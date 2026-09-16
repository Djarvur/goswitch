package main

// Watchdog limit resolution corpus (plan 04-06 deviation fix): the matrix
// runner passes its limit through runCaseWatchdog, and 04-04's signature
// change turned the "default, please" zero into a literal time.After(0) —
// every matrix case then failed with "did not complete within 0s" (live
// finding of the first D-48 double-run dispatch, run 35107406144). The
// resolution is pure logic, so it is pinned here without a live desktop;
// the live step-runner itself is exercised by `mise run e2e-matrix-v3`.

import (
	"testing"
	"time"
)

// TestWatchdogLimit_ZeroMeansDefault pins the caseSpec.watchdog convention
// ("A non-zero watchdog overrides the default case deadline"): zero is the
// REQUEST for the default deadline, never a zero-length budget — the same
// substitution the -case registry path has always applied to spec.watchdog.
func TestWatchdogLimit_ZeroMeansDefault(t *testing.T) {
	if got := watchdogLimit(0); got != caseTimeout {
		t.Fatalf("watchdogLimit(0) = %v, want the caseTimeout default %v", got, caseTimeout)
	}
}

// TestWatchdogLimit_ExplicitOverrides pins the override half of the
// convention: a caller's explicit budget (perf: N × its repeat budget)
// passes through untouched.
func TestWatchdogLimit_ExplicitOverrides(t *testing.T) {
	explicit := 90 * time.Second
	if got := watchdogLimit(explicit); got != explicit {
		t.Fatalf("watchdogLimit(%v) = %v, want the explicit budget unchanged", explicit, got)
	}
}
