# Deferred Items — Phase 08

Out-of-scope discoveries logged during execution (deviation rules: scope boundary — pre-existing issues are not fixed by plan executors).

## 2026-10-05 — 08-01 (Task 2, `mise run ci`)

Pre-existing test flakes observed while establishing the green iteration on the docs-only tree. Both reproduce only intermittently; both are in code paths untouched by this plan (the plan's diff is 62 added lines in `docs/SPEC.md`, zero `.go` files). `mise run ci` passed green (exit 0) on retry; no code was changed in response.

1. **`internal/ctlsvc` — `TestRun_OnConnHookCalledOnce`: "OnConn called 0 times, want exactly 1"**
   - Environment: the owner's live desktop runs a real `goswitchd` (systemd user unit, active) that owns `org.djarvur.goswitch` on the real session bus (`DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1001/bus`); suite logs show `another goswitchd instance already registered reply=2` on flaky runs.
   - Passes standalone (`go test -race -count=1 ./internal/ctlsvc/ -run TestRun_OnConnHookCalledOnce`) and in most package runs; suspected name-release/ownership race between sequential tests on a shared session bus.
2. **`internal/config` — `TestWatch_BrokenAutocorrectBlockKeepsLastGood`: LastError non-nil after the repaired reload**
   - Debounce-timing test (200 ms debounce + `waitUntil` polling) under `-race` and parallel package execution; failed once under machine load, passed on subsequent runs.

Suggested follow-up (owner): isolate ctlsvc tests from the live session bus (private bus or `-p 1` serialization) and review the watch test's LastError-clearing window; both are candidates for a small hardening todo, not phase-8 scope.
