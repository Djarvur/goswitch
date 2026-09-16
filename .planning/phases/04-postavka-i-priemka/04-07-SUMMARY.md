---
phase: 04-postavka-i-priemka
plan: 07
subsystem: docs-release
tags: [readme, bilingual, acceptance-uat, windows-ledger, perf-budget, release-gate, inst-01]

# Dependency graph
requires:
  - phase: 04-postavka-i-priemka
    provides: "04-01: goswitchctl install/uninstall facts; 04-02: selfcheck six steps; 04-03: release channels (.goreleaser.yaml + release.yml); 04-04: perf-report.txt live numbers + the latency escalation; 04-05: matrix v3; 04-06: D-48 double-run mechanics + ci-runner owner procedure"
provides:
  - "Двуязычная пара README.md (EN, первичен) + README.ru.md (RU): установка одной командой goswitchctl install (канал релиз-архива + go install), selfcheck, жесты/CLI, конфиг (дефолты до первого YAML), perf-таблица из живого прогона, расширенное -debug-предупреждение, troubleshooting, uninstall --purge"
  - "docs/ACCEPTANCE.md — письменный чек-лист приёмки владельца D-49 (формат 03-UAT): 6 пунктов expected/result/note (формальный D-48 свежесессионный прогон, установка с нуля по README, живые жесты, hot-reload конфига, счётчики status, uninstall-восстановление) + Summary-блок"
  - "WINDOWS-ledger сведён глагонами: #2 → fixed (решение владельца 2026-09-15), #4 → waived с причиной (GTK4-Wayland platform limitation, UAT 2026-09-15); open_count честно = 1 (пункт #5 — решение владельца по методике латентности)"
  - "Третий воспроизводимый вердикт perf-гейта (p95 184.4 мс ≥ 50 мс) на HEAD 8ac3451 — полная доказательная база чекпоинта владельцу"
affects: [решение владельца по WINDOWS #5 (методика INST-03), тег v1.0.0 (заблокирован), verify-work фазы (UAT по docs/ACCEPTANCE.md)]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
actuals:
  tokens: 7647    # chars/4 over the realized diff (30586 chars, 4 files)
  tasks: 2        # of 3 — Task 3 stopped at the red perf gate per the plan's explicit rule (no tag)
  commits: 3      # MEASURED: git rev-list --count gsd-plan-head-before-04-07(657fbc4)..HEAD

# Tech tracking
tech-stack:
  added: []       # только доки и ledger-глаголы; ноль зависимостей
  patterns:
    - "README-таблица D-46: числа из последнего perf-report.txt + строка методики (окно ydotool→AT-SPI, оверхед инжектора назван явно) + примечание об обновлении каждым релизом"
    - "ledger-сверка только глаголами gsd-tools (fixed/waive с причиной из решений владельца) — никаких ручных правок статусов"

key-files:
  created:
    - README.ru.md
    - docs/ACCEPTANCE.md
  modified:
    - README.md
    - .planning/WINDOWS.md

key-decisions:
  - "Perf-таблица опубликована с ЧЕСТНЫМИ живыми числами и методикой (окно включает ~170 мс инжекции стенда; реакция демона < 2 мс): репо-README не публичен до тега, а тег заблокирован гейтом — при смене методики владельцем таблица обновляется до тега (D-46)"
  - "WINDOWS #4 waive выполнен глаголом с дословной причиной из решения владельца 2026-09-15; #5 НЕ тронут — исполнитель не вправе решать вопрос методики владельца (Rule 4-класс), open_count 1 — честное состояние"
  - "docs/ACCEPTANCE.md включает формальный D-48 свежесессионный прогон пунктом 1 (ДО локальной установки — прогон матрицы сам управляет источниками через snapshot/restore; процедура docs/ci-runner.md от 04-06)"
  - "Task 3 ОСТАНОВЛЕН на красном perf-гейте по явному правилу плана (любой красный = СТОП, тег не ставить) — релиз v1.0.0 НЕ публиковался; запретительных проверок (STAMPED/RELEASE-ASSETS/PROXY) существовать не должно"

patterns-established:
  - "Чек-лист приёмки фазы живёт в docs/ACCEPTANCE.md (discretion плана), формат 03-UAT, порядок: свежесессионный D-48 → install → жесты → конфиг → status → uninstall"

requirements-completed: []  # INST-01 (shared ID 04-01..04-07) НЕ закрывается этим планом: канал GitHub Releases пуст — тег v1.0.0 заблокирован красным perf-гейтом (WINDOWS #5); mark-complete будет исполнен после публикации релиза

coverage:
  - id: D1
    description: "Двуязычная пара README: установка одной командой (оба канала), selfcheck/uninstall, perf-таблица из живого прогона, приватность -debug, troubleshooting; команды/числа синхронны; голого write-cache нет"
    requirement: INST-01
    verification:
      - kind: other
        ref: "grep-гейт README-COMPLETE (goswitchctl install/selfcheck/uninstall/p95 в обоих языках) — pass"
        status: pass
      - kind: other
        ref: "grep-гейт NO-BARE-CACHE (ни одного 'ibus write-cache' вне #/> строк в обоих файлах) — pass"
        status: pass
      - kind: other
        ref: "grep-гейт синхрона: счётчик 'goswitchctl install' en=4 ru=4 — pass"
        status: pass
      - kind: other
        ref: "числа таблицы == perf-report.txt (прогон 2026-09-16T14:58:14Z: p50 166.9 / p95 184.4 / p99 186.5 мс, VmHWM 12368 кБ = 12.4 МБ)"
        status: pass
    human_judgment: true
    rationale: "Фактическая достаточность и полная синхронность пары (D-50) — flagged verification:judgment плана: грей-зоны формулировок и полнота для читателя проверяются владельцем/верификатором на UAT"
  - id: D2
    description: "docs/ACCEPTANCE.md — чек-лист приёмки владельца D-49 в формате 03-UAT (6 пунктов с expected/result/note + Summary-блок)"
    requirement: INST-01
    verification:
      - kind: other
        ref: "grep-гейт CHECKLIST-OK (expected: / goswitchctl install / uninstall / Summary) — pass"
        status: pass
    human_judgment: true
    rationale: "Чек-лист — артефакт прохода: смысл гейта — его ПРОХОЖДЕНИЕ владельцем в одной сессии (SPEC §7.2) на verify-work фазы; автоматизация здесь по определению неполна"
  - id: D3
    description: "WINDOWS-ledger сверен глаголами: #2 fixed, #4 waived с причиной, счётчики фронматтера согласованы; open_count = 1 (только #5)"
    verification:
      - kind: other
        ref: "gsd-tools windows fixed 2 / waive 4 <reason> — оба ok; фронматтер: open_count 1, waived_count 2, fixed_count 3 (gsp-гейт LEDGER-CLEAN == 0 недостижим честно, см. Deviations)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Релиз v1.0.0 (тег → release-workflow → ассеты + checksums + прокси-видимость): НЕ ВЫПОЛНЕН — останов на красном гейте"
    requirement: INST-01
    verification:
      - kind: e2e
        ref: "mise run e2e-perf (2026-09-16T14:58:14Z): budget FAIL — p95 184.4 ms >= 50 ms (третий подряд: 204.4 / 190.2 / 184.4); правило плана: красный гейт = СТОП, тега нет"
        status: fail
    human_judgment: true
    rationale: "Disposition вердикта — решение владельца (WINDOWS #5): (а) санкционировать uinput-инжектор стенда, (б) переопределить окно методики, (в) принять; без решения публикация запрещена планом"
  - id: D5
    description: "Остальные гейты тега зелёные на HEAD 8ac3451 (доказательная база чекпоинта): ci, install-cycle, матрица v3 локально"
    verification:
      - kind: e2e
        ref: "mise run ci — зелёный (build+vet+lint+test -race); mise run e2e-install-cycle — PASS (preflight 4/4 + полный жизненный цикл); go run ./test/e2e -matrix matrix-v3.yaml — 31/31 PASS"
        status: pass
    human_judgment: false

# Metrics
duration: 16min
completed: 2026-09-16
status: halted   # designed stop: красный perf-гейт (WINDOWS #5, решение владельца) — Task 3 (публикация v1.0.0) не исполнялся по правилу плана
---

# Phase 4 Plan 7: Публикация и приёмка Summary

**Двуязычный README с одно-командной установкой и живой perf-таблицей, письменный чек-лист приёмки владельца и сведённый WINDOWS-ledger — готово; публикация v1.0.0 ЧЕСТНО ОСТАНОВЛЕНА красным perf-гейтом (третий подряд FAIL: p95 184.4 мс ≥ 50 мс) — ждёт решения владельца по методике (WINDOWS #5)**

## Performance

- **Duration:** 16 min (14:46–15:02 UTC; живые гейты ~10 мин из них)
- **Started:** 2026-09-16T14:46:40Z
- **Completed:** 2026-09-16T15:02:00Z (останов на гейте Task 3)
- **Tasks:** 2 of 3 (Task 3: сверка гейтов исполнена, публикация — нет)
- **Files modified:** 4

## Accomplishments

- **Двуязычная пара README (D-50)**: README.md (EN, первичен) переписан, README.ru.md создан зеркально — что это и как работает (IBus engine, без root, сосуществование с keyd/xremap), требования, установка ОДНОЙ командой `goswitchctl install` (канал А: релизный архив; канал Б: `go install ...@v1.0.0` с честной оговоркой про `dev`), verify через `goswitchctl selfcheck` (шесть проверок, самопочинка кеша), жесты и CLI, конфиг (вшитые дефолты до первого YAML; схема и hot reload — ссылка на docs/CONFIG.md), perf-таблица, приватность, troubleshooting (пропавший движок → selfcheck; GTK4-Wayland ограничение), uninstall `--purge`, MIT. НИ ОДНОГО ручного шага ibus/systemctl и ни одного упоминания голого `ibus write-cache` (греп-гейт NO-BARE-CACHE в обоих языках); счётчики команд установки синхронны (en=4 ru=4).
- **docs/ACCEPTANCE.md (D-49)** — чек-лист приёмки владельца в формате 03-UAT: 6 нумерованных пунктов с expected/result/note (формальный D-48 свежесессионный двойной прогон по процедуре docs/ci-runner.md; установка с нуля по README без root + selfcheck шесть ok; живой набор всех жестов; hot-reload конфига с last-good-отказом; счётчики status; uninstall с восстановлением стола) + Summary-блок; владелец заполняет result/note в одной сессии (SPEC §7.2).
- **WINDOWS-ledger сведён глаголами** (Pitfall 11): #2 → fixed (владелец подтвердил чтение D-24 на verify-гейте Фазы 3, 2026-09-15), #4 → waived с дословной причиной из решения владельца (GTK4-Wayland forwarded-events platform limitation, UAT 2026-09-15, REQUIREMENTS MACR-01 caveat). open_count честно = 1: пункт #5 (латентный вердикт) — ожидание решения владельца, не вправе закрывать исполнитель.
- **Финальная сверка гейтов тега** (Task 3, до СТОПа): `mise run ci` ЗЕЛЁНЫЙ; `mise run e2e-install-cycle` PASS (preflight 4/4, полный жизненный цикл); матрица v3 локально **31/31 PASS**; `mise run e2e-perf` — **FAIL: p95 184.4 мс ≥ 50 мс** (p50 166.9, p99 186.5; VmHWM 12.4 МБ — память держит бюджет; прогон 2026-09-16T14:58:14Z, 40/40 повторов). По явному правилу плана (любой красный = СТОП, тег не ставить) публикация НЕ исполнялась: тега нет, релиза нет, запретительные грейты (STAMPED/RELEASE-ASSETS/PROXY) не порождались.

## Task Commits

1. **Task 1: Двуязычный README (D-50/D-46/INST-01)** - `b694e64` (docs)
2. **Task 2: docs/ACCEPTANCE.md (D-49) + WINDOWS-ledger сверка** - `8ac3451` (docs)
3. **Task 3: сверка гейтов → СТОП на красном perf-гейте; refresh README-таблицы на свежайший прогон** - `dd645f6` (docs); **тег/релиз — НЕТ (заблокированы)**

## Files Created/Modified

- `README.md` — переписан: публичный фронт EN (установка/selfcheck/жесты/конфиг/perf/приватность/troubleshooting/uninstall)
- `README.ru.md` — НОВЫЙ: русское зеркало, синхрон команд/чисел/версий дословно
- `docs/ACCEPTANCE.md` — НОВЫЙ: чек-лист UAT-гейта фазы (D-49, формат 03-UAT)
- `.planning/WINDOWS.md` — ledger: #2 fixed, #4 waived с причиной; open_count 1 (только #5)

## Decisions Made

- Perf-таблица в README публикует ЧЕСТНЫЕ живые числа с методикой (окно включает ~170 мс инжекции стенда ydotool 0.1.8; внутренняя реакция демона < 2 мс; память — с запасом) и примечанием об актуализации каждым релизом: репо-README не публичен до тега, а тег заблокирован — при смене методики таблица обновляется до тега (D-46-цикл).
- README-таблица обновлена до свежайшего прогона (dd645f6): perf-report.txt после гейт-сверки содержит новый прогон — числа обеих языков синхронно подняты (166.9/184.4/186.5, 12.4 МБ).
- docs/ACCEPTANCE.md ставит формальный D-48 свежесессионный прогон пунктом 1 ДО установки (прогон матрицы сам управляет источниками ввода через snapshot/restore; свежая сессия — его предусловие).
- Таблица производительности и чек-лист не раскрывают содержимого полей — только числа и вердикты (T-04-07-04).

## Deviations from Plan

### Auto-fixed Issues / documented departures

**1. [Rule 3 - Blocking, эскалирован] Гейт LEDGER-CLEAN (open_count: 0) недостижим честно**
- **Found during:** Task 2
- **Issue:** план писался при open-парах #2/#4 (Pitfall 11); 04-04 ДОБАВИЛА пункт #5 (unmet-truth по латентному вердикту, 2026-09-16T13:51Z — ПОСЛЕ ревизии плана). Предписанные глаголы исполнены точно (#2 fixed, #4 waived с причиной), но open_count == 1, а не 0.
- **Fix:** НЕ форсировать: waив #5 = решение вопроса методики за владельца (запрещено планом и правилами). Задокументировано здесь; #5 — центральный предмет чекпоинта.
- **Files modified:** .planning/WINDOWS.md (только глаголами)
- **Verification:** фронматтер согласован (open 1 / waived 2 / fixed 3 / total 6); глаголы вернули ok
- **Committed in:** 8ac3451

**2. (дополнение в духе плана) ACCEPTANCE.md включает формальный D-48-прогон пунктом 1**
- **Found during:** Task 2
- **Issue:** список пунктов плана (6 групп) не называл свежесессионный D-48-гейт — но это критерий №2 фазы, а 04-06 зарезервировала его «за шагом приёмки 04-07/verify-work» с готовой процедурой в docs/ci-runner.md
- **Fix:** добавлен пунктом 1 (до установки — свежая сессия его предусловие), остальные пункты сохранены дословно по плану
- **Files modified:** docs/ACCEPTANCE.md
- **Verification:** CHECKLIST-OK; порядок и содержание пунктов 2–6 соответствуют плану
- **Committed in:** 8ac3451

---

**Total deviations:** 2 (1 blocking-форма, эскалированная владельцу; 1 документированное дополнение объёма чек-листа). **Impact on plan:** цель Task 2 (ship-гигиена) достигнута настолько, насколько это честно возможно без решения владельца; на публикацию не влияет (она и так остановлена тем же #5).

## Issues Encountered

**ГЛАВНОЕ — красный гейт бюджета латентности (останов Task 3 по правилу плана):**
- `mise run e2e-perf` на HEAD 8ac3451: **p95 184.4 мс ≥ 50 мс** — третий подряд воспроизводимый FAIL (204.4 → 190.2 → 184.4; разброс объясним шумом стола). p50 166.9, p99 186.5, 40/40 повторов, VmHWM 12.4 МБ (памятный бюджет ДЕРЖИТСЯ с запасом).
- Декомпозиция 04-04 в силе: окно D-44 (ydotool→AT-SPI) доминируется оверхедом инжектора ydotool 0.1.8 (~80–125 мс/событие); реакция демона 0.2–1.8 мс. С пинненым инжектором окно математически не может уложиться в 50 мс.
- **Ожидает решения владельца (WINDOWS #5)**, варианты записаны владельцу 04-04: (а) санкционировать фолбэк STACK.md — собственный uinput-инжектор ~150 LOC stdlib ioctls, только для стенда; (б) переопределить окно методики (t0 на событие инжекции, исключив серийную стоимость инжектора); (в) принять результат как есть.
- После решения: обновить таблицу README при необходимости → перегнать `mise run e2e-perf` (зелёный или (в)) → закрыть/завайвить #5 → поставить тег. Тег ДО мержа фазы не публикуется: release.yml не на default branch (origin/main = d815e72) — либо мерж фазы, либо bootstrap прецедентом 02-07 (заметка в шапке release.yml, требование плана).
- Прочее: известный средовой флейк combo-word-layout (deferred-items) не воспроизвёлся — матрица v3 зелёная 31/31 с первого прогона.

## Known Stubs

None — все созданные артефакты полны; единственный незакрытый пункт (публикация релиза) — не заглушка, а честный останов на гейте.

## User Setup Required

None - no external service configuration required. (Проход docs/ACCEPTANCE.md — шаг приёмки владельца на verify-work фазы, не настройка.)

## Next Phase Readiness

- Всё, кроме публикации, готово и зелёное на HEAD dd645f6: доки-пара, чек-лист, ledger (один честно открытый пункт), ci/install-cycle/v3.
- Для завершения плана после решения владельца по WINDOWS #5: (1) при методике (а)/(б) — обновить стенд/таблицу и перегнать e2e-perf; (2) закрыть #5 (fixed или waived с причиной); (3) мерж фазы (или bootstrap 02-07); (4) тег v1.0.0 → gh run watch (release) → проверки ассетов/checksum/-version/прокси по Task 3; (5) владелец проходит docs/ACCEPTANCE.md (включая формальный D-48 прогон) на verify-work фазы.

---
*Phase: 04-postavka-i-priemka*
*Completed: 2026-09-16 (частично — статус halted)*

## Self-Check: PASSED

- Files on disk: README.md, README.ru.md, docs/ACCEPTANCE.md, .planning/WINDOWS.md (open_count: 1) — all FOUND
- Commits: b694e64, 8ac3451, dd645f6 — all present in git log on gsd/phase-04-postavka-i-priemka
- commits: 3 MEASURED from ledger gsd-plan-head-before-04-07 (657fbc4)
- Gates at stop: ci GREEN, e2e-install-cycle PASS, matrix v3 31/31 PASS, e2e-perf FAIL (p95 184.4 >= 50) — the STOP evidence
