// The persist writer: menu-toggle persistence into the user's YAML
// document (plan 07-03). The tracer stage carries the schema shape only —
// the flip path lands with the tracer GREEN (the RED-stub precedent of
// 06-03: the RED commit carries the shape, GREEN the behavior).
package config

import (
	"errors"
	"fmt"
)

// errWriterStub is the tracer-stage refusal — every toggle attempt fails
// loudly until the flip path is implemented, never a silent no-op.
var errWriterStub = errors.New("writer: not implemented")

// SetAutocorrectEnabled persists the autocorrect master switch into the
// YAML document at path — the menu toggle's storage (plan 07-05 is the
// first caller). The document's comments and structure survive the write;
// an absent file is the ensure branch (Task 2).
func SetAutocorrectEnabled(path string, on bool) error {
	return fmt.Errorf("set autocorrect.enabled = %v in %s: %w", on, path, errWriterStub)
}
