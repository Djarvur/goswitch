package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSkillgen_ParseRealConfigTable pins the generator to the real
// docs/CONFIG.md key table (the D-8-10 single source): at least the 20
// documented rows parse (the 2026-10-06 owner revision dropped the a11y
// app-list row — the golden boundary rides the data source), every row
// carries exactly five cells, and every key name is section-qualified — a
// dot in the first cell, with a prefix from the closed section set.
func TestSkillgen_ParseRealConfigTable(t *testing.T) {
	t.Parallel()

	rows, err := parseKeyTable("../../docs/CONFIG.md")
	if err != nil {
		t.Fatalf("parseKeyTable: %v", err)
	}
	if len(rows) < 20 {
		t.Fatalf("parseKeyTable returned %d rows, want at least 20", len(rows))
	}
	sections := []string{"hotkeys.", "timeouts.", "correction.", "macr.", "autocorrect.", "sound.", "a11y."}
	for i, row := range rows {
		if len(row.cells) != keyTableCells {
			t.Errorf("row %d: %d cells, want %d", i, len(row.cells), keyTableCells)

			continue
		}
		key := row.cells[0]
		if !strings.Contains(key, ".") {
			t.Errorf("row %d: key %s is not section-qualified (no dot)", i, key)
		}
		qualified := false
		for _, prefix := range sections {
			if strings.HasPrefix(key, "`"+prefix) {
				qualified = true

				break
			}
		}
		if !qualified {
			t.Errorf("row %d: key %s carries no known section prefix", i, key)
		}
	}
}

// TestSkillgen_MalformedRowHardError pins the dictgen discipline on the
// table: a row with a wrong cell count or an empty required cell is a hard
// generation error naming the offending line — never a silent skip and
// never a partial result.
func TestSkillgen_MalformedRowHardError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		table string
	}{
		{
			name:  "four-cells",
			table: "| `hotkeys.tap_key` | string | `shift_r` | closed table |\n",
		},
		{
			name:  "empty-default-cell",
			table: "| `timeouts.tap_window_ms` | int |  | (0, 2000] | окно различения тапов |\n",
		},
		{
			name:  "key-without-dot",
			table: "| `hotkeys` | string | `shift_r` | closed table | клавиша тапов |\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "CONFIG.md")
			doc := "# Справочник\n\n" + tableHeader + "\n" + tableSeparator + "\n" + tc.table
			if err := os.WriteFile(path, []byte(doc), filePerm); err != nil {
				t.Fatalf("write fixture: %v", err)
			}
			rows, err := parseKeyTable(path)
			if err == nil {
				t.Fatalf("parseKeyTable accepted the malformed table (%d rows), want a hard error", len(rows))
			}
			if rows != nil {
				t.Errorf("parseKeyTable returned %d partial rows alongside the error", len(rows))
			}
			if !strings.Contains(err.Error(), "CONFIG.md:") {
				t.Errorf("error %q does not name the offending file:line", err)
			}
		})
	}
}

// TestSkillgen_RenderDeterministic pins the emission: two renders of the
// same rows are byte-equal, and the output carries the region title, the
// source signature, the canonical table header, and every row verbatim.
func TestSkillgen_RenderDeterministic(t *testing.T) {
	t.Parallel()

	rows, err := parseKeyTable("../../docs/CONFIG.md")
	if err != nil {
		t.Fatalf("parseKeyTable: %v", err)
	}
	first := renderRegion(rows)
	second := renderRegion(rows)
	if !bytes.Equal(first, second) {
		t.Error("two renders of the same rows differ byte-wise")
	}
	if len(first) == 0 {
		t.Fatal("renderRegion produced empty output")
	}
	for _, marker := range []string{regionTitle, "docs/CONFIG.md", tableHeader} {
		if !bytes.Contains(first, []byte(marker)) {
			t.Errorf("rendered region misses %q", marker)
		}
	}
	for _, row := range rows {
		if !bytes.Contains(first, []byte(row.line)) {
			t.Errorf("rendered region misses the verbatim row %s", row.line)
		}
	}
}

// TestSkillgen_RegionSentinelsRequired pins the replace contract: a skill
// file missing either sentinel (or carrying them in reverse order) is a
// hard error — the generator never creates a region at an arbitrary spot
// from scratch.
func TestSkillgen_RegionSentinelsRequired(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		content string
	}{
		{
			name:    "no-begin",
			content: "# Skill\n\nТело скилла без региона.\n",
		},
		{
			name:    "no-end",
			content: "# Skill\n\n" + beginMarker + "\n\nСтарый регион без конца.\n",
		},
		{
			name:    "reversed",
			content: "# Skill\n\n" + endMarker + "\n\n" + beginMarker + "\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "SKILL.md")
			if err := os.WriteFile(path, []byte(tc.content), filePerm); err != nil {
				t.Fatalf("write fixture: %v", err)
			}
			updated, err := replaceRegion(path, []byte("новый регион"))
			if err == nil {
				t.Fatalf("replaceRegion accepted a sentinel-less file (%d bytes), want an error", len(updated))
			}
			if updated != nil {
				t.Errorf("replaceRegion returned %d bytes of partial content alongside the error", len(updated))
			}
		})
	}
}

// TestSkillgen_CommittedFileInSync is the CI sync gate (D-8-10,
// T-08-08-01): re-rendering the region from the current docs/CONFIG.md must
// reproduce the committed skills/goswitch-config/SKILL.md byte for byte, so
// pr-sanity enforces the sync through the ordinary `mise run test` with no
// workflow change. RED note: until plan task 2 creates the file, this test
// fails on the missing artifact — the honest red the plan pins.
func TestSkillgen_CommittedFileInSync(t *testing.T) {
	t.Parallel()

	current, err := os.ReadFile("../../skills/goswitch-config/SKILL.md")
	if err != nil {
		t.Fatalf("read committed SKILL.md: %v", err)
	}
	rows, err := parseKeyTable("../../docs/CONFIG.md")
	if err != nil {
		t.Fatalf("parseKeyTable: %v", err)
	}
	updated, err := replaceRegion("../../skills/goswitch-config/SKILL.md", renderRegion(rows))
	if err != nil {
		t.Fatalf("replaceRegion: %v", err)
	}
	if !bytes.Equal(updated, current) {
		t.Error("committed SKILL.md drifted from docs/CONFIG.md — run `mise run skillgen-regen` and commit the region")
	}
}
