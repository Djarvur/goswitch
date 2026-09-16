---
phase: 04-postavka-i-priemka
plan: 02
subsystem: install
tags: [selfcheck, version-stamping, cli, ibus, systemd-user, e2e, tdd]
requires:
  - internal/install package (Runner/ActiveEngines seams, writeCache, listEngineHasGoswitch) from 04-01
  - goswitchctl install/uninstall subcommands and the install-cycle live case from 04-01
  - ctlsvc single-line key=value status canon (Phase 3, 03-06)
provides:
  - internal/install Selfcheck(ctx, io.Writer) error — the six-step D-41 live audit with per-step fix hints and the exactly-once env-cache repair; ctlStatus seam with the session-bus production probe
  - goswitchctl selfcheck subcommand (own 60 s budget) — the one-command post-install diagnosis (D-39/D-41)
  - D-37 version identity: cmd/goswitchd package vars version/commit/date (dev fallback) + -version flag + versionString(); session.Status.Version; renderStatus version= leading token — goreleaser-ready for 04-03
  - install-cycle e2e now self-checks right after install (the roadmap №1 criterion proven live)
affects: [04-03 release channels (stamps the D-37 vars), 04-07 README install docs (selfcheck step), verify-work UAT (selfcheck gate item)]
tech-stack:
  added: []
  patterns:
    - checks-table audit with fail-fast verdict lines and fix hints (e2e-preflight pattern productized)
    - probe-repair-once-then-verdict for a repairable red state (exactly one env-cache re-run, T-04-02-01)
    - ldflags-stamped package vars with an honest dev fallback (goreleaser default -X contract)
key-files:
  created:
    - internal/install/selfcheck.go
    - internal/install/selfcheck_test.go
    - cmd/goswitchd/version.go
    - cmd/goswitchd/main_test.go
  modified:
    - cmd/goswitchd/main.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - internal/ctlsvc/ctlsvc.go
    - internal/ctlsvc/ctlsvc_test.go
    - internal/install/install.go
    - cmd/goswitchctl/main.go
    - test/e2e/case_install.go
key-decisions:
  - "The wire component keeps its scheme version 0.1.0 — engine.NewComponent is NOT tied to the build stamp; the build identity surfaces only via goswitchd -version and the status version= token (research Pattern 3, plan-mandated SUMMARY record)"
  - "SetVersion-on-construction (not a NewActor signature change): the daemon pins its build into the actor once; the snapshot lifts it — NewActor's signature and every existing call site stay untouched"
  - "Config step validates ~/.config/goswitch/config.yaml (the plan-pinned path) and treats a MISSING file as green defaults — install deliberately generates no config, so no file can trip the strict decode (D-33); docs/CONFIG.md's goswitch.yaml example is a flag-passed name, the selfcheck conventional path is the planner's contract"
  - "Multi-line red causes (the YAML decoder's error) are whitespace-flattened into the one-verdict-per-line FAIL shape — the renderStatus canon applied to audit output"
  - "cmd/goswitchctl WAS modified — but for the selfcheck dispatch (Task 2), never for version parsing: parseStatusLine stayed generic and needed zero client edits for the new version= token"
requirements-completed: []
duration: 24 min
completed: 2026-09-16
actuals:
  tokens: 10200
  tasks: 2
  commits: 4
plan_head_before: 2d9ee51bb23fa689f7a8f811fe2b20c2cc96be45
commits: 4
status: complete
---

# Phase 4 Plan 02: Selfcheck D-41 + версия сборки D-37 Summary

**`goswitchctl selfcheck` — six live D-41 steps (version → component → unit → engine → config → source) each printing an ok/FAIL verdict with a fix hint and repairing the registry cache exactly once, plus the D-37 build identity (`goswitchd -version`, `version=` leading the status line) proven live inside the install cycle.**

## Performance

- **Duration:** 24 min
- **Started:** 2026-09-16T12:30:02Z
- **Completed:** 2026-09-16T12:54:23Z
- **Tasks:** 2
- **Files modified:** 13 (4 created, 9 modified)

## Accomplishments

- **D-37 version identity end to end**: package vars `version`/`commit`/`date` with the honest dev fallback (goreleaser's default `-X main.version/main.commit/main.date` — no custom ldflags needed), the `-version` flag printing `goswitchd dev` and exiting 0 before the daemon ever starts, the actor's `Status.Version` pinned at construction, and `renderStatus` leading with `version=` while `config_error` stays the closing token of the untouched key=value grammar. Release builds of 04-03 only need to stamp the vars.
- **`goswitchctl selfcheck` (D-41, INST-01)**: the six-step live audit in the preflight style — version (via the daemon's control service), component-visible (`ibus list-engine`), unit-active, engine-registered (ListActiveEngines — the deliberately different, live view per Pitfall 2), config (missing file = green defaults; invalid file = FAIL with path + flattened parse reason), input-source. Every red verdict names its fix; the first red ends the run and is the non-zero exit; output carries verdicts/paths/reasons only (D-20/D-21).
- **The repairable red state repairs itself once** (research Pitfall 1, T-04-02-01): a missing component triggers exactly one env-carrying `ibus write-cache` re-run and a re-check — never a loop (pinned by corpus and by the T-04-02-01 threat disposition).
- **install-cycle self-checks now** (roadmap criterion №1): the live e2e case runs `goswitchctl selfcheck` right after install — PASS with the six ok verdicts on the real desktop; the red side was also proven live against the uninstalled desktop (FAIL version + exit 1).

## TDD Gate Compliance

| Task | RED | GREEN | REFACTOR | Status |
|------|-----|-------|----------|--------|
| 1 (version identity) | ✓ d597a88 (8 corpus tests fail on assertions vs stubs; RED_EVIDENCE_OK) | ✓ 110b5a7 | — (not needed) | Pass |
| 2 (selfcheck) | ✓ 1b8e62f (14 corpus entries fail on assertions vs the stub audit; RED_EVIDENCE_OK) | ✓ cb32ad3 | — (not needed) | Pass |

Evidence records: `.planning/phases/04-postavka-i-priemka/red-evidence/04-02-task{1,2}-red.json` (TAP-transcribed per repo precedent, both verified RED_EVIDENCE_OK by the gate).

## Task Commits

1. **Task 1 RED: version identity corpus vs the stubs** - `d597a88` (test)
2. **Task 1 GREEN: -version + actor Version + status version= token** - `110b5a7` (feat)
3. **Task 2 RED: selfcheck D-41 corpus vs the stub audit** - `1b8e62f` (test)
4. **Task 2 GREEN: selfcheck six live steps + install-cycle step** - `cb32ad3` (feat)

**Plan metadata:** see the docs commit following this SUMMARY.

## Files Created/Modified

- `internal/install/selfcheck.go` — the D-41 audit: checks table, per-step verdicts, one-shot repair, ctlStatus seam + production session-bus probe
- `internal/install/selfcheck_test.go` — 6 top-level tests / 15 subtests: green transcript, repair-once pin, exhausted-repair red, every red hint, config no-file/valid/invalid, seam behavior
- `cmd/goswitchd/version.go` — stamped build vars (dev fallback), `versionString()`, `runVersion()`
- `cmd/goswitchd/main.go` — `-version` flag with the early exit 0 before `run()`
- `cmd/goswitchd/main_test.go` — version line corpus (dev/stamped/partial)
- `internal/session/actor.go` — `SetVersion` + `Status.Version` lifted into the snapshot
- `internal/ctlsvc/ctlsvc.go` — `renderStatus` leads with `version=`
- `cmd/goswitchctl/main.go` — `selfcheck` dispatch, `selfcheckTimeout` 60 s, errUsage extended
- `internal/install/install.go` — `ctlStatus` field + `WithCtlStatus` seam + production default
- `test/e2e/case_install.go` — selfcheck step right after the install-side verification

## Decisions Made

- Wire component stays on scheme version `0.1.0` (plan-mandated record): the IBus component metadata is not the build identity — stamping surfaces only through `-version` and status.
- Missing config = green defaults is the enforced consequence of "install generates no config": the daemon runs on built-in defaults until the first user document, and no generated file can trip the strict decode (D-33).
- `SetVersion` construction pin instead of a `NewActor` signature change — zero churn across the existing corpus.
- FAIL lines flatten multi-line causes (the YAML decoder's) into one verdict line — audit output stays line-oriented and grep-able.

## Deviations from Plan

None - plan executed exactly as written. (The GREEN-phase corpus fixes — the is-active stub's arg shape and the unpinned no-version hint — were normal RED→GREEN test iteration on the new corpus itself, not deviations from plan behavior; the plan's pinned behaviors all held.)

## Issues Encountered

None beyond the above iteration. The live install-cycle and the live red-path probe both passed on the first run after GREEN; the desktop was left healthy and uninstalled (the case's teardown verified the restore).

## Verification Results

- `mise exec -- go test ./cmd/goswitchd/ ./internal/session/ ./internal/ctlsvc/ -race -count=1` — PASS
- `mise exec -- go test ./internal/install/ -race -count=1` — PASS (15 pre-existing + 15 new/subtests)
- `mise run ci` (build + vet + golangci-lint strict + test -race) — PASS, 0 lint issues
- `mise exec -- go run ./cmd/goswitchd -version` — prints `goswitchd dev`, exit 0
- `mise run e2e-install-cycle` — PASS live (install → **selfcheck six ok** → correction → status → reinstall → correction → uninstall → desktop restored)
- Live red contract: `goswitchctl selfcheck` on the uninstalled desktop — `FAIL version: … fix: systemctl --user start goswitchd`, exit 1

## Coverage

```yaml
coverage:
  - id: D1
    description: "D-37 build identity: package vars with dev fallback, -version flag exits 0 before run(), versionString format"
    requirement: INST-01
    verification:
      - {kind: unit, ref: "cmd/goswitchd/main_test.go#TestVersionString", status: pass}
      - {kind: unit, ref: "cmd/goswitchd/main_test.go#TestVersionFlag", status: pass}
      - {kind: command, ref: "mise exec -- go run ./cmd/goswitchd -version → 'goswitchd dev', exit 0", status: pass}
    human_judgment: false
  - id: D2
    description: "Version rides the status wire: actor Status.Version pinned at construction; renderStatus leads with version=; grammar and config_error position preserved; client parser untouched"
    requirement: INST-01
    verification:
      - {kind: unit, ref: "internal/session/actor_test.go#TestStatusCarriesVersion", status: pass}
      - {kind: unit, ref: "internal/ctlsvc/ctlsvc_test.go#TestRenderStatusVersionToken", status: pass}
    human_judgment: false
  - id: D3
    description: "goswitchctl selfcheck — six D-41 steps, ok/FAIL verdicts with fix hints, exactly-one env-cache repair, fail-fast non-zero exit, no user content in output"
    requirement: INST-01
    verification:
      - {kind: unit, ref: "internal/install/selfcheck_test.go#TestSelfcheck_AllGreen", status: pass}
      - {kind: unit, ref: "internal/install/selfcheck_test.go#TestSelfcheck_ComponentRepair", status: pass}
      - {kind: unit, ref: "internal/install/selfcheck_test.go#TestSelfcheck_ComponentRedAfterRepair", status: pass}
      - {kind: unit, ref: "internal/install/selfcheck_test.go#TestSelfcheck_EachRedPath", status: pass}
      - {kind: unit, ref: "internal/install/selfcheck_test.go#TestSelfcheck_ConfigNoFileGreen", status: pass}
      - {kind: unit, ref: "internal/install/selfcheck_test.go#TestSelfcheck_StatusProbeSeam", status: pass}
      - {kind: command, ref: "live: goswitchctl selfcheck on the uninstalled desktop → FAIL version + exit 1", status: pass}
    human_judgment: false
  - id: D4
    description: "Install cycle self-checks live: goswitchctl selfcheck runs right after install and prints six ok verdicts, exit 0"
    requirement: INST-01
    verification:
      - {kind: e2e, ref: "mise run e2e-install-cycle → PASS install-cycle (six ok lines asserted in-case)", status: pass}
    human_judgment: false
  - id: D5
    description: "Plan-flagged prohibitions: selfcheck never silently greens a red state, and never prints user config content (D-20/D-21) — judgment items"
    verification: []
    human_judgment: true
    rationale: "Both are the plan's flagged verification:judgment prohibitions — the corpus pins FAIL-on-every-red-path and verdicts/paths-only output, but 'no silent green anywhere' and 'no user content' are owner-judgment statements over the whole surface; verifier classifies at UAT"
```

## Known Stubs

None. All seams have production backings; the audit is fully wired to the live desktop.

## Next Phase Readiness

- D-37 mechanics are goreleaser-ready: 04-03 needs no custom ldflags — stamp `main.version`/`main.commit`/`main.date` (goreleaser defaults) and release builds identify themselves; `go install` builds honestly report `dev`.
- INST-01's self-check criterion is proven live end-to-end; 04-07's README can document `install → selfcheck → (use) → uninstall` as the whole user journey.
- The selfcheck version step assumes a post-04-02 daemon (a pre-version reply is a named red with the restart hint) — a fresh install from the same build never hits it.

---
*Phase: 04-postavka-i-priemka*
*Completed: 2026-09-16*

## Self-Check: PASSED

- Files exist: internal/install/selfcheck.go, selfcheck_test.go, cmd/goswitchd/version.go, cmd/goswitchd/main_test.go, 04-02-SUMMARY.md — all FOUND
- Commits: d597a88, 110b5a7, 1b8e62f, cb32ad3 — all FOUND on gsd/phase-04-postavka-i-priemka
- Acceptance criteria re-run: all task verify blocks green (unit corpora, mise run ci, go run -version, live install-cycle, live red-exit contract)
- Commits measured from ledger gsd-plan-head-before-04-02 (2d9ee51): 4
