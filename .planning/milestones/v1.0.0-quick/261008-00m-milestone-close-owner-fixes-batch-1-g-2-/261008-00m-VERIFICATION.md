---
phase: 261008-00m-milestone-close-owner-fixes-batch-1-g-2-
verified: 2026-10-07T22:45:03Z
status: passed
score: 6/6 must-haves verified
covered_digest: "v1:sha256:9ea7be0ab6bf5d7009855680f76208ca76194ba74af8f1857f28e68da987eb96"
covered_files:
  - .planning/quick/261008-00m-milestone-close-owner-fixes-batch-1-g-2-/261008-00m-PLAN.md
  - .planning/quick/261008-00m-milestone-close-owner-fixes-batch-1-g-2-/261008-00m-SUMMARY.md
  - docs/SPEC.md
  - docs/adr/ADR-004-buffer-reset-triggers.md
  - docs/adr/ADR-006-two-engine-revision.md
  - internal/install/install.go
  - internal/install/install_internal_test.go
  - internal/install/selfcheck.go
  - internal/install/selfcheck_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - test/e2e/matrix.go
  - test/e2e/matrix_test.go
behavior_unverified: 0
overrides_applied: 0
---

# Quick 261008-00m: Milestone-close owner fixes batch 1 (G-2-3 + G-5-5) Verification Report

**Task Goal:** G-2-3 — own-engine flip must not reset the correction buffer (ADR-004 revision, fix in session actor lifecycle handling, unit test modeling a mixed word typed across a flip, live matrix v1 word-mixed row settles to «паиghbdtn»); G-5-5 — half-wrapped input sources auto-completed to the correct configuration with refusal only if wrapping fails; spec-delta before code per D-55 for both; plus the orchestrator-authorized follow-up: matrix stand hermeticity (plain cases pin the base config).
**Verified:** 2026-10-07T22:45:03Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| - | ----- | ------ | -------- |
| 1 | A word typed across the daemon's OWN flip corrects as a whole — «gfb»→flip→«привет» settles to «паиghbdtn» over one (-9,9) deletion (G-2-3) | ✓ VERIFIED | `internal/session/actor.go`: `flipCredit` field (L263, a.mu-only) armed in `flipTo` at L2017 ONLY after the switcher seam's SetGlobalEngine round trip returns nil; consumed in `HandleLifecycle` L788-790; `TestActor_OwnFlipMixedWordCorrectsWhole` (actor_test.go L1469) models the real flip — switcher stub via `SetSwitcher`, synthetic FocusOut/FocusIn pair — and PASSES under `-race`: exactly one RequireSurroundingText, one (-9,9) deletion, last commit «паиghbdtn», INFO done. Live confirmation (orchestrator-recorded, not re-run): matrix v1 on HEAD 4dc4877 16/16, word-mixed → «паиghbdtn» |
| 2 | A REAL focus loss still hard-resets the correction buffer (ADR-004 semantics survive) | ✓ VERIFIED | Code: the credit branch requires `kind == LifecycleFocusOut && flipCredit > 0` (L788) — a `Reset` kind never consumes; only `buf.HardReset()` is skipped (L793-795), every other side effect (FSM reset, surr clear, resolvePending, clearAfter, combo/MACR, timer) stays. Behavioral: `TestActor_OwnFlipCreditConsumedOnce` (second FocusOut after the synthetic one → empty-buffer refusal, 0 require calls), `TestActor_OwnFlipFailedSwitcherArmsNothing` (failed switcher → no credit → real reset), pre-existing `TestBuffer_ResetByFocusOut` (FocusOut and Reset with no credits) — all PASS under `-race` |
| 3 | Half-wrapped sources auto-complete to the canonical configuration at every detection point; refusal ONLY when completion fails (G-5-5) | ✓ VERIFIED | `wrapSources` (install.go L837-880): completion via `renderWrapped` + first-occurrence dedupe (L865-877); residue `errMixedSources`+hint (L859-860, deliberately ordered before nothing-wrappable when goswitch material present), unsupported-kind and nothing-wrappable `errUnsupportedPair` (L844-848, L862-863); owner verbatim in the doc comment. `resolveWrapInput` mixed-live branch routes through the wrap attempt (L902-905); `saveState` gate comment revised (L528-534) with the FULL wrap computation as the atomic gate. `checkInputSource` (selfcheck.go L196-246): heal = wrapSources → exactly ONE `gsettings set` (L218) → re-read → parse + ownership verification (L221-235) → green naming engines; refusal issued BEFORE any mutating call (L212-216). Behavioral: `TestWrapSourcesHalfWrappedCompletes` (both forms), `TestWrapSourcesRefusalTable` (3 residue forms + unsupported + nothing-wrappable; empty value, no raw line, fix hint), `TestSelfcheck_InputSourceMixedHeals` (exactly one set, asserted), `TestSelfcheck_InputSourceResidueStaysRed` (ZERO set calls, asserted, D-53 rationale) — all PASS under `-race` |
| 4 | Both revisions recorded in spec BEFORE code (D-55) | ✓ VERIFIED | Commit `0781a5f` (parent of both RED commits) touches EXACTLY docs/adr/ADR-004-buffer-reset-triggers.md, docs/adr/ADR-006-two-engine-revision.md, docs/SPEC.md — zero .go files. ADR-004: Status line "Amended 2026-10-07 — G-2-3" (L9) + dated amendment section (L81) citing the live-matrix evidence (`gfbghbdtn` readback, runes:6 source:привет), the owner decision, and the preserved real-focus-loss semantics. ADR-006: Status line (L18) + dated amendment section (L280) with the owner verbatim quote and the superseded immediate-refusal contract noted for the audit trail. SPEC §4.3 (L126-135) and §4.4 (L139-148): dated HTML-comment amendments with the old bullet preserved verbatim inside the comment and the new rule below — the house pattern |
| 5 | Every iteration ends green (mise run ci at each task boundary) | ✓ VERIFIED | Orchestrator-recorded: full `mise run ci` green at every task boundary (executor-reported + spot-verified commits). Verifier spot checks: `go build ./...` OK, `go vet` clean on both touched packages, six named test runs green under `-race` (see Behavioral Spot-Checks). Lint issues reported by the executor were fixed inside their own GREEN iterations (SUMMARY CI table) |
| 6 | FOLLOW-UP: plain matrix cases pin the base config instead of adopting the desktop owner's live config (hermeticity) | ✓ VERIFIED | `test/e2e/matrix.go`: `caseNeedsBaseEstablishment` predicate (config_base → true, no-reload-steps plain → true, reload-steps → false) + `runMatrixCase` routes every such case through `establishCaseConfigBase`/`matrixCaseBaseConfig` (commit 4dc4877). `TestMatrixCaseNeedsBaseEstablishment` pins all three rows and PASSES. Live confirmation (orchestrator-recorded): matrix v1 on HEAD 4dc4877 16/16 — the previously-FAILING word-after-space case now passes; matrix v2 20/21 with the single FAIL being the documented environmental focus-stealing class, not a product regression |

**Score:** 6/6 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `docs/adr/ADR-004-buffer-reset-triggers.md` | "Amended 2026-10-07" Status line + dated amendment section | ✓ VERIFIED | L9 + L81; substantive (mechanism, evidence, owner verdict, discriminating credit mechanism) |
| `docs/adr/ADR-006-two-engine-revision.md` | "Amended 2026-10-07" Status line + dated amendment section | ✓ VERIFIED | L18 + L280; substantive (what changes, what does not, owner verbatim, superseded-contract audit trail) |
| `docs/SPEC.md` | §4.3 + §4.4 dated deltas, old text verbatim | ✓ VERIFIED | L126-135, L139-148; house HTML-comment pattern, 7 dated mentions total |
| `internal/session/actor.go` | flipTo arms one credit on successful round trip; HandleLifecycle FocusOut consumes it, skips only the hard reset | ✓ VERIFIED | L263 (field), L2009-2017 (arm), L787-795 (consume); doc comments cite ADR-004 amendment 2026-10-07; wired — all flipTo callers serialize on a.mu (HandleKey/ExpiryAt/ToggleMode/SwitchMode lock at entry; internal callers downstream; modeSwitchChord documents "caller holds the mutex") |
| `internal/session/actor_test.go` | Own-flip mixed-word test + real-focus-loss guards | ✓ VERIFIED | `TestActor_OwnFlipMixedWordCorrectsWhole` (L1469), `TestActor_OwnFlipCreditConsumedOnce` (L1519), `TestActor_OwnFlipFailedSwitcherArmsNothing` (L1550); `mixedRU` const with pre-existing literals re-pointed (behavior-neutral, `TestActor_MixedWordInvertsPerChar` passes) |
| `internal/install/install.go` | wrapSources completion arbiter + dedupe; resolveWrapInput routing; revised saveState gate | ✓ VERIFIED | L837-880, L892-921, L528-546; wired — called from resolveWrapInput, saveState, and selfcheck's heal (one arbiter, confirmed by grep: the only non-test `errMixedSources` production use is the selfcheck residue refusal) |
| `internal/install/selfcheck.go` | checkInputSource heals the completable half-state; residue stays red | ✓ VERIFIED | L196-246; heal = one set → re-read → ownership verify → green; zero-write refusals |
| `internal/install/install_internal_test.go` | Refusal table re-pinned; completion test added | ✓ VERIFIED | `TestWrapSourcesHalfWrappedCompletes` (L140), `TestWrapSourcesRefusalTable` (L174) with all 6 rows and the superseded-contract doc comment; fixtures L48-62 |
| `internal/install/selfcheck_test.go` | Mixed test re-pinned to heal-green; residue red test added | ✓ VERIFIED | `TestSelfcheck_InputSourceMixedHeals` (L486, renamed per SUMMARY deviation 2 — no stale MixedRejected remains), `TestSelfcheck_InputSourceResidueStaysRed` (L529) |
| `test/e2e/matrix.go` + `matrix_test.go` | (Follow-up) hermeticity predicate + routing + RED-first pin | ✓ VERIFIED | `caseNeedsBaseEstablishment`/`caseHasReloadSteps`/`runMatrixCase` routing (4dc4877); `TestMatrixCaseNeedsBaseEstablishment` (matrix_test.go L616) |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | -- | --- | ------ | ------- |
| `flipTo` (successful SetGlobalEngine under a.mu) | `HandleLifecycle(LifecycleFocusOut)` | synthetic-lifecycle credit, armed before any dispatch (both sides on a.mu) | ✓ WIRED | Arm L2017 in the `err == nil` branch only; WARN failure and nil-seam degradation arm nothing; consume L788-790 exactly once; `Reset` never consumes |
| `TestActor_OwnFlipMixedWordCorrectsWhole` | live matrix v1 word-mixed row «паиghbdtn» | the unit corpus exercises the same flip the e2e row does (stub switcher + synthetic pair; no direct mixed-token feed) | ✓ WIRED | Fixture-only blind spot closed; e2e expectations untouched by the G-2-3/G-5-5 commits (`git diff e76f6b0..2fa8b53 -- test/e2e/` empty); live row settled «паиghbdtn» (orchestrator-recorded 16/16) |
| `wrapSources` → renderWrapped + dedupe | `resolveWrapInput` / saveState gate / `checkInputSource` heal | ONE completion arbiter | ✓ WIRED | All three call sites route through wrapSources; refusals return before any mutating call (selfcheck L212-216 precedes L218's set) |
| selfcheck heal | `i.call(binGSettings, "set", ...)` → re-read → "ok input-source <engines>" | single verified write | ✓ WIRED | `TestSelfcheck_InputSourceMixedHeals` asserts exactly one set with the healed args and the green verdict string; residue test asserts zero set invocations via the call log |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| `checkInputSource` heal | `healed` list | `wrapSources(parsed live gsettings line)` — parsed via `activate.ParseSourceTuples`, raw line never reaches the argument vector | Yes (unit-pinned through the Installer's call seam; the fake records the exact set args) | ✓ FLOWING |
| `wrapSources` result | rendered tuple strings | rendered only from parsed Go values (closed-enum mapping) | Yes | ✓ FLOWING |
| `flipCredit` | arm/consume counter | real `flipTo` path + real `HandleLifecycle` calls in tests | Yes (transition exercised; live row confirms on the real bus) | ✓ FLOWING |

No value in the touched paths terminates in a static return or hardcoded literal.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Own-flip mixed word corrects whole (-9,9 → «паиghbdtn») | `go test -race ./internal/session/ -run 'TestActor_OwnFlip'` | 3/3 PASS | ✓ PASS |
| Real focus loss + failed-switcher still reset; re-pointed literal test | `go test -race ./internal/session/ -run 'TestBuffer_ResetByFocusOut\|TestActor_MixedWordInvertsPerChar'` | PASS | ✓ PASS |
| Half-wrapped completion + refusal table + heal + residue-zero-write | `go test -race ./internal/install/ -run 'TestWrapSources\|TestSelfcheck_InputSource'` | 13/13 subtests PASS | ✓ PASS |
| Hermeticity predicate (plain/base/reload rows) | `go test ./test/e2e/ -run 'TestMatrixCaseNeedsBaseEstablishment'` | PASS | ✓ PASS |
| Tree compiles; touched packages vet-clean | `go build ./... && go vet ./internal/session/ ./internal/install/` | OK, exit 0 | ✓ PASS |

Full `mise run ci` was NOT re-run (orchestrator constraint; single named tests only) — boundary-green evidence is the orchestrator's recorded full runs.

### Probe Execution

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| Live matrix v1 | `mise run e2e-matrix` (orchestrator-run, HEAD 4dc4877) | 16/16 PASS — word-mixed → «паиghbdtn» | PASS (recorded; out of verifier scope by instruction) |
| Live matrix v2 | `mise run e2e-matrix` (orchestrator-run, fix tree) | 20/21 — phrase-mixed PASS; super-space-alive FAIL = documented environmental focus-stealing class (green on idle-desktop nightlies), not a product regression | PASS (recorded; single FAIL accounted) |

### Commit Structure (TDD + D-55)

| Commit | Claim | Verified |
| ------ | ----- | -------- |
| `0781a5f` | docs-only spec-delta (3 files, zero .go) | ✓ `git show --name-only` exact match |
| `77c91b7` → `4a0c887` | G-2-3 RED (test-only) → GREEN | ✓ RED commit touches only actor_test.go; `flipCredit` count in actor.go at 77c91b7 = 0 (fix absent); assertions demand (-9,9)/«паиghbdtn» |
| `e0e21ec` → `2fa8b53` | G-5-5 RED (test-only) → GREEN | ✓ RED touches only the two test files; heal/wrap code absent at e0e21ec (grep = 0) |
| `07d143b` → `4dc4877` | hermeticity RED (test-only) → GREEN | ✓ RED touches only matrix_test.go; `caseNeedsBaseEstablishment` absent at 07d143b (grep = 0) |
| all 7 | no `.planning/` artifacts committed by the executor | ✓ `.planning/` files appear only in the orchestrator's own `9cb032c` |
| e2e scope | e2e untouched by the original plan's commits | ✓ `git diff e76f6b0..2fa8b53 -- test/e2e/` empty; e2e changes come only from the authorized follow-up |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ----------- | ----------- | ------ | -------- |
| QUICK-261008-00m-milestone-close-owner-fixes | 261008-00m-PLAN.md | Both owner fixes (G-2-3, G-5-5) via spec-delta first, strict TDD, green iterations | ✓ SATISFIED | Truths 1-5 above; no orphaned requirements (quick task — the requirement is defined by the plan itself; REQUIREMENTS.md carries no additional IDs for this task) |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | No TBD/FIXME/XXX debt markers, no placeholder/stub patterns, no empty implementations in any file touched by this task | — | — |

### Human Verification Required

None. Every behavior-dependent truth has a passing behavioral test at the project's canonical seams, and the live matrix evidence (v1 16/16, v2 20/21) was recorded by the orchestrator on the fix tree. The plan's acceptance contract (unit corpus + orchestrator-run `e2e-matrix`) is fully met; `e2e-install-cycle` was optional per the plan and its absence is not a must-have gap — the install/selfcheck paths are behavior-pinned through the Installer call seam exactly as the pre-existing corpus pins them.

### Gaps Summary

None. SUMMARY.md claims were adversarially re-derived from the codebase and held at every level:

- **Exists:** all 9 PLAN artifacts + the 2 follow-up e2e files present.
- **Substantive:** flipCredit arm/consume logic, completion+dedupe, heal-with-verification, and all docs amendments read as real implementations, not markers.
- **Wired:** the credit is armed/consumed under the same mutex (serialization confirmed at every flipTo call site); wrapSources is the single arbiter behind all three detection points; no stale old-contract reference remains.
- **Behavior:** six named test runs green under `-race`, including the exact state transitions (credit consumed once; failed switcher arms nothing; exactly-one-set heal; zero-write refusal) and the hermeticity predicate rows.
- **Commit discipline:** D-55 docs-only base, three RED→GREEN pairs with fix code provably absent at each RED tree, executor commits carry zero `.planning/` files.

Informational (non-blocking, not gaps): the matrix v2 `super-space-alive` FAIL is the documented environmental focus-stealing class (green on idle-desktop nightlies); the executor noted one flaky `internal/ctlsvc` `TestSvc_ReloadInvalidKeepsLastGood` under full parallel `-race` (inotify instance exhaustion; passes in isolation) — environmental, unrelated to this task's changes.

---

_Verified: 2026-10-07T22:45:03Z_
_Verifier: Claude (gsd-verifier)_
