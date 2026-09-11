---
phase: "01"
plan: "04"
subsystem: adr-package
tags: [adr, experiment, layout-switching, tap-semantics, capability-ladder, buffer-reset, macr-01, spec-delta, m0-gate]
requires:
  phase: "01"
  provides: "e2e stand skeleton (plan 01-03), engine seam (plan 01-01), hotkey FSM + layout tables (plan 01-02)"
provides:
  - "docs/adr/d01-experiment-log.md — evidence journal: per-probe rows (command, observable, verdict), 15 Verdict lines across two full runs, DECISION row (Option B)"
  - "test/e2e -case d01-probe + mise task e2e-d01 — the repeatable D-01 probe matrix with observable asserts and machine-checked restore"
  - "docs/adr/ADR-001..ADR-005 (Accepted at M0) — layout switching (Option B internal flip), tap semantics (classic-with-wait, 300 ms), replacement capability ladder, buffer reset triggers, MACR-01 mechanism (global rules + per-app YAML lists + AT-SPI identity, Super-intercept degradation ladder)"
  - "docs/SPEC.md — owner's M0-era edits (§4.1 Right Shift tap defaults, §4.5 macros requirement, go install) + applied spec-deltas: §5 NFR reformulation, §4.3 buffer reset without mouse click"
affects:
  - docs/SPEC.md
  - engine/engine.go
  - engine/factory.go
actuals:
  tokens: 20210
  tasks: 3
  commits: 3
  plan_head_before: 5781f7b867a32b3e611c45c2a4d439c2419388de
tech-stack:
  added: []
  patterns:
    - "observable-only proof: a probe verdict requires the FocusOut/FocusIn engine pair in the daemon log plus typed-text readback; dconf values are never evidence (keyboard.js 46 ignores external writes)"
    - "journal-as-evidence-chain: the e2e case appends Verdict rows to the ADR journal; ADR-001 must cite the journal — decisions stay traceable to live observations"
    - "spec-delta flow: ADR carries the ready delta text; the text is applied to SPEC only after the owner gate approves it"
key-files:
  created:
    - test/e2e/case_d01.go
    - docs/adr/d01-experiment-log.md
    - docs/adr/ADR-001-layout-switching-mechanism.md
    - docs/adr/ADR-002-tap-semantics.md
    - docs/adr/ADR-003-replacement-capability-ladder.md
    - docs/adr/ADR-004-buffer-reset-triggers.md
    - docs/adr/ADR-005-macr-01-owner-checkpoint.md
  modified:
    - test/e2e/main.go
    - test/e2e/README.md
    - engine/engine.go
    - engine/factory.go
    - engine/engine_test.go
    - mise.toml
    - docs/SPEC.md
key-decisions:
  - "D-01 resolved by experiment: NO probe produced the double observation (focus pair + Cyrillic readback) — Option B internal flip wins; gsettings writes are runtime-ignored, ydotool combos miss our engines, SetGlobalEngine switches the engine but never the XKB group, Shell.Eval refuses without unsafe mode, micro-extension requires re-login"
  - "M0 gate passed 2026-09-10: owner approved the ADR pack verbatim («утверждено») — no edits, no alternative MACR-01 mechanism requested; verbatim approval recorded in ADR-005 Status"
  - "ADR-001..005 Accepted; spec-deltas applied to SPEC §5 (waitless < 50 ms vs Right Shift bounded by the discrimination window) and §4.3 (mouse click removed, mandatory surrounding-text check, best-effort cursor-jump reset)"
  - "ADR-005 mechanism: global rules + optional per-app YAML lists as primary; AT-SPI observer as the app-identity fallback source; FocusIn hints rejected (FocusIn carries no arguments on 1.5.29); Super-intercept gets a documented degradation ladder (detect consumed combo, WARN log + goswitchctl visibility, configurable alternative modifier)"
  - "STATE blocker «Decision #1» closed by the kill-criterion spike before any correction code (Phase 2 builds on a decided contract)"
patterns-established:
  - "probe-matrix e2e case shape: snapshot/restore desktop state, per-probe observable asserts, DECISION row derived in-journal, teardown re-reads both gsettings keys and fails on drift"
  - "ADR format for this repo: Status / Context (CONTEXT decisions quoted verbatim) / Decision / Consequences / Reversibility (ratings quoted from CONTEXT)"
requirements-completed: [TEST-02, INTEG-04]
coverage:
  - id: d01-probe-case
    description: "e2e case d01-probe: five-probe matrix with observable verdicts, machine-checked restore, journal append; mise task e2e-d01 (live session, not in ci)"
    requirement: TEST-02
    verification:
      - kind: e2e
        ref: "mise run e2e-d01 → PASS at checkpoint-stop state (RESTORED printed, extension dir EXT REMOVED, 15 Verdict lines); no Go code changed after that run (both later commits are docs-only)"
        status: pass
      - kind: unit
        ref: "mise run ci → exit 0 (re-run after continuation docs commit, 2026-09-11)"
        status: pass
    human_judgment: false
  - id: d01-journal
    description: "Experiment journal: every executed probe has a row with command, observable (not dconf), verdict; decision derivable (Option B)"
    verification:
      - kind: automated
        ref: "grep -c 'Verdict|Вердикт' docs/adr/d01-experiment-log.md = 15 ≥ 4 → VERDICTS-OK"
        status: pass
    human_judgment: false
  - id: adr-pack
    description: "Five ADRs in unified format; ADR-001 cites the journal and decides Option B with D-03 consequences and the A3/INTEG-04 re-activation requirement; ADR-005 carries the D-12 mechanism (app detection + Super-intercept degradation ladder), no MACR code in Phase 1"
    requirement: INTEG-04
    verification:
      - kind: automated
        ref: "sections check CHECKED (Status/Decision/Consequences/Reversibility in all 5); ADR-001→journal ref = 1; ADR-005 D-12 = 3, FocusIn = 3, degradation ladder = 2"
        status: pass
    human_judgment: false
  - id: m0-gate
    description: "Blocking owner checkpoint M0: ADR pack review including the MACR-01 mechanism and both spec-deltas"
    verification:
      - kind: manual_procedural
        ref: "Owner response at the checkpoint (2026-09-10): «утверждено» — recorded verbatim in ADR-005 Status; resolved the blocking-human gate, continuation verified commits 145d0d3 + a82b3c1 before applying deltas"
        status: pass
    human_judgment: false
  - id: spec-deltas-applied
    description: "SPEC §5 NFR reformulation (ADR-002 text) and §4.3 buffer-reset rewrite (ADR-004 text) applied on top of the owner's uncommitted working-tree edits (preserved, not reverted); macros section renumbered 4.4→4.5 to fix a duplicate section number"
    verification:
      - kind: automated
        ref: "grep: §4.3 line carries the ADR-004 delta text; §5 line carries the ADR-002 delta text; single 4.4 + single 4.5 headings; ADR-005 refs updated to §4.5"
        status: pass
    human_judgment: false
duration: 40 min active
completed: 2026-09-11T10:57:00+03:00
status: complete
---

# Phase 01 Plan 04: ADR-пакет и каркас IBus-движка — эксперимент D-01 + ADR-пак + гейт M0 Summary

D-01 experiment journal with a 15-probe verdict matrix and ADR-001..005 resolving every phase fork (Option B internal flip), owner-approved at gate M0 with spec-deltas §5/§4.3 applied to SPEC.

## Performance

- Duration: ~40 min active across two agents (00:18–00:50 experiment + ADR pack; 10:52–10:57 continuation: spec-deltas, ADR statuses, verification, close-out); the overnight gap is the blocking M0 checkpoint awaiting the owner
- Started: 2026-09-11T00:18 (+03:00, after 5781f7b)
- Completed: 2026-09-11T10:57 (+03:00)
- Tasks: 3/3 (task 3 = the M0 checkpoint, resolved by owner approval)
- Files: 14 changed (+1039/−35)

## Accomplishments

- **Task 1:** `d01-probe` e2e case + `mise run e2e-d01` — five-probe matrix (gsettings write, ydotool super+space, ydotool alt+shift_l, `ibus engine` SetGlobalEngine, Shell.Eval, conditional micro-extension) with observable-only criteria (FocusOut/FocusIn pair in the daemon log + typed-text readback), journal rows appended per run, machine-checked teardown (both gsettings keys re-read and compared; drift → FAIL). Canonical run PASS: RESTORED printed, extension dir removed, 15 Verdict lines. Decision derivable: zero switched probes → **Option B (internal flip)**.
- **Task 2:** ADR pack 002→005 first, ADR-001 last from the journal. ADR-002 tap semantics (classic-with-wait, 300 ms window, modifier discrimination, fourth-tap rule, fsm_test.go as executable spec); ADR-003 replacement ladder (DeleteSurroundingText → Backspace×N in runes with abort-don't-garbage → clipboard, selected per-client by caps bit 1<<5); ADR-004 buffer reset (D-06: mouse click out, mandatory surrounding-text check, best-effort cursor-jump reset; ready §4.3 delta text); ADR-005 MACR-01 mechanism per D-12 (global rules + per-app YAML lists primary, AT-SPI identity fallback, FocusIn hints rejected on 1.5.29 wire evidence, Super-intercept degradation ladder). ADR-001 written last: Option B wins, kill-criteria table with observed failures per candidate (incl. the keyboard.js 46 finding: external SetGlobalEngine never moves the XKB group), D-03 "one owner" consequences, A3/INTEG-04 re-activation requirement.
- **Task 3 (M0 gate):** owner approved the pack («утверждено», 2026-09-10). Continuation applied: spec-delta §5 (waitless actions < 50 ms; Right Shift actions bounded by the discrimination window, 300 ms default) and §4.3 (buffer reset without mouse click) on top of the owner's own uncommitted SPEC edits; ADR-001..005 Status Proposed → Accepted; ADR-005 carries the verbatim approval record. Plan verification re-run green.

## Task Commits

| Task | Type | Hash | Subject |
|------|------|------|---------|
| 1 | feat | 145d0d3 | D-01 probe matrix — source-switch experiment with observable verdicts |
| 2 | docs | a82b3c1 | ADR pack — five architecture decisions, ADR-001 last from the journal |
| 3 | docs | 0a48f76 | M0 gate passed — spec-deltas applied, ADR-001..005 Accepted |

## Files Created/Modified

Created: `test/e2e/case_d01.go`, `docs/adr/d01-experiment-log.md`, `docs/adr/ADR-001-layout-switching-mechanism.md`, `docs/adr/ADR-002-tap-semantics.md`, `docs/adr/ADR-003-replacement-capability-ladder.md`, `docs/adr/ADR-004-buffer-reset-triggers.md`, `docs/adr/ADR-005-macr-01-owner-checkpoint.md`.
Modified: `test/e2e/main.go`, `test/e2e/README.md`, `engine/engine.go`, `engine/factory.go`, `engine/engine_test.go`, `mise.toml`, `docs/SPEC.md`.

## Decisions Made

- **Option B — internal flip** (ADR-001): no probe gave the double observation; one engine `goswitch` with `<layout>us</layout>`, EN/RU as internal daemon mode; XKB group never changes; SetGlobalEngine stays a stand/diagnostics tool, not a layout mechanism.
- **M0 approved «утверждено»** (2026-09-10): the whole pack accepted as written — no edits, no alternative MACR-01 mechanism requested; recorded verbatim in ADR-005 Status.
- **Spec-deltas applied:** §5 performance NFR now distinguishes waitless actions (< 50 ms) from Right Shift actions (bounded by the discrimination window, 300 ms default, configurable); §4.3 buffer reset drops the unobservable mouse-click trigger in favor of the mandatory pre-correction surrounding-text check and best-effort cursor-jump reset.
- **MACR-01 mechanism fixed** (ADR-005): global rules + optional per-app lists in YAML (primary), AT-SPI observer for app identity (fallback), documented degradation ladder for GNOME-consumed Super+Буква (detect → WARN log + goswitchctl visibility → configurable alternative modifier); no MACR code in Phase 1.
- **Owner's working-tree SPEC edits preserved and committed as the base** (§4.1 defaults now Right-шift taps, new macros requirement section, `go install` delivery); the new macros section renumbered 4.4 → 4.5 to resolve a duplicate section number (see Deviations #5).

## Deviations from Plan

**1. [Rule 1 - Bug] Malformed gsettings sources value in probe 1**
- Found during: Task 1 (first live run)
- Issue: the probe wrote an invalid GVariant to `sources`, so the probe itself was testing garbage.
- Fix: corrected the value construction; the journal row was re-run and now carries the valid-form observation.
- Files: `test/e2e/case_d01.go`
- Verification: re-run verdict row in `docs/adr/d01-experiment-log.md`; canonical `mise run e2e-d01` PASS with RESTORED.
- Commit: 145d0d3

**2. [Rule 1 - Bug] Live hang of run 2 (unbounded external-CLI call)**
- Found during: Task 1 (second run)
- Issue: an external CLI call in the case had no deadline and hung the live run.
- Fix: structural bounding — 15 s `runCmd` deadline for every external call, 180 s per-case watchdog, per-probe progress prints for diagnosis.
- Files: `test/e2e/case_d01.go`, `test/e2e/main.go`
- Verification: subsequent runs completed within bounds; canonical PASS.
- Commit: 145d0d3

**3. [Rule 2 - Missing critical observability] Engine name missing from the Engine object and focus logs**
- Found during: Task 1
- Issue: `focus_in`/`focus_out` log records did not identify which engine fired — the probe's observable criterion (the FocusOut(en)/FocusIn(ru) pair) was unreadable.
- Fix: engine `name` added to the Engine object and included in focus_in/focus_out log records.
- Files: `engine/engine.go`, `engine/factory.go`, `engine/engine_test.go`
- Verification: journal observables cite `focus_out(en)+N / focus_in(ru)+N`; `mise run ci` green including engine tests.
- Commit: 145d0d3

**4. [Minor - Scope] README updated beyond the files_modified list**
- Found during: Task 1
- Issue: `test/e2e/README.md` was updated (d01 case table + live-session tasks) though not listed in the plan's files_modified.
- Fix: kept — the stand's run book must cover the new live case; documented here as a scope note.
- Files: `test/e2e/README.md`
- Verification: README describes e2e-d01 and the live-session caveats.
- Commit: 145d0d3

**5. [Rule 1 - Bug] Duplicate SPEC section number after the owner's edit (continuation)**
- Found during: Task 3 (spec-delta application)
- Issue: the owner's uncommitted edit added "### 4.4 Коррекция клавиатурных макросов" while "### 4.4 Интеграция с окружением" already existed (referenced as §4.4 by the committed 01-03-PLAN); ADR-005 also referenced §4.4 for macros — ambiguous spec references.
- Fix: owner's new section renumbered to §4.5 (content and position untouched, owner's edits otherwise preserved verbatim); ADR-005's two references updated to §4.5; the pre-existing §4.4 (Интеграция) keeps its number so committed plan references stay valid.
- Files: `docs/SPEC.md`, `docs/adr/ADR-005-macr-01-owner-checkpoint.md`
- Verification: grep — single `### 4.4` + single `### 4.5` heading; no other §4.4-for-macros references in docs/ or .planning/.
- Commit: 0a48f76

Totals: 5 deviations (3× Rule 1, 1× Rule 2, 1 minor scope note). None architectural (Rule 4); all documented and verified.

## Issues Encountered

- The blocking M0 checkpoint split the plan across two executors (tasks 1–2 before, delta application after approval); continuation verified commits 145d0d3/a82b3c1 and the clean-except-owner-edits tree before proceeding.
- `mise run e2e-d01` was not re-run in the continuation: both post-run commits (a82b3c1, 0a48f76) are documentation-only, so the checkpoint-time PASS (RESTORED + extension removed) remains the canonical evidence; `mise run ci` was re-run green after the final docs commit.
- Journal timestamps show wall-clock times of the live runs (00:34–00:46); two full matrix runs are recorded (10 verdict rows + 2 DECISION rows + 3 auxiliary verdict-bearing rows).

## User Setup Required

- None for this plan. (Carried from 01-03, not reopened here: the INTEG-02 live keyd checklist in `test/e2e/README.md` remains an owner-manual item.)

## Next Phase Readiness

- Every Phase-01 architecture fork is closed with an Accepted ADR: Phase 2 correction code builds on the capability ladder (ADR-003), the buffer contract (ADR-004), and the corrected SPEC §4.3/§5 wording; Phase 3 planning inherits the MACR-01 mechanism and YAML shape requirements from ADR-005.
- The STATE blocker «Decision #1» is closed by the kill-criterion spike before any correction code was written.
- D-03 "one owner" UX documentation obligations (single goswitch source, indicator non-learning) are recorded in ADR-001 Consequences for the Phase 4 docs.
- Stale-name ibus race and keycode/keyval semantics on the live IM path remain noted in 01-03-SUMMARY for Phase 2 observation.

## Self-Check: PASSED

- Commits 145d0d3, a82b3c1, 0a48f76 present on gsd/phase-01-adr-paket-i-karkas-ibus-dvizhka (git log verified).
- All seven created files exist under docs/adr/ and test/e2e/ (checked via git diff --name-status 5781f7b..HEAD).
- Verification re-run green: `mise run ci` exit 0; VERDICTS-OK (15 ≥ 4); ADR sections CHECKED; ADR-001→journal, ADR-005 D-12/FocusIn/degradation-ladder counts non-zero; SPEC carries both delta texts with unique 4.4/4.5 headings.
