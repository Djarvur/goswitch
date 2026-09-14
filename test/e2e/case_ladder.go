package main

import (
	"context"
	"fmt"
)

// Ladder cases of plan 02-05: the replacement ladder's ACTUAL level and the
// buffer reset triggers, both on the Chromium driver of plan 02-02 — the
// surface the ADR-003 matrix case names, and one that survives Enter/Escape
// (Pitfall 6 is exactly why these cases are NOT on zenity).

// actualLevelMark is the debug-record oracle of the ladder: the correction
// record's FIRST attribute after msg is the ladder level (D-21), so the
// substring pins the level the daemon ACTUALLY chose for this client —
// never the caps-bit assumption (ADR-003). LIVE FINDING (2026-09-15, plan
// 02-05 first run): google-chrome 153 reports caps 0x29 — CapSurroundingText
// INCLUDED — and applies DeleteSurroundingText; the actual level on this
// target is 1, and the plan's level-2 assumption (ibus#2354) is obsolete
// here. The unit corpus (TestActor_Level2NoCaps) carries the level-2
// contract; every matrix surface on this desktop (zenity, chromium, GTE)
// reports the surrounding-text bit, so no live level-2 witness exists.
const actualLevelMark = `"msg":"correction","level":1`

// verifyMatchMark is the verify-after oracle: the level-1 deletion is
// ack-less on 1.5.29, and the post-correction round (02-05) proves the
// client actually applied the replacement — the doubled-text defect of
// Pitfall 3 would land here as outcome:mismatch.
const verifyMatchMark = `"msg":"correction verify","outcome":"match"`

// runLadderChromium fixes the ACTUAL ladder level on the Chromium class
// (CORR-07, ADR-003, Pitfall 3): the 02-02 driver opens the fixture page,
// "ghbdtn" is injected through the physical path and the double tap runs
// the correction. The oracle is threefold: the daemon's -debug record pins
// the ladder level the daemon chose for THIS client, the verify-after round
// reports a match (the client applied the replacement — the ack-less
// deletion's compensation), and the content-exact readback holds «привет»:
// the doubled "ghbdtn привет" would mean the client silently ignored the
// deletion (Pitfall 3 live).
func runLadderChromium(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	if err := s.startChromium(ctx); err != nil {
		return err
	}
	defer s.closeChromium()

	if err := s.waitChromiumInput(ctx, 0); err != nil {
		return err
	}
	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return err
	}
	if err := s.waitChromiumInput(ctx, len(wordProbeEN)); err != nil {
		return fmt.Errorf("ladder-chromium injected field: %w", err)
	}

	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"action","n":2`, decisionWait); err != nil {
		return fmt.Errorf("ladder-chromium double-tap decision: %w", err)
	}
	// The actual-level oracle first (ADR-003): the case pins what the daemon
	// really did for THIS client, then proves the field and the verify-after
	// round agree.
	if err := s.waitForLog(ctx, actualLevelMark, correctionWait); err != nil {
		return fmt.Errorf("ladder-chromium actual ladder level: %w", err)
	}
	if err := s.waitForLog(ctx, `"msg":"correction","outcome":"done"`, correctionWait); err != nil {
		return fmt.Errorf("ladder-chromium correction: %w", err)
	}
	if err := s.waitForLog(ctx, verifyMatchMark, correctionWait); err != nil {
		return fmt.Errorf("ladder-chromium verify-after: %w", err)
	}

	if err := s.waitChromiumInput(ctx, len([]rune(wordResultRU))); err != nil {
		return fmt.Errorf("ladder-chromium applied correction: %w", err)
	}
	got, err := s.readChromiumText(ctx)
	if err != nil {
		return err
	}
	if got != wordResultRU {
		return fmt.Errorf("ladder-chromium readback %q, want %q (doubled text = ignored deletion, Pitfall 3)",
			got, wordResultRU)
	}

	return nil
}
