---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
verified: 2026-09-20T14:32:40Z
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
  - engine/engine.go
  - engine/engine_test.go
  - engine/keys.go
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

covered_digest: "v1:sha256:a4d4246434dafbd9332d81b51e12225e1de1833eabe335b0ceb61a9a57dda88c"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 20/20
  previous_verified: 2026-09-16T19:19:27Z
  reason: "fingerprint refresh after Phase-4 gap closure (04-08 G-4-1 witness-probe fix; 04-09 G-4-2 nightly gate wiring, commits 994ee2a..04d00ad) reworked shared covered files — test/e2e harness only (matrix.go quiesce budget, focus_helper.py, quiesce_test.go new, e2e-matrix.yml schedule/preflight); zero daemon production code changed"
  gaps_closed: []
  gaps_remaining: []
  regressions: []
  resolved_human_items:
    - "Human item 1 (matrix v2 re-run at HEAD) — CLOSED: owner-authorized live run 2026-09-15T21:41+03:00, 21/21 PASS at HEAD 54ea460 (03-UAT.md item 1); superseded by the stronger Phase-4 D-48 double-run gate — CI run 35108412175 (re-fetched: success) at 9f2fd75 with BOTH matrix-v3 runs 31/31, and v3 is a strict superset of v2 (all 21 v2 case names verified present)"
    - "Human item 2 (WINDOWS ledger #2, D-24 corpus reading) — CLOSED: owner accepted 2026-09-15, homogeneous-wholesale resolution confirmed, ledger #2 resolved (03-UAT.md item 2)"
    - "Human item 3 (WINDOWS ledger #4, GTK4-Wayland forwarded events) — CLOSED: owner accepted 2026-09-15 as a documented GNOME platform limitation; REQUIREMENTS.md MACR-01 flipped Complete with the caveat, footer refreshed (03-UAT.md item 3; commit 9069287)"
    - "Human item 4 (zenity selection verdict reversal, 03-07 D5) — CLOSED: owner acknowledged 2026-09-15; corrected actual pinned by select-all-zenity and re-proven live in the 21/21 run at 54ea460 (03-UAT.md item 4)"
---

# Phase 3: Фразы, выделение, переключение и конфигурация — Verification Report

**Phase Goal (ROADMAP.md):** «Полный словарь жестов — тройной тап исправляет фразу, двойной при выделении исправляет выделенное, смешанный текст корректируется по ADR-семантике; одиночный Right Shift переключает раскладку механизмом по ADR; все хоткеи и таймауты настраиваются YAML-конфигом с hot reload и управляются через `goswitchctl` — всё подтверждает зелёная e2e-матрица v2 полной широты.»
**Verified:** 2026-09-20T14:32:40Z
**Status:** passed
**Re-verification:** Yes — stale-fingerprint refresh on the Phase-4 gap-closure tree (HEAD 04d00ad, branch gsd/phase-04-postavka-i-priemka)

## Re-verification Scope and Method

The phase passed verification at close (2026-09-15T15:55:41Z, status
human_needed with all 4 human items) and was re-verified on the Phase-4 tree
at 42d10a1 (2026-09-16T19:19:27Z, status passed). Since that second pass the
codebase advanced through Phase-4 GAP CLOSURE only (04-08 G-4-1, 04-09
G-4-2; commits 994ee2a..04d00ad). This pass re-checked every must-have
against the CURRENT tree; live-desktop re-runs were NOT needed — recorded
evidence plus current-tree static/automated checks.

**What Phase-4 gap closure actually changed in phase-3-owned files**
(diff 42d10a1..HEAD): ONLY the e2e measurement harness, and only for
robustness —

- `test/e2e/matrix.go` (+28): quiesce-gate timing constants —
  `witnessProbeTimeout` 4s→10s, named `witnessQuiesceWindow`; behavior
  (probe-then-wait loop) unchanged.
- `test/e2e/focus_helper.py` (+156): additive `focused-app` idle-preflight
  primitive (G-4-2); existing `focus`/`witness` contracts untouched.
- `test/e2e/quiesce_test.go` (new, +53): headless corpus pinning the
  quiesce-window invariant (04-08 staged RED→GREEN).
- `.github/workflows/e2e-matrix.yml` (+96): nightly `schedule` trigger,
  idle-desktop preflight step, D48_SKIP soft mode — all ADDITIVE; the
  v1/v2/v3 matrix routes (:194-216) and the double-run D-48 gate are
  intact.

**The ENTIRE daemon production tree is bit-identical** from the green
double-run commit (9f2fd75, CI 35108412175, 31/31 ×2) through the previous
verification HEAD (42d10a1) to the current HEAD (04d00ad): `git diff
9f2fd75..HEAD -- cmd engine internal layouts` shows exactly one file,
`cmd/goswitchctl/main_test.go` (+72, test corpus). Zero correction /
selection / switching / config / ctlsvc production lines changed since the
green proof. All pinned contracts survive (details per truth below).

**Evidence-currency note (the three red matrix runs at 9ff56cb):** CI runs
35142914380 / 35188570260 / 35271719601 (2026-09-16/17, all dispatched at
the pinned pre-fix SHA 9ff56cb) FAILED. 04-UAT.md root-causes these as
Phase-4 gap G-4-1 — the e2e harness's OWN a11y witness probe (4s budget)
exceeding its timeout on large fresh-session desktop trees (~3.0s bare
walk + load), a HARNESS defect: "no test/e2e code changed since the green
double-run proof at 9f2fd75"; the lone select-partial-gte content miss is
attributed to the degraded bridge ("case green in every healthy run" —
including both runs this report cites). Fixed by 04-08 (focus-first
traversal + 10s floor, pinned by the now-green quiesce_test.go corpus,
live probe smoke 3.0-4.7s on the degraded session); attempt 3 re-ran the
pinned pre-fix SHA so it could not contain the fix. No green FULL-matrix
live run exists yet at the post-fix tree — that re-validation is wired
into the unattended nightly D-48 gate (04-09, G-4-2 resolution, commit
04d00ad). This is Phase-4 acceptance currency, not a phase-3 regression:
the daemon code under test is bit-identical to the code that passed 31/31
twice, and the v2/v3 case corpora and workflow routes are unchanged
(re-verified below).

**The one previously behavior-unverified truth remains closed** by
recorded evidence: (a) the owner-authorized live matrix v2 re-run — 21/21
PASS at HEAD 54ea460 (03-UAT.md item 1); (b) superseded upward by Phase
4's D-48 double-run gate — CI run 35108412175 (re-fetched: success) at
9f2fd75 with BOTH matrix-v3 runs 31/31, and matrix-v3.yaml is a strict
superset of v2 (re-verified this pass: all 21 v2 case names present in
v3's 31, `comm -23` empty). Production code is identical 9f2fd75→HEAD.

## Verification Approach Note (MVP mode discrepancy)

ROADMAP.md declares `**Mode:** mvp` for Phase 3, but the goal is a Russian
capability statement, not a canonical User Story — `gsd_run query
user-story.validate` → `valid: false` (checked at initial verification).
Per `verify-mvp-mode.md` the MVP user-flow framing fires only when BOTH
`mode: mvp` AND a user-story goal are present; standard goal-backward
verification against the 5 Success Criteria + the 7 plans' must_haves
applies. Unchanged from prior passes; `/gsd mvp-phase 3` remains available
to the owner if a User-Story goal is wanted.

## Stale-criteria note (D-34/D-35, per 03-CONTEXT)

Unchanged: two SC-3 phrasings are recorded STALE against accepted
ADR-001/ADR-002 («святой порядок: свитч на первый тап»; «индикатор GNOME
актуален»), verified via the working interpretation (native Super mechanism
not broken; mode-record oracle). Owner-preauthorized (D-34, BIND), not an
override.

## Live-gate evidence policy

Live e2e drives the owner's desktop and is not re-run by the verifier
(Phase 1/2/3 precedent). Machine evidence re-fetched THIS pass: `gh run
list --workflow e2e-matrix.yml` → latest success remains 35108412175
(conclusion success, workflow_dispatch, headSha 9f2fd75 — the D-48
double-run gate, both v3 runs 31/31); the three later failures at 9ff56cb
are the G-4-1 harness finding documented above and in 04-UAT.md
(resolved by 04-08; nightly re-validation wired by 04-09). Headless gates
re-run at HEAD 04d00ad: `go build ./...` OK; full `go test ./... -race
-count=1` → 14/14 packages ok (now including the new quiesce corpus in
test/e2e); `go vet ./...` clean; `mise exec -- golangci-lint run` → 0
issues (upstream exhaustruct deprecation warning only).

## User Flow Coverage

User capability (goal, non-story form): «полный словарь жестов коррекции и
переключения, настраиваемый YAML-ом и управляемый goswitchctl». All rows
re-checked on the current tree.

| Step | Expected | Evidence | Status |
|------|----------|----------|--------|
| Type a wrong-layout phrase, triple-tap Right Shift | whole phrase converts `ghbdtn ghbdtn`→`привет привет` (registers, 70-rune phrase, RU→EN) | matrix-v2.yaml 8 phrase rows (all present in v3, re-verified by name); actor.go:578/:1383 Triple→startPhraseCorrection; TestActor_TripleTapCorrectsPhrase green in this pass's -race run | ✓ |
| Select wrong text, double-tap | exactly the selection converts, both geometries, 3 surfaces | matrix-v2.yaml 5 select rows; verify.go:24/:41 SelectionRange/VerifyRangeAt; TestActor_DoubleTapSelectionCorrects/SelectionLeftToRight green | ✓ |
| Type mixed-script text, tap | only foreign runs convert (`gfbпривет`→`паипривет`) | TestConvertRuns_GoldenCorpus verbatim (D-22/D-23), green; word-mixed/phrase-mixed rows | ✓ |
| Single Right Shift | internal mode flips EN↔RU both ways (mode-record oracle) | layout-single row; TestActor_FlipOnSingle green; grep gsettings internal/session/ → 0 hits re-confirmed (the gsettings hits remain ALL internal/install — Phase-4 lifecycle code outside the switch path) | ✓ |
| Shift+RightCtrl combo | word corrected FIRST, then mode flips | TestActor_ComboWordThenFlip; settleCombo at terminals actor.go:601/:723/:1342/:1352 | ✓ |
| Edit the YAML config | applied without restart; typo → last-good + visible error | reload-window / reload-invalid-last-good / ctl-status-reload rows; watcher pins + TestActor_HotReload* green | ✓ |
| goswitchctl status / reload / correct | live answers, --json, exit≠0 contract, D-32 error visibility | ctl-smoke case; ctlsvc corpus incl. over-the-wire name-guard; status line leads with version= (D-37, ctlsvc.go:135) — counts-and-states-only invariant re-affirmed in code | ✓ |
| Super+letter in configured apps | intercepted → Ctrl+letter forwarded (wire-proven) | TestActor_MACR* corpus; macr-super-letter row (relay oracle; GTK4 client boundary = documented platform limitation, WINDOWS #4 closed by owner) | ✓ |
| Outcome: green matrix proves the lot | green runs recorded | 21/21 v2 live at 54ea460 (UAT) + 31/31 v3 twice in CI 35108412175 (re-fetched success); v3 ⊇ v2 re-verified; production code identical 9f2fd75→HEAD | ✓ |

## Goal Achievement

### Observable Truths

Consolidated from the 5 ROADMAP Success Criteria (primary contract) merged
with the 7 plans' must_haves. All 20 re-verified on the current tree
(HEAD 04d00ad); production line numbers from the 42d10a1 pass remain valid
verbatim — those files are bit-identical.

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC-1: Triple Right Shift corrects the whole phrase since the last hard reset (CORR-02, D-25) — live `ghbdtn ghbdtn`→`привет привет` | ✓ VERIFIED | actor.go:578/:1383-1387 Triple→startPhraseCorrection over Phrase()/ReplacePhrase (buffer.go:117/:126); TestActor_TripleTapCorrectsPhrase/EmptyBuffer/PhraseVerifyMismatch green under -race (re-run this pass); 8 phrase rows in matrix-v2 (names re-verified), carried into v3 |
| 2 | SC-1: Double Right Shift under an active selection corrects exactly the selected range, both geometric directions (CORR-03, D-30, Pitfall 6) | ✓ VERIFIED | SelectionRange/VerifyRangeAt (verify.go:24/:41); selection branch + selectionState in actor.go; TestActor_DoubleTapSelectionCorrects/SelectionLeftToRight/SelectionMixedConverts green; 5 select rows live green in both the 21/21 (54ea460) and 31/31 (35108412175) runs |
| 3 | Clipboard round-trip rung opt-in (default OFF), order after verify-mismatch, best-effort restore, stdin-only content (D-28/D-29, ADR-003 rung C) | ✓ VERIFIED | clipboard.go:154 cmd.Stdin-only content path, CR-03 kill-vs-empty fix intact; TestActor_ClipboardRungDisabledByDefault/AfterMismatch/OffKeyPath + TestClipboard_KilledNotEmpty/DeadlineKillNotEmpty green; no live rung driver on this desktop (documented exclusion; WINDOWS #4 family closed) |
| 4 | SC-2: mixed text converts by runs — anchor = last letter, foreign runs independently, own runs untouched, digits neutral (CORR-06, D-22/D-23) | ✓ VERIFIED | runs.go:63 ConvertRuns (run scanner unchanged); TestConvertRuns_GoldenCorpus verbatim from CONTEXT specifics — green this pass |
| 5 | Уточнение D-22: homogeneous wholesale (Phase-2 behavior intact); D-24 changed=false = success-without-changes; D-16→D-23 succession (`gfbпривет`→`паипривет`) | ✓ VERIFIED | wholesale branch in runs.go; D-24 `!changed` branches intact at actor.go:1346/:1425; matrix-v1 word-mixed row expects паипривет; owner confirmed the corpus reading (WINDOWS #2 resolved, 03-UAT item 2) |
| 6 | Backspace cap D-27: default 50, level 2 not built over cap, zero deletions on refusal (reason backspace-cap) | ✓ VERIFIED | plan.go:30 DefaultBackspaceCap=50 (re-grepped); LevelNone; actor.go:1530 skipCorrection("backspace-cap"); TestBuildPlan_BackspaceCap green |
| 7 | SC-3: single Right Shift flips internal mode EN↔RU behind `layoutset`; oracle = mode log record, never gsettings (SWCH-01, D-34/ADR-001) | ✓ VERIFIED | flipScript actor.go:1190 (re-grepped); TestActor_FlipOnSingle green; grep gsettings internal/session/ → 0 hits (re-run this pass). The gsettings references introduced since the initial pass remain ALL in internal/install (Phase-4 install/uninstall lifecycle — reads/restores org.gnome.desktop.input-sources at install time), outside the daemon's switch path; contract intact |
| 8 | SC-3/SWCH-04: tap-timing conflict resolved per ADR-002; window + tap key configurable; FSM corpus semantically untouched (D-35); tap_key and verify_wait_ms actually consumed (CR-01/CR-02) | ✓ VERIFIED | NewFSM(window, tapKeyval) + SetTapKey (fsm.go:66/:85); applySnapshot feeds a.verifyWait (:669-675) and a.tapKeyval (:684); TestFSM_ConfigurableTapKey/SetTapKey + TestActor_HotReloadTapKey/VerifyWait green |
| 9 | SC-3/SWCH-02: combo Shift+RightCtrl corrects the word THEN flips (D-36 order), binding renavigable, series Reset, buffer not fed (Pitfall 4) | ✓ VERIFIED | combo branch above mode branches; settleCombo at every pipeline terminal (actor.go:601/:620/:723/:1342/:1352); TestActor_ComboWordThenFlip/EmptyBufferStillFlips/ConfigurableBinding/DoesNotFeedBuffer + TestFSM_ModifierUse green |
| 10 | SC-3/SWCH-03 (working interpretation, D-34 pre-authorized): native Super mechanism not broken — daemon keeps its IBus name and corrects after the shell chord | ✓ VERIFIED | super-space-alive row live PASS ×4 at acceptance, carried in v2→v3 (name re-verified in both corpora); literal MRU/indicator wording STALE per D-34, not required; gsettings scoping as truth 7 |
| 11 | SC-4/CONF-01: all hotkeys/timeouts/params in YAML (hotkeys/timeouts/correction/macr); strict KnownFields decode, named-field errors, range ceilings (D-31/D-33) | ✓ VERIFIED | load.go:28 KnownFields(true) (re-grepped); config.go per-section Validate; TestLoad_UnknownKeyRejected + TestValidate_* green; docs/CONFIG.md complete reference (unchanged since last pass) |
| 12 | SC-4/CONF-02: hot reload without restart — dir-watch + debounce, last-good + WARN, one snapshot per event, stale-timer no-re-arm (D-32, Pitfall 8) | ✓ VERIFIED | watch.go:89-90 atomic.Pointer, Reload() :167, "config reload rejected" :237; AttachConfig + applySnapshot; TestWatch_* + TestActor_HotReloadWindowNewSeries/OptionsAndCombo/InvalidKeepsLastGood green |
| 13 | SC-4/CONF-03: Caramba correspondence table in docs, no name cloning, model-difference caveat (D-31) | ✓ VERIFIED | docs/CONFIG.md:105-121 — table maps all keys to Caramba purposes with the first-tap vs window-expiry caveat; unchanged by Phase 4 |
| 14 | SC-4/INST-02: goswitchctl status(--json)/reload/correct against a live daemon; primary-owner name guard; recover shim per method; status carries machine state only, never user text (D-32 visible) | ✓ VERIFIED | ctlsvc.go:37-40 ErrNotPrimaryOwner (re-grepped), :91/:102/:120 recoverMethod on all 3 methods, :166 recover shim; TestSvc_Status/ReloadConfig(×3)/CorrectNow/NameGuard (over a real private dbus-daemon)/RecoverShim green; ctl-smoke live ×4 at acceptance. D-37's leading `version=` token (ctlsvc.go:135) is additive; counts-and-states-only invariant intact |
| 15 | SC-5/MACR-01: implemented per ADR-005 — interception above combo/mode branches (RU pin), burst on Super release, consumed-upstream WARN + counters, per-app a11y observer with degradation ladder + lazy start, alt_modifier default-off (b.3) | ✓ VERIFIED | actor.go:123-124/:999 macrPendingKeyval/macrIntercepted; appid.go:41/:163 ErrBusClosed/FocusedApp; TestActor_MACR* corpus green; live: probe verdicts + wire-relay oracle + per-app ×3; GTK4 client boundary = documented platform limitation (WINDOWS #4 closed by owner; REQUIREMENTS.md MACR-01 Complete with caveat) |
| 16 | SC-5: e2e-matrix green in full breadth (21 v2 cases: 8 phrase incl. registers + 70-rune A3 pin, 5 selection rows across 3 surfaces both geometries, mixed, combo, single, super-space, 3 reload rows, MACR; clipboard row excluded with the documented spike-verdict comment) — **at the verified HEAD** | ✓ VERIFIED (evidence chain intact) | (a) 03-UAT.md item 1 — owner-authorized live `mise run e2e-matrix-v2` 21/21 PASS at HEAD 54ea460; (b) D-48 double-run gate — CI 35108412175 re-fetched success at 9f2fd75, BOTH matrix-v3 runs 31/31; v3 re-verified this pass a strict superset (comm -23 of case names empty; v2 still 21 cases on disk, v3 31); production code identical 9f2fd75→HEAD (single test-corpus file). The 3 red runs at the pinned pre-fix SHA 9ff56cb are the G-4-1 HARNESS defect (root-caused, fixed by 04-08, corpus green; see Evidence-currency note) — no daemon-code regression signal: no test/e2e case spec changed since the green proof, and the one content miss is bridge-attributed |
| 17 | anchorPos seam carries the anchor to the actor on every surface; per-surface actuals spike-pinned (Pitfall 1; A1/A2/A5) | ✓ VERIFIED | engine.go:87-89/:190-195 signature + wire call + DEBUG trace; implementors unchanged; zenity verdict corrected + owner-acknowledged (03-UAT item 4); select-all-zenity row green in both matrix runs |
| 18 | D-30 continuity: no selection → the Phase-2 word path unchanged; Phase-2 corpus green without expectation edits | ✓ VERIFIED | TestActor_NoSelectionStillWord present and green in this pass's -race run; v1 matrix 16/16 at acceptance; v1 path still routed in the workflow (:194-195) |
| 19 | Key/binding provenance: closed name tables from ibuskeysyms.h; KeyV=0x076, KeyControlR=0xffe4, KeySuperL/R; ParseBinding + FamilyMask press-side rule | ✓ VERIFIED | names.go; engine/keys.go:36-51 with ibuskeysyms.h provenance comments; TestParseBinding/TestParseBinding_KeyvalProvenance/TestFamilyMask green |
| 20 | CI circuit: workflow routes the full-breadth matrix by input (v1 AND v2 paths preserved); first green run recorded; mise task registry complete | ✓ VERIFIED (wording superseded) | Re-checked this pass on the current workflow: default input matrix-v3.yaml (:46, D-48), v1/v2 routes preserved in BOTH run steps (:194-216), D48_SKIP soft mode additive and schedule-only; mise still carries 32 e2e-* task lines incl. e2e-matrix-v2/:v3 (mise.toml unchanged). Substance — CI circuit wired to the full-breadth matrix with a recorded green run — holds: green v3 31/31 ×2 at 35108412175 (re-fetched) on top of the green v2 21/21 at 54ea460 |

**Score:** 20/20 truths verified (0 present-but-behavior-unverified)

Behavior-dependent truths note: every state-transition/invariant truth in
this phase (phrase/selection pipelines, verify-after debounce, hot-reload
window arming, combo settlement fan-out, MACR release-time burst, clipboard
kill-vs-empty, name guard over the wire) is covered by a named behavioral
test re-run green under -race at the current HEAD by this verifier; the
live-desktop dimension is covered by the green v2 21/21 run at 54ea460
(UAT) AND the green v3 31/31 double-run at 9f2fd75 (CI 35108412175,
re-fetched) — both on production code bit-identical to the current tree.

### Decision Coverage

`gsd_run query check.decision-coverage-verify` (03-CONTEXT.md), re-run this
pass: **15/15 decisions honored** by shipped artifacts, none missing
(non-blocking gate).

### Advisory (New Scope, Unevidenced)

Re-verification ran; the Step 7 scan over files modified since the previous
verification (matrix.go, focus_helper.py, quiesce_test.go, e2e-matrix.yml,
plus planning docs) found zero debt markers (TBD/FIXME/XXX/TODO/HACK/
PLACEHOLDER all 0), zero stubs, zero empty implementations. Three
informational notes, none evidence-bearing against phase-3 must-haves:

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| 1 | CI matrix runs 35142914380/35188570260/35271719601 (all at pinned pre-fix SHA 9ff56cb) are red — G-4-1 harness defect (a11y witness probe budget), root-caused and fixed by 04-08 (quiesce corpus green); nightly D-48 gate (04-09) will produce the first green live run at a post-fix tree | other (e2e harness currency) | HAS deterministic evidence both ways (04-UAT.md root-cause; fix corpus green in this pass's -race run; daemon code bit-identical to the green double-run) — recorded so the evidence chain stays auditable; a phase-3 must-have asserts the green breadth at the verified HEAD, which stands |
| 2 | golangci-lint v2.13+ warns `exhaustruct` is deprecated (replaced by `exhaustruct_v5`); .golangci.yml still names the old linter | other (tooling config) | Re-confirmed this pass (same warning, 0 issues); config belongs to the repo-wide quality gate, not a phase-3 truth. Resolve in a housekeeping pass |
| 3 | internal/install shells out to gsettings (28 refs) — the initial pass's "zero gsettings in internal/" grep no longer holds verbatim | other (scope note) | Carried from the previous pass: by design, Phase-4 D-39 install lifecycle reads/restores input sources at INSTALL time; the switch path (internal/session) re-verified at 0 hits this pass, so truths 7/10's contract is intact. Recorded so future greps scope correctly |

### Required Artifacts

`gsd_run query verify.artifacts` per plan at initial verification: all 7
plans valid. This pass: all 71 covered files exist on the current tree
(existence check before fingerprinting; focus_helper.py and quiesce_test.go
added to coverage per phases-1/2 precedent — they carry the e2e surface
interaction the phase-3 matrix evidence depends on, and quiesce_test.go now
pins matrix.go's gate). The gap-closure diff touched only the harness
surfaces of them (see Re-verification Scope). Substantiveness re-confirmed
by grep of every pinned symbol (per-truth evidence above); zero stub
markers.

### Key Link Verification

Initial pass: 5/7 plans fully verified by the tool; the 3 tool-shape false
negatives manually disproven (ctlsvc→session direction, goswitchctl bus
name constants, matrix.yaml↔matrix.go strict schema). Re-checked this pass
on the current tree:

- ctlsvc.go still returns `session.Status` (:124 area) — direction intact
  (file bit-identical).
- cmd/goswitchctl bus name/path constants intact (the only production-tree
  change since the green proof is main_test.go — constants file untouched).
- matrix.go kind vocabulary (:154-158 matrixKindType..Select) still parses
  matrix-v2.yaml (21 cases on disk, names re-verified) and matrix-v3.yaml
  (31 cases); the quiesce-constant edit sits at :683+, after the vocabulary
  block.
- All other links unchanged: actor→ConvertRuns (zero `Detect(` in
  actor.go), actor→engine emitters, main→AttachConfig, main→ctlsvc.Run.

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| internal/session/actor.go | correction ranges | buf.Phrase()/activeRange + selectionState from live SetSurroundingText pushes | ✓ select rows green in both recorded matrix runs | ✓ FLOWING |
| internal/session/actor.go | config snapshot | Watcher.Snapshot() over atomic.Pointer (fsnotify dir events / ctl Reload) | ✓ reload rows change window/verify_wait live in recorded runs | ✓ FLOWING |
| internal/session/actor.go | MACR counters | macrIntercepted/macrConsumed increments on live key events | ✓ super-intercept gated in macr row, carried in v3 | ✓ FLOWING |
| internal/session/actor.go | version (D-37) | daemon build identity pinned via SetVersion at construction | ✓ additive; unit-pinned in the corpus | ✓ FLOWING |
| internal/ctlsvc/ctlsvc.go | status line | renderStatus ← StatusSnapshot (leading version=, ctlsvc.go:135) + watcher LastError/ConfigPath | ✓ corrections_done observed live in ctl-smoke at acceptance | ✓ FLOWING |
| test/e2e/matrix.go | report | per-case runMatrixCaseIsolated results over real daemon+surfaces | ✓ 21 PASS lines (54ea460) + 31/31 ×2 (CI 35108412175) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Build | `go build ./...` | exit 0 | ✓ PASS |
| Full headless corpus (single full-suite run) | `go test ./... -race -count=1` | 14/14 packages ok (incl. new quiesce corpus) | ✓ PASS |
| Vet | `go vet ./...` | clean | ✓ PASS |
| Lint (strict) | `mise exec -- golangci-lint run` | 0 issues (upstream exhaustruct deprecation warning only) | ✓ PASS |
| Named behavioral tests exist | `go test ./internal/... -list '.*'` (enumeration only) | all sampled pins present: TripleTapCorrectsPhrase, DoubleTapSelectionCorrects, SelectionLeftToRight, NoSelectionStillWord, ConvertRuns_GoldenCorpus, FlipOnSingle, ComboWordThenFlip, HotReloadTapKey, HotReloadInvalidKeepsLastGood, ClipboardRungOffKeyPath, FSM_SetTapKey, MACRIntercepts, Load_UnknownKeyRejected, Svc_NameGuard, VerifyAfterLevel1 | ✓ PASS |
| CI matrix run (recorded, re-fetched) | `gh run list --workflow e2e-matrix.yml` | latest success = 35108412175 at 9f2fd75 (D-48 double-run gate, both v3 runs 31/31); 3 later failures at pinned pre-fix SHA 9ff56cb = G-4-1 harness finding (fixed by 04-08, corpus green this pass) | ✓ PASS (with currency note) |
| Live e2e re-run | not re-run (drives owner's desktop) | recorded UAT 21/21 at 54ea460 + CI 35108412175 cited | ? SKIP (by policy) |
| Matrix v2 breadth preserved | case-name diff v2 vs v3 | v2 still 21 cases (names listed); v3 = strict superset (31; `comm -23` empty) | ✓ PASS |
| Production-tree identity | `git diff 9f2fd75..HEAD -- cmd engine internal layouts` | single file: cmd/goswitchctl/main_test.go (+72 test corpus) | ✓ PASS |

### Probe Execution

Step 7c: no `scripts/*/tests/probe-*.sh` convention in this repo. The
phase's runnable probes remain the mise e2e tasks (live-session; covered by
the recorded-evidence policy above) and the CI gate (re-run this pass:
build + vet + lint + full -race suite, all green). The matrix CI run is the
phase-declared probe target — green at 9f2fd75 (35108412175) and, for v2
specifically, at 54ea460 (UAT); the post-04-08 harness's first green live
validation arrives via the nightly D-48 gate wired by 04-09 (Phase-4
acceptance currency, not a phase-3 contract).

### Test Quality Audit

Unchanged from prior passes (Phase-4 gap closure added quiesce_test.go —
a harness-invariant corpus for matrix.go, which IS phase-3-covered, and it
runs green in this pass's -race suite): phase-3 requirement-linked corpora
remain all-active, non-circular, value/behavioral-level. This pass
re-confirmed: the single env-guard skip in TestSvc_NameGuard still executed
(dbus-daemon present); expected values remain hand-authored literals
(ghbdtn/привет/паипривет); matrix oracles rune-exact. Disabled tests on
requirements: **0**. Circular patterns: **0**.

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
| MACR-01 | 03-05, 03-07 | Super+letter → Ctrl+letter per ADR-005 | ✓ SATISFIED (Complete with GTK4 caveat — owner decision, UAT item 3) | truth 15; REQUIREMENTS.md:45/:108 aligned (Complete with the platform caveat) |
| INST-02 | 03-06, 03-07 | goswitchctl status/reload/correct | ✓ SATISFIED | truth 14 |

Orphaned requirements: **none** — REQUIREMENTS.md re-grepped this pass:
all 12 IDs present, all `[x]`, traceability table maps exactly these 12
IDs to Phase 3, every traceability row Complete; every plan requirement
field resolves to them.

### Anti-Patterns Found

Debt-marker gate re-run this pass on all covered files modified since the
previous verification (matrix.go, focus_helper.py, quiesce_test.go,
e2e-matrix.yml) AND on the whole changed covered set: **ZERO**
TBD/FIXME/XXX and zero TODO/HACK/PLACEHOLDER.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| .golangci.yml | — | `exhaustruct` linter name deprecated upstream since v2.13.0 (→ exhaustruct_v5) | ℹ️ Info | Tooling housekeeping; 0 lint issues today (re-confirmed); not a phase-3 regression (see Advisory 2) |
| internal/session/actor.go | logCorrectionDone | source/result words at DEBUG (review IN-01) | ℹ️ Info | File bit-identical; D-21-permitted debug channel, -debug flag warns; INFO records content-free |
| engine/engine.go | Engine.caps | write-only field (review IN-02) | ℹ️ Info | Unchanged (engine untouched since prior pass); dead state carried as review Info debt |
| internal/hotkey/names.go | ParseBinding | degenerate bindings accepted (review IN-04) | ℹ️ Info | Unchanged; harmless (WR-02 closed the tap_key side) |
| test/e2e/README.md | — | -case table stale vs registry (review IN-03) | ℹ️ Info | Docs-only; unchanged by gap closure |

The 8 in-scope review findings (CR-01..03, WR-01..05) remain FIXED with
RED→GREEN pins, re-confirmed green in this pass's -race run.

Prohibition audit: the prior passes' 15 judgment-tier verdicts carry
forward — all programmatically-upheld ones re-confirmed this pass (no
autocorrection without a gesture — startCorrection still only from FSM
Double/Triple expiry, combo branch, CorrectNow; INFO content-free;
KnownFields strict; last-good + WARN; clipboard content stdin-only, never
logged/argv; selection exactness pinned; zero gsettings in the switch path
(re-grepped); TestFSM_ModifierUse green; consumed-upstream WARN + counters;
status surface counts/states/version only). The formerly live-only judgments
(teardown restores the desktop; matrix never red from stand fault;
session-bus same-uid isolation) rest on TWO green recorded runs (21/21 at
54ea460 per UAT, and 31/31 ×2 at 35108412175 with the D-48 fresh-session
preflight + double-run gate) — evidence materially stronger than at the
initial pass.

### Human Verification Required

**None — all 4 items from the initial verification remain RESOLVED**,
recorded in 03-UAT.md (status: complete, 4 passed, 0 issues):

1. ~~Matrix v2 re-run at HEAD~~ → owner-authorized live run 21/21 PASS at
   HEAD 54ea460 (2026-09-15); superseded by the D-48 double-run gate (v3
   31/31 ×2, CI 35108412175, re-fetched success at 9f2fd75).
2. ~~WINDOWS ledger #2 (D-24 corpus reading)~~ → owner accepted the
   уточнение D-22 resolution 2026-09-15; ledger #2 resolved.
3. ~~WINDOWS ledger #4 (GTK4-Wayland forwarded events)~~ → owner accepted
   as a documented GNOME platform limitation; REQUIREMENTS.md MACR-01
   flipped Complete with the caveat.
4. ~~Zenity selection verdict reversal (03-07 D5)~~ → owner acknowledged;
   corrected actual pinned by select-all-zenity and re-proven live.

### Gaps Summary

**No gaps.** All 20 must-haves verified on the current tree (HEAD 04d00ad):
Phase-4 gap closure (04-08/04-09) touched phase-3 files only in the e2e
measurement harness — quiesce budget constants, an additive focus_helper
subcommand, a new harness-invariant corpus, and additive workflow
preflight/schedule steps — while the ENTIRE daemon production tree is
bit-identical to the code that passed the green double-run (31/31 ×2 at
9f2fd75) and the previous two verification passes. Every pinned contract
survives: the switch path's gsettings abstinence, the D-32
counts-and-states-only status invariant (leading version token), the
strict-decode/hot-reload config core, the phrase/selection/runs/combo/MACR
pipelines, and the matrix breadth (v2 intact at 21 cases, provably subsumed
by the green v3 31/31 double-run; routes re-verified in both workflow run
steps). Headless gates re-run green at HEAD (build, full -race suite 14/14,
vet, strict lint 0 issues). The three red CI runs at the pinned pre-fix SHA
are the Phase-4 G-4-1 harness finding — root-caused, fixed, corpus green —
not a phase-3 regression. Status remains **passed**; the fingerprint is
refreshed over all 71 covered files (focus_helper.py and quiesce_test.go
newly covered, matching phases-1/2 precedent).

---

_Verified: 2026-09-20T14:32:40Z_
_Verifier: Claude (gsd-verifier)_
