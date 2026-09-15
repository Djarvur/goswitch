package correct_test

import (
	"testing"

	"github.com/Djarvur/goswitch/internal/correct"
)

// The golden mixed-text corpus of Phase 3 (03-CONTEXT specifics, D-22/D-23)
// plus the homogeneous and refusal families. Every case names its decision.
func TestConvertRuns_GoldenCorpus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
		why  string
	}{
		{
			// D-22: якорь RU от последней буквы (…вет) — чужой прогон ghbdtn
			// конвертируется, свой не трогается (D-23).
			name: "latin then cyrillic, anchor RU",
			in:   wordEN + wordRU,
			want: wordRU + wordRU,
			why:  "D-22/D-23",
		},
		{
			// D-23: granular per-run conversion — gfb→паи by key position.
			name: "short latin run, anchor RU",
			in:   "gfb" + wordRU,
			want: "паи" + wordRU,
			why:  "D-23",
		},
		{
			// D-22: якорь EN — теперь чужой прогон русский, конвертируется он.
			name: "cyrillic then latin, anchor EN",
			in:   wordRU + wordEN,
			want: wordEN + wordEN,
			why:  "D-22",
		},
		{
			// D-15/D-22: цифры нейтральны и якорь не сдвигают — якорь RU от
			// последней буквы, конвертируется только ghbdtn.
			name: "digits are neutral to the anchor",
			in:   wordEN + "2026" + wordRU,
			want: wordRU + "2026" + wordRU,
			why:  "D-15/D-22",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, changed, ok := correct.ConvertRuns([]rune(tc.in))
			if !ok {
				t.Fatalf("ConvertRuns(%q) ok = false, want true (%s)", tc.in, tc.why)
			}
			if !changed {
				t.Fatalf("ConvertRuns(%q) changed = false, want true — a foreign run exists (%s)", tc.in, tc.why)
			}
			if got := string(out); got != tc.want {
				t.Errorf("ConvertRuns(%q) = %q, want %q (%s)", tc.in, got, tc.want, tc.why)
			}
		})
	}
}

// TestConvertRuns_HomogeneousWholesale pins the уточнение D-22 (2026-09-15):
// the anchor governs ONLY mixed text — a single-script range converts
// WHOLESALE by its composition, the Phase 2 behavior that keeps matrix v1
// green (ghbdtn→привет, привет→ghbdtn). Register is preserved per rune
// through the same generated tables (CORR-05).
//
// D-24 resolution (deviation, plan 03-01 task 2): the plan's
// nothing-to-convert example ("привет", якорь RU → out==input) presumes the
// strict anchor applying to homogeneous ranges — the exact reading the
// уточнение D-22 rejects ("строгий якорь везде превращает базовый кейс
// ghbdtn→привет в no-op"). Under the locked semantics every word/phrase
// range resolves to a conversion, so ConvertRuns's changed=false success
// surface is reserved for ranges anchored from outside their own
// composition (the selection path of plan 03-03); it is implemented, and
// the D-20 refusal family stays reserved for no-letters and unmapped
// letters alone — pinned by TestConvertRuns_Refusals below.
func TestConvertRuns_HomogeneousWholesale(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "latin word converts wholesale", in: wordEN, want: wordRU},
		{name: "cyrillic word converts wholesale", in: wordRU, want: wordEN},
		{name: "digits ride along", in: wordENDigits, want: wordRU + "2026"},
		// The register pin on a MIXED range: the foreign run converts with
		// its own per-rune register (G→П, f→а, b→и).
		{name: "register per rune on the foreign run", in: "Gfb" + wordRU, want: "Паи" + wordRU},
		{name: "punctuation of the token rides along", in: wordENComma, want: wordRU + ","},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, changed, ok := correct.ConvertRuns([]rune(tc.in))
			if !ok || !changed {
				t.Fatalf("ConvertRuns(%q) = (%q, %v, %v), want a wholesale conversion", tc.in, string(out), changed, ok)
			}
			if got := string(out); got != tc.want {
				t.Errorf("ConvertRuns(%q) = %q, want %q (уточнение D-22)", tc.in, got, tc.want)
			}
		})
	}
}

// TestConvertRuns_Refusals pins the D-20 refusal family of the run
// conversion: a range with no letters at all has nothing to anchor on, and
// an unmapped letter in a run that must convert fails the WHOLE range — a
// partial conversion (out with some runes already mapped) must never leak.
func TestConvertRuns_Refusals(t *testing.T) {
	t.Parallel()

	t.Run("no letters", func(t *testing.T) {
		t.Parallel()
		out, _, ok := correct.ConvertRuns([]rune("2026 ,.!"))
		if ok {
			t.Fatalf("ConvertRuns(letterless) ok = true, want false — no direction to convert to")
		}
		if out != nil {
			t.Errorf("ConvertRuns(letterless) out = %q, want nil — nothing was converted", string(out))
		}
	})

	t.Run("unmapped letter fails wholesale", func(t *testing.T) {
		t.Parallel()
		// 'é' is a Latin-script letter with no key position in either
		// layout — Convert must refuse it, and the refusal must abort the
		// whole range before any rune is emitted (T-03-01-03).
		out, _, ok := correct.ConvertRuns([]rune("ghé" + wordRU))
		if ok {
			t.Fatalf("ConvertRuns(mixed with unmapped 'é') ok = true, want false")
		}
		if out != nil {
			t.Errorf("partial conversion leaked: out = %q, want nil — abort, не мусорить", string(out))
		}
	})
}
