---
status: complete
phase: 02-Коррекция слова EN↔RU
source: [02-VERIFICATION.md]
started: 2026-09-15T01:05:00Z
updated: 2026-09-15T04:30:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Owner decision — ADR-003 wording vs live ladder finding

expected: Owner annotates ADR-003/ROADMAP SC-3 wording («fallback-уровень» → «фактический уровень») or accepts the recorded deviation; WINDOWS ledger #1 → resolved. Nothing in code needs to change — documentation-acceptance decision on owner-gated documents (ADR-003 left deliberately unedited, Accepted).
result: pass
note: "owner accepted the recorded deviation as-is (Option B, 2026-09-15): ADR-003 wording left unedited; WINDOWS ledger #1 resolved"

### 2. Visual no-jump confirmation (goal clause «без визуального прыжка»)

expected: With goswitch-en active, type `ghbdtn` in a real GNOME input field (e.g. gedit or a browser box), double-tap Right Shift — the word becomes `привет` in place: one atomic replacement over exactly the wrong-word range, no visible cursor jump, flicker, or transient doubled text. (Content-exactness is e2e-proven; the perceptual property is inherently visual.)
result: issue
reported: "по двойному шифту ничего не происходит"
severity: major

## Summary

total: 2
passed: 1
issues: 1
pending: 0
skipped: 0
blocked: 0

## Gaps

- gap_id: G-02-2
  truth: "Double Right Shift corrects the last wrong-layout word in a real GNOME input field (visual, in-place, no jump)"
  status: failed
  reason: "User reported: по двойному шифту ничего не происходит (nothing happens on double Right Shift)"
  severity: major
  test: 2
  artifacts: []
  missing: []
