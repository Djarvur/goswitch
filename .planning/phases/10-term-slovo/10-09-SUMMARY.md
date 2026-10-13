---
phase: 10-term-slovo
plan: 09
subsystem: host-closure-validation
tags: [gap-closure, host-install, g-10-1, closure-validation, human-check-pending, owner-reacceptance, read-only-host]

requires:
  - phase: 10-term-slovo plan 07
    provides: флап-стойкая классификация verifyDead + identity-якорь [gnome-terminal, wezterm] + ac_role_errors (причина A, в сборке)
  - phase: 10-term-slovo plan 08
    provides: defer-until-confirmed доставка commit + stale-cache амендмент (причина B, в сборке)
  - phase: 10-term-slovo plan 06
    provides: процесс установки на хост (build → release-dir → goswitchctl install), точка отката v1.1.1, форма human-check 10-UAT
  - phase: 10-term-slovo UAT
    provides: разрыв G-10-1 (недостающий элемент 5 — host-ревалидация), Current Test форма приёмки
provides:
  - Gap-сборка закрытия G-10-1 v1.1.1-124-g7ace3fd (HEAD 7ace3fd: оба фикса 10-07+10-08) установлена и активна на живом столе владельца (release-dir, unit re-point, откат v1.1.1 нетронут); датированный ряд host-install-gap в adr-008-spike-log.md
  - Фиксированный путь goswitchctl для всех машинных проверок хоста (не на PATH — новейший releases по mtime): /home/nil/.local/share/goswitch/releases/v1.1.1-124-g7ace3fd/goswitchctl
  - Машинная предпроверка хоста зелёная: unit active, чистая регистрация движка (component registered engines:2, ошибок нет), status config_valid=true, счётчик role-ошибок 10-07 виден (ac_role_errors=0) — тихий эпизод 11:00:16 более не воспроизводим незаметно; после установки хост строго read-only
  - Секция «## Closure validation (G-10-1)» в DEBUG-term-host-live-failures.md: машинные предпроверки с датами, шестипунктовая инструкция re-acceptance (5 исходных пунктов + флап-бурст + молчание автокоррекции), критерии журнальных рядов приёмки (нет verify-timeout/mismatch, done changed=true не потерян по E3, флипы строго после settlement), вердикт — честный ОЖИДАНИЕ (не синтезируется)
affects: [phase-verification (G-10-1 закрывается ТОЛЬКО вердиктом владельца — маршрутизируется через UAT end-of-phase; rejected = СТОП без кода), milestone-harvest (owner-зачёт критерия 1)]

actuals:
  tokens: 8520         # chars/4 over the realized diff 7ace3fd..HEAD (34081 chars, 2 files, +352)
  tasks: 2
  commits: 2           # MEASURED: git rev-list --count gsd-plan-head-before-10-09..HEAD (df0c9e3, b90498f)

tech-stack:
  added: []
  patterns:
    - "Установка gap-сборки — существующий процесс 10-06: build из рабочего дерева в НОВЫЙ release-dir (путь отката сохранён нетронутым), фикс. путь ctl записывается в ряд установки ДО всякой проверки"
    - "Closure validation как жанр секции debug-файла: машинные предпроверки (read-only) + инструкция human-check + критерии журнальных рядов + вердикт-плейсхолдер «не синтезируется» — приёмка и её свидетельства разнесены по ролям исполнителя"

key-files:
  created:
    - ".planning/phases/10-term-slovo/10-09-SUMMARY.md"
  modified:
    - "docs/adr/adr-008-spike-log.md (секция «Фаза 10 host-ревалидация G-10-1 (план 10-09)»: ряд host-install-gap 2026-10-14 00:46 — версия/коммит, маркер «gap G-10-1 closure build», фикс. путь ctl, откат)"
    - ".planning/debug/DEBUG-term-host-live-failures.md (секция «## Closure validation (G-10-1)»: машинные предпроверки 00:46–00:49 MSK, шестипунктовая инструкция re-acceptance, критерии журнальных рядов, вердикт ОЖИДАНИЕ; frontmatter updated)"

key-decisions:
  - "Установка — существующим процессом 10-06 в НОВЫЙ release-dir v1.1.1-124-g7ace3fd (не перезапись бинарника старого каталога): откат ~/.local/share/goswitch/releases/v1.1.1/goswitchctl install сохранён нетронутым и назван в ряду ДО human-check (T-10-09-01)"
  - "Фикс. путь goswitchctl резолвится ОДИН раз и вписывается в ряд установки — все машинные проверки зовут только его (воспроизводимость рядов, нет зависимости от PATH/mtime на момент проверки)"
  - "Вердикт re-acceptance НЕ синтезируется: машинные предпроверки фиксируют видимость диагностики, но не заменяют приёмку (T-10-09-02); Closure validation несёт честный ОЖИДАНИЕ, вердикт маршрутируется владельцу через UAT end-of-phase; при approved — ряд закрытия G-10-1, при rejected — СТОП без правок кода, откат по запросу"
  - "Хост после установки строго read-only: инъекций нет, окон нет, чужие сервисы не тронуты, ydotool не запускался — все жесты приёмки исполняет владелец (prohibitions плана соблюдены)"

patterns-established:
  - "Closure validation (G-10-1) — формат закрытия host-разрывов: установка с откатом-до-приёмки → read-only машинные предпроверки → инструкция владельца с флап-пунктом → критерии журнальных рядов → вердикт-плейсхолдер; владелец — единственный валидный оракул"

requirements-completed: [TERM-02]  # shared-ID gate: уже Complete (открыт 10-06, checkbox [x] в REQUIREMENTS.md) — mark-complete переисполнен идемпотентно, write_set_complete: false

coverage:
  - id: D1
    description: "Хост несёт gap-сборку (оба фикса 10-07+10-08), unit active, версия/коммит записаны датированным рядом host-install-gap 2026-10-14 00:46, откат-путь (reinstall v1.1.1) назван ДО human-check"
    requirement: TERM-02
    verification:
      - kind: other
        ref: "systemctl --user is-active goswitchd = active; ExecStart=…/releases/v1.1.1-124-g7ace3fd/goswitchd (старт 00:46:57 MSK); journal 00:46:57 component registered (engines:2), ошибок регистрации нет; docs/adr/adr-008-spike-log.md ряд host-install-gap (маркер «gap G-10-1 closure build», HEAD 7ace3fd); mise run ci зелёный на HEAD перед установкой"
        status: pass
    human_judgment: false
  - id: D2
    description: "Машинная предпроверка хоста зелёная (read-only): goswitchctl status по фикс. пути парсится, config_valid=true, счётчик role-ошибок 10-07 присутствует (ac_role_errors=0); autocorrect_fired=0 на старте; хост после установки строго read-only"
    requirement: TERM-02
    verification:
      - kind: other
        ref: "/home/nil/.local/share/goswitch/releases/v1.1.1-124-g7ace3fd/goswitchctl status → config_valid=true config_error= (пуст), ac_role_errors=0, autocorrect_fired=0 (ac_skip_* — только при ненулевых причинах, формат 10-07 пиннут ctlsvc-тестом); секция Closure validation debug-файла, предпроверки 00:46–00:49 MSK"
        status: pass
    human_judgment: false
  - id: D3
    description: "Секция «## Closure validation (G-10-1)» в debug-файле: машинные предпроверки с датами, шестипунктовая инструкция re-acceptance владельца (5 исходных пунктов 10-UAT + флап-бурст + молчание автокоррекции с машинной проверкой), критерии журнальных рядов приёмки, путь отката в инструкции"
    requirement: TERM-02
    verification:
      - kind: other
        ref: "grep 'Closure validation' и grep 'gap G-10-1 closure build' — PASS (verify Task 1 целиком зелёный); grep 're-acceptance' — PASS; инструкция содержит все 6 пунктов и откат-путь"
        status: pass
    human_judgment: false
  - id: D4
    description: "Re-acceptance владельца: жесты на живом столе по шестипунктовой инструкции, журнальные ряды приёмки, вердикт approved/rejected — закрывает/не закрывает G-10-1"
    requirement: TERM-02
    verification:
      - kind: other
        ref: "Вердикт — ОЖИДАНИЕ на момент авторинга SUMMARY: никогда не синтезируется (human_check_mode плана + T-10-09-02). Автоматизированный grep verify Task 2 'approved|rejected' совпадает С ТЕКСТОМ ИНСТРУКЦИИ/ОЖИДАНИЯ, не с записанным вердиктом — честный статус: unrun до ответа владельца. Routing: UAT end-of-phase (workflow.human_verify_mode); запись дословного вердикта в Closure validation — за сессией приёмки"
        status: unknown
    human_judgment: true
    rationale: "Жест владельца на живом столе — закрывающий оракул G-10-1: стенд принципиально не воспроизводит хостовую флап-тряску ядра 7.2.9-2-t2 (E1/E2b), машинные проверки видимость диагностики фиксируют, но приёмку не заменяют; вердикт доставлен владельцу и ждёт ответа"
  - id: D5
    description: "Flagged assumption плана: формальная классификация края TERM-02 остаётся unresolved (SPEC без ## Edge Coverage); край host-ревалидации выражен human-check-пунктами Task 2; формальный вердикт — за фазовой верификацией"
    requirement: TERM-02
    verification:
      - kind: other
        ref: ".planning/phases/10-term-slovo/10-09-PLAN.md — Flagged Assumptions: {\"requirement_id\":\"TERM-02\",\"category\":\"unclassified\",\"status\":\"unresolved\"}"
        status: unknown
    human_judgment: true
    rationale: "Flagged assumption никогда не авто-разрешается: verifier обязан классифицировать край явно, не тихий pass (прецедент 10-04/10-06/10-07/10-08)"

duration: 8min
completed: 2026-10-14
status: complete
---

# Phase 10 Plan 09: host-ревалидация G-10-1 — gap-сборка на хосте, машинные предпроверки, re-acceptance владельца Summary

**Gap-сборка закрытия G-10-1 (v1.1.1-124-g7ace3fd = HEAD 7ace3fd: флап-стойкая классификация 10-07 + defer-until-confirmed доставка 10-08) установлена и активна на живом столе владельца существующим процессом 10-06 — откат-путь v1.1.1 назван до приёмки, фикс. путь goswitchctl записан в ряд установки; машинные предпроверки зелёные строго read-only (unit active, чистая регистрация, config_valid=true, ac_role_errors=0 — диагностика 10-07 машинно видна); секция Closure validation (G-10-1) открыта в debug-файле с шестипунктовой инструкцией re-acceptance и критериями журнальных рядов — вердикт владельца честно ОЖИДАЕТ (human_judgment, маршрутизация через UAT end-of-phase): разрыв закрывается ТОЛЬКО его ответом.**

## Performance

- **Duration:** 8 min
- **Started:** 2026-10-13T21:44:09Z
- **Completed:** 2026-10-14T00:52:49+03:00 (21:52:49Z)
- **Tasks:** 2/2 (Task 1 auto — установка gap-сборки + машинные предпроверки + датированные ряды; Task 2 auto+human-check — инструкция re-acceptance доставлена, вердикт ОЖИДАНИЕ, human_judgment: true)
- **Files modified:** 2 репозиторных (docs/adr/adr-008-spike-log.md, .planning/debug/DEBUG-term-host-live-failures.md — обе из files_modified frontmatter) — measured diff 7ace3fd..HEAD: 2 files, +352

## Accomplishments

- **Установка (10-06 процесс):** build goswitchd+goswitchctl из рабочего дерева (HEAD 7ace3fd, mise run ci зелёный; оба корневых фикса G-10-1) → release-dir v1.1.1-124-g7ace3fd → `goswitchctl install` (unit re-point, ibus restart, sources wrapped goswitch-en/goswitch-ru, engine activated goswitch-en). Unit active, ExecStart указывает на новую сборку (старт 00:46:57 MSK), журнал поднятия чистый: config loaded → connected → engine created goswitch-en → component registered (engines:2) → ctl service listening → engine reactivated.
- **Откат-путь назван ДО приёмки:** `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install` — релизный каталог v1.1.1 нетронут (release-dir не перезаписывался); записан в ряд установки, повторён в инструкции владельца (T-10-09-01 mitigate).
- **Фикс. путь машинных проверок:** `/home/nil/.local/share/goswitch/releases/v1.1.1-124-g7ace3fd/goswitchctl` (не на PATH; резолв новейшим по mtime среди releases) — вписан в ряд установки, все проверки зовут только его.
- **Машинные предпроверки (read-only):** status парсится, config_valid=true, config_error пуст; счётчик role-ошибок 10-07 присутствует (ac_role_errors=0) — тихий эпизод 11:00:16 более не воспроизводим незаметно (key_links плана); autocorrect_fired=0 на старте. После установки хост строго read-only: инъекций нет, окон нет, чужие сервисы не тронуты, ydotool не запускался.
- **Closure validation (G-10-1):** секция в DEBUG-term-host-live-failures.md — предпроверки с датами, шестипунктовая инструкция re-acceptance (ghbdtn→привет в wezterm и GNOME Terminal; флап-бурст «пощёлкать фокусами и СРАЗУ жест» — проверка фикса 10-07 в её сценарии, воспроизводит тряску 16:20:05; середина фразы; «ytn?»→«нет,»; молчание автокоррекции с машинной проверкой autocorrect_fired), критерии журнальных рядов приёмки (нет verify-timeout/mismatch skip-рядов; done changed=true не потерян по E3 — следующий push трекает исправленное слово; флип-рубежи строго после settlement-рубежей).
- **Датированный ряд в spike-log:** секция «Фаза 10 host-ревалидация G-10-1 (план 10-09)», ряд host-install-gap 2026-10-14 00:46 с маркером «gap G-10-1 closure build», версией/коммитом, фикс. путём ctl и откатом.
- Строгая итерация: mise run ci зелёный на HEAD до и после установки (задачи плана — install/docs, кода не пишут; go.mod/go.sum нетронуты).

## Task Commits

1. **Task 1: Установка gap-сборки на хост + машинные предпроверки** - `df0c9e3` (docs: host-install-gap ряд + Closure validation секция с предпроверками; verify зелёный целиком)
2. **Task 2: Owner re-acceptance — приёмка критерия 1 (human-check)** - `b90498f` (docs: шестипунктовая инструкция re-acceptance + критерии журнальных рядов + вердикт ОЖИДАНИЕ, не синтезируется; маршрутизация UAT end-of-phase)

**Plan metadata:** финальный docs-коммит (SUMMARY+STATE+ROADMAP)

## Files Created/Modified

- `docs/adr/adr-008-spike-log.md` — секция «Фаза 10 host-ревалидация G-10-1 (план 10-09)»: ряд host-install-gap (версия/коммит/маркер/фикс.путь/откат +Verdict «установлено, re-acceptance за владельцем»)
- `.planning/debug/DEBUG-term-host-live-failures.md` — секция «## Closure validation (G-10-1)»: установленная сборка, фикс. путь, откат ДО приёмки, машинные предпроверки 00:46–00:49 MSK, инструкция re-acceptance (6 пунктов), критерии журнальных рядов, вердикт-плейсхолдер ОЖИДАНИЕ + routing верификатора; frontmatter updated

## Decisions Made

См. key-decisions. Ядро: установка существующим процессом в НОВЫЙ release-dir (откат нетронут, назван до приёмки); фикс. путь ctl в ряду установки; вердикт не синтезируется — честный ОЖИДАНИЕ с routing через UAT end-of-phase (T-10-09-02); хост строго read-only после установки (prohibitions).

## Deviations from Plan

None — plan executed exactly as written. (Код задачами не писался: install/documentation план; поведение — сборка уже закоммиченных фиксов 10-07/10-08.)

## Issues Encountered

- **Task 2 human-check — вердикт ОЖИДАНИЕ (не отклонение):** план и human_check_mode запрещают синтез вердикта владельца. Автоматизированный verify Task 2 технически матчится (`approved|rejected` встречаются в тексте инструкции/ОЖИДАНИЯ), но записанного вердикта НЕТ — честный статус: unrun до ответа владельца; задокументировано в coverage D4 (human_judgment: true, status: unknown) и в секции debug-файла. Routing: владелец получает инструкцию через UAT на end-of-phase; при approved — ряд закрытия G-10-1 дописывается в Closure validation; при rejected — СТОП без правок кода (план фиксирует свидетельства, не чинит), откат по запросу немедленно.
- Счётчики демона на момент установки были нулевыми (daemon перезапускался после UAT-эпизода 16:19–16:23): база сравнения для приёмки чистая — любые skip-ряды приёмки будут видны с нуля.
- `state advance-plan` первый вызов вернул plans_outstanding (SUMMARY ещё не существовал) — переисполнен после Write SUMMARY (порядок create_summary → update_tracking).

## User Setup Required

**Owner re-acceptance (human-check закрытия G-10-1) — на живом столе, gap-сборка активна:**
1. wezterm: набрать «ghbdtn», двойной правый Shift → увидеть «привет».
2. GNOME Terminal: то же.
3. Флап-бурст: быстро пощёлкать фокус между окнами туда-сюда и СРАЗУ жест в терминале — исправление срабатывает.
4. Слово в середине фразы исправляется жестом.
5. «ytn?» + жест → «нет,».
6. Автокоррекция молчала во время всех проб (`.../v1.1.1-124-g7ace3fd/goswitchctl status` — autocorrect_fired не вырос).
Ответ «approved» или описание отклонений. Откат по запросу: `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install`.

## Next Phase Readiness

- Фаза 10 исполнена целиком (9/9 планов): причины A и B разрыва G-10-1 закрыты в коде и доказаны юнит-корпусами + байт-точными стенд-рядами; хост несёт сборку обоих фиксов. Оставшийся хвост — вердикт re-acceptance владельца (D4) и формальная классификация края TERM-02 (D5) — оба human_judgment для фазовой верификации/UAT harvest.
- Верификатору: Closure validation — готовая форма для записи вердикта (критерии журнальных рядов уже сформулированы); при approved дописать ряд «gap closed by owner re-acceptance, дата»; при rejected — STOP-рубеж, код не правится вне нового цикла планирования.
- Блокеров нет; TERM-02 остаётся Complete (shared-ID gate, mark-complete идемпотентен: write_set_complete: false).

---
*Phase: 10-term-slovo*
*Completed: 2026-10-14*
