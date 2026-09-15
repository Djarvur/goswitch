package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Combo-gesture artifacts of plan 03-04 (SWCH-02/D-36): the combo-word-layout
// case proves the word-then-layout gesture live — the word pipeline corrects
// ghbdtn→привет, the mode flip record lands strictly AFTER the settled
// completion record (the SPEC action name fixes the order: "correct the word
// AND switch the layout"), and the zenity stdout oracle pins the field
// content. The combo injection name is canonicalized by a live probe
// (Pitfall 5: ydotool 0.1.8 resolves unknown names by first-letter fallback,
// so every name is proven before it is trusted).

// comboLogMark is the daemon's combo entry record — a NEW grep-stable slog
// shape, additive to the "action"/"mode" contracts the matrix greps.
const comboLogMark = `"msg":"combo","kind":"word-layout"`

// comboProbeWait budgets one combo-name candidate: the press round-trips
// through uinput→mutter→ibus before the daemon log shows the record.
const comboProbeWait = 3 * time.Second

// comboCandidates returns the ydotool key-name candidates for the
// Shift+Control_R physical combo, in try order. A function, not a var: the
// strict lint forbids mutable globals (matrix.go idiom).
func comboCandidates() []string {
	return []string{"Shift_R+Control_R", "SHIFT_R+CTRL_R", "Shift+Control_R"}
}

// runComboWordLayout proves the D-36 gesture end to end on the live desktop:
// with goswitch-en active and "ghbdtn" typed into the zenity entry, the
// canonicalized combo injection fires the word pipeline (correction done:
// the field settles at "привет") and then flips the mode EN→RU — the log
// order oracle pins that the flip record lands strictly after the settled
// completion record, and the stdout oracle pins the corrected word.
func runComboWordLayout(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	name, err := s.probeComboInjection(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("combo-word-layout: canonical combo injection %q\n", name)

	return s.comboWordLayoutRound(ctx, name)
}

// probeComboInjection opens a throwaway entry and tries the combo-name
// candidates until one produces the daemon's combo record — the live probe
// that canonicalizes the injection (the 03-03 select-all precedent). The
// probe runs on an EMPTY field: the working combo refuses with empty-buffer
// and flips the mode (switching is the primary intent when there is no
// word), so the probe flips it back before handing over. Stray input of
// failed candidates and the probe's buffer state die with the throwaway
// surface (closeEntrySurface → FocusOut hard-reset).
func (s *stand) probeComboInjection(ctx context.Context) (string, error) {
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return "", errors.New("combo probe needs the zenity entry surface (locked-session fallback engaged?)")
	}

	for _, cand := range comboCandidates() {
		before := s.countSub(comboLogMark)
		if err := s.pressKey(ctx, cand); err != nil {
			continue
		}
		if err := s.waitForNew(ctx, comboLogMark, before+1, comboProbeWait); err != nil {
			continue
		}
		// The empty-field combo flipped the mode; flip it back so the round
		// starts from EN exactly like the daemon's boot state.
		if err := s.injectKeys(ctx, "Shift_R"); err != nil {
			return "", err
		}
		if err := s.waitForLog(ctx, `"msg":"mode","to":"en"`, decisionWait); err != nil {
			return "", fmt.Errorf("combo probe flip-back: %w", err)
		}

		return cand, nil
	}

	return "", fmt.Errorf("no combo injection name produced the combo record (tried %v)", comboCandidates())
}

// comboWordLayoutRound drives one full combo round on a FRESH entry:
// ghbdtn typed, the combo injected, the correction-done and mode records
// gated as NEW occurrences (the probe already produced one of each), the
// D-36 log order pinned via record timestamps, and both oracles — the
// content-exact readback and the zenity stdout — checked.
func (s *stand) comboWordLayoutRound(ctx context.Context, name string) error {
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("combo round needs the zenity entry surface (locked-session fallback engaged?)")
	}

	keyBase := s.countSub(`"msg":"key"`)
	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"key"`, keyBase+minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("combo round key visibility: %w", err)
	}

	comboBase := s.countSub(comboLogMark)
	doneBase := s.countSub(`"msg":"correction","outcome":"done"`)
	modeBase := s.countSub(`"msg":"mode","to":"ru"`)
	if err := s.pressKey(ctx, name); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, comboLogMark, comboBase+1, decisionWait); err != nil {
		return fmt.Errorf("combo round gesture: %w", err)
	}
	if err := s.waitForNew(ctx, `"msg":"correction","outcome":"done"`, doneBase+1, correctionWait); err != nil {
		return fmt.Errorf("combo round word correction: %w", err)
	}
	if err := s.waitForNew(ctx, `"msg":"mode","to":"ru"`, modeBase+1, decisionWait); err != nil {
		return fmt.Errorf("combo round layout flip: %w", err)
	}

	// D-36 order oracle: the flip record of THIS round (the last mode-ru
	// record) must postdate the round's settled completion record.
	tDone, err := s.lastRecordTime(`"msg":"correction","outcome":"done"`)
	if err != nil {
		return fmt.Errorf("combo round done timestamp: %w", err)
	}
	tMode, err := s.lastRecordTime(`"msg":"mode","to":"ru"`)
	if err != nil {
		return fmt.Errorf("combo round mode timestamp: %w", err)
	}
	if !tMode.After(tDone) {
		return fmt.Errorf("combo order (D-36): mode flip at %v is not after correction done at %v", tMode, tDone)
	}

	if err := s.waitZenityText(ctx, wordResultRU); err != nil {
		return fmt.Errorf("combo round applied correction: %w", err)
	}
	out, err := s.closeZenity(ctx)
	if err != nil {
		return err
	}
	if out != wordResultRU {
		return fmt.Errorf("combo-word-layout oracle: entry printed %q, want %q", out, wordResultRU)
	}

	return nil
}

// lastRecordTime returns the JSON timestamp of the LAST log record
// containing sub — firstRecordTime's sibling for rounds that repeat within
// one run (the probe's mode record precedes the round's pinned one).
func (s *stand) lastRecordTime(sub string) (time.Time, error) {
	var last time.Time
	found := false
	for _, line := range s.logLines() {
		if !strings.Contains(line, sub) {
			continue
		}
		ts, err := parseLogTime(line)
		if err != nil {
			return time.Time{}, err
		}
		last = ts
		found = true
	}
	if !found {
		return time.Time{}, fmt.Errorf("no daemon log record containing %q", sub)
	}

	return last, nil
}
