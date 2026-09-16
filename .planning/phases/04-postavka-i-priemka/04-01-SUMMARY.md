---
phase: 04-postavka-i-priemka
plan: 01
subsystem: install
tags: [install, ibus, systemd-user, cli, e2e, tdd]
requires:
  - engine.NewComponent/NewEngineDesc wire identity (engine/types.go)
  - ctlsvc single-instance name org.djarvur.goswitch (Phase 3)
  - e2e stand registry/snapshot-restore machinery (test/e2e)
provides:
  - internal/install package (Runner/ActiveEngines/RegistryProbe seams, Install/Uninstall, writeAtomic/writeVerified)
  - goswitchctl install / uninstall [--purge] subcommands (own 120 s budget)
  - install-cycle live e2e case + mise task e2e-install-cycle
  - path contracts: ~/.config/ibus/component/goswitch.xml, ~/.config/systemd/user/goswitchd.service, ~/.local/share/goswitch/install-state.json
affects: [04-02 selfcheck (builds on install seams), 04-03 release channels, 04-07 README install docs]
tech-stack:
  added: []
  patterns:
    - command-runner seam with env parameter (clipboard pattern extended)
    - atomic write-then-rename with byte-exact read-back verification (ASVS V14)
    - probe-verify-retry-once cache registration (Pitfall 4)
    - standalone case registry entries (stand spawns no daemon; unit daemon owns the bus)
key-files:
  created:
    - internal/install/install.go
    - internal/install/install_test.go
    - internal/install/install_internal_test.go
    - test/e2e/case_install.go
  modified:
    - cmd/goswitchctl/main.go
    - test/e2e/main.go
    - test/e2e/preflight.go
    - test/e2e/matrix.go
    - mise.toml
key-decisions:
  - "IBUS_COMPONENT_PATH REPLACES the ibus scan path — every goswitch write-cache carries user dir + /usr/share/ibus/component ':'-joined (a user-only value strips the system components from the registry cache and the next daemon start dies without its config component — live finding)"
  - "Registry verification probes the CACHE FILE before the restart (list-engine stays stale until the daemon restarts); after the async `ibus restart` the registry view is gated by a bounded-wait list-engine (registrationWait 20 s / poll 500 ms)"
  - "Uninstall restore shape-validates the saved value BEFORE gsettings set; malformed state falls back to [('xkb', 'us')] with a visible report (ASVS V5)"
  - "install/uninstall get their own installTimeout (120 s), never ctlCallTimeout; report output is paths/verdicts only (D-20/D-21)"
requirements-completed: []
duration: 47 min
completed: 2026-09-16
actuals:
  tokens: 31000
  tasks: 3
  commits: 5
plan_head_before: 357a647a9c9a011284fcbea13054217d825a82ea
commits: 5
status: complete
---

# Phase 4 Plan 01: Трассер — internal/install + goswitchctl install/uninstall + живой install-cycle Summary

**One-liner:** `goswitchctl install` puts goswitch on the desktop without root in one command — component XML, absolute-ExecStart user unit, single-owner sources takeover — with a full verified rollback in `uninstall`, proven live end to end (install → correction through the unit daemon → status → reinstall → uninstall → restored desktop).

## Accomplishments

- **internal/install package** (D-39/D-40/D-42): the full install sequence (daemon-binary preflight → verbatim sources save → component XML → env-carrying write-cache with registry verification and retry-once → absolute-ExecStart unit → daemon-reload → enable --now → ibus restart → bounded-wait list-engine + ListActiveEngines → single-owner gsettings → engine activation) and the matching rollback (idempotent stop/disable → artifact removal → env write-cache → restart → shape-validated verbatim restore with xkb fallback → state removal → optional --purge). All subprocesses ride a Runner seam with per-call deadlines; file writes are atomic with byte-exact read-back verification (ASVS V14/T-04-01-01).
- **cmd/goswitchctl**: `install` and `uninstall [--purge]` subcommands with their own installTimeout (120 s — never ctlCallTimeout) and step-by-step path/verdict reports; errUsage extended.
- **Live install-cycle e2e** (INST-01, SPEC §7 "полный цикл без человека"): a standalone registry case (the stand spawns no daemon — the installed unit daemon owns the bus) proving the FULL product lifecycle on the owner desktop: install → unit active → registry + live registration → **ghbdtn→привет correction through the installed unit daemon** (log-independent oracles) → `goswitchctl status` answers mode= → harmless reinstall → second correction → uninstall → unit inactive, artifacts gone, registry clean, sources == snapshot (machine check).
- **Research pitfalls locked as tests**: every write-cache carries IBUS_COMPONENT_PATH (Pitfall 1, corrected to user+system dirs — see deviations), cache-file verification + exactly-one retry (Pitfall 4), state file in ~/.local/share/goswitch 0600 never ~/.config/goswitch (Pitfall 8), snapshot/teardown insurance around the live case (Pitfall 9), restart visibility as a bounded wait (Pitfall 2/A6).

## TDD Gate Compliance

| Task | RED | GREEN | REFACTOR | Status |
|------|-----|-------|----------|--------|
| 1 (tracer) | ✓ d066503 (9 corpus tests fail on assertions vs the stub; RED_EVIDENCE_OK) | ✓ 2347265 | — (not needed) | Pass |
| 2 (tdd) | ✓ 9ae9ab4 (TestInstall_UnitHijackOverwritten fails on the missing read-back verdict; RED_EVIDENCE_OK) | ✓ 21862fb | — (not needed) | Pass |
| 3 (auto) | — (live case extension; not a tdd task) | ✓ 0250528 | — | Pass |

Note on Task 2's RED run: 13 of 14 corpus tests already passed — Task 1's GREEN naturally delivered the idempotency/retry/purge behaviors the plan staged into Task 2; the genuinely missing read-back verification was the intentional RED (target test failed on its assertion; gate verdict RED_EVIDENCE_OK, target_test_failed).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] IBUS_COMPONENT_PATH replaces the scan path — user-only env poisoned the registry**
- **Found during:** Task 1 (first live install-cycle run)
- **Issue:** The plan/research pattern (env = user component dir only) STRIPPED every system component (org.freedesktop.IBus.Config/Panel/Simple…) from the registry cache — the env var replaces, not extends, the scan path. The running daemon never re-read the cache, so the poison was latent until install's `ibus restart`: the fresh daemon could not find its config component ("Can not execute default config program", exit 255) and the desktop IBus went down. Assumption A6's restart chain initially looked falsified; the real cause was the cache content.
- **Fix:** every write-cache now carries `IBUS_COMPONENT_PATH=<userdir>:/usr/share/ibus/component` (the ':' delimiter is documented in ibusregistry.h); corpus pin updated. Desktop repaired via plain `ibus write-cache` + unit restart; the repaired cache was then re-poisoned once by the still-buggy uninstall before the fix landed.
- **Files modified:** internal/install/install.go, internal/install/install_test.go
- **Verification:** live install-cycle PASS twice (run with fix + full-lifecycle run); cache contents verified (system components present, goswitch present only while installed)
- **Commit:** 2347265

**2. [Rule 1 - Bug] `ibus restart` is asynchronous — the registry view needs a bounded wait**
- **Found during:** Task 1 (second live install-cycle run)
- **Issue:** The plan's sequence verifies `ibus list-engine` immediately after write-cache — but list-engine reads the daemon's startup registry, which stays stale until the restart later in the sequence; and after the (asynchronous) restart, an immediate one-shot check raced the fresh daemon's startup.
- **Fix:** immediate verification probes the cache FILE (research A3's own remedy, via a new RegistryProbe seam so the corpus can script miss/hit); the list-engine gate moved AFTER `ibus restart` as a bounded wait (20 s / 500 ms poll). The plan's A6 chain (fresh-XML → env-cache → restart → list-engine) was then proven true live.
- **Files modified:** internal/install/install.go, internal/install/install_test.go
- **Verification:** live install-cycle PASS; manual research-style probe of the exact chain confirmed
- **Commit:** 2347265

**3. [Rule 3 - Blocker] RegistryProbe seam + standalone case mechanics**
- **Found during:** Task 1 GREEN
- **Issue:** scripting cache miss/hit (Pitfall 4 pin) needs a seam the plan's option list didn't name; disabling the stand-daemon spawn needed a registry-shape change (plan delegated the location to the executor).
- **Fix:** `WithRegistryProbe` option (production = cache-file grep); `caseSpec{fn, standalone}` registry entries; setupStand skips the daemon start and preflight drops daemon-bound checks for standalone cases; matrix.go call site updated.
- **Files modified:** internal/install/install.go, internal/install/install_test.go, test/e2e/main.go, test/e2e/preflight.go, test/e2e/matrix.go
- **Verification:** mise run ci green; standalone case runs without a stand daemon
- **Commit:** 2347265

**Total deviations:** 3 auto-fixed (2 live-behavior bug fixes + 1 enabling seam/mechanics). **Impact:** the installer is now live-true where the plan's research assumptions (A3/A6/Pitfall 1 sketch) were incomplete; the desktop contract the corpus pins is the one that actually survives a live daemon restart.

## Verification Results

- `mise exec -- go test ./internal/install/ -race -count=1` — PASS (14 tests + 2 subtest groups incl. white-box TestVerifyWritten)
- `mise run ci` (build + vet + golangci-lint strict + test -race) — PASS, 0 lint issues
- `mise exec -- go test ./cmd/... ./internal/... -race -count=1` — PASS (no regressions)
- `mise run e2e-install-cycle` — PASS live (full lifecycle incl. two corrections through the unit daemon)
- `mise run e2e-ctl` — PASS live (CLI regression after the dispatcher restructure)

## Coverage

```yaml
coverage:
  - deliverable: "internal/install Install() — full D-39/D-40 sequence with env on every write-cache"
    verification: [{kind: tests, ref: "internal/install#TestInstall_Sequence", status: pass},
                   {kind: tests, ref: "internal/install#TestInstall_EveryWriteCacheCarriesEnv", status: pass},
                   {kind: command, ref: "mise run e2e-install-cycle", status: pass}]
    human_judgment: false
  - deliverable: "component XML mirrors the wire identity, empty <exec> (Pattern 2)"
    verification: [{kind: tests, ref: "internal/install#TestInstall_ComponentXMLMirrorsWireIdentity", status: pass}]
    human_judgment: false
  - deliverable: "absolute-ExecStart user unit, dist semantics preserved (ASVS V14)"
    verification: [{kind: tests, ref: "internal/install#TestInstall_UnitAbsoluteExecStart", status: pass}]
    human_judgment: false
  - deliverable: "idempotent install-over-install, retry-once, read-back verify, idempotent uninstall, purge semantics"
    verification: [{kind: tests, ref: "internal/install#TestInstall_OverInstallIdempotent", status: pass},
                   {kind: tests, ref: "internal/install#TestInstall_WriteCacheRetryOnce", status: pass},
                   {kind: tests, ref: "internal/install#TestInstall_UnitHijackOverwritten", status: pass},
                   {kind: tests, ref: "internal/install#TestVerifyWritten", status: pass},
                   {kind: tests, ref: "internal/install#TestUninstall_Idempotent", status: pass},
                   {kind: tests, ref: "internal/install#TestUninstall_PurgeRemovesUserConfig", status: pass}]
    human_judgment: false
  - deliverable: "corrupt-state safe fallback (ASVS V5)"
    verification: [{kind: tests, ref: "internal/install#TestUninstall_CorruptStateFallsBack", status: pass}]
    human_judgment: false
  - deliverable: "goswitchctl install/uninstall with own time budget"
    verification: [{kind: command, ref: "mise run e2e-install-cycle (drives the real binary)", status: pass},
                   {kind: command, ref: "mise run e2e-ctl (dispatcher regression)", status: pass}]
    human_judgment: false
  - deliverable: "live full lifecycle: install → correction via unit daemon → status → reinstall → uninstall → restored desktop"
    verification: [{kind: command, ref: "mise run e2e-install-cycle — PASS install-cycle", status: pass}]
    human_judgment: false
  - deliverable: "no-root promise — installer writes nothing outside $HOME (prohibition, verification: judgment)"
    verification: []
    human_judgment: true
    rationale: "unit corpus pins all paths under the fake $HOME and the live run left no system traces, but 'nothing outside $HOME' is the plan's flagged judgment item — verifier classifies"
```

## Requirements

INST-01 is declared by plans 04-01..04-07; per the shared-ID gate it stays unmarked until the sibling plans finish (their SUMMARY files unlock `requirements mark-complete`).

## Known Stubs

None. (T-04-01-05 mitigations — selfcheck repair and the README one-command documentation — are 04-02/04-07 scope by plan, not stubs here.)

## Issues Encountered

- The two live findings above took the desktop's IBus down twice during bring-up (both times repaired in-session via plain `ibus write-cache` + systemd unit restart; owner sources were never lost — the state file/teardown restored them). The final tree leaves the desktop healthy and the installer safe by construction.
- A stray untracked `e2e` binary at the repo root predates this plan (build artifact, untouched, uncommitted).

## Self-Check: PASSED

- Files exist: internal/install/install.go, install_test.go, install_internal_test.go, test/e2e/case_install.go — all FOUND
- Commits: d066503, 2347265, 9ae9ab4, 21862fb, 0250528 — all FOUND on gsd/phase-04-postavka-i-priemka
- Acceptance criteria re-run: all task verify blocks green (unit corpus, mise ci, cmd/internal regression, live install-cycle, live ctl regression)
- commits measured from ledger gsd-plan-head-before-04-01 (357a647): 5
