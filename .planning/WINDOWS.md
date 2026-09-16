---
schema_version: 1
open_count: 1
waived_count: 2
fixed_count: 3
total_count: 6
last_updated: 2026-09-16T14:53:16.148Z
---

# Broken Windows Ledger

> Cross-phase defect register. With `workflow.windows_enforce` enabled, `/gsd-ship` blocks while `open_count > 0`.
> Waive with `gsd-tools windows waive <id> "<reason>"` (reason required).
> Mark fixed with `gsd-tools windows fixed <id>`.

| id | phase | kind | file | line | description | status | reason | recorded_at | resolved_at |
|----|-------|------|------|------|-------------|--------|--------|-------------|-------------|
| 1 | 02 | deviation | test/e2e/case_ladder.go |  | Plan oracle level:2 falsified live: google-chrome 153 reports caps 0x29 and applies DeleteSurroundingText — actual level is 1 (ibus#2354 obsolete on target); case pins actual level 1 + verify-after match; no live level-2 witness exists on this desktop (all surfaces report the bit), unit corpus carries level 2 | fixed |  | 2026-09-14T22:27:01.392Z | 2026-09-15T05:19:04.926Z |
| 2 | 3 | deviation | internal/correct/runs.go | 57 | D-24 nothing-to-convert example unsatisfiable under уточнение D-22 (homogeneous converts wholesale per matrix v1); changed=false surface implemented, reserved for selection path 03-03 — owner confirms at verify gate | fixed |  | 2026-09-15T08:57:29.492Z | 2026-09-16T14:53:15.673Z |
| 3 | 3 | deviation | internal/correct/runs.go | 57 | test | waived | junk entry appended by executor output-shape probe — not a real defect, removed intent | 2026-09-15T08:57:41.296Z | 2026-09-15T08:58:03.115Z |
| 4 | 3 | unmet-truth | test/e2e/case_macr.go | 595 | GTK4-Wayland (GNOME 46) does not apply IBus-forwarded key events: the relay reaches the focused InputContext (dbus-monitor pinned) but the widget never acts — MACR's Ctrl+letter remap, the ADR-003 level-2 Backspace replay and the D-28 Ctrl+V burst all stop at the client boundary (03-05 Deviation 3) | waived | owner accepted GTK4-Wayland forwarded-events platform limitation (UAT 2026-09-15, REQUIREMENTS MACR-01 caveat) | 2026-09-15T13:23:14.477Z | 2026-09-16T14:53:16.148Z |
| 5 | 04 | unmet-truth | test/e2e/perf.go | 1 | INST-03 latency budget verdict is FAIL under the prescribed ydotool→AT-SPI window (p95 190-205ms >= 50ms over two full runs): the window is dominated by ydotool 0.1.8 per-event injection overhead (~80-125ms/event), daemon reaction is 0.2-1.8ms; owner methodology decision required before the D-46 README table publishes latency numbers | open |  | 2026-09-16T13:51:53.747Z |  |
| 6 | 04 | deviation | test/e2e/main.go | 201 | 04-04 (1b443eb) turned the matrix path's zero watchdog limit into a literal time.After(0): every matrix case (v1/v2/v3) failed instantly with 'did not complete within 0s' — live-caught by the first D-48 double-run dispatch (run 35107406144); fixed structurally by watchdogLimit resolving 0 to caseTimeout inside runCaseWatchdog (9f2fd75, RED->GREEN tests) | fixed |  | 2026-09-16T14:35:30.840Z | 2026-09-16T14:35:35.768Z |

````json
[
  {
    "id": 1,
    "kind": "deviation",
    "phase": "02",
    "file": "test/e2e/case_ladder.go",
    "line": null,
    "description": "Plan oracle level:2 falsified live: google-chrome 153 reports caps 0x29 and applies DeleteSurroundingText — actual level is 1 (ibus#2354 obsolete on target); case pins actual level 1 + verify-after match; no live level-2 witness exists on this desktop (all surfaces report the bit), unit corpus carries level 2",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-09-14T22:27:01.392Z",
    "resolved_at": "2026-09-15T05:19:04.926Z"
  },
  {
    "id": 2,
    "kind": "deviation",
    "phase": "3",
    "file": "internal/correct/runs.go",
    "line": 57,
    "description": "D-24 nothing-to-convert example unsatisfiable under уточнение D-22 (homogeneous converts wholesale per matrix v1); changed=false surface implemented, reserved for selection path 03-03 — owner confirms at verify gate",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-09-15T08:57:29.492Z",
    "resolved_at": "2026-09-16T14:53:15.673Z"
  },
  {
    "id": 3,
    "kind": "deviation",
    "phase": "3",
    "file": "internal/correct/runs.go",
    "line": 57,
    "description": "test",
    "status": "waived",
    "reason": "junk entry appended by executor output-shape probe — not a real defect, removed intent",
    "recorded_at": "2026-09-15T08:57:41.296Z",
    "resolved_at": "2026-09-15T08:58:03.115Z"
  },
  {
    "id": 4,
    "kind": "unmet-truth",
    "phase": "3",
    "file": "test/e2e/case_macr.go",
    "line": 595,
    "description": "GTK4-Wayland (GNOME 46) does not apply IBus-forwarded key events: the relay reaches the focused InputContext (dbus-monitor pinned) but the widget never acts — MACR's Ctrl+letter remap, the ADR-003 level-2 Backspace replay and the D-28 Ctrl+V burst all stop at the client boundary (03-05 Deviation 3)",
    "status": "waived",
    "reason": "owner accepted GTK4-Wayland forwarded-events platform limitation (UAT 2026-09-15, REQUIREMENTS MACR-01 caveat)",
    "recorded_at": "2026-09-15T13:23:14.477Z",
    "resolved_at": "2026-09-16T14:53:16.148Z"
  },
  {
    "id": 5,
    "kind": "unmet-truth",
    "phase": "04",
    "file": "test/e2e/perf.go",
    "line": 1,
    "description": "INST-03 latency budget verdict is FAIL under the prescribed ydotool→AT-SPI window (p95 190-205ms >= 50ms over two full runs): the window is dominated by ydotool 0.1.8 per-event injection overhead (~80-125ms/event), daemon reaction is 0.2-1.8ms; owner methodology decision required before the D-46 README table publishes latency numbers",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-09-16T13:51:53.747Z",
    "resolved_at": null
  },
  {
    "id": 6,
    "kind": "deviation",
    "phase": "04",
    "file": "test/e2e/main.go",
    "line": 201,
    "description": "04-04 (1b443eb) turned the matrix path's zero watchdog limit into a literal time.After(0): every matrix case (v1/v2/v3) failed instantly with 'did not complete within 0s' — live-caught by the first D-48 double-run dispatch (run 35107406144); fixed structurally by watchdogLimit resolving 0 to caseTimeout inside runCaseWatchdog (9f2fd75, RED->GREEN tests)",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-09-16T14:35:30.840Z",
    "resolved_at": "2026-09-16T14:35:35.768Z"
  }
]
````
