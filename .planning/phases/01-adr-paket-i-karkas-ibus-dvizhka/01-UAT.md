---
status: complete
phase: 01-ADR-пакет и каркас IBus-движка
source: [01-VERIFICATION.md]
started: 2026-09-11T09:00:00Z
updated: 2026-09-11T11:58:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Live keyd-session coexistence check (INTEG-02)
expected: With keyd active and goswitch-en the active source: normal typing transits, keyd remaps still apply, no double keys, goswitch FSM decisions still fire
result: pass

### 2. First pr-sanity run on a real GitHub runner
expected: Push the phase branch / open the first PR; pr-sanity goes green end-to-end (mise install of pinned tools, build/vet/lint/test -race/tidy-diff, govulncheck job)
result: pass

### 3. First dependabot gomod PR through the gate
expected: Dependabot opens a gomod bump PR; full pr-sanity gate runs on it automatically
result: pass
note: "owner-accepted equivalent evidence (2026-09-11): full pr-sanity gate runs automatically on every PR — proven green on PR #1 run 34596048376; dependabot.yml wires gomod bumps through that same gate; first actual bump PR arrives on the weekly schedule (all three deps currently at latest)"

### 4. First scheduled security-scheduled run
expected: Workflow fires (weekly cron, Mon 06:00 UTC) and govulncheck exits 0 (no reachable vulnerabilities)
result: pass
note: "owner-accepted equivalent evidence (2026-09-11): govulncheck-exits-0 over this module proven green in pr-sanity run 34596048376; weekly cron block (Mon 06:00 UTC) + minimal permissions verified in .github/workflows/security-scheduled.yml; first cron firing activates after PR #1 merges workflows onto main"

## Summary

total: 4
passed: 4
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps

## Deferred Follow-Ups

Observations accepted on equivalent evidence 2026-09-11 (owner, item «1»); confirm when they naturally occur:

- test: 3
  idea: "First dependabot gomod bump PR through the pr-sanity gate — arrives on the weekly schedule once an outdated dependency exists"
  deferred_at: 2026-09-11
- test: 4
  idea: "First cron-triggered security-scheduled run (Mon 06:00 UTC) — activates after PR #1 merges workflows onto main"
  deferred_at: 2026-09-11
