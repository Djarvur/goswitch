package main

// The term-word-drain oracle's unit corpus (WR-08 of the 10-REVIEW): the
// cause-honest half of drainMarkOrder must judge the CURRENT round's segment
// only. The stand matrix accumulates all three combo rows into ONE journal
// (restartDaemonWithArgs never rotates), so an unscoped Contains lets a
// PREVIOUS row's honest release cause close the CURRENT row even when the
// current release rode a dishonest reason (the round-2 geometry — a grace or
// verify release — the very class the oracle exists to convict).

import (
	"strings"
	"testing"
)

// drainSyntheticRow builds one round's delivery-mark segment: the three
// ordered marks plus the flip-released row with the GIVEN reason (an empty
// reason omits the release row — the regression form where the round never
// released at all).
func drainSyntheticRow(reason string) string {
	b := &strings.Builder{}
	b.WriteString(`{"msg":"correction delivery","stage":"chunks-sent","forwards":7}` + "\n")
	b.WriteString(`{"msg":"correction delivery","stage":"commit-sent","latency_ms":144}` + "\n")
	b.WriteString(`{"msg":"correction delivery","stage":"flip-armed"}` + "\n")
	if reason != "" {
		release := `{"msg":"correction delivery","stage":"flip-released","reason":"` +
			reason + `","latency_ms":104}` + "\n"
		b.WriteString(release)
	}

	return b.String()
}

// TestDrainMarkOrder_HonestCausePasses pins the positive form first: a
// single honest round — the release reason a keystroke or a push — orders
// and closes cause-honest.
func TestDrainMarkOrder_HonestCausePasses(t *testing.T) {
	for _, cause := range []string{"keystroke", "push"} {
		log := drainSyntheticRow(cause)
		ok, obs := drainMarkOrder(log)
		if !ok {
			t.Errorf("cause %q: honest round judged not-ok (%s)", cause, obs)
		}
	}
}

// TestDrainMarkOrder_DishonestCauseInCurrentRoundFails is the WR-08 RED
// form: the CURRENT round's release rode a dishonest reason (grace — the
// round-2 geometry back from the dead) and the journal carries an EARLIER
// row's honest push release. The oracle must fail the row on the current
// segment; an unscoped whole-journal Contains would close it on the
// previous row's cause. RED truth: the cause search greps the WHOLE
// accumulated journal — the dishonest round passes on the borrowed cause.
func TestDrainMarkOrder_DishonestCauseInCurrentRoundFails(t *testing.T) {
	log := drainSyntheticRow("push") + // row 1 (the earlier combo): honest push release
		drainSyntheticRow("grace") // row 2 (the current combo): the dishonest release
	ok, obs := drainMarkOrder(log)
	if ok {
		t.Fatalf("a dishonest grace release closed the row on a previous row's push cause (%s)", obs)
	}
}

// TestDrainMarkOrder_NoReleaseInCurrentRoundFails pins the sibling form: the
// current round never released at all — no release row after its arm. The
// borrowed marks of the earlier row must not close it.
func TestDrainMarkOrder_NoReleaseInCurrentRoundFails(t *testing.T) {
	log := drainSyntheticRow("push") + // row 1: honest release
		drainSyntheticRow("") // row 2: armed, never released
	ok, obs := drainMarkOrder(log)
	if ok {
		t.Fatalf("a never-released round closed the row on the previous row's marks (%s)", obs)
	}
}
