package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// m1-case timing knobs and floors.
const (
	minKeyEvents   = 6 // ghbdtn yields 12 records (press+release); 6 is the floor
	keyWait        = 10 * time.Second
	decisionWait   = 5 * time.Second
	focusWait      = 5 * time.Second
	witnessWait    = 5 * time.Second
	surfaceStep    = 2 * time.Second // one activation step before escalation
	witnessPoll    = 100 * time.Millisecond
	shellAppPrefix = "gnome-shell:"
	zenityPrompt   = "goswitch e2e"
)

// searchEntryMark identifies the shell's own text entries in the AT-SPI
// witness output. GNOME Shell exposes them as PASSWORD_TEXT nodes
// (live-verified in phase 01-03): the overview search entry and the
// lock-screen auth prompt. Both take injected text, route it through the
// active IBus engine and discard it (search on close, password never
// commits) — the disposal properties the stand needs from a surface it
// activates itself.
const searchEntryMark = ":PASSWORD_TEXT:"

// surfaceKind names which input surface a case opened — it decides how the
// surface closes and which desktop-input oracle applies.
type surfaceKind int

const (
	// surfaceShell is the shell's own entry (search or lock prompt).
	surfaceShell surfaceKind = iota
	// surfaceZenity is the stand-spawned zenity entry window.
	surfaceZenity
)

// runM1Gate proves the M1 gate (phase success criterion 2): with goswitch-en
// the ACTIVE input source, physically injected keys reach the engine, the
// tap FSM decides on a double Right Shift, and the lifecycle order
// registered → focus_in holds by log timestamps.
func runM1Gate(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
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
	if err := s.waitForNew(ctx, `"msg":"key"`, minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("M1 key visibility: %w", err)
	}

	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"action","n":2`, decisionWait); err != nil {
		return fmt.Errorf("M1 double-tap decision: %w", err)
	}

	return assertRegisteredBeforeFocusIn(s)
}

// activateGoswitch makes goswitch-en the session's input source through
// IBus SetGlobalEngine (the `ibus engine` CLI) and waits for the resulting
// focus_in. The plan's primary mechanism — writing gsettings
// sources+current — is runtime-ignored by GNOME 46's shell (live-verified
// this phase: the sources list updates but no focus_in arrives and no keys
// route), so SetGlobalEngine is the working activation; gsettings stays
// snapshotted and restored by the teardown regardless.
func (s *stand) activateGoswitch(ctx context.Context) error {
	if _, err := runCmd(ctx, "ibus", "engine", "goswitch-en"); err != nil {
		return fmt.Errorf("activate goswitch-en via SetGlobalEngine: %w", err)
	}

	return s.waitForLog(ctx, "focus_in", focusWait)
}

// openEntrySurface activates the stand's own input surface and verifies
// through the AT-SPI witness that it owns keyboard focus BEFORE any text is
// injected — injected text never reaches the owner's windows (TEST-03).
// Preferred surface: a spawned zenity entry, which takes focus on map
// (background grabs are refused by Wayland focus-stealing prevention —
// GTK4 errors, GTK3 returns false, live-verified; the helper keeps a
// best-effort focus subcommand for platforms where grabs work). On a
// locked session the zenity window cannot surface, so the fallback wakes
// the shell's own entry instead: super opens the overview search, Enter
// reveals the lock-screen prompt.
func (s *stand) openEntrySurface(ctx context.Context) (surfaceKind, error) {
	focused, err := s.focusWitness(ctx)
	if err != nil {
		return 0, err
	}
	if s.shellEntryFocused(focused) {
		return surfaceShell, nil
	}
	if err := s.startZenity(ctx); err != nil {
		return 0, err
	}
	if err := s.waitZenityEntry(ctx); err == nil {
		return surfaceZenity, nil
	} else if ctx.Err() != nil {
		s.reapZenity()

		return 0, fmt.Errorf("surface activation aborted: %w", ctx.Err())
	}
	// The zenity window stayed hidden (locked session) — fall back to the
	// shell's own entry.
	s.reapZenity()

	return s.openShellEntry(ctx)
}

// closeEntrySurface dismisses the surface opened by openEntrySurface.
func (s *stand) closeEntrySurface(ctx context.Context, kind surfaceKind) error {
	switch kind {
	case surfaceZenity:
		_, err := s.closeZenity(ctx)

		return err
	case surfaceShell:
		return s.pressKey(ctx, "Escape")
	case surfaceChromium:
		// Driver-managed surface (surface.go): chromium cases close their
		// instance through the driver, never through the entry-surface path.
		return nil
	}

	return nil
}

// openShellEntry wakes the shell's text entry: super opens the overview
// search (unlocked) or wakes the lock screen (locked); when a non-entry
// node holds focus afterwards (lock-screen shield button), Enter reveals
// the auth prompt.
func (s *stand) openShellEntry(ctx context.Context) (surfaceKind, error) {
	if err := s.pressKey(ctx, "super"); err != nil {
		return 0, err
	}
	if err := s.waitShellEntry(ctx, surfaceStep); err == nil {
		return surfaceShell, nil
	} else if cerr := ctx.Err(); cerr != nil {
		return 0, fmt.Errorf("surface activation aborted: %w", cerr)
	}
	if err := s.pressKey(ctx, "enter"); err != nil {
		return 0, err
	}
	if err := s.waitShellEntry(ctx, witnessWait); err != nil {
		return 0, err
	}

	return surfaceShell, nil
}

// shellEntryFocused reports whether one of the shell's own text entries
// has focus per the witness output. The gnome-shell prefix matters: an
// app's own password field (e.g. a password manager) must not pass for
// the stand's surface, or injection would type into it.
func (s *stand) shellEntryFocused(witness string) bool {
	return strings.HasPrefix(witness, shellAppPrefix) && strings.Contains(witness, searchEntryMark)
}

// startZenity spawns a self-focusing text entry — the stand's input surface
// on an unlocked session: a freshly mapped window takes keyboard focus
// without any grab (live-verified in phase 01-03). The entry's stdout (the
// typed text, printed on Enter) is the delivery readback oracle.
func (s *stand) startZenity(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "zenity", "--entry", "--title=goswitch-e2e", "--text="+zenityPrompt)
	out := &bytes.Buffer{}
	cmd.Stdout = out
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start zenity: %w", err)
	}
	s.zenity = cmd
	s.zenityOut = out

	return nil
}

// waitZenityEntry polls the witness until the spawned entry owns keyboard
// focus — the gate that makes injection safe.
func (s *stand) waitZenityEntry(ctx context.Context) error {
	deadline := time.Now().Add(witnessWait)
	for {
		now, err := s.focusWitness(ctx)
		if err != nil {
			return err
		}
		if strings.HasPrefix(now, "zenity:") && strings.Contains(now, ":TEXT") {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("zenity entry did not take focus (witness %q)", now)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// closeZenity dismisses the entry with Enter (OK prints the text) and
// returns the collected stdout — the proof the typed text was delivered.
func (s *stand) closeZenity(ctx context.Context) (string, error) {
	if s.zenity == nil {
		return "", nil
	}
	if err := s.pressKey(ctx, "enter"); err != nil {
		s.reapZenity()

		return "", err
	}
	done := make(chan struct{})
	go func() {
		_ = s.zenity.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(killGrace):
		s.reapZenity()
	}
	out := strings.TrimSpace(s.zenityOut.String())
	s.zenity = nil

	return out, nil
}

// reapZenity force-kills a leftover entry (teardown and error paths); a
// no-op once the entry was closed or reaped.
func (s *stand) reapZenity() {
	if s.zenity == nil {
		return
	}
	if s.zenity.Process != nil {
		_ = s.zenity.Process.Kill()
	}
	_ = s.zenity.Wait()
	s.zenity = nil
}

// focusWitness reports the currently focused input surface
// ("<app>:<role>:chars=N") through the AT-SPI helper — the verification
// primitive that the stand's own surface activation worked (TEST-03's
// intent under Wayland constraints).
func (s *stand) focusWitness(ctx context.Context) (string, error) {
	out, err := runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "witness")
	if err != nil {
		return "", fmt.Errorf("focus witness: %w", err)
	}

	return out, nil
}

// waitShellEntry polls the witness until one of the shell's text entries
// reports focused — proof the injected activation key reached the
// compositor and the shell entry owns keyboard focus.
func (s *stand) waitShellEntry(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		now, err := s.focusWitness(ctx)
		if err != nil {
			return err
		}
		if s.shellEntryFocused(now) {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("shell entry did not take focus (witness %q)", now)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// assertRegisteredBeforeFocusIn checks the lifecycle order by JSON record
// timestamps: the component registered strictly before the first focus_in.
func assertRegisteredBeforeFocusIn(s *stand) error {
	registered, err := s.firstRecordTime("component registered")
	if err != nil {
		return err
	}
	focusIn, err := s.firstRecordTime("focus_in")
	if err != nil {
		return err
	}
	if !registered.Before(focusIn) {
		return fmt.Errorf("lifecycle order violated: component registered at %v is not before focus_in at %v",
			registered, focusIn)
	}

	return nil
}
