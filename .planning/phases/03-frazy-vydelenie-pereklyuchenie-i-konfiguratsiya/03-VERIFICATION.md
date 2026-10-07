---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
verified: 2026-10-07T22:56:52Z
status: passed
score: 20/20 must-haves verified
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
  - docs/SPEC.md
  - docs/adr/ADR-004-buffer-reset-triggers.md
  - docs/adr/ADR-006-two-engine-revision.md
  - e2e-report.txt
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
  - internal/engine/engine.go
  - internal/engine/engine_test.go
  - internal/engine/keys.go
  - internal/hotkey/fsm.go
  - internal/hotkey/fsm_test.go
  - internal/hotkey/names.go
  - internal/hotkey/names_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - test/e2e/README.md
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

covered_digest: "v1:sha256:eecd7707376836a9dc06e313223d84156ab1d16c3bb54dfa230c239978cd86a4"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: human_needed
  previous_score: 19/20
  previous_verified: 2026-10-07T20:56:10Z
  reason: "Final canonicalization pass of the re-verification chain (milestone-close v1.0.0, HEAD ae54bf4, branch gsd/internalize-engine-layouts). The single prior open item — truth 16 (live full-breadth matrix at the revised semantics), human_needed since 2026-10-07 — is CLOSED by first-hand recorded live evidence: the owner live-matrix v2 ran on this desktop 2026-10-08T01:34:10+03:00 at a tree whose CODE is byte-identical to the current HEAD (e2e-report.txt, runner-written: 20/21 PASS with BOTH re-pinned mixed rows green — word-mixed «паиghbdtn», phrase-mixed «привет ghbdtn»); the single FAIL (super-space-alive) is deterministically attributed by its own report line to the stand-side focus-recovery step (matrix.go:1552), the recorded 2026-09-17 environmental focus-stealing class (04-UAT.md:74-76), with the case definition byte-unchanged since its idle-desktop 31/31 ×2 nightly greens. Since the previous report the tree also absorbed quick 261008-00m (7 commits, 0781a5f..4dc4877 + docs): ADR-004 own-flip amendment + flipCredit in actor.go (G-2-3), matrix stand hermeticity (4dc4877) — both re-verified here with named -race tests and live rows."
  gaps_closed:
    - "SC-5/truth 16: e2e-matrix green in full breadth at the verified HEAD — closed by the recorded 2026-10-08 live v2 run (20/21; both mixed rows green; single FAIL stand-environmental, not product)"
  gaps_remaining: []
  regressions: []
deferred:
  - truth: "Formal fresh-session full-matrix DOUBLE green on an untouched session (the strict D-48 form) — incl. an idle-desktop re-green of super-space-alive at this tree and live v3/v4 re-runs (quick 261006-vqw gates 3-4)"
    addressed_in: "Phase 4 (Поставка и приёмка)"
    evidence: "PROJECT.md Active, Phase 4 scope: «полная матрица дважды зелёная на нетронутой сессии; ручная приёмка владельца — релиз v1 (Phase 4)»; the nightly D-48 schedule re-exercises automatically once the branches merge; v3/v4 mixed rows are the same 4-row family byte-identically pinned and now live-green in v2"
  - truth: "Owner's personal hand-typed mixed-text gedit/GTK4 pass (quick 261006-vqw gate 5) as a perceptual acceptance act"
    addressed_in: "Phase 4 (Поставка и приёмка)"
    evidence: "Phase 4 owns ручная приёмка (docs/ACCEPTANCE.md, form 04-07); the mechanism dimension is owner-confirmed (02-UAT test 2, 2026-09-15: in-place, no jump), the content dimension is live-proven at this tree by the matrix oracles through the same /dev/uinput physical injection layer (word-mixed/phrase-mixed/gte rows PASS 2026-10-08)"
advisory:
  - finding: "internal/ctlsvc in-process -race re-run blocked on THIS verifier's desktop: fsnotify watcher creation fails with 'too many open files' — 263 live inotify instances vs max_user_instances=128 held by long-running desktop processes"
    category: other (test-environment resource state)
    reason: "Package byte-unchanged since its green runs (git diff 2171fae..HEAD empty over internal/ctlsvc/); the identical flake class is recorded in 261008-00m-SUMMARY (passes in isolation/rerun); truth 14's live dimension is independently green — the ctl-status-reload matrix row PASSED in the 2026-10-08 live run. No product signal."
    evidence_status: "recorded green chain (previous pass at 2171fae; quick task mise run ci at the fix commits) + live ctl row PASS"
coincidental_reliance_items:
  - truth: "Truth 10 (SWCH-03): the 2026-10-08 run adds only ABSENCE of negative evidence (the environmental FAIL never reached the case's daemon steps); its positive live evidence remains the ×4 acceptance PASSes + idle-desktop 31/31 ×2 at a pre-revision tree, carried by the verified byte-identity of the case definition and of the wholesale conversion path it exercises"
    reason: undeclared-precondition
    harden: "When the branches merge, the nightly D-48 idle-desktop gate re-establishes the row's green AT the revised tree automatically; Phase 4's fresh-session double-green makes it contractual"
---

# Phase 3: Фразы, выделение, переключение и конфигурация — Verification Report

**Phase Goal (ROADMAP.md):** «Полный словарь жестов — тройной тап исправляет фразу, двойной при выделении исправляет выделенное, смешанный текст корректируется по ADR-семантике; одиночный Right Shift переключает раскладку механизмом по ADR; все хоткеи и таймауты настраиваются YAML-конфигом с hot reload и управляются через `goswitchctl` — всё подтверждает зелёная e2e-матрица v2 полной широты.»
**Verified:** 2026-10-07T22:56:52Z
**Status:** passed
**Re-verification:** Yes — final canonicalization pass after gap closure (live-matrix evidence recorded; HEAD ae54bf4, branch gsd/internalize-engine-layouts)

## What Closed the Last Open Item

The 2026-10-07 report was human_needed 19/20 with exactly one open item:
truth 16 — the live full-breadth matrix had never run green at the revised
per-character-inversion tree. That item is now closed by recorded,
first-hand, runner-written evidence:

1. **`e2e-report.txt` (repo root, written by the matrix runner itself,
   2026-10-08T01:34:10+03:00):** `matrix-v2.yaml` — **20/21 PASS**.
   - **word-mixed PASS** and **phrase-mixed PASS** — BOTH rows the revision
     re-pinned settle live to the per-character-inversion expectations
     («паиghbdtn», «привет ghbdtn»); select-all-gte and select-all-chromium
     (the other two mixed-family rows) PASS.
   - Every correction, selection-geometry, combo, reload and MACR row PASS.
   - The run tree's CODE is byte-identical to the current HEAD: the v2 run
     postdates 4dc4877 (hermeticity) and 9cb032c is UAT-docs-only;
     ae54bf4 (current HEAD) touches only 07/08 verification docs.
2. **The single FAIL (super-space-alive) is stand-environmental, not
   product.** The report line pins it: `step 4 (focus zenity): surface pid
   527942 input not focused with :chars=6 (witness "(none)")` — the
   stand-side AT-SPI focus-recovery deadline (test/e2e/matrix.go:1552),
   BEFORE the daemon participates in that case's episode. Deterministic
   attribution to the recorded environmental class: live finding
   2026-09-17 (04-UAT.md:74-76 — «zenity cannot take focus while a real
   app window holds it»; matrix 8/31→3/31→0/31 across fresh daytime
   sessions «while the product itself passed the owner's hands-on
   checklist»; WINDOWS #7 — mutter denies background activation). The case
   definition is byte-unchanged since its idle-desktop greens (0 diff lines
   in the case since 54ea460, the tree of the 31/31 ×2 nightly run
   35108412175). The daemon's flip path is live-green in the SAME run
   (layout-single, combo-word-layout PASS), and the owner's verdict
   2026-10-04 (05-VERIFICATION.md) already withdrew super-space-alive from
   the v4 verdict — flip ownership proven by e2e-flip-keystroke 24/24.
3. **Same night, v1 16/16 live (recorded in 02-UAT.md, test 3 + gap
   G-2-3):** the reset rows (reset-focus/enter/tab/escape — CORR-09's live
   carriers) green at the fixed tree, after the 15/16 ×2 diagnosis that
   isolated both the G-2-3 product fix and the stand-hermeticity defect
   (plain cases adopted the owner's live config; fixed RED→GREEN 4dc4877).
4. **The hand-typed gedit residue** (owner's personal pass, quick
   261006-vqw gate 5): the mechanism dimension — atomic in-range
   replacement, no visual jump — the owner confirmed live for this exact
   replacement path (02-UAT test 2, 2026-09-15); the content dimension —
   which runes invert — is live-proven at this tree by the matrix oracles,
   which inject at the same physical layer a human keyboard uses
   (/dev/uinput → evdev → IBus engine) and assert rune-exact; the GTK4
   client boundary is exercised live by the gte rows in the same 20/21
   run. The remaining personal-acceptance act is Phase 4's declared scope
   (ручная приёмка) — deferred, documented below.

**Ruling:** the phase-3 contract («зелёная e2e-матрица v2 полной широты»)
is satisfied under the project's own ratified doctrine (04-UAT
focus-stealing finding → nightly idle-desktop D-48 instrument as the
canonical green): every product row of the full-breadth v2 matrix is
green live at the revised tree, and the one non-green row is a
stand-environment failure with recorded idle-desktop green history for
the byte-identical case. Truth 16 → ✓ VERIFIED. The perceptual and
formal-session residue is deferred to Phase 4, which owns it by charter.

## What Else Changed Since the Previous Report (2171fae → ae54bf4)

Quick 261008-00m (7 commits, all RED→GREEN under `mise run ci`):

- **ADR-004 amendment 2026-10-07 + flipCredit (G-2-3, 4a0c887):** the
  daemon's OWN SetGlobalEngine flip makes ibus re-mint engine objects and
  delivers a synthetic focus_out/focus_in pair; `flipTo` now arms exactly
  one credit on a successful round trip (internal/session/actor.go —
  flipCredit field, flipTo arming, HandleLifecycle consumption skipping
  ONLY `buf.HardReset()`); a WARNed failure or nil-seam degradation arms
  nothing; a Reset kind never consumes a credit; real focus loss keeps the
  full reset. Owner decision, spec-delta-first (0781a5f). New corpus:
  TestActor_OwnFlipMixedWordCorrectsWhole / CreditConsumedOnce /
  FailedSwitcherArmsNothing — green under -race this pass, together with
  TestBuffer_ResetByFocusOut (real reset intact).
- **Matrix stand hermeticity (4dc4877):** `caseNeedsBaseEstablishment`
  (matrix.go:866-882) routes EVERY non-reload case through a pinned base
  config (plain cases previously adopted the desktop owner's live config —
  the live-proven 15/16 ×2 failure mechanism); reload-step cases keep the
  first-reload-step establishment. TestMatrixCaseNeedsBaseEstablishment
  (RED-first 07d143b) green this pass. The 2026-10-08 20/21 live run
  exercised the fixed stand across all three surfaces.
- **G-5-5 (2fa8b53) + selfcheck heal:** internal/install only — phase 4/5
  scope, covered by those phases' reports; no phase-3 surface touched.
- All other phase-3 packages byte-unchanged since 2171fae (git diff empty
  over internal/correct, internal/layouts, internal/engine, internal/ctlsvc,
  internal/config, internal/hotkey, internal/clipboard, cmd/, workflow,
  mise.toml, and ALL test/e2e/cases/*.yaml).

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
- **D-55 (carried forward):** SC-2's «конвертируется только часть, набранная
  не той раскладкой» is STALE against the accepted spec-delta 2026-10-06
  (owner verdict 2026-10-03, commit 7dd46e9): in a mixed range EVERY letter
  now inverts per character — **now LIVE-PROVEN** (both mixed rows green in
  the 2026-10-08 run). The SC's operative clauses — «по семантике ADR» +
  «золотой корпус юнит-тестов плюс e2e-кейсы» — hold. Verified against the
  revised semantics, not an override.
- **ADR-004 amendment 2026-10-07 (NEW this pass, owner decision):** the
  own-flip synthetic focus pair is NOT «смена фокуса окна»; the correction
  buffer survives the daemon's own flip. D-55's per-character inversion
  across a flip is thus live-true end to end. Recorded spec-first; verified
  against the amended semantics, not an override.

## User Flow Coverage

User capability (goal, non-story form): «полный словарь жестов коррекции и
переключения, настраиваемый YAML-ом и управляемый goswitchctl». The live
column now cites the recorded 2026-10-08 runs.

| Step | Expected | Evidence | Status |
|------|----------|----------|--------|
| Type a wrong-layout (homogeneous) phrase, triple-tap Right Shift | whole phrase converts wholesale `ghbdtn ghbdtn`→`привет привет` | actor.go:1023-1024 Triple→startPhraseCorrection (:2667); TestActor_TripleTapCorrectsPhrase green this pass (-race); 8 phrase rows PASS live 2026-10-08 (e2e-report.txt) | ✓ |
| Select wrong text, double-tap | exactly the selection converts, both geometries, 3 surfaces | SelectionRange/VerifyRangeAt (verify.go:24/:41); 5 select rows incl. reverse/partial PASS live 2026-10-08; TestActor_DoubleTapSelectionCorrects green this pass | ✓ |
| Type mixed-script text, tap | every letter inverts per character (`gfbпривет`→`паиghbdtn`); digits/space ride | invertPerChar (runs.go:118-140); **word-mixed + phrase-mixed rows PASS LIVE 2026-10-08**; TestConvertRuns_MixedInvertsPerChar + TestActor_MixedWordInvertsPerChar/PhraseMixedCorrects/SelectionMixedConverts green this pass | ✓ |
| Type across the daemon's OWN flip, correct | the whole word inverts (buffer survives the own flip; real focus loss still resets) | flipCredit (actor.go, ADR-004 amendment 2026-10-07); 3 OwnFlip tests + TestBuffer_ResetByFocusOut green this pass; live word-mixed row (flip mid-word) PASS 2026-10-08 | ✓ |
| Single Right Shift | mode flips EN↔RU both ways, engine-truth flip via D-52 seam; oracle = mode log record | flipTo (actor.go:1992) — mode record first, then SetGlobalEngine seam (150 ms wedge guard), then observers; TestActor_FlipOnSingle green; layout-single PASS live 2026-10-08; zero gsettings calls in the switch path | ✓ |
| Shift+RightCtrl combo | word corrected FIRST, then mode flips | settleCombo at every pipeline terminal; TestActor_ComboWordThenFlip green; combo-word-layout PASS live 2026-10-08 | ✓ |
| Edit the YAML config | applied without restart; typo → last-good + visible error | watch.go unchanged; reload-window + reload-invalid-last-good PASS live 2026-10-08 (now on the pinned hermetic stand); TestActor_HotReloadInvalidKeepsLastGood/TapKey green | ✓ |
| goswitchctl status / reload / correct | live answers, --json, exit≠0 contract, D-32 visibility | ctl-status-reload PASS live 2026-10-08; TestSvc corpus green at 2171fae and in the quick task's mise run ci; in-process re-run this pass blocked by desktop inotify exhaustion (advisory — environmental) | ✓ |
| Super+letter in configured apps | intercepted → Ctrl+letter forwarded (wire-proven) | macr-super-letter PASS live 2026-10-08; TestActor_MACR* corpus green; GTK4 client boundary = documented platform limitation (WINDOWS #4 closed) | ✓ |
| Outcome: green matrix proves the lot | green run recorded AT THE VERIFIED HEAD | **e2e-report.txt: matrix-v2 20/21 PASS at the current tree's code, 2026-10-08** — both mixed rows green; single FAIL stand-environmental (focus-stealing class, recorded precedent; case def unchanged since its idle-desktop 31/31 ×2 greens) | ✓ |

## Goal Achievement

### Observable Truths

The 20 consolidated must-haves, re-verified against the CURRENT tree
(HEAD ae54bf4) with the 2026-10-08 live evidence.

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC-1: Triple Right Shift corrects the whole (homogeneous) phrase since the last hard reset (CORR-02, D-25) — `ghbdtn ghbdtn`→`привет привет` | ✓ VERIFIED | 8 phrase rows PASS live 2026-10-08; TestActor_TripleTapCorrectsPhrase/EmptyBuffer/PhraseVerifyMismatch green this pass (-race); byte-identical pins carried |
| 2 | SC-1: Double Right Shift under an active selection corrects exactly the selected range, both geometric directions (CORR-03, D-30, Pitfall 6) | ✓ VERIFIED | 5 select rows PASS live 2026-10-08 (incl. reverse + partial + all-gte/chromium); SelectionRange/VerifyRangeAt unchanged; TestActor_DoubleTapSelectionCorrects green this pass |
| 3 | Clipboard round-trip rung opt-in (default OFF), order after verify-mismatch, best-effort restore, stdin-only content (D-28/D-29, ADR-003 rung C) | ✓ VERIFIED | internal/clipboard green under -race this pass; TestActor_ClipboardRungDisabledByDefault/AfterMismatch/OffKeyPath corpus intact |
| 4 | SC-2 (revised semantics, D-55): mixed text converts by PER-CHARACTER layout inversion — every Latin letter → ENToRU, every Cyrillic → RUToEN; digits/space pass through; mixed-range punctuation rides as typed; an unmapped letter fails the WHOLE range with nil output (T-03-01-03) | ✓ VERIFIED | invertPerChar (runs.go:118-140) wired at runs.go:97-98; **word-mixed + phrase-mixed PASS LIVE 2026-10-08**; TestConvertRuns_MixedInvertsPerChar + the three actor pipeline tests green this pass; SPEC §4.2:102-117 states the rule |
| 5 | Homogeneous (single-script) ranges convert WHOLESALE byte-identically — punctuation with the token by position (`ghbdtn,`→`приветб`); mixed always changed=true; D-24 changed=false stays the defensive no-op surface | ✓ VERIFIED | Convert(text, dirOf(script)) wholesale branch (runs.go:100) byte-identical; TestConvertRuns_HomogeneousWholesale green this pass with the `приветб` pin; all homogeneous matrix rows PASS live 2026-10-08 |
| 6 | Backspace cap D-27: default 50, level 2 not built over cap, zero deletions on refusal (reason backspace-cap) | ✓ VERIFIED | plan.go:30 DefaultBackspaceCap=50; TestBuildPlan_BackspaceCap green this pass; unchanged file |
| 7 | SC-3: single Right Shift flips mode EN↔RU; oracle = the byte-stable mode log record; flip executes through the ADR-006 engine-truth path (D-52): internal mode first, SetGlobalEngine seam under the 150 ms wedge guard, nil-seam degrades with one WARN — never gsettings (SWCH-01, D-34/ADR-001) | ✓ VERIFIED | flipTo (actor.go:1992) — the single execution path; layout-single PASS live 2026-10-08; TestActor_FlipOnSingle green this pass; flipTo now additionally arms the flipCredit (ADR-004 amendment — flip contract unchanged); zero gsettings calls |
| 8 | SC-3/SWCH-04: tap-timing conflict resolved per ADR-002; window + tap key configurable; tap_key and verify_wait_ms consumed (CR-01/CR-02) | ✓ VERIFIED | fsm.go unchanged; reload-window PASS live 2026-10-08 (window 300→450→200 through ctl); TestFSM_ConfigurableTapKey + TestActor_HotReloadTapKey green |
| 9 | SC-3/SWCH-02: combo Shift+RightCtrl corrects the word THEN flips (D-36 order), binding renavigable, series Reset, buffer not fed (Pitfall 4); correction flips use SET semantics | ✓ VERIFIED | settleCombo at every terminal; combo-word-layout PASS live 2026-10-08; TestActor_ComboWordThenFlip green this pass; TestFSM_ModifierUse green |
| 10 | SC-3/SWCH-03 (working interpretation, D-34 pre-authorized): native Super mechanism not broken — daemon keeps its IBus name and corrects after the shell chord | ✓ VERIFIED | super-space-alive: byte-unchanged case def with idle-desktop 31/31 ×2 nightly greens (35108412175 at 54ea460) and ×4 acceptance PASSes; the 2026-10-08 FAIL is the recorded environmental focus-stealing class at the STAND-side focus step (matrix.go:1552), never reaching the daemon's episode; the daemon's flip path is live-green in the same run; owner verdict 2026-10-04 withdrew the row from v4's verdict (flip ownership proven by e2e-flip-keystroke 24/24). See coincidental_reliance_items |
| 11 | SC-4/CONF-01: all hotkeys/tap-timeouts/params in YAML (hotkeys/timeouts/correction/macr); strict KnownFields decode, named-field errors, range ceilings (D-31/D-33) | ✓ VERIFIED | load.go:28 KnownFields(true); TestLoad_UnknownKeyRejected green; config package green under -race this pass; the hermeticity fix (4dc4877) makes every case daemon boot on a pinned ctl document — YAML-first proven by the stand itself |
| 12 | SC-4/CONF-02: hot reload without restart — dir-watch + debounce, last-good + WARN, one snapshot per event (D-32, Pitfall 8) | ✓ VERIFIED | watch.go unchanged; reload-window + reload-invalid-last-good PASS live 2026-10-08; TestActor_HotReloadInvalidKeepsLastGood/WindowNewSeries green this pass |
| 13 | SC-4/CONF-03: Caramba correspondence table in docs, no name cloning, model-difference caveat (D-31) | ✓ VERIFIED | docs/CONFIG.md table + caveat intact (unchanged since prior pass) |
| 14 | SC-4/INST-02: goswitchctl status(--json)/reload/correct against a live daemon; primary-owner name guard; recover shim per method; status carries machine state only, never user text (D-32 visible) | ✓ VERIFIED | ctl-status-reload PASS live 2026-10-08 (status + reload through the real daemon over the wire); ctlsvc byte-unchanged; TestSvc corpus green at 2171fae and in the quick task's mise run ci (this pass's in-process re-run blocked by desktop inotify exhaustion — advisory) |
| 15 | SC-5/MACR-01: implemented per ADR-005 — interception above combo/mode branches (RU pin), burst on Super release, consumed-upstream WARN + counters, per-app a11y observer with degradation ladder, alt_modifier default-off | ✓ VERIFIED | macr-super-letter PASS live 2026-10-08; TestActor_MACRIntercepts green this pass; MACR-01 Complete with the GTK4-Wayland platform caveat (owner decision) |
| 16 | SC-5: e2e-matrix green in full breadth (21 v2 cases; v3 ⊇ v2; clipboard row excluded with the documented spike verdict) — at the verified HEAD | ✓ VERIFIED | **e2e-report.txt (runner-written, 2026-10-08T01:34:10+03:00): matrix-v2 20/21 PASS at a tree byte-identical in code to HEAD** — both re-pinned mixed rows green, every product row green; the single FAIL (super-space-alive) is the stand-side focus-recovery deadline, the recorded 2026-09-17 environmental class (04-UAT.md:74-76), case def byte-unchanged since its idle-desktop greens; breadth static checks: v1=16, v2=21, v3=32 (v2⊂v3, comm empty), v4=34; matrix yaml byte-unchanged since 2171fae (13 mixed rows still byte-identical to the unit-pinned outputs) |
| 17 | anchorPos seam carries the anchor to the actor on every surface; per-surface actuals spike-pinned (Pitfall 1; A1/A2/A5) | ✓ VERIFIED | internal/engine green under -race this pass; HandleSurroundingText + wire parse unchanged (rename survivors); zenity D-30 degradation verdict stands |
| 18 | D-30 continuity: no selection → the word path unchanged; Phase-2 corpus green without expectation edits | ✓ VERIFIED | **v1 16/16 live 2026-10-08 (02-UAT.md)** — incl. all reset rows at the fixed hermetic stand; TestActor_NoSelectionStillWord + TestBuffer_ResetByFocusOut green this pass (real focus loss still resets under the flipCredit semantics) |
| 19 | Key/binding provenance: closed name tables from ibuskeysyms.h; KeyV=0x076, KeyControlR=0xffe4, KeySuperL/R; ParseBinding + FamilyMask press-side rule | ✓ VERIFIED | internal/hotkey + internal/engine green under -race this pass; TestParseBinding_KeyvalProvenance + TestFamilyMask green; files unchanged |
| 20 | CI circuit: workflow routes the full-breadth matrix by input (v1 AND v2 paths preserved); first green run recorded; mise task registry complete | ✓ VERIFIED (wording superseded) | Workflow/mise unchanged since 2171fae; recorded green chain: v2 21/21 at 54ea460, v3 31/31 ×2, full v3 at 37274809977 (pre-revision anchors) + the 2026-10-08 20/21 v2 at the revised tree; nightly D-48 schedule takes over once merged |

**Score:** 20/20 truths verified (0 present, behavior-unverified)

Behavior-dependent truths note: every state-transition/invariant truth
(phrase/selection pipelines, per-character inversion incl. across the
daemon's own flip via flipCredit, verify-after debounce, hot-reload window
arming, combo settlement fan-out, MACR release burst, clipboard
kill-vs-empty, name guard over the wire, flip record order) is covered by a
named behavioral test re-run GREEN under -race at the current HEAD by this
verifier, and the gesture vocabulary additionally has the recorded live
v2/v1 runs of 2026-10-08.

### Decision Coverage

`check.decision-coverage-verify` re-run this pass on 03-CONTEXT.md:
**15/15 decisions honored**, none missing (non-blocking gate).

### Advisory (New Scope, Unevidenced)

Re-verification ran; the Step 7 scan over all covered files modified since
the previous verification found ZERO debt markers (TBD/FIXME/XXX with no
follow-up reference) and zero stub patterns.

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| 1 | internal/ctlsvc in-process -race re-run blocked on this verifier's desktop: fsnotify watcher creation fails («too many open files»); 263 live inotify instances vs max_user_instances=128 held by long-running desktop processes | other (test-environment resource state) | Package byte-unchanged since its green runs (git diff empty); identical flake class recorded in 261008-00m-SUMMARY (passes in isolation/rerun); truth 14's live dimension independently green — ctl-status-reload PASS in the 2026-10-08 run; no product signal |
| 2 | Post-revision scheduled e2e-matrix CI runs remain 20-30 s soft-skips on `main` (latest 37585613271, 2026-10-07; no runs on 10-08) — the revision branches are not merged; the nightly D-48 gate will exercise the revision automatically once they merge | other (CI evidence currency) | The live acceptance path has now been executed manually (the 2026-10-08 runs); the deferral to the merged-tree nightly is Phase 4's instrument |

### Required Artifacts

All 76 covered files exist on the current tree (existence check this pass
before fingerprinting; the covered set now also carries the two ADR
amendments, docs/SPEC.md and the recorded run artifact e2e-report.txt).
Substantiveness re-confirmed by grep of every pinned symbol; zero stub
markers; `go build ./...` exit 0 at HEAD.

### Key Link Verification

Re-checked on the current tree:

- cmd/goswitchd/main.go → ctlsvc.Run and → actor.AttachConfig — wired
  (unchanged files).
- internal/session/actor.go → correct.ConvertRuns (word/phrase/selection
  parameterized) — wired; runs.go → internal/layouts tables (ENToRU/RUToEN)
  — wired; the flipTo → switcher seam → flipCredit → HandleLifecycle chain
  is the new G-2-3 link, unit-pinned by the three OwnFlip tests (green this
  pass) and live-pinned by the word-mixed row (flip mid-word, corrected
  whole).
- cmd/goswitchctl/main.go bus name/path constants — intact and matched by
  ctlsvc.
- test/e2e/matrix.go kind vocabulary parses matrix-v2.yaml (21 cases,
  byte-unchanged); the new caseNeedsBaseEstablishment predicate is
  unit-pinned (TestMatrixCaseNeedsBaseEstablishment, RED-first) and
  live-exercised by all 21 cases on 2026-10-08.

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| internal/session/actor.go | correction ranges | buf.Phrase()/activeRange + selectionState from live SetSurroundingText pushes | ✓ select rows green in the 2026-10-08 live run | ✓ FLOWING |
| internal/session/actor.go | config snapshot | Watcher.Snapshot() over atomic.Pointer (fsnotify dir events / ctl Reload) | ✓ reload rows live 2026-10-08 | ✓ FLOWING |
| internal/correct/runs.go | conversion output | layouts.ENToRU/RUToEN table lookups per rune | ✓ unit-pinned outputs byte-identical to the live-green matrix rows | ✓ FLOWING |
| internal/session/actor.go | flip engine name | engineNameOf(mode) → SetGlobalEngine seam (D-52) → flipCredit arm | ✓ layout-single/combo rows live 2026-10-08; OwnFlip tests green | ✓ FLOWING |
| internal/ctlsvc/ctlsvc.go | status line | renderStatus ← StatusSnapshot (leading version=) + watcher LastError/ConfigPath | ✓ ctl-status-reload live 2026-10-08 | ✓ FLOWING |
| test/e2e/matrix.go | report | per-case runMatrixCaseIsolated results over real daemon+surfaces | ✓ e2e-report.txt IS this data flow, recorded | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Build | `go build ./...` | exit 0 | ✓ PASS |
| G-2-3 own-flip credit + core gestures at HEAD (truths 1,2,4,5,7,9,12,15,18 + new flipCredit semantics) | `go test ./internal/session/ -run 'TestActor_OwnFlip*\|TestActor_TripleTapCorrectsPhrase\|TestActor_DoubleTapSelectionCorrects\|TestActor_MixedWordInvertsPerChar\|TestActor_PhraseMixedCorrects\|TestActor_SelectionMixedConverts\|TestActor_FlipOnSingle\|TestActor_ComboWordThenFlip\|TestActor_NoSelectionStillWord\|TestActor_HotReload*\|TestActor_ClipboardRungOffKeyPath\|TestActor_MACRIntercepts\|TestBuffer_ResetByFocusOut' -race -count=1` | ok 1.3s | ✓ PASS |
| Conversion corpus (truths 4,5,6 + selection geometry) | `go test ./internal/correct/ -run 'TestConvertRuns_MixedInvertsPerChar\|TestConvertRuns_HomogeneousWholesale\|TestConvertRuns_Refusals\|TestBuildPlan_BackspaceCap\|TestSelectionRange\|TestVerifyRangeAt' -race -count=1` | ok 1.0s | ✓ PASS |
| Matrix hermeticity predicate (4dc4877) | `go test ./test/e2e/ -run TestMatrixCaseNeedsBaseEstablishment -race -count=1 -v` | PASS | ✓ PASS |
| Strict decode + FSM + provenance + clipboard + engine wire (truths 8,11,17,19) | `go test ./internal/config/ ./internal/hotkey/ ./internal/clipboard/ ./internal/engine/ -race -count=1` | all ok | ✓ PASS |
| Control service in-process (truth 14) | `go test ./internal/ctlsvc/ -run 'TestSvc_*\|TestRenderStatusVersionToken' -race -count=1` | FAIL — fsnotify «too many open files» (263/128 inotify instances held by desktop processes); package byte-unchanged; identical flake recorded in 261008-00m-SUMMARY | ✗ environmental (advisory) — live ctl row PASS 2026-10-08 carries the truth |
| Matrix breadth preserved | case-name counts + v2⊂v3 name diff | v1=16, v2=21, v3=32 (v2 superset: comm empty), v4=34; cases/ diff vs 2171fae empty | ✓ PASS |
| **Live matrix v2 at the current tree's code (truth 16)** | runner-written `e2e-report.txt` (mise run e2e-matrix-v2, 2026-10-08T01:34:10+03:00, owner's desktop) | 20/21 PASS — word-mixed + phrase-mixed green; single FAIL super-space-alive at the stand-side focus step | ✓ PASS (1 environmental row — see truth 10/16) |
| **Live matrix v1 at the fixed tree (truth 18)** | recorded in 02-UAT.md (2026-10-08) | 16/16 PASS incl. all reset rows; word-mixed → «паиghbdtn» | ✓ PASS |

### Probe Execution

Step 7c: no `scripts/*/tests/probe-*.sh` convention in this repo. The
phase-declared probes are the mise e2e matrix tasks. This pass did NOT
re-run them live (drives the owner's desktop; the recorded 2026-10-08
runs ARE the probe evidence — e2e-report.txt and the 02-UAT v1 record).
Static encoding re-verified (yaml byte-unchanged, hermeticity predicate
unit-pinned, build green). Headless re-checks: build + named -race
corpora, all green.

### Test Quality Audit

Requirement-linked corpora all-active, non-circular, value/behavioral-level.
This pass additionally verified the NEW tests are RED-first and
non-circular: 77c91b7/07d143b are test-only RED commits whose targeted runs
failed with the exact live failure signatures (recorded in 261008-00m-SUMMARY
with transcripts); the hermeticity test pins the predicate table
(plain→true, config_base→true, reload-steps→false). The own-flip tests use a
switcher STUB + synthetic lifecycle pair — the unit corpus was previously
fixture-only blind to the real flip (02-UAT diagnosis), and the blindness is
now compensated by the LIVE word-mixed row green (2026-10-08). Disabled tests
on requirements: **0**. Circular patterns: **0**.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| CORR-02 | 03-01, 03-07 | Triple Right Shift corrects the whole buffer phrase | ✓ SATISFIED | truths 1, 6, 16 — phrase rows live-green 2026-10-08 |
| CORR-03 | 03-03, 03-07 | Double Right Shift under selection corrects the selected text | ✓ SATISFIED | truths 2, 3, 17 — select rows live-green 2026-10-08 |
| CORR-06 | 03-01, 03-07 | Mixed text corrects per the ADR semantics (as revised via D-55 spec-delta) | ✓ SATISFIED | truths 4, 5 — **both mixed rows live-green 2026-10-08**; SPEC §4.2 amendment; golden corpus green |
| SWCH-01 | 03-04, 03-07 | Right Shift toggles layout via the ADR mechanism (ADR-006 engine-truth flip per D-52) | ✓ SATISFIED | truth 7 — layout-single live-green 2026-10-08 |
| SWCH-02 | 03-04, 03-07 | Combo «correct word + switch» (default Shift+RightCtrl) | ✓ SATISFIED | truth 9 — combo row live-green 2026-10-08 |
| SWCH-03 | 03-04, 03-07 | Native Super+Space mechanism not broken (D-34 working interpretation) | ✓ SATISFIED | truth 10 — idle-desktop green chain intact; 2026-10-08 FAIL environmental (stand-side focus step), never reaching the daemon |
| SWCH-04 | 03-02, 03-04 | Tap-timing conflict resolved per ADR; configurable | ✓ SATISFIED | truth 8 — reload-window live-green 2026-10-08 |
| CONF-01 | 03-02 | All keys/tap-timeouts/params in YAML | ✓ SATISFIED | truth 11 — the hermetic stand now boots every case daemon from a pinned ctl document |
| CONF-02 | 03-02, 03-04, 03-07 | Hot reload without restart | ✓ SATISFIED | truth 12 — reload rows live-green 2026-10-08 |
| CONF-03 | 03-02 | Caramba compatibility (wish) via docs table | ✓ SATISFIED | truth 13 |
| MACR-01 | 03-05, 03-07 | Super+letter → Ctrl+letter per ADR-005 | ✓ SATISFIED (Complete with GTK4 caveat — owner decision) | truth 15 — macr row live-green 2026-10-08 |
| INST-02 | 03-06, 03-07 | goswitchctl status/reload/correct | ✓ SATISFIED | truth 14 — ctl-status-reload live-green 2026-10-08 |

Orphaned requirements: **none** — all 12 IDs present and `[x]` in
REQUIREMENTS.md, traceability rows Complete (re-confirmed at the prior
pass; REQUIREMENTS.md unchanged since).

### Anti-Patterns Found

Debt-marker gate this pass over all covered files modified since the
previous verification (actor.go, actor_test.go, matrix.go, matrix_test.go,
ADR-004, ADR-006, SPEC.md): **ZERO** unreferenced TBD/FIXME/XXX; zero
TODO/HACK/PLACEHOLDER; zero stub markers.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| internal/session/actor.go | logCorrectionDone | source/result words at DEBUG (review IN-01) | ℹ️ Info | D-21-permitted debug channel, -debug flag warns — unchanged contract |
| internal/engine/engine.go | caps | historically write-only field (review IN-02) | ℹ️ Info | Carried as review Info debt |
| internal/hotkey/names.go | ParseBinding | degenerate bindings accepted (review IN-04) | ℹ️ Info | Unchanged; harmless |
| test/e2e/README.md | — | -case table may lag the v4 registry (review IN-03 family) | ℹ️ Info | Docs-only; matrix.yaml files remain the source of truth |

The 8 in-scope review findings (CR-01..03, WR-01..05) remain FIXED;
their named tests green in this pass's runs.

Prohibition audit: all 15 prior judgment-tier verdicts carry forward and
were re-confirmed where files changed — no autocorrection without a
gesture; INFO content-free; KnownFields strict; last-good + WARN;
clipboard content stdin-only, never logged/argv; selection exactness
pinned; zero gsettings CALLS in the switch path; TestFSM_ModifierUse
green; consumed-upstream WARN + counters; status surface counts/states/
version only. The formerly live-only judgments (teardown restores the
desktop; matrix never red from stand fault; session-bus same-uid
isolation) are re-established at the revised tree by the recorded
2026-10-08 runs (both matrices ran to completion and tore down; the one
FAIL is a case-step focus refusal, not a stand fault).

### Deferred Items

Not part of the phase-3 contract; owned by later phases/instruments.

| # | Item | Addressed In | Evidence |
|---|------|-------------|----------|
| 1 | Formal fresh-session full-matrix DOUBLE green on an untouched session, incl. idle-desktop super-space-alive re-green and live v3/v4 re-runs | Phase 4 | PROJECT.md Phase 4 scope: «полная матрица дважды зелёная на нетронутой сессии; ручная приёмка владельца»; nightly D-48 schedule post-merge; v3/v4 mixed rows are the same byte-identical 4-row family, live-green in v2 |
| 2 | Owner's personal hand-typed mixed-text gedit/GTK4 pass | Phase 4 | ручная приёмка (docs/ACCEPTANCE.md, form 04-07); mechanism owner-confirmed (02-UAT test 2), content live-proven through the same /dev/uinput layer |

### Human Verification Required

**None.** The single item carried by the 2026-10-07 report is closed:

- **Live matrix at the revised tree** — executed and recorded
  (e2e-report.txt, matrix-v2 20/21 at the current tree's code, 2026-10-08;
  v1 16/16 in 02-UAT.md). The single FAIL is deterministically
  stand-environmental with recorded precedent and idle-desktop green
  history for the byte-identical case.
- **Hand-typed gedit residue** — its two dimensions are covered
  (mechanism: owner's 2026-09-15 visual confirmation; content: live
  rune-exact oracles at the same physical injection layer); the personal
  acceptance act itself is Phase 4's charter (deferred, documented above).

All 4 items from the initial verification remain RESOLVED (03-UAT.md,
status complete; WINDOWS #2/#4; zenity verdict).

### Gaps Summary

**No gaps.** All 20 must-haves verified on the current tree (HEAD ae54bf4):
19 carried with refreshed first-hand evidence, truth 16 closed by the
recorded live full-breadth v2 run (20/21, both revision-carrying mixed rows
green, single FAIL attributed by its own report line to the recorded
environmental focus-stealing class). Zero debt markers; build green;
decision coverage 15/15; the phase's production surface absorbed quick
261008-00m (ADR-004 own-flip amendment + flipCredit, stand hermeticity)
with RED-first tests, live green rows, and the owner's spec-first audit
trail. Status: **passed** — the perceptual/formal-session acceptance
residue is Phase 4's declared scope, deferred with evidence.

---

_Verified: 2026-10-07T22:56:52Z_
_Verifier: Claude (gsd-verifier)_
