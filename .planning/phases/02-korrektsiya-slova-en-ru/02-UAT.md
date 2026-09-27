---
status: complete
phase: 02-Коррекция слова EN↔RU
source: [02-VERIFICATION.md]
started: 2026-09-15T01:05:00Z
updated: 2026-09-15T04:40:00Z
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
result: pass
note: "initial report «по двойному шифту ничего не происходит» was environmental, not code: no daemon running on the live desktop and active engine was xkb:us::eng — the test precondition (goswitch-en active) was unmet. Retest with live setup (daemon started, ibus engine goswitch-en): owner confirmed in-place replacement, no jump. Daemon log independently confirms: action n:2 → correction done source:ghbdtn result:привет level:1 runes:6 latency_ms:0 → correction verify outcome:match (2026-09-15T08:36). Engine restored to xkb:us::eng, daemon stopped after test."

## Summary

total: 2
passed: 2
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps

- gap_id: G-02-2
  truth: "Double Right Shift corrects the last wrong-layout word in a real GNOME input field (visual, in-place, no jump)"
  status: resolved
  reason: "User reported: по двойному шифту ничего не происходит (nothing happens on double Right Shift)"
  severity: major
  test: 2
  artifacts: []
  missing: []
  root_cause: "environmental — test precondition unmet: goswitchd not running on the live desktop (no systemd unit until Phase 4), goswitch engine not registered, active engine xkb:us::eng. No code defect."
  resolved_by: "live-setup retest (daemon + ibus engine goswitch-en), owner pass + daemon-log evidence"
  resolved_at: 2026-09-15
