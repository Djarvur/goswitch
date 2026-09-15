package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Djarvur/goswitch/internal/config"
	"github.com/Djarvur/goswitch/internal/hotkey"
)

// validDoc is the complete schema document every Validate case starts from
// (the plan's config_schema corpus, all sections, all keys).
func validDoc() config.Config {
	return config.Config{
		Hotkeys:    config.Hotkeys{TapKey: "shift_r", WordLayoutCombo: "shift+ctrl_r"},
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
	if got.Hotkeys.TapKey != "shift_r" {
		t.Errorf("hotkeys.tap_key = %q, want shift_r", got.Hotkeys.TapKey)
	}
	if got.Hotkeys.WordLayoutCombo != "shift+ctrl_r" {
		t.Errorf("hotkeys.word_layout_combo = %q, want shift+ctrl_r", got.Hotkeys.WordLayoutCombo)
	}
	if got.Timeouts.TapWindowMs != 300 {
		t.Errorf("timeouts.tap_window_ms = %d, want 300", got.Timeouts.TapWindowMs)
	}
	if time.Duration(got.Timeouts.TapWindowMs)*time.Millisecond != hotkey.DefaultWindow {
		t.Errorf(
			"timeouts.tap_window_ms ≠ hotkey.DefaultWindow (%v)",
			hotkey.DefaultWindow,
		)
	}
	if got.Timeouts.VerifyWaitMs != 100 {
		t.Errorf("timeouts.verify_wait_ms = %d, want 100", got.Timeouts.VerifyWaitMs)
	}
	if got.Correction.BackspaceCap != 50 {
		t.Errorf("correction.backspace_cap = %d, want 50 (D-27)", got.Correction.BackspaceCap)
	}
	if got.Correction.ClipboardRung {
		t.Error("correction.clipboard_rung = true, want false (D-28 opt-in)")
	}
	if got.MACR.Enabled {
		t.Error("macr.enabled = true, want false")
	}
	if got.MACR.AltModifier != "" {
		t.Errorf("macr.alt_modifier = %q, want empty (ADR-005 b.3: not introduced)", got.MACR.AltModifier)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Defaults() does not validate: %v", err)
	}
}

// TestValidate_Ranges pins the DoS ceilings (T-03-02-02, ASVS V5): every
// boundary violation rejects the whole config with the field named in the
// error — a giant value must never ride through.
func TestValidate_Ranges(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		mutate    func(*config.Config)
		wantField string
	}{
		{
			name:      "tap window zero",
			mutate:    func(c *config.Config) { c.Timeouts.TapWindowMs = 0 },
			wantField: "timeouts.tap_window_ms",
		},
		{
			name:      "tap window negative",
			mutate:    func(c *config.Config) { c.Timeouts.TapWindowMs = -1 },
			wantField: "timeouts.tap_window_ms",
		},
		{
			name:      "tap window above ceiling",
			mutate:    func(c *config.Config) { c.Timeouts.TapWindowMs = 2001 },
			wantField: "timeouts.tap_window_ms",
		},
		{
			name:      "verify wait zero",
			mutate:    func(c *config.Config) { c.Timeouts.VerifyWaitMs = 0 },
			wantField: "timeouts.verify_wait_ms",
		},
		{
			name:      "verify wait above ceiling",
			mutate:    func(c *config.Config) { c.Timeouts.VerifyWaitMs = 2001 },
			wantField: "timeouts.verify_wait_ms",
		},
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
			wantField: "macr.apps",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := validDoc()
			tc.mutate(&cfg)
			err := cfg.Validate()
			if err == nil {
				t.Fatalf("%s accepted, want a whole-config rejection", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantField) {
				t.Errorf("error %q does not name the field %q", err, tc.wantField)
			}
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
// resolve through hotkey.ParseBinding, an unknown token rejects the whole
// config with the field named, and the documented defaults pass.
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
			wantField: "hotkeys.tap_key",
		},
		{
			name:      "tap key empty",
			mutate:    func(c *config.Config) { c.Hotkeys.TapKey = "" },
			wantField: "hotkeys.tap_key",
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
			err := cfg.Validate()
			if err == nil {
				t.Fatalf("%s accepted, want a whole-config rejection", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantField) {
				t.Errorf("error %q does not name the field %q", err, tc.wantField)
			}
		})
	}
}

// TestValidate_MACRFields pins the MACR grammar: letters are required once
// the layer is on, only single lowercase a–z tokens are allowed, and the
// alternative modifier is a two-name closed set (ADR-005 b.3 candidates).
func TestValidate_MACRFields(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		mutate    func(*config.Config)
		wantField string
	}{
		{
			name: "enabled without letters",
			mutate: func(c *config.Config) {
				c.MACR.Enabled = true
			},
			wantField: "macr.letters",
		},
		{
			name: "cyrillic letter",
			mutate: func(c *config.Config) {
				c.MACR.Enabled = true
				c.MACR.Letters = "a,я"
			},
			wantField: "macr.letters",
		},
		{
			name: "multi-letter token",
			mutate: func(c *config.Config) {
				c.MACR.Enabled = true
				c.MACR.Letters = "ab,z"
			},
			wantField: "macr.letters",
		},
		{
			name: "uppercase letter",
			mutate: func(c *config.Config) {
				c.MACR.Enabled = true
				c.MACR.Letters = "A"
			},
			wantField: "macr.letters",
		},
		{
			name: "alt modifier not a candidate",
			mutate: func(c *config.Config) {
				c.MACR.AltModifier = "shift"
			},
			wantField: "macr.alt_modifier",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := validDoc()
			tc.mutate(&cfg)
			err := cfg.Validate()
			if err == nil {
				t.Fatalf("%s accepted, want a whole-config rejection", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantField) {
				t.Errorf("error %q does not name the field %q", err, tc.wantField)
			}
		})
	}

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
