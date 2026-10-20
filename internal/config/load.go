package config

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// errEmptyDocument refuses a config file with no document at all: an
// explicit -config must yield a working config, never silent defaults.
var errEmptyDocument = errors.New("empty config document")

// Load reads and validates the YAML config at path. Decoding is strict
// (KnownFields, D-33 — the working pattern of test/e2e/matrix.go:113-116):
// an unknown key rejects the whole document; an empty document is a
// validate-time refusal.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer func() { _ = f.Close() }() // read-only: the close error carries no signal

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true) // D-33: a typo'd key invalidates the whole config.

	var cfg Config
	switch err := dec.Decode(&cfg); {
	case errors.Is(err, io.EOF):
		return nil, fmt.Errorf("validate config: %w", errEmptyDocument)
	case err != nil:
		return nil, fmt.Errorf("decode config: %w", err)
	}
	// The absent-key normalization (plan 10-07, G-10-1): yaml.v3 decodes an
	// ABSENT key to nil and an EXPLICIT `[]` to a non-nil empty slice —
	// the two forms are distinguished exactly here. An absent key IS the
	// default identity anchor (the source was dead by construction — 10-UAT
	// missing #2); the explicit empty list stays the conscious opt-out
	// (classification not configured), and an explicit non-empty list
	// passes as written. The anchor itself is valid by construction (two
	// non-empty RE2 literals), Validate still runs after.
	if cfg.Correction.Terminals == nil {
		cfg.Correction.Terminals = DefaultTerminals()
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}
