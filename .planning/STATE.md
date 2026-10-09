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
- [Phase 9 / 09-02] ENV DIAGNOSIS LEAD (orchestrator, 2026-10-09 ~13:50): broken-boot correlation brackets ALL 12 failures. Boot -1 (Oct 1 10:59 → Oct 8 18:39:05) hosted every successful v1.0/v1.1 e2e run; boot 0 (since Oct 8 18:39:26; kernel 7.2.9-2-t2-noble + Mesa 25.2.8-...4, both upgraded Oct 8 00:24 while still on the old boot) hosts the 8 night spike attempts (23:32–23:53 Oct 8) AND probes 1–4. Owner relogins (probes 1–3) never left boot 0, so the kernel/Mesa delta was never tested. Ruled out this session: focus-mode (click), punto-switcher daemon (unit masked, inactive; autostart entry no-ops), bitwarden (fully removed 13:15 with owner approval — probe 4 microdelta shows it DID perturb focus events but is not the root cause), GNOME extensions (no installs/changes since Sep 30, predating the breakage; all were present and working on boot -1). NEXT RUNG (owner, ~3 min): reboot and pick "Ubuntu, with Linux 7.2.9-1-t2-noble" in GRUB → Advanced options (new Mesa userspace stays — this isolates the KERNEL), then run /gsd-execute-phase 9; its executor re-verifies the focus gate by design before any live run. If still broken on -1: live extension bisection, then Mesa downgrade, are the next brackets. If fixed: stay pinned on 7.2.9-1 for e2e work and report the regression to t2-linux.

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
