package correct_test

import (
	"testing"

	"github.com/Djarvur/goswitch/internal/correct"
)

// TestMatchesSuffix pins the ADR-004 suffix rule: the text before the
// cursor must END with the token. A tail after the token inside the
// surrounding text is a mismatch — the verification waits for the token to
// be last — and text shorter than the token is not a match either.
func TestMatchesSuffix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		text  string
		token string
		want  bool
	}{
		{name: "ends with token", text: "abc " + wordEN, token: wordEN, want: true},
		{name: "tail after token in text", text: "abc " + wordEN + " ", token: wordEN, want: false},
		{name: "text shorter than token", text: "gfb", token: wordEN, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := correct.MatchesSuffix([]rune(tc.text), []rune(tc.token)); got != tc.want {
				t.Errorf("MatchesSuffix(%q, %q) = %v, want %v", tc.text, tc.token, got, tc.want)
			}
		})
	}
}

// TestSelectionRange pins the D-30 selection geometry: the range is the
// half-open interval between cursor and anchor in EITHER order — the client
// reports the pair from its own selection direction, so right-to-left
// (cursor > anchor) and left-to-right (cursor < anchor) describe the SAME
// range — and a selection is active exactly when the positions differ.
func TestSelectionRange(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		cursor       uint32
		anchor       uint32
		wantStart    uint32
		wantEnd      uint32
		wantActive   bool
	}{
		{name: "left to right", cursor: 0, anchor: 6, wantStart: 0, wantEnd: 6, wantActive: true},
		{name: "right to left", cursor: 6, anchor: 0, wantStart: 0, wantEnd: 6, wantActive: true},
		{name: "interior range", cursor: 9, anchor: 3, wantStart: 3, wantEnd: 9, wantActive: true},
		{name: "collapsed", cursor: 5, anchor: 5, wantActive: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			start, end, active := correct.SelectionRange(tc.cursor, tc.anchor)
			if active != tc.wantActive {
				t.Errorf("SelectionRange(%d, %d) active = %v, want %v", tc.cursor, tc.anchor, active, tc.wantActive)
			}
			if !tc.wantActive {
				return // inactive: the range half is undefined by contract
			}
			if start != tc.wantStart || end != tc.wantEnd {
				t.Errorf("SelectionRange(%d, %d) = (%d, %d), want (%d, %d)",
					tc.cursor, tc.anchor, start, end, tc.wantStart, tc.wantEnd)
			}
		})
	}
}

// TestVerifyRangeAt pins the verify-after rule of the selection correction
// (Pitfall 6): MatchesSuffix only checks the prefix-before-cursor, which a
// selection right of the cursor escapes — the range check must find the
// converted text AT the selection position, and only there.
func TestVerifyRangeAt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		text  string
		start uint32
		want  string
		ok    bool
	}{
		{name: "range at start", text: "привет мир", start: 0, want: "привет", ok: true},
		{name: "shifted range", text: "привет мир", start: 1, want: "привет", ok: false},
		{name: "range past the text", text: "привет", start: 2, want: "привет", ok: false},
		{name: "want longer than text", text: "привет", start: 0, want: "привет мир", ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := correct.VerifyRangeAt([]rune(tc.text), tc.start, []rune(tc.want)); got != tc.ok {
				t.Errorf("VerifyRangeAt(%q, %d, %q) = %v, want %v", tc.text, tc.start, tc.want, got, tc.ok)
			}
		})
	}
}
