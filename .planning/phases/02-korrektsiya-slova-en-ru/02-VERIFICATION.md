---
phase: 02-korrektsiya-slova-en-ru
verified: 2026-09-15T17:59:16Z
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
covered_digest: "v1:sha256:4a9afc75dccc0daaa1a646afa4cc72c30f3f02ba6338d07363af81f51d54d605"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 23/23
  previous_verified: 2026-09-15T05:40:42Z
  reason: "fingerprint refresh after Phase 3 execution (commits 6fafa98..c23e98c) reworked shared covered files"
  gaps_closed: []
  gaps_remaining: []
  regressions: []
  closed_deferred_items:
    - "deferred-items #2 (verify-after settles on stale first push → false mismatch INFO) — CLOSED by WR-05 (d94a3ce): first stale answer re-requires, only the second concludes; pinned by TestActor_VerifyAfterLevel1 subtests, re-run PASS"
  still_open_deferred_items:
    - "deferred-items #1 (persistent AT-SPI shell-bridge wedge) — environmental, unchanged"
    - "deferred-items #3 (case_m1.go:135 ydotool \"Escape\" name trap in the unexercised shell-surface fallback) — unchanged, warning stands"
human_verification: [] # both prior owner items resolved 2026-09-15 in 02-UAT.md (2/2 pass, owner-accepted) — final, not re-opened
---

# Phase 2: Коррекция слова EN↔RU — Verification Report

**Phase Goal (ROADMAP.md):** «Пользователь двойным нажатием Right Shift исправляет последнее слово, набранное не в той раскладке, в любом поле ввода GNOME — направление определяется автоматически, регистр сохраняется, замена происходит ровно по диапазону неверного текста без визуального прыжка; это подтверждает зелёная e2e-матрица v1.»
**Verified:** 2026-09-15T17:59:16Z
**Status:** passed
**Re-verification:** Yes — fingerprint-refresh after Phase 3 (prior pass: 2026-09-15T05:40:42Z, UAT 2/2 owner-accepted; no prior gaps)

## Re-verification Approach

Phase 3 (7 plans + 8 code-review fix commits CR-01..03 / WR-01..05,
d715937..c23e98c) reworked files this phase covers. Every covered file was
classified with `git diff b52e85c..HEAD` (b52e85c = the commit that recorded
the prior `passed` verdict):

- **Reworked → deep re-check** (SC-level contracts re-proven at HEAD):
  `internal/session/actor.go` (+1295/-rework: tap_key wiring CR-01,
  verify_wait_ms CR-02, clipboard rung off-mutex WR-01, verify-after
  debounce WR-05, D-23 run pipeline, selection/combo/MACR branches),
  `internal/correct/{plan,verify,buffer}.go` (D-27 cap, VerifyRangeAt,
  SelectionRange), `test/e2e/{matrix,main}.go`,
  `test/e2e/cases/matrix-v1.yaml` (one row successed, see below),
  `case_word.go`, `engine/{engine,keys}.go`, `cmd/goswitchd/main.go`,
  `mise.toml`, `.github/workflows/e2e-matrix.yml`, `go.mod/go.sum`.
- **Untouched → regression check** (existence + sanity): all 14 PLAN/SUMMARY
  files, `internal/correct/{convert,direction}.go`, `engine/{factory,types}.go`,
  `test/e2e/{case_d01,case_ladder,case_surface}.go`, `surface.go`,
  `focus_helper.py`, fixtures, docs — all present and unchanged.

## Designed Successions Touching Phase 2 Truths (roadmap-sanctioned, not deviations)

Phase 3's ROADMAP SC-2/SC-3 (mixed text, D-27) and STATE.md decisions
D-16→D-23 sanctioned four changes on Phase 2 contract surface:

1. **D-16 → D-23 (mixed word):** the mixed word now CONVERTS its foreign
   run (`gfbпривет`→`паипривет`) instead of refusal. Matrix v1 row renamed
   `word-mixed-untouched`→`word-mixed`, oracle updated; pure single-script
   words still convert WHOLESALE through the same `Convert` (runs.go:84 —
   «уточнение D-22 … keeps matrix v1 green»), so register/direction/exact-range
   invariants of every other v1 row are untouched. Live-green in v2 run
   34984814995 (`word-mixed` PASS).
2. **Direction mechanism: `correct.Detect` → `ConvertRuns` anchor scan.**
   SC-2's contract (direction by buffer composition) now holds via the
   last-letter anchor in `ConvertRuns` (runs.go:63); `Detect` remains as the
   corpus-pinned reference (its unit corpus still green) but is
   production-orphaned — see Anti-Patterns.
3. **WR-05 verify-after debounce:** the first mismatching post-correction
   push now re-requires (`stale-retry`); only the SECOND stale answer
   concludes the INFO mismatch. The SC-level invariant (mismatch counts
   exactly once, NEVER auto-repairs) is preserved and re-pinned; a GENUINE
   mismatch (client ignored the deletion) still concludes — delayed by one
   re-require round, never masked. Deferred item #2 is thereby CLOSED.
4. **Config-driven tap key / verify budget (CR-01/CR-02):** defaults are
   pinned in `NewActor` (`hotkey.KeyvalShiftR` into BOTH the FSM and the
   actor's `tapKeyval`; `defaultVerifyWait`); `applySnapshot` re-resolves
   only a CHANGED, parseable name (last-good discipline), and config
   validation rejects modifier-bearing tap keys (WR-02, `errTapKeyBare`).
   The default double-Right-Shift FSM semantics are unchanged.

## Live-gate evidence policy (unchanged)

Live e2e (`mise run e2e-matrix` / `e2e-matrix-v2`) drives the owner's
desktop — NOT re-run by the verifier (Phase 1/2 precedent). Machine-generated
GitHub-side evidence, independently re-fetched:

| Run | Workflow | headSha | Result | Relevance |
|-----|----------|---------|--------|-----------|
| 34984814995 | e2e-matrix (dispatch) | **21df37c** (12 commits before HEAD; the 8 code-fix commits CR-01..03/WR-01..05 intervene) | success, **21/21 PASS** (log re-fetched: phrase-en-ru/ru-en/registers/after-space, long-phrase, phrase-gte/phrase-chromium, word-mixed, 5 select rows, layout-single, combo-word-layout, super-space-alive, 3 reload rows, macr) | Live proof of the successed `word-mixed` row + phrase-width analogs of every v1 word semantic (registers, both directions, tail separator, 3 surfaces) + live FocusOut buffer reset (super-space-alive) + single-tap flip (layout-single). Recency caveat: the 8 fix commits after 21df37c are default-path-neutral (defaults pinned in NewActor; rung opt-in OFF by default; debounce touches only the verdict observability) and are unit-pinned under -race at HEAD — but no live-desktop matrix run exists after them. |
| 34913812782 | e2e-matrix (dispatch) | 419799d (Phase 2 code state) | success, 16/16 PASS incl. three GTE registers, reset-* rows, fallback-chromium | The literal v1 row set live-green at the Phase 2 state. Since then exactly ONE v1 row changed (the D-23 succession); no post-succession v1 CI dispatch exists — noted honestly. |
| pr-sanity | — | latest 34914315850 @ 6fad577 (Phase 2 era) | — | **No pr-sanity run exists for the phase-3 branch or HEAD** (workflow triggers: push to main / PR / dispatch; the branch was pushed without a PR). Compensated by the verifier re-running the identical gate set locally (below). |
| runner | — | — | green106 online, labels self-hosted,Linux,X64,gnome | CI circuit alive. |

Headless gates re-run by the verifier at HEAD c23e98c — all green:
`go build ./...` OK; `go vet ./...` OK; `go test -race ./...` all 11 packages
ok (engine, appid, clipboard, config, correct, ctlsvc, hotkey, logging,
session, layouts, test/e2e); `mise run lint` 0 issues; `go mod tidy` +
`git diff` — no drift.

## User Flow Coverage

User capability (goal, non-story form — MVP user-flow framing discrepancy
recorded in the initial pass, unchanged).

| Step | Expected | Evidence | Status |
|------|----------|----------|--------|
| Type a word in the wrong layout | `ghbdtn` (registers/mixed/digits variants) lands in the field | matrix-v1.yaml cases intact; case_word.go registry wired (main.go:204+) | ✓ |
| Double-tap Right Shift | FSM Double → correction pipeline | TestActor_DoubleTapCorrects **re-run PASS at HEAD**; NewActor default KeyvalShiftR; live `action n:2` in both CI matrix runs | ✓ |
| Word corrected, direction automatic | EN→RU and RU→EN; mixed → foreign run (D-23 succession) | TestActor_MixedWordConvertsForeignRuns PASS; v2 live word-mixed + phrase-en-ru/phrase-ru-en PASS; ConvertRuns anchor | ✓ |
| Register preserved | three registers | TestConvert_CasePreserved green; v2 live phrase-registers («Привет ПРИВЕТ») PASS; v1 GTE registers live-green @419799d | ✓ |
| Exact-range replacement, no jump | delete+commit over the range, one transaction | BuildPlan rune arithmetic + TestActor_AfterSpaceCorrects re-run PASS; v2 live phrase-after-space (rune-exact trailing space) PASS; perceptual check owner-accepted (UAT 2) | ✓ |
| Works in any GNOME input field | GTK/Chromium/GtkSourceView | v1 16/16 @419799d; v2 phrase-gte/phrase-chromium/select-* live @21df37c | ✓ |
| Green matrix proves it | 16/16 v1; exit≠0 | infra green at HEAD (matrix_test.go under -race); live artifacts cited above | ✓ |

## Goal Achievement

### Observable Truths

Same 23 consolidated truths as the accepted prior pass; evidence refreshed
to the HEAD tree (R = regression-checked, untouched file; D = deep-checked,
reworked file; succession wording where Phase 3 superseded a mechanism).

| # | Truth | Status | Evidence at HEAD c23e98c |
|---|-------|--------|--------------------------|
| 1 | SC-1: gnome-text-editor double Right Shift corrects `ghbdtn`→`привет`, `GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет` | ✓ VERIFIED | v1 rows intact (matrix-v1.yaml); live 16/16 @419799d; register semantics re-proven live at phrase width (v2 phrase-registers @21df37c) + unit TestActor_DoubleTapCorrects/TestConvert_CasePreserved at HEAD |
| 2 | SC-2: direction auto-detected from buffer composition both ways + unit corpus | ✓ VERIFIED (mechanism succession) | direction now via ConvertRuns last-letter anchor (runs.go:63-91); single-script → wholesale Convert — TestActor_DoubleTapCorrects/word cases re-run PASS; TestDetect_* corpus green (Detect now corpus-reference, see Anti-Patterns) |
| 3 | SC-3: replacement exactly over the wrong range, both ladder levels, abort-not-litter | ✓ VERIFIED | BuildPlan rune arithmetic intact (plan.go); D-27 added LevelNone cap for level 2 ONLY (level 1 uncapped — comment pinned); level-1 live on 3 surfaces (both CI runs); level 2 unit-pinned re-run PASS; abort discipline TestActor_VerifyPaths green; Chromium actual level recorded (owner-accepted UAT 1) |
| 4 | SC-4: buffer cleared on Enter, Tab, Escape, focus change | ✓ VERIFIED | isResetKeyval table + HandleLifecycle HardReset unchanged semantics; TestActor_HardResetKeyvals / TestBuffer_ResetByFocusOut **re-run PASS**; live reset-* @419799d + live FocusOut reset in v2 super-space-alive @21df37c |
| 5 | SC-5: e2e-matrix v1 green — YAML cases, PASS/FAIL report, exit≠0 | ✓ VERIFIED | matrix.go strict loader + per-case isolation + exitCode green at HEAD (matrix_test.go under -race); v1 file 16 cases intact; live v1 16/16 @419799d; successed row live-green in v2 @21df37c (no post-succession v1 dispatch — recency noted) |
| 6 | Token semantics D-13/D-14/D-15 + toggle invariant | ✓ VERIFIED | buffer.go Push/Token/Tail/ReplaceToken (+ Phase 3 backspace/phrase helpers, additive); buffer corpus green under -race (D) |
| 7 | Mixed word (CORR-04) — WAS D-16 refusal | ✓ VERIFIED (D-23 succession) | mixed word now converts foreign runs (паипривет); oracle updated in case_word.go + matrix v1 row + live v2 word-mixed PASS; digits/punctuation neutral (runs.go D-15 continuation); homogeneous words keep Phase 2 wholesale conversion — matrix v1 stays green by design |
| 8 | Conversion preserves register per rune; non-letters identical (CORR-05) | ✓ VERIFIED | convert.go untouched (R); ConvertRuns reuses Convert as-is (never forked); TestConvert_CasePreserved green; live registers @21df37c |
| 9 | Suffix verification ADR-004 (token+tail range) | ✓ VERIFIED | MatchesSuffix intact; VerifyRangeAt added for selection (additive); wired at pendingVerdict/afterVerdict; corpus green |
| 10 | BuildPlan both levels, all counts in runes (CORR-07) | ✓ VERIFIED | len([]rune) arithmetic intact; +backspaceCap param (D-27); TestBuildPlan corpus (extended) green |
| 11 | internal/correct purity: no godbus/time/goroutines | ✓ VERIFIED | purity grep re-run: 0 hits in production files |
| 12 | Surface drivers: fresh chromium + standalone GTE, witness-gated | ✓ VERIFIED | surface.go untouched (R); smoke cases in registry |
| 13 | Preflight rejects missing binaries | ✓ VERIFIED | preflight.go (mark-constant refactor only); package green |
| 14 | Tracer: double tap runs buffer→direction→convert→verify→delete+commit (CORR-01) | ✓ VERIFIED | TestActor_DoubleTapCorrects **re-run PASS**; live word/phrase cases in both CI runs |
| 15 | Word after space corrected with separator (D-13) | ✓ VERIFIED | TestActor_AfterSpaceCorrects **re-run PASS**; v2 phrase-after-space live PASS (rune-exact trailing space) |
| 16 | Log discipline D-20/D-21 | ✓ VERIFIED | logCorrectionDone; TestActor_DebugCorrectionRecord green in full -race run |
| 17 | Flip on Single: mode EN↔RU, INFO mode record, XKB untouched | ✓ VERIFIED | flipScript; TestActor_FlipOnSingle **re-run PASS**; live layout-single (both directions, D-34 mode oracle) @21df37c |
| 18 | RU consumption commits Cyrillic; script-true buffer in ALL printing branches incl. digits | ✓ VERIFIED | feedKey unchanged invariants; TestActor_RUDigitsFullPipeline **re-run PASS** |
| 19 | Level 2: ForwardKeyEvent(BackSpace)×(token+tail runes) burst → one commit, in order (CORR-07) | ✓ VERIFIED | executeLevel2 + emitter wire pin; TestActor_Level2NoCaps/WithTailAndRunes **re-run PASS**; over-cap range now refuses (D-27, LevelNone) instead of an unbounded burst |
| 20 | Verify-after with epoch: mismatch → INFO counter, never auto-repair; stale timers no-op | ✓ VERIFIED (WR-05 semantics) | settleAfter: first stale → `stale-retry` + re-Require (same epoch, re-armed deadline), second stale → mismatch once; genuine mismatch NOT masked; epoch guard retained; TestActor_VerifyAfterLevel1 3 subtests **re-run PASS** (match quiet / mismatch-once-no-repair / intermediate-push race); deferred item #2 CLOSED |
| 21 | Reset triggers wired live (CORR-09) | ✓ VERIFIED | reset-* live @419799d; FocusOut reset live @21df37c; unit re-run PASS |
| 22 | Matrix infrastructure: strict decoder, per-case isolation, closed vocabularies | ✓ VERIFIED | matrix.go KnownFields + v2 step kinds (select/combo/reload) added additively; matrix_test.go green at HEAD under -race |
| 23 | CI circuit D-19: dispatch-only workflow on self-hosted gnome runner | ✓ VERIFIED | e2e-matrix.yml (WR-04: matrix input via env, never shell-interpolated); run 34984814995 success from the phase-3 branch; runner green106 online (gnome label); pr-sanity: no branch run exists — local gate re-run compensates |

**Score:** 23/23 truths verified (0 present-but-behavior-unverified)

Behavior-dependent truths note: every state-transition/invariant truth is
covered by a named behavioral test re-run by the verifier under -race at
HEAD (all PASS — spot-check table below); the live-desktop dimension rests
on the two independently re-fetched CI artifacts with the recency caveats
recorded in the evidence table.

### Decision Coverage

9/9 decisions honored at the initial pass (02-CONTEXT.md); Phase 3 did not
touch Phase 2 decisions. Non-blocking gate, unchanged.

### Required Artifacts

`verify.artifacts` re-run per plan at HEAD: **27/28** — 02-01 5/5, 02-02 4/4,
02-03 4/4, 02-04 3/3, 02-05 4/4, 02-06 4/4, 02-07 3/4. The single "miss" is
the same verifier-tool `~` false negative as the prior pass —
`~/.config/systemd/user/goswitch-ci-runner.service` exists (750 bytes,
re-checked). All artifacts substantive.

### Key Link Verification

`verify.key-links` re-run per plan at HEAD: **15/16**. The one unverified
link is the same plan-spec false negative as the prior pass (02-01
buffer.go→direction.go, pattern `\.Token\(\)`); the actual consumer wiring
is `actor.go:1282` (`a.buf.Token()` → `startRangeCorrection` → ConvertRuns)
— present, wired, behavior-tested.

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| internal/session/actor.go | token/tail | buf.Token()/Tail() fed by HandleKey from live ProcessKeyEvent | ✓ live corrections in both CI matrix runs | ✓ FLOWING |
| internal/session/actor.go | surr/sel | HandleSurroundingText live pushes + Require rounds | ✓ live word/phrase/selection corrections verified | ✓ FLOWING |
| internal/correct/runs.go | out/changed | ConvertRuns anchor scan over the live range | ✓ word-mixed live PASS (паипривет) @21df37c | ✓ FLOWING |
| internal/correct/plan.go | Plan | BuildPlan from token/tail/converted + caps bit | ✓ real caps 0x29 live; level dispatch exercised | ✓ FLOWING |
| test/e2e/matrix.go | report | per-case runMatrixCaseIsolated results | ✓ 16/16 + 21/21 PASS lines in re-fetched CI logs | ✓ FLOWING |
| layouts/tables.go | ENToRU/RUToEN | xkb-generated (Phase 1) | ✓ RU-commit path live (phrase-ru-en) | ✓ FLOWING |

### Behavioral Spot-Checks (re-run at HEAD c23e98c, all -race -count=1)

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full corpus, all packages | `go test -race ./...` | 11 pkgs ok | ✓ PASS |
| Default double-tap pipeline | `TestActor_DoubleTapCorrects` | PASS | ✓ PASS |
| WR-05 verify-after debounce (match / mismatch-once / intermediate-push race) | `TestActor_VerifyAfterLevel1` (3 subtests) | PASS | ✓ PASS |
| Exact range + tail separator | `TestActor_AfterSpaceCorrects` | PASS | ✓ PASS |
| Level-2 ladder | `TestActor_Level2NoCaps` / `Level2WithTailAndRunes` | PASS | ✓ PASS |
| Reset keyvals + FocusOut | `TestActor_HardResetKeyvals` / `TestBuffer_ResetByFocusOut` | PASS | ✓ PASS |
| Flip / RU consumption | `TestActor_FlipOnSingle` / `RUDigitsFullPipeline` | PASS | ✓ PASS |
| D-23 mixed word | `TestActor_MixedWordConvertsForeignRuns` | PASS | ✓ PASS |
| CR-01 tap key (default + hot swap) | `TestActor_HotReloadTapKey` (real-timer subtest) | PASS | ✓ PASS |
| CR-02 verify budget | `TestActor_HotReloadVerifyWait` | PASS | ✓ PASS |
| WR-01 rung off key path + after mismatch | `TestActor_ClipboardRungOffKeyPath` / `ClipboardRungAfterMismatch` | PASS | ✓ PASS |
| Gates | build / vet / lint / tidy | OK / OK / 0 issues / no drift | ✓ PASS |
| Live e2e re-run | not re-run (drives owner's desktop) | CI artifacts cited | ? SKIP (by policy) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` convention; the phase's runnable probes are
the mise e2e tasks (live-session; covered by the CI-artifact policy) and the
CI gate (re-run, PASS). Not applicable beyond that.

### Test Quality Audit

Disabled tests on Phase 2 requirement-linked files: **0** (the only
`t.Skipf` in the tree is ctlsvc_test.go:374 — a Phase 3 file, an
environmental dbus-daemon guard, not requirement-linked here). Circular
patterns: **0** (hand-authored oracles; layout tables regenerate from system
xkb). Assertion strength: rune-exact equality / op-log order / log-shape
scans — no existence-only assertions on requirement-linked tests.

### Requirements Coverage

| Requirement | Source Plan | Status | Evidence |
|-------------|------------|--------|----------|
| CORR-01 | 02-03, 02-04 | ✓ SATISFIED | truths 14, 17; live word/phrase both directions |
| CORR-04 | 02-01, 02-04 | ✓ SATISFIED | truths 2, 7, 18; direction by composition via ConvertRuns anchor |
| CORR-05 | 02-01 | ✓ SATISFIED | truth 8; registers live at phrase width + v1 GTE registers @419799d |
| CORR-07 | 02-01, 02-03, 02-05 | ✓ SATISFIED | truths 3, 9, 10, 15, 19, 20 |
| CORR-09 | 02-05 | ✓ SATISFIED | truths 4, 21 |
| TEST-04 | 02-02, 02-06, 02-07 | ✓ SATISFIED | truths 5, 12, 22, 23; REQUIREMENTS.md already records «широта наращивается в Фазе 3» |

Orphaned requirements: **none** (REQUIREMENTS.md maps exactly these 6 IDs,
all Complete).

### Anti-Patterns Found

Debt-marker gate re-run at HEAD: **ZERO** TBD/FIXME/XXX and zero
TODO/HACK/PLACEHOLDER across all phase-covered Go/Python/YAML/TOML/workflow
files.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| internal/correct/direction.go | — | `Detect` production-orphaned by the D-23 succession (no production caller; runs.go:19 cites it as the classifyRune reference; its corpus stays green) | ⚠️ Warning | Dead code carried by its own tests; SC-2 contract unaffected (anchor scan proves it live). Candidate: annotate as corpus-reference or fold into the run scanner — owner cleanup decision, no defect |
| test/e2e/case_m1.go | 135 | ydotool "Escape" name trap in the shell-surface fallback (deferred-items #3, unchanged) | ⚠️ Warning | Unexercised fallback path only; one-word fix queued for next touch |
| desktop environment | — | persistent AT-SPI shell-bridge wedge (deferred-items #1, unchanged) | ℹ️ Info | Environmental; remedy documented in the runner book |
| .planning/.../02-VALIDATION.md | — | unfilled draft template (unchanged) | ℹ️ Info | Process artifact; not gated |

Judgment: no blocker. The Detect orphaning is the only new observation this
pass; it has no failing test and no reproducible defect (evidence gate: not
a blocker), and the succession that caused it is roadmap-sanctioned.

### Advisory (New Scope, Unevidenced)

None — no finding was downgraded from blocker this pass (re-verification
gate applied; the one new observation, Detect orphaning, was classified a
warning on its own merits, not an unevidenced blocker).

### Human Verification Required

None — both prior owner items were resolved 2026-09-15 (ADR-003 wording
deviation accepted; visual no-jump confirmed with live daemon log evidence),
recorded in 02-UAT.md (2/2 pass, owner-accepted). Final; not re-opened.

### Gaps Summary

**No gaps.** All 23 consolidated truths verified at HEAD c23e98c: reworked
files deep-checked (the four code-review fixes on Phase 2 contract surface —
WR-05 debounce, CR-01 tap key, CR-02 verify budget, WR-01 rung — each
default-path-neutral and behavior-pinned by tests re-run under -race), all
untouched files regression-checked, all headless gates green, both live CI
artifacts independently re-fetched with honest recency caveats (v2 21/21 on
21df37c, 12 commits / 8 code fixes before HEAD; literal v1 rows last
live-green on 419799d with exactly one row successed since). Four
roadmap-sanctioned successions documented; deferred item #2 closed by WR-05;
deferred items #1/#3 unchanged warnings. UAT 2/2 final.

---

_Verified: 2026-09-15T17:59:16Z_
_Verifier: Claude (gsd-verifier) — fingerprint-refresh re-verification post-Phase-3_
