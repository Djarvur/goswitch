package correct

import "unicode"

// Dir is the correction direction of a pure-script token.
type Dir int

// Correction directions.
const (
	// ENtoRU converts Latin letters to their ru-layout key-position twins.
	ENtoRU Dir = iota + 1
	// RUtoEN converts Cyrillic letters back to their us-layout positions.
	RUtoEN
)

// Detect classifies the token by its letter script (CORR-04): Latin-only →
// ENtoRU, Cyrillic-only → RUtoEN. A single foreign letter makes the word
// mixed and refuses it (D-16); digits and punctuation are neutral and never
// make a word mixed (D-15). A token with no letters at all has no
// direction. No dictionaries or frequency heuristics — the script
// composition of the buffer is the only input (SPEC §11).
func Detect(token []rune) (Dir, bool) {
	var en, ru int
	for _, r := range token {
		switch {
		case unicode.Is(unicode.Latin, r):
			en++
		case unicode.Is(unicode.Cyrillic, r):
			ru++
		}
	}
	switch {
	case en > 0 && ru > 0:
		return 0, false // mixed — silent refusal (D-16)
	case en > 0:
		return ENtoRU, true
	case ru > 0:
		return RUtoEN, true
	}

	return 0, false // no letters — no direction
}
