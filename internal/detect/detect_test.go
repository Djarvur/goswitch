package detect_test

import (
	"reflect"
	"slices"
	"sort"
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

// The real baked data of plan 06-02: the golden corpus judges the detector
// against the dictionaries and trigram tables it will meet in production.
var (
	//nolint:gochecknoglobals // the production-shaped inputs, wrapped once
	realData = detect.Data{RU: layouts.DictRU, EN: layouts.DictEN}
	//nolint:gochecknoglobals // the production-shaped inputs, wrapped once
	realTrigrams = detect.Trigrams{RU: layouts.TriRU, EN: layouts.TriEN}
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

// ---- The golden corpus on the real baked data (Discretion D-52: the
// ---- unit golden corpus is MANDATORY — criterion 1 of the phase).

// A corpus entry: the token, the layout it pretends to be typed in, and
// the class it belongs to.
type corpusCase struct {
	name string
	tok  string
	mode string
}

// hitsDict mirrors the detector's own two-probe lookup (exact spelling,
// then the normalized query) — the corpus side needs the same definition
// of a dictionary hit its expectations are built on.
func hitsDict(dict []string, w string) bool {
	probe := func(s string) bool {
		i := sort.SearchStrings(dict, s)

		return i < len(dict) && dict[i] == s
	}
	if probe(w) {
		return true
	}
	norm := strings.ReplaceAll(strings.ToLower(w), "ё", "е")

	return norm != w && probe(norm)
}

// corpusWords samples at most want words, every stride-th dictionary
// entry, keeping only pure-letter words long enough to clear the default
// MinWordLen whose lowercase spelling is itself an entry (the register
// corpus relies on the normalized probe reaching it). Deterministic — no
// randomness anywhere.
func corpusWords(dict []string, stride, want int) []string {
	words := make([]string, 0, want)
	for i := 0; i < len(dict) && len(words) < want; i += stride {
		w := dict[i]
		tooShort := len([]rune(w)) < 4
		notALetter := strings.IndexFunc(w, func(r rune) bool { return !unicode.IsLetter(r) }) >= 0
		if tooShort || notALetter || !hitsDict(dict, strings.ToLower(w)) {
			continue
		}
		words = append(words, strings.ToLower(w))
	}

	return words
}

// registers returns the three register spellings of a lowercase word.
func registers(w string) []string {
	r := []rune(w)

	return []string{w, string(unicode.ToUpper(r[0])) + string(r[1:]), strings.ToUpper(w)}
}

// wrongLayoutCases builds the D-52(а) corpus: real dictionary words of
// BOTH languages, each in all three registers, tokenized as the remap a
// user types when the wrong layout is active. A token that accidentally
// lands in the current dictionary is skipped — that case belongs to the
// veto, not to this class.
func wrongLayoutCases() []corpusCase {
	cases := make([]corpusCase, 0, 200)
	cases = append(cases,
		corpusCase{name: "flagship ghbdtn", tok: "ghbdtn", mode: testModeEN},
		corpusCase{name: "flagship Ghbdtn", tok: "Ghbdtn", mode: testModeEN},
		corpusCase{name: "flagship GHBDTN", tok: "GHBDTN", mode: testModeEN},
		corpusCase{name: "flagship руддщ", tok: "руддщ", mode: testModeRU},
		corpusCase{name: "flagship Руддщ", tok: "Руддщ", mode: testModeRU},
		corpusCase{name: "flagship РУДДЩ", tok: "РУДДЩ", mode: testModeRU},
		// The plan's Pitfall-2 example: the remap of Vfrcbv is the
		// dictionary word максим — by the D-52 definition this is NOT an
		// OOV case but a dictionary-path wrong-layout, exactly like ghbdtn.
		corpusCase{name: "name vfrcbv", tok: "vfrcbv", mode: testModeEN},
		corpusCase{name: "name Vfrcbv", tok: "Vfrcbv", mode: testModeEN},
		corpusCase{name: "name VFRCBV", tok: "VFRCBV", mode: testModeEN},
		corpusCase{name: "keyboard row", tok: "йцукен", mode: testModeRU},
	)
	for _, w := range corpusWords(layouts.DictRU, 3000, 30) {
		for _, form := range registers(w) {
			tok, ok := correct.Convert([]rune(form), correct.RUtoEN)
			if ok && !hitsDict(layouts.DictEN, string(tok)) {
				cases = append(cases, corpusCase{name: "ru " + form, tok: string(tok), mode: testModeEN})
			}
		}
	}
	for _, w := range corpusWords(layouts.DictEN, 2000, 30) {
		for _, form := range registers(w) {
			tok, ok := correct.Convert([]rune(form), correct.ENtoRU)
			if ok && !hitsDict(layouts.DictRU, string(tok)) {
				cases = append(cases, corpusCase{name: "en " + form, tok: string(tok), mode: testModeRU})
			}
		}
	}

	return cases
}

// legitCases builds the veto corpus: raw dictionary entries of BOTH
// languages (the canonical spelling a user types correctly), checked on
// their own mode.
func legitCases() []corpusCase {
	cases := make([]corpusCase, 0, 500)
	cases = append(cases,
		corpusCase{name: "caps entry", tok: "АЗС", mode: testModeRU},
		corpusCase{name: "proper noun", tok: "Москва", mode: testModeRU},
		corpusCase{name: "proper noun 2", tok: "Россия", mode: testModeRU},
		corpusCase{name: "proper noun 3", tok: "Александр", mode: testModeRU},
		corpusCase{name: "proper noun 4", tok: "Питер", mode: testModeRU},
	)
	for _, w := range corpusWords(layouts.DictRU, 600, 220) {
		cases = append(cases, corpusCase{name: "ru " + w, tok: w, mode: testModeRU})
	}
	for _, w := range corpusWords(layouts.DictEN, 350, 220) {
		cases = append(cases, corpusCase{name: "en " + w, tok: w, mode: testModeEN})
	}

	return cases
}

// oovCases builds the Pitfall-2 class: words OUT OF VOCABULARY on both
// sides — the token misses its own dictionary (exact and normalized) and
// its remap misses the other one. Proper names, nicks, abbreviations,
// keyboard garbage; both directions.
func oovCases() []corpusCase {
	cases := []corpusCase{
		{name: "moscow in latin", tok: "vjcrdf", mode: testModeEN},
		{name: "piter in latin", tok: "gbtnth", mode: testModeEN},
		{name: "rossiya in latin", tok: "hjccbz", mode: testModeEN},
		{name: "alexander in latin", tok: "fktrcfylh", mode: testModeEN},
		{name: "name in latin 1", tok: "ghbytn", mode: testModeEN},
		{name: "nick 1", tok: "pjaka", mode: testModeEN},
		{name: "nick 2", tok: "rjvgk", mode: testModeEN},
		{name: "nick 3", tok: "cjkysq", mode: testModeEN},
		{name: "nick 4", tok: "zcsolx", mode: testModeEN},
		{name: "nick 5", tok: "vbkbz", mode: testModeEN},
		{name: "nick 6", tok: "cyjdvb", mode: testModeEN},
		{name: "tech term", tok: "kubernetes", mode: testModeEN},
		{name: "tech term 2", tok: "postgresql", mode: testModeEN},
		{name: "tech term 3", tok: "javascript", mode: testModeEN},
		{name: "tech term 4", tok: "bluetooth", mode: testModeEN},
		{name: "abbrev", tok: "wifi", mode: testModeEN},
		{name: "keyboard walk", tok: "trololo", mode: testModeEN},
		{name: "keyboard walk 2", tok: "zzzqqq", mode: testModeEN},
		{name: "moscow", tok: "москва", mode: testModeRU},
		{name: "piter", tok: "питер", mode: testModeRU},
		{name: "name 1", tok: "зулихан", mode: testModeRU},
		{name: "name 2", tok: "айгуль", mode: testModeRU},
		{name: "surname", tok: "крутько", mode: testModeRU},
		{name: "nick 7", tok: "жужа", mode: testModeRU},
		{name: "nick 8", tok: "пупкин", mode: testModeRU},
		{name: "keyboard walk 3", tok: "ывапрол", mode: testModeRU},
		{name: "yo nick", tok: "ёшкин", mode: testModeRU},
		{name: "food", tok: "эчпочмак", mode: testModeRU},
		{name: "slang", tok: "хайп", mode: testModeRU},
		{name: "slang 2", tok: "мемас", mode: testModeRU},
		{name: "slang 3", tok: "ойбога", mode: testModeRU},
		{name: "name 3", tok: "кобейн", mode: testModeRU},
		{name: "keyboard walk 4", tok: "ждлорп", mode: testModeRU},
		{name: "keyboard walk 5", tok: "фывапролдж", mode: testModeRU},
	}

	return cases
}

// mixedCases builds the CORR-06 class: one foreign letter fails the whole
// conversion — the detector abstains, never a partial rewrite.
func mixedCases() []corpusCase {
	return []corpusCase{
		{name: "latin head in en", tok: "gфпривет", mode: testModeEN},
		{name: "latin tail in ru", tok: "приветg", mode: testModeRU},
		{name: "cyrillic tail in en", tok: "helloмир", mode: testModeEN},
		{name: "cyrillic head in ru", tok: "мирhello", mode: testModeRU},
	}
}

// The yo words of the corpus, named once (Pitfall 3 class).
const (
	yoElka   = "ёлка"
	yoEshe   = "ещё"
	yoYozhik = "ёжик"
)

// yoCases builds the Pitfall-3 class: yo-spellings of folded dictionary
// words — the ru-side tokens must stay cur-hits (the normalized probe
// finds the е-entry), and the same words typed in the en layout must be
// confidently corrected back.
func yoCases() []corpusCase {
	cases := []corpusCase{
		{name: yoElka, tok: yoElka, mode: testModeRU},
		{name: yoEshe, tok: yoEshe, mode: testModeRU},
		{name: yoYozhik, tok: yoYozhik, mode: testModeRU},
		{name: "Ёлка", tok: "Ёлка", mode: testModeRU},
		{name: "ЁЖИК", tok: "ЁЖИК", mode: testModeRU},
	}
	for _, w := range []string{yoElka, yoEshe, yoYozhik} {
		tok, ok := correct.Convert([]rune(w), correct.RUtoEN)
		if ok {
			cases = append(cases, corpusCase{name: "latin " + w, tok: string(tok), mode: testModeEN})
		}
	}

	return cases
}

// TestCorpus_WrongLayoutBothDirections pins criterion 1, detection side:
// ghbdtn-class words of BOTH directions, in all three registers, are found
// CONFIDENTLY through the dictionary path on the real baked data.
func TestCorpus_WrongLayoutBothDirections(t *testing.T) {
	t.Parallel()

	cases := wrongLayoutCases()
	if len(cases) < 40 {
		t.Fatalf("wrong-layout corpus under-sized: %d cases, want >= 40", len(cases))
	}

	en, ru := 0, 0
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := detect.Check([]rune(tc.tok), tc.mode, realData, realTrigrams, detect.DefaultParams())
			dir := correct.RUtoEN
			if tc.mode == testModeEN {
				dir = correct.ENtoRU
			}
			want := detect.Verdict{
				WrongLayout: true,
				Confident:   true,
				Dir:         dir,
				Reason:      detect.ReasonDictWrongLayout,
			}
			if got != want {
				t.Errorf("Check(%q, %q) = %+v, want %+v", tc.tok, tc.mode, got, want)
			}
		})
		if tc.mode == testModeEN {
			en++
		} else {
			ru++
		}
	}
	if en < 20 || ru < 20 {
		t.Fatalf("wrong-layout corpus unbalanced: en=%d ru=%d, want >= 20 each", en, ru)
	}
}

// TestCorpus_LegitWordsZeroFalsePositives pins criterion 1, safety side:
// NOT ONE legit word of the >= 400-strong both-alphabet corpus is ever a
// wrong-layout on its own mode — the false-positive budget is exactly 0.
// A raw dictionary entry must also answer the cur-hit veto verbatim: the
// dictionary hit suppresses, completeness is safety.
func TestCorpus_LegitWordsZeroFalsePositives(t *testing.T) {
	t.Parallel()

	cases := legitCases()
	if len(cases) < 400 {
		t.Fatalf("legit corpus under-sized: %d cases, want >= 400", len(cases))
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := detect.Check([]rune(tc.tok), tc.mode, realData, realTrigrams, detect.DefaultParams())
			if got.WrongLayout {
				t.Errorf("FALSE POSITIVE: Check(%q, %q) = %+v, want no wrong-layout", tc.tok, tc.mode, got)
			}
			want := detect.Verdict{Reason: detect.ReasonDictCurHit}
			if got != want {
				t.Errorf("Check(%q, %q) = %+v, want the cur-hit veto %+v", tc.tok, tc.mode, got, want)
			}
		})
	}
}

// TestCorpus_YoWords pins the yo-fold (Pitfall 3): ё-spellings resolve to
// the е-entries — a cur-hit on the ru side, a confident correction from
// the latin side, never a dictionary miss.
func TestCorpus_YoWords(t *testing.T) {
	t.Parallel()

	for _, tc := range yoCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := detect.Check([]rune(tc.tok), tc.mode, realData, realTrigrams, detect.DefaultParams())
			if tc.mode == testModeRU {
				want := detect.Verdict{Reason: detect.ReasonDictCurHit}
				if got != want {
					t.Errorf("Check(%q, ru) = %+v, want the cur-hit veto %+v", tc.tok, got, want)
				}

				return
			}
			want := detect.Verdict{
				WrongLayout: true,
				Confident:   true,
				Dir:         correct.ENtoRU,
				Reason:      detect.ReasonDictWrongLayout,
			}
			if got != want {
				t.Errorf("Check(%q, en) = %+v, want %+v", tc.tok, got, want)
			}
		})
	}
}

// TestCorpus_OOVNames pins Pitfall 2: an out-of-vocabulary word — token
// AND remap both dictionary misses — never earns a CONFIDENT wrong-layout;
// unsure is the honest answer for names, nicks and keyboard garbage. The
// construction preconditions are asserted loudly: a case whose either side
// HITS a dictionary is not OOV and must not silently sit in this corpus.
func TestCorpus_OOVNames(t *testing.T) {
	t.Parallel()

	cases := oovCases()
	if len(cases) < 20 {
		t.Fatalf("OOV corpus under-sized: %d cases, want >= 20", len(cases))
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			own, other := layouts.DictEN, layouts.DictRU
			dir := correct.ENtoRU
			if tc.mode == testModeRU {
				own, other = layouts.DictRU, layouts.DictEN
				dir = correct.RUtoEN
			}
			conv, ok := correct.Convert([]rune(tc.tok), dir)
			if hitsDict(own, tc.tok) || !ok || hitsDict(other, string(conv)) {
				t.Fatalf("corpus construction: %q is not OOV (own hit=%v, conv %q other hit=%v)",
					tc.tok, hitsDict(own, tc.tok), string(conv), ok && hitsDict(other, string(conv)))
			}

			got := detect.Check([]rune(tc.tok), tc.mode, realData, realTrigrams, detect.DefaultParams())
			if got.WrongLayout && got.Confident {
				t.Errorf("OOV %q judged CONFIDENTLY: %+v, want unsure at most", tc.tok, got)
			}
		})
	}
}

// TestCorpus_MixedTokens pins the mixed refusal on real shapes: one
// foreign letter fails the whole conversion — abstain, never a partial
// rewrite.
func TestCorpus_MixedTokens(t *testing.T) {
	t.Parallel()

	for _, tc := range mixedCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := detect.Check([]rune(tc.tok), tc.mode, realData, realTrigrams, detect.DefaultParams())
			want := detect.Verdict{Reason: detect.ReasonAbstainConvertFail}
			if got != want {
				t.Errorf("Check(%q, %q) = %+v, want %+v", tc.tok, tc.mode, got, want)
			}
		})
	}
}

// TestCorpus_ReasonVocabularyClosed pins the privacy contract (D-20/D-21,
// Pitfall 6): every reason the corpus ever collects comes from the closed
// eight-constant vocabulary — enumerated here verbatim, exactly eight —
// the Verdict struct has exactly the four token-free fields (reflection
// pin), and no corpus string ever equals a verdict's string channel.
func TestCorpus_ReasonVocabularyClosed(t *testing.T) {
	t.Parallel()

	// The closed vocabulary, enumerated: exactly eight constants.
	reasons := map[string]bool{
		detect.ReasonDictWrongLayout:    true,
		detect.ReasonDictCurHit:         true,
		detect.ReasonDictBothHit:        true,
		detect.ReasonTrigramWrongLayout: true,
		detect.ReasonTrigramUnsure:      true,
		detect.ReasonAbstainShort:       true,
		detect.ReasonAbstainNoLetters:   true,
		detect.ReasonAbstainConvertFail: true,
	}
	if len(reasons) != 8 {
		t.Fatalf("the reason vocabulary drifted: %d constants, want exactly 8", len(reasons))
	}

	// Structural privacy pin: the verdict is exactly four token-free
	// fields, and Reason is the only string among them.
	vt := reflect.TypeOf(detect.Verdict{})
	wantFields := []string{"WrongLayout", "Confident", "Dir", "Reason"}
	if vt.NumField() != len(wantFields) {
		t.Fatalf("Verdict has %d fields, want exactly %d", vt.NumField(), len(wantFields))
	}
	for i, name := range wantFields {
		field := vt.Field(i)
		if field.Name != name {
			t.Errorf("Verdict field %d = %q, want %q", i, field.Name, name)
		}
		if field.Type.Kind() == reflect.String && field.Name != "Reason" {
			t.Errorf("Verdict field %q is a string — a second string channel breaks the closed vocabulary", field.Name)
		}
	}

	// Every corpus class, every verdict: the reason stays in the set, and
	// no corpus token leaks into it.
	all := make([]corpusCase, 0, 900)
	all = append(all, wrongLayoutCases()...)
	all = append(all, legitCases()...)
	all = append(all, oovCases()...)
	all = append(all, mixedCases()...)
	all = append(all, yoCases()...)
	all = append(all, corpusCase{tok: "", mode: testModeEN}, corpusCase{tok: "12345", mode: testModeEN})
	tokens := map[string]bool{}
	for _, tc := range all {
		tokens[tc.tok] = true
		got := detect.Check([]rune(tc.tok), tc.mode, realData, realTrigrams, detect.DefaultParams())
		if !reasons[got.Reason] {
			t.Errorf("Check(%q, %q) reported an unclosed reason %q", tc.tok, tc.mode, got.Reason)
		}
	}
	for reason := range reasons {
		if tokens[reason] {
			t.Errorf("corpus token %q collides with the reason vocabulary", reason)
		}
	}
}
