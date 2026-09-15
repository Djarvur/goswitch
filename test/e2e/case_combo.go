package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
// so every name is proven before it is trusted). The Task-2 extension runs
// the PASS part under -config and pins the live reload on the running
// daemon (CONF-02); the layout-single and super-space-alive cases close
// SWCH-01/SWCH-03 under the D-34 oracles.

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
	return []string{"Shift_R+Control_R", comboCanonicalName, "Shift+Control_R"}
}

// comboWindowStart/comboWindowReload are the tap_window_ms values of the
// case's live-reload section: the daemon starts at 500, the rewrite drops
// it to 200 — the reload-application record plus a deciding single tap pin
// that the RUNNING daemon picked the new document up (CONF-02).
const (
	comboWindowStart  = 500
	comboWindowReload = 200
	reloadApplyWait   = 10 * time.Second // the watcher's 200 ms debounce plus a live-desktop margin
)

// comboConfigTmpl is the COMPLETE config document of the case (no defaults
// overlay — the 03-02 strict-parse decision), parameterized by the tap
// window.
const comboConfigTmpl = `hotkeys:
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
`

// writeComboConfig writes the case's complete config document with the
// given tap window.
func writeComboConfig(path string, windowMs int) error {
	if err := os.WriteFile(path, []byte(fmt.Sprintf(comboConfigTmpl, windowMs)), configFilePerm); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}

	return nil
}

// runComboWordLayout proves the D-36 gesture end to end on the live desktop
// and the CONF-02 live consumption around it: the combo injection is
// canonicalized by a probe on the default daemon, then the PASS part runs
// under a -config daemon (tap_window_ms 500 written BEFORE the start), and
// a rewrite to 200 is pinned applied by the reload record — with a single
// tap deciding afterwards on the running daemon.
func runComboWordLayout(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	name, err := s.probeComboInjection(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("combo-word-layout: canonical combo injection %q\n", name)

	cfgPath := filepath.Join(s.tmpDir, "combo-word-layout.yaml")
	if err := writeComboConfig(cfgPath, comboWindowStart); err != nil {
		return fmt.Errorf("combo-word-layout: write temp config: %w", err)
	}
	if err := s.restartDaemonWithArgs("-config", cfgPath); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"config loaded"`, registrationWait); err != nil {
		return fmt.Errorf("combo-word-layout daemon -config spawn: %w", err)
	}
	if err := s.waitForLog(ctx, componentRegisteredMark, registrationWait); err != nil {
		return fmt.Errorf("combo-word-layout daemon re-registration: %w", err)
	}
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	if err := s.comboWordLayoutRound(ctx, name); err != nil {
		return err
	}

	// Live reload on the RUNNING daemon: rewrite the window, pin the
	// application by the reload record, then a single tap must still decide
	// (the flip record of the fresh, post-reload series).
	if err := writeComboConfig(cfgPath, comboWindowReload); err != nil {
		return fmt.Errorf("combo-word-layout: rewrite temp config: %w", err)
	}
	if err := s.waitForLog(ctx, `"msg":"config reloaded"`, reloadApplyWait); err != nil {
		return fmt.Errorf("combo-word-layout live reload application: %w", err)
	}
	enBase := s.countSub(`"msg":"mode","to":"en"`)
	if err := s.injectKeys(ctx, "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"mode","to":"en"`, enBase+1, decisionWait); err != nil {
		return fmt.Errorf("combo-word-layout post-reload single tap: %w", err)
	}

	return nil
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

	// The word must settle as LATIN before the gesture: a mid-word mode
	// flip (a live-desktop flake class) fails here, early and diagnosable,
	// instead of corrupting the correction oracle.
	if err := s.waitZenityText(ctx, wordProbeEN); err != nil {
		return fmt.Errorf("combo round typed word: %w", err)
	}

	if err := s.gateComboRound(ctx, name); err != nil {
		return err
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

// gateComboRound fires the combo and gates its three records as NEW
// occurrences (the probe already produced one of each), then pins the D-36
// order: the flip record of THIS round (the last mode-ru record) must
// postdate the round's settled completion record.
func (s *stand) gateComboRound(ctx context.Context, name string) error {
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

// runLayoutSingle proves SWCH-01 live under the D-34 oracle: a single Right
// Shift flips the INTERNAL mode EN→RU and a second flips it back RU→EN —
// the daemon's own mode records are the ONLY oracle (the internal flip
// never touches the session's XKB state, so there is no desktop source
// state to read — and reading one would be the forbidden Pitfall 3 of
// ADR-001).
func runLayoutSingle(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("layout-single needs the zenity entry surface (locked-session fallback engaged?)")
	}

	for _, tc := range []struct {
		tap  string
		mark string
	}{
		{"first tap flips EN→RU", `"msg":"mode","to":"ru"`},
		{"second tap flips RU→EN", `"msg":"mode","to":"en"`},
	} {
		base := s.countSub(tc.mark)
		if err := s.injectKeys(ctx, "Shift_R"); err != nil {
			return err
		}
		if err := s.waitForNew(ctx, tc.mark, base+1, decisionWait); err != nil {
			return fmt.Errorf("layout-single %s: %w", tc.tap, err)
		}
	}

	return nil
}

// superSettleWait lets the overview map before the closing esc lands.
const superSettleWait = 500 * time.Millisecond

// waitZenityFocused polls the witness until the zenity entry owns keyboard
// focus again — the settle gate after focus-churning episodes (the
// overview episode and the engine re-activation it forces).
func (s *stand) waitZenityFocused(ctx context.Context) error {
	deadline := time.Now().Add(witnessWait)
	for {
		now, err := s.focusWitness(ctx)
		if err != nil {
			return err
		}
		if strings.HasPrefix(now, "zenity:") {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("zenity entry not focused (witness %q)", now)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// runSuperSpaceAlive proves the D-34 working interpretation of SWCH-03: the
// shell's native Super mechanism is not broken by goswitch's presence —
// after the native Super key fires (the overview toggle the Super+Space
// switcher rides on), the daemon still owns its IBus component name
// (goswitchNameOwner) and a word correction still runs end to end through
// the engine.
func runSuperSpaceAlive(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	if kind != surfaceZenity {
		defer func() { _ = s.closeEntrySurface(ctx, kind) }()

		return errors.New("super-space-alive needs the zenity entry surface (locked-session fallback engaged?)")
	}

	// The native Super mechanism: the bare Super key toggles the shell
	// overview (closed right back with the proven esc name) — the shell
	// key handling the Super+Space switcher rides on. The Super+Space
	// chord itself is NOT injectable on this stack (every name spelling
	// falls back to its first letter — live finding 2026-09-15), and the
	// alt+Shift_L alternate binding was tried and removed: it can
	// genuinely switch the session's input source — a destructive side
	// effect on the owner's desktop the stand must not cause.
	if err := s.pressKey(ctx, "super"); err != nil {
		return fmt.Errorf("super-space-alive inject super: %w", err)
	}
	if err := sleepCtx(ctx, superSettleWait); err != nil {
		return err
	}
	if err := s.pressKey(ctx, "esc"); err != nil { // close the overview Super opened
		return err
	}

	// The daemon must still own its IBus name after the system chords fired.
	owner, err := goswitchNameOwner(ctx)
	if err != nil {
		return fmt.Errorf("super-space-alive name probe: %w", err)
	}
	if owner == "" {
		return errors.New("super-space-alive: the goswitch component name lost its owner after the native chords")
	}

	return s.superSpaceCorrectionRound(ctx, kind)
}

// superSpaceCorrectionRound proves the engine still works after the native
// chords: re-activate (the overview episode may have unbound the engine),
// WAIT for the entry to hold focus again (typing before the rebind races
// the IM renegotiation and loses early keys — the live finding of the
// first runs of this case), then run one full word correction.
func (s *stand) superSpaceCorrectionRound(ctx context.Context, kind surfaceKind) error {
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()

	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	if err := s.waitZenityFocused(ctx); err != nil {
		return fmt.Errorf("super-space-alive post-overview focus: %w", err)
	}
	keyBase := s.countSub(`"msg":"key"`)
	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"key"`, keyBase+minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("super-space-alive key visibility: %w", err)
	}
	// The word must settle as LATIN before the taps (the same early-flake
	// gate as the combo round).
	if err := s.waitZenityText(ctx, wordProbeEN); err != nil {
		return fmt.Errorf("super-space-alive typed word: %w", err)
	}
	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"action","n":2`, decisionWait); err != nil {
		return fmt.Errorf("super-space-alive double-tap decision: %w", err)
	}
	if err := s.waitForLog(ctx, `"msg":"correction","outcome":"done"`, correctionWait); err != nil {
		return fmt.Errorf("super-space-alive post-chord correction: %w", err)
	}
	if err := s.waitZenityText(ctx, wordResultRU); err != nil {
		return fmt.Errorf("super-space-alive applied correction: %w", err)
	}
	out, err := s.closeZenity(ctx)
	if err != nil {
		return err
	}
	if out != wordResultRU {
		return fmt.Errorf("super-space-alive oracle: entry printed %q, want %q", out, wordResultRU)
	}

	return nil
}
