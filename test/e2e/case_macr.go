package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Djarvur/goswitch/engine"
	"github.com/Djarvur/goswitch/internal/appid"
)

// MACR spike artifacts of plan 03-05 (MACR-01/A7, ADR-005, Pitfall 7): the
// macr-probe case runs BEFORE any interception logic exists and pins, from
// the daemon's -debug ProcessKeyEvent trace, what a Super+letter injection
// ACTUALLY delivers to the IME — the letter with Mod4, or nothing (the
// chord consumed upstream by a mutter keybinding). The case PASSES on the
// fact of execution, never on the delivery: a negative verdict is a valid
// spike outcome (ADR-005 b.2 — MACR then degrades to the consumed-upstream
// detect) and feeds the SUMMARY table plus the Task-2 live-case letter
// choice. The Task-2 macr-super-letter case extends this file.

// macrProbeWait budgets one probe injection: the uinput→mutter→ibus round
// trip before the daemon log shows the records. macrQuiescence lets the
// release half of a burst land before the capture.
const (
	macrProbeWait  = 3 * time.Second
	macrQuiescence = 400 * time.Millisecond
	// macrMonitorSettle lets the dbus-monitor attach before the observed
	// traffic flows; macrRelayArgs is the argument-block depth scanned
	// under one relay header (three uint32 args, plus slack).
	macrMonitorSettle = 1500 * time.Millisecond
	macrRelayArgs     = 5
)

// macrProbeLetters are the candidate letters of the delivery scan (A7), in
// try order: 'a' is the plan's candidate (bound to toggle-application-view
// here — the scan prints it as a bound row without injecting, because the
// chord opens the shell overlay and steals the keyboard: live finding
// 2026-09-15, press delivered then focus lost), the rest exist to find a
// letter whose chord actually stays with the IME. The scan stops at the
// first DELIVERED letter and that letter becomes the Task-2 live case's
// macr.letters entry. A function, not a var: the strict lint forbids
// mutable globals (matrix.go idiom).
func macrProbeLetters() []string {
	return []string{"a", "b", "c", "z"}
}

// gsettingsLineParts is the column count of one `gsettings list-recursively`
// line: schema, key, value — anything shorter carries no binding.
const gsettingsLineParts = 3

// macrBoundSuperLetters reads the desktop's Super+letter bindings from the
// three GNOME keybinding schemas (read-only, no injection): a bound chord
// belongs to the shell and must not be probed — toggle-application-view
// (Super+a here) opens the app grid, and the overlay steals the keyboard
// from the entry (the empirically observed failure mode of the first probe
// runs). Returns letter → action name.
func macrBoundSuperLetters(ctx context.Context) map[string]string {
	bound := make(map[string]string)
	for _, schema := range []string{
		"org.gnome.shell.keybindings",
		"org.gnome.desktop.wm.keybindings",
		"org.gnome.settings-daemon.plugins.media-keys",
	} {
		out, err := runCmd(ctx, "gsettings", "list-recursively", schema)
		if err != nil {
			continue // a missing schema is a verdict-less row, never a failure
		}
		for _, line := range strings.Split(out, "\n") {
			// schema key [']'<Super>x['] ... — capture the action and the letter.
			fields := strings.Fields(line)
			if len(fields) < gsettingsLineParts {
				continue
			}
			var letter string
			for _, tok := range fields[2:] {
				// values print GVariant-wrapped: ['<Super>a'] — strip the
				// list brackets before the quote, then the closing pair.
				tok = strings.TrimPrefix(tok, "[")
				l, ok := strings.CutPrefix(tok, "'<Super>")
				if !ok {
					continue
				}
				l = strings.Trim(l, "']")
				if len(l) == 1 && l[0] >= 'a' && l[0] <= 'z' {
					letter = l

					break
				}
			}
			if letter != "" {
				bound[letter] = fields[1]
			}
		}
	}

	return bound
}

// keyRecord is one parsed "msg":"key" DEBUG record of the daemon log — the
// ProcessKeyEvent trace the whole spike verdict is derived from (keyval and
// mods arrive as 0x-prefixed hex strings, exactly as the engine traces
// them).
type keyRecord struct {
	Keyval  string `json:"keyval"`
	Keycode int    `json:"keycode"`
	Release bool   `json:"release"`
	Mods    string `json:"mods"`
}

// keyRecords parses the daemon log's key-trace records in order.
func (s *stand) keyRecords() []keyRecord {
	var recs []keyRecord
	for _, line := range s.logLines() {
		if !strings.Contains(line, `"msg":"key"`) {
			continue
		}
		var rec keyRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue // a non-key record that merely mentions the mark — skip
		}
		recs = append(recs, rec)
	}

	return recs
}

// parseHexKeyField parses one 0x-prefixed hex field of a key record; an
// unparsable value reads as 0 and only ever weakens a verdict, never fakes
// one.
func parseHexKeyField(v string) uint32 {
	u, err := strconv.ParseUint(strings.TrimPrefix(v, "0x"), 16, 32)
	if err != nil {
		return 0
	}

	return uint32(u)
}

// keyvalIsSuper reports whether the keyval is one of the Super modifier
// keysyms (XK_Super_L=0xffeb, XK_Super_R=0xffec — ibuskeysyms.h).
func keyvalIsSuper(keyval uint32) bool {
	return keyval == 0xffeb || keyval == 0xffec
}

// letterPressWithMod4 returns the first letter PRESS carrying Mod4 in the
// records — the exact event the Task-2 interception branch keys on.
func letterPressWithMod4(recs []keyRecord) (keyRecord, bool) {
	for _, r := range recs {
		kv := parseHexKeyField(r.Keyval)
		if r.Release || kv < 'a' || kv > 'z' {
			continue
		}
		if parseHexKeyField(r.Mods)&engine.MaskMod4 != 0 {
			return r, true
		}
	}

	return keyRecord{}, false
}

// superPressRelease reports which of the Super modifier's press/release
// halves appear as separate key-trace records — the observability the
// consumed-upstream detect needs (ADR-005 b.1/b.2).
func superPressRelease(recs []keyRecord) (press, release bool) {
	for _, r := range recs {
		if !keyvalIsSuper(parseHexKeyField(r.Keyval)) {
			continue
		}
		if r.Release {
			release = true
		} else {
			press = true
		}
	}

	return press, release
}

// formatKeyRecords renders one probe's observed records for the verdict
// table — the exact (keyval, keycode, mods) tuples the engine saw.
func formatKeyRecords(recs []keyRecord) string {
	if len(recs) == 0 {
		return "(no key records)"
	}
	parts := make([]string, 0, len(recs))
	for _, r := range recs {
		kind := "press"
		if r.Release {
			kind = "release"
		}
		parts = append(parts, fmt.Sprintf("%s keyval=%s keycode=%d mods=%s",
			kind, r.Keyval, r.Keycode, r.Mods))
	}

	return strings.Join(parts, "; ")
}

// runMacrProbe executes the A7 delivery probe (plan 03-05 Task 1): three
// probes against the -debug daemon on a focused zenity surface —
// (a) a super+letter scan for a delivered letter, (b) the same physical key
// after the internal flip to RU mode (the interception input must be
// layout-independent), (c) the bare Super press/release (the
// consumed-upstream path's raw material). Every probe's observed records
// are printed; the grep-stable verdict line "letter-delivered: yes|no"
// closes the case.
func runMacrProbe(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("macr-probe needs the zenity entry surface (locked-session fallback engaged?)")
	}

	// (a) the delivery scan: bound letters become verdict rows without
	// injection (their chords belong to the shell — Super+a here opens the
	// application view, stealing the keyboard after the press reaches the
	// IME); the first unbound letter whose chord reaches ProcessKeyEvent
	// with Mod4 becomes the canonical probe letter.
	bound := macrBoundSuperLetters(ctx)
	for _, letter := range macrProbeLetters() {
		if action, isBound := bound[letter]; isBound {
			fmt.Printf("macr-probe: (a) super+%s: shell-bound (%s) — not probed,"+
				" the chord opens a shell overlay\n", letter, action)
		}
	}
	delivered, err := s.macrScanSuperLetters(ctx, bound)
	if err != nil {
		return err
	}
	if delivered.letter == "" {
		fmt.Println("letter-delivered: no")
		fmt.Println("consumed-upstream-detect: see (c) below")

		return s.macrSuperSoloProbe(ctx)
	}
	press, _ := letterPressWithMod4(delivered.records)
	fmt.Printf("letter-delivered: yes (letter %q keyval=%s mods=%s)\n",
		delivered.letter, press.Keyval, press.Mods)

	// (b) the same physical key under the internal RU mode: the flip never
	// touches the session XKB group (ADR-001 Option B), so the interception
	// decision input must arrive identically — pinned, not assumed.
	ruRecs, err := s.macrSuperLetterInRU(ctx, delivered.letter)
	if err != nil {
		return err
	}
	ruPress, ruOK := letterPressWithMod4(ruRecs)
	fmt.Printf("macr-probe: (b) super+%s in RU mode: %s\n", delivered.letter, formatKeyRecords(ruRecs))
	fmt.Printf("macr-probe: (b) ru-mode delivery: %v", ruOK)
	if ruOK {
		fmt.Printf(" (keyval=%s mods=%s)\n", ruPress.Keyval, ruPress.Mods)
	} else {
		fmt.Println()
	}

	return s.macrSuperSoloProbe(ctx)
}

// macrLetterProbe is one scan step's outcome: the letter tried and every key
// record the round produced.
type macrLetterProbe struct {
	letter  string
	records []keyRecord
}

// macrScanSuperLetters probes super+letter per candidate (skipping the
// shell-bound letters) until one delivers a letter press with Mod4 AND
// leaves the entry holding focus. Delivery alone is not usability: a
// shell-bound chord (Super+a is toggle-application-view here — live finding
// 2026-09-15) reaches the IME with its press but then opens a shell overlay
// that steals the keyboard — the interception burst would land in the
// overlay, never in the field. Every tried letter's records are printed as
// a verdict row (the consumed-upstream rows are the spike's finding too).
func (s *stand) macrScanSuperLetters(ctx context.Context, bound map[string]string) (macrLetterProbe, error) {
	for _, letter := range macrProbeLetters() {
		if _, isBound := bound[letter]; isBound {
			continue // printed as a bound row by the caller — never injected
		}
		recs, err := s.macrSuperLetterRound(ctx, letter)
		if err != nil {
			return macrLetterProbe{}, err
		}
		fmt.Printf("macr-probe: (a) super+%s: %s\n", letter, formatKeyRecords(recs))
		if press, ok := letterPressWithMod4(recs); ok {
			focused, ferr := s.focusWitness(ctx)
			if ferr == nil && strings.HasPrefix(focused, "zenity:") {
				fmt.Printf("macr-probe: (a) super+%s delivered: keyval=%s mods=%s (entry kept focus)\n",
					letter, press.Keyval, press.Mods)

				return macrLetterProbe{letter: letter, records: recs}, nil
			}
			witness := "(witness failed)"
			if ferr == nil {
				witness = focused
			}
			fmt.Printf("macr-probe: (a) super+%s press delivered but focus stolen (witness %q)"+
				" — shell-bound chord, not a usable interception letter\n", letter, witness)
		}
		if err := s.macrRefocus(ctx); err != nil {
			return macrLetterProbe{}, err
		}
	}

	return macrLetterProbe{}, nil
}

// macrSuperLetterRound injects one super+letter chord and captures the key
// records it produced (a timeout capturing nothing is itself a verdict —
// the chord was consumed upstream).
func (s *stand) macrSuperLetterRound(ctx context.Context, letter string) ([]keyRecord, error) {
	base := len(s.keyRecords())
	if err := s.pressKey(ctx, "super+"+letter); err != nil {
		return nil, fmt.Errorf("inject super+%s: %w", letter, err)
	}
	// At minimum the Super press must show; the letter halves may never
	// arrive (that IS the probe). A quiet timeout falls through to the
	// capture below.
	_ = s.waitForNew(ctx, `"msg":"key"`, base+1, macrProbeWait)
	if err := sleepCtx(ctx, macrQuiescence); err != nil {
		return nil, err
	}

	return keyRecordsSince(s.keyRecords(), base), nil
}

// macrSuperLetterInRU flips the internal mode to RU (single Right-Shift
// tap, D-34 mode record oracle) and repeats the canonical chord: the RU
// mode branch must never see the letter first — the Task-2 MACR branch sits
// above it. Flips the mode back to EN afterwards.
func (s *stand) macrSuperLetterInRU(ctx context.Context, letter string) ([]keyRecord, error) {
	if err := s.injectKeys(ctx, "Shift_R"); err != nil {
		return nil, err
	}
	if err := s.waitForLog(ctx, `"msg":"mode","to":"ru"`, decisionWait); err != nil {
		return nil, fmt.Errorf("macr-probe RU flip: %w", err)
	}
	recs, err := s.macrSuperLetterRound(ctx, letter)
	if err != nil {
		return nil, err
	}
	if err := s.injectKeys(ctx, "Shift_R"); err != nil {
		return nil, err
	}
	if err := s.waitForLog(ctx, `"msg":"mode","to":"en"`, decisionWait); err != nil {
		return nil, fmt.Errorf("macr-probe EN flip-back: %w", err)
	}

	return recs, nil
}

// macrSuperSoloProbe is probe (c): the bare Super press and release with no
// letter in between — the raw shape of the consumed-upstream path. On this
// desktop the bare Super opens the shell overview (the super-space-alive
// finding), so the round dismisses it with the proven esc name and settles
// the entry's focus back before closing. The verdict line names whether
// BOTH halves are visible as separate ProcessKeyEvent records — the
// observability the Task-2 detect keys on.
func (s *stand) macrSuperSoloProbe(ctx context.Context) error {
	base := len(s.keyRecords())
	if err := s.pressKey(ctx, "super"); err != nil {
		return fmt.Errorf("inject super: %w", err)
	}
	_ = s.waitForNew(ctx, `"msg":"key"`, base+1, macrProbeWait)
	if err := sleepCtx(ctx, macrQuiescence); err != nil {
		return err
	}
	recs := keyRecordsSince(s.keyRecords(), base)
	fmt.Printf("macr-probe: (c) super solo: %s\n", formatKeyRecords(recs))

	press, release := superPressRelease(recs)
	viable := press && release
	fmt.Printf("macr-probe: (c) super press=%v release=%v\n", press, release)
	fmt.Printf("consumed-upstream-detect: %v\n", viable)

	// Close whatever surface the bare Super opened and give the entry its
	// keyboard back — the deferred closeEntrySurface needs a live entry.
	return s.macrRefocus(ctx)
}

// macrRefocus recovers the entry's keyboard after a shell-bound chord took
// it (Super+a opens the application view here — the chord reaches the IME
// with its press, then the shell surface takes over): dismiss the overlay
// with the proven esc name, then poll with the pid-keyed grabFocus pokes of
// waitZenityEntry's recovery idiom — the refused call still re-issues the
// window activation, so the entry wins focus back on the active desktop. A
// no-op when the entry never lost focus.
func (s *stand) macrRefocus(ctx context.Context) error {
	now, err := s.focusWitness(ctx)
	if err != nil {
		return err
	}
	if strings.HasPrefix(now, "zenity:") {
		return nil
	}
	if err := s.pressKey(ctx, "esc"); err != nil {
		return err
	}
	if err := sleepCtx(ctx, superSettleWait); err != nil {
		return err
	}
	pid := strconv.Itoa(s.zenity.Process.Pid)
	deadline := time.Now().Add(surfaceFocusWait)
	nextGrab := time.Now() // focus was already lost — poke immediately
	for {
		now, err := s.focusWitness(ctx)
		if err != nil {
			return err
		}
		if strings.HasPrefix(now, "zenity:") {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("zenity entry did not regain focus (witness %q)", now)
		}
		if time.Now().After(nextGrab) {
			// Best-effort poke: a no-op while the window tree is not up.
			_, _ = runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "grab-input-pid", pid)
			nextGrab = time.Now().Add(zenityGrabEvery)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// keyRecordsSince returns recs[from:] — the records captured after a
// baseline count.
func keyRecordsSince(recs []keyRecord, from int) []keyRecord {
	if from >= len(recs) {
		return nil
	}

	return recs[from:]
}

// Task-2 live case (plan 03-05): the interception itself, driven through a
// -config daemon, with the IBus WIRE as the oracle. The live lessons that
// shaped this case: (1) the interception's Ctrl+letter burst cannot fire
// while the physical Super is held — the client still tracks the held
// modifier, so the remap rides the Super RELEASE; (2) even a correctly
// relayed forward does not reach the widget on this GTK4-Wayland desktop:
// the dbus-monitor capture shows ibus-daemon forwarding the burst to the
// focused InputContext (mutter's IM client), and the widget never applies
// it (no cut, no cursor move — the same never-live-driven family as the
// ADR-003 level-2 replay and the D-28 Ctrl+V burst; the unit corpus owns
// the behavioral contract). The case therefore pins everything up to the
// client boundary — consume + INFO record + the on-bus relay — and pins
// the field's unchanged state as the documented actual.
const (
	macrSuperLetter = "x"
	macrText        = "hello"
)

// macrConfigTmpl is the case's complete config document (no defaults
// overlay — the 03-02 strict-parse decision): the layer ON with the
// interception set 'x' (free of shell bindings on this desktop; the Task-1
// spike).
const macrConfigTmpl = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
macr:
  enabled: true
  letters: "x"
  apps: []
  alt_modifier: ""
`

// macrIbusCapture is a bounded dbus-monitor capture of the live IBus bus,
// filtered to ForwardKeyEvent blocks — the ground truth of the remap's
// RELAY (the engine's emissions and ibus-daemon's forwarding to the
// focused input context both ride this bus).
type macrIbusCapture struct {
	cmd  *exec.Cmd
	path string
}

// shq single-quotes s for the monitoring command line.
func shq(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// startMacrIbusCapture resolves the live IBus address and starts the
// filtered monitor in its own process group (stop() reaps the whole
// group — the pipe keeps dbus-monitor and grep alive past the shell).
func startMacrIbusCapture(ctx context.Context) (*macrIbusCapture, error) {
	addr, err := runCmd(ctx, "ibus", "address")
	if err != nil {
		return nil, fmt.Errorf("ibus address: %w", err)
	}
	if addr == "" || !strings.HasPrefix(addr, "unix:") {
		return nil, fmt.Errorf("ibus address output %q is not a unix address", addr)
	}
	tmpDir, err := os.MkdirTemp("", "goswitch-macr-monitor-*")
	if err != nil {
		return nil, fmt.Errorf("create monitor temp dir: %w", err)
	}
	path := filepath.Join(tmpDir, "forwards.txt")
	//nolint:noctx // diagnostics-only capture; the stand owns the lifetime (stop kills the group)
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf("dbus-monitor --address %s 2>/dev/null | grep --line-buffered -A6 ForwardKeyEvent >> %s",
			shq(addr), shq(path)))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start dbus-monitor: %w", err)
	}
	// The monitor needs a beat to attach before the observed traffic flows
	// (it lost the injection race on this bus in the first runs).
	if err := sleepCtx(ctx, macrMonitorSettle); err != nil {
		return nil, err
	}

	return &macrIbusCapture{cmd: cmd, path: path}, nil
}

// stop kills the capture's whole process group and waits it out
// (idempotent — a nil Process marks a stopped capture).
func (c *macrIbusCapture) stop() {
	if c.cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-c.cmd.Process.Pid, syscall.SIGTERM)
	_ = c.cmd.Wait()
	c.cmd.Process = nil
}

// relayedCount counts the capture lines where ibus-daemon FORWARDED a
// ForwardKeyEvent carrying keyval to the focused input context — the relay
// header carries the InputContext path; the engine's own emissions (the
// other half of the capture) do not. The keyval is looked up in the header
// line's argument block (the next few lines), line-based so the grep
// grouping never matters.
func (c *macrIbusCapture) relayedCount(keyval uint32) int {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return 0
	}
	lines := strings.Split(string(data), "\n")
	want := fmt.Sprintf("uint32 %d", keyval)
	count := 0
	for i, line := range lines {
		if !strings.Contains(line, "org.freedesktop.IBus.InputContext") ||
			!strings.Contains(line, "member=ForwardKeyEvent") {
			continue
		}
		for _, arg := range lines[i:min(i+macrRelayArgs, len(lines))] {
			if strings.Contains(arg, want) {
				count++

				break
			}
		}
	}

	return count
}

// waitForRelay polls until the daemon relayed want keyval events (the
// wait-for-condition idiom, never a fixed sleep).
func (c *macrIbusCapture) waitForRelay(ctx context.Context, keyval uint32, want int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		got := c.relayedCount(keyval)
		if got >= want {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout after %v: ibus relay of keyval %d seen %d times, want >= %d",
				timeout, keyval, got, want)
		}
		if err := sleepCtx(ctx, logPollInterval); err != nil {
			return err
		}
	}
}

// macrRelayHalves is the wire oracle's count: both letter halves (press
// and release) of the remap relayed by ibus-daemon.
const macrRelayHalves = 2

// startMacrConfigDaemon writes the case's config document, restarts the
// daemon on it and re-activates the engine — the -config spawn ladder the
// combo case established.
func startMacrConfigDaemon(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	cfgPath := filepath.Join(s.tmpDir, "macr-super-letter.yaml")
	if err := os.WriteFile(cfgPath, []byte(macrConfigTmpl), configFilePerm); err != nil {
		return fmt.Errorf("macr-super-letter: write temp config: %w", err)
	}
	if err := s.restartDaemonWithArgs("-config", cfgPath); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"config loaded"`, registrationWait); err != nil {
		return fmt.Errorf("macr-super-letter daemon -config spawn: %w", err)
	}
	if err := s.waitForLog(ctx, "component registered", registrationWait); err != nil {
		return fmt.Errorf("macr-super-letter daemon re-registration: %w", err)
	}

	return s.activateGoswitch(ctx)
}

// runMacrSuperLetter proves the interception live (MACR-01) up to the
// client boundary: a daemon on a config document with the layer on
// intercepts super+x (the INFO record names the config letter; the press
// is consumed — the chord's letter never reaches the client raw) and the
// remap burst rides the Super release onto the IBus bus, where
// ibus-daemon relays both letter halves to the focused input context. The
// field's unchanged content is PINNED as the documented GTK4-Wayland
// actual (the widget does not apply forwarded chords on this desktop).
func runMacrSuperLetter(ctx context.Context, s *stand) error {
	if err := startMacrConfigDaemon(ctx, s); err != nil {
		return err
	}

	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("macr-super-letter needs the zenity entry surface (locked-session fallback engaged?)")
	}

	keyBase := s.countSub(`"msg":"key"`)
	if err := s.injectText(ctx, macrText); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"key"`, keyBase+minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("macr-super-letter key visibility: %w", err)
	}
	// The word must settle before the chord — the same early-flake gate as
	// the combo round.
	if err := s.waitZenityText(ctx, macrText); err != nil {
		return fmt.Errorf("macr-super-letter typed text: %w", err)
	}

	capture, err := startMacrIbusCapture(ctx)
	if err != nil {
		return err
	}
	defer capture.stop()

	// The chord: super+x through the same injection-name class the Task-1
	// probe proved for unbound letters; the remap rides the release.
	if err := s.pressKey(ctx, "super+"+macrSuperLetter); err != nil {
		return fmt.Errorf("macr-super-letter inject super+%s: %w", macrSuperLetter, err)
	}
	if err := s.waitForNew(ctx, `"msg":"super intercept","key":"`+macrSuperLetter+`"`, 1, correctionWait); err != nil {
		return fmt.Errorf("macr-super-letter interception record: %w", err)
	}
	// The wire oracle: both letter halves relayed by ibus-daemon to the
	// focused input context (keyval 'x' = 0x78 = 120).
	if err := capture.waitForRelay(ctx, uint32(macrSuperLetter[0]), macrRelayHalves, correctionWait); err != nil {
		return fmt.Errorf("macr-super-letter wire relay: %w", err)
	}

	out, err := s.closeZenity(ctx)
	if err != nil {
		return err
	}
	if out != macrText {
		return fmt.Errorf(
			"macr-super-letter oracle: entry printed %q, want %q — the GTK4-Wayland non-action is the actual",
			out, macrText)
	}

	return nil
}

// Task-3 live case (plan 03-05, ADR-005 a): the per-app list scoped by the
// LIVE a11y observer (the A4 spike). The observer only learns from a
// focus GAIN — a freshly mapped window emits one (an already-focused
// window does not; the poke proved useless live) — so every observer
// boundary in the case is crossed by a FRESH zenity: the case's own
// observer starts first and reads the identity from surface A's map, and
// the daemon's LAZY observer (started by the first key event, which
// always postdates its surface's focus gain) is fed by surface C's map.
// The observed identity becomes the config's macr.apps entry — the
// vocabulary the daemon's own observer reports — and the interception
// fires under that matched focus. An unavailable observer is a VALID A4
// verdict: the case prints it and passes (the documented degradation —
// the unit corpus owns the ladder).
const (
	macrPerAppIdentWait = 10 * time.Second    // the observer's first identity budget
	macrPerAppProbeText = "hello"             // the typed probe of the interception round
	macrPerAppStarter   = "starter-hello-too" // surface B only starts the daemon's observer
)

// macrPerAppConfigTmpl parameterizes the layer's document by the observed
// app identity (the list entry the daemon's observer must match).
const macrPerAppConfigTmpl = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
macr:
  enabled: true
  letters: "x"
  apps: ["%s"]
  alt_modifier: ""
`

// waitFocusedApp polls the observer until it holds an identity.
func waitFocusedApp(ctx context.Context, obs *appid.Observer) (string, error) {
	deadline := time.Now().Add(macrPerAppIdentWait)
	for {
		if app, err := obs.FocusedApp(); err == nil && app != "" {
			return app, nil
		}
		if time.Now().After(deadline) {
			app, err := obs.FocusedApp()
			if err != nil {
				return app, fmt.Errorf("app identity poll: %w", err)
			}

			return app, nil
		}
		if err := sleepCtx(ctx, logPollInterval); err != nil {
			return "", err
		}
	}
}

// startPerAppDaemon writes the per-app document (the observed identity as
// the list entry), restarts the daemon on it and re-activates the engine.
func startPerAppDaemon(ctx context.Context, s *stand, app string) error {
	cfgPath := filepath.Join(s.tmpDir, "macr-per-app.yaml")
	doc := fmt.Sprintf(macrPerAppConfigTmpl, app)
	if err := os.WriteFile(cfgPath, []byte(doc), configFilePerm); err != nil {
		return fmt.Errorf("macr-per-app: write temp config: %w", err)
	}
	if err := s.restartDaemonWithArgs("-config", cfgPath); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"config loaded"`, registrationWait); err != nil {
		return fmt.Errorf("macr-per-app daemon -config spawn: %w", err)
	}
	if err := s.waitForLog(ctx, "component registered", registrationWait); err != nil {
		return fmt.Errorf("macr-per-app daemon re-registration: %w", err)
	}

	return s.activateGoswitch(ctx)
}

// runMacrPerApp proves the per-app scoping live: the observed a11y
// identity of the stand's own surface becomes the macr.apps entry, and
// the daemon intercepts super+x only under that focus. Prints the
// per-app verdict lines; an unavailable observer degrades to the
// documented A4 verdict (still a PASS — the ladder is the design).
func runMacrPerApp(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}

	// The case-side observer probe (A4) starts BEFORE the surface: A's
	// own map-focus event is its first meal.
	obs, obsErr := appid.Start(ctx)
	if obsErr != nil {
		fmt.Printf("per-app-observer: unavailable (%v) — the documented A4 degradation\n", obsErr)

		return nil
	}

	app, err := s.macrObserveZenity(ctx, obs)
	if err != nil {
		return err
	}
	if app == "" {
		return nil // the A4 degradation verdict, already printed
	}
	if err := startPerAppDaemon(ctx, s, app); err != nil {
		return err
	}

	// Surface B only starts the daemon's lazy observer (its key events
	// carry the first applySnapshot); its own focus gain predates the
	// start, so it cannot feed the cache — surface C's map does.
	if err := s.macrStarterRound(ctx); err != nil {
		return err
	}

	return s.macrInterceptRound(ctx)
}

// macrObserveZenity opens a fresh zenity and reads the identity the
// observer gleans from its map-focus event — "" (after printing the
// verdict) when the live observer produced none.
func (s *stand) macrObserveZenity(ctx context.Context, obs *appid.Observer) (string, error) {
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return "", err
	}
	if kind != surfaceZenity {
		_ = s.closeEntrySurface(ctx, kind)

		return "", errors.New("macr-per-app needs the zenity entry surface (locked-session fallback engaged?)")
	}
	app, appErr := waitFocusedApp(ctx, obs)
	if appErr != nil || app == "" {
		fmt.Printf("per-app-observer: no identity observed (%v) — the documented A4 degradation\n", appErr)
		_ = s.closeEntrySurface(ctx, kind)

		return "", nil
	}
	fmt.Printf("per-app-observer: focused identity %q\n", app)
	if !strings.Contains(strings.ToLower(app), "zenity") {
		_ = s.closeEntrySurface(ctx, kind)

		return "", fmt.Errorf("macr-per-app: observed identity %q does not look like the stand's zenity surface", app)
	}

	return app, s.closeEntrySurface(ctx, kind)
}

// macrStarterRound runs the short typing round that lands the daemon's
// first key events — the applySnapshot that starts its lazy observer.
func (s *stand) macrStarterRound(ctx context.Context) error {
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	if kind != surfaceZenity {
		_ = s.closeEntrySurface(ctx, kind)

		return errors.New("macr-per-app needs the starter zenity surface (locked-session fallback engaged?)")
	}
	if err := s.injectText(ctx, macrPerAppStarter); err != nil {
		return err
	}
	if err := s.waitZenityText(ctx, macrPerAppStarter); err != nil {
		return fmt.Errorf("macr-per-app starter round: %w", err)
	}

	return s.closeEntrySurface(ctx, kind)
}

// macrInterceptRound opens the TARGET surface (its fresh map-focus event
// feeds the daemon's observer), types the probe, fires the chord and
// gates the interception record; the unchanged readback closes the round.
func (s *stand) macrInterceptRound(ctx context.Context) error {
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("macr-per-app needs the target zenity surface (locked-session fallback engaged?)")
	}

	keyBase := s.countSub(`"msg":"key"`)
	if err := s.injectText(ctx, macrPerAppProbeText); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"key"`, keyBase+minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("macr-per-app key visibility: %w", err)
	}
	if err := s.waitZenityText(ctx, macrPerAppProbeText); err != nil {
		return fmt.Errorf("macr-per-app typed text: %w", err)
	}

	interceptBase := s.countSub(`"msg":"super intercept"`)
	if err := s.pressKey(ctx, "super+"+macrSuperLetter); err != nil {
		return fmt.Errorf("macr-per-app inject super+%s: %w", macrSuperLetter, err)
	}
	if err := s.waitForNew(ctx, `"msg":"super intercept"`, interceptBase+1, correctionWait); err != nil {
		return fmt.Errorf("macr-per-app interception under the matched focus: %w", err)
	}

	out, err := s.closeZenity(ctx)
	if err != nil {
		return err
	}
	if out != macrPerAppProbeText {
		return fmt.Errorf(
			"macr-per-app oracle: entry printed %q, want %q (the GTK4-Wayland non-action)",
			out, macrPerAppProbeText)
	}
	fmt.Println("per-app-observer: interception fired under the matched focus")

	return nil
}
