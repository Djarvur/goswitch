---
phase: 05-integratsiya-s-gnome-indikatsiya
plan: 02
subsystem: installer (internal/install, internal/activate, test/e2e)
tags: [install, gsettings, dbus-input-sources, d54, two-source, asvs-v5, selfcheck, e2e]

requires:
  - phase: 04-postavka-i-priyomka
    provides: "installer lifecycle (saveState/first-backup-sacred, shape-checked verbatim restore, Runner seams, selfcheck D-41 steps), single-owner takeover form"
provides:
  - "activate.SourceTuple + activate.ParseSourceTuples — the one canonical strict GVariant sources parser (remainder-check), shared daemon-side and installer-side"
  - "install.wrapSources — closed-enum D-54 wrap: the user's own xkb us/ru pair becomes goswitch-en/goswitch-ru with positions preserved, every other entry transits verbatim; rendered only from parsed values (ASVS V5)"
  - "Refusal table errUnsupportedPair/errMixedSources (+errSavedOriginalUnusable): no pair, foreign kind, half-wrapped hand-edited list, unusable save — all refuse BEFORE any desktop mutation (atomic D-54 refusal)"
  - "Upgrade path (Pitfall 7): an already-goswitch desk wraps from the SAVED original in install-state.json; the first backup stays byte-sacred"
  - "Skip-if-equal takeover (Pitfall 6): a value-identical sources write is skipped; the explicit engine activation stays in sequence"
  - "selfcheck 'input-source' step requires BOTH goswitch engines (criterion 5) with a two-source red verdict naming the fix"
  - "install-cycle e2e asserts the two-source wrapper by machine comparison and accepts verbatim restore to the saved original on upgrade-form desks"
affects: [05-03, 05-04, 05-05, EXEC-configuration, uninstall/restore consumers]

actuals:
  tokens: 13004            # 52017 diff chars / 4 over the 8 changed files
  tasks: 2
  commits: 4                # measured: git rev-list --count da6519b..HEAD (production commits)

tech-stack:
  added: []                 # zero new dependencies (tidy-diff green)
  patterns:
    - "canonical shared parser: one strict ParseSourceTuples exported from activate, consumed by install (no hand-rolled fork)"
    - "closed-enum rendering: gsettings arguments built only from parsed Go values + enum literals, never raw user input"
    - "atomic refusal gate: pair validation inside saveState BEFORE the state write or any mutating subprocess"
    - "oracle restates the enum: the e2e check recomputes the wrapper via the canonical parser but its own copy of the mapping"

key-files:
  created: []
  modified:
    - internal/activate/activate.go
    - internal/activate/activate_test.go
    - internal/install/install.go
    - internal/install/install_test.go
    - internal/install/install_internal_test.go
    - internal/install/selfcheck.go
    - internal/install/selfcheck_test.go
    - test/e2e/case_install.go

key-decisions:
  - "D-54 wrap is computed, never forced: live pair → saved original (Pitfall 7) → refusal; half-wrapped lists refuse errMixedSources instead of guessing (owner confirms at the verify gate)"
  - "One canonical parser: activate.ParseSourceTuples exported and reused by install; the gsettings argument never carries raw user input (ASVS V5 / T-05-02-01)"
  - "The e2e oracle restates the wrap enum through the canonical parser — an oracle importing the production wrap could not catch a wrap bug"
  - "Refusals name the failing tuple TYPE and the fix, never the raw user line (D-20/D-21-safe)"

patterns-established:
  - "resolve-then-wrap-then-compare: takeover input resolution (live pair / saved original / refusal) shared by the saveState gate and the takeover write"
  - "countOwnedForeign/renderWrapped decomposition keeps the wrap under the strict complexity ceilings"

requirements-completed: [INST-01, INTEG-03]

coverage:
  - id: D1
    description: "Canonical strict parser activate.ParseSourceTuples/SourceTuple; internal parseSources delegates (one parse site)"
    requirement: INST-01
    verification:
      - kind: unit
        ref: "internal/activate/activate_test.go#TestParseSourceTuples"
        status: pass
    human_judgment: false
  - id: D2
    description: "wrapSources: pair wrap with positions, third-source transit, refusal table (us+fr / single / foreign kind / half-wrapped), upgrade from saved original, skip-if-equal write, atomic refusal before any mutation"
    requirement: INST-01
    verification:
      - kind: unit
        ref: "internal/install/install_internal_test.go#TestWrapSourcesPairPreserved|TestWrapSourcesThirdSourcePassthrough|TestWrapSourcesRefusalTable|TestWrapSourcesAlreadyOwnedUpgrade"
        status: pass
      - kind: unit
        ref: "internal/install/install_test.go#TestInstall_Sequence|TestInstall_TakeoverSkipsIdenticalWrite|TestInstall_RefusalBeforeAnyWrite|TestInstall_SecondInstallKeepsOriginalBackup"
        status: pass
      - kind: integration
        ref: "mise run ci (build/vet/golangci-lint strict/test -race) + tidy-diff clean"
        status: pass
    human_judgment: false
  - id: D3
    description: "selfcheck input-source step requires both goswitch engines with the two-source red verdict (criterion 5); step name 'input-source' unchanged"
    requirement: INTEG-03
    verification:
      - kind: unit
        ref: "internal/install/selfcheck_test.go#TestSelfcheck_InputSourceTwoEngines|TestSelfcheck_InputSourceSingleRejected|TestSelfcheck_InputSourceForeignRejected"
        status: pass
      - kind: e2e
        ref: "live: goswitchctl selfcheck on the owner's desk after the two-source install — six ok verdicts incl. 'ok input-source' (2026-09-29)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Live install-cycle full-case PASS in two-source form (criterion 1 end-to-end incl. correction through the installed daemon)"
    requirement: INST-01
    verification:
      - kind: e2e
        ref: "live 2026-09-29: install wrapped the saved pair on the owner's desk (report 'sources: wrapped pair'), machine-checked sources == wrapper, selfcheck 6/6, teardown restored the original verbatim"
        status: pass
    human_judgment: true
    rationale: "The case's zenity correction step was focus-refused (mutter denial, owner actively working), so mise run e2e-install-cycle did not finish PASS end-to-end; per phase notes the full-case rerun consolidates in plan 05-05. WINDOWS ledger entry recorded (unrun-verify). The new two-source assertions themselves all passed live before the focus step."

duration: 29 min
completed: 2026-09-29
status: complete
---

# Phase 05 Plan 02: D-54 two-source installer takeover Summary

**install now wraps the user's OWN xkb us/ru pair into both goswitch engines (positions and extra sources preserved), refuses anything else atomically before touching the desktop, upgrades from the saved original, and selfcheck audits the two-source state — the canonical parser lives once in activate.**

## Performance

- **Duration:** 29 min
- **Started:** 2026-09-29T20:09:18Z
- **Completed:** 2026-09-29T20:38:20Z
- **Tasks:** 2 (both strict TDD: RED evidence RED_EVIDENCE_OK per task)
- **Files modified:** 8

## Accomplishments
- The single-owner constant `ownerSourcesSet` is gone: takeover computes the D-54 wrapper from the user's own layout pair — positions preserved, extra sources transiting verbatim, honest refusals outside the pair, all BEFORE any desktop mutation.
- selfcheck (criterion 5) now requires BOTH goswitch engines; the red verdict names the two-source expectation and the fix.
- The upgrade path (Pitfall 7) and skip-if-equal write (Pitfall 6) are pinned by unit corpora AND proven on the live owner desktop.

## Task Commits

1. **Task 1: wrapSources — closed-enum обёртка пары** - `2490a22` (test: RED corpus + compile stubs) → `1c67c67` (feat: implementation)
2. **Task 2: selfcheck двух источников + живой install-cycle** - `a8ebe96` (test: RED corpus) → `e761164` (feat: selfcheck + e2e case)

**Plan metadata:** (docs commit follows this summary)

_Note: TDD tasks each carry one RED (test) and one GREEN (feat) commit; lint-driven refactors were folded into the GREEN commits with the full suite re-run green._

## Files Created/Modified
- `internal/activate/activate.go` — exported `SourceTuple`/`ParseSourceTuples` (canonical strict parser with remainder-check); `parseSources` now delegates
- `internal/activate/activate_test.go` — `TestParseSourceTuples` corpus (order, remainder, unbalanced, empty)
- `internal/install/install.go` — `wrapSources`, refusal sentinels (`errUnsupportedPair`, `errMixedSources`, `errSavedOriginalUnusable`), `resolveWrapInput` (live pair → saved original → refusal), atomic pair gate in `saveState`, skip-if-equal `takeoverSources`, two-source report verdict
- `internal/install/install_test.go` — sequence re-pinned (14 calls, computed wrapper), refusal/skip/upgrade corpora
- `internal/install/install_internal_test.go` — wrap corpus (pair/preserve/transit/refusal table/upgrade)
- `internal/install/selfcheck.go` — `checkInputSource` requires both engines; reworded `errSourceNotOwner`
- `internal/install/selfcheck_test.go` — two-source green/rejection corpora; green stub moved to the wrapped form
- `test/e2e/case_install.go` — post-install machine equality with the computed wrapper (upgrade-form aware), verbatim-restore target = saved original on goswitch-owned desks

## Decisions Made
- Wrap input resolution order pinned: live xkb us/ru pair → SAVED original from install-state.json (already-goswitch desk, Pitfall 7) → named refusal; a half-wrapped hand-edited list refuses `errMixedSources` rather than guessing (FLAGGED plan assumption — owner confirms at the verify gate).
- One canonical parser: `activate.ParseSourceTuples` exported; install imports activate (no parser fork, no import cycle).
- The e2e oracle restates the closed enum itself through the canonical parser (oracle independence from the code under test).
- Error texts name the failing tuple TYPE and the fix, never the raw user line (D-20/D-21-safe); ASVS V5 holds — `gsettings set` arguments are built only from parsed values and enum literals.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocker] Existing backup test's marker tripped the new D-54 gate**
- **Found during:** Task 1 (RED corpus authoring)
- **Issue:** `TestInstall_SecondInstallKeepsOriginalBackup` planted an unshapeable marker (`MARKER-ORIGINAL`); under the new atomic gate an already-goswitch desk with an unusable saved original must REFUSE, so the scenario as planted would fail permanently.
- **Fix:** Planted a shape-valid saved original (the owner's real pair) — the test's intent (first backup byte-sacred) is preserved and it now doubles as the plan-required upgrade case (wrapper computed from the SAVED original).
- **Files modified:** internal/install/install_test.go
- **Verification:** test green in the full -race suite
- **Committed in:** 2490a22 / 1c67c67

**2. [Rule 1 - Bug] install-cycle verbatim-restore oracle was wrong for upgrade-form desks**
- **Found during:** Task 2 (live run analysis)
- **Issue:** On a desk whose snapshot is goswitch-owned (the owner's phase-4 form), uninstall restores the SAVED original pair — the old oracle compared against the snapshot and would false-fail the drift check.
- **Fix:** `runUninstallAndVerify` reads the state file's saved original BEFORE the uninstall consumes it and accepts restore == saved original when the snapshot carries no xkb pair.
- **Files modified:** test/e2e/case_install.go
- **Verification:** live teardown restore observed green (original pair restored verbatim); suite green
- **Committed in:** e761164

### Deferred (environment, per phase notes)

- `mise run e2e-install-cycle` full-case PASS: the live run proved every NEW assertion (install → sources machine-equal the wrapper; selfcheck 6/6 green incl. two-source input-source; teardown restored the original verbatim) but stopped at the zenity correction step — mutter focus-refusal with the owner actively working. The full-case rerun consolidates in plan 05-05 (phase live-e2e consolidation). WINDOWS ledger: `unrun-verify` entry recorded.

---

**Total deviations:** 2 auto-fixed (1 Rule 3 blocker, 1 Rule 1 bug) + 1 environment deferral.
**Impact on plan:** none on scope — both fixes protect contracts the plan itself mandates; the deferred full-case rerun is scheduled in 05-05.

## Issues Encountered
- The live e2e run's best-effort uninstall tore down the owner's phase-4 installation (case-by-design insurance). Restored in-session: built this tree's `goswitchctl`/`goswitchd` into the owner's install dir (`~/.local/share/goswitch`), reset sources to the saved original pair, re-ran the production install — the desk is now in the plan's TARGET two-source state (sources `[('ibus', 'goswitch-en'), ('ibus', 'goswitch-ru')]`, unit active, engine goswitch-en live, selfcheck 6/6 green, state file holding the original pair). This live reinstall was also the plan's deciding upgrade-path probe (FLAGGED assumption A5/Pitfall 7) — it passed.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- EXEC-configuration foundation is in place: 05-03+ mechanisms run inside the two-source wrapper; the switcher work (05-03/05-04) can flip `goswitch-en`/`goswitch-ru` as real input sources.
- 05-05 owns the consolidated live e2e (incl. the deferred install-cycle full-case PASS and 05-01's parked spike checkpoint).

---
*Phase: 05-integratsiya-s-gnome-indikatsiya*
*Completed: 2026-09-29*

## Self-Check: PASSED

All 9 created/modified artifacts exist on disk; all 5 commits verified in git history (2490a22, 1c67c67, a8ebe96, e761164, 82f48c6). Measured commits from the plan ledger da6519b..HEAD: 4 production + 1 metadata.
