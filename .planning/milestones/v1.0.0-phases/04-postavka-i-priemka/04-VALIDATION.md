---
phase: "4"
slug: "postavka-i-priemka"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-15"
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Audit filled 2026-09-16 (state B — reconstructed from the seven SUMMARYs, plans and the tree):
> every plan-pinned test exists in the tree, the full `-race` suite plus `mise run ci` are green at
> HEAD a888f05 (re-run by this audit), and the live evidence chain (e2e-install-cycle, e2e-perf ×3,
> matrix v3 31/31 local + double-run on runner green106, published release v1.0.0) is recorded in the
> SUMMARYs and independently spot-checked (PROXY-OK re-verified live).

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib, -race per D-08) |
| **Config file** | mise.toml (D-09 — mise is the only scenario interface; goreleaser 2.18.1 pinned under [tools]) |
| **Quick run command** | `mise exec -- go test ./... -race -count=1` |
| **Full suite command** | `mise run ci` |
| **Live acceptance** | `mise run e2e-install-cycle`, `mise run e2e-perf`, `mise run e2e-matrix-v3` (31 cases), `mise run e2e-ctl` |

## Sampling Rate

- **After every task commit:** Run `mise exec -- go test ./... -race -count=1`
- **After every plan wave:** Run `mise run ci`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 04-01 install/uninstall + live install-cycle | 04-01 | 1 | INST-01 | T-04-01-01..05 | atomic write + read-back verify (V14), env on EVERY write-cache (Падение 1), idempotent over-install/uninstall, corrupt-state xkb fallback (V5), D-20/D-21 output | unit (fakeRunner corpus) + live e2e | `go test -race ./internal/install/ -count=1` ; `mise run e2e-install-cycle` ; `mise run e2e-ctl` | ✅ TestInstall_Sequence, TestInstall_EveryWriteCacheCarriesEnv, TestInstall_ComponentXMLMirrorsWireIdentity, TestInstall_UnitAbsoluteExecStart, TestInstall_DaemonBinaryMissing, TestInstall_StateFileOutsideConfigDir, TestInstall_SecondInstallKeepsOriginalBackup, TestInstall_OverInstallIdempotent, TestInstall_WriteCacheRetryOnce, TestInstall_UnitHijackOverwritten, TestVerifyWritten, TestUninstall_FullRollback, TestUninstall_Idempotent, TestUninstall_CorruptStateFallsBack, TestUninstall_PurgeRemovesUserConfig; live full lifecycle PASS 2026-09-16 (×2 during bring-up + re-run at tag gates 04-07, selfcheck step added by 04-02) | ✅ green |
| 04-02 selfcheck D-41 + version D-37 | 04-02 | 2 | INST-01 | T-04-02-01..03 | repair-exactly-once then red, fail-fast hints per red path, missing config = green defaults, -version exits 0 pre-run, version= leads status canon | unit + live e2e | `go test -race ./internal/install/ ./cmd/goswitchd/ ./internal/session/ ./internal/ctlsvc/ -run 'Selfcheck\|Version\|CarriesVersion\|RenderStatusVersionToken' -count=1` ; `mise run e2e-install-cycle` | ✅ TestSelfcheck_AllGreen, TestSelfcheck_ComponentRepair, TestSelfcheck_ComponentRedAfterRepair, TestSelfcheck_EachRedPath, TestSelfcheck_ConfigNoFileGreen, TestSelfcheck_StatusProbeSeam, TestVersionString, TestVersionFlag, TestStatusCarriesVersion, TestRenderStatusVersionToken; + gap-fill (this audit): cmd/goswitchctl#TestParseStatusLine_ExtractsVersionToken, TestParseStatusLine_StampedVersionToken, TestStatusJSON_VersionTokenSurvives; live: selfcheck six ok inside install-cycle; live red probe on uninstalled desktop (FAIL + exit 1) recorded | ✅ green |
| 04-03 release pipeline (D-38) | 04-03 | 3 | INST-01 | T-04-03-01..04 | checksums.txt integrity, contents:write scoped to release.yml only (SCOPE-OK), goreleaser mise-pinned, tidy-diff clean | config + grep gates; snapshot proofs superseded live by v1.0.0 | `mise exec -- goreleaser check` ; `mise exec -- go mod tidy && git diff --exit-code -- go.mod go.sum` ; greps RELEASE-WF-OK / SCOPE-OK; live: release run 35130309174 + CHECKSUM-OK / STAMPED-OK | ✅ (config gates green per SUMMARY; no Go corpus — config plan) | ✅ green |
| 04-04 perf acceptance INST-03 | 04-04 | 2 | INST-03 | T-04-04-01..03 | budget = gate not metric (exit≠0 both boundaries), percentile formula pin, /proc VmHWM parser, prod-form daemon (no -debug), no field content in report | unit (math) + live e2e | `go test -race ./test/e2e/ -run 'Percentile\|ProcStatus\|BudgetGate' -count=1` ; `mise run e2e-perf` | ✅ TestPercentile_Exact, TestPercentile_SingleSample, TestReadProcStatus, TestPerfBudgetGate; live: 3 full 40/40 runs 2026-09-16 — memory half PASSES (VmHWM 12364–12368 kB < 51200), latency half gate FAILS as designed (p95 204.4/190.2/184.4 ms ≥ 50, exit≠0, reproducible) — **verdict WAIVED by owner decision (в) 2026-09-16, WINDOWS #5 (accept-as-is; daemon reaction 0.2–1.8 ms, window dominated by ydotool 0.1.8 injector; README publishes the honest budget-fail status)** | ✅ green (infra) / waived (latency verdict) |
| 04-05 matrix v3 (criterion №2) | 04-05 | 1 | INST-01 (nearest — no dedicated REQ-ID; TEST-04 closed Phase 2/3) | T-04-05-01..03 | v2 frozen (V2-FROZEN), superset by names (V2-SUBSET-OK) + ≥10 new cases (COUNT-OK), closed surface vocabulary, spike-pinned expectations | unit (decoder, phase-3 corpus) + live e2e | `go test -race ./test/e2e/ -run 'Matrix' -count=1` ; V2-SUBSET-OK / COUNT-OK / V2-FROZEN gates ; `mise run e2e-matrix-v3` | ✅ TestMatrixDecode, TestMatrixDecode_RejectsNew, TestMatrixKeyNames_Canonical; live 31/31 PASS exit 0 locally 2026-09-16 (+ clean re-run at 04-07 tag gates after the executor tree-switch incident was voided by a clean repeat) | ✅ green |
| 04-06 D-48 double-run | 04-06 | 4 | INST-01 (nearest — criterion №2) | T-04-06-01..03 | no interference between run #1 and #2, fresh_session + loginctl preflight (30-min threshold, loud fail), watchdog 0 = default (regression from 04-04 fixed RED→GREEN) | unit + live CI dispatch | `go test -race ./test/e2e/ -run Watchdog -count=1` ; grep DOUBLE-RUN-WIRED ; `gh run view 35108412175` | ✅ TestWatchdogLimit_ZeroMeansDefault, TestWatchdogLimit_ExplicitOverrides; live run 35108412175 conclusion success — both sequential v3 runs `matrix: 31/31 PASS`, 0 FAIL rows | ✅ green |
| 04-07 README + ACCEPTANCE + ledger + release | 04-07 | 5 | INST-01 | T-04-07-01..04 | no bare `ibus write-cache` documented (NO-BARE-CACHE), EN==RU command/count sync, ledger open_count 0 (LEDGER-CLEAN), release integrity (assets/checksums/stamp/proxy) | grep gates + live public-artifact checks | README-COMPLETE / NO-BARE-CACHE / sync / CHECKLIST-OK / LEDGER-CLEAN greps ; `gh release view v1.0.0` + `sha256sum -c checksums.txt` + proxy `.info` | ✅ RELEASE-ASSETS-OK / STAMPED-OK (`goswitchd 1.0.0` + sha == tag commit 2bc8c6a) / CHECKSUM-OK / PROXY-OK recorded; PROXY-OK independently re-verified by this audit (`{"Version":"v1.0.0","Hash":"2bc8c6ade5a0…"}`) | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

## Wave 0 Requirements

- [x] Existing infrastructure covers all phase requirements (go test + mise + test/e2e stand from Phases 1–3; internal/install and the perf/watchdog corpora arrived with the phase's own TDD tasks — no separate Wave 0 was required).

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| D-49 acceptance checklist execution (install from README → selfcheck → live gestures → config reload → status counters → uninstall) | INST-01 | The owner UAT session IS the manual gate per SPEC §7.2 — the checklist is the artifact; passing it is the next phase step (verify-work), by design not automatable | Owner executes docs/ACCEPTANCE.md (6 items, expected/result/note + Summary) in one live session; criteria 03-UAT format |
| D-48 formal fresh-session double run (`fresh_session=true`) | INST-01 (criterion №2) | Requires owner GDM relogin; a loginctl workflow shell step is not unit-testable; the double-run MECHANICS are proven live (run 35108412175, both 31/31) on the current session — mechanics vs gate deliberately separated by plan 04-06 | docs/ci-runner.md «D-48 double-run gate»: owner relogin → dispatch e2e-matrix with fresh_session=true → success criterion: both v3 runs green in ONE launch |

*All other phase behaviors have automated verification (unit corpora + the four live mise scenarios).*

## Validation Audit 2026-09-16

| Metric | Count |
|--------|-------|
| Gaps found | 1 |
| Resolved | 1 |
| Escalated | 0 |

Notes:
- **Gap resolved (this audit):** the client half of the D-37 wire contract had zero tests — `parseStatusLine`/`statusJSON` in cmd/goswitchctl carried the `version=` token unpinned (plan 04-02 made `TestParseStatusLineParsesVersion` conditional: «если корпус клиента рядом — иначе grep-гейт»; no client corpus existed, so «client parser needs no edits» was reasoning, not a falsifiable test). Filled with `cmd/goswitchctl/main_test.go` (3 behavioral tests: generic extraction with version= leading and the config_error-tail grammar; stamped `1.0.0` extraction per the 04-03/04-07 goreleaser no-'v' pin; --json typing contract). Green; `mise run ci` green after 1 lint iteration (lll line length in the new file).
- **Plan-vs-tree name reconciliation:** 04-01's SUMMARY coverage list omits four plan-declared corpus tests (TestInstall_DaemonBinaryMissing, TestInstall_StateFileOutsideConfigDir, TestInstall_SecondInstallKeepsOriginalBackup, TestUninstall_FullRollback) — all four verified present and green in the tree by this audit. Same-behavior naming deviations: none beyond that.
- **04-04 percentile pin:** the plan's pin (p95 == p99 == 100 ms at n=10) is arithmetically inconsistent with the plan's own formula `s[int(float64(len(s)-1)*p)]` (yields 90); the test pins the formula's actual output (p50=50, p95=p99=90) and the formula is preserved verbatim (SUMMARY Deviation 1) — honest fix, no weakening.
- **INST-03 disposition (recorded, not flagged):** memory half CLOSED by measurement (VmHWM ~12.4 MB < 50 MB, three runs); latency half honestly FAILED the prescribed ydotool→AT-SPI window (p95 184–205 ms across three full runs, gate exit≠0 each time) and is **WAIVED** by owner decision (в) 2026-09-16 — WINDOWS ledger #5, open_count 0; README publishes the real numbers with the budget-fail status in both languages. REQUIREMENTS.md intentionally leaves INST-03 unchecked pending phase verify-work (04-07-SUMMARY decision, MACR-01 Phase-3 precedent) — the disposition is recorded, this is not a missing test.
- **Release evidence supersedes config proofs:** 04-03's snapshot/dry-run proofs were superseded by the actual v1.0.0 publication (tag → run 35130309174 success → assets + checksums + stamp == tag commit); the STAMPED-OK grep adaptation (`goswitchd 1.0.0` + sha equality, goreleaser strips the 'v') is a documented stricter equivalent, not a weaker check (04-07-SUMMARY Deviation 1).
- Judgment-flagged prohibitions (no writes outside $HOME; no user content in install/selfcheck/perf output; no silent green; v2 frozen; spike-pinned expectations; no interference between double runs; EN==RU doc sync) are pinned by the corpus/gates to the extent testable; their whole-surface judgment residue belongs to the phase verifier at UAT, as flagged in each plan.

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** validated 2026-09-16 (audit reconstructed from SUMMARYs + tree at HEAD a888f05; full -race suite and `mise run ci` green by this audit; gap test added and green; live evidence chain — install-cycle, e2e-perf, matrix v3 local + double-run on green106, release v1.0.0 — recorded and spot-checked)
