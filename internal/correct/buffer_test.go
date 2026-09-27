package correct_test

import (
	"testing"

	"github.com/Djarvur/goswitch/internal/correct"
)

// The SPEC correction word in both layouts, named so every corpus case
// states its contract (style of the timing constants in fsm_test.go).
const (
	wordEN       = "ghbdtn"
	wordRU       = "привет"
	wordENDigits = "ghbdtn2026"
	wordENComma  = "ghbdtn,"
)

// push feeds every rune of s into the buffer.
func push(b *correct.Buffer, s string) {
	for _, r := range s {
		b.Push(r)
	}
}

// TestBuffer_TokenRules pins the core token contract: a plain word is the
// token with an empty tail; a space finishes the word but keeps it in the
// buffer as the correction target (D-13); Backspace pops that space;
// HardReset empties everything.
func TestBuffer_TokenRules(t *testing.T) {
	t.Parallel()

	b := correct.NewBuffer()
	push(b, wordEN)
	if got := string(b.Token()); got != wordEN {
		t.Fatalf("Token() = %q, want %q", got, wordEN)
	}
	if got := string(b.Tail()); got != "" {
		t.Fatalf("Tail() = %q, want empty", got)
	}
	b.Push(' ')
	if got := string(b.Token()); got != wordEN {
		t.Errorf("Token() after space = %q, want %q (D-13: the last word lives on)", got, wordEN)
	}
	if got := string(b.Tail()); got != " " {
		t.Errorf("Tail() after space = %q, want %q", got, " ")
	}
	b.Backspace()
	if got := string(b.Token()); got != wordEN {
		t.Errorf("Token() after Backspace over the space = %q, want %q", got, wordEN)
	}
	if got := string(b.Tail()); got != "" {
		t.Errorf("Tail() after Backspace over the space = %q, want empty", got)
	}
	b.HardReset()
	if got := b.Token(); got != nil {
		t.Errorf("Token() after HardReset = %q, want nil", string(got))
	}
}

// TestBuffer_TwoTokens pins that the current token wins while it is being
// typed, and that a trailing separator hands the target back to the last
// finished word with the separator as its tail.
func TestBuffer_TwoTokens(t *testing.T) {
	t.Parallel()

	t.Run("current token wins", func(t *testing.T) {
		t.Parallel()
		b := correct.NewBuffer()
		push(b, "ghbdtn privet")
		if got := string(b.Token()); got != "privet" {
			t.Errorf("Token() = %q, want %q", got, "privet")
		}
		if got := string(b.Tail()); got != "" {
			t.Errorf("Tail() = %q, want empty", got)
		}
	})

	t.Run("last word lives after trailing space", func(t *testing.T) {
		t.Parallel()
		b := correct.NewBuffer()
		push(b, "ghbdtn privet ")
		if got := string(b.Token()); got != "privet" {
			t.Errorf("Token() = %q, want %q", got, "privet")
		}
		if got := string(b.Tail()); got != " " {
			t.Errorf("Tail() = %q, want %q", got, " ")
		}
	})
}

// TestBuffer_WordAfterSpace pins the D-13 lifecycle of a word around a
// separator: separated but still correctable, then superseded by the next
// word being typed, which in turn survives its own trailing separator.
func TestBuffer_WordAfterSpace(t *testing.T) {
	t.Parallel()

	t.Run("word then space", func(t *testing.T) {
		t.Parallel()
		b := correct.NewBuffer()
		push(b, wordEN)
		b.Push(' ')
		if got := string(b.Token()); got != wordEN {
			t.Errorf("Token() = %q, want %q", got, wordEN)
		}
		if got := string(b.Tail()); got != " " {
			t.Errorf("Tail() = %q, want %q", got, " ")
		}
	})

	t.Run("next word supersedes", func(t *testing.T) {
		t.Parallel()
		b := correct.NewBuffer()
		push(b, "ghbdtn vj")
		if got := string(b.Token()); got != "vj" {
			t.Errorf("Token() = %q, want %q", got, "vj")
		}
		if got := string(b.Tail()); got != "" {
			t.Errorf("Tail() = %q, want empty", got)
		}
	})

	t.Run("next word then space", func(t *testing.T) {
		t.Parallel()
		b := correct.NewBuffer()
		push(b, "ghbdtn vj ")
		if got := string(b.Token()); got != "vj" {
			t.Errorf("Token() = %q, want %q", got, "vj")
		}
		if got := string(b.Tail()); got != " " {
			t.Errorf("Tail() = %q, want %q", got, " ")
		}
	})
}

// TestBuffer_PunctuationToken pins D-14: a rune whose key is letter-capable
// in either layout (',' shares a key with RU 'б', ';' with RU 'ж') stays
// inside the token; runes that are symbols on both sides ('/', '-') are
// boundaries and land in the tail.
func TestBuffer_PunctuationToken(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		in    string
		token string
		tail  string
	}{
		{name: "comma joins the token", in: wordENComma, token: wordENComma, tail: ""},
		{name: "semicolon joins the token", in: "ghbdtn;", token: "ghbdtn;", tail: ""},
		{name: "slash is a boundary", in: "ghbdtn/", token: wordEN, tail: "/"},
		{name: "dash is a boundary", in: "ghbdtn-", token: wordEN, tail: "-"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b := correct.NewBuffer()
			push(b, tc.in)
			if got := string(b.Token()); got != tc.token {
				t.Errorf("Token() = %q, want %q", got, tc.token)
			}
			if got := string(b.Tail()); got != tc.tail {
				t.Errorf("Tail() = %q, want %q", got, tc.tail)
			}
		})
	}
}

// TestBuffer_DigitsInToken pins D-15: digits are neutral token members —
// they extend the word and even form a token of their own.
func TestBuffer_DigitsInToken(t *testing.T) {
	t.Parallel()

	t.Run("digits extend the word", func(t *testing.T) {
		t.Parallel()
		b := correct.NewBuffer()
		push(b, wordENDigits)
		if got := string(b.Token()); got != wordENDigits {
			t.Errorf("Token() = %q, want %q", got, wordENDigits)
		}
	})

	t.Run("digits alone form a token", func(t *testing.T) {
		t.Parallel()
		b := correct.NewBuffer()
		push(b, "2026")
		if got := string(b.Token()); got != "2026" {
			t.Errorf("Token() = %q, want %q", got, "2026")
		}
	})
}

// TestBuffer_BackspacePop pins the honest pop: removing the last rune of a
// one-rune token hands the target back to the previous word — with the
// separator intact in the tail, because the buffer must keep mirroring the
// field ("ghbdtn " after the pop) and the token+tail replacement geometry
// of CORR-07 depends on that separator — and eight pops empty the phrase.
func TestBuffer_BackspacePop(t *testing.T) {
	t.Parallel()

	b := correct.NewBuffer()
	push(b, "ghbdtn v")
	b.Backspace()
	if got := string(b.Token()); got != wordEN {
		t.Errorf("Token() after Backspace = %q, want %q (back to the previous word)", got, wordEN)
	}
	if got := string(b.Tail()); got != " " {
		t.Errorf("Tail() after Backspace = %q, want %q (the separator stays)", got, " ")
	}
	for range 7 {
		b.Backspace()
	}
	if got := b.Token(); got != nil {
		t.Errorf("Token() after 8 backspaces = %q, want nil (buffer empty)", string(got))
	}
	if got := b.Tail(); got != nil {
		t.Errorf("Tail() after 8 backspaces = %q, want nil", string(got))
	}
}

// TestBuffer_PhraseReplace pins the phrase-level API of Phase 3 (CORR-02,
// D-25): Phrase() returns the WHOLE buffer since the last hard reset — both
// words and the separator, not the token — and ReplacePhrase swaps the entire
// [0, len) range, after which recompute() has rederived the token from the
// new contents (the buffer keeps mirroring the field — the same invariant
// ReplaceToken carries for the word range).
func TestBuffer_PhraseReplace(t *testing.T) {
	t.Parallel()

	b := correct.NewBuffer()
	push(b, wordEN+" "+wordEN)
	if got := string(b.Phrase()); got != wordEN+" "+wordEN {
		t.Fatalf("Phrase() = %q, want %q — the whole phrase, not the token", got, wordEN+" "+wordEN)
	}
	// The phrase and the token are different views of the same buffer.
	if got := string(b.Token()); got != wordEN {
		t.Errorf("Token() = %q, want %q — the token stays the last word", got, wordEN)
	}

	b.ReplacePhrase([]rune(wordRU + " " + wordRU))
	if got := string(b.Phrase()); got != wordRU+" "+wordRU {
		t.Fatalf("Phrase() after ReplacePhrase = %q, want %q", got, wordRU+" "+wordRU)
	}
	// recompute invariant: the token is rederived from the replaced phrase —
	// the trailing word of the new contents, ready for the next correction.
	if got := string(b.Token()); got != wordRU {
		t.Errorf("Token() after ReplacePhrase = %q, want %q (recompute from the new runes)", got, wordRU)
	}
	if got := string(b.Tail()); got != "" {
		t.Errorf("Tail() after ReplacePhrase = %q, want empty (no separator after the last word)", got)
	}
}

// TestBuffer_PhraseEmpty pins the phrase view of an empty buffer: no phrase,
// and ReplacePhrase on it is a no-op (the triple-tap empty-buffer refusal is
// decided by the actor before ever reaching here).
func TestBuffer_PhraseEmpty(t *testing.T) {
	t.Parallel()

	b := correct.NewBuffer()
	if got := b.Phrase(); got != nil {
		t.Fatalf("Phrase() of an empty buffer = %q, want nil", string(got))
	}
	b.ReplacePhrase([]rune(wordRU))
	if got := b.Phrase(); got != nil {
		t.Errorf("ReplacePhrase on an empty buffer left Phrase() = %q, want nil", string(got))
	}
}

// TestBuffer_ReplaceTokenToggle pins the toggle invariant: after a
// successful correction the buffer holds the corrected token with the tail
// preserved, so the buffer keeps equal to the text before the cursor and a
// repeated correction converts the word back.
func TestBuffer_ReplaceTokenToggle(t *testing.T) {
	t.Parallel()

	b := correct.NewBuffer()
	push(b, "ghbdtn ")
	b.ReplaceToken([]rune(wordRU))
	if got := string(b.Token()); got != wordRU {
		t.Fatalf("Token() after ReplaceToken = %q, want %q", got, wordRU)
	}
	if got := string(b.Tail()); got != " " {
		t.Errorf("Tail() after ReplaceToken = %q, want %q (tail preserved)", got, " ")
	}
	dir, ok := correct.Detect(b.Token())
	if !ok || dir != correct.RUtoEN {
		t.Fatalf("Detect(corrected token) = (%d, %v), want (%d, true)", dir, ok, correct.RUtoEN)
	}
	back, ok := correct.Convert(b.Token(), dir)
	if !ok || string(back) != wordEN {
		t.Errorf("Convert(corrected token) = (%q, %v), want (ghbdtn, true)", string(back), ok)
	}
}
