# Фаза 12: TERM-строка — Research

**Researched:** 2026-10-31
**Domain:** терминальный путь СТРОКИ (wezterm-only) поверх проверенной машины записи Фазы 10: канал чтения CopyMode-рецепт → PRIMARY, verify-then-act по объёму выделение↔буфер, отключение тройного Shift в VTE (D-02)
**Confidence:** MEDIUM — машина записи и жестовый каркас HIGH (живые ряды Фазы 10); канал ЧТЕНИЯ строки — наименее доказанное звено всего дизайна (ни одного зелёного ряда CopyMode-рецепта не существует; leg-probe красный) — см. Summary и Open Question 1

## Summary

Строка = «та же машина, другой чорд и объём выделения» — формула CONTEXT.md точна для ЗАПИСИ: write-путь строки обязан остаться байт-тем же чанкированным level-2 (пары press+release BackSpace, pacing 20 мс, drain-settle 150 мс, отложенный флип, verify-after без самолечения), зелёные ряды которого заморожены в `adr-008-spike-log.md` на обоих терминалах. Новое в фазе — три вещи: (1) КАНАЛ ЧТЕНИЯ — тройной Shift должен получить объём всей строки, а он у демона есть только через выделение (CopyMode-рецепт в конфиге wezterm → PRIMARY → `wl-paste --primary -n`; примитив `internal/clipboard.Primary` уже написан и ждёт первого потребителя); (2) ДОГОВОР ОБЪЁМА — verify-then-act «выделение ↔ буфер» (SC3): хвост выделения сверяется с буфером нажатий демона, расхождение = отказ, не угадывание; именно этот якорь честно решает THE жёсткий вопрос строки — глифы промпта в выделении никогда не стираются и не инвертируются, потому что N BackSpace-реплея считается по буферу, а не по выделению; (3) ЖЕСТОВЫЙ СЛОВАРЬ (D-02) — тройной Shift в VTE отключается (сегодня он де-факто работает: единый маршруто ОО1 «unified» Фазы 10 гонит ЛЮБОЙ терминальный диапазон, включая фразовый, в level-2), в wezterm становится строкой, в обычных приложениях остаётся фразой из буфера.

**Центральная находка исследования:** канал чтения — единственное звено строки, у которого НЕТ ни одного зелёного живого ряда. VTE select-chord красный (kill-критерий, ряд 18:19:27); leg-probe — форварднутый через демона чорд `Home, shift+End` НЕ поднимает выделение в PRIMARY (пять красных рядов 20:34:47/21:17:23/21:31:40/21:46:33/09:48:46/18:16:26); механика пробы CopyMode-рецепта в коде есть (`recipeChordRow`, физический `ydotool shift+home` на wayland-фикстуре с рецептом), но все прогоны `spike-wezterm` прервались — зелёного ряда рецепта в журнале спайка не существует. Открытым остаётся и вопрос ТРИГГЕРА выделения: кто нажимает рецепт-чорд в продакшн-жесте — демон (форварднутый `shift+home`, что leg-probe-класс не доказывает) или пользователь (двухдвижковый флоу). По дисциплине проекта (probe-first, прецедент leg-probe 10-06) это значит: первый содержательный план фазы — сел-проба с датированными рядами, и только её вердикт открывает задачу машины строки; планирование обязано заложить ветку «лег не подключён» → эскалация владельцу (Rule 4), а не тихий фолбэк.

**Primary recommendation:** три волны — (A) сел-проба канала чтения (физический чорд-рецепт → PRIMARY на wayland- и xim-фикстурах; форварднутый рецепт-чорд через probe-forward; датированные ряды в журнале спайка) + юнит-каркас жестового словаря D-02; (B) машина строки (классификация-шов, verify-then-act по буферному якорю, End-нормализация с задокументированным побочным эффектом, level-2 байт-прежний) + живые ряды; (C) отключение VTE-тройного + re-pin замороженных негативов + установка/UAT. Ноль новых зависимостей; go.mod заморожен.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions (Implementation Decisions)

**Сужение области (вердикт владельца)**
- **D-01:** Строка — **только wezterm**. Kill-критерий VTE сработал и принят: клавиатурного выделения в VTE нет (живая проба select-chord 18:19:27 в `docs/adr/adr-008-spike-log.md` — сентинел не сменился; апстрим-просьба с 2002). SC2 роадмапа закрывается сужением, вердикт — владельца, дан 2026-10-31. Альтернативы отклонены: «снять фазу» — не потребовалось; «VTE-примитив» (реконструкция строки из буфера нажатий демона) — **отвергнут**, ADR-008 D3 («единый примитив захвата строки — не существует, вопрос закрыт») остаётся закрытым.

**Жест тройного Shift в VTE**
- **D-02:** В VTE тройной правый Shift **отключён** — не делает ничего (ни строка, ни фраза из буфера). Терминальный жестовый словарь VTE: двойной Shift = слово (Фаза 10), тройной — молчание. В wezterm тройной = строка.

### Claude's Discretion
- Новых делегированных зон нет — «как» полностью закреплено ADR-008 (Accepted) и verdict-таблицей Фазы 9: End-нормализация и её побочный эффект задокументированы (SC1), verify-then-act при расхождении длин — отказ, не угадывание (SC3), клип-политика PRIMARY-only (D5), точка аборта до первого разрушающего действия (D6), граница интерактивные шеллы bash/zsh внутри / канонический режим и TUI вне (D7).

### Locked scope per CONTEXT.md (domain)
- wezterm-only; VTE: тройной Shift отключён (ничего не делает); обычные приложения не затронуты (тройной = фраза из буфера, как было).
- Чтение — клавиатурное выделение по CopyMode-рецепту → PRIMARY; запись — проверенная машина Фазы 10 (verify-then-act → N×BackSpace-реплей + commit).

### Deferred Ideas (OUT OF SCOPE)
- VTE-примитив реконструкции строки — не отложен, а **отвергнут** (D-01): reopening ADR-008 D3 не планируется; если апстрим VTE когда-нибудь даст клавиатурное выделение — вопрос переоткроется сам собой.

### Project Constraints (from AGENTS.md — bind all phases, no CLAUDE.md in repo)
1. **Строгий TDD** — каждая задача на поведение red → green → refactor; verify-блоки прогоняют тесты на каждом шаге.
2. **Зелёная итерация** — итерация закрывается только при `go build ./...`, `go test -race ./...`, `golangci-lint run` без ошибок.
3. **Проверочные сценарии — mise, не make** — задачи в `mise.toml` `[tasks]`, инструменты в `[tools]`.
4. **golangci-lint — максимально строгий конфиг** (v2, `linters.default: all`, depguard deny `unsafe`, formatters gofmt+goimports, `local-prefixes: github.com/Djarvur/goswitch`).
5. **GitHub Actions актуальны** — PR sanity (build+vet+lint+test -race+tidy-diff+govulncheck) на каждый push/PR; фазовых изменений CI не требуется.
- Процесс: изменения через PR в защищённый main, conventional commits; D-55-класс дисциплина «спека/ADR прежде кода» — переоткрытие решений ADR-008 D1–D7 = датированный амендмент, не тихая правка кода.
- Стек: Go ≥ 1.23, только stdlib + godbus/dbus + минимальные зависимости; uinput-инъекция ВНУТРИ демона запрещена (аксиома §3.1); `-debug`-журнал демона раскрывает нажатия («passwords become visible») — дисциплина D-20/D-21.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| TERM-03 | Пользователь может исправить всю набранную строку/фразу в терминале (растяжка). Feasibility-gated: VTE без клавиатурного выделения — сужение до wezterm, вердикт владельца (дан 2026-10-31, D-01) | Write-машина Фазы 10 байт-переиспользуется (executeLevel2/replayBurst пары/drain-settle/deferred flip — VERIFIED строки actor.go ниже); канал чтения = CopyMode-рецепт → PRIMARY (примитив Primary готов, потребителя нет; probe-first задача обязательна — leg-probe red); договор объёма SC3 = буферный якорь выделения; D-02 = словарь жестов VTE молчание/wezterm строка/обычные фраза |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Различение жеста (тройной: VTE-молчание / wezterm-строка / обычное-фраза) | internal/session (шов classifyTerminal + identity-якорь) | — | Один шов терминального класса уже кормит вето D-55 и конвейер записи; VTE-идентичность видима (якорь «gnome-terminal» матчится), wezterm-идентичность невидима (off a11y bus) — различение отрицательное, fail-closed |
| Триггер выделения строки (End + shift+Home / рецепт-чорд) | internal/session → internal/engine (Emitter.ForwardKeyEvent) | конфиг wezterm (CopyMode-рецепт) | Единственный клиентский рычаг демона — форвард клавиш; рецепт в конфиге wezterm превращает shift+Home в модальное выделение; подключённость форварднутого чорда НЕ доказана (leg-probe red) — probe-first |
| Чтение PRIMARY | internal/clipboard (Primary — subprocess wl-paste --primary -n) | — | Первый реальный потребитель примитива; вызов OFF-mutex по контракту, бюджет 1500 мс |
| Verify-then-act (выделение ↔ буфер) | internal/session + internal/correct (Buffer) | — | Буфер демона — единственное знание «что набрано»; хвост выделения обязан совпасть с буфером, иначе отказ (SC3); решает проблему глифов промпта |
| Конвертация строки | internal/correct (ConvertRuns/Convert — без изменений) | — | Homogeneous wholesale / per-char inversion уже реализованы; строки НЕ вносятся в конвертер |
| Геометрия замены (N×BackSpace + commit) | internal/session (executeLevel2, BuildPlan) — БАЙТ-НЕТРОНУТ | — | Всё доказано живыми рядами; строка отличается только диапазоном (больше слова) и входом |
| Отложенный флип / delivery-метки / WR-01 | internal/session — БАЙТ-НЕТРОНУТЫ | — | Машины 10-08/10-17 покрывают строку как есть (терминальная ветка) |
| CopyMode-рецепт в конфиге wezterm | Пользовательский конфиг (~/.wezterm.lua — только через чекпойнт с backup) | стенд-фикстуры (--config-file) | Рецепт — client-side; фикстуры стенда несут свои файлы (Pitfall 7: конфиг владельца не трогается экспериментами) |
| Живые ряды/негативы | test/e2e (семейство case_term*) + mise (NOT in ci) | журнал adr-008-spike-log.md | Прецедент 10-05/10-06: appendSpikeRow, арминг по key-trace, pty-оракул, two-greens |

## Standard Stack

### Core (всё in-repo — ноль новых модулей)

| Компонент | Версия | Роль в фазе | Основание |
|-----------|--------|-------------|-----------|
| internal/session (actor.go) | in-repo, 4612 строк | Жестовый словарь (Triple), шов classifyTerminal, новый вход строки, level-2 байт-прежний | [VERIFIED: internal/session/actor.go:1630-1637 — `for _, action := range a.fsm.Feed(hotkey.TimerExpired{}, now) { … case hotkey.Triple: a.startPhraseCorrection()`; :3964-3982 — `terminal := a.classifyTerminal()` → `if a.caps&correct.CapSurroundingText == 0 \|\| terminal {` → `a.executeLevel2(rng, converted, armed, terminal)`] |
| internal/correct | in-repo | BuildPlan (cap), ConvertRuns, Buffer.Phrase/ReplacePhrase — без правок | [VERIFIED: internal/correct/plan.go:30 — `const DefaultBackspaceCap = 50`; internal/correct/runs.go:78-107 — ConvertRuns wholesale/per-char] |
| internal/clipboard | in-repo | `Primary()` — чтение PRIMARY, первый потребитель | [VERIFIED: internal/clipboard/clipboard.go:99-108 — «Primary reads the PRIMARY selection byte-exactly (wl-paste --primary --no-newline) — the read half of the wezterm CopyMode leg … The call is OFF-mutex by contract»; grep потребителей: только тесты] |
| internal/engine (keys.go) | in-repo | KeyHome 0xff50 / KeyEnd 0xff57 / KeyLeft 0xff51 / KeyRight 0xff53, MaskShift, pairs-дисциплина форвардов | [VERIFIED: internal/engine/keys.go:59-62 — `KeyHome = 0xff50`, `KeyEnd = 0xff57`, `KeyLeft = 0xff51`, `KeyRight = 0xff53`] |
| internal/hotkey (fsm.go) | in-repo | Тройной уже различается машиной: MaxTaps=3, Triple action | [VERIFIED: internal/hotkey/fsm.go:16 — `KeyvalShiftR = 0xffe2`; :20 — `MaxTaps = 3`; :30 — `Triple`] |
| test/e2e + probe.go | in-repo | Кейс семейства term-line; probe-forward верб с keyspec `[mod+]{Home\|End\|Left\|Right\|BackSpace}` — keyspec shift+home УЖЕ валиден | [VERIFIED: internal/session/probe.go:25 — «keyspec must be [mod+]{Home\|End\|Left\|Right\|BackSpace} or one key»; test/e2e/case_spike_wezterm.go:777-796 — recipeChordRow (ydotool shift+home → wl-paste --primary)] |
| Стенд goswitch-stand (VM) | kvm, 192.168.122.150 | Площадка живых рядов (venue-verdict vm-primary 2026-10-12) | [CITED: docs/adr/adr-008-spike-log.md — блок «Wave-0 площадка Фазы 10», ряд venue-verdict 14:20] |
| Стенд-фикстуры wezterm | 3 файла in-repo | xim (транспорт владельца), wayland (РЕЦЕПТ CopyMode уже внутри), stand (ноябрьский дефолт) | [VERIFIED: test/e2e/fixtures/gs-e2e-wezterm-wayland.lua:10-18 — Recipe Multiple{ActivateCopyMode, SetSelectionMode Cell, MoveToStartOfLineContent, Close}; gs-e2e-wezterm-xim.lua и gs-e2e-wezterm-stand.lua — рецепта НЕТ, прочитаны] |

### Supporting (инструменты стенда/хоста — живые проверки этой сессии)

| Инструмент | Версия (живая проверка 2026-10-31) | Когда используется |
|------------|-----------------------------------|---------------------|
| wezterm-nightly | 20261010-083911-e2a98e1b (held; `wezterm --version`) | Фикстура и целевой терминал |
| gnome-terminal (VTE) | 3.52.0-1ubuntu2 | Негативные ряды D-02 (тройной молчит), двойной-слово не сломать |
| ydotool | 0.1.8-3build1 | Физический путь: набор, чорды выделения (модификаторы ДОХОДЯТ физическим путём — select-chord пробы так и устроены) |
| wl-clipboard | 2.2.1-1build1 | wl-paste --primary --no-newline чтение; сентинелы |
| mise | tasks в mise.toml (61 задача; golangci-lint = "2" в [tools]) | ci/e2e-задачи фазы |

**Installation:** ничего. `go.mod`/`go.sum` заморожены (нулевые новые модули; прецедент всех фаз вехи).

## Package Legitimacy Audit

**Неприменимо — фаза не ставит ни одного внешнего пакета.** Весь код — in-repo (internal/*, test/e2e); инструменты стенда — системные пакеты Ubuntu 24.04, проверенные живыми dpkg-пробами выше.

**Packages removed due to [SLOP] verdict:** none.
**Packages flagged as suspicious [SUS]:** none.

## Architecture Patterns

### System Architecture Diagram — путь строки (три исхода жеста)

```
                    ТРОЙНОЙ правый Shift (FSM Triple, 300 мс окно)
                                   │
                     startPhraseCorrection (actor.go:1635-1636)
                                   │
                        classifyTerminal() [единый шов]
                ┌───────────────────┼─────────────────────────┐
        терминал? НЕТ            терминал? ДА                терминал? ДА +
        (обычное приложение     + VTE-идентичность           identity неизвестна
         — байт-как сегодня)    (якорь «gnome-terminal»      (wezterm off-a11y-bus,
                │                матчится)                    поведенческий класс)
                │                       │                          │
             ФРАЗА                 D-02: МОЛЧАНИЕ               СТРОКА (новый вход)
         из буфера демона        — return без единого        1. форвард End (+пара)
         (Phrase→startRange       emitted байта, без флипа   2. форвард shift+Home —
         Correction→caps/        (двойной Shift = слово       CopyMode-рецепт →
         level-1/level-2)        в VTE НЕ ТРОНУТ)             выделение активное → PRIMARY
                │                                           3. wl-paste --primary -n
                │                                           4. VERIFY-THEN-ACT: хвост
                │                                              выделения == буфер демона?
                │                                              НЕТ → ОТКАЗ (не угадывание)
                │                                              ДА ↓ (глифы промпта вне цели)
                │                                           5. executeLevel2 (БАЙТ-ПРЕЖНИЙ):
                │                                              N=len(буфер) пар BackSpace
                │                                              чанки ≤8, pacing 20мс,
                │                                              drain-settle 150мс → commit
                │                                              инвертированного буфера
                │                                           6. отложенный флип + verify-after
                │                                              (noRung) + delivery-метки
                ▼                                                  ▼
          [курсор: позиция = конец коммита]        [побочный эффект End-нормализации
                                                    задокументирован; отказ = НОЛЬ
                                                    разрушений до первого форварда]
```

Точки входа: жест тройного Shift (FSM), debug-верб probe-forward (селен-пробы), e2e-кейс семейства term-line. Внешние зависимости: конфиг wezterm (рецепт), PRIMARY-селекция клиента, wl-clipboard. Решения: ветвление по classifyTerminal + identity; fail-closed отказ при любом расхождении.

### Pattern 1: Probe-first канал чтения (гейт фазы)

**What:** до всякой задачи машины строки — датированные ряды сел-пробы в журнале спайка, по форме leg-probe 10-06: сентинел в PRIMARY → чорд → settle 400 мс (chordSettle) → `wl-paste --primary -n`. Три ячейки: (a) ФИЗИЧЕСКИЙ чорд `ydotool shift+home` на wayland-фикстуре (рецепт уже в файле) — доказывает сам рецепт; (b) ФИЗИЧЕСКИЙ чорд на xim-фикстуре + рецепт в неё добавленный — доказывает рецепт на транспорте владельца (ведь XIM с 2026-10-24) и мост X11-PRIMARY → wl-paste --primary; (c) ФОРВАРДНУТЫЙ чорд `goswitchctl probe-forward shift+home` (keyspec валиден [VERIFIED: internal/session/probe.go:25]) — доказывает/опровергает самотриггер выделения демоном.
**When:** первый план фазы, до машины.
**Why:** leg-probe red доказал только отсутствие выделения от ПЛОСКИХ форварднутых Home/shift+End (которые без рецепта уходят в pty как навигация readline); связка «форварднутый чорд + рецепт в конфиге» не пробована НИ РАЗУ — это ровно тот класс «фикс из предположения», который проект гонит датированными рядами.
**Verdict-развилка:** (c) зелёный → жест «один тройной Shift исправляет строку» достижим; (c) красный → демоy-самотриггер мёртв: эскалация владельцу (Rule 4-класс) с вариантами «двухдвижковый флоу: пользователь жмёт рецепт-чорд (Shift+Home), тройной исправляет выделенное» / «снять строку» — фолбэк «буфер вместо выделения» запрещён D-01 (это и есть отвергнутый VTE-примитив по духу, и SC3 теряет свой якорь).

### Pattern 2: Буферный якорь verify-then-act — ответ на глифы промпта (SC3)

**What:** выделение (строка с промптом, например `nil@host:~$ ghbdtn`) сверяется с буфером демона: буфер обязан быть хвостом содержимого выделения. Цель коррекции — ТОЛЬКО буферная часть: N BackSpace = len(буфер) (не len(выделения)!), commit = инвертированный буфер. Глифы промпта никогда не стираются (BackSpace из конца строки не проходит за начало строки шелла — readline/zle не отдают промпт) и никогда не инвертируются (в ConvertRuns не попадают).
**Why это решает THE жёсткий вопрос:** wholesale-конвертация строка-целиком превратила бы латинские буквы промпта (`nil@host`) в кириллицу («нил@хост») — катастрофа; якорь делает промпт геометрически вне цели без единого эвристического «детектора промпта» (его и не существует честного).
**Refusal-ветки (все fail-closed, до первого разрушающего действия — D6):** хвост выделения ≠ буфер (recall истории Up, Ctrl+A-редактирование, Ctrl+U, чужой буфер PRIMARY, RPROMPT в хвосте) → отказ; выделение пусто/не сменилось (лег мёртв) → отказ; буфер пуст → отказ; len(буфера) > backspace_cap → отказ backspace-cap (D-27).
**Граница честности (документируется рядом D7-класса):** строка корректируется только для флоу «набрал строку целиком в текущем контексте, курсор где угодно»; recall истории — вне (буфер не видел текст). Это СУЖЕНИЕ по сравнению с наивным «исправь что в строке» — оно честное: демон знает только то, что пропустил через себя.

### Pattern 3: End-нормализация и её побочный эффект (SC1)

**What:** жест строки форвардит End ДО выделения — курсор уходит в конец строки, shift+Home поднимает выделение «вся строка», хвост-якорь срабатывает, реплей стирает ровно буферную часть от конца строки, commit ставит инвертированный текст. Финальная позиция курсора = конец (исправленного) текста.
**Побочный эффект (задокументировать, не устранять — формулировка SC1):** исходная mid-line позиция курсора демону НЕИЗВЕСТИМА (терминал не сообщает позицию; не-печатные клавиши буфер не кормят [VERIFIED: internal/session/actor.go:3240-3243 — `case !printableKeyval(ev.Keyval) || ev.Mods&comboMask != 0: … return false`]). Поэтому в договорном флоу «набор в конец строки» (доминирующий) курсор после коррекции стоит ровно там же — «курсор возвращается в исходную позицию» выполнен тривиально; в mid-line флоу курсор остаётся на конце строки, суффикс не теряется (он был внутри выделения, инвертировался по-символьно и закоммитился) — смещение фиксируется в docs/CONFIG.md рядом с терминальной клаузой. Альтернатива с восстановлением (двойное чтение: shift+Home до End → длина суффикса = разность двух чтений → Left× суффикс после коммита) возможна, но удваивает канал чтения и форвард-трафик — закладывать только по явному требованию владельца (Open Question 2).
**Дисциплина пар:** End/Left/стрелки — повторяемые клавиши; КАЖДЫЙ форвард едет парой press→gap→release (амендмент (7) ADR-008: press-only взводит клиентский автоповтор — «появилось и исчезло»). shift+Home — чорд: press-пары обоих членов, порядок релизов — как в forwardCtrlV/probe-дисциплине.

### Pattern 4: Словарь жестов D-02 — различение VTE/wezterm на одном шве

**What:** startPhraseCorrection получает ту же классификацию, что и запись: терминал + identity-якорь матчит VTE-паттерн (`gnome-terminal`) → жест молча завершается (ни строки, ни фразы, ни флипа; наблюдаемость — решение Open Question 3, рекомендация: D-20-класс refusal-запись `triple-disabled-vte`, не тишина — иначе владелец не отличит «отключено» от «сломалось»); терминал + identity неизвестна → строка (wezterm-класс); не-терминал → фраза из буфера байт-как сегодня.
**Почему различение отрицательное и почему это ок:** wezterm отсутствует на a11y-шине («wezterm is off the a11y bus» — VERIFIED прецедент 09-RESEARCH/case_autocorrect.go), положительной идентичности у демона нет и не появится; VTE — GTK-гражданин, якорь `gnome-terminal` матчится (row «identity-якорь … классификация VTE-контуров стала identity-истинной» [CITED: adr-008-spike-log.md:330]). Degradation-класс: VTE с невидимой identity (a11y сломан, класс-черк host ac_role_errors) уедет в строку-ветку → сел-лег не поднимет выделение (VTE) → отказ fail-closed, ноль разрушений — приемлемая деградация, задокументировать. Третьи терминалы (kitty/…): как wezterm — попытка строки с честным отказом, если лег не подключён (вне capability-таблицы Фазы 9 — гарантий нет).

### Pattern 5: Живые ряды семейства term-line (форма 10-05/10-06)

**What:** кейс `case_term_line_wezterm.go` + mise-задача (NOT in ci), секция журнала «Фаза 12 живые ряды wezterm (term-line-wezterm)»: арминг по key-trace (x+BackSpace, ретрай-лестница 3), наборы корпусом (echo-префикс + ghbdtn — промпт-класс случай у фикстуры `bash -c 'cat > file'` — промпта нет, ВСЯ строка набрана; синтетический промпт эмулируется префиксом набора, например `echo `), два подряд зелёных раунда (×2 дисциплина арминг-гонки), mid-line ряд, clip-frozen сентинел (CLIPBOARD байт-нетронут; PRIMARY затёрт выделением — задокументировать пред-стейт, D5), VTE-негатив (тройной в VTE: НОЛЬ emitted записей демона + двойной-слово остаётся зелёным), re-pin замороженных негативов автокоррекции (term-negative-witnesses, autocorrect-terminal-silent) и слова (term-word-vte/wezterm) — регресс-гейт SC3.
**Oracles:** pty-дельта байт-точна (форма drain-кейса: RAW захват stty, порядок 0x7f строго прежде байтов слова); delivery-метки chunks→commit→flip-armed→flip-released(причина); refusal-ряды отказов (буфер≠выделение — синтетический раскол: набор + Up-arrow recall) — «отказ, не угадывание» доказан живьём.

### Anti-Patterns to Avoid

- **Строить машину строки до сел-пробы:** leg-probe-класс красного ряда обязан существовать ДО кода записи строки — иначе вся волна B строится на неподключённом леге.
- **Угадывать при расхождении:** любой mismatch буфер↔выделение — отказ; «поправим как думаем» = инвариант TERM-04-класса нарушен в ручном канале.
- **Commit-over-selection / paste-чорды / DST:** отвергнуты по построению (ADR-008 D4; DST = compositor-killer, ретест-гейт закрыт до 51.1+main [CITED: adr-008-spike-log.md, секция round-4 retest-gate]).
- **Инвертировать промпт:** никакого «детектора промпта» и никакого wholesale по всему выделению — только буферный якорь (Pattern 2).
- **Тихие правки ADR-008 D1–D7:** новое решение = датированный амендмент на новых живых рядах (дисциплина Status Accepted).
- **Трогать ~/.wezterm.lua без чекпойнта:** recipe в ежедневный конфиг — только backup + явный owner-чекпойнт (прецедент 09-03), стенд живёт на --config-file фикстурах.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Чтение PRIMARY | свой Wayland/X11 клиент или парсинг селекций | `internal/clipboard.Primary` (готов, протестирован, флаги запинены, empty-состояние определено) | Уже несёт stdin/argv-безопасность, таймаут, Runner-шов [VERIFIED: internal/clipboard/clipboard.go:99-120] |
| Конвертация строки | новый «line converter» | `correct.ConvertRuns` (один вход для слова/фразы/выделения уже) | Per-char inversion/wholesale/рефузалы уже живые и пиннуты корпусом |
| Арифметика замены | собственный подсчёт BackSpace | `correct.BuildPlan` (runes, cap, tail-геометрия) | D-27 cap и rune-координаты уже верны; строка отличается только входом |
| Пейcинг/пары/чанки/драйн | новая «line-машина» | `executeLevel2` как есть (терминальная ветка) | Шесть раундов G-10-1 отполировали ровно эту машину; строка = тот же шов |
| Форвард чордов выделения | новый транспорт | `Emitter.ForwardKeyEvent` + парная дисциплина (probe.go:179-183 форма) | Единственный лег; keyspec-грамматика probe-верба уже валидирует shift+home |
| Журнальные ряды вердиктов | заметки в планах | `appendSpikeRow`-паттерн в кейсе | d01/adr-008 прецедент; ADR-амендменты цитируют только ряды |
| Фикстуры wezterm | правка конфига владельца | `--config-file` стенд-файлы (рецепт уже в wayland-фикстуре) | Pitfall 7; A3-форма спавна закреплена журналом |

**Key insight:** единственный по-настоящему новый код фазы — вход строки (классификация + сел-лег + буферный якорь + D-02-словарь). Конвертация, геометрия, доставка, флип, метки, негативы — уже написаны и доказаны рядами; их правка = регрессионный риск без причины.

## Runtime State Inventory

> Фаза не rename/refactor, но мутирует живое состояние — инвентаризация того, что переживёт коммиты.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — демона не хранит строки; буфер в памяти, PRIMARY/CLIPBOARD — внешний мир (D5: CLIPBOARD байт-нетронут; PRIMARY затирается выделением жеста — задокументировать в ряде/доках) | — |
| Live service config | `~/.wezterm.lua` владельца: `enable_wayland = false` [VERIFIED: живая проба 2026-10-31, строка 86] — РЕЦЕПТА CopyMode в нём нет; юнит goswitchd хоста на release v1.1.1-226-g8bd4edc (round-6 сборка) + пост-фазный 946b4cd ждёт установки | Рецепт в ежедневный конфиг — чекпойнт владельца + backup + rollback-строка (прецедент 09-03/10-06); установка фазовой сборки — ритуал release-dir (10-06 форма), путь отката назван ДО приёмки |
| OS-registered state | sources-takeover стенда [('ibus','goswitch-en'),('ibus','goswitch-ru')] / хоста — не меняется фазой; hold wezterm-nightly (apt-mark) — не менять | Assert-only |
| Secrets/env vars | `-debug`-журнал раскрывает нажатия (строки СЕКРЕТНЫ, если владелец так набрал) | Дисциплина D-20/D-21: ряды несут hex/счётчики, корпус синтетический; содержимое выделения в логи НЕ ПОПАДАЕТ ни на каком уровне |
| Build artifacts | release-dir'ы хоста (v1.1.1…226) — не перезаписывать; зеркало стенда ~/goswitch — rsync-рецепт 10-01 | Новые release-dir; зеркало re-sync перед живыми рядами |
| Coordination | 11-01 Task 1 (Orca Round 2) правит ТОТ ЖЕ actor.go — HandleLifecycle ~1282-1360; фаза 12 работает в диапазонах жестов/строки (~1630, ~3800-4400) | Integration-ordering (форма 11-01): Round 2 PR сливается ДО фазового PR; ветка фазы ребейзится на post-fix main; номера строк actor.go пере-пиняются по фактическому файлу |

## Common Pitfalls

### Pitfall 1: Триггер выделения принят работающим без пробы
**What goes wrong:** машина строки строится на «демон форвардит shift+Home и читает PRIMARY» — а форварднутый чорд не поднимает выделение (весь класс leg-probe red), и вся волна мертва на первом живом ряде.
**Why:** leg-probe проверял плоские Home/shift+End без рецепта; связка «рецепт в конфиге + форварднутый чорд» не пробована; #4615-класс (потеря модификаторов на форварднутых чордах) — открытый hazard.
**How to avoid:** Pattern 1 — сел-проба три ячейки (физика-wayland / физика-xim / форвард) с вердикт-развилкой до кода машины.
**Warning signs:** план, где первый контакт с PRIMARY — в задаче записи.

### Pitfall 2: Промпт инвертируется wholesale-конвертацией
**What goes wrong:** цель коррекции = всё выделение → `nil@host:~$ ghbdtn` конвертируется EN→RU целиком («нил@хост:~$ привет»), pty получает разрушенную команду.
**Why:** ConvertRuns честно конвертирует всё, что ей дали; латиница промпта для неё — конвертируемый текст.
**How to avoid:** буферный якорь (Pattern 2) — конвертируется ТОЛЬКО буферная часть; юнит-корпус с промпт-подобным префиксом пинует геометрию (N по буферу, commit по буферу, промпт вне).
**Warning signs:** любой план, где Backspaces считается от len(выделения).

### Pitfall 3: Расхождение буфер↔выделение «починено» угадыванием
**What goes wrong:** после Up-arrow recall (или Ctrl+A-редактирования) хвост не совпал — машина «помогает» тем, что в буфере, и стирает часть чужой строки.
**Why:** соблазн использовать буфер как истину вместо селекции.
**How to avoid:** SC3 дословно: расхождение → отказ (refusal D-20-класса), ноль форвардов. Живой отказ-ряд в матрице (синтетический recall-случай).
**Warning signs:** ветка «если хвост частично совпал — обрежем».

### Pitfall 4: Курсор-восстановление посчитано из того, чего демон не знает
**What goes wrong:** план обещает «курсор вернулся в исходную mid-line позицию» через Left×N — а N неоткуда (исходный оффсет курсора демону неизвестен; позицию терминал не сообщает).
**Why:** SC1 сформулирован по договорному флоу; mid-line — побочный эффект End-нормализации.
**How to avoid:** Pattern 3 — задокументировать эффект (курсор на конце строки в mid-line флоу; суффикс цел — инвертирован и закоммичен); вариант с двух-чтением — только по решению владельца.
**Warning signs:** «Left×len(suffix)» без источника len.

### Pitfall 5: Wrapped/многострочная строка считается в договоре
**What goes wrong:** длинная строка переносится на несколько ЭКРАННЫХ рядов — MoveToStartOfLineContent в CopyMode двигает по визуальному ряду, выделение захватывает только последний экран-ряд; якорь ломается или коррекция частична.
**Why:** CopyMode-селекция визуальная, не логическая [CITED: wezterm copymode.md — «Move to start of indented line = ^/Alt+M»; движения по строкам экрана].
**How to avoid:** граница договора: однострочные (невернутые) команды — в договоре; wrapped/PS2-продолжения — вне, отказ по якорю (частичное совпадение хвоста = расхождение). Зафиксировать в docs/CONFIG.md рядом с D7.
**Warning signs:** acceptance-ряд с 100+ символами в одном pty-ряде.

### Pitfall 6: Backspace cap молча режет длинные строки
**What goes wrong:** строка длиннее cap → LevelNone → silent backspace-cap отказ; владелец читает «не работает».
**Why:** DefaultBackspaceCap = 50 [VERIFIED: internal/correct/plan.go:30], конфиг `[1, 500]` [VERIFIED: internal/config/config.go:24-25 — `minBackspaceCap = 1`, `maxBackspaceCap = 500`].
**How to avoid:** поведение существующее — оставить; задокументировать границу в CONFIG.md + refusal-ряд в матрице (строка > cap → отказ байт-честно). Тайминг-примечание: 50 пар × 20 мс + drain 150 мс ≈ 1.2 с — лучшее время строки превышает §5-бюджет; терминальное завершение best-effort (прецедент формулировки 09-RESEARCH для спеки).
**Warning signs:** предложение поднять дефолт cap без owner-гейта.

### Pitfall 7: D-02 сломал слово в VTE или фразу в обычных приложениях
**What goes wrong:** отключая тройной в VTE, задевают двойной (слово) в VTE или фразовый путь GTK-приложений; замороженные ряды красные.
**Why:** один шов обслуживает три жеста; VTE-ветка фразы сегодня ЖИВАЯ (unified routing) — отключение это регрессия по решению, её надо изолировать ровно в Triple-входе терминала-VTE.
**How to avoid:** юнит-матрица словаря {VTE, wezterm, GTK} × {Single, Double, Triple}: Single флип везде; Double слово в терминалах/выделение-слово в GTK; Triple фраза в GTK / молчание VTE / строка wezterm. Re-pin: term-word-vte, term-word-wezterm, term-negative-witnesses, autocorrect-terminal-silent — все зелёные до закрытия.
**Warning signs:** правки в startCorrection/executeLevel2 вместо входа startPhraseCorrection.

### Pitfall 8: CLIPBOARD/PRIMARY-политика нарушена чтением
**What goes wrong:** чтение PRIMARY затирает… ничего — но выделение жеста ПЕРЕЗАПИСЫВАЕТ прежний PRIMARY пользователя (copy-on-select), и это надо документировать; CLIPBOARD трогать нельзя вовсе.
**Why:** D5: PRIMARY-only, CLIPBOARD байт-нетронут; судьба затёртого PRIMARY — задокументированная оговорка (прецедент clip-frozen рядов).
**How to avoid:** clip-frozen ряд строки (сентинел в CLIPBOARD проходит сквозь жест байт-нетронутым; PRIMARY пред-стейт — сентинел, затёртие задокументировано в Наблюдаемое).
**Warning signs:** любой wl-copy в терминальном пути.

### Pitfall 9: Субпроцесс чтения под мьютексом актора
**What goes wrong:** wl-paste (до 1.5 с бюджета) вызван под a.mu — клавиатурный путь демона стоит.
**Why:** Primary() по контракту OFF-mutex [VERIFIED: internal/clipboard/clipboard.go:104-107]; runClipboardRung — прецедент off-mutex ступени с WR-01-классом сериализацией.
**How to avoid:** сел-лег — off-mutex ступень (форма runClipboardRung): чтение вне мьютекса, результат — короткая секция под мьютексом для вердикта; emitter-вызовы wire-safe off-mutex (godbus Emit — прецедент того же комментария).
**Warning signs:** `a.Primary(ctx)` под удерживаемым a.mu.

### Pitfall 10: Один зелёный ряд = победа
**What goes wrong:** arming-гонка wezterm даёт одиночный ложнозелёный прогон (живая находка 10-06).
**Why:** uinput-чурн реминтит движок; первые клавиши теряются.
**How to avoid:** ×2 два подряд зелёных в одном запуске — для каждого вердиктного ряда строки; арминг-гейт с ретрай-лестницей (3).
**Warning signs:** кейс с одним раундом на исход.

### Pitfall 11: Транспорт-дрейф владельца (XIM с 2026-10-24)
**What goes wrong:** селен-пробы гоняются только на wayland-фикстуре ночника, а ежедневный конфиг владельца — XIM (`enable_wayland = false` [VERIFIED: ~/.wezterm.lua:86, живая проба 2026-10-31; причина в комментарии :80 — окно без заголовка]) — поведение селекции/форвардов на XIM не измерено, UAT владельца красный.
**Why:** Фаза 10 уже наступала на класс «проверено на одном транспорте, принято на другом» (10-VERIFICATION SC1 round-5, RESOLVED только 2026-10-31).
**How to avoid:** сел-проба обязана иметь xim-ячейку (рецепт добавляется в копию xim-фикстуры); UAT-зачёт — под текущим конфигом владельца; write-машина транспорт-агностична (доказано), но чорды/селекция — клиент-сайд.
**Warning signs:** план без xim-ряда.

### Pitfall 12: Конфликт правок с 11-01 (HandleLifecycle Round 2)
**What goes wrong:** параллельная работа в actor.go валит ребейз/пины.
**Why:** 11-01 Task 1 правит HandleLifecycle (Round 2 Orca fix) в том же файле.
**How to avoid:** integration-ordering: Round 2 в main ДО фазового PR; фазовая ветка ребейзится; строковые пины пересчитываются (форма уже предписана в 11-01-PLAN).
**Warning signs:** фазовый PR, открытый на pre-fix main.

## Code Examples

### Точка входа жеста — различение Triple (сегодня, байт-точно; фаза добавляет словарь внутри startPhraseCorrection)
```go
// Source: internal/session/actor.go:1628-1638 (прочитано 2026-10-31)
for _, action := range a.fsm.Feed(hotkey.TimerExpired{}, now) {
    slog.Info("action", "n", int(action))
    switch action {
    case hotkey.Double:
        a.startCorrection()
    case hotkey.Single:
        a.flipTo(oppositeMode(a.mode))
    case hotkey.Triple:
        a.startPhraseCorrection()
    }
}
```

### Существующий терминальный маршруто записи — строка едет ТЕМ ЖЕ швом
```go
// Source: internal/session/actor.go:3964-3981 (прочитано 2026-10-31)
terminal := a.classifyTerminal()
if a.caps&correct.CapSurroundingText == 0 || terminal {
    // …the whole terminal class — against a terminal the buffer is
    // the only target source (the stretch of plan 10-02)…
    a.executeLevel2(rng, converted, armed, terminal)
    return
}
```

### Дисциплина пар форвардов (амендмент (7)) — форма для End/Left/чорда выделения
```go
// Source: internal/session/actor.go:4292-4301 (прочитано 2026-10-31)
sleep(a.level2ForwardGap)
// … free-abort window (plan 10-17) …
eng.ForwardKeyEvent(engine.KeyBackSpace, backSpaceKeycode, 0)
sleep(a.level2ForwardGap)
eng.ForwardKeyEvent(engine.KeyBackSpace, backSpaceKeycode, engine.MaskRelease)
```

### Verify-after терминала — noRung (строке нужен ПРЕ-чек, а этот остаётся пост-гейтом)
```go
// Source: internal/session/actor.go:2569-2580 (прочитано 2026-10-31)
func (a *Actor) armAfterVerifyTerminal(converted, tail []rune) {
    a.verifyEpoch++
    a.clearAfter() // never two open rounds
    a.after = &pendingAfter{
        expected: concatRunes(converted, tail),
        epoch:    a.verifyEpoch,
        noRung:   true,
    }
    epoch := a.verifyEpoch
    a.after.deadline = time.AfterFunc(a.verifyWait, func() { a.afterExpiry(epoch) })
    a.emitter().RequireSurroundingText()
}
```

### Чтение PRIMARY — готовый примитив (OFF-mutex контракт)
```go
// Source: internal/clipboard/clipboard.go:99-108 (прочитано 2026-10-31)
// Primary reads the PRIMARY selection byte-exactly (wl-paste --primary
// --no-newline) — the read half of the wezterm CopyMode leg (ADR-008 D5:
// PRIMARY-only, CLIPBOARD never touched by the terminal path). …
// The call is OFF-mutex by contract — the caller reads it
// outside the actor's mutex, never under it (the runClipboardRung
// precedent, actor.go) — and bounded by the same cmdTimeout.
func (c *Clipboard) Primary(ctx context.Context) (saved []byte, had bool, err error)
```

### CopyMode-рецепт — уже в wayland-фикстуре; для xim-ячейки добавляется копией файла
```lua
-- Source: test/e2e/fixtures/gs-e2e-wezterm-wayland.lua:10-18 (прочитано 2026-10-31)
table.insert(keys, {
  key = 'Home', mods = 'SHIFT',
  action = wezterm.action.Multiple{
    wezterm.action.ActivateCopyMode,
    wezterm.action.CopyMode { SetSelectionMode = 'Cell' },
    wezterm.action.CopyMode 'MoveToStartOfLineContent',
    wezterm.action.CopyMode 'Close',
  },
})
-- Close LEAVES the selection active; copy_on_select (default) кладёт
-- выделение в PRIMARY — читается wl-paste --primary --no-newline
```

### Сел-проба форварднутого чорда — верб уже умеет keyspec
```bash
# Source: internal/session/probe.go:25 — keyspec grammar (прочитано 2026-10-31)
# "probe keyspec must be [mod+]{Home|End|Left|Right|BackSpace} or one key"
goswitchctl probe-forward shift+home   # ячейка (c) Pattern 1; fail-closed без -debug
```

### Cap-граница строки
```go
// Source: internal/correct/plan.go:26-30 (прочитано 2026-10-31)
// DefaultBackspaceCap is the default bound of one level-2 Backspace series
// (D-27: a configurable ~50-press ceiling; …) A range longer than the cap
// never replays — the correction takes the silent backspace-cap refusal
// instead of hammering the key bus.
const DefaultBackspaceCap = 50
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Тройной Shift в терминале = фраза из буфера (unified routing Фазы 10) | D-02: VTE — молчание; wezterm — строка; GTK — фраза | решение владельца 2026-10-31 | Отключение в VTE — осознанная регрессия, изолируется в Triple-входе |
| Уровень-1 (DST) как замена терминальной записи | Только парные форвард-бекспейсы (level-2) | бисекция 2026-10-16 + round-6 2026-10-19 | DST депрекирован навсегда на текущем стеке (ретест-гейт закрыт до 51.1+main) |
| Терминальная строка «если-примитив-найдётся» (VTE-надежда) | wezterm-only, D-01 | 2026-10-31 | SC2 роадмапа закрыт сужением; reopened ADR-008 D3 не планируется |
| Буфер-доверие как единственный механизм чтения wezterm | + PRIMARY-лег строки (если сел-проба зелёная) | эта фаза | Primary() получает первого потребителя; слово остаётся буферным |
| wezterm нативный Wayland (ноябрь 2026-10-12…24) | wezterm XIM/XWayland (откат владельца, причина — CSD-рамки) | 2026-10-24 | Все клиент-сайд пробы строки обязаны иметь xim-ячейку (Pitfall 11) |

**Deprecated/outdated:**
- Флэт «строка как большая фраза из буфера» — это отвергнутый D-01 примитив; буфер в строке живёт ТОЛЬКО как якорь верификации, не как источник объёма.
- `SelectTextAtMouseCursor 'CommandLine'`-класс семантических зон — требует OSC 133 shell-integration; в договоре не участвует (фикстуры без интеграции; рецепт Cell+MoveToStartOfLineContent не требует ничего) [CITED: wezterm SelectTextAtMouseCursor.md — SemanticZone since 20210203 «See Shell Integration docs»].

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Физический чорд Shift+Home с рецептом в конфиге wezterm поднимает выделение строки в PRIMARY (copy_on_select) — механика `recipeChordRow` существует, но зелёного ряда НЕТ (все spike-wezterm заезды aborted) | Pattern 1, Pitfall 1 | Весь канал чтения мёртв → строка откатывается к owner-развилке (двухдвижковый флоу/снятие) |
| A2 | Форварднутый демоном `probe-forward shift+home` триггерит рецепт в конфиге wezterm (форварднутые клавиши проходят client-side key-dispatch) | Pattern 1 | Красный ряд = самотриггер невозможен; жест «один тройной Shift исправляет» недостижим без user-чорда (owner-развилка) |
| A3 | Мост X11 PRIMARY ↔ Wayland primary-selection (XWayland-селекции читаются wl-paste --primary) работает на стеке владельца/стенда | Pattern 1 ячейка (b) | xim-ячейка сел-пробы красная → строка на текущем конфиге владельца невозможна, только под нативным Wayland |
| A4 | End/Left форварды повторяемы → парная дисциплина амендмента (7) распространяется на них; чорд shift+Home форвардится парой press/release обоих членов без потери модификатора | Pattern 3 | Одиночные форварды взводят автоповтор клиента («появилось и исчезло» — round-6 класс); потеря Shift = End вместо выделения |
| A5 | Хвост выделения bash/zsh совпадает с буфером в договорном флоу (промпт до начала строки контента; RPROMPT нет в хвосте) | Pattern 2 | Ложные отказы на реальных промптах владельца → юзабилити-провал (лечится refusal-рядами и/или owner-чекпойнтом по его PS1) |
| A6 | Время сел-лега (2 пары форвардов + settle + wl-paste) + реплей строки ≤ ~1.5-2 с — приемлемо как best-effort (SPEC §5 терминальная клауза) | Pitfall 6 | UX-жалоба; не блокер корректности |
| A7 | «Молчание» D-02 реализуемо как тихое завершение Triple-входа без побочных эффектов (FSM/флип/буфер) | Pattern 4 | Если молчание нарушит инварианты combo/флипа — нужен refusal-записной вариант (Open Question 3) |

## Open Questions

1. **Триггер выделения строки — гейт-вопрос фазы**
   - What we know: физический рецепт-чорд не доказан ни разу (A1); форварднутый — тем более (A2); leg-probe-класс красный на плоских чордах.
   - What's unclear: проходит ли форварднутая клавиша через client-side key-dispatch wezterm (рецепт сработает) или уходит в pty мимо конфига.
   - Recommendation: probe-first задача (Pattern 1) в первом плане; вердикт-развилка на владельца при красной ячейке (c) заранее вписана в план (Rule 4-форма).
2. **Курсор: документировать побочный эффект или строить two-read восстановление?**
   - What we know: исходный mid-line оффсет демону неизвестен; SC1-формулировка допускает «задокументированный» эффект.
   - Recommendation: дефолт — документировать (Pattern 3); two-read вариант держать как явную развилку чекпойнта владельца, если mid-line UX важен.
3. **Наблюдаемость D-02-молчания**: тихий no-op vs refusal-запись (`triple-disabled-vte`) в D-20-словаре.
   - Recommendation: запись — иначе диагностика «почему не работает» невозможна; не несёт содержимого.
4. **Ключ конфига для словаря строки**: reuse `correction.terminals` + неявное VTE-молчание vs отдельный якорь line-капабельных терминалов.
   - Recommendation: не добавлять ключ без необходимости (zero-config дефолт мы уже якорим [gnome-terminal, wezterm]); решение за планировщиком при Victoria-проверке формата конфига.
5. **Деплой рецепта в `~/.wezterm.lua` владельца**: чекпойнт с backup в плане C; конфликт с кастомными keys владельца (конфиг несёт свои `keys`) проверяется чтением конфига на чекпойнте.
6. **Приёмка/UAT**: owner-зачёт строки (selftype-класс G-10-1) под текущим XIM-транспортом; формулировка human-check в UAT заранее.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| wezterm-nightly (хост) | целевой терминал, UAT | ✓ | 20261010-083911-e2a98e1b (held) | — |
| wezterm на стенде ВМ | живые ряды | ✓ (10-06..10-18 ряды) | nightly, stand-фикстура | — |
| gnome-terminal/VTE (хост+стенд) | D-02-негативы, слово-регрессия | ✓ | 3.52.0 | — |
| ydotool + /dev/uinput + input-группа | физические чорды/набор | ✓ | 0.1.8 | собственный uinject (~150 LOC, в test/e2e уже есть) |
| wl-clipboard | PRIMARY чтение/сентинелы | ✓ | 2.2.1 | — |
| Стенд goswitch-stand (VM, ssh) | живые ряды vm-primary | ✓ (venue-verdict 2026-10-12) | стоковое ядро ВМ | owner-desktop (только UAT) |
| probe-forward верб | сел-проба ячейка (c) | ✓ in-repo (debug-gated, fail-closed) | — | — |
| mise + golangci-lint | зелёная итерация | ✓ ([tools] mise.toml) | golangci-lint = "2" | — |
| goswitchd -debug + ctlsvc | оракулы key-trace | ✓ | — | — |

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** none — вся площадка Фаз 9/10 жива и переиспользуется.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `-race` (testify запрещён конвенцией); живые e2e `test/e2e` — mise-задачи NOT in ci |
| Config file | `mise.toml` (ci/lint/test/e2e-*), строгий `.golangci.yml` v2 |
| Quick run command | `mise exec -- go test ./internal/... -race -count=1` |
| Full suite command | `mise run ci` (build + vet + golangci-lint v2 strict + test -race; прецедент ~10-15 с) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TERM-03 | Сел-проба канала чтения: физический рецепт-чорд → PRIMARY (wayland + xim ячейки); форварднутый чорд ячейка; датированные ряды | e2e (стенд) | `mise run e2e-term-line-probe` (новая) | ❌ Wave 0 |
| TERM-03 | Словарь жестов D-02: VTE Triple = 0 emitted; Double-слово жив; GTK Triple = фраза; wezterm Triple = строка-ветка | unit | `go test -race ./internal/session -run TestGestureDict_ -count=1` | ❌ Wave 0 |
| TERM-03 | Буферный якорь: промпт вне цели; N по буферу; mismatch→отказ (recall/Ctrl+U/пустой буфер/чужой PRIMARY) | unit | `go test -race ./internal/session ./internal/correct -run TestLineAnchor_ -count=1` | ❌ Wave 0 |
| TERM-03 | End-нормализация: пары форвардов End/Left; порядок select→read→burst→commit | unit | `go test -race ./internal/session -run TestLineDelivery_ -count=1` | ❌ Wave 0 |
| TERM-03 | Живые ряды wezterm: ×2 два зелёных, mid-line, clip-frozen, отказ-ряды, pty-дельта/метки оракулы | e2e (стенд) | `mise run e2e-term-line-wezterm` (новая) | ❌ Wave 0 |
| TERM-03 | VTE-негатив строки + слово-регрессия + замороженные негативы автокоррекции | e2e (стенд) | `mise run e2e-term-negative-witnesses` + re-pin `e2e-term-word-vte` / `e2e-term-word-wezterm` / `e2e-autocorrect-terminal-silent` | ✅ (переиспользуются) |
| TERM-03 | Установка + owner-зачёт строки (selftype, текущий XIM-транспорт) | manual | ритуал release-dir + UAT human-check | ❌ финальный план |

### Sampling Rate
- **Per task commit:** `mise exec -- go test ./internal/... -race -count=1`.
- **Per wave merge:** `mise run ci` + соответствующие живые ряды волны.
- **Phase gate:** все замороженные ряды (word-vte/wezterm, flip, negative-witnesses, autocorrect-terminal-silent) зелёные + новые line-ряды ×2 зелёные + `mise run ci` зелёный до `/gsd:verify-work`; owner UAT — финальный зачёт.

### Wave 0 Gaps
- [ ] `test/e2e/case_term_line_probe.go` (или расширение семейства) — сел-проба, TERM-03 Wave 0
- [ ] стенд-фикстура с рецептом на xim-транспорте (копия gs-e2e-wezterm-xim.lua + recipe) — TERM-03
- [ ] юнит-корпус: TestGestureDict_*, TestLineAnchor_*, TestLineDelivery_* (internal/session)
- [ ] mise-задачи e2e-term-line-* (NOT in ci)
- [ ] интеграционный порядок с 11-01 Round 2 (ребейз-дисциплина) — координационная, не тестовая, ячейка

## Security Domain

`security_enforcement: true` (config), ASVS Level 1. Поверхность фазы — расширение терминального пути демона и стенд-пробы.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | n/a (нет auth-поверхности) |
| V3 Session Management | no | n/a |
| V4 Access Control | yes | стенд-этикет (только свои фикстуры; хост после установки read-only); debug-вербы fail-closed без -debug (прецедент 10-13/10-16) |
| V5 Input Validation | yes | содержимое выделения = НЕдоверенный ввод: никогда не в логах (D-20/D-21), конвертация через strict-таблицы (unmapped → refusal всего диапазона [VERIFIED: internal/correct/runs.go:133-135]); keyspec-грамматика probe-верба |
| V6 Cryptography | no | n/a |
| V14 Config | yes | правка `~/.wezterm.lua` — чекпойнт+backup+rollback; `--config-file` фикстуры stand-owned; без shell-интерполяций |
| V16 Data Logging | yes | `-debug` раскрывает нажатия («passwords become visible» [CITED: cmd/goswitchd/main.go:38 прецедент]); строка может нести секреты (пользователь выделил строку с токеном) — содержимое выделения в журнал НЕ ПОПАДАЕТ ни на каком уровне; только длины/исходы/латентности |

### Known Threat Patterns for this phase

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Секрет в выделенной строке утекает через журнал/артефакты | Information Disclosure | D-20/D-21: hex/счётчики/булевы в рядах; pty-артефакты tmp-only; синтетический корпус |
| Инвертированный промпт = разрушенная команда в шелле | Tampering (integrity) | Буферный якорь: промпт вне цели; wholesale по всему выделению запрещён геометрией (юнит-пин) |
| Гонка: тройной Shift во время чужой коррекции/флипа | DoS (UX) / Elevation of state | WR-01 level2Busy покрывает всё окно машины (прецедент T-10-17-01); второй вход = correction-busy refusal; флип паркуется (parkGestureFlip) |
| PRIMARY-затирание пользовательского выделения жестом | Tampering (UX data) | D5-оговорка: задокументировано в ряде/доках; CLIPBOARD байт-нетронут (clip-frozen ряд) |
| DST-эмиссия (запрещённый примитив) роняет gnome-shell | Denial of Service | DST вне терминального пути по построению; ретест-гейт 10-16 закрыт (мьютер-краш, mutter fix отсутствует до 51.1+main) |

## Sources

### Primary (HIGH — прочитано/проверено в этой сессии)
- `internal/session/actor.go` — :1630-1637 (FSM Triple dispatch), :2569-2580 (armAfterVerifyTerminal noRung), :3179-3268 (feedKey; не-печатные не кормят буфер :3240-3243), :3407-3427 (classifyTerminal), :3893-3907 (startPhraseCorrection), :3928-4008 (startRangeCorrection; terminal-маршруто :3964-3982), :4107-4228 (executeLevel2), :4287-4316 (replayBurst пары)
- `internal/correct/plan.go:26-30, 57-85` — DefaultBackspaceCap=50, BuildPlan геометрия; `internal/correct/runs.go:78-140` — ConvertRuns/refusals; `internal/correct/buffer.go:127-148` — Phrase/ReplacePhrase
- `internal/clipboard/clipboard.go:99-120` — Primary() (OFF-mutex контракт, флаги, empty-состояние)
- `internal/engine/keys.go:59-62` — KeyHome/KeyEnd/KeyLeft/KeyRight keyvals; `internal/hotkey/fsm.go:13-31` — KeyvalShiftR/MaxTaps/Triple
- `internal/session/probe.go:25, 171` — keyspec-грамматика (shift+home валиден), ProbeForward
- `internal/config/config.go:24-25, 44-51` — cap границы [1,500], дефолт 50; identity-якорь `gnome-terminal`/`wezterm` дефолты
- `test/e2e/fixtures/gs-e2e-wezterm-wayland.lua` (рецепт), `gs-e2e-wezterm-xim.lua`, `gs-e2e-wezterm-stand.lua` (без рецепта) — прочитаны дословно
- `test/e2e/case_term_word_wezterm.go` (весь), `test/e2e/case_spike_wezterm.go:709-803` (recipeChordRow), `test/e2e/wezterm.go:1-60` (фикстурные константы)
- `docs/adr/ADR-008-terminal-transport.md` (весь — D1-D7 + 7 амендментов + Reversibility), `docs/adr/adr-008-spike-log.md` (весь — kill-ряд 18:19:27, leg-probe red ×6, venue vm-primary, round-3..6 секции)
- `.planning/phases/09-spayk-i-defekt-wezterm/09-RESEARCH.md` (весь — capability-таблица, CopyMode-рецепт-кандидат, SC-карта)
- `.planning/phases/10-term-slovo/10-VERIFICATION.md:195-218, 330-345` — откат владельца 2026-10-24, RESOLVED 2026-10-31, живые ряды машины сегодня
- `.planning/phases/11-menu-punkt-v-chernyy-spisok/11-01-PLAN.md` (HandleLifecycle Round 2 — координация actor.go)
- `.planning/phases/12-term-stroka/12-CONTEXT.md`, `12-DISCUSSION-LOG.md`, `.planning/REQUIREMENTS.md` (TERM-03), `.planning/ROADMAP.md:128-140`, `docs/SPEC.md:128-188` (§4.3 D-55 клауза, §4.4 транспорт)
- Живые пробы 2026-10-31: `~/.wezterm.lua:80,86` (enable_wayland=false), `wezterm --version`, dpkg (wezterm-nightly/gnome-terminal/wl-clipboard/ydotool), mise.toml

### Secondary (MEDIUM)
- wezterm официальные доки (raw.githubusercontent.com/wezterm/wezterm/main/docs/copymode.md, SelectTextAtMouseCursor.md, ActivateCopyMode.md — скачаны curl 2026-10-31): CopyMode-семантика, SetSelectionMode Cell/Word/Line/SemanticZone/Block, MoveToStartOfLineContent (визуальный ряд), SemanticZone требует shell-integration
- Community/web: мост X11-PRIMARY ↔ Wayland primary-selection в GNOME/mutter (wp_primary_selection_unstable_v1) — [LOW→MEDIUM, кэш research-store; живой ряд обязателен]
- ADR-008-нутренние CITED-строки журнала спайка (см. выше) — формально Secondary как записанные вердикты прежних сессий, но для проекта это канон

### Tertiary (LOW)
- WebSearch-выдача по sync-инструментам клипбордов (reddit/arch) — только контекст, вердиктов не несёт

## Metadata

**Confidence breakdown:**
- Машина записи строки (write): **HIGH** — байт-тот же шов executeLevel2 с шестью раундами живых рядов и реплей-корпусом; никаких правок не требуется
- Жестовый словарь D-02 / классификация: **HIGH** — все швы прочитаны, матрица {терминал×жест} выводима из кода; риск только в регрессии соседних веток (пины существуют)
- Канал чтения строки (read leg): **LOW→MEDIUM** — примитив Primary готов [VERIFIED], но вся цепочка «чорд→селекция→PRIMARY→чтение» не имеет ни одного зелёного живого ряда; фаза обязана открыться сел-пробой (Pattern 1) и держать owner-развилку при red
- Договор объёма/промпт-глифы: **MEDIUM** — якорная схема выводится из SC3 и существующей семантики, но промпт-формы владельца (PS1/RPROMPT) не измерены (A5)
- Площадка/инструменты: **HIGH** — всё живо, venue-verdict vm-primary стоит

**Research date:** 2026-10-31
**Valid until:** ~2026-11-30 (стабильно: стек запинен; риск дрейфа — владелец меняет wezterm-транспорт/конфиг — уже учтён как Pitfall 11/чекпойнт-форма)
