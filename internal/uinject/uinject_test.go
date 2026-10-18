//nolint:testpackage // drives the unexported resolve/clock seams — the sound corpus in-package precedent
package uinject

// The uinject pure-logic corpus (TDD): the keycode table, the 24-byte
// input_event encoders, the pre-typing resolve gate and the closed-injector
// guard — the ONLY tests of the package. The device is never opened here:
// Open belongs to the driver's battery start, and no test ever injects
// (or sleeps — the clock is a seam, never exercised on a real timer).

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
	"testing"
	"time"
)

// redTime is the fixed stamp the encoder vectors use: sec=1, usec=500 —
// every timeval byte hand-computable.
func redTime() time.Time {
	return time.Unix(1, 500000)
}

// TestKeycodeDigits pins the digit row: KEY_1..KEY_9 = 2..10 and KEY_0
// capping the row at 11 (input-event-codes.h lines 77-86) — the marker's
// vocabulary (the focus proof types "7" = keycode 8).
func TestKeycodeDigits(t *testing.T) {
	for _, tc := range []struct {
		name string
		want uint16
	}{
		{"1", 2}, {"2", 3}, {"3", 4}, {"4", 5}, {"5", 6},
		{"6", 7}, {"7", 8}, {"8", 9}, {"9", 10}, {"0", 11},
	} {
		code, ok := Keycode(tc.name)
		if !ok {
			t.Fatalf("Keycode(%q) missed the table, want %d", tc.name, tc.want)
		}
		if code != tc.want {
			t.Fatalf("Keycode(%q) = %d, want %d", tc.name, code, tc.want)
		}
	}
}

// TestKeycodeCorpusLetters pins the corpus letters: the positional QWERTY
// rows of /usr/include/linux/input-event-codes.h (T..P = 20..25, G..L =
// 34..38, B..M = 48..50), NOT alphabetical — and NOT the work order's
// g=35/h=36, which are KEY_H and KEY_J: the header and the transcription
// in internal/session/probe.go agree on KEY_G=34, KEY_H=35, and a letter
// off by one row position types a different character entirely (the corpus
// would stop being ghbdtn).
func TestKeycodeCorpusLetters(t *testing.T) {
	for _, tc := range []struct {
		name string
		want uint16
	}{
		{"g", 34}, // KEY_G
		{"h", 35}, // KEY_H
		{"b", 48}, // KEY_B
		{"d", 32}, // KEY_D
		{"t", 20}, // KEY_T
		{"n", 49}, // KEY_N
		{"u", 22}, // KEY_U (the Ctrl-U clear's letter half)
	} {
		code, ok := Keycode(tc.name)
		if !ok {
			t.Fatalf("Keycode(%q) missed the table, want %d", tc.name, tc.want)
		}
		if code != tc.want {
			t.Fatalf("Keycode(%q) = %d, want %d (linux/input-event-codes.h)", tc.name, code, tc.want)
		}
	}
}

// TestKeycodeNamed pins the named keys: the chord modifiers and the
// reserved vocabulary — LeftCtrl (the Ctrl-U clear), ShiftRight (the
// double-tap correction gesture), BackSpace and Enter.
func TestKeycodeNamed(t *testing.T) {
	for _, tc := range []struct {
		name string
		want uint16
	}{
		{"LeftCtrl", 29},   // KEY_LEFTCTRL
		{"ShiftRight", 54}, // KEY_RIGHTSHIFT
		{"BackSpace", 14},  // KEY_BACKSPACE
		{"Enter", 28},      // KEY_ENTER
	} {
		code, ok := Keycode(tc.name)
		if !ok {
			t.Fatalf("Keycode(%q) missed the table, want %d", tc.name, tc.want)
		}
		if code != tc.want {
			t.Fatalf("Keycode(%q) = %d, want %d (linux/input-event-codes.h)", tc.name, code, tc.want)
		}
	}
}

// TestKeycodeMiss pins the fail-closed miss: a name outside the table
// (ydotool spellings included) reads code 0 and ok=false — never a guessed
// code. The table is case-sensitive by design: "G" is the shifted GLYPH,
// the table carries the KEY.
func TestKeycodeMiss(t *testing.T) {
	for _, name := range []string{"", "Shift_R", "ctrl+u", "G", "П", "привет", "space"} {
		code, ok := Keycode(name)
		if ok {
			t.Fatalf("Keycode(%q) resolved to %d, want a loud miss", name, code)
		}
		if code != 0 {
			t.Fatalf("Keycode(%q) miss carries code %d, want 0", name, code)
		}
	}
}

// TestEncodePressBytes pins the 24-byte input_event layout byte-for-byte:
// timeval (sec int64, usec int64, little-endian) then type, code and the
// press value — the kernel ABI verified against the installed headers
// (sizeof(struct input_event) = 24, timeval = 16).
func TestEncodePressBytes(t *testing.T) {
	ev := EncodePress(34, redTime()) // KEY_G press at sec=1, usec=500
	want := []byte{
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // tv_sec = 1
		0xf4, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // tv_usec = 500
		0x01, 0x00, // type = EV_KEY
		0x22, 0x00, // code = 34 (KEY_G)
		0x01, 0x00, 0x00, 0x00, // value = 1 (press)
	}
	if !bytes.Equal(ev[:], want) {
		t.Fatalf("EncodePress bytes = %x, want %x", ev[:], want)
	}

	// The independent field-level readback (literal offsets, not the
	// package's own constants).
	if sec := int64(binary.LittleEndian.Uint64(ev[0:8])); sec != 1 {
		t.Fatalf("tv_sec = %d, want 1", sec)
	}
	if usec := int64(binary.LittleEndian.Uint64(ev[8:16])); usec != 500 {
		t.Fatalf("tv_usec = %d, want 500", usec)
	}
	if typ := binary.LittleEndian.Uint16(ev[16:18]); typ != 1 {
		t.Fatalf("type = %d, want 1 (EV_KEY)", typ)
	}
	if code := binary.LittleEndian.Uint16(ev[18:20]); code != 34 {
		t.Fatalf("code = %d, want 34 (KEY_G)", code)
	}
	if val := int32(binary.LittleEndian.Uint32(ev[20:24])); val != 1 {
		t.Fatalf("value = %d, want 1 (press)", val)
	}
}

// TestEncodeReleaseBytes pins the release form: the same layout, value 0.
func TestEncodeReleaseBytes(t *testing.T) {
	ev := EncodeRelease(11, redTime()) // KEY_0 release
	want := []byte{
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // tv_sec = 1
		0xf4, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // tv_usec = 500
		0x01, 0x00, // type = EV_KEY
		0x0b, 0x00, // code = 11 (KEY_0)
		0x00, 0x00, 0x00, 0x00, // value = 0 (release)
	}
	if !bytes.Equal(ev[:], want) {
		t.Fatalf("EncodeRelease bytes = %x, want %x", ev[:], want)
	}
}

// TestEncodeSynBytes pins the separator: type EV_SYN (0), code 0 (the
// SYN_REPORT report), value 0 — the flush that hands the pending events
// to the kernel's readers.
func TestEncodeSynBytes(t *testing.T) {
	ev := EncodeSyn(redTime())
	want := []byte{
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // tv_sec = 1
		0xf4, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // tv_usec = 500
		0x00, 0x00, // the SYN type
		0x00, 0x00, // the SYN_REPORT report
		0x00, 0x00, 0x00, 0x00, // the zero value
	}
	if !bytes.Equal(ev[:], want) {
		t.Fatalf("EncodeSyn bytes = %x, want %x", ev[:], want)
	}
}

// TestResolveCorpus pins the pre-typing gate on the driver's whole typing
// vocabulary: the marker digit and the correction corpus resolve cleanly,
// an empty text vacuously.
func TestResolveCorpus(t *testing.T) {
	for _, text := range []string{"", "7", "ghbdtn"} {
		if err := resolve(text); err != nil {
			t.Fatalf("resolve(%q): %v, want nil", text, err)
		}
	}
}

// TestResolveMiss pins the loud miss: the first rune outside the table
// stops the typing BEFORE any key leaves, wrapping ErrKeyUnknown and
// naming the rune (code point and all) for the abort row.
func TestResolveMiss(t *testing.T) {
	err := resolve("ghпр")
	if err == nil {
		t.Fatal("resolve of a non-table rune returned nil, want a loud error")
	}
	if !errors.Is(err, ErrKeyUnknown) {
		t.Fatalf("resolve miss %v does not wrap ErrKeyUnknown", err)
	}
	if !strings.Contains(err.Error(), "п") || !strings.Contains(err.Error(), "U+043F") {
		t.Fatalf("resolve miss %v does not name the rune and its code point", err)
	}
}

// TestCloseDouble pins the lifecycle guard: a closed (or never-opened)
// injector refuses Close loudly instead of issuing ioctls on a stale fd —
// the sentinel fd is the zero-device constructor's state, so the guard is
// pure logic the corpus can pin without opening anything.
func TestCloseDouble(t *testing.T) {
	inj := &Injector{fd: -1, sleep: realSleep}
	if err := inj.Close(); err == nil {
		t.Fatal("Close on a closed injector returned nil, want a loud error")
	}
}
