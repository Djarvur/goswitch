---
phase: 02-korrektsiya-slova-en-ru
plan: 04
subsystem: input-correction
tags: [ibus, layout-switching, script-flip, key-consumption, e2e, tdd]

requires:
  - phase: 02-03
    provides: tracer correction pipeline (engine emitters, EventHandler+Emitter extension, actor two-phase verify, script-true buffer feeding outside combos)
provides:
  - внутренний флип скрипта (ADR-001 Option B): одиночный Right Shift переключает режим движка EN↔RU на решении FSM Single, XKB-группа не меняется
  - условное потребление в ProcessKeyEvent (вердикт EventHandler — контракт observer-false Фазы 1 заменён)
  - RU-режим: потребление чистых печатных с коммитом кириллицы по layouts.ENToRU; немапленное и тождественное — транзит
  - script-true буфер во ВСЕХ ветвях печати (коммит движка и транзит клиента), включая цифры в RU-режиме
  - лог-контракт INFO {"msg":"mode","to":"ru"|"en"} — синхронизация стенда (Pitfall 5)
  - живые кейсы word-ru-en / word-mixed + mise-задачи e2e-word-ru / e2e-word-mixed
affects: [02-05, 02-06, 02-07, phase-03]

actuals:
  tokens: 9300   # chars/4 over the realized diff (c738b1f..43c0631: 625+/49- over 10 files)
  tasks: 2
  commits: 4
plan_head_before: c738b1f32080c46f615df8d513aec929a6a53ddb9

tech-stack:
  added: []
  patterns:
    - "потребление решает логика, транспорт возвращает: ProcessKeyEvent → handler.HandleKey(ev) verdict (условное потребление вместо всегда-false)"
    - "script-true кормление буфера ветка-за-веткой: руна в поле ⇒ руна в буфере, независимо от того, кто её вставил (коммит движка или транзит клиента)"
    - "мод-запись INFO как контракт синхронизации e2e: стенд ждёт {\"msg\":\"mode\",\"to\":…} перед следующим шагом (флип на истечении окна, не мгновенно)"

key-files:
  created:
    - .planning/phases/02-korrektsiya-slova-en-ru/red-evidence/02-04-task1-red.json
  modified:
    - internal/session/actor.go
    - internal/session/actor_test.go
    - engine/engine.go
    - engine/engine_test.go
    - cmd/goswitchd/main.go
    - test/e2e/case_word.go
    - test/e2e/main.go
    - mise.toml
    - test/e2e/README.md

key-decisions:
  - "comboMask (Ctrl/Alt/Super, лэтчи разрешены) управляет и потреблением RU — буквальная маска плана mods&^MaskShift==0 уморила бы буфер на NumLock-столе (живая находка 02-03: все буквы с mods 0x10)"
  - "Тождественная карта ('2'→'2') в RU: пин выбран транзитом (false, без коммита) — руна всё равно в буфере, инвариант script-true от выбора не зависит (пин TestActor_RUScriptTrueAllBranches)"
  - "TestActor_MixedWordUntouched зелёный сразу — Detect из 02-01 уже отказывал смешанному; по распоряжению плана RED-коммит не нужен, пин нулевых разрушительных вызовов закоммичен с живыми кейсами"

patterns-established:
  - "Flip-on-Single: решение FSM исполняется flipScript() на истечении окна, идемпотентный toggle, ноль вызовов на sink"
  - "feedKey: единая точка решения потребления с инвариантом script-true в каждой ветви печати"

requirements-completed: [CORR-01, CORR-04]

coverage:
  - id: D1
    description: "Флип-режим: одиночный Right Shift переключает EN↔RU на решении Single, INFO mode-запись, ноль внешних вызовов"
    requirement: CORR-04
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_FlipOnSingle"
        status: pass
      - kind: e2e
        ref: "mise run e2e-word-ru → PASS word-ru-en (live, 2026-09-15: flip gated on mode record)"
        status: pass
    human_judgment: false
  - id: D2
    description: "RU-потребление печатных с коммитом кириллицы по таблице (g→п, [→х, ?→,); немапленное/тождественное — транзит; Ctrl-комбо не потребляются и не кормят буфер; EN транзитен как в Фазе 1"
    requirement: CORR-04
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_RUConsumesPrintable (+RUTransitUnmapped, RUCtrlModifiedNotCommitted, ENTransitUnchanged)"
        status: pass
      - kind: unit
        ref: "engine/engine_test.go#TestEngine_ProcessKeyEventConsumePropagation"
        status: pass
    human_judgment: false
  - id: D3
    description: "Инвариант script-true в обеих ветвях печати: цифры в RU-режиме в буфере независимо от consume/транзита; полный пайплайн привет2026→ghbdtn2026 с диапазоном ровно 10 рун (не 14 байт)"
    requirement: CORR-04
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_RUScriptTrueAllBranches (+TestActor_RUDigitsFullPipeline, TestActor_ScriptTrueBuffer)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Оба направления исправляются живьём: привет→ghbdtn (режим RU) — кейс word-ru-en, оракул stdout ghbdtn"
    requirement: CORR-01
    verification:
      - kind: e2e
        ref: "mise run e2e-word-ru → PASS word-ru-en (live, 2026-09-15)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Смешанное слово gfbпривет не трогается вовсе: INFO-отказ reason:mixed-script, ноль DeleteSurroundingText/CommitText — юнит-пин и живой кейс word-mixed"
    requirement: CORR-04
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_MixedWordUntouched"
        status: pass
      - kind: e2e
        ref: "mise run e2e-word-mixed → PASS word-mixed (live, 2026-09-15, gfbпривет без изменений)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Регресс 02-03 не сломан: e2e-word и e2e-word-space остаются PASS после смены контракта потребления"
    verification:
      - kind: e2e
        ref: "mise run e2e-word && mise run e2e-word-space → PASS word-en-ru, PASS word-after-space (live, 2026-09-15)"
        status: pass
      - kind: automated
        ref: "mise run ci → exit 0 (build/vet/lint/test -race)"
        status: pass
    human_judgment: false

duration: 15 min
completed: 2026-09-14
status: complete
---

# Phase 2 Plan 4: Внутренний флип скрипта Summary

**Одиночный Right Shift переключает режим движка EN↔RU (ADR-001 Option B): в RU движок потребляет печатные и коммитит кириллицу по таблицам, буфер script-true во всех ветвях печати, оба направления коррекции и смешанное слово работают живьём.**

## Performance

- **Duration:** 15 min
- **Started:** 2026-09-14T21:31:25Z
- **Completed:** 2026-09-14T21:46:46Z
- **Tasks:** 2
- **Files modified:** 10

## TDD Gate Compliance

| Gate | Commit | Evidence |
|------|--------|----------|
| RED | `c5b5171` test(02-04) | RED_EVIDENCE_OK: TestActor_FlipOnSingle + 5 target tests fail on assertions (34 tests, 28 pass, 6 fail); record at red-evidence/02-04-task1-red.json |
| GREEN (Task 1) | `fbf4d1c` feat(02-04) | mise exec go test ./internal/session/ ./engine/ ./internal/correct/ -race — ok; mise run ci — exit 0 |
| Task 2 | `d691481` feat(02-04) | Mixed-word pin green on arrival (plan-authorized: no RED commit); live cases PASS |

REFACTOR: not needed — feedKey already isolates the branch-local invariant, printableKeyval exists; no behavior-neutral cleanup worth a commit.

## Accomplishments

- Флип-режим (Task 1): scriptMode (modeEN старт) + flipScript на решении FSM Single; INFO {"msg":"mode","to":"ru"|"en"} — контракт синхронизации стенда (Pitfall 5); флип — чистое состояние демона, ноль вызовов на sink, XKB-группа не участвует.
- Условное потребление: engine.ProcessKeyEvent возвращает вердикт EventHandler (контракт observer-false Фазы 1 осознанно заменён; паника-сдерживание INTEG-05 не тронуто и зелёное).
- RU-потребление по таблице: чистый печатный press → ENToRU → CommitText(кириллица) + consume=true ('g'→п, '['→х, '?'→,); тождественная карта и немапленное — транзит; Ctrl/Alt/Super-комбо и непечатные — транзит без кормления буфера.
- Инвариант script-true пинен для каждой ветви печати: руна в поле ⇒ руна в буфере — скоммиченная кириллица в RU-commit, исходная руна в транзите (включая цифры '2' в RU: привет2026 → диапазон удаления ровно 10 рун, коррекция ghbdtn2026).
- Живые кейсы (Task 2): word-ru-en — полный цикл флип→RU-набор (движок коммитит «привет»)→тап-тап→«ghbdtn», оракул stdout; word-mixed — gfbпривет нетронуто с записью mixed-script в логе. Оба PASS живьём с первого прогона; регресс e2e-word/e2e-word-space зелёный; mise run ci зелёный.

## Task Commits

1. **Task 1: Флип-режим — Single→переключение, потребление, script-true буфер** — `c5b5171` (test/RED) + `fbf4d1c` (feat/GREEN)
2. **Task 2: Живые кейсы word-ru-en / word-mixed** — `d691481` (feat)

**Plan metadata:** `43c0631` (docs: e2e README) + SUMMARY commit (docs: complete plan)

## Files Created/Modified

- `internal/session/actor.go` — scriptMode/flipScript/feedKey: флип на Single, потребление с коммитом по таблице, script-true кормление в каждой ветви
- `internal/session/actor_test.go` — корпус 02-04 (9 новых тестов: флип, потребление, транзит-гарды, инвариант, полный пайплайн, смешанное слово)
- `engine/engine.go` — ProcessKeyEvent возвращает вердикт обработчика; doc-контракт обновлён
- `engine/engine_test.go` — consumeStubHandler + таблица пропагации; decline-таблица переозаглавлена; паника-тесты не тронуты
- `cmd/goswitchd/main.go` — комментарий проводки актуализирован (ключи потребляются в RU)
- `test/e2e/case_word.go` — runWordRUEN, runWordMixed, helper flipToRU (тап + гейт на мод-записи)
- `test/e2e/main.go` — реестр word-ru-en/word-mixed + флаг-справка
- `mise.toml` — задачи e2e-word-ru, e2e-word-mixed (конвенция e2e-*, NOT in ci)
- `test/e2e/README.md` — две строки канонических инвокаций
- `.planning/.../red-evidence/02-04-task1-red.json` — RED-доказательство (RED_EVIDENCE_OK)

## Decisions Made

- **comboMask вместо буквальной Shift-only маски плана.** План предписывал `mods&^MaskShift == 0` для ветки потребления, но живая находка 02-03 (TestActor_LatchedModsStillFeed: на NumLock-столе каждая буква приходит с mods 0x10) делает её нерабочей — вся живая печать транзитила бы латиницей и флип был бы мёртв на реальном столе. Действует та же маска, что кормит буфер (Ctrl/Alt/Super запрещены, лэтчи разрешены); собственный верификат плана (живой e2e-word-ru) это подтверждает.
- **Тождественная карта в RU — транзит (пин выбора).** '2'→'2': consume=false без коммита, buf.Push('2') — руна появится в поле транзитом клиента и обязана быть в буфере. Инвариант script-true от выбора не зависит — пин TestActor_RUScriptTrueAllBranches фиксирует и отсутствие коммита, и попадание в буфер.
- **Коммит движка при a.eng == nil невозможен → транзит.** Если emitter ещё не привязан, RU-ветка не потребляет: клиент вставит исходную руну, буфер берёт её же — инвариант сохранён в обоих случаях.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] comboMask вместо mods&^MaskShift==0 в ветке RU-потребления**
- **Found during:** Task 1 (GREEN)
- **Issue:** буквальная маска плана противоречит живой находке 02-03 (NumLock-лэтч 0x10 на каждой букве) — потребление никогда бы не сработало на реальном столе, план's own e2e gate (word-ru-en) упал бы
- **Fix:** та же comboMask, что кормит буфер; маска задокументирована в константе
- **Files modified:** internal/session/actor.go
- **Verification:** TestActor_LatchedModsStillFeed зелёный; live PASS word-ru-en
- **Committed in:** fbf4d1c

**2. [Rule 3 - Blocking] funcorder/cyclop/gosec/gofmt strict-lint конформность**
- **Found during:** Task 1 (GREEN) и Task 2
- **Issue:** строгий линт: непэкспортированные методы после VerifyExpiry (funcorder), G115 на rune-конверсии EN-ветки, сложность runWordMixed 16>15
- **Fix:** перенос feedKey/flipScript, #nosec G115 с обоснованием, извлечение helper'а flipToRU (заодно устранив дублирование тап+гейт)
- **Files modified:** internal/session/actor.go, test/e2e/case_word.go
- **Verification:** mise run lint 0 issues
- **Committed in:** fbf4d1c, d691481

**3. [Rule 3 - Blocking] README стенда отстал бы от реестра кейсов**
- **Found during:** Task 2 (close-out)
- **Issue:** test/e2e/README.md перечисляет канонические mise-инвокации — две новые задачи создали бы дрейф документации
- **Fix:** две строки в списке инвокаций (прецедент 02-03: a7c3dc3)
- **Files modified:** test/e2e/README.md
- **Verification:** список соответствует реестру pickCase
- **Committed in:** 43c0631

---

**Total deviations:** 3 auto-fixed (1 bug, 2 blocking)
**Impact on plan:** Все правки продиктованы строгими гейтами владельца и живыми находками предшествующих планов; объём плана не расширен.

## Issues Encountered

None — оба живых кейса PASS с первого прогона (экран разблокирован, zenity-паттерны Фазы 1).

## TDD Red-Evidence Note

Первичная запись RED-доказательства была отклонена валидатором (INVALID_RED: zero_tests_discovered) — валидатор парсит TAP-сводку (`# tests/pass/fail` + `not ok N - …`), а не сырой вывод `go test -v`; запись переоформлена в TAP (34/28/6) и прошла как RED_EVIDENCE_OK. На сам RED это не влияло: падения целевых тестов были подлинными, на утверждениях.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Готово для 02-05 (коррекция сброса/Backspace-ветка, no-caps уровень 2): потребление и script-true буфер уже в рантайме, инвариант «то же без caps не выходит за буфер» пинен TestActor_RUDigitsFullPipeline на fake-sink с caps.
- SWCH-01 (Фаза 3) остаётся за консолидацией источников и UX-доками; Triple-решение по-прежнему Debug-заглушка (02-05).
- Известная поверхность для Фазы 3: commit в RU-режиме при живом a.eng всегда сквозь эмиттер — если появятся дополнительные инъекции, держать их под тем же инвариантом script-true.

## Self-Check: PASSED

- Commits: c5b5171, fbf4d1c, d691481, 43c0631 — все в git log ✓
- mise run ci exit 0 ✓
- mise run e2e-word-ru / e2e-word-mixed / e2e-word / e2e-word-space — все PASS живьём ✓
- RED предшествует GREEN: test(02-04) c5b5171 < feat(02-04) fbf4d1c ✓

---
*Phase: 02-korrektsiya-slova-en-ru*
*Completed: 2026-09-14*
