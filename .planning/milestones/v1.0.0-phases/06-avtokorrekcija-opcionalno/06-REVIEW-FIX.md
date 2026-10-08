---
phase: 06-avtokorrekcija-opcionalno
fixed_at: 2026-10-02T23:05:00Z
review_path: .planning/phases/06-avtokorrekcija-opcionalno/06-REVIEW.md
iteration: 1
findings_in_scope: 3
fixed: 3
skipped: 0
status: all_fixed
---

# Phase 6: Code Review Fix Report

**Fixed at:** 2026-10-02T23:05:00Z
**Source review:** `.planning/phases/06-avtokorrekcija-opcionalno/06-REVIEW.md`
**Iteration:** 1
**Branch:** `gsd/phase-06-avtokorrekcija-opcionalno` (fixes landed via review-fix worktree, fast-forwarded)
**Scope:** Critical + Warning (the 4 Info findings left documented, unfixed — per orchestrator instruction)

**Summary:**
- Findings in scope: 3 (CR-01, CR-02, WR-01)
- Fixed: 3
- Skipped: 0

## Fixed Issues

### CR-01: Async autocorrect confirm executes against moved-on field/buffer state

**Files modified:** `internal/session/actor.go`, `internal/session/actor_test.go`
**Commits:** RED `3e64e95`, GREEN `d5f11e4`
**Fix mechanism chosen:** content re-validation (the review's primary hint; NOT the generation-bump alternative — one mechanism, applied consistently).

- `acPayload` gained `expectToken`/`expectTail` — a snapshot of the buffer as the arming event LEAVES it, taken in `armAutoCorrect` after the boundary's own mutation. The reset branch of `feedKey` now captures the decision BEFORE `HardReset` (the plan-06-06 pin) and launches the confirm AFTER it, so a reset-boundary payload's expectation is the legitimately emptied buffer.
- `autoConfirm`'s fired branch re-validates under the mutex: `!slices.Equal(a.buf.Token(), payload.expectToken) || !slices.Equal(a.buf.Tail(), payload.expectTail)` → `recordACAbstain("payload-stale")` (new closed slug), silence, zero emitter ops. The check sits inside the allowed-role case (after the role-err handling) so the existing `role-timeout` counter semantics of `TestAutoCorrect_RoleDeadlineBounded` stay byte-identical.
- RED evidence: `TestAutoConfirm_RevalidatesArmedPayload` drives the exact review interleaving deterministically (boundary → role call blocked on the double's release channel → cached post-separator push lands → interleaved keystroke `x` mutates the buffer while its own client push is in flight → release). Pre-fix the cached-push fast path fired delete(-7,7)+commit against the stale geometry (`fired = 1`).

### CR-02: Reset-boundary arming races the reset key's own client-side effect

**Files modified:** `internal/session/actor.go`, `internal/session/actor_test.go`
**Commits:** RED `92991c8`, GREEN `2d9ea85`
**Fix mechanism chosen:** always-verify per the review hint — `startRangeCorrection` never takes the cached-push shortcut for `rng.acBoundary` payloads:

```go
if rng.acBoundary || !correct.MatchesSuffix(a.surr, a.pending.match) { ... RequireSurroundingText() ... }
```

The cbfa46c pre-boundary early-answer machinery decides the space case; a post-reset Require answer names the field as the reset key left it → mismatch → fail-closed `verify-mismatch`. Tab/Escape stay in the boundary set (the owner's behavior decision, explicitly out of scope). Manual (Double/Triple/CorrectNow) paths are byte-unchanged (`acBoundary` false → the same condition as before).
- RED evidence: `TestAutoConfirm_ResetBoundaryAlwaysVerifies` (Escape boundary, cache pre-set to the armed word, confirm gated on the role double) — pre-fix the cached fast path fired 2 correction ops with no verify round.

### WR-01: D-53 conjunction evaluated non-atomically (arm-time app, confirm-time role)

**Files modified:** `internal/session/actor.go`, `internal/session/actor_test.go`
**Commits:** RED `802de00`, GREEN `7c8d3cc` (+ `48013f2` lint refactor extracting `acConfirmRefusals`)
**Fix mechanism chosen:** store the resolved app in `acPayload.app`; re-check under the mutex at the confirm's fired threshold — same bridge-namespace equality as arm time, via the existing `acFocusedApp()` (a cached observer read, mutex-safe; `FocusedApp` never blocks). A lost identity source counts `app-unknown`; a different app counts the new closed slug `app-changed` — both fail-closed silence.
- RED evidence: `TestAutoConfirm_RechecksFocusedApp` (focus swapped to the unlisted `org.gnome.Zenity` inside the role-RTT window, role double answers 61) — pre-fix the confirm paired gedit's white-list verdict with Zenity's role and fired.
- Deviation from the review sketch (documented): the review's one-liner counted both `!ok` and `app != payload.app` as `app-unknown`; the fix keeps `app-unknown` for the dead-source case and adds the honest closed slug `app-changed` for the focus-moved case (same fail-closed semantics, diagnosable counters). Coordinator contract ("mismatch → fail-closed silence + skip counter") is met.

## Skipped Issues

None — all in-scope findings fixed. The 4 Info findings (IN-01 counter semantics, IN-02 per-boundary INFO noise, IN-03 godbus pending-call accumulation, IN-04 e2e fixture Wait race) remain documented in 06-REVIEW.md per the orchestrator's scope instruction.

## Verification

- All gates ran in the **isolated review-fix worktree** (`/home/nil/DiskD/W/Djarvur/goswitch/.claude/worktrees/rf-06-450653-1790969692`, branch `gsd-reviewfix/06-450653`, fast-forwarded to `gsd/phase-06-avtokorrekcija-opcionalno` at cleanup) — reproducible from the main checkout at the ff tip `48013f2`.
- Per fix: RED test committed first and demonstrated failing (exact failure output in the fix sections above), then the GREEN fix; the full `internal/session` + `internal/correct` corpus stayed green at every GREEN commit (`go test -race -count=1`).
- After all three: `mise exec -- go test ./internal/session/ ./internal/correct/ -race -count=1` green; full `mise run ci` (build + vet + golangci-lint v2 strict + `go test -race -count=1 ./...`) green. The known flake `TestRun_OnConnHookCalledOnce` did not fire; no re-run needed.
- Lint follow-up `48013f2`: `funlen` (autoConfirm 44 > 40 statements — the two re-checks extracted into `acConfirmRefusals`) and one `lll` corpus line. Behavior-neutral; the three review-fix tests re-run green after it.

---

_Fixed: 2026-10-02T23:05:00Z_
_Fixer: Claude (gsd-code-fixer), go-ultimate strict discipline, strict TDD (RED→GREEN per finding)_
_Iteration: 1_
