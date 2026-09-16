---
status: testing
phase: 04-Поставка и приёмка
source: [04-VERIFICATION.md]
started: 2026-09-16T18:40:00Z
updated: 2026-09-16T18:40:00Z
---

## Current Test

number: 1
name: D-49 owner manual acceptance — execute docs/ACCEPTANCE.md in one live session
expected: |
  Every checklist item lands its expected result; Summary block filled; any miss is a
  release-blocking finding (fix + re-tag v1.0.1 per the documented release cycle)
awaiting: user response

## Tests

### 1. D-49 owner manual acceptance
expected: execute docs/ACCEPTANCE.md (6 items + Summary) in one live session — install from
README on the live desktop (release-archive channel, no root) → `goswitchctl selfcheck` all
green → live gesture set (flip/word/phrase/selection/combo/mixed/reload) → status counters
grow → uninstall → desktop fully restored
result: [pending]

### 2. D-48 formal fresh-session run
expected: owner relogin → dispatch e2e-matrix with fresh_session=true per
docs/ci-runner.md «D-48 double-run gate» → loginctl preflight passes (session age ≤ 30 min),
then both sequential v3 runs report matrix: 31/31 PASS, job conclusion success
result: [pending]

## Summary

total: 2
passed: 0
issues: 0
pending: 2
skipped: 0
blocked: 0

## Gaps
