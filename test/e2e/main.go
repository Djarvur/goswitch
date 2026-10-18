// Command e2e drives the goswitch live-session e2e stand (SPEC §7, TEST-02):
// it activates input surfaces, injects physical keys with ydotool and
// asserts on the goswitchd structured log. It is a CLI, not a go test: the
// observable behavior exists only inside a real GNOME Wayland session, and
// the exit-code contract (non-zero on FAIL) is the signal the mise tasks
// forward (D-09).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/Djarvur/goswitch/internal/uinject"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/Djarvur/goswitch/internal/engine"
)

// Stand timing knobs: every value is a behavioral constant of the stand,
// named once instead of sprinkled through call sites.
const (
	defaultPacingMs      = 40 // -pacing default: ms between injected keystrokes
	tapWindowFactor      = 3  // key taps use pacing×3 so a multi-tap fits the 300 ms window
	logPollInterval      = 25 * time.Millisecond
	killGrace            = 3 * time.Second   // SIGTERM → SIGKILL grace for the daemon
	teardownTimeout      = 15 * time.Second  // whole-teardown budget on Ctrl-C
	restoreVerifyTimeout = 10 * time.Second  // post-teardown gsettings readback budget
	cmdTimeout           = 15 * time.Second  // hard deadline for EVERY external CLI call
	caseTimeout          = 180 * time.Second // hard watchdog over the whole case body
	excerptLines         = 40                // daemon log tail dumped on FAIL
)

// gsettings input-source keys — the live-desktop state the stand snapshots
// and restores around every case (T-03-01: losing them loses the owner's
// layout).
const (
	gsettingsSchema = "org.gnome.desktop.input-sources"
	keySources      = "sources"
	keyCurrent      = "current"
)

// config is the stand's CLI surface.
type config struct {
	caseName      string
	pacing        int
	logPath       string
	helper        string
	matrixPath    string
	weztermConfig string
	// The selfdrive surface (the round-5 instrument): a dedicated branch
	// BELOW — it must never reach setupStand, the owner's host runs the
	// production daemon and the driver touches nothing but the focused
	// pane's marker discipline and the script log it watches.
	selfdrive    bool
	ssPaneLog    string
	ssPaneType   string
	ssOut        string
	ssSettle     time.Duration
	ssBaseline   int
	ssTiming     int
	ssTimingWait time.Duration
	ssDry        bool
}

// caseSpec is one registry entry: the case body plus the stand-shape flag.
// standalone cases (install-cycle) run WITHOUT the stand daemon — the
// single-instance ctlsvc name and the IBus registration belong to the
// daemon the case itself installs as a systemd unit. noSurfacePreflight
// skips the chromium/gte/gedit launch checks (the matrixPreflight
// philosophy: the surface drivers prove their binaries inside the cases
// that use them — a terminal-only case on a stand without those binaries
// must not fail their preflight). A non-zero watchdog overrides the default
// case deadline (perf: N × its repeat budget — raised, never removed: a
// live-session stand must never strand the owner's desktop).
type caseSpec struct {
	fn                 func(context.Context, *stand) error
	standalone         bool
	noSurfacePreflight bool
	watchdog           time.Duration
}

// desktopSnapshot is the live-desktop state the teardown restores.
type desktopSnapshot struct {
	sources    string
	current    string
	engineName string // empty when no global engine is set
}

// stand owns the per-run environment: the daemon subprocess, its log file,
// the stand's input surface and the captured desktop state.
type stand struct {
	cfg           config
	daemonBin     string
	tmpDir        string
	logPath       string
	logFile       *os.File
	daemon        *exec.Cmd
	standalone    bool // no stand daemon: the case owns the bus (install-cycle)
	unitWasActive bool // the setup lifecycle stopped an active user unit — teardown restores it
	zenity        *exec.Cmd
	zenityOut     *bytes.Buffer
	chromium      *exec.Cmd
	chromiumX11   *exec.Cmd // the x11/XWayland window-mode instance (04-05)
	gte           *exec.Cmd
	gedit         *exec.Cmd // the GTK3-generation editor instance (04-05)
	pwFixture     *exec.Cmd // the GTK4 password-entry fixture (06-07)
	pwFixtureOut  *bytes.Buffer
	wezterm       *exec.Cmd // the spawned terminal of the terminal-silent negative (06-07)
	snap          desktopSnapshot
	caseCfgPath   string // the matrix case's -config doc ("" until a reload step establishes it)
}

func main() {
	os.Exit(run())
}

// run wires the whole stand: registry, setup, preflight, the case and the
// caseListUsage is the -case help text: every registry name, "|" joined.
func caseListUsage() string {
	return "case to run: m1-gate | ibus-restart | kill9-survive | d01-probe | chromium-smoke | gte-smoke" +
		" | gedit-smoke | x11-smoke" +
		" | word-en-ru | word-after-space | word-ru-en | word-mixed | phrase-en-ru | phrase-mixed" +
		" | ladder-chromium | reset-escape | select-smoke | select-correct | select-clipboard" +
		" | combo-word-layout | layout-single | super-space-alive | macr-probe | macr-super-letter" +
		" | macr-per-app | ctl-smoke | switch-spike | two-source-flip | external-flip-sync" +
		" | flip-keystroke" +
		" | spike-smoke" +
		" | spike-wezterm | spike-caps | wezterm-flip | term04-negative" +
		" | autocorrect-fires | autocorrect-password-silent | autocorrect-terminal-silent" +
		" | autocorrect-blocklist-silent | menu-v2" +
		" | term-word-vte" +
		" | term-word-wezterm" +
		" | term-word-vte-flip" +
		" | term-cap-delivery" +
		" | term-word-drain" +
		" | install-cycle | perf | perf-autocorrect"
}

// parseFlags fills the stand's CLI surface from os.Args.
func parseFlags(cfg *config) {
	flag.StringVar(&cfg.caseName, "case", "", caseListUsage())
	flag.IntVar(&cfg.pacing, "pacing", defaultPacingMs,
		"milliseconds between injected keystrokes (raise on a loaded machine)")
	flag.StringVar(&cfg.logPath, "log", "", "daemon log path (default: a temp file removed in teardown)")
	flag.StringVar(&cfg.helper, "helper", "test/e2e/focus_helper.py",
		"path to the AT-SPI helper script (legacy pre-09-02 cases only;"+
			" the spike family never calls it)")
	flag.StringVar(&cfg.matrixPath, "matrix", "",
		"YAML case matrix to run (multi-doc cases; every case gets a fresh daemon)")
	flag.StringVar(&cfg.weztermConfig, "wezterm-config", "",
		"fixture config for wezterm-flip (default: gs-e2e-wezterm-xim.lua;"+
			" the wayland fixture re-targets the oracle without touching the registry)")
	flag.BoolVar(&cfg.selfdrive, "selfdrive", false,
		"run the host self-drive driver against the pane the owner opened"+
			" (no stand daemon; NOT in ci — live injections are orchestrator-invoked)")
	flag.StringVar(&cfg.ssPaneType, "pane-type", "",
		"selfdrive aux: type this text into the FOCUSED window via uinject and exit")
	flag.StringVar(&cfg.ssPaneLog, "pane-log", ssDefaultPaneLog,
		"selfdrive: the pane's `script -q` transcript the owner opened")
	flag.StringVar(&cfg.ssOut, "selfdrive-out", ssDefaultOut,
		"selfdrive: the dated verdict rows' append target")
	flag.DurationVar(&cfg.ssSettle, "settle", ssDefaultSettle,
		"selfdrive: the post-gesture settle before the choreography parse")
	flag.IntVar(&cfg.ssBaseline, "baseline-runs", ssDefaultRuns,
		"selfdrive: baseline battery runs (no manipulation)")
	flag.IntVar(&cfg.ssTiming, "timing-runs", ssDefaultRuns,
		"selfdrive: timing-axis battery runs")
	flag.DurationVar(&cfg.ssTimingWait, "timing-wait", ssDefaultTimingWait,
		"selfdrive: the typing→gesture wait of the timing axis")
	flag.BoolVar(&cfg.ssDry, "dry-run", false,
		"selfdrive: preconditions + parser self-check only — no injections, no writes")
	flag.Parse()
}

// run wires the whole stand: registry, setup, preflight, the case and the
// ALWAYS-run teardown. Returning an int keeps the exit-code contract in one
// place; defers are honored because run returns normally. The deferred
// finisher also machine-verifies the gsettings restore (plan 01-04, T-04-02)
// and may downgrade a PASS to FAIL through the named return.
func run() (exit int) {
	var cfg config
	parseFlags(&cfg)

	// The selfdrive branch owns its whole lifecycle: it runs against the
	// owner's production daemon on the live host — no stand daemon, no
	// desktop snapshot/restore may ever wrap it.
	if cfg.ssPaneType != "" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		inj, err := uinject.Open()
		if err != nil {
			fmt.Fprintf(os.Stderr, "pane-type: %v\n", err)

			return 1
		}
		defer func() { _ = inj.Close() }()

		const paneTypeDelay = 40
		if err := inj.Type(ctx, cfg.ssPaneType, paneTypeDelay); err != nil {
			fmt.Fprintf(os.Stderr, "pane-type: %v\n", err)

			return 1
		}

		return 0
	}

	if cfg.selfdrive {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		return runSelfDrive(ctx, cfg)
	}

	// The matrix branch owns its whole stand lifecycle: case isolation needs
	// a setupStand → case → teardown cycle PER CASE (02-06), so the single
	// shared stand of the -case path must not wrap it.
	if cfg.matrixPath != "" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		return runMatrixFile(ctx, cfg, cfg.matrixPath)
	}

	spec, err := pickCase(cfg.caseName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", cfg.caseName, err)

		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s, err := setupStand(ctx, cfg, spec.standalone)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL %s: stand setup: %v\n", cfg.caseName, err)

		return 1
	}
	defer func() {
		s.teardown()
		if rerr := s.verifyRestored(); rerr != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: teardown restore verification: %v\n", cfg.caseName, rerr)
			exit = 1
		}
	}()

	if err := preflight(ctx, s, spec.noSurfacePreflight); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", cfg.caseName, err)

		return 1
	}

	fmt.Printf("=== e2e case %s (daemon journal %s) ===\n", cfg.caseName, s.logPath)

	// Zero watchdog means the default: runCaseWatchdog resolves it
	// (watchdogLimit) — the same convention the matrix path relies on.
	if err := runCaseWatchdog(ctx, cfg.caseName, spec.fn, s, spec.watchdog); err != nil {
		s.printLogExcerpt()
		fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", cfg.caseName, err)

		return 1
	}

	fmt.Println("PASS " + cfg.caseName)

	return 0
}

// watchdogLimit resolves the case deadline: zero is the REQUEST for the
// default (the caseSpec.watchdog convention — "a non-zero watchdog
// overrides the default case deadline"), never a zero-length budget. The
// matrix runner passes 0 for exactly this meaning; the 04-04 signature
// change briefly turned it into a literal time.After(0) that failed every
// matrix case instantly (live finding of the first D-48 dispatch, run
// 35107406144) — the resolution lives HERE so no caller can re-trip it.
func watchdogLimit(explicit time.Duration) time.Duration {
	if explicit == 0 {
		return caseTimeout
	}

	return explicit
}

// runCaseWatchdog runs the case body under a hard deadline (the case's own
// watchdog override, or the default caseTimeout — see watchdogLimit) and
// cancels its context on expiry: a live-session stand must never strand
// the owner's desktop. A stuck case fails fast, names the case, and
// control falls through to the teardown contract (restore + machine
// verification).
func runCaseWatchdog(
	ctx context.Context,
	name string,
	fn func(context.Context, *stand) error,
	s *stand,
	limit time.Duration,
) error {
	limit = watchdogLimit(limit)
	caseCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- fn(caseCtx, s) }()

	select {
	case err := <-done:
		return err
	case <-time.After(limit):
		cancel()

		return fmt.Errorf("watchdog: case %q did not complete within %v", name, limit)
	}
}

// pickCase resolves the case name through the registry; the registry is
// built per call (no mutable globals) and lives in caseRegistry (the funlen
// budget keeps the two apart).
func pickCase(name string) (caseSpec, error) {
	spec, ok := caseRegistry()[name]
	if !ok {
		return caseSpec{}, fmt.Errorf("unknown or missing -case %q (registry: m1-gate, ibus-restart, kill9-survive,"+
			" d01-probe, chromium-smoke, gte-smoke, gedit-smoke, x11-smoke, word-en-ru,"+
			" word-after-space, word-ru-en, word-mixed, phrase-en-ru, phrase-mixed, ladder-chromium,"+
			" reset-escape, select-smoke, select-correct, select-clipboard, combo-word-layout,"+
			" layout-single, super-space-alive, macr-probe, macr-super-letter, macr-per-app,"+
			" ctl-smoke, switch-spike, two-source-flip, external-flip-sync, flip-keystroke,"+
			" spike-smoke, spike-wezterm, spike-caps, wezterm-flip, term04-negative,"+
			" autocorrect-fires, autocorrect-password-silent, autocorrect-terminal-silent,"+
			" autocorrect-blocklist-silent, menu-v2,"+
			" term-word-vte, term-negative-witnesses,"+
			" term-word-wezterm, term-word-vte-flip,"+
			" term-cap-delivery, term-word-drain,"+
			" install-cycle, perf, perf-autocorrect)", name)
	}

	return spec, nil
}

// caseRegistry is the case table of the stand (built per call — no mutable
// globals).
func caseRegistry() map[string]caseSpec {
	return map[string]caseSpec{
		"m1-gate":                     {fn: runM1Gate},
		"ibus-restart":                {fn: runIbusRestart},
		"kill9-survive":               {fn: runKill9Survive},
		"d01-probe":                   {fn: runD01Probe},
		"chromium-smoke":              {fn: runChromiumSmoke},
		"gte-smoke":                   {fn: runGTESmoke},
		"gedit-smoke":                 {fn: runGeditSmoke},
		"x11-smoke":                   {fn: runChromiumX11Smoke},
		"word-en-ru":                  {fn: runWordENRU},
		"word-after-space":            {fn: runWordAfterSpace},
		"word-ru-en":                  {fn: runWordRUEN},
		"word-mixed":                  {fn: runWordMixed},
		"phrase-en-ru":                {fn: runPhraseENRU},
		"phrase-mixed":                {fn: runPhraseMixed},
		"ladder-chromium":             {fn: runLadderChromium},
		"reset-escape":                {fn: runResetEscape},
		"select-smoke":                {fn: runSelectSmoke},
		"select-correct":              {fn: runSelectCorrect},
		"select-clipboard":            {fn: runSelectClipboard},
		"combo-word-layout":           {fn: runComboWordLayout},
		"layout-single":               {fn: runLayoutSingle},
		"super-space-alive":           {fn: runSuperSpaceAlive},
		"macr-probe":                  {fn: runMacrProbe},
		"macr-super-letter":           {fn: runMacrSuperLetter},
		"macr-per-app":                {fn: runMacrPerApp},
		"ctl-smoke":                   {fn: runCtlSmoke},
		"switch-spike":                {fn: runSwitchSpike, standalone: true},
		"two-source-flip":             {fn: runTwoSourceFlip, standalone: true},
		"external-flip-sync":          {fn: runExternalFlipSync, standalone: true},
		"flip-keystroke":              {fn: runFlipKeystroke, standalone: true},
		"spike-smoke":                 {fn: runSpikeSmoke, standalone: true, watchdog: spikeSmokeWatchdog},
		"spike-wezterm":               {fn: runSpikeWezterm, standalone: true, watchdog: recipeWatchdog},
		"term04-negative":             {fn: runTerm04Negative, watchdog: recipeWatchdog},
		"spike-caps":                  {fn: runSpikeCaps, standalone: true, watchdog: capsWatchdog},
		"wezterm-flip":                {fn: runWeztermFlip, standalone: true, watchdog: flipWatchdog},
		"autocorrect-fires":           {fn: runAutocorrectFires},
		"autocorrect-password-silent": {fn: runAutocorrectPasswordSilent},
		// autocorrect-terminal-silent drives zenity + wezterm only — the
		// chromium/gte/gedit preflight checks are surfaces it never uses
		// (the noSurfacePreflight philosophy of 10-05; the VM stand live
		// finding of plan 10-06: the negative is machine-reproducible
		// there now).
		"autocorrect-terminal-silent": {
			fn:                 runAutocorrectTerminalSilent,
			noSurfacePreflight: true,
			watchdog:           autocorrectWatchdog,
		},
		"autocorrect-blocklist-silent": {fn: runAutocorrectBlocklistSilent},
		"menu-v2":                      {fn: runMenuV2},
		"term-word-vte":                {fn: runTermWordVTE, noSurfacePreflight: true, watchdog: recipeWatchdog},
		"term-word-wezterm":            {fn: runTermWordWezterm, noSurfacePreflight: true, watchdog: recipeWatchdog},
		"term-word-vte-flip":           {fn: runTermWordVTEFlip, noSurfacePreflight: true, watchdog: recipeWatchdog},
		// term-cap-delivery (plan 10-13): six probe rows (3 probes × 2
		// clients), each probe on a fresh fixture with its own focus gate —
		// six spawn+gate cycles plus the daemon/config setup ride inside
		// the standard recipe budget.
		"term-cap-delivery": {fn: runTermCapDelivery, noSurfacePreflight: true, watchdog: recipeWatchdog},
		// term-word-drain (plan 10-17): the round-4 pty-drain matrix — one
		// full correction round per pacing combination (default/tight/loose
		// through the -debug override), each with a raw pty capture, so
		// three daemon restarts plus three fixture+observer rounds ride
		// inside the budget.
		"term-word-drain": {fn: runTermWordDrain, noSurfacePreflight: true, watchdog: recipeWatchdog},
		"term-negative-witnesses": {
			fn:                 runTermNegativeWitnesses,
			noSurfacePreflight: true,
			watchdog:           autocorrectWatchdog,
		},
		"install-cycle":    {fn: runInstallCycle, standalone: true},
		"perf":             {fn: runPerf, watchdog: perfSamples * perfRepeatBudget},
		"perf-autocorrect": {fn: runPerfAutocorrect, watchdog: perfSamples * perfRepeatBudget},
	}
}

// setupStand builds the daemon, snapshots the desktop and starts the
// daemon subprocess with -debug and both std streams into the log file
// (T-03-03: the log lives in $TMPDIR and is removed unless -log pinned
// it). A standalone case (install-cycle) skips the START — the daemon
// binary is still built, because the installed unit runs it.
func setupStand(ctx context.Context, cfg config, standalone bool) (*stand, error) {
	tmpDir, err := os.MkdirTemp("", "goswitch-e2e-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	s := &stand{cfg: cfg, tmpDir: tmpDir, daemonBin: filepath.Join(tmpDir, "goswitchd"), standalone: standalone}

	logFile, err := createLog(cfg.logPath)
	if err != nil {
		s.cleanupFiles()

		return nil, err
	}
	s.logFile = logFile
	s.logPath = logFile.Name()
	fmt.Printf("stand setup: daemon journal %s\n", s.logPath)

	fmt.Printf("stand setup: building goswitchd...\n")
	if err := buildDaemon(ctx, s.daemonBin); err != nil {
		s.discardLog(logFile)

		return nil, err
	}

	snap, err := snapshotDesktop(ctx)
	if err != nil {
		s.discardLog(logFile)

		return nil, err
	}
	s.snap = snap
	fmt.Printf("stand setup: desktop state snapshotted\n")

	if !standalone {
		fmt.Printf("stand setup: spawning the -debug daemon...\n")
		if err := s.startSetupDaemon(ctx); err != nil {
			s.discardLog(logFile)

			return nil, err
		}
	}

	return s, nil
}

// createLog opens the daemon log — a pinned path or a temp file.
func createLog(pinned string) (*os.File, error) {
	if pinned != "" {
		f, err := os.Create(pinned)
		if err != nil {
			return nil, fmt.Errorf("create daemon log %s: %w", pinned, err)
		}

		return f, nil
	}
	f, err := os.CreateTemp("", "goswitch-e2e-*.log")
	if err != nil {
		return nil, fmt.Errorf("create daemon log: %w", err)
	}

	return f, nil
}

// discardLog closes and removes an unused temp log on the setup-failure
// path (a pinned -log path is kept for inspection).
func (s *stand) discardLog(f *os.File) {
	_ = f.Close()
	if s.cfg.logPath == "" {
		_ = os.Remove(f.Name())
	}
	s.cleanupFiles()
}

// buildDaemon compiles the daemon fresh so the stand always tests the
// current tree.
func buildDaemon(ctx context.Context, bin string) error {
	cmd := exec.CommandContext(ctx, "go", "build", "-o", bin, "./cmd/goswitchd")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build goswitchd: %w: %s", err, strings.TrimSpace(out.String()))
	}

	return nil
}

// startDaemon spawns the daemon subprocess; readiness (component
// registered) is a preflight check, not an assumption here.
//
// it only AFTER the desktop state is restored — a CommandContext kill would
// fire on Ctrl-C before that restore runs.
func (s *stand) startDaemon() error {
	return s.startDaemonArgs()
}

// startDaemonArgs spawns the daemon with -debug plus extra arguments — the
// config-dependent cases append their -config after writing it.
//
//nolint:noctx // the daemon must outlive the run context: teardown SIGTERMs
func (s *stand) startDaemonArgs(args ...string) error {
	argv := append([]string{"-debug"}, args...)
	cmd := exec.Command(s.daemonBin, argv...)
	cmd.Stdout = s.logFile
	cmd.Stderr = s.logFile
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start goswitchd: %w", err)
	}
	s.daemon = cmd

	return nil
}

// startDaemonPlain spawns the daemon WITHOUT -debug — the production form
// the perf run measures (T-04-04-03: the -debug trace would be neither
// honest nor private; the acceptance numbers come from the prod shape).
func (s *stand) startDaemonPlain() error {
	return s.startDaemonPlainArgs()
}

// startDaemonPlainArgs spawns the daemon with extra arguments and NO
// -debug — the perf-autocorrect form (plan 06-08): a config attached, the
// prod shape kept (T-04-04-03).
//
//nolint:noctx // the daemon must outlive the run context: teardown SIGTERMs
func (s *stand) startDaemonPlainArgs(args ...string) error {
	cmd := exec.Command(s.daemonBin, args...)
	cmd.Stdout = s.logFile
	cmd.Stderr = s.logFile
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start goswitchd (prod form): %w", err)
	}
	s.daemon = cmd

	return nil
}

// restartDaemonWithArgs stops the running daemon and spawns a fresh one
// with extra arguments: the config cases build their command line after
// setup has already started the default daemon.
func (s *stand) restartDaemonWithArgs(args ...string) error {
	s.stopDaemon()

	return s.startDaemonArgs(args...)
}

// stopDaemon terminates the daemon with a bounded graceful wait.
func (s *stand) stopDaemon() {
	if s.daemon == nil || s.daemon.Process == nil {
		return
	}
	_ = s.daemon.Process.Signal(syscall.SIGTERM)
	s.waitDaemonExit()
}

// waitDaemonExit reaps the daemon subprocess, killing it after the grace
// period (used both after SIGTERM and after an external kill -9).
func (s *stand) waitDaemonExit() {
	done := make(chan struct{})
	go func() {
		_ = s.daemon.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(killGrace):
		_ = s.daemon.Process.Kill()
		<-done
	}
	s.daemon = nil
}

// startSetupDaemon is the setup daemon's full lifecycle (the pre-preflight
// spawn every non-standalone case depends on): stop any stale stand daemon
// from a previous run, stop the user unit when it is active (its daemon
// holds the component name and never allows replacement — the stand daemon
// could never register over it; the teardown restores it), await the name
// release with the generous budget, and only then spawn — with the
// registered-mark retry ladder, because registration is retry-until-ready,
// never a single attempt (the 2026-10-13 preflight finding: an unguarded
// spawn died with "another goswitchd instance already registered" and the
// preflight burned its whole budget on a dead daemon).
func (s *stand) startSetupDaemon(ctx context.Context) error {
	// The unit snapshot rides BEFORE the stale-reap: on the VM stand the
	// unit daemon itself runs -debug (the 10-01 oracle drop-in), so the
	// pgrep below matches it — a snapshot taken after the reap would read
	// "inactive" and the teardown would silently skip the restore (the
	// first term-word-vte run left the stand unit down, live finding
	// 2026-10-12). On a host whose unit is not -debug the order is
	// byte-equivalent.
	unit := unitState(ctx)
	s.unitWasActive = unit == unitActiveState
	if err := stopStaleStandDaemon(ctx); err != nil {
		return err
	}
	if s.unitWasActive {
		fmt.Printf("stand setup: the user unit goswitchd is active — stopping it (restored at teardown)\n")
		if _, err := runCmd(ctx, "systemctl", "--user", "stop", "goswitchd"); err != nil {
			return fmt.Errorf("stop unit daemon: %w", err)
		}
	}
	fmt.Printf("stand setup: awaiting the component name release (budget %v) before the spawn\n",
		daemonReleaseWait)
	if err := awaitNameFree(ctx, "before the setup daemon spawn"); err != nil {
		return err
	}

	return s.startDaemonRegistered(ctx)
}

// teardown is the ALWAYS-run cleanup (T-03-01/T-03-03): gsettings are
// restored (only the keys the stand actually changed), the daemon is
// stopped, and the global engine is re-asserted AFTER the daemon dies —
// ibus-daemon unsets the global engine when an engine component's
// connection closes (live-verified in phase 01-03), so restoring before
// the stop would be undone by it. run()'s defer guarantees all of this on
// FAIL and on Ctrl-C alike.
func (s *stand) teardown() {
	ctx, cancel := context.WithTimeout(context.Background(), teardownTimeout)
	defer cancel()

	s.restoreGsettings(ctx)
	hadDaemon := s.daemon != nil
	s.stopDaemon()
	if hadDaemon {
		// The shutdown half of the daemon lifecycle: reap any lingering
		// stand daemon (always -debug; the user's unit daemon never is) so
		// the next case never inherits a holder — the release lag beyond
		// the process death is covered by the next start's patient
		// awaitNameFree.
		reapStaleStandDaemons(ctx)
	}
	if s.unitWasActive {
		restoreUnit(ctx, unitActiveState)
	}
	s.restoreEngine(ctx)
	s.reapZenity()
	s.closeChromium()
	s.closeGTE()
	s.reapPasswordFixture()
	s.closeWezterm()
	_ = s.logFile.Close()
	if s.cfg.logPath == "" {
		_ = os.Remove(s.logPath)
	}
	s.cleanupFiles()
}

// cleanupFiles removes the stand's temp directory (daemon binary).
func (s *stand) cleanupFiles() {
	_ = os.RemoveAll(s.tmpDir)
}

// holdAction is the await's routing for one poll where the name did not
// probe free: which remediation the state calls for.
type holdAction int

const (
	holdProbeAgain holdAction = iota // the name is absent from the bus — free, re-probe
	holdWait                         // a live owner is releasing (or the restart already fired) — wait
	holdRestart                      // the poisoned entry: listed, ownerless, unclaimable — restart ibus once
)

// nameHoldDecision routes one held poll from the bus facts. A listed name
// with NO owner is the poisoned-entry signature (the 2026-10-13 finding:
// every flag set returned queued/exists while GetNameOwner reported no
// owner — no spawn could ever register on that ibus generation), and one
// bounded ibus restart is the cure; a live owner is the ordinary slow
// release. The caller's budget stays in charge either way — this is a
// pure routing decision, never a deadline change.
func nameHoldDecision(listed bool, ownerPID string, restarted bool) (holdAction, bool) {
	switch {
	case !listed:
		return holdProbeAgain, false
	case ownerPID != "":
		return holdWait, false
	case !restarted:
		return holdRestart, true
	default:
		return holdWait, false
	}
}

// daemonAlive reports whether the stand daemon process still exists
// (signal 0 probes existence without disturbing it). No daemon = not alive.
func (s *stand) daemonAlive() bool {
	return s.daemon != nil && s.daemon.Process != nil &&
		s.daemon.Process.Signal(syscall.Signal(0)) == nil
}

// lastErrorLineBound caps one journal error record's length in the
// fail-fast diagnostic (a full record with a long field would flood the
// console).
const lastErrorLineBound = 300

// lastErrorLine returns the journal's last ERROR-level record (bounded),
// or "" when the journal holds none — the diagnostic the fail-fast wait
// carries so a daemon death names its own cause.
func lastErrorLine(journal string) string {
	var found string
	for _, ln := range strings.Split(journal, "\n") {
		if strings.Contains(ln, `"level":"ERROR"`) {
			found = ln
		}
	}
	if len(found) > lastErrorLineBound {
		found = found[:lastErrorLineBound]
	}

	return found
}

// waitForLogWhileAlive blocks until the daemon journal contains mark — and
// fails FAST when the daemon process exits first, carrying the journal's
// last error record (the name-conflict death) instead of a generic timeout
// after the full budget (the 2026-10-13 preflight finding).
func (s *stand) waitForLogWhileAlive(ctx context.Context, mark string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if s.countSub(mark) >= 1 {
			return nil
		}
		if !s.daemonAlive() {
			journal, rerr := os.ReadFile(s.logPath)
			lastErr := ""
			if rerr == nil {
				lastErr = lastErrorLine(string(journal))
			}
			if lastErr == "" {
				return fmt.Errorf("the daemon process exited before %q landed (journal %s)", mark, s.logPath)
			}

			return fmt.Errorf("the daemon process exited before %q landed — last journal error: %s",
				mark, lastErr)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout after %v: daemon log has %d occurrences of %q, want >= 1",
				timeout, s.countSub(mark), mark)
		}
		if err := sleepCtx(ctx, logPollInterval); err != nil {
			return fmt.Errorf("wait for %q aborted: %w", mark, err)
		}
	}
}

// ibusRestartSettle is the pause after an ibus restart before the next
// probe: the fresh generation needs a beat to write its address files and
// export its bus.
const ibusRestartSettle = 2 * time.Second

// staleStopGrace is how long a SIGTERM'd stale daemon gets to exit on its
// own before the SIGKILL escalation (the socket close is what releases the
// component name — a stalled shutdown holds it indefinitely).
const staleStopGrace = 10 * time.Second

// stopStaleStandDaemon stops a LIVE previous stand daemon, if any — a
// previous run left it holding the component name (stand daemons always
// run -debug; the user's unit daemon never does). The stop half of the
// setup lifecycle: the await lives in awaitNameFree.
func stopStaleStandDaemon(ctx context.Context) error {
	return stopStaleStandDaemonWith(ctx, staleStopGrace)
}

// stopStaleStandDaemonWith is the escalated core: every matching holder is
// SIGTERM'd, and after the grace the survivors are SIGKILLed — a daemon
// stalled in its own shutdown keeps the name held until its socket closes,
// so waiting forever behind it is the bug this escalation exists to kill.
func stopStaleStandDaemonWith(ctx context.Context, grace time.Duration) error {
	out, err := runCmd(ctx, "pgrep", "-f", "goswitchd -debug")
	if err != nil {
		fmt.Printf("no live stand daemon found — nothing to stop\n")

		return nil //nolint:nilerr // pgrep exit 1 = no live stand daemon — the good outcome
	}
	for _, pid := range strings.Fields(out) {
		fmt.Printf("stale stand daemon pid %s — SIGTERM\n", pid)
		_, _ = runCmd(ctx, "kill", pid)
	}
	deadline := time.Now().Add(grace)
	for {
		out, err := runCmd(ctx, "pgrep", "-f", "goswitchd -debug")
		if err != nil {
			return nil //nolint:nilerr // no match = every holder exited
		}
		if time.Now().After(deadline) {
			for _, pid := range strings.Fields(out) {
				fmt.Printf("stale stand daemon pid %s — SIGKILL (stuck past the grace)\n", pid)
				_, _ = runCmd(ctx, "kill", "-9", pid)
			}

			return nil
		}
		if err := sleepCtx(ctx, logPollInterval); err != nil {
			return err
		}
	}
}

// awaitNameFree blocks until the component's well-known name is claimable —
// the non-destructive DoNotQueue probe of waitNameFree under the generous
// daemonReleaseWait budget (the spike cases' patient replacement for the
// fixed 15 s budget the slow-releasing previous daemon outlived). The
// start half of the lifecycle: after this returns nil, the spawn cannot
// lose a name race.
func awaitNameFree(ctx context.Context, why string) error {
	deadline := time.Now().Add(daemonReleaseWait)
	restarted := false
	for {
		// Re-discover on every iteration: an ibus restart (below) mints a
		// fresh generation and address — a cached one would probe a dead
		// bus forever.
		free, perr := probeNameFreeAt()
		if perr != nil {
			return fmt.Errorf("name probe (%s): %w", why, perr)
		}
		if free {
			return nil
		}
		listed, ownerPID, qerr := ibusNameState(ctx)
		if qerr != nil {
			return fmt.Errorf("name state probe (%s): %w", why, qerr)
		}
		action, doRestart := nameHoldDecision(listed, ownerPID, restarted)
		if doRestart {
			fmt.Printf("the component name is POISONED on this ibus generation (listed, ownerless, "+
				"unclaimable under every flag set) — restarting ibus once (%s)\n", why)
			if _, rerr := runCmd(ctx, "ibus", "restart"); rerr != nil {
				return fmt.Errorf("ibus restart (%s): %w", why, rerr)
			}
			restarted = true
			if err := sleepCtx(ctx, ibusRestartSettle); err != nil {
				return err
			}

			continue
		}
		if action == holdProbeAgain {
			if err := sleepCtx(ctx, logPollInterval); err != nil {
				return err
			}

			continue
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s still held after %v (%s) — the previous holder has not released the component name",
				goswitchComponentName, daemonReleaseWait, why)
		}
		if err := sleepCtx(ctx, logPollInterval); err != nil {
			return err
		}
	}
}

// probeNameFreeAt probes the component name claimable on the CURRENT ibus
// generation (the address is re-discovered per call — never cached).
func probeNameFreeAt() (bool, error) {
	addr, err := engine.Discover()
	if err != nil {
		return false, fmt.Errorf("discover ibus bus: %w", err)
	}

	return probeNameFree(addr)
}

// ibusNameState reports whether the component's well-known name is listed
// on the live ibus bus and, when owned, the owner's pid ("" when ownerless
// — the poisoned-entry signature). ListNames is the presence oracle: the
// poisoned entry appears there with NO owner while every claim attempt is
// refused.
func ibusNameState(ctx context.Context) (listed bool, ownerPID string, err error) {
	addr, err := engine.Discover()
	if err != nil {
		return false, "", fmt.Errorf("discover ibus bus: %w", err)
	}
	conn, err := dbus.Dial(addr)
	if err != nil {
		return false, "", fmt.Errorf("dial ibus bus: %w", err)
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "e2e: close ibus name-state connection: %v\n", cerr)
		}
	}()
	if err := conn.Auth([]dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}); err != nil {
		return false, "", fmt.Errorf("dbus auth: %w", err)
	}
	if err := conn.Hello(); err != nil {
		return false, "", fmt.Errorf("dbus hello: %w", err)
	}
	obj := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus")
	namesCall := obj.CallWithContext(ctx, "org.freedesktop.DBus.ListNames", 0)
	if namesCall.Err != nil {
		return false, "", fmt.Errorf("ListNames: %w", namesCall.Err)
	}
	listed = bodyMentions(namesCall.Body, goswitchComponentName)
	ownerCall := obj.CallWithContext(ctx, "org.freedesktop.DBus.GetNameOwner", 0, goswitchComponentName)
	if ownerCall.Err != nil {
		var dbusErr dbus.Error
		if errors.As(ownerCall.Err, &dbusErr) && dbusErr.Name == "org.freedesktop.DBus.Error.NameHasNoOwner" {
			return listed, "", nil
		}

		return listed, "", fmt.Errorf("GetNameOwner: %w", ownerCall.Err)
	}
	ownerPID = ownerPIDFromCall(ctx, obj, ownerCall)

	return listed, ownerPID, nil
}

// ownerPIDFromCall resolves the unix pid of the unique name carried in a
// successful GetNameOwner reply ("" when the reply shape surprises).
func ownerPIDFromCall(ctx context.Context, obj dbus.BusObject, ownerCall *dbus.Call) string {
	if len(ownerCall.Body) != 1 {
		return ""
	}
	unique, ok := ownerCall.Body[0].(string)
	if !ok {
		return ""
	}
	pidCall := obj.CallWithContext(ctx,
		"org.freedesktop.DBus.GetConnectionUnixProcessID", 0, unique)
	if pidCall.Err != nil || len(pidCall.Body) != 1 {
		return ""
	}
	pid, ok := pidCall.Body[0].(uint32)
	if !ok {
		return ""
	}

	return strconv.FormatUint(uint64(pid), 10)
}

// reapStaleStandDaemons SIGTERMs any lingering stand daemon — the
// shutdown-path half of the daemon lifecycle (the pid-точечный discipline:
// stand daemons always run -debug and live under a temp dir, so the
// pattern never matches the user's unit daemon).
func reapStaleStandDaemons(ctx context.Context) {
	out, err := runCmd(ctx, "pgrep", "-f", "goswitchd -debug")
	if err != nil {
		return
	}
	for _, pid := range strings.Fields(out) {
		_, _ = runCmd(ctx, "kill", pid)
	}
}

// snapshotDesktop captures the live-desktop state: both gsettings
// input-source keys and the IBus global engine name (empty when none is
// set — `ibus engine` exits non-zero then, which is not a stand failure).
func snapshotDesktop(ctx context.Context) (desktopSnapshot, error) {
	sources, err := runCmd(ctx, "gsettings", "get", gsettingsSchema, keySources)
	if err != nil {
		return desktopSnapshot{}, fmt.Errorf("snapshot %s: %w", keySources, err)
	}
	current, err := runCmd(ctx, "gsettings", "get", gsettingsSchema, keyCurrent)
	if err != nil {
		return desktopSnapshot{}, fmt.Errorf("snapshot %s: %w", keyCurrent, err)
	}
	engineName, _ := runCmd(ctx, "ibus", "engine")

	return desktopSnapshot{sources: sources, current: current, engineName: engineName}, nil
}

// restoreGsettings puts back the two input-source keys — but only the ones
// that actually differ from the snapshot. A value-identical `gsettings set`
// still triggers GNOME's input-source re-evaluation, which live-unsets the
// global engine (phase 01-03 finding); the stand's activation path
// (SetGlobalEngine) never mutates these keys, so in practice nothing is
// written and the owner's desktop is left untouched.
func (s *stand) restoreGsettings(ctx context.Context) {
	for _, kv := range []struct{ key, value string }{
		{keySources, s.snap.sources},
		{keyCurrent, s.snap.current},
	} {
		now, err := runCmd(ctx, "gsettings", "get", gsettingsSchema, kv.key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "e2e: read gsettings %s for restore: %v\n", kv.key, err)

			continue
		}
		if now == kv.value {
			continue
		}
		if _, err := runCmd(ctx, "gsettings", "set", gsettingsSchema, kv.key, kv.value); err != nil {
			fmt.Fprintf(os.Stderr, "e2e: restore gsettings %s: %v\n", kv.key, err)
		}
	}
}

// verifyRestored machine-checks the teardown contract (plan 01-04, T-04-02):
// after the restore, BOTH gsettings input-source keys must read back equal
// to the snapshot. A mismatch is a named FAIL — the owner's desktop state
// was not put back — so it must fail the exit code, not a human's glance.
func (s *stand) verifyRestored() error {
	ctx, cancel := context.WithTimeout(context.Background(), restoreVerifyTimeout)
	defer cancel()

	for _, kv := range []struct{ key, snapshot string }{
		{keySources, s.snap.sources},
		{keyCurrent, s.snap.current},
	} {
		now, err := runCmd(ctx, "gsettings", "get", gsettingsSchema, kv.key)
		if err != nil {
			return fmt.Errorf("verify restore: read %s: %w", kv.key, err)
		}
		if now != kv.snapshot {
			return fmt.Errorf("restore verification FAILED for %s: snapshot %q, now %q (desktop state not restored)",
				kv.key, kv.snapshot, now)
		}
	}

	return nil
}

// restoreEngine re-asserts the snapshotted global engine. Called after the
// daemon's death (see teardown) and never to a goswitch engine — a snapshot
// captured while a goswitch engine was global would otherwise point the
// desktop at a dead engine after teardown.
func (s *stand) restoreEngine(ctx context.Context) {
	name := s.snap.engineName
	if name == "" || strings.Contains(name, "goswitch") {
		if fallback := s.fallbackEngine(); fallback != "" {
			name = fallback
		}
	}
	if name == "" {
		return
	}
	if _, err := runCmd(ctx, "ibus", "engine", name); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: restore ibus engine %q: %v\n", name, err)
	}
}

// fallbackEngine derives a plain XKB engine name from the first ('xkb',
// layout) tuple of the snapshotted sources — the best-effort desktop-safe
// engine when the snapshot itself is unusable.
func (s *stand) fallbackEngine() string {
	layout := s.snap.sources
	if i := strings.Index(layout, "'xkb'"); i >= 0 {
		rest := layout[i+len("'xkb'"):]
		start := strings.Index(rest, "'")
		end := strings.Index(rest[start+1:], "'")
		if start >= 0 && end >= 0 {
			return "xkb:" + rest[start+1:start+1+end] + "::eng"
		}
	}

	return ""
}

// runCmd runs a command capturing stdout; a non-zero exit reports the
// command line and stderr for the named diagnostics. Every call carries its
// own hard deadline (cmdTimeout): a congested D-Bus, a wedged ydotool or a
// stalled helper must fail the named check instead of freezing the stand —
// a frozen stand strands the owner's desktop state (live finding, run 2 of
// the D-01 matrix: the runner sat 4+ minutes inside an unbounded call).
func runCmd(ctx context.Context, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}

	return strings.TrimSpace(out.String()), nil
}

// sleepCtx sleeps in abortable chunks so Ctrl-C interrupts waits.
func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("sleep aborted: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}

// countSub counts substring occurrences in the daemon log.
func (s *stand) countSub(sub string) int {
	data, err := os.ReadFile(s.logPath)
	if err != nil {
		return 0
	}

	return strings.Count(string(data), sub)
}

// logText returns the whole daemon log as one string — the input of the
// position-based flip-marks oracles (the countSub sibling that needs
// offsets, not counts).
func (s *stand) logText() string {
	data, err := os.ReadFile(s.logPath)
	if err != nil {
		return ""
	}

	return string(data)
}

// waitForLog blocks until the daemon log contains sub — a
// wait-for-condition poll, never a fixed sleep (Pitfall 6).
func (s *stand) waitForLog(ctx context.Context, sub string, timeout time.Duration) error {
	return s.waitForNew(ctx, sub, 1, timeout)
}

// waitForNew blocks until the daemon log contains at least want occurrences
// of sub; resilience cases use the count to require NEW events after a
// restart rather than any old match.
func (s *stand) waitForNew(ctx context.Context, sub string, want int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		got := s.countSub(sub)
		if got >= want {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout after %v: daemon log has %d occurrences of %q, want >= %d",
				timeout, got, sub, want)
		}
		if err := sleepCtx(ctx, logPollInterval); err != nil {
			return fmt.Errorf("wait for %q aborted: %w", sub, err)
		}
	}
}

// logLines returns the daemon log split into JSON records.
func (s *stand) logLines() []string {
	data, err := os.ReadFile(s.logPath)
	if err != nil {
		return nil
	}

	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

// firstRecordTime returns the JSON timestamp of the first log record
// containing sub.
func (s *stand) firstRecordTime(sub string) (time.Time, error) {
	for _, line := range s.logLines() {
		if strings.Contains(line, sub) {
			return parseLogTime(line)
		}
	}

	return time.Time{}, fmt.Errorf("no daemon log record containing %q", sub)
}

// parseLogTime extracts the time field of one JSON log record.
func parseLogTime(line string) (time.Time, error) {
	var rec struct {
		Time string `json:"time"`
	}
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		return time.Time{}, fmt.Errorf("parse log record %q: %w", line, err)
	}
	ts, err := time.Parse(time.RFC3339Nano, rec.Time)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time in %q: %w", line, err)
	}

	return ts, nil
}

// printLogExcerpt dumps the tail of the daemon log to aid FAIL triage.
func (s *stand) printLogExcerpt() {
	lines := s.logLines()
	start := max(0, len(lines)-excerptLines)
	fmt.Fprintf(os.Stderr, "--- daemon log excerpt (last %d records) ---\n%s\n",
		len(lines)-start, strings.Join(lines[start:], "\n"))
}

// injectText types text through ydotool with the configured pacing. The
// ydotoold-unavailable stderr notice on 0.1.8 is benign (it writes to
// /dev/uinput directly).
func (s *stand) injectText(ctx context.Context, text string) error {
	if _, err := runCmd(ctx, "ydotool", "type", "--key-delay", strconv.Itoa(s.cfg.pacing), text); err != nil {
		return fmt.Errorf("inject %q: %w", text, err)
	}

	return nil
}

// injectKeys taps key names with pacing×3 between them so a multi-tap
// lands inside the 300 ms disambiguation window even on a loaded machine.
func (s *stand) injectKeys(ctx context.Context, keys ...string) error {
	prefix := s.keyCmdPrefix()
	args := make([]string, 0, len(keys)+len(prefix))
	args = append(args, prefix...)
	args = append(args, keys...)
	if _, err := runCmd(ctx, "ydotool", args...); err != nil {
		return fmt.Errorf("inject keys %v: %w", keys, err)
	}

	return nil
}

// keyCmdPrefix is the fixed ydotool key invocation with the window-paced
// delay (tapWindowFactor).
func (s *stand) keyCmdPrefix() []string {
	return []string{"key", "--key-delay", strconv.Itoa(s.cfg.pacing * tapWindowFactor)}
}

// pressKey sends one key event (super, Escape, ...).
func (s *stand) pressKey(ctx context.Context, key string) error {
	if _, err := runCmd(ctx, "ydotool", "key", key); err != nil {
		return fmt.Errorf("press %q: %w", key, err)
	}

	return nil
}
