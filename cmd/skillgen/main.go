// Command skillgen rewrites the generated config-reference region of
// skills/goswitch-config/SKILL.md from the key table of docs/CONFIG.md —
// the single source of the config documentation (D-8-10). Table rows are
// copied verbatim, so the two documents cannot disagree by construction,
// and the committed SKILL.md is the golden.
//
// It is a dev-side tool (Directive 3, the dictgen discipline): the mise
// task skillgen-regen (`go run ./cmd/skillgen`) rewrites the region from
// the repository root, and CI never regenerates — it only verifies, via
// `go run ./cmd/skillgen -check` and the CommittedFileInSync golden test
// inside the ordinary `mise run test`. Parsing follows the golden
// generator discipline: hard errors on data anomalies (a wrong cell count,
// an empty cell, a key without a section dot, missing or misordered region
// sentinels), never silent skips, and regeneration is byte-for-byte
// deterministic.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

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

	// regionRowsCap seeds the row slice above the 21 documented keys.
	regionRowsCap = 32

	// regionFrameNl counts the newlines spliced around the region: one
	// after the begin sentinel, one before the end sentinel.
	regionFrameNl = 2

	filePerm = 0o644
)

// tableRow is one parsed key-table line: the raw verbatim markdown line and
// its five cells (markdown-escaped pipes stay inside a cell).
type tableRow struct {
	line  string
	cells []string
}

// Static error values; call sites only ever wrap them with %w.
var (
	errHeaderMissing  = errors.New("key table header not found")
	errSepMissing     = errors.New("table header is not followed by a separator line")
	errEmptyTable     = errors.New("key table parsed to zero rows")
	errCellsWrong     = errors.New("table row does not have exactly five cells")
	errCellEmpty      = errors.New("table row has an empty cell")
	errKeyUnqualified = errors.New("table row key is not section-qualified (no dot)")
	errBeginMissing   = errors.New("generated region begin sentinel not found")
	errEndMissing     = errors.New("generated region end sentinel not found")
	errSentinelDup    = errors.New("generated region sentinel appears more than once")
	errSentinelOrder  = errors.New("generated region end sentinel precedes the begin sentinel")
	errRegionDrift    = errors.New(
		"generated region drifted from docs/CONFIG.md — run `mise run skillgen-regen` and commit",
	)
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "skillgen: %v\n", err)
		os.Exit(1)
	}
}

// run dispatches the whole pipeline: parse the CONFIG.md key table, render
// the region, and splice it into the SKILL.md content. Without -check the
// updated bytes are written back; with -check they are compared byte for
// byte against the committed file and any drift exits non-zero — the CI
// gate shape (regenerate-and-compare, the tidy-diff form).
func run() error {
	check := flag.Bool("check", false,
		"сверить закоммиченный SKILL.md с регенерацией из docs/CONFIG.md и выйти с кодом 1 при расхождении")
	flag.Parse()

	rows, err := parseKeyTable(configPath)
	if err != nil {
		return err
	}
	updated, err := replaceRegion(skillPath, renderRegion(rows))
	if err != nil {
		return err
	}
	if !*check {
		if err := os.WriteFile(skillPath, updated, filePerm); err != nil {
			return fmt.Errorf("write %s: %w", skillPath, err)
		}

		return nil
	}
	current, err := os.ReadFile(skillPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", skillPath, err)
	}
	if !bytes.Equal(updated, current) {
		return fmt.Errorf("%s: %w", skillPath, errRegionDrift)
	}

	return nil
}

// parseKeyTable extracts the key-table rows from the CONFIG.md markdown.
// The canonical header line anchors the table, a separator-shaped line must
// follow, and every subsequent line beginning with "|" is a data row until
// the first line that is not. Each row must carry exactly five non-empty
// cells (markdown-escaped pipes stay inside a cell) with a
// section-qualified key first; any violation is a hard error naming the
// offending line — the dictgen discipline, never a silent skip.
func parseKeyTable(path string) ([]tableRow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	lines := strings.Split(string(data), "\n")
	headerIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == tableHeader {
			headerIdx = i

			break
		}
	}
	if headerIdx < 0 {
		return nil, fmt.Errorf("%s: %w", path, errHeaderMissing)
	}
	sepIdx := headerIdx + 1
	if sepIdx >= len(lines) || !isSeparatorLine(lines[sepIdx]) {
		return nil, fmt.Errorf("%s:%d: %w", path, sepIdx+1, errSepMissing)
	}
	rows := make([]tableRow, 0, regionRowsCap)
	for i := sepIdx + 1; i < len(lines); i++ {
		line := lines[i]
		if !strings.HasPrefix(line, "|") {
			break // the table ends at the first non-row line
		}
		row, err := parseRow(line)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, i+1, err)
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%s: %w", path, errEmptyTable)
	}

	return rows, nil
}

// parseRow validates one table line: exactly five non-empty cells with a
// section-qualified key in the first.
func parseRow(line string) (tableRow, error) {
	cells := splitCells(line)
	if len(cells) != keyTableCells {
		return tableRow{}, fmt.Errorf("%w: got %d", errCellsWrong, len(cells))
	}
	for i, cell := range cells {
		if cell == "" {
			return tableRow{}, fmt.Errorf("%w: cell %d", errCellEmpty, i+1)
		}
	}
	if !strings.Contains(cells[0], ".") {
		return tableRow{}, fmt.Errorf("%w: %s", errKeyUnqualified, cells[0])
	}

	return tableRow{line: line, cells: cells}, nil
}

// splitCells splits one markdown table line into its cells: boundaries are
// unescaped pipes, a `\|` escape stays inside the cell verbatim, and
// surrounding whitespace is trimmed from every cell.
func splitCells(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	cells := make([]string, 0, keyTableCells)
	var cell strings.Builder
	for i := 0; i < len(trimmed); i++ {
		switch c := trimmed[i]; {
		case c == '\\' && i+1 < len(trimmed) && trimmed[i+1] == '|':
			cell.WriteString(`\|`)
			i++ // the escape consumes the pipe too; the loop adds one
		case c == '|':
			cells = append(cells, strings.TrimSpace(cell.String()))
			cell.Reset()
		default:
			cell.WriteByte(c)
		}
	}
	cells = append(cells, strings.TrimSpace(cell.String()))

	return cells
}

// isSeparatorLine reports whether the line is a markdown table separator:
// dashes, colons, pipes and blanks only, with at least one dash.
func isSeparatorLine(line string) bool {
	if !strings.Contains(line, "-") {
		return false
	}
	for _, r := range line {
		switch r {
		case '-', ':', '|', ' ', '\t':
		default:
			return false
		}
	}

	return true
}

// renderRegion emits the generated markdown region: the title, the source
// signature naming the single source, and the canonical table carrying
// every row line verbatim. No clocks, no maps, no environment reads — two
// calls on the same rows are byte-equal by construction.
func renderRegion(rows []tableRow) []byte {
	var out bytes.Buffer
	out.WriteString(regionTitle)
	out.WriteString("\n\n")
	out.WriteString("Источник: `docs/CONFIG.md` — единственный источник этого раздела (D-8-10): " +
		"правки ключей вносятся только туда; регион пересобирается `mise run skillgen-regen`, " +
		"сверка — `go run ./cmd/skillgen -check`. Регион генерированный — не редактируйте его вручную.")
	out.WriteString("\n\n")
	out.WriteString(tableHeader)
	out.WriteString("\n")
	out.WriteString(tableSeparator)
	out.WriteString("\n")
	for _, row := range rows {
		out.WriteString(row.line)
		out.WriteString("\n")
	}

	return out.Bytes()
}

// replaceRegion splices the region into the skill file between the two
// sentinels and returns the updated bytes in memory — the caller decides
// whether to write them. The manual frame and the sentinels themselves are
// preserved; a missing, duplicated or misordered sentinel is a hard error,
// because the generator never invents a region at an arbitrary spot.
func replaceRegion(path string, region []byte) ([]byte, error) {
	current, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	beginIdx := bytes.Index(current, []byte(beginMarker))
	if beginIdx < 0 {
		return nil, fmt.Errorf("%s: %w", path, errBeginMissing)
	}
	if bytes.Contains(current[beginIdx+len(beginMarker):], []byte(beginMarker)) {
		return nil, fmt.Errorf("%s: %w", path, errSentinelDup)
	}
	endIdx := bytes.Index(current, []byte(endMarker))
	if endIdx < 0 {
		return nil, fmt.Errorf("%s: %w", path, errEndMissing)
	}
	if bytes.Contains(current[endIdx+len(endMarker):], []byte(endMarker)) {
		return nil, fmt.Errorf("%s: %w", path, errSentinelDup)
	}
	if endIdx < beginIdx {
		return nil, fmt.Errorf("%s: %w", path, errSentinelOrder)
	}
	head := current[:beginIdx+len(beginMarker)]
	tail := current[endIdx:]
	updated := make([]byte, 0, len(head)+len(region)+len(tail)+regionFrameNl)
	updated = append(updated, head...)
	updated = append(updated, '\n')
	updated = append(updated, region...)
	updated = append(updated, '\n')
	updated = append(updated, tail...)

	return updated, nil
}
