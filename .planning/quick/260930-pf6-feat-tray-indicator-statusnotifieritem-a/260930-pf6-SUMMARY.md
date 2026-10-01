---
status: complete
phase: 260930-pf6-feat-tray-indicator-statusnotifieritem-a
plan: 260930-pf6-PLAN.md
date: 2026-09-30
commits:
  - 2b0cc08  # test(260930-pf6): pin the EN/RU tray pixmap renderer with golden tests (RED)
  - 053e2c8  # feat(260930-pf6): deterministic EN/RU ARGB32 tray pixmap renderer
  - 168aca7  # test(260930-pf6): pin the SNI item corpus against watcher/emitter fakes (RED)
  - 209ae06  # feat(260930-pf6): SNI item — watcher probe, register, export the live item, ModeChanged with one-WARN degradations
  - f9270f7  # test(260930-pf6): pin the ModeDisplay seam — install push, observer last in flip and sync (RED)
  - 78e79a5  # feat(260930-pf6): actor ModeDisplay seam — install push, observer last in flipTo and syncMode
  - b7311cf  # test(260930-pf6): pin the ctlsvc OnConn hook (RED)
  - a23dc82  # feat(260930-pf6): tray indicator wired — ctlsvc OnConn hook, daemon attaches the item
  - 9063703  # fix(260930-pf6): serve Menu as the /NO_DBUSMENU sentinel (verification gap)
verification: human_needed # 5/5 programmatic truths verified live; owner pixel confirmation pending
---

# SUMMARY: tray indicator (StatusNotifierItem/AppIndicator)

## What was built

goswitchd publishes an EN/RU tray icon on the session bus (SNI/AppIndicator)
— the owner's decision after GNOME Shell's own input-source indicator proved
dead on this desktop (exhaustively, 2026-09-30). The icon always matches the
daemon's script mode and flips on every chord, correction flip, and external
sync.

- `internal/indicator/pixmap.go` — deterministic in-code ARGB32 renderer
  (embedded 1-bit E/N/R/U font, fixed 24×24 canvas, white-on-transparent),
  pinned by golden files (`-update` regen).
- `internal/indicator/indicator.go` — the org.kde.StatusNotifierItem at
  /StatusNotifierItem on the daemon's EXISTING ctl connection and bus name
  (org.djarvur.goswitch): watcher probe → double export (SNI + Properties) →
  RegisterStatusNotifierItem; ModeChanged swaps the pixmap under the item's
  own mutex and emits NewIcon. Every degradation (no watcher, probe/register/
  export error, emit failure, unknown symbol) is exactly one WARN and an
  inert display — the daemon never blocks or dies. `Menu` is served as the
  `/NO_DBUSMENU` sentinel: the ubuntu-appindicators watcher DESTROYS an item
  refusing Menu (NEEDED_PROPERTIES = ['Id','Menu'], verification gap, fixed
  9063703).
- `internal/session/actor.go` — the ModeDisplay seam (SetModeDisplay, the
  SetSwitcher mirror): install pushes the CURRENT mode (startup + late
  install self-sync); the observer fires LAST in flipTo and syncMode — the
  D-36 record order with the display appended. Nil display = zero behavior
  change (the whole pre-existing corpus green untouched).
- `internal/ctlsvc/ctlsvc.go` — the optional OnConn hook in Deps, called
  once after the bus name and the ctl object are live; its failure is a WARN
  and serving continues (ctl-failure-never-kills-the-daemon).
- `cmd/goswitchd/main.go` — wiring: OnConn attaches the indicator and
  installs it as the actor's display, before engine.Run (live from the
  first generation).

Zero new module deps (godbus in-module only, tidy-diff clean), zero config
keys, zero installed icon files, no SPEC/ADR edits.

## Verification

- Per-task strict red→green; every iteration green on `mise run ci`
  (build + vet + golangci-lint strict + go test -race -count=1 + tidy-diff).
- gsd-plan-checker: VERIFICATION PASSED (all dimensions).
- gsd-verifier: 4/5 live-verified, 1 gap found (Menu sentinel — the watcher
  evicted the item ~3 s after every attach) → fixed (9063703) → re-verified
  live: registration sticks (t+1s…t+25s), IconPixmap flips en→ru→en on the
  real bus (md5 cce7bd2d24ee ↔ 57014693527e), no indicator WARNs. Status:
  human_needed — the owner's pixel confirmation in the top bar is the last
  gate (260930-pf6-VERIFICATION.md).

## Post-deploy verification (executed by the orchestrator)

```bash
systemctl --user restart goswitchd
busctl --user get-property org.kde.StatusNotifierWatcher /StatusNotifierWatcher \
  org.kde.StatusNotifierWatcher RegisteredStatusNotifierItems   # contains org.djarvur.goswitch
busctl --user call org.djarvur.goswitch /StatusNotifierItem org.freedesktop.DBus.Properties \
  Get ss org.kde.StatusNotifierItem IconPixmap                  # (iiay), flips with the mode
```

## Deviations

- The executor subagent died mid-Task-2 (runtime quota limit) after the RED
  commit; the orchestrator completed Tasks 2–4 inline against the same plan,
  same TDD discipline, same gate bar.
- Two implementation-time fixes beyond the corpus the executor left behind:
  export the LIVE item (not `new(Item)`) so the watcher reads the real
  pixmap; return the zero dbus.Variant on Properties error paths
  (MakeVariant(nil) panics in signature computation).
- `mise run test` flaked once under load (a timing-sensitive test during
  concurrent bus probes); two consecutive full runs green, direct run green.
