package main

// The phase-10 VTE flip live case (plan 10-08, TERM-02, G-10-1 cause B):
// the same VTE word scenario as term-word-vte but under
// correction.flip_after_correction: true — the owner rule «режим следует за
// скриптом результата» riding the plan's deferred-flip machine. The E3
// verdict forbids the flip from following the commit at 2–8 ms (the
// churn-delayed client discards the stale engine's signal, the committed
// word is lost), so the row's oracle is the JOURNAL DELIVERY ORDER: the
// done record, then the settlement (the client's confirming push — the
// verify-after match), and only then the mode/switch marks —
// SetGlobalEngine must not appear between done and settlement. Two rows:
// flip-word-bash (the order + pty oracle) and flip-clip-frozen (the
// CLIPBOARD sentinel byte-untouched under the flip config; its typing runs
// in the FLIPPED mode, so the correction converts back and the pty oracle
// carries the Latin corpus). Every row lands a dated verdict row in the
// phase journal; a failure lands the honest-abort row and fails the case —
// never a retry loop.

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
)

// The case's journal section and marks: the section header text is the
// contract (TestSpikeJournalSectionsExist pins the sections). The
// settlement marks are the verify-after's DEBUG CONCLUSIONS — the plan's
// settlement рубеж is «push match / verify»: the client's confirming push
// (match) or the quiet budget conclusion (timeout). The stand's VTE
// answers Requires with a lagged cache (the live finding of the first
// run), so the honest row accepts both forms; the stale-retry record is
// the WR-05 intermediate, never a settlement. The stand daemon runs
// -debug, so the DEBUG records are greppable; the switch mark is the
// daemon wiring's SetGlobalEngine INFO record.
const (
	secTermFlip = "Фаза 10 живой ряд VTE flip (term-word-vte-flip)"

	termFlipSettleMatch   = `"msg":"correction verify","outcome":"match"`
	termFlipSettleTimeout = `"msg":"correction verify","outcome":"timeout"`
	termFlipSettleMis     = `"msg":"correction verify","outcome":"mismatch"`
	termFlipModeMark      = `"msg":"mode","to"`
	termFlipSwitchMark    = `"msg":"switch_engine"`

	// termFlipDocFlipLine replaces the frozen false pin of the
	// autocorrectConfigDoc form — the case owns its document (the 03-02
	// no-overlay rule); the frozen configs are never touched.
	termFlipDocFlipLine = "  flip_after_correction: true # plan 10-08: the deferred flip rides the live rows"
)

// isSettleRecord reports whether the journal line is a verify-after
// CONCLUSION (the plan's settlement рубеж in either form).
func isSettleRecord(ln string) bool {
	return strings.Contains(ln, termFlipSettleMatch) ||
		strings.Contains(ln, termFlipSettleTimeout) ||
		strings.Contains(ln, termFlipSettleMis)
}

// settleCount counts the verify-after conclusions in the daemon log.
func (s *stand) settleCount() int {
	n := 0
	for _, ln := range s.logLines() {
		if isSettleRecord(ln) {
			n++
		}
	}

	return n
}

// termFlipConfigDoc is the case's complete config document: the
// autocorrectConfigDoc form with the flip gate ON — the mechanism row's
// own document, not an edit of the frozen pins.
const termFlipConfigDoc = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
` + termFlipDocFlipLine + `
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
`

// termFlipRow appends one dated verdict row to the case's journal section.
func termFlipRow(label, cmd, observed, verdict string) error {
	return appendSpikeRow(secTermFlip, label+" "+time.Now().Format("15:04:05"), cmd, observed, verdict)
}

// flipWordRound drives one word round and gates the JOURNAL DELIVERY
// ORDER: done, then the settlement (the client's confirming push — the
// verify-after match), then the mode/switch marks. The gate reads the
// daemon journal between the marks — a mode/switch record between done and
// settlement is the E3 loss signature and fails the row. The final newline
// (the pty flush) comes only after the settlement is proven.
func (s *stand) flipWordRound(ctx context.Context, caseName, text string) error {
	keyBase := s.countSub(keyTraceMark)
	if err := s.injectText(ctx, text); err != nil {
		return fmt.Errorf("%s: type %q: %w", caseName, text, err)
	}
	if err := s.waitForNew(ctx, keyTraceMark, keyBase+minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("%s: key visibility: %w", caseName, err)
	}
	actionBase := s.countSub(actionDoubleMark)
	doneBase := s.countSub(correctionDoneMa)
	levelBase := s.countSub(termLevel2Mark)
	settleBase := s.settleCount()
	modeBase := s.countSub(termFlipModeMark)
	switchBase := s.countSub(termFlipSwitchMark)
	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return fmt.Errorf("%s: double gesture: %w", caseName, err)
	}
	if err := s.waitForNew(ctx, actionDoubleMark, actionBase+1, decisionWait); err != nil {
		return fmt.Errorf("%s: double-tap decision: %w", caseName, err)
	}
	if err := s.waitForNew(ctx, correctionDoneMa, doneBase+1, correctionWait); err != nil {
		return fmt.Errorf("%s: correction done: %w", caseName, err)
	}
	if err := s.waitForNew(ctx, termLevel2Mark, levelBase+1, correctionWait); err != nil {
		return fmt.Errorf("%s: terminal branch level-2 mark: %w", caseName, err)
	}
	// The settlement: the verify-after's CONCLUSION — the client's
	// confirming push (match) or the quiet budget (timeout). Either closes
	// the delivery question before the mode moves (the E3 contract); the
	// WR-05 stale-retry record is an intermediate, not a settlement.
	settleDeadline := time.Now().Add(correctionWait)
	for s.settleCount() <= settleBase {
		if time.Now().After(settleDeadline) {
			return fmt.Errorf("%s: settlement (verify-after match/timeout) never concluded within %v",
				caseName, correctionWait)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
	if err := s.waitForNew(ctx, termFlipModeMark, modeBase+1, correctionWait); err != nil {
		return fmt.Errorf("%s: mode record after the settlement: %w", caseName, err)
	}
	if err := s.waitForNew(ctx, termFlipSwitchMark, switchBase+1, correctionWait); err != nil {
		return fmt.Errorf("%s: switch record after the settlement: %w", caseName, err)
	}
	if err := s.assertFlipDeliveryOrder(); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}

	return s.injectText(ctx, "\n")
}

// assertFlipDeliveryOrder greps the daemon journal for the E3 contract: no
// mode/switch record may sit between the LAST done record and the
// settlement that follows it — SetGlobalEngine between the commit and its
// confirmation is the loss window the plan forbids by construction.
func (s *stand) assertFlipDeliveryOrder() error {
	log := s.logLines()
	lastDone, settle, firstModeAfterDone := -1, -1, -1
	for i, ln := range log {
		if strings.Contains(ln, correctionDoneMa) {
			lastDone, settle, firstModeAfterDone = i, -1, -1

			continue
		}
		if lastDone < 0 {
			continue
		}
		switch {
		case settle < 0 && isSettleRecord(ln):
			settle = i
		case settle < 0 &&
			(strings.Contains(ln, termFlipModeMark) || strings.Contains(ln, termFlipSwitchMark)):
			if firstModeAfterDone < 0 {
				firstModeAfterDone = i
			}
		}
	}
	if lastDone < 0 || settle < 0 {
		return fmt.Errorf("delivery order unreadable: done@%d settle@%d", lastDone, settle)
	}
	if firstModeAfterDone >= 0 && firstModeAfterDone < settle {
		return fmt.Errorf("E3 signature: a mode/switch record sits between done@%d and settlement@%d "+
			"(first@%d) — the flip followed the commit", lastDone, settle, firstModeAfterDone)
	}

	return nil
}

// flipWordRowCore is the flip case's word row: fixture, arming gate, the
// order-gated round and the pty-hex oracle.
func (s *stand) flipWordRowCore(
	ctx context.Context, caseName, rowTag string, text, wantHex string,
) error {
	tf, err := s.openTermFixture(ctx, caseName, rowTag, capturePane("bash", "-c"))
	if err != nil {
		return err
	}
	defer tf.close()
	if err := s.termArmGate(ctx, caseName); err != nil {
		return err
	}
	if err := s.flipWordRound(ctx, caseName, text); err != nil {
		return err
	}
	hexr, cerr := waitForCaptureHex(ctx, tf.ptyPath, wantHex, termCaptureWait)
	observed := fmt.Sprintf("venue: %s; pty_hex=%s; delivery_order=done<settlement<mode", venueMarker(ctx), hexr)
	cmd := "armGate + type " + text + " + double Shift_R + journal order oracle (done<settle<mode) + pty-hex"
	if cerr == nil {
		return termFlipRow(rowTag, cmd, observed,
			"Verdict: pass (pty байт-точен; флип освободился подтверждением доставки — E3 порядок держится)")
	}
	if rerr := termFlipRow(rowTag, cmd, observed,
		"Verdict: FAIL (pty не байт-точен или порядок доставки нарушен)"); rerr != nil {
		return rerr
	}

	return cerr
}

// flipRowClipFrozen is the frozen CLIPBOARD row under the flip config: the
// sentinel rides CLIPBOARD through the round and wl-paste must read it
// back byte-identical (ADR-008 D4/D5 — the deferred flip changes the MODE
// delivery, never the clipboard). The round runs in the FLIPPED mode (row
// 1's correction flipped EN→RU), so the typed Latin corpus commits as
// Cyrillic and the correction converts BACK — the pty oracle carries the
// Latin corpus, the delivery order oracle still holds.
func (s *stand) flipRowClipFrozen(ctx context.Context, caseName string) error {
	const rowTag = "flip-clip-frozen"
	tf, err := s.openTermFixture(ctx, caseName, rowTag, capturePane("bash", "-c"))
	if err != nil {
		return err
	}
	defer tf.close()
	if err := s.termArmGate(ctx, caseName); err != nil {
		return err
	}
	if err := clipCopy(ctx, primarySentinelMark); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	before := clipReadNoNewline(ctx)
	if before != primarySentinelMark {
		return fmt.Errorf("%s: CLIPBOARD pre-state %q, want sentinel %q", caseName, before, primarySentinelMark)
	}
	if err := s.flipWordRound(ctx, caseName, corpusLatin); err != nil {
		return err
	}
	wantHex := hex.EncodeToString([]byte(corpusLatin + "\n"))
	hexr, cerr := waitForCaptureHex(ctx, tf.ptyPath, wantHex, termCaptureWait)
	after := clipReadNoNewline(ctx)
	observed := fmt.Sprintf("venue: %s; clip_before=%s; clip_after=%s; pty_hex=%s"+
		" (набор шёл в перевернутом режиме — коррекция вернула латиницу; CLIPBOARD перезаписан сентинелом — синтетика)",
		venueMarker(ctx), before, after, hexr)
	cmd := "printf %s GS-SPIKE-SENTINEL | wl-copy → flip word round → wl-paste --no-newline"
	if cerr == nil && after == primarySentinelMark {
		return termFlipRow(rowTag, cmd, observed,
			"Verdict: pass (CLIPBOARD байт-нетронут и при активной машине флипа; слово исправлено)")
	}
	if rerr := termFlipRow(rowTag, cmd, observed,
		"Verdict: FAIL (клипборд изменён или слово не исправлено)"); rerr != nil {
		return rerr
	}
	if cerr != nil {
		return cerr
	}

	return fmt.Errorf("%s: CLIPBOARD changed: %q, want sentinel %q", caseName, after, primarySentinelMark)
}

// runTermWordVTEFlip drives the plan-10-08 flip rows on the stand: the
// flip-enabled config daemon, the observer warm round, the order-gated
// word row and the frozen CLIPBOARD row. A failure lands the honest-abort
// row and fails the case (exit contract).
func runTermWordVTEFlip(ctx context.Context, s *stand) error {
	const caseName = "term-word-vte-flip"
	err := runTermFlipRounds(ctx, s)
	if err == nil {
		fmt.Printf("%s: flip word row and flip-clip-frozen green (rows in %s)\n", caseName, spikeJournal)

		return nil
	}
	obs := "case aborted: " + boundRowText(err.Error())
	if rerr := termFlipRow(caseName, "VTE flip round", obs, "Verdict: aborted"); rerr != nil {
		fmt.Fprintf(os.Stderr, "e2e: %s deferred journal row: %v\n", caseName, rerr)
	}

	return err
}

// runTermFlipRounds is the row program of the flip case (extracted so the
// honest-abort wrapper stays under the complexity budget).
func runTermFlipRounds(ctx context.Context, s *stand) error {
	const caseName = "term-word-vte-flip"
	if err := startAutocorrectConfigDaemon(ctx, s, termFlipConfigDoc); err != nil {
		return err
	}
	if err := s.termObserverRound(ctx, caseName); err != nil {
		return err
	}
	if err := s.flipWordRowCore(ctx, caseName, "flip-word-bash", corpusLatin, corpusHexRU); err != nil {
		return err
	}

	return s.flipRowClipFrozen(ctx, caseName)
}
