---
phase: 261001-fg3-feat-interactive-tray-daemon-re-attaches
plan: 01
subsystem: indicator
tags: [dbus, sni, statusnotifieritem, dbusmenu, ibus, gnome, tray, systemd, godbus]

# Dependency graph
requires:
  - phase: 260930-pf6 (quick, tray indicator v0)
    provides: the SNI item on the ctl connection, the ModeDisplay seam, the pixmap renderer, the /NO_DBUSMENU sentinel
  - phase: 05 (ADR-006 two-source model)
    provides: actor.flipTo — the single execution path of every flip; the switcher seam; the D-36 record order
provides:
  - indicator supervisor — NameOwnerChanged re-attach + ~30s health check with one-WARN/one-INFO discipline
  - org.canonical.dbusmenu v3 object at /Menu (toggle first, status, reload) with sentinel fallback
  - item SNI Activate for menu-less environments
  - session.Actor.ToggleMode — the public gesture into flipTo
  - ctlsvc Deps.OnConn carrying the serve context; daemon wiring of callbacks + supervisor
affects: [e2e matrix (menu/gesture cases), verifier (live desktop human-check), future spec-delta on the indicator surface]

# Actuals (#2632)
actuals:
  tokens: 20050    # chars/4 over the realized diff (80199 diff chars, 12 files, +1588/-76)
  tasks: 4
  commits: 8

# Tech tracking
tech-stack:
  added: []        # zero new module deps (godbus v5.2.2 only; tidy-diff clean at every gate)
  patterns:
    - supervisor pattern: single-writer select loop over ctx/ticks/signals; unexported methods driven by synthetic channels in the corpus, exported Supervise for live wiring
    - re-registration seam: registerItem re-runs the FULL probe -> exports -> register sequence on every re-attach (godbus re-export is idempotent replacement)
    - recover-shimmed interactive surface: menu Event and item Activate open with recoverMenuCall (the ctlsvc recoverMethod precedent)

key-files:
  created:
    - internal/indicator/supervisor.go
    - internal/indicator/supervisor_test.go
    - internal/indicator/menu.go
    - internal/indicator/menu_test.go
  modified:
    - internal/indicator/indicator.go
    - internal/indicator/indicator_test.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - internal/ctlsvc/ctlsvc.go
    - internal/ctlsvc/ctlsvc_test.go
    - cmd/goswitchd/main.go
    - README.md

key-decisions:
  - "Registry membership accepts service or service-prefixed path (itemsContainService): the ubuntu-appindicators watcher lists items as <bus name><object path>; a literal match would re-register the item every ~30s tick on the live machine (Rule 2 correctness fix over the plan's literal wording)"
  - "Item state split: disabled -> registered (supervisor-managed, revivable) + emitDead (permanent per connection) — ModeChanged always swaps the pixmap but emits only while registered"
  - "dbusmenu Event with nil callbacks = silent no-op; item Activate with nil Toggle = one-WARN no-op per item lifetime (the emitWarned discipline pattern)"
  - "Menu is stateless (immutable cb) — no mutex of its own; the recover shim bounds whatever the callbacks let escape (T-FG3-04 satisfied vacuously)"
  - "OnConn ctx hand-off pinned in the corpus via a channel (fatcontext forbids storing a context into a captured closure variable)"
  - "healthInterval = 30s compile-time constant; zero new config keys, zero new deps, no SPEC/ADR edits (scope guard held: 12 files, all in the plan's list)"

patterns-established:
  - "Channel-driven corpus for select loops: unexported supervisor methods (check/onWatcherSignal) tested directly for deterministic pins; run() tested with synthetic channels for exit + beat consumption"
  - "Wire-shape pins via dbus.SignatureOf(x) != dbus.ParseSignatureMust(\"(ia{sv}av)\") — the 02-03 AttrList av/au lesson applied to menuLayout"

requirements-completed:
  - QUICK-261001-fg3-interactive-tray

coverage:
  - id: D1
    description: "ToggleMode — public gesture into the flipTo pipeline (mode record, switcher target engine, panel symbol, display last — D-36 order)"
    requirement: QUICK-261001-fg3-interactive-tray
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_ToggleModeFlips"
        status: pass
    human_judgment: false
  - id: D2
    description: "Supervisor self-heal: eviction detected on the ~30s tick or the NameOwnerChanged event, full re-registration, one-WARN/one-INFO discipline, silent steady state, retry on next beat, ctx exit"
    requirement: QUICK-261001-fg3-interactive-tray
    verification:
      - kind: unit
        ref: "internal/indicator/supervisor_test.go (TestSupervisorEvictionHealsOnTick, TestSupervisorOwnerChangedReattaches, TestSupervisorOwnerChangedGuards, TestSupervisorFailedReattachRetriesNextTick, TestSupervisorConsecutiveFailuresOneWarn, TestSupervisorRunExitsOnContextDone, TestSupervisorRunConsumesBeats)"
        status: pass
    human_judgment: false
  - id: D3
    description: "DBusMenu at /Menu: v3 properties, static layout (toggle FIRST, reload greyed without callbacks), clicked dispatch with recover containment, AboutToShow false, /NO_DBUSMENU sentinel fallback with registration intact"
    requirement: QUICK-261001-fg3-interactive-tray
    verification:
      - kind: unit
        ref: "internal/indicator/menu_test.go (TestMenuProperties, TestMenuLayout, TestMenuGetGroupProperties, TestMenuAboutToShow, TestMenuEventDispatch, TestMenuEventPanicContained, TestItemActivateInvokesToggle, TestItemActivateNilToggleWarnsOnce) + indicator_test.go#TestAttachMenuExportFailureServesSentinel"
        status: pass
    human_judgment: false
  - id: D4
    description: "Daemon wiring: OnConn carries the serve ctx; menu callbacks wired (ToggleMode, notification, reload); supervisor started on the Run context; README note"
    requirement: QUICK-261001-fg3-interactive-tray
    verification:
      - kind: unit
        ref: "internal/ctlsvc/ctlsvc_test.go (TestRun_OnConnHookCalledOnce, TestRun_OnConnHookFailureNeverAbortsServing) + mise run ci (build/vet/lint/test -race across ./...)"
        status: pass
    human_judgment: true
    rationale: "The live end-state (icon re-attachment on the real desktop, menu interaction through GNOME's appindicator bridge, notification rendering) is the orchestrator's post-merge human-check — the corpus is hermetic by design (no live bus in mise ci)"
  - id: D5
    description: "Live desktop verification: reboot/shell-restart re-attachment (event path), manual eviction self-heal within one ~30s tick (poll path), menu clicking end-to-end"
    requirement: QUICK-261001-fg3-interactive-tray
    verification: []
    human_judgment: true
    rationale: "Live desktop work is orchestrator-owned per the execution constraints — exact steps recorded under 'Post-deploy verification steps (live desktop)' below"

# Metrics
duration: 51min
completed: 2026-10-01
status: complete
---

# Quick Task 261001-fg3: Interactive tray — self-healing supervisor + DBusMenu Summary

**The tray icon is no longer a one-shot bet on boot ordering: the daemon re-attaches when the shell's SNI watcher appears (NameOwnerChanged) or silently evicts the item (~30s health check), and the item gained a DBusMenu («Переключить раскладку» first) plus SNI Activate — zero new deps, zero config keys, every degradation a contained one-WARN.**

## Performance

- **Duration:** 51 min
- **Tasks:** 4/4 (each strict red -> green: RED commit, then GREEN commit)
- **Commits:** 8 (measured: `git rev-list --count bd6eb0f..HEAD`), all on `gsd/phase-05-integratsiya-s-gnome-indikatsiya`
- **Gates:** `mise run ci` (build + vet + golangci-lint strict + `go test -race -count=1 ./...`) and `mise run tidy-diff` green before every GREEN commit; RED commits fail to compile by design (new method/struct pins before the implementation exists)

## Commits

| Hash | Type | Content |
|------|------|---------|
| 6851280 | test | RED — TestActor_ToggleModeFlips (full flipTo pipeline pin) |
| 6bd360a | feat | ToggleMode + display-order op-label constants (goconst) |
| e7335a7 | test | RED — supervisor corpus + item seam extensions (registered/emitDead split, revive, RegisteredStatusNotifierItems) |
| fdf6539 | feat | supervisor.go — NameOwnerChanged re-attach + ~30s health check, Supervise wiring |
| 2c1832c | test | RED — DBusMenu corpus + item menu/Activate seams (Callbacks surface, registerItem menu sequence) |
| d96dcb2 | feat | menu.go — dbusmenu v3 object at /Menu + SNI Activate + sentinel fallback; main.go call site |
| d3399c0 | test | RED — ctlsvc OnConn corpus carries the serve ctx (non-nil pin) |
| fd19bf7 | feat | OnConn ctx signature, daemon menu callbacks + supervisor start, README note |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Correctness] Registry membership accepts the service-prefixed path**
- **Found during:** Task 2 (implementation)
- **Issue:** The plan's literal check "our service missing from the list" assumes `RegisteredStatusNotifierItems` entries equal the service name. The ubuntu-appindicators watcher lists items as `<bus name><object path>` (e.g. `org.djarvur.goswitch/StatusNotifierItem`); a literal equality match would be permanently "missing" on the live machine and re-register the item on EVERY ~30s tick.
- **Fix:** `itemsContainService` accepts `service` or `service+"/"` prefix; both wire shapes pinned by the corpus (`TestSupervisorRecognizesBareServiceItem`, suffixed form in the steady-state test).
- **Files modified:** internal/indicator/supervisor.go, supervisor_test.go
- **Commit:** fdf6539

**2. [Rule 3 - Blocking] Strict-lint mechanical fixes (green-iteration directive 2)**
- **Found during:** every task
- **Issue/fix pairs:** funcorder demanded exported methods before unexported ones (ToggleMode moved into the exported block; Item/Menu helpers reordered); goconst demanded the display-order op labels become constants; err113 demanded a static sentinel for the variant-type mismatch (`errItemsType`); mnd demanded the NameOwnerChanged body arity become `ownerChangedBodyParts`; revive var-naming demanded `Id`->`ID` in the wire structs; fatcontext forbade storing a context into a captured closure variable (the OnConn ctx pin switched to a channel hand-off); funlen/lll/cyclop/dogsled line-length and complexity refactors in the test corpus.
- **Files modified:** the touched test/source files above
- **Commits:** 6bd360a, fdf6539, d96dcb2, fd19bf7 (spread across the GREEN commits)

### Deliberate scope notes (not deviations)

- `watcherIface` is a new constant with the same spelling as `watcherName` (the plan's "new constant" instruction — the item's sniIface is the different string).
- The supervisor treats a CLOSED signals channel as "stop reacting, keep health beats" (nil-channel select guard) — defensive, unpinned by the plan, one line.
- The eviction corpus's export pin grew to 8 entries (menu + item exports for attach + re-attach) — forced by the full-sequence re-registration the plan itself specifies.

## Verification Results

- `go test -race -count=1 ./internal/session/` — green (ToggleMode pinned; entire existing corpus untouched and green).
- `go test -race -count=1 ./internal/indicator/` — green (supervisor corpus: eviction self-heal on tick, event re-attach + guards, one-WARN/one-INFO discipline, retry-on-next-beat, revival icon refresh, ctx exit, beat consumption; menu corpus: v3 properties, pinned layout, dispatch containment, Activate, sentinel fallback; v0 degradation pins hold reworded).
- `go test -race -count=1 ./internal/ctlsvc/ ./cmd/...` — green (OnConn call-once + non-nil ctx + failure WARN-and-continue on the new signature).
- `mise run ci` + `mise run tidy-diff` — green at every task gate; go.mod/go.sum untouched (godbus v5.2.2 remains the only module dep).

## Post-deploy verification steps (live desktop — ORCHESTRATOR work, not executed here)

After merging and reinstalling the binary (`go build ./cmd/goswitchd` -> `systemctl --user restart goswitchd`, or the install flow):

1. **Boot-race / watcher-appears-late (event path):** reboot, or `systemctl --user restart goswitchd` right after `ibus restart`. The icon appears WITHOUT a manual daemon restart. Journal (`journalctl --user -u goswitchd -f`): the attach-time one-WARN (if the watcher was absent at start) then, on the watcher's appearance, exactly one INFO `tray indicator re-registered`.
2. **Silent eviction (poll path):** simulate the shell dropping the item (e.g. `ibus restart`, or killing/restarting gnome-shell's appindicator extension). Within one ~30s tick the icon returns; the journal shows exactly ONE WARN `tray indicator lost` (reason: `item missing from the watcher's registry`) followed by exactly one INFO `tray indicator re-registered`. Repeated failed ticks must NOT repeat the WARN (one-WARN discipline).
3. **Menu interaction:** click the icon — GNOME renders the menu. First item «Переключить раскладку» flips icon EN<->RU AND the typing layout (the same flipTo path; journal shows the byte-stable `mode to=ru/en` record). «Статус» posts a desktop notification with mode + version. «Перечитать конфиг» is greyed without `-config`; with `-config` it re-reads the YAML (broken edit -> WARN `config reload rejected`, last-good keeps serving).
4. **Menu-less fallback:** SNI Activate (e.g. `busctl --user call org.djarvur.goswitch /StatusNotifierItem org.kde.StatusNotifierItem Activate iix 0 0`-shaped call or any applet that calls Activate) toggles the mode exactly once.
5. **Regression sanity:** `goswitchctl status` unchanged; correction gestures still work with the menu open/closed; no journal spam at idle (steady-state ticks are silent).

## Self-Check: PASSED

- Files exist: internal/indicator/supervisor.go, supervisor_test.go, menu.go, menu_test.go — FOUND (committed in fdf6539 / d96dcb2).
- All 8 commit hashes found in `git log bd6eb0f..HEAD` — FOUND.
- Working tree clean of task modifications (docs artifacts left uncommitted per the orchestrator's constraint).
