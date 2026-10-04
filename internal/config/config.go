// Package config holds the daemon's YAML configuration: the schema of all
// hotkeys, timeouts, correction and MACR parameters (D-31), strict decoding
// where a typo'd key invalidates the whole document (D-33) and the hot
// reload watcher publishing last-good snapshots (D-32). The package stays
// free of D-Bus and the engine — it resolves bindings through the pure
// internal/hotkey name tables.
package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Djarvur/goswitch/internal/hotkey"
)

// Range ceilings and list caps of the schema — the DoS guard rails every
// Validate error cites (T-03-02-02, ASVS V5).
const (
	maxTapWindowMs  = 2000
	maxVerifyWaitMs = 2000
	minBackspaceCap = 1
	maxBackspaceCap = 500
	maxMACRApps     = 64

	maxAutocorrectApps    = 64
	minAutocorrectWordLen = 2
	maxAutocorrectWordLen = 16
)

// The documented default values (config_schema): the tap window is
// ADR-002's DefaultWindow in milliseconds, the verify wait the Phase 2
// surrounding-text timer, the Backspace cap D-27's series bound. The
// autocorrect thresholds are the detector's start values pending the
// internal/detect golden corpus (plan 06-05) — change the defaults and
// detect.Params TOGETHER, never one side.
const (
	defaultTapWindowMs   = 300
	defaultVerifyWaitMs  = 100
	defaultBackspaceCap  = 50
	defaultMinWordLen    = 4
	defaultTrigramMargin = 2.0
	defaultTrigramFloor  = 1.0
)

// Validation refusals — package sentinels (the err113 discipline of the
// strict lint: no dynamically created error values), each wrapped with the
// offending field and value so every error names its field (D-33).
var (
	errTapWindowRange     = errors.New("must be in (0, 2000] ms")
	errVerifyWaitRange    = errors.New("must be in (0, 2000] ms")
	errBackspaceCapRange  = errors.New("must be in [1, 500]")
	errTapKeyBare         = errors.New("must be a bare key name — modifiers belong to word_layout_combo")
	errMACRLettersReq     = errors.New("is required when macr.enabled is true")
	errMACRLetterToken    = errors.New("must be a single lowercase letter a-z")
	errMACRAppsOverCeil   = errors.New("entries, at most 64 allowed")
	errMACRAltModifierSet = errors.New("must be ctrl_l, ctrl_r or empty (empty = not introduced)")

	errAutocorrectAppsOverCeil    = errors.New("entries, at most 64 allowed")
	errAutocorrectMinWordLenRange = errors.New("must be in [2, 16]")
	errAutocorrectThresholdRange  = errors.New("must be positive, with trigram_margin >= trigram_floor")
)

// altModifierCandidates is the closed set of alternative MACR modifiers
// (ADR-005 b.3) — empty means "do not introduce one". A function, not a
// package-level var: the strict lint forbids mutable globals (matrix.go
// idiom).
func altModifierCandidates() []string {
	return []string{"", "ctrl_l", "ctrl_r"}
}

// Hotkeys are the daemon's key bindings, expressed as closed-table name
// strings resolved by hotkey.ParseBinding (D-31: name strings, combos
// "+"-joined with the key last).
type Hotkeys struct {
	TapKey          string `yaml:"tap_key"`
	WordLayoutCombo string `yaml:"word_layout_combo"`
	ModeSwitchChord string `yaml:"mode_switch_chord"`
}

// Timeouts are the timing parameters, in milliseconds.
type Timeouts struct {
	TapWindowMs  int `yaml:"tap_window_ms"`
	VerifyWaitMs int `yaml:"verify_wait_ms"`
}

// Correction carries the replacement-ladder parameters: the D-27 Backspace
// series cap, the D-28 opt-in clipboard rung and the post-correction script
// flip switch (owner decision 2, quick plan 260927-vu8).
type Correction struct {
	BackspaceCap        int  `yaml:"backspace_cap"`
	ClipboardRung       bool `yaml:"clipboard_rung"`
	FlipAfterCorrection bool `yaml:"flip_after_correction"`
}

// MACR is the Super→Ctrl remapping layer (ADR-005): a global switch, the
// remapped letter set, per-app allow lists and the alternative modifier —
// which defaults to empty, i.e. NOT introduced (ADR-005 b.3).
type MACR struct {
	Enabled     bool     `yaml:"enabled"`
	Letters     string   `yaml:"letters"`
	Apps        []string `yaml:"apps"`
	AltModifier string   `yaml:"alt_modifier"`
}

// Autocorrect is the automatic wrong-layout correction layer's schema
// section (D-54, opt-in): the global switch, the per-app white list of
// bridge-namespace names — exact string matches, order carries no
// meaning — and the detector thresholds. The zero value is the OFF
// state: a document without the section decodes disabled, never
// activated (default off everywhere, D-54).
type Autocorrect struct {
	Enabled       bool     `yaml:"enabled"`
	Apps          []string `yaml:"apps"`
	MinWordLen    int      `yaml:"min_word_len"`
	TrigramMargin float64  `yaml:"trigram_margin"`
	TrigramFloor  float64  `yaml:"trigram_floor"`
}

// Config is the whole daemon configuration: exactly the five sections
// hotkeys / timeouts / correction / macr / autocorrect (D-31, D-54).
type Config struct {
	Hotkeys     Hotkeys     `yaml:"hotkeys"`
	Timeouts    Timeouts    `yaml:"timeouts"`
	Correction  Correction  `yaml:"correction"`
	MACR        MACR        `yaml:"macr"`
	Autocorrect Autocorrect `yaml:"autocorrect"`
}

// Defaults returns the documented built-in defaults: the daemon runs on
// them when -config is absent, so they must equal the Phase 2 behavior —
// the tap window is ADR-002's DefaultWindow (300 ms), the Backspace cap
// D-27's 50, the clipboard rung off (D-28), MACR off with no
// alternative modifier (ADR-005 b.3), the post-correction script flip
// ON (owner decision 2, 2026-09-27: the mode follows a changed
// correction) and the autocorrect layer OFF with an empty white list —
// the zero value is the off state (D-54 default off everywhere).
func Defaults() Config {
	return Config{
		Hotkeys: Hotkeys{
			TapKey:          defTapKey,
			WordLayoutCombo: defCombo,
			ModeSwitchChord: defModeSwitchChord,
		},
		Timeouts: Timeouts{
			TapWindowMs:  defaultTapWindowMs,
			VerifyWaitMs: defaultVerifyWaitMs,
		},
		Correction: Correction{
			BackspaceCap:        defaultBackspaceCap,
			ClipboardRung:       false,
			FlipAfterCorrection: true,
		},
		MACR: MACR{
			Enabled:     false,
			Letters:     "",
			Apps:        nil,
			AltModifier: "",
		},
		// The thresholds are the detector's start values pending the
		// internal/detect golden corpus (plan 06-05) — change the defaults
		// and detect.Params TOGETHER, never one side.
		Autocorrect: Autocorrect{
			Enabled:       false,
			Apps:          nil,
			MinWordLen:    defaultMinWordLen,
			TrigramMargin: defaultTrigramMargin,
			TrigramFloor:  defaultTrigramFloor,
		},
	}
}

// The documented default binding names (shared with the corpus).
const (
	defTapKey          = "shift_r"
	defCombo           = "shift+ctrl_r"
	defModeSwitchChord = "super+space"
)

// Validate enforces the ranges and closed vocabularies; every error names
// the offending field, and any violation invalidates the whole config
// (D-33 — never a partially applied file).
func (c Config) Validate() error {
	if err := c.Hotkeys.validate(); err != nil {
		return err
	}
	if err := c.Timeouts.validate(); err != nil {
		return err
	}
	if err := c.Correction.validate(); err != nil {
		return err
	}
	if err := c.MACR.validate(); err != nil {
		return err
	}

	return c.Autocorrect.validate()
}

// validate resolves both bindings through the closed hotkey name tables —
// an unknown name anywhere rejects the config (D-33). The tap key must
// additionally be a BARE key (WR-02): tap semantics are key-only — the FSM
// tracks a single keyval and no modifier state — so a modifier-bearing
// tap_key would pass the binding grammar and "apply" while its modifier
// tokens are silently ignored. The combo is the only binding that uses
// modifiers.
func (h Hotkeys) validate() error {
	if strings.Contains(h.TapKey, "+") {
		return fmt.Errorf("hotkeys.tap_key %q: %w", h.TapKey, errTapKeyBare)
	}
	if _, err := hotkey.ParseBinding(h.TapKey); err != nil {
		return fmt.Errorf("hotkeys.tap_key %q: %w", h.TapKey, err)
	}
	if _, err := hotkey.ParseBinding(h.WordLayoutCombo); err != nil {
		return fmt.Errorf("hotkeys.word_layout_combo %q: %w", h.WordLayoutCombo, err)
	}
	if h.ModeSwitchChord == "" {
		// The chord is DISABLED by an empty value — the missing-key
		// compatibility rule (the macr.alt_modifier empty precedent):
		// documents written before the key exist load unchanged with the
		// chord off, never defaulted on.
		return nil
	}
	if _, err := hotkey.ParseBinding(h.ModeSwitchChord); err != nil {
		return fmt.Errorf("hotkeys.mode_switch_chord %q: %w", h.ModeSwitchChord, err)
	}

	return nil
}

// validate enforces the millisecond ceilings: a giant window or verify
// wait is a DoS vector, not a setting (T-03-02-02).
func (t Timeouts) validate() error {
	if t.TapWindowMs <= 0 || t.TapWindowMs > maxTapWindowMs {
		return fmt.Errorf("timeouts.tap_window_ms = %d: %w", t.TapWindowMs, errTapWindowRange)
	}
	if t.VerifyWaitMs <= 0 || t.VerifyWaitMs > maxVerifyWaitMs {
		return fmt.Errorf("timeouts.verify_wait_ms = %d: %w", t.VerifyWaitMs, errVerifyWaitRange)
	}

	return nil
}

// validate enforces the Backspace series cap range (D-27: the ladder's
// destructive series is bounded).
func (c Correction) validate() error {
	if c.BackspaceCap < minBackspaceCap || c.BackspaceCap > maxBackspaceCap {
		return fmt.Errorf(
			"correction.backspace_cap = %d: %w",
			c.BackspaceCap, errBackspaceCapRange,
		)
	}

	return nil
}

// validate enforces the MACR grammar: letters are required once the layer
// is on and only single lowercase a–z tokens are allowed; the app list is
// capped; the alternative modifier is the two-name closed set of ADR-005
// b.3 candidates.
func (m MACR) validate() error {
	if m.Enabled && m.Letters == "" {
		return fmt.Errorf("macr.letters %w", errMACRLettersReq)
	}
	if m.Letters != "" {
		for _, tok := range strings.Split(m.Letters, ",") {
			if len(tok) != 1 || tok[0] < 'a' || tok[0] > 'z' {
				return fmt.Errorf("macr.letters token %q: %w", tok, errMACRLetterToken)
			}
		}
	}
	if len(m.Apps) > maxMACRApps {
		return fmt.Errorf("macr.apps has %d %w", len(m.Apps), errMACRAppsOverCeil)
	}
	if !slices.Contains(altModifierCandidates(), m.AltModifier) {
		return fmt.Errorf("macr.alt_modifier %q: %w", m.AltModifier, errMACRAltModifierSet)
	}

	return nil
}

// validate enforces the autocorrect grammar (D-54): the white list is
// capped unconditionally — a giant list is a DoS vector even while the
// layer is dormant (T-06-04-01, the maxMACRApps precedent). The threshold
// ranges bite only on an ACTIVE section — enabled with a non-empty white
// list, the only shape that can ever fire (T-06-04-03) — so the zero
// value stays valid and documents without the section (or with the
// feature switched off) decode off, never defaulted on (default off
// everywhere, D-54). Every error names its field (D-33).
func (a Autocorrect) validate() error {
	if len(a.Apps) > maxAutocorrectApps {
		return fmt.Errorf("autocorrect.apps has %d %w", len(a.Apps), errAutocorrectAppsOverCeil)
	}
	if !a.Enabled || len(a.Apps) == 0 {
		return nil
	}
	if a.MinWordLen < minAutocorrectWordLen || a.MinWordLen > maxAutocorrectWordLen {
		return fmt.Errorf("autocorrect.min_word_len = %d: %w", a.MinWordLen, errAutocorrectMinWordLenRange)
	}
	if a.TrigramFloor <= 0 {
		return fmt.Errorf("autocorrect.trigram_floor = %v: %w", a.TrigramFloor, errAutocorrectThresholdRange)
	}
	if a.TrigramMargin <= 0 || a.TrigramMargin < a.TrigramFloor {
		return fmt.Errorf("autocorrect.trigram_margin = %v: %w", a.TrigramMargin, errAutocorrectThresholdRange)
	}

	return nil
}
