---
phase: 06-avtokorrekcija-opcionalno
plan: 10
subsystem: e2e-stand
tags: [golang, e2e, ibus, dbus, ydotool, atspi, flip, gap-closure, live-proof]
requires:
  - phase: 06-09
    provides: the lock-free AttachEngine (atomic emitter slot) — the mechanism this plan proves live on the wire; ADR-006 Amendment 2026-10-02 awaiting its live evidence line
  - phase: 05-05
    provides: the reversible two-source window discipline (unit stop/restoreSpikeWindow, waitNameFree, startDaemonRegistered), the busFlipRound oracle forms (flipMarksPaired, `ibus engine` readback) and the spikePinMode journal-gated pin
provides:
  - e2e case flip-keystroke (runFlipKeystroke): daemon flip → mode record wait → 'd' injected with NO pause → field-tail oracle, both directions ×3, per-round journal gates (paired order D-36 + zero new deadline-abort WARNs)
  - main.go registration (`-case flip-keystroke`) and mise task e2e-flip-keystroke (outside [tasks.ci])
  - live wired evidence for ADR-006 Amendment 2026-10-02: 24/24 immediate letters across 4 consecutive runs (pre-fix probe lost 5/5), INFO-form flips, measured RTT 6.5–9.2 ms
affects: [phase-06-verification, UAT WINDOWS #13 (owner verdict), ADR-006 evidence table]
tech-stack:
  added: []
  patterns:
    - "immediate-keystroke-after-flip oracle: the mode record is the keystroke's starting gun — the letter races the switching window on purpose (the exact window the pre-fix self-block swallowed)"
    - "round-scoped WARN baseline: switchEngineWarnCount gates only NEW aborts since the round's own baseline — environmental WARNs outside the rounds are reported honestly, never gate the case (T-06-10-02)"
key-files:
  created: []
  modified:
    - test/e2e/case_switch.go
    - test/e2e/main.go
    - mise.toml
key-decisions:
  - "G-6-1 live proof landed: the e2e case flip-keystroke injects 'd' with NO pause after the daemon flip's mode record — 24/24 letters delivered across 4 consecutive runs (pre-fix probe lost 5/5)"
  - "Journal audit post-fix: every switch_engine record INFO-form, zero deadline-abort WARNs in a full run, RTT mode→switch_engine 6.5–9.2 ms (median 8.32 ms, n=8) — vs the pre-fix WARN-on-every-flip picture; the 150 ms wedge guard is silent on a healthy flip"
  - "The case never rewrites the sources list (flip = SetGlobalEngine, ADR-006 2026-09-30 amendment) and gates only its own rounds — the formal fresh-session D-48 verdict (WINDOWS #13) stays with the owner on UAT, not replaced by this case"
patterns-established:
  - "Tail oracle over an accumulating field: waitZenityTail polls the content-exact AT-SPI readback for the field's last rune — consecutive rounds always expect the OTHER letter, so a stale tail can never satisfy the poll"
  - "flip-keystroke pause reservation: the only pauses ride AFTER the letter (FSM window 350 ms + settle 200 ms, the probe's form); any pause BEFORE the letter would weaken the oracle (T-06-10-01) and is forbidden"
requirements-completed: [SWCH-01, SWCH-02]
coverage:
  - id: D1
    description: "e2e case flip-keystroke exists, registered in main.go, mise task outside [tasks.ci]: daemon flip → mode record → IMMEDIATE 'd' → field tail equals the direction expectation (en→ru 'в', ru→en 'd'), both directions ×3, paired order D-36 and zero round WARNs asserted per round"
    requirement: SWCH-01
    verification:
      - kind: e2e
        ref: "mise run e2e-flip-keystroke (PASS, 6/6 immediate letters)"
        status: pass
      - kind: other
        ref: "gates: go build/vet ./test/e2e + flip-keystroke in main.go + tasks.e2e-flip-keystroke in mise.toml + NOT-IN-CI grep"
        status: pass
    human_judgment: false
  - id: D2
    description: "Live proof is double-green and regression-clean: two back-to-back flip-keystroke runs (12/12 letters, D-48 form), two-source-flip green (D-36/D-52 pair untouched), journal audit (INFO-form flips, measured RTT) recorded, mise run ci green"
    requirement: SWCH-02
    verification:
      - kind: e2e
        ref: "mise run e2e-flip-keystroke && mise run e2e-flip-keystroke (DOUBLE-GREEN, 12/12)"
        status: pass
      - kind: e2e
        ref: "mise run e2e-two-source-flip (PASS — pair order + factual-engine readback both directions)"
        status: pass
      - kind: other
        ref: "journal audit of a pinned-log run (/tmp/goswitch-06-10-audit.log): 8/8 switch_engine INFO, 0 WARN-level records in the run, RTT 6.50–9.16 ms (median 8.32)"
        status: pass
      - kind: other
        ref: "mise run ci (build + vet + golangci-lint + go test -race ./...)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Fresh-session D-48 gate / WINDOWS #13 verdict (do the matrix flip rows stay green on a session-age-fresh login): deliberately NOT produced by this plan — the case proves direction- and repetition-independence on the current session; the formal session-freshness verdict is the owner's UAT call"
    verification: []
    human_judgment: true
    rationale: "Session freshness is an environmental fact about the owner's login, machine-checkable only by the D-48 fresh_session preflight on the acceptance-nightly path — the plan explicitly leaves this verdict to the owner (WINDOWS #13)"
duration: 16min
completed: 2026-10-03
status: complete
actuals:
  tokens: 4600
  tasks: 2
  commits: 2
  plan_head_before: 02ac2ae1935367452f66391a9643873d4b0e959d
---

# Phase 06 Plan 10: Gap Closure G-6-1 Live Wired Proof Summary

**The flip-keystroke e2e case proves live that the first letter after a daemon flip lands in the field — 24/24 immediate letters across 4 consecutive runs (pre-fix probe lost 5/5), every flip INFO-form with RTT 6.5–9.2 ms and zero deadline-abort WARNs.**

## Performance

- **Duration:** 16 min
- **Started:** 2026-10-03T19:41:33Z
- **Completed:** 2026-10-03T19:57:30Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- G-6-1 closed LIVE on the wire: after a daemon flip (single Shift_R tap) the FIRST letter injected with no pause reaches the field in both directions — en→ru leaves 'в' (RU-conversion of the physical 'd'), ru→en leaves 'd' — deterministically, 6 flips per run, 24/24 letters across the plan's 4 runs (probe v2 baseline: 5/5 losses through the same daemon path)
- Journal form of a healthy flip re-confirmed live: every switch_engine record of the audited run landed INFO (no err member) — the pre-fix picture (WARN `switch_engine context deadline exceeded` on every flip, engine created +1 ms after the abort) is gone; measured RTT mode→switch_engine 6.50–9.16 ms, median 8.32 ms (n=8, incl. 2 setup pin taps; the 6 round flips: 6.85–9.16 ms) — the 150 ms wedge-guard deadline is silent on a healthy flip
- Double green in the D-48 form: two back-to-back `mise run e2e-flip-keystroke` runs, no intervention between — 12/12 immediate letters
- Pair contract regression: `mise run e2e-two-source-flip` green — the D-36 record order (mode, then switch_engine naming the target engine) and the D-52 factual-engine readback did not move under the 06-09 fix
- The oracle was NOT weakened anywhere: injection stays immediate after the mode record (T-06-10-01), direction expectations are the probe's, per-round WARN gates count only NEW aborts from the round's own baseline (T-06-10-02); the pause-constant reserve of Task 2 stayed unused — case_switch.go was not touched after Task 1's commit

## Task Commits

1. **Task 1: e2e case flip-keystroke + registration + mise task** - `9f78017` (feat)
2. **Task 2: double green + two-source-flip regression + journal audit** - no commit (verification-only: zero file delta; the case_switch.go pause-constant reserve stayed unused per the plan's own condition)

**Plan metadata:** docs commit (this SUMMARY + STATE/ROADMAP/REQUIREMENTS).

_Note: Task 2's evidence lives in this SUMMARY and the live runs; the production code of the daemon was not touched by this plan at all._

## Live Results

| Run | Case | Rounds | Result | Letters |
| --- | ---- | ------ | ------ | ------- |
| 1 (Task 1 verify) | flip-keystroke | en→ru ×3, ru→en ×3 | PASS | 6/6 |
| 2 (Task 2, run 1) | flip-keystroke | en→ru ×3, ru→en ×3 | PASS | 6/6 |
| 3 (Task 2, run 2, back-to-back) | flip-keystroke | en→ru ×3, ru→en ×3 | PASS | 6/6 |
| 4 (Task 2, journal audit, pinned log) | flip-keystroke | en→ru ×3, ru→en ×3 | PASS | 6/6 |
| 5 (regression) | two-source-flip | en→ru, ru→en | PASS | — (pair order + readback) |

**Total: 24/24 immediate letters vs the pre-fix 5/5 LOST baseline (`/tmp/gsy-settle2.sh`, UAT G-6-1). Double-green verdict: PASS (D-48 form). Regression verdict: PASS.**

**Journal audit (run 4, `/tmp/goswitch-06-10-audit.log`):**

- 8 switch_engine records, ALL `level: INFO` (no err member); zero WARN-level records of any kind in the whole run
- RTT mode record → switch_engine record per flip: 6.50, 6.85, 7.34, 7.44, 8.32, 8.34, 9.10, 9.16 ms — median 8.32 ms (flips 1–2 are the setup's journal-gated EN pin taps; the 6 flip-keystroke rounds are flips 3–8: 9.16, 7.44, 7.34, 6.85, 8.34, 9.10 ms)
- Old picture for contrast (UAT journal 00:05:04–00:05:17): WARN on EVERY flip, `engine created` +1 ms after the abort, 5/5 keys swallowed
- Honest note: no WARN occurred anywhere in any run, inside or outside the rounds — nothing to excuse

## Files Created/Modified

- `test/e2e/case_switch.go` — `runFlipKeystroke` + `flipKeystrokeRound` + `waitZenityTail` + `expectedFlipLetter` + `switchEngineWarnLine`/`switchEngineWarnCount` + timing constants (flipKeystrokeReps=3, flipKeyFsmWait=350 ms, flipKeySettle=200 ms, flipKeystrokeKey='d')
- `test/e2e/main.go` — `-case flip-keystroke` registration (standalone, default watchdog), usage + error listings
- `mise.toml` — `[tasks.e2e-flip-keystroke]` beside the switch-family tasks, outside `[tasks.ci]` (NOT-IN-CI gate green)

## Decisions Made

- Setup establishes the probe's precondition explicitly: the case activates goswitch-en via the stand's `activateGoswitch` (external SetGlobalEngine) before the journal-gated EN pin — the unit stop leaves the global engine unset on the stand, while the owner's live probe session always carried one (see Deviations)
- Sources list never rewritten (per plan): the flip is a SetGlobalEngine and works with any registered engine; the audit run proved it on the stand's own desktop
- The fresh-session verdict (WINDOWS #13) is left to the owner's UAT pass — this plan's numbers feed it, they do not replace it

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Setup was missing engine activation — taps routed nowhere without it**
- **Found during:** Task 1 (case design against the stand's lifecycle)
- **Issue:** the plan's setup list (unit stop → waitNameFree → startDaemonRegistered → openEntrySurface → spikePinMode) omits activating a goswitch engine; on the stand the unit stop leaves the global engine unset (ibus-daemon unsets it when an engine connection closes — phase 01-03 finding), so Shift_R taps route nowhere and `spikePinMode` can never see a mode record. The owner's probe never hit this because the owner's live session had a goswitch engine already global
- **Fix:** one `activateGoswitch` call (the word cases' external SetGlobalEngine + focus_in wait) between the surface opening and the pin — it never rewrites gsettings, so the plan's "no sources rewrite" constraint holds, and the sync correction it provokes gives the journal its first mode record
- **Files modified:** test/e2e/case_switch.go (part of the Task 1 commit)
- **Verification:** all 4 live runs green; pin settles with ≤2 taps per the existing discipline
- **Committed in:** 9f78017

---

**Total deviations:** 1 auto-fixed (1 blocking).
**Impact on plan:** the activation step makes the case deterministic on the stand without touching the oracle; no scope creep.

## Issues Encountered

- `mise run ci` hit the known `internal/ctlsvc` conn-hook flake once per task gate ("post-export connection hook failed: hook exploded"); each re-run — the single sanctioned retry — was fully green across all packages. Logged per the green-iteration discipline, no code action
- The four live runs and the regression left the desktop clean: gsettings restored, unit goswitchd `active` after every run (the case's restoreSpikeWindow discipline held)

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- ADR-006 Amendment 2026-10-02 now has its missing live evidence line: the mechanism fix (06-09) is proven on the wire — the UAT G-6-1 issue can be closed on these numbers, with WINDOWS #13 (fresh-session matrix verdict) remaining the owner's call
- The flip-keystroke case is a permanent regression gate for the flip path's key-delivery property: any return of the self-block (or a wedged bus) fails the round with the journal excerpt and a named WARN count
- Phase 06 verification can proceed: all 8 plans of the phase have SUMMARYs; the remaining UAT items (2–5) are owner-judgment passes

## Self-Check: PASSED

- All 3 modified files exist on disk (case_switch.go, main.go, mise.toml) plus this SUMMARY
- Task commit 9f78017 present in git log; wiring verified live (case registered, mise task present, NOT-IN-CI)
- Live claims spot-checked: 5 runs table matches run output above; audit numbers parsed from the pinned log (/tmp/goswitch-06-10-audit.log); unit goswitchd `active` after all runs
- Ledger measure: 1 code commit before this docs commit (plan_head_before 02ac2ae → 9f78017); final count incl. docs commit: 2

---
*Phase: 06-avtokorrekcija-opcionalno*
*Completed: 2026-10-03*
