package main

// The phase-12 line sel-probe case (plan 12-01, TERM-03, OQ1 — the PHASE
// GATE): the read channel of the line correction (chord → CopyMode recipe →
// PRIMARY) has ZERO green live rows today — the leg-probe reds of 10-06
// measured FLAT chords (Home, shift+End) with NO recipe in the config, and
// the «recipe in the config + forwarded chord» pair was never probed
// (12-RESEARCH A1/A2/A3). Before any line machinery (plan 12-02) the probe
// lands three dated cells in the phase journal: (a) the PHYSICAL recipe
// chord on the wayland fixture (the recipe already lives there), (b) the
// same physical chord on the xim-line fixture — the owner's transport since
// 2026-10-24 plus the X11-PRIMARY → wl-paste --primary bridge (A3, Pitfall
// 11), and (c) the DAEMON-FORWARDED chord in the exact future-gesture order
// (the End pair first, then the shift+home pair through probe-forward) —
// the self-trigger question. Cell (c) is the phase gate: green opens the
// line machine of 12-02; red STOPS the phase for the owner's decision —
// the buffer-instead-of-selection fallback is FORBIDDEN by D-01 (the
// rejected VTE-primitive in spirit; SC3 loses its anchor; ADR-008 D3 stays
// closed) and is neither offered nor executed. Privacy (D-20/D-21,
// T-12-01-01): the selection may carry secrets the daemon never typed, so
// its content NEVER enters a row or a log — rows carry the sentinel mark,
// the change FACT, the length and the latency only. Every cell freezes a
// CLIPBOARD sentinel through the round (the clip-frozen class) and documents
// the PRIMARY pre-state's fate (the sentinel overwritten by the selection,
// D5). A single honest row is verdict-bearing (the leg-probe form of
// 10-06 — no ×2 discipline for a probe); a FAIL/abort lands the
// honest-abort row, never a retry loop.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	secLineProbe = "Фаза 12 сел-проба канала чтения (term-line-probe)"

	// weztermXimLineFixture is the xim body (enable_wayland=false — the
	// owner's transport) plus the CopyMode Shift+Home recipe byte-identical
	// to the wayland fixture — the instrument of cells (b) and (c) (Pitfall
	// 11: a plan without an xim row is forbidden).
	weztermXimLineFixture = "gs-e2e-wezterm-xim-line.lua"
)

// lineProbeRow appends one dated verdict row to the case's journal section
// (the appendSpikeRow form; the label carries the row tag, the time the
// date).
func lineProbeRow(label, cmd, observed, verdict string) error {
	return appendSpikeRow(secLineProbe, label+" "+time.Now().Format("15:04:05"), cmd, observed, verdict)
}

// openLineFixture acquires one probe round: the stand-owned config spawn
// (--config-file TOP-LEVEL form), the pane-ready proof and the daemon
// focus_in gate above the spawn baseline — the openWezFixture form with the
// fixture parameterized (this family's three cells ride three different
// configs; the owner's ~/.wezterm.lua is never touched, Pitfall 7).
func (s *stand) openLineFixture(
	ctx context.Context, caseName, tag, fixture, clickLabel string,
) (*wezFixtureRound, error) {
	readyPath := filepath.Join(s.tmpDir, tag+"-ready")
	ptyPath := filepath.Join(s.tmpDir, tag+"-in")
	focusBase := s.focusBaseline()
	wcap, err := startWeztermCapture(ctx, weztermFixturePath(fixture), clickLabel,
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
	if werr := s.waitFocusInFrom(ctx, caseName, tag+" pane", focusBase); werr != nil {
		wf.close()

		return nil, fmt.Errorf("%s: %w", caseName, werr)
	}

	return wf, nil
}

// lineProbeTypeCorpus types the synthetic corpus into the pane and proves
// its delivery by the daemon key-trace (the wezLegProbeRow visibility form):
// the corpus is the line CONTENT the recipe selects — without it an empty
// selection would make the PRIMARY read ambiguous between «chord dead» and
// «selected nothing».
func (s *stand) lineProbeTypeCorpus(ctx context.Context, caseName string) error {
	keyBase := s.countSub(keyTraceMark)
	if err := s.injectText(ctx, corpusLatin); err != nil {
		return fmt.Errorf("%s: probe token type: %w", caseName, err)
	}

	return s.waitForNew(ctx, keyTraceMark, keyBase+minKeyEvents, keyWait)
}

// lineProbeSentinels freezes the two pre-states every cell needs: the
// CLIPBOARD sentinel (must ride the whole cell untouched — the clip-frozen
// class; the selection path may only ever touch PRIMARY) and the PRIMARY
// sentinel (the chord's change target). A CLIPBOARD pre-state that is not
// the sentinel is a harness defect — the change-fact verdict would ride an
// unknown pre-state, so the cell refuses instead.
func lineProbeSentinels(ctx context.Context, caseName string) error {
	if err := clipCopy(ctx); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	if before := clipReadNoNewline(ctx); before != primarySentinelMark {
		return fmt.Errorf("%s: CLIPBOARD pre-state %q, want sentinel %q", caseName, before, primarySentinelMark)
	}

	return primarySentinel(ctx)
}

// linePrimaryChanged is the change FACT of one cell: PRIMARY left both the
// empty state and the sentinel state — the recipe lifted real content. The
// content itself never leaves this function's caller as data.
func linePrimaryChanged(primary string) bool {
	return primary != "" && primary != primarySentinelMark
}

// lineProbeFacts renders the privacy-safe observable of one cell
// (D-20/D-21, T-12-01-01): the sentinel mark, the change fact, the length,
// the clip-frozen fact and the latency — never the selection content.
func lineProbeFacts(venue, primary string, changed bool, latency time.Duration, clipFrozen bool) string {
	return fmt.Sprintf(
		"venue: %s; sentinel=%s; primary_changed=%v; primary_len=%d; clip_frozen=%v; latency=%v"+
			" (содержимое выделения в ряд не попадает — D-20/D-21; PRIMARY пред-стейт — сентинел,"+
			" затёртие выделением задокументировано — D5)",
		venue, primarySentinelMark, changed, len(primary), clipFrozen, latency.Round(time.Millisecond))
}

// lineProbePhysicalCell runs one PHYSICAL chord cell — (a) on the wayland
// fixture (the recipe's own ground), (b) on the xim-line fixture (the
// owner's transport plus the X11-PRIMARY bridge, A3): the arming ladder
// warms the uinput channel (the 10-06 churn finding — the first ydotool
// call of a fresh fixture can be silently eaten), the corpus types the line
// content, the sentinels freeze both pre-states, the ydotool chord presses
// the recipe, the settle and the PRIMARY read-back answer the change FACT.
// The verdict is honest in both branches and never synthesized from
// expectations.
func (s *stand) lineProbePhysicalCell(
	ctx context.Context, caseName, tag, rowTag, fixture, clickLabel string,
) (bool, error) {
	wf, err := s.openLineFixture(ctx, caseName, tag, fixture, clickLabel)
	if err != nil {
		return false, err
	}
	defer wf.close()
	if err := s.wezArmGate(ctx, caseName); err != nil {
		return false, err
	}
	if err := s.lineProbeTypeCorpus(ctx, caseName); err != nil {
		return false, err
	}
	if err := lineProbeSentinels(ctx, caseName); err != nil {
		return false, err
	}
	chordStart := time.Now()
	if err := s.injectKeys(ctx, "shift+home"); err != nil {
		return false, fmt.Errorf("%s: %s: recipe chord: %w", caseName, rowTag, err)
	}
	if err := sleepCtx(ctx, chordSettle); err != nil {
		return false, err
	}
	primary := primaryReadNoNewline(ctx)
	changed := linePrimaryChanged(primary)
	clipFrozen := clipReadNoNewline(ctx) == primarySentinelMark
	observed := lineProbeFacts(venueMarker(ctx), primary, changed, time.Since(chordStart), clipFrozen)
	verdict := "Verdict: " + rowTag + " red — CopyMode-рецепт выделения не дал (sentinel не сменился)"
	if changed {
		verdict = "Verdict: " + rowTag + " green — CopyMode-рецепт выделяет (PRIMARY сменился)"
	}

	return changed, lineProbeRow(rowTag, "ydotool shift+home (рецепт CopyMode в "+fixture+")", observed, verdict)
}

// lineProbeForwardPair emits one keyspec through the daemon's probe verb and
// proves the emission pair landed in the daemon journal (the verb's
// contract — the wezLegProbeRow form; every forward IS a press+release
// pair, amendment (7): press-only arms the client's autorepeat).
func (s *stand) lineProbeForwardPair(ctx context.Context, caseName, ctlBin, spec string) error {
	forwardBase := s.countSub(probeForwardMark)
	if _, _, err := runCtl(ctx, ctlBin, "probe-forward", spec); err != nil {
		return fmt.Errorf("%s: probe-forward %s: %w", caseName, spec, err)
	}

	return s.waitForNew(ctx, probeForwardMark, forwardBase+probeEmissionRecords, keyWait)
}

// lineProbeForwardCell runs the SELF-TRIGGER cell (c) — the phase gate
// (OQ1): a fresh xim-line round (the future gesture runs on the owner's
// transport), the key-trace arming ladder, the corpus, the sentinels, then
// the daemon-forwarded chord in the EXACT future-gesture order — the End
// pair first (the cursor normalization of Pattern 3), then the shift+home
// pair (the ProbeForward pair form) — and the PRIMARY read-back. green: the
// self-trigger is alive (the one-triple-Shift gesture is reachable); red:
// the daemon cannot raise the selection itself — an owner decision (Rule 4
// class), never a code fallback: D-01 forbids the
// buffer-instead-of-selection primitive (the rejected VTE-primitive in
// spirit; ADR-008 D3 stays closed).
func (s *stand) lineProbeForwardCell(ctx context.Context, caseName, ctlBin string) (bool, error) {
	const rowTag = "сел-проба c (форвард)"
	wf, err := s.openLineFixture(ctx, caseName, "line-c", weztermXimLineFixture,
		"gs-e2e term-line-probe (xim-line c)")
	if err != nil {
		return false, err
	}
	defer wf.close()
	if err := s.wezArmGate(ctx, caseName); err != nil {
		return false, err
	}
	if err := s.lineProbeTypeCorpus(ctx, caseName); err != nil {
		return false, err
	}
	if err := lineProbeSentinels(ctx, caseName); err != nil {
		return false, err
	}
	chordStart := time.Now()
	// The exact future-gesture sequence (Pattern 1): the End pair BEFORE
	// the chord — each member rides its own press+release pair.
	if err := s.lineProbeForwardPair(ctx, caseName, ctlBin, "End"); err != nil {
		return false, err
	}
	if err := s.lineProbeForwardPair(ctx, caseName, ctlBin, "shift+home"); err != nil {
		return false, err
	}
	if err := sleepCtx(ctx, chordSettle); err != nil {
		return false, err
	}
	primary := primaryReadNoNewline(ctx)
	changed := linePrimaryChanged(primary)
	clipFrozen := clipReadNoNewline(ctx) == primarySentinelMark
	observed := lineProbeFacts(venueMarker(ctx), primary, changed, time.Since(chordStart), clipFrozen)
	verdict := "Verdict: " + rowTag + " red — самотриггер выделения мёртв (форварднутый чорд рецепт не поднял)"
	if changed {
		verdict = "Verdict: " + rowTag + " green — самотриггер выделения жив (форварднутый чорд поднял PRIMARY)"
	}

	return changed, lineProbeRow(rowTag,
		"sentinel → goswitchctl probe-forward End → probe-forward shift+home → wl-paste --primary",
		observed, verdict)
}

// lineProbeSummaryRow lands the summary row — the phase-gate fork stated
// FROM THE OBSERVED CELLS (never expectations): a green (c) opens the line
// machine of 12-02 (red a/b at a green c document the transport boundary —
// the channel is not proven on every transport; XIM UAT stays a separate
// acceptance); a red (c) stops the phase for the owner's decision — the
// buffer fallback is forbidden by D-01 and is neither offered nor executed.
func lineProbeSummaryRow(ctx context.Context, greenA, greenB, greenC bool) error {
	observed := fmt.Sprintf(
		"venue: %s; (a) физика-wayland=%v; (b) физика-xim=%v; (c) форвард=%v (только наблюдаемые ячеек)",
		venueMarker(ctx), greenA, greenB, greenC)
	verdict := "Verdict: сел-проба итог — (c) green: сел-лег самотриггера доказан," +
		" гейт фазы 12 открыт (машина строки 12-02)"
	if !greenC {
		verdict = "Verdict: сел-проба итог — (c) red: самотриггер мёртв — СТОП фазы, checkpoint владельцу" +
			" (D-01: фолбэк «буфер вместо выделения» запрещён — отвергнутый VTE-примитив по духу," +
			" SC3 теряет якорь, ADR-008 D3 не переоткрывается)"
	} else if !greenA || !greenB {
		verdict += "; граница: канал доказан не на всех транспортах (UAT на XIM — отдельная приёмка)"
	}

	return lineProbeRow("сел-проба итог",
		"три ячейки: физика-wayland / физика-xim / форвард (End-пара + shift+home-пара через probe-forward)",
		observed, verdict)
}

// runLineProbeCells is the probe's row program (extracted to keep the
// honest-abort wrapper small): the ctl client, the three cells in the
// plan's fixed order, then the summary row.
func runLineProbeCells(ctx context.Context, s *stand) error {
	const caseName = "term-line-probe"
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}
	greenA, err := s.lineProbePhysicalCell(ctx, caseName, "line-a",
		"сел-проба a (физика-wayland)", weztermWaylandFixture, "gs-e2e term-line-probe (wayland a)")
	if err != nil {
		return err
	}
	greenB, err := s.lineProbePhysicalCell(ctx, caseName, "line-b",
		"сел-проба b (физика-xim)", weztermXimLineFixture, "gs-e2e term-line-probe (xim-line b)")
	if err != nil {
		return err
	}
	greenC, err := s.lineProbeForwardCell(ctx, caseName, ctlBin)
	if err != nil {
		return err
	}

	return lineProbeSummaryRow(ctx, greenA, greenB, greenC)
}

// runTermLineProbe drives the phase-12 sel-probe on the stand: the plain
// -debug stand daemon (the probe reads its journal; probe-forward is
// debug-gated and fails closed without -debug), three cells in the fixed
// order and the summary row. A cell verdict — green OR red — is a completed
// probe; only a harness failure aborts (the honest-abort row and the exit
// contract, never a retry loop).
func runTermLineProbe(ctx context.Context, s *stand) error {
	const caseName = "term-line-probe"
	err := runLineProbeCells(ctx, s)
	if err == nil {
		fmt.Printf("%s: three cells + summary row landed (rows in %s)\n", caseName, spikeJournal)

		return nil
	}
	obs := "case aborted: " + boundRowText(err.Error())
	if rerr := lineProbeRow(caseName, "line sel-probe cell", obs, "Verdict: aborted"); rerr != nil {
		fmt.Fprintf(os.Stderr, "e2e: %s deferred journal row: %v\n", caseName, rerr)
	}

	return err
}
