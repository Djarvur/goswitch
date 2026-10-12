# Phase 10: TERM-слово — Research

**Researched:** 2026-10-12
**Domain:** терминальный канал ручной коррекции слова (read из собственного буфера демона, write N×BackSpace-реплей + commit_text) на живом демоне goswitchd; дефект конвертации пунктуации; арминг-вето TERM-04; GNOME 46 Wayland (Ubuntu 24.04, T2 MacBook + VM-стенд)
**Confidence:** HIGH для всего, что прочитано в репо или проверено живым в этой сессии; MEDIUM для дизайнерских развилок, где живые свидетельства 2026-10-12 входят в противоречие с черновыми решениями ADR-008 (still Proposed) — эти развилки помечены как чекпойнты владельца

## Summary

Фаза закрывает TERM-02: двойной правый Shift исправляет последнее слово в wezterm и GNOME Terminal через собственный канал демона. **Фундамент жив и проверен владельцем 2026-10-12**: клавиши текут в буфер демона в wezterm (nightly 20261010, нативный Wayland, text-input v3 → mutter → IBus — тот же канал, что у GTK), ConvertRuns приготовляет замену, commit_text видимо садится в новое wezterm [CITED: .planning/debug/2026-10-12-wezterm-nightly-probe.md, раздел ~11:05]. Конвейер останавливает только одна ступень: ADR-004 verify ждёт подтверждения от приложения (свежий SetSurroundingText-пуш или ответ на RequireSurroundingText в бюджет verify_wait_ms=100), а wezterm окружающий текст не передаёт никогда, и контекст mutter на Require не отвечает — журнал: `action n=2` → `correction skipped: verify-timeout` ровно через 101 мс. Отсюда ядро фазы: **заменить апп-верификацию на терминальный канал** — чтение/таргет из собственного буфера нажатий, запись executeLevel2-формы (ForwardKeyEvent BackSpace ×N + commit_text), verify-after «логирует и останавливается, никогда не чинит», бесплатный аборт до первого BackSpace, чанки ≤8 клавиш с перечеком фокуса.

Вторая линия — **дефект ядра коррекции, найденный владельцем 2026-10-12** («ytn?» → «нет?», SPEC требует «нет,»): конвейер слова скармливает конвертеру только буквенный прогон `buf.Token()`, а соседняя пунктуация того же неправильно-раскладочного отрезка едет в `tail` «как набрана» — хотя таблицы содержат все знаки (ENToRU['?']=',') и ConvertRuns монораскладочный диапазон конвертирует ЦЕЛИКОМ с пунктуацией. Дефект не wezterm-специфичен, он во всех приложениях; место правки — геометрия диапазона слова в шве buffer→startRangeCorrection, не конвертер.

Третья линия — **двухразрешительная модель (критерий 5)**: на HEAD вето автокоррекции для терминалов фактически живёт в confirm-стадии (`autoConfirm`: роли 61/79/94 пропущены, всё остальное — отказ), а `autoCorrectBoundary` (точка армирования) роли НЕ проверяет — инвариант D-55 «вето в точке армирования ДО детектора, независимо от наблюдаемости identity» буквально не выполнен. Сегодня терминалы молчат (83 воздержания, 0 срабатываний 2026-10-12), но по побочным причинам (role-unknown у невидимого wezterm, role-forbidden у видимого VTE). Фаза ставит вето в точку армирования и закрепляет замороженными негативными рядами (vim-insert/htop-свидетели).

**Главный не решённый вопрос планирования** — классификация терминального контекста: VTE виден через AT-SPI роль 60, wezterm вне a11y-шины (identity-источника нет вообще), при этом ночная сборка СМЕНИЛА caps-подпись wezterm (0x9 без бита surrounding → бит установлен — это доказано самим фактом verify-timeout: pending создаётся только при бите). Плюс честное противоречие: VTE-коррекция слова УЖЕ работает сегодня на level 1 (живая запись 2026-10-09 21:22:39 в журнале спайка — DeleteSurroundingText в собственном терминале владельца), тогда как черновик ADR-008 D3/D4 предписывает буфер+BackSpace-реплей. Обе развилки — чекпойнты владельца, Research даёт механику и цену каждого варианта.

**Primary recommendation:** строить фазу вокруг одного шва классификации контекста (роль 60 / сконфигурированная identity / поведенческий класс verify-dead), который обслуживает И терминальный конвейер слова, И арминг-вето TERM-04; терминальный конвейер = буфер-таргет + level-2 чанкованный реплей + verify-after без ранга клипборда; дефект пунктуации чинится расширением конвертируемого пробега слова в `internal/correct`; всё под строгим TDD с фейковым эмиттером, живые ряды — в VM-стенде `goswitch-stand` (запущен, проверено живым в этой сессии) с хостовым owner-зачётом.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| TERM-02 | Исправление слова жестом (двойной правый Shift) в wezterm и GNOME Terminal. Чтение — собственный буфер нажатий демона (VTE) / primary-выделение (wezterm, рецепт CopyMode); запись — N×BackSpace-реплей + `commit_text` через канал IBus (не внешний инжектор, не paste), с безопасной точкой отмены до первого разрушающего действия и verify-сверкой | Весь конвейер уже существует в параметрической форме: `startRangeCorrection` (internal/session/actor.go:2745-2811) уже умеет level-2 путь при отсутствии бита caps (executeLevel2, actor.go:2881-2900); недостающее — (1) маршрутизация терминального контекста мимо апп-верификации, (2) чанкирование и перечек фокуса, (3) verify-after-вариант без клипборд-ранга, (4) classify-шов. Механика и развилки — Architecture Patterns 1-3 |
| (попутно, критерии 1/5) | ghbdtn→привет включая середину строки, bash и zsh; автокоррекция молчит при работающей ручной | Дефект пунктуации («ytn?»→«нет?») — Pattern 3; арминг-вето и замороженные негативные ряды — Pattern 4; середина строки — Pitfall 2 и Open Question 2 |
</phase_requirements>

## Project Constraints (from workspace AGENTS.md / conventions)

- **1. Строгий TDD** — каждое поведение red → green → refactor; verify-блоки гоняют тесты на каждом шаге.
- **2. Зелёная итерация** — конец каждой задачи: `go build ./...`, `go test -race ./...`, `golangci-lint run` без ошибок.
- **3. mise, не make** — новые e2e-кейсы = mise-задачи (`NOT in ci` семья, прецедент `e2e-d01`/`e2e-word`).
- **4. golangci-lint строгий** (v2, `linters.default: all`) — никаких послаблений для нового кода.
- **5. GitHub Actions актуальны** — изменений CI фаза не требует (живые ряды — ручные mise-задачи).
- **Стек:** Go ≥ 1.23, stdlib + godbus + минимум зависимостей — **ноль новых модулей** (все механизмы уже в репо или OS-инструменты).
- **Аксиома архитектуры:** инъекция текста только через IBus engine канал; uinput в демоне запрещён; внешние инжекторы и paste-чорды в терминальном пути запрещены (SPEC §4.4 rev D-55, ADR-008 D1/D4).
- **D-55:** спека прежде кода — изменения семантики коррекции оформляются датированной ревизией SPEC/ADR-008 (амендмент), не тихой правкой кода.
- **Процесс:** conventional commits, PR в защищённый main; вывод RESEARCH — по конвенции проекта на русском, термины/идентификаторы в English.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Классификация «этот контекст — терминал» | `internal/session` (шов классификатора при контексте) | `internal/config` (сконфигурированная identity), `internal/appid` (роль 60) | Единственная точка, которую потребляют И ручной конвейер, И арминг-вето; role 60 видит только VTE, wezterm требует конфиг/поведенческий класс |
| Чтение таргета слова в терминале | `internal/session` (a.buf — уже существует) | `internal/correct` (Token/Tail) | Буфер нажатий уже зеркалирует поле через feedKey; нового кода чтения нет |
| Verify слова в терминале | `internal/session` (маршрутизация мимо апп-верификации) | `wl-paste --primary` subprocess, OFF-mutex (WR-01 образец) | VTE пушит surrounding text (существующий verify работает); wezterm — только primary-рецепт или буфер-доверие |
| Запись (замена) | `internal/session` (executeLevel2 + чанкирование + перечек фокуса) | `internal/engine` Emitter (ForwardKeyEvent/CommitText — без изменений интерфейса) | Emitter уже умеет всё нужное; new code — только оркестрация бурста |
| Verify-after терминального пути | `internal/session` (вариант armAfterVerify БЕЗ ранга клипборда) | — | Существующий settleAfter на mismatch армит clipboard rung (Ctrl+V — SIGINT-класс в терминале) — ранг запрещён ADR-008 D4 |
| Конвертация пунктуации со словом | `internal/correct` (геометрия пробега) | `internal/layouts` (таблицы без изменений) | Конвертер уже конвертирует пунктуацию монодиапазона; дефект — в том, ЧТО попадает в конвертер |
| Арминг-вето TERM-04 | `internal/session` (autoCorrectBoundary, до detect.Check) | классификационный шов | Сейчас вето де-факто в autoConfirm (confirm) — инвариант требует точку армирования |
| Замороженные негативные ряды (vim-insert/htop) | `test/e2e` (живые кейсы, NOT in ci) | mise-задачи | Прецедент `autocorrect-terminal-silent` (counters-оракул) |
| CLIPBOARD байт-нетронут | `test/e2e` (замороженный ряд с сентинелом) | `internal/clipboard` (чтение-сентинел через wl-paste) | Контракт D5; сентинел-дисциплина уже в стенде спайка |
| Рецепт CopyMode wezterm | конфиг wezterm (документация/фикстура стенда) | — | Это сторона wezterm, не демона; для owned-конфига — чекпойнт владельца (прецедент 09-03) |

## Standard Stack

### Core (ноль новых Go-модулей — всё уже в репо, прочитано в этой сессии)

| Component | Version | Purpose | Why |
|-----------|---------|---------|-----|
| `internal/session` (actor.go, 2998 LOC) | in-repo | Конвейер коррекции: startCorrection/startRangeCorrection/executeLevel2, verify-машина (pending/pendingAfter), FSM жестов, арминг автокоррекции | Единственное место оркестрации; фаза меняет маршрутизацию и добавляет терминальную ветку ЗДЕСЬ |
| `internal/correct` | in-repo | Чистая библиотека: Buffer (Token/Tail/ReplaceToken), ConvertRuns, BuildPlan, MatchesSuffix/VerifyRangeAt, D-27 cap | Правка пунктуационного дефекта — здесь (геометрия пробега), плюс тесты корпуса |
| `internal/engine` Emitter | in-repo | `ForwardKeyEvent(keyval, keycode, state)`, `CommitText(IBusText)`, `DeleteSurroundingText`, `RequireSurroundingText` [VERIFIED: internal/engine/engine.go:70-76] | Интерфейса хватает на весь терминальный путь; KeyBackSpace=0xff08 и keycode 14 уже есть [VERIFIED: internal/engine/keys.go:30, internal/session/actor.go:53] |
| `internal/clipboard` | in-repo | wl-clipboard subprocess-клиент (`--no-newline`/`--trim-newline`/`--clear`, timeout 1500 мс, stdin-дисциплина) [VERIFIED: internal/clipboard/clipboard.go:16-58] | Флаг `--primary` сегодня НЕ проброшен — чтение PRIMARY для wezterm-лега добавляется сюда (Runner-шов уже тестируемый) |
| `internal/ctlsvc` ProbeForward | in-repo (фаза 9) | Debug-gated форвард клавиш через живой движок; keyspec-грамматика Home/End/Left/Right ± shift/ctrl [VERIFIED: internal/ctlsvc/ctlsvc.go:81-152, internal/session/probe.go:14-60] | Готовый канал управления для select-chord проб (лег primary-чтения) — e2e управляет чордами без ydotool-расширений |
| `internal/appid` | in-repo | Bridge-namespace identity из a11y focus-событий [VERIFIED: internal/appid/appid.go:13-15,145-155] | Источник роли 60 для VTE; для wezterm источника НЕТ (вне шины) — честное ограничение, не баг |
| `internal/config` | in-repo | Строгий YAML-декодер (KnownFields), секции Correction/Autocorrect и т.д. [VERIFIED: internal/config/config.go:79-210] | Новые ключи (терминальная identity, чанки/бюджеты) требуют схему+валидацию+docs/CONFIG.md — иначе валидатор отвергнет документ |
| mise-стенд `test/e2e` | in-repo | Go-раннер кейсов: startVTECapture (gnome-terminal --window -- PROG, pty-захват), startWezterm (--always-new-process), armGate, appendSpikeRow, сентинелы PRIMARY [VERIFIED: test/e2e/case_spike_caps.go:34-130, test/e2e/case_autocorrect.go:751-770] | Живые ряды фазы строятся расширением этой семьи |
| systemd user unit + ydotool 0.1.8 + wl-clipboard 2.2.1 | OS (проверено живым) | Демон активен; инжекция физическим путём; сентинелы клипборда | Без изменений |

### Alternatives Considered

| Вместо | Можно | Tradeoff |
|--------|-------|----------|
| Буфер-доверие + primary-read для wezterm | Ждать апп-верификацию дольше (поднять verify_wait_ms) | Бессмысленно: wezterm не пушит surrounding text ВООБЩЕ (live finding 2026-10-12) — таймаут любой длины лишь затягивает отказ |
| Level-2 реплей в VTE (единообразно) | Оставить VTE на level 1 + surrounding-verify (работает сегодня — живая запись 21:22:39 2026-10-09) | Level-1 строго безопаснее для середины строки (верифицирует реальное поле); но TERM-02 буквой предписывает реплей, а caps VTE лживы исторически (RHBZ 1847347 класс). Развилка — чекпойнт владельца (Open Question 1) |
| Роль 60 как единственный терминальный сигнал | Классификационный шов: роль + конфиг + поведенческий класс verify-dead | Роль 60 не видит wezterm вовсе; без поведенческого класса wezterm-ветка невозможна. Альтернативы нет |
| Расширение диапазона слова в Buffer.Token() | Отдельный метод `ConvertibleStretch()` рядом с Token/Phrase | Правка Token() меняет семантику автокоррекционного payload'а и существующие golden'ы; отдельный метод оставляет D-13 токен-границы нетронутыми (рекомендация) |

**Installation:** ничего не ставится. `go.mod` не меняется. В VM-стенде (см. Environment Availability) потребуется поставить goswitchd + wezterm-nightly .deb + gnome-terminal (есть в distro) + ydotool — это задача стенда, не фазы.

**Version verification (живые пробы этой сессии, 2026-10-12):** wezterm 20261010-083911-e2a98e1b (nightly, apt-mark hold); GNOME Terminal 3.52.0 / VTE 0.76.0; wl-clipboard 2.2.1; go 1.27.1; mise 2026.10.7; golangci-lint 2.13.2 (mise-managed); ядро хоста 7.2.9-2-t2-noble; goswitchd active (user unit).

## Package Legitimacy Audit

**Не применимо — фаза не ставит внешних пакетов.** `go.mod` не меняется (нулевой новый импорт подтверждён обзором всех швов: каждый механизм — in-repo или OS-инструмент).

**Packages removed due to [SLOP] verdict:** none.
**Packages flagged as suspicious [SUS]:** none.

## Архитектурные находки: что именно менять (карта кода)

Все ссылки прочитаны в этой сессии; номера строк — HEAD.

### A. Почему wezterm-коррекция падает (точный механизм)

1. Жест Double → `ExpiryAt` → `startCorrection` → `startRangeCorrection` [VERIFIED: internal/session/actor.go:1043-1061, 2623-2634, 2745-2811].
2. `startRangeCorrection` создаёт `pending` и ждёт апп-верификацию **только если** `a.caps & CapSurroundingText != 0` [VERIFIED: internal/session/actor.go:2779-2810 — ветка `if a.caps&correct.CapSurroundingText == 0 { a.executeLevel2(...); return }` идёт ДО создания pending]. Наблюдаемый verify-timeout поэтому **доказывает**: контекст wezterm-nightly через mutter несёт бит surrounding (старый XIM-клиент репортил 0x9 — без бита — и попадал бы в level-2 сразу; подпись сменилась со сменой сборки).
3. Бюджет `verifyWait` (default 100 мс, ключ `timeouts.verify_wait_ms`) истекает → `VerifyExpiry` → `skipCorrection("verify-timeout")` → `settleCombo()` [VERIFIED: internal/session/actor.go:1069-1082]. Строка ничего не мутирует — отказ безопасный, но коррекции нет.
4. RequireSurroundingText от контекста mutter не отвечает никогда, спонтанные пуши wezterm не шлёт (live finding 2026-09-14 в коде: комментарий к startRangeCorrection, actor.go:2735-2744 «Transport adaptation… neither GTK nor the mutter input context answers RequireSurroundingText — clients push SetSurroundingText spontaneously»; wezterm не пушит вовсе).

Вывод: для wezterm (и любого клиента класса «бит есть, пушей нет») апп-верификация недостижима в принципе — терминальная ветка обязана её обходить, а не ждать дольше.

### B. Почему «ytn?» → «нет?» (дефект пунктуации, точный механизм)

1. Слово-диапазон собирается как `token: a.buf.Token(), tail: a.buf.Tail()` [VERIFIED: internal/session/actor.go:2629-2633].
2. `tokenCapable` = буквы (letterCapable: юникод-буквы или знаки, чья пара по таблице — буква) + цифры; `?` в таблице ЕСТЬ (ENToRU['?']=','), но её образ не буква → `?` НЕ токен-член, она падает в tail [VERIFIED: internal/correct/buffer.go:186-205; internal/layouts/tables.go:44 — `'?':  ','`].
3. Конвертируется ТОЛЬКО token: `ConvertRuns(rng.token)` [VERIFIED: internal/session/actor.go:2754]; tail коммитится «как набрана»: `commit = append(commit, converted...); commit = append(commit, tail...)` [VERIFIED: internal/correct/plan.go:57-61].
4. При этом специка §4.2 требует обратного: «преобразуется КАЖДЫЙ знак таблицы — буквы, скобочный ряд, запятая (клавиша „б“), точка (клавиша „ю“), „?“ (русская запятая на Shift+/)… пунктуация по-прежнему конвертируется вместе со словом» (монораскладочный диапазон) [VERIFIED: docs/SPEC.md:99,118-126].

Замечание: конвертер уже умеет нужное — монопатия ConvertRuns конвертирует ВЕСЬ диапазон с пунктуацией [VERIFIED: internal/correct/runs.go:66-70,94-106]; дефект исключительно в том, какой диапазон ему скармливается. Правка — в шве диапазона (новая геометрия «конвертируемого пробега»: token + соседние руны tail с табличной позицией), с сохранением: D-13 (последнее слово после разделителя), D-14/D-15 (границы токена для автокоррекционного payload), правила смешанного диапазона (пунктуация едет как набрана — spec-delta 2026-10-06 [VERIFIED: internal/correct/runs.go:52-62]), T-03-01-03 (знак вне таблицы своего скрипта отвергает весь диапазон — расширение пробега расширяет и поверхность отказа).

### C. Почему вето TERM-04 не в той точке

- Точка армирования — `autoCorrectBoundary`: проверяет enabled, бит caps, blocklist, запускает детектор `detect.Check` — роли НЕ проверяет [VERIFIED: internal/session/actor.go:2272-2329].
- Ролевое решение живёт в `autoConfirm` (асинхронный confirm, после детектора): `switch role { case acRoleText(61), acRoleEntry(79), acRoleDocumentText(94): … default: recordACAbstain(acReasonRoleForbidden) }` — роль 60 (терминал) уходит в default [VERIFIED: internal/session/actor.go:82-86 (константы), 2447-2540 (switch)]. У невидимого wezterm запрос роли падает → role-unknown отказ [VERIFIED: internal/session/actor.go:2458-2466].
- Инвариант D-55: «Вето проверяется в точке армирования автокоррекции ДО запуска детектора; …не зависит от наблюдаемости identity» [VERIFIED: docs/SPEC.md:569-578]. Поведенчески терминалы молчат (83 воздержания/0 срабатываний 2026-10-12), но буквальное исполнение инварианта требует переноса вето в `autoCorrectBoundary` — чем и является критерий 5 этой фазы.

## Recommended Plan Decomposition (черновик для планировщика)

**Три плана, две волны** (config: `parallelization: true`, `granularity: fine`; TDD-директива):

| План | Содержание | Depends on |
|------|-----------|------------|
| **10-A: ядро коррекции + классификация** (unit-first) | (1) дефект пунктуации: конвертируемый пробег слова в `internal/correct` + golden-корпус («ytn?»→«нет,», регистры, смешанный диапазон не меняется, `é`-отказ расширяется честно); (2) классификационный шов терминального контекста (роль 60 / конфиг-identity / verify-dead-класс) + арминг-вето в `autoCorrectBoundary`; (3) D-24-наблюдаемость (дискриминатор done-with-changes vs done-unchanged на INFO, без содержимого — D-20/D-21); (4) чанкированный executeLevel2 с перечеком фокуса + verify-after-вариант без клипборд-ранга (фейковый эмиттер пинит форму бурста) | — |
| **10-B: живые ряды VTE** | Кейс term-word-vte: gnome-terminal-фикстура (pty-захват), ydotool набор + жест, оракул pty == привет; bash и zsh; середина строки (по решённому OQ 2); замороженный ряд CLIPBOARD-сентинел байт-нетронут; негативные ряды vim-insert/htop (автокоррекция молчит при зелёной ручной); re-pin `e2e-autocorrect-terminal-silent` | 10-A |
| **10-C: живые ряды wezterm + owner-зачёт** | Кейс term-word-wezterm (nightly, два подряд зелёных раунда — дисциплина арминг-гонки); лег primary-чтения (CopyMode-рецепт в фикстуре стенда; wl-paste --primary через internal/clipboard + ProbeForward select-chord) — по решённому OQ 2/3; ADR-008 амендмент (Evidence-ячейки D1-D7 заполняются строками журналов, Proposed→Accepted); owner UAT | 10-B |

Развилки OQ 1 (VTE: level-1 остаётся vs единый level-2) и OQ 2 (середина строки: механика защиты) влияют на объём 10-A/10-B — закрепить чекпойнтом владельца ДО планирования задач.

## Architecture Patterns

### System Architecture Diagram — терминальная коррекция слова (целевое состояние)

```
[владелец: печатает ghbdtn в wezterm / gnome-terminal]
        │ (физическая клавиатура)
        ▼
┌─────────────────── ibus-daemon ────────────────────┐
│  ProcessKeyEvent → goswitch engine (goswitch-en)    │
│        │                                            │
│        ▼                                            │
│  feedKey: printable → pushAndArm(r) → buf = "ytn?"  │
│  (автокоррекционный boundary: classifyContext()=    │
│   terminal → ВЕТО до detect.Check — TERM-04)        │
└────────┬────────────────────────────────────────────┘
         │ двойной правый Shift → FSM Double → ExpiryAt
         ▼
┌───────────── startRangeCorrection (терминальная ветка) ─────────────┐
│ 1. ТАРГЕТ: buf.Token() + конвертируемый пробег → "ytn?"             │
│ 2. ConvertRuns → "нет," (монодиапазон: пунктуация вместе)           │
│ 3. ПРЕ-ВЕРИФИЯ (класс-зависимо, до первого разрушающего действия):  │
│    • VTE (пушит surrounding): MatchesSuffix по свежему пушу          │
│      (как сегодня — работает, живая запись 2026-10-09)               │
│    • wezterm (verify-dead): (a) буфер-доверие ИЛИ                    │
│      (b) select-chord → wl-paste --primary (OFF-mutex, WR-01) →      │
│      сравнение с токеном; mismatch → тихий отказ, ничего не тронуто  │
│ 4. ЗАПИСЬ (one-way, чанки ≤8 ForwardKeyEvent BackSpace,              │
│    между чанками перечек фокуса: FocusOut → стоп)                    │
│    + ОДИН CommitText("нет,")  → буфер ReplaceToken                   │
└────────┬─────────────────────────────────────────────────────────────┘
         ▼
┌───────────── verify-after (терминальный вариант) ───────────────────┐
│ свежий пуш (VTE): suffix == converted+tail?                          │
│   match → тихо; mismatch → INFO-запись и СТОП — никогда не «чинит»,  │
│   НИКОГДА не армит clipboard rung (paste-чорд = SIGINT-класс)        │
│ пушей нет (wezterm): тихий таймаут (послеExpiry-семантика)           │
└──────────────────────────────────────────────────────────────────────┘
```

Точки входа: физический набор и жест (FSM). Решения: классификация контекста → выбор verify-ветки; все отказы до первого BackSpace бесплатны. Внешние зависимости: ibus-daemon/mutter (транспорт), wl-paste (только wezterm-лег, вне мьютекса).

### Pattern 1: Классификационный шов терминального контекста (центральная задача фазы)

**What:** одно место, отвечающее «этот input context — терминал?», с тремя источниками признака:

| Источник | Покрывает | Механизм | Статус свидетельства |
|----------|-----------|----------|----------------------|
| AT-SPI роль 60 | GNOME Terminal/VTE | живой GetRole в момент решения (источник internal/appid уже подключён актору через UseRole) [VERIFIED: internal/session/actor.go:572-580] | роль 60 = `RoleTerminal // terminal — never corrected` [VERIFIED: internal/appid/appid.go:73-79 через 09-RESEARCH; константы ролей в actor.go:82-86] |
| Сконфигурированная identity | терминалы, видимые a11y (декларативно), будущие | ключ конфига (напр. `autocorrect.terminals` / `correction.terminal.apps`) — строжий декодер требует схему+docs/CONFIG.md | прецедент формулировки инварианта: «роль AT-SPI 60 ИЛИ сконфигурированный терминальный движок/identity» [VERIFIED: docs/SPEC.md:572-574] |
| Поведенческий класс verify-dead | wezterm-nightly (вне a11y, бит surrounding есть, пушей нет никогда) | обучение на факте: verify-раунд этого контекста истёк по таймауту при живом контексте → контекст помечен verify-dead на время жизни контекста (FocusOut/Reset снимает) | единственный рабочий признак для wezterm: identity-источника нет (appid строится ТОЛЬКО из a11y bridge-путей [VERIFIED: internal/appid/appid.go:145-155]), а бит caps у ночника установлен — доказано фактом verify-timeout (pending создаётся только при бите [VERIFIED: internal/session/actor.go:2779-2810]) |

**When to use:** потребители — (1) маршрутизация ручной коррекции (терминальная ветка verify/записи), (2) арминг-вето TERM-04 в `autoCorrectBoundary`. Один шов, два потребителя — иначе вето и конвейер разъедутся.

**Развилка, которую закрывает владелец (OQ 1):** VTE пушит surrounding text и коррекция слова на level 1 работает СЕГОДНЯ (запись в журнале спайка: «21:22:39 — живая коррекция … в собственном терминале владельца, уровень 1» [CITED: docs/adr/adr-008-spike-log.md:24]). Варианты: (а) единый терминальный канал level-2 для обоих терминалов (буква TERM-02, черновик ADR-008 D3/D4); (б) VTE сохраняет существующий level-1 путь с настоящей апп-верификацией (строго безопаснее для середины строки), терминальная ветка — только для verify-dead контекстов (wezterm). Research не решает за владельца; ADR-008 ещё Proposed — амендмент легален.

### Pattern 2: Терминальный конвейер записи — чанкированный level-2 с перечеком фокуса

**What:** executeLevel2 существует, но сегодня это один слепой бурст: `for range plan.Backspaces { eng.ForwardKeyEvent(engine.KeyBackSpace, backSpaceKeycode, 0) }` + один commit [VERIFIED: internal/session/actor.go:2892-2895] — без чанков, без перечека фокуса, без verify-after. Критерий 4 требует: чанки ≤8 клавиш, между чанками перечек фокуса (реальный FocusOut → стоп раунда), verify-after «логирует и останавливается».

**Существующая база для перечека и verify-after:**
- Жизненный цикл: `HandleLifecycle(FocusOut/Reset)` уже ретирит pending и чистит буфер [VERIFIED: internal/session/actor.go:770-812] — чанк-граница проверяет «контекст ещё тот» (поколение контекста/эпоха раунда), а не изобретает новый механизм;
- Verify-after машина уже есть: `armAfterVerify` (ожидание suffix == converted+tail, WR-05-дебаунс первого mismatch, тихий таймаут) [VERIFIED: internal/session/actor.go:1578-1629, settleAfter 1528-1549]. **Обязательная дельта:** терминальный вариант НЕ армит clipboard rung — сегодня mismatch возвращает `round.paste` и ставит условие `a.opts.ClipboardRung && …` [VERIFIED: internal/session/actor.go:1548] — в терминале ранг запрещён по построению (ADR-008 D4: Ctrl+V в шелле = SIGINT-класс), не только дефолтом конфига.

**Точка аборта:** бесплатность аборта до первого BackSpace структурная — все отказы (empty-buffer, convert-failed, no-engine, backspace-cap D-27=50, верификация) происходят до эмуляции бурста [VERIFIED: internal/session/actor.go:2748-2776, correct/plan.go:77-82]. После первого BackSpace раунд one-way: реплей+commit завершаются, дальше только лог.

**Example (форма чанка, псевдокод поверх существующего Emitter):**
```go
// Терминальная запись: чанки ≤8, между чанками — контекст ещё тот?
const terminalChunk = 8
for done := 0; done < plan.Backspaces; {
    n := min(terminalChunk, plan.Backspaces-done)
    for range n {
        eng.ForwardKeyEvent(engine.KeyBackSpace, backSpaceKeycode, 0)
    }
    done += n
    if done < plan.Backspaces && a.contextGeneration != gen0 {
        a.stopRoundNoRepair() // фокус ушёл между чанками: стоп, verify-after логирует
        return
    }
}
eng.CommitText(engine.NewIBusText(string(plan.Commit)))
```

### Pattern 3: Конвертируемый пробег слова (фикс пунктуации)

**What:** геометрия диапазона слова расширяется от «буквенный токен» до «максимальный пробег знаков с табличной позицией, примыкающий к токену» (хвостовая сторона — живой кейс «ytn?»; головная сторона — «(ytn»/«»ytn», скобочный ряд SPEC §4.2). Рекомендация — НОВЫЙ метод рядом с Token/Tail (не переписывать Token(): семантика D-13/D-14 автокоррекционного payload'а и десятки golden'ов остаются байт-в-байт).

**Инварианты, которые нельзя сломать:**
- D-13: «ghbdtn » → «привет » — пробел не в таблице, едет в tail как набран [VERIFIED: correct/plan.go:44-56];
- Смешанный диапазон: пунктуация едет как набрана (у неё нет посимвольного направления) [VERIFIED: correct/runs.go:52-62,109-118] — расширение пробега обязано сначала определить моно/микс ВСЕГО пробега, потом конвертировать;
- T-03-01-03: знак вне таблицы своего скрипта отвергает весь диапазон — расширенный пробег расширяет отказ честно (документировать в golden'е);
- Регистр по руне, биективность (DR-правка §4.2 2026-09-28) — не трогаются.

### Pattern 4: Арминг-вето TERM-04 (критерий 5)

**What:** в `autoCorrectBoundary` ПЕРЕД `detect.Check` добавляется терминальная проверка (Pattern 1): контекст классифицирован терминалом → воздержание (закрытый slug, напр. `terminal-veto`), детектор не запускается, confirm не назначается. Confirm-стадия оставляет существующую ролевую логику как второй пояс (61/79/94 — только они и сегодня [VERIFIED: internal/session/actor.go:2447-2540]).

**Замороженные негативные ряды:** прецедент — `autocorrect-terminal-silent` (wezterm, counters-оракул: autocorrect_enabled=true, fired=0, positive-control keystroke против глухоты) [VERIFIED: test/e2e/case_autocorrect.go:779-810]. Фаза добавляет vim-insert/htop-свидетелей ВНУТРИ gnome-terminal: автокоррекция молчит, ручная коррекция в том же классе контекста работает. Осторожно: htop/vim — TUI, для РУЧНОЙ коррекции вне контракта (D7), свидетель только про молчание автокоррекции, не про ручную.

### Pattern 5: D-24-наблюдаемость (побочный вопрос живого находждения 14:19:08)

**What:** «instant-done» и исполненная коррекция пишут ИДЕНТИЧНУЮ INFO-строку `correction outcome=done` — различающий DEBUG-репкорд (level/source/result) на INFO не виден [VERIFIED: internal/session/actor.go:2761-2769 (D-24-ветка) vs 2907-2916 (logCorrectionDone)]. Журнальная триажная метка «instant-done» у исследователей 2026-10-12 не могла отличить «нечего менять» от «исполнилось, деталей не видно».

**Сверка с §4.3 по коду (ответ на побочный вопрос находждения):**
- D-24-ветка (done без изменений) НИКОГДА не флипает: `settleCorrectionFlip` в ней не вызывается; флипает только settleCombo — а он стреляет лишь при жесте-КОМБО Shift+RightCtrl (`comboPending=true` ставится только в combo-ветке feedKey [VERIFIED: internal/session/actor.go:2149-2160; settleCombo 1199-1205]) и в VerifyExpiry [VERIFIED: actor.go:1081]. У combo флип — первичное намерение жеста (§4.1 «Исправить слово и переключить раскладку»), спеке не противоречит.
- Наблюдаемый «mode to ru» через 56 мкс после done означает ЛИБО жест был комбо (флип по дизайну), ЛИБО коррекция исполнилась с изменениями (settleCorrectionFlip → scriptOf(converted); DEBUG-деталь невидим). Обе ветки спеко-совместимы; загадка растворяется наблюдаемостью, не правкой поведения.

**Рекомендация:** INFO-запись исхода получает дискриминатор `changed:true/false` (без содержимого — D-20/D-21); e2e-греп-формы проверить на совместимость (стенд грепает «correction»/«outcome done»).

### Anti-Patterns to Avoid

- **Ждать апп-верификацию дольше в wezterm** (поднять verify_wait_ms): контекст класса verify-dead не ответит никогда — растёт только цена отказа.
- **Слепой чинящий verify-after:** после разрушающего действия никакой достройки/допечатки — лог и стоп (критерий 4 дословно; терминал = команды).
- **Клипборд-ранг в терминальном пути** даже при `clipboard_rung: true`: Ctrl+V-форвард в шелле — SIGINT/вставка-мусор класс; ранг жёстко выключен в терминальной ветке.
- **Чинить пунктуацию внутри ConvertRuns:** конвертер прав; правка геометрии пробега в шве буфера — иначе смешанный диапазон и D-13 поедут.
- **Ставить вето TERM-04 в blocklist:** прямо запрещено инвариантом (blocklist не может исключить невидимое; wezterm невидим) [VERIFIED: docs/SPEC.md:578-586].
- **Ключеться на caps терминалов:** VTE ложно анонсировал surrounding (0x29, журнал спайка 18:19:29 [CITED: docs/adr/adr-008-spike-log.md:139]), ночник wezterm СМЕНИЛ подпись с 0x9 на бит-установлен — caps записываются, решения на них не ключуются (Pitfall 3 фазы 9).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Инжекция набора/жеста в живые окна | свой uinput-инжектор | ydotool 0.1.8 (физический путь, стенд-проверен) | 0.1.8 бездемонный, /dev/uinput + input-группа готовы |
| Чтение PRIMARY для wezterm-лега | свой Wayland-клиент | `wl-paste --primary --no-newline` через internal/clipboard Runner-шов | Пакет уже владеет дисциплиной (stdin-only контент, timeout, empty-состояние); добавить флаг — не транспорт |
| Управление select-чордами из e2e | новые CLI-хаки | debug-gated `goswitchctl probe-forward` (фаза 9) | Форвард — сигнал только собственного соединения движка; keyspec-грамматика и fail-closed уже в коде [VERIFIED: internal/session/probe.go:14-60] |
| Verify-машина таймеров | новые таймеры/горутины | pending/pendingAfter + epoch-дисциплина (уже в акторе) | Pitfall 4 ADR-004: ожидание живёт в таймере, не в хендлере; эпохи гасят stale-дедлайны |
| Оракул результата в терминале | AT-SPI-чтение терминала (wezterm вне шины; VTE-свидетель недостоверен — журнал спайка 2026-10-11) | pty-захват: pane-программа `cat > file`, hexlify-сравнение | Байт-точная земля без a11y-зависимости (прецедент P4-оракула) |
| Замороженные ряды/журнал вердиктов | чат/планы как источник вердиктов | appendSpikeRow-паттерн (датированные строки дописывает кейс) | d01/adr-008-журнальный прецедент; ADR-008 Evidence заполняется только строками журнала |

**Key insight:** весь терминальный путь — ОРКЕСТРАЦИЯ поверх уже написанных механизмов (буфер, конвертер, level-2, verify-after, ProbeForward, pty-оракул). Нового транспортного кода нет; риск фазы — семантика (классификация, границы пробега, безопасность середины строки), а не механизмы.

## Runtime State Inventory

Не rename/миграция, но фаза мутирует живое состояние — инвентаризация для планировщика:

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Установленный демон | на хосте активен goswitchd **v1.1.1** (релиз) — БЕЗ арминг-вето и терминальной ветки; HEAD-код опережает установку | живые ряды гоняют ДЕМОН СТЕНДА из рабочего дерева (стенд уже поднимает свой бинарник; правило «unit-демон остановлен перед живым прогоном»); owner-зачёт — после установки сборки фазы (существующий процесс обновления) |
| Конфиг владельца | `~/.config/goswitch/config.yaml` (если существует) — строгий декодер: незнакомые ключи отвергнут документ ЦЕЛИКОМ | новая схема — обратно совместима (zero-value = старое поведение, прецедент D-54); docs/CONFIG.md синхронно |
| Конфиг wezterm | `~/.wezterm.lua` — enable_wayland=true строка 78 (флип владелец-подтверждён); CopyMode-рецепт для primary-лега потребует правки | только через чекпойнт владельца с backup (прецедент 09-03); фикстура стенда несёт СВОЙ конфиг (`--config-file` top-level форма: `wezterm --config-file … start --always-new-process -- PROG` — закреплено журналом спайка A3 [CITED: docs/adr/adr-008-spike-log.md:25]) |
| Клипборд владельца | CLIPBOARD/PRIMARY живого стола | сентинелы только синтетические (GS-SPIKE-SENTINEL класс); кейс восстанавливает пред-стейт или документирует затирание (D5) |
| VM-стенд | `goswitch-stand` (libvirt) — state running [VERIFIED: `virsh list --all`, эта сессия]; внутри требуется: goswitchd+регистрация движка, wezterm-nightly .deb, gnome-terminal, ydotool+input-группа, wl-clipboard | установка/проверка стенда — задача плана 10-B/10-C (или отдельная Wave 0); см. Environment Availability |
| apt-состояние хоста | wezterm-nightly on hold (hi-флаг dpkg) | не трогать; откаты задокументированы в debug-файле |

## Common Pitfalls

### Pitfall 1: Одиночный зелёный wezterm-оракул
**What:** раунд прошёл раз — объявили победу; утром у владельца не работает.
**Why:** арминг-гонка IME wezterm (один прогон доставил 7/7 клавиш, другой 0 — живая находка 2026-10-02 в кейсе terminal-silent) + смена движков при флипе реминтит контексты.
**Avoid:** каждый wezterm-ряд = два подряд полных раунда в одном запуске кейса; арминг-гейт (dummy-клавиша в key-trace до раунда, ограниченные ретраи); SPEC §7.2 rev D-55 закрепил дисциплину [VERIFIED: docs/SPEC.md:418-423].
**Знаки:** план с приёмкой «e2e зелёный» без ×2.

### Pitfall 2: Середина строки при буфер-доверии
**What:** реплей стирает не то: токен слился со «старыми» рунами буфера (курсор ушёл стрелками/кликом/Tab-_completion — демону стрелки ВИДНЫ (не-printable, в буфер не едут [VERIFIED: actor.go:2190-2193]), клик и Ctrl+U/Ctrl+W — нет), либо буферистая-сталь после Ctrl+U воспроизводит «старые» руны конвертированными.
**Why:** буфер зеркалирует ПОСЛЕДОВАТЕЛЬНОСТЬ НАБОРА, не поле; без surrounding-verify расхождение невидимо. GTK-путь это чинит обязательной сверкой (ADR-004); у verify-dead контекста сверки нет.
**Avoid:** три механизма на выбор планировщика/владельца (OQ 2): (а) ядовитый флаг — стрелки/Control-комбо/непечатаемые навигационные ядовят буфер до следующего reset (Enter) → отказ «abort, не мусорить» (но это режет середину строки после стрелок — против критерия 1, если владелец так работает); (б) primary-read verify для wezterm (select-chord → сравнение с токеном — re-arm доверия); (в) для VTE — существующая surrounding-верификация (работает сегодня). Честно задокументировать границу контракта (D7: шеллы bash/zsh — в контракте).
**Знаки:** e2e-кейс середины строки отсутствует при приёмке «середина работает».

### Pitfall 3: verify-after с клипборд-рангом в терминале
**What:** mismatch после терминальной коррекции армит ранг → форвард Ctrl+V в шелл.
**Why:** общий settleAfter возвращает paste-полезную нагрузку при ClipboardRung [VERIFIED: actor.go:1546-1548]; флаг в дефолтном конфиге false, но «false по конфигу» ≠ «запрещено по построению».
**Avoid:** терминальная ветка verify-after без ранга вовсе (отдельный метод/флаг раунда); юнит-тест пинит «mismatch в терминале не порождает Ctrl+V ни при каком конфиге».

### Pitfall 4: Слепое доверие старой capability-строке wezterm
**What:** план опирается на «wezterm caps=0x9 → level-2 без верификации уже работает» (запись фазы 9).
**Why:** 0x9 — подпись СТАРОГО XIM-клиента 20240203; ночник через mutter несёт бит surrounding (доказано verify-timeout) — старый «бесплатный» level-2 путь больше не берётся.
**Avoid:** живой re-pin caps в первом же кейсе 10-C (строка журнала); решения — по текущей подписи.

### Pitfall 5: Строгий декодер против нового ключа конфига
**What:** добавили ключ в код, забыли docs/CONFIG.md/валидацию — документ владельца с новым ключом отвергается (или наоборот: старый демон + новый документ).
**Why:** KnownFields-декодер и D-32 контракты (валидация всего документа, last-good).
**Avoid:** схема+валидация+docs/CONFIG.md+golden одним PR; zero-value = выключено/старое поведение.

### Pitfall 6: Расширение пробега расширяет поверхность отказа
**What:** «ytn é» (é вне таблицы латиницы) — после фикса весь пробег включая пунктуацию отвергается; владелец замечает «раньше слово чинилось».
**Why:** T-03-01-03 — частичный вывод невозможен; пробег стал шире.
**Avoid:** golden-кейс на отказ расширенного пробега; граница пробега останавливается на знаке БЕЗ табличной позиции (пробел и пр.), но НЕ на знаке с позицией — поведение задокументировать в CONFIG.md/README.

### Pitfall 7: Свидетельские механизмы e2e на этой машине
**What:** кейс ждёт AT-SPI-свидетеля фокуса — вечное ожидание.
**Why:** на ядре хоста 7.2.9-2-t2 AT-SPI-репортинг фокуса недостоверен глобально, фоновые окна клавиатурно мертвы (журнал спайка 2026-10-11) [CITED: docs/adr/adr-008-spike-log.md:223-229].
**Avoid:** живые ряды — в VM-стенде (стоковое ядро, тестируемо) с маркерами арминг-гейта по key-trace демона (не по a11y); на хосте — только owner-запуск/owner-клик по runbook.

## Code Examples

### Текущая level-2 запись (точка расширения чанкированием)
```go
// Source: internal/session/actor.go:2881-2900 (executeLevel2, HEAD)
eng := a.emitter()
plan := correct.BuildPlan(rng.token, rng.tail, converted, a.caps, a.backspaceCap())
if plan.Level == correct.LevelNone {
    a.skipCorrection("backspace-cap")
    a.settleCombo()
    return
}
for range plan.Backspaces {
    eng.ForwardKeyEvent(engine.KeyBackSpace, backSpaceKeycode, 0)
}
eng.CommitText(engine.NewIBusText(string(plan.Commit)))
```

### Точка маршрутизации verify (куда встаёт классификация)
```go
// Source: internal/session/actor.go:2779-2810 (startRangeCorrection, HEAD)
if a.caps&correct.CapSurroundingText == 0 {
    // Сегодня: level-2 сразу, без верификации.
    a.executeLevel2(rng, converted, armed)
    return
}
// Сегодня: pending + RequireSurroundingText + бюджет verifyWait.
// Фаза: классифицировать контекст ДО этой развилки:
//   терминал с пушами (VTE)      → существующий verify-путь (или level-2, OQ 1);
//   verify-dead (wezterm-nightly) → терминальная ветка: буфер-таргет,
//   (опц.) primary-read verify OFF-mutex → чанкированный level-2.
a.pending = &pendingFix{…}
```

### Verify-after без ранга клипборда (образец дельты)
```go
// Существующая машина: internal/session/actor.go:1578-1597 (armAfterVerify)
// + settleAfter:1548 — mismatch отдаёт round.paste при ClipboardRung.
// Терминальный вариант: та же эпоха/дебаунс/тихий таймаут,
// но mismatch → только slog.Info("correction verify", "outcome", "mismatch")
// и СТОП; поле paste не заполняется, ранг не проверяется вовсе.
```

### Конвертируемый пробег (контрактный пример)
```go
// «ytn?» (EN-набор): сегодня token="ytn", tail="?" → «нет?».
// Целевое: convertible-stretch = "ytn?" (все руны с табличной позицией,
// mono-Latin) → ConvertRuns → «нет,». Пробел останавливает пробег
// («ghbdtn » → «привет », D-13 не меняется).
// Смешанный: «gfb?» — микс-диапазон, '?' едет как набрана (spec-delta
// 2026-10-06) — расширение пробега НЕ меняет микс-семантику.
```

### Мыло-лега primary-чтения (wezterm, опционально по OQ 2/3)
```bash
# Сентинел-дисциплина (прецедент фазы 9):
printf 'GS-SPIKE-SENTINEL' | wl-copy --primary      # известный пред-стейт
# select-chord (daemon ProbeForward или ядовитый флаг — см. OQ 2/3)
sleep 0.2                                            # subprocess-нога 150-400 мс, OFF-mutex
wl-paste --primary --no-newline                      # ожидаемый префикс токена, НЕ сентинел
```

### Замороженный ряд CLIPBOARD (критерий 3)
```bash
printf 'CLIP-SENTINEL' | wl-copy                    # пред-стейт CLIPBOARD
# …терминальная коррекция слова…
test "$(wl-paste --no-newline)" = "CLIP-SENTINEL"   # байт-нетронут — ряд матрицы
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| wezterm 20240203 XIM-клиент (caps 0x9, desync engine→stale-context, воспроизведён дважды) | wezterm-nightly 20261010, нативный Wayland, text-input v3 → mutter → IBus; commit_text видимо садится | 2026-10-12 (установлен, hold) | Дефект флипа TERM-01 снят сменой сборки; для фазы 10 сменилась caps-подпись — старые строки capability-таблицы wezterm исторические |
| Апп-верификация как безусловная ступень (ADR-004) | Двухклассовая модель: пушащие клиенты (GTK/VTE) — апп-верификация; verify-dead (wezterm) — терминальный канал | эта фаза | SPEC §4.3 rev D-55 уже закрепил право ADR-008 определять терминальную семантику [VERIFIED: docs/SPEC.md:152-160] |
| Пунктуация едет «как набрана» в слове (де-факто) | Конвертируемый пробег: пунктуация монодиапазона вместе со словом (§4.2) | эта фаза (дефект 2026-10-12) | Касается ВСЕХ приложений, не только терминалов |
| Вето автокоррекции на confirm-стадии (ADR-007-эпоха) | Вето в точке армирования до детектора (инвариант D-55) | эта фаза (критерий 5) | Закрытый slug воздержания; замороженные негативные ряды |
| Живые e2e только на хосте (owner-present) | VM-стенд goswitch-stand (libvirt, стоковое ядро) как машинная площадка живых рядов | 2026-10-12 (владелец одобрил, стенд running) | Owner-free регрессия терминальных рядов становится достижимой; хост — owner-зачёт |

**Deprecated/outdated:** строки capability-таблицы wezterm из журнала спайка (старый XIM-клиент) — исторические; вердикт «шорткат VTE не работает» (pty_hex пуст, flip_pair=false, 18:19:29) — получен в период клавиатурно-мёртвых фоновых окон (средовой дефект), НЕ транспортный факт; lore «gsettings current read-only» и xdotool/X11-инструменты — вне Wayland-мира проекта.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Контекст wezterm-nightly несёт бит surrounding-caps (вывод из verify-timeout + кода pending-условия) | Архитектурная находка A | Если бит когда-то снимается (mutter-апдейт), ветка маршрутизации сместится к level-2-без-верификации — классификация должна быть живой, не зашитой |
| A2 | Чанкированный ForwardKeyEvent-бурст доставляется шеллом как серия BackSpace без потерь (интервалы не нужны) | Pattern 2 | Если бурст теряет клавиши (перегрузка очереди), чанки обретут паузы — первом же живом ряде pty-оракул это покажет; план включает юнит на форму бурста + живой ряд |
| A3 | VTE-пуши surrounding text достаточны для MatchesSuffix-верификации в терминальной ветке (как в GTK-пути) | Pattern 1 таблица | Если VTE-пуши обрезаны/запаздывают, VTE-ветка падает на verify-timeout — тогда OQ 1 решается в пользу level-2-единообразия; живой ряд 10-B проверит |
| A4 | `wl-paste --primary` возвращает выделение, сделанное форварднутым select-чордом в wezterm-nightly (#4615-ловушка не перенеслась на ночник) | Pattern 2 / OQ 3 | Если чорд теряет модификаторы — лег не работает; фолбэк: буфер-доверие с ядовитым флагом (OQ 2а); ProbeForward-проба — первый таск 10-C |
| A5 | Поведенческий класс verify-dead стабилен в пределах жизни контекста (FocusOut/Reset снимает) | Pattern 1 | Мигающий класс даст переменную маршрутизацию; дешёвая страховка — лог смен класса (без содержимого) |
| A6 | Мид-лайн кейс владельца = «слово набрано там, где курсор» (не «после стрелок с кликом») | Pitfall 2 / OQ 2 | Если владельцу нужен вариант «стрелки → набор → коррекция», без primary-read (wezterm) он небезопасен — чекпойнт сформулирует границу контракта |
| A7 | Установка сборки фазы на хост/VM владельцем не ломает ежедневный флип (установленный v1.1.1 → HEAD-поведение) | Runtime State | Регрессия флипа откатывается установкой релиза; owner-зачёт гейтит |

## Open Questions (чекпойнты владельца/планировщика)

1. **VTE: единый level-2 или сохранение level-1?**
   - Знаем: TERM-02 и черновик ADR-008 D3/D4 предписывают буфер+BackSpace-реплей; ЖИВАЯ запись 2026-10-09 21:22:39 показывает работающую level-1 коррекцию (слово И фраза) в собственном терминале владельца через surrounding-верификацию [CITED: docs/adr/adr-008-spike-log.md:24]. ADR-008 всё ещё Proposed (09-03 не исполнен).
   - Неясно: жертвовать ли работающим level-1 VTE-путём ради единообразия (буква критерия 3), или амендментом ADR-008 закрепить двухклассовую запись (VTE=level-1 с апп-верификацией, verify-dead=терминальная ветка).
   - Recommendation: чекпойнт владельца в начале планирования; на уровне безопасности вариант (б) строго сильнее (верифицирует реальное поле — закрывает Pitfall 2 для VTE), вариант (а) сильнее по букве требования и единообразию тест-матрицы.
2. **Середина строки: механизм защиты буфер-доверия.**
   - Знаем: стрелки демону видны (ядовитый флаг возможен), клик/Ctrl+U — нет; критерий 1 требует середину строки в wezterm (bash/zsh).
   - Неясно: какой флоу владельца каноничен («набор сразу в середину» безопасен буфер-доверием при граничном символе-разделителе; «стрелки/клик → набор» — нет) и что считать контрактом (D7).
   - Recommendation: спросить владельца о реальном флоу; wezterm — primary-read лег (OQ 3), VTE — апп-пуши (OQ 1б) или отказ как граница контракта.
3. **Primary-read лег wezterm: в объёме фазы или контингенция?**
   - Знаем: критерий 2 говорит «при необходимости»; #4615-ловушка belonged к старой сборке; ProbeForward и wl-paste --primary готовы; рецепт CopyMode требует правки wezterm-конфига (фикстура стенда — свободно; ~/.wezterm.lua — чекпойнт).
   - Recommendation: проба чорда+чтения — первый live-таск 10-C; в объём — только если проба зелёная и OQ 2 требует верификации.
4. **Имя/место конфиг-ключей терминального класса** (напр. `correction.terminal.apps` vs секция `terminals`) и состав (identity-паттерны? бюджеты чанков?) — дискреция планировщика в рамках прецедентов схемы (zero-value=off, KnownFields, docs/CONFIG.md). identity-паттерны матчатся против ЧЕГО для wezterm — открытый подвопрос (у невидимого приложения identity нет; ключ практически обслуживает a11y-видимые терминалы и будущие классы).
5. **Судьба ProbeForward после фазы** — уже поставленный фазой 9 owner-гейт (09-03); фазе 10 он нужен для select-chord проб — подтвердить, что гейт остаётся открытым до конца фазы 10.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| wezterm (nightly, native Wayland) | целевой терминал №1 | ✓ живым | 20261010-083911-e2a98e1b (hold) | откат 20240203+X11 бессмыслен для фазы (старый дефект) |
| gnome-terminal / VTE | целевой терминал №2 | ✓ живым | 3.52.0 / VTE 0.76.0 | — |
| goswitchd (user unit) | весь продукт | ✓ active (v1.1.1 установлен; стенд гоняет HEAD-сборку) | см. Runtime State | — |
| ydotool + /dev/uinput + input-группа | инжекция набора/жеста | ✓ (v1.0/фаза-9 верификация) | 0.1.8 | ~150-LOC uinput-инжектор (паркован) |
| wl-clipboard | сентинелы CLIPBOARD/PRIMARY | ✓ живым | 2.2.1 | — |
| go / mise / golangci-lint | зелёная итерация | ✓ живым | 1.27.1 / 2026.10.7 / 2.13.2 | — |
| VM-стенд goswitch-stand (libvirt) | owner-free живые ряды | ✓ state: running [VERIFIED: virsh list, эта сессия] | Ubuntu 24.04 GNOME (стоковое ядро) | хост + owner-runbook (остаётся для owner-зачёта) |
| ВМ-внутренности (goswitchd+движки, wezterm-nightly, ydotool, wl-clipboard) | живые ряды в VM | ✗ не проверено (стенд строится параллельно) | — | установка — Wave-0 задача; фоново-клавиатурная гипотеза ВМ требует первой же пробы |
| AT-SPI/a11y | роль 60 (VTE) | ✓ (gnome-terminal на шине) | 2.52.0 | для wezterm источника НЕТ — поведенческий класс (Pattern 1) |
| ядро хоста 7.2.9-2-t2 | — (дефект атрибуции фокуса) | ✗ известный дефект фоновых окон | 7.2.9-2-t2-noble | VM-стенд обходит; хост — только owner-запуск окон |

**Missing dependencies with no fallback:** нет блокирующих; неопределённость одна — клавиатурная интерактивность фоновых окон ВНУТРИ VM («to be proven», первая проба стенда).

**Missing dependencies with fallback:** VM-внутренности (установка — задача), primary-read лег (фолбэк буфер-доверие+ядовитый флаг), AT-SPI-свидетели (фолбэк key-trace арминг-гейт).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `-race`; живой e2e-стенд `test/e2e` (Go-раннер, `PASS/FAIL`, exit≠0) — не `go test` |
| Config file | `mise.toml` ([tasks]; `ci` = build+vet+lint+test), строгий `.golangci.yml` v2 |
| Quick run command | `mise exec -- go test ./internal/session/ ./internal/correct/ ./internal/clipboard/ -race -count=1` |
| Full suite command | `mise run ci` (headless); живые ряды — новые mise-задачи семьи `e2e-term-*` (NOT in ci) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TERM-02 | «ytn?»→«нет,» — пунктуация монопробега (все приложения) | unit/golden (TDD) | `mise exec -- go test ./internal/correct/ -race -count=1` | ❌ Wave 0 (10-A) |
| TERM-02 | D-13/микс-инварианты не сломаны расширением пробега | unit/golden | та же команда (корпус расширяется, старые золотые зелёные) | ✅ (существующий корпус) |
| TERM-02 | чанкированный бурст: форма ForwardKeyEvent-серии ≤8, перечек фокуса, один commit | unit (фейковый эмиттер) | `mise exec -- go test ./internal/session/ -race -count=1` | ❌ Wave 0 (10-A) |
| TERM-02 | verify-after терминальной ветки: mismatch → лог+стоп, ранг клипборда не армится НИКОГДА | unit | та же команда | ❌ Wave 0 (10-A) |
| TERM-04 | вето в точке армирования (класс-терминал → воздержание до detect.Check) | unit | та же команда | ❌ Wave 0 (10-A) |
| TERM-02 | слово в VTE: ghbdtn→привет (pty-оракул), bash и zsh | live e2e | `mise run e2e-term-word-vte` (новый) | ❌ Wave 0 (10-B) |
| TERM-02 | CLIPBOARD байт-нетронут (замороженный ряд, сентинел) | live e2e | внутри `e2e-term-word-vte` / `e2e-term-word-wezterm` | ❌ Wave 0 (10-B/10-C) |
| TERM-02 (кр.5) | негатив: автокоррекция молчит (vim-insert/htop-свидетели) при зелёной ручной | live e2e | `mise run e2e-term-negative-witnesses` (новый) + re-pin `mise run e2e-autocorrect-terminal-silent` | ❌ / ✅ (re-pin существует [VERIFIED: mise-задача e2e-… terminal-silent семья]) |
| TERM-02 | слово в wezterm ×2 подряд (арминг-гонка), caps re-pin строкой журнала | live e2e | `mise run e2e-term-word-wezterm` (новый; два раунда внутри) | ❌ Wave 0 (10-C) |
| — | D-24-наблюдаемость: исход done несёт changed-дискриминатор | unit + grep-формы стенда | `mise exec -- go test ./internal/session/ -race -count=1` | ❌ Wave 0 (10-A) |

### Sampling Rate
- **Per task commit:** `mise run ci` (директива 2).
- **Per wave merge:** `mise run ci` + живые ряды волны на площадке (VM-стенд при доказанной клавиатурности; иначе owner-runbook хоста).
- **Phase gate:** `e2e-term-word-vte` + `e2e-term-word-wezterm` (×2) + `e2e-autocorrect-terminal-silent` зелёные; CLIPBOARD-ряд зелёный; ADR-008 Accepted с Evidence-строками; owner UAT.

### Wave 0 Gaps
- [ ] `internal/correct` — convertible-stretch геометрия + golden-корпус (TERM-02 пунктуация)
- [ ] `internal/session` — классификационный шов + арминг-вето + чанкированный level-2 + verify-after-без-ранга + D-24-дискриминатор (unit-корпус)
- [ ] `internal/clipboard` — `--primary` флаг чтения (если OQ 3 в объёме)
- [ ] `test/e2e/case_term_word.go` (или семья) — VTE/wezterm/negative-кейсы + pty/hex/сентинел-хелперы
- [ ] mise-задачи `e2e-term-*` (NOT in ci)
- [ ] docs/CONFIG.md + схема/валидация новых ключей
- [ ] Wave-0 проба VM-стенда: клавиатурность фоновых окон (гейт выбора площадки живых рядов)

## Security Domain

`security_enforcement: true`, ASVS level 1. Поверхность фазы:

| ASVS Category | Applies | Standard Control |
|---------------|---------|------------------|
| V2 Authentication | no | n/a |
| V3 Session Management | no | n/a |
| V4 Access Control | yes | живые пробы — только на поверхностях стенда/VM; окна владельца — только по runbook (правило с фазы 9) |
| V5 Input Validation | yes | новый конфиг — строгий декодер (KnownFields) + валидация потолков (чанки/бюджеты — положительные целые); keyspec ProbeForward уже strict [VERIFIED: internal/session/probe.go:14-23]; identity-паттерны — RE2 с явной анкеровкой (прецедент D-53) |
| V6 Cryptography | no | n/a |
| V14 Config | yes | zero-value = старое поведение; docs/CONFIG.md синхронно; no argv-интерполяции (wl-paste — флаги из литералов, контент только stdin/pipe — дисциплина internal/clipboard) |
| V16 Data Logging | yes | слово/пунктуация НИКОГДА не в логах (D-20/D21; logCorrectionDone пишет содержимое только на DEBUG [VERIFIED: actor.go:2907-2916]); -debug-трейс = клавиши видимы («passwords become visible») — синтетический корпус (ghbdtn) в живых рядах; pty-захваты tmp-only, hexlify, не коммитятся |

### Known Threat Patterns for this phase

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Слепой BackSpace-реплей по рассинхронизированному буферу (клик/Ctrl+U невидимы) → порча командной строки | Tampering (user data) | точка аборта до первого BackSpace; verify-after лог+стоп без самолечения; границы контракта D7; OQ 2-механизмы (ядовитый флаг/primary-read) |
| Терминальный verify-after армит paste-ранг → Ctrl+V в шелл (SIGINT/вставка) | Tampering | ранг запрещён в терминальной ветке по построению (юнит-пин), не конфигом |
| Арминг-вето обойдено при невидимой identity (wezterm) | Elevation (автокоррекция там, где запрещена) | вето в точке армирования по классу контекста, fail-closed; замороженные негативные ряды |
| Сентинелы/pty-артефакты уносят пользовательский текст в коммиты | Information Disclosure | синтетический корпус; tmp-only; hex-форма в журналах (журнальная дисциплина спайка) |
| Новый конфиг-ключ открывает DoS демона (нулевые/отрицательные бюджеты чанков) | Denial of Service | валидация диапазонов при загрузке (прецедент [1,500]/[2,16] коридоров схемы) |

## Sources

### Primary (HIGH — прочитано/проверено в этой сессии, 2026-10-12)
- `internal/session/actor.go` — startCorrection 2623-2634; startRangeCorrection 2745-2811 (caps-развилка 2779-2787); executeCorrection 2822-2842; executeLevel2 2881-2900; VerifyExpiry 1069-1082; ExpiryAt 1032-1061; HandleLifecycle 770-824; HandleSurroundingText-комментарий WR-01 826-845; settleAfter/WR-05 1519-1549; armAfterVerify/Range 1578-1617; settleCombo 1199-1205; settleCorrectionFlip 1207-1235; feedKey (combo 2149-2160, buffer-ветки 2166-2218, software keymap 2194-2210); pushAndArm 2226-2231; armAutoCorrect 2243-2252; autoCorrectBoundary 2272-2329; autoConfirm+ролевой switch 2447-2540; acConfirmRefusals 2542-2578; структуры correctionRange/acPayload/selectionSpec/pendingFix/pendingAfter ~379-483; константы ролей 82-86; backSpaceKeycode 53
- `internal/correct/` — buffer.go (Token/Tail 86-109, letterCapable/tokenCapable 186-205, recompute 165-183, Phrase/Replace*); runs.go (ConvertRuns 78-107, invertPerChar 109-140); plan.go (BuildPlan 57-85, D-27 cap 26-30); verify.go (MatchesSuffix 9-16, VerifyRangeAt 41-52); direction/convert (детектор, Convert)
- `internal/engine/` — engine.go:70-76 (Emitter), keys.go (маски, caps-биты, KeyBackSpace=0xff08, навигационные keyvals), types.go
- `internal/clipboard/clipboard.go` — wl-клиент, флаги/timeout/stdin-дисциплина
- `internal/ctlsvc/ctlsvc.go:81-152`, `cmd/goswitchctl/main.go:185-191`, `internal/session/probe.go` — ProbeForward debug-gate и keyspec-грамматика
- `internal/config/config.go:79-210` — секции схемы (Correction/Autocorrect/…)
- `internal/layouts/tables.go` — ENToRU['?']=',' (строка 44), '/'→'.' (23)
- `docs/SPEC.md` — §4.2 (99-126), §4.3 rev D-55 (146-160), §4.4 rev D-55 (177-188), §5 rev D-55 (363-372), §7.2 rev D-55 (413-423), §11 инвариант D-55 (559-586)
- `docs/adr/ADR-008-terminal-transport.md` — D1-D7 (86-92), Consequences (94-120); `ADR-003` (лестница); `ADR-004` (+amendment G-2-3)
- `docs/adr/adr-008-spike-log.md` — средовой блокиратор (11-31), VTE-ряды (133-142), HALT-запись (223-229), ночной вердикт wezterm (231-251), живая запись level-1 в терминале владельца (24)
- `.planning/debug/2026-10-12-wezterm-nightly-probe.md` — протокольный след text-input v3 (69-88), корень verify-timeout (90-118), дефект пунктуации (120-143)
- `.planning/phases/09-…/09-RESEARCH.md`, `09-VALIDATION.md` — прецеденты формата и стенд-дисциплины
- `test/e2e/case_autocorrect.go:751-810` (startWezterm/terminal-silent), `case_spike_caps.go` (VTE-раунд/pty-захват), `case_word.go`, mise.toml (задачи ci/e2e)
- Живые пробы этой сессии: `wezterm --version`, `gnome-terminal --version`, `wl-paste --version`, `go/mise/golangci-lint --version`, `uname -r`, `virsh list --all` (goswitch-stand running), `systemctl --user is-active goswitchd`, `dpkg -l wezterm-nightly` (hold)

### Secondary (MEDIUM — коммиченные исследования фаз 9/вехи)
- `.planning/research/{STACK,FEATURES,ARCHITECTURE,PITFALLS}.md` — апстрим-ловушки wezterm (#4615 модификаторы, #5125/#1772), RHBZ 1847347 (VTE surrounding), VTE-отсутствие keyboard-selection (source-verified), три семантики стирания

### Tertiary (LOW)
- Веб-lore терминального транспорта — гейтится живыми пробами фазы (правило ADR-008); в решениях не участвует

## Metadata

**Confidence breakdown:**
- Карта кода/механизмы: **HIGH** — каждый шов прочитан в этой сессии с номерами строк
- Терминальный конвейер (design): **MEDIUM-HIGH** — опирается на работающие механизмы; развилки OQ 1/2 требуют владельца
- Дефект пунктуации: **HIGH** — механизм подтверждён кодом + живым репортом владельца + SPEC-текстом
- Живые ряды/стенд: **MEDIUM** — VM-стенд running, но клавиатурность фоновых окон внутри ВМ не доказана; хостовый дефект ядра известен и обходится

**Research date:** 2026-10-12
**Valid until:** ~2026-11-11 (стек локально закреплён; риск дрейфа — mutter/wezterm-nightly обновления при снятом hold и статус VM-стенда)
