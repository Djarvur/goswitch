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
// sorted, the SearchStrings contract. The pair "анаграмма"/"fyfuhfvvf"
// (the RUToEN remap of the word) deliberately sits in BOTH fixtures: the
// both-hit veto needs a token whose dictionary hit exists on both sides at
// once (a random cross-script anagram), which the script-disjoint real
// data cannot provide (its disjointness is pinned by layouts/dict_test.go).
var (
	//nolint:gochecknoglobals // the shared fixture corpus: every test judges the same dictionaries
	fixtureRU = buildFixture([]string{"привет", "мир", "работает", "небо", "анаграмма"}, layouts.DictRU, 1499, 200)
	//nolint:gochecknoglobals // the shared fixture corpus: every test judges the same dictionaries
	fixtureEN = buildFixture([]string{"hello", "world", "layout", "test", "fyfuhfvvf"}, layouts.DictEN, 997, 200)

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

// TestCheck_DictBothHitVetoes pins the second veto: a token whose
// dictionary hit exists on BOTH sides (the random-anagram case — the
// "анаграмма"/"fyfuhfvvf" remap pair sits in both fixtures) is nobody's
// wrong-layout.
func TestCheck_DictBothHitVetoes(t *testing.T) {
	t.Parallel()

	got := detect.Check([]rune("анаграмма"), testModeRU, fixtures, trigrams, detect.DefaultParams())
	want := detect.Verdict{Reason: detect.ReasonDictBothHit}
	if got != want {
		t.Errorf("Check(анаграмма, ru) = %+v, want %+v", got, want)
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

// The synthetic trigram tables of the fallback corpus: built around the
// "abvgd"↔"фимпв" remap pair — the latin side scores garbage (no table
// entries), the cyrillic side scores a strong uniform profile. Independent
// of the baked layouts data by design (D-07: the behavior corpus pins the
// scoring arithmetic, not the live tables).
var (
	//nolint:gochecknoglobals // the shared fallback corpus tables
	triStrongRU = map[string]float64{
		"фим": -1.0,
		"имп": -1.0,
		"мпв": -1.0,
		// the junction windows of the doubled word "фимпввфимпв" — the
		// length-normalization pin doubles through the same letter, and the
		// doubled profile must stay as strong as the single one
		"пвв": -1.0,
		"ввф": -1.0,
		"вфи": -1.0,
	}
	//nolint:gochecknoglobals // the mirror-side table of the ru-mode case
	triStrongEN = map[string]float64{
		"abd": -1.0,
		"bdg": -1.0,
		"dgd": -1.0,
	}
)

// trigramParams returns thresholds on the log10 scale that let the strong
// synthetic profile through: margin 1.0, floor -3.0.
func trigramParams() detect.Params {
	return detect.Params{MinWordLen: 4, TrigramMargin: 1.0, TrigramFloor: -3.0}
}

// TestTrigram_WrongLayout pins D-52(б): a token both dictionaries miss
// whose other-language remap profile beats the current-language profile by
// more than the margin and stays above the floor is a confident
// wrong-layout — in both directions.
func TestTrigram_WrongLayout(t *testing.T) {
	t.Parallel()

	tri := detect.Trigrams{RU: triStrongRU, EN: triStrongEN}
	cases := []struct {
		name string
		tok  string
		mode string
		dir  correct.Dir
	}{
		{name: "garbage latin, strong ru remap", tok: "abvgd", mode: testModeEN, dir: correct.ENtoRU},
		{name: "garbage cyrillic, strong en remap", tok: "фивпв", mode: testModeRU, dir: correct.RUtoEN},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := detect.Check([]rune(tc.tok), tc.mode, fixtures, tri, trigramParams())
			want := detect.Verdict{
				WrongLayout: true,
				Confident:   true,
				Dir:         tc.dir,
				Reason:      detect.ReasonTrigramWrongLayout,
			}
			if got != want {
				t.Errorf("Check(%q, %q) = %+v, want %+v", tc.tok, tc.mode, got, want)
			}
		})
	}
}

// TestTrigram_UnsureBelowMargin pins the relative gate: when the other
// language wins by less than the margin the fallback stays unsure.
func TestTrigram_UnsureBelowMargin(t *testing.T) {
	t.Parallel()

	tri := detect.Trigrams{RU: triStrongRU, EN: triStrongEN}
	p := detect.Params{MinWordLen: 4, TrigramMargin: 5.5, TrigramFloor: -3.0}

	got := detect.Check([]rune("abvgd"), testModeEN, fixtures, tri, p)
	want := detect.Verdict{Reason: detect.ReasonTrigramUnsure}
	if got != want {
		t.Errorf("Check(abvgd, en) = %+v, want %+v", got, want)
	}
}

// TestTrigram_UnsureBelowFloor pins the absolute gate: a plausibility
// below the floor is insufficient even at a large margin win — the
// fallback never bets on a word it cannot vouch for.
func TestTrigram_UnsureBelowFloor(t *testing.T) {
	t.Parallel()

	tri := detect.Trigrams{RU: triStrongRU, EN: triStrongEN}
	p := detect.Params{MinWordLen: 4, TrigramMargin: 1.0, TrigramFloor: 0.0}

	got := detect.Check([]rune("abvgd"), testModeEN, fixtures, tri, p)
	want := detect.Verdict{Reason: detect.ReasonTrigramUnsure}
	if got != want {
		t.Errorf("Check(abvgd, en) = %+v, want %+v", got, want)
	}
}

// TestTrigram_NeutralTrigramsUnknown pins the unseen-window penalty: a
// word whose trigrams are absent from BOTH tables scores the neutral
// penalty on both sides (the -6.0 log10 floor pinned by the neutralPenalty
// comment), the race is a draw, and a draw is never confident.
func TestTrigram_NeutralTrigramsUnknown(t *testing.T) {
	t.Parallel()

	tri := detect.Trigrams{RU: map[string]float64{}, EN: map[string]float64{}}
	p := trigramParams()

	got := detect.Check([]rune("zzzzz"), testModeEN, fixtures, tri, p)
	want := detect.Verdict{Reason: detect.ReasonTrigramUnsure}
	if got != want {
		t.Errorf("Check(zzzzz, en) = %+v, want %+v", got, want)
	}
}

// TestTrigram_NormalizedByLength pins the length normalization: doubling
// the word through its last letter (the same trigram classes, twice as
// many windows plus junctions) never changes the verdict.
func TestTrigram_NormalizedByLength(t *testing.T) {
	t.Parallel()

	tri := detect.Trigrams{RU: triStrongRU, EN: triStrongEN}

	got := detect.Check([]rune("abvgddabvgd"), testModeEN, fixtures, tri, trigramParams())
	want := detect.Verdict{
		WrongLayout: true,
		Confident:   true,
		Dir:         correct.ENtoRU,
		Reason:      detect.ReasonTrigramWrongLayout,
	}
	if got != want {
		t.Errorf("Check(abvgddabvgd, en) = %+v, want %+v", got, want)
	}
}

// TestTrigram_DictHitSkipsScoring pins the fallback's reach: a
// current-dictionary hit is decided by the dict path ALONE — the tables
// here would flip the verdict if the scoring were ever consulted, and the
// expected answer is still the veto.
func TestTrigram_DictHitSkipsScoring(t *testing.T) {
	t.Parallel()

	// hello → руддщ: a profile that would scream ru if the scoring ran.
	tri := detect.Trigrams{
		RU: map[string]float64{"руд": -1.0, "удд": -1.0, "ддщ": -1.0},
		EN: map[string]float64{},
	}

	got := detect.Check([]rune("hello"), testModeEN, fixtures, tri, trigramParams())
	want := detect.Verdict{Reason: detect.ReasonDictCurHit}
	if got != want {
		t.Errorf("Check(hello, en) = %+v, want %+v", got, want)
	}
}
