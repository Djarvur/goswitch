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

// SelectionRange resolves the wire pair of a surrounding-text push into the
// selection's half-open rune range (D-30): the interval between cursor and
// anchor in either order, active exactly when the positions differ.
//
// RED stub (plan 03-03 task 2): returns the inactive shape so the corpus
// fails on assertions.
func SelectionRange(cursorPos, anchorPos uint32) (start, end uint32, active bool) {
	return 0, 0, false
}

// VerifyRangeAt is the range-anchored verify rule of the selection
// correction (Pitfall 6): the converted text must sit AT the selection
// position — MatchesSuffix only checks the prefix before the cursor, which a
// selection right of the cursor escapes.
//
// RED stub (plan 03-03 task 2): always false so the corpus fails on
// assertions.
func VerifyRangeAt(_ []rune, _ uint32, _ []rune) bool {
	return false
}
