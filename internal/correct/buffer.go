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
	"unicode"

	"github.com/Djarvur/goswitch/layouts"
)

// Buffer mirrors the text before the cursor as a rune phrase with token
// boundaries (D-13/D-14/D-15): the buffer keeps everything typed since the
// last hard reset, a separator finishes the current token without evicting
// it from the buffer (D-13 — the last word stays correctable after a
// space), and token membership follows the letter-capable key union of both
// layouts plus digits (D-14/D-15).
type Buffer struct{}

// NewBuffer returns an empty Buffer.
func NewBuffer() *Buffer { return &Buffer{} }

// Push feeds one typed rune: token-capable runes extend the current token,
// any other rune finishes the token and lands in the buffer as a boundary.
func (b *Buffer) Push(rune) {}

// Backspace pops the last rune — honestly, whatever it was: the token under
// the cursor, the boundary tail, or a finished word; token coordinates are
// rederived from the buffer contents.
func (b *Buffer) Backspace() {}

// HardReset clears the whole phrase (Enter/Tab/Escape/FocusOut/Reset —
// CORR-09 triggers, wired by the actor in later plans).
func (b *Buffer) HardReset() {}

// Token returns the correction target as a fresh slice: the current token
// when it is non-empty, else the last finished one (D-13). It returns nil
// when the buffer holds no token at all.
func (b *Buffer) Token() []rune { return nil }

// Tail returns the boundary runes between the end of Token and the end of
// the buffer — the separator a correction after a space must delete and
// recommit together with the converted token (D-13/D-14, ADR-003 geometry).
// It returns nil when there is no token or no tail.
func (b *Buffer) Tail() []rune { return nil }

// ReplaceToken swaps the Token range for repl and keeps the tail, so the
// buffer keeps mirroring the field after a correction and a repeated
// correction converts the word back (toggle invariant, plan 02-01).
func (b *Buffer) ReplaceToken([]rune) {}

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
