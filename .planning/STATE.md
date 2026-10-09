---
gsd_state_version: "1.0"
milestone: v1.2
milestone_name: Терминалы и видимость приложений
current_phase: 09
current_phase_name: Спайк и дефект wezterm
status: executing
stopped_at: "09-02 blocked at live-session gate: 4th probe (13:21, bitwarden removed, idle desk) still fails; next rung is owner-side: bitwarden autostart off + relogin"
last_updated: "2026-10-09T09:30:00.000Z"
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
- [Phase 9 / 09-02] Live-session gate STILL BLOCKED after four probes: relogin 10:13 applied the formal fresh-session cure; probes 1 (10:39, age 25 min = within D-48 gate), 2 (10:48), 3 (11:01, idle desk, confound removed) and 4 (13:21, prime suspect REMOVED — owner closed all 7 bitwarden processes with explicit approval; bitwarden was autostarting headless `--ozone-platform=x11` every session, which explains why the relogin did not cure) all reproduce the identical ground truth: AT-SPI witness pinned at gnome-shell:WINDOW:chars=-1 through 30× grabFocus pokes, focused-inputs=(none), zenity dialog immune to ydotool Enter, stdout empty. Probe 4 microdelta: focus_out/focus_in pairs DID fire at zenity spawn (13:21:08, absent in probes 2–3) — removing bitwarden changed the map dynamics, but focus still does not stick (no stable zenity:TEXT frame, Enter undelivered). Focus machinery alive (owner typing through the daemon 10:14:20/10:33:02). NEXT RUNG IS OWNER-SIDE (executor forbidden): disable bitwarden autostart, relogin with it NOT running, then one witness probe. All four spike cases + TERM-04 regression still need working injection surfaces; running them now would write false verdicts into adr-008-spike-log.md and risk typing into owner windows (T-09-02-03). Committed RED (c21e043) + GREEN (bbd463d) verified (unit corpus, mise run ci green, keys verbatim vs ibuskeysyms.h, go.mod untouched); drafted fixture/case files stay uncommitted in-tree pending the live smoke run.
- [Phase 9 / 09-02] LIVE BISECT VERDICTS (orchestrator, 2026-10-09 ~20:00, all live/no reboot, state restored after each): (1) daemon out of input loop (goswitchd stop + `ibus engine xkb:us::eng`) → probe FAIL — daemon cleared; (2) ALL 6 GNOME extensions disabled → probe FAIL — extensions cleared; (3) client path cleared (GDK_BACKEND=x11 zenity via XWayland FAILS same as native Wayland); (4) /dev/input per-device capture blocked by uaccess for this shell — phantom-event capture not possible non-root. Probe harness rebuilt inline (zenity entry + Atspi StateType walk + ydotool key 28) — reproduces the defect (rc=124, AT-SPI witness pinned gnome-shell|FOCUSED). Conclusion: map-time focus-grant is yanked at compositor level; with userspace exonerated live, kernel 7.2.9-2 (T2, updated Oct 8 00:24) remains the only delta — fix requires reboot/kernel update (owner declines). ADAPTATION DECISION (owner-authorized 2026-10-09 «справиться без шаманства»): 09-02 proceeds in MANUAL-ACTIVATION mode — machine maps fixture windows, OWNER CLICKS the window (the production gesture anyway: users focus terminals by clicking), machine waits for a stable-focus marker (daemon focus_in / AT-SPI ACTIVE on the fixture), then injects/observes. Every verdict line in adr-008-spike-log.md gets `activation: manual-click` so the evidence stays honest; caps-table rows record per-terminal whether the yank affects them at all (all probes so far were zenity — terminals are genuinely untested). If even owner-clicked surfaces reject injected keys → deeper defect, STOP with evidence (that would threaten the phase, not just its convenience).

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

Last session: 2026-10-09T09:30:00.000Z
Stopped at: 09-02 blocked at live-session gate: 4th probe (13:21, bitwarden removed, idle desk) still fails; next rung is owner-side: bitwarden autostart off + relogin
Resume file: .planning/phases/09-spayk-i-defekt-wezterm/09-02-PLAN.md

## Operator Next Steps

- Одобрить роадмап (оркестратор презентует); затем /gsd:plan-phase 9
