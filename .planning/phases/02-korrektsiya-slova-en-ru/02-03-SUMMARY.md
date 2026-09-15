---
phase: 02-korrektsiya-slova-en-ru
plan: "03"
subsystem: correct
tags: [ibus, dbus, tdd, tracer, layout-correction, surrounding-text, zenity, e2e]

requires:
  - phase: 02-korrektsiya-slova-en-ru
    provides: internal/correct (Buffer/Detect/Convert/MatchesSuffix/BuildPlan — формулы 02-01, использованы как есть)
  - phase: 02-korrektsiya-slova-en-ru
    provides: e2e-стенд (zenity-драйвер case_m1.go, AT-SPI witness, фокус-восстановление 02-02)
provides:
  - engine.Emitter (RequireSurroundingText/DeleteSurroundingText/CommitText) — исходящие сигналы лестницы ADR-003; EventHandler расширен HandleSurroundingText/HandleCapabilities/AttachEngine; factory.CreateEngine привязывает sink
  - Исправленный wire-тип AttrList.Attributes = av (дневной дефект Фазы 1 ронял ibus-daemon 1.5.29 SEGV)
  - internal/session.Actor — двухфазная коррекция CORR-01: кормление буфера (лэтч-модификаторы разрешены, комбо Ctrl/Alt/Super нет), Double → Token/Detect/Convert → ADR-004-сверка (кэш спонтанных пушей + Require-раунд ≤100мс таймером) → DeleteSurroundingText(token+tail) + CommitText(converted+tail), ReplaceToken-тоггл; D-20/D-21 лог-контракты
  - Живые кейсы реестра word-en-ru и word-after-space + mise-задачи e2e-word / e2e-word-space
  - Трассировка живого транспорта: GTK/mutter НИКОГДА не отвечают RequireSurroundingText (только спонтанные пуши) — документировано в коде и здесь
affects: [02-04 (флип-режим: Solution Deferred готово), 02-05 (уровень 2: BuildPlan уже считает, no-surrounding-отказ — точка подключения), 02-06 (матрица: лог-контракты "level":1/«correction skipped» — грепабельны)]

actuals:
  tokens: 17628   # chars/4 over the realized diff dd461b8..HEAD (estimate was 28000)
  tasks: 2
  commits: 5      # MEASURED: git rev-list --count dd461b8..HEAD
  plan_head_before: dd461b88dc35260f8d58224fd9fd8b971c2f9eca

tech-stack:
  added: []
  patterns:
    - "godbus WithOutgoingInterceptor — headless-пин wire-формы исходящих сигналов (in-package seam, server-side SASL у godbus нет)"
    - "decodeIBusText: godbus декодирует STRUCT в []any — типизированная структура на входящем пути невозможна, разварьирование позиционное и защитное"
    - "Сверка ADR-004 по свежести данных: кэш последнего спонтанного SetSurroundingText проверяется первым, Require-раунд — фолбэк в таймере 100 мс"
    - "Оракул точного содержимого не проходит через runCmd (TrimSpace съедает значимый хвостовой пробел)"

key-files:
  created:
    - engine/emitter_wire_test.go
    - test/e2e/case_word.go
    - .planning/phases/02-korrektsiya-slova-en-ru/red-evidence/02-03-task1-red.json
  modified:
    - engine/engine.go
    - engine/types.go
    - engine/factory.go
    - engine/engine_test.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - test/e2e/case_m1.go
    - test/e2e/main.go
    - test/e2e/README.md
    - mise.toml

key-decisions:
  - "AttachEngine принимает engine.Emitter (интерфейс швa), а не буквальный *Engine плана: план одновременно требует fake-sink тесты актёра — конкретный тип сделал бы двойник невозможным; интерфейс объявлен в точке-потребителе шва"
  - "Верификация сверяет суффикс token+tail (весь диапазон замены), не токен: буквальный pending.token плана Task 1 противоречит его же D-13 пину Task 2 («abc ghbdtn », 11 → коррекция исполняется); Task 2 — контракт, 02-01 разрешил аналогичное противоречие так же"
  - "Кэш спонтанных surrounding-пушей — первичный вход сверки: живьём доказано (gdbus-монитор шины + строки бинарников), что ни GTK4/GTK3, ни mutter не имеют обработчика require-surrounding-text; прототип владельца работает тем же кэшем; Require-раунд сохранён в таймере 100 мс для клиентов, которые его реализуют — инвариант ADR-004 (без сверки не заменять, несовпадение/таймаут → тихий отказ) не нарушен"
  - "кормление буфера: Mods&comboMask==0 (Ctrl/Alt/Super запрещены), лэтчи (NumLock/CapsLock) разрешены — keyval уже XKB-переведённый символ; живая находка: все буквы приходили с mods 0x10 (NumLock), буквальный гард плана морил буфер голодом"
  - "AttrList.Attributes = []dbus.Variant (av): дневной 'au' триггерил format-assertion ibus-daemon 1.5.29 и убивал его SEGV (journal, 2026-09-14); wire-тест пинит сигнатуры (sa{sv}sv)/(sa{sv}av)"
  - "Разделитель в живом кейсе вводится injectText(\" \"): ydotool 0.1.8 резолвит имя \"space\" в keycode 31 = физическая S (двигок видел keyval 0x73, коррекция дала приветЫ)"

patterns-established:
  - "Живой трассер вскрывает wire-дефекты, недоказуемые headless: AttrList-SEGV нашёл именно сквозной прогон"
  - "Диагностика оракула через параллельное чтение поля из шелла при падающем прогоне стенда — быстрый способ отделить баг поля от бага читателя"

requirements-completed: [CORR-01, CORR-07]

coverage:
  - id: D1
    description: "Трассер CORR-01: двойной тап Right Shift запускает двухфазную коррекцию в актёре (буфер→направление→конвертация→сверка→DeleteSurroundingText+CommitText)"
    requirement: CORR-01
    verification:
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_DoubleTapCorrects|TestActor_CachedSurroundingCorrects|TestActor_LatchedModsStillFeed
        status: pass
      - kind: e2e
        ref: "mise run e2e-word → PASS word-en-ru (live, 2026-09-15, оракул stdout «привет» + witness settle 6)"
        status: pass
    human_judgment: false
  - id: D2
    description: "CORR-07 уровень 1: замена ровно по диапазону токен+хвост — ровно один DeleteSurroundingText(-6,6)/(-7,7) и один CommitText(converted+tail), без прыжка курсора"
    requirement: CORR-07
    verification:
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_AfterSpaceCorrects (ровно один (-7,7)+CommitText("привет "))
        status: pass
      - kind: unit
        ref: internal/correct/plan_test.go#TestBuildPlan_BothLevels (формулы 02-01, актёр их вызывает)
        status: pass
      - kind: e2e
        ref: "mise run e2e-word-space → PASS word-after-space (content-exact AT-SPI оракул «привет » с пробелом)"
        status: pass
    human_judgment: false
  - id: D3
    description: "ADR-004 abort-дисциплина: mismatch/timeout → тихий отказ с INFO-причиной, НИ ОДНОГО разрушительного вызова; late-текст не воскрешает коррекцию"
    verification:
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_VerifyPaths (mismatch/timeout/late-text) + TestActor_EmptyBufferNoDestructive + TestActor_TokenRefusals
        status: pass
    human_judgment: false
  - id: D4
    description: "D-20/D-21 приватность логов: INFO — только outcome/reason без содержимого слов; DEBUG — level первым атрибутом после msg, с source/result/latency"
    verification:
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_DebugCorrectionRecord (сканирует каждую INFO-запись на отсутствие слов)
        status: pass
      - kind: e2e
        ref: "живые логи: {\"msg\":\"correction\",\"outcome\":\"done\"} и {\"msg\":\"correction\",\"level\":1,...} в /tmp word-run логах"
        status: pass
    human_judgment: false
  - id: D5
    description: "engine.Emitter-эмиттеры и wire-форма сигналов (Delete/Require симметричны CommitText; AttrList av)"
    verification:
      - kind: unit
        ref: engine/emitter_wire_test.go#TestEmitters_DeleteAndRequire (интерсептор godbus, сигнатуры и позиционные аргументы) + engine/engine_test.go#TestEngine_SurroundingTextForward
        status: pass
    human_judgment: false
  - id: D6
    description: "Toggle после коррекции: повторный двойной тап конвертирует обратно (ReplaceToken-инвариант 02-01 в конвейере)"
    verification:
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_ToggleRepeat (две коррекции, CommitText привет→ghbdtn)
        status: pass
    human_judgment: false
  - id: D7
    description: "Тоггл-стойкость стенда: zenity-ожидание переживает focus-stealing denial активного стола; обе mise-задачи живые и задокументированы"
    verification:
      - kind: e2e
        ref: "m1-gate PASS 3× подряд после hardening; e2e-word/e2e-word-space PASS; README обновлён"
        status: pass
    human_judgment: false

duration: 86min
completed: 2026-09-14
status: complete
---

# Phase 2 Plan 03: Трассер коррекции слова EN↔RU Summary

**Двойной Right Shift исправляет ghbdtn→привет в живом zenity сквозным конвейером (буфер→направление→сверка ADR-004→DeleteSurroundingText+CommitText) — трассер вскрыл и закрыл три транспортных дефекта, включая SEGV ibus-daemon от дневного AttrList-wire-типа; word-after-space доказывает геометрию D-13 с сохранённым пробелом**

## Performance

- **Duration:** 86 min (19:58–21:25 UTC 2026-09-14/15, включая живую отладку на активном столе владельца)
- **Started:** 2026-09-14T19:58:23Z
- **Completed:** 2026-09-14T21:24:45Z
- **Tasks:** 2 / 2
- **Files modified:** 13

## Accomplishments
- CORR-01 живьём: `mise run e2e-word` — инжекция ghbdtn → тап-тап → DeleteSurroundingText+CommitText → оракул stdout «привет»; конвейер замкнут от evdev до поля ввода
- D-13 живьём: `mise run e2e-word-space` — `ghbdtn␠` → `привет␠` с content-exact оракулом (главный геометрический тест фазы, Pitfall 1 — баг хвоста прототипа — закрыт сквозным доказательством)
- Трассер выполнил своё назначение: первый живой CommitText вскрыл wire-дефект Фазы 1 (AttrList 'au' вместо 'av'), ронявший ibus-daemon SEGV — исправлен и запинен сигнатурным тестом
- Живая трассировка транспорта: GTK/mutter не отвечают RequireSurroundingText (только спонтанные пуши) — сверка ADR-004 перестроена вокруг кэша пушей с сохранением инварианта «без сверки не заменять»; Require-раунд остался фолбэком в таймере 100 мс
- D-20/D-21 соблюдены дословно: INFO-причины без слов (включая защитные no-engine/convert-failed вне словаря плана), DEBUG с level первым атрибутом

## Task Commits

Each task was committed atomically:

1. **Task 1 pre-fix: zenity focus-recovery** — `fc2d0ac` (fix)
2. **Task 1 RED: tracer-correction corpus** — `e0ff2de` (test)
3. **Task 1 GREEN: pipeline + live case** — `558add1` (feat)
4. **Task 2: word-after-space D-13** — `73be9b3` (feat)
5. **README: cases documentation** — `a7c3dc3` (docs)

**Plan metadata:** см. финальный docs-коммит.

## TDD Gate Compliance

| Task | RED | GREEN | REFACTOR | Evidence |
|------|-----|-------|----------|----------|
| 1 | `e0ff2de` (5 тестов падают на утверждениях, 14 прежних зелёные) | `558add1` | — не потребовался | `red-evidence/02-03-task1-red.json` → RED_EVIDENCE_OK (target_test_failed) |
| 2 | — не положен планом | `73be9b3` | — | оба behavior-кейса (AfterSpace/Toggle) зелёные на появлении — Task 1 уже нёс хвостовой путь; план оговаривает это явно («если всё зелёное — RED-коммит не нужен») |

Gate-последовательность `test(02-03)` → `feat(02-03)` проверена по git log.

## Files Created/Modified
- `internal/session/actor.go` — двухфазная коррекция: feeding (comboMask), startCorrection (кэш+Require), executeCorrection, VerifyExpiry, токен-отказы с D-20-причинами
- `internal/session/actor_test.go` — fakeSink, 7 новых тестов + TestActor_LatchedModsStillFeed (регрессия NumLock)
- `engine/engine.go` — Emitter-интерфейс, EventHandler+3 метода, эмиттеры Delete/Require, decodeIBusText, фуннелирование SetSurroundingText/SetCapabilities
- `engine/types.go` — AttrList.Attributes → []dbus.Variant (av)
- `engine/factory.go` — AttachEngine после привязки conn/path
- `engine/emitter_wire_test.go` — in-package wire-пины через WithOutgoingInterceptor (nolint:testpackage обоснован)
- `engine/engine_test.go` — recordingHandler расширен, TestEngine_SurroundingTextForward, TestEmitters_DetachedQuiet
- `test/e2e/case_word.go` — runWordENRU / runWordAfterSpace, waitZenityChars/Text, readFocusedTextRaw
- `test/e2e/case_m1.go` — waitZenityEntry: grab-реактивация (grace 2 c, каждые 2 c, бюджет 15 c)
- `test/e2e/main.go`, `mise.toml`, `test/e2e/README.md` — реестр, задачи, документация

## Decisions Made
- См. key-decisions в frontmatter — все продиктованы живыми находками трассера (монитор шины gdbus, journal SEGV, строки бинарников GTK/mutter, параллельные readback-пробы)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] AttrList wire-тип ронял ibus-daemon SEGV**
- **Found during:** Task 1 (первый живой CommitText)
- **Issue:** AttrList.Attributes был []uint32 ('au'); ibus-daemon 1.5.29 парсит атрибуты как "av" — format assertion → SEGV демона ввода всего стола (journal: status=11/SEGV, рестарт-счётчик 2). Дефект Фазы 1, headless-недоказуемый.
- **Fix:** Attributes []dbus.Variant; wire-тест пинит сигнатуры payload (sa{sv}sv)/(sa{sv}av)
- **Files modified:** engine/types.go, engine/emitter_wire_test.go
- **Verification:** PASS word-en-ru живьём; ibus-daemon жив; wire-тест зелёный
- **Committed in:** 558add1

**2. [Rule 1 - Bug] NumLock-лэтч морил буфер голодом**
- **Found during:** Task 1 (живой прогон: action n:2 → reason:"empty-buffer")
- **Issue:** план пиннул гард `Mods&^MaskShift == 0`; на столе владельца NumLock залочен — ВСЕ буквы приходят с mods 0x10, буфер не кормился
- **Fix:** comboMask = Control|Mod1|Mod4 (лэтчи разрешены — keyval уже XKB-переведён); регресс-пин TestActor_LatchedModsStillFeed
- **Files modified:** internal/session/actor.go, internal/session/actor_test.go
- **Verification:** живой PASS; юнит-регресс зелёный
- **Committed in:** 558add1

**3. [Rule 1 - Bug] godbus STRUCT декодируется в []any, не в типизированную структуру**
- **Found during:** Task 1 (surrounding_text приходит, сверка не разрешается)
- **Issue:** план предполагал разварьирование входящего варианта в IBusText; `text.Value().(IBusText)` не работает никогда (декодер godbus даёт []any позиционно)
- **Fix:** decodeIBusText — позиционное защитное разварьирование (comma-ok на каждом поле); тест TestEngine_SurroundingTextForward
- **Files modified:** engine/engine.go, engine/engine_test.go
- **Verification:** юнит + живой конвейер (пуши курсора 1..7 дошли до актёра)
- **Committed in:** 558add1

**4. [Rule 1 - Transport] Клиенты GTK/mutter не отвечают RequireSurroundingText**
- **Found during:** Task 1 (Require на шине (gdbus-монитор — корректный path/iface/member/args), ответа нет 100 мс и позже; в libgtk-3/4 и libmutter-14 нет ни одной require-surrounding строки)
- **Issue:** план-протокол «Require → свежий SetSurroundingText ≤100 мс» неисполним на этом стеке клиентов: они ПУШАЮТ surrounding спонтанно при каждом изменении (курсор 1..6 при наборе) — так работает и прототип владельца
- **Fix:** кэш последнего спонтанного пуша — первичный вход сверки (свежесть = после последней клавиши); Require-раунд сохранён как фолбэк в таймере verifyWait; сброс кэша на FocusOut/Reset; инвариант ADR-004 (mismatch/timeout → тихий отказ, ноль удалений) не тронут
- **Files modified:** internal/session/actor.go
- **Verification:** TestActor_CachedSurroundingCorrects + живые PASS обоих кейсов (latency_ms:0 — сверка из кэша)
- **Committed in:** 558add1

**5. [Rule 1 - Plan-internal] Диапазон сверки = токен+хвост, не токен**
- **Found during:** Task 1 design (при подготовке хвостового пути)
- **Issue:** буквальный pending.token плана Task 1 противоречит его же D-13 пину Task 2: текст «abc ghbdtn » (курсор после пробела) не ЗАКАНЧИВАЕТСЯ токеном «ghbdtn» — after-space-коррекция отказалась бы всегда
- **Fix:** pendingFix.match = token+tail (ровно диапазон замены лестницы) — TestActor_AfterSpaceCorrects исполняется, mismatch-кейсы не затронуты
- **Files modified:** internal/session/actor.go
- **Committed in:** 558add1

**6. [Rule 3 - Blocking] ydotool резолвит имя "space" в физическую S**
- **Found during:** Task 2 (поле после коррекции — «приветЫ»)
- **Issue:** `ydotool key space` шлёт keycode 31 (KEY_S): движок видел keyval 0x73, токен вырос на s, хвост исчез
- **Fix:** разделитель вводится проверенным путём печати injectText(" ") (keycode 57 подтверждён в логе)
- **Files modified:** test/e2e/case_word.go
- **Committed in:** 73be9b3

**7. [Rule 3 - Blocking] runCmd TrimSpace съедал хвостовой пробел оракула**
- **Found during:** Task 2 (chars=7, чтение «привет» 5 секунд подряд; параллельный шелл-проба того же поля — «привет »)
- **Issue:** generic runCmd обрезает surrounding-whitespace — значимый для D-13 хвостовой пробел исчезал в читателе, не в поле
- **Fix:** readFocusedTextRaw — без трима (только \n помощника)
- **Files modified:** test/e2e/case_word.go
- **Committed in:** 73be9b3

**8. [Rule 3 - Blocking] zenity-ожидание не переживает focus-stealing denial активного стола**
- **Found during:** precondition-прогон m1-gate (два FAIL подряд)
- **Issue:** mutter отказывает свежесмапленному GTK4-окну в фокусе при активности владельца (та же находка 02-02 для GTE); бюджет 5 с с одним-двумя poke недостаточен
- **Fix:** waitZenityEntry — pid-ключевой grab-input-pid poke каждые 2 c в бюджете 15 c (паттерн 02-02); m1-gate PASS 3× подряд
- **Files modified:** test/e2e/case_m1.go
- **Verification:** живые прогоны
- **Committed in:** fc2d0ac

**9. [Docs] README вне файловых списков задач**
- **Found during:** Task 2
- **Issue:** новые mise-задачи не отражены в README (конвенция 02-02)
- **Fix:** строки в списке прогона + таблица флагов
- **Files modified:** test/e2e/README.md
- **Committed in:** a7c3dc3

---

**Total deviations:** 9 auto-fixed (5 bug, 3 blocking, 1 docs)
**Impact on plan:** Все — назначение трассера: план предполагал транспорт, отличный от фактического; семантика плана (инварианты ADR-003/004, D-13..D-21) соблюдена строже буквы его транспортных допущений. Роста объёма нет.

## Issues Encountered
- SEGV ibus-daemon (отклонение 1) оставил zombie-владельца имени org.freedesktop.IBus.goswitch (имя в ListNames без owner; GetNameOwner — no such name); снят штатным `ibus restart` — наш демон пережил рестарт шины и перерегистрировался (generation 1 в логе — живое доказательство INTEG-04-механизма Фазы 1)
- Активность владельца за столом (вечер, Chrome с почтой) давала интермиттентные focus-отказы — устойчивость обеспечена отклонением 8; все живые прогоны завершены на разблокированном экране, human-action checkpoint не потребовался
- cursorPos в behavior-спецификации Task 1 плана указан 9 для «abc ghbdtn» (10 рун) — арифметическая описка; корпус использует самовычисляемую длину (runeLen)

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- 02-04 (флип): решение Single уже диспетчеризуется в ExpiryAt (slog.Debug "action deferred") — точка подключения готова; буфер кормится и в RU-режиме (script-true keyvals)
- 02-05 (уровень 2): BuildPlan вычисляет оба уровня; отказ no-surrounding — точка подключения ForwardKeyEvent-пути
- 02-06 (матрица): лог-контракты грепабельны дословно (`"msg":"correction","outcome":"done"`, `"reason":…`, `"level":1` первым атрибутом); движок word-кейсов переносится в YAML-шаги
- Осторожно: повторная активация демона при живом zombie-имени (после падений ibus-daemon) — single-instance guard честно падает; лечение штатное `ibus restart`

## Self-Check: PASSED

- Files: engine/emitter_wire_test.go, test/e2e/case_word.go, red-evidence/02-03-task1-red.json, 02-03-SUMMARY.md — FOUND
- Commits: fc2d0ac, e0ff2de, 558add1, 73be9b3, a7c3dc3 — FOUND в git log
- Plan verification re-run: mise run ci exit 0; mise run e2e-word PASS; mise run e2e-word-space PASS; юнит-корпус -race зелёный
