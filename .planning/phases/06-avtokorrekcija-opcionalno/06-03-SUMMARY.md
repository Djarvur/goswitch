---
phase: 06-avtokorrekcija-opcionalno
plan: 03
subsystem: appid
tags: [at-spi, a11y-bus, godbus, role-enum, fail-closed, d-53, tdd, seam-injection]

# Dependency graph
requires:
  - phase: 03 (internal/appid Observer)
    provides: a11y-Observer с фокус-событиями bridge-namespace, ErrBusClosed-сентинел-дисциплина, Start/GetAddress-прецедент вызова, синтетический фид-корпус
  - phase: 06-avtokorrekcija-opcionalno (06-01)
    provides: ADR-007 (Proposed) — живой GetRole в момент решения, отказ от кэш-семантики, fail-closed (D-53)
provides:
  - Observer хранит (focusSender, focusPath) последнего focus-gain под тем же мьютексом, что и имя приложения; loss/не-bridge события пару держат (cache-keep, зеркально FocusedApp)
  - Role(ctx) (uint32, error) — живой org.a11y.atspi.Accessible.GetRole-запрос в момент решения через инжектируемый шов roleCall; Start инжектирует живое godbus-замыкание по собственной a11y-связи
  - ErrRoleUnknown (errors.Is-дружелюбный сентинел, образец ErrBusClosed): пустой путь / nil-шов / ошибка шва — error return, потребитель (план 06-06) трактует как молчание
  - enum-константы RolePasswordText=40, RoleTerminal=60, RoleText=61, RoleEntry=79, RoleDocumentText=94 (gi Atspi 2.52.0, живо верифицированы 06-RESEARCH Q4); GetRoleName не вызывается нигде
  - Корпус appid_test.go: 8 тестов поверх существующих (пара хранится/держится/неизвестна; enum-пин; форма вызова; nil-шов; гонки под -race)
affects: [06-06 session (RoleSource-потребление: ensureAppid тайп-ассертит src.(RoleSource), конъюнкция D-53), 06-07 e2e (живой пин ролей Chromium/VTE)]

# Actuals (#2632)
actuals:
  tokens: 6027        # chars/4 over the realized diff (24 108 bytes) — оценка плана 35 000 пере-оценила объём: расширение существующего пакета, не новый механизм
  tasks: 2
  commits: 4          # MEASURED: git rev-list --count gsd-plan-head-before-06-03..HEAD

# Tech tracking
tech-stack:
  added: []           # ноль новых зависимостей — godbus v5.2.2 уже в go.mod (граница стека соблюдена, tidy-diff зелёный)
  patterns:
    - "Живой D-Bus-запрос по сохранённому адресу события: кэш пары — только адресат запроса, ответ никогда не кэшируется (D-53 дословно)"
    - "Инжектируемый шов через variadic Option (WithRoleCall) — New остаётся сигнатурой фазы 3, живое замыкание задаётся в Start, тесты кормят рекордеры"
    - "Сентинел-дисциплина ErrBusClosed воспроизведена: ErrRoleUnknown = roleUnknownError{} (errors.Is-дружелюбный конкретный тип, без мутабельного глобала)"

key-files:
  created:
    - .planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-03-task1-red.json
    - .planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-03-task2-red.json
  modified:
    - internal/appid/appid.go
    - internal/appid/appid_test.go

key-decisions:
  - "Тип пары — dbus.Sender, а не сырой string: godbus доставляет Signal.Sender как plain string, наблюдатель конвертирует на границе сигнала (dbus.Sender(sig.Sender)), живое замыкание конвертирует обратно (conn.Object(string(sender), path) — Conn.Object принимает string); тип шва самодокументирует wire-роль"
  - "Role НЕ проверяет o.err (флаг выхода наблюдателя): арбитр — живой вызов (D-53); мёртвая связь валит wire-вызов fail-closed, а пара под закрытым фидом остаётся честным адресом последнего gain"
  - "Роль никогда не резолвится по строковому имени (T-06-03-04): сравнение ТОЛЬКО по enum-числу; греп-гейт GetRoleName==0 в appid.go; имена локализованы и расходятся («text box» wire vs «text» gi-nick для того же 61)"
  - "Снапшот пары под мьютексом, затем вызов шва ВНЕ блокировки (T-06-03-02): ни одной блокировки на D-Bus round trip; шов задаётся при конструкции и не мутирует — чтение поля без замкки гонобезопасно"
  - "requirements-completed — провизорная метка плана: MACR-ACL фиксируется в REQUIREMENTS.md только при new-milestone v1.1.0 (06-RESEARCH phase_requirements; файл сегодня несёт только v1.0.0 REQ-ID)"

patterns-established:
  - "RoleSource-поверхность (неявная): метод Role(ctx context.Context) (uint32, error) — план 06-06 тайп-ассертит src.(RoleSource) в точке потребления, интерфейс определяется у потребителя (AppidSource-прецедент)"
  - "RED-evidence для Go-корпуса: TAP-транскрипция вывода go test -v в record JSON (прецедент 04-01/04-04) — check tdd-red-evidence парсит только node-TAP"

requirements-completed: ["MACR-ACL (app×role политика)"]  # провизорный ID фронта плана; формальная фиксация REQ-ID — при new-milestone v1.1.0 (D-51); ролевая половина политики (условие 2 конъюнкции D-53)

coverage:
  - id: D1
    description: "Observer хранит (focusSender, focusPath) последнего focus-gain под тем же мьютексом, что и имя приложения; loss-события и немаркированные пути пару не сбрасывают (cache-keep); свежий наблюдатель = неизвестно"
    requirement: "MACR-ACL (app×role политика)"
    verification:
      - kind: unit
        ref: "tests/internal/appid/appid_test.go#TestObserver_StoresFocusPair"
        status: pass
      - kind: unit
        ref: "tests/internal/appid/appid_test.go#TestObserver_LossKeepsPair"
        status: pass
      - kind: unit
        ref: "tests/internal/appid/appid_test.go#TestObserver_FreshObserverUnknown"
        status: pass
    human_judgment: false
  - id: D2
    description: "Role(ctx) — живой GetRole-запрос через инжектируемый шов по паре последнего gain; пустой путь/nil-шов/ошибка шва — error return (errors.Is(err, ErrRoleUnknown) для первых двух); enum-константы запинены живо-верифицированными значениями; GetRoleName отсутствует; гонобезопасно под -race"
    requirement: "MACR-ACL (app×role политика)"
    verification:
      - kind: unit
        ref: "tests/internal/appid/appid_test.go#TestRole_EnumConstants"
        status: pass
      - kind: unit
        ref: "tests/internal/appid/appid_test.go#TestRole_EmptyPathUnknown"
        status: pass
      - kind: unit
        ref: "tests/internal/appid/appid_test.go#TestRole_LiveCallShape"
        status: pass
      - kind: unit
        ref: "tests/internal/appid/appid_test.go#TestRole_NilSeamUnknown"
        status: pass
      - kind: unit
        ref: "tests/internal/appid/appid_test.go#TestRole_ConcurrentWithGain"
        status: pass
      - kind: other
        ref: "grep-гейты плана: GetRoleName==0; RolePasswordText/ErrRoleUnknown/org.a11y.atspi.Accessible.GetRole присутствуют — ROLE-SEAM-OK"
        status: pass
    human_judgment: false
  - id: D3
    description: "Start инжектирует живое godbus-замыкание org.a11y.atspi.Accessible.GetRole по собственной a11y-связи (CallWithContext().Store() — форма GetAddress-прецедента)"
    verification:
      - kind: other
        ref: "grep-гейт: wire-литерал org.a11y.atspi.Accessible.GetRole в appid.go; форма вызова — прецедент appid.go GetAddress"
        status: pass
    human_judgment: true
    rationale: "Живая шина в headless-корпусе не участвует по построению шва (тесты кормят рекордеры); реальный ответ GetRole по пути фокус-события пинится e2e плана 06-07 (предположения плана A2/A3/A6: Chromium-пароль и VTE — живой пин на 06-07), форма вызова верифицирована busctl в 06-RESEARCH Q4"

# Metrics
duration: 23min
completed: 2026-10-01
status: complete
---

# Phase 6 Plan 3: internal/appid — (sender,path) + Role(ctx) Summary

**Живой источник роли AT-SPI для политики D-53: Observer хранит (sender,path) последнего фокус-gain с cache-keep семантикой, Role(ctx) делает живой org.a11y.atspi.Accessible.GetRole по enum-числу через инжектируемый шов, любое «неизвестно» — error return (ErrRoleUnknown), потребитель молчит**

## Performance

- **Duration:** 23 min (21:01–21:24 UTC)
- **Started:** 2026-10-01T21:01:03Z
- **Completed:** 2026-10-01T21:23:28Z
- **Tasks:** 2 (оба tdd, строгий RED→GREEN)
- **Files modified:** 2 (+2 red-evidence записи)

## Accomplishments
- Observer расширен полями focusSender (dbus.Sender) / focusPath под существующим мьютексом: на каждом bridge-gain пишутся все три значения одной критической секцией; loss/не-bridge пару держат (cache-keep, зеркально FocusedApp)
- Role(ctx) (uint32, error): снапшот пары под мьютексом → один вызов инжектируемого шва вне блокировки; пустой путь/nil-шов → fmt.Errorf("role: %w", ErrRoleUnknown); шов ошибается — ошибка доходит до потребителя (fail-closed по построению); Role не пишет в лог вовсе
- Start инжектирует живое замыкание conn.Object(string(sender), path).CallWithContext(ctx, "org.a11y.atspi.Accessible.GetRole", 0).Store(&role) — форма GetAddress-прецедента; New остался сигнатурой фазы 3 (variadic Option — существующие вызовечки не тронуты)
- enum-константы 40/60/61/79/94 запинены тестом с комментарием-источником (gi Atspi 2.52.0, живая верификация 2026-09-27); GetRoleName в appid.go — ноль вхождений (греп-гейт)
- Полный корпус 10 top-level тестов зелёный под -race; mise run ci (build+vet+lint+test -race) зелёный; ноль новых зависимостей

## Task Commits

Each task was committed atomically (TDD: RED → GREEN per task):

1. **Task 1: Хранение (focusSender, focusPath) на focus-gain** — `889e3b7` (test, RED) → `e37a295` (feat, GREEN)
2. **Task 2: Role(ctx) — живой GetRole, сентинел, константы ролей** — `19e2ff4` (test, RED) → `00d7278` (feat, GREEN)

**Plan metadata:** `*(этот коммит — docs(06-03))*

_Note: REFACTOR-коммиты не понадобились — линтер-риффы (intrange/fatcontext/wrapcheck) вошли в GREEN-коммиты задач._

## TDD Gate Compliance

| Task | RED | GREEN | REFACTOR | RED evidence |
| ---- | --- | ----- | -------- | ------------ |
| 1 | `889e3b7` ✓ | `e37a295` ✓ | — (вошёл в GREEN) | RED_EVIDENCE_OK — TestObserver_StoresFocusPair упал на утверждении против стаба (exit 1, 3/5 красных, 2 старых зелёных) |
| 2 | `19e2ff4` ✓ | `00d7278` ✓ | — (не требовался) | RED_EVIDENCE_OK — TestRole_EnumConstants упал на живых значениях против нулевых стабов (exit 1, 9/10 зелёных: EmptyPathUnknown/NilSeamUnknown/LiveCallShape зелёные по построению — шов и сентинел уже землись GREEN'ом Task 1, прецедент continuity-pin 04-01) |

Обе записи: `.planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-03-task{1,2}-red.json` (TAP-транскрипция per repo precedent — verb парсит только node-TAP).

## Files Created/Modified
- `internal/appid/appid.go` — поля focusSender/focusPath/roleCall; Option/WithRoleCall; ErrRoleUnknown + roleUnknownError; константы Role*; getRoleMethod; Role(ctx); Start-замыкание
- `internal/appid/appid_test.go` — focusSignal с sender-полем; корпус-константы sender/path; +8 тестов (пара/loss/свежий/enum/пустой-путь/форма/nil-шов/гонки)
- `.planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-03-task1-red.json` — RED-evidence Task 1
- `.planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-03-task2-red.json` — RED-evidence Task 2

## Decisions Made
- **dbus.Sender как тип шва** — godbus отдаёт Signal.Sender строкой; конвертация на границах (сигнал → поле → замыкание), тип документирует wire-роль; conn.Object(string(sender), path) — план сам предписывал строку на вызове
- **Role не смотрит o.err** — арбитр неизвестности живой вызов (D-53): мёртвая связь = ошибка wire-вызова, fail-closed; проверка o.err добавила бы непротестированную ветку без изменения исхода
- **Сентинел в финальной форме уже в Task 1 GREEN** — err113 запрещает динамические ошибки в internal-коде, одноразовый плейсхолдер был бы линт-нарушением; Task 2 RED поэтому пинит enum (остальные Role-тесты зелёные по построению, задокументировано)
- **Снапшот-затем-вызов** — мьютекс не держится на D-Bus round trip (T-06-03-02); roleCall иммутабелен после конструкции — чтение без замкки безопасно

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking lint] wrapcheck: wrap транспортной ошибки перенесён из Role в Start-замыкание**
- **Found during:** Task 2 (GREEN)
- **Issue:** скетч плана возвращал Store-ошибку из замыкания сырой (`return role, err`) — wrapcheck флагает незамкнутую ошибку внешнего пакета (dbus.Call.Store); повторный wrap в Role дал бы «get role: get role: …»
- **Fix:** замыкание оборачивает `fmt.Errorf("get role: %w", err)` (wire-слой называет операцию), Role отдаёт ответ шва как есть; errors.Is-контракт прежний, префикс единственный и чистый
- **Files modified:** internal/appid/appid.go
- **Verification:** TestRole_LiveCallShape (errors.Is(err, errSeamDead)) зелёный; пакет и mise run ci зелёные
- **Committed in:** 00d7278 (Task 2 commit)

**2. [Rule 3 - Blocking lint] err113: сентинел ErrRoleUnknown встал в Task 1 GREEN, а не в Task 2**
- **Found during:** Task 1 (GREEN)
- **Issue:** минимальный «неизвестно»-ответ Task 1 требует ошибки; динамический errors.New/fmt.Errorf без %w в internal-коде запрещён err113 (конфиг: только test/e2e исключён)
- **Fix:** сентинел в финальной форме (ErrRoleUnknown + roleUnknownError) объявлен в Task 1 GREEN; Task 2 RED от этого не пострадал — целевой RED-тест (EnumConstants) падает на живых значениях, три Role-теста зелёные по построению (прецедент continuity-pin 04-01, задокументировано в evidence-записи)
- **Files modified:** internal/appid/appid.go
- **Verification:** RED_EVIDENCE_OK задачи 2; оба гейта закрыты
- **Committed in:** e37a295 (Task 1 GREEN)

**3. [Rule 3 - Blocking lint] intrange + fatcontext в тестовом корпусе**
- **Found during:** Task 2 (GREEN)
- **Issue:** классические циклы `for i := 0; i < 50; i++` — intrange требует range-over-int (Go 1.22+, модуль 1.23); именованный ctx-параметр шва — fatcontext
- **Fix:** циклы переведены на `for i := range 50` / `for range 50`; fatcontext подавлен точечным `//nolint:fatcontext` с обоснованием (ctx-параметр шва и ЕСТЬ утверждение — Role обязан донести ctx вызывающего до wire-вызова, D-53)
- **Files modified:** internal/appid/appid_test.go
- **Verification:** golangci-lint run — 0 issues; mise run ci зелёный
- **Committed in:** 00d7278 (Task 2 commit)

---

**Total deviations:** 3 auto-fixed (3 × Rule 3 blocking-lint). **Impact on plan:** все три — принуждение линтера к эквивалентным по поведению формам; контракт D-53 и errors.Is-дисциплина не изменены. Скоуп-крипа нет.

## Issues Encountered
- Транзиентный сбой первого `mise run ci` на закрытии Task 2: голый FAIL от одного пакета при параллельной нагрузке (окно tail не поймало имя; два немедленных полных перегона — test и ci — зелёные). Классифицировано как флак под нагрузкой, не дефект кода; оба закрытия задач подтверждены зелёными полными прогонами
- Первый RED-прогон Task 2 поймал баг самого тест-фикстуры (TestRole_LiveCallShape делил один feed-канал между двумя наблюдателями — канал обслуживает одного получателя, второй голодал): фикстура исправлена до записи evidence — красный прогон остался только у целевого теста на planned-утверждении

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Планы 06-04/06-05 (wave 2/3: конфиг autocorrect + детектор) не зависят от appid — могут идти параллельно
- План 06-06 (актор): ensureAppid тайп-ассертит src.(RoleSource); конъюнкция D-53 получает Role + FocusedApp от одного Observer; warn-once-дисциплина — на потребителе
- План 06-07 (e2e): живой пин ролей Chromium-пароля и VTE (предположения A2/A3 этого плана), форма GetRole уже верифицирована busctl

## Self-Check: PASSED

- Все key-files существуют на диске (appid.go, appid_test.go, 2 × red-evidence, SUMMARY)
- Все 4 коммита задач найдены в git log (889e3b7, e37a295, 19e2ff4, 00d7278); измерено `git rev-list --count gsd-plan-head-before-06-03..HEAD` = 4
- TDD-гейт: оба RED-коммита предшествуют своим GREEN (889e3b7<e37a295, 19e2ff4<00d7278); check tdd-red-evidence — RED_EVIDENCE_OK ×2
- Верификация плана перезапущена на финальном дереве: `go test ./internal/appid/ -race -count=1` ok; греп-гейты ROLE-SEAM-OK / GetRoleName=0 / enum-пины; `mise run ci` зелёный

---
*Phase: 06-avtokorrekcija-opcionalno*
*Completed: 2026-10-01*
