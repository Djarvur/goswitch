package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
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
		t.Errorf("tap_window_ms = %d, want the explicit file's 777 — the explicit -config is priority", lc.cfg.Timeouts.TapWindowMs)
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
