---
phase: 02-korrektsiya-slova-en-ru
plan: "02"
subsystem: testing
tags: [e2e, at-spi, chromium, gnome-text-editor, gtk4, ydotool, live-session]

# Dependency graph
requires:
  - phase: 01-ibus-engine-adresatsiya-i-e2e-stend
    provides: e2e-стенд (main.go registry/watchdog/teardown, focus_helper.py AT-SPI мост, префлайт-таблица)
provides:
  - Драйверы поверхностей chromium (startChromium/waitChromiumInput/readChromiumText/closeChromium) и gnome-text-editor (startGTE/waitGTEInput/readGTEText/closeGTE) в test/e2e/surface.go
  - Абстракция surfaceDriver + runSurfaceSmoke (case_surface.go) — скелет для кейсов 02-05/02-06
  - Префлайт-проверки chromium-launch и gnome-text-editor-launch
  - Живые smoke-кейсы реестра chromium-smoke и gte-smoke + mise-задачи e2e-chromium-smoke/e2e-gte-smoke
  - focus_helper.py: focused-text/focused-inputs (мост из Task 1) + pid-ключевые focused-input-pid/focused-text-pid/grab-input-pid и GTK4-совместимый read_text (GetStringAtOffset)
affects: [02-korrektsiya-slova-en-ru (02-05 лестница, 02-06 матрица), 04-postavka-i-priyomka (e2e CI)]

# Actuals (#2632) — same estimateTokens scale (chars/4 over the realized diff)
actuals:
  tokens: 8900
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []  # нулевые новые зависимости — только stdlib + существующий python3-gi мост
  patterns:
    - "surfaceDriver (open/await/read/close) — поверхность как непрозрачная ценность для smoke и матрицы"
    - "pid-ключевой AT-SPI witness/readback — экземплярная точность при разделяемом имени приложения (Google Chrome, gnome-text-editor)"
    - "grabFocus-реактивация: отказанный AT-SPI вызов всё равно триггерит свежий activation request GTK4 — детерминированное лечение focus-stealing denial"
    - "XDG_DATA_HOME-изоляция GTK-приложения: сессия/черновики владельца нетронуты (session.gvariant, drafts/)"
    - "focused-inputs witness-гейт: инжекция запрещена, пока оболочечный entry держит фокус (AT-SPI focus lag)"

key-files:
  created:
    - test/e2e/surface.go
    - test/e2e/fixtures/input.html
    - test/e2e/case_surface.go
  modified:
    - test/e2e/main.go
    - test/e2e/preflight.go
    - test/e2e/focus_helper.py
    - test/e2e/case_m1.go
    - test/e2e/case_d01.go
    - test/e2e/case_resilience.go
    - test/e2e/README.md
    - mise.toml
    - .gitignore

key-decisions:
  - "Chromium-бинарник закреплён константой google-chrome (deb 153.x, AT-SPI имя \"Google Chrome\" — live-verified); смена на snap chromium — правка двух констант в surface.go"
  - "Драйвер GTE не следует букве плана «spawn без аргументов»: gnome-text-editor single-instance и реставрирует сессию владельца даже с --ignore-session (46.3); --standalone + XDG_DATA_HOME в temp-каталог стенда — единственный способ получить собственный процесс, чистый документ и нетронутый draft-store владельца"
  - "Pid-ключевой witness вместо app-name: оба имени поверхностей разделяются с инстансами владельца; chars-гейт оставлен как второй барьер"
  - "grabFocus-реактивация: GTK4 отвечает на input-node grabFocus ошибкой и всё равно выполняет widget-grab как свежий activation request — PASS доказан под непрерывным wiggl'ом указателя (симуляция активности владельца)"
  - "read_text хелпера: сначала современный GetStringAtOffset (TextGranularity.LINE) — GTK4-мост отказывает на deprecated GetTextAtOffset, Chromium обслуживает оба"

patterns-established:
  - "Surface driver: 4 операции (open/await/read/close) + surfaceKind enum, расширяемый из surface.go без правки case_m1.go"
  - "Живой smoke как доказательство драйвера БЕЗ коррекции: inject ghbdtn → readback ghbdtn, коррекция подключается позже поверх готовых драйверов"

requirements-completed: [TEST-04]  # копия из frontmatter плана; marking в REQUIREMENTS.md отложен shared-ID гейтом (02-06 тоже декларирует TEST-04)

coverage:
  - id: D1
    description: "Драйвер chromium: свежий инстанс (temp --user-data-dir), witness-гейт, focused readback, PID-close; профиль владельца не тронут"
    requirement: TEST-04
    verification:
      - kind: e2e
        ref: "mise run e2e-chromium-smoke → PASS chromium-smoke (live, 2026-09-14, повторно после Task 2)"
        status: pass
      - kind: other
        ref: "grep 'user-data-dir' test/e2e/surface.go && grep 'chromium' test/e2e/preflight.go → SURFACE-OK"
        status: pass
    human_judgment: false
  - id: D2
    description: "Драйвер gnome-text-editor: standalone-инстанс с изолированным XDG_DATA_HOME, pid-keyed witness + grabFocus-реактивация, readback, SIGKILL-close; сессия и черновики владельца нетронуты"
    requirement: TEST-04
    verification:
      - kind: e2e
        ref: "mise run e2e-gte-smoke → PASS gte-smoke (live, 3 прогона подряд + 1 под непрерывной pointer-активностью)"
        status: pass
      - kind: other
        ref: "grep 'gnome-text-editor' test/e2e/preflight.go && grep 'gte-smoke' main.go case_surface.go → GTE-OK"
        status: pass
    human_judgment: false
  - id: D3
    description: "Префлайт-проверки обеих поверхностей: spawn → witness → kill, по одной человеко-читаемой строке диагностики на каждую"
    verification:
      - kind: e2e
        ref: "preflight chromium-launch: ok / gnome-text-editor-launch: ok в каждом живом прогоне"
        status: pass
    human_judgment: false
  - id: D4
    description: "Тройной гейт (D-08) зелёный поверх изменений стенда: mise run ci"
    verification:
      - kind: other
        ref: "mise run ci (build+vet+lint+test -race) → 0 issues, exit 0"
        status: pass
    human_judgment: false

# Metrics
duration: 54min
completed: 2026-09-14
status: complete
---

# Phase 2 Plan 02: Поверхности e2e-стенда (chromium + gnome-text-editor) Summary

**Два драйвера живых поверхностей с witness-гейтами и pid-точным AT-SPI readback; GTE-драйвер победил mutter focus-stealing denial через grabFocus-реактивацию — доказано PASS'ом под непрерывной активностью указателя**

## Performance

- **Duration:** 54 min (плюс checkpoint-пауза на заблокированный экран до континуации)
- **Started:** 2026-09-14T18:58:54Z (континуация после human-action checkpoint)
- **Completed:** 2026-09-14T19:53:24Z
- **Tasks:** 2 / 2
- **Files modified:** 12

## Accomplishments
- Драйвер chromium по канону плана: свежий инстанс `--user-data-dir` в temp, `--force-renderer-accessibility` (без него AT-SPI-дерево свежего инстанса пусто — live-verified), focused readback, PID-close; smoke PASS с первого живого прогона
- Драйвер gnome-text-editor, переосмысленный по живым находкам: single-instance GTK-приложение реставрирует сессию владельца даже с `--ignore-session` — изоляция `--standalone` + `XDG_DATA_HOME` в temp; PID-keyed witness; grabFocus-реактивация фокуса
- Слой фокуса стенда усилен: `focused-inputs`-гейт отказывается инжектировать, пока оболочечный entry держит фокус (AT-SPI focus lag), pid-ключевые команды хелпера дают экземплярную точность при разделяемых именах приложений
- Оба smoke доказаны многократно, включая прогон gte-smoke под непрерывным wiggl'ом указателя (симуляция активности владельца за столом)

## Task Commits

Each task was committed atomically:

1. **Task 1: Драйвер chromium** — `412dff8` (feat)
2. **Task 2: Драйвер gnome-text-editor** — `078a23e` (feat)

**Plan metadata:** см. финальный docs-коммит ниже.

## Files Created/Modified
- `test/e2e/surface.go` — оба драйвера поверхностей, enum расширения surfaceKind, waitInputFocus/readFocusedText мосты, withEnv
- `test/e2e/fixtures/input.html` — статическая страница `<input id="e">` для Chromium-кейсов
- `test/e2e/case_surface.go` — surfaceDriver абстракция, runSurfaceSmoke, runChromiumSmoke, runGTESmoke
- `test/e2e/preflight.go` — проверки chromium-launch и gnome-text-editor-launch
- `test/e2e/main.go` — stand.chromium/gte, реестр chromium-smoke/gte-smoke, teardown closeChromium/closeGTE
- `test/e2e/focus_helper.py` — focused-text/focused-inputs/focused-text-pid/focused-input-pid/grab-input-pid; read_text через GetStringAtOffset
- `test/e2e/case_m1.go`, `case_d01.go`, `case_resilience.go` — exhaustive-ветки enum (lint D-10)
- `test/e2e/README.md`, `mise.toml` — задачи e2e-chromium-smoke / e2e-gte-smoke, таблица флагов
- `.gitignore` — `__pycache__/` (продукт запусков хелпера)

## Decisions Made
- См. key-decisions в frontmatter — все решения продиктованы живыми находками на столе владельца и зафиксированы с датами верификации

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 — Bug] readback GTK4-поверхностей: мост хелпера не читал GtkSourceView**
- **Found during:** Task 2 (живая отладка драйвера GTE)
- **Issue:** GTK4 AT-SPI-мост отвечает на deprecated `get_text_at_offset` ошибкой «deprecated in favor of GetStringAtOffset» (Chromium обслуживает оба) — readback пуст
- **Fix:** `read_text` сначала зовёт современный `get_string_at_offset(0, TextGranularity.LINE)` (живьём подтверждено: content='ghbdtn'), deprecated-путь и get_text остались фолбэками
- **Files modified:** test/e2e/focus_helper.py
- **Verification:** PASS gte-smoke (readback == ghbdtn)
- **Committed in:** 078a23e

**2. [Rule 1 — Bug] mutter focus-stealing denial: окно GTE детерминированно не получает фокус при активности пользователя**
- **Found during:** Task 2 (chromium-smoke упал на gte-префлайте; матрица A: 5/6 отказов при spawn сразу после kill-окна; контролируемый тест: pointer-событие ПОСЛЕ map → отказ 4/4 на обоих GDK-бэкендах)
- **Issue:** без activation token новому окну GTK4 отказывается в фокусе всякий раз, когда за его map-запросом следует реальный input-событие (даже движение указателя) — драйвер плана («spawn, окно получит фокус») недетерминирован
- **Fix:** waitGTEInput — poll + периодический pid-ключевой `grab-input-pid`: отказанный AT-SPI grabFocus на input-node всё равно триггерит свежий activation request GTK4, чей timestamp новее отказа
- **Files modified:** test/e2e/surface.go, test/e2e/focus_helper.py
- **Verification:** PASS gte-smoke под непрерывным wiggl'ом указателя весь прогон; 3/3 обычных PASS; chromium-smoke PASS
- **Committed in:** 078a23e

**3. [Rule 2 — Missing Critical] spawn GTE «без аргументов» реставрирует сессию владельца и пишет черновики в его draft-store**
- **Found during:** Task 2 (живое зондирование)
- **Issue:** gnome-text-editor single-instance (spawn делегирует в процесс владельца — PID-close невозможен) и 46.3-сборка восстанавливает session.gvariant даже с `--ignore-session`; autosave черновиков (3 с) записал бы probe-слово в ~/.local/share/org.gnome.TextEditor/drafts/; была и одна аномалия «fresh doc chars=22» до изоляции
- **Fix:** `--standalone --ignore-session` + `XDG_DATA_HOME=<tmp>/gte-data` (withEnv): свой процесс, чистый документ, черновики и сессия умирают с temp-каталогом стенда
- **Files modified:** test/e2e/surface.go
- **Verification:** после каждого прогона drafts/ владельца нетронут (только их старый файл 10 Sep), temp-drafts удаляются teardown
- **Committed in:** 078a23e

**4. [Rule 3 — Blocking] exhaustive-линтер потребовал ветки surfaceGnomeTextEditor в трёх существующих switch'ах**
- **Found during:** Task 2 (первый прогон mise run lint)
- **Issue:** файлы case_m1/d01/resilience не входили в <files> задачи, но строгий конфиг (`linters.default: all`) запрещает неполные switch по surfaceKind
- **Fix:** по одной комментарий-ветке на файл в стиле Task 1 (surfaceChromium)
- **Files modified:** test/e2e/case_m1.go, test/e2e/case_d01.go, test/e2e/case_resilience.go
- **Verification:** golangci-lint 0 issues
- **Committed in:** 078a23e

**5. [Docs] README.md и .gitignore вне файловых списков задач**
- **Found during:** Task 1/2
- **Issue:** новые mise-задачи не отражены в README (конвенция Task 1), `test/e2e/__pycache__/` остаётся untracked после запусков хелпера
- **Fix:** строка в таблице флагов + задача в списке прогона; `__pycache__/` в .gitignore
- **Files modified:** test/e2e/README.md, .gitignore
- **Verification:** README консистентен реестру; git status чист от runtime-мусора
- **Committed in:** 412dff8 (README-строка chromium), 078a23e (gte), 412dff8 (.gitignore)

---

**Total deviations:** 5 auto-fixed (2 bug, 1 missing critical, 1 blocking, 1 docs)
**Impact on plan:** Все — прямые следствия живого стола; планная семантика (ввод→readback без коррекции, запрет чужих поверхностей, нетронутый профиль владельца) соблюдена строже буквы.

## Issues Encountered
- Живой chromium-smoke прошёл с первого прогона; GTE потребовал трассировки через 7 живых зондов (single-instance, session.gvariant, drafts, focus-stealing матрицы A/B/C, GDK_BACKEND=x11 — не помог, grabFocus — помог). Все находки — в key-decisions и комментариях кода с датами.
- Континуация стартовала после human-action checkpoint (экран был заблокирован) — задокументирован как нормальный поток, не отклонение.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Драйверы готовы к 02-05 (лестница) и 02-06 (YAML-матрица) без изменений: surfaceDriver-абстракция уже параметризуется поверхностью
- Для кейсов с коррекцией: readback остаётся focused/pid-ключевым; grab-реактивация включена только в GTE-драйвер (chromium форсирует активацию сам) — если матрица встретит отказ фокуса на chromium, паттерн grab-input-pid переносится одним вызовом
- REQUIREMENTS.md: TEST-04 не отмечен — shared-ID гейт (02-06 тоже декларирует; отметится при закрытии 02-06)

## TDD Gate Compliance
План type: execute, задачи type="auto" без tdd="true"; предикат task.is-behavior-adding → false (проверено gsd-tools). Гейты RED/GREEN не применялись: единственный честный оракул этих задач — живая сессия (класс задачи заявлен планом: smoke-доказательство драйверов, D-07/D-09), тройной гейт D-08 зелёный на каждом шаге.

## Self-Check: PASSED

- Files: surface.go, fixtures/input.html, case_surface.go, 02-02-SUMMARY.md — FOUND
- Commits: 412dff8 (Task 1), 078a23e (Task 2) — FOUND
- Plan verification re-run: mise run e2e-chromium-smoke PASS, mise run e2e-gte-smoke PASS (3× + 1 под pointer-вигглом), mise run ci exit 0

---
*Phase: 02-korrektsiya-slova-en-ru*
*Completed: 2026-09-14*
