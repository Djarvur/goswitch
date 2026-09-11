---
gsd_state_version: "1.0"
current_phase: 02
current_phase_name: Коррекция слова EN↔RU
status: executing
stopped_at: Completed 02-01-PLAN.md
last_updated: "2026-09-11T22:35:19.951Z"
last_activity: 2026-09-12
last_activity_desc: Phase 02 execution started
state_head: 888effd5bf8d6caf29fb1bf11e16f4181def4b95
progress:
  total_phases: 4
  completed_phases: 1
  total_plans: 12
  completed_plans: 6
  percent: 25
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-09-11)

**Core value:** По горячей клавише исправить текст, набранный не в той раскладке (EN↔RU), в любом поле ввода GNOME Wayland — через IBus engine, без root и без конфликтов с keyd/xremap.
**Current focus:** Phase 02 — Коррекция слова EN↔RU

## Current Position

Phase: 02 (Коррекция слова EN↔RU) — EXECUTING
Plan: 2 of 7
Status: Ready to execute
Last activity: 2026-09-12 — Phase 02 execution started

Progress: [███░░░░░░░] 25%

## Performance Metrics

**Velocity:**

- Total plans completed: 5
- Average duration: -
- Total execution time: 0 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1. ADR-пакет и каркас IBus-движка | 0/? | - | - |
| 2. Коррекция слова EN↔RU | 0/? | - | - |
| 3. Фразы, выделение, переключение и конфигурация | 0/? | - | - |
| 4. Поставка и приёмка | 0/? | - | - |
| 01 | 5 | - | - |

**Recent Trend:**

- Last 5 plans: -
- Trend: -

*Updated after each plan completion*
**Per-Plan Metrics:**

| Plan | Duration | Tasks | Files |
|------|----------|-------|-------|
| Phase 1 P01 | 35 min | 2 tasks | 19 files |
| Phase 01 P02 | 20 min | 2 tasks | 8 files |
| Phase 01 P03 | 80 min | 3 tasks | 13 files |
| Phase 01 P04 | 40 min | 3 tasks | 14 files |
| Phase 01 P05 | 8 min | 2 tasks | 7 files |
| Phase 02 P01 | 10 min | 2 tasks | 12 files |

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
- [Phase 1]: 01-02: layout tables join by xkb KEY POSITION (not character) — asymmetric pairs (&↔?, @↔", #↔№, $↔;, ^↔:, |↔/) fall out of the join; plan bullet's wrong per-char pairs (h→и, b→б, d→д) corrected to the string-level SPEC truth (h→р, b→и, d→в)
- [Phase 1]: 01-02: FSM TimerExpired honors the last tap's deadline — stale AfterFunc timers are no-ops; required by the intervening-key-cancel pin and race-safe for the adapter's re-arm
- [Phase 1]: 01-02: generator's Cyrillic table keyed by keysym NAME with values resolved from ibuskeysyms.h (Latin-1 direct 0x20–0xff); unknown names fail generation loudly — golden corpus pins at generation time, never runtime
- [Phase 1]: 01-02: KeyvalShiftR (0xffe2) lives in internal/hotkey, not imported from engine — pure package stays D-Bus-free; dependency direction will be engine → hotkey
- [Phase 01]: [Phase 1] 01-03: engine activation is IBus SetGlobalEngine — GNOME 46 shell runtime-ignores gsettings sources/current writes (live-verified); both keys stay snapshotted, restored only-if-changed
- [Phase 01]: [Phase 1] 01-03: e2e surface = self-focused zenity entry (primary) / shell PASSWORD_TEXT entry (locked fallback) — AT-SPI grabFocus refused under Wayland; injection gated on the AT-SPI witness
- [Phase 01]: [Phase 1] 01-03: teardown re-asserts the global engine AFTER daemon death — ibus-daemon unsets it on engine-component disconnect; never restores to a goswitch engine (xkb fallback)
- [Phase 01]: [Phase 1] 01-03: readback oracles are length-based — the desktop's active XKB group maps ghbdtn to привет, content comparison would be layout-dependent
- [Phase 01]: [Phase 01] 01-04 M0 passed: owner approved the ADR pack («утверждено», 2026-09-10) — Option B internal flip (D-01 experiment, zero switched probes), ADR-001..005 Accepted; STATE blocker «Decision #1» closed by kill-criterion spike before any correction code
- [Phase 01]: [Phase 01] 01-04: spec-deltas applied to SPEC on top of owner edits — §5 (waitless actions < 50 ms; Right Shift actions bounded by the 300 ms discrimination window) and §4.3 (mouse click out of reset triggers; mandatory surrounding-text check + best-effort cursor-jump reset); owner macros section renumbered 4.4→4.5 to fix duplicate numbering
- [Phase 01]: [Phase 01] 01-05: CI installs tools via mise install from mise.toml (not setup-go/golangci-lint-action — the action's version input would duplicate the mise pin and drift); CI and local mise run ci are the same tasks (D-09/D-11a)
- [Phase 01]: [Phase 01] 01-05: CONVENTIONS.md directives are single-line bullets — generate-claude-md's summarizer drops numbered lists/indented sub-bullets; directives must stay transport-safe for AGENTS.md regeneration
- [Phase 01]: verify-work 2026-09-11: UAT 4/4 — keyd-сосуществование подтверждено владельцем; первый pr-sanity зелёный на реальном раннере (PR #1, run 34596048376: build/vet/lint/test 39с, govulncheck 29с); dependabot-PR и cron-запуск приняты владельцем по эквивалентным доказательствам (активируются после мержа PR #1)
- [Phase 01]: SECURITY.md создан при verify-work (post-hook): 21 угроза из threat-моделей 5 планов, 21 закрыто / 0 открыто (ASVS L1), 3 принятых риска задокументированы
- [Phase 02]: [02-01] Both ladder levels replace token+tail and recommit converted+tail — deleting exactly the token at a non-empty tail strands the cursor after the tail (D-13 pin geometry, Pitfall 1)
- [Phase 02]: [02-01] CapSurroundingText mirrored as 1<<5 in internal/correct instead of importing engine — pure packages never import the D-Bus-bound engine (Phase 1 dependency direction precedent)
- [Phase 02]: [02-01] Backspace keeps the separator in Tail(): buffer mirrors the field («ghbdtn » after the pop) — plan's Tail=="" expectation was internally inconsistent with its own D-13 pin (Pitfall 1)
- [Phase 02]: [02-01] Corpus literals named as constants (wordEN/wordRU) per goconst of the strict lint; targeted #nosec G115 with justification on the plan.go int conversions

### Pending Todos

None yet.

### Blockers/Concerns

- Phase 1: ADR Decision #1 — ЗАКРЫТО планом 01-04 (2026-09-11): kill-criterion спайк D-01 прогнан до кода коррекции, ноль switched-проб, побеждает Option B (внутренний флип); ADR-001 Accepted на гейте M0.
- Phase 1: MACR-01 — РЕШЕНО владельцем 2026-09-10 (D-12): в v1, внутри goswitch; keyd отвергнут. Механизм утверждён на M0 (ADR-005 Accepted): глобальные правила + per-app YAML списки, идентичность — AT-SPI; код в Фазе 3.
- Go-работа всех фаз идёт по скиллу go-ultimate (project skill, .zcode/skills/) — конвенции и ревью-чеклист оттуда.

## Deferred Items

Items acknowledged and deferred at milestone close, most recent first:

| Category | Item | Status | Deferred At | Milestone |
|----------|------|--------|-------------|-----------|
| *(none)* | | | | |

## Session Continuity

Last session: 2026-09-11T22:35:19.905Z
Stopped at: Completed 02-01-PLAN.md
Resume file: None
