package correct_test

import (
	"testing"

	"github.com/Djarvur/goswitch/internal/correct"
)

// TestConvert_CasePreserved pins CORR-05 in both directions: register is
// preserved per character through the key-position tables, digits and
// non-letter token runes pass through identically (D-14/D-15 — the comma of
// wordENComma stays a comma, it is not converted to RU 'б'), and a letter
// missing from the table fails the whole conversion.
//
// The allow-set extension of quick plan 260927-vu8 (SPEC §4.2 «включая
// знаки», owner decision 1 revised 2026-09-27): the EN bracket-row symbols
// '[', ']', ';', the apostrophe and '`' — the EN symbols that never serve
// as Russian prose punctuation in the user's intent — convert together
// with the letters ('['→'х' is the live defect shape «несколько
// дополнительны[»), while
// ',' and '.' stay ride-alongs (universal Russian prose punctuation —
// "ghbdtn," intends the separator, D-14) and so does the whole remaining
// non-letter class ('@'-class, digits, space, slash). In the RU→EN
// direction every non-letter rides: a ';' in a Russian-typed text is the RU
// keyboard's own punctuation (Shift+4), not one of the EN symbols.
func TestConvert_CasePreserved(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		dir  correct.Dir
		want string
		ok   bool
	}{
		{name: "lower en-ru", in: wordEN, dir: correct.ENtoRU, want: wordRU, ok: true},
		{name: "upper en-ru", in: "GHBDTN", dir: correct.ENtoRU, want: "ПРИВЕТ", ok: true},
		{name: "capitalized en-ru", in: "Ghbdtn", dir: correct.ENtoRU, want: "Привет", ok: true},
		{name: "lower ru-en", in: wordRU, dir: correct.RUtoEN, want: wordEN, ok: true},
		{name: "upper ru-en", in: "ПРИВЕТ", dir: correct.RUtoEN, want: "GHBDTN", ok: true},
		{name: "capitalized ru-en", in: "Привет", dir: correct.RUtoEN, want: "Ghbdtn", ok: true},
		{name: "digits identical", in: wordENDigits, dir: correct.ENtoRU, want: "привет2026", ok: true},
		{name: "token punctuation identical (D-14)", in: wordENComma, dir: correct.ENtoRU, want: "привет,", ok: true},
		// The allow-set bracket row converts with the word (owner decision 1,
		// 260927-vu8) — '[' is the live defect shape: the word converted, the
		// trailing RU 'х' key stayed.
		{
			name: "bracket converts with the word (live defect)",
			in:   wordEN + "[", dir: correct.ENtoRU, want: wordRU + "х", ok: true,
		},
		{name: "close bracket converts", in: "]", dir: correct.ENtoRU, want: "ъ", ok: true},
		{name: "semicolon converts", in: ";", dir: correct.ENtoRU, want: "ж", ok: true},
		{name: "apostrophe converts", in: "'", dir: correct.ENtoRU, want: "э", ok: true},
		{name: "backtick converts", in: "`", dir: correct.ENtoRU, want: "ё", ok: true},
		// Ride-along pins (the allow-set boundary): universal Russian prose
		// punctuation and the rest of the non-letter class never convert —
		// their table targets ('@'→'"', '#'→'№', '?'→',', '/'→'.') are
		// bypassed exactly like the comma's.
		{name: "dot rides along", in: wordEN + ".", dir: correct.ENtoRU, want: wordRU + ".", ok: true},
		{name: "at rides along", in: wordEN + "@", dir: correct.ENtoRU, want: wordRU + "@", ok: true},
		{
			name: "hash rides along",
			in:   wordEN + "#", dir: correct.ENtoRU, want: wordRU + "#", ok: true,
		},
		{
			name: "question rides along",
			in:   wordEN + "?", dir: correct.ENtoRU, want: wordRU + "?", ok: true,
		},
		{name: "space and slash ride along", in: wordEN + " /", dir: correct.ENtoRU, want: wordRU + " /", ok: true},
		// RU→EN: every non-letter rides (non-letter↔non-letter) — a ';' or
		// '"' in a Russian-typed text is RU-keyboard punctuation, and
		// RUToEN has no '[' entry anyway.
		{
			name: "ru quote rides along (non-letter)",
			in:   wordRU + "\"", dir: correct.RUtoEN, want: wordEN + "\"", ok: true,
		},
		{name: "ru semicolon rides along", in: wordRU + ";", dir: correct.RUtoEN, want: wordEN + ";", ok: true},
		{name: "unmapped letter fails", in: "é", dir: correct.ENtoRU, want: "", ok: false},
		{name: "unknown direction fails", in: wordEN, dir: 0, want: "", ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := correct.Convert([]rune(tc.in), tc.dir)
			if ok != tc.ok {
				t.Fatalf("Convert(%q, %d) ok = %v, want %v", tc.in, tc.dir, ok, tc.ok)
			}
			if string(got) != tc.want {
				t.Errorf("Convert(%q, %d) = %q, want %q", tc.in, tc.dir, string(got), tc.want)
			}
		})
	}
}
