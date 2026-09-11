package correct_test

import (
	"testing"

	"github.com/Djarvur/goswitch/internal/correct"
)

// TestConvert_CasePreserved pins CORR-05 in both directions: register is
// preserved per character through the key-position tables, digits and
// non-letter token runes pass through identically (D-14/D-15 — the comma of
// "ghbdtn," stays a comma, it is not converted to RU 'б'), and a letter
// missing from the table fails the whole conversion.
func TestConvert_CasePreserved(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		dir  correct.Dir
		want string
		ok   bool
	}{
		{name: "lower en-ru", in: "ghbdtn", dir: correct.ENtoRU, want: "привет", ok: true},
		{name: "upper en-ru", in: "GHBDTN", dir: correct.ENtoRU, want: "ПРИВЕТ", ok: true},
		{name: "capitalized en-ru", in: "Ghbdtn", dir: correct.ENtoRU, want: "Привет", ok: true},
		{name: "lower ru-en", in: "привет", dir: correct.RUtoEN, want: "ghbdtn", ok: true},
		{name: "upper ru-en", in: "ПРИВЕТ", dir: correct.RUtoEN, want: "GHBDTN", ok: true},
		{name: "capitalized ru-en", in: "Привет", dir: correct.RUtoEN, want: "Ghbdtn", ok: true},
		{name: "digits identical", in: "ghbdtn2026", dir: correct.ENtoRU, want: "привет2026", ok: true},
		{name: "token punctuation identical (D-14)", in: "ghbdtn,", dir: correct.ENtoRU, want: "привет,", ok: true},
		{name: "unmapped letter fails", in: "é", dir: correct.ENtoRU, want: "", ok: false},
		{name: "unknown direction fails", in: "ghbdtn", dir: 0, want: "", ok: false},
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
