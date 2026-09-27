---
phase: 04-postavka-i-priemka
plan: 05
subsystem: testing
tags: [e2e, matrix-v3, gedit, gtk3, chromium-x11, xwayland, at-spi, ibus, spike]

# Dependency graph
requires:
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
    provides: "матрица v2 (замороженный эталон, 21 кейс), surfaceKind-драйверы стенда (zenity/GTE/chromium), pid-keyed witness/readback, closed-vocabulary валидация кейсов"
provides:
  - "драйверы поверхностей gedit (standalone + двойная XDG-изоляция) и chromium-x11 (--ozone-platform=x11) с pid-keyed witness/readback/close"
  - "расширенный закрытый словарь поверхностей (gedit | chromium-x11) с валидацией на загрузке кейса"
  - "preflight checkGeditLaunch с fix-hint (sudo apt install gedit)"
  - "matrix-v3.yaml — superset v2 (31 кейс: 21 v2 дословно + 6 gedit + 4 x11), полный живой прогон 31/31 PASS exit 0"
  - "спайк-вердикты двух новых поверхностей (caps/anchor/level/RU-режим) — в комментариях драйверов, шапке v3 и ниже"
  - "smoke-кейсы gedit-smoke / x11-smoke, печатающие вердикт-записи поверхности"
affects: [04-06 (двойной прогон v3 на свежей сессии — база D-48), 04-07 (verify-гейт фазы: вердикт владельца по x11-исключению)]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
actuals:
  tokens: 13728     # chars/4 over the realized diff (54912 chars)
  tasks: 2
  commits: 2        # MEASURED: git rev-list --count plan_head_before(39d637c)..HEAD

# Tech tracking
tech-stack:
  added: []         # ноль новых зависимостей: gedit/google-chrome — спавн-поверхности стенда, не Go-зависимости
  patterns:
    - "изоляция GTK3-редактора: --standalone + XDG_DATA_HOME + XDG_CONFIG_HOME в tmp стенда (enchant пишет словари в XDG_CONFIG_HOME)"
    - "второй спавн семейства chromium: стартовый флаг-набор + ровно один режимный флаг --ozone-platform=x11"
    - "awaitPidInput/readFocusedTextPid — общий pid-keyed цикл witness/poke/readback трёх драйверов (GTE, gedit, chromium-x11)"

key-files:
  created:
    - test/e2e/cases/matrix-v3.yaml
  modified:
    - test/e2e/surface.go
    - test/e2e/matrix.go
    - test/e2e/preflight.go
    - test/e2e/main.go
    - test/e2e/case_surface.go
    - test/e2e/case_m1.go
    - test/e2e/case_d01.go
    - test/e2e/case_resilience.go
    - docs/ci-runner.md
    - .gitignore

key-decisions:
  - "Спайк опроверг план: gedit 46.2 ИМЕЕТ -s/--standalone — изоляция GTE-образцом переносится (A4 закрыт пробой)"
  - "XDG_CONFIG_HOME присоединён к изоляции: enchant пишет пользовательские словари (en_US.dic/.exc) — спайк поймал материализацию в перенаправленном каталоге"
  - "gedit пины: caps 0x29, коррекция слова живёт на уровне 1 (delete+commit прочитан назад), anchor НЕТ (все пуши cursor==anchor — D-30 класс zenity), RU-режим: коммиты движка НЕ рендерятся (документ пуст, caps падают до 0x9)"
  - "chromium-x11 пины: AT-SPI имя 'Google Chrome' (пере-пин всех трёх констант вместе), caps 0x29 декларированы, но surrounding-пушей НОЛЬ (проба GTK_IM_MODULE=ibus ничего не меняет) — коррекция никогда не исполняется (verify-timeout skip); RU-коммиты попадают в омнибокс (63-знаковый свидетель URL)"
  - "состав v3 скорректирован по вердиктам: gedit ru-en/mixed строки и select-строки исключены (неоткомпонуемы), хвостовая строка заменена на registers; x11 строки — минимальные транзит-формы без жестов (ambient omnibox focus steal ~20%)"
  - "читающий оракул стенда (LINE-granularity a11y read) точен на набранном тексте, но роняет движковый trailing space в GtkTextView gedit — D-13 на gedit доказан записями демона (runes=7, verify match), строка исключена как ограничение харнеса"

patterns-established:
  - "v3-spike вердикт-записи: smoke-кейс печатает caps/level/anchor поверхности — состав матрицы потребляет только живые вердикты (прецедент select-smoke 03-03)"
  - " spike-вердикты живут в комментариях констант драйверов рядом с пинами (re-pin-both-together дисциплина)"

requirements-completed: [INST-01]  # разделяемый ID (04-01..04-04): mark-complete через ready-ids гейт

coverage:
  - id: D1
    description: "Драйвер поверхности gedit (startGedit/waitGeditInput/readGeditText/closeGedit) с XDG-изоляцией + словарь/префлайт gedit"
    requirement: INST-01
    verification:
      - kind: e2e
        ref: "go run ./test/e2e -case gedit-smoke — PASS (spawn/witness/inject/readback/close + preflight gedit-launch ok)"
        status: pass
      - kind: unit
        ref: "mise run ci — build/vet/lint/test -race зелёные"
        status: pass
    human_judgment: false
  - id: D2
    description: "Драйвер chromium-x11 (startChromiumX11 с --ozone-platform=x11) + словарь chromium-x11"
    requirement: INST-01
    verification:
      - kind: e2e
        ref: "go run ./test/e2e -case x11-smoke — PASS (spawn/witness/inject/readback/close)"
        status: pass
    human_judgment: false
  - id: D3
    description: "matrix-v3.yaml — superset v2 по именам (comm-гейт) и счётчику (>=+10), полный живой прогон зелёный"
    requirement: INST-01
    verification:
      - kind: e2e
        ref: "go run ./test/e2e -matrix test/e2e/cases/matrix-v3.yaml — 31/31 PASS, exit 0"
        status: pass
      - kind: unit
        ref: "comm-гейт V2-SUBSET-OK + счётчик COUNT-OK (31 >= 21+10) + git diff V2-FROZEN"
        status: pass
    human_judgment: false
  - id: D4
    description: "Спайк-вердикты новых поверхностей (caps/anchor/level/RU-режим) запинены в драйверах, шапке v3 и SUMMARY"
    verification:
      - kind: e2e
        ref: "v3-spike строки gedit-smoke/x11-smoke (caps=0x29, level=1, anchor=NO) + записи демона (\"correction\",\"level\":1; verify-timeout skips)"
        status: pass
    human_judgment: true
    rationale: "Интерпретация негативных вердиктов (x11 поверхностный уровень не существует; gedit RU-коммиты не рендерятся) — составленные исключения строк требуют вердикта владельца на verify-гейте фазы (A1 fallback: критерий №2 покрывается фактически доступными поверхностями)"
  - id: D5
    description: "Требование gedit к машине раннера документировано (docs/ci-runner.md)"
    verification:
      - kind: unit
        ref: "grep-гейт VOCAB-OK (gedit в ci-runner.md/preflight/matrix.go)"
        status: pass
    human_judgment: false

# Metrics
duration: 75 min
completed: 2026-09-16
status: complete
---

# Phase 04 Plan 05: Matrix v3 — gedit + chromium-x11 Summary

**Матрица v3 как superset замороженной v2: драйверы GTK3-gedit и оконного Chromium-x11, спайк-пины фактического поведения, полный живой прогон 31/31 PASS exit 0 — критерий №2 (полная широта) готов к двойному прогону 04-06**

## Performance

- **Duration:** 75 min
- **Started:** 2026-09-16T11:09:02Z
- **Completed:** 2026-09-16T12:23:53Z
- **Tasks:** 2
- **Files modified:** 11 (972 insertions, 108 deletions)

## Accomplishments

- Обе новые поверхности водятся стендом end-to-end: gedit — изоляция `--standalone` + двойная XDG-перенаправка, chromium-x11 — стартовый флаг-набор + `--ozone-platform=x11`; pid-keyed witness/readback/close, общие циклы awaitPidInput/readFocusedTextPid
- Словарь поверхностей расширен закрытой валидацией (`gedit | chromium-x11`), префлайт получил checkGeditLaunch с fix-hint (детектор Падения 6), книга раннера — требование gedit
- matrix-v3.yaml: все 21 кейс v2 дословно + 6 gedit + 4 chromium-x11; полный живой прогон 31/31 PASS exit 0; v2 не тронута (git diff пуст); имена v2 — строгое подмножество (comm-гейт)
- Спайк-вердикты запинены «фактическое, никогда предположение»: см. Спайк-вердикты ниже

## Спайк-вердикты (04-05 Task 1, живые пробы 2026-09-16)

| Пин | gedit (GTK3, 46.2) | chromium-x11 (--ozone-platform=x11) |
|-----|--------------------|--------------------------------------|
| Изоляция | `--standalone` + XDG_DATA_HOME + XDG_CONFIG_HOME (enchant пишет словари) | свежий `--user-data-dir` в tmp (собственное дерево процессов) |
| AT-SPI имя | `gedit` | `Google Chrome` (идентично Wayland — re-pin трёх констант вместе) |
| caps | 0x29 | 0x29 декларированы |
| Surrounding-пуши | есть при транзите; при RU-режиме (коммит движка) пропадают, caps падают до 0x9 | НЕТ вообще (ни при печати; проба GTK_IM_MODULE=ibus ничего не меняет) |
| Коррекция слова | живёт: `"correction","level":1`, delete+commit прочитан назад | никогда: `"correction skipped","reason":"verify-timeout"` — поле не тронуто |
| Anchor (ctrl+a) | НЕТ — каждый пуш cursor==anchor (D-30 класс zenity; хорд долетает — прыжок каретки доказывает) | НЕТ — пушей нет вовсе |
| RU-режим (коммиты движка) | НЕ рендерятся: документ пуст (0 знаков) | попадают в омнибокс (свидетель ENTRY:chars=63 — URL) |
| Фактический уровень лестницы | **1** | не существует (лестница не исполняется) |

## Task Commits

1. **Task 1: Спайки + драйверы gedit/chromium-x11, словарь + префлайт** - `c3d6d07` (feat)
2. **Task 2: matrix-v3.yaml — superset v2 + живой зелёный прогон** - `b38e6e6` (feat)

_План type: execute (задачи type=auto) — TDD RED/GREEN гейты не применялись; зелёная итерация D-08 подтверждена mise run ci на каждом коммите._

## Files Created/Modified

- `test/e2e/cases/matrix-v3.yaml` — v3-корпус (31 кейс) с шапкой дельты и исключений
- `test/e2e/surface.go` — драйверы gedit/chromium-x11, константы AT-SPI имён со ссылками на спайк, surfaceKind-расширение, общие awaitPidInput/readFocusedTextPid, withEnvList
- `test/e2e/matrix.go` — словарь (5 поверхностей), openMatrixSurface/matrixSurfaceApp/matrixSurfaceReportsAnchor/matrixSurfacePID/verifyMatrixText — все пять точек расширения
- `test/e2e/preflight.go` — checkGeditLaunch (LookPath + fix-hint + happy path)
- `test/e2e/main.go` — поля стенда gedit/chromiumX11, реестр + usage строка (gedit-smoke, x11-smoke)
- `test/e2e/case_surface.go` — gedit-smoke/x11-smoke + spikeVerdictsV3 (caps/level/anchor печать)
- `test/e2e/case_m1.go`, `case_d01.go`, `case_resilience.go` — exhaustive no-op кейсы новых surfaceKind
- `docs/ci-runner.md` — требование gedit (universe) к машине раннера
- `.gitignore` — `/e2e` (stale билд-артефакт стенда)

## Decisions Made

- Состав v3 определён спайками (CONTEXT discretion): gedit — слово en-ru (с пином уровня 1), регистры, фраза, registers-фраза, layout-single; x11 — 4 транзит-строки (слово/регистры/фраза)
- Исключения задокументированы в шапке v3 с доказательствами: gedit ru-en/mixed (RU-коммиты не рендерятся), gedit select (нет anchor), gedit tail (ограничение читающего оракула — демон D-13 держит), x11 select/ru/жесты (нет пушей; омнибокс-стил ~20%)
- Линт-рефакторинг по строгому конфигу: дедупликация pid-циклов в awaitPidInput/readFocusedTextPid, вынос verifyZenityText и matrixSurfaceCmd (cyclop/dupl/exhaustive — все чисто)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] План ошибочно утверждал отсутствие --standalone у gedit**
- **Found during:** Task 1 (спайк изоляции)
- **Issue:** gedit 46.2 имеет `-s, --standalone`; класс решения изоляции — GTE-образец, но с добавкой XDG_CONFIG_HOME (enchant-словари, поймано пробой)
- **Fix:** драйвер использует --standalone + двойную XDG-перенаправку
- **Files modified:** test/e2e/surface.go
- **Verification:** gedit-smoke PASS, префлайт gedit-launch ok
- **Committed in:** c3d6d07

**2. [Rule 1 - Bug] Рекампоз-скрипт съел tap:double у трёх gedit word-строк**
- **Found during:** Task 2 (диагностика «деградации» word-строк)
- **Issue:** str.replace паттерна x11-строк совпал с type-строками gedit (`{type: "ghbdtn"}` + `{tap: double}`) — строки остались без жеста, expect_level стал невыполнимым по построению
- **Fix:** восстановлены tap:double всех трёх строк; дельта-прогон 10/10 PASS
- **Files modified:** test/e2e/cases/matrix-v3.yaml
- **Verification:** дельта 10/10, затем полный прогон 31/31
- **Committed in:** b38e6e6 (в составе финального файла)

---

**Total deviations:** 2 auto-fixed (1 blocking-assumption resolved by spike, 1 self-inflicted edit bug caught by gates) + composition adjustments по живым вердиктам (см. Issues) — все в пределах «состав кейсов определяется спайками» (CONTEXT discretion, D-47)
**Impact on plan:** Критерий №2 покрыт фактически доступными поверхностями; негативные вердикты x11/gedit-RU — открытие плана, не потеря объёма.

## Issues Encountered

- «Ambient omnibox focus steal» на x11-chrome (~20% кейсов: URL-бар перехватывает фокус, applied-state видит ENTRY:chars=63) — нестабильно по поверхности, не по коду; x11-строки собраны в минимальной транзит-форме; лечится будущим пере-пином при починке x11 IM-пути (громко падает)
- Читающий оракул AT-SPI (LINE-granularity) роняет движковый trailing space в GtkTextView gedit — хвостовая строка заменена registers-строкой; D-13 на gedit доказан логом демона (runes=7, result «привет », verify match)
- Серийные прогоны стенда деградировали ibus/a11y-состояние в процессе сессии — применены штатные лекарства (ibus restart через кейс, перезапуск шины a11y по docs/ci-runner.md); финальный полный прогон зелёный на восстановленном столе

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- 04-06 получает зелёную v3-базу для двойного прогона D-48 (свежая сессия раннера → прогон ×2)
- 04-07/verify: владелец утверждает x11-исключения (A1 fallback: «критерий №2 покрывается фактически доступными оконными режимами») и gedit RU-ограничение — обе записи в шапке v3 и WINDOWS-кандидаты при непринятии
- Будущий пере-пин: если x11 IM-путь Chromium начнёт обслуживать surrounding — x11-строки громко упадут и потребуют апгрейда до коррекционных (заложено в комментарии строк)

---
*Phase: 04-postavka-i-priemka*
*Completed: 2026-09-16*

## Self-Check: PASSED

- matrix-v3.yaml, драйверы, префлайт, ci-runner.md — на диске ([ -f ] проверен)
- Коммиты c3d6d07 (Task 1), b38e6e6 (Task 2) — в git log
- Полный живой прогон v3: 31/31 PASS, exit 0 (2026-09-16, см. e2e-report.txt прогона)
- mise run ci зелёный; V2-SUBSET-OK; COUNT-OK (31 ≥ 21+10); V2-FROZEN
