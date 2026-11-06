---
phase: 13-a11y-retsepty-telegram-zcode
plan: "02"
subsystem: testing
tags: [a11y, at-spi, electron, zcode, vm-stand, probe, selfcheck]

requires:
  - phase: 13-a11y-retsepty-telegram-zcode
    provides: поднятый стенд (демон active -debug, autocorrect on), откалиброванный scripts/a11y-selfcheck.py, ряды Telegram-пробы
provides:
  - приёмка A11Y-01 по живому дереву (RED→GREEN пара ×2, роль поля 79)
  - задокументированная граница слоя 3 для Electron-моста ZCode (материал residual-risk докам 13-03)
  - контракт D-06, доказанный живой матрицей 0/1/2 (скрипт без изменений)
  - парковочное состояние стенда с дословными откатами
  - три рецептовые находки установки ZCode (chrome-sandbox 4755, self-rewrite override, argv-перезапись)
affects: [13-03-docs, a11y-recipes]

actuals:
  tokens: 5928
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "экранный дифференциал дерева как доказательство флага (cmdline недоступен на Electron 44)"
    - "контроль оракула zenity перед вердиктом по границе (Pitfall 2 применён к Electron-рунгу)"

key-files:
  created: []
  modified:
    - docs/adr/adr-008-spike-log.md

key-decisions:
  - "Приёмка A11Y-01 зачтена по живому дереву (exit 0 + номер роли 79 ×2 подряд) — форма truth #1 плана; слой 3 зафиксирован честным красным с границей"
  - "cmdline-доказательство заменено: pid-множества ДО/ПОСЛЕ + экранный дифференциал дерева + прямой запуск с явным флагом — буквальная форма структурно невозможна на Electron 44"
  - "Все VM-local артефакты оставлены как артефакты принятых рецептов (осознанно, с командами отката); демон оставлен active -debug — ВМ не хост"

requirements-completed: [A11Y-01]

coverage:
  - id: D1
    description: "ZCode на ВМ с .desktop override присутствует в живом AT-SPI-дереве: selfcheck zcode → exit 0 с номером роли текст-класса (79), RED→GREEN пара ×2"
    requirement: A11Y-01
    verification:
      - kind: e2e
        ref: "vm: /usr/bin/python3 scripts/a11y-selfcheck.py zcode → exit 0 (r1 и r2, роль 79); verify-блок задачи 1 (selfcheck_exit=0)"
        status: pass
      - kind: e2e
        ref: "journal: ряды «a11y-проба zcode r1/r2 red|green» (docs/adr/adr-008-spike-log.md, секция Фазы 13)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Граница слоя 3 автокоррекции в ZCode: мост Electron отвечает обходу, но не эмитит identity-события → демон skip role-unknown (fail-closed); контроль zenity fired +1 в той же сессии"
    requirement: A11Y-01
    verification:
      - kind: e2e
        ref: "vm: дельты goswitchctl status до/после сентинела ×2 (fired без роста, abstained/skip_role_unknown +1, «autocorrect role unavailable»); dbus-monitor захват 0 state-changed; журнал демона 0 zcode-identity"
        status: pass
    human_judgment: true
    rationale: "честный красный слоя 3 — продуктовая граница моста Chromium на этом столе; владелец решает, как её формулировать в рецептах 13-03 (residual-risk)"
  - id: D3
    description: "Контракт self-check D-06 живой матрицей: exit 0 (zcode, роль 79), exit 1 (живой процесс без дерева), exit 2 (нет процесса) и exit 2 (usage) — все коды совпали, вердикты человекочитаемы"
    requirement: A11Y-01
    verification:
      - kind: e2e
        ref: "vm: rc0=0 rc1=1 rc2=2 rcu=2 (ряд «a11y-selfcheck контракт»); verify-блок задачи 2 (test rc0 -eq 0 -a rc2 -eq 2 -a rcu -eq 2) — TASK2_VERIFY_OK"
        status: pass
      - kind: unit
        ref: "host: /usr/bin/python3 -m py_compile scripts/a11y-selfcheck.py"
        status: pass
    human_judgment: false
  - id: D4
    description: "Стенд в парковочном состоянии: gsettings-тройка возвращена дословно, живые экземпляры остановлены, судьба каждого VM-local артефакта записана с командой отката"
    requirement: A11Y-01
    verification:
      - kind: e2e
        ref: "vm: gsettings get sources → [('xkb', 'us')]; ряд «a11y-стенд итог» (lock-enabled=true, idle-delay=300); verify задачи 3 — TASK3_VERIFY_OK"
        status: pass
    human_judgment: false

duration: 31min
completed: 2026-11-03
status: complete
---

# Phase 13 Plan 02: ZCode на ВМ — приёмка A11Y-01 по живому дереву Summary

**RED→GREEN пара ×2 по живому AT-SPI-дереву (plain-стаб 2 ноды → override-дерево 43+ нод, зонд exit 0 + номер роли поля 79), слой 3 — честный красный с границей «мост Electron не пушит identity-события», контракт self-check доказан живой матрицей 0/1/2, стенд запаркован с дословными откатами**

## Performance

- **Duration:** 31 min (17:40-18:11 UTC)
- **Started:** 2026-11-03T17:40:18Z
- **Completed:** 2026-11-03T18:10:54Z
- **Tasks:** 3
- **Files modified:** 1 (docs/adr/adr-008-spike-log.md — 7 рядов)

## Accomplishments

- A11Y-01 принят по живому дереву: ZCode 3.17.1/Electron 44.7.0 на ВМ (rsync 811 М), user-local .desktop override (шаблон CONFIG.md:385-392) + полный рестарт → зонд `scripts/a11y-selfcheck.py zcode` exit 0 с номером роли поля 79 [entry] ×2 раунда подряд; RED-форма доказана живьём (plain-запуск → exit 1, стаб 2 ноды, ×2)
- Слой 3 вскрыл и дважды воспроизвёл границу: дерево отвечает внешнему обходу (GetRole зонда работает), но мост Chromium не эмитит identity/state-события на шину (0 state-changed в захвате, 0 zcode-identity в журнале демона) → демон skip role-unknown (fail-closed); контроль zenity той же сессией fired +1 — цепочка оракула жива, красный специфичен мосту ZCode
- Контракт D-06 доказан живой матрицей всех кодов: 0 (zcode, роль 79) / 1 (telegram-tarball pid 26397, живой без дерева) / 2 (несуществующее имя) / 2 (usage); аудит тела расхождений не нашёл — diff скрипта 0
- Стенд запаркован: sources `[('xkb', 'us')]`, lock-enabled=true, idle-delay=300 (дословно, проверено get после set); судьба каждого VM-local артефакта записана с командой отката; хост чист по D-05 (override не создавался, живых прогонов не было)

## Task Commits

1. **Task 1: ZCode на ВМ — rsync, RED, override + полный рестарт, GREEN ×2** — `2876219` (docs)
2. **Task 2: Контракт self-check D-06 — живая матрица 0/1/2** — `910ad1f` (docs)
3. **Task 3: Итоговое состояние стенда — парковка с откатами** — `e9d7b18` (docs)

**Plan metadata:** см. ниже (final commit)

## Files Created/Modified

- `docs/adr/adr-008-spike-log.md` — 7 новых рядов секции Фазы 13: `a11y-механика zcode install`, `a11y-проба zcode r1/r2 red|green`, `a11y-selfcheck контракт`, `a11y-стенд итог`
- `scripts/a11y-selfcheck.py` — НЕ изменён (аудит задачи 2 расхождений с матрицей не нашёл; план допускал минимальный дифф — он оказался нулевым)

## Decisions Made

- Приёмка A11Y-01 зачтена в форме truth #1 плана (живое дерево: exit 0 + номер роли ×2); слой 3 записан честным красным — план явно допускал этот вердикт («вердикт из ряда, не из ожидания»)
- Доказательство флага у нового процесса заменено с буквального `/proc/<pid>/cmdline` (невозможно на Electron 44) на тройку: непересекающиеся pid-множества ДО/ПОСЛЕ + экранный дифференциал дерева (тот же welcome-экран: 2 ноды plain против 43 нод с флагом) + прямой запуск с явным флагом (ground truth, exit 0)
- Ячейка exit-1 матрицы исполнена на живом tarball-инстансе Telegram (pid 26397, вердикт 13-01 красный ×2) вместо snap — тот же класс машины, ZCode-override не тронут ячейкой
- Все артефакты (snap/flatpak/tarball//opt/ZCode/override/конфиг) оставлены на ВМ как артефакты принятых рецептов — осознанно, с командами отката каждого; демон оставлен active -debug (ВМ не хост, решение записано с полным откатом)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] chrome-sandbox SUID на ВМ-копии ZCode**
- **Found during:** Task 1 (стейджинг)
- **Issue:** первый запуск FATAL «SUID sandbox helper… mode 4755» — на хосте chrome-sandbox 755 (unprivileged userns), на ВМ userns ограничен AppArmor (kernel.apparmor_restrict_unprivileged_userns=1)
- **Fix:** `sudo chown root:root /opt/ZCode/chrome-sandbox && sudo chmod 4755 /opt/ZCode/chrome-sandbox` — только на ВМ-копии
- **Files modified:** нет (VM-local); факт в ряду «a11y-механика zcode install»
- **Verification:** запуск жив, все последующие старты без FATAL
- **Committed in:** 2876219 (ряд журнала)

**2. [Rule 1 - Bug in check] буквальная форма cmdline-доказательства невыполнима на Electron 44**
- **Found during:** Task 1 (шаг 4)
- **Issue:** Electron перезаписывает argv главного процесса на «ZCode» ≤1 с после exec, флаг не наследуется ни одним дочерним процессом, environ затёрт (966 байт) — `grep -- '--force-renderer-accessibility' /proc/<pid>/cmdline` даёт 1 всегда, даже при явно переданном флаге (ground-truth прогон)
- **Fix:** доказательство заменено на: непересекающиеся pid-множества + экранный дифференциал дерева + прямой запуск с флагом; verify-блок задачи 1 исполнен дословно и честно показал `selfcheck_exit=0` / `flag_in_cmdline=1`; замена — в ряду «a11y-механика zcode install»
- **Files modified:** docs/adr/adr-008-spike-log.md
- **Committed in:** 2876219

**3. [Rule 3 - Blocking] mise run ci падал на pollution PATH хост-сессии**
- **Found during:** Task 2 (хост-гейты)
- **Issue:** «compile: version go1.23.12 does not match go tool version go1.27.3» — сессионный PATH нёс системный /usr/local/go/bin (go1.27.3) раньше пина mise 1.23.12 при mise-заданном GOROOT; к коду отношения не имеет (диff демона пуст)
- **Fix:** запуск ci с исключением записи PATH (`PATH=$(… grep -v '^/usr/local/go/bin$' …)`); дополнительно `go clean -cache`; зелёный 47 c и 35 c (два прогона)
- **Files modified:** нет; заметка в ряду «a11y-selfcheck контракт»
- **Committed in:** 910ad1f

**4. [Rule 2 - Missing Critical finding, не фикс] ZCode перезаписывает user-local override при каждом старте**
- **Found during:** Task 1 (запуск через override)
- **Issue:** deep-link регистрация приложения сбрасывает Exec на `"/opt/ZCode/zcode" %U` (2/2 запусков) — gtk-launch успевает прочитать НАШ файл до старта приложения (флаг доходит), но файл на диске затирается; рецепт 13-03 обязан это назвать
- **Fix:** на время плана override пересоздавался перед каждым GREEN-запуском; как дефект записи — материал докам 13-03, не код
- **Committed in:** 2876219 (ряд)

---

**Total deviations:** 4 (1 blocking, 1 check-form, 1 environment, 1 documented finding)
**Impact on plan:** ни одна не меняет границу фазы (демон — ноль строк, `git diff --exit-code -- internal/ cmd/` чист в каждом коммите); все четыре — материал рецептов 13-03.

## Issues Encountered

- Транзиент: `pkill -f` по строке «/opt/ZCode/zcode» убивал собственный ssh-bash (самосовпадение) — сессия падала с 255; ушёл на pgrep с фильтрацией собственного pid и bracket-паттерн `zcod[e]`; на репо/стенде не сказалось
- Материальный факт для рецепта: роль поля ввода API-ключа ZCode = 40 (password-класс, «видимо, но неисправимо») — первое текст-классное поле (79) появляется только на главном экране после выхода из онбординга; GREEN достижим без логина (поля авторизации измеримы, что и предсказывал план)

## Known Stubs

None — проба доведена до машинных вердиктов во всех слоях; слой 3 — не стаб, а задокументированная граница (ряды r1/r2 green).

## User Setup Required

None — внешняя конфигурация не требуется.

## Next Phase Readiness

- 13-03 (рецепты в README×2/CONFIG.md) имеет весь материал: живая валидация override (exit 0 + роль 79 ×2), три обязательные строки рецепта установки (rsync → chmod 4755 → override), caveats (self-rewrite override; флаги только при старте; API-ключ-поле роль 40), строка запуска self-check (`/usr/bin/python3 scripts/a11y-selfcheck.py <app>`), residual-risk слой 3 для ZCode и недостижимость Telegram (вердикт владельца из 13-01)
- Стенд запаркован; докам 13-03 стенд не нужен

## Self-Check: PASSED

- SUMMARY.md exists; commits 2876219 / 910ad1f / e9d7b18 present in git log
- journal rows grep: 6 matches across the 4 mandated row names (r1 red/green, r2 red/green, контракт, итог)

---
*Phase: 13-a11y-retsepty-telegram-zcode*
*Completed: 2026-11-03*
