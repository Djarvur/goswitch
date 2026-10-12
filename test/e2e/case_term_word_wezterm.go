package main

// The phase-10 wezterm live word case (plan 10-06, TERM-02 criterion 1):
// inside the VM stand (the vm-primary venue verdict) one wezterm-nightly
// window, the caps re-pin row BEFORE any routing decision (Pitfall 4: the
// legacy 0x9 strings are historical — the XIM client is gone), then the
// ×2 two-greens discipline (Pitfall 1: the IME arming race makes a single
// green round a known false victory) — two full word rounds in one case
// run — plus the mid-line contract row (OQ 2 «contract»: the word typed
// straight at the cursor behind a prefix separator), the frozen CLIPBOARD
// sentinel row and the OQ 3 primary-leg probe (a capability probe for the
// record: sentinel in PRIMARY → the select chord through the daemon's own
// probe-forward channel → wl-paste --primary). The wezterm context rides
// the verify-dead behavioral class (plan 10-03): a fresh context learns it
// on the first gesture (the app-verification budget expires — no pushes
// come), the SECOND gesture rides the learned class to done+level-2. Every
// row lands a dated verdict row in the phase journal with the venue
// marker; a failure lands the honest-abort row and fails the case (exit
// contract) — never a retry loop.

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// The case's journal section and frozen marks: the section header text is
// the contract (the journal owns it; TestSpikeJournalSectionsExist pins
// it). termLevel2Mark pins the TERMINAL write branch — the DEBUG
// correction record's level attribute (the chunked branch of plan 10-04;
// the GTK pending machine logs Level1 and must never answer here).
const (
	secTermWez = "Фаза 10 живые ряды wezterm (term-word-wezterm)"

	// wezLearnMark is the verify-dead learning transition (setVerifyDead's
	// INFO marker): the first gesture on a fresh nightly context expires
	// the app-verification budget while the context lives and the
	// behavioral class learns the terminal routing for the context's
	// remaining life (plan 10-03).
	wezLearnMark = `"msg":"terminal class"`

	// wezSkipVerifyMark is the learning gesture's skip record (the closed
	// verify-timeout reason — skipCorrection's vocabulary).
	wezSkipVerifyMark = `"msg":"correction skipped","reason":"verify-timeout"`

	// wezCapsMark is the daemon capabilities record the re-pin row reads
	// (the caps bitmap of the nightly context through the daemon).
	wezCapsMark = `"msg":"capabilities"`

	// wezCapsWait bounds the caps-record wait after the fixture engages.
	wezCapsWait = 5 * time.Second

	// wezFixture is the stand-owned nightly config (minimal on purpose:
	// the owner's ~/.wezterm.lua is NEVER touched — Pitfall 7; NO
	// enable_wayland override — the nightly rides its native Wayland
	// default, the 2026-10-12 verdict; the phase-9 xim/wayland fixture
	// flags are historical instruments of a build that is gone).
	wezFixture = "gs-e2e-wezterm-stand.lua"

	// wezArmGateTries bounds the arming ladder (the plan's «ограниченные
	// ретраи»): the uinput churn of the first ydotool calls re-mints the
	// engine (the live finding of the first run — the BackSpace press
	// landed in the disable window of the old mint and only the release
	// traced), so the self-erasing probe pair retries until the floor is
	// met, never assuming delivery.
	wezArmGateTries = 3

	// wezCapsNone is the honest "(none)" observed when no capabilities
	// record landed for the fixture (Pitfall 3: the value is recorded,
	// never a routing key).
	wezCapsNone = "(none)"
)

// wezRow appends one dated verdict row to the case's journal section (the
// appendSpikeRow form; the label carries the row tag, the time the date).
func wezRow(label, cmd, observed, verdict string) error {
	return appendSpikeRow(secTermWez, label+" "+time.Now().Format("15:04:05"), cmd, observed, verdict)
}

// wezFixtureRound is one open wezterm round: the capture paths and the
// window handle (close() reaps the pane by the tmp marker and the gui by
// the fixture-name scope — the spike family's teardown discipline; the
// EOF chord is not traced on the stand, so the kill is the fallback).
type wezFixtureRound struct {
	wcap      *weztermCapture
	readyPath string
	ptyPath   string
}

// close tears the fixture down: the pane kill then the launcher/gui reap.
// Safe on the nil/error paths.
func (wf *wezFixtureRound) close() {
	if wf == nil {
		return
	}
	killTermPane(wf.readyPath)
	wf.wcap.close()
}

// openWezFixture acquires the wezterm round: the stand-owned config spawn
// (--config-file TOP-LEVEL form — the A3 journal pin; after the start
// subcommand the CLI rejects it outright), the pane-ready proof and the
// daemon focus_in gate above the spawn baseline (map focus usually
// self-fires on the stand's stock kernel).
func (s *stand) openWezFixture(ctx context.Context, caseName, tag string) (*wezFixtureRound, error) {
	readyPath := filepath.Join(s.tmpDir, tag+"-ready")
	ptyPath := filepath.Join(s.tmpDir, tag+"-in")
	focusBase := s.focusBaseline()
	wcap, err := startWeztermCapture(ctx, weztermFixturePath(wezFixture), "gs-e2e term-word-wezterm",
		capturePane("bash", "-c")(readyPath, ptyPath)...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", caseName, err)
	}
	wf := &wezFixtureRound{wcap: wcap, readyPath: readyPath, ptyPath: ptyPath}
	if werr := waitPaneReady(ctx, caseName, readyPath); werr != nil {
		wf.close()

		return nil, fmt.Errorf("%s: %w", caseName, werr)
	}
	fmt.Printf("%s: %s pane ready — awaiting daemon focus_in (map focus usually self-fires)\n", caseName, tag)
	if werr := s.waitFocusInFrom(ctx, caseName, tag+" wezterm pane", focusBase); werr != nil {
		wf.close()

		return nil, fmt.Errorf("%s: %w", caseName, werr)
	}

	return wf, nil
}

// wezObservedCaps returns the newest capabilities record that landed after
// the spawn baseline (the caps-record count taken BEFORE the fixture
// spawned) — the fixture context's own signature, not an older context's.
// "(none)" is the honest observed when nothing landed (Pitfall 3: the
// value is recorded, never a routing key).
func (s *stand) wezObservedCaps(base int) string {
	lines := s.logLines()
	seen := 0
	last := wezCapsNone
	for _, ln := range lines {
		if !strings.Contains(ln, wezCapsMark) {
			continue
		}
		seen++
		if seen <= base {
			continue
		}
		var rec struct {
			Caps string `json:"caps"`
		}
		if err := json.Unmarshal([]byte(ln), &rec); err == nil && rec.Caps != "" {
			last = rec.Caps
		}
	}

	return last
}

// wezCapsRepin lands the caps re-pin row BEFORE any routing decision
// (Pitfall 4): the observed SetCapabilities signature of the nightly
// context through the daemon. The row records what THIS build announces —
// the routing is the live classification, never keyed on the value.
func (s *stand) wezCapsRepin(ctx context.Context, capsBase int) error {
	deadline := time.Now().Add(wezCapsWait)
	for s.countSub(wezCapsMark) <= capsBase && time.Now().Before(deadline) {
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
	observed := fmt.Sprintf("venue: %s; caps=%s (наблюдаемая подпись ночника; "+
		"легаси-строки 0x9 исторические — Pitfall 4)", venueMarker(ctx), s.wezObservedCaps(capsBase))

	return wezRow("caps re-pin", "SetCapabilities в журнале демона для контекста wezterm-nightly (до решений)",
		observed, "Verdict: подпись записана до решений (маршрутизация живая — classifyTerminal, не зашитая)")
}

// wezArmGate is the rounds' arming gate — BY THE DAEMON KEY-TRACE (the
// plan's gate, never an AT-SPI witness) under the plan's bounded retry
// ladder: one self-erasing probe pair (x + BackSpace — the venue-proven
// named key) per attempt, the floor is the pair's 4 records above the
// attempt's own baseline. The self-erasing pair leaves the tty line
// buffer empty, so no probe bytes ride into the row oracles; a floor
// still unmet after the retries fails the round honestly.
func (s *stand) wezArmGate(ctx context.Context, caseName string) error {
	for attempt := 1; attempt <= wezArmGateTries; attempt++ {
		base := s.countSub(keyTraceMark)
		if err := s.injectText(ctx, "x"); err != nil {
			return fmt.Errorf("%s: arming probe type: %w", caseName, err)
		}
		if err := s.injectKeys(ctx, "BackSpace"); err != nil {
			return fmt.Errorf("%s: arming probe erase: %w", caseName, err)
		}
		if err := s.waitForNew(ctx, keyTraceMark, base+armProbeKeys, keyWait); err == nil {
			return nil
		}
		fmt.Printf("%s: arming attempt %d/%d below the key-trace floor — retrying\n",
			caseName, attempt, wezArmGateTries)
	}

	return fmt.Errorf("%s: arming gate: key-trace floor unmet after %d attempts", caseName, wezArmGateTries)
}

// wezGestureToTerminalDone drives the Double Right Shift gesture and waits
// for the terminal write branch to execute: a fresh nightly context learns
// its behavioral class on the first gesture (the verify-timeout skip plus
// the terminal-class marker) and the SECOND gesture rides the learned
// class to done+level-2; an already-classified context answers on the
// first. The returned flag reports the observed learning transition (the
// row's observable). A second gesture that still fails to execute is an
// error — the handshake is the documented mechanism, never a retry loop.
func (s *stand) wezGestureToTerminalDone(ctx context.Context, caseName string) (bool, error) {
	actionBase := s.countSub(actionDoubleMark)
	doneBase := s.countSub(correctionDoneMa)
	levelBase := s.countSub(termLevel2Mark)
	learnBase := s.countSub(wezLearnMark)
	skipBase := s.countSub(wezSkipVerifyMark)
	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return false, fmt.Errorf("%s: double gesture: %w", caseName, err)
	}
	if err := s.waitForNew(ctx, actionDoubleMark, actionBase+1, decisionWait); err != nil {
		return false, fmt.Errorf("%s: double-tap decision: %w", caseName, err)
	}
	learned := false
	deadline := time.Now().Add(correctionWait)
	for {
		if s.countSub(termLevel2Mark) > levelBase && s.countSub(correctionDoneMa) > doneBase {
			return learned, nil
		}
		if !learned && s.countSub(wezLearnMark) > learnBase && s.countSub(wezSkipVerifyMark) > skipBase {
			learned = true

			break
		}
		if time.Now().After(deadline) {
			return false, fmt.Errorf("%s: neither done+level-2 nor the verify-dead learning within %v",
				caseName, correctionWait)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return false, err
		}
	}
	// The learning gesture's round closed by the verify-timeout — the
	// SECOND gesture rides the learned class (exactly one, the mechanism).
	actionBase = s.countSub(actionDoubleMark)
	doneBase = s.countSub(correctionDoneMa)
	levelBase = s.countSub(termLevel2Mark)
	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return false, fmt.Errorf("%s: second gesture (learned class): %w", caseName, err)
	}
	if err := s.waitForNew(ctx, actionDoubleMark, actionBase+1, decisionWait); err != nil {
		return false, fmt.Errorf("%s: second double-tap decision: %w", caseName, err)
	}
	if err := s.waitForNew(ctx, correctionDoneMa, doneBase+1, correctionWait); err != nil {
		return false, fmt.Errorf("%s: correction done on the learned class: %w", caseName, err)
	}

	return learned, s.waitForNew(ctx, termLevel2Mark, levelBase+1, correctionWait)
}

// wezWordRound drives one word-correction round into the focused fixture:
// the typed text (gated by the key-trace), the gesture-to-done handshake
// and the line-ending keystroke that flushes the tty line buffer into the
// capture (venue mechanics: Enter rides ydotool type, never a named key).
func (s *stand) wezWordRound(ctx context.Context, caseName, text string) (bool, error) {
	keyBase := s.countSub(keyTraceMark)
	if err := s.injectText(ctx, text); err != nil {
		return false, fmt.Errorf("%s: type %q: %w", caseName, text, err)
	}
	if err := s.waitForNew(ctx, keyTraceMark, keyBase+minKeyEvents, keyWait); err != nil {
		return false, fmt.Errorf("%s: key visibility: %w", caseName, err)
	}
	learned, err := s.wezGestureToTerminalDone(ctx, caseName)
	if err != nil {
		return false, err
	}

	return learned, s.injectText(ctx, "\n")
}

// captureSize returns the capture file's current size (0 when absent) —
// the offset a round's delta oracle reads from.
func captureSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}

	return fi.Size()
}

// waitForCaptureDelta polls the capture until it grows past offset, then
// hexlifies the growth — the byte-exact delta oracle of one round (each
// round's line flushes whole at its newline; the arming probe is a
// self-erasing pair and never enters the line buffer).
func waitForCaptureDelta(ctx context.Context, path string, offset int64, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	var hexr string
	for {
		if data, err := os.ReadFile(path); err == nil && int64(len(data)) > offset {
			return hex.EncodeToString(data[offset:]), nil
		}
		if time.Now().After(deadline) {
			return hexr, fmt.Errorf("capture %s never grew past offset %d within %v", path, offset, timeout)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return hexr, err
		}
	}
}

// wezWordRow is one word-round row: the round body plus the byte-exact
// delta oracle. Pass/FAIL lands the row; a FAIL also fails the case (the
// exit contract).
func (s *stand) wezWordRow(
	ctx context.Context, caseName string, wf *wezFixtureRound, rowTag, text, wantHex string,
) error {
	offset := captureSize(wf.ptyPath)
	learned, err := s.wezWordRound(ctx, caseName, text)
	if err != nil {
		return err
	}
	hexr, cerr := waitForCaptureDelta(ctx, wf.ptyPath, offset, termCaptureWait)
	observed := fmt.Sprintf("venue: %s; learning=%v; pty_delta_hex=%s", venueMarker(ctx), learned, hexr)
	cmd := "armGate + type " + text + " + double Shift_R (learning handshake) + pty-delta oracle"
	if cerr == nil && hexr == wantHex {
		return wezRow(rowTag, cmd, observed,
			"Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона)")
	}
	if cerr == nil {
		cerr = fmt.Errorf("capture delta %s, want exactly %s", hexr, wantHex)
	}
	if rerr := wezRow(rowTag, cmd, observed, "Verdict: FAIL (pty-дельта не байт-точна)"); rerr != nil {
		return rerr
	}

	return cerr
}

// wezClipFrozenRow is the frozen CLIPBOARD row: a sentinel of the GS-SPIKE
// class rides CLIPBOARD through the whole correction round and wl-paste
// must read it back byte-identical — the terminal write branch never
// touches CLIPBOARD (noRung verify-after, ADR-008 D4/D5). The sentinel
// overwrite of the pre-state is documented in the row (D5).
func (s *stand) wezClipFrozenRow(ctx context.Context, caseName string, wf *wezFixtureRound) error {
	const rowTag = "clip-frozen"
	if err := clipCopy(ctx, primarySentinelMark); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	before := clipReadNoNewline(ctx)
	if before != primarySentinelMark {
		return fmt.Errorf("%s: CLIPBOARD pre-state %q, want sentinel %q", caseName, before, primarySentinelMark)
	}
	offset := captureSize(wf.ptyPath)
	learned, err := s.wezWordRound(ctx, caseName, corpusLatin)
	if err != nil {
		return err
	}
	hexr, cerr := waitForCaptureDelta(ctx, wf.ptyPath, offset, termCaptureWait)
	after := clipReadNoNewline(ctx)
	observed := fmt.Sprintf("venue: %s; learning=%v; clip_before=%s; clip_after=%s; pty_delta_hex=%s"+
		" (CLIPBOARD перезаписан сентинелом — синтетика, пред-стейт задокументирован затиранием)",
		venueMarker(ctx), learned, before, after, hexr)
	cmd := "printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline"
	if cerr == nil && hexr == corpusHexRU+"0a" && after == primarySentinelMark {
		return wezRow(rowTag, cmd, observed,
			"Verdict: pass (CLIPBOARD байт-нетронут терминальной коррекцией, слово исправлено)")
	}
	if cerr == nil && after != primarySentinelMark {
		cerr = fmt.Errorf("CLIPBOARD changed: %q, want sentinel %q", after, primarySentinelMark)
	}
	if rerr := wezRow(rowTag, cmd, observed,
		"Verdict: FAIL (клипборд изменён или слово не исправлено)"); rerr != nil {
		return rerr
	}

	return cerr
}

// wezLegProbeRow drives the OQ 3 primary-leg probe — a capability probe
// for the record (the OQ 2 owner verdict keeps the shipping mechanism
// buffer-trust WITHOUT the leg): a sentinel rides PRIMARY, the select
// chord (Home, then Shift+End — the phase-9 keyspecs) rides the
// debug-gated probe-forward channel of the daemon's own engine connection
// (ForwardKeyEvent is a signal only that connection can emit), and
// wl-paste --primary answers whether the forwarded chord moved the field
// content onto PRIMARY. green: the leg COULD read the field (the volume
// question still belongs to the OQ 2 verdict); red: the leg is not
// connectable — the boundary the OQ 2 verdict already documents.
func (s *stand) wezLegProbeRow(ctx context.Context, caseName, ctlBin string) error {
	const rowTag = "leg-probe"
	keyBase := s.countSub(keyTraceMark)
	if err := s.injectText(ctx, corpusLatin); err != nil {
		return fmt.Errorf("%s: probe token type: %w", caseName, err)
	}
	if err := s.waitForNew(ctx, keyTraceMark, keyBase+minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("%s: probe token visibility: %w", caseName, err)
	}
	if err := primarySentinel(ctx); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	forwardBase := s.countSub(probeForwardMark)
	if _, _, err := runCtl(ctx, ctlBin, "probe-forward", "Home"); err != nil {
		return fmt.Errorf("%s: probe-forward Home: %w", caseName, err)
	}
	if err := s.waitForNew(ctx, probeForwardMark, forwardBase+probeEmissionRecords, keyWait); err != nil {
		return fmt.Errorf("%s: Home emission: %w", caseName, err)
	}
	forwardBase = s.countSub(probeForwardMark)
	if _, _, err := runCtl(ctx, ctlBin, "probe-forward", "shift+End"); err != nil {
		return fmt.Errorf("%s: probe-forward shift+End: %w", caseName, err)
	}
	if err := s.waitForNew(ctx, probeForwardMark, forwardBase+probeEmissionRecords, keyWait); err != nil {
		return fmt.Errorf("%s: shift+End emission: %w", caseName, err)
	}
	if err := sleepCtx(ctx, chordSettle); err != nil {
		return err
	}
	primary := primaryReadNoNewline(ctx)
	observed := fmt.Sprintf("venue: %s; sentinel=%s; primary_after=%s; token_hex=%s",
		venueMarker(ctx), primarySentinelMark, primary, corpusHexLatin)
	cmd := "sentinel → goswitchctl probe-forward Home, shift+End → wl-paste --primary"
	if primary == corpusLatin || strings.HasPrefix(primary, corpusLatin) {
		return wezRow(rowTag, cmd, observed,
			"Verdict: leg-probe green (форварднутый чорд поднял выделение в PRIMARY — лег подключаем"+
				" технически; объём лега по-прежнему решает вердикт OQ 2 «contract»)")
	}

	return wezRow(rowTag, cmd, observed,
		"Verdict: leg-probe red (лег не подключён — буфер-доверие остаётся единственным механизмом"+
			" по вердикту OQ 2 «contract»; граница D7 задокументирована)")
}

// runTermWordWezterm drives the phase-10 wezterm word rows on the stand:
// the enabled-layer config daemon, one nightly fixture, the caps re-pin
// row, the two-greens word rounds, the mid-line contract row, the frozen
// CLIPBOARD sentinel and the leg-probe. A failure lands the honest-abort
// row and fails the case (exit contract).
func runTermWordWezterm(ctx context.Context, s *stand) error {
	const caseName = "term-word-wezterm"
	err := runTermWezRounds(ctx, s)
	if err == nil {
		fmt.Printf("%s: two green word rounds, mid-line, clip-frozen and leg-probe rows green (rows in %s)\n",
			caseName, spikeJournal)

		return nil
	}
	obs := "case aborted: " + boundRowText(err.Error())
	if rerr := wezRow(caseName, "wezterm word round", obs, "Verdict: aborted"); rerr != nil {
		fmt.Fprintf(os.Stderr, "e2e: %s deferred journal row: %v\n", caseName, rerr)
	}

	return err
}

// runTermWezRounds is the row program of the wezterm case (extracted so
// the honest-abort wrapper above stays under the complexity budget): the
// two-greens word rounds ride ONE fixture (the learned class lives with
// the input context — the daily-usage shape the criterion describes).
func runTermWezRounds(ctx context.Context, s *stand) error {
	const caseName = "term-word-wezterm"
	if err := startAutocorrectConfigDaemon(ctx, s, autocorrectConfigDoc); err != nil {
		return err
	}
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}
	capsBase := s.countSub(wezCapsMark)
	wf, err := s.openWezFixture(ctx, caseName, "wez")
	if err != nil {
		return err
	}
	defer wf.close()
	if err := s.wezCapsRepin(ctx, capsBase); err != nil {
		return err
	}
	// Round A (the learning round on a fresh context) and round B (the
	// full repeat — the two-greens verdict needs BOTH green in one run).
	// The wants are HEX TEXT: the delta oracle hexlifies the capture, so
	// the line ending rides as the literal "0a".
	for _, row := range []struct{ tag, text, want string }{
		{"word-a", corpusLatin, corpusHexRU + "0a"},
		{"word-b", corpusLatin, corpusHexRU + "0a"},
		{"mid-line", termMidPrefix + corpusLatin, termMidPrefixHex + corpusHexRU + "0a"},
	} {
		if err := s.wezArmGate(ctx, caseName); err != nil {
			return err
		}
		if err := s.wezWordRow(ctx, caseName, wf, row.tag, row.text, row.want); err != nil {
			return err
		}
	}

	// Clip-frozen: the delivery proof first, then the sentinel, then the
	// round (the sentinel must ride THROUGH the round untouched).
	if err := s.wezArmGate(ctx, caseName); err != nil {
		return err
	}
	if err := s.wezClipFrozenRow(ctx, caseName, wf); err != nil {
		return err
	}

	return s.wezLegProbeRow(ctx, caseName, ctlBin)
}
