---
phase: 05-integratsiya-s-gnome-indikatsiya
plan: 04
subsystem: engine-ibus-sync
tags: [ibus, dbus, signals, sync, reactivation, tdd, adr-006]

requires:
  - phase: 05-01
    provides: spike verdicts P2 (AddMatch works on the ibus socket — GlobalEngineChanged delivered to the daemon's connection) and P4 (the SetGlobalEngine initiator receives its OWN signal — echo suppression mandatory); ADR-006 d52-literal
  - phase: 05-03
    provides: the generation-scoped BindSwitcher seam and its serve law (rebind beside PostRegister on EVERY successful RequestName), engine.NameEN/NameRU literals, the switch_engine journal record, StatusSnapshot.Engine, the fake-ibus-bus test stand
provides:
  - engine.Config.OnGlobalEngine — the sync-listener input fed by the GlobalEngineChanged dispatcher AND FocusIn
  - engine.Config.BindGlobalEngine — the generation-scoped GetGlobalEngine reader seam, bound before PostRegister
  - stormHandler — the daemon connection's dropping dbus.SignalHandler (bounded storm memory, one WARN, Terminate = errBusClosed)
  - Actor.SyncEngine — closed-enum sync {goswitch-en→EN, goswitch-ru→RU, foreign→WARN}, own-flip echo confirms silently, never calls the switcher
  - activate.IfOwned(ctx, run, globalEngine) — factual-engine priority over the dead gsettings current key, cold-bus fallback, foreign skip
  - cmd/goswitchd wiring — OnGlobalEngine→actor.SyncEngine, BindGlobalEngine→reader→IfOwned
affects: [05-05 (live proofs: external-flip-sync, ibus-restart ru survival), README third-source documentation]

actuals:
  tokens: 16263
  tasks: 2
  commits: 4

tech-stack:
  added: []
  patterns:
    - "generation-scoped sync seams on the ibus connection: BindGlobalEngine follows the BindSwitcher law (rebind per generation in serve, reader bound BEFORE PostRegister so same-generation IfOwned consults its own reader)"
    - "dropping signal handler: strict non-blocking deliver replaces godbus's default deferred-blocking-goroutine fan-out; Terminate closing the registered channels IS the bus-loss verdict"
    - "closed-enum sync input: only engine.NameEN/NameRU can move the daemon's mode; echo = same-mode silent confirmation; correction records but never flips the bus"

key-files:
  created:
    - engine/conn_sync_test.go
    - .planning/phases/05-integratsiya-s-gnome-indikatsiya/red-evidence/05-04-task1-red.json
    - .planning/phases/05-integratsiya-s-gnome-indikatsiya/red-evidence/05-04-task2-red.json
  modified:
    - engine/conn.go
    - engine/engine.go
    - engine/factory.go
    - engine/conn_switcher_test.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - internal/activate/activate.go
    - internal/activate/activate_test.go
    - cmd/goswitchd/main.go
    - cmd/goswitchd/main_test.go

key-decisions:
  - "stormHandler replaces godbus's default signal handler: v5.2.2's default defers a BLOCKING goroutine per overflow signal (deferredDeliver) — unbounded memory under a storm; the custom handler drops on a full channel with one bounded WARN per generation and its Terminate closes the registered channels, preserving the errBusClosed bus-loss verdict (T-05-04-02)"
  - "Echo suppression per spike P4 = SyncEngine's same-mode branch: the daemon's own flip comes back as GlobalEngineChanged with the mode already matching — a confirmation without a record; a drift corrects with the byte-stable mode record then a 'mode corrected' WARN, panel emit after the records (flipTo order minus the switcher leg)"
  - "SyncEngine NEVER calls the switcher — the correction moves only the internal mode, so the flip loop is impossible by construction (single-writer, T-05-04-01); pinned by TestActor_SyncEngineNeverSwitches"
  - "activate.IfOwned: a successful factual read outranks everything — goswitch name reactivates directly (Pitfall 3: GNOME 46 never writes current), foreign name skips with DEBUG WITHOUT consulting the dead key, reader failure falls back to the legacy current-index derivation unchanged"
  - "BindGlobalEngine binds BEFORE PostRegister in serve (a reorder of the 05-03 seam order): the same generation's reactivation must consult its own generation's reader, never a dead closure"
  - "Sync-listener corpus lives in the in-package engine/conn_sync_test.go: engine_test.go is the external-package file and cannot reach unexported serve/dispatchSignals/onGlobalEngine (the conn_switcher_test.go sanctioned seam)"

patterns-established:
  - "SyncEngine seam: two observers (signal dispatcher + FocusIn) converge into ONE closed-enum sync input on the actor; sync follows, never fights"
  - "hermetic wiring pins: a foreign factual name makes IfOwned skip before any subprocess, so the main.go reader→IfOwned handoff is provable without touching the desktop"

requirements-completed: [INTEG-04]

coverage:
  - id: D1
    description: "GlobalEngineChanged subscription + path/member-filtered dispatch into Config.OnGlobalEngine on the daemon's ibus connection; serve-contour verdicts (ctx-done → nil, closed channel → errBusClosed) survive the dispatcher refactor"
    requirement: INTEG-04
    verification:
      - kind: unit
        ref: "tests/engine/conn_sync_test.go#TestGlobalEngineChangedDispatched"
        status: pass
      - kind: unit
        ref: "tests/engine/conn_sync_test.go#TestWaitBusLossVerdictsPreserved"
        status: pass
    human_judgment: false
  - id: D2
    description: "Signal-storm bounded: the connection's dropping handler keeps memory at signalBufferSize, warns exactly once per generation, the dispatcher stays responsive and serve exits nil on cancel (T-05-04-02)"
    requirement: INTEG-04
    verification:
      - kind: unit
        ref: "tests/engine/conn_sync_test.go#TestSignalStormBounded"
        status: pass
    human_judgment: false
  - id: D3
    description: "FocusIn forwards THIS engine's wire name into the same sync seam, recoverHandler-wrapped; EventHandler.HandleLifecycle signature untouched (OQ3)"
    requirement: INTEG-04
    verification:
      - kind: unit
        ref: "tests/engine/conn_sync_test.go#TestFocusInForwardsEngineName"
        status: pass
    human_judgment: false
  - id: D4
    description: "Per-generation GetGlobalEngine reader bound before PostRegister; parses the engine name from the desc variant (positional field 2 of the types.go wire order); nil seam = no-op"
    requirement: INTEG-04
    verification:
      - kind: unit
        ref: "tests/engine/conn_sync_test.go#TestBindGlobalEnginePerGeneration"
        status: pass
    human_judgment: false
  - id: D5
    description: "Actor.SyncEngine: closed enum, own-flip echo confirms silently (P4), drift corrects with the byte-stable mode record + WARN + panel emit, foreign names warn without touching state, zero switcher calls from any sync input (T-05-04-01/03)"
    requirement: INTEG-04
    verification:
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_SyncEngineFollowsEngine"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_SyncEngineForeignWarns"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_SyncEngineNeverSwitches"
        status: pass
    human_judgment: false
  - id: D6
    description: "activate.IfOwned prefers the factual engine (ru survives a restart with current=0 — the Pitfall-3 regression pin), cold bus falls back to the index path, foreign factual names skip without consulting the dead key; void contract and test knobs untouched"
    requirement: INTEG-04
    verification:
      - kind: unit
        ref: "tests/internal/activate/activate_test.go#TestIfOwned_RuActiveSurvivesRestart"
        status: pass
      - kind: unit
        ref: "tests/internal/activate/activate_test.go#TestIfOwned_PrefersGlobalEngine"
        status: pass
      - kind: unit
        ref: "tests/internal/activate/activate_test.go#TestIfOwned_ColdBusFallsBackToIndex"
        status: pass
      - kind: unit
        ref: "tests/internal/activate/activate_test.go#TestIfOwned_ForeignGlobalEngineSkips"
        status: pass
    human_judgment: false
  - id: D7
    description: "Daemon wiring: OnGlobalEngine → actor.SyncEngine (mode corrects), BindGlobalEngine → reader, PostRegister hands the reader to IfOwned (hermetic foreign-skip handoff pin)"
    requirement: INTEG-04
    verification:
      - kind: unit
        ref: "tests/cmd/goswitchd/main_test.go#TestMain_WiringSyncAndReader"
        status: pass
    human_judgment: false
  - id: D8
    description: "Live behavior: an indicator click (external flip) is followed by the daemon within a gesture; a daemon/ibus restart with goswitch-ru active keeps ru (INTEG-04 live criterion)"
    verification: []
    human_judgment: true
    rationale: "Assigned by this plan's own verification section to plan 05-05's live-proofs matrix (external-flip-sync, ibus-restart cases on the GNOME desktop) — the unit level is proven here; the desktop truth is 05-05's deliverable."

duration: 46 min
completed: 2026-09-29
status: complete
---

# Phase 05 Plan 04: Sync Listener and Factual-Engine Reactivation Summary

**The daemon now FOLLOWS the factual active engine — GlobalEngineChanged and FocusIn converge into a closed-enum SyncEngine that never fights the bus, and the restart reactivation prefers the live GetGlobalEngine truth over GNOME 46's dead `current` key (Pitfall 3).**

## Performance

- **Duration:** 46 min
- **Started:** 2026-09-29T22:32:05Z
- **Completed:** 2026-09-29T23:18:12Z
- **Tasks:** 2 (both TDD: RED → GREEN, RED_EVIDENCE_OK ×2)
- **Files modified:** 13 (10 production/test sources, 1 new test corpus, 2 red-evidence records)

## Accomplishments
- Sync listener live-viable per spike P2: AddMatchSignal(path=/org/freedesktop/IBus, iface=org.freedesktop.IBus) + a buffered channel registered BEFORE the post-registration seams (early signals buffer, never vanish), dispatched through a path/member filter into `Config.OnGlobalEngine`; the waitBusLoss verdicts (ctx→nil, close→errBusClosed) survived the refactor byte-identically
- Echo suppression (spike P4 obligation): the daemon's own flip returns as a same-mode GlobalEngineChanged and confirms silently; a real drift corrects the internal mode with the byte-stable `"msg":"mode"` record + a `mode corrected` WARN — and SyncEngine structurally CANNOT call the switcher, so the flip loop is impossible by construction (T-05-04-01)
- Pitfall 3 closed at unit level: `IfOwned` reactivates the FACTUAL engine from the generation-scoped GetGlobalEngine reader; the RED run literally reproduced the defect (`ibus engine goswitch-en` where goswitch-ru was due) and GREEN pins `TestIfOwned_RuActiveSurvivesRestart` green
- Storm safety (T-05-04-02): the daemon's connection drops signals on a full channel (godbus's default would park a goroutine per overflow signal), warns exactly once per generation, and its Terminate closes the channels — which IS the bus-loss verdict

## Task Commits

1. **Task 1 RED: sync-listener corpus** - `ee961ad` (test)
2. **Task 1 GREEN: dispatcher + FocusIn feed + GetGlobalEngine reader + stormHandler** - `9b3874f` (feat)
3. **Task 2 RED: SyncEngine/IfOwned/wiring corpus** - `869baf0` (test)
4. **Task 2 GREEN: SyncEngine + IfOwned priority + main wiring** - `e218ab3` (feat)

**Plan metadata:** committed with the SUMMARY (docs)

_Note: no REFACTOR commits — the lint-driven cleanups (funcorder placement, funlen/cyclop extractions) folded into the GREEN commits per the green-iteration gate._

## Files Created/Modified
- `engine/conn.go` — stormHandler (dropping dbus.SignalHandler), subscribeGlobalEngine (AddMatch + buffered channel + reader seam before PostRegister), dispatchSignals (GlobalEngineChanged → OnGlobalEngine with path/member filter), newGlobalEngineReader + engineDescName (positional field 2), dialIbus/logRegistered extractions
- `engine/engine.go` — FocusIn forwards `e.name` into the sync seam behind recoverHandler; HandleLifecycle untouched
- `engine/factory.go` — factory carries onGlobalEngine into every minted engine
- `engine/conn_sync_test.go` — the in-package sync corpus (5 behavior tests) + stand emit/flood helpers
- `engine/conn_switcher_test.go` — fakeBus stand extensions: write mutex, client tracking, AddMatch/RemoveMatch + GetGlobalEngine answers
- `internal/session/actor.go` — SyncEngine (closed enum, echo-safe, switcher-free) + syncMode
- `internal/session/actor_test.go` — TestActor_SyncEngineFollowsEngine / ForeignWarns / NeverSwitches
- `internal/activate/activate.go` — IfOwned(ctx, run, globalEngine): factual-engine priority, cold-bus fallback, foreign skip
- `internal/activate/activate_test.go` — TestIfOwned_PrefersGlobalEngine / RuActiveSurvivesRestart / ColdBusFallsBackToIndex / ForeignGlobalEngineSkips
- `cmd/goswitchd/main.go` — engineConfig: OnGlobalEngine→actor.SyncEngine, BindGlobalEngine→(string,bool) reader, PostRegister→IfOwned(reader)
- `cmd/goswitchd/main_test.go` — TestMain_WiringSyncAndReader (hermetic handoff pin)

## Decisions Made
- **stormHandler over godbus's default signal handler** — the plan's memory-bound mitigation (T-05-04-02) is unachievable with the default: v5.2.2 defers a blocking goroutine per overflow signal (default_handler.go deliver → deferredDeliver). The custom handler does a strict non-blocking deliver, drops with ONE bounded WARN per generation, and Terminate closes the registered channels — preserving the errBusClosed verdict that the serve loop reconnects on.
- **BindGlobalEngine binds BEFORE PostRegister** — a deliberate reorder of the seam order from 05-03: the same generation's reactivation must consult its own generation's reader (a reader captured later would miss the first PostRegister or use a stale closure).
- **Foreign factual name skips without consulting `current`** — the factual GetGlobalEngine answer outranks the dead gsettings key in every branch where it is readable; the key is consulted only on a cold bus.
- **Sync corpus in an in-package test file** — the dispatcher, the sync seam and the serve cycle are unexported; the external-package engine_test.go cannot reach them (the conn_switcher_test.go in-package precedent).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Custom dropping signal handler (stormHandler)**
- **Found during:** Task 1 (GREEN design)
- **Issue:** godbus v5.2.2's default signal handler does NOT drop overflow: `signalChannelData.deliver` spawns a blocking `deferredDeliver` goroutine per signal that does not fit a full channel — a signal storm parks unbounded goroutines and queues unbounded `*Signal` values, contradicting the plan's own mitigation "память ограничена signalBufferSize" (T-05-04-02, severity medium, disposition mitigate).
- **Fix:** a custom `dbus.SignalHandler` installed via `dbus.Dial(addr, dbus.WithSignalHandler(...))`: strict non-blocking fan-out, drop counted with one WARN per generation, `Terminate()` closing the registered channels (the interface godbus's `Terminator` contract already calls on Close — the bus-loss verdict rides on it, mirroring the default handler).
- **Files modified:** engine/conn.go, engine/conn_sync_test.go (TestSignalStormBounded proves drops: 64-signal storm over a parked consumer dispatches ≤ signalBufferSize+1 and warns exactly once)
- **Verification:** mise run ci green; TestSignalStormBounded green under -race
- **Committed in:** 9b3874f

**2. [Rule 3 - Blocking] Sync corpus placed in engine/conn_sync_test.go (new in-package file) + fakeBus stand extensions**
- **Found during:** Task 1 (RED authoring)
- **Issue:** the plan's files list named engine/engine_test.go for the corpus, but that file is `package engine_test` (external) and cannot reach the unexported serve/dispatchSignals/onGlobalEngine the behavior tests must drive; the fakeBus stand also had no way to emit signals or answer AddMatch/GetGlobalEngine.
- **Fix:** new in-package engine/conn_sync_test.go (the sanctioned conn_switcher_test.go seam style); fakeBus gained a write mutex, client-connection tracking, `emitSignal`, `setGlobalEngine`, and AddMatch/RemoveMatch/GetGlobalEngine answers; engine_test.go unchanged (its corpus still passes — the EventHandler-signature-unchanged proof).
- **Files modified:** engine/conn_sync_test.go (new), engine/conn_switcher_test.go
- **Verification:** full engine corpus green under -race
- **Committed in:** ee961ad, 9b3874f

---

**Total deviations:** 2 auto-fixed (1 missing-critical, 1 blocking test-placement)
**Impact on plan:** none on the contract — both deviations serve the plan's own must-haves (storm-bound truth, behavior-test reachability); zero new dependencies; go.mod/go.sum untouched.

## Issues Encountered
- Three test-design bugs caught by the corpus's own execution and fixed in-session: TestBindGlobalEnginePerGeneration's shared `reached` channel let generation 0's leftover signal satisfy generation 1's await (dedicated `readerReached` channel); TestSignalStormBounded counted dispatches via a channel drained before the parked callback wrote (mutex-guarded counter + a `stormEpisode` helper); TestMain_WiringSyncAndReader invoked a reader variable the engineConfig closure never populates (the handoff is proven hermetically through PostRegister + a foreign-name skip instead).
- `mise run ci` lint iterated the strict gates (funcorder, funlen, cyclop, gochecknoglobals, lll) into GREEN — no lint debt left behind.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- 05-05 (live proofs) receives the complete sync mechanism: subscription + dispatcher + SyncEngine + factual-priority reactivation, all unit-pinned; the live-proofs plan owns external-flip-sync (indicator click followed), the ibus-restart ru-survival case (INTEG-04 live criterion, D8 above) and the unresolved spike P3 XKB-truth question.
- The ADR-006 obligations for 05-04 (sync listener + self-echo suppression, P2/P4) are discharged; ADR-006 moves to Accepted on 05-05's live evidence per its status clause.
- README third-source documentation (the honest exit from under goswitch) remains a 05-05 Task 2 item per the phase plan.

---
*Phase: 05-integratsiya-s-gnome-indikatsiya*
*Completed: 2026-09-29*

## Self-Check: PASSED

- All 10 created/modified source and planning files exist on disk (checked by path).
- All 5 plan commits exist in history: `ee961ad` (test 05-04 RED T1), `9b3874f` (feat 05-04 T1), `869baf0` (test 05-04 RED T2), `e218ab3` (feat 05-04 T2), `fe9846e` (docs 05-04 plan complete).
- Measured `git rev-list --count` from the plan ledger base `22d5f40` to HEAD: 5 commits (4 production/RED + 1 docs metadata) — matches the frontmatter `commits: 4` production count plus the metadata commit.
- `mise run ci` green on the final tree (build + vet + golangci-lint strict + test -race, all packages).
