package main

import (
	"bytes"
	"errors"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Corpus literals named as constants (goconst discipline of the strict lint).
const (
	wordABC = "abc"
	wordBCD = "bcd"
	wordABE = "abe"
	wordBEF = "bef"
	wordBBB = "bbb"
	wordCCC = "ccc"
)

// assertParsed parses path and pins the exact sorted word list plus the fold
// and minimum-length invariants over every word.
func assertParsed(t *testing.T, path string, want []string) {
	t.Helper()

	got, err := parseDIC(path)
	if err != nil {
		t.Fatalf("parseDIC: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("parseDIC = %q, want %q", got, want)
	}
	for _, w := range got {
		if strings.ContainsAny(w, "ёЁ") {
			t.Errorf("word %q still contains ё/Ё", w)
		}
		if len([]rune(w)) < 2 {
			t.Errorf("word %q is shorter than 2 runes", w)
		}
	}
}

// TestParseDIC pins the hunspell .dic pipeline on fixtures: the count line is
// skipped and verified loudly, affix flags are stripped, 1-letter words are
// dropped, ё/Ё fold to е/Е, duplicates collapse, and the result is sorted.
func TestParseDIC(t *testing.T) {
	t.Parallel()

	t.Run("ru-fixture", func(t *testing.T) {
		t.Parallel()

		// ёлка folds to елка (collapsing with the explicit елка entry), Ёлка
		// keeps its register (Елка), both ещё fold to еще, and а/б drop.
		assertParsed(t, "testdata/ru_fixture.dic", []string{"Елка", "елка", "еще", "кот", "мир", "привет"})
	})

	t.Run("en-fixture", func(t *testing.T) {
		t.Parallel()

		assertParsed(t, "testdata/en_fixture.dic", []string{"cat", "dog", "hello", "run", "world", "zebra"})
	})

	t.Run("refusals", func(t *testing.T) {
		t.Parallel()

		// Every anomaly is a loud error, never a silent skip (T-06-02-04).
		// A nil want pins "any non-nil error" — the os-level open failure.
		cases := []struct {
			name  string
			path  string // fixed testdata path; empty writes the content below
			write string // fixture content for a temp file
			want  error
		}{
			{name: "count-mismatch", path: "testdata/ru_bad_count.dic", want: errCountMismatch},
			{name: "bad-count-line", write: "not-a-number\nword\n", want: errBadCountLine},
			{name: "empty-dictionary", write: "0\n", want: errEmptyDict},
			{name: "missing-file", path: "testdata/absent.dic"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				path := tc.path
				if tc.write != "" {
					path = filepath.Join(t.TempDir(), tc.name+".dic")
					if err := os.WriteFile(path, []byte(tc.write), 0o600); err != nil {
						t.Fatalf("write fixture: %v", err)
					}
				}
				_, err := parseDIC(path)
				if tc.want == nil {
					if err == nil {
						t.Errorf("parseDIC(%s) returned nil error, want an open failure", path)
					}

					return
				}
				if !errors.Is(err, tc.want) {
					t.Errorf("parseDIC(%s) error = %v, want %v", path, err, tc.want)
				}
			})
		}
	})
}

// wantGram asserts one trigram value equals log10(count/total).
func wantGram(t *testing.T, table map[string]float64, gram string, count, total float64) {
	t.Helper()

	want := math.Log10(count / total)
	if math.Abs(table[gram]-want) > 1e-12 {
		t.Errorf("table[%s] = %v, want %v", gram, table[gram], want)
	}
}

// TestTrigramTrain pins the trigram training: 3-rune windows over every word,
// words shorter than three runes contribute nothing, values are
// log10(count/total), the ceiling keeps the top-k with an alphabetical
// tie-break, and repeated training is deterministic.
func TestTrigramTrain(t *testing.T) {
	t.Parallel()

	t.Run("windows-and-log10", func(t *testing.T) {
		t.Parallel()

		got := trainTrigrams([]string{wordABC, wordABC, wordABC + "d"}, 10)
		if len(got) != 2 {
			t.Fatalf("len(table) = %d, want 2 (%s x3, %s x1)", len(got), wordABC, wordBCD)
		}
		wantGram(t, got, wordABC, 3, 4)
		wantGram(t, got, wordBCD, 1, 4)
		for gram, val := range got {
			if val > 0 {
				t.Errorf("table[%s] = %v is positive; log-probabilities are non-positive", gram, val)
			}
		}
	})

	t.Run("short-words-contribute-nothing", func(t *testing.T) {
		t.Parallel()

		if got := trainTrigrams([]string{"ab", "a", ""}, 10); len(got) != 0 {
			t.Errorf("len(table) = %d, want 0 for words under 3 runes", len(got))
		}
	})

	t.Run("ceiling-with-alphabetical-tie-break", func(t *testing.T) {
		t.Parallel()

		got := trainTrigrams([]string{wordABC + "d", wordABE + "f"}, 2)
		// All four trigrams tie at count 1; the ceiling keeps the two
		// alphabetically first ones.
		if len(got) != 2 {
			t.Fatalf("len(table) = %d, want 2", len(got))
		}
		for _, gram := range []string{wordABC, wordABE} {
			if _, ok := got[gram]; !ok {
				t.Errorf("table misses %q; tie-break is not alphabetical", gram)
			}
		}
		for _, gram := range []string{wordBCD, wordBEF} {
			if _, ok := got[gram]; ok {
				t.Errorf("table keeps %q past the ceiling", gram)
			}
		}
	})

	t.Run("deterministic", func(t *testing.T) {
		t.Parallel()

		words := []string{"ёлка", "привет", "ghbdtn", "ещё", "mir"}
		first := trainTrigrams(words, 8)
		second := trainTrigrams(words, 8)
		if len(first) == 0 {
			t.Fatal("training produced an empty table")
		}
		for gram, val := range first {
			if second[gram] != val {
				t.Errorf("table[%s] differs across runs: %v vs %v", gram, val, second[gram])
			}
		}
	})
}

// TestEmitDeterministic pins the emission: identical inputs render byte-equal
// gofmt-clean sources whose headers carry the DO NOT EDIT marker, the
// generate directive, the package clause, and the data license notice.
func TestEmitDeterministic(t *testing.T) {
	t.Parallel()

	t.Run("dict-slice", func(t *testing.T) {
		t.Parallel()

		words := []string{wordBBB, wordCCC}
		first, err := emitSortedSlice("DictRU", words)
		if err != nil {
			t.Fatalf("emitSortedSlice: %v", err)
		}
		second, err := emitSortedSlice("DictRU", words)
		if err != nil {
			t.Fatalf("emitSortedSlice: %v", err)
		}
		if !bytes.Equal(first, second) {
			t.Error("two emissions of the same input differ byte-wise")
		}
		for _, marker := range []string{"DO NOT EDIT", "//go:generate go run ./dictgen", "package layouts", "Lebedev"} {
			if !bytes.Contains(first, []byte(marker)) {
				t.Errorf("emitted source misses %q in the header", marker)
			}
		}
		en, err := emitSortedSlice("DictEN", []string{wordBBB})
		if err != nil {
			t.Fatalf("emitSortedSlice: %v", err)
		}
		if !bytes.Contains(en, []byte("SCOWL")) {
			t.Error("DictEN source misses the SCOWL license notice")
		}
	})

	t.Run("trigrams", func(t *testing.T) {
		t.Parallel()

		table := map[string]float64{"абв": -1.5, "где": -2.25}
		first, err := emitTrigrams("TriRU", table)
		if err != nil {
			t.Fatalf("emitTrigrams: %v", err)
		}
		second, err := emitTrigrams("TriRU", table)
		if err != nil {
			t.Fatalf("emitTrigrams: %v", err)
		}
		if !bytes.Equal(first, second) {
			t.Error("two trigram emissions of the same input differ byte-wise")
		}
		if !bytes.Contains(first, []byte("TriRU")) {
			t.Error("emitted trigram source misses the var name")
		}
	})
}

// TestEmitRefusals pins the emission's loud refusals: an unsorted dictionary
// slice and an unknown var name are hard errors, never silent output.
func TestEmitRefusals(t *testing.T) {
	t.Parallel()

	if _, err := emitSortedSlice("DictRU", []string{"b", "a"}); !errors.Is(err, errUnsortedWords) {
		t.Errorf("emitSortedSlice unsorted error = %v, want errUnsortedWords", err)
	}
	if _, err := emitSortedSlice("DictXX", []string{"a"}); !errors.Is(err, errUnknownName) {
		t.Errorf("emitSortedSlice unknown name error = %v, want errUnknownName", err)
	}
}

// TestGenerateLive pins the full dev-machine run against the installed
// hunspell dictionaries: the parsed counts land inside the measured envelopes.
// Measured on hunspell-ru 1:24.2.1-1 / hunspell-en-us 1:2020.12.07-2:
// 146 269 ru entries − 8 short forms = 146 261 unique, ё→е folding merges
// 7347 е/ё lemma pairs → 138 914 (the research Q3 figure of 146 261 predates
// the folding decision); en: 79 013 → 78 951. CI machines without hunspell
// skip the test — the golden contract (CORR-08) never requires the live input.
func TestGenerateLive(t *testing.T) {
	t.Parallel()

	if _, err := os.Stat(dictRUPath); err != nil {
		t.Skipf("live ru dictionary is absent (%v) — golden CI path", err)
	}
	if _, err := os.Stat(dictENPath); err != nil {
		t.Skipf("live en dictionary is absent (%v) — golden CI path", err)
	}

	ru, err := parseDIC(dictRUPath)
	if err != nil {
		t.Fatalf("parseDIC(%s): %v", dictRUPath, err)
	}
	if len(ru) < 137000 || len(ru) > 140500 {
		t.Errorf("len(DictRU) = %d, want within [137000, 140500]", len(ru))
	}
	if !slices.IsSorted(ru) {
		t.Error("live ru dictionary is not sorted")
	}

	en, err := parseDIC(dictENPath)
	if err != nil {
		t.Fatalf("parseDIC(%s): %v", dictENPath, err)
	}
	if len(en) < 78000 || len(en) > 80000 {
		t.Errorf("len(DictEN) = %d, want within [78000, 80000]", len(en))
	}
	if !slices.IsSorted(en) {
		t.Error("live en dictionary is not sorted")
	}
}
