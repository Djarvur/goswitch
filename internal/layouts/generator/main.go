// Command generator rebuilds layouts/tables.go from the system XKB symbol
// files and the IBus keysym header.
//
// It is a dev-side tool (CORR-08): `go generate ./layouts` runs it, the
// generated file is committed, and CI only runs the golden tests over the
// committed tables — no xkb dependency in CI. The ru side resolves the
// file's default variant (winkeys on Ubuntu 24.04) and recursively expands
// its same-file includes (ru(common)); the keypad include kpdl(comma) is
// skipped as keypad-only.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// System inputs read by the generator; dev-side only, never touched in CI.
const (
	usSymbolsPath = "/usr/share/X11/xkb/symbols/us"
	ruSymbolsPath = "/usr/share/X11/xkb/symbols/ru"
	keysymsHeader = "/usr/include/ibus-1.0/ibuskeysyms.h"
	generatedFile = "tables.go"

	filePerm = 0o644

	// A keysym value in this range maps to its Latin-1 rune directly.
	latin1Min = 0x20
	latin1Max = 0xff

	// The generator joins the first two shift levels of every key.
	levelsPerKey = 2
)

// Static error values; call sites only ever wrap them with %w.
var (
	errNoDefaultSection = errors.New("no default xkb_symbols section")
	errDuplicateSection = errors.New("duplicate xkb_symbols section")
	errNoSuchSection    = errors.New("has no xkb_symbols section")
	errCrossFileInclude = errors.New("include references another symbols file")
	errMissingKey       = errors.New("layout is missing key")
	errConflictingEntry = errors.New("maps to two runes")
	errUnknownKeysym    = errors.New("keysym is not defined")
	errNoRuneMapping    = errors.New("keysym has no rune mapping")
)

var (
	sectionRe = regexp.MustCompile(`^xkb_symbols\s+"([^"]+)"\s*\{`)
	keyRe     = regexp.MustCompile(`^key\s+<([A-Za-z0-9]+)>\s*\{?\s*\[([^]]*)]`)
	includeRe = regexp.MustCompile(`^include\s+"([^"]+)"`)
	defineRe  = regexp.MustCompile(`^#define\s+IBUS_KEY_(\S+)\s+(0x[0-9a-fA-F]+|\d+)$`)
)

// keyDef is one `key <NAME> {[ lvl1, lvl2, ... ]}` statement.
type keyDef struct {
	name   string
	levels []string
}

// bodyEntry is one ordered statement of an xkb_symbols section: an include
// directive (include != "") or a key definition.
type bodyEntry struct {
	include string
	key     keyDef
}

// symbolsFile is the parsed form of one xkb symbols file.
type symbolsFile struct {
	sections    map[string][]bodyEntry
	defaultName string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "generator: %v\n", err)
		os.Exit(1)
	}
}

// run performs the whole pipeline: load keysym values, resolve both
// effective layouts, join them by key position, and write tables.go.
func run() error {
	order := keyOrder()
	cyrillic := cyrillicRunes()
	keyvals, err := loadKeysymValues(keysymsHeader)
	if err != nil {
		return err
	}
	usKeys, err := effectiveDefaultLayout(usSymbolsPath, "us", order)
	if err != nil {
		return err
	}
	ruKeys, err := effectiveDefaultLayout(ruSymbolsPath, "ru", order)
	if err != nil {
		return err
	}
	enToRu, ruToEn, err := buildTables(usKeys, ruKeys, keyvals, order, cyrillic)
	if err != nil {
		return err
	}
	src, err := emitTables(enToRu, ruToEn)
	if err != nil {
		return err
	}
	if err := os.WriteFile(generatedFile, src, filePerm); err != nil {
		return fmt.Errorf("write %s: %w", generatedFile, err)
	}

	return nil
}

// keyOrder returns the 47 xkb key names of the us(basic) alphanumeric block;
// the EN↔RU join is performed on these positions only (LSGT and keypad keys
// are deliberately out of scope).
func keyOrder() []string {
	return []string{
		"TLDE",
		"AE01", "AE02", "AE03", "AE04", "AE05", "AE06",
		"AE07", "AE08", "AE09", "AE10", "AE11", "AE12",
		"AD01", "AD02", "AD03", "AD04", "AD05", "AD06",
		"AD07", "AD08", "AD09", "AD10", "AD11", "AD12",
		"AC01", "AC02", "AC03", "AC04", "AC05", "AC06",
		"AC07", "AC08", "AC09", "AC10", "AC11",
		"AB01", "AB02", "AB03", "AB04", "AB05", "AB06",
		"AB07", "AB08", "AB09", "AB10",
		"BKSL",
	}
}

// cyrillicRunes returns the keysym-name → Unicode table for the Cyrillic
// names used by the ru layout (and numerosign). Values outside the Latin-1
// range have no direct rune correspondence, so this table is the authority;
// the golden corpus pins it at generation time, never at runtime.
func cyrillicRunes() map[string]rune {
	return map[string]rune{
		"Cyrillic_A": 'А', "Cyrillic_a": 'а',
		"Cyrillic_BE": 'Б', "Cyrillic_be": 'б',
		"Cyrillic_CHE": 'Ч', "Cyrillic_che": 'ч',
		"Cyrillic_DE": 'Д', "Cyrillic_de": 'д',
		"Cyrillic_E": 'Э', "Cyrillic_e": 'э',
		"Cyrillic_EF": 'Ф', "Cyrillic_ef": 'ф',
		"Cyrillic_EL": 'Л', "Cyrillic_el": 'л',
		"Cyrillic_EM": 'М', "Cyrillic_em": 'м',
		"Cyrillic_EN": 'Н', "Cyrillic_en": 'н',
		"Cyrillic_ER": 'Р', "Cyrillic_er": 'р',
		"Cyrillic_ES": 'С', "Cyrillic_es": 'с',
		"Cyrillic_GHE": 'Г', "Cyrillic_ghe": 'г',
		"Cyrillic_HA": 'Х', "Cyrillic_ha": 'х',
		"Cyrillic_HARDSIGN": 'Ъ', "Cyrillic_hardsign": 'ъ',
		"Cyrillic_IE": 'Е', "Cyrillic_ie": 'е',
		"Cyrillic_IO": 'Ё', "Cyrillic_io": 'ё',
		"Cyrillic_I": 'И', "Cyrillic_i": 'и',
		"Cyrillic_KA": 'К', "Cyrillic_ka": 'к',
		"Cyrillic_O": 'О', "Cyrillic_o": 'о',
		"Cyrillic_PE": 'П', "Cyrillic_pe": 'п',
		"Cyrillic_SHA": 'Ш', "Cyrillic_sha": 'ш',
		"Cyrillic_SHCHA": 'Щ', "Cyrillic_shcha": 'щ',
		"Cyrillic_SHORTI": 'Й', "Cyrillic_shorti": 'й',
		"Cyrillic_SOFTSIGN": 'Ь', "Cyrillic_softsign": 'ь',
		"Cyrillic_TE": 'Т', "Cyrillic_te": 'т',
		"Cyrillic_TSE": 'Ц', "Cyrillic_tse": 'ц',
		"Cyrillic_U": 'У', "Cyrillic_u": 'у',
		"Cyrillic_VE": 'В', "Cyrillic_ve": 'в',
		"Cyrillic_YA": 'Я', "Cyrillic_ya": 'я',
		"Cyrillic_YERU": 'Ы', "Cyrillic_yeru": 'ы',
		"Cyrillic_YU": 'Ю', "Cyrillic_yu": 'ю',
		"Cyrillic_ZE": 'З', "Cyrillic_ze": 'з',
		"Cyrillic_ZHE": 'Ж', "Cyrillic_zhe": 'ж',
		"numerosign": '№',
	}
}

// loadKeysymValues parses `#define IBUS_KEY_<name> <value>` lines into a
// name → keysym value table.
func loadKeysymValues(header string) (map[string]uint32, error) {
	data, err := os.ReadFile(header)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", header, err)
	}
	values := make(map[string]uint32)
	for _, raw := range strings.Split(string(data), "\n") {
		m := defineRe.FindStringSubmatch(strings.TrimSpace(raw))
		if m == nil {
			continue
		}
		val, err := strconv.ParseUint(m[2], 0, 32)
		if err != nil {
			return nil, fmt.Errorf("keysym %s: %w", m[1], err)
		}
		values[m[1]] = uint32(val)
	}

	return values, nil
}

// effectiveDefaultLayout reads one symbols file and resolves its default
// xkb_symbols section into an effective keymap: same-file includes are
// expanded depth-first in order, and later key statements override earlier
// ones (winkeys overriding ru(common)).
func effectiveDefaultLayout(path, selfName string, order []string) (map[string][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	file, err := parseSymbols(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if file.defaultName == "" {
		return nil, fmt.Errorf("symbols file %s: %w", selfName, errNoDefaultSection)
	}
	keys := make(map[string][]string, len(order))
	if err := expandLayout(file, selfName, file.defaultName, order, keys); err != nil {
		return nil, fmt.Errorf("resolve %s default %q: %w", selfName, file.defaultName, err)
	}

	return keys, nil
}

// parseSymbols splits a symbols file into named sections of ordered
// statements, tracking which section the `default` prelude marks.
func parseSymbols(data string) (symbolsFile, error) {
	file := symbolsFile{sections: make(map[string][]bodyEntry)}
	var current string
	var pendingDefault bool
	for _, raw := range strings.Split(data, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case line == "" || strings.HasPrefix(line, "//"):
			continue
		case strings.HasPrefix(line, "default"):
			pendingDefault = true

			continue
		}
		if m := sectionRe.FindStringSubmatch(line); m != nil {
			if _, dup := file.sections[m[1]]; dup {
				return symbolsFile{}, fmt.Errorf("xkb_symbols %q: %w", m[1], errDuplicateSection)
			}
			current = m[1]
			if pendingDefault {
				file.defaultName = m[1]
				pendingDefault = false
			}

			continue
		}
		if current == "" {
			continue
		}
		if m := includeRe.FindStringSubmatch(line); m != nil {
			file.sections[current] = append(file.sections[current], bodyEntry{include: m[1]})

			continue
		}
		if m := keyRe.FindStringSubmatch(line); m != nil {
			entry := bodyEntry{key: keyDef{name: m[1], levels: splitLevels(m[2])}}
			file.sections[current] = append(file.sections[current], entry)
		}
	}

	return file, nil
}

// splitLevels splits the bracket body of a key statement into trimmed level
// names, dropping empties.
func splitLevels(bracket string) []string {
	fields := strings.Split(bracket, ",")
	levels := make([]string, 0, len(fields))
	for _, field := range fields {
		if name := strings.TrimSpace(field); name != "" {
			levels = append(levels, name)
		}
	}

	return levels
}

// expandLayout resolves one section into out, honouring statement order:
// includes expand first, key overrides after an include win.
func expandLayout(file symbolsFile, selfName, section string, order []string, out map[string][]string) error {
	entries, ok := file.sections[section]
	if !ok {
		return fmt.Errorf("symbols file %s: %w %q", selfName, errNoSuchSection, section)
	}
	for _, entry := range entries {
		if entry.include != "" {
			if err := expandInclude(file, selfName, entry.include, order, out); err != nil {
				return err
			}

			continue
		}
		if !slices.Contains(order, entry.key.name) || len(entry.key.levels) < levelsPerKey {
			continue
		}
		out[entry.key.name] = entry.key.levels[:levelsPerKey:levelsPerKey]
	}

	return nil
}

// expandInclude resolves one include directive: same-file sections recurse,
// the keypad-only kpdl include is skipped, anything cross-file is an error.
func expandInclude(
	file symbolsFile, selfName, ref string,
	order []string, out map[string][]string,
) error {
	incFile, incSection := splitInclude(ref)
	switch {
	case incFile == "kpdl":
		// Keypad-only (kpdl(comma) inside ru(common)); outside the join.

		return nil
	case incFile != selfName:
		return fmt.Errorf("%q: %w", ref, errCrossFileInclude)
	}
	if err := expandLayout(file, selfName, incSection, order, out); err != nil {
		return fmt.Errorf("expand %q: %w", ref, err)
	}

	return nil
}

// splitInclude splits an include reference into file and section parts:
// "ru(common)" → ("ru", "common"), "us" → ("us", "").
func splitInclude(ref string) (file, section string) {
	name, inner, found := strings.Cut(ref, "(")
	if !found {
		return name, ""
	}

	return name, strings.TrimSuffix(inner, ")")
}

// buildTables joins the two effective layouts by key position and produces
// both direction tables; a rune colliding with a different mapping is a hard
// error, not a silent overwrite.
func buildTables(
	usKeys, ruKeys map[string][]string,
	keyvals map[string]uint32,
	order []string,
	cyrillic map[string]rune,
) (enToRu, ruToEn map[rune]rune, err error) {
	enToRu = make(map[rune]rune, len(order)*levelsPerKey)
	ruToEn = make(map[rune]rune, len(order)*levelsPerKey)
	for _, name := range order {
		usLevels, ok := usKeys[name]
		if !ok {
			return nil, nil, fmt.Errorf("us %w %s", errMissingKey, name)
		}
		ruLevels, ok := ruKeys[name]
		if !ok {
			return nil, nil, fmt.Errorf("ru %w %s", errMissingKey, name)
		}
		if err := joinKey(usLevels, ruLevels, keyvals, cyrillic, enToRu, ruToEn); err != nil {
			return nil, nil, fmt.Errorf("key %s: %w", name, err)
		}
	}

	return enToRu, ruToEn, nil
}

// joinKey adds both shift levels of one key position to the tables.
func joinKey(
	usLevels, ruLevels []string,
	keyvals map[string]uint32,
	cyrillic map[string]rune,
	enToRu, ruToEn map[rune]rune,
) error {
	for i := range levelsPerKey {
		enRune, err := runeFor(usLevels[i], keyvals, cyrillic)
		if err != nil {
			return err
		}
		ruRune, err := runeFor(ruLevels[i], keyvals, cyrillic)
		if err != nil {
			return err
		}
		if prev, dup := enToRu[enRune]; dup && prev != ruRune {
			return fmt.Errorf("ENToRU[%q]: %w: %q and %q", enRune, errConflictingEntry, prev, ruRune)
		}
		if prev, dup := ruToEn[ruRune]; dup && prev != enRune {
			return fmt.Errorf("RUToEN[%q]: %w: %q and %q", ruRune, errConflictingEntry, prev, enRune)
		}
		enToRu[enRune] = ruRune
		ruToEn[ruRune] = enRune
	}

	return nil
}

// runeFor resolves a keysym name to its rune: Latin-1 values map directly,
// everything else goes through the Cyrillic table. Unresolved names are hard
// errors so a layout change fails generation, not runtime.
func runeFor(name string, keyvals map[string]uint32, cyrillic map[string]rune) (rune, error) {
	val, ok := keyvals[name]
	if !ok {
		return 0, fmt.Errorf("%q: %w in %s", name, errUnknownKeysym, keysymsHeader)
	}
	if val >= latin1Min && val <= latin1Max {
		return rune(val), nil
	}
	r, ok := cyrillic[name]
	if !ok {
		return 0, fmt.Errorf("%q (0x%04x): %w", name, val, errNoRuneMapping)
	}

	return r, nil
}

const fileHeader = `// Code generated by goswitch/layouts/generator. DO NOT EDIT.

//go:generate go run ./generator

// Package layouts holds the generated EN↔RU key-position tables used to
// correct text typed in the wrong layout (CORR-08).
package layouts

`

const enToRuDoc = `// ENToRU maps a character typed on the us(basic) XKB layout to the character
// at the same key position of the effective ru layout (winkeys, the default
// variant on Ubuntu 24.04). Both shift levels are merged into one map; the
// join is by key position, not by character.
`

const ruToEnDoc = `// RUToEN maps a character typed on the effective ru (winkeys) layout to the
// character at the same key position of us(basic). Position-inverse of ENToRU.
`

// emitTables renders the generated Go source, gofmt-clean by construction.
func emitTables(enToRu, ruToEn map[rune]rune) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(fileHeader)
	writeMap(&buf, "ENToRU", enToRuDoc, enToRu)
	writeMap(&buf, "RUToEN", ruToEnDoc, ruToEn)
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated source: %w", err)
	}

	return src, nil
}

// writeMap renders one map declaration with entries sorted by key rune so
// regeneration is byte-for-byte deterministic.
func writeMap(buf *bytes.Buffer, name, doc string, table map[rune]rune) {
	keys := make([]rune, 0, len(table))
	for k := range table {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	buf.WriteString(doc)
	buf.WriteString("var " + name + " = map[rune]rune{\n")
	for _, k := range keys {
		fmt.Fprintf(buf, "\t%s: %s,\n", strconv.QuoteRune(k), strconv.QuoteRune(table[k]))
	}
	buf.WriteString("}\n\n")
}
