---
phase: 261001-fg3-feat-interactive-tray-daemon-re-attaches
verified: 2026-10-01T09:42:19Z
status: passed
score: 6/6 must-haves verified
covered_files:
  - .planning/quick/261001-fg3-feat-interactive-tray-daemon-re-attaches/261001-fg3-PLAN.md
  - .planning/quick/261001-fg3-feat-interactive-tray-daemon-re-attaches/261001-fg3-SUMMARY.md
  - internal/session/actor.go
  - internal/session/actor_test.go
  - internal/indicator/indicator.go
  - internal/indicator/indicator_test.go
  - internal/indicator/supervisor.go
  - internal/indicator/supervisor_test.go
  - internal/indicator/menu.go
  - internal/indicator/menu_test.go
  - internal/ctlsvc/ctlsvc.go
  - internal/ctlsvc/ctlsvc_test.go
  - cmd/goswitchd/main.go
  - README.md
covered_digest: "v1:sha256:cb8febebd3dbae99e2b9720ccc756657aee90505b3a035e3d06a33000c9b4926"
behavior_unverified: 0
overrides_applied: 0
gaps: []
human_verification:
  - test: "Boot-race / watcher-appears-late (event path): reboot, or restart goswitchd right after ibus restart. Watch journalctl --user -u goswitchd."
    expected: "The icon appears WITHOUT a manual daemon restart. Journal: the attach-time one WARN («tray indicator attach failed») — then NOTE: the recovery itself may be SILENT on the event path (no «tray indicator re-registered» INFO; see warning W-2). The icon's return is the recovery verdict, not the INFO line."
    why_human: "Real GNOME SNI watcher, real bus broadcast, real boot ordering — no headless test can exercise the live bridge."
  - test: "Silent eviction (poll path): simulate the shell dropping the item (ibus restart, or restarting gnome-shell's appindicator extension)."
    expected: "Within one ~30s tick the icon returns; the journal shows exactly ONE WARN «tray indicator lost» (reason field naming the verdict) followed by exactly one INFO «tray indicator re-registered». Repeated failed ticks do NOT repeat the WARN."
    why_human: "Requires the live watcher's registry and the real 30s ticker pacing."
  - test: "Menu interaction: click the tray icon on GNOME."
    expected: "The menu opens (GNOME renders the item as a menu button). First item «Переключить раскладку» flips icon EN<->RU AND the typing layout (journal shows the byte-stable mode to=ru/en record). «Статус» posts a desktop notification with mode + version. «Перечитать конфиг» is greyed without -config; with -config it re-reads the YAML."
    why_human: "GNOME's ubuntu-appindicators bridge rendering and the notification daemon are external surfaces."
  - test: "Menu-less fallback: call SNI Activate on the item (e.g. busctl --user call org.djarvur.goswitch /StatusNotifierItem org.kde.StatusNotifierItem Activate iix 0 0), or any applet that calls Activate."
    expected: "The mode toggles exactly once — the same flipTo path as the menu item."
    why_human: "Requires a live D-Bus peer driving the exported item object."
  - test: "Regression sanity after the tray work."
    expected: "goswitchctl status unchanged; correction gestures still work with the menu open/closed; no journal spam at idle (steady-state ~30s ticks are silent)."
    why_human: "Live-session end-state check."
---

# Quick Task 261001-fg3: Interactive tray — self-healing supervisor + DBusMenu Verification Report

**Task Goal:** feat: interactive tray — daemon re-attaches when the SNI watcher appears (NameOwnerChanged) and re-registers on silent eviction via a periodic health check (~30s); DBusMenu on the item (toggle as first item, status and reload items); SNI Activate wired to the actor's mode toggle for menu-less environments.
**Verified:** 2026-10-01T09:42:19Z
**Status:** human_needed (all codebase must-haves verified; live-desktop leg is the orchestrator's post-merge step per the execution constraints)
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Tray icon self-heals: re-registers immediately on NameOwnerChanged, within one ~30s tick on the poll path, no daemon restart | ✓ VERIFIED | `internal/indicator/supervisor.go` — `check()` (NameHasOwner probe → RegisteredStatusNotifierItems → eviction verdict → tryReattach), `onWatcherSignal` (defensive 3-part body + non-empty-new-owner guard, pinned by TestSupervisorOwnerChangedGuards), `Supervise` live wiring (AddMatchSignal + buffered channel + real ticker, ctx-driven teardown). Behavioral pins: TestSupervisorEvictionHealsOnTick (full re-registration sequence + one NewIcon), TestSupervisorOwnerChangedReattaches, TestSupervisorRunConsumesBeats (real run loop consumes a synthetic beat and heals), TestSupervisorRunExitsOnContextDone. The `itemsContainService` deviation (service or service+"/" prefix) is CORRECT: the SNI watcher's registry lists items as `<bus name><object path>` — a literal match would re-register every tick; both wire shapes pinned (TestSupervisorRecognizesBareServiceItem + the suffixed `watcherItem` constant in the steady-state corpus) |
| 2 | Log discipline: one WARN per loss, one INFO per revival, silent steady state, failed re-attach waits for the NEXT beat | ✓ VERIFIED | TestSupervisorConsecutiveFailuresOneWarn (5 failed beats → exactly 1 WARN, 0 INFO), TestSupervisorFailedReattachRetriesNextTick (failed+healed beats → exactly 1 WARN + 1 INFO total; 3 register calls), TestSupervisorSteadyStateSilent (zero log output), TestSupervisorEvictionHealsOnTick (1 WARN + 1 INFO; post-heal beat adds nothing). See warning W-2 for the boot-race event-path caveat |
| 3 | Menu FIRST item «Переключить раскладку» → the SAME flipTo path (D-36 order); «Статус» notification; «Перечитать конфиг» greyed without -config | ✓ VERIFIED | `internal/indicator/menu.go` — GetLayout root kids pinned in order (toggle, status, reload) by TestMenuLayout; the `(ia{sv}av)` wire shape pinned via `dbus.SignatureOf(menuLayout{}) != dbus.ParseSignatureMust(...)`. Event dispatch (id 1/2/3, "clicked" only) by TestMenuEventDispatch. `internal/session/actor.go` ToggleMode → `a.flipTo(oppositeMode(a.mode))` under a.mu; TestActor_ToggleModeFlips pins switch targets goswitch-ru/goswitch-en, byte-stable `"msg":"mode","to"` records, and ops strictly symbol:ru → display:ru → symbol:en → display:en (D-36). `cmd/goswitchd/main.go` wires Toggle: actor.ToggleMode, Status: notifyStatus (org.freedesktop.Notifications.Notify, mode+version only), Reload set only when a watcher exists |
| 4 | SNI Activate triggers the same toggle (menu-less environments) | ✓ VERIFIED | `internal/indicator/indicator.go` `Activate(_, _ int32)` under sniIface — recover-shimmed, drives the same `cb.Toggle`; TestItemActivateInvokesToggle (exactly once), TestItemActivateNilToggleWarnsOnce (one WARN per item lifetime) |
| 5 | Every degradation one-WARN and non-fatal: sentinel fallback with registration intact; recover shim; daemon keeps typing | ✓ VERIFIED | TestAttachMenuExportFailureServesSentinel (per-call export failure predicate → exactly 1 WARN, Menu property = /NO_DBUSMENU sentinel, register call intact); `recoverMenuCall` contains panics below Event/Activate with an ERROR + stack (TestMenuEventPanicContained); nil-callback Event is a silent no-op (TestMenuEventDispatch); the ctl attach failure path WARN-and-continues (TestRun_OnConnHookFailureNeverAbortsServing) |
| 6 | Zero new module deps, zero new config keys, no SPEC/ADR edits; scope guard (12 files) | ✓ VERIFIED | `git diff bd6eb0f..HEAD -- go.mod go.sum` is EMPTY (tidy-diff green); healthInterval is a compile-time 30s constant (no config key); diff stat = exactly the 12 files in the plan's files_modified list (+1588/−76); no SPEC/ADR files touched |

**Score:** 6/6 truths verified (0 present, behavior-unverified)

### Warnings (non-blocking)

| # | Finding | Severity | Detail |
|---|---------|----------|--------|
| W-1 | Flaky test `TestRun_OnConnHookCalledOnce` under `-race` (~5–10% of runs) | ⚠️ Warning | Reproduced twice in ~30 package runs: `ctlsvc_test.go:647: OnConn called 0 times, want exactly 1`. Root cause: `waitCtlOwner` polls GetNameOwner on a separate bus connection and can observe the name owned in the window between `Run`'s RequestName completing and its goroutine reaching the OnConn call. The racing assertion pattern is byte-identical in the pre-task commit bd6eb0f — PRE-EXISTING, not a regression of this task. One full `mise run ci` run hit it; a clean rerun completed green (all packages ok, golangci-lint 0 issues, exit 0). Suggested follow-up (out of this task's scope): have waitCtlOwner additionally await the hook signal before the call-count read |
| W-2 | Boot-race event-path recovery logs NO «tray indicator re-registered» INFO | ⚠️ Warning | Traced deterministically: a failed initial attach (the 2026-10-01 boot race) leaves `sup.lost == false` (attach() warns via its own path and never calls markLost), so when NameOwnerChanged later fires and tryReattach succeeds, `wasLost` is false and no INFO is logged — the recovery is silent. This matches the plan's Task-2 letter («if lost was set, exactly one INFO») but contradicts the plan's own post-deploy step 1 and SUMMARY step 1, which expect the INFO after the boot-race recovery. The eviction/poll path (lost set) satisfies the INFO discipline fully. Human check item 1 above is phrased so the silent-but-successful recovery is not misjudged. Optional follow-up: seed `sup.lost = true` on a failed attach (or log the INFO on the first successful registration after a failed attach); no corpus test currently covers the failed-attach→recover sequence |

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/indicator/supervisor.go` | Re-attachable lifecycle: NameOwnerChanged + ~30s health check, channel seams, ctx exit | ✓ VERIFIED | 220 lines; substantive; born in attach(), Supervise called from main.go:133 — wired |
| `internal/indicator/menu.go` | org.canonical.dbusmenu v3 at /Menu: GetLayout/GetGroupProperties/AboutToShow/Event, static 3-item layout | ✓ VERIFIED | 206 lines; exported in registerItem() before the item registration — wired |
| `internal/indicator/indicator.go` | Watcher seam + RegisteredStatusNotifierItems, registered/emitDead split + revive, Activate, Menu property | ✓ VERIFIED | All present; registerItem runs probe → menu exports → item exports → register; Menu property decided before registration |
| `internal/session/actor.go` | Public ToggleMode into the existing flipTo | ✓ VERIFIED | +14 lines, purely additive ( ToggleMode + doc comment); called from main.go callbacks |
| `internal/ctlsvc/ctlsvc.go` | Deps.OnConn gains the Run context as first parameter | ✓ VERIFIED | `OnConn func(ctx context.Context, conn *dbus.Conn) error`; Run passes its serve ctx at ctlsvc.go:219 |
| `cmd/goswitchd/main.go` | Menu callbacks wired; supervisor started on the Run context | ✓ VERIFIED | Callbacks built from actor/reload in scope; `go item.Supervise(connCtx, conn)` with the ctx received through OnConn |
| `README.md` | 2–3 line touch: menu on click, first item toggles, self-healing icon | ✓ VERIFIED | "Input sources and the mode indicator" section extended with exactly the prescribed note (+7 lines) |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| supervisor check() | Watcher seam (NameHasOwner + RegisteredStatusNotifierItems) | per-tick probe + registry read | ✓ WIRED | connWatcher implements both; sup.w seeded in attach() |
| NameOwnerChanged match | tryReattach → registerItem → item revive | AddMatchSignal(iface=org.freedesktop.DBus, member=NameOwnerChanged, arg0=watcherName) on the same ctl conn | ✓ WIRED | Supervise matches the watcher name; onWatcherSignal guards name/body/new-owner then re-runs the full registerItem |
| menu Event(id=1) / item Activate | Callbacks.Toggle → actor.ToggleMode → flipTo | main.go cb wiring | ✓ WIRED | D-36 order pinned end-to-end by TestActor_ToggleModeFlips |
| menu Event(id=2) | Callbacks.Status → org.freedesktop.Notifications.Notify | notifyStatus on the SAME conn, pure D-Bus | ✓ WIRED | No subprocess; body = mode + version only (T-03-06-03) |
| menu Event(id=3) | Callbacks.Reload → config Watcher.Reload | reloadConfig closure; nil without -config | ✓ WIRED | cb.Reload set only when reload != nil; menuProps serves item 3 `enabled = cb.Reload != nil` (greyed) |
| SNI Menu property | menu export outcome | /Menu when served; /NO_DBUSMENU sentinel fallback with one WARN | ✓ WIRED | menuPath decided during registerItem before the watcher can read it; TestAttachMenuExportFailureServesSentinel pins the fallback |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| Item properties | it.pix | PixmapFor(symbol) composition, swapped by ModeChanged | Yes | ✓ FLOWING |
| Item Menu property | it.menuPath | registerItem's export outcome | Yes | ✓ FLOWING |
| Status notification body | st.Mode, st.Version | actor.StatusSnapshot() | Yes | ✓ FLOWING |
| Menu layout/labels | static constants | plan-specified static menu | Static BY DESIGN | ✓ N/A (not a stub — the plan mandates a static layout with compile-time ids/labels) |
| Supervisor signals/ticks | sup.signals / sup.ticks | live conn + time.Ticker in Supervise; synthetic channels in the corpus by design | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Touched-package corpus green | `go test -race -count=1 ./internal/indicator/ ./internal/session/ ./internal/ctlsvc/ ./cmd/...` | all ok | ✓ PASS |
| Full CI gate | `mise run ci` | first run: FAIL on flaky W-1; clean rerun: build+vet+lint(0 issues)+test all ok, exit 0 | ✓ PASS (after W-1 rerun) |
| Dependency freeze | `mise run tidy-diff` | clean; go.mod/go.sum untouched since bd6eb0f | ✓ PASS |
| Commit trail | `git log bd6eb0f..HEAD --oneline` | 8/8 claimed commits present (6851280, 6bd360a, e7335a7, fdf6539, 2c1832c, d96dcb2, d3399c0, fd19bf7) | ✓ PASS |
| Key named tests | TestActor_ToggleModeFlips, TestSupervisorEvictionHealsOnTick, TestSupervisorOwnerChangedReattaches, TestMenuLayout, TestAttachMenuExportFailureServesSentinel, TestRun_OnConnHookCalledOnce | all green within the package runs above | ✓ PASS |

### Probe Execution

SKIPPED — no `scripts/*/tests/probe-*.sh` probes declared for this task; the mise ci/tidy-diff gates above are the task's runnable checks.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| QUICK-261001-fg3-interactive-tray | 261001-fg3-PLAN.md | Interactive tray: self-healing supervisor + DBusMenu + Activate | ✓ SATISFIED | Truths 1–6 all verified; D1–D4 coverage entries in the SUMMARY map to passing named tests (D5 is the live-desktop leg → human verification) |

No orphaned requirements: REQUIREMENTS.md carries no phase mapping for this quick task (plan-local requirement ID).

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| internal/indicator/supervisor_test.go | QUICK-261001-fg3 | 11 | 0 | 0 | Behavioral (state transitions, exact log counts, call sequences) | SOUND |
| internal/indicator/menu_test.go | QUICK-261001-fg3 | 8 | 0 | 0 | Behavioral + wire-shape (signature pins) | SOUND |
| internal/indicator/indicator_test.go | QUICK-261001-fg3 | (updated) | 0 | 0 | Behavioral (export sequences, sentinel, emit discipline) | SOUND |
| internal/session/actor_test.go | QUICK-261001-fg3 | (updated) | 0 | 0 | Behavioral (D-36 op order, byte-stable records) | SOUND |
| internal/ctlsvc/ctlsvc_test.go | QUICK-261001-fg3 | (updated) | 0 (startTestBus t.Skip is an environment skip when dbus-daemon is absent) | 0 | Behavioral (call-once, ctx hand-off, failure degradation) | SOUND — W-1 flake noted |

**Disabled tests on requirements:** 0. **Circular patterns detected:** 0. **Insufficient assertions:** 0.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | None. No TBD/FIXME/XXX/HACK/placeholder markers in any touched source file; no empty implementations; no hardcoded-empty data flowing to output | — | — |

### Decision Coverage

SKIPPED cleanly — no CONTEXT.md exists for this quick task (no `<decisions>` block to check).

### Human Verification Required

The live-desktop leg is the ORCHESTRATOR's post-merge step (after binary reinstall), per the execution constraints. Five items are listed in the frontmatter `human_verification` block; they mirror the SUMMARY's "Post-deploy verification steps" with one correction: item 1's journal expectation is amended per warning W-2 (the boot-race event-path recovery may be silent — the icon's return is the verdict, not the INFO line).

### Gaps Summary

None. All six must-have truths verified against the codebase with behavioral test evidence; all artifacts exist, are substantive, and are wired; all six key links connected; zero new dependencies; scope guard held exactly (12 files). Two warnings recorded (W-1 pre-existing flaky OnConn test; W-2 boot-race INFO silence — a plan-internal inconsistency whose implementation side is defensible but whose post-deploy journal narrative needed correcting in the human-check items above). Status is human_needed solely for the orchestrator-owned live-desktop leg.

---

_Verified: 2026-10-01T09:42:19Z_
_Verifier: Claude (gsd-verifier)_

## Orchestrator live leg (2026-10-01, post-merge) + one gap found and fixed

Deployed binaries from fd19bf7. The supervisor's live behavior matched the
design IMMEDIATELY: the first registration after restart did not stick (the
known re-registration race), and the health check healed it within one beat —
journal shows exactly one WARN `tray indicator lost` (item missing from the
watcher's registry) + one INFO `tray indicator re-registered`, then the
registration stuck (verified across a full 35 s window).

**GAP (found during the live leg, fixed in e8dae38):** every menu dispatch
answered UnknownInterface — the executor wrote `menuIface =
"org.canonical.dbusmenu"`; the real DBusMenu protocol is
`com.canonical.dbusmenu`. The corpus could not catch it (fakes compare the
same constant — a wrong interface NAME is invisible to a self-consistent
corpus), and the item served the correct /Menu path, so the watcher
destroy/re-register loop ran silently every 30 s. Fix: the correct constant
PLUS a literal wire-name pin (TestMenuWireNamesPinned) so a protocol-name
drift can never pass the corpus again.

Post-fix live evidence (daemon from e8dae38):
- `com.canonical.dbusmenu.GetLayout(0,1,[])` returns the 3-item layout —
  «Переключить раскладку» (id 1, enabled), «Статус» (id 2), «Перечитать
  конфиг» (id 3, disabled without -config);
- `com.canonical.dbusmenu.Event(1, "clicked", ...)` flips the mode en→ru and
  the served IconPixmap follows (md5 swap) — the menu toggle is LIVE;
- SNI Activate flips too; selfcheck 6×ok; supervisor healed the restart race
  once more and the registration sticks.

Remaining human check: the owner sees the icon, clicks it — the menu opens,
«Переключить раскладку» flips the layout/icon, «Статус» posts a notification
(«Перечитать конфиг» is greyed without -config, by design).

## Owner confirmation (2026-10-01)

The owner confirmed visually: the tray icon with its menu works — the menu
opens, «Переключить раскладку» flips the layout/icon. All human gates of the
interactive-tray task are closed.
