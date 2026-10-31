# Phase 11: MENU — пункт «в чёрный список» - Research

**Researched:** 2026-10-31
**Domain:** трей-меню DBusMenu (internal/indicator), identity-наблюдатель a11y (internal/appid), атомарная запись конфига (internal/config/writer), blocklist-гейт автокоррекции (internal/session)
**Confidence:** HIGH — все компоненты in-repo, прочитаны в этом сеансе; внешних зависимостей фаза не добавляет

<user_constraints>
## User Constraints (from CONTEXT.md)

### Implementation Decisions (locked)

- **D-01:** Когда текущее приложение уже в блок-листе, пункт показывается с галочкой (отмеченное состояние); клик по отмеченному пункту — no-op (защита от дублей обеспечивает безопасность клика).
- **D-02:** Имя приложения берётся из кэша наблюдателя фокуса (`internal/appid`); живые a11y-запросы при открытии меню НЕ делаются. Открытие меню мгновенное, цена — пустой кэш до первого focus-события после старта демона (ведёт себя как «личность неизвестна», см. D-03).
- **D-03:** При неизвестной личности (snap, wezterm/winit, пустой кэш) пункт виден, но неактивен, **без имени приложения и без пояснительного текста** — решение владельца «не пишем ничего». Это сужает критерий 3 роадмапа («неактивен с объяснением») — владелец сузил осознанно 2026-10-31; текст роадмапа не переписывать, планировать по этому решению.
- **D-04:** `macroChordWhat` меняется с `" — смена режима"` на `" — язык"` — тот же хвост, что у строки одиночного шифта («1× … — язык»); итоговая строка меню: «Super+Space — язык».

### Claude's Discretion
- **Размещение пункта в меню** (владелец: «где — решай сам»): в группе автокоррекции, сразу под тумблером «Автокоррекция» — действие относится к автокоррекции; отдельную секцию не заводить. Реестр menuID и layout-карусель расширить по существующим паттернам menu.go.

### Deferred Ideas (OUT OF SCOPE)
- Удаление приложения из чёрного списка через меню — новая возможность, вне объёма MENU-01 (кандидат после v1.2).

### Specifics (from CONTEXT.md)
- Строка меню super+space: «Super+Space — язык»
- Галочка на пункте для уже заблокированного приложения
- Никакого объяснительного текста при неизвестной личности

### Established Patterns (CONTEXT.md, подтверждены кодом)
- Уведомления владельцу — строками меню, не десктоп-уведомлениями → отказ записи (потолок 64) показывать состоянием в меню.
- Hot-reload ленивый (`internal/config/watch.go`): применяется на следующем клавиатурном событии — достаточно для «замолкает без рестарта»; гейт blocklist читается при армировании (`internal/session/actor.go`).
- Новая запись идёт через существующий writer → watcher → ленивая перечитка конфига; отдельный канал проталкивания в actor не нужен.
- Имя приложения в меню — через сеттер из сессионного слоя (владелец кэша appid), по образцу `SetCorrections`/`SetKeys`.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| MENU-01 | Динамический пункт трей-меню «Добавить ИмяПриложения в чёрный список»: QuoteMeta имени, без дублей, потолок 64, атомарная запись + живая перечитка; неизвестная личность → пункт неактивен, демон не угадывает | Паттерн расширения menu.go (реестр ID, layout, дельты), паттерн click-композиции `configToggle` (main.go), новый writer-метод поверх `rewriteAtomically`, механизм свежести идентичности (см. «Свежесть идентичности» — центральная находка), fold blocklist в `applySnapshot` (actor.go:2188-2194) |
</phase_requirements>

## Summary

Фаза целиком собирается из готовых in-repo механизмов — новых внешних зависимостей **ноль**, новой инфраструктуры **ноль**. Пункт меню — это расширение статического каркаса DBusMenu (`menu.go`): новый menuID (16), вставка в `rootKidsLocked` сразу после `menuIDACToggle` (позиция в layout не меняет существующих ID — e2e-кейс `menu-v2` пинит id 1/2/4/5), property-builder по образцу `checkProps` (галочка) + `standardProps` (enabled), сеттер по образцу `SetKeys`/`SetCorrections`, дельты `ItemsPropertiesUpdated`. Клик идёт через `Callbacks` → click-композицию в `cmd/goswitchd/main.go` по образцу `configToggle.flip` (read → write → push → sync reload), запись — новый метод в `internal/config/writer.go` поверх готовых `rewriteAtomically`/`writeAtomically` (Node round-trip + temp+rename 0600).

**Центральная находка исследования:** буквальное прочтение D-02 («кэш наблюдателя») + D-03 («snap, wezterm/winit → пункт неактивен») **несовместимо с текущей семантикой кэша appid**. Наблюдатель держит последнюю известную личность сквозь потери фокуса (`focusGain` возвращает `gained=false` на loss-события и на пути без bridge-маркера — appid.go:141-156, 287-314; ничего не чистится), поэтому когда пользователь уходит в wezterm/snap, кэш продолжает отдавать **старое** приложение: пункт показал бы АКТИВНОЕ имя не того приложения, и клик молча заблокировал бы его. D-03 («демон не угадывает») достигается только если у «известной личности» появляется признак **свежести**: подтверждено gain-событием после последнего настоящего FocusOut движка. Рекомендуемый механизм — см. «Свежесть идентичности» ниже; он не трогает семантику гейтов D-53 (блок-гейт продолжает читать кэш как сегодня).

Второй существенный факт: `CONTEXT.md` описывает hot-reload как «fsnotify watcher», но код давно на **stat-опросе 2 с** (fsnotify выведен из эксплуатации — `watch.go:21`, решение владельца 2026-10-08, D-32 REV). Для фазы это ничего не меняет (атомарный rename писателя виден опросом; ленивый fold — на следующем событии клавиатуры), но планировщику и ревьюеру нужно правильное имя механизма.

**Primary recommendation:** расширять меню по существующим паттернам (ID 16, позиция под тумблером «Автокоррекция»), добавить в appid-наблюдателя признак свежести (gain-хук + сброс на реальный FocusOut), writer-метод `AppendAutocorrectBlocklist(path, pattern)` с дедупом/потолком/pre-write валидацией строгим декодером, и не начинать наблюдателя a11y принудительно при выключенной автокоррекции без явного решения (см. Open Questions 1).

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Отрисовка пункта меню (label/enabled/галочка) | internal/indicator (DBusMenu-сервер) | — | меню — «глупый» рендерер снапшота; property-ответы считаются AT REQUEST TIME из снапшота (канон menu.go) |
| Источник имени приложения (свежесть, кэш) | internal/appid + internal/session (Actor) | — | наблюдателем владеет актор (`ensureAppid`, actor.go:2835); наружу — через `MenuSync`-сеттер (CONTEXT: по образцу SetCorrections/SetKeys) |
| Решение о клике (known/covered/full → писать/нет) | cmd/goswitchd (click-композиция) | internal/session (Status/аксессоры) | тогглы меню уже живут в main.go (`configToggle`); актор — источник истины о состоянии |
| Запись в конфиг (QuoteMeta-паттерн, дедуп, потолок, atomic) | internal/config/writer.go | — | единственная точка записи документов (`setDocumentToggle`-прецедент: ONE persist path) |
| Замолкание автокоррекции в приложении | internal/session (гейт при армировании) | internal/config (watch: ленивая перечитка) | blocklist-гейт уже в `autoCorrectBoundary` (actor.go:3342-3348); фаза ничего в нём не меняет |
| Документация (README, CONFIG.md) | docs/ | — | «смена режима» в README.md:102/README.en.md; контракт apps_blocklist в CONFIG.md:90 |

## Standard Stack

### Core (in-repo — новые внешние зависимости ЗАПРЕЩЕНЫ constraint'ом проекта)

| Component | Version | Purpose | Why |
|-----------|---------|---------|-----|
| `internal/indicator/menu.go` | in-repo | DBusMenu-сервер: каркас `menuLayout`/`menuItemProps`, реестр menuID, `AboutToShow`+`layoutRevision`, `checkProps`, сеттеры `Set*`, дельты `ItemsPropertiesUpdated`/`LayoutUpdated` | весь динамический контракт меню уже здесь; пункт = +1 ID, +1 props-ветка, +1 сеттер [VERIFIED: internal/indicator/menu.go:56-91, 444-487] |
| `internal/appid` | in-repo | кэш личности сфокусированного приложения (bridge-namespace, `FocusedApp`), сентинелы `ErrRoleUnknown`/`ErrBusClosed` | единственный источник identity; ленивый старт через `ensureAppid` [VERIFIED: internal/appid/appid.go:253-258; internal/session/actor.go:2835-2850] |
| `internal/config/writer.go` | in-repo | атомарная запись YAML: `yaml.Node` round-trip, `rewriteAtomically` (temp same-dir + rename, 0600), helpers `mapping`/`scalarString`/`emptySeq` | критерий 2 (atomic + золотой дифф) садится на готовый механизм [VERIFIED: internal/config/writer.go:116-139, 301-350] |
| `internal/config` (валидация) | in-repo | потолок `maxAutocorrectBlocklist = 64`, regex-валидация, строгий `KnownFields`-декодер (`Load`) | критерии 2 и 4: pre-write валидация = прогон `Load`-семантики по байтам до rename [VERIFIED: internal/config/config.go:28, 472-501; internal/config/load.go:27-28] |
| `internal/session` (`Actor`) | in-repo | владелец appid-наблюдателя, `MenuSync`-пуши, fold `applySnapshot` (blocklist перекомпилируется при изменении списка), гейт | сеттер-пуш и fold-дисциплина готовы [VERIFIED: internal/session/actor.go:1889-1895, 2188-2203, 2282-2303] |
| `godbus/dbus/v5` v5.2.2 | go.mod | D-Bus транспорт (меню, a11y-шина) | без изменений [CITED: .planning STACK.md / go.mod] |
| `gopkg.in/yaml.v3` v3.0.1 | go.mod | Node round-trip writer'а | без изменений |

### Supporting
| Component | Purpose | When to Use |
|-----------|---------|-------------|
| `regexp.QuoteMeta` (stdlib) | экранирование имени приложения в regex-паттерн | в click-композиции перед записью; writer получает ГОТОВЫЙ паттерн (писатель не знает про identity) |
| `test/e2e/case_menu.go` | e2e-драйвер меню по D-Bus (`com.canonical.dbusmenu.Event`, без ydotool/мыши) | новый e2e-кейс пункта по этому образцу [VERIFIED: test/e2e/case_menu.go:31-48] |
| `mise` tasks (`mise.toml`) | build/vet/lint/test/ci-гейты | гейты зелёной итерации; e2e-задача добавляется по образцу `e2e-*` |

**Installation:** ничего не устанавливается; go.mod не меняется (гейт `mise run tidy-diff`).

## Package Legitimacy Audit

Внешние пакеты фазой **не добавляются** (constraint: stdlib + godbus + минимум зависимостей). Проверок реестра не требуется; `go.mod` не должен измениться.

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| — (none) | — | — | — | — | — | No new installs |

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### System Architecture Diagram

```
                          a11y bus (AT-SPI2)
   [GTK/Qt app bridge] --StateChanged focused=1 (gain)--> internal/appid.Observer
        (bridge namespace = identity)                          |  run(): cache app + (sender,path)
                                                               |  НОВОЕ: push-хук OnGain(name)
                                                               v
                    engine FocusOut (real) --(сброс свежести)--> internal/session.Actor
                                                               |  снапшот: имя, fresh, covered, full
                                                               |-- MenuSync.SetBlocklistApp(...) --> indicator.Menu (снапшот + ItemsPropertiesUpdated)
                                                               |
  [user opens tray menu] --> gnome-shell renderer --> com.canonical.dbusmenu GetLayout/GetGroupProperties
                                                               |  props AT REQUEST TIME из снапшота
                                                               v
  [user clicks item] --> com.canonical.dbusmenu.Event(id=16) --> indicator.Menu.dispatchClick
                                                               |  Callbacks.BlockCurrentApp (nil = no-op)
                                                               v
                              cmd/goswitchd click-композиция (образец configToggle.flip)
                               1. identity known? fresh? covered (галочка -> no-op)? full (-> no-op, состояние в меню)
                               2. pattern = regexp.QuoteMeta(identity)
                               3. config.AppendAutocorrectBlocklist(path, pattern)
                                  read -> Node round-trip -> append -> strict pre-validate -> temp+rename
                               4. push: SetBlocklistApp(name, covered=true) + sync reload
                                                               v
                    config.Watcher (stat-опрос 2 c, watch.go:21) --> reparse (strict Load + Validate)
                                                               v
                    Actor.applySnapshot (следующее клавиатурное событие) --> refreshACBlocklist
                                                               v
                    autoCorrectBoundary: re.MatchString(identity) -> abstain ac_app_blocked
```

### Pattern 1: Расширение реестра menuID и layout-карусели
**What:** новый пункт = новая константа ID (следующий свободный: **16**, после `menuIDReload = 15`), вставка в `rootKidsLocked` на нужной ПОЗИЦИИ (сразу после `menuIDACToggle` — дискреция, закреплённая в CONTEXT), props-ветка в `itemPropsLocked`, case в `dispatchClick`.
**Почему новый ID, а не перенумерация:** e2e-кейс `menu-v2` жёстко пинит существующие id (`menuIDACToggle = int32(4)` и т.д., case_menu.go:40-43, «a layout change breaks this case loudly, by design»); unit-корпус `TestMenuLayout` пинит точный порядок детей. ID 16 с позицией 4 в списке детей сохраняет все существующие пины (правятся только ожидания длины/порядка в `TestMenuLayout` — и это честная RED-ступень TDD).
**Важно:** `GetGroupProperties` фильтрует по диапазону `id < menuIDEN || id > menuIDReload` (menu.go:365) — верхнюю границу надо расширить, иначе id 16 молча выпадает из ответов (renderer не увидит props).

### Pattern 2: Сеттер состояния извне (образец SetKeys/SetCorrections)
**What:** `MenuSync` (actor.go:1889-1895) расширяется методом вида `SetBlocklistApp(name string, covered bool)` (точную сигнатуру — включая состояния fresh/full — выбирает планировщик; можно одним сеттером с полями, можно парой). Дисциплина сеттера: dedupe identical values (никаких эмиссий на равных значениях), label-only изменения — `ItemsPropertiesUpdated`, эмиссия — с ПУСТЫМ слайс-литералом removed-props (Pitfall 7: nil ломает сигнатуру и паникует первую живую эмиссию).
**Push-моменты:** (а) gain-хук наблюдателя (новое имя); (б) реальный FocusOut (свежесть=false); (в) fold `applySnapshot` (covered/full пересчитаны по новому документу); (г) install-пуш при attach (образец: `menu.SetKeys(...)` в OnConn, main.go:275).

### Pattern 3: Свежесть идентичности (центральная находка)
**What:** кэш appid — «последняя известная», не «текущая»: `focusGain` отбрасывает loss-события и пути без `/a11y/`-маркера, `run()` хранит кэш до следующего gain (appid.go:141-156, 287-314). Без коррекции семантики пункт в wezterm/snap покажет активное имя ПРЕДЫДУЩЕГО приложения — клик запишет мусор в blocklist, что прямо противоречит D-03 («демон не угадывает, не пишет мусорные паттерны»). Критерий 3 роадмапа называет snap/wezterm случаями «неизвестной личности» — значит, свежесть обязательна.
**Механизм (рекомендация):** подтверждённая личность = gain-событие, пришедшее ПОСЛЕ последнего реального FocusOut движка (или после старта демона).
- `Observer` получает опциональный хук (Option вида `WithGainHook(func(name string))`), вызываемый из `run()` ВНЕ критической секции (после store) — не держит `appid.mu` при вызове, чтобы не инвертировать порядок блокировок (существующая односторонняя дисциплина actor.mu → sink.mu, pushA11y-комментарий actor.go:2250-2263).
- daemon-проводка соединяет хук с публичным методом актора (например `actor.PushFocusedApp(name)`): под мьютексом обновляет снапшотные поля (имя, fresh=true), пересчитывает covered/full, пушит дельту в меню.
- `HandleLifecycle` реальный FocusOut (ветка G-2-3: НЕ synthetic — без flipCredit, actor.go:1302) ставит fresh=false и пушит.
- Гейты D-53 НЕ трогаются: `acFocusedApp` продолжает читать кэш как сегодня (blocklist-семантика «неизвестная не запрет» не меняется).
**Гонка gain-vs-FocusIn:** a11y-gain и engine-FocusIn — независимые транспорты; gain может прийти позже FocusIn. Хук это закрывает: после прихода gain снапшот обновится независимо от событий движка (сходимость к истине, не атомарность). Клик-only сценарий («переключил окно → открыл трей») покрыт, т.к. gain обрабатывается мгновенно по приходу.
**Поздние/ложные gains:** быстрое A→B→A сходится к последнему gain; гарантия — сходимость, задокументировать в комментарии.

### Pattern 4: Click-композиция (образец configToggle.flip)
**What:** новая композиция в main.go рядом с `menuToggles`: read (состояние из актора: identity/fresh/covered/full) → guards → write → push → sync reload. Гварды:
- пустой путь конфига — отказ с WARN (образец WR-03 `configToggle.flip`, main.go:476-480);
- личность неизвестна/несвежа — молчаливый no-op (fail-closed: renderer и так не должен слать клик по enabled:false — [CITED: github.com/ubuntu/gnome-shell-extension-appindicator, dbusMenu.js — disabled-актор нерективен], но daemon-гвард обязателен как второй пояс);
- covered (галочка) — no-op по D-01 (для checkmark использовать ту же дисциплину матча, что у гейта: любой скомпилированный паттерн совпал с identity — покрыто; это делает галочку честной: «коррекция уже молчит»);
- full (потолок) — no-op, отказ уже показан состоянием пункта.
Запись: `pattern := regexp.QuoteMeta(identity)`; после записи — `push` (новое covered-состояние) и синхронный reload (`reloadConfig` + `actor.FoldAppliedConfig`, образец `toggleReload`, main.go:249-255) — «замолкает без рестарта» в том же клике, без ожидания stat-опроса.

### Pattern 5: Writer-метод добавления в список (новый, поверх готового)
**What:** `AppendAutocorrectBlocklist(path, pattern string) error` (имя — за планировщиком) в writer.go:
1. read → `yaml.Unmarshal` в `yaml.Node` (образец `setDocumentToggle` writer.go:116-139, включая ветку absent-file: создать полный дефолтный документ с паттерном на месте — `buildFullDocument` + вставка);
2. найти секцию `autocorrect` (`sectionAutocorrect = "autocorrect"`, writer.go:26), ключ `apps_blocklist` (`keyAppsBlocklist`, writer.go:30); дедуп: паттерн уже в списке → no-op success (idempotent, образец `TestWriter_IdempotentFlip`);
3. потолок: `len(list) >= maxAutocorrectBlocklist` (config.go:28, константа в том же пакете) → sentinel-error (дисциплина err113: пакетный sentinel вида `errAutocorrectBlocklistOverCeil`, обёртка с путём);
4. append: к SequenceNode добавить `scalarString(pattern)`; при null/отсутствующем ключе — создать последовательность (образец `upsertToggle`/`buildSectionNode`, но сигнатура buildSectionNode bool-специфична — не переиспользовать вслепую);
5. **pre-write валидация** (критерии 2 и 4): закодировать в байты и прогнать строгий декод (`KnownFields(true)` + `Validate()` — семантика `Load`, load.go:27-28) по байтам в памяти; отказ → ошибка, оригинал не тронут;
6. `rewriteAtomically` (temp same-dir, rename, 0600 — writer.go:326-350).
**YAML-стили:** `scalarString` ставит Style 0 (plain) — энкодер yaml.v3 сам закавычит спец-символы (имена с `#`, `:` и т.п.), ручное квотирование не нужно и не должно добавляться.

### Pattern 6: D-04 — переименование хвоста super+space
`macroChordWhat = " — смена режима"` → `" — язык"` (menu.go:84). Метка собирается в `bindingProps(m.modeChord, macroChordWhat)` → «super+space — язык». Обновить пин в `TestMenuLayout` (menu_test.go:160: «super+space — смена режима») и документацию: README.md:102, README.en.md (строка «`<сочетание> — смена режима`»). Проверить e2e на строковые пины хвоста — grep «смена режима» по test/ чист (VERIFIED: grep по test/, docs/ пуст кроме README).

### Recommended Project Structure
Новая логика ложится в существующие файлы (структура не растёт):
```
internal/indicator/menu.go      # +ID 16, +props-ветка, +сеттер, +dispatchClick case, D-04
internal/indicator/menu_test.go # +пины layout/states/deltas
internal/appid/appid.go         # +WithGainHook (Option), вызов вне критсекции
internal/session/actor.go       # +PushFocusedApp, +fresh-поле, +fold-пуш covered/full, MenuSync-расширение
internal/session/actor_test.go  # +fakeMenuSync-метод, +кейсы fresh/unconfirmed/full
internal/config/writer.go       # +AppendAutocorrectBlocklist (+sentinel потолка)
internal/config/writer_test.go  # +golden-дифф, дедуп, потолок, regex-метасимволы, strict-load
cmd/goswitchd/main.go           # +click-композиция, +Callbacks.BlockCurrentApp, +gain-хук проводка, +install-пуш
cmd/goswitchd/main_test.go      # +композиция-кейсы
test/e2e/case_menu*.go          # +кейс menu-blocklist (Event-клик по id 16)
mise.toml                       # +e2e задача
README.md, README.en.md, docs/CONFIG.md, docs/SPEC.md # D-04 + новое поведение пункта
```

### Anti-Patterns to Avoid
- **Не угадывать личность:** никогда не писать в blocklist паттерны вида `:1.NN` (D-Bus sender-имена) и не «заполнять» имя из чужих источников; нет свежей identity — пункт неактивен (D-03).
- **Не делать живых a11y-запросов при открытии меню** (D-02) — только кэш/снапшот; живой Role-запрос существует только для гейта при армировании.
- **Не проталкивать конфиг в актор мимо watcher:** запись → watcher → fold; отдельный канал не нужен (CONTEXT Integration Points; ленивый fold на следующем клавиатурном событии + синхронный reload из клика покрывают критерий 1).
- **Не переиспользовать `buildSectionNode`** для list-секции: сигнатура bool-специфична (`buildSectionNode(section, on)`).
- **Не эммитить nil-slice** в `ItemsPropertiesUpdated` (Pitfall 7, menu.go:93-101).
- **Не забывать `doubleUnderscores`** для имени приложения в label (GTK-мнемоники: dbusMenu.js режет одиночный `_`, menu.go:645-651).
- **Не менять гейты D-53/блок-гейт/терминальное вето** — фаза только добавляет поверхность управления.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Атомарная запись YAML с сохранением комментариев | свой writer поверх os.WriteFile | `rewriteAtomically`/`writeAtomically` (writer.go:301-350) | Node round-trip + same-dir temp + rename + 0600 уже пинены тестами (T-07-03-01/02) |
| Валидация документа после правки | ручная проверка полей | строгий `KnownFields`-декодер + `Validate()` (семантика `Load`) | единственное место, где битый regex/потолок/лишний ключ видны до rename (D-33) |
| Экранирование regex-метасимволов | ручные замены | `regexp.QuoteMeta` | имена вида `org.gnome.Weather`, `App[1]`, `wezterm (2)` |
| Галочка/enable/visible-орнаменты | свой протокол | `toggleProps`/`standardProps`/`hiddenProps` + `stateInt` | dbusMenu.js:731-739 читает ровно эти ключи (menu.go:604-613) |
| Динамика меню при закрытом меню | свой протокол обновлений | `ItemsPropertiesUpdated` (renderer флагает и переигрывает при открытии) + `LayoutUpdated` только на смену набора | канон menu.go (research A4/Q1 фазы 07) |
| Драйвер e2e меню | ydotool/мышь | `com.canonical.dbusmenu.Event` по живому демону | кейс menu-v2 доказал канал живьём (case_menu.go:3-14) |
| Дедуп/потолок списка | проверки в click-композиции | writer-метод (same package, константа `maxAutocorrectBlocklist`) | единственный писатель документа; композиция не должна знать схему |

**Key insight:** весь критичный путь (запись, валидация, перечитка, гейт) уже пинен корпусом; фаза — «последний сантиметр» UI поверх него. Опасность не в новых механизмах, а в семантике кэша appid (свежесть) и в точных пинах ID/сигнатур.

## Runtime State Inventory

> Фаза содержит элемент rename (D-04: строка UI) и добавляет запись данных в существующий пользовательский конфиг. Полный инвентарь:

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | `~/.config/goswitch/config.yaml` → `autocorrect.apps_blocklist` — фаза ДОБАВЛЯЕТ записи (QuoteMeta-паттерны) через writer; существующие документы могут не содержать секцию/ключ (вопрос валиден) | новый writer-метод (код); миграции данных нет |
| Live service config | None — verified: меню живёт на собственном экспорте демона; внешний DBusMenu-состояния не хранит (renderer перечитывает layout по revision/сигналам) | none |
| OS-registered state | None — verified: systemd user unit, имена шин, dconf-ключи фазу не затрагивают | none |
| Secrets/env vars | None — verified: новых env/секретов нет; D-20/D-21 (ни слово, ни содержимое секций в журнале) распространяется на новый writer-путь (прецедент writer.go:1-8 «logs nothing») | none |
| Build artifacts | None — verified: rename затрагивает только Go-константу и markdown; `mise run tidy-diff` гейтит go.mod-дрейф | none |

Rename-поверхность D-04: `internal/indicator/menu.go:84` (константа), `internal/indicator/menu_test.go:160` (пин), `README.md:102`, `README.en.md` (документация). Runtime-систем с закэшированной старой строкой нет (label считается AT REQUEST TIME из снапшота).

## Common Pitfalls

### Pitfall 1: Кэш appid держит устаревшее имя (главная ловушка фазы)
**What goes wrong:** пункт показывает активное имя предыдущего приложения, когда пользователь уже в wezterm/snap/поле без bridge; клик блокирует не то приложение.
**Why it happens:** `focusGain` отбрасывает loss-события и немаркированные пути; `run()` хранит кэш до следующего gain — семантика «последняя известная», полезная для гейтов, опасная для UI.
**How to avoid:** свежесть (Pattern 3): сброс на реальный FocusOut, подтверждение gain-хуком; состояние fresh=false рендерится как неизвестная личность (D-03).
**Warning signs:** в тестах — сценарий «gain(app A) → real FocusOut → открытие меню» обязан показывать неактивный пункт; если зелёный без fresh-поле — семантика не покрыта.

### Pitfall 2: Наблюдатель не запущен при выключенной автокоррекции
**What goes wrong:** `ensureAppid` стартует a11y-наблюдателя только при `len(MACRApps) > 0 || AutoCorrectEnabled` (actor.go:2836 — verbatim: `if a.appidStarted || (len(a.opts.MACRApps) == 0 && !a.opts.AutoCorrectEnabled) { return }`); с выключенной автокоррекцией (D-54 дефолт) кэш пуст навсегда — пункт мёртв до первого включения тумблера.
**How to avoid:** принять как честное сужение (пункт про автокоррекцию; пользователь включает её тем же меню) ИЛИ расширить условие/стартовать при проводке (Open Question 1 — решение за планировщиком/чекпоинтом).
**Warning signs:** e2e/юнит с автокоррекцией off не должен ждать имени.

### Pitfall 3: nil вместо пустого слайса в эмиссии
**What goes wrong:** паника на первой живой эмиссии `ItemsPropertiesUpdated` (godbus не считает сигнатуру a(ias) от nil).
**How to avoid:** только `[]removedProps{}` (menu.go:93-101 — Pitfall 7 VERIFIED).
**Warning signs:** юнит `TestMenuSignalFormPins`; живой клик.

### Pitfall 4: Имя приложения в label без экранирования мнемоник
**What goes wrong:** `org.gnome.Tab_1` рендерится как `org.gnome.Tab1` (dbusMenu.js режет одиночный `_`).
**How to avoid:** `doubleUnderscores(name)` в билдере label (прецедент `macroTapProps`/`bindingProps`).

### Pitfall 5: Диапазон GetGroupProperties и е2e-пины ID
**What goes wrong:** id 16 отфильтрован `id > menuIDReload` — renderer не получает props (пункт без label/enabled); либо правка «сдвигает» существующие ID и ломает кейс menu-v2/юнит-пины.
**How to avoid:** ID 16, позиция в `rootKidsLocked` после `menuIDACToggle`, расширить верхнюю границу фильтра; обновить `wantKids` в `TestMenuLayout` (это ожидаемая RED-ступень), не трогать существующие числа.

### Pitfall 6: Потолок 64 без состояния в меню
**What goes wrong:** клик на заполненном списке молча ничего не делает — критерий 4 требует уведомление из трея (строкой/состоянием в меню; десктоп-уведомления на этом столе невидимы — вердикт UAT фазы 7).
**How to avoid:** состояние full — отдельное видимое состояние пункта (например disabled + label «Чёрный список заполнен — 64/64»; формулировка — дискреция), пересчитывается на каждом fold; пишется только writer'ом (sentinel), композиция never bypass.
**Warning signs:** юнит: 64 паттерна в документе → fold → пуш full-состояния; клик no-op.

### Pitfall 7: Метасимволы в имени и strict-декодер
**What goes wrong:** имя с `*?()[]` без QuoteMeta ломает паттерн (или Compile отказывает → ВЕСЬ документ отвергнут при перечитке — D-33, last-good остаётся, но пункт «сделал всё хуже»).
**How to avoid:** `QuoteMeta` в композиции; pre-write строгая валидация в writer'е (критерий 4: после записи строгий декодер принимает файл целиком — юнит-тест с именем `org.example.App[1] (test)` — цитата-образец критерия).
**Warning signs:** `TestWriter_FullDocumentStrictLoad`-образец расширить append-версией.

### Pitfall 8: Гонка «gain позже FocusIn»
**What goes wrong:** наивная оценка свежести в момент FocusIn: gain ещё не дошёл — пункт неактивен в активном приложении (первичный сценарий «переключил окно → клик в трей»).
**How to avoid:** обновление снапшота по приходу gain (хук), не только на событиях движка; сходимость вместо атомарности (Pattern 3).

### Pitfall 9: Lock-инверсия appid.mu → actor.mu
**What goes wrong:** хук, вызванный из `run()` под мьютексом наблюдателя, берёт actor.mu — инверсия к существующему порядку (актор под своим мьютексом зовёт `FocusedApp()`) → дедлок под нагрузкой.
**How to avoid:** хук вызывается ПОСЛЕ снятия appid.mu (вне критсекции); в хуке не читать наблюдателя. Дисциплина односторонних порядков — pushA11y-комментарий (actor.go:2250-2263).

### Pitfall 10: Пуш дельты не того типа
**What goes wrong:** смена label/enabled/галочки — property-динамика; эмиссия `LayoutUpdated`/revision-бампа на каждое изменение имени сделает лишние полные перечитывания (revision — ТОЛЬКО на смену набора пунктов, menu.go:42-45).
**How to avoid:** пункт ВСЕГДА в layout (виден и в unknown-состоянии — D-03 «пункт виден»), поэтому все его изменения — `ItemsPropertiesUpdated`; revision не трогается вовсе.

## Code Examples

Все цитаты — verbatim из прочитанных в этом сеансе файлов.

### Реестр ID и метки (меню)
```go
// Source: internal/indicator/menu.go:56-91 (фрагмент)
	menuIDRoot        = int32(0)
	menuIDEN          = int32(1)
	menuIDRU          = int32(2)
	menuIDSep1        = int32(3)
	menuIDACToggle    = int32(4)
	menuIDSoundToggle = int32(5)
	menuIDMacro1      = int32(6)
	menuIDMacro2      = int32(7)
	menuIDMacro3      = int32(8)
	menuIDMacroCombo  = int32(9)
	menuIDMacroChord  = int32(10)
	menuIDSep2        = int32(11)
	menuIDSettings    = int32(12)
	menuIDAbout       = int32(13)
	menuIDStatus      = int32(14)
	menuIDReload      = int32(15)
```
```go
// Source: internal/indicator/menu.go:83-84
	macroComboWhat = " — слово и язык"
	macroChordWhat = " — смена режима"
```
→ D-04: `" — смена режима"` → `" — язык"`.

### Порядок детей root и фильтр диапазона
```go
// Source: internal/indicator/menu.go:444-457
func (m *Menu) rootKidsLocked() []int32 {
	ids := []int32{
		menuIDEN, menuIDRU, menuIDSep1, menuIDACToggle, menuIDSoundToggle,
		menuIDMacro1, menuIDMacro2, menuIDMacro3,
	}
	if m.wordCombo != "" {
		ids = append(ids, menuIDMacroCombo)
	}
	if m.modeChord != "" {
		ids = append(ids, menuIDMacroChord)
	}

	return append(ids, menuIDSep2, menuIDSettings, menuIDAbout, menuIDStatus, menuIDReload)
}
```
```go
// Source: internal/indicator/menu.go:364-366
	for _, id := range ids {
		if id < menuIDEN || id > menuIDReload {
			continue
		}
```

### Свойства пункта: галочка и enable (готовые билдеры)
```go
// Source: internal/indicator/menu.go:607-613
func toggleProps(label, kind string, on bool) map[string]dbus.Variant {
	p := standardProps(label, true)
	p[menuPropToggleType] = dbus.MakeVariant(kind)
	p[menuPropToggleState] = dbus.MakeVariant(stateInt(on))

	return p
}
```
```go
// Source: internal/indicator/menu.go:599-602
func standardProps(label string, enabled bool) map[string]dbus.Variant {
	return baseProps(label, enabled, true)
}
```

### Сеттер с dedupe и эмиссией дельт (образец для нового сеттера)
```go
// Source: internal/indicator/menu.go:262-273
func (m *Menu) SetCorrections(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if n == m.corrections {
		return
	}
	m.corrections = n
	m.emitSignal(signalItemsPropsUpdated,
		[]menuItemProps{m.labelDelta(menuIDStatus)},
		[]removedProps{})
}
```

### MenuSync-шов и push из fold
```go
// Source: internal/session/actor.go:1889-1895
type MenuSync interface {
	SetMode(symbol string)
	SetAutocorrectEnabled(on bool)
	SetSoundEnabled(on bool)
	SetKeys(tap, combo, chord string)
	SetCorrections(n int)
}
```
```go
// Source: internal/session/actor.go:2297-2302
	if a.menuSync != nil {
		a.menuSync.SetAutocorrectEnabled(snap.Autocorrect.Enabled)
		a.menuSync.SetSoundEnabled(a.soundEnabled)
		a.menuSync.SetKeys(snap.Hotkeys.TapKey, snap.Hotkeys.WordLayoutCombo, snap.Hotkeys.ModeSwitchChord)
		a.menuSync.SetCorrections(a.corrDone)
	}
```
```go
// Source: internal/session/actor.go:2188-2194 (fold blocklist)
	a.opts.AutoCorrectEnabled = snap.Autocorrect.Enabled
	a.opts.AutoCorrectBlocklist = snap.Autocorrect.AppsBlocklist
	a.opts.AutoCorrectMinWordLen = snap.Autocorrect.MinWordLen
	a.opts.AutoCorrectMargin = snap.Autocorrect.TrigramMargin
	a.opts.AutoCorrectFloor = snap.Autocorrect.TrigramFloor
	a.refreshACBlocklist(snap.Autocorrect.AppsBlocklist)
	a.refreshACTerminals(snap.Correction.Terminals)
```

### ensureAppid — точное условие ленивого старта
```go
// Source: internal/session/actor.go:2835-2839
func (a *Actor) ensureAppid() {
	if a.appidStarted || (len(a.opts.MACRApps) == 0 && !a.opts.AutoCorrectEnabled) {
		return
	}
	a.appidStarted = true
```

### Кэш-keep семантика наблюдателя (корень Pitfall 1)
```go
// Source: internal/appid/appid.go:300-313
	for {
		select {
		case <-ctx.Done():
			return
		case sig, ok := <-feed:
			if !ok {
				return
			}
			if name, gained := focusGain(sig); gained {
				o.mu.Lock()
				o.app = name
				o.focusSender = dbus.Sender(sig.Sender)
				o.focusPath = sig.Path
				o.mu.Unlock()
			}
		}
	}
```
```go
// Source: internal/appid/appid.go:148-156
	kind, ok := sig.Body[0].(string)
	if !ok || kind != "focused" {
		return "", false
	}
	state, ok := sig.Body[1].(int32)
	if !ok || state != 1 {
		return "", false
	}
	name := bridgeApp(sig.Path)
```

### Click-композиция-образец и синхронный reload
```go
// Source: cmd/goswitchd/main.go:475-494
func (t *configToggle) flip() {
	if t.path == "" {
		slog.Warn("config toggle unavailable", "component", "tray indicator", "reason", "no config path")

		return
	}
	on := !t.read()
	if err := t.write(t.path, on); err != nil {
		slog.Warn("config toggle write failed", "component", "tray indicator", "error", err)

		return
	}
	t.push(on)
	if t.reload == nil {
		slog.Warn("config toggle apply deferred", "component", "tray indicator", "reason", "no config source")

		return
	}
	t.reload()
}
```
```go
// Source: cmd/goswitchd/main.go:249-255
			toggleReload := syncReload
			if reload != nil {
				toggleReload = func() {
					syncReload()
					actor.FoldAppliedConfig()
				}
			}
```

### Writer: константы секции/ключа, atomic rename
```go
// Source: internal/config/writer.go:25-35 (фрагмент)
	sectionAutocorrect = "autocorrect"
	sectionSound       = "sound"
	keyEnabled         = "enabled"

	keyAppsBlocklist    = "apps_blocklist"
```
```go
// Source: internal/config/writer.go:326-331
func writeAtomically(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp config %s: %w", path, err)
	}
```

### Потолок и валидация blocklist
```go
// Source: internal/config/config.go:28
	maxAutocorrectBlocklist = 64
```
```go
// Source: internal/config/config.go:479-486
	for i, pattern := range a.AppsBlocklist {
		if strings.TrimSpace(pattern) == "" {
			return fmt.Errorf("autocorrect.apps_blocklist[%d] %q: %w", i, pattern, errAutocorrectBlocklistEmpty)
		}
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("autocorrect.apps_blocklist[%d] = %q: %w", i, pattern, errAutocorrectBlocklistRegex)
		}
	}
```
```go
// Source: internal/config/load.go:27-28
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true) // D-33: a typo'd key invalidates the whole config.
```

### Статус-поверхность (для click-композиции)
```go
// Source: internal/session/actor.go:1140-1143
	AutoCorrectEnabled     bool
	AutoCorrectFired       int
	AutoCorrectAbstained   int
	AutoCorrectSkipReasons map[string]int
```
Композиции-тогглы читают именно Status: `read: func() bool { return actor.StatusSnapshot().AutoCorrectEnabled }` (main.go:371). Идентичность можно довести новым полем Status (identity — не слово, D-20/D-21 не нарушает; данные конфиг-области) либо отдельным аксессором — выбор планировщика.

### Hot-reload: stat-опрос, не fsnotify
```go
// Source: internal/config/watch.go:21
const defaultPollInterval = 2 * time.Second
```
(Запись атомарным rename меняет mtime+size → опрос видит; fold — на следующем клавиатурном событии; клик дополнительно делает синхронный reload — критерий 1 покрыт дважды.)

### e2e-пины ID и клик-канал
```go
// Source: test/e2e/case_menu.go:40-43
	menuIDEN          = int32(1)
	menuIDRU          = int32(2)
	menuIDACToggle    = int32(4)
	menuIDSoundToggle = int32(5)
```
```go
// Source: test/e2e/case_menu.go:32-38 (фрагмент)
	menuBusName = "org.djarvur.goswitch"
	menuObjPath = "/Menu"
	menuIface   = "com.canonical.dbusmenu"

	menuEventMethod = "com.canonical.dbusmenu.Event"
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| fsnotify- watch конфига (directory+debounce) | stat-опрос mtime+size каждые 2 с (`defaultPollInterval`) | 2026-10-08 (D-32 REV, решение владельца; Electron исчерпывает inotify-бюджет) | CONTEXT.md code_context упоминает fsnotify — планировать по коду: watcher = Watcher/poll; для фазы разницы нет (rename виден опросу) |
| blocklist как белый список `autocorrect.apps` | чёрный список `apps_blocklist`, D-53 | 2026-10-04 | семантика готова, фаза только пишет в неё |
| Десктоп-уведомления | строки/состояния в меню | 2026-10-05 (UAT фазы 7) | отказ потолка — состояние пункта |

**Deprecated/outdated:**
- godbus v4, goibus-зависимости — не в этой фазе (см. STACK.md What NOT to Use).
- «Ленивый reload видит только rename-записи» (CONTEXT) — механизм теперь poll, вывод тот же.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Renderer (gnome-shell appindicator) не диспатчит `clicked` для `enabled:false` пунктов — [CITED: github.com/ubuntu/gnome-shell-extension-appindicator dbusMenu.js, MEDIUM: поведение стандартное для St.Widget non-reactive, первоисточник в этом сеансе не прочитан построчно] | Pattern 4 | Клик по неактивному пункту дошёл бы до демона — гвард композиции (fail-closed no-op) обязан быть; риск закрыт вторым поясом |
| A2 | gain-события a11y доходят регулярно на реальном столе при фокусе bridge-приложений (включая поле-свитчи) — [ASSUMED: поведение atk-bridge; в этом сеансе живой пробы не было] | Pattern 3, Pitfall 8 | Пункт мог бы залипать в unknown в активных приложениях; закрывается юнит-корпусом + живой e2e-приёмкой |
| A3 | Реальный engine FocusOut приходит на каждое переключение окон (включая уход в невидимые приложения) — [VERIFIED: in-repo, машина фазы 10: flap-пары FocusOut/FocusIn, actor.go:1272-1347, STATE.md ряды] | Pattern 3 | Если бы FocusOut не приходил, fresh залип бы true — но это код самого демона, пинен корпусом |
| A4 | Существующие документы пользователей могут не иметь секции `autocorrect`/ключа `apps_blocklist` (writer обязан уметь absent-ветку) — [VERIFIED: writer.go:116-139 ensure-ветка для тогглов; для списка — вывод из D-54 «nil blocklist» дефолта] | Pattern 5 | Отсутствие ветки = ошибка записи на «чистых» установках; тест-кейс обязателен |
| A5 | Формулировка full-состояния («Чёрный список заполнен — 64/64») — дискреция; CONTEXT锁定 только механизм «состоянием в меню» | Pitfall 6 | Низкий: строка правится без риска |
| A6 | zenity/stand-fixture приложения дают наблюдаемую identity на стенде (кейс autocorrect-blocklist-silent уже наблюдает fixture identity) — [VERIFIED: test/e2e/case_autocorrect.go:34-50, main.go:421-422] | Validation Architecture | Низкий: e2e-кейс опирается на доказанный канал |

## Open Questions

1. **Старт a11y-наблюдателя при выключенной автокоррекции (Pitfall 2).**
   - What we know: `ensureAppid` стартует наблюдателя только при непустом MACR-списке или включённой автокоррекции (verbatim выше). Пункт меню существует всегда; при выключенной автокоррекции кэш пуст навсегда → пункт постоянно неактивен.
   - What's unclear: принять ли как честное поведение (пункт про автокоррекцию; D-03 трактует пустой кэш как unknown), или расширить условие старта (всегда при проводке / при первом запросе меню).
   - Recommendation: стартовать наблюдателя при проводке демона безусловно (новый публичный метод актора, вызываемый рядом с `actor.SetMenuSync(menu)` в OnConn; отказ старта — существующий WARN-degradation, пункт остаётся неактивным). Экономика a11y-соединения мала (один сокет + регистрация), а пункт делает identity постоянной способностью демона. Если владелец/планировщик сочтёт излишним — fallback: документировать «пункт активен после первого включения автокоррекции».
2. **Сигнатура нового сеттера MenuSync (один метод vs несколько).**
   - What we know: состояния пункта — (name|unknown) × fresh × covered × full; существующие сеттеры атомарные по одному факту с dedupe.
   - Recommendation: один сеттер с явными полями (например `SetBlocklistApp(app string, fresh, covered, full bool)`) — одна точка dedupe, одна эмиссия на изменение; альтернатива (пара сеттеров) размазывает dedupe. Выбор за планировщиком.
3. **Формулировка label при неизвестной личности.**
   - What we know: D-03 — пункт виден, неактивен, без имени и без пояснительного текста («не пишем ничего»).
   - Recommendation: статическая метка «Добавить в чёрный список» (без имени) + enabled:false. Буквальное «ничего не пишем» в label сделало бы пункт пустой строкой — против духа «пункт виден»; трактовка «без имени приложения» как «метка без подстановки имени». Если владелец имел в виду именно пустой пункт — правка одной строки.
4. **Spec-документирование нового пункта.**
   - What we know: канонический макет меню пинен (287999c) и упомянут в SPEC (строка 520); фаза меняет макет; дисциплина вехи — «спека прежде кода» (прецедент D-55 фазы 9).
   - Recommendation: в план включить датированную ревизию-блок SPEC §4.3 (меню v2 → новый пункт + D-03/D-04) и CONFIG.md (apps_blocklist: «пункт меню пишет QuoteMeta-имя, дедуп, потолок-64-отказ видим в меню»; меню-раздел CONFIG.md:59) — до/вместе с кодом, по устоявшемуся шаблону дословных ревизий.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | сборка/тесты | ✓ | go1.27.1 (module `go 1.23`) [VERIFIED: `go build ./...` зелёный в этом сеансе] | — |
| golangci-lint v2 | зелёная итерация | ✓ (mise `[tools]`) | 2 [VERIFIED: mise.toml] | mise install |
| mise | задачи гейтов | ✓ | [VERIFIED: mise.toml задачи build/vet/lint/test/ci/tidy-diff] | — |
| godbus/yaml.v3 | транспорт/writer | ✓ (go.mod) | v5.2.2 / v3.0.1 | — |
| Живой GNOME-сеанс | e2e menu-blocklist (Event-клики) | ✓ (целевая машина) | GNOME 46 | e2e вне `mise run ci` (как все e2e-*) |
| a11y-шина + bridge-приложение (zenity-прецедент) | identity для e2e | ✓ [VERIFIED: case_autocorrect.go fixture-канал] | at-spi2 2.52 | юнит-корпус покрывает семантику без живой шины |
| ydotool | НЕ требуется (клики по D-Bus Event) | — | — | — |

**Missing dependencies with no fallback:** none
**Missing dependencies with fallback:** none

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `-race` (TDD red→green→refactor — инженерная директива 1) |
| Config file | none — go test + `.golangci.yml` (v2, strict); гейты через mise |
| Quick run command | `go test -race -count=1 ./internal/indicator/ ./internal/config/ ./internal/session/` |
| Full suite command | `mise run ci` (build+vet+lint+test) ; e2e отдельными задачами, вне ci |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| MENU-01.1 | Пункт виден/активен с точным именем; id 16 в layout под тумблером | unit (RED: `TestMenuLayout` wantKids) | `go test -race -count=1 ./internal/indicator/ -run TestMenuLayout` | ✅ (правится) |
| MENU-01.1 | Дельты сеттера (label/toggle-state/enabled), dedupe, empty-removed-slice | unit | `go test -race -count=1 ./internal/indicator/ -run 'TestMenuToggleAndVersionDeltas|TestMenuSignalFormPins'` | ✅ (расширяется) |
| MENU-01.1 | Клик id 16 → композиция: guards (unknown/covered/full), QuoteMeta, write, push, sync reload | unit (main_test.go) | `go test -race -count=1 ./cmd/goswitchd/` | ✅ |
| MENU-01.2 | Writer append: golden-дифф «+1 строка, комментарии целы», round-trip строгим Load, absent-ветка, дедуп-idempotent | unit | `go test -race -count=1 ./internal/config/ -run TestWriter` | ✅ (расширяется) |
| MENU-01.2 | Потолок 64: sentinel-отказ, оригинал не тронут | unit | `go test -race -count=1 ./internal/config/` | ❌ новый тест (Wave 0 не нужен — файл/паттерн есть) |
| MENU-01.4 | Имя с метасимволами (`org.example.App[1]`) — QuoteMeta + strict Load зелёный после записи | unit | `go test -race -count=1 ./internal/config/ -run TestWriter` | ❌ новый тест |
| MENU-01.3 | Свежесть: gain→fresh, real-FocusOut→unknown, гонка gain-позже-FocusIn, D-53 гейты байт-неизменны | unit (actor_test, fakeAppid + fakeMenuSync) | `go test -race -count=1 ./internal/session/` | ✅ (расширяется) |
| MENU-01.1-4 | Живой цикл: Event-клик id 16 → документ вырос ровно на паттерн → строгий Load → status; клик по отмеченному — no-op; автокоррекция молчит в заблокированном приложении (oracle-машина blocklist-silent) | e2e (live GNOME, вне ci) | `go run ./test/e2e -case menu-blocklist` (новая mise-задача) | ❌ новый кейс по образцу case_menu.go |

### Sampling Rate
- **Per task commit:** quick run (три пакета) + `golangci-lint run`
- **Per wave merge:** `mise run ci`
- **Phase gate:** `mise run ci` зелёный; e2e menu-blocklist — по живой приёмке фазы (UAT end-of-phase, конфиг `human_verify_mode: end-of-phase`)

### Wave 0 Gaps
- [ ] none — инфраструктура (testing, .golangci.yml, mise, e2e-каркас, фикстуры writer/menu/actor) существует; новые тесты ложатся в существующие файлы по готовым паттернам

## Security Domain

security_enforcement включён (config.json), ASVS level 1.

### Applicable ASVS Categories
| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | локальный user-демон, без аутентификации |
| V3 Session Management | no | — |
| V4 Access Control | no (n/a) | доступ = владение сессионной шиной; конфиг 0600 |
| V5 Input Validation | yes | `regexp.QuoteMeta` (метасимволы имени), строгий `KnownFields` (D-33), потолки 64 (`maxAutocorrectBlocklist`), pre-write валидация, atomic writer |
| V6 Cryptography | no | криптографии в фазе нет |

### Known Threat Patterns for goswitch
| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Regex-инъекция через имя приложения | Tampering | `QuoteMeta` до записи; валидация `regexp.Compile` в `Autocorrect.validate` как вторая ступень |
| ReDoS через паттерны blocklist | DoS | Go regexp = RE2 (линейный), потолок 64 записей, потолок уже в Validate (T-06-04-01) |
| Подмена identity на a11y-шине (любой сеансовый процесс может эмитить StateChanged с чужим bridge-путём) → клик записывает «не то» имя | Tampering (низкий) | сеанс локален и однопользовательский; потолок 64 ограничивает ущерб; запись — только явным кликом владельца; в журнале — только факт записи без содержимого секции (D-20/D-21) — документировать в SPEC-ревизии |
| Частичный запись конфига | Tampering | temp+rename атомарно; pre-write строгая валидация; отказ оставляет оригинал байт-нетронутым (T-07-03-01, TestWriter_BrokenSourceUntouched) |

## Sources

### Primary (HIGH confidence)
- internal/indicator/menu.go — реестр ID (56-91), removedProps Pitfall 7 (93-101), Menu (133-164), сеттеры (205-326), GetLayout/GetGroupProperties/AboutToShow/Event/dispatchClick (335-430), rootKidsLocked (444-457), itemPropsLocked (462-487), билдеры props (590-634), doubleUnderscores (645-651) — прочитано целиком в этом сеансе
- internal/appid/appid.go — wire-факты (30-50), сентинелы (52-74), focusGain/bridgeApp (141-171), Start (181-248), FocusedApp (253-258), run cache-keep (287-314)
- internal/config/writer.go — константы (25-48), SetAutocorrectEnabled/EnsureDocument (58-106), setDocumentToggle (116-139), buildFullDocument (166-184), upsertToggle (191-246), rewriteAtomically/writeAtomically (301-350)
- internal/config/config.go — потолок 64 (28), Validate (333-355), Autocorrect.validate (472-501)
- internal/config/load.go — строгий декодер (20-53)
- internal/config/watch.go — stat-опрос (13-21), Watcher/Reload/reparse (85-256)
- internal/session/actor.go — Status (1128-1163), StatusSnapshot (1179-1214), Options (607-643), MenuSync (1889-1895), ensureAppid-условие (2835-2850), applySnapshot fold (2141-2203), pushA11y lock-дисциплина (2250-2271), pushMenuSync (2273-2303), autoCorrectBoundary гейт (3322-3385), HandleLifecycle FocusOut (1272-1375), acFocusedApp (3519-3547)
- internal/indicator/indicator.go — Callbacks (98-107), attach/registerItem (224-299), ModeChanged, recoverMenuCall
- cmd/goswitchd/main.go — OnConn-проводка меню (230-307), newMenuToggles (346-392), configToggle.flip (475-494), startConfigWatcher/loadConfig (543-600)
- internal/config/writer_test.go — golden/idempotent/strict-load паттерны (97-206, 425-)
- internal/indicator/menu_test.go — TestMenuLayout wantKids (131-190), TestMenuMacroVisibility (199-245), fakeEmitter
- internal/session/actor_test.go — fakeAppid (3274+), fakeMenuSync (5003-5057)
- test/e2e/case_menu.go — пины ID (31-48), Event-канал, menuConfigDoc; test/e2e/main.go — реестр кейсов (410-430); case_autocorrect.go — blocklist-silent oracle
- .planning/phases/11-.../11-CONTEXT.md, 11-DISCUSSION-LOG.md — D-01..D-04, дискреции
- .planning/REQUIREMENTS.md — MENU-01; .planning/STATE.md — контекст фаз 9-10
- docs/SPEC.md — D-53 ревизия (485-510), сводка 2026-10-05 (264-303); docs/CONFIG.md — apps_blocklist контракт (90), тумблеры (59); docs/adr/ADR-005 — identity через AT-SPI-наблюдатель
- Живая база: `go build ./...` + `go test -race ./internal/config/ -run TestWriter...` — зелёные в этом сеансе

### Secondary (MEDIUM confidence)
- [at-spi2-core/NEWS (GitHub)](https://github.com/GNOME/at-spi2-core/blob/main/NEWS), [Table of AT-SPI events – a11y.org](https://www.a11y.org/) — state-changed:focused detail1=1/0 семантика; потеря-эмиссия исторически не универсальна (обосновывает ненадёжность loss-сигнала как единственного)
- [ubuntu/gnome-shell-extension-appindicator](https://github.com/ubuntu/gnome-shell-extension-appindicator) — dbusMenu.js: enabled:false → non-reactive актор, clicked не диспатчится (A1)

### Tertiary (LOW confidence)
- none (все LOW-уровневые допущения помечены [ASSUMED] в Assumptions Log)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — все компоненты in-repo, прочитаны; ноль новых зависимостей
- Architecture: HIGH — паттерны меню/writer/composition пинены корпусом фаз 7-8; fresh-механизм — рекомендация, деривативы прочитаны
- Pitfalls: HIGH — Pitfall 1 (кэш-keep) VERIFIED по исходнику; остальные — кодовые факты
- Свежесть/сходимость: MEDIUM — живое поведение a11y-gains (A2) не пробилось в этом сеансе

**Research date:** 2026-10-31
**Valid until:** 2026-11-30 (стабильная кодовая база; внешних источников почти нет)
