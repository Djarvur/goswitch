package main

// The spike-journal row corpus (plan 09-02): the dated verdict rows are the
// phase's evidence, so the append path is pinned twice — every section
// constant must name a section that EXISTS in the phase journal (the R1
// regression: a renamed secTerm04 silently routed term04's rows to a
// "no section" error, and a whole case's evidence never landed), and a
// classified arming round must land its verdict row under the right
// section with the verdict and activation cells intact.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// spikeSections is the section-constant set of the spike case family — the
// corpus iterates all of them so a future rename cannot dodge the pin.
func spikeSections() map[string]string {
	return map[string]string{
		"secSmoke":   secSmoke,
		"secHazards": secHazards,
		"secLadder":  secLadder,
		"secRecipe":  secRecipe,
		"secMatrix":  secMatrix,
		"secCaps":    secCaps,
		"secOracle":  secOracle,
		"secTerm04":  secTerm04,
	}
}

// spikeJournalPath locates the phase journal from the test's working
// directory: the checkout layout is repo-root/docs/adr, and the candidates
// cover the invocation dirs the test binary sees (the package dir and the
// repo root). A wrong-file hit cannot pass: the header assertions below
// require every spike section to be present.
func spikeJournalPath(t *testing.T) string {
	t.Helper()
	for _, rel := range []string{
		"../docs/adr/adr-008-spike-log.md",
		"../../docs/adr/adr-008-spike-log.md",
		"docs/adr/adr-008-spike-log.md",
	} {
		if _, err := os.Stat(rel); err == nil {
			return rel
		}
	}
	t.Fatal("the phase journal adr-008-spike-log.md not found from the test cwd")

	return ""
}

// TestSpikeJournalSectionsExist pins every section constant against the
// REAL phase journal: the header text is data owned by the journal (rows
// must land under it), so a cosmetic code-side rename breaks the append
// path — exactly the term04 silent-evidence-loss finding of 2026-10-13.
func TestSpikeJournalSectionsExist(t *testing.T) {
	data, err := os.ReadFile(spikeJournalPath(t))
	if err != nil {
		t.Fatalf("read the phase journal: %v", err)
	}
	journal := string(data)
	for constant, section := range spikeSections() {
		if !strings.Contains(journal, "## "+section) {
			t.Fatalf("%s = %q has no matching '## ' header in the phase journal"+
				" — rows for this section silently never land; the header text is"+
				" the contract (fix the constant, not the journal)", constant, section)
		}
	}
}

// TestAppendArmRowLandsVerdictRow pins the classified-round row end to end:
// an arming outcome routed through appendArmRowTo must land under the named
// section of the given journal with the probe, the evidence counts, the
// activation field and the verdict cell — and a dead arming must carry the
// harness-fail verdict, a stale one the reproduced-desync verdict.
func TestAppendArmRowLandsVerdictRow(t *testing.T) {
	headers := "## " + secOracle + "\n\n| Проба | Команда | Наблюдаемое | Вердикт |\n" +
		"|---|---|---|---|\n" +
		"## " + secTerm04 + "\n\n| Проба | Команда | Наблюдаемое | Вердикт |\n" +
		"|---|---|---|---|\n"
	path := filepath.Join(t.TempDir(), "adr-008-spike-log.md")
	if err := os.WriteFile(path, []byte(headers), probeFilePerm); err != nil {
		t.Fatalf("seed journal: %v", err)
	}

	stale := armReport{outcome: armStaleContext, engineDelta: 2, tries: 1}
	if err := appendArmRowTo(path, secOracle, "wezterm-flip раунд 1",
		"armGate probe", stale, "/tmp/pty-in"); err != nil {
		t.Fatalf("appendArmRowTo(stale): %v", err)
	}
	dead := armReport{outcome: armDead, tries: 5}
	if err := appendArmRowTo(path, secTerm04, "term04-negative",
		"armGate probe", dead, "/tmp/pty-in"); err != nil {
		t.Fatalf("appendArmRowTo(dead): %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read journal back: %v", err)
	}
	journal := string(data)
	for _, want := range []string{
		"reproduced: engine→stale-context desync (XIM)",
		"harness-fail: arming probe reached neither engine nor pty",
		activationMark,
		"engine_traced=true (probe-key records delta 2)",
		"engine_traced=false",
	} {
		if !strings.Contains(journal, want) {
			t.Fatalf("appended journal lacks %q:\n%s", want, journal)
		}
	}
	// Section routing: the stale row under secOracle, the dead row under
	// secTerm04 — a row under the wrong section is evidence misfiled.
	term04Head := strings.Index(journal, "## "+secTerm04)
	if term04Head < 0 {
		t.Fatalf("the journal lost its %q section:\n%s", secTerm04, journal)
	}
	oraclePart := journal[:term04Head]
	if !strings.Contains(oraclePart, "reproduced: engine→stale-context desync (XIM)") {
		t.Fatalf("the desync row did not land under %q:\n%s", secOracle, oraclePart)
	}
	term04Part := journal[term04Head:]
	if !strings.Contains(term04Part, "harness-fail: arming probe reached neither engine nor pty") {
		t.Fatalf("the harness-fail row did not land under %q:\n%s", secTerm04, term04Part)
	}
}
