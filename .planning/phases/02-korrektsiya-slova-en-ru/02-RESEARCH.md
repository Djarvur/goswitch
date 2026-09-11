# Phase 2: Коррекция слова EN↔RU - Research

**Researched:** 2026-09-11
**Domain:** IBus engine text correction (word-level EN↔RU), buffer/word semantics, capability-ladder replacement, e2e YAML matrix on a live GNOME Wayland session + self-hosted CI runner
**Confidence:** HIGH (in-repo code read line-by-line; target-system environment audited live; external claims limited and cross-checked)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-13:** Слово, уже отделённое пробелом, исправляется двойным тапом (Punto-поведение): буфер помнит последнее слово и после пробела. `ghbdtn <space>` → тап-тап → `привет <space>`. Пробел завершает слово, но не «отправляет» его из буфера. — **Reversibility:** costly
- **D-14:** Диапазон замены = весь «токен» до границы: примыкающая пунктуация входит. `ghbdtn,` → `привет,` одним диапазоном замены. Ключи с буквами определяются по объединению EN+RU слоёв (позиция `;` = RU `ж`, `,` = RU `б` и т.д.).
- **D-15:** Цифры внутри слова входят в диапазон замены и идут через таблицу тождественно; направление определяется только по буквам. `ghbdtn2026` → `привет2026`. Нейтральные символы «смешанным» слово не делают.
- **D-16:** Слово с буквами обеих раскладок (`gfbпривет`) в Фазе 2 не трогаем вовсе — тихий отказ. Порог строгий: любая одна «чужая» буква делает слово смешанным. Нейтральные символы (цифры, общая пунктуация) смешанным слово не считают. — **Reversibility:** reversible
- **D-17:** Матрица v1 покрывает zenity entry (GTK-класс, уже в стенде Фазы 1: self-focused, AT-SPI-witness) + Chromium (обязателен по ADR-003 — подтверждает фактический fallback-уровень). Терминал отложен. — **Reversibility:** reversible
- **D-18:** Набор кейсов v1 — полный словесный: оба направления (`ghbdtn`→`привет`, `привет`→`ghbdtn`), три регистра (`GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет`, `ghbdtn`→`привет`), слово после пробела (D-13), пунктуация/цифры внутри (D-14/D-15), смешанное слово → не тронуто (D-16). YAML-формат кейсов, отчёт PASS/FAIL, код выхода ≠ 0 при падении — по TEST-04.
- **D-19:** Прогон матрицы живёт на self-hosted GNOME CI-раннере, поднимаемом уже в Фазе 2. Требования к раннеру — из STACK: GNOME-сессия, пользователь в группе `input`, python3-gi + gir1.2-atspi. — **Reversibility:** costly
- **D-20:** При молчаливой отмене коррекции (сверка не сошлась / смешанное слово / буфер пуст) пользователь не видит ничего; демон пишет причину отказа в журнал (INFO, причина без содержимого слова).
- **D-21:** Детали УДАЧНЫХ коррекций (исход→результат, уровень лестницы, latency) — только в `-debug`; на INFO — только счётчики.

### Claude's Discretion

- Точная таблица «буквенных клавиш» (объединение EN+RU слоёв по позиции) — сверена с генератором таблиц Фазы 1 (`layouts/`); принцип задан D-14. → Выведена в этом исследовании, см. «Буквенно-ёмкие клавиши».
- Механика «буфер помнит последнее слово после пробела» (хранить одно слово vs хранить всю фразу с указателем) — планировщик выбирает с прицелом на CORR-02 (фразы, Фаза 3). → Рекомендация: фразовый буфер + границы токенов, см. «Pattern 2».
- Порядок задач, разбиение на планы, детали self-hosted раннера (runner-группы, метки, триггер manual-dispatch vs на-PR) — исполнитель.

### Deferred Ideas (OUT OF SCOPE)

- Терминальный класс в e2e-матрице (clipboard-уровень wl-copy/wl-paste + включение конфигом) — Фаза 3/4 (Q3 спеки)
- OSD-уведомление при отказе/переключении — не запрашивалось
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CORR-01 | Двойной Right Shift исправляет последнее слово EN↔RU по позиции клавиш | FSM `Double` уже эмитится актёром; конвейер коррекции: буфер → направление → сверка → лестница (Patterns 1–4); таблицы `layouts/` готовы |
| CORR-04 | Направление определяется автоматически по составу буфера | Классификация рун: ASCII-буква → EN, кириллица → RU, прочее → нейтрально; в Фазе 2 «большинство» вырождается в чистоту скрипта (D-16: любая чужая буква → тихий отказ) |
| CORR-05 | Регистр сохраняется посимвольно | Таблицы несут оба shift-уровня на позицию клавиши (`'G':'П'` и `'g':'п'`) — преобразование = побайтовая lookup-таблица, регистр сохраняется конструктивно |
| CORR-07 | Замена ровно по диапазону без «прыжка» (surrounding text; fallback Backspace×N) | Лестница ADR-003 по битам `SetCapabilities`; арифметика диапазона с учётом «хвоста» после слова (Pitfall 2 — реальный баг прототипа); Chromium-кейс фиксирует фактический уровень |
| CORR-09 | Буфер очищается по Enter, Tab, Escape, смене фокуса | Все триггеры наблюдаемы: keyvals 0xff0d/0xff09/0xff1b + `LifecycleFocusOut`/`LifecycleReset`; клик-скачок курсора — best-effort через surrounding text (ADR-004) |
| TEST-04 | YAML-матрица «ввод → ожидание», отчёт PASS/FAIL, код выхода ≠ 0 | Шаговый YAML-формат поверх стенда Фазы 1 (`test/e2e`); AT-SPI readback + stdout-оракул zenity; self-hosted runner (D-19) |
</phase_requirements>

## Project Constraints (from AGENTS.md / CONVENTIONS.md — authoritative)

No `CLAUDE.md` exists; workspace `AGENTS.md` (GSD-generated from PROJECT/STACK/CONVENTIONS) is authoritative. Directives binding this phase (owner, 2026-09-10):

1. **Строгий TDD** — red → green → refactor; Verify-блоки задач прогоняют тесты на каждом шаге (D-07).
2. **Зелёная итерация** — каждая задача завершается только при `go build ./...` + `go test -race ./...` + `golangci-lint run` без ошибок (D-08).
3. **Сценарии — mise, не make**; mise управляет инструментами (`mise.toml [tools]`: go 1.23, golangci-lint 2.13.2 установлены per-project — проверено) (D-09).
4. **golangci-lint максимально строгий** с Фазы 1 (`.golangci.yml` v2, `default: all`, depguard deny `unsafe`) — новые пакеты обязаны проходить (D-10).
5. **GitHub Actions актуальны**: e2e-workflow Фазы 2 добавляется к `pr-sanity.yml`/`security-scheduled.yml` (D-11/D-19).
6. **Go-работа — по скиллу go-ultimate** (project skill, `.zcode/skills/go-ultimate/`): no `pkg/`, `internal/` для приватного, interfaces at point of use, `errors.Is/As`, `slices`/`maps`, vanilla `testing` + `package xxx_test`, без мутабельных глобалов; project-file precedence — AGENTS.md выше скилла.
7. **Приватность логов**: INFO — метаданные/счётчики; keystroke-трассировка — только `-debug` (D-20/D-21, INST-04).
8. Git workflow: фича-ветки; работа уже на `gsd/phase-02-korrektsiya-slova-en-ru`.

## Summary

Фаза 2 — это конвейер из четырёх чистых блоков (буфер-слов → детектор направления → сверки → выбор уровня лестницы) плюс два проводных изменения в существующем коде: `ProcessKeyEvent` перестаёт быть всегда-`false` (в RU-режиме внутреннего флипа движок потребляет клавиши и коммитит кириллицу), и эмиттеры `DeleteSurroundingText`/`ForwardKeyEvent`/`RequireSurroundingText` добавляются в `engine/` симметрично существующему `CommitText`. Вся спецификация поведения уже зафиксирована: ADR-002 (семантика тапов, исполнимый корпус `fsm_test.go`), ADR-003 (лестница, «abort, не мусорить»), ADR-004 (триггеры сброса + обязательная сверка), D-13..D-21 (границы слова, смешанное слово, матрица, логи). Реализация — преимущественно `internal/correct` (чистая логика) + проводка в `internal/session` + e2e-матрица в `test/e2e`.

**Ключевая находка №1 — RU→EN требует внутреннего флипа уже в Фазе 2.** Кейсы `привет`→`ghbdtn` (D-18) и смешанное слово `gfbпривет` (D-16) выполнимы, только если кириллица попадает в поле ввода ЧЕРЕЗ движок (в режиме RU движок потребляет латинский keyval и коммитит кириллический символ по таблице — паттерн прототипа владельца, проверенный на этой машине). Все внешние пути закрыты живыми пробами D-01: `SetGlobalEngine` не двигает XKB-группу (ввод остаётся латиницей), gsettings-записи игнорируются рантаймом, а системный xkb:ru источник уводит фокус с движка (сброс буфера по CORR-09). ADR-001 явно назначает это Фазе 2: «Переключение „на лету“ скрипта выдачи делает Фаза 2 через таблицы layouts (CORR-01..08), XKB-группа не участвует». Открыт один вопрос-гейт: чем переключать режим в рантайме (рекомендация — проводка FSM `Single`→флип; альтернатива — debug-сеттер; см. Open Questions №1).

**Ключевая находка №2 — «хвост» после слова.** D-13 требует исправлять слово, уже отделённое пробелом, но `DeleteSurroundingText(-n, n)`/Backspace×N без учёта разделителя портят текст: для `ghbdtn␠` курсор стоит после пробела, и удаление 6 рун снимает `␠hbdtn`, оставляя `g` → `gпривет`. Это реальный баг прототипа владельца (`punto_engine.py:186-199` — `delete_word` считает `n = len(word)` без хвоста; ручная коррекция после пробела в прототипе портила текст). goswitch обязан считать диапазон = токен + хвост: для лестницы уровня 1 удалять ровно токен (`offset = -(len(tail)+len(token))`, `nchars = len(token)`), для уровня 2 удалять токен+хвост и перекоммитить `исправленный_токен + хвост`.

**Ключевая находка №3 — окружение готово.** zenity 4.0.1 (GTK4), google-chrome 153 (deb) + chromium 152 (snap), gnome-text-editor 46.3, ydotool 0.1.8, python3-gi+atspi, `/dev/uinput` (группа `input`), mise-инструменты — всё установлено и проверено живьём; отсутствует только GitHub Actions runner (его Фаза 2 и поднимает, D-19). Прототип punto-switcher-daemon замаскирован — конфликтов нет.

**Primary recommendation:** построить `internal/correct` как чистый пакет (буфер-фраза + токены, направление, конверсия, сверка, арифметика лестницы) с TDD-корпусом; провести флип-режим через актёра с проводкой `Single`→флип (подтвердить на plan-review); расширить `engine/` тремя эмиттерами; надстроить шаговую YAML-матрицу над стендом Фазы 1; поднять self-hosted runner на этой машине с триггером `workflow_dispatch` и concurrency-группой.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Буфер набора, токены, границы слова | `internal/correct` (чистая логика, без D-Bus) | `internal/session` (проводка) | Headless-TDD (D-07), ноль зависимостей от транспорта |
| Детектор направления + конверсия регистров | `internal/correct` + `layouts/` (готовые таблицы) | — | Чистые функции над таблицами; корпус слов — golden |
| Режим скрипта (внутренний флип EN/RU) | `internal/session` (актёр: состояние + решение) | `engine/` (коммит) | ADR-001: режим — внутреннее состояние демона |
| Исполнение замены (лестница уровней) | `engine/` (эмиттеры сигналов) | `internal/session` (выбор уровня по caps) | Транспортная обязанность адаптера; решение — логика |
| Сверка с surrounding text | `internal/session` (ожидание/сравнение) | `engine/` (приём `SetSurroundingText`, эмит `RequireSurroundingText`) | ADR-004: обязательный протокол до коррекции |
| Триггеры сброса буфера | `engine/` (keyvals/lifecycle → актёр) | — | Все триггеры наблюдаемы на уровне IME (ADR-004) |
| e2e-матрица TEST-04 | `test/e2e` (Go-раннер) | `focus_helper.py` (AT-SPI readback) | Надстройка над стендом Фазы 1 |
| Self-hosted GNOME runner (D-19) | `.github/workflows/e2e-*.yml` + actions-runner (systemd user unit в сессии) | `mise` tasks | CI-контур; локальный прогон = те же mise-задачи |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/godbus/dbus/v5` | v5.2.2 (уже в go.mod) | Эмиттеры сигналов движка (`DeleteSurroundingText`, `ForwardKeyEvent`, `RequireSurroundingText`) симметрично `CommitText` | Единственная зависимость демона; `conn.Emit(path, iface.SIGNAL, args...)` — тот же вызов, что в `Engine.CommitText` [VERIFIED: engine/engine.go:314] |
| `gopkg.in/yaml.v3` | **v3.0.1** (проверено через proxy.golang.org живьём в этой сессии: `{"Version":"v3.0.1","Time":"2022-05-27T08:05:30Z"}`) | YAML-кейсы e2e-матрицы (TEST-04) | Стандартный YAML 1.2-парсер, ноль транзитивных зависимостей, заморожен (аудит-поверхность минимальна); уже в STACK как конфиг-парсер Фазы 3 — матрица начинает его использование |
| In-repo `internal/correct` | новый пакет | Буфер/токены/направление/сверка/лестница-арифметика | Чистая логика по go-ultimate: D-Bus-free, полный unit-корпус |
| In-repo `internal/hotkey` + `internal/session` | Фаза 1 | FSM тапов + актёр | `Double` уже эмитится: `slog.Info("action", "n", int(action))` [VERIFIED: internal/session/actor.go:98] — Фаза 2 подключает обработчик вместо/вместе с логом |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| ydotool | 0.1.8-3build1 (dpkg, проверено) | Инжекция физического пути в матрице | Уже используется стендом; для кейса «клик-скачок курсора» — `ydotool move`/`click` (префлайт-самотест добавить) |
| AT-SPI helper (`test/e2e/focus_helper.py`) | Фаза 1 | witness + `text <app>` readback | `cmd_text` уже реализован как «bridge to Phase 2 TEST-04» [VERIFIED: test/e2e/focus_helper.py:11-12] |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `gopkg.in/yaml.v3` (кейсы матрицы) | JSON + stdlib | Ноль новых зависимостей, но TEST-04/SPEC §7.2 явно требуют YAML; владелец читает/правит кейсы руками |
| Проводка `Single`→флип в Фазе 2 | debug/CLI-сеттер режима | Сеттер не трогает трассируемость SWCH-01 (Фаза 3), но создаёт одноразовый тестовый канал, который Фаза 3 снесёт; см. Open Questions №1 |

**Installation:**

```bash
go get gopkg.in/yaml.v3@v3.0.1   # единственное новое зависимости; go mod tidy в гейте итерации
```

**Version verification:** `go list -m -versions gopkg.in/yaml.v3` → v3.0.1 (последняя); proxy.golang.org `@latest` → v3.0.1 — проверено в этой сессии.

## Package Legitimacy Audit

Go-модули проверяются через proxy.golang.org (авторитетный реестр Go; npm/pypi- seam неприменим). Никаких кандидатов из WebSearch/тренировочных данных в фазе нет — оба модуля уже присутствуют в STACK (утверждённый research Фазы 0, проверенный через официальный прокси).

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| `gopkg.in/yaml.v3` | proxy.golang.org | v3.0.1 от 2022-05-27 (заморожен) | де-факто стандарт экосистемы | github.com/go-yaml/yaml | OK | Approved |
| `github.com/godbus/dbus/v5` | proxy.golang.org | v5.2.2 (2025-12-29, из STACK) | ~1.2k stars, продакшн ibus-bamboo | github.com/godbus/dbus | OK | Approved (уже в go.mod) |

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### System Architecture Diagram

```
ydotool (физический путь, e2e) / живая клавиатура
        │ evdev → compositor (XKB group 'us' неизменен, ADR-001)
        ▼
   ibus-daemon ── ProcessKeyEvent(keyval,keycode,state) ──► engine.Engine (go/адаптер)
        │                                                       │ decode → EngineEvent
        │                                                       ▼
        │                                            internal/session.Actor (мьютекс-сериализация)
        │                                            ├─ hotkey.FSM (тапы; Double=слово, Single=флип*)
        │                                            ├─ correct.Buffer (фраза + токены; сбросы)
        │                                            └─ correct.Plan (направление → сверка → уровень)
        │                                                       │
        │                                                       ▼ эмиттеры engine/
        │      caps&31? ── RequireSurroundingText ─► (асинхр. ожидание SetSurroundingText ≤ ~100мс)
        │            │ сверка: буферный токен == суффикс текста перед курсором?
        │            │ нет ──► тихий отказ (D-20, INFO-причина)
        │            ▼ да
        │      уровень 1: DeleteSurroundingText(-(tail+n), n) → CommitText(fixed) → verify-after
        │      уровень 2: ForwardKeyEvent(BackSpace)×(n+tail) → CommitText(fixed + tail)
        ▼                                                       │
   клиент (GTK4/GTK3/Chromium) применяет удаление/коммит ◄─────┘
        ▲
        │  SetCapabilities / SetSurroundingText / FocusIn/Out/Reset — обратно в движок
   e2e-стенд: focus_helper.py (AT-SPI readback) + лог демона (-debug: уровень лестницы) → PASS/FAIL
```

\* проводка `Single`→флип — рекомендация, см. Open Questions №1.

### Recommended Project Structure

```
internal/correct/          # НОВЫЙ чистый пакет (нет D-Bus, нет времени, нет горутин)
├── buffer.go              # фразовый буфер: push-символ, backspace-pop, границы, hard-reset, token()
├── direction.go           # классификация рун + правило D-16 (тихий отказ)
├── convert.go             # таблицы layouts/ → конверсия с сохранением регистра
├── verify.go              # свера буферного токена с surrounding text (суффиксное правило)
├── plan.go                # выбор уровня лестницы по caps + арифметика диапазона (offset/nchars/N)
└── *_test.go              # корпусы: слова обоих направлений, регистры, цифры/пунктуация, смешанные, сбросы
internal/session/
├── actor.go               # + режим скрипта (флип), состояние коррекции (двухфазность), проводка Double/Single
└── actor_test.go          # + интеграция с fake-эмиттером
engine/
├── engine.go              # + эмиттеры DeleteSurroundingText/ForwardKeyEvent/RequireSurroundingText
│                          #   ProcessKeyEvent: возврат true в RU-режиме для потреблённых клавиш
└── types.go, keys.go      # без изменений (CapSurroundingText и keyvals уже есть)
test/e2e/
├── matrix.go              # НОВОЕ: загрузка YAML, прогон шагов, отчёт PASS/FAIL, exit≠0
├── surface.go             # НОВОЕ: драйверы поверхностей (zenity / chromium / gnome-text-editor*)
├── cases/                 # НОВОЕ: YAML-кейсы матрицы v1 (D-18)
├── fixtures/input.html    # НОВОЕ: страница с <input> для Chromium
└── main.go, focus_helper.py, preflight.go   # расширить: -matrix флаг, mouse-самотест, chromium-префлайт
.github/workflows/e2e-matrix.yml   # НОВОЕ: self-hosted runner (D-19)
docs/                      # runner-инструкция (установка actions-runner в сессию)
```

### Pattern 1: Лестница замены с учётом «хвоста» (ADR-003 + D-13/D-14)

**What:** замена = удалить диапазон неверного текста, закоммитить исправленный. Диапазон = токен (буквы+пунктуация+цифры по D-14/D-15); между концом токена и курсором могут стоять boundary-символы (хвост — пробел и т.п.).
**When to use:** каждая коррекция; уровень по битам caps ТЕКУЩЕГО input context.

```go
// Source: ADR-003 + вывод этого исследования (арифметика хвоста)
// n = len([]rune(token)); tail = len([]rune(boundaryTail)) — ВСЁ В РУНАХ

// Уровень 1 (caps & CapSurroundingText): удалить ровно токен, хвост не трогать
engine.DeleteSurroundingText(-(int32(tail + n)), uint32(n))
engine.CommitText(engine.NewIBusText(converted))
// verify-after: RequireSurroundingText + сравнение позиции/текста (сигнал без ack на 1.5.29)

// Уровень 2 (без бита): удалить токен+хвост, вернуть хвост в коммите
for i := 0; i < n+tail; i++ {
    engine.ForwardKeyEvent(engine.KeyBackSpace, 14, 0) // keyval 0xff08, keycode 14
}
engine.CommitText(engine.NewIBusText(converted + boundaryTail))
```

[VERIFIED: engine/keys.go:25 `KeyBackSpace = 0xff08`; engine/keys.go:18 `CapSurroundingText = 1 << 5`; engine/engine.go:310-317 CommitText-эмиттер как образец]

### Pattern 2: Фразовый буфер с границами токенов (дискреция CONTEXT)

**What:** буфер хранит всё введённое с последнего hard-reset как рунную строку + метки границ; «последнее слово» = текущий токен, а если он пуст — предыдущий завершённый (D-13). Backspace выталкивает последнюю руну (естественная модель). Фаза 3 (CORR-02 «фраза») переиспользует тот же буфер без переделки.
**When to use:** единственный источник «что набрал пользователь»; сверка (ADR-004) сравнивает токен с текстом перед курсором.

```go
// Source: D-13/D-14/D-15 + дискреция CONTEXT (выбор в пользу CORR-02)
type Buffer struct {
    runes []rune          // всё с последнего hard-reset
    tokenStart int        // индекс начала текущего токена
    lastWordLen int       // длина предыдущего завершённого токена (для D-13)
}
// push(char): letterCapable(char)||digit(char) → в токен; иначе граница:
//   завершить токен (lastWordLen = длина, если >0), символ — в хвост.
// backspace(): вытолкнуть последнюю руну (и токен, и хвост, и прошлые слова — честно).
// hardReset(): полностью очистить (Enter/Tab/Escape/FocusOut/Reset — CORR-09).
```

### Pattern 3: Внутренний флип скрипта (ADR-001 Option B, паттерн прототипа)

**What:** режим EN/RU — состояние демона. В EN-режиме всё идёт транзитом (`ProcessKeyEvent → false`, символ вставляет клиент через XKB us). В RU-режиме движок потребляет печатную клавишу и коммитит кириллицу по таблице; буфер в обоих режимах записывает символ, реально появившийся в поле (script-true представление — из него работает CORR-04).

```go
// Source: прототип владельца punto_engine.py:394-414 (проверен на этой машине), адаптировано
if mode == RU && printable && mods&^MaskShift == 0 {
    if ru, ok := layouts.ENToRU[ch]; ok && ru != ch {
        e.CommitText(engine.NewIBusText(string(ru))) // 'g'→'п', '['→'х', '@'→'"'
        return true                                   // потреблено: клиент не вставит латиницу
    }
    return false // тождественная карта ('(', цифры) и непокрытое — транзитом
}
// ключ с шифтом уже приходит XKB-переведённым: 'G'→'П', '@'→'"' — таблица несёт оба уровня
```

[VERIFIED: layouts/tables.go:84 `'g': 'п'`, :46 `'G': 'П'`, :72 `['': 'х'`, :45 `'@': '"'`]

### Pattern 4: Сверка перед коррекцией (ADR-004) — двухфазная, без блокировки ответа

**What:** перед заменой — `RequireSurroundingText`, ограниченное ожидание свежего `SetSurroundingText` (актор: таймер-таймаут ~100 мс, продолжение под мьютексом), затем суффиксное сравнение: текст перед курсором обязан ЗАКАНЧИВАТЬСЯ токеном (хвост допустим после). Несовпадение/таймаут → тихий отказ (D-20). Без caps-бита сверка невозможна — доверие буферу + явные триггеры сброса (задокументированная деградация ADR-004).
**When to use:** любая коррекция на caps-клиенте. `ProcessKeyEvent` всегда отвечает мгновенно — ожидания живут в таймерах актёра, не в обработчике (тайминговое требование IBus).

### Буквенно-ёмкие клавиши (дискреция CONTEXT — сверено с генератором)

Правило D-14: клавиша «буквенно-ёмкая», если хотя бы один из слоёв (EN или RU) даёт на её позиции букву. Выведено из сгенерированной таблицы [VERIFIED: layouts/tables.go:13-108] — дословные пары:

- Буквы обеих раскладок: `a-z`, `A-Z` (напр. `'a': 'ф'`, `'Z': 'Я'`).
- Ёмкие по RU-стороне: `` ` ``→`ё`, `~`→`Ё`, `'`→`э`, `"`→`Э`, `;`→`ж`, `:`→`Ж`, `[`→`х`, `{`→`Х`, `]`→`ъ`, `}`→`Ъ`, `,`→`б`, `<`→`Б`, `.`→`ю`, `>`→`Ю` (tables.go:77, 16, 40, 39, 72, 105, 74, 106, 25, 38, 27, 43).
- НЕ ёмкие (обе стороны — символы; = граница токена): `/`→`.` (tables.go:28), `?`→`,` ( :44), `\`, `|`→`/` ( :105), `-`, `=`, `_`, `+`, `!`, `(`, `)`, `*`, `&`, `@`→`"` ( :45 — кавычка, не буква), `#`→`№` (№ — символ), `$`→`;`, `^`→`:`, `%`, пробел, Tab.
- Цифры `0-9`: НЕ ёмкие, но НЕ граница — нейтральный символ внутри токена (D-15: `'2': '2'`, таблица тождественна).
- Буферное правило: токен = максимальный запуск {буквенно-ёмкие} ∪ {цифры}; направление — только по буквам; букв нет → тихий отказ.

### Pattern 5: Шаговый YAML-кейс матрицы (TEST-04 / D-18)

```yaml
# Source: TEST-04 + D-18 (полный словесный набор)
name: word-after-space-en-ru
surface: zenity          # zenity | chromium (| gnome-text-editor — см. Open Questions №2)
mode: en                 # стартовый режим движка (внутренний флип)
steps:
  - {type: "ghbdtn"}     # ydotool type — физический путь
  - {key: space}
  - {tap: double}        # двойной Right Shift; стенд ждёт решения в логе перед следующим шагом
expect_text: "привет "   # AT-SPI readback (или stdout-оракул zenity при закрытии)
---
name: reset-escape-noop
surface: chromium
mode: en
steps:
  - {type: "ghbdtn"}
  - {key: Escape}
  - {tap: double}
expect_text: "ghbdtn"    # CORR-09: Escape очистил буфер — коррекции нет
```

Смешанное слово (D-16): `steps: [{type: "gfb"}, {tap: single}, {type: "ghbdtn"}]` → ожидание «не тронуто» (`gfbпривет` остаётся). После `{tap: single}` стенд ОБЯЗАН дождаться `{"msg":"action","n":1}` в логе (флип происходит на истечении окна, не мгновенно) — иначе RU-часть наберётся в EN-режиме.

### Pattern 6: Self-hosted GNOME runner (D-19)

- actions-runner ставится на ЭТУ машину (единственная GNOME-цель; владелец осознанно выбрал D-19), ярлык `gnome`, запуск systemd **user**-юнитом внутри графической сессии (`After=graphical-session.target`; юниту нужны `WAYLAND_DISPLAY`, `XDG_RUNTIME_DIR`, `DBUS_SESSION_BUS_ADDRESS` — импорт окружения сессии).
- Workflow `e2e-matrix.yml`: `runs-on: [self-hosted, gnome]`, триггер `workflow_dispatch` (SPEC §8: e2e — manual dispatch; также безопасность — см. Security Domain), `concurrency: {group: e2e-gnome}` — один прогон; два стенда одновременно печатали бы в чужие окна.
- Шаги: checkout → `mise install` → `mise run e2e-matrix` → upload отчёта (artifact). Локальный и CI-прогон — одни и те же mise-задачи (D-09, прецедент `pr-sanity.yml`).

### Anti-Patterns to Avoid

- **Считать Backspace×N в байтах:** кириллица = 2 байта UTF-8; N — всегда `len([]rune(...))` (ADR-003: «подсчёт рун — часть контракта буфера»).
- **Игнорировать хвост после слова:** коррекция после пробела без учёта хвоста даёт `gпривет` (реальный баг прототипа, Pitfall 2).
- **Блокировать ответ `ProcessKeyEvent`** ожиданием сверки/таймеров — тайминговое требование IBus: ответ за единицы мс; ожидания — таймерами в актёре (прецедент Фазы 1: решение FSM только на истечении окна).
- **Словари/автоисправления:** SPEC §11 — только по хоткею; CORR-04 — композиция буфера, без hunspell (словари прототипа — для auto-режима, не нашего).
- **Глобальные предположения о клиентах:** caps-карта — per input context (ADR-003: «движок не предполагает, что все окна одинаковы»); уровень выбирается на каждую коррекцию.
- **Логировать содержимое слов на INFO:** D-20/D-21 — только причина отказа (без слова) на INFO; исход→результат — `-debug`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| YAML-кейсы | Свой парсер/JSON | `gopkg.in/yaml.v3` (+ `decoder.KnownFields(true)` — строгая схема) | TEST-04 требует YAML; строгий декодер ловит опечатки в кейсах |
| Классификация «буква/кириллица» | Регулярки/свои диапазоны | `unicode.IsLetter` + проверка Latin/Cyrillic-блоков (`unicode.Is(unicode.Latin, r)` / `unicode.Is(unicode.Cyrillic, r)`) | Ё/Ё и прочие края; stdlib уже даёт корректные таблицы |
| Подсчёт рун | `len(string)` | `[]rune` / `utf8.RuneCountInString` | Кириллица мультибайтна — главный источник «ровно в 2 раза больше Backspace» |
| Таблицы раскладок | Ручные карты | `layouts.ENToRU` / `layouts.RUToEN` (готовы, golden-пиннуты) | CORR-08 завершён в Фазе 1; обе таблицы покрывают знаки и оба регистра |
| AT-SPI чтение | Новый Go-клиент a11y | `focus_helper.py text <app>` (готов) | python3-gi предписан SPEC §7.3; `cmd_text` уже «bridge to Phase 2» |
| Физическая инжекция | uinput-инжектор | ydotool 0.1.8 (проверен префлайтом) | Уже в стенде; собственный инжектор — только как fallback (STACK) |

**Key insight:** вся «инженерия» фазы — в чистой логике буфера/направления/диапазонов; транспорт (эмиттеры), инжекция и чтение — готовые блоки Фазы 1.

## Common Pitfalls

### Pitfall 1: Курсор после хвоста ломает арифметику удаления
**What goes wrong:** `ghbdtn␠` + тап-тап → в поле остаётся `gпривет` (удаление сняло `␠n,t,d,b,h` вместо слова).
**Why it happens:** DeleteSurroundingText/Backspace работают ОТ КУРСОРА; между курсором и словом стоит boundary-символ. Прототип владельца содержит ровно этот баг: `delete_word` считает `n = len(word)` без хвоста (`punto_engine.py:186-199`); спасало его лишь то, что авто-путь срабатывал до пересылки разделителя.
**How to avoid:** диапазон = токен + хвост; уровень 1 — `offset=-(tail+n), nchars=n`; уровень 2 — `N=(tail+n)` Backspace и перекоммит хвоста (Pattern 1).
**Warning signs:** e2e-кейс «слово после пробела» (D-13) — первый упавший.

### Pitfall 2: Байты вместо рун
**What goes wrong:** `привет` (12 байт) удаляется 12 Backspace — стирается вдвое больше текста.
**Why:** `len(str)` в Go — байты.
**How to avoid:** буфер хранит `[]rune` (ADR-003 следствие: «буфер хранит руны, не байты»); юнит-тест на кириллическом слове фиксирует.
**Warning signs:** RU→EN-кейсы падают «ровно вдвое» агрессивнее.

### Pitfall 3: Chromium не применяет delete-surrounding (caps-бит может «врать»)
**What goes wrong:** уровень 1 выбран по caps-биту, клиент молча игнорирует удаление → после CommitText в поле `ghbdtn привет` (дубль).
**Why:** [CITED: github.com/ibus/ibus/issues/2354 — «Chromium-based apps: 'delete surrounding text' doesn't work»] — подтверждено поиском в этой сессии; ожидалось исследованием Фазы 0.
**How to avoid:** e2e фиксирует ФАКТИЧЕСКИЙ уровень (ADR-003: «кейс фиксирует, какой уровень реально сработал»); verify-after после удаления обязателен (log+счётчик при расхождении, без авто-ремонта — остаточный риск по ADR-003).
**Warning signs:** Chromium-кейс матрицы показывает дубль текста.

### Pitfall 4: Ожидание сверки в обработчике D-Bus
**What goes wrong:** `ProcessKeyEvent` отвечает за секунды → ibus-daemon считает движок зависшим.
**Why:** сверка — асинхронный раунд (Require → SetSurroundingText).
**How to avoid:** двухфазная коррекция в актёре (таймер-таймаут ~100 мс, продолжение по приходу SetSurroundingText); обработчик всегда отвечает мгновенно.
**Warning signs:** задержки ввода всего стола при коррекции.

### Pitfall 5: RU-часть кейса набрана до флипа
**What goes wrong:** стенд шлёт `{tap: single}` и сразу печатает «привет»-клавиши — флип ещё не произошёл (решение Single — на истечении 300 мс окна, ADR-002), буквы уходят транзитом латиницей.
**How to avoid:** после каждого `{tap}` — `waitForLog({"msg":"action","n":K})` перед следующим шагом (прецедент: m1-gate ждёт `"n":2`).
**Warning signs:** смешанный кейс даёт `gfbghbdtn` вместо `gfbпривет`.

### Pitfall 6: zenity закрывается от Enter/Escape
**What goes wrong:** кейс «сброс по Enter» на zenity не может сделать тап-тап — диалог закрыт.
**How to avoid:** сброс-кейсы по классам: Enter/Escape на zenity проверяются ОРАКУЛОМ закрытия (stdout == набранное без коррекции), полный цикл «сброс → тап-тап → без изменений» — на Chromium (поле переживает Enter/Escape); смена фокуса — две поверхности (Pattern 5).
**Warning signs:** кейс падает на «окно исчезло».

### Pitfall 7: Buffer desync на no-cap клиентах (клик-скачок)
**What goes wrong:** пользователь кликнул в середину слова (движок кликов не видит), тап-тап → Backspace×N стирает чужой текст.
**Why:** без caps сверка невозможна — деградация, зафиксированная ADR-004.
**How to avoid:** на no-cap-клиентах доверять буферу только при отсутствии явных триггеров сброса; Backspace-клавиша выталкивает руну из буфера; Ctrl-модифицированные комбинации не кормят буфер (прецедент прототипа). Остаточный риск задокументировать.
**Warning signs:** юнит-корпус «backspace/модификаторы» не покрывает сценарий.

### Pitfall 8: Оркестратор матрицы забыл, что он печатает на живом столе
**What goes wrong:** два прогона стенда параллельно инжектируют в чужие окна; CI-раннер стартует матрицу, пока владелец печатает.
**How to avoid:** `concurrency`-группа workflow; матрица открывает СВОИ поверхности (прецедент: zenity self-focus + AT-SPI-witness перед инжекцией); префлайт остаётся fail-fast.
**Warning signs:** «flaky» падения только при ручной работе за машиной.

## Code Examples

### Классификация направления (CORR-04 + D-16)

```go
// Source: CORR-04/D-16 + REQUIREMENTS — чистая функция, корпус в unit
func Direction(token []rune) (dir Direction, ok bool) {
    en, ru := 0, 0
    for _, r := range token {
        switch {
        case unicode.Is(unicode.Latin, r):
            en++
        case unicode.Is(unicode.Cyrillic, r):
            ru++
        } // цифры/пунктуация — нейтральны (D-15/D-16)
    }
    switch {
    case en > 0 && ru > 0:
        return 0, false // смешанное — тихий отказ (D-16), Фаза 3 — CORR-06
    case en > 0:
        return ENtoRU, true // «букв какой раскладки больше» при чистом скрипте = 100%
    case ru > 0:
        return RUtoEN, true
    }
    return 0, false // букв нет — направление не определено, отказ
}
```

### Суффиксное правило сверки (ADR-004)

```go
// Source: ADR-004 decision п.1 + D-13/D-14
// textBeforeCursor обязан ЗАКАНЧИВАТЬСЯ токеном (после токена — только хвост boundary).
func MatchesSuffix(textBeforeCursor []rune, token []rune) bool {
    n := len(token)
    if len(textBeforeCursor) < n {
        return false
    }
    return slices.Equal(textBeforeCursor[len(textBeforeCursor)-n:], token)
}
```

### Эмиттеры уровня лестницы (симметрично CommitText)

```go
// Source: engine/engine.go:310-317 (CommitText) — тот же образец; сигнатуры из
// 01-RESEARCH.md Pattern 1 (живая интроспекция IBus 1.5.29)
func (e *Engine) DeleteSurroundingText(offset int32, nchars uint32) {
    if e.conn == nil { return }
    if err := e.conn.Emit(e.path, ifaceEngine+".DeleteSurroundingText", offset, nchars); err != nil {
        slog.Error("delete surrounding emit failed", "error", err)
    }
}
func (e *Engine) ForwardKeyEvent(keyval, keycode, state uint32) { /* ... ifaceEngine+".ForwardKeyEvent" ... */ }
func (e *Engine) RequireSurroundingText() { /* ... ifaceEngine+".RequireSurroundingText" ... */ }
```

### Триггеры сброса — уже определённые константы

```go
// Source: [VERIFIED: engine/keys.go:25-32 — дословно]
//	KeyBackSpace = 0xff08
//	KeyTab       = 0xff09
//	KeyReturn    = 0xff0d
//	KeyEscape    = 0xff1b
//	KeyShiftL    = 0xffe1
//	KeyShiftR    = 0xffe2
//	KeyControlL  = 0xffe3
//	KeySpace     = 0x020
// Hard reset: KeyReturn|KeyTab|KeyEscape (+ KP_Enter 0xff8b — keypad-вариант Enter) + LifecycleFocusOut|LifecycleReset
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `delete_surrounding_text` с bool-ack (GI-биндинги) | Сигнал fire-and-forget на IBus 1.5.29 | 1.5.x (verified в исследовании Фазы 1 по ibusengine.c) | verify-after обязателен; «bool» старых биндингов незначим |
| uinput-инъекция текста (WaylandSwitcher) | IBus `commit_text` + лестница | Аксиома SPEC §3.1 | нет конфликтов с keyd/xremap; e2e-инжекция остаётся uinput (симулирует человека) |
| Переключение раскладки внешними командами | Внутренний флип (ADR-001, 2026-09-10) | D-01 эксперимент | XKB-группа сессии не меняется; RU-текст появляется только через коммиты движка |

**Deprecated/outdated:**
- Двухдвижковая модель активации (two-engine) как механизм ПЕРЕКЛЮЧЕНИЯ — убита экспериментом D-01 (все пробы not-switched); регистрация двух движков из Фазы 1 остаётся безвредной до консолидации в Фазе 3.
- Лор «`gsettings current` переключает раскладку» — опровергнуто живьём (keyboard.js 46 игнорирует внешние записи).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | zenity 4.0.1 (GTK4-on-Wayland, text-input-v3 через mutter) применяет DeleteSurroundingText | Pitfalls/Matrix | Уровень 1 не работает на GTK-классе — вся матрица падает на уровень 2 (Backspace); лестница вырождается, ADR-003 пересматривать не нужно (уровень 2 самодостаточен) |
| A2 | Проводка FSM `Single`→флип входит в Фазу 2 (рекомендация; подтверждение — plan-review) | Pattern 3, Open Q1 | Если нет — RU→EN и смешанные кейсы требуют debug-сеттера режима (одноразовый канал); объём сопоставим, трассируемость SWCH-01 чище |
| A3 | google-chrome (deb, 153.x) — Chromium-поверхность матрицы (snap 152 — альтернатива) | Matrix | Snap-конфайнмент мог бы ломать AT-SPI/uinput; deb проще. Автообновление Chrome — дрейф версии на раннере |
| A4 | Порядок ForwardKeyEvent-бёрст → CommitText сохраняется ibus-daemon для всех клиентов | Pattern 1 | Для GTK доказано прототипом; для Chromium — пока нет (матрица докажет); при нарушении — pacing между forwards |
| A5 | Эта машина (T2, живая сессия) становится self-hosted раннером | Pattern 6 | Если владелец предпочтёт ВМ — воспроизвести требования STACK (GNOME-сессия, input-группа, python3-gi, gir1.2-atspi) |
| A6 | Буфер хранит реально появившиеся символы (script-true), не физические keyval | Pattern 2 | Альтернатива (физические клавиши + тег режима) ломает CORR-04 «по составу буфера» и D-16 — потребует решения владельца |
| A7 | Матрица v1 = zenity + Chromium (D-17); упоминание gnome-text-editor в критерии 1 — наследие ROADMAP-формулировки | Open Q2 | Если владелец хочет буквального gnome-text-editor — третий драйвер поверхности (установлен: 46.3; усилие малое) |
| A8 | `ydotool move/click` (0.1.8) работает для кейса «клик-скачок курсора» | Pitfalls | Нет — кейс через скачок другим способом (AT-SPI недоступен для установки курсора; тогда кейс заменить на «сверка не сошлась» юнит-покрытием + e2e без клика) |

## Open Questions

1. **Триггер внутреннего флипа в Фазе 2** (главный вопрос планирования)
   - What we know: RU→EN и смешанные кейсы требуют режима RU у движка; ADR-001: «Переключение „на лету“ скрипта выдачи делает Фаза 2». FSM `Single` уже эмитится; SWCH-01 формально в Фазе 3; ADR-002 уже зафиксировал тайминги (окно 300 мс), т.е. SWCH-04 не открывается.
   - What's unclear: тянет ли Фаза 2 проводку `Single`→флип (следствие ADR-001) или строго хранит SWCH-01 нетронутым до Фазы 3.
   - Recommendation: провести `Single`→флип в Фазе 2 (естественно для e2e через физический путь; владелец получает dogfood-able демон; SWCH-01 Фазы 3 остаётся за «одним хозяином» D-03 — консолидация источников, UX-доки). Обсудить на plan-review.
2. **gnome-text-editor в матрице v1?**
   - What we know: критерий успеха 1 фазы дословно называет gnome-text-editor; D-17锁定 матрицу как zenity+Chromium. gnome-text-editor 46.3 установлен; GtkSourceView — ДРУГОЙ GTK-виджет-класс (multiline), потенциально другая caps-карта — реальная матричная ценность.
   - Recommendation: добавить как третий (дешёвый) драйвер поверхности, если plan-review не возражает; минимум — подтвердить у владельца, что D-17 заменяет формулировку критерия.
3. **Судьба буфера после удачной коррекции**
   - What we know: не специфицировано. Варианты: заменить токен в буфере на исправленный (сверка остаётся зелёной — повторный тап-тап даёт toggle-поведение) или очистить.
   - Recommendation: заменить (бесплатный toggle — знакомое по Punto поведение; инвариант «буфер == текст перед курсором» сохраняется). Зафиксировать в плане одним правилом.
4. **Таймаут ожидания SetSurroundingText после RequireSurroundingText**
   - What we know: бюджет действия Right Shift ≤ 300 мс (SPEC §5 после spec-delta); локальный D-Bus RTT ~1-10 мс; медленный клиент может ответить дольше.
   - Recommendation: ~100 мс таймаут, при истечении — тихий отказ (счётчик); замерить фактические латентности в e2e-отчёте.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| zenity | GTK-поверхность матрицы | ✓ | 4.0.1-1build3 (dpkg) | — |
| google-chrome | Chromium-класс (ADR-003) | ✓ | 153.0.8010.36 | chromium snap 152.0.7977.64 |
| chromium (snap) | альтернатива Chromium-класса | ✓ | 152.0.7977.64 | google-chrome |
| gnome-text-editor | Open Q2 (критерий 1) | ✓ | 46.3-0ubuntu2 | — |
| ydotool | инжекция (уже в стенде) | ✓ | 0.1.8-3build1 | свой uinput-инжектор (~150 LOC, STACK) |
| /dev/uinput + группа input | инжекция | ✓ | root:input 0660+, user в `input` | — |
| /usr/bin/python3 + gi + Atspi | AT-SPI witness/readback | ✓ | atspi OK (живая проверка) | — |
| IBus | транспорт | ✓ | 1.5.29-rc2 | — |
| mise (go 1.23.12, golangci-lint 2.13.2) | задачи/гейты | ✓ | mise ls проверен | — |
| GNOME Wayland сессия | e2e | ✓ | wayland-0 жив | — |
| GitHub Actions runner | D-19 CI | ✗ | — | **поднять в этой фазе** (установка + systemd user unit + ярлык `gnome`) |

**Missing dependencies with no fallback:**
- GitHub Actions self-hosted runner — provisioning является частью объёма фазы (D-19), не блокатором планирования.

**Missing dependencies with fallback:**
- ydotool mouse (move/click) для кейса «клик-скачок» — если 0.1.8 mouse-команды поведут себя плохо, кейс заменяется (A8).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (+`-race`), vanilla, `package xxx_test` (go-ultimate) |
| Config file | нет (стандартные флаги); гейты — `mise.toml` |
| Quick run command | `go test -race -count=1 ./internal/... ./engine/...` |
| Full suite command | `mise run ci` (build + vet + lint + test + tidy-diff) |
| Live e2e | `mise run e2e-matrix` (живая сессия; НЕ в ci — прецедент e2e-m1) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CORR-01 | двойной тап → конвейер коррекции | unit (актёр+fake-эмиттер) + e2e | `go test -race ./internal/session/...` | частично (actor_test.go Фазы 1) — Wave 0 |
| CORR-04 | направление по составу | unit golden-корпус слов | `go test -race ./internal/correct/ -run TestDirection` | ❌ Wave 0 |
| CORR-05 | регистр посимвольно | unit (таблицы + convert) | `go test -race ./internal/correct/ -run TestConvert` | ❌ Wave 0 |
| CORR-07 | диапазон без прыжка, лестница | unit (plan/арифметика) + e2e (уровень из лога) | `go test -race ./internal/correct/ -run TestPlan` | ❌ Wave 0 |
| CORR-09 | сбросы Enter/Tab/Escape/FocusOut/Reset | unit (buffer) + e2e noop-кейсы | `go test -race ./internal/correct/ -run TestBuffer` | ❌ Wave 0 |
| TEST-04 | матрица YAML, PASS/FAIL, exit≠0 | e2e (живая сессия/раннер) | `mise run e2e-matrix` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** quick run + `golangci-lint run` (зелёная итерация D-08)
- **Per wave merge:** `mise run ci`
- **Phase gate:** `mise run ci` зелёный + `mise run e2e-matrix` зелёная на раннере (D-19) перед `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/correct/{buffer,direction,convert,verify,plan}_test.go` — корпусы CORR-01/04/05/07/09
- [ ] `internal/session/actor_test.go` — расширить: флип-режим, двухфазная коррекция, fake-эмиттер
- [ ] `test/e2e/matrix.go` + `test/e2e/cases/*.yaml` + `fixtures/input.html` — TEST-04
- [ ] `engine/` wire-тесты новых эмиттеров (по образцу `wire_test.go`)
- [ ] mise-задача `e2e-matrix`; префлайт: chromium-запуск, mouse-самотест
- [ ] `.github/workflows/e2e-matrix.yml` + runner-инструкция (docs/)

## Security Domain

ASVS L1 (config: security_enforcement=true, level=1, block_on=high).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | локальный демон без аутентификации; IBus-сокет = граница доверия сессии (Фаза 1 оценена) |
| V3 Session Management | no | — |
| V4 Access Control | no (частично — runner) | контроль доступа CI: `workflow_dispatch`-только (запускать могут роли с write), минимальные `permissions:` (прецедент pr-sanity: `contents: read`) |
| V5 Input Validation | yes | strict YAML (`KnownFields`), godbus-типизация аргументов, recover-шимы на каждом экспортированном методе (INTEG-05), comma-ok на вариантах |
| V6 Cryptography | no | — |

### Known Threat Patterns for {IBus engine daemon + self-hosted runner}

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Печать на живом столе владельца из CI (раннер в сессии) | Elevation/Tampering | `workflow_dispatch`-только (fork-PR не может запустить), `concurrency`-группа, раннер-группа с ограничением репозитория; обновить SECURITY.md |
| Keystroke-логгинг (буфер = всё, что набрал пользователь) | Information Disclosure | keystroke-трассировка только `-debug` (README прецедент); INFO — причины отказов без содержимого слов, счётчики (D-20/D-21) |
| Паника обработчика → смерть ввода стола | DoS | recover-шимы уже на всех методах (Фаза 1); новые эмиттеры — под теми же шимами через godbus пути |
| Кейс-инъекция через YAML (матрица читает файлы) | Tampering | строгий декодер; кейсы в git (review); раннер исполняет код репозитория только по dispatch |
| Чтение AT-SPI чужих окон | Information Disclosure | helper читает только по имени приложения, которое сам стенд запустил (`focus_helper.py text <app>`) |

## Sources

### Primary (HIGH confidence)
- Встроечное чтение кода этой сессией: `internal/hotkey/fsm.go`, `internal/session/actor.go`, `engine/{engine,types,keys,conn,factory}.go`, `layouts/tables.go`, `test/e2e/{main,case_m1,preflight}.go`, `focus_helper.py`, `cmd/goswitchd/main.go`, `mise.toml`, `.github/workflows/pr-sanity.yml`, `.golangci.yml`
- `docs/SPEC.md` §4.1–4.3, §5, §7, §8; `docs/adr/ADR-001..004`; `docs/adr/d01-experiment-log.md` — дословные цитаты
- Прототип владельца `/home/nil/.local/share/punto-switcher/punto_engine.py` — референс флипа/буфера/лестницы (включая баг хвоста: строки 186-199)
- Живой аудит окружения (dpkg, mise ls, systemctl, /dev/uinput, atspi-import) — эта сессия
- proxy.golang.org — yaml.v3 v3.0.1 (живой ответ `@latest`)

### Secondary (MEDIUM confidence)
- [ibus/ibus#2354 — Chromium-based apps: 'delete surrounding text' doesn't work](https://github.com/ibus/ibus/issues/2354) — подтверждён поиском этой сессии (дайджест закэширован)
- [GTK Development Blog — Input methods in GTK 4](https://blogs.gnome.org/gtk/2018/03/06/input-methods-in-gtk-4/), [IBusInputContext docs](https://ibus.github.io/docs/ibus-1.5/IBusInputContext.html), [GTK Integration (DeepWiki)](https://deepwiki.com/ibus/ibus/3.1-gtk-integration) — путь DeleteSurroundingText через GtkIMContext; фактическое применение zenity/GTK4 — эмпирика матрицы
- `.planning/research/{STACK,ARCHITECTURE,FEATURES,SUMMARY}.md` — исследовательская база Фазы 0 (сигнатуры сигналов 1.5.29, caps-биты, ibus#2054/#2600/#2423)

### Tertiary (LOW confidence)
- Поведение mutter text-input-v3 релея delete_surrounding → GTK4-клиент — не проверялось отдельно; закрывается e2e-кейсом GTK-класса

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — обе зависимости уже в STACK/прокси; код читан построчно
- Architecture: HIGH — все решения зафиксированы ADR-002/003/004 + D-13..D-21; проводка флипа — MEDIUM (Open Q1)
- Pitfalls: HIGH — баг хвоста воспроизведён из реального прототипа; рунный подсчёт — требование ADR-003; Chromium #2354 подтверждён

**Research date:** 2026-09-11
**Valid until:** 2026-10-11 (30 дней — стабильная доменная область; пере-проверить только версии Chrome/Chromium на раннере к приёмке фазы)
