---
phase: 260930-toa-feat-single-source-input-model-selfcheck
verified: 2026-09-30T22:40:00Z
status: human_needed
score: 5/5 must-haves verified
covered_files:
  - ".planning/quick/260930-toa-feat-single-source-input-model-selfcheck/260930-toa-PLAN.md"
  - ".planning/quick/260930-toa-feat-single-source-input-model-selfcheck/260930-toa-SUMMARY.md"
  - "README.md"
  - "docs/adr/ADR-006-two-engine-revision.md"
  - "internal/install/install.go"
  - "internal/install/install_internal_test.go"
  - "internal/install/install_test.go"
  - "internal/install/selfcheck.go"
  - "internal/install/selfcheck_test.go"
  - "test/e2e/case_install.go"
  - "test/e2e/case_resilience.go"
  - "test/e2e/case_switch.go"
covered_digest: "v1:sha256:40747cdeffb7f67df300d4af1615a631176fcfe5f556f506549c46d177c1192e"
behavior_unverified: 0
overrides_applied: 0
human_verification:
  - test: "Live end-state on the owner's desktop AFTER merge, with freshly built binaries: gsettings set org.gnome.desktop.input-sources sources \"[('ibus', 'goswitch-en')]\"; systemctl --user restart goswitchd; then goswitchctl selfcheck; inspect the top bar; Super+Space; goswitchctl status."
    expected: "selfcheck prints all six ok lines including 'ok input-source goswitch-en'; GNOME's native input indicator is GONE from the top bar (one source); the goswitch tray icon is present and reads en; Super+Space flips typing (tray icon follows to ru); goswitchctl status reports mode= matching the factual engine."
    why_human: "Live GNOME session state — panel indicator visibility, tray icon rendering and real-time flip behavior are physical-desktop observations no grep or unit test can see. The executor deliberately did not touch the live desktop; this leg is orchestrator post-merge work by constraint."
re_verification:
  previous_status: none
  previous_score: n/a
  gaps_closed: []
  gaps_remaining: []
  regressions: []
---

# Quick Task 260930-toa Verification Report: feat — single-source input model (installer + selfcheck + e2e oracle + README/ADR-006)

**Task Goal:** single-source input model — selfcheck and installer accept 1..N goswitch-wrapped sources (owner decision 2026-09-30 evening); README user-model rewrite + ADR-006 amendment; live machine moves to `[('ibus','goswitch-en')]`.
**Verified:** 2026-09-30T22:40:00Z
**Status:** human_needed (all code/doc must-haves verified; the live-desktop leg is orchestrator post-merge work — marked human_needed, NOT a task failure)
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Single xkb 'us'/'ru' source wraps to exactly its goswitch engine; pair still wraps with positions preserved; already-goswitch lists keep the skip-if-equal takeover | ✓ VERIFIED | `internal/install/install.go`: `carriesWrappableXKB` (L721–733, ≥1 xkb us/ru), `minWrapped = 1` (L134; `pairTuples` gone — grep clean), `renderWrapped` maps us→goswitch-en / ru→goswitch-ru positionally (L753–773), `takeoverSources` records `wrappedEngines` BEFORE the skip-if-equal early return (L878–891). Behavioral: `TestWrapSourcesSingleSource` (single us → `goswitch-en`, single ru → `goswitch-ru`), `TestWrapSourcesPairPreserved` (both orders), `TestInstall_SequenceSingleSource` (full sequence: exactly one `gsettings set sources "[('ibus', 'goswitch-en')]"`, `ibus engine goswitch-en` kept, chords untouched), `TestInstall_TakeoverSkipsIdenticalWrite` — all pass under `-race`. |
| 2 | Refusals fire for foreign residue (us+fr, us+ru+fr), nothing wrappable, unsupported kinds, half-wrapped; atomic gate fails BEFORE any mutation and without creating the state file | ✓ VERIFIED | `wrapSources` gate order matches the plan exactly (L787–818): mixed owned+foreign → `errMixedSources`; unsupported kind → `errUnsupportedPair`; `wrapped < minWrapped` → `errUnsupportedPair` (reworded single-source text, L153–156); foreign residue (`len(parts)-wrapped-owned > 0`) → `errMixedSources` + shared `foreignResidueHint` with the D-53 rationale (L165–170). **The documented deviation is real and correct:** `saveState` gates on `resolveWrapInput(prior)` AND `wrapSources(raw)` before any write (L505–519) — with only the resolution gate, a us+fr desktop would have written the state file before `takeoverSources` refused, breaking the must-have; the plan's literal wording was internally inconsistent and the fix preserves the stronger contract. Behavioral: `TestWrapSourcesRefusalTable` (5 rows incl. already-owned-single → `errUnsupportedPair`), `TestWrapSourcesForeignResidueRefused` (us+ru+fr), and `TestInstall_RefusalBeforeAnyWrite` (us+fr desktop → refusal with fix hint, ZERO mutating calls recorded, state file never created — this passing test is the behavioral proof of the full-wrap gate deviation). |
| 3 | Selfcheck input-source green on ≥1 goswitch engine + zero foreign; verdict names the engines; zero-goswitch and goswitch-beside-foreign are distinct reds naming their fix (D-53) | ✓ VERIFIED | `internal/install/selfcheck.go` `checkInputSource` (L186–212): reads sources, parses with `activate.ParseSourceTuples` (the SAME grammar the installer wraps with — one parser both sides), splits with `countOwnedForeign`; `owned == 0` → reworded `errSourceNotOwner` ("carry no goswitch engine (goswitch-en or goswitch-ru) — fix: goswitchctl install"); `foreign > 0` → `errMixedSources` + `foreignResidueHint` (D-53); else "input-source " + joined engine names in list order — config literals only (D-20/D-21). Behavioral: `TestSelfcheck_InputSourceTwoEngines` (verdict "ok input-source goswitch-en, goswitch-ru"), `TestSelfcheck_InputSourceSingleGreen` (green, names goswitch-en, asserts goswitch-ru ABSENT), `TestSelfcheck_InputSourceForeignRejected` (red + hintInstall, "no goswitch engine"), `TestSelfcheck_InputSourceMixedRejected` (red, D-53 "no keys" rationale) — all pass. |
| 4 | e2e oracle recomputes the ≥1 wrap for single-source snapshots; ibus-restart precondition accepts any goswitch count; runtime flip oracles untouched | ✓ VERIFIED | `test/e2e/case_install.go`: `carriesOracleSource` (L139–151, restates ≥1 — never imports production), `wrapForCheck` success condition `!hasUS && !hasRU` → "no wrappable xkb us/ru source" (L183–207), both call sites updated (L220 installSourcesWrapped, L364 runUninstallAndVerify). `case_switch.go` L1216–1235: `requireGoswitchDesktop` (count-agnostic, failure "sources carry no goswitch engine — install first"), called from `pinRestartEngineRU` (case_resilience.go L142). Independent grep scan of the e2e tree: remaining "two-source" mentions are self-contained case windows (`writeSourcesIfDiffers(twoSources)` + restore in switch-spike/two-source-flip/d01-probe) and INTEG-04 case naming whose mechanism checks `isGoswitchEngine` — no remaining hard desktop pin. Confirms the SUMMARY's scan claim. |
| 5 | README documents the single-source user model (one source recommended, native indicator hidden by design, SNI tray icon is the indicator, chords untouched, no tray-click flipping); ADR-006 amendment with evidence pointers; SPEC.md untouched | ✓ VERIFIED | README "Input-source handover" (L55–65): touches ONLY the sources key, chords snapshotted verbatim, switching = goswitch gestures. "Input sources and the mode indicator" (L67–96): wraps ≥1, single us/ru mappings, "ONE goswitch source is the recommended shape… hidden — by design", tray icon (StatusNotifierItem) as the indicator, foreign-beside-goswitch refused + selfcheck red, ibus-restart semantics kept. Grep gates: zero hits for "clears GNOME" / "Both engines sit"; zero "click" mentions (no tray-click flipping documented). ADR-006: Status pointer "Amended 2026-09-30" (L11) + full "## Amendment 2026-09-30 — single-source model" section (L174–221) carrying the owner's verbatim intent, WHY obsolete, WHAT changes, WHAT IS UNAFFECTED (SetGlobalEngine, 05-04 sync, tray, XKB safety-net, v1.0.0 precedent), evidence pointers (260930-pf6-VERIFICATION.md + evening probes), SPEC.md-not-edited note. `git diff 28014f7..HEAD -- SPEC.md internal/session internal/indicator internal/ctlsvc` is empty. |

**Score:** 5/5 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/install/install.go` | wrapSources ≥1 gate + foreign-residue refusal, carriesWrappableXKB, reworded static errors, wrappedEngines for the report | ✓ VERIFIED | All present and wired (see Truths 1–2); `report()` renders "sources: wrapped (…)" (L921) |
| `internal/install/install_internal_test.go` | single-source success corpus, updated refusal table, foreign-residue test | ✓ VERIFIED | All consts/tests per plan; `wrapTripleWrapped` deleted (grep clean) |
| `internal/install/install_test.go` | named report pin, single-source install sequence, atomic refusal test kept | ✓ VERIFIED | `assertSequenceState` pins "sources: wrapped (goswitch-en, goswitch-ru)" (L323); `TestInstall_SequenceSingleSource` L335–382; `TestInstall_RefusalBeforeAnyWrite` L1155–1182 with us+fr and all assertions |
| `internal/install/selfcheck.go` + `selfcheck_test.go` | ≥1+no-foreign predicate, named-engines verdict, mixed-red and single-green tests | ✓ VERIFIED | See Truth 3; `assertSelfcheckLines` last green line updated (L145) |
| `test/e2e/case_install.go`, `case_switch.go`, `case_resilience.go` | oracle wraps ≥1 source; count-agnostic desktop precondition | ✓ VERIFIED | See Truth 4; `requireTwoSourceDesktop` gone (grep clean) |
| `README.md` + `docs/adr/ADR-006-two-engine-revision.md` | rewritten input-source handover and indicator model; amendment section + status pointer | ✓ VERIFIED | See Truth 5 |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| `resolveWrapInput` → `carriesWrappableXKB` → `wrapSources` | atomic refusal gate admits single-source desktops before any mutation | saveState calls both (install.go L510–519); takeoverSources recomputes (L870–877) | ✓ WIRED | Behaviorally proven by `TestInstall_RefusalBeforeAnyWrite` (zero mutating calls, no state file) and `TestInstall_SequenceSingleSource` (single-source passes the same gate) |
| `takeoverSources` → `Installer.wrappedEngines` → `report()` | report names the engines actually wrapped | recorded before skip-if-equal early return (L878–891), rendered L921 | ✓ WIRED | Report pins on both pair and single-source sequences assert the exact line |
| selfcheck `checkInputSource` → `activate.ParseSourceTuples` + `countOwnedForeign` | same sources grammar on both sides, never substring guessing | selfcheck.go L191–197 | ✓ WIRED | Direct import + call; matches installer's parser |
| e2e `wrapForCheck`/`carriesOracleSource` ↔ production `wrapSources` | oracle restates (never imports) the ≥1 rule | closed-enum restatement in case_install.go | ✓ WIRED | Compiles in the passing e2e package; restates, does not import |
| README/ADR-006 amendment ↔ shipped behavior | documented model matches the code | one source / tray icon / chords untouched / ≥1 wrap all cross-checked | ✓ WIRED | No contradiction found between docs and code |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| `wrapSources` | tuples | `activate.ParseSourceTuples(raw)` — the live/saved gsettings line | Yes (parsed desktop value) | ✓ FLOWING |
| `checkInputSource` verdict | names | goswitch tuples parsed from the live `gsettings get` reply | Yes | ✓ FLOWING |
| `report()` sources line | `wrappedEngines` | engine IDs parsed from the COMPUTED wrapped value | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| install package corpus (wrap/refusal/sequence/selfcheck tests) | `go test -race -count=1 ./internal/install/` | ok 1.068s | ✓ PASS |
| e2e package unit tests | `go test -race -count=1 ./test/e2e/` | ok 1.019s | ✓ PASS |
| Full CI gate suite (build + vet + golangci-lint strict + `go test -race -count=1 ./...` + tidy-diff) | `mise run ci` | all 13 packages ok, 0 lint issues, no go.mod drift | ✓ PASS |
| Must-have tests exist | `go test -list '.*' ./internal/install/` | all 9 named tests listed | ✓ PASS |
| Stale-wording grep gates | `grep -c "clears GNOME\|Both engines sit" README.md` | 0 | ✓ PASS |

### Probe Execution

No `scripts/*/tests/probe-*.sh` probes declared or conventional for this repo; `mise run ci` is the project's probe surface and was executed in a fresh process (see above).

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| QUICK-260930-toa-single-source-input-model | 260930-toa-PLAN.md | installer/selfcheck/e2e accept 1..N goswitch-wrapped sources; docs amended | ✓ SATISFIED | Truths 1–5; not tracked in `.planning/REQUIREMENTS.md` (quick-task-local ID — no orphaned-requirement issue: REQUIREMENTS.md maps nothing else to this quick task) |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none) | — | TBD/FIXME/XXX/HACK/PLACEHOLDER scan clean across all 8 code files; no empty implementations, no hardcoded-empty props | — | — |

### Human Verification Required

The single remaining leg is the **live-desktop end-state — ORCHESTRATOR post-merge work** (the executor deliberately did not touch the live desktop or the installed unit). Full steps are documented in the SUMMARY's "Post-deploy verification" section and the plan's Task 4 human-check.

### 1. Live end-state on the owner's desktop (after merge, freshly built binaries)

**Test:** `gsettings set org.gnome.desktop.input-sources sources "[('ibus', 'goswitch-en')]"` → `systemctl --user restart goswitchd` → run `goswitchctl selfcheck`; inspect the GNOME top bar; press Super+Space; run `goswitchctl status`.
**Expected:** selfcheck prints all six ok lines including `ok input-source goswitch-en`; the native GNOME input indicator is GONE (one source); the goswitch tray icon is present and reads en; Super+Space flips typing and the tray icon follows to ru; `goswitchctl status` reports `mode=` matching the factual engine.
**Why human:** Panel-indicator visibility, tray rendering and real-time flip behavior are physical-desktop observations — no grep, unit test or headless run can see them.

### Gaps Summary

None. All five must-have truths are verified against the codebase with passing behavioral tests; the documented Rule-1 deviation (saveState gating the FULL wrap computation) was independently confirmed as both real and necessary — it is the only reading under which the plan's own atomic-refusal must-have and its pinning test can coexist. The SUMMARY's claims checked out in every particular (diff stat +500/−234 over 10 files, 6 commits, e2e scan, deferred notes, untouched SPEC/session/indicator/ctlsvc). The task goal is achieved in the codebase; `human_needed` reflects only the by-design orchestrator leg (live desktop migration + observation).

---

_Verified: 2026-09-30T22:40:00Z_
_Verifier: Claude (gsd-verifier)_

## Orchestrator live leg (2026-09-30, post-merge)

Executed on the owner's desktop with binaries built from 0e34a83:
- sources set to `[('ibus', 'goswitch-en')]`, unit restarted;
- `goswitchctl selfcheck` — six ok lines, `ok input-source goswitch-en`;
- SNI item registered and sticking (t+1s…t+10s after restart; an initial
  absence right after the FIRST restart proved to be a re-registration race
  of the same bus name in the watcher, cleared by the restart — not a code
  regression; v0 has no re-register-on-watcher-loss by design);
- IconPixmap flips en→ru→en deterministically on the real bus (md5
  cce7bd2d24 ↔ 5701469352); `goswitchctl status` mode=en matches the
  factual engine; `current` stays 0 with nothing fighting it.

Remaining human check: the owner confirms the native GNOME input indicator
is GONE from the top bar (single source) while the goswitch tray icon and
Super+Space flipping keep working.

Orchestrator note (post-verification observation): the daemon journal
carried two `switch_engine: context deadline exceeded` WARNs tonight (17:56,
22:19) AFTER the 150 ms budget was live — occasional genuinely-slow
SetGlobalEngine round trips under load, not the old 40 ms lie (which is
fixed and tested). Recorded in the project memory for UAT watch; no action
in this task.
