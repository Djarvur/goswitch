package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Word-correction timing and corpus constants (plan 02-03). The corpus pair
// is the SPEC example word; the settle gate proves the client APPLIED the
// replacement before the Enter-oracle closes the entry.
const (
	correctionWait = 5 * time.Second
	wordProbeEN    = "ghbdtn"
	wordResultRU   = "привет"
	// wordResultRUAfterSpace is the D-13 expectation: the corrected word
	// WITH its separator — the trailing space is load-bearing.
	wordResultRUAfterSpace = "привет "
	// wordMixedExpected is the D-16 oracle: the mixed word exactly as
	// assembled — the EN part typed in EN mode (transit) plus the RU part
	// committed by the engine after the flip — and left untouched.
	wordMixedExpected = "gfb" + wordResultRU
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

// runWordAfterSpace proves the D-13 geometry on the real desktop — the
// phase's main geometric test (Pitfall 1): a word already separated by a
// space is corrected TOGETHER with the separator, "ghbdtn " → "привет ".
// The deletion must cover token+tail: deleting only the token would leave
// the space, land the commit behind it and produce "gпривет " — the exact
// defect of the owner's prototype (punto_engine.py delete_word). The oracle is
// content-exact AT-SPI readback: the character-count settle gate alone
// cannot tell "привет " from "gпривет " (both 7 runes).
func runWordAfterSpace(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("word-after-space needs the zenity entry surface (locked-session fallback engaged?)")
	}

	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"key"`, minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("word-after-space key visibility: %w", err)
	}
	// The separator arrives through the typing path, not pressKey("space"):
	// ydotool 0.1.8's key-name parser resolves "space" to keycode 31 — the
	// physical S key (live-verified 2026-09-14: the engine saw keyval 0x73,
	// the token grew a trailing s and the correction produced приветЫ).
	if err := s.injectText(ctx, " "); err != nil {
		return err
	}

	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"action","n":2`, decisionWait); err != nil {
		return fmt.Errorf("word-after-space double-tap decision: %w", err)
	}
	if err := s.waitForLog(ctx, `"msg":"correction","outcome":"done"`, correctionWait); err != nil {
		return fmt.Errorf("word-after-space correction: %w", err)
	}

	if err := s.waitZenityChars(ctx, len([]rune(wordResultRUAfterSpace))); err != nil {
		return fmt.Errorf("word-after-space applied correction: %w", err)
	}
	if err := s.waitZenityText(ctx, wordResultRUAfterSpace); err != nil {
		return fmt.Errorf("word-after-space oracle (Pitfall 1 geometry): %w", err)
	}

	out, err := s.closeZenity(ctx)
	if err != nil {
		return err
	}
	// zenity prints the entry text on OK, but closeZenity trims surrounding
	// whitespace of the printed line — the trailing space is pinned by the
	// AT-SPI readback above; stdout pins the word itself.
	if out != wordResultRU {
		return fmt.Errorf("word-after-space stdout oracle: entry printed %q, want %q", out, wordResultRU)
	}

	return nil
}

// flipToRU performs the single-tap script flip and gates on its mode
// record: the Single decision fires at window expiry (~300 ms after the
// tap), so every following step must wait for the record — typing earlier
// would land in EN mode (Pitfall 5).
func (s *stand) flipToRU(ctx context.Context, caseName string) error {
	if err := s.injectKeys(ctx, "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"mode","to":"ru"`, decisionWait); err != nil {
		return fmt.Errorf("%s single-tap flip: %w", caseName, err)
	}

	return nil
}

// runWordRUEN proves the second correction direction end to end (CORR-01,
// D-18, plan 02-04): the daemon starts in its EN mode, a single Right Shift
// flips the engine's script mode to RU at window expiry (the mode log record
// is the sequencing gate — the flip fires ~300 ms after the tap, Pitfall 5),
// "ghbdtn" typed through the physical path is consumed key by key and
// committed as «привет» by the engine, and the double tap corrects the
// Cyrillic word BACK to "ghbdtn" — one full loop of the ADR-001 Option B
// flip: flip → RU typing → correction, with the stdout oracle proving the
// final field content.
func runWordRUEN(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("word-ru-en needs the zenity entry surface (locked-session fallback engaged?)")
	}

	// Flip EN → RU: the Single decision fires at window expiry, so the
	// INFO mode record — not the tap — gates the typing step (Pitfall 5).
	if err := s.flipToRU(ctx, "word-ru-en"); err != nil {
		return err
	}

	// Physical Latin keys: the engine consumes each press and commits the
	// Cyrillic rune of the same key position — the entry settles at привет.
	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return err
	}
	if err := s.waitZenityText(ctx, wordResultRU); err != nil {
		return fmt.Errorf("word-ru-en RU typing: %w", err)
	}

	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"action","n":2`, decisionWait); err != nil {
		return fmt.Errorf("word-ru-en double-tap decision: %w", err)
	}
	if err := s.waitForLog(ctx, `"msg":"correction","outcome":"done"`, correctionWait); err != nil {
		return fmt.Errorf("word-ru-en correction: %w", err)
	}

	// The correction must replace привет with ghbdtn — content-exact.
	if err := s.waitZenityText(ctx, wordProbeEN); err != nil {
		return fmt.Errorf("word-ru-en applied correction: %w", err)
	}
	out, err := s.closeZenity(ctx)
	if err != nil {
		return err
	}
	if out != wordProbeEN {
		return fmt.Errorf("word-ru-en oracle: entry printed %q, want %q", out, wordProbeEN)
	}

	return nil
}

// runWordMixed proves the D-16 refusal live (plan 02-04): "gfb" typed in EN
// mode transits, the flip switches to RU, "ghbdtn" is committed as «привет»
// — the field holds the mixed word gfbпривет assembled through both real
// printing branches, and the double tap must leave it EXACTLY as it is: the
// daemon logs the mixed-script skip reason (D-20) and both oracles — the
// content-exact readback and the zenity stdout — print the untouched word.
func runWordMixed(ctx context.Context, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("word-mixed needs the zenity entry surface (locked-session fallback engaged?)")
	}

	if err := s.injectText(ctx, "gfb"); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"key"`, minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("word-mixed key visibility: %w", err)
	}

	// Flip EN → RU and wait for the mode record before typing (Pitfall 5).
	if err := s.flipToRU(ctx, "word-mixed"); err != nil {
		return err
	}

	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return err
	}
	if err := s.waitZenityChars(ctx, len([]rune(wordMixedExpected))); err != nil {
		return fmt.Errorf("word-mixed RU typing: %w", err)
	}

	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"action","n":2`, decisionWait); err != nil {
		return fmt.Errorf("word-mixed double-tap decision: %w", err)
	}
	// D-16/D-20: the refusal lands in the log with its reason — and the
	// field must not change by one rune.
	if err := s.waitForLog(ctx, `"reason":"mixed-script"`, correctionWait); err != nil {
		return fmt.Errorf("word-mixed D-16 refusal: %w", err)
	}

	if err := s.waitZenityText(ctx, wordMixedExpected); err != nil {
		return fmt.Errorf("word-mixed untouched oracle (D-16): %w", err)
	}
	out, err := s.closeZenity(ctx)
	if err != nil {
		return err
	}
	if out != wordMixedExpected {
		return fmt.Errorf("word-mixed oracle: entry printed %q, want the untouched %q", out, wordMixedExpected)
	}

	return nil
}

// readFocusedTextRaw reads the focused input's text content-exactly — the
// D-13 oracle is whitespace-sensitive and the trailing separator is
// load-bearing, so this path strips nothing but the helper's newline (the
// generic runCmd trims surrounding whitespace and eats the trailing space;
// live finding 2026-09-15: the stand's oracle read "привет" while a
// parallel shell read of the same field returned "привет ").
func (s *stand) readFocusedTextRaw(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/usr/bin/python3", s.cfg.helper, "focused-text")
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("focused-text readback: %w: %s", err, strings.TrimSpace(errOut.String()))
	}

	return strings.TrimSuffix(out.String(), "\n"), nil
}

// waitZenityText polls the content-exact readback until the focused entry
// holds exactly want — the AT-SPI bridge can lag the content update behind
// the character count (live-verified 2026-09-14: a trailing space showed
// in chars immediately but in the text a beat later), so the oracle rides
// the lag out instead of reading once.
func (s *stand) waitZenityText(ctx context.Context, want string) error {
	deadline := time.Now().Add(witnessWait)
	var last string
	for {
		out, err := s.readFocusedTextRaw(ctx)
		if err != nil {
			return err
		}
		last = out
		if out == want {
			return nil
		}
		if time.Now().After(deadline) {
			witness, _ := s.focusWitness(ctx)
			inputs, _ := runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "focused-inputs")

			return fmt.Errorf("entry did not settle to %q (readback %q, witness %q, focused inputs %q)",
				want, last, witness, inputs)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
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
