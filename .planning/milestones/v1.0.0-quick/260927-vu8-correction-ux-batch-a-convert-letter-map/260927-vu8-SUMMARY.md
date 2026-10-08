---
phase: 260927-vu8-correction-ux-batch-a
plan: 01
subsystem: correction + session + config + e2e
tags: [correction, allow-set, bracket-row, flip-after-correction, owner-decision, spec-4.2]
requires:
  - layouts.ENToRU/RUToEN bracket-row entries ('['→'х', ']'→'ъ', ';'→'ж', '''→'э', '`'→'ё')
  - internal/session actor settlement sites (executeCorrection, executeLevel2, settleCombo/D-36)
  - internal/config strict KnownFields decode + Defaults() (03-02)
provides:
  - correct.bracketRowAllowSet — the explicit EN→RU allow-set in Convert (SPEC §4.2 «включая знаки», owner decision 1 revised)
  - session.Options.FlipAfterCorrection + settleCorrectionFlip (owner decision 2, D-36 order extended)
  - config.Correction.FlipAfterCorrection (`correction.flip_after_correction`, default ON, hot-reload live)
  - e2e expect_flip_after_done case pin + baseline-counted mode gates in matrix.go
affects: []
tech-stack:
  added: []
  patterns: [allow-set-as-function (no mutable globals), flip-after-settle guard chain, baseline-count log gates]
key-files:
  created: []
  modified:
    - internal/correct/convert.go
    - internal/correct/convert_test.go
    - internal/correct/runs.go
    - internal/correct/runs_test.go
    - internal/config/config.go
    - internal/config/config_test.go
    - internal/config/load_test.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - cmd/goswitchd/main.go
    - docs/CONFIG.md
    - docs/SPEC.md
    - test/e2e/matrix.go
    - test/e2e/case_combo.go
    - test/e2e/case_ctl.go
    - test/e2e/case_select.go
    - test/e2e/case_macr.go
    - test/e2e/cases/matrix-v3.yaml
decisions:
  - "Allow-set is direction-gated (EN→RU only): the plan's GREEN wording («in the allow-set AND the direction's table carries it») read literally would convert RU ';'→'$' (RUToEN carries it), but the plan's own RED pin demands «RU ';' rides» — the pin wins; a ';' in a Russian-typed text is RU-keyboard punctuation (Shift+4), and the owner decision scopes the set to «the EN symbols»"
  - "Missing flip_after_correction key decodes false (off), not the built-in default — keeps every pre-existing config fragment loading unchanged under the strict decoder; the completeness duty moves to the document author (documented in CONFIG.md)"
  - "Selector/selection correction never flips (owner decision 2 scopes the flip to word+phrase); D-24 done-without-change and every refusal are structurally excluded — only the two changed-success sites call settleCorrectionFlip"
  - "e2e matrices v2/v3 needed NO flow re-balancing: the audit showed every correction step is terminal in every case (nothing types after a changed correction); only the gates needed hardening and config-carrying cases needed the explicit flip:false pin"
metrics:
  duration: 32 min
  completed: 2026-09-27
status: complete
actuals:
  tokens: 13294   # chars/4 over the realized diff (53176 chars, +699/−43 lines, 18 files)
  tasks: 3
  commits: 5      # MEASURED: git rev-list --count 64cffc8..HEAD
plan_head_before: 64cffc8c801fed1d3aab949a35f29af98a9ad58d
---

# Quick Task 260927-vu8: Correction UX batch A (allow-set letter map + flip after correction) Summary

Bracket-row punctuation now converts during correction (`ghbdtn[`→`приветх`, the live defect fixed) via an explicit EN→RU allow-set, and a changed word/phrase/level-2 correction flips the internal script mode exactly once after the settled done record — gated by the new `correction.flip_after_correction` key (default ON, hot-reload live, D-36 order extended).

## Tasks

### Task 1 — Allow-set bracket-row conversion in correct.Convert (STRICT TDD)

- **RED** (`4c59844`): corpus extended before touching Convert — 5 allow-set conversion cases (incl. the live-defect shape `ghbdtn[`→`приветх`), ride-along pins (',' '.' '@'-class, digits/space/slash; RU '"'/';' ride in RU→EN), wholesale/mixed/lone-bracket runs cases. `go test -race ./internal/correct/` exit 1: exactly the six target cases failed on the planned assertions (`Convert("ghbdtn[", ENtoRU) = "привет[", want "приветх"` etc.); all ride-along pins green. RED evidence persisted at `/tmp/vu8-red-evidence.json`; note: the SDK `check tdd-red-evidence` verb only parses node:test TAP output and cannot classify `go test` output (verdict `zero_tests_discovered` is a parser-format artifact — the genuine gate conditions, nonzero exit + named target tests failing on the planned assertions, are met and recorded above).
- **GREEN** (`f17b817`): `bracketRowAllowSet()` (function returning map, strict-lint idiom) + the Convert loop change + doc comments on Convert and ConvertRuns (scriptNeutral: anchor-neutral in mixed ranges, converted inside wholesale ranges). No fork of Convert (T-03-01-03); runs.go is comment-only.
- **REFACTOR**: D-14 goldens pass unmodified; comma/dot literals across internal/correct reconciled (none affected — buffer_test's letterCapable pins are Buffer membership, untouched). No changes → no refactor commit per the TDD contract.

### Task 2 — flip_after_correction end to end (STRICT TDD)

- **RED** (`d90f186`): actor corpus (word/phrase/level-2 flip-on pins with zero extra sink calls and the done→mode order pin; gate-off at zero value and explicit false; four refusal no-flip pins; combo single-flip pin; hot-reload true→false/false→true) + config corpus (TestDefaults ON, full-doc decode, false/true round-trip). Scaffolding so targets fail on assertions, not the build: the `Correction.FlipAfterCorrection` yaml field and the `Options.FlipAfterCorrection` field exist unwired. RED run exit 1: TestActor_FlipAfterWordCorrection / …PhraseCorrection / …Level2Correction / TestActor_HotReloadFlipAfterCorrection / TestDefaults failed on the planned assertions; all pins green.
- **GREEN** (`1decb2b`): Defaults() ON; `settleCorrectionFlip` (no-op on comboPending or off) called at exactly the two changed-success sites after logCorrectionDone and before settleCombo; applySnapshot fold (CONF-02); main.go SetOptions wiring; Options doc comment; CONFIG.md (schema row, both example documents, Caramba row, pre-existing-document note) and SPEC.md (§4.3 bullet + the §4.2 clarifying parenthetical: only «[ ] ; ' \`» convert, ',' and '.' ride as universal prose punctuation).
- **REFACTOR**: whole session/config corpora green unmodified (the zero-value-off reading keeps every pre-existing test's world intact); lint-driven test restructuring (funlen/goconst/lll) folded into the green commit per the зелёная-итерация directive. No further changes → no refactor commit.

### Task 3 — e2e reconciliation (deliberate updates only)

- Gates hardened: `tapMatrix`'s single-tap mode gate and `comboMatrixStep`'s mode gate now count a NEW record over a baseline taken BEFORE the gesture (the case_combo.go:124 pattern) — the old whole-log `count ≥ 1` could be satisfied early by a correction's own mode record under the default-ON flip.
- New live pin: `flip-after-word-correction` case in matrix-v3.yaml (zenity, default daemon = flip ON) with the new case-level `expect_flip_after_done` field checked by `verifyFlipAfterDone` (last done record followed by the mode record — the live D-36 order extension; the YAML schema cannot express ordering, so the pin rides the case level).
- Config-carrying cases pin the pre-batch semantics: `flip_after_correction: false` added to all six written documents (comboConfigTmpl, ctlConfigYAML — also the matrix reload base, ctlBrokenYAML, clipboardConfigYAML, macrConfigTmpl, macrPerAppConfigTmpl).
- Flows audited case by case: every YAML-matrix correction step (v2 and v3, including all gedit/x11 rows) and every standalone case_word/case_phrase flow ends at its correction — **no case types after a changed correction**, so no balancing single-tap steps or re-orderings were needed; no pin was weakened.
- **Live-run verify: deferred to orchestrator.** The matrices drive the owner's real desktop and are deliberately outside `mise run ci` (docs/ci-runner.md); this task's green gate was `go vet ./test/e2e/` + headless `go test -race ./test/e2e/` + `mise run ci` (all green). The owner's live acceptance run of matrix-v2/v3 remains the orchestrator's step.

## Deviations from Plan

None — plan executed as written. Two in-plan ambiguities were resolved toward the plan's own pinned tests (recorded under Decisions): the direction-gated allow-set (RED pin «RU ';' rides» over the GREEN prose) and the hot-reload test's RU-mode second episode (the flip from the first episode makes the buffer script-true Cyrillic, so the field push carries `wordRU` — the test drives the field truth the daemon must mirror).

## Known Stubs

None.

## Authentication Gates

None.

## Verification

- `mise run ci` green after every task (build + vet + golangci-lint 0 issues + `go test -race -count=1 ./...`).
- Correct corpus: five-symbol allow-set converts wholesale; ',' '.' ride (D-14 goldens unmodified); mixed-path neutrality (`gfbпривет[`→`паипривет[`); '@'-class/digits ride; «é» still refuses wholesale; lone `[` still refuses no-letters.
- Session corpus: flip-on-success pins (word/phrase/level-2, one mode record strictly after done, zero extra sink calls); combo single-flip; D-24 and four refusal no-flip pins; gate-off (zero value and false); hot-reload toggles.
- Config: TestDefaults ON; decode round-trip; defaults live in Defaults() only.
- e2e: compiles (vet), headless tests green, gates baseline-counted, one new default-ON live case.

## Self-Check: PASSED

- Commits verified: 4c59844, f17b817, d90f186, 1decb2b, f679b95 — all present on gsd/quick-correction-ux (git log).
- Files verified: all 18 modified files exist in the tree; no unexpected deletions in any commit (diff-filter=D checked per commit).
- Commit count measured: 5 = `git rev-list --count 64cffc8..HEAD`.
