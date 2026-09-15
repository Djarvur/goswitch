package correct

import "slices"

// MatchesSuffix implements the mandatory pre-correction verification of
// ADR-004: the text before the cursor must END with the token — any tail
// after the token in the surrounding text is a mismatch, and the correction
// aborts ("abort, не мусорить") rather than edit at a guess.
func MatchesSuffix(textBeforeCursor, token []rune) bool {
	n := len(token)
	if len(textBeforeCursor) < n {
		return false
	}

	return slices.Equal(textBeforeCursor[len(textBeforeCursor)-n:], token)
}
