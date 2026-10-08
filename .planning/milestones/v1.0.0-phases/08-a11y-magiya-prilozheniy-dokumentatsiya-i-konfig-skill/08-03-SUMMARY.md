---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
plan: 03
subsystem: a11y
tags: [a11y, reconciler, gsettings, toolkit-accessibility, org-a11y-status, godbus, dbus, argv-discipline, warn-once, fire-and-forget, tdd]

# Dependency graph
requires:
  - phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill (план 08-01)
    provides: "SPEC §4 ревизия 2026-10-05 — записанный контракт: две ручки, read-verify-then-set, снятие запрещено, пояс без снапшота"
  - phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill (план 08-02)
    provides: "Секция A11y в internal/config + (a A11y) Active() — источник желаемого состояния для фолда 08-05"
  - phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
    provides: "sound.go канон (package-doc, warn-once/closeEpisode, no-pipes fork shape, reaped goroutine) + clipboard.go Runner seam (deadline-bounded чтение)"
provides:
  - "Пакет internal/a11y: type Reconciler + New(runner, status) + (*Reconciler).Apply(active bool) — fire-and-forget серия под собственным мьютексом, вне акторного (WR-01); diff-gate lastApplied делает повторные Apply и устаревшие серии no-op'ами"
  - "Серия применения: read-verify-then-set org.gnome.desktop.interface toolkit-accessibility (ключ уже true — ноль записей, dconf-чурн запрещён); Apply(false) — ноль подпроцессов (D-8-4); отказ чтения — fail toward desired"
  - "Пояс org.a11y.Status.IsEnabled=true при КАЖДОЙ активации-эпизоде через seam StatusSetter; session-scoped, без снапшота и revert-долга (A5, Pitfall 2)"
  - "ЭКСПОРТИРОВАННЫЕ конструкторы NewExecRunner() Runner и NewDBusStatusSetter() StatusSetter — имена-контракт для wiring 08-05 (правки internal/a11y планом 08-05 не предполагаются)"
  - "Argv-дисциплина T-08-03-01: argv обеих форм рендерится ТОЛЬКО из package-литералов; сигнатура Apply(active bool) не принимает строки — списку приложений нет пути в argv и в логи (D-20/D-21)"
  - "Degradation-дисциплина sound.go-канона: закрытый reason-словарь (key-read-failed/key-write-failed/belt-set-failed), ОДИН WARN на reason за эпизод, чистая активация переоткрывает бюджет"
  - "Корпус 12 тестов: серия/уже-true/деактивация/литеральный argv/неблокируемость+сходимость/no-pipes/пояс на каждой активации/молчание пояса на деактивации/бюджет WARN'ов/fail-toward-desired"
affects: [08-04 (snapshot/revert uninstall), 08-05 (wiring: фолд зовёт sink.Apply(snap.A11y.Active()) с NewExecRunner()+NewDBusStatusSetter()), 08-06 (доки описывают семантику)]

# Actuals (#2632) — pairs with the plan's estimate to calibrate future estimates.
actuals:
  tokens: 8778    # chars/4 over the realized diff (35113 chars: a11y.go + a11y_test.go + 2 red-evidence JSON)
  tasks: 2
  commits: 4      # MEASURED: git rev-list --count gsd-plan-head-before-08-03..HEAD
plan_head_before: c1f0c0d43731bad56cd6d0868305167bc0a1b4c4

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Reconciler-канон: Apply фиксирует желаемое под своим mu и spawn'ит серию (fire-and-forget); вся серия держит mu пакета — односторонний порядок актор.mu → a11y.mu; stale-серия под дифф-гейтом — no-op"
    - "Две формы подпроцессов в одном Runner-сиве: get — CommandContext с таймаутом пакета и захваченным stdout (clipboard execRunner), set — no-pipes exec.Command + reaped goroutine (sound.go newPlayProc:192)"
    - "Belt-адаптер: один godbus Properties.Set на сессионной шине (org.a11y.Bus → org.a11y.Status.IsEnabled), коннект на один вызов — эпизоды пользовательского темпа"

key-files:
  created:
    - internal/a11y/a11y.go
    - internal/a11y/a11y_test.go
    - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/red-evidence/08-03-task1-red.json
    - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/red-evidence/08-03-task2-red.json

key-decisions:
  - "Пояс-вызов landed в Task 2 по его GREEN-пункту (1) «вызов StatusSetter в серии активации» — Task 2 RED пал честно на belt-утверждениях (RED_EVIDENCE_OK), green-by-design не потребовался; Task-1-корпус пояс не пинит, Task-1 done-критерий пояс не называет"
  - "closeEpisode — только после ПОЛНОСТЬЮ чистой активации (sound.go:320-325 канон): отказ пояса не проваливает эпизод (lastApplied=true), но эпизод остаётся открытым — повторные отказы молчат до успешной активации; оба бюджет-теста зелёные в этой форме"
  - "read() несёт собственный контекст-таймаут пакета (1500 мс, форма clipboard call()), set — без дедлайна (nothing kills a set), пояс — таймаут внутри адаптера (beltTimeout 1500 мс); серия передаёт context.Background() — fire-and-forget отвязан от вызывающего"

patterns-established:
  - "internal/a11y Reconciler — канон best-effort sink'а с желаемым состоянием (расширение sound.go-дисциплины: diff-gate + сериализация серии + fail-toward-desired)"
  - "Belt-корпус: запись WARN'ов через recordSink (slog-хендлер-двойка, прецедент watch_test) — warn-once бюджеты тестируемы без реального стола"

requirements-completed: ["D-8-3", "D-8-6"]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Серия применения: Apply(true) при ключе false — ровно один get + один set с литеральным argv; ключ уже true — ноль set (анти-dconf-чурн, Pitfall 4); Apply неблокирующ (серия переживает заблокированный get) и последовательность Apply(true)→Apply(false)→Apply(true) сходится под -race"
    requirement: "D-8-3"
    verification:
      - kind: unit
        ref: "internal/a11y/a11y_test.go#TestA11y_ActivateSetsKeyWhenFalse"
        status: pass
      - kind: unit
        ref: "internal/a11y/a11y_test.go#TestA11y_AlreadyTrueNoSet"
        status: pass
      - kind: unit
        ref: "internal/a11y/a11y_test.go#TestA11y_ApplyDoesNotBlockAndConverges"
        status: pass
    human_judgment: false
  - id: D2
    description: "Argv-дисциплина и формы подпроцессов: каждый argv из package-литералов (список приложений не входит в seam вообще), diff-gate держит повторную Apply(true) на нуле вызовов, set — no-pipes форма с фиксированным argv (T-08-03-01/02/04/05)"
    requirement: "D-8-6"
    verification:
      - kind: unit
        ref: "internal/a11y/a11y_test.go#TestA11y_FixedArgvOnlyLiterals"
        status: pass
      - kind: unit
        ref: "internal/a11y/a11y_test.go#TestA11yWireLiterals"
        status: pass
      - kind: unit
        ref: "internal/a11y/a11y_test.go#TestA11y_SetCmdNoPipesForm"
        status: pass
    human_judgment: false
  - id: D3
    description: "Деактивация ноль-касаний (D-8-4): Apply(false) — ноль вызовов runner (ни get, ни set) и ноль вызовов пояса, на свежем reconciler'е и после активации"
    requirement: "D-8-3"
    verification:
      - kind: unit
        ref: "internal/a11y/a11y_test.go#TestA11y_DeactivateNeverTouchesKey"
        status: pass
      - kind: unit
        ref: "internal/a11y/a11y_test.go#TestA11y_BeltNotCalledOnDeactivation"
        status: pass
    human_judgment: false
  - id: D4
    description: "Пояс org.a11y.Status.IsEnabled + дисциплина эпизодов: пояс при каждой активации (уже-true и ре-активация включительно), молчание на деактивации; отказы (пояс/чтение) — один WARN на reason, эпизод жив (set выполняется, fail toward desired), успешная активация переоткрывает бюджет (closeEpisode-форма)"
    requirement: "D-8-6"
    verification:
      - kind: unit
        ref: "internal/a11y/a11y_test.go#TestA11y_BeltSetOnEveryActivation"
        status: pass
      - kind: unit
        ref: "internal/a11y/a11y_test.go#TestA11y_BeltFailureWarnOnce"
        status: pass
      - kind: unit
        ref: "internal/a11y/a11y_test.go#TestA11y_ReadFailureWarnOnceAndFailsTowardDesired"
        status: pass
      - kind: unit
        ref: "internal/a11y/a11y_test.go#TestA11y_SuccessfulEpisodeReopensBudget"
        status: pass
      - kind: other
        ref: "grep-гейт плана: org.a11y.Bus / org.a11y.Status / IsEnabled / godbus/dbus/v5 в a11y.go → BELT-OK"
        status: pass
    human_judgment: false
  - id: D5
    description: "Продакшн-адаптеры с запиненными именами-контрактом 08-05: NewExecRunner() (get = CommandContext с таймаутом и захваченным stdout, set = no-pipes + reaped goroutine) и NewDBusStatusSetter() (godbus Properties.Set на org.a11y.Bus → org.a11y.Status.IsEnabled=true)"
    requirement: "D-8-6"
    verification:
      - kind: other
        ref: "grep-гейт плана: org.gnome.desktop.interface / toolkit-accessibility / func (r *Reconciler) Apply / go func → RECONCILER-OK"
        status: pass
      - kind: other
        ref: "mise exec -- go vet ./internal/a11y/ && mise exec -- go build ./... → BUILD-VET-OK"
        status: pass
    human_judgment: false
  - id: D6
    description: "Живая итерация: mise run ci (build + vet + lint strict + test -race ./... + tidy-diff) exit 0 на выходе плана; go.sum не меняется (godbus уже в go.mod)"
    verification:
      - kind: other
        ref: "mise run ci → CI_EXIT=0 (прогон /tmp/08-03-ci-run2.log, internal/a11y ok)"
        status: pass
    human_judgment: false
  - id: D7
    description: "Фактический эффект на живом столе: реальный gsettings ставит ключ, реальная сессионная шина принимает Properties.Set, приложения из списка подхватывают a11y-дерево при следующем запуске"
    verification: []
    human_judgment: true
    rationale: "Юнит-корпус доказывает дисциплину через seam-фейки; живой эффект требует реальной GNOME-сессии и wiring'а 08-05 (демон ещё не зовёт reconciler) — владелец/verify-work проверит после 08-05 на живом столе"

# Metrics
duration: 27min
completed: 2026-10-05
status: complete
---

# Phase 8 Plan 03: Reconciler a11y-магии Summary

**Пакет internal/a11y: fire-and-forget reconciler (read-verify-then-set toolkit-accessibility + пояс org.a11y.Status.IsEnabled на каждой активации, снятие запрещено, argv из одних литералов, warn-once эпизоды) с экспортированными NewExecRunner()/NewDBusStatusSetter() — контрактом wiring'а 08-05.**

## Performance

- **Duration:** 27 min
- **Started:** 2026-10-05T16:58:01Z
- **Completed:** 2026-10-05T17:25:34Z
- **Tasks:** 2 (4 TDD-коммита: RED→GREEN ×2)
- **Files modified:** 2 кода (+882 строки) + 2 red-evidence JSON

## Accomplishments

- Новый пакет internal/a11y (D-8-3/D-8-6): `Reconciler.Apply(active bool)` фиксирует желаемое под своим мьютексом и spawn'ит серию — вызов никогда не блокирует фолд и не возвращает ошибку; вся серия под `mu` пакета (вне акторного, WR-01), дифф-гейт `lastApplied` делает повторные Apply и устаревшие серии no-op'ами, последовательность Apply(true)→Apply(false)→Apply(true) сходится к последнему состоянию (пин под -race)
- Read-verify-then-set глобального ключа `org.gnome.desktop.interface toolkit-accessibility` (research Pitfall 4): get через deadline-bounded Runner с захваченным stdout → trim → set только при расхождении; отказ чтения — один WARN и set всё равно выполняется (fail toward desired); Apply(false) — ноль подпроцессов (D-8-4)
- Пояс `org.a11y.Status.IsEnabled=true` (резолюция research «ship the belt») при КАЖДОЙ активации-эпизоде — «уже true» тоже активация (session-scoped, A5); production-адаптер `NewDBusStatusSetter()` — один godbus Properties.Set на сессионной шине, без root, без снапшота и revert-долга (Pitfall 2)
- Argv-дисциплина T-08-03-01: argv рендерится ТОЛЬКО из package-литералов (тест пинит и значения литералов, и членство каждого элемента), сигнатура Apply не принимает строки — списку приложений нет пути ни в argv, ни в логи (D-20/D-21); set — no-pipes fork shape с reaped goroutine (T-08-03-05)
- Degradation-дисциплина sound.go-канона: закрытый reason-словарь (key-read-failed / key-write-failed / belt-set-failed), один WARN на reason за эпизод, чистая активация переоткрывает бюджет (closeEpisode)
- Полный TDD-цикл обоих задач: RED `d9f55c3` (5/7 падений, RED_EVIDENCE_OK) → GREEN `8b72eec`; RED `db701ec` (4/12 падений, RED_EVIDENCE_OK) → GREEN `9a80c19`; `mise run ci` exit 0, go.sum не тронут

## Task Commits

Each task was committed atomically:

1. **Task 1 (трейсер): серия применения** — RED `d9f55c3` (test), GREEN `8b72eec` (feat); REFACTOR не потребовался (реализация легла в каноническую форму сразу)
2. **Task 2: пояс + дисциплина эпизодов + зелёная итерация** — RED `db701ec` (test), GREEN `9a80c19` (feat; включает lint-фиксы lll/goconst/gocritic/gofmt/nolintlint)

**Plan metadata:** следует отдельным docs-коммитом (SUMMARY + STATE + ROADMAP + REQUIREMENTS).

_Note: TDD-гейты: test(08-03) `d9f55c3` → feat(08-03) `8b72eec`; test(08-03) `db701ec` → feat(08-03) `9a80c19`._

## Files Created/Modified

- `internal/a11y/a11y.go` — package-doc по канону sound.go; литералы ключа/схемы/глаголов; Runner/StatusSetter seam'ы; Reconciler (desired/lastApplied/warned под своим mu); Apply/run/episode/read/setKey/belt; warn/closeEpisode; newSetCmd (no-pipes); NewExecRunner(); NewDBusStatusSetter()
- `internal/a11y/a11y_test.go` — корпус 12 тестов над fakes (recording runner с блокирующим get, recording belt, recordSink slog-хендлер): серия/уже-true/деактивация/литералы/неблокируемость+сходимость/no-pipes/пояс×3+бюджет/fail-toward-desired
- `red-evidence/08-03-task1-red.json`, `red-evidence/08-03-task2-red.json` — RED-свидетельства (оба RED_EVIDENCE_OK)

## Decisions Made

- **Пояс-вызов — в Task 2 GREEN** (его action-пункт 1 «вызов StatusSetter в серии активации»): Task-1-корпус пояс не пинит, поэтому Task 2 RED пал честно на belt-утверждениях (ровно как план предсказывал: «пояс ещё не вызывается / бюджет не переоткрывается») — оба RED_свидетельства валидированы verb'ом, green-by-design не понадобился
- **closeEpisode — только после полностью чистой активации**: отказ пояса не проваливает эпизод (lastApplied=true, серия жива), но бюджет не переоткрывает — повторные отказы молчат до успешной активации; форма sound.go:320-325, оба бюджет-кейса зелёные
- **Контексты по формам**: read — таймаут пакета 1500 мс (clipboard call()-форма), set — без дедлайна (nothing kills a set, sound.go), пояс — beltTimeout 1500 мс внутри адаптера; серия передаёт context.Background() — fire-and-forget отвязан от вызывающего
- **Reconciler владеет status-поля с nil-гвардом**: nil StatusSetter = пояс-заглушка «успех» (wiring 08-05 передаёт настоящий адаптер)

## Deviations from Plan

**1. [Rule 3 - Blocking] lint-гейт `mise run ci`: 10 замечаний строгого конфига после Task 2 GREEN**
- **Found during:** Task 2 (зелёная итерация — verify `<automated>mise run ci</automated>` красный)
- **Issue:** lll ×6 (длинные nolint-комментарии, однострочный set-argv, два длинных t.Errorf), goconst ×1 (литерал "false\n" ×8), gocritic rangeValCopy ×1 (копия slog.Record ~288 байт в range), gofmt ×1, nolintlint ×2 (неиспользуемые //nolint:ireturn на NewExecRunner/NewDBusStatusSetter — ireturn эти возвращаемые типы не флагует)
- **Fix:** разнесены длинные строки (argv в переменную, joined в переменную), литералы ответов в константы answerFalse/answerTrue, range по индексу, gofmt, лишние nolint сняты
- **Files modified:** internal/a11y/a11y.go, internal/a11y/a11y_test.go
- **Verification:** mise run lint → 0 issues; mise run ci → exit 0
- **Committed in:** 9a80c19 (коммит Task 2 GREEN)

---

**Total deviations:** 1 auto-fixed (1 blocking — lint-гейт зелёной итерации).
**Impact on plan:** формат и формы без изменения поведения; ни один behavior-кейс не изменился.

## TDD Gate Compliance

- Task 1: RED `test(08-03) d9f55c3` предшествует GREEN `feat(08-03) 8b72eec`; RED-свидетельство `RED_EVIDENCE_OK` (5 целевых тестов пали на сходимости серии против RED-стаба-формы по прецеденту 06-04/08-02; raw /tmp/08-03-t1-red-raw.txt, запись red-evidence/08-03-task1-red.json). REFACTOR — без изменений, коммита нет.
- Task 2: RED `test(08-03) db701ec` предшествует GREEN `feat(08-03) 9a80c19`; RED-свидетельство `RED_EVIDENCE_OK` (4 целевых belt-теста пали на «пояс ещё не вызывается»; TestA11y_BeltNotCalledOnDeactivation green-by-design — отрицательное утверждение тривиально истинно при неподключённом поясе, честно записано; raw /tmp/08-03-t2-red-raw.txt, запись red-evidence/08-03-task2-red.json). REFACTOR — без отдельного коммита (lint-фиксы вошли в GREEN по прецеденту 08-02).

## Issues Encountered

- Первый полный прогон `mise run ci` после lint-фиксов упал на одном тесте вне internal/a11y (в выводе сохранился только хвост — падение не в пакете плана, корпус a11y в соседних прогонах зелёный); немедленный повторный `mise run test` и полный `mise run ci` зелёные (exit 0, лог /tmp/08-03-ci-run2.log). По заметкам плана перезапущен однократно и зафиксирован здесь: транзиент, подозреваемые — известные pre-existing флаки из deferred-items.md (ctlsvc TestRun_OnConnHookCalledOnce, config TestWatch_BrokenAutocorrectBlockKeepsLastGood), оба вне зоны плана.
- Гейт `requirements.ready-ids`: D-8-3 и D-8-6 объявлены также планами 08-05/08-06 без SUMMARY — shared-ID-гейт #2388 блокирует mark-complete до их завершения; IDs проставятся последним завершившим планом волны.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `internal/a11y` самодостаточен: `New(a11y.NewExecRunner(), a11y.NewDBusStatusSetter())` + `sink.Apply(snap.A11y.Active())` — полный контракт wiring'а 08-05 (имена конструкторов запинены, правки пакета планом 08-05 не предполагаются); snapshot/revert uninstall — 08-04; строки a11y.* в docs/CONFIG.md — 08-06
- Живой эффект на столе (постановка ключа реальным gsettings, приём Properties.Set шиной, рестарт-семантика приложений) — предмет UAT после 08-05 (coverage D7, human_judgment)
- Открытых блокеров нет

## Self-Check: PASSED

- internal/a11y/a11y.go, internal/a11y/a11y_test.go, red-evidence/08-03-task1-red.json, red-evidence/08-03-task2-red.json — FOUND (в дифе плана, 882+)
- Commits d9f55c3, 8b72eec, db701ec, 9a80c19 — FOUND in git log; ledger gsd-plan-head-before-08-03 → 4 коммита
- Гейт RECONCILER-OK (toolkit-accessibility/Apply/go func) и BELT-OK (org.a11y.Bus/Status/IsEnabled/godbus) — FOUND
- go.sum — git diff пуст (godbus уже в go.mod, ноль новых пакетов)
- mise run ci — exit 0 (финальный прогон 2026-10-05T17:2xZ, internal/a11y ok 1.140s)

---
*Phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill*
*Completed: 2026-10-05*
