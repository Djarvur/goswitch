//nolint:testpackage // binds unexported serve — the sanctioned in-package seam (02-PATTERNS § engine)
package engine

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

// serveBudget bounds every wait on a serve generation: a healthy fake bus
// completes a generation in milliseconds, so the budget only turns a missing
// bind point into an assertion failure instead of a hung suite.
const serveBudget = 5 * time.Second

// callBudget bounds one switcher round trip against the fake bus.
const callBudget = 5 * time.Second

// errSaslHandshake is the stand's static refusal: the client left the SASL
// script the fake bus answers.
var errSaslHandshake = errors.New("fake bus: unexpected sasl line")

// errInjectedFail is the stand's injected SetGlobalEngine failure — the
// Warn path's observable cause without a wedged bus.
var errInjectedFail = errors.New("fake bus: injected switch failure")

// errFailMember is the D-Bus error name of the injected SetGlobalEngine
// failure — the shape every bus error reply carries.
const errFailMember = "org.freedesktop.DBus.Error.Failed"

// switcherRecorder is the BindSwitcher seam double: it counts the binds and
// keeps each generation's closure — the corpus's assertion surface for the
// per-generation rebinding (T-05-03-03).
type switcherRecorder struct {
	mu    sync.Mutex
	binds int
	fns   []func(ctx context.Context, engineName string) error
}

// bind is the Config.BindSwitcher callback: record the closure.
func (r *switcherRecorder) bind(fn func(ctx context.Context, engineName string) error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.binds++
	r.fns = append(r.fns, fn)
}

// count snapshots the bind count.
func (r *switcherRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.binds
}

// closure returns generation gen's closure, failing the test when the seam
// never bound that many generations.
func (r *switcherRecorder) closure(t *testing.T, gen int) func(ctx context.Context, engineName string) error {
	t.Helper()

	r.mu.Lock()
	defer r.mu.Unlock()

	if gen >= len(r.fns) {
		t.Fatalf("BindSwitcher fired %d times, want generation %d — the seam never bound", len(r.fns), gen)
	}

	return r.fns[gen]
}

// seamConfig builds the registration payload of the seam corpus with the
// seam bound to bind (nil = the nil-seam test's world).
func seamConfig(bind func(func(context.Context, string) error)) *Config {
	engines := []EngineDesc{
		NewEngineDesc(NameEN, "goswitch English (US)", "en", "us", "en"),
		NewEngineDesc(NameRU, "goswitch Русская", "ru", "ru", "ru"),
	}

	return &Config{
		Component:    NewComponent(engines),
		Engines:      engines,
		BindSwitcher: bind,
	}
}

// serveOnce starts one serve generation on its own goroutine; the caller
// cancels the context after the seam fired and reads the returned channel
// for the clean nil.
func serveOnce(ctx context.Context, cfg *Config, generation int) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- serve(ctx, cfg, generation)
	}()

	return errCh
}

// awaitSeam blocks until the generation reached its post-RequestName seam —
// or fails the test when serve died first or the budget expired.
func awaitSeam(t *testing.T, reached <-chan struct{}, errCh <-chan error, generation int) {
	t.Helper()

	select {
	case <-reached:
	case err := <-errCh:
		t.Fatalf("serve(gen %d) returned before the seam fired: %v", generation, err)
	case <-time.After(serveBudget):
		t.Fatalf("serve(gen %d) never reached the post-RequestName seam", generation)
	}
}

// boundSeam drives one serve generation against a fresh stand and returns
// the bound closure plus the stand — the shared setup of the seam corpus.
func boundSeam(t *testing.T) (func(ctx context.Context, engineName string) error, *fakeBus) {
	t.Helper()

	captureJSONLogs(t) // quiet the serve INFO records
	rec := &switcherRecorder{}
	reached := make(chan struct{}, 1)
	bus := newFakeBus(t)
	cfg := seamConfig(func(bind func(context.Context, string) error) {
		rec.bind(bind)
		reached <- struct{}{}
	})
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	awaitSeam(t, reached, serveOnce(ctx, cfg, 0), 0)

	return rec.closure(t, 0), bus
}

// syncLogBuffer is a mutex-guarded JSON log buffer: serve logs from its own
// goroutine while the corpus reads the records back.
type syncLogBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

// Write appends under the guard.
func (b *syncLogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

// String snapshots the buffer under the guard.
func (b *syncLogBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

// captureJSONLogs redirects the default logger into a guarded JSON buffer
// and restores it at cleanup — the journal-record corpus's reading surface.
func captureJSONLogs(t *testing.T) *syncLogBuffer {
	t.Helper()

	buf := &syncLogBuffer{}
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	return buf
}

// fakeBus is a minimal in-process ibus-daemon stand-in: a unix listener
// whose accepted connections walk a hand-rolled EXTERNAL SASL server (godbus
// ships no server side) and then a hand-rolled method-call dispatch loop
// built on godbus's exported DecodeMessage/EncodeTo — answering the three
// calls serve makes (Hello, RequestName, RegisterComponent) plus
// SetGlobalEngine, whose targets the stand records. Torn down with the test.
type fakeBus struct {
	done chan struct{}

	mu       sync.Mutex
	switches []string
	fail     bool

	// writeMu serializes every write to the client connection: the
	// dispatch loop answers method calls from its own goroutine while the
	// sync corpus emits raw signal messages from the test goroutine (the
	// direction the real bus fans GlobalEngineChanged out in — spike P2).
	writeMu sync.Mutex
	// client is the live client connection (the most recent accepted one);
	// nil when no client is connected.
	client net.Conn
	// globalEngine is what the stand's GetGlobalEngine answers with (the
	// wire name inside a variant-wrapped EngineDesc); empty = refuses.
	globalEngine string
}

// setClient records the live client connection under the write guard.
func (fb *fakeBus) setClient(c net.Conn) {
	fb.writeMu.Lock()
	defer fb.writeMu.Unlock()

	fb.client = c
}

// clientConn snapshots the live client connection (nil when disconnected).
func (fb *fakeBus) clientConn() net.Conn {
	fb.writeMu.Lock()
	defer fb.writeMu.Unlock()

	return fb.client
}

// newFakeBus starts the stand and points Discover's IBUS_ADDRESS override at
// its socket — the same env the live socket uses, so serve's address
// discipline is exercised unchanged.
func newFakeBus(t *testing.T) *fakeBus {
	t.Helper()

	sock := filepath.Join(t.TempDir(), "ibus-fake.sock")
	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "unix", sock)
	if err != nil {
		t.Fatalf("listen fake ibus socket: %v", err)
	}
	fb := &fakeBus{done: make(chan struct{})}
	t.Cleanup(func() {
		close(fb.done)
		_ = ln.Close()
	})
	go fb.accept(ln)
	t.Setenv("IBUS_ADDRESS", "unix:path="+sock)

	return fb
}

// accept serves the listener until the test's cleanup closes it.
func (fb *fakeBus) accept(ln net.Listener) {
	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		go fb.handle(c)
	}
}

// handle walks one connection through SASL, then dispatches its method
// calls until the connection dies or the test's cleanup tears the stand down.
func (fb *fakeBus) handle(c net.Conn) {
	defer func() { _ = c.Close() }()

	fb.setClient(c)
	defer fb.setClient(nil)

	br := bufio.NewReader(c)
	if err := fb.sasl(c, br); err != nil {
		return
	}
	fb.dispatch(c, br)
}

// sasl answers the minimal EXTERNAL server half of godbus's client auth:
// \0 + AUTH → REJECTED EXTERNAL, the EXTERNAL line → OK, the fd negotiation
// → ERROR, BEGIN → done — the same script conn.Auth(userAuth()) drives on
// serve's client side.
func (fb *fakeBus) sasl(c net.Conn, br *bufio.Reader) error {
	var nul [1]byte
	if _, err := io.ReadFull(br, nul[:]); err != nil {
		return err
	}
	if _, err := saslLine(br); err != nil { // "AUTH"
		return err
	}
	if _, err := io.WriteString(c, "REJECTED EXTERNAL\r\n"); err != nil {
		return err
	}
	if _, err := saslLine(br); err != nil { // "AUTH EXTERNAL <hex uid>"
		return err
	}
	if _, err := io.WriteString(c, "OK 0123456789abcdef\r\n"); err != nil {
		return err
	}
	line, err := saslLine(br) // "NEGOTIATE_UNIX_FD" or "BEGIN"
	if err != nil {
		return err
	}
	if strings.HasPrefix(line, "NEGOTIATE_UNIX_FD") {
		if _, err := io.WriteString(c, "ERROR\r\n"); err != nil { // no fd passing
			return err
		}
		if line, err = saslLine(br); err != nil { // "BEGIN"
			return err
		}
	}
	if line != "BEGIN" {
		return errSaslHandshake
	}

	return nil
}

// saslLine reads one \r\n-terminated SASL line.
func saslLine(br *bufio.Reader) (string, error) {
	line, err := br.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimRight(line, "\r\n"), nil
}

// dispatch reads method calls and answers them until the connection dies:
// Hello and RequestName at the bus object (unique name, primary owner — the
// single-instance guard must pass for the generation to register),
// RegisterComponent (accepted unwatched — the payload is the variant-wrapped
// wire struct), SetGlobalEngine (the D-52 act itself — recorded, verdict
// per the fail flag), AddMatch/RemoveMatch (the match-rule bookkeeping the
// bus object serves) and GetGlobalEngine (the factual-engine reader of
// 05-04). Every reply carries the call's serial — the shape the client's
// pending-call tracker matches on.
func (fb *fakeBus) dispatch(c net.Conn, br *bufio.Reader) {
	for {
		msg, err := dbus.DecodeMessage(br)
		if err != nil {
			return
		}
		if msg.Type != dbus.TypeMethodCall {
			continue
		}
		reply := fb.answer(msg)
		reply.Headers[dbus.FieldReplySerial] = dbus.MakeVariant(msg.Serial())
		fb.writeMu.Lock()
		err = reply.EncodeTo(c, binary.LittleEndian)
		fb.writeMu.Unlock()
		if err != nil {
			return
		}
	}
}

// answer maps one method call to its reply message.
func (fb *fakeBus) answer(msg *dbus.Message) *dbus.Message {
	member, _ := msg.Headers[dbus.FieldMember].Value().(string)
	switch member {
	case "Hello":
		return replyMsg(dbus.TypeMethodReply, "", ":1.fake")
	case "RequestName":
		return replyMsg(dbus.TypeMethodReply, "", dbus.RequestNameReplyPrimaryOwner)
	case "RegisterComponent":
		return replyMsg(dbus.TypeMethodReply, "", nil)
	case "SetGlobalEngine":
		name, _ := msg.Body[0].(string)
		fb.mu.Lock()
		fb.switches = append(fb.switches, name)
		fail := fb.fail
		fb.mu.Unlock()
		if fail {
			return replyMsg(dbus.TypeError, errFailMember, errInjectedFail.Error())
		}

		return replyMsg(dbus.TypeMethodReply, "", nil)
	case "AddMatch", "RemoveMatch":
		// The match-rule bookkeeping: the real bus confirms with an empty
		// reply — the forwarding the rule buys is the live side (spike P2).
		return replyMsg(dbus.TypeMethodReply, "", nil)
	case "GetGlobalEngine":
		fb.mu.Lock()
		name := fb.globalEngine
		fb.mu.Unlock()
		if name == "" {
			return replyMsg(dbus.TypeError, errFailMember, "fake bus: no global engine")
		}
		desc := NewEngineDesc(name, "goswitch stand", "xx", "us", "xx")

		return replyMsg(dbus.TypeMethodReply, "", dbus.MakeVariant(desc))
	default:
		// An unknown member must fail the caller fast, never hang it.
		return replyMsg(dbus.TypeError, errFailMember, "fake bus: unknown member "+member)
	}
}

// replyMsg builds a reply (or error) message with the given body; the
// reply-serial header is the dispatcher's job.
func replyMsg(kind dbus.Type, errName, body any) *dbus.Message {
	msg := &dbus.Message{
		Type:    kind,
		Flags:   0,
		Body:    []any{},
		Headers: map[dbus.HeaderField]dbus.Variant{},
	}
	if body != nil {
		msg.Body = []any{body}
	}
	if len(msg.Body) > 0 {
		msg.Headers[dbus.FieldSignature] = dbus.MakeVariant(dbus.SignatureOf(msg.Body...))
	}
	if errName != "" {
		msg.Headers[dbus.FieldErrorName] = dbus.MakeVariant(errName)
	}

	return msg
}

// switchTargets snapshots the recorded flip targets.
func (fb *fakeBus) switchTargets() []string {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	return append([]string(nil), fb.switches...)
}

// setFail flips the verdict of the subsequent SetGlobalEngine replies.
func (fb *fakeBus) setFail(on bool) {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	fb.fail = on
}

// setGlobalEngine fixes what the stand's GetGlobalEngine answers with —
// the factual active engine of the 05-04 reader (Pitfall 3).
func (fb *fakeBus) setGlobalEngine(name string) {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	fb.globalEngine = name
}

// TestSwitcherBindsPerGeneration pins the generation-scoped rebinding
// (T-05-03-03): serve hands the seam a FRESH closure on every generation
// that cleared RequestName — generation 0 and the reconnect alike — and a
// dead generation's closure is never re-issued: after the reconnect the old
// closure's connection is closed (its call fails) while the new closure
// works against the new bus.
func TestSwitcherBindsPerGeneration(t *testing.T) {
	captureJSONLogs(t) // quiet the serve INFO records
	rec := &switcherRecorder{}
	reached := make(chan struct{}, 4)
	cfg := seamConfig(func(bind func(context.Context, string) error) {
		rec.bind(bind)
		reached <- struct{}{}
	})

	_ = newFakeBus(t) // generation 0
	ctx0, cancel0 := context.WithCancel(context.Background())
	errCh := serveOnce(ctx0, cfg, 0)
	awaitSeam(t, reached, errCh, 0)
	cancel0()
	if err := <-errCh; err != nil {
		t.Fatalf("serve(gen 0) = %v, want a clean nil on ctx cancel", err)
	}

	bus1 := newFakeBus(t) // a new ibus-daemon generation — a new socket, a new address
	ctx1, cancel1 := context.WithCancel(context.Background())
	errCh = serveOnce(ctx1, cfg, 1)
	awaitSeam(t, reached, errCh, 1)

	if got := rec.count(); got != 2 {
		t.Fatalf("BindSwitcher fired %d times across two generations, want exactly 2", got)
	}
	// The rebinding proof runs while generation 1 is still serving: the
	// generation-0 closure's connection is already closed (serve returned
	// and closed it), the generation-1 closure's is live.
	flip0, flip1 := rec.closure(t, 0), rec.closure(t, 1)
	callCtx, callCancel := context.WithTimeout(context.Background(), callBudget)
	defer callCancel()
	if err := flip0(callCtx, NameEN); err == nil {
		t.Error("the generation-0 switcher still answers after the reconnect — a dead generation's seam must be dead")
	}
	if err := flip1(callCtx, NameRU); err != nil {
		t.Errorf("generation-1 switcher(%q) = %v, want a live call on the new bus", NameRU, err)
	}
	if got := bus1.switchTargets(); !slices.Equal(got, []string{NameRU}) {
		t.Errorf("reconnected bus targets = %q, want exactly [%s]", got, NameRU)
	}

	cancel1()
	if err := <-errCh; err != nil {
		t.Fatalf("serve(gen 1) = %v, want a clean nil on ctx cancel", err)
	}
}

// TestSwitcherCallsSetGlobalEngine pins the D-52 act: the bound closure
// calls org.freedesktop.IBus.SetGlobalEngine with the engine name as its
// single string argument — the round trip through the real godbus marshaller
// and the stand's typed decode is the serialization proof (the
// emitter_wire_test discipline, transport side).
func TestSwitcherCallsSetGlobalEngine(t *testing.T) {
	flip, bus := boundSeam(t)

	callCtx, callCancel := context.WithTimeout(context.Background(), callBudget)
	defer callCancel()
	if err := flip(callCtx, NameRU); err != nil {
		t.Fatalf("switcher(%q) = %v, want a clean bus call", NameRU, err)
	}
	if err := flip(callCtx, NameEN); err != nil {
		t.Fatalf("switcher(%q) = %v, want a clean bus call", NameEN, err)
	}
	if got := bus.switchTargets(); !slices.Equal(got, []string{NameRU, NameEN}) {
		t.Errorf("SetGlobalEngine targets = %q, want exactly [%s %s] — one name per call", got, NameRU, NameEN)
	}
}

// TestSwitcherJournalRecords pins the flip journal (T-05-03-02): a clean
// call writes the switch_engine record at INFO, a failed call the SAME
// record at WARN with the cause — and the engine name (a config literal,
// D-20/D-21) is the only payload: the record's key set admits no user text.
func TestSwitcherJournalRecords(t *testing.T) {
	flip, bus := boundSeam(t)
	buf := captureJSONLogs(t) // after the setup — its own capture owns the default logger now

	callCtx, callCancel := context.WithTimeout(context.Background(), callBudget)
	defer callCancel()
	if err := flip(callCtx, NameRU); err != nil {
		t.Fatalf("switcher(%q) = %v, want a clean call", NameRU, err)
	}
	bus.setFail(true)
	if err := flip(callCtx, NameEN); err == nil {
		t.Fatalf("switcher(%q) = nil, want the injected failure", NameEN)
	}

	logged := buf.String()
	ruLine, ok := journalLine(logged, NameRU)
	if !ok {
		t.Fatalf("switch_engine record for %s missing; log:\n%s", NameRU, logged)
	}
	enLine, ok := journalLine(logged, NameEN)
	if !ok {
		t.Fatalf("switch_engine record for %s missing; log:\n%s", NameEN, logged)
	}
	if !strings.Contains(ruLine, `"level":"INFO"`) {
		t.Errorf("success record not at INFO: %s", ruLine)
	}
	if !strings.Contains(enLine, `"level":"WARN"`) {
		t.Errorf("failure record not at WARN: %s", enLine)
	}
	if !strings.Contains(enLine, `"err":`) {
		t.Errorf("failure record carries no cause: %s", enLine)
	}
	// The no-user-text pin (D-20/D-21), structural: the record's keys admit
	// only the envelope, the engine literal and the error cause.
	for _, line := range []string{ruLine, enLine} {
		for key := range jsonKeys(t, line) {
			switch key {
			case "time", "level", "msg", "engine", "err":
			default:
				t.Errorf("switch_engine record carries key %q — only the engine literal"+
					" and the cause belong: %s", key, line)
			}
		}
	}
}

// journalLine returns the switch_engine log line naming the given engine.
func journalLine(logged, engine string) (string, bool) {
	for _, line := range strings.Split(logged, "\n") {
		if strings.Contains(line, `"msg":"switch_engine"`) && strings.Contains(line, `"engine":"`+engine+`"`) {
			return line, true
		}
	}

	return "", false
}

// jsonKeys extracts the key set of one JSON log line.
func jsonKeys(t *testing.T, line string) map[string]any {
	t.Helper()

	var rec map[string]any
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		t.Fatalf("log line is not JSON: %q: %v", line, err)
	}

	return rec
}

// TestSwitcherNilNoOp pins the seam's nil contract (the PostRegister shape):
// without BindSwitcher the serve cycle is exactly the pre-seam one —
// register, PostRegister, park — and never panics on the absent seam.
func TestSwitcherNilNoOp(t *testing.T) {
	captureJSONLogs(t) // quiet the serve INFO records
	reached := make(chan struct{}, 1)
	sewn := make(chan int, 1)
	cfg := seamConfig(nil)
	cfg.PostRegister = func(_ context.Context, generation int) {
		sewn <- generation
		reached <- struct{}{}
	}
	_ = newFakeBus(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	awaitSeam(t, reached, serveOnce(ctx, cfg, 0), 0)
	if gen := <-sewn; gen != 0 {
		t.Errorf("PostRegister generation = %d, want 0 — the pre-seam cycle must be intact", gen)
	}
}

// TestNameConstants pins the engine-name literals (D-20): config-level
// constants, the single source the flip path derives its SetGlobalEngine
// target from and the journal records name.
func TestNameConstants(t *testing.T) {
	if NameEN != "goswitch-en" {
		t.Errorf("NameEN = %q, want %q", NameEN, "goswitch-en")
	}
	if NameRU != "goswitch-ru" {
		t.Errorf("NameRU = %q, want %q", NameRU, "goswitch-ru")
	}
}
