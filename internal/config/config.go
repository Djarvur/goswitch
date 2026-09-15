// Package config holds the daemon's YAML configuration: the schema of all
// hotkeys, timeouts, correction and MACR parameters (D-31), strict decoding
// where a typo'd key invalidates the whole document (D-33) and the hot
// reload watcher publishing last-good snapshots (D-32). The package stays
// free of D-Bus and the engine — it resolves bindings through the pure
// internal/hotkey name tables.
package config

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
	BackspaceCap int  `yaml:"backspace_cap"`
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
// them when -config is absent, so they must equal the Phase 2 behavior.
func Defaults() Config {
	return Config{} // RED stub: the corpus pins the documented values.
}

// Validate enforces the ranges and closed vocabularies; every error names
// the offending field, and any violation invalidates the whole config
// (D-33 — never a partially applied file).
func (c Config) Validate() error {
	return nil // RED stub: the corpus pins the rejection behavior.
}
