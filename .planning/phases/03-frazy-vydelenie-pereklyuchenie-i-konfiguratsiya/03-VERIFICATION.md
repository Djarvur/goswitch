---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
verified: 2026-09-16T19:19:27Z
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

covered_digest: "v1:sha256:653571db7ff881699d5946f8d830e45c1442741527a5cc7dcddbd3829bc460de"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: human_needed
  previous_score: 19/20
  previous_verified: 2026-09-15T15:55:41Z
  reason: "fingerprint refresh after Phase 4 execution (commits 357a647..42d10a1) reworked shared covered files — install package, selfcheck/version identity (D-37), perf harness, matrix v3 (D-47), double-run gate (D-48)"
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
**Verified:** 2026-09-16T19:19:27Z
**Status:** passed
**Re-verification:** Yes — stale-fingerprint refresh on the Phase-4 tree (HEAD 42d10a1, branch gsd/phase-04-postavka-i-priemka)

## Re-verification Scope and Method

The phase passed verification at close (2026-09-15T15:55:41Z, status
human_needed with all 4 human items) and the codebase then advanced through
Phase 4 (поставка и приёмка: install/uninstall lifecycle, selfcheck + D-37
version identity, goreleaser pipeline, perf harness, matrix v3, double-run
gate, docs/release). This pass re-checked every must-have against the
CURRENT tree; live-desktop re-runs were NOT needed — recorded evidence plus
current-tree static/automated checks.

**What Phase 4 actually changed in phase-3-owned files** (diff
30ca7ae..HEAD): purely additive D-37 version identity — `Actor.SetVersion`,
`Status.Version`, a leading `version=` token in `renderStatus`, a
`-version` flag on goswitchd — plus test/e2e harness growth (matrix v3,
watchdog, fresh-session preflight, perf case) and the new
`internal/install` package. No correction/selection/switching/config
pipeline line was reworked. All pinned contracts survive (details per truth
below).

**The one previously behavior-unverified truth is closed by recorded
evidence:** (a) the owner-authorized live matrix v2 re-run — 21/21 PASS at
HEAD 54ea460, post-review-fixes (03-UAT.md item 1); (b) superseded upward by
Phase 4's D-48 double-run gate — CI run 35108412175 (re-fetched by this
verifier: conclusion success, workflow_dispatch, headSha 9f2fd75,
2026-09-16T14:25Z) with BOTH matrix-v3 runs 31/31, and matrix-v3.yaml is a
strict superset of v2 (all 21 v2 case names verified present, 10 new
gedit/x11 rows). After 9f2fd75, HEAD differs only by docs and a
goswitchctl status-line parser test corpus — zero production code.

## Verification Approach Note (MVP mode discrepancy)

ROADMAP.md declares `**Mode:** mvp` for Phase 3, but the goal is a Russian
capability statement, not a canonical User Story — `gsd_run query
user-story.validate` → `valid: false` (checked at initial verification).
Per `verify-mvp-mode.md` the MVP user-flow framing fires only when BOTH
`mode: mvp` AND a user-story goal are present; standard goal-backward
verification against the 5 Success Criteria + the 7 plans' must_haves
applies. Unchanged from the initial pass; `/gsd mvp-phase 3` remains
available to the owner if a User-Story goal is wanted.

## Stale-criteria note (D-34/D-35, per 03-CONTEXT)

Unchanged: two SC-3 phrasings are recorded STALE against accepted
ADR-001/ADR-002 («святой порядок: свитч на первый тап»; «индикатор GNOME
актуален»), verified via the working interpretation (native Super mechanism
not broken; mode-record oracle). Owner-preauthorized (D-34, BIND), not an
override.

## Live-gate evidence policy

Live e2e drives the owner's desktop and is not re-run by the verifier
(Phase 1/2/3 precedent). Machine evidence re-fetched THIS pass: `gh run
view 35108412175` → conclusion success, event workflow_dispatch, headSha
9f2fd75, workflow e2e-matrix (the D-48 double-run gate; run log per commit
4946e28: both v3 runs 31/31). Headless gates re-run at HEAD 42d10a1: `go
build ./...` OK; full `go test ./... -race -count=1` → 14/14 packages ok
(includes the new internal/install package); `go vet ./...` clean;
`golangci-lint run` → 0 issues.

## User Flow Coverage

User capability (goal, non-story form): «полный словарь жестов коррекции и
переключения, настраиваемый YAML-ом и управляемый goswitchctl». All rows
re-checked on the current tree.

| Step | Expected | Evidence | Status |
|------|----------|----------|--------|
| Type a wrong-layout phrase, triple-tap Right Shift | whole phrase converts `ghbdtn ghbdtn`→`привет привет` (registers, 70-rune phrase, RU→EN) | matrix-v2.yaml 8 phrase rows (all present in v3); actor.go:578/:1383 Triple→startPhraseCorrection; TestActor_TripleTapCorrectsPhrase in the green -race corpus | ✓ |
| Select wrong text, double-tap | exactly the selection converts, both geometries, 3 surfaces | matrix-v2.yaml 5 select rows; verify.go:24/:41 SelectionRange/VerifyRangeAt; TestActor_DoubleTapSelectionCorrects/SelectionLeftToRight green | ✓ |
| Type mixed-script text, tap | only foreign runs convert (`gfbпривет`→`паипривет`) | TestConvertRuns_GoldenCorpus verbatim (D-22/D-23), green; word-mixed/phrase-mixed rows | ✓ |
| Single Right Shift | internal mode flips EN↔RU both ways (mode-record oracle) | layout-single row; TestActor_FlipOnSingle green; zero gsettings in internal/session (the 28 new gsettings hits are ALL internal/install — Phase-4 lifecycle code outside the switch path) | ✓ |
| Shift+RightCtrl combo | word corrected FIRST, then mode flips | TestActor_ComboWordThenFlip; settleCombo at terminals actor.go:601/:723/:1342/:1352 | ✓ |
| Edit the YAML config | applied without restart; typo → last-good + visible error | reload-window / reload-invalid-last-good / ctl-status-reload rows; watcher pins + TestActor_HotReload* green | ✓ |
| goswitchctl status / reload / correct | live answers, --json, exit≠0 contract, D-32 error visibility | ctl-smoke case; ctlsvc corpus incl. over-the-wire name-guard; status line now LEADS with version= (D-37, additive) — counts-and-states-only invariant re-affirmed in code | ✓ |
| Super+letter in configured apps | intercepted → Ctrl+letter forwarded (wire-proven) | TestActor_MACR* corpus; macr-super-letter row (relay oracle; GTK4 client boundary = documented platform limitation, WINDOWS #4 closed by owner) | ✓ |
| Outcome: green matrix proves the lot | green runs recorded | 21/21 v2 live at 54ea460 (UAT) + 31/31 v3 twice in CI 35108412175 (re-fetched success); v3 ⊇ v2 | ✓ |

## Goal Achievement

### Observable Truths

Consolidated from the 5 ROADMAP Success Criteria (primary contract) merged
with the 7 plans' must_haves. All 20 re-verified on the current tree; line
numbers refreshed to HEAD 42d10a1.

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC-1: Triple Right Shift corrects the whole phrase since the last hard reset (CORR-02, D-25) — live `ghbdtn ghbdtn`→`привет привет` | ✓ VERIFIED | actor.go:578/:1383-1387 Triple→startPhraseCorrection over Phrase()/ReplacePhrase (buffer.go:117/:126); TestActor_TripleTapCorrectsPhrase/EmptyBuffer/PhraseVerifyMismatch green under -race (re-run this pass); 8 phrase rows in matrix-v2, carried into v3 |
| 2 | SC-1: Double Right Shift under an active selection corrects exactly the selected range, both geometric directions (CORR-03, D-30, Pitfall 6) | ✓ VERIFIED | SelectionRange/VerifyRangeAt (verify.go:24/:41); selection branch + selectionState in actor.go; TestActor_DoubleTapSelectionCorrects/SelectionLeftToRight/SelectionMixedConverts green; 5 select rows live green in both the 21/21 (54ea460) and 31/31 (35108412175) runs |
| 3 | Clipboard round-trip rung opt-in (default OFF), order after verify-mismatch, best-effort restore, stdin-only content (D-28/D-29, ADR-003 rung C) | ✓ VERIFIED | clipboard.go:154 cmd.Stdin-only content path, CR-03 kill-vs-empty fix intact; TestActor_ClipboardRungDisabledByDefault/AfterMismatch/OffKeyPath + TestClipboard_KilledNotEmpty/DeadlineKillNotEmpty green; no live rung driver on this desktop (documented exclusion; WINDOWS #4 family closed) |
| 4 | SC-2: mixed text converts by runs — anchor = last letter, foreign runs independently, own runs untouched, digits neutral (CORR-06, D-22/D-23) | ✓ VERIFIED | runs.go:63 ConvertRuns (run scanner unchanged); TestConvertRuns_GoldenCorpus verbatim from CONTEXT specifics — green this pass |
| 5 | Уточнение D-22: homogeneous wholesale (Phase-2 behavior intact); D-24 changed=false = success-without-changes; D-16→D-23 succession (`gfbпривет`→`паипривет`) | ✓ VERIFIED | wholesale branch in runs.go; D-24 `!changed` branches intact at actor.go:1346/:1425 (shifted by additive D-37 lines); matrix-v1 word-mixed row expects паипривет; owner confirmed the corpus reading (WINDOWS #2 resolved, 03-UAT item 2) |
| 6 | Backspace cap D-27: default 50, level 2 not built over cap, zero deletions on refusal (reason backspace-cap) | ✓ VERIFIED | plan.go:30 DefaultBackspaceCap=50, LevelNone; actor.go:1530 skipCorrection("backspace-cap"); TestBuildPlan_BackspaceCap green |
| 7 | SC-3: single Right Shift flips internal mode EN↔RU behind `layoutset`; oracle = mode log record, never gsettings (SWCH-01, D-34/ADR-001) | ✓ VERIFIED | flipScript actor.go:1190; TestActor_FlipOnSingle green; grep gsettings internal/session/ → 0 hits. The 28 gsettings references introduced since the initial pass are ALL in internal/install (Phase-4 install/uninstall lifecycle — reads/restores org.gnome.desktop.input-sources at install time), outside the daemon's switch path; contract intact |
| 8 | SC-3/SWCH-04: tap-timing conflict resolved per ADR-002; window + tap key configurable; FSM corpus semantically untouched (D-35); tap_key and verify_wait_ms actually consumed (CR-01/CR-02) | ✓ VERIFIED | NewFSM(window, tapKeyval) + SetTapKey (fsm.go:66/:85); applySnapshot feeds a.verifyWait (:669-675) and a.tapKeyval (:684); TestFSM_ConfigurableTapKey/SetTapKey + TestActor_HotReloadTapKey/VerifyWait green |
| 9 | SC-3/SWCH-02: combo Shift+RightCtrl corrects the word THEN flips (D-36 order), binding renavigable, series Reset, buffer not fed (Pitfall 4) | ✓ VERIFIED | combo branch above mode branches; settleCombo at every pipeline terminal (actor.go:601/:620/:723/:1342/:1352); TestActor_ComboWordThenFlip/EmptyBufferStillFlips/ConfigurableBinding/DoesNotFeedBuffer + TestFSM_ModifierUse green |
| 10 | SC-3/SWCH-03 (working interpretation, D-34 pre-authorized): native Super mechanism not broken — daemon keeps its IBus name and corrects after the shell chord | ✓ VERIFIED | super-space-alive row live PASS ×4 at acceptance, carried in v2→v3; literal MRU/indicator wording STALE per D-34, not required; gsettings scoping as truth 7 |
| 11 | SC-4/CONF-01: all hotkeys/timeouts/params in YAML (hotkeys/timeouts/correction/macr); strict KnownFields decode, named-field errors, range ceilings (D-31/D-33) | ✓ VERIFIED | load.go:28 KnownFields(true); config.go per-section Validate; TestLoad_UnknownKeyRejected + TestValidate_* green; docs/CONFIG.md complete reference |
| 12 | SC-4/CONF-02: hot reload without restart — dir-watch + debounce, last-good + WARN, one snapshot per event, stale-timer no-re-arm (D-32, Pitfall 8) | ✓ VERIFIED | watch.go:89-90 atomic.Pointer, Reload() :167, "config reload rejected" :237; AttachConfig + applySnapshot; TestWatch_* + TestActor_HotReloadWindowNewSeries/OptionsAndCombo/InvalidKeepsLastGood green |
| 13 | SC-4/CONF-03: Caramba correspondence table in docs, no name cloning, model-difference caveat (D-31) | ✓ VERIFIED | docs/CONFIG.md:105-121 — table maps all keys to Caramba purposes with the first-tap vs window-expiry caveat; unchanged by Phase 4 |
| 14 | SC-4/INST-02: goswitchctl status(--json)/reload/correct against a live daemon; primary-owner name guard; recover shim per method; status carries machine state only, never user text (D-32 visible) | ✓ VERIFIED | ctlsvc.go:37-40 ErrNotPrimaryOwner, :91/:102/:120 recoverMethod on all 3 methods, :166 recover shim; TestSvc_Status/ReloadConfig(×3)/CorrectNow/NameGuard (over a real private dbus-daemon)/RecoverShim green; ctl-smoke live ×4 at acceptance. Phase-4 D-37 ADDS a leading `version=` token to renderStatus — additive, comment re-affirms «Counts and states only — never user text (T-03-06-03)»; D-32 invariant intact. The initial pass's narrower wording ("mode/counters/config only") is superseded by the later roadmap decision D-37, not broken |
| 15 | SC-5/MACR-01: implemented per ADR-005 — interception above combo/mode branches (RU pin), burst on Super release, consumed-upstream WARN + counters, per-app a11y observer with degradation ladder + lazy start, alt_modifier default-off (b.3) | ✓ VERIFIED | actor.go:123-124/:999 macrPendingKeyval/macrIntercepted; appid.go:41/:163 ErrBusClosed/FocusedApp; TestActor_MACR* corpus green; live: probe verdicts + wire-relay oracle + per-app ×3; GTK4 client boundary = documented platform limitation (WINDOWS #4 closed by owner; REQUIREMENTS.md MACR-01 flipped Complete with caveat) |
| 16 | SC-5: e2e-matrix green in full breadth (21 v2 cases: 8 phrase incl. registers + 70-rune A3 pin, 5 selection rows across 3 surfaces both geometries, mixed, combo, single, super-space, 3 reload rows, MACR; clipboard row excluded with the documented spike-verdict comment) — **at the verified HEAD** | ✓ VERIFIED (closed this pass) | Previously ⚠ (green runs predated the review-fix commits). Now: (a) 03-UAT.md item 1 — owner-authorized live `mise run e2e-matrix-v2` 21/21 PASS at HEAD 54ea460 (post 7152cd3..d94a3ce fixes); (b) D-48 double-run gate — CI 35108412175 re-fetched success at 9f2fd75, BOTH matrix-v3 runs 31/31; v3 verified a strict superset (comm -23 of case names empty); production code identical 9f2fd75→HEAD (only docs + goswitchctl parser test corpus after). matrix-v2.yaml still 21 cases on disk |
| 17 | anchorPos seam carries the anchor to the actor on every surface; per-surface actuals spike-pinned (Pitfall 1; A1/A2/A5) | ✓ VERIFIED | engine.go:87-89/:190-195 signature + wire call + DEBUG trace; implementors unchanged; zenity verdict corrected + owner-acknowledged (03-UAT item 4); select-all-zenity row green in both matrix runs |
| 18 | D-30 continuity: no selection → the Phase-2 word path unchanged; Phase-2 corpus green without expectation edits | ✓ VERIFIED | TestActor_NoSelectionStillWord present and green in this pass's -race run; v1 matrix 16/16 at acceptance; v1 path still routed in the workflow |
| 19 | Key/binding provenance: closed name tables from ibuskeysyms.h; KeyV=0x076, KeyControlR=0xffe4, KeySuperL/R; ParseBinding + FamilyMask press-side rule | ✓ VERIFIED | names.go; engine/keys.go:36-51 with ibuskeysyms.h provenance comments; TestParseBinding/TestParseBinding_KeyvalProvenance/TestFamilyMask green |
| 20 | CI circuit: workflow routes the full-breadth matrix by input (v1 AND v2 paths preserved); first green run recorded; mise task registry complete | ✓ VERIFIED (wording superseded) | Initial wording was "defaults to matrix-v2.yaml". Phase-4 D-48 legitimately moved the default input to matrix-v3.yaml (e2e-matrix.yml:44) while PRESERVING the v1 and v2 routes (:133-138, :149-152); mise now carries 32 e2e-* task lines. Substance — CI circuit wired to the full-breadth matrix with a recorded green run — holds and is stronger: green v3 31/31 ×2 at 35108412175 (re-fetched) on top of the green v2 21/21 at 54ea460 |

**Score:** 20/20 truths verified (0 present-but-behavior-unverified)

Behavior-dependent truths note: every state-transition/invariant truth in
this phase (phrase/selection pipelines, verify-after debounce, hot-reload
window arming, combo settlement fan-out, MACR release-time burst, clipboard
kill-vs-empty, name guard over the wire) is covered by a named behavioral
test re-run green under -race at the current HEAD by this verifier; the
live-desktop dimension is covered by the green v2 21/21 run at 54ea460
(UAT) AND the green v3 31/31 double-run at 9f2fd75 (CI 35108412175,
re-fetched) — both postdating every review fix.

### Decision Coverage

`gsd_run query check.decision-coverage-verify` (03-CONTEXT.md), re-run this
pass: **15/15 decisions honored** by shipped artifacts, none missing
(non-blocking gate).

### Advisory (New Scope, Unevidenced)

Re-verification ran; the Step 7 scan over files modified since the previous
verification (actor.go, ctlsvc.go, both cmd mains, matrix.go, main.go,
preflight.go, e2e-matrix.yml, mise.toml, go.mod/go.sum) found zero debt
markers, zero stubs, zero empty implementations. Two informational notes,
neither evidence-bearing against phase-3 must-haves:

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| 1 | golangci-lint v2.13+ warns `exhaustruct` is deprecated (replaced by `exhaustruct_v5`); .golangci.yml still names the old linter | other (tooling config) | New since the initial pass (newer linter release); zero lint issues either way; config belongs to the repo-wide quality gate, not a phase-3 truth. Resolve in a housekeeping pass |
| 2 | internal/install now shells out to gsettings (28 refs) — the initial pass's "zero gsettings in internal/" grep no longer holds verbatim | other (scope note) | By design: Phase-4 D-39 install lifecycle reads/restores input sources at INSTALL time; the switch path (internal/session) still has zero gsettings, so truths 7/10's contract is intact. Recorded so future greps scope correctly |

### Required Artifacts

`gsd_run query verify.artifacts` per plan at initial verification: all 7
plans valid. This pass: all 67 covered files exist on the current tree
(existence check before fingerprinting); the phase-4 diff touched only
additive surfaces of them (see Re-verification Scope). Substantiveness
re-confirmed by grep of every pinned symbol (per-truth evidence above);
zero stub markers.

### Key Link Verification

Initial pass: 5/7 plans fully verified by the tool; the 3 tool-shape false
negatives manually disproven (ctlsvc→session direction, goswitchctl bus
name constants, matrix.yaml↔matrix.go strict schema). Re-checked this pass
on the current tree:

- ctlsvc.go still returns `session.Status` (:124 area) — direction intact.
- cmd/goswitchctl bus name/path constants intact (selfcheck and version
  additions did not touch them).
- matrix.go kind vocabulary (:154-158 matrixKindType..Select) still parses
  matrix-v2.yaml (21 cases decode; TestMatrix corpus green) and now also
  matrix-v3.yaml (31 cases; new select/combo/reload rows use the same
  closed vocabularies).
- All other links unchanged: actor→ConvertRuns (zero `Detect(` in
  actor.go), actor→engine emitters, main→AttachConfig, main→ctlsvc.Run.

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| internal/session/actor.go | correction ranges | buf.Phrase()/activeRange + selectionState from live SetSurroundingText pushes | ✓ select rows green in both recorded matrix runs | ✓ FLOWING |
| internal/session/actor.go | config snapshot | Watcher.Snapshot() over atomic.Pointer (fsnotify dir events / ctl Reload) | ✓ reload rows change window/verify_wait live in recorded runs | ✓ FLOWING |
| internal/session/actor.go | MACR counters | macrIntercepted/macrConsumed increments on live key events | ✓ super-intercept gated in macr row, carried in v3 | ✓ FLOWING |
| internal/session/actor.go | version (NEW, D-37) | daemon build identity pinned via SetVersion at construction | ✓ additive; unit-pinned in the new corpus | ✓ FLOWING |
| internal/ctlsvc/ctlsvc.go | status line | renderStatus ← StatusSnapshot (now leading with version=) + watcher LastError/ConfigPath | ✓ corrections_done observed live in ctl-smoke at acceptance | ✓ FLOWING |
| test/e2e/matrix.go | report | per-case runMatrixCaseIsolated results over real daemon+surfaces | ✓ 21 PASS lines (54ea460) + 31/31 ×2 (CI 35108412175) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Build | `go build ./...` | exit 0 | ✓ PASS |
| Full headless corpus (single full-suite run) | `go test ./... -race -count=1` | 14/14 packages ok (incl. new internal/install) | ✓ PASS |
| Vet | `go vet ./...` | clean | ✓ PASS |
| Lint (strict) | `mise exec -- golangci-lint run` | 0 issues (upstream exhaustruct deprecation warning only) | ✓ PASS |
| Named behavioral tests exist | `go test ./internal/... -list '.*'` (enumeration only) | all 15 sampled pins present: TripleTapCorrectsPhrase, DoubleTapSelectionCorrects, NoSelectionStillWord, ConvertRuns_GoldenCorpus, FlipOnSingle, ComboWordThenFlip, HotReloadTapKey/VerifyWait, ClipboardRungOffKeyPath, Clipboard_KilledNotEmpty, FSM_SetTapKey, MACRIntercepts, Load_UnknownKeyRejected, Svc_NameGuard, VerifyAfterLevel1 | ✓ PASS |
| CI matrix run (recorded, re-fetched) | `gh run view 35108412175` | conclusion success, workflow_dispatch, headSha 9f2fd75 (D-48 double-run gate: both v3 runs 31/31) | ✓ PASS |
| Live e2e re-run | not re-run (drives owner's desktop) | recorded UAT 21/21 at 54ea460 + CI 35108412175 cited | ? SKIP (by policy) |
| Matrix v2 breadth preserved | case-name diff v2 vs v3 | v2 still 21 cases; v3 = strict superset (31; all 21 v2 names present) | ✓ PASS |

### Probe Execution

Step 7c: no `scripts/*/tests/probe-*.sh` convention in this repo. The
phase's runnable probes remain the mise e2e tasks (live-session; covered by
the recorded-evidence policy above) and the CI gate (re-run this pass:
build + vet + lint + full -race suite, all green). The matrix CI run is the
phase-declared probe target — green at 9f2fd75 (35108412175) and, for v2
specifically, at 54ea460 (UAT).

### Test Quality Audit

Unchanged from the initial pass (Phase 4 added new test files — install,
selfcheck, perf, watchdog, goswitchctl/main — which belong to Phase 4's
verification): phase-3 requirement-linked corpora remain all-active,
non-circular, value/behavioral-level. This pass re-confirmed: the single
env-guard skip in TestSvc_NameGuard still executed (dbus-daemon present);
expected values remain hand-authored literals (ghbdtn/привет/паипривет);
matrix oracles rune-exact. Disabled tests on requirements: **0**. Circular
patterns: **0**.

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
| MACR-01 | 03-05, 03-07 | Super+letter → Ctrl+letter per ADR-005 | ✓ SATISFIED (Complete with GTK4 caveat — owner decision, UAT item 3) | truth 15; REQUIREMENTS.md:45/:108 now aligned (flipped Complete with the platform caveat, footer refreshed) |
| INST-02 | 03-06, 03-07 | goswitchctl status/reload/correct | ✓ SATISFIED | truth 14 |

Orphaned requirements: **none** — REQUIREMENTS.md traceability maps exactly
these 12 IDs to Phase 3; every plan requirement field resolves to them.
The initial pass's bookkeeping warning (MACR-01 checkbox Pending) is
RESOLVED: commit 9069287 flipped it Complete with the caveat.

### Anti-Patterns Found

Debt-marker gate re-run on all phase-3 covered files modified since the
previous verification: **ZERO** TBD/FIXME/XXX and zero TODO/HACK/PLACEHOLDER.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| repo root `./e2e` | — | Stray untracked Go build artifact, still present locally (rebuilt 2026-09-16 by Phase-4 e2e activity) | ℹ️ Info | RESOLVED at repo-hygiene level: now gitignored (`.gitignore:38` `/e2e`) — cannot be committed |
| .golangci.yml | — | `exhaustruct` linter name deprecated upstream since v2.13.0 (→ exhaustruct_v5) | ℹ️ Info | Tooling housekeeping; 0 lint issues today; not a phase-3 regression (see Advisory) |
| internal/session/actor.go | logCorrectionDone | source/result words at DEBUG (review IN-01) | ℹ️ Info | Unchanged by Phase 4; D-21-permitted debug channel, -debug flag warns; INFO records content-free |
| engine/engine.go | Engine.caps | write-only field (review IN-02) | ℹ️ Info | Unchanged (engine untouched by Phase 4); dead state carried as review Info debt |
| internal/hotkey/names.go | ParseBinding | degenerate bindings accepted (review IN-04) | ℹ️ Info | Unchanged; harmless (WR-02 closed the tap_key side) |
| test/e2e/README.md | — | -case table stale vs registry (review IN-03) | ℹ️ Info | Docs-only; unchanged by Phase 4 (which updated the perf table section) |

The 8 in-scope review findings (CR-01..03, WR-01..05) remain FIXED with
RED→GREEN pins, re-confirmed green in this pass's -race run.

Prohibition audit: the initial pass's 15 judgment-tier verdicts carry
forward — all programmatically-upheld ones re-confirmed this pass (no
autocorrection without a gesture — startCorrection still only from FSM
Double/Triple expiry, combo branch, CorrectNow; INFO content-free;
KnownFields strict; last-good + WARN; clipboard content stdin-only, never
logged/argv; selection exactness pinned; zero gsettings in the switch path;
TestFSM_ModifierUse green; consumed-upstream WARN + counters; status
surface counts/states/version only). The formerly live-only judgments
(teardown restores the desktop; matrix never red from stand fault;
session-bus same-uid isolation) now rest on TWO green recorded runs
(21/21 at 54ea460 per UAT, and 31/31 ×2 at 35108412175 with the D-48
fresh-session preflight + double-run gate) — evidence materially stronger
than at the initial pass.

### Human Verification Required

**None — all 4 items from the initial verification are RESOLVED**, recorded
in 03-UAT.md (status: complete, 4 passed, 0 issues):

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

**No gaps.** All 20 must-haves verified on the current tree (HEAD 42d10a1):
the phase-4 diff touched phase-3 code only additively (D-37 version
identity) and every pinned contract survives — the switch path's gsettings
abstinence, the D-32 counts-and-states-only status invariant (now with a
leading version token), the strict-decode/hot-reload config core, the
phrase/selection/runs/combo/MACR pipelines, and the matrix breadth (v2
intact at 21 cases and provably subsumed by the green v3 31/31 double-run).
Headless gates re-run green at HEAD (build, full -race suite 14/14, vet,
strict lint 0 issues). The evidence-currency gap and all four human items
from the initial pass are closed by recorded evidence (03-UAT.md 4/4; CI
35108412175 re-fetched success). Status advances from human_needed to
**passed**; the fingerprint is refreshed over all 67 covered files.

---

_Verified: 2026-09-16T19:19:27Z_
_Verifier: Claude (gsd-verifier)_
