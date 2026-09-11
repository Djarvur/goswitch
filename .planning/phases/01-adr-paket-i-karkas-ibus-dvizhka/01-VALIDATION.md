---
phase: "01"
slug: "adr-paket-i-karkas-ibus-dvizhka"
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-11"
---

# Phase 01 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Reconstructed from artifacts (State B) by execute-phase's nyquist step hook on 2026-09-11.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (+ go test -race in `mise run ci`) |
| **Config file** | mise.toml (single tool source per D-09/D-10) |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `mise run ci` (build + vet + golangci-lint + test -race + tidy-diff) |
| **Estimated runtime** | ~30 seconds unit; live e2e cases 30–180 s each |

Live e2e (GNOME Wayland session required, run individually):
`mise run e2e-m1` · `mise run e2e-ibus-restart` · `mise run e2e-kill9-survive` · `mise run e2e-d01`

---

## Sampling Rate

- **After every task commit:** `go test ./...`
- **After every plan wave:** `mise run ci`
- **Before `/gsd:verify-work`:** full suite green
- **Max feedback latency:** ~30 s

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 01-01 registration | 01-01 | 1 | INTEG-01 | — | N/A | integration | `mise run e2e-m1` (ListActiveEngines live gate) | ✅ | ✅ green |
| 01-01 observer + logging | 01-01 | 1 | INST-04 | — | key trace gated behind -debug | unit | `go test ./internal/logging/` | ✅ | ✅ green |
| 01-01 panic recover | 01-01 | 1 | INTEG-05 | — | N/A | unit | `go test ./engine/` (TestEngine_RecoverContainsPanic) | ✅ | ✅ green |
| 01-02 layout tables | 01-02 | 2 | CORR-08 | — | N/A | unit | `go test ./layouts/` (golden corpus) | ✅ | ✅ green |
| 01-02 tap FSM | 01-02 | 2 | TEST-01 | — | N/A | unit | `go test ./internal/hotkey/ -race` (deterministic corpus) | ✅ | ✅ green |
| 01-03 session actor | 01-03 | 3 | TEST-01 | — | N/A | unit | `go test ./internal/session/ -race` | ✅ | ✅ green |
| 01-03 e2e stand + injection | 01-03 | 3 | TEST-02, TEST-03, INTEG-03 | — | witness-gate blocks injection into owner windows | e2e | `mise run e2e-m1` | ✅ | ✅ green |
| 01-03 ibus-restart resilience | 01-03 | 3 | INTEG-04 | — | N/A | e2e | `mise run e2e-ibus-restart` | ✅ | ✅ green |
| 01-03 kill9 survival | 01-03 | 3 | INTEG-05 | — | N/A | e2e | `mise run e2e-kill9-survive` | ✅ | ✅ green |
| 01-04 D-01 experiment | 01-04 | 4 | TEST-02, INTEG-04 | — | N/A | e2e | `mise run e2e-d01` (journal: 15 verdicts) | ✅ | ✅ green |
| 01-04 ADR package + M0 | 01-04 | 4 | — | — | N/A | manual_procedural | owner approval «утверждено» 2026-09-10 (recorded in ADR-005) | ✅ | ✅ green |
| 01-05 CI workflows | 01-05 | 5 | TEST-01, INST-04, INTEG-03 | — | workflow `contents: read` only | other | `yaml.v3` parse + structure greps + `mise run ci && mise run tidy-diff` | ✅ | ✅ green |
| 01-05 systemd unit + README | 01-05 | 5 | INST-04 | — | -debug keystroke tracing documented private | other | `systemd-analyze verify dist/systemd/user/goswitchd.service` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements — go test + mise task chain established in plan 01-01; no Wave 0 needed.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Live keyd coexistence (no evdev conflict) | INTEG-02 | keyd.service disabled on this machine; start requires root | test/e2e/README.md § INTEG-02 — 4-step owner checklist |
| First pr-sanity run on real push | TEST-01, INST-04, INTEG-03 | GitHub Actions behavior provable only on a real runner | Push phase branch → PR → observe pr-sanity green |
| First dependabot PR | INTEG-03 | Requires GitHub-side scheduling | Wait for weekly gomod scan; merge a bumped PR through pr-sanity |
| First scheduled govulncheck run | INST-04 | Requires GitHub-side cron | Observe security-scheduled weekly run exit 0 |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30 s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-09-11 (gap gate: manual-only selected for environment-gated checks — root-keyd and GitHub-side CI activations; nyquist auditor dispatch skipped as non-test-writable)

## Validation Audit 2026-09-11 (verify-work)

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

Manual-Only table re-checked during UAT (01-UAT.md): INTEG-02 keyd coexistence — owner-verified **pass**; first pr-sanity on real runner — **proven green** (PR #1, run 34596048376: build/vet/lint/test 39s ✓, govulncheck 29s ✓); first dependabot PR and first scheduled security run — deferred follow-ups (activate after PR #1 merges workflows onto main; govulncheck substance already green). All automated rows remain green; `nyquist_compliant: true` unchanged.
