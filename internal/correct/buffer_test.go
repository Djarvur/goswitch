package correct_test

import (
	"testing"

	"github.com/Djarvur/goswitch/internal/correct"
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
	push(b, "ghbdtn")
	if got := string(b.Token()); got != "ghbdtn" {
		t.Fatalf("Token() = %q, want %q", got, "ghbdtn")
	}
	if got := string(b.Tail()); got != "" {
		t.Fatalf("Tail() = %q, want empty", got)
	}
	b.Push(' ')
	if got := string(b.Token()); got != "ghbdtn" {
		t.Errorf("Token() after space = %q, want %q (D-13: the last word lives on)", got, "ghbdtn")
	}
	if got := string(b.Tail()); got != " " {
		t.Errorf("Tail() after space = %q, want %q", got, " ")
	}
	b.Backspace()
	if got := string(b.Token()); got != "ghbdtn" {
		t.Errorf("Token() after Backspace over the space = %q, want %q", got, "ghbdtn")
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
		push(b, "ghbdtn")
		b.Push(' ')
		if got := string(b.Token()); got != "ghbdtn" {
			t.Errorf("Token() = %q, want %q", got, "ghbdtn")
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
		{name: "comma joins the token", in: "ghbdtn,", token: "ghbdtn,", tail: ""},
		{name: "semicolon joins the token", in: "ghbdtn;", token: "ghbdtn;", tail: ""},
		{name: "slash is a boundary", in: "ghbdtn/", token: "ghbdtn", tail: "/"},
		{name: "dash is a boundary", in: "ghbdtn-", token: "ghbdtn", tail: "-"},
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
		push(b, "ghbdtn2026")
		if got := string(b.Token()); got != "ghbdtn2026" {
			t.Errorf("Token() = %q, want %q", got, "ghbdtn2026")
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
	if got := string(b.Token()); got != "ghbdtn" {
		t.Errorf("Token() after Backspace = %q, want %q (back to the previous word)", got, "ghbdtn")
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

// TestBuffer_ReplaceTokenToggle pins the toggle invariant: after a
// successful correction the buffer holds the corrected token with the tail
// preserved, so the buffer keeps equal to the text before the cursor and a
// repeated correction converts the word back.
func TestBuffer_ReplaceTokenToggle(t *testing.T) {
	t.Parallel()

	b := correct.NewBuffer()
	push(b, "ghbdtn ")
	b.ReplaceToken([]rune("привет"))
	if got := string(b.Token()); got != "привет" {
		t.Fatalf("Token() after ReplaceToken = %q, want %q", got, "привет")
	}
	if got := string(b.Tail()); got != " " {
		t.Errorf("Tail() after ReplaceToken = %q, want %q (tail preserved)", got, " ")
	}
	dir, ok := correct.Detect(b.Token())
	if !ok || dir != correct.RUtoEN {
		t.Fatalf("Detect(corrected token) = (%d, %v), want (%d, true)", dir, ok, correct.RUtoEN)
	}
	back, ok := correct.Convert(b.Token(), dir)
	if !ok || string(back) != "ghbdtn" {
		t.Errorf("Convert(corrected token) = (%q, %v), want (ghbdtn, true)", string(back), ok)
	}
}
