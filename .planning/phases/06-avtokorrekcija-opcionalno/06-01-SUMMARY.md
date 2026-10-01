---
phase: 06-avtokorrekcija-opcionalno
plan: 01
subsystem: docs
tags: [spec-delta, adr, autocorrect, milestone-v1.1, at-spi, fail-closed, licensing]

# Dependency graph
requires:
  - phase: 05-integratsiya-s-gnome
    provides: ADR-006-two-engine-revision (номер 006 занят — ADR-007 ренумерован с 006); ADR-формат-эталон ADR-005
provides:
  - SPEC §2 пункт 3 — автокоррекция как возвращённое содержимое milestone v1.1.0 (D-51: opt-in, default off, white/black list), указатели на §11 и ADR-007
  - SPEC §10 — запрещающий буллет автокоррекции заменён датированным указателем (решение владельца 2026-09-27); три остальных буллета нетронуты
  - SPEC §11 — старый вердикт дословно (аудит-трейл) + датированное решение D-51 в стиле §4.1
  - ADR-007 (Proposed) — механика: гибридный детектор (hunspell + своя триграммная мини-модель), живой GetRole вместо кэша, fail-closed, лицензионная позиция доноров — контракт для кодовых планов 06-02..06-07
affects: [06-02 dictgen, 06-03 appid Role, 06-04 config, 06-05 detect, 06-06 actor hook, 06-07 e2e, 06-08 acceptance (перевод ADR-007 в Accepted)]

# Actuals (#2632)
actuals:
  tokens: 5004
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "spec-delta с аудит-трейлом: старый вердикт сохраняется дословно, датированное решение добавляется рядом в стиле §4.1 (HTML-комментарий)"
    - "ADR kill-таблица «кандидат | вердикт | обоснование» по формату ADR-005"

key-files:
  created:
    - docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md
  modified:
    - docs/SPEC.md

key-decisions:
  - "D-55 соблюдён: спека правится ДО кода детектора — в дифе плана ноль .go-файлов"
  - "ADR-007 ренумерован 006→007 (006 занят ревизией ADR-001 Фазы 5); ADR-006 не изменён"
  - "Милстоун-переключатели (PROJECT.md/REQUIREMENTS.md/VERSION) не тронуты — D-51: артефакты v1.1 ведутся аддитивно до complete-milestone v1.0.0"

patterns-established:
  - "Spec-delta аудита-трейла: держать старый текст видимым, новое решение добавлять датированным HTML-комментарием + абзацем (прецедент §4.1)"
  - "ADR-007 как контракт фазы: кодовые планы ссылаются на kill-таблицу, fail-closed и лицензионную позицию"

requirements-completed: []  # frontmatter plan carries free-text AC labels, not REQUIREMENTS.md IDs — see Deviations/Notes

coverage:
  - id: D1
    description: "SPEC §2 пункт 3 переформулирован: автокоррекция — возвращённое содержимое milestone v1.1.0 (D-51), формулировка «явно ВНЕ объёма» удалена"
    requirement: "SPEC §10/§11 (spec-delta)"
    verification:
      - kind: other
        ref: "grep -q v1.1.0 && grep -q 2026-09-27 && grep -q ADR-007 && ! grep -q 'явно ВНЕ объёма первой версии' docs/SPEC.md → SPEC-DELTA-OK"
        status: pass
    human_judgment: false
  - id: D2
    description: "SPEC §11: старый вердикт «отвергнута владельцем осознанно…» сохранён дословно, датированное решение D-51 добавлено рядом"
    requirement: "SPEC §10/§11 (spec-delta)"
    verification:
      - kind: other
        ref: "grep 'отвергнута владельцем осознанно' docs/SPEC.md → AUDIT-TRAIL-OK (T-06-01-01 mitigate)"
        status: pass
    human_judgment: false
  - id: D3
    description: "SPEC §10: запрещающий буллет заменён датированным указателем; три остальных буллета нетронуты (Wayland-композиторы — ровно 1 вхождение)"
    requirement: "SPEC §10/§11 (spec-delta)"
    verification:
      - kind: other
        ref: "grep -c ' Wayland-композиторы кроме GNOME' docs/SPEC.md = 1 → AUDIT-TRAIL-OK"
        status: pass
    human_judgment: false
  - id: D4
    description: "ADR-007 создан (Proposed): 5 секций ADR-формата, дословные D-52..D-55, kill-таблица (hunspell / своя модель / живой GetRole / fail-closed / лицензии доноров), GTK3-факт роль 61, честная поправка Easy Switcher (Q1b)"
    requirement: "MACR-ACL (app×role политика)"
    verification:
      - kind: other
        ref: "grep 5 секций + Proposed → ADR-SKELETON-OK; grep D-52/D-53/GetRole/whatlanggo/fail-closed/Easy Switcher/hunspell/61 → ADR-CONTENT-OK"
        status: pass
    human_judgment: true
    rationale: "Полнота и точность дословных цитат решений владельца и лицензионной формулировки — суждение владельца на гейте verify-work фазы (перевод ADR-007 в Accepted — план 06-08)"
  - id: D5
    description: "Гейты неизменности: ADR-006 и формальные милстоун-переключатели не тронуты (D-51 freeze)"
    requirement: "SPEC §10/§11 (spec-delta)"
    verification:
      - kind: other
        ref: "git diff --exit-code -- docs/adr/ADR-006-two-engine-revision.md .planning/PROJECT.md .planning/REQUIREMENTS.md (пусто); .planning/STATE.md diff байт-в-байт равен предсуществующей правке позиции оркестратора → SCOPE-FROZEN"
        status: pass
    human_judgment: false

# Metrics
duration: 6min
completed: 2026-10-01
status: complete
---

# Phase 6 Plan 1: Spec-delta + ADR-007 Summary

**SPEC §2/§10/§11 переведены под решение D-51 (автокоррекция = milestone v1.1.0, opt-in, default off) с сохранённым аудит-трейлом; ADR-007 (Proposed, ренумерован с 006) фиксирует механику: гибридный детектор, живой GetRole, fail-closed, лицензионная позиция доноров.**

## Performance

- **Duration:** 6 min
- **Started:** 2026-10-01T20:07:19Z
- **Completed:** 2026-10-01T20:13:58Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Spec-delta §2/§10/§11 по D-55 выполнен ДО кода детектора: пункт 3 §2 больше не «вне объёма» — он содержимое milestone v1.1.0 на предписанных §11 условиях; §10 без запрещающего буллета (три остальных нетронуты); §11 несёт старый вердикт дословно + датированное решение владельца 2026-09-27 в стиле §4.1
- ADR-007-autocorrect-hybrid-detector-and-role-policy.md создан в формате ADR-005, Status Proposed: дословные D-52..D-55 в Context, kill-таблица Decision (словарный путь hunspell / своя триграммная мини-модель vs whatlanggo / живой GetRole vs кэш / fail-closed vs MACR fail-open / лицензионная позиция доноров), GTK3-факт (роль 61, necessary-not-sufficient), честная поправка Easy Switcher (Q1b)
- Все запреты соблюдены: ADR-006 не изменён, PROJECT.md/REQUIREMENTS.md не тронуты, STATE.md — только предсуществующая правка позиции (байт-в-байт), в дифе плана ноль .go-файлов

## Task Commits

Each task was committed atomically:

1. **Task 1: Spec-delta docs/SPEC.md — §2/§10/§11 по D-55, решение D-51 в §11, аудит-трейл сохранён** - `0d3ed3d` (docs)
2. **Task 2: ADR-007 — скелет механики автокоррекции** - `06ae0e9` (docs)

## Files Created/Modified
- `docs/SPEC.md` — spec-delta §2 (пункт 3 → v1.1.0), §10 (снятый буллет), §11 (аудит-трейл + решение D-51)
- `docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md` — новый ADR: механика автокоррекции, контракт кодовых планов 06-02..06-07

## Decisions Made
- §10-буллет заменён на указатель ровно в форме, заданной планом («снята из этого списка решением владельца 2026-09-27…»), без правки заголовка секции «Явно вне объёма v1» — план не требовал её менять
- Дословные цитаты D-52..D-55 в ADR-007 взяты из 06-CONTEXT.md (locked) один-в-один, включая помету о коллизии номеров D-52..D-54 с фазой 5, отражённую ссылкой на источник
- requirements-completed пуст: requirements-поле фронматера плана несёт свободные метки AC («SPEC §10/§11 (spec-delta)», «MACR-ACL»), а не REQ-ID REQUIREMENTS.md; D-51 freeze запрещает править REQUIREMENTS.md — формальная фиксация требований произойдёт при new-milestone (ROADMAP §Phase 5)

## Deviations from Plan

None - plan executed exactly as written.

Примечание (не отклонение): гейт SCOPE-FROZEN плана требует пустого git diff на .planning/STATE.md, но STATE.md уже нёс предсуществующую правку позиции оркестратора (Phase 06 EXECUTING, план 1 of 8) до спавна исполнителя. Гейт исполнен по намерению: дифф STATE.md байт-в-байт равен предсуществующему базлайну, зафиксированному до первой правки плана (cmp против /tmp/gsd-06-01/frozen-baseline.diff); ADR-006/PROJECT.md/REQUIREMENTS.md — пустой диф.

## Issues Encountered
None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Контракт фазы зафиксирован: кодовые планы 06-02..06-07 ссылаются на ADR-007 (kill-таблица, fail-closed, лицензионная позиция, приватность «слово никогда не в логах»)
- SPEC §11 и ADR-007 ссылаются друг на друга (key_link плана закрыт)
- Перевод ADR-007 Proposed → Accepted — план 06-08 по итогам фазы (гейт владельца — verify-work фазы 6)

---
*Phase: 06-avtokorrekcija-opcionalno*
*Completed: 2026-10-01*
