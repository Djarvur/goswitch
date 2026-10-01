package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"syscall"
	"time"
)

// Resilience-case timing knobs.
const (
	reRegisterWait       = 20 * time.Second // ibus restart: new daemon + our backoff + re-registration
	respawnWait          = 15 * time.Second // kill9: subprocess respawn to a fresh registration
	activationAttempts   = 3                // engine activation retries right after a bus restart
	activationRetryDelay = 1 * time.Second
	kill9Text            = "test" // what kill9-survive types with the daemon dead
)

// The two-source restart semantics (LIVE FINDING, the first 05-05 runs,
// deterministic ×2): gnome-shell does NOT follow daemon SetGlobalEngine
// flips in its own current-source state (the spec-delta lag taken to the
// restart limit), so after `ibus restart` the SHELL re-activates its own
// current source (index 0, goswitch-en) before the daemon's backoff even
// lands — and the daemon's IfOwned then correctly reactivates the FACTUAL
// engine (en) and the sync listener corrects the mode to follow. A
// daemon-flipped ru is therefore a restart-lost state BY MECHANISM, not a
// daemon bug: the factual=ru branch of IfOwned is pinned hermetically
// (05-04 corpus); the shell-gesture flips (Super+Space, which the shell
// tracks) do survive. The case pins the MECHANISM truth — reactivation of
// the factual engine + the mode following it — and never re-asserts ru
// against the shell (that would be the forbidden war of mechanisms).

// runIbusRestart proves INTEG-04 live: `ibus restart` tears down the private
// bus (socket path included); the daemon's reconnect loop must re-discover
// the NEW socket and re-register (ibus#2910), after which keys flow again.
// The 05-05 two-source extension pins the reactivation MECHANISM: a
// goswitch engine comes back, it IS the factual engine the bus reports,
// and the daemon's mode follows it — the IfOwned factual-priority +
// sync-follow pair of 05-04 proven live.
func runIbusRestart(ctx context.Context, s *stand) error {
	if err := s.activateGoswitchRetry(ctx); err != nil {
		return err
	}
	if err := s.pinRestartEngineRU(ctx); err != nil {
		return err
	}
	keysBefore := s.countSub(`"msg":"key"`)
	reactivationsBefore := s.countSub(engineReactivatedMark)

	if _, err := runCmd(ctx, "ibus", "restart"); err != nil {
		return fmt.Errorf("ibus restart: %w", err)
	}
	if err := s.waitForLog(ctx, "re-registered", reRegisterWait); err != nil {
		return fmt.Errorf("INTEG-04 re-registration: %w (the reconnect loop never reached the new socket)", err)
	}

	// The reactivation mechanism assertion (the LIVE FINDING semantics in
	// the file header comment): what came back, and did the mode follow.
	if err := s.verifyRestartReactivation(ctx, reactivationsBefore); err != nil {
		return err
	}

	// GNOME re-activates the plain xkb sources after a bus restart —
	// goswitch is not in gsettings sources, so re-activate explicitly,
	// then type on a fresh surface.
	if err := s.activateGoswitchRetry(ctx); err != nil {
		return err
	}
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()

	if err := s.injectText(ctx, "ghbdtn"); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"key"`, keysBefore+minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("INTEG-04 post-restart key flow: %w", err)
	}

	return nil
}

// verifyRestartReactivation proves the two-source reactivation mechanism
// after the restart: a NEW engine-reactivation record within the
// re-registration budget (IfOwned runs at PostRegister, so the record may
// already be in the log when the re-registration line is seen), naming a
// goswitch engine that agrees with the bus readback (the FACTUAL engine —
// never a dead-key guess), with the mode following it (the sync path
// closed the loop).
func (s *stand) verifyRestartReactivation(ctx context.Context, before int) error {
	if err := s.waitForNew(ctx, engineReactivatedMark, before+1, reRegisterWait); err != nil {
		return fmt.Errorf("INTEG-04 two-source reactivation: %w (no owned engine came back)", err)
	}
	reactivated, err := s.lastReactivatedEngine()
	if err != nil {
		return err
	}
	if !isGoswitchEngine(reactivated) {
		return fmt.Errorf("INTEG-04 two-source reactivation: reactivated %q — not a goswitch engine", reactivated)
	}
	readback, err := runCmd(ctx, "ibus", "engine")
	if err != nil {
		return fmt.Errorf("INTEG-04 factual readback: %w", err)
	}
	if readback != reactivated {
		return fmt.Errorf("INTEG-04 two-source reactivation: reactivated %q but the bus reports %q —"+
			" the reactivation is not the factual engine", reactivated, readback)
	}
	last, lerr := s.lastModeMark()
	if lerr != nil {
		return lerr
	}
	wantMode := flipTargetEN
	if reactivated == spikeEngineRU {
		wantMode = flipTargetRU
	}
	if last != wantMode {
		return fmt.Errorf("INTEG-04 two-source reactivation: reactivated %q but the mode reads %q, want %q —"+
			" the mode did not follow the factual engine", reactivated, last, wantMode)
	}

	return nil
}

// pinRestartEngineRU makes goswitch-ru the active engine before the bus
// restart — the mechanism assertion needs a NON-DEFAULT engine active
// across it, so the case can observe what the restart does to a
// daemon-flipped state (the shell resets it; the daemon follows the
// factual engine). The pin rides the busFlipRound discipline (one Shift_R
// tap in the stand's own focused entry, the journal pair as proof, the
// readback as the factual confirmation) — never an assumed boot default,
// never a Super+Space injection (Pitfall 8). The goswitch-owned desktop
// is a case precondition and is checked, not assumed — of any sources
// count: with one source the flip mechanism still works (SetGlobalEngine
// targets any registered engine).
func (s *stand) pinRestartEngineRU(ctx context.Context) error {
	if err := s.requireGoswitchDesktop(ctx); err != nil {
		return err
	}
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("ibus-restart needs the zenity entry surface (locked-session fallback engaged?)")
	}

	return s.busFlipRound(ctx, flipTargetRU)
}

// runKill9Survive proves INTEG-05 live: SIGKILL is uncatchable by design
// (Pitfall 1), so the case asserts the DESKTOP side — typing through a
// plain engine still reaches the focused entry while the daemon is dead —
// and then respawns the daemon and waits for its fresh registration.
//
// The desktop-input oracle depends on the surface: the shell entry exposes
// a char count through the witness (AT-SPI readback, the Phase-2 bridge),
// the zenity entry prints its text on close (Enter) — length-checked,
// because the active XKB group maps the same key to a different letter
// (live finding: typing on this desktop yields привет for ghbdtn).
func runKill9Survive(ctx context.Context, s *stand) error {
	if err := s.activateGoswitchRetry(ctx); err != nil {
		return err
	}
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	charsBefore := -1
	if kind == surfaceShell {
		if charsBefore, err = s.witnessChars(ctx); err != nil {
			_ = s.closeEntrySurface(ctx, kind)

			return err
		}
	}
	registrations := s.countSub(componentRegisteredMark)

	if err := s.killAndType(ctx); err != nil {
		_ = s.closeEntrySurface(ctx, kind)

		return err
	}
	// Respawn (no systemd unit in the skeleton — the runner stands in for
	// Restart=on-failure) and wait for a fresh registration in the same
	// log; the input-alive oracle reads the entry only afterwards, since
	// the dead daemon's log cannot witness its own death.
	if err := s.respawnAndWait(ctx, registrations); err != nil {
		_ = s.closeEntrySurface(ctx, kind)

		return err
	}

	return s.kill9Oracle(ctx, kind, charsBefore)
}

// killAndType kills the daemon, moves the session to a plain engine and
// types with the daemon dead — the phase whose observable is purely
// desktop-side. The plain-engine switch is explicit (the plan says gsettings
// sources/current; live finding: SetGlobalEngine is the path GNOME 46
// honors) — ibus-daemon already unset the global engine on the component's
// death, and this is what keeps the desktop usable.
func (s *stand) killAndType(ctx context.Context) error {
	if err := s.killDaemon9(); err != nil {
		return err
	}
	if _, err := runCmd(ctx, "ibus", "engine", "xkb:us::eng"); err != nil {
		return fmt.Errorf("switch to plain xkb after daemon death: %w", err)
	}

	return s.injectText(ctx, kill9Text)
}

// respawnAndWait restarts the daemon subprocess and waits for its fresh
// registration in the same log (one more occurrence than before).
func (s *stand) respawnAndWait(ctx context.Context, registrations int) error {
	if err := s.startDaemon(); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, componentRegisteredMark, registrations+1, respawnWait); err != nil {
		return fmt.Errorf("INTEG-05 respawn registration: %w", err)
	}

	return nil
}

// kill9Oracle asserts the desktop input stayed alive while the daemon was
// dead, through whichever oracle the surface offers, and closes it.
func (s *stand) kill9Oracle(ctx context.Context, kind surfaceKind, charsBefore int) error {
	switch kind {
	case surfaceShell:
		after, err := s.witnessChars(ctx)
		if err != nil {
			_ = s.pressKey(ctx, "Escape")

			return err
		}
		if after <= charsBefore {
			_ = s.pressKey(ctx, "Escape")

			return fmt.Errorf("INTEG-05 desktop input dead: shell entry chars %d -> %d after kill -9",
				charsBefore, after)
		}

		return s.pressKey(ctx, "Escape")
	case surfaceZenity:
		out, err := s.closeZenity(ctx)
		if err != nil {
			return err
		}
		if got := len([]rune(out)); got != len(kill9Text) {
			return fmt.Errorf("INTEG-05 desktop input dead: zenity readback %q (%d chars, want %d)",
				out, got, len(kill9Text))
		}

		return nil
	case surfaceChromium:
		// Driver-managed surface (surface.go): resilience cases stay on the
		// entry surfaces.
		return nil
	case surfaceGnomeTextEditor:
		// Driver-managed surface (surface.go): same as chromium — resilience
		// cases stay on the entry surfaces.
		return nil
	case surfaceGedit:
		// Driver-managed surface (surface.go): same as chromium (04-05).
		return nil
	case surfaceChromiumX11:
		// Driver-managed surface (surface.go): same as chromium (04-05).
		return nil
	}

	return nil
}

// killDaemon9 kills the daemon with SIGKILL — the uncatchable death — and
// reaps the subprocess so the respawn starts from a clean slot.
func (s *stand) killDaemon9() error {
	if s.daemon == nil || s.daemon.Process == nil {
		return errors.New("daemon is not running")
	}
	if err := syscall.Kill(s.daemon.Process.Pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("kill -9 daemon: %w", err)
	}
	s.waitDaemonExit()

	return nil
}

// activateGoswitchRetry activates the engine, tolerating transient failures
// right after an ibus-daemon restart (the CLI can hit the bus before it
// serves names again).
func (s *stand) activateGoswitchRetry(ctx context.Context) error {
	var err error
	for range activationAttempts {
		err = s.activateGoswitch(ctx)
		if err == nil {
			return nil
		}
		if werr := sleepCtx(ctx, activationRetryDelay); werr != nil {
			return werr
		}
	}

	return err
}

// witnessCharsRe extracts the char count from a witness line.
var witnessCharsRe = regexp.MustCompile(`chars=(\d+)`)

// witnessChars reads the focused entry's character count through the AT-SPI
// witness — the desktop-input-alive oracle for kill9-survive on a locked
// session (the daemon is dead; its log cannot witness anything). This is
// also the Phase-2 bridge: AT-SPI text readback grows into the TEST-04
// case matrix.
func (s *stand) witnessChars(ctx context.Context) (int, error) {
	out, err := s.focusWitness(ctx)
	if err != nil {
		return 0, err
	}
	m := witnessCharsRe.FindStringSubmatch(out)
	if m == nil {
		return 0, fmt.Errorf("witness line carries no char count: %q", out)
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, fmt.Errorf("parse char count in %q: %w", out, err)
	}

	return n, nil
}
