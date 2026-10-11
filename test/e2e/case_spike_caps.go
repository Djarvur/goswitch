package main

// The capability-table case for the second terminal (plan 09-02, TERM-01):
// one live gnome-terminal/VTE round under the manual-activation protocol —
// the injection rows the kill criterion and the Phase-10 gate need (the
// caps-table rows for WEZTERM are written by the wezterm-flip case's
// rounds, which own those surfaces and their data). Observation rows
// record what the session honestly yields; feasibility-gated deferrals are
// marked deferred-with-reason, never silently skipped.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Knobs of the TERM-04 negative round.
const (
	controlRetries = 3
	// secTerm04 names the journal section — the HEADER TEXT IS THE
	// CONTRACT (data owned by docs/adr/adr-008-spike-log.md, where the
	// section was created): a cosmetic code-side rename silently routes
	// the case's rows to a "no section" error and the evidence never
	// lands (the 2026-10-13 finding). TestSpikeJournalSectionsExist pins
	// it.
	secTerm04 = "TERM-04 негатив (attach-семейство, rev D-55 §7.2)"
)

// runSpikeCaps drives the VTE round: a fresh gnome-terminal window running
// the pty-capturing program, engaged by the daemon focus_in, then the
// injection round — delivery count, the flip, the commit-landing shortcut
// check (flip → RU → type → pty already привет? — the decisive cheap probe
// for the VTE word path), and the select-chord with the PRIMARY sentinel (the
// expected verdict: VTE has no keyboard selection → the Phase-12 kill
// criterion).
func runSpikeCaps(ctx context.Context, s *stand) error {
	const caseName = "spike-caps"
	verdicted := false
	defer func() {
		if verdicted {
			return
		}
		if err := appendSpikeRow(secCaps, "VTE итог "+time.Now().Format("15:04:05"),
			"VTE раунд", "case aborted before the verdict", "Verdict: aborted"); err != nil {
			fmt.Fprintf(os.Stderr, "e2e: %s deferred journal row: %v\n", caseName, err)
		}
	}()

	outcome, err := s.openVTERound(ctx, caseName)
	if err != nil {
		return err
	}
	if outcome != armWindow {
		// The arming row (desync or dead) is recorded by the arming; the
		// injection rounds are skipped (every keystroke would ride the stale
		// context or go nowhere) — the case still closes honestly with the
		// branch noted.
		if err := appendSpikeRow(secCaps, "VTE итог "+time.Now().Format("15:04:05"),
			"VTE раунд (ветка арминга)", "инъекционные раунды не исполнялись — арминг не доказал доставку в окно",
			"Verdict: arming branch (раунды пропущены, наблюдение записано)"); err != nil {
			return err
		}
		verdicted = true

		return nil
	}

	// Row 2 (ProcessKeyEvent delivery) and the flip for the landing round.
	flipPair, err := s.vteDeliveryAndFlip(ctx, caseName)
	if err != nil {
		return err
	}

	// Row 4 (select-chord → PRIMARY, channel (a) only — the kill-criterion
	// question): sentinel → shift+home → read.
	if _, cerr := s.vteChordRow(ctx, caseName); cerr != nil {
		return cerr
	}

	// The pty ground truth (the shortcut check datum) rides after the
	// chord; then the teardown sends the EOF so the window closes itself
	// (the pid-точечный discipline: no name-wide kills).
	if err := s.sendVTETeof(ctx, caseName); err != nil {
		return err
	}
	if err := sleepCtx(ctx, acTerminalSettle); err != nil {
		return err
	}
	hex, err := readPtyHexArmed(filepath.Join(s.tmpDir, caseName+"-in"))
	if err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	if err := s.vteLandingRow(flipPair, hex); err != nil {
		return err
	}

	// Row 1 (caps reporting): the observed SetCapabilities value for the
	// VTE context — recorded, never a decision key (Pitfall 3).
	capsVal := s.lastCapsValue() // "(none)" is the honest observed when no caps record landed
	if err := appendSpikeRow(secCaps, "VTE caps "+time.Now().Format("15:04:05"),
		"SetCapabilities в журнале демона для контекста VTE",
		"caps="+capsVal+" (0x29-класс объявляет surrounding — недоверие, Pitfall 3)",
		"Verdict: значение записано, решение на caps не ключуется"); err != nil {
		return err
	}
	verdicted = true

	return appendSpikeRow(secCaps, "VTE итог "+time.Now().Format("15:04:05"),
		"VTE раунд", "все пробы раунда исполнены, строки выше", "Verdict: executed")
}

// vteCapture holds the gnome-terminal client process for the teardown reap
// (the actual window lives in gnome-terminal-server; the pane's EOF closes
// it — the client process is only a zombie to reap).
type vteCapture struct {
	cmd *exec.Cmd
}

// startVTECapture spawns a fresh gnome-terminal window running the
// pty-capturing program (--window: never a tab in an owner window).
//
//nolint:noctx // the window must outlive the case context: the caller closes it
func startVTECapture(_ context.Context, readyPath, ptyPath string) (*vteCapture, error) {
	cmd := exec.Command(vteBin, "--window", "--",
		"bash", "-c", "echo ready > "+readyPath+"; cat > "+ptyPath)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s capture: %w", vteBin, err)
	}

	return &vteCapture{cmd: cmd}, nil
}

// close reaps the client process (the window itself closes via the pane's
// EOF — sendVTETeof).
func (v *vteCapture) close() {
	if v == nil || v.cmd == nil || v.cmd.Process == nil {
		return
	}
	_ = v.cmd.Process.Kill()
	_ = v.cmd.Wait()
	v.cmd = nil
}

// sendVTETeof delivers Ctrl+D to the focused pane (the ydotool 0.1.8 named
// sequence ctrl+d — the colon form typed "2332" instead, the live finding
// shared with the owner-fixture close) so the cat sees EOF and exits,
// closing the window by the plan's own discipline.
func (s *stand) sendVTETeof(ctx context.Context, caseName string) error {
	if _, err := runCmd(ctx, "ydotool", "key", "ctrl+d"); err != nil {
		return fmt.Errorf("%s: VTE EOF: %w", caseName, err)
	}

	return nil
}

// vteChordRow drives the VTE select-chord probe (sentinel → shift+home →
// PRIMARY read) and records the kill-criterion row. Returns the PRIMARY
// read for the caller's logs.
func (s *stand) vteChordRow(ctx context.Context, caseName string) (string, error) {
	if err := primarySentinel(ctx); err != nil {
		return "", err
	}
	if err := s.injectKeys(ctx, "shift+home"); err != nil {
		return "", fmt.Errorf("%s: VTE chord: %w", caseName, err)
	}
	if err := sleepCtx(ctx, chordSettle); err != nil {
		return "", err
	}
	primary := primaryReadNoNewline(ctx)
	killVerdict := "Verdict: выделения нет (sentinel не сменился) — kill-критерий Фазы 12: " +
		"строка сужается до wezterm, вердикт о сужении — владелец (чекпойнт 09-03)"
	if primary != "" && primary != primarySentinelMark {
		killVerdict = "Verdict: ВЫДЕЛЕНИЕ ЕСТЬ (sentinel сменился)" +
			" — kill-критерий НЕ срабатывает, VTE имеет выделение с клавиатуры"
	}
	err := appendSpikeRow(secCaps, "VTE select-chord (a) "+time.Now().Format("15:04:05"),
		"ydotool shift+home после набора корпуса",
		"sentinel="+primarySentinelMark+"; primary_after="+primary, killVerdict)

	return primary, err
}

// vteLandingRow records the commit-landing shortcut check row (the pty hex
// against the Cyrillic corpus).
func (s *stand) vteLandingRow(flipPair bool, hex string) error {
	shortcutVerdict := "Verdict: шорткат НЕ работает (pty != привет — словесная коррекция VTE требует полного канала)"
	if strings.Contains(hex, corpusHexRU) {
		shortcutVerdict = "Verdict: шорткат работает (pty содержит привет" +
			" — буфер+commit без канала чтения: VTE почти бесплатен)"
	}

	return appendSpikeRow(secCaps, "VTE посадка+шорткат "+time.Now().Format("15:04:05"),
		"флип → RU → набор ghbdtn → pty hex",
		fmt.Sprintf("flip_pair=%v; pty_hex=%s", flipPair, hex), shortcutVerdict)
}

// openVTERound opens the VTE round's window: unit window, stand daemon,
// the capture spawn, the pane-ready proof, the daemon focus_in gate (the
// unified detector) and the arming gate. A stale or dead arming records its
// row and is reported as outcome != armWindow (the caller records the
// branch) — never an error.
func (s *stand) openVTERound(ctx context.Context, caseName string) (armOutcome, error) {
	unit, err := openSpikeWindow(ctx, caseName)
	if err != nil {
		return armDead, err
	}
	defer func() { s.restoreSpikeWindow(ctx, unit) }()

	if _, err := startSpikeDaemon(ctx, s, caseName); err != nil {
		return armDead, err
	}

	ptyPath := filepath.Join(s.tmpDir, caseName+"-in")
	readyPath := filepath.Join(s.tmpDir, caseName+"-ready")
	focusBase := s.focusBaseline()
	vcap, err := startVTECapture(ctx, readyPath, ptyPath)
	if err != nil {
		return armDead, fmt.Errorf("%s: %w", caseName, err)
	}
	defer vcap.close()
	if err := waitPaneReady(ctx, caseName, readyPath); err != nil {
		return armDead, fmt.Errorf("%s: %w", caseName, err)
	}

	fmt.Printf("%s: awaiting daemon focus_in for the gnome-terminal fixture (map focus usually self-fires)\n", caseName)
	if err := s.waitFocusInFrom(ctx, caseName, "gnome-terminal window", focusBase); err != nil {
		return armDead, fmt.Errorf("%s: %w", caseName, err)
	}

	report, aerr := s.armGate(ctx, caseName, ptyPath)
	if aerr != nil {
		return armDead, fmt.Errorf("%s: %w", caseName, aerr)
	}
	if report.outcome != armWindow {
		if rerr := appendArmRow(secCaps, caseName,
			fmt.Sprintf("armGate probe (ydotool type %s + Enter)", armProbeText),
			report, ptyPath); rerr != nil {
			return report.outcome, rerr
		}

		return report.outcome, nil
	}

	return armWindow, nil
}

// vteDeliveryAndFlip runs the EN-mode delivery probe (row 2) and the flip
// that sets the landing round's RU mode. Returns the flip-pair flag.
func (s *stand) vteDeliveryAndFlip(ctx context.Context, caseName string) (bool, error) {
	keys, err := s.typeAndCountKeys(ctx)
	if err != nil {
		return false, err
	}
	deliveryVerdict := "Verdict: клавиши доставлены (движок видит набор VTE)"
	if keys == 0 {
		deliveryVerdict = "Verdict: клавиши НЕ доставлены (движок глух — диагностика армирования)"
	}
	if err := appendSpikeRow(secCaps, "VTE доставка "+time.Now().Format("15:04:05"),
		"ydotool type ghbdtn, счётчик key-записей",
		fmt.Sprintf("key_records=%d", keys), deliveryVerdict); err != nil {
		return false, err
	}

	// The flip feeds the commit-landing shortcut check (row 3): flip → RU →
	// type → the pty must read привет if the VTE word path is nearly free
	// (buffer+commit, no read channel).
	return s.flipAndWaitPair(ctx, caseName)
}

// t04Grace is the residual budget knob of the negative round's zenity
// control (each control try waits this long for the key-trace record).
const t04Grace = 60 * time.Second

// runTerm04Negative carries the TERM-04 frozen negative under the unified
// observation layer (rev D-55 §7.2: замороженные негативные ряды перенесены
// в дисциплину нового семейства). The layer is ENABLED (the frozen
// document); the deafness control is a stand-spawned zenity keystroke, the
// negative surface a stand-spawned plain wezterm (the owner's default
// config — the daily transport). ONE mode: no pastes, no clicks required —
// the daemon focus_in gates both surfaces. Oracles: zero fired records,
// zero correction records, status counters fired=0.
func runTerm04Negative(ctx context.Context, s *stand) error {
	const caseName = "term04-negative"
	verdicted := false
	defer func() {
		if verdicted {
			return
		}
		if err := appendSpikeRow(secTerm04, caseName+" "+time.Now().Format("15:04:05"),
			"негатив", "case aborted before the verdict", "Verdict: aborted"); err != nil {
			fmt.Fprintf(os.Stderr, "e2e: %s deferred journal row: %v\n", caseName, err)
		}
	}()

	// The ENABLED layer (the frozen document — active state, empty
	// blocklist): the stand daemon restarted WITH the autocorrect config.
	// The full lifecycle precedes the spawn (see startTerm04Daemon).
	if err := startTerm04Daemon(ctx, s, caseName); err != nil {
		return err
	}
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}

	// The control surface: a stand-spawned zenity entry (reaped by the
	// teardown), engaged by the unified focus detector.
	if err := s.spawnZenityControl(ctx, caseName); err != nil {
		return err
	}
	if err := s.controlWithRetries(ctx, caseName); err != nil {
		return err
	}

	// The negative surface: a stand-spawned plain wezterm (the owner's
	// daily transport — no stand config), engaged by the same detector; the
	// negative expects NO engagement of the autocorrect layer.
	if err := s.t04NegativeSurface(ctx, caseName); err != nil {
		return err
	}

	if err := s.injectText(ctx, acWord); err != nil {
		return err
	}
	if err := sleepCtx(ctx, acTerminalSettle); err != nil {
		return err
	}
	if err := s.assertAcSilent(caseName); err != nil {
		return err
	}
	if err := ctlStatusHas(ctx, ctlBin, "autocorrect_enabled=true", "autocorrect_fired=0"); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	// Best-effort pane EOF: the negative window's default shell exits, the
	// window closes itself (the pid-точечный discipline).
	_, _ = runCmd(ctx, "ydotool", "key", "ctrl+d")
	verdicted = true

	return appendSpikeRow(secTerm04, caseName+" "+time.Now().Format("15:04:05"),
		"autocorrect enabled + обычный wezterm (транспорт владельца): набор ghbdtn",
		"fired=0; correction-записей 0; статус активного слоя зелёный (контроль глухоты: zenity-клавиша видима)",
		"Verdict: автокоррекция молчит в терминале (негатив жив, TERM-04 не деградировал)")
}

// startTerm04Daemon is the TERM-04 daemon lifecycle: the user unit is
// stopped (the single-instance rule, restored at teardown), any stale stand
// daemon is stopped, the component name must be claimable with the generous
// budget, and the stand daemon restarts WITH the autocorrect config (the
// ENABLED layer of the frozen document). The 2026-10-13 finding: this path
// had NO name wait and NO unit stop, so the daemon died with "another
// goswitchd instance already registered" 17 s into the previous case's
// shutdown. Registration precedes activation — the zenity control engages
// the freshly registered engine when it spawns.
func startTerm04Daemon(ctx context.Context, s *stand, caseName string) error {
	unit, err := openSpikeWindow(ctx, caseName)
	if err != nil {
		return err
	}
	defer func() { s.restoreSpikeWindow(ctx, unit) }()
	if err := stopStaleStandDaemon(ctx); err != nil {
		return err
	}
	fmt.Printf("%s: awaiting the component name release (budget %v) before the spawn\n",
		caseName, daemonReleaseWait)
	if err := awaitNameFree(ctx, daemonReleaseWait, "before the stand daemon spawn"); err != nil {
		return err
	}
	cfgPath := filepath.Join(s.tmpDir, "autocorrect.yaml")
	if err := os.WriteFile(cfgPath, []byte(autocorrectConfigDoc), configFilePerm); err != nil {
		return fmt.Errorf("%s: write temp config: %w", caseName, err)
	}
	if err := s.restartDaemonWithArgs("-config", cfgPath); err != nil {
		return err
	}
	for _, mark := range []string{configLoadedMark, componentRegisteredMark, ctlListeningMark} {
		if err := s.waitForLog(ctx, mark, registrationWait); err != nil {
			return fmt.Errorf("%s: daemon -config spawn (%s): %w", caseName, mark, err)
		}
	}

	return nil
}

// spawnZenityControl spawns the deafness control's zenity entry (ONE mode:
// the case owns the process; the teardown reaps it) and engages it through
// the unified focus detector.
//
//nolint:noctx // the dialog must outlive the case context: the teardown reaps s.zenity
func (s *stand) spawnZenityControl(ctx context.Context, caseName string) error {
	// The baseline rides BEFORE the spawn: the dialog self-focuses within
	// milliseconds and its pairs must sit above the marker's baseline (the
	// 2026-10-14 live finding: a post-settle baseline sat above them and
	// the case hung on its own marker).
	focusBase := s.focusBaseline()
	cmd := exec.Command("zenity", "--entry",
		"--title=TERM-04 control (goswitch-e2e)", "--text=goswitch-e2e")
	cmd.Stdout = &bytes.Buffer{}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%s: spawn zenity control: %w", caseName, err)
	}
	s.zenity = cmd
	fmt.Printf("%s: zenity control spawned — awaiting daemon focus_in (map focus usually self-fires)\n", caseName)

	return s.waitFocusInFrom(ctx, caseName, "zenity control dialog", focusBase)
}

// t04NegativeSurface spawns the negative surface itself: plain wezterm —
// the owner's daily transport (no stand config; the pane is the default
// shell), engaged through the unified focus detector.
//
//nolint:noctx // the window must outlive the case context: the teardown reaps s.wezterm
func (s *stand) t04NegativeSurface(ctx context.Context, caseName string) error {
	focusBase := s.focusBaseline()
	cmd := exec.Command(weztermBin, "start", "--always-new-process")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%s: spawn negative wezterm: %w", caseName, err)
	}
	s.wezterm = cmd
	fmt.Printf("%s: negative wezterm spawned — awaiting daemon focus_in (map focus usually self-fires)\n", caseName)

	return s.waitFocusInFrom(ctx, caseName, "negative wezterm (owner transport)", focusBase)
}

// controlWithRetries proves the control surface's key flow: bounded "x"
// injections until ONE lands in the key trace (each try is one boundary-free
// letter — a miss types a single harmless character wherever focus is; the
// console notes keep the owner oriented).
func (s *stand) controlWithRetries(ctx context.Context, caseName string) error {
	for range controlRetries {
		base := s.countSub(keyTraceMark)
		if err := s.injectText(ctx, "x"); err != nil {
			return fmt.Errorf("%s: control inject: %w", caseName, err)
		}
		if werr := s.waitForNew(ctx, keyTraceMark, base+1, t04Grace); werr == nil {
			fmt.Printf("%s: control keystroke visible — zenity focused and routing\n", caseName)

			return nil
		}
	}

	return fmt.Errorf("%s: control keystroke never reached the daemon after %d tries —"+
		" the zenity dialog may not hold focus (a click on it would fire focus_in)", caseName, controlRetries)
}
