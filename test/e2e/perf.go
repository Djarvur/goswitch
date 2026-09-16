package main

import (
	"time"
)

// RED stubs (plan 04-04 Task 1, D-07): the pure perf mathematics of the
// INST-03 acceptance run, pinned by perf_test.go before implementation.

// percentile is the p50/p95/p99 estimator over one homogeneous sample.
func percentile(samples []time.Duration, p float64) time.Duration {
	return 0 // RED stub
}

// parseProcStatus parses VmHWM/VmRSS out of a /proc/<pid>/status document.
func parseProcStatus(data string) (hwm, rss int64, err error) {
	return 0, 0, nil // RED stub
}

// budgetGate turns the INST-03 budget into an acceptance gate.
func budgetGate(p95 time.Duration, vmHWMKB int64) error {
	return nil // RED stub
}
