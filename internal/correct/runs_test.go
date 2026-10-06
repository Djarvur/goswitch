package correct_test

import (
	"testing"

	"github.com/Djarvur/goswitch/internal/correct"
)

// The golden mixed-text corpus (spec-delta 2026-10-06; owner verdict, UAT
// of phase 6, 2026-10-03, commit 7dd46e9): mixed text corrects by
// PER-CHARACTER layout inversion — every Latin letter maps through ENToRU,
// every Cyrillic letter through RUToEN, no anchor, no last-letter rule
// (the D-22/D-23 anchor semantics are superseded; the WINDOWS-12 readback
// freeze was rejected by the owner). The homogeneous and refusal families
// below carry the preserved behavior.
func TestConvertRuns_MixedInvertsPerChar(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
		why  string
	}{
		{
			// Every letter inverts: ghbdtn→привет AND привет→ghbdtn — the
			// "flip every character" intent (supersedes the D-22 anchor).
			name: "latin then cyrillic, both invert",
			in:   wordEN + wordRU,
			want: wordRU + wordEN,
			why:  "spec-delta 2026-10-06 (supersedes D-22/D-23)",
		},
		{
			// The owner-verdict case: gfb→паи AND привет→ghbdtn in one
			// pass — each letter through its own script's table.
			name: "short latin run, both scripts invert",
			in:   "gfb" + wordRU,
			want: "паиghbdtn",
			why:  "spec-delta 2026-10-06",
		},
		{
			// Mirror direction: the Cyrillic head inverts to Latin, the
			// Latin tail to Cyrillic — no anchor on the last letter.
			name: "cyrillic then latin, both invert",
			in:   wordRU + wordEN,
			want: wordEN + wordRU,
			why:  "spec-delta 2026-10-06 (supersedes D-22/D-23)",
		},
		{
			// Digits are self-inverse ('5'→'5' in both tables) — they ride
			// while every letter around them inverts.
			name: "digits ride, letters all invert",
			in:   wordEN + "2026" + wordRU,
			want: wordRU + "2026" + wordEN,
			why:  "spec-delta 2026-10-06: the neutral rule",
		},
		{
			// Owner decision 1 (260927-vu8): '[' has no per-character
			// direction — in a MIXED range it rides as typed; only a
			// wholesale single-script range converts it.
			name: "allow-set bracket rides in a mixed range",
			in:   "gfb" + wordRU + "[",
			want: "паиghbdtn[",
			why:  "spec-delta 2026-10-06: the neutral rule",
		},
		{
			// Register is preserved per rune through the generated tables
			// (CORR-05): G→П, f→а, b→и, and привет→ghbdtn back.
			name: "register per rune through the inversion",
			in:   "Gfb" + wordRU,
			want: "Паиghbdtn",
			why:  "CORR-05 under spec-delta 2026-10-06",
		},
		{
			// Space has no key position — its inversion is itself; the
			// letters on both sides still invert.
			name: "space rides, letters on both sides invert",
			in:   "gfb " + wordRU,
			want: "паи ghbdtn",
			why:  "spec-delta 2026-10-06: the neutral rule",
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
				t.Fatalf("ConvertRuns(%q) changed = false, want true — a mixed range always converts (%s)", tc.in, tc.why)
			}
			if got := string(out); got != tc.want {
				t.Errorf("ConvertRuns(%q) = %q, want %q (%s)", tc.in, got, tc.want, tc.why)
			}
		})
	}
}

// TestConvertRuns_HomogeneousWholesale pins the single-script half of the
// conversion (уточнение D-22, 2026-09-15; carried over byte-identically by
// the spec-delta of 2026-10-06): a single-script range converts WHOLESALE
// by its composition — the Phase 2 behavior that keeps matrix v1 green
// (ghbdtn→привет, привет→ghbdtn). Register is preserved per rune through
// the same generated tables (CORR-05); punctuation converts with the token
// — the wholesale surface the mixed range's neutral rule deliberately does
// not share.
//
// D-24: ConvertRuns's changed=false success surface is reserved for ranges
// arriving from outside their own composition (the selection path of plan
// 03-03); it is implemented, and the D-20 refusal family stays reserved
// for no-letters and unmapped letters alone — pinned by
// TestConvertRuns_Refusals below.
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
		// Full by-position conversion (owner directive 2026-09-28): the
		// token's punctuation converts with it — the comma is the 'б' key.
		{name: "punctuation converts with the token", in: wordENComma, want: wordRU + "б"},
		// Owner decision 1 (260927-vu8): in a WHOLESALE single-script range
		// the allow-set bracket-row symbol converts with the token — the
		// run-level form of the live defect «несколько дополнительны[».
		{name: "bracket converts with the token wholesale", in: wordEN + "[", want: wordRU + "х"},
		// The RU direction has no '[' entry — an intentional EN symbol rides
		// as typed (the allow-set converts EN→RU only).
		{name: "en bracket rides in the ru direction", in: wordRU + "[", want: wordEN + "["},
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
			t.Fatalf("ConvertRuns(letterless) ok = false, want false — no direction to convert to")
		}
		if out != nil {
			t.Errorf("ConvertRuns(letterless) out = %q, want nil — nothing was converted", string(out))
		}
	})

	t.Run("lone bracket is no letters", func(t *testing.T) {
		t.Parallel()
		// Owner decision 1 (260927-vu8): even an allow-set symbol alone has
		// nothing to anchor on — the D-20 no-letters refusal stands.
		out, _, ok := correct.ConvertRuns([]rune("["))
		if ok {
			t.Fatalf("ConvertRuns(lone '[') ok = true, want false — no letters, no anchor")
		}
		if out != nil {
			t.Errorf("ConvertRuns(lone '[') out = %q, want nil — nothing was converted", string(out))
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
