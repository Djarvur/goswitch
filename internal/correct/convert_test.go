package correct_test

import (
	"testing"

	"github.com/Djarvur/goswitch/internal/correct"
)

// TestConvert_CasePreserved pins CORR-05 in both directions: register is
// preserved per character through the key-position tables. FULL by-position
// conversion (owner directive 2026-09-28 — «изучи раскладку и сделай все
// как надо», superseding the D-14 ride-along AND the 2026-09-27 allow-set):
// EVERY table-mapped rune of the token converts — letters, the bracket row,
// AND the comma/dot/question class — because the tables carry both shift
// levels per physical key, making the mapping a bijection: typing the same
// physical keys in the other mode yields exactly the converted text
// ("hf,jnftn"→"работает", "ghbdtn,"→"приветб", "ghbdtn?"→"привет," — the
// RU comma lives on the '?' key). A non-letter MISSING from the table
// (space, tab) rides along; a LETTER missing from the table fails the
// whole conversion.
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
		// Full by-position conversion (owner directive 2026-09-28): the
		// comma and the dot convert with the word — the live defect
		// «ра,отает» was exactly "hf,jnftn" corrected with the comma left
		// riding.
		{name: "comma converts (supersedes D-14)", in: wordENComma, dir: correct.ENtoRU, want: "приветб", ok: true},
		{name: "dot converts", in: wordEN + ".", dir: correct.ENtoRU, want: "приветю", ok: true},
		{name: "live defect word", in: "hf,jnftn", dir: correct.ENtoRU, want: "работает", ok: true},
		{name: "question becomes the ru comma", in: wordEN + "?", dir: correct.ENtoRU, want: wordRU + ",", ok: true},
		{name: "slash becomes the ru dot", in: wordEN + "/", dir: correct.ENtoRU, want: wordRU + ".", ok: true},
		{name: "hash becomes numero", in: wordEN + "#", dir: correct.ENtoRU, want: wordRU + "№", ok: true},
		{name: "at becomes the ru quote", in: wordEN + "@", dir: correct.ENtoRU, want: wordRU + "\"", ok: true},
		{name: "tilde becomes yo", in: wordEN + "~", dir: correct.ENtoRU, want: wordRU + "Ё", ok: true},
		{name: "bracket converts with the word", in: wordEN + "[", dir: correct.ENtoRU, want: wordRU + "х", ok: true},
		{name: "close bracket converts", in: "]", dir: correct.ENtoRU, want: "ъ", ok: true},
		{name: "semicolon converts", in: ";", dir: correct.ENtoRU, want: "ж", ok: true},
		{name: "apostrophe converts", in: "'", dir: correct.ENtoRU, want: "э", ok: true},
		{name: "backtick converts", in: "`", dir: correct.ENtoRU, want: "ё", ok: true},
		// Ride-along pins: only table-MISSING non-letters ride — space (and
		// tab) sit on no layout key pair.
		{name: "space rides along", in: wordEN + " ", dir: correct.ENtoRU, want: wordRU + " ", ok: true},
		// RU→EN mirrors the bijection: the RU comma returns to the '?' key,
		// the RU dot to the '/' key.
		{name: "ru comma becomes question", in: wordRU + ",", dir: correct.RUtoEN, want: wordEN + "?", ok: true},
		{name: "ru dot becomes slash", in: wordRU + ".", dir: correct.RUtoEN, want: wordEN + "/", ok: true},
		{name: "ru tilde becomes yo-back", in: wordRU + "~", dir: correct.RUtoEN, want: wordEN + "~", ok: true},
		// RU→EN: every non-letter rides (non-letter↔non-letter) — a ';' or
		// '"' in a Russian-typed text is RU-keyboard punctuation, and
		// The bijection's other half: RU-keyboard punctuation returns to
		// its EN key ('"' is the 2-key's '@', ';' the 4-key's '$').
		{name: "ru quote becomes at", in: wordRU + "\"", dir: correct.RUtoEN, want: wordEN + "@", ok: true},
		{name: "ru semicolon becomes dollar", in: wordRU + ";", dir: correct.RUtoEN, want: wordEN + "$", ok: true},
		// '[' sits on no RU key — a genuine table miss rides in RU→EN.
		{name: "en bracket rides in ru text", in: wordRU + "[", dir: correct.RUtoEN, want: wordEN + "[", ok: true},
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
