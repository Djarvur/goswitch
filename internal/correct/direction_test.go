package correct_test

import (
	"testing"

	"github.com/Djarvur/goswitch/internal/correct"
)

// TestDetect_PureScript pins CORR-04: the direction comes from the letter
// script of the token alone, digits ride along neutrally, and a token with
// no letters has no direction.
func TestDetect_PureScript(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		token string
		dir   correct.Dir
		ok    bool
	}{
		{name: "latin word", token: "ghbdtn", dir: correct.ENtoRU, ok: true},
		{name: "cyrillic word", token: "привет", dir: correct.RUtoEN, ok: true},
		{name: "latin word with digits", token: "ghbdtn2026", dir: correct.ENtoRU, ok: true},
		{name: "no letters at all", token: "2026", dir: 0, ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir, ok := correct.Detect([]rune(tc.token))
			if dir != tc.dir || ok != tc.ok {
				t.Errorf("Detect(%q) = (%d, %v), want (%d, %v)", tc.token, dir, ok, tc.dir, tc.ok)
			}
		})
	}
}

// TestDetect_MixedRefused pins D-16: any single foreign letter makes the
// word mixed and refuses it — the strict threshold, word untouched.
func TestDetect_MixedRefused(t *testing.T) {
	t.Parallel()

	for _, token := range []string{"gfbпривет", "приветg", "ghbdtnп"} {
		t.Run(token, func(t *testing.T) {
			t.Parallel()
			dir, ok := correct.Detect([]rune(token))
			if ok || dir != 0 {
				t.Errorf("Detect(%q) = (%d, %v), want (0, false): one foreign letter refuses the word", token, dir, ok)
			}
		})
	}
}

// TestDetect_NeutralNotMixed pins D-15/D-16: digits and punctuation are
// neutral — they never make a word mixed.
func TestDetect_NeutralNotMixed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		token string
		dir   correct.Dir
	}{
		{name: "latin with digits", token: "ghbdtn2026", dir: correct.ENtoRU},
		{name: "cyrillic with digits", token: "привет2026", dir: correct.RUtoEN},
		{name: "latin with token punctuation", token: "ghbdtn,;", dir: correct.ENtoRU},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir, ok := correct.Detect([]rune(tc.token))
			if !ok || dir != tc.dir {
				t.Errorf("Detect(%q) = (%d, %v), want (%d, true): neutrals do not mix", tc.token, dir, ok, tc.dir)
			}
		})
	}
}

// TestDetect_BothShiftLevels pins that both shift levels classify by script
// — the register never changes the direction (CORR-04/CORR-05).
func TestDetect_BothShiftLevels(t *testing.T) {
	t.Parallel()

	t.Run("mixed register latin", func(t *testing.T) {
		t.Parallel()
		dir, ok := correct.Detect([]rune("GHBdtN"))
		if !ok || dir != correct.ENtoRU {
			t.Errorf("Detect(\"GHBdtN\") = (%d, %v), want (ENtoRU, true)", dir, ok)
		}
	})

	t.Run("mixed register cyrillic", func(t *testing.T) {
		t.Parallel()
		dir, ok := correct.Detect([]rune("ПриВет"))
		if !ok || dir != correct.RUtoEN {
			t.Errorf("Detect(\"ПриВет\") = (%d, %v), want (RUtoEN, true)", dir, ok)
		}
	})
}
