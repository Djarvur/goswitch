package main

import "testing"

// TestParseMenuToggles pins the menu case's file oracle: the persisted
// document's autocorrect.enabled and sound.enabled decode from the exact
// bytes the toggle writes (the 07-03 Node round-trip keeps every section;
// the daemon flips exactly these two scalars per click, plan 07-05).
func TestParseMenuToggles(t *testing.T) {
	t.Parallel()

	bothOff := []byte(menuConfigDoc)
	ac, sound, err := parseMenuToggles(bothOff)
	if err != nil {
		t.Fatalf("parse menu document: %v", err)
	}
	if ac || sound {
		t.Errorf("start document decoded ac=%t sound=%t, want false/false", ac, sound)
	}

	bothOn := []byte(`hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
autocorrect:
  enabled: true
  apps_blocklist: []
sound:
  enabled: true
`)
	ac, sound, err = parseMenuToggles(bothOn)
	if err != nil {
		t.Fatalf("parse toggled document: %v", err)
	}
	if !ac || !sound {
		t.Errorf("toggled document decoded ac=%t sound=%t, want true/true", ac, sound)
	}

	split := []byte(`autocorrect:
  enabled: true
sound:
  enabled: false
`)
	ac, sound, err = parseMenuToggles(split)
	if err != nil {
		t.Fatalf("parse split document: %v", err)
	}
	if !ac || sound {
		t.Errorf("split document decoded ac=%t sound=%t, want true/false", ac, sound)
	}
}
