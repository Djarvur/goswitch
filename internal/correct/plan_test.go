package correct_test

import (
	"testing"

	"github.com/Djarvur/goswitch/internal/correct"
)

// assertPlan reports every field mismatch of got against want.
func assertPlan(t *testing.T, got, want correct.Plan) {
	t.Helper()
	if got.Level != want.Level {
		t.Errorf("Level = %d, want %d", got.Level, want.Level)
	}
	if got.Offset != want.Offset {
		t.Errorf("Offset = %d, want %d", got.Offset, want.Offset)
	}
	if got.NChars != want.NChars {
		t.Errorf("NChars = %d, want %d", got.NChars, want.NChars)
	}
	if got.Backspaces != want.Backspaces {
		t.Errorf("Backspaces = %d, want %d", got.Backspaces, want.Backspaces)
	}
	if string(got.Commit) != string(want.Commit) {
		t.Errorf("Commit = %q, want %q", string(got.Commit), string(want.Commit))
	}
}

// TestBuildPlan_BothLevels pins CORR-07 for both ladder levels: with the
// surrounding-text capability bit the plan deletes token+tail and recommits
// converted+tail; without it the backspaces count token+tail and the tail
// rides in the commit; an empty tail degenerates to the bare token range.
func TestBuildPlan_BothLevels(t *testing.T) {
	t.Parallel()

	t.Run("level 1 with tail", func(t *testing.T) {
		t.Parallel()
		got := correct.BuildPlan([]rune(wordEN), []rune(" "), []rune(wordRU), correct.CapSurroundingText)
		assertPlan(t, got, correct.Plan{
			Level:  correct.Level1,
			Offset: -7,
			NChars: 7,
			Commit: []rune(wordRU + " "),
		})
	})

	t.Run("level 2 with tail", func(t *testing.T) {
		t.Parallel()
		got := correct.BuildPlan([]rune(wordEN), []rune(" "), []rune(wordRU), 0)
		assertPlan(t, got, correct.Plan{
			Level:      correct.Level2,
			Backspaces: 7,
			Commit:     []rune(wordRU + " "),
		})
	})

	t.Run("empty tail degenerates", func(t *testing.T) {
		t.Parallel()
		got := correct.BuildPlan([]rune(wordEN), nil, []rune(wordRU), correct.CapSurroundingText)
		assertPlan(t, got, correct.Plan{
			Level:  correct.Level1,
			Offset: -6,
			NChars: 6,
			Commit: []rune(wordRU),
		})
	})
}

// TestBuildPlan_RunesNotBytes pins ADR-003 Pitfall 2: a Cyrillic token of 6
// runes — 12 UTF-8 bytes — counts 6, never 12; the byte count would erase
// twice the text.
func TestBuildPlan_RunesNotBytes(t *testing.T) {
	t.Parallel()

	if len(wordRU) != 12 { // corpus sanity: the byte count the arithmetic must NOT use
		t.Fatalf("corpus sanity: len(%q) = %d bytes, want 12", wordRU, len(wordRU))
	}
	got := correct.BuildPlan([]rune(wordRU), nil, []rune(wordEN), 0)
	if got.Backspaces != 6 {
		t.Errorf("Backspaces = %d, want 6 (runes, not the 12 bytes of %q)", got.Backspaces, wordRU)
	}
}

// TestPlan_TailArithmetic pins the tail geometry of CORR-07 (Pitfall 1):
// the cursor sits after the separator, so both levels delete token+tail
// and recommit converted+tail — deleting exactly the token would strand
// the cursor after the tail and commit the word behind it; a token without
// a tail is the degenerate case. All counts in runes.
func TestPlan_TailArithmetic(t *testing.T) {
	t.Parallel()

	t.Run("level 1 deletes token and tail", func(t *testing.T) {
		t.Parallel()
		got := correct.BuildPlan([]rune(wordEN), []rune(" "), []rune(wordRU), correct.CapSurroundingText)
		assertPlan(t, got, correct.Plan{
			Level:  correct.Level1,
			Offset: -7,
			NChars: 7,
			Commit: []rune(wordRU + " "),
		})
	})

	t.Run("level 2 counts token and tail in runes", func(t *testing.T) {
		t.Parallel()
		got := correct.BuildPlan([]rune(wordRU), []rune(" "), []rune(wordEN), 0)
		assertPlan(t, got, correct.Plan{
			Level:      correct.Level2,
			Backspaces: 7,
			Commit:     []rune(wordEN + " "),
		})
	})

	t.Run("degenerate no tail", func(t *testing.T) {
		t.Parallel()
		got := correct.BuildPlan([]rune(wordRU), nil, []rune(wordEN), 0)
		assertPlan(t, got, correct.Plan{
			Level:      correct.Level2,
			Backspaces: 6,
			Commit:     []rune(wordEN),
		})
	})
}
