---
phase: 260930-toa-feat-single-source-input-model-selfcheck
plan: 01
subsystem: install (goswitchctl install/selfcheck) + e2e corpus + docs
tags: [installer, selfcheck, single-source, input-sources, ADR-006, D-53, e2e-oracle]
requires:
  - ADR-006 two-engine revision (d52-literal flip, accepted 05-05)
  - quick 260930-pf6 tray indicator (SNI) — the live mode indicator
  - quick 260930-nxd — switch chords untouched discipline
provides:
  - installer wraps >=1 wrappable xkb us/ru source (single 'us' → goswitch-en, single 'ru' → goswitch-ru, pair → both)
  - foreign-residue refusal (us+fr, us+ru+fr) with the D-53 rationale, atomic before any mutation
  - selfcheck input-source green on >=1 goswitch engine + zero foreign, verdict names the engines found
  - e2e install-cycle oracle >=1 restatement + count-agnostic ibus-restart precondition
  - README single-source user model + ADR-006 Amendment 2026-09-30
affects: [goswitchctl install, goswitchctl selfcheck, test/e2e install-cycle, docs]
tech-stack:
  added: []
  patterns: [shared verdict text const (foreignResidueHint) across wrap gate and audit; oracle restates the production rule, never imports it]
key-files:
  created: []
  modified:
    - internal/install/install.go
    - internal/install/install_internal_test.go
    - internal/install/install_test.go
    - internal/install/selfcheck.go
    - internal/install/selfcheck_test.go
    - test/e2e/case_install.go
    - test/e2e/case_resilience.go
    - test/e2e/case_switch.go
    - README.md
    - docs/adr/ADR-006-two-engine-revision.md
decisions:
  - "saveState's atomic gate runs the FULL wrap computation (resolveWrapInput + wrapSources), not just the resolution — Rule 1 fix, see Deviations"
  - "foreign-residue verdict text shared as the foreignResidueHint const — the wrap gate and the selfcheck audit carry one D-53 text"
  - "single-source desktops keep the skip-if-equal takeover (owner decision 2026-09-30: one source hides the dead native indicator)"
metrics:
  duration: 17 min
  completed: 2026-09-30
status: complete
actuals:
  tokens: 14942 # chars/4 over the realized diff (59,767 chars, 10 files, +500/−234)
  tasks: 4
  commits: 6 # MEASURED: git rev-list --count gsd-plan-head-before-260930-toa..HEAD
plan_head_before: 28014f7eced201b49239a44735b1c6f294b5b0c3
---

# Quick Task 260930-toa Summary: feat — single-source input model (installer + selfcheck + e2e oracle + README/ADR-006)

Installer and selfcheck accept 1..N goswitch-wrapped sources: a single xkb
'us'/'ru' desktop wraps into exactly its one goswitch engine, foreign
residue beside goswitch engines is refused atomically, selfcheck's
input-source step is green on any goswitch-only list of count ≥ 1 and
names the engines it found, and README/ADR-006 document the model with
the SNI tray icon as the mode indicator.

## Tasks Completed

| # | Task | Commits | Files |
|---|------|---------|-------|
| 1 | installer wraps ≥1 wrappable xkb source; foreign residue refused; report names the engines (TDD) | e62d055 (RED), f817216 (GREEN) | install.go, install_internal_test.go, install_test.go |
| 2 | selfcheck input-source green on ≥1 goswitch engine, named verdicts, mixed stays red (TDD) | 6ce5962 (RED), bfebd28 (GREEN) | selfcheck.go, selfcheck_test.go |
| 3 | e2e corpus: install-cycle oracle wraps ≥1 source; ibus-restart precondition count-agnostic | ec2883e | case_install.go, case_resilience.go, case_switch.go |
| 4 | README user model + ADR-006 amendment | 0e34a83 | README.md, docs/adr/ADR-006-two-engine-revision.md |

## Verification Results

- `go test -race -count=1 ./internal/install/` green after each task (RED runs watched the new expectations fail for the right reasons).
- `mise run ci` (build + vet + golangci-lint strict + `go test -race -count=1 ./...`) green after every task and at the end: 0 lint issues, all packages ok.
- `mise run tidy-diff` green — no go.mod/go.sum drift (no new deps; activate.ParseSourceTuples was already imported on both sides).
- Doc grep gates: no "clears GNOME" / "Both engines sit" in README; "Amendment 2026-09-30" present in ADR-006.
- Rename grep gate: `requireTwoSourceDesktop` gone from test/e2e.
- Transient note: one `mise run test` invocation failed once without a localized package; three consecutive full `mise run ci`/`mise run test` runs and a `go test -race -count=5 ./internal/install/` stress run were all green afterwards. Not reproduced; not attributable to this change (the install package is the only one touched on the code side and it passed 5×).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] saveState's atomic gate now validates the FULL wrap, not just the resolution**
- **Found during:** Task 1 GREEN (TestInstall_RefusalBeforeAnyWrite caught it)
- **Issue:** The plan's literal spec renames carriesXKBPair → carriesWrappableXKB ("true when the parsed list carries AT LEAST ONE xkb 'us' or 'ru' entry") and keeps saveState's gate as `resolveWrapInput(prior)` alone. With the ≥1 predicate, a us+fr desktop RESOLVES fine (it carries a wrappable entry), so the refusal moved to takeoverSources — AFTER the state file was written and write-cache/ibus-restart ran. This contradicted the plan's own must-have: "the atomic refusal gate still fails BEFORE any mutation and without creating the state file" (TestInstall_RefusalBeforeAnyWrite pins zero mutating calls and no state file for exactly the us+fr desktop).
- **Fix:** saveState gates on `resolveWrapInput(prior)` AND `wrapSources(resolved)` — the full computation. Single-source desktops pass; us+fr, half-wrapped and unwrappable desktops refuse before the backup write and before a single desktop setting changes. resolveWrapInput's resolution contract is unchanged.
- **Files modified:** internal/install/install.go (saveState + gate comment)
- **Commit:** f817216

### E2E scan result (Task 3 verify-and-leave)

The only hard both-engines sources pin in the e2e tree was
`requireTwoSourceDesktop` — renamed to `requireGoswitchDesktop` (green on
any goswitch-owned count, failure message "sources carry no goswitch
engine — install first"). The remaining "two-source" mentions are NOT
desktop pins and were deliberately left untouched:

- `switch-spike` (05-01) and `two-source-flip` (05-05) CREATE their own
  reversible two-source windows (`writeSourcesIfDiffers(twoSources)` +
  `restoreSpikeWindow`) — self-contained case setups, count-agnostic
  w.r.t. the desktop they run on; they are the historical proof record of
  the 05-01/05-05 mechanism (deliberately preserved, not stale).
- `d01-probe` likewise builds its own two-source window (historical
  experiment record).
- `INTEG-04 two-source reactivation` strings in case_resilience.go are
  case NAMING; the mechanism checks `isGoswitchEngine` (any goswitch
  engine), count-agnostic.
- The runtime flip oracles (busFlipRound journal pair + readback),
  external-flip-sync, flip-after-done, matrix v3 (including the frozen
  word-mixed/phrase-mixed rows, WINDOWS #12) — all sources-count-agnostic,
  untouched as planned.
- `installSelfcheck`'s `Contains("ok input-source")` still matches the
  named verdict ("ok input-source goswitch-en") — left as planned.

## Deferred Items (per Task 4)

1. **SPEC.md still documents the two-source takeover** — owner routes the
   specification change through the spec-delta process in the v1.1 pass
   (deliberately not edited here; ADR-006 carries the amendment and says so).
2. **Known interaction, owner-accepted:** after the desktop migrates to one
   source, a later `goswitchctl install` re-wraps from the SAVED original in
   the state file (the first backup is sacred, Pitfall 7), so a reinstall
   re-imposes the saved pair — v1.1 spec-delta candidate.

## Known Stubs

None — no stubs, skipped tests, or unrun verifications were introduced.
All `<verify>` blocks were executed; the Task 4 human-check is a live
desktop step that is ORCHESTRATOR work by constraint (block below), not a
skipped automated check.

## Post-deploy verification (ORCHESTRATOR work — live end-state on the owner's desktop, AFTER merge, with freshly built binaries)

The executor did NOT touch the live desktop or the installed unit. Steps
for the orchestrator:

1. Build fresh binaries from the merged tree (`mise run build` or
   `go build ./cmd/goswitchd ./cmd/goswitchctl`), place them where the
   installed unit/CLI resolve (the install command re-writes the unit's
   ExecStart to the binary's directory).
2. Set the sources to ONE goswitch engine:
   `gsettings set org.gnome.desktop.input-sources sources "[('ibus', 'goswitch-en')]"`
3. `systemctl --user restart goswitchd`
4. Verify:
   - `goswitchctl selfcheck` prints all six ok lines, including
     `ok input-source goswitch-en`;
   - GNOME's native input indicator is GONE from the top bar (one source);
   - the goswitch tray icon is present and reads en;
   - Super+Space flips typing (the tray icon follows to ru);
   - `goswitchctl status` reports `mode=` matching the factual engine.
5. (Optional negative probe) temporarily add a foreign source
   (`gsettings set ... "[('ibus', 'goswitch-en'), ('xkb', 'us')]"`) and
   confirm `goswitchctl selfcheck` goes RED on input-source with the D-53
   rationale — then restore the single-source list.

## Self-Check: PASSED

- All 10 planned files modified and committed (verified via
  `git log/diff 28014f7..HEAD`); working tree clean for code paths.
- All 6 commit hashes exist on `gsd/phase-05-integratsiya-s-gnome-indikatsiya`
  (e62d055, f817216, 6ce5962, bfebd28, ec2883e, 0e34a83).
- No docs artifacts (SUMMARY/STATE/PLAN) committed by the executor; ROADMAP.md untouched, per constraints.
