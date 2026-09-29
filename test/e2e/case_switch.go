package main

import "errors"

// Switch-spike contracts (plan 05-01, SWCH-03 criterion 2): the RED stubs
// of the verdict renderer, the probe plan and the XKB-class derivation —
// the GREEN step implements them together with the live case.

// spikeReportPath is the spike's verdict report at the repo root (the
// perf-report.txt/e2e-report.txt convention).
const spikeReportPath = "switch-spike-report.txt"

// spikeIndicatorLine is the report's final line: the owner's indicator
// verdict (follows / not-follows — eyes only, SWCH-03), filled at the
// Task-2 checkpoint; the case always writes the unknown placeholder.
const spikeIndicatorLine = "indicator-verdict: unknown"

// The closed verdict set of every spike line.
const (
	verdictFollows    = "follows"
	verdictNotFollows = "not-follows"
	verdictError      = "error"
)

// Readback classes of the P3 typing rounds (D-20: the typed word itself
// never reaches the report — its script class does).
const (
	readbackCyrillic = "cyrillic"
	readbackLatin    = "latin"
)

// spikeProbe is one probe's rendered verdict row.
type spikeProbe struct {
	ID       string
	Name     string
	Observed string
	Verdict  string
	Detail   string
}

// spikeProbeSpec pins the probe plan: fixed, complete, ordered.
type spikeProbeSpec struct {
	ID   string
	Name string
}

func newSpikeProbe(id, name, observed, verdict, detail string) (spikeProbe, error) {
	return spikeProbe{}, errors.New("switch-spike: not implemented")
}

func renderVerdictTable(probes []spikeProbe) string { return "" }

func renderSpikeReport(probes []spikeProbe) string { return "" }

func spikeProbePlan() []spikeProbeSpec { return nil }

func xkbVerdictFromClasses(enClass, ruClass string) (string, error) {
	return "", errors.New("switch-spike: not implemented")
}
