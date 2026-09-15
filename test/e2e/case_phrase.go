package main

import (
	"context"
	"errors"
	"fmt"
)

// Phrase-correction corpus constants (plan 03-01). The tracer phrase is two
// SPEC words and the separator between them — the smallest phrase that is
// MORE than a token, so the case proves the range really spans the whole
// buffer (D-25), not just the last word.
const (
	phraseProbeEN = wordProbeEN + " " + wordProbeEN   // "ghbdtn ghbdtn"
	phraseResult  = wordResultRU + " " + wordResultRU // "привет привет"
	// phraseMixedResult is the D-26 oracle: a phrase with words in both
	// layouts converts by the SAME run semantics as a mixed word — the EN
	// word typed in EN mode converts, the RU-committed word stays.
	phraseMixedResult = phraseResult // "привет привет"
)

// runPhraseENRU proves the phrase tracer live (CORR-02, D-25, plan 03-01):
// with goswitch-en active, a two-word phrase typed into the zenity entry and
// a TRIPLE Right Shift, the daemon corrects the WHOLE phrase through the
// same pipeline as the word — buffer range → RequireSurroundingText →
// suffix verification over the full phrase → DeleteSurroundingText +
// CommitText — and the entry ends up holding "привет привет", verified by
// the character-count settle gate and the zenity stdout oracle.
func runPhraseENRU(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("phrase-en-ru needs the zenity entry surface (locked-session fallback engaged?)")
	}

	if err := s.injectText(ctx, phraseProbeEN); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"key"`, minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("phrase-en-ru key visibility: %w", err)
	}

	if err := s.injectKeys(ctx, "Shift_R", "Shift_R", "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"action","n":3`, decisionWait); err != nil {
		return fmt.Errorf("phrase-en-ru triple-tap decision: %w", err)
	}
	if err := s.waitForLog(ctx, `"msg":"correction","outcome":"done"`, correctionWait); err != nil {
		return fmt.Errorf("phrase-en-ru correction: %w", err)
	}

	if err := s.waitZenityChars(ctx, len([]rune(phraseResult))); err != nil {
		return fmt.Errorf("phrase-en-ru applied correction: %w", err)
	}
	out, err := s.closeZenity(ctx)
	if err != nil {
		return err
	}
	if out != phraseResult {
		return fmt.Errorf("phrase-en-ru oracle: entry printed %q, want %q", out, phraseResult)
	}

	return nil
}

// runPhraseMixed proves the D-26 phrase semantics live (plan 03-01 task 2):
// "ghbdtn" typed in EN mode transits, the single-tap flip switches the
// engine to RU, the second word — typed through the same physical Latin
// path — arrives as engine-committed «привет», so the field holds the mixed
// phrase "ghbdtn привет". The TRIPLE tap corrects it by the SAME run
// semantics as the mixed word: the anchor is the last letter of the phrase
// (RU), the foreign EN word converts, the Russian word stays — the entry
// ends up holding "привет привет".
func runPhraseMixed(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("phrase-mixed needs the zenity entry surface (locked-session fallback engaged?)")
	}

	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"key"`, minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("phrase-mixed key visibility: %w", err)
	}
	// The separator rides the typing path (ydotool "space" resolves to the
	// physical S key — the 02-03 live trap).
	if err := s.injectText(ctx, " "); err != nil {
		return err
	}

	// Flip EN → RU and wait for the mode record before typing (Pitfall 5):
	// the RU part arrives as engine commits of the same physical Latin keys.
	if err := s.flipToRU(ctx, "phrase-mixed"); err != nil {
		return err
	}
	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return err
	}
	if err := s.waitZenityText(ctx, wordProbeEN+" "+wordResultRU); err != nil {
		return fmt.Errorf("phrase-mixed mixed phrase assembled: %w", err)
	}

	if err := tripleTapCorrects(ctx, s, "phrase-mixed"); err != nil {
		return err
	}

	if err := s.waitZenityText(ctx, phraseMixedResult); err != nil {
		return fmt.Errorf("phrase-mixed applied correction (D-26): %w", err)
	}
	out, err := s.closeZenity(ctx)
	if err != nil {
		return err
	}
	if out != phraseMixedResult {
		return fmt.Errorf("phrase-mixed oracle: entry printed %q, want %q", out, phraseMixedResult)
	}

	return nil
}

// tripleTapCorrects injects the triple Right Shift and gates on both log
// records of a phrase correction — the decision (n=3) and the completion.
func tripleTapCorrects(ctx context.Context, s *stand, caseName string) error {
	if err := s.injectKeys(ctx, "Shift_R", "Shift_R", "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"action","n":3`, decisionWait); err != nil {
		return fmt.Errorf("%s triple-tap decision: %w", caseName, err)
	}
	if err := s.waitForLog(ctx, `"msg":"correction","outcome":"done"`, correctionWait); err != nil {
		return fmt.Errorf("%s correction: %w", caseName, err)
	}

	return nil
}
