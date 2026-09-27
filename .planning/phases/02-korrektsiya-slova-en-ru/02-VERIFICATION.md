---
phase: 02-korrektsiya-slova-en-ru
verified: 2026-09-20T14:30:42Z
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
covered_digest: "v1:sha256:f89ccbb16ec76d3aa9ccff87916670f2eb7647c006e595ba675adbe7b5882cf1"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 23/23
  previous_verified: 2026-09-16T19:19:44Z
  reason: "fingerprint refresh after Phase 4 gap closure (commits 42d10a1..04d00ad, plans 04-08/04-09): focus-first witness traversal + 10s quiesce budget in the shared e2e harness, D-48 v2 nightly schedule + idle-desktop preflight in the workflow, ci-runner docs — phase-2 correction core byte-identical throughout"
  gaps_closed: []
  gaps_remaining: []
  regressions: []
  closed_deferred_items:
    - "deferred-items #2 (verify-after settles on stale first push → false mismatch INFO) — CLOSED by WR-05 (d94a3ce, phase 3): first stale answer re-requires, only the second concludes; pinned by TestActor_VerifyAfterLevel1 subtests, re-run PASS at HEAD 04d00ad"
  still_open_deferred_items:
    - "deferred-items #1 (persistent AT-SPI shell-bridge wedge) — environmental, unchanged; in fact caused the three 9ff56cb D-48 fresh-session failures (04-UAT attempts 1–3), remedy documented in ci-runner.md"
    - "deferred-items #3 (case_m1.go:135 ydotool \"Escape\" name trap in the unexercised shell-surface fallback) — unchanged, line re-pinned at HEAD"
human_verification: [] # both prior owner items resolved 2026-09-15 in 02-UAT.md (2/2 pass, owner-accepted) — final, not re-opened; the pending post-fix D-48 fresh-session green run is phase-4's own tracked human item (04-VERIFICATION.md), not phase-2 scope
---

# Phase 2: Коррекция слова EN↔RU — Verification Report

**Phase Goal (ROADMAP.md):** «Пользователь двойным нажатием Right Shift исправляет последнее слово, набранное не в той раскладке, в любом поле ввода GNOME — направление определяется автоматически, регистр сохраняется, замена происходит ровно по диапазону неверного текста без визуального прыжка; это подтверждает зелёная e2e-матрица v1.»
**Verified:** 2026-09-20T14:30:42Z
**Status:** passed
**Re-verification:** Yes — fingerprint-refresh after Phase 4 gap closure (prior pass: 2026-09-16T19:19:44Z post-Phase-4-mainline, 23/23; original pass 2026-09-15T05:40:42Z, UAT 2/2 owner-accepted; no prior gaps)

## Re-verification Approach

Since the prior fingerprint tree 42d10a1, 19 commits landed (phase-4 gap
closure: plan 04-08 — focus-first witness traversal + 10s quiesce budget
resolving G-4-1; plan 04-09 — D-48 v2 nightly schedule + idle-desktop
preflight resolving G-4-2). Every covered file was classified with
`git diff 42d10a1..HEAD`:

- **Touched → deep-checked (all phase-4 harness/ops work, additive to
  phase-2 semantics):**
  `test/e2e/focus_helper.py` — focus-first witness traversal (FOCUSED
  checked on frames before descent, only focused subtrees walked) + new
  `focused-app` mode; the pre-04-08 `cmd_witness` body is kept **verbatim**
  as the full-walk fallback and the witness output contract
  (`app:role:chars=N` / `(none)`) is unchanged, so the matrix readback
  oracle phase-2 cases rely on keeps its semantics by construction;
  `test/e2e/matrix.go` — quiesce constants only (`witnessProbeTimeout`
  4s→10s, `witnessQuiesceWindow` named constant); the strict
  `KnownFields(true)` loader, per-case isolation and exitCode are
  untouched; `.github/workflows/e2e-matrix.yml` — nightly `schedule`
  trigger + idle-desktop preflight + soft mode (`D48_SKIP`); dispatch-only
  discipline kept (still zero push/PR triggers, T-02-07-01), both matrix
  run steps and the v1/v2 dispatch regression arms preserved (re-read at
  HEAD), WR-04 env-input discipline kept (`GITHUB_EVENT_NAME` via env);
  `docs/ci-runner.md` — D-48 v2 procedure.
- **Untouched → regression check (`git diff 42d10a1..HEAD` empty):** all
  14 PLAN/SUMMARY files, the whole correction core
  `internal/correct/{buffer,convert,direction,plan,runs,verify}.go`
  (+tests), `engine/*` (+tests), `internal/hotkey/fsm.go`,
  `internal/session/actor.go` (+tests), `cmd/goswitchd/main.go`,
  `layouts/`, `test/e2e/{case_word,case_ladder,case_d01,case_m1,case_resilience,case_surface}.go`,
  `test/e2e/{main,matrix_test,preflight,surface}.go`,
  `cases/matrix-v1.yaml` (16 cases, D-13 pin-geometry header intact),
  `cases/matrix-v2.yaml`, `mise.toml`, `README` fixtures, `.gitignore`,
  `SECURITY.md`, `go.mod`/`go.sum`. The entire phase-2 production surface
  is **byte-identical** to the tree of the green D-48 double-run
  (35108412175 @ 9f2fd75).
- **New files, deliberately NOT added to covered_files:**
  `test/e2e/quiesce_test.go` and `scripts/d48-nightly-dispatch.sh` are
  phase-4 artifacts (quiesce-window invariant pins, nightly dispatcher) —
  they carry phase-4 behavior, tracked by phase-4's own verification
  fingerprint; no phase-2 behavior lives in them.

The D-23 mixed-word succession, direction-anchor mechanism, WR-05 debounce,
CR-01/CR-02 config plumbing and WR-01 rung semantics documented by the prior
passes are unchanged — their pinned tests re-ran green at HEAD 04d00ad
(below).

## Live-gate evidence policy (unchanged)

Live e2e (`mise run e2e-matrix*`) drives the owner's desktop — NOT re-run by
the verifier (Phase 1/2/3/4 precedent). GitHub-side evidence re-fetched
2026-09-20 via `gh run list` / `gh api`:

| Run | Workflow | headSha | Result | Relevance |
|-----|----------|---------|--------|-----------|
| **35108412175** | e2e-matrix (dispatch) | **9f2fd75** (2026-09-16) | success, **D-48 double-run: 31/31 PASS TWICE** | Still the strongest live proof of phase-2 semantics: every word/phrase/register/exact-range/surface invariant green twice on one session. Production code at 9f2fd75 is **byte-identical to HEAD 04d00ad** (`git diff` on internal/, engine/, cmd/, layouts/ is empty) — only the e2e harness changed since. |
| 35271719601 | e2e-matrix (dispatch) | 9ff56cb (2026-09-17 20:35Z) | failure | D-48 fresh-session attempt 3 — focus-stealing finding (`witness gnome-shell:WINDOW:chars=-1` while zenity waits). Environmental/desktop-busy, diagnosed in 04-UAT; addressed by 04-09's idle-desktop preflight. **Not a phase-2 regression** — phase-2 correction semantics never implicated. |
| 35188570260 | e2e-matrix (dispatch) | 9ff56cb (2026-09-17 06:08Z) | failure (3/31) | D-48 attempt 2 — degraded a11y bridge (deferred-items #1). Same class. |
| 35142914380 | e2e-matrix (dispatch) | 9ff56cb (2026-09-16 19:49Z) | failure (8/31) | D-48 attempt 1 — witness full-tree walk crossed the 4s probe budget on a fresh-session tree (the G-4-1 scalability defect, fixed by 04-08: focus-first + 10s). Harness defect, never a product one. |
| 34913812782 | e2e-matrix (dispatch) | 419799d (2026-09-15, phase 2 code state) | success, 16/16 PASS | Still the ONLY live run of the literal v1 file's reset rows — matrix-v1.yaml is byte-identical since, so this remains the reset-row live witness. |
| pr-sanity | — | latest 35130001841 @ 80d8534 (success) | success | pr-sanity circuit alive and green (build + vet + lint + test -race + tidy + govulncheck). |
| runner | — | — | **green106 online** (re-checked via API 2026-09-20, labels self-hosted,Linux,X64,gnome) | CI circuit alive. |

**Honest recency note:** the last GREEN full-matrix run (35108412175)
predates the 04-08/04-09 harness fixes; the three later runs failed on
environmental/harness-scale causes (bridge wedge, focus-stealing, walk
budget) — all three diagnosed and structurally addressed by 04-08/04-09
with local live evidence recorded in 04-08-SUMMARY (timed witness runs
against live zenity/GTE, 2976–3165 ms, exit 0). The formal post-fix green
D-48 fresh-session run is phase-4's own tracked human item
(04-VERIFICATION.md truth 7 / human item 2: "the double-run MECHANICS are
already proven live"). It is NOT a phase-2 gap: the phase-2 production
surface is byte-identical to the green-run tree, and no failure was ever
attributed to correction semantics.

Headless gates re-run by the verifier at HEAD 04d00ad — all green:
`go build ./...` OK; `go vet ./...` OK; `go test -race -count=1 ./...` all
**14** packages ok; `mise run lint` 0 issues; `go mod tidy` + `git diff` —
no drift.

## User Flow Coverage

User capability (goal, non-story form — MVP user-flow framing discrepancy
recorded in the initial pass, unchanged).

| Step | Expected | Evidence | Status |
|------|----------|----------|--------|
| Type a word in the wrong layout | `ghbdtn` (registers/mixed/digits variants) lands in the field | matrix-v1.yaml 16 cases intact (byte-identical, re-read at HEAD); case_word.go registry wired | ✓ |
| Double-tap Right Shift | FSM Double → correction pipeline | TestActor_DoubleTapCorrects **PASS in the -race run at HEAD 04d00ad**; NewActor default KeyvalShiftR; live in D-48 double-run @9f2fd75 (production-identical code) | ✓ |
| Word corrected, direction automatic | EN→RU and RU→EN; mixed → foreign run (D-23) | TestActor_MixedWordConvertsForeignRuns PASS; live word-mixed + phrase-en-ru/ru-en + gedit/x11 word rows (both directions) PASS @9f2fd75; ConvertRuns anchor (runs.go, R) | ✓ |
| Register preserved | three registers | TestConvert_CasePreserved green (correct corpus under -race); live phrase-registers + gedit-phrase-registers PASS @9f2fd75; v1 GTE registers live @419799d | ✓ |
| Exact-range replacement, no jump | delete+commit over the range, one transaction | BuildPlan rune arithmetic + TestActor_AfterSpaceCorrects PASS; live phrase-after-space PASS @9f2fd75; perceptual check owner-accepted (UAT 2) | ✓ |
| Works in any GNOME input field | GTK/Chromium/GtkSourceView + gedit + chromium-X11 drivers | live v3 31/31 twice @9f2fd75: zenity/GTE/chromium/gedit/x11 rows all PASS | ✓ |
| Green matrix proves it | 16/16 v1; exit≠0 | infra green at HEAD (matrix_test.go + quiesce_test.go under -race, strict loader intact); live artifacts cited above | ✓ |

## Goal Achievement

### Observable Truths

Same 23 consolidated truths as the accepted prior passes; evidence refreshed
to the HEAD tree 04d00ad (R = regression-checked, byte-identical since
42d10a1; D = deep-checked, touched by phase-4 gap closure; succession
wording where a later phase superseded a mechanism).

| # | Truth | Status | Evidence at HEAD 04d00ad |
|---|-------|--------|--------------------------|
| 1 | SC-1: gnome-text-editor double Right Shift corrects `ghbdtn`→`привет`, `GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет` | ✓ VERIFIED | v1 rows intact (matrix-v1.yaml, R); live 16/16 @419799d; register semantics re-proven live at phrase width (phrase-registers, gedit-phrase-registers @9f2fd75) + unit TestActor_DoubleTapCorrects/TestConvert_CasePreserved at HEAD |
| 2 | SC-2: direction auto-detected from buffer composition both ways + unit corpus | ✓ VERIFIED (mechanism succession) | direction via ConvertRuns last-letter anchor (runs.go, R — untouched); single-script → wholesale Convert — TestActor_DoubleTapCorrects/word cases PASS at HEAD; live both directions @9f2fd75; TestDetect_* corpus green (Detect remains corpus-reference, see Anti-Patterns) |
| 3 | SC-3: replacement exactly over the wrong range, both ladder levels, abort-not-litter | ✓ VERIFIED | BuildPlan rune arithmetic intact (plan.go, R); D-27 level-2-only cap pinned; level-1 live on 5 surfaces (v3 @9f2fd75); level 2 unit-pinned PASS; abort discipline TestActor_VerifyPaths green; Chromium actual level owner-accepted (UAT 1) |
| 4 | SC-4: buffer cleared on Enter, Tab, Escape, focus change | ✓ VERIFIED | isResetKeyval table + HardReset unchanged (R); TestActor_HardResetKeyvals / TestBuffer_ResetByFocusOut **PASS at HEAD**; live reset-* @419799d + live FocusOut reset (super-space-alive) @9f2fd75 |
| 5 | SC-5: e2e-matrix v1 green — YAML cases, PASS/FAIL report, exit≠0 | ✓ VERIFIED | matrix.go strict loader (`KnownFields(true)`, D — quiesce constants only changed) + per-case isolation + exitCode green (matrix_test.go + quiesce_test.go under -race at HEAD); v1 file 16 cases intact; live v1 16/16 @419799d; v1 semantics live-green in v3 31/31 twice @9f2fd75 |
| 6 | Token semantics D-13/D-14/D-15 + toggle invariant | ✓ VERIFIED | buffer.go Push/Token/Tail/ReplaceToken untouched (R); buffer corpus green under -race |
| 7 | Mixed word (CORR-04) — D-23 succession (was D-16 refusal) | ✓ VERIFIED | mixed word converts foreign runs (паипривет); oracle intact in case_word.go (R) + matrix v1 word-mixed row; live v3 word-mixed PASS @9f2fd75; digits/punctuation neutral (D-15 continuation) |
| 8 | Conversion preserves register per rune; non-letters identical (CORR-05) | ✓ VERIFIED | convert.go untouched (R); ConvertRuns reuses Convert as-is (never forked); TestConvert_CasePreserved green; live registers at phrase width + gedit @9f2fd75 |
| 9 | Suffix verification ADR-004 (token+tail range) | ✓ VERIFIED | MatchesSuffix intact (verify.go, R); VerifyRangeAt for selection (additive); wired at pendingVerdict/afterVerdict; corpus green |
| 10 | BuildPlan both levels, all counts in runes (CORR-07) | ✓ VERIFIED | len([]rune) arithmetic intact (R); +backspaceCap param (D-27); TestBuildPlan corpus green |
| 11 | internal/correct purity: no godbus/time/goroutines | ✓ VERIFIED | purity grep re-run at HEAD: **0 hits** in production files |
| 12 | Surface drivers: fresh chromium + standalone GTE, witness-gated | ✓ VERIFIED | surface.go untouched (R); witness readback contract preserved through the 04-08 focus-first rework (verbatim full-walk fallback, output format unchanged; local timed witness runs 2976–3165 ms, exit 0 — 04-08-SUMMARY); live gedit/x11/smoke rows green @9f2fd75 |
| 13 | Preflight rejects missing binaries | ✓ VERIFIED | preflight.go untouched since 42d10a1 (R); package green under -race; live preflight ok in D-48 log |
| 14 | Tracer: double tap runs buffer→direction→convert→verify→delete+commit (CORR-01) | ✓ VERIFIED | TestActor_DoubleTapCorrects **PASS at HEAD**; live word/phrase cases in D-48 double-run |
| 15 | Word after space corrected with separator (D-13) | ✓ VERIFIED | TestActor_AfterSpaceCorrects **PASS at HEAD**; live phrase-after-space PASS @9f2fd75 |
| 16 | Log discipline D-20/D-21 | ✓ VERIFIED | logCorrectionDone; TestActor_DebugCorrectionRecord green; D-20 also pinned by ClipboardRungAfterMismatch subtest («no content in INFO») |
| 17 | Flip on Single: mode EN↔RU, INFO mode record, XKB untouched | ✓ VERIFIED | flipScript; TestActor_FlipOnSingle **PASS at HEAD**; live layout-single + gedit-layout-single @9f2fd75 |
| 18 | RU consumption commits Cyrillic; script-true buffer in ALL printing branches incl. digits | ✓ VERIFIED | feedKey unchanged invariants; TestActor_RUDigitsFullPipeline **PASS at HEAD**; live phrase-ru-en @9f2fd75 |
| 19 | Level 2: ForwardKeyEvent(BackSpace)×(token+tail runes) burst → one commit, in order (CORR-07) | ✓ VERIFIED | executeLevel2 + emitter wire pin; TestActor_Level2NoCaps/WithTailAndRunes **PASS at HEAD**; over-cap range refuses (D-27, LevelNone) |
| 20 | Verify-after with epoch: mismatch → INFO counter, never auto-repair; stale timers no-op | ✓ VERIFIED (WR-05 semantics) | settleAfter unchanged; first stale → `stale-retry` + re-Require, second stale → mismatch once; epoch guard retained; TestActor_VerifyAfterLevel1 3 subtests **PASS at HEAD 04d00ad**; deferred item #2 stays CLOSED |
| 21 | Reset triggers wired live (CORR-09) | ✓ VERIFIED | reset-* live @419799d; FocusOut reset live @9f2fd75; unit re-run PASS |
| 22 | Matrix infrastructure: strict decoder, per-case isolation, closed vocabularies | ✓ VERIFIED | matrix.go KnownFields intact (D — quiesce window retune additive, invariant now pinned by quiesce_test.go, green under -race); matrix_test.go green at HEAD |
| 23 | CI circuit D-19: dispatch-only workflow on self-hosted gnome runner | ✓ VERIFIED | e2e-matrix.yml re-checked (D — nightly schedule + soft-mode preflight added; still zero push/PR triggers, permissions contents:read, WR-04 env-input kept, v1/v2 dispatch arms preserved); runner green106 online re-checked via API 2026-09-20; pr-sanity green; dispatch circuit's live green remains 35108412175 at production-identical code |

**Score:** 23/23 truths verified (0 present-but-behavior-unverified)

Behavior-dependent truths note: every state-transition/invariant truth is
covered by a named behavioral test that ran PASS in the verifier's single
`go test -race -count=1 ./...` run at HEAD 04d00ad (all 40 TestActor_* and
24 correct-package tests enumerated by `-list` first); the live-desktop
dimension rests on the D-48 double-run at production-code-identical-to-HEAD
(9f2fd75).

### Required Artifacts

`verify.artifacts` re-run per plan in both prior passes: **27/28** — the
single "miss" is the same verifier-tool `~` false negative
(`~/.config/systemd/user/goswitch-ci-runner.service` exists, 750 bytes).
No plan frontmatter changed since (all 14 PLAN/SUMMARY files byte-identical
— `git diff 42d10a1..HEAD` empty). All artifacts re-confirmed to exist at
HEAD; all substantive.

### Key Link Verification

15/16 per the prior passes; the one unverified link is the same plan-spec
false negative (02-01 buffer.go→direction.go, pattern `\.Token\(\)`); the
actual consumer wiring is `internal/session/actor.go` (`a.buf.Token()` →
`startRangeCorrection` → ConvertRuns) — re-read at HEAD, present, wired,
behavior-tested (actor.go untouched since the green runs).

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| internal/session/actor.go | token/tail | buf.Token()/Tail() fed by HandleKey from live ProcessKeyEvent | ✓ live corrections in the D-48 double-run @9f2fd75 | ✓ FLOWING |
| internal/session/actor.go | surr/sel | HandleSurroundingText live pushes + Require rounds | ✓ live word/phrase/selection corrections @9f2fd75 | ✓ FLOWING |
| internal/correct/runs.go | out/changed | ConvertRuns anchor scan over the live range | ✓ word-mixed live PASS @9f2fd75 | ✓ FLOWING |
| internal/correct/plan.go | Plan | BuildPlan from token/tail/converted + caps bit | ✓ real caps 0x29 live; level dispatch exercised | ✓ FLOWING |
| test/e2e/matrix.go | report | per-case runMatrixCaseIsolated results | ✓ 31/31 PASS twice in re-fetched D-48 log | ✓ FLOWING |
| layouts/tables.go | ENToRU/RUToEN | xkb-generated (Phase 1) | ✓ RU-commit path live (phrase-ru-en, gedit/x11 rows) | ✓ FLOWING |

### Behavioral Spot-Checks (re-run at HEAD 04d00ad)

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full corpus, all packages (includes every named test below) | `go test -race -count=1 ./...` | 14 pkgs ok (exit 0) | ✓ PASS |
| Named-test existence proof | `go test ./internal/session -list TestActor` / `./internal/correct -list Test` | 40 + 24 names enumerated, incl. all previously pinned tests | ✓ PASS |
| Gates | `go build ./...` / `go vet ./...` / `mise run lint` / `go mod tidy`+diff | OK / OK / 0 issues / no drift | ✓ PASS |
| Purity invariant (truth 11) | grep godbus/time./go func in internal/correct non-test | 0 hits | ✓ PASS |
| v1 matrix integrity | matrix-v1.yaml case count + strict-schema header | 16 cases, D-13 header intact, byte-identical | ✓ PASS |
| Correction-core immutability | `git diff 42d10a1..HEAD -- internal/ engine/ cmd/ layouts/` | empty | ✓ PASS |
| Debt markers on gap-closure-touched files | grep TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER on 6 changed files | 0 hits | ✓ PASS |
| Live e2e re-run | not re-run (drives owner's desktop) | D-48 double-run 31/31×2 @9f2fd75 re-fetched 2026-09-20; runner online; pr-sanity green | ? SKIP (by policy) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` convention; the phase's runnable probes are
the mise e2e tasks (live-session; covered by the CI-artifact policy) and the
CI gate (re-run, PASS). Not applicable beyond that.

### Test Quality Audit

Disabled tests on Phase 2 requirement-linked files: **0** (the only
`t.Skipf` in the tree remains the ctlsvc environmental dbus-daemon guard —
not requirement-linked here; phase-4 gap closure added zero skips on
phase-2 files; quiesce_test.go adds positive pins only). Circular
patterns: **0** (hand-authored oracles; layout tables regenerate from
system xkb). Assertion strength: rune-exact equality / op-log order /
log-shape scans — no existence-only assertions on requirement-linked
tests.

### Requirements Coverage

| Requirement | Source Plan | Status | Evidence |
|-------------|------------|--------|----------|
| CORR-01 | 02-03, 02-04 | ✓ SATISFIED | truths 14, 17; live word/phrase both directions incl. gedit/x11 @9f2fd75 |
| CORR-04 | 02-01, 02-04 | ✓ SATISFIED | truths 2, 7, 18; direction by composition via ConvertRuns anchor |
| CORR-05 | 02-01 | ✓ SATISFIED | truth 8; registers live at phrase width + gedit @9f2fd75 + v1 GTE registers @419799d |
| CORR-07 | 02-01, 02-03, 02-05 | ✓ SATISFIED | truths 3, 9, 10, 15, 19, 20 |
| CORR-09 | 02-05 | ✓ SATISFIED | truths 4, 21 |
| TEST-04 | 02-02, 02-06, 02-07 | ✓ SATISFIED | truths 5, 12, 22, 23; REQUIREMENTS.md re-read at HEAD: all 6 IDs mapped Phase 2, Complete, "Phase 2: 6" — unchanged |

Orphaned requirements: **none** (REQUIREMENTS.md maps exactly these 6 IDs
to Phase 2, all Complete).

### Anti-Patterns Found

Debt-marker gate re-run at HEAD 04d00ad across all phase-covered
Go/Python/YAML/TOML/workflow files **and** the six files touched by phase-4
gap closure (focus_helper.py, matrix.go, quiesce_test.go, e2e-matrix.yml,
d48-nightly-dispatch.sh, ci-runner.md): **ZERO** TBD/FIXME/XXX and zero
TODO/HACK/PLACEHOLDER.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| internal/correct/direction.go | — | `Detect` production-orphaned by the D-23 succession (re-confirmed at HEAD: zero production callers; runs.go cites it as the classifyRune reference; corpus stays green) | ⚠️ Warning | Dead code carried by its own tests; SC-2 contract unaffected (anchor scan proves it live). Candidate: annotate as corpus-reference or fold into the run scanner — owner cleanup decision, no defect. File untouched since prior pass |
| test/e2e/case_m1.go | 135 | ydotool "Escape" name trap in the shell-surface fallback (deferred-items #3, unchanged; line re-pinned at HEAD) | ⚠️ Warning | Unexercised fallback path only; one-word fix queued for next touch. File untouched since prior pass |
| desktop environment | — | persistent AT-SPI shell-bridge wedge (deferred-items #1, unchanged) | ℹ️ Info | Environmental; materially caused the three 9ff56cb D-48 failures; remedy documented in ci-runner.md and now also guarded by the 04-09 idle preflight |
| .planning/.../02-VALIDATION.md | — | unfilled draft template (unchanged) | ℹ️ Info | Process artifact; not gated |

Judgment: no blocker. Both warnings carried forward unchanged — their files
are byte-identical to the prior pass's tree, so no regression is even
possible; nothing new was introduced on phase-2 contract surface.

### Advisory (New Scope, Unevidenced)

None. Every phase-4 gap-closure change intersecting phase-2 covered files
was diff-read (focus-first witness keeps the readback contract by
construction with a verbatim fallback; matrix.go quiesce retune is
constant-only; workflow additions preserve dispatch-only and WR-04),
gate-proven (build/vet/lint/race/tidy at HEAD) and, where it touches live
behavior, locally live-proven per 04-08-SUMMARY with the formal post-fix
green run correctly owned by phase-4's verification.

### Human Verification Required

None — both prior owner items were resolved 2026-09-15 (ADR-003 wording
deviation accepted; visual no-jump confirmed with live daemon log evidence),
recorded in 02-UAT.md (2/2 pass, owner-accepted; re-read at HEAD: status
complete, passed: 2). Final; not re-opened. (The pending D-48
fresh-session green run after the 04-08/04-09 harness fixes is phase-4's
own tracked human item in 04-VERIFICATION.md — phase-4 scope.)

### Gaps Summary

**No gaps.** All 23 consolidated truths verified at HEAD 04d00ad. Phase-4
gap closure (04-08/04-09) changed only the shared e2e harness (focus-first
witness traversal with verbatim full-walk fallback, quiesce budget 10s now
pinned by quiesce_test.go) and ops infrastructure (nightly schedule +
idle-desktop preflight, ci-runner docs); the entire phase-2 production
surface — internal/correct, engine, hotkey, session, cmd, layouts,
case_word.go, matrix-v1.yaml — is byte-identical to the tree of the green
D-48 double-run 31/31×2 (35108412175 @ 9f2fd75). All headless gates green
at HEAD (build, vet, 14-package -race suite, lint 0 issues, tidy clean).
The three failed runs at 9ff56cb were environmental/harness-scale (bridge
wedge, focus-stealing, walk budget), diagnosed and structurally fixed by
the very plans that produced this refresh; none implicated phase-2
correction semantics. The two carried warnings (Detect orphaning,
case_m1.go:135 Escape trap) and the environmental AT-SPI note are
unchanged on untouched files. Deferred item #2 stays closed (WR-05, pinned
tests PASS at HEAD). UAT 2/2 final. Fingerprint recomputed with the
canonical verb.

---

_Verified: 2026-09-20T14:30:42Z_
_Verifier: Claude (gsd-verifier) — fingerprint-refresh re-verification post-Phase-4-gap-closure_
