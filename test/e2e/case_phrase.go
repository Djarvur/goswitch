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
