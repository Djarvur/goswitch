//nolint:testpackage // drives the unexported reader/fake seams — the sound corpus in-package precedent
package sound

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

// toneOf builds one tone's byte payload: frames of int16-LE pairs.
func toneOf(samples ...int16) []byte {
	out := make([]byte, 2*len(samples))
	for i, s := range samples {
		binary.LittleEndian.PutUint16(out[2*i:], uint16(s)) // #nosec G115 -- the int16 bit pattern IS the PCM encoding
	}

	return out
}

// readWithBudget runs one reader callback off the test goroutine — a
// reader that lost its silence-default BLOCKS, and the budget turns the
// hang into a loud failure instead of a wedged corpus.
func readWithBudget(t *testing.T, r *toneReader, n int) ([]byte, bool) {
	t.Helper()
	done := make(chan []byte, 1)
	go func() {
		buf := make([]byte, n)
		read, _ := r.Read(buf)
		done <- buf[:read]
	}()
	select {
	case got := <-done:
		return got, true
	case <-time.After(pollBudget):
		return nil, false
	}
}

// TestToneReader_SilenceWithoutPending pins the blocking contract that
// killed the prototype: with no pending tone the reader answers SILENCE
// through the select's default arm — stream Start() returns without a
// pending tone, never a hang.
func TestToneReader_SilenceWithoutPending(t *testing.T) {
	r := newToneReader()

	got, ok := readWithBudget(t, r, silenceChunk)
	if !ok {
		t.Fatal("the reader blocked without a pending tone — Start() would hang (the prototype's SIGQUIT)")
	}
	if len(got) != silenceChunk {
		t.Fatalf("silence read = %d bytes, want %d", len(got), silenceChunk)
	}
	if !bytes.Equal(got, make([]byte, silenceChunk)) {
		t.Error("the silence is not zeroed — int16 silence must be zero pairs")
	}
}

// TestToneReader_ServesPendingTone pins the tone path: a fed tone's bytes
// come out whole and in order (a short read at the tone's end is legal
// io.Reader form), then the reader falls back to silence — no truncation,
// no duplication.
func TestToneReader_ServesPendingTone(t *testing.T) {
	r := newToneReader()
	tone := toneOf(100, -200, 300, -400, 500, -600)
	r.feed(tone)

	first, ok := readWithBudget(t, r, 4)
	if !ok {
		t.Fatal("the reader never served the pending tone")
	}
	if !bytes.Equal(first, tone[:4]) {
		t.Errorf("first read = %v, want the tone's first bytes %v", first, tone[:4])
	}

	second, ok := readWithBudget(t, r, silenceChunk)
	if !ok {
		t.Fatal("the reader never served the tone's remainder")
	}
	if !bytes.Equal(second, tone[4:]) {
		t.Errorf("second read = %v, want the tone's remainder %v", second, tone[4:])
	}

	third, ok := readWithBudget(t, r, silenceChunk)
	if !ok {
		t.Fatal("the reader never fell back to silence after the tone")
	}
	if !bytes.Equal(third, make([]byte, silenceChunk)) {
		t.Error("the post-tone read is not silence")
	}
}

// TestToneReader_NoBusySpin pins the wakeup discipline: one Read call is
// exactly one wakeup — the reader runs no internal loop and spawns no
// goroutine, so the quiet stream's callback rate is the caller's (the
// server-paced drain), never a spin.
func TestToneReader_NoBusySpin(t *testing.T) {
	r := newToneReader()

	for range 3 {
		got, ok := readWithBudget(t, r, silenceChunk)
		if !ok {
			t.Fatal("a quiet read blocked — the reader looped or waited")
		}
		if len(got) != silenceChunk {
			t.Fatalf("quiet read = %d bytes, want exactly one chunk per wakeup", len(got))
		}
	}
}
