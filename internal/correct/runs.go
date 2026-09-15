package correct

import "unicode"

// scriptKind classifies a rune for the run segmentation of a correction
// range: Latin and Cyrillic letters form runs; every other rune — digits,
// shared punctuation, spaces, letters of other scripts — is neutral (D-15
// continuation: neutrals build no segment and shift no anchor).
type scriptKind int

// Script classes of the run scan.
const (
	scriptNeutral scriptKind = iota
	scriptLatin
	scriptCyrillic
)

// classifyRune maps one rune to its script class — the same Latin/Cyrillic
// test Detect uses (direction.go), factored for the run scanner.
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
// run converts through ENtoRU, a Cyrillic run through RUtoEN.
func dirOf(kind scriptKind) Dir {
	if kind == scriptCyrillic {
		return RUtoEN
	}

	return ENtoRU
}

// ConvertRuns converts a correction range by the script composition of its
// letters (CORR-06, D-22/D-23/D-24) — the ONE conversion entry of the
// daemon for every range kind (word, phrase, later selection):
//
//   - No Latin or Cyrillic letters at all: ok=false — nothing to anchor on,
//     the actor takes the D-20 silent refusal.
//   - Letters of one script only: the range converts WHOLESALE to the other
//     layout (уточнение D-22, 2026-09-15: the anchor governs only mixed
//     text — the Phase 2 composition behavior keeps matrix v1 green).
//   - Letters of both scripts: the anchor is the script of the LAST letter
//     (digits and other neutrals shift it, D-15/D-22); every maximal run of
//     foreign-script letters converts independently through Convert — own
//     runs stay as typed, neutrals ride along unchanged.
//
// changed=false with ok=true is the D-24 success-without-changes outcome
// ("все буквы уже в якорной раскладке" — a successful operation, never a
// refusal or a WARN). Under the уточнение the word and phrase ranges of
// this plan always resolve to a conversion; that surface serves ranges
// anchored from outside their own composition (the selection path of plan
// 03-03). An unmapped letter in a run that must convert fails the WHOLE
// range (T-03-01-03: no partial conversion ever leaks — Convert is reused
// as is, never forked).
func ConvertRuns(text []rune) (out []rune, changed, ok bool) {
	var hasLatin, hasCyrillic bool
	anchor := scriptNeutral
	for _, r := range text {
		switch classifyRune(r) {
		case scriptLatin:
			hasLatin = true
			anchor = scriptLatin
		case scriptCyrillic:
			hasCyrillic = true
			anchor = scriptCyrillic
		case scriptNeutral:
			// neutrals build no segment and shift no anchor (D-15/D-22)
		}
	}
	switch {
	case !hasLatin && !hasCyrillic:
		return nil, false, false // no letters — D-20 refusal
	case hasLatin && hasCyrillic:
		return convertForeignRuns(text, anchor)
	default:
		converted, cok := Convert(text, dirOf(anchor))
		if !cok {
			return nil, false, false // unmapped letter — wholesale refusal
		}

		return converted, true, true
	}
}

// convertForeignRuns implements the anchor rule over a mixed range
// (D-22/D-23): each maximal run of same-script letters is a segment, runs
// of the anchor script stay as typed, foreign runs convert through Convert
// each on its own, neutral runes pass through unchanged.
func convertForeignRuns(text []rune, anchor scriptKind) (out []rune, changed, ok bool) {
	result := make([]rune, 0, len(text))
	converted := false
	runStart, runScript := -1, scriptNeutral
	closeRun := func(end int) bool {
		if runStart == -1 {
			return true
		}
		run := text[runStart:end]
		if runScript == anchor {
			result = append(result, run...)
		} else {
			c, cok := Convert(run, dirOf(runScript))
			if !cok {
				return false // unmapped letter aborts the whole range
			}
			result = append(result, c...)
			converted = true
		}
		runStart, runScript = -1, scriptNeutral

		return true
	}
	for i, r := range text {
		kind := classifyRune(r)
		if kind != scriptNeutral {
			if runStart != -1 && kind != runScript && !closeRun(i) {
				return nil, false, false
			}
			if runStart == -1 {
				runStart, runScript = i, kind
			}

			continue
		}
		if !closeRun(i) {
			return nil, false, false
		}
		result = append(result, r)
	}
	if !closeRun(len(text)) {
		return nil, false, false
	}

	// A mixed range always holds at least one foreign run, so converted is
	// true here; the D-24 no-op surface is documented on ConvertRuns.
	return result, converted, true
}
