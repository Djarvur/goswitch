---
phase: 04-postavka-i-priemka
plan: 06
subsystem: testing
tags: [ci, github-actions, e2e, matrix-v3, d48, mise, loginctl, self-hosted-runner]

# Dependency graph
requires:
  - phase: 04-postavka-i-priemka
    provides: "04-05: matrix-v3.yaml (31 кейса, живой прогон 31/31) — база двойного прогона; 04-03: прецедент разделения contents: write (release.yml alone); 02-07/03-07: bootstrap-модель диспатча и живое доказательство «определение исполняется от ref»"
provides:
  - "D-48 операционализован как CI-механика: workflow e2e-matrix гоняет v3 ДВАЖДЫ ПОДРЯД в одном job без вмешательства между прогонами; живое доказательство — run 35108412175, оба прогона 31/31 PASS, conclusion success"
  - "вход fresh_session (bool, default false) + loginctl-префлайт свежести сессии (порог 30 минут — константа шага D48_MAX_SESSION_AGE_SEC, превышение = падение с подсказкой relogin)"
  - "mise-задача e2e-matrix-v3 — единый интерфейс локально и в CI (D-09)"
  - "docs/ci-runner.md: раздел «D-48 double-run gate» — процедура владельца (перелогин → dispatch fresh_session=true → оба прогона зелёные)"
  - "починен регресс матричного вочдога (04-04, 1b443eb): нулевой лимит снова означает дефолт 180s — матрицы v1/v2/v3 снова исполняемы везде"
affects: [04-07 (verify-гейт фазы: формальный свежесессионный прогон D-48 по процедуре из ci-runner.md), будущие релизные циклы (двойной прогон — часть приёмки)]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
actuals:
  tokens: 4764     # chars/4 over the realized diff (19056 chars, 6 files)
  tasks: 2
  commits: 3       # MEASURED: git rev-list --count plan_head_before(34c95b4)..HEAD

# Tech tracking
tech-stack:
  added: []        # ноль новых зависимостей: loginctl/GNU date — системные утилиты раннера
  patterns:
    - "гейт-инпут workflow: bool-вход + именованная константа порога в env шага + падение с подсказкой фикса (fail-fast preflight паттерн стенда, перенесённый в CI)"
    - "нулевая длительность = «дефолт, пожалуйста» — резолюция внутри общего механизма (watchdogLimit), не на_call-сайтах (класс ловушки 04-04 закрыт структурно)"

key-files:
  created:
    - test/e2e/watchdog_test.go
  modified:
    - .github/workflows/e2e-matrix.yml
    - mise.toml
    - docs/ci-runner.md
    - test/e2e/main.go
    - test/e2e/matrix.go

key-decisions:
  - "Механика vs гейт зафиксированы раздельно (assumption плана): план доказал двойной прогон на ТЕКУЩЕЙ сессии (fresh_session=false, run 35108412175); формальный свежесессионный прогон (fresh_session=true после перелогина владельца) — шаг приёмки фазы 04-07/verify-work по процедуре из docs/ci-runner.md"
  - "loginctl-факты (живая находка): show-session -p Since — НЕ свойство (пусто), Timestamp --value отдаёт ЧЕЛОВЕЧЕСКИЙ формат (не usec) — префлайт парсит GNU date -d, неразбираемый штамп падает громко с подсказкой relogin"
  - "Регресс вочдога чинен структурно: резолюция 0→caseTimeout перенесена внутрь runCaseWatchdog (watchdogLimit) — ни один вызывный сайт больше не может создать нулевой бюджет; дублирующая подстановка в -case-пути удалена"
  - "Bootstrap-PR не потребовался: workflow уже зарегистрирован на main (d815e72), а диспатч исполняет ОПРЕДЕЛЕНИЕ от ref (живое доказательство 03-07) — достаточно запушить фаза-ветку и диспатчить --ref"

patterns-established:
  - "D-48-дисциплина двойного прогона: между прогонами НЕТ шагов очистки/пересоздания — любой шаг между ними = подмена проверки (запинено комментариями шагов workflow и пином в этом SUMMARY)"

requirements-completed: [INST-01]  # разделяемый ID (04-01..04-07): mark-complete через ready-ids гейт — блокирован 04-07 (ещё без SUMMARY), отложено до его закрытия

coverage:
  - id: D1
    description: "mise-задача e2e-matrix-v3 — единый интерфейс двойного прогона локально и в CI (D-09)"
    requirement: INST-01
    verification:
      - kind: e2e
        ref: "CI-шаг mise run e2e-matrix-v3 (run 35108412175) — matrix: 31/31 PASS, дважды"
        status: pass
      - kind: unit
        ref: "греп-гейт DOUBLE-RUN-WIRED (задача в mise.toml + ровно 2 шага в workflow)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Workflow e2e-matrix: дефолт v3, ДВА последовательных шага прогона без вмешательства между ними, MATRIX через env (WR-04), concurrency/permissions сохранены"
    requirement: INST-01
    verification:
      - kind: e2e
        ref: "gh run view 35108412175 — conclusion success; оба шага 'matrix v3 run' success; NO-FAIL-ROWS (0 FAIL-строк в полном логе)"
        status: pass
      - kind: unit
        ref: "структурная сверка YAML (ruby yaml parse): шаги смежные, \${{ }} отсутствует в run-блоках, fresh_session default false"
        status: pass
    human_judgment: false
  - id: D3
    description: "Свежесть сессии проверяема: вход fresh_session + loginctl-префлайт (порог 30 минут, именованная константа, подсказка relogin)"
    requirement: INST-01
    verification:
      - kind: unit
        ref: "логика парсинга Timestamp прожита локально на реальном выводе loginctl этой машины (= green106): sid=2, GNU date -d парсит, возраст 2 дня > порога → WOULD-FAIL корректен"
        status: pass
    human_judgment: true
    rationale: "Путь fresh_session=true ВНУТРИ workflow не исполнялся живьём — он требует перелогина владельца; это и есть формальный D-48-гейт, зарезервированный за приёмкой фазы (04-07/verify-work) по процедуре из docs/ci-runner.md. Разделение механика-vs-гейт — осознанное решение плана, не пробел"
  - id: D4
    description: "Процедура D-48-гейта владельца документирована в docs/ci-runner.md (перелогин → dispatch fresh_session=true → критерий: оба прогона зелёные в одном запуске)"
    requirement: INST-01
    verification:
      - kind: unit
        ref: "греп-гейт RUNNER-DOC-OK (relogin + fresh_session в книге раннера); трёхшаговая процедура с критерием успеха и примечанием о loginctl-префлайте"
        status: pass
    human_judgment: false
  - id: D5
    description: "Живое доказательство механики двойного прогона: один dispatch на green106, оба последовательных v3-прогона зелёные"
    requirement: INST-01
    verification:
      - kind: e2e
        ref: "run 35108412175: run#1 'matrix: 31/31 PASS' (14:27:45Z) → сразу run#2 'matrix: 31/31 PASS' (14:29:57Z), conclusion success, 0 FAIL-строк"
        status: pass
    human_judgment: false
  - id: D6
    description: "Регресс матричного вочдога (04-04) найден и починен: нулевой лимит = дефолт caseTimeout, матрицы снова исполняемы"
    verification:
      - kind: unit
        ref: "test/e2e/watchdog_test.go TestWatchdogLimit_ZeroMeansDefault/_ExplicitOverrides (RED→GREEN); mise run ci зелёный"
        status: pass
      - kind: e2e
        ref: "живой прогон v2 локально (PASS-строки случаев восстановились) + двойной прогон v3 в CI (D5)"
        status: pass
    human_judgment: false

# Metrics
duration: 22 min
completed: 2026-09-16
status: complete
---

# Phase 04 Plan 06: D-48 двойной прогон матрицы v3 Summary

**D-48 операционализован и доказан живьём: один dispatch на green106 гоняет полную матрицу v3 ДВАЖДЫ ПОДРЯД в одном job — 31/31 PASS и снова 31/31 PASS без единого шага между прогонами; свежесть сессии проверяема через fresh_session + loginctl-префлайт, процедура владельца — в книге раннера; по пути найден и починен спящий регресс матричного вочдога от 04-04**

## Performance

- **Duration:** 22 min (14:10–14:32 UTC; два фоновых CI-прогона ~19 мин суммарно)
- **Started:** 2026-09-16T14:10:29Z
- **Completed:** 2026-09-16T14:32:19Z (фиксация метаданных — позже)
- **Tasks:** 2
- **Files modified:** 6 (196 insertions, 25 deletions)

## Accomplishments

- Workflow `e2e-matrix.yml`: дефолт matrix-пути — v3; вход `fresh_session` (bool, default false, «D-48 gate: set true only after owner relogin»); префлайт свежести сессии через loginctl с порогом 30 минут (именованная константа `D48_MAX_SESSION_AGE_SEC`, превышение = падение с «owner relogin required (D-48)»); ДВА последовательных шага «matrix v3 run #1/#2» — между ними ТОЛЬКО комментарии, никаких шагов очистки/пересоздания (суть D-48 запинена в комментариях шагов); MATRIX через env, ноль shell-интерполяции (WR-04); contents: read и concurrency e2e-gnome без cancel-in-progress сохранены (T-04-06-02/03)
- Живое доказательство механики (Task 2): dispatch `--ref gsd/phase-04-postavka-i-priemka`, run **35108412175** — run #1 `matrix: 31/31 PASS` (14:27:45Z), сразу run #2 `matrix: 31/31 PASS` (14:29:57Z), conclusion **success**, ноль FAIL-строк в полном логе. Счётчики прогонов равны (31 = 31 — superset v3)
- mise-задача `e2e-matrix-v3` по конвенции описаний (план-реф 04-06, трейлер «live GNOME session; NOT in ci») — локальный и CI-прогон зовут одну задачу (D-09)
- docs/ci-runner.md: раздел «D-48 double-run gate» — процедура владельца из трёх шагов (перелогин GDM → раннер возвращается сам через graphical-session.target → dispatch `fresh_session=true` → критерий: оба прогона зелёные в одном запуске) + примечание о loginctl-префлайте; туда же добавлено уточнение 03-07 об исполнении определения от ref
- Находка D-48-качества до того, как она стала находкой прогона №2: регресс вочдога (см. Deviations) убит, матричный контур снова здравствует

## Task Commits

1. **Task 1: mise-задача v3 + двойной прогон workflow + fresh_session-префлайт + процедура в книге раннера** - `fbd3072` (feat)
2. **Task 2: живое доказательство механики — dispatch на green106** - прогоны 35107406144 (failure — нашёл регресс) и **35108412175 (success)**; правки по находкам: `2526a8a` (fix, префлайт), `9f2fd75` (fix, вочдог)

_План type: execute (задачи type=auto) — плановые TDD RED/GREEN гейты не применялись; фикс вочдога volée-RED→GREEN (watchdog_test.go) — добровольная дисциплина поверх deviation-правил._

## Files Created/Modified

- `.github/workflows/e2e-matrix.yml` — v3-дефолт, двойной прогон, fresh_session + loginctl-префлайт, шапка с цитатой D-48
- `mise.toml` — `[tasks.e2e-matrix-v3]`
- `docs/ci-runner.md` — раздел «D-48 double-run gate» + уточнение про ref-определение
- `test/e2e/main.go` — `watchdogLimit` (0 = дефолт caseTimeout), вызов из `runCaseWatchdog`, упрощённый -case-путь
- `test/e2e/matrix.go` — комментарий вызывного сайта (0 = дефолт, резолюция внутри)
- `test/e2e/watchdog_test.go` — NEW: корпус резолюции лимита (RED→GREEN)

## Decisions Made

- Механика-vs-гейт: план доказал двойной прогон на текущей сессии; формальный свежесессионный гейт — за владельцем на приёмке фазы (04-07/verify-work), процедура готова в книге раннера
- Парсинг времени входа: `loginctl show-session -p Timestamp --value` + GNU `date -d` (живые факты о loginctl — см. Deviations 1); неразбираемый штамп = громкое падение с подсказкой relogin
- Регресс вочдога закрыт структурно (резолюция внутри `runCaseWatchdog`), а не точечно на вызывном сайте — класс ловушки «нулевой бюджет по недосмотру» невозможен отныне ни для одного пути
- Диспатч без bootstrap-PR: определение воркфлоуса исполняется от диспатч-ветки (03-07), регистрация на main уже есть (d815e72) — прецедент 02-07 остаётся запасным на случай deregistration

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] loginctl-префлайт из плана не работает как написан**
- **Found during:** Task 2 (сухая проверка логики префлайта на этой машине до повторного диспатча)
- **Issue:** план предполагал usec-арифметику; живой пробой: `show-session -p Since --value` — пусто (Since — колонка таблицы, не свойство), `Timestamp --value` — человеческий формат «Mon 2026-09-14 21:54:47 MSK»; арифметика умирала на bash syntax error — префлайт падал бы мусорной ошибкой на формальном гейте
- **Fix:** парсинг GNU `date -d`; проверка разбираемости с громким падением и подсказкой relogin
- **Files modified:** .github/workflows/e2e-matrix.yml
- **Verification:** логика прожита на реальном выводе loginctl (sid=2, парс ок, возраст 2 дня → WOULD-FAIL корректен)
- **Committed in:** 2526a8a

**2. [Rule 1 - Bug] Регресс матричного вочдога от 04-04: каждый кейс каждой матрицы падал мгновенно**
- **Found during:** Task 2 (первый dispatch, run 35107406144: все 31 кейс «watchdog: did not complete within 0s», conclusion failure) — воспроизведён локально на v2
- **Issue:** 04-04 (1b443eb) добавил параметр `limit` в `runCaseWatchdog` и механически передал `0` из матричного пути — `time.After(0)` срабатывает мгновенно; матрицы v1/v2/v3 были неисполнимы с этого коммита (04-05 её зелёный прогон предшествовал регрессу; 04-03/04-04 матрицу не гоняли)
- **Fix:** `watchdogLimit` резолвит 0 → caseTimeout ВНУТРИ `runCaseWatchdog` (конвенция caseSpec.watchdog задокументирована); дублирующая подстановка в -case-пути удалена; RED→GREEN тесты
- **Files modified:** test/e2e/main.go, test/e2e/matrix.go, test/e2e/watchdog_test.go
- **Verification:** локальный прогон v2 (случаи снова PASS) + повторный dispatch: оба прогона v3 31/31 (run 35108412175); mise run ci зелёный
- **Committed in:** 9f2fd75

**3. [Rule 3 - Blocking] Verify-команда плана использует флаг, которого нет в установленном gh**
- **Found during:** Task 2 (гейт «gh run view --workflow e2e-matrix --log» → «unknown flag: --workflow»)
- **Issue:** установленная версия gh CLI не принимает `--workflow` у `gh run view`
- **Fix:** тот же гейт через явный id: `gh run list --workflow e2e-matrix --limit 1 --json databaseId` → `gh run view <id> --log`; семантика гейта (лог последнего прогона, ноль FAIL-строк, пустой лог = падение) не тронута
- **Files modified:** нет (адаптация команды, не кода)
- **Verification:** NO-FAIL-ROWS на логе run 35108412175
- **Committed in:** — (протокольная адаптация)

---

**Total deviations:** 3 auto-fixed (1 bug с живым воспроизведением, 2 blocking-формы). **Impact on plan:** находка №2 — именно то, ради чего двойной прогон существует: спящий регресс, невидимый юнит-гейтам, пойман первым же живым прогоном. Без неё формальный гейт приёмки упал бы на ровном месте.

## Issues Encountered

- Первый dispatch (35107406144) упал целиком из-за регресса №2 — это «находка D-48-качества» в терминах плана, но не про «первый прогон портит состояние второму»: падали ОБА будущих прогона одинаково (вочдог), а не второй после первого. Чинено по правилу плана: не ослаблять второй прогон, чинить стенд и повторять dispatch
- Свежесессионный формальный прогон НЕ исполнялся (требует перелогина владельца) — сознательно оставлен шагом приёмки фазы; механика префлайта валидирована локально на реальном loginctl-выводе

## Known Stubs

None — ни заглушек, ни незакрытых <verify>: все гейты исполнены (DOUBLE-RUN-WIRED, RUNNER-DOC-OK, conclusion success, NO-FAIL-ROWS, mise run ci).

## User Setup Required

None - no external service configuration required. (Формальный D-48-гейт требует ПЕРЕЛОГИНА владельца в момент приёмки фазы — это шаг процедуры из docs/ci-runner.md, не настройка.)

## Next Phase Readiness

- 04-07 (последний план фазы) получает: живую механику D-48, готовую процедуру свежесессионного прогона, здравствующий матричный контур (регресс закрыт) — формальный гейт: перелогин → dispatch fresh_session=true → оба прогона зелёные
- INST-01 остаётся открытой до 04-07 (разделяемый ID: ready-ids гейт блокирует mark-complete, пока 04-07 без SUMMARY) — ожидаемое поведение, не блокер
- WINDOWS.md: регресс вочдога записан в реестр как deviation со статусом fixed (кросс-плановый след 04-04→04-06)

---
*Phase: 04-postavka-i-priemka*
*Completed: 2026-09-16*

## Self-Check: PASSED

- Все 6 файлов плана и SUMMARY на диске ([ -f ] проверен)
- Коммиты fbd3072 (Task 1), 2526a8a + 9f2fd75 (фиксы Task 2), 894e3f0 (SUMMARY) — в git log
- commits: 3 MEASURED от plan_head_before 34c95b4 (совпадает с frontmatter)
- run 35108412175: conclusion success, оба прогона «matrix: 31/31 PASS», 0 FAIL-строк
- mise run ci зелёный на HEAD 9f2fd75
