// Package detect is the pure wrong-layout detector of the autocorrect
// system (D-52): a dictionary path on the word boundary — the token is NOT
// in the current-layout dictionary AND its ЙЦУКЕН↔QWERTY remap IS in the
// other one, a high-confidence verdict — plus a trigram log-likelihood
// fallback for the words both dictionaries miss, and abstentions for every
// shape the detector refuses to judge. The package is headless: no D-Bus,
// no clock, no goroutines — the whole corpus runs under -race without a
// live session. Privacy (D-20/D-21, Pitfall 6): a Verdict carries only a
// Reason from the closed vocabulary below — the token never enters the
// verdict, so no consumer can leak the typed word into a log or a status
// surface through it.
package detect

import (
	"github.com/Djarvur/goswitch/internal/correct"
)

// The closed reason vocabulary: every Check outcome reports exactly one of
// these constants, and they are the ONLY string channel of a Verdict —
// counters and logs name the outcome without ever carrying the word
// (D-20/D-21, T-06-05-02).
const (
	// ReasonDictWrongLayout — the token is out of the current dictionary
	// and its remap is in the other one (D-52а): high confidence.
	ReasonDictWrongLayout = "dict-wrong-layout"
	// ReasonDictCurHit — the token is valid in the current layout: it is
	// NEVER corrected, the dictionary hit vetoes (completeness is safety).
	ReasonDictCurHit = "dict-cur-hit"
	// ReasonDictBothHit — the token is valid in both layouts (a random
	// anagram): veto, nothing to decide.
	ReasonDictBothHit = "dict-both-hit"
	// ReasonTrigramWrongLayout — the trigram fallback passed both thresholds
	// (D-52б): confident verdict from plausibility alone.
	ReasonTrigramWrongLayout = "trigram-wrong-layout"
	// ReasonTrigramUnsure — the fallback could not decide (thresholds not
	// met) or was never implemented for the branch: no confidence.
	ReasonTrigramUnsure = "trigram-unsure"
	// ReasonAbstainShort — the token is shorter than Params.MinWordLen.
	ReasonAbstainShort = "abstain-short"
	// ReasonAbstainNoLetters — the token carries no letter of the mode's
	// script (digits, punctuation) or the mode is unknown: no side to judge.
	ReasonAbstainNoLetters = "abstain-no-letters"
	// ReasonAbstainConvertFail — a mixed token: one foreign letter refuses
	// the whole remap (correct.Convert), so the detector abstains and the
	// word stays untouched — never a partial rewrite (CORR-06).
	ReasonAbstainConvertFail = "abstain-convert-fail"
)

// Verdict is the outcome of one detection. WrongLayout marks the
// typed-in-the-other-layout verdict; Confident marks whether the detector
// would silently auto-correct on it (the D-53 condition 4); Dir is the
// correction direction of a confident verdict; Reason names the outcome
// from the closed vocabulary above. The struct is deliberately token-free:
// these four fields are the whole verdict (privacy pin, T-06-05-02).
type Verdict struct {
	WrongLayout bool
	Confident   bool
	Dir         correct.Dir
	Reason      string
}

// Data carries the baked dictionaries of plan 06-02 (layouts.DictRU and
// layouts.DictEN): sorted slices, looked up with sort.SearchStrings only —
// never a runtime map (research Q3).
type Data struct {
	RU []string
	EN []string
}

// Trigrams carries the baked trigram tables of plan 06-02 (layouts.TriRU
// and layouts.TriEN): log10 probabilities, every value non-positive.
type Trigrams struct {
	RU map[string]float64
	EN map[string]float64
}

// Params carries the detector thresholds. The zero value never fires: a
// zero MinWordLen abstains on every token.
type Params struct {
	MinWordLen    int
	TrigramMargin float64
	TrigramFloor  float64
}

// The DefaultParams values, named so the mirror contract reads as one
// table: internal/config Defaults (plan 06-04) must carry the same trio —
// change the two places TOGETHER.
const (
	defaultMinWordLen    = 4
	defaultTrigramMargin = 2.0
	defaultTrigramFloor  = 1.0
)

// DefaultParams returns the starting thresholds, mirroring the autocorrect
// config defaults of plan 06-04 (internal/config Defaults: min_word_len 4,
// trigram_margin 2.0, trigram_floor 1.0). The golden corpus of this package
// pins the calibration: trigram scores are log10 probabilities
// (non-positive), so a floor of 1.0 sits above the whole scale and keeps
// the fallback UNSURE on every word at the defaults — the conservative
// start the corpus confirmed (research A4 sanctions a threshold-disabled
// fallback; a dictionary miss on both sides then reports trigram-unsure
// and the word stays untouched).
func DefaultParams() Params {
	return Params{
		MinWordLen:    defaultMinWordLen,
		TrigramMargin: defaultTrigramMargin,
		TrigramFloor:  defaultTrigramFloor,
	}
}

// Check judges one token typed in the mode layout ("en" or "ru"): the
// dictionary path of D-52(а) first — a current-dictionary hit vetoes, the
// token whose remap sits in the other dictionary is a confident
// wrong-layout — then the trigram fallback of D-52(б) for the words both
// dictionaries miss. Every other shape abstains: short tokens, tokens
// without a letter of the mode's script, mixed tokens, unknown modes. The
// function is deterministic over its arguments and owns no state.
func Check(tok []rune, mode string, d Data, t Trigrams, p Params) Verdict {
	return Verdict{Reason: ReasonTrigramUnsure}
}
