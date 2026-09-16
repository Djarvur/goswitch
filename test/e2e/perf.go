package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// INST-03 acceptance budgets (plan 04-04): the latency and memory budgets
// are a GATE on the perf run (exit != 0 over budget), never a reference
// metric. The named constants are the methodology's numbers (D-44).
const (
	perfLatencyBudget  = 50 * time.Millisecond // reaction to the hot chord < 50 ms
	perfMemoryBudgetKB = 50 * 1024             // peak RSS < 50 MB
)

// The resident witness protocol pieces: the helper interpreter (PATH
// python3 is linuxbrew without gi — focus_helper.py docstring), the line
// format "<ts> <app> <pid> <text>" (the text is the fourth field and may
// contain spaces, hence SplitN), the stdout scanner buffer bounds and the
// line queue depth.
const (
	python3HelperBin  = "/usr/bin/python3"
	witnessLineFields = 4
	witnessScanInit   = 64 * 1024
	witnessScanMax    = 1024 * 1024
	witnessLineQueue  = 64
)

// Perf-run knobs (plan 04-04): N, the observation budget and the witness
// mode are the methodology's numbers (D-43/D-44) — named once here, cited
// by the report. The witness mode is PINNED by the Task-1 live smoke: the
// event listener delivered (zenity, line well under 2 s, pid-tagged), so
// the run is event-driven with NO polling quantum; witness-poll (fixed
// 5 ms quantum, WITNESS_POLL_QUANTUM_S in focus_helper.py) is the
// documented fallback if a future desktop proves the listener unstable.
const (
	perfSamples      = 40               // N: homogeneous hot-case repeats (D-43)
	perfSampleWait   = 5 * time.Second  // per-repeat witness observation budget
	perfRepeatBudget = 15 * time.Second // per-repeat wall budget; watchdog = N × this
	perfReportPath   = "perf-report.txt"
	perfWitnessMode  = "witness-events"
	perfSurfaceApp   = "zenity"
)

// The reported percentile ranks (D-43): the median and the two tail
// quantiles the acceptance table publishes.
const (
	perfP50 = 0.50
	perfP95 = 0.95
	perfP99 = 0.99
)

// percentile sorts a copy of the sample and indexes it — the research
// formula s[int(float64(len(s)-1)*p)] (04-RESEARCH Don't Hand-Roll: no
// stats dependency, no interpolation).
func percentile(samples []time.Duration, p float64) time.Duration {
	if len(samples) == 0 {
		return 0
	}
	s := append([]time.Duration(nil), samples...)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })

	return s[int(float64(len(s)-1)*p)]
}

// parseProcStatus extracts VmHWM/VmRSS (kB) from one /proc/<pid>/status
// document — the D-45 oracle mechanics. A missing VmHWM line is a named
// error: the peak oracle must never report a silent zero.
func parseProcStatus(data string) (hwm, rss int64, err error) {
	for _, line := range strings.Split(data, "\n") {
		if v, ok := strings.CutPrefix(line, "VmHWM:"); ok {
			hwm = parseProcStatusKB(v)
		}
		if v, ok := strings.CutPrefix(line, "VmRSS:"); ok {
			rss = parseProcStatusKB(v)
		}
	}
	if hwm == 0 {
		return 0, 0, errors.New("no VmHWM line")
	}

	return hwm, rss, nil
}

// parseProcStatusKB converts one " 12345 kB" status value to an int64.
func parseProcStatusKB(v string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(v), "kB")), 10, 64)

	return n
}

// readProcStatus reads the measured process's VmHWM/VmRSS once — pid
// identity is the stand's own subprocess (s.daemon.Process.Pid), never a
// pgrep/pkill pattern (research anti-pattern, live hit).
func readProcStatus(pid int) (hwm, rss int64, err error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, 0, fmt.Errorf("read /proc/%d/status: %w", pid, err)
	}

	return parseProcStatus(string(data))
}

// budgetGate turns the INST-03 budgets into the acceptance gate: any side
// at or over its boundary fails the run (exit != 0).
func budgetGate(p95 time.Duration, vmHWMKB int64) error {
	if p95 >= perfLatencyBudget {
		return fmt.Errorf("budget violated: p95 %v >= %v", p95, perfLatencyBudget)
	}
	if vmHWMKB >= perfMemoryBudgetKB {
		return fmt.Errorf("budget violated: vm_hwm %d kB >= %d kB", vmHWMKB, perfMemoryBudgetKB)
	}

	return nil
}

// witnessEvent is one parsed witness line: "<RFC3339> <app> <pid> <text>".
// The leading timestamp is stamped by the helper at OBSERVATION arrival —
// it is the t1 of the D-44 window, and the text readback latency behind it
// stays out of the measured interval.
type witnessEvent struct {
	at   time.Time
	app  string
	pid  int
	text string
}

// parseWitnessLine splits one witness protocol line. The text (last field)
// may contain spaces; the app name and pid may not.
func parseWitnessLine(line string) (witnessEvent, error) {
	parts := strings.SplitN(line, " ", witnessLineFields)
	if len(parts) != witnessLineFields {
		return witnessEvent{}, fmt.Errorf("malformed witness line %q", line)
	}
	at, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return witnessEvent{}, fmt.Errorf("witness timestamp %q: %w", parts[0], err)
	}
	pid := -1 // "pid if available" — the helper prints - when it cannot
	if parts[2] != "-" {
		if parsed, cerr := strconv.Atoi(parts[2]); cerr == nil {
			pid = parsed
		}
	}

	return witnessEvent{at: at, app: parts[1], pid: pid, text: parts[3]}, nil
}

// perfWitness is the resident AT-SPI witness client: ONE python spawn per
// call (pid identity, shutdown by pid — never pkill/pgrep -f), a line
// reader feeding a channel, and stderr captured under a mutex for the
// failure diagnostics (T-04-04-02: a wedged witness fails the wait loudly).
type perfWitness struct {
	cmd   *exec.Cmd
	lines chan string
	mu    sync.Mutex
	errs  []string
}

// startWitness spawns the helper's resident mode — witness-events (event
// listener, one spawn per run) or witness-poll (the A2 fallback, one spawn
// per measured surface, its pid passed as the watch target). The witness
// rides the given context: a watchdog cancellation kills it, so a wedged
// helper can never outlive the case.
func startWitness(ctx context.Context, helper, mode string, args ...string) (*perfWitness, error) {
	argv := append([]string{helper, mode}, args...)
	cmd := exec.CommandContext(ctx, python3HelperBin, argv...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("witness stdout pipe: %w", err)
	}
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("witness stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start witness %s: %w", mode, err)
	}
	w := &perfWitness{cmd: cmd, lines: make(chan string, witnessLineQueue)}
	go func() {
		sc := bufio.NewScanner(out)
		sc.Buffer(make([]byte, 0, witnessScanInit), witnessScanMax)
		for sc.Scan() {
			w.lines <- sc.Text()
		}
	}()
	go func() {
		sc := bufio.NewScanner(errPipe)
		for sc.Scan() {
			w.mu.Lock()
			w.errs = append(w.errs, sc.Text())
			w.mu.Unlock()
		}
	}()

	return w, nil
}

// await reads witness lines until pred accepts one; the matched line's
// helper-side timestamp is the return value (t1). A timeout, a run-context
// cancellation or a dead witness all fail the wait loudly.
func (w *perfWitness) await(
	ctx context.Context, timeout time.Duration, pred func(witnessEvent) bool,
) (time.Time, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return time.Time{}, fmt.Errorf("witness await aborted: %w", ctx.Err())
		case <-timer.C:
			return time.Time{}, fmt.Errorf("witness event did not arrive within %v (witness stderr: %s)",
				timeout, w.stderr())
		case line := <-w.lines:
			ev, err := parseWitnessLine(line)
			if err != nil {
				continue // protocol noise is never a measurement
			}
			if pred(ev) {
				return ev.at, nil
			}
		}
	}
}

// stderr returns the witness's collected stderr (diagnostics only).
func (w *perfWitness) stderr() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	return strings.Join(w.errs, "; ")
}

// shutdown kills the witness process by pid and reaps it.
func (w *perfWitness) shutdown() {
	if w.cmd.Process != nil {
		_ = w.cmd.Process.Kill()
	}
	_ = w.cmd.Wait()
}

// runPerf is the dedicated INST-03 acceptance run (D-43/D-44/D-45): one
// homogeneous hot case — the Shift+RightCtrl combo correcting ghbdtn→привет
// AND flipping the mode in a single chord (the live diagnostic pinned the
// press-side immediate fire: press→correction-done ≈ 0.3–1.8 ms, no
// discrimination window) — repeated N times under a resident event
// witness, reported as p50/p95/p99 plus the memory checkpoints, gated by
// the budget (exit != 0 over budget). The daemon under measurement is the
// PRODUCTION form: no -config (built-in defaults), no -debug (T-04-04-03 —
// the numbers must be honest).
func runPerf(ctx context.Context, s *stand) error {
	regBase := s.countSub(componentRegisteredMark)
	s.stopDaemon()
	if err := s.startDaemonPlain(); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, componentRegisteredMark, regBase+1, registrationWait); err != nil {
		return fmt.Errorf("perf prod-form daemon registration: %w", err)
	}
	if err := s.activateGoswitchFresh(ctx); err != nil {
		return fmt.Errorf("perf activation: %w", err)
	}
	_, rssStart, err := readProcStatus(s.daemon.Process.Pid)
	if err != nil {
		return fmt.Errorf("perf VmRSS checkpoint 1: %w", err)
	}
	combo, err := s.probeComboInjection(ctx)
	if err != nil {
		return fmt.Errorf("perf combo canonicalization: %w", err)
	}
	w, err := startWitness(ctx, s.cfg.helper, perfWitnessMode)
	if err != nil {
		return err
	}
	defer w.shutdown()

	samples := make([]time.Duration, 0, perfSamples)
	for i := range perfSamples {
		sample, err := perfRepeat(ctx, s, combo, w)
		if err != nil {
			return fmt.Errorf("perf repeat %d/%d: %w", i+1, perfSamples, err)
		}
		samples = append(samples, sample)
	}

	hwm, rssEnd, err := readProcStatus(s.daemon.Process.Pid)
	if err != nil {
		return fmt.Errorf("perf VmHWM oracle: %w", err)
	}
	p50 := percentile(samples, perfP50)
	p95 := percentile(samples, perfP95)
	p99 := percentile(samples, perfP99)
	verdict := budgetGate(p95, hwm)
	if err := writePerfReport(combo, samples, p50, p95, p99, rssStart, rssEnd, hwm, verdict); err != nil {
		return err
	}

	return verdict
}

// activateGoswitchFresh is the count-based form of activateGoswitch,
// required after the in-case daemon restart: the log already carries the
// old daemon's focus_in records, so a plain waitForLog would pass on a
// stale match while the fresh engine is not global yet.
func (s *stand) activateGoswitchFresh(ctx context.Context) error {
	base := s.countSub("focus_in")
	if _, err := runCmd(ctx, "ibus", "engine", "goswitch-en"); err != nil {
		return fmt.Errorf("activate goswitch-en: %w", err)
	}

	return s.waitForNew(ctx, "focus_in", base+1, focusWait)
}

// perfRepeat drives ONE homogeneous hot-case sample (D-43). A fresh zenity
// entry per repeat: the closing Enter doubles as the ADR-004 buffer reset,
// so no repeat inherits the previous one's buffer state. The probe word is
// typed and settled (quiesce) BEFORE t0; t0 stamps immediately before the
// chord injection; t1 is the witness's text-changed observation carrying
// the converted word (its helper-side timestamp — the D-44 window
// ydotool→AT-SPI, honest upper bound, ydotool spawn included). The
// flip-back tap and the zenity stdout readback confirm the repeat AFTER
// the sample, outside the window.
func perfRepeat(ctx context.Context, s *stand, combo string, w *perfWitness) (time.Duration, error) {
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return 0, err
	}
	if kind != surfaceZenity {
		return 0, errors.New("perf needs the zenity entry surface (locked-session fallback engaged?)")
	}
	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return 0, fmt.Errorf("type the probe word: %w", err)
	}
	if err := s.waitZenityText(ctx, wordProbeEN); err != nil {
		return 0, fmt.Errorf("probe word settle (quiesce): %w", err)
	}

	t0 := time.Now() // the D-44 window opens immediately before the injection
	if err := s.pressKey(ctx, combo); err != nil {
		return 0, fmt.Errorf("combo injection: %w", err)
	}
	t1, err := w.await(ctx, perfSampleWait, func(ev witnessEvent) bool {
		return ev.app == perfSurfaceApp && ev.text == wordResultRU && !ev.at.Before(t0)
	})
	if err != nil {
		return 0, fmt.Errorf("witness observation: %w", err)
	}

	// The combo corrected AND flipped to RU; the next repeat needs EN
	// typing, so the tap-back gates on ITS OWN mode record before the
	// close — count-based, never a plain waitForLog: the probe's flip-back
	// record is already in the log and a stale match would close the
	// surface with the tap window still open.
	enBase := s.countSub(`"msg":"mode","to":"en"`)
	if err := s.injectKeys(ctx, "Shift_R"); err != nil {
		return 0, fmt.Errorf("flip-back tap: %w", err)
	}
	if err := s.waitForNew(ctx, `"msg":"mode","to":"en"`, enBase+1, decisionWait); err != nil {
		return 0, fmt.Errorf("flip-back record: %w", err)
	}
	out, err := s.closeZenity(ctx)
	if err != nil {
		return 0, fmt.Errorf("surface close: %w", err)
	}
	if out != wordResultRU {
		return 0, fmt.Errorf("repeat readback: entry printed %q, want %q", out, wordResultRU)
	}

	return t1.Sub(t0), nil
}

// writePerfReport renders the acceptance report — stdout AND
// perf-report.txt (the artifact the README's D-46 table transcribes): the
// methodology block naming the witness mechanism and its quantum, the
// percentile rows, the memory checkpoints and the budget verdict. Field
// CONTENT is never printed (T-04-04-01): numbers, units and verdicts only.
func writePerfReport(
	combo string, samples []time.Duration, p50, p95, p99 time.Duration,
	rssStart, rssEnd, hwm int64, verdict error,
) error {
	var b strings.Builder
	fmt.Fprintf(&b, "perf report — %s\n", time.Now().UTC().Format(time.RFC3339))
	b.WriteString("methodology:\n")
	b.WriteString("  window: ydotool injection -> AT-SPI text-changed observation" +
		" (honest upper bound; includes the ydotool spawn)\n")
	if perfWitnessMode == "witness-events" {
		b.WriteString("  witness: resident Atspi.EventListener on object:text-changed" +
			" (event-driven; no polling quantum; one spawn per run)\n")
	} else {
		b.WriteString("  witness: resident line-protocol poller witness-poll" +
			" (fixed 5 ms quantum — WITNESS_POLL_QUANTUM_S in focus_helper.py)\n")
	}
	fmt.Fprintf(&b, "  case: %d repeats of one homogeneous hot case — combo %s:"+
		" word correction + layout flip in one chord (D-43)\n", len(samples), combo)
	b.WriteString("  surface: zenity entry, fresh per repeat (the closing Enter is the ADR-004 buffer reset)\n")
	b.WriteString("  daemon: production form (no -config, no -debug; built-in defaults)\n")
	fmt.Fprintf(&b, "samples: %d\n", len(samples))
	fmt.Fprintf(&b, "p50: %.1f ms\n", float64(p50)/float64(time.Millisecond))
	fmt.Fprintf(&b, "p95: %.1f ms\n", float64(p95)/float64(time.Millisecond))
	fmt.Fprintf(&b, "p99: %.1f ms\n", float64(p99)/float64(time.Millisecond))
	fmt.Fprintf(&b, "vmrss_start_kb: %d\n", rssStart)
	fmt.Fprintf(&b, "vmrss_end_kb: %d\n", rssEnd)
	fmt.Fprintf(&b, "vm_hwm_kb: %d\n", hwm)
	if verdict != nil {
		fmt.Fprintf(&b, "budget: FAIL (%v)\n", verdict)
	} else {
		fmt.Fprintf(&b, "budget: PASS (p95 < %v, vm_hwm_kb < %d)\n", perfLatencyBudget, perfMemoryBudgetKB)
	}

	report := b.String()
	fmt.Print(report)
	if err := os.WriteFile(perfReportPath, []byte(report), configFilePerm); err != nil {
		return fmt.Errorf("write %s: %w", perfReportPath, err)
	}

	return nil
}
