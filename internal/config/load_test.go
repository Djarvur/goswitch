package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Djarvur/goswitch/internal/config"
)

// fullDocYAML is the complete schema document of the plan's config_schema —
// every key present, the documented defaults as values.
const fullDocYAML = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
`

// writeConfig writes the corpus to a temp file and returns its path.
func writeConfig(t *testing.T, corpus string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "goswitch.yaml")
	if err := os.WriteFile(path, []byte(corpus), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return path
}

// TestLoad_ValidDoc pins the happy path: the full schema document decodes
// into Config with every field in place (D-31, exact snake_case keys).
func TestLoad_ValidDoc(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(writeConfig(t, fullDocYAML))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Hotkeys.TapKey != defTapKey {
		t.Errorf("hotkeys.tap_key = %q, want %q", cfg.Hotkeys.TapKey, defTapKey)
	}
	if cfg.Hotkeys.WordLayoutCombo != defCombo {
		t.Errorf("hotkeys.word_layout_combo = %q, want %q", cfg.Hotkeys.WordLayoutCombo, defCombo)
	}
	if cfg.Timeouts.TapWindowMs != 300 {
		t.Errorf("%s = %d, want 300", fieldTapWindow, cfg.Timeouts.TapWindowMs)
	}
	if cfg.Timeouts.VerifyWaitMs != 100 {
		t.Errorf("%s = %d, want 100", fieldVerifyWait, cfg.Timeouts.VerifyWaitMs)
	}
	if cfg.Correction.BackspaceCap != 50 {
		t.Errorf("correction.backspace_cap = %d, want 50", cfg.Correction.BackspaceCap)
	}
	if cfg.Correction.ClipboardRung {
		t.Error("correction.clipboard_rung = true, want false")
	}
	if cfg.MACR.Enabled {
		t.Error("macr.enabled = true, want false")
	}
	if cfg.MACR.Letters != "" || cfg.MACR.AltModifier != "" {
		t.Errorf("macr.letters/alt_modifier = %q/%q, want empty/empty", cfg.MACR.Letters, cfg.MACR.AltModifier)
	}
	if len(cfg.MACR.Apps) != 0 {
		t.Errorf("macr.apps = %v, want empty", cfg.MACR.Apps)
	}
}

// TestLoad_UnknownKeyRejected pins the strict decoder (D-33): a typo'd
// section key or a typo'd field inside a section invalidates the whole
// document — the user who misspells a timeout name must get an error, not
// the illusion of an applied setting.
func TestLoad_UnknownKeyRejected(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		corpus string
	}{
		{
			name:   "typo'd section",
			corpus: strings.Replace(fullDocYAML, "timeouts:", "timeots:", 1),
		},
		{
			name:   "typo'd field",
			corpus: strings.Replace(fullDocYAML, "backspace_cap:", "backspace_capp:", 1),
		},
		{
			name:   "unknown top-level section",
			corpus: fullDocYAML + "extra:\n  key: value\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := config.Load(writeConfig(t, tc.corpus)); err == nil {
				t.Fatalf("%s accepted, want strict rejection", tc.name)
			}
		})
	}
}

// TestLoad_MissingFileRejected pins the open-error path: an explicit
// -config pointing nowhere is a visible start refusal.
func TestLoad_MissingFileRejected(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "absent.yaml")
	if _, err := config.Load(path); err == nil {
		t.Fatal("missing config file accepted, want an open error")
	}
}

// TestLoad_EmptyDocumentRejected pins the "explicit -config must yield a
// working config" rule (prohibition 2): an empty document is a refusal,
// never a silent fall-through to built-in defaults.
func TestLoad_EmptyDocumentRejected(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		corpus string
	}{
		{name: "empty file", corpus: ""},
		{name: "whitespace only", corpus: "\n\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := config.Load(writeConfig(t, tc.corpus)); err == nil {
				t.Fatalf("%s accepted, want a refusal", tc.name)
			}
		})
	}
}

// TestLoad_RangeViolationRejected pins that Load runs Validate: a document
// that decodes cleanly but violates a range is rejected at load time too.
func TestLoad_RangeViolationRejected(t *testing.T) {
	t.Parallel()

	corpus := strings.Replace(fullDocYAML, "tap_window_ms: 300", "tap_window_ms: 100000000", 1)
	if _, err := config.Load(writeConfig(t, corpus)); err == nil {
		t.Fatal("giant tap_window_ms accepted at load, want range rejection")
	}
}
