---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
plan: 05
subsystem: session+daemon
tags: [a11y, wiring, actor-fold, seam, diff-gate, self-sync, reconciler, wr01, d32, tdd, tracer]

# Dependency graph
requires:
  - phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill (план 08-02)
    provides: "(a A11y) Active() — единственный источник ACTIVE-семантики, который фолд читает"
  - phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill (план 08-03)
    provides: "internal/a11y Reconciler с Apply(active bool), fire-and-forget серией под собственным mu и запиненными конструкторами NewExecRunner()/NewDBusStatusSetter()"
  - phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
    provides: "SoundSink/SetSoundSink канон seam'а (doc-форма, self-sync, nil-деградация) и refreshACBlocklist форма дифф-гейта"
provides:
  - "Seam A11ySink на точке использования в акторе (type A11ySink interface { Apply(active bool) }, doc-канон SoundSink: fire-and-forget, никогда не блокирует hot path, никогда не паникует; nil — тихий no-op)"
  - "(*Actor).SetA11ySink — форма SetSoundSink с self-sync: установка после фолда активного состояния немедленно пушит Apply(true) (OnConn-порядок: стартовый фолд уже отработал)"
  - "Фолд a11y в applySnapshot: pushA11y читает ТОЛЬКО snap.A11y.Active() и пушит ровно на изменение булева (форма refreshACBlocklist) — повторный фолд того же состояния не зовёт sink ни при каком темпе фолдов"
  - "Wiring демона: OnConn собирает reconciler на production-адаптерах a11y.New(a11y.NewExecRunner(), a11y.NewDBusStatusSetter()) и ставит sink одним вызовом рядом с SetSoundSink; применение едет существующим фолдом (старт — FoldAppliedConfig в newActor + self-sync, reload — applySnapshot), ноль новых watcher-веток"
  - "Корпус a11y-фолда: 6 тестов (активация/дифф-гейт/деактивация/пустой-список/self-sync/nil-no-op)"
affects: [08-06 (доки описывают a11y-секцию и её применение), 08-VALIDATION (живая UAT на столе)]

# Actuals (#2632) — pairs with the plan's estimate to calibrate future estimates.
actuals:
  tokens: 3670    # chars/4 over the realized diff (14682 chars: actor.go + actor_test.go + main.go + red-evidence)
  tasks: 2
  commits: 3      # MEASURED: git rev-list --count gsd-plan-head-before-08-05..HEAD
plan_head_before: 70df4146d46969d2e97de98a8087e9c322e44e5d

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Seam-канон актора продолжен: A11ySink определён на точке использования, Set-сеттер с self-sync последнего сфолденного состояния, nil = тихая деградация — третий экземпляр формы (MenuSync, SoundSink, A11ySink)"
    - "Фолд-пуш с дифф-гейтом на булеве состояния (pushA11y рядом с pushMenuSync): чтение семантики ТОЛЬКО из метода конфига (Active()), состояние-кэш в поле актора, вызов sink под actor.mu легален т.к. контракт fire-and-forget (actor.mu → sink.mu односторонний, WR-01)"

key-files:
  created:
    - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/red-evidence/08-05-task1-red.json
  modified:
    - internal/session/actor.go
    - internal/session/actor_test.go
    - cmd/goswitchd/main.go

key-decisions:
  - "Фолд вынесен в метод pushA11y (форма pushMenuSync), а не инлайн в applySnapshot: строгий funlen (40 утверждений) + канон «per-fold push chain» — семантика та же, план называл форму «рядом с autocorrect-фолдом»; комментарий WR-01/одностороннего порядка перенесён в doc pushA11y"
  - "Стартовое применение при отсутствующем watcher'е (no-config старт) не выполняется вовсе: FoldAppliedConfig живёт под if watcher != nil в newActor, а Defaults() неактивны — Apply(0) не нужен; при активном документе порядок «фолд → OnConn → SetA11ySink(self-sync)» доставляет начальное Apply без новых строк (проверено чтением main.go:488-496)"
  - "TestActor_A11yNilSinkNoOp — green-by-design в RED (отрицательное утверждение тривиально истинно без фолда), честно записано в RED-свидетельство по прецеденту 08-03 (TestA11y_BeltNotCalledOnDeactivation)"

patterns-established:
  - "pushA11y — канон фолда состояния-булева во внешний reconciler: Active()-источник, дифф-гейт на поле, self-sync сеттера"

requirements-completed: ["D-8-1", "D-8-3", "D-8-6"]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Фолд актора доставляет Active()-состояние в sink ровно на изменение: активация ровно один Apply(true); повторный фолд и фолд с другим списком при том же Active — ноль повторных; деактивация доставляет Apply(false); enabled+пустой список — Apply(false) (Active()-семантика, не enabled); nil-sink — тихий no-op"
    requirement: "D-8-3"
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_A11yFoldTriggersSinkOnActivation"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_A11yFoldDiffGated"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_A11yDeactivatePropagates"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_A11yEnabledEmptyListInactive"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_A11yNilSinkNoOp"
        status: pass
    human_judgment: false
  - id: D2
    description: "Self-sync установки: фолд активного документа без sink, затем SetA11ySink → немедленный Apply(true) на fake (форма SetSoundSink) — стартовое применение не зависит от порядка установки"
    requirement: "D-8-3"
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_A11ySinkInstallSelfSyncs"
        status: pass
    human_judgment: false
  - id: D3
    description: "Существующий акторный корпус (звук/меню/autocorrect) зелёный без правок; ACTIVE-семантика только из Active() (греп-гейты FOLD-OK и ACTIVE-SINGLE-SOURCE); -race зелёный"
    requirement: "D-8-3"
    verification:
      - kind: other
        ref: "mise exec -- go test ./internal/session/ -race -count=1 → exit 0 (282 теста, /tmp/08-05-t1-green2.txt)"
        status: pass
      - kind: other
        ref: "греп-гейт плана: type A11ySink interface + SetA11ySink + snap.A11y.Active() → FOLD-OK; !grep a11yActive = .*len\\(|A11y.Enabled && → ACTIVE-SINGLE-SOURCE"
        status: pass
    human_judgment: false
  - id: D4
    description: "Wiring демона: OnConn-блок содержит a11y.New(a11y.NewExecRunner(), a11y.NewDBusStatusSetter()) + actor.SetA11ySink рядом с SetSoundSink; в дифе main.go ноль строк fsnotify/NewWatcher (применение только через существующий фолд); build+vet зелёные"
    requirement: "D-8-3"
    verification:
      - kind: other
        ref: "греп-гейт плана: SetA11ySink + a11y.New в cmd/goswitchd/main.go → WIRING-OK; git diff HEAD -- main.go без ^+.*(fsnotify|NewWatcher) → NO-NEW-WATCHER"
        status: pass
      - kind: other
        ref: "mise exec -- go build ./... && mise exec -- go vet ./... → BUILD-VET-OK"
        status: pass
    human_judgment: false
  - id: D5
    description: "Зелёная итерация на выходе плана: mise run ci (build + vet + golangci-lint strict + test -race ./... + tidy-diff) exit 0; кодовый контур фазы (session+a11y+config+install) под -race зелёный"
    requirement: "D-8-6"
    verification:
      - kind: other
        ref: "mise run ci → CI_EXIT=0 (прогон /tmp/08-05-ci-run1.log, первая попытка)"
        status: pass
      - kind: other
        ref: "mise exec -- go test ./internal/session/ ./internal/a11y/ ./internal/config/ ./internal/install/ -race -count=1 → exit 0 (/tmp/08-05-plan-verify.txt)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Живой эффект на столе: демон при старте и на hot reload реально ставит toolkit-accessibility и пояс IsEnabled для приложений из списка"
    verification: []
    human_judgment: true
    rationale: "Юнит-корпус доказывает проводку через seam-фейки; живой эффект требует реальной GNOME-сессии — предмет 08-VALIDATION Manual-Only (FLAGGED assumption плана: headless-деградация в WARN-эпизоды 08-03, демон стартует, магия отложена)"

# Metrics
duration: 9min
completed: 2026-10-05
status: complete
---

# Phase 8 Plan 05: Wiring a11y-магии Summary

**A11ySink-seam с дифф-гейтом в акторе (pushA11y читает только snap.A11y.Active(), пушит ровно на изменение булева, SetA11ySink self-syncs) и wiring демона: reconciler на production-адаптерах a11y.New(NewExecRunner, NewDBusStatusSetter) одним вызовом рядом с SetSoundSink — старт и reload едут существующим фолдом, ноль новых веток.**

## Performance

- **Duration:** 9 min
- **Started:** 2026-10-05T18:09:16Z
- **Completed:** 2026-10-05T18:18:21Z
- **Tasks:** 2 (3 коммита: RED → GREEN трейсера + wiring)
- **Files modified:** 3 кода (+252 строки) + 1 red-evidence JSON

## Accomplishments

- Seam `A11ySink` на точке использования в акторе (D-8-3): `type A11ySink interface { Apply(active bool) }` рядом с SoundSink, doc-канон SoundSink — fire-and-forget: реализация никогда не блокирует hot path актора (серия подпроцессов идёт на собственной сериализации reconciler'а 08-03, вне actor.mu — WR-01), никогда не паникует, каждый отказ — её собственный WARN-эпизод; nil-sink — тихий no-op
- Фолд `pushA11y` в applySnapshot-цепочке (рядом с pushMenuSync): читает ТОЛЬКО `snap.A11y.Active()` — единственный источник ACTIVE-семантики (08-02), ноль вычислений enabled/len в акторе (греп-гейт ACTIVE-SINGLE-SOURCE) — и пушит ровно на изменение булева `a.a11yActive` (форма refreshACBlocklist): повторный фолд того же состояния и фолд с другим списком, но тем же Active — ноль вызовов sink; вместе с diff-gate и read-verify-then-set reconciler'а — двойная идемпотентность (T-08-05-02)
- `SetA11ySink` — форма SetSoundSink с self-sync: установка после фолда активного состояния немедленно пушит `Apply(true)` — стартовое применение (D-8-3 reconcile-at-start) не зависит от порядка «фолд в newActor → OnConn → установка»
- Wiring демона: OnConn-блок собирает reconciler на production-адаптерах `a11y.New(a11y.NewExecRunner(), a11y.NewDBusStatusSetter())` (запиненные имена 08-03 сверены с кодом — совпали дословно) и ставит sink одним вызовом рядом с SetSoundSink; отдельной reload-ветки и watcher'а нет (D-32-контур, греп-гейт NO-NEW-WATCHER на дифе main.go)
- Полный TDD-цикл трейсера: RED `7e8ef39` (5 целевых тестов пали на push-count утверждениях, RED_EVIDENCE_OK; nil-no-op green-by-design записан честно) → GREEN `d4d019c`; wiring `6014f7e`; `mise run ci` exit 0 с первой попытки

## Task Commits

Each task was committed atomically:

1. **Task 1 (трейсер): фолд актора** — RED `7e8ef39` (test), GREEN `d4d019c` (feat; включает lint-фиксы lll/goconst/funlen по прецеденту 08-02/08-03); REFACTOR не потребовался — pushA11y легла в каноническую форму сразу
2. **Task 2: wiring демона** — `6014f7e` (feat; glue — tdd-исключение, верификация build/vet/ci + грепы)

**Plan metadata:** следует отдельным docs-коммитом (SUMMARY + STATE + ROADMAP + REQUIREMENTS).

_Note: TDD-гейты: test(08-05) `7e8ef39` → feat(08-05) `d4d019c`._

## Files Created/Modified

- `internal/session/actor.go` — A11ySink-интерфейс, поля a11ySink/a11yActive, SetA11ySink с self-sync, pushA11y в applySnapshot (doc WR-01/односторонний порядок блокировок)
- `internal/session/actor_test.go` — корпус 6 тестов: fakeA11ySink, a11yListedApp-константа, a11yDoc-хелпер; активация/дифф-гейт/деактивация/пустой-список/self-sync/nil-no-op
- `cmd/goswitchd/main.go` — импорт internal/a11y + блок конструирования reconciler'а и SetA11ySink в OnConn (+11 строк, ноль новых веток)
- `red-evidence/08-05-task1-red.json` — RED-свидетельство Task 1 (RED_EVIDENCE_OK)

## Decisions Made

- **pushA11y — отдельный метод, а не инлайн**: строгий funlen актора (лимит 40 утверждений, у applySnapshot стало 43) и канон «per-fold push chain» (pushMenuSync) — фолд вынесен в метод с полным doc-комментарием WR-01 и одностороннего порядка actor.mu → sink.mu; семантика идентична плану
- **Стартовое применение через self-sync, не через новую строку**: FoldAppliedConfig в newActor отрабатывает ДО OnConn-установки sink'а — начальное Apply доставляет self-sync SetA11ySink; при no-config старте (нет watcher'а) фолда нет и Defaults неактивны — Apply не нужен
- **Линт-фиксы в GREEN-коммите** (прецедент 08-02/08-03): lll ×4 (перенос длинных t.Errorf/комментариев), goconst ×1 (литерал "zcode" ×7 → a11yListedApp), funlen ×1 (вынос pushA11y)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] lint-гейт строгого конфига: 6 замечаний после GREEN-реализации**
- **Found during:** Task 1 (грязная итерация после GREEN — golangci-lint красный до коммита)
- **Issue:** funlen ×1 (applySnapshot 43 > 40 утверждений после инлайн-фолда), goconst ×1 ("zcode" ×7), lll ×4 (два t.Errorf и две строки сообщений > 120)
- **Fix:** фолд вынесен в метод pushA11y (форма pushMenuSync), "zcode" → константа a11yListedApp, длинные строки разнесены
- **Files modified:** internal/session/actor.go, internal/session/actor_test.go
- **Verification:** mise exec -- golangci-lint run internal/session/... → 0 issues; mise run ci → exit 0
- **Committed in:** d4d019c (коммит Task 1 GREEN)

---

**Total deviations:** 1 auto-fixed (1 blocking — lint-гейт зелёной итерации).
**Impact on plan:** форма без изменения поведения; все behavior-кейсы плана зелёные как названы.

## TDD Gate Compliance

- Task 1: RED `test(08-05) 7e8ef39` предшествует GREEN `feat(08-05) d4d019c`; RED-свидетельство `RED_EVIDENCE_OK` (5 целевых тестов пали на push-count утверждениях против RED-стаба-формы — interface + install-only SetA11ySink без фолда и self-sync, прецедент 06-04/08-02; raw /tmp/08-05-t1-red-raw.txt, запись red-evidence/08-05-task1-red.json). TestActor_A11yNilSinkNoOp — green-by-design (отрицательное утверждение), честно записано. REFACTOR — без отдельного коммита (lint-фиксы вошли в GREEN по прецеденту 08-02/08-03).
- Итерационная заметка: первая форма TestActor_A11yNilSinkNoOp сверялась с AutoCorrectEnabled — дефолтный документ держит autocorrect OFF (D-54); санити-утверждение переведено на дефолтно-истинный SoundEnabled ДО фиксации RED-свидетельства.
- Tracer-гейт: verify трейсера (3 automated-чеков) перезапущен end-to-end после GREEN — exit 0, FOLD-OK, ACTIVE-SINGLE-SOURCE; режим end-of-phase/automated-only — продолжение без checkpoint.

## Issues Encountered

- None — `mise run ci` зелёный с первой попытки; известные pre-existing флаки (ctlsvc TestRun_OnConnHookCalledOnce, config TestWatch_BrokenAutocorrectBlockKeepsLastGood) не воспроизвелись, перезапуск не потребовался.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Кодовая часть a11y-магии завершена: схема (08-02) → reconciler (08-03) → snapshot/revert (08-04) → фолд+wiring (этот план); `goswitchd` применяет состояние при старте и на каждом hot reload, идемпотентно на всех трёх уровнях
- Остались доки/skill: 08-06 (README/CONFIG.md строки a11y.*), 08-07/08-08; живая UAT на столе — 08-VALIDATION Manual-Only (D6 coverage)
- Открытых блокеров нет

## Self-Check: PASSED

- internal/session/actor.go, internal/session/actor_test.go, cmd/goswitchd/main.go, red-evidence/08-05-task1-red.json — FOUND (в дифе плана, 252+)
- Commits 7e8ef39, d4d019c, 6014f7e — FOUND in git log; ledger gsd-plan-head-before-08-05 → 3 коммита
- Гейты: FOLD-OK, ACTIVE-SINGLE-SOURCE, WIRING-OK, NO-NEW-WATCHER, BUILD-VET-OK — FOUND
- mise run ci — exit 0 (первая попытка, /tmp/08-05-ci-run1.log)
- go.mod/go.sum — не тронуты (tidy-diff зелёный, ноль новых пакетов)

---
*Phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill*
*Completed: 2026-10-05*
