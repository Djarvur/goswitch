---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
plan: 06
subsystem: input-correction
tags: [ctl, dbus, session-bus, goswitchctl, inst-02, d-32, tdd, e2e]

# Dependency graph
requires:
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 02
    provides: config schema + Watcher (Snapshot/LastError, strict decode D-33), -config flag wiring
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 04
    provides: applySnapshot one-read-per-event succession, live case scaffolding (restartDaemonWithArgs, waitForLog)
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 05
    provides: MACRCounters{SuperIntercepted, ConsumedUpstream} — the fixed goswitchctl counter surface
provides:
  - internal/ctlsvc: Run(ctx, deps) — the daemon's SECOND godbus connection owning org.djarvur.goswitch (primary-owner guard, exported ErrNotPrimaryOwner) with Status/ReloadConfig/CorrectNow exported at /org/djarvur/goswitch, recoverMethod shim on every method
  - internal/session: Status type + Actor.StatusSnapshot() (mode/correction counters/skip reasons/MACR counters/config status) and Actor.CorrectNow() (the Double dispatch point, no tap, immediate reply)
  - internal/session: correction outcome counters wired into EVERY done/skipped path (D-24 counts done)
  - internal/config: Watcher.Reload() (synchronous forced re-read sharing the debounce core) + Watcher.ConfigPath()
  - cmd/goswitchctl: status (--json) / reload / correct; "daemon not running?" verdict; exit-1 contract
  - cmd/goswitchd: newActor split + startCtl on the signal ctx (ctl failure logged, never fatal)
  - test/e2e: ctl-smoke case + mise task e2e-ctl
affects: [03-07-matrix-v2]

actuals:
  tokens: 18370   # chars/4 over the realized diff (git diff c0956e6..HEAD = 73479 chars)
  tasks: 2
  commits: 3       # measured: git rev-list --count c0956e6..HEAD
  plan_head_before: c0956e6832a6c0c01f6298f379317f974c9e7efc

tech-stack:
  added: []
  patterns:
    - "key=value status canon: ONE D-Bus string is both the human report and the --json source (config_error always LAST with a whitespace-flattened value, skip reasons dash-flattened) — no properties, no signals, the plan's FLAGGED minimal-methods assumption held live"
    - "private dbus-daemon spawn (temp config-file + --print-address=1 + --nofork, killed by the test's cleanup context) as the hermetic bus of the name-guard corpus — RequestName semantics and the wire-level method call are unit-pinned without touching the live session"
    - "ReplaceExisting WITHOUT AllowReplacement is the real single-instance semantics on the session bus: a second live instance answers EXISTS/IN_QUEUE (never PrimaryOwner), and a dead connection's name auto-release covers the restart — the conn.go flag set reused verbatim"

key-files:
  created:
    - internal/ctlsvc/ctlsvc.go
    - internal/ctlsvc/ctlsvc_test.go
    - cmd/goswitchctl/main.go
    - test/e2e/case_ctl.go
    - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/red-evidence/03-06-task1-red.json
  modified:
    - internal/session/actor.go
    - internal/session/actor_test.go
    - internal/config/watch.go
    - cmd/goswitchd/main.go
    - test/e2e/main.go
    - mise.toml

key-decisions:
  - "The Status wire contract is a single-line key=value string: human-readable as printed, machine-parsed by goswitchctl's --json client-side — the daemon exposes no second method; config_error is rendered LAST and whitespace-flattened so a multi-line parse error cannot break the token grammar"
  - "ctlsvc's deps are three single-method interfaces at the point of use (StatusSnapshotProvider/Reloader/Corrector); the package imports internal/session ONLY for the Status struct in the interface signature — the plan's literal \"no session types\" wording was resolved against its own key_links (ctl→session is the prescribed direction), adapters duplicating the nine fields would add no decoupling"
  - "ErrNotPrimaryOwner is EXPORTED (the plan named conn.go's private errNotPrimaryOwner): the errors.Is pin lives in ctlsvc_test, an external test package — and the guard is real: our RequestName carries no AllowReplacement, so a second live instance can never replace us"
  - "Correction outcome counters live in the actor as skipCorrection()/countCorrectionDone() helpers replacing every direct \"correction skipped\" slog site; the D-24 success-without-change counts DONE (outcome done either way) — the full reason vocabulary (empty-buffer, no-letters, convert-failed, no-engine, verify-mismatch, verify-timeout, backspace-cap, clipboard-unavailable) is countable from outside"
  - "A ctl-service failure NEVER kills the daemon: startCtl logs the error and the daemon keeps serving input through the engine connection (the appid degradation precedent — the desktop's input rides the process)"
  - "ReloadConfig answers a D-Bus error on rejection (the CLI exits 1 with the rejection text naming the key) and on the no-config daemon (\"no config file\") — the exit contract treats both as failures worth signaling, while Status stays the always-succeeding diagnostic"

requirements-completed: [INST-02]

coverage:
  - id: D1
    description: "The session-bus control service: org.djarvur.goswitch owned with the primary-owner guard (a second Run answers ErrNotPrimaryOwner), the object at /org/djarvur/goswitch serving Status/ReloadConfig/CorrectNow over the wire, every method behind a recover shim that converts a panic into a *dbus.Error with the daemon still answering"
    requirement: INST-02
    verification:
      - kind: unit
        ref: internal/ctlsvc/ctlsvc_test.go#TestSvc_NameGuard
        status: pass
      - kind: unit
        ref: internal/ctlsvc/ctlsvc_test.go#TestSvc_RecoverShim
        status: pass
      - kind: e2e
        ref: mise run e2e-ctl (PASS ×4 on the final tree — the daemon holds the name and answers all three subcommands live)
        status: pass
    human_judgment: false
  - id: D2
    description: "goswitchctl: status (human key=value + --json single-line typed JSON), reload, correct; the friendly 'daemon not running?' verdict when the name has no owner; exit ≠ 0 on every failure (broken reload, unreachable daemon, unknown subcommand)"
    requirement: INST-02
    verification:
      - kind: e2e
        ref: mise run e2e-ctl (PASS ×4 — steps 1/2/4 pin the outputs live)
        status: pass
      - kind: command
        ref: "live pin: goswitchctl status (and status --json) with no daemon → 'daemon not running?' + exit 1; unknown subcommand → usage + exit 1; only third-party import is godbus/dbus (stdlib flag/os.Args — no CLI framework)"
        status: pass
    human_judgment: false
  - id: D3
    description: "D-32 closed end to end: a broken document's forced reload answers with the error naming the unknown key and exits non-zero; the LAST-GOOD ERROR stays visible in status (config_valid=false + the error text) WHILE the daemon keeps correcting under the last-good snapshot; the restored config reloads with applied and the status turns valid"
    requirement: INST-02
    verification:
      - kind: unit
        ref: internal/ctlsvc/ctlsvc_test.go#TestSvc_ReloadInvalidKeepsLastGood
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_StatusConfigFields
        status: pass
      - kind: e2e
        ref: mise run e2e-ctl (steps 2–4: last-good visible in status, forced correction done under the broken on-disk document, restored reload applied)
        status: pass
    human_judgment: false
  - id: D4
    description: "The status snapshot's counters: mode EN/RU, corrections done/skipped with the skip-reason breakdown, the MACR super counters (ADR-005 b.2 surface) — counts and states only, never typed or corrected text (T-03-06-03)"
    requirement: INST-02
    verification:
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_StatusSnapshot
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_StatusCounters
        status: pass
      - kind: unit
        ref: internal/ctlsvc/ctlsvc_test.go#TestSvc_Status
        status: pass
      - kind: e2e
        ref: mise run e2e-ctl (corrections_done=1 in the status right after the settled forced correction)
        status: pass
    human_judgment: false
  - id: D5
    description: "CorrectNow — the forced word correction (Q5 word semantics): the Double decision's internal point launched with no tap and no FSM round trip, immediate acknowledgment, the two-phase ADR-004 discipline intact (nothing destructive before the surrounding verdict), settlement in the counters/log"
    requirement: INST-02
    verification:
      - kind: unit
        ref: internal/ctlsvc/ctlsvc_test.go#TestSvc_CorrectNow
        status: pass
      - kind: e2e
        ref: mise run e2e-ctl (step 3: correction done record, the entry settles at привет, both oracles)
        status: pass
    human_judgment: false

duration: 22 min
completed: 2026-09-15
status: complete
---

# Phase 3 Plan 6: Управление демоном — internal/ctlsvc + goswitchctl Summary

**Контрольная поверхность демона на session bus: internal/ctlsvc владеет org.djarvur.goswitch (primary-owner guard, ErrNotPrimaryOwner, recover-шим на каждом методе — паника метода не убивает демон), goswitchctl гоняет status (key=value + --json) / reload / correct против живого демона, last-good ошибка D-32 видна в status ПОКА демон продолжает корректировать на последнем валидном конфиге, счётчики коррекций и MACR-перехватов — в снапшоте актора (никакого текста коррекций, T-03-06-03).**

## Performance

- **Duration:** 22 min (13:36–13:58 UTC)
- **Tasks:** 2 (1 tdd: RED→GREEN with RED_EVIDENCE_OK, 1 auto)
- **Files modified:** 11 (10 code/config + 1 red-evidence record)
- **Commits:** 3 (measured from plan_head_before c0956e6)

## Accomplishments

- **internal/ctlsvc (Task 1):** the daemon's second godbus connection on the session bus — `Run(ctx, deps)` takes org.djarvur.goswitch with NameFlagReplaceExisting and refuses anything but RequestNameReplyPrimaryOwner (the conn.go:93-101 guard verbatim; no AllowReplacement on our side, so a second live instance can never replace us — the guard is semantic, not decorative), exports `Svc` at /org/djarvur/goswitch with Status/ReloadConfig/CorrectNow, each opening with `recoverMethod` (INTEG--05 continuation: a panic below a control call becomes an org.freedesktop.DBus.Error.Failed reply + an ERROR log with the stack — the daemon keeps serving desktop input). Deps are three single-method interfaces; Reload may be nil (the no-config daemon answers "no config file").
- **Status canon:** `renderStatus` — one single-line key=value string: mode, corrections_done/skipped, one `skip_<reason>` token per reason (dashes flattened, sorted), super_intercepted, super_upstream_consumed, then the config triple with config_error ALWAYS last and whitespace-flattened (a multi-line yaml parse error cannot break the grammar); no-config renders `config=none`. No typed or corrected text anywhere (T-03-06-03 — the D-20 prohibition extended to the control surface).
- **session.Status + counters (Task 1):** `Actor.StatusSnapshot()` fills under the mutex from the counters the actor now keeps — `skipCorrection(reason)` (counts + the existing INFO record) replaced every direct "correction skipped" site across the word/phrase/selection/level-2/clipboard paths; `logCorrectionDone` became a counting method; the D-24 success-without-change counts done. `Actor.CorrectNow()` — the Double dispatch point with one config snapshot folded in, immediate `correctStartedReply`, settlement in the counters/log (the async contract documented on the method). The D-32 config fields come from an optional `configStatus` seam type-asserted on the attached source (the watcher grew `ConfigPath()`; `LastError()` existed since 03-02).
- **Watcher.Reload:** synchronous forced re-read sharing the debounce path's parse-and-publish core under the same reloadMu — publish ("applied (tap_window_ms=N)") or last-good + the error (returned AND visible in status).
- **goswitchctl + daemon wiring (Task 2):** stdlib-only CLI (flag/os.Args; the only third-party import is godbus — the grep-pinned no-framework rule), thin main + `run(ctx, args) error`; `--json` converts the same status line into typed single-line JSON client-side; ServiceUnknown/NameHasNoOwner → "daemon not running?" + exit 1; a rejected reload exits 1 with the rejection text. goswitchd builds the actor once (`newActor`) and hands it to both engine.Run and ctlsvc.Run on the signal ctx — a ctl failure is logged, never fatal.
- **Live ctl-smoke (Task 2):** temp-config daemon; (1) status + --json carry mode/config facts; (2) broken reload: exit ≠ 0, error names verify_wait_mss, status shows config_valid=false + the error (D-32); (3) forced `correct` on a typed word UNDER the still-broken on-disk document — the daemon corrects on the last-good snapshot, corrections_done=1 lands in the next status, both oracles hold привет; (4) restored reload: applied, config_valid=true, the watcher's own "config reloaded" record. PASS ×4 on the final tree.
- **Hermetic name-guard corpus:** TestSvc_NameGuard spawns a private dbus-daemon (temp config, --print-address, --nofork, killed by cleanup) — the first Run owns the name, a plain client calls Status OVER THE WIRE (the method/signature/export are unit-pinned, not just the in-process struct), the second Run answers ErrNotPrimaryOwner, ctx-cancel returns nil. No live session dependency in `mise run ci`.
- **Regressions:** `mise run ci` green at every task boundary (0 lint issues under the strict v2 config); live word-en-ru and layout-single PASS after the counter refactor touched every correction path.

## Task Commits

Each task committed atomically (TDD: RED before GREEN):

1. **Task 1: internal/ctlsvc + StatusSnapshot/CorrectNow + Reload** — `58435b3` (test, RED) + `21a9219` (feat, GREEN)
2. **Task 2: goswitchctl + проводка демона + ctl-smoke** — `db95d70` (feat)

**Plan metadata:** this commit (docs: complete plan)

## TDD Gate Compliance

The tdd task followed RED → GREEN with machine-validated red evidence:

| Task | RED commit | Evidence record | Verdict | GREEN commit |
|------|-----------|-----------------|---------|--------------|
| 1 | `58435b3` test(03-06) | red-evidence/03-06-task1-red.json | RED_EVIDENCE_OK (6 new top-level tests failing on assertions — Status/ReloadConfig/CorrectNow/RecoverShim/NameGuard in ctlsvc, StatusSnapshot in session; 79 prior green; compile stubs only) | `21a9219` feat(03-06) |

No REFACTOR commit — the GREEN implementations landed lint-clean under the strict v2 config (all findings fixed inside the GREEN iterations per D-08).

## Decisions Made

- The six key-decisions of the frontmatter, in brief: the key=value status canon is both the human report and the JSON source (no second method); ctlsvc holds session only as the Status data type beside three single-method interfaces; ErrNotPrimaryOwner is exported and the guard is semantically real; the correction counters wrap every done/skipped site with D-24 counting done; a ctl failure never kills the daemon; reload rejections and the no-config daemon answer D-Bus errors (exit 1), Status is the always-succeeding diagnostic.
- The daemon's own readiness record `"ctl service listening"` is the gate the live case waits for before its first goswitchctl call (grep-stable, like "component registered").

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Plan-literal ambiguity] "ctlsvc не импортирует session-типы напрямую" vs the plan's own key_links**
- **Found during:** Task 1 (interface design)
- **Issue:** The plan pins Deps as small interfaces AND wires ctl→session (methods call actor.StatusSnapshot()) — but StatusSnapshot's return type IS a session type, so the literal "no session import" reading is unsatisfiable without adapter structs duplicating the nine Status fields in cmd/goswitchd (which PATTERNS.md explicitly does not ask for).
- **Fix:** Deps stays three single-method interfaces at the point of use (tests drive fakes); ctlsvc imports internal/session solely for the Status struct in StatusSnapshotProvider's signature — the prescribed dependency direction (session never imports ctlsvc) is intact.
- **Files modified:** internal/ctlsvc/ctlsvc.go
- **Verification:** TestSvc_Status/TestSvc_NameGuard green over fake providers; mise run ci green
- **Committed in:** `21a9219`

**2. [Rule 1 - Naming] The sentinel is exported as ErrNotPrimaryOwner, not the plan's lowercase errNotPrimaryOwner**
- **Found during:** Task 1 RED (the corpus cannot see a private sentinel from ctlsvc_test)
- **Issue:** The plan names conn.go's private `errNotPrimaryOwner` as the template, but the pinned `errors.Is` check lives in an external test package.
- **Fix:** Exported `ErrNotPrimaryOwner` (the appid.ErrBusClosed precedent for exported sentinels); the errors.Is discipline is unchanged.
- **Files modified:** internal/ctlsvc/ctlsvc.go
- **Verification:** TestSvc_NameGuard green
- **Committed in:** `58435b3` / `21a9219`

**3. [Rule 3 - Blocking] Strict-lint findings inside the GREEN iterations (D-08)**
- **Found during:** Task 1 and Task 2 GREEN (mise run ci)
- **Issue:** err113 (dynamic errors in tests and CLI), errcheck (fmt.Fprintln), funlen/gocognit (test and case functions over the ceilings), mnd (300/450 literals), perfsprint (constant fmt.Errorf), gofmt alignment, one lll line.
- **Fix:** Package-level static errors (errStatusRejected, errCtlOwnerTimeout, errUsage, errCtlZenityNeeded, errCtlEmptyReply); printLine helper with one explicit discard; TestSvc_ReloadConfig split into three top-level tests over a shared newReloadSvc builder; TestActor_StatusSnapshot split into Snapshot/Counters/ConfigFields; the e2e case split into four step helpers; ctlWindowStart/ctlWindowRestored constants; errors.New where no formatting is needed.
- **Verification:** mise run ci green after the fixes (lint 0 issues)
- **Committed in:** `21a9219`, `db95d70`

---

**Total deviations:** 3 auto-fixed (1 plan-literal resolution, 1 naming-for-test-visibility, 1 blocking lint round)
**Impact on plan:** All resolved toward the locked must-haves; no scope creep — the FLAGGED assumptions (transport: session bus; methods without properties/signals) both held and need no owner escalation beyond the verify gate's routine confirmation.

## Issues Encountered

None beyond the deviations above. Environmental notes: the private dbus-daemon of the name-guard corpus resolved to the linuxbrew 1.16.2 binary via PATH on this machine (system dbus-daemon equivalents work the same; the corpus skips with a named diagnostic only if no dbus-daemon exists at all — it ran in every gate here).

## Known Limitations (for the verify gate)

- **CorrectNow's reply is an acknowledgment, not an outcome** — by design (the plan's async-contract wording): the pipeline's settlement (ADR-004 verification, ladder execution, refusals) lands in the status counters and the log; a scripted caller polls status (the live case shows corrections_done=1 arriving).
- **The status token grammar reserves the LAST field for config_error** — the parser treats everything after `config_error=` as that value; a future field must be inserted BEFORE the config triple, never after it (documented on renderStatus and parseStatusLine).
- **The dbus-daemon skip guard**: TestSvc_NameGuard skips (with a named diagnostic) only on machines with no dbus-daemon at all — everywhere this project runs (the target and the GNOME CI runner have one) the pin executes.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- For the 03-07 matrix v2: `goswitchctl status` is a non-UI oracle — a matrix step can gate on the log shapes OR the ctl counters (corrections_done/skip reasons/super_intercepted); the ctl-smoke case's step helpers are the template for any ctl-driven case.
- The full status vocabulary is now fixed: mode, corrections_done, corrections_skipped, skip_<reason>*, super_intercepted, super_upstream_consumed, config=none | config_path/config_valid/config_error — matrix cases and goswitchctl stay in this vocabulary.
- The per-app observer's identity (03-05) is NOT yet in the status line — the focused-app surface was left for the owner to request (the snapshot struct can carry it without a wire change to the existing tokens).

## Self-Check: PASSED

All key-files exist on disk; all three task commits found in history (58435b3, 21a9219, db95d70); the plan ledger measured 3 commits from plan_head_before c0956e6 (matches `commits:` in frontmatter); `mise run ci` green (exit 0, 0 lint issues) at the final tree; mise run e2e-ctl PASS ×4, e2e-word and e2e-layout-single PASS on the final tree.
