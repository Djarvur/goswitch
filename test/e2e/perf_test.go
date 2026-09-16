package main

// Perf mathematics corpus (plan 04-04 Task 1, D-07): percentile, the
// /proc status parser and the INST-03 budget gate are pure logic pinned
// headless; the live measurement loop itself is `mise run e2e-perf` (the
// matrix_test.go discipline: headless corpus first, live run second).

import (
	"strings"
	"testing"
	"time"
)

// TestPercentile_Exact pins the percentile formula
// s[int(float64(len(s)-1))*p on the [10..100] sample: p50 = 50,
// p95 = p99 = 100 — exact values, no interpolation (04-RESEARCH
// Don't Hand-Roll: sort + index, no stats dependency).
func TestPercentile_Exact(t *testing.T) {
	t.Parallel()

	samples := []time.Duration{
		10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond,
		40 * time.Millisecond, 50 * time.Millisecond, 60 * time.Millisecond,
		70 * time.Millisecond, 80 * time.Millisecond, 90 * time.Millisecond,
		100 * time.Millisecond,
	}
	if got := percentile(samples, 0.50); got != 50*time.Millisecond {
		t.Errorf("percentile(p50) = %v, want %v", got, 50*time.Millisecond)
	}
	if got := percentile(samples, 0.95); got != 100*time.Millisecond {
		t.Errorf("percentile(p95) = %v, want %v", got, 100*time.Millisecond)
	}
	if got := percentile(samples, 0.99); got != 100*time.Millisecond {
		t.Errorf("percentile(p99) = %v, want %v", got, 100*time.Millisecond)
	}
}

// TestPercentile_SingleSample pins the degenerate one-sample case: every
// percentile of [42] is 42.
func TestPercentile_SingleSample(t *testing.T) {
	t.Parallel()

	for _, p := range []float64{0.50, 0.95, 0.99} {
		if got := percentile([]time.Duration{42 * time.Millisecond}, p); got != 42*time.Millisecond {
			t.Errorf("percentile(single, p=%v) = %v, want %v", p, got, 42*time.Millisecond)
		}
	}
}

// TestReadProcStatus pins the /proc/<pid>/status parser on fixture strings:
// VmHWM/VmRSS kB values are extracted, a missing VmHWM line is a named
// error — the D-45 oracle must never report zero silently.
func TestReadProcStatus(t *testing.T) {
	t.Parallel()

	const fixture = "Name:\tgoswitchd\nVmHWM:\t 12345 kB\nVmRSS:\t 6789 kB\nThreads:\t4\n"
	hwm, rss, err := parseProcStatus(fixture)
	if err != nil {
		t.Fatalf("parseProcStatus: %v", err)
	}
	if hwm != 12345 {
		t.Errorf("VmHWM = %d, want 12345", hwm)
	}
	if rss != 6789 {
		t.Errorf("VmRSS = %d, want 6789", rss)
	}

	const noHWM = "Name:\tgoswitchd\nVmRSS:\t 6789 kB\nThreads:\t4\n"
	if _, _, err := parseProcStatus(noHWM); err == nil || !strings.Contains(err.Error(), "no VmHWM line") {
		t.Errorf("parseProcStatus(missing VmHWM) error = %v, want the named 'no VmHWM line' error", err)
	}
}

// TestPerfBudgetGate pins the INST-03 budget as a GATE, not a metric, on
// both sides: the under-budget triple passes; p95 AT the 50 ms boundary and
// VmHWM AT the 50 MB boundary each fail (the gate exits non-zero).
func TestPerfBudgetGate(t *testing.T) {
	t.Parallel()

	if err := budgetGate(49*time.Millisecond, 49*1024); err != nil {
		t.Errorf("budgetGate(under budget) = %v, want nil", err)
	}
	if err := budgetGate(50*time.Millisecond, 49*1024); err == nil {
		t.Error("budgetGate(p95 = 50 ms) = nil, want error — the latency budget is a gate")
	}
	if err := budgetGate(10*time.Millisecond, 50*1024); err == nil {
		t.Error("budgetGate(vmHWM = 50 MB) = nil, want error — the memory budget is a gate")
	}
}
