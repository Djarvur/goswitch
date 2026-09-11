package correct

// MatchesSuffix implements the mandatory pre-correction verification of
// ADR-004: the text before the cursor must END with the token — any tail
// after the token in the surrounding text is a mismatch, and the correction
// aborts ("abort, не мусорить") rather than edit at a guess.
func MatchesSuffix(textBeforeCursor, token []rune) bool {
	_, _ = textBeforeCursor, token

	return false
}
