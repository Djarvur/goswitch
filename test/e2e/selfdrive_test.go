package main

// The selfdrive pty-log parser corpus (the round-5 instrument, TDD): the
// driver's verdicts ride ONE machine — parsePaneEvents turns a raw
// `script -q` fragment into merged erase/text runs, analyzeRunWindow
// classifies the choreography. Every vector here is a synthetic fragment:
// the marker echo + its erase, the word echo, the correction's erase run,
// the committed word — and (the defect) the second delete run. Content
// discipline: the fragments carry the test corpus only.

import (
	"strings"
	"testing"
)

// ssEchoErase is the cooked-mode echo of ONE erased character: the line
// discipline (or readline) answers a BackSpace with backspace-space-
// backspace.
func ssEchoErase(n int) string {
	return strings.Repeat("\b \b", n)
}

// TestSelfdriveParse_EraseForms pins the erase vocabulary: the cooked
// \b-space-\b triple is ONE erase, a bare backspace is one erase, the raw
// 0x7f capture form is one erase.
func TestSelfdriveParse_EraseForms(t *testing.T) {
	for _, tc := range []struct {
		name string
		frag []byte
		want int // total erase count across the fragment
	}{
		{"cooked-triple", ssFrag(ssEchoErase(3)), 3},
		{"bare-backspace", ssFrag("\b\b\b"), 3},
		{"raw-del", ssFrag("\x7f\x7f\x7f\x7f\x7f\x7f"), 6},
		{"mixed", ssFrag("\b \b\x7f\b"), 3},
	} {
		events := parsePaneEvents(tc.frag)
		if got := ssCountKind(events, ssEvErase); got != tc.want {
			t.Fatalf("%s: erase count = %d, want %d; events: %s", tc.name, got, tc.want, ssSymbolic(events))
		}
	}
}

// TestSelfdriveParse_TextCountsRunes pins the rune discipline: a UTF-8
// word's text run counts RUNES (the choreography compares 6 against 6),
// never bytes.
func TestSelfdriveParse_TextCountsRunes(t *testing.T) {
	events := parsePaneEvents([]byte(ssWordRU))
	if len(events) != 1 || events[0].kind != ssEvText || events[0].runes != len([]rune(ssWordRU)) {
		t.Fatalf("привет parsed as %s, want one text run of %d runes", ssSymbolic(events), len([]rune(ssWordRU)))
	}
}

// TestSelfdriveParse_NeutralNoise pins the neutral vocabulary: terminal
// escape sequences and CR/LF never become events and never split a text
// run into phantom choreography.
func TestSelfdriveParse_NeutralNoise(t *testing.T) {
	events := parsePaneEvents([]byte("\x1b[?25l" + ssWordRU + "\r\n" + ssWordRU))
	if len(events) != 1 || events[0].kind != ssEvText || events[0].runes != 2*len([]rune(ssWordRU)) {
		t.Fatalf("noise-prefixed words parsed as %s, want one merged text run", ssSymbolic(events))
	}
}

// TestSelfdriveParse_MergesConsecutiveRuns pins the run merging: adjacent
// same-kind output is one event — six erase echoes land as E6, not six
// events (the choreography zone must read E6 T6, never noise).
func TestSelfdriveParse_MergesConsecutiveRuns(t *testing.T) {
	events := parsePaneEvents(ssFrag(corpusLatin, ssEchoErase(6), ssWordRU))
	if got := ssSymbolic(events); got != "T6 E6 T6" {
		t.Fatalf("word/erase/word parsed as %q, want \"T6 E6 T6\"", got)
	}
}

// TestSelfdriveClassify_Clean pins the green choreography: marker, marker
// erase, the word echo, the correction's erase run, the committed word —
// and NOTHING after it.
func TestSelfdriveClassify_Clean(t *testing.T) {
	frag := ssFrag(ssMarker, ssEchoErase(1), corpusLatin, ssEchoErase(6), ssWordRU)
	rep := analyzeRunWindow(parsePaneEvents(frag), len([]rune(corpusLatin)))
	if rep.choreo != ssChoreoClean {
		t.Fatalf("clean choreography classified %q (second-delete=%v, events %q), want clean",
			rep.choreo.word(), rep.secondDel, rep.events)
	}
	if rep.secondDel {
		t.Fatal("clean choreography carries second-delete=true")
	}
}

// TestSelfdriveClassify_Replay pins the defect signature: a SECOND erase
// run after the committed word — the client replayed its queued forwards
// (the 2026-10-17 wezterm double-delete). A replay that re-inserts as well
// (the full E T E T form) is still a replay: the symbolic events string
// distinguishes the tail for the journal juxtaposition.
func TestSelfdriveClassify_Replay(t *testing.T) {
	frag := ssFrag(ssMarker, ssEchoErase(1), corpusLatin, ssEchoErase(6), ssWordRU, ssEchoErase(6))
	rep := analyzeRunWindow(parsePaneEvents(frag), len([]rune(corpusLatin)))
	if rep.choreo != ssChoreoReplay || !rep.secondDel {
		t.Fatalf("replay choreography classified %q second-delete=%v, want replay+true",
			rep.choreo.word(), rep.secondDel)
	}

	frag = ssFrag(ssMarker, ssEchoErase(1), corpusLatin, ssEchoErase(6), ssWordRU, ssEchoErase(6), ssWordRU)
	rep = analyzeRunWindow(parsePaneEvents(frag), len([]rune(corpusLatin)))
	if rep.choreo != ssChoreoReplay {
		t.Fatalf("replay-with-reinsert classified %q, want replay (events %q)", rep.choreo.word(), rep.events)
	}
}

// TestSelfdriveClassify_NegativeForms pins the honest negatives: nothing
// arrived (empty), the word echoed but no correction fired (no-correction),
// the forwards landed without a visible commit (erase-only), and a foreign
// shape (unexpected).
func TestSelfdriveClassify_NegativeForms(t *testing.T) {
	words := len([]rune(corpusLatin))
	for _, tc := range []struct {
		name string
		frag []byte
		want ssChoreo
	}{
		{"empty", ssFrag(ssMarker, ssEchoErase(1)), ssChoreoEmpty},
		{"no-correction", ssFrag(ssMarker, ssEchoErase(1), corpusLatin), ssChoreoNoGesture},
		{"erase-only", ssFrag(ssMarker, ssEchoErase(1), corpusLatin, ssEchoErase(6)), ssChoreoEraseOnly},
		{"unexpected-double-word", ssFrag(ssMarker, ssEchoErase(1), corpusLatin, corpusLatin), ssChoreoUnexpected},
		{
			"unexpected-after-commit",
			ssFrag(ssMarker, ssEchoErase(1), corpusLatin, ssEchoErase(6), ssWordRU, ssWordRU),
			ssChoreoUnexpected,
		},
	} {
		rep := analyzeRunWindow(parsePaneEvents(tc.frag), words)
		if rep.choreo != tc.want {
			t.Fatalf("%s: classified %q (events %q), want %q", tc.name, rep.choreo.word(), rep.events, tc.want.word())
		}
	}
}

// TestSelfdriveClassify_ToleratesLeadingNoise pins the lenient marker
// skip: a prompt redraw before the marker echo must not break the
// classification (the window offset is taken before the marker, but the
// pane may have flushed unrelated bytes into the same window).
func TestSelfdriveClassify_ToleratesLeadingNoise(t *testing.T) {
	frag := ssFrag("\x1b[?2004h", ssMarker, ssEchoErase(1), corpusLatin, ssEchoErase(6), ssWordRU)
	rep := analyzeRunWindow(parsePaneEvents(frag), len([]rune(corpusLatin)))
	if rep.choreo != ssChoreoClean {
		t.Fatalf("noise-tolerant clean classified %q (events %q), want clean", rep.choreo.word(), rep.events)
	}
}

// ssCountKind totals the rune counts of one event kind.
func ssCountKind(events []ssEvent, kind ssEvKind) int {
	n := 0
	for _, e := range events {
		if e.kind == kind {
			n += e.runes
		}
	}

	return n
}
