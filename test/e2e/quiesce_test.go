package main

// Quiesce-budget corpus (plan 04-08, gap G-4-1): the formal fresh-session
// D-48 gate failed HEALTHY cases in both attempts (runs 35142914380 and
// 35188570260, 04-UAT.md) because the witness probe's bare full-tree walk
// (~3.0s across 6100 shell+gjs nodes on a fresh session) crossed the 4s
// witnessProbeTimeout under any additive latency, exhausting the 30s
// quiesce window. The resolution pinned here is pure constant arithmetic,
// so it is proven without a live desktop; the live step-runner itself is
// exercised by `mise run e2e-matrix-v3`.

import (
	"testing"
	"time"
)

// TestWitnessProbeBudget_ScalabilityFloor pins the probe-budget floor
// (G-4-1 fix direction, defense in depth): the measured worst case is the
// bare full-tree walk of a fresh-session desktop (3673 gnome-shell + 2437
// gjs nodes, ~3.0s) and it crossed the former 4s budget under load; 10s
// keeps roughly 3x headroom over that measured worst case while the
// helper's focus-first walk keeps real probes sub-second.
func TestWitnessProbeBudget_ScalabilityFloor(t *testing.T) {
	if witnessProbeTimeout < 10*time.Second {
		t.Fatalf("witnessProbeTimeout = %v, want >= 10s (G-4-1: the bare fresh-session walk ~3.0s crossed the former 4s budget)", witnessProbeTimeout)
	}
}

// TestMatrixQuiesce_WindowCoversProbes pins the quiesce-window coverage
// invariant: the window must cover a FULL probe on every attempt plus the
// inter-probe pauses (2 attempts × (10s probe + 100ms pause) = 20.2s <=
// window). The invariant catches future constant drift — anyone raising
// the probe budget without re-proportioning the window fails here before
// a live run does.
func TestMatrixQuiesce_WindowCoversProbes(t *testing.T) {
	need := time.Duration(witnessQuieceAttempts) * (witnessProbeTimeout + witnessPoll)
	if need > witnessQuiesceWindow {
		t.Fatalf("quiesce window %v does not cover %d full probes plus pauses (%v); the deadline would exhaust mid-probe", witnessQuiesceWindow, witnessQuieceAttempts, need)
	}
}

// TestMatrixQuiesce_WindowSource pins the single-source rule: the loop
// deadline and the error text read ONE named constant, never a re-derived
// product (the expression was duplicated in matrixQuiesce before 04-08).
func TestMatrixQuiesce_WindowSource(t *testing.T) {
	if witnessQuiesceWindow != time.Duration(witnessQuieceAttempts)*surfaceFocusWait {
		t.Fatalf("witnessQuiesceWindow = %v, want witnessQuieceAttempts*surfaceFocusWait = %v — the deadline and the error text must share one source", witnessQuiesceWindow, time.Duration(witnessQuieceAttempts)*surfaceFocusWait)
	}
}
