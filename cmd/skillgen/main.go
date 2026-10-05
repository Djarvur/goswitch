// Command skillgen rewrites the generated config-reference region of
// skills/goswitch-config/SKILL.md from the key table of docs/CONFIG.md —
// the single source of the config documentation (D-8-10). Table rows are
// copied verbatim, so the two documents cannot disagree by construction,
// and the committed SKILL.md is the golden: the -check mode re-renders the
// region and compares it byte for byte, exiting non-zero on drift.
//
// RED stub: the pipeline bodies and the dispatch land in the GREEN commit;
// the shape below is what the corpus compiles against (the 06-04/08-02
// RED-stub precedent).
package main

// Paths are relative to the repository root — the tool always runs from
// there (the mise task, the -check invocation and the test corpus all
// assume it).
const (
	configPath = "docs/CONFIG.md"
	skillPath  = "skills/goswitch-config/SKILL.md"

	// The generated region sits between these two comment sentinels; the
	// manual frame around them is never touched by the generator.
	beginMarker = "<!-- goswitch-config:generated BEGIN -->"
	endMarker   = "<!-- goswitch-config:generated END -->"

	// regionTitle heads the generated region.
	regionTitle = "## Ключи config.yaml"

	// The canonical table header; parse hard-errors when the source table
	// departs from it, so the emitted header is the source header.
	tableHeader = "| Key | Type | Default | Range / vocabulary | Meaning |"

	// The canonical separator line that must follow the header.
	tableSeparator = "|-----|-----|---------|--------------------|---------|"

	// keyTableCells is the pinned column count of the key table.
	keyTableCells = 5

	filePerm = 0o644
)

// tableRow is one parsed key-table line: the raw verbatim markdown line and
// its five cells (markdown-escaped pipes stay inside a cell).
type tableRow struct {
	line  string
	cells []string
}

// parseKeyTable extracts and validates the key-table rows of the CONFIG.md
// markdown.
func parseKeyTable(path string) ([]tableRow, error) {
	return nil, nil // RED stub
}

// renderRegion emits the generated markdown region for the rows.
func renderRegion(rows []tableRow) []byte {
	return nil // RED stub
}

// replaceRegion splices the region between the sentinels of the skill file.
func replaceRegion(path string, region []byte) ([]byte, error) {
	return nil, nil // RED stub
}

func main() {} // RED stub: the dispatch lands in GREEN.
