# Roadmap: goswitch

## Overview

Путь от архитектурных решений к работающему продукту: сначала ADR-пакет по всем развилкам (механизм переключения раскладки, тайминги тапов, capability ladder замены, очистка буфера, судьба макросов) и каркас IBus-движка, который видит клавиши как активный input source и переживает рестарты ibus-daemon и паники; затем ядро ценности — коррекция слова EN↔RU с первой зелёной e2e-матрицей; затем полный набор жестов (фраза, выделение, смешанный текст), переключение раскладки и YAML-конфигурация с hot reload; финал — установка без root, производительность и приёмка владельца. Структура следует вехам `docs/SPEC.md` §9 (M0–M4): M0+M1 → Фаза 1, M2 → Фаза 2, M3 → Фаза 3, M4 → Фаза 4.

## Phases

**Phase Numbering:**

- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: ADR-пакет и каркас IBus-движка** - Архитектурные решения по всем развилкам + работающий engine на godbus: регистрация, перехват клавиш, устойчивость к рестартам и паникам (M0+M1) (completed 2026-09-11)
- [x] **Phase 2: Коррекция слова EN↔RU** - Двойной Right Shift исправляет последнее слово в любой раскладке с сохранением регистра, точно по диапазону; e2e-матрица v1 зелёная (M2) (completed 2026-09-15)
- [x] **Phase 3: Фразы, выделение, переключение и конфигурация** - Полный набор жестов коррекции, переключение раскладки, YAML-конфиг с hot reload и goswitchctl; e2e-матрица v2 зелёная (M3) (completed 2026-09-15)
- [ ] **Phase 4: Поставка и приёмка** - Установка без root, производительность (<50 мс / <50 МБ), полная матрица дважды зелёная, ручная приёмка владельца (M4)
- [ ] **Phase 5: Интеграция с GNOME: индикация и двухисточниковое переключение** - Два goswitch-источника в GNOME с нативным индикатором en/ru, переключение через SetGlobalEngine, раскладки оборачиваются из выбора пользователя; ревизия ADR-001
- [x] **Phase 6: Система автокоррекции** - Default-off автокоррекция: гибридный детектор неверной раскладки (словарь + триграммы), политика app×role (unknown → молчание), spec-delta §10/§11 (решение D-51) (completed 2026-10-04)

## Phase Details

### Phase 1: ADR-пакет и каркас IBus-движка

**Goal**: Все архитектурные развилки закрыты утверждёнными ADR, а `goswitchd` работает как IBus engine: видит клавиатурные события как активный input source, логирует их, переживает рестарт ibus-daemon и собственные паники, не конфликтуя с keyd/xremap — фундамент, на котором безопасно строить коррекцию и переключение.
**Mode:** mvp
**Depends on**: Nothing (first phase)
**Requirements**: CORR-08, INTEG-01, INTEG-02, INTEG-03, INTEG-04, INTEG-05, TEST-01, TEST-02, TEST-03, INST-04
**Success Criteria** (what must be TRUE):

  1. ADR-пак утверждён владельцем: механизм переключения раскладки (two-engine регистрация vs внутренний флип — Decision #1), семантика single/double/triple Right Shift против бюджета <50 мс (кандидат — Caramba-подобный «свитч на первый тап»), capability ladder замены текста, триггеры очистки буфера (клик мыши переопределён — на уровне IME ненаблюдаем), механизм и судьба MACR-01 (гейт M0 «ревью владельца»)
  2. `goswitchd` зарегистрирован как IBus engine (`_RegisterComponent` через собственный адаптер `engine/` на godbus) и, будучи активным input source в GNOME Wayland, логирует события клавиш — e2e-скелет (ydotool → структурированный лог, самостоятельная активация целевого окна, fail-fast диагностика) подтверждает гейт M1 «событие видно в логе»
  3. `ibus restart` и `kill -9` движка не роняют ввод рабочего стола: цикл переподключения и перерегистрации с бэкоффом (ibus#2910) и recover-шим на каждом экспортированном D-Bus обработчике работают с первого коммита D-Bus-слоя
  4. Обычный набор идёт транзитом при активном движке — без EVIOCGRAB и захвата клавиатуры, включая сессию с работающим keyd/xremap; демон работает поверх дефолтного IM-стека Ubuntu 24.04 GNOME Wayland без root
  5. Headless CI зелёный: golden-тесты генерируемых `go:generate` таблиц ЙЦУКЕН↔QWERTY (включая знаки `[ ] ; ' , . /`), unit-корпус FSM хоткеев на синтетических потоках событий, логи структурированные с уровнями и debug-трассировкой клавиш

**Plans**: 5/5 plans executed
Plans:
**Wave 1**

- [x] 01-01-PLAN.md — Walking skeleton: goswitchd на шине IBus (адаптер engine/, наблюдатель, recover-шим, цикл перерегистрации, логи) — INTEG-01..05, INST-04

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 01-02-PLAN.md — Чистые пакеты: генерируемые таблицы ЙЦУКЕН↔QWERTY (golden) + FSM тапов (корпус) — CORR-08, TEST-01

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 01-03-PLAN.md — e2e-стенд: инжекция ydotool, AT-SPI фокус, кейсы m1-gate/ibus-restart/kill9-survive, актор FSM в демоне — TEST-02/03, живые INTEG-02..05

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 01-04-PLAN.md — Эксперимент D-01 + ADR-пак 001..005 + гейт M0 (ревью владельца, MACR-01, spec-deltas)

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 01-05-PLAN.md — CI с полным гейтом go-ultimate (tidy+govulncheck), dev systemd unit, README о конфиденциальности -debug, отклонение lint в AGENTS.md — TEST-01, INST-04, INTEG-03

### Phase 2: Коррекция слова EN↔RU

**Goal**: Пользователь двойным нажатием Right Shift исправляет последнее слово, набранное не в той раскладке, в любом поле ввода GNOME — направление определяется автоматически, регистр сохраняется, замена происходит ровно по диапазону неверного текста без визуального прыжка; это подтверждает зелёная e2e-матрица v1.
**Mode:** mvp
**Depends on**: Phase 1
**Requirements**: CORR-01, CORR-04, CORR-05, CORR-07, CORR-09, TEST-04
**Success Criteria** (what must be TRUE):

  1. В gnome-text-editor двойной Right Shift исправляет последнее слово: `ghbdtn`→`привет`, `GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет` (e2e: инжекция ydotool, чтение результата через AT-SPI)
  2. Направление EN→RU / RU→EN определяется автоматически по составу буфера (букв какой раскладки больше) — e2e-кейсы в обе стороны плюс юнит-корпус слов
  3. Замена затрагивает ровно диапазон неверного слова без визуального «прыжка»: `DeleteSurroundingText`+`CommitText` там, где поддержано; fallback Backspace×N с подсчётом в рунах и «abort, не мусорить» там, где нет — минимум один Chromium-кейс в матрице подтверждает fallback-уровень
  4. Буфер очищается по Enter, Tab, Escape и смене фокуса окна — поведение подтверждено эмпирически e2e-кейсами на клиентских классах (GTK/Chromium)
  5. e2e-матрица v1 зелёная: YAML-кейсы «ввод → ожидание» (gnome-text-editor + Chromium), отчёт PASS/FAIL по кейсам, код выхода ≠ 0 при падении

**Plans**: 7/7 planned
Plans:
**Wave 1**

- [x] 02-01-PLAN.md — Чистый пакет internal/correct: буфер-фраза/направление/конверсия/сверка/арифметика лестницы обоих уровней (headless корпус) — CORR-04, CORR-05, CORR-07
- [x] 02-02-PLAN.md — Поверхности стенда: драйверы chromium + gnome-text-editor (fixtures, префлайт, smoke-кейсы) — TEST-04

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 02-03-PLAN.md — Трассер: двойной Right Shift исправляет ghbdtn→привет сквозным конвейером (буфер→направление→сверка→лестница ур.1) + слово после пробела живьём (D-13) — CORR-01, CORR-07

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 02-04-PLAN.md — Флип-режим EN↔RU (Single→флип, потребление+коммит кириллицы, script-true буфер во всех ветвях), RU→EN-коррекция, смешанное слово нетронуто (D-16) — CORR-01, CORR-04

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 02-05-PLAN.md — Лестница ур.2 (Backspace×N по рунам), verify-after, триггеры сброса буфера (CORR-09), живой ladder-chromium на драйвере 02-02 — CORR-07, CORR-09

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 02-06-PLAN.md — YAML-матрица v1: полный словесный набор D-18 (16 кейсов, изоляция кейса свежим демоном, три регистра на gnome-text-editor) — TEST-04

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 02-07-PLAN.md — Self-hosted GNOME-раннер (systemd user unit, ярлык gnome) + workflow e2e-matrix (dispatch-only, bootstrap-регистрация), первый зелёный CI-прогон — TEST-04, D-19

### Phase 3: Фразы, выделение, переключение и конфигурация

**Goal**: Полный словарь жестов — тройной тап исправляет фразу, двойной при выделении исправляет выделенное, смешанный текст корректируется по ADR-семантике; одиночный Right Shift переключает раскладку механизмом по ADR; все хоткеи и таймауты настраиваются YAML-конфигом с hot reload и управляются через `goswitchctl` — всё подтверждает зелёная e2e-матрица v2 полной широты.
**Mode:** mvp
**Depends on**: Phase 2
**Requirements**: CORR-02, CORR-03, CORR-06, SWCH-01, SWCH-02, SWCH-03, SWCH-04, CONF-01, CONF-02, CONF-03, MACR-01, INST-02
**Success Criteria** (what must be TRUE):

  1. Тройной Right Shift исправляет всю фразу из буфера; двойной Right Shift при активном выделении исправляет выделенный текст (включая clipboard round-trip там, где `commit_text` не заменяет выделение)
  2. Смешанный текст корректируется по семантике ADR: конвертируется только часть, набранная не той раскладкой — золотой корпус юнит-тестов плюс e2e-кейсы
  3. Одиночный Right Shift переключает раскладку EN↔RU механизмом по ADR Фазы 1 (two-engine gsettings-путь или внутренний флип за интерфейсом `layoutset`); конфликт таймингов single/double/triple разрешён схемой по ADR (святой порядок: свитч на первый тап, коррекция — повторным тапом в окне); родное Super+Space и MRU продолжают работать, индикатор GNOME актуален; Shift+RightCtrl исправляет слово и переключает раскладку одной комбинацией
  4. Все горячие клавиши, таймауты тапов и параметры задаются в YAML-конфиге (схема биндингов Caramba-совместима по мере возможности); изменения применяются без перезапуска демона (hot reload); `goswitchctl` показывает статус, перечитывает конфиг и корректирует принудительно
  5. e2e-матрица v2 зелёная во всей широте (фразы, выделение, смешанный текст, регистры, разные приложения); MACR-01 реализован — замена Super+Буква → Ctrl+Буква по механизму ADR-005 (в v1, внутри goswitch — решение D-12)

**Plans**: 7/7 planned
Plans:
**Wave 1**

- [x] 03-01-PLAN.md — Трассер: тройной тап исправляет фразу (живой e2e) + run-конвейер смешанного текста D-22..D-24 с золотым корпусом + кап Backspace D-27 — CORR-02, CORR-06
- [x] 03-02-PLAN.md — internal/config: YAML-схема (hotkeys/timeouts/correction/macr), strict-декод D-33, fsnotify hot reload + last-good D-32, -config у демона, docs/CONFIG.md + таблица Caramba — CONF-01..03, SWCH-04

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 03-03-PLAN.md — Выделение: anchorPos-seam, спайк фактической ступени per-surface, геометрия в обе стороны D-30, clipboard-ступень opt-in D-28/D-29 — CORR-03

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 03-04-PLAN.md — Комбо Shift+RightCtrl (слово→флип D-36), живое потребление конфиг-снапшотов (hot reload CONF-02), layout-single/super-space кейсы D-34 — SWCH-01..03, CONF-02

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 03-05-PLAN.md — MACR-01 по ADR-005: спайк доставки Super+Буква, перехват Super→Ctrl, consumed-upstream, per-app appid с деградацией — MACR-01

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 03-06-PLAN.md — internal/ctlsvc (session bus org.djarvur.goswitch) + cmd/goswitchctl status/reload/correct, last-good видимость D-32 — INST-02

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 03-07-PLAN.md — Матрица v2 полной широты (шаги select/combo/reload, три поверхности, живой стол + раннер green106) — приёмка M3, пере-доказывает CORR-02/03/06, SWCH-01..03, CONF-02, MACR-01, INST-02

### Phase 4: Поставка и приёмка

**Goal**: Продукт ставится без root по документированной инструкции, укладывается в бюджет производительности, полная e2e-матрица зелёная дважды подряд на нетронутой сессии, и владелец принял его вручную — релиз v1.
**Mode:** mvp
**Depends on**: Phase 3
**Requirements**: INST-01, INST-03
**Success Criteria** (what must be TRUE):

  1. Установка без root с нуля на чистой Ubuntu 24.04: systemd user unit, регистрация engine в IBus (компонент под `$HOME` через `IBUS_COMPONENT_PATH`, `ibus write-cache`), `go install` или бинарник из GitHub releases; self-check после установки и uninstall работают
  2. Полная e2e-матрица зелёная дважды подряд на нетронутой сессии, с покрытием capability-tier'ов замены (gnome-text-editor, gedit, Chrome, включая оконные режимы Chromium)
  3. Производительность подтверждена измерением: p95 реакции на горячую клавишу < 50 мс (таймстемпы ydotool→AT-SPI в e2e-отчёте), потребление памяти демона < 50 МБ
  4. Ручная приёмка владельца пройдена (одна сессия по §7.2 спеки); README и инструкция установки опубликованы

**Plans**: 7/7 planned
Plans:

- [x] 04-09-PLAN.md

- [x] 04-08-PLAN.md

**Wave 1**

- [x] 04-01-PLAN.md — Трассер: internal/install + goswitchctl install/uninstall («один хозяин», env-cache, полный откат) + живой install-cycle — INST-01, D-39/D-40/D-42
- [x] 04-05-PLAN.md — Матрица v3: спайки + драйверы gedit (GTK3) и chromium-x11, словарь/префлайт, matrix-v3.yaml (superset v2, v2 заморожена) — D-47, критерий №2

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 04-02-PLAN.md — goswitchctl selfcheck (шесть шагов D-41, ремонт env-cache) + прошивка версии (var version, -version, status version=) — INST-01, D-37/D-41
- [x] 04-04-PLAN.md — Perf-приёмка: резидентный событийный свидетель AT-SPI, perf-кейс N=40 комбо-аккорда, p50/p95/p99 + VmHWM, бюджет-гейт exit≠0 — INST-03, D-43/D-44/D-45

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 04-03-PLAN.md — Релизный конвейер: .goreleaser.yaml (дефолтная прошивка), release.yml (тег v*, contents:write только здесь), goreleaser-пин mise — INST-01, D-37/D-38

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 04-06-PLAN.md — Двойной прогон D-48: mise e2e-matrix-v3 + workflow (два прогона подряд, fresh_session-префлайт loginctl) + процедура в docs/ci-runner.md — критерий №2

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 04-07-PLAN.md — Двуязычный README + perf-таблица (D-50/D-46), docs/ACCEPTANCE.md (D-49), сверка WINDOWS-ledger, релиз v1.0.0 (тег → ассеты → каналы; гейт T3 требует доказанного двойного прогона 04-06) — INST-01, критерий №4

## Progress

**Execution Order:**
Phases execute in numeric order: 2 → 2.1 → 2.2 → 3 → 3.1 → 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. ADR-пакет и каркас IBus-движка | 5/5 | Complete    | 2026-09-11 |
| 2. Коррекция слова EN↔RU | 7/7 | Complete    | 2026-09-15 |
| 3. Фразы, выделение, переключение и конфигурация | 7/7 | Complete    | 2026-09-15 |
| 4. Поставка и приёмка | 9/9 | In Progress|  |

## Coverage

30/30 v1 требований замаплено (CORR 9, SWCH 4, CONF 3, INTEG 5, MACR 1, TEST 4, INST 4), без сирот и дубликатов. Таблица трассировки — в `.planning/REQUIREMENTS.md`.

Примечания:

- ADR-пакет Фазы 1 — ворота для всей коррекции и переключения: Decision #1 (two-engine vs внутренний флип) и семантика тапов должны быть закрыты до кода буфера/коррекции.
- recover-шим и цикл перерегистрации — обязательные элементы первого коммита D-Bus-слоя (падение движка = смерть ввода всего рабочего стола).
- MACR-01 — в v1, реализуется внутри goswitch (решение владельца D-12 от 2026-09-10: keyd отвергнут — сложен и ненадёжен; философия goswitch — упрощение конфигурации). ADR-005 Фазы 1 выбирает механизм (идентификация приложений, поведение при перехвате Super системой), судьба MACR-01 больше не открыта.

### Phase 5: Интеграция с GNOME: индикация и двухисточниковое переключение

**Goal**: goswitch переключает раскладку как полноценный гражданин GNOME: в sources прописаны ДВА goswitch-источника (обёрнутые из текущего выбора пользователя), верхняя панель показывает нативный индикатор активной раскладки (en/ru), флип (одиночный Shift, Super+Space, флип после коррекции) переключает источники через SetGlobalEngine на шине IBus — видимым для GNOME образом. Ревизия ADR-001 (Option B «внутренний флип» → Option A «двухдвижковое переключение») фиксируется ADR-006. v1 оборачивает пару us↔ru (раскладка внутри EngineDesc); произвольные пары из xkb — веха v1.1.

**Mode:** standard
**Depends on**: Phase 4
**Requirements**: SWCH-01..04, INTEG-01..05, INST-01
**Success Criteria** (what must be TRUE):

  1. install читает текущие sources пользователя и оборачивает КАЖДЫЙ источник пары us/ru в goswitch-движок его раскладки (`('xkb','us')→('ibus','goswitch-en')`, `('xkb','ru')→('ibus','goswitch-ru')`); пара без us/ru — честный отказ с объяснением; uninstall восстанавливает оригинальный список дословно (ASVS-дисциплина сохранена)
  2. В верхней панели GNOME виден индикатор активной раскладки (en/ru), меняющийся при каждом переключении — без panel-свойств, нативно (2+ источника); живая проверка владельцем
  3. Все жесты переключения (одиночный Shift, Super+Space, флип после коррекции) работают через SetGlobalEngine на ibus-соединении демона (не subprocess), <50 мс; состояние демона не расходится с активным источником (sync при FocusIn; WARN при неудаче переключения)
  4. Режим после коррекции = скрипт результата (правило 2026-09-28) сохраняется: lat→cyr коррекция оставляет активным goswitch-ru и наоборот
  5. GNOME-переключение по клавишам остаётся у goswitch (wm.keybindings-чорды очищены install'ом, восстанавливаются при uninstall — уже в коде); selfcheck отражает двухисточниковое состояние
  6. e2e-матрица зелёная дважды на живом столе в двухисточниковой конфигурации (ночной D-48 гейт), включая кейсы переключения; акт переключения наблюдаем в журнале (`SetGlobalEngine` запись)

**Plans:** 5/5 plans executed

Plans:
**Wave 1**

- [x] 05-01-PLAN.md — живой спайк двухисточникового окна (switch-spike): следует ли GNOME Shell 46 за внешним SetGlobalEngine (индикатор/XKB/сигнал), вердикт владельца + ADR-006 скелет (SWCH-03)
- [x] 05-02-PLAN.md — install оборачивает пару пользователя в два goswitch-движка (D-54): closed-enum wrap, отказы, апгрейд-путь, selfcheck двух источников, живой install-cycle (INST-01, INTEG-03)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 05-03-PLAN.md — шов флипа engine.BindSwitcher → actor.flipTo: все жесты через SetGlobalEngine на шине демона, порядок записей mode→switch_engine, дедлайн без блокировки (SWCH-01/02/04, INTEG-01/05)

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 05-04-PLAN.md — sync-контур: подписка GlobalEngineChanged + FocusIn-имя → Actor.SyncEngine (без flip-loop), activate.IfOwned предпочитает GetGlobalEngine мёртвому current (INTEG-04)

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 05-05-PLAN.md — живые кейсы переключения (two-source-flip, external-flip-sync, ibus-restart), матрица зелёная в двухисточниковой конфигурации, ADR-006 Accepted, README (SWCH-03, INTEG-02, критерий 6)

### Phase 6: Система автокоррекции

**Goal**: Опциональная (default off) автокоррекция — молчаливое исправление слова, набранного не в той раскладке, без горячей клавиши. Безопасность гибридным детектором (словарь набранной раскладки + ремап + словарь другой раскладки; триграммный fallback для OOV) и политикой app×role на живых AT-SPI-данных (unknown → молчание). SPEC §10/§11 правится spec-delta внутри фазы (решение владельца D-51 от 2026-09-27: новый milestone v1.1.0; донор-референс xneur — GPL, только чтение; whatlanggo — MIT, допустим; словарные данные — hunspell/частотники, лицензии проверяет research). Артефакты: 06-RESEARCH.md (перенос 05→06).

**Mode:** standard
**Depends on**: Phase 5
**Requirements**: SPEC §10/§11 (spec-delta), CORR-01..09, MACR-ACL (app×role политика)
**Success Criteria** (what must be TRUE):

  1. Гибридный детектор: слово, набранное в неверной раскладке, опознаётся (словарь + ремап + словарь; OOV — триграммы) с точностью, зафиксированной на корпусе; ложные срабатывания на легитимных словах обоих алфавитов измерены и в бюджете
  2. Политика app×role: автокоррекция применяется только в полях с подтверждённой ролью (AT-SPI живой запрос); unknown → молчание — на терминалах/паролях/редакторах кода коррекции нет
  3. Default off: включается YAML-ключом; hot reload; счётчики автокоррекций в goswitchctl status; каждая автокоррекция логируется (без содержимого, D-20/D-21)
  4. Двойной Shift остаётся ручным переопределением поверх автокоррекции
  5. e2e-матрица расширена автокоррекционными кейсами (вкл/выкл, роль поля, OOV), ночная двойная зелёная

**Plans:** 10/10 plans complete

Plans:
**Wave 1**

- [x] 06-01-PLAN.md — Spec-delta §2/§10/§11 + ADR-007 скелет (D-55: спека ДО кода; D-51 v1.1.0) — SPEC §10/§11, MACR-ACL

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 06-02-PLAN.md — layouts/dictgen: hunspell → golden DictRU/DictEN (sorted slices) + TriRU/TriEN, лицензионные нотисы, dictgen-regen — CORR-08
- [x] 06-03-PLAN.md — internal/appid: (sender,path) + Role(ctx) живой GetRole enum + ErrRoleUnknown — MACR-ACL
- [x] 06-04-PLAN.md — internal/config: секция autocorrect (default off, strict decode, потолки) + docs/CONFIG.md — MACR-ACL

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 06-05-PLAN.md — internal/detect: словарный путь D-52(а) + триграммный fallback D-52(б) + обязательный golden-корпус (0 ложных) — CORR-04/05/06

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 06-06-PLAN.md — Актор: граница слова (Buffer.PushFeed), конъюнкция D-53 fail-closed, счётчики/Status, renderStatus-токены — CORR-01/07/09, MACR-ACL

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 06-07-PLAN.md — e2e: fires / password-silent (GTK4-фикстура) / terminal-silent + матрица v4 (superset v3) + ночной конвейер на v4 — CORR-01, MACR-ACL

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 06-08-PLAN.md — Приёмка: perf p95-гейт (Pitfall 5), двойной регресс v4, README, ADR-007 Accepted — CORR-01..09 (регресс)

**Wave 7** *(gap closure G-6-1, blocked on Wave 6)*

- [x] 06-09-PLAN.md — G-6-1: самоблокировка фабрики на флипе — lock-free AttachEngine + reattach-свидетели фаб-шины + правка ADR-006 — SWCH-01/SWCH-02 (gap G-6-1)

**Wave 8** *(blocked on Wave 7 completion — 06-10 depends_on 06-09)*

- [x] 06-10-PLAN.md — G-6-1: живое доказательство «первая буква после флипа доезжает» — e2e flip-keystroke + журнальный аудит — SWCH-01/SWCH-02 (gap G-6-1; живой оракул гоняется только против исправленного актора 06-09) — выполнен 2026-10-04 (SUMMARY d988560)

### Phase 7: Меню v2 и чёрный список автокоррекции

**Goal:** Довести автокоррекцию до пользовательского качества и выпустить v1.1.0: трей-меню v2 (EN/RU переключение двумя пунктами с отметкой текущего, тумблер автокоррекции с записью в конфиг, живые горячие клавиши из конфига, «Настройки…» = config.yaml в редакторе, версия), чёрный список приложений `autocorrect.apps_blocklist` (regex) с ревизией D-53: белый список `apps` удаляется, автокоррекция стреляет везде, где полевой role-гейт пропускает, кроме совпавших с blocklist; spec-delta §11 + поправка ADR-007 до кода.
**Requirements**: SPEC §11 (spec-delta), MACR-ACL (blocklist-ревизия), CORR-01..09 (регресс), SWCH-01..04 (регресс)
**Depends on:** Phase 6
**Success Criteria** (what must be TRUE):

  1. Меню v2 живьём: EN/RU пункты переключают движок с отметкой текущего; тумблер автокоррекции persist-ит в конфиг (hot reload подхватывает); макросы отображаются актуальными значениями и обновляются на reload; «Настройки…» открывает config.yaml; версия видна.
  2. Blocklist: `apps_blocklist` (regex, подстрока, анкеровка явная) — совпадение запрещает автокоррекцию при любом остальном; пустой = ничего не запрещено; unknown-идентичность приложения НЕ запрет; hot reload; strict decode; docs/CONFIG.md дополнен.
  3. Белый список `autocorrect.apps` удалён из схемы, кода, корпуса и документации; D-53-молчание заменено «blocklist + полевой role-гейт» (spec-delta §11, ADR-007 amendment); регресс-корпус и матричные строки перезакреплены на новую семантику.
  4. Регресс фаз 5-6 зелёный: `mise run ci`, матрица v4 (строки автокоррекции перезакреплены), flip-keystroke 24/24.
  5. Релиз v1.1.0: тег → goreleaser → ассеты; README/CONFIG.md соответствуют v1.1.0.

**Plans:** 8/8 plans complete

Plans:
**Wave 1** *(spec-delta до кода, D-55 — блокирует все кодовые планы)*

- [x] 07-01-PLAN.md — SPEC §11 третья ревизия (blocklist, 2026-10-04) + ADR-007 amendment — SPEC §11
- [x] 07-02-PLAN.md — Схема: apps_blocklist (regex-валидация, пустой паттерн, потолок, ACTIVE-условие) + механическая адаптация актора + docs/CONFIG.md — MACR-ACL
- [x] 07-03-PLAN.md — Персист: SetAutocorrectEnabled/EnsureDocument/DefaultPath (Node round-trip, temp+rename, 0600) + adopt+watch дефолтного пути — MACR-ACL
- [x] 07-04-PLAN.md — Гейт: unknown⇒пропуск (только идентичность), confirm blocklist-only, удаление WR-01-равенства, ensureAppid по enabled — MACR-ACL, CORR-01..09 (регресс)

**Wave 4** *(blocked on 07-03 + 07-04)*

- [x] 07-05-PLAN.md — Меню v2: радиопара EN/RU (SwitchMode→flipTo), тумблер с персистом, живые макро-инфопункты, Настройки (xdg-open no-pipes), О программе; ItemsPropertiesUpdated-динамика — SWCH-01..04 (регресс)

**Wave 5** *(blocked on Waves 1-4; 07-06 и 07-08 параллельны — файлы не пересекаются)*

- [x] 07-06-PLAN.md — e2e: упрощение fires/silents, новый autocorrect-blocklist-silent, Event-драйв case_menu (оба тумблера переживают рестарт), матричная config_base, mise-задачи — MACR-ACL, CORR/SWCH (регресс)
- [x] 07-08-PLAN.md — Звуки переключения: package internal/sound (canberra→paplay no-pipes, комплектные тоны, WARN-best-effort), SoundSink-шов актора (flipTo-флип + fire автокоррекции), gating sound.enabled через hot reload, wiring — SPEC §11 (решение «Звуки при переключении»)

**Wave 6** *(blocked on Waves 1-5)*

- [x] 07-07-PLAN.md — Релизная готовность v1.1.0: README (blocklist, меню v2 со «Звуком», звуки), полный регресс (ci + матрица ×2 + flip-keystroke 24/24 + новые кейсы), RELEASE-READINESS (тег — post-merge действие владельца) — CORR/SWCH (регресс), критерии 4-5

### Phase 8: a11y-магия приложений, документация и конфиг-skill

**Goal:** Разобранный беклог (3 todo от 2026-10-05; порядок волн = порядок зависимостей). (1) **a11y-магия** — конфиг-секция приложений (имена/паттерны), которым goswitch автоматически включает accessibility-магию (gsettings toolkit-accessibility, snap-оверрайды, флаги запуска — механику исследовать при планировании; прецедент ZCode подтверждён владельцем): демон или установщик применяет её сам. (2) **Документация** — сверка README.ru/README/CONFIG.md/SPEC с фактическим поведением v1.1.0 (меню, тумблеры, звуки, blocklist, adopt+watch, «где молчит»), исправление расхождений; реструктура «русский-первый» (README.ru главная, README — перевод, синхронность поддерживается). (3) **Skill «Конфигурация goswitch» для AI** — SKILL.md (.zcode/skills): все секции config.yaml, команды goswitchctl, трей-меню, юнит — ключи/дефолты/диапазоны синхронно с CONFIG.md, процедуры (включить автокоррекцию, blocklist, хоткеи, звуки, a11y-магия), диагностика (goswitchctl status, журнал, известные ограничения); механизм синхрона с доками — генерация из CONFIG.md или lint-гейт.
**Requirements**: backlog-derived (D-8-1..D-8-11 из 08-CONTEXT.md; phase req_ids: null)
**Depends on:** Phase 7
**Plans:** 3/8 plans executed

Plans:
**Wave 1**

- [x] 08-01-PLAN.md — SPEC §4 spec-delta (ревизия 2026-10-05): контракт секции a11y ДО кода (D-55, волна-гейт кодовых планов)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 08-02-PLAN.md — Схема секции a11y (enabled/apps, RE2, потолок 64, Active(), strict decode, hot reload) — D-8-1/D-8-5
- [x] 08-03-PLAN.md — Reconciler internal/a11y (read-verify-then-set toolkit-accessibility + пояс org.a11y.Status.IsEnabled, warn-once) — D-8-3/D-8-6
- [ ] 08-04-PLAN.md — Снапшот ключа в install-state.json + uninstall restore only-if-present — D-8-4

**Wave 3** *(blocked on Wave 2 completion)*

- [ ] 08-05-PLAN.md — Wiring: A11ySink-фолд актора с дифф-гейтом + сборка reconciler'а в goswitchd — D-8-3

**Wave 4** *(blocked on Wave 3 completion)*

- [ ] 08-06-PLAN.md — docs/CONFIG.md русский-первый + сверка со схемой/поведением (CONFIG.md + SPEC в форме 08-01) + секция a11y (restart-семантика, рецепт оверрайда) — D-8-8/D-8-9/D-8-6

**Wave 5** *(blocked on Wave 4 completion)*

- [ ] 08-07-PLAN.md — README русский-первый (README.md = RU полная, README.en.md = EN, README.ru.md удалён), якоря — D-8-7
- [ ] 08-08-PLAN.md — Skill .zcode/skills/goswitch-config/SKILL.md: генератор cmd/skillgen из CONFIG.md + golden-гейт в mise run test + задача skillgen-regen — D-8-10/D-8-11
