package main

// Matrix runner corpus (plan 02-06, TEST-04, D-07 headless part): the
// decoder and the exit-code contract are pure logic, so they are pinned
// here without a live desktop; the live step-runner itself is exercised by
// `mise run e2e-matrix` (D-08's green iteration runs this corpus first).

import (
	"errors"
	"testing"
)

// matrixCorpusValid is the Pattern 5 sample of 02-RESEARCH verbatim: the
// D-13 geometry case and the CORR-09 Escape-reset noop — the two shapes the
// whole matrix vocabulary builds on.
const matrixCorpusValid = `name: word-after-space-en-ru
surface: zenity
mode: en
steps:
  - {type: "ghbdtn"}
  - {key: space}
  - {tap: double}
expect_text: "привет "
expect_level: 1
---
name: reset-escape-noop
surface: chromium
mode: en
steps:
  - {type: "ghbdtn"}
  - {key: Escape}
  - {tap: double}
expect_text: "ghbdtn"
`

// TestMatrixDecode_Valid pins the happy path: the Pattern 5 corpus decodes
// into exact field values — names, surfaces, modes, every step kind and the
// snake_case expect_* fields (trailing space of the D-13 oracle included).
func TestMatrixDecode_Valid(t *testing.T) {
	t.Parallel()

	cases, err := loadMatrixCases([]byte(matrixCorpusValid))
	if err != nil {
		t.Fatalf("loadMatrixCases: %v", err)
	}
	if len(cases) != 2 {
		t.Fatalf("decoded %d cases, want 2", len(cases))
	}

	t.Run("word-after-space-en-ru", func(t *testing.T) {
		t.Parallel()

		assertAfterSpaceCase(t, cases[0])
	})

	t.Run("reset-escape-noop", func(t *testing.T) {
		t.Parallel()

		second := cases[1]
		if second.Name != "reset-escape-noop" {
			t.Errorf("Name = %q, want %q", second.Name, "reset-escape-noop")
		}
		if second.Surface != "chromium" {
			t.Errorf("Surface = %q, want %q", second.Surface, "chromium")
		}
		if second.Steps[1].Key != "Escape" {
			t.Errorf("Steps[1].Key = %q, want %q", second.Steps[1].Key, "Escape")
		}
		if second.ExpectLevel != 0 {
			t.Errorf("ExpectLevel = %d, want 0 (unset)", second.ExpectLevel)
		}
	})
}

// assertAfterSpaceCase pins the D-13 corpus case field by field.
func assertAfterSpaceCase(t *testing.T, first matrixCase) {
	t.Helper()

	if first.Name != "word-after-space-en-ru" {
		t.Errorf("Name = %q, want %q", first.Name, "word-after-space-en-ru")
	}
	if first.Surface != "zenity" || first.Mode != "en" {
		t.Errorf("Surface/Mode = %q/%q, want zenity/en", first.Surface, first.Mode)
	}
	if len(first.Steps) != 3 {
		t.Fatalf("Steps = %d, want 3", len(first.Steps))
	}
	if first.Steps[0].Type != "ghbdtn" {
		t.Errorf("Steps[0].Type = %q, want %q", first.Steps[0].Type, "ghbdtn")
	}
	if first.Steps[1].Key != "space" {
		t.Errorf("Steps[1].Key = %q, want %q", first.Steps[1].Key, "space")
	}
	if first.Steps[2].Tap != "double" {
		t.Errorf("Steps[2].Tap = %q, want %q", first.Steps[2].Tap, "double")
	}
	if first.ExpectText != "привет " {
		t.Errorf("ExpectText = %q, want %q", first.ExpectText, "привет ")
	}
	if first.ExpectLevel != 1 {
		t.Errorf("ExpectLevel = %d, want 1", first.ExpectLevel)
	}
}

// TestMatrixDecode_UnknownFieldRejected pins the strict schema (T-02-06-01):
// a typo'd field must fail decoding loudly, never pass silently — the
// KnownFields contract of 02-RESEARCH "Don't Hand-Roll".
func TestMatrixDecode_UnknownFieldRejected(t *testing.T) {
	t.Parallel()

	const corpus = `name: typo-case
surface: zenity
mode: en
steps:
  - {type: "ghbdtn"}
expct_text: "привет"
`
	if _, err := loadMatrixCases([]byte(corpus)); err == nil {
		t.Fatal("decode with unknown field expct_text succeeded, want an error")
	}
}

// TestMatrixDecode_MissingRequiredRejected pins the mandatory fields: a
// case without steps, without a name or without expect_text is rejected by
// validation, and a step carrying no (or more than one) kind is an error —
// the "exactly one step kind" rule. Corpora ride YAML flow mappings: one
// case per line keeps the rejection table scannable.
func TestMatrixDecode_MissingRequiredRejected(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		corpus string
	}{
		{
			name:   "no steps",
			corpus: `{name: no-steps, surface: zenity, mode: en, expect_text: "привет"}`,
		},
		{
			name:   "no name",
			corpus: `{surface: zenity, mode: en, steps: [{type: ghbdtn}], expect_text: "привет"}`,
		},
		{
			name:   "no expect_text",
			corpus: `{name: no-expect, surface: zenity, mode: en, steps: [{type: ghbdtn}]}`,
		},
		{
			name:   "step without a kind",
			corpus: `{name: empty-step, surface: zenity, mode: en, steps: [{}], expect_text: "привет"}`,
		},
		{
			name: "step with two kinds",
			corpus: "{name: greedy-step, surface: zenity, mode: en," +
				` steps: [{type: ghbdtn, tap: double}], expect_text: "привет"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := loadMatrixCases([]byte(tc.corpus)); err == nil {
				t.Fatalf("decode of %q succeeded, want an error", tc.name)
			}
		})
	}
}

// TestMatrixDecode_MultiDoc pins the `---`-separated multi-document
// loading: three well-formed cases in one file decode to exactly three.
func TestMatrixDecode_MultiDoc(t *testing.T) {
	t.Parallel()

	const corpus = `name: one
surface: zenity
mode: en
steps:
  - {type: "a"}
expect_text: "ф"
---
name: two
surface: chromium
mode: en
steps:
  - {type: "b"}
expect_text: "и"
---
name: three
surface: gnome-text-editor
mode: ru
steps:
  - {type: "c"}
expect_text: "с"
`
	cases, err := loadMatrixCases([]byte(corpus))
	if err != nil {
		t.Fatalf("loadMatrixCases: %v", err)
	}
	if len(cases) != 3 {
		t.Fatalf("decoded %d cases, want 3", len(cases))
	}
	if cases[2].Surface != "gnome-text-editor" || cases[2].Mode != "ru" {
		t.Errorf("third case = %q/%q, want gnome-text-editor/ru", cases[2].Surface, cases[2].Mode)
	}
}

// TestMatrixStepValidation pins the closed vocabularies of the schema: tap
// accepts only single|double|triple, surface only the three driver names,
// mode only en|ru — everything else fails validation with the field named.
func TestMatrixStepValidation(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		corpus string
	}{
		{
			name:   "tap quadruple",
			corpus: `{name: v, surface: zenity, mode: en, steps: [{tap: quadruple}], expect_text: x}`,
		},
		{
			name:   "surface terminal",
			corpus: `{name: v, surface: terminal, mode: en, steps: [{type: ghbdtn}], expect_text: x}`,
		},
		{
			name:   "mode xx",
			corpus: `{name: v, surface: zenity, mode: xx, steps: [{type: ghbdtn}], expect_text: x}`,
		},
		{
			name:   "focus unknown surface",
			corpus: `{name: v, surface: chromium, mode: en, steps: [{focus: terminal}], expect_text: x}`,
		},
		{
			name:   "key unknown name",
			corpus: `{name: v, surface: chromium, mode: en, steps: [{key: WakeUp}], expect_text: x}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := loadMatrixCases([]byte(tc.corpus)); err == nil {
				t.Fatalf("decode of %q succeeded, want an error", tc.name)
			}
		})
	}
}

// TestMatrixReport_ExitCode pins the TEST-04 exit contract without a live
// desktop: a report with at least one FAIL computes exit code 1, an
// all-PASS report computes 0.
func TestMatrixReport_ExitCode(t *testing.T) {
	t.Parallel()

	var green matrixReport
	green.add("word-en-ru", nil)
	green.add("reset-escape", nil)
	if code := green.exitCode(); code != 0 {
		t.Errorf("all-PASS exitCode = %d, want 0", code)
	}

	var mixed matrixReport
	mixed.add("word-en-ru", nil)
	mixed.add("reset-escape", errors.New(`expected "ghbdtn", got "привет"`))
	if code := mixed.exitCode(); code != 1 {
		t.Errorf("report with one FAIL exitCode = %d, want 1", code)
	}
	if mixed.failed() != 1 || mixed.passed() != 1 {
		t.Errorf("counts = %d failed / %d passed, want 1/1", mixed.failed(), mixed.passed())
	}
}
