---
phase: 02-korrektsiya-slova-en-ru
plan: "01"
subsystem: correct
tags: [go, runes, layouts, tdd, layout-correction, punto, pure-package]

requires:
  - phase: 01-adr-paket-i-karkas-ibus-dvizhka
    provides: layouts.ENToRU/RUToEN golden-pinned key-position tables (CORR-08) and the pure-package pattern (internal/hotkey)
provides:
  - internal/correct — headless word-correction library: phrase Buffer with token boundaries, Detect, Convert, MatchesSuffix, BuildPlan (both ladder levels)
  - Buffer API — NewBuffer/Push/Backspace/HardReset/Token/Tail/ReplaceToken; letter-capable key union (D-14) + digits (D-15); last word survives a separator (D-13); ReplaceToken toggle invariant
  - Detect — script classification with the single-foreign-letter mixed refusal (D-16), neutrals never mix (CORR-04)
  - Convert — per-rune key-position conversion, register preserved constructively, non-letter token runes identical (CORR-05)
  - MatchesSuffix — the ADR-004 mandatory suffix verification (text before cursor must END with the token)
  - BuildPlan — token+tail replacement geometry of BOTH ladder levels, all counts in runes (CORR-07, ADR-003); Level-2 formulas ready for plan 02-05 without rework
  - TDD corpus of five *_test.go pinning D-13..D-16, both shift levels, both directions, suffix rule, tail arithmetic
affects: [02-03 tracer wiring, 02-04 flip mode, 02-05 ladder level 2, 02-06 e2e matrix]

actuals:
  tokens: 15620   # chars/4 over the realized diff 148c28b..HEAD (estimate was 20000)
  tasks: 2
  commits: 4      # MEASURED: git rev-list --count 148c28b..HEAD (ledger base)

tech-stack:
  added: []
  patterns:
    - token = maximal run of letter-capable keys (union of EN+RU layers by position) ∪ digits — letterCapable derived from the generated tables, no hand-written key list
    - buffer coordinates rederived from contents after destructive mutations (recompute) — Push stays O(1), pop/replace stay honest
    - both ladder levels delete token+tail and recommit converted+tail — deleting exactly the token would strand the cursor after the tail (Pitfall 1 geometry)
    - capability bit mirrored as a constant in the pure package (CapSurroundingText = 1<<5) — pure packages never import the D-Bus-bound engine (Phase 1 dependency direction)

key-files:
  created:
    - internal/correct/buffer.go
    - internal/correct/direction.go
    - internal/correct/convert.go
    - internal/correct/verify.go
    - internal/correct/plan.go
    - internal/correct/buffer_test.go
    - internal/correct/direction_test.go
    - internal/correct/convert_test.go
    - internal/correct/verify_test.go
    - internal/correct/plan_test.go
    - .planning/phases/02-korrektsiya-slova-en-ru/red-evidence/02-01-task1-red.json
    - .planning/phases/02-korrektsiya-slova-en-ru/red-evidence/02-01-task2-red.json
  modified: []

key-decisions:
  - "Both ladder levels replace token+tail and recommit converted+tail: deleting exactly the token at a non-empty tail puts the cursor AFTER the tail and the commit lands behind it (D-13 pin `ghbdtn ` → `привет ` allows only token+tail) — planner's geometric correction over RESEARCH Pattern 1, implemented as pinned"
  - "CapSurroundingText (1<<5) mirrored from engine/keys.go instead of importing engine — the package purity contract (no godbus transitively) follows the Phase 1 precedent (KeyvalShiftR lives in internal/hotkey); plan key_link anticipated this («константа без D-Bus-импорта самого транспорта»)"
  - "Backspace keeps the separator in Tail(): buffer "ghbdtn v" + Backspace → Token "ghbdtn", Tail " " — the buffer must keep mirroring the field, and the token+tail geometry depends on that separator (Pitfall 1); see Deviations"
  - "Corpus literals named as constants (wordEN/wordRU/...), house style of fsm_test.go timing constants — goconst of the max-strict lint demands it (D-10)"
  - "Targeted `#nosec G115` annotations with written justification on the int→int32/uint32 conversions of plan.go — first such conversions in the codebase; a phrase since the last hard reset is bounded by human typing, far below 2^31 runes"

patterns-established:
  - "Red-evidence records persisted per task under red-evidence/ in the phase dir — TAP transcription of go test -v validated by `check tdd-red-evidence`"
  - "Named golden-corpus constants shared across a test package (correct_test) satisfy goconst without weakening the corpus"

requirements-completed: [CORR-04, CORR-05, CORR-07]

coverage:
  - id: D1
    description: "Phrase buffer with token boundaries: D-13 last-word-after-space, D-14 punctuation membership, D-15 digits, backspace pop, ReplaceToken toggle"
    requirement: CORR-04
    verification:
      - kind: unit
        ref: internal/correct/buffer_test.go#TestBuffer_TokenRules|TwoTokens|WordAfterSpace|PunctuationToken|DigitsInToken|BackspacePop|ReplaceTokenToggle
        status: pass
    human_judgment: false
  - id: D2
    description: "Direction detector: script classification, single foreign letter refuses the word, neutrals never mix, both shift levels"
    requirement: CORR-04
    verification:
      - kind: unit
        ref: internal/correct/direction_test.go#TestDetect_PureScript|MixedRefused|NeutralNotMixed|BothShiftLevels
        status: pass
    human_judgment: false
  - id: D3
    description: "Layout conversion preserving register per character in both directions, digits and token punctuation identical, unmapped letter fails"
    requirement: CORR-05
    verification:
      - kind: unit
        ref: internal/correct/convert_test.go#TestConvert_CasePreserved
        status: pass
    human_judgment: false
  - id: D4
    description: "Suffix verification of ADR-004: text before cursor must end with the token; tail-after-token in the text and shorter text are mismatches"
    verification:
      - kind: unit
        ref: internal/correct/verify_test.go#TestMatchesSuffix
        status: pass
    human_judgment: false
  - id: D5
    description: "Ladder arithmetic of both levels in runes: token+tail geometry, empty-tail degenerate case, Cyrillic 6-runes-not-12-bytes pin"
    requirement: CORR-07
    verification:
      - kind: unit
        ref: internal/correct/plan_test.go#TestBuildPlan_BothLevels|RunesNotBytes|TestPlan_TailArithmetic
        status: pass
    human_judgment: false
  - id: D6
    description: "Package purity and green iteration: no godbus/time imports, mise run ci (build+vet+lint+test -race) green"
    verification:
      - kind: automated
        ref: grep-purity-gate + mise run ci
        status: pass
    human_judgment: false

duration: 10min
completed: 2026-09-11
status: complete
---

# Phase 2 Plan 01: Чистый пакет internal/correct Summary

**Headless word-correction library: phrase buffer with letter-capable token boundaries (D-13..D-15), script detector with mixed-word refusal (D-16), case-preserving conversion, ADR-004 suffix verification and both-level rune ladder arithmetic — 16-test corpus green under -race, zero D-Bus/time imports**

## Performance

- **Duration:** 10 min (22:23–22:34 UTC 2026-09-11)
- **Started:** 2026-09-11T22:23:46Z
- **Completed:** 2026-09-11T22:34:00Z
- **Tasks:** 2
- **Files:** 12 created

## Accomplishments

- `internal/correct` exists as a pure library: the whole word-correction logic of Phase 2 is now provable headless under `-race` without a live session (CORR-04/CORR-05/CORR-07 formulas locked before any engine wiring)
- Token semantics implemented verbatim: `,`/`;` inside the token (letter-capable keys), `/`/`-` as boundaries, digits neutral members, the last word correctable after a space, honest backspace pop, ReplaceToken toggle invariant
- BuildPlan ships BOTH ladder levels with the token+tail geometry — plans 02-03 (Level 1) and 02-05 (Level 2) wire them without touching the arithmetic
- Strict TDD honored: two RED commits (stub-anchored corpora, `RED_EVIDENCE_OK` gate records committed under `red-evidence/`) each precede their GREEN commit

## Task Commits

Each task was committed atomically (TDD: test → feat per task):

1. **Task 1 RED: word-correction corpus against the stub** — `fae2359` (test)
2. **Task 1 GREEN: phrase buffer, direction detector, layout conversion** — `1acf59b` (feat)
3. **Task 2 RED: suffix-verification and ladder-arithmetic corpus** — `81c6166` (test)
4. **Task 2 GREEN: suffix verification and both-level ladder arithmetic** — `888effd` (feat)

No REFACTOR commit: the GREEN implementations landed clean, no behavior-preserving cleanup was owed.

## TDD Gate Compliance

| Task | RED | GREEN | REFACTOR | Evidence |
|------|-----|-------|----------|----------|
| 1 | `fae2359` (12 tests, 11 fail on assertions) | `1acf59b` | — not needed | `red-evidence/02-01-task1-red.json` → RED_EVIDENCE_OK |
| 2 | `81c6166` (16 tests, 4 fail on assertions) | `888effd` | — not needed | `red-evidence/02-01-task2-red.json` → RED_EVIDENCE_OK |

Gate sequence `test(02-01)` → `feat(02-01)` twice — verified via `git log --grep`. Target tests (`TestBuffer_TokenRules`, `TestBuildPlan_BothLevels`) failed on assertions, not compilation; stubs returned zero values so every corpus failure was a behavioral assertion.

## Files Created/Modified

- `internal/correct/buffer.go` — phrase Buffer: Push/Backspace/HardReset/Token/Tail/ReplaceToken, letterCapable from the table union, coordinate recompute
- `internal/correct/direction.go` — Dir, Detect: unicode Latin/Cyrillic classification, D-16 refusal
- `internal/correct/convert.go` — Convert: per-rune table lookup, register preserved constructively
- `internal/correct/verify.go` — MatchesSuffix: ADR-004 suffix rule
- `internal/correct/plan.go` — Level/Plan/BuildPlan + CapSurroundingText mirror: both-level token+tail geometry in runes
- `internal/correct/{buffer,direction,convert,verify,plan}_test.go` — the corpus (16 top-level tests, all parallel, external package)
- `red-evidence/02-01-task{1,2}-red.json` — machine-verified RED gate records

## Decisions Made

See key-decisions in frontmatter. Additional notes:

- Backspace keeps the separator in `Tail()` — the honest-buffer invariant («буфер == текст перед курсором») outranks the plan's literal expectation; see Deviations
- No dictionaries/heuristics anywhere in Detect (SPEC §11, plan prohibition `resolved`): classification is pure script composition — the prohibition's `verification: judgment` is satisfied by the direction corpus alone

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] TestBuffer_BackspacePop tail expectation corrected from "" to " "**
- **Found during:** Task 1 (corpus authoring)
- **Issue:** The plan pinned `Tail()==""` after `"ghbdtn v"` + Backspace — but the buffer after an honest one-rune pop is `"ghbdtn "`, the exact state the plan's own TestBuffer_TokenRules pins as `Tail()==" "` (same buffer, same D-13 rule, two different expected values). Honoring `""` would encode Pitfall 1 (the tail-ignoring range arithmetic, threat T-02-01-01) into the corpus: a correction after that backspace would delete only 6 runes and recommit `привет ` behind the surviving space — the `gпривет` prototype bug this phase exists to fix.
- **Fix:** Corpus pins `Tail()==" "` with a comment citing the mirror invariant; Backspace×8→empty still proves the one-rune pop.
- **Files modified:** internal/correct/buffer_test.go
- **Verification:** TestBuffer_BackspacePop green; TestBuffer_TokenRules green on the identical buffer state
- **Committed in:** fae2359 (Task 1 RED)

**2. [Rule 3 - Blocking] goconst blocked the green iteration — corpus literals named**
- **Found during:** Task 1 (GREEN gate)
- **Issue:** The max-strict lint (D-10, `default: all`) flagged 10 goconst issues on repeated corpus literals (`ghbdtn` ×12, `привет` ×4, `ghbdtn2026` ×4, `ghbdtn,` ×3) — `mise run ci` failed.
- **Fix:** Named corpus constants (`wordEN`/`wordRU`/`wordENDigits`/`wordENComma`), the house pattern of fsm_test.go's timing constants; no linter disabled.
- **Files modified:** internal/correct/{buffer,direction,convert}_test.go
- **Verification:** golangci-lint 0 issues; mise run ci green
- **Committed in:** 1acf59b (Task 1 GREEN)

---

**Total deviations:** 2 auto-fixed (1 bug in plan expectation, 1 blocking lint gate)
**Impact on plan:** Both fixes preserve the plan's intent — the first resolves an internal contradiction in favor of the plan's own D-13/Pitfall-1 pins; the second is style-level naming. No scope creep.

## Issues Encountered

None beyond the deviations above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plans 02-03 (tracer: engine wiring, Level 1) and 02-05 (Level 2 actor) consume `Buffer`/`Detect`/`Convert`/`MatchesSuffix`/`BuildPlan` as-is — no formula rework expected
- `correct.CapSurroundingText` mirrors `engine.CapSurroundingText`; keep the two in sync (both cite ibustypes.h)
- The toggle invariant (ReplaceToken) is pinned here; plan 02-03's TestActor_ToggleRepeat builds on it

## Self-Check: PASSED

All 13 created files exist on disk; all 4 task commits (fae2359, 1acf59b, 81c6166, 888effd) found in git log; `mise run ci` green at final code state; purity grep gates green.
