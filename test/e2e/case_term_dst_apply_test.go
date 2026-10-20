package main

// The round-6 dst-apply case's unit corpus (the 10-13 scope continuation,
// strict TDD): the raw-delta inspection is the experiment's ONLY oracle —
// whether the deleted characters actually left the terminal's input line —
// so the byte arithmetic and the verdict classifiers are pinned BEFORE any
// stand run. Every fixture here is the synthetic corpus (ghbdtn → привет,
// Pitfall 6); no daemon, no window, no bus — pure functions.

import (
	"testing"
	"time"
)

// TestDstFocusInAge pins the warm gate's time anchor (the bisection's
// trigger boundary is TIME from the context's fresh focus_in — crash at
// +0.5 s, survive at +1.5 s): the helper must read the LAST focus_in
// record, skip unparseable lines (an overestimated age is the UNSAFE
// direction) and refuse when no parseable record exists.
func TestDstFocusInAge(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	stamp := func(sec int, msg string) string {
		return `{"time":"` + now.Add(-time.Duration(sec)*time.Second).
			Format(time.RFC3339Nano) + `","level":"INFO","msg":"` + msg + `"}`
	}
	for _, tt := range []struct {
		name    string
		lines   []string
		wantAge time.Duration
		wantOK  bool
	}{
		{
			name:    "age of the last focus_in",
			lines:   []string{stamp(10, "focus_in"), stamp(9, "key")},
			wantAge: 10 * time.Second,
			wantOK:  true,
		},
		{
			name:    "multiple records the last wins",
			lines:   []string{stamp(30, "focus_in"), stamp(4, "focus_in")},
			wantAge: 4 * time.Second,
			wantOK:  true,
		},
		{
			name:    "unparseable last skipped to the parseable one",
			lines:   []string{stamp(20, "focus_in"), "garbage line", `{"time":"nope","msg":"focus_in"}`},
			wantAge: 20 * time.Second,
			wantOK:  true,
		},
		{
			name:   "no focus_in at all refuses",
			lines:  []string{stamp(5, "key")},
			wantOK: false,
		},
		{
			name:   "empty journal refuses",
			lines:  nil,
			wantOK: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			age, ok := dstFocusInAge(tt.lines, now)
			if ok != tt.wantOK || (ok && age != tt.wantAge) {
				t.Fatalf("dstFocusInAge = (%v, %v), want (%v, %v)", age, ok, tt.wantAge, tt.wantOK)
			}
		})
	}
}

// TestInspectDstDelta pins the token-anchored byte arithmetic of the apply
// cells: the erase classes, the last-occurrence anchor (the arming
// residue's erases never count) and the delivery control.
func TestInspectDstDelta(t *testing.T) {
	t.Parallel()
	token := []byte(corpusLatin) // ghbdtn — 6 ASCII bytes

	for _, tt := range []struct {
		name  string
		delta []byte
		want  dstInspect
	}{
		{
			// The APPLIED form: the client synthesized one erase byte per
			// deleted rune after the typed token.
			name:  "applied exact six DEL after token",
			delta: append(append([]byte{}, token...), 0x7f, 0x7f, 0x7f, 0x7f, 0x7f, 0x7f),
			want:  dstInspect{tokenSeen: true, delAfter: 6},
		},
		{
			name:  "ignored token only no erase bytes",
			delta: append([]byte{}, token...),
			want:  dstInspect{tokenSeen: true},
		},
		{
			name:  "partial three of six",
			delta: append(append([]byte{}, token...), 0x7f, 0x7f, 0x7f),
			want:  dstInspect{tokenSeen: true, delAfter: 3},
		},
		{
			// The arming-probe discipline: erases BEFORE the token (the
			// self-erasing pair's residue in the raw capture) never count.
			name:  "erases before token not counted",
			delta: append(append([]byte{0x78, 0x7f, 0x7f}, token...), 0x7f),
			want:  dstInspect{tokenSeen: true, delAfter: 1},
		},
		{
			name:  "no token in delta at all",
			delta: []byte{0x7f, 0x7f, 0x7f},
			want:  dstInspect{},
		},
		{
			name:  "empty delta",
			delta: nil,
			want:  dstInspect{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := inspectDstDelta(tt.delta); got != tt.want {
				t.Fatalf("inspectDstDelta(% x) = %+v, want %+v", tt.delta, got, tt.want)
			}
		})
	}
}

// TestInspectDstDeltaWordOrder pins the word-bearing half of the oracle:
// the choreography's order predicate (every erase strictly before the
// word's first byte) and the BS-byte erase class.
func TestInspectDstDeltaWordOrder(t *testing.T) {
	t.Parallel()
	token := []byte(corpusLatin) // ghbdtn
	word := []byte(drainWordRU)  // привет — 12 UTF-8 bytes

	for _, tt := range []struct {
		name  string
		delta []byte
		want  dstInspect
	}{
		{
			// The full choreography's green form: erases first, then the
			// replacement word — the final line net is the word.
			name: "green erases then word",
			delta: append(append(append([]byte{}, token...),
				0x7f, 0x7f, 0x7f, 0x7f, 0x7f, 0x7f), word...),
			want: dstInspect{tokenSeen: true, delAfter: 6, wordSeen: true, orderOK: true},
		},
		{
			// A client may synthesize the BS byte (0x08) instead of DEL —
			// the erase class counts both.
			name: "BS byte form",
			delta: append(append(append([]byte{}, token...),
				0x08, 0x08, 0x08, 0x08, 0x08, 0x08), word...),
			want: dstInspect{tokenSeen: true, bsAfter: 6, wordSeen: true, orderOK: true},
		},
		{
			// The late-deletion defect shape: an erase byte AFTER the word
			// started — the order predicate must refuse it.
			name: "order broken erase after word start",
			delta: append(append(append([]byte{}, token...), word...),
				0x7f, 0x7f, 0x7f),
			want: dstInspect{tokenSeen: true, delAfter: 3, wordSeen: true},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := inspectDstDelta(tt.delta); got != tt.want {
				t.Fatalf("inspectDstDelta(% x) = %+v, want %+v", tt.delta, got, tt.want)
			}
		})
	}
}

func TestDstApplyVerdictClass(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		ins  dstInspect
		want string
	}{
		{"no token delivery control fails", dstInspect{}, dstVerdictNoToken},
		{"full count", dstInspect{tokenSeen: true, delAfter: 6}, dstVerdictApplied},
		{"full count mixed bytes", dstInspect{tokenSeen: true, delAfter: 5, bsAfter: 1}, dstVerdictApplied},
		{"zero erases", dstInspect{tokenSeen: true}, dstVerdictIgnored},
		{"under count", dstInspect{tokenSeen: true, bsAfter: 4}, dstVerdictPartial},
		{"over count", dstInspect{tokenSeen: true, delAfter: 7}, dstVerdictPartial},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := dstApplyVerdictClass(tt.ins); got != tt.want {
				t.Fatalf("dstApplyVerdictClass(%+v) = %q, want %q", tt.ins, got, tt.want)
			}
		})
	}
}

func TestDstChoreoVerdictClass(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		ins  dstInspect
		want string
	}{
		{"no token", dstInspect{}, dstVerdictNoToken},
		{"no word commit never landed", dstInspect{tokenSeen: true, delAfter: 6}, dstVerdictNoWord},
		{"ignored deletion concat form", dstInspect{tokenSeen: true, wordSeen: true}, dstChoreoConcat},
		{"partial deletion",
			dstInspect{tokenSeen: true, delAfter: 2, wordSeen: true, orderOK: true}, dstVerdictPartial},
		{"order broken late erases",
			dstInspect{tokenSeen: true, delAfter: 6, wordSeen: true}, dstChoreoOrder},
		{"pass full construction",
			dstInspect{tokenSeen: true, delAfter: 6, wordSeen: true, orderOK: true}, dstChoreoPass},
		{"pass mixed erase bytes",
			dstInspect{tokenSeen: true, delAfter: 5, bsAfter: 1, wordSeen: true, orderOK: true},
			dstChoreoPass},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := dstChoreoVerdictClass(tt.ins); got != tt.want {
				t.Fatalf("dstChoreoVerdictClass(%+v) = %q, want %q", tt.ins, got, tt.want)
			}
		})
	}
}
