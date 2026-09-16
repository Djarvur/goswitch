---
status: testing
phase: 04-Поставка и приёмка
source: [04-VERIFICATION.md]
started: 2026-09-16T18:40:00Z
updated: 2026-09-16T19:45:00Z
---

## Current Test

number: 2
name: D-48 formal fresh-session run — owner relogin, then dispatch e2e-matrix fresh_session=true
expected: |
  loginctl preflight passes (session age <= 30 min), then both sequential v3 runs report
  matrix: 31/31 PASS, job conclusion success
awaiting: user response

## Tests

### 1. D-49 owner manual acceptance
expected: execute docs/ACCEPTANCE.md (6 items + Summary) in one live session — install from
README on the live desktop (release-archive channel, no root) → `goswitchctl selfcheck` all
green → live gesture set (flip/word/phrase/selection/combo/mixed/reload) → status counters
grow → uninstall → desktop fully restored
result: pass (owner, 2026-09-16)

### 2. D-48 formal fresh-session run
expected: owner relogin → dispatch e2e-matrix with fresh_session=true per
docs/ci-runner.md «D-48 double-run gate» → loginctl preflight passes (session age ≤ 30 min),
then both sequential v3 runs report matrix: 31/31 PASS, job conclusion success
result: [pending]

## Summary

total: 2
passed: 1
issues: 0
pending: 1
skipped: 0
blocked: 0

## Gaps
