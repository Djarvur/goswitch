package main

// The wezterm spike fixture (plan 09-02): a dedicated terminal window
// (--always-new-process, never a delegation to an owner instance — the
// chromium-profile isolation lesson) running a pty-capturing program, plus
// the unified observation primitives every spike case drives: the focus
// detector (daemon focus_in above the spawn baseline — the only focus
// signal; no AT-SPI anywhere), the pty read-back, the arming gate of the
// IME channel (classifyArm: window / stale-context desync / dead) and the
// PRIMARY/flip-round helpers. ONE mode: the case always spawns the fixture
// itself and runs from the owner's interactive terminal; map focus usually
// self-fires, a click fires it when it does not, and the 30-minute budget
// is the only wait.

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Djarvur/goswitch/internal/engine"
)

// Fixture knobs: every value is a behavioral constant of the spike stand,
// named once (the stand-knobs discipline of main.go).
const (
	// weztermXimFixture is the stand config mirroring the owner's transport
	// (enable_wayland=false — the owner's ~/.wezterm.lua is NEVER touched,
	// Pitfall 7).
	weztermXimFixture = "gs-e2e-wezterm-xim.lua"

	// weztermWaylandFixture is the stand config of the native-Wayland recipe
	// trial (enable_wayland=true + the CopyMode Shift+Home select recipe —
	// the Multiple-sequence from the phase research). Still a stand-owned
	// file: the owner's ~/.wezterm.lua stays untouched (Pitfall 7).
	weztermWaylandFixture = "gs-e2e-wezterm-wayland.lua"

	// vteBin is the GNOME Terminal surface of the caps round (a GTK3
	// client of the same IM channel the probe measures).
	vteBin = "gnome-terminal"

	// flipPairOutEN and flipPairInRU are the daemon's own flip records (the
	// SetGlobalEngine switch reflects as the engine focus pair — the
	// two-source-flip precedent): P2's observable.
	flipPairOutEN = `"msg":"focus_out","engine":"goswitch-en"`
	flipPairInRU  = `"msg":"focus_in","engine":"goswitch-ru"`

	// corpus and expected hexes: the synthetic corpus discipline (Pitfall 6)
	// — hex only in journals, never the raw bytes.
	corpusLatin    = "ghbdtn"
	corpusHexLatin = "67686264746e"             // ghbdtn
	corpusHexRU    = "d0bfd180d0b8d0b2d0b5d182" // привет

	// keyTraceMark is the -debug key-trace record prefix the delivery
	// counters read (typeAndCountKeys, the arm gate's diagnostic).
	keyTraceMark = `"msg":"key"`
	// probeKeyMark is the key-trace record of the arming probe's FIRST
	// digit (keyval 0x31 = '1'): the engine-traced signal of the arming
	// gate (the live-smoke finding, 2026-10-12: the stale-context desync
	// leaves the probe in the ENGINE trace while the fixture pty stays
	// empty). Residual risk, documented honestly: an owner typing a '1'
	// into an IM-routed window mid-gate would also grow this count — the
	// rows print the numbers, the evidence stays auditable.
	probeKeyMark = `"keyval":"0x31"`
	// probeForwardMark is the probe verb's emission record (the smoke's
	// emission contract).
	probeForwardMark = `"msg":"probe forward"`
	// probeEmissionRecords is the emission pair's floor: one press + one
	// release record.
	probeEmissionRecords = 2

	// armGateTries bounds the arming race's retries; armGateSettle is the
	// pause between attempts (the live finding: one run delivers, the next
	// does not — a gate, not an assumption).
	armGateTries  = 5
	armGateSettle = 300 * time.Millisecond
	// armGateWait bounds the wait for the arming probe line to land in the
	// fixture pty capture per attempt.
	armGateWait = 2 * time.Second

	// armProbeText is the arming probe: digits only — byte-identical on
	// every layout, never consumed by the daemon's RU script, never an
	// autocorrect candidate — so the probe measures ONLY the delivery
	// channel. It is typed, then flushed out of the tty line discipline
	// into the capture by an Enter: armProbeEnter.
	armProbeText = "1234"
)

// armProbeEnter is the Enter keystroke of the arming probe, in ydotool
// 0.1.8's documented key-sequence form (its `key --help`: "modifiers and
// keys, separated by plus"; the binary's table names it KEY_ENTER). The
// 1.0-style colon form ("28:1") is NOT 0.1.8 syntax: the live-smoke finding
// (2026-10-12) is that each colon arg typed its FIRST CHARACTER — the
// owner's prompt collected "123422123422…" (probe + two '2's per try) and
// the R1 teardown's "29:1 32:1 …" typed "2332" instead of the closing
// Ctrl+D. A missing real Enter would leave a canonical tty unflushed and
// misread a healthy round as stale, so the named form is load-bearing.
//
//nolint:gochecknoglobals // the 0.1.8 named Enter sequence (read-only)
var armProbeEnter = []string{"KEY_ENTER"}

// armOutcome classifies the arming gate's result (the verdict semantics of
// the 2026-10-12 live smoke: the engine→stale-context desync is the
// TERM-01 defect honestly measured — a completed trace, not a failure).
type armOutcome int

const (
	// armWindow: the probe landed in the fixture pty — window delivery
	// proven (the engine delta is a diagnostic, present or not).
	armWindow armOutcome = iota
	// armStaleContext: the engine traced the probe but the fixture pty
	// stayed empty — the engine delivered/forwarded the keys to a stale IM
	// context instead of the freshly focused XIM fixture. A legitimate
	// measured outcome: record it, keep the batch going.
	armStaleContext
	// armDead: the probe reached NEITHER the engine trace NOR the fixture
	// pty — the injection itself broke. The only genuine harness FAIL.
	armDead
)

// classifyArm decides the gate's outcome from its two observables: the
// fixture pty's probe presence wins (window delivery), then the engine's
// probe-class trace (stale-context desync), then nothing (dead injection).
func classifyArm(ptyHasProbe bool, engineDelta int) armOutcome {
	switch {
	case ptyHasProbe:
		return armWindow
	case engineDelta > 0:
		return armStaleContext
	default:
		return armDead
	}
}

// armReport is the gate's classified outcome with its evidence counts —
// what the verdict rows print (the evidence-honesty discipline).
type armReport struct {
	outcome     armOutcome
	engineDelta int // probe-class (keyval 0x31) records the engine gained
	tries       int // attempts actually made
}

const (
	// captureFixedArgs is the fixed argument prefix of the capture spawn
	// (--config-file, path, --config, font_size, value, start,
	// --always-new-process, --): the config flags are TOP-LEVEL — after
	// the start subcommand the CLI rejects them outright (the 21:22 live
	// finding: "unexpected argument", no pane ever ran).
	captureFixedArgs = 8

	// fixtureFontSize is the big-font override that makes the fixture
	// window obvious from across the room (the ergonomics amendment).
	fixtureFontSize = "20"

	// Focus-wait knobs (the unified observation layer): the fixture is
	// engaged when the daemon focus_in lands — a terminal-child window
	// usually self-focuses on map, the owner's click fires it when it does
	// not. spikeClickWait bounds that wait — 30 minutes: a timeout is a
	// pause, not a failure (the console note names the awaited surface;
	// the 2026-10-12 directive removed the sound cues — silence is the
	// default); manualFocusSettle lets map-time focus bounces pass before
	// the baseline is taken.
	spikeClickWait    = 30 * time.Minute
	manualFocusSettle = 3 * time.Second

	// daemonReleaseWait bounds the pre-spawn wait for the component name to
	// become claimable — generous (60s+) because the previous case's daemon
	// releases slowly (ibus-daemon + D-Bus teardown lag; the live finding
	// of 2026-10-13: term04's daemon died with "another goswitchd instance
	// already registered" 17 s into the previous case's shutdown, and
	// spike-wezterm timed out of the old fixed 15 s budget into a 0-byte
	// journal).
	daemonReleaseWait = 90 * time.Second

	// focusInMark is the daemon's client-context focus record — THE focus
	// signal of the unified observation layer (the 2026-10-11/12 findings:
	// the daemon focus pair fires reliably on real focus while the AT-SPI
	// witness stays pinned to gnome-shell and the caps mint is not
	// guaranteed; a terminal-child window usually self-focuses on map, and
	// the owner's click fires the pair when it does not).
	focusInMark = `"msg":"focus_in"`

	// paneReadyWait bounds the wait for the pane program's readiness marker
	// (the A3 PROG-passing proof: `wezterm start -- PROG` must actually run
	// the program — the 21:22 live finding: --config-file AFTER the
	// subcommand is rejected outright, so no pane ever ran).
	paneReadyWait = 15 * time.Second

	// chordSettle is the settle budget after a select-chord before the
	// PRIMARY read-back (the research's spike-pinned ~50-100 ms knob, with
	// headroom for the yank session).
	chordSettle = 400 * time.Millisecond

	// primarySentinelMark is the PRIMARY pre-state marker (Pitfall 5):
	// non-secret by design, asserted to CHANGE to the typed corpus prefix.
	primarySentinelMark = "GS-SPIKE-SENTINEL"

	// flipWait bounds the wait for the daemon's flip pair (focus_out en +
	// focus_in ru) after the Right Shift tap.
	flipWait = 5 * time.Second

	// clickNoteEvery is the await-progress console cadence of the focus
	// wait (the owner sees the surface's name in the output while waiting).
	clickNoteEvery = 15 * time.Second

	// spikeSmokeWatchdog is the spike-smoke case deadline: one 30-minute
	// focus wait, daemon spawn and probe round with headroom.
	spikeSmokeWatchdog = 45 * time.Minute

	// flipRounds is the oracle's round count — the two-greens contract
	// (Pitfall 1): a single green is void by house rule.
	flipRounds = 2

	// Batch-round watchdogs: each round carries one 30-minute focus wait
	// plus the probe set (flip: two such rounds).
	flipWatchdog   = 75 * time.Minute
	recipeWatchdog = 45 * time.Minute
	capsWatchdog   = 45 * time.Minute

	// autocorrectWatchdog rides the frozen TERM-04 negative under the
	// unified budgets (zenity + wezterm waits).
	autocorrectWatchdog = 75 * time.Minute
)

// waitPaneReady waits for the pane program's readiness marker file — the
// A3 PROG-passing proof (the pane must actually run the program before any
// focus gating means anything).
func waitPaneReady(ctx context.Context, caseName, readyPath string) error {
	deadline := time.Now().Add(paneReadyWait)
	for {
		if _, err := os.Stat(readyPath); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s: pane program never created its readiness marker %s within %v"+
				" (PROG passing broken?)", caseName, readyPath, paneReadyWait)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// noteGate is the shared await progress of the focus/launch waits: the
// console note every clickNoteEvery until the wait resolves (the owner
// sees the surface's name in the output while waiting). The former sound
// cue is gone at the owner's request (2026-10-12: the beeps fired while
// merely awaiting the paste — silence is the default); the notes carry the
// progress alone.
type noteGate struct {
	nextNote time.Time
}

func newNoteGate() noteGate {
	return noteGate{nextNote: time.Now().Add(clickNoteEvery)}
}

// due reports whether the console note is due, re-arming the interval.
func (g *noteGate) due() bool {
	if time.Now().After(g.nextNote) {
		g.nextNote = time.Now().Add(clickNoteEvery)

		return true
	}

	return false
}

// focusBaseline snapshots the focus_in record count — the SPAWN baseline
// of the focus detector. It must be captured BEFORE the fixture spawns: a
// terminal-child window usually self-focuses within milliseconds and an
// owner click lands seconds later, and a baseline taken after the wait
// starts would sit above the very pairs it waits for (the 2026-10-14 live
// finding: the zenity control's pairs at T+1 s sat under the post-settle
// baseline and the case hung on its own marker for the whole budget).
func (s *stand) focusBaseline() int {
	return s.countSub(focusInMark)
}

// waitFocusInFrom is THE focus detector of the unified observation layer
// (the only one): the surface is proven engaged when a fresh daemon
// focus_in lands ABOVE the caller-captured spawn baseline, read from the
// pinned -log journal. No AT-SPI, no caps fingerprint, no click REQUIRED —
// the pair arriving IS the proceed signal; the 30-minute budget is the
// only wait (a timeout is a pause, never a failure of the experiment
// protocol). The await is logged every 15 s so a timeout names the
// surface that was waiting.
//
// Plan 10-08 (the fresh-session live finding of 2026-10-13): a freshly
// booted stand session does not hand map-focus to spawned gnome-terminal
// windows, so the gate PROVOKES the pair machine-side — after the first
// quiet seconds it drives a ydotool center click (the manual-activation
// equivalent, automated; the venue's stock kernel keeps background windows
// keyboard-alive). The honesty is untouched: the click is only a
// provocation, the proceed signal stays the daemon focus_in above the
// baseline, and every nudge is logged.
func (s *stand) waitFocusInFrom(ctx context.Context, caseName, surface string, focusBase int) error {
	deadline := time.Now().Add(spikeClickWait)
	notes := newNoteGate()
	nextNudge := time.Now().Add(focusNudgeAfter)
	for {
		if s.countSub(focusInMark) > focusBase {
			fmt.Printf("%s: %s engaged (daemon focus_in above the spawn baseline) — proceeding\n",
				caseName, surface)

			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s: awaiting focus for %s: no daemon focus_in within %v"+
				" above the spawn baseline (focus_in since baseline: %d; a click on the window would fire it)",
				caseName, surface, spikeClickWait, s.countSub(focusInMark)-focusBase)
		}
		if notes.due() {
			fmt.Printf("%s: awaiting daemon focus_in for %s (map focus usually self-fires; a click fires it)\n",
				caseName, surface)
		}
		if time.Now().After(nextNudge) {
			nextNudge = time.Now().Add(focusNudgeEvery)
			if err := nudgeFocusClick(ctx); err != nil {
				fmt.Printf("%s: focus nudge failed (continuing to wait): %v\n", caseName, err)
			}
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// focusNudgeAfter is the quiet time before the gate's first machine click,
// focusNudgeEvery the cadence after it: long enough that a self-firing map
// focus is never disturbed, short enough that the gate's own budget
// absorbs several nudges. nudgeHalfDivisor halves the screen dimensions to
// the center; the xrandr probe reads the live geometry of the stand's
// Xwayland display.
const (
	focusNudgeAfter  = 5 * time.Second
	focusNudgeEvery  = 5 * time.Second
	nudgeHalfDivisor = 2
	nudgeXrandrProbe = "xrandr --current 2>/dev/null | grep -o '[0-9]*x[0-9]* screen'"
)

// nudgeFocusClick drives one ydotool center click — the machine form of
// the venue's manual-activation step (plan 10-08): mutter maps spawned
// windows centered, so the click lands on the fixture the gate is waiting
// for. The screen size is read from xrandr when it answers (the Xwayland
// display of the stand), the documented 1280x800 otherwise. Errors are the
// caller's to log and survive: a failed nudge must not fail the gate.
func nudgeFocusClick(ctx context.Context) error {
	x, y := 640, 400
	if out, err := runCmd(ctx, "sh", "-c", nudgeXrandrProbe); err == nil {
		var w, h int
		if _, serr := fmt.Sscan(strings.ReplaceAll(out, "x", " "), &w, &h); serr == nil && w > 0 && h > 0 {
			x, y = w/nudgeHalfDivisor, h/nudgeHalfDivisor
		}
	}
	if _, err := runCmd(ctx, "ydotool", "mousemove", "--delay", "100",
		strconv.Itoa(x), strconv.Itoa(y)); err != nil {
		return fmt.Errorf("mousemove: %w", err)
	}
	if _, err := runCmd(ctx, "ydotool", "click", "--delay", "100", "1"); err != nil {
		return fmt.Errorf("click: %w", err)
	}
	fmt.Printf("focus nudge: ydotool center click at %dx%d\n", x, y)

	return nil
}

// weztermFixturePath resolves a spike fixture config under test/e2e/fixtures.
func weztermFixturePath(name string) string {
	return filepath.Join("test", "e2e", "fixtures", name)
}

// weztermCapture is one dedicated wezterm window running a pty-capturing
// program. The launcher process is held for the teardown — and the gui
// process is reaped by the fixture-name scope, because the live finding
// (2026-10-10 21:25) is that wezterm-gui is a separate pid that SURVIVES
// the launcher's kill (an orphan window would strand the owner's desktop).
type weztermCapture struct {
	cmd     *exec.Cmd
	fixture string // the config file name — the gui-reap scope (never matches the owner's own instance)
}

// startWeztermCapture spawns wezterm with a stand-owned config and a
// pty-capturing PROG. --config-file rides BEFORE the start subcommand (a
// top-level flag; the 21:22 live finding pinned the rejection otherwise);
// -- makes the program boundary explicit.
//
//nolint:noctx // the window must outlive the case context: the caller closes it by pid
func startWeztermCapture(_ context.Context, configPath, clickLabel string, prog ...string) (*weztermCapture, error) {
	args := make([]string, 0, len(prog)+captureFixedArgs)
	args = append(args, "--config-file", configPath,
		"--config", "font_size="+fixtureFontSize,
		"start", "--always-new-process", "--")
	// The pane sets its own title before anything else — the window reads
	// it (the case's label, unmissable on the desktop).
	if len(prog) >= 3 && prog[0] == "bash" && prog[1] == "-c" {
		prog[2] = "printf '\\033]2;" + clickLabel + "\\007'; " + prog[2]
	}
	args = append(args, prog...)
	cmd := exec.Command(weztermBin, args...)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s capture: %w", weztermBin, err)
	}

	return &weztermCapture{cmd: cmd, fixture: filepath.Base(configPath)}, nil
}

// close terminates the captured window: the launcher is killed by pid (the
// closeWezterm discipline) and the gui is reaped under the fixture-name
// scope — the pattern matches only stand windows spawned with OUR config
// file, never the owner's own instance. A no-op once closed.
func (w *weztermCapture) close() {
	if w == nil {
		return
	}
	if w.cmd != nil && w.cmd.Process != nil {
		_ = w.cmd.Process.Kill()
		_ = w.cmd.Wait()
		w.cmd = nil
	}
	if w.fixture == "" {
		return
	}
	//nolint:noctx // best-effort teardown reap; pkill outlives the case context harmlessly
	reap := exec.Command("pkill", "-9", "-f", "--", w.fixture)
	_ = reap.Run()
}

// readPtyHexArmed is the oracle read of a captured pty — the encoding
// discipline: hex and verdicts go to journals and commits, raw captures
// stay in tmp files and are never committed (Pitfall 6) — for a capture
// the arming gate has written into: the gate's probe is stripped from the
// head before hexlification (stripArmingProbe), so the verdicts compare
// exactly what the round under test typed — never the gate's own probe. A
// capture without the probe is a named harness failure (the gate
// guarantees it before any oracle read), never a plausible wrong hex.
func readPtyHexArmed(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read pty capture %s: %w", path, err)
	}
	if !strings.Contains(string(data), armProbeText) {
		return "", fmt.Errorf("read pty capture %s: no arming probe %q — the capture is not armed", path, armProbeText)
	}

	return hex.EncodeToString(stripArmingProbe(data)), nil
}

// armGate is the formal arming gate of the wezterm spike: inject the
// arming probe line (armProbeText + Enter) and classify where it landed —
// classifyArm pins the semantics. Bounded retries with settle — the race
// is a live finding (autocorrect-terminal-silent, 2026-10-02), so the gate
// retries instead of assuming; a stale-context signature stops the retries
// after one clean demonstration (every extra try just types more junk into
// the stale context, the 2026-10-12 "1234221234…" finding).
//
// The three outcomes (the R1 + live-smoke findings):
//   - armWindow: the probe landed in the FIXTURE's own pty capture — the
//     injection-side proof that ydotool → compositor → focused-window
//     reaches the very surface the oracle reads;
//   - armStaleContext: the engine traced the probe but the fixture pty
//     stayed empty — engine focus is desynchronized from real compositor
//     focus for XIM clients and the keys went to a stale IM context. The
//     TERM-01 defect mechanism, honestly measured — a verdict, not a
//     failure;
//   - armDead: neither signal — the injection itself broke. The only FAIL.
//
// The daemon journal is evidence and diagnostic here, never by itself a
// gate: the R1 finding (engine deaf while the window showed input) and the
// 2026-10-12 finding (engine tracing while the window got nothing) are the
// two halves of the same desync, and neither journal state alone proves
// delivery.
func (s *stand) armGate(ctx context.Context, caseName, ptyPath string) (armReport, error) {
	base := s.countSub(probeKeyMark)
	for try := 1; try <= armGateTries; try++ {
		if err := s.injectText(ctx, armProbeText); err != nil {
			return armReport{outcome: armDead, tries: try}, fmt.Errorf("%s: arm gate probe: %w", caseName, err)
		}
		if err := s.injectKeys(ctx, armProbeEnter...); err != nil {
			return armReport{outcome: armDead, tries: try}, fmt.Errorf("%s: arm gate enter: %w", caseName, err)
		}
		if werr := waitForArmingProbe(ctx, ptyPath, armGateWait); werr == nil {
			report := armReport{outcome: armWindow, engineDelta: s.countSub(probeKeyMark) - base, tries: try}
			fmt.Printf("%s: arm gate passed — probe landed in the fixture pty %s"+
				" (probe-key records delta %d)\n", caseName, ptyPath, report.engineDelta)

			return report, nil
		}
		engineDelta := s.countSub(probeKeyMark) - base
		if outcome := classifyArm(false, engineDelta); outcome == armStaleContext {
			fmt.Printf("%s: arm gate: engine traced the probe (probe-key records delta %d)"+
				" but the fixture pty %s stayed empty — engine→stale-context desync\n",
				caseName, engineDelta, ptyPath)

			return armReport{outcome: armStaleContext, engineDelta: engineDelta, tries: try}, nil
		}
		if err := sleepCtx(ctx, armGateSettle); err != nil {
			return armReport{outcome: armDead, engineDelta: engineDelta, tries: try}, err
		}
	}

	return armReport{outcome: armDead, tries: armGateTries},
		fmt.Errorf("%s: arming gate: probe %q reached neither the engine trace nor the fixture pty %s after %d tries",
			caseName, armProbeText, ptyPath, armGateTries)
}

// waitForArmingProbe blocks until the fixture's pty capture contains the
// arming probe line — the file-based sibling of waitForNew (the fixture
// pty capture is the arming gate's observable; the daemon journal is not).
// The poll interval and the abortable-sleep discipline are shared with the
// journal waits.
func waitForArmingProbe(ctx context.Context, path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if data, err := os.ReadFile(path); err == nil && strings.Contains(string(data), armProbeText) {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout after %v: file %s has no %q", timeout, path, armProbeText)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return fmt.Errorf("wait for %q in %s aborted: %w", armProbeText, path, err)
		}
	}
}

// stripArmingProbe removes the arming-probe bytes from the head of a pty
// capture: the gate writes strictly before the round's own content (arm
// retries included), so every arming byte is prefix material — a repeated
// run of the probe followed by a line ending. The ending differs with the
// tty mode (canonical flush: \n; raw passthrough: \r), so the strip eats
// the probe plus any run of CR/LF. Nothing after the run is touched: the
// round's own content never starts with the digit probe.
func stripArmingProbe(data []byte) []byte {
	rest := string(data)
	for {
		trimmed := false
		if after, ok := strings.CutPrefix(rest, armProbeText); ok {
			rest, trimmed = after, true
		}
		for _, eol := range []string{"\r\n", "\n", "\r"} {
			if after, ok := strings.CutPrefix(rest, eol); ok {
				rest, trimmed = after, true

				break
			}
		}
		if !trimmed {
			break
		}
	}

	return []byte(rest)
}

// primarySentinel sets a known PRIMARY pre-state (Pitfall 5: the sentinel
// discipline — a chord probe asserts the sentinel CHANGED to the expected
// prefix, never a stale read).
func primarySentinel(ctx context.Context) error {
	//nolint:noctx // best-effort clipboard write; bounded by the caller's step
	cmd := exec.Command("bash", "-c", "printf %s "+primarySentinelMark+" | wl-copy --primary")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("set PRIMARY sentinel: %w", err)
	}

	return sleepCtx(ctx, chordSettle)
}

// primaryReadNoNewline reads PRIMARY without the trailing newline (the
// `-n` discipline); an empty read is a defined state, not an error
// (Pitfall 5) — some wl-clipboard versions exit non-zero on empty PRIMARY.
func primaryReadNoNewline(ctx context.Context) string {
	out, err := runCmd(ctx, "wl-paste", "--primary", "--no-newline")
	if err != nil {
		return ""
	}

	return out
}

// probeTransportIdentity reads the transport facts of the running
// wezterm-gui process (P0): the WAYLAND_DISPLAY/DISPLAY environ pair and
// the xlsclients listing — XWayland client vs native Wayland. The first
// gui pid wins (the stand owns the only fixture windows; the owner's own
// instance, if any, uses the same binary — the pid list is reported for
// the journal either way).
func probeTransportIdentity(ctx context.Context) (waylandDisplay, display, xls string, pids []string, err error) {
	out, perr := runCmd(ctx, "pgrep", "-a", "wezterm-gui")
	if perr != nil {
		return "", "", "", nil, fmt.Errorf("probe transport: no wezterm-gui process: %w", perr)
	}
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		pids = append(pids, strings.Fields(line)[0])
	}
	envData, rerr := os.ReadFile("/proc/" + pids[0] + "/environ")
	if rerr != nil {
		return "", "", "", pids, fmt.Errorf("probe transport: read environ: %w", rerr)
	}
	for _, kv := range strings.Split(string(envData), "\x00") {
		switch {
		case strings.HasPrefix(kv, "WAYLAND_DISPLAY="):
			waylandDisplay = kv
		case strings.HasPrefix(kv, "DISPLAY="):
			display = kv
		}
	}
	xls, err = runCmd(ctx, "xlsclients")

	return waylandDisplay, display, xls, pids, err
}

// commitMonitor is the P3 wire channel: dbus-monitor on the IBus socket
// filtered to CommitText signals (the plan's sanctioned fallback — the
// daemon logs no commit records). Counts only: the capture buffer stays in
// memory, never in journals or commits (Pitfall 6).
type commitMonitor struct {
	cmd *exec.Cmd
	out *bytes.Buffer
}

// startCommitMonitor begins watching CommitText signals on the IBus
// socket.
//
//nolint:noctx // the monitor must outlive individual case steps; the caller closes it
func startCommitMonitor() (*commitMonitor, error) {
	addr, err := engine.Discover()
	if err != nil {
		return nil, fmt.Errorf("commit monitor: discover ibus: %w", err)
	}
	out := &bytes.Buffer{}
	cmd := exec.Command("dbus-monitor", "--address", addr,
		"type='signal',interface='org.freedesktop.IBus.Engine',member='CommitText'")
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("commit monitor: start dbus-monitor: %w", err)
	}

	return &commitMonitor{cmd: cmd, out: out}, nil
}

// count returns the CommitText signals seen so far.
func (m *commitMonitor) count() int {
	if m == nil {
		return 0
	}

	return strings.Count(m.out.String(), "member=CommitText")
}

// close stops the monitor and reaps the process.
func (m *commitMonitor) close() {
	if m == nil || m.cmd == nil || m.cmd.Process == nil {
		return
	}
	_ = m.cmd.Process.Kill()
	_ = m.cmd.Wait()
	m.cmd = nil
}

// surfaceHandle is the round's acquired fixture: the pty ground-truth path
// plus the teardown closure (spawn kill).
type surfaceHandle struct {
	ptyPath string
	close   func()
}

// openFixture acquires the round's fixture surface — ONE mode: the case
// always spawns the dedicated wezterm window itself (the case runs from
// the owner's interactive terminal, the only supported invocation). The
// pane's readiness is proven, then the surface is engaged by the daemon
// focus_in (waitFocusInFrom — map focus usually self-fires, the owner's click
// fires it when it does not); the caller runs the arm gate before any
// injection.
func (s *stand) openFixture(
	ctx context.Context, caseName, configPath, clickLabel, tag string, afterReady func() error,
) (*surfaceHandle, error) {
	readyPath := filepath.Join(s.tmpDir, tag+"-ready")
	ptyPath := filepath.Join(s.tmpDir, tag+"-in")
	focusBase := s.focusBaseline()
	wcap, err := startWeztermCapture(ctx, configPath, clickLabel,
		"bash", "-c", "echo ready > "+readyPath+"; cat > "+ptyPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", caseName, err)
	}
	if err := waitPaneReady(ctx, caseName, readyPath); err != nil {
		wcap.close()

		return nil, fmt.Errorf("%s: %w", caseName, err)
	}
	fmt.Printf("%s: fixture spawned (%s) — pane ready\n", caseName, clickLabel)
	if afterReady != nil {
		if err := afterReady(); err != nil {
			wcap.close()

			return nil, err
		}
	}
	fmt.Printf("%s: awaiting daemon focus_in for %s (map focus usually self-fires; a click fires it; budget %v)\n",
		caseName, clickLabel, spikeClickWait)
	if err := s.waitFocusInFrom(ctx, caseName, clickLabel, focusBase); err != nil {
		wcap.close()

		return nil, fmt.Errorf("%s: %w", caseName, err)
	}

	return &surfaceHandle{ptyPath: ptyPath, close: func() { wcap.close() }}, nil
}
