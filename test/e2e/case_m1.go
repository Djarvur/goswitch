package main

import (
	"context"
	"fmt"
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
)

// searchEntryMark identifies the stand's shell-owned input surface in the
// AT-SPI witness output. GNOME Shell exposes text entries as PASSWORD_TEXT
// nodes (live-verified in phase 01-03): the overview search entry when the
// session is unlocked, the lock-screen auth prompt when it is locked. Both
// take injected text, route it through the active IBus engine and discard
// it (search on close, password never commits) — exactly the disposal
// properties the stand needs from a surface it activates itself.
const searchEntryMark = ":PASSWORD_TEXT:"

// runM1Gate proves the M1 gate (phase success criterion 2): with goswitch-en
// the ACTIVE input source, physically injected keys reach the engine, the
// tap FSM decides on a double Right Shift, and the lifecycle order
// registered → focus_in holds by log timestamps.
func runM1Gate(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}

	if err := s.typeInSearch(ctx, "ghbdtn"); err != nil {
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

// typeInSearch opens the shell's search entry, types text into it and
// closes it again: the text lands in the search box — discarded on Escape —
// and never in the owner's windows. The search entry routes keys through
// the active IBus engine (live-verified: 12 key records for a 6-char type).
func (s *stand) typeInSearch(ctx context.Context, text string) (err error) {
	if oerr := s.openSearchSurface(ctx); oerr != nil {
		return oerr
	}
	defer func() {
		if cerr := s.closeSearchSurface(ctx); cerr != nil && err == nil {
			err = cerr
		}
	}()

	return s.injectText(ctx, text)
}

// openSearchSurface activates the shell's own text entry — the one input
// surface the stand can activate itself — and verifies through the AT-SPI
// witness that the entry (PASSWORD_TEXT node) actually took focus. On an
// unlocked session the super key opens the overview search; on a locked
// one it wakes the lock screen, whose auth prompt needs a second
// activation (Enter on the shield) before the entry appears. Background
// focus grabs are refused by Wayland focus-stealing prevention
// (live-verified: AT-SPI grabFocus errors on GTK4 windows and returns
// false on GTK3), so this replaces the plan's grabFocus step; the helper
// keeps its best-effort focus subcommand for platforms where grabs work.
// The idempotency check comes first: when the entry already has focus
// (overview left open by a crashed earlier run, or an already-awake lock
// prompt), pressing super would toggle the overview shut.
func (s *stand) openSearchSurface(ctx context.Context) error {
	focused, err := s.focusWitness(ctx)
	if err != nil {
		return err
	}
	if s.shellEntryFocused(focused) {
		return nil
	}
	if err := s.pressKey(ctx, "super"); err != nil {
		return err
	}
	if err := s.waitShellEntry(ctx, surfaceStep); err == nil {
		return nil
	} else if cerr := ctx.Err(); cerr != nil {
		return fmt.Errorf("surface activation aborted: %w", cerr)
	}
	// Focus sits on a non-entry shell node (lock-screen shield button,
	// overview preview): activate it — Enter reveals the lock prompt and
	// is inert on the empty overview search.
	if err := s.pressKey(ctx, "enter"); err != nil {
		return err
	}

	return s.waitShellEntry(ctx, witnessWait)
}

// shellEntryFocused reports whether one of the shell's own text entries
// has focus per the witness output. The gnome-shell prefix matters: an
// app's own password field (e.g. a password manager) must not pass for
// the stand's surface, or injection would type into it.
func (s *stand) shellEntryFocused(witness string) bool {
	return strings.HasPrefix(witness, shellAppPrefix) && strings.Contains(witness, searchEntryMark)
}

func (s *stand) closeSearchSurface(ctx context.Context) error {
	return s.pressKey(ctx, "Escape")
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
