package main

// The selfdrive pty-log parser corpus (the round-5 instrument, TDD): the
// driver's verdicts ride ONE machine — parsePaneEvents turns a raw
// `script -q` fragment into merged erase/text runs, analyzeRunWindow
// classifies the choreography. Every vector here is a synthetic fragment:
// the digit marker echo + the Ctrl-U clear echo, the word echo, the
// correction's erase run, the committed word — and (the defect) the second
// delete run. Content discipline: the fragments carry the test corpus only.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/Djarvur/goswitch/internal/uinject"
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

// TestSelfdriveClassify_Clean pins the green choreography: marker, the
// Ctrl-U clear echo, the word echo, the correction's erase run, the
// committed word — and NOTHING after it.
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

// TestSelfdriveMarker_IsDigit pins the layout independence of the focus
// proof: the marker is a DIGIT — identical keystrokes in EN and RU (the
// ЙЦУКЕН digit row matches QWERTY's), so the liveness probe cannot type
// Cyrillic garbage when the pane's IBus mode is not the one the battery
// assumes. The 2026-10-18 first live battery: the "." marker landed as
// «ю» in the RU-mode pane, the check grepped for a ghost and aborted 80
// times while «ю»-letters accumulated in the owner's line.
func TestSelfdriveMarker_IsDigit(t *testing.T) {
	runes := []rune(ssMarker)
	if len(runes) != 1 {
		t.Fatalf("marker %q carries %d runes, want exactly 1", ssMarker, len(runes))
	}
	if !unicode.IsDigit(runes[0]) {
		t.Fatalf("marker %q is not a digit — it would not survive an EN/RU layout mismatch", ssMarker)
	}
}

// TestSelfdriveClassify_CtrlUClear pins the line-clear discipline: the run
// opens with the marker echo followed by the Ctrl-U clear echo — ONE erase
// run of ANY length, because the clear wipes the marker plus whatever
// stray junk shared the line (the owner's leftovers, a previous run) — and
// the classification must see straight through it to the corpus echo. A
// window holding nothing past the clear is honestly empty.
func TestSelfdriveClassify_CtrlUClear(t *testing.T) {
	words := len([]rune(corpusLatin))
	for _, tc := range []struct {
		name  string
		clear int
		frag  []byte
		want  ssChoreo
	}{
		{
			"clear-marker-only",
			1,
			ssFrag(ssMarker, ssEchoErase(1), corpusLatin, ssEchoErase(words), ssWordRU),
			ssChoreoClean,
		},
		{
			"clear-marker-plus-junk",
			5,
			ssFrag(ssMarker, ssEchoErase(5), corpusLatin, ssEchoErase(words), ssWordRU),
			ssChoreoClean,
		},
		{
			"clear-only-empty-window",
			4,
			ssFrag(ssMarker, ssEchoErase(4)),
			ssChoreoEmpty,
		},
	} {
		rep := analyzeRunWindow(parsePaneEvents(tc.frag), words)
		if rep.choreo != tc.want {
			t.Fatalf("%s: classified %q (events %q), want %q", tc.name, rep.choreo.word(), rep.events, tc.want.word())
		}
	}
}

// TestSelfdriveParseMode pins the daemon-mode read behind the battery's
// layout gate: goswitchctl status reports `mode=en|ru` among its
// key=value fields, and the parser reads that token and nothing else — a
// report without a mode field reads empty (fail-closed: no guessed mode,
// the gate would refuse to proceed).
func TestSelfdriveParseMode(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status string
		want   string
	}{
		{"en", "version=v1.1.1 mode=en corrections_done=3 corrections_skipped=1 config_valid=true", "en"},
		{"ru", "version=v1.1.1 mode=ru corrections_done=0 corrections_skipped=0 config_valid=true", "ru"},
		{"mode-last", "version=dev config=none mode=ru", "ru"},
		{"no-mode", "version=v1.1.1 config_valid=true", ""},
		{"empty", "", ""},
	} {
		if got := ssParseMode(tc.status); got != tc.want {
			t.Fatalf("%s: ssParseMode = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// TestSelfdriveInjectorVocabulary pins the driver's contract with the
// in-repo uinject table: EVERY rune the battery types (the marker digit,
// the correction corpus) resolves, and so do the chord keys (the Ctrl-U
// clear's LeftCtrl and u, the double Right Shift gesture) — a miss here
// would be a loud abort row at battery start; the corpus pins it before
// any battery runs at all.
func TestSelfdriveInjectorVocabulary(t *testing.T) {
	for _, r := range ssMarker + corpusLatin {
		if _, ok := uinject.Keycode(string(r)); !ok {
			t.Fatalf("rune %q of the driver's typing vocabulary (marker %q, corpus %q) has no keycode",
				r, ssMarker, corpusLatin)
		}
	}
	for _, name := range []string{ssKeyLeftCtrl, "u", "ShiftRight", "ShiftRight"} {
		if _, ok := uinject.Keycode(name); !ok {
			t.Fatalf("chord key %q of the driver's gestures has no keycode", name)
		}
	}
}

// TestSelfdriveCtrlUChord pins the explicit Ctrl-U chord the ydotool
// "ctrl+u" form became: exactly four steps, modifier down FIRST and up
// LAST (a true simultaneous chord), the letter between — every step a
// name the injector's table resolves.
func TestSelfdriveCtrlUChord(t *testing.T) {
	if len(ssCtrlUSteps) != 4 {
		t.Fatalf("the Ctrl-U chord carries %d steps, want 4 (ctrl down, u down, u up, ctrl up)", len(ssCtrlUSteps))
	}
	want := []struct {
		name string
		down bool
	}{
		{ssKeyLeftCtrl, true},
		{"u", true},
		{"u", false},
		{ssKeyLeftCtrl, false},
	}
	for i, step := range ssCtrlUSteps {
		if step.name != want[i].name || step.down != want[i].down {
			t.Fatalf("chord step %d = {%s down=%v}, want {%s down=%v}",
				i, step.name, step.down, want[i].name, want[i].down)
		}
		if _, ok := uinject.Keycode(step.name); !ok {
			t.Fatalf("chord step %d names %q — outside the injector's keycode table", i, step.name)
		}
	}
}

// TestSelfdriveCtlPath_NewestRelease pins the release resolution: with no
// goswitchctl on PATH the battery reads the daemon through the
// newest-by-mtime binary among ~/.local/share/goswitch/releases/*/
// (the plans' `ls -t .../releases/*/goswitchctl | head -1`), never a
// guessed version; no releases at all is a loud error, not a silent fall
// through.
func TestSelfdriveCtlPath_NewestRelease(t *testing.T) {
	home := t.TempDir()
	mk := func(rel string, mod time.Time) string {
		dir := filepath.Join(home, ".local", "share", "goswitch", "releases", rel)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		bin := filepath.Join(dir, "goswitchctl")
		if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(bin, mod, mod); err != nil {
			t.Fatal(err)
		}

		return bin
	}
	mk("v1.1.1", time.Now().Add(-time.Hour))
	newest := mk("v1.1.1-190-g29bbde6", time.Now())
	mk("v1.1.1-9-gb1dcb05", time.Now().Add(-time.Minute))

	got, err := ssNewestReleaseCtl(home)
	if err != nil {
		t.Fatalf("ssNewestReleaseCtl: %v", err)
	}
	if got != newest {
		t.Fatalf("ssNewestReleaseCtl = %q, want the newest release %q", got, newest)
	}

	if _, err := ssNewestReleaseCtl(t.TempDir()); err == nil {
		t.Fatal("ssNewestReleaseCtl on a home without releases returned nil error, want a loud failure")
	}
}
