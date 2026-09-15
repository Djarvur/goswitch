---
phase: 01-adr-paket-i-karkas-ibus-dvizhka
verified: 2026-09-15T03:55:00Z
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
covered_digest: "v1:sha256:6e49a675c1c1cae667604c0394efdd0dee32e8f650da650c5620fe5c9a2ba621"
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

**Verified:** 2026-09-15T03:55:00Z
**Status:** passed
**Re-verification:** Yes — fingerprint refresh after Phase 2 modified shared covered files

## Why this re-verification ran

Phase 1 passed verification 2026-09-11 (20/20) and UAT 4/4 (owner-accepted
2026-09-11, 01-UAT.md). Phase 2 execution then legitimately modified covered
files. Git classification of every change since the prior `verified:`
timestamp (2026-09-11T12:10:00Z):

- **Phase-1 merge only** (`b982780`, PR #1 squash, 17:42 UTC — the verified
  content itself landing): engine/conn.go, engine/address*.go, engine/wire_test.go,
  internal/hotkey/*, internal/logging/*, layouts/*, docs/SPEC.md, docs/adr/*
  (ADR-001..005 + journal), dist/systemd/user/goswitchd.service,
  .github/workflows/{pr-sanity,security-scheduled}.yml, .github/dependabot.yml,
  .golangci.yml, README.md, AGENTS.md → regression-checked, not deep-re-read.
- **Phase 2 rework** (full 3-level + behavioral re-check): engine/{engine.go,
  engine_test.go, factory.go, keys.go, types.go}, internal/session/{actor.go,
  actor_test.go}, cmd/goswitchd/main.go, test/e2e/{case_d01.go, case_m1.go,
  case_resilience.go, focus_helper.py, main.go, preflight.go, README.md},
  go.mod, go.sum, mise.toml, .gitignore.

The prior report had no `gaps:` section; its truths, evidence and live-gate
policy carried over as the map. UAT resolution (4/4 pass) is final and was
not re-opened.

## Designed successions (documented, not gaps)

Two plan-level Phase 1 details were superseded by Phase 2's roadmap-sanctioned
work. The SC-level contracts both still hold:

1. **ProcessKeyEvent consumption.** Plan 01-01 said «returns false for all
   events без исключения». Phase 2's goal (word correction EN↔RU) requires the
   engine to consume keys in RU script mode, so the verdict is now delegated
   to the EventHandler (engine.go:147-170). The SC-2/INTEG-02 contract — normal
   typing transits — is preserved and behaviorally pinned: a nil handler stays
   a pure observer by construction; a declining handler transits everything
   (`TestEngine_ProcessKeyEventReturnsFalse`, reworked to the consume contract);
   EN mode is explicitly «Phase 1 semantics — transit» (actor.go:390-396,
   `TestActor_ENTransitUnchanged`); bare modifiers, non-printables and true
   combos are never consumed (actor.go:356-373). All green under `-race`.
2. **go.mod dependency set.** Plan 01-01 said godbus was the only external
   dependency. Phase 2 added `gopkg.in/yaml.v3 v3.0.1` for the e2e YAML matrix
   — a core technology in the approved stack (research/STACK.md) since before
   Phase 1. The minimal-dep constraint holds: 2 direct deps + 1 indirect
   (golang.org/x/sys via godbus).

## Goal Achievement

### Observable Truths

| # | Truth | Status | Re-verification evidence (current tree) |
|---|-------|--------|------------------------------------------|
| 1 | SC-1: ADR-001..005 exist in unified format (Status/Context/Decision/Consequences/Reversibility), CONTEXT decisions quoted | ✓ VERIFIED | Untouched since PR #1; re-checked: each of the 5 files carries 5/5 section headers, Status = Accepted (гейт M0) |
| 2 | SC-1: D-01 journal — ≥4 verdict rows, decision derivable | ✓ VERIFIED | `grep -c "Verdict\|Вердикт" docs/adr/d01-experiment-log.md` = **15**; file untouched; 0 length-only-oracle rows (taint check re-run: 0 matches) |
| 3 | SC-1: ADR-001 written last from journal — winner/Option B + kill-criteria, losers' observed failures, D-03 consequences | ✓ VERIFIED | File untouched since the verified pass (git: only b982780) |
| 4 | SC-1: M0 owner gate passed; MACR-01 fixed in ADR-005 (D-12), spec-deltas applied; no MACR-01 code | ✓ VERIFIED | All 5 ADRs Accepted; SPEC.md untouched; grep for macr over cmd/ internal/ engine/ layouts/ (incl. new internal/correct/) = **0** |
| 5 | SC-2: programmatic registration on the private IBus bus (no XML, no root), two engines goswitch-en/-ru (INTEG-01) | ✓ VERIFIED | engine/conn.go byte-unchanged since the live proof (git: only b982780); factory.go rework only attaches the emitter sink; **fresh machine-generated live evidence**: e2e-matrix CI run 34913812782 (2026-09-15, self-hosted GNOME) 16/16 PASS with a fresh daemon per case — registration + activation re-proven on the current tree; TestAddress_* green |
| 6 | SC-2: as active source, engine receives ProcessKeyEvent, logs keys; normal typing transits | ✓ VERIFIED | Evolved per Phase 2 (see Designed successions §1): transit contract behaviorally pinned — TestEngine_ProcessKeyEventReturnsFalse, TestActor_ENTransitUnchanged, TestActor_RUTransitUnmapped all green under `-race`; key trace still slog.Debug-only (engine.go:159); live: matrix word cases type text through the engine on the current tree |
| 7 | SC-2: e2e skeleton — ydotool injection, self-activation, fail-fast preflight, exit-code contract (TEST-02/03) | ✓ VERIFIED | main.go: `os.Exit(run())`, non-zero on FAIL (re-read); preflight keeps all 6 Phase 1 named checks (injection self-test, ibus address, uinput writable, /usr/bin/python3 gi, engine registered, log heartbeat) + 2 new surface checks; focus_helper.py keeps /usr/bin/python3 shebang + AT-SPI grabFocus; m1 assertions intact (`"msg":"key"` ≥ min, `"msg":"action","n":2`, registered-before-focus_in by timestamps) |
| 8 | SC-3: ibus restart → re-registration on the new socket, keys flow again (INTEG-04) | ✓ VERIFIED | Reconnect implementation (engine/conn.go Run/serve: fresh Discover per cycle, signal-channel bus-loss, jittered backoff) byte-unchanged since the committed live PASS; case_resilience.go reworked for multi-surface but assertions intact (literal `re-registered` wait + post-restart `"msg":"key"` growth, lines 34-55); mise task e2e-ibus-restart present |
| 9 | SC-3: recover shim contains handler panics (INTEG-05) | ✓ VERIFIED | Re-counted: `defer recoverHandler` on **21/21** exported Engine methods + CreateEngine = every exported D-Bus handler; **TestEngine_RecoverContainsPanic green under -race (re-run this pass)** — injected panic, safe replies, ERROR log |
| 10 | SC-3: kill -9 → desktop input alive, daemon respawns and re-registers | ✓ VERIFIED | case_resilience.go kill9 path intact (SIGKILL → plain-engine switch → inject → desktop oracle via witness/zenity readback → respawnAndWait registrations+1 `component registered`); dist unit `Restart=on-failure` unchanged; committed live PASS; mise task present |
| 11 | SC-4: no EVIOCGRAB / evdev / uinput capture in product code | ✓ VERIFIED | Structural grep over engine/ cmd/ internal/ layouts/ (now including Phase 2's internal/correct/) + OpenFile scan: **CLEAN** |
| 12 | SC-4: works over default Ubuntu 24.04 GNOME Wayland IM stack without root (INTEG-03) | ✓ VERIFIED | systemd **user** unit unchanged (PartOf=graphical-session.target, ExecStart=%h/go/bin — no root anywhere); conn.go EXTERNAL/SO_PEERCRED auth unchanged; mise user-local |
| 13 | SC-4: keyd/xremap non-conflict (INTEG-02) | ✓ VERIFIED | No-capture gate (truth 11) + transit (truth 6) + README INTEG-02 owner checklist intact (test/e2e/README.md:167); live keyd session was owner-UAT-accepted 2026-09-11 |
| 14 | SC-5: golden tests of generated tables incl. `[ ] ; ' , . /` + asymmetric pairs (CORR-08) | ✓ VERIFIED | layouts/ untouched; TestGolden_* green under -race; live re-proof: matrix punctuation/digit cases PASS on the current tree |
| 15 | SC-5: tables generated by go:generate, deterministic, CI xkb-free | ✓ VERIFIED | **Re-run this pass: `go generate ./layouts && git diff --exit-code` → byte-identical**; generated-marker + go:generate header intact |
| 16 | SC-5: FSM unit corpus on synthetic streams (TEST-01) | ✓ VERIFIED | All 11 TestFSM_* enumerated, green under -race; purity grep (goroutines/chans/time.Now/Sleep in fsm.go) clean; fsm.go untouched |
| 17 | SC-5: structured logs with levels; key trace only behind -debug (INST-04) | ✓ VERIFIED | logging.go untouched (LevelVar INFO/DEBUG); TestLogging_JSONShape/LevelFiltering/KeyTraceGated green; -debug password warning in main.go:21; README «Режим отладки и конфиденциальность» (README.md:13) |
| 18 | SC-5: headless CI gates green (build/vet/lint/test -race/tidy-diff) | ✓ VERIFIED | **Verifier re-ran all gates this pass**: `go build ./...` OK; `go vet ./...` OK; `go test -race ./...` — 8 packages ok (incl. new internal/correct and test/e2e matrix decoder); `mise run lint` → 0 issues; `go mod tidy` → zero drift |
| 19 | Plan 01-05: CI triplet (pr-sanity + dependabot gomod weekly + scheduled govulncheck), e2e excluded from CI | ✓ VERIFIED | All three workflow files untouched; **fresh machine evidence**: pr-sanity green ×3 on this branch (latest run 34914315850, 2026-09-15); security-scheduled cron actually fired green on schedule 2026-09-14 (run 34838604697 — the UAT deferred observation now has real evidence); dependabot dynamic runs green; e2e remains CI-excluded except the dispatch-only e2e-matrix workflow (Phase 2 artifact) |
| 20 | Plan 01-03: session actor serializes FSM, logs `{"msg":"action","n":N}` at expiry, wired into daemon | ✓ VERIFIED | Wiring intact: `session.NewActor(hotkey.DefaultWindow)` at cmd/goswitchd/main.go:53, actor = Config.Handler; mutex-serialized single entry point preserved through the Phase 2 rework; ExpiryAt still emits `slog.Info("action", "n", int(action))` (actor.go:259); **TestActor_WindowTimerFiresAutomatically + TestActor_Serialization green under -race (re-run this pass)** |

**Score:** 20/20 truths verified (0 present-but-behavior-unverified)

Behavior-dependent truths note: panic containment (#9), FSM window semantics
(#16), actor expiry/timer/serialization (#20), transit contract (#6) are all
covered by behavioral tests that ran green in this pass's `go test -race`
suite. Live-bus truths (#5, #8, #10) rest on committed live-run evidence per
the phase's evidence policy — #5 additionally re-proven live on the current
tree by the 16/16 e2e-matrix CI run; #8/#10 core implementation (conn.go) is
byte-unchanged since their live proof and their asserting e2e cases survive
the Phase 2 surface refactor with assertions intact.

### Advisory (New Scope, Unevidenced)

None. The Step 7 scan over all covered files (debt markers, stub patterns,
empty implementations) found nothing new; the only observation is the
informational lint notice below, which is tooling news, not a code defect.

### Required Artifacts

`gsd_run query verify.artifacts` per plan, re-run this pass: **31/31 passed,
0 issues** — 01-01 (8/8), 01-02 (5/5), 01-03 (6/6), 01-04 (7/7), 01-05 (5/5).

### Key Link Verification

`gsd_run query verify.key-links` per plan, re-run this pass: **14/14
verified** ("Pattern found in source/target") across all five plans —
including main.go→session.NewActor, e2e→daemon subprocess, ADR-001→journal,
ADR-004→SPEC delta, pr-sanity→mise.toml, README→systemd unit,
AGENTS.md→CONVENTIONS.md regeneration chain.

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| engine/conn.go (unchanged) | addr | `Discover()` per cycle: $IBUS_ADDRESS → newest ~/.config/ibus/bus file | ✓ real socket discovery (TestAddress_* green; live-registered in the 16/16 matrix run) | ✓ FLOWING |
| engine/engine.go (Phase 2 rework) | ev / consume | decodeEvent of ProcessKeyEvent args; verdict from EventHandler | ✓ matrix word cases prove keys flow and corrections commit on the live bus | ✓ FLOWING |
| internal/session/actor.go (Phase 2 rework) | action n / corrections | FSM Feed at timer expiry; ladder via Emitter | ✓ TestActor_* (29 tests) + live matrix corrections | ✓ FLOWING |
| layouts/tables.go (unchanged) | ENToRU/RUToEN | generated from system xkb | ✓ regeneration byte-identical this pass | ✓ FLOWING |
| docs/adr/d01-experiment-log.md (unchanged) | verdict rows | appended by d01-probe live runs | ✓ 15 rows; append path intact in reworked case_d01.go (recordVerdict) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Build | `go build ./...` | exit 0 | ✓ PASS |
| Vet | `go vet ./...` | exit 0 | ✓ PASS |
| Full race suite (single run, incl. TestEngine_RecoverContainsPanic, TestEngine_ProcessKeyEventReturnsFalse, TestActor_WindowTimerFiresAutomatically, TestActor_Serialization, TestActor_ENTransitUnchanged, TestFSM_*, TestGolden_*, TestLogging_*) | `go test -race ./...` | 8 packages ok, 0 failures | ✓ PASS |
| Lint (strict v2 config) | `mise run lint` | 0 issues | ✓ PASS |
| Tidy drift | `go mod tidy` + diff | zero drift | ✓ PASS |
| Generation determinism | `go generate ./layouts && git diff --exit-code` | byte-identical | ✓ PASS |
| Panic containment count | `grep -c "defer recoverHandler" engine/*.go` | 21 (engine.go) + 1 (factory.go) = all exported handlers | ✓ PASS |
| No device capture | `grep -riE "EVIOCGRAB\|uinput\|/dev/input\|evdev" engine/ cmd/ internal/ layouts/` | no matches | ✓ PASS |
| FSM purity | `grep -E "go func\|chan\|time.Now\|Sleep" internal/hotkey/fsm.go` | no matches | ✓ PASS |
| Journal verdict count | `grep -c "Verdict\|Вердикт" docs/adr/d01-experiment-log.md` | 15 (≥4) | ✓ PASS |
| CI green on current branch | `gh run list` | pr-sanity ×3 success (latest 34914315850); e2e-matrix 34913812782 success; security-scheduled 34838604697 success (cron-fired) | ✓ PASS |
| Live e2e matrix artifact | `gh run download 34913812782 -n e2e-report` | 16/16 PASS (word-en-ru … word-gte-capitalized), 2026-09-15T03:36 | ✓ PASS |
| Live m1/ibus-restart/kill9/d01 mise tasks | not re-run (drive owner's desktop) | committed Phase 1 evidence + intact assertions + unchanged conn.go | ? SKIP (by policy) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` conventions in this repo. The runnable probes
are the mise tasks: headless CI gate re-run this pass (PASS); live-session e2e
tasks excluded by the binding evidence policy — their machine-generated CI
counterpart (e2e-matrix, dispatch-only) was fetched and is green 16/16.

### Requirements Coverage

All 10 Phase 1 requirement IDs remain mapped to Phase 1 and Complete in
REQUIREMENTS.md traceability; no orphans.

| Requirement | Source Plan | Status | Evidence |
|-------------|------------|--------|----------|
| CORR-08 | 01-02 | ✓ SATISFIED | truths #14-15 |
| INTEG-01 | 01-01 | ✓ SATISFIED | truth #5; CommitText/DeleteSurroundingText/ForwardKeyEvent emitters now live-driven (matrix corrections) |
| INTEG-02 | 01-01, 01-03 | ✓ SATISFIED | truths #6, #11, #13; live keyd owner-UAT-accepted |
| INTEG-03 | 01-01, 01-03, 01-05 | ✓ SATISFIED | truths #11-12; CI live on GitHub runners |
| INTEG-04 | 01-01, 01-03, 01-04 | ✓ SATISFIED | truth #8 |
| INTEG-05 | 01-01, 01-03 | ✓ SATISFIED | truths #9-10 |
| TEST-01 | 01-02, 01-03, 01-05 | ✓ SATISFIED | truths #16-18 |
| TEST-02 | 01-03, 01-04 | ✓ SATISFIED | truth #7; stand re-proven live by the 16/16 matrix run |
| TEST-03 | 01-03 | ✓ SATISFIED | truth #7 (focus_helper.py contract intact) |
| INST-04 | 01-01, 01-05 | ✓ SATISFIED | truths #17, #19 |

### Anti-Patterns Found

Debt-marker gate re-run over all covered Go/Python files: **ZERO**
TBD/FIXME/XXX, zero TODO/HACK/PLACEHOLDER. Carry-forward warnings from the
prior pass (files unchanged, still non-blocking quality debt for later
hardening): WR-01 re-registered generation semantics (conn.go:119-123),
WR-02 NameFlagReplaceExisting inertness (conn.go:88-101), WR-03 length-only
oracle limitation in locked-session d01 re-runs.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| engine/conn.go | 93, 122 | WR-01/WR-02 carry-forward (unchanged since prior pass) | ⚠️ Warning | Boot-time retry race + rapid-respawn race; canonical paths demonstrably green (16/16 live matrix) |
| .golangci.yml → lint output | — | golangci-lint reports `exhaustruct` linter deprecated since v2.13.0 (replaced by exhaustruct_v5) — informational tooling notice, config still applies, 0 issues | ℹ️ Info | Cosmetic future config rename; no code impact |

### Human Verification Required

None new. The four environment-gated items from the prior pass were resolved
by UAT (01-UAT.md: 4/4 pass, owner-accepted 2026-09-11) and are not re-opened.
The two deferred-equivalent-evidence observations recorded there stand as
deferred follow-ups; note that since then the scheduled security-scheduled run
has actually fired green (2026-09-14, run 34838604697), giving the fourth item
real observed evidence — the UAT record remains authoritative.

### Gaps Summary

**No gaps.** All 20 must-have truths verified on the current (post-Phase-2)
tree; 31/31 artifacts pass; 14/14 key links wired; requirements 10/10 with
zero orphans; zero debt markers; all headless gates re-run green by the
verifier; fresh machine-generated live evidence (16/16 e2e-matrix on the
self-hosted GNOME runner, three green pr-sanity runs on this branch) covers
the registration/key-flow/transit chain on the exact code under verification.
Two Phase-1 plan-level details were superseded by Phase 2's designed behavior
(consumption delegation; yaml.v3 dependency) — both documented above, both
with the SC-level contract preserved and behaviorally pinned.

---

_Verified: 2026-09-15T03:55:00Z_
_Verifier: Claude (gsd-verifier)_
