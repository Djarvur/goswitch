// Command dictgen bakes the hunspell ru/en dictionaries into committed golden
// Go sources for the autocorrect detector (D-52): sorted dictionary slices
// (DictRU/DictEN) and bounded trigram tables (TriRU/TriEN).
//
// It is a dev-side tool (CORR-08): `go generate ./layouts` runs it, the
// generated files are committed, and CI only runs the golden tests over the
// committed data — no hunspell dependency in CI. Parsing follows the golden
// generator discipline of layouts/generator: hard errors on data anomalies,
// never silent skips, and regeneration is byte-for-byte deterministic.
package main

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"go/format"
	"maps"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
)

// System inputs read by the generator, and the outputs it rewrites; the
// inputs exist on the dev machine only, never in CI (CORR-08).
const (
	dictRUPath = "/usr/share/hunspell/ru_RU.dic"
	dictENPath = "/usr/share/hunspell/en_US.dic"

	dictRUFile  = "dict_ru.go"
	dictENFile  = "dict_en.go"
	trigramFile = "trigrams.go"

	filePerm = 0o644

	// maxTrigramsPerLang caps each trigram table; the golden test pins the
	// ceiling loudly.
	maxTrigramsPerLang = 4096

	// minWordLen drops 1-letter entries; trigramLen is the n-gram window.
	minWordLen = 2
	trigramLen = 3
)

// Static error values; call sites only ever wrap them with %w.
var (
	errBadCountLine  = errors.New("first line is not a valid entry count")
	errCountMismatch = errors.New("declared entry count does not match the file")
	errEmptyDict     = errors.New("dictionary parsed to zero words")
	errUnknownName   = errors.New("has no license notice")
	errUnsortedWords = errors.New("dictionary slice is not sorted")
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "dictgen: %v\n", err)
		os.Exit(1)
	}
}

// run performs the whole pipeline: parse both hunspell dictionaries, train
// both trigram tables, and rewrite the three golden sources.
func run() error {
	ru, err := parseDIC(dictRUPath)
	if err != nil {
		return err
	}
	en, err := parseDIC(dictENPath)
	if err != nil {
		return err
	}
	triRU := trainTrigrams(ru, maxTrigramsPerLang)
	triEN := trainTrigrams(en, maxTrigramsPerLang)
	for _, out := range []struct {
		file string
		emit func() ([]byte, error)
	}{
		{dictRUFile, func() ([]byte, error) { return emitSortedSlice("DictRU", ru) }},
		{dictENFile, func() ([]byte, error) { return emitSortedSlice("DictEN", en) }},
		{trigramFile, func() ([]byte, error) { return emitTrigramsFile(triRU, triEN) }},
	} {
		src, err := out.emit()
		if err != nil {
			return err
		}
		if err := os.WriteFile(out.file, src, filePerm); err != nil {
			return fmt.Errorf("write %s: %w", out.file, err)
		}
	}

	return nil
}

// parseDIC reads one hunspell .dic file and returns its sorted, deduplicated
// word list: the count line is verified loudly, affix flags after '/' are
// stripped, words shorter than two runes are dropped, and ё/Ё fold to е/Е
// (Pitfall 3 — the ru dictionary is an е-dictionary).
func parseDIC(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		if serr := sc.Err(); serr != nil {
			return nil, fmt.Errorf("scan %s: %w", path, serr)
		}

		return nil, fmt.Errorf("%s: %w", path, errEmptyDict)
	}
	first := strings.TrimSpace(sc.Text())
	declared, err := strconv.Atoi(first)
	if err != nil || declared < 0 {
		return nil, fmt.Errorf("%s: %w: %q", path, errBadCountLine, first)
	}

	seen := make(map[string]struct{}, declared)
	words := make([]string, 0, declared)
	entries := 0
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		entries++
		word := foldYO(beforeFlags(line))
		if len([]rune(word)) < minWordLen {
			continue
		}
		if _, dup := seen[word]; dup {
			continue
		}
		seen[word] = struct{}{}
		words = append(words, word)
	}
	if serr := sc.Err(); serr != nil {
		return nil, fmt.Errorf("scan %s: %w", path, serr)
	}
	if entries != declared {
		return nil, fmt.Errorf("%s: %w: declared %d entries, found %d", path, errCountMismatch, declared, entries)
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("%s: %w", path, errEmptyDict)
	}
	slices.Sort(words)

	return words, nil
}

// beforeFlags strips the affix-flag suffix: everything from the first '/' on.
func beforeFlags(s string) string {
	if i := strings.IndexByte(s, '/'); i >= 0 {
		return s[:i]
	}

	return s
}

// foldYO replaces ё with е and Ё with Е so a typed е-form finds ё-spelled
// dictionary entries and vice versa; the ru dictionary itself is е-spelled.
func foldYO(s string) string {
	if !strings.ContainsAny(s, "ёЁ") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case 'ё':
			r = 'е'
		case 'Ё':
			r = 'Е'
		}
		b.WriteRune(r)
	}

	return b.String()
}

// trainTrigrams trains a trigram table over words: every 3-rune window counts
// once per occurrence, values are log10(count/total) — non-positive by
// construction — and only the top-k entries survive with an alphabetical
// tie-break so the table is deterministic. A non-positive cap yields the
// empty table; words shorter than three runes contribute nothing.
func trainTrigrams(words []string, k int) map[string]float64 {
	counts := make(map[string]int)
	total := 0
	for _, w := range words {
		runes := []rune(w)
		for i := 0; i+trigramLen <= len(runes); i++ {
			counts[string(runes[i:i+trigramLen])]++
			total++
		}
	}
	// Alphabetical order first, then a STABLE sort by descending count: ties
	// keep their alphabetical order, so the top-k cut is deterministic.
	grams := slices.Sorted(maps.Keys(counts))
	slices.SortStableFunc(grams, func(a, b string) int {
		return cmp.Compare(counts[b], counts[a])
	})
	if k > 0 && len(grams) > k {
		grams = grams[:k]
	}
	table := make(map[string]float64, len(grams))
	for _, g := range grams {
		table[g] = math.Log10(float64(counts[g]) / float64(total))
	}

	return table
}

const genMarker = "// Code generated by goswitch/layouts/dictgen. DO NOT EDIT.\n"

const dictRUHeader = genMarker + `
//go:generate go run ./dictgen

// Source: hunspell ru_RU (Debian hunspell-ru 1:24.2.1-1). Copyright 1997-2008
// Alexander I. Lebedev, custom-bsd-4-clauses; the verbatim notice lives in
// docs/LICENSE-data.md. Baked from the full dictionary with the yo-to-ye fold
// applied (Pitfall 3), so no yo glyph occurs anywhere in this file.
package layouts

// DictRU holds the full Russian dictionary sorted for sort.SearchStrings: a
// detector dictionary hit SUPPRESSES correction, so full coverage is safety
// (D-52). Never build a runtime map from it — the slice lives in rodata.
`

const dictENHeader = genMarker + `
//go:generate go run ./dictgen

// Source: hunspell en_US (Debian hunspell-en-us 1:2020.12.07-2, SCOWL and
// public-domain contributors). Copyright Kevin Atkinson and others,
// permissive; the verbatim notices live in docs/LICENSE-data.md.
package layouts

// DictEN holds the full American English dictionary sorted for
// sort.SearchStrings: a detector dictionary hit SUPPRESSES correction, so
// full coverage is safety (D-52). Never build a runtime map from it.
`

const triFileHeader = genMarker + `
//go:generate go run ./dictgen

// Source: trained by layouts/dictgen over the baked dictionaries themselves —
// ru data © 1997-2008 Alexander I. Lebedev (custom-bsd-4-clauses), en data
// © Kevin Atkinson and SCOWL/public-domain contributors (permissive); the
// verbatim notices live in docs/LICENSE-data.md.
package layouts
`

const triDocRU = `// TriRU holds log10 trigram probabilities of the ru dictionary for the
// detector's fallback scoring (D-52), capped at maxTrigramsPerLang entries
// by top count with an alphabetical tie-break.
`

const triDocEN = `// TriEN holds log10 trigram probabilities of the en dictionary for the
// detector's fallback scoring (D-52), capped at maxTrigramsPerLang entries
// by top count with an alphabetical tie-break.
`

// Names of the trigram tables (goconst discipline of the strict lint).
const (
	nameTriRU = "TriRU"
	nameTriEN = "TriEN"
)

// emitSortedSlice renders one generated dictionary file: the DO NOT EDIT
// header with the go:generate directive and the data license notice, followed
// by the sorted []string literal — never a map literal (a runtime map costs
// +10–12 МБ RSS against ≈0 heap for a rodata slice, 06-RESEARCH Q3).
func emitSortedSlice(name string, words []string) ([]byte, error) {
	if !slices.IsSorted(words) {
		return nil, fmt.Errorf("%s: %w", name, errUnsortedWords)
	}
	header, err := sliceHeader(name)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteString(header)
	fmt.Fprintf(&buf, "var %s = []string{\n", name)
	for _, w := range words {
		fmt.Fprintf(&buf, "\t%s,\n", strconv.Quote(w))
	}
	buf.WriteString("}\n")

	return formatSource(buf.Bytes(), name)
}

// emitTrigramsFile renders the whole trigram file: one header, then both
// documented tables with keys sorted for byte-for-byte deterministic
// regeneration.
func emitTrigramsFile(triRU, triEN map[string]float64) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(triFileHeader)
	buf.WriteString("\n")
	for _, table := range []struct {
		name string
		t    map[string]float64
	}{
		{nameTriRU, triRU},
		{nameTriEN, triEN},
	} {
		if table.name == nameTriEN {
			buf.WriteString("\n")
		}
		body, err := emitTrigrams(table.name, table.t)
		if err != nil {
			return nil, err
		}
		buf.Write(body)
	}

	return formatSource(buf.Bytes(), trigramFile)
}

// emitTrigrams renders one documented trigram-map declaration: keys sorted so
// regeneration is byte-for-byte deterministic, float values in the shortest
// deterministic form.
func emitTrigrams(name string, t map[string]float64) ([]byte, error) {
	doc, err := trigramDoc(name)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteString(doc)
	fmt.Fprintf(&buf, "var %s = map[string]float64{\n", name)
	for _, g := range slices.Sorted(maps.Keys(t)) {
		fmt.Fprintf(&buf, "\t%s: %s,\n", strconv.Quote(g), strconv.FormatFloat(t[g], 'g', -1, 64))
	}
	buf.WriteString("}\n")

	return buf.Bytes(), nil
}

// sliceHeader returns the file header for a dictionary var name.
func sliceHeader(name string) (string, error) {
	switch name {
	case "DictRU":
		return dictRUHeader, nil
	case "DictEN":
		return dictENHeader, nil
	default:
		return "", fmt.Errorf("%s: %w", name, errUnknownName)
	}
}

// trigramDoc returns the var doc comment for a trigram table name.
func trigramDoc(name string) (string, error) {
	switch name {
	case nameTriRU:
		return triDocRU, nil
	case nameTriEN:
		return triDocEN, nil
	default:
		return "", fmt.Errorf("%s: %w", name, errUnknownName)
	}
}

// formatSource renders generated source gofmt-clean by construction.
func formatSource(src []byte, name string) ([]byte, error) {
	out, err := format.Source(src)
	if err != nil {
		return nil, fmt.Errorf("format %s: %w", name, err)
	}

	return out, nil
}
