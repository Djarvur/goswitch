---
phase: 04-postavka-i-priemka
verified: 2026-10-07T20:59:03Z
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
covered_digest: "v1:sha256:522c329aafb1887584bafefab2c148f5780ac33f1b4b1fd96cb2a93c3982ab3a"
behavior_unverified: 0
overrides_applied: 1
overrides:
  - must_have: "Производительность: p95 реакции на горячую клавишу < 50 мс (INST-03, latency half)"
    reason: "Owner waiver decision (в) 2026-09-16, WINDOWS #5 (closed): the prescribed ydotool→AT-SPI window is dominated by ydotool 0.1.8 injector overhead (~80–125 ms/event); daemon reaction is 0.2–1.8 ms. v1.0.0 ships the honest as-measured p50/p95/p99 with the budget-fail status published in both READMEs. Re-verified 2026-10-07: the honest numbers and the «exceeds/превышает» budget paragraph are still published in README.md:385-397 and README.en.md:375-387."
    accepted_by: "owner (Daniel Podolsky), recorded by goswitch phase execution"
    accepted_at: "2026-09-16"
re_verification:
  previous_status: human_needed
  previous_score: 9/11
  gaps_closed:
    - "D-49 owner manual acceptance (previous human item 1): passed live by the owner 2026-09-16, recorded in 04-UAT.md test 1 (result: pass, issues: 0)"
    - "D-48 formal fresh-session gate (previous human item 2, truth 7): redefined by owner directive 2026-09-17 («remove the human») into the unattended nightly gate D-48 v2 — gaps G-4-1/G-4-2 resolved by plans 04-08/04-09, gate artifacts wired and re-verified on the current tree, owner accepted the UAT test by evidence 2026-10-07 (machine-checked freshness on demand remains the gate)"
    - "Judgment-tier prohibition residue (previous human item 3): whole-surface nothing-outside-$HOME/no-user-content/never-silently-green — exercised during the owner's live D-49 session (issues: 0) and re-checked by this verifier (grep + live probes, below)"
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
**Verified:** 2026-10-07T20:59:03Z
**Status:** passed
**Re-verification:** Yes — fingerprint-staleness refresh at milestone close (v1.0.0 RELEASED; tree evolved since 2026-09-16: phases 5–8 merged, engine/ + layouts/ moved under internal/ in 5427d7a, README pair restructured russian-first in 6f2742d, matrix v3→v4 in 2c41b83, nightly gate built per owner directive)

## Goal Achievement

All must-have paths re-derived against the CURRENT tree. The engine/ + layouts/ move did not touch any phase-4 covered file (internal/install, internal/ctlsvc, internal/session were already under internal/; the selfcheck→engine probe link compiles and its corpus is green). README.ru.md no longer exists — the bilingual pair is now README.md (RU, канон) + README.en.md (EN), with ACCEPTANCE.md links re-pointed (b65b62f). The e2e matrix grew v3→v4 in phase 6 by deliberate superset extension (2c41b83 «nightly on v4»); matrix-v3.yaml and its mise task remain intact.

### Observable Truths

| # | Truth | Status | Evidence |
| --- | ----- | ------ | -------- |
| 1 | Root-less install in one command (SC-1, D-39/D-40): component XML under `$HOME`, systemd user unit, env-cache registration with IBUS_COMPONENT_PATH on every write, unit active, sources takeover, activation | ✓ VERIFIED | Fresh behavioral evidence this session: TestInstall_SequenceSingleSource PASS under `-race`; full rollback corpus green (row 3); install-state machinery intact at internal/install/install.go:368 (Install), :632 (writeCache), :648 (componentPathEnv); recorded live install-cycle ×2 (SUMMARYs, VALIDATION audit) |
| 2 | selfcheck D-41: six steps, ok/FAIL + fix hint, exit≠0 on red, exactly-once env-cache repair | ✓ VERIFIED | TestSelfcheck_AllGreen PASS under `-race` this session; selfcheck.go intact (D-41 step comments :141+); live green/red probes recorded in 04-02-SUMMARY |
| 3 | Uninstall full rollback (SC-1, D-42): unit stop/disable, artifacts removed, cache re-written with env, sources restored with readback, state file; `--purge` opt-in | ✓ VERIFIED | TestUninstall_FullRollback + TestUninstall_Idempotent PASS under `-race` this session; saveState :534 / restoreSources :986 / restoreSwitchBinding :1033 / restoreToolkitAccessibility :1087 all present (phase 8 later extended the a11y restore — out of phase-4 scope) |
| 4 | Release channels: release tar.gz + checksums, `go install` channel, version stamping vs honest dev (SC-1, D-37/D-38) | ✓ VERIFIED | **Re-verified live today:** `gh release view v1.0.0` → published 2026-09-16T17:47:32Z, isDraft false, assets `goswitch_1.0.0_linux_amd64.tar.gz` (3.6 MB) + `checksums.txt`; proxy `v1.0.0.info` Hash `2bc8c6a…` == `git rev-parse v1.0.0`; `go run ./cmd/goswitchd -version` → `goswitchd dev`, exit 0 (honest unstamped fallback); `go build ./...` green |
| 5 | Matrix v3 full breadth with capability-tier coverage (SC-2, D-47): v2 frozen superset + gedit (GTK3) + chromium-x11 (XWayland) | ✓ VERIFIED | matrix-v3.yaml intact (31 cases at the v1.0.0 tag; one later-phase reconciliation case `flip-after-word-correction` appended by f679b95, mixed rows re-pinned per-character by f9ed497 — deliberate superset growth, v2 untouched); mise tasks e2e-matrix/-v2/-v3/-v4 all present (mise.toml:97-109); matrix-v4 (34 cases) added by phase 6 WITHOUT removing v3; spike-pinned expectations still in the v3 header |
| 6 | D-48 double-run mechanics: matrix runs TWICE back-to-back in one job, any fall = job failure, no interference between runs | ✓ VERIFIED | e2e-matrix.yml carries exactly two sequential run steps (run #1 / run #2, lines ~198/217) with the D-48 comment «ПРОГОН №2 СРАЗУ после №1… вмешательство между прогонами = подмена проверки D-48»; D48_SKIP gates only *skip* runs on preflight failure — they never insert steps between the two runs; live proof run 35108412175 (31/31 twice, conclusion success) stands |
| 7 | D-48 formal gate on a genuinely untouched session (SC-2 «на нетронутой сессии») | ✓ VERIFIED | **Gate redefined by owner directive 2026-09-17** («придумай, как тестировать это без участия человека», 04-UAT.md G-4-2): the formal evidence is now the unattended nightly D-48 v2 double-run with machine-checked freshness — every in-repo artifact is wired on the current tree (truth 13) and the machine-side chain is documented for one-time owner application (docs/ci-runner.md «Автономный ночной гейт (D-48 v2)»). G-4-1 (witness scalability) and G-4-2 (human in the loop) resolved by plans 04-08/04-09; owner accepted the UAT test by evidence 2026-10-07; the human-acceptance dimension is tracked by the orchestrator's UAT session, not by this report |
| 8 | Performance measured: daemon memory < 50 MB (SC-3, INST-03 memory half) | ✓ VERIFIED | VmHWM 12364–12368 kB (< 51200 kB) across three full live runs 2026-09-16 (perf-report.txt); budget gate unit-pinned (TestPerfBudgetGate) and fired exit≠0 on the latency boundary in every run; perf.go parser/gate intact (named tests green in the -race corpus) |
| 9 | Performance measured: p95 hotkey reaction < 50 ms (SC-3, INST-03 latency half) | ✓ PASSED (override) | Measured honestly: p95 184.4 ms (p50 166.9 / p99 186.5) in the prescribed ydotool→AT-SPI window — gate FAIL by design, reproducible. **Override carried forward:** owner decision (в) 2026-09-16, WINDOWS #5 closed — injector-dominated window, daemon reaction 0.2–1.8 ms; README publishes the real numbers with the explicit budget-fail paragraph in both languages (re-verified today, truth 11). Phase 6's later perf gate (Pitfall-5 delta ≤ +5 ms) passed live — the phase-4 numbers remain the published baseline |
| 10 | Owner manual acceptance passed (SC-4, D-49, SPEC §7.2) | ✓ VERIFIED | Recorded owner verdict: 04-UAT.md test 1 «result: pass (owner, 2026-09-16)», issues: 0 — the live session covered install from README → selfcheck → gesture set → reload → status → uninstall; docs/ACCEPTANCE.md checklist present and wired (item 1 = the formal D-48 gate, links re-pointed to README.en.md by b65b62f); the GHBDTN copy fix (6e90c34) still in place; milestone-close acceptance dimension → orchestrator UAT session |
| 11 | README + install instructions published (SC-4, D-50/D-46): bilingual pair, one-command install, perf table from the live run | ✓ VERIFIED | Pair restructured russian-first (6f2742d): README.md (RU, канон) + README.en.md (EN); `goswitchctl install` appears 6× in each; zero bare `ibus write-cache` documented anywhere (grep clean — Pitfall 1 prohibition holds); perf table matches perf-report.txt in both (p50 166.9 / p95 184.4 / p99 186.5, peak RSS 12.4 MB) with the honest «**exceeds** / **превышает**» budget paragraph (README.en.md:387, README.md:397) |
| 12 | (04-08, G-4-1) Focus-first witness traversal + witnessProbeTimeout 10s + witnessQuiesceWindow single source; exhaustive modes untouched | ✓ VERIFIED | Quiesce corpus 3/3 PASS under `-race` **run by this verifier today** (TestWitnessProbeBudget_ScalabilityFloor, TestMatrixQuiesce_WindowCoversProbes, TestMatrixQuiesce_WindowSource); constants at matrix.go:772/:779/:787, deadline+error single-source :798/:810; **live witness probe run by this verifier today:** `gnome-terminal-server:TERMINAL:chars=-1`, exit 0, instant — focus-first pass + frame-fallback arm both exercised; commit 4bf01fc diff hunks touch only the docstring, char_count and the witness path — focused_inputs (:307) / cmd_text (:290) / focused_input (:327) outside every hunk, and the witness functions' only executable callers are cmd_witness :254/:256 (call-graph grep) |
| 13 | (04-09, G-4-2) Nightly D-48 v2 gate wired: schedule + idle-desktop preflight with soft mode, focused-app primitive, dispatch script, machine-setup docs; nothing in-repo executes sudo | ✓ VERIFIED | All structural gates re-run green by this verifier: cron `10 1 * * *` (:61); idle preflight calls `focus_helper.py focused-app` for ALL triggers (:100) with error arm (:115) and schedule soft arm (warning + D48_SKIP=1, :111-112); fresh-session preflight armed on `fresh_session==true \|\| schedule` (:139) with folded soft/hard failure; exactly 3 `env.D48_SKIP != '1'` gates (run #1, run #2, upload); double run intact; `scripts/d48-nightly-dispatch.sh` — bash -n clean, exec bit, gh guard + `gh auth status`, `D48_REF:-main`, `--repo Djarvur/goswitch`, `fresh_session=true`, matrix-v4 (updated by 2c41b83 to track the live superset); docs/ci-runner.md:191 «Автономный ночной гейт (D-48 v2)» with all three blocks (AutomaticLoginEnable :228, d48-nightly-relogin.timer :248, goswitch-d48-dispatch ExecStart→script :298); **live focused-app probe run by this verifier today:** `(none)` twice, exit 0, one line — the arm 04-09 could not reproduce is now live-proven; sudo grep in scripts/ + workflows + mise + test/e2e = 2 hits, both benign (a comment stating the prohibition; a preflight error hint «sudo apt install gedit» addressed to the human) |

**Score:** 13/13 truths verified (12 VERIFIED + 1 PASSED override; 0 present-behavior-unverified)

**Verification-method notes:**
- Single named tests only, never the full suite: quiesce corpus (`go test ./test/e2e/ -run 'WitnessProbeBudget|MatrixQuiesce' -race`) and install/selfcheck/uninstall corpus (`go test ./internal/install/ -run 'TestInstall_Sequence|TestSelfcheck_AllGreen|TestUninstall_FullRollback|TestUninstall_Idempotent' -race`) — all green; plus `go build ./...` and the `-version` flag contract.
- Live read-only a11y probes executed by this verifier on the owner's desktop (the plans' own verify commands): `focused-app` ×2 → `(none)`/`(none)` exit 0; `witness` → `gnome-terminal-server:TERMINAL:chars=-1` exit 0. Both binary contracts (one line, stable, exit 0) hold live.
- Recorded (not re-run) evidence: live-desktop install-cycle, perf ×3, matrix double-run 35108412175, release run 35130309174 — recorded in SUMMARYs and re-corroborated today by cheap re-checks (release publication, proxy identity, tag/commit match, file structure).
- `verify.artifacts`: 8/9 plans all-pass; 04-07 flags `README.ru.md` — the file was absorbed by the russian-first restructure (6f2742d); the deliverable exists as README.md + README.en.md and was content-verified above (truth 11).
- `verify.key-links`: 5/9 plans tool-clean; the 7 pseudo-component links the tool cannot resolve were all verified manually on the current tree — install save→restore (install-state.json const :63, saveState :534, restore chain :986+), README pair→`goswitchctl install` (6 mentions each), тег v1.0.0→release.yml (tag 2bc8c6a, release published), workflow→focused-app (:100), systemd unit→dispatch script (ci-runner.md:298), root timer→autologin (ci-runner.md:248/:228, documented-only per plan boundary).

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/install/install.go` (+tests) | Install/Uninstall/writeAtomic, Runner seam, XML+unit render | ✓ VERIFIED | wired (cmd/goswitchctl dispatch, e2e case_install); named corpus green -race today |
| `internal/install/selfcheck.go` (+test) | six-step D-41 audit + corpus | ✓ VERIFIED | TestSelfcheck_AllGreen green today |
| `cmd/goswitchctl/main.go`, `cmd/goswitchd/{main,version}.go` (+tests) | install/uninstall/selfcheck; version vars, -version | ✓ VERIFIED | live `goswitchd dev` exit 0 today |
| `internal/session/actor.go`, `internal/ctlsvc/ctlsvc.go` (+tests) | Version in status snapshot, `version=` token | ✓ VERIFIED | intact post engine-move; corpus green at HEAD (phase 8 verification 44/44) |
| `test/e2e/perf.go` + `perf_test.go` + `focus_helper.py` | percentile, /proc parser, resident witness, budget gate; witness focus-first + focused-app | ✓ VERIFIED | named corpus green; py_compile clean; live probes by this verifier |
| `test/e2e/surface.go`, `matrix.go`, `preflight.go`, `cases/matrix-v3.yaml`, `quiesce_test.go` | v3 corpus + drivers; 10s floor + window pins | ✓ VERIFIED | 31-case tag corpus intact (+1 documented later-phase case); quiesce corpus green today |
| `test/e2e/case_install.go`, `watchdog_test.go`, `main.go` | install-cycle case; watchdog default | ✓ VERIFIED | intact |
| `.github/workflows/e2e-matrix.yml` | double run + preflights + soft mode + nightly schedule | ✓ VERIFIED | structure re-verified line-by-line today |
| `.github/workflows/release.yml`, `.goreleaser.yaml`, `mise.toml` | tag v* → goreleaser; two builds, checksums; goreleaser pin | ✓ VERIFIED | proven live by the published release; config gates green per SUMMARYs |
| `README.md`, `README.en.md`, `docs/ACCEPTANCE.md`, `docs/ci-runner.md` | bilingual pair, UAT checklist, D-48 + D-48 v2 procedures | ✓ VERIFIED | content checks above; README.ru.md → README.en.md rename handled |
| `scripts/d48-nightly-dispatch.sh` | gh-guarded session dispatcher | ✓ VERIFIED | bash -n, exec bit, content verified |
| `.planning/phases/04-postavka-i-priemka/red-evidence/` | TDD RED records for phase tasks | ✓ VERIFIED | 04-01×2, 04-02×2, 04-04, 04-08 staged two-form RED all present |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| cmd/goswitchctl/main.go | internal/install/install.go | thin-client dispatch | ✓ WIRED | tool-verified |
| internal/install/install.go | ibus write-cache env | IBUS_COMPONENT_PATH every call | ✓ WIRED | tool-verified; writeCache/componentPathEnv intact |
| install save | uninstall restore | install-state.json (0600) | ✓ WIRED | manual: const :63, saveState :534, restore :986+ |
| cmd/goswitchctl/main.go | internal/install/selfcheck.go | .Selfcheck( dispatch | ✓ WIRED | tool-verified |
| internal/ctlsvc/ctlsvc.go | internal/session/actor.go | `version=` token | ✓ WIRED | tool-verified |
| internal/install/selfcheck.go | internal/engine (moved 5427d7a) | live engine probe | ✓ WIRED | imports re-pointed to internal/engine; corpus compiles+green |
| .goreleaser.yaml | cmd/goswitchd/main.go | ldflags → main.version | ✓ WIRED | stamped 1.0.0 in the published binary |
| .github/workflows/release.yml | mise.toml | mise [tools] pin | ✓ WIRED | tool-verified |
| test/e2e/perf.go | focus_helper.py / /proc | witness line protocol; VmHWM | ✓ WIRED | tool-verified |
| .github/workflows/e2e-matrix.yml | mise.toml | two × e2e-matrix-v3/v4 | ✓ WIRED | manual: run #1/#2 with input-selected task |
| test/e2e/cases/matrix-v3.yaml | test/e2e/matrix.go | closed surface vocabulary | ✓ WIRED | tool-verified |
| .github/workflows/e2e-matrix.yml (idle preflight) | test/e2e/focus_helper.py (focused-app) | `/usr/bin/python3 … focused-app` | ✓ WIRED | manual: :100; live `(none)` ×2 today |
| systemd user unit goswitch-d48-dispatch (docs) | scripts/d48-nightly-dispatch.sh | ExecStart=%h/goswitch/scripts/… | ✓ WIRED | manual: ci-runner.md:298; documented-only per plan boundary |
| root timer d48-nightly-relogin | GDM autologin | restart gdm → autologin | ✓ WIRED (documented) | manual: ci-runner.md:248/:228; owner applies by design |
| README pair | goswitchctl install | one-command instruction | ✓ WIRED | manual: 6 mentions each |
| тег v1.0.0 | release.yml → published release | tag trigger | ✓ WIRED | tag 2bc8c6a; release live with 2 assets today |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| README perf table | p50/p95/p99, VmHWM | perf-report.txt (live runs ×3) | yes | ✓ FLOWING |
| release assets | binary + checksums | goreleaser build at tag 2bc8c6a | yes | ✓ FLOWING |
| selfcheck steps | live system state (unit, component, engine, config, source) | systemctl/ibus/gsettings probes | yes | ✓ FLOWING |
| focused-app / witness output | live AT-SPI tree | D-Bus a11y bus (verified live today) | yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Quiesce budget corpus (G-4-1 fix) | `go test ./test/e2e/ -run 'WitnessProbeBudget\|MatrixQuiesce' -race -count=1` | 3/3 PASS | ✓ PASS |
| Install/selfcheck/uninstall corpus | `go test ./internal/install/ -run 'TestInstall_Sequence\|TestSelfcheck_AllGreen\|TestUninstall_FullRollback\|TestUninstall_Idempotent' -race -count=1` | 4/4 PASS | ✓ PASS |
| Workspace build (directive 2, cheap form) | `go build ./...` | ok | ✓ PASS |
| Version flag contract (D-37) | `go run ./cmd/goswitchd -version` | `goswitchd dev`, exit 0 | ✓ PASS |
| focused-app idle-preflight primitive (04-09 Task 1 verify) | `/usr/bin/python3 test/e2e/focus_helper.py focused-app` ×2 | `(none)` / `(none)`, exit 0, one line, stable | ✓ PASS |
| witness focus-first traversal + fallback arm (04-08) | `/usr/bin/python3 test/e2e/focus_helper.py witness` | `gnome-terminal-server:TERMINAL:chars=-1`, exit 0, instant | ✓ PASS |
| Release publication | `gh release view v1.0.0` + proxy `.info` | published, 2 assets, proxy Hash == tag commit | ✓ PASS |
| Live-desktop behaviors (install-cycle, perf window, matrix double run) | recorded runs 35108412175 / 35130309174, perf-report.txt | recorded | ✓ EVIDENCED (not re-run) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` probes exist (the project's probe surface is mise tasks + named go tests). The phase-declared live scenarios are covered by the spot-check table above.

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
| ----------- | ----------- | ----------- | ------ | -------- |
| INST-01 | 04-01…04-09 (04-08/04-09 via scope_note nearest-coverage) | Установка без root: systemd user unit, регистрация engine в IBus, `go install` + бинарник из GitHub releases | ✓ SATISFIED (REQUIREMENTS.md: Complete) | Truths 1–7, 11–13: install/uninstall/selfcheck proven live + unit-pinned; release v1.0.0 published and re-verified live today; nightly gate wired |
| INST-03 | 04-04 | Реакция на горячую клавишу < 50 мс; память < 50 МБ | ✓ SATISFIED WITH WAIVER (memory verified; latency waived by owner, WINDOWS #5 closed) | Truths 8–9. REQUIREMENTS.md checkbox intentionally unchecked pending the owner's acceptance-time flip (04-07-SUMMARY decision, MACR-01 precedent) — a milestone-close/orchestrator act, not a missing implementation |

No orphaned requirements: REQUIREMENTS.md maps exactly {INST-01, INST-03} to Phase 4; both are claimed by plans.

### Decision Coverage

Gate `check.decision-coverage-verify` (non-blocking): **14/14 trackable CONTEXT.md decisions honored** in shipped artifacts; `not_honored: []`. No decision vanished during execution.

### Prohibitions

| Prohibition | Tier | Status | Evidence |
| ----------- | ---- | ------ | -------- |
| Zero bare `ibus write-cache` documented anywhere (Pitfall 1) | automated | ✓ HOLDS | grep over README.md/README.en.md/docs: 0 hits (previous verification: 0) |
| (04-08 p1, judgment) Focus-first must NOT contaminate exhaustive modes (focused-inputs/cmd_text family) | judgment | ✓ HOLDS — enforcement evidence recorded | Deterministic: commit 4bf01fc diff hunks (docstring, char_count, witness path only) leave focused_inputs :307 / cmd_text :290 / focused_input :327 untouched; call-graph grep — witness functions' only executable callers are cmd_witness :254/:256; behavioral: exhaustive focused-inputs drives injection targeting in every matrix case and the matrix stayed functional across the phase 6–8 runs (3×35-case live runs + phase-8 UAT). Verdict is structural/LLM-judge — semantic residue (focus-lag enumeration correctness) remains covered by the 02-02 contract and its own corpus, not by this check |
| (04-08 p2, automated) Witness prints only role + char count (app:role:chars=N), never surface text (D-20/D-21) | automated | ✓ ENFORCED | Live output today: `gnome-terminal-server:TERMINAL:chars=-1` — grammar exact; no text content in any arm |
| (04-09, judgment) Nothing in the repository executes sudo/root; root blocks are documentation only | judgment | ✓ HOLDS | sudo grep over scripts/ + .github/workflows/ + mise.toml + test/e2e/*.go: 2 hits — a comment stating the prohibition (dispatch script :10) and a preflight error hint addressed to the human (preflight.go:193); root blocks exist only inside ci-runner.md copy-paste documentation |

### Anti-Patterns Found

Debt-marker scan (TBD/FIXME/XXX) across all 30 phase-4 source/config/doc files including the 04-08/04-09 additions (quiesce_test.go, d48-nightly-dispatch.sh, README.en.md): **zero markers**. No stub patterns; py_compile clean; bash -n clean.

Re-verification evidence gate (#3304) applied: no carried-forward gaps exist (previous file had none), and no Step-7 blocker was found on any file — no regression or new-scope findings to weigh. The one advisory entry (ACCEPTANCE.md result fields unfilled) is recorded in frontmatter `advisory:` and does not affect the verdict.

### Deferred Items

| # | Item | Disposition | Evidence |
|---|------|-------------|----------|
| 1 | combo-word-layout historical flake (04 deferred-items.md: post-reload tap after closeZenity in an unusual focus state) | Informational — never a failed truth; no later phase claims it; post-release maintenance material | .planning/phases/04-postavka-i-priemka/deferred-items.md; candidate remedy documented in place |

### Human Verification Required

None from this verifier. The phase's two designed human gates are resolved with recorded evidence: D-49 passed live by the owner (04-UAT.md test 1, 2026-09-16), D-48 redefined by owner directive into the unattended nightly gate D-48 v2 whose in-repo artifacts are all wired and re-verified here, with the owner's acceptance-by-evidence recorded 2026-10-07. The milestone-close human-acceptance dimension is tracked by the orchestrator's UAT session. The one judgment-tier residue worth a glance during that session (non-blocking): the semantic focus-lag property of the exhaustive focused-inputs modes after the 04-08 refactor — structural evidence says untouched, and ordinary nightly use continues to exercise it.

### Gaps Summary

None. No truth FAILED, no artifact missing/stub (the single tool flag — README.ru.md — is the documented russian-first rename, deliverable present as README.md + README.en.md), no key link unwired, no blocker anti-pattern, zero debt markers. The tree evolution since the previous verification (internal/ move, README restructure, matrix v3→v4, nightly gate) is coherent later-phase work that preserves every phase-4 deliverable; the fresh-session gate now exists in human-free machine-checked form per the owner's own directive.

Recorded notes:
- The latency half of INST-03 remains waived by owner decision (в), WINDOWS #5 closed — the honest numbers and budget-fail status are still published in both READMEs (re-verified today).
- docs/ACCEPTANCE.md checklist result fields remain unfilled (advisory above); the owner's recorded verdicts live in 04-UAT.md and planning history.
- The dispatch script and workflow default now target matrix-v4 (35-row superset) — the D-48 gate tracks the live matrix by design (2c41b83); the phase-4-era 31/31 v3 proof (run 35108412175) and docs/ACCEPTANCE.md item-1 criterion predate that growth and remain the recorded v1.0.0 evidence.

---

_Verified: 2026-10-07T20:59:03Z_
_Verifier: Claude (gsd-verifier) — re-verification at milestone close_
