---
status: complete
phase: 02-Коррекция слова EN↔RU
source: [02-VERIFICATION.md]
started: 2026-09-15T01:05:00Z
updated: 2026-10-07T21:00:00Z
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

### 3. Live witness — re-pinned mixed-text row (post 261006-vqw)

expected: Live matrix v1 on the current tree: 16/16 PASS — word-mixed settles to «паиghbdtn» (per-character inversion across the flip, owner verdict 2026-10-06).
result: issue
reported: "matrix 15/16 — FAIL word-mixed: entry did not settle to «паиghbdtn» (readback «gfbghbdtn», witness zenity:TEXT:chars=9); daemon log: correction level:1 runes:6 source:привет result:ghbdtn — только русская часть"
severity: major
note: "Живой прогон владельца 2026-10-07 (в этой сессии). Механизм: собственный флип движка
порождает синтетический focus_out/focus_in (SetGlobalEngine), и буфер коррекции жёстко
сбрасывается по ADR-004 — слово, набранное через флип, теряет до-флиповую часть; исправляется
только пост-флиповый токен. Юнит-корпус зелёный, потому что подаёт смешанный токен в буфер
напрямую, без реального флипа (fixture-only слепое пятно). Тот же механизм валит phrase-mixed
в матрицах v3/v4. См. Gaps G-2-3."

## Summary

total: 3
passed: 2
issues: 1
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

- gap_id: G-2-3
  truth: "Mixed word typed across a flip inverts per-character as a whole («gfb»+flip+«привет» → «паиghbdtn», owner verdict 2026-10-06)"
  status: failed
  reason: "Live matrix v1 15/16 (owner run 2026-10-07): word-mixed settles to gfbghbdtn — correction covered only the post-flip token (log: runes:6 source:привет). Own-engine flip produces synthetic focus_out/focus_in; the ADR-004 focus reset hard-clears the correction buffer mid-word."
  severity: major
  test: 3
  artifacts: []
  missing:
    - "ADR-004 revision (D-55): the daemon's own engine flip must NOT reset the correction context — synthetic lifecycle transition ≠ real focus change"
    - "Fix: distinguish flip-internal lifecycle events from real focus changes; preserve buffer across own flip; unit test modeling flip mid-word"
