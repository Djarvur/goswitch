package main

// Matrix runner stub (RED phase of plan 02-06 Task 1): the schema types and
// the seams the corpus pins, with the loader and the report logic still
// unimplemented — every target test fails on its assertions, never on a
// compile error.

import "errors"

// matrixStep is one step of a case: exactly one of the four fields is set
// (validated), the step kind it names decides the runner's action.
type matrixStep struct {
	Type  string `yaml:"type"`
	Key   string `yaml:"key"`
	Tap   string `yaml:"tap"`
	Focus string `yaml:"focus"`
}

// matrixCase is one "type → expectation" case of the YAML matrix (TEST-04,
// 02-RESEARCH Pattern 5 schema).
type matrixCase struct {
	Name        string       `yaml:"name"`
	Surface     string       `yaml:"surface"`
	Mode        string       `yaml:"mode"`
	Steps       []matrixStep `yaml:"steps"`
	ExpectText  string       `yaml:"expect_text"`
	ExpectLevel uint8        `yaml:"expect_level"`
}

// loadMatrixCases decodes the strict multi-document YAML matrix. Stub: the
// GREEN phase implements the yaml.v3 decoder with KnownFields(true) and the
// schema validation.
func loadMatrixCases(data []byte) ([]matrixCase, error) {
	return nil, errors.New("matrix loader not implemented (RED stub)")
}

// matrixEntry is one case outcome: err == nil means PASS.
type matrixEntry struct {
	name string
	err  error
}

// matrixReport accumulates per-case outcomes and computes the run's exit
// code (TEST-04: non-zero on any FAIL).
type matrixReport struct {
	entries []matrixEntry
}

// add records one case outcome.
func (r *matrixReport) add(name string, err error) {
	r.entries = append(r.entries, matrixEntry{name: name, err: err})
}

// passed counts PASS entries.
func (r *matrixReport) passed() int { return 0 } // stub

// failed counts FAIL entries.
func (r *matrixReport) failed() int { return 0 } // stub

// exitCode computes the run's exit code. Stub: the GREEN phase implements
// the "any FAIL → 1" contract.
func (r *matrixReport) exitCode() int { return 0 }
