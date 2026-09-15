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
// anchor in either order — the client reports the pair from its own
// selection direction, so right-to-left and left-to-right describe the SAME
// range — active exactly when the positions differ (no selection
// observable → the caller keeps the word path, Phase 2 behavior).
func SelectionRange(cursorPos, anchorPos uint32) (start, end uint32, active bool) {
	if cursorPos == anchorPos {
		return 0, 0, false
	}
	start, end = cursorPos, anchorPos
	if start > end {
		start, end = end, start
	}

	return start, end, true
}

// VerifyRangeAt is the range-anchored verify rule of the selection
// correction (Pitfall 6): the converted text must sit AT the selection
// position — MatchesSuffix only checks the prefix before the cursor, which a
// selection right of the cursor escapes. Out-of-range positions and an
// overlong want are mismatches, never panics.
func VerifyRangeAt(text []rune, start uint32, want []rune) bool {
	// #nosec G115 -- start indexes a real input field, far below 2^31 runes;
	// on the 64-bit target an int always holds a uint32.
	if int(start) >= len(text) {
		return false
	}
	if len(text)-int(start) < len(want) {
		return false
	}

	return slices.Equal(text[int(start):int(start)+len(want)], want)
}
