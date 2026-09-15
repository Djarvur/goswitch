---
schema_version: 1
open_count: 1
waived_count: 1
fixed_count: 1
total_count: 3
last_updated: 2026-09-15T08:58:03.115Z
---

# Broken Windows Ledger

> Cross-phase defect register. With `workflow.windows_enforce` enabled, `/gsd-ship` blocks while `open_count > 0`.
> Waive with `gsd-tools windows waive <id> "<reason>"` (reason required).
> Mark fixed with `gsd-tools windows fixed <id>`.

| id | phase | kind | file | line | description | status | reason | recorded_at | resolved_at |
|----|-------|------|------|------|-------------|--------|--------|-------------|-------------|
| 1 | 02 | deviation | test/e2e/case_ladder.go |  | Plan oracle level:2 falsified live: google-chrome 153 reports caps 0x29 and applies DeleteSurroundingText — actual level is 1 (ibus#2354 obsolete on target); case pins actual level 1 + verify-after match; no live level-2 witness exists on this desktop (all surfaces report the bit), unit corpus carries level 2 | fixed |  | 2026-09-14T22:27:01.392Z | 2026-09-15T05:19:04.926Z |
| 2 | 3 | deviation | internal/correct/runs.go | 57 | D-24 nothing-to-convert example unsatisfiable under уточнение D-22 (homogeneous converts wholesale per matrix v1); changed=false surface implemented, reserved for selection path 03-03 — owner confirms at verify gate | open |  | 2026-09-15T08:57:29.492Z |  |
| 3 | 3 | deviation | internal/correct/runs.go | 57 | test | waived | junk entry appended by executor output-shape probe — not a real defect, removed intent | 2026-09-15T08:57:41.296Z | 2026-09-15T08:58:03.115Z |

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
    "status": "open",
    "reason": "",
    "recorded_at": "2026-09-15T08:57:29.492Z",
    "resolved_at": null
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
  }
]
````
