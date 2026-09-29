//nolint:testpackage // binds unexported serve/dispatchSignals and the engine's sync seam — the sanctioned in-package corpus (conn_switcher_test.go precedent)
package engine

import (
	"context"
	"encoding/binary"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

// foreignMember and foreignPath name signal traffic that must NOT reach the
// sync listener: the dispatcher filters on the exact IBus path/member pair.
var (
	foreignMember = "NameOwnerChanged"
	foreignPath   = dbus.ObjectPath("/org/freedesktop/DBus")
	foreignIface  = "org.freedesktop.DBus"
	wrongPath     = dbus.ObjectPath("/org/freedesktop/IBus/Engine/1")
)

// emitSignal writes one raw signal message to the stand's live client —
// the direction the real bus fans GlobalEngineChanged out in (spike P2:
// AddMatch works on the ibus socket). Best-effort: a write to a dead
// client simply fails and is ignored.
func (fb *fakeBus) emitSignal(member string, path dbus.ObjectPath, iface string, body ...any) {
	c := fb.clientConn()
	if c == nil {
		return
	}
	msg := &dbus.Message{
		Type:    dbus.TypeSignal,
		Flags:   0,
		Body:    body,
		Headers: map[dbus.HeaderField]dbus.Variant{},
	}
	msg.Headers[dbus.FieldPath] = dbus.MakeVariant(path)
	msg.Headers[dbus.FieldInterface] = dbus.MakeVariant(iface)
	msg.Headers[dbus.FieldMember] = dbus.MakeVariant(member)
	if len(body) > 0 {
		msg.Headers[dbus.FieldSignature] = dbus.MakeVariant(dbus.SignatureOf(body...))
	}
	fb.writeMu.Lock()
	defer fb.writeMu.Unlock()
	_ = msg.EncodeTo(c, binary.LittleEndian)
}

// awaitSeamSignal drives one serve generation to its post-RequestName seam
// where the signal stream is already plumbed (serve registers the channel
// before the seams — an emission after the seam cannot be lost).
func awaitSeamSignal(t *testing.T, reached <-chan struct{}, errCh <-chan error) {
	t.Helper()
	awaitSeam(t, reached, errCh, 0)
}

// drainDispatched empties the dispatch channel without blocking.
func drainDispatched(ch chan string) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

// TestWaitBusLossVerdictsPreserved pins the serve-contour verdicts through
// the 05-04 dispatcher refactor: ctx-done → nil, closed channel →
// errBusClosed — byte-identical to the pre-phase drain (the corpus's
// regression pin; the dispatcher's new behavior has its own tests below).
func TestWaitBusLossVerdictsPreserved(t *testing.T) {
	t.Run("context done returns nil", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		signals := make(chan *dbus.Signal, 1)
		errCh := make(chan error, 1)
		go func() { errCh <- dispatchSignals(ctx, &Config{}, signals) }()
		cancel()
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("dispatchSignals() = %v, want nil on ctx done", err)
			}
		case <-time.After(serveBudget):
			t.Fatal("dispatchSignals never returned after ctx cancellation")
		}
	})

	t.Run("closed channel returns errBusClosed", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		signals := make(chan *dbus.Signal, 1)
		errCh := make(chan error, 1)
		go func() { errCh <- dispatchSignals(ctx, &Config{}, signals) }()
		close(signals)
		select {
		case err := <-errCh:
			if !errors.Is(err, errBusClosed) {
				t.Fatalf("dispatchSignals() = %v, want errBusClosed on channel close", err)
			}
		case <-time.After(serveBudget):
			t.Fatal("dispatchSignals never returned after channel close")
		}
	})
}

// TestGlobalEngineChangedDispatched pins the sync listener (criterion 3)
// through the live stand: a GlobalEngineChanged signal on the IBus
// path/interface reaches Config.OnGlobalEngine with its engine-name payload
// (spike P2), while foreign traffic — another member, another path — never
// dispatches. The serve generation exits nil on cancellation afterwards:
// the dispatcher must not block shutdown.
func TestGlobalEngineChangedDispatched(t *testing.T) {
	captureJSONLogs(t) // quiet the serve INFO records
	dispatched := make(chan string, 16)
	reached := make(chan struct{}, 1)
	cfg := seamConfig(func(func(context.Context, string) error) { reached <- struct{}{} })
	cfg.OnGlobalEngine = func(name string) { dispatched <- name }
	bus := newFakeBus(t)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	errCh := serveOnce(ctx, cfg, 0)
	awaitSeamSignal(t, reached, errCh)

	bus.emitSignal("GlobalEngineChanged", ibusPath, ibusService, NameRU)
	select {
	case name := <-dispatched:
		if name != NameRU {
			t.Fatalf("OnGlobalEngine = %q, want %s", name, NameRU)
		}
	case <-time.After(serveBudget):
		t.Fatal("GlobalEngineChanged never dispatched — the sync listener is dead")
	}
	drainDispatched(dispatched)

	bus.emitSignal(foreignMember, foreignPath, foreignIface, ":1.a", "", ":1.b")
	bus.emitSignal("GlobalEngineChanged", wrongPath, ibusService, NameEN)
	select {
	case name := <-dispatched:
		t.Errorf("foreign signal dispatched (%q) — the path/member filter is broken", name)
	case <-time.After(100 * time.Millisecond):
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("serve() = %v, want a clean nil on ctx cancel", err)
		}
	case <-time.After(serveBudget):
		t.Fatal("serve never returned after cancellation — the dispatcher blocks shutdown")
	}
}

// TestSignalStormBounded pins the storm mitigation (T-05-04-02): the
// connection's signal handler drops on a full registered channel (godbus's
// default handler instead defers a blocking goroutine per overflow signal —
// unbounded memory under a storm), warns ONCE (bounded WARN), the dispatcher
// keeps consuming, and serve still exits nil on cancellation. The first
// dispatch is held in a gate so the buffer provably fills and the drops
// provably happen.
func TestSignalStormBounded(t *testing.T) {
	buf := captureJSONLogs(t)
	dispatched := make(chan string, signalBufferSize+8)
	gate := make(chan struct{})
	entered := make(chan struct{}, 1)
	var once sync.Once
	reached := make(chan struct{}, 1)
	cfg := seamConfig(func(func(context.Context, string) error) { reached <- struct{}{} })
	cfg.OnGlobalEngine = func(name string) {
		once.Do(func() { close(entered) })
		<-gate
		dispatched <- name
	}
	bus := newFakeBus(t)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	errCh := serveOnce(ctx, cfg, 0)
	awaitSeamSignal(t, reached, errCh)

	// One signal: the dispatcher picks it up and parks inside the callback.
	bus.emitSignal("GlobalEngineChanged", ibusPath, ibusService, NameRU)
	select {
	case <-entered:
	case <-time.After(serveBudget):
		t.Fatal("the dispatcher never picked up the first signal")
	}

	// The storm: far more than the buffer holds while the consumer is
	// parked — every delivery beyond one buffered is a guaranteed drop.
	const storm = 64
	for range storm {
		bus.emitSignal("GlobalEngineChanged", ibusPath, ibusService, NameRU)
	}
	// Drain window: a desktop-class unix socket moves these tiny messages
	// in microseconds; the wait makes "all storm messages processed" the
	// operative assumption for the drop arithmetic below.
	time.Sleep(200 * time.Millisecond)

	close(gate)
	drainDispatched(dispatched)
	// The parked first dispatch plus everything still in the bounded
	// buffer — everything else was dropped: memory stayed at the buffer.
	total := len(dispatched)
	if total > signalBufferSize+1 {
		t.Errorf("dispatched %d of %d storm signals — the memory bound is broken", total, storm)
	}
	if total == 0 {
		t.Error("no storm signal dispatched — the dispatcher died in the storm")
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("serve() = %v, want a clean nil on ctx cancel", err)
		}
	case <-time.After(serveBudget):
		t.Fatal("serve never returned after the storm — the dispatcher blocked shutdown")
	}

	// The drop policy warns exactly once: a storm must not flood the journal.
	if got := strings.Count(buf.String(), `"msg":"signal dropped"`); got != 1 {
		t.Errorf("signal-dropped WARN count = %d, want exactly 1; log:\n%s", got, buf.String())
	}
}

// TestFocusInForwardsEngineName pins the FocusIn sync input (criterion 3):
// FocusIn carries THIS engine's wire name into the sync seam — the second
// observer feeding SyncEngine — and the recoverHandler contract wraps the
// path: a panicking sync callback stays contained (safe nil reply).
func TestFocusInForwardsEngineName(t *testing.T) {
	t.Run("focus in forwards the engine name", func(t *testing.T) {
		var mu sync.Mutex
		var got []string
		eng := NewEngine(nil, NameRU)
		eng.onGlobalEngine = func(name string) {
			mu.Lock()
			defer mu.Unlock()
			got = append(got, name)
		}

		if err := eng.FocusIn(); err != nil {
			t.Fatalf("FocusIn() err = %v, want nil", err)
		}
		mu.Lock()
		defer mu.Unlock()
		if len(got) != 1 || got[0] != NameRU {
			t.Errorf("sync seam got %q, want exactly [%s] — FocusIn must carry the engine name", got, NameRU)
		}
	})

	t.Run("panicking sync callback is contained", func(t *testing.T) {
		captureJSONLogs(t) // contain the ERROR record of the injected panic
		eng := NewEngine(nil, NameEN)
		eng.onGlobalEngine = func(string) { panic("injected sync panic") }

		if err := eng.FocusIn(); err != nil {
			t.Errorf("FocusIn() err = %v, want the safe nil reply (recoverHandler must wrap the sync path)", err)
		}
	})
}

// readerRecorder is the BindGlobalEngine seam double: it counts the binds
// and keeps each generation's reader closure.
type readerRecorder struct {
	mu    sync.Mutex
	binds int
	fns   []func(ctx context.Context) (string, error)
}

func (r *readerRecorder) bind(fn func(ctx context.Context) (string, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.binds++
	r.fns = append(r.fns, fn)
}

func (r *readerRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.binds
}

func (r *readerRecorder) closure(t *testing.T, gen int) func(ctx context.Context) (string, error) {
	t.Helper()

	r.mu.Lock()
	defer r.mu.Unlock()

	if gen >= len(r.fns) {
		t.Fatalf("BindGlobalEngine fired %d times, want generation %d — the seam never bound", len(r.fns), gen)
	}

	return r.fns[gen]
}

// TestBindGlobalEnginePerGeneration pins the generation-scoped reader seam
// (Pitfall 3): serve hands BindGlobalEngine a FRESH GetGlobalEngine reader
// on every generation that cleared RequestName, bound BEFORE PostRegister —
// the same-generation reactivation must consult the reader — and the reader
// parses the engine name out of the variant-wrapped EngineDesc (wire order
// of types.go: positional field 2). The nil seam stays a no-op: the
// pre-seam cycle of TestSwitcherNilNoOp runs with BindGlobalEngine nil.
func TestBindGlobalEnginePerGeneration(t *testing.T) {
	captureJSONLogs(t) // quiet the serve INFO records
	rec := &readerRecorder{}
	reached := make(chan struct{}, 8)
	cfg := seamConfig(func(func(context.Context, string) error) { reached <- struct{}{} })
	cfg.BindGlobalEngine = func(get func(context.Context) (string, error)) {
		rec.bind(get)
		reached <- struct{}{}
	}

	callCtx, callCancel := context.WithTimeout(context.Background(), callBudget)
	defer callCancel()

	bus0 := newFakeBus(t) // generation 0
	bus0.setGlobalEngine(NameRU)
	ctx0, cancel0 := context.WithCancel(context.Background())
	errCh := serveOnce(ctx0, cfg, 0)
	awaitSeam(t, reached, errCh, 0)
	name, err := rec.closure(t, 0)(callCtx)
	if err != nil {
		t.Fatalf("generation-0 reader = %v, want a clean read", err)
	}
	if name != NameRU {
		t.Errorf("generation-0 reader = %q, want %s — the desc-variant parse is broken", name, NameRU)
	}
	cancel0()
	if err := <-errCh; err != nil {
		t.Fatalf("serve(gen 0) = %v, want a clean nil on ctx cancel", err)
	}

	bus1 := newFakeBus(t) // a new ibus-daemon generation
	bus1.setGlobalEngine(NameEN)
	ctx1, cancel1 := context.WithCancel(context.Background())
	errCh = serveOnce(ctx1, cfg, 1)
	awaitSeam(t, reached, errCh, 1)
	if got := rec.count(); got != 2 {
		t.Fatalf("BindGlobalEngine fired %d times across two generations, want exactly 2", got)
	}
	name, err = rec.closure(t, 1)(callCtx)
	if err != nil {
		t.Fatalf("generation-1 reader = %v, want a live read on the new bus", err)
	}
	if name != NameEN {
		t.Errorf("generation-1 reader = %q, want %s — a dead generation's reader must not survive", name, NameEN)
	}
	cancel1()
	if err := <-errCh; err != nil {
		t.Fatalf("serve(gen 1) = %v, want a clean nil on ctx cancel", err)
	}
}
