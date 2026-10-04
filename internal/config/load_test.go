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
  flip_after_correction: true
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
autocorrect:
  enabled: false
  apps_blocklist: []
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
`

// docNoAutocorrectYAML is the complete PRE-phase document — the schema
// before plan 06-04 added the autocorrect section — every shipped user
// document's shape.
const docNoAutocorrectYAML = `hotkeys:
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

// writeConfig writes the corpus to a temp file and returns its path.
func writeConfig(t *testing.T, corpus string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "goswitch.yaml")
	if err := os.WriteFile(path, []byte(corpus), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return path
}

// assertDecodeDefaults fails unless cfg's autocorrect section carries the
// documented default values — the shared tail of the decode corpus.
func assertDecodeDefaults(t *testing.T, cfg *config.Config) {
	t.Helper()

	if cfg.Autocorrect.Enabled {
		t.Error("autocorrect.enabled = true, want false")
	}
	if len(cfg.Autocorrect.AppsBlocklist) != 0 {
		t.Errorf("autocorrect.apps_blocklist = %v, want empty", cfg.Autocorrect.AppsBlocklist)
	}
	if cfg.Autocorrect.MinWordLen != 4 || cfg.Autocorrect.TrigramMargin != 2.0 || cfg.Autocorrect.TrigramFloor != 1.0 {
		t.Errorf(
			"autocorrect thresholds = %d/%v/%v, want 4/2.0/1.0",
			cfg.Autocorrect.MinWordLen, cfg.Autocorrect.TrigramMargin, cfg.Autocorrect.TrigramFloor,
		)
	}
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
	if !cfg.Correction.FlipAfterCorrection {
		t.Error("correction.flip_after_correction = false, want true")
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
	assertDecodeDefaults(t, cfg)
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

// TestLoad_DocumentWithoutSectionDecodesOff pins the D-54 decode contract
// (plan prohibition 1): a document without the autocorrect section decodes
// into the ZERO value — disabled, empty white list, no defaults overlay in
// Load — and the whole file stays valid. Defaults() is the no-flag path
// only (the 03-02 complete-document precedent).
func TestLoad_DocumentWithoutSectionDecodesOff(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(writeConfig(t, docNoAutocorrectYAML))
	if err != nil {
		t.Fatalf("Load(pre-phase document): %v", err)
	}
	got := cfg.Autocorrect
	zero := got.Enabled || got.AppsBlocklist != nil ||
		got.MinWordLen != 0 || got.TrigramMargin != 0 || got.TrigramFloor != 0
	if zero {
		t.Errorf(
			"autocorrect = %+v, want the zero value (no defaults overlay in Load; absent section = off, D-54)",
			got,
		)
	}
}

// TestLoad_StrictDecodeRejectsTypo pins D-33 propagation to the new
// section (T-06-04-02): a typo'd key inside the autocorrect section
// invalidates the WHOLE document — an `enbled` typo can never silently
// half-apply the section or quietly disable it.
func TestLoad_StrictDecodeRejectsTypo(t *testing.T) {
	t.Parallel()

	corpus := strings.Replace(fullDocYAML, "autocorrect:\n  enabled: false", "autocorrect:\n  enbled: true", 1)
	_, err := config.Load(writeConfig(t, corpus))
	if err == nil {
		t.Fatal("typo'd autocorrect key accepted, want a whole-document rejection (D-33)")
	}
	if !strings.Contains(err.Error(), "enbled") {
		t.Errorf("error %q does not name the unknown field enbled", err)
	}
}

// TestLoad_EmptyAppsEnabledIsValid pins the kill-switch shape at load
// level: `enabled: true` with an empty blocklist and the documented
// thresholds is a VALID document — under the blocklist semantics (D-53)
// an empty list forbids nothing, so the layer is ACTIVE wherever the
// field's role gate lets it through.
func TestLoad_EmptyAppsEnabledIsValid(t *testing.T) {
	t.Parallel()

	corpus := strings.Replace(fullDocYAML, "autocorrect:\n  enabled: false", "autocorrect:\n  enabled: true", 1)
	cfg, err := config.Load(writeConfig(t, corpus))
	if err != nil {
		t.Fatalf("Load(enabled, empty blocklist): %v", err)
	}
	if !cfg.Autocorrect.Enabled {
		t.Error("autocorrect.enabled = false, want true (the document's value)")
	}
	if len(cfg.Autocorrect.AppsBlocklist) != 0 {
		t.Errorf("autocorrect.apps_blocklist = %v, want empty — nothing is forbidden", cfg.Autocorrect.AppsBlocklist)
	}
}

// TestLoad_AutocorrectRangeViolationRejected pins that Load runs the
// autocorrect validation: an ACTIVE section (enabled, non-empty white
// list) with an out-of-range threshold is rejected at load time, not at
// the first keystroke.
func TestLoad_AutocorrectRangeViolationRejected(t *testing.T) {
	t.Parallel()

	active := strings.Replace(
		fullDocYAML,
		"autocorrect:\n  enabled: false\n  apps_blocklist: []",
		"autocorrect:\n  enabled: true\n  apps_blocklist: [\"org.gnome.Gedit\"]",
		1,
	)
	broken := strings.Replace(active, "min_word_len: 4", "min_word_len: 1", 1)
	if _, err := config.Load(writeConfig(t, broken)); err == nil {
		t.Fatal("active section with min_word_len 1 accepted at load, want range rejection")
	}
}

// TestLoad_FlipAfterCorrectionDecode pins the decode round-trip of the
// post-correction flip switch (owner decision 2, 260927-vu8): a document
// carrying `flip_after_correction: false` decodes false, `true` decodes
// true — a plain bool, no range to validate.
func TestLoad_FlipAfterCorrectionDecode(t *testing.T) {
	t.Parallel()

	offDoc := strings.Replace(fullDocYAML, "flip_after_correction: true", "flip_after_correction: false", 1)
	cfg, err := config.Load(writeConfig(t, offDoc))
	if err != nil {
		t.Fatalf("Load (flip off): %v", err)
	}
	if cfg.Correction.FlipAfterCorrection {
		t.Error("correction.flip_after_correction = true, want false (the document's value)")
	}

	onDoc := fullDocYAML
	cfg, err = config.Load(writeConfig(t, onDoc))
	if err != nil {
		t.Fatalf("Load (flip on): %v", err)
	}
	if !cfg.Correction.FlipAfterCorrection {
		t.Error("correction.flip_after_correction = false, want true (the document's value)")
	}
}

// legacyAppsDocYAML is the PRE-revision document: the autocorrect section
// carries the white-list key the D-53 revision REMOVED. The strict
// decoder must reject the whole file (D-33 — the old key gets no silent
// half-support, and an old white list is never transferred to the
// blocklist mechanically: the two keys carry opposite intents).
const legacyAppsDocYAML = `hotkeys:
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
autocorrect:
  enabled: false
  apps: []
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
`

// TestLoad_StrictDecodeRejectsLegacyAppsKey pins the D-33 hard rename
// (plan 07-02): the white-list key is GONE from the schema — a document
// carrying it is rejected WHOLE by the strict decoder, loudly, never
// half-applied or quietly ignored (T-07-02-04).
func TestLoad_StrictDecodeRejectsLegacyAppsKey(t *testing.T) {
	t.Parallel()

	_, err := config.Load(writeConfig(t, legacyAppsDocYAML))
	if err == nil {
		t.Fatal("legacy autocorrect.apps key accepted, want a whole-document rejection (D-33)")
	}
	if !strings.Contains(err.Error(), "apps") {
		t.Errorf("error %q does not name the removed field apps", err)
	}
}

// soundDocYAML renders the complete document with the given sound section
// body — the decode corpus's template for the sound keys (the chordDoc
// idiom: full document, one section under test).
func soundDocYAML(body string) string {
	return strings.Replace(fullDocYAML, "autocorrect:", "sound:\n"+body+"autocorrect:", 1)
}

// TestSound_AbsentSectionMeansOn pins the default-ON contract (owner
// decision, 07-CONTEXT): a document without the sound section decodes
// the ZERO value — and the section's effective switch reads ON, because
// Load never overlays defaults (03-02) and the pointer-bool's nil is the
// "on" verdict; a plain bool's zero value would have read "off".
func TestSound_AbsentSectionMeansOn(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(writeConfig(t, fullDocYAML))
	if err != nil {
		t.Fatalf("Load(document without sound): %v", err)
	}
	if !cfg.Sound.EffectiveEnabled() {
		t.Error("Sound.EffectiveEnabled() = false for an absent section, want true (default ON)")
	}
	if cfg.Sound.AutocorrectEvent != "" {
		t.Errorf("Sound.AutocorrectEvent = %q, want empty — Load never overlays defaults", cfg.Sound.AutocorrectEvent)
	}
}

// TestSound_ExplicitOffAndCustomEvent pins the decode round-trip of the
// sound keys: an explicit enabled: false silences the sounds, a custom
// autocorrect_event reaches the accessor verbatim, and an omitted event
// name reads as the built-in default event.
func TestSound_ExplicitOffAndCustomEvent(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(writeConfig(t, soundDocYAML("  enabled: false\n")))
	if err != nil {
		t.Fatalf("Load(sound off): %v", err)
	}
	if cfg.Sound.EffectiveEnabled() {
		t.Error("Sound.EffectiveEnabled() = true for enabled: false, want false (the document's value)")
	}

	cfg, err = config.Load(writeConfig(t, soundDocYAML("  enabled: true\n  autocorrect_event: custom\n")))
	if err != nil {
		t.Fatalf("Load(sound custom event): %v", err)
	}
	if got := cfg.Sound.EffectiveAutocorrectEvent(); got != "custom" {
		t.Errorf("Sound.EffectiveAutocorrectEvent() = %q, want %q (the document's value)", got, "custom")
	}

	cfg, err = config.Load(writeConfig(t, soundDocYAML("  enabled: true\n")))
	if err != nil {
		t.Fatalf("Load(sound default event): %v", err)
	}
	want := config.DefaultSoundAutocorrectEvent
	if got := cfg.Sound.EffectiveAutocorrectEvent(); got != want {
		t.Errorf("Sound.EffectiveAutocorrectEvent() = %q, want the default %q", got, want)
	}
}

// TestSound_StrictDecodeUnknownKey pins D-33 propagation to the new
// section: an unknown key inside the sound section invalidates the WHOLE
// document — the strict decoder covers the section automatically, no new
// code in load.go.
func TestSound_StrictDecodeUnknownKey(t *testing.T) {
	t.Parallel()

	_, err := config.Load(writeConfig(t, soundDocYAML("  enabled: true\n  sound_effect: wat\n")))
	if err == nil {
		t.Fatal("unknown sound key accepted, want a whole-document rejection (D-33)")
	}
	if !strings.Contains(err.Error(), "sound_effect") {
		t.Errorf("error %q does not name the unknown field sound_effect", err)
	}
}
