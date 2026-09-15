# Phase 3: Фразы, выделение, переключение и конфигурация - Research

**Researched:** 2026-09-15
**Domain:** IBus-engine daemon feature build-out: phrase/selection/mixed-text correction, layout-switch combos, YAML hot reload, goswitchctl control CLI, Super→Ctrl macro layer (GNOME Wayland, Go, IBus 1.5.29)
**Confidence:** HIGH (code claims verified against the working tree read this session; protocol claims verified against installed headers/man pages; ADR pack is binding)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Смешанный текст (CORR-06 — закрыт Q-маркер SPEC §4.2)**
- **D-22:** Якорь «правильной» раскладки — раскладка последней БУКВЫ корректируемого текста (буфера/фразы/выделения); конвертируются только прогоны «чужих» букв. Буквальное чтение SPEC §4.2 «относительно последней использованной». Цифры и общая пунктуация якорь не сдвигают (продолжение D-15): `ghbdtn2026привет` — якорь RU от `…вет`, конвертируется только `ghbdtn`.
- **D-23:** Гранулярность — прогонами: каждый непрерывный прогон букв одной раскладки конвертируется независимо, если он чужой относительно якоря. ЕДИНЫЙ код-путь для слова (двойной тап), фразы (тройной) и выделения: `ghbdtnпривет` → `приветпривет`; `gfbпривет` → `паипривет`.
- **D-24:** «Нечего конвертировать» (все буквы уже в якорной раскладке) — УСПЕШНАЯ операция без изменений, не отказ и не WARN. Владелец выбрал это вопреки рекомендации «тихий отказ». Пустой буфер/отсутствие текста остаются тихим отказом по D-20.

**Границы фразы (CORR-02)**
- **D-25:** «Вся фраза» = весь буфер с последнего жёсткого сброса (Enter, Tab, смена фокуса, Escape — SPEC §4.3), без искусственного капа на длину. Обязательная сверка с surrounding text (ADR-004) сама отрезает неваличный хвост — «abort, не мусорить».
- **D-26:** Фраза со словами в разных раскладках конвертируется той же семантикой прогонов (D-22/D-23): якорь = последняя буква фразы, чужие прогоны конвертируются, свои не трогаются. Отдельной «фразовой» семантики нет.
- **D-27:** Лестница ADR-003 для фраз: DeleteSurroundingText приоритетен; ступень Backspace×N получает настраиваемый кап серии (дефолт ~50 нажатий, YAML); если обе ступени недоступны и clipboard-ступень выключена — тихий отказ D-20 с причиной в лог.

**Выделение и clipboard (CORR-03)**
- **D-28:** Clipboard round-trip для замены выделения — opt-in по конфигу (последовательность с ADR-003 rung C «только по конфигу»); в дефолт-конфиге ВЫКЛЮЧЕН. Порядок ступеней: commit_text/DeleteSurroundingText → clipboard round-trip (если включён).
- **D-29:** После clipboard round-trip демон best-effort восстанавливает исходное содержимое буфера обмена (сохранить перед → вернуть после); неудача восстановления — только запись в лог, не ошибка операции.
- **D-30:** Приложение не сообщает выделение (anchor == cursor в SetSurroundingText или surrounding text недоступен) → тихий отказ D-20 с причиной в лог. Двойной тап БЕЗ активного выделения продолжает исправлять последнее слово — поведение Фазы 2 не меняется.

**YAML-конфиг и Caramba (CONF-01..03)**
- **D-31:** Своя чистая схема ключей (секции уровня `hotkeys` / `timeouts` / `correction` / `macr` — точную структуру выбирает планировщик); совместимость с Caramba (CONF-03, wish) — таблицей соответствия в документации, БЕЗ клонирования Caramba-имён: их схема привязана к их модели «свитч на первый тап», наша — к ADR-002.
- **D-32:** Hot reload при невалидной правке — last-good + WARN: демон продолжает работать на последней валидной конфигурации, WARN в лог, ошибка видна в `goswitchctl status`. Atomic-rename редакторов уже покрыт паттерном STACK (directory-watch + debounce, fsnotify).
- **D-33:** Строгий парсер: неизвестный ключ = невалидный конфиг (отклоняется целиком, last-good остаётся). yaml.v3 strict-режим (KnownFields) — опечатка в ключе не молчит.

**Переключение раскладки (SWCH-01..04) — примирение с ADR**
- **D-34 (перенос ADR-001/D-03, BIND):** Одиночный Right Shift переключает ВНУТРЕННИЙ режим EN↔RU за интерфейсом `layoutset`. «Один хозяин»: в источниках ввода только goswitch; Super+Space ничего не переключает (и не обязан — переключать нечего); системный индикатор GNOME всегда показывает один источник goswitch и внутренний режим НЕ отражает — индикатору не верить, компенсации нет (OSD отложена). Формулировки SWCH-03 («Super+Space и MRU продолжают работать») и критерия №3 ROADMAP УСТАРЕЛИ против принятых ADR-001/ADR-002 — верификатор фазы НЕ должен требовать их буквального исполнения; рабочая интерпретация: нативный механизм Super+Space не сломан присутствием goswitch (если пользователь добавит другие источники — механизм жив), e2e-оракул переключения — внутреннее состояние движка (решение FSM/`layoutset` в логе), НЕ gsettings.
- **D-35 (перенос ADR-002, BIND):** Конфликт таймингов single/double/triple уже разрешён ADR-002: классика с ожиданием, ВСЕ действия на истечении окна (дефолт 300 мс, настраивается — теперь реально через YAML), ничего на самом тапе; модификаторная дискриминация, правило четвёртого тапа, отмена посторонней клавишей — исполняются существующим корпусом `internal/hotkey/fsm_test.go`.
- **D-36 (SWCH-02, из SPEC §4.1):** Комбо Shift+RightCtrl (переназначается): ИСПРАВИТЬ последнее слово (семантика двойного тапа: сверка ADR-004, границы D-13..D-15), ЗАТЕМ переключить внутренний режим (toggle EN↔RU, тот же `layoutset`). Порядок зафиксирован именем действия в SPEC («исправить слово и переключить раскладку»). Q-маркер SPEC `[Q: конфликтует с прошлым опытом дублирования]` — дефолт остаётся Shift+RightCtrl (переназначение в YAML покрывает неудобство); живую эргономику проверяет e2e.

### Claude's Discretion
- Точная структура YAML-схемы (имена ключей внутри секций, формат биндингов) — планировщик; принципы D-31..D-33 фиксированы. В конфиг обязаны вместиться: окно тапов, все биндинги (вкл. комбо SWCH-02), кап Backspace (D-27), вкл/выкл clipboard-ступени (D-28), MACR: глобальный вкл/выкл + per-app списки + альтернативные модификаторы (ADR-005).
- Транспорт `goswitchctl` ↔ демон (кандидат по STACK: D-Bus на session bus — свой интерфейс демона; альтернатива — локальный сокет) — исследователь/планировщик.
- Формат вывода `goswitchctl status` (человекочитаемый; `--json` опционален) и состав: режим EN/RU, конфиг (путь, валидность, last-good ошибка), счётчики коррекций, перехваченные Super-комбинации (ADR-005).
- Механика восстановления буфера (D-29) и clipboard round-trip — исполнитель (wl-copy/wl-paste по STACK).
- Реализация MACR-01 по ADR-005: перехват Super+Буква в ProcessKeyEvent, эмуляция Ctrl+Буква через ForwardKeyEvent, AT-SPI-наблюдатель только при per-app списках; дефолт альтернативного модификатора НЕ вводится до живой проверки (ADR-005 b.3) — фаза проверяет и решает.
- Ширина e2e-матрицы v2 «разные приложения»: минимум zenity (GTK) + Chromium (уже в стенде); расширение (Electron и др.) — планировщик по результатам исследования поверхностей.
- Порядок задач, разбиение на планы, детали self-hosted раннера — исполнитель.

### Deferred Ideas (OUT OF SCOPE)
- Терминальный класс (clipboard-уровень для терминалов, SPEC §6 Q3) — остаётся отложенным из Фазы 2 (там владелец снял терминал из v1); в требования Фазы 3 НЕ входит, кандидат на Фазу 4/v2.
- OSD-уведомление при переключении раскладки (компенсация «необучаемого» индикатора, ADR-001) — вне объёма, переносится.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CORR-02 | Тройной Right Shift исправляет всю фразу из буфера | FSM уже эмитит Triple на истечении окна; Buffer хранит всю фразу — нужен API `заменить весь буфер`; единый run-based конвейер D-22/D-23/D-26; лестница D-27 с капом Backspace в YAML |
| CORR-03 | Двойной Right Shift при активном выделении исправляет выделенное | anchorPos уже приходит на провод SetSurroundingText, но НЕ доходит до EventHandler — расширение seam; выделение = anchor ≠ cursor (D-30); clipboard round-trip opt-in D-28/D-29 |
| CORR-06 | Смешанный текст: конвертируется только «чужая» часть (семантика ADR/D-22..D-24) | Обобщение `internal/correct` до прогонов; `Detect`/`Convert` переиспользуются; золотой корпус `ghbdtnпривет`→`приветпривет` и др. |
| SWCH-01 | Right Shift переключает раскладку EN↔RU (механизм ADR-001 — внутренний флип) | УЖЕ реализовано в `actor.flipScript` (план 02-04); Фаза 3 делает окно настраиваемым, оракул — запись `mode` в логе (D-34) |
| SWCH-02 | Комбо «исправить слово и переключить раскладку» (дефолт Shift+RightCtrl, переназначается) | FSM не знает комбо — расширение: press Control_R при зажатом Shift_R; IBUS_KEY_Control_R=0xffe4; порядок слово→флип фиксирован D-36 |
| SWCH-03 | Родное GNOME-переключение Super+Space не сломано (рабочая интерпретация D-34) | ydotool super+space физически активирует GNOME-свитчер (проба D-01-2) — e2e может инжектировать и убедиться, что goswitch не мешает; буквальный MRU-критерий устарел |
| SWCH-04 | Конфликт таймингов single/double/triple (ADR-002) | Исполнен корпусом `fsm_test.go`; Фаза 3 только выносит окно в YAML (D-35) |
| CONF-01 | Все хоткеи, таймауты и параметры — YAML-конфиг | Новый `internal/config`: yaml.v3 KnownFields + fsnotify v1.10.1 (новая зависимость, верифицирована) |
| CONF-02 | Изменения применяются без перезапуска (hot reload) | Directory-watch + debounce + `atomic.Pointer[Config]`; last-good по D-32 |
| CONF-03 | Совместимость схемы биндингов с Caramba (wish) | Таблица соответствия в доках, БЕЗ клонирования имён (D-31) — документационная задача, не кодовая |
| MACR-01 | Замена Super+Буква → Ctrl+Буква по приложениям (механизм ADR-005) | Перехват в ProcessKeyEvent (Mods&MaskMod4), ForwardKeyEvent Ctrl+Буква по прототипной последовательности; AT-SPI-наблюдатель опционален при per-app списках |
| INST-02 | CLI `goswitchctl`: статус, перечитать конфиг, принудительно скорректировать | Транспорт: D-Bus session bus (godbus `ConnectSessionBus`, проверено в module cache); имя org.djarvur.goswitch по ARCHITECTURE |
</phase_requirements>

## Project Constraints (from AGENTS.md / CONVENTIONS.md)

CLAUDE.md отсутствует; конвенции живут в `AGENTS.md` (генерирован из CONVENTIONS.md) и BIND-директивах фаз 1–2. Planner обязан нести их в каждый план:

- **Строгий TDD** (D-07): каждая поведенческая задача — red → green → refactor; Verify-блоки прогоняют тесты на каждом шаге.
- **Зелёная итерация** (D-08): план завершается только при `go build ./...`, `go test -race ./...`, `golangci-lint run` без ошибок. Никаких «поправим линт потом».
- **Сценарии — mise, не make** (D-09): все задачи сборки/тестов/линта/e2e — задачи `mise.toml`; Makefile не создаётся.
- **golangci-lint максимально строгий** (D-10): `.golangci.yml` v2 `default: all`, depguard deny `unsafe`, потолки сложности, `local-prefixes: github.com/Djarvur/goswitch`. Замечания: `tagliatelle` требует yaml-теги в snake_case; `nonamedreturns` отключён (recover-шимы требуют named returns); `hugeParam` threshold 512.
- **GitHub Actions** (D-11): pr-sanity на push/PR; dependabot weekly; cron govulncheck. Новых workflow в Фазе 3 не требуется (e2e-раннер green106 уже поднят, Фаза 2).
- **Приватность логов** (D-20/D-21): INFO — счётчики/причины без содержимого; детали (source/result) — только `-debug`. Распространяется на ВСЕ новые пути: фраза, выделение, clipboard (содержимое буфера обмена в лог — запрещено), MACR.
- **Go-скилл**: вся Go-работа идёт по скиллу go-ultimate (`.zcode/skills/go-ultimate/SKILL.md`): stdlib-first, `internal/` для приватного, `cmd/<name>/` для бинарников, контекст-first сигнатуры, `errors.Is/As`, `-race` всегда, без `pkg/`.
- **Зависимости**: модуль разрешает ровно godbus + yaml.v3 (+ fsnotify с этой фазы); gomodguard_v2 отключён с комментарием «module allows exactly godbus» — при добавлении fsupdate проверить, что конфиг не нужно синхронизировать (сейчас список не прописан жёстко — disable с обоснованием уже стоит).
- **Git workflow**: работа на ветке `gsd/phase-03-...` (уже создана), conventional commits, изменения через PR.

## Summary

Фаза 3 — надстройка над зрелым каркасом, а не новый фундамент. Три из четырёх больших блоков имеют готовые точки врезки, прочитанные в дереве этой сессией: FSM уже эмитит `Single/Double/Triple` на истечении окна (ADR-002 исполнен), `Double` уже запускает конвейер коррекции со сверкой ADR-004 и лестницей ADR-003, `Single` уже переключает внутренний режим (`flipScript`, план 02-04 — SWCH-01 фактически работает), `Triple` — пустая заглушка с отладочной записью. Буфер `internal/correct.Buffer` хранит ВСЮ фразу с указателем на токен — фраза это `b.runes` целиком, нужен только метод уровня «заменить весь буфер». `SetSurroundingText` на проводе уже несёт `anchorPos` (семантика: «Anchor position of selection in @text» — установленный заголовок), но seam `EventHandler.HandleSurroundingText(text string, cursorPos uint32)` его отбрасывает — детект выделения (D-30) требует расширения сигнатуры. Реальная новая инженерия: (1) run-based конвейер смешанного текста (D-22..D-24) в чистом `internal/correct` — золотой корпус, голова фазы; (2) пакет `internal/config` (yaml.v3 strict + fsnotify v1.10.1 — единственная новая зависимость) с hot reload last-good; (3) `goswitchctl` + ctlsvc на session bus (godbus уже в зависимостях, `ConnectSessionBus` сверен по module cache); (4) MACR-01 по ADR-005 с AT-SPI-наблюдателем как опциональным spike-гейтом.

Две живые неизвестности, которые план обязан закрыть прогонами, а не предположениями: поведение замены ВЫДЕЛЕНИЯ per-surface (commit_text заменяет активное выделение как «набор поверх выделения»? DeleteSurroundingText с положительным offset? — матрица v2 пинирует фактическое поведение zenity/Chromium/GTE так же, как Фаза 2 пинила фактический уровень лестницы) и доставка Super+Буква в ProcessKeyEvent (ADR-005 утверждает доставку модификатора; связка буквы для незабинженных комбинаций — live-прога). Обе деградируют по уже принятым правилам (clipboard opt-in D-28; WARN «consumed-upstream» ADR-005 b.2), то есть не блокируют архитектуру.

**Primary recommendation:** строить фазу в порядке «чистая логика → конфиг → провод → ctl/MACR → матрица v2»: сначала run-based конвейер в `internal/correct` (единый код-путь слова/фразы/выделения по D-23) + расширение Buffer, затем `internal/config` (strict + fsnotify + last-good) и конфигурируемость FSM (окно/биндинги/комбо), затем фраза/выделение в акторе с расширением seam anchorPos, затем ctlsvc + goswitchctl, MACR последним (его перехват изолирован в ProcessKeyEvent), и матрица v2 наращивается поверх готовых драйверов поверхностей. Каждая live-неизвестность получает ранний пробный кейс (spike) до встраивания основной логики.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Run-segmentation смешанного текста (CORR-06) | `internal/correct` (чистая логика) | — | Чистые функции над таблицами `layouts/`; headless-корпус под `-race`; ни D-Bus, ни часов (паттерн пакета) |
| Фраза CORR-02 (диапазон + лестница D-27) | `internal/correct` (геометрия) + `internal/session` (оркестрация) | `engine` (эмиттеры) | Арифметика замены — Plan/BuildPlan; таймеры сверки — актор |
| Детект выделения (CORR-03, D-30) | `engine` (декод anchorPos) → `internal/session` | — | Провод уже несёт anchorPos; seam расширяется, актор хранит anchor в кэше сверки |
| Clipboard round-trip (D-28/D-29) | `internal/session` (orкестрация) | subprocess wl-copy/wl-paste | Демон никогда не тянет Wayland-клиента (STACK); подпроцессы по прототипному паттерну |
| Переключение раскладки (SWCH-01, D-34) | `internal/session` (scriptMode = «layoutset») | — | ADR-001 Option B: чистое состояние демона; gsettings не участвует; оракул — запись `mode` в логе |
| Комбо SWCH-02 (D-36) | `internal/hotkey` (распознавание) + `internal/session` (порядок слово→флип) | — | Распознавание жеста — FSM/актор; порядок действий — актор |
| YAML-схема + strict decode (CONF-01, D-31/D-33) | `internal/config` (новый) | — | Чистый пакет: декод + валидация + дефолты; без D-Bus |
| Hot reload (CONF-02, D-32) | `internal/config` (watcher) | `internal/session` (потребление снапшотов) | fsnotify только в watcher'е; публикация через `atomic.Pointer[Config]` |
| Транспорт goswitchctl (INST-02) | `internal/ctlsvc` (новый, session bus) + `cmd/goswitchctl` | `internal/session` (in-process вызовы) | Второе godbus-соединение к session bus; CLI — клиент того же интерфейса |
| MACR-01 перехват (ADR-005) | `internal/session` (ветка в обработке клавиши) | `engine` (ForwardKeyEvent) | Super виден в Mods на ProcessKeyEvent; подмена — эмуляция Ctrl+Буква |
| MACR app-identity (per-app списки) | `internal/appid` (новый, AT-SPI через a11y bus) | — | Подключается ТОЛЬКО при per-app списках (ADR-005); деградация WARN → глобальные правила |
| e2e-оракулы v2 | `test/e2e` (стенд) | лог демона | Оракул переключения — внутреннее состояние (D-34), НЕ gsettings; оракул замены — контентный readback AT-SPI |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go (module `go 1.23`, toolchain mise 1.23) | 1.23 | язык | CONSTRAINT проекта; менять нельзя — совместимость релизных бинарников |
| `github.com/godbus/dbus/v5` | **v5.2.2** (в go.mod; сверено по proxy.golang.org этой сессией) | IBus-провод + НОВОЕ: session bus для ctlsvc и клиента goswitchctl | Единственный поддерживаемый нативный Go D-Bus; `ConnectSessionBus`/`SessionBus` сверен по module cache: conn.go:61 `func SessionBus() (conn *Conn, err error)`, conn.go:137 `func ConnectSessionBus(opts ...ConnOption) (*Conn, error)` [VERIFIED: module cache v5.2.2/conn.go:61,137] |
| `gopkg.in/yaml.v3` | **v3.0.1** (в go.mod) | strict-декод конфига (KnownFields, D-33) | Уже используется матрицей e2e (`dec.KnownFields(true)` — matrix.go:115); zero-dep |
| `github.com/fsnotify/fsnotify` | **v1.10.1** — НОВАЯ зависимость Фазы 3 (сверено: proxy.golang.org `{"Version":"v1.10.1","Time":"2026-05-04"}`; модуль присутствует в локальном кэше) | hot reload (CONF-02) | Стандартный inotify-врапер из STACK; API сверен по кэшу: `NewWatcher`, `Watcher.Add`, Events, ops Create/Write/Remove/Rename [VERIFIED: module cache fsnotify@v1.10.1/fsnotify.go] |
| wl-clipboard (`wl-copy`/`wl-paste`) | 2.2.1-1build1 — установлен в системе [VERIFIED: dpkg + /usr/bin/wl-copy] | clipboard round-trip D-28/D-29 через subprocess | STACK-паттерн без Wayland-зависимости в Go; man-страницы прочитаны локально (см. Code Examples) |
| systemd user units, ydotool 0.1.8, AT-SPI (python3-gi), google-chrome 153, gnome-text-editor 46.3, zenity | установлены [VERIFIED: phase-2 stand green] | e2e-стенд | Без изменений из Фазы 2 |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| stdlib `log/slog` | — | структурные логи новых событий (reload, mode, super-consumed, clipboard restore) | всегда |
| stdlib `flag` / `os.Args` | — | CLI `goswitchctl` (status/reload/correct) + флаг `-config` демона | CLI-фреймворк запрещён конвенцией |
| stdlib `os/exec` | — | wl-copy/wl-paste, gsettings (только стенд) | clipboard round-trip; контент — через stdin, никогда argv |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| fsnotify | polling `os.Stat` mtime | STACK: «genuinely acceptable» для одного файла; fsnotify оставляем как стандартный крошечный вариант — и STACK уже предписал directory-watch паттерн |
| D-Bus session bus для goswitchctl | локальный unix-сокет | Сокет = рукодельный протокол + свой path/permissions; D-Bus уже в зависимостях, session bus даёт SO_PEERCRED-изоляцию того же пользователя и автозапуск; вердикт исследователя: session bus (кандидат STACK подтверждён) |
| yaml.v3 strict | goccy/go-yaml strict | Механическая замена; НЕ менять в этой фазе (yaml.v3 уже в дереве и в матрице) |
| python3-gi helper как AT-SPI-наблюдатель демона | чистый Go на a11y bus | Демон НЕ должен требовать python; a11y bus интроспектируем (STACK) → чистый Go наблюдатель жизнеспособен; helper остаётся только в стендe |

**Installation:**
```bash
go get github.com/fsnotify/fsnotify@v1.10.1
go mod tidy   # golang.org/x/sys уже indirect — станет прямой транзитивкой fsnotify
```

**Version verification (выполнено этой сессией):**
```bash
curl -s https://proxy.golang.org/github.com/fsnotify/fsnotify/@latest   # {"Version":"v1.10.1","Time":"2026-05-04T10:39:17Z"}
curl -s https://proxy.golang.org/github.com/godbus/dbus/v5/@latest      # {"Version":"v5.2.2","Time":"2025-12-29"}
curl -s https://proxy.golang.org/gopkg.in/yaml.v3/@latest               # {"Version":"v3.0.1"}
```

## Package Legitimacy Audit

Шов `package-legitimacy check` не поддерживает экосистему Go (требует npm|pypi|crates) — для Go авторитетный реестр это proxy.golang.org, проверка выполнена напрямую + источник подтверждён официальным репозиторием и STACK-исследованием:

| Package | Registry | Age | Downloads/Зрелость | Source Repo | Verdict | Disposition |
|---------|----------|-----|--------------------|-------------|---------|-------------|
| github.com/fsnotify/fsnotify | proxy.golang.org | первая версия ~2012 (13+ лет) | де-факто стандарт Go-файлвотчинга; v1.10.1 tag на официальном git (hash 76b01a6e8f5...) | github.com/fsnotify/fsnotify | OK | Approved |
| github.com/godbus/dbus/v5 | proxy.golang.org | v5.x с 2017 | в проекте с Фазы 1, продакшн-след ibus-bamboo | github.com/godbus/dbus | OK | Approved |
| gopkg.in/yaml.v3 | proxy.golang.org | v3 с 2019, финал 2022 | экосистемный дефолт YAML | github.com/go-yaml/yaml (gopkg.in зеркало) | OK | Approved |
| wl-clipboard (внешний бинарник, не Go-модуль) | Ubuntu archive (dpkg) | проект с 2018 | стандартный Wayland CLI | github.com/bugaevc/wl-clipboard | OK | Approved |

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

Никаких иных новых пакетов фаза не вводит: MACR-наблюдатель — godbus по a11y bus; clipboard — subprocess; CLI — stdlib.

## Architecture Patterns

### System Architecture Diagram

Поток данных Фазы 3 (новые элементы помечены ★):

```
клавиатура → mutter → ibus-daemon → engine.ProcessKeyEvent
                                          │
                       ┌──────────────────┴──────────────────────┐
                       │ internal/session.Actor (1 goroutine, mutex)
                       │  ├─ hotkey.FSM (окно/биндинги из ★config)
                       │  │    Single → flipScript (SWCH-01, есть)
                       │  │    Double → выделение? ★CORR-03 : слово (есть)
                       │  │    Triple → ★фраза CORR-02
                       │  │    ★комбо Shift+RCtrl → слово, затем флип (SWCH-02)
                       │  ├─ ★MACR-ветка: Super+Буква → ForwardKeyEvent(Ctrl+Буква)
                       │  │    └─ per-app? → ★internal/appid → a11y bus (опц.)
                       │  └─ Buffer (вся фраза) + correct.* (★run-based)
                       │        correct.Detect/ConvertRuns/BuildPlan
                       ▼
          эмиттеры Emitter: CommitText / DeleteSurroundingText /
          ForwardKeyEvent / RequireSurroundingText ────────► ibus-daemon → поле
                       ▲
          SetSurroundingText(text, cursorPos, ★anchorPos) — seam расширяется:
          выделение = anchor ≠ cursor (D-30); сверка ADR-004 как в Фазе 2
                       │
          ★clipboard-ступень (opt-in D-28): wl-paste (save) → wl-copy (stdin)
          → ForwardKeyEvent(Ctrl+V) → wl-copy (restore, best-effort D-29)

  ★internal/config ──fsnotify(dir-watch)+debounce──> yaml.v3 KnownFields strict
        │  invalid → WARN + last-good (D-32); ok → atomic.Pointer[Config]
        └─> снапшоты читают: FSM (окно), correct (капы), MACR, clipboard-флаг

  ★internal/ctlsvc (session bus, org.djarvur.goswitch) ◄──★cmd/goswitchctl
        └─> Status / ReloadConfig / CorrectNow → in-process в актор
```

Первичный use case прослеживается: тапы → FSM-решение на истечении окна → run-конвертация → лестница/commit → сверка → verify-after — тот же протокол Фазы 2, расширенный на три новых диапазона (фраза/выделение/смешанный) и три новых источника параметров (config).

### Recommended Project Structure (дельта к дереву)

```
goswitch/
├── cmd/goswitchd/main.go        # + флаг -config; сборка актора из config.Snapshot
├── cmd/goswitchctl/main.go      # НОВЫЙ: status | reload | correct (INST-02)
├── engine/                      # расширение: anchorPos в EventHandler; keys.go + KeyControlR
├── internal/config/             # НОВЫЙ: схема, strict decode, watcher, last-good
│   ├── config.go                #    тип Config + Defaults() + Validate()
│   ├── load.go                  #    yaml.NewDecoder + KnownFields(true) (D-33)
│   ├── watch.go                 #    fsnotify dir-watch + debounce + atomic.Pointer
│   └── *_test.go                #    корпус: strict/defaults/last-good/debounce
├── internal/correct/            # расширение: runs.go (D-22/D-23), buffer.go (фраза)
│   ├── runs.go                  #    НОВЫЙ: сегментация прогонов + якорь
│   └── runs_test.go             #    золотой корпус смешанного текста
├── internal/hotkey/fsm.go       # расширение: конфигурируемые биндинги + комбо
├── internal/session/actor.go    # расширение: Triple/выделение/комбо/MACR/фраза
├── internal/ctlsvc/             # НОВЫЙ: D-Bus сервис на session bus
├── internal/appid/              # НОВЫЙ (опц., только per-app MACR): AT-SPI наблюдатель
└── test/e2e/                    # расширение: шаги select/combo/reload, cases/matrix-v2.yaml
```

### Pattern 1: Единый run-based конвейер коррекции (D-23)

**What:** один код-путь для слова/фразы/выделения: диапазон → сегментация на прогоны букв → якорь → конвертация чужих прогонов → BuildPlan → лестница.
**When to use:** все CORR-02/03/06 обработки.
**Example:** см. Code Examples §1.

### Pattern 2: Снапшоты конфига через atomic.Pointer (hot reload без локов)

**What:** watcher публикует иммутабельный `*Config`; актор читает `cfg.Load()` один раз на событие.
**When to use:** CONF-02; окно FSM, капы, MACR-флаги.
**Example:** см. Code Examples §3.

### Pattern 3: Directory-watch + debounce для atomic-rename редакторов

**What:** fsnotify следит за КАТАЛОГОМ конфига, события фильтруются по имени файла, debounce 150–300 мс, затем полный ре-парс и атомарная подмена (STACK, «Stack Patterns by Variant»). vim/kwrite/gedit пишут через write-temp+rename — вотчинг пути теряет inode.
**When to use:** watch.go. Пере-добавлять watch после rename не нужно при вотчинге каталога.

### Pattern 4: Clipboard round-trip по прототипу (D-28/D-29)

**What:** сохранить буфер (`wl-paste --no-newline`; пустой буфер = ненулевой exit → восстановление = `wl-copy --clear`), подменить (`wl-copy --trim-newline` с контентом через STDIN), вставить поверх выделения (ForwardKeyEvent Ctrl+V — валидированная последовательность прототипа), восстановить (`wl-copy` stdin).
**When to use:** только когда ступень включена конфигом И commit/delete не сработали (verify-after mismatch).

### Pattern 5: ctlsvc на session bus (второе godbus-соединение)

**What:** демон открывает `dbus.ConnectSessionBus()`, `RequestName("org.djarvur.goswitch")`, экспортирует объект с методами Status/ReloadConfig/CorrectNow; goswitchctl — клиент. ARHITECTURE.md предписывал `ctlsvc` именно так («second godbus connection in same process»).
**When to use:** INST-02.

### Pattern 6: MACR-перехват до буфера (ADR-005)

**What:** ветка обработки клавиши ВЫШЕ feedKey: если Super в Mods + буква из настроенного множества + (глобально или app совпал) → consume=true + ForwardKeyEvent(Ctrl+Буква). Буфер не кормится (comboMask уже исключает Mod4 — actor.go:35,370).
**When to use:** MACR-01; в обеих раскладках (RU-коммит не должен сработать на Super+буква — ветка MACR идёт раньше mode-ветки).

### Anti-Patterns to Avoid

- **Блокировать ProcessKeyEvent ожиданием окна/сверки** — реплай синхронный; все ожидания в таймерах, пере-входящих в актор (действующий инвариант Фазы 1–2, actor.go комментарий «The ProcessKeyEvent answer never waits»).
- **Логировать содержимое буфера обмена / выделения на INFO** — приватность D-20/D-21 распространяется на clipboard: только счётчики и причины; содержание — максимум `-debug`.
- **Передавать clipboard-контент аргументом командной строки** — argv виден в `ps` однопользовательским процессам и может быть распарсен как флаги (`-`-префиксы; man wl-copy явно оговаривает `--`). Только stdin-пайп (паттерн прототипа).
- **Клонировать Caramba-имена ключей** — D-31 запрещает: их схема привязана к «свитч на первый тап», наша — к ADR-002; совместимость = таблица соответствия в доках.
- **Считать gsettings ораколом переключения** — D-34/ADR-001: чтение dconf запрещено как оракул (Pitfall 3 ADR-001); оракул — запись `mode` в логе демона.
- **Делать выделение через чтение clipboard по умолчанию** — D-28: clipboard-ступень opt-in и выключена по умолчанию; первичный путь — surrounding text (anchor ≠ cursor).
- **Хардкодить новые keyval'ы из памяти** — значения брать из `/usr/include/ibus-1.0/ibuskeysyms.h` (см. Code Examples §6; ловушка класса 02-05: имя ≠ клавиша).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Файловые уведомления | epoll/inotify через syscall | fsnotify v1.10.1 | edge-кеймы atomic-rename, масок, переподключения; STACK предписывает |
| YAML-парсинг | регэкспы/свой парсер | yaml.v3 + KnownFields | strict-семантика D-33 из коробки (`dec.KnownFields(true)` — рабочий паттерн matrix.go:114-116) |
| D-Bus-транспорт ctl | свой unix-socket протокол | godbus session bus | имя/activation/peer-creds бесплатно; godbus уже в дереве |
| Буфер обмена Wayland | Wayland-клиент на Go | subprocess wl-copy/wl-paste | STACK: никакого Wayland-клиента; прототип валидирован на этой машине |
| Тайминги тапов | новая схема дискриминации | существующий FSM-корпус | D-35: правила ADR-002 уже исполнены `fsm_test.go` (11 тестов) — конфигурируемость НЕ должна их сломать |
| Таблицы раскладок | новые таблицы | существующие `layouts/` | D-23: «таблицы layouts/ переиспользуются как есть» (CONTEXT code_context) |

**Key insight:** фаза — это композиция уже проверенных примитивов; единственная принципиально новая механика — run-сегментация (чистая функция) и AT-SPI-наблюдатель (опциональный, за spike-гейтом).

## Runtime State Inventory

> Фаза — функциональная надстройка (greenfield-фичи), НЕ rename/refactor/migration. Раздел опущён согласно шаблону. (Краткая справка на случай смежных вопросов: фаза НЕ переименовывает D-Bus-имена, пути или ключи конфигурации; component name `org.freedesktop.IBus.goswitch` неизменен — matrix.go:82; записей ОС/сервисов с фазовыми строками нет.)

## Common Pitfalls

### Pitfall 1: anchorPos теряется на seam — выделение «невидимо» для актора
**What goes wrong:** провод несёт anchorPos, но `HandleSurroundingText(t.Text, cursorPos)` (engine.go:192) его отбрасывает; актор не может отличить выделение от простого курсора.
**Why it happens:** Фазе 2 якорь был не нужен.
**How to avoid:** расширить сигнатуру seam до `(text string, cursorPos, anchorPos uint32)` одной правкой; все реализаторы/даблы обновить в той же задаче (интерфейс мал: Actor + test doubles).
**Warning signs:** e2e-кейс выделения даёт «no-selection» отказ при видимом выделении.

### Pitfall 2: strict-конфиг молча проходит из-за нестрогого декодера
**What goes wrong:** опечатка в ключе YAML игнорируется → пользователь думает, что применил таймаут.
**Why:** дефолт yaml.v3 не строгий.
**How to avoid:** `dec.KnownFields(true)` в load.go (D-33); юнит-тест на неизвестный ключ, отсутствующее поле, неверный тип; last-good при любой ошибке (D-32).
**Warning signs:** тест «typo-key rejected» отсутствует; reload не пишет WARN при битом конфиге.

### Pitfall 3: clipboard round-trip ломает буфер пользователя (или восстанавливает с \n)
**What goes wrong:** (a) restore перезаписывает пустой буфер текстом; (b) `wl-paste` без `--no-newline` добавляет перевод строки, `wl-copy` без `--trim-newline` копирует хвостовой \n из пайпа.
**Why:** man wl-clipboard: wl-paste аппендит newline по умолчанию; wl-copy форкается и сервит буфер в фоне.
**How to avoid:** сохранять `wl-paste --no-newline` (байт-точно); пустой буфер (ненулевой exit) → восстановление через `wl-copy --clear`; запись через stdin c `--trim-newline`; неудача восстановления — только WARN (D-29).
**Warning signs:** e2e не проверяет содержимое буфера до/после; лог содержит clipboard-контент.

### Pitfall 4: комбо Shift+RightCtrl убивает серию тапов вместо действия
**What goes wrong:** press Control_R при зажатом Shift_R в нынешнем FSM ставит `sawKey=true` (модификаторное использование) и молча гасит серию — комбо не существует для FSM.
**Why:** FSM знает только Shift_R-тапы (fsm.go:85-98).
**How to avoid:** распознавать комбо до/aside FSM-ветки (актор видит press Control_R при зажатом Shift_R); конфигурируемый биндинг (D-31) — таблица имя→keyval; НЕ ломать модификаторную дискриминацию для остальных клавиш (корпус fsm_test.go должен остаться зелёным).
**Warning signs:** TestFSM_ModifierUse падает после refactor; e2e комбо не даёт ни коррекции, ни флипа.

### Pitfall 5: имена клавиш ydotool разрешаются не в те физические клавиши
**What goes wrong:** `space`→физическая S, `Escape`→физическая E (живые находки 02-03/02-05) — селект-шаги `ctrl+a`/`shift+Right` могут попасть не туда.
**Why:** fallback первой буквы в таблице имён ydotool 0.1.8.
**How to avoid:** живая проба каждого нового ключевого имени в матрице (профлайт/отдельный smoke-кейс) до включения в схему; канонизировать рабочие имена в `matrixKeyNames()`-таблицу, как сделано для esc/enter/tab.
**Warning signs:** readback после селекта показывает пустую замену/не тот диапазон.

### Pitfall 6: верификация выделения сверяет суффикс «до курсора», а выделение может быть правее курсора
**What goes wrong:** `MatchesSuffix` проверяет только текст перед курсором; при cursor < anchor (выделение правее) geometry уровня 1 отличается (положительный offset удаления).
**Why:** Фаза 2 имела только диапазон перед курсором.
**How to avoid:** план замены для выделения вычисляет offset со знаком по взаимному положению anchor/cursor; verify-after для выделения — наличие конвертированного текста на месте диапазона; всё пинится юнит-тестами геометрии + e2e.
**Warning signs:** e2e кейс «select left-to-right vs right-to-left» отсутствует.

### Pitfall 7: MACR-перехват ловит Super-комбинации GNOME и ломает оболочку
**What goes wrong:** перехват Super+буква для букв, забинженных mutter (обзор, скриншоты), конфликтуен; также перехват буквы, когда пользователь хочет системную комбинацию.
**Why:** mutter биндит часть Super+буква раньше IME (ADR-005 b).
**How to avoid:** (a) перехватывать только сконфигурированное множество букв; (b) детект «Super вниз → нет буквы → Super вверх» = consumed-upstream → WARN `super-combo-consumed-upstream` + счётчик в goswitchctl (ADR-005 b.2); (c) живая проба фактической доставки Super+буква до написания логики.
**Warning signs:** в логе нет записей Super press/release при живой сессии; e2e не различает «не дошло» и «не сконфигурировано».

### Pitfall 8: hot reload окна рассинхронизирует FSM-таймер
**What goes wrong:** окно сменилось, а взведённый `time.AfterFunc` ждёт по-старому (или серия, начатая при старом окне, валидируется новым).
**Why:** актор армит таймер значением на момент тапа (actor.go:530-535).
**How to avoid:** простое правило: новое окно действует на серии, начатые после reload; взведённый таймер дожить по старому значению (FSM-инвариант «stale timer is no-op» уже это переживает — fsm_test TestFSM_WindowEdges). Зафиксировать юнит-тестом.
**Warning signs:** flaky-тест после reload; попытка «перевичать» таймер в reload-пути (не надо).

### Pitfall 9: фраза длиннее surrounding-окна клиента
**What goes wrong:** клиент пушит усечённый surrounding text → сверка ADR-004 не находит весь диапазон фразы → вечный verify-timeout.
**Why:** клиенты ограничивают окно surrounding text.
**How to avoid:** сверять ДОСТУПНУЮ часть: суффикс фразы, попавший в окно сверки (D-25: «сверка сама отрезает неваличный хвост» — обратное тоже верно: невалидная голова = отказ); живая проба на фразах ~50-100 рун; при капе Backspace (D-27) — тихий отказ с причиной.
**Warning signs:** e2e «длинная фраза» зелёный только на коротких прогонах.

## Code Examples

### §1. Run-сегментация смешанного текста (ядро CORR-06, скелет по D-22/D-23)

```go
// internal/correct/runs.go — скелет; семантика D-22/D-23/D-24, корпус в runs_test.go
// Якорь: раскладка последней БУКВЫ диапазона (цифры/общая пунктуация не сдвигают,
// продолжение D-15). Чужие прогоны конвертируются каждым независимо (Convert
// переиспользуется как есть — convert.go:15).
//
// Золотые примеры (CONTEXT specifics, обязательны в корпусе):
//   ghbdtnпривет   → приветпривет   (якорь RU от …вет)
//   gfbпривет      → паипривет
//   приветghbdtn   → ghbdtnghbdtn   (якорь EN)
//   ghbdtn2026привет → привет2026привет (цифры якорь не сдвигают)
```

Существующие примитивы, на которые опирается скелет [VERIFIED: internal/correct/direction.go:22-42 — `Detect` возвращает `(Dir, bool)`: Latin-only → `ENtoRU`, Cyrillic-only → `RUtoEN`, смешанный/без букв → `false`; internal/correct/convert.go:15-40 — `Convert` отображает буквы через таблицу, не-буквы едут тождественно, отсутствующая буква валит всю конвертацию]:

```go
// Существующие сигнатуры (дословно, direction.go / convert.go):
func Detect(token []rune) (Dir, bool)
func Convert(token []rune, dir Dir) ([]rune, bool)
```

### §2. Расширение seam выделения (дословные текущие формы)

Текущий провод [VERIFIED: engine/engine.go:187-197]:

```go
func (e *Engine) SetSurroundingText(text dbus.Variant, cursorPos, anchorPos uint32) (err *dbus.Error) {
	...
	slog.Debug("surrounding_text", "cursor_pos", cursorPos, "anchor_pos", anchorPos)
	if e.handler != nil {
		if t, ok := decodeIBusText(text); ok {
			e.handler.HandleSurroundingText(t.Text, cursorPos) // anchorPos теряется — расширить
		}
	}
	...
}
```

Текущий seam [VERIFIED: engine/engine.go:88-94]:

```go
// HandleSurroundingText delivers the surrounding text the client
// reported (the runes before the anchor at cursorPos) — the input of the
// ADR-004 pre-correction verification.
HandleSurroundingText(text string, cursorPos uint32)
```

Целевая форма: `HandleSurroundingText(text string, cursorPos, anchorPos uint32)`; семантика якоря из установленного заголовка [VERIFIED: /usr/include/ibus-1.0/ibusengine.h:430 — `@anchor_pos: (out) (allow-none): Anchor position of selection in @text.`]. Выделение активно ⇔ `anchorPos != cursorPos` (D-30). Кэш актора `surr` (actor.go:199) дополняется якорем.

### §3. Hot reload: strict decode + directory-watch + debounce (STACK-паттерн, адаптировано)

```go
// internal/config/load.go — strict decode (D-33); паттерн уже работает в матрице:
// [VERIFIED: test/e2e/matrix.go:113-116] dec := yaml.NewDecoder(...); dec.KnownFields(true)
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil { return nil, fmt.Errorf("open config: %w", err) }
	defer f.Close()
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true) // D-33: неизвестный ключ = невалидный конфиг целиком
	var cfg Config
	if err := dec.Decode(&cfg); err != nil { return nil, fmt.Errorf("decode config: %w", err) }
	if err := cfg.Validate(); err != nil { return nil, fmt.Errorf("validate config: %w", err) }
	return &cfg, nil
}

// internal/config/watch.go — скелет watcher'а (fsnotify v1.10.1, API сверен по module cache)
// КАТАЛОГ, не файл: vim/kwrite/gedit пишут write-temp+rename → inode файла меняется.
// Debounce 150–300 мс, публикация atomic.Pointer, невалид = WARN + last-good (D-32).
w, _ := fsnotify.NewWatcher()
_ = w.Add(filepath.Dir(path))          // каталог
for ev := range w.Events {
	if filepath.Base(ev.Name) != filepath.Base(path) { continue }
	if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 { continue }
	// → debounce timer → Load(path): ok → ptr.Store(cfg); err → slog.Warn("config reload rejected", "error", err)
}
```

### §4. Clipboard round-trip (валидированный прототипный паттерн + локальные man-факты)

Прототип владельца (живая машина, прочитан этой сессией) [VERIFIED: /home/nil/.local/share/punto-switcher/punto_engine.py:273-305]:

```python
sel = subprocess.run(['wl-paste', '--primary'], capture_output=True, text=True, timeout=1.5).stdout
...
p = subprocess.Popen(['wl-copy'], stdin=subprocess.PIPE, text=True); p.communicate(new_text)
self.forward_key_event(IBus.KEY_Control_L, 29, 0)
self.forward_key_event(IBus.KEY_v, 55, 4)               # 4 = CONTROL_MASK
self.forward_key_event(IBus.KEY_v, 55, 0x40000004)      # release
self.forward_key_event(IBus.KEY_Control_L, 29, 0x40000004)
```

Флаги wl-clipboard 2.2.1 [VERIFIED: локальные man-страницы, dpkg 2.2.1-1build1]: `wl-copy` по умолчанию **форкается и сервит запросы в фоне** (`-f` — foreground); `-n/--trim-newline` — не копировать хвостовой перевод строки; `-c/--clear`; `-p/--primary`; `wl-paste -n/--no-newline` — байт-точный вывод. Go-скелет:

```go
// сохранение (перед подменой): байт-точное; пустой буфер = ненулевой exit → restore через wl-copy --clear
out, err := exec.Command("wl-paste", "--no-newline").Output()
// подмена: контент ТОЛЬКО через stdin (argv утекает в ps и парсится как флаги)
cmd := exec.Command("wl-copy", "--trim-newline")
cmd.Stdin = strings.NewReader(replacement)
_ = cmd.Run() // форкнет и будет сервить в фоне — это норма
// вставка поверх выделения: ForwardKeyEvent-последовательность Ctrl+V из прототипа
// восстановление: тот же stdin-паттерн с out; неудача — только WARN (D-29)
```

### §5. ctlsvc на session bus (godbus, сверен по module cache)

[VERIFIED: module cache godbus v5.2.2/conn.go:61 `func SessionBus() (conn *Conn, err error)`; conn.go:137 `func ConnectSessionBus(opts ...ConnOption) (*Conn, error)`]

```go
// internal/ctlsvc — скелет сервиса (имя по ARCHITECTURE.md: org.djarvur.goswitch)
conn, err := dbus.ConnectSessionBus()          // второе соединение демона, session bus
reply, _ := conn.RequestName("org.djarvur.goswitch", dbus.NameFlagReplaceExisting)
_ = conn.Export(&Svc{actor: actor, cfg: cfgPtr}, dbus.ObjectPath("/org/djarvur/goswitch"), "org.djarvur.goswitch")

// cmd/goswitchctl — клиент (stdlib flag; никакого CLI-фреймворка — конвенция)
conn, _ := dbus.ConnectSessionBus()
obj := conn.Object("org.djarvur.goswitch", dbus.ObjectPath("/org/djarvur/goswitch"))
call := obj.CallWithContext(ctx, "org.djarvur.goswitch.Status", 0)
```

Методы (дискреция планировщика по составу; минимум INST-02 + ADR-005 b.2): `Status() (s, error)` — режим EN/RU, конфиг (путь/валидность/last-good ошибка), счётчики коррекций, перехваченные Super-комбинации; `ReloadConfig() (s, error)`; `CorrectNow() (s, error)` — принудительный запуск конвейера слова (семантика двойного тапа без тапа). goswitchctl читаем человеком, `--json` опционален.

### §6. Ключевые константы (дословно из установленных заголовков и дерева)

[VERIFIED: /usr/include/ibus-1.0/ibuskeysyms.h:191 — `#define IBUS_KEY_Control_R 0xffe4`] — НОВЫЙ keyval для комбо SWCH-02; KeyControlL = 0xffe3 уже есть [VERIFIED: engine/keys.go:35].
Существующие маски/ключи [VERIFIED: engine/keys.go:5-12,24-37]: `MaskShift = 1<<0; MaskControl = 1<<2; MaskMod1 = 1<<3; MaskMod4 = 1<<6; MaskRelease = 1<<30`; `KeyBackSpace=0xff08, KeyTab=0xff09, KeyReturn=0xff0d, KeyEscape=0xff1b, KeyKPEnter=0xff8b, KeyShiftL=0xffe1, KeyShiftR=0xffe2, KeyControlL=0xffe3, KeySpace=0x020`.
FSM [VERIFIED: internal/hotkey/fsm.go:11-19]: `DefaultWindow = 300 * time.Millisecond`; `KeyvalShiftR = 0xffe2`; `MaxTaps = 3`.
Точка инъекции окна сегодня [VERIFIED: cmd/goswitchd/main.go:53 — `actor := session.NewActor(hotkey.DefaultWindow)`] — Фаза 3 подставляет значение из config.

### §7. Диспетчеризация FSM-решений сегодня (куда врезаются Triple/комбо/выделение)

[VERIFIED: internal/session/actor.go:253-269]:

```go
for _, action := range a.fsm.Feed(hotkey.TimerExpired{}, now) {
	slog.Info("action", "n", int(action))
	switch action {
	case hotkey.Double:
		a.startCorrection()
	case hotkey.Single:
		a.flipScript()
	case hotkey.Triple:
		slog.Debug("action deferred", "n", int(action)) // 02-05 phrase  ← врезка CORR-02
	}
}
```

Выделение (D-30) врезается в `startCorrection`: при активном якоре (anchor≠cursor из последнего пуша) диапазон = выделение; иначе — слово как сегодня. Флип-комбо (D-36) — «слово, затем flipScript()» в этом же свитче/маршрутизаторе.

### §8. e2e-инъекция селекта (ydotool 0.1.8 — синтаксис проверен по локальному --help)

[VERIFIED: `ydotool key --help`, 0.1.8-3build1]: «Each key sequence can be any number of modifiers and keys, separated by plus (+). For example: alt+r Alt+F4 CTRL+alt+f3…» — комбинации вида `ctrl+a` поддержаны синтаксисом. Ловушка имён (Pitfall 5) закрывается живой пробой: `shift+Right`, `ctrl+a`, `Control_R` — рабочие имена пинируются smoke-кейсом до матрицы.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Переключение раскладки через gsettings/two-engine (STACK-эра исследования) | Внутренний флип за «один хозяин» (ADR-001 Accepted, D-34 BIND) | гейт M0 2026-09-10 | gsettings-механика мертва для продукта; «layoutset» = интерфейс над scriptMode; оракул e2e — лог демона |
| «Свитч на первый тап» (Caramba-модель, кандидат SWCH-04) | Классика с ожиданием, все действия на истечении окна (ADR-002) | гейт M0 2026-09-10 | CONF-03 = таблица соответствия имён, НЕ клонирование схемы |
| Выделение через clipboard-хак прототипа (wl-paste --primary) | Surrounding-text якорь как первичный детект (D-30); clipboard — opt-in ступень замены (D-28) | CONTEXT фазы 3, 2026-09-15 | clipboard-чтение больше не источник выделения; путь прототипа остаётся механикой ступени |
| Направление слова голосованием букв (CORR-04, Фаза 2) | Прогоны с якорем «последняя буква» для смешанного (D-22/D-23); чистый диапазон — прежний контракт (см. Open Questions Q1) | CONTEXT фазы 3 | единый конвейер; матрица v1 обязана остаться зелёной |

**Deprecated/outdated:**
- Формулировки SWCH-03/критерия №3 ROADMAP про Super+Space/MRU — устарели против ADR-001 (D-34); верификатор не требует буквального исполнения.
- `hotkey.DefaultWindow` как единственный источник окна — с Фазы 3 дефолт живёт в config; константа остаётся значением дефолта (пин `TestFSM_DefaultWindow` сохранить).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `CommitText` заменяет активное выделение в GTK/Chromium-классе (семантика «набор поверх выделения»), а `DeleteSurroundingText` применим к диапазону выделения со знаком offset | CORR-03 / Patterns | Замена выделения пойдёт по clipboard-ступени везде или молча откажет; матрица v2 обязана пинировать фактическое поведение per-surface до фиксации порядка ступеней |
| A2 | GTK-класс (zenity/GTK4, Gnome Text Editor) пушит anchorPos, отражающий реальное выделение; поведение Chromium 153 с якорем неизвестно | CORR-03 | Если Chromium якорь не шлёт — его кейсы выделения дают тихий отказ D-30 (документированная деградация, не блокер) |
| A3 | Surrounding text клиенты присылают с диапазоном, покрывающим выделение/фразу матричных размеров | CORR-02/03 | Сверка ADR-004 отклонит длинные диапазоны → тихий отказ; живая проба на фразах ~50–100 рун до фиксации кейсов |
| A4 | AT-SPI-наблюдатель на a11y bus (org.a11y.Bus.GetAddress → focus-events) реализуем на godbus с приемлемой задержкой | MACR / appid | Деградация по ADR-005: WARN «app-identity unavailable» + только глобальные правила; per-app списки откатываются на spike-решение |
| A5 | Селект-инъекции ydotool (`ctrl+a`, `shift+Right`) работают на поверхностях стенда (синтаксис подтверждён, разрешение имён — нет) | e2e v2 | Кейсы выделения нестабильны; закрывается живым smoke-прогоном до матрицы (Pitfall 5) |
| A6 | Семантика якоря D-22 применяется к СМЕШАННЫМ диапазонам; однородный диапазон конвертируется целиком в другую раскладку (непрерывность CORR-01/04 и зелёной матрицы v1) | CORR-06 / Q1 | При буквальном якоре для чистого текста `ghbdtn` стал бы no-op — прямое противоречие зелёной матрице v1; требует подтверждения золотым корпусом (Q1) |
| A7 | Super+незабинженная-буква доходит до ProcessKeyEvent (модификатор — подтверждён ADR-005 b.1; буква — предположение) | MACR | Если буква не доходит — весь MACR вырождается в детект consumed-upstream; живая проба до реализации |
| A8 | `wl-copy`-форк не конфликтует с буфер-менеджером GNOME на цели (GNOME Shell не перехватывает владение агрессивно) | Clipboard | Restore может быть перекрыт буфер-менеджером; D-29 уже квалифицирует восстановление как best-effort |
| A9 | Для goswitchctl достаточно методов без свойств/сигналов (только вызовы) | ctlsvc | Расширение интерфейса позже — дешёвое (новый метод на том же объекте) |

## Open Questions

1. **Семантика D-24 «нечего конвертировать» и якорь для однородного диапазона (A6).**
   - What we know: D-22 определяет якорь для смешанного текста; D-23 требует единый код-путь; матрица v1 (зелёная, регрессионный арест) требует `ghbdtn`→`привет` для чистого слова.
   - What's unclear: буквальное применение якоря-по-последней-букве к ОДНОРОДНОМУ тексту даёт no-op (D-24) и ломает CORR-01; D-24 заявлен как достижимый кейс «все буквы уже в якорной раскладке», который при смешанном якоре арифметически недостижим.
   - Recommendation: закрепить в золотом корпусе: (a) смешанный диапазон → якорь = последняя буква, конвертируются чужие прогоны; (b) однородный диапазон → конвертация целиком в другую раскладку (непрерывность CORR-01/04); (c) post-segmentation чужих прогонов нет (защитный случай, напр. диапазон без букв при фразе/выделении) → успех-без-изменений D-24; (d) пустой буфер/нет текста → тихий отказ D-20. Если планировщик/владелец читает D-22 иначе — сверка на discuss-гейте до корпуса.
2. **Фактическое поведение замены выделения per-surface (A1/A2).**
   - What we know: механики доступны (commit, DeleteSurroundingText ±offset, clipboard); D-28 фиксирует порядок ступеней.
   - What's unclear: какая ступень реально срабатывает в zenity/GTK4, GTE, Chromium 153.
   - Recommendation: ранний живой smoke (spike) по образцу 02-05 «ladder-chromium»: пинировать фактическую ступень per-surface, затем переносить в expect_level-подобные поля матрицы v2.
3. **Доставка Super+Буква и матрица перехваченных GNOME комбинаций (A7).**
   - Recommendation: spike ProcessKeyEvent-трассы при живых Super+буква (забинженная и свободная) до реализации MACR; результат определяет дефолт множества и объём WARN-детекта.
4. **Формат биндингов в YAML (дискреция).**
   - Recommendation: имена-строки («Shift_R», «Shift+Control_R», «Super+c»→«Ctrl+c» для MACR), резолв через таблицу имён к keyval из keys.go; неизвестное имя = невалидный конфиг (D-33). Секции ровно `hotkeys` / `timeouts` / `correction` / `macr` (D-31), yaml-теги snake_case (tagliatelle).
5. **CorrectNow — состав принудительной коррекции.**
   - Recommendation: минимум — слово (семантика двойного тапа); расширение scope-аргументом (`word|phrase`) дёшево, но не обязательно для INST-02.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| wl-clipboard (wl-copy/wl-paste) | CORR-03/D-28/D-29 | ✓ | 2.2.1-1build1 [VERIFIED: dpkg] | — (ступень opt-in, при отсутствии бинарника — тихий отказ с причиной) |
| ydotool | e2e v2 (селекты, тапы, комбо) | ✓ | 0.1.8-3build1 | — |
| AT-SPI (python3-gi, gir1.2-atspi) | e2e-стенд | ✓ | 3.48.2 / 2.52.0 (STACK, стенд Фазы 2 зелёный) | — |
| a11y bus (org.a11y.Bus) | MACR per-app наблюдатель | ✓ (стандарт GNOME-сессии; адрес через session bus) | — | Деградация ADR-005: WARN + глобальные правила |
| session D-Bus | goswitchctl/ctlsvc | ✓ | `DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1001/bus` [VERIFIED: env] | — |
| google-chrome | e2e-матрица v2 (Chromium-поверхность) | ✓ | 153.x (STATE) | snap chromium (константы стенда, surface.go:19-26) |
| gnome-text-editor | e2e (многострочная поверхность) | ✓ | 46.3 (STATE) | — |
| zenity | e2e (GTK-класс) | ✓ | системный | — |
| proxy.golang.org (сеть) | `go get fsnotify@v1.10.1` | ✓ | проверено этой сессией | vendoring не требуется |
| Go/golangci-lint (mise) | все задачи | ✓ | mise 1.23 / lint 2 | — |

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** none (все присутствуют).

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing`, `package xxx_test`, `-race` обязательно (go-ultimate; D-08) |
| Config file | none — стандартный `go test`; mock-фреймворки запрещены конвенцией (testify отсутствует) |
| Quick run command | `mise run test` (= `go test -race -count=1 ./...`) |
| Full suite command | `mise run ci` (= build + vet + lint + test) |
| Live e2e | `mise run e2e-matrix` + новые e2e-задачи mise для spike-кейсов (живая сессия, НЕ в headless CI) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CORR-02 | Тройной тап: фраза из буфера конвертируется, лестница D-27, кап Backspace | unit + e2e | `go test -race ./internal/correct/ ./internal/session/ -run 'Phrase\|BackspaceCap'` | ❌ Wave 0 |
| CORR-03 | Выделение: anchor≠cursor → диапазон выделения; замена; clipboard opt-in + restore | unit + e2e | `go test -race ./internal/session/ -run Selection` | ❌ Wave 0 |
| CORR-06 | Run-сегментация: якорь последней буквы, чужие прогоны, цифры нейтральны, D-24 no-op | unit (golden) | `go test -race ./internal/correct/ -run Runs` | ❌ Wave 0 |
| SWCH-01 | Single → флип; окно из конфига; оракул = запись mode | unit (есть частично 02-04) + e2e | `go test -race ./internal/session/ -run 'Mode\|Window'` | ◐ расширить |
| SWCH-02 | Комбо Shift+RCtrl: слово затем флип; переназначение | unit + e2e | `go test -race ./internal/hotkey/ ./internal/session/ -run Combo` | ❌ Wave 0 |
| SWCH-04 | Окно настраивается; корпус FSM не сломан | unit | `go test -race ./internal/hotkey/` | ✅ (fsm_test.go) + новые |
| CONF-01/02/03 | strict decode, defaults, unknown-key reject, hot reload last-good, debounce | unit | `go test -race ./internal/config/` | ❌ Wave 0 |
| MACR-01 | Super+буква → consume + ForwardKeyEvent(Ctrl+буква); consumed-upstream WARN; per-app деградация | unit + spike/e2e | `go test -race ./internal/session/ -run MACR` | ❌ Wave 0 |
| INST-02 | ctlsvc методы; goswitchctl вызовы; статус с last-good | unit (+интеграция на живой шине e2e) | `go test -race ./internal/ctlsvc/` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `mise run test` (+ узкие `-run` фильтры в TDD-цикле)
- **Per wave merge:** `mise run ci` (build/vet/lint/test-race — зелёная итерация D-08)
- **Phase gate:** `mise run ci` зелёной + e2e-матрица v2 зелёная на живом столе/раннере (полная широта: фразы, выделение, смешанный текст, регистры, разные приложения) перед `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/correct/runs_test.go` — золотой корпус CORR-06 (примеры CONTEXT specifics дословно)
- [ ] `internal/config/config_test.go` / `load_test.go` / `watch_test.go` — strict/defaults/last-good/debounce (watcher за интерфейсом для синтетических событий)
- [ ] `internal/hotkey/fsm_test.go` — расширение: конфигурируемый биндинг, комбо (существующие 11 тестов остаются зелёными без правок семантики)
- [ ] `internal/session/actor_test.go` — фраза/выделение/комбо/MACR/anchorPos-seam
- [ ] `internal/ctlsvc/*_test.go` — методы сервиса
- [ ] e2e: smoke-кейсы селекта/Super-доставки/clipboard (spike) → `cases/matrix-v2.yaml` + расширение схемы шагов (select/combo/config-reload)
- [ ] Инфраструктура: флага `-config` у демона нет — Wave 0 (мелочь, но блокирует e2e конфиг-кейсы)

## Security Domain

`security_enforcement: true` (config.json), ASVS L1. Фаза расширяет поверхность: конфиг, clipboard, session-bus сервис, подмена клавиш.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|------------------|
| V2 Authentication | no | однопользовательский демон; session bus идентифицирует пира ядром (SO_PEERCRED) |
| V3 Session Management | no | — |
| V4 Access Control | yes | ctlsvc на session bus:同一-пользовательский scope (адрес `/run/user/1001/bus`); единственный владелец имени через RequestName — повторPattern single-instance guard как engine/conn.go:93-101 |
| V5 Input Validation | yes | yaml.v3 `KnownFields(true)` (D-33) + Validate() (диапазоны: window >0, cap >0, имена биндингов из закрытой таблицы); буфер обмена — внешние данные, в лог не попадают |
| V6 Cryptography | no | — |
| V12 File/Config | yes | конфиг по XDG-пути без выхода за `$HOME`; atomic-rename учтён directory-watch'ем |

### Known Threat Patterns for {stack}

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Инъекция в argv subprocess (clipboard-контент как флаг/утечка в ps) | Tampering / Information Disclosure | контент только через stdin-пайп (паттерн прототипа); `--` не нужен при stdin-подходе |
| Утечка набранного/clipboard-контента в логи | Information Disclosure | D-20/D-21: INFO — только счётчики/причины; содержимое — `-debug`; clipboard-контент — никогда (расширение действующего инварианта) |
| Подмена конфига соседним процессом пользователя | Tampering | угроза в рамках L1 принята (однопользовательская сессия); strict-parse ограничивает влияние; last-good предотвращает поломку на битом файле |
| Захват имени ctlsvc/компонента чужим процессом | Spoofing | RequestName-ответ проверяется (PrimaryOwner) — существующий guard engine/conn.go; ctlsvc аналогично |
| DoS через гигантский конфиг/значения капов | DoS | Validate(): потолки (window, backspace cap ≤ разумного max), длины списков MACR |
| Вредоносный YAML (billion-laughs) | DoS | yaml.v3 v3.0.1 содержит фиксы известных DoS (STACK); строгая схема ограничивает глубину |

## Sources

### Primary (HIGH confidence)
- Рабочее дерево, прочитанное этой сессией: engine/engine.go, conn.go, keys.go, types.go, factory.go; internal/session/actor.go; internal/hotkey/fsm.go (+fsm_test.go); internal/correct/{buffer,convert,direction,plan,verify}.go; cmd/goswitchd/main.go; test/e2e/{main,matrix,surface}.go; mise.toml; go.mod; .golangci.yml — все цитаты с номерами строк
- Установленные заголовки цели: /usr/include/ibus-1.0/ibusengine.h:413-421 (delete_surrounding_text), :430 (anchor_pos), ibuskeysyms.h:191 (Control_R=0xffe4)
- Локальные man-страницы wl-clipboard 2.2.1 (dpkg-пакет) — флаги fork/serve, --primary, --trim-newline, --no-newline, --clear
- Прототип владельца /home/nil/.local/share/punto-switcher/punto_engine.py:273-305 (_convert_selection — валидированный на этой машине clipboard round-trip)
- Go module proxy (authoritative registry): fsnotify v1.10.1, godbus v5.2.2, yaml.v3 v3.0.1 — сверено этой сессией
- Module cache: godbus v5.2.2/conn.go:61,137 (SessionBus/ConnectSessionBus), fsnotify@v1.10.1 (API surface)
- ADR-001..005 + d01-experiment-log (BIND), 03-CONTEXT.md (D-22..D-36), REQUIREMENTS.md, STATE.md
- .planning/research/STACK.md, ARCHITECTURE.md (фаза 0) — канонические паттерны hot reload, ctlsvc, e2e-стенд
- `ydotool key --help` на цели (0.1.8) — синтаксис modifier+key

### Secondary (MEDIUM confidence)
- [mankier — wl-clipboard man](https://www.mankier.com/1/wl-clipboard), [github.com/bugaevc/wl-clipboard](https://github.com/bugaevc/wl-clipboard) — кросс-проверка локальных man
- [GTK4 IMContext.get_surrounding_with_selection](https://docs.gtk.org/gtk4//method.IMContext.get_surrounding_with_selection.html), [gtk-text-input.xml](https://gitlab.zrythm.org/zrythm/gtk/-/blob/4.14.1/gtk/gtk-text-input.xml) — GTK доставляет anchor выделения в IM
- [ibus#2423](https://github.com/ibus/ibus/issues/2423) — GTK синхронизирует anchor_pos с выбором (и известные сбои синхронизации)
- [IBus.Engine class docs](https://lazka.github.io/pgi-docs/IBus-1.0/classes/Engine.html) — commit_text/surrounding API

### Tertiary (LOW confidence)
- Поведение Chromium 153 с anchorPos/commit-over-selection — не проверено живьём; закрыто spike'ом фазы (A1/A2)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — все версии сверены по proxy.golang.org/module cache; единственная новая зависимость fsnotify v1.10.1 проверена и описана STACK'ом
- Architecture: HIGH — врезки прочитаны в коде с номерами строк; паттерны продолжают действующие Фазы 1–2
- Pitfalls: HIGH для кодовых/локальных (P1–P4, P8 — из прочитанного кода и man); MEDIUM для live-поведения поверхностей (A1/A2/A5) — каждое закрыто spike-шагом в плане
- Mixed-text semantics: HIGH для механики, MEDIUM для границы D-24/однородный диапазон (Open Question Q1, рекомендация дана)

**Research date:** 2026-09-15
**Valid until:** 2026-10-15 (стабильный домен; live-поведение поверхностей — до первого spike'а фазы)
