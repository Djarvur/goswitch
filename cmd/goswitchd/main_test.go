package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Djarvur/goswitch/engine"
	"github.com/Djarvur/goswitch/internal/config"
	"github.com/Djarvur/goswitch/internal/hotkey"
	"github.com/Djarvur/goswitch/internal/session"
)

// restoreBuildVars puts the package build coordinates back after a subtest
// mutated them — the corpus never leaks state between tests.
func restoreBuildVars(t *testing.T) {
	t.Helper()
	prev := [3]string{version, commit, date}
	t.Cleanup(func() { version, commit, date = prev[0], prev[1], prev[2] })
}

// TestVersionString pins the D-37 version line: a dev build (unstamped
// commit/date) prints exactly "goswitchd dev"; a release build with the
// goreleaser-stamped coordinates prints name, version, commit and date
// without hanging punctuation.
func TestVersionString(t *testing.T) {
	t.Run("dev fallback: no commit, no date", func(t *testing.T) {
		restoreBuildVars(t)
		version, commit, date = "dev", "", ""

		if got := versionString(); got != "goswitchd dev" {
			t.Errorf("versionString() = %q, want %q", got, "goswitchd dev")
		}
	})

	t.Run("stamped release build", func(t *testing.T) {
		restoreBuildVars(t)
		version, commit, date = "v1.0.0", "abc123", "2026-01-01"

		want := "goswitchd v1.0.0 abc123 2026-01-01"
		if got := versionString(); got != want {
			t.Errorf("versionString() = %q, want %q", got, want)
		}
	})

	t.Run("stamped commit without date", func(t *testing.T) {
		restoreBuildVars(t)
		version, commit, date = "v0.9.0", "deadbee", ""

		want := "goswitchd v0.9.0 deadbee"
		if got := versionString(); got != want {
			t.Errorf("versionString() = %q, want %q", got, want)
		}
	})
}

// TestVersionFlag pins the -version flag behavior: runVersion prints the
// version line (and nothing else) — main calls it and exits 0 before run()
// is ever reached, so the daemon never starts for a version query.
func TestVersionFlag(t *testing.T) {
	restoreBuildVars(t)
	version, commit, date = "dev", "", ""

	var buf bytes.Buffer
	runVersion(&buf)

	if got := buf.String(); got != "goswitchd dev\n" {
		t.Errorf("runVersion output = %q, want %q", got, "goswitchd dev\n")
	}
}

// TestEngineConfigWiresSwitcher pins the D-52 wiring: engineConfig hands the
// BindSwitcher closure to the actor BEFORE engine.Run — a flip gesture after
// the wiring reaches the seam, and the status snapshot's engine follows the
// mode the seam flipped into.
func TestEngineConfigWiresSwitcher(t *testing.T) {
	actor := session.NewActor(time.Second)
	cfg := engineConfig(actor)
	if cfg.BindSwitcher == nil {
		t.Fatal("engineConfig leaves BindSwitcher nil — the D-52 flip would never reach the bus")
	}

	switches := make(chan string, 4)
	cfg.BindSwitcher(func(_ context.Context, name string) error {
		switches <- name

		return nil
	})

	actor.HandleKey(engine.EngineEvent{Keyval: hotkey.KeyvalShiftR})
	actor.HandleKey(engine.EngineEvent{Keyval: hotkey.KeyvalShiftR, Release: true})
	actor.ExpiryAt(2 * time.Second) // past the one-second window — the Single decision

	select {
	case name := <-switches:
		if name != engine.NameRU {
			t.Errorf("wired switcher received %q, want %s", name, engine.NameRU)
		}
	default:
		t.Fatal("the flip never reached the wired switcher — the closure does not feed the actor")
	}
	if st := actor.StatusSnapshot(); st.Engine != engine.NameRU {
		t.Errorf("snapshot engine = %q, want %s", st.Engine, engine.NameRU)
	}
}

// TestMain_WiringSyncAndReader pins the 05-04 wiring: OnGlobalEngine is the
// actor's SyncEngine (every observed engine name corrects the daemon's
// mode — criterion 3), BindGlobalEngine installs a working factual-engine
// reader, and PostRegister hands that reader to activate.IfOwned — the
// reactivation consults the reader BEFORE any gsettings read, so a foreign
// factual name skips without a single subprocess (the hermetic proof of the
// handoff).
func TestMain_WiringSyncAndReader(t *testing.T) {
	actor := session.NewActor(time.Second)
	cfg := engineConfig(actor)

	if cfg.OnGlobalEngine == nil {
		t.Fatal("engineConfig leaves OnGlobalEngine nil — the sync listener is dead (criterion 3)")
	}
	cfg.OnGlobalEngine(engine.NameRU)
	if st := actor.StatusSnapshot(); st.Mode != "ru" {
		t.Errorf("mode after OnGlobalEngine(%s) = %q, want ru — the sync must reach the actor", engine.NameRU, st.Mode)
	}

	if cfg.BindGlobalEngine == nil {
		t.Fatal("engineConfig leaves BindGlobalEngine nil — IfOwned could never consult the factual engine (Pitfall 3)")
	}

	// PostRegister hands the bound reader to IfOwned: a FOREIGN factual
	// name makes IfOwned skip without a single subprocess — the consulted
	// channel proves the handoff (IfOwned consults the reader before any
	// gsettings read), the fast return proves the skip. The reader itself
	// is engine-package surface (its own corpus); the reader-to-IfOwned
	// semantics live in the activate corpus — this pin is the wiring.
	consulted := make(chan struct{}, 1)
	cfg.BindGlobalEngine(func(context.Context) (string, error) {
		consulted <- struct{}{}

		return "mozc", nil // foreign factual name: IfOwned must skip without subprocesses
	})
	done := make(chan struct{})
	go func() {
		cfg.PostRegister(context.Background(), 0)
		close(done)
	}()
	select {
	case <-consulted:
	case <-time.After(2 * time.Second):
		t.Fatal("IfOwned never consulted the bound reader — the wiring handoff is broken")
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("PostRegister never returned — the foreign-engine skip is stuck")
	}
}

// adoptDefaultDocYAML is a complete valid document planted at the adopted
// default path with a distinguishable tap window (555).
const adoptDefaultDocYAML = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
  mode_switch_chord: super+space
timeouts:
  tap_window_ms: 555
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
autocorrect:
  enabled: false
  apps_blocklist: []
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
`

// adoptExplicitDocYAML is the explicit -config competitor — same shape, a
// distinguishable tap window (777).
const adoptExplicitDocYAML = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
  mode_switch_chord: super+space
timeouts:
  tap_window_ms: 777
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
autocorrect:
  enabled: false
  apps_blocklist: []
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
`

// writeAdoptDoc plants a document at path inside the temp user config dir.
func writeAdoptDoc(t *testing.T, path, corpus string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(corpus), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

// TestLoadConfigAdoptExisting pins the adopt branch for an EXISTING
// default-path file: without -config the daemon loads it — the same
// contract as an explicit -config (values, fromFile, the factual path).
func TestLoadConfigAdoptExisting(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path := config.DefaultPath()
	writeAdoptDoc(t, path, adoptDefaultDocYAML)

	lc, err := loadConfig("")
	if err != nil {
		t.Fatalf("loadConfig with an existing default file: %v", err)
	}
	if !lc.fromFile {
		t.Error("loadConfig reports defaults, want a file load — the adopt contract")
	}
	if lc.path != path {
		t.Errorf("loaded path = %q, want the adopted default path %q", lc.path, path)
	}
	if lc.cfg.Timeouts.TapWindowMs != 555 {
		t.Errorf("tap_window_ms = %d, want the file's 555 — the file content must apply", lc.cfg.Timeouts.TapWindowMs)
	}
}

// TestLoadConfigAdoptAbsent pins the 04-02 contract under adopt: an absent
// default file stays the green defaults start — the daemon watches the
// default path but creates NOTHING at start (generation is a user action,
// never startup).
func TestLoadConfigAdoptAbsent(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path := config.DefaultPath()

	lc, err := loadConfig("")
	if err != nil {
		t.Fatalf("loadConfig with an absent default file: %v", err)
	}
	if lc.fromFile {
		t.Error("loadConfig reports a file load, want the defaults start")
	}
	if lc.path != path {
		t.Errorf("watch path = %q, want the default path %q — the first toggle's file must be watchable", lc.path, path)
	}
	if lc.cfg.Timeouts.TapWindowMs != config.Defaults().Timeouts.TapWindowMs {
		t.Errorf("tap_window_ms = %d, want the built-in default", lc.cfg.Timeouts.TapWindowMs)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("stat after loadConfig = %v, want ErrNotExist — the start must not create the file", err)
	}
}

// TestLoadConfigAdoptBroken pins the loud refusal (T-07-03-03): a broken
// default-path file refuses the start with the Load error — silently
// serving defaults would hide an on-disk/behavior divergence the user
// cannot see any other way.
func TestLoadConfigAdoptBroken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path := config.DefaultPath()
	writeAdoptDoc(t, path, "{ not yaml")

	_, err := loadConfig("")
	if err == nil {
		t.Fatal("loadConfig on a broken default file returned nil — the divergence went unnoticed")
	}
	if !strings.Contains(err.Error(), "decode config") {
		t.Errorf("error %q does not carry the Load error — the refusal must name the cause", err)
	}
}

// TestLoadConfigExplicitPriority pins the explicit -config contract
// byte-as-today: with both files present the explicit one wins, and its
// load is mandatory (fromFile).
func TestLoadConfigExplicitPriority(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	writeAdoptDoc(t, config.DefaultPath(), adoptDefaultDocYAML)
	explicit := filepath.Join(tmp, "explicit.yaml")
	writeAdoptDoc(t, explicit, adoptExplicitDocYAML)

	lc, err := loadConfig(explicit)
	if err != nil {
		t.Fatalf("loadConfig with an explicit -config: %v", err)
	}
	if !lc.fromFile {
		t.Error("loadConfig reports defaults for an explicit -config, want a mandatory file load")
	}
	if lc.path != explicit {
		t.Errorf("loaded path = %q, want the explicit path %q", lc.path, explicit)
	}
	if lc.cfg.Timeouts.TapWindowMs != 777 {
		t.Errorf("tap_window_ms = %d, want the explicit file's 777 — explicit wins", lc.cfg.Timeouts.TapWindowMs)
	}
}

// TestAdoptWatcherServesDefaultsOnAbsentFile pins the adopt+watch
// construction: the watcher starts on the ABSENT default path (serving
// the defaults through the adopt loader) so the file the first toggle
// creates is picked up by the same watch — the non-nil watcher in the
// absent-file scenario is the acceptance evidence.
func TestAdoptWatcherServesDefaultsOnAbsentFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	lc, err := loadConfig("")
	if err != nil {
		t.Fatalf("loadConfig with an absent default file: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // the watcher loop exits on ctx — goroutine hygiene

	w, err := startConfigWatcher(ctx, lc, "")
	if err != nil {
		t.Fatalf("startConfigWatcher on the absent default path: %v", err)
	}
	if w == nil {
		t.Fatal("watcher is nil — the adopt+watch contract requires an unconditional watcher")
	}
	if w.ConfigPath() != lc.path {
		t.Errorf("watch path = %q, want the adopted default path %q", w.ConfigPath(), lc.path)
	}
	if got := w.Snapshot().Timeouts.TapWindowMs; got != config.Defaults().Timeouts.TapWindowMs {
		t.Errorf("watcher snapshot tap_window_ms = %d, want the served defaults", got)
	}
}

// syncBuffer is the guarded log buffer of the wiring corpus (the session
// corpus's captureLogs discipline, local copy — dupl is relaxed in tests).
type syncBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

// Write appends to the buffer under the guard.
func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

// String snapshots the buffer under the guard.
func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

// captureLogs redirects the process default logger into a guarded buffer.
func captureLogs(t *testing.T) *syncBuffer {
	t.Helper()
	buf := &syncBuffer{}
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, nil)))
	t.Cleanup(func() { slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil))) })

	return buf
}

// The menu-wiring corpus of plan 07-05: the persisted-toggle compositions
// (the pinned snapshot → write → push → reload order, both toggles) and the
// settings launcher (the no-pipes xdg-open shape, T-07-05-01/02).

// toggleTestPath is the factual-path stand-in of the composition corpus.
const toggleTestPath = "/tmp/goswitch-toggle.yaml"

// errWriteInjected is the toggle-writer failure of the composition corpus.
var errWriteInjected = errors.New("write exploded")

// fakeCfgSource is the config-source double of the composition corpus: a
// value document standing in for the watcher's Snapshot contract.
type fakeCfgSource struct{ cfg config.Config }

// Snapshot returns the document.
func (s fakeCfgSource) Snapshot() config.Config { return s.cfg }

// fakeProc is the started-process double of the launcher seam: the REAL
// exec.Cmd shape is recorded for the argv/pipes pins, Start/Wait are
// counted, Start can fail.
type fakeProc struct {
	mu       sync.Mutex
	cmd      *exec.Cmd
	startErr error
	started  int
	waited   int
}

// Start counts and answers the canned verdict — nothing spawns.
func (f *fakeProc) Start() error {
	f.mu.Lock()
	f.started++
	f.mu.Unlock()

	return f.startErr
}

// Wait counts the reap.
func (f *fakeProc) Wait() error {
	f.mu.Lock()
	f.waited++
	f.mu.Unlock()

	return nil
}

// counts snapshots the Start/Wait counters under the guard — the reaper
// goroutine races the test's polling.
func (f *fakeProc) counts() (started, waited int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.started, f.waited
}

// installFakeProc swaps the launcher seam for the counting double and
// restores it on cleanup.
func installFakeProc(t *testing.T, startErr error) *fakeProc {
	t.Helper()

	fp := &fakeProc{startErr: startErr}
	prev := newEditorProc
	newEditorProc = func(bin, path string) procStarter {
		// The REAL shape for the argv/pipes pins — nothing spawns.
		fp.cmd = exec.CommandContext(context.Background(), bin, path)

		return fp
	}
	t.Cleanup(func() { newEditorProc = prev })

	return fp
}

// waitFor polls until cond turns true — the reap is asynchronous.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("condition never became true")
		}
		time.Sleep(time.Millisecond)
	}
}

// appliedReader folds one toggle's truth into the actor's snapshot: the
// config document carries the applied value, one ExpiryAt drives the fold.
func appliedReader(t *testing.T, section string, applied bool) *session.Actor {
	t.Helper()

	actor := session.NewActor(time.Hour)
	cfg := config.Defaults()
	switch section {
	case "autocorrect":
		cfg.Autocorrect.Enabled = applied
	case "sound":
		if applied {
			cfg.Sound = config.Sound{} // the absent section reads enabled
		} else {
			off := false
			cfg.Sound = config.Sound{Enabled: &off}
		}
	}
	actor.AttachConfig(fakeCfgSource{cfg: cfg})
	actor.ExpiryAt(2 * time.Hour) // one fold: the applied truth lands in the snapshot

	return actor
}

// TestConfigToggleComposition pins BOTH persisted toggles' composition
// (plan 07-05): the click reads the APPLIED value from the actor's status
// snapshot, inverts it, writes it through the config writer, pushes the
// menu, then applies synchronously through the reload — in that order, for
// autocorrect and for sound alike. A failed write stops everything.
func TestConfigToggleComposition(t *testing.T) {
	t.Run("autocorrect-order-and-inversion", func(t *testing.T) {
		assertToggleFlip(t, appliedReader(t, "autocorrect", true),
			func(a *session.Actor) bool { return a.StatusSnapshot().AutoCorrectEnabled }, false)
	})

	t.Run("sound-order-and-inversion", func(t *testing.T) {
		assertToggleFlip(t, appliedReader(t, "sound", false),
			func(a *session.Actor) bool { return a.StatusSnapshot().SoundEnabled }, true)
	})

	t.Run("write-failure-stops-everything", func(t *testing.T) {
		buf := captureLogs(t)
		pushed, reloaded := false, false
		tg := &configToggle{
			path:   toggleTestPath,
			read:   func() bool { return false },
			write:  func(string, bool) error { return errWriteInjected },
			push:   func(bool) { pushed = true },
			reload: func() { reloaded = true },
		}

		tg.flip()

		if pushed || reloaded {
			t.Errorf("a failed write continued (push %t, reload %t) — the state did not change", pushed, reloaded)
		}
		if !strings.Contains(buf.String(), "config toggle write failed") {
			t.Errorf("the write failure is not WARNed; log:\n%s", buf.String())
		}
	})
}

// assertToggleFlip runs one composition order pin: snapshot →
// write(inverted) → push → reload, with the write value asserted against
// the inversion of the applied truth.
func assertToggleFlip(t *testing.T, actor *session.Actor, read func(*session.Actor) bool, wantWrite bool) {
	t.Helper()

	var order []string
	tg := &configToggle{
		path: toggleTestPath,
		read: func() bool {
			order = append(order, "snapshot")

			return read(actor)
		},
		write: func(path string, on bool) error {
			if path != toggleTestPath {
				t.Errorf("write path = %q, want the factual config path", path)
			}
			if on != wantWrite {
				t.Errorf("write value = %t, want the inverted %t", on, wantWrite)
			}
			order = append(order, "write")

			return nil
		},
		push:   func(on bool) { order = append(order, "push") },
		reload: func() { order = append(order, "reload") },
	}

	tg.flip()

	want := []string{"snapshot", "write", "push", "reload"}
	if !slices.Equal(order, want) {
		t.Errorf("composition order = %q, want exactly %q", order, want)
	}
}

// TestToggleNilReloadKeepsWrite pins the degenerate no-source guard (plan
// 07-05): nil reload is impossible after the 07-03 adopt (the watcher is
// unconditional), but the defensive path keeps the write AND the push — the
// file is the truth, the status catches up through the next fold — and
// WARNs the deferred apply.
func TestToggleNilReloadKeepsWrite(t *testing.T) {
	buf := captureLogs(t)
	wrote, pushed := false, false
	tg := &configToggle{
		path: toggleTestPath,
		read: func() bool { return true },
		write: func(string, bool) error {
			wrote = true

			return nil
		},
		push:   func(bool) { pushed = true },
		reload: nil,
	}

	tg.flip()

	if !wrote || !pushed {
		t.Errorf("nil reload dropped the write (wrote %t) or the push (pushed %t) — the file is the truth",
			wrote, pushed)
	}
	if !strings.Contains(buf.String(), `"level":"WARN"`) {
		t.Errorf("the deferred apply is not WARNed; log:\n%s", buf.String())
	}
}

// TestOpenConfigEditorLauncher pins the settings launcher's happy path
// (plan 07-05, research Q7): the document is ensured first, then xdg-open
// starts on the pinned path as its ONLY argument with NO piped descriptors
// (the fork-shaped wl-copy precedent) and is reaped by a Wait goroutine.
func TestOpenConfigEditorLauncher(t *testing.T) {
	var ensured []string
	fp := installFakeProc(t, nil)
	editor := &configEditor{
		ensure: func(path string) error {
			ensured = append(ensured, path)

			return nil
		},
	}

	editor.open("/tmp/goswitch-config.yaml")

	if !slices.Equal(ensured, []string{"/tmp/goswitch-config.yaml"}) {
		t.Errorf("ensure paths = %q, want exactly the pinned config path", ensured)
	}
	if fp.cmd == nil {
		t.Fatal("no process was built — the launch never happened")
	}
	if !slices.Equal(fp.cmd.Args, []string{"xdg-open", "/tmp/goswitch-config.yaml"}) {
		t.Errorf("argv = %q, want exactly [xdg-open, path] — never a shell, never user data (T-07-05-01)", fp.cmd.Args)
	}
	if fp.cmd.Stdout != nil || fp.cmd.Stderr != nil {
		t.Error("the editor got piped descriptors — the fork-shaped grandchild would deadlock a Run (T-07-05-02)")
	}
	started, _ := fp.counts()
	if started != 1 {
		t.Errorf("Start calls = %d, want exactly 1", started)
	}
	waitFor(t, func() bool {
		_, waited := fp.counts()

		return waited == 1
	})
}

// TestOpenConfigEditorStartErrorWarnsOnce pins the one-WARN-per-lifetime
// degradation of a failed launch: the menu keeps working, the second
// failure is silent (the emitWarned precedent).
func TestOpenConfigEditorStartErrorWarnsOnce(t *testing.T) {
	buf := captureLogs(t)
	fp := installFakeProc(t, errWriteInjected)
	editor := &configEditor{ensure: func(string) error { return nil }}

	editor.open("/tmp/goswitch-config.yaml")
	editor.open("/tmp/goswitch-config.yaml")

	waitFor(t, func() bool {
		started, _ := fp.counts()

		return started == 2
	})
	if got := strings.Count(buf.String(), "config editor start failed"); got != 1 {
		t.Errorf("start failures warned %d times, want exactly one per editor lifetime; log:\n%s", got, buf.String())
	}
}

// TestOpenConfigEditorEnsureFailureSkipsLaunch pins the Q7 guard: a failed
// ensure never launches — an editor on an empty buffer would
// strict-decode-refuse on save.
func TestOpenConfigEditorEnsureFailureSkipsLaunch(t *testing.T) {
	buf := captureLogs(t)
	fp := installFakeProc(t, nil)
	editor := &configEditor{ensure: func(string) error { return errWriteInjected }}

	editor.open("/tmp/goswitch-config.yaml")

	started, _ := fp.counts()
	if fp.cmd != nil || started != 0 {
		t.Error("a failed ensure still launched the editor — the editor would open an empty buffer (Q7)")
	}
	if !strings.Contains(buf.String(), `"level":"WARN"`) {
		t.Errorf("the ensure failure is not WARNed; log:\n%s", buf.String())
	}
}

// The sound-wiring corpus of plan 07-08: the daemon's sound surface — the
// startup config feeds the gate (default ON), the installed sink hears the
// flips, and the applied document's switch gates through the snapshot fold
// without a restart.

// fakeSoundSink is the sound-seam double of the wiring corpus: both tones
// counted under a mutex.
type fakeSoundSink struct {
	mu    sync.Mutex
	flips int
	fires int
}

func (f *fakeSoundSink) Flip() {
	f.mu.Lock()
	f.flips++
	f.mu.Unlock()
}

func (f *fakeSoundSink) AutoCorrect() {
	f.mu.Lock()
	f.fires++
	f.mu.Unlock()
}

func (f *fakeSoundSink) flipCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.flips
}

// foldSource is the snapshot fake of the watcher's Snapshot contract — the
// reload corpus mutates the document between gestures.
type foldSource struct {
	mu  sync.Mutex
	cfg config.Config
}

func (s *foldSource) Snapshot() config.Config {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.cfg
}

func (s *foldSource) set(cfg config.Config) {
	s.mu.Lock()
	s.cfg = cfg
	s.mu.Unlock()
}

// TestNewActor_SoundGatingFromStartupConfig pins the startup feed (plan
// 07-08): the daemon's Options carry the document's effective sound switch
// — the defaults read ON (the owner's verdict), an explicit enabled: false
// mutes the daemon from the first gesture.
func TestNewActor_SoundGatingFromStartupConfig(t *testing.T) {
	t.Run("defaults: sound ON", func(t *testing.T) {
		actor := newActor(config.Defaults(), nil)
		sink := &fakeSoundSink{}
		actor.SetSoundSink(sink)

		actor.ToggleMode()

		if got := sink.flipCount(); got != 1 {
			t.Errorf("flip tones under the defaults = %d, want 1 — sound is default ON", got)
		}
	})

	t.Run("enabled: false mutes", func(t *testing.T) {
		cfg := config.Defaults()
		off := false
		cfg.Sound.Enabled = &off
		actor := newActor(cfg, nil)
		sink := &fakeSoundSink{}
		actor.SetSoundSink(sink)

		actor.ToggleMode()

		if got := sink.flipCount(); got != 0 {
			t.Errorf("flip tones with sound.enabled=false = %d, want 0", got)
		}
	})
}

// TestNewActor_SoundHotReloadGating pins the D-32 contour (plan 07-08):
// the applied document's switch gates through the snapshot fold — muting
// the document silences the next flip WITHOUT a restart, a reload back to
// ON sounds again.
func TestNewActor_SoundHotReloadGating(t *testing.T) {
	cfg := config.Defaults()
	off := false
	cfg.Sound.Enabled = &off
	src := &foldSource{cfg: cfg}
	actor := newActor(cfg, nil)
	actor.AttachConfig(src)
	sink := &fakeSoundSink{}
	actor.SetSoundSink(sink)

	actor.ExpiryAt(time.Second) // one fold of the muted document
	actor.ToggleMode()
	if got := sink.flipCount(); got != 0 {
		t.Fatalf("flip tones under the muted document = %d, want 0 — the fold gates without a restart", got)
	}

	on := true
	cfg.Sound.Enabled = &on
	src.set(cfg)
	actor.ExpiryAt(2 * time.Second) // the fold of the re-enabled document
	actor.ToggleMode()
	if got := sink.flipCount(); got != 1 {
		t.Errorf("flip tones after the re-enabled reload = %d, want 1 — hot reload re-arms the tones", got)
	}
}

// TestNewActor_StartupFoldInForce pins the startup fold (plan 07-06, the
// restart-survival pin of the menu-v2 case): a daemon built on a LOADED
// document reports the document truth from the FIRST status read — the
// fold the first key event would run runs at startup, never a stale off
// (the live finding: the restart step saw autocorrect_enabled=false and
// sound_enabled=false against a true/true document on disk).
func TestNewActor_StartupFoldInForce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	const doc = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: false
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
autocorrect:
  enabled: true
  apps_blocklist: []
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
sound:
  enabled: true
`
	if err := os.WriteFile(path, []byte(doc), 0o600); err != nil {
		t.Fatalf("write config document: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config document: %v", err)
	}
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	w, err := config.NewWatcher(ctx, path)
	if err != nil {
		t.Fatalf("create watcher: %v", err)
	}

	actor := newActor(*cfg, w)
	st := actor.StatusSnapshot()
	if !st.AutoCorrectEnabled {
		t.Errorf("autocorrect option at startup = false, want the document truth true — no key-event wait")
	}
	if !st.SoundEnabled {
		t.Errorf("sound option at startup = false, want the document truth true")
	}
}
