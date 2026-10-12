package correct

import "slices"

// ConvertibleStretch returns the correction target of the word path as a
// fresh slice — the convertible stretch of plan 10-02 (SPEC §4.2): the
// active range the buffer already bounds. The GREEN step of the plan
// extends this range over the adjacent runes that carry a table position;
// the RED scaffold keeps the Phase 2 geometry (the stretch equals the
// token) so the corpus proves the defect before the fix. It returns nil
// when the buffer holds no token at all.
func (b *Buffer) ConvertibleStretch() []rune {
	start, end, ok := b.stretchRange()
	if !ok {
		return nil
	}

	return slices.Clone(b.runes[start:end])
}

// StretchTail returns the boundary runes after the convertible stretch —
// the counterpart of Tail for the stretch geometry. It returns nil when
// there is no token or nothing follows the stretch.
func (b *Buffer) StretchTail() []rune {
	_, end, ok := b.stretchRange()
	if !ok || end == len(b.runes) {
		return nil
	}

	return slices.Clone(b.runes[end:])
}

// ReplaceStretch swaps the convertible stretch range for repl and keeps
// what follows it — the replace half of the stretch geometry: the swap must
// cover the whole run, or the typed punctuation would survive beside the
// corrected word. recompute() rederives the token coordinates from the
// replaced contents, so the buffer keeps mirroring the field and a repeated
// correction converts the run back (the toggle invariant).
func (b *Buffer) ReplaceStretch(repl []rune) {
	start, end, ok := b.stretchRange()
	if !ok {
		return
	}
	next := make([]rune, 0, len(b.runes)-(end-start)+len(repl))
	next = append(next, b.runes[:start]...)
	next = append(next, repl...)
	next = append(next, b.runes[end:]...)
	b.runes = next
	b.recompute()
}

// stretchRange returns the [start, end) rune range of the convertible
// stretch — the RED scaffold binds it to the active range; GREEN widens it
// in both directions over runes with a table position.
func (b *Buffer) stretchRange() (start, end int, ok bool) {
	return b.activeRange()
}
