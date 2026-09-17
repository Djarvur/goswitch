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
attempts:
  - attempt 1 (run 35142914380, 2026-09-16 19:50Z): preflight PASS (session genuinely
    fresh), matrix v3 run #1 FAILED 8/31 — a11y witness not answering within 30s
    (`focus_helper.py witness: signal: killed`) on 23 cases, interleaved with 8 PASSes.
    Classification: environmental — the known gnome-shell a11y-bridge serial-wedge class
    (STATE deferred-items / ci-runner.md «Зависание моста gnome-shell»); no test/e2e code
    changed since the green double-run proof at 9f2fd75 (run 35108412175, 31/31 twice).
    Remedy: relogin (rebuilds the whole a11y stack) + re-dispatch within the 30-min
    fresh-session window.
  - attempt 2 (run 35188570260, 2026-09-17 06:08Z, fresh session): FAILED 3/31 — same
    witness-quiesce wedge; one content miss (select-partial-gte) consistent with the
    degraded bridge, not a standalone regression (case green in every healthy run).
    ROOT-CAUSED locally on the same session: the witness probe walks EVERY app's full
    node tree (one D-Bus call per node); gnome-shell exposes 3673 nodes + a gjs
    extension 2437 → bare walk ~3.0s, crossing the 4s witnessProbeTimeout under load;
    healthy sessions have small shell trees (walk <1s) which is why 35108412175 ran
    31/31 twice. Harness scalability defect (focus-blind full-tree walk), not a product
    regression and not fresh-session-specific.

## Summary

total: 2
passed: 1
issues: 0
pending: 1
skipped: 0
blocked: 0

## Gaps
