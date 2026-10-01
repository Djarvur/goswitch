package detect_test

import (
	"slices"
	"strings"
	"testing"
	"unicode"

	"github.com/Djarvur/goswitch/internal/correct"
	"github.com/Djarvur/goswitch/internal/detect"
	"github.com/Djarvur/goswitch/layouts"
)

// The test modes — the actor's mode record is a string (D-34), and Check
// accepts exactly the "en"/"ru" spellings (the unknown-mode pin below).
const (
	testModeEN = "en"
	testModeRU = "ru"
)

// The test dictionaries: a hand-picked core per language plus a
// deterministic stride sample of the real baked data for plausibility —
// sorted, the SearchStrings contract. "anagram" deliberately sits in BOTH
// fixtures: the both-hit veto needs a word valid in two layouts at once (a
// random cross-script anagram), which the script-disjoint real data cannot
// provide (its disjointness is pinned by layouts/dict_test.go).
var (
	//nolint:gochecknoglobals // the shared fixture corpus: every test judges the same dictionaries
	fixtureRU = buildFixture([]string{"привет", "мир", "работает", "небо", "anagram"}, layouts.DictRU, 1499, 200)
	//nolint:gochecknoglobals // the shared fixture corpus: every test judges the same dictionaries
	fixtureEN = buildFixture([]string{"hello", "world", "layout", "test", "anagram"}, layouts.DictEN, 997, 200)

	//nolint:gochecknoglobals // the fixtures above, wrapped once for Check
	fixtures = detect.Data{RU: fixtureRU, EN: fixtureEN}
	//nolint:gochecknoglobals // the real baked tables, wrapped once for Check
	trigrams = detect.Trigrams{RU: layouts.TriRU, EN: layouts.TriEN}
)

// buildFixture merges core with a stride sample of baked pure-letter words
// (long enough to clear the default MinWordLen) and sorts the result.
func buildFixture(core, baked []string, stride, want int) []string {
	words := slices.Clone(core)
	for i := 0; i < len(baked) && len(words)-len(core) < want; i += stride {
		w := baked[i]
		tooShort := len([]rune(w)) < 4
		notALetter := strings.IndexFunc(w, func(r rune) bool { return !unicode.IsLetter(r) }) >= 0
		if tooShort || notALetter {
			continue
		}
		words = append(words, w)
	}
	slices.Sort(words)

	return words
}

// TestCheck_DictWrongLayout pins D-52(а): a token out of the current
// dictionary whose remap sits in the other one is a confident wrong-layout
// in both directions.
func TestCheck_DictWrongLayout(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		tok  string
		mode string
		dir  correct.Dir
	}{
		{name: "ghbdtn typed in en", tok: "ghbdtn", mode: testModeEN, dir: correct.ENtoRU},
		{name: "руддщ typed in ru", tok: "руддщ", mode: testModeRU, dir: correct.RUtoEN},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := detect.Check([]rune(tc.tok), tc.mode, fixtures, trigrams, detect.DefaultParams())
			want := detect.Verdict{
				WrongLayout: true,
				Confident:   true,
				Dir:         tc.dir,
				Reason:      detect.ReasonDictWrongLayout,
			}
			if got != want {
				t.Errorf("Check(%q, %q) = %+v, want %+v", tc.tok, tc.mode, got, want)
			}
		})
	}
}

// TestCheck_DictCurHitVetoes pins the veto: a word valid in the current
// layout is NEVER a wrong-layout, whatever the other side would say (D-52:
// every dictionary hit suppresses correction — completeness is safety).
func TestCheck_DictCurHitVetoes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		tok  string
		mode string
	}{
		{name: "hello typed in en", tok: "hello", mode: testModeEN},
		{name: "привет typed in ru", tok: "привет", mode: testModeRU},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := detect.Check([]rune(tc.tok), tc.mode, fixtures, trigrams, detect.DefaultParams())
			want := detect.Verdict{Reason: detect.ReasonDictCurHit}
			if got != want {
				t.Errorf("Check(%q, %q) = %+v, want %+v", tc.tok, tc.mode, got, want)
			}
		})
	}
}

// TestCheck_DictBothHitVetoes pins the second veto: a word valid in BOTH
// layouts (the random-anagram case — "anagram" sits in both fixtures) is
// nobody's wrong-layout.
func TestCheck_DictBothHitVetoes(t *testing.T) {
	t.Parallel()

	got := detect.Check([]rune("anagram"), testModeEN, fixtures, trigrams, detect.DefaultParams())
	want := detect.Verdict{Reason: detect.ReasonDictBothHit}
	if got != want {
		t.Errorf("Check(anagram, en) = %+v, want %+v", got, want)
	}
}

// TestCheck_AbstainShort pins the length gate: an empty, nil or 1-rune
// token abstains always (probe CORR-07/empty) — no branch may answer
// wrong-layout for it.
func TestCheck_AbstainShort(t *testing.T) {
	t.Parallel()

	want := detect.Verdict{Reason: detect.ReasonAbstainShort}
	for _, tok := range []string{"", "г", "h"} {
		got := detect.Check([]rune(tok), testModeRU, fixtures, trigrams, detect.DefaultParams())
		if got != want {
			t.Errorf("Check(%q, ru) = %+v, want %+v", tok, got, want)
		}
	}
	if got := detect.Check(nil, testModeEN, fixtures, trigrams, detect.DefaultParams()); got != want {
		t.Errorf("Check(nil, en) = %+v, want %+v", got, want)
	}
}

// TestCheck_AbstainNoLetters pins the script gate: a token without a
// letter of the mode's script has nothing to judge — digits and
// punctuation carry no script of their own.
func TestCheck_AbstainNoLetters(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		tok  string
		mode string
	}{
		{name: "digits in en", tok: "12345", mode: testModeEN},
		{name: "digits in ru", tok: "12345", mode: testModeRU},
		{name: "punctuation in en", tok: "!!!?", mode: testModeEN},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := detect.Check([]rune(tc.tok), tc.mode, fixtures, trigrams, detect.DefaultParams())
			want := detect.Verdict{Reason: detect.ReasonAbstainNoLetters}
			if got != want {
				t.Errorf("Check(%q, %q) = %+v, want %+v", tc.tok, tc.mode, got, want)
			}
		})
	}
}

// TestCheck_UnknownModeAbstains pins the mode contract: only the
// "en"/"ru" spellings are judgable — anything else has no target side at
// all and falls into the no-letters abstention.
func TestCheck_UnknownModeAbstains(t *testing.T) {
	t.Parallel()

	want := detect.Verdict{Reason: detect.ReasonAbstainNoLetters}
	for _, mode := range []string{"", "xx", "EN", "english"} {
		got := detect.Check([]rune("ghbdtn"), mode, fixtures, trigrams, detect.DefaultParams())
		if got != want {
			t.Errorf("Check(ghbdtn, %q) = %+v, want %+v", mode, got, want)
		}
	}
}

// TestCheck_AbstainConvertFail pins the mixed-token refusal: one foreign
// letter fails the whole correct.Convert, and the detector abstains —
// never a partial rewrite (CORR-06 via the D-16 refusal reuse).
func TestCheck_AbstainConvertFail(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		tok  string
		mode string
	}{
		{name: "latin head cyrillic tail in en", tok: "gфпривет", mode: testModeEN},
		{name: "cyrillic head latin tail in ru", tok: "приветg", mode: testModeRU},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := detect.Check([]rune(tc.tok), tc.mode, fixtures, trigrams, detect.DefaultParams())
			want := detect.Verdict{Reason: detect.ReasonAbstainConvertFail}
			if got != want {
				t.Errorf("Check(%q, %q) = %+v, want %+v", tc.tok, tc.mode, got, want)
			}
		})
	}
}

// TestCheck_DirMatchesCorrect pins the direction vocabulary: a confident
// verdict's Dir is exactly the correct-package direction of the mode —
// ENtoRU for en, RUtoEN for ru (the remap is reused, never re-derived).
func TestCheck_DirMatchesCorrect(t *testing.T) {
	t.Parallel()

	got := detect.Check([]rune("ghbdtn"), testModeEN, fixtures, trigrams, detect.DefaultParams())
	if got.Dir != correct.ENtoRU {
		t.Errorf("Check(ghbdtn, en).Dir = %v, want %v", got.Dir, correct.ENtoRU)
	}
	got = detect.Check([]rune("руддщ"), testModeRU, fixtures, trigrams, detect.DefaultParams())
	if got.Dir != correct.RUtoEN {
		t.Errorf("Check(руддщ, ru).Dir = %v, want %v", got.Dir, correct.RUtoEN)
	}
}
