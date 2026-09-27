---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
plan: 01
subsystem: input-correction
tags: [ibus, layout-conversion, phrase-correction, run-segmentation, tdd, e2e]

# Dependency graph
requires:
  - phase: 02-korrektsiya-slova-en-ru
    provides: word-correction pipeline (Detect/Convert/BuildPlan/ADR-004 verify), script-true buffer, e2e matrix v1
provides:
  - Buffer.Phrase()/ReplacePhrase() — whole-buffer correction range (D-25)
  - correctionRange-parameterized actor pipeline — one code path for word (Double) and phrase (Triple) ranges (D-23)
  - correct.ConvertRuns — run-wise mixed-text conversion with last-letter anchor (D-22/D-23)
  - BuildPlan(backspaceCap) + DefaultBackspaceCap=50 + LevelNone + actor reason backspace-cap (D-27)
  - live e2e cases phrase-en-ru, phrase-mixed; updated word-mixed oracle (D-16→D-23 succession)
affects: [03-03-selection, 03-04-config, 03-07-matrix-v2]

actuals:
  tokens: 19309   # chars/4 over the realized diff (git diff 8d0ddfb..HEAD)
  tasks: 2
  commits: 4      # measured: git rev-list --count 8d0ddfb..HEAD
  plan_head_before: 8d0ddfba86a111c259123430b1c079e0d6390e65

tech-stack:
  added: []
  patterns:
    - "correctionRange{token, tail, replace} — the actor's pipeline parameterized by range; the selection path (03-03) joins the same seam"
    - "script-classified run scan (classifyRune) generalizing Detect's per-rune test; neutrals build no segment and shift no anchor"

key-files:
  created:
    - internal/correct/runs.go
    - internal/correct/runs_test.go
    - test/e2e/case_phrase.go
  modified:
    - internal/correct/buffer.go
    - internal/correct/buffer_test.go
    - internal/correct/plan.go
    - internal/correct/plan_test.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - test/e2e/case_word.go
    - test/e2e/cases/matrix-v1.yaml
    - test/e2e/main.go
    - test/e2e/README.md
    - mise.toml

key-decisions:
  - "Mixed-range anchor = script of the LAST letter; digits/punctuation/other-script letters are neutral (no segment, no anchor shift); each foreign run converts independently through Convert reused as-is"
  - "Homogeneous ranges convert wholesale by composition (уточнение D-22, 2026-09-15) — matrix v1 and the Phase-2 corpus stay green untouched"
  - "BuildPlan refuses level 2 via LevelNone when token+tail exceeds the Backspace cap; level 1 (one DeleteSurroundingText call) is never capped — the ceiling bounds the destructive key series only"
  - "phrase-mixed injection follows the 02-04 live-proven path: physical Latin keys after the flip, the engine commits the Cyrillic runes (the plan's parenthetical «RU-коммиты движка» names exactly this)"

patterns-established:
  - "Run-scan pattern: classifyRune/scriptKind in the pure correct package — per-rune script classes with neutral default, anchor from the last classified letter"
  - "TDD red-evidence records transcribed to node-test TAP (the checker's discovery format) — go test -v output converted mechanically, content faithful"

requirements-completed: [CORR-02, CORR-06]

coverage:
  - id: D1
    description: "Triple Right Shift corrects the whole phrase from the buffer live: ghbdtn ghbdtn → привет привет in zenity through FSM Triple → range-phrase → ADR-004 verify → ADR-003 ladder (CORR-02)"
    requirement: CORR-02
    verification:
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_TripleTapCorrectsPhrase
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_TripleTapEmptyBuffer
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_PhraseVerifyMismatch
        status: pass
      - kind: e2e
        ref: mise run e2e-phrase (PASS phrase-en-ru, stdout oracle "привет привет")
        status: pass
    human_judgment: false
  - id: D2
    description: "Mixed text converts by runs (D-22/D-23): anchor = last letter, foreign runs convert independently, own runs untouched — golden corpus verbatim from CONTEXT plus live word/phrase cases (CORR-06)"
    requirement: CORR-06
    verification:
      - kind: unit
        ref: internal/correct/runs_test.go#TestConvertRuns_GoldenCorpus
        status: pass
      - kind: unit
        ref: internal/correct/runs_test.go#TestConvertRuns_HomogeneousWholesale
        status: pass
      - kind: unit
        ref: internal/correct/runs_test.go#TestConvertRuns_Refusals
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_MixedWordConvertsForeignRuns
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_PhraseMixedCorrects
        status: pass
      - kind: e2e
        ref: mise run e2e-phrase-mixed && mise run e2e-word-mixed (both PASS)
        status: pass
    human_judgment: false
  - id: D3
    description: "Backspace series cap D-27: BuildPlan refuses level 2 via LevelNone when token+tail > cap; DefaultBackspaceCap = 50; level 1 ignores the cap; actor refuses with reason backspace-cap before any ForwardKeyEvent"
    verification:
      - kind: unit
        ref: internal/correct/plan_test.go#TestBuildPlan_BackspaceCap
        status: pass
    human_judgment: false
  - id: D4
    description: "D-24 semantics: «нечего конвертировать» is a SUCCESS without changes (outcome done, no WARN, no replacement calls) — implemented as the changed=false branch of ConvertRuns/actor, reserved for externally anchored ranges (selection 03-03); under the уточнение D-22 no word/phrase input reaches it"
    verification: []
    human_judgment: true
    rationale: "The plan's literal D-24 example (привет, якорь RU → out==input) contradicts the plan's own must-have truth #3, matrix v1 word-ru-en and Task 1's tracer pin (all require привет→ghbdtn wholesale). Resolved toward the must-haves; the owner confirms the corpus reading at the verify gate (the plan's FLAGGED assumption names exactly this confirmation)."
  - id: D5
    description: "Succession D-16 → D-23: the mixed word is no longer refused — word-mixed oracle changes from untouched gfbпривет to run-converted паипривет (unit corpus, live case, matrix v1 row)"
    verification:
      - kind: e2e
        ref: mise run e2e-matrix (16/16 PASS with the renamed word-mixed row expecting паипривет)
        status: pass
    human_judgment: true
    rationale: "Planned semantics change of an owner-accepted Phase-2 refusal (D-16); the plan requires the succession to be visible to the owner at the verify gate, so it routes to the human even though the automated oracle passed."

duration: 26 min
completed: 2026-09-15
status: complete
---

# Phase 3 Plan 1: Трассер фразы + run-конвейер Summary

**Triple Right Shift corrects a whole typed phrase live (ghbdtn ghbdtn → привет привет), mixed text converts by script runs under a last-letter anchor, and the Backspace ladder gains its D-27 cap — word and phrase now share one range-parameterized pipeline.**

## Performance

- **Duration:** 26 min
- **Started:** 2026-09-15T08:29:32Z
- **Completed:** 2026-09-15T08:56:00Z
- **Tasks:** 2 (both tdd: RED→GREEN, red-evidence RED_EVIDENCE_OK ×2)
- **Files modified:** 14 (13 planned + test/e2e/README.md)

## Accomplishments

- Triple-tap phrase correction end to end: `hotkey.Triple` now drives the same two-phase pipeline as the word over the WHOLE buffer range (`Buffer.Phrase()`/`ReplacePhrase()`, D-25); live zenity case proves `ghbdtn ghbdtn` → `привет привет` with the ADR-004 verify and the (-13,13) ladder geometry.
- Run-wise mixed conversion (D-22/D-23): `correct.ConvertRuns` segments a range into maximal same-script letter runs; the anchor is the script of the last letter (neutrals shift nothing); foreign runs convert independently through `Convert` reused as-is; homogeneous ranges still convert wholesale (уточнение D-22) so matrix v1 and the whole Phase-2 corpus stay green without a single expectation edit.
- Backspace cap (D-27): `BuildPlan` takes `backspaceCap` (`correct.DefaultBackspaceCap = 50`); over-cap ranges return `LevelNone` and the actor refuses with reason `backspace-cap` before emitting one ForwardKeyEvent; level 1 (single DeleteSurroundingText) is never capped.
- Planned succession D-16 → D-23: the mixed word converts its foreign run instead of being refused — `gfbпривет` → `паипривет` in the unit corpus, the live word-mixed case and matrix v1 (row renamed `word-mixed`).
- Single conversion code path (D-23): the actor calls `ConvertRuns` for both word and phrase ranges; `Detect` no longer appears in actor.go (grep-pinned by acceptance); the selection path (03-03) joins the same `correctionRange` seam.

## Task Commits

Each task was committed atomically (TDD: RED before GREEN):

1. **Task 1: Трассер — тройной Right Shift исправляет фразу** — `7e3b7f8` (test, RED) + `ffbc6e4` (feat, GREEN)
2. **Task 2: Run-сегментация D-22/D-23/D-24 + кап D-27 + сукцессия D-16→D-23** — `540fc47` (test, RED) + `b92987c` (feat, GREEN)

**Plan metadata:** this commit (docs: complete plan)

## Files Created/Modified

- `internal/correct/runs.go` (new) — `ConvertRuns` + run scanner (`classifyRune`, `convertForeignRuns`)
- `internal/correct/runs_test.go` (new) — golden corpus (4 CONTEXT specifics with D-22/D-23 comments), wholesale family, refusal family
- `internal/correct/buffer.go` — `Phrase()` (slices.Clone idiom) + `ReplacePhrase()` (whole-range replace + recompute)
- `internal/correct/buffer_test.go` — phrase API pins incl. the recompute invariant
- `internal/correct/plan.go` — `LevelNone`, `DefaultBackspaceCap`, `BuildPlan(..., backspaceCap)` with the D-27 refusal
- `internal/correct/plan_test.go` — cap corpus (over/at cap, token+tail, level-1 exempt)
- `internal/session/actor.go` — `correctionRange` parameterization, Triple dispatch, `ConvertRuns` conversion step, `refusalReason` (no-letters | convert-failed), D-24 branch, backspace-cap refusal
- `internal/session/actor_test.go` — triple-tap corpus, mixed word/phrase conversion pins, succession replacements
- `test/e2e/case_phrase.go` (new) — `runPhraseENRU`, `runPhraseMixed` + `tripleTapCorrects` helper
- `test/e2e/case_word.go` — word-mixed oracle → `паипривет`, waits for correction-done
- `test/e2e/cases/matrix-v1.yaml` — `word-mixed` row expects `паипривет`
- `test/e2e/main.go` — registry: phrase-en-ru, phrase-mixed
- `test/e2e/README.md` — case list and the word-mixed description updated for the succession
- `mise.toml` — tasks `e2e-phrase`, `e2e-phrase-mixed` (live GNOME session; NOT in ci)

## TDD Gate Compliance

Both tasks followed RED → GREEN with machine-validated red evidence:

| Task | RED commit | Evidence record | Verdict | GREEN commit |
|------|-----------|-----------------|---------|--------------|
| 1 | `7e3b7f8` test(03-01) | red-evidence/03-01-task1-red.json | RED_EVIDENCE_OK (4 failing on assertions, 46 pre-existing green) | `ffbc6e4` feat(03-01) |
| 2 | `540fc47` test(03-01) | red-evidence/03-01-task2-red.json | RED_EVIDENCE_OK (16 failing on assertions, 110 pre-existing green) | `b92987c` feat(03-01) |

No REFACTOR commit — the GREEN implementations landed lint-clean under the strict v2 config (cyclop/lll/exhaustive/paramTypeCombine fixed inside the GREEN iterations).

## Decisions Made

- Mixed-anchor conversion scans per-rune script classes with a neutral default; the wholesale rule keys on script presence (both scripts → anchor rule; one script → Phase-2 wholesale). This keeps `Detect`'s Latin/Cyrillic test as the single classification idiom (factored into `classifyRune`).
- The phrase-mixed live case types the RU part through physical Latin keys after the flip (the 02-04 live-proven path — the engine commits the Cyrillic runes), matching the plan's «RU-коммиты движка» parenthetical rather than its literal `injectText("привет")` (ydotool cannot type Cyrillic through the EN XKB group).
- `test/e2e/README.md` joined the change set (not in the plan's file list) so the renamed word-mixed case and the new phrase cases stay documented.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Plan inconsistency] D-24 nothing-to-convert example is unsatisfiable as written**
- **Found during:** Task 2 (RED authoring)
- **Issue:** The plan's `TestConvertRuns_NothingToConvert` example («привет», якорь RU → out==input, changed=false) presumes the strict anchor applying to homogeneous ranges. That directly contradicts the plan's own must-have truth #3 (`привет`→`ghbdtn` wholesale), `TestConvertRuns_HomogeneousWholesale`, Task 1's already-committed tracer pin (`TestActor_TripleTapCorrectsPhrase`) and matrix v1 `word-ru-en`. Proof of unreachability: matrix v1 forces homogeneous wholesale; a mixed range always holds ≥1 foreign run; therefore no word/phrase input can produce changed=false — the уточнение D-22 itself says the strict anchor «превращает базовый кейс ghbdtn→привет в no-op» and rejects it.
- **Fix:** Implemented the must-have semantics (mixed → anchor rule; homogeneous → wholesale). The D-24 surface exists in code: `ConvertRuns` computes `changed` honestly and the actor's `!changed` branch logs `correction outcome:done` with zero sink calls and no WARN — reserved for externally anchored ranges (the selection path of 03-03). The literal no-op test was not written; the resolution is pinned as a doc note on `TestConvertRuns_HomogeneousWholesale` and routes to the owner at the verify gate (coverage D4), exactly the confirmation the plan's FLAGGED assumption schedules.
- **Files modified:** internal/correct/runs.go, internal/correct/runs_test.go, internal/session/actor.go
- **Verification:** golden corpus + wholesale corpus green under -race; `mise run ci` green
- **Committed in:** `540fc47` / `b92987c`

**2. [Rule 3 - Blocking] Red-evidence checker requires node-test TAP format**
- **Found during:** Task 1 (RED evidence recording)
- **Issue:** `gsd check tdd-red-evidence` discovers tests from node-test TAP lines (`ok N - name` / `# tests N`); raw `go test -v` output yields `zero_tests_discovered` → INVALID_RED.
- **Fix:** The real `go test -v` output is mechanically transcribed to TAP in the record (per-test result lines + counts derived from the actual run, assertion excerpts preserved). Both records validate RED_EVIDENCE_OK.
- **Files modified:** .planning/phases/03-.../red-evidence/03-01-task{1,2}-red.json
- **Verification:** `gsd check tdd-red-evidence` → RED_EVIDENCE_OK for both records
- **Committed in:** `7e3b7f8`, `540fc47`

**3. [Rule 3 - Blocking] Strict-lint findings inside GREEN iterations**
- **Found during:** Task 1 and Task 2 (mise run ci)
- **Issue:** lll (line length), cyclop (runPhraseMixed 16>15), exhaustive (missing scriptNeutral arm), paramTypeCombine (combined bool results).
- **Fix:** Line wraps; extracted `tripleTapCorrects` helper; added the neutral arm; combined `changed, ok bool`. Fixed inside the same iteration per D-08 (no «поправим линт потом»).
- **Verification:** `mise run ci` green after each fix
- **Committed in:** `ffbc6e4`, `b92987c`

---

**Total deviations:** 3 auto-fixed (1 plan-inconsistency resolution, 2 blocking tooling/lint)
**Impact on plan:** All resolved toward the plan's locked must-haves; no scope creep. Deviation 1 needs the owner's corpus confirmation at the verify gate (by design of the FLAGGED assumption).

## Issues Encountered

None beyond the deviations above. Live gates all green on the target desktop: `e2e-phrase`, `e2e-phrase-mixed`, `e2e-word-mixed` PASS; `e2e-matrix` 16/16 PASS (fresh daemon per case).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The `correctionRange` seam is the cut-in point for selection (03-03) and the combo (D-36); the conversion step is uniformly `ConvertRuns`.
- The Backspace cap is a constant (`correct.DefaultBackspaceCap`) awaiting YAML wiring in 03-04.
- The D-24 changed=false surface awaits its first reachable caller (selection with an external anchor, 03-03) — coverage D4 tracks the owner's confirmation.
- Assumption A3 (surrounding window vs long phrases) stays scheduled for the matrix v2 live probe (03-07), per the plan's FLAGGED note.

## Self-Check: PASSED

All key-files exist on disk; all four task commits found in history; plan ledger measured 4 commits from plan_head_before 8d0ddfb (matches `commits:` in frontmatter).
