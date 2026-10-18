package session_test

// The round-5 emission-sequence instrument (the 2026-10-17 wezterm
// double-delete hunt): the terminal delivery machine's INFO rows gain a
// per-burst identity — a monotonic emission-seq counter and a t_rel_ms
// timeline offset on EVERY «correction delivery» row, plus burst-window
// open/close rows bracketing the wire-emission window. The instrument is
// content-free by absolute rule (D-20/D-21): the juxtaposition with the
// selfdrive pty log must stay decisive — the daemon says one burst (one
// seq) while the pty shows two delete runs ⇒ the client replayed its
// queue; two seqs on the wire ⇒ the daemon fired twice.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Djarvur/goswitch/internal/engine"
	"github.com/Djarvur/goswitch/internal/session"
)

// dmBurstStage is the burst-window row's stage label (goconst).
const dmBurstStage = "burst-window"

// dmRow is one parsed «correction delivery» log record: the attributes the
// instrument pins, decoded from the JSON line.
type dmRow struct {
	line        string
	stage       string
	phase       string
	forwards    int
	emissionSeq int
	tRelMs      int
	hasForwards bool
}

// dmRows parses the captured log into the delivery rows in arrival order.
func dmRows(t *testing.T, logged string) []dmRow {
	t.Helper()

	var rows []dmRow
	for _, ln := range strings.Split(logged, "\n") {
		if !strings.Contains(ln, `"msg":"correction delivery"`) {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(ln), &rec); err != nil {
			t.Fatalf("delivery row is not JSON: %q: %v", ln, err)
		}
		row := dmRow{line: ln, stage: dmStr(rec, "stage"), phase: dmStr(rec, "phase")}
		if v, ok := rec["forwards"].(float64); ok {
			row.forwards, row.hasForwards = int(v), true
		}
		seq, ok := rec["emission-seq"].(float64)
		if !ok {
			t.Fatalf("delivery row without emission-seq (the round-5 instrument is missing): %q", ln)
		}
		rel, ok := rec["t_rel_ms"].(float64)
		if !ok {
			t.Fatalf("delivery row without t_rel_ms (the round-5 instrument is missing): %q", ln)
		}
		row.emissionSeq, row.tRelMs = int(seq), int(rel)
		rows = append(rows, row)
	}

	return rows
}

// dmStr returns a string attribute of the decoded record ("" when absent
// or not a string).
func dmStr(rec map[string]any, key string) string {
	s, _ := rec[key].(string)

	return s
}

// dmStages renders the rows' stage sequence — the order oracle's input.
func dmStages(rows []dmRow) []string {
	stages := make([]string, 0, len(rows))
	for _, r := range rows {
		stages = append(stages, r.stage+":"+r.phase)
	}

	return stages
}

// dmOneSeq fails unless every row carries the given burst identity.
func dmOneSeq(t *testing.T, rows []dmRow, want int) {
	t.Helper()

	for _, r := range rows {
		if r.emissionSeq != want {
			t.Fatalf("delivery row of burst %d carries emission-seq %d, want %d; line:\n%s",
				want, r.emissionSeq, want, r.line)
		}
	}
}

// dmRunTerminalRound drives one literal terminal correction round (the
// learned class, the typed corpus word, the double Right Shift gesture).
func dmRunTerminalRound(t *testing.T, a *session.Actor, word string) {
	t.Helper()

	typeWord(a, word)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
}

// TestDeliveryBurst_WindowRowsBracketTheWire pins the bracket: one terminal
// round lands exactly one burst-window open before chunks-sent and one
// close after commit-sent, with the forward count riding both.
func TestDeliveryBurst_WindowRowsBracketTheWire(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	teachTerminalClass(t, a)

	dmRunTerminalRound(t, a, wordEN)

	rows := dmRows(t, buf.String())
	var opens, closes int
	for _, r := range rows {
		if r.stage != dmBurstStage {
			continue
		}
		if !r.hasForwards || r.forwards != len(wordEN) {
			t.Fatalf("burst-window %s forwards = %d (present=%v), want %d; line:\n%s",
				r.phase, r.forwards, r.hasForwards, len(wordEN), r.line)
		}
		switch r.phase {
		case "open":
			opens++
		case "close":
			closes++
		}
	}
	if opens != 1 || closes != 1 {
		t.Fatalf("burst-window rows: %d opens, %d closes, want 1 and 1; stages: %v",
			opens, closes, dmStages(rows))
	}

	dmAssertBracketOrder(t, rows)
}

// dmAssertBracketOrder pins the wire-window sequence: burst-window open →
// chunks-sent → commit-sent → burst-window close, each strictly after the
// previous (a bracket that loses its order proves nothing about the wire).
func dmAssertBracketOrder(t *testing.T, rows []dmRow) {
	t.Helper()

	joined := dmStages(rows)
	want := []string{dmBurstStage + ":open", "chunks-sent:", "commit-sent:", dmBurstStage + ":close"}
	at := 0
	for _, w := range want {
		found := false
		for ; at < len(joined); at++ {
			if strings.HasPrefix(joined[at], w) {
				found = true
				at++

				break
			}
		}
		if !found {
			t.Fatalf("delivery rows out of order — %q never follows in sequence %v", w, joined)
		}
	}
}

// TestDeliveryBurst_SeqOnEveryRowAndMonotonicAcrossRounds pins the identity:
// every delivery row of a burst carries the burst's seq, the timeline opens
// at t_rel_ms=0, and the NEXT burst on the same context carries seq+1 —
// the juxtaposition contract against the pty log.
func TestDeliveryBurst_SeqOnEveryRowAndMonotonicAcrossRounds(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	teachTerminalClass(t, a)

	dmRunTerminalRound(t, a, wordEN)
	rowsOne := dmRows(t, buf.String())
	if len(rowsOne) == 0 {
		t.Fatalf("no delivery rows for the first burst; log:\n%s", buf.String())
	}
	dmOneSeq(t, rowsOne, 1)
	if rowsOne[0].tRelMs != 0 {
		t.Fatalf("the burst's first row opens at t_rel_ms=%d, want 0 (the timeline anchor)", rowsOne[0].tRelMs)
	}

	// The second round on the same context: fresh token, same machine —
	// the burst identity must MOVE (two seqs on the wire ⇒ a genuine
	// second daemon emission, never a client replay).
	dmRunTerminalRound(t, a, wordEN)
	rowsTwo := dmRows(t, buf.String())
	var seqTwo []dmRow
	for _, r := range rowsTwo {
		if r.emissionSeq == 2 {
			seqTwo = append(seqTwo, r)
		}
	}
	if len(seqTwo) == 0 {
		t.Fatalf("no delivery rows carry emission-seq 2 after the second round; stages: %v", dmStages(rowsTwo))
	}
	for _, r := range rowsTwo {
		if r.emissionSeq != 1 && r.emissionSeq != 2 {
			t.Fatalf("delivery row carries emission-seq %d, want 1 or 2; line:\n%s", r.emissionSeq, r.line)
		}
	}
	if seqTwo[0].stage != "burst-window" || seqTwo[0].phase != "open" {
		t.Fatalf("the second burst opens at %s:%s, want burst-window:open; stages: %v",
			seqTwo[0].stage, seqTwo[0].phase, dmStages(rowsTwo))
	}
}

// TestDeliveryBurst_FlipRowsCarryTheirBurst pins the flip rows: the armed
// terminal flip and its keystroke release both stamp the burst the flip
// was armed in — even though the release lands long after the burst closed.
func TestDeliveryBurst_FlipRowsCarryTheirBurst(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	teachTerminalClass(t, a)
	a.SetOptions(flipOnOptions())

	dmRunTerminalRound(t, a, wordEN)

	rows := dmRows(t, buf.String())
	dmOneSeq(t, rows, 1)
	var armed dmRow
	for _, r := range rows {
		if r.stage == "flip-armed" {
			armed = r
		}
	}
	if armed.line == "" {
		t.Fatalf("no flip-armed row; stages: %v", dmStages(rows))
	}

	a.HandleKey(engine.EngineEvent{Keyval: 'x'}) // the honest release cause: the next keystroke

	rows = dmRows(t, buf.String())
	var released dmRow
	for _, r := range rows {
		if r.stage == "flip-released" {
			released = r
		}
	}
	if released.line == "" {
		t.Fatalf("no flip-released row after the keystroke; stages: %v", dmStages(rows))
	}
	dmOneSeq(t, rows, 1)
	if released.tRelMs < armed.tRelMs {
		t.Fatalf("flip-released t_rel_ms=%d sits before flip-armed t_rel_ms=%d — the timeline ran backwards",
			released.tRelMs, armed.tRelMs)
	}
}

// TestDeliveryBurst_RowsAreContentFree pins the absolute D-20/D-21 rule for
// the instrument: no «correction delivery» row ever carries the typed token
// or the committed word — the instrument stays content-free even when the
// correction corpus rides the same process log.
func TestDeliveryBurst_RowsAreContentFree(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	teachTerminalClass(t, a)

	dmRunTerminalRound(t, a, wordEN)

	for _, r := range dmRows(t, buf.String()) {
		if strings.Contains(r.line, wordEN) || strings.Contains(r.line, wordRU) {
			t.Fatalf("delivery row carries field content — the D-20/D-21 rule is broken; line:\n%s", r.line)
		}
	}
}
