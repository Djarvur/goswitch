---
phase: 12-term-stroka
plan: 01
subsystem: testing
tags: [e2e-stand, wezterm, primary-selection, copymode-recipe, sel-probe, phase-gate, probe-first]

# Dependency graph
requires:
  - phase: 09-spayk-i-defekt-wezterm
    provides: ADR-008 + журнал спайка (контракт датированных рядов), capability-таблица Фазы 9 (рецепт CopyMode как кандидат канала чтения), kill-ряд VTE 18:19:27
  - phase: 10-term-slovo
    provides: терминальная машина слова (машина round-4: пары press+release, чанки, drain-settle, отложенный флип), семейство case_term* и стенд-механика vm-primary, leg-probe красные ряды 10-06
provides:
  - Кейс term-line-probe (измерительный инструмент в репо): три ячейки сел-пробы + сводный итог, honest-abort обёртка, регистрация main.go/mise (NOT in ci)
  - Фикстура gs-e2e-wezterm-xim-line.lua (xim-тело + рецепт CopyMode байт-тот же с wayland-фикстурой)
  - Датированное доказательство: канал «чорд → CopyMode-рецепт → PRIMARY» мёртв на корне — физический рецепт не поднимает выделение ни на wayland, ни на xim, ни форвардом (A1/A2/A3), при функциональном PRIMARY-канале (сентинел-раундтрип) и доставке чорда (трейс движка)
  - Закрытие гейта OQ1 вердиктом владельца 2026-11-03: цель строки уже достигнута машиной слова Фазы 10 на длинных пробегах — примитив чтения выделения не нужен; 12-02+ retired superseded
affects: [phase-12-closure, ROADMAP-v1.2, phase-13]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
actuals:
  tokens: 7500          # chars/4 over the realized diff (398 insertions + journal rows + SUMMARY)
  tasks: 2
  commits: 3            # MEASURED: git rev-list --count 911773c..HEAD at SUMMARY time (2 task + 1 closure row); closing metadata commits follow

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Сел-проба канала: сентинел-раундтрип доказывает инструмент чтения ДО вердикта канала (primary_len==сентинел ⇒ красное — не артефакт wl-paste)"
    - "Приватность рядов пробы: только сентинел-марка/факт смены/длина/латентность — содержимое выделения никогда не попадает в журнал (D-20/D-21)"

key-files:
  created:
    - "test/e2e/case_term_line_probe.go"
    - "test/e2e/fixtures/gs-e2e-wezterm-xim-line.lua"
    - ".planning/phases/12-term-stroka/12-01-SUMMARY.md"
  modified:
    - "test/e2e/main.go (реестр + usage оба места)"
    - "mise.toml (e2e-term-line-probe, NOT in ci)"
    - "test/e2e/spike_journal_test.go (secLineProbe пин)"
    - "docs/adr/adr-008-spike-log.md (секция «Фаза 12 сел-проба канала чтения (term-line-probe)»: pending → 3 красные ячейки + итог + mechanics + закрытие)"

key-decisions:
  - "Гейт OQ1 закрыт вердиктом владельца 2026-11-03 (дословно: «коррекция всей строки в wezterm у меня работает, а коррекция выделения мне не нужна»): цель строки достигнута существующей машиной слова Фазы 10 на длинных пробегах (журнал хоста: 160 done / 0 rejected, до 47 символов одним жестом); примитив чтения выделения мёртв на живой пробе и владельцем не запрошен"
  - "Фолбэк «буфер вместо выделения» не предлагался и не исполнялся — D-01 соблюдена дословно (отвергнутый VTE-примитив по духу, SC3 теряет якорь)"
  - "Кейс term-line-probe остаётся в репо как измерительный инструмент (NOT in ci); рецепт CopyMode из capability-таблицы Фазы 9 помечен неактуальным для применения «чтение строки»"

patterns-established:
  - "Probe-first гейт с вердикт-развилкой на владельца: красная ячейка → честные ряды + checkpoint, никогда тихий фолбэк"
  - "Механика двухзаездной честности: honest-abort заезда на средстве (uinput-деградация) + документированное лекарство (virsh reboot, прецедент 2026-10-13) + полный заезд на свежей сессии"

requirements-completed: [TERM-03]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Кейс term-line-probe + xim-line фикстура + регистрации (main.go реестр/usage, mise NOT in ci) + секция журнала, запиненная TestSpikeJournalSectionsExist (RED→GREEN)"
    requirement: TERM-03
    verification:
      - kind: unit
        ref: "test/e2e/spike_journal_test.go#TestSpikeJournalSectionsExist (secLineProbe пин: RED до появления секции, GREEN после)"
        status: pass
      - kind: other
        ref: "mise exec -- go vet ./test/e2e && go build ./... && greps (реестр/фикстура/mise/пин) && mise run ci — все зелёные"
        status: pass
    human_judgment: false
  - id: D2
    description: "Живые ряды сел-пробы на стенде vm-primary: три датированные ячейки (a)/(b)/(c) + сводный итог, все честные RED; privacy-дисциплина (D-20/D-21) и clip-frozen (D4/D5) выдержаны"
    requirement: TERM-03
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#сел-проба-a-физика-wayland-23:16:57 / b / c / итог 23:17:04 (заезд 2 exit 0; заезд 1 honest-abort)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Развязка гейта: вердикт владельца (дословно) + журнальные ряды хоста (160 done / 0 rejected / max 47 символов) → фаза закрыта, 12-02+ retired superseded"
    verification: []
    human_judgment: true
    rationale: "Закрывающий оракул — решение владельца на checkpoint:decision gate=blocking-human плюс пассивная сверка журнала его хоста; машинного оракула для «мне не нужна коррекция выделения» не существует"

# Metrics
duration: 42min
completed: 2026-11-03
status: complete
plan_head_before: 911773ca2d661a2a962e0035971c25dafefaf06b
---

# Phase 12 Plan 01: Сел-проба канала чтения (term-line-probe) Summary

**Три красные ячейки живой пробы (физика-wayland / физика-xim / форвард) доказали: канал CopyMode→PRIMARY мёртв на корне; владелец закрыл гейт вердиктом «строка уже работает машиной слова» — фаза завершена без машины строки, 12-02+ retired.**

## Performance

- **Duration:** 42 min (executed 2026-11-02 22:52 – 23:22 UTC; owner verdict 2026-11-03)
- **Started:** 2026-11-02T22:52:45Z
- **Completed:** 2026-11-03 (owner closure of the checkpoint)
- **Tasks:** 2/2
- **Files modified:** 6

## Accomplishments

- Кейс term-line-probe собран по готовым хелперам (recipeChordRow/wezLegProbeRow-формы): три ячейки сел-пробы с honest-вердиктами обеих ветвей, CLIPBOARD-сентинел сквозь ячейки (clip-frozen класс), privacy-дисциплина D-20/D-21 (только сентинел-марка/факт смены/длина/латентность — содержимое выделения ни в один ряд), honest-abort обёртка, teardown.
- Фикстура gs-e2e-wezterm-xim-line.lua: xim-тело владельца + рецепт CopyMode байт-тот же с wayland-фикстурой (md5 совпадает); регистрации main.go (реестр + usage оба места) и mise (NOT in ci); секция журнала открыта датированным pending-рядом и запинена контрактом секций (RED→GREEN).
- Живой прогон на стенде vm-primary (заезд 2 после документированного лекарства среды virsh-reboot): три датированных ряда + итог — все ячейки RED: физический рецепт-чорд не поднимает PRIMARY ни на wayland, ни на xim, ни форвардом через probe-forward; при этом канал чтения функционален (сентинел-раундтрип wl-copy/wl-paste --primary) и чорд доставлялся (трейс движка чистыми парами 0xffe1/0xff50) — красное это свойство рецепта, не инструментария.
- Развилка гейта исполнена дословно: checkpoint:decision gate=blocking-human владельцу (СТОП, 12-02 не стартовал, фолбэк «буфер вместо выделения» не предлагался — D-01); вердикт владельца 2026-11-03 закрыл фазу: цель уже достигнута машиной слова Фазы 10.

## Task Commits

Each task was committed atomically:

1. **Task 1: трассер — кейс term-line-probe + xim-line фикстура + регистрации + секция** - `12c49fe` (feat)
2. **Task 2: живой прогон пробы — три датированных ряда + итог (все RED)** - `231d443` (docs)
3. **Закрытие по вердикту владельца — итоговая строка журнала** - `8656321` (docs)

**Plan metadata:** closing commits (SUMMARY/ROADMAP/REQUIREMENTS/STATE) follow this file.

## Files Created/Modified

- `test/e2e/case_term_line_probe.go` — кейс: три ячейки, сводный итог, honest-abort
- `test/e2e/fixtures/gs-e2e-wezterm-xim-line.lua` — xim-тело + рецепт CopyMode (байт-тот же)
- `test/e2e/main.go` — регистрация term-line-probe (noSurfacePreflight, recipeWatchdog) + usage
- `mise.toml` — задача e2e-term-line-probe (NOT in ci)
- `test/e2e/spike_journal_test.go` — пин секции secLineProbe
- `docs/adr/adr-008-spike-log.md` — секция «Фаза 12 сел-проба канала чтения (term-line-probe)»: intro + pending + 3 ячейки + abort-ряд заезда 1 + итог + mechanics + закрытие

## Decisions Made

See key-decisions. Core: гейт закрыт вердиктом владельца (не кодом); кейс остаётся измерительным инструментом; D-01/D3 соблюдены дословно.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Ветка при старте не соответствовала предписанию оркестратора**
- **Found during:** init (до Task 1)
- **Issue:** оркестратор предписал gsd/phase-12-term-stroka «уже checkout»; фактически HEAD был на gsd/phase-13-a11y-retsepty-telegram-zcode (6 docs-only коммитов планирования Фазы 13 впереди)
- **Fix:** переключение на предписанную ветку (runtime-маркеры STATE/state.json оркестратора восстановимы — отброшены); коммиты Фазы 13 не тронуты
- **Verification:** git log — оба task-коммита на gsd/phase-12-term-stroka
- **Committed in:** 12c49fe (контекст коммита)

**2. [Rule 3 - Blocking] Стендовая среда требовала ре-активации**
- **Found during:** Task 2 precondition
- **Issue:** /tmp/stand-env.sh исчез (перезагрузка хоста); зеркало устарело; юнит стенда перенацелен на goswitch-rel (Orca Round 2) — шаг «restart unit -debug» рецепта 10-01 устарел
- **Fix:** stand-env.sh восстановлен по рецепту 10-01; rsync зеркала + go build на ВМ; кейс несёт собственный -debug-демон (доказанная форма), запаркованный юнит возвращён в active
- **Committed in:** 231d443 (mechanics-строка журнала)

**3. [Rule 3 - Blocking] Заезд 1 оборван на арминг-гейте (uinput-деградация)**
- **Found during:** Task 2
- **Issue:** key-trace floor unmet ×3 в ячейке (b) — класс деградации uinput 2026-10-13 (стенд запаркован с 2026-11-01)
- **Fix:** документированное лекарство — virsh reboot (прецедент boot 21:27:35); заезд 2 на свежей сессии — полный прогон, exit 0; не ретрай-цикл (honest-abort ряд заезда 1 в журнале)
- **Committed in:** 231d443

**4. [Rule 3 - Blocking] Дрейф инструментария mise на хосте**
- **Found during:** Task 1 verify
- **Issue:** `mise run` в неактивированной оболочке резолвит системный go1.27.3 при экспортированном mise GOROOT 1.23.12 → compile version mismatch на любой задаче
- **Fix:** гейты исполняются через mise-активированную оболочку (буквальный `mise run ci` зелёный) / `mise exec --`; конфиг mise не менялся (CI-раннеры без системного go не затронуты)
- **Committed in:** — (окружение хоста, не репо)

---

**Total deviations:** 4 auto-fixed (4 Rule 3 blocking). **Impact:** ни одного отступления от контрактов задач; все — восстановление окружения (ветка/стенд/инструментарий), зафиксированы журнальными рядами и контекстом коммитов.

## Issues Encountered

- Ячейка (a) наблюдалась дважды (заезды 1 и 2) — одинаковая красная сигнатура; развилка плана предполагала зелёные физические ячейки, фактическая краснота глубже (A1 опровергнут) — доведена до владельца в checkpoint дословно.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Фаза 12 закрыта по решению владельца 2026-11-03: TERM-03 зачтён достижением существующей машины (см. REQUIREMENTS/ROADMAP); планы 12-02…12-05 помечены retired superseded (файлы сохранены).
- Следующая: Фаза 13 (A11Y — рецепты Telegram/ZCode), планы 13-01…13-03 верифицированы на ветке планирования; блокеров нет.

---
*Phase: 12-term-stroka*
*Completed: 2026-11-03*
