---
status: testing
phase: 01-ADR-пакет и каркас IBus-движка
source: [01-VERIFICATION.md]
started: 2026-09-11T09:00:00Z
updated: 2026-09-11T09:00:00Z
---

## Current Test

number: 1
name: Live keyd-session coexistence check (INTEG-02 owner checklist in test/e2e/README.md § INTEG-02)
expected: |
  With keyd active and goswitch-en the active source: normal typing transits,
  keyd remaps still apply, no double keys, goswitch FSM decisions still fire
awaiting: user response

## Tests

### 1. Live keyd-session coexistence check (INTEG-02)
expected: With keyd active and goswitch-en the active source: normal typing transits, keyd remaps still apply, no double keys, goswitch FSM decisions still fire
result: [pending]

### 2. First pr-sanity run on a real GitHub runner
expected: Push the phase branch / open the first PR; pr-sanity goes green end-to-end (mise install of pinned tools, build/vet/lint/test -race/tidy-diff, govulncheck job)
result: [pending]

### 3. First dependabot gomod PR through the gate
expected: Dependabot opens a gomod bump PR; full pr-sanity gate runs on it automatically
result: [pending]

### 4. First scheduled security-scheduled run
expected: Workflow fires (weekly cron, Mon 06:00 UTC) and govulncheck exits 0 (no reachable vulnerabilities)
result: [pending]

## Summary

total: 4
passed: 0
issues: 0
pending: 4
skipped: 0
blocked: 0

## Gaps
