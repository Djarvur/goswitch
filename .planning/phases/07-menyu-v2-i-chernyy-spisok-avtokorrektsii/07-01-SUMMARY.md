---
phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
plan: 01
subsystem: docs
tags: [spec-delta, adr, blocklist, autocorrect, dbus-menu, sound, d53, d55]

# Dependency graph
requires:
  - phase: 06-sistema-avtokorrektsii
    provides: SPEC §11 с двумя прежними ревизиями (вердикт-отказ + D-51), ADR-007 Accepted (D-52/D-53/D-54/D-55), форма датированного append §4.1
provides:
  - SPEC §11 — третья датированная ревизия (2026-10-04, ревизия D-53): blocklist-семантика (fire ⇔ enabled ∧ role-гейт ∧ NOT apps_blocklist; unknown-идентичность НЕ запрет; confirm blocklist-only, payload-stale; ручная коррекция вне blocklist; default off D-54) + решение «Звуки при переключении» (секция sound, enabled default ON, отдельный тон автокоррекции, canberra/paplay субпроцесс, WARN-best-effort, тумблер «Звук»)
  - ADR-007 amendment «Amendment (2026-10-04, фаза 07 — ревизия D-53: white-list → blocklist)»: контекст с дословными решениями владельца, вердикт-таблица, следствия (слаг app-blocked, миграция CONFIG.md, перезакрепление корпуса), обратимость
  - SPEC-DELTA-OK: волна-гейт снят — кодовые планы 07-02..07-08 строят на записанном контракте
affects: [07-02-schema, 07-03-persist, 07-04-gate, 07-05-menu, 07-06-e2e, 07-08-sound]

# Actuals (#2632)
actuals:
  tokens: 2573        # chars/4 over the realized diff (10294 chars, docs-only)
  tasks: 2
  commits: 3          # MEASURED: git rev-list --count ccf7c80..HEAD (ledger gsd-plan-head-before-07-01)

# Tech tracking
tech-stack:
  added: []           # ноль новых зависимостей — план документальный
  patterns:
    - "Датированный append spec-delta: HTML-комментарий §4.1-формы + абзац «Решение владельца (<дата>, ревизия)» — третья ревизия рядом с прежними (аудит-трейл)"
    - "ADR amendment по прецеденту ADR-005: датированная секция в конце существующего ADR, статус не меняется, новый файл не создаётся"

key-files:
  created: []
  modified:
    - docs/SPEC.md
    - docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md

key-decisions:
  - "Ревизия D-53 зафиксирована в спеке ДО кода (D-55, прецедент 06-01): §11 — третья датированная ревизия + ADR-007 amendment (append, Accepted сохранён, слаг app-blocked); SPEC-DELTA-OK для 07-02..07-07"
  - "Решение «Звуки при переключении» внесено в §11 ревизии 2026-10-04 ДО кода плана 07-08: секция sound (enabled default ON, отдельный тон автокоррекции), субпроцесс canberra-gtk-play (paplay-фолбэк), отказ — только WARN, тумблер «Звук» по каноническому макету 287999c"

patterns-established:
  - "Третья датированная ревизия §11: append-only, оба прежних вердикта байт-нетронуты (прецедент D-51-записи 06-01)"
  - "ADR amendment: датированная секция Контекст/Вердикты/Следствия/Обратимость без нового номера ADR"

requirements-completed: ["SPEC §11 (spec-delta)"]  # дословно из frontmatter плана; не REQ-ID в REQUIREMENTS.md — mark-complete корректно no-op

coverage:
  - id: D1
    description: "SPEC §11 — третья датированная ревизия (2026-10-04, ревизия D-53) с полной locked-семантикой blocklist и решением о звуках"
    requirement: "SPEC §11 (spec-delta)"
    verification:
      - kind: other
        ref: "grep -c 2026-10-04 docs/SPEC.md + grep apps_blocklist → S11-OK"
        status: pass
      - kind: other
        ref: "grep «НЕ запрет» + payload-stale/перенос фокуса → SEMANTICS-OK"
        status: pass
      - kind: other
        ref: "grep sound + canberra + «Звук» → SOUND-REV-OK"
        status: pass
      - kind: other
        ref: "git diff --numstat docs/SPEC.md: 37 insertions / 0 deletions (относительно ccf7c80)"
        status: pass
    human_judgment: false
  - id: D2
    description: "ADR-007 amendment-секция ревизии D-53: Контекст (дословные решения владельца), вердикт-таблица, Следствия (app-blocked, миграция, корпус), Обратимость"
    requirement: "SPEC §11 (spec-delta)"
    verification:
      - kind: other
        ref: "grep «Amendment (2026-10-04» + apps_blocklist + app-blocked → ADR-OK"
        status: pass
      - kind: other
        ref: "grep fail-closed + §11|SPEC → XREF-OK; git diff --numstat: 55 insertions / 0 deletions"
        status: pass
    human_judgment: false
  - id: D3
    description: "Аудит-трейл цел: оба прежних вердикта §11 (2026-09-27-формы) не тронуты; §10-bullet «Wayland-композиторы» count==1; новых ADR-файлов нет; статус ADR-007 остался Accepted"
    verification:
      - kind: other
        ref: "grep «отвергнута владельцем осознанно» + «решение владельца (2026-09-27, D-51)» → TRAIL-OK"
        status: pass
      - kind: other
        ref: "ls docs/adr/ | grep -c ^ADR-008 == 0 → NO-NEW-ADR; git diff ccf7c80..HEAD: 0 deletions в обоих файлах"
        status: pass
    human_judgment: false

# Metrics
duration: 11 min
completed: 2026-10-04
status: complete
---

# Phase 7 Plan 1: SPEC §11 spec-delta + ADR-007 amendment Summary

**Третья датированная ревизия §11 (blocklist D-53, 2026-10-04) и ADR-007 amendment записаны ДО кода — SPEC-DELTA-OK для всех кодовых планов фазы, аудит-трейл обоих прежних вердиктов цел.**

## Performance

- **Duration:** 11 min (20:35–20:46 UTC)
- **Started:** 2026-10-04T20:35:12Z
- **Completed:** 2026-10-04T20:46:24Z
- **Tasks:** 2/2
- **Files modified:** 2 (только docs; ноль .go)

## Accomplishments

- SPEC §11 несёт третью датированную ревизию (2026-10-04, ревизия D-53): полная конъюнкция срабатывания (enabled ∧ role-гейт ∧ surrounding-text ∧ уверен ∧ длина ∧ НЕ apps_blocklist), regex-подстрока с явной анкеровкой, unknown-идентичность НЕ запрет, confirm blocklist-only на ТЕКУЩЕМ приложении (payload-stale), ручная коррекция вне blocklist, default off (D-54) — и ОТДЕЛЬНЫМ пунктом решение «Звуки при переключении» (секция sound, enabled по умолчанию включён, отдельный тон автокоррекции, canberra/paplay, WARN-best-effort, тумблер «Звук» по макету 287999c).
- ADR-007 дополнен amendment-секцией «Amendment (2026-10-04, фаза 07 — ревизия D-53: white-list → blocklist)»: дословные решения владельца из 07-CONTEXT, вердикт-таблица (white-list → blocklist; unknown ⇒ пропуск при fail-closed роли/капсов/детектора; ACTIVE = enabled: true с обязательными порогами; confirm на текущей идентичности, равенство арм/confirm удалено), следствия для кодовых планов (слаг app-blocked, миграция docs/CONFIG.md, перезакрепление silence-matrix и e2e-оракулов), обратимость.
- Аудит-трейл нетронут: вердикт-отказ и ревизия D-51 в §11 байт-сохранены, §10-bullet «Wayland-композиторы» — ровно один, статус ADR-007 остался Accepted, новых ADR-файлов нет; оба дифа — только добавления (37+ и 55+ строк, 0 удалений).
- SPEC-DELTA-OK: волна-гейт фазы снят, кодовые планы 07-02..07-08 строят на записанном контракте (D-55 соблюдён — ноль .go в дифе плана).

## Task Commits

Каждая задача закоммичена атомарно:

1. **Task 1: SPEC §11 — третья датированная ревизия (blocklist, 2026-10-04)** - `d882474` (docs)
2. **Task 2: ADR-007 amendment-секция (ревизия D-53) + перекрёстные греп-гейты** - `ea772c5` (docs)
3. **Фикс переноса строки (см. Deviations)** - `12d73f1` (docs)

**Plan metadata:** см. финальный коммит docs(07-01) ниже.

## Files Created/Modified

- `docs/SPEC.md` — §11: HTML-комментарий §4.1-формы (2026-10-04) + абзац ревизии D-53 + абзац решения о звуках; всё — append после абзаца D-51, перед §12
- `docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md` — amendment-секция в конце (Контекст/Вердикты/Следствия/Обратимость), статус Accepted не менялся

## Decisions Made

- Формулировки ревизии взяты дословно по смыслу из 07-CONTEXT (Locked semantics, Owner Decisions, решение о звуках) — три источника семантики (SPEC ↔ ADR ↔ CONTEXT) согласованы перекрёстными ссылками §11↔ADR-007.
- ADR-007 amendment НЕ выносит решение о звуках (вне предметной области ADR о детекторе/роли) — звуки записаны только в §11, как предписано планом.
- Amendment прямо фиксирует: инвертируется ровно один сегмент конъюнкции D-53 (идентичность), роль/капсы/детектор остаются fail-closed — защита от half-переворота (Pitfall 4 research).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Пин-фраза «не ограничивается» разрывалась переносом строки**
- **Found during:** Task 1 (ре-прогон acceptance criteria)
- **Issue:** собственный перенос строки в добавленном абзаце разделил «не / ограничивается» — семантический пин «ручная коррекция вне blocklist» не грепался одной строкой (будущие гейт-гейты кодовых планов могли не найти формулировку)
- **Fix:** минимальный rewrap — «blocklist
не ограничивается» целиком в строке; содержимое не менялось
- **Files modified:** docs/SPEC.md
- **Verification:** grep «не ограничивается» + «Default off» → AC3-PASS; net diff против ccf7c80 остался additions-only (37+/0-)
- **Committed in:** 12d73f1

---

**Total deviations:** 1 auto-fixed (1 blocking — форматирование собственного текста задачи)
**Impact on plan:** косметическое; ни семантики, ни объёма не меняет.

## Issues Encountered

- `state update-progress` корректно пропущен вербом («phase scope is unscoped, not complete») — не ошибка.
- `requirements mark-complete` корректно no-op: требование плана «SPEC §11 (spec-delta)» — свободная формулировка, не REQ-ID из REQUIREMENTS.md (верб разделил строку и не нашёл ID); requirements-completed в frontmatter SUMMARY скопированы дословно.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- SPEC-DELTA-OK: план 07-02 (схема apps_blocklist + механическая адаптация актора + docs/CONFIG.md) может стартовать — контракт §11/ADR-007 записан, слаг app-blocked запинен, семантика ACTIVE-условия и confirm-гейта зафиксирована в вердикт-таблице amendment'а.
- Аудит-трейл для верификатора: оба прежних вердикта §11 читаются рядом с новой ревизией; git-история правок — d882474 → ea772c5 → 12d73f1.
- Зелёная итерация подтверждена: `mise run ci` (build + vet + golangci-lint + test -race) — все пакеты ok на дереве с правками.

## Self-Check: PASSED

- FOUND: docs/SPEC.md; FOUND: docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md
- FOUND commits: d882474, ea772c5, 12d73f1 (git log --all)
- Все acceptance criteria обеих задач re-run: AC1..AC6 PASS (греп-гейты S11-OK / TRAIL-OK / SEMANTICS-OK / SOUND-REV-OK / ADR-OK / NO-NEW-ADR / XREF-OK)
- План-level verification: оба дифа additions-only; 7 пунктов Locked semantics 07-CONTEXT покрыты формулировками ревизии; ноль .go файлов; `mise run ci` зелёный

---
*Phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii*
*Completed: 2026-10-04*
