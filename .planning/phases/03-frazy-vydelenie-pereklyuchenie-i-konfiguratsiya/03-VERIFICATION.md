---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
verified: 2026-09-15T15:55:41Z
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
  - cmd/goswitchctl/main.go
  - cmd/goswitchd/main.go
  - docs/CONFIG.md
  - engine/engine.go
  - engine/engine_test.go
  - engine/keys.go
  - .github/workflows/e2e-matrix.yml
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
  - test/e2e/main.go
  - test/e2e/matrix.go
  - test/e2e/matrix_test.go
  - test/e2e/preflight.go
  - test/e2e/README.md
covered_digest: "v1:sha256:7f4491ae20dbe9c66f4bebc1e61323a43873536d6e109c9326746765fb486783"
behavior_unverified: 1
overrides_applied: 0
behavior_unverified_items:
  - truth: "SC-5a: e2e-матрица v2 зелёная во всей широте на верифицируемом HEAD"
    test: "Run `mise run e2e-matrix-v2` on the live desktop at HEAD 44a4c65 (and optionally dispatch the e2e-matrix workflow on green106 at HEAD)"
    expected: "21/21 PASS, exit 0, teardown restores desktop state — re-establishing the acceptance evidence over the 8 review-fix commits (7152cd3..d94a3ce) that postdate the recorded green runs"
    why_human: "The recorded green matrix evidence (live 21/21 and CI run 34984814995) is at headSha 21df37c; HEAD adds 737 lines of behavioral Go/YAML changes on paths every matrix case exercises (WR-05 verify-after debounce, WR-03 preflight UID auth, CR-01 tap-key wiring, CR-02 verifyWait). No post-fix matrix run exists (gh run list shows none; the stray ./e2e binary predates even 21df37c). Live e2e drives the owner's desktop and is not re-run by the verifier (Phase 1/2 precedent). The headless corpus is green at HEAD, so this is an evidence-currency gap, not a defect signal."
human_verification:
  - test: "Re-run the acceptance matrix at HEAD: `mise run e2e-matrix-v2` (live desktop); optionally dispatch e2e-matrix on green106 at 44a4c65"
    expected: "21/21 PASS with exit 0 and a restored desktop — closes the evidence-currency gap left by the review-fix commits (see behavior_unverified_items)"
    why_human: "Live e2e drives the owner's GNOME session; verifier policy re-fetches machine evidence but never drives the desktop. The only green matrix runs (live + CI 34984814995) are at 21df37c, 8 behavioral fix commits before HEAD."
  - test: "Owner decision — WINDOWS ledger #2 (open): confirm the D-24 corpus reading. The plan's literal nothing-to-convert example (привет → out==input) was resolved toward the must-haves (homogeneous ranges convert wholesale per уточнение D-22; the changed=false surface is implemented and reserved for externally anchored ranges)"
    expected: "Owner accepts the уточнение D-22 resolution (matrix v1 word-ru-en keeps привет→ghbdtn wholesale) and marks ledger #2 resolved — or requests the strict-anchor reading, which would re-open the golden corpus"
    why_human: "Planned semantics resolution routed to the verify gate by the plan's FLAGGED assumption; the automated oracle passes either way — this is an owner-intent confirmation, not a code check."
  - test: "Owner decision — WINDOWS ledger #4 (open, unmet-truth): GTK4-Wayland (GNOME 46) does not apply IBus-forwarded key events. The daemon side is wire-proven (dbus-monitor captures engine burst + ibus-daemon relay to the focused InputContext) but GTK4 widgets never act — affecting the MACR Ctrl+letter end-effect, a live ADR-003 level-2 witness, and the D-28 Ctrl+V rung driver"
    expected: "Owner either accepts this as a documented GNOME platform limitation (wire relay = the provable boundary) or schedules investigation on other surfaces; then flips the REQUIREMENTS.md MACR-01 checkbox/traceability to match the decision (currently inconsistent: implementation verified, checkbox still Pending)"
    why_human: "Platform truth verified live by the executor; no automated oracle can decide 'document vs investigate'. The requirement's end-user-visible effect on GTK4 surfaces is unprovable on this desktop."
  - test: "Owner eyeball — the corrected zenity selection verdict (03-07 coverage D5): a silently-surviving ctrl+a selection on zenity makes CommitText REPLACE the selected residue (field settles at the converted word alone); this reverses the 03-03 spike table's transparency verdict, which was probe-contaminated"
    expected: "Owner acknowledges the corrected actual (matrix row select-all-zenity pins it: expect_text \"ghbdtn\", expect_sel_step degraded)"
    why_human: "A live-truth reversal between two phase documents; the automated oracle pins the new behavior, but the reversal itself is exactly what the verify gate exists to surface."
---

# Phase 3: Фразы, выделение, переключение и конфигурация — Verification Report

**Phase Goal (ROADMAP.md):** «Полный словарь жестов — тройной тап исправляет фразу, двойной при выделении исправляет выделенное, смешанный текст корректируется по ADR-семантике; одиночный Right Shift переключает раскладку механизмом по ADR; все хоткеи и таймауты настраиваются YAML-конфигом с hot reload и управляются через `goswitchctl` — всё подтверждает зелёная e2e-матрица v2 полной широты.»
**Verified:** 2026-09-15T15:55:41Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Verification Approach Note (MVP mode discrepancy)

ROADMAP.md declares `**Mode:** mvp` for Phase 3, but the goal is a Russian
capability statement, not a canonical User Story — `gsd_run query
user-story.validate` → `valid: false`. Per `verify-mvp-mode.md` the MVP
user-flow framing fires only when BOTH `mode: mvp` AND a user-story goal are
present; a non-story goal is surfaced as a discrepancy. Same resolution as
Phase 1/2 verifications: standard goal-backward verification against the 5
Success Criteria + the 7 plans' must_haves, with a User Flow Coverage table
below as evidence framing. **Discrepancy surfaced for the owner:** run
`/gsd mvp-phase 3` if a User-Story goal is wanted; no action required
otherwise.

## Stale-criteria note (D-34/D-35, per 03-CONTEXT)

Two SC-3 phrasings are recorded as STALE against accepted ADR-001/ADR-002:
«святой порядок: свитч на первый тап» (ADR-002 decides all actions at window
expiry) and «родное Super+Space и MRU продолжают работать, индикатор GNOME
актуален» (ADR-001 «один хозяин»: the indicator shows one source and does not
reflect the internal mode). Per D-34 the verifier did NOT require their
literal fulfillment; the working interpretation (native Super mechanism not
broken by goswitch's presence; mode-record oracle, never gsettings) was
verified instead. This is owner-preauthorized (D-34, BIND), not an override.

## Live-gate evidence policy

Live e2e (all `mise run e2e-*` tasks) drives the owner's desktop and was NOT
re-run by the verifier (Phase 1/2 precedent). Machine evidence was
independently re-fetched: `gh run view 34984814995` → `conclusion: success`,
`event: workflow_dispatch`, `headBranch: gsd/phase-03-frazy-...` (the phase
branch), workflow `e2e-matrix`. Headless gates WERE re-run at HEAD: full
`go test ./... -race -count=1` (12/12 packages ok) and `mise run ci`
(build + vet + golangci-lint strict + race; 0 lint issues).

**Evidence-currency finding (the one material gap):** the green matrix runs
(live 21/21 at 17:54 local, CI 34984814995 at headSha 21df37c) PREDATE the 8
review-fix commits (7152cd3..d94a3ce, 18:26–18:45 local) — 737
insertions/153 deletions of Go/YAML changes on paths every matrix case
exercises: `internal/session/actor.go` (278 lines: CR-01 tap-key wiring,
CR-02 verifyWait, WR-01 rung off-mutex, WR-05 verify-after debounce),
`internal/clipboard/clipboard.go` (CR-03), `internal/config/config.go` (WR-02),
`internal/hotkey/fsm.go` (SetTapKey), `test/e2e/matrix.go` (WR-03 UID auth —
used by every case's preflight), `matrix-v2.yaml` (verify_wait_ms 250 reload).
No post-fix matrix or CI dispatch exists (`gh run list` shows 2 runs total).
The headless corpus at HEAD — including the new RED→GREEN pins for every fix
(TestActor_HotReloadTapKey, TestActor_HotReloadVerifyWait,
TestClipboard_KilledNotEmpty/DeadlineKillNotEmpty,
TestActor_ClipboardRungOffKeyPath, TestActor_VerifyAfterLevel1 debounced,
TestFSM_SetTapKey) — is green, so this is an evidence-currency gap routed to
human re-run, not a defect signal. Phase 2's verifier explicitly checked the
same property and found only docs drift; here code drifted.

## User Flow Coverage

User capability (goal, non-story form): «полный словарь жестов коррекции и
переключения, настраиваемый YAML-ом и управляемый goswitchctl».

| Step | Expected | Evidence | Status |
|------|----------|----------|--------|
| Type a wrong-layout phrase, triple-tap Right Shift | whole phrase converts `ghbdtn ghbdtn`→`привет привет` (registers, 70-rune phrase, RU→EN) | matrix-v2.yaml 8 phrase rows; actor.go Triple→startPhraseCorrection; TestActor_TripleTapCorrectsPhrase (re-run green) | ✓ |
| Select wrong text, double-tap | exactly the selection converts, both geometries, 3 surfaces | matrix-v2.yaml 5 select rows (zenity degraded pin, GTE LTR+partial+reverse, chromium RTL); TestActor_DoubleTapSelectionCorrects/SelectionLeftToRight | ✓ |
| Type mixed-script text, tap | only foreign runs convert (`gfbпривет`→`паипривет`) | TestConvertRuns_GoldenCorpus verbatim (D-22/D-23); word-mixed/phrase-mixed rows | ✓ |
| Single Right Shift | internal mode flips EN↔RU both ways (mode-record oracle) | layout-single row; TestActor_FlipOnSingle; zero gsettings in actor.go (grep) | ✓ |
| Shift+RightCtrl combo | word corrected FIRST, then mode flips | TestActor_ComboWordThenFlip (sink order + log order); combo-word-layout row | ✓ |
| Edit the YAML config | applied without restart; typo → last-good + visible error | reload-window / reload-invalid-last-good / ctl-status-reload rows; watcher corpus + TestActor_HotReload* | ✓ |
| goswitchctl status / reload / correct | live answers, --json, exit≠0 contract, D-32 error visibility | ctl-smoke case; ctlsvc corpus incl. over-the-wire name-guard on a private dbus-daemon | ✓ |
| Super+letter in configured apps | intercepted → Ctrl+letter forwarded (wire-proven) | TestActor_MACR* corpus; macr-super-letter row (relay oracle; GTK4 client boundary → Human item 3) | ✓ |
| Outcome: green matrix v2 proves the lot | 21/21 PASS live + CI | green at 21df37c (re-fetched); NOT re-run at HEAD after review fixes | ⚠ human |

## Goal Achievement

### Observable Truths

Consolidated from the 5 ROADMAP Success Criteria (primary contract) merged
with the 7 plans' must_haves (≈45 plan-level truths deduplicate into these).

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC-1: Triple Right Shift corrects the whole phrase since the last hard reset (CORR-02, D-25) — live `ghbdtn ghbdtn`→`привет привет` | ✓ VERIFIED | actor.go:561 Triple→startPhraseCorrection over Phrase()/ReplacePhrase (buffer.go:117-134); TestActor_TripleTapCorrectsPhrase/EmptyBuffer/PhraseVerifyMismatch green under -race (re-run at HEAD); 8 live phrase rows green at acceptance commit |
| 2 | SC-1: Double Right Shift under an active selection corrects exactly the selected range, both geometric directions (CORR-03, D-30, Pitfall 6) | ✓ VERIFIED | SelectionRange/VerifyRangeAt (verify.go:24-52); selection branch actor.go:1295+; TestActor_DoubleTapSelectionCorrects/SelectionLeftToRight/SelectionMixedConverts green; 5 select rows live at acceptance |
| 3 | Clipboard round-trip rung opt-in (default OFF), order after verify-mismatch, best-effort restore, stdin-only content (D-28/D-29, ADR-003 rung C) | ✓ VERIFIED | clipboard.go (cmd.Stdin-only, pinned flags, CR-03 kill-vs-empty fix :86); TestActor_ClipboardRungDisabledByDefault/AfterMismatch/OffKeyPath + TestClipboard_KilledNotEmpty/DeadlineKillNotEmpty green; no live rung driver exists on this desktop (documented exclusion, WINDOWS #4 family) |
| 4 | SC-2: mixed text converts by runs — anchor = last letter, foreign runs independently, own runs untouched, digits neutral (CORR-06, D-22/D-23) | ✓ VERIFIED | runs.go ConvertRuns (144 LOC, run scanner); TestConvertRuns_GoldenCorpus verbatim from CONTEXT specifics — re-run green |
| 5 | Уточнение D-22: homogeneous wholesale (Phase-2 behavior intact); D-24 changed=false = success-without-changes; D-16→D-23 succession (`gfbпривет`→`паипривет`) | ✓ VERIFIED | runs.go:44-91 (wholesale branch) + actor.go:1330/:1409 D-24 branches (outcome done, no WARN, zero sink calls); matrix-v1 word-mixed row expects паипривет; v1 16/16 at acceptance — owner confirmation of the corpus reading → Human item 2 |
| 6 | Backspace cap D-27: default 50, level 2 not built over cap, zero deletions on refusal (reason backspace-cap) | ✓ VERIFIED | plan.go DefaultBackspaceCap=50, LevelNone; actor.go:1514 skipCorrection("backspace-cap"); TestBuildPlan_BackspaceCap green |
| 7 | SC-3: single Right Shift flips internal mode EN↔RU behind `layoutset`; oracle = mode log record, never gsettings (SWCH-01, D-34/ADR-001) | ✓ VERIFIED | flipScript actor.go:1167; TestActor_FlipOnSingle green; grep gsettings internal/ → 0 hits; layout-single live both directions at acceptance |
| 8 | SC-3/SWCH-04: tap-timing conflict resolved per ADR-002; window + tap key configurable; FSM corpus (11+ tests) semantically untouched (D-35); post-review: tap_key and verify_wait_ms actually consumed (CR-01/CR-02) | ✓ VERIFIED | NewFSM(window, tapKeyval) + SetTapKey (fsm.go:66-89); actor tapKeyval/tapKeyName/verifyWait fields fed by applySnapshot (:653-661); TestFSM_ConfigurableTapKey/SetTapKey + TestActor_HotReloadTapKey/VerifyWait green — the review found these knobs dead and the fixes are wired and pinned |
| 9 | SC-3/SWCH-02: combo Shift+RightCtrl corrects the word THEN flips (D-36 order), binding renavigable, series Reset, buffer not fed (Pitfall 4) | ✓ VERIFIED | combo branch above mode branches (actor.go:1215-1225, FamilyMask held-subset predicate); settleCombo at every pipeline terminal (:599-609, :585, :707); TestActor_ComboWordThenFlip/EmptyBufferStillFlips/ConfigurableBinding/DoesNotFeedBuffer + TestFSM_ModifierUse green |
| 10 | SC-3/SWCH-03 (working interpretation, D-34 pre-authorized): native Super mechanism not broken — daemon keeps its IBus name and corrects after the shell chord | ✓ VERIFIED | super-space-alive row (live PASS ×4 at acceptance; Super+Space itself is not ydotool-injectable — documented); literal MRU/indicator wording STALE per D-34, not required |
| 11 | SC-4/CONF-01: all hotkeys/timeouts/params in YAML (hotkeys/timeouts/correction/macr); strict KnownFields decode, named-field errors, range ceilings (D-31/D-33) | ✓ VERIFIED | load.go:28 KnownFields(true); config.go per-section Validate; TestLoad_UnknownKeyRejected + TestValidate_* green; docs/CONFIG.md complete reference |
| 12 | SC-4/CONF-02: hot reload without restart — dir-watch + debounce, last-good + WARN, one snapshot per event, stale-timer no-re-arm (D-32, Pitfall 8) | ✓ VERIFIED | watch.go (atomic.Pointer, dir-watch :127, "config reload rejected" :237, Reload() shared core); AttachConfig + applySnapshot; TestWatch_* + TestActor_HotReloadWindowNewSeries/OptionsAndCombo/InvalidKeepsLastGood green |
| 13 | SC-4/CONF-03: Caramba correspondence table in docs, no name cloning, model-difference caveat (D-31) | ✓ VERIFIED | docs/CONFIG.md:105-121 — table maps all keys to Caramba purposes with the first-tap vs window-expiry caveat |
| 14 | SC-4/INST-02: goswitchctl status(--json)/reload/correct against a live daemon; primary-owner name guard; recover shim per method; status carries mode/counters/config only (D-32 visible) | ✓ VERIFIED | ctlsvc.go (ErrNotPrimaryOwner :40, guard :193-200, recoverMethod on all 3 methods); cmd/goswitchctl stdlib-only; TestSvc_Status/ReloadConfig(×3)/CorrectNow/NameGuard (over a real private dbus-daemon)/RecoverShim + ctl-smoke live ×4 at acceptance |
| 15 | SC-5/MACR-01: implemented per ADR-005 — interception above combo/mode branches (RU pin), burst on Super release, consumed-upstream WARN + counters, per-app a11y observer with degradation ladder + lazy start, alt_modifier default-off (b.3) | ✓ VERIFIED | actor.go MACR branch (:955-1008, macrPendingKeyval); appid.go (FocusedApp, ErrBusClosed, StateChanged match); TestActor_MACRIntercepts/ConsumedUpstream/RULayoutStillIntercepts/AltModifier/DisabledByDefault/PerAppMatch/AppidDegradation/GlobalWhenNoApps/AppidLazyStart green; live: probe verdicts + wire-relay oracle + per-app ×3 — client-side effect on GTK4 stops at the boundary → Human item 3 |
| 16 | SC-5: e2e-matrix v2 green in full breadth (21 cases: 8 phrase incl. registers + 70-rune A3 pin, 5 selection rows across 3 surfaces both geometries, mixed, combo, single, super-space, 3 reload rows, MACR; clipboard row excluded with the documented spike-verdict comment) — **at the verified HEAD** | ⚠ PRESENT_BEHAVIOR_UNVERIFIED | matrix-v2.yaml (21 cases verified on disk); green live 21/21 AND CI 34984814995 (success, re-fetched) — both at 21df37c; the 8 review-fix commits (737 lines on exercised paths) postdate every green run; headless corpus green at HEAD. One re-run closes it → Human item 1 |
| 17 | anchorPos seam carries the anchor to the actor on every surface; per-surface actuals spike-pinned (Pitfall 1; A1/A2/A5) | ✓ VERIFIED | engine.go:89/:190-195 signature + wire call + DEBUG trace; all implementors updated; spike table in 03-03-SUMMARY; zenity verdict corrected live in 03-07 (→ Human item 4) |
| 18 | D-30 continuity: no selection → the Phase-2 word path unchanged; Phase-2 corpus green without expectation edits | ✓ VERIFIED | TestActor_NoSelectionStillWord (green by design through both RED rounds); v1 matrix 16/16 at acceptance with only the planned word-mixed succession edit |
| 19 | Key/binding provenance: closed name tables from ibuskeysyms.h; KeyV=0x076, KeyControlR=0xffe4, KeySuperL/R; ParseBinding + FamilyMask press-side rule | ✓ VERIFIED | names.go:103-146; keys.go:36-51 provenance comments; TestParseBinding/TestParseBinding_KeyvalProvenance/TestFamilyMask green |
| 20 | CI circuit: workflow defaults to matrix-v2.yaml (v1 path preserved); first green green106 v2 run recorded; mise task registry complete (17 e2e tasks) | ✓ VERIFIED | e2e-matrix.yml:29/:69-76 (post-WR-04 env form); gh run 34984814995 re-fetched (success, phase branch, 2026-09-15T14:54Z); mise.toml e2e-* tasks verified |

**Score:** 19/20 truths verified (1 present, behavior-unverified at HEAD)

Behavior-dependent truths note: every state-transition/invariant truth in
this phase (phrase/selection pipelines, verify-after debounce, hot-reload
window arming, combo settlement fan-out, MACR release-time burst, clipboard
kill-vs-empty, name guard over the wire) is covered by a named behavioral
test **re-run green under -race at HEAD by the verifier**; the live-desktop
dimension is covered by the green matrix runs at the acceptance commit —
whose currency at HEAD is the single ⚠ item.

### Decision Coverage

`gsd_run query check.decision-coverage-verify` (03-CONTEXT.md): **15/15
decisions honored** by shipped artifacts, none missing (non-blocking gate).

### Required Artifacts

`gsd_run query verify.artifacts` per plan: **all 7 plans report valid**.
Substantiveness cross-checked by the verifier directly (all key files exist,
6218 LOC over the 20 core artifacts; zero stub markers; all symbols wired as
specified in artifacts_produced sections).

### Key Link Verification

`gsd_run query verify.key-links` per plan: 5/7 plans fully verified by the
tool; plan 06 reported 2 failures and plan 07 one — **all three are
tool-shape false negatives, manually disproven**:
- Plan 06 `internal/ctlsvc → internal/session` ("EISDIR"): ctlsvc.go:50
  `StatusSnapshot() session.Status`, :119 CorrectNow — the prescribed
  direction (session never imports ctlsvc).
- Plan 06 `cmd/goswitchctl → internal/ctlsvc` ("EISDIR"): main.go:26-27
  `org.djarvur.goswitch` / `/org/djarvur/goswitch` bus name+path constants.
- Plan 07 `matrix-v2.yaml → matrix.go` (pattern miss): the link is the
  strict schema — matrix.go:143-145 defines matrixKindSelect/Combo/Reload,
  :243-245 the yaml fields, :398-400 the closed vocabularies; matrix-v2.yaml
  uses exactly those kinds (verified per-row).

All other links verified by the tool, including actor→ConvertRuns (the D-23
single-code-path pin: grep shows zero `Detect(` in actor.go), actor→engine
emitters, main→AttachConfig, main→ctlsvc.Run, workflow→matrix-v2.yaml.

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| internal/session/actor.go | correction ranges | buf.Phrase()/activeRange + selectionState from live SetSurroundingText pushes | ✓ live pushes (select rows correct through both geometries) | ✓ FLOWING |
| internal/session/actor.go | config snapshot | Watcher.Snapshot() over atomic.Pointer (fsnotify dir events / ctl Reload) | ✓ reload rows change window/verify_wait live at acceptance | ✓ FLOWING |
| internal/session/actor.go | MACR counters | macrIntercepted/macrConsumed increments on live key events | ✓ super-intercept records gated in macr row | ✓ FLOWING |
| internal/ctlsvc/ctlsvc.go | status line | renderStatus ← StatusSnapshot counters + watcher LastError/ConfigPath | ✓ corrections_done=1 observed in ctl-smoke after forced correct | ✓ FLOWING |
| test/e2e/matrix.go | report | per-case runMatrixCaseIsolated results over real daemon+surfaces | ✓ 21 PASS lines in the CI run log (21df37c) | ✓ FLOWING (at acceptance commit; HEAD re-run → Human item 1) |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full headless corpus | `mise exec -- go test ./... -race -count=1` | 12/12 packages ok (engine, appid, clipboard, config, correct, ctlsvc, hotkey, logging, session, layouts, test/e2e) | ✓ PASS |
| Full CI gate | `mise run ci` | exit 0, lint 0 issues, all tests ok | ✓ PASS |
| Key behavioral tests exist | `go test ./internal/session/ -list '.*'` + correct package | 30+ named pins enumerated incl. all review-fix tests (HotReloadTapKey/VerifyWait, ClipboardRungOffKeyPath) | ✓ PASS |
| CI matrix run (live) | `gh run view 34984814995` (re-fetched) | conclusion success, workflow_dispatch, phase branch, workflow e2e-matrix | ✓ PASS (at 21df37c) |
| Post-review-fix matrix run | `gh run list --workflow e2e-matrix` | only 2 runs total — none at HEAD | ✗ FAIL (evidence-currency → Human item 1) |
| Commits exist | `git cat-file -t` × 31 hashes from SUMMARYs | all present | ✓ PASS |
| Matrix v2 breadth | case names in matrix-v2.yaml | 21 cases; select/combo/reload kinds present; clipboard exclusion documented | ✓ PASS |
| Live e2e re-run | not re-run (drives owner's desktop) | CI artifact + live reports at acceptance commit cited instead | ? SKIP (by policy) |

### Probe Execution

Step 7c: no `scripts/*/tests/probe-*.sh` convention in this repo. The
phase's runnable probes are the mise e2e tasks (live-session; covered by the
CI-artifact policy above) and the CI gate (re-run, PASS). The matrix-v2 CI
run IS the phase-declared probe target — executed and green at 21df37c.

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| internal/correct/*_test.go (7 files) | CORR-02/03/06 | all | 0 | 0 | value (rune-exact golden corpus) | OK |
| internal/session/actor_test.go (~90 tests) | CORR-02/03, SWCH-01/02, CONF-02, MACR-01 | all | 0 | 0 | behavioral (op-log order, sink sequences, log shapes) | OK |
| internal/config/*_test.go, internal/hotkey/*_test.go | CONF-01/02, SWCH-04 | all | 0 | 0 | value (decode errors, ranges, wire-truth masks) | OK |
| internal/ctlsvc/ctlsvc_test.go | INST-02 | all | 1 env-guard | 0 | behavioral (over-the-wire on a private dbus-daemon) | OK |
| internal/clipboard/clipboard_test.go | CORR-03 | all | 0 | 0 | value + real-kill shapes (CR-03 corpus forges real children) | OK |
| internal/appid/appid_test.go | MACR-01 | all | 0 | 0 | behavioral (fake a11y feed) | OK |
| test/e2e/matrix_test.go | SC-5 | all | 0 | 0 | value (strict decode, exit code) | OK |

- Disabled tests on requirements: **0** (the single `t.Skipf` in
  TestSvc_NameGuard fires only on machines with no dbus-daemon at all — it
  executed in every gate here and on the CI runner; documented).
- Circular patterns: **0** — expected values are hand-authored literals
  (ghbdtn/привет/паипривет…); the golden corpus is independent of the system.
- Assertion strength: matrix oracles are rune-exact; actor corpus pins exact
  emitter sequences and counters. No existence-only assertions on
  requirement-linked tests.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| CORR-02 | 03-01, 03-07 | Triple Right Shift corrects the whole buffer phrase | ✓ SATISFIED | truths 1, 6, 16 (phrase rows) |
| CORR-03 | 03-03, 03-07 | Double Right Shift under selection corrects the selected text | ✓ SATISFIED | truths 2, 3, 17 |
| CORR-06 | 03-01, 03-07 | Mixed text: only the wrong-layout part converts (ADR semantics) | ✓ SATISFIED | truths 4, 5 |
| SWCH-01 | 03-04, 03-07 | Right Shift toggles layout via the ADR mechanism | ✓ SATISFIED | truth 7 |
| SWCH-02 | 03-04, 03-07 | Combo «correct word + switch» (default Shift+RightCtrl) | ✓ SATISFIED | truth 9 |
| SWCH-03 | 03-04, 03-07 | Native Super+Space mechanism not broken (D-34 working interpretation; literal wording STALE) | ✓ SATISFIED | truth 10 |
| SWCH-04 | 03-02, 03-04 | Tap-timing conflict resolved per ADR; configurable | ✓ SATISFIED | truth 8 |
| CONF-01 | 03-02 | All keys/timeouts/params in YAML | ✓ SATISFIED | truth 11 |
| CONF-02 | 03-02, 03-04, 03-07 | Hot reload without restart | ✓ SATISFIED | truth 12 |
| CONF-03 | 03-02 | Caramba compatibility (wish) via docs table | ✓ SATISFIED | truth 13 |
| MACR-01 | 03-05, 03-07 | Super+letter → Ctrl+letter per ADR-005 | ✓ SATISFIED (daemon side; client-boundary owner decision → Human item 3) | truth 15 |
| INST-02 | 03-06, 03-07 | goswitchctl status/reload/correct | ✓ SATISFIED | truth 14 |

Orphaned requirements: **none** — REQUIREMENTS.md traceability maps exactly
these 12 IDs to Phase 3; every plan requirement field resolves to them.

Bookkeeping warning: REQUIREMENTS.md still shows **MACR-01 `[ ]`/Pending**
while the mechanism is implemented, unit-pinned and matrix-covered — coherent
with WINDOWS #4 being open, but inconsistent with the other 11 rows (all
flipped Complete). Resolve together with Human item 3; the file's «Last
updated» footer is also stale.

### Anti-Patterns Found

Debt-marker gate: **ZERO** TBD/FIXME/XXX and zero TODO/HACK/PLACEHOLDER
across all phase Go/YAML/TOML/workflow/doc files (re-grepped).

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| repo root `./e2e` | — | Stray untracked 5.9 MB Go ELF binary (mtime 17:37, predates the matrix commit) — a `go build ./test/e2e` artifact never cleaned | ℹ️ Info | Untracked, never committed — repo hygiene only; delete or gitignore |
| .planning/REQUIREMENTS.md | 45, 108 | MACR-01 checkbox Pending vs implemented+verified reality (all other phase-3 rows flipped) | ⚠️ Warning (bookkeeping) | Resolves with Human item 3; footer «Last updated» stale |
| internal/session/actor.go | 1536-1538 | logCorrectionDone writes source/result words at DEBUG (review IN-01, unfixed by scope decision) | ℹ️ Info | D-21-permitted debug channel, -debug flag warns; no INFO leak (verified: INFO records carry outcome only) |
| engine/engine.go | 103-109 | Engine.caps write-only field (review IN-02, unfixed) | ℹ️ Info | Dead state; first future reader inherits a race — carried as review Info debt |
| internal/hotkey/names.go | 123-146 | ParseBinding accepts degenerate bindings (review IN-04, unfixed) | ℹ️ Info | Harmless today (WR-02 closed the tap_key side); grammar lacks canonical-form check |
| test/e2e/README.md | — | -case table stale vs registry (review IN-03, unfixed) | ℹ️ Info | Docs-only |

The 8 in-scope review findings (CR-01..03, WR-01..05) are all FIXED with
RED→GREEN evidence at commits 7152cd3..d94a3ce, each fix pinned by a named
test re-run green at HEAD by this verifier. The 6 Info findings were
explicitly out of the fix scope (REVIEW-FIX.md) and are carried above.

Prohibition audit (15 judgment-tier across the 7 plans; ADR-550 D4 —
non-authoritative LLM-judge verdicts, `unverified-prohibition — human review
recommended` for the live-only ones): programmatically upheld — no
autocorrection without a gesture (startCorrection called only from FSM
Double/Triple expiry, the combo branch, CorrectNow: actor.go:416/:558/:1221);
INFO records content-free (skipCorrection/logCorrectionDone verified);
KnownFields strict decode; last-good + WARN; clipboard content never in any
log or argv (cmd.Stdin only); selection range exactness unit-pinned; save/
restore byte-exact with --clear for empty; zero gsettings in internal/ (grep);
TestFSM_ModifierUse green; consumed-upstream WARN with counters; status
surface carries counts/states only. Live-only judgments (teardown restores
the desktop; matrix never red from stand fault; session-bus same-uid
isolation) rest on the green runs at 21df37c — human review recommended
alongside Human item 1.

### Human Verification Required

1. **Re-run the acceptance matrix at HEAD** — `mise run e2e-matrix-v2` on the
   live desktop (optionally dispatch e2e-matrix on green106 at 44a4c65).
   Expected: 21/21 PASS, exit 0, desktop restored. Why: every recorded green
   matrix run predates the 8 review-fix commits (737 behavioral lines); the
   verifier does not drive the owner's session.

2. **Owner decision — WINDOWS #2 (D-24 corpus reading).** Confirm the
   уточнение D-22 resolution: homogeneous ranges convert wholesale (matrix v1
   `привет`→`ghbdtn` stays green); the nothing-to-convert surface
   (changed=false → outcome done, no WARN) is implemented and reserved for
   externally anchored ranges. Expected: accept → ledger #2 resolved.

3. **Owner decision — WINDOWS #4 (GTK4-Wayland forwarded-events).** The
   daemon side of MACR-01/level-2/Ctrl+V is wire-proven (engine burst +
   ibus-daemon relay captured); GTK4 widgets never apply forwarded events on
   this desktop. Expected: accept as a documented platform limitation or
   schedule investigation; then align the REQUIREMENTS.md MACR-01 row with
   the decision.

4. **Owner eyeball — zenity selection verdict reversal (03-07 D5).** A
   silently-surviving ctrl+a selection on zenity makes the commit replace the
   selected residue (field settles at the converted word alone) — the 03-03
   spike's transparency verdict was probe-contaminated. Expected: acknowledge
   the corrected actual pinned in select-all-zenity.

### Gaps Summary

**No code gaps.** All 19 substance truths verified with behavioral evidence
re-run green at HEAD; all artifacts substantive and wired; all key links
proven (3 tool false negatives manually disproven); 12/12 requirements
satisfied with zero orphans; decision coverage 15/15; zero debt markers;
zero disabled/circular tests; `mise run ci` green (0 lint issues). The
`human_needed` status comes from: (1) the evidence-currency gap on the
acceptance matrix — the only green live/CI matrix runs predate the review-fix
commits, so the phase's crown criterion («всё подтверждает зелёная e2e-матрица
v2») is machine-proven only at 21df37c, not at HEAD — and (2)-(4) the owner
decision items the plans deliberately routed to the verify gate (WINDOWS #2,
#4, and the zenity verdict reversal). One command (`mise run e2e-matrix-v2`)
plus three owner acknowledgments close everything.

---

_Verified: 2026-09-15T15:55:41Z_
_Verifier: Claude (gsd-verifier)_
