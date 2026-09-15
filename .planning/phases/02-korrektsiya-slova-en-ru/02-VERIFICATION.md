---
phase: 02-korrektsiya-slova-en-ru
verified: 2026-09-15T00:55:00Z
status: human_needed
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
  - internal/correct/verify.go
  - internal/correct/verify_test.go
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
  - test/e2e/fixtures/input.html
  - test/e2e/focus_helper.py
  - test/e2e/main.go
  - test/e2e/matrix.go
  - test/e2e/matrix_test.go
  - test/e2e/preflight.go
  - test/e2e/surface.go
covered_digest: "v1:sha256:a8a243cd1b6f2bc5fa0a050e776b43236a493bf674f50855a025f0d4aaf23ca0"
behavior_unverified: 0
overrides_applied: 0
human_verification:
  - test: "Owner decision on the live ladder finding (WINDOWS ledger #1, open): google-chrome 153 reports caps 0x29 (CapSurroundingText set) and APPLIES DeleteSurroundingText — the actual level on this target is 1 for every matrix surface (zenity, chromium, GTE); the plan/ADR-003 assumption of a Chromium level-2 fallback witness (ibus#2354 lore) is falsified on this desktop and no live level-2 witness exists (level-2 contract is carried by TestActor_Level2NoCaps/Level2WithTailAndRunes). ADR-003 was deliberately left unedited (Accepted, owner gate)."
    expected: "Owner reviews the finding and either annotates ADR-003/ROADMAP SC-3 wording («fallback-уровень» → «фактический уровень») or accepts the deviation as recorded; WINDOWS ledger entry #1 moves from open to resolved."
    why_human: "Accepted-ADR text and roadmap wording are owner-gated documents; the executor explicitly deferred this edit to verify-work. Nothing in code needs to change — this is a documentation-acceptance decision."
  - test: "Visual no-jump confirmation (phase goal clause «без визуального прыжка»): with goswitch-en active, type ghbdtn in a real GNOME input field (e.g. gedit or a browser box), double-tap Right Shift and watch the replacement."
    expected: "The word becomes привет in place — one atomic replacement over exactly the wrong-word range, no visible cursor jump, flicker or transient doubled text."
    why_human: "e2e proves content-exactness of the field after the gesture (rune-exact AT-SPI readback, no doubling) and the range arithmetic is unit-pinned, but the perceptual «no visual jump» property of the DeleteSurroundingText+CommitText transaction is inherently visual and cannot be asserted through AT-SPI."
---

# Phase 2: Коррекция слова EN↔RU — Verification Report

**Phase Goal (ROADMAP.md):** «Пользователь двойным нажатием Right Shift исправляет последнее слово, набранное не в той раскладке, в любом поле ввода GNOME — направление определяется автоматически, регистр сохраняется, замена происходит ровно по диапазону неверного текста без визуального прыжка; это подтверждает зелёная e2e-матрица v1.»
**Verified:** 2026-09-15T00:55:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Verification Approach Note (MVP mode discrepancy)

ROADMAP.md declares `**Mode:** mvp` for Phase 2, but the goal is a
Russian user-capability statement, not a canonical User Story
(`As a …, I want to …, so that …`) — confirmed via
`gsd_run query user-story.validate` → `valid: false`. Per
`gsd-core/references/verify-mvp-mode.md` the MVP user-flow framing fires
only when BOTH `mode: mvp` AND a user-story goal are present, and a
non-story goal must be surfaced as a discrepancy — the same resolution
Phase 1's 01-VERIFICATION.md applied. Standard goal-backward verification
against the 5 Success Criteria + plan must_haves was therefore applied.
Because the goal is still user-action-shaped, a User Flow Coverage table
is included below as evidence framing. **Discrepancy surfaced for the
owner:** run `/gsd mvp-phase 2` if a User-Story goal is wanted; no action
required otherwise.

## Live-gate evidence policy

Live e2e (`mise run e2e-matrix` and the per-case tasks) drives the owner's
live desktop and was NOT re-run by the verifier (Phase 1 precedent).
Instead, the live matrix claim was verified against **machine-generated,
GitHub-side evidence independently re-fetched by the verifier**:

- `gh run view 34913812782` → `conclusion: success`, `event:
  workflow_dispatch`, `headBranch: gsd/phase-02-korrektsiya-slova-en-ru`
  (the phase branch), workflow `e2e-matrix`, created 2026-09-15T00:35:29Z.
- The run's `e2e-report` artifact was **re-downloaded from GitHub** by the
  verifier (`gh run download` → fresh dir) and matches byte-for-byte the
  executor-cited copy: **16/16 PASS** including all three gnome-text-editor
  registers, both directions, the untouched mixed word, all four reset
  cases and the chromium case.
- Evidence currency: the run's headSha `419799d` differs from HEAD
  `6fad577` by **docs/planning files only** (`git diff --stat` verified —
  REQUIREMENTS/ROADMAP/STATE/02-07-SUMMARY/ci-runner.md); zero Go/YAML
  code changed between the green CI run and the verified HEAD.
- Headless gates (test corpus under -race, full `mise run ci`, purity
  greps, `go mod tidy` drift) WERE re-run by the verifier and are green.

## User Flow Coverage

User capability (goal, non-story form): «Пользователь двойным нажатием
Right Shift исправляет последнее слово, набранное не в той раскладке, в
любом поле ввода GNOME».

| Step | Expected | Evidence | Status |
|------|----------|----------|--------|
| Type a word in the wrong layout | `ghbdtn` (and `GHBDTN`/`Ghbdtn`/`привет`/`ghbdtn2026`/`ghbdtn,`) lands in the field | matrix-v1.yaml cases; AT-SPI readback path test/e2e/matrix.go:835 verifyMatrixText | ✓ |
| Double-tap Right Shift | FSM Double → two-phase correction pipeline | actor.go:261 startCorrection; live log `{"msg":"action","n":2}` gated in every case | ✓ |
| Word is corrected, direction automatic | EN→RU and RU→EN both work; mixed word untouched | CI artifact: word-en-ru, word-ru-en, word-mixed-untouched PASS; Detect corpus direction_test.go | ✓ |
| Register preserved | three registers on gnome-text-editor (criterion 1 verbatim) | CI artifact: word-gte / word-gte-upper / word-gte-capitalized PASS | ✓ |
| Replacement exactly over the wrong range | token+tail deleted and recommitted, one transaction | BuildPlan formulas + TestActor_AfterSpaceCorrects; live word-after-space PASS with rune-exact «привет » | ✓ |
| Works in any GNOME input field | GTK (zenity), Chromium, GtkSourceView (gnome-text-editor) | CI artifact 16/16 across three surfaces | ✓ |
| Green matrix proves it | 16/16 PASS, exit≠0 contract | Re-downloaded CI artifact + matrix_test.go exit-code pin | ✓ |
| Outcome: no visual jump | replacement perceived as in-place | content-exactness proven; perceptual check → Human Verification item 2 | ⚠ human |

## Goal Achievement

### Observable Truths

Consolidated from the 5 ROADMAP Success Criteria (primary contract) merged
with the seven plans' must_haves (≈40 plan-level truths deduplicate into
these).

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC-1: gnome-text-editor double Right Shift corrects `ghbdtn`→`привет`, `GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет` | ✓ VERIFIED | matrix-v1.yaml cases word-gte/-upper/-capitalized; **re-downloaded CI artifact: all three PASS** (AT-SPI readback, per-case fresh daemon) |
| 2 | SC-2: direction auto-detected from buffer script both ways + unit word corpus | ✓ VERIFIED | correct.Detect (direction.go:22) Latin/Cyrillic counts; TestDetect_* green under -race; live word-en-ru + word-ru-en + word-mixed-untouched PASS in CI artifact |
| 3 | SC-3: replacement exactly over the wrong-word range, no jump; both ladder levels; abort-not-litter; Chromium matrix case pins the level | ✓ VERIFIED (see Human item 1) | Level 1 live on all 3 surfaces (CI artifact, exact-range `DeleteSurroundingText(-n,n)+CommitText`); level 2 unit-pinned (TestActor_Level2NoCaps/Level2WithTailAndRunes **re-run PASS**; wire pin TestEmitters_ForwardKeyEvent); abort discipline TestActor_VerifyPaths/EmptyBufferNoDestructive; Chromium case pins ACTUAL level — plan's level-2 assumption falsified live (chrome 153 applies the deletion), deviation documented + WINDOWS ledger #1 open for owner |
| 4 | SC-4: buffer cleared on Enter, Tab, Escape, focus change — e2e-proven on client classes | ✓ VERIFIED | isResetKeyval table (actor.go:554) + HandleLifecycle FocusOut/Reset→HardReset; TestActor_HardResetKeyvals/TestBuffer_ResetByFocusOut **re-run PASS**; live reset-escape/reset-enter/reset-tab/reset-focus all PASS in CI artifact |
| 5 | SC-5: e2e-matrix v1 green — YAML cases, PASS/FAIL report, exit≠0 | ✓ VERIFIED | matrix.go strict KnownFields loader + per-case setupStand/teardown isolation + exitCode(); matrix_test.go (6 tests) green; **CI run success + artifact 16/16**; corrupted-expectation → FAIL/exit 1 proven live by executor (02-06 D4) |
| 6 | Token semantics D-13/D-14/D-15 + toggle invariant (02-01) | ✓ VERIFIED | buffer.go Push/Token/Tail/ReplaceToken/Backspace + corpus (7 buffer tests) green under -race |
| 7 | Mixed-word refusal D-16; digits/punctuation neutral (CORR-04) | ✓ VERIFIED | Detect single-foreign-letter refusal; TestDetect_MixedRefused/NeutralNotMixed green; live word-mixed-untouched PASS («gfbпривет» unchanged) |
| 8 | Conversion preserves register per rune; non-letters identical (CORR-05) | ✓ VERIFIED | convert.go table lookup; TestConvert_CasePreserved green; live word-upper/word-capitalized PASS |
| 9 | Suffix verification ADR-004 (token+tail range) | ✓ VERIFIED | verify.go MatchesSuffix; TestMatchesSuffix green; wired at actor.go:201/214/457 |
| 10 | BuildPlan both levels, all counts in runes (CORR-07, Pitfall 2) | ✓ VERIFIED | plan.go:43 len([]rune) arithmetic; TestBuildPlan_BothLevels/RunesNotBytes green |
| 11 | internal/correct purity: no godbus/time/goroutines | ✓ VERIFIED | grep gate re-run: PURE-PACKAGE; corpus headless green under -race |
| 12 | Surface drivers: fresh chromium (temp --user-data-dir) + standalone GTE (XDG_DATA_HOME isolation), witness-gated | ✓ VERIFIED | surface.go:71 (--user-data-dir), :131 (--standalone --ignore-session), :132 (XDG_DATA_HOME temp); witness/pid-keyed focus paths; smoke cases in registry |
| 13 | Preflight rejects missing binaries with one-line diagnostics | ✓ VERIFIED | preflight.go chromium-launch / gnome-text-editor-launch checks present |
| 14 | Tracer: double tap runs buffer→direction→convert→verify→delete+commit end-to-end (CORR-01) | ✓ VERIFIED | TestActor_DoubleTapCorrects **re-run PASS**; live word-en-ru PASS (CI artifact) |
| 15 | Word after space corrected with separator (D-13, Pitfall 1) | ✓ VERIFIED | TestActor_AfterSpaceCorrects (-7,7 + «привет ») PASS; live word-after-space PASS, rune-exact trailing space |
| 16 | Log discipline D-20/D-21: INFO outcome/reason without words; DEBUG level-first | ✓ VERIFIED | logCorrectionDone (actor.go:507); TestActor_DebugCorrectionRecord scans INFO records — green |
| 17 | Flip on Single: mode EN↔RU, INFO mode record, XKB untouched (02-04) | ✓ VERIFIED | flipScript (actor.go:340); TestActor_FlipOnSingle PASS; live word-ru-en gated on `{"msg":"mode","to":"ru"}` |
| 18 | RU consumption commits Cyrillic via ENToRU; script-true buffer in ALL printing branches incl. digits | ✓ VERIFIED | feedKey (actor.go:356-398); TestActor_RUConsumesPrintable/RUScriptTrueAllBranches/RUDigitsFullPipeline (10-rune range, not 14 bytes) **re-run PASS** |
| 19 | Level 2: ForwardKeyEvent(BackSpace)×(token+tail runes) burst → one commit, in order (CORR-07) | ✓ VERIFIED | executeLevel2 (actor.go:493) + engine.ForwardKeyEvent emitter + wire pin; TestActor_Level2NoCaps/WithTailAndRunes **re-run PASS** (op-log order) |
| 20 | Verify-after with epoch: mismatch → INFO counter, never auto-repair; stale timers no-op | ✓ VERIFIED | armAfterVerify/pendingAfter/verifyEpoch (actor.go:309-331); TestActor_VerifyAfterLevel1 **re-run PASS** (match quiet / mismatch counts once) |
| 21 | Reset triggers wired live (CORR-09): Escape/Enter/Tab/focus cases green | ✓ VERIFIED | CI artifact reset-escape/reset-enter/reset-tab/reset-focus PASS; ydotool name traps (space→S, Escape→E) fixed via typing path / `esc` |
| 22 | Matrix infrastructure: strict decoder, per-case fresh-daemon isolation, expect_level = actual level, order-independence | ✓ VERIFIED | matrix.go KnownFields(true), runMatrixCaseIsolated (setupStand per case), closed vocabularies; matrix_test.go 6/6; shuffled-order 16/16 (executor, documented) |
| 23 | CI circuit D-19: dispatch-only workflow on self-hosted gnome runner, first green dispatch | ✓ VERIFIED | e2e-matrix.yml (workflow_dispatch only, [self-hosted, gnome], concurrency e2e-gnome no-cancel, contents: read, mise run e2e-matrix, artifact if:always()); runner unit active + exactly 1 gnome-labeled runner online (re-checked); run 34913812782 success from the phase branch; bootstrap commit d815e72 on main |

**Score:** 23/23 truths verified (0 present-but-behavior-unverified)

Behavior-dependent truths note: every state-transition/invariant truth in
this phase (correction pipeline, verify abort paths, level-2 burst order,
verify-after epoch, reset table, flip, script-true feeding, toggle) is
covered by a named behavioral test **re-run by the verifier under -race**
(all PASS); the live-desktop dimension is covered by the independently
re-fetched CI matrix artifact at the exact code state of HEAD.

### Decision Coverage

`gsd_run query check.decision-coverage-verify` (02-CONTEXT.md): **9/9
decisions honored** by shipped artifacts, none missing (non-blocking gate).

### Required Artifacts

`gsd_run query verify.artifacts` per plan: **27/28 passed** — 02-01 5/5,
02-02 4/4, 02-03 4/4, 02-04 3/3, 02-05 4/4, 02-06 4/4, 02-07 3/4. The
single "miss" is a **verifier-tool false negative**: the 02-07 artifact
`~/.config/systemd/user/goswitch-ci-runner.service` exists at
`/home/nil/.config/systemd/user/goswitch-ci-runner.service` (750 bytes,
verified) — gsd-tools does not expand `~`; the plan itself declares the
unit "вне git". All artifacts are substantive (no stub markers; internal/
correct 882 LOC, actor 603, matrix runner 910, corpus 46k+ LOC tests).

### Key Link Verification

`gsd_run query verify.key-links` per plan: **15/16 verified**. The one
unverified link (02-01 buffer.go→direction.go, pattern `\.Token\(\)`) is a
**plan-spec false negative**: the declared from/to are both inside the pure
package, but the actual consumer is the actor —
`internal/session/actor.go:415` (`a.buf.Token()`) feeding
`correct.Detect(token)` at :421, with `ReplaceToken` at :482/:499. Wiring
exists and is behavior-tested. All other links verified by the tool,
including actor→engine emitters (AttachEngine), matrix→stand primitives,
workflow→mise task, ci-runner doc→systemd precedent.

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| internal/session/actor.go | token/tail | buf.Token()/Tail() fed by HandleKey from live ProcessKeyEvent | ✓ live key events (CI matrix: corrections fired on injected input) | ✓ FLOWING |
| internal/session/actor.go | surr | HandleSurroundingText from real client pushes (cache) + Require round | ✓ live pushes observed (word cases correct through verification) | ✓ FLOWING |
| internal/correct/plan.go | Plan | BuildPlan from token/tail/converted + caps bit from HandleCapabilities | ✓ real caps 0x29 observed live; level dispatch exercised | ✓ FLOWING |
| test/e2e/matrix.go | report | per-case runMatrixCaseIsolated results | ✓ 16 PASS lines in re-downloaded CI artifact | ✓ FLOWING |
| layouts/tables.go | ENToRU/RUToEN | xkb-generated (Phase 1, byte-identical regeneration) | ✓ RU-commit path live (word-ru-en) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Headless corpus (4 pkgs) | `go test ./internal/correct/ ./internal/session/ ./engine/ ./test/e2e/ -race -count=1` | all ok | ✓ PASS |
| Full CI gate (build+vet+lint+race) | `mise run ci` | exit 0, lint 0 issues, 7 pkgs ok | ✓ PASS |
| Level-2 ladder + verify-after | `go test ./internal/session/ -run 'TestActor_Level2NoCaps\|Level2WithTailAndRunes\|VerifyAfterLevel1' -race -v` | 3/3 PASS (5 subtests) | ✓ PASS |
| Tracer/flip/reset/mixed/digits/toggle | `go test ./internal/session/ -run 'DoubleTapCorrects\|AfterSpaceCorrects\|HardResetKeyvals\|RUDigitsFullPipeline\|MixedWordUntouched\|FlipOnSingle\|ToggleRepeat' -race -v` | 7/7 PASS | ✓ PASS |
| Package purity | `! grep godbus && ! grep '"time"' internal/correct/` | PURE-PACKAGE | ✓ PASS |
| Module tidy (no drift beyond yaml.v3) | `go mod tidy && git diff --exit-code -- go.mod go.sum` | clean | ✓ PASS |
| CI matrix run (live) | `gh run view 34913812782` + `gh run download` (re-fetched) | conclusion success; artifact 16/16 PASS, byte-identical to executor copy | ✓ PASS |
| Runner live state | `systemctl --user is-active` + `gh api runners` | active; exactly 1 gnome-labeled | ✓ PASS |
| Commits exist | `git cat-file -t` × 24 hashes from SUMMARYs | all present | ✓ PASS |
| Live e2e re-run (16 cases) | not re-run (drives owner's desktop) | CI artifact at HEAD-equivalent code cited instead | ? SKIP (by policy) |

### Probe Execution

Step 7c: no `scripts/*/tests/probe-*.sh` convention in this repo. The
phase's runnable probes are the mise e2e tasks (live-session; covered by
the CI-artifact policy above) and the CI gate (re-run, PASS). Not
applicable beyond that.

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| internal/correct/*_test.go (5 files, 16 tests) | CORR-04/05/07 | 16 | 0 | 0 | value (rune-exact) | OK |
| internal/session/actor_test.go (29 tests) | CORR-01/04/07/09 | 29 | 0 | 0 | behavioral (op-log order, log-shape scans) | OK |
| engine/{engine,emitter_wire}_test.go | CORR-07 | 9 | 0 | 0 | value + wire-signature | OK |
| test/e2e/matrix_test.go (6 tests) | TEST-04 | 6 | 0 | 0 | value (decode errors, exit code) | OK |

- Disabled tests on requirements: **0** (grep for skip patterns clean).
- Circular patterns: **0** — expected values are hand-authored literals
  (ghbdtn/привет…) independent of the system; layout tables regenerate
  byte-identically from system xkb (external source).
- Assertion strength: matrix oracles are rune-exact UTF-8 equality; actor
  corpus pins exact emitter argument sequences and log shapes. No
  existence-only assertions found on requirement-linked tests.
- 81 test functions across 7 packages; 16 matrix cases ≥ the D-18 corpus
  list (6+1+1+5+3).

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| CORR-01 | 02-03, 02-04 | Double Right Shift corrects last word EN↔RU | ✓ SATISFIED | truths 14, 17; live word-en-ru + word-ru-en PASS |
| CORR-04 | 02-01, 02-04 | Auto direction by buffer composition | ✓ SATISFIED | truths 2, 7, 18; no dictionaries (grep clean) |
| CORR-05 | 02-01 | Register preserved per rune | ✓ SATISFIED | truth 8; three registers live on zenity + GTE |
| CORR-07 | 02-01, 02-03, 02-05 | Exact range, both ladder levels, no jump | ✓ SATISFIED | truths 3, 9, 10, 15, 19, 20 (wording deviation → Human item 1) |
| CORR-09 | 02-05 | Buffer reset Enter/Tab/Escape/focus | ✓ SATISFIED | truth 4, 21 |
| TEST-04 | 02-02, 02-06, 02-07 | YAML matrix, report, exit≠0, multiple apps | ✓ SATISFIED | truths 5, 12, 22, 23 |

Orphaned requirements: **none** — REQUIREMENTS.md traceability maps exactly
these 6 IDs to Phase 2 (all marked Complete, updated at HEAD); every plan
requirement field resolves to them.

### Anti-Patterns Found

Debt-marker gate: **ZERO** TBD/FIXME/XXX and zero TODO/HACK/PLACEHOLDER
across all phase Go/Python/YAML/TOML/workflow/doc files (re-grepped).

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| internal/session/actor.go | 195-220 | Verify-after settles on the FIRST post-correction push — clients can push the intermediate post-delete pre-commit state → false `correction verify outcome:mismatch` INFO on a healthy correction (deferred-items #2, open) | ⚠️ Warning | Log-noise only: no repair, no field damage; matrix deliberately does not pin the match record; candidate daemon-side fix (settled-answer grace) needs its own plan |
| test/e2e/case_m1.go | closeEntrySurface | `pressKey(ctx, "Escape")` — ydotool 0.1.8 resolves "Escape" to physical E; locked-session shell fallback would type 'e' instead of dismissing (deferred-items #3, open) | ⚠️ Warning | Stand tooling fallback path only, not exercised by any phase case; one-word fix ("esc") queued for next touch of the shell surface |
| desktop environment | — | Persistent AT-SPI shell-bridge wedge after repeated surface SIGKILLs (deferred-items #1, open) | ℹ️ Info | Environmental, not code; remedy (a11y bus restart) documented in the runner book; matrix carries a between-cases quiesce gate |
| .planning/.../02-VALIDATION.md | — | Left as unfilled draft template (status: draft, nyquist_compliant: false) while the phase's actual validation ran through PLAN verify-blocks + SUMMARY coverage records | ℹ️ Info | Process-artifact gap only; no plan claims it and the ROADMAP does not gate on it |
| test/e2e/case_ladder.go | 23 | `actualLevelMark = "level":1` — no live level-2 witness exists on this desktop (all surfaces report the caps bit) | ⚠️ Warning (owner decision) | Level-2 correctness rests on the unit corpus + wire pin (adequate for the contract); live witness would need a no-caps client — see Human item 1 |

Judgment: none of the warnings invalidates a phase must-have. The
verify-after race degrades an observability counter (documented,
matrix-decoupled); the ydotool/AT-SPI items are stand/environment tooling
with documented remedies; the level-witness question is the owner decision
already flagged. No gaps.

### Human Verification Required

1. **Owner decision — live ladder finding vs ADR-003/ROADMAP wording**
   **Test:** review WINDOWS ledger entry #1 (kind: deviation, status:
   open) and the 02-05 finding: google-chrome 153 reports caps 0x29 and
   APPLIES DeleteSurroundingText — actual level 1 on every matrix surface;
   ibus#2354 lore is obsolete on this target; no live level-2 witness
   exists (level-2 contract carried by TestActor_Level2NoCaps /
   Level2WithTailAndRunes / TestEmitters_ForwardKeyEvent).
   **Expected:** owner either annotates ADR-003/ROADMAP SC-3 («fallback» →
   «фактический уровень») or accepts the deviation; ledger entry resolved.
   **Why human:** ADR-003 is Accepted under the owner gate and was
   deliberately left unedited by the executor pending this review.

2. **Visual no-jump confirmation (goal clause «без визуального прыжка»)**
   **Test:** with goswitch-en active, type `ghbdtn` in a real GNOME field
   (gedit / browser box), double-tap Right Shift, watch the replacement.
   **Expected:** word becomes «привет» in place — atomic replacement over
   exactly the wrong range, no cursor jump/flicker/doubling.
   **Why human:** e2e proves rune-exact field content and the range
   arithmetic is unit-pinned, but the perceptual no-jump property of the
   DeleteSurroundingText+CommitText transaction is visual by nature.

### Gaps Summary

**No gaps.** All 23 consolidated must-have truths verified; 27/28
artifacts pass (1 verifier-tool `~` false negative, file present); 15/16
key links wired (1 plan-spec false negative, wiring proven in actor.go);
6/6 requirements satisfied with zero orphans; decision coverage 9/9; zero
debt markers; zero disabled/circular tests; `mise run ci` green and the
16/16 matrix evidence independently re-fetched from GitHub at the exact
code state of HEAD. The `human_needed` status comes solely from the two
owner items above (ADR-003 wording acceptance + perceptual no-jump check).

---

_Verified: 2026-09-15T00:55:00Z_
_Verifier: Claude (gsd-verifier)_
