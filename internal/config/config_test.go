package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Djarvur/goswitch/internal/config"
	"github.com/Djarvur/goswitch/internal/hotkey"
)

// Corpus literals named once for the strict lint's goconst (the 02-01
// wordEN/wordRU idiom); shared with load_test.go of this test package.
const (
	defTapKey        = "shift_r"
	defCombo         = "shift+ctrl_r"
	defChord         = "super+space"
	fieldTapWindow   = "timeouts.tap_window_ms"
	fieldVerifyWait  = "timeouts.verify_wait_ms"
	fieldTapKey      = "hotkeys.tap_key"
	fieldChord       = "hotkeys.mode_switch_chord"
	fieldMACRLetters = "macr.letters"
	fieldMACRApps    = "macr.apps"
	fieldMACRAltMod  = "macr.alt_modifier"
)

// validDoc is the complete schema document every Validate case starts from
// (the plan's config_schema corpus, all sections, all keys).
func validDoc() config.Config {
	return config.Config{
		Hotkeys:    config.Hotkeys{TapKey: defTapKey, WordLayoutCombo: defCombo},
		Timeouts:   config.Timeouts{TapWindowMs: 300, VerifyWaitMs: 100},
		Correction: config.Correction{BackspaceCap: 50, ClipboardRung: false},
		MACR:       config.MACR{Enabled: false, Letters: "", Apps: nil, AltModifier: ""},
	}
}

// TestDefaults pins the documented defaults (config_schema): the tap window
// is ADR-002's DefaultWindow expressed in milliseconds, the Backspace cap
// is D-27's 50, the clipboard rung is OFF by default (D-28) and MACR never
// activates on its own — the alternative modifier is not introduced until
// the live probe clears it (ADR-005 b.3).
func TestDefaults(t *testing.T) {
	t.Parallel()

	got := config.Defaults()
	if got.Hotkeys.TapKey != defTapKey {
		t.Errorf("hotkeys.tap_key = %q, want %q", got.Hotkeys.TapKey, defTapKey)
	}
	if got.Hotkeys.WordLayoutCombo != defCombo {
		t.Errorf("hotkeys.word_layout_combo = %q, want %q", got.Hotkeys.WordLayoutCombo, defCombo)
	}
	if got.Hotkeys.ModeSwitchChord != defChord {
		t.Errorf("%s = %q, want %q (owner decision 2: goswitch owns Super+Space)",
			fieldChord, got.Hotkeys.ModeSwitchChord, defChord)
	}
	if got.Timeouts.TapWindowMs != 300 {
		t.Errorf("%s = %d, want 300", fieldTapWindow, got.Timeouts.TapWindowMs)
	}
	if time.Duration(got.Timeouts.TapWindowMs)*time.Millisecond != hotkey.DefaultWindow {
		t.Errorf("%s ≠ hotkey.DefaultWindow (%v)", fieldTapWindow, hotkey.DefaultWindow)
	}
	if got.Timeouts.VerifyWaitMs != 100 {
		t.Errorf("%s = %d, want 100", fieldVerifyWait, got.Timeouts.VerifyWaitMs)
	}
	if got.Correction.BackspaceCap != 50 {
		t.Errorf("correction.backspace_cap = %d, want 50 (D-27)", got.Correction.BackspaceCap)
	}
	if got.Correction.ClipboardRung {
		t.Error("correction.clipboard_rung = true, want false (D-28 opt-in)")
	}
	if !got.Correction.FlipAfterCorrection {
		// Owner decision 2: the mode follows a changed correction.
		t.Error("correction.flip_after_correction = false, want true")
	}
	if got.MACR.Enabled {
		t.Error("macr.enabled = true, want false")
	}
	if got.MACR.AltModifier != "" {
		t.Errorf("%s = %q, want empty (ADR-005 b.3: not introduced)", fieldMACRAltMod, got.MACR.AltModifier)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Defaults() does not validate: %v", err)
	}
}

// assertRejected fails unless cfg violates validation with the given field
// named in the error — the shared shape of every whole-config rejection.
func assertRejected(t *testing.T, name string, cfg config.Config, wantField string) {
	t.Helper()

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("%s accepted, want a whole-config rejection", name)
	}
	if !strings.Contains(err.Error(), wantField) {
		t.Errorf("error %q does not name the field %q", err, wantField)
	}
}

// TestValidate_TimeoutRanges pins the millisecond ceilings (T-03-02-02,
// ASVS V5): a giant or non-positive window/verify-wait rejects the whole
// config with the field named.
func TestValidate_TimeoutRanges(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		mutate    func(*config.Config)
		wantField string
	}{
		{
			name:      "tap window zero",
			mutate:    func(c *config.Config) { c.Timeouts.TapWindowMs = 0 },
			wantField: fieldTapWindow,
		},
		{
			name:      "tap window negative",
			mutate:    func(c *config.Config) { c.Timeouts.TapWindowMs = -1 },
			wantField: fieldTapWindow,
		},
		{
			name:      "tap window above ceiling",
			mutate:    func(c *config.Config) { c.Timeouts.TapWindowMs = 2001 },
			wantField: fieldTapWindow,
		},
		{
			name:      "verify wait zero",
			mutate:    func(c *config.Config) { c.Timeouts.VerifyWaitMs = 0 },
			wantField: fieldVerifyWait,
		},
		{
			name:      "verify wait above ceiling",
			mutate:    func(c *config.Config) { c.Timeouts.VerifyWaitMs = 2001 },
			wantField: fieldVerifyWait,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := validDoc()
			tc.mutate(&cfg)
			assertRejected(t, tc.name, cfg, tc.wantField)
		})
	}
}

// TestValidate_CorrectionRanges pins the Backspace cap range (D-27) and
// the MACR app-list ceiling: boundary violations reject the whole config
// with the field named.
func TestValidate_CorrectionRanges(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		mutate    func(*config.Config)
		wantField string
	}{
		{
			name:      "backspace cap zero",
			mutate:    func(c *config.Config) { c.Correction.BackspaceCap = 0 },
			wantField: "correction.backspace_cap",
		},
		{
			name:      "backspace cap above ceiling",
			mutate:    func(c *config.Config) { c.Correction.BackspaceCap = 501 },
			wantField: "correction.backspace_cap",
		},
		{
			name: "macr apps above ceiling",
			mutate: func(c *config.Config) {
				c.MACR.Apps = make([]string, 65)
			},
			wantField: fieldMACRApps,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := validDoc()
			tc.mutate(&cfg)
			assertRejected(t, tc.name, cfg, tc.wantField)
		})
	}
}

// TestValidate_BoundariesAccepted pins that the ceilings are inclusive:
// 2000 ms windows, a 500-press cap and a 64-entry app list all validate.
func TestValidate_BoundariesAccepted(t *testing.T) {
	t.Parallel()

	cfg := validDoc()
	cfg.Timeouts.TapWindowMs = 2000
	cfg.Timeouts.VerifyWaitMs = 2000
	cfg.Correction.BackspaceCap = 500
	cfg.MACR.Apps = make([]string, 64)
	if err := cfg.Validate(); err != nil {
		t.Errorf("ceiling values rejected: %v", err)
	}
}

// TestValidate_Bindings pins the binding resolution path: both key names
// resolve through hotkey.ParseBinding and an unknown token rejects the
// whole config with the field named.
func TestValidate_Bindings(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		mutate    func(*config.Config)
		wantField string
	}{
		{
			name:      "tap key unknown",
			mutate:    func(c *config.Config) { c.Hotkeys.TapKey = "shift_x" },
			wantField: fieldTapKey,
		},
		{
			name:      "tap key empty",
			mutate:    func(c *config.Config) { c.Hotkeys.TapKey = "" },
			wantField: fieldTapKey,
		},
		{
			// WR-02: tap semantics are key-only — the FSM tracks one keyval
			// and no modifier state, so a modifier-bearing tap_key would
			// "apply" while its modifier tokens are silently ignored.
			name:      "tap key combo (modifier-bearing)",
			mutate:    func(c *config.Config) { c.Hotkeys.TapKey = "ctrl+shift_r" },
			wantField: fieldTapKey,
		},
		{
			name:      "tap key redundant family modifier",
			mutate:    func(c *config.Config) { c.Hotkeys.TapKey = "shift+shift_r" },
			wantField: fieldTapKey,
		},
		{
			name:      "combo key unknown",
			mutate:    func(c *config.Config) { c.Hotkeys.WordLayoutCombo = "shift+ctrl_q" },
			wantField: "hotkeys.word_layout_combo",
		},
		{
			name:      "combo modifier unknown",
			mutate:    func(c *config.Config) { c.Hotkeys.WordLayoutCombo = "shif+ctrl_r" },
			wantField: "hotkeys.word_layout_combo",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := validDoc()
			tc.mutate(&cfg)
			assertRejected(t, tc.name, cfg, tc.wantField)
		})
	}
}

// TestValidate_MACRLetters pins the letter grammar: required once the
// layer is on, only single lowercase a–z tokens allowed (the Cyrillic
// keyboard has no business here — MACR forwards Latin Ctrl-shortcuts).
func TestValidate_MACRLetters(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		mutate func(*config.Config)
	}{
		{
			name: "enabled without letters",
			mutate: func(c *config.Config) {
				c.MACR.Enabled = true
			},
		},
		{
			name: "cyrillic letter",
			mutate: func(c *config.Config) {
				c.MACR.Enabled = true
				c.MACR.Letters = "a,я"
			},
		},
		{
			name: "multi-letter token",
			mutate: func(c *config.Config) {
				c.MACR.Enabled = true
				c.MACR.Letters = "ab,z"
			},
		},
		{
			name: "uppercase letter",
			mutate: func(c *config.Config) {
				c.MACR.Enabled = true
				c.MACR.Letters = "A"
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := validDoc()
			tc.mutate(&cfg)
			assertRejected(t, tc.name, cfg, fieldMACRLetters)
		})
	}
}

// TestValidate_MACRAltModifier pins the alternative-modifier closed set
// (ADR-005 b.3): ctrl_l/ctrl_r/empty validate, everything else — including
// a real modifier name that is not a candidate — rejects.
func TestValidate_MACRAltModifier(t *testing.T) {
	t.Parallel()

	cfg := validDoc()
	cfg.MACR.AltModifier = "shift"
	assertRejected(t, "alt modifier not a candidate", cfg, fieldMACRAltMod)

	accepted := validDoc()
	accepted.MACR = config.MACR{
		Enabled:     true,
		Letters:     "a,z",
		Apps:        []string{"google-chrome"},
		AltModifier: "ctrl_l",
	}
	if err := accepted.Validate(); err != nil {
		t.Errorf("valid MACR block rejected: %v", err)
	}
}

// chordDocYAML renders a complete document with the given
// hotkeys.mode_switch_chord YAML value — the decode corpus's template (the
// clipboard_rung idiom: full document, one key under test).
func chordDocYAML(value string) string {
	return `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
  mode_switch_chord: ` + value + `
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
`
}

// chordDocNoKeyYAML is the same document WITHOUT the mode_switch_chord key
// — every pre-existing document's shape.
const chordDocNoKeyYAML = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
`

// TestLoad_ModeSwitchChord pins the decode/validation surface of the chord
// key (owner decision 2, quick plan 260927-way): an explicit value decodes
// and validates, an EXPLICIT empty string decodes "" and validates (chord
// disabled — the missing-key compatibility rule of the vu8 batch, the
// macr.alt_modifier empty precedent, so every pre-existing document loads
// unchanged), and a garbage value refuses the WHOLE document (D-33).
func TestLoad_ModeSwitchChord(t *testing.T) {
	t.Parallel()

	t.Run("explicit value decodes and validates", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.Load(writeConfig(t, chordDocYAML("ctrl+space")))
		if err != nil {
			t.Fatalf("Load(ctrl+space doc): %v", err)
		}
		if cfg.Hotkeys.ModeSwitchChord != "ctrl+space" {
			t.Errorf("%s = %q, want %q", fieldChord, cfg.Hotkeys.ModeSwitchChord, "ctrl+space")
		}
	})

	t.Run("explicit empty disables", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.Load(writeConfig(t, chordDocYAML(`""`)))
		if err != nil {
			t.Fatalf("Load(empty doc): %v", err)
		}
		if cfg.Hotkeys.ModeSwitchChord != "" {
			t.Errorf("%s = %q, want %q (empty = the chord is disabled)", fieldChord, cfg.Hotkeys.ModeSwitchChord, "")
		}
	})

	t.Run("document without the key loads unchanged", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.Load(writeConfig(t, chordDocNoKeyYAML))
		if err != nil {
			t.Fatalf("Load(pre-batch doc): %v", err)
		}
		if cfg.Hotkeys.ModeSwitchChord != "" {
			t.Errorf("%s = %q, want %q — a missing key decodes disabled, never defaulted",
				fieldChord, cfg.Hotkeys.ModeSwitchChord, "")
		}
	})

	t.Run("garbage refuses the whole document", func(t *testing.T) {
		t.Parallel()

		_, err := config.Load(writeConfig(t, chordDocYAML("super+spaces")))
		if err == nil {
			t.Fatalf("Load(garbage chord) = nil error, want a whole-document rejection (D-33)")
		}
		if !strings.Contains(err.Error(), fieldChord) {
			t.Errorf("error %q does not name the field %q", err, fieldChord)
		}
	})
}
