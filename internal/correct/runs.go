package correct

import (
	"unicode"

	"github.com/Djarvur/goswitch/internal/layouts"
)

// scriptKind classifies a rune for the script scan of a correction range:
// Latin and Cyrillic letters carry a conversion direction of their own;
// every other rune — digits, shared punctuation, spaces, letters of other
// scripts — is neutral: it has no direction to invert to and passes
// through the conversion unchanged (spec-delta 2026-10-06).
type scriptKind int

// Script classes of the rune scan.
const (
	scriptNeutral scriptKind = iota
	scriptLatin
	scriptCyrillic
)

// classifyRune maps one rune to its script class — the same Latin/Cyrillic
// test Detect uses (direction.go), factored for the range scanner.
func classifyRune(r rune) scriptKind {
	switch {
	case unicode.Is(unicode.Latin, r):
		return scriptLatin
	case unicode.Is(unicode.Cyrillic, r):
		return scriptCyrillic
	default:
		return scriptNeutral
	}
}

// dirOf is the conversion direction that leaves the given script: a Latin
// range converts through ENtoRU, a Cyrillic range through RUtoEN.
func dirOf(kind scriptKind) Dir {
	if kind == scriptCyrillic {
		return RUtoEN
	}

	return ENtoRU
}

// ConvertRuns converts a correction range by the script composition of its
// letters (CORR-06) — the ONE conversion entry of the daemon for every
// range kind (word, phrase, selection):
//
//   - No Latin or Cyrillic letters at all: ok=false — nothing to convert,
//     the actor takes the D-20 silent refusal.
//   - Letters of both scripts (mixed): PER-CHARACTER layout inversion —
//     every Latin letter maps through ENToRU, every Cyrillic letter
//     through RUToEN, each independently of position (owner verdict, UAT
//     of phase 6, 2026-10-03, commit 7dd46e9; spec-delta 2026-10-06 —
//     supersedes the D-22/D-23 last-letter semantics, whose WINDOWS-12
//     readback freeze was rejected by the owner). The neutral rule is
//     explicit: a rune whose inversion equals itself passes unchanged —
//     digits ('5' inverts to '5' in both tables) — as does a rune with no
//     key position (space); a table-mapped punctuation rune in a MIXED
//     range also rides as typed, because a per-character direction is
//     undefined for it (',' is a source key of BOTH tables with different
//     images). A letter missing from its script's table fails the WHOLE
//     range before any rune is emitted (T-03-01-03: no partial conversion
//     ever leaks).
//   - Letters of one script only: the range converts WHOLESALE to the
//     other layout through Convert — punctuation converts with the token
//     (','→'б'), table-missing non-letters ride, register is preserved
//     per rune (CORR-05). The spec-delta leaves this branch
//     byte-identical.
//
// A mixed range always returns ok=true, changed=true — every mapped
// letter inverts into the other script. changed=false with ok=true is the
// D-24 success-without-changes outcome (a successful operation, never a
// refusal or a WARN); under inversion and wholesale conversion no letter
// maps to itself, so ConvertRuns never produces that outcome — the
// actor's branch stays as the defensive no-op surface.
func ConvertRuns(text []rune) (out []rune, changed, ok bool) {
	var hasLatin, hasCyrillic bool
	script := scriptNeutral
	for _, r := range text {
		switch classifyRune(r) {
		case scriptLatin:
			hasLatin = true
			script = scriptLatin
		case scriptCyrillic:
			hasCyrillic = true
			script = scriptCyrillic
		case scriptNeutral:
			// a neutral rune carries no direction of its own
			// (spec-delta 2026-10-06)
		}
	}
	switch {
	case !hasLatin && !hasCyrillic:
		return nil, false, false // no letters — D-20 refusal
	case hasLatin && hasCyrillic:
		return invertPerChar(text)
	default:
		converted, cok := Convert(text, dirOf(script))
		if !cok {
			return nil, false, false // unmapped letter — wholesale refusal
		}

		return converted, true, true
	}
}

// invertPerChar implements the mixed-text semantics (spec-delta
// 2026-10-06): every Latin letter maps through ENToRU, every Cyrillic
// letter through RUToEN, each on its own; a neutral rune — digits, space,
// punctuation (a per-character direction is undefined for it: ',' is a
// source key of BOTH tables) — passes through unchanged. A letter missing
// from its script's table fails the WHOLE range with a nil output before
// any rune is emitted (T-03-01-03: no partial conversion ever leaks). A
// mixed range always converts: every mapped letter inverts into the other
// script, so the pass reports ok=true, changed=true.
func invertPerChar(text []rune) (out []rune, changed, ok bool) {
	result := make([]rune, 0, len(text))
	for _, r := range text {
		var table map[rune]rune
		switch classifyRune(r) {
		case scriptLatin:
			table = layouts.ENToRU
		case scriptCyrillic:
			table = layouts.RUToEN
		case scriptNeutral:
			result = append(result, r)

			continue
		}
		mapped, mappedOK := table[r]
		if !mappedOK {
			return nil, false, false // unmapped letter — the WHOLE range refuses
		}
		result = append(result, mapped)
	}

	return result, true, true
}
