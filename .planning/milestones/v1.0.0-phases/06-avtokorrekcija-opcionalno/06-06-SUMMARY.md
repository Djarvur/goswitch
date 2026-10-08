---
phase: 06-avtokorrekcija-opcionalno
plan: 06
subsystem: session
tags: [autocorrect, actor, word-boundary, d-53, fail-closed, live-getrole, off-mutex, counters, status-tokens, tdd]

# Dependency graph
requires:
  - phase: 06-avtokorrekcija-opcionalno (06-03)
    provides: appid.Observer.Role(ctx) (uint32, error) + ErrRoleUnknown + enum 40/60/61/79/94 — RoleSource-шов
  - phase: 06-avtokorrekcija-opcionalno (06-04)
    provides: snap.Autocorrect.{Enabled,Apps,MinWordLen,TrigramMargin,TrigramFloor} — источник Options-фолда, дефолты off 4/2.0/1.0
  - phase: 06-avtokorrekcija-opcionalno (06-05)
    provides: detect.Check(tok, mode, Data, Trigrams, Params) — уверенный вердикт = условие 4 конъюнкции, замкнутые слаги причин
  - phase: 02-05 (internal/correct)
    provides: Buffer Push/Token/Tail/ReplaceToken зеркало (D-13) и конвейер startRangeCorrection (единственный путь замены, D-52)
provides:
  - Buffer.PushFeed(r rune) bool — вся логика Push + честный флаг границы («разделитель завершил непустой токен», проверка до мутации); Push — тонкая обёртка, нулевой риппл
  - Акторный хук границы: feedKey зовёт PushFeed на всех трёх сайтах Push и вооружает ДО CORR-09 HardReset; consume-вердикты и reset-семантика байт-нетронуты; FocusOut — НЕ граница
  - Конъюнкция D-53: дешёвые гейты под мьютексом (enabled → white-list точным равенством → капс → детектор), затем armed-payload handoff — живой GetRole ВНЕ мьютекса под autoRoleTimeout 25 мс; любое «неизвестно» → молчание со слагом (fail-closed, инверсия MACR-деградации)
  - Подтверждение только через СУЩЕСТВУЮЩИЙ конвейер startRangeCorrection (единственный путь замены; греп-пин CommitText-сайтов неизменен); generation-счётчик отменяет устаревший confirm; re-check enabled на re-entry
  - Счётчики D-54: AutoCorrectStats{Fired, Abstained, Reasons} + AutoCorrectCounters(); Status += 4 поля; renderStatus: autocorrect_enabled/fired/abstained + сортированные ac_skip_<slug> перед config-блоком (config_error последним); goswitchctl/cmd mains — ноль правок
affects: [06-07 (e2e: fires/password-silent/terminal-silent по счётчикам-оракулу), 06-08 (матрица/приёмка, perf p95)]

# Actuals (#2632)
actuals:
  tokens: 16739       # chars/4 over the realized diff (66 955 bytes) — оценка плана 72 000 близка
  tasks: 3
  commits: 7          # MEASURED: git rev-list --count gsd-plan-head-before-06-06..HEAD

# Tech tracking
tech-stack:
  added: []           # ноль новых зависимостей (go.mod не тронут; tidy-diff зелёный)
  patterns:
    - "Граница слова как флаг буфера: PushFeed возвращает «разделитель завершил непустой токен», актор вооружает решение только на границе — горячий путь клавиш не тронут"
    - "Armed-payload confirm вне мьютекса: дешёвые гейты под a.mu, живой GetRole с собственным дедлайном в горутине, вердикт возвращается под мьютексом с re-check enabled+generation"
    - "Fail-closed silence-matrix как корпус политики: каждая ячейка «неизвестно» — ноль операций + ровно один слаг; направление инвертировано относительно MACR fail-open"

key-files:
  created:
    - .planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-06-task1-red.json
    - .planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-06-task2-red.json
    - .planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-06-task3-red.json
  modified:
    - internal/correct/buffer.go
    - internal/correct/buffer_test.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - internal/ctlsvc/ctlsvc.go
    - internal/ctlsvc/ctlsvc_test.go

key-decisions:
  - "Подтверждение (autoConfirm) с Task-1 — асинхронная горутина с generation-счётчиком (план: «запускаем confirm-горутину»): новая граница отменяет устаревший confirm, роль — одна за раз; Task-1-форма исполняла pipeline без ролевого гейта (scaffold), Task-2 встраила D-53-контуры в ту же форму"
  - "Off-состояние — ноль счётчиков (behavior-блок TestActor_AutocorrectDisabledSilent старше шorthand'а действия «причина disabled»): feedKey-гейт enabled стоит ДО всякого счёта, байт-как-сегодня; слаг disabled в словаре не нужен"
  - "Длина слова гейтится самим детектором (Params.MinWordLen из Options): отдельная актop-проверка дублировала бы слаг abstain-short — один источник, тот же наблюдаемый исход"
  - "RoleSource-константы 61/79/94 определены в акторе с комментарием-источником appid (план: числовые литералы); сравнение ТОЛЬКО по числу — GetRoleName-ловушка"
  - "renderStatus-тест живёт в ctlsvc_test.go (прецедент TestRenderStatusVersionToken), end-to-end — в session_test с реальным актором как StatusSnapshotProvider; goswitchctl-парсер generic — CLI не тронут (CLI-WIRING-FROZEN)"

patterns-established:
  - "Silence-matrix-корпус: таблица ячеек «неизвестно→слаг» с общим assertSilence (ноль операций, ровно один счёт) — переиспользуемая форма для политик fail-closed"
  - "fakeRole double с started/release-каналами и ctx- honoring: доказывает off-mutex (HandleKey жив при блокированном вызове) и дедлайн-ограниченность (подтверждение завершается без release)"

requirements-completed: ["CORR-01", "CORR-07", "CORR-09", "MACR-ACL (app×role политика)"]  # провизорные метки плана: CORR-01/07/09 заблокированы шлагбаумом #2388 (разделяются с 06-07/06-08 без SUMMARY); свободная метка MACR-ACL не размечается verb'ом (пробел-лейбл, прецедент 06-02/06-05); формальная фиксация REQ-ID — при new-milestone v1.1.0 (D-51)

coverage:
  - id: D1
    description: "Граница слова: Buffer.PushFeed возвращает честный флаг (пустой/повторный разделитель false, после ReplaceToken true), feedKey зовёт её на всех трёх сайтах Push, reset-граница вооружает ДО HardReset при транзите Enter, забой и FocusOut не границы, consume-вердикты нетронуты; ручной Double конвертирует привет обратно после автокоррекции (критерий 4)"
    requirement: "CORR-01"
    verification:
      - kind: unit
        ref: "tests/internal/correct/buffer_test.go#TestBuffer_PushFeed"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_AutocorrectDisabledSilent + TestActor_AutocorrectBoundaryArms + TestActor_ResetKeyBoundary + TestActor_BackspaceNoBoundary + TestActor_FocusOutNoBoundary + TestActor_ManualOverrideAfterAutocorrect"
        status: pass
    human_judgment: false
  - id: D2
    description: "Дешёвая половина конъюнкции D-53 под мьютексом: enabled (off = ноль поведения и счётчиков), непустой white-list точным равенством bridge-namespace (расширение пространства не матчит), капс surrounding-text, фокус-identity (nil/ошибка → app-unknown с warn-once), detect.Check уверенный wrong-layout на запечённых данных (короткое/unsure — слагы детектора)"
    requirement: "MACR-ACL (app×role политика)"
    verification:
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestAutoCorrect_SilenceMatrix (10 ячеек, включая namespace-extension) + TestAutoCorrect_FailClosedNoFailOpen"
        status: pass
    human_judgment: false
  - id: D3
    description: "Роль-контур D-53: живой GetRole ВНЕ мьютекса под autoRoleTimeout 25 мс (HandleKey жив при блокированном вызове; без release confirm завершается role-timeout), роли 61/79/94 разрешены, 40/60/прочее/ошибка — молчание со слагом, warn-once per эпизод со сбросом здоровым ответом, re-check enabled на re-entry, fired только через startRangeCorrection (греп-пин CommitText-сайтов), счётчики Fired/Abstained/Reasons, слово никогда в журнале"
    requirement: "MACR-ACL (app×role политика)"
    verification:
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestAutoCorrect_FiresThroughPipeline + TestAutoCorrect_RoleDeadlineBounded + TestAutoCorrect_WarnOncePerEpisode + TestAutoCorrect_ReEntryRechecks + TestAutoCorrect_CountersAndReasons"
        status: pass
      - kind: other
        ref: "греп-гейт SINGLE-PATH-OK: grep -c CommitText actor.go == HEAD~1 (5) — второй механизм замены отсутствует"
        status: pass
    human_judgment: false
  - id: D4
    description: "Status/ctl-поверхность D-54: StatusSnapshot несёт 4 новых поля с клоном карты (мутация снапшота не портит актор), applySnapshot фолдит все пять полей живьём (появление/исчезновение секции, пороги, одноразовый старт наблюдателя), renderStatus добавляет autocorrect_enabled/fired/abstained + сортированные ac_skip_<slug> перед config-блоком, config_error последним; cmd mains и goswitchctl — ноль правок"
    requirement: "CORR-09"
    verification:
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestStatus_AutocorrectFields + TestApplySnapshot_AutocorrectFold"
        status: pass
      - kind: unit
        ref: "tests/internal/ctlsvc/ctlsvc_test.go#TestRenderStatus_AutocorrectTokens"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestCtlStatus_EndToEnd"
        status: pass
      - kind: other
        ref: "греп-гейт CLI-WIRING-FROZEN: git diff --exit-code -- cmd/goswitchctl/main.go cmd/goswitchd/main.go"
        status: pass
    human_judgment: false
  - id: D5
    description: "Зелёная итерация D-08: mise run ci (build+vet+lint strict+test -race) зелёный на закрытии каждой задачи и на финальном дереве; весь session/correct/ctlsvc корпус (ручные контракты CORR-01..09) зелёный под -race; ноль новых зависимостей"
    requirement: "CORR-07"
    verification:
      - kind: other
        ref: "mise run ci exit 0 (2026-10-02, /tmp/06-06-ci4.log; финальный прогон после пинов допущений); go.mod/go.sum не тронуты"
        status: pass
    human_judgment: false
  - id: D6
    description: "Живые негативы на реальном столе (реальный GetRole по a11y-шине, пароль/терминал не тронуты живьём, p95-сравнение латентности) — по плану носится планами 06-07/06-08 (e2e-кейсы по счётчикам-оракулу; юнит-пины этого плана их сужают)"
    verification: []
    human_judgment: true
    rationale: "Юнит-корпус доказывает политику на фейках шва; живое поведение GetRole на настоящей a11y-шине и e2e-негативы (GTK4-пароль, wezterm-счётчики) — скоуп плана 06-07, perf-сравнение — 06-08 (Pitfall 5)"

# Metrics
duration: 38min
completed: 2026-10-01
status: complete
---

# Phase 6 Plan 6: Акторная интеграция автокоррекции (D-53) Summary

**Граница слова через Buffer.PushFeed вооружает конъюнкцию D-53 в акторе: дешёвые гейты под мьютексом, живой GetRole вне его под дедлайном 25 мс, любое «неизвестно» — молчание со слагом-счётчиком (fail-closed), замена — только через существующий конвейер; status/ctl получают autocorrect-токены, ручной Double остаётся переопределением, cmd-файлы не тронуты.**

## Performance

- **Duration:** 38 min (22:53–23:31 UTC)
- **Started:** 2026-10-01T22:53:49Z
- **Completed:** 2026-10-01T23:31:55Z
- **Tasks:** 3 (каждая — полный RED→GREEN по RED_EVIDENCE_OK)
- **Files modified:** 9 (6 кода/тестов + 3 red-evidence)

## Accomplishments

- Buffer.PushFeed(r) bool: вся логика Push переехала в флаг-несущую форму («разделитель завершил непустой токен», проверка до мутации), Push — тонкая обёртка; существующий буферный корпус без правок
- feedKey зовёт PushFeed на всех трёх сайтах Push и вооружает решение ДО CORR-09 HardReset; consume-вердикты веток, reset-семантика и FocusOut-поведение байт-нетронуты (корпус пинит транзит границы и «FocusOut — не граница»)
- Дешёвая половина конъюнкции D-53 под a.mu: enabled (off — ноль счётчиков, D-54), white-list точным равенством bridge-namespace, капс surrounding-text, фокус-identity с warn-once эпизодами, detect.Check на реальных запечённых данных (DictRU/EN, TriRU/EN) — каждый отказ считает свой слаг
- Живой GetRole — вне мьютекса, под autoRoleTimeout 25 мс (половина бюджета SPEC §5): HandleKey остаётся живым при блокированном вызове, confirm завершается role-timeout без release; 61/79/94 — fired через СУЩЕСТВУЮЩИЙ startRangeCorrection (греп-пин CommitText-сайтов), 40/60/прочее/ошибка — молчание; generation отменяет устаревший confirm, re-check enabled глушит несбывшийся
- Счётчики D-54 и ctl-поверхность: AutoCorrectCounters/AutoCorrectStats, Status += 4 поля с клоном карты, applySnapshot фолдит секцию живьём (включая одноразовый старт наблюдателя), renderStatus несёт autocorrect_enabled/fired/abstained + сортированные ac_skip_<slug> перед config-блоком; goswitchctl и cmd mains — ноль правок (CLI-WIRING-FROZEN)
- mise run ci зелёный на закрытии каждой задачи; полный session/correct/ctlsvc корпус (все ручные контракты) зелёный под -race; ноль новых зависимостей

## Task Commits

Each task was committed atomically (TDD: RED → GREEN per task):

1. **Task 1: Граница слова — PushFeed + хук дешёвых гейтов** — `3009559` (test, RED) → `bd620f8` (feat, GREEN)
2. **Task 2: Роль-контур D-53 — живой GetRole вне мьютекса, fail-closed, счётчики** — `e543603` (test, RED) → `9f92c8c` (feat, GREEN)
3. **Task 3: Status/ctl-поверхность — фолд, renderStatus-токены** — `9102bac` (test, RED) → `302542b` (feat, GREEN)
4. **Пины допущений (после GREEN)** — `e33893a` (test — continuity-pin: точное равенство namespace, FocusOut-не-граница; оба зелёные против уже поставленного контура)

**Plan metadata:** _этот коммит — docs(06-06)_

## TDD Gate Compliance

| Task | RED | GREEN | REFACTOR | RED evidence |
| ---- | --- | ----- | -------- | ------------ |
| 1 | `3009559` ✓ | `bd620f8` ✓ | — (риффы в GREEN) | RED_EVIDENCE_OK — 5 целевых падений на утверждениях (PushFeed-флаг, арминг, reset-до-сброса, ручной override); DisabledSilent/BackspaceNoBoundary зелёные по построению (off/нет-границы = состояние стаба) |
| 2 | `e543603` ✓ | `9f92c8c` ✓ | — (funlen/goconst/unused-риффы в GREEN) | RED_EVIDENCE_OK — 10 падений по всему контуру (4 ролевые ячейки матрицы, warn-once, дедлайн, re-entry, счётчики, Fired-подсчёт); 6 зелёных по построению (синхронные ячейки Task-1-гейтов + FailClosed) |
| 3 | `9102bac` ✓ | `302542b` ✓ | — (appendCombine/lll-риффы в GREEN) | RED_EVIDENCE_OK — 4/4 падают (поля не заполняются, фолда нет, токенов нет) |

Все три RED-корпуса прошли `check tdd-red-evidence` с RED_EVIDENCE_OK (целевой тест падает на утверждении запланированного поведения). Коммит `e33893a` — test-тип ПОСЛЕ GREEN: пины допущений плана (namespace-extension ячейка, FocusOut) против уже поставленного поведения — прецедент continuity-pin 06-03/06-04, гейт RED-до-GREEN не нарушен (все три пары предшествуют своим GREEN).

## Files Created/Modified

- `internal/correct/buffer.go` — PushFeed(r) bool (логика Push + флаг границы до мутации), Push — обёртка
- `internal/correct/buffer_test.go` — TestBuffer_PushFeed (три сабтеста: буквы/разделитель, пустой/повторный, честность после ReplaceToken)
- `internal/session/actor.go` — Options += 5 AutoCorrect-полей; autoRoleTimeout=25ms; acPayload{rng,dir,word,gen}; константы ролей 61/79/94 и слагов; RoleSource + UseRole; ensureAppid — расширенный гейт + тайп-ассерт; поля acFired/acAbstained/acReasons/acGeneration/acAppWarned/acRoleWarned; pushAndArm/armAutoCorrect/autoCorrectBoundary/acFocusedApp/warnACApp/warnACRole/recordACAbstain/autoConfirm; AutoCorrectStats/AutoCorrectCounters; Status += 4 поля + StatusSnapshot-фолл; applySnapshot-фолд секции
- `internal/session/actor_test.go` — корпус: граница/сброс/забой/FocusOut/manual-override; fakeRole (started/release/ctx-honoring); silence-matrix (10 ячеек) + assertSilence; fires/warn-once/дедлайн/re-entry/счётчики; Status/fold/end-to-end; fireAutocorrect-танец с eventually-поллингом
- `internal/ctlsvc/ctlsvc.go` — renderStatus: autocorrect-блок перед config-блоком
- `internal/ctlsvc/ctlsvc_test.go` — TestRenderStatus_AutocorrectTokens (порядок, сортировка, подчёркивания, config_error последним)
- `.planning/.../red-evidence/06-06-task{1,2,3}-red.json` — RED-записи (все RED_EVIDENCE_OK)

## Decisions Made

См. key-decisions: асинхронный confirm с generation уже в Task-1-скаффолде; off = ноль счётчиков (behavior старше shorthand'а действия); длина слова — гейт детектора; роли — числовые константы в акторе; renderStatus-тест в ctlsvc-корпусе.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking lint] Константа autoRoleTimeout потребовала keepalive в Task 1**
- **Found during:** Task 1 (GREEN — mise run ci)
- **Issue:** план требует const autoRoleTimeout в actor.go уже на закрытии Task 1 (verify-греп), но её потребитель (ctx живого вызова) появляется только в Task 2 — strict-линт (unused) не пропускает неиспользуемую константу, зелёная итерация D-08 — жёсткий гейт
- **Fix:** `var _ = autoRoleTimeout` keepalive с комментарием в Task-1 GREEN, заменён реальным использованием ctx в Task-2 GREEN
- **Files modified:** internal/session/actor.go
- **Verification:** mise run ci зелёный на обоих закрытиях; греп-гейт Task 1 (autoRoleTimeout присутствует) выполнен
- **Committed in:** bd620f8 (keepalive), 9f92c8c (реальное использование)

**2. [Rule 1 - Bug] Тестовое зеркало Task-1-корпуса под ролевым гейтом**
- **Found during:** Task 2 (GREEN — полный корпус)
- **Issue:** после посадки ролевого гейта fires-путь требует двой роли: fireAutocorrect/ResetKeyBoundary не имели его — TestActor_ResetKeyBoundary краснел («boundary did not arm»: confirm без роли молчит)
- **Fix:** fireAutocorrect инсталлирует allowed-role double (форма «fires»), ResetKeyBoundary — явно; Task-1-ассерты не изменились
- **Files modified:** internal/session/actor_test.go
- **Verification:** весь session-корпус зелёный под -race
- **Committed in:** 9f92c8c

**3. [Rule 1 - Bug] Фикстуры Task-3 корпуса: включённый слой и носитель клона**
- **Found during:** Task 3 (GREEN)
- **Issue:** TestStatus_AutocorrectFields не включал слой (SetOptions пропущен — armed-полл висел); clone-проба писала в nil-карту (maps.Clone(nil) = nil после чистого fired-раунда); acEnabledCfg наследовал FlipAfterCorrection из встроенных дефолтов — тест ловил чужой флип режима
- **Fix:** SetOptions(acOptions()) в fields-тесте; clone-проба после одной абстиненции (карта несёт статью); acEnabledCfg пинит FlipAfterCorrection=false с комментарием (счётчики, не mode-правило)
- **Files modified:** internal/session/actor_test.go
- **Verification:** весь корпус зелёный; TestStatus_AutocorrectFields проверяет клон на реальной статье
- **Committed in:** 302542b

---

**Total deviations:** 3 auto-fixed (1 Rule 3, 2 Rule 1). **Impact on plan:** все — согласование с собственными behavior-блоками и жёсткими гейтами (lint/зелёная итерация); скоуп-крипа нет, производственный контур соответствует must_haves дословно.

## Issues Encountered

- Первый прогон Task-2 GREEN поймал баг самой фикстуры CountersAndReasons: `role.set(40)` мутировал double, ЗАМЕНЁННЫЙ fireAutocorrect'ом (свой `&fakeRole{61}`) — граница «для запрета» уходила в fired; фикс — явная переинсталляция запрета (строка в GREEN-коммите)
- Один транзиентный сбой `mise run ci` возник из моих же рук: два параллельных ci-прогона в одной команде контендили ctlsvc-сокет/шину; последующие одиночные прогоны (включая финальный /tmp/06-06-ci4.log) зелёные с нулём FAIL — известный флак ctlsvc не воспроизводился

## Known Stubs

None — реализация полная: ролевой шов, конъюнкция, счётчики, статус и фолд живут; заглушек «not available/TODO» нет.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- 06-07 (e2e) получает всё необходимое: счётчики-оракул (autocorrect_fired/abstained + ac_skip_* в goswitchctl status), fired-INFO-класс (`"msg":"autocorrect","reason":"fired"`), конфиг-секция с живым фолдом; e2e-кейсы fires/password-silent/terminal-silent садятся на готовые поверхности без правок актора
- 06-08 (матрица/приёмка): perf-сравнение p95 до/после (Pitfall 5) — решение по-прежнему только на границе слова, GetRole под 25 мс вне мьютекса; весь юнит-регресс зелёный на финальном дереве
- Границы контракта для e2e: off-состояние = ноль счётчиков и ноль записей; fires — через ПРОБЕЛ (reset-граница вооружает до HardReset — живой эффект не несущая часть приёмки, план-допущение); silence-matrix слаги — замкнутый словарь для оракула

---
*Phase: 06-avtokorrekcija-opcionalno*
*Completed: 2026-10-01*

## Self-Check: PASSED

- Файлы: buffer.go/buffer_test.go, actor.go/actor_test.go, ctlsvc.go/ctlsvc_test.go, 3 red-evidence JSON — существуют на диске (9/9 FOUND)
- Коммиты: 3009559, bd620f8, e543603, 9f92c8c, 9102bac, 302542b, e33893a — присутствуют в git log; measured commits от ledger gsd-plan-head-before-06-06: 7
- TDD-гейт: test(06-06) `3009559`/`e543603`/`9102bac` предшествуют своим feat(06-06); все три RED-записи — RED_EVIDENCE_OK по check tdd-red-evidence
- План-level verification перезапущена на финальном дереве: `go test ./internal/correct/ ./internal/session/ ./internal/ctlsvc/ -race -count=1` ok; `mise run ci` exit 0 (без FAIL); греп-гейты HOOK-OK / SINGLE-PATH-OK (CommitText: 5==5) / CLI-WIRING-FROZEN
