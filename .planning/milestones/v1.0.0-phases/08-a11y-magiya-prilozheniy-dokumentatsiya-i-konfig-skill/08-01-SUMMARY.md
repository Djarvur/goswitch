---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
plan: 01
subsystem: docs
tags: [spec-delta, a11y, config-schema, gsettings, dbus, gnome, toolkit-accessibility]

# Dependency graph
requires:
  - phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
    provides: D-53 blocklist-ревизия §11 (форма датированных ревизий), D-54 zero-value-off прецедент, install snapshot-дисциплина install-state.json
provides:
  - "SPEC §4 — датированная ревизия 2026-10-05: контракт секции a11y (a11y.enabled/a11y.apps, RE2, потолок 64, zero-value off)"
  - "Контракт двух ручек магии: gsettings org.gnome.desktop.interface toolkit-accessibility + пояс org.a11y.Status.IsEnabled=true (D-Bus Properties.Set)"
  - "Демон-reconcile (старт + hot reload, D-8-3) и restart-семантика «подхватят при следующем запуске»"
  - "Откат только через goswitchctl uninstall из install-state.json (restore only-if-present, D-8-4); динамическое снятие ключа исключено"
  - "argv-дисциплина ASVS V5 и резолюция per-app оверрайдов: документируется, вне v1-дифа (рецепт — docs/CONFIG.md, план 08-06)"
affects: [08-02 (схема a11y), 08-03 (демон-reconcile), 08-04 (snapshot/revert), 08-05 (wave-downstream), 08-06 (docs/CONFIG.md)]

# Actuals (#2632)
actuals:
  tokens: 1535    # chars/4 over realized diff (6141 chars added in docs/SPEC.md)
  tasks: 2
  commits: 1      # MEASURED: git rev-list --count gsd-plan-head-before-08-01..HEAD at SUMMARY time (task 2 is verification-only)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Спека прежде кода (D-55): датированная ревизия §4 добавлением — волна-гейт для кодовых планов"
    - "Аудит-трейл: HTML-комментарий-пометка + датированный абзац, прежние ревизии не трогаются (форма §4.1/§11)"

key-files:
  created:
    - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/deferred-items.md
  modified:
    - docs/SPEC.md

key-decisions:
  - "a11y.enabled / a11y.apps — канонические имена секции (discretion планировщика по D-8-1 «секция типа a11y.apps»)"
  - "Ревизия консервативна по A1: демон ставит ОБЕ ручки (ключ + пояс), контракт верен при любой композиции стартовых проверок Chromium/Electron"
  - "Spec-less fallback пропущен ВИДИМО: у фазы нет requirement-ID для probe-падения; источник требований — D-8-1..D-8-11 из 08-CONTEXT.md; пробы не запускались, покрытие дают истины планов и эта ревизия"

patterns-established:
  - "Ревизия §4 только добавляет датированный блок; прежние ревизии (flip_after_correction, D-53) не удаляются и не переписываются"

requirements-completed: ["D-8-1", "D-8-2", "D-8-3", "D-8-4", "D-8-6"]

coverage:
  - id: D1
    description: "SPEC §4 несёт датированную ревизию 2026-10-05 с полным контрактом a11y-магии (секция enabled/apps/потолок 64/zero-off, две ручки, reconcile старт+hot reload, restart-семантика, uninstall-revert only-if-present, argv-дисциплина, резолюция оверрайдов); аудит-трейл цел (только добавления, §10/§11 не тронуты)"
    requirement: "D-8-1"
    verification:
      - kind: other
        ref: "плановый гейт 1: grep 2026-10-05/toolkit-accessibility/a11y/install-state docs/SPEC.md → S4-OK"
        status: pass
      - kind: other
        ref: "плановый гейт 2: git diff HEAD --numstat -- docs/SPEC.md → 62 insertions, 0 deletions → ONLY-ADDITIONS"
        status: pass
      - kind: other
        ref: "плановый гейт 3: git diff HEAD --name-only | grep .go$ → пусто (ноль Go-файлов в дифе)"
        status: pass
    human_judgment: true
    rationale: "Ревизия — прозаический контракт, на который кодовые планы 08-02..08-04 строят реализацию; соответствие дословным решениям владельца (D-8-1/D-8-2) и качество русской прозы требует человеческой сверки на verify-гейте"
  - id: D2
    description: "Зелёная итерация mise run ci достигнута на docs-only дереве (build + vet + lint 0 issues + test -race, exit 0)"
    verification:
      - kind: other
        ref: "mise run ci → EXIT=0"
        status: pass
    human_judgment: false

# Metrics
duration: 11min
completed: 2026-10-05
status: complete
---

# Phase 8 Plan 01: SPEC §4 ревизия (a11y-магия приложений) Summary

**Spec-delta до кода: датированная ревизия §4 SPEC (2026-10-05) фиксирует полный контракт a11y-магии — секция a11y.enabled/a11y.apps, две ручки (toolkit-accessibility + IsEnabled), демон-reconcile, restart-семантика, uninstall-revert из install-state.json, argv-дисциплина.**

## Performance

- **Duration:** 11 min
- **Started:** 2026-10-05T16:16:19Z
- **Completed:** 2026-10-05T16:27:30Z
- **Tasks:** 2
- **Files modified:** 1 (docs/SPEC.md, +62 строки) + 1 planning-артефакт (deferred-items.md)

## Accomplishments

- §4 SPEC несёт датированную ревизию 2026-10-05 «a11y-магия приложений» в форме §4.1/§11 (HTML-комментарий-пометка + датированные абзацы) со всем контрактом: секция `a11y` (`a11y.enabled` default off, `a11y.apps` — RE2 regex, потолок 64, анкеровка явная, валидация на Load, zero-value = off), две ручки магии, демон-reconcile (старт + hot reload), «приложение подхватит при следующем запуске», откат только через uninstall из install-state.json (restore only-if-present), argv-дисциплина, резолюция «пер-app оверрайды — документируется, вне v1-дифа»
- Аудит-трейл цел: 62 вставки, 0 удалений; прежние ревизии §4/§11 (flip_after_correction, D-53, звуки) дословно на месте; §10/§11 не тронуты; ноль .go в дифе (D-55 соблюдён)
- Волна-гейт фазы открыт: кодовые планы 08-02 (схема), 08-03 (демон-reconcile), 08-04 (snapshot/revert) строят на записанном контракте, а не на интерпретации research
- Зелёная итерация `mise run ci` (build + vet + lint 0 issues + test -race) достигнута на docs-only дереве — прецедент 07-01 подтверждён

## Task Commits

Each task was committed atomically:

1. **Task 1: SPEC §4 — датированная ревизия (a11y-магия, 2026-10-05)** - `5293daa` (docs)
2. **Task 2: Зелёная итерация на docs-only дереве** - без коммита: verify-only задача (`mise run ci`), изменений дерева не создаёт — коммитить нечего

**Plan metadata:** отдельный metadata-коммит (SUMMARY + STATE + ROADMAP) следует за этим файлом.

_Note: TDD-гейты не применялись — план type: execute, docs-only._

## Files Created/Modified

- `docs/SPEC.md` — §4: ревизия 2026-10-05, контракт a11y-магии (только добавления, +62)
- `.planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/deferred-items.md` — журнал out-of-scope находок (два pre-existing флаки тестов)

## Decisions Made

- `a11y.enabled` / `a11y.apps` зафиксированы как канонические имена (D-8-1 давал discretion «секция типа a11y.apps»)
- По A1 ревизия консервативна: демон ставит ОБЕ ручки (постоянный ключ + session-scoped пояс), поэтому контракт верен при любой композиции трёх стартовых проверок Chromium; uninstall восстанавливает только ключ — пояс снапшота не требует (A5)
- Restart-семантика записана как семантика механизма («Chromium/Electron проверяют ключ однократно при старте, без подписки»), не как ограничение
- **Spec-less fallback пропущен ВИДИМО** (записанное решение must_haves): у фазы нет requirement-ID для probe-падения — источник требований: D-8-1..D-8-11 из 08-CONTEXT.md + goal ROADMAP; пробы не запускались, покрытие даётся истинами планов и этой ревизией

## Deviations from Plan

None - plan executed exactly as written.

---

**Total deviations:** 0 auto-fixed.
**Impact on plan:** нет.

## Issues Encountered

- **Pre-existing flaky tests при прогоне `mise run ci` (out of scope, не чинились).** Первые прогоны падали на разных тестах: `internal/ctlsvc TestRun_OnConnHookCalledOnce` (гонка владения D-Bus-именем `org.djarvur.goswitch` с ЖИВЫМ демоном владельца на реальной session bus — лог «already registered reply=2»; standalone проходит) и `internal/config TestWatch_BrokenAutocorrectBlockKeepsLastGood` (debounce-тайминг 200 мс под `-race` при параллельных пакетах). Оба не связаны с планом (диф — 62 строки markdown, ноль .go); оба проходят на повторных прогонах. Финальный `mise run ci` — зелёный (exit 0). Задокументировано в `deferred-items.md` фазы с suggested follow-up.
- Plan-level verification: все три автоматических гейта плана прошли (S4-OK; ONLY-ADDITIONS 62/0; ноль .go); гейты «GIT-DIFF-FAILED» не сработали — git доступен, гейты вычислены.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Контракт a11y-магии записан в спеке — блокировка 08-02/08-03/08-04 снята (wave 2)
- 08-06 (docs/CONFIG.md) получит из ревизии якорь для рецепта per-app оверрайдов
- Открытых блокеров нет; pre-existing флаки тестов — в deferred-items.md, не блокируют фазу

---
*Phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill*
*Completed: 2026-10-05*

## Self-Check: PASSED

- docs/SPEC.md — FOUND (ревизия §4 в дереве; commit 5293daa — FOUND in git log)
- 08-01-SUMMARY.md — FOUND в каталоге фазы
- deferred-items.md — FOUND в каталоге фазы
- commits с ledger gsd-plan-head-before-08-01: 1 (5293daa, docs)

