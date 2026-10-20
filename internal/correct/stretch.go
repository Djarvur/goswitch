package correct

import (
	"slices"

	"github.com/Djarvur/goswitch/internal/layouts"
)

// tablePosition reports whether r carries a key position in either layout
// table — the membership predicate of the convertible stretch (SPEC §4.2,
// plan 10-02). It is deliberately coarser than letterCapable (D-14): the
// stretch reaches over every rune the tables can carry — letters, digits,
// and the punctuation mapped by key position («?» sits on the key whose RU
// image is «,»), while a rune without a position (the space of D-13, «é» in
// the Latin direction) stops it.
func tablePosition(r rune) bool {
	_, en := layouts.ENToRU[r]
	_, ru := layouts.RUToEN[r]

	return en || ru
}

// ConvertibleStretch returns the correction target of the word path as a
// fresh slice — the convertible stretch of plan 10-02 (SPEC §4.2): the
// active range (D-13) extended in BOTH directions over the adjacent runes
// that carry a table position. A mono-layout run converts whole —
// «ytn?» into «нет,», «(ytn» into «(нет» — while a rune without a position
// stops the run (the space: «ghbdtn » keeps its tail). The geometry of the
// feed changes, never the converter: ConvertRuns receives the stretch and
// keeps its semantics byte-for-byte (the mixed run still rides its
// punctuation as typed, spec-delta 2026-10-06). It returns nil when the
// buffer holds no token at all. The autocorrect payload keeps consuming the
// bare Token() (D-14/D-15 unchanged) — the stretch is the manual word
// path's range only.
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
// stretch: the active range walked outward in both directions while the
// neighboring rune carries a table position — the tail side (the living
// case «ytn?») and the head side («(ytn») alike.
func (b *Buffer) stretchRange() (start, end int, ok bool) {
	start, end, ok = b.activeRange()
	if !ok {
		return 0, 0, false
	}
	for start > 0 && tablePosition(b.runes[start-1]) {
		start--
	}
	for end < len(b.runes) && tablePosition(b.runes[end]) {
		end++
	}

	return start, end, true
}
