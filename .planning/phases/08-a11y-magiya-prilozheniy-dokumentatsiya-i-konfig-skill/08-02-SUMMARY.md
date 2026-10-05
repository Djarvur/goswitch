---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
plan: 02
subsystem: config
tags: [a11y, config-schema, yaml, re2, regex, strict-decode, hot-reload, tdd, zero-value-off]

# Dependency graph
requires:
  - phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill (план 08-01)
    provides: "SPEC §4 ревизия 2026-10-05 — записанный контракт секции a11y (enabled/apps, RE2, потолок 64, zero-off, компиляция на Load)"
  - phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
    provides: "Канон секции с валидацией (Autocorrect: потолки, сентинелы err113, TrimSpace-отказ до Compile, ошибки поле+индекс) и last-good watch-кейс 07-02"
provides:
  - "Секция A11y в internal/config: type A11y{Enabled bool yaml:\"enabled\"; Apps []string yaml:\"apps\"}; Config.A11y yaml:\"a11y\" — седьмая секция схемы"
  - "(a A11y) Active() bool — единственное определение ACTIVE-семантики (enabled && непустой список); потребитель — фолд актора плана 08-05"
  - "Валидация на Load: безусловный потолок maxA11yApps=64, blank-отказ ДО regexp.Compile, компиляция каждого паттерна — ошибки называют a11y.apps[i] (D-33)"
  - "Strict decode и hot-reload last-good распространены на секцию бесплатно — load.go/watch.go с нулём правок (прецедент 06-04)"
  - "Корпус a11y: 10 тестов (потолок/blank/regex/порядок/Active/Defaults/absent-off/strict-decode/декод-раундтрип/hot-reload)"
affects: [08-03 (демон-reconcile применяет Active()), 08-05 (фолд актора читает snap.A11y.Active()), 08-06 (docs/CONFIG.md строки a11y.*)]

# Actuals (#2632) — pairs with the plan's estimate to calibrate future estimates.
actuals:
  tokens: 5335    # chars/4 over the realized diff (21339 chars across the 4 changed files)
  tasks: 2
  commits: 3      # MEASURED: git rev-list --count gsd-plan-head-before-08-02..HEAD
plan_head_before: 26e820ac88122c689c9364978af37e40058347f8

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Новая секция схемы = копия канона Autocorrect: const-потолок рядом (:21-32), сентинелы в err113-блоке (:61-70), validate() с TrimSpace-отказом ДО Compile и ошибками «поле[индекс]» — в диспетче Config.Validate()"
    - "RED-стаб 06-04/07-02: RED-коммит несёт форму схемы (struct + поле Config + стаб Active()), GREEN — документированную семантику; strict-decode/last-good кейсы зелёные сразу (green-by-design)"

key-files:
  created:
    - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/red-evidence/08-02-task1-red.json
  modified:
    - internal/config/config.go
    - internal/config/config_test.go
    - internal/config/load_test.go
    - internal/config/watch_test.go

key-decisions:
  - "Потолок 64 безусловен по enabled (DoS-потолок, прецедент T-06-04-01) — тест пинит 65 паттернов при enabled:false → отказ"
  - "Compile-результаты НЕ кэшируются в схеме: валидация на Load — проверка; потребители (08-03/08-05) компилируют своё"
  - "Секция несёт ровно два ключа (enabled/apps): набор магии фиксирован D-8-6, per-app ключей нет — strict decode сам отвергает лишнее (тест пинит)"
  - "Тест 2 green-by-design (валидация на Load уже отвергает документ до apply) — честно зафиксирован по распоряжению плана, прецедент 02-04"

patterns-established:
  - "A11y.validate() — точная форма Autocorrect.validate() (:356-385): потолок «X has %d %w», цикл TrimSpace→Compile с «a11y.apps[%d] %q: %w»"

requirements-completed: ["D-8-1", "D-8-5"]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Секция a11y в схеме: type A11y (enabled/apps), Config.A11y (седьмая секция, doc-коммент «seven sections»), безусловный потолок maxA11yApps=64, сентинелы errA11yAppsOverCeil/Regex/Empty, A11y.validate() в диспетче — blank-отказ до компиляции, каждая ошибка называет a11y.apps и индекс"
    requirement: "D-8-5"
    verification:
      - kind: unit
        ref: "internal/config/config_test.go#TestValidate_A11yCeilingUnconditional"
        status: pass
      - kind: unit
        ref: "internal/config/config_test.go#TestValidate_A11yBlankPattern"
        status: pass
      - kind: unit
        ref: "internal/config/config_test.go#TestValidate_A11yRegexCompile"
        status: pass
      - kind: unit
        ref: "internal/config/config_test.go#TestValidate_A11yOrderIrrelevant"
        status: pass
    human_judgment: false
  - id: D2
    description: "Active()-семантика: enabled && непустой список — единственное определение ACTIVE; zero-value off; Defaults().A11y выключен (Enabled false, Apps nil)"
    requirement: "D-8-1"
    verification:
      - kind: unit
        ref: "internal/config/config_test.go#TestA11yActive"
        status: pass
      - kind: unit
        ref: "internal/config/config_test.go#TestDefaults_A11yOff"
        status: pass
    human_judgment: false
  - id: D3
    description: "Обратная совместимость и strict decode: документ без секции a11y грузится без изменений (zero-off, Pitfall 7); неизвестный ключ внутри секции инвалидирует ВЕСЬ файл (KnownFields(true) бесплатно); load.go/watch.go — ноль правок"
    requirement: "D-8-1"
    verification:
      - kind: unit
        ref: "internal/config/load_test.go#TestLoad_A11yAbsentMeansOff"
        status: pass
      - kind: unit
        ref: "internal/config/load_test.go#TestLoad_A11yStrictDecodeUnknownKey"
        status: pass
      - kind: unit
        ref: "internal/config/load_test.go#TestLoad_A11yDecodesEnabledAndApps"
        status: pass
      - kind: other
        ref: "git diff 26e820a..HEAD --name-only | grep internal/config/(load|watch).go → пусто (LOAD-WATCH-ZERO-EDITS)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Hot reload пин: битый паттерн a11y.apps → reload отклонён, last-good держится, WARN «config reload rejected»; починка применяется без рестарта (реальный config.Load)"
    requirement: "D-8-5"
    verification:
      - kind: unit
        ref: "internal/config/watch_test.go#TestWatch_BrokenA11yPatternKeepsLastGood"
        status: pass
    human_judgment: false
  - id: D5
    description: "Зелёная итерация на выходе плана: mise run ci (build + vet + golangci-lint strict + test -race ./... + tidy-diff) exit 0"
    verification:
      - kind: other
        ref: "mise run ci → CI_EXIT=0 (финальный прогон после lll-фиксов)"
        status: pass
    human_judgment: false

# Metrics
duration: 12min
completed: 2026-10-05
status: complete
---

# Phase 8 Plan 02: Схема секции a11y Summary

**Секция a11y в конфиг-схеме (enabled/apps RE2, безусловный потолок 64, blank-отказ до компиляции, Active() как единственный источник ACTIVE-семантики) со strict decode и hot-reload last-good бесплатно — load.go/watch.go не тронуты.**

## Performance

- **Duration:** 12 min
- **Started:** 2026-10-05T16:36:21Z
- **Completed:** 2026-10-05T16:48:52Z
- **Tasks:** 2
- **Files modified:** 4 (config.go +80/−6, config_test.go +170, load_test.go +84, watch_test.go +96)

## Accomplishments

- Схема несёт седьмую секцию `a11y`: `type A11y` (Enabled/Apps с yaml-тегами), `Config.A11y`, doc-комментарии по канону Autocorrect (D-8-1/D-8-5/D-8-6: zero-value off, набор магии фиксирован, per-app ключей нет); `(a A11y) Active()` — единственное определение ACTIVE-семантики для потребителя 08-05
- Валидация на Load (D-8-5/D-53 канон): безусловный потолок `maxA11yApps = 64` (65 паттернов при enabled:false → отказ — тест пинит), blank/whitespace-паттерн — отдельный сентинел `errA11yAppsEmpty` ДО `regexp.Compile` (ловушка «матчит всё»), некомпилируемый паттерн — `errA11yAppsRegex`, каждая ошибка называет `a11y.apps` и индекс (D-33); `Defaults()` несёт выключенную секцию
- Strict decode (D-33) и last-good hot reload (D-32) распространены на секцию БЕСПЛАТНО: load.go/watch.go — ноль правок (гейт `LOAD-WATCH-ZERO-EDITS` на всём дифе плана), оба свойства пинятся тестами
- TDD: RED `d894611` (6 целевых падений на утверждениях, `check tdd-red-evidence` → `RED_EVIDENCE_OK`) → GREEN `79d7b69` → hot-reload пин `c9efe29`; `mise run ci` зелёный (exit 0)

## Task Commits

Each task was committed atomically:

1. **Task 1 (трейсер): схема a11y сквозь Load** — RED `d894611` (test), GREEN `79d7b69` (feat); REFACTOR не потребовался (реализация легла в форму канона без чистки)
2. **Task 2: Hot-reload пин last-good + зелёная итерация** — `c9efe29` (test; включает lll-фиксы корпуса Task 1)

**Plan metadata:** следует отдельным docs-коммитом (SUMMARY + STATE + ROADMAP + red-evidence).

_Note: TDD-гейты: test(08-02) `d894611` предшествует feat(08-02) `79d7b69`._

## Files Created/Modified

- `internal/config/config.go` — тип A11y, Active(), Config.A11y (седьмая секция), maxA11yApps=64, сентинелы errA11yApps*, A11y.validate() в диспетче, Defaults off
- `internal/config/config_test.go` — корпус валидации/Active/Defaults (6 тестов + activeA11yDoc-хелпер)
- `internal/config/load_test.go` — корпус Load: absent-off, strict-decode, декод-раундтрип с compile-гейтом, blank-отказ на load-уровне (4 теста + a11yDocYAML-хелпер)
- `internal/config/watch_test.go` — TestWatch_BrokenA11yPatternKeepsLastGood (форма blocklist-кейса 07-02, реальный config.Load)
- `.planning/phases/08-.../red-evidence/08-02-task1-red.json` — RED-свидетельство Task 1 (RED_EVIDENCE_OK, TAP-транскрипция по прецеденту 07-04)

## Decisions Made

- Потолок 64 безусловен по enabled (T-08-02-01, прецедент T-06-04-01): тест пинит отказ 65 паттернов на спящей секции — ENABLED-обусловленный потолок пропускал бы гигантский список
- Compile-результаты не кэшируются в схеме: A11y.validate() — только проверка; компиляцию для матчинга выполнят потребители (08-03/08-05)
- Два дополнительных load-уровень кейса (TestLoad_A11yDecodesEnabledAndApps, TestLoad_A11yBlankPatternRejected) сверх восьми именованных планом — зеркалят форму TestLoad_AutocorrectRangeViolationRejected и пинят истину «компиляция каждого паттерна на Load» на уровне Load, а не только Validate

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] lll: две строки корпуса Task 1 превышали лимит 120 символов**
- **Found during:** Task 2 (зелёная итерация `mise run ci` — lint-гейт красный)
- **Issue:** golangci-lint lll: config_test.go:839 (122 символа, таблица TestA11yActive) и load_test.go:467 (132 символа, вызов Load с anchored-паттерном)
- **Fix:** перенос case в многострочную форму / вынос YAML-тела в переменную `anchored`
- **Files modified:** internal/config/config_test.go, internal/config/load_test.go
- **Verification:** mise run ci exit 0
- **Committed in:** c9efe29 (коммит Task 2)

---

**Total deviations:** 1 auto-fixed (1 blocking).
**Impact on plan:** косметика формата тестовых строк; ни поведения, ни объёма.

## TDD Gate Compliance

- Task 1: RED `test(08-02) d894611` предшествует GREEN `feat(08-02) 79d7b69`; RED-свидетельство `RED_EVIDENCE_OK` (6 целевых тестов пали на запланированных утверждениях: nil-ошибка на потолке/blank/regex, Active()-стаб; raw-вывод /tmp/08-02-t1-red-raw.txt, запись red-evidence/08-02-task1-red.json). REFACTOR — без изменений, коммита нет (tdd.md: коммит только при изменениях).
- Task 2: **green-by-design, честно зафиксирован** — кейс TestWatch_BrokenA11yPatternKeepsLastGood зелёный сразу (валидация на Load уже отвергает документ до apply; watcher-кода не требуется, D-32 распространяется автоматически), по прямому распоряжению плана и прецеденту 02-04 (TestActor_MixedWordUntouched). RED-свидетельство не фабриковалось; коммит формы `test(08-02)` `c9efe29`.

## Issues Encountered

- None — известные pre-existing флаки (ctlsvc TestRun_OnConnHookCalledOnce, config TestWatch_BrokenAutocorrectBlockKeepsLastGood) в прогоне `mise run ci` не воспроизвелись; единственный прогон после lll-фиксов зелёный с первого раза.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `snap.A11y.Active()` — готовый источник ACTIVE-семантики для фолда актора (08-05); применение магии — 08-03; snapshot/revert — 08-04; строки a11y.* в docs/CONFIG.md — 08-06 (yaml-теги и потолок схемы зафиксированы)
- Волна-гейт: схема самодостаточна (`go build ./...` зелёный, потребителей у поля пока нет)
- Открытых блокеров нет

## Self-Check: PASSED

- internal/config/config.go / config_test.go / load_test.go / watch_test.go — FOUND (в дифе плана, 424+/6−)
- Commits d894611, 79d7b69, c9efe29 — FOUND in git log; ledger gsd-plan-head-before-08-02 → 3 коммита
- Гейт «семь секций» — FOUND (config.go:197)
- LOAD-WATCH-ZERO-EDITS — git diff 26e820a..HEAD не содержит load.go/watch.go
- mise run ci — exit 0 (финальный прогон 2026-10-05T16:47Z)

---
*Phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill*
*Completed: 2026-10-05*
