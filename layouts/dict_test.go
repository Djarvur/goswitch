package layouts_test

import (
	"slices"
	"strings"
	"testing"
	"unicode"

	"github.com/Djarvur/goswitch/layouts"
)

// maxTrigramsPerLang is the generator ceiling pinned here INDEPENDENTLY of
// layouts/dictgen: if the generator ever loses its cap, this test fails
// loudly (T-06-02-02).
const maxTrigramsPerLang = 4096

// TestGolden_DictSorted pins that both baked dictionaries are sorted — the
// contract the detector's sort.SearchStrings lookup relies on (research Q3).
func TestGolden_DictSorted(t *testing.T) {
	t.Parallel()

	if !slices.IsSorted(layouts.DictRU) {
		t.Error("DictRU is not sorted")
	}
	if !slices.IsSorted(layouts.DictEN) {
		t.Error("DictEN is not sorted")
	}
}

// TestGolden_DictDisjoint pins that the two dictionary scripts are disjoint:
// no Latin letter in any ru word, no Cyrillic letter in any en word. A mixed
// word would mean a broken remap verdict in the detector (D-52а).
func TestGolden_DictDisjoint(t *testing.T) {
	t.Parallel()

	for _, w := range layouts.DictRU {
		if strings.IndexFunc(w, func(r rune) bool { return unicode.Is(unicode.Latin, r) }) >= 0 {
			t.Errorf("DictRU word %q contains a Latin letter", w)
		}
	}
	for _, w := range layouts.DictEN {
		if strings.IndexFunc(w, func(r rune) bool { return unicode.Is(unicode.Cyrillic, r) }) >= 0 {
			t.Errorf("DictEN word %q contains a Cyrillic letter", w)
		}
	}
}

// TestGolden_DictNoYo pins the yo-to-ye fold of the ru dictionary (Pitfall 3 /
// R3): no word contains ё or Ё, so a typed е-form always finds its entry.
func TestGolden_DictNoYo(t *testing.T) {
	t.Parallel()

	for _, w := range layouts.DictRU {
		if strings.ContainsRune(w, 'ё') || strings.ContainsRune(w, 'Ё') {
			t.Errorf("DictRU word %q still contains a yo glyph", w)
		}
	}
}

// TestGolden_DictEnvelope pins the measured dictionary sizes so a package
// update or a corrupted regeneration fails loudly. Measured by dictgen on
// hunspell-ru 1:24.2.1-1 / hunspell-en-us 1:2020.12.07-2: ru 146 269 entries
// minus 8 short forms = 146 261 unique, the yo-to-ye fold merges 7347 е/ё
// lemma pairs (the research Q3 figure of 146 261 predates the fold), en
// 79 013 entries → 78 951. Bounds allow the drift of a dictionary package
// update; a deliberate re-pin follows every accepted drift.
func TestGolden_DictEnvelope(t *testing.T) {
	t.Parallel()

	if len(layouts.DictRU) < 137000 || len(layouts.DictRU) > 140500 {
		t.Errorf("len(DictRU) = %d, want within [137000, 140500]", len(layouts.DictRU))
	}
	if len(layouts.DictEN) < 78000 || len(layouts.DictEN) > 80000 {
		t.Errorf("len(DictEN) = %d, want within [78000, 80000]", len(layouts.DictEN))
	}
}

// TestGolden_TrigramsBounded pins the trigram tables: non-empty, within the
// per-language ceiling, and every value a non-positive log-probability.
func TestGolden_TrigramsBounded(t *testing.T) {
	t.Parallel()

	if len(layouts.TriRU) == 0 || len(layouts.TriEN) == 0 {
		t.Fatalf("trigram tables empty: TriRU=%d, TriEN=%d", len(layouts.TriRU), len(layouts.TriEN))
	}
	if len(layouts.TriRU) > maxTrigramsPerLang {
		t.Errorf("len(TriRU) = %d, want <= %d", len(layouts.TriRU), maxTrigramsPerLang)
	}
	if len(layouts.TriEN) > maxTrigramsPerLang {
		t.Errorf("len(TriEN) = %d, want <= %d", len(layouts.TriEN), maxTrigramsPerLang)
	}
	for gram, val := range layouts.TriRU {
		if val > 0 {
			t.Errorf("TriRU[%q] = %v is positive; log-probabilities are non-positive", gram, val)
		}
	}
	for gram, val := range layouts.TriEN {
		if val > 0 {
			t.Errorf("TriEN[%q] = %v is positive; log-probabilities are non-positive", gram, val)
		}
	}
}
