---
phase: 09-spayk-i-defekt-wezterm
verified: 2026-11-06T16:11:14Z
status: passed
score: 5/5 must-haves verified
covered_files: [".planning/REQUIREMENTS.md", ".planning/phases/09-spayk-i-defekt-wezterm/09-01-PLAN.md", ".planning/phases/09-spayk-i-defekt-wezterm/09-01-SUMMARY.md", ".planning/phases/09-spayk-i-defekt-wezterm/09-02-PLAN.md", ".planning/phases/09-spayk-i-defekt-wezterm/09-02-SUMMARY.md", ".planning/phases/09-spayk-i-defekt-wezterm/09-03-PLAN.md", ".planning/phases/09-spayk-i-defekt-wezterm/09-04-PLAN.md", ".planning/phases/09-spayk-i-defekt-wezterm/evidence-20261011-validate-daemon.log", "cmd/goswitchctl/main.go", "cmd/goswitchd/main.go", "docs/SPEC.md", "docs/adr/ADR-008-terminal-transport.md", "docs/adr/adr-008-spike-log.md", "internal/ctlsvc/ctlsvc.go", "internal/engine/keys.go", "internal/session/probe.go", "internal/session/probe_test.go", "mise.toml", "test/e2e/case_spike_caps.go", "test/e2e/case_spike_wezterm.go", "test/e2e/fixtures/gs-e2e-wezterm-stand.lua", "test/e2e/fixtures/gs-e2e-wezterm-wayland.lua", "test/e2e/fixtures/gs-e2e-wezterm-xim.lua", "test/e2e/main.go", "test/e2e/wezterm.go", "test/e2e/wezterm_test.go"]
covered_digest: "v4:sha256:1806259a075e12d992da08071e460489648236857b268d94f006b53885af7011"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 5/5
  gaps_closed:
    - "Запрет 09-02 «содержимое pty-захватов и dbus-мониторов НЕ коммитится» (Pitfall 6): регрессия merge c54a388 ЗАКРЫТА лекарством №1 из Gaps Summary прошлого отчёта — merge переделан как ded4c1f (те же родители 5544d7e + f2b9ddc): оба журнала (evidence-20261011-flip2-stuck.log 975 строк, evidence-20261012-t04-stuck.log 66 строк) и 14 файлов .planning/tmp/ выведены из индекса (git diff c54a388..ded4c1f: 16 файлов, −1347 строк, +2 в .gitignore); инвариант 09-02-SUMMARY:121/:189 восстановлен — единственный tracked-журнал фазы снова бесконтентный evidence-20261011-validate-daemon.log (прецедент 4953d78); .gitignore:40-41 добавляет механический guard (.planning/tmp/ и .planning/phases/*/evidence-*.log; check-ignore подтверждает покрытие обоих excised-файлов); история локальна, не публиковалась — экспозиции живого ввода нет; решение владельца не потребовалось — исполнено собственное лекарство запрета (вариант 1 прошлого отчёта), override не вводился"
  gaps_remaining: []
  regressions: []
---

# Phase 9: Спайк и дефект wezterm — Verification Report (STALE-REFRESH + gap-closure)

**Phase Goal:** Переключение раскладки работает в wezterm (ежедневный терминал владельца), а терминальный транспорт закреплён в SPEC и ADR до всякого кода: живой спайк производит verdict-таблицу возможностей, на которой строятся Фазы 10 и 12
**Verified:** 2026-11-06T16:11:14Z
**Status:** passed (5/5 SC, оба требования, все артефакты и связи зелёные; единственный гэп прошлого прогона — запрет-регрессия merge c54a388 — закрыт переделанным merge ded4c1f, см. Gap Resolution)
**Re-verification:** Yes — STALE-REFRESH (прошлый отчёт 2026-11-01T20:58:00Z passed устарел от записей трекинга закрытий Фаз 10–13 и merge) + gap-closure 2026-11-06T16:11:14Z: прогон против c54a388 вернул gaps_found по запрету 09-02 (закоммиченные журналы живого ввода); лекарство №1 исполнено — HEAD теперь ded4c1f, инвариант восстановлен, гэп закрыт этим же прогоном (свидетельства git перечитаны на ded4c1f). Прежние human-пункты закрыты 2026-11-01 (живой ряд на ВМ goswitch-stand, делегирование владельца) — перенесены как закрытые, НЕ переоткрыты.

## Goal Achievement

Контракт фазы — 5 Success Criteria из ROADMAP.md. Диспозиции планов неизменны: 09-01 complete (ac676c4, 2a6e67a); 09-02 HALTED на запроектированной владельцем остановке (2026-10-12, ядро-результаты стоят; живые раунды R2–R6 не требуются — контекст реверификации); 09-03/09-04 superseded (#2349-форма, штампы `status: superseded` + `retired: 2026-10-20` перечитаны в обоих планах). UAT фазы закрыт 2026-10-31 (9/9 pass, mtime/содержимое не менялись); SECURITY: threats_open 0.

**Что изменилось с прошлой верификации (2026-11-01T20:58:00Z), перечитано прогоном против merged-дерева (изначально HEAD c54a388, после лекарства — ded4c1f):**

1. **Код фазы байт-неизменен, кроме одного owner-вердиктного фикса**: `git log 713ac42..HEAD -- docs/SPEC.md docs/adr/ADR-008-terminal-transport.md internal/ctlsvc internal/engine cmd/goswitchctl cmd/goswitchd test/e2e/wezterm.go test/e2e/case_spike_*.go test/e2e/fixtures/gs-e2e-wezterm-{xim,wayland,stand}.lua` — пусто (SPEC и ADR-008 не тронуты ни Фазами 12/13, ни merge). Единственное изменение в internal/session: 5a259ad (2026-11-02 19:19 +0300, fix(session): verifyDead rides app memory) — по вердикту владельца 2026-11-02 «память по приложению» (Orca churn, 5 мёртвых жестов vs 26 коррекций в журнале хоста); classifyTerminal теперь консультирует память класса; юнит-корпус текущего HEAD зелёный под -race (этот прогон, internal/session ok 32.9s). flipTo/флип-машина этим коммитом не ослаблена (TestTerminalClass_VerifyDeadSurvivesClientResetAfterOwnFlip в корпусе).
2. **Журнал спайка вырос чисто аддитивно**: 719 → 800 строк; `git diff 713ac42..HEAD -- docs/adr/adr-008-spike-log.md` = один hunk в конце (:717+), **0 удалённых строк**. Дописаны датированные секции Фазы 12 (сел-проба :720–750, закрытие по вердикту владельца 2026-11-03) и Фазы 13 (a11y-проба :754+, трассер). Ни одна вердикт-строка Фазы 9 не изменена.
3. **Регистрации спайка целы**: test/e2e/main.go :403–407 (spike-smoke/spike-wezterm/spike-caps/wezterm-flip; добавления Фазы 12 term-line-probe :405-область и секции — аддитивны); mise.toml :155/:159/:163/:171 (e2e-spike-*/wezterm-flip). Новый пин-тест test/e2e/spike_journal_test.go (Фаза 12/13) УСИЛИВАЕТ проводку фазы: TestSpikeJournalSectionsExist требует существование каждой секции журнала, включая фазовые.
4. **Трекинг-файлы**: REQUIREMENTS TERM-01/TERM-04 `[x]` Complete (:12/:15, трассируемость :50/:51; TERM-03 закрыт Фазой 12 :54); ROADMAP все четыре consumer-строки «Depends on: Phase 9» целы (:61/:97/:131/:167); STATE.md :87 (стена Telegram AppArmor, живая проба 2026-10-08) и :113 (close-out фазы) на месте.
5. **НАХОДКА прошлого прогона — ЗАКРЫТА ЛЕКАРСТВОМ**: merge c54a388 закоммитил оба «сознательно untracked» журнала демона с живым вводом → merge переделан как ded4c1f (те же родители), журналы и .planning/tmp/ выведены из git, .gitignore получил механический guard — см. Gap Resolution.

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC1: Переключение раскладки работает в wezterm на живом столе (хоткей → последующий ввод даёт символы новой раскладки) | ✓ VERIFIED (закрытый пункт, переносится как закрытый) | Полная цепь: рецепт 7365a93 (2026-10-12), G-10-1 «осталось» 2026-10-20, вердикт владельца 2026-10-31; живой чек на пост-Round-2 дереве ЗАКРЫТ 2026-11-01 живым видимым рядом на ВМ goswitch-stand (владелец делегировал: «Тестируй сам в виртуалке»): демон g2c00c10, wezterm ночника 20261010-083911-e2a98e1b (нативный Wayland/IBus), после собственного флипа демона на ru физические ghbdtn отрисованы панелью как «привет», журнал несёт mode to ru/en + switch_engine (см. `closed:` в архиве pastого отчёта; 11-UAT Test 2). Регрессия на текущем дереве: единственный пост-verbatim коммит в internal/session — 5a259ad (owner-вердикт 2026-11-02), корпус -race зелёный на HEAD (этот прогон). Живой стол не переоткрывается — по контексту реверификации пункт закрыт владельцем |
| 2 | SC2: SPEC несёт датированную spec-delta ревизию D-55 (до кода): слово обязателен, строка feasibility-gated, автокоррекция в терминалах НИКОГДА; двухразрешительная модель; вето в точке армирования, не в apps_blocklist | ✓ VERIFIED | SPEC байт-неизменен с прошлой верификации (0 коммитов 713ac42..HEAD): `grep -c 'rev D-55' docs/SPEC.md` = **10** по всем пяти секциям; «точке/точки армирования» =1, «в терминалах разрешена» =3 — целы; хронология «до кода» неизменна |
| 3 | SC3: ADR-008 доведён до Accepted на свидетельствах живого спайка, не на веб-источниках | ✓ VERIFIED | docs/adr/ADR-008-terminal-transport.md:5 `**Accepted** (2026-10-12, план 10-06)`; **0 веб-ссылок**; 12 ссылок на adr-008-spike-log.md; файл байт-неизменен с прошлой верификации |
| 4 | SC4: Вердикт-таблица возможностей терминалов существует — фиксирует kill-критерий строковой растяжки и чорды слова для Фаз 10/12 | ✓ VERIFIED (рассеянная форма, принятая владельцем) | adr-008-spike-log.md — **800 строк** (было 719; +81 — чисто аддитивные ряды Фаз 12/13, 0 удалений). Секция «Capability-таблица терминалов» :131; десинк-ряды :59/:60/:141/:171/:173/:176–178 + итог «defect-reproduced» :178 целы. Потребители живы: ROADMAP:131 «Depends on: Phase 9 (kill-критерий)» (Фаза 12 — закрыта, потребив таблицу: ряд :750 честно фиксирует, что примитив CopyMode→PRIMARY на живой пробе мёртв и владельцем не запрошен — потребительский вердикт, не регрессия строк фазы); Фаза 10 закрыта на этой таблице |
| 5 | SC5: A11Y-вердикты записаны для деградации MENU-идентичности (Фаза 11) и планирования Фазы 13 | ✓ VERIFIED (рассеянная форма) | STATE.md:87 — стена Telegram snap AppArmor (живая проба 2026-10-08) с гейтом планирования Фазы 13; STATE.md:113 — close-out фазы (вердикты рассеяны); ROADMAP:167 «Depends on: Phase 9 (A11Y-зонд-матрица, AppArmor-вердикт)». Оба пункта прошлого deferred ДОСТАВЛЕНЫ Фазой 13 (complete 2026-11-06): flatpak/лестница D-03 — ряды 07e93ae; ZCode-override + self-check D-06 — 2876219/910ad1f (живая матрица 0/1/2 совпала) |

**Score:** 5/5 истин verified (закрытый SC1 переносится закрытым; запрет-регрессия c54a388 не была истиной — она закрыта лекарством ded4c1f, см. Gap Resolution)

### Deferred Items

Новых deferred-пунктов нет (все фазы вехи закрыты — откладывать некуда). Прежние два deferred прошлого отчёта (flatpak-ряд Telegram; ZCode .desktop-override) доставлены Фазой 13 (complete 2026-11-06, ряды журнала 07e93ae/2876219/910ad1f) — закрыты, больше не числятся.

### Advisory (New Scope, Unevidenced)

None — находка прошлого прогона имела детерминированное git-свидетельство (история коммита c54a388, grep-счётчики содержания), поэтому классифицировалась как блокирующий гэп по гейту #3304 (флаг-файл git-модифицирован после прошлого verified), а не как консультативная запись; теперь она закрыта детерминированным же лекарством ded4c1f (см. Gap Resolution).

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| docs/SPEC.md | rev D-55 в §4.3/§4.4/§5/§7.2/§11 с аудит-трейлами | ✓ VERIFIED | 10 маркеров rev D-55; байт-неизменен с прошлой верификации (0 коммитов) |
| docs/adr/ADR-008-terminal-transport.md | Accepted, D1–D7, Evidence из живого журнала | ✓ VERIFIED | Accepted 2026-10-12 (план 10-06); 0 веб-ссылок; 12 ссылок на spike-log; байт-неизменен |
| docs/adr/adr-008-spike-log.md | Журнал спайка: датированные вердикт-строки, честные aborted | ✓ VERIFIED | 800 строк (719+81, 0 удалений); десинк-ряды :59/:60/:141/:171/:173/:176–178 + defect-reproduced :178 целы; прирост — чисто аддитивные секции Фаз 12/13 |
| internal/session/probe.go + probe_test.go | ParseProbeKey/ProbeForward, строгая грамматика, fail-closed | ✓ VERIFIED | ProbeForward :171; пакет ok под -race в этом прогоне (32.9s) |
| internal/engine/keys.go | KeyHome/KeyEnd вербатим из ibuskeysyms.h | ✓ VERIFIED | 0xff50/0xff57 (:59–60) с дословной цитатой заголовка (:55–56); файл не менялся |
| internal/ctlsvc/ctlsvc.go | ProbeForward с recoverMethod, nil-dep fail-closed | ✓ VERIFIED | errProbeDisabled :50 «probe disabled (daemon not under -debug)»; ok под -race |
| cmd/goswitchctl/main.go, cmd/goswitchd/main.go | подкоманда probe-forward; -debug-гейт | ✓ VERIFIED | оба пакета ok под -race; файлы не менялись с прошлой верификации |
| test/e2e/wezterm.go + fixtures xim/wayland/stand | фикстура wezterm+pty, сентинелы PRIMARY, арминг | ✓ VERIFIED | xim/wayland/stand fixtures на месте, не менялись; wezterm.go не менялся |
| case_spike_wezterm.go, case_spike_caps.go | кейсы spike-smoke/spike-wezterm/spike-caps/wezterm-flip | ✓ VERIFIED | регистрации main.go :403–407 целы; mise-задачи :155/:159/:163/:171 целы; пакеты ok под -race |
| evidence-20261011-validate-daemon.log | бесконтентный журнал демона в git (прецедент 4953d78) | ✓ VERIFIED | tracked (единственный tracked-журнал фазы после лекарства ded4c1f); 0 кириллицы; не менялся |
| evidence-20261011-flip2-stuck.log, evidence-20261012-t04-stuck.log | «сознательно НЕ в git» (09-02-SUMMARY:189) | ✓ VERIFIED (инвариант восстановлен в ded4c1f) | регрессия merge c54a388 закрыта: оба файла выведены из индекса переделанным merge ded4c1f (те же родители), остались на диске untracked; gitignore:41 механически блокирует повторный коммит (check-ignore подтверждён) |
| 09-03-PLAN.md / 09-04-PLAN.md | штампы superseded (#2349) | ✓ VERIFIED | status: superseded + retired: 2026-10-20 в обоих |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| cmd/goswitchctl/main.go | internal/session/probe.go | ProbeForward (CLI → ctlsvc → session → engine.Emitter) | ✓ WIRED | errProbeDisabled :50; ProbeForward :171; цепочка ok под -race |
| docs/adr/ADR-008-terminal-transport.md | docs/adr/adr-008-spike-log.md | Evidence-ячейки D1–D7 → именованные датированные ряды | ✓ WIRED | 12 ссылок; ячейки цитируют конкретные ряды |
| docs/SPEC.md §11 | internal/session (машина армирования) | вето по терминальному классу в точке армирования | ✓ WIRED | SPEC :586 (байт-неизменен); classifyTerminal эволюционировал только owner-вердиктным 5a259ad (память по приложению), корпус зелёный |
| test/e2e/case_spike_*.go | docs/adr/adr-008-spike-log.md | кейсы дописывают датированные вердикт-строки | ✓ WIRED | ПРОВЕРЕНО ТЕСТОМ: TestSpikeJournalSectionsExist (spike_journal_test.go, Фаза 12/13) пинит существование секций и приземление рядов — ok под -race |
| ROADMAP (Фазы 10/11/12/13) | adr-008-spike-log.md (capability-ряды) | kill-критерий / A11Y-вердикт как вход планирования | ✓ WIRED | ROADMAP:61/:97/:131/:167 — все четыре строки целы после закрытий Фаз 12/13 |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| adr-008-spike-log.md | датированные вердикт-строки | живые пробы кейсов + журналы демона | Да (activation-метки, pty-пути, hex) | ✓ FLOWING |
| ADR-008 Evidence D1–D7 | ссылки на ряды | строки журнала спайка | Да (ряды существуют и датированы) | ✓ FLOWING |
| evidence-20261011-flip2-stuck.log | surrounding_text живого ввода | -debug-журнал демона | Да — и ПОТОМУ запретный к коммиту класс; после лекарства ded4c1f снова вне git (untracked на диске, статус-кво всех верификаций до c54a388) | ✓ ЗАПРЕТ СОБЛЮДЁН (гигиена Pitfall 6 восстановлена) |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Сборка | `mise exec -- go build ./...` | exit 0 | ✓ PASS |
| Vet | `mise exec -- go vet ./...` | exit 0 | ✓ PASS |
| Целевой -race прогон covered-пакетов (единственный тест-прогон верификатора) | `mise exec -- go test -race -count=1 ./internal/session/ ./internal/ctlsvc/ ./internal/engine/ ./cmd/goswitchd/ ./cmd/goswitchctl/ ./test/e2e/` | все 6 пакетов ok (session 32.9s, e2e 3.0s) — на дереве merge (c54a388; код-файлы идентичны в ded4c1f — лекарство тронуло только индекс .planning и .gitignore), включая пост-verbatim фикс 5a259ad | ✓ PASS |
| SPEC несёт D-55 во всех 5 секциях | `grep -c 'rev D-55' docs/SPEC.md` | 10 (≥5) | ✓ PASS |
| ADR-008 без веб-источников | `grep -c 'https\?://' ADR-008` | 0 | ✓ PASS |
| Десинк-вердикты в журнале | grep 'reproduced: engine→stale-context' spike-log | :59, :60, :141, :171, :173, :176–178 + defect-reproduced :178 | ✓ PASS |
| Spike-регистрации/задачи пережили Фазы 12/13 | grep main.go/mise.toml | :403–407 / :155/:159/:163/:171 целы; добавления (term-line-probe, secA11yProbe) аддитивны | ✓ PASS |
| Прирост журнала аддитивен | `git diff 713ac42..HEAD -- spike-log` | 1 hunk в конце, 0 удалённых строк | ✓ PASS |
| Запрет-инвариант после лекарства (gap-closure) | `git ls-files .planning/phases/09-*/ \| grep evidence` + `git check-ignore -v` обоих excised-журналов + `git diff c54a388..ded4c1f --stat` | tracked только validate-daemon; оба excised покрыты .gitignore:41; diff: 16 файлов, −1347 строк, +2 (.gitignore) | ✓ PASS |
| Живые пробы (pgrep/apt-mark/journalctl; mise e2e-spike-*) | — | — | ? SKIPPED (директива 6: проверки продукта — только на ВМ goswitch-stand; вердикт-строки существуют от живых сессий, верифицированы чтением) |

### Probe Execution

| Probe | Command | Result | Status |
|-------|---------|--------|-------|
| mise e2e-spike-smoke / e2e-spike-wezterm / e2e-caps / wezterm-flip | — | — | ? SKIPPED (live-GNOME; директива 6) |

Пробы фазы — живые кейсы настольного GNOME; их запуск на хосте запрещён директивой 6. Вердикт-строки этих проб существуют от живых сессий (журнал спайка, датированы) — верифицированы чтением, не перезапуском. Закрытый живый чек SC1 (ВМ-стенд, 2026-11-01) переносится как закрытый и не переоткрывается.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| TERM-01 | 09-01, 09-02 (09-03 superseded → Фаза 10) | Переключение раскладки работает в wezterm | ✓ SATISFIED | REQUIREMENTS.md:12 `[x]`; трассируемость :50 «TERM-01 \| Phase 9 \| Complete»; рецепт 7365a93; G-10-1 «осталось» 2026-10-20; живой ВМ-ряд 2026-11-01 (закрытый пункт) |
| TERM-04 | 09-01 (негатив-свидетели → Фаза 10) | Автокоррекция в терминалах никогда; вето spec-delta, не apps_blocklist | ✓ SATISFIED | REQUIREMENTS.md:15 `[x]`; трассируемость :51 Complete; SPEC §11 rev D-55 (:586) перечитан (байт-неизменен); witness vim/htop fired=0 (журнал); corpus -race зелёный |

Orphaned: нет — фазе 9 назначены TERM-01/TERM-04, обе покрыты планами и закрыты.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| .planning/phases/09-spayk-i-defekt-wezterm/evidence-20261011-flip2-stuck.log | весь файл | (было) Закоммичен живой ввод владельца (surrounding_text из -debug-журнала демона): 396 surrounding-записей, 180 уникальных значений, 179 строк с кириллицей — merge c54a388 | ✗→✓ RESOLVED (ded4c1f) | Файл выведен из индекса переделанным merge; на диске untracked; .gitignore:41 блокирует повтор; экспозиции нет (история локальна, не публиковалась) |
| .planning/phases/09-spayk-i-defekt-wezterm/evidence-20261012-t04-stuck.log | весь файл | (было) Закоммичен без решения тем же merge (бесконтентен: 0 кириллицы, но входил в объявленную untracked-пару) | ⚠️→✓ RESOLVED (ded4c1f) | Тот же вывод из индекса + тот же gitignore-guard |

TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER: 0 совпадений во всех код-файлах covered-набора (вкл. изменённые Фазами 12/13 main.go/mise.toml).

Прочие запреты (перечитаны на текущем дереве):

- «Вето НЕ через apps_blocklist» — VERIFIED (SPEC §11 неизменен; вето в точке армирования);
- «ADR-008 без веб-ссылок» — VERIFIED (0);
- «Ревизии не стирают прежний текст» — VERIFIED (SPEC байт-неизменен с прошлой верификации);
- «go.mod/go.sum не меняются [коммитами фазы]» — VERIFIED (исторически: c21e043^..7365a93 = 0; Фазы 10–13 зависимостей не вносили);
- «Зондовый код не вырастает в продуктовый конвейер» — VERIFIED (dep только под -debug, fail-closed errProbeDisabled :50);
- «Содержимое pty-захватов и живой ввод не коммитятся» — VERIFIED (после лекарства: tracked-журнал фазы единственный и бесконтентный — validate-daemon 4953d78; flip2/t04 вне git, gitignore-guard активен; covered-набор этого отчёта перечисляет только санкционированный журнал).

### Decision Coverage

SKIPPED — в каталоге фазы нет 09-CONTEXT.md (гейт чисто пропускается по правилу).

### Human Verification Required

Открытых пунктов нет. Единственный human-пункт прошлого отчёта (живой флип в wezterm на пост-Round-2 дереве) закрыт 2026-11-01 владельцем через делегированный ВМ-стенд (живой видимый ряд: демон g2c00c10, ghbdtn → «привет» в панели wezterm, журнал mode to ru/en + switch_engine; см. 09-UAT.md / 11-UAT Test 2) — переносится закрытым, не переоткрыт. Решение по найденному гэпу НЕ потребовалось: исполнено лекарство №1 (вычистка) — то самое, которое запрет предписывал как собственное средство восстановления; override-запись владельца не вводилась.

### Gap Resolution

Гэп прошлого прогона (единственный, статус gaps_found при 5/5 SC) — запрет-регрессия гигиены, внесённая merge c54a388 (2026-11-06 18:49:46 +0300): коммит объединения веток Фаз 12/13 молча добавил в git оба журнала демона, которые контракт фазы (09-02-SUMMARY:121/:189, запрет 09-02-PLAN Pitfall 6) держал «сознательно НЕ в git: несут живой ввод владельца» (flip2: 396 surrounding-записей, 180 уникальных значений живого ввода; файлы отсутствовали в обоих родителях merge — внесены при самом слиянии).

**Закрытие (2026-11-06, HEAD ded4c1f, проверено этим прогоном):**

1. **Инвариант восстановлен**: merge переделан как ded4c1f с теми же родителями (5544d7e + f2b9ddc); `git ls-files` по каталогу фазы показывает единственный tracked-журнал — бесконтентный evidence-20261011-validate-daemon.log (прецедент 4953d78). `git diff c54a388..ded4c1f`: evidence-20261011-flip2-stuck.log (−975), evidence-20261012-t04-stuck.log (−66), 14 файлов .planning/tmp/ (−306), .gitignore (+2) — итого 16 файлов, −1347/+2. Оба журнала остались на диске untracked — статус-кво всех верификаций до регрессии.
2. **Механический guard**: .gitignore:40-41 (.planning/tmp/ и .planning/phases/*/evidence-*.log); `git check-ignore -v` подтверждает покрытие обоих excised-файлов — случайный ре-коммит шаблоном исключён (validate-daemon остаётся tracked: gitignore не действует на уже-tracking-файлы).
3. **Экспозиции нет**: история локальна, никогда не публиковалась — живой ввод владельца не покидал машину.
4. **Решение владельца не потребовалось**: из двух лекарств прошлого отчёта исполнено №1 (вычистка) — собственное средство запрета; override (лекарство №2) не вводился, covered-набор отчёта перечисляет только санкционированный бесконтентный журнал.

Проверка воспроизводимости digest: fingerprint пересчитан по covered-набору без двух excised-журналов и повторен после записи отчёта — совпадает (см. covered_digest frontmatter).

---

_Verified: 2026-11-06T16:11:14Z_
_Verifier: Claude (gsd-verifier)_
