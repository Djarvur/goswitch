---
phase: 02-korrektsiya-slova-en-ru
verified: 2026-10-07T22:51:15Z
status: passed
score: 23/23 must-haves verified
covered_files:
  - .github/workflows/e2e-matrix.yml
  - .gitignore
  - .planning/phases/02-korrektsiya-slova-en-ru/02-01-PLAN.md
  - .planning/phases/02-korrektsiya-slova-en-ru/02-01-SUMMARY.md
  - .planning/phases/02-korrektsiya-slova-en-ru/02-02-PLAN.md
  - .planning/phases/02-korrektsiya-slova-en-ru/02-02-SUMMARY.md
  - .planning/phases/02-korrektsiya-slova-en-ru/02-03-PLAN.md
  - .planning/phases/02-korrektsiya-slova-en-ru/02-03-SUMMARY.md
  - .planning/phases/02-korrektsiya-slova-en-ru/02-04-PLAN.md
  - .planning/phases/02-korrektsiya-slova-en-ru/02-04-SUMMARY.md
  - .planning/phases/02-korrektsiya-slova-en-ru/02-05-PLAN.md
  - .planning/phases/02-korrektsiya-slova-en-ru/02-05-SUMMARY.md
  - .planning/phases/02-korrektsiya-slova-en-ru/02-06-PLAN.md
  - .planning/phases/02-korrektsiya-slova-en-ru/02-06-SUMMARY.md
  - .planning/phases/02-korrektsiya-slova-en-ru/02-07-PLAN.md
  - .planning/phases/02-korrektsiya-slova-en-ru/02-07-SUMMARY.md
  - .planning/quick/261006-vqw-mixed-text-per-character-layout-inversio/261006-vqw-PLAN.md
  - .planning/quick/261006-vqw-mixed-text-per-character-layout-inversio/261006-vqw-SUMMARY.md
  - .planning/quick/261007-0yg-move-engine-and-layouts-under-internal-o/261007-0yg-PLAN.md
  - .planning/quick/261007-0yg-move-engine-and-layouts-under-internal-o/261007-0yg-SUMMARY.md
  - .planning/quick/261008-00m-milestone-close-owner-fixes-batch-1-g-2-/261008-00m-PLAN.md
  - .planning/quick/261008-00m-milestone-close-owner-fixes-batch-1-g-2-/261008-00m-SUMMARY.md
  - README.md
  - SECURITY.md
  - cmd/goswitchd/main.go
  - docs/SPEC.md
  - docs/adr/ADR-004-buffer-reset-triggers.md
  - docs/ci-runner.md
  - go.mod
  - go.sum
  - internal/correct/buffer.go
  - internal/correct/buffer_test.go
  - internal/correct/convert.go
  - internal/correct/convert_test.go
  - internal/correct/direction.go
  - internal/correct/direction_test.go
  - internal/correct/plan.go
  - internal/correct/plan_test.go
  - internal/correct/runs.go
  - internal/correct/runs_test.go
  - internal/correct/verify.go
  - internal/correct/verify_test.go
  - internal/engine/emitter_wire_test.go
  - internal/engine/engine.go
  - internal/engine/engine_test.go
  - internal/engine/factory.go
  - internal/engine/keys.go
  - internal/engine/types.go
  - internal/hotkey/fsm.go
  - internal/hotkey/fsm_test.go
  - internal/layouts/tables.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - test/e2e/README.md
  - test/e2e/case_d01.go
  - test/e2e/case_ladder.go
  - test/e2e/case_m1.go
  - test/e2e/case_resilience.go
  - test/e2e/case_surface.go
  - test/e2e/case_word.go
  - test/e2e/cases/matrix-v1.yaml
  - test/e2e/cases/matrix-v2.yaml
  - test/e2e/fixtures/input.html
  - test/e2e/focus_helper.py
  - test/e2e/main.go
  - test/e2e/matrix.go
  - test/e2e/matrix_test.go
  - test/e2e/preflight.go
  - test/e2e/surface.go
covered_digest: "v1:sha256:e338d15e58e9bce109ff2143c1e764e5c44a67aed60b266bb728a121562eda4a"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: human_needed
  previous_score: 22/23
  previous_verified: 2026-10-07T07:49:59Z
  reason: "human-item closure (final canonicalization pass of the re-verification chain): both human verification items of the 2026-10-07 report are resolved — (1) the live matrix v1 witness was recorded 2026-10-08 (16/16 PASS, word-mixed → «паиghbdtn») after quick 261008-00m fixed the synthetic-focus buffer reset the first live attempt exposed (ADR-004 amendment, RED 77c91b7 → GREEN 4a0c887) plus stand hermeticity (RED 07d143b → GREEN 4dc4877), recorded in 02-UAT.md test 3 and gap G-2-3; (2) the hand-typed mixed-text gedit check is closed by composition (mechanism identity code-proven + owner's visual confirmation of that mechanism at UAT 2 + live content/range oracle), ruling recorded in the report body. All 22 previously verified truths regression-checked: every file behind them is untouched since the prior pass (git log empty); actor.go/matrix.go changes deep-checked"
  gaps_closed:
    - "SC-5 (live dimension): the current matrix-v1.yaml — with its two re-pinned rows word-punct («приветб») and word-mixed («паиghbdtn») — has no live green run yet"
  gaps_remaining: []
  regressions: []
---

# Phase 2: Коррекция слова EN↔RU — Verification Report

**Phase Goal (ROADMAP.md):** «Пользователь двойным нажатием Right Shift исправляет последнее слово, набранное не в той раскладке, в любом поле ввода GNOME — направление определяется автоматически, регистр сохраняется, замена происходит ровно по диапазону неверного текста без визуального прыжка; это подтверждает зелёная e2e-матрица v1.»
**Verified:** 2026-10-07T22:51:15Z
**Status:** passed (23/23; both human items of the 2026-10-07 report closed — one by recorded live evidence, one by ruling recorded below)
**Re-verification:** Yes — human-item closure after quick 261008-00m (chain: original pass 2026-09-15 UAT 2/2 owner-accepted → passed 23/23 2026-09-20 → human_needed 22/23 2026-10-07 at fingerprint refresh → this pass)

## What changed since the 2026-10-07 report (HEAD ae54bf4, last code commit 4dc4877)

Quick **261008-00m** (milestone-close owner fixes, batch 1) plus docs-only commits
on top. Exactly three code/doc surfaces changed for phase 2; `git log` proves
**every other phase-2 file untouched** since the prior pass (internal/correct/*,
internal/engine/*, internal/layouts/*, internal/hotkey/*, matrix-v1.yaml,
case_word.go, surface drivers, preflight, workflow, main.go — all empty logs):

1. **flipCredit — own engine flip keeps the correction buffer** (G-2-3;
   spec-first: ADR-004 amendment 2026-10-07 + SPEC §4.3 dated bullet, commit
   0781a5f; RED 77c91b7 with the exact live failure signature `(-6,6)`
   «ghbdtn» vs wanted `(-9,9)` «паиghbdtn»; GREEN 4a0c887):
   `internal/session/actor.go` — `flipTo` arms exactly one synthetic-lifecycle
   credit AFTER a successful `SetGlobalEngine` round trip (ibus re-mints
   engine objects → a synthetic focus_out/focus_in pair arrives);
   `HandleLifecycle`'s FocusOut branch consumes exactly one credit and skips
   ONLY `buf.HardReset()` — FSM reset, surrounding-cache clear,
   pending-round retirement, clearAfter, combo/MACR cleanup and timer stop
   all stay. A WARNed switcher failure or nil-seam degradation arms nothing;
   the Reset lifecycle kind never consumes a credit. Owner decision recorded
   verbatim in ADR-004's amendment section.
2. **Stand hermeticity** (07d143b RED → 4dc4877 GREEN):
   `test/e2e/matrix.go` — `caseNeedsBaseEstablishment` predicate: EVERY case
   daemon now boots on a pinned `-config` document (plain cases previously
   adopted the desktop owner's live config, firing owner autocorrect inside
   v1 cases — the second half of the 15/16 live failure); reload-step cases
   are the single exception (their first reload step IS the establishment).
   Strict loader untouched: `dec.KnownFields(true)` still at matrix.go:367;
   matrix_test.go +49/+15 pins the predicate (TestMatrixCaseBaseConfig,
   TestMatrixCaseNeedsBaseEstablishment — PASS in this verifier's run).
3. **G-5-5 auto-wrap** (2fa8b53) touched `internal/install/*` only — phase-5
   scope, tracked by phase 5's own verification; not a phase-2 surface.

**No code changed after 4dc4877** (`git log 4dc4877..HEAD -- internal/ test/
cmd/ go.mod go.sum` is empty; the two commits above it are planning docs) —
the live 16/16 witness of 2026-10-08 therefore ran on code byte-identical to
this HEAD.

## Closure of the two human verification items

### Item 1 — Live matrix v1 witness: CLOSED by recorded live evidence

First live attempt (2026-10-07) FAILED 15/16 — word-mixed settled to
«gfbghbdtn»: the daemon's own flip generated a synthetic focus_out/focus_in
and `HandleLifecycle` hard-reset the correction buffer mid-word, correcting
only the post-flip token (daemon log: single correction `runes:6
source:привет`). The prior pass was right to withhold the live dimension.
Diagnosed from the full daemon log; the unit corpus was blind to the real
flip (fixture-only); fixed spec-first then RED→GREEN (above), plus the
hermeticity defect (plain cases adopting the owner's live config) in the same
batch. **Final live run 2026-10-08: matrix v1 16/16 PASS, exit 0 —
word-mixed settled to «паиghbdtn», word-punct to «приветб», all other rows
unchanged green** (registers, D-13 separator, D-15 digits, resets, both
directions). Recorded in 02-UAT.md test 3 (result: pass, 2026-10-08) and gap
G-2-3 (status: resolved, resolved_by 261008-00m); commit 9cb032c. Truth 5's
live dimension is witnessed; the last-standing PRESENT_BEHAVIOR_UNVERIFIED
upgrades to VERIFIED.

### Item 2 — Hand-typed mixed-text gedit check (the perceptual dimension): CLOSED by composition — ruling

The owner has not personally performed the gedit check (offered, pending).
Ruled **satisfied** on the following chain — each link independently
evidenced, none assumed:

1. **Mechanism identity is code-proven, not analogous.** The correction
   pipeline has ONE conversion entry and ONE emission path with no
   content-type branch: `startRangeCorrection` (actor.go:2745) →
   `correct.ConvertRuns` (:2754; the mixed/homogeneous branch is entirely
   inside content production — runs.go `invertPerChar` vs `Convert`) →
   BuildPlan → level 1 emits exactly one `DeleteSurroundingText(plan.Offset,
   plan.NChars)` + one `CommitText(plan.Commit)` (:2835-2836); level 2 emits
   the backspace burst + one CommitText (:2895). A mixed-word correction and
   a plain-word correction differ in NO emitted call sequence — only in the
   committed bytes.
2. **The perception is a property of that mechanism, owner-confirmed.** The
   «no visual jump» clause was accepted by the owner visually for THIS
   emission sequence on THIS desktop (02-UAT test 2, 2026-09-15: in-place
   replacement, no jump, corroborated by the daemon log
   `correction done … level:1 runes:6 → verify outcome:match`).
3. **The mixed path's content and range dimensions are now live-proven.**
   The 2026-10-08 matrix word-mixed row performs the exact hand-typed
   sequence (type `gfb`, single-tap flip, type `ghbdtn`, double-tap) via
   ydotool on a live surface, and the AT-SPI oracle reads back the EXACT
   final field content «паиghbdtn» — a range error would corrupt surrounding
   text and fail the oracle; a content error fails it trivially. 16/16.
4. **The owner's eyes have been on every mixed-semantics decision:** verdict
   2026-10-03 (superseding D-22/D-23), directive 2026-10-06 (per-character
   inversion), ADR-004 amendment decision 2026-10-07 (own flip ≠ focus
   change).

**Honest residual:** no human has watched a mixed correction's transient in
real time; the ruling rests on there being no causal pathway for the
transient to be content-dependent given links 1–3 (visual dynamics belong to
the emission sequence, not the bytes). If the owner's live use ever shows a
jump on mixed text, this reopens as a defect report — nothing in this
closure forecloses that. On the evidence standard applied throughout this
chain (recorded live witnesses + owner verdicts + code identity), the item
is satisfied; it is NOT counted as an open human-verification item.

## User Flow Coverage

User capability (goal, non-story form — MVP user-flow framing discrepancy
recorded in the initial pass, unchanged).

| Step | Expected | Evidence | Status |
|------|----------|----------|--------|
| Type a word in the wrong layout | `ghbdtn` (registers/mixed/digits variants) lands in the field | matrix-v1.yaml 16 cases intact at HEAD; case_word.go registry wired | ✓ |
| Double-tap Right Shift | FSM Double → correction pipeline | TestActor_DoubleTapCorrects PASS in this verifier's `-race` run at HEAD; live 16/16 2026-10-08 | ✓ |
| Word corrected, direction automatic | EN→RU and RU→EN; mixed → per-character inversion (D-55) | named conversion+actor tests PASS at HEAD; live both directions @9f2fd75 + Oct-5 + the 2026-10-08 mixed row | ✓ |
| Register preserved | three registers | TestConvert_CasePreserved PASS (this verifier's run); live registers @9f2fd75 + Oct-5; v1 register rows live @419799d | ✓ |
| Exact-range replacement, no jump | delete+commit over the range, one transaction | BuildPlan rune arithmetic + single emission path (above); live oracle exact-content 16/16 2026-10-08; perceptual check owner-accepted (UAT 2) + closure ruling for mixed | ✓ |
| Works in any GNOME input field | GTK/Chromium/GtkSourceView + gedit + chromium-X11 drivers | live v3 31/31 twice @9f2fd75; v4 31/34×2 @c84709c (fails = designed pre-re-pin mixed rows + transient); drivers untouched since | ✓ |
| Green matrix proves it | 16/16 v1; exit≠0 | **live 16/16 PASS 2026-10-08 on the current code** (UAT test 3; G-2-3 resolved); infra green at HEAD | ✓ |

## Goal Achievement

### Observable Truths

Same 23 consolidated truths as the accepted prior passes; re-derived at HEAD
ae54bf4 (R = regression-checked — every backing file untouched since the
prior pass; D = deep-checked — changed by 261008-00m; wording updated where
the owner-approved ADR-004 amendment refined a mechanism).

| # | Truth | Status | Evidence at HEAD ae54bf4 |
|---|-------|--------|--------------------------|
| 1 | SC-1: gnome-text-editor double Right Shift corrects `ghbdtn`→`привет`, `GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет` | ✓ VERIFIED | v1 register rows unchanged (R); live 16/16 @419799d + registers re-proven live @9f2fd75 + Oct-5 + the 2026-10-08 matrix run; TestActor_DoubleTapCorrects/TestConvert_CasePreserved PASS at HEAD (this verifier's `-race` run) |
| 2 | SC-2: direction auto-detected both ways + unit corpus | ✓ VERIFIED (mechanism succession) | homogeneous branch `Convert(text, dirOf(script))` intact (runs.go, R); TestActor_DoubleTapCorrects + TestDetect_* corpus green; live both directions @9f2fd75 + Oct-5; mixed ranges per-character inversion per owner verdict / D-55 (truth 7) |
| 3 | SC-3: replacement exactly over the wrong range, both ladder levels, abort-not-litter | ✓ VERIFIED | BuildPlan rune arithmetic intact (plan.go, R); D-27 level-2-only cap pinned (TestActor_Level2NoCaps PASS at HEAD); level-1 live on 5 surfaces @9f2fd75 + Oct-5; live oracle exact final content 2026-10-08; abort discipline green; Chromium actual level owner-accepted (UAT 1) |
| 4 | SC-4: buffer cleared on Enter, Tab, Escape, focus change | ✓ VERIFIED | isResetKeyval table + HardReset unchanged for the real triggers (R); **ADR-004 amendment 2026-10-07 (owner decision, spec-first)** refines the focus trigger: the own flip's synthetic focus_out is not «смена фокуса окна» — flipCredit skips ONLY the HardReset on a consumed credit; real focus loss (no credit), failed round trips and the Reset kind keep the full reset (actor.go HandleLifecycle, D); TestActor_HardResetKeyvals **PASS at HEAD** + TestActor_OwnFlipCreditConsumedOnce / OwnFlipFailedSwitcherArmsNothing PASS; v1 reset rows unchanged and live @419799d; FocusOut reset live @9f2fd75 |
| 5 | SC-5: e2e-matrix v1 green — YAML cases, PASS/FAIL report, exit≠0 | ✓ VERIFIED | Infra at HEAD: strict loader `dec.KnownFields(true)` (matrix.go:367), per-case isolation + exitCode pinned by matrix_test.go (named tests PASS in this verifier's run); v1 file 16 cases, D-13 pin-geometry header intact, re-pinned rows «приветб»/«паиghbdtn» byte-consistent with the unit-pinned ConvertRuns outputs. **Live dimension CLOSED:** 2026-10-08 live run 16/16 PASS on the current code (no code after 4dc4877) — 02-UAT test 3, G-2-3 resolved; the first attempt's 15/16 honestly exposed and fixed two real defects (synthetic-focus reset, stand hermeticity) before the green |
| 6 | Token semantics D-13/D-15 + toggle invariant (D-14 ride-along superseded by owner directive 2026-09-28: table-mapped punctuation converts with the token) | ✓ VERIFIED | buffer corpus green incl. TestBuffer_PushFeed (PASS in this verifier's run); punctuation conversion pinned by TestConvertRuns_HomogeneousWholesale and live-green in the Oct-5 runs + v1 word-punct row 2026-10-08; digits ride (v1 word-digits row, live @419799d) |
| 7 | Mixed word (CORR-04) — per-character inversion (spec-delta D-55; supersedes the D-23 foreign-run succession) | ✓ VERIFIED | `invertPerChar` wired at the mixed branch (runs.go, R); TestConvertRuns_MixedInvertsPerChar + TestActor_MixedWordInvertsPerChar / TestActor_PhraseMixedCorrects / TestActor_SelectionMixedConverts **PASS at HEAD**; unmapped-letter whole-range refusal pinned (TestConvertRuns_Refusals); **live word-mixed row green 2026-10-08 («паиghbdtn»)** — and the flip-across-word live failure (G-2-3) fixed by flipCredit with the whole-word assertion now unit-pinned (TestActor_OwnFlipMixedWordCorrectsWhole PASS) |
| 8 | Conversion preserves register per rune; non-letters per the 2026-09-28 rule (CORR-05) | ✓ VERIFIED | register logic unchanged (convert.go, R); TestConvert_CasePreserved PASS; per-rune register under inversion pinned («Паиghbdtn»); live registers @9f2fd75 + Oct-5 |
| 9 | Suffix verification ADR-004 (token+tail range) | ✓ VERIFIED | MatchesSuffix intact (verify.go, R); TestMatchesSuffix **PASS at HEAD**; wired at pendingVerdict/afterVerdict (actor.go re-read at :2543) |
| 10 | BuildPlan both levels, all counts in runes (CORR-07) | ✓ VERIFIED | len([]rune) arithmetic intact (R); TestBuildPlan_BothLevels **PASS at HEAD** |
| 11 | internal/correct purity: no godbus/time/goroutines | ✓ VERIFIED | purity grep re-run at HEAD: 0 hits in production files (internal/layouts import is a data-table package, inside the purity contract) |
| 12 | Surface drivers: fresh chromium + standalone GTE, witness-gated | ✓ VERIFIED | surface.go, case_surface.go, focus_helper.py untouched since prior pass (R); live gedit/x11/smoke rows @9f2fd75; v4 breadth rows @c84709c |
| 13 | Preflight rejects missing binaries | ✓ VERIFIED | preflight.go untouched (R); live preflight ok in the D-48 log; fresh-session/idle preflight steps observed in the nightly logs |
| 14 | Tracer: double tap runs buffer→direction→convert→verify→delete+commit (CORR-01) | ✓ VERIFIED | TestActor_DoubleTapCorrects **PASS at HEAD**; pipeline wiring re-read at current lines (buf.Token :2247/:2300/:2630 → startRangeCorrection :2745/:2509/:2629/:2720 → ConvertRuns :2672/:2754); live word/phrase cases @9f2fd75 + Oct-5 + 2026-10-08 |
| 15 | Word after space corrected with separator (D-13) | ✓ VERIFIED | TestActor_AfterSpaceCorrects **PASS at HEAD**; live phrase-after-space @9f2fd75; v1 word-after-space row unchanged, live 2026-10-08 |
| 16 | Log discipline D-20/D-21 | ✓ VERIFIED | logCorrectionDone; TestActor_DebugCorrectionRecord present (129 TestActor* enumerated); D-20 pinned by ClipboardRungAfterMismatch |
| 17 | Flip on Single: mode EN↔RU, INFO mode record, XKB untouched | ✓ VERIFIED | flipScript intact; flipTo now additionally arms one credit per successful round trip (D — ADR-004 amendment; mode record, EN↔RU semantics and XKB non-touchment unchanged); TestActor_FlipOnSingle + TestActor_OwnFlip* **PASS at HEAD**; live layout-single @9f2fd75; live flip-across-word flipCredit behavior 2026-10-08 |
| 18 | RU consumption commits Cyrillic; script-true buffer in ALL printing branches incl. digits | ✓ VERIFIED | feedKey invariants (R); TestActor_RUDigitsFullPipeline **PASS at HEAD**; live phrase-ru-en @9f2fd75 |
| 19 | Level 2: ForwardKeyEvent(BackSpace)×(token+tail runes) burst → one commit, in order (CORR-07) | ✓ VERIFIED | executeLevel2 + emitter wire pin (R); TestActor_Level2NoCaps/WithTailAndRunes **PASS at HEAD**; over-cap range refuses (D-27, LevelNone) |
| 20 | Verify-after with epoch: mismatch → INFO counter, never auto-repair; stale timers no-op | ✓ VERIFIED (WR-05 semantics) | settleAfter unchanged (R); HandleLifecycle FocusOut still retires pending rounds even on a consumed credit (only HardReset skipped — context-scoped sides stay, D); TestActor_VerifyAfterLevel1 **PASS at HEAD** |
| 21 | Reset triggers wired live (CORR-09) | ✓ VERIFIED | reset-* live @419799d; FocusOut reset live @9f2fd75; TestActor_HardResetKeyvals PASS at HEAD; ADR-004 amendment refines (not weakens) the focus trigger — real focus loss keeps the full reset, owner decision recorded |
| 22 | Matrix infrastructure: strict decoder, per-case isolation, closed vocabularies | ✓ VERIFIED | matrix.go KnownFields(true) intact at :367; hermeticity predicate added (caseNeedsBaseEstablishment, D) — loader untouched; TestMatrixCaseBaseConfig/NeedsBaseEstablishment PASS in this verifier's run; quiesce invariant pinned by quiesce_test.go |
| 23 | CI circuit D-19: dispatch-only workflow on self-hosted gnome runner | ✓ VERIFIED | e2e-matrix.yml untouched since prior pass (R; zero push/PR triggers, WR-04 env-input kept); pr-sanity green at prior HEAD 1efc6fa (37543084250); live green circuit evidence: 35108412175 (31/31×2), owner Oct-5 runs, **owner live matrix 2026-10-08 16/16**; nightlies soft-skip by design (session age), honestly recorded |

**Score:** 23/23 truths verified (0 present-behavior-unverified)

### Advisory (New Scope, Unevidenced)

New-scope findings from Step 7 with no deterministic evidence — reported,
not blocking. Included because re-verification ran.

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| — | None — the Step 7 scan across every file changed since the prior pass (actor.go, actor_test.go, matrix.go, matrix_test.go, ADR-004/ADR-006, SPEC.md) found no new concern | — | — |

### Required Artifacts

All 70 covered files exist at HEAD (67 carried + 3 added: the 261008-00m
PLAN/SUMMARY and docs/adr/ADR-004-buffer-reset-triggers.md, amended by this
batch). The moved engine/layouts artifacts remain substantive at their
internal/ paths. `verify.artifacts` per plan not re-run — the plan
frontmatter is byte-identical since 42d10a1 and the prior passes' single "~"
false negative stands (the verifier tool cannot see
`~/.config/systemd/user/goswitch-ci-runner.service`, which exists).

### Key Link Verification

Carried: 15/16; the one unverified link remains the plan-spec false negative
(02-01 buffer.go→direction.go, pattern `\.Token\(\)`); the actual consumer
wiring re-read at current HEAD lines: `internal/session/actor.go`
`a.buf.Token()` (:2247/:2300/:2630) → `startRangeCorrection` (:2745) →
`ConvertRuns` (:2754) — present, wired, behavior-tested. The re-pin links
remain wired (runs.go → invertPerChar → both layout tables; case_word.go
oracle ↔ unit pins). New link from 261008-00m verified wired:
`flipTo` success path → `flipCredit++` → `HandleLifecycle` FocusOut branch
consumes exactly one credit → HardReset skipped (actor.go, deep-read; pinned
by the three TestActor_OwnFlip* tests).

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| internal/session/actor.go | token/tail | buf.Token()/Tail() fed by HandleKey from live ProcessKeyEvent | ✓ live corrections @9f2fd75 + Oct-5 + 2026-10-08 | ✓ FLOWING |
| internal/correct/runs.go | out/changed | invertPerChar / Convert over the live range | ✓ live mixed row green 2026-10-08; homogeneous rows live @Oct-5 | ✓ FLOWING |
| internal/correct/plan.go | Plan | BuildPlan from token/tail/converted + caps bit | ✓ live level dispatch @9f2fd75 | ✓ FLOWING |
| test/e2e/matrix.go | report | per-case runMatrixCaseIsolated results (now on pinned base config) | ✓ 31/31×2 @9f2fd75; 16/16 2026-10-08 | ✓ FLOWING |
| internal/layouts/tables.go | ENToRU/RUToEN | xkb-generated (Phase 1) | ✓ live RU-commit path @9f2fd75 | ✓ FLOWING |

### Behavioral Spot-Checks (this verifier, at HEAD ae54bf4)

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Actor pipeline incl. the new credit invariants | `go test -race -count=1 -run 'TestActor_DoubleTapCorrects\|TestActor_MixedWordInvertsPerChar\|TestActor_OwnFlipMixedWordCorrectsWhole\|TestActor_OwnFlipCreditConsumedOnce\|TestActor_OwnFlipFailedSwitcherArmsNothing\|TestActor_AfterSpaceCorrects\|TestActor_VerifyAfterLevel1\|TestActor_FlipOnSingle\|TestActor_HardResetKeyvals\|TestActor_RUDigitsFullPipeline\|TestActor_Level2NoCaps\|TestActor_Level2WithTailAndRunes' ./internal/session` | ok (exit 0) | ✓ PASS |
| Correction core (mixed inversion, wholesale+punctuation, refusals, register, plan, suffix, PushFeed) | `go test -race -count=1 -run 'TestConvert_CasePreserved\|TestConvertRuns_MixedInvertsPerChar\|TestConvertRuns_HomogeneousWholesale\|TestConvertRuns_Refusals\|TestBuildPlan_BothLevels\|TestMatchesSuffix\|TestBuffer_PushFeed' ./internal/correct` | ok (exit 0) | ✓ PASS |
| Matrix hermeticity predicate + strict decode | `go test -race -count=1 -run 'TestMatrixCaseBaseConfig\|TestMatrixCaseNeedsBaseEstablishment\|TestMatrixDecode_ConfigBaseAndFresh$\|TestMatrixDecode_ConfigBaseAndFreshRejections' ./test/e2e` | ok (exit 0) | ✓ PASS |
| Named-test existence proof | `-list` enumerations: 129 TestActor*, 25 correct tests | all pinned tests present incl. the 3 new OwnFlip tests | ✓ PASS |
| Gates | `go build ./...` / `go vet ./...` | OK / OK | ✓ PASS |
| Purity invariant (truth 11) | grep godbus/time./go func in internal/correct non-test | 0 hits | ✓ PASS |
| Disabled-test scan | grep t.Skip across phase test files | 0 (matches are assertion field names only) | ✓ PASS |
| Full `-race` suite / lint / tidy | not re-run (rule: no full suite) | last full green: 261008-00m `mise run ci` iterations at their HEADs (per SUMMARY); pr-sanity green at 1efc6fa | ? SKIP (by rule) |
| Live e2e re-run | not re-run (drives owner's desktop; policy) | live witness RECORDED 2026-10-08 16/16 (UAT test 3); this verifier does not re-drive the desktop | ? SKIP (by policy; witness current — no code after 4dc4877) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` convention; the phase's runnable probes remain
the mise e2e tasks (live-session; policy above) and the CI gate (pr-sanity
green at the prior code HEAD). Not applicable beyond that.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ----------- | ----------- | ------ | -------- |
| CORR-01 | 02-03, 02-04 | Double-tap correction pipeline | ✓ SATISFIED | truths 14, 17; live word/phrase both directions @9f2fd75 + Oct-5 + 2026-10-08 |
| CORR-04 | 02-01, 02-04 | Direction/wrong-layout correction incl. mixed | ✓ SATISFIED | truths 2, 7, 18; mixed per-character inversion live-witnessed 2026-10-08 incl. flip-across-word (flipCredit) |
| CORR-05 | 02-01 | Register preservation | ✓ SATISFIED | truth 8; registers live @9f2fd75 + Oct-5 |
| CORR-07 | 02-01, 02-03, 02-05 | Exact-range, both ladder levels, verify-after | ✓ SATISFIED | truths 3, 9, 10, 15, 19, 20 |
| CORR-09 | 02-05 | Buffer reset triggers | ✓ SATISFIED | truths 4, 21; ADR-004 amendment (owner decision, spec-first) refines the focus trigger with the synthetic-credit mechanism |
| TEST-04 | 02-02, 02-06, 02-07 | e2e matrix + CI circuit | ✓ SATISFIED | truths 5, 12, 22, 23 — infra verified AND the current v1 file live-green 2026-10-08 |

Orphaned requirements: **none** (exactly these 6 IDs mapped Phase 2 in
REQUIREMENTS.md, all Complete).

### Anti-Patterns Found

Debt-marker gate re-run at HEAD across every file changed since the prior
pass (actor.go, actor_test.go, matrix.go, matrix_test.go, ADR-004, ADR-006,
SPEC.md): **ZERO** TBD/FIXME/XXX and zero TODO/HACK/PLACEHOLDER.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| internal/correct/direction.go | — | `Detect` production-orphaned (file untouched since prior pass; zero production callers outside direction.go; TestDetect_* corpus green) | ⚠️ Warning | Carried unchanged from all prior passes. Owner cleanup decision, no defect |
| test/e2e/case_m1.go | 135 | ydotool "Escape" name trap in the unexercised shell-surface fallback (deferred-items #3; file untouched since prior pass) | ⚠️ Warning | Carried unchanged |
| desktop environment | — | persistent AT-SPI shell-bridge wedge + aged-session D48_SKIP nightlies (deferred-items #1; environmental) | ℹ️ Info | Environmental; remedy documented in ci-runner.md; the live witness now exists locally (2026-10-08), reducing dependence on the nightly circuit |
| .planning/.../02-VALIDATION.md | — | unfilled draft template (unchanged) | ℹ️ Info | Process artifact; not gated |

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| internal/correct/{runs,convert,buffer,plan,verify,direction}_test.go | CORR-04/05/07 | 25 | 0 | 0 | value+behavioral (exact outputs, refusals, register) | VALID |
| internal/session/actor_test.go | CORR-01/04/07/09 | 129 | 0 | 0 | behavioral (full pipeline, credit state transitions) | VALID |
| test/e2e/matrix_test.go | TEST-04 | all enumerated | 0 | 0 | value (loader strictness, establishment predicate) | VALID |
| test/e2e/case_word.go + matrix-v1.yaml | TEST-04 | 16 v1 cases | 0 | 0 | behavioral vs AT-SPI oracle (live 2026-10-08) | VALID — oracle values are owner-pinned directives, not system-generated |

**Disabled tests on requirements:** 0 · **Circular patterns:** 0 ·
**Insufficient assertions:** 0

### Decision Coverage

All trackable CONTEXT.md decisions are honored by shipped artifacts (9/9,
gsd-tools check.decision-coverage-verify, this pass). Non-blocking gate; no
findings.

### Human Verification Required

**None — both prior items closed.** Item 1 by recorded live evidence
(2026-10-08, 16/16, UAT test 3 / G-2-3). Item 2 by the closure ruling
recorded above (mechanism identity code-proven + owner's UAT-2 visual
confirmation of that mechanism + live content/range oracle); reopen as a
defect report if live use ever shows a jump on mixed text.

### Gaps Summary

**No gaps.** The re-verification chain is fully converged: the 2026-10-07
report's single open item (the live witness for the two re-pinned v1 rows)
was resolved the honest way — the first live attempt FAILED 15/16, exposing
a real defect (the daemon's own flip hard-resetting the correction buffer
via a synthetic focus transition) and a stand-hermeticity defect (plain
cases adopting the owner's live config); both were fixed spec-first with
RED→GREEN TDD in quick 261008-00m, and the final live run on the fixed tree
went 16/16 with word-mixed settling to «паиghbdtn» exactly as re-pinned. No
code has changed since that witness. All 23 truths verified at HEAD; every
previously verified truth regression-checked untouched or deep-checked
(flipCredit, hermeticity); named `-race` tests, build, vet, purity, debt
scan, decision coverage (9/9) all green. The perceptual mixed-text item is
closed by the composition ruling recorded above — with its residual stated
plainly rather than absorbed silently. No overrides carried or needed;
nothing deferred (no later-phase dependency remains for phase-2 scope).

---

_Verified: 2026-10-07T22:51:15Z_
_Verifier: Claude (gsd-verifier) — human-item closure re-verification (live witness 2026-10-08 + composition ruling)_
