---
phase: 01-adr-paket-i-karkas-ibus-dvizhka
verified: 2026-09-11T08:46:33Z
status: human_needed
score: 20/20 must-haves verified
covered_files:
  - .github/dependabot.yml
  - .github/workflows/pr-sanity.yml
  - .github/workflows/security-scheduled.yml
  - .gitignore
  - .golangci.yml
  - .planning/codebase/CONVENTIONS.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-01-PLAN.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-01-SUMMARY.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-02-PLAN.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-02-SUMMARY.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-03-PLAN.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-03-SUMMARY.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-04-PLAN.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-04-SUMMARY.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-05-PLAN.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-05-SUMMARY.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-REVIEW.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-VALIDATION.md
  - AGENTS.md
  - README.md
  - cmd/goswitchd/main.go
  - dist/systemd/user/goswitchd.service
  - docs/SPEC.md
  - docs/adr/ADR-001-layout-switching-mechanism.md
  - docs/adr/ADR-002-tap-semantics.md
  - docs/adr/ADR-003-replacement-capability-ladder.md
  - docs/adr/ADR-004-buffer-reset-triggers.md
  - docs/adr/ADR-005-macr-01-owner-checkpoint.md
  - docs/adr/d01-experiment-log.md
  - engine/address.go
  - engine/address_test.go
  - engine/conn.go
  - engine/engine.go
  - engine/engine_test.go
  - engine/factory.go
  - engine/keys.go
  - engine/types.go
  - engine/wire_test.go
  - go.mod
  - go.sum
  - internal/hotkey/fsm.go
  - internal/hotkey/fsm_test.go
  - internal/logging/logging.go
  - internal/logging/logging_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - layouts/generator/main.go
  - layouts/tables.go
  - layouts/tables_test.go
  - mise.toml
  - test/e2e/README.md
  - test/e2e/case_d01.go
  - test/e2e/case_m1.go
  - test/e2e/case_resilience.go
  - test/e2e/focus_helper.py
  - test/e2e/main.go
  - test/e2e/preflight.go
covered_digest: "v1:sha256:e9224cb553af4d9e311e2e37f3f8bd0cfd882d11219652a0b19cb73562fe73af"
behavior_unverified: 0
overrides_applied: 0
human_verification:
  - test: "Live keyd-session coexistence check (INTEG-02 owner checklist in test/e2e/README.md § INTEG-02)"
    expected: "With keyd active and goswitch-en the active source: normal typing transits, keyd remaps still apply, no double keys, goswitch FSM decisions still fire"
    why_human: "keyd.service is failed/disabled on this machine and starting it requires root — cannot be automated; phase documents this honestly as «архитектурно обоснованной, но не подтверждённой живьём»"
  - test: "Push the phase branch / open the first PR and observe pr-sanity on a real GitHub runner"
    expected: "pr-sanity goes green end-to-end (mise install of pinned tools, build/vet/lint/test -race/tidy-diff, govulncheck job)"
    why_human: "GitHub-runner behavior (mise.run availability, mise install of pinned go/golangci-lint) is external-service integration provable only on a real push; branch gsd/phase-01-… has never been pushed (verified: no origin counterpart)"
  - test: "Wait for the first weekly dependabot gomod PR and merge it through pr-sanity"
    expected: "Dependabot opens a gomod bump PR; full pr-sanity gate runs on it automatically"
    why_human: "Requires GitHub-side scheduling; locally only the YAML structure and the gomod ecosystem entry are provable"
  - test: "Observe the first scheduled security-scheduled run (weekly cron, Mon 06:00 UTC)"
    expected: "Workflow fires and govulncheck exits 0 (no reachable vulnerabilities)"
    why_human: "Requires GitHub-side cron activation; govulncheck is network-side by design and not mirrored locally"
---

# Phase 01: ADR-пакет и каркас IBus-движка — Verification Report

**Phase Goal (ROADMAP.md):** «Все архитектурные развилки закрыты утверждёнными ADR, а `goswitchd` работает как IBus engine: видит клавиатурные события как активный input source, логирует их, переживает рестарт ibus-daemon и собственные паники, не конфликтуя с keyd/xremap — фундамент, на котором безопасно строить коррекцию и переключение.»
**Verified:** 2026-09-11T08:46:33Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Verification Approach Note (MVP mode discrepancy)

ROADMAP.md declares `**Mode:** mvp` for Phase 1, but the goal is an
architectural-capability statement, not a User Story
(«As a …, I want to …, so that …»). Per
`gsd-core/references/verify-mvp-mode.md` the MVP user-flow framing fires only
when BOTH `mode: mvp` AND a user-story goal are present (the convention is
"set by `/gsd mvp-phase` per Phase 2"). Standard goal-backward verification
against the 5 Success Criteria + plan must_haves was therefore applied.
**Discrepancy surfaced for the owner:** run `/gsd mvp-phase 1` if a
User-Story goal is wanted; no action required otherwise.

## Live-gate evidence policy

Live e2e cases (`mise run e2e-m1`, `e2e-ibus-restart`, `e2e-kill9-survive`,
`e2e-d01`) drive the owner's live desktop and were NOT re-run by the verifier.
Per the phase's own coordination contract, live-gated must-haves are judged on
**committed evidence**: the appended-by-tool experiment journal
`docs/adr/d01-experiment-log.md` (git-committed, timestamped, observable-based),
the SUMMARY coverage records of two independent executor agents, the
01-VALIDATION.md per-task map, and direct code inspection of the asserting
e2e cases (which this verifier read in full — the assertions are specific and
non-vacuous). Headless gates (`mise run ci` = build+vet+lint+test -race,
`go generate` determinism) WERE re-run by the verifier and are green.

## Goal Achievement

### Observable Truths

Consolidated from the 5 ROADMAP Success Criteria (primary contract) merged
with the plans' must_haves (42 plan-level truths deduplicate into these).

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC-1: ADR-001..005 exist in unified format (Status/Context/Decision/Consequences/Reversibility), CONTEXT decisions D-01..D-06 quoted | ✓ VERIFIED | All 5 files read; section headers confirmed at lines 3-98 of each; D-01 quoted verbatim in ADR-001 Context |
| 2 | SC-1: D-01 experiment journal — ≥4 verdict rows, decision derivable | ✓ VERIFIED | `grep -c "Verdict\|Вердикт" docs/adr/d01-experiment-log.md` = **15** (10 probe verdicts + 2 DECISION + 3 auxiliary, two full runs 00:34–00:46); every row carries command + observable + verdict; DECISION: Option B derived in-journal |
| 3 | SC-1: ADR-001 written last from journal — winner (or Option B + kill-criteria), losers with observed failures, D-03 consequences | ✓ VERIFIED | ADR-001 cites journal at :36; Decision = «Побеждает вариант B — внутренний флип»; kill-criteria table lists per-candidate observable failures (incl. the keyboard.js 46 finding); Consequences carries D-03 |
| 4 | SC-1: M0 owner gate passed; MACR-01 mechanism fixed in ADR-005 (D-12: v1, inside goswitch); spec-deltas applied | ✓ VERIFIED | All 5 ADRs Status=Accepted with «утверждено» recorded verbatim (ADR-005 Status); SPEC §5 (line 111) carries the ADR-002 NFR delta; §4.3 carries the ADR-004 delta (mouse click removed, mandatory surrounding-text check); unique §4.4/§4.5 headings; **no MACR-01 code exists** (grep over cmd/ internal/ engine/ = NONE) |
| 5 | SC-2: goswitchd registers programmatically on the private IBus bus (no XML, no root), two engines goswitch-en/-ru (INTEG-01) | ✓ VERIFIED | engine/conn.go serve(): Dial→Auth(EXTERNAL/SO_PEERCRED)→Hello→RequestName guard→Factory export→RegisterComponent; committed live evidence: SUMMARY 01-01 D2 — ListActiveEngines contains both engines with correct descs, SIGTERM exit 0 (`ibus list-engine` adaptation documented as deviation — XML-registry-only on 1.5.29) |
| 6 | SC-2: as active input source the engine receives ProcessKeyEvent and logs key events; normal typing transits | ✓ VERIFIED | engine/engine.go:128-142 returns `false, nil` unconditionally (observer by construction); TestEngine_ProcessKeyEventReturnsFalse green under -race (re-run by verifier); key trace at slog.Debug only; committed live evidence: 12 key records for injected ghbdtn (keycodes 34/35/48), m1-gate PASS |
| 7 | SC-2: e2e skeleton — ydotool injection, self-activation of target surface, fail-fast preflight, exit-code contract (TEST-02/03) | ✓ VERIFIED | test/e2e/main.go `os.Exit(run())`; 6 named preflight checks (injection self-test, ibus address, uinput writable, /usr/bin/python3 gi, ListActiveEngines registration, log heartbeat); witness-gated zenity/shell dual-path surface (documented deviation: AT-SPI grabFocus refused under Wayland); /usr/bin/python3 absolute path; committed m1-gate PASS |
| 8 | SC-3: ibus restart → re-registration on the new socket, keys flow again (INTEG-04, ibus#2910) | ✓ VERIFIED | engine/conn.go Run/serve: fresh `Discover()` per cycle (address never cached), bus-loss = signal-channel close (godbus semantics confirmed by reviewer), 1-2 s jittered backoff; case_resilience.go:34-53 asserts literal `re-registered` + new `"msg":"key"` events; committed e2e-ibus-restart PASS (re-registered generation=1) |
| 9 | SC-3: recover shim contains handler panics — desktop input never lost (INTEG-05) | ✓ VERIFIED | `recoverHandler` deferred on all 21 exported D-Bus handlers (counted in engine.go); **TestEngine_RecoverContainsPanic green under -race — behavioral test re-run by the verifier** |
| 10 | SC-3: kill -9 → desktop input alive through a plain engine, daemon respawns and re-registers | ✓ VERIFIED | case_resilience.go kill9 path: SIGKILL → desktop-typing oracle via readback char counts → respawnAndWait waits registrations+1 `component registered`; dist systemd unit `Restart=on-failure`; committed e2e-kill9-survive PASS |
| 11 | SC-4: no EVIOCGRAB / evdev / uinput capture in product code | ✓ VERIFIED | Structural grep over engine/ cmd/ internal/ layouts/: **CLEAN** — no EVIOCGRAB, no /dev/input, no os.OpenFile on devices; injection lives only in the stand (test/e2e) |
| 12 | SC-4: daemon works over the default Ubuntu 24.04 GNOME Wayland IM stack without root (INTEG-03) | ✓ VERIFIED | IBus socket auth EXTERNAL via SO_PEERCRED (conn.go userAuth); systemd **user** unit (PartOf=graphical-session.target); mise user-local; committed live runs under the session user |
| 13 | SC-4: keyd/xremap non-conflict (INTEG-02) | ✓ VERIFIED (architectural) | No-capture structural gate + ProcessKeyEvent false + live transit proven through the engine; honest doc contract in test/e2e/README.md § INTEG-02 (4-step owner checklist; live keyd session remains a **human item** — keyd needs root and is disabled here) |
| 14 | SC-5: golden tests of generated tables incl. `[ ] ; ' , . /` + asymmetric pairs &↔?, @↔", #↔№, $↔;, ^↔:, \|↔/ (CORR-08) | ✓ VERIFIED | TestGolden_SpecExamples/Punctuation/TableSize green under -race (re-run); corpus pins all three registers + 16 punctuation pairs; tables.go carries the pairs (verified lines 16-137) |
| 15 | SC-5: tables generated by `go:generate` from xkb, deterministic, CI xkb-free | ✓ VERIFIED | `// Code generated … DO NOT EDIT.` first line + `//go:generate`; **verifier re-ran `go generate ./layouts && git diff --exit-code` → byte-identical**; generator is dev-side (package main, not imported by tests) |
| 16 | SC-5: FSM unit corpus on synthetic streams — window edges 299/300/301, modifier use, intervening key, 4th tap, FocusOut, glitch, burst of 10 (TEST-01) | ✓ VERIFIED | 11 TestFSM_* functions enumerated and green under -race; purity confirmed: no goroutines/channels/time.Now in fsm.go; decisions fire only at TimerExpired |
| 17 | SC-5: structured logs with levels; key trace only behind -debug (INST-04) | ✓ VERIFIED | internal/logging.Setup LevelVar INFO/DEBUG; TestLogging_JSONShape/LevelFiltering/KeyTraceGated green; -debug flag carries an explicit password warning (main.go:21); README § «Режим отладки и конфиденциальность» |
| 18 | SC-5: headless CI gates green (build/vet/lint/test -race/tidy-diff) | ✓ VERIFIED | **verifier re-ran `timeout 300 mise run ci` → exit 0**: build ok, vet ok, lint 0 issues, test -race ok (5 packages), tidy-diff clean |
| 19 | Plan 01-05: CI triplet (pr-sanity + dependabot gomod weekly + scheduled govulncheck), e2e excluded from CI | ✓ VERIFIED | Workflow contents confirmed (mise install, all gates, separate govulncheck job, `contents: read`, concurrency; the only "e2e" mention is the comment explaining deliberate exclusion); dependabot gomod weekly; cron `0 6 * * 1`; **GitHub-side activation pending = human items** (branch never pushed) |
| 20 | Plan 01-03: session actor serializes FSM, logs `{"msg":"action","n":N}` at expiry, wired into daemon | ✓ VERIFIED | internal/session/actor.go (mutex + AfterFunc + exact log shape); wiring `session.NewActor(hotkey.DefaultWindow)` in cmd/goswitchd/main.go:53; 5 TestActor_* green under -race incl. TestActor_WindowTimerFiresAutomatically and TestActor_Serialization |

**Score:** 20/20 truths verified (0 present-but-behavior-unverified)

Behavior-dependent truths note: recover containment (#9), FSM window
semantics (#16), actor expiry logging (#20) are covered by passing behavioral
tests re-run by the verifier. Live-bus truths (#5-#8, #10) rest on committed
live-run evidence as declared above; the asserting e2e cases were read and
their assertions are specific (literal log substrings, count baselines,
timestamp ordering, readback char counts).

### Required Artifacts

`gsd_run query verify.artifacts` per plan: **31/31 passed, 0 issues** across
01-01 (8/8), 01-02 (5/5), 01-03 (6/6), 01-04 (7/7), 01-05 (5/5). All
artifacts exist, are substantive (line counts verified: engine 700+ LOC total,
generator 466, e2e suite multi-file), and none carry stub markers.

### Key Link Verification

`gsd_run query verify.key-links` per plan: **14/14 verified** ("Pattern found
in source/target") across all five plans — including the daemon→actor wiring,
e2e→daemon subprocess path, ADR-001→journal citation, ADR-004→SPEC delta,
pr-sanity→mise.toml single-version-source, README→systemd unit, and
AGENTS.md→CONVENTIONS.md regeneration chain.

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| engine/conn.go | addr | `Discover()` per cycle: $IBUS_ADDRESS → newest ~/.config/ibus/bus file | ✓ real socket discovery (TestAddress_* green; live-registered) | ✓ FLOWING |
| engine/engine.go | ev (EngineEvent) | decodeEvent of ProcessKeyEvent args | ✓ live: 12 key records for injected ghbdtn (committed) | ✓ FLOWING |
| internal/session/actor.go | action n | FSM Feed at timer expiry | ✓ TestActor_* + live `{"msg":"action","n":2}` in m1-gate | ✓ FLOWING |
| layouts/tables.go | ENToRU/RUToEN | generated from system xkb symbols (us basic + ru winkeys) | ✓ regeneration byte-identical on this machine | ✓ FLOWING |
| docs/adr/d01-experiment-log.md | verdict rows | appended by `test/e2e -case d01-probe` live runs | ✓ 15 rows with timestamps + observables | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full headless gate (build+vet+lint+race+tidy-diff) | `timeout 300 mise run ci` | exit 0; lint 0 issues; 5 pkgs ok | ✓ PASS |
| Test corpus enumeration | `timeout 120 go test ./... -list '.*'` | 30 test functions across 5 packages | ✓ PASS |
| Panic containment (INTEG-05) | TestEngine_RecoverContainsPanic (in ci run) | green under -race | ✓ PASS |
| Observer transit (INTEG-02 core) | TestEngine_ProcessKeyEventReturnsFalse (in ci run) | green under -race | ✓ PASS |
| Generation determinism (CORR-08) | `go generate ./layouts && git diff --exit-code` | byte-identical | ✓ PASS |
| FSM purity | `grep -E "go func\|chan\|time.Now\|Sleep" internal/hotkey/fsm.go` | no matches | ✓ PASS |
| Journal verdict count (D-01) | `grep -c "Verdict\|Вердикт" docs/adr/d01-experiment-log.md` | 15 (≥4 required) | ✓ PASS |
| No evdev capture (INTEG-02 gate) | `grep -riE "EVIOCGRAB\|uinput\|/dev/input\|evdev" engine/ cmd/ internal/ layouts/` | no matches | ✓ PASS |
| Journal not oracle-limited (WR-03 taint check) | `grep -c "length-only" docs/adr/d01-experiment-log.md` | 0 — all D-01 rows had a working text oracle | ✓ PASS |
| Phase commits exist | `git cat-file -t` × 16 hashes from SUMMARYs | all OK | ✓ PASS |
| Live e2e (m1/ibus-restart/kill9/d01) | not re-run (drives owner's desktop) | committed evidence cited in truths #5-#10 | ? SKIP (by policy) |

### Probe Execution

Step 7c: no `scripts/*/tests/probe-*.sh` conventions in this repo; the
phase's runnable probes are the mise e2e tasks (live-session, excluded by
policy above) and the CI gate (re-run, PASS). Not applicable beyond that.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| CORR-08 | 01-02 | xkb-generated ЙЦУКЕН↔QWERTY tables incl. punctuation, go:generate | ✓ SATISFIED | truths #14-15; REQUIREMENTS traceability Complete |
| INTEG-01 | 01-01 | IBus engine registration, commit_text injection path | ✓ SATISFIED | truth #5; CommitText emitter shipped |
| INTEG-02 | 01-01, 01-03 | keyd/xremap compat, no EVIOCGRAB | ✓ SATISFIED (architectural + live transit) | truths #6, #11, #13; live keyd checklist = human item |
| INTEG-03 | 01-01, 01-03, 01-05 | default IM stack, no root | ✓ SATISFIED | truths #11-12; GitHub CI side = human item |
| INTEG-04 | 01-01, 01-03, 01-04 | re-registration after ibus restart | ✓ SATISFIED | truths #5, #8 |
| INTEG-05 | 01-01, 01-03 | panic does not kill engine/input | ✓ SATISFIED | truths #9-10 |
| TEST-01 | 01-02, 01-03, 01-05 | headless unit/golden corpus | ✓ SATISFIED | truths #16-18 |
| TEST-02 | 01-03, 01-04 | e2e stand: physical injection + readback | ✓ SATISFIED | truth #7; d01 journal proves the stand live |
| TEST-03 | 01-03 | stand self-activates target window | ✓ SATISFIED | truth #7 (witness-gated zenity/shell path) |
| INST-04 | 01-01, 01-05 | structured logs, levels, debug key trace | ✓ SATISFIED | truths #17, #19; README privacy section |

Orphaned requirements: **none** — REQUIREMENTS.md traceability maps exactly
these 10 IDs to Phase 1, all marked Complete, all claimed by plans.

### Anti-Patterns Found

Debt-marker gate: **ZERO** TBD/FIXME/XXX across all phase Go/Python/MD/YML/
TOML/service files. Zero TODO/HACK/PLACEHOLDER in code. No empty-return stub
patterns in product code (CommitText is a documented functional primitive for
Phase 2, present and wired to a real Emit).

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| engine/conn.go | 119-123 | WR-01: first successful registration after a failed attempt logs `re-registered` (generation counts attempts, not registrations) | ⚠️ Warning | Observability-contract drift + spurious-flake risk in preflight/checkEngineRegistered on the systemd boot path (ibus not yet up); does not affect the demonstrated paths (live runs had attempt-0 success) |
| engine/conn.go | 88-101 | WR-02: `NameFlagReplaceExisting` is inert without incumbent `AllowReplacement`; rapid respawn after SIGKILL can fatal-exit on stale `Exists` | ⚠️ Warning | Narrow race in the kill9/respawn path; canonical run passed; systemd would mask it with a retry; hardening candidate for Phase 2 |
| test/e2e/case_d01.go | 309-401 | WR-03: on locked sessions the length-only oracle can never report `switched`, DECISION row doesn't state the limitation | ⚠️ Warning | Tooling limitation for future re-runs only — the committed journal has 0 oracle-limited rows, so the recorded Option B decision is untainted |
| test/e2e/case_d01.go, workflows, misc | — | REVIEW Info findings IN-01..IN-11 (dead store `caps`, modsMask drop of SUPER/HYPER/META bits, unpinned mise.run installer, no github-actions dependabot ecosystem, generator cycle guard, etc.) | ℹ️ Info | Phase-2/3 hardening backlog; none block the phase goal |

**Judgment (per coordination instruction):** none of the three warnings
invalidates a phase must-have. WR-01/WR-02 degrade supported-but-secondary
paths (boot-time retry race, rapid respawn race) that the canonical e2e
matrix demonstrably passed; WR-03 does not taint the committed D-01 verdicts.
The phase goal gates on behavior demonstrated by the e2e matrix and ADR
closure — both hold. All three are quality debt for Phase 2 hardening, in
line with 01-REVIEW.md (0 critical).

### Human Verification Required

1. **Live keyd coexistence (INTEG-02)**
   **Test:** run the 4-step owner checklist in `test/e2e/README.md` § INTEG-02 with keyd active (requires root to start keyd.service — failed/disabled on this machine).
   **Expected:** normal typing transits, keyd remaps still apply, no double keys, goswitch FSM decisions still fire.
   **Why human:** needs root + a live keyd session; the phase honestly documents compat as «архитектурно обоснованной, но не подтверждённой живьём».

2. **First pr-sanity run on a real GitHub runner**
   **Test:** push the phase branch (`gsd/phase-01-adr-paket-i-karkas-ibus-dvizhka` — currently has no origin counterpart) / open the PR.
   **Expected:** pr-sanity green end-to-end (mise install of pinned tools; build/vet/lint/test -race/tidy-diff; govulncheck job).
   **Why human:** GitHub-runner behavior is external-service integration; local `mise run ci` (the same tasks) is green.

3. **First dependabot PR**
   **Test:** wait for the weekly gomod scan; merge a bumped PR through pr-sanity.
   **Expected:** dependabot opens the PR; the full gate runs on it automatically.
   **Why human:** requires GitHub-side scheduling.

4. **First scheduled govulncheck run**
   **Test:** observe the weekly security-scheduled run (Mon 06:00 UTC).
   **Expected:** workflow fires, govulncheck exits 0.
   **Why human:** requires GitHub-side cron; network-side by design.

These are exactly the four items the phase's own 01-VALIDATION.md contract
pre-declared as Manual-Only (environment-gated) — no new gaps were found.

### Gaps Summary

**No gaps.** All 20 consolidated must-have truths verified; 31/31 artifacts
pass; 14/14 key links wired; requirements 10/10 satisfied with zero orphans;
zero debt markers; the three code-review warnings judged non-blocking quality
debt. The `human_needed` status comes solely from the four environment-gated
owner checks above (keyd root requirement + three GitHub-side CI activations),
which the phase validation contract itself scheduled as manual-only.

---

_Verified: 2026-09-11T08:46:33Z_
_Verifier: Claude (gsd-verifier)_
