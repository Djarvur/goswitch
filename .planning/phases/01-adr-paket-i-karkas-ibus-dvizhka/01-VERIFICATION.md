---
phase: 01-adr-paket-i-karkas-ibus-dvizhka
verified: 2026-10-07T23:10:20Z
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
covered_digest: "v1:sha256:e3866f09bcfded8f1bde1c1788114fdd0544c96f1b77f347a55e5336ca7fc513"
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
    evidence: "Phase 4 goal «Продукт ставится без root по документированной инструкции» + SC-1; Phase 4-01 INST-01 implemented goswitchctl install, which writes the same goswitchd.service into .config/systemd/user (internal/install/install.go — unitDirRel = \".config/systemd/user\", unitFile = \"goswitchd.service\"); Phase 4-07 and the 08-07 README rewrite (RU full v1.1.0) both document that installer (README:30 «одна команда: goswitchctl install», idempotent, no-root at README:17). The Phase 1 SC-level contract (user unit, no root, truth #12) is untouched; only the plan-level grep pattern (goswitchd\\.service in README) still does not match — re-confirmed unchanged this pass (README.md 0 commits since prior pass)."
---

# Phase 01: ADR-пакет и каркас IBus-движка — Verification Report

**Phase Goal (ROADMAP.md):** «Все архитектурные развилки закрыты утверждёнными ADR, а `goswitchd` работает как IBus engine: видит клавиатурные события как активный input source, логирует их, переживает рестарт ibus-daemon и собственные паники, не конфликтуя с keyd/xremap — фундамент, на котором безопасно строить коррекцию и переключение.»

**Verified:** 2026-10-07T23:10:20Z
**Status:** passed
**Re-verification:** Yes — second staleness refresh of the milestone-close chain (after the 2026-10-07T07:43:52Z pass); quick task 261008-00m landed after it, changing tree code this phase covers

## Why this re-verification ran

Phase 1 passed its first verification 2026-09-11 (20/20), was owner-UAT-accepted
2026-09-11 (01-UAT.md, 4/4), and has been re-verified five times since (2026-09-15
×2, 2026-09-16, 2026-09-20, 2026-10-07T07:43:52Z — all 20/20, previous verdict
`passed` with an empty `gaps:` list). After the 2026-10-07T07:43:52Z pass, quick
task 261008-00m (milestone-close owner fixes) landed two commits that touch
files this phase covers:

- **4a0c887** (2026-10-08T00:46+03:00) `fix(261008-00m): own engine flip keeps
  the correction buffer (G-2-3)` — `internal/session/actor.go` (+56 lines:
  `flipCredit` field, `HandleLifecycle` synthetic-credit branch, `flipTo`
  arming) and `internal/session/actor_test.go` (77c91b7 RED + 4a0c887 green).
- **0781a5f** (2026-10-08T00:38+03:00) `docs(261008-00m): spec-deltas —
  ADR-004 own-flip revision … (D-55)` — `docs/adr/ADR-004-buffer-reset-triggers.md`
  (dated amendment) + `docs/SPEC.md` (§4.3 dated delta, old bullet preserved
  verbatim) + ADR-006 amendment (phase-5 scope, not this phase's ADR pack).

Git classification of every covered file since the prior `verified:`
timestamp (2026-10-07T07:43:52Z), at HEAD 8292dd4 — exactly 4 covered files
changed:

- `internal/session/actor.go` (1 commit: 4a0c887)
- `internal/session/actor_test.go` (2 commits: 77c91b7 RED, 4a0c887 green)
- `docs/adr/ADR-004-buffer-reset-triggers.md` (1 commit: 0781a5f)
- `docs/SPEC.md` (1 commit: 0781a5f)

**Untouched since prior pass (0 commits):** all 12 `internal/engine/*` +
`internal/layouts/*` files, `internal/hotkey/*`, `internal/logging/*`,
`cmd/goswitchd/main.go`, `dist/systemd/user/goswitchd.service`, the other four
ADR files + d01 journal, go.mod/go.sum, mise.toml, .golangci.yml,
.github/{dependabot.yml,workflows/*}, README.md, AGENTS.md,
.planning/codebase/CONVENTIONS.md, all five Phase 1 PLAN/SUMMARY pairs,
01-REVIEW.md, 01-VALIDATION.md, all test/e2e/* files. All 55 covered paths
re-checked for existence at the current tree — none missing; the covered-file
set is unchanged (the quick task's internal/install/* edits are Phase 4
coverage, and ADR-006 is Phase 5 coverage — neither belongs to this phase's
set).

## This pass's delta audits

### ADR-004 amendment (0781a5f) — the ADR pack contract holds

The amendment is **docs-only, dated, and additive**: an `Amended 2026-10-07 —
G-2-3` note under Status plus an appended `## Amendment 2026-10-07 — G-2-3 …`
section; the original Context/Decision/Consequences/Reversibility text is
untouched (diff is +36/-0). Contract checks re-run this pass:

- All five ADR-001..005 Status headers still read «Accepted — гейт M0
  (ревью владельца, план 01-04 Task 3)» — re-grepped, 5/5.
- ADR-004 unified format intact: all 5 canonical headers
  (Status/Context/Decision/Consequences/Reversibility) still present, plus the
  additive amendment appendix (6 `##` sections total).
- The amendment records the owner decision (G-2-3, 2026-10-07) with the live
  evidence trail (matrix v1 word-mixed readback, daemon journal line), exactly
  the audit-trail pattern of the owner-sanctioned D-55 spec-delta series; the
  paired `docs/SPEC.md` §4.3 delta marks the revision with a dated comment and
  preserves the previous bullet verbatim as audit trail.
- d01 journal untouched: `grep -c "Verdict\|Вердикт"` = **15** (≥4).

### actor.go flipCredit delta (4a0c887) — the Phase 1 actor contract holds

The delta adds `flipCredit int` (actor.go:263, documented «Accessed only under
a.mu»), arms exactly one credit in `flipTo` after a successful `SetGlobalEngine`
round trip (actor.go:2017; the WARN-failure and nil-seam paths arm nothing),
and in `HandleLifecycle`'s FocusOut branch consumes exactly one credit to skip
ONLY `a.buf.HardReset()` (actor.go:788-793) — every context-scoped side effect
(FSM reset, surrounding-cache clear, pending retirement, timer stop) stays; a
`Reset` kind never consumes a credit.

Phase-1 contract re-verification on the changed file:

- **Serialization:** `HandleLifecycle` still opens with `a.mu.Lock();
  defer a.mu.Unlock()` (actor.go:771-772). All 6 `flipTo` call sites hold the
  mutex: `ExpiryAt` (locks at entry), `ToggleMode`, `SwitchMode` (direct
  locks), `settleCombo` / `modeSwitchChord` (documented caller-holds), and
  `settleCorrectionFlip` whose callers (`executeCorrection`,
  `executeSelectionCorrection`, `executeLevel2`) sit on locked HandleKey /
  handleSurroundingLocked / startRangeCorrection paths. `flipCredit` has
  exactly 3 references (field, decrement, increment) — all under the lock.
  `TestActor_Serialization` green `-race` this pass.
- **Action log at expiry:** `slog.Info("action", "n", int(action))` still at
  actor.go:1051 inside `ExpiryAt` — the e2e oracle shape unchanged.
- **Daemon wiring:** re-grepped at HEAD — `engine.Run(ctx,
  engineConfig(actor))` (main.go:88), `session.NewActor(window)` +
  `actor.SetVersion(version)` (main.go:487-488), `engine.Config{Handler:
  actor, …}` (main.go:550). main.go itself untouched since prior pass.
- **No device capture:** structural grep re-run over internal/{engine,layouts,
  hotkey,logging,session} + cmd/goswitchd — CLEAN (no EVIOCGRAB, no
  /dev/uinput, no /dev/input; the only "evdev" matches remain the later-phase
  comment-only keycode references in actor.go).
- **Recover shim / FSM purity / engine+layouts layers:** files untouched
  (0 commits since prior pass); named tests re-run green this pass (table
  below).

**Live-bus evidence status:** unchanged from the prior pass — the committed
live proof (e2e-matrix run 35108412175, 31/31 PASS at head 9f2fd75) remains
valid because every `internal/engine/` file is still byte-identical to the
live-proven head (0 commits since the prior pass's R100 verification); the
flipCredit delta lives entirely in `internal/session/actor.go`, downstream of
the proven protocol layer.

**CI note:** the newest pr-sanity run remains the one at 1efc6fa
(2026-10-06T22:50:22Z, success) — the post-prior-pass commits (all in the
quick-261008-00m series, tip docs-only) have no CI run yet. The regression
evidence this pass is therefore local: `go build ./...` green, scoped
`golangci-lint run ./internal/session/...` = **0 issues**, and 8 named tests
green under `-race` (table below), including both new flipCredit tests.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Re-verification evidence (current tree, HEAD 8292dd4) |
|---|-------|--------|------------------------------------------|
| 1 | SC-1: ADR-001..005 exist in unified format, CONTEXT decisions quoted | ✓ VERIFIED (regression) | All 5 ADR files: Status = «Accepted — гейт M0 (ревью владельца, план 01-04 Task 3)» re-grepped 5/5 this pass; ADR-004's dated amendment is additive (+36/-0), all 5 canonical headers intact |
| 2 | SC-1: D-01 journal — ≥4 verdict rows, decision derivable | ✓ VERIFIED (regression) | File untouched; `grep -c "Verdict\|Вердикт"` = **15** (re-run this pass) |
| 3 | SC-1: ADR-001 written last from journal — winner/Option B + kill-criteria, losers' failures, D-03 consequences | ✓ VERIFIED (regression) | File untouched (0 commits since prior pass) |
| 4 | SC-1: M0 owner gate passed; MACR-01 in ADR-005 (D-12), spec-deltas applied; no MACR-01 code in Phase 1 | ✓ VERIFIED (regression) | All 5 ADRs Accepted; SPEC.md's new delta (0781a5f) is the same owner-spec-delta D-55 pattern (dated, old text preserved); MACR code ownership with later phases unchanged |
| 5 | SC-2: programmatic registration on the private IBus bus (no XML, no root), two engines goswitch-en/-ru | ✓ VERIFIED | internal/engine/ untouched since prior pass (byte-identical to live-proven head); TestAddress green `-race` this pass; main.go wiring intact (88/487-488/550) |
| 6 | SC-2: as active source, engine receives ProcessKeyEvent, logs keys; normal typing transits | ✓ VERIFIED | engine.go untouched (decode → `slog.Debug("key",…)` at :166, DEBUG-only); **TestActor_ENTransitUnchanged green `-race` this pass**; the actor-side transit code path is untouched by the flipCredit delta (credit logic only in HandleLifecycle/flipTo) |
| 7 | SC-2: e2e skeleton — ydotool injection, self-activation, fail-fast preflight, exit-code contract | ✓ VERIFIED (regression) | test/e2e/* untouched (0 commits since prior pass); case_m1.go oracles intact |
| 8 | SC-3: ibus restart → re-registration on the new socket, keys flow again | ✓ VERIFIED (regression) | conn.go untouched; case_resilience.go "re-registered" wait intact; mise task e2e-ibus-restart present (mise.toml untouched) |
| 9 | SC-3: recover shim contains handler panics | ✓ VERIFIED (regression) | engine.go untouched — shim count re-confirmed 21/21 handlers + CreateEngine in the prior pass; **TestEngine_RecoverContainsPanic green `-race` this pass** |
| 10 | SC-3: kill -9 → desktop input alive, daemon respawns and re-registers | ✓ VERIFIED (regression) | case_resilience.go + dist unit untouched; committed live PASS stands; engine layer byte-identical |
| 11 | SC-4: no EVIOCGRAB / evdev / uinput capture in product code | ✓ VERIFIED (re-widened this pass) | Structural grep re-run over internal/{engine,layouts,hotkey,logging,session} + cmd/goswitchd at HEAD — **CLEAN**, including the changed actor.go (flipCredit adds no device access) |
| 12 | SC-4: works over default Ubuntu 24.04 GNOME Wayland IM stack without root | ✓ VERIFIED (regression) | systemd **user** unit untouched; conn.go untouched; README:17 «без root везде» (README untouched this pass) |
| 13 | SC-4: keyd/xremap non-conflict | ✓ VERIFIED | No-capture gate (truth 11, re-run) + EN transit (truth 6, green test) + e2e README INTEG-02 section intact (untouched); live keyd session owner-UAT-accepted 2026-09-11 |
| 14 | SC-5: golden tests of generated tables incl. `[ ] ; ' , . /` + asymmetric pairs | ✓ VERIFIED (regression) | internal/layouts/ untouched; **TestGolden_Punctuation green `-race` this pass** |
| 15 | SC-5: tables generated by go:generate, deterministic, CI xkb-free | ✓ VERIFIED (regression) | tables.go untouched (`//go:generate go run ./generator` + DO-NOT-EDIT intact); mise.toml untouched; CI tidy-diff green at 1efc6fa |
| 16 | SC-5: FSM unit corpus on synthetic streams | ✓ VERIFIED (regression) | internal/hotkey/fsm.go untouched — purity re-grepped clean this pass (no goroutines/chans/time.Now/Sleep); **TestFSM_SingleAtWindowExpiry green `-race` this pass** |
| 17 | SC-5: structured logs with levels; key trace only behind -debug | ✓ VERIFIED (regression) | internal/logging/logging.go untouched; key trace still slog.Debug-only (engine.go:166, untouched); main.go:38 `-debug` flag preserved; actor's action log still INFO (actor.go:1051) |
| 18 | SC-5: headless CI gates green (build/vet/lint/test -race/tidy-diff) | ✓ VERIFIED | pr-sanity success at 1efc6fa (2026-10-06T22:50:22Z, newest run); post-prior-pass delta covered locally this pass: `go build ./...` green, `golangci-lint run ./internal/session/...` = 0 issues, 8 named tests green `-race` (the quick task's own gates ran green per its verified record, commit e00e11f «verified 6/6, live 16/16») |
| 19 | Plan 01-05: CI triplet (pr-sanity + dependabot gomod weekly + scheduled govulncheck), e2e excluded from CI | ✓ VERIFIED (regression) | All three workflow/config files untouched; security-scheduled success 2026-10-05 (weekly cadence); e2e remains dispatch-only |
| 20 | Plan 01-03: session actor serializes FSM, logs `{"msg":"action","n":N}` at expiry, wired into daemon | ✓ VERIFIED (deep, on the changed file) | Wiring re-grepped at HEAD (main.go:88/487-488/550); HandleKey and HandleLifecycle both hold `a.mu` across the whole event (actor.go:506, 771-772); expiry emits `slog.Info("action", "n", int(action))` (actor.go:1051); all flipTo call sites lock-audited (see delta audit above); **TestActor_Serialization + TestActor_ENTransitUnchanged green `-race` this pass; the new TestActor_OwnFlipCreditConsumedOnce + TestActor_OwnFlipMixedWordCorrectsWhole (the delta's own behavior tests) green `-race` this pass** |

**Score:** 20/20 truths verified (0 present-but-behavior-unverified)

Behavior-dependent truths note: panic containment (#9), FSM window semantics
(#16), actor serialization (#20), EN transit contract (#6) each had their
single named behavioral test run green under `-race` **this pass** at HEAD
8292dd4 — the actor tests now run against the flipCredit-bearing code and
still pass. The delta's own state-transition behavior (credit armed on
successful round trip, consumed exactly once by the first FocusOut, never by
Reset) is exercised by the quick task's two named tests, also green `-race`
this pass. Live-bus truths (#5, #8, #10) rest on committed live-run evidence,
with the engine protocol layer untouched since the live-proven head.

### Advisory (New Scope, Unevidenced)

Re-verification ran; the Step 7 debt-marker scan over the 4 covered files
changed since the prior pass (actor.go, actor_test.go, ADR-004, SPEC.md) found
**zero** TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER markers and zero stub patterns.
No new-scope findings; advisory entries: none.

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| 1 | None — only carry-forward informational notes below (conn.go race semantics; e2e README wording; lint tooling notice) | — | — |

### Required Artifacts

No new artifacts this pass. All 55 covered paths re-checked for existence at
HEAD 8292dd4 — none missing. The 4 changed files remain substantive and wired
(actor.go: imported and driven by cmd/goswitchd/main.go:487/550; ADR-004 +
SPEC.md: the ADR↔spec delta cross-references are present in both directions).

### Key Link Verification

Manual wiring re-verification focused on links touching changed files:

- main.go → internal/session actor: `session.NewActor` (487), `SetVersion`
  (488), `Handler: actor` (550) — WIRED (unchanged, re-grepped)
- internal/session/actor.go → internal/engine lifecycle kinds:
  `engine.LifecycleFocusOut` / `engine.LifecycleReset` in HandleLifecycle —
  WIRED (compiles; `go build ./...` green this pass)
- internal/session/actor.go → switcher seam: `a.flipCredit++` only on the
  successful `sw(ctx, …)` round-trip branch (actor.go:2008-2017) — WIRED per
  ADR-004 amendment's mechanism paragraph
- docs/adr/ADR-004 → docs/SPEC.md §4.3: amendment cites the SPEC revision,
  SPEC §4.3 delta cites «ADR-004 amendment 2026-10-07» — WIRED (both directions)
- All other key links (main.go → internal/engine, engine → godbus, e2e →
  daemon, .golangci.yml → conn.go exclusion, mise.toml → layouts task,
  workflows → mise tasks, ADR-001 → d01 journal): files untouched since prior
  pass's WIRED verdicts — carry
- README.md → dist/systemd/user/goswitchd.service: unchanged
  designed-succession state — carried in frontmatter `deferred`

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| internal/session/actor.go (changed) | flipCredit | armed by flipTo's successful `sw()` round trip; consumed by HandleLifecycle FocusOut | ✓ TestActor_OwnFlipCreditConsumedOnce green `-race` this pass | ✓ FLOWING |
| internal/session/actor.go (changed) | action n / version / config snapshot | FSM Feed at timer expiry; version pinned at construction | ✓ TestActor_Serialization green at HEAD; action-log line intact (:1051) | ✓ FLOWING |
| internal/engine/conn.go, engine.go, layouts/tables.go, d01 journal (untouched) | addr / ev / tables / verdict rows | unchanged from prior pass's ✓ FLOWING verdicts (0 commits) | ✓ carry | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Build (current tree) | `go build ./...` | exit 0 | ✓ PASS |
| Actor serialization (Phase 1 contract) | `go test -race -count=1 -run 'TestActor_Serialization' ./internal/session/` | ok 1.0s | ✓ PASS |
| EN transit (Phase 1 semantics) | `go test -race -count=1 -run 'TestActor_ENTransitUnchanged' ./internal/session/` | ok 1.0s | ✓ PASS |
| Own-flip credit consumed once (delta behavior) | `go test -race -count=1 -run 'TestActor_OwnFlipCreditConsumedOnce\|TestActor_OwnFlipMixedWordCorrectsWhole' ./internal/session/` | ok 1.0s | ✓ PASS |
| Recover-shim behavior | `go test -race -count=1 -run 'TestEngine_RecoverContainsPanic' ./internal/engine/` | ok 1.0s | ✓ PASS |
| Address discovery | `go test -race -count=1 -run 'TestAddress' ./internal/engine/` | ok 1.0s | ✓ PASS |
| Golden punctuation tables | `go test -race -count=1 -run 'TestGolden_Punctuation' ./internal/layouts/` | ok 1.0s | ✓ PASS |
| FSM window expiry | `go test -race -count=1 -run 'TestFSM_SingleAtWindowExpiry' ./internal/hotkey/` | ok 1.0s | ✓ PASS |
| Lint on changed package | `mise exec -- golangci-lint run ./internal/session/...` | 0 issues | ✓ PASS |
| No device capture (re-widened) | grep over internal/{engine,layouts,hotkey,logging,session} + cmd/goswitchd | CLEAN | ✓ PASS |
| FSM purity | `grep -nE "go func\|chan \|time.Now\|Sleep" internal/hotkey/fsm.go` | no matches | ✓ PASS |
| flipCredit lock discipline | grep of all flipTo call sites + HandleLifecycle | 6/6 call sites hold a.mu; flipCredit referenced only under lock | ✓ PASS |
| ADR statuses | grep «Accepted — гейт M0» over docs/adr/ADR-00[1-5] | 5/5 | ✓ PASS |
| ADR-004 canonical headers | `grep -n "^## " docs/adr/ADR-004-buffer-reset-triggers.md` | 5 canonical + 1 dated amendment appendix | ✓ PASS |
| Journal verdict count | `grep -c "Verdict\|Вердикт" docs/adr/d01-experiment-log.md` | 15 (≥4) | ✓ PASS |
| Daemon wiring | grep of cmd/goswitchd/main.go | engine.Run (88), session.NewActor (487), SetVersion (488), Handler: actor (550) | ✓ PASS |
| Action-log oracle line | `grep -n 'slog.Info("action"' internal/session/actor.go` | line 1051 | ✓ PASS |
| Debt markers (changed covered files) | grep TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER over the 4 changed files | zero matches | ✓ PASS |
| Headless CI gate | `gh run list --workflow=pr-sanity.yml` | success at 1efc6fa (2026-10-06T22:50:22Z, newest run); post-pass delta covered by local build+lint+8 named `-race` tests | ✓ PASS (with CI-lag note) |
| Scheduled security scan | `gh run list --workflow=security-scheduled.yml` | success 2026-10-05 (weekly cadence) | ✓ PASS |
| Live m1/ibus-restart/kill9/d01 mise tasks | not re-run (drive owner's desktop) | committed Phase 1 evidence + intact assertions + untouched engine layer | ? SKIP (by policy) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` conventions in this repo. The runnable probes
are the mise tasks: headless CI gates ran green at 1efc6fa (newest pr-sanity
run); the verifier additionally ran `go build`, scoped golangci-lint, and the
eight named behavioral tests locally this pass (table above); live-session
e2e tasks excluded by the binding evidence policy (committed live proof
e2e-matrix 31/31 at the untouched engine head stands).

### Requirements Coverage

All 10 Phase 1 requirement IDs remain mapped to Phase 1 and Complete in
REQUIREMENTS.md traceability («Phase 1: 10»); no orphans. None of the 4
changed files alters any requirement's evidence base — statuses carry with
this pass's spot-check refresh:

| Requirement | Source Plan | Status | Evidence |
|-------------|------------|--------|----------|
| CORR-08 | 01-02 | ✓ SATISFIED | truths #14-15 (layouts untouched; golden test green this pass) |
| INTEG-01 | 01-01 | ✓ SATISFIED | truth #5; engine files untouched since live-proven head |
| INTEG-02 | 01-01, 01-03 | ✓ SATISFIED | truths #6, #11, #13; no-capture grep re-run CLEAN this pass |
| INTEG-03 | 01-01, 01-03, 01-05 | ✓ SATISFIED | truths #11-12; CI live, pr-sanity green at 1efc6fa |
| INTEG-04 | 01-01, 01-03, 01-04 | ✓ SATISFIED | truth #8 |
| INTEG-05 | 01-01, 01-03 | ✓ SATISFIED | truths #9-10 (recover test green this pass) |
| TEST-01 | 01-02, 01-03, 01-05 | ✓ SATISFIED | truths #16-18 (FSM test green this pass) |
| TEST-02 | 01-03, 01-04 | ✓ SATISFIED | truth #7; stand untouched, live-proven |
| TEST-03 | 01-03 | ✓ SATISFIED | truth #7 (focus_helper.py untouched) |
| INST-04 | 01-01, 01-05 | ✓ SATISFIED | truths #17, #19 |

### Decision Coverage

01-CONTEXT.md is untouched since prior passes; the decision set has not
changed. The tool-level `check decision-coverage-verify` gate reported
`skipped: CONTEXT.md missing` under current gsd-tools path resolution in
prior passes — warning-only, non-blocking; nothing about the decision corpus
changed since. (The ADR-004 amendment documents a NEW owner decision, G-2-3,
made at milestone close under the quick-task workflow — it amends the ADR's
contract via the owner-sanctioned spec-delta pattern and belongs to the
milestone-close record, not to Phase 1's original decision corpus.)

### Anti-Patterns Found

Debt-marker scan re-run over the 4 covered files changed since the prior
pass: **ZERO** TBD/FIXME/XXX, zero TODO/HACK/PLACEHOLDER, zero stub patterns
(including the new flipCredit code — the credit branch has no placeholder
paths; the WARN and nil-seam degradation branches are real handling, not
stubs). Carry-forward warnings (files untouched this pass, still non-blocking
quality debt documented in prior passes): boot-time retry race semantics
(internal/engine/conn.go), NameFlagReplaceExisting inertness, length-only
oracle limitation in locked-session d01 re-runs.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| internal/engine/conn.go | (untouched) | carry-forward boot-time retry race + rapid-respawn race (documented quality debt; gosec G404 exclusion target) | ⚠️ Warning | Canonical paths demonstrably green (31/31 live matrix at untouched engine head); hardening debt for future work |
| lint tooling | — | golangci-lint `exhaustruct` deprecation notice (v2.13+ → exhaustruct_v5) — re-observed this pass on the scoped run, 0 issues | ℹ️ Info | Cosmetic future config rename |
| test/e2e/README.md | 220 | INTEG-02 section carries the Phase 1 wording «возвращающий `false` (наблюдатель)» — predates the consumption succession; architectural claim remains true | ℹ️ Info | Documentation nuance only |
| README.md | install section | documents `goswitchctl install` instead of the plan-01-05 manual unit-copy dev instruction | ℹ️ Info | Designed succession (see `deferred`); SC-level no-root contract intact |
| CI freshness | — | newest pr-sanity run predates the quick-261008-00m code commits; local build + lint + 8 named `-race` tests cover the delta this pass | ℹ️ Info | Next push/PR run of pr-sanity closes the gap mechanically |

### Human Verification Required

None new. The four environment-gated items from the original pass were
resolved by UAT (01-UAT.md: 4/4 pass, owner-accepted 2026-09-11) and are not
re-opened. No ⚠️ PRESENT_BEHAVIOR_UNVERIFIED truths exist on this pass — every
behavior-dependent truth had its named test run green this pass, including the
delta's own credit state-transition tests. The G-2-3 live acceptance itself
(owner UAT test 3, matrix v1 16/16 live at the quick task) is the owner's own
completed live verification of the amended behavior. Live-session e2e re-runs
remain policy-excluded.

### Gaps Summary

**No gaps.** All 20 must-have truths verified on the current tree at HEAD
8292dd4. The staleness introduced by quick task 261008-00m is fully resolved:

1. The ADR-004 amendment is a dated, additive, owner-decision record in the
   exact D-55 spec-delta pattern; the ADR pack contract (5 ADRs Accepted,
   unified format, journal trail) re-verified intact.
2. The actor.go flipCredit delta preserves every Phase 1 actor seam:
   serialization (mutex at every entry and every flipTo call site, race-clean),
   the `{"msg":"action","n":N}` expiry oracle (actor.go:1051), daemon wiring
   (main.go:88/487-488/550), and the no-capture guarantee (grep CLEAN
   re-widened over the changed package). Named regression tests green under
   `-race`, plus the delta's own two behavior tests.
3. covered_files re-derived (set unchanged, 55/55 exist) and the fingerprint
   recomputed over the current tree.

Zero debt markers, zero orphans, zero regressions, zero human-verification
items. Phase 1 remains a sound foundation on the milestone-closed tree.

---

_Verified: 2026-10-07T23:10:20Z_
_Verifier: Claude (gsd-verifier)_
