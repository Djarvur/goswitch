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
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
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
	caseName   string
	pacing     int
	logPath    string
	helper     string
	matrixPath string
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
	cfg       config
	daemonBin string
	tmpDir    string
	logPath   string
	logFile   *os.File
	daemon    *exec.Cmd
	zenity    *exec.Cmd
	zenityOut *bytes.Buffer
	chromium  *exec.Cmd
	gte       *exec.Cmd
	snap      desktopSnapshot
}

func main() {
	os.Exit(run())
}

// run wires the whole stand: registry, setup, preflight, the case and the
// ALWAYS-run teardown. Returning an int keeps the exit-code contract in one
// place; defers are honored because run returns normally. The deferred
// finisher also machine-verifies the gsettings restore (plan 01-04, T-04-02)
// and may downgrade a PASS to FAIL through the named return.
func run() (exit int) {
	var cfg config
	flag.StringVar(&cfg.caseName, "case", "",
		"case to run: m1-gate | ibus-restart | kill9-survive | d01-probe | chromium-smoke | gte-smoke"+
			" | word-en-ru | word-after-space | word-ru-en | word-mixed | phrase-en-ru | phrase-mixed"+
			" | ladder-chromium | reset-escape | select-smoke | select-correct")
	flag.IntVar(&cfg.pacing, "pacing", defaultPacingMs,
		"milliseconds between injected keystrokes (raise on a loaded machine)")
	flag.StringVar(&cfg.logPath, "log", "", "daemon log path (default: a temp file removed in teardown)")
	flag.StringVar(&cfg.helper, "helper", "test/e2e/focus_helper.py", "path to the AT-SPI helper script")
	flag.StringVar(&cfg.matrixPath, "matrix", "",
		"YAML case matrix to run (multi-doc cases; every case gets a fresh daemon)")
	flag.Parse()

	// The matrix branch owns its whole stand lifecycle: case isolation needs
	// a setupStand → case → teardown cycle PER CASE (02-06), so the single
	// shared stand of the -case path must not wrap it.
	if cfg.matrixPath != "" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		return runMatrixFile(ctx, cfg, cfg.matrixPath)
	}

	caseFn, err := pickCase(cfg.caseName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", cfg.caseName, err)

		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s, err := setupStand(ctx, cfg)
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

	if err := preflight(ctx, s); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", cfg.caseName, err)

		return 1
	}

	if err := runCaseWatchdog(ctx, cfg.caseName, caseFn, s); err != nil {
		s.printLogExcerpt()
		fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", cfg.caseName, err)

		return 1
	}

	fmt.Println("PASS " + cfg.caseName)

	return 0
}

// runCaseWatchdog runs the case body under a hard deadline (caseTimeout) and
// cancels its context on expiry: a live-session stand must never strand the
// owner's desktop. A stuck case fails fast, names the case, and control
// falls through to the teardown contract (restore + machine verification).
func runCaseWatchdog(
	ctx context.Context,
	name string,
	fn func(context.Context, *stand) error,
	s *stand,
) error {
	caseCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- fn(caseCtx, s) }()

	select {
	case err := <-done:
		return err
	case <-time.After(caseTimeout):
		cancel()

		return fmt.Errorf("watchdog: case %q did not complete within %v", name, caseTimeout)
	}
}

// pickCase resolves the case name through the registry; the registry is
// built per call (no mutable globals).
func pickCase(name string) (func(context.Context, *stand) error, error) {
	registry := map[string]func(context.Context, *stand) error{
		"m1-gate":          runM1Gate,
		"ibus-restart":     runIbusRestart,
		"kill9-survive":    runKill9Survive,
		"d01-probe":        runD01Probe,
		"chromium-smoke":   runChromiumSmoke,
		"gte-smoke":        runGTESmoke,
		"word-en-ru":       runWordENRU,
		"word-after-space": runWordAfterSpace,
		"word-ru-en":       runWordRUEN,
		"word-mixed":       runWordMixed,
		"phrase-en-ru":     runPhraseENRU,
		"phrase-mixed":     runPhraseMixed,
		"ladder-chromium":  runLadderChromium,
		"reset-escape":     runResetEscape,
		"select-smoke":     runSelectSmoke,
		"select-correct":   runSelectCorrect,
	}
	fn, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown or missing -case %q (registry: m1-gate, ibus-restart, kill9-survive,"+
			" d01-probe, chromium-smoke, gte-smoke, word-en-ru, word-after-space, word-ru-en, word-mixed,"+
			" phrase-en-ru, phrase-mixed, ladder-chromium, reset-escape, select-smoke, select-correct)", name)
	}

	return fn, nil
}

// setupStand builds the daemon, snapshots the desktop and starts the
// daemon subprocess with -debug and both std streams into the log file
// (T-03-03: the log lives in $TMPDIR and is removed unless -log pinned it).
func setupStand(ctx context.Context, cfg config) (*stand, error) {
	tmpDir, err := os.MkdirTemp("", "goswitch-e2e-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	s := &stand{cfg: cfg, tmpDir: tmpDir, daemonBin: filepath.Join(tmpDir, "goswitchd")}

	logFile, err := createLog(cfg.logPath)
	if err != nil {
		s.cleanupFiles()

		return nil, err
	}
	s.logFile = logFile
	s.logPath = logFile.Name()

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

	if err := s.startDaemon(); err != nil {
		s.discardLog(logFile)

		return nil, err
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
//
//nolint:noctx // the daemon must outlive the run context: teardown SIGTERMs
func (s *stand) startDaemon() error {
	cmd := exec.Command(s.daemonBin, "-debug")
	cmd.Stdout = s.logFile
	cmd.Stderr = s.logFile
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start goswitchd: %w", err)
	}
	s.daemon = cmd

	return nil
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
	s.stopDaemon()
	s.restoreEngine(ctx)
	s.reapZenity()
	s.closeChromium()
	s.closeGTE()
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
