---
phase: 06-avtokorrekcija-opcionalno
plan: 04
subsystem: config
tags: [yaml, strict-decode, d-33, hot-reload, d-32, validation, ceilings, default-off, d-54, docs, tdd]

# Dependency graph
requires:
  - phase: 03 (internal/config MACR-секция)
    provides: структурный шаблон схема→Defaults→validate с сентинелами, maxMACRApps-потолок, complete-document контракт 03-02 (без overlay-дефолтов в Load), D-32/D-33 машина watcher'а
  - phase: 06-avtokorrekcija-opcionalno (06-01)
    provides: spec-delta §10/§11 + ADR-007 — opt-in, default off (D-54), конъюнкция D-53
provides:
  - Секция Autocorrect в YAML-схеме (enabled/apps/min_word_len/trigram_margin/trigram_floor): strict decode D-33 и hot reload D-32 наследуются БЕЗ нового кода — load.go/watch.go с нулём правок, битый autocorrect-блок инвалидирует весь файл (last-good держится)
  - Default off везде (D-54): Defaults() = Autocorrect{Enabled:false, Apps:nil, MinWordLen:4, TrigramMargin:2.0, TrigramFloor:1.0}; нулевое значение = off; отсутствующая секция декодируется в ноль без ошибок (никакого overlay в Load)
  - Валидация с потолками по прецеденту MACR: maxAutocorrectApps=64 безусловно (DoS-гвард T-06-04-01); диапазоны [2,16] / >0 / margin≥floor — только на АКТИВНОЙ секции; каждая ошибка именует своё поле (сентинелы errAutocorrectAppsOverCeil/errAutocorrectMinWordLenRange/errAutocorrectThresholdRange)
  - Публичный контракт docs/CONFIG.md: пять секций, 5 строк таблицы байт-со схеме, два примера (дефолты + включённая секция), семантика отсутствующих ключей (= off), правило приватности «слово никогда не попадает ни в лог, ни в status»
  - Кросс-плановый контракт данных для 06-06 (applySnapshot-фолд) и 06-05 (DefaultParams): имена/типы полей запинены (defaults 4/2.0/1.0 — менять парой мест)
affects: [06-05 detect (DefaultParams + золотой корпус порогов), 06-06 session (Options-фолд snap.Autocorrect.*), 06-07 ctlsvc (autocorrect-статус-токены), 06-08 (e2e-кейсы)]

# Actuals (#2632)
actuals:
  tokens: 7900        # chars/4 over the realized diff (31 600 bytes) — оценка плана 42 000 пере-оценила: аддитивная секция по готовому MACR-шаблону
  tasks: 2
  commits: 3          # MEASURED: git rev-list --count gsd-plan-head-before-06-04..HEAD

# Tech tracking
tech-stack:
  added: []           # ноль новых зависимостей (yaml.v3 уже в go.mod; tidy-diff зелёный)
  patterns:
    - "MACR-шаблон расширения схемы: struct с yaml-тегами + поле Config + Defaults-off + (секция).validate с сентинелами + одна строка в цепочке Validate"
    - "Активно-секционный гейт валидации порогов: диапазоны кусают только enabled+непустой список (единственная форма, которая может сработать), потолок списка — безусловно"

key-files:
  created:
    - .planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-04-task1-red.json
  modified:
    - internal/config/config.go
    - internal/config/config_test.go
    - internal/config/load_test.go
    - internal/config/watch_test.go
    - docs/CONFIG.md

key-decisions:
  - "Диапазоны порогов валидируются только на АКТИВНОЙ секции (enabled + непустой apps): запиненные планом поведения (нулевая секция валидна; enabled+пустой список валиден) делают безусловную проверку диапазонов самопротиворечивой; потолок white-list при этом безусловный (гигантский список — DoS-вектор даже в спящем слое, T-06-04-01), а АКТИВНАЯ секция с пропущенными порогами честно отвергается (complete-document контракт 03-02)"
  - "Каждая пороговая ошибка именует СВОЁ поле: trigram_floor ≤ 0 → autocorrect.trigram_floor, а не комбинированную форму из action-текста плана (только margin) — must_haves-истина «каждая ошибка именует своё поле» старше стенографии действия"
  - "RED-stub паттерн (прецедент 06-03 continuity-pin): RED-коммит несёт только форму схемы (struct + поле Config + нулевой Defaults), GREEN ставит документированные дефолты/потолок/валидацию; тесты распространения D-32/D-33 зелёные by-design — strict decode и last-good reload не потребовали ни строчки в load.go/watch.go"
  - "requirements-completed — провизорная метка плана: MACR-ACL и SPEC §10/§11 разделяются с 06-05..06-08 (SUMMARY отсутствуют) — REQUIREMENTS.md чекбоксы остаются последнему объявившему плану (шлагбаум #2388); формальная фиксация REQ-ID — при new-milestone v1.1.0 (D-51, прецедент 06-03)"

patterns-established:
  - "Активно-секционный гейт валидации: потолок списка безусловно, диапазоны параметров — только когда секция может сработать (enabled + непустой список); нулевое значение = off остаётся валидным всегда"
  - "Сквозной hot-reload тест на реальном парсере: config.Load как WithLoader над синтетическим фидом событий и настоящим temp-файлом — свойство пинится на схеме, а не на скриптованном лоадере"

requirements-completed: ["MACR-ACL (app×role политика)", "SPEC §10/§11 (spec-delta)"]  # конфигурационная половина MACR-ACL (условие 1 конъюнкции D-53) + исполнение условий возврата §11; ID разделяются с 06-05..06-08 — чекбоксы REQUIREMENTS.md при последнем объявившем плане (#2388), формальная фиксация при v1.1.0 (D-51)

coverage:
  - id: D1
    description: "Секция Autocorrect со strict decode D-33, дефолтами off (Defaults: enabled:false, apps:nil, 4/2.0/1.0) и валидацией с потолками: 64-entry white-list валиден, 65 — отказ с именем поля и лимитом; диапазоны [2,16] и >0 с margin≥floor на активной секции; границы 2/16 и margin==floor валидны"
    requirement: "MACR-ACL (app×role политика)"
    verification:
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestDefaults_AutocorrectOff"
        status: pass
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestValidate_AutocorrectAppsCeil"
        status: pass
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestValidate_AutocorrectRanges"
        status: pass
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestValidate_AutocorrectRangeBoundaries"
        status: pass
      - kind: unit
        ref: "tests/internal/config/load_test.go#TestLoad_AutocorrectRangeViolationRejected"
        status: pass
    human_judgment: false
  - id: D2
    description: "Распространение D-32/D-33 на новую секцию БЕЗ нового кода: опечатка в ключе секции инвалидирует весь документ; ломаный autocorrect-блок (margin<floor) на reload отклоняется целиком — last-good держится, WARN «config reload rejected», починка применяется без рестарта; load.go/watch.go — ноль правок (git diff пуст)"
    requirement: "SPEC §10/§11 (spec-delta)"
    verification:
      - kind: unit
        ref: "tests/internal/config/load_test.go#TestLoad_StrictDecodeRejectsTypo"
        status: pass
      - kind: unit
        ref: "tests/internal/config/watch_test.go#TestWatch_BrokenAutocorrectBlockKeepsLastGood"
        status: pass
      - kind: other
        ref: "git diff bf7cc57..HEAD -- internal/config/load.go internal/config/watch.go — пусто (ноль правок, D-32/D-33 наследуются)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Декод-контракт default off (D-54): документ без секции декодируется в нулевую Autocorrect без ошибок (никакого overlay в Load); enabled:true + пустой apps — валидный, но глухой документ (пустой список глуше флага, T-06-04-03); порядок apps не влияет на семантику (set-семантика)"
    requirement: "SPEC §10/§11 (spec-delta)"
    verification:
      - kind: unit
        ref: "tests/internal/config/load_test.go#TestLoad_DocumentWithoutSectionDecodesOff"
        status: pass
      - kind: unit
        ref: "tests/internal/config/load_test.go#TestLoad_EmptyAppsEnabledIsValid"
        status: pass
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestValidate_AutocorrectDormantShapesValid"
        status: pass
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestValidate_AppsOrderIrrelevant"
        status: pass
    human_judgment: false
  - id: D4
    description: "docs/CONFIG.md — публичный контракт пяти секций: вводная «exactly five sections», семантика отсутствующих ключей (= off), 5 строк таблицы байт-со схеме (имена/дефолты/диапазоны), пример дефолтов и пример включённой секции, правило приватности — набранное и скорректированное слово никогда не попадает ни в один уровень лога, ни в goswitchctl status"
    requirement: "SPEC §10/§11 (spec-delta)"
    verification:
      - kind: other
        ref: "греп-гейты плана: min_word_len/trigram_margin/five sections/никогда в docs/CONFIG.md — CONFIG-DOC-OK; таблица сверена со схемой построчно (defaults false/[]/4/2.0/1.0, диапазоны [2,16], >0, ≥floor)"
        status: pass
      - kind: other
        ref: "mise run ci (build + vet + golangci-lint strict + go test -race ./...) — зелёный на обоих задачах"
        status: pass
    human_judgment: true
    rationale: "Полнота/читабельность пользовательской документации — суждение верификатора; автоматика доказывает присутствие и байт-согласованность со схемой, но не адекватность формулировок для пользователя"

# Metrics
duration: 10min
completed: 2026-10-01
status: complete
---

# Phase 6 Plan 4: internal/config — секция autocorrect (D-54) Summary

**Конфиг-поверхность автокоррекции: пятая секция схемы со strict decode D-33 и hot reload D-32, унаследованными без единой правки load.go/watch.go; default off везде (enabled:false + пустой white-list + пороги 4/2.0/1.0), потолок 64 и поименованные ошибки валидации по прецеденту MACR; docs/CONFIG.md дополнен до публичного контракта пяти секций с правилом приватности**

## Performance

- **Duration:** 10 min (21:35–21:45 UTC)
- **Started:** 2026-10-01T21:35:28Z
- **Completed:** 2026-10-01T21:45:36Z
- **Tasks:** 2 (Task 1 tdd — строгий RED→GREEN; Task 2 auto)
- **Files modified:** 5 (+1 red-evidence запись)

## Accomplishments
- `type Autocorrect` (enabled/apps/min_word_len/trigram_margin/trigram_floor) — пятая секция Config; нулевое значение = off (D-54): документ без секции декодируется в ноль без ошибок, никакого overlay-слияния с Defaults в Load (прецедент 03-02)
- Defaults() несёт документированную off-форму с порогами 4/2.0/1.0 и комментарием-кросс-ссылкой: значения пинируются золотым корпусом internal/detect (план 06-05) — менять ТОЛЬКО парой мест
- (Autocorrect).validate: потолок maxAutocorrectApps=64 безусловно (T-06-04-01); диапазоны [2,16] и >0 с margin≥floor — на активной секции (enabled + непустой white-list), так что никакая комбинация отсутствующих/пустых значений не включает фичу (T-06-04-03, empty-apps+enabled валиден, но глух); каждая ошибка именует поле (сентинелы err113-дисциплины)
- Strict decode и hot reload распространены автоматически: опечатка `enbled` отвергает весь файл; ломаный блок (margin<floor) на reload отклоняется, last-good держится, WARN «config reload rejected», починка применяется — **load.go/watch.go с нулём правок** (гейт плана: git diff пуст)
- docs/CONFIG.md: «exactly five sections», семантика отсутствующих ключей (= off), 5 строк таблицы, два полных примера, приватность — «слово никогда (никогда) не попадает ни в один уровень лога, ни в goswitchctl status» (строже D-20/D-21, Pitfall 6)
- mise run ci (build+vet+lint strict+test -race) зелёный на закрытие обеих задач; ноль новых зависимостей

## Task Commits

Each task was committed atomically (TDD: RED → GREEN for Task 1):

1. **Task 1: Секция Autocorrect — схема, дефолты off, валидация** — `d616906` (test, RED) → `3706b60` (feat, GREEN)
2. **Task 2: Hot-reload пин + docs/CONFIG.md** — `8981397` (test+docs)

**Plan metadata:** _этот коммит — docs(06-04)_

_Note: REFACTOR-коммит не понадобился — линтер-риффы (appGedit-константа, разбиение boundaries-сабтеста, assertDecodeDefaults-хелпер) вошли в GREEN-коммит Task 1 (зелёная итерация D-08 до коммита)._

## TDD Gate Compliance

| Гейт | Коммит | Статус |
|------|--------|--------|
| RED | `d616906` test(06-04) | RED_EVIDENCE_OK — TestDefaults_AutocorrectOff упал на целевом утверждении (MinWordLen = 0, want 4) против RED-stub; exit 1; запись red-evidence/06-04-task1-red.json |
| GREEN | `3706b60` feat(06-04) | весь пакет зелёный под -race, 32 top-level теста |
| REFACTOR | — | не требуется (риффы вошли в GREEN, поведение не менялось) |

Green-by-design в RED (запинено, не чинено — свойства уже держались): TestLoad_StrictDecodeRejectsTypo (D-33 уже отвергает опечатку секции), TestLoad_DocumentWithoutSectionDecodesOff (overlay в Load не существует), TestLoad_EmptyAppsEnabledIsValid, TestValidate_AutocorrectDormantShapesValid, TestValidate_AppsOrderIrrelevant; все 19 дофазовых тестов остались зелёными.

## Files Created/Modified
- `internal/config/config.go` — type Autocorrect, Config.Autocorrect («five sections»), Defaults-off 4/2.0/1.0, maxAutocorrectApps=64, три сентинела, (Autocorrect).validate, хвост цепочки Validate
- `internal/config/config_test.go` — корпус: дефолты off, потолок 64, диапазоны+границыы, спящие формы, порядок apps
- `internal/config/load_test.go` — декод-корпус: полный документ с секцией, документ без секции (= ноль), strict-decode опечатка, empty-apps+enabled, range-отказ на загрузке
- `internal/config/watch_test.go` — TestWatch_BrokenAutocorrectBlockKeepsLastGood на реальном config.Load
- `docs/CONFIG.md` — секция autocorrect: пять секций, таблица, примеры, приватность

## Decisions Made
См. key-decisions: активно-секционный гейт диапазонов (потолок — безусловно), поименная ошибка trigram_floor, RED-stub паттерн, провизорная метка requirements (шлагбаум #2388 — ID разделяются с 06-05..06-08).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Безусловная проверка диапазонов из action-текста противоречила запиненным поведениям плана**
- **Found during:** Task 1 (GREEN)
- **Issue:** action-текст перечисляет проверки порогов без условия; в буквальном виде TestLoad_EmptyAppsEnabledIsValid ({enabled:true, apps:[]} → MinWordLen=0) и prohibition 1 (нулевая секция валидна) проваливаются — план внутренне противоречив
- **Fix:** диапазоны кусают только АКТИВНУЮ секцию (enabled + непустой white-list — единственная форма, которая может сработать по D-53); потолок списка оставлен безусловным
- **Files modified:** internal/config/config.go
- **Verification:** весь корпус зелёный, включая оба запиненных поведения; активная секция с пропущенными порогами честно отвергается на загрузке
- **Committed in:** 3706b60

**2. [Rule 1 - Bug] Ошибка trigram_floor ≤ 0 именовала бы чужое поле**
- **Found during:** Task 1 (GREEN)
- **Issue:** action-текст задаёт одну комбинированную ошибку с именем autocorrect.trigram_margin для всех трёх пороговых нарушений — нарушает must_haves-истину «каждая ошибка именует своё поле»
- **Fix:** отдельные строки ошибок: floor → autocorrect.trigram_floor, margin (≤0 или <floor) → autocorrect.trigram_margin; sentinel errAutocorrectThresholdRange общий
- **Files modified:** internal/config/config.go
- **Verification:** TestValidate_AutocorrectRanges — каждое нарушение возвращает своё имя поля
- **Committed in:** 3706b60

---

**Total deviations:** 2 auto-fixed (2 × Rule 1 — сверки внутриплановых противоречий в пользу behavior-контракта плана).
**Impact on plan:** оба фикса необходимы для согласованности с собственным behavior-блоком плана; область та же, расширений нет.

## Issues Encountered
- Структура со slice-полем несравнима (got != want не компилируется) — пины дефолтов переписаны на по-полевые сравнения (точнее и в духе corpus-стиля)
- Строгий линт на тест-корпусе: lll 121/120, funlen 72/60, cyclop 18/15, goconst «org.gnome.Gedit»×3 — устранены до коммита (зелёная итерация директивы 2)
- check tdd-red-evidence парсит только канонический TAP-хвост (# tests / # pass / # fail) — запись переоформлена, вердикт RED_EVIDENCE_OK

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- 06-05 (detect + DefaultParams) получает запиненную форму полей и дефолты 4/2.0/1.0 с обязательством менять парой; 06-06 (actor-фолд) — snap.Autocorrect.{Enabled,Apps,MinWordLen,TrigramMargin,TrigramFloor}; конъюнкция потребителя обязана требовать непустой список (здесь пин «валиден, но глух»)
- Оба requirement-ID остаются для последнего объявившего плана (06-05..06-08 без SUMMARY) — шлагбаум #2388

## Self-Check: PASSED

- Все key-files существуют на диске (config.go, config_test.go, load_test.go, watch_test.go, docs/CONFIG.md, red-evidence/06-04-task1-red.json, SUMMARY)
- Все коммиты задач найдены в git log (d616906, 3706b60, 8981397), метадата-коммит — HEAD; измерено `git rev-list --count gsd-plan-head-before-06-04..HEAD` = 4; рабочее дерево чистое
- TDD-гейт: RED `d616906` предшествует GREEN `3706b60`; check tdd-red-evidence — RED_EVIDENCE_OK (TestDefaults_AutocorrectOff, target_test_failed)
- Верификация плана перезапущена на финальном дереве: `go test ./internal/config/ -race -count=1` ok; греп-гейты SCHEMA-OK / CONFIG-DOC-OK; `mise run ci` зелёный; `git diff bf7cc57..HEAD -- internal/config/load.go internal/config/watch.go` пуст (ноль правок)

---
*Phase: 06-avtokorrekcija-opcionalno*
*Completed: 2026-10-01*
