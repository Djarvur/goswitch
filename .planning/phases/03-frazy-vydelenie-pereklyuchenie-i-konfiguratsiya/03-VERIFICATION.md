---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
verified: 2026-10-07T20:56:10Z
status: human_needed
score: 19/20 must-haves verified
covered_files:

  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-01-PLAN.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-01-SUMMARY.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-02-PLAN.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-02-SUMMARY.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-03-PLAN.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-03-SUMMARY.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-04-PLAN.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-04-SUMMARY.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-05-PLAN.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-05-SUMMARY.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-06-PLAN.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-06-SUMMARY.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-07-PLAN.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-07-SUMMARY.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-CONTEXT.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-REVIEW.md
  - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-REVIEW-FIX.md
  - .github/workflows/e2e-matrix.yml
  - cmd/goswitchctl/main.go
  - cmd/goswitchd/main.go
  - docs/CONFIG.md
  - internal/engine/engine.go
  - internal/engine/engine_test.go
  - internal/engine/keys.go
  - go.mod
  - go.sum
  - internal/appid/appid.go
  - internal/appid/appid_test.go
  - internal/clipboard/clipboard.go
  - internal/clipboard/clipboard_test.go
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/config/load.go
  - internal/config/load_test.go
  - internal/config/watch.go
  - internal/config/watch_test.go
  - internal/correct/buffer.go
  - internal/correct/buffer_test.go
  - internal/correct/plan.go
  - internal/correct/plan_test.go
  - internal/correct/runs.go
  - internal/correct/runs_test.go
  - internal/correct/verify.go
  - internal/correct/verify_test.go
  - internal/ctlsvc/ctlsvc.go
  - internal/ctlsvc/ctlsvc_test.go
  - internal/hotkey/fsm.go
  - internal/hotkey/fsm_test.go
  - internal/hotkey/names.go
  - internal/hotkey/names_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - test/e2e/case_combo.go
  - test/e2e/case_ctl.go
  - test/e2e/case_m1.go
  - test/e2e/case_macr.go
  - test/e2e/case_phrase.go
  - test/e2e/case_resilience.go
  - test/e2e/case_select.go
  - test/e2e/case_word.go
  - test/e2e/cases/matrix-v1.yaml
  - test/e2e/cases/matrix-v2.yaml
  - test/e2e/cases/matrix-v3.yaml
  - test/e2e/focus_helper.py
  - test/e2e/main.go
  - test/e2e/matrix.go
  - test/e2e/matrix_test.go
  - test/e2e/preflight.go
  - test/e2e/quiesce_test.go
  - test/e2e/README.md

covered_digest: "v1:sha256:9dc643d242510d79711ce1dcd77ca7cbf9be3227cc886b52af87219884590ac6"
behavior_unverified: 1
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 20/20
  previous_verified: 2026-09-20T14:32:40Z
  reason: "Milestone-close (v1.0.0) fingerprint refresh. Since the previous pass (HEAD 04d00ad) the tree advanced through Phases 5-8 AND two quick tasks that changed phase-3-owned production code: (1) 261006-vqw — mixed-text semantics revised to per-character layout inversion as THE mixed-text semantics (owner verdict 2026-10-03, commit 7dd46e9; D-55 spec-delta a24afa9 committed before code; runs.go invertPerChar 4da90d8; golden corpus re-pinned RED-first; 13 matrix rows v1..v4 re-pinned), which SUPERSEDES the previously recorded D-22/D-23 run semantics; (2) 261007-0yg — engine/ and layouts/ moved under internal/ (5427d7a, SPEC §8 rev), making the old covered paths engine/*.go, layouts/*.go STALE — all paths re-derived against the current tree (internal/engine/, internal/layouts/; layouts/* files left coverage — never phase-3-covered). Also switchTimeout 40→150 ms (actor.go const, documented SetGlobalEngine wedge guard, G-6-1 — not a YAML param, truth 11 unaffected). The previous passes' 'production tree bit-identical to the green matrix runs' argument is therefore DEAD: the daemon code AND the matrix expectations changed. 19/20 truths re-verified on the current tree (pinned symbols + wiring + named behavioral tests re-run green under -race by this verifier); truth 16 (green full-breadth matrix AT THE VERIFIED HEAD) drops to PRESENT_BEHAVIOR_UNVERIFIED — no green live matrix run exists at the revised tree; the owner live-matrix acceptance is the pending quick-task 261006-vqw human gate (STATE.md: 'owner live-matrix + gedit check pending'). No code regression: every re-pinned expectation is byte-identical to unit-pinned outputs, and the changed behavior is owner-directed."
  gaps_closed: []
  gaps_remaining: []
  regressions: []
behavior_unverified_items:
  - truth: "SC-5/truth 16: e2e-matrix green in full breadth at the verified HEAD — the live-desktop proof of the whole gesture vocabulary"
    test: "Owner live run at the current tree: mise run e2e-matrix-v2 (minimum; v3/v4 for full breadth) per the pending quick-task 261006-vqw gates"
    expected: "21/21 v2 PASS with the re-pinned mixed rows: phrase-mixed 'привет ghbdtn', word-mixed 'паиghbdtn', select-all-gte/chromium 'привет ghbdtn'; all homogeneous/degradation/selection-geometry rows unchanged-green"
    why_human: "Drives the owner's live GNOME desktop (IBus + real GTK/chromium/zenity surfaces); cannot run headless. The unit/session behavior IS test-proven at this tree (named corpus green under -race); what no recorded run exercises is the live full-breadth matrix at the revised semantics"
human_verification:
  - test: "Owner live-matrix re-run at the revised tree: mise run e2e-matrix-v2 (and ideally -v3/-v4), plus the hand-typed mixed-text gedit/GTK4 check"
    expected: "All 21 v2 cases PASS with the per-character-inversion expectations ('паиghbdtn', 'привет ghbdtn'); homogeneous/refusal/geometry rows unchanged-green; every letter inverts on the hand-typed check"
    why_human: "Live-desktop acceptance act reserved to the owner — the same 5 pending gates declared by quick task 261006-vqw (its VERIFICATION.md frontmatter), which this item consolidates for phase 3; the nightly D-48 CI gate will re-exercise it automatically only after the revision branches merge to main"
---

# Phase 3: Фразы, выделение, переключение и конфигурация — Verification Report

**Phase Goal (ROADMAP.md):** «Полный словарь жестов — тройной тап исправляет фразу, двойной при выделении исправляет выделенное, смешанный текст корректируется по ADR-семантике; одиночный Right Shift переключает раскладку механизмом по ADR; все хоткеи и таймауты настраиваются YAML-конфигом с hot reload и управляются через `goswitchctl` — всё подтверждает зелёная e2e-матрица v2 полной широты.»
**Verified:** 2026-10-07T20:56:10Z
**Status:** human_needed
**Re-verification:** Yes — milestone-close fingerprint refresh (v1.0.0, HEAD 2171fae, branch gsd/internalize-engine-layouts)

## Re-verification Scope and Method

The phase passed at close (2026-09-15, human_needed with 4 items, all later
resolved) and was re-verified passed twice on later trees (42d10a1, then
04d00ad on 2026-09-20 — both times with the daemon production tree
bit-identical to the green matrix evidence). Since 04d00ad the tree advanced
through Phases 5–8 AND two quick tasks that — for the first time — changed
phase-3-owned production code. This pass re-derived every path and re-checked
every must-have on the current tree (HEAD 2171fae).

**What actually changed in phase-3-owned files since 04d00ad:**

1. **Mixed-text semantics revision (quick 261006-vqw, owner verdict
   2026-10-03, commit 7dd46e9; D-55 spec-delta a24afa9 committed before
   code).** Per-character layout inversion is now THE mixed-text semantics,
   superseding D-22/D-23 (the WINDOWS-12 readback freeze was rejected by the
   owner). `internal/correct/runs.go`: `invertPerChar` (:118-140) in the
   mixed branch (:97-98); the homogeneous wholesale branch byte-identical;
   `internal/session/actor.go` wiring re-pinned; golden corpus re-pinned
   RED-first (independently reproduced at d7a8a5e by the quick task's
   verifier); 13 e2e matrix rows across v1..v4 re-pinned; case_word.go
   oracle re-pinned. SPEC §4.2 amended with the dated audit trail
   (docs/SPEC.md:88-117); README matches.
2. **Package internalization (quick 261007-0yg, commit 5427d7a, SPEC §8
   rev).** `engine/` → `internal/engine/`, `layouts/` → `internal/layouts/`
   — pure renames plus import-path lines. Old covered paths
   engine/engine.go, engine/engine_test.go, engine/keys.go are STALE; this
   pass re-derives them as internal/engine/{engine.go,engine_test.go,keys.go}
   (all exist; engine.go/keys.go content unchanged by the move).
3. **switchTimeout 40→150 ms** (internal/session/actor.go:63) — the
   SetGlobalEngine round-trip WEDGE GUARD, not a YAML parameter: the
   measured live flip RTT is ~41-45 ms, so the old 40 ms deadline fired
   before completion on every healthy flip (05-05 live proofs; G-6-1).
   Phase-3 truth 11 scopes YAML to the user-facing tap timeouts/keys
   (SC-4's «таймауты тапов») — unaffected.
4. **Phases 5–8 additive surfaces** on the shared files: flipTo unified
   execution path with the D-52 switcher seam (actor.go:1954), indicator/
   menu/sound/a11y observer seams, autocorrect (06), blocklist (07),
   config/ctlsvc/config.md additive sections, matrix-v4.yaml (34 cases),
   workflow default input matrix-v4 + v4 routes.

**Method.** Full checks where paths/wiring/semantics changed (runs.go,
actor.go, matrix rows, ctlsvc/config/CONFIG.md/names.go diffs reviewed);
quick regression (pinned-symbol greps + named behavioral tests) for the
surviving truths. Named tests re-run green under `-race` BY THIS VERIFIER at
the current tree: TestConvertRuns_MixedInvertsPerChar,
TestConvertRuns_HomogeneousWholesale, TestConvertRuns_Refusals,
TestBuildPlan_BackspaceCap, TestSelectionRange, TestVerifyRangeAt,
TestActor_TripleTapCorrectsPhrase, TestActor_DoubleTapSelectionCorrects,
TestActor_MixedWordInvertsPerChar, TestActor_PhraseMixedCorrects,
TestActor_SelectionMixedConverts, TestActor_FlipOnSingle,
TestActor_ComboWordThenFlip, TestActor_NoSelectionStillWord,
TestActor_HotReloadInvalidKeepsLastGood, TestActor_HotReloadTapKey,
TestActor_ClipboardRungOffKeyPath, TestActor_MACRIntercepts,
TestSvc_NameGuard, TestSvc_Status, TestSvc_RecoverShim,
TestRenderStatusVersionToken, TestLoad_UnknownKeyRejected,
TestFSM_ConfigurableTapKey, TestFSM_ModifierUse,
TestParseBinding_KeyvalProvenance, TestFamilyMask; the clipboard and engine
wire packages green. Full suite NOT re-run per re-verification policy (the
quick task's `mise run ci` — build+vet+lint+`go test -race ./...`, 19
packages — was green at f9ed497, and 5427d7a/2171fae are docs/moves on top).

**Evidence-currency finding (the reason this pass is human_needed).** The
green live-matrix evidence chain ends BEFORE the semantics revision: the
last real full live run is CI 37274809977 (success, 24m46s) at 780463e —
the phase-6 merge, which does NOT contain the revision
(`git merge-base --is-ancestor f9ed497 780463e` → false). The scheduled
runs of 2026-10-03..10-07 are 20–30 s soft-skips on `main` (the revision
branches are NOT merged to main; `f9ed497` not an ancestor of origin/main),
so they exercise neither the new code nor the full matrix. The owner
live-matrix acceptance is recorded PENDING in STATE.md (quick 261006-vqw,
status «Needs Review», «owner live-matrix + gedit check pending») with the
5 declared live gates in the quick task's VERIFICATION.md. Truth 16 is
therefore ⚠️ PRESENT_BEHAVIOR_UNVERIFIED at this tree — present (21 v2
cases on disk, rows byte-identical to the unit-pinned outputs, runner and
routes wired) and wired, with the unit/session behavior test-proven green,
but with no recorded live full-breadth run at the revised semantics. This
is revision-acceptance currency, not a phase-3 code regression.

## Verification Approach Note (MVP mode discrepancy)

Unchanged from all prior passes: ROADMAP declares `**Mode:** mvp` but the
goal is not a canonical User Story (`user-story.validate` → `valid: false`,
checked at initial verification); standard goal-backward verification
against the 5 Success Criteria + the 7 plans' must_haves applies.

## Stale-criteria notes (working interpretations, owner-preauthorized)

- **D-34/D-35 (carried forward):** SC-3's literal «святой порядок: свитч на
  первый тап» and «индикатор GNOME актуален» wordings remain STALE against
  accepted ADR-001/ADR-002; verified via the working interpretation (native
  Super mechanism not broken; mode-record oracle). Owner-preauthorized
  (D-34, BIND), not an override.
- **D-55 (NEW this pass):** SC-2's «конвертируется только часть, набранная
  не той раскладкой» is STALE against the accepted spec-delta 2026-10-06
  (owner verdict 2026-10-03, commit 7dd46e9; WINDOWS-12 freeze REJECTED):
  in a mixed range EVERY letter now inverts per character. The SC's
  operative clauses — «по семантике ADR» (the ADR semantics, as revised
  through the project's spec-delta mechanism) + «золотой корпус юнит-тестов
  плюс e2e-кейсы» — hold: SPEC §4.2 amended with the audit trail, golden
  corpus green, e2e cases re-pinned (live green pending — truth 16).
  Verified against the revised semantics, not an override.

## User Flow Coverage

User capability (goal, non-story form): «полный словарь жестов коррекции и
переключения, настраиваемый YAML-ом и управляемый goswitchctl». Re-checked
on the current tree; mixed rows now carry the revised expectations.

| Step | Expected | Evidence | Status |
|------|----------|----------|--------|
| Type a wrong-layout (homogeneous) phrase, triple-tap Right Shift | whole phrase converts wholesale `ghbdtn ghbdtn`→`привет привет` | actor.go:1023-1024 Triple→startPhraseCorrection (:2667) over Phrase()/ReplacePhrase; TestActor_TripleTapCorrectsPhrase green this pass; 8 phrase rows in matrix-v2 (21 cases on disk) | ✓ |
| Select wrong text, double-tap | exactly the selection converts, both geometries, 3 surfaces | SelectionRange/VerifyRangeAt (verify.go:24/:41); selection branch + selectionState in actor.go; TestActor_DoubleTapSelectionCorrects/SelectionLeftToRight corpus green; select rows re-pinned where mixed (`привет ghbdtn`) | ✓ |
| Type mixed-script text, tap | every letter inverts per character (`gfbпривет`→`паиghbdtn`); digits/space ride | invertPerChar (runs.go:118-140); TestConvertRuns_MixedInvertsPerChar + TestActor_MixedWordInvertsPerChar/PhraseMixedCorrects/SelectionMixedConverts green this pass | ✓ |
| Single Right Shift | mode flips EN↔RU both ways, engine-truth flip via D-52 seam; oracle = mode log record | flipTo (actor.go:1954) — mode record first, then SetGlobalEngine seam (150 ms wedge guard), then observers; TestActor_FlipOnSingle green; grep gsettings internal/session/ → 1 hit, a comment (:1067), zero calls | ✓ |
| Shift+RightCtrl combo | word corrected FIRST, then mode flips | settleCombo at every pipeline terminal (actor.go:1047/:1432/:1165); TestActor_ComboWordThenFlip green; correction-flip SET semantics (owner rule 2026-09-28) additive | ✓ |
| Edit the YAML config | applied without restart; typo → last-good + visible error | watch.go unchanged (atomic snapshot, Reload, «config reload rejected»); TestActor_HotReloadInvalidKeepsLastGood/TapKey green | ✓ |
| goswitchctl status / reload / correct | live answers, --json, exit≠0 contract, D-32 visibility | ctlsvc.go: ErrNotPrimaryOwner (:40), recoverMethod on all 3 methods (:100/:111/:129), leading `version=` (:144); additive autocorrect/sound tokens + OnConn hook; TestSvc_NameGuard/Status/RecoverShim green | ✓ |
| Super+letter in configured apps | intercepted → Ctrl+letter forwarded (wire-proven) | TestActor_MACR* corpus green; macr row in v2/v3; GTK4 client boundary = documented platform limitation (WINDOWS #4 closed) | ✓ |
| Outcome: green matrix proves the lot | green runs recorded AT THE VERIFIED HEAD | unit/session behavior green at HEAD; live full-breadth green ends at the PRE-revision trees (54ea460 21/21 v2; CI 35108412175 31/31 v3 ×2; 37274809977 phase-6 tree); owner live re-run pending (quick 261006-vqw gates) | ⚠️ present, behavior-unverified |

## Goal Achievement

### Observable Truths

The 20 consolidated must-haves of the previous passes, re-verified against
the CURRENT tree (HEAD 2171fae) with the revised mixed-text semantics.

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC-1: Triple Right Shift corrects the whole (homogeneous) phrase since the last hard reset (CORR-02, D-25) — `ghbdtn ghbdtn`→`привет привет` | ✓ VERIFIED | actor.go:1023-1024/:2667 startPhraseCorrection over buffer.go Phrase()/ReplacePhrase; TestActor_TripleTapCorrectsPhrase/EmptyBuffer/PhraseVerifyMismatch green this pass (-race); 8 phrase rows in matrix-v2, carried in v3 (32 cases ⊇ v2) |
| 2 | SC-1: Double Right Shift under an active selection corrects exactly the selected range, both geometric directions (CORR-03, D-30, Pitfall 6) | ✓ VERIFIED | SelectionRange/VerifyRangeAt (verify.go:24/:41) unchanged; selection branch + selectionState intact; TestActor_DoubleTapSelectionCorrects green this pass (byte-identical pin per the revision audit); 5 select rows, mixed ones re-pinned to `привет ghbdtn` |
| 3 | Clipboard round-trip rung opt-in (default OFF), order after verify-mismatch, best-effort restore, stdin-only content (D-28/D-29, ADR-003 rung C) | ✓ VERIFIED | clipboard package unchanged since last pass; internal/clipboard green under -race this pass; TestActor_ClipboardRungDisabledByDefault/AfterMismatch/OffKeyPath (OffKeyPath green this pass) |
| 4 | SC-2 (revised semantics, D-55): mixed text converts by PER-CHARACTER layout inversion — every Latin letter → ENToRU, every Cyrillic → RUToEN, each independent of position; digits/space pass through (their inversion equals themselves / no key position); mixed-range punctuation rides as typed (per-char direction undefined); an unmapped letter fails the WHOLE range with nil output (T-03-01-03) | ✓ VERIFIED | invertPerChar (runs.go:118-140) wired at runs.go:97-98 — the ONE conversion entry for word (:2620), phrase (:2702) and selection ranges; TestConvertRuns_MixedInvertsPerChar green this pass; TestActor_MixedWordInvertsPerChar/PhraseMixedCorrects/SelectionMixedConverts green; SPEC §4.2:102-117 states the rule with the both-tables ',' argument |
| 5 | Homogeneous (single-script) ranges convert WHOLESALE byte-identically — punctuation with the token by position (owner directive 2026-09-28: `ghbdtn,`→`приветб`); mixed always changed=true; D-24 changed=false stays the defensive no-op surface; the old D-16→D-23 succession example is superseded (`gfbпривет`→`паиghbdtn`, matrix rows re-pinned) | ✓ VERIFIED | Convert(text, dirOf(script)) wholesale branch (runs.go:100) — byte-identical per the revision audit (TestConvertRuns_Refusals, TestActor_DoubleTapSelectionCorrects byte-identical); TestConvertRuns_HomogeneousWholesale green this pass with the `приветб` pin; D-24 `!changed` branches intact in actor.go; old `паипривет` wording STALE per D-55 (see stale-criteria note) |
| 6 | Backspace cap D-27: default 50, level 2 not built over cap, zero deletions on refusal (reason backspace-cap) | ✓ VERIFIED | plan.go:30 DefaultBackspaceCap=50; actor.go:2835 skipCorrection("backspace-cap"); backspaceCap() resolver (:1906); TestBuildPlan_BackspaceCap green this pass |
| 7 | SC-3: single Right Shift flips mode EN↔RU; oracle = the byte-stable mode log record; the flip now executes through the ADR-006 engine-truth path (D-52): internal mode switches immediately, SetGlobalEngine leaves through the switcher seam under the 150 ms wedge guard, nil-seam degrades to the internal flip with one WARN — never gsettings (SWCH-01, D-34/ADR-001) | ✓ VERIFIED | flipTo (actor.go:1954-1998) — the single execution path of EVERY flip gesture; mode record FIRST (the e2e oracle, Pitfall 5), then switcher, then observers; TestActor_FlipOnSingle green this pass; grep gsettings internal/session/ → 1 comment hit (:1067), zero calls; SetSwitcher seam (:836) with generation-scoped degradation |
| 8 | SC-3/SWCH-04: tap-timing conflict resolved per ADR-002; window + tap key configurable; tap_key and verify_wait_ms consumed (CR-01/CR-02) | ✓ VERIFIED | fsm.go unchanged (NewFSM(window, tapKeyval), SetTapKey); applySnapshot still feeds verifyWait/tapKeyval; TestFSM_ConfigurableTapKey + TestActor_HotReloadTapKey green this pass |
| 9 | SC-3/SWCH-02: combo Shift+RightCtrl corrects the word THEN flips (D-36 order), binding renavigable, series Reset, buffer not fed (Pitfall 4); correction flips use SET semantics naming the target from the corrected script (owner rule 2026-09-28, additive) | ✓ VERIFIED | settleCombo at every terminal (actor.go:1047/:1432; settleCombo :1165; settleCorrectionFlip :1190); TestActor_ComboWordThenFlip/EmptyBufferStillFlips/ConfigurableBinding/DoesNotFeedBuffer corpus present, ComboWordThenFlip green this pass; TestFSM_ModifierUse green |
| 10 | SC-3/SWCH-03 (working interpretation, D-34 pre-authorized): native Super mechanism not broken — daemon keeps its IBus name and corrects after the shell chord | ✓ VERIFIED | super-space-alive row live PASS ×4 at acceptance, carried in v2→v3 (name present in both corpora on disk); literal MRU/indicator wording STALE per D-34, not required; chord support formalized by KeyvalSpace (names.go, additive) |
| 11 | SC-4/CONF-01: all hotkeys/tap-timeouts/params in YAML (hotkeys/timeouts/correction/macr); strict KnownFields decode, named-field errors, range ceilings (D-31/D-33) | ✓ VERIFIED | load.go:28 KnownFields(true); TestLoad_UnknownKeyRejected green this pass; config.go grew +254 lines of later-phase sections under the same strict decode; docs/CONFIG.md complete reference incl. new sections (grew +388 lines); switchTimeout is a hardcoded wedge-guard const, deliberately not a YAML key (see scope note) |
| 12 | SC-4/CONF-02: hot reload without restart — dir-watch + debounce, last-good + WARN, one snapshot per event (D-32, Pitfall 8) | ✓ VERIFIED | watch.go unchanged since last pass (atomic snapshot, Reload, rejection path); TestActor_HotReloadInvalidKeepsLastGood/WindowNewSeries/OptionsAndCombo corpus present, the first two named tests green this pass |
| 13 | SC-4/CONF-03: Caramba correspondence table in docs, no name cloning, model-difference caveat (D-31) | ✓ VERIFIED | docs/CONFIG.md:293+ — the table and the «не клонирует ни имена… ни модель» caveat intact after the +388-line update |
| 14 | SC-4/INST-02: goswitchctl status(--json)/reload/correct against a live daemon; primary-owner name guard; recover shim per method; status carries machine state only, never user text (D-32 visible) | ✓ VERIFIED | ctlsvc.go: ErrNotPrimaryOwner (:40, wired into serve :203/:225), recoverMethod on all 3 methods (:100/:111/:129, shim :189), leading `version=` token (:144); ADDITIVE this period: autocorrect/sound state tokens + ac_skip_ slugs + OnConn hook (WARN-and-continue) — counts-and-states-only invariant intact; TestSvc_NameGuard/Status/ReloadConfig/CorrectNow/RecoverShim green this pass (NameGuard over a real private dbus-daemon) |
| 15 | SC-5/MACR-01: implemented per ADR-005 — interception above combo/mode branches (RU pin), burst on Super release, consumed-upstream WARN + counters, per-app a11y observer with degradation ladder, alt_modifier default-off | ✓ VERIFIED | macrPendingKeyval/macrIntercepted intact (actor.go:203-204/:1724); MACRStats (:593); TestActor_MACRIntercepts green this pass; MACR-01 Complete with the GTK4-Wayland forwarded-events platform caveat (owner, REQUIREMENTS.md:45/:108) |
| 16 | SC-5: e2e-matrix green in full breadth (21 v2 cases; v3 ⊇ v2; clipboard row excluded with the documented spike verdict) — **at the verified HEAD** | ⚠️ PRESENT_BEHAVIOR_UNVERIFIED | Breadth present and wired: matrix-v2.yaml still 21 cases, v3 32 (superset re-verified by name-diff, comm empty), v4 34 (later-phase); the 13 mixed rows re-pinned BYTE-IDENTICAL to the unit-pinned ConvertRuns outputs («паиghbdtn» ×4, «привет ghbdtn» ×9); runner/routes/double-run intact in the workflow; unit+session behavior green at HEAD (named corpus this pass; full -race suite green at f9ed497 per the quick task's mise run ci). NOT proven at this tree: a green LIVE full-breadth run — the recorded green chain (54ea460 21/21; CI 35108412175 31/31 ×2; CI 37274809977 at the phase-6 merge) all PREDATES the semantics revision; post-revision scheduled CI runs are 20-30 s soft-skips on main, which does not contain the revision; the owner live-matrix gate is pending (quick 261006-vqw human gates, STATE.md «Needs Review»). See Human Verification |
| 17 | anchorPos seam carries the anchor to the actor on every surface; per-surface actuals spike-pinned (Pitfall 1; A1/A2/A5) | ✓ VERIFIED | internal/engine/engine.go:88-90 HandleSurroundingText(text, cursorPos, anchorPos) + wire parse :196-201 (pure-rename survivors); verify.go SelectionRange; zenity D-30 degradation verdict stands (owner-acknowledged); spike verdicts re-confirmed in STATE.md |
| 18 | D-30 continuity: no selection → the word path unchanged; Phase-2 corpus green without expectation edits | ✓ VERIFIED | TestActor_NoSelectionStillWord present and green this pass; v1 matrix 16 cases on disk (its two mixed/punctuation rows re-pinned per the revision — deliberate, owner-directed, covered by truth 5) |
| 19 | Key/binding provenance: closed name tables from ibuskeysyms.h; KeyV=0x076, KeyControlR=0xffe4, KeySuperL/R; ParseBinding + FamilyMask press-side rule | ✓ VERIFIED | internal/engine/keys.go (rename survivor) + names.go — additive KeyvalSpace=0x020 with provenance cite (:376) and the documented no-family-mask rationale; TestParseBinding_KeyvalProvenance + TestFamilyMask green this pass |
| 20 | CI circuit: workflow routes the full-breadth matrix by input (v1 AND v2 paths preserved); first green run recorded; mise task registry complete | ✓ VERIFIED (wording superseded) | Workflow re-checked at HEAD: input routes v1/v2/v3/v4 preserved in BOTH run steps (:202-226, 8 mise invocations — the D-48 double-run structure intact), default now matrix-v4; mise.toml carries 46 e2e-* task lines; recorded green runs exist (v2 21/21 at 54ea460; v3 31/31 ×2 at CI 35108412175; v3 full at 37274809977) — all PRE-revision; live re-validation at the revised tree is the pending owner gate + the nightly D-48 schedule once merged |

**Score:** 19/20 truths verified (1 present, behavior-unverified — truth 16)

Behavior-dependent truths note: every state-transition/invariant truth in
this phase (phrase/selection pipelines, per-character inversion, verify-after
debounce, hot-reload window arming, combo settlement fan-out, MACR release
burst, clipboard kill-vs-empty, name guard over the wire, flip record order)
is covered by a named behavioral test re-run GREEN under -race at the current
HEAD by this verifier. The live-desktop dimension of truth 16 is the single
open item.

### Decision Coverage

`check.decision-coverage-verify` re-run this pass on 03-CONTEXT.md:
**15/15 decisions honored**, none missing (non-blocking gate).

### Advisory (New Scope, Unevidenced)

Re-verification ran; the Step 7 scan over all covered files modified since
the previous verification found ZERO debt markers (TBD/FIXME/XXX with no
follow-up reference) and zero stub patterns. Recorded currency findings
(evidence-backed, none a phase-3 code regression):

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| 1 | Post-revision scheduled e2e-matrix CI runs (37102882019/37183775834/37429931098/37585613271, 2026-10-03..10-07) complete in 20-30 s on `main` — soft-skips, not live matrix runs; `main` does not contain the revision (f9ed497 not an ancestor) | other (CI evidence currency) | Deterministic evidence both ways (gh run list durations; merge-base check); the nightly D-48 gate will exercise the revision automatically once the branches merge; the owner gate is the live path — tracked as truth 16's human item, not a gap |
| 2 | The last full live matrix success (CI 37274809977, 24m46s, success) ran at 780463e — the phase-6 merge tree, pre-revision | other (evidence currency) | Same finding family as #1; recorded so the evidence chain stays auditable |
| 3 | .golangci.yml exhaustruct deprecation (previous Advisory 2) has EVOLVED: the disable entry now names both `exhaustruct` and `exhaustruct_v5` (:130-131) | other (tooling config) | Prior advisory partially addressed on the current tree; belongs to the repo-wide quality gate, not a phase-3 truth |
| 4 | grep scoping: `gsettings` now has 1 hit inside internal/session — a COMMENT (actor.go:1067, the SWCH regression-guard note); the switch path remains call-free | other (scope note) | Contract intact; recorded so future greps scope correctly |

### Required Artifacts

All 71 covered files exist on the current tree (existence check this pass
before fingerprinting; the three engine files re-derived under internal/ per
the 5427d7a move; layouts/ files were never phase-3-covered and stay out).
Substantiveness re-confirmed by grep of every pinned symbol (per-truth
evidence above); zero stub markers; `go build ./...` exit 0 at HEAD.

### Key Link Verification

Re-checked on the current tree:

- cmd/goswitchd/main.go → ctlsvc.Run (:216) and → actor.AttachConfig (:509) — wired.
- internal/ctlsvc/ctlsvc.go → session.Status (StatusSnapshotProvider :142) — direction intact.
- internal/session/actor.go → correct.ConvertRuns (:2620 word, :2702 phrase/selection parameterized) — wired; runs.go → internal/layouts tables (ENToRU/RUToEN :124/:126) — wired.
- cmd/goswitchctl/main.go bus name/path constants (org.djarvur.goswitch, :30-31) — intact and matched by ctlsvc.
- test/e2e/matrix.go kind vocabulary parses matrix-v2.yaml (21 cases) — the re-pinned expect_text values ride the same strict schema; matrix-v4 parsed by the same driver (v4 route live in the workflow).

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| internal/session/actor.go | correction ranges | buf.Phrase()/activeRange + selectionState from live SetSurroundingText pushes | ✓ select rows green in recorded matrix runs; unit pipeline green at HEAD | ✓ FLOWING |
| internal/session/actor.go | config snapshot | Watcher.Snapshot() over atomic.Pointer (fsnotify dir events / ctl Reload) | ✓ reload rows live in recorded runs; HotReload tests green at HEAD | ✓ FLOWING |
| internal/correct/runs.go | conversion output | layouts.ENToRU/RUToEN table lookups per rune | ✓ unit-pinned outputs byte-identical to the re-pinned matrix expectations | ✓ FLOWING |
| internal/session/actor.go | flip engine name | engineNameOf(mode) → SetGlobalEngine seam (D-52) | ✓ TestActor_FlipOnSingle green; live flip proofs 05-05 | ✓ FLOWING |
| internal/ctlsvc/ctlsvc.go | status line | renderStatus ← StatusSnapshot (leading version=) + watcher LastError/ConfigPath | ✓ corrections_done observed live at acceptance; additive tokens test-pinned | ✓ FLOWING |
| test/e2e/matrix.go | report | per-case runMatrixCaseIsolated results over real daemon+surfaces | ✓ 21/21 + 31/31 ×2 recorded (pre-revision trees); wired at HEAD | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Build | `go build ./...` | exit 0 | ✓ PASS |
| Mixed inversion (truths 4/5) | `go test ./internal/correct/ -run 'TestConvertRuns_MixedInvertsPerChar\|TestConvertRuns_HomogeneousWholesale\|TestConvertRuns_Refusals\|TestBuildPlan_BackspaceCap\|TestSelectionRange\|TestVerifyRangeAt' -race -count=1` | ok 1.0s | ✓ PASS |
| Core gestures at HEAD (truths 1,2,7,8,9,12,15,18 + mixed pipeline) | `go test ./internal/session/ -run 'TestActor_TripleTapCorrectsPhrase\|TestActor_DoubleTapSelectionCorrects\|TestActor_MixedWordInvertsPerChar\|TestActor_PhraseMixedCorrects\|TestActor_SelectionMixedConverts\|TestActor_FlipOnSingle\|TestActor_ComboWordThenFlip\|TestActor_NoSelectionStillWord\|TestActor_HotReloadInvalidKeepsLastGood\|TestActor_HotReloadTapKey\|TestActor_ClipboardRungOffKeyPath\|TestActor_MACRIntercepts' -race -count=1` | ok 1.4s | ✓ PASS |
| Control service (truth 14) | `go test ./internal/ctlsvc/ -run 'TestSvc_NameGuard\|TestSvc_Status\|TestSvc_RecoverShim\|TestRenderStatusVersionToken' -race -count=1` | ok 1.1s | ✓ PASS |
| Strict decode (truth 11) | `go test ./internal/config/ -run TestLoad_UnknownKeyRejected -race -count=1` | ok | ✓ PASS |
| FSM/tap + provenance (truths 8, 19) | `go test ./internal/hotkey/ -run 'TestFSM_ConfigurableTapKey\|TestFSM_ModifierUse\|TestParseBinding_KeyvalProvenance\|TestFamilyMask' -race -count=1` | ok | ✓ PASS |
| Clipboard + engine wire (truths 3, 17) | `go test ./internal/clipboard/ -race` and `go test ./internal/engine/ -run 'TestParseBindings\|TestWire'` | both ok | ✓ PASS |
| Matrix breadth preserved | case-name counts + v2⊂v3 name diff | v1=16, v2=21, v3=32 (v2 superset: comm empty), v4=34 | ✓ PASS |
| CI matrix runs (recorded, re-fetched) | `gh run list --workflow e2e-matrix.yml` | latest success 37585613271 (2026-10-07, schedule, ~25 s — soft-skip on main); last full live success 37274809977 (24m46s) at the PRE-revision phase-6 merge 780463e | ✓ PASS (with currency note — see truth 16) |
| Live e2e re-run | not re-run (drives owner's desktop) | pending owner gate (quick 261006-vqw) | ? SKIP (by policy) → human verification |

### Probe Execution

Step 7c: no `scripts/*/tests/probe-*.sh` convention in this repo. The
phase-declared probes remain the mise e2e matrix tasks (live-session —
SKIP by the recorded-evidence policy; their static encoding of the revised
semantics is verified and their live PASS is the owner's pending
acceptance act) and the CI gate. Headless re-checks this pass: build +
named -race corpora, all green; the full headless suite was last run green
by the quick task's `mise run ci` at f9ed497 (19 packages), with
5427d7a/2171fae being a pure move + docs on top.

### Test Quality Audit

Unchanged in character from prior passes: requirement-linked corpora
all-active, non-circular, value/behavioral-level. This pass re-confirmed:
the golden corpus is now split into MixedInvertsPerChar /
HomogeneousWholesale / Refusals (the revision's RED-first re-pin,
independently reproduced at the RED commit d7a8a5e by the quick task's
verifier — the failing set was exactly the four re-pinned tests); expected
values remain hand-authored literals (приветб/паиghbdtn/привет ghbdtn);
matrix oracles rune-exact. Disabled tests on requirements: **0** (the
TestSvc_NameGuard dbus env-guard still executed — dbus-daemon present).
Circular patterns: **0**.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| CORR-02 | 03-01, 03-07 | Triple Right Shift corrects the whole buffer phrase | ✓ SATISFIED | truths 1, 6, 16 |
| CORR-03 | 03-03, 03-07 | Double Right Shift under selection corrects the selected text | ✓ SATISFIED | truths 2, 3, 17 |
| CORR-06 | 03-01, 03-07 | Mixed text corrects per the ADR semantics («точная семантика — ADR фазы решения») — semantics revised to per-character inversion via D-55 spec-delta | ✓ SATISFIED | truths 4, 5 (revised semantics, SPEC §4.2 amendment, golden corpus green; live e2e pending — truth 16) |
| SWCH-01 | 03-04, 03-07 | Right Shift toggles layout via the ADR mechanism | ✓ SATISFIED | truth 7 (now the ADR-006 engine-truth flip per D-52) |
| SWCH-02 | 03-04, 03-07 | Combo «correct word + switch» (default Shift+RightCtrl) | ✓ SATISFIED | truth 9 |
| SWCH-03 | 03-04, 03-07 | Native Super+Space mechanism not broken (D-34 working interpretation) | ✓ SATISFIED | truth 10 |
| SWCH-04 | 03-02, 03-04 | Tap-timing conflict resolved per ADR; configurable | ✓ SATISFIED | truth 8 |
| CONF-01 | 03-02 | All keys/tap-timeouts/params in YAML | ✓ SATISFIED | truth 11 |
| CONF-02 | 03-02, 03-04, 03-07 | Hot reload without restart | ✓ SATISFIED | truth 12 |
| CONF-03 | 03-02 | Caramba compatibility (wish) via docs table | ✓ SATISFIED | truth 13 |
| MACR-01 | 03-05, 03-07 | Super+letter → Ctrl+letter per ADR-005 | ✓ SATISFIED (Complete with GTK4 caveat — owner decision, UAT item 3) | truth 15; REQUIREMENTS.md:45/:108 |
| INST-02 | 03-06, 03-07 | goswitchctl status/reload/correct | ✓ SATISFIED | truth 14 |

Orphaned requirements: **none** — REQUIREMENTS.md re-grepped: all 12 IDs
present and `[x]` (29 checked boxes total), traceability rows Complete.

### Anti-Patterns Found

Debt-marker gate this pass over ALL covered files modified since the
previous verification (25-file scan incl. actor.go, runs.go, config.go,
ctlsvc.go, names.go, engine.go, keys.go, CONFIG.md, all changed e2e files,
workflow, mise.toml): **ZERO** unreferenced TBD/FIXME/XXX; zero
TODO/HACK/PLACEHOLDER; zero stub markers.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| internal/session/actor.go | logCorrectionDone :2855 | source/result words at DEBUG (review IN-01) | ℹ️ Info | INFO record content-free («outcome=done»); D-21-permitted debug channel, -debug flag warns — unchanged contract |
| internal/engine/engine.go | caps :109 | historically write-only field (review IN-02) | ℹ️ Info | Move survivor; :238 comment documents the capability funnel — carried as review Info debt |
| internal/hotkey/names.go | ParseBinding | degenerate bindings accepted (review IN-04) | ℹ️ Info | Unchanged (the space addition is orthogonal); harmless (WR-02 closed the tap_key side) |
| test/e2e/README.md | — | -case table may lag the v4 registry (review IN-03 family) | ℹ️ Info | Docs-only; matrix.yaml files remain the source of truth |

The 8 in-scope review findings (CR-01..03, WR-01..05) remain FIXED with
RED→GREEN pins; their named tests green in this pass's runs.

Prohibition audit: the prior 15 judgment-tier verdicts carry forward and
were re-confirmed programmatically this pass where the files changed — no
autocorrection without a gesture (startCorrection only from FSM
Double/Triple expiry, combo branch, CorrectNow); INFO content-free;
KnownFields strict (TestLoad_UnknownKeyRejected green); last-good + WARN
(HotReloadInvalidKeepsLastGood green); clipboard content stdin-only, never
logged/argv; selection exactness pinned (DoubleTapSelectionCorrects green);
zero gsettings CALLS in the switch path (1 comment hit only);
TestFSM_ModifierUse green; consumed-upstream WARN + counters;
status surface counts/states/version only (additive tokens are counts and
booleans; no user text). The formerly live-only judgments (teardown
restores the desktop; matrix never red from stand fault; session-bus
same-uid isolation) rest on the recorded green runs — all of which predate
the revision; their re-establishment at the revised tree is part of the
pending owner live gate.

### Human Verification Required

**1 item (new this pass — the only open item in the phase):**

1. **Owner live-matrix re-run at the revised tree** — run `mise run
   e2e-matrix-v2` (minimum; `-v3`/`-v4` for full breadth) on the live GNOME
   desktop, plus the hand-typed mixed-text check in gedit/GTK4 (type a
   Latin word, flip, type a Cyrillic word, correct). Expected: 21/21 v2
   PASS with the per-character-inversion rows (`паиghbdtn`, `привет
   ghbdtn`); homogeneous/degradation/geometry rows unchanged-green; every
   letter inverts by hand. Why human: drives the owner's live desktop;
   the recorded green chain predates the 2026-10-06 semantics revision and
   post-revision CI runs are soft-skips on a main without the revision.
   This consolidates the 5 pending gates declared by quick task
   261006-vqw (its VERIFICATION.md frontmatter) for the phase-3 contract.

All 4 items from the initial verification remain RESOLVED (03-UAT.md,
status complete: matrix v2 21/21 at 54ea460 — now the pre-revision anchor;
WINDOWS #2/#4; zenity verdict).

### Gaps Summary

**No gaps.** All artifacts present and wired on the current tree; 19/20
must-haves verified with first-hand behavioral evidence at HEAD 2171fae;
zero debt markers; build green; decision coverage 15/15. The phase's
production surface absorbed phases 5–8 and two quick tasks cleanly: the
mixed-text semantics revision (D-55) is owner-directed, spec-delta-first,
RED-first, and consistently pinned from unit corpus through e2e rows; the
package internalization is a pure rename this report's paths now reflect.
The single open item is evidence currency, not code: truth 16's live
full-breadth green proof has not yet been re-established at the revised
semantics — the owner live-matrix gate (quick 261006-vqw, STATE.md
«Needs Review») is the wired path, and the nightly D-48 CI gate takes over
once the branches merge to main. Status: **human_needed** (1 item).

---

_Verified: 2026-10-07T20:56:10Z_
_Verifier: Claude (gsd-verifier)_
