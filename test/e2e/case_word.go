package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Word-correction timing and corpus constants (plan 02-03). The corpus pair
// is the SPEC example word; the settle gate proves the client APPLIED the
// replacement before the Enter-oracle closes the entry.
const (
	correctionWait = 5 * time.Second
	wordProbeEN    = "ghbdtn"
	wordResultRU   = "привет"
)

// runWordENRU proves the core-value tracer (CORR-01, CORR-07 level 1):
// with goswitch-en active, "ghbdtn" typed into the zenity entry and a
// double Right Shift, the daemon corrects the last word through the whole
// pipeline — buffer → direction → RequireSurroundingText → suffix
// verification → DeleteSurroundingText + CommitText — and the entry ends
// up holding "привет", verified by both the live field content and the
// zenity stdout oracle.
func runWordENRU(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("word-en-ru needs the zenity entry surface (locked-session fallback engaged?)")
	}

	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"key"`, minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("word-en-ru key visibility: %w", err)
	}

	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"action","n":2`, decisionWait); err != nil {
		return fmt.Errorf("word-en-ru double-tap decision: %w", err)
	}
	if err := s.waitForLog(ctx, `"msg":"correction","outcome":"done"`, correctionWait); err != nil {
		return fmt.Errorf("word-en-ru correction: %w", err)
	}

	if err := s.waitZenityChars(ctx, len([]rune(wordResultRU))); err != nil {
		return fmt.Errorf("word-en-ru applied correction: %w", err)
	}
	out, err := s.closeZenity(ctx)
	if err != nil {
		return err
	}
	if out != wordResultRU {
		return fmt.Errorf("word-en-ru oracle: entry printed %q, want %q", out, wordResultRU)
	}

	return nil
}

// waitZenityChars polls the witness until the zenity entry holds exactly
// wantChars characters — the settle gate proving the client APPLIED the
// correction before the case closes the entry: the correction-done log
// record fires when the daemon emits, the field updates a round-trip
// later, and a client that ignored DeleteSurroundingText would hold the
// doubled text (12 characters, not 6 — Pitfall 3, assumption A1).
func (s *stand) waitZenityChars(ctx context.Context, wantChars int) error {
	want := fmt.Sprintf("zenity:TEXT:chars=%d", wantChars)
	deadline := time.Now().Add(witnessWait)
	for {
		now, err := s.focusWitness(ctx)
		if err != nil {
			return err
		}
		if now == want {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("entry did not settle to %s (witness %q)", want, now)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}
