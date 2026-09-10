---
gsd_state_version: "1.0"
current_phase: 1
current_phase_name: ADR-пакет и каркас IBus-движка
status: executing
stopped_at: Completed 01-01-PLAN.md
last_updated: "2026-09-10T18:59:44.790Z"
last_activity: 2026-09-10
last_activity_desc: Phase 1 execution started
state_head: e9dc626831cc4f438476e751b8970e54d20554f7
progress:
  total_phases: 4
  completed_phases: 0
  total_plans: 5
  completed_plans: 1
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-09-10)

**Core value:** По горячей клавише исправить текст, набранный не в той раскладке (EN↔RU), в любом поле ввода GNOME Wayland — через IBus engine, без root и без конфликтов с keyd/xremap.
**Current focus:** Phase 1 — ADR-пакет и каркас IBus-движка

## Current Position

Phase: 1 (ADR-пакет и каркас IBus-движка) — EXECUTING
Plan: 2 of 5
Status: Ready to execute
Last activity: 2026-09-10 — Phase 1 execution started

Progress: [░░░░░░░░░░] 0%

## Performance Metrics

**Velocity:**

- Total plans completed: 0
- Average duration: -
- Total execution time: 0 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1. ADR-пакет и каркас IBus-движка | 0/? | - | - |
| 2. Коррекция слова EN↔RU | 0/? | - | - |
| 3. Фразы, выделение, переключение и конфигурация | 0/? | - | - |
| 4. Поставка и приёмка | 0/? | - | - |

**Recent Trend:**

- Last 5 plans: -
- Trend: -

*Updated after each plan completion*
**Per-Plan Metrics:**

| Plan | Duration | Tasks | Files |
|------|----------|-------|-------|
| Phase 1 P01 | 35 min | 2 tasks | 19 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Roadmap: структура следует вехам SPEC.md §9 — M0+M1→Фаза 1, M2→Фаза 2, M3→Фаза 3, M4→Фаза 4; granularity coarse.
- Roadmap: recover-шим и цикл перерегистрации — day-one элементы Фазы 1 (падение движка = смерть ввода рабочего стола).
- Roadmap: TEST-02/03 (e2e-стенд) в Фазе 1 как скелет (ydotool → лог), матрица кейсов (TEST-04) — Фаза 2.
- Roadmap: INST-04 (логи/трассировка) в Фазе 1 — гейт M1 требует событий в логе.
- [Phase 1]: RequestName targets the component name org.freedesktop.IBus.goswitch — org.freedesktop.IBus is reserved by ibus-daemon (live-probed); single-instance guard preserved
- [Phase 1]: Live-gate proof uses ListActiveEngines (ibus list-engine reads only the XML registry) and layout-independent keycode assertions — keyvals follow the focused client's XKB group
- [Phase 1]: nonamedreturns disabled in strict lint config: the INTEG-05 recover shim requires named returns for zero-value panic replies

### Pending Todos

None yet.

### Blockers/Concerns

- Phase 1: ADR Decision #1 (two-engine vs внутренний флип) — конфликтующие данные по `gsettings set current` на GNOME 46; решается kill-criterion спайком в Фазе 1 до любого кода коррекции.
- Phase 1: MACR-01 — РЕШЕНО владельцем 2026-09-10 (D-12): в v1, внутри goswitch; keyd отвергнут. ADR-005 выбирает механизм. Открытых вопросов владельца нет.
- Go-работа всех фаз идёт по скиллу go-ultimate (project skill, .zcode/skills/) — конвенции и ревью-чеклист оттуда.

## Deferred Items

Items acknowledged and deferred at milestone close, most recent first:

| Category | Item | Status | Deferred At | Milestone |
|----------|------|--------|-------------|-----------|
| *(none)* | | | | |

## Session Continuity

Last session: 2026-09-10T18:59:44.763Z
Stopped at: Completed 01-01-PLAN.md
Resume file: None
