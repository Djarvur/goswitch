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
result: skipped
reason: "Deferred follow-up: first dependabot PR is a weekly-scheduled future event; fetch-updates API unavailable (404); mechanism verified via dependabot.yml on main + pr-sanity gate proven green in test 2"

### 4. First scheduled security-scheduled run
expected: Workflow fires (weekly cron, Mon 06:00 UTC) and govulncheck exits 0 (no reachable vulnerabilities)
result: skipped
reason: "Deferred follow-up: workflow not yet on default branch (registers after PR #1 merges — dispatch API 404s until then); govulncheck-exits-0 substance already proven by green govulncheck job in pr-sanity run 34596048376; cron block verified in .github/workflows/security-scheduled.yml"

## Summary

total: 4
passed: 2
issues: 0
pending: 0
skipped: 2
blocked: 0

## Gaps

## Deferred Follow-Ups

- test: 3
  idea: "First dependabot gomod bump PR through the pr-sanity gate — arrives on the weekly schedule once an outdated dependency exists (all three deps currently at latest)"
  deferred_at: 2026-09-11
- test: 4
  idea: "First cron-triggered security-scheduled run (Mon 06:00 UTC) — activates after PR #1 merges workflows onto main; govulncheck substance already green in pr-sanity"
  deferred_at: 2026-09-11
