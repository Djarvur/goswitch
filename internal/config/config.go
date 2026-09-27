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
)

// The documented default values (config_schema): the tap window is
// ADR-002's DefaultWindow in milliseconds, the verify wait the Phase 2
// surrounding-text timer, the Backspace cap D-27's series bound.
const (
	defaultTapWindowMs  = 300
	defaultVerifyWaitMs = 100
	defaultBackspaceCap = 50
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
}

// Timeouts are the timing parameters, in milliseconds.
type Timeouts struct {
	TapWindowMs  int `yaml:"tap_window_ms"`
	VerifyWaitMs int `yaml:"verify_wait_ms"`
}

// Correction carries the replacement-ladder parameters: the D-27 Backspace
// series cap and the D-28 opt-in clipboard rung.
type Correction struct {
	BackspaceCap  int  `yaml:"backspace_cap"`
	ClipboardRung bool `yaml:"clipboard_rung"`
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

// Config is the whole daemon configuration: exactly the four sections
// hotkeys / timeouts / correction / macr (D-31).
type Config struct {
	Hotkeys    Hotkeys    `yaml:"hotkeys"`
	Timeouts   Timeouts   `yaml:"timeouts"`
	Correction Correction `yaml:"correction"`
	MACR       MACR       `yaml:"macr"`
}

// Defaults returns the documented built-in defaults: the daemon runs on
// them when -config is absent, so they must equal the Phase 2 behavior —
// the tap window is ADR-002's DefaultWindow (300 ms), the Backspace cap
// D-27's 50, the clipboard rung off (D-28) and MACR off with no
// alternative modifier (ADR-005 b.3).
func Defaults() Config {
	return Config{
		Hotkeys: Hotkeys{
			TapKey:          defTapKey,
			WordLayoutCombo: defCombo,
		},
		Timeouts: Timeouts{
			TapWindowMs:  defaultTapWindowMs,
			VerifyWaitMs: defaultVerifyWaitMs,
		},
		Correction: Correction{
			BackspaceCap:  defaultBackspaceCap,
			ClipboardRung: false,
		},
		MACR: MACR{
			Enabled:     false,
			Letters:     "",
			Apps:        nil,
			AltModifier: "",
		},
	}
}

// The documented default binding names (shared with the corpus).
const (
	defTapKey = "shift_r"
	defCombo  = "shift+ctrl_r"
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

	return c.MACR.validate()
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
