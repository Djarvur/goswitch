package main

// The wezterm spike case family (plan 09-02, TERM-01): live-session probe
// cases that execute the phase's evidence program and append dated verdict
// rows to docs/adr/adr-008-spike-log.md (the d01-log precedent: the case
// owns the appends; verdict CONTENT is data, never a pass criterion — a
// probe that EXECUTED and recorded its observation is a green case).
//
// Manual activation (2026-10-09 protocol, owner-authorized): the session's
// map-time focus grant is broken (the journal's environment block), so
// every fixture window is focused by the OWNER'S CLICK — the production
// gesture. Every injection is gated on a focus marker (the AT-SPI witness
// or a fresh daemon focus_in pair; armGate stays the injection-side proof)
// and never fires into an unfocused surface; every journal row carries
// activation: manual-click.

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// Spike identifiers: the journal the cases append to (Evidence-правило
// ADR-008: строки журнала, не URL) and the journal sections each case
// family appends to (the header created them empty).
const (
	spikeJournal = "docs/adr/adr-008-spike-log.md"

	secSmoke   = "Дым пробы"
	secHazards = "A3 и hazard-строки"

	// The journal sections the spike case family appends to (the header
	// created them empty).
	secLadder = "Лестница первопричины P0–P5 (wezterm, базовый транспорт)"
	secRecipe = "Проба нативно-Wayland рецепта"
	secMatrix = "Вердикт-матрица первопричины"
	secCaps   = "Capability-таблица терминалов"
	secOracle = "Оракул wezterm-flip"

	// activationMark prefixes every row's Наблюдаемое cell — the
	// evidence-honesty marker of the manual-activation protocol (a verdict
	// produced on a clicked surface, never on an auto-granted one).
	activationMark = "activation: manual-click"
)

// runSpikeSmoke proves the probe channel end to end on the live desktop
// (Task 1's tracer slice): the stand daemon under -debug owns the ctl name,
// a dedicated wezterm window captures a pty, the built goswitchctl drives
// probe-forward "x" through the daemon, and the case asserts the EMISSION
// (the verb's contract) and records the delivery truth (pty hex) as a
// journal datum — delivered AND not-delivered are both green outcomes.
func runSpikeSmoke(ctx context.Context, s *stand) error {
	const caseName = "spike-smoke"
	// Registered first → runs last (the switch-spike defer order): the
	// journal row lands even when a later step fails (the row names the
	// abortion as the observed).
	verdicted := false
	defer func() {
		if verdicted {
			return
		}
		if err := appendSpikeRow(secSmoke, caseName+" "+time.Now().Format("15:04:05"),
			"gather smoke evidence", "case aborted before the verdict", "Verdict: aborted"); err != nil {
			fmt.Fprintf(os.Stderr, "e2e: %s deferred journal row: %v\n", caseName, err)
		}
	}()

	// The single-instance window: unit stopped when it was active, the
	// reversible restore registered (gsettings + stand daemon + unit).
	unit, err := openSpikeWindow(ctx, caseName)
	if err != nil {
		return err
	}
	defer func() { s.restoreSpikeWindow(ctx, unit) }()

	ctlBin, err := startSpikeDaemon(ctx, s, caseName)
	if err != nil {
		return err
	}
	surf, err := s.openSmokeCapture(ctx, caseName)
	if err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	defer surf.close()

	// A3 (indirect form, xim fixture): the spawn with --config-file came up
	// and took the owner's click — a rejected config path would have failed
	// the spawn.
	if err := appendSpikeRow(secHazards, "A3 indirect "+time.Now().Format("15:04:05"),
		"wezterm start --config-file "+weztermXimFixture,
		"spawn up and owner-clicked; the option did not refuse the path"+
			" (direct WAYLAND_DISPLAY re-probe belongs to the wayland fixture run)",
		"Verdict: config-file accepted (indirect)"); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}

	if err := s.armGate(ctx, caseName); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	hex, err := probeSmokeDelivery(ctx, s, ctlBin, surf.ptyPath)
	if err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	verdict := "Verdict: not-delivered (эмиссия была, клиент не получил)"
	if hex == "78" {
		verdict = "Verdict: delivered (hex байта 'x' в pty)"
	}
	verdicted = true

	return appendSpikeRow(secSmoke, caseName+" "+time.Now().Format("15:04:05"),
		"goswitchctl probe-forward x",
		"emitted=true; pty_hex="+hex,
		verdict)
}

// openSmokeCapture acquires the smoke's fixture surface through the full
// manual-activation ladder: the pane program must RUN (the readiness marker
// — A3's PROG-passing proof; the 21:22 live finding: a rejected flag shape
// means NO pane, NO window, NO verdict), the map-time auto-grant datum is
// observed BEFORE any click (2026-10-09 protocol item 3, spawn mode only —
// an owner-launched window's map moment is not observable), and the
// fixture is focused by the OWNER'S CLICK (the caps-0x9 fingerprint gate).
func (s *stand) openSmokeCapture(ctx context.Context, caseName string) (*surfaceHandle, error) {
	afterReady := func() error {
		return s.observeWeztermAutoGrant(ctx, caseName)
	}

	return s.openFixture(ctx, caseName, weztermFixturePath(weztermXimFixture),
		"КЛИК 1 из 6 — spike-smoke", caseName, afterReady)
}

// probeSmokeDelivery drives the probe verb through the live daemon and
// reads the capture back: the EMISSION is asserted (two "probe forward"
// records above base — the verb's contract), the DELIVERY is returned as
// the pty hex datum (both outcomes are legal verdicts, never a failure).
func probeSmokeDelivery(ctx context.Context, s *stand, ctlBin, ptyPath string) (string, error) {
	emissionBase := s.countSub(probeForwardMark)
	if _, _, err := runCtl(ctx, ctlBin, "probe-forward", "x"); err != nil {
		return "", fmt.Errorf("probe-forward call: %w", err)
	}
	if err := s.waitForNew(ctx, probeForwardMark, emissionBase+probeEmissionRecords, keyWait); err != nil {
		return "", fmt.Errorf("probe emission journal: %w", err)
	}

	return readPtyHex(ptyPath)
}

// observeWeztermAutoGrant records the per-terminal map-time auto-grant
// datum (2026-10-09 protocol item 3: terminals are genuinely untested —
// one no-click, no-injection observation per run, wezterm first): the
// freshly mapped fixture either engages its IM context on its own within
// the budget or does not — both outcomes are journal data. The marker is
// the wezterm caps fingerprint (caps 0x9 — the context engagement).
func (s *stand) observeWeztermAutoGrant(ctx context.Context, caseName string) error {
	marker, err := s.watchWeztermAutoGrant(ctx, caseName, mapAutoGrantBudget)
	if err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	observed := "no auto-grant within " + mapAutoGrantBudget.String() + " (no caps-0x9 fingerprint above base)"
	verdict := "Verdict: map-time auto-grant absent (manual-click protocol required)"
	if marker != "" {
		observed = "auto-grant marker=" + marker + " within " + mapAutoGrantBudget.String()
		verdict = "Verdict: map-time auto-grant present (" + marker + ")"
	}

	return appendSpikeRow(secHazards, "map-auto-grant wezterm "+time.Now().Format("15:04:05"),
		"wezterm --config-file ... start --always-new-process (xim fixture), observation only",
		observed, verdict)
}

// openSpikeWindow applies the stand's single-instance rule (T-09-02-04): a
// live goswitchd unit owns the ctl name and the component names, so it is
// stopped before the stand daemon spawns. The returned unit snapshot feeds
// restoreSpikeWindow in the caller's teardown.
func openSpikeWindow(ctx context.Context, caseName string) (string, error) {
	unit := unitState(ctx)
	fmt.Printf("%s: unit goswitchd state %q\n", caseName, unit)
	if unit == unitActiveState {
		if _, err := runCmd(ctx, "systemctl", "--user", "stop", "goswitchd"); err != nil {
			return "", fmt.Errorf("%s: stop unit daemon: %w", caseName, err)
		}
	}

	return unit, nil
}

// startSpikeDaemon brings the stand's own daemon up inside the spike
// window: the ctl client is built first (the case drives it), the component
// name must be free, and the -debug daemon must reach its registered mark
// (the switch-spike spawn ladder).
func startSpikeDaemon(ctx context.Context, s *stand, caseName string) (string, error) {
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return "", err
	}
	if err := waitNameFree(ctx); err != nil {
		return "", fmt.Errorf("%s: %w", caseName, err)
	}
	if err := s.startDaemonRegistered(ctx); err != nil {
		return "", fmt.Errorf("%s: %w", caseName, err)
	}

	return ctlBin, nil
}

// appendSpikeRow inserts one verdict row at the END of the named journal
// section (the d01-log precedent, section-scoped: the journal carries
// several tables and each case family appends only its own). Every row
// carries the activation: manual-click marker (the 2026-10-09 protocol's
// evidence-honesty rule).
func appendSpikeRow(section, probe, cmd, observed, verdict string) error {
	data, err := os.ReadFile(spikeJournal)
	if err != nil {
		return fmt.Errorf("open journal %s: %w", spikeJournal, err)
	}
	text := string(data)
	head := strings.Index(text, "## "+section)
	if head < 0 {
		return fmt.Errorf("journal %s has no section %q", spikeJournal, section)
	}
	headerEnd := strings.Index(text[head:], "\n")
	if headerEnd < 0 {
		return fmt.Errorf("journal %s section %q is malformed", spikeJournal, section)
	}
	headerEnd += head + 1
	tail := strings.Index(text[head:], "\n## ")
	if tail < 0 {
		tail = len(text) - head
	}

	// Skip backwards over the section's trailing blank lines so the row
	// lands directly under the last line of its table — but never before
	// the section header's own newline (an empty section appends under the
	// header).
	insert := head + tail
	for insert > headerEnd && (text[insert-1] == '\n' || text[insert-1] == ' ' || text[insert-1] == '\t') {
		insert--
	}
	observed = strings.NewReplacer("|", "/").Replace(activationMark + "; " + observed)
	row := fmt.Sprintf("| %s | %#q | %s | %s |\n", probe, cmd, observed, verdict)
	out := text[:insert] + "\n" + row + text[insert:]
	//nolint:gosec // G703 false positive: the journal path is a repo-relative constant
	if werr := os.WriteFile(spikeJournal, []byte(out), probeFilePerm); werr != nil {
		return fmt.Errorf("append journal row: %w", werr)
	}

	return nil
}

// appendMatrixRow writes one verdict-matrix row (P0..P4 → вердикт) — the
// section's own column shape.
func appendMatrixRow(p0, p1, p2, p3, p4, verdict string) error {
	data, err := os.ReadFile(spikeJournal)
	if err != nil {
		return fmt.Errorf("open journal %s: %w", spikeJournal, err)
	}
	mark := "## " + secMatrix
	text := string(data)
	head := strings.Index(text, mark)
	if head < 0 {
		return fmt.Errorf("journal %s has no section %q", spikeJournal, secMatrix)
	}
	// The matrix table is the section's only table; the row goes under its
	// last line.
	headEnd := head + len(mark)
	insert := len(text)
	if tail := strings.Index(text[headEnd:], "\n## "); tail >= 0 {
		insert = headEnd + tail
	}
	for insert > headEnd && (text[insert-1] == '\n' || text[insert-1] == ' ' || text[insert-1] == '\t') {
		insert--
	}
	row := fmt.Sprintf("\n| %s | %s | %s | %s | %s | %s |\n", p0, p1, p2, p3, p4, verdict)
	out := text[:insert] + row + text[insert:]
	//nolint:gosec // G703 false positive: the journal path is a repo-relative constant
	if werr := os.WriteFile(spikeJournal, []byte(out), probeFilePerm); werr != nil {
		return fmt.Errorf("append matrix row: %w", werr)
	}

	return nil
}

// runWeztermFlip is the two-round flip oracle (RESEARCH § E2E Oracle, the
// phase's RED instrument on TERM-01) — and, under the minimal-click batch
// re-plan (2026-10-10), it also carries the P0–P4 ladder instruments: each
// round is a fresh fixture spawn clicked by the owner, the round's flip →
// type → pty hex is both the oracle datum and the ladder's ground truth,
// and round 2 additionally drives the select-chord probes (sentinel
// discipline, both channels). Verdicts: both rounds hex(привет) → fixed;
// no round hex(привет) → defect-reproduced (exit 0 — the honest RED);
// mixed → flaky (exit 1). The -wezterm-config flag re-targets the fixture
// without touching the registry.
func runWeztermFlip(ctx context.Context, s *stand) error {
	const caseName = "wezterm-flip"
	verdicted := false
	defer func() {
		if verdicted {
			return
		}
		if err := appendSpikeRow(secOracle, caseName+" "+time.Now().Format("15:04:05"),
			"flip rounds", "case aborted before the verdict", "Verdict: aborted"); err != nil {
			fmt.Fprintf(os.Stderr, "e2e: %s deferred journal row: %v\n", caseName, err)
		}
	}()

	unit, err := openSpikeWindow(ctx, caseName)
	if err != nil {
		return err
	}
	defer func() { s.restoreSpikeWindow(ctx, unit) }()

	ctlBin, err := startSpikeDaemon(ctx, s, caseName)
	if err != nil {
		return err
	}

	fixture := weztermXimFixture
	if s.cfg.weztermConfig != "" {
		fixture = s.cfg.weztermConfig
	}

	green, latin := 0, 0
	for round := 1; round <= flipRounds; round++ {
		hex, flipOK, keys, commits, err := s.flipRound(ctx, caseName, ctlBin, fixture, round)
		if err != nil {
			return err
		}
		switch hex {
		case corpusHexRU:
			green++
		case corpusHexLatin:
			latin++
		}
		if err := appendFlipRoundRow(caseName, round, flipOK, keys, commits, hex); err != nil {
			return err
		}
	}

	verdicted = true

	return finalFlipVerdict(caseName, green)
}

// appendFlipRoundRow records one oracle round row.
func appendFlipRoundRow(caseName string, round int, flipPair bool, keys, commits int, hex string) error {
	verdict := "defect-reproduced (Latin-байты корпуса)"
	if hex == corpusHexRU {
		verdict = "fixed (hex привет)"
	}

	return appendSpikeRow(secOracle,
		fmt.Sprintf("%s раунд %d %s", caseName, round, time.Now().Format("15:04:05")),
		"flip → type ghbdtn → pty hex",
		fmt.Sprintf("flip_pair=%v; key_records=%d; commits=%d; pty_hex=%s", flipPair, keys, commits, hex),
		"Verdict: "+verdict)
}

// finalFlipVerdict records the oracle's closing verdict row from the round
// tally: both green → fixed; none green → defect-reproduced (exit 0 — the
// honest RED); mixed → flaky.
func finalFlipVerdict(caseName string, green int) error {
	var observed, verdict string
	switch green {
	case flipRounds:
		fmt.Printf("%s: verdict: fixed\n", caseName)
		observed, verdict = "оба pty_hex==привет", "verdict: fixed"
	case 0:
		fmt.Printf("%s: verdict: defect-reproduced\n", caseName)
		observed = "ни одного pty_hex==привет (латиница/пусто — дефект воспроизведён, честный RED фазы)"
		verdict = "verdict: defect-reproduced"
	default:
		fmt.Printf("%s: verdict: flaky\n", caseName)
		observed, verdict = "раунды разошлись", "verdict: flaky"
	}

	return appendSpikeRow(secOracle, "wezterm-flip итог "+time.Now().Format("15:04:05"),
		"два раунда", observed, verdict)
}

// roundOneObservations records the round-1-only data: the auto-grant datum
// and the P0 transport-identity probe (both observation-only, no clicks).
func (s *stand) roundOneObservations(ctx context.Context, caseName string) error {
	if err := s.observeWeztermAutoGrant(ctx, caseName); err != nil {
		return err
	}

	return s.probeP0Row(ctx, caseName)
}

// flipRound executes one oracle round: fresh fixture spawn → pane ready →
// (round 1: the auto-grant datum + the P0 transport probe) → owner click →
// arm gate (the deafness control) → flip (P2) → type (P1) → pty teardown →
// hex (P4), with the CommitText wire monitor (P3) spanning the type. Round
// 2 adds the select-chord probes before teardown.
func (s *stand) flipRound(
	ctx context.Context,
	caseName, ctlBin, fixture string,
	round int,
) (hex string, flipPair bool, keys, commits int, err error) {
	tag := fmt.Sprintf("%s-r%d", caseName, round)
	var afterReady func() error
	if round == 1 && !s.cfg.ownerLaunch {
		afterReady = func() error { return s.roundOneObservations(ctx, caseName) }
	}
	surf, err := s.openFixture(ctx, caseName, weztermFixturePath(fixture),
		fmt.Sprintf("КЛИК %d из 6 — wezterm-flip раунд %d", round+1, round),
		tag, afterReady)
	if err != nil {
		return "", false, 0, 0, fmt.Errorf("%s: %w", caseName, err)
	}
	defer surf.close()
	if err := s.armGate(ctx, caseName); err != nil {
		return "", false, 0, 0, fmt.Errorf("%s: %w", caseName, err)
	}

	monitor, err := startCommitMonitor()
	if err != nil {
		return "", false, 0, 0, fmt.Errorf("%s: %w", caseName, err)
	}
	defer monitor.close()

	flipPair, keys, err = s.driveFlipRound(ctx, caseName, ctlBin, round == flipRounds)
	if err != nil {
		return "", flipPair, keys, 0, fmt.Errorf("%s: %w", caseName, err)
	}
	commits = monitor.count()

	surf.close()
	hex, err = readPtyHex(surf.ptyPath)
	if err != nil {
		return "", flipPair, keys, commits, fmt.Errorf("%s: %w", caseName, err)
	}
	if round == 1 {
		if err := s.ladderRows(ctx, flipPair, keys, commits, hex); err != nil {
			return hex, flipPair, keys, commits, err
		}
	}

	return hex, flipPair, keys, commits, nil
}

// flipAndWaitPair taps the daemon's own flip (single Right Shift) and waits
// for the engine focus pair (P2). A missing pair is recorded, never fatal —
// the ladder's h2 class.
func (s *stand) flipAndWaitPair(ctx context.Context, caseName string) (bool, error) {
	baseOut, baseIn := s.countSub(flipPairOutEN), s.countSub(flipPairInRU)
	if err := s.injectKeys(ctx, "Shift_R"); err != nil {
		return false, fmt.Errorf("%s: flip tap: %w", caseName, err)
	}
	outErr := s.waitForNew(ctx, flipPairOutEN, baseOut+1, flipWait)
	inErr := s.waitForNew(ctx, flipPairInRU, baseIn+1, flipWait)
	pair := outErr == nil && inErr == nil
	fmt.Printf("%s: flip pair %v (out=%v in=%v)\n", caseName, pair, outErr == nil, inErr == nil)

	return pair, nil
}

// typeAndCountKeys injects the synthetic corpus and returns the key-trace
// record count above base (P1's delivery datum).
func (s *stand) typeAndCountKeys(ctx context.Context) (int, error) {
	base := s.countSub(keyTraceMark)
	if err := s.injectText(ctx, corpusLatin); err != nil {
		return 0, err
	}
	if err := sleepCtx(ctx, acTerminalSettle); err != nil {
		return 0, err
	}

	return s.countSub(keyTraceMark) - base, nil
}

// chordProbes drives the wezterm select-chord rows (the caps table's row 4)
// under the sentinel discipline (Pitfall 5): channel (a) the physical
// shift+home (the terminal's own selection semantics), channel (b) the
// daemon probe-forward (ForwardKeyEvent delivery — the #4615 Shift-lost
// signature question).
func (s *stand) chordProbes(ctx context.Context, caseName, ctlBin string) error {
	if err := primarySentinel(ctx); err != nil {
		return err
	}
	if err := s.injectKeys(ctx, "shift+home"); err != nil {
		return fmt.Errorf("%s: chord (a) shift+home: %w", caseName, err)
	}
	if err := sleepCtx(ctx, chordSettle); err != nil {
		return err
	}
	primaryA := primaryReadNoNewline(ctx)
	verdictA := "Verdict: выделения нет (sentinel не сменился)"
	if strings.HasPrefix(primaryA, corpusLatin) || strings.HasPrefix(primaryA, "привет") {
		verdictA = "Verdict: выделение есть (sentinel сменился на корпус)"
	}
	if err := appendSpikeRow(secCaps, "wezterm select-chord (a) "+time.Now().Format("15:04:05"),
		"ydotool shift+home после набора корпуса",
		"sentinel="+primarySentinelMark+"; primary_after="+primaryA, verdictA); err != nil {
		return err
	}

	if err := primarySentinel(ctx); err != nil {
		return err
	}
	if _, _, err := runCtl(ctx, ctlBin, "probe-forward", "shift+Home"); err != nil {
		return fmt.Errorf("%s: chord (b) probe-forward: %w", caseName, err)
	}
	if err := s.waitForLog(ctx, probeForwardMark, keyWait); err != nil {
		return fmt.Errorf("%s: chord (b) emission: %w", caseName, err)
	}
	if err := sleepCtx(ctx, chordSettle); err != nil {
		return err
	}
	primaryB := primaryReadNoNewline(ctx)
	verdictB := "Verdict: forwarded-chord выделения нет (sentinel не сменился; курсор-прыжок или глухота)"
	if strings.HasPrefix(primaryB, corpusLatin) || strings.HasPrefix(primaryB, "привет") {
		verdictB = "Verdict: forwarded-chord выделение есть (#4615 Shift не потерян)"
	}

	return appendSpikeRow(secCaps, "wezterm select-chord (b) "+time.Now().Format("15:04:05"),
		"goswitchctl probe-forward shift+Home",
		"sentinel="+primarySentinelMark+"; primary_after="+primaryB, verdictB)
}

// probeP0Row runs the P0 transport-identity probe (environ + xlsclients on
// the live fixture process) and records the row.
func (s *stand) probeP0Row(ctx context.Context, caseName string) error {
	waylandDisplay, display, xls, pids, err := probeTransportIdentity(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	observed := fmt.Sprintf("pids=%v; %s; %s; xlsclients=%q", pids, waylandDisplay, display, xls)
	verdict := "Verdict: XIM-client (WAYLAND_DISPLAY отсутствует, DISPLAY есть — гипотеза подтверждена)"
	if waylandDisplay != "" {
		verdict = "Verdict: native-wayland (WAYLAND_DISPLAY есть — XIM-гипотеза опровергнута," +
			" лестница идёт под h1/h2/h3)"
	}

	return appendSpikeRow(secLadder, "P0 "+time.Now().Format("15:04:05"),
		"/proc/<wezterm-gui>/environ + xlsclients", observed, verdict)
}

// driveFlipRound is the round's instrumented middle: the flip (P2), the
// typed corpus (P1), and — on the final round — the select-chord probes,
// closed by the settle budget. Returns the flip-pair flag and the key
// record count.
func (s *stand) driveFlipRound(
	ctx context.Context, caseName, ctlBin string, final bool,
) (flipPair bool, keys int, err error) {
	flipPair, err = s.flipAndWaitPair(ctx, caseName)
	if err != nil {
		return flipPair, 0, err
	}
	keys, err = s.typeAndCountKeys(ctx)
	if err != nil {
		return flipPair, keys, err
	}
	if final {
		if err := s.chordProbes(ctx, caseName, ctlBin); err != nil {
			return flipPair, keys, err
		}
	}

	return flipPair, keys, sleepCtx(ctx, acTerminalSettle)
}

// ladderRows records the round-1 ladder rows (P1/P2/P3/P4) and the verdict
// matrix row from the oracle's own data.
func (s *stand) ladderRows(ctx context.Context, flipPair bool, keys, commits int, hex string) error {
	_ = ctx
	keyVerdict := "Verdict: клавиши доставлены (keys+armGate)"
	if keys == 0 {
		keyVerdict = "Verdict: клавиши отсутствуют после settle" +
			" (h1-класс; демонский фолбэк без input-контекста невозможен)"
	}
	if err := appendSpikeRow(secLadder, "P1 "+time.Now().Format("15:04:05"),
		"ydotool type ghbdtn, счётчик key-записей",
		fmt.Sprintf("key_records=%d (armGate прошёл — подложка видит модификатор)", keys), keyVerdict); err != nil {
		return err
	}

	pairVerdict := "Verdict: пара есть (флип достиг демона)"
	if !flipPair {
		pairVerdict = "Verdict: пары нет (h2-класс: контекст wezterm не перечекался на новый движок)"
	}
	if err := appendSpikeRow(secLadder, "P2 "+time.Now().Format("15:04:05"),
		"тап Right Shift, ожидание focus_out(en)/focus_in(ru)",
		fmt.Sprintf("pair=%v", flipPair), pairVerdict); err != nil {
		return err
	}

	commitVerdict := "Verdict: коммиты есть (демонская сторона работает, разрыв в доставке клиенту — h3)"
	if commits == 0 {
		commitVerdict = "Verdict: коммитов нет при видимых клавишах (modeRU не достигнут / consume не случился)"
	}
	if err := appendSpikeRow(secLadder, "P3 "+time.Now().Format("15:04:05"),
		"dbus-monitor CommitText (только счётчик)",
		fmt.Sprintf("commit_signals=%d", commits), commitVerdict); err != nil {
		return err
	}

	hexClass := "латиница ghbdtn (" + corpusHexLatin + ")"
	if hex == corpusHexRU {
		hexClass = "кириллица привет (" + corpusHexRU + ")"
	}
	p4Verdict := "Verdict: посадка дефекта зафиксирована (Latin-байты — сигнатура дефекта)"
	if hex == corpusHexRU {
		p4Verdict = "Verdict: посадка фикса (pty == привет)"
	}

	if err := appendSpikeRow(secLadder, "P4 "+time.Now().Format("15:04:05"),
		"pty-грунт: hex-сверка захвата",
		"pty_hex="+hex+" ("+hexClass+")", p4Verdict); err != nil {
		return err
	}

	return appendMatrixRow("XIM", "keys", "pair", fmt.Sprintf("commits=%d", commits), "Latin",
		ladderVerdict(flipPair, keys, commits, hex))
}

// ladderVerdict classifies the round per the RESEARCH verdict matrix.
func ladderVerdict(flipPair bool, keys, commits int, hex string) string {
	switch {
	case hex == corpusHexRU:
		return "fixed — дефект не воспроизведён в этом раунде"
	case keys == 0:
		return "h1: XIM-контекст армирования мёртв (keys absent) — демонский фолбэк невозможен, честная ветка — рецепт"
	case !flipPair:
		return "h2: контекст wezterm не перечекался на новый движок — кандидат pre-arm/re-arm на focus_in"
	case commits > 0:
		return "XIM commit-delivery break (h3): коммиты есть, клиент не получил" +
			" — кандидаты pre-arm / forward-кириллицы (spike-probe)"
	default:
		return "consume-разрыв: клавиши и флип есть, коммитов нет — modeRU не достигнут"
	}
}

// recipeRound executes the wayland round's injection phase: owner click →
// arm gate → flip (P2') → type → the CopyMode chord observation (before
// teardown!) → pty read. Returns the flip-pair flag and the pty hex.
func (s *stand) recipeRound(
	ctx context.Context, s2 *stand, caseName, ptyPath string,
) (flipPair bool, hex string, err error) {
	clickErr := s2.waitWeztermClick(ctx, caseName, "wezterm wayland fixture")
	if clickErr != nil {
		return false, "", clickErr
	}
	if err := s2.armGate(ctx, caseName); err != nil {
		return false, "", err
	}
	fp, ferr := s2.flipAndWaitPair(ctx, caseName)
	if ferr != nil {
		return fp, "", ferr
	}
	if _, terr := s2.typeAndCountKeys(ctx); terr != nil {
		return fp, "", terr
	}
	if serr := sleepCtx(ctx, acTerminalSettle); serr != nil {
		return fp, "", serr
	}
	flipPair = fp
	// The CopyMode select-chord observation rides BEFORE teardown — the
	// window must still hold focus (primary record; the full sentinel round
	// lives in the flip case's chord probes).
	if err := recipeChordRow(ctx, s2, caseName); err != nil {
		return flipPair, "", err
	}
	hex, hexErr := readPtyHex(ptyPath)
	if hexErr != nil {
		return flipPair, "", hexErr
	}

	return flipPair, hex, nil
}

// recipeP0Row records the P0' transport re-probe row under the wayland
// fixture (the direct A3: WAYLAND_DISPLAY must appear).
func recipeP0Row(ctx context.Context, caseName string) error {
	waylandDisplay, display, xls, pids, err := probeTransportIdentity(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	observed := fmt.Sprintf("pids=%v; %s; %s; xlsclients=%q", pids, waylandDisplay, display, xls)
	p0Verdict := "Verdict: транспорт сменился (WAYLAND_DISPLAY есть — рецепт применился)"
	if waylandDisplay == "" {
		p0Verdict = "Verdict: A3-провал — транспорт НЕ сменился" +
			" (переход на --config enable_wayland=true; зафиксировать, какой путь сработал)"
	}

	return appendSpikeRow(secRecipe, "P0' "+time.Now().Format("15:04:05"),
		"/proc/<wezterm-gui>/environ под wayland-фикстурой", observed, p0Verdict)
}

// recipeChordRow drives the CopyMode select-chord observation and records
// the row (the window must still hold focus — the caller runs it before
// teardown).
func recipeChordRow(ctx context.Context, s *stand, caseName string) error {
	if err := primarySentinel(ctx); err != nil {
		return err
	}
	if err := s.injectKeys(ctx, "shift+home"); err != nil {
		return fmt.Errorf("%s: CopyMode chord: %w", caseName, err)
	}
	if err := sleepCtx(ctx, chordSettle); err != nil {
		return err
	}
	primary := primaryReadNoNewline(ctx)
	chordVerdict := "Verdict: CopyMode-рецепт выделения не дал (sentinel не сменился)"
	if primary != "" && primary != primarySentinelMark {
		chordVerdict = "Verdict: CopyMode-рецепт выделяет (PRIMARY сменился)"
	}

	return appendSpikeRow(secRecipe, "select-chord CopyMode "+time.Now().Format("15:04:05"),
		"ydotool shift+home (рецепт CopyMode в фикстуре)",
		"sentinel="+primarySentinelMark+"; primary_after="+primary, chordVerdict)
}

// runSpikeWezterm is the native-Wayland recipe trial (Pattern 2): one stand
// round on the wayland fixture — the P0' WAYLAND_DISPLAY re-probe (direct
// A3), the flip → type → pty round under the new transport, and the
// CopyMode select-chord observation. The owner's ~/.wezterm.lua is never
// touched; the stability ×2 round is deferred with reason (the batch's
// click budget).
func runSpikeWezterm(ctx context.Context, s *stand) error {
	const caseName = "spike-wezterm"
	verdicted := false
	defer func() {
		if verdicted {
			return
		}
		if err := appendSpikeRow(secRecipe, caseName+" "+time.Now().Format("15:04:05"),
			"рецепт-проба", "case aborted before the verdict", "Verdict: aborted"); err != nil {
			fmt.Fprintf(os.Stderr, "e2e: %s deferred journal row: %v\n", caseName, err)
		}
	}()

	unit, err := openSpikeWindow(ctx, caseName)
	if err != nil {
		return err
	}
	defer func() { s.restoreSpikeWindow(ctx, unit) }()

	if _, err := startSpikeDaemon(ctx, s, caseName); err != nil {
		return err
	}

	surf, err := s.openFixture(ctx, caseName, weztermFixturePath(weztermWaylandFixture),
		"КЛИК 4 из 6 — spike-wezterm (wayland)", "recipe", nil)
	if err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	defer surf.close()

	// P0' re-probe (direct A3): under enable_wayland=true the process
	// environ must carry WAYLAND_DISPLAY.
	if err := recipeP0Row(ctx, caseName); err != nil {
		return err
	}

	flipPair, hex, err := s.recipeRound(ctx, s, caseName, surf.ptyPath)
	if err != nil {
		return err
	}
	recipeVerdict := "Verdict: рецепт НЕ исправил (pty != привет — нативный IME 20240203 не работает," +
		" hazard #5125-класса; выбор ветки — владелец, чекпойнт 09-03)"
	if hex == corpusHexRU {
		recipeVerdict = "Verdict: рецепт работает (pty == привет под нативным Wayland)"
	}
	if err := appendSpikeRow(secRecipe, "P1'–P4' "+time.Now().Format("15:04:05"),
		"флип → набор ghbdtn → pty hex под нативным Wayland",
		fmt.Sprintf("flip_pair=%v; pty_hex=%s", flipPair, hex), recipeVerdict); err != nil {
		return err
	}
	verdicted = true

	return appendSpikeRow(secRecipe, "стабильность "+time.Now().Format("15:04:05"),
		"второй раунд в одном запуске",
		"deferred: клик-бюджет батча (минимально-честный прогон 09-02); two-greens остаётся контрактом 09-03",
		"Verdict: deferred-with-reason (не исполнено, помечено честно)")
}
