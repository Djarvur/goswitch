---
gsd_state_version: "1.0"
milestone: v1.2
milestone_name: Терминалы и видимость приложений
current_phase: 09
current_phase_name: Спайк и дефект wezterm
status: executing
stopped_at: "09-02 blocked at live-session gate: session focus drift; RED/GREEN committed, drafts in tree"
last_updated: "2026-10-09T06:41:57.511Z"
last_activity: 2026-10-08
last_activity_desc: Phase 09 execution started
state_head: bbd463d61c84aafa925f12dd2379c3723c374f69
progress:
  total_phases: 5
  completed_phases: 0
  total_plans: 4
  completed_plans: 1
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-10-08)

**Core value:** По горячей клавише исправить текст, набранный не в той раскладке (EN↔RU), в любом поле ввода GNOME Wayland — через IBus engine, без root и без конфликтов с keyd/xremap.
**Current focus:** Phase 09 — Спайк и дефект wezterm

## Current Position

Phase: 09 (Спайк и дефект wezterm) — EXECUTING
Plan: 2 of 4
Status: Ready to execute
Last activity: 2026-10-08 — Phase 09 execution started

Progress: [░░░░░░░░░░] 0%

## Performance Metrics

**Velocity (наследие v1.0.0):**

- Total plans completed: 61 (8 фаз вехи v1.0.0, 2026-09-10..2026-10-08)
- Веха v1.2: 0 plans

**Per-Plan Metrics:**

| Plan | Duration | Tasks | Files |
|------|----------|-------|-------|
| Phase 09 P01 | 10 min | 2 tasks | 2 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Roadmap v1.2 (spike-first, research SUMMARY): Фаза 9 = spec-delta D-55 + ADR-008 + живой capability-спайк (фикс TERM-01 включён, TERM-04-вето в SPEC до кода, A11Y-зонд рано/параллельно) → слово (10) → MENU (11, независима, полные in-repo прецеденты) → строка (12, feasibility-gated kill-критерием) → A11Y-рецепты (13, демон не меняется)
- Read channel per-terminal: VTE не имеет keyboard-selection (source-verified) — слово через собственный буфер нажатий демона; wezterm — CopyMode-рецепт + wl-paste --primary; запись ВСЕГДА N×BackSpace + commit_text (никогда commit-over-selection, никогда paste, CLIPBOARD байт-нетронут)
- Telegram snap структурно AppArmor-блокирован для входящих AT-SPI (живая проба 2026-10-08) — решение владельца (unconfined tarball/flatpak vs демон-fallback) гейтит планирование Фазы 13
- Конвенции всех фаз вехи: строгий TDD (red→green→refactor), зелёная итерация (build / test -race / golangci-lint), mise-задачи (не make), golangci-lint с первой фазы, spec-delta до кода (D-55), изменения через PR в защищённый main, go-ultimate skill
- [Phase 09]: Вето автокоррекции в терминалах живёт в точке армирования (роль AT-SPI 60, проверка ДО детектора), не в apps_blocklist; двухразрешительная модель и транспорт терминалов закреплены SPEC rev D-55 в §4.3/§4.4/§5/§7.2/§11 — Спека прежде кода (D-55); формулировка через blocklist противоречила бы D-53 «blocklist не может исключить невидимое»; TERM-04 инвариант дословно из 09-RESEARCH
- [Phase 09]: ADR-008 «Терминальный транспорт» создан Proposed: D1–D7 с пустыми Evidence-ячейками-плейсхолдерами журнала adr-008-spike-log.md; заполняет только живой спайк (09-02), Accepted ставит 09-03 — Anti-Pattern «фикс из issue-tracker'а»: апстрим-факты (20240203 последний тег, ловушки #5125/#1772/#4615) упомянуты без веб-ссылок; grep-гейт отсутствия ссылок зелёный

### Pending Todos

None yet.

### Blockers/Concerns

- [Phase 13]: planning blocked on Telegram owner decision (лестница: snap → deb/flatpak с разовой проверкой a11y-reachability → демон-fallback)
- [Phase 12]: feasibility-gated — VTE kill-критерий может сузить строку до wezterm-only; вердикт о сужении за владельцем; фаза может быть снята без провала вехи
- [Phase 10]: пер-терминальная семантика клавиш (VTE пропускает Shift+Home в шелл, wezterm #4615 может терять Shift) — истина только в живой capability-таблице Фазы 9, не в веб-источниках
- [Phase 9 / 09-02] Live-session gate BLOCKED: the GNOME session's keyboard-focus machinery is drifted — no window can take focus (map-activation denied; AT-SPI FOCUSED dead for GTK3+GTK4; zenity Enter-test keystrokes go nowhere; witness stuck at gnome-shell:WINDOW:chars=-1). Precedented WINDOWS class (rows 7/10; fresh-session/login-state family; survives ibus restart). Formal cure: owner relogin (fresh-session gate, docs/ci-runner.md). All four spike cases + TERM-04 regression need working injection surfaces; running them now would write false verdicts into adr-008-spike-log.md. Committed RED (c21e043) + GREEN (bbd463d) verified (unit corpus, mise run ci green, keys verbatim vs ibuskeysyms.h, go.mod untouched); drafted fixture/case files stay uncommitted in-tree pending the live smoke run.

## Deferred Items

Items acknowledged and deferred at milestone close, most recent first:

| Category | Item | Status | Deferred At | Milestone |
|----------|------|--------|-------------|-----------|
| deferred_items | 02: персистентный AT-SPI wedge после серий SIGKILL (bus-restart remedy; кандидат CI-runner book) | acknowledged | 2026-10-08 | v1.0.0 |
| deferred_items | 02: verify-after на stale первом пуше (grace-window кандидат) | acknowledged | 2026-10-08 | v1.0.0 |
| deferred_items | 02: ydotool "Escape" резолвится в физическую E (однословный фикс "esc") | acknowledged | 2026-10-08 | v1.0.0 |
| deferred_items | 04: combo-word-layout флак после reload при необычном фокусе (surface-hold кандидат) | acknowledged | 2026-10-08 | v1.0.0 |
| deferred_items | 05: префлайт injection-selftest тап = реальный флип на установленном столе | acknowledged | 2026-10-08 | v1.0.0 |
| deferred_items | 06: session drift первых клавиш после флипа (WINDOWS #13, средовой; ночной D-48 — машинный гейт) | acknowledged | 2026-10-08 | v1.0.0 |
| deferred_items | 07/08: флаки ctlsvc/config/appid под -race нагрузкой (bus-isolation hardening кандидат) | acknowledged | 2026-10-08 | v1.0.0 |

## Session Continuity

Last session: 2026-10-09T06:41:45.023Z
Stopped at: 09-02 blocked at live-session gate: session focus drift; RED/GREEN committed, drafts in tree
Resume file: .planning/phases/09-spayk-i-defekt-wezterm/09-02-PLAN.md

## Operator Next Steps

- Одобрить роадмап (оркестратор презентует); затем /gsd:plan-phase 9
