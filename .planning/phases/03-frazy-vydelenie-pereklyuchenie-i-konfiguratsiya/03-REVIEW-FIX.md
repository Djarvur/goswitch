---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
fixed_at: 2026-09-15T15:44:08Z
review_path: .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-REVIEW.md
iteration: 1
findings_in_scope: 8
fixed: 8
skipped: 0
status: all_fixed
---

# Phase 3: Code Review Fix Report

**Fixed at:** 2026-09-15T15:44:08Z
**Source review:** .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-REVIEW.md
**Iteration:** 1
**Scope:** Critical + Warning (CR-01..03, WR-01..05); Info findings out of scope

**Summary:**
- Findings in scope: 8
- Fixed: 8
- Skipped: 0

**Verification:** every fix was developed test-first (RED → GREEN) and the
full gate `mise run ci` (build + vet + golangci-lint strict + `go test
-race`) ran green after EACH fix and once more after the last one. The gates
ran inside the isolated review-fix worktree
(`.claude/worktrees/rf-03-1022466-1789485307`, branch
`gsd-reviewfix/03-1022466`, since fast-forwarded into
`gsd/phase-03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya` and removed)
— the numbers are reproducible from the branch history at the listed
commits, not from the torn-down worktree directory.

## Fixed Issues

### CR-01: `hotkeys.tap_key` is validated, documented, and silently ignored

**Files modified:** `internal/hotkey/fsm.go`, `internal/session/actor.go` (+ tests `internal/hotkey/fsm_test.go`, `internal/session/actor_test.go`)
**Commit:** 7152cd3
**Applied fix:** `FSM.SetTapKey` added as the hot-reload seam (a CHANGED key
disarms the in-flight series — its taps belonged to the replaced key; a
same-key set is a no-op). The actor gained `tapKeyval`/`tapKeyName` state
(constructor default `KeyvalShiftR`); `applySnapshot` re-resolves the name
through `hotkey.ParseBinding` with the comboName last-good discipline;
`HandleKey`'s timer-arming comparison now uses `a.tapKeyval`. The daemon
needs no `cmd/goswitchd` change: with `-config` the watcher snapshot governs
from the FIRST event (applySnapshot runs before every FSM feed), and the
no-config default equals the constructor default. Pinned by
`TestFSM_SetTapKey` and `TestActor_HotReloadTapKey` (a Shift_L-configured
document decides through the REAL AfterFunc timer — expiry never injected).

### CR-02: `timeouts.verify_wait_ms` is validated, documented, and silently ignored

**Files modified:** `internal/session/actor.go` (+ test `internal/session/actor_test.go`, e2e case `test/e2e/cases/matrix-v2.yaml`)
**Commit:** c92996c
**Applied fix:** the package const became `defaultVerifyWait` (the actor's
START value); the budget in force is the `verifyWait` actor field, fed by
`applySnapshot` from `snap.Timeouts.VerifyWaitMs` (>0 guard, last-good
discipline). All four `time.AfterFunc` sites (both verify-after deadlines,
both pending-fix deadlines) now read `a.verifyWait`. Pinned by
`TestActor_HotReloadVerifyWait`: with a 1200 ms document the
verify-timeout skip must NOT land inside the old 100 ms budget and must
land by the new one. The e2e `ctl-status-reload` case now reloads
`verify_wait_ms: 250` (base document is 100) so the "applied" reply reflects
an actual change, not a no-op write.

### CR-03: a deadline-killed wl-paste is misread as the empty clipboard (data loss)

**Files modified:** `internal/clipboard/clipboard.go`, `internal/clipboard/clipboard_test.go`
**Commit:** a8343da
**Applied fix (requires human verification of the classification logic):**
`Save` treats an `*exec.ExitError` as empty ONLY when it is a genuine
non-zero exit — `exitErr.ExitCode() >= 0` (a signal kill reports -1,
verified against a live killed subprocess) — and the caller's context is
not done (`ctx.Err() == nil`). A killed subprocess surfaces as an error, so
`runClipboardRung` takes the `clipboard-unavailable` refusal and never
clears the user's clipboard. Corpus: the "cancelled context" fake now
returns the REAL os/exec kill shape (child started before the wait, killed
by the deadline); `killedPasteErr`/`cleanPasteExitErr` forge the two wire
shapes from real children (the zero-value `&exec.ExitError{}` panics on
`ExitCode()` and was replaced); `TestClipboard_KilledNotEmpty` pins the
signal-kill row and `TestClipboard_DeadlineKillNotEmpty` pins the
production runner end-to-end (a hanging wl-paste script SIGKILLed by the
package's own 1500 ms budget must surface as an error).

### WR-01: the clipboard rung holds the actor mutex for up to ~4.5 s

**Files modified:** `internal/session/actor.go` (+ test `internal/session/actor_test.go`)
**Commit:** 16c4354
**Applied fix (requires human verification of the concurrency restructure):**
`HandleSurroundingText` is split into a mutex-held core
(`handleSurroundingLocked`, returns the rung payload + armed flag) and the
rung dispatch outside the lock. `runClipboardRung` snapshots the emitter and
clipboard client under a brief lock, then runs Save/Set/Ctrl+V/Restore with
NO actor mutex; counters go through `recordSkip` (brief re-acquisition). A
dedicated `rungMu` serializes rungs via `TryLock` — a second mismatch
concluding mid-flight skips with the new `clipboard-rung-busy` reason
instead of blocking or interleaving a second wl-copy/wl-paste pair.
Emitter calls off-mutex are wire-safe (godbus `Emit` is goroutine-safe).
Pinned by `TestActor_ClipboardRungOffKeyPath`: while a wedged wl-paste
parks the rung, key events complete and a concurrent rung attempt skips
busy; after release the single rung finishes its full contract.

### WR-02: `tap_key` validation accepts modifier-bearing bindings

**Files modified:** `internal/config/config.go` (+ test `internal/config/config_test.go`)
**Commit:** ed18c2d
**Applied fix:** `Hotkeys.validate` rejects any "+"-bearing `tap_key` with
the new `errTapKeyBare` sentinel before the `ParseBinding` resolution —
the string rule also catches the degenerate `shift+shift_r` whose ModMask
would equal the bare key's. Corpus rows: "tap key combo
(modifier-bearing)" and "tap key redundant family modifier".

### WR-03: `goswitchNameOwner` authenticates EXTERNAL with the PID

**Files modified:** `test/e2e/matrix.go`
**Commit:** f1f3788
**Applied fix:** `dbus.AuthExternal(strconv.Itoa(os.Getuid()))` — the same
discipline as `engine/conn.go`, `internal/appid`, and `preflight.go`, with
a comment naming why the PID ever worked (SO_PEERCRED tolerance).

### WR-04: workflow interpolates `inputs.matrix` directly into shell

**Files modified:** `.github/workflows/e2e-matrix.yml`
**Commit:** 92fd92e
**Applied fix:** the dispatch input now travels through the step's `env:`
block (`MATRIX: ${{ inputs.matrix }}`) and the shell reads `"$MATRIX"` —
the canonical expression-injection defense; the step name still shows the
input for readability.

### WR-05: the verify-after stale-answer race pastes a duplicate word with the rung on

**Files modified:** `internal/session/actor.go` (+ tests `internal/session/actor_test.go`, comment update in `test/e2e/matrix.go`)
**Commit:** d94a3ce
**Applied fix (requires human verification of the debounce logic):** the
review's primary suggestion — the after-verdict is debounced. The FIRST
mismatching push re-issues one `RequireSurroundingText` and re-arms the
round's deadline (same epoch, same verify budget, new `mismatches` state)
instead of concluding; only a SECOND mismatching push concludes the INFO
verdict (and arms the rung), while a matching push settles quietly. A
client that never answers settles as the quiet timeout — no verdict, no
rung, no duplicate paste. Corpus: `TestActor_VerifyAfterLevel1` now pins
both halves (the debounced first answer re-requires; the intermediate
post-delete push ends in a match) and `settledWordCorrection` drives the
two-stale-answer protocol for the rung corpus. matrix.go's "deferred item"
comment now records the daemon-side debounce.

## Skipped Issues

None — all in-scope findings were fixed.

---

_Fixed: 2026-09-15T15:44:08Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
