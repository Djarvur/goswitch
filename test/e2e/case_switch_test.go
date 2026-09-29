package main

// Switch-spike pure-logic corpus (plan 05-01, SWCH-03 criterion 2, D-07):
// the verdict renderer, the probe plan and the XKB-class verdict derivation
// are pure functions, so they are pinned here without a live desktop; the
// live probe sequence itself is exercised by `mise run e2e-switch-spike`
// on the owner's desktop. The indicator verdict is not code-checkable at
// all (SWCH-03, the phase's only non-code criterion) — it belongs to the
// owner's eyes at the Task-2 checkpoint, and the report ends with its
// placeholder line.

import (
	"strings"
	"testing"
)

// sampleProbes builds the four-probe corpus the way the live case does:
// named probes, observed as engine names/counts/classes only (D-20 — never
// the typed text), verdicts from the closed set.
func sampleProbes(t *testing.T) []spikeProbe {
	t.Helper()
	rows := []struct {
		id, name, observed, verdict, detail string
	}{
		{"P1", "shell-activation", "readback=goswitch-en focus_in(goswitch)=+1", verdictFollows, ""},
		{"P2", "external-flip-signal", "flip=goswitch-ru signal=GlobalEngineChanged(goswitch-ru)", verdictFollows, ""},
		{"P3", "xkb-truth", "en_pinned=cyrillic ru_pinned=latin", verdictFollows, ""},
		{"P4", "self-echo", "initiator-signal=none within 6s", verdictNotFollows, ""},
	}
	probes := make([]spikeProbe, 0, len(rows))
	for _, r := range rows {
		p, err := newSpikeProbe(r.id, r.name, r.observed, r.verdict, r.detail)
		if err != nil {
			t.Fatalf("newSpikeProbe(%s): %v", r.id, err)
		}
		probes = append(probes, p)
	}

	return probes
}

func TestRenderSwitchVerdict(t *testing.T) {
	probes := sampleProbes(t)
	table := renderVerdictTable(probes)
	lines := strings.Split(strings.TrimSuffix(table, "\n"), "\n")
	if len(lines) != len(probes) {
		t.Fatalf("renderer produced %d lines, want %d (one per probe):\n%s", len(lines), len(probes), table)
	}
	for i, p := range probes {
		line := lines[i]
		if !strings.HasPrefix(line, p.ID+" "+p.Name+" | ") {
			t.Errorf("line %d = %q, want prefix %q", i, line, p.ID+" "+p.Name+" | ")
		}
		if !strings.Contains(line, "| "+p.Verdict) {
			t.Errorf("line %d = %q misses verdict %q", i, line, p.Verdict)
		}
	}

	// A non-error verdict never carries a renderer-appended reason: the
	// row ends at the verdict token (an observed value may legitimately
	// contain parentheses of its own, e.g. a delivered signal's engine).
	for i, p := range probes {
		if p.Verdict != verdictError && !strings.HasSuffix(lines[i], "| "+p.Verdict) {
			t.Errorf("line %d = %q must end at the verdict token", i, lines[i])
		}
	}

	// The error verdict carries its one-line reason (D-20: types and
	// counts, one line — never field content).
	errProbe, err := newSpikeProbe("P2", "external-flip-signal", "signal=none within 6s", verdictError,
		"no GlobalEngineChanged within 6s")
	if err != nil {
		t.Fatalf("newSpikeProbe error row: %v", err)
	}
	errTable := renderVerdictTable([]spikeProbe{errProbe})
	if !strings.Contains(errTable, "error (no GlobalEngineChanged within 6s)") {
		t.Errorf("error verdict line misses the one-line reason:\n%s", errTable)
	}
	if trimmed := strings.TrimSuffix(errTable, "\n"); strings.ContainsAny(trimmed, "\n\r") {
		t.Errorf("reason must stay on the verdict line, got a multi-line row:\n%s", errTable)
	}

	// The closed verdict set is enforced at construction — the report's
	// verdicts must come from {follows, not-follows, error} only.
	for _, bad := range []string{"switched", "not-switched", "", "FOLLOWS"} {
		if _, err := newSpikeProbe("P1", "shell-activation", "observed", bad, ""); err == nil {
			t.Errorf("verdict %q accepted — outside the closed set {follows, not-follows, error}", bad)
		}
	}

	// D-20: the renderer's input corpus carries no typed text, so any leak
	// could only come from the renderer synthesizing it — pin the absence.
	if strings.Contains(table, wordProbeEN) {
		t.Errorf("table carries the typed probe word (D-20):\n%s", table)
	}
}

func TestSwitchSpikeProbesOrdered(t *testing.T) {
	checkSpikePlanOrder(t)
	probes := sampleProbes(t)
	checkSpikeReportShape(t, probes)
	checkXKBClassVerdicts(t)
}

// checkSpikePlanOrder pins the fixed probe sequence: P1..P4, complete,
// in the research-prescribed order.
func checkSpikePlanOrder(t *testing.T) {
	t.Helper()
	plan := spikeProbePlan()
	want := []spikeProbeSpec{
		{ID: "P1", Name: "shell-activation"},
		{ID: "P2", Name: "external-flip-signal"},
		{ID: "P3", Name: "xkb-truth"},
		{ID: "P4", Name: "self-echo"},
	}
	if len(plan) != len(want) {
		t.Fatalf("probe plan has %d entries, want %d", len(plan), len(want))
	}
	for i, w := range want {
		if plan[i] != w {
			t.Errorf("plan[%d] = %+v, want %+v", i, plan[i], w)
		}
	}
}

// checkSpikeReportShape renders the sample corpus and pins the report
// shape: every probe line exactly once, in order, and the indicator
// verdict line last; the P3 line carries BOTH readback classes.
func checkSpikeReportShape(t *testing.T, probes []spikeProbe) {
	t.Helper()
	report := renderSpikeReport(probes)
	lines := strings.Split(strings.TrimSuffix(report, "\n"), "\n")
	if len(lines) != len(probes)+1 {
		t.Fatalf("report has %d lines, want %d probe lines + the indicator line:\n%s",
			len(lines), len(probes)+1, report)
	}
	pos := 0
	for _, p := range probes {
		found := -1
		for i := pos; i < len(lines)-1; i++ {
			if strings.HasPrefix(lines[i], p.ID+" ") {
				found = i

				break
			}
		}
		if found < 0 {
			t.Errorf("probe %s line missing (or out of order) in:\n%s", p.ID, report)

			continue
		}
		pos = found + 1
	}
	if last := lines[len(lines)-1]; last != spikeIndicatorLine {
		t.Errorf("report must END with the indicator-verdict line %q, got %q", spikeIndicatorLine, last)
	}

	p3Line := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "P3 ") {
			p3Line = line

			break
		}
	}
	if p3Line == "" {
		t.Fatalf("no P3 line in the report:\n%s", report)
	}
	if !strings.Contains(p3Line, "en_pinned=cyrillic") || !strings.Contains(p3Line, "ru_pinned=latin") {
		t.Errorf("P3 line misses the readback classes (en_pinned/ru_pinned):\n%s", p3Line)
	}
}

// checkXKBClassVerdicts pins the P3 derivation: the verdict comes from
// en_pinned ONLY — ru_pinned is the safety-net witness (cyrillic expected
// in both XKB worlds), never a discriminator.
func checkXKBClassVerdicts(t *testing.T) {
	t.Helper()
	verdictCases := []struct{ en, ru, want string }{
		{readbackCyrillic, readbackLatin, verdictFollows},
		{readbackCyrillic, readbackCyrillic, verdictFollows},
		{readbackLatin, readbackCyrillic, verdictNotFollows},
		{readbackLatin, readbackLatin, verdictNotFollows},
	}
	for _, tc := range verdictCases {
		got, err := xkbVerdictFromClasses(tc.en, tc.ru)
		if err != nil {
			t.Errorf("xkbVerdictFromClasses(%s, %s): %v", tc.en, tc.ru, err)

			continue
		}
		if got != tc.want {
			t.Errorf("xkbVerdictFromClasses(%s, %s) = %q, want %q", tc.en, tc.ru, got, tc.want)
		}
	}
	if _, err := xkbVerdictFromClasses("gibberish", readbackCyrillic); err == nil {
		t.Errorf("xkbVerdictFromClasses accepted a non-class readback value")
	}
}
