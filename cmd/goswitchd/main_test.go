package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/Djarvur/goswitch/engine"
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
