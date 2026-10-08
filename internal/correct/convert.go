package correct

import (
	"unicode"

	"github.com/Djarvur/goswitch/internal/layouts"
)

// Convert maps EVERY table-mapped rune of the token through the
// key-position layout table chosen by dir (CORR-05: register is preserved
// constructively — the generated tables carry both shift levels per
// position, so 'g'→'п' and 'G'→'П'). Full by-position conversion (owner
// directive 2026-09-28, superseding the D-14 ride-along and the 2026-09-27
// allow-set): punctuation converts with the word — the comma is 'б' and
// the question mark is the RU comma's key ("hf,jnftn"→"работает") —
// making the mapping a bijection: the same physical keys typed in the
// other mode yield exactly the converted text. A non-letter MISSING from
// the table (space, tab) rides through identically; a letter missing from
// the table fails the whole conversion.
func Convert(token []rune, dir Dir) ([]rune, bool) {
	var table map[rune]rune
	switch dir {
	case ENtoRU:
		table = layouts.ENToRU
	case RUtoEN:
		table = layouts.RUToEN
	default:
		return nil, false
	}
	out := make([]rune, 0, len(token))
	for _, r := range token {
		if !unicode.IsLetter(r) {
			// A non-letter converts when the table carries its key position
			// (','→'б', '?'→','), rides along when it does not (space, tab).
			if m, ok := table[r]; ok {
				out = append(out, m)

				continue
			}
			out = append(out, r) // table-missing non-letters ride along identically

			continue
		}
		m, ok := table[r]
		if !ok {
			return nil, false
		}
		out = append(out, m)
	}

	return out, true
}
