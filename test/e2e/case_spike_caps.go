package main

// The capability-table case for the second terminal (plan 09-02, TERM-01):
// one live gnome-terminal/VTE round under the manual-activation protocol —
// the injection rows the kill criterion and the Phase-10 gate need (the
// caps-table rows for WEZTERM are written by the wezterm-flip case's
// rounds, which own those surfaces and their data). Observation rows
// record what the session honestly yields; feasibility-gated deferrals are
// marked deferred-with-reason, never silently skipped.

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Attach-mode knobs of the TERM-04 negative round.
const (
	controlRetries = 3
	secTerm04      = "TERM-04 негатив (attach-семейство, rev D-55 §7.2)"
)

// runSpikeCaps drives the VTE round: a fresh gnome-terminal window running
// the pty-capturing program, clicked by the owner, then the injection
// round — delivery count, the flip, the commit-landing shortcut check
// (flip → RU → type → pty already привет? — the decisive cheap probe for
// the VTE word path), and the select-chord with the PRIMARY sentinel (the
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

	stale, err := s.openVTERound(ctx, caseName)
	if err != nil {
		return err
	}
	if stale {
		// The desync row is recorded by the arming; the injection rounds
		// are skipped (every keystroke would ride the stale context) — the
		// case still closes honestly with the branch noted.
		if err := appendSpikeRow(secCaps, "VTE итог "+time.Now().Format("15:04:05"),
			"VTE раунд (desync-ветка)", "инъекционные раунды не исполнялись — арминг застал desync",
			"Verdict: reproduced-desync branch (раунды пропущены)"); err != nil {
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
// the capture spawn, the pane-ready proof, the owner click gate and the
// a11y identity row (the witness form for the focused GTK3 citizen),
// closing with the arming gate. A reproduced desync is reported as
// stale=true (the caller records the branch) — never an error.
func (s *stand) openVTERound(ctx context.Context, caseName string) (bool, error) {
	unit, err := openSpikeWindow(ctx, caseName)
	if err != nil {
		return false, err
	}
	defer func() { s.restoreSpikeWindow(ctx, unit) }()

	if _, err := startSpikeDaemon(ctx, s, caseName); err != nil {
		return false, err
	}

	ptyPath := filepath.Join(s.tmpDir, caseName+"-in")
	readyPath := filepath.Join(s.tmpDir, caseName+"-ready")
	vcap, err := startVTECapture(ctx, readyPath, ptyPath)
	if err != nil {
		return false, fmt.Errorf("%s: %w", caseName, err)
	}
	defer vcap.close()
	if err := waitPaneReady(ctx, caseName, readyPath); err != nil {
		return false, fmt.Errorf("%s: %w", caseName, err)
	}

	fmt.Printf("%s: fixture window mapped — awaiting owner click (gnome-terminal)\n", caseName)
	if err := s.waitSurfaceClick(ctx, caseName, "gnome-terminal window", "gnome-terminal"); err != nil {
		return false, fmt.Errorf("%s: %w", caseName, err)
	}
	// Row 5 (a11y identity): the witness form for the VTE surface — the
	// GTK3 citizen SHOULD be nameable while focused.
	witness, _ := s.focusWitness(ctx)
	if err := appendSpikeRow(secCaps, "VTE a11y "+time.Now().Format("15:04:05"),
		"focus_helper.py witness при сфокусированном VTE",
		"witness="+witness, "Verdict: форма идентичности записана (наблюдаемое выше)"); err != nil {
		return false, err
	}

	report, aerr := s.armGate(ctx, caseName, ptyPath)
	if aerr != nil {
		return false, fmt.Errorf("%s: %w", caseName, aerr)
	}
	if report.outcome == armStaleContext {
		if rerr := appendDesyncRow(secCaps, caseName,
			fmt.Sprintf("armGate probe (ydotool type %s + Enter)", armProbeText),
			report, ptyPath); rerr != nil {
			return false, rerr
		}

		return true, nil
	}

	return false, nil
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

// t04ZenityCmd is the owner's paste for the TERM-04 deafness control (the
// out file's creation is the launch marker; the dialog stays until the
// owner closes it).
const t04ZenityCmd = "zenity --entry" +
	" --title=\"КЛИК 5 из 6 — TERM-04 control (goswitch-e2e)\"" +
	" --text=goswitch-e2e > /tmp/gs-owner-t04z-out 2>/dev/null"

// t04ZenityOut and t04GraceKnobs are the attach-mode markers and budgets of
// the negative round.
const (
	t04ZenityOut = "/tmp/gs-owner-t04z-out"
	t04Grace     = 60 * time.Second
)

// runTerm04Negative carries the TERM-04 frozen negative under the
// attach-mode family discipline (rev D-55 §7.2: замороженные негативные
// ряды перенесены в дисциплину нового семейства; the automation-era frozen
// case cannot run its zenity witness gate on the drifted session — the a11y
// FOCUSED frames are dead even for genuinely focused windows). The layer is
// ENABLED (the frozen document); the deafness control is an owner-launched
// zenity keystroke, the negative surface an owner-launched plain wezterm
// (the owner's default config — the daily transport). Oracles: zero fired
// records, zero correction records, status counters fired=0.
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
	// blocklist).
	if err := startAutocorrectConfigDaemon(ctx, s, autocorrectConfigDoc); err != nil {
		return err
	}
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}

	// The control surface: owner-launched zenity. The launch marker is the
	// out file; the focus gate is the control keystroke itself (the a11y
	// witness is dead on this session — even genuinely focused windows
	// never name themselves; the VTE a11y row above pins it).
	fmt.Printf("%s: OWNER-LAUNCH — вставить в СВОЙ терминал:\n  %s\n", caseName, t04ZenityCmd)
	if err := s.waitMarker(ctx, caseName, t04ZenityOut, "zenity control"); err != nil {
		return err
	}
	if err := s.controlWithRetries(ctx, caseName); err != nil {
		return err
	}

	// The negative surface: owner-launched plain wezterm (the owner's daily
	// transport). No markers exist by design — the negative expects NO
	// engagement — so the round waits a fixed, silent grace for the owner to
	// launch and click it, states the assumption, and injects.
	if err := s.t04NegativeGrace(ctx, caseName); err != nil {
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
	verdicted = true

	return appendSpikeRow(secTerm04, caseName+" "+time.Now().Format("15:04:05"),
		"autocorrect enabled + владельческий wezterm: набор ghbdtn",
		"fired=0; correction-записей 0; статус активного слоя зелёный (контроль глухоты: zenity-клавиша видима)",
		"Verdict: автокоррекция молчит в терминале (негатив жив, TERM-04 не деградировал)")
}

// t04NegativeGrace is the fixed silent grace before the negative
// injection: the owner launches and clicks the plain wezterm window while
// the case waits (console notes only — the 2026-10-12 no-sound directive);
// no markers exist for this surface by design.
func (s *stand) t04NegativeGrace(ctx context.Context, caseName string) error {
	fmt.Printf("%s: OWNER-LAUNCH — вставить в СВОЙ терминал:\n  wezterm start --always-new-process\n", caseName)
	fmt.Printf("%s: окно wezterm должно быть НАЖАТО в течение %v; инъекция после паузы\n",
		caseName, t04Grace)
	deadline := time.Now().Add(t04Grace)
	notes := newNoteGate()
	for {
		if time.Now().After(deadline) {
			return nil
		}
		if notes.due() {
			fmt.Printf("%s: негативная поверхность: окно wezterm запущено и нажато? инъекция через %v\n",
				caseName, time.Until(deadline).Round(time.Second))
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// waitMarker waits for a launch marker file (the owner's paste created it).
func (s *stand) waitMarker(ctx context.Context, caseName, path, what string) error {
	deadline := time.Now().Add(spikeClickWait)
	notes := newNoteGate()
	for {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s: awaiting OWNER LAUNCH of %s: no marker %s within %v",
				caseName, what, path, spikeClickWait)
		}
		if notes.due() {
			fmt.Printf("%s: awaiting owner launch of %s (marker %s)\n", caseName, what, path)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// controlWithRetries proves the control surface's key flow: bounded "x"
// injections until ONE lands in the key trace (each try is one boundary-free
// letter — a miss types a single harmless character wherever focus is; the
// console notes keep the owner on the dialog).
func (s *stand) controlWithRetries(ctx context.Context, caseName string) error {
	if err := sleepCtx(ctx, t04Grace); err != nil {
		return err
	}
	for range controlRetries {
		base := s.countSub(keyTraceMark)
		if err := s.injectText(ctx, "x"); err != nil {
			return fmt.Errorf("%s: control inject: %w", caseName, err)
		}
		if werr := s.waitForNew(ctx, keyTraceMark, base+1, keyWait); werr == nil {
			fmt.Printf("%s: control keystroke visible — zenity focused and routing\n", caseName)

			return nil
		}
	}

	return fmt.Errorf("%s: control keystroke never reached the daemon after %d tries —"+
		" is the zenity dialog focused?", caseName, controlRetries)
}
