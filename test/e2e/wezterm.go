package main

// The wezterm spike fixture (plan 09-02): a dedicated terminal window
// (--always-new-process, never a delegation to an owner instance — the
// chromium-profile isolation lesson) running a pty-capturing program, plus
// the probe primitives the spike cases drive: pty hex read-back, the
// arming gate of the wezterm IME race, and the manual-activation gates
// (owner-click focus marker + map-time auto-grant watch — the
// 2026-10-09 protocol). (The PRIMARY-sentinel and flip-round primitives
// join with the cases that drive them — the strict unused gate forbids
// dead fixture code.)

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	// vteBin is the GNOME Terminal surface of the caps round — a GTK3 a11y
	// citizen (the witness CAN name it) with the false-advertising 0x29 caps
	// class.
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

	// keyTraceMark is the -debug key-trace record prefix the arming gate and
	// the delivery counters read.
	keyTraceMark = `"msg":"key"`
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
	// armGateWait bounds the wait for ONE new key-trace record per attempt.
	armGateWait = 2 * time.Second

	// captureFixedArgs is the fixed argument prefix of the capture spawn
	// (--config-file, path, --config, font_size, value, start,
	// --always-new-process, --): the config flags are TOP-LEVEL — after
	// the start subcommand the CLI rejects them outright (the 21:22 live
	// finding: "unexpected argument", no pane ever ran).
	captureFixedArgs = 8

	// fixtureFontSize is the big-font override that makes the fixture
	// window obvious from across the room (the ergonomics amendment).
	fixtureFontSize = "20"

	// Manual-activation knobs (2026-10-09 protocol, owner-authorized; the
	// 2026-10-10 ergonomics amendment): the session's map-time focus grant
	// is broken (the journal's environment block), so fixture windows are
	// focused by the OWNER'S CLICK — the production gesture. spikeClickWait
	// bounds the per-surface wait for that click — 30 minutes: a timeout is
	// a pause, not a failure (the owner clicks when the sound cue fires,
	// no sustained watching); manualFocusSettle lets map-time focus bounces
	// pass before the click-marker base is taken.
	spikeClickWait     = 30 * time.Minute
	manualFocusSettle  = 3 * time.Second
	mapAutoGrantBudget = 12 * time.Second

	// Click-request sound cue (the ergonomics amendment): while a surface
	// awaits its click the cue fires every cueEvery; the click is the
	// cue's own stop.
	cueEvery = 25 * time.Second
	cueSound = "/usr/share/sounds/freedesktop/stereo/alarm-clock-elapsed.oga"

	// weztermCapsMark is the wezterm client fingerprint in the SetCapabilities
	// record — caps 0x9 (preedit+focus, NO surrounding bit; the 2026-10-02
	// live finding). The fingerprint, not a bare focus pair, is the click
	// marker: a pair (and even a fresh non-wezterm context) can come from
	// the owner's own windows re-engaging — the 21:17/21:22 live findings
	// false-passed twice on the owner's terminal while the fixture had not
	// even mapped. VTE false-advertises 0x29 — the values distinguish the
	// client classes.
	weztermCapsMark = `"msg":"capabilities","caps":"0x9"`
	// capsMark is the unqualified SetCapabilities record (diagnostics only:
	// the timeout error reports what DID land).
	capsMark = `"msg":"capabilities"`

	// focusInMark is the daemon's client-context focus record — the click
	// confirmation of the attach mode (2026-10-11 finding: the owner's
	// click fires the daemon focus pair reliably while the AT-SPI witness
	// stays pinned to gnome-shell and the caps mint is not guaranteed; the
	// arm gate remains the injection-side proof).
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

	// clickNoteEvery is the await-progress log cadence of waitWeztermClick
	// (the owner sees the window's name in the output while waiting).
	clickNoteEvery = 15 * time.Second

	// spikeSmokeWatchdog is the spike-smoke case deadline: one 30-minute
	// click wait plus the auto-grant watch, daemon spawn and probe round
	// with headroom.
	spikeSmokeWatchdog = 45 * time.Minute

	// flipRounds is the oracle's round count — the two-greens contract
	// (Pitfall 1): a single green is void by house rule.
	flipRounds = 2

	// Batch-round watchdogs: each round carries one 30-minute click wait
	// plus the probe set (flip: two such rounds).
	flipWatchdog   = 75 * time.Minute
	recipeWatchdog = 45 * time.Minute
	capsWatchdog   = 45 * time.Minute

	// autocorrectWatchdog rides the frozen TERM-04 negative under the
	// manual-activation budgets (zenity + wezterm waits).
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

// watchWeztermAutoGrant observes whether the freshly mapped fixture takes
// keyboard focus ON ITS OWN (no click, no injection) within the budget —
// the per-terminal map-time auto-grant datum (2026-10-09 protocol item 3).
// The marker is the wezterm caps fingerprint (the context engaged the
// engine); "" means no engagement within the budget. The caller records it
// as a journal row.
func (s *stand) watchWeztermAutoGrant(ctx context.Context, caseName string, budget time.Duration) (string, error) {
	base := s.countSub(weztermCapsMark)
	deadline := time.Now().Add(budget)
	for {
		if s.countSub(weztermCapsMark) > base {
			return "caps-0x9", nil
		}
		if time.Now().After(deadline) {
			fmt.Printf("%s: wezterm did NOT auto-grant focus within %v (witness %q)\n",
				caseName, budget, s.focusWitnessSafe(ctx))

			return "", nil
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return "", err
		}
	}
}

// playCue fires the click-request sound (non-blocking; the player is
// reaped in the background).
func playCue() {
	//nolint:noctx // fire-and-forget cue; reaped in a goroutine
	cmd := exec.Command("canberra-gtk-play", "-f", cueSound)
	if err := cmd.Start(); err != nil {
		return
	}
	go func() { _ = cmd.Wait() }()
}

// cueGate is the shared await progress for the click gates: the note every
// clickNoteEvery on the console and the sound every cueEvery until the
// click lands (the sound IS the owner's prompt — no watching required).
type cueGate struct {
	nextNote time.Time
	nextCue  time.Time
}

func newCueGate() cueGate {
	return cueGate{nextNote: time.Now().Add(clickNoteEvery), nextCue: time.Now().Add(cueEvery)}
}

// due fires the sound when its slot elapsed and reports whether the
// console note is due.
func (g *cueGate) due() bool {
	note := false
	if time.Now().After(g.nextNote) {
		note = true
		g.nextNote = time.Now().Add(clickNoteEvery)
	}
	if time.Now().After(g.nextCue) {
		playCue()
		g.nextCue = time.Now().Add(cueEvery)
	}

	return note
}

// focusWitnessSafe is the witness read that never fails the caller — a
// diagnostic read for log lines (errors render as the empty string).
func (s *stand) focusWitnessSafe(ctx context.Context) string {
	w, err := s.focusWitness(ctx)
	if err != nil {
		return ""
	}

	return w
}

// waitWeztermClick gates an injection on the owner's click (manual
// activation): the fixture is proven focused when a fresh wezterm caps
// fingerprint (caps 0x9 — the client-class context engagement) lands above
// the post-settle base. Unbound markers (bare focus pairs, foreign caps)
// do NOT qualify — the 21:17/21:22 live findings false-passed on the
// owner's own windows while the fixture had not even mapped. The await is
// logged every 15 s so a timeout names the window that was waiting.
func (s *stand) waitWeztermClick(ctx context.Context, caseName, surface string) error {
	if err := sleepCtx(ctx, manualFocusSettle); err != nil {
		return err
	}
	base := s.countSub(weztermCapsMark)
	baseAllCaps := s.countSub(capsMark)
	deadline := time.Now().Add(spikeClickWait)
	cue := newCueGate()
	for {
		if s.countSub(weztermCapsMark) > base {
			playCue()
			fmt.Printf("%s: %s focused via caps-0x9 context fingerprint — owner click registered\n",
				caseName, surface)

			return nil
		}
		if time.Now().After(deadline) {
			// Reached only when the fingerprint never landed — every caps
			// record since arming is therefore foreign (the diagnostic:
			// what DID engage instead).
			return fmt.Errorf("%s: awaiting owner click on %s: no caps-0x9 fingerprint within %v"+
				" (foreign caps records since arming: %d; witness %q)",
				caseName, surface, spikeClickWait, s.countSub(capsMark)-baseAllCaps, s.focusWitnessSafe(ctx))
		}
		if cue.due() {
			fmt.Printf("%s: awaiting owner click on %s (witness %q)\n", caseName, surface, s.focusWitnessSafe(ctx))
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
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
	// it (the ergonomics amendment: «КЛИК N из M — <case>», unmissable).
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

// readPtyHex returns the captured pty bytes hexlified — the encoding
// discipline: hex and verdicts go to journals and commits, raw captures
// stay in tmp files and are never committed (Pitfall 6).
func readPtyHex(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read pty capture %s: %w", path, err)
	}

	return hex.EncodeToString(data), nil
}

// armGate is the formal arming gate of the wezterm IME race: tap a bare
// modifier (a modifier writes no pty bytes) and demand ONE new key-trace
// record; bounded retries with settle — the race is a live finding
// (autocorrect-terminal-silent, 2026-10-02), so the gate retries instead of
// assuming.
func (s *stand) armGate(ctx context.Context, caseName string) error {
	for range armGateTries {
		base := s.countSub(keyTraceMark)
		if err := s.injectKeys(ctx, "Shift_L"); err != nil {
			return fmt.Errorf("%s: arm gate tap: %w", caseName, err)
		}
		if werr := s.waitForNew(ctx, keyTraceMark, base+1, armGateWait); werr == nil {
			return nil
		}
		if err := sleepCtx(ctx, armGateSettle); err != nil {
			return err
		}
	}

	return fmt.Errorf("%s: arming gate: no key-trace record after %d modifier taps", caseName, armGateTries)
}

// waitSurfaceClick gates an injection on the owner's click for a surface
// that the witness CAN name (the VTE — a GTK3 a11y citizen) OR whose
// context mint shows in the daemon log (any fresh caps record — the only
// new window on the stand during its round). Unbound focus pairs alone do
// NOT qualify (the 21:17/21:22 false-pass class).
func (s *stand) waitSurfaceClick(ctx context.Context, caseName, surface, witnessPrefix string) error {
	if err := sleepCtx(ctx, manualFocusSettle); err != nil {
		return err
	}
	baseAllCaps := s.countSub(capsMark)
	deadline := time.Now().Add(spikeClickWait)
	cue := newCueGate()
	for {
		witness, werr := s.focusWitness(ctx)
		named := werr == nil && witnessPrefix != "" && strings.HasPrefix(witness, witnessPrefix)
		minted := s.countSub(capsMark) > baseAllCaps
		if named || minted {
			playCue()
			marker := "fresh context mint"
			if named {
				marker = "a11y witness"
			}
			fmt.Printf("%s: %s focused via %s — owner click registered\n", caseName, surface, marker)

			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s: awaiting owner click on %s: no focus marker within %v (witness %q)",
				caseName, surface, spikeClickWait, witness)
		}
		if cue.due() {
			fmt.Printf("%s: awaiting owner click on %s (witness %q)\n", caseName, surface, witness)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
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

// ownerCapture is the attach-mode handle of an OWNER-LAUNCHED fixture:
// marker paths only — the process belongs to the owner (launch attribution,
// 2026-10-11: background-spawned windows are keyboard-dead on the drifted
// kernel; owner-launched ones type).
type ownerCapture struct {
	readyPath string
	ptyPath   string
}

// runbookCmd renders the owner's paste-ready launch command for a round
// (repo root cwd): the stand fixture, the big font, the labeled title, and
// the marker paths the case waits for.
func runbookCmd(configPath, clickLabel, readyPath, ptyPath string) string {
	return "wezterm --config-file " + configPath + " --config font_size=" + fixtureFontSize +
		" start --always-new-process -- " +
		"bash -c \"printf '\\033]2;" + clickLabel + "\\007'; touch " + readyPath + " " + ptyPath +
		"; echo ready > " + readyPath + "; cat > " + ptyPath + "\""
}

// awaitOwnerFixture is the attach-mode fixture acquisition: the case prints
// the runbook command, waits for the owner to LAUNCH it (the ready marker)
// and CLICK it (a daemon focus_in above the case-start baseline — the only
// working click signal on this session: the AT-SPI witness pins to
// gnome-shell for everything, and the caps mint is not guaranteed; the
// arm gate stays the injection-side proof). No spawn handle exists; the
// teardown is the pane's EOF.
func (s *stand) awaitOwnerFixture(
	ctx context.Context, caseName, configPath, clickLabel, tag string,
) (*ownerCapture, error) {
	readyPath := "/tmp/gs-owner-" + tag + "-ready"
	ptyPath := "/tmp/gs-owner-" + tag + "-in"
	focusBase := s.countSub(focusInMark)
	fmt.Printf("%s: OWNER-LAUNCH — вставить в СВОЙ терминал (repo root):\n  %s\n",
		caseName, runbookCmd(configPath, clickLabel, readyPath, ptyPath))

	deadline := time.Now().Add(spikeClickWait)
	cue := newCueGate()
	for {
		if _, err := os.Stat(readyPath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("%s: awaiting OWNER LAUNCH of %s: no ready marker within %v",
				caseName, clickLabel, spikeClickWait)
		}
		if cue.due() {
			fmt.Printf("%s: awaiting owner launch (ready marker %s)\n", caseName, readyPath)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return nil, err
		}
	}
	for {
		if s.countSub(focusInMark) > focusBase {
			playCue()
			fmt.Printf("%s: %s engaged (daemon focus_in above case-start baseline) — owner click registered\n",
				caseName, clickLabel)

			return &ownerCapture{readyPath: readyPath, ptyPath: ptyPath}, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("%s: awaiting owner click on %s: no daemon focus_in within %v"+
				" (focus_in since case start: %d; witness %q)",
				caseName, clickLabel, spikeClickWait,
				s.countSub(focusInMark)-focusBase, s.focusWitnessSafe(ctx))
		}
		if cue.due() {
			fmt.Printf("%s: awaiting owner click on %s (daemon focus_in is the click signal)\n", caseName, clickLabel)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return nil, err
		}
	}
}

// close delivers the pane's EOF (raw evdev codes: 29 left-ctrl, 32 d) so
// the owner-launched window closes itself — the pid-точечный discipline
// without a process handle.
func (o *ownerCapture) close(ctx context.Context, caseName string) {
	if o == nil {
		return
	}
	if _, err := runCmd(ctx, "ydotool", "key", "29:1", "32:1", "32:0", "29:0"); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: %s owner fixture EOF: %v (the window may need a manual close)\n", caseName, err)
	}
}

// surfaceHandle is the round's acquired fixture: the pty ground-truth path
// plus the teardown closure (spawn kill or owner-pane EOF).
type surfaceHandle struct {
	ptyPath string
	close   func()
}

// openFixture acquires the round's fixture surface: attach mode when the
// case runs -owner-launch (the owner pastes the runbook command and
// clicks — the attribution protocol), the direct spawn otherwise. Both
// paths end with the fixture engaged (caps fingerprint) and return the
// handle; the caller runs the arm gate before any injection.
func (s *stand) openFixture(
	ctx context.Context, caseName, configPath, clickLabel, tag string, afterReady func() error,
) (*surfaceHandle, error) {
	if s.cfg.ownerLaunch {
		oc, err := s.awaitOwnerFixture(ctx, caseName, configPath, clickLabel, tag)
		if err != nil {
			return nil, err
		}

		return &surfaceHandle{ptyPath: oc.ptyPath, close: func() { oc.close(ctx, caseName) }}, nil
	}

	readyPath := filepath.Join(s.tmpDir, tag+"-ready")
	ptyPath := filepath.Join(s.tmpDir, tag+"-in")
	wcap, err := startWeztermCapture(ctx, configPath, clickLabel,
		"bash", "-c", "echo ready > "+readyPath+"; cat > "+ptyPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", caseName, err)
	}
	if err := waitPaneReady(ctx, caseName, readyPath); err != nil {
		wcap.close()

		return nil, fmt.Errorf("%s: %w", caseName, err)
	}
	if afterReady != nil {
		if err := afterReady(); err != nil {
			wcap.close()

			return nil, err
		}
	}
	fmt.Printf("%s: fixture window mapped — awaiting owner click (%s)\n", caseName, clickLabel)
	if err := s.waitWeztermClick(ctx, caseName, clickLabel); err != nil {
		wcap.close()

		return nil, fmt.Errorf("%s: %w", caseName, err)
	}

	return &surfaceHandle{ptyPath: ptyPath, close: func() { wcap.close() }}, nil
}
