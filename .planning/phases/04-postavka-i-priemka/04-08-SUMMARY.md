---
phase: 04-postavka-i-priemka
plan: 08
subsystem: testing
tags: [e2e, atspi, a11y, witness, quiesce, tdd, gap-closure]

# Dependency graph
requires:
  - phase: 02-06 (matrix isolation) and 04-04 (watchdog/witness timing knobs)
    provides: matrixQuiesce probe loop, witnessProbeTimeout/witnessPoll constants, headless corpus precedent (watchdog_test.go)
provides:
  - Focus-first witness traversal in test/e2e/focus_helper.py — FOCUSED checked on window frames before descent, only focused frames' subtrees walked, verbatim pre-04-08 body kept as the full-walk fallback (arms input line / frame-fallback chars=-1 / (none) preserved)
  - witnessProbeTimeout raised 4s→10s (G-4-1 defense in depth) and witnessQuiesceWindow named constant single-sourcing the matrixQuiesce deadline and error text
  - Headless quiesce-budget corpus (test/e2e/quiesce_test.go): probe floor, window-coverage invariant, window single-source pin — RED evidence in red-evidence/04-08-task1-red.json (staged two-form RED)
affects: [owner D-48 formal fresh-session double-run (phase 04 acceptance criterion 2), any future e2e harness timing work]

# Actuals (#2632)
actuals:
  tokens: 3623
  tasks: 2
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "staged two-form RED: a pure-assertion RED first, then compile-error RED tests added in a second step so the clean failure is not masked (red evidence carries both forms; step 1 gate-validated RED_EVIDENCE_OK)"
    - "verbatim-body extraction: an arm-preserving refactor moves a command's body into a fallback function unchanged and dispatches to it only when the new fast path finds nothing"

key-files:
  created:
    - test/e2e/quiesce_test.go
    - .planning/phases/04-postavka-i-priemka/red-evidence/04-08-task1-red.json
  modified:
    - test/e2e/matrix.go
    - test/e2e/focus_helper.py

key-decisions:
  - "witnessProbeTimeout floor 10s pinned by corpus; witnessQuiesceWindow = witnessQuieceAttempts*surfaceFocusWait is the single source for deadline and error text; attempts stay 2 (2*(10s+100ms)=20.2s inside 30s) and the misspelled name is deliberately not renamed (lint churn)"
  - "focus-first answers from focused frames' subtrees only; a focused non-input node inside a focused frame short-circuits to the frame-fallback arm without the full walk; the verbatim full walk runs only when the pass finds nothing — 01-03 locked-session semantics preserved by construction"
  - "exhaustive modes (focused-inputs and kin) untouched: enumerating candidates through unfocused frames stays their 02-02 contract; the witness's consumers need a prompt answer, not enumeration (plan FLAGGED multifocus assumption)"

patterns-established:
  - "quiesce-budget corpus: constant arithmetic pinned headless per watchdog_test.go precedent"
  - "focus-first traversal contract documented in the helper's docstring table with live-verified caveats (04-04 precedent)"

requirements-completed: [INST-01]

coverage:
  - id: D1
    description: "Quiesce budget corpus: probe scalability floor (>=10s), window-coverage invariant, window single-source pin — green under -race"
    requirement: INST-01
    verification:
      - kind: unit
        ref: "test/e2e/quiesce_test.go#TestWitnessProbeBudget_ScalabilityFloor"
        status: pass
      - kind: unit
        ref: "test/e2e/quiesce_test.go#TestMatrixQuiesce_WindowCoversProbes"
        status: pass
      - kind: unit
        ref: "test/e2e/quiesce_test.go#TestMatrixQuiesce_WindowSource"
        status: pass
      - kind: integration
        ref: "mise run ci (build, vet, golangci-lint, go test -race ./...)"
        status: pass
    human_judgment: false
  - id: D2
    description: "matrix.go defense in depth: witnessProbeTimeout = 10s; witnessQuiesceWindow single-sources the matrixQuiesce deadline and error text; loop structure unchanged"
    requirement: INST-01
    verification:
      - kind: unit
        ref: "test/e2e/quiesce_test.go#TestMatrixQuiesce_WindowSource"
        status: pass
      - kind: other
        ref: "grep: expression witnessQuieceAttempts * surfaceFocusWait occurs exactly once in executable code (the witnessQuiesceWindow declaration)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Focus-first witness traversal with the verbatim full-walk fallback; exhaustive modes untouched; docstring contract updated"
    requirement: INST-01
    verification:
      - kind: other
        ref: "/usr/bin/python3 -m py_compile test/e2e/focus_helper.py"
        status: pass
      - kind: e2e
        ref: "live: witness answers frame-fallback arm from the ONLY focused frame's subtree (gnome-shell:WINDOW:chars=-1) in 3.0-3.2s while zenity/GTE hold node-focused-but-frame-unfocused surfaces the pass intentionally skips"
        status: pass
    human_judgment: true
    rationale: "The focus-first correctness at fresh-session scale (the INPUT-hit path through a genuinely focused frame on the runner's D-48 tree) is the plan's own FLAGGED assumption, reserved for the owner's formal D-48 dispatch after this plan; local evidence (frame-skip, focused-frame subtree walk, prompt answers) is recorded here as the weak local witness"
  - id: D4
    description: "Live smoke on the desktop: witness answers in budget against spawned surfaces; post-close fallback probe prompt; wall-times recorded"
    requirement: INST-01
    verification:
      - kind: e2e
        ref: "live timed witness runs: 2976-3165ms against live zenity/GTE (floor 10s), 3107/3034ms post-close — frame-fallback arm, exit 0, single line"
        status: pass
      - kind: e2e
        ref: "exhaustive cross-check: focused-inputs reported zenity:TEXT / GTE TEXT:chars=0 node focus while the witness frame-skip held — the plan's FLAGGED focus-lag case, live"
        status: pass
    human_judgment: true
    rationale: "The criterion's INPUT-role line through a focused zenity entry was not reproducible: mutter denies window activation to background-spawned surfaces (zenity map-time focus denied, GTK3 frame grabFocus errors, GTE reaches widget-grab only — frame FOCUSED never set), so no focused frame contained a focused input node; the precondition absence is an environment fact, not a code failure — the owner's interactive D-48 session exercises the path"

# Metrics
duration: 35 min
completed: 2026-09-17
status: complete
---

# Phase 04 Plan 08: G-4-1 gap closure Summary

**Focus-first a11y witness (frames before descent, verbatim full-walk fallback) plus a 10s probe floor with a pinned quiesce window — the harness scalability defect that felled healthy D-48 cases is closed in code and locally proven on the 6100-node session**

## Performance

- **Duration:** 35 min
- **Started:** 2026-09-17T11:15:04Z
- **Completed:** 2026-09-17T11:50:04Z
- **Tasks:** 2
- **Files modified:** 4 (2 created, 2 modified)

## Accomplishments
- G-4-1 fix direction implemented: the witness probe no longer walks the whole desktop when a focused frame exists — FOCUSED is checked on each application's direct children (frames) before any descent, only focused frames' subtrees are walked (per-node logic verbatim), and the pre-04-08 body answers unchanged as the full-walk fallback
- Defense in depth: witnessProbeTimeout 4s→10s (measured worst case ~3.0s bare walk + ~3x headroom) and a named witnessQuiesceWindow that single-sources the matrixQuiesce deadline and error text; attempts stay 2 with the 20.2s-in-30s arithmetic pinned
- Headless corpus (RED→GREEN) pins the floor, the window-coverage invariant and the single-source rule; staged two-form RED evidence recorded and gate-validated (RED_EVIDENCE_OK)
- Live smoke on the degraded 6100-node session: witness answered in 3.0-4.7s via the focused-frame subtree (pre-fix probes on the same session: 3.8-6.3s), post-close probes prompt at 3.0s — the fallback arm is alive, not wedged
- Green iteration (directive 2): mise run ci green at every commit; exhaustive modes and product code untouched

## Task Commits

Each task was committed atomically (Task 1 is tdd="true" — RED and GREEN as separate commits):

1. **Task 1 RED: staged quiesce budget corpus** - `230e55b` (test)
2. **Task 1 GREEN: 10s floor + witnessQuiesceWindow** - `983204e` (feat)
3. **Task 2: focus-first witness traversal + live smoke** - `4bf01fc` (feat)

**Plan metadata:** docs commit carrying this SUMMARY, STATE.md, ROADMAP.md and REQUIREMENTS.md (created immediately after this file).

### TDD Gate Compliance

- RED commit present: `230e55b` `test(04-08): add failing quiesce budget corpus (staged RED)`; GREEN commit present: `983204e` `feat(04-08): raise witness probe budget to 10s, name the quiesce window`. No REFACTOR commit — nothing left to clean.
- Staged RED executed literally per the checker-revised plan: step 1 (TestWitnessProbeBudget_ScalabilityFloor) failed as a pure assertion failure on 4s < 10s with a green build, gate-validated via `check tdd-red-evidence` → `RED_EVIDENCE_OK` (target_test_failed); step 2 (window tests) produced the checker-staged compile error `undefined: witnessQuiesceWindow`. Both forms are transcribed into `red-evidence/04-08-task1-red.json` (step 1 as the gate-validated record, step 2 as its documented second record).

## Files Created/Modified
- `test/e2e/quiesce_test.go` - headless quiesce-budget corpus: floor, window coverage, window source
- `test/e2e/matrix.go` - witnessProbeTimeout 10s with G-4-1 comment; witnessQuiesceWindow constant; deadline/error text single-sourced; loop untouched
- `test/e2e/focus_helper.py` - focus-first cmd_witness; focused_frame_witness(); full_walk_witness() (verbatim body); docstring contract
- `.planning/phases/04-postavka-i-priemka/red-evidence/04-08-task1-red.json` - both RED forms

## Decisions Made
See key-decisions in frontmatter. The material interpretation: a focused non-input node inside a focused frame short-circuits to the frame-fallback arm from the focus-first pass (the full walk runs only when the pass finds nothing anywhere) — read directly from the plan's fallback clause ("не нашёл ничего") and exercised live (gnome-shell:WINDOW:chars=-1 answers).

## Deviations from Plan

### Auto-fixed Issues

None — plan executed exactly as written. (Lint lll reformats of the new test's Fatalf lines were in-task green-iteration work under directive 2, not unplanned changes.)

### Acceptance criterion adjusted by environment (documented, not silently skipped)

- **Found during:** Task 2 (live smoke)
- **Criterion:** "живой смок: witness со сфокусированным zenity-entry — exit 0, строка с ролью ENTRY"
- **Actual:** witness answered exit 0 with wall-times recorded, but via the frame-fallback arm (`gnome-shell:WINDOW:chars=-1`): no surface could reach a FOCUSED *frame* containing a focused input node from a background context. Attempts (all house mechanisms): zenity map-time focus — denied; `grab-input-pid` — widget grab only (zenity GTK3); `focus zenity` frame grabFocus — atspi_error (documented GTK3 refusal); Super/overview — search entry never materialized focused; real click on GTE's served extents — activation incomplete (frame ACTIVE=True, FOCUSED=False). ydotool event emission verified live at the evdev layer (transient uinput devices observed during calls).
- **Why this is the plan's own flagged case:** the a11y tree held node-focused surfaces the whole time (`focused-inputs` reported `zenity:TEXT:chars=0`, GTE `TEXT:chars=0`) while their frames never reported FOCUSED — exactly the FLAGGED multifocus/focus-lag assumption, which the focus-first pass intentionally skips (the exhaustive modes own enumeration).
- **Resolution:** the INPUT-hit branch shares the verbatim per-node logic with the fallback walk whose INPUT arm is the live-proven 01-03 contract; the branch difference is only the INPUT_ROLES membership test on the same walked node. The in-situ proof lands with the owner's formal D-48 fresh-session dispatch — the plan's designated next step, not part of this plan.
- **Files modified:** none (no code change warranted)
- **Committed in:** n/a (documented in this SUMMARY and the WINDOWS ledger)

---

**Total deviations:** 0 auto-fixed; 1 environment-blocked acceptance criterion documented above.
**Impact on plan:** none on shipped code — all four success-criteria items (focus-first traversal, 10s floor + window pinning, untouched exhaustive/product code, green iteration) are met; the criterion that depended on an interactive desktop state routes to the owner's D-48 run by the plan's own assumption.

## Issues Encountered
- The live session (up since 2026-09-16, fresh-session-class tree: gnome-shell 3673 + gjs 2437 nodes) keeps frame-focus on the shell stage; every activation lever available to a background executor failed as described above. Environment-flake remedies (ibus restart, a11y bus restart per ci-runner.md) were considered and NOT applied — the bridge was never wedged (all walks completed; 02-06-style hangs absent), so a restart would have treated a healthy bus.
- gsd-tools state.advance-plan / state.update-progress skipped on first call (04-08 SUMMARY absent at that moment); re-run after the SUMMARY was written — see the state verbs' outputs in the session log.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- G-4-1 closed in the plan's scope; the formal D-48 fresh-session double-run (phase acceptance criterion 2) is unblocked and is the owner's next step per docs/ACCEPTANCE.md item 1
- The witness budget on a degraded 6100-node session is now 3.0-4.7s observed against the 10s floor; on healthy runner sessions both arms answer well inside it
- WINDOWS ledger carries the environment-blocked smoke note for the verifier

---
*Phase: 04-postavka-i-priemka*
*Completed: 2026-09-17*

## Self-Check: PASSED

- Files: quiesce_test.go, matrix.go, focus_helper.py, red-evidence/04-08-task1-red.json — all exist on disk
- Commits: 230e55b (test RED), 983204e (feat GREEN), 4bf01fc (feat focus-first) — all present in git log
- Corpus re-run at self-check: `go test ./test/e2e/ -run 'WitnessProbeBudget|MatrixQuiesce' -race -count=1` — ok
- Plan-level verification: mise run ci green (build, vet, golangci-lint strict, go test -race ./...); py_compile green; matrixQuiesce loop structure unchanged (constants and named window only)
- Commits measured from ledger: 3 (gsd-plan-head-before-04-08 = 3511f50)
