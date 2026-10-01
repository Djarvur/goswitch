// Command dictgen bakes the hunspell ru/en dictionaries into committed golden
// Go sources for the autocorrect detector (D-52): sorted dictionary slices
// (DictRU/DictEN) and bounded trigram tables (TriRU/TriEN).
//
// It is a dev-side tool (CORR-08): `go generate ./layouts` runs it, the
// generated files are committed, and CI only runs the golden tests over the
// committed data — no hunspell dependency in CI. TDD RED scaffold: the
// pipeline functions below are stubbed declarations; the corpus in main_test.go
// pins their behavior before the implementation lands (D-07).
package main

import "errors"

// System inputs read by the generator; dev-side only, never touched in CI.
const (
	dictRUPath = "/usr/share/hunspell/ru_RU.dic"
	dictENPath = "/usr/share/hunspell/en_US.dic"
)

var (
	errBadCountLine  = errors.New("first line is not a valid entry count")
	errCountMismatch = errors.New("declared entry count does not match the file")
	errEmptyDict     = errors.New("dictionary parsed to zero words")
	errUnknownName   = errors.New("has no license notice")
	errUnsortedWords = errors.New("dictionary slice is not sorted")
)

func main() {}

// parseDIC reads one hunspell .dic file and returns its sorted, deduplicated
// word list: the count line is verified loudly, affix flags after '/' are
// stripped, words shorter than two runes are dropped, and ё/Ё fold to е/Е.
func parseDIC(path string) ([]string, error) { return nil, nil }

// trainTrigrams trains a trigram table over words: every 3-rune window counts
// once, values are log10(count/total), and only the top-k entries survive with
// an alphabetical tie-break so the table is deterministic.
func trainTrigrams(words []string, k int) map[string]float64 { return nil }

// emitSortedSlice renders one generated dictionary file: a DO NOT EDIT header
// with the go:generate directive and the data license notice, followed by the
// sorted []string literal — never a map literal.
func emitSortedSlice(name string, words []string) ([]byte, error) { return nil, nil }

// emitTrigrams renders one documented trigram-map declaration with keys sorted
// for byte-for-byte deterministic regeneration.
func emitTrigrams(name string, t map[string]float64) ([]byte, error) { return nil, nil }
