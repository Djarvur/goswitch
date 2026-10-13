---
gsd_state_version: "1.0"
milestone: v1.2
milestone_name: Терминалы и видимость приложений
current_phase: 10
current_phase_name: TERM-слово
status: executing
stopped_at: Completed 10-08-PLAN.md
last_updated: "2026-10-13T21:38:45.859Z"
last_activity: 2026-10-13
last_activity_desc: Phase 10 execution started
state_head: 4b65c1782b2cc9215bd5228ad96448f0ab3acd0d
progress:
  total_phases: 5
  completed_phases: 0
  total_plans: 13
  completed_plans: 9
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-10-08)

**Core value:** По горячей клавише исправить текст, набранный не в той раскладке (EN↔RU), в любом поле ввода GNOME Wayland — через IBus engine, без root и без конфликтов с keyd/xremap.
**Current focus:** Phase 10 — TERM-слово

## Current Position

Phase: 10 (TERM-слово) — EXECUTING
Plan: 9 of 9
Status: Ready to execute
Last activity: 2026-10-13 — Phase 10 execution started

Progress: [░░░░░░░░░░] 0%

## Performance Metrics

**Velocity (наследие v1.0.0):**

- Total plans completed: 61 (8 фаз вехи v1.0.0, 2026-09-10..2026-10-08)
- Веха v1.2: 0 plans

**Per-Plan Metrics:**

| Plan | Duration | Tasks | Files |
|------|----------|-------|-------|
| Phase 09 P01 | 10 min | 2 tasks | 2 files |
| Phase 10 P01 | 37min | 3 tasks | 1 files |
| Phase 10 P02 | 28min | 3 tasks | 13 files |
| Phase 10 P03 | 31min | 2 tasks | 2 files |
| Phase 10 P04 | 39min | 3 tasks | 5 files |
| Phase 10 P05 | 102min | 3 tasks | 12 files |
| Phase 10 P06 | 67min | 3 tasks | 8 files |
| Phase 10 P07 | 150min | 3 tasks | 13 files |
| Phase 10 P08 | 141min | 2 tasks | 13 files |

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
- [Phase 10]: Терминальный класс — один шов classifyTerminal на двух потребителей: вето D-55 в точке армирования (после blocklist, до detect.Check) и терминальный конвейер 10-04; роль 60 запрашивается в момент решения под autoRoleTimeout, verify-dead учится на истёкшем verify-бюджете при живом контексте — План предписывал вето ДО детектора независимо от наблюдаемости identity; confirm-стадия де-факто держала ролевое решение с эпохи ADR-007 — перенос в autoCorrectBoundary делает инвариант буквальным, а confirm остаётся вторым поясом
- [Phase 10]: Терминальная ветка записи: OQ 1 unified — весь терминальный класс (роль 60/паттерн/verify-dead) идёт в чанкированный level-2 (≤8, перечек ctxEpoch между чанками), pending-машина ADR-004 для терминалов не существует — Единый канал — буква TERM-02 и ADR-008 D3/D4; жертвуется работающий level-1 путь VTE: середина строки VTE защищается на уровне буфера, не поля; разъезд вето D-55 и конвейера записи невозможен (один шов classifyTerminal)
- [Phase 10]: Середина строки — OQ 2 contract (D7): буфер-доверие только для флоу «набор сразу в позицию курсора», «стрелки → набор» вне защиты честно; poison-флаг и primary-read НЕ ставились, Primary остаётся примитивом без потребителя; граница зафиксирована в docs/CONFIG.md, амендмент ADR-008 — 10-06 — Вердикты владельца 2026-10-12 из checkpoint'а применены дословно; контрактная граница — не настраиваемый режим: отдельного выключателя нет
- [Phase 10]: A2 ответен живым рядом 10-05: тесный бурст ForwardKeyEvent+CommitText переставляется в VTE (форварды стоят в GDK-очереди клиента, коммит синхронен; живой hex «ghbdtn»→«ghприв») — executeLevel2 пакирован (sleep 20ms перед каждой эмиссией), wire-порядок и чанк-геометрия не тронуты, после фикса ряды байт-точные
- [Phase 10]: Наблюдатель a11y демона регистрирует собственный интерес в at-spi реестре (RegisterEvent Object:state-changed) — мосты GTK эмитят фокус-события только при живой регистрации (0→153 StateChanged в живой пробе); raw bus match слеп на голом столе, отказ регистрации честно валит Start
- [Phase 10]: Plan 10-07: флап-окно стирания verifyDead — внутренняя константа 10 мс с журнальным происхождением (пары 1–3 мс, 16:20:05), не конфиг-ключ; Reset и синтетика флипа стирают немедленно, ctxEpoch/HardReset семантика 10-04 байт-неизменна
- [Phase 10]: Plan 10-07: дефолтный identity-якорь [gnome-terminal, wezterm] едет И нормализацией config.Load (nil→якорь, явный [] — opt-out), И в Defaults() (no-file демон классифицирует идентично); якорь доезжает существующим applySnapshot-фолдом, шов не правился
- [Phase 10]: Plan 10-07: role-диагностика громкая и липкая — счётчик ошибок эпизода в status (ac_role_errors=, обнуляется здоровым ответом :2741-семантикой), WARN с окном переармивания 30 мин (происхождение: эпизод 11:00:16 молчал весь день)
- [Phase 10]: Plan 10-07: встроенная форма terminal-silent-repin (term-negative-witnesses) считает correction-записи ОБЩЕГО журнала кейса — в них падают done+verify-after позитивного контроля (до-существующий артефакт формы, не регрессия); вердикт негатива несёт STANDALONE форма (ряд 19:07:31 зелёный при коде 10-07)

- [Phase 10]: Plan 10-08: машина отложенного флипа defer-until-confirmed — SetGlobalEngine только после подтверждения доставки (подтверждающий push / клавиша / flipConfirmGrace 300 мс / mismatch / тихий таймаут); release переиспользует flipTo целиком (D-36 не дублируется); ни один release-сайт не сидит на FocusOut
- [Phase 10]: Plan 10-08: owner-approved Rule 4 амендмент — nil-cache word-раунд с ответом-без-набранного-пробега = терминальный контекст (класс учится, раунд исполняет терминальную ветку); fail-closed для кэш-свидетельств сохранён; дрейф 0x9→0x29 документирован датированными рядами

### Pending Todos

None yet.

### Blockers/Concerns

- [Phase 13]: planning blocked on Telegram owner decision (лестница: snap → deb/flatpak с разовой проверкой a11y-reachability → демон-fallback)
- [Phase 12]: feasibility-gated — VTE kill-критерий может сузить строку до wezterm-only; вердикт о сужении за владельцем; фаза может быть снята без провала вехи
- [Phase 10 / 10-08] RESOLVED (owner approved semantic amendment 2026-10-13): контрольная точка Task 2 — дрейф подписи клиента wezterm-nightly (0x9→0x29 + Require-ответы протухшим кэшем) закрыт амендментом маршрутизации (pendingStaleCache: nil-cache раунд + ответ-без-набранного-пробега = терминальный контекст, класс учится, раунд исполняет терминальную ветку); fail-closed для кэш-свидетельств сохранён; все четыре живых кейса зелёные на одной свежей сессии
- [Phase 10]: пер-терминальная семантика клавиш (VTE пропускает Shift+Home в шелл, wezterm #4615 может терять Shift) — истина только в живой capability-таблице Фазы 9, не в веб-источниках
- [Phase 9 / 09-02] Live-session gate STILL BLOCKED after four probes: relogin 10:13 applied the formal fresh-session cure; probes 1 (10:39, age 25 min = within D-48 gate), 2 (10:48), 3 (11:01, idle desk, confound removed) and 4 (13:21, prime suspect REMOVED — owner closed all 7 bitwarden processes with explicit approval; bitwarden was autostarting headless `--ozone-platform=x11` every session, which explains why the relogin did not cure) all reproduce the identical ground truth: AT-SPI witness pinned at gnome-shell:WINDOW:chars=-1 through 30× grabFocus pokes, focused-inputs=(), zenity dialog immune to ydotool Enter, stdout empty. Probe 4 microdelta: focus_out/focus_in pairs DID fire at zenity spawn (13:21:08, absent in probes 2–3) — removing bitwarden changed the map dynamics, but focus still does not stick (no stable zenity:TEXT frame, Enter undelivered). Focus machinery alive (owner typing through the daemon 10:14:20/10:33:02). NEXT RUNG IS OWNER-SIDE (executor forbidden): disable bitwarden autostart, relogin with it NOT running, then one witness probe. All four spike cases + TERM-04 regression still need working injection surfaces; running them now would write false verdicts into adr-008-spike-log.md and risk typing into owner windows (T-09-02-03). Committed RED (c21e043) + GREEN (bbd463d) verified (unit corpus, mise run ci green, keys verbatim vs ibuskeysyms.h, go.mod untouched); drafted fixture/case files stay uncommitted in-tree pending the live smoke run.
- [Phase 9 / 09-02] LIVE BISECT VERDICTS (orchestrator, 2026-10-09 ~20:00, all live/no reboot, state restored after each): (1) daemon out of input loop (goswitchd stop + `ibus engine xkb:us::eng`) → probe FAIL — daemon cleared; (2) ALL 6 GNOME extensions disabled → probe FAIL — extensions cleared; (3) client path cleared (GDK_BACKEND=x11 zenity via XWayland FAILS same as native Wayland); (4) /dev/input per-device capture blocked by uaccess for this shell — phantom-event capture not possible non-root. Probe harness rebuilt inline (zenity entry + Atspi StateType walk + ydotool key 28) — reproduces the defect (rc=124, AT-SPI witness pinned gnome-shell|FOCUSED). Conclusion: map-time focus-grant is yanked at compositor level; with userspace exonerated live, kernel 7.2.9-2 (T2, updated Oct 8 00:24) remains the only delta — fix requires reboot/kernel update (owner declines). ADAPTATION DECISION (owner-authorized 2026-10-09 «справиться без шаманства»): 09-02 proceeds in MANUAL-ACTIVATION mode — machine maps fixture windows, OWNER CLICKS the window (the production gesture anyway: users focus terminals by clicking), machine waits for a stable-focus marker (daemon focus_in / AT-SPI ACTIVE on the fixture), then injects/observes. Every verdict line in adr-008-spike-log.md gets `activation: manual-click` so the evidence stays honest; caps-table rows record per-terminal whether the yank affects them at all (all probes so far were zenity — terminals are genuinely untested). If even owner-clicked surfaces reject injected keys → deeper defect, STOP with evidence (that would threaten the phase, not just its convenience).
- [Phase 9 / 09-02] Manual-activation harness is LIVE and honest (19:52-19:56 run): spike-smoke built its manual gates into the drafted cases (waitManualClick: AT-SPI witness OR fresh daemon focus_in pair above post-settle base; await logged every 15 s; activation: manual-click marker auto-prepended to every journal row; per-run map-time auto-grant observation row), fixture window mapped and the case waited the FULL 3-minute budget — NO CLICK ARRIVED (owner apparently away), so the case FAILED at the gate exactly as designed: no injection into the unfocused surface, journal abort row written, unit daemon restored to active. The harness itself needs no further work; live runs need the OWNER PRESENT and clicking each mapped window (zenity removed from the spike flow — the click replaces it).
- [Phase 9 / 09-02] MANUAL ACTIVATION EXHAUSTED (clean run 21:37–21:41, orchestrator hands-off per hygiene notice; owner present and confirmed «clicked every window, including the wezterm fixtures»): the fixture mapped and its pane program RAN (readiness marker — the A3 PROG fix: `--config-file` is a TOP-LEVEL flag, after `start` the CLI rejects it outright, pinned 21:22 after two aborts where NO window ever existed), yet ZERO IM contexts of ANY kind engaged the stand daemon during the full 3-minute click window (0 caps records — run 3, contaminated 21:26–21:35, had shown 4 foreign 0x29-class mints incl. engine/4 goswitch-ru; the clean run shows even re-mints require the owner clicking their OWN already-running apps). Sequence of gate strengthenings pinned two false-pass classes first (bare focus pairs 21:17; foreign context mints 21:22 — incl. one stray probe-forward 'x' likely landed in the owner's terminal, noted honestly): the final caps-0x9 client-fingerprint gate is sound. CONCLUSION: the compositor focus-yank defeats CLICK-granted focus for NEW surfaces — the defect is deeper than activation; no injection evidence is producible on this session, honestly (stop-condition a). The spike program (tasks 1–3 live parts) stays halted; REMAINING RUNG (owner-side): reboot into kernel 7.2.9-1 (boot-0 correlation; the Oct 8 7.2.9-2-t2 update is the only uncleared delta) and re-run the whole program on a working session — or abort/defer the phase. Harness commits pending the live green; RED (c21e043) + GREEN (bbd463d) stand.
- [Phase 9 / 09-02] BATCH2 HALTED BY OWNER (2026-10-11 ~19:10, писк/окна): R1 spike-smoke записал FAIL «arming gate: no key-trace record after 5 modifier taps» ПРИ подтверждённом клике (focus-пары демона 18:57:50) и видимой инъекции в окне — противоречие зафиксировано дословно в adr-008-spike-log.md, разбор за следующим заходом. R2–R6 не исполнены. Центральное свидетельство среды подтверждено валидацией: фоновые окна мертвы для клавиатуры (auto-grant нестабилен / AT-SPI ✗ / EWMH ✗ / клик ✗), окна владельца живы; AT-SPI-репортинг фокуса недостоверен глобально — свидетеля убрать из маркеров стенда. Runbook «владелец запускает → кликает → инъекция видна» доставлен; остался нерешённым арминг/трейс на стороне кейса. Все процессы пакета убиты, агент остановлен; живая часть 09-02 ждёт нового захода (сначала починить arming-gate, без звуковых циклов).
- [Phase 9 / 09-02] OWNER ENDED LIVE-TEST PARTICIPATION (2026-10-12 ~02:00): after a day of harness failures the owner refused the tester role outright. Live rounds R2–R6 remain unexecuted and stay blocked on the environment: background-spawned windows are keyboard-dead on kernel 7.2.9-2-t2 (attribution defect, proven), and machine focus-detection proved unfixable at acceptable cost (six harness patches, all superseded). WHAT STANDS COMMITTED: the phase's core scientific result — engine→stale-context desync (XIM) reproduced twice, dated verdict rows + preserved daemon journals (a7ac579, 4953d78, 5a5b9f3, evidence-20261011-*.log); harness rewritten headless-green through 47e0583 (one observation layer, silent, patient daemon lifecycle, pty-armed gates, replay-pinned classifiers). PATHS FORWARD (owner's call, urgent): t2-linux kernel fix lands → re-run the 5-case runbook as-is; or a CI-runner desktop with working focus → run headless-owned; or resume owner-run commands if ever desired. The spike log's desync rows are already usable by phases 10/12 as feasibility evidence per the plan's own semantics.
- Ячейка silence-matrix role-60 мигрирована на арминг-ожидание terminal-veto (D-55 переносит вердикт из confirm в арминг) — записана в WINDOWS.md как deviation, открытой до фазовой верификации
- [Phase 10 / 10-08] CHECKPOINT (blocking-human, 2026-10-13): Задача 1 (машина отложенного флипа defer-until-confirmed) закоммичена (5a47555 RED, bfdeaf1 GREEN, mise ci зелёный). Задача 2 (стенд: re-pin замороженных рядов + term-word-vte-flip + амендмент ADR-008) заблокирована дрейфом стенда: wezterm-nightly клиент на ВМ минтит caps 0x29 (surrounding-бит анонсируется) против запиненного 0x9 и отвечает на RequireSurroundingText протухшим кэшем; подпись стабильна через ibus-restart и reboot ВМ, пакеты старше зелёных рядов 10-06. Починка в демоне = смена семантик 10-04/ADR-004 (Rule 4). Ждёт вердикт владельца: (1) семантическая дельта плана, (2) ручная активация стенда кликами, (3) поиск источника дрейфа 0x9→0x29, (4) закрыть план с задокументированным окруженческим хвостом. Возобновление: SendMessage агенту agent_5a61967f-ffab-4383-afff-962bf2450d7b с выбранным вариантом. 10-09 (хост-ревалидация) стоит за 10-08.

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

Last session: 2026-10-13T21:38:45.810Z
Stopped at: Completed 10-08-PLAN.md
Resume file: None

## Operator Next Steps

- Одобрить роадмап (оркестратор презентует); затем /gsd:plan-phase 9
- [Phase 9 / 09-02] HALT (owner, 2026-10-11 вечер): attach-батч остановлен после R1 — клик подтверждён (focus_in демона, AT-SPI-свидетель исключён из гейта), инъекция ВИДИМА в окне владельцем, но key-trace демона пуст (armGate FAIL 5/5) — слой (3) клиент↔движок разорван; R2–R6 не исполнялись. Окруженческие находки стоят: лаунч-атрибуция (фон-спавн мёртв, owner-launch печатает), пути активации (grab/EWMH/auto-grant ✗). Возобновление: owner-runbook R2–R6 (готов, команды в журнале батча) после чина арм/трейс-гейтинга; всё записано в adr-008-spike-log.md (HALT-запись).
