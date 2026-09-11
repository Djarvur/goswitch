package correct_test

import (
	"testing"

	"github.com/Djarvur/goswitch/internal/correct"
)

// TestMatchesSuffix pins the ADR-004 suffix rule: the text before the
// cursor must END with the token. A tail after the token inside the
// surrounding text is a mismatch — the verification waits for the token to
// be last — and text shorter than the token is not a match either.
func TestMatchesSuffix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		text  string
		token string
		want  bool
	}{
		{name: "ends with token", text: "abc " + wordEN, token: wordEN, want: true},
		{name: "tail after token in text", text: "abc " + wordEN + " ", token: wordEN, want: false},
		{name: "text shorter than token", text: "gfb", token: wordEN, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := correct.MatchesSuffix([]rune(tc.text), []rune(tc.token)); got != tc.want {
				t.Errorf("MatchesSuffix(%q, %q) = %v, want %v", tc.text, tc.token, got, tc.want)
			}
		})
	}
}
