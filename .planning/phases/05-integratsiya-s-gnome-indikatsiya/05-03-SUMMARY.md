---
phase: 05-integratsiya-s-gnome-indikatsiya
plan: 03
subsystem: daemon-flip-mechanism
tags: [ibus, set-global-engine, dbus, flip, tdd, godbus]

requires:
  - phase: 05-01
    provides: "D-52 verdict d52-literal (ADR-006 Proposed), live spike facts (P2 GlobalEngineChanged delivery, P4 self-echo), engine/conn.go seam precedents"
  - phase: 05-02
    provides: "two-source install takeover — the desktop state the flip acts on"
provides:
  - "engine.Config.BindSwitcher — the per-generation SetGlobalEngine seam (rebound beside PostRegister on every generation that clears RequestName; nil = no-op)"
  - "engine.NameEN/NameRU — the single source of the flip-target engine-name literals"
  - "switch_engine journal record (Info success / Warn with cause on failure, engine literal only — the additive e2e mark for 05-05)"
  - "session.Actor.SetSwitcher + flipTo — the single execution path of ALL four flip sites (Single expiry, settleCombo, settleCorrectionFlip SET semantics, modeSwitchChord) with a hard 40 ms switchTimeout deadline"
  - "StatusSnapshot.Engine — the active engine's name on the goswitchctl status surface"
  - "cmd/goswitchd wiring: engineConfig.BindSwitcher = actor.SetSwitcher before engine.Run"
affects: [05-04-sync-listener, 05-05-e2e-two-source, goswitchctl-status-consumers]

actuals:
  tokens: 13500 # chars/4 over the plan diff (1068 insertions / 36 deletions across 10 files)
  tasks: 2
  commits: 4 # measured: git rev-list --count db7a96e..HEAD

tech-stack:
  added: [] # zero new dependencies — godbus v5.2.2 only, go.mod untouched (tidy-diff green)
  patterns:
    - "generation-scoped seam: config callback invoked beside PostRegister per serve generation — a dead generation's closure never survives the reconnect (T-05-03-03)"
    - "deadline-under-mutex flip: the SetGlobalEngine call holds a.mu bounded by switchTimeout=40 ms — chosen over the async WR-01 handoff so record order stays deterministic and rapid flips keep final bus state = final mode (why pinned in the flipTo doc comment)"
    - "fake-ibus-bus test stand: unix-socket listener + hand-rolled EXTERNAL SASL server + godbus exported DecodeMessage/EncodeTo dispatch loop — the wire act proven end-to-end hermetically"
    - "flip journal discipline: mode record (byte-stable e2e oracle) → switch_engine (engine literal only, D-20/D-21) → UpdateModeSymbol, pinned by hook-order tests"

key-files:
  created:
    - engine/conn_switcher_test.go
    - .planning/phases/05-integratsiya-s-gnome-indikatsiya/red-evidence/05-03-task1-red.json
    - .planning/phases/05-integratsiya-s-gnome-indikatsiya/red-evidence/05-03-task2-red.json
  modified:
    - engine/conn.go
    - engine/types.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - cmd/goswitchd/main.go
    - cmd/goswitchd/main_test.go
    - .golangci.yml

key-decisions:
  - "Flip execution = SetGlobalEngine on the daemon's ibus connection through the BindSwitcher seam (D-52 literal, ADR-006) — direct D-Bus call, never subprocess"
  - "Deadline-under-mutex (switchTimeout 40 ms) chosen over the async WR-01-style handoff: the plan offered either mechanism; sync execution keeps the pinned record order and makes rapid flips final-state-correct"
  - "In-process fake ibus bus for the seam corpus (hand-rolled SASL server; godbus NewConn starts no workers) — hermetic wire proof without the live desktop"
  - "Engine-seam test file is package engine (in-package) — serve/generation binding is unexported; the emitter_wire_test.go //nolint:testpackage precedent"

patterns-established:
  - "Generation-scoped rebinding beside PostRegister: any upward handoff from the ibus connection must be re-issued per serve generation (05-04's sync listener follows the same law)"
  - "flipTo(target) choke point: every new flip gesture must funnel through it — duplication of the mode/record/switch/emit sequence is forbidden"
  - "SetSwitcher-style setter mirrors AttachEngine for connection-owned seams handed to the actor"

requirements-completed: [SWCH-01, SWCH-02, SWCH-04, INTEG-01, INTEG-05]

coverage:
  - id: D1
    description: "BindSwitcher seam: per-generation rebinding, SetGlobalEngine wire act with a single string argument, switch_engine journal (Info/Warn), nil no-op"
    requirement: SWCH-01
    verification:
      - kind: unit
        ref: "engine/conn_switcher_test.go#TestSwitcherBindsPerGeneration"
        status: pass
      - kind: unit
        ref: "engine/conn_switcher_test.go#TestSwitcherCallsSetGlobalEngine"
        status: pass
      - kind: unit
        ref: "engine/conn_switcher_test.go#TestSwitcherJournalRecords"
        status: pass
      - kind: unit
        ref: "engine/conn_switcher_test.go#TestSwitcherNilNoOp"
        status: pass
    human_judgment: false
  - id: D2
    description: "NameEN/NameRU literals pinned; EngineDesc wire order untouched"
    requirement: INTEG-01
    verification:
      - kind: unit
        ref: "engine/conn_switcher_test.go#TestNameConstants"
        status: pass
      - kind: unit
        ref: "engine/wire_test.go#TestWire_FieldOrder"
        status: pass
    human_judgment: false
  - id: D3
    description: "All four flip sites route through flipTo with pinned order (mode → switch → emit; combo: done → mode → switch), failure WARN-not-fatal, one-WARN nil degradation, 40 ms deadline with bounded stall"
    requirement: SWCH-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_FlipRoutesThroughSwitcher"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_FlipOrderPinned"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_SwitcherFailureWarnsNotFatal"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_SwitcherDeadlineBounded"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_NilSwitcherInternalFlip"
        status: pass
    human_judgment: false
  - id: D4
    description: "Correction flip SET-semantics in two-engine form: the RESULT script's engine activates (lat→cyr→goswitch-ru, cyr→lat→goswitch-en)"
    requirement: SWCH-04
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_CorrectionFlipSetsResultEngine"
        status: pass
    human_judgment: false
  - id: D5
    description: "StatusSnapshot carries the active engine; goswitchd wires BindSwitcher into the actor before engine.Run"
    requirement: INTEG-05
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestStatus_ReportsActiveEngine"
        status: pass
      - kind: unit
        ref: "cmd/goswitchd/main_test.go#TestEngineConfigWiresSwitcher"
        status: pass
    human_judgment: false
  - id: D6
    description: "The flip ACT on the live desktop: SetGlobalEngine really switches the active engine mid-session and the switch_engine mark shows in the daemon journal (layout-single additive mark, two-source-flip case)"
    verification: []
    human_judgment: true
    rationale: "Live evidence consolidates in plan 05-05 by design (plan verification section: «Живые e2e-доказательства … план 05-05»); the unit corpus proves the wire shape against the fake bus, not the owner's live session"

duration: 47 min
completed: 2026-09-29
status: complete
---

# Phase 05 Plan 03: Механизм флипа через SetGlobalEngine (BindSwitcher seam) Summary

**Каждый жест переключения (одиночный Shift, Super+Space-чорд, флип после коррекции, комбо) теперь исполняется как org.freedesktop.IBus.SetGlobalEngine на ibus-соединении демона — через поколенческий шов engine.Config.BindSwitcher и единый путь actor.flipTo с жёстким дедлайном 40 мс; RU-ветка, тайминги FSM и байт-стабильный e2e-оракул записи mode нетронуты, добавлена аддитивная журнальная запись switch_engine.**

## Performance

- **Duration:** 47 min
- **Started:** 2026-09-29T21:38:43Z
- **Completed:** 2026-09-29T22:26:06Z
- **Tasks:** 2/2
- **Files modified:** 10 (7 production/test, 1 lint config, 2 red-evidence records)

## Accomplishments
- Механизм флипа ADR-006 (d52-literal) исполним: шов BindSwitcher привязывается на КАЖДОЕ поколение соединения (рядом с PostRegister), мёртвое замыкание не переживает реконнект — корпус доказывает это живым round trip'ом через подменный ibus-сокет.
- Все четыре флип-сайта актора воронкой через flipTo: порядок записей mode → switch → UpdateModeSymbol запинен (комбо: done → mode → switch — расширение D-36); сбой шины — WARN с именем движка, набор продолжается; nil-шов — ровно один WARN; дедлайн 40 мс с дисциплиной ограниченного стэлла запинены.
- Критерий 4 в двухдвижковой форме: активным становится движок СКРИПТА РЕЗУЛЬТАТА (оба направления); StatusSnapshot несёт активный движок; wiring в main запинен до engine.Run.

## Task Commits

1. **Task 1 RED: шов engine — корпус BindSwitcher + пины имён** - `992071b` (test)
2. **Task 1 GREEN: BindSwitcher seam — SetGlobalEngine, switch_engine, NameEN/NameRU** - `e624727` (feat)
3. **Task 2 RED: flipTo-through-seam корпус (8 целевых тестов)** - `c8f1baa` (test)
4. **Task 2 GREEN: flipTo исполняет все флип-жесты через шов** - `1a0a402` (feat)

**Plan metadata:** follows in the docs commit (docs: complete plan)

_Note: both tasks are TDD — each has a RED (test) and GREEN (feat) commit; REFACTOR needed no separate commit (cleanup folded into GREEN, tests still green)._

## Files Created/Modified
- `engine/conn.go` — Config.BindSwitcher seam + newSwitcher closure (SetGlobalEngine call shape verbatim from RegisterComponent; switch_engine journal at INFO/Warn)
- `engine/types.go` — NameEN/NameRU literals (wire order untouched)
- `engine/conn_switcher_test.go` — NEW: seam corpus over an in-process fake ibus bus (hand-rolled EXTERNAL SASL + godbus exported-API dispatch)
- `internal/session/actor.go` — flipTo (single execution path, switchTimeout=40ms, WARN-not-fatal, one-WARN nil degradation), SetSwitcher, engineNameOf/oppositeMode, StatusSnapshot.Engine; setScriptMode folded into flipTo's same-target guard
- `internal/session/actor_test.go` — switcher corpus: routing, order, correction-flip SET, failure, deadline, nil degradation, Status.Engine
- `cmd/goswitchd/main.go` — engineConfig.BindSwitcher = actor.SetSwitcher (wiring before engine.Run)
- `cmd/goswitchd/main_test.go` — wiring pin end to end
- `.golangci.yml` — wrapcheck relaxed for test paths (fake-bus plumbing errors surface verbatim as assertion fuel)

## Decisions Made
- **Механизм защиты от затыка шины: жёсткий дедлайн под мьютексом** (switchTimeout = 40 мс), не асинхронный WR-01-handoff — план предлагал «или/или»; синхронный вызов сохраняет детерминизм запиненного порядка записей и корректность финального состояния шины при быстрых флипах; обоснование закреплено комментарием-почему в flipTo. Живая латентность (A6) измеряется в 05-05.
- **Корпус шва — на подменном ibus-сокете**, не на живом столе: godbus не имеет серверной стороны — SASL-сервер написан вручную, диспетчер вызовов — на экспортированных DecodeMessage/EncodeTo (NewConn не стартует воркеров — это делает Auth). Доказательство wire-акта герметично и работает в headless CI.
- **Тест-файл шва в пакете engine** (не engine_test, как в букве плана): serve/поколенческая привязка неэкспортированы; прецедент emitter_wire_test.go (`//nolint:testpackage`) санкционирует именно это.

## Deviations from Plan

### Auto-fixed Issues (Deviation Rules)

**1. [Rule 3 - Blocker] BindSwitcher signature used a Go keyword as parameter name**
- **Found during:** Task 1
- **Issue:** the plan's literal signature `BindSwitcher func(switch func(ctx, engineName string) error)` is invalid Go — `switch` is a reserved keyword
- **Fix:** parameter renamed to `flip` (`BindSwitcher func(flip func(ctx context.Context, engineName string) error)`); semantics identical
- **Files modified:** engine/conn.go
- **Verification:** build + full corpus green
- **Committed in:** 992071b (stub), e624727 (GREEN)

**2. [Rule 3 - Blocker] No hermetic path to test serve-level rebinding**
- **Found during:** Task 1
- **Issue:** no in-repo fake bus existed; godbus ships no server side (no SASL server, no worker start in NewConn) — the per-generation rebinding behavior (T-05-03-03) was untestable as planned
- **Fix:** ~150 test-only LOC: fakeBus stand (unix listener + hand-rolled EXTERNAL SASL + exported-API dispatch loop answering Hello/RequestName/RegisterComponent/SetGlobalEngine)
- **Files modified:** engine/conn_switcher_test.go
- **Verification:** the seam corpus proves the wire act end-to-end; whole suite green under -race
- **Committed in:** 992071b, e624727

**3. [Plan-letter resolution] TestActor_SwitcherDeadlineBounded's mutex probe realized as the bounded-stall probe**
- **Found during:** Task 2
- **Issue:** the behavior text lists both probes (deadline ≤40 ms AND "call does not hold a.mu"), but the action mandates choosing ONE mechanism ("выбрать один механизм") and the must_haves say «дедлайн И/ИЛИ вынос из-под a.mu»; the chosen deadline mechanism holds a.mu boundedly by design
- **Fix:** the test pins the deadline at entry (≤40 ms) plus the real invariant — a Status snapshot served THROUGH a seam stuck until the deadline completes within budget, so the keystroke path can never hang; the why-comment in flipTo documents the choice
- **Files modified:** internal/session/actor_test.go, internal/session/actor.go
- **Verification:** corpus green; if 05-05's live A6 latency exceeds the budget, the WR-01 off-mutex ladder remains the documented escalation
- **Committed in:** c8f1baa, 1a0a402

---

**Total deviations:** 3 auto-resolved (2 × Rule 3 blockers, 1 plan-letter resolution within the plan's own either/or)
**Impact on plan:** none on the contract surface — the seam signature, wire act, journal shape, order pins and timing semantics match the must_haves; the deviations are test-infrastructure and parameter-naming level.

## Issues Encountered
- golangci-lint strict findings on the new test file (lll, noctx, errcheck, nolintlint, wrapcheck) — all fixed in-session; the wrapcheck relaxation for test paths was added to `.golangci.yml` with justification, mirroring the existing test-file relaxations (no deferred lint).
- A first RED run hung: the deadline test blocked on an unguarded `<-entered` against the stub (the seam is never called at RED) — guarded with a budget timeout; the 8-test RED then completed in ~5 s.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- **05-04 (sync listener)** consumes GlobalEngineChanged + FocusIn names through the SAME generation-scoped seam law; P4 self-echo suppression (the initiator receives its OWN signal — live-proven in 05-01) is its obligation; BindSwitcher's rebinding point shows exactly where the listener subscription belongs.
- **05-05 (live e2e)** owns the live proof of D6 (the flip ACT on the owner's desktop: layout-single's additive switch_engine mark, two-source-flip case) and the A6 latency measurement against the 40 ms budget; the honest-budget line lands in ADR-006 if exceeded.
- goswitchctl status consumers may surface the new `Status.Engine` field (rendering untouched this plan — the struct field is the contract).
- WINDOWS ledger: no entries — no stubs, no skipped tests, all `<verify>` blocks ran green (`mise run ci` at each task).

---
*Phase: 05-integratsiya-s-gnome-indikatsiya*
*Completed: 2026-09-29*
