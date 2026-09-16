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
	python3HelperBin  = "/usr/bin/python3" //nolint:unused // staged (Task 2 consumes)
	witnessLineFields = 4                  //nolint:unused // staged (Task 2 consumes)
	witnessScanInit   = 64 * 1024          //nolint:unused // staged (Task 2 consumes)
	witnessScanMax    = 1024 * 1024        //nolint:unused // staged (Task 2 consumes)
	witnessLineQueue  = 64                 //nolint:unused // staged (Task 2 consumes)
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
//
//nolint:unused // staged for the perf case (04-04 Task 2 is the consumer)
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
//
//nolint:unused // staged for the perf case (04-04 Task 2 is the consumer)
type witnessEvent struct {
	at   time.Time
	app  string
	pid  int
	text string
}

// parseWitnessLine splits one witness protocol line. The text (last field)
// may contain spaces; the app name and pid may not.
//
//nolint:unused // staged for the perf case (04-04 Task 2 is the consumer)
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
//
//nolint:unused // staged for the perf case (04-04 Task 2 is the consumer)
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
//
//nolint:unused // staged for the perf case (04-04 Task 2 is the consumer)
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
//
//nolint:unused // staged for the perf case (04-04 Task 2 is the consumer)
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
//
//nolint:unused // staged for the perf case (04-04 Task 2 is the consumer)
func (w *perfWitness) stderr() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	return strings.Join(w.errs, "; ")
}

// shutdown kills the witness process by pid and reaps it.
//
//nolint:unused // staged for the perf case (04-04 Task 2 is the consumer)
func (w *perfWitness) shutdown() {
	if w.cmd.Process != nil {
		_ = w.cmd.Process.Kill()
	}
	_ = w.cmd.Wait()
}
