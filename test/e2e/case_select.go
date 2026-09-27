package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Djarvur/goswitch/internal/clipboard"
)

// Selection-spike artifacts of plan 03-03 (CORR-03, A1/A2/A5): the spike
// pins, per surface, whether the client pushes a selection anchor at all
// (anchor != cursor in the surrounding-text trace, D-30) and what the
// replacement ladder ACTUALLY does while a selection is active — the same
// "pin the actual, never the assumption" discipline as the 02-05
// ladder-chromium case. The case PASSES with whatever verdict the desktop
// produces; the verdicts feed the Task-2 selection branch and the matrix v2
// fields (03-07).

// surroundingMark identifies the daemon's DEBUG surrounding_text records —
// the trace that carries both wire positions (cursor_pos, anchor_pos).
const surroundingMark = `"msg":"surrounding_text"`

// selectProbeWait budgets one candidate injection: the client needs a beat
// to push a fresh surrounding text after the selection key lands.
const selectProbeWait = 2 * time.Second

// selectSettleWait lets the field and the verify-after round settle before
// the readback classifies the actual rung.
const selectSettleWait = 1500 * time.Millisecond

// selectAllCandidates returns the ydotool key-name candidates for a
// select-all injection, in try order (Pitfall 5: ydotool 0.1.8 resolves
// unknown names by first-letter fallback — "Escape"→E, "space"→S — so every
// name is proven live before it is trusted; the working one is reported in
// the verdict). A function, not a var: the strict lint forbids mutable
// globals (matrix.go idiom).
func selectAllCandidates() []string {
	return []string{selectAllCanonical, "ctrl+A", "Control+a"}
}

// selectFieldEN is the spike's probe field: two SPEC words typed in EN mode
// (13 runes). The Task-1 daemon has no selection branch yet, so the double
// tap runs the WORD pipeline over the last token — exactly the probe the
// spike wants: what does the ladder do while a selection is active?
const selectFieldEN = wordProbeEN + " " + wordProbeEN

// selectExpectSuffixApplied is the field after a delete+commit the selection
// was transparent to: the last ghbdtn replaced by привет, head untouched.
const selectExpectSuffixApplied = wordProbeEN + " " + wordResultRU

// selectAllCanonical is the select-all injection name proven live by the
// Task-1 spike on every anchor-reporting surface of the stand (GTE:
// cursor=0/anchor=len; chromium: cursor=len/anchor=0 — ydotool 0.1.8
// resolves the combo correctly; "Control+a" types a literal "ca" through the
// first-letter fallback, Pitfall 5).
const selectAllCanonical = "ctrl+a"

// selectCorrectResult is the select-correct oracle: the whole selected mixed
// field converts by the run rule — the foreign EN word to привет, the
// engine-committed Cyrillic word re-committed unchanged (D-23 inside the
// selection).
const selectCorrectResult = wordResultRU + " " + wordResultRU

// selectClipboardProbe is the replacement the clipboard case puts into the
// clipboard — no edge whitespace, so runCmd's trimming cannot eat a byte of
// the compared content.
const selectClipboardProbe = "goswitch-e2e-clip-probe"

// clipboardConfigYAML is the COMPLETE config document of the rung case (a
// config must be complete — no defaults overlay, the 03-02 strict-parse
// decision) with the D-28 switch ON.
const clipboardConfigYAML = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: true
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
`

// configFilePerm is the temp config's mode (mnd).
const configFilePerm = 0o600

// runSelectClipboard proves the D-28 rung's live mechanics
// (unreachable-primary form, exactly the plan's fallback): on this desktop
// every anchor-reporting surface applies the primary rung under a selection
// (spike table: GTE delete+commit, chromium commit-replaces-selection), so
// no live correction can reach a verify-after mismatch — the daemon-side
// rung has no live driver yet. The case therefore pins what CAN be driven
// live: (a) the daemon spawns with -config and arms the rung (the config
// loaded record — the flag wiring of D-28), and (b) the round-trip
// mechanics themselves run through the daemon's own clipboard package on
// the live session — Save byte-exactly, Set through stdin-only wl-copy
// (verified by wl-paste), best-effort Restore of the owner's bytes. The
// matrix v2 decision (03-07) carries the unreachability note.
func runSelectClipboard(ctx context.Context, s *stand) error {
	// (a) the daemon-side wiring: a complete temp config with the rung on,
	// a fresh daemon on -config, the config-loaded record in the log.
	cfgPath := filepath.Join(s.tmpDir, "clipboard-rung.yaml")
	if err := os.WriteFile(cfgPath, []byte(clipboardConfigYAML), configFilePerm); err != nil {
		return fmt.Errorf("select-clipboard: write temp config: %w", err)
	}
	if err := s.restartDaemonWithArgs("-config", cfgPath); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"config loaded"`, registrationWait); err != nil {
		return fmt.Errorf("select-clipboard daemon -config spawn: %w", err)
	}
	if err := s.waitForLog(ctx, componentRegisteredMark, registrationWait); err != nil {
		return fmt.Errorf("select-clipboard daemon re-registration: %w", err)
	}

	return clipboardRoundTripLive(ctx)
}

// clipboardRoundTripLive drives the daemon's own clipboard client over the
// live session: save the owner's clipboard, put the probe in, verify the
// paste source holds it, restore, verify the owner's bytes came back
// (whitespace-trimmed comparison — wl-copy --trim-newline may strip one
// trailing newline of the original, the pinned-flag nuance).
func clipboardRoundTripLive(ctx context.Context) error {
	c := clipboard.New()
	saved, had, err := c.Save(ctx)
	if err != nil {
		return fmt.Errorf("select-clipboard save: %w", err)
	}
	if err := c.Set(ctx, []byte(selectClipboardProbe)); err != nil {
		return fmt.Errorf("select-clipboard set: %w", err)
	}
	got, err := runCmd(ctx, "wl-paste", "--no-newline")
	if err != nil {
		return fmt.Errorf("select-clipboard paste readback: %w", err)
	}
	if got != selectClipboardProbe {
		return fmt.Errorf("select-clipboard: clipboard holds %q after Set, want the probe %q",
			got, selectClipboardProbe)
	}
	if err := c.Restore(ctx, saved, had); err != nil {
		return fmt.Errorf("select-clipboard restore: %w", err)
	}
	after, afterErr := runCmd(ctx, "wl-paste", "--no-newline")
	if had {
		if afterErr != nil || strings.TrimSpace(after) != strings.TrimSpace(string(saved)) {
			return fmt.Errorf("select-clipboard: owner clipboard not restored (readback %q, err %w, saved %q)",
				after, afterErr, string(saved))
		}

		return nil
	}
	if afterErr == nil && strings.TrimSpace(after) != "" && after != selectClipboardProbe {
		return fmt.Errorf("select-clipboard: empty original came back as %q", after)
	}

	return nil
}

// surroundingReport is one parsed surrounding_text debug record: both wire
// positions of the client's push.
type surroundingReport struct {
	cursorPos uint32
	anchorPos uint32
}

// selectionActive reports whether the push describes an active selection
// (D-30: anchor != cursor).
func (r surroundingReport) selectionActive() bool {
	return r.anchorPos != r.cursorPos
}

// lastSurrounding parses the NEWEST surrounding_text record from the daemon
// log — the freshest selection state the client reported.
func (s *stand) lastSurrounding() (surroundingReport, bool) {
	var rep surroundingReport
	found := false
	for _, line := range s.logLines() {
		if !strings.Contains(line, surroundingMark) {
			continue
		}
		var rec struct {
			CursorPos uint32 `json:"cursor_pos"`
			AnchorPos uint32 `json:"anchor_pos"`
		}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		rep = surroundingReport{cursorPos: rec.CursorPos, anchorPos: rec.AnchorPos}
		found = true
	}

	return rep, found
}

// waitNewSurrounding blocks until the daemon log holds MORE than want
// surrounding_text records, then reports the newest one — the push that
// landed after the injection under probe.
func (s *stand) waitNewSurrounding(ctx context.Context, want int) (surroundingReport, bool) {
	deadline := time.Now().Add(selectProbeWait)
	for {
		if s.countSub(surroundingMark) > want {
			rep, ok := s.lastSurrounding()

			return rep, ok
		}
		if time.Now().After(deadline) {
			return surroundingReport{}, false
		}
		if err := sleepCtx(ctx, logPollInterval); err != nil {
			return surroundingReport{}, false
		}
	}
}

// probeSelectAll tries every select-all candidate name until one produces an
// active selection in the trace (a fresh push with anchor != cursor). A
// candidate that resolves to the wrong physical key leaves no anchor change
// (or types a stray character — visible in the readback, harmless for the
// verdict) and the probe moves on.
func (s *stand) probeSelectAll(ctx context.Context) (string, surroundingReport, bool) {
	for _, cand := range selectAllCandidates() {
		before := s.countSub(surroundingMark)
		if err := s.pressKey(ctx, cand); err != nil {
			continue
		}
		if rep, ok := s.waitNewSurrounding(ctx, before); ok && rep.selectionActive() {
			return cand, rep, true
		}
	}

	return "", surroundingReport{}, false
}

// selectionVerdict classifies the ACTUAL replacement rung the surface
// executed while the selection was active, from the settled readback plus
// the daemon's own records: transparent (delete+commit applied at the
// cursor, the selection ignored), commit-replaces-selection (the commit
// landed as typing-over-the-selection), or nothing-applied (mismatch skip
// or the client ignored the ladder).
func (s *stand) selectionVerdict(readback string) string {
	switch readback {
	case selectExpectSuffixApplied:
		return "delete+commit applied, selection transparent"
	case wordResultRU:
		return "commit replaced the active selection"
	case selectFieldEN:
		if s.countSub(`"reason":"verify-mismatch"`) > 0 {
			return "nothing applied (verify-mismatch skip)"
		}

		return "nothing applied"
	default:
		return fmt.Sprintf("other (readback %q)", readback)
	}
}

// selectSurfaceDriver bundles the per-surface calls of the spike probe so
// the driver-owned surfaces (GTE, chromium) share one probe body.
type selectSurfaceDriver struct {
	surface   string
	start     func(context.Context) error
	close     func()
	waitChars func(context.Context, int) error
	readback  func(context.Context) (string, error)
}

// runSelectSpike drives one driver-owned surface through the full probe:
// type the field, find a working select-all injection, fire the double tap
// under the active selection, print the verdict.
func (s *stand) runSelectSpike(ctx context.Context, d selectSurfaceDriver) error {
	if err := d.start(ctx); err != nil {
		return err
	}
	defer d.close()

	if err := d.waitChars(ctx, 0); err != nil {
		return fmt.Errorf("select-smoke %s surface: %w", d.surface, err)
	}
	if err := s.injectText(ctx, selectFieldEN); err != nil {
		return err
	}
	if err := d.waitChars(ctx, len([]rune(selectFieldEN))); err != nil {
		return fmt.Errorf("select-smoke %s injected field: %w", d.surface, err)
	}

	name, rep, ok := s.probeSelectAll(ctx)
	s.reportSelectProbe(d.surface, name, rep, ok)

	return s.doubleTapUnderSelection(ctx, d.readback)
}

// runSelectSmoke drives the selection spike over every driver surface of the
// stand (zenity, gnome-text-editor, chromium) and prints the per-surface
// verdict table. The case PASSes with any verdict — it pins the actual
// behavior, it does not assert an assumption (the matrix v2 rows of 03-07
// will carry these verdicts).
func runSelectSmoke(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}

	if err := selectSpikeZenity(ctx, s); err != nil {
		return err
	}
	if err := s.runSelectSpike(ctx, selectSurfaceDriver{
		surface:   "gnome-text-editor",
		start:     s.startGTE,
		close:     s.closeGTE,
		waitChars: s.waitGTEInput,
		readback:  s.readGTEText,
	}); err != nil {
		return err
	}

	return s.runSelectSpike(ctx, selectSurfaceDriver{
		surface:   "chromium",
		start:     s.startChromium,
		close:     s.closeChromium,
		waitChars: s.waitChromiumInput,
		readback:  s.readChromiumText,
	})
}

// runSelectCorrect proves the selection branch live (CORR-03, D-30, plan
// 03-03 task 2) on the surface the spike table gave a working anchor:
// gnome-text-editor. The mixed field "ghbdtn привет" is assembled through
// both real printing branches (EN transit, RU engine commits after the
// flip), ctrl+a selects it all (the canonical spike name; GTE pushes
// cursor=0/anchor=len — the positive-offset geometry of Pitfall 6), and the
// double Right Shift corrects exactly the SELECTION: the field settles at
// "привет привет" — the foreign EN word converted, the Cyrillic word
// untouched, nothing outside the range modified.
func runSelectCorrect(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	if err := s.startGTE(ctx); err != nil {
		return err
	}
	defer s.closeGTE()

	if err := s.waitGTEInput(ctx, 0); err != nil {
		return fmt.Errorf("select-correct surface: %w", err)
	}
	if err := s.assembleMixedField(ctx, "select-correct"); err != nil {
		return err
	}
	mixed := wordProbeEN + " " + wordResultRU
	if err := s.waitGTEInput(ctx, len([]rune(mixed))); err != nil {
		return fmt.Errorf("select-correct mixed field assembled: %w", err)
	}

	// Select all and gate on the anchor push itself: the selection must be
	// observable before the tap means anything (D-30 — a surface without
	// the anchor cannot carry this case).
	before := s.countSub(surroundingMark)
	if err := s.pressKey(ctx, selectAllCanonical); err != nil {
		return err
	}
	rep, ok := s.waitNewSurrounding(ctx, before)
	if !ok || !rep.selectionActive() {
		return fmt.Errorf("select-correct: no selection anchor after %s (report %+v)"+
			" — selection unobservable", selectAllCanonical, rep)
	}

	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"action","n":2`, decisionWait); err != nil {
		return fmt.Errorf("select-correct double-tap decision: %w", err)
	}
	if err := s.waitForLog(ctx, `"msg":"correction","outcome":"done"`, correctionWait); err != nil {
		return fmt.Errorf("select-correct selection correction: %w", err)
	}

	return s.waitGTEText(ctx, selectCorrectResult)
}

// assembleMixedField types "ghbdtn привет" into the focused surface through
// both real printing branches: the EN word transits, the single-tap flip
// switches the engine to RU, the second word is committed by the engine as
// Cyrillic — the mixed-selection input of the select cases.
func (s *stand) assembleMixedField(ctx context.Context, caseName string) error {
	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"key"`, minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("%s key visibility: %w", caseName, err)
	}
	// The separator rides the typing path (ydotool "space" resolves to the
	// physical S key — the 02-03 live trap).
	if err := s.injectText(ctx, " "); err != nil {
		return err
	}
	if err := s.flipToRU(ctx, caseName); err != nil {
		return err
	}

	return s.injectText(ctx, wordProbeEN)
}

// waitGTEText polls the pid-keyed GTE readback until the document holds
// exactly want — the AT-SPI bridge can lag the field update, so the oracle
// rides the lag out instead of reading once (the waitZenityText idiom).
func (s *stand) waitGTEText(ctx context.Context, want string) error {
	deadline := time.Now().Add(witnessWait)
	var last string
	for {
		out, err := s.readGTEText(ctx)
		if err == nil {
			last = out
			if out == want {
				return nil
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("select-correct: document did not settle to %q (readback %q)", want, last)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// selectSpikeZenity runs the selection probe on the zenity GTK entry — the
// surface whose surrounding pushes carry no selection anchor (live finding
// of the first spike run, 2026-09-15: the selection key lands but no push
// with anchor != cursor ever arrives, so the selection is unobservable D-30).
func selectSpikeZenity(ctx context.Context, s *stand) error {
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("select-smoke needs the zenity entry surface (locked-session fallback engaged?)")
	}

	if err := s.injectText(ctx, selectFieldEN); err != nil {
		return err
	}
	if err := s.waitZenityChars(ctx, len([]rune(selectFieldEN))); err != nil {
		return fmt.Errorf("select-smoke zenity injected field: %w", err)
	}

	name, rep, ok := s.probeSelectAll(ctx)
	s.reportSelectProbe("zenity", name, rep, ok)

	return s.doubleTapUnderSelection(ctx, s.readFocusedTextRaw)
}

// reportSelectProbe prints the anchor half of the verdict: whether the
// surface pushes a selection anchor at all and with which working
// select-all name (a surface without an anchor makes every selection case
// on it a documented D-30 degradation, never a blocker).
func (s *stand) reportSelectProbe(surface, name string, rep surroundingReport, ok bool) {
	if !ok {
		fmt.Printf("select-smoke %-18s anchor=NO select-name=(none worked)\n", surface)

		return
	}
	fmt.Printf("select-smoke %-18s anchor=YES select-name=%s cursor=%d anchor=%d\n",
		surface, name, rep.cursorPos, rep.anchorPos)
}

// doubleTapUnderSelection fires the double Right Shift under whatever state
// the select probe left, waits for the decision and the correction records,
// settles, readbacks the field and prints the rung verdict.
func (s *stand) doubleTapUnderSelection(ctx context.Context, readback func(context.Context) (string, error)) error {
	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"action","n":2`, decisionWait); err != nil {
		return fmt.Errorf("select-smoke double-tap decision: %w", err)
	}
	// Either the correction completes or a refusal lands — both are verdicts.
	_ = s.waitForLog(ctx, `"msg":"correction"`, correctionWait)
	if err := sleepCtx(ctx, selectSettleWait); err != nil {
		return err
	}
	got, err := readback(ctx)
	if err != nil {
		return fmt.Errorf("select-smoke readback: %w", err)
	}
	fmt.Printf("select-smoke rung verdict: %s\n", s.selectionVerdict(got))

	return nil
}
