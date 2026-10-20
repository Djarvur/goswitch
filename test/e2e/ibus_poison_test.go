package main

// The name-hold decision corpus (plan 09-02): when the component name will
// not be claimed, the await must DISTINGUISH a live holder (slow release —
// wait) from a poisoned name entry (listed, ownerless, unclaimable — the
// ibus generation is corrupt and one bounded restart is the cure) from a
// genuinely free name. The 2026-10-13 finding: a poisoned entry made every
// spawn die with "another instance owns the goswitch name" while the probe
// waited forever — the three states are not distinguishable by the probe
// alone.

import (
	"testing"
	"time"
)

// TestNameHoldDecision pins the await's routing table: the pure decision
// from the bus facts (name listed? owner pid?) and whether the one-shot
// ibus restart already fired.
func TestNameHoldDecision(t *testing.T) {
	for _, tc := range []struct {
		name        string
		listed      bool
		ownerPID    string
		restarted   bool
		wantAction  holdAction
		wantRestart bool
	}{
		{
			name:       "not listed — free, probe again",
			listed:     false,
			wantAction: holdProbeAgain,
		},
		{
			name:       "live owner — slow release, wait",
			listed:     true,
			ownerPID:   "1182235",
			wantAction: holdWait,
		},
		{
			name:        "listed, ownerless, restart not yet fired — the poison: restart once",
			listed:      true,
			restarted:   false,
			wantAction:  holdRestart,
			wantRestart: true,
		},
		{
			name:       "listed, ownerless, restart already fired — keep waiting inside the budget",
			listed:     true,
			restarted:  true,
			wantAction: holdWait,
		},
	} {
		got, restart := nameHoldDecision(tc.listed, tc.ownerPID, tc.restarted)
		if got != tc.wantAction {
			t.Fatalf("%s: nameHoldDecision = %v, want %v", tc.name, got, tc.wantAction)
		}
		if restart != tc.wantRestart {
			t.Fatalf("%s: restart = %v, want %v", tc.name, restart, tc.wantRestart)
		}
	}
}

// TestBodyMentionsStringSlice pins the ListNames reply shape: the name
// list decodes as []string inside the body, and the poison detector's
// presence oracle must see through it (the regression that made `listed`
// permanently false and busy-spun the await).
func TestBodyMentionsStringSlice(t *testing.T) {
	body := []any{[]string{"org.freedesktop.IBus", "org.freedesktop.IBus.Config", "org.freedesktop.IBus.goswitch"}}
	if !bodyMentions(body, "org.freedesktop.IBus.goswitch") {
		t.Fatal("bodyMentions missed the name inside a []string reply (ListNames shape)")
	}
	if bodyMentions(body, "org.freedesktop.IBus.other") {
		t.Fatal("bodyMentions matched a name that is not in the list")
	}
}

// TestNameHoldDecisionBudgetStillBounds pins that the decision never
// suspends the budget: the caller's deadline check stays in charge — the
// restart action is a one-shot, everything else is a wait-and-reprobe.
func TestNameHoldDecisionBudgetStillBounds(t *testing.T) {
	deadline := time.Now().Add(50 * time.Millisecond)
	action, _ := nameHoldDecision(true, "", false)
	if action != holdRestart {
		t.Fatalf("poisoned name must route to the restart, got %v", action)
	}
	time.Sleep(60 * time.Millisecond)
	if time.Now().Before(deadline) {
		t.Fatal("the decision must not manipulate the caller's deadline")
	}
}
