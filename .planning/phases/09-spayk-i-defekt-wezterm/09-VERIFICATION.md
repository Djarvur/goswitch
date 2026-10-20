---
phase: 09-spayk-i-defekt-wezterm
verified: 2026-10-20T17:30:00Z
status: passed
score: 5/5 must-haves verified
covered_files:
  - .planning/REQUIREMENTS.md
  - .planning/ROADMAP.md
  - .planning/STATE.md
  - .planning/WINDOWS.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-01-PLAN.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-01-SUMMARY.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-02-PLAN.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-02-SUMMARY.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-03-PLAN.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-04-PLAN.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-RESEARCH.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-VALIDATION.md
  - .planning/phases/09-spayk-i-defekt-wezterm/evidence-20261011-flip2-stuck.log
  - .planning/phases/09-spayk-i-defekt-wezterm/evidence-20261011-validate-daemon.log
  - .planning/phases/09-spayk-i-defekt-wezterm/evidence-20261012-t04-stuck.log
  - cmd/goswitchctl/main.go
  - cmd/goswitchctl/main_test.go
  - cmd/goswitchd/main.go
  - docs/SPEC.md
  - docs/adr/ADR-008-terminal-transport.md
  - docs/adr/adr-008-spike-log.md
  - internal/ctlsvc/ctlsvc.go
  - internal/ctlsvc/ctlsvc_test.go
  - internal/engine/keys.go
  - internal/session/probe.go
  - internal/session/probe_test.go
  - mise.toml
  - test/e2e/case_spike_caps.go
  - test/e2e/case_spike_wezterm.go
  - test/e2e/fixtures/gs-e2e-wezterm-stand.lua
  - test/e2e/fixtures/gs-e2e-wezterm-wayland.lua
  - test/e2e/fixtures/gs-e2e-wezterm-xim.lua
  - test/e2e/main.go
  - test/e2e/wezterm.go
  - test/e2e/wezterm_test.go
covered_digest: "v1:sha256:a8f1549bcf92aea63542e1fc34f69131e1f7115a136cbc95950d4aef5b8332ef"
behavior_unverified: 0
overrides_applied: 0
deferred: # Items addressed in later phases — not actionable gaps
  - truth: "Flatpak-ряд достижимости Telegram не исполнен живой пробой (09-04 retired)"
    addressed_in: "Phase 13"
    evidence: "Goal Фазы 13: «Telegram: решение владельца по лестнице (unconfined tarball/deb/flatpak vs демон-fallback) принято и исполнено» — SC2"
  - truth: "Проба ZCode .desktop-override с --force-renderer-accessibility не исполнялась (09-04 retired)"
    addressed_in: "Phase 13"
    evidence: "SC1 Фазы 13: «ZCode после user-local .desktop override с --force-renderer-accessibility и рестарта присутствует в живом AT-SPI-дереве»"
---

# Phase 9: Спайк и дефект wezterm — Verification Report

**Phase Goal:** Переключение раскладки работает в wezterm (ежедневный терминал владельца), а терминальный транспорт закреплён в SPEC и ADR до всякого кода: живой спайк производит verdict-таблицу возможностей, на которой строятся Фазы 10 и 12
**Verified:** 2026-10-20T17:30:00Z
**Status:** passed
**Re-verification:** No — первичная формальная верификация фазы (закрытие 5e957d0..b29b7ff после завершения Фазы 10)

## Goal Achievement

Контракт фазы — 5 Success Criteria из ROADMAP.md. Дискпозиции планов: 09-01 complete (commits ac676c4, 2a6e67a); 09-02 HALTED на запроектированной владельцем остановке (2026-10-12, ядро-результаты стоят); 09-03/09-04 superseded (доставлено Фазой 10 / рассеяно; retire_reason в frontmatter обоих планов, #2349-форма).

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC1: Переключение раскладки работает в wezterm на живом столе (хоткей → последующий ввод даёт символы новой раскладки) | ✓ VERIFIED | Прямые живые доказательства на хосте, прочитанные этим верификатором (read-only): `wezterm --version` = **20261010-083911-e2a98e1b** (ночная сборка из рецепта 7365a93), `apt-mark showhold` = wezterm-nightly, `~/.wezterm.lua` = `enable_wayland = true`; журнал демона `journalctl --user -u goswitchd` за 2026-10-19/20 несёт **16 живых рядов** mode-switch/flip (в т.ч. `combo kind mode-switch-chord` 2026-10-20 12:09:53 и correction delivery flip-armed/flip-released 15:06:09 с latency_ms 5–41). Человеческий оракул закрыт и записан: G-10-1, UAT Test 6 verdict **approved «осталось»** 2026-10-20 (10-VERIFICATION.md: passed 51/51, selftype-ряд 11:44:44 сверен тем верификатором). Десинк engine→stale-context (XIM) воспроизведён дважды датированными строками журнала спайка (21:34:26, 21:51:11 + машинный вердикт 22:28:44/54, `verdict: defect-reproduced`) — первопричина источник-диагностирована, путь нативного Wayland её обходит. **Хвост (записан, не блокирует):** буквальная форма «два подряд зелёных e2e-оракула в одном запуске» не исполнялась — остановка владельца 2026-10-12 (WINDOWS.md #21, status open); подмена оракула (ежедневный флоу владельца + стенд-ряды Фазы 10, байт-точные `d0bf…d182 0a` = «привет\n») зафиксирована в ROADMAP-аннотациях и 09-02-SUMMARY |
| 2 | SC2: SPEC несёт датированную spec-delta ревизию D-55 (до кода): слово обязательна, строка feasibility-gated, автокоррекция в терминалах НИКОГДА; двухразрешительная модель; вето в точке армирования, не в apps_blocklist | ✓ VERIFIED | `grep -c 'rev D-55' docs/SPEC.md` = **10** (≥5) по всем пяти секциям: §4.3 (:146 — терминальное разрешение + :155 feasibility-gated), §4.4 (:177 — транспорт ForwardKeyEvent+commit_text, wezterm нативный Wayland с фолбэком), §5 (:363 — «первое действие на проводе», best-effort, дословный аудит-трейл прежнего пункта), §7.2 (:413 — «два подряд… одиночный зелёный ничтожен», замороженные негативы «без ослабления»), §11 (:586 — «вето не живёт в blocklist… единственный дом вето — точка армирования… ДО всякого запуска детектора», коллизия с D-53 названа, роль AT-SPI 60 процитирована). Аудит-трейл-комментарии (HTML) на месте во всех секциях. Хронология «до кода»: D-55 закоммичен 2026-10-08 19:50 (ac676c4, 2a6e67a), зондовый код начинается позже (c21e043) |
| 3 | SC3: ADR-008 доведён до Accepted на свидетельствах живого спайка, не на веб-источниках | ✓ VERIFIED | docs/adr/ADR-008-terminal-transport.md: `**Accepted** (2026-10-12, план 10-06)`; все 7 строк D1–D7 имеют заполненные Evidence-ячейки со ссылками на именованные датированные ряды журнала (venue-probe r1/r2 14:12/14:16, word-bash/zsh 19:17:39/41, term-word-wezterm 20:34:29/31, clip-frozen, witness-vim/htop, leg-probe); **0 веб-ссылок** (`grep -c 'https\?://'` = 0); 12 ссылок на adr-008-spike-log.md; **7 амендментов** живыми вердиктами (2026-10-12 10-06, 10-13 10-08, 10-14 10-10, 10-14 10-11, 10-16 10-16, 10-16 10-17, 10-19 round-6). Перевод в Accepted исполнен Фазой 10 (план 10-06) на рядах, порождённых спайком Фазы 9 — требование «закреплён до кода, Accepted на живых свидетельствах» выполнено (Proposed 2026-10-08 до кода; Accepted 2026-10-12 на живых строках) |
| 4 | SC4: Вердикт-таблица возможностей терминалов существует — фиксирует kill-критерий строковой растяжки и чорды слова для Фаз 10/12 | ✓ VERIFIED | docs/adr/adr-008-spike-log.md (675 строк): секция «Capability-таблица терминалов» с VTE-рядами (a11y 18:19:12, доставка 18:19:15, select-chord **kill-критерий** 18:19:27 — «строка сужается до wezterm, вердикт о сужении — владелец (чекпойнт 09-03)», посадка+шорткат, caps 0x29), строка десинка spike-caps 22:30:05; классы слова/клип-политики/негативов заполнены стенд-рядами Фазы 10 («Фаза 10 живые ряды VTE» :265 — word-bash/zsh/mid-line/clip-frozen байт-точные; «Фаза 10 живые ряды wezterm» :335; witness-vim/htop :19:29 fired=0 veto Δ=1; leg-probe). Потребитель доказан делом: Фаза 10 построена на этой таблице и закрыта владельцем (G-10-1 «осталось» 2026-10-20); SC2 Фазы 12 в ROADMAP ссылается ровно на этот kill-критерий. **Отклонение формы (записано честно):** единая таблица «5×2» не собиралась — вердикты живут датированными рядами секций журнала (09-02-SUMMARY coverage D4) |
| 5 | SC5: A11Y-вердикты записаны для деградации MENU-идентичности (Фаза 11) и планирования Фазы 13 | ✓ VERIFIED (рассеянная форма) | Стена Telegram snap AppArmor — живая проба 2026-10-08, STRUCTURED-вердикт в .planning/STATE.md:79 («структурно AppArmor-блокирован для входящих AT-SPI»); гейт планирования Фазы 13 — решение владельца (STATE.md:113). ZCode/Electron-идентичность — фаза-7 UAT-вердикты, процитированы в Depends-on Фазы 11 ROADMAP («Electron-идентичность существует только после включения a11y»). a11y-невидимость wezterm — находка ac_role_errors Фазы 10. Отдельный журнал d09-…-log.md и кейс case_spike_a11y.go **намеренно не созданы** (09-04 superseded, retire_reason; отсутствие подтверждено на диске). Остатки (flatpak-ряд, ZCode-override проба) принадлежат исполнению Фазы 13 — см. deferred |

**Score:** 5/5 истин verified (0 present-behavior-unverified; SC1 имеет прямое поведенческое доказательство — живые журнальные ряды прочитаны этим верификатором, человеческий вердикт владельца записан и закрыт в 10-UAT.md)

### Deferred Items

| # | Item | Addressed In | Evidence |
|---|------|-------------|----------|
| 1 | Flatpak-ряд достижимости Telegram не исполнен живой пробой (09-04 retired) | Phase 13 | Goal/SC2 Фазы 13: решение владельца по лестнице unconfined tarball/deb/flatpak vs демон-fallback |
| 2 | Проба ZCode .desktop-override с --force-renderer-accessibility не исполнялась (09-04 retired) | Phase 13 | SC1 Фазы 13: user-local .desktop override + приёмка по живому AT-SPI-дереву |

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| docs/SPEC.md | rev D-55 в §4.3/§4.4/§5/§7.2/§11 с аудит-трейлами | ✓ VERIFIED | 10 маркеров rev D-55; HTML-комментарии аудит-трейла; коммит ac676c4 |
| docs/adr/ADR-008-terminal-transport.md | Proposed → Accepted, D1–D7, Evidence из живого журнала | ✓ VERIFIED | Accepted 2026-10-12 (план 10-06); 0 веб-ссылок; 7 амендментов |
| docs/adr/adr-008-spike-log.md | Журнал спайка: датированные вердикт-строки | ✓ VERIFIED | 675 строк: HALT-записи 2026-10-11/12, лестница P0 (XIM-гипотеза опровергнута), capability-таблица, десинк-строки, nightly-вердикты, секции Фазы 10 |
| internal/session/probe.go + probe_test.go | ParseProbeKey/ProbeForward, строгая грамматика, fail-closed | ✓ VERIFIED | 192 строки + 219 строк тестов; `go test ./internal/session/ -race -run "TestParseProbeKey\|TestProbeForward"` — ok |
| internal/engine/keys.go | KeyHome/KeyEnd/KeyLeft/KeyRight вербатим из ibuskeysyms.h | ✓ VERIFIED | :59–62, цитата заголовка в комментарии :55 |
| internal/ctlsvc/ctlsvc.go | ProbeForward с recoverMethod, nil-dep fail-closed | ✓ VERIFIED | :156–170, dep-поле :90, проводка только под -debug |
| cmd/goswitchctl/main.go | подкоманда probe-forward | ✓ VERIFIED | :107/:112/:330, errUsage :53 |
| test/e2e/wezterm.go + fixtures xim/wayland/stand | фикстура wezterm+pty, сентинелы PRIMARY, арминг | ✓ VERIFIED | 661 строка; wayland.lua: enable_wayland=true + CopyMode Shift+Home Multiple; xim.lua: enable_wayland=false |
| test/e2e/case_spike_wezterm.go, case_spike_caps.go | кейсы spike-smoke/spike-wezterm/spike-caps/wezterm-flip | ✓ VERIFIED | 879 + 448 строк; зарегистрированы в test/e2e/main.go (:401–402); mise-задачи e2e-spike-* (mise.toml :155–171) |
| evidence-журналы | журнал демона десинка в git; живой ввод НЕ в git | ✓ VERIFIED | evidence-20261011-validate-daemon.log tracked (только hex-синтетика 0x31–0x34, живого ввода нет); flip2/t04-журналы на диске, сознательно untracked (гигиена Pitfall 6) |
| docs/adr/d09-a11y-probe-log.md, test/e2e/case_spike_a11y.go | — (09-04 retired) | n/a (намеренно отсутствуют) | Рассеянные вердикты покрывают потребителей (Фазы 11/13); retire_reason в 09-04-PLAN.md |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| cmd/goswitchctl/main.go | internal/session/probe.go | ProbeForward (CLI → ctlsvc → session → engine) | ✓ WIRED | grep: case "probe-forward" :107 → runProbeForward :330 → ctlsvc.ProbeForward :163 → session.ProbeForward :171 → engine.Emitter |
| docs/adr/ADR-008-terminal-transport.md | docs/adr/adr-008-spike-log.md | Evidence-ячейки D1–D7 → именованные датированные ряды | ✓ WIRED | 12 ссылок; каждая ячейка D1–D7 цитирует конкретные ряды (даты/время) |
| docs/SPEC.md §11 | internal/appid/appid.go | роль AT-SPI 60 (RoleTerminal) | ✓ WIRED | SPEC :580 цитирует «роль 60 — RoleTerminal»; appid.go:83 `RoleTerminal uint32 = 60`; вето потреблено в internal/session/actor.go:3401 |
| test/e2e/case_spike_*.go | docs/adr/adr-008-spike-log.md | кейсы дописывают датированные вердикт-строки | ✓ WIRED | строки с activation-метками (manual-click/focus-detected) и Verdict-полями на месте |
| gs-e2e-wezterm-wayland.lua (рецепт) | ~/.wezterm.lua владельца | enable_wayland=true | ✓ WIRED (доставлено) | Фикстура — тот же рецепт; конфиг владельца применён (7365a93), откат задокументирован в журнале; проверено чтением ~/.wezterm.lua |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| adr-008-spike-log.md | датированные вердикт-строки | живые пробы кейсов + журнал демона | Да (activation-метки, pty_hex, hex-свидетельства) | ✓ FLOWING |
| ADR-008 Evidence D1–D7 | ссылки на ряды | строки журнала спайка | Да (именованные ряды существуют и датированы) | ✓ FLOWING |
| VERIFICATION-свидетельства SC1 | journalctl --user -u goswitchd | живой демон на хосте владельца | Да (16 рядов 2026-10-19/20 прочитаны верификатором) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Зондовый юнит-корпус зелёный под -race | `go test ./internal/session/ -race -count=1 -run "TestParseProbeKey\|TestProbeForward"` | ok, 1.094s | ✓ PASS |
| Ночная сборка wezterm поставлена (рецепт 7365a93) | `wezterm --version` | 20261010-083911-e2a98e1b | ✓ PASS |
| Рецепт применён в конфиге владельца + hold | `sed -n '74,82p' ~/.wezterm.lua; apt-mark showhold` | enable_wayland = true; wezterm-nightly | ✓ PASS |
| Живой флип в wezterm (цель фазы) | `journalctl --user -u goswitchd --since 2026-10-19 \| grep mode-switch\|flip` | 16 рядов, вкл. mode-switch-chord 2026-10-20 12:09:53, latency 5–41 мс | ✓ PASS |
| SPEC несёт D-55 во всех 5 секциях | `grep -c 'rev D-55' docs/SPEC.md` | 10 (≥5) | ✓ PASS |
| ADR-008 без веб-источников, со ссылками на журнал | `grep -c 'https\?://' ADR-008; grep -c 'adr-008-spike-log' ADR-008` | 0; 12 | ✓ PASS |
| Desync-вердикты в журнале | grep 'reproduced: engine→stale-context' spike-log | строки 21:34:26, 21:51:11, 22:28:44/54 + итог defect-reproduced | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
|-------|---------|--------|--------|
| mise e2e-spike-smoke / e2e-spike-wezterm / e2e-spike-caps / e2e-wezterm-flip | — | — | ? SKIPPED (live-GNOME) |

Пробы фазы — живые кейсы настольного GNOME (mise.toml: «live GNOME session; NOT in ci»); запуск означал бы спавн окон и инъекции на столе владельца, живое тестирование прекращено его решением 2026-10-12. Вердикт-строки этих проб существуют от живых сессий (журнал спайка, датированы) — верифицированы чтением, не перезапуском.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| TERM-01 | 09-01, 09-02 (09-03 superseded → Фаза 10) | Переключение раскладки работает в wezterm | ✓ SATISFIED | Рецепт (nightly + enable_wayland=true) поставлен планом 09-02 (7365a93); приёмка — живой ежедневный флоу владельца + вердикт G-10-1 «осталось» 2026-10-20 (Фаза 10, UAT Test 6 approved); ADR-008 D2 Evidence — дифференциальный вердикт сборки |
| TERM-04 | 09-01 (негатив-свидетели → Фаза 10 10-03/10-05) | Автокоррекция в терминалах никогда не срабатывает; вето spec-delta, не apps_blocklist | ✓ SATISFIED | SPEC §11 rev D-55 (вето в точке армирования, роль AT-SPI 60, blocklist отвергнут явно + коллизия D-53); код: internal/session/actor.go:3401 (role == acRoleTerminal) на internal/appid RoleTerminal; замороженные негативы: SPEC §7.2 «без ослабления» + witness-vim/witness-htop (fired=0, veto Δ=1, spike-log) + терминальный veto-счётчик хоста (10-UAT: terminal-veto ×11) |

Оба требования доставлены (TERM-01 — фикс 09-02 + зачёт Фазой 10; TERM-04 — спека 09-01 + вето и свидетели Фазы 10 — требование, доставленное поздней фазой, засчитывается вехе с указанием места посадки). **Книжkeeping-заметка оркестратору:** чекбоксы TERM-01/TERM-04 в .planning/REQUIREMENTS.md всё ещё `[ ]` Pending — 09-01/09-02-SUMMARY сознательно отложили отметку на фазовую верификацию (shared-ID gate); данный отчёт — вход для отметки.

Orphaned requirements: нет — таблица трассируемости назначает фазе 9 только TERM-01/TERM-04; A11Y-02(spike)-часть ретированного 09-04 остаётся за Фазой 13 (гейт — решение владельца), её спайковый вклад (стена AppArmor) записан в STATE.md.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | TBD/FIXME/XXX/TODO/placeholder: не найдено ни в одном файле фазы | — | — |

Prohibitions-проверки (все закрыты детерминированными доказательствами):
- «Вето НЕ через apps_blocklist» — VERIFIED (SPEC §11 отвергает blocklist явно);
- «ADR-008 без веб-ссылок» — VERIFIED (0 ссылок);
- «Ревизии не стирают прежний текст» — VERIFIED (аудит-трейл-комментарии во всех 5 секциях; §5 дословно);
- «go.mod/go.sum не меняются; clipboard/indicator/config не тронуты» — VERIFIED (`git log c21e043^..7365a93 -- go.mod go.sum` пуст; diff — 28 файлов, пакетов нет в списке);
- «Зондовый код не вырастает в продуктовый конвейер» — VERIFIED по объёму диффа (только probe.go/keys.go/фикстуры/кейсы/харнесс; степпер-машина пришла Фазой 10 в её собственном скоупе);
- «pty-захваты и живой ввод не коммитятся» — VERIFIED (трекнутый журнал содержит только hex-синтетику; flip2/t04 с живым вводом — untracked на диске).

### Decision Coverage

SKIPPED — в каталоге фазы нет *-CONTEXT.md (гейт чисто пропускается по правилу).

### Human Verification Required

Не требуется. Человеческие оракулы фазы уже исполнены и записаны: владелец лично вёл живые пробы (решения о ветке рецепта 2026-10-12, остановка участия 2026-10-12 — STATE.md Blockers), закрывающий вердикт G-10-1 «осталось» 2026-10-20 записан и approved в 10-UAT.md (awaiting: closed, 0 pending). Поведенческая истина SC1 проверена верификатором напрямую по живому журналу демона.

### Carried Tails (записаны, не блокируют — остановки по решению владельца)

1. **Живые раунды R2–R6, полная лестница P1–P5, двухраундовый RED-оракул в одной форме запуска** — не исполнены (владелец завершил участие в живых тестах 2026-10-12; фоновые окна клавиатурно мертвы на ядре T2 7.2.9-2). Носители: WINDOWS.md #21 (unrun-verify, open), HALT-строки журнала спайка, 09-02-SUMMARY «Not Done». Пути восстановления записаны тогда же (чин ядра t2 / CI-runner desktop / owner-run); диагностическая цель лестницы погашена источник-диагнозом десинка и Wayland-обходом.
2. **Вердикт-матрица первопричины P1–P5** — шапка таблицы в журнале пуста (P1–P5 не шли) — погашена содержательно: D2 Evidence ADR-008 несёт вердикты ветки/сборки, десинк-класс воспроизведён и источник-диагностирован.
3. **Чекбоксы TERM-01/TERM-04 в REQUIREMENTS.md** — Pending при доставленных требованиях (см. Requirements Coverage) — отметить при закрытии.

### Gaps Summary

Отсутствуют. Все пять Success Criteria подтверждены доказательствами на диске и живым состоянием хоста; отклонения формы (5×2-таблица → датированные ряды; two-greens оракул → владельческий вердикт + стенд-ряды Фазы 10; d09-журнал → рассеянные вердикты) суть запроектированные владельцем остановки/ретировки, зафиксированные в ROADMAP-аннотациях, retire_reason и WINDOWS-леджере — не провалы исполнения. Синтеза недостающих рядов нет нигде (проверено чтением журнала: честные aborted-строки стоят как aborted).

---

_Verified: 2026-10-20T17:30:00Z_
_Verifier: Claude (gsd-verifier)_
