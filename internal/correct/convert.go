package correct

import (
	"unicode"

	"github.com/Djarvur/goswitch/layouts"
)

// bracketRowAllowSet returns the explicit allow-set of the EN bracket-row
// symbols that convert during correction (SPEC §4.2 «включая знаки», owner
// decision 1, revised 2026-09-27): '[', ']', ';', the apostrophe and '`'
// are the EN symbols that never serve as Russian prose punctuation in the
// user's intent — typed while intending Russian they mean х ъ ж э ё (the
// live defect: «несколько дополнительны[» — the word converted, the
// trailing RU 'х' key stayed). ',' and '.' are deliberately EXCLUDED —
// universal Russian prose punctuation: "ghbdtn," intends the separator
// (D-14), so the comma and the dot keep riding along unchanged, as does
// every other non-letter ('@'-class, digits, space, slash). The set applies
// in the EN→RU direction only: a ';' in a Russian-typed text is the RU
// keyboard's own punctuation (Shift+4) and rides in RU→EN, exactly like
// '"'. The owner may extend the set later. A function, not a package-level
// var: the strict lint forbids mutable globals (the altModifierCandidates/
// matrix.go idiom).
func bracketRowAllowSet() map[rune]bool {
	return map[rune]bool{
		'[':  true,
		']':  true,
		';':  true,
		'\'': true,
		'`':  true,
	}
}

// Convert maps every letter of the token through the key-position layout
// table chosen by dir (CORR-05: register is preserved constructively — the
// generated tables carry both shift levels per position, so 'g'→'п' and
// 'G'→'П'). The allow-set bracket-row symbols (bracketRowAllowSet) convert
// together with the letters in the EN→RU direction (SPEC §4.2 «включая
// знаки», owner decision 1 revised 2026-09-27); every other non-letter of
// the token rides through identically (D-14/D-15: "ghbdtn,"→"привет,"); a
// letter missing from the table fails the whole conversion.
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
	allow := bracketRowAllowSet()
	out := make([]rune, 0, len(token))
	for _, r := range token {
		if !unicode.IsLetter(r) {
			// A non-letter converts only when it is in the EN→RU allow-set
			// and the direction's table carries it; every other non-letter
			// rides along identically (the D-14 comma, the '.' and the whole
			// '@'-class — owner decision 1, 260927-vu8).
			if dir == ENtoRU && allow[r] {
				if m, ok := table[r]; ok {
					out = append(out, m)

					continue
				}
			}
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
