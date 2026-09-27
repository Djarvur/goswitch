---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
plan: 07
subsystem: input-correction
tags: [ibus, e2e, matrix-v2, selection, combo, hot-reload, macr, acceptance, tdd]

# Dependency graph
requires:
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 03
    provides: the per-surface spike table (anchor behavior, canonical ctrl+a, actual rungs), selection push parsing (lastSurrounding/waitNewSurrounding), startDaemonArgs/restartDaemonWithArgs
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 04
    provides: the canonical combo name (SHIFT_R+CTRL_R), the combo/mode log shapes, the reload-application record, the super-space D-34 interpretation
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 05
    provides: the MACR free letter x, the super-intercept record, the -config spawn ladder
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 06
    provides: goswitchctl + buildCtl/runCtl, the status config_valid vocabulary, the ctl-listening readiness record
provides:
  - test/e2e matrix v2 — 21-case acceptance matrix of the whole phase breadth (критерий №5 ROADMAP), green on the live desktop AND on the CI runner green106
  - matrixStep kinds select/combo/reload (strict schema, dispatch, summary arms) + matrixCase.expect_sel_step (the actual selection rung pin, pre-correction window)
  - matrixKeyNames replenished with the live-probed canonical spellings: ctrl+a, SHIFT_R+CTRL_R, super+x, home, shift+right, shift+end (KEY_-prefixed forms fall back to first letters — never use)
  - the reload step drives config application THROUGH goswitchctl (INST-02 in the matrix) and gates BOTH log forms + status validity
  - mise task e2e-matrix-v2; .github/workflows/e2e-matrix.yml matrix input defaulting to matrix-v2.yaml (v1 path preserved)
affects: [04-delivery, phase-verification]

actuals:
  tokens: 16293   # chars/4 over the realized diff (git diff e7d5255..HEAD = 65172 chars)
  tasks: 2
  commits: 3      # measured: git rev-list --count e7d5255..HEAD (the docs commit follows)
  plan_head_before: e7d52558d72e2bb246a1d42bf1adf8a59d055a96

tech-stack:
  added: []
  patterns:
    - "active-push selection gate: pure caret moves push surrounding text too (Home pushes cursor=0/anchor=0), so a selection step must wait for the NEWEST push to be ACTIVE — 'some new push arrived' races lagging caret pushes (live finding of the first v2 run)"
    - "pre-correction verification window: the correction itself makes clients push transient active states (cursor=7/anchor=0 mid-delete) — a selection-observability pin must only scan pushes BEFORE the first correction-family record"
    - "readback-marker name probing: to prove a ydotool chord name, press the candidate and type a marker over whatever state it left — the field content discriminates worked/failed/fallback-typed; surrounding pushes are unreliable as the probe oracle"

key-files:
  created:
    - test/e2e/cases/matrix-v2.yaml
    - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/red-evidence/03-07-task1-red.json
  modified:
    - test/e2e/matrix.go
    - test/e2e/matrix_test.go
    - test/e2e/main.go
    - test/e2e/case_ctl.go
    - test/e2e/case_select.go
    - test/e2e/case_combo.go
    - test/e2e/case_macr.go
    - test/e2e/case_m1.go
    - test/e2e/case_resilience.go
    - test/e2e/preflight.go
    - test/e2e/README.md
    - mise.toml
    - .github/workflows/e2e-matrix.yml

key-decisions:
  - "The matrix v2 breadth landed as 21 cases: 8 phrase rows (both directions, D-26 mixed, registers, trailing-space geometry, the 70-rune long-phrase A3 pin — the surrounding window holds the whole phrase on zenity, no truncation), 5 selection rows across all three surfaces and BOTH geometries, mixed text, combo, single tap, super-space, 3 reload rows, MACR — with the clipboard row EXCLUDED by a comment citing the 03-03 spike verdict (no primary-incompatible surface exists; the rung is unreachable and select-clipboard carries the mechanics)"
  - "The reload step drives the application THROUGH goswitchctl (forced reload + status), not the bare watcher: one step then covers CONF-02 and INST-02 together — the CLI exit/reply contract, BOTH daemon log forms (config reloaded / WARN config reload rejected) and the D-32 status visibility are asserted per step; the fragment applies by key-line replace-or-append (the appended unknown key IS the broken-edit mechanics of D-33)"
  - "The select step's gate waits for an ACTIVE new push — live finding: pure caret moves push too, and the lagging Home push raced the shift+end gate on the first run; expect_sel_step verifies over the PRE-correction window only (the correction itself pushes transient active states)"
  - "The 03-03 spike's zenity verdict is CORRECTED by the v2 live run: a SURVIVING ctrl+a selection on zenity makes the commit REPLACE the selected residue — the field settles at the converted word alone ('ghbdtn'); the spike saw transparency only because its failed probe candidates (ctrl+A, Control+a) had destroyed the selection before the double tap. The daemon-side degradation stands (no anchor push → the word path decides); the client-side actual is now pinned"
  - "GTE's surrounding-push map is operation-specific (live): typing, ctrl+a and shift+right push; Home and shift+end NEVER do. The reverse (RTL) geometry is therefore proven on GTE by six gated shift+right extensions from the home caret (cursor=6/anchor=0 — the same shape chromium's native ctrl+a gives), and select-reverse via shift+end is not expressible as a gated step on GTE"
  - "The overview episode's FocusOut hard-resets the daemon buffer (first-run live finding) — the super-space row retypes the word over a select-all AFTER the focus recovery, mirroring the 03-04 case's type-after-episode order"
  - "workflow_dispatch uses the workflow definition FROM THE DISPATCHED REF (live-proven: the run's step name is the new 'e2e matrix (test/e2e/cases/matrix-v2.yaml)' form) — the first green green106 run on matrix v2 is https://github.com/Djarvur/goswitch/actions/runs/34984814995 (21/21 PASS in the run log)"

requirements-completed: [CORR-02, CORR-03, CORR-06, SWCH-01, SWCH-02, SWCH-03, CONF-02, MACR-01, INST-02]

coverage:
  - id: D1
    description: "The three phase-3 step kinds in the strict schema + dispatcher: select/combo/reload decode with per-kind vocabularies, the one-kind counter counts all seven, unknown fields/names/expectations are decode errors; fragment application replaces by key, appends unknown keys, rejects malformed lines; both reload log forms distinguished"
    verification:
      - kind: unit
        ref: test/e2e/matrix_test.go#TestMatrixDecode_SelectComboReload
        status: pass
      - kind: unit
        ref: test/e2e/matrix_test.go#TestMatrixDecode_RejectsNew
        status: pass
      - kind: unit
        ref: test/e2e/matrix_test.go#TestMatrixKeyNames_Canonical
        status: pass
      - kind: unit
        ref: test/e2e/matrix_test.go#TestMatrixStep_RunDispatch
        status: pass
      - kind: command
        ref: "mise exec -- go test ./test/e2e -run TestMatrixDecode -race -count=1 -v | grep -c '^--- PASS' → 6 (was 4)"
        status: pass
    human_judgment: false
  - id: D2
    description: "matrix-v2.yaml covers the full phase breadth on the live desktop: 21/21 PASS, exit 0, per-case fresh daemon, desktop state restored (teardown contract)"
    verification:
      - kind: command
        ref: "mise run e2e-matrix-v2 → matrix: 21/21 PASS (2026-09-15T17:46 report; first run 18/21 drove three live findings, all fixed and re-proven)"
        status: pass
    human_judgment: false
  - id: D3
    description: "The v1 regression stays green and the CI gate holds at the final tree"
    verification:
      - kind: command
        ref: "mise run e2e-matrix → 16/16 PASS; mise run ci → exit 0, 0 lint issues"
        status: pass
    human_judgment: false
  - id: D4
    description: "The runner gate: e2e-matrix.yml defaults to matrix-v2.yaml and the first green matrix-v2 run on green106 is recorded"
    verification:
      - kind: command
        ref: "gh run view 34984814995 → conclusion success, step 'e2e matrix (test/e2e/cases/matrix-v2.yaml)', run log 'matrix: 21/21 PASS' (https://github.com/Djarvur/goswitch/actions/runs/34984814995)"
        status: pass
    human_judgment: false
  - id: D5
    description: "The zenity spike-table correction (a surviving ctrl+a selection makes the commit replace the selected residue — field settles at the converted word alone) reverses a documented 03-03 verdict; the owner confirms the corrected reading at the verify gate"
    verification:
      - kind: e2e
        ref: test/e2e/cases/matrix-v2.yaml#select-all-zenity (PASS — readback "ghbdtn", no pre-correction active push)
        status: pass
    human_judgment: true
    rationale: "The pin-the-actual discipline produced a documented behavior change against the 03-03 spike table (the spike's probe methodology had destroyed the selection before its own double tap); the automated oracle passes, but the corrected verdict is exactly the kind of live-truth reversal the verify gate exists to surface."

duration: 48 min
completed: 2026-09-15
status: complete
---

# Phase 3 Plan 7: Матрица v2 полной широты Summary

**Приёмочная матрица фазы из 21 кейса зелёная на живом столе (21/21) и на CI-раннере green106 (run 34984814995): фразы до 70 рун, выделение в обеих геометриях на трёх поверхностях, комбо, одиночный тап, super-space, reload через goswitchctl (обе лог-формы + last-good), MACR — со строгой схемой шагов select/combo/reload и канонизацией имён по живым пробам; по ходу прогона живьём исправлена вердикт-строка zenity из спайк-таблицы 03-03.**

## Performance

- **Duration:** 48 min (14:10–14:59 UTC)
- **Tasks:** 2 (1 tdd: RED→GREEN with RED_EVIDENCE_OK, 1 auto)
- **Files modified:** 15 (13 code/config/docs + 1 case matrix + 1 red-evidence record)
- **Commits:** 3 production (measured from plan_head_before e7d5255)

## Accomplishments

- **Step schema (Task 1, TDD):** `matrixStep` grew `Select`/`Combo`/*matrixReload` with the one-kind counter counting all seven; per-kind closed vocabularies (a select/combo name must BOTH resolve through `matrixKeyNames` and belong to its chord set); `reload` carries a fragment + an applied/rejected expectation. `matrixStepKind` is the dispatch's single classification source; `stepSummary` names the new kinds in FAIL lines.
- **Canonical name table (Pitfall 5):** `matrixKeyNames` replenished with the spike/probe-pinned spellings — `ctrl+a` (03-03), `SHIFT_R+CTRL_R` (03-04), `super+x` (03-05), plus the 03-07 readback-probed `home`/`shift+right`/`shift+end`. Live probe verdicts: lowercase forms work; `KEY_HOME`-style prefixed forms fall back to first letters ("abcdefkX") and must never enter a case.
- **Step runners:** the select step presses the canonical name and — on anchor-reporting surfaces — gates on an ACTIVE new push; the combo step gates on the `"combo"` record + the mode flip (the action/mode shapes untouched); the reload step establishes the case's `-config` daemon (complete base document, restart, readiness ladder, engine re-activation, focus settle), applies fragments by key-line replace/append, then drives the application THROUGH goswitchctl and asserts the CLI exit/reply, the daemon's log form (both variants) and the status validity.
- **matrix-v2.yaml (Task 2):** 21 cases — 8 phrase rows (en↔ru, D-26 mixed, registers, trailing-space D-13 geometry, the 70-rune long-phrase A3 pin: the surrounding window holds the whole phrase on zenity, the correction converts all 70 runes), 5 selection rows (zenity degraded pin, GTE LTR, chromium native RTL commit-replaces, GTE partial [0,6), GTE reverse geometry built from six gated shift+right extensions — cursor=6/anchor=0), word-mixed, combo-word-layout, layout-single (both directions), super-space-alive, reload-window / reload-invalid-last-good / ctl-status-reload, macr-super-letter. The clipboard row is EXCLUDED with the 03-03 spike-verdict comment (no primary-incompatible surface — the rung is unreachable on this desktop).
- **Live gates:** `mise run e2e-matrix-v2` 21/21 PASS on the owner's desktop (teardown contract held — no orphaned names/processes; preflight-verified between runs); `mise run e2e-matrix` 16/16 PASS (v1 regression); `mise run ci` green with 0 lint issues at every iteration boundary.
- **CI runner:** `.github/workflows/e2e-matrix.yml` gained a `matrix` path input defaulting to `test/e2e/cases/matrix-v2.yaml` (v1 selectable for the regression dispatch); dispatched on the phase branch — the run proves the REF's workflow definition executes (the new step name shows in the run) and finished green: https://github.com/Djarvur/goswitch/actions/runs/34984814995 (21/21 PASS in the run log).

## Task Commits

Each task committed atomically (TDD: RED before GREEN):

1. **Task 1: Шаги select/combo/reload + канонизация имён** — `eb42be9` (test, RED) + `63cff95` (feat, GREEN)
2. **Task 2: matrix-v2.yaml + зелёный прогон на столе и раннере** — `21df37c` (feat)

**Plan metadata:** this commit (docs: complete plan)

## TDD Gate Compliance

The tdd task followed RED → GREEN with machine-validated red evidence:

| Task | RED commit | Evidence record | Verdict | GREEN commit |
|------|-----------|-----------------|---------|--------------|
| 1 | `eb42be9` test(03-07) | red-evidence/03-07-task1-red.json | RED_EVIDENCE_OK (3 new positive tests failing on assertions against compile stubs — SelectComboReload/KeyNames_Canonical/RunDispatch; 7 green, RejectsNew green-by-design as the stub rejects everything, the 03-03 continuity-pin precedent) | `63cff95` feat(03-07) |

No REFACTOR commit — the GREEN implementation landed lint-clean under the strict v2 config (all findings fixed inside the iterations per D-08).

## Decisions Made

- The six key-decisions of the frontmatter, in brief: the 21-case breadth with the documented clipboard exclusion; the reload step as the INST-02+CONF-02 vehicle (goswitchctl-driven, both log forms, status validity, key-line fragment mechanics); the active-push selection gate + the pre-correction verification window; the corrected zenity verdict (surviving selection eats the commit — the spike's transparency was probe-contaminated); GTE's operation-specific push map (Home/shift+end never push — reverse geometry proven via gated shift+right extensions); the buffer-killing overview episode (type after it, retype over a select-all).
- The `super+x` key step gates on the interception record by name class (documented in the vocabulary): every matrix chord case arms MACR through a reload step first, so the gate is the proof, not an assumption.
- The first v2 run (18/21) was the plan's own hypothesis test working as designed: each FAIL classified as either a case-authoring error (super-space ordering, gate race) or a live-truth correction (zenity selection semantics) — none were daemon defects; the daemon's behavior was correct in every failing signature.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Live truth] The 03-03 spike's zenity verdict was probe-contaminated**
- **Found during:** Task 2 (the first full matrix run)
- **Issue:** The spike table's zenity row ("selection transparent to the delete+commit") was an artifact: the spike's failed select-all candidates (ctrl+A, Control+a) had typed stray characters and destroyed the selection before its double tap. With ONLY the canonical ctrl+a pressed (the matrix case), the silently-selected range makes the subsequent CommitText REPLACE the selected residue — the field settles at "ghbdtn" (the converted word alone), not "ghbdtn ghbdtn".
- **Fix:** The select-all-zenity row pins the corrected actual: expect_text "ghbdtn", expect_sel_step degraded with the pre-correction-window check (the daemon-side degradation stands — no anchor push arrives, the word path decides; only the client-side outcome changed). Documented in the YAML comment and coverage D5 routes it to the owner at the verify gate.
- **Files modified:** test/e2e/cases/matrix-v2.yaml, test/e2e/matrix.go (preCorrectionActiveSelection)
- **Verification:** the row PASSes; the repro log showed the exact push sequence (post-correction cursor=7/anchor=0 → cursor=6/anchor=6)
- **Committed in:** `21df37c`

**2. [Rule 1 - Live truth] GTE pushes nothing for Home/shift+end — the plan-literal reverse row was inexpressible**
- **Found during:** Task 2 (the first full matrix run)
- **Issue:** The plan's select-reverse shape (Home + shift+end) forms the RTL selection client-side, but GTE emits NO surrounding push for either chord — the daemon never learns the selection, so no gated reverse row (and no selection conversion) is possible that way.
- **Fix:** The reverse geometry is proven with the push-capable chord instead: Home, then six shift+right extensions (the last one a gated select step) — cursor=6/anchor=0, the exact RTL shape chromium's native ctrl+a gives — and the tap converts [0,6). The push map itself (typing/ctrl+a/shift+right push; Home/shift+end never) is documented in the README and here.
- **Files modified:** test/e2e/cases/matrix-v2.yaml
- **Verification:** select-reverse-gte PASS (the gated extension's push observed cursor=6/anchor=0)
- **Committed in:** `21df37c`

**3. [Rule 3 - Case authoring] super-space-alive: the overview FocusOut kills the buffer**
- **Found during:** Task 2 (the first full matrix run)
- **Issue:** With the word typed BEFORE the Super/Escape episode, the taps fired (`action n:2`) but the correction refused with `empty-buffer` — the overview's FocusOut had hard-reset the buffer; and with typing moved after the episode, the focus step's char-count gate failed on the empty field (it expects the expect_text rune count).
- **Fix:** The row types the word before (feeding the focus gate), then — after the focus recovery — selects all and retypes the word over the selection: the field is refreshed AND the live buffer re-fed; the double tap corrects.
- **Files modified:** test/e2e/cases/matrix-v2.yaml
- **Verification:** super-space-alive PASS (repro-verified before the full re-run)
- **Committed in:** `21df37c`

**4. [Rule 3 - Blocking] Strict-lint findings inside the iterations (D-08)**
- **Found during:** Tasks 1–2 (mise run ci)
- **Issue:** goconst (11: the cross-file "component registered"/"SHIFT_R+CTRL_R"/"ctrl+a" literals the new code joined; "key"/chord strings), funlen/gocyclo/cyclop (validate + the RunDispatch corpus), gocritic (ifElseChain), gofmt, gosec G703 (the stand's own temp path flagged as traversal).
- **Fix:** componentRegisteredMark/comboCanonicalName/selectExtendRight/selectExtendEnd constants adopted across the package (eight files, mechanical); validateKind/validateChordName/validateKeyName split; assertStepKindSummary extracted; driveCaseReload split into assertReloadReply/assertReloadStatus; kind constants; gofmt; #nosec G703 with the stand-internal-path argument.
- **Verification:** mise run ci green after each fix (0 lint issues)
- **Committed in:** `63cff95`, `21df37c`

---

**Total deviations:** 4 auto-fixed (2 live-truth pins, 1 case-authoring fix, 1 blocking lint round)
**Impact on plan:** All resolved toward the locked must-haves; no scope creep. Deviations 1–2 are exactly the "фактические ступени/имена — запинены, не предполагаемы" discipline the plan's success criteria demand; the zenity verdict correction (coverage D5) is the one item the owner should eyeball at the verify gate.

## Issues Encountered

Environmental, all recovered without product impact: (a) the first scratch name-probe run wedged the AT-SPI bridge mid-run (the known 02-06/03-04 trap — serial surface churn); healed by the documented a11y-bus restart; the probe was rewritten readback-based (marker-over-state), which also fixed its oracle reliability; (b) one transient fresh-GTE focus flake on a re-run (registry reap race — a retry sufficed); (c) one orphaned stand daemon after the wedged probe run — reaped before the next preflight.

## Known Limitations (for the verify gate)

- **The zenity selection actual (corrected):** a silently-surviving ctrl+a selection on zenity makes the daemon's CommitText replace the selected residue — the field settles at the converted word alone. The daemon-side degradation (D-30: no anchor → the word path) stands; the matrix pins the client-side outcome. Owner confirmation routed via coverage D5.
- **GTE's push map is operation-specific:** Home/shift+end selections are invisible to the daemon on GTE (no push ever); such selections degrade to the word path exactly like zenity's ctrl+a. Reverse-geometry coverage rides on shift+right extensions (GTE) and the native inverted ctrl+a (chromium).
- **The select-partial row's ungated extensions:** the five ungated shift+right key steps are verified only by the final field outcome (a wrong range would produce a different text) — per-extension pushes are observed but not asserted individually.
- **The GTK4-Wayland forwarded-events limitation** (03-05) carries into the macr-super-letter row unchanged: the interception is proven by the consume + the INFO record; the forwarded Ctrl+x burst still does not move GTK4 widgets.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 3 is COMPLETE (7/7 plans): критерий №5 holds — matrix v2 green across the live desktop and green106 — and the phase is ready for `/gsd:verify-work`.
- The matrix v2 schema (select/combo/reload + expect_sel_step) is the acceptance base any future phase extends; the canonical name table grows only through live probes (the readback-marker pattern documented in tech-stack).
- The runner workflow now takes a matrix path input — future matrix versions only add a mise task + a default flip.

## Self-Check: PASSED

All key-files exist on disk; all three production commits found in history (eb42be9, 63cff95, 21df37c); the plan ledger measured 3 commits from plan_head_before e7d5255 (matches `commits:` in frontmatter); `mise run ci` green (exit 0, 0 lint issues) at the final tree; e2e-matrix-v2 21/21 PASS (live desktop) and green106 run 34984814995 conclusion success with 21/21 in its log; e2e-matrix 16/16 PASS.
