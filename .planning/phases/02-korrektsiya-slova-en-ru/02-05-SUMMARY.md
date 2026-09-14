---
phase: 02-korrektsiya-slova-en-ru
plan: 05
subsystem: input-correction
tags: [ibus, replacement-ladder, level2-backspace, verify-after, buffer-reset, e2e, tdd]

requires:
  - phase: 02-01
    provides: pure correction library (Buffer, Detect, Convert, BuildPlan with BOTH ladder formulas counted in runes)
  - phase: 02-02
    provides: surface drivers (chromium driver + runSurfaceSmoke abstraction, AT-SPI witness/readback)
  - phase: 02-03
    provides: tracer pipeline (engine emitters, actor two-phase verify with spontaneous-push cache, zenity focus recovery)
  - phase: 02-04
    provides: script flip on Single, RU consumption with Cyrillic commits, script-true buffer in all branches
provides:
  - эмиттер engine.Engine.ForwardKeyEvent(keyval, keycode, state uint32) + Emitter-seam на четыре примитива лестницы
  - исполнение уровня 2 лестницы (ADR-003): клиент без CapSurroundingText — немедленный бёрст ForwardKeyEvent(BackSpace)×(токен+хвост в рунах) → один CommitText(исправленное+хвост), доверие буферу (деградация ADR-004)
  - verify-after после каждой коррекции уровня 1: RequireSurroundingText от актёра, epoch-защита (verifyEpoch/pendingAfter), INFO {"msg":"correction verify","outcome":"mismatch"} счётчиком расхождения, без автоповторения; FocusOut/Reset и новая коррекция гасят раунд
  - таблица сброса CORR-09: press Return/KP_Enter(0xff8b)/Tab/Escape → buf.HardReset() (транзит, не потребление); engine.KeyKPEnter = 0xff8b
  - живые кейсы ladder-chromium / reset-escape + mise-задачи e2e-ladder-chromium / e2e-reset-escape
affects: [02-06, 02-07, phase-03]

actuals:
  tokens: 14200   # chars/4 over the realized diff (46514bb..4de8155: ~56.7k diff chars over 13 files)
  tasks: 2
  commits: 5
plan_head_before: 46514bb967dc39b21dbc606a6be7f6be46b1e48d

tech-stack:
  added: []
  patterns:
    - "verify-after с epoch-защитой: каждый раунд пост-сверки носит монотонный тег, таймер сверки сверяет тег под мьютексом — устаревший дедлайн от сменённого раунда не-op"
    - "порядок бёрст→коммит пинится единым op-log фейк-синка (запись вызовов в порядке прибытия), не отдельными счётчиками"
    - "оракул ФАКТИЧЕСКОГО уровня лестницы: подстрока '\"msg\":\"correction\",\"level\":N' в -debug-логе демона — кейс фиксирует что реально сделал демон, не предположение caps-бита (ADR-003)"

key-files:
  created:
    - test/e2e/case_ladder.go
    - .planning/phases/02-korrektsiya-slova-en-ru/red-evidence/02-05-task1-red.json
    - .planning/phases/02-korrektsiya-slova-en-ru/red-evidence/02-05-task2-red.json
    - .planning/phases/02-korrektsiya-slova-en-ru/deferred-items.md
  modified:
    - engine/engine.go
    - engine/keys.go
    - engine/engine_test.go
    - engine/emitter_wire_test.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - test/e2e/main.go
    - test/e2e/README.md
    - mise.toml

key-decisions:
  - "LIVE FINDING (2026-09-15, первый прогон ladder-chromium): google-chrome 153 сообщает caps 0x29 — CapSurroundingText УСТАНОВЛЕН — и ПРИМЕНЯЕТ DeleteSurroundingText; фактический уровень на цели 1, ibus#2354 (Pitfall 3) на этой цели устарел; оракул кейса переведён с предположения level:2 на фактический level:1 + пин verify-after match + readback"
  - "Живого свидетеля уровня 2 на этом столе НЕТ: все поверхности матрицы (zenity, chromium, GTE) сообщают бит (0x29); контракт уровня 2 несёт юнит-корпус (TestActor_Level2NoCaps/Level2WithTailAndRunes); матрица 02-06 берёт expect_level из фактических находок"
  - "LIVE FINDING: ydotool 0.1.8 резолвит имя 'Escape' в физическую клавишу E (fallback на первую букву: keyval 0x65/keycode 18); рабочее имя — 'esc' (0xff1b/keycode 1) — тот же класс ловушки, что 'space'→S из 02-03; closeEntrySurface (02-03, вне скоупа) — в deferred-items"
  - "verify-after гонит RequireSurroundingText сам после коррекции — пин кэш-пуша 02-03 (require=0) обновлён до require=1 (ровно один, verify-after); пин отказа no-surrounding снят — заменён уровнем 2 (TestActor_Level2NoCaps)"
  - "Reset — состояние движка, не потребление: Return/KP_Enter/Tab/Escape делают HardReset и транзит (клиент видит свою клавишу); FSM intervening-key-отмена не тронута — она уже питается теми же событиями"

patterns-established:
  - "isResetKeyval-таблица CORR-09 в актёре: keyval → HardReset, расширение списком не меняет архитектуру"
  - "logCorrectionDone: общая пара INFO-счётчик (D-20) + DEBUG-запись с level первым атрибутом после msg (оракул матрицы, D-21) для обоих уровней лестницы"

requirements-completed: [CORR-07, CORR-09]

coverage:
  - id: D1
    description: "Уровень 2 лестницы: эмиттер ForwardKeyEvent + исполнение бёрст→коммит по рунам на no-caps клиенте (деградация ADR-004 — доверие буферу)"
    requirement: CORR-07
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_Level2NoCaps"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_Level2WithTailAndRunes"
        status: pass
      - kind: unit
        ref: "engine/emitter_wire_test.go#TestEmitters_ForwardKeyEvent"
        status: pass
    human_judgment: false
  - id: D2
    description: "Verify-after после коррекции уровня 1: сверка суффикса исправленное+хвост, INFO mismatch счётчиком, без автоповторения, epoch-защита от устаревших таймеров"
    requirement: CORR-07
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_VerifyAfterLevel1"
        status: pass
    human_judgment: false
  - id: D3
    description: "Живой кейс ladder-chromium: фактический уровень лестницы в chromium зафиксирован -debug-записью (фактически 1), verify-after match, readback «привет» без дубля (Pitfall 3)"
    requirement: CORR-07
    verification:
      - kind: e2e
        ref: "mise run e2e-ladder-chromium"
        status: pass
    human_judgment: false
  - id: D4
    description: "Триггеры сброса буфера CORR-09: Return/KP_Enter(0xff8b)/Tab/Escape в корпусе актёра, FocusOut/Reset, Backspace-pop (пин 02-01), Ctrl-изоляция (пин 02-04)"
    requirement: CORR-09
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_HardResetKeyvals"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestBuffer_ResetByFocusOut"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestBuffer_CtrlIsolation"
        status: pass
      - kind: unit
        ref: "internal/correct/buffer_test.go#TestBuffer_BackspacePop"
        status: pass
    human_judgment: false
  - id: D5
    description: "Живой кейс reset-escape: ghbdtn + Escape + двойной тап → поле не тронуто, отказ empty-buffer в логе"
    requirement: CORR-09
    verification:
      - kind: e2e
        ref: "mise run e2e-reset-escape"
        status: pass
    human_judgment: false

# Metrics
duration: 38 min
completed: 2026-09-14
status: complete
---

# Phase 2 Plan 5: Лестница замены полная — уровень 2 + verify-after + триггеры сброса Summary

**Уровень 2 лестницы (Backspace×N по рунам + перекоммит), verify-after с epoch-защитой и все триггеры сброса CORR-09; живой Chromium-кейс опроверг предположение level:2 — google-chrome 153 сообщает CapSurroundingText и применяет удаление, фактический уровень 1 закреплён**

## Performance

- **Duration:** 38 min
- **Started:** 2026-09-14T21:49:59Z
- **Completed:** 2026-09-14T22:28:30Z
- **Tasks:** 2
- **Files modified:** 13 (9 code + 4 planning artifacts)

## Accomplishments

- Ladder level 2 исполняется: no-caps клиент получает бёрст ForwardKeyEvent(0xff08, 14, 0)×(токен+хвост в рунах) и один CommitText(исправленное+хвост) — немедленно, сверка невозможна (деградация ADR-004), порядок бёрст→коммит пинится op-log
- Verify-after: каждая коррекция уровня 1 сама просит свежий surrounding text; расхождение — INFO-счётчик `{"msg":"correction verify","outcome":"mismatch"}`, повторной коррекции нет; устаревшие таймеры гасятся epoch-тегом
- Триггеры сброса CORR-09: Return/KP_Enter(0xff8b)/Tab/Escape — HardReset с транзитом; FocusOut/Reset; Backspace-pop и Ctrl-изоляция запинены корпусом
- Живые кейсы ladder-chromium и reset-escape PASS на драйвере 02-02 без его изменений; регресс всех word-кейсов и `mise run ci` зелёные

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: уровень 2 + verify-after + пин эмиттера** — `5e22ba2` (test)
2. **Task 1 GREEN: исполнение уровня 2, verify-after, живой кейс ladder-chromium** — `a656b38` (feat)
3. **Task 2 RED: корпус сбросов CORR-09 + KeyKPEnter** — `1e28ec0` (test)
4. **Task 2 GREEN: таблица сброса, живой кейс reset-escape** — `ebe7b8f` (feat)
5. **docs: e2e README + deferred-items** — `4de8155` (docs)

**Plan metadata:** (this commit)

_Примечание: обе задачи TDD — RED-коммит предшествует GREEN; RED_EVIDENCE_OK-записи в `red-evidence/02-05-task{1,2}-red.json`._

## TDD Gate Compliance

| Plan | RED | GREEN | REFACTOR | Status |
|------|-----|-------|----------|--------|
| 02-05 Task 1 | `5e22ba2` (RED_EVIDENCE_OK: 39 tests, 4 fail on assertions) | `a656b38` | — not needed (helper extraction done in GREEN) | Pass |
| 02-05 Task 2 | `1e28ec0` (RED_EVIDENCE_OK: 58 tests, 1 fail — HardResetKeyvals ×4 subtests) | `ebe7b8f` | — not needed | Pass |

REFACTOR-коммиты не потребовались: рефакторинг (logCorrectionDone, fakeSink op-log) вошёл в GREEN-шаги как очевидная очистка, поведение не менялось, корпус зелёный до и после.

## Files Created/Modified

- `engine/engine.go` — эмиттер ForwardKeyEvent (ifaceEngine-сигнал, 3×uint32), Emitter-seam на 4 примитива
- `engine/keys.go` — KeyKPEnter = 0xff8b (XK_KP_Enter)
- `engine/emitter_wire_test.go` — wire-пин эмиттера (порядок аргументов, интерфейс, путь)
- `engine/engine_test.go` — KeyKPEnter в пине констант; detached-quiet расширен
- `internal/session/actor.go` — уровень 2 (executeLevel2), verify-after (armAfterVerify/pendingAfter/verifyEpoch), таблица isResetKeyval, logCorrectionDone, clearAfter в FocusOut/Reset/новом раунде
- `internal/session/actor_test.go` — op-log фейк-синка; 5 новых тестов; эволюция двух пинов 02-03
- `test/e2e/case_ladder.go` — runLadderChromium (фактический уровень + verify-after + readback), runResetEscape (esc, empty-buffer, нетронутое поле)
- `test/e2e/main.go`, `mise.toml`, `test/e2e/README.md` — реестр, задачи, документация

## Decisions Made

- См. key-decisions во frontmatter; все решения продиктованы живыми прогонами (caps 0x29 chromium, резолв имён ydotool) — план-предположения скорректированы по доказательствам, архитектура не менялась.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 — Plan assumption falsified live] Оракул ladder-chromium: level:2 → фактический level:1**
- **Found during:** Task 1 (первый живой прогон e2e-ladder-chromium)
- **Issue:** План предписывал `waitForLog '"msg":"correction","level":2'` — предположение ADR-003/ibus#2354, что Chromium не сообщает/не применяет surrounding text. Живой прогон: google-chrome 153 шлёт caps **0x29** (CapSurroundingText установлен), пушит SetSurroundingText при наборе, коррекция ушла уровнем 1, удаление ПРИМЕНЕНО, verify-after дал match, поле = «привет».
- **Fix:** Кейс пинирует ФАКТИЧЕСКИЙ уровень (1) — ровно то, зачем ADR-003 требовал кейс («фиксирует, какой уровень реально сработал»), плюс пин verify-after match и readback (сторожа Pitfall 3 сохранены). Уровень 2 живьём на этом столе невоспроизводим: zenity/chromium/GTE — все сообщают бит (0x29); контракт уровня 2 несёт юнит-корпус. Матрица 02-06 берёт expect_level из этой находки.
- **Files modified:** test/e2e/case_ladder.go, mise.toml
- **Verification:** `mise run e2e-ladder-chromium` — PASS (level:1 + done + verify match + readback «привет»); caps-запись 0x29 в закреплённом логе
- **Committed in:** a656b38
- **Owner attention:** находка противоречит контексту ADR-003 (масштаб проблемы Pitfall 3 на цели) — вынести на verify-work; ADR не редактировался (Accepted, гейт владельца). Запись в WINDOWS-ledger (kind: deviation).

**2. [Rule 1 — Bug in own new code] Байтовый счёт вместо рунного в двух местах**
- **Found during:** Task 1 (прогон RED / живой прогон кейса)
- **Issue:** `len(wordRU)` = 12 байт ≠ 6 рун — в TestActor_Level2WithTailAndRunes (RED-прогон показал «want 13») и в settle-гейте кейса (`chars=12` при 6 рунах — упало живьём). Ирония Pitfall 2 в собственном коде.
- **Fix:** `len([]rune(...))` в обоих местах.
- **Files modified:** internal/session/actor_test.go, test/e2e/case_ladder.go
- **Verification:** корпус зелёный; кейс PASS
- **Committed in:** 5e22ba2 (тест, до RED-коммита), a656b38 (кейс)

**3. [Rule 1 — Blocking] ydotool 0.1.8: имя 'Escape' резолвится в клавишу E**
- **Found during:** Task 2 (живой прогон reset-escape: в логе нет 0xff1b, токен вырос до «ghbdtne», коррекция сработала по чужому слову)
- **Issue:** `ydotool key Escape` — fallback на первую букву: keyval 0x65/keycode 18. Рабочее имя — `esc` (0xff1b/keycode 1, доказано живой пробой с демоном-наблюдателем).
- **Fix:** кейс шлёт `esc` с комментарием-пином находки. Пред-существующий той же ловушки код (closeEntrySurface, 02-03) — вне скоупа, записан в deferred-items.md.
- **Files modified:** test/e2e/case_ladder.go
- **Verification:** `mise run e2e-reset-escape` — PASS; в логе 0xff1b/keycode 1 и `reason:empty-buffer`
- **Committed in:** ebe7b8f

**4. [Corpus evolution, plan-mandated] Снят пин отказа no-surrounding; обновлён пин кэш-пуша**
- **Found during:** Task 1 GREEN
- **Issue:** С уровнями 2/verify-after два пина 02-03 стали ложно-красными: подкейс «no surrounding capability» (отказ заменён исполнением уровня 2) и «cache hit → require 0» (теперь ровно один Require — verify-after).
- **Fix:** Подкейс снят (замещён TestActor_Level2NoCaps, комментарий-ссылка оставлен); кэш-пин ожидает ровно 1 verify-after Require.
- **Files modified:** internal/session/actor_test.go
- **Verification:** весь корпус зелёный под -race
- **Committed in:** a656b38

---

**Total deviations:** 4 auto-fixed (3× Rule 1, 1× corpus evolution mandated by the plan feature)
**Impact on plan:** Все — следствия живых прогонов, ради которых кейсы существуют. Архитектура, формулы и объём не менялись. Живой level-2 свидетель отсутствует на цели (неустранимо в объёме плана — поверхностей без бита на столе нет), контракт закрыт корпусом.

## Issues Encountered

None beyond the deviations above — обе живые находки (chromium caps, ydotool names) обнаружены и закрыты в рамках своих задач.

## Known Stubs

None — заглушек не осталось.

## User Setup Required

None — внешних сервисов нет.

## Next Phase Readiness

- Готово к 02-06 (матрица): expect_level-кейсы брать из фактических уровней (chromium=1 — живая находка), reset-tab-кейс живьём планируется именно там (по плану 02-05)
- Для владельца на verify-work: находка об устаревании ibus#2354 на цели (google-chrome 153 применяет delete-surrounding) — контекст ADR-003 стоит дополнить; сам ADR не редактировался (Accepted)
- deferred-items.md: ydotool-имя "Escape" в closeEntrySurface (02-03) — поправить при следующем касании шелл-поверхности

---
*Phase: 02-korrektsiya-slova-en-ru*
*Completed: 2026-09-14*

## Self-Check: PASSED

- All 4 created artifacts exist on disk (case_ladder.go, KeyKPEnter in keys.go, both red-evidence records)
- All 5 commits present in history (5e22ba2, a656b38, 1e28ec0, ebe7b8f, 4de8155)
- Acceptance spot-greps: ForwardKeyEvent emitter, KeyKPEnter=0xff8b
- git rev-list 46514bb..HEAD = 5 (matches actuals.commits)
