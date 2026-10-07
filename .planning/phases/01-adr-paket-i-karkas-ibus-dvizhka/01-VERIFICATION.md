---
phase: 01-adr-paket-i-karkas-ibus-dvizhka
verified: 2026-10-07T07:43:52Z
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
  - go.mod
  - go.sum
  - internal/engine/address.go
  - internal/engine/address_test.go
  - internal/engine/conn.go
  - internal/engine/engine.go
  - internal/engine/engine_test.go
  - internal/engine/factory.go
  - internal/engine/keys.go
  - internal/engine/types.go
  - internal/engine/wire_test.go
  - internal/hotkey/fsm.go
  - internal/hotkey/fsm_test.go
  - internal/layouts/generator/main.go
  - internal/layouts/tables.go
  - internal/layouts/tables_test.go
  - internal/logging/logging.go
  - internal/logging/logging_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - test/e2e/README.md
  - test/e2e/case_d01.go
  - test/e2e/case_m1.go
  - test/e2e/case_resilience.go
  - test/e2e/focus_helper.py
  - test/e2e/main.go
  - test/e2e/preflight.go
covered_digest: "v1:sha256:e3d59ec78d9f2db5022700f4083857d3b1b271c7d9b31398979fc860959d3715"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 20/20
  gaps_closed: []
  gaps_remaining: []
  regressions: []
gaps: []
deferred:
  - truth: "Plan 01-05 key link: README.md documents the unit placement (cp dist/systemd/user/goswitchd.service → ~/.config/systemd/user/)"
    addressed_in: "Phase 4"
    evidence: "Phase 4 goal «Продукт ставится без root по документированной инструкции» + SC-1; Phase 4-01 INST-01 implemented goswitchctl install, which writes the same goswitchd.service into .config/systemd/user (internal/install/install.go:60-61 at the current tree — unitDirRel = \".config/systemd/user\", unitFile = \"goswitchd.service\"); Phase 4-07 and the 08-07 README rewrite (RU full v1.1.0) both document that installer (README:30 «одна команда: goswitchctl install», idempotent, no-root at README:17). The Phase 1 SC-level contract (user unit, no root, truth #12) is untouched; only the plan-level grep pattern (goswitchd\\.service in README) still does not match — verified again this pass."
---

# Phase 01: ADR-пакет и каркас IBus-движка — Verification Report

**Phase Goal (ROADMAP.md):** «Все архитектурные развилки закрыты утверждёнными ADR, а `goswitchd` работает как IBus engine: видит клавиатурные события как активный input source, логирует их, переживает рестарт ibus-daemon и собственные паники, не конфликтуя с keyd/xremap — фундамент, на котором безопасно строить коррекцию и переключение.»

**Verified:** 2026-10-07T07:43:52Z
**Status:** passed
**Re-verification:** Yes — fingerprint-staleness refresh at milestone close (v1.0.0, all 8 phases shipped); `engine/` and `layouts/` moved under `internal/` (commit 5427d7a, SPEC §8 rev 2026-10-07, owner spec-delta D-55)

## Why this re-verification ran

Phase 1 passed its first verification 2026-09-11 (20/20), was owner-UAT-accepted
2026-09-11 (01-UAT.md, 4/4 — final, not re-opened), and was re-verified four
times since (2026-09-15 ×2, 2026-09-16, 2026-09-20 — all 20/20, previous
verdict `passed` with an empty `gaps:` list). Since the prior `verified:`
timestamp (2026-09-20T14:32:52Z) the tree evolved through Phases 5–8 and a
quick-task series; the one change that stales this phase's fingerprints is
**commit 5427d7a (2026-10-07): `engine/` → `internal/engine/`,
`layouts/` → `internal/layouts/`** (git mv, history-preserving; SPEC §8
revision D-55).

Git classification of every covered file since the prior timestamp, at HEAD
1efc6fa:

- **Moved 1:1 by 5427d7a (single commit):** all 12 `engine/` + `layouts/`
  covered files. Product files are **R100 = byte-identical** renames
  (address.go, conn.go, engine.go, factory.go, keys.go, types.go,
  layouts/generator/main.go, layouts/tables.go); the five test files
  (address_test.go, engine_test.go, wire_test.go, tables_test.go) are
  R098/R099 — the diff is **import-block rewrites only** (verified: each diff
  is a new `package *_test`/import header importing
  `github.com/Djarvur/goswitch/internal/{engine,layouts}`; commit message
  documents "rewrite import paths in 20 files; gofmt re-sorts").
- **Untouched since prior pass (0 commits):** docs/adr/* (all 6),
  dist/systemd/user/goswitchd.service, internal/hotkey/*, internal/logging/*,
  .github/{dependabot.yml,workflows/pr-sanity.yml,workflows/security-scheduled.yml},
  AGENTS.md, .planning/codebase/CONVENTIONS.md, all five Phase 1
  PLAN/SUMMARY pairs, 01-REVIEW.md, 01-VALIDATION.md,
  test/e2e/{case_m1.go, focus_helper.py}.
- **Modified by later phases (shared files, later-phase scope, Phase 1
  seams re-checked below):** cmd/goswitchd/main.go (24 commits — menu-v2,
  sound sink, a11y reconciler wiring), internal/session/actor.go (+test;
  correction semantics, sound fold, a11y fold, per-character inversion),
  go.mod/go.sum (sound stack: +jfreymuth/oggvorbis, +jfreymuth/pulse), mise
  (mise.toml — skillgen + dictgen-regen task path updates), README.md
  (rewritten RU v1.1.0, 08-07), docs/SPEC.md (13 revisions incl. D-55),
  .golangci.yml (exclusion paths for the move), .gitignore,
  test/e2e/{README.md, case_d01.go, case_resilience.go, main.go,
  preflight.go}.

Mechanical plan-level `verify.artifacts`/`verify.key-links` re-runs are not
meaningful this pass: the five PLAN files are frozen history whose artifact
paths say `engine/…`/`layouts/…`. Per the re-verification contract, every
moved path was **re-derived against the current tree** and re-checked
manually (existence + substance + wiring), documented per truth below.

**Live-bus evidence status this pass:** the committed live proof of the
registration/key-flow chain (e2e-matrix run 35108412175, 2026-09-16T14:25Z,
31/31 PASS at head 9f2fd75) remains valid for the IBus protocol layer: every
file of `internal/engine/` — conn.go, factory.go, engine.go, address.go,
keys.go, types.go — is **byte-identical (R100) to the live-proven head**.
The daemon shell around it (cmd/goswitchd/main.go, internal/session/) was
legitimately extended by Phases 5–8, each of which landed through this
project's own verified e2e workflow and green CI; Phase 1's SC-level
mechanisms (reconnect loop, re-registration, respawn, recover) live in the
byte-identical engine package.

## Designed successions (documented, not gaps)

1. **Package relocation** — `internal/engine/`, `internal/layouts/`
   (commit 5427d7a; owner spec-delta D-55, SPEC §8 rev 2026-10-07). All
   importers rewritten; `.golangci.yml` gosec exclusions repointed to
   `internal/engine/conn\.go`; mise dictgen-regen task repointed to
   `./internal/layouts` (mise.toml:216); tables.go keeps its DO-NOT-EDIT
   marker + `//go:generate go run ./generator` (relative — path-correct under
   the new location).
2. **go.mod dependency set**: 3 → 5 direct deps (added
   jfreymuth/oggvorbis + jfreymuth/pulse — Phase 8 sound stack, owner
   spec-delta 261006-squ). The Phase 1 core set is untouched: godbus
   v5.2.2, yaml.v3 v3.0.1, fsnotify v1.10.1; golang.org/x/sys indirect.
   Minimal-dep constraint holds per the approved spec-deltas; CI tidy-diff
   green at HEAD (pr-sanity 1efc6fa, 2026-10-06T22:50Z).
3. **Daemon shell growth** — main.go now 580 lines (menu-v2, sound sink,
   a11y reconciler); the Phase 1 wiring seams survive: `engine.Run(ctx,
   engineConfig(actor))` (main.go:88), actor built via `newActor()` →
   `session.NewActor(window)` + `actor.SetVersion(version)` (main.go:487-488),
   `engine.Config{Handler: actor, …}` (main.go:550).
4. **Actor growth** — actor.go now carries correction/sound/a11y folds
   (Phases 3/5/6/8); the Phase 1 contracts are pinned by still-present,
   still-green tests: `a.mu` held across HandleKey (actor.go:495-496),
   `slog.Info("action", "n", int(action))` at expiry (actor.go:1017),
   TestActor_ENTransitUnchanged explicitly "pins the EN mode as the untouched
   Phase 1" transit (actor_test.go:1274-1278).
5. **Items 1–9 of the prior pass's successions** (ProcessKeyEvent consumption
   delegation to the EventHandler; go 1.23 language version + core dep set;
   actor construction from startup config; FSM constructor + hot-reload
   seams; HandleSurroundingText anchorPos seam; MACR-01 by Phase 3 per
   Accepted ADR-005; additive keys.go keyvals; mise.toml task contracts)
   — all re-confirmed on the current tree this pass.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Re-verification evidence (current tree, HEAD 1efc6fa) |
|---|-------|--------|------------------------------------------|
| 1 | SC-1: ADR-001..005 exist in unified format, CONTEXT decisions quoted | ✓ VERIFIED (regression) | docs/adr/* untouched since prior pass; re-counted at new state: each of the 5 files carries 5/5 `## Status/Context/Decision/Consequences/Reversibility` headers; each Status = «Accepted — гейт M0 (ревью владельца, план 01-04 Task 3)» |
| 2 | SC-1: D-01 journal — ≥4 verdict rows, decision derivable | ✓ VERIFIED (regression) | Untouched; `grep -c "Verdict\|Вердикт"` = **15** (re-run this pass) |
| 3 | SC-1: ADR-001 written last from journal — winner/Option B + kill-criteria, losers' failures, D-03 consequences | ✓ VERIFIED (regression) | File untouched (0 commits since prior pass) |
| 4 | SC-1: M0 owner gate passed; MACR-01 in ADR-005 (D-12), spec-deltas applied; no MACR-01 code in Phase 1 | ✓ VERIFIED (regression + succession) | All 5 ADRs Accepted; SPEC.md revised only via owner spec-deltas (13 commits, D-55 series); MACR code ownership with later phases per the Accepted ADR-005 decision chain — decision-record contract intact |
| 5 | SC-2: programmatic registration on the private IBus bus (no XML, no root), two engines goswitch-en/-ru | ✓ VERIFIED | internal/engine/conn.go + factory.go **byte-identical (R100)** to the live-proven head; TestAddress green `-race` this pass; main.go imports `internal/engine` (main.go:29) and runs `engine.Run` (main.go:88); committed live evidence (e2e-matrix 31/31, head 9f2fd75) valid for the protocol layer per the byte-identity above |
| 6 | SC-2: as active source, engine receives ProcessKeyEvent, logs keys; normal typing transits | ✓ VERIFIED | internal/engine/engine.go byte-identical: decode → `slog.Debug("key",…)` (engine.go:166, DEBUG-only) → handler verdict; **TestActor_ENTransitUnchanged green `-race` this pass** (pins Phase 1 EN-transit semantics explicitly); key-flow chain live-proven at the byte-identical engine head |
| 7 | SC-2: e2e skeleton — ydotool injection, self-activation, fail-fast preflight, exit-code contract | ✓ VERIFIED | test/e2e/main.go `os.Exit(run())` (line 101); preflight keeps all 6 Phase 1 named checks (injection-selftest, ibus-address, uinput-writable, python-gi, engine-registered, daemon-log-heartbeat — later phases only appended app-launch checks); case_m1.go untouched (`"msg":"key"` ≥ min at :63, `"msg":"action","n":2` at :70); focus_helper.py untouched (`focus <app>` grabFocus contract intact) |
| 8 | SC-3: ibus restart → re-registration on the new socket, keys flow again | ✓ VERIFIED | conn.go byte-identical (R100); case_resilience.go "re-registered" wait intact (line 56; the two later commits ec2883e made the precondition oracle count-agnostic — an oracle robustness tweak, no Phase 1 semantic change); mise task e2e-ibus-restart present (mise.toml:45) |
| 9 | SC-3: recover shim contains handler panics | ✓ VERIFIED (deep re-count at new path) | `defer recoverHandler` on **21/21** exported Engine D-Bus handlers (internal/engine/engine.go — 26 exported methods = 21 handlers + 5 outgoing emitters: the prior 4 + later-phase `UpdateModeSymbol`, correctly unshimmed as an engine→client signal) + CreateEngine (factory.go); **TestEngine_RecoverContainsPanic green `-race` this pass** |
| 10 | SC-3: kill -9 → desktop input alive, daemon respawns and re-registers | ✓ VERIFIED | case_resilience.go kill9 path intact (runKill9Survive :167, respawnAndWait `registrations+1` :226); dist unit `Restart=on-failure` unchanged (byte-untouched file); committed live PASS stands; engine layer byte-identical |
| 11 | SC-4: no EVIOCGRAB / evdev / uinput capture in product code | ✓ VERIFIED (deep, re-widened at new paths) | Structural grep re-run over internal/{engine,layouts,hotkey,logging,session} + cmd/goswitchd — **CLEAN**: no EVIOCGRAB, no /dev/uinput, no /dev/input access; the only "evdev" matches are comments describing physical keycode semantics (actor.go:1597,1879,1890 — later-phase MACR keycode tables, comment-only) |
| 12 | SC-4: works over default Ubuntu 24.04 GNOME Wayland IM stack without root | ✓ VERIFIED (regression) | systemd **user** unit byte-untouched (PartOf=graphical-session.target :12, Restart=on-failure :20, no User=); conn.go EXTERNAL/SO_PEERCRED auth byte-identical; README:17 still states «без root везде»; mise user-local |
| 13 | SC-4: keyd/xremap non-conflict | ✓ VERIFIED | No-capture gate (truth 11) + EN transit (truth 6, green test) + e2e README INTEG-02 section intact (test/e2e/README.md:215-233, incl. the manual keyd-session owner checklist); live keyd session owner-UAT-accepted 2026-09-11 |
| 14 | SC-5: golden tests of generated tables incl. `[ ] ; ' , . /` + asymmetric pairs | ✓ VERIFIED | internal/layouts/tables.go byte-identical (R100); tables_test.go import-path-only diff; **TestGolden_Punctuation green `-race` this pass** (TestGolden_SpecExamples, TestGolden_TableSize also present) |
| 15 | SC-5: tables generated by go:generate, deterministic, CI xkb-free | ✓ VERIFIED | `//go:generate go run ./generator` (relative) at internal/layouts/tables.go:3 + DO-NOT-EDIT marker intact; mise task repointed to `./internal/layouts` (mise.toml:216); tables.go byte-identical since the prior pass's verified `go generate && git diff --exit-code` run — determinism carries; CI tidy-diff green at HEAD |
| 16 | SC-5: FSM unit corpus on synthetic streams | ✓ VERIFIED (deep) | internal/hotkey/fsm.go untouched — purity re-grepped clean (no goroutines/chans/time.Now/Sleep); 13 TestFSM_ tests; **TestFSM_SingleAtWindowExpiry green `-race` this pass** |
| 17 | SC-5: structured logs with levels; key trace only behind -debug | ✓ VERIFIED | internal/logging/logging.go untouched (slog.LevelVar INFO/DEBUG :22-26); key trace still slog.Debug-only (engine.go:166 re-grepped); main.go:38 `-debug` flag with password warning preserved; privacy contract re-verified in the rewritten README (README:407-413 — «Флаг `-debug` включает трассировку клавиш… каждое нажатие», warning against leaving it on) |
| 18 | SC-5: headless CI gates green (build/vet/lint/test -race/tidy-diff) | ✓ VERIFIED | **pr-sanity success on the exact current HEAD 1efc6fa** (run 2026-10-06T22:50:22Z — build+vet+strict lint+`test -race`+tidy-diff+govulncheck over the moved packages); verifier re-ran `go build ./...` (green) and 6 named behavioral tests (all green `-race`) locally this pass; prior pass re-ran the full gate green at 04d00ad |
| 19 | Plan 01-05: CI triplet (pr-sanity + dependabot gomod weekly + scheduled govulncheck), e2e excluded from CI | ✓ VERIFIED | All three workflow/config files untouched; pr-sanity green at HEAD (1efc6fa) and on PRs; security-scheduled green 2026-10-05 (weekly cadence continues); dependabot.yml unchanged; e2e remains excluded from push CI (dispatch-only e2e-matrix) |
| 20 | Plan 01-03: session actor serializes FSM, logs `{"msg":"action","n":N}` at expiry, wired into daemon | ✓ VERIFIED (deep) | Wiring re-grepped at HEAD: `session.NewActor(window)` (main.go:487 inside newActor), `actor.SetVersion(version)` (488), `engine.Config{Handler: actor, …}` (550), `engine.Run(ctx, engineConfig(actor))` (88); HandleKey holds `a.mu` across the whole event (actor.go:495-496); expiry emits `slog.Info("action", "n", int(action))` (actor.go:1017); **TestActor_Serialization + TestActor_ENTransitUnchanged + TestActor_RUTransitUnmapped all present, first two green `-race` this pass** |

**Score:** 20/20 truths verified (0 present-but-behavior-unverified)

Behavior-dependent truths note: panic containment (#9), FSM window semantics
(#16), actor serialization (#20), EN transit contract (#6) each had their
single named behavioral test run green under `-race` **this pass** at HEAD
1efc6fa (TestEngine_RecoverContainsPanic, TestFSM_SingleAtWindowExpiry,
TestActor_Serialization, TestActor_ENTransitUnchanged — plus TestAddress and
TestGolden_Punctuation for #5/#14). Full-suite runs were not repeated per the
re-verification scope; the last full `-race` gate ran green on this exact
HEAD in CI (pr-sanity, 2026-10-06T22:50Z). Live-bus truths (#5, #8, #10) rest
on committed live-run evidence per the phase's binding evidence policy, with
the engine protocol layer byte-identical (R100) to the live-proven head.

### Advisory (New Scope, Unevidenced)

Re-verification ran; the Step 7 debt-marker scan over all covered files that
changed since the prior pass (the 12 moved files at their new paths, main.go,
actor.go, go.mod, mise.toml, .golangci.yml, README.md, and the touched e2e
files) found **zero** TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER markers and zero
stub patterns. No new-scope findings; advisory entries: none.

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| 1 | None — only carry-forward informational notes below (conn.go race semantics; e2e README wording; lint tooling notice) | — | — |

### Required Artifacts

All 12 moved artifacts re-verified at their current paths
(internal/engine/{address,conn,engine,factory,keys,types}.go +
internal/layouts/{tables.go, tables_test.go, generator/main.go} +
their tests): exist, substantive (byte-identical R100 product code /
import-only test diffs), wired (imported by cmd/goswitchd/main.go and
internal/session). Prior pass's mechanical artifact score (31/31) carries by
byte-identity; plan-level patterns are frozen history (see Why-this-ran).

### Key Link Verification

Manual wiring re-verification at the current tree (plan-level patterns
staled by the rename — see Why-this-ran):

- main.go → internal/engine: import (main.go:29) + `engine.Run` (88) + `engine.Config{Handler: actor}` (550) — WIRED
- main.go → internal/session: `newActor()` → `session.NewActor` (487), `SetVersion` (488) — WIRED
- internal/engine → godbus v5.2.2: go.mod unchanged for godbus — WIRED
- e2e → daemon subprocess: case_m1/case_resilience spawn+log-oracle paths intact — WIRED
- .golangci.yml → internal/engine/conn.go gosec exclusion: repointed by 5427d7a — WIRED
- mise.toml → internal/layouts generation task: repointed (mise.toml:216) — WIRED
- pr-sanity.yml → mise.toml tasks: workflows untouched — WIRED
- AGENTS.md → .planning/codebase/CONVENTIONS.md regeneration chain: both untouched — WIRED
- ADR-001 → d01-experiment-log.md: both untouched — WIRED
- README.md → dist/systemd/user/goswitchd.service: **NOT_WIRED mechanically, designed succession** — carried in frontmatter `deferred` (README documents `goswitchctl install`, which writes the same unit via internal/install/install.go:60-61; SC-level no-root contract holds, truth #12)

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| internal/engine/conn.go (byte-identical) | addr | `Discover()`: $IBUS_ADDRESS → suffix-filtered freshest ~/.config/ibus/bus file | ✓ real socket discovery (TestAddress green; live-registered in the 31/31 matrix run) | ✓ FLOWING |
| internal/engine/engine.go (byte-identical) | ev / consume | decodeEvent of ProcessKeyEvent args; verdict from EventHandler | ✓ live-proven at byte-identical engine head (run 35108412175) | ✓ FLOWING |
| internal/session/actor.go | action n / version / config snapshot | FSM Feed at timer expiry; version pinned at construction (SetVersion main.go:488) | ✓ TestActor_Serialization green at HEAD; action-log line intact (:1017) | ✓ FLOWING |
| internal/layouts/tables.go (byte-identical) | ENToRU/RUToEN | generated from system xkb via ./generator | ✓ TestGolden_Punctuation green; DO-NOT-EDIT marker intact | ✓ FLOWING |
| docs/adr/d01-experiment-log.md (untouched) | verdict rows | appended by d01-probe live runs | ✓ 15 rows; case_d01.go append path intact (only a later gocritic sprintf fix) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Build (current tree, post-move) | `go build ./...` | exit 0 | ✓ PASS |
| Recover-shim behavior | `go test -race -count=1 -run 'TestEngine_RecoverContainsPanic' ./internal/engine/` | ok 1.0s | ✓ PASS |
| Address discovery | `go test -race -count=1 -run 'TestAddress' ./internal/engine/` | ok 1.0s | ✓ PASS |
| Golden punctuation tables | `go test -race -count=1 -run 'TestGolden_Punctuation' ./internal/layouts/` | ok 1.0s | ✓ PASS |
| Actor serialization | `go test -race -count=1 -run 'TestActor_Serialization' ./internal/session/` | ok 1.0s | ✓ PASS |
| EN transit (Phase 1 semantics) | `go test -race -count=1 -run 'TestActor_ENTransitUnchanged' ./internal/session/` | ok 1.0s | ✓ PASS |
| FSM window expiry | `go test -race -count=1 -run 'TestFSM_SingleAtWindowExpiry' ./internal/hotkey/` | ok 1.0s | ✓ PASS |
| Recover-shim count (deep) | `grep -c "defer recoverHandler" internal/engine/engine.go` (21) + factory.go (1) | 21 handlers + CreateEngine; 26 exported methods = 21 handlers + 5 emitters (0 unshimmed handlers) | ✓ PASS |
| No device capture (re-widened) | grep over internal/{engine,layouts,hotkey,logging,session} + cmd/goswitchd | comment-only matches (keycode semantics); no access | ✓ PASS |
| FSM purity | `grep -nE "go func\|chan \|time.Now\|Sleep" internal/hotkey/fsm.go` | no matches | ✓ PASS |
| Journal verdict count | `grep -c "Verdict\|Вердикт" docs/adr/d01-experiment-log.md` | 15 (≥4) | ✓ PASS |
| Daemon wiring | grep of cmd/goswitchd/main.go | engine.Run (88), newActor→session.NewActor (487), Handler: actor (550), SetVersion (488) | ✓ PASS |
| Key trace level | `grep -n 'slog.Debug("key"' internal/engine/engine.go` | line 166, DEBUG-only | ✓ PASS |
| Debt markers (changed covered files) | grep TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER over all changed/moved covered files | zero matches | ✓ PASS |
| Headless CI gate on current HEAD | `gh run list --workflow=pr-sanity.yml` | success at 1efc6fa (2026-10-06T22:50:22Z) — full go-ultimate gate incl. `-race` suite + lint + tidy-diff + govulncheck | ✓ PASS |
| Scheduled security scan | `gh run list --workflow=security-scheduled.yml` | success 2026-10-05 (weekly cadence) | ✓ PASS |
| Live m1/ibus-restart/kill9/d01 mise tasks | not re-run (drive owner's desktop) | committed Phase 1 evidence + intact assertions + byte-identical engine layer | ? SKIP (by policy) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` conventions in this repo. The runnable
probes are the mise tasks: headless CI gates ran green on this exact HEAD in
pr-sanity (2026-10-06T22:50Z, full gate including the `-race` suite over the
moved packages) — the verifier additionally ran the six named behavioral
tests locally (table above); live-session e2e tasks excluded by the binding
evidence policy (their committed live proof, e2e-matrix 31/31 at the
byte-identical engine head, stands).

### Requirements Coverage

All 10 Phase 1 requirement IDs remain mapped to Phase 1 and Complete in
REQUIREMENTS.md traceability (re-checked this pass: lines 94-121, «Phase 1:
10»); no orphans.

| Requirement | Source Plan | Status | Evidence |
|-------------|------------|--------|----------|
| CORR-08 | 01-02 | ✓ SATISFIED | truths #14-15 (tables byte-identical at internal/layouts/, golden test green) |
| INTEG-01 | 01-01 | ✓ SATISFIED | truth #5; conn.go/factory.go byte-identical to live-proven head |
| INTEG-02 | 01-01, 01-03 | ✓ SATISFIED | truths #6, #11, #13; e2e README INTEG-02 checklist intact; live keyd owner-UAT-accepted |
| INTEG-03 | 01-01, 01-03, 01-05 | ✓ SATISFIED | truths #11-12; CI live on GitHub runners, pr-sanity green at HEAD |
| INTEG-04 | 01-01, 01-03, 01-04 | ✓ SATISFIED | truth #8 |
| INTEG-05 | 01-01, 01-03 | ✓ SATISFIED | truths #9-10 |
| TEST-01 | 01-02, 01-03, 01-05 | ✓ SATISFIED | truths #16-18 |
| TEST-02 | 01-03, 01-04 | ✓ SATISFIED | truth #7; stand live-proven by the 31/31 matrix run |
| TEST-03 | 01-03 | ✓ SATISFIED | truth #7 (focus_helper.py untouched; `focus` contract intact) |
| INST-04 | 01-01, 01-05 | ✓ SATISFIED | truths #17, #19 |

### Decision Coverage

The phase's 01-CONTEXT.md is byte-untouched since the prior pass's honored
verdict; the decision set has not changed. (The tool-level
`check decision-coverage-verify` gate reported `skipped: CONTEXT.md missing`
under current gsd-tools path resolution in the prior pass — warning-only,
non-blocking; nothing about the decision corpus changed since.)

### Anti-Patterns Found

Debt-marker scan re-run over all covered files changed or moved since the
prior pass: **ZERO TBD/FIXME/XXX, zero TODO/HACK/PLACEHOLDER, zero stub
patterns** (including the 12 moved files at their new paths and the
later-phase-modified shared files). Carry-forward warnings (files
byte-identical/moved, still non-blocking quality debt documented in prior
passes): boot-time retry race semantics (internal/engine/conn.go, moved R100
from engine/conn.go), NameFlagReplaceExisting inertness, length-only oracle
limitation in locked-session d01 re-runs.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| internal/engine/conn.go | (moved R100 from engine/conn.go) | carry-forward boot-time retry race + rapid-respawn race (documented quality debt; also the gosec G404 exclusion target, repointed correctly by 5427d7a) | ⚠️ Warning | Canonical paths demonstrably green (31/31 live matrix at byte-identical engine head); hardening debt for future work |
| lint tooling | — | golangci-lint `exhaustruct` deprecation notice (v2.13+ → exhaustruct_v5) — carry-forward info, 0 issues, pr-sanity green at HEAD | ℹ️ Info | Cosmetic future config rename |
| test/e2e/README.md | 220 | INTEG-02 section still carries the Phase 1 wording «возвращающий `false` (наблюдатель)» — predates the consumption succession; the architectural claim (no evdev layer, keyd below IBus) remains true | ℹ️ Info | Documentation nuance only |
| README.md | install section | documents `goswitchctl install` instead of the plan-01-05 manual unit-copy dev instruction | ℹ️ Info | Designed succession (see `deferred`); SC-level no-root contract intact |

### Human Verification Required

None new. The four environment-gated items from the original pass were
resolved by UAT (01-UAT.md: 4/4 pass, owner-accepted 2026-09-11) and are not
re-opened. No ⚠️ PRESENT_BEHAVIOR_UNVERIFIED truths exist on this pass — every
behavior-dependent truth had its named test run green this pass. Live-session
e2e re-runs remain policy-excluded (owner's desktop; committed live evidence
valid via byte-identity of the engine layer).

### Gaps Summary

**No gaps.** All 20 must-have truths verified on the current tree at HEAD
1efc6fa. The fingerprint staleness is fully resolved: the 12 moved
engine/layouts files were re-verified at their `internal/` paths (product
files byte-identical R100; test files import-path-only diffs, diff-checked);
wiring re-verified manually where mechanical plan patterns staled (main.go →
internal/engine/session imports and calls; .golangci.yml and mise.toml path
repoints; README/install succession re-confirmed); six named behavioral tests
green under `-race` this pass; the full headless gate ran green on this exact
HEAD in CI (pr-sanity 2026-10-06T22:50Z). Later-phase changes to shared files
(main.go, actor.go, go.mod, README, SPEC, mise.toml) are owner-sanctioned
successions (spec-delta D-55 series) that preserve every Phase 1 seam; the
one non-verified plan-level key link (README → unit file) remains the
recorded designed succession in `deferred`, with its SC-level no-root
contract holding. Zero debt markers, zero orphans, zero regressions. Phase 1
remains a sound foundation on the milestone-closed tree.

---

_Verified: 2026-10-07T07:43:52Z_
_Verifier: Claude (gsd-verifier)_
