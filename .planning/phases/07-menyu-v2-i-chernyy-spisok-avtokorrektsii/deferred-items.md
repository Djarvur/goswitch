# Deferred Items (phase 07)

| Category | Item | Status | Deferred At | Plan |
|----------|------|--------|-------------|------|
| pre-existing-flake | `TestRun_OnConnHookCalledOnce` (internal/ctlsvc) — "OnConn called 0 times, want exactly 1" under parallel full-suite load; timing flake, passes 4/5 isolated and on package re-run; unrelated to the 07-02 diff (config/session only) | open | 2026-10-05 | 07-02 (out of scope — scope boundary) |
| post-release-tuning | задержка звука (spawn подпроцесса-плеера canberra-gtk-play ~150-300 мс) — тюнинг после релиза: ранее-хук в flipTo, pw-play/постоянный плеер, предзагрузка темы | open | 2026-10-05 | 07 UAT (owner verdict, tests 1-3) |

## 07-03 execution (2026-10-04)

- [Flake, pre-existing 07-02] `internal/config/watch_test.go` TestWatch_BrokenBlocklistPatternKeepsLastGood — occasional (~1/15 package runs) failure: after the repaired reload the test's waitUntil observes the snapshot update, but Watcher.reparse clears lastErr AFTER storing the snapshot, so the immediate LastError() read can still see the stale rejection ("must be a valid regular expression after the repaired reload, want nil"). Not caused by 07-03 (writer.go is not in the watcher path); candidate fix is a lastErr-settled wait or reordering the two stores in reparse — owner of watch.go semantics to decide. Out-of-scope per the executor scope boundary; re-run on occurrence.
