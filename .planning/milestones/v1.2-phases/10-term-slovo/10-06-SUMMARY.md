---
phase: 10-term-slovo
plan: 06
subsystem: live-e2e-wezterm-adr
tags: [wezterm-live-rows, two-greens, caps-repin, leg-probe, adr-accepted, terminal-silent-repin, host-install, vm-stand]

requires:
  - phase: 10-term-slovo plan 01
    provides: вердикт площадки vm-primary, стенд goswitch-stand (/tmp/stand-env.sh, rsync-рецепт, unit -debug, арминг по key-trace)
  - phase: 10-term-slovo plan 04
    provides: терминальная ветка записи (чанкированный level-2, done+level-2 оракулы), вердикты владельца OQ 1 unified / OQ 2 contract
  - phase: 10-term-slovo plan 05
    provides: живые ряды VTE (word-bash/word-zsh/clip-frozen — земля для D-ячеек), paced executeLevel2, приём «journal rows пишутся кейсом»
  - phase: 09-spayk-i-defekt-wezterm
    provides: журнал adr-008-spike-log.md (контракт датированных строк), wezterm-nightly 20261010 в ВМ (hold), ADR-008 Proposed с пустыми Evidence-ячейками
provides:
  - Кейс term-word-wezterm: живые ряды wezterm-nightly по pane-pty дельта-оракулу — ×2 двухподряд-дисциплина (Pitfall 1), caps re-pin до решений (наблюдено 0x9), мид-лайн контракт-ряд, замороженный CLIPBOARD-сентинел, OQ 3 leg-probe через probe-forward
  - ADR-008 доведён до Accepted: Evidence D1–D7 заполнены ссылками на датированные ряды журнала (ноль веб-ссылок), амендмент фиксирует вердикты OQ 1/OQ 2, caps re-pin, venue verdict и leg-probe red — критерий Фазы 9, переданный этой фазе
  - Leg-probe вердикт (OQ 3): red — форварднутый select-чорд не поднимает выделение в PRIMARY, лег не подключён; буфер-доверие остаётся единственным механизмом чтения wezterm (граница D7 подтверждена пробой)
  - Re-pin негатива autocorrect-terminal-silent зелёный на площадке (fired=0, ноль correction-записей при обоих режимах доставки IME) — blocked-ряд 10-05 закрыт; гейты кейса переведены на daemon focus_in
  - Сборка фазы v1.1.1-94-ge62a5e4 установлена на хост (goswitchctl install, unit re-point, откат A7 назван); инструкция owner-зачёта (5 пунктов VALIDATION + молчание автокоррекции) доставлена для harvest конца фазы
affects: [phase-verification (критерии 1/3 ROADMAP и критерий Accepted Фазы 9 закрыты живыми рядами; human-check владельца — финальный зачёт), phase-12 (kill-критерий строки: leg-probe red подтверждает отсутствие примитива захвата у ночника)]

# Actuals (#2632) — pairs with the plan's `estimate` (60000) to calibrate future estimates.
actuals:
  tokens: 17652        # chars/4 over the realized diff bb8314e..HEAD (70609 chars, 8 files)
  tasks: 3
  commits: 4           # MEASURED: git rev-list --count gsd-plan-head-before-10-06..HEAD

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Дельта-оракул pty: смещение-снимок до раунда + hex-дельта после — байт-точная земля раунда, иммунная к накоплениюcapture-файла; завершение строки в hex-тексте — литера «0a», не байт"
    - "Арминг-гейт с ограниченной ретрай-лестницей (план §7.2): uinput-churn ydotool реминтит движок — probe-пара может потерять press в disable-окно старого мина; ретрай до floor, не предположение доставки"
    - "Фокус-гейты стенда — ТОЛЬКО daemon focus_in над spawn-базлайном (waitFocusInFrom): AT-SPI-свидетель прибит к gnome-shell глобально и на ночнике проходил бы тривиально; пережило кейс 06-07 (перевод гейтов zenity-контроля и wezterm-ноги)"
    - "Caps re-pin до решений: наблюдаемая подпись контекста читается из журнала демона ПОСЛЕ spawn-базлайна счётчика caps-записей — подпись ночника 0x9 на стенде, решение на caps не ключуется"

key-files:
  created:
    - "test/e2e/case_term_word_wezterm.go (кейс term-word-wezterm: wezterm-фикстура stand-конфигом, caps re-pin, ×2 раунда с learning-handshake, дельта-оракул, сентинел CLIPBOARD, leg-probe)"
    - "test/e2e/fixtures/gs-e2e-wezterm-stand.lua (СТАНД-СВОЙ конфиг ночника — минимальный, без enable_wayland override; ~/.wezterm.lua хоста не тронут)"
  modified:
    - "test/e2e/main.go (регистрация term-word-wezterm: реестр+usage+error-список; autocorrect-terminal-silent — noSurfacePreflight)"
    - "test/e2e/case_autocorrect.go (гейты terminal-silent → waitFocusInFrom: zenity-контроль и wezterm-нога)"
    - "test/e2e/spike_journal_test.go (secTermWez в корпусе секций)"
    - "mise.toml (задача e2e-term-word-wezterm — NOT in ci)"
    - "docs/adr/ADR-008-terminal-transport.md (Status Accepted, Evidence D1–D7 из строк журнала, амендмент-секция, Reversibility)"
    - "docs/adr/adr-008-spike-log.md (секция «Фаза 10 живые ряды wezterm»: caps re-pin, word-a/b, mid-line, clip-frozen, leg-probe ×2 захода, mechanics, re-pin green, host-install)"

key-decisions:
  - "Кейс едет на ОДНОЙ фиксстуре все ряды: выученный verify-dead класс живёт с контекстом ввода (daily-usage форма критерия); на стенде ночник минтит caps=0x9 — коррекция едет без-капс веткой level-2 ПЕРВЫМ же жестом (learning=false), рукопожатие обучения остаётся терпящей формой, не требованием"
  - "Дельта-оракул вместо contains: смещение-снимок до раунда, сравнение hex-дельты НА РАВЕНСТВО — байт-точная земля каждого раунда («привет\n» ровно), накопленные байты прошлых раундов вне сравнения"
  - "Leg-probe исполнен дословно по плану (Home, затем Shift+End keyspec'ами фазы 9 через goswitchctl probe-forward) на МИНИМАЛЬНОЙ стенд-конфигурации ночника — вердикт red честный: без CopyMode-рецепта форварднутый чорд не создаёт выделения; лег НЕ подключается (вердикт OQ 2 «contract» и не требовал его), internal/clipboard.Primary остаётся примитивом без потребителя"
  - "Гейты terminal-silent переведены на daemon focus_in (waitFocusInFrom) — prescription передачи 10-05: AT-SPI-свидетель прибит к gnome-shell глобально; доставка доказывается key-trace floor'ом контроля, оракулы негатива не тронуты"
  - "Установка на хост — существующий процесс goswitchctl install в новый release-dir v1.1.1-94-ge62a5e4 (не перезапись бинарника старого каталога — путь отката v1.1.1 сохранён нетронутым)"

patterns-established:
  - "Вест-механика живых рядов wezterm на стенде: ssh-env DISPLAY=:0+XAUTHORITY, ретрай-лестница арминга, hex-«0a»-форма оракула, caps re-pin по базлайну — строка term-word-wezterm-mechanics"
  - "Двухзелёная дисциплина Pitfall 1 подтверждена на обоих терминалах фазы (VTE 10-05, wezterm 10-06) — одиночный зелёный раунд вердиктом не считается"

requirements-completed: [TERM-02]  # declared by ALL six 10-xx plans — this is the LAST declaring plan; the shared-ID gate opens here (manual owner acceptance rides the human-check harvest)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Кейс term-word-wezterm: wezterm-фикстура на СТАНД-СВОЕМ конфиге (--config-file top-level), регистрации в main.go (реестр+usage+error), mise-задача NOT in ci, секция журнала пиннута корпусом"
    requirement: TERM-02
    verification:
      - kind: other
        ref: "mise run ci зелёный на HEAD (build+vet+golangci-lint v2 strict 0 issues+test -race); grep term-word-wezterm в main.go, e2e-term-word-wezterm в mise.toml"
        status: pass
      - kind: unit
        ref: "test/e2e/spike_journal_test.go — secTermWez в корпусе, заголовок секции существует в журнале (контракт)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Критерий 1 ROADMAP: слово исправляется в wezterm по pane-pty дельта-оракулу — ДВА ПОДРЯД зелёных раунда в одном запуске кейса (word-a 20:34:29 + word-b 20:34:31, повтор официальным прогоном 21:17:16/18), байт-точное «привет\\n», done+level-2 в журнале демона"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#word-a-203429 + #word-b-203431 (pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a оба; exit кейса 0) — повтор независимым прогоном #word-a-211716 + #word-b-211718"
        status: pass
    human_judgment: false
  - id: D3
    description: "Caps re-pin ночника записан строкой журнала ДО решений: наблюдаемая подпись wezterm-nightly 20261010 на стенде — 0x9 (окружающий-бит не анонсируется); легаси-строки исторические, маршрутизация живая (classifyTerminal), решение на caps не ключуется"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#caps-re-pin-203427 (caps=0x9; повтор #caps-re-pin-211714) — запись в журнале датирована, подпись названа"
        status: pass
    human_judgment: false
  - id: D4
    description: "Мид-лайн контракт-ряд по вердикту OQ 2 «contract»: префикс+слово одним флоу, оракул — байт-смежное «echo привет\\n» в дельте"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#mid-line-211720 — pty_delta_hex=6563686f20d0bfd180d0b8d0b2d0b5d1820a (ровно; граница D7 не нарушена)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Критерий 3 ROADMAP: замороженный CLIPBOARD-сентинел байт-нетронут терминальной коррекцией в wezterm-ряду; paste-чорды/commit-over-selection отсутствуют по построению (noRung, 10-04)"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#clip-frozen-211722 — clip_before=clip_after=GS-SPIKE-SENTINEL, pty_delta_hex=полное «привет\\n» (и #clip-frozen-203445)"
        status: pass
    human_judgment: false
  - id: D6
    description: "OQ 3 закрыт вердиктом leg-probe: red — форварднутый select-чорд (Home, Shift+End через probe-forward) не поднимает выделение в PRIMARY; лег не подключён, буфер-доверие остаётся по OQ 2, граница D7 подтверждена пробой; OQ 5 подтверждён открытым precondition'ом (probe-forward принят каналом на стенде)"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#leg-probe-211723 (sentinel=primary_after=GS-SPIKE-SENTINEL; повтор #leg-probe-203447) — честный red, вердикт-строка green|red в журнале"
        status: pass
    human_judgment: false
  - id: D7
    description: "ADR-008 доведён до Accepted на живых свидетельствах (критерий Фазы 9, переданный этой фазе): Evidence D1–D7 — ссылки на датированные ряды журнала, ноль веб-ссылок; амендмент цитирует OQ 1 unified / OQ 2 contract, caps re-pin, venue verdict, leg-probe"
    requirement: TERM-02
    verification:
      - kind: other
        ref: "grep Accepted docs/adr/ADR-008-terminal-transport.md; 7 строк | D[1-7]; счётчик adr-008-spike-log = 10; grep http = 0"
        status: pass
    human_judgment: false
  - id: D8
    description: "Re-pin негатива autocorrect-terminal-silent зелёный на площадке: fired=0, ноль correction-записей при обоих режимах доставки IME (полный counters-оракул), exit 0 — blocked-ряд 10-05 (19:59:37) закрыт"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#autocorrect-terminal-silent-repin-211340 — Verdict: pass; гейты кейса переведены на daemon focus_in (fix 0d70cea)"
        status: pass
    human_judgment: false
  - id: D9
    description: "Сборка фазы v1.1.1-94-ge62a5e4 установлена на хост (goswitchctl install: unit re-point, ibus restart, sources wrapped, engine activated); human-check владельца доставлен — 5 пунктов приёмки критерия 1 VALIDATION + пункт молчания автокоррекции; откат A7 (reinstall v1.1.1) назван в ряду"
    requirement: TERM-02
    verification:
      - kind: other
        ref: "systemctl --user is-active goswitchd = active; journalctl 00:19:41 component registered; ExecStart=…/v1.1.1-94-ge62a5e4/goswitchd; docs/adr/adr-008-spike-log.md#host-install-0019"
        status: pass
    human_judgment: true
    rationale: "Финальное подтверждение критерия 1 — жест владельца на живом столе (VALIDATION Manual-Only): ghbdtn→привет в wezterm и GNOME Terminal, середина строки, «ytn?»→«нет,», молчание автокоррекции при ручных пробах; harvest конца фазы"
  - id: D10
    description: "Flagged assumption плана: формальная классификация края TERM-02 остаётся unresolved (SPEC без ## Edge Coverage) — край wezterm-слова покрыт живыми рядами (×2, мид-лайн, клипборд, caps-дрейф), вердикт классификации — за фазовой верификацией"
    requirement: TERM-02
    verification:
      - kind: other
        ref: ".planning/phases/10-term-slovo/10-06-PLAN.md — Flagged Assumptions: {\"requirement_id\":\"TERM-02\",\"category\":\"unclassified\",\"status\":\"unresolved\"}"
        status: unknown
    human_judgment: true
    rationale: "Flagged assumption плана никогда не авто-разрешается: verifier обязан классифицировать край явно, не тихий pass"

# Metrics
duration: 67min
completed: 2026-10-12
status: complete
---

# Phase 10 Plan 06: живые ряды wezterm ×2, leg-probe, ADR-008 Accepted и owner-зачёт Summary

**Критерий 1 ROADMAP закрыт живыми рядами на ВМ: слово исправляется в wezterm-nightly байт-точно по pane-pty дельта-оракулу — два подряд зелёных раунда в одном запуске (повторено независимым прогоном), caps re-pin ночника записан до решений (наблюдено 0x9), CLIPBOARD байт-нетронут, мид-лайн контракт-ряд зелёный, leg-probe честно red (лег не подключён — D7-граница подтверждена пробой); ADR-008 доведён до Accepted на журнальных свидетельствах D1–D7 без единой веб-ссылки, re-pin негатива terminal-silent стал машинно-зелёным на площадке, сборка фазы v1.1.1-94-ge62a5e4 установлена на хост с путём отката — owner-зачёт (5 пунктов + молчание автокоррекции) доставлен на harvest конца фазы.**

## Performance

- **Duration:** 67 min
- **Started:** 2026-10-12T20:13:38Z
- **Completed:** 2026-10-12T21:20:47Z
- **Tasks:** 3/3 (Task 1 tracer — кейс + caps re-pin живым прогоном, гейт перезапущен end-to-end зелёным; Tasks 2-3 — официальный прогон ×2, re-pin, амендмент ADR, установка на хост)
- **Files modified:** 8 репозиторных (2 новых файла кейса/фикстуры, main.go×2-направления, case_autocorrect.go, spike_journal_test.go, mise.toml, ADR-008, журнал) — measured diff 8 files, +684/−113

## Accomplishments

- **Критерий 1 (слово в wezterm, ×2):** word-a 20:34:29 и word-b 20:34:31 — pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a («привет\n» РОВНО) в обоих раундах одного запуска; независимый повтор 21:17:16/18 — тот же байт-точный результат, exit кейса 0; done+level-2 грейпы пиннят терминальную ветку 10-04. Дисциплина Pitfall 1 (одиночный зелёный = ложная победа) исполнена и на wezterm, и ранее на VTE.
- **Caps re-pin (Pitfall 4):** ряд 20:34:27 записан ДО решений — ночник на стенде минтит контекст с caps=0x9 (окружающий-бит НЕ анонсируется); коррекция едет без-капс веткой level-2 первым же жестом, рукопожатие обучения — терпящая форма; маршрутизация живая, решение на caps не ключуется.
- **Критерий 3:** clip-frozen 20:34:45/21:17:22 — сентинел GS-SPIKE-SENTINEL до и после коррекции байт-равен; paste-чорды и commit-over-selection отсутствуют по построению (noRung verify-after 10-04).
- **Мид-лайн (OQ 2 contract):** mid-line 21:17:20 — pty_delta_hex=6563686f20+привет+0a байт-смежно, префикс не тронут.
- **OQ 3 закрыт:** leg-probe 20:34:47/21:17:23 — sentinel не сменился после форварднутого чорда Home+Shift+End → Verdict: leg-probe red; лег не подключается, буфер-доверие остаётся единственным механизмом чтения wezterm, internal/clipboard.Primary остаётся примитивом без потребителя. OQ 5 подтверждён открытым precondition'ом (probe-forward принят каналом на стенде, отвечает «probe forwarded»).
- **ADR-008 → Accepted:** Status переведён (история Proposed 2026-09 → обязанность 09-03 → исполнено 10-06), все 7 Evidence-ячеек D1–D7 ссылаются на конкретные датированные ряды журнала (venue-probe, VTE-ряды, wezterm-ряды, сентинелы, свидетели), ноль веб-ссылок; амендмент-секция фиксирует OQ 1 unified / OQ 2 contract, caps re-pin, venue verdict vm-primary, двухзелёную дисциплину и leg-probe red.
- **Re-pin негатива машинно-зелёный:** autocorrect-terminal-silent на площадке — fired=0, ноль correction-записей при обоих режимах доставки IME, exit 0 (ряд 21:13:40) — закрывает blocked-ряд 10-05; гейты кейса переведены на daemon focus_in (передача 10-05 исполнена).
- **Хост возвращён в строй:** goswitchctl install из рабочего дерева → release-dir v1.1.1-94-ge62a5e4, unit active, журнал поднятия 00:19:41, откат A7 назван; инструкция owner-зачёта доставлена (harvest конца фазы).
- Строгая итерация: mise run ci зелёный на каждом коммите; новых модулей ноль (go.mod/go.sum нетронуты).

## Task Commits

1. **Task 1 (tracer): кейс term-word-wezterm + caps re-pin живым прогоном + регистрации + mise** - `231c750` (feat); tracer-гейт: verify перезапущен end-to-end зелёным (auto_advance) — ⚡ Tracer verified end-to-end — expanding
2. **Task 2: официальный прогон ×2 + re-pin негатива + ADR-008 Accepted** - `0d70cea` (fix: гейты terminal-silent → daemon focus_in + noSurfacePreflight), `e62a5e4` (docs: амендмент ADR + ряды 21:17 + re-pin-ряд)
3. **Task 3: сборка фазы на хост + owner-зачёт** - `c3b436a` (docs: host-install ряд + откат + human-check доставка)

**Plan metadata:** финальный docs-коммит (SUMMARY+STATE+ROADMAP+REQUIREMENTS)

_Note: кейс-файлы — не behavior-задачи плана (tdd-атрибутов нет): живые вердикты — verify-свидетельство плана; сам кейс прошёл честную эволюцию из четырёх заездов (см. Deviations)_

## Files Created/Modified

- `test/e2e/case_term_word_wezterm.go` — кейс term-word-wezterm: одна фиксстура на все ряды, caps re-pin по базлайну, wezArmGate (ретрай-лестница 3), дельта-оракул (снимок-смещение + hex-равенство), learning-handshake жеста, CLIPBOARD-сентинел, leg-probe через runCtl probe-forward, honest-abort строка
- `test/e2e/fixtures/gs-e2e-wezterm-stand.lua` — СТАНД-СВОЙ минимальный конфиг ночника (без enable_wayland override; ~/.wezterm.lua хоста не тронут)
- `test/e2e/main.go` — term-word-wezterm в реестре/usage/error-списке (noSurfacePreflight, recipeWatchdog); autocorrect-terminal-silent + noSurfacePreflight
- `test/e2e/case_autocorrect.go` — гейты terminal-silent (zenity-контроль и wezterm-нога) переведены на waitFocusInFrom (daemon focus_in над spawn-базлайном)
- `test/e2e/spike_journal_test.go` — secTermWez в корпусе секций
- `mise.toml` — задача e2e-term-word-wezterm (NOT in ci)
- `docs/adr/ADR-008-terminal-transport.md` — Status Accepted, Evidence D1–D7, амендмент-секция, Reversibility переписаны
- `docs/adr/adr-008-spike-log.md` — секция «Фаза 10 живые ряды wezterm»: 13 датированных рядов (caps re-pin ×2, word-a/b ×2, mid-line ×2, clip-frozen ×2, leg-probe ×2, mechanics, re-pin green, host-install)

## Decisions Made

См. key-decisions. Ядро: одна фиксстура на все ряды (выученный класс живёт с контекстом — daily-usage форма критерия 1); дельта-оракул равенством вместо contains (байт-точная земля раунда, иммунная к накоплению файла); leg-probe исполнен дословно по плану на минимальной конфигурации — честный red без попыток «улучшить» пробу конфиг-рецептом (объём лега решает вердикт OQ 2, и он уже сказал «contract»); гейты terminal-silent переведены на daemon focus_in по prescription передачи 10-05; установка на хост — goswitchctl install в НОВЫЙ release-dir (путь отката v1.1.1 сохранён нетронутым).

## Deviations from Plan

**[Rule 3 - Blocking] Арминг-гейт VTE-формы непереносим на ночник: uinput-churn реминтит движок**
- **Found during:** Task 1, прогон 2 (20:30:49: BackSpace press сел в disable-окно старого мина, трейснулся только release — floor не собран)
- **Issue:** каждое ydotool-вызов=create/destroy uinput-устройства; на свежем wezterm-контексте первый вызов реминтит движок и probe-пара теряет половину
- **Fix:** wezArmGate — ограниченная ретрай-лестница (3 попытки), базлайн счётчика пере-снимается на попытку; ровно то, что план предписывал («dummy-клавиша в трейсе до раунда, ограниченные ретраи»), в форме VTE-кейса отсутствовавшей
- **Files modified:** test/e2e/case_term_word_wezterm.go
- **Verification:** попытка 2 зелёная в заезде 20:34; официальный прогон 21:17 — первая попытка
- **Commit:** 231c750

**[Rule 1 - Bug] Оракул сравнивал hex-текст с байтом \n (заезд 20:33:52 — честный FAIL при байт-точной земле)**
- **Found during:** Task 1, прогон 3
- **Issue:** want был corpusHexRU+"\n" (байт перевода строки) против hex-ТЕКСТА дельты — pty_delta_hex был d0bfd180…d1820a (корректен), want — нет
- **Fix:** want в hex-литеральной форме corpusHexRU+"0a" (и в clip-ряде); заезд записан честным FAIL-следом механики
- **Files modified:** test/e2e/case_term_word_wezterm.go
- **Verification:** заезд 20:34 зелёный байт-в-байт
- **Commit:** 231c750

**[Rule 3 - Blocking] Re-pin autocorrect-terminal-silent непроизводим на площадке дважды: preflight требовал chromium, затем zenity-гейт ждал AT-SPI-свидетеля**
- **Found during:** Task 2, прогоны 1-2 (preflight chromium-launch FAIL; затем 30-минутный бюджет waitZenityEntry — witness прибит к gnome-shell, тот же исход, что blocked-ряд 10-05)
- **Issue:** (а) кейс 06-07 предшествовал noSurfacePreflight-флагу 10-05 и требовал поверхности, которых не использует; (б) фокус-гейты zenity-контроля и wezterm-ноги стояли на глобально недостоверном AT-SPI-свидетеле — передача 10-05 прямо предписывала daemon focus_in
- **Fix:** noSurfacePreflight в записи реестра; оба гейта → waitFocusInFrom (базлайн до спавна); оракулы негатива (counters, fired=0) не тронуты; на хосте изменение байт-эквивалентно по исходу (focus_in гейт строже свидетеля)
- **Files modified:** test/e2e/main.go, test/e2e/case_autocorrect.go
- **Verification:** прогон на площадке 21:13 — PASS (fired=0, оба режима доставки), ряд 21:13:40 закрывает blocked-ряд 19:59:37
- **Commit:** 0d70cea

**Строки неудачных заездов затёрты rsync-синхронизациями зеркала (20:30:49 и 20:33:52)**
- **Found during:** Task 2 (слияние рядов)
- **Issue:** rsync хост→ВМ перезаписывает зеркальный журнал копией хостовой — строки заездов 2-3 потеряны из журнала
- **Fix:** причины заездов восстановлены ДОСЛОВНО из сессионного вывода и записаны рядом term-word-wezterm-mechanics с пометкой о затирании; ни один вердикт на них не опирается; слияние рядов финальных прогонов выполнено БЕЗ последующего rsync
- **Files modified:** docs/adr/adr-008-spike-log.md
- **Commit:** 231c750 (mechanics-ряд)

**Total deviations:** 3 auto-fixed (1 Rule 1 — форма оракула; 2 Rule 3 — ретрай-лестница арминга, гейты/noSurfacePreflight terminal-silent) + 1 журнальная реставрация. **Impact:** нулевое на контракт задач — все три задачи исполнены; критерии 1/3 ROADMAP закрыты живыми рядами, ADR-008 Accepted, re-pin негатива стал машинно-воспроизводимым (усиление фазы).

## Issues Encountered

- Четыре захода живого прогона term-word-wezterm до зелёного (20:30:26 ssh-env setxkbmap-refusal — setup-отказ до строк; 20:30:49 armGate-floor; 20:33:52 форма оракула; 20:34 — зелёный): каждый FAIL — первопричина с фиксом, ни одного ретрай-цикла одной формы; официальный прогон Task 2 (21:17) зелёный с первой попытки.
- Re-pin terminal-silent потратил полный 30-минутный бюджет waitZenityEntry на второй заезд (AT-SPI-свидетель) — тот же механизм, что blocked-ряд 10-05; устранён переводом гейтов (0d70cea), следующий прогон зелёный за ~90 с.
- VERSION-строка хостового демона после установки «dev» (сборка рабочего дерева — документированное поведение go-build формы); коммит сборки пиннат журнальной строкой host-install.

## User Setup Required

**Owner-зачёт (human-check, harvest конца фазы) — приёмка критерия 1 на живом столе:**
1. В wezterm набрать `ghbdtn`, двойной правый Shift → увидеть «привет».
2. То же в GNOME Terminal.
3. Середина строки: набрать слово в середине фразы и исправить жестом.
4. Набрать «ytn?» + жест → «нет,» (пунктуация вместе со словом).
5. Убедиться, что автокоррекция молчала в терминале во время ручных проб.
Откат при просьбе владельца: `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install` (релизный каталог на месте).

## Next Phase Readiness

- Фаза 10 закрыта планами: критерии 1/3/4/5 ROADMAP закрыты живыми рядами и юнит-корпусом (1 — этот план, 3 — clip-frozen обоих терминалов, 4 — 10-04, 5 — 10-05), критерий 2 — 10-05; фазовая верификация получает: живые ряды обоих терминалов, ADR-008 Accepted с амендментом, leg-probe red (входное данное kill-критерия Фазы 12 — примитива захвата строки у ночника без CopyMode-рецепта нет), re-pin негатива машинным, установленную на хосте сборку и human-check владельца на harvest.
- Открытые гвосты для верификатора: классификация края TERM-02 (flagged assumption, D10) и ручная приёмка критерия 1 владельцем (D9) — оба human_judgment, не тихий pass.
- Блокеров нет; TERM-02 открывается этим планом (shared-ID gate — последний из шести объявивших планов).

---
*Phase: 10-term-slovo*
*Completed: 2026-10-12*

## Self-Check: PASSED

SUMMARY exists; commits 231c750 / 0d70cea / e62a5e4 / c3b436a verified in git log; ledger recount: 4 commits from plan base bb8314e (gsd-plan-head-before-10-06); production file list matches plan frontmatter (case_term_word_wezterm.go, main.go, mise.toml, ADR-008, spike-log + fixture/journal-test по контрактам пин-теста и секции); stub-скан новых файлов чист (TODO/FIXME/placeholder — ноль); mise run ci зелёный на HEAD; go.mod/go.sum нетронуты; pty-артефакты tmp-only в ВМ, журнал — hex/счётчики.
