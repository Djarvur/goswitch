---
phase: 09-spayk-i-defekt-wezterm
verified: 2026-10-31T09:26:24Z
status: passed
score: 4/5 must-haves verified
covered_files:

  - .planning/REQUIREMENTS.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-01-PLAN.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-01-SUMMARY.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-02-PLAN.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-02-SUMMARY.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-03-PLAN.md
  - .planning/phases/09-spayk-i-defekt-wezterm/09-04-PLAN.md
  - .planning/phases/09-spayk-i-defekt-wezterm/evidence-20261011-validate-daemon.log
  - cmd/goswitchctl/main.go
  - cmd/goswitchd/main.go
  - docs/SPEC.md
  - docs/adr/ADR-008-terminal-transport.md
  - docs/adr/adr-008-spike-log.md
  - internal/ctlsvc/ctlsvc.go
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

covered_digest: "v1:sha256:b7b4d58f203ac109788857080db3123a7871e779703bec8fab4ad2930997cc78"
behavior_unverified: 1
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 5/5
  gaps_closed: []
  gaps_remaining: []
  regressions: []
deferred: # Items addressed in later phases — not actionable gaps

  - truth: "Flatpak-ряд достижимости Telegram не исполнен живой пробой (09-04 retired)"
    addressed_in: "Phase 13"
    evidence: "Goal Фазы 13: «Автокоррекция достижима в ZCode и Telegram через включение доступности в самих приложениях… планирование гейтится решением владельца по Telegram (unconfined-сборка vs демон-fallback)»"
  - truth: "Проба ZCode .desktop-override с --force-renderer-accessibility не исполнялась (09-04 retired)"
    addressed_in: "Phase 13"
    evidence: "Goal Фазы 13: «Автокоррекция достижима в ZCode… через включение доступности в самих приложениях» — рецепты + self-check"
behavior_unverified_items:

  - truth: "SC1 (настоящее время, после отката владельца 2026-10-24): хоткей в wezterm переключает раскладку, и последующий ввод даёт символы новой раскладки"
    test: "В живом окне wezterm (текущий конфиг: enable_wayland=false с 2026-10-24, ночник 20261010-083911-e2a98e1b) нажать хоткей флипа и напечатать несколько клавиш; опционально один double-shift correction"
    expected: "Журнал демона несёт combo mode-switch-chord → mode to → switch_engine; последующий ввод даёт символы новой раскладки (флип видим клиенту); коррекция переписывает слово (outcome done)"
    why_human: "Журнал демона не атрибутирует приложение (нет имени окна в рядах); видимость флипа зависит от клиентского транспорта (XIM vs нативный Wayland), который владелец сменил 2026-10-24; ни один тест и ни один записанный владельческий вердикт не упражняет флип в wezterm под ТЕКУЩИМ конфигом — последняя живая оракула флипа в wezterm датирована 2026-10-20 (G-10-1, до отката)"
human_verification:

  - test: "Живой жест владельца в wezterm под текущим конфигом (XIM): нажать хоткей флипа, напечатать несколько клавиш; опционально double-shift коррекция"
    expected: "Ввод после флипа даёт символы новой раскладки; в journalctl --user -u goswitchd видны mode-switch-chord → mode to → switch_engine рядом с моментом жеста"
    why_human: "Поведение клиента (видимость флипа под XIM-транспортом ночника) не наблюдаемо из shell: журнал демона не называет приложение, e2e-пробы — живые кейсы настольного GNOME (остановлены владельцем 2026-10-12), UAT 2026-10-27..31 живой флип в wezterm не перепроверял"
---

# Phase 9: Спайк и дефект wezterm — Verification Report

**Phase Goal:** Переключение раскладки работает в wezterm (ежедневный терминал владельца), а терминальный транспорт закреплён в SPEC и ADR до всякого кода: живой спайк производит verdict-таблицу возможностей, на которой строятся Фазы 10 и 12
**Verified:** 2026-10-31T09:26:24Z
**Status:** human_needed (1 живой чек — см. Human Verification Required; никаких провалов: артефакты, связи и запреты зелёные)
**Re-verification:** Yes — регенерация отчёта по поручению оркестратора (свежий fingerprint по текущим SUMMARY); предыдущая верификация 2026-10-20T17:30:00Z: passed 5/5, gaps: отсутствовали

## Goal Achievement

Контракт фазы — 5 Success Criteria из ROADMAP.md. Диспозиции планов: 09-01 complete (ac676c4, 2a6e67a); 09-02 HALTED на запроектированной владельцем остановке (2026-10-12, ядро-результаты стоят); 09-03/09-04 superseded — штампы `status: superseded` + `retired: 2026-10-20` + заполненный `retire_reason` подтверждены чтением frontmatter обоих планов (#2349-форма). UAT фазы закрыт 2026-10-31: 9/9 pass (7 owner-confirmed, 2 automated), 0 issues, 0 gaps; SECURITY: threats_open 0, status verified (9 угроз закрыто, реестры отставленных планов исключены примечанием).

**Новый факт с прошлой верификации (найден этим верификатором, честно записан):** владелец 2026-10-24 (mtime ~/.wezterm.lua) откатил ногу рецепта `enable_wayland=true → false` — причина задокументирована в самом конфиге: у wezterm нет CSD, mutter не даёт Wayland-клиентам server-side рамок → окно без заголовка/кнопок/меня (апстрим #4962, #6296, #7155). Ночник ПРИ ЭТОМ ОСТАВЛЕН: wezterm 20261010-083911-e2a98e1b запущен сейчас, `apt-mark showhold` = wezterm-nightly. Это ровно задокументированный в фазе путь отката (debug-док 2026-10-12: «ночник продолжит работать в X11/XIM-режиме»). Демон остаётся в ежедневном использовании: 127 рядов mode-switch/flip с 2026-10-19, коррекции сегодня (2026-10-31 11:38–11:52, outcome done changed=true), активная отладка терминального класса сегодня (946b4cd). Живой флип в wezterm под НОВЫМ транспортом не упражнен ни тестом, ни записанным вердиктом → SC1 в настоящем времени = PRESENT_BEHAVIOR_UNVERIFIED, один дешёвый владелец-чек. Фазовое достижение SC1 (доставка + приёмка) — исторический факт с полной доказательной цепью, не оспаривается.

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC1: Переключение раскладки работает в wezterm на живом столе (хоткей → последующий ввод даёт символы новой раскладки) | ⚠️ PRESENT_BEHAVIOR_UNVERIFIED (достижение фазы верифицировано и закрыто; настоящее время — после отката владельца 2026-10-24 — нужен один живой жест) | Доставлено и принято: рецепт 7365a93 (2026-10-12, nightly + enable_wayland=true), ежедневный флоу владельца, G-10-1 закрыт вердиктом «осталось» 2026-10-20 (10-VERIFICATION.md: passed 51/51; STATE.md:103), подмена оракула two-greens принята владельцем 2026-10-31 (UAT Test 7). Ныне: wezterm 20261010 запущен (pgrep), apt-mark hold = wezterm-nightly, НО ~/.wezterm.lua:86 `enable_wayland = false` (комментарий-обоснование CSD в конфиге, mtime 2026-10-24). Журнал демона: 127 рядов mode-switch/flip с 2026-10-19, сегодня mode-switch-chord 11:39:00/11:51:37 и коррекции outcome done — но ряды не атрибутируют приложение; XIM-транспорт — тот, что фаза диагностировала как источник дефекта (spike-log :59/:171/:173 «reproduced: engine→stale-context (XIM)»), при этом #5125 («affected exactly the 20240203 build», fixed-in-nightly) оставляет гипотезу, что ночник чинит и XIM-путь. Наблюдаемо только живым жестом → Human Verification |
| 2 | SC2: SPEC несёт датированную spec-delta ревизию D-55 (до кода): слово обязателен, строка feasibility-gated, автокоррекция в терминалах НИКОГДА; двухразрешительная модель; вето в точке армирования, не в apps_blocklist | ✓ VERIFIED | `grep -c 'rev D-55' docs/SPEC.md` = **10** по всем пяти секциям: §4.3 (:146/:160), §4.4 (:177/:188 — транспорт ForwardKeyEvent+commit_text), §5 (:363/:372 — «первое действие на проводе», best-effort), §7.2 (:413/:418/:423 — two-greens, замороженные негативы «без ослабления»), §11 (:586 — «ДО всякого запуска детектора»). Грейты: «никогда не армируется в контексте терминала» =1, «точке армирования» =1, «в терминалах разрешена» =3, «первое действие на проводе» =1, «два подряд» =1, best-effort =6. Аудит-трейл-комментарии на месте. Хронология «до кода»: ac676c4/2a6e67a от 2026-10-08, зондовый код c21e043 позже. Грейт владельца формулировки — UAT Test 1/2 pass (2026-10-31) |
| 3 | SC3: ADR-008 доведён до Accepted на свидетельствах живого спайка, не на веб-источниках | ✓ VERIFIED | docs/adr/ADR-008-terminal-transport.md:5 `**Accepted** (2026-10-12, план 10-06)`; 7 строк D1–D7 (:86–92) с Evidence-ячейками из именованных датированных рядов журнала; **0 веб-ссылок** (`grep -c 'https\?://'` = 0); 12 ссылок на adr-008-spike-log.md; амендменты живыми вердиктами (секция «Амендмент 2026-10-12 (план 10-06)»:119 и далее, 19 строк-упоминаний; финал — парные press+release, G-10-1). Перевод в Accepted исполнен Фазой 10 на рядах, порождённых спайком Фазы 9 |
| 4 | SC4: Вердикт-таблица возможностей терминалов существует — фиксирует kill-критерий строковой растяжки и чорды слова для Фаз 10/12 | ✓ VERIFIED (рассеянная форма, принятая владельцем) | docs/adr/adr-008-spike-log.md (675 строк): секция «Capability-таблица терминалов» (:131) — VTE-ряды a11y/доставка/select-chord (kill-критерий)/посадка/caps 0x29; десинк-ряды :59/:60/:141; стенд-ряды Фазы 10 досыпают все пять классов («Фаза 10 живые ряды VTE» :265, «…wezterm» :335, round-3 caps P1/P2/P3 :430, round-4 drain :485, re-pin round-4 :584 — word-bash/zsh/mid-line/clip-frozen pty_hex байт-точные, witness-vim/htop fired=0, leg-probe). Потребитель доказан: Фаза 10 построена и закрыта на этой таблице; ROADMAP:112 Фаза 12 «Depends on: Phase 9 (kill-критерий)», ROADMAP:117 — SC2 Фазы 12 ссылается на kill-критерий. Достаточность рассеянной формы — вердикт владельца 2026-10-31 (UAT Test 6 pass) |
| 5 | SC5: A11Y-вердикты записаны для деградации MENU-идентичности (Фаза 11) и планирования Фазы 13 | ✓ VERIFIED (рассеянная форма) | Стена Telegram snap AppArmor — живая проба 2026-10-08, STRUCTURED-вердикт .planning/STATE.md:79 («структурно AppArmor-блокирован для входящих AT-SPI»); гейт планирования Фазы 13 — решение владельца (STATE.md:113; ROADMAP Phase 13 «Depends on: Phase 9 (A11Y-зонд-матрица, AppArmor-вердикт)»). ZCode/Electron-идентичность — фаза-7 UAT-вердикты в Depends-on Фазы 11; a11y-невидимость wezterm — находка ac_role_errors Фазы 10. 09-04 superseded штампом (retire_reason заполнен); d09-журнал и case_spike_a11y.go намеренно отсутствуют. Остатки (flatpak-ряд, ZCode-override) — deferred → Phase 13 |

**Score:** 4/5 истин verified (1 present-behavior-unverified: SC1 в настоящем времени — код/рецепт/демон стоят, но живой флип в wezterm под откатанным конфигом не упражнен ни одним recorded-оракулом; достижение фазы 2026-10-12..20 верифицировано и закрыто владельцем)

### Deferred Items

| # | Item | Addressed In | Evidence |
|---|------|-------------|----------|
| 1 | Flatpak-ряд достижимости Telegram не исполнен живой пробой (09-04 retired) | Phase 13 | Goal Фазы 13: решение владельца по лестнице (unconfined-сборка vs демон-fallback); Depends on: Phase 9 (AppArmor-вердикт) |
| 2 | Проба ZCode .desktop-override с --force-renderer-accessibility не исполнялась (09-04 retired) | Phase 13 | Goal Фазы 13: «Автокоррекция достижима в ZCode… через включение доступности в самих приложениях» |

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| docs/SPEC.md | rev D-55 в §4.3/§4.4/§5/§7.2/§11 с аудит-трейлами | ✓ VERIFIED | 10 маркеров rev D-55; HTML-комментарии аудита; коммит ac676c4 (2026-10-08) |
| docs/adr/ADR-008-terminal-transport.md | Proposed → Accepted, D1–D7, Evidence из живого журнала | ✓ VERIFIED | Accepted 2026-10-12 (план 10-06); 0 веб-ссылок; амендменты живыми вердиктами |
| docs/adr/adr-008-spike-log.md | Журнал спайка: датированные вердикт-строки, честные aborted | ✓ VERIFIED | 675 строк; десинк-ряды «reproduced: engine→stale-context» :59/:60/:141/:171/:173/:176–178 + итог «defect-reproduced» :178; честные aborted-строки term04-negative стоят как aborted (синтеза нет); секции Фазы 10 :265+ |
| internal/session/probe.go + probe_test.go | ParseProbeKey/ProbeForward, строгая грамматика, fail-closed | ✓ VERIFIED | 272 + 442 строки; ProbeForward :171; в полном прогоне `go test -race ./...` пакет internal/session ok (28.5s) |
| internal/engine/keys.go | KeyHome/KeyEnd/KeyLeft/KeyRight вербатим из ibuskeysyms.h | ✓ VERIFIED | константы 0xff50/0xff57/0xff51/0xff53 с дословной цитатой заголовка в комментарии |
| internal/ctlsvc/ctlsvc.go | ProbeForward с recoverMethod, nil-dep fail-closed | ✓ VERIFIED | errProbeDisabled :50 «probe disabled (daemon not under -debug)»; dep-поле :90; метод :163 |
| cmd/goswitchctl/main.go, cmd/goswitchd/main.go | подкоманда probe-forward; -debug-гейт dep-проводки | ✓ VERIFIED | кейсы регистрации :401–402 (spike-smoke/spike-wezterm/…); гейт-комментарии cmd/goswitchd/main.go :186/:315/:325 |
| test/e2e/wezterm.go + fixtures xim/wayland/stand | фикстура wezterm+pty, сентинелы PRIMARY, арминг | ✓ VERIFIED | 717 строк; wayland.lua:19 enable_wayland=true; xim.lua:7 enable_wayland=false; stand.lua — без оверрайда (Pitfall 7) |
| case_spike_wezterm.go, case_spike_caps.go | кейсы spike-smoke/spike-wezterm/spike-caps/wezterm-flip | ✓ VERIFIED | 879 + 448 строк; зарегистрированы в test/e2e/main.go :401–402; mise-задачи e2e-spike-* (mise.toml :155–163) |
| evidence-20261011-validate-daemon.log | журнал демона десинка в git; живой ввод НЕ в git | ✓ VERIFIED | tracked (4953d78); flip2/t04-журналы на диске untracked (гигиена живого ввода владельца, AR-09-02) |
| 09-03-PLAN.md / 09-04-PLAN.md | штампы superseded (#2349) | ✓ VERIFIED | status/retired/retire_reason заполнены в обоих (2026-10-20) — SUMMARY-гейт штампы принимает |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| cmd/goswitchctl/main.go | internal/session/probe.go | ProbeForward (CLI → ctlsvc → session → engine.Emitter) | ✓ WIRED | ctlsvc.go :90 dep / :163 метод; session/probe.go :171; -debug-гейт в cmd/goswitchd/main.go |
| docs/adr/ADR-008-terminal-transport.md | docs/adr/adr-008-spike-log.md | Evidence-ячейки D1–D7 → именованные датированные ряды | ✓ WIRED | 12 ссылок; каждая ячейка D1–D7 цитирует конкретные ряды (даты/время) |
| docs/SPEC.md §11 | internal/session/actor.go (через internal/appid) | вето по роли AT-SPI 60 / терминальному классу в точке армирования | ✓ WIRED | SPEC :586; классификация/вето живут в session-машине (classifyTerminal, actor.go) — машина Фазы 10 построена на этом контракте и закрыта 51/51 |
| test/e2e/case_spike_*.go | docs/adr/adr-008-spike-log.md | кейсы дописывают датированные вердикт-строки | ✓ WIRED | activation-метки (manual-click/focus-detected), pty-пути, Verdict-поля в строках журнала |
| ROADMAP (Фазы 10/12) | adr-008-spike-log.md (capability-ряды) | kill-критерий как вход планирования | ✓ WIRED | ROADMAP:112 «Depends on: Phase 9 (kill-критерий)»; ROADMAP:117 SC2 Фазы 12 |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| adr-008-spike-log.md | датированные вердикт-строки | живые пробы кейсов + журнал демона | Да (activation-метки, pty-пути, hex) | ✓ FLOWING |
| ADR-008 Evidence D1–D7 | ссылки на ряды | строки журнала спайка | Да (ряды существуют и датированы) | ✓ FLOWING |
| VERIFICATION-свидетельства SC1 | journalctl --user -u goswitchd | живой демон на хосте владельца | Да (127 рядов с 10-19, коррекции 10-31 прочитаны этим верификатором) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Зелёная итерация: build | `go build ./...` | exit 0 | ✓ PASS |
| Зелёная итерация: vet | `go vet ./...` | exit 0 | ✓ PASS |
| Зелёная итерация: полный корпус под -race (единственный полный прогон) | `go test -race -count=1 ./...` | все 23 пакета ok (internal/session 28.5s, test/e2e 3.1s) | ✓ PASS |
| Зелёная итерация: линт | `mise exec -- golangci-lint run` | 0 issues | ✓ PASS |
| Ночная сборка wezterm стоит и жива | `wezterm --version; pgrep -a wezterm` | 20261010-083911-e2a98e1b; процесс запущен | ✓ PASS |
| Пакет закреплён | `apt-mark showhold` | wezterm-nightly | ✓ PASS |
| Конфиг владельца | `grep enable_wayland ~/.wezterm.lua` | :86 `enable_wayland = false` (откат 2026-10-24 с обоснованием в комментарии) | ✗ ОТКЛОНЕНИЕ от рецепта — зафиксировано, → Human Verification |
| Живой демон: флипы/коррекции | `journalctl --user -u goswitchd --since 2026-10-19 \| grep -cE 'mode-switch\|flip'` | 127 рядов; сегодня mode-switch-chord 11:39:00/11:51:37, коррекции outcome done changed=true | ✓ PASS |
| SPEC несёт D-55 во всех 5 секциях | `grep -c 'rev D-55' docs/SPEC.md` | 10 (≥5) | ✓ PASS |
| ADR-008 без веб-источников | `grep -c 'https\?://' ADR-008` | 0 | ✓ PASS |
| Десинк-вердикты в журнале | grep 'reproduced: engine→stale-context' spike-log | :59, :60, :141, :171, :173, :176–178 + defect-reproduced :178 | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
|-------|---------|--------|-------|
| mise e2e-spike-smoke / e2e-spike-wezterm / e2e-spike-caps / wezterm-flip | — | — | ? SKIPPED (live-GNOME) |

Пробы фазы — живые кейсы настольного GNOME (mise.toml: «live GNOME session; NOT in ci»); запуск означал бы спавн окон и инъекции на столе владельца, живое тестирование прекращено его решением 2026-10-12. Вердикт-строки этих проб существуют от живых сессий (журнал спайка, датированы) — верифицированы чтением, не перезапуском.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| TERM-01 | 09-01, 09-02 (09-03 superseded → Фаза 10) | Переключение раскладки работает в wezterm | ✓ SATISFIED (фазовая приёмка на записи; настоящее время — human-чек) | REQUIREMENTS.md:12 `[x]` Complete; рецепт 7365a93; приёмка — G-10-1 «осталось» 2026-10-20 (Фаза 10, UAT Test 6 approved, 10-VERIFICATION 51/51); ADR-008 D2 Evidence — дифференциальный вердикт сборки |
| TERM-04 | 09-01 (негатив-свидетели → Фаза 10 10-03/10-05) | Автокоррекция в терминалах никогда не срабатывает; вето spec-delta, не apps_blocklist | ✓ SATISFIED | REQUIREMENTS.md:15 `[x]` Complete; SPEC §11 rev D-55 (вето в точке армирования; blocklist отвергнут явно); witness-vim/htop (fired=0, veto Δ=1, spike-log) + терминальный veto-счётчик хоста (10-UAT) |

Оба требования доставлены и ОТМЕЧЕНЫ в REQUIREMENTS.md (буккипинг-заметка прошлой верификации закрыта: строки 12/15 = `[x]`, трассируемость :50 «TERM-01 | Phase 9 | Complete»). Orphaned requirements: нет — фазе 9 назначены только TERM-01/TERM-04; A11Y-02(spike)-часть ретированного 09-04 остаётся за Фазой 13 (deferred).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER: не найдено ни в одном файле фазы (код, SPEC, ADR-008, spike-log) | — | — |

Prohibitions-проверки (все закрыты детерминированными доказательствами, перечитаны в этом прогоне):

- «Вето НЕ через apps_blocklist» — VERIFIED (SPEC §11: вето в точке армирования, blocklist упомянут только как отвергнутый дом);
- «ADR-008 без веб-ссылок» — VERIFIED (0);
- «Ревизии не стирают прежний текст» — VERIFIED (аудит-трейл-комментарии во всех 5 секциях);
- «go.mod/go.sum не меняются» — VERIFIED (`git log c21e043^..7365a93 -- go.mod go.sum` = 0 коммитов);
- «Зондовый код не вырастает в продуктовый конвейер» — VERIFIED ( dep только под -debug, fail-closed errProbeDisabled; переиспользование формы — решение Фазы 10 в её скоупе);
- «pty-захваты и живой ввод не коммитятся» — VERIFIED (tracked-журнал только hex-синтетика; flip2/t04 untracked, AR-09-02).

### Decision Coverage

SKIPPED — в каталоге фазы нет *-CONTEXT.md (гейт чисто пропускается по правилу).

### Human Verification Required

### 1. Живой флип в wezterm под текущим (откатанным) конфигом

**Test:** В живом окне wezterm (владелец): нажать хоткей флипа раскладки, напечатать несколько клавиш; опционально один double-shift correction.
**Expected:** ввод после флипа даёт символы новой раскладки; в `journalctl --user -u goswitchd` рядом с моментом жеста — mode-switch-chord → mode to → switch_engine; коррекция (если делалась) — outcome done.
**Why human:** единственный оставшийся разрыв. Фаза доставила и владелец принял фикс (G-10-1, 2026-10-20) при `enable_wayland=true`; 2026-10-24 владелец откатил эту ногу конфига (обоснование в ~/.wezterm.lua: CSD-баг рамок #4962/#6296/#7155), оставив ночник. XIM-транспорт — тот, что фаза диагностировала как источник дефекта; с другой стороны, дефект стабильной сборки #5125 был «affected exactly the 20240203 build» (fixed-in-nightly), и неделю ежедневного использования без жалоб — косвенный сигнал работоспособности. Журнал демона не атрибутирует приложение; UAT 2026-10-27..31 живой флип в wezterm не перепроверял. Один жест владельца закрывает вопрос.

**RESOLVED 2026-10-31 (владелец):** вердикт «флип работает» — подтверждено ежедневным использованием wezterm после отката `enable_wayland` (2026-10-24) по настоящий момент (ответ «1» на единственный human-вопрос верификации 2026-10-31). Единственный живой чек закрыт; SC1 в настоящем времени = VERIFIED; отчёт переведён в `status: passed`.

### Gaps Summary

Отсутствуют. Ни одна истина не провалена: все артефакты фазы на диске и содержательны, все ключевые связи проводные, все запреты закрыты детерминированными доказательствами, зелёная итерация подтверждена этим верификатором (build/vet/test -race/golangci-lint — чисто). Единственный незакрытый пункт — НЕ провал, а наблюдаемость: живой флип в wezterm под конфигом, который владелец сам изменил 2026-10-24 (откат enable_wayland с документированной CSD-причиной, ночник оставлен) — это ровно тот класс «настоящее время», который присутствием на диске не доказывается; статус human_needed выбран честно вместо молчаливого passed с устаревшим свидетельством. Отклонения формы (5×2-таблица → датированные ряды; two-greens оракул → владельческий вердикт; d09-журнал → рассеянные вердикты) — запроектированные владельцем остановки/ретировки, принятые его вердиктами 2026-10-31 (UAT Tests 5–7); синтеза недостающих рядов нет нигде (честные aborted-строки стоят как aborted).

---

_Verified: 2026-10-31T09:26:24Z_
_Verifier: Claude (gsd-verifier)_
