---
phase: 11-menu-punkt-v-chernyy-spisok
plan: 01
subsystem: session
tags: [orca-terminal, verify-dead-class, lifecycle-reset, re-mint-aftermath, tdd, spec-delta, blocklist-menu]

requires:
  - phase: 10-term-slovo
    provides: терминальный класс verify-dead (classifyTerminal, setVerifyDead, flap-окно 10-07), машина отложенного флипа (10-08), ритуал установки на хост (10-06)
  - phase: main via PR #24
    provides: Round 1 фикс 946b4cd — карве-аут synthetic FocusOut (слит до этого плана)
provides:
  - Round 2 фикс клиентского LifecycleReset: выученный терминальный класс переживает эхо-Reset окна re-mint aftermath (flipAftermathWindow=2 с); настоящий Reset вне окна стирает байт-прежне — PR #25 слит в защищённый main (58b07cc)
  - Машинно-верифицированное подтверждение гипотезы Round 2: элиминация 22/22 рядов стирки дня = ветка Reset (журнал green106 обеих жизней демона)
  - Установленная на хост сборка Round 2 (v1.2.0-alpha.2-15-g2c00c10, unit active; откат назван до верификации)
  - Датированная spec-ревизия меню v2 в docs/SPEC.md (пункт «Добавить ИмяПриложения в чёрный список», состояния D-01/D-03, QuoteMeta-действие, freshness, D-04 « — язык», принятый риск a11y-шины) и контракт apps_blocklist в docs/CONFIG.md (+regen SKILL.md)
affects: [11-02 (пункт меню/freshness/D-04 — правит HandleLifecycle на post-fix main), 11-03 (writer-метод), 11-04 (клик-композиция), 11-05 (e2e), phase-verification (UAT end-of-phase)]

actuals:
  tokens: 11212
  tasks: 2
  commits: 7

tech-stack:
  added: []
  patterns:
    - "Re-mint aftermath carve-out: клиентское эхо ре-минта (Reset на новом engine-объекте) приравнено к синтетической паре для поведенческого класса; окно с журнальным происхождением, single-slot дисциплина T-10-07-04, гашение реальным FocusOut"
    - "Подтверждение гипотезы машинной элиминацией журнала вместо отдельного probe-инсталла: инструмент наблюдения встроен в сам фикс (INFO-ряд lifecycle aftermath=true)"

key-files:
  created:
    - ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-01-SUMMARY.md"
  modified:
    - "internal/session/actor.go (Round 2 фикс: verifyDeadOnContextDeath, armFlipAftermath/cancelFlipAftermath/flipAftermathExpiry, UseFlipAftermathTimer, arm в flipTo, контрактные комментарии)"
    - "internal/session/actor_test.go (TestTerminalClass_VerifyDeadSurvivesClientResetAfterOwnFlip RED→GREEN, TestTerminalClass_VerifyDeadClearedByResetOutsideAftermath)"
    - "docs/SPEC.md (ревизия 2026-10-31: пункт меню, D-01..D-04, freshness, принятый риск)"
    - "docs/CONFIG.md (меню-абзац + строка apps_blocklist)"
    - "skills/goswitch-config/SKILL.md (regen региона из CONFIG.md)"
    - ".planning/debug/orca-terminal-second-shift.md (ряды Round 2: подтверждение, install, честный journal-verify)"

key-decisions:
  - "Round 2 карве-аут: клиентский Reset в окне flipAftermathWindow (2 с, журнальное происхождение — эхо 946 мс ряда 13:55) не стирает класс; buffer hard-reset/ctxEpoch/кэши — байт-прежний полный reset; настоящий Reset вне окна и реальный FocusOut (гасит окно) стирают — fail-closed сохранён"
  - "Подтверждение гипотезы — машинная элиминация журнала (22/22) вместо отдельного probe-инсталла; прямой INFO-маркер встроен в фикс; живой ряд пассивной верификации дозапишется на следующей сессии владельца"
  - "Integration-ordering исполнен дословно: PR #25 (новая fix-ветка) слит в защищённый main ДО всякого фазового PR; фазовая ветка влита на post-fix main (f943e60) — 11-02/11-04 правят актуальный HandleLifecycle"
  - "Спека прежде кода: ревизия меню v2 добавляет блок, семантические строки D-53 байт-целы (grep-гейт); сужение D-03 зафиксировано решением владельца 2026-10-31, текст роадмапа не менялся"

patterns-established:
  - "verifyDeadOnContextDeath: дисциплина стирки класса при смерти контекста — один шов (deferred/immediate/ничего) с явным контрактом"
  - "flipAftermathWindow: внутренняя константа с журнальным происхождением, никогда не конфиг-ключ (прецедент focusFlapWindow)"

requirements-completed: [MENU-01]

coverage:
  - id: D1
    description: "Round 2 фикс: выученный терминальный класс переживает клиентский Reset окна re-mint aftermath; настоящий Reset вне окна стирает (fail-closed); TDD полный цикл RED→GREEN"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestTerminalClass_VerifyDeadSurvivesClientResetAfterOwnFlip (RED f1f0019 — gsd check tdd-red-evidence RED_EVIDENCE_OK; GREEN 2c00c10) + #TestTerminalClass_VerifyDeadClearedByResetOutsideAftermath + пины Round 1 (OwnFlip, SyntheticFocusOut) + #TestTerminalClass_VerifyDeadClearedByReset"
        status: pass
      - kind: other
        ref: "go build ./..., go vet ./..., golangci-lint 0 issues, go test -race ./... — зелёные на post-fix main; go.mod/go.sum нетронуты"
        status: pass
    human_judgment: false
  - id: D2
    description: "Сборка Round 2 установлена на хост: НОВЫЙ release-dir v1.2.0-alpha.2-15-g2c00c10 ( прежние каталоги нетронуты), путь отката назван ДО верификации, unit active"
    verification:
      - kind: other
        ref: "systemctl --user is-active goswitchd = active; journal 23:12:00 component registered + engine activated; строка «Round 2 install» в .planning/debug/orca-terminal-second-shift.md"
        status: pass
    human_judgment: false
  - id: D3
    description: "Пассивная journal-верификация живого цикла: следующая коррекция с ОДНОГО нажатия после flip+эхо-Reset, INFO-ряд lifecycle aftermath=true, ноль строк terminal class verify_dead=false между flip-released и следующей коррекцией"
    verification:
      - kind: other
        ref: ".planning/debug/orca-terminal-second-shift.md#round-2-journal-verify — честный ряд-ожидание (владелец вышел из сессии 23:21:15, loginctl — только gdm); оракул сформулирован, ряд завершится на следующей сессии"
        status: unknown
    human_judgment: true
    rationale: "Живой ряд требует сессии владельца (окно Orca — Wayland-нативное, вне X11/AT-SPI-достижимости исполнителя; инъекция в живую сессию — T-09-02-03); владелец освобождён от Round 2 («проверяй сам в виртуалке»), by-ear зачёт — UAT end-of-phase (прецедент G-10-1); вердикт не синтезируется"
  - id: D4
    description: "Датированная spec-ревизия меню v2 (2026-10-31): пункт + состояния D-01/D-03 + QuoteMeta-действие + freshness (D-02) + D-04 « — язык» + принятый риск подмены identity; семантические строки D-53 байт-целы"
    requirement: MENU-01
    verification:
      - kind: other
        ref: "grep-гейты: '2026-10-31', ' — язык', 'Неизвестная идентичность приложения НЕ запрет' (docs/SPEC.md); 'пункт меню', 'QuoteMeta' (docs/CONFIG.md) — все зелёные"
        status: pass
    human_judgment: false
  - id: D5
    description: "Контракт apps_blocklist дополнен поведением пункта меню; сгенерированный регион SKILL.md синхронизирован"
    verification:
      - kind: unit
        ref: "cmd/skillgen#TestSkillgen_CommittedFileInSync (зелёный после mise-задачи skillgen-regen / go run ./cmd/skillgen)"
        status: pass
    human_judgment: false

# Metrics
duration: 51min
completed: 2026-10-31
status: complete
---

# Phase 11 Plan 01: Подготовка фазы — Round 2 фикс + spec-ревизия меню v2 Summary

**Клиентский Reset окна re-mint aftermath больше не стирает выученный терминальный класс (PR #25 → main, сборка установлена), и SPEC §4.3 несёт датированную ревизию пункта меню «в чёрный список» с контрактами CONFIG.md — спека прежде кода**

## Performance

- **Duration:** 51 min
- **Started:** 2026-10-31T19:46:47Z
- **Completed:** 2026-10-31T20:57Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Round 2 закрыт машинной стороной: гипотеза клиентского Reset ПОДТВЕРЖДЕНА машинной элиминацией журнала дня (22/22 рядов стирки = ветка Reset), полный цикл RED→GREEN запинен TestTerminalClass_VerifyDeadSurvivesClientResetAfterOwnFlip, фикс слит в защищённый main через НОВУЮ fix-ветку и PR #25 (58b07cc; PR #24 не переоткрывался), сборка v1.2.0-alpha.2-15-g2c00c10 установлена и active
- SPEC §11 несёт датированный ревизионный блок меню v2 (2026-10-31): пункт «Добавить ИмяПриложения в чёрный список» с состояниями D-01/D-03/потолок-64, QuoteMeta-действие с дедупом и живой перечиткой, freshness-семантика (D-02), D-04 « — язык», принятый риск подмены identity; D-53 байт-целы
- docs/CONFIG.md дополнен (меню-абзац + строка apps_blocklist), регион SKILL.md перегенерирован, TestSkillgen_CommittedFileInSync зелёный
- Integration-ordering: фазовая ветка влита на post-fix main ДО всякого фазового PR — 11-02/11-04 работают с актуальным HandleLifecycle

## Task Commits

Each task was committed atomically:

1. **Task 1 (код, fix-ветка fix/orca-terminal-client-reset-aftermath):** RED `f1f0019` (test), GREEN `2c00c10` (feat) — слиты в main через PR #25 (merge `58b07cc`)
2. **Task 1 (фазовая ветка):** `7008542` (docs: ряды Round 2 в debug-док) + `f943e60` (merge post-fix main в фазовую ветку)
3. **Task 2:** `7150799` (docs: SPEC ревизия + CONFIG контракт) + `0870f50` (docs: regen SKILL.md)

**Plan metadata:** final docs commit (SUMMARY + STATE + ROADMAP)

_Note: TDD-задача Task 1 несёт RED (test) и GREEN (feat) коммиты в fix-ветке; REFACTOR-изменений после последнего зелёного прогона не было (извлечение verifyDeadOnContextDeath вошло в GREEN — требование гейта cyclop 17>15)._

## Files Created/Modified

- `internal/session/actor.go` — карве-аут aftermath: Reset в окне flipAftermathWindow не стирает класс (INFO-маркер lifecycle aftermath=true); verifyDeadOnContextDeath (извлечение ради cyclop); armFlipAftermath/cancelFlipAftermath/flipAftermathExpiry; UseFlipAftermathTimer; arm в успешной flipTo round trip; контрактные комментарии (HandleLifecycle doc, credit-заметка, flipTo doc)
- `internal/session/actor_test.go` — TestTerminalClass_VerifyDeadSurvivesClientResetAfterOwnFlip (полный цикл, RED на main) + TestTerminalClass_VerifyDeadClearedByResetOutsideAftermath (fail-closed пин)
- `docs/SPEC.md` — ревизионный блок 2026-10-31 (пункт меню, состояния, действие, freshness, D-04, принятый риск)
- `docs/CONFIG.md` — меню-абзац + примечание в строке apps_blocklist
- `skills/goswitch-config/SKILL.md` — regen региона таблицы
- `.planning/debug/orca-terminal-second-shift.md` — ряды Round 2: подтверждение (22/22 элиминация), install, честный journal-verify ряд-ожидание

## Decisions Made

- Карве-аут только для ПОВЕДЕНЧЕСКОГО КЛАССА: buffer hard-reset, ctxEpoch, кэши и ретайр pending — байт-прежний полный reset (минимальная семантическая дельта по плану)
- Окно 2 с: удвоение наблюденных ~0.9 с переподключения медленного клиента; внутренняя константа с журнальным происхождением, не конфиг-ключ
- Подтверждение гипотезы элиминацией журнала вместо probe-инсталла (меньше рестартов демона на машине владельца); прямой маркер встроен в фикс
- Реальный FocusOut гасит aftermath-окно (Reset после реальной смены контекста обязан стирать — fail-closed)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] SKILL.md разошёлся с CONFIG.md после ревизии**
- **Found during:** Task 2 (финальная зелёная итерация)
- **Issue:** TestSkillgen_CommittedFileInSync красный — правка строки apps_blocklist в docs/CONFIG.md требует регенерации сгенерированного региона skills/goswitch-config/SKILL.md
- **Fix:** mise-задача skillgen-regen (go run ./cmd/skillgen) — регион перегенерирован, тест зелёный
- **Files modified:** skills/goswitch-config/SKILL.md
- **Verification:** go test -race ./cmd/skillgen/ + полный suite зелёный
- **Committed in:** 0870f50

**2. [Rule 3 - Blocking] mise run ci в неинтерактивной оболочке исполнителя падает на toolchain-разъезде**
- **Found during:** Task 1 (зелёная итерация)
- **Issue:** сессионная env несёт GOROOT=/usr/local/go (1.27.3) и mise не инжектирует tool-пути в PATH задач — go1.23.12 из mise.toml грузил чужие tool-бинарники («compile: version go1.23.12 does not match go tool version go1.27.3»); dbus-monitor на ibus-сокете тоже закрыт (eavesdropping запрещён)
- **Fix:** те же гейты исполнялись напрямую с заблокированными mise-инструментами: go build ./..., go vet ./..., golangci-lint 2.15.0 (из mise-инсталла), go test -race ./..., go mod tidy + git diff --exit-code — согласовано со STACK.md («build with current 1.27.x», язык модуля go 1.23); дубль-проверка в обеих деревьях
- **Files modified:** нет (исполнение, не код)
- **Verification:** все гейты зелёные в fix-ветке и на post-fix main
- **Committed in:** (изменений кода не требовалось)

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking-окружение)
**Impact on plan:** оба отклонения — обслуживание зелёной итерации и дисциплины генерируемых артефактов; объём плана не менялся

### Незакрытый живой ряд (честно, не синтезируется)

Пассивная journal-верификация (verify-критерий «следующая коррекция с одного нажатия» + INFO-ряд lifecycle aftermath=true) НЕ снята: владелец завершил сессию 23:21:15 (к 23:40 в loginctl только gdm-экран входа), natural-циклы невозможны, а драйв цикла исполнителем небезопасен (окно Orca — Wayland-нативное: X11-тулинг слеп, второй инстанс с --user-data-dir отказал на Wayland-connect, AT-SPI из сторонних клиентов недостоверен — фаза 9, инъекция в живую сессию — T-09-02-03). Демон активен в reconnect-лупе (Discover() пере-читает адрес на каждой попытке — подключится сам при новом входе владельца); ряд дозапишется на следующей сессии (оракул сформулирован в debug-доке). By-ear зачёт — UAT end-of-phase (прецедент G-10-1) — маршрутизация самого плана. Запись в WINDOWS.md (unrun-verify, открытая).

## Issues Encountered

- Лог-зонд dbus-monitor на приватном ibus-сокете невозможен (старый dbus-daemon без Monitoring, eavesdropping отклонён) — усилено решение об элиминации
- Второй инстанс Orca.AppImage (--user-data-dir) не spawn'ится: Wayland-connect refused у ozone (адмиссия демона Orca перевела запуск на существующий daemon-сокет)
- После logout'а владельца (23:21:15) стартованный вручную демон остался в reconnect-лупе на устаревшем адресе — само-восстановление конструктивно (пере-discovery на каждой попытке), перезапуск вручную не потребовался

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Post-fix main несёт Round 2 — планы 11-02/11-04 (HandleLifecycle, actor.go ~1282-1360) работают с актуальными номерами строк (фазовая ветка уже влита на post-fix main)
- SPEC-ревизия 2026-10-31 пинит макет пункта до кода 11-02 (дисциплина вехи, прецедент D-55); реестр menuID расширяется по паттернам menu.go
- Открытый хвост: ряд «Round 2 journal verify» завершится на следующей сессии владельца; закрытие G-фазы — по UAT end-of-phase (MENU-01 остаётся открытой в REQUIREMENTS до 11-02, shared-ID гейт)

## Self-Check: PASSED

- Файлы: docs/SPEC.md, docs/CONFIG.md, skills/goswitch-config/SKILL.md, .planning/debug/orca-terminal-second-shift.md, internal/session/actor.go, internal/session/actor_test.go — FOUND; тест TestTerminalClass_VerifyDeadSurvivesClientResetAfterOwnFlip в дереве
- Коммиты: 7150799, f943e60, 7008542, 0870f50 (фазовая ветка), f1f0019, 2c00c10 (fix-ветка, в main через 58b07cc) — все FOUND в git log
- Верификация: go test -race ./... зелёный на post-fix main; golangci-lint 0 issues; go.mod/go.sum нетронуты; systemctl is-active goswitchd = active; grep-гейты SPEC/CONFIG зелёные
- Commits measured: git rev-list --count gsd-plan-head-before-11-01..HEAD = 7

---
*Phase: 11-menu-punkt-v-chernyy-spisok*
*Completed: 2026-10-31*
