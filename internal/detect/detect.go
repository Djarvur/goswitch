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
	"sort"
	"strings"
	"unicode"

	"github.com/Djarvur/goswitch/internal/correct"
)

// Input modes: the layout the token was typed in. Check takes the mode as
// a string because the actor's mode record (D-34) is the caller's state.
const (
	modeEN = "en"
	modeRU = "ru"
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
	if len(tok) < p.MinWordLen {
		return Verdict{Reason: ReasonAbstainShort}
	}

	// An unknown mode has no target side at all and falls into the same
	// abstention as a letterless token (the closed-input pin).
	dir, known := dirOf(mode)
	if !known || !hasScriptLetters(tok, mode) {
		return Verdict{Reason: ReasonAbstainNoLetters}
	}

	// Dictionary path (D-52а): both memberships first, then one verdict —
	// a current hit vetoes BEFORE the other side is even considered, and a
	// word valid in both layouts (a random anagram) is nobody's
	// wrong-layout.
	inCur := inDict(dictOf(d, mode), tok)
	conv, convOK := correct.Convert(tok, dir)
	inOther := convOK && inDict(dictOf(d, otherMode(mode)), conv)

	switch {
	case inCur && inOther:
		return Verdict{Reason: ReasonDictBothHit}
	case inCur:
		return Verdict{Reason: ReasonDictCurHit}
	case inOther:
		return Verdict{
			WrongLayout: true,
			Confident:   true,
			Dir:         dir,
			Reason:      ReasonDictWrongLayout,
		}
	case !convOK:
		return Verdict{Reason: ReasonAbstainConvertFail}
	}

	// Neither dictionary answered (D-52б): the trigram fallback scores the
	// plausibility of the token in the current layout against the remap in
	// the other one.
	if trigramWrongLayout(tok, conv, mode, t, p) {
		return Verdict{
			WrongLayout: true,
			Confident:   true,
			Dir:         dir,
			Reason:      ReasonTrigramWrongLayout,
		}
	}

	return Verdict{Reason: ReasonTrigramUnsure}
}

// dirOf resolves the mode to its correction direction: en-typed tokens
// convert ENtoRU, ru-typed ones RUtoEN.
func dirOf(mode string) (correct.Dir, bool) {
	switch mode {
	case modeEN:
		return correct.ENtoRU, true
	case modeRU:
		return correct.RUtoEN, true
	}

	return 0, false
}

// otherMode returns the opposite layout of a known mode.
func otherMode(mode string) string {
	if mode == modeEN {
		return modeRU
	}

	return modeEN
}

// dictOf picks the dictionary of the mode's own layout.
func dictOf(d Data, mode string) []string {
	if mode == modeEN {
		return d.EN
	}

	return d.RU
}

// triOf picks the trigram table of the mode's own language.
func triOf(t Trigrams, mode string) map[string]float64 {
	if mode == modeEN {
		return t.EN
	}

	return t.RU
}

// hasScriptLetters reports whether the token carries at least one letter
// of the mode's script: an en-typed token needs a latin letter, a ru-typed
// one a cyrillic letter. Digits and punctuation have no script of their
// own; an unknown mode matches nothing.
func hasScriptLetters(tok []rune, mode string) bool {
	for _, r := range tok {
		switch {
		case mode == modeEN && unicode.Is(unicode.Latin, r):
			return true
		case mode == modeRU && unicode.Is(unicode.Cyrillic, r):
			return true
		}
	}

	return false
}

// inDict reports a dictionary hit for either the exact spelling or the
// normalized one (normalizeLookup): the exact probe keeps the capitalized
// entries (АЗС, Москва) suppressive, the normalized probe lets a register
// or yo variant of a lowercase entry (ПРИВЕТ, ёлка) still find it — a miss
// here would feed the fallback and erode the veto that makes full
// dictionary coverage safety (Pitfall 3).
func inDict(dict []string, word []rune) bool {
	s := string(word)
	if has(dict, s) {
		return true
	}

	norm := normalizeLookup(s)

	return norm != s && has(dict, norm)
}

// has binary-searches one sorted dictionary — the only lookup mechanism
// (research Q3: ~17 comparisons over 146k entries, zero heap).
func has(dict []string, w string) bool {
	i := sort.SearchStrings(dict, w)

	return i < len(dict) && dict[i] == w
}

// normalizeLookup folds a query the same way the 06-02 dictgen folded the
// data: lower case for register-free matching, ё→е for the yo-less ru
// dictionary (Pitfall 3).
func normalizeLookup(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), "ё", "е")
}

// trigramWrongLayout is the D-52(б) fallback verdict for the words both
// dictionaries miss: the other-language profile of the remap must beat the
// current-language profile of the token by more than the margin AND be
// absolutely plausible above the floor — both thresholds come from Params,
// no literals live in this branch.
func trigramWrongLayout(tok, conv []rune, mode string, t Trigrams, p Params) bool {
	curScore := scoreLang(tok, triOf(t, mode))
	otherScore := scoreLang(conv, triOf(t, otherMode(mode)))

	return otherScore-curScore > p.TrigramMargin && otherScore > p.TrigramFloor
}

// neutralPenalty is the log10 probability of a trigram the training
// dictionaries never saw (1e-6): below every baked entry (the tables floor
// near -4.7), so an unknown window penalizes more than any known one while
// staying finite. The golden corpus pins the value through behavior.
const neutralPenalty = -6.0

// trigramLen is the n of the n-gram model: the detector scores 3-rune
// windows, the tables of plan 06-02 are trained over the same shape.
const trigramLen = 3

// scoreLang averages the log10 trigram probabilities of the word under one
// language table: Σ log P over the 3-rune windows, normalized by the
// window count so doubling the word never changes the score. Windows
// absent from the table contribute neutralPenalty; a word shorter than one
// window has no evidence and scores 0 — never above a log-scale floor, so
// it can never become a confident verdict on its own.
func scoreLang(word []rune, table map[string]float64) float64 {
	windows := len(word) - trigramLen + 1
	if windows <= 0 {
		return 0
	}

	var sum float64
	for i := range windows {
		if p, ok := table[string(word[i:i+trigramLen])]; ok {
			sum += p

			continue
		}
		sum += neutralPenalty
	}

	return sum / float64(windows)
}
