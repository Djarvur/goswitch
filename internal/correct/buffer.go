// Package correct is the pure word-correction library of Phase 2: a
// typed-phrase buffer with token boundaries (D-13/D-14/D-15), the
// script-direction detector with the mixed-word refusal (D-16), the
// case-preserving layout conversion (CORR-05), the mandatory pre-correction
// suffix verification (ADR-004) and the replacement-ladder arithmetic of
// both levels counted in runes (ADR-003). The package is headless: no D-Bus,
// no clock, no goroutines — the whole corpus runs under -race without a
// live session, and the engine wiring arrives in plan 02-03.
package correct

import (
	"slices"
	"unicode"

	"github.com/Djarvur/goswitch/layouts"
)

// Buffer mirrors the text before the cursor as a rune phrase with token
// boundaries (D-13/D-14/D-15): the buffer keeps everything typed since the
// last hard reset, a separator finishes the current token without evicting
// it from the buffer (D-13 — the last word stays correctable after a
// space), and token membership follows the letter-capable key union of both
// layouts plus digits (D-14/D-15). All coordinates and lengths are rune
// indexes, never bytes (ADR-003).
type Buffer struct {
	runes     []rune
	curStart  int // start of the token being built; == len(runes) when none
	lastStart int // start of the last finished token
	lastLen   int // length of the last finished token; 0 when there is none
}

// NewBuffer returns an empty Buffer.
func NewBuffer() *Buffer { return &Buffer{} }

// Push feeds one typed rune: a token-capable rune (letter-capable key or a
// digit) extends the current token, any other rune finishes the current
// token and lands in the buffer as a boundary.
func (b *Buffer) Push(r rune) {
	if tokenCapable(r) {
		b.runes = append(b.runes, r)

		return
	}
	if cur := len(b.runes) - b.curStart; cur > 0 {
		b.lastStart, b.lastLen = b.curStart, cur
	}
	b.runes = append(b.runes, r)
	b.curStart = len(b.runes)
}

// Backspace pops the last rune — honestly, whatever it was: the token under
// the cursor, the boundary tail, or a finished word. Token coordinates are
// rederived from the buffer contents, so the buffer keeps mirroring the
// field.
func (b *Buffer) Backspace() {
	if len(b.runes) == 0 {
		return
	}
	b.runes = b.runes[:len(b.runes)-1]
	b.recompute()
}

// HardReset clears the whole phrase (Enter/Tab/Escape/FocusOut/Reset —
// CORR-09 triggers, wired by the actor in later plans).
func (b *Buffer) HardReset() {
	b.runes = nil
	b.curStart = 0
	b.lastStart = 0
	b.lastLen = 0
}

// Token returns the correction target as a fresh slice: the current token
// when it is non-empty, else the last finished one (D-13). It returns nil
// when the buffer holds no token at all.
func (b *Buffer) Token() []rune {
	start, end, ok := b.activeRange()
	if !ok {
		return nil
	}

	return slices.Clone(b.runes[start:end])
}

// Tail returns the boundary runes between the end of Token and the end of
// the buffer — the separator a correction after a space must delete and
// recommit together with the converted token (D-13/D-14, ADR-003 geometry).
// It returns nil when there is no token or no tail.
func (b *Buffer) Tail() []rune {
	_, end, ok := b.activeRange()
	if !ok || end == len(b.runes) {
		return nil
	}

	return slices.Clone(b.runes[end:])
}

// ReplaceToken swaps the Token range for repl and keeps the tail, so the
// buffer keeps mirroring the field after a correction and a repeated
// correction converts the word back (toggle invariant, plan 02-01).
func (b *Buffer) ReplaceToken(repl []rune) {
	start, end, ok := b.activeRange()
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

// activeRange returns the [start, end) rune range of the correction target:
// the current token when it is non-empty, else the last finished one
// (D-13 — the last word stays correctable after a separator). ok is false
// when the buffer holds no token at all.
func (b *Buffer) activeRange() (start, end int, ok bool) {
	if b.curStart < len(b.runes) {
		return b.curStart, len(b.runes), true
	}
	if b.lastLen > 0 {
		return b.lastStart, b.lastStart + b.lastLen, true
	}

	return 0, 0, false
}

// recompute rederives the token coordinates from the buffer contents after
// a destructive mutation (an honest pop or a token replacement): the
// current token is the maximal token-capable run touching the end, and the
// last finished token is the run before the boundary runes that precede it.
func (b *Buffer) recompute() {
	i := len(b.runes)
	for i > 0 && tokenCapable(b.runes[i-1]) {
		i--
	}
	b.curStart = i
	for i > 0 && !tokenCapable(b.runes[i-1]) {
		i--
	}
	end := i
	for i > 0 && tokenCapable(b.runes[i-1]) {
		i--
	}
	b.lastStart, b.lastLen = i, end-i
}

// letterCapable reports whether r sits on a key that yields a letter in at
// least one of the two layouts (D-14: the union of the EN and RU layers by
// key position — ',' shares its key with RU 'б' and ';' with RU 'ж', while
// '/' and '-' are symbols on both sides and therefore token boundaries).
func letterCapable(r rune) bool {
	if unicode.IsLetter(r) {
		return true
	}
	if m, ok := layouts.ENToRU[r]; ok && unicode.IsLetter(m) {
		return true
	}
	m, ok := layouts.RUToEN[r]

	return ok && unicode.IsLetter(m)
}

// tokenCapable reports whether r extends the current token: a
// letter-capable key, or a digit — digits are neutral token members (D-15).
func tokenCapable(r rune) bool {
	return letterCapable(r) || unicode.IsDigit(r)
}
