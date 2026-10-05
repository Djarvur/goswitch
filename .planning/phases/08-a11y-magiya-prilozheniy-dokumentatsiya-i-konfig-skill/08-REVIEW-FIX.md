---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
fixed_at: 2026-10-05T20:41:15Z
review_path: .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-REVIEW.md
iteration: 1
findings_in_scope: 2
fixed: 2
skipped: 1
status: partial
---

# Phase 8: Code Review Fix Report

**Fixed at:** 2026-10-05T20:41:15Z
**Source review:** `.planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-REVIEW.md`
**Iteration:** 1

**Summary:**
- Findings in scope: 2 (CR-01, WR-01 — WR-02 excluded by owner-decision scope)
- Fixed: 2
- Skipped: 1 (WR-02 — out of fixer scope by instruction)

## Fixed Issues

### CR-01: Reconciler holds its mutex across subprocess/D-Bus I/O — `Apply` blocks the actor's keystroke hot path for the whole in-flight series

**Files modified:** `internal/a11y/a11y.go`, `internal/a11y/a11y_test.go`
**Commits:** `3db244d` (RED — failing corpus), `1927bf9` (GREEN — the fix)
**Status:** fixed — requires human verification (concurrency-semantics change; the regression corpus pins the contract, but the lock-split reasoning deserves a human pass)

**Applied fix:** the reconciler's single mutex was split in two. `seriesMu` now serializes apply series one-at-a-time and is the only lock held across an episode's subprocess/D-Bus I/O; `mu` became a short state lock guarding `desired`/`lastApplied`/`warned` only — never held across I/O or logging. `Apply` touches only the short state lock before spawning the series goroutine, so a fold issued mid-episode returns immediately instead of waiting out the in-flight probe+belt legs (the proven ≈3 s worst-case keystroke stall is gone). The stale-series gate moved into `runSeries(desired)`: the captured desired state is checked against the LIVE desired and last-applied state in one short critical section (superseded or already-applied series return before any subprocess); the deactivation branch (D-8-4, zero subprocesses) shares that same section so the live state cannot drift between check and record; the episode runs with no state lock held; and a post-episode re-check re-reads the live desired state so a config change that landed mid-episode supersedes the stale series' convergence record and is honored by the follow-up series. `episode()` now returns health and the warn-budget reopen (`clear(r.warned)`) folds into the post-episode section; `warn()` takes the state lock around the map check-and-record only, with the `slog.Warn` call outside it.

**Preserved semantics:** idempotence (the diff gate), warn-once per episode per reason with the healthy-activation reopen, argv-literal discipline (untouched — no argv code changed), D-8-4 deactivation-touches-nothing, and the existing corpus (`TestA11y_ApplyDoesNotBlockAndConverges` stale-series/convergence sequence) — all green under `-race -count=2`.

**Regression proof:** the RED tests reproduce the freeze on the pre-fix code shape. `TestA11y_ApplyDoesNotWaitForInFlightSeries` wedges the first series inside its `get` probe (the existing `blockGet` fake seam, observed via a `waitProbeWedged` poll — at that moment the old `run()` provably holds the mutex across I/O), then issues `Apply(false)` and requires it to return within the corpus' 100 ms settle window. On the pre-fix code both new tests failed with «Apply waited for the in-flight series — the fire-and-forget contract is broken (CR-01)» (0.10 s, the blocking duration); the same tests pass on the fix. A second test, `TestA11y_MidEpisodeDesiredChangeHonoredByFollowUp`, pins the follow-up-episode semantics: a mid-episode `Apply(false)` supersedes the stale episode's convergence record and a later re-activation runs a full fresh probe+set+belt episode. The pre-existing corpus was unaffected by the RED commit (only the two new tests failed).

### WR-01: SKILL.md manual frame claims `goswitchctl install` intercepts the GNOME `switch-input-source` binding

**Files modified:** `skills/goswitch-config/SKILL.md`, `docs/CONFIG.md`
**Commit:** `8d68ea4`
**Status:** fixed

**Applied fix:** the manual frame (SKILL.md «Переназначить хоткеи», item 3) now states the actual mechanism: the default `super+space` works because goswitch owns input-source switching (install wraps the sources into goswitch engines) and the daemon follows external engine changes; install does NOT touch the GNOME `switch-input-source` binding (it is snapshotted only for the uninstall rollback) — matching the pinned contract in `internal/install/install.go:100-112` and the install report line. The same false claim was corrected at its sources: the `hotkeys.mode_switch_chord` Meaning cell (`docs/CONFIG.md:55`) and the Caramba-comparison row (`docs/CONFIG.md:292`, "install передаёт биндинг … goswitch"), and the generated region was rebuilt with `mise run skillgen-regen` so the SKILL.md key-table row picks up the corrected source verbatim. `go run ./cmd/skillgen -check` passes (byte-in-sync). No `a11y.apps` rows were touched (WR-02 scope exclusion respected).

## Skipped Issues

### WR-02: `a11y.apps` regex patterns are validated but never matched at runtime

**File:** `internal/config/config.go:183-195, 428-453` (docs: `skills/goswitch-config/SKILL.md:156`, source `docs/CONFIG.md:73`)
**Reason:** owner decision required (per D-8-8 and the fix-scope instruction): either implement the documented per-app matching or restate the list as the on/off gate it actually is. Both options change user-visible behavior/contract — out of fixer scope. `docs/CONFIG.md` a11y rows deliberately untouched.
**Original issue:** the documented per-app regex matching semantics do not exist in the code; the app list's only runtime effect is the non-emptiness of the boolean `Active()` gate.

## Verification

Ran in the ISOLATED WORKTREE `/home/nil/DiskD/W/Djarvur/goswitch/.claude/worktrees/rf-08-2784249-1791232088` (branch `gsd-reviewfix/08-2784249`, cut from `gsd/phase-08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill` at `9a8b9d5`) — the numbers below are reproducible from that tree's commits after the fast-forward, not from a stale checkout:

- `go test -race -count=1 ./internal/a11y/` on the RED commit: exactly the two new tests fail, pre-existing corpus green (bug reproduced).
- `go test -race -count=2 ./internal/a11y/` on the GREEN commit: ok (includes all warn-budget, deactivation, argv-literal, and belt corpus).
- `mise run ci` (build + vet + golangci-lint strict + `go test -race -count=1 ./...`, all 15 packages): green, 0 issues.
- `go run ./cmd/skillgen -check`: byte-in-sync after the regen.
- Working tree clean; no uncommitted changes left behind.

---

_Fixed: 2026-10-05T20:41:15Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
