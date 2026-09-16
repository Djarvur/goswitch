---
phase: 02-korrektsiya-slova-en-ru
verified: 2026-09-16T19:19:44Z
status: passed
score: 23/23 must-haves verified
covered_files:
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
  - .github/workflows/e2e-matrix.yml
  - .gitignore
  - SECURITY.md
  - cmd/goswitchd/main.go
  - docs/ci-runner.md
  - engine/emitter_wire_test.go
  - engine/engine.go
  - engine/engine_test.go
  - engine/factory.go
  - engine/keys.go
  - engine/types.go
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
  - internal/hotkey/fsm.go
  - internal/hotkey/fsm_test.go
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
covered_digest: "v1:sha256:019420070eb5a11083f2cf3b057e97d7f4b1e05d63e9eb27fbb89fa47d57ba51"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 23/23
  previous_verified: 2026-09-15T17:59:16Z
  reason: "fingerprint refresh after Phase 4 execution (commits 54ea460..42d10a1): D-37 version identity + e2e v3/gedit/x11/perf/D-48 infrastructure additively reworked covered files; phase-2 correction core untouched"
  gaps_closed: []
  gaps_remaining: []
  regressions: []
  closed_deferred_items:
    - "deferred-items #2 (verify-after settles on stale first push → false mismatch INFO) — CLOSED by WR-05 (d94a3ce, phase 3): first stale answer re-requires, only the second concludes; pinned by TestActor_VerifyAfterLevel1 subtests, re-run PASS at HEAD 42d10a1"
  still_open_deferred_items:
    - "deferred-items #1 (persistent AT-SPI shell-bridge wedge) — environmental, unchanged"
    - "deferred-items #3 (case_m1.go:135 ydotool \"Escape\" name trap in the unexercised shell-surface fallback) — unchanged, warning stands"
human_verification: [] # both prior owner items resolved 2026-09-15 in 02-UAT.md (2/2 pass, owner-accepted) — final, not re-opened; phase 4's own pending UAT (04-UAT.md) is phase-4 scope, not phase-2
---

# Phase 2: Коррекция слова EN↔RU — Verification Report

**Phase Goal (ROADMAP.md):** «Пользователь двойным нажатием Right Shift исправляет последнее слово, набранное не в той раскладке, в любом поле ввода GNOME — направление определяется автоматически, регистр сохраняется, замена происходит ровно по диапазону неверного текста без визуального прыжка; это подтверждает зелёная e2e-матрица v1.»
**Verified:** 2026-09-16T19:19:44Z
**Status:** passed
**Re-verification:** Yes — fingerprint-refresh after Phase 4 (prior pass: 2026-09-15T17:59:16Z post-Phase-3, 23/23; original pass 2026-09-15T05:40:42Z, UAT 2/2 owner-accepted; no prior gaps)

## Re-verification Approach

Phase 4 (7 plans, commits 54ea460..42d10a1) advanced the tree through
delivery/acceptance. Every covered file was classified with
`git diff 54ea460..HEAD` (54ea460 = the phase-3-close commit that recorded
the prior `passed` fingerprint):

- **Touched → targeted re-check** (all additive to phase-2 semantics):
  `internal/session/actor.go` (+ D-37 `version` field, `SetVersion`,
  `Status.Version` lift — correction pipeline untouched; all 16 named
  behavioral tests re-run PASS under -race),
  `cmd/goswitchd/main.go` (+ `-version` flag, `actor.SetVersion(version)`
  call — additive, before `run()`),
  `test/e2e/{main,matrix,preflight,surface}.go` (+ v3 matrix, gedit /
  chromium-x11 surfaces, fresh-session preflight, watchdog — v1/v2 paths and
  the strict `KnownFields` loader preserved),
  `test/e2e/case_{d01,m1,resilience,surface}.go` (+ gedit/x11 branch arms),
  `test/e2e/focus_helper.py`, `mise.toml` (+ install/release/perf/v3 tasks;
  e2e-matrix v1/v2/v3 tasks all present),
  `.github/workflows/e2e-matrix.yml` (D-48 double-run gate, v3 default —
  explicit v1/v2 regression dispatch paths kept; WR-04 env-input discipline
  kept), `.gitignore`, `docs/ci-runner.md` (D-48 procedure).
- **Untouched → regression check** (existence + sanity; `git diff
  54ea460..HEAD` empty): all 14 PLAN/SUMMARY files, the whole correction
  core `internal/correct/{buffer,convert,direction,plan,runs,verify}.go`
  (+tests), `engine/*` (+tests), `internal/hotkey/fsm.go`, `test/e2e/case_word.go`,
  `case_ladder.go`, `cases/matrix-v1.yaml` (16 cases, D-13 pin-geometry
  header intact), `cases/matrix-v2.yaml`, `matrix_test.go`, `README` fixtures.
  The D-13 pin geometry, ladder levels and buffer semantics live where the
  prior pass pinned them: `internal/correct` + `internal/session` + `engine`.

The D-23 mixed-word succession, direction-anchor mechanism, WR-05 debounce,
CR-01/CR-02 config plumbing and WR-01 rung semantics documented by the prior
pass are unchanged by phase 4 — their pinned tests re-ran green (below).

## Live-gate evidence policy (unchanged)

Live e2e (`mise run e2e-matrix*`) drives the owner's desktop — NOT re-run by
the verifier (Phase 1/2/3 precedent). Machine-generated GitHub-side
evidence, independently re-fetched 2026-09-16:

| Run | Workflow | headSha | Result | Relevance |
|-----|----------|---------|--------|-----------|
| **35108412175** | e2e-matrix (dispatch) | **9f2fd75** (2026-09-16T14:25; between it and HEAD 42d10a1 only docs/planning commits + one test-only commit 100af04 — **production code identical to HEAD**) | success, **D-48 double-run: 31/31 PASS TWICE** (log re-fetched: both `matrix: 31/31 PASS` lines; phrase-en-ru/ru-en/mixed/registers/after-space, long-phrase, phrase-gte, phrase-chromium, word-mixed, 5 select rows, layout-single, combo-word-layout, super-space-alive, 3 reload rows, macr-super-letter, gedit word/word-upper/word-capital/phrase/phrase-registers/layout-single, x11 word-en-ru/word-upper/word-capital/phrase) | The strongest live evidence yet for phase-2 semantics: every word/phrase/register/exact-range/surface invariant green at HEAD-equivalent code, twice consecutively on one session (D-48: run #1 does not poison run #2). Includes the successed `word-mixed` (D-23) and FocusOut reset (`super-space-alive`). |
| 35107406144 | e2e-matrix (dispatch) | fbd3072 (2026-09-16T14:16) | failure | The first D-48 dispatch — killed by the 04-06 watchdog bug (`time.After(0)` when the limit input was absent), fixed in 9f2fd75 and re-run green as 35108412175. Documented honestly; not a phase-2 regression (the failure was in matrix-runner plumbing, fixed same day). |
| 34984814995 | e2e-matrix (dispatch) | 21df37c (2026-09-15, phase 3) | success, 21/21 PASS | Superseded as primary evidence by 35108412175; retained in history. |
| 34913812782 | e2e-matrix (dispatch) | 419799d (2026-09-15, phase 2 code state) | success, 16/16 PASS incl. three GTE registers, reset-escape/enter/tab/focus rows, fallback-chromium | Still the ONLY live run of the literal v1 file's reset rows — matrix-v1.yaml is byte-identical since, so this remains the reset-row live witness. All other v1 semantics are re-covered live by v3 above. |
| pr-sanity | — | latest 35130001841 / 35129888432 @ 80d8534 / f37dc57 (2026-09-16) | success | pr-sanity circuit alive and green (build + vet + lint + test -race + tidy + govulncheck). |
| runner | — | — | green106 online, labels self-hosted,Linux,X64,gnome | CI circuit alive (re-checked via API 2026-09-16). |

Headless gates re-run by the verifier at HEAD 42d10a1 — all green:
`go build ./...` OK; `go vet ./...` OK; `go test -race -count=1 ./...` all
**14** packages ok (engine, appid, clipboard, config, correct, ctlsvc,
hotkey, install, logging, session, layouts, cmd/goswitchctl, cmd/goswitchd,
test/e2e — 3 new packages from phase 4, all additive); `mise run lint`
0 issues; `go mod tidy` + `git diff` — no drift.

## User Flow Coverage

User capability (goal, non-story form — MVP user-flow framing discrepancy
recorded in the initial pass, unchanged).

| Step | Expected | Evidence | Status |
|------|----------|----------|--------|
| Type a word in the wrong layout | `ghbdtn` (registers/mixed/digits variants) lands in the field | matrix-v1.yaml 16 cases intact (byte-identical); case_word.go registry wired (main.go:244–247 re-checked) | ✓ |
| Double-tap Right Shift | FSM Double → correction pipeline | TestActor_DoubleTapCorrects **re-run PASS at HEAD 42d10a1**; NewActor default KeyvalShiftR; live in D-48 double-run @9f2fd75 | ✓ |
| Word corrected, direction automatic | EN→RU and RU→EN; mixed → foreign run (D-23) | TestActor_MixedWordConvertsForeignRuns PASS; live word-mixed + phrase-en-ru/ru-en + gedit/x11 word rows (both directions) PASS @9f2fd75; ConvertRuns anchor | ✓ |
| Register preserved | three registers | TestConvert_CasePreserved green (20-test correct corpus under -race); live phrase-registers + gedit-phrase-registers PASS @9f2fd75; v1 GTE registers live @419799d | ✓ |
| Exact-range replacement, no jump | delete+commit over the range, one transaction | BuildPlan rune arithmetic + TestActor_AfterSpaceCorrects re-run PASS; live phrase-after-space PASS @9f2fd75; perceptual check owner-accepted (UAT 2) | ✓ |
| Works in any GNOME input field | GTK/Chromium/GtkSourceView — now also gedit + chromium-X11 drivers | live v3 31/31 twice @9f2fd75: zenity/GTE/chromium/gedit/x11 rows all PASS | ✓ |
| Green matrix proves it | 16/16 v1; exit≠0 | infra green at HEAD (matrix_test.go under -race, strict loader intact); live artifacts cited above | ✓ |

## Goal Achievement

### Observable Truths

Same 23 consolidated truths as the accepted prior passes; evidence refreshed
to the HEAD tree 42d10a1 (R = regression-checked, untouched since 54ea460;
D = deep-checked, phase-4-touched file; succession wording where a later
phase superseded a mechanism).

| # | Truth | Status | Evidence at HEAD 42d10a1 |
|---|-------|--------|--------------------------|
| 1 | SC-1: gnome-text-editor double Right Shift corrects `ghbdtn`→`привет`, `GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет` | ✓ VERIFIED | v1 rows intact (matrix-v1.yaml, R); live 16/16 @419799d; register semantics re-proven live at phrase width (phrase-registers, gedit-phrase-registers @9f2fd75) + unit TestActor_DoubleTapCorrects/TestConvert_CasePreserved at HEAD |
| 2 | SC-2: direction auto-detected from buffer composition both ways + unit corpus | ✓ VERIFIED (mechanism succession) | direction via ConvertRuns last-letter anchor (runs.go:63–91, R — untouched); single-script → wholesale Convert — TestActor_DoubleTapCorrects/word cases re-run PASS; live both directions @9f2fd75; TestDetect_* corpus green (Detect remains corpus-reference, see Anti-Patterns) |
| 3 | SC-3: replacement exactly over the wrong range, both ladder levels, abort-not-litter | ✓ VERIFIED | BuildPlan rune arithmetic intact (plan.go, R); D-27 level-2-only cap pinned; level-1 live on 5 surfaces now (v3 @9f2fd75); level 2 unit-pinned re-run PASS; abort discipline TestActor_VerifyPaths green; Chromium actual level owner-accepted (UAT 1) |
| 4 | SC-4: buffer cleared on Enter, Tab, Escape, focus change | ✓ VERIFIED | isResetKeyval table re-read at HEAD (Return/KPEnter/Tab/Escape) + HardReset unchanged; TestActor_HardResetKeyvals / TestBuffer_ResetByFocusOut **re-run PASS**; live reset-* @419799d + live FocusOut reset (super-space-alive) @9f2fd75 |
| 5 | SC-5: e2e-matrix v1 green — YAML cases, PASS/FAIL report, exit≠0 | ✓ VERIFIED | matrix.go strict loader (`KnownFields(true)` re-read at HEAD, D — phase-4 additions additive) + per-case isolation + exitCode green (matrix_test.go under -race); v1 file 16 cases intact; live v1 16/16 @419799d; v1 semantics live-green in v3 31/31 twice @9f2fd75 |
| 6 | Token semantics D-13/D-14/D-15 + toggle invariant | ✓ VERIFIED | buffer.go Push/Token/Tail/ReplaceToken untouched (R); buffer corpus green under -race |
| 7 | Mixed word (CORR-04) — D-23 succession (was D-16 refusal) | ✓ VERIFIED | mixed word converts foreign runs (паипривет); oracle intact in case_word.go (R) + matrix v1 word-mixed row; live v3 word-mixed PASS @9f2fd75; digits/punctuation neutral (D-15 continuation) |
| 8 | Conversion preserves register per rune; non-letters identical (CORR-05) | ✓ VERIFIED | convert.go untouched (R); ConvertRuns reuses Convert as-is (never forked); TestConvert_CasePreserved green; live registers at phrase width + gedit @9f2fd75 |
| 9 | Suffix verification ADR-004 (token+tail range) | ✓ VERIFIED | MatchesSuffix intact (verify.go, R); VerifyRangeAt for selection (additive); wired at pendingVerdict/afterVerdict; corpus green |
| 10 | BuildPlan both levels, all counts in runes (CORR-07) | ✓ VERIFIED | len([]rune) arithmetic intact (R); +backspaceCap param (D-27); TestBuildPlan corpus green (part of 20-test correct re-run) |
| 11 | internal/correct purity: no godbus/time/goroutines | ✓ VERIFIED | purity grep re-run at HEAD: **0 hits** in production files |
| 12 | Surface drivers: fresh chromium + standalone GTE, witness-gated | ✓ VERIFIED | surface.go reworked additively (gedit/x11 drivers added, existing zenity/chromium/GTE drivers untouched); smoke cases in registry; live gedit/x11/smoke rows green @9f2fd75 |
| 13 | Preflight rejects missing binaries | ✓ VERIFIED | preflight.go extended (fresh-session preflight additive; existing binary checks intact); package green under -race; live preflight 6× ok in D-48 log |
| 14 | Tracer: double tap runs buffer→direction→convert→verify→delete+commit (CORR-01) | ✓ VERIFIED | TestActor_DoubleTapCorrects **re-run PASS**; live word/phrase cases in D-48 double-run |
| 15 | Word after space corrected with separator (D-13) | ✓ VERIFIED | TestActor_AfterSpaceCorrects **re-run PASS**; live phrase-after-space PASS @9f2fd75 |
| 16 | Log discipline D-20/D-21 | ✓ VERIFIED | logCorrectionDone; TestActor_DebugCorrectionRecord green in re-run; D-20 also re-pinned by ClipboardRungAfterMismatch subtest («no content in INFO») |
| 17 | Flip on Single: mode EN↔RU, INFO mode record, XKB untouched | ✓ VERIFIED | flipScript; TestActor_FlipOnSingle **re-run PASS**; live layout-single + gedit-layout-single @9f2fd75 |
| 18 | RU consumption commits Cyrillic; script-true buffer in ALL printing branches incl. digits | ✓ VERIFIED | feedKey unchanged invariants; TestActor_RUDigitsFullPipeline **re-run PASS**; live phrase-ru-en @9f2fd75 |
| 19 | Level 2: ForwardKeyEvent(BackSpace)×(token+tail runes) burst → one commit, in order (CORR-07) | ✓ VERIFIED | executeLevel2 + emitter wire pin; TestActor_Level2NoCaps/WithTailAndRunes **re-run PASS**; over-cap range refuses (D-27, LevelNone) |
| 20 | Verify-after with epoch: mismatch → INFO counter, never auto-repair; stale timers no-op | ✓ VERIFIED (WR-05 semantics) | settleAfter: first stale → `stale-retry` + re-Require, second stale → mismatch once; genuine mismatch NOT masked; epoch guard retained; TestActor_VerifyAfterLevel1 3 subtests **re-run PASS at HEAD** (match quiet / mismatch-once-no-repair / intermediate-push debounced); deferred item #2 stays CLOSED |
| 21 | Reset triggers wired live (CORR-09) | ✓ VERIFIED | reset-* live @419799d; FocusOut reset live @9f2fd75; unit re-run PASS |
| 22 | Matrix infrastructure: strict decoder, per-case isolation, closed vocabularies | ✓ VERIFIED | matrix.go KnownFields + v3 step kinds added additively (R for v1/v2 vocab); matrix_test.go + watchdog_test.go green at HEAD under -race |
| 23 | CI circuit D-19: dispatch-only workflow on self-hosted gnome runner | ✓ VERIFIED | e2e-matrix.yml re-checked (D-48 double-run added; dispatch-only trigger, permissions contents:read, WR-04 env-input kept); run 35108412175 success at HEAD-equivalent code; runner green106 online; pr-sanity green 2026-09-16 |

**Score:** 23/23 truths verified (0 present-but-behavior-unverified)

Behavior-dependent truths note: every state-transition/invariant truth is
covered by a named behavioral test re-run by the verifier under -race at
HEAD 42d10a1 (all PASS — spot-check table below); the live-desktop dimension
now rests on the D-48 double-run at production-code-identical-to-HEAD
(9f2fd75) — the strongest recency of any pass so far.

### Required Artifacts

`verify.artifacts` re-run per plan at HEAD: **27/28** — 02-01 5/5, 02-02 4/4,
02-03 4/4, 02-04 3/3, 02-05 4/4, 02-06 4/4, 02-07 3/4. The single "miss" is
the same verifier-tool `~` false negative as both prior passes —
`~/.config/systemd/user/goswitch-ci-runner.service` exists (750 bytes,
re-checked 2026-09-16). All artifacts substantive.

### Key Link Verification

`verify.key-links` re-run per plan at HEAD: **15/16**. The one unverified
link is the same plan-spec false negative as both prior passes (02-01
buffer.go→direction.go, pattern `\.Token\(\)`); the actual consumer wiring
is `internal/session/actor.go` (`a.buf.Token()` → `startRangeCorrection` →
ConvertRuns) — present, wired, behavior-tested.

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| internal/session/actor.go | token/tail | buf.Token()/Tail() fed by HandleKey from live ProcessKeyEvent | ✓ live corrections in the D-48 double-run @9f2fd75 | ✓ FLOWING |
| internal/session/actor.go | surr/sel | HandleSurroundingText live pushes + Require rounds | ✓ live word/phrase/selection corrections verified @9f2fd75 | ✓ FLOWING |
| internal/correct/runs.go | out/changed | ConvertRuns anchor scan over the live range | ✓ word-mixed live PASS @9f2fd75 | ✓ FLOWING |
| internal/correct/plan.go | Plan | BuildPlan from token/tail/converted + caps bit | ✓ real caps 0x29 live; level dispatch exercised | ✓ FLOWING |
| test/e2e/matrix.go | report | per-case runMatrixCaseIsolated results | ✓ 31/31 PASS twice in re-fetched D-48 log | ✓ FLOWING |
| layouts/tables.go | ENToRU/RUToEN | xkb-generated (Phase 1) | ✓ RU-commit path live (phrase-ru-en, gedit/x11 rows) | ✓ FLOWING |

### Behavioral Spot-Checks (re-run at HEAD 42d10a1, all -race -count=1)

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full corpus, all packages | `go test -race -count=1 ./...` | 14 pkgs ok (exit 0) | ✓ PASS |
| Default double-tap pipeline | `TestActor_DoubleTapCorrects` | PASS | ✓ PASS |
| WR-05 verify-after debounce (match / mismatch-once / intermediate-push) | `TestActor_VerifyAfterLevel1` (3 subtests, verbose-confirmed) | PASS | ✓ PASS |
| Exact range + tail separator | `TestActor_AfterSpaceCorrects` | PASS | ✓ PASS |
| Level-2 ladder | `TestActor_Level2NoCaps` / `Level2WithTailAndRunes` | PASS | ✓ PASS |
| Reset keyvals + FocusOut | `TestActor_HardResetKeyvals` / `TestBuffer_ResetByFocusOut` | PASS | ✓ PASS |
| Flip / RU consumption | `TestActor_FlipOnSingle` / `RUDigitsFullPipeline` | PASS | ✓ PASS |
| D-23 mixed word | `TestActor_MixedWordConvertsForeignRuns` | PASS | ✓ PASS |
| CR-01 tap key (default + hot swap) | `TestActor_HotReloadTapKey` (real-timer subtest) | PASS | ✓ PASS |
| CR-02 verify budget | `TestActor_HotReloadVerifyWait` | PASS | ✓ PASS |
| WR-01 rung off key path + after mismatch | `TestActor_ClipboardRungOffKeyPath` / `ClipboardRungAfterMismatch` | PASS | ✓ PASS |
| Abort discipline / log record | `TestActor_VerifyPaths` / `TestActor_DebugCorrectionRecord` | PASS | ✓ PASS |
| correct-package corpus (registers, plans, verify ranges, runs) | 20 named tests, `-run 'TestConvert_CasePreserved|TestDetect|TestBuildPlan|TestPlan|TestBuffer|TestVerifyRangeAt|TestMatchesSuffix|TestRunsConvert'` | 20/20 PASS | ✓ PASS |
| Gates | build / vet / lint / tidy | OK / OK / 0 issues / no drift | ✓ PASS |
| Live e2e re-run | not re-run (drives owner's desktop) | D-48 double-run 31/31×2 @9f2fd75 re-fetched | ? SKIP (by policy) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` convention; the phase's runnable probes are
the mise e2e tasks (live-session; covered by the CI-artifact policy) and the
CI gate (re-run, PASS). Not applicable beyond that.

### Test Quality Audit

Disabled tests on Phase 2 requirement-linked files: **0** (the only
`t.Skipf` in the tree remains the ctlsvc environmental dbus-daemon guard —
not requirement-linked here; phase 4 added no skips on phase-2 files).
Circular patterns: **0** (hand-authored oracles; layout tables regenerate
from system xkb). Assertion strength: rune-exact equality / op-log order /
log-shape scans — no existence-only assertions on requirement-linked tests.

### Requirements Coverage

| Requirement | Source Plan | Status | Evidence |
|-------------|------------|--------|----------|
| CORR-01 | 02-03, 02-04 | ✓ SATISFIED | truths 14, 17; live word/phrase both directions incl. gedit/x11 @9f2fd75 |
| CORR-04 | 02-01, 02-04 | ✓ SATISFIED | truths 2, 7, 18; direction by composition via ConvertRuns anchor |
| CORR-05 | 02-01 | ✓ SATISFIED | truth 8; registers live at phrase width + gedit @9f2fd75 + v1 GTE registers @419799d |
| CORR-07 | 02-01, 02-03, 02-05 | ✓ SATISFIED | truths 3, 9, 10, 15, 19, 20 |
| CORR-09 | 02-05 | ✓ SATISFIED | truths 4, 21 |
| TEST-04 | 02-02, 02-06, 02-07 | ✓ SATISFIED | truths 5, 12, 22, 23; REQUIREMENTS.md mapping unchanged (re-read at HEAD) |

Orphaned requirements: **none** (REQUIREMENTS.md maps exactly these 6 IDs,
all Complete).

### Anti-Patterns Found

Debt-marker gate re-run at HEAD 42d10a1: **ZERO** TBD/FIXME/XXX and zero
TODO/HACK/PLACEHOLDER across all phase-covered Go/Python/YAML/TOML/workflow
files.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| internal/correct/direction.go | — | `Detect` production-orphaned by the D-23 succession (re-confirmed: only definition, zero production callers; runs.go cites it as the classifyRune reference; corpus stays green) | ⚠️ Warning | Dead code carried by its own tests; SC-2 contract unaffected (anchor scan proves it live). Candidate: annotate as corpus-reference or fold into the run scanner — owner cleanup decision, no defect |
| test/e2e/case_m1.go | 135 | ydotool "Escape" name trap in the shell-surface fallback (deferred-items #3, unchanged; line re-pinned at HEAD) | ⚠️ Warning | Unexercised fallback path only; one-word fix queued for next touch |
| desktop environment | — | persistent AT-SPI shell-bridge wedge (deferred-items #1, unchanged) | ℹ️ Info | Environmental; remedy documented in the runner book |
| .planning/.../02-VALIDATION.md | — | unfilled draft template (unchanged) | ℹ️ Info | Process artifact; not gated |

Judgment: no blocker. Both warnings carried forward unchanged from the prior
pass with fresh evidence at HEAD; nothing new was introduced by phase 4 on
phase-2 contract surface.

### Advisory (New Scope, Unevidenced)

None — every phase-4 change intersecting phase-2 covered files was verified
additive (diff-read), gate-proven (build/vet/lint/race/tidy) and, where it
touches live behavior, live-proven by the D-48 double-run at
code-identical-to-HEAD.

### Human Verification Required

None — both prior owner items were resolved 2026-09-15 (ADR-003 wording
deviation accepted; visual no-jump confirmed with live daemon log evidence),
recorded in 02-UAT.md (2/2 pass, owner-accepted). Final; not re-opened.
(Phase 4's own pending UAT items in 04-UAT.md are phase-4 scope.)

### Gaps Summary

**No gaps.** All 23 consolidated truths verified at HEAD 42d10a1: phase 4's
changes to covered files are additive (D-37 version identity, e2e v3/gedit/
x11/perf/D-48 infrastructure) and leave the phase-2 correction core
byte-identical (whole of internal/correct, engine, hotkey, case_word.go,
matrix-v1.yaml untouched since the prior fingerprint tree); all headless
gates green at HEAD (build, vet, 14-package -race suite, lint 0 issues,
tidy clean); live evidence upgraded to the strongest recency yet — D-48
double-run 31/31 PASS twice at 9f2fd75, whose production code is identical
to HEAD (only docs + one test-only commit between). The two carried
warnings (Detect orphaning, case_m1.go:135 Escape trap) and the
environmental AT-SPI note are unchanged, each with fresh HEAD evidence.
Deferred item #2 stays closed (WR-05, pinned tests re-run PASS). UAT 2/2
final. Fingerprint recomputed with the canonical verb.

---

_Verified: 2026-09-16T19:19:44Z_
_Verifier: Claude (gsd-verifier) — fingerprint-refresh re-verification post-Phase-4_
