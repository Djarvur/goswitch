package ctlsvc_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/Djarvur/goswitch/engine"
	"github.com/Djarvur/goswitch/internal/config"
	"github.com/Djarvur/goswitch/internal/ctlsvc"
	"github.com/Djarvur/goswitch/internal/session"
)

// ctlTestWindow keeps the actor's real AfterFunc out of every test (the
// session corpus's farWindow discipline): the forced correction drives the
// pipeline directly, no tap series is ever armed.
const ctlTestWindow = time.Hour

// configFilePerm is the temp config's mode (mnd).
const configFilePerm = 0o600

// ctlWordEN/ctlWordRU are the corpus pair of the forced correction (the
// 02-01 pair, named once for goconst).
const (
	ctlWordEN = "ghbdtn"
	ctlWordRU = "привет"
)

// fakeStatus is the snapshot-provider double of the service corpus: one
// canned snapshot, returned verbatim.
type fakeStatus struct {
	snap session.Status
}

// StatusSnapshot returns the canned snapshot.
func (f fakeStatus) StatusSnapshot() session.Status { return f.snap }

// switchableStatus flips between a healthy snapshot and a panicking one —
// the recover-shim corpus drives the SAME service through both.
type switchableStatus struct {
	panicNext bool
}

// StatusSnapshot panics on demand — the panic the shim must contain.
func (s *switchableStatus) StatusSnapshot() session.Status {
	if s.panicNext {
		panic("snapshot exploded")
	}

	return session.Status{Mode: "en", ConfigValid: true}
}

// ctlSink is the engine.Emitter double of the forced-correction corpus:
// every emitter call recorded under a mutex (the guarded-sink style of the
// session corpus's fakeSink).
type ctlSink struct {
	mu       sync.Mutex
	requires int
	deletes  int
	commits  []string
}

// RequireSurroundingText records the verification request.
func (c *ctlSink) RequireSurroundingText() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.requires++
}

// DeleteSurroundingText records the ladder deletion.
func (c *ctlSink) DeleteSurroundingText(offset int32, nchars uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deletes++
}

// ForwardKeyEvent records nothing the corpus asserts on — the word path
// emits none.
func (c *ctlSink) ForwardKeyEvent(keyval, keycode, state uint32) {}

// CommitText records the committed payload.
func (c *ctlSink) CommitText(text engine.IBusText) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.commits = append(c.commits, text.Text)
}

// requireCount snapshots the verification-request count.
func (c *ctlSink) requireCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.requires
}

// deleteCount snapshots the deletion count.
func (c *ctlSink) deleteCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.deletes
}

// commitTexts snapshots the committed payloads.
func (c *ctlSink) commitTexts() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return append([]string(nil), c.commits...)
}

// TestSvc_Status pins the Status rendering: the single-line key=value
// report carries the mode, the correction counters with their skip-reason
// breakdown, the Super interception counters and the config string (path +
// validity + the D-32 last-good error) — counts and states only, never
// user text (T-03-06-03). The no-config daemon renders config=none.
func TestSvc_Status(t *testing.T) {
	svc := ctlsvc.NewSvc(ctlsvc.Deps{Status: fakeStatus{snap: session.Status{
		Mode:                  "ru",
		CorrectionsDone:       3,
		CorrectionsSkipped:    1,
		SkipReasons:           map[string]int{"empty-buffer": 1},
		SuperIntercepted:      2,
		SuperUpstreamConsumed: 4,
		ConfigPath:            "/tmp/goswitch-ctl-status.yaml",
		ConfigValid:           false,
		ConfigError:           "decode config: field verify_wait_mss not found",
	}}})

	reply, err := svc.Status()
	if err != nil {
		t.Fatalf("Status error = %v, want nil", err)
	}
	for _, want := range []string{
		"mode=ru",
		"corrections_done=3",
		"corrections_skipped=1",
		"skip_empty_buffer=1",
		"super_intercepted=2",
		"super_upstream_consumed=4",
		"config_path=/tmp/goswitch-ctl-status.yaml",
		"config_valid=false",
		"config_error=decode config: field verify_wait_mss not found",
	} {
		if !strings.Contains(reply, want) {
			t.Errorf("Status reply %q missing %q", reply, want)
		}
	}

	t.Run("no config renders config=none", func(t *testing.T) {
		svc := ctlsvc.NewSvc(ctlsvc.Deps{Status: fakeStatus{}})
		reply, err := svc.Status()
		if err != nil {
			t.Fatalf("Status error = %v, want nil", err)
		}
		if !strings.Contains(reply, "config=none") {
			t.Errorf("no-config Status reply %q missing config=none", reply)
		}
	})
}

// TestSvc_ReloadConfig pins the D-32 behavior through the control surface:
// a valid document is applied (the served snapshot moves, the reply says
// so); a rejected one keeps the last-good serving, answers with the error
// naming the offending key and leaves it visible through LastError — the
// status surface's raw material.
func TestSvc_ReloadConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "ctl-reload.yaml")
	if err := os.WriteFile(cfgPath, []byte(ctlConfigYAML(300)), configFilePerm); err != nil {
		t.Fatalf("write config: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// The debounce never fires inside the test (time.Hour) — Reload is the
	// synchronous driver under test; the watcher's own event path has its
	// own corpus (03-02).
	w, err := config.NewWatcher(ctx, cfgPath, config.WithDebounce(time.Hour))
	if err != nil {
		t.Fatalf("new watcher: %v", err)
	}
	svc := ctlsvc.NewSvc(ctlsvc.Deps{Reload: w})

	t.Run("valid reload applies", func(t *testing.T) {
		if err := os.WriteFile(cfgPath, []byte(ctlConfigYAML(450)), configFilePerm); err != nil {
			t.Fatalf("rewrite config: %v", err)
		}
		reply, rerr := svc.ReloadConfig()
		if rerr != nil {
			t.Fatalf("ReloadConfig error = %v, want nil", rerr)
		}
		if !strings.Contains(reply, "applied") {
			t.Errorf("ReloadConfig reply %q missing applied", reply)
		}
		if got := w.Snapshot().Timeouts.TapWindowMs; got != 450 {
			t.Errorf("served window after reload = %d, want 450", got)
		}
		if lerr := w.LastError(); lerr != nil {
			t.Errorf("LastError after valid reload = %v, want nil", lerr)
		}
	})

	t.Run("invalid reload keeps last-good and surfaces the error", func(t *testing.T) {
		if err := os.WriteFile(cfgPath, []byte(ctlBrokenYAML), configFilePerm); err != nil {
			t.Fatalf("break config: %v", err)
		}
		reply, rerr := svc.ReloadConfig()
		if rerr == nil {
			t.Fatalf("ReloadConfig error = nil (reply %q), want the rejection", reply)
		}
		if !strings.Contains(rerr.Error(), "verify_wait_mss") {
			t.Errorf("ReloadConfig error %q names no unknown key", rerr.Error())
		}
		if got := w.Snapshot().Timeouts.TapWindowMs; got != 450 {
			t.Errorf("served window after rejected reload = %d, want last-good 450", got)
		}
		if lerr := w.LastError(); lerr == nil {
			t.Error("LastError after rejected reload = nil, want the rejection (D-32)")
		}
	})

	t.Run("no config file", func(t *testing.T) {
		svc := ctlsvc.NewSvc(ctlsvc.Deps{})
		reply, rerr := svc.ReloadConfig()
		if rerr == nil {
			t.Fatalf("ReloadConfig on the no-config daemon = nil error (reply %q), want the report", reply)
		}
		if !strings.Contains(rerr.Error(), "no config file") {
			t.Errorf("ReloadConfig no-config error %q missing the report", rerr.Error())
		}
	})
}

// TestSvc_CorrectNow pins the forced word correction (INST-02, the Q5 word
// semantics): the control method launches the actor's word pipeline — the
// RequireSurroundingText round arms on the sink — with NO tap anywhere,
// nothing destructive before the surrounding verdict, and the matching
// client push settles it exactly like the Double decision would.
func TestSvc_CorrectNow(t *testing.T) {
	a := session.NewActor(ctlTestWindow)
	sink := &ctlSink{}
	a.AttachEngine(sink)
	a.HandleCapabilities(engine.CapSurroundingText)
	for _, r := range ctlWordEN {
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r)})
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r), Release: true})
	}

	svc := ctlsvc.NewSvc(ctlsvc.Deps{Correct: a})
	reply, err := svc.CorrectNow()
	if err != nil {
		t.Fatalf("CorrectNow error = %v, want nil", err)
	}
	if reply == "" {
		t.Fatal("CorrectNow reply is empty, want the immediate acknowledgment")
	}
	if got := sink.requireCount(); got != 1 {
		t.Fatalf("RequireSurroundingText calls after forced correction = %d, want 1", got)
	}
	if got := sink.deleteCount(); got != 0 {
		t.Fatalf("two-phase violation: %d deletions before the surrounding verdict, want 0", got)
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Fatalf("two-phase violation: %d commits before the surrounding verdict, want 0", len(texts))
	}

	// The matching client push settles the armed round: the ladder fires
	// exactly as it does for the Double decision.
	a.HandleSurroundingText(ctlWordEN, runeLen(ctlWordEN), runeLen(ctlWordEN))
	if got := sink.deleteCount(); got != 1 {
		t.Fatalf("deletions after the matching push = %d, want 1", got)
	}
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != ctlWordRU {
		t.Fatalf("commits after the matching push = %q, want exactly one %q", texts, ctlWordRU)
	}
}

// runeLen is the cursor position at the end of s (the surrounding-text
// argument shape, self-computed).
func runeLen(s string) uint32 {
	return uint32(len([]rune(s)))
}

// TestSvc_RecoverShim pins the panic containment (T-03-06-02, the INTEG-05
// continuation): a panic below a control method surfaces as a *dbus.Error
// and the service still answers the next healthy call — the daemon lives.
func TestSvc_RecoverShim(t *testing.T) {
	st := &switchableStatus{}
	svc := ctlsvc.NewSvc(ctlsvc.Deps{Status: st})

	st.panicNext = true
	reply, err := svc.Status()
	if err == nil {
		t.Fatalf("panic below Status surfaced no error (reply %q) — the shim must convert it to a *dbus.Error", reply)
	}
	var derr dbus.Error
	if !errors.As(err, &derr) {
		t.Fatalf("Status error after panic = %T, want *dbus.Error", err)
	}

	st.panicNext = false
	reply, err = svc.Status()
	if err != nil {
		t.Fatalf("Status after the contained panic = %v, want nil (the daemon lives)", err)
	}
	if !strings.Contains(reply, "mode=en") {
		t.Errorf("Status reply after the contained panic = %q, missing mode=en", reply)
	}
}

// testBusConfig is the minimal permissive dbus-daemon configuration of the
// guard corpus: a private session bus on a temp socket with every send,
// receive and own allowed (the corpus is the only client).
const testBusConfig = `<!DOCTYPE busconfig PUBLIC "-//freedesktop//DTD D-Bus Bus Configuration 1.0//EN"
 "http://www.freedesktop.org/standards/dbus/1.0/busconfig.dtd">
<busconfig>
  <type>session</type>
  <listen>unix:tmpdir=</listen>
  <policy context="default">
    <allow own="*"/>
    <allow send_destination="*"/>
    <allow receive_type="method_call"/>
    <allow receive_type="method_return"/>
    <allow receive_type="error"/>
    <allow receive_type="signal"/>
  </policy>
</busconfig>
`

// startTestBus spawns a private dbus-daemon and points the process's
// session-bus address at it — the hermetic bus of the guard corpus (no
// live session is touched). The daemon dies with the test's cleanup
// context.
func startTestBus(t *testing.T) {
	t.Helper()

	bin, err := exec.LookPath("dbus-daemon")
	if err != nil {
		t.Skipf("dbus-daemon not found (%v) — the name-guard pin needs a private bus", err)
	}
	dir := t.TempDir()
	confPath := filepath.Join(dir, "bus.conf")
	// The listen address needs the actual temp dir appended after the
	// tmpdir= prefix (the config template leaves it empty).
	conf := strings.Replace(testBusConfig, "unix:tmpdir=", "unix:tmpdir="+dir, 1)
	if err := os.WriteFile(confPath, []byte(conf), configFilePerm); err != nil {
		t.Fatalf("write bus config: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, bin, "--config-file="+confPath, "--print-address=1", "--nofork")
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("bus stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start dbus-daemon: %v", err)
	}
	line, err := bufio.NewReader(out).ReadString('\n')
	if err != nil {
		t.Fatalf("read bus address: %v", err)
	}
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", strings.TrimSpace(line))
}

// waitCtlOwner polls the private bus until the control name has an owner —
// the first Run's export completed. A wait-for-condition poll, never a
// fixed sleep.
func waitCtlOwner(t *testing.T, timeout time.Duration) error {
	t.Helper()

	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("probe connect: %w", err)
	}
	defer func() { _ = conn.Close() }()

	deadline := time.Now().Add(timeout)
	for {
		var owner string
		call := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").
			CallWithContext(context.Background(), "org.freedesktop.DBus.GetNameOwner", 0, ctlsvc.BusName)
		if call.Err == nil {
			if err := call.Store(&owner); err == nil {
				return nil
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: %s never gained an owner", ctlsvc.BusName)
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// TestSvc_NameGuard pins the single-instance guard (V4/T-03-06-01) and the
// wire serving: on a private bus, the first Run takes org.djarvur.goswitch
// as primary owner and serves Status to a plain client; a SECOND Run on
// the same bus answers ErrNotPrimaryOwner — the sentinel the daemon's
// caller checks with errors.Is (the conn.go idiom).
func TestSvc_NameGuard(t *testing.T) {
	startTestBus(t)

	deps := ctlsvc.Deps{Status: fakeStatus{snap: session.Status{Mode: "en"}}}

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	run1 := make(chan error, 1)
	go func() { run1 <- ctlsvc.Run(ctx1, deps) }()

	if err := waitCtlOwner(t, 5*time.Second); err != nil {
		t.Fatalf("first Run never owned the name: %v", err)
	}

	// Wire serving: a plain client call answers the rendered status.
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = conn.Close() }()
	var reply string
	if err := conn.Object(ctlsvc.BusName, ctlsvc.ObjectPath).
		CallWithContext(context.Background(), ctlsvc.BusName+".Status", 0).Store(&reply); err != nil {
		t.Fatalf("client Status call: %v", err)
	}
	if !strings.Contains(reply, "mode=en") {
		t.Errorf("wire Status reply = %q, missing mode=en", reply)
	}

	// The guard: a second service on the same bus must NOT get the name.
	if err := ctlsvc.Run(context.Background(), deps); !errors.Is(err, ctlsvc.ErrNotPrimaryOwner) {
		t.Fatalf("second Run error = %v, want ErrNotPrimaryOwner", err)
	}

	cancel1()
	select {
	case err := <-run1:
		if err != nil {
			t.Errorf("first Run exit = %v, want nil on ctx cancel", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("first Run never returned after ctx cancel")
	}
}

// ctlConfigYAML is a COMPLETE config document with the given tap window
// (the 03-02 strict-parse rule: no defaults overlay).
func ctlConfigYAML(windowMs int) string {
	return fmt.Sprintf(`hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: %d
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
`, windowMs)
}

// ctlBrokenYAML carries one unknown key — the strict decoder names it in
// the rejection (D-33), which the reload reply and the status string must
// both surface (D-32).
const ctlBrokenYAML = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_mss: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
`
