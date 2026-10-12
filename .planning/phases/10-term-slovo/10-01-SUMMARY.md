---
phase: 10-term-slovo
plan: 01
subsystem: testing
tags: [e2e-stand, libvirt-vm, ydotool, ibus, keyboard-liveness, wave-0, venue-verdict]

# Dependency graph
requires:
  - phase: 09-spayk-i-defekt-wezterm
    provides: ADR-008 spike journal (контракт датированных строк), debug-дневник стенда (оркестраторская проба ~15:30 KEYBOARD_ALIVE=true, состав стенда), дефект атрибуции ядра 7.2.9-2-t2 как мотивация VM-площадки
provides:
  - Вердикт площадки vm-primary для живых рядов 10-05/10-06 (строка venue-verdict в журнале ADR-008, ссылается на два venue-probe раунда)
  - Готовый стенд goswitch-stand: goswitchd рабочего дерева (87908b0) active с -debug, движок goswitch-en глобальный, ydotool+uinput 0660, wl-clipboard, gnome-terminal, зеркало ~/goswitch, /tmp/stand-env.sh
  - Закреплённая стенд-механика в журнале (venue-mechanics): sources-takeover переживает uinput-churn, arming-ключ BackSpace, Enter только type+\n, idle-lock выключен, unlock-gate
affects: [10-02, 10-03, 10-04, 10-05, 10-06, phase-11-menu]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
actuals:
  tokens: 2399          # chars/4 over the journal diff 87908b0..HEAD (plan carried estimate 35000)
  tasks: 3
  commits: 5            # MEASURED: git rev-list --count gsd-plan-head-before-10-01..HEAD (3 task + metadata + self-check)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Venue-probe протокол: двухступенчатый критерий клавиатурности (key-trace демона -debug + hex байтов корпуса в pty), два подряд зелёных раунда до вердикта"
    - "Стенд-гейты по key-trace демона, никогда по AT-SPI-свидетелю (Pitfall 7); corpus одной ydotool type-командой с embedded \\n"

key-files:
  created:
    - ".planning/phases/10-term-slovo/10-01-SUMMARY.md"
  modified:
    - "docs/adr/adr-008-spike-log.md (4 датированные строки: venue-ready, venue-mechanics, venue-probe r1, venue-probe r2, venue-verdict)"

key-decisions:
  - "Площадка живых рядов Фазы 10 = vm-primary: ряды 10-05/10-06 исполняются внутри ВМ через /tmp/stand-env.sh; хост — только owner-UAT (ядро 7.2.9-2-t2 клавиатурно мертво для фоновых окон)"
  - "Sources-takeover в ВМ (sources=[('ibus','goswitch-en')], эквивалент install-шага ADR-001/D-40) — без него gsd re-apply на uinput-churn ydotool сбрасывает глобальный движок в foreign xkb и key-trace оракул умирает"
  - "Стенд без idle-lock (lock-enabled=false, idle-delay=0): autologin-сессия самоблокируется на простое и глушит клавиатуру под экраном блокировки"
  - "Откаты стенда задокументированы в строке venue-verdict: sources [('xkb','us')], lock-enabled=true/idle-delay=300"

patterns-established:
  - "Venue-probe discipline: арминг-гейт по key-trace (BackSpace x3), корпус одной type-командой, hex и счётчики в журнал, pty-артефакты tmp-only"
  - "Two-consecutive-greens применяется не только к живым рядам, но и к самой пробе площадки"

requirements-completed: []  # TERM-02 declared by ALL six 10-xx plans — this plan delivered only its Wave-0 venue leg; stays open per shared-ID gate until 10-06

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Клавиатурность стенд-spawn фоновых окон ВМ доказана (venue-probe r1+r2, keyboard-live)"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#venue-probe-r1-2026-10-12-1412 (key-trace delta 14 + pty hex 67686264746e0a; flip control d0bfd180d0b8d0b2d0b5d1820a)"
        status: pass
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#venue-probe-r2-2026-10-12-1416 (key-trace delta 14 + pty hex 67686264746e0a; flip control pass)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Вердикт площадки живых рядов vm-primary записан со ссылкой на обе строки venue-probe и следствием для 10-05/10-06"
    requirement: TERM-02
    verification:
      - kind: other
        ref: "grep venue-verdict docs/adr/adr-008-spike-log.md → 1 строка, токен vm-primary, ссылки r1/r2, следствие 10-05/10-06 (verify-блок Task 3)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Стендовые внутренности и зеркало репо готовы для живых рядов (goswitchd -debug active, инструменты, ~/goswitch, /tmp/stand-env.sh)"
    verification:
      - kind: other
        ref: "ssh stand@192.168.122.150 'systemctl --user is-active goswitchd && command -v ydotool wl-copy gnome-terminal' → active + 3 пути (verify-блок Task 1)"
        status: pass
    human_judgment: false

# Metrics
duration: 37min
completed: 2026-10-12
status: complete
---

# Phase 10 Plan 01: Wave-0 стенд — клавиатурная проба VM goswitch-stand + вердикт площадки Summary

**Клавиатурность стенд-spawn окон ВМ доказана дважды подряд (key-trace + pty hex, venue-probe keyboard-live ×2) и площадка живых рядов Фазы 10 зафиксирована вердиктом vm-primary в журнале ADR-008.**

## Performance

- **Duration:** 37 min
- **Started:** 2026-10-12T13:42:38Z
- **Completed:** 2026-10-12T14:20:37Z (commit close ~14:22)
- **Tasks:** 3/3
- **Files modified:** 1 (docs/adr/adr-008-spike-log.md — журнальные строки; plan journal-only)

## Accomplishments

- Wave-0 пре-реквизит VALIDATION Фазы 10 закрыт: фоново-заспавненные (ssh → GNOME-сессия) окна ВМ goswitch-stand доставляют клавиши в демон (-debug key-trace) и байты в pty — два подряд зелёных раунда, каждый с полным флип-контролем (shift_r → mode ru → switch_engine goswitch-ru → кириллица «привет» в pty — полный канал ProcessKeyEvent→commit).
- Вердикт площадки vm-primary записан (venue-verdict 14:20) со ссылкой на venue-probe r1 14:12 + r2 14:16 и операционным следствием: живые ряды 10-05/10-06 исполняются внутри ВМ, мир пассивен для владельца.
- Стенд доведён до полной готовности живых рядов: goswitchd рабочего дерева active с -debug (оракул), sources-takeover, ydotool/uinput, wl-clipboard, зеркало ~/goswitch, /tmp/stand-env.sh.
- Стенд-механика закреплена для кейсов 10-05/10-06 (venue-mechanics строка): unlock-gate, arming по key-trace BackSpace, Enter только `type $'\n'`, EOF-чорда не трейсится — teardown reap по pid.

## Task Commits

Each task was committed atomically:

1. **Task 1: Готовность стенда и внутренности** - `d7539e9` (docs)
2. **Task 2: Клавиатурная проба (venue-probe ×2)** - `61dc503` (docs)
3. **Task 3: Вердикт площадки (venue-verdict)** - `cc2e2d8` (docs)

**Plan metadata:** see final docs commit (`docs(10-01): complete …`)

## Files Created/Modified

- `docs/adr/adr-008-spike-log.md` — секция «2026-10-12 ~13:42-14:20 UTC — Wave-0 площадка Фазы 10»: 5 датированных строк (venue-ready, venue-mechanics, venue-probe r1, venue-probe r2, venue-verdict) в контрактном табличном формате журнала
- `/tmp/stand-env.sh` (tmp-only, не коммитится) — VM_ADDR=192.168.122.150, VM_USER=stand, VM_REPO=/home/stand/goswitch
- Внутри ВМ (не в репо): unit + drop-in -debug, udev-правило uinput 0660, wl-clipboard, sources-takeover, зеркало ~/goswitch

## Decisions Made

See key-decisions. Core: vm-primary; sources-takeover как обязательная часть внутренностей стенда; stand idle-lock off; откаты задокументированы.

## Deviations from Plan

**[Rule 3 - Blocking] /dev/uinput был root:root 0600 (план требует 0660)**
- **Found during:** Task 1
- **Issue:** ydotool не мог открыть uinput; uaccess-ACL из дневника оркестратора отсутствовал
- **Fix:** udev-правило 60-goswitch-stand-uinput.rules (KERNEL=="uinput", MODE="0660", GROUP="input") + chown/chmod; запись открытием подтверждена
- **Files modified:** (ВМ) /etc/udev/rules.d/60-goswitch-stand-uinput.rules
- **Verification:** open-for-write от пользователя stand — OK; инъекции раундов прошли
- **Commit:** d7539e9 (строка venue-ready)

**[Rule 3 - Blocking] wl-clipboard отсутствовал в ВМ**
- **Found during:** Task 1
- **Issue:** wl-copy/wl-paste не установлены (план требует)
- **Fix:** apt-get install wl-clipboard 2.2.1-1build1 — только архив Ubuntu 24.04 (T-10-01-03 соблюдён)
- **Commit:** d7539e9

**[Rule 2 - Missing critical] sources-takeover отсутствовал в ВМ**
- **Found during:** Task 2 (обнаружен по падению key-trace)
- **Issue:** gsd re-apply на create/destroy uinput-устройства ydotool сбрасывал глобальный движок в foreign xkb:us::eng — key-trace оракул (основной критерий пробы) умирал после первой инъекции
- **Fix:** gsettings sources="[('ibus','goswitch-en')]" (эквивалент install-шага ADR-001/D-40; прежнее значение [('xkb','us')] задокументировано для отката); движок переживает churn — проверено
- **Commit:** 61dc503 (строка venue-mechanics)

**[Rule 3 - Blocking] Сессия ВМ самоблокировалась (LockedHint=yes)**
- **Found during:** Task 2
- **Issue:** autologin-сессия простояла 55 мин → idle-delay 300s + lock-enabled=true → пробы шли под экраном блокировки (окна мапятся, key-trace/pty мертвы — ложная сигнатура «keyboard-dead»)
- **Fix:** loginctl unlock-session 1; на стенде lock-enabled=false, idle-delay=0 (иначе живые ряды упрутся в то же)
- **Commit:** 61dc503

Инструментальные находки (не отклонения плана, закреплены в журнале): ydotool 0.1.8 не резолвит `key Return` (падает в посимвольный набор — keyval 0x72 в трейсе), модификаторы не доходят до ProcessKeyEvent, EOF-чорда ctrl+d не трейсится. Обходы: Enter через `type $'…\n'`, arming BackSpace x3, teardown reap по pid.

**Total deviations:** 4 auto-fixed (2 Rule 3 blocking, 1 Rule 2 missing-critical, 1 Rule 3 environment). **Impact:** ни одного отступления от контракта задач; все находки — доводка стендовых внутренностей, которые план предписывал «проверить/довести»; зафиксированы датированными строками журнала и питают механику 10-05/10-06.

## Issues Encountered

- Первая серия попыток Task 2 (13:50-14:00) дала честный FAIL-след (key-trace пуст, pty пуст) — причина найдена (lock-экран), попытки не вошли в вердикт; вердикт опирается на два подряд зелёных чистых раунда 14:12/14:16.
- Пробная строка армирования модификаторами из R1-батча фазы 9 на стенде не воспроизводится (0 записей) — arming-ключ заменён на BackSpace; для кейсов 10-05/10-06 арминг-гейты должны использовать ключ, доходящий до ProcessKeyEvent.

## User Setup Required

None - no external service configuration required. (Стенд самодостаточен; доступ владельца не нужен — что и подтверждает вердикт vm-primary.)

## Next Phase Readiness

- Планы 10-02…10-04 (correct-ядро, классификация+вето, терминальная ветка) не зависят от стенда и готовы к исполнению; живые ряды 10-05/10-06 получили площадку (vm-primary), оракул (-debug key-trace), зеркало и стенд-механику в журнале.
- For wezterm-ряда 10-06: wezterm-nightly 20261010 уже на борту ВМ (поставлен оркестратором, hold); precondition 10-06 (из kронштейна угроз) выполнен средой.
- Отсутствуют блокеры. TERM-02 остаётся открытой до 10-06 (shared-ID, шесть планов).

---
*Phase: 10-term-slovo*
*Completed: 2026-10-12*

## Self-Check: PASSED

SUMMARY exists; commits d7539e9 / 61dc503 / cc2e2d8 verified in git log; final metadata commit 3786d37 + self-check sync 54801b4; ledger recount 5 commits from plan base 87908b0.
