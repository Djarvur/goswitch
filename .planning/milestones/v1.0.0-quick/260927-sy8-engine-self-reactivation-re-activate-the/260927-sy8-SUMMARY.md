---
phase: quick-260927-sy8
plan: 01
subsystem: daemon/ibus-engine
tags: [ibus, reactivation, recovery, sy8, daemon-restart]
requires:
  - engine.Run connect→register→serve loop (engine/conn.go)
  - gsettings input-sources wire shape (org.gnome.desktop.input-sources sources/current)
  - install.go activateEngine precedent (`ibus engine <name>` is the honored activation path)
provides:
  - internal/activate package — IfOwned ownership check + activation + bounded retry (void contract, never fatal)
  - engine.Config.PostRegister hook — fires after every successful component (re)registration
  - daemon self-reactivation wired in cmd/goswitchd engineConfig
  - README (EN+RU) recovery contract for daemon restarts and the Super+Space caveat
affects:
  - Phase 4 UAT (unattended daemon restarts now restore the owned engine)
  - nightly D-48 gate (unit restart no longer silently disables correction)
tech-stack:
  added: [] # stdlib only (os/exec, regexp, strconv, log/slog) — zero new dependencies
  patterns:
    - Runner subprocess seam (lean 2-arg form of the install.go Runner precedent)
    - strict GVariant text parse with malformed→skip semantics
    - optional callback seam on engine.Config for post-registration work
key-files:
  created:
    - internal/activate/activate.go
    - internal/activate/activate_test.go
  modified:
    - engine/conn.go
    - cmd/goswitchd/main.go
    - README.md
    - README.ru.md
decisions:
  - "engineConfig signature kept unchanged: the PostRegister closure takes its ctx from serve's per-generation context (the daemon signal lineage) — no plumbing through run()"
  - "internal/activate does NOT import internal/install: the Runner seam is deliberately re-defined lean (no env/stdin — `ibus engine` needs neither); duplication is one exec call, accepted per plan"
  - "Retry knobs are unexported package vars (activationAttempts/activationRetryDelay) overridable by the corpus with t.Cleanup restore; package tests are never parallel, keeping -race deterministic"
  - "parseCurrent returns a static sentinel (errMalformedCurrent) wrapped with the raw value — err113-clean; malformed current value appears in the DEBUG log field"
metrics:
  duration: 19 min
  completed: 2026-09-27
  tasks: 3
  files: 6
status: complete
commits: 4
plan_head_before: b331bd49449f9648dbfbc4e42de2311042c5652f
actuals:
  tokens: 6100 # chars/4 over the realized diff (24414 chars, 598 insertions / 1 deletion, 6 files)
  tasks: 3
  commits: 4 # MEASURED: git rev-list --count b331bd4..HEAD
---

# Quick Task 260927-sy8: Engine Self-Reactivation Summary

**IfOwned + PostRegister: the daemon re-activates its own goswitch engine after every (re)registration when goswitch owns the current GNOME input source — a `systemctl --user restart goswitchd` no longer silently kills correction.**

## What Was Built

1. **`internal/activate`** (new, strict TDD): `IfOwned(ctx, run)` reads
   `gsettings get org.gnome.desktop.input-sources sources|current`, parses both strictly
   (single-quoted 2-tuple GVariant grammar; `uint32 <N>`), and activates
   `ibus engine <current-name>` ONLY when the current index points at a `goswitch-` engine.
   Up to 3 attempts, 500 ms apart, ctx-aware (never sleeps past cancellation). INFO
   `engine reactivated` on success, DEBUG for skips/attempts, WARN on exhausted retries —
   void contract, zero fatal paths (T-SY8-01..03 mitigations in place: argv-form exec,
   10 s per-call CommandContext timeout, gsettings-absent degradation).
2. **`engine.Config.PostRegister`** (glue): invoked synchronously in `serve` after the
   "component registered"/"re-registered" log on EVERY generation, before `waitBusLoss`;
   nil = no-op (zero behavior change for existing callers, no existing tests touched).
3. **Daemon wiring**: `engineConfig` sets `PostRegister` to
   `activate.IfOwned(ctx, activate.NewExecRunner())` — ctx comes from serve's generation
   context, so `engineConfig`'s signature stayed unchanged.
4. **README.md + README.ru.md**: Troubleshooting gains the self-reactivation bullet
   (journal line `engine reactivated`) and the Super+Space single-source caveat —
   documented, NOT fixed; remedy `ibus engine goswitch-en` or unit restart.

## Tasks

| Task | Name | Commit(s) | Files |
| ---- | ---- | --------- | ----- |
| 1 | internal/activate — owned-source check, activation, bounded retry (strict TDD) | b0274fd (RED) → 5a1ed0a (GREEN+REFACTOR) | internal/activate/activate.go, activate_test.go |
| 2 | PostRegister hook — engine.Config + daemon wiring | 37cfd9b | engine/conn.go, cmd/goswitchd/main.go |
| 3 | README recovery contract (EN + RU) | c543d2b | README.md, README.ru.md |

## Verification

- Task 1: `go test -race -count=1 ./internal/activate/` — 14-case corpus green (owned
  single/mid-list → exactly one activation of the CURRENT source; foreign xkb/ibus,
  out-of-range, malformed, gsettings-failure → zero ibus calls; exhausted retries warn
  after exactly `activationAttempts` calls; flaky retry succeeds; pre-cancelled ctx →
  zero calls, no sleeps). `mise run ci` green.
- Task 2: `mise run ci` + `go test -race -count=1 ./engine/ ./cmd/goswitchd/` green with
  zero edits to existing tests; grep confirms the hook fires after the registration log.
- Task 3: grep gate (`engine reactivated` present in both READMEs) + `mise run ci` green.
- `mise run ci` (build + vet + golangci-lint v2 strict + `go test -race -count=1 ./...`)
  green after every task — engineering directives 1-2 held.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Corpus stub indexed an empty error slice**
- **Found during:** Task 1 GREEN (first run)
- **Issue:** `stubDesktop` computed `min(attempt, len(ibusErrs)-1)` → `[-1]` panic when
  called with no ibus error script (the owned-path cases).
- **Fix:** empty `ibusErrs` short-circuits to success before the index math.
- **Files modified:** internal/activate/activate_test.go
- **Commit:** 5a1ed0a (folded into the GREEN commit)

**2. [Rule 1 - Lint] REFACTOR pass tightened three strict-lint findings**
- **Found during:** Task 1 REFACTOR
- **Issue:** err113 (dynamic error in parseCurrent), lll (nolint comment 139 chars),
  revive context-as-argument vs thelper t-first conflict in the runIfOwned helper.
- **Fix:** static `errMalformedCurrent` sentinel wrapped with the raw value; shortened
  nolint comment; runIfOwned lost its ctx param (background ctx inside) and the
  cancelled-context test grew its own inline body.
- **Files modified:** internal/activate/activate.go, internal/activate/activate_test.go
- **Commit:** 5a1ed0a

No other deviations — plan executed as written.

## Known Stubs

None. No placeholder code, no unwired data paths.

## Threat Flags

None — the only new trust-boundary surface (gsettings output → parser → exec argv) is
exactly the plan's T-SY8-01/02/03 and carries its prescribed mitigations (strict grammar
+ prefix gate + range check before exec; argv form, no shell; per-call timeout; void
degradation).

## Human-Check Status (Task 3)

- Automated parts DONE: grep gates, README edits (EN+RU), `mise run ci`.
- **Live desktop restart-and-type check: DEFERRED TO ORCHESTRATOR** (per run instructions):
  `systemctl --user restart goswitchd` → type `ghbdtn` without `ibus engine` → expect
  `привет` and journal `component registered` → `engine reactivated`; then the
  foreign-source guard (xkb current + restart → no activation). Recorded in the
  WINDOWS ledger as open (`unrun-verify`, quick-260927-sy8) — close it after the live run.

## Self-Check: PASSED

- Files exist: internal/activate/activate.go, internal/activate/activate_test.go — FOUND
- Commits exist: b0274fd, 5a1ed0a, 37cfd9b, c543d2b — all FOUND on gsd/phase-04-postavka-i-priemka
- `git rev-list --count b331bd4..HEAD` = 4 (matches frontmatter `commits: 4`)
- No file deletions in the plan diff (598 insertions, 1 deletion across 6 files)
