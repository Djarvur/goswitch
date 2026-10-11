package main

// The arming-gate primitives corpus (plan 09-02, the R1 fix): the gate keys
// on the fixture's own pty content stream — inject a probe line, demand it
// in the capture, strip the probe lines from the oracle reads — because the
// R1 live finding (2026-10-11) is that injected keys reach the fixture
// window WITHOUT traversing the IM engine, so the daemon key-trace can
// never arm this gate. The primitives are pure file/text logic, so they
// are pinned here without a live desktop; the live gate itself is exercised
// by the spike cases under `mise run e2e-*`.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestWaitForArmingProbe_Lands pins the pass half of the gate's wait: a
// probe that appears in the capture after a delay resolves the wait (the
// arming probe is written by the fixture's cat, i.e. asynchronously — the
// wait must poll, never single-shot).
func TestWaitForArmingProbe_Lands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capture")
	if err := os.WriteFile(path, []byte("prefix\n"), probeFilePerm); err != nil {
		t.Fatalf("seed capture: %v", err)
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = os.WriteFile(path, []byte("prefix\n"+armProbeText+"\n"), probeFilePerm)
	}()

	if err := waitForArmingProbe(context.Background(), path, 2*time.Second); err != nil {
		t.Fatalf("waitForArmingProbe = %v, want nil once the probe lands", err)
	}
}

// TestWaitForArmingProbe_TimeoutNamesTheTarget pins the fail half: a probe
// that never lands times out with the file and the probe named — the R1
// failure named neither ("no key-trace record"), which cost a whole batch
// round of triage.
func TestWaitForArmingProbe_TimeoutNamesTheTarget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capture")
	if err := os.WriteFile(path, []byte("unrelated\n"), probeFilePerm); err != nil {
		t.Fatalf("seed capture: %v", err)
	}

	err := waitForArmingProbe(context.Background(), path, 120*time.Millisecond)
	if err == nil {
		t.Fatal("waitForArmingProbe = nil, want timeout error")
	}
	for _, want := range []string{path, armProbeText} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("waitForArmingProbe error %q does not name %q", err.Error(), want)
		}
	}
}

// TestWaitForArmingProbe_AbortStopsTheWait pins the Ctrl-C contract: an
// aborted context interrupts the poll (sleepCtx discipline) instead of
// burning the whole timeout.
func TestWaitForArmingProbe_AbortStopsTheWait(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capture")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan error, 1)
	go func() { done <- waitForArmingProbe(ctx, path, 30*time.Second) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("waitForArmingProbe = nil, want abort error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("waitForArmingProbe ignored the cancelled context")
	}
}

// TestNoteGate_FreshGateHoldsTheFirstInterval pins the wait-gate progress
// contract after the sound-cue removal (the owner's 2026-10-12 directive:
// silence is the default): a fresh gate holds its note interval — the
// waits start quiet, the console note is the only progress signal.
func TestNoteGate_FreshGateHoldsTheFirstInterval(t *testing.T) {
	if g := newNoteGate(); g.due() {
		t.Fatal("fresh noteGate.due = true, want quiet for the first interval")
	}
}

// TestNoteGate_NoteCadence pins the note half of the gate: once the note
// interval has elapsed exactly one due() fires and the interval re-arms —
// the 15 s console cadence the long waits keep.
func TestNoteGate_NoteCadence(t *testing.T) {
	g := noteGate{nextNote: time.Now().Add(-time.Second)}
	if !g.due() {
		t.Fatal("noteGate.due = false with the interval elapsed, want one note")
	}
	if g.due() {
		t.Fatal("noteGate.due = true twice in a row, want the interval re-armed")
	}
}

// TestClassifyArm pins the arming gate's verdict semantics (the live-smoke
// finding, 2026-10-12: the engine traced every probe while the fixture pty
// stayed empty — the engine→stale-context desync IS the TERM-01 defect
// measured, not a harness failure): the pty probe wins whenever it lands
// (window delivery proven, engine delta merely diagnostic), an engine trace
// without pty bytes is the reproduced desync, and only the absence of BOTH
// signals means the injection itself broke.
func TestClassifyArm(t *testing.T) {
	for _, tc := range []struct {
		name        string
		ptyHasProbe bool
		engineDelta int
		want        armOutcome
	}{
		{"window delivery wins", true, 40, armWindow},
		{"window delivery without engine trace", true, 0, armWindow},
		{"engine traced, pty empty — desync", false, 40, armStaleContext},
		{"engine traced once, pty empty", false, 2, armStaleContext},
		{"neither signal — dead injection", false, 0, armDead},
	} {
		if got := classifyArm(tc.ptyHasProbe, tc.engineDelta); got != tc.want {
			t.Fatalf("%s: classifyArm(%v, %d) = %v, want %v", tc.name, tc.ptyHasProbe, tc.engineDelta, got, tc.want)
		}
	}
}

// TestStripArmingProbe pins the oracle-read stripping: the gate's repeated
// probe bytes plus their line endings go from the head of the capture
// (retries included, canonical \n and raw \r endings alike), and the typed
// corpus stays byte-exact; a capture without probe bytes passes through
// untouched.
func TestStripArmingProbe(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"single probe prefix", armProbeText + "\n" + corpusLatin, corpusLatin},
		{"retried probes", armProbeText + "\n" + armProbeText + "\n" + corpusLatin, corpusLatin},
		{"raw mode endings", armProbeText + "\r" + armProbeText + "\r" + corpusLatin, corpusLatin},
		{"probe only", armProbeText + "\n", ""},
		{"smoke shape", armProbeText + "\nx", "x"},
		{"no probe passthrough", corpusLatin, corpusLatin},
	} {
		if got := string(stripArmingProbe([]byte(tc.in))); got != tc.want {
			t.Fatalf("%s: stripArmingProbe(%q) = %q, want %q", tc.name, tc.in, got, tc.want)
		}
	}
}

// TestReadPtyHexArmed_StripsAndHexes pins the armed oracle read end to end:
// the capture's arming lines are stripped before hexlification, so the flip
// verdicts compare exactly what the round under test typed.
func TestReadPtyHexArmed_StripsAndHexes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capture")
	if err := os.WriteFile(path, []byte(armProbeText+"\n"+armProbeText+"\n"+corpusLatin), probeFilePerm); err != nil {
		t.Fatalf("seed capture: %v", err)
	}

	hex, err := readPtyHexArmed(path)
	if err != nil {
		t.Fatalf("readPtyHexArmed: %v", err)
	}
	if hex != corpusHexLatin {
		t.Fatalf("readPtyHexArmed = %s, want the bare corpus hex %s", hex, corpusHexLatin)
	}
}

// TestReadPtyHexArmed_UnarmedCaptureIsAnError pins the misuse guard: the
// gate guarantees an arming line before any oracle read, so a capture
// without one is a named harness failure — never a plausible wrong hex.
func TestReadPtyHexArmed_UnarmedCaptureIsAnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capture")
	if err := os.WriteFile(path, []byte(corpusLatin), probeFilePerm); err != nil {
		t.Fatalf("seed capture: %v", err)
	}

	if _, err := readPtyHexArmed(path); err == nil {
		t.Fatal("readPtyHexArmed = nil error for a capture without the arming line, want error")
	}
}
