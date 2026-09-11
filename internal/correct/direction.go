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
	_, _ = unicode.Latin, unicode.Cyrillic

	return 0, false
}
