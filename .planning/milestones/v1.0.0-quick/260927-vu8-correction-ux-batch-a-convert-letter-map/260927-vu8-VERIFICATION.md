---
phase: 260927-vu8-correction-ux-batch-a
verified: 2026-09-27T21:27:49Z
status: human_needed
score: 7/7 must-haves verified
covered_files:
  - ".planning/quick/260927-vu8-correction-ux-batch-a-convert-letter-map/260927-vu8-PLAN.md"
  - ".planning/quick/260927-vu8-correction-ux-batch-a-convert-letter-map/260927-vu8-SUMMARY.md"
  - cmd/goswitchd/main.go
  - docs/CONFIG.md
  - docs/SPEC.md
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/config/load_test.go
  - internal/correct/convert.go
  - internal/correct/convert_test.go
  - internal/correct/runs.go
  - internal/correct/runs_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - test/e2e/case_combo.go
  - test/e2e/case_ctl.go
  - test/e2e/case_macr.go
  - test/e2e/case_select.go
  - test/e2e/cases/matrix-v3.yaml
  - test/e2e/matrix.go
covered_digest: "v1:sha256:a63b001e815932e1a6dc9f37160d790216dd4ac4f341e89c0c12326a08f6b8e6"
behavior_unverified: 0
overrides_applied: 0
human_verification:
  - test: "Run the live-desktop e2e matrices (mise run e2e / docs/ci-runner.md) on the owner's GNOME session — matrix-v2 and matrix-v3, including the new flip-after-word-correction zenity case"
    expected: "All cases pass; in flip-after-word-correction the corrected text is привет AND the daemon log shows the mode record strictly after the settled done record (expect_flip_after_done, verifyFlipAfterDone); no gate flakes from correction mode records under the default-ON flip"
    why_human: "The matrices drive the owner's real desktop (ydotool, AT-SPI, live ibus) — deliberately outside mise run ci per the plan and docs/ci-runner.md; deferred to the orchestrator/owner acceptance step by the plan itself (pending, not failed)"
---

# Quick Task 260927-vu8: Correction UX batch A Verification Report

**Task Goal:** correction UX batch A: convert letter-mapped punctuation in corrections and flip the script mode after a successful word/phrase correction (config-gated, default on)
**Verified:** 2026-09-27T21:27:49Z
**Status:** human_needed (all 7 must-haves verified; one plan-deferred live-desktop acceptance run pending)
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Allow-set bracket-row symbols convert together with the word (`ghbdtn[` → `приветх`) per SPEC 4.2 / owner decision 1 revised | ✓ VERIFIED | `internal/correct/convert.go:24-32` — `bracketRowAllowSet()` holds exactly `[` `]` `;` `'` `` ` ``; the Convert loop converts them EN→RU (`convert.go:60-66`); all five table entries exist in `layouts/tables.go` (`'['→'х'` :72, `']'→'ъ'` :74, `';'→'ж'` :40, `''''→'э'` :20, ``'`'→'ё'`` :77). Behavioral: `convert_test.go:48-53` (all five, incl. the live-defect shape) and `runs_test.go:111` pass in the green corpus run |
| 2 | Comma and dot ride along unchanged — D-14 golden intact — as do space, parentheses, slash, symbol-row class, digits | ✓ VERIFIED | D-14 golden `convert_test.go:44` (`wordENComma`→`привет,`) and `runs_test.go:105` pass UNMODIFIED; ride-along pins `convert_test.go:60-61` (dot, '@'-class) pass; code path `convert.go:67` (non-allow-set non-letters append identically); ',' and '.' deliberately absent from the allow-set with the owner-decision comment |
| 3 | Allow-set symbols stay scriptNeutral for the run anchor; convert inside a wholesale range, ride as typed inside a mixed range | ✓ VERIFIED | `classifyRune` (runs.go:20-29) untouched — the symbols are neither Latin nor Cyrillic, so they build no segment and shift no anchor; behavioral: `runs_test.go:54-60` mixed-neutrality pin (`gfbпривет[`→`паипривет[` — bracket rides, only the Latin run converts) and `runs_test.go:114` (`привет[`→`ghbdtn[` — RU direction has no '[' entry, rides); wholesale conversion pinned at `runs_test.go:111` |
| 4 | A changed word (double) / phrase (triple) correction flips the script mode exactly once, strictly after the settled done record (D-36 extended) | ✓ VERIFIED | `settleCorrectionFlip` (`actor.go:651-656`) called at exactly two sites: `executeCorrection` :1523 (after `logCorrectionDone` :1522, before `settleCombo` :1524) and `executeLevel2` :1573 (after :1572, before :1574). Behavioral: `TestActor_FlipAfterWordCorrection` (:3192), `TestActor_FlipAfterPhraseCorrection` (:3231), `TestActor_FlipAfterLevel2Correction` (:3266) each pin exactly one mode record, the done-before-mode `strings.Index` order pin, and (word path) zero extra sink calls — all pass in the green run |
| 5 | The D-24 done-without-change outcome and every refusal never flip | ✓ VERIFIED | Refusals behaviorally pinned: `TestActor_RefusalsNeverFlip` (:3351) drives empty-buffer / no-letters / verify-mismatch / verify-timeout with the option ON and asserts zero mode records — passes. D-24: deterministic static exclusion — `settleCorrectionFlip` has exactly two call sites (grep-verified), neither on the `!changed` branch (`actor.go:1456-1464` calls only countCorrectionDone + slog + settleCombo); additionally the branch is unreachable through `ConvertRuns` for word/phrase (runs.go:94 hardcodes `changed=true` wholesale; convertForeignRuns always `converted=true` :148), so no runtime state leak is possible on a path whose body contains no flip call |
| 6 | A combo gesture still flips exactly once (settleCombo/D-36) — the correction flip never doubles it | ✓ VERIFIED | Guard `if a.comboPending || !a.opts.FlipAfterCorrection { return }` (`actor.go:652`) — the helper no-ops when a combo is pending and settleCombo (called after it, :1524/:1574) fires the gesture's single flip. Behavioral: `TestActor_ComboStillSingleFlipWithCorrectionFlip` (:3395, option ON over a correctable word → exactly one mode record) passes; pre-existing `TestActor_ComboWordThenFlip` (zero-Options) stays green in the full run |
| 7 | `correction.flip_after_correction` gates the behavior; built-in defaults ON; hot reload applies document changes live (CONF-02) | ✓ VERIFIED | Config: field `config.go:79` (`yaml:"flip_after_correction"`), `Defaults()` true `config.go:120`; `TestDefaults` asserts ON (`config_test.go:67-69`); decode round-trip false/true pinned (`load_test.go:165-185`). Wiring: `applySnapshot` fold `actor.go:720` (per-event, CONF-02); daemon `main.go:128-131` feeds `cfg.Correction.FlipAfterCorrection` through SetOptions. Behavioral: `TestActor_HotReloadFlipAfterCorrection` (:3472) pins attached-doc-ON, true→false stops the flip for the NEXT correction, false→true restores it — passes |

**Score:** 7/7 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/correct/convert.go` | Explicit allow-set for bracket-row symbols | ✓ VERIFIED | Substantive (79 lines, allow-set fn + converted loop + owner-decision comments); wired — the single conversion entry reused by ConvertRuns (no fork, T-03-01-03) |
| `internal/config/config.go` | `Correction.FlipAfterCorrection`, Defaults() true | ✓ VERIFIED | Field :79 with yaml tag; Defaults true :120; no Validate change (plain bool — per plan) |
| `internal/session/actor.go` | Options field, applySnapshot fold, settleCorrectionFlip at the two changed-success sites | ✓ VERIFIED | `Options.FlipAfterCorrection` :160 with doc comment :148-152; fold :720; helper :651-656; call sites :1523/:1573 exactly |
| `cmd/goswitchd/main.go` | SetOptions wiring | ✓ VERIFIED | :131 `FlipAfterCorrection: cfg.Correction.FlipAfterCorrection` in newActor's SetOptions literal; cfg from `loadConfig` (Defaults()/Load) |
| `docs/CONFIG.md` + `docs/SPEC.md` | Documented key and behavior | ✓ VERIFIED | CONFIG.md: schema row :31 (default true), both example documents :64/:84, completeness note :19 (missing key = off, not defaulted), Caramba correspondence row :127. SPEC.md: §4.2 уточнение (only «[» «]» «;» apostrophe backtick convert; comma/dot ride) + §4.3 bullet :92 (flip after changed correction, D-36 order, default on, D-24/refusal exclusion) |
| `test/e2e` reconciled | Gates, flows, one new live case | ✓ VERIFIED | Both mode gates baseline-counted (`comboMatrixStep` matrix.go:996-1016, `tapMatrix` :1228-1262 — countSub before, waitForNew base+1 after, rationale cited); new live case `flip-after-word-correction` matrix-v3.yaml:373-380 with case-level `expect_flip_after_done` enforced by `verifyFlipAfterDone` (matrix.go:1432-1457, last-occurrence D-36 order check); `flip_after_correction: false` pinned in all six written config documents (case_combo.go:62, case_ctl.go:36/:58, case_select.go:87, case_macr.go:481/:723) |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| convert.go allow-set | layouts.ENToRU bracket-row entries | table lookup `table[r]` inside the allow-set branch | ✓ WIRED | All five allow-set runes resolve in the generated table (tables.go :20/:40/:72/:74/:77); a set entry the table lacked would ride (defensive `ok` check) — none missing |
| actor.applySnapshot | config.Correction.FlipAfterCorrection | CONF-02 value fold | ✓ WIRED | `actor.go:720` `a.opts.FlipAfterCorrection = snap.Correction.FlipAfterCorrection`, beside the ClipboardRung fold; exercised live by the hot-reload test |
| settleCorrectionFlip | comboPending guard ↔ flipScript | guard chain | ✓ WIRED | `actor.go:651-656`: comboPending check precedes the option check, flipScript is the only flip action; call order (helper then settleCombo) at both sites guarantees the guard reads the flag before settleCombo clears it |
| cmd/goswitchd newActor SetOptions | config.Load/Defaults | startup config literal | ✓ WIRED | main.go:104-118 loadConfig (Defaults() when -config absent, Load otherwise) → newActor :125-141 SetOptions |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| settleCorrectionFlip | `a.opts.FlipAfterCorrection` | config document via cfgSrc.Snapshot() / Defaults() via SetOptions | Yes — real YAML decode or built-in default | ✓ FLOWING |
| Convert allow-set branch | `layouts.ENToRU` | generated key-position tables (`layouts/tables.go`, go:generate) | Yes — generated 1:1 table data | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| correct corpus (allow-set + ride-alongs + refusals) | `go test -race -count=1 ./internal/correct/` | ok 1.015s | ✓ PASS |
| config corpus (Defaults ON + decode round-trip) | `go test -race -count=1 ./internal/config/` | ok 1.120s | ✓ PASS |
| session corpus (flip-on pins, gate-off, refusals, combo single-flip, hot reload) | `go test -race -count=1 ./internal/session/` | ok 3.175s | ✓ PASS |
| Full green-iteration gate (build + vet + golangci-lint + `go test -race ./...`) | `mise run ci` | all 13 packages ok incl. test/e2e headless; Finished, exit 0 | ✓ PASS |
| e2e driver compiles + headless tests | included in `mise run ci` (test/e2e ok 1.017s) | ok | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
|-------|---------|--------|--------|
| (none declared) | — | No `scripts/*/tests/probe-*.sh` probes exist or are declared for this task; the plan's live verification surface is the e2e matrix run (see Human Verification) | N/A |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| SPEC-4.2 | 260927-vu8-PLAN | Ключе-позиционная конвертация «включая знаки» | ✓ SATISFIED | Allow-set conversion implemented and tested; SPEC.md §4.2 carries the owner's revised уточнение verbatim |
| OWNER-DEC-1 | 260927-vu8-PLAN | Revised allow-set rule: «[ ] ; ' \`» convert, ',' and '.' ride as universal prose punctuation | ✓ SATISFIED | convert.go allow-set with the exact boundary; D-14 goldens pass unmodified; boundary ride-along pins pass |
| OWNER-DEC-2 | 260927-vu8-PLAN | Flip script mode after a changed word/phrase correction, default on, D-24/refusals never flip, combo single-flip | ✓ SATISFIED | settleCorrectionFlip at exactly the two changed-success sites; full behavioral pin set (truths 4-7) |
| CONF-02 (key_links) | 260927-vu8-PLAN | Hot-reload value-fold applies document changes live | ✓ SATISFIED | applySnapshot fold + TestActor_HotReloadFlipAfterCorrection (true→false→true) |

REQUIREMENTS.md traceability maps phases 1-4 only; no requirement IDs in REQUIREMENTS.md are mapped to this quick task, so there are no orphaned requirements. The plan's `requirements: [SPEC-4.2, OWNER-DEC-1, OWNER-DEC-2]` reference docs/SPEC.md and the owner decisions quoted in the plan objective — all three satisfied.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none) | — | No TBD/FIXME/XXX/HACK/PLACEHOLDER markers, no empty implementations, no hardcoded-empty data in any of the 18 changed files | — | — |

### Deviations Reviewed (from SUMMARY claims — checked against code, all legitimate)

1. **Direction-gated allow-set (EN→RU only).** The plan's GREEN prose («in the allow-set AND the direction's table carries it») read literally would convert RU ';'→'$' (RUToEN carries it), but the plan's own RED pin demands «RU ';' rides». The code implements EN→RU-only (`convert.go:60` `dir == ENtoRU && allow[r]`) and both directions are pinned (`runs_test.go:114`). The pin wins; consistent with the owner decision's wording («the EN symbols»). Legitimate resolution of an in-plan ambiguity.
2. **case_word.go, case_phrase.go, matrix-v2.yaml listed in plan `files_modified` but unchanged.** Task 3's flow-rebalancing instructions were conditional on the audit finding cases that type after a changed correction; the audit (SUMMARY decision 4) found every correction step terminal in every case. The new live case was correctly added to matrix-v3.yaml as the plan instructed. No pin weakened — verified by reading the gate code and the v3 case.

### Human Verification Required

### 1. Live-desktop e2e matrix run (plan-deferred acceptance step — pending, not failed)

**Test:** Run matrix-v2 and matrix-v3 on the owner's real GNOME session per docs/ci-runner.md, including the new `flip-after-word-correction` zenity case (type `ghbdtn`, double tap, expect `привет`).
**Expected:** All cases pass. In the new case, the corrected text lands AND `verifyFlipAfterDone` proves the mode record strictly follows the settled done record in the daemon log. The hardened baseline-counted mode gates must not flake under the default-ON flip.
**Why human:** The matrices drive the owner's real desktop (ydotool injection, AT-SPI reading, live ibus daemon) and are deliberately excluded from `mise run ci` by the plan (Task 3: "the live run is the owner's acceptance step") and by docs/ci-runner.md. This verification deliberately did not execute them, per the orchestrator's instruction to record the live run as pending.

Observation for that run: no live case pins the punctuation conversion itself (the plan recorded during grounding that probes are letters+space only, and Task 3's new live pin covers only the flip) — the punctuation fix's live evidence is the owner's own «несколько дополнительны[» repro, worth a manual glance during the acceptance run.

### Gaps Summary

None. All 7 must-have truths are verified against the codebase with behavioral test evidence: the allow-set conversion and its exact comma/dot boundary, the anchor-neutrality of the allow-set symbols, the single post-done flip on word/phrase/level-2 success, the D-24/refusal exclusions, the combo single-flip guarantee, and the config gate with default-ON and live hot reload. Documentation (CONFIG.md schema row + both examples + completeness note + Caramba row; SPEC §4.2 уточнение + §4.3 bullet) matches the implementation. All 5 commits (4c59844, f17b817, d90f186, 1decb2b, f679b95) are conventional on `gsd/quick-correction-ux`; `mise run ci` re-run by this verifier is green. The only outstanding item is the plan-deferred live-desktop matrix run, recorded above as pending human/orchestrator acceptance — the basis for status `human_needed`.

---

_Verified: 2026-09-27T21:27:49Z_
_Verifier: Claude (gsd-verifier)_
