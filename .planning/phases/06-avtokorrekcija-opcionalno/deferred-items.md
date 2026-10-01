# Deferred Items — Phase 06

Out-of-scope discoveries logged during execution (per the executor scope-boundary rule: pre-existing issues in unrelated files are not fixed by the plan that stumbles on them).

## 06-05 (internal/detect)

- **Flaky test (pre-existing, unrelated package):** `internal/ctlsvc` `TestRun_OnConnHookCalledOnce` failed once during a full `mise run ci` ("OnConn called 0 times, want exactly 1") in the middle of the 06-05 run; a local unix-socket bus race is visible in the same log ("another goswitchd instance already registered reply=2"). It passes 4/4 on isolated re-runs and the full `mise run ci` is green on the identical tree. Not touched by 06-05 (a new pure package with no imports of ctlsvc/engine). Candidate for a stability look (socket/timeout hygiene) in a future plan.
