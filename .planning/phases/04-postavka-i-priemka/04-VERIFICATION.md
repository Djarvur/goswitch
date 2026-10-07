---
phase: 04-postavka-i-priemka
verified: 2026-10-07T23:08:38Z
status: passed
score: 13/13 must-haves verified
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
  - .planning/phases/04-postavka-i-priemka/04-08-PLAN.md
  - .planning/phases/04-postavka-i-priemka/04-08-SUMMARY.md
  - .planning/phases/04-postavka-i-priemka/04-09-PLAN.md
  - .planning/phases/04-postavka-i-priemka/04-09-SUMMARY.md
  - .planning/phases/04-postavka-i-priemka/red-evidence/04-08-task1-red.json
  - README.md
  - README.en.md
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
  - scripts/d48-nightly-dispatch.sh
  - test/e2e/case_install.go
  - test/e2e/cases/matrix-v3.yaml
  - test/e2e/focus_helper.py
  - test/e2e/main.go
  - test/e2e/matrix.go
  - test/e2e/perf.go
  - test/e2e/perf_test.go
  - test/e2e/preflight.go
  - test/e2e/quiesce_test.go
  - test/e2e/surface.go
  - test/e2e/watchdog_test.go
covered_digest: "v1:sha256:13ff04f69498d40a1d8d8aa2a4e059ff7758c89b0ac905ee7f3f4b89eb5f2bdb"
behavior_unverified: 0
overrides_applied: 1
overrides:
  - must_have: "Производительность: p95 реакции на горячую клавишу < 50 мс (INST-03, latency half)"
    reason: "Owner waiver decision (в) 2026-09-16, WINDOWS #5 (closed): the prescribed ydotool→AT-SPI window is dominated by ydotool 0.1.8 injector overhead (~80–125 ms/event); daemon reaction is 0.2–1.8 ms. v1.0.0 ships the honest as-measured p50/p95/p99 with the budget-fail status published in both READMEs. Re-verified 2026-10-07: the honest numbers and the «exceeds/превышает» budget paragraph are still published in README.md:397 and README.en.md:387."
    accepted_by: "owner (Daniel Podolsky), recorded by goswitch phase execution"
    accepted_at: "2026-09-16"
re_verification:
  previous_status: passed
  previous_score: 13/13
  gaps_closed: [] # staleness refresh — the previous pass had no gaps; this round re-validates the fingerprint after quick 261008-00m changed tree code covered by this phase
  gaps_remaining: []
  regressions: []
advisory:
  - finding: "docs/ACCEPTANCE.md result fields remain unfilled (frontmatter status: pending, Summary pending: 6) even though the owner's D-49 verdict is recorded in 04-UAT.md (pass, 2026-09-16) and the 2026-10-07 acceptance-by-evidence is recorded in planning history; the checklist document itself was never back-filled"
    category: other
    reason: "ACCEPTANCE.md is by design the owner's hands-on fill-in document; the verdicts exist elsewhere in committed planning artifacts, so this is transcription bookkeeping, not a missing or unwired artifact. Resolves if the owner (or the closing UAT session) transcribes the verdicts into the checklist."
    evidence_status: "none provided"
coincidental_reliance_items: []
---

# Phase 4: Поставка и приёмка — Verification Report (Re-verification at Milestone Close)

**Phase Goal:** Продукт ставится без root по документированной инструкции, укладывается в бюджет производительности, полная e2e-матрица зелёная дважды подряд на нетронутой сессии, и владелец принял его вручную — релиз v1.
**Verified:** 2026-10-07T23:08:38Z
**Status:** passed
**Re-verification:** Yes — second fingerprint-staleness refresh at milestone close. The 2026-10-07T20:59:03Z pass was invalidated because quick task 261008-00m landed AFTER it, changing phase-4-covered tree code: internal/install/* (wrapSources completion+dedupe, resolveWrapInput routing, saveState gate, checkInputSource self-heal — G-5-5, RED e0e21ec → GREEN 2fa8b53; spec-delta 0781a5f), plus internal/session/actor.go+test (G-2-3 flipCredit) and test/e2e/matrix.go+test (stand hermeticity follow-up).

## Goal Achievement

All must-have paths re-derived against the CURRENT tree. The staleness window touched exactly seven covered files (git log --since=2026-10-07T20:59:03Z): internal/install/{install.go, install_internal_test.go, selfcheck.go, selfcheck_test.go}, internal/session/{actor.go, actor_test.go}, test/e2e/matrix.go — all already in the covered list; no new impl files, so the covered_files derivation is unchanged and only the digest was recomputed. Every other phase-4 deliverable (goreleaser, release workflow, e2e-matrix double-run, dispatch script, README pair, ci-runner docs) has an empty git log for the window — the previous pass's line-level verification stands. **Behavioral contract change absorbed:** per the owner's verbatim directive («автоматически приводить конфигурацию к правильной, а если не получилось - отказ», spec-delta first per D-55, ADR-006 amendment + SPEC §4.3/§4.4), half-wrapped input sources (one goswitch tuple + one wrappable foreign) now AUTO-COMPLETE to the canonical configuration, with refusal only when completion fails — the 2026-09-30 immediate-`errMixedSources` contract is superseded; the residue (foreign layout outside us/ru), unsupported-kind and nothing-wrappable refusals keep their error identities.

### Observable Truths

| # | Truth | Status | Evidence |
| --- | ----- | ------ | -------- |
| 1 | Root-less install in one command (SC-1, D-39/D-40): component XML under `$HOME`, systemd user unit, env-cache registration with IBUS_COMPONENT_PATH on every write, unit active, sources takeover, activation | ✓ VERIFIED | Fresh behavioral evidence this session: TestInstall_Sequence + TestInstall_SequenceSingleSource PASS under `-race` (2026-10-07T23:08Z run); the takeover gate now proves the resolved sources COMPLETE into a goswitch-owned list before the backup (saveState install.go:535-549, full wrap computation as the atomic gate — a completable half-state passes and the takeover writes the completed list, residue refuses before any desktop mutation); install-state machinery intact (writeCache, componentPathEnv); live install-cycle ×2 recorded (SUMMARYs, VALIDATION audit) |
| 2 | selfcheck D-41: six steps, ok/FAIL + fix hint, exit≠0 on red, exactly-once env-cache repair | ✓ VERIFIED | TestSelfcheck_AllGreen PASS under `-race` this session; checkInputSource (selfcheck.go:196-246) now heals the completable half-state — wrapSources (one arbiter) → exactly ONE `gsettings set` (:218) → re-read → parse+ownership verification → green engine-naming verdict; refusal issued BEFORE any mutating call; D-41 step structure otherwise intact |
| 3 | Uninstall full rollback (SC-1, D-42): unit stop/disable, artifacts removed, cache re-written with env, sources restored with readback, state file; `--purge` opt-in | ✓ VERIFIED | TestUninstall_FullRollback + TestUninstall_Idempotent PASS under `-race` this session; the G-5-5 commit did not touch the restore chain (saveState backup discipline preserved — first backup sacred, Pitfall 7; resolveWrapInput:892-921 routes the saved-original path unchanged for the already-owned live desktop) |
| 4 | Release channels: release tar.gz + checksums, `go install` channel, version stamping vs honest dev (SC-1, D-37/D-38) | ✓ VERIFIED | Recorded live evidence stands (release published 2026-09-16T17:47:32Z, assets + checksums.txt, proxy Hash == tag commit; `goswitchd dev` honest fallback); release surfaces untouched in the staleness window (git log empty); `go build ./...` green this session |
| 5 | Matrix v3 full breadth with capability-tier coverage (SC-2, D-47): v2 frozen superset + gedit (GTK3) + chromium-x11 (XWayland) | ✓ VERIFIED | matrix-v3.yaml untouched in the window (git log empty); matrix.go's only change is the hermeticity predicate (caseNeedsBaseEstablishment routing, 4dc4877) — case surface and quiesce machinery intact (constants re-confirmed at matrix.go:772/:787) |
| 6 | D-48 double-run mechanics: matrix runs TWICE back-to-back in one job, any fall = job failure, no interference between runs | ✓ VERIFIED | e2e-matrix.yml untouched in the window (git log empty) — exactly two sequential run steps with the D-48 comment, D48_SKIP gates only skip; quick regression greps this session: 6 `run e2e-matrix-v` task references, nightly cron `10 1 * * *` at :61; live proof run 35108412175 stands |
| 7 | D-48 formal gate on a genuinely untouched session (SC-2 «на нетронутой сессии») | ✓ VERIFIED | Unchanged disposition: gate redefined by owner directive 2026-09-17 into the unattended nightly D-48 v2 with machine-checked freshness; all in-repo artifacts untouched in the window and previously line-verified (truth 13); owner accepted the UAT test by evidence 2026-10-07; human-acceptance dimension tracked by the orchestrator's UAT session |
| 8 | Performance measured: daemon memory < 50 MB (SC-3, INST-03 memory half) | ✓ VERIFIED | VmHWM 12364–12368 kB (< 51200 kB) across three full live runs 2026-09-16 (perf-report.txt); budget gate unit-pinned (TestPerfBudgetGate); perf.go/perf_test.go untouched in the window |
| 9 | Performance measured: p95 hotkey reaction < 50 ms (SC-3, INST-03 latency half) | ✓ PASSED (override) | Override carried forward verbatim: owner decision (в) 2026-09-16, WINDOWS #5 closed — injector-dominated window, daemon reaction 0.2–1.8 ms. Re-affirmed this session: the honest numbers and the budget-fail paragraph remain published (README.md:397 «**превышает**», README.en.md:387 «**exceeds**»); READMEs untouched in the window |
| 10 | Owner manual acceptance passed (SC-4, D-49, SPEC §7.2) | ✓ VERIFIED | Recorded owner verdict: 04-UAT.md test 1 «result: pass (owner, 2026-09-16)», issues: 0; docs/ACCEPTANCE.md untouched in the window (advisory on unfilled result fields stands); GHBDTN copy fix in place |
| 11 | README + install instructions published (SC-4, D-50/D-46): bilingual pair, one-command install, perf table from the live run | ✓ VERIFIED | Pair untouched in the window (git log empty); quick regression greps this session re-confirmed the budget paragraphs in both files; zero bare `ibus write-cache` documented (previous grep clean; docs untouched) |
| 12 | (04-08, G-4-1) Focus-first witness traversal + witnessProbeTimeout 10s + witnessQuiesceWindow single source; exhaustive modes untouched | ✓ VERIFIED | matrix.go changed in the window (hermeticity predicate only) — quiesce corpus RE-RUN green under `-race` this session (TestWitnessProbeBudget_ScalabilityFloor, TestMatrixQuiesce_WindowCoversProbes, TestMatrixQuiesce_WindowSource); constants intact at matrix.go:772 (witnessProbeTimeout=10s), :787 (window), single-source deadline :798, deadline+error :798/:810; the hermeticity predicate does not touch the witness path; live witness probe (`gnome-terminal-server:TERMINAL:chars=-1`, exit 0) recorded at the previous pass — surface.go/focus_helper.py untouched |
| 13 | (04-09, G-4-2) Nightly D-48 v2 gate wired: schedule + idle-desktop preflight with soft mode, focused-app primitive, dispatch script, machine-setup docs; nothing in-repo executes sudo | ✓ VERIFIED | All artifacts untouched in the window (git log empty) — the previous line-level verification stands (cron :61, idle preflight :100 with error/soft arms, fresh-session gate :139, exactly 3 D48_SKIP gates, bash -n clean dispatch script with exec bit, ci-runner.md D-48 v2 blocks; sudo grep = 2 benign hits); quick regression greps this session: cron and task references confirmed in place |

**Score:** 13/13 truths verified (12 VERIFIED + 1 PASSED override; 0 present-behavior-unverified)

**Verification-method notes (this refresh):**
- Quick regression scope per orchestrator: the install/selfcheck truths touched by the G-5-5 wrap completion got full behavioral re-verification; the unchanged-but-covered surfaces got existence + structural sanity (git log empty for the window + targeted greps).
- Single named-test runs only, never the full suite: `go test -race -count=1 ./internal/install/ -run 'TestWrapSources|TestSelfcheck_InputSource|TestInstall_Sequence|TestSelfcheck_AllGreen|TestUninstall_FullRollback|TestUninstall_Idempotent'` — 17/17 subtests PASS; `go test -race -count=1 ./test/e2e/ -run 'WitnessProbeBudget|MatrixQuiesce'` — PASS; `go build ./...` — OK.
- The G-2-3 actor change and the hermeticity predicate inside this window's file set are behaviorally verified by the quick task's own verification (261008-00m-VERIFICATION.md: truths 1–2 and 6, named tests green under `-race`, live matrix v1 16/16) — not duplicated here beyond the compile + quiesce regression.
- Refusal identities confirmed twice over: `errors.Is` pins in TestWrapSourcesRefusalTable (3 residue rows → errMixedSources; unsupported-kind + nothing-wrappable → errUnsupportedPair; plus empty-value, no-raw-line, fix-hint guarantees) and TestSelfcheck_InputSourceResidueStaysRed (ZERO set calls, D-53 rationale) — all green this session.
- Fingerprint re-derived via `verification.fingerprint` on the current tree: `v1:sha256:13ff04f69498d40a1d8d8aa2a4e059ff7758c89b0ac905ee7f3f4b89eb5f2bdb` (file set unchanged from the previous pass; all 54 files exist).

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/install/install.go` (+tests) | Install/Uninstall/writeAtomic, Runner seam, XML+unit render; G-5-5 completion arbiter | ✓ VERIFIED | wrapSources :837-880 (completion via renderWrapped + first-occurrence dedupe; residue/unsupported/nothing-wrappable refusals with stable identities), resolveWrapInput :892-921, saveState gate :535-549; named corpus green -race this session |
| `internal/install/selfcheck.go` (+test) | six-step D-41 audit + corpus; half-state heal | ✓ VERIFIED | checkInputSource :196-246 heal path verified by reading + TestSelfcheck_InputSourceMixedHeals / TestSelfcheck_InputSourceResidueStaysRed green this session |
| `cmd/goswitchctl/main.go`, `cmd/goswitchd/{main,version}.go` (+tests) | install/uninstall/selfcheck; version vars, -version | ✓ VERIFIED | untouched in the window; build green this session |
| `internal/session/actor.go`, `internal/ctlsvc/ctlsvc.go` (+tests) | Version in status snapshot, `version=` token | ✓ VERIFIED | actor.go changed in the window (G-2-3 flipCredit only — behaviorally verified by the quick task corpus); `version=` token re-confirmed at ctlsvc.go:144 |
| `test/e2e/perf.go` + `perf_test.go` + `focus_helper.py` | percentile, /proc parser, resident witness, budget gate | ✓ VERIFIED | untouched in the window; previous named-corpus + live-probe evidence stands |
| `test/e2e/surface.go`, `matrix.go`, `preflight.go`, `cases/matrix-v3.yaml`, `quiesce_test.go` | v3 corpus + drivers; 10s floor + window pins | ✓ VERIFIED | matrix.go's window change is the hermeticity predicate; quiesce corpus green -race this session; constants intact |
| `test/e2e/case_install.go`, `watchdog_test.go`, `main.go` | install-cycle case; watchdog default | ✓ VERIFIED | untouched in the window |
| `.github/workflows/e2e-matrix.yml` | double run + preflights + soft mode + nightly schedule | ✓ VERIFIED | untouched in the window; cron + task-reference greps re-confirmed this session |
| `.github/workflows/release.yml`, `.goreleaser.yaml`, `mise.toml` | tag v* → goreleaser; two builds, checksums; goreleaser pin | ✓ VERIFIED | untouched in the window; proven live by the published release |
| `README.md`, `README.en.md`, `docs/ACCEPTANCE.md`, `docs/ci-runner.md` | bilingual pair, UAT checklist, D-48 + D-48 v2 procedures | ✓ VERIFIED | untouched in the window; budget paragraphs re-grepped this session (README.md:397, README.en.md:387) |
| `scripts/d48-nightly-dispatch.sh` | gh-guarded session dispatcher | ✓ VERIFIED | untouched in the window; exec bit confirmed this session |
| `.planning/phases/04-postavka-i-priemka/red-evidence/` | TDD RED records for phase tasks | ✓ VERIFIED | all records present; the new-window RED evidence (e0e21ec/77c91b7/07d143b) lives in the quick task's commit history per D-55 |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| cmd/goswitchctl/main.go | internal/install/install.go | thin-client dispatch | ✓ WIRED | untouched in the window |
| internal/install/install.go | ibus write-cache env | IBUS_COMPONENT_PATH every call | ✓ WIRED | writeCache/componentPathEnv untouched; build green |
| install save | uninstall restore | install-state.json (0600) | ✓ WIRED | saveState gate now computes the FULL wrap atomically before the backup; restore chain untouched |
| cmd/goswitchctl/main.go | internal/install/selfcheck.go | .Selfcheck( dispatch | ✓ WIRED | untouched in the window |
| internal/ctlsvc/ctlsvc.go | internal/session/actor.go | `version=` token | ✓ WIRED | re-confirmed ctlsvc.go:144 this session |
| internal/install/selfcheck.go | internal/engine | live engine probe | ✓ WIRED | imports intact; package compiles and corpus green |
| .goreleaser.yaml | cmd/goswitchd/main.go | ldflags → main.version | ✓ WIRED | stamped 1.0.0 in the published binary |
| .github/workflows/release.yml | mise.toml | mise [tools] pin | ✓ WIRED | untouched in the window |
| test/e2e/perf.go | focus_helper.py / /proc | witness line protocol; VmHWM | ✓ WIRED | untouched in the window |
| .github/workflows/e2e-matrix.yml | mise.toml | two × e2e-matrix-v3/v4 | ✓ WIRED | untouched; 6 task references re-grepped |
| test/e2e/cases/matrix-v3.yaml | test/e2e/matrix.go | closed surface vocabulary | ✓ WIRED | yaml untouched; matrix.go window change is routing-only |
| .github/workflows/e2e-matrix.yml (idle preflight) | test/e2e/focus_helper.py (focused-app) | `/usr/bin/python3 … focused-app` | ✓ WIRED | untouched; live `(none)` ×2 recorded at the previous pass |
| systemd user unit goswitch-d48-dispatch (docs) | scripts/d48-nightly-dispatch.sh | ExecStart=%h/goswitch/scripts/… | ✓ WIRED (documented) | untouched; exec bit confirmed |
| root timer d48-nightly-relogin | GDM autologin | restart gdm → autologin | ✓ WIRED (documented) | untouched; owner applies by design |
| README pair | goswitchctl install | one-command instruction | ✓ WIRED | untouched in the window |
| тег v1.0.0 | release.yml → published release | tag trigger | ✓ WIRED | release live, untouched in the window |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| checkInputSource heal | `healed` list | wrapSources(parsed live gsettings line) — one arbiter, raw line never reaches the argument vector (D-20/ASVS V5) | yes — TestSelfcheck_InputSourceMixedHeals asserts the exact set args + green verdict (unit-pinned through the call seam) | ✓ FLOWING |
| saveState gate | resolved wrap input | readSources(live gsettings) → resolveWrapInput → wrapSources | yes — TestInstall_Sequence* green; residue refuses before any mutation | ✓ FLOWING |
| README perf table | p50/p95/p99, VmHWM | perf-report.txt (live runs ×3) | yes | ✓ FLOWING |
| release assets | binary + checksums | goreleaser build at tag 2bc8c6a | yes | ✓ FLOWING |
| selfcheck steps | live system state | systemctl/ibus/gsettings probes | yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | -------- | ------ | ------ |
| G-5-5 corpus: completion + re-pinned refusal table + heal + residue-zero-write + install/uninstall/selfcheck regression | `go test -race -count=1 ./internal/install/ -run 'TestWrapSources\|TestSelfcheck_InputSource\|TestInstall_Sequence\|TestSelfcheck_AllGreen\|TestUninstall_FullRollback\|TestUninstall_Idempotent'` | 17/17 subtests PASS (incl. TestWrapSourcesRefusalTable 6/6 rows with `errors.Is` identity pins; TestWrapSourcesHalfWrappedCompletes 2/2; TestSelfcheck_InputSourceMixedHeals exactly-one-set; TestSelfcheck_InputSourceResidueStaysRed zero-write) | ✓ PASS |
| Quiesce budget corpus (truth 12, on the window-changed matrix.go) | `go test -race -count=1 ./test/e2e/ -run 'WitnessProbeBudget\|MatrixQuiesce'` | PASS | ✓ PASS |
| Workspace build | `go build ./...` | OK | ✓ PASS |
| Refusal identities (structural read, plus the `errors.Is` pins above) | read install.go:844-848/:859-863, selfcheck.go:212-216 | errMixedSources = residue forms; errUnsupportedPair = unsupported-kind + nothing-wrappable; both stable, both pinned by test | ✓ PASS |
| D-48 v2 / README quick greps | cron :61, task refs, budget paragraphs :397/:387, dispatch exec bit | all in place | ✓ PASS |
| Live-desktop behaviors (install-cycle, perf window, matrix double run, release) | recorded runs 35108412175 / 35130309174, perf-report.txt | recorded | ✓ EVIDENCED (not re-run) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` probes exist (the project's probe surface is mise tasks + named go tests). The phase-declared live scenarios are covered by the spot-check table above; the quick task's live matrix runs (v1 16/16 on the fix tree) are recorded in 261008-00m-VERIFICATION.md.

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
| ----------- | ----------- | ----------- | ------ | -------- |
| INST-01 | 04-01…04-09 (04-08/04-09 via scope_note nearest-coverage) | Установка без root: systemd user unit, регистрация engine в IBus, `go install` + бинарник из GitHub releases | ✓ SATISFIED (REQUIREMENTS.md: Complete) | Truths 1–7, 11–13: install/uninstall/selfcheck behaviorally re-proven this session on the G-5-5 tree; release v1.0.0 published; nightly gate wired and untouched |
| INST-03 | 04-04 | Реакция на горячую клавишу < 50 мс; память < 50 МБ | ✓ SATISFIED WITH WAIVER (memory verified; latency waived by owner, WINDOWS #5 closed) | Truths 8–9. REQUIREMENTS.md checkbox intentionally unchecked pending the owner's acceptance-time flip (04-07-SUMMARY decision, MACR-01 precedent) — a milestone-close/orchestrator act, not a missing implementation |

No orphaned requirements: REQUIREMENTS.md maps exactly {INST-01, INST-03} to Phase 4; both are claimed by plans.

### Decision Coverage

Gate `check.decision-coverage-verify` (non-blocking): 14/14 trackable CONTEXT.md decisions honored at the previous pass; the window's decisions (G-2-3/G-5-5 revisions) were spec-delta'd per D-55 and verified in 261008-00m-VERIFICATION.md truth 4 — no decision vanished during execution.

### Prohibitions

| Prohibition | Tier | Status | Evidence |
| ----------- | ---- | ------ | -------- |
| Zero bare `ibus write-cache` documented anywhere (Pitfall 1) | automated | ✓ HOLDS | READMEs + docs untouched in the window (previous grep: 0 hits) |
| (04-08 p1, judgment) Focus-first must NOT contaminate exhaustive modes | judgment | ✓ HOLDS — enforcement evidence recorded | matrix.go's window change adds routing predicates only; witness path and exhaustive focused-inputs families untouched (constants and call sites re-confirmed); previous structural + call-graph evidence stands |
| (04-08 p2, automated) Witness prints only role + char count, never surface text (D-20/D-21) | automated | ✓ ENFORCED | surface.go/witness path untouched; live grammar proof recorded at the previous pass; the NEW wrap/heal verdicts are D-20-safe by construction and test (refusal-table pin: verdict must not carry the raw user line) |
| (04-09, judgment) Nothing in the repository executes sudo/root | judgment | ✓ HOLDS | scripts/, workflows, mise.toml, test/e2e untouched in the window (previous grep: 2 benign hits) |
| (G-5-5 superseded contract — informational) half-wrapped lists were refused outright | judgment | ✓ SUPERSEDED BY OWNER DECISION | ADR-006 amendment + SPEC §4.3/§4.4 deltas (0781a5f, docs-only before code); completion now mandated; residue refusals keep identities (this session's corpus) |

### Anti-Patterns Found

Debt-marker scan (TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER) across all seven window-changed files: **zero markers**. No stub patterns in the changed code (the completion/dedupe/heal logic is fully implemented with test-pinned behavior); build green.

Re-verification evidence gate (#3304) applied: the previous verification had no `gaps:` list (carried-forward gap set empty), and the only window-modified files are exactly the refresh scope — no new-scope Step-7 findings arose, so no advisory downgrade was needed. The one carried advisory (ACCEPTANCE.md result fields unfilled) remains in frontmatter `advisory:` and does not affect the verdict.

### Deferred Items

| # | Item | Disposition | Evidence |
|---|------|-------------|----------|
| 1 | combo-word-layout historical flake (04 deferred-items.md: post-reload tap after closeZenity in an unusual focus state) | Informational — never a failed truth; no later phase claims it; post-release maintenance material | .planning/phases/04-postavka-i-priemka/deferred-items.md; candidate remedy documented in place |

### Human Verification Required

None from this verifier. The phase's two designed human gates are resolved with recorded evidence: D-49 passed live by the owner (04-UAT.md test 1, 2026-09-16), D-48 redefined by owner directive into the unattended nightly gate D-48 v2 whose in-repo artifacts are wired and re-confirmed untouched, with the owner's acceptance-by-evidence recorded 2026-10-07. The milestone-close human-acceptance dimension is tracked by the orchestrator's UAT session.

### Gaps Summary

None. No truth FAILED, no artifact missing/stub, no key link unwired, no blocker anti-pattern, zero debt markers. The staleness window (quick 261008-00m) is a coherent owner-mandated behavior revision that was spec-delta'd before code (D-55), verified 6/6 by the quick task's own verification, and re-proven here on the phase-4 must-have surface: the install/selfcheck corpus is green under `-race` on the current tree, the residue/unsupported refusal identities are intact and test-pinned, and every other phase-4 deliverable is bit-identical to the previously line-verified state (empty git log for the window). The wrap-completion change strengthens truth 1 (the install gate now completes a completable half-state instead of refusing it) and truth 2 (selfcheck self-heals the same state) — both directions covered by fresh named-test evidence.

Recorded notes:
- The latency half of INST-03 remains waived by owner decision (в), WINDOWS #5 closed — the honest numbers and budget-fail status are still published in both READMEs (re-grepped this session).
- docs/ACCEPTANCE.md checklist result fields remain unfilled (advisory above); the owner's recorded verdicts live in 04-UAT.md and planning history.
- The dispatch script and workflow default target matrix-v4 (live superset) by design; the phase-4-era 31/31 v3 proof (run 35108412175) remains the recorded v1.0.0 evidence.
- The G-2-3 (session actor) and hermeticity (matrix routing) changes inside this window belong to the quick task's verified scope; this refresh confirmed the phase-4-adjacent surfaces they touch (version token, quiesce machinery) remain intact.

---

_Verified: 2026-10-07T23:08:38Z_
_Verifier: Claude (gsd-verifier) — second staleness refresh at milestone close_
