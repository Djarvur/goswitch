package main

import (
	"slices"
	"strings"
	"testing"
)

// TestParseBlocklistDoc pins the blocklist case's file-oracle decoder: the
// persisted document's autocorrect.enabled and apps_blocklist decode from
// the exact bytes the case writes (the start document) and from the exact
// bytes the writer's Node round-trip lands after the click (the flow
// one-element form, verified against the 11-03 encoder).
func TestParseBlocklistDoc(t *testing.T) {
	t.Parallel()

	enabled, list, err := parseBlocklistDoc([]byte(blocklistMenuConfigDoc))
	if err != nil {
		t.Fatalf("parse start document: %v", err)
	}
	if !enabled {
		t.Errorf("start document decoded enabled=false, want true (the silence oracle needs the ON layer)")
	}
	if len(list) != 0 {
		t.Errorf("start document decoded list %q, want empty", list)
	}

	appended := strings.Replace(blocklistMenuConfigDoc,
		"apps_blocklist: []", "apps_blocklist: [org.gtk.application.gs_e2e_password]", 1)
	enabled, list, err = parseBlocklistDoc([]byte(appended))
	if err != nil {
		t.Fatalf("parse appended document: %v", err)
	}
	if !enabled {
		t.Errorf("appended document decoded enabled=false, want true")
	}
	if want := []string{"org.gtk.application.gs_e2e_password"}; !slices.Equal(list, want) {
		t.Errorf("appended document decoded list %q, want %q", list, want)
	}
}

// TestAssertOnlyBlocklistLine pins the byte-footprint oracle on the exact
// before/after pair the writer's round-trip produces (verified against the
// 11-03 encoder): exactly the empty-list line swapped for the one-element
// line carrying the pattern; every other line byte-identical. Any extra
// change — a reordered section, a reformatted comment, a blank line —
// must fail the oracle.
func TestAssertOnlyBlocklistLine(t *testing.T) {
	t.Parallel()

	const app = "org.gtk.application.gs_e2e_password"
	before := []byte(blocklistMenuConfigDoc)
	after := []byte(strings.Replace(blocklistMenuConfigDoc,
		"apps_blocklist: []", "apps_blocklist: ["+app+"]", 1))

	if err := assertOnlyBlocklistLine(before, after, app); err != nil {
		t.Fatalf("clean one-line swap refused: %v", err)
	}

	// A second touched line (the sound scalar flipped) must fail — the
	// click may never move anything but the blocklist line.
	drifting := []byte(strings.Replace(string(after), "sound:\n  enabled: false", "sound:\n  enabled: true", 1))
	if err := assertOnlyBlocklistLine(before, drifting, app); err == nil {
		t.Fatalf("two-line diff accepted — the footprint oracle is deaf")
	}

	// The pattern in the added line is pinned verbatim: a different
	// element (the un-meta'd name when QuoteMeta was skipped) must fail.
	otherPattern := []byte(strings.Replace(blocklistMenuConfigDoc,
		"apps_blocklist: []", "apps_blocklist: [other.app]", 1))
	if err := assertOnlyBlocklistLine(before, otherPattern, app); err == nil {
		t.Fatalf("wrong pattern accepted — the oracle does not pin the QuoteMeta literal")
	}
}
