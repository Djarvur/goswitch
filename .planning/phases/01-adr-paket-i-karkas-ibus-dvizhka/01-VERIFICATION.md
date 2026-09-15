---
phase: 01-adr-paket-i-karkas-ibus-dvizhka
verified: 2026-09-15T17:57:51Z
status: passed
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
covered_digest: "v1:sha256:18c60a0fc7ee44e72c1383e99f3c59aafcb54bc549acab8e1cb606355d973453"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 20/20
  gaps_closed: []
  gaps_remaining: []
  regressions: []
---

# Phase 01: ADR-пакет и каркас IBus-движка — Verification Report

**Phase Goal (ROADMAP.md):** «Все архитектурные развилки закрыты утверждёнными ADR, а `goswitchd` работает как IBus engine: видит клавиатурные события как активный input source, логирует их, переживает рестарт ibus-daemon и собственные паники, не конфликтуя с keyd/xremap — фундамент, на котором безопасно строить коррекцию и переключение.»

**Verified:** 2026-09-15T17:57:51Z
**Status:** passed
**Re-verification:** Yes — fingerprint refresh after Phase 3 modified shared covered files

## Why this re-verification ran

Phase 1 passed verification 2026-09-11 (20/20) and UAT 4/4 (owner-accepted
2026-09-11, 01-UAT.md — final, not re-opened), then re-verified 2026-09-15T03:55:00Z
after Phase 2. Phase 3 execution (7 plans + 8 code-review fix commits
d715937..c23e98c) has since legitimately modified covered files again. Git
classification of every change since the prior `verified:` timestamp
(2026-09-15T03:55:00Z), per file of the covered set:

- **Untouched — regression-check only** (0 commits): engine/{conn,factory,
  address,types}.go, engine/{address,wire}_test.go, internal/logging/*,
  layouts/*, docs/SPEC.md, docs/adr/* (ADR-001..005 + journal),
  dist/systemd/user/goswitchd.service, .github/workflows/{pr-sanity,
  security-scheduled}.yml, .github/dependabot.yml, .golangci.yml, .gitignore,
  README.md, AGENTS.md, test/e2e/{focus_helper.py,case_d01.go}, all five
  Phase 1 PLAN/SUMMARY pairs.
- **Phase 3 rework — deep re-check** (full 3-level + behavioral at HEAD
  c23e98c): internal/session/{actor.go,actor_test.go} (21 commits; WR-01/
  CR-02/WR-05 fixes among them), internal/hotkey/{fsm.go,fsm_test.go}
  (CR-01 configurable tap key), cmd/goswitchd/main.go (config wiring),
  engine/{engine.go,engine_test.go} (anchorPos seam), engine/keys.go
  (+4 keyval constants), test/e2e/{main.go,preflight.go,case_m1.go,
  case_resilience.go,README.md}, mise.toml (new e2e tasks), go.mod/go.sum
  (fsnotify added).

New Phase 3 packages (internal/{config,ctlsvc,correct,appid,clipboard},
cmd/goswitchctl, test/e2e matrix-v2 surface, .github/workflows/e2e-matrix.yml)
are NOT absorbed into covered_files: they carry no primary Phase 1 contract
weight — their only Phase 1 obligation, the no-device-capture prohibition
(SC-4/INTEG-02), is verified structurally over them (truth #11) without
fingerprinting. Conservative by design.

The prior report had no `gaps:` section; its 20 truths, evidence and
live-gate policy carried over as the map.

## Designed successions (documented, not gaps)

Plan-level Phase 1 details superseded by later roadmap-sanctioned work. The
SC-level contracts all still hold on the current tree:

1. **ProcessKeyEvent consumption** (carried from the Phase 2 re-verification,
   extended by Phase 3): plan 01-01 said «returns false for all events». The
   verdict is delegated to the EventHandler (engine.go:148-171); Phase 3 adds
   MACR Super+letter consumption and the configurable-tap path — all through
   the same handler verdict. Nil handler = pure observer by construction;
   declining handler transits everything (`TestEngine_ProcessKeyEventReturnsFalse`,
   re-run green); EN mode is explicitly «Phase 1 semantics — transit»
   (`TestActor_ENTransitUnchanged`, re-run green); bare modifiers,
   non-printables and true combos are never consumed (actor.go:425-433, 1241).
2. **go.mod dependency set**: plan 01-01 said godbus only. Now 3 direct deps
   — godbus v5.2.2, yaml.v3 v3.0.1 (Phase 2, e2e matrix), fsnotify v1.10.1
   (Phase 3, config hot reload) — every one a core technology in the approved
   stack (research/STACK.md). +1 indirect (golang.org/x/sys via godbus).
   Minimal-dep constraint holds.
3. **Actor construction wiring**: prior tree had `session.NewActor(hotkey.
   DefaultWindow)`; main.go now builds the actor from the startup config
   (main.go:114-126) with the watcher attached. The no-config default equals
   `hotkey.DefaultWindow` (300 ms, ADR-002) — pinned by config.TestDefaults;
   actor remains `engine.Config.Handler` (main.go:142).
4. **FSM constructor**: `NewFSM(window)` → `NewFSM(window, tapKeyval)` +
   SetWindow/SetTapKey hot-reload seams (CR-01/CONF-02, D-35). Purity
   preserved (no goroutines/channels/clock in fsm.go); the Phase 1 corpus
   expectations are untouched and still green; new tests extend them.
5. **HandleSurroundingText seam widened** with anchorPos (D-30: a selection
   is active exactly when anchorPos != cursorPos) — decode path and recover
   shim unchanged, wire test strengthened to assert both positions.
6. **MACR-01**: the Phase 1 truth «no MACR-01 code in Phase 1» held at phase
   close (grep-verified then). Phase 3 implemented the MACR layer
   (internal/appid + the actor's interception path) per Accepted ADR-005 —
   the roadmap-sanctioned next step, exactly what ADR-005's mechanism
   decision was for.
7. **engine/keys.go**: +KeyControlR 0xffe4, KeySuperL 0xffeb, KeySuperR
   0xffec, KeyV 0x076 — additive, each verified verbatim against
   ibuskeysyms.h per in-file citations.
8. **mise.toml**: ~20 new e2e tasks (word/phrase/select/combo/ctl/macr/
   matrix-v1/v2); the Phase 1 contracts — [tools] pins (go 1.23,
   golangci-lint 2), tasks build/test/lint/vet/tidy-diff/ci, no Makefile —
   unchanged.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Re-verification evidence (current tree, HEAD c23e98c) |
|---|-------|--------|------------------------------------------|
| 1 | SC-1: ADR-001..005 exist in unified format (Status/Context/Decision/Consequences/Reversibility), CONTEXT decisions quoted | ✓ VERIFIED (regression) | Untouched since prior pass; re-checked: each of the 5 files carries 5/5 section headers, Status = Accepted (гейт M0) |
| 2 | SC-1: D-01 journal — ≥4 verdict rows, decision derivable | ✓ VERIFIED (regression) | Untouched; `grep -c "Verdict\|Вердикт"` = **15** |
| 3 | SC-1: ADR-001 written last from journal — winner/Option B + kill-criteria, losers' observed failures, D-03 consequences | ✓ VERIFIED (regression) | File untouched (git: 0 commits since prior pass) |
| 4 | SC-1: M0 owner gate passed; MACR-01 fixed in ADR-005 (D-12), spec-deltas applied; no MACR-01 code in Phase 1 | ✓ VERIFIED (regression + succession) | All 5 ADRs Accepted, SPEC.md untouched — the owner-gate contract intact. Phase 3 now ships MACR code per Accepted ADR-005 (designed succession §6); the truth held at phase close and its decision-record contract still holds |
| 5 | SC-2: programmatic registration on the private IBus bus (no XML, no root), two engines goswitch-en/-ru (INTEG-01) | ✓ VERIFIED | engine/conn.go + factory.go byte-unchanged since their live proof; **fresh machine-generated live evidence**: e2e-matrix CI run 34984814995 (2026-09-15T14:54:29Z, this branch, self-hosted GNOME, head 21df37c) — **21/21 PASS** with a fresh daemon per case; registration + activation re-proven on the Phase 3 lineage; TestAddress_* green at HEAD |
| 6 | SC-2: as active source, engine receives ProcessKeyEvent, logs keys; normal typing transits | ✓ VERIFIED (deep) | engine.go re-read at HEAD: decode → slog.Debug("key",…) (engine.go:160, still DEBUG-only) → handler verdict, nil → false (engine.go:156-171); TestEngine_ProcessKeyEventReturnsFalse, TestActor_ENTransitUnchanged, TestActor_RUTransitUnmapped all green in this pass's `-race` suite at HEAD |
| 7 | SC-2: e2e skeleton — ydotool injection, self-activation, fail-fast preflight, exit-code contract (TEST-02/03) | ✓ VERIFIED (deep) | main.go: `os.Exit(run())` (line 82), FAIL paths print + non-zero; preflight keeps all 6 Phase 1 named checks (injection-selftest, ibus-address, uinput-writable, python-gi, engine-registered, daemon-log-heartbeat) + surface checks; focus_helper.py byte-unchanged (/usr/bin/python3 + AT-SPI grabFocus); m1 assertions intact (`"msg":"key"` ≥ min, `"msg":"action","n":2`, registered-before-focus_in by timestamps, case_m1.go:63-89, 322-334) |
| 8 | SC-3: ibus restart → re-registration on the new socket, keys flow again (INTEG-04) | ✓ VERIFIED | conn.go (the reconnect implementation: fresh Discover per cycle, signal-channel bus-loss, jittered backoff) byte-unchanged since the committed live PASS; case_resilience.go reworked for the multi-surface runner but assertions intact (literal `re-registered` wait line 34 + respawn registrations+1); mise task e2e-ibus-restart present |
| 9 | SC-3: recover shim contains handler panics (INTEG-05) | ✓ VERIFIED (deep) | Re-counted at HEAD: `defer recoverHandler` on **21/21** exported Engine handler methods (engine.go — the 25 exported methods are 21 D-Bus handlers + 4 outgoing emitters; engine.go grew NO new exported handler since the prior pass) + CreateEngine (factory.go:30); **TestEngine_RecoverContainsPanic re-run green under -race at HEAD this pass** |
| 10 | SC-3: kill -9 → desktop input alive, daemon respawns and re-registers | ✓ VERIFIED | case_resilience.go kill9 path intact (SIGKILL → plain-engine switch → inject → desktop oracle → respawnAndWait `component registered` count+1, lines 60-130); dist unit `Restart=on-failure` unchanged; committed live PASS stands |
| 11 | SC-4: no EVIOCGRAB / evdev / uinput capture in product code | ✓ VERIFIED (deep, widened) | Structural grep over engine/ cmd/ internal/ layouts/ — now including Phase 3's internal/{appid,clipboard,config,correct,ctlsvc} and cmd/goswitchctl — **CLEAN**: no EVIOCGRAB, no /dev/uinput, no /dev/input, no evdev access; no OpenFile/ioctl in product code; the only "evdev" matches are comments describing physical keycode semantics for ForwardKeyEvent bursts (actor.go:858,1132,1143) |
| 12 | SC-4: works over default Ubuntu 24.04 GNOME Wayland IM stack without root (INTEG-03) | ✓ VERIFIED (regression) | systemd **user** unit unchanged (PartOf=graphical-session.target, Restart=on-failure — no root anywhere); conn.go EXTERNAL/SO_PEERCRED auth unchanged; mise user-local |
| 13 | SC-4: keyd/xremap non-conflict (INTEG-02) | ✓ VERIFIED | No-capture gate (truth 11) + transit (truth 6) + e2e README INTEG-02 owner checklist intact (test/e2e/README.md:196-219); live keyd session was owner-UAT-accepted 2026-09-11 |
| 14 | SC-5: golden tests of generated tables incl. `[ ] ; ' , . /` + asymmetric pairs (CORR-08) | ✓ VERIFIED (regression) | layouts/ untouched; TestGolden_{SpecExamples,Punctuation,TableSize} green in this pass's suite at HEAD |
| 15 | SC-5: tables generated by go:generate, deterministic, CI xkb-free | ✓ VERIFIED (re-run) | **Re-run this pass: `go generate ./layouts && git diff --exit-code` → byte-identical**; DO-NOT-EDIT marker + go:generate header intact |
| 16 | SC-5: FSM unit corpus on synthetic streams (TEST-01) | ✓ VERIFIED (deep) | fsm.go reworked by CR-01 (configurable tap key) — purity re-checked clean (no goroutines/chans/time.Now/Sleep); corpus now 15 tests: the 11 Phase 1 TestFSM_* preserved green + TestFSM_{ConfigurableTapKey,SetTapKey}, TestParseBinding*, TestFamilyMask; all green under -race at HEAD |
| 17 | SC-5: structured logs with levels; key trace only behind -debug (INST-04) | ✓ VERIFIED | logging.go untouched (LevelVar INFO/DEBUG); key trace still slog.Debug-only (engine.go:160; surrounding_text/capabilities also DEBUG, lifecycle INFO — engine.go:192,237,248); TestLogging_JSONShape/LevelFiltering/KeyTraceGated green; -debug password warning in main.go:26; README «Режим отладки и конфиденциальность» (README.md:13, untouched) |
| 18 | SC-5: headless CI gates green (build/vet/lint/test -race/tidy-diff) | ✓ VERIFIED (re-run) | **Verifier re-ran all gates this pass at HEAD**: `go build ./...` OK; `go vet ./...` OK; `go test -race -count=1 ./...` — 11 test-bearing packages ok, 0 failures (incl. engine, hotkey, session, layouts, logging + Phase 3 packages); `mise run lint` → 0 issues; `go mod tidy` + `git diff --exit-code` → zero drift |
| 19 | Plan 01-05: CI triplet (pr-sanity + dependabot gomod weekly + scheduled govulncheck), e2e excluded from CI | ✓ VERIFIED | All three workflow/config files untouched; security-scheduled cron actually fired green 2026-09-14 (run 34838604697 — weekly cadence, no newer scheduled run exists yet); dependabot dynamic runs green; e2e remains CI-excluded except the dispatch-only e2e-matrix workflow; **no pr-sanity run on this branch yet** (Phase 3 worked without a PR; triggers are push-to-main/PR/dispatch) — the headless substitute is this pass's verifier-local gate re-run at HEAD (truth 18), and the live substitute is run 34984814995 |
| 20 | Plan 01-03: session actor serializes FSM, logs `{"msg":"action","n":N}` at expiry, wired into daemon | ✓ VERIFIED (deep) | Wiring intact through the Phase 3 config rework: `session.NewActor(window)` at cmd/goswitchd/main.go:116, actor = `engine.Config.Handler` at main.go:142; HandleKey holds `a.mu` across the whole event (actor.go:434-437) — WR-01's rungMu only moves the clipboard rung (a Phase 3 feature) off the actor mutex with busy-skip, the FSM event path stays serialized; ExpiryAt still emits `slog.Info("action", "n", int(action))` (actor.go:555); **TestActor_Serialization + TestActor_ENTransitUnchanged re-run green under -race at HEAD**; 62 TestActor_* tests in the suite |

**Score:** 20/20 truths verified (0 present-but-behavior-unverified)

Behavior-dependent truths note: panic containment (#9), FSM window
semantics (#16), actor expiry/timer/serialization (#20), transit contract
(#6) are all covered by behavioral tests that ran green in this pass's
`go test -race ./...` at HEAD (the four named tests above additionally
re-run individually). Live-bus truths (#5, #8, #10) rest on committed
live-run evidence per the phase's binding evidence policy — #5 additionally
re-proven live by the 21/21 e2e-matrix CI run 34984814995 (2026-09-15,
after the prior verified timestamp, head 21df37c); #8/#10 core
implementation (conn.go) is byte-unchanged since their live proof and their
asserting e2e cases survive the Phase 3 runner refactor with assertions
intact. Recency caveat, stated honestly: run 34984814995 executed at
21df37c — 12 commits before HEAD (docs + the 8 code-review fixes); those
fixes' live-path changes (actor/fsm/matrix runner) are covered by the
HEAD-green unit suite, not by a newer live matrix run (none exists).

### Advisory (New Scope, Unevidenced)

Re-verification ran; the Step 7 scan over all covered files found no new
blocking pattern. Advisory entries: none.

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| 1 | None — only the informational notes below (lint tooling notice; doc wording) | — | — |

### Required Artifacts

`gsd_run query verify.artifacts` per plan, re-run this pass: **31/31 passed,
0 issues** — 01-01 (8/8), 01-02 (5/5), 01-03 (6/6), 01-04 (7/7), 01-05 (5/5).

### Key Link Verification

`gsd_run query verify.key-links` per plan, re-run this pass: **14/14
verified** across all five plans — including main.go→engine.Run
(cmd/goswitchd/main.go:66), e2e→daemon subprocess, ADR-001→journal,
ADR-004→SPEC delta, pr-sanity→mise.toml, README→systemd unit,
AGENTS.md→CONVENTIONS.md regeneration chain.

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| engine/conn.go (unchanged) | addr | `Discover()` per cycle: $IBUS_ADDRESS → suffix-filtered freshest ~/.config/ibus/bus file | ✓ real socket discovery (TestAddress_* green; live-registered in the 21/21 matrix run) | ✓ FLOWING |
| engine/engine.go (Phase 3 seam widening) | ev / consume | decodeEvent of ProcessKeyEvent args; verdict from EventHandler | ✓ matrix cases prove keys flow and corrections/MACR decisions act on the live bus | ✓ FLOWING |
| internal/session/actor.go (Phase 3 rework) | action n / corrections / snapshot | FSM Feed at timer expiry; one config Snapshot per event (actor.go:436) | ✓ 62 TestActor_* green at HEAD + live matrix corrections | ✓ FLOWING |
| layouts/tables.go (unchanged) | ENToRU/RUToEN | generated from system xkb | ✓ regeneration byte-identical this pass | ✓ FLOWING |
| docs/adr/d01-experiment-log.md (unchanged) | verdict rows | appended by d01-probe live runs | ✓ 15 rows; append path intact in case_d01.go (byte-unchanged) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Build | `go build ./...` | exit 0 | ✓ PASS |
| Vet | `go vet ./...` | exit 0 | ✓ PASS |
| Full race suite at HEAD (single run) | `go test -race -count=1 ./...` | 11 test-bearing packages ok, 0 failures | ✓ PASS |
| Named: panic containment | `go test -race -run TestEngine_RecoverContainsPanic ./engine/` | PASS | ✓ PASS |
| Named: EN transit unchanged | `go test -race -run 'TestActor_ENTransitUnchanged\|TestActor_Serialization' ./internal/session/` | both PASS | ✓ PASS |
| Lint (strict v2 config) | `mise run lint` | 0 issues | ✓ PASS |
| Tidy drift | `go mod tidy` + `git diff --exit-code -- go.mod go.sum` | zero drift | ✓ PASS |
| Generation determinism | `go generate ./layouts && git diff --exit-code -- layouts/` | byte-identical | ✓ PASS |
| Recover-shim count | `grep -c "defer recoverHandler" engine/engine.go` (21) + factory.go (1) | 21 handlers + CreateEngine; 0 new exported handlers since prior pass | ✓ PASS |
| No device capture (widened) | `grep -rniE "EVIOCGRAB\|/dev/uinput\|/dev/input\|evdev" engine/ cmd/ internal/ layouts/` | comment-only matches (keycode semantics); no access | ✓ PASS |
| FSM purity | `grep -nE "go func\|chan \|time.Now\|Sleep" internal/hotkey/fsm.go` | no matches | ✓ PASS |
| Journal verdict count | `grep -c "Verdict\|Вердикт" docs/adr/d01-experiment-log.md` | 15 (≥4) | ✓ PASS |
| CI green | `gh run list` | e2e-matrix 34984814995 success (this branch, 2026-09-15T14:54Z); security-scheduled 34838604697 success (cron 2026-09-14); dependabot dynamic green | ✓ PASS |
| Live e2e matrix v2 artifact | `gh run download 34984814995 -n e2e-report` | 21/21 PASS (phrase/select/layout/combo/reload/ctl/macr families), 2026-09-15T17:55+03:00 | ✓ PASS |
| Live m1/ibus-restart/kill9/d01 mise tasks | not re-run (drive owner's desktop) | committed Phase 1 evidence + intact assertions + unchanged conn.go | ? SKIP (by policy) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` conventions in this repo. The runnable probes
are the mise tasks: headless CI gates re-run this pass (PASS, table above);
live-session e2e tasks excluded by the binding evidence policy — their
machine-generated CI counterpart (e2e-matrix v2, dispatch-only) was fetched
and is green 21/21 at head 21df37c.

### Requirements Coverage

All 10 Phase 1 requirement IDs remain mapped to Phase 1 and Complete in
REQUIREMENTS.md traceability; no orphans.

| Requirement | Source Plan | Status | Evidence |
|-------------|------------|--------|----------|
| CORR-08 | 01-02 | ✓ SATISFIED | truths #14-15 |
| INTEG-01 | 01-01 | ✓ SATISFIED | truth #5; emitters live-driven (matrix corrections) |
| INTEG-02 | 01-01, 01-03 | ✓ SATISFIED | truths #6, #11, #13; live keyd owner-UAT-accepted |
| INTEG-03 | 01-01, 01-03, 01-05 | ✓ SATISFIED | truths #11-12; CI live on GitHub runners |
| INTEG-04 | 01-01, 01-03, 01-04 | ✓ SATISFIED | truth #8 |
| INTEG-05 | 01-01, 01-03 | ✓ SATISFIED | truths #9-10 |
| TEST-01 | 01-02, 01-03, 01-05 | ✓ SATISFIED | truths #16-18 |
| TEST-02 | 01-03, 01-04 | ✓ SATISFIED | truth #7; stand re-proven live by the 21/21 matrix run |
| TEST-03 | 01-03 | ✓ SATISFIED | truth #7 (focus_helper.py contract intact) |
| INST-04 | 01-01, 01-05 | ✓ SATISFIED | truths #17, #19 |

### Decision Coverage

`gsd_run query check.decision-coverage-verify`: **12/12 trackable
CONTEXT.md decisions honored** by shipped artifacts; `not_honored: []`.
Warning-only gate; no status impact.

### Anti-Patterns Found

Debt-marker gate re-run over all covered Go/Python/YAML/TOML files:
**ZERO TBD/FIXME/XXX, zero TODO/HACK/PLACEHOLDER, zero stub patterns**
in product code. Carry-forward warnings from the prior pass (files
unchanged, still non-blocking quality debt): WR-01* (prior pass numbering)
re-registered generation semantics (conn.go:119-123), NameFlagReplaceExisting
inertness (conn.go:88-101), length-only oracle limitation in locked-session
d01 re-runs. Note: the Phase 3 code-review WR-01..WR-05/CR-01..CR-03 items
are a different, disjoint numbering series (03-REVIEW-FIX.md) — all fixed
with tests before this verification.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| engine/conn.go | 93, 122 | carry-forward boot-time retry race + rapid-respawn race (unchanged since prior passes) | ⚠️ Warning | Canonical paths demonstrably green (21/21 live matrix); documented quality debt for hardening |
| .golangci.yml → lint output | — | golangci-lint reports `exhaustruct` deprecated since v2.13.0 (replaced by exhaustruct_v5) — informational tooling notice, config still applies, 0 issues | ℹ️ Info | Cosmetic future config rename; no code impact |
| test/e2e/README.md | 201 | INTEG-02 section still carries the Phase 1 wording «`ProcessKeyEvent`, возвращающий `false` (наблюдатель)» — original b982780 text predating the consumption succession; the architectural claim (no evdev layer, keyd below IBus) remains true | ℹ️ Info | Documentation nuance only; a wording refresh matching the conditional-consumption model would help future readers |

### Human Verification Required

None new. The four environment-gated items from the original pass were
resolved by UAT (01-UAT.md: 4/4 pass, owner-accepted 2026-09-11) and are
not re-opened. No ⚠️ PRESENT_BEHAVIOR_UNVERIFIED truths exist on this pass.

### Gaps Summary

**No gaps.** All 20 must-have truths verified on the current (post-Phase-3)
tree at HEAD c23e98c; 31/31 artifacts pass; 14/14 key links wired;
requirements 10/10 with zero orphans; decision coverage 12/12; zero debt
markers; all headless gates re-run green by the verifier at HEAD; fresh
machine-generated live evidence (e2e-matrix 21/21 on the self-hosted GNOME
runner, this branch, 2026-09-15) covers the registration/key-flow/transit
chain on the Phase 3 lineage, with the 12-commit recency gap to HEAD
covered headlessly and stated honestly. Phase 3's rework of shared files
(actor, FSM, engine seam, daemon wiring, e2e runner, go.mod) preserved
every SC-level Phase 1 contract — consumption delegation, recover-shim
coverage, no-device-capture over the widened package set, DEBUG-gated key
trace, actor serialization and action logging — with eight further designed
successions documented above.

---

_Verified: 2026-09-15T17:57:51Z_
_Verifier: Claude (gsd-verifier)_
