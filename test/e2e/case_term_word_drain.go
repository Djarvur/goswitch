package main

// The phase-10 round-4 pty-drain case (plan 10-17, TERM-02 / G-10-1): the
// byte-order proof of the delivery machine. Per pacing combination of the
// explicit -debug matrix (the daemon-side pacing override, fail-closed
// without -debug) a fresh gnome-terminal fixture with a RAW pty capture
// (stty raw — the capture is a byte stream, not a cooked line) runs one
// full correction round, and the oracles are:
//
//  1. the pty ORDER delta: every BackSpace byte (0x7f) sits STRICTLY
//     before the replacement word's bytes — each erase reached the client
//     before the commit (the pty-drain proof of the round-4 machine);
//  2. the journal delivery marks: chunks-sent → commit-sent ordered, the
//     flip-armed mark after the commit, and the flip-released mark only
//     AFTER its honest release cause (a confirming push or a keystroke —
//     never inside the correction window);
//  3. the frozen CLIPBOARD sentinel byte-untouched (ADR-008 D4/D5).
//
// The code is delivered by plan 10-17; the LIVE rows are plan 10-18's
// work. Every row lands a dated verdict row in the case's journal section;
// a failure lands the honest-abort row and fails the case (exit contract)
// — never a retry loop.

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// The case's journal section and the daemon-side marks it orders: the
// section header text is the contract (the journal owns it;
// TestSpikeJournalSectionsExist pins it). The delivery-mark family is the
// round-4 machine's INFO rows — outcomes and counters only (D-20/D-21).
const (
	secTermDrain = "Фаза 10 round-4 pty-drain (term-word-drain)"

	drainChunksMark  = `"msg":"correction delivery","stage":"chunks-sent"`
	drainCommitMark  = `"msg":"correction delivery","stage":"commit-sent"`
	drainFlipArmMark = `"stage":"flip-armed"`
	drainFlipRelMark = `"stage":"flip-released"`

	// The honest release causes of the flip mark (the mark's attribute
	// order is stage, reason — the case greps the exact forms): a release
	// with any other reason (grace, verify) would be the round-2 geometry
	// back from the dead.
	drainFlipRelPush = `"stage":"flip-released","reason":"push"`
	drainFlipRelKey  = `"stage":"flip-released","reason":"keystroke"`

	// drainWordRU is the replacement word's literal — the pty-order
	// oracle's needle (the corpus, not user data).
	drainWordRU = "привет"

	// The pacing-matrix environment of the -debug daemon (the mirrors of
	// cmd/goswitchd's constants — a main package cannot be imported; the
	// case owns its literal copies and the row records the values, so a
	// daemon-side rename fails the row loudly at 10-18).
	envDrainGap    = "GOSWITCH_DEBUG_LEVEL2_FORWARD_GAP_MS"
	envDrainSettle = "GOSWITCH_DEBUG_TERMINAL_DRAIN_SETTLE_MS"

	// drainBackSpaceByte is the erase byte the raw pty stream carries for
	// every forwarded BackSpace.
	drainBackSpaceByte = 0x7f

	// drainFlipWait bounds the flip-released wait per attempt: a pushing
	// VTE client releases the flip itself (the confirming push), a
	// push-less client needs the case's probe keystroke.
	drainFlipWait = 3 * time.Second
)

// drainConfigDoc is the case's complete config document (the 03-02
// no-overlay rule; the autocorrectConfigDoc form) with the HOST-PARITY
// flip gate: correction.flip_after_correction rides true — the owner's
// configuration, and the only one that exercises the flip half of the
// delivery-mark oracle (armed after the commit, released only by a cause).
const drainConfigDoc = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true # host parity: the flip arms and its marks are on oracle duty
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

// drainCombo is one pacing-matrix combination: the -debug override values
// applied to the row's daemon (empty = the journal-proven defaults —
// the matrix's first cell proves the defaults too, not only the lever).
type drainCombo struct {
	name     string
	gapMs    string
	settleMs string
}

// drainMatrix is the explicit latency matrix of the plan: the defaults,
// a tight pacing (the queue gets less machine time per forward), a loose
// one (a long gap and an extended delivery window — the client gets an
// ample term). The ORDER oracle must hold at EVERY combination — that is
// the machine's by-construction promise.
func drainMatrix() []drainCombo {
	return []drainCombo{
		{name: "default"},
		{name: "tight", gapMs: "5", settleMs: "150"},
		{name: "loose", gapMs: "50", settleMs: "400"},
	}
}

// drainRow appends one dated verdict row to the case's journal section
// (the appendSpikeRow form).
func drainRow(label, cmd, observed, verdict string) error {
	return appendSpikeRow(secTermDrain, label+" "+time.Now().Format("15:04:05"), cmd, observed, verdict)
}

// drainPane builds the RAW-capture pane program: the line discipline is
// turned off so the capture is the byte stream itself (every forwarded
// BackSpace lands as 0x7f, every commit as raw UTF-8) — the cooked `cat`
// of the 10-05 case would deliver only the net lines and erase the very
// ordering this case proves.
func drainPane(ready, pty string) []string {
	return []string{"bash", "-c", "stty raw -echo; echo ready > " + ready + "; cat > " + pty}
}

// applyPacingEnv sets (or unsets — the empty value is the DEFAULT cell)
// the -debug override variables for the NEXT daemon spawn: exec.Command
// inherits the case process's environment, so the values ride the restart.
func applyPacingEnv(gapMs, settleMs string) error {
	for _, kv := range []struct{ name, v string }{
		{envDrainGap, gapMs},
		{envDrainSettle, settleMs},
	} {
		if kv.v == "" {
			if err := os.Unsetenv(kv.name); err != nil {
				return fmt.Errorf("unset %s: %w", kv.name, err)
			}

			continue
		}
		if err := os.Setenv(kv.name, kv.v); err != nil {
			return fmt.Errorf("set %s=%s: %w", kv.name, kv.v, err)
		}
	}

	return nil
}

// drainRestartDaemon (re)spawns the case's -debug daemon on the case's own
// config document with the combo's override environment in force, waits
// out the readiness ladder and re-activates the engine (the
// startAutocorrectConfigDaemon ladder, parameterized).
func drainRestartDaemon(ctx context.Context, s *stand, combo drainCombo) error {
	const caseName = "term-word-drain"
	if err := applyPacingEnv(combo.gapMs, combo.settleMs); err != nil {
		return err
	}
	cfgPath := filepath.Join(s.tmpDir, "term-drain.yaml")
	if err := os.WriteFile(cfgPath, []byte(drainConfigDoc), configFilePerm); err != nil {
		return fmt.Errorf("%s: write config: %w", caseName, err)
	}
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	if err := s.restartDaemonWithArgs("-config", cfgPath); err != nil {
		return err
	}
	for _, mark := range []string{configLoadedMark, componentRegisteredMark, ctlListeningMark} {
		if err := s.waitForLog(ctx, mark, registrationWait); err != nil {
			return fmt.Errorf("%s: drain daemon -config spawn (%s): %w", caseName, mark, err)
		}
	}

	return s.activateGoswitchRetry(ctx)
}

// waitForDrainPty polls the raw capture until the replacement word's bytes
// are present AND every BackSpace byte sits strictly before them — the
// pty-order oracle. Returns the capture's hex for the row.
func waitForDrainPty(ctx context.Context, path string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	var hexr, note string
	for {
		data, err := os.ReadFile(path)
		if err == nil {
			hexr = hex.EncodeToString(data)
			lastBS, wordAt, ok := drainPtyOrder(data)
			switch {
			case ok:
				return hexr, nil
			case wordAt < 0:
				note = "слово не дошло до pty"
			case lastBS < 0:
				note = "ни одного 0x7f до слова — форварды не дошли"
			default:
				note = "0x7f после байтов слова — порядок нарушен"
			}
		}
		if time.Now().After(deadline) {
			return hexr, fmt.Errorf("pty-порядок не сошёлся за %v (%s; last hex %s)", timeout, note, hexr)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return hexr, err
		}
	}
}

// drainPtyOrder reports the last BackSpace byte before the replacement
// word, the word's offset, and whether the ORDER oracle holds: the word is
// present, at least one erase precedes it, and NO erase byte follows it.
func drainPtyOrder(data []byte) (lastBS, wordAt int, ok bool) {
	wordAt = bytes.Index(data, []byte(drainWordRU))
	if wordAt < 0 {
		return 0, -1, false
	}
	lastBS = -1
	for i, b := range data[:wordAt] {
		if b == drainBackSpaceByte {
			lastBS = i
		}
	}
	for _, b := range data[wordAt:] {
		if b == drainBackSpaceByte {
			return lastBS, wordAt, false
		}
	}

	return lastBS, wordAt, lastBS >= 0
}

// drainMarkOrder is the journal half of the oracle: within the whole log,
// the marks must be ordered chunks-sent < commit-sent < flip-armed, and
// the flip-released mark must exist strictly AFTER its release cause — the
// last key-trace record that followed the commit (a confirming push or the
// probe keystroke releases; nothing releases inside the window). The cause
// search is scoped to the CURRENT round's segment — the tail from the
// flip-armed mark (WR-08 of the 10-REVIEW): the matrix accumulates all
// three combo rows into ONE journal (restartDaemonWithArgs never
// rotates), and an unscoped whole-log Contains would close this row on a
// PREVIOUS row's honest cause even when the current release rode a
// dishonest reason — the round-2 geometry the oracle exists to convict.
func drainMarkOrder(log string) (ok bool, observed string) {
	chunks := strings.LastIndex(log, drainChunksMark)
	commit := strings.LastIndex(log, drainCommitMark)
	armed := strings.LastIndex(log, drainFlipArmMark)
	released := strings.LastIndex(log, drainFlipRelMark)
	ordered := chunks >= 0 && commit > chunks && armed > commit && released > armed
	tail := log
	if armed >= 0 {
		tail = log[armed:]
	}
	cause := strings.Contains(tail, drainFlipRelPush) || strings.Contains(tail, drainFlipRelKey)
	ok = ordered && cause
	observed = fmt.Sprintf("marks: chunks@%d commit@%d flip-armed@%d flip-released@%d cause-honest=%v",
		chunks, commit, armed, released, cause)

	return ok, observed
}

// drainFlipReleased waits for the flip-released mark; a pushing VTE client
// releases the flip itself (the confirming push), a push-less one gets the
// case's probe keystroke as the honest cause.
func (s *stand) drainFlipReleased(ctx context.Context, base int) error {
	if err := s.waitForNew(ctx, drainFlipRelMark, base+1, drainFlipWait); err == nil {
		return nil
	}
	if err := s.injectText(ctx, "x"); err != nil {
		return fmt.Errorf("release probe keystroke: %w", err)
	}

	return s.waitForNew(ctx, drainFlipRelMark, base+1, drainFlipWait)
}

// runDrainRow drives one full matrix combination: fresh daemon with the
// combo's override, observer warm round, raw-capture fixture, the typed
// word + the double Right Shift gesture, and the three oracles.
func runDrainRow(ctx context.Context, s *stand, combo drainCombo) error {
	const caseName = "term-word-drain"
	tag := "row-" + combo.name
	if err := drainRestartDaemon(ctx, s, combo); err != nil {
		return err
	}
	if err := s.termObserverRound(ctx, caseName); err != nil {
		return err
	}
	tf, err := s.openTermFixture(ctx, caseName, tag, drainPane)
	if err != nil {
		return err
	}
	defer tf.close()
	if err := s.termArmGate(ctx, caseName); err != nil {
		return err
	}

	// The frozen CLIPBOARD pre-state (the D5 sentinel overwrite is
	// documented in the row, the termRowClipFrozen form).
	if err := clipCopy(ctx); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	before := clipReadNoNewline(ctx)
	if before != primarySentinelMark {
		return fmt.Errorf("%s: CLIPBOARD pre-state %q, want sentinel %q", caseName, before, primarySentinelMark)
	}

	keyBase := s.countSub(keyTraceMark)
	if err := s.injectText(ctx, corpusLatin); err != nil {
		return fmt.Errorf("%s: type %q: %w", caseName, corpusLatin, err)
	}
	if err := s.waitForNew(ctx, keyTraceMark, keyBase+minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("%s: key visibility: %w", caseName, err)
	}

	return s.drainRoundOracles(ctx, tf, combo)
}

// drainRoundOracles drives the gesture and the case's three oracles, and
// lands the row (pass or FAIL with the observables — the exit contract).
func (s *stand) drainRoundOracles(ctx context.Context, tf *termFixture, combo drainCombo) error {
	const caseName = "term-word-drain"
	tag := "row-" + combo.name
	bases := map[string]int{
		"action":   s.countSub(actionDoubleMark),
		"done":     s.countSub(correctionDoneMa),
		"level2":   s.countSub(termLevel2Mark),
		"chunks":   s.countSub(drainChunksMark),
		"commit":   s.countSub(drainCommitMark),
		"armed":    s.countSub(drainFlipArmMark),
		"released": s.countSub(drainFlipRelMark),
	}
	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return fmt.Errorf("%s: double gesture: %w", caseName, err)
	}
	for _, w := range []struct {
		base, mark, tag string
		wait            time.Duration
	}{
		{"action", actionDoubleMark, "double-tap decision", decisionWait},
		{"done", correctionDoneMa, "correction done", correctionWait},
		{"level2", termLevel2Mark, "terminal branch level-2 mark", correctionWait},
		{"chunks", drainChunksMark, "chunks-sent mark", correctionWait},
		{"commit", drainCommitMark, "commit-sent mark", correctionWait},
		{"armed", drainFlipArmMark, "flip-armed mark", correctionWait},
	} {
		if err := s.waitForNew(ctx, w.mark, bases[w.base]+1, w.wait); err != nil {
			return fmt.Errorf("%s: %s: %w", caseName, w.tag, err)
		}
	}

	// Oracle 1: the pty order delta.
	ptyHex, perr := waitForDrainPty(ctx, tf.ptyPath, termCaptureWait)
	// Oracle 2: the journal delivery-mark order (the flip-released wait
	// rides its own cause discipline).
	if ferr := s.drainFlipReleased(ctx, bases["released"]); ferr != nil && perr == nil {
		perr = fmt.Errorf("flip-released mark: %w", ferr)
	}
	// Oracle 3: the frozen CLIPBOARD sentinel.
	after := clipReadNoNewline(ctx)
	markOK, markObs := drainMarkOrder(s.logText())

	return drainRowVerdict(ctx, combo, tag, perr, markOK, markObs, after, ptyHex)
}

// drainRowVerdict lands the row with the honest verdict for the oracles'
// joint outcome (the exit contract: a FAIL also fails the case).
func drainRowVerdict(
	ctx context.Context, combo drainCombo, tag string, perr error, markOK bool, markObs, after, ptyHex string,
) error {
	observed := fmt.Sprintf("venue: %s; combo=%s (gap=%q settle=%q); pty_hex=%s; marks %s; clip_after=%s",
		venueMarker(ctx), combo.name, combo.gapMs, combo.settleMs, ptyHex, markObs, after)
	cmd := "armGate + type " + corpusLatin + " + double Shift_R + pty-порядок (0x7f прежде байтов слова)" +
		" + порядок delivery-меток + release-причина флипа"
	switch {
	case perr == nil && markOK && after == primarySentinelMark:
		return drainRow(tag, cmd, observed,
			"Verdict: pass (каждый 0x7f в pty строго прежде байтов слова; delivery-метки упорядочены, "+
				"флип-релиз после своей причины; CLIPBOARD байт-нетронут)")
	case perr == nil && markOK:
		return drainRow(tag, cmd, observed, "Verdict: FAIL (CLIPBOARD изменён)")
	case perr == nil:
		return drainRow(tag, cmd, observed, "Verdict: FAIL (порядок delivery-меток нарушен)")
	default:
		if rerr := drainRow(tag, cmd, observed, "Verdict: FAIL (pty-порядок)"); rerr != nil {
			return rerr
		}

		return perr
	}
}

// runTermWordDrain drives the phase-10 round-4 pty-drain matrix: one full
// correction round per pacing combination, each with the pty-order,
// delivery-mark and CLIPBOARD oracles. A failure lands the honest-abort
// row and fails the case (the exit contract).
func runTermWordDrain(ctx context.Context, s *stand) error {
	const caseName = "term-word-drain"
	var ran []string
	for _, combo := range drainMatrix() {
		if err := runDrainRow(ctx, s, combo); err != nil {
			obs := "case aborted: " + boundRowText(err.Error())
			if rerr := drainRow(caseName, "pty-drain matrix ["+strings.Join(ran, ",")+"]", obs,
				"Verdict: aborted"); rerr != nil {
				fmt.Fprintf(os.Stderr, "e2e: %s deferred journal row: %v\n", caseName, rerr)
			}

			return err
		}
		ran = append(ran, combo.name)
	}
	fmt.Printf("%s: pty-drain rows green for the whole pacing matrix (rows in %s)\n", caseName, spikeJournal)

	return nil
}
