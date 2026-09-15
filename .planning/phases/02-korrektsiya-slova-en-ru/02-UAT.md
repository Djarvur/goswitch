---
status: testing
phase: 02-Коррекция слова EN↔RU
source: [02-VERIFICATION.md]
started: 2026-09-15T01:05:00Z
updated: 2026-09-15T01:05:00Z
---

## Current Test

number: 1
name: Owner decision on the live ladder finding (ADR-003 wording vs reality — WINDOWS ledger #1)
expected: |
  Owner reviews the finding (google-chrome 153 reports CapSurroundingText and APPLIES
  DeleteSurroundingText — actual ladder level is 1 on every matrix surface; the Chromium
  level-2 fallback assumption from ibus#2354 lore is falsified on this desktop; the level-2
  contract is carried by TestActor_Level2NoCaps / TestActor_Level2WithTailAndRunes unit corpus)
  and either annotates ADR-003 / ROADMAP SC-3 wording («fallback-уровень» → «фактический
  уровень») or accepts the deviation as recorded. WINDOWS ledger entry #1 moves from open
  to resolved.
awaiting: user response

## Tests

### 1. Owner decision — ADR-003 wording vs live ladder finding

expected: Owner annotates ADR-003/ROADMAP SC-3 wording («fallback-уровень» → «фактический уровень») or accepts the recorded deviation; WINDOWS ledger #1 → resolved. Nothing in code needs to change — documentation-acceptance decision on owner-gated documents (ADR-003 left deliberately unedited, Accepted).
result: [pending]

### 2. Visual no-jump confirmation (goal clause «без визуального прыжка»)

expected: With goswitch-en active, type `ghbdtn` in a real GNOME input field (e.g. gedit or a browser box), double-tap Right Shift — the word becomes `привет` in place: one atomic replacement over exactly the wrong-word range, no visible cursor jump, flicker, or transient doubled text. (Content-exactness is e2e-proven; the perceptual property is inherently visual.)
result: [pending]

## Summary

total: 2
passed: 0
issues: 0
pending: 2
skipped: 0
blocked: 0

## Gaps
