---
phase: quick-260927-sy8
verified: 2026-09-27T22:05:00Z
status: human_needed
score: 3/4 must-haves verified
covered_files:
  - ".planning/quick/260927-sy8-engine-self-reactivation-re-activate-the/260927-sy8-PLAN.md"
  - ".planning/quick/260927-sy8-engine-self-reactivation-re-activate-the/260927-sy8-SUMMARY.md"
  - "README.md"
  - "README.ru.md"
  - "cmd/goswitchd/main.go"
  - "engine/conn.go"
  - "internal/activate/activate.go"
  - "internal/activate/activate_test.go"
covered_digest: "v1:sha256:50d8d276dbe1082cfe58e36d88fca869a4a604306800b6d310dc1c7ad883c285"
behavior_unverified: 1
overrides_applied: 0
re_verification: none
behavior_unverified_items:
  - truth: "After `systemctl --user restart goswitchd` (ibus-daemon already up), the daemon re-activates the goswitch engine by itself when goswitch owns the current GNOME input source — keystrokes route through the engine without a manual `ibus engine` call"
    test: "On the live GNOME desktop: `systemctl --user restart goswitchd`, then immediately type `ghbdtn` in any text field WITHOUT running `ibus engine` by hand; then `journalctl --user -u goswitchd --since -2min`"
    expected: "Correction produces `привет` within the first word or two; journal shows `component registered` followed by `engine reactivated`"
    why_human: "The full transition (real ibus-daemon accepting SetGlobalEngine and GNOME routing keystrokes through the engine) only exists on the live desktop — the unit corpus proves the decision logic at the Runner seam, no test on this machine exercises the live registration→activation→routing chain. Known-open WINDOWS #8 (unrun-verify), deferred to orchestrator per run instructions"
coincidental_reliance_items: []
human_verification:
  - test: "Live owned-path restart: `systemctl --user restart goswitchd` on the running desktop, type `ghbdtn` without manual `ibus engine`; check `journalctl --user -u goswitchd --since -2min`"
    expected: "`привет` appears within the first word or two; journal shows `component registered` then `engine reactivated`"
    why_human: "Live ibus-daemon + GNOME keystroke routing is unobservable to automated checks on this machine; this is the deferred Task 3 human-check (WINDOWS #8)"
  - test: "Live foreign-source guard: switch the GNOME current input source to a plain xkb source with goswitch NOT current, `systemctl --user restart goswitchd`, observe the active engine"
    expected: "The active engine is unchanged — no goswitch activation; journal records a skip entry (DEBUG-level; the level-independent observable is that the active source did not change)"
    why_human: "Same live-bus limitation; confirms the no-hijack guard end-to-end beyond the unit corpus"
---

# Quick Task 260927-sy8: Engine Self-Reactivation — Verification Report

**Task Goal:** engine self-reactivation — re-activate the global IBus engine on daemon (re)registration when goswitch owns the current input source.
**Verified:** 2026-09-27T22:05:00Z
**Status:** human_needed (all automated deliverables verified; the plan's own Task 3 live desktop check remains pending human confirmation — WINDOWS #8, not counted against the code)
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | After daemon restart, goswitch re-activates its engine by itself when it owns the current input source — keystrokes route through without manual `ibus engine` | ⚠️ PRESENT_BEHAVIOR_UNVERIFIED | Present + wired + unit-proven: `internal/activate` corpus (14 cases, `go test -race` green) proves owned → exactly one `ibus engine <current-name>`; `engine/conn.go:132-133` fires `cfg.PostRegister(ctx, generation)` on EVERY generation after the registration log; `cmd/goswitchd/main.go:160-162` wires `activate.IfOwned`. The live end-to-end transition (real ibus-daemon + GNOME routing) is the deferred Task 3 human-check — WINDOWS #8. See Human Verification |
| 2 | When the current input source is NOT goswitch (xkb or foreign ibus engine), (re)registration never activates anything | ✓ VERIFIED | `TestIfOwnedForeignCurrentSkipped` (xkb current, foreign-ibus current → zero ibus calls) passes under -race; code gate `internal/activate/activate.go:128-139`: index range check + `goswitch-` prefix check before any exec |
| 3 | gsettings/ibus unavailability degrades to WARN/DEBUG logs; the daemon keeps serving and never exits on reactivation failure | ✓ VERIFIED | `IfOwned` is void with zero fatal paths (no error return, no `os.Exit`/`log.Fatal`/panic in the package); `TestIfOwnedGSettingsFailureSkipped` (both keys fail → normal return, zero ibus calls) passes; retry exhaustion lands in `slog.Warn` (activate.go:227) and returns; serve continues to `waitBusLoss` (conn.go:136) |
| 4 | Super+Space single-source engine-context disable is documented (NOT fixed) with `ibus engine goswitch-en` as remedy, in both READMEs | ✓ VERIFIED | README.md:168-181 and README.ru.md:170-183, both under `## Troubleshooting` (EN line 155, RU line 157): self-reactivation bullet naming the `engine reactivated` journal line + Super+Space caveat explicitly stated as not-live-detected, remedy `ibus engine goswitch-en`, `journalctl --user -u goswitchd` pointer |

**Score:** 3/4 truths verified (1 present, behavior-unverified — the live desktop restart check)

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/activate/activate.go` | Ownership check + activation + bounded retry, subprocess seam, zero fatal paths | ✓ VERIFIED | 242 LOC. Exact plan surface: `Runner` seam (activate.go:72), `NewExecRunner` (10 s `CommandContext` timeout, argv form, stderr folded into error — :77-93), `IfOwned` void contract (:101-142), strict GVariant tuple grammar (:66, :153-175), `uint32 <N>` current parse (:184-196), bounded retry 3×500 ms ctx-aware (:202-228), log levels INFO/DEBUG/WARN per plan |
| `internal/activate/activate_test.go` | Fake-runner table corpus covering owner/foreign/malformed/retry/cancel | ✓ VERIFIED | 302 LOC, 14 cases — every plan behavior case present: owner single + mid-list (activates CURRENT source `goswitch-ru`, not always `-en`), foreign xkb, foreign ibus, out-of-range, garbage + empty sources, negative/non-numeric/empty current, gsettings down on both keys, retry exhausted (exactly 3 calls → WARN path), flaky retry (2 calls), pre-cancelled ctx (0 calls, wall-clock bounded < 400 ms) |
| `engine/conn.go` | `engine.Config.PostRegister` invoked after every successful (re)registration | ✓ VERIFIED | Field at :37-43 with the plan's doc contract (cheap, non-fatal, inline before waitBusLoss); invoked at :132-133 after the `component registered`/`re-registered` log block (:127-131) and before `waitBusLoss` (:136); `Run` loops `for generation := 0; ; generation++` so generation 0 AND every reconnect generation fire it |
| `cmd/goswitchd/main.go` | PostRegister wired to `activate.IfOwned` with the exec runner | ✓ VERIFIED | Import `internal/activate` at :19; `engineConfig` sets `PostRegister: func(ctx){ activate.IfOwned(ctx, activate.NewExecRunner()) }` (:160-162) with the idempotent-double-activation comment (:148-149); the ONLY production `engine.Config{` construction site |
| `README.md` | Recovery contract under Troubleshooting | ✓ VERIFIED | Both bullets present, `engine reactivated` + `ibus engine goswitch-en` remedy + journalctl pointer |
| `README.ru.md` | Mirror of EN contract | ✓ VERIFIED | Content parity confirmed (RU lines 170-183 mirror the EN bullets) |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | -- | --- | ------ | ------- |
| `engine.serve` (after "component registered"/"re-registered" log) | `cfg.PostRegister(ctx, generation)` → `activate.IfOwned` → `ibus engine <goswitch-*>` | inline call before `waitBusLoss` | ✓ WIRED | conn.go:132-133 → main.go:160-161 → activate.go:141/210 (`run(ctx, "ibus", ["engine", name])`, name prefix-gated). Fires on generation 0 and reconnect generations (`Run` loop, conn.go:55) |
| `cmd/goswitchd` `engineConfig` | `PostRegister` closure → `activate.NewExecRunner` (stdlib os/exec, argv form) | closure over daemon signal ctx lineage | ✓ WIRED | main.go:160-162 → activate.go:77-93 (`exec.CommandContext(ctx, name, args...)` — argv form, no shell; per-call 10 s timeout) |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| `activate.IfOwned` | sources/current → engine name | `gsettings get org.gnome.desktop.input-sources sources\|current` via `NewExecRunner` in production; recording fake through the `Runner` seam in tests | Yes (real subprocess in production wiring) | ✓ FLOWING |

Tampering surface (T-SY8-01) verified mitigated: tuple regex `\('([^']*)',\s*'([^']*)'\)` admits no quote into the captured name, `goswitch-` prefix + index-range gate precede exec, argv form cannot escape into a second command.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Package builds | `go build ./...` | success | ✓ PASS |
| Activate corpus (owned/foreign/malformed/retry/cancel) | `go test -race -count=1 ./internal/activate/` | ok 1.018s | ✓ PASS |
| Engine + daemon suites untouched and green | `go test -race -count=1 ./engine/ ./cmd/goswitchd/` | ok 1.018s / ok 1.013s | ✓ PASS |
| Full CI gate (build + vet + golangci-lint v2 strict + `go test -race -count=1 ./...`) | `mise run ci` | all 14 packages ok | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| N/A | — | no probe-*.sh declared or conventional for this task; the CI gate above is the runnable check | N/A |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| SY8 | 260927-sy8-PLAN (`requirements: [SY8]`) | Task-local requirement: daemon re-activates its owned engine after every (re)registration; foreign sources never hijacked; degradation never fatal; Super+Space caveat documented | ✓ SATISFIED (code level; live confirmation pending per truth 1) | Truths 2-4 verified; truth 1 verified at code+unit level with the live check open |
| INTEG-04 (adjacent) | .planning/REQUIREMENTS.md | Engine re-registers after ibus-daemon restart (ibus#2910) | ✓ STRENGTHENED | Re-registration now also re-activates the owned source — no regression: engine suite green, zero edits to existing tests (commit 37cfd9b touched only conn.go + main.go) |
| SWCH-03 (adjacent) | .planning/REQUIREMENTS.md | Native Super+Space keeps working | ✓ DOCUMENTED | Super+Space single-source caveat documented, not fixed, exactly per plan scope |

No orphaned requirements: quick tasks are not phase-mapped in REQUIREMENTS.md, and no REQUIREMENTS.md ID maps to this task unclaimed. The SY8 label exists only inside the quick task (plan objective is its definition); no REQUIREMENTS.md/SPEC.md entry carries that literal ID.

### Commits

| Commit | Claim | Verified |
| ------ | ----- | -------- |
| b0274fd | test(04): activate package corpus (RED) | ✓ exists; adds ONLY activate_test.go (302 lines) — genuine red-first (in-package test referencing not-yet-written symbols) |
| 5a1ed0a | fix(04): internal/activate (GREEN+REFACTOR) | ✓ exists; activate.go 242 lines + test fixes |
| 37cfd9b | feat(04): PostRegister hook + daemon wiring | ✓ exists; conn.go +13, main.go +10, no test edits |
| c543d2b | docs(04): README recovery contract | ✓ exists; README.md +13, README.ru.md +14 |

All on `gsd/phase-04-postavka-i-priemka`, conventional style with `(04)` scope and `(SY8)` refs — matches plan. Working tree: no source files modified since the commits (only planning docs).

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
| --------- | ---------- | ------ | ------- | -------- | --------------- | ------- |
| internal/activate/activate_test.go | SY8 | 14 | 0 | 0 | Behavioral (exact subprocess call sequences; wall-clock bound on cancellation) | VALID |

Disabled tests on requirements: 0. Circular patterns: 0 (fake runner is an independent oracle). `t.Parallel` absent per plan (package-global retry knobs + t.Cleanup restore keep -race deterministic).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | No TBD/FIXME/XXX/TODO/HACK/placeholder markers, no empty implementations, no stub returns in any of the 6 modified files | — | — |

### Human Verification Required

### 1. Live owned-path restart (the original defect's fix proof)

**Test:** On the live GNOME desktop: `systemctl --user restart goswitchd`, then immediately type `ghbdtn` in any text field WITHOUT running `ibus engine` by hand. Then `journalctl --user -u goswitchd --since -2min`.
**Expected:** Correction produces `привет` within the first word or two; journal shows `component registered` followed by `engine reactivated`.
**Why human:** Real ibus-daemon + GNOME keystroke routing is unobservable to automated checks on this machine; the unit corpus proves the decision logic at the Runner seam only. This is the Task 3 human-check explicitly deferred to the orchestrator (WINDOWS #8, `unrun-verify`) — pending, not failed.

### 2. Live foreign-source guard

**Test:** Switch the GNOME current input source to a plain xkb source with goswitch NOT current; `systemctl --user restart goswitchd`; observe the active engine.
**Expected:** Active engine unchanged — no goswitch activation; journal records a skip entry (DEBUG level; the level-independent observable is the unchanged active source).
**Why human:** Same live-bus limitation; confirms the no-hijack guard end-to-end beyond the unit corpus.

## Gaps Summary

No gaps. All six artifacts exist, are substantive, and are wired end to end; both key links trace cleanly; the full CI gate (build + vet + golangci-lint v2 strict + `go test -race -count=1 ./...`, 14 packages) is green on the commit head; all four commits verified on the declared branch with genuine red-first TDD sequencing. The single non-verified truth is behavior-dependent by nature: the live desktop restart transition was deferred by the run itself to the orchestrator and stands open as WINDOWS #8 (`unrun-verify`) — per instructions it is recorded as pending human confirmation, not as a failure. Status `human_needed` reflects exactly that: automated checks passed, two live checks await the human (the second, the foreign-source guard, is this verifier's addition for end-to-end completeness).

---

_Verified: 2026-09-27T22:05:00Z_
_Verifier: Claude (gsd-verifier)_
