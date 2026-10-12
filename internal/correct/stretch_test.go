package correct_test

import (
	"testing"

	"github.com/Djarvur/goswitch/internal/correct"
)

// The stretch corpus words (plan 10-02): the SPEC correction word with the
// punctuation of the same wrong-layout run, named so every case states its
// contract (the buffer_test.go style).
const (
	stretchENPunct  = "ytn?"     // the owner's live defect: word + trailing table punctuation
	stretchRUPunct  = "нет,"     // the WHOLE run converts — «нет,», never «нет?»
	stretchHeadPunc = "(ytn"     // the head side: an opening bracket joins the run
	stretchHeadRU   = "(нет"     // and converts with it
	stretchMixPunc  = "gфb?"     // a MIXED-script run with table punctuation
	stretchMixOut   = "пaи?"     // mixed semantics: punctuation rides AS TYPED («a» is the LATIN image of «ф»)
	stretchRuns2026 = "ytn?2026" // digits stay members across the extended run
)

// TestConvertibleStretch_WordWithPunctuation pins the phase-10 defect
// (owner finding 2026-10-12): the convertible stretch of a buffer holding
// «ytn?» is the WHOLE run — word plus the trailing table punctuation — and
// ConvertRuns turns it into «нет,». The token-only geometry of Phase 2
// («ytn» → «нет», «?» riding the tail as typed) is the defect this corpus
// pins out (SPEC §4.2: every table sign of the mono-layout run converts).
func TestConvertibleStretch_WordWithPunctuation(t *testing.T) {
	t.Parallel()

	b := correct.NewBuffer()
	push(b, stretchENPunct)
	if got := string(b.ConvertibleStretch()); got != stretchENPunct {
		t.Fatalf("ConvertibleStretch() = %q, want %q — the trailing «?» carries a table position",
			got, stretchENPunct)
	}
	if got := string(b.StretchTail()); got != "" {
		t.Fatalf("StretchTail() = %q, want empty — the run ends at the buffer end", got)
	}
	out, changed, ok := correct.ConvertRuns(b.ConvertibleStretch())
	if !ok || !changed {
		t.Fatalf("ConvertRuns(%q) = (%q, changed=%v, ok=%v), want a wholesale conversion",
			stretchENPunct, string(out), changed, ok)
	}
	if string(out) != stretchRUPunct {
		t.Fatalf("ConvertRuns(%q) = %q, want %q — «ytn?» converts whole, never «нет?»",
			stretchENPunct, string(out), stretchRUPunct)
	}
}

// TestConvertibleStretch_SpaceStopsRun pins the D-13 invariant byte-for-byte
// through the new geometry: a space carries NO table position, so the run
// stops at the word and the separator stays the tail — «ghbdtn » still
// corrects to «привет » (converted stretch + tail, the old golden shape).
func TestConvertibleStretch_SpaceStopsRun(t *testing.T) {
	t.Parallel()

	b := correct.NewBuffer()
	push(b, wordEN+" ")
	if got := string(b.ConvertibleStretch()); got != wordEN {
		t.Fatalf("ConvertibleStretch() = %q, want %q — a space has no table position and stops the run", got, wordEN)
	}
	if got := string(b.StretchTail()); got != " " {
		t.Fatalf("StretchTail() = %q, want %q — the separator stays the tail (D-13)", got, " ")
	}
	out, changed, ok := correct.ConvertRuns(b.ConvertibleStretch())
	if !ok || !changed {
		t.Fatalf("ConvertRuns(%q) = ok=%v changed=%v, want a wholesale conversion", wordEN, ok, changed)
	}
	if got := string(out) + string(b.StretchTail()); got != wordRU+" " {
		t.Fatalf("converted+tail = %q, want %q — D-13 stays byte-identical", got, wordRU+" ")
	}
}

// TestConvertibleStretch_HeadSidePunctuation pins the head side of the run
// (SPEC §4.2): an opening bracket before the word carries a table position
// («(» maps to «(» by key position), so «(ytn» is ONE run converting to
// «(нет» — the correction must reach left of the token, not only into the
// tail.
func TestConvertibleStretch_HeadSidePunctuation(t *testing.T) {
	t.Parallel()

	b := correct.NewBuffer()
	push(b, stretchHeadPunc)
	if got := string(b.ConvertibleStretch()); got != stretchHeadPunc {
		t.Fatalf("ConvertibleStretch() = %q, want %q — the head-side bracket joins the run", got, stretchHeadPunc)
	}
	if got := string(b.StretchTail()); got != "" {
		t.Fatalf("StretchTail() = %q, want empty", got)
	}
	out, changed, ok := correct.ConvertRuns(b.ConvertibleStretch())
	if !ok || !changed {
		t.Fatalf("ConvertRuns(%q) = ok=%v changed=%v, want a wholesale conversion", stretchHeadPunc, ok, changed)
	}
	if string(out) != stretchHeadRU {
		t.Fatalf("ConvertRuns(%q) = %q, want %q — the bracket converts with the word",
			stretchHeadPunc, string(out), stretchHeadRU)
	}
}

// TestConvertibleStretch_DigitsRemainMembers pins the tokenCapable legacy
// (D-15) under the new geometry: digits keep their membership — inside the
// token they ride the run («ghbdtn2026»), and across an extended boundary
// they stay members too («ytn?2026» is one run: the digits of the next
// token join through the table punctuation between).
func TestConvertibleStretch_DigitsRemainMembers(t *testing.T) {
	t.Parallel()

	t.Run("digits inside the token ride the run", func(t *testing.T) {
		t.Parallel()
		b := correct.NewBuffer()
		push(b, wordENDigits)
		if got := string(b.ConvertibleStretch()); got != wordENDigits {
			t.Fatalf("ConvertibleStretch() = %q, want %q", got, wordENDigits)
		}
		out, changed, ok := correct.ConvertRuns(b.ConvertibleStretch())
		if !ok || !changed || string(out) != wordRU+"2026" {
			t.Fatalf("ConvertRuns(%q) = (%q, %v, %v), want (%q, true, true)",
				wordENDigits, string(out), changed, ok, wordRU+"2026")
		}
	})

	t.Run("digits across the extended boundary stay members", func(t *testing.T) {
		t.Parallel()
		b := correct.NewBuffer()
		push(b, stretchRuns2026)
		if got := string(b.ConvertibleStretch()); got != stretchRuns2026 {
			t.Fatalf("ConvertibleStretch() = %q, want %q — «?2026» joins through the table punctuation",
				got, stretchRuns2026)
		}
		out, changed, ok := correct.ConvertRuns(b.ConvertibleStretch())
		if !ok || !changed || string(out) != stretchRUPunct+"2026" {
			t.Fatalf("ConvertRuns(%q) = (%q, %v, %v), want (%q, true, true)",
				stretchRuns2026, string(out), changed, ok, stretchRUPunct+"2026")
		}
	})
}

// TestConvertibleStretch_MixedPunctRidesAsTyped pins the mixed-script
// semantics (spec-delta 2026-10-06) over the NEW geometry: for a mixed run
// the punctuation still rides AS TYPED — «gфb?» inverts per character into
// «паи?» with the «?» untouched. runs.go is NOT touched by the stretch: the
// converter was always right, only the fed range grows.
func TestConvertibleStretch_MixedPunctRidesAsTyped(t *testing.T) {
	t.Parallel()

	b := correct.NewBuffer()
	push(b, stretchMixPunc)
	if got := string(b.ConvertibleStretch()); got != stretchMixPunc {
		t.Fatalf("ConvertibleStretch() = %q, want %q — the table punctuation joins the mixed run", got, stretchMixPunc)
	}
	out, changed, ok := correct.ConvertRuns(b.ConvertibleStretch())
	if !ok || !changed {
		t.Fatalf("ConvertRuns(%q) = ok=%v changed=%v, want a mixed conversion", stretchMixPunc, ok, changed)
	}
	if string(out) != stretchMixOut {
		t.Fatalf("ConvertRuns(%q) = %q, want %q — mixed semantics: the «?» rides as typed",
			stretchMixPunc, string(out), stretchMixOut)
	}
}

// TestConvertibleStretch_UnmappedRefusesWholeStretch pins T-03-01-03 over
// the new geometry, honestly extended: «é» has no table position in the
// Latin direction but IS letter-capable, so it forms (with the space before
// it stopping the run) a one-rune stretch that ConvertRuns refuses WHOLE —
// no partial output ever leaks from a stretch carrying an unmapped letter.
func TestConvertibleStretch_UnmappedRefusesWholeStretch(t *testing.T) {
	t.Parallel()

	b := correct.NewBuffer()
	push(b, "ytn é")
	if got := string(b.ConvertibleStretch()); got != "é" {
		t.Fatalf("ConvertibleStretch() = %q, want %q — the space stops the run at the unmapped letter", got, "é")
	}
	out, changed, ok := correct.ConvertRuns(b.ConvertibleStretch())
	if ok {
		t.Fatalf("ConvertRuns(%q) = (%q, %v), want the wholesale refusal (ok=false)", "é", string(out), changed)
	}
	if out != nil {
		t.Fatalf("ConvertRuns(%q) out = %q, want nil — no partial conversion leaks", "é", string(out))
	}
}

// TestConvertibleStretch_RegisterPreserved pins CORR-05 over the new
// geometry: the per-rune register preservation is untouched — «YTN?»
// converts to «НЕТ,», the punctuation with it.
func TestConvertibleStretch_RegisterPreserved(t *testing.T) {
	t.Parallel()

	b := correct.NewBuffer()
	push(b, "YTN?")
	if got := string(b.ConvertibleStretch()); got != "YTN?" {
		t.Fatalf("ConvertibleStretch() = %q, want %q", got, "YTN?")
	}
	out, changed, ok := correct.ConvertRuns(b.ConvertibleStretch())
	if !ok || !changed {
		t.Fatalf("ConvertRuns(%q) = ok=%v changed=%v, want a wholesale conversion", "YTN?", ok, changed)
	}
	if string(out) != "НЕТ," {
		t.Fatalf("ConvertRuns(%q) = %q, want %q — the register is preserved per rune", "YTN?", string(out), "НЕТ,")
	}
}

// TestConvertibleStretch_ReplaceStretchCoversRun pins the replace half of
// the stretch: ReplaceStretch swaps the WHOLE run («ytn?» → «нет,») — not
// just the token (which would leave the typed «?» behind the corrected
// word) — and the buffer keeps mirroring the field: the corrected run is
// the next token, and a repeated conversion toggles back to «ytn?».
func TestConvertibleStretch_ReplaceStretchCoversRun(t *testing.T) {
	t.Parallel()

	b := correct.NewBuffer()
	push(b, stretchENPunct)
	b.ReplaceStretch([]rune(stretchRUPunct))
	if got := string(b.Token()); got != stretchRUPunct {
		t.Fatalf("Token() after ReplaceStretch = %q, want %q — the replace covers the whole run", got, stretchRUPunct)
	}
	if got := string(b.Tail()); got != "" {
		t.Fatalf("Tail() after ReplaceStretch = %q, want empty", got)
	}
	dir, ok := correct.Detect(b.Token())
	if !ok || dir != correct.RUtoEN {
		t.Fatalf("Detect(corrected run) = (%d, %v), want (%d, true)", dir, ok, correct.RUtoEN)
	}
	back, ok := correct.Convert(b.Token(), dir)
	if !ok || string(back) != stretchENPunct {
		t.Errorf("Convert(corrected run) = (%q, %v), want (%q, true) — the toggle round-trips",
			string(back), ok, stretchENPunct)
	}
}
