---
phase: 10
plan: 16
subsystem: session/correction-safety
tags: [dst-guard, safety, mutter-crash, bisection-round4, g-10-1]
requires:
  - "Бисекция 2026-10-16 (.planning/debug/2026-10-16-stand-gnome-shell-dst-crash.md, R2/R6b/R7) — вердикт (a)/(c)"
  - "Двухфазный уровень-1 10-11 (executeCorrection/flushLevel1Confirm) — точка гварда"
  - "executeLevel2 (форвард-ветка, WR-01) — форма диверсии"
provides:
  - "Гвард DST-эмиссии в executeCorrection: DeleteSurroundingText только при перегенерационном свидетельстве материализованности (surrSeen/surrEng)"
  - "Диверсия неэвиденсированных раундов в executeLevel2-форму с INFO-рубежом dst-guarded"
  - "Счётчик corr_dst_guarded в goswitchctl status (форма ac_role_errors)"
  - "Ретест-гейт DST-маршрута, запиненный рядами spike-log round-4 (закрыт на всех существующих версиях mutter до 51.1 + main)"
affects:
  - "10-17 (round-4 машина): строится поверх дерева, чистого от DST-риска; флип-семантику не трогает"
  - "10-18 (стенд/установка/приёмка): устанавливаемая сборка несёт гвард"
tech-stack:
  added: []
  patterns:
    - "generation-scoped evidence (surrSeen + surrEng — bind к эмиттеру поколения, форма flapEng)"
    - "safety-guard by state evidence (гвард по свидетельству состояния, не по параметрам/таймингу)"
key-files:
  created:
    - ".planning/phases/10-term-slovo/evidence-20261016-10-16-red-task1.json"
  modified:
    - "internal/session/actor.go"
    - "internal/session/actor_test.go"
    - "internal/ctlsvc/ctlsvc.go"
    - "docs/adr/adr-008-spike-log.md"
    - "docs/adr/ADR-008-terminal-transport.md"
key-decisions:
  - "Свидетельство материализованности биндится к эмиттеру поколения (surrEng, форма flapEng) поверх плана-минимума (только булев флаг): churn-реминт эмиттера без жизненной пары не наследует свидетельство мёртвой генерации — иначе гвард не закрывал бы остаточный маршрут T-10-16-02"
  - "Диверсия зовёт executeLevel2 с terminal=false: раунд классифицирован НЕ-терминальным, чтобы дойти до уровня-1 — форма фолбэка сохраняет классификационную правду раунда"
  - "Пин апстрим-фикса честно UNPINNED: точный NULL/garbage-гард delete_surrounding НЕ СУЩЕСТВУЕТ апстримом (проверено GitLab REST API: main + теги 47.0..51.1) — ретест-гейт закрыт на ВСЕХ существующих версиях mutter, сильнее планового «условия на 46.2»"
patterns-established:
  - "Гвард саксесс-точки по свидетельству состояния (state evidence), не по параметрам эмиссии"
  - "Счётчик диверсии безопасности в status по форме ac_role_errors"
requirements-completed:
  - "requirements: [TERM-02] (см. assumptions плана: край TERM-02 — за владельцем, UAT pending; plans 10-17/10-18 разделяют ID)"
coverage:
  human_judgment: false
  rationale: "Все verify-блоки трёх задач исполнены и зелёные: RED-гейт на неизменённом HEAD (RED_EVIDENCE_OK), корпус TestDstGuard_* зелёный, регресс-рампа зелёная, mise run ci зелёный целиком, grep-гейты документов зелёные, go.mod/go.sum заморожены"
actuals:
  tokens: 12020
  tasks: 3
  commits: 3
  plan_head_before: 44ca636c4536b239c142fb96da15122aaea1bcc5
duration: 18min
completed: 2026-10-16
status: complete
---

# Phase 10 Plan 16: DST-гард девственного окна + пин апстрим-фикса mutter Summary

**Гвард DeleteSurroundingText в коррекционной машине: DST только при перегенерационном свидетельстве материализованности surrounding-состояния, иначе громкая диверсия в форварды — класс краха gnome-shell (mutter 46.2, strlen по NULL, бисекция 2026-10-16) закрыт на уровне машины; апстрим-фикс честно unpinned (не существует вплоть до 51.1 + main), ретест-гейт запинен рядами spike-log.**

## Performance

| Метрика | План | Факт |
|---|---|---|
| Токены (chars/4 по диффу 48080 байт) | 50000 | ~12020 |
| Задачи | 3 | 3 |
| Коммиты | — | 3 (измерено: rev-list 44ca636..HEAD) |
| Длительность | — | 18 мин |

## Accomplishments

- **Гвард DST-эмиссии (Task 2, GREEN):** в `executeCorrection` (единственная продакшн-точка DeleteSurroundingText) эмиссия разрешена только при `surrSeen && surrEng == текущий эмиттер`; без свидетельства — диверсия в существующую executeLevel2-форму (чанки ≤8, pacing 20 мс, ctxEpoch-перечек, один commit, WR-01) с громким INFO-рубежом `dst-guarded` (латентность, без содержимого — D-20/D-21) и счётчиком `corr_dst_guarded` в status. Require не подавлен; машина 10-11, selection-путь, шов classifyTerminal, вето D-55, flip=false — байт-нетронуты.
- **Перегенерационное свидетельство:** `surrSeen`/`surrEng` set в `handleSurroundingLocked` (каждый пуш/материализующий ответ, bind к эмиттеру), clear ТОЧНО в FocusOut/Reset-ветке (`a.surr=nil`, `ctxEpoch++`) — живёт и умирает с кэшем окружения.
- **Строгий TDD (Task 1, RED):** TestDstGuard_* корпус — КРАСНЫЙ на неизменённом HEAD (обе негативные формы падают на delete(-6,6) к девственному контексту), evidence-файл RED_EVIDENCE_OK (target_test_failed); пин MaterializedContextByteUnchanged зелёный уже на HEAD.
- **Документация (Task 3):** секция «Фаза 10 round-4: DST-гард девственного окна и пин апстрим-фикса» в spike-log (ряды safety-guard / mutter-fix pin / retest-gate) + амендрмент ADR-008 2026-10-16 с Reversibility-строкой; mise run ci зелёный целиком.
- **Находка пина (важнее плановой гипотезы):** точный фикс «NULL/garbage-гард surrounding-состояния» апстримом НЕ СУЩЕСТВУЕТ — краш-сайт `strlen (text_input->surrounding.text)` без NULL-гарда жив на main и на всех проверенных тегах (47.0/48.0/48.1/49.8/50.6/51.1); запинено частичное харденение 48.0 (2c42a284 — границы арифметики после strlen; fb86d5d5 — сброс surrounding на смену фокуса, text остаётся NULL). Ретест-гейт закрыт на всех существующих версиях mutter, не только 46.2; noble-бэкпорта нет (краш воспроизведён на 46.2-1ubuntu0.24.04.18).

## Task Commits

| Task | Name | Commit |
| ---- | ---- | ------ |
| 1 | TestDstGuard_* корпус — КРАСНЫЙ на неизменённом HEAD | 410473f |
| 2 | Гвард DST-эмиссии — свидетельство + диверсия в форварды | 3bd0aaf |
| 3 | Секция round-4 в spike-log + амендрмент ADR-008 | c727d86 |

## Files

- internal/session/actor.go — гвард в executeCorrection; surrSeen/surrEng (set/clear-сайты); corrDstGuarded; Status.CorrDstGuarded
- internal/session/actor_test.go — TestDstGuard_NeverMaterializedContextGetsForwards / _MaterializedContextByteUnchanged / _EvidenceDiesWithGeneration (+ counter-пины GREEN-этапа)
- internal/ctlsvc/ctlsvc.go — токен `corr_dst_guarded=` в renderStatus (форма ac_role_errors, старые ключи не тронуты)
- docs/adr/adr-008-spike-log.md — секция round-4 (safety-guard + mutter-fix pin + retest-gate)
- docs/adr/ADR-008-terminal-transport.md — амендрмент 2026-10-16 + Reversibility
- .planning/phases/10-term-slovo/evidence-20261016-10-16-red-task1.json — RED-свидетельство (RED_EVIDENCE_OK)

## Decisions Made

- **surrEng — bind свидетельства к эмиттеру поколения** (форма flapEng): план предписывал только булев флаг с clear в FocusOut/Reset-ветке; анализ достижимых путей показал, что остаточный маршрут бисекции (кэш, переживший смену поколения при churn-пропущенной жизненной паре — AttachEngine это чистый store указателя) такой флаг не закрывает — гвард получался никогда не срабатывающим. Rule 2: добавлена проверка `a.surrEng != eng` (interface identity — оба живых реализации указатели, дисциплина flapEng :384-394). Все план-пиннутые сайты set/clear соблюдены дословно.
- **Диверсия с terminal=false:** раунд, дошедший до executeCorrection, классифицирован НЕ-терминальным (иначе 10-04 увёл бы его в executeLevel2 раньше) — фолбэк сохраняет классификационную правду (immediate settle, без терминального verify-after).
- **Пин UNPINNED по причине отсутствия фикса:** веб-проверка доступными средствами (GitLab REST API через curl — штатного веб-инструментария нет, все движки отключены в config.json; локально dpkg/changelog) показала: гвард NULL-текста не существует ни на одной версии; ряд фиксирует источники и предел инструментария дословно по плану.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Свидетельство дополнено bind'ом к эмиттеру поколения (surrEng)**
- **Found during:** Task 2 (анализ достижимых путей executeCorrection перед реализацией)
- **Issue:** булев флаг с clear ТОЛЬКО в FocusOut/Reset-ветке не закрывает остаточный маршрут плана (T-10-16-02: «свидетельство переживает смену поколения контекста (host churn)») — churn может пропустить жизненную пару, AttachEngine не сбрасывает состояние, и гвард на этом маршруте не сработал бы вовсе (обе RED-формы остались бы красными после GREEN)
- **Fix:** свидетельство = `surrSeen && surrEng == текущий эмиттер` (interface identity, форма flapEng); set/clear-сайты плана соблюдены дословно
- **Files modified:** internal/session/actor.go
- **Commit:** 3bd0aaf

**2. [Rule 1 - Bug] gofmt-выравнивание поля surrSeen после вставки в struct**
- **Found during:** Task 2 (mise run lint)
- **Fix:** `gofmt -w internal/session/actor.go`
- **Files modified:** internal/session/actor.go
- **Commit:** 3bd0aaf

**3. [Rule 3 - Blocking] RED-evidence в TAP-формате чекера**
- **Found during:** Task 1 (gsd_run check tdd-red-evidence вернул INVALID_RED/zero_tests_discovered на сыром go-test-выводе)
- **Fix:** запись перестроена в TAP-форму чекера (`not ok N - Test` + `# tests/# pass/# fail`) механической конвертацией РЕАЛЬНОГО прогона (полный -race -v прогон пакета, 226 tests / 2 fail) — форма evidence-20261014-10-10-red-task1.json; RED_EVIDENCE_OK получен на тот же прогон
- **Files modified:** .planning/phases/10-term-slovo/evidence-20261016-10-16-red-task1.json
- **Commit:** 410473f

### Plan-Delta (конструкция RED-сценариев, в границах плана)

Формулировка RED-сценариев плана («новый раунд к свежему контексту после FocusOut/Reset — DST уходит на HEAD») буквально невоспроизводима: после FocusOut/Reset кэш пуст, маршрут 10-10 (pendingStaleCache) уводит такой раунд в форварды уже на HEAD. Обе RED-формы построены на ДОСТИЖИМОМ маршруте до executeCorrection (единственном, где DST вообще эмитится): кэш предыдущей генерации, переживший churn-реминт эмиттера; наблюдаемый отказ RED — тот же, что предписан планом (delete-запись на проводе к контексту без свидетельства текущей генерации). FocusOut/Reset-форма ~:1166 присутствует в TestDstGuard_EvidenceDiesWithGeneration как обещано. Задокументировано в комментариях корпуса.

**Всего отклонений: 3 авто-фикса + 1 plan-delta конструкции тестов; на приёмку не влияют (все acceptance-критерии задач исполнены дословно).**

## Issues Encountered

- Чекер TDD-гейта ожидает TAP-формат (`not ok N - ...`), а не сырой вывод go test — преодолено конвертацией реального прогона (см. Deviation 3).
- Пин апстрим-фикса: гипотеза плана «фикс запинен коммит+версия» не подтвердилась — фикса нет апстримом вовсе; это УСИЛИВАЕТ гейт (закрыт на всех версиях до 51.1), а не ослабляет безопасность.
- Хост несёт mutter 46.2-1ubuntu0.24.04.16, стенд бисекции — .18; оба без фикса (крах воспроизведён на .18) — противоречия нет, noble-бэкпорта нет по changelog до июня 2026.

## User Setup Required

- Ничего; хост не мутирован (демон владельца v1.1.1-142 продолжает работу; гвард поедет на стол владельца установкой плана 10-18).

## TDD Gate Compliance

- Task 1: RED-коммит `test(10-16)` 410473f на неизменённом actor.go; `gsd_run check tdd-red-evidence` → **RED_EVIDENCE_OK (target_test_failed)**; evidence-файл закоммичен.
- Task 2: GREEN-коммит `feat(10-16)` 3bd0aaf после RED_EVIDENCE_OK; корпус зелёный, регресс-рампа зелёная.
- REFACTOR не требовался (правка минимальна, golangci-lint 0 issues).

## Next Phase Readiness

- **10-17 (round-4 машина: форварды + drain-settle + флип вне окна):** стартует поверх дерева, чистого от DST-риска; flip-семантику гвард не трогал — 10-17 ведёт её свободно.
- **10-18 (стенд + установка + owner-приёмка):** устанавливаемая сборка несёт гвард и счётчик corr_dst_guarded; журнальный рубеж dst-guarded — ожидаемая (редкая) строка при диверсиях.
- **Ретест DST-маршрута:** закрыт до появления апстрим-фикса NULL/garbage-гарда; перепроверка пина — новый датированный ряд в секции round-4 spike-log при любом обновлении стека.
- G-10-1 остаётся открытым (TERM-02 за владельцем, UAT end-of-phase; 10-16 не претендует на фикс потери — это safety-половина round-4).

## Self-Check: PASSED

Все заявленные файлы существуют; все три коммита (410473f, 3bd0aaf, c727d86) найдены в git log; commits=3 измерено `git rev-list --count 44ca636..HEAD`; mise run ci зелёный на финальном HEAD; go.mod/go.sum нетронуты; пост-коммитная проверка удалений — пусто.
