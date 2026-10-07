---
phase: 02-korrektsiya-slova-en-ru
verified: 2026-10-07T07:49:59Z
status: human_needed
score: 22/23 must-haves verified
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
  - README.md
  - SECURITY.md
  - cmd/goswitchd/main.go
  - docs/SPEC.md
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
covered_digest: "v1:sha256:5672f34b1f2ef6bc71fec1f2fce86748816e76280a0049e09aa8bf880d511f76"
behavior_unverified: 1
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 23/23
  previous_verified: 2026-09-20T14:30:42Z
  reason: "fingerprint-staleness refresh at milestone close (v1.0.0): quick task 261006-vqw revised the mixed-text semantics to per-character layout inversion (spec-delta D-55, 2026-10-06/07 — D-22/D-23 superseded by owner verdict 2026-10-03, commit 7dd46e9) and quick task 261007-0yg internalized engine/ and layouts/ under internal/ (commit 5427d7a, history-preserving git mv) — the prior covered paths engine/*.go and layouts/tables.go are stale; all paths re-derived against the current tree"
  gaps_closed: []
  gaps_remaining: []
  regressions: []
behavior_unverified_items:
  - truth: "SC-5 (live dimension): the current matrix-v1.yaml — with its two re-pinned rows word-punct («приветб», owner directive 2026-09-28) and word-mixed («паиghbdtn», spec-delta 2026-10-06) — has no live green run yet"
    test: "On the live GNOME stand (after gsd/internalize-engine-layouts merges, or locally by the owner): mise run e2e-matrix"
    expected: "16/16 PASS, exit 0; the re-pinned rows read «приветб» and «паиghbdtn»; all other rows unchanged green (registers, D-13 separator, D-15 digits, resets, both directions)"
    why_human: "Drives the owner's live desktop (IBus engine + ydotool injection + AT-SPI readback); headless tests only prove the strict loader/oracle plumbing, never the live desktop"
human_verification:
  - test: "Live matrix v1 on the GNOME stand: mise run e2e-matrix (phase-2 slice of 261006-vqw human item 1)"
    expected: "16/16 PASS — after the double tap the field reads «паиghbdtn» (word-mixed, every letter inverted) and «приветб» (word-punct, the comma is the б key); homogeneous/register/reset rows unchanged green; exit code 0"
    why_human: "Live-desktop gate: real IBus engine, real GTK/chromium surfaces, real ydotool injection — not reproducible headless"
  - test: "Hand-typed mixed-text check in gedit/GTK4 (261006-vqw human item 5): type a Latin word, flip, type a Cyrillic word (e.g. gfb + привет), double-tap Right Shift"
    expected: "Every letter inverts (паиghbdtn); digits/space/punctuation in the mixed range ride unchanged; no visual jump; correction-done record in the daemon log"
    why_human: "Real-time keyboard/IME interaction and visual result — not reproducible programmatically"
---

# Phase 2: Коррекция слова EN↔RU — Verification Report

**Phase Goal (ROADMAP.md):** «Пользователь двойным нажатием Right Shift исправляет последнее слово, набранное не в той раскладке, в любом поле ввода GNOME — направление определяется автоматически, регистр сохраняется, замена происходит ровно по диапазону неверного текста без визуального прыжка; это подтверждает зелёная e2e-матрица v1.»
**Verified:** 2026-10-07T07:49:59Z
**Status:** human_needed (22/23 verified; 1 truth present-and-wired but its live-desktop witness for the re-pinned rows is pending — the same item 261006-vqw tracks)
**Re-verification:** Yes — fingerprint-staleness refresh at milestone close (prior pass 2026-09-20T14:30:42Z, 23/23; original pass 2026-09-15T05:40:42Z, UAT 2/2 owner-accepted; no prior gaps)

## Re-verification Approach

Since the prior fingerprint tree (04d00ad) the phase-2 surface changed through two
quick tasks and the later phases that shipped after it:

1. **261006-vqw — mixed-text per-character layout inversion** (commits
   a24afa9/3c7f564/4da90d8/f9ed497, spec-delta D-55 BEFORE code, strict-TDD
   RED→GREEN): `internal/correct/runs.go` — the mixed branch of `ConvertRuns`
   now calls `invertPerChar` (every Latin letter through ENToRU, every
   Cyrillic through RUToEN, neutrals ride, unmapped letter fails the WHOLE
   range); the D-22/D-23 last-letter anchor (`convertForeignRuns`) is deleted.
   `convert.go` — table-mapped punctuation converts with the token
   (","→"б"; owner directive 2026-09-28; "ghbdtn,"→"приветб"). Golden mixed
   corpus re-pinned in `runs_test.go`/`actor_test.go`; e2e oracles re-pinned
   in `case_word.go` and matrices v1..v4. The homogeneous wholesale branch is
   behavior-identical (261006-vqw truth 3: byte-level verification vs the
   pre-plan tree, including the deliberate 3-line mixed-case move only).
2. **261007-0yg — internalization** (commit 5427d7a, history-preserving
   `git mv` + import-path rewrite in 20 files; go.mod/go.sum untouched):
   `engine/*` → `internal/engine/*`, `layouts/tables.go` →
   `internal/layouts/tables.go`. All prior covered paths re-derived; the
   moved phase-2 files (engine.go, factory.go, keys.go, types.go,
   engine_test.go, emitter_wire_test.go, tables.go) verified present and
   substantive at the new paths.
3. **Later phases (5–8)** grew `internal/session/actor.go` (+1576 lines:
   autocorrect, selection, blocklist, menu callbacks) and
   `test/e2e/matrix.go` (+263 lines: v3/v4 support) around the phase-2
   pipeline; `cmd/goswitchd/main.go` +472; workflow +44 (nightly schedule +
   soft-mode preflight). The phase-2 load-bearing surfaces were re-checked:
   `dec.KnownFields(true)` still at matrix.go:367; the mixed branch wiring
   `runs.go:98 → invertPerChar`, homogeneous `runs.go:100 →
   Convert(text, dirOf(script))`; actor wiring `buf.Token()` →
   `startRangeCorrection` → `ConvertRuns` (actor.go:2195/2248/2457/2577).

**Classification of the diff** (`git diff 04d00ad..HEAD` per file): deep-checked —
runs.go, convert.go, buffer.go (+PushFeed, Push now a wrapper; token semantics
unchanged), runs/convert/buffer tests, case_word.go, matrix-v1.yaml (2 rows),
matrix.go, matrix.go loader, workflow, engine+layouts at new paths; regression-checked
byte-identical or import-path-only — plan.go, verify.go, direction.go, their tests,
hotkey/fsm.go, surface.go, case_surface.go, case_ladder.go, case_resilience.go,
focus_helper.py, fixtures, case_m1.go, case_d01.go (journal-row quoting), preflight.go
(import path only), mise.toml, README/SECURITY/.gitignore, all 14 phase PLAN/SUMMARY files.

## Live-gate evidence policy (unchanged) and its honest state

Live e2e (`mise run e2e-matrix*`) drives the owner's desktop — NOT re-run by
the verifier (phase 1–8 precedent). GitHub-side evidence re-fetched 2026-10-07
via `gh run list` / `gh run view --log`:

| Run | Trigger | headSha | Result | Relevance |
|-----|---------|---------|--------|-----------|
| 37585613271 (2026-10-07) | schedule | 42cbe72 | "success" — **soft-skip** (session age 398222s > 1800s, D48_SKIP; no cases ran) | Not a matrix witness. Also predates the re-pin: origin/master is at 42cbe72 (phase-7 merge); the re-pin commits live on this branch, pending merge |
| 37429931098 (2026-10-06) | schedule | 42cbe72 | "success" — **soft-skip** (D48_SKIP) | Not a matrix witness |
| 37183775834 / 37102882019 (10-04/10-03) | schedule | e13af69 | "success" — **soft-skips** (D48_SKIP ×3 lines each) | Not matrix witnesses |
| Owner live runs 2026-10-05 (07-UAT test 4, tree c84709c/4b75680) | local | c84709c | matrix v4 ×2, **31/34 both** — 3 fails = 2 designed WINDOWS-12 mixed rows (byte-identical to the ledger; the freeze the owner then REJECTED) + 1 ibus activation transient | The most recent REAL live run of the correction semantics: homogeneous word/punctuation rows green live; only the mixed rows failed — as designed, pre-re-pin |
| 35108412175 (2026-09-16) | dispatch | 9f2fd75 | success, **D-48 double-run 31/31 PASS twice** | Historic live proof of word/phrase/register/exact-range/surface invariants at phase-3-era production code (superseded for mixed semantics by the re-pin) |
| pr-sanity 37543084250 (2026-10-06 22:50Z) | push | **1efc6fa (HEAD)** | success | pr-sanity green at the current HEAD (build + vet + lint + test -race + tidy + govulncheck) |
| runner | — | — | **green106 online** (API re-check 2026-10-07; labels self-hosted, Linux, X64, gnome) | CI circuit alive |

**Honest recency note (the one open item):** NO live run has ever executed the
re-pinned mixed rows. The 2026-10-05 owner runs predate the re-pin (their
mixed rows failed exactly as the then-designed WINDOWS-12 ledger expected,
prompting the owner verdict that produced the re-pin); the four "green"
nightlies since 2026-10-03 are all D48_SKIP soft-skips; and the re-pin itself
sits on `gsd/internalize-engine-layouts` pending merge (nightlies run
origin/master@42cbe72). Headless, the re-pinned semantics are fully proven
(RED→GREEN TDD independently reproduced in 261006-vqw's verification; the
matrix expect_text values are byte-identical to the unit-pinned
`ConvertRuns` outputs). The live witness is tracked as 261006-vqw's
human-needed items; phase-2's slice of it is truth 5 below (SC-5) and is
surfaced in Human Verification Required.

## User Flow Coverage

User capability (goal, non-story form — MVP user-flow framing discrepancy
recorded in the initial pass, unchanged).

| Step | Expected | Evidence | Status |
|------|----------|----------|--------|
| Type a word in the wrong layout | `ghbdtn` (registers/mixed/digits variants) lands in the field | matrix-v1.yaml 16 cases intact at HEAD (2 rows re-pinned per owner-approved spec-delta); case_word.go registry wired | ✓ |
| Double-tap Right Shift | FSM Double → correction pipeline | TestActor_DoubleTapCorrects PASS in this verifier's `-race` run at HEAD 1efc6fa; live in D-48 double-run @9f2fd75 and Oct-5 owner runs | ✓ |
| Word corrected, direction automatic | EN→RU and RU→EN; mixed → per-character inversion (D-55, supersedes D-23 foreign-run) | TestActor_MixedWordInvertsPerChar + TestConvertRuns_MixedInvertsPerChar PASS at HEAD; homogeneous direction via `Convert(text, dirOf(script))` (runs.go:100); live both directions @9f2fd75 + Oct-5 | ✓ (mixed live row pending — see truth 5) |
| Register preserved | three registers | TestConvert_CasePreserved PASS (this verifier's run); live phrase-registers + gedit-phrase-registers @9f2fd75; v1 GTE register rows unchanged since live 16/16 @419799d | ✓ |
| Exact-range replacement, no jump | delete+commit over the range, one transaction | BuildPlan rune arithmetic (plan.go untouched) + TestActor_AfterSpaceCorrects PASS; live phrase-after-space @9f2fd75; perceptual check owner-accepted (UAT 2) | ✓ |
| Works in any GNOME input field | GTK/Chromium/GtkSourceView + gedit + chromium-X11 drivers | live v3 31/31 twice @9f2fd75; v4 31/34×2 @c84709c (fails = designed mixed rows + transient); surface.go/focus_helper.py untouched since | ✓ |
| Green matrix proves it | 16/16 v1; exit≠0 | infra green at HEAD (strict loader matrix.go:367, matrix_test.go in the quick-task `mise run ci` greens); **live green of the current v1 file pending** — 2 rows re-pinned post-pin, no live run yet | ⚠️ live witness pending |

## Goal Achievement

### Observable Truths

Same 23 consolidated truths as the accepted prior passes; evidence re-derived
at HEAD 1efc6fa (R = regression-checked, byte-identical or import-path-only
since 04d00ad; D = deep-checked, changed by the quick tasks / later phases;
wording updated where the owner-approved spec-delta superseded a mechanism).

| # | Truth | Status | Evidence at HEAD 1efc6fa |
|---|-------|--------|--------------------------|
| 1 | SC-1: gnome-text-editor double Right Shift corrects `ghbdtn`→`привет`, `GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет` | ✓ VERIFIED | v1 register rows unchanged (R); live 16/16 @419799d + registers re-proven live at phrase width @9f2fd75 and Oct-5; TestActor_DoubleTapCorrects/TestConvert_CasePreserved PASS at HEAD (this verifier's `-race` run) |
| 2 | SC-2: direction auto-detected both ways + unit corpus | ✓ VERIFIED (mechanism succession) | homogeneous branch `Convert(text, dirOf(script))` intact (runs.go:100, R); single-script → wholesale — TestActor_DoubleTapCorrects PASS; TestDetect_* corpus green (Detect remains corpus-reference, see Anti-Patterns); live both directions @9f2fd75 + Oct-5; mixed ranges now per-character inversion per owner verdict 2026-10-03 / spec-delta D-55 (truth 7) |
| 3 | SC-3: replacement exactly over the wrong range, both ladder levels, abort-not-litter | ✓ VERIFIED | BuildPlan rune arithmetic intact (plan.go, R); D-27 level-2-only cap pinned (TestBuildPlan_BackspaceCap enumerated, TestActor_Level2NoCaps PASS); level-1 live on 5 surfaces @9f2fd75 + Oct-5; abort discipline green; Chromium actual level owner-accepted (UAT 1) |
| 4 | SC-4: buffer cleared on Enter, Tab, Escape, focus change | ✓ VERIFIED | isResetKeyval table + HardReset unchanged (R); TestActor_HardResetKeyvals **PASS at HEAD**; v1 reset rows (Escape/Enter/Tab/FocusOut) unchanged and live @419799d; FocusOut reset live @9f2fd75 |
| 5 | SC-5: e2e-matrix v1 green — YAML cases, PASS/FAIL report, exit≠0 | ⚠️ PRESENT_BEHAVIOR_UNVERIFIED | Infra verified at HEAD: strict loader `dec.KnownFields(true)` (matrix.go:367), per-case isolation + exitCode pinned by matrix_test.go (green in the 261006-vqw/261007-0yg `mise run ci` full `-race` runs); v1 file 16 cases, D-13 pin-geometry header intact, the 2 re-pinned rows byte-consistent with the unit-pinned ConvertRuns outputs. **Live dimension open:** no live run has executed the current v1 file — the re-pinned word-punct («приветб») and word-mixed («паиghbdtn») rows postdate every live run (last real: 2026-10-05 v4 ×2, pre-re-pin; nightlies since are D48_SKIP soft-skips; re-pin pending merge). Live witness = the human item below; 14 of 16 rows' semantics remain live-witnessed from earlier runs |
| 6 | Token semantics D-13/D-15 + toggle invariant (D-14 ride-along superseded by owner directive 2026-09-28: table-mapped punctuation now converts with the token) | ✓ VERIFIED | PushFeed adds the boundary flag only; Push wraps it — token semantics unchanged (buffer.go, D); buffer corpus green incl. TestBuffer_PushFeed, TestBuffer_TokenRules (enumerated; TestBuffer_PushFeed in this verifier's run); punctuation conversion pinned by TestConvertRuns_HomogeneousWholesale **PASS at HEAD** ("ghbdtn,"→"приветб") and live-green in the Oct-5 owner runs; digits ride (v1 word-digits row unchanged, live @419799d) |
| 7 | Mixed word (CORR-04) — per-character inversion (spec-delta D-55 2026-10-06; supersedes the D-23 foreign-run succession) | ✓ VERIFIED | `invertPerChar` wired at the mixed branch (runs.go:98→:118); TestConvertRuns_MixedInvertsPerChar (7 cases: both-orders inversion, digits/space/bracket ride, per-rune register «Паиghbdtn») + TestActor_MixedWordInvertsPerChar / TestActor_PhraseMixedCorrects / TestActor_SelectionMixedConverts **PASS at HEAD** (this verifier's `-race` run); TDD RED shape independently reproduced at d7a8a5e (exactly the 4 re-pinned tests fail) per 261006-vqw verification; e2e oracle case_word.go re-pinned («паиghbdtn»); unmapped-letter whole-range refusal pinned (TestConvertRuns_Refusals, é case, ok=false out=nil); live row rides truth 5's pending witness |
| 8 | Conversion preserves register per rune; non-letters per the 2026-09-28 rule (CORR-05) | ✓ VERIFIED | register logic unchanged (convert.go table pair, D — non-letter branch revised only); TestConvert_CasePreserved PASS; per-rune register under inversion pinned («Паиghbdtn»); live registers @9f2fd75 + Oct-5 |
| 9 | Suffix verification ADR-004 (token+tail range) | ✓ VERIFIED | MatchesSuffix intact (verify.go, R); VerifyRangeAt for selection (additive); TestMatchesSuffix **PASS at HEAD**; wired at pendingVerdict/afterVerdict (actor.go re-read) |
| 10 | BuildPlan both levels, all counts in runes (CORR-07) | ✓ VERIFIED | len([]rune) arithmetic intact (R); +backspaceCap param; TestBuildPlan_BothLevels/RunesNotBytes/TailArithmetic **PASS at HEAD** |
| 11 | internal/correct purity: no godbus/time/goroutines | ✓ VERIFIED | purity grep re-run at HEAD: **0 hits** in production files (the new internal/layouts import is a data-table package, inside the purity contract) |
| 12 | Surface drivers: fresh chromium + standalone GTE, witness-gated | ✓ VERIFIED | surface.go, case_surface.go, focus_helper.py byte-identical since 04d00ad (R); live gedit/x11/smoke rows @9f2fd75; v4 breadth rows @c84709c |
| 13 | Preflight rejects missing binaries | ✓ VERIFIED | preflight.go: import-path-only change (R); package green; live preflight ok in the D-48 log; the fresh-session/idle preflight steps re-observed live in the 2026-10-06/07 nightly logs ("desktop idle", session-age soft-skip) |
| 14 | Tracer: double tap runs buffer→direction→convert→verify→delete+commit (CORR-01) | ✓ VERIFIED | TestActor_DoubleTapCorrects **PASS at HEAD**; pipeline wiring re-read (buf.Token → startRangeCorrection → ConvertRuns, actor.go:2577/2457); live word/phrase cases @9f2fd75 + Oct-5 |
| 15 | Word after space corrected with separator (D-13) | ✓ VERIFIED | TestActor_AfterSpaceCorrects **PASS at HEAD**; live phrase-after-space PASS @9f2fd75; v1 word-after-space row unchanged |
| 16 | Log discipline D-20/D-21 | ✓ VERIFIED | logCorrectionDone; TestActor_DebugCorrectionRecord present (127 TestActor* enumerated; full -race suite green in the 261006-vqw/261007-0yg ci runs); D-20 pinned by ClipboardRungAfterMismatch («no content in INFO») |
| 17 | Flip on Single: mode EN↔RU, INFO mode record, XKB untouched | ✓ VERIFIED | flipScript; TestActor_FlipOnSingle **PASS at HEAD**; single flipTo path intact through phase-7 menu rewiring (07-VERIFICATION regression); live layout-single @9f2fd75 |
| 18 | RU consumption commits Cyrillic; script-true buffer in ALL printing branches incl. digits | ✓ VERIFIED | feedKey invariants; TestActor_RUDigitsFullPipeline **PASS at HEAD**; live phrase-ru-en @9f2fd75 |
| 19 | Level 2: ForwardKeyEvent(BackSpace)×(token+tail runes) burst → one commit, in order (CORR-07) | ✓ VERIFIED | executeLevel2 + emitter wire pin; TestActor_Level2NoCaps/WithTailAndRunes **PASS at HEAD**; over-cap range refuses (D-27, LevelNone) |
| 20 | Verify-after with epoch: mismatch → INFO counter, never auto-repair; stale timers no-op | ✓ VERIFIED (WR-05 semantics) | settleAfter unchanged; first stale → re-Require, second stale → mismatch once; epoch guard retained; TestActor_VerifyAfterLevel1 **PASS at HEAD** |
| 21 | Reset triggers wired live (CORR-09) | ✓ VERIFIED | reset-* live @419799d; FocusOut reset live @9f2fd75; TestActor_HardResetKeyvals PASS at HEAD |
| 22 | Matrix infrastructure: strict decoder, per-case isolation, closed vocabularies | ✓ VERIFIED | matrix.go KnownFields(true) intact at :367 (D — +263 lines are v3/v4 additive surface, loader untouched); matrix_test.go green (quick-task ci runs); quiesce invariant still pinned by quiesce_test.go |
| 23 | CI circuit D-19: dispatch-only workflow on self-hosted gnome runner | ✓ VERIFIED | e2e-matrix.yml re-read at HEAD (D — +44 lines: nightly schedule + soft-mode preflights; still **zero push/PR triggers**, WR-04 env-input kept, matrix input defaults v4); runner green106 **online** (API 2026-10-07); pr-sanity green at HEAD 1efc6fa (37543084250); live green circuit evidence: 35108412175 @9f2fd75 (31/31×2) + owner Oct-5 runs; nightlies soft-skip by design (session age), honestly recorded above |

**Score:** 22/23 truths verified (1 present-and-wired, live witness pending)

### Required Artifacts

All 67 covered files re-verified to exist at HEAD (path re-derivation:
engine/* → internal/engine/*, layouts/tables.go → internal/layouts/tables.go;
`verify.artifacts` per plan not re-run — the plan frontmatter is byte-identical
since 42d10a1 and the prior passes' single "~" false negative stands: the
verifier tool cannot see `~/.config/systemd/user/goswitch-ci-runner.service`,
which exists, 750 bytes). The moved engine/layouts artifacts are substantive
at their new paths (engine.go 468 lines, factory.go 76, keys.go 56, types.go
214, tables.go 207 + the later-phase additions conn.go/address.go/dict*/trigrams.go
tracked by their own phases).

### Key Link Verification

Carried: 15/16; the one unverified link remains the plan-spec false negative
(02-01 buffer.go→direction.go, pattern `\.Token\(\)`); the actual consumer
wiring re-read at HEAD: `internal/session/actor.go` `a.buf.Token()`
(:2195/:2248/:2577) → `startRangeCorrection` → `ConvertRuns` — present, wired,
behavior-tested. New links from the re-pin verified wired: runs.go:98 →
`invertPerChar` → both layout tables; case_word.go oracle ↔ unit pins
(byte-identical outputs, 261006-vqw key-link table).

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| internal/session/actor.go | token/tail | buf.Token()/Tail() fed by HandleKey from live ProcessKeyEvent | ✓ live corrections @9f2fd75 + Oct-5 | ✓ FLOWING |
| internal/correct/runs.go | out/changed | invertPerChar / Convert over the live range | ✓ pinned by the passing mixed/wholesale corpus; live homogeneous rows @Oct-5; live mixed row pending (truth 5) | ✓ FLOWING |
| internal/correct/plan.go | Plan | BuildPlan from token/tail/converted + caps bit | ✓ live level dispatch @9f2fd75 | ✓ FLOWING |
| test/e2e/matrix.go | report | per-case runMatrixCaseIsolated results | ✓ 31/31×2 @9f2fd75; 31/34×2 @c84709c | ✓ FLOWING |
| internal/layouts/tables.go | ENToRU/RUToEN | xkb-generated (Phase 1) | ✓ live RU-commit path @9f2fd75 | ✓ FLOWING |

### Behavioral Spot-Checks (this verifier, at HEAD 1efc6fa)

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Correction core (mixed inversion, wholesale+punctuation, refusals, register, plan, suffix, PushFeed) | `go test -race -count=1 -run 'TestConvert_CasePreserved\|TestConvertRuns_MixedInvertsPerChar\|TestConvertRuns_HomogeneousWholesale\|TestConvertRuns_Refusals\|TestDetect_PureScript\|TestBuildPlan_BothLevels\|TestMatchesSuffix\|TestBuffer_PushFeed' ./internal/correct` | ok (exit 0) | ✓ PASS |
| Actor pipeline (double-tap, mixed inversion, after-space, verify-after, flip, resets, RU digits, level 2) | `go test -race -count=1 -run 'TestActor_DoubleTapCorrects\|TestActor_MixedWordInvertsPerChar\|TestActor_AfterSpaceCorrects\|TestActor_VerifyAfterLevel1\|TestActor_FlipOnSingle\|TestActor_HardResetKeyvals\|TestActor_RUDigitsFullPipeline\|TestActor_Level2NoCaps\|TestActor_Level2WithTailAndRunes' ./internal/session` | ok (exit 0) | ✓ PASS |
| Named-test existence proof | `go test ./internal/correct -list Test` (25 names) / `./internal/session -list TestActor` (127 names) | all previously pinned tests present | ✓ PASS |
| Gates | `go build ./...` / `go vet ./...` | OK / OK | ✓ PASS |
| Purity invariant (truth 11) | grep godbus/time./go func in internal/correct non-test | 0 hits | ✓ PASS |
| v1 matrix integrity | case count + strict-schema header + re-pinned rows | 16 cases; header intact; word-punct «приветб» (:62), word-mixed «паиghbdtn» (:88) | ✓ PASS |
| Full `-race` suite / `mise run lint` / tidy | not re-run (rule: no full suite) | last full green: 261006-vqw `mise run ci` (19 pkgs) and 261007-0yg ci at their HEADs; pr-sanity green at 1efc6fa | ? SKIP (by rule; circuit evidence green) |
| Live e2e re-run | not re-run (drives owner's desktop) | see the live-evidence table — the re-pinned rows' live witness is the open human item | ? SKIP (by policy) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` convention; the phase's runnable probes remain
the mise e2e tasks (live-session; policy above) and the CI gate (pr-sanity
green at HEAD; e2e-matrix nightlies soft-skip by design). Not applicable beyond that.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ----------- | ----------- | ------ | -------- |
| CORR-01 | 02-03, 02-04 | Double-tap correction pipeline | ✓ SATISFIED | truths 14, 17; live word/phrase both directions @9f2fd75 + Oct-5 |
| CORR-04 | 02-01, 02-04 | Direction/wrong-layout correction incl. mixed | ✓ SATISFIED | truths 2, 7, 18; mixed now per-character inversion (owner-approved D-55) |
| CORR-05 | 02-01 | Register preservation | ✓ SATISFIED | truth 8; registers live @9f2fd75 + Oct-5 |
| CORR-07 | 02-01, 02-03, 02-05 | Exact-range, both ladder levels, verify-after | ✓ SATISFIED | truths 3, 9, 10, 15, 19, 20 |
| CORR-09 | 02-05 | Buffer reset triggers | ✓ SATISFIED | truths 4, 21 |
| TEST-04 | 02-02, 02-06, 02-07 | e2e matrix + CI circuit | ✓ SATISFIED | truths 5, 12, 22, 23 — infra verified; the live green of the 2 re-pinned v1 rows rides the pending human item |

Orphaned requirements: **none** (REQUIREMENTS.md re-read at HEAD: exactly these
6 IDs mapped Phase 2, all Complete; "Phase 2: 6").

### Anti-Patterns Found

Debt-marker gate re-run at HEAD 1efc6fa across all phase-covered
Go/YAML/TOML/workflow files including every file changed by the two quick
tasks (runs.go, convert.go, buffer.go, case_word.go, matrix-v1.yaml, moved
engine/layouts, actor.go, main.go): **ZERO** TBD/FIXME/XXX and zero
TODO/HACK/PLACEHOLDER.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| internal/correct/direction.go | — | `Detect` production-orphaned (re-confirmed at HEAD: zero production callers outside direction.go; classifyRune cites it; TestDetect_* corpus green) | ⚠️ Warning | Carried unchanged from both prior passes; file untouched since. Owner cleanup decision, no defect |
| test/e2e/case_m1.go | 135 | ydotool "Escape" name trap in the unexercised shell-surface fallback (deferred-items #3; line re-pinned at HEAD: `return s.pressKey(ctx, "Escape")` in the surfaceShell branch) | ⚠️ Warning | Carried unchanged; file byte-identical since prior pass |
| desktop environment | — | persistent AT-SPI shell-bridge wedge + aged-session nightlies (deferred-items #1; all four 10-03..10-07 "successes" are D48_SKIP soft-skips) | ℹ️ Info | Environmental; remedy documented in ci-runner.md; the fresh-session gate stays with the nightly pipeline/owner (WINDOWS #13 precedent) |
| .planning/.../02-VALIDATION.md | — | unfilled draft template (unchanged) | ℹ️ Info | Process artifact; not gated |

### Human Verification Required

Two items for phase-2 scope (the remaining 261006-vqw live items — matrices
v2/v3/v4 — belong to phases 3/6 scope and stay tracked there):

### 1. Live matrix v1 — re-pinned rows witness

**Test:** On the live GNOME stand (after `gsd/internalize-engine-layouts`
merges, or locally by the owner): `mise run e2e-matrix`.
**Expected:** 16/16 PASS, exit 0 — word-mixed reads «паиghbdtn» (every letter
inverted), word-punct reads «приветб»; registers, D-13 separator, D-15
digits, resets and both-direction rows unchanged green.
**Why human:** Drives the owner's live desktop (IBus + ydotool + AT-SPI);
headless tests only prove the loader/oracle plumbing.

### 2. Hand-typed mixed-text check in gedit/GTK4

**Test:** Type a Latin word, flip, type a Cyrillic word (e.g. `gfb` + привет),
double-tap Right Shift.
**Expected:** Every letter inverts («паиghbdtn»); digits/space/punctuation in
the mixed range ride unchanged; no visual jump; correction-done record in the
daemon log.
**Why human:** Real-time keyboard/IME interaction and visual result — not
reproducible programmatically.

### Gaps Summary

**No gaps.** All 22 headless-verifiable truths hold at HEAD 1efc6fa: the
owner-approved mixed-text revision (per-character inversion, spec-delta D-55
before code) is implemented, TDD-proven RED→GREEN, wired into the one
conversion entry, and pinned end-to-end from the layout tables through the
actor pipeline to the re-pinned e2e oracles; the wholesale branch and the
entire rest of the correction core survived the quick tasks and phases 5–8
with byte-identical or trivially-diffable changes; the internalization move
is a clean history-preserving git mv. Gates green (build, vet, the named
`-race` test sets, pr-sanity at HEAD). The single open item is the
live-desktop witness for the two re-pinned v1 rows (truth 5's live
dimension): no live run has executed them yet — the 2026-10-05 owner runs
predate the re-pin (their mixed rows failed exactly as the then-designed
WINDOWS-12 ledger expected, which is what prompted the owner verdict behind
the re-pin), and every nightly since is a D48_SKIP soft-skip on
origin/master@42cbe72, which does not yet contain the re-pin. That pending
witness is exactly what 261006-vqw's human-needed items track; phase-2's
slice is surfaced above. UAT 2/2 (2026-09-15) remains final; no prior human
item re-opened. No overrides carried (prior file had none).

---

_Verified: 2026-10-07T07:49:59Z_
_Verifier: Claude (gsd-verifier) — fingerprint-staleness re-verification at milestone close (mixed-text D-55 re-pin + internalization)_
