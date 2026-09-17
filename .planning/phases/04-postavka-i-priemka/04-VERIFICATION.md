---
phase: 04-postavka-i-priemka
verified: 2026-09-16T18:34:39Z
status: human_needed
score: 9/11 must-haves verified
covered_files:
  - .github/workflows/e2e-matrix.yml
  - .github/workflows/release.yml
  - .goreleaser.yaml
  - .planning/REQUIREMENTS.md
  - .planning/WINDOWS.md
  - .planning/phases/04-postavka-i-priemka/04-01-PLAN.md
  - .planning/phases/04-postavka-i-priemka/04-01-SUMMARY.md
  - .planning/phases/04-postavka-i-priemka/04-02-PLAN.md
  - .planning/phases/04-postavka-i-priemka/04-02-SUMMARY.md
  - .planning/phases/04-postavka-i-priemka/04-03-PLAN.md
  - .planning/phases/04-postavka-i-priemka/04-03-SUMMARY.md
  - .planning/phases/04-postavka-i-priemka/04-04-PLAN.md
  - .planning/phases/04-postavka-i-priemka/04-04-SUMMARY.md
  - .planning/phases/04-postavka-i-priemka/04-05-PLAN.md
  - .planning/phases/04-postavka-i-priemka/04-05-SUMMARY.md
  - .planning/phases/04-postavka-i-priemka/04-06-PLAN.md
  - .planning/phases/04-postavka-i-priemka/04-06-SUMMARY.md
  - .planning/phases/04-postavka-i-priemka/04-07-PLAN.md
  - .planning/phases/04-postavka-i-priemka/04-07-SUMMARY.md
  - README.md
  - README.ru.md
  - cmd/goswitchctl/main.go
  - cmd/goswitchd/main.go
  - cmd/goswitchd/main_test.go
  - cmd/goswitchd/version.go
  - docs/ACCEPTANCE.md
  - docs/ci-runner.md
  - internal/ctlsvc/ctlsvc.go
  - internal/ctlsvc/ctlsvc_test.go
  - internal/install/install.go
  - internal/install/install_internal_test.go
  - internal/install/install_test.go
  - internal/install/selfcheck.go
  - internal/install/selfcheck_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - test/e2e/case_install.go
  - test/e2e/cases/matrix-v3.yaml
  - test/e2e/focus_helper.py
  - test/e2e/main.go
  - test/e2e/matrix.go
  - test/e2e/perf.go
  - test/e2e/perf_test.go
  - test/e2e/preflight.go
  - test/e2e/surface.go
  - test/e2e/watchdog_test.go
covered_digest: "v1:sha256:1680b28d20f8dd5b7c1cfc7d9275a02126799fa2c499ca46407181a923a7d66a"
behavior_unverified: 0
overrides_applied: 1
overrides:
  - must_have: "Производительность: p95 реакции на горячую клавишу < 50 мс (INST-03, latency half)"
    reason: "Owner waiver decision (в) 2026-09-16, WINDOWS #5 (closed): the prescribed ydotool→AT-SPI window is dominated by ydotool 0.1.8 injector overhead (~80–125 ms/event); daemon reaction is 0.2–1.8 ms. v1.0.0 ships the honest as-measured p50/p95/p99 with the budget-fail status published in both READMEs."
    accepted_by: "owner (Daniel Podolsky), recorded by goswitch phase execution"
    accepted_at: "2026-09-16"
human_verification:
  - test: "D-49 owner manual acceptance: execute docs/ACCEPTANCE.md (6 items + Summary) in one live session — install from README on the live desktop (release-archive channel, no root) → goswitchctl selfcheck all green → live gesture set (flip/word/phrase/selection/combo/mixed/reload) → status counters grow → uninstall → desktop fully restored"
    expected: "Every checklist item lands its expected result; Summary block filled; any miss is a release-blocking finding (fix + re-tag v1.0.1 per the documented release cycle)"
    why_human: "The owner UAT session IS the phase's designed acceptance gate (goal: «владелец принял его вручную», SPEC §7.2); requires a live GNOME desktop, real keyboard gestures and the owner's judgment — not machine-checkable"
  - test: "D-48 formal fresh-session run: owner relogin → dispatch e2e-matrix with fresh_session=true per docs/ci-runner.md «D-48 double-run gate» → both sequential v3 runs green in ONE launch"
    expected: "loginctl preflight passes (session age ≤ 30 min), then both mise run e2e-matrix-v3 steps report matrix: 31/31 PASS, job conclusion success"
    why_human: "Requires an owner GDM relogin to produce a genuinely untouched session; the double-run MECHANICS are already proven live (run 35108412175: 31/31 twice back-to-back, conclusion success) — the fresh-session gate is by design an acceptance-time step (plan 04-06 mechanics-vs-gate split)"
---

# Phase 4: Поставка и приёмка — Verification Report

**Phase Goal:** Продукт ставится без root по документированной инструкции, укладывается в бюджет производительности, полная e2e-матрица зелёная дважды подряд на нетронутой сессии, и владелец принял его вручную — релиз v1.
**Verified:** 2026-09-16T18:34:39Z
**Status:** human_needed
**Re-verification:** No — initial verification

**Mode note:** ROADMAP marks this phase `mode: mvp`, but the goal is a delivery/acceptance goal, not a user-story («As a…, I want to…, so that…» — `user-story.validate` returns false). Consistent with the phases 1–3 precedent (all `mode: mvp`, all verified goal-backward), this report applies the goal-backward methodology with the ROADMAP success criteria as the contract; the UAT-script framing (MVP user-flow first) is preserved inside the human-verification section, which leads with the owner's hands-on acceptance.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Root-less install in one command (SC-1, D-39/D-40): component XML under `$HOME`, systemd user unit, env-cache registration, unit active, sources takeover, activation | ✓ VERIFIED | `gsd_run verify.artifacts` pass on all 7 plans; internal/install corpus green (TestInstall_Sequence, TestInstall_EveryWriteCacheCarriesEnv, TestInstall_ComponentXMLMirrorsWireIdentity, …); live install-cycle PASS ×2 during bring-up + re-run at 04-07 tag gates (SUMMARYs; VALIDATION audit spot-check); key links CLI→install and IBUS_COMPONENT_PATH-on-every-write-cache verified in source |
| 2 | selfcheck D-41: six steps, ok/FAIL + fix hint, exit≠0 on red, exactly-once env-cache repair | ✓ VERIFIED | TestSelfcheck_AllGreen/ComponentRepair/ComponentRedAfterRepair/EachRedPath/ConfigNoFileGreen green in the -race run; live selfcheck green inside install-cycle; live red probe on uninstalled desktop (FAIL + exit 1) recorded in 04-02-SUMMARY |
| 3 | Uninstall full rollback (SC-1, D-42): unit stop/disable, artifacts removed, cache re-written with env, sources restored with readback, state file removed; `--purge` opt-in | ✓ VERIFIED | TestUninstall_FullRollback, TestUninstall_Idempotent, TestUninstall_CorruptStateFallsBack green; live install-cycle teardown (verifyRestored-style) PASS; state-file save/restore pair present at internal/install/install.go:402/617. Advisory CR-01 (below) does not contradict the must-have as written — the plan's own contract prescribes the `[('xkb','us')]` fallback-with-report for a missing state file; the review proposes a stricter no-op contract (owner fix decision) |
| 4 | Release channels: release tar.gz + checksums, `go install` channel, version stamping vs honest dev (SC-1, D-37/D-38) | ✓ VERIFIED | **Re-verified live this session:** GitHub Release v1.0.0 published 2026-09-16T17:47:32Z, assets `goswitch_1.0.0_linux_amd64.tar.gz` + `checksums.txt`, not draft; proxy `v1.0.0.info` → `{"Version":"v1.0.0","Hash":"2bc8c6ade5a0…"}` == tag commit (git rev-parse v1.0.0 → 2bc8c6a); release.yml triggers on `tags: ['v*']` with `contents: write` scoped to this workflow only; `go run ./cmd/goswitchd -version` → `goswitchd dev`, exit 0 (honest unstamped fallback); goreleaser absent from go.mod (0 hits) |
| 5 | Matrix v3 full breadth with capability-tier coverage (SC-2, D-47): v2 frozen superset + gedit (GTK3) + chromium-x11 (XWayland window mode) | ✓ VERIFIED | matrix-v2.yaml untouched by phase 4 (last commits c92996c/21df37c, both phase 3); matrix-v3.yaml = 31 cases (21 v2 verbatim + 6 gedit + 4 chromium-x11, counts verified in file); closed-surface-vocabulary validation + gedit preflight in tree; local run 31/31 PASS exit 0 recorded (04-05-SUMMARY + re-run at tag gates); spike-pinned expectations in the v3 header and driver comments |
| 6 | D-48 double-run mechanics: v3 runs TWICE back-to-back in one job, any fall = job failure, no interference between runs | ✓ VERIFIED | e2e-matrix.yml contains two sequential `mise run e2e-matrix-v3` steps (lines 138/154); fresh_session input + loginctl preflight with `D48_MAX_SESSION_AGE_SEC: 1800` wired; live proof run 35108412175: run#1 `matrix: 31/31 PASS` (14:27:45Z) → run#2 `matrix: 31/31 PASS` (14:29:57Z), conclusion success, 0 FAIL rows; TestWatchdogLimit_ZeroMeansDefault green (the 04-04 watchdog regression found live by run 35107406144 is fixed structurally, WINDOWS #6) |
| 7 | D-48 formal gate on a genuinely untouched session (SC-2 «на нетронутой сессии») | → HUMAN (by design) | Acceptance-time step per plan 04-06's mechanics-vs-gate split and adjudication #2: procedure documented in docs/ci-runner.md; requires owner relogin — see Human Verification item 2. NOT a missing automated check |
| 8 | Performance measured: daemon memory < 50 MB (SC-3, INST-03 memory half) | ✓ VERIFIED | VmHWM 12364–12368 kB (< 51200 kB) across three full 40/40 live runs 2026-09-16; budget gate is unit-pinned (TestPerfBudgetGate) and fired exit≠0 on the latency boundary in every run (never silent); /proc parser unit-pinned (TestReadProcStatus); daemon spawned prod-form (no -debug) |
| 9 | Performance measured: p95 hotkey reaction < 50 ms (SC-3, INST-03 latency half) | ✓ PASSED (override) | Measured honestly: p95 184.4 ms (p50 166.9 / p99 186.5) in the prescribed ydotool→AT-SPI window across 3 full runs — gate FAIL by design, reproducible. **Override applied:** owner decision (в) 2026-09-16, WINDOWS #5 closed — the window is dominated by ydotool 0.1.8 injector overhead (~80–125 ms/event), daemon reaction 0.2–1.8 ms; README publishes the real numbers with the explicit budget-fail status in both languages (verified below, truth 11) |
| 10 | Owner manual acceptance passed (SC-4, D-49, SPEC §7.2) | → HUMAN (the designed gate) | docs/ACCEPTANCE.md checklist exists (6 items expected/result/note + Summary + Gaps); the corrupted `ГHBDTN` example found by UI review was fixed (commit 6e90c34, verified: line 50 now `GHBDTN`); execution is the verify-work UAT itself — see Human Verification item 1 |
| 11 | README + install instructions published (SC-4, D-50/D-46): bilingual pair, one-command install, perf table from the live run | ✓ VERIFIED | README.md (EN, primary) + README.ru.md (RU) present, structurally mirrored; `goswitchctl install` appears 4× in each (one-command install, no manual ibus/systemctl steps); zero bare `ibus write-cache` documented anywhere in README/docs (grep clean — Pitfall 1 prohibition holds); perf table matches perf-report.txt exactly (p50 166.9 / p95 184.4 / p99 186.5 ms, peak RSS 12.4 MB) with the honest «**exceeds** / **превышает**» budget paragraph in both languages; known copy residue: `## Troubleshooting` untranslated in README.ru.md:157 (advisory, below) |

**Score:** 9/11 truths verified (8 VERIFIED + 1 PASSED override; 2 routed to human verification by phase design; 0 present-behavior-unverified)

**Verification-method notes:**
- Full workspace suite run once this session at HEAD 8c18606: `go build ./...` + `go test -race -count=1 ./...` — 14/14 packages ok, 0 failures.
- Live-desktop runs (install-cycle, e2e-perf ×3, matrix v3 local + CI double-run, release run) were NOT re-run — recorded evidence in the SUMMARYs and the independent VALIDATION audit stands; cheap re-checks (release publication, proxy, tag/commit identity, greps, unit suite) were re-executed fresh by this verifier.
- `verify.artifacts`: all 7 plans — valid (exists + substantive). `verify.key-links`: 5/7 plans all-verified; 04-01 and 04-07 each had one pseudo-component link the tool cannot resolve («install save (install-state.json)», «README.md/README.ru.md», «тег v1.0.0») — all three verified manually above (truths 3, 4, 11).

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/install/install.go` | Install/Uninstall/writeAtomic, Runner seam, XML+unit render | ✓ VERIFIED | wired: imported by cmd/goswitchctl, exercised by e2e case_install |
| `internal/install/install_test.go` + `install_internal_test.go` | fakeRunner corpus | ✓ VERIFIED | green in -race run |
| `internal/install/selfcheck.go` + `selfcheck_test.go` | six-step D-41 audit + corpus | ✓ VERIFIED | green; repair-once pin |
| `cmd/goswitchctl/main.go` | install/uninstall/selfcheck subcommands, own budgets | ✓ VERIFIED | thin-client dispatch verified |
| `cmd/goswitchd/version.go` + `main.go` + `main_test.go` | version vars, -version flag | ✓ VERIFIED | live `goswitchd dev` exit 0 |
| `internal/session/actor.go`, `internal/ctlsvc/ctlsvc.go` | Version in status snapshot, `version=` token | ✓ VERIFIED | TestStatusCarriesVersion, TestRenderStatusVersionToken green |
| `test/e2e/perf.go` + `perf_test.go` + `focus_helper.py` | percentile, /proc parser, resident witness, budget gate | ✓ VERIFIED | unit math green; live ×3 recorded |
| `test/e2e/surface.go`, `matrix.go`, `preflight.go`, `cases/matrix-v3.yaml` | gedit + chromium-x11 drivers, closed vocabulary, v3 corpus | ✓ VERIFIED | 31/31 live PASS recorded |
| `test/e2e/case_install.go`, `watchdog_test.go`, `main.go` | install-cycle case; watchdog default fix | ✓ VERIFIED | TestWatchdogLimit_* green |
| `.github/workflows/e2e-matrix.yml` | double run + fresh_session preflight | ✓ VERIFIED | steps verified in file |
| `.github/workflows/release.yml` | tag v* → goreleaser release, contents:write scoped | ✓ VERIFIED | proven live by run 35130309174 |
| `.goreleaser.yaml`, `mise.toml` | two builds, tar.gz, checksums; goreleaser 2.18.1 pinned | ✓ VERIFIED | config gates green per SUMMARY + live release |
| `README.md`, `README.ru.md`, `docs/ACCEPTANCE.md`, `docs/ci-runner.md` | bilingual pair, UAT checklist, D-48 procedure | ✓ VERIFIED | content checks above |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| cmd/goswitchctl/main.go | internal/install/install.go | thin-client dispatch | ✓ WIRED | tool-verified |
| internal/install/install.go | ibus write-cache env | IBUS_COMPONENT_PATH every call | ✓ WIRED | tool-verified |
| install save | uninstall restore | install-state.json (0600, `$HOME/.local/share/goswitch/`) | ✓ WIRED | manual: const at install.go:61, save :402, restore :617 |
| cmd/goswitchctl/main.go | internal/install/selfcheck.go | .Selfcheck( dispatch | ✓ WIRED | tool-verified |
| internal/ctlsvc/ctlsvc.go | internal/session/actor.go | `version=` token from snapshot | ✓ WIRED | tool-verified |
| internal/install/selfcheck.go | engine live-probe | ListActiveEngines | ✓ WIRED | tool-verified |
| .goreleaser.yaml | cmd/goswitchd/main.go | default ldflags → main.version vars | ✓ WIRED | tool-verified; stamped 1.0.0 in the published binary |
| .github/workflows/release.yml | mise.toml | mise install from [tools] pin | ✓ WIRED | tool-verified |
| test/e2e/perf.go | focus_helper.py | resident witness line protocol | ✓ WIRED | tool-verified |
| test/e2e/perf.go | /proc/<pid>/status | VmHWM oracle by daemon pid | ✓ WIRED | tool-verified |
| test/e2e/cases/matrix-v3.yaml | test/e2e/matrix.go | closed surface vocabulary | ✓ WIRED | tool-verified |
| test/e2e/surface.go | test/e2e/matrix.go | matrixSurfaceApp re-pin discipline | ✓ WIRED | tool-verified |
| .github/workflows/e2e-matrix.yml | mise.toml | two × `mise run e2e-matrix-v3` | ✓ WIRED | manual: lines 138/154 |
| README.md | perf-report.txt | D-46 live numbers | ✓ WIRED | tool-verified; numbers match byte-for-byte |
| README pair | goswitchctl install | one-command instruction | ✓ WIRED | manual: 4 mentions each EN/RU |
| тег v1.0.0 | release.yml | tag triggers publication | ✓ WIRED | manual: tag → 2bc8c6a; release.yml `tags: ['v*']`; release run success |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Full workspace green iteration (directive 2) | `go build ./... && go test -race -count=1 ./...` | 14/14 packages ok | ✓ PASS |
| Version flag contract (D-37) | `go run ./cmd/goswitchd -version` | `goswitchd dev`, exit 0 | ✓ PASS |
| Matrix corpora load + decode | covered by suite (TestMatrixDecode, TestMatrixDecode_RejectsNew) | ok | ✓ PASS |
| Budget gate is a gate, not a metric | covered by suite (TestPerfBudgetGate) + live perf-report.txt `budget: FAIL` with exit≠0 ×3 | ok | ✓ PASS |
| Release publication | `gh release view v1.0.0` + proxy `.info` curl | published, 2 assets; proxy Hash == tag commit | ✓ PASS |
| Live-desktop behaviors (install-cycle, perf window, matrix v3 ×2, selfcheck on live desktop) | recorded runs 35108412175 / 35130309174, perf-report.txt, SUMMARYs | recorded | ✓ EVIDENCED (not re-run) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` probes exist (the project's probe surface is mise tasks). Phase-declared live scenarios are covered by the recorded-evidence row above; `mise run ci` green at HEAD a888f05 per the VALIDATION audit, and build+test re-run green by this verifier at HEAD 8c18606.

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
| ----------- | ----------- | ----------- | ------ | -------- |
| INST-01 | 04-01, 04-02, 04-03, 04-05*, 04-06*, 04-07 (*nearest-coverage scope notes) | Установка без root: systemd user unit, регистрация engine в IBus, `go install` + бинарник из GitHub releases | ✓ SATISFIED (REQUIREMENTS.md: Complete) | Truths 1–6, 11: install/uninstall/selfcheck proven live + unit-pinned; release v1.0.0 published and re-verified live |
| INST-03 | 04-04 | Реакция на горячую клавишу < 50 мс; память < 50 МБ | ✓ SATISFIED WITH WAIVER (memory verified; latency waived by owner, WINDOWS #5 closed) | Truths 8–9. REQUIREMENTS.md intentionally leaves the checkbox unchecked pending the owner UAT flip (04-07-SUMMARY decision, MACR-01 Phase-3 precedent) — flipping is the owner's acceptance-time act, not a missing implementation |

No orphaned requirements: REQUIREMENTS.md maps exactly {INST-01, INST-03} to Phase 4; both are claimed by plans and accounted for above. All other requirement rows in REQUIREMENTS.md predate this phase.

### Anti-Patterns Found

Debt-marker scan of all 25 phase-4 source/config/doc files: **zero** TBD/FIXME/XXX/HACK/PLACEHOLDER markers. No stub patterns (no `return null`/empty-report implementations; every empty-state output is a documented contract, e.g. `config (defaults: no config file)`).

Known review findings (recorded, owner fix decision pending — none invalidates a must-have as written):

| File | Finding | Severity | Impact on verdicts |
| ---- | ------- | -------- | ------------------ |
| internal/install/install.go:616-636 | CR-01: uninstall without install state force-sets sources to `[('xkb','us')]` (review argues for a true no-op) | Critical (review scale) | Advisory. The 04-01 must-have's own wording prescribes exactly this fallback-with-report for an absent state file, and its prohibition (never leave the desk without input) is honored; the review's stricter no-op contract is an improvement decision for the owner, routed to the fix queue with the UI-review copy finding (misleading "unreadable or malformed" verdict line) |
| internal/install/install.go:322 | WR-01: uninstall swallows non-gone unit-step errors | Warning | Advisory — error-propagation hardening, no must-have contradiction |
| internal/session/actor.go:1054 | WR-02: per-app MACR a11y D-Bus call under actor mutex in the keystroke path | Warning | Advisory (phase-3 code touched by 04-02 only for the Version field); input-plane availability hardening for the owner's queue |
| test/e2e/surface.go:517 | WR-03: recovery poke without char filter on x11 surface | Warning | Advisory — stand reliability, not product |
| test/e2e/case_m1.go:254 | WR-04: concurrent `Cmd.Wait` in zenity teardown | Warning | Advisory — latent stand race |
| internal/install/install.go:692 | IN-02: component XML version "0.1.0" ≠ release stamp | Info | Documented plan decision (04-02 assumption: wire scheme version ≠ build identity) — decided deviation, not a defect |
| IN-01, IN-03..06 | doc comments, dup derivation, positional-arg leniency, floating lint pin, non-exec daemon preflight | Info | Advisory backlog |
| README.ru.md:157 | `## Troubleshooting` untranslated (D-50 mirror miss at heading level; commands/numbers/versions do sync) | Warning (UI review) | Advisory — facts-sync prohibition not violated; copy fix queued with UI-review findings (priority-1 `ГHBDTN` already fixed in 6e90c34) |

### Human Verification Required

1. **Owner manual acceptance — D-49 (the phase's designed gate)**
   **Test:** Execute /home/nil/DiskD/W/Djarvur/goswitch/docs/ACCEPTANCE.md top to bottom in one live session (install from README → selfcheck → gestures → config reload → status → uninstall).
   **Expected:** All six items land; Summary filled; the desktop is fully restored after uninstall.
   **Why human:** The goal text itself demands «владелец принял его вручную»; requires a live desktop and the owner's judgment.
2. **D-48 formal fresh-session double run**
   **Test:** Owner relogin → dispatch e2e-matrix with `fresh_session=true` per docs/ci-runner.md.
   **Expected:** Preflight passes; both v3 runs report 31/31 PASS in one launch.
   **Why human:** Requires a genuine GDM relogin; mechanics already proven live (run 35108412175, 31/31 twice, success).
3. **Judgment-tier prohibition residue** (per plan flags): whole-surface confirmation that install/uninstall/selfcheck/perf write nothing outside `$HOME`, print no user content, and never go silently green — spot-checkable during items 1–2 (the streaming install output, the FAIL lines and the budget-fail exit make silent green structurally unlikely).

### Gaps Summary

None. No truth FAILED, no artifact missing/stub, no key link unwired, no blocker anti-pattern. The two open items are the phase's own designed human gates (owner UAT per D-49; formal fresh-session D-48 run), not missing automated work — both have their procedures documented (docs/ACCEPTANCE.md, docs/ci-runner.md) and their mechanics proven live.

Recorded notes:
- The latency half of INST-03 is not a gap: measured, gate-enforced, waived by owner decision (в) with WINDOWS #5 closed and the honest numbers published.
- deferred-items.md records a historical flake of the phase-3 `combo-word-layout` case (post-reload tap after closeZenity; unusual focus state; candidate remedy noted). Not a phase-4 truth; no later phase exists to absorb it — the owner may fold the remedy into post-release maintenance.
- All WINDOWS entries are closed (open_count 0): #1 fixed, #2 fixed, #3 waived (junk), #4 waived (GTK4 platform limitation), #5 waived (owner decision в), #6 fixed.

---

_Verified: 2026-09-16T18:34:39Z_
_Verifier: Claude (gsd-verifier)_
