---
status: complete
phase: 260930-nxd-fix-switchtimeout-40-150ms-actor-go-inst
plan: 01
subsystem: session, install
tags: [switch-deadline, ibus-flip, gnome-keybindings, adr-006, tdd]
requires: []
provides:
  - flip deadline (switchTimeout 150 ms) above the measured live SetGlobalEngine RTT — no spurious deadline-exceeded WARN on healthy flips
  - install that leaves org.gnome.desktop.wm.keybindings switch-input-source / -backward untouched (ADR-006 two-source)
affects:
  - internal/session (flipTo deadline budget)
  - internal/install (install sequence, report verdict)
tech-stack:
  added: []
  patterns:
    - seam-corpus floor assertion (entry budget strictly above the measured live RTT maximum)
    - absence assertion on the recorded subprocess calls (no writes to a schema)
key-files:
  created: []
  modified:
    - internal/session/actor.go
    - internal/session/actor_test.go
    - internal/install/install.go
    - internal/install/install_test.go
decisions:
  - flipBudget test mirror pins the deadline in (liveSwitchRTTMax 45 ms, 150 ms] at seam entry — the corpus now proves healthy flips complete inside the budget
  - install writes NOTHING to the wm.keybindings schema; saveState still snapshots both chords verbatim for the uninstall restore (restore path byte-identical)
  - installCallCount test helper tracks the shorter install sequence (14→12) so the uninstall-suffix window stays correct
metrics:
  duration: 13 min
  completed: 2026-09-30
  commits: 4
  plan_head_before: 09aa10c
actuals:
  tasks: 2
  commits: 4
---

# Quick Task 260930-nxd Summary: switchTimeout 40→150 ms (actor) + install stops clearing GNOME switch bindings (ADR-006)

Two live-evidenced fixes: the flip's SetGlobalEngine deadline raised above the measured ~41-45 ms live round trip so it stops lying `engine switch failed` on every healthy flip, and install stops clearing the GNOME layout-switch chords so the user's own visible switch stays live under the ADR-006 two-source model.

## Tasks Completed

| # | Task | Commits | Files |
|---|------|---------|-------|
| 1 | switchTimeout 40→150 ms — the deadline becomes a real wedge guard above the measured live RTT | 30e34d9 (RED), 8daa027 (GREEN) | internal/session/actor.go, internal/session/actor_test.go |
| 2 | install leaves GNOME switch-input-source bindings untouched (ADR-006 two-source makes clearing obsolete) | 0bbd0fd (RED), 65670b8 (GREEN) | internal/install/install.go, internal/install/install_test.go |

## Task 1 — switchTimeout 40→150 ms

- **RED (30e34d9):** `flipBudget` test mirror moved to 150 ms; new `liveSwitchRTTMax = 45 ms` constant (live evidence: 05-05 proofs + 2026-09-30 journal) with the seam-entry budget pinned in `(45 ms, 150 ms]`; `TestActor_SwitcherDeadlineBounded` failed exactly on the floor assertion (39.99 ms ≤ 45 ms).
- **GREEN (8daa027):** `switchTimeout = 150 * time.Millisecond`; the doc comments at the constant and inside the flipTo block state the new truth — the deadline is the failure-path wedge guard (Pitfall 4 / T-05-03-01), NOT a slice of the <50 ms SPEC reaction budget; healthy flips complete inside it, a wedged bus costs at most the deadline. `flipTo` logic untouched (the `context.WithTimeout` call already consumes the constant).
- Scope guard held: no SPEC reaction-budget text touched elsewhere; `TestActor_FlipRoutesThroughSwitcher` and the whole seam corpus green.

## Task 2 — install leaves the switch chords untouched

- **RED (0bbd0fd):** `TestInstall_Sequence` drops the two expected `set ... []` entries and asserts via the new `assertSwitchBindingsUntouched` helper that NO `gsettings set` touches `gsettingsKeybindingsSchema` during install; `assertSequenceState` flips the report assertion to the preserved verdict `switch-input-source: left untouched (saved for uninstall)`; `TestInstall_SecondInstallKeepsOriginalBackup` now proves a junk live binding value never overwrites the sacred backup nor reaches `gsettings set`. Both failed against the still-clearing code; `TestUninstall_FullRollback` stayed green.
- **GREEN (65670b8):** deleted the `clearSwitchBinding` call from Install (takeover flows straight into activateEngine, with an ADR-006 rationale comment at the site), the `clearSwitchBinding` function, and the orphaned `clearedSwitchBindings` constant; report line updated to the preserved verdict; comment blocks rewritten to the ADR-006 two-source rationale (sources are a pair of goswitch engines, the chord cycles them as a first-class visible switch, the daemon follows external flips via the 05-04 sync listener — live proof external-flip-sync, commit 1caf5b4). `gsettingsKeybindingsSchema`, both key constants, both fallback constants, `restoreSwitchBinding`, and the uninstall flow byte-identical; `saveState` still snapshots both pre-install bindings VERBATIM.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] installCallCount test helper tracked the old install sequence**
- **Found during:** Task 2 GREEN
- **Issue:** `installCallCount = 14` (the fixed prefix `uninstallCalls` strips to isolate the uninstall phase) counted the install sequence WITH the two clearing writes. After the GREEN removal, a happy-path install records 12 calls and the stale count swallowed `stop`/`disable` into the install prefix — `TestUninstall_FullRollback` failed on its unchanged sequence assertion.
- **Fix:** `installCallCount` 14→12 with its counting comment updated. The FullRollback expected sequence and all its assertions stayed byte-identical; the test is green again.
- **Files modified:** internal/install/install_test.go
- **Commit:** 65670b8

**2. [lint gate] lll + gofmt violations in the Task 2 RED test edit**
- **Found during:** Task 2 RED (pre-commit gate)
- **Issue:** one 127-char `t.Errorf` line (max 120) and const-block alignment drift after inserting `junkSwitchBinding`.
- **Fix:** wrapped the message, `gofmt -w`. No behavior change.
- **Files modified:** internal/install/install_test.go
- **Commit:** 0bbd0fd

## Verification Results

- `go test -race -count=1 -run 'TestActor_SwitcherDeadlineBounded|TestActor_FlipRoutesThroughSwitcher' ./internal/session/` — ok
- `go test -race -count=1 -run 'TestInstall_|TestUninstall_' ./internal/install/` — ok
- `mise run ci` (build + vet + golangci-lint strict + test -race, whole repo) — green after every iteration; 0 lint issues
- grep: no `clearedSwitchBindings`/`clearSwitchBinding` anywhere in *.go; no `40 * time.Millisecond` in internal/session/
- RED commits verified failing on the intended assertions only (Task 1: floor assertion, 39.99 ms ≤ 45 ms; Task 2: sequence mismatch + absence assertion), with `TestUninstall_FullRollback` green in both REDs

## Known Stubs

None — no stubs, placeholders, or unwired data introduced.

## Success Criteria Assessment

- Flip deadline: 150 ms sits above the measured 41-45 ms live RTT — healthy flips complete inside the budget, no spurious `switch_engine: context deadline exceeded` WARN; a wedged bus still costs at most the deadline (unit corpus pins the mechanism and the floor). Live-desktop confirmation of the quiet log belongs to the next e2e/live pass (out of scope here).
- Install: writes nothing to the wm.keybindings schema (absence-pinned on the recorded calls); the state file still carries both pre-install bindings verbatim; Super+Space keeps cycling the two goswitch engines with the 05-04 sync listener following external flips (live proof 1caf5b4).
- Uninstall: restores both chords verbatim exactly as before this plan (`TestUninstall_FullRollback` unchanged and green).
- README wording about no-longer-switching GNOME chords dates from the cleared era — deliberately out of scope per the plan's scope guard (needs a separate doc pass).

## Self-Check: PASSED

- internal/session/actor.go: `switchTimeout = 150 * time.Millisecond` present, truthful comments — FOUND
- internal/session/actor_test.go: `flipBudget = 150ms`, `liveSwitchRTTMax = 45ms`, floor assertion — FOUND
- internal/install/install.go: no `clearSwitchBinding`/`clearedSwitchBindings`, preserved report line — FOUND (grep)
- internal/install/install_test.go: absence helper + preserved-verdict assertion — FOUND
- Commits 30e34d9, 8daa027, 0bbd0fd, 65670b8 — FOUND in git log on gsd/phase-05-integratsiya-s-gnome-indikatsiya
