package correct

import (
	"unicode"

	"github.com/Djarvur/goswitch/layouts"
)

// Convert maps every letter of the token through the key-position layout
// table chosen by dir (CORR-05: register is preserved constructively — the
// generated tables carry both shift levels per position, so 'g'→'п' and
// 'G'→'П'). Digits and other non-letter runes of the token pass through
// identically (D-14/D-15: "ghbdtn,"→"привет,"); a letter missing from the
// table fails the whole conversion.
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
			out = append(out, r) // digits and token punctuation ride along identically

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
