# Phase 7: Меню v2 и чёрный список автокоррекции - Research

**Researched:** 2026-10-02
**Domain:** DBusMenu v3 wire capabilities GNOME 46 (ubuntu-appindicators), persist-конфиг-тумблер (yaml.v3 Node round-trip + fsnotify self-write), blocklist-семантика D-53-ревизии (config schema + actor gate), e2e-репины, spec-delta/ADR-007 amendment
**Confidence:** HIGH (рендерер меню прочитан по исходникам расширения на этой машине; шина запробована живьём на работающем демоне; yaml/fsnotify/regex-механики доказаны пробами этой сессии; весь затронутый in-repo код прочитан)

## Summary

Фаза закрывает два UX-пробела фазы 6: автокоррекция без интерфейса и белый список,
требующий «наблюдаемой» идентичности (06-UAT Deferred). Оба решения владельца
зафиксированы в 07-CONTEXT.md дословно; research подтвердил реализуемость **без единой
новой зависимости**: DBusMenu v3 поверх существующего канала (quick 261001-fg3),
yaml.v3 Node round-trip для persist-тумблера, godbus Emit для динамических обновлений,
xdg-open для «Настроек…».

**Рендерер меню полностью верифицирован по источнику.** Расширение `ubuntu-appindicators`
(этот стол, файлы 2024-09) поддерживает ровно тот словарь свойств, который нужен меню v2:
`toggle-type` (`radio`/`checkmark`) + `toggle-state` → точки/галочки пунктов, `enabled` →
серые пункты, динамические обновления сигналами `ItemsPropertiesUpdated` (инкрементально)
и `LayoutUpdated` (полная перечитка; номер ревизии расширением НЕ сравнивается), подменю
(`children-display='submenu'`) и разделители. Критичная ловушка: `_updateLabel` вырезает
`_`-мнемоники регэкспом `/_([^_])/, '$1'` — метки горячих клавиш вида `shift_r` отрисуются
как «shiftr» (удвоение `__` спасает).

**Blocklist-ревизия D-53** переворачивает ровно один сегмент конъюнкции: white-list
(`app-not-listed`/`no-apps` отказники, «неизвестно → молчание») → blocklist («совпадение ⇒
запрет, неизвестно ⇒ пропуск»). Ролевый гейт, 25 мс контур, off-состояние и деградации
остаются байт-как-сегодня. Каскад последствий вскрыт и задокументирован: (1) активная
секция конфига — теперь просто `enabled: true` (пустой blocklist валиден и стреляет) ⇒
пороги обязаны присутствовать при enabled; (2) persist-тумблер требует, чтобы демон без
`-config` ПОДХВАТЫВАЛ дефолтный `~/.config/goswitch/config.yaml` — иначе «переживает
перезапуск» не выполняется на установленной системе (юнит стартует без `-config`, живой
факт этой сессии); (3) persist-запись через temp+rename порождает ровно один дебаунснутый
echo-reload (проба живьём) — безвреден.

**Primary recommendation:** spec-delta §11 + ADR-007 amendment до кода (прецедент 06-01);
затем schema+writer (`apps_blocklist`, persist-райтер с Node-раундтрипом и adopt-семантикой
дефолтного пути), меню v2 (радиопара EN/RU + чекмарк-тумблер + живые макро-пункты +
«Настройки…» + «О программе», обновления через ItemsPropertiesUpdated), реворк гейта актора
(arm: blocklist-матч отказывает, unknown пропускает; confirm: blocklist-матч на текущей
идентичности отказывает, unknown пропускает), репины юнит-корпуса и e2e (новый негативный
кейс blocklist на sentinel-идентичности фикстуры), релиз v1.1.0 по поезду Фазы 4.

<user_constraints>
## User Constraints (from CONTEXT.md — 07-CONTEXT.md, статус locked)

### Locked Decisions (решения владельца, дословно по смыслу)

- **Меню — переключение языка двумя отдельными пунктами EN и RU** «как во всех
  переключателях»: клик по пункту = включить этот язык; текущий отмечен.
- **Меню — вкл/выкл автокоррекции**: тумблер с записью в конфиг (переживает перезапуск),
  hot reload подхватывает.
- **Меню — версия приложения** (п. «О программе»).
- **Меню — клавиатурные макросы показывать сразу**: живые значения из конфига
  (tap-серия: 1 тап — язык, 2 — слово, 3 — фраза; word_layout_combo;
  mode_switch_chord), обновляются на reload конфига.
- **Меню — «Настройки…»**: запуск редактора для config.yaml (все 15 ключей конфига
  описаны в docs/CONFIG.md — «важных» для меню владелец не выделил, остальное покрывает
  редактор).
- **Белый список НЕ НУЖЕН** («белый список не нужен»): `autocorrect.apps` удаляется из
  схемы, кода, корпуса и документации.
- **Чёрный список обязателен**: ключ `autocorrect.apps_blocklist`, regex (подстрока;
  анкеровка `^…$` явная), совпадение = запрет автокоррекции.

### Locked semantics (ревизия D-53 — до кода, D-55-дисциплина)

- НОВОЕ ПРАВИЛО: автокоррекция исправляет ⇔ `enabled` ∧ живой role-гейт поля
  (текстовый ввод; пароли/терминалы — никогда) ∧ surrounding-text применим ∧
  уверенный вердикт ∧ длина ок ∧ приложение НЕ совпало с `apps_blocklist`.
- Неизвестная идентичность приложения НЕ запрет (не «молчание по D-53-unknown»):
  blocklist не может исключить то, чего не видно; полевой role-гейт остаётся фильтром
  безопасности. (Снимает и UX-пробел 06-UAT: zenity теперь стреляет.)
- Default off сохраняется (D-54): `enabled: false` + пустой blocklist.
- Ручная коррекция (Double/фраза/выделение) не ограничивается blocklist — явное
  действие пользователя.
- Спека правится spec-delta ДО кода: SPEC §11 (дополнение решения), ADR-007 amendment
  (ревизия D-53: white-list → blocklist), docs/CONFIG.md, README.

### Claude's Discretion

- Вид представления макросов в меню (инфо-пункты vs подменю) — по эргономике DBusMenu.
- Механика persist-тумблера: атомарная правка YAML с сохранением комментариев
  (yaml.v3 Node-раундтрип) vs перегенерация — по простоте и аудируемости.
- Формат regex-матчинга: MatchString-подстрока (RE2); валидация компиляции на Load.
- Способ запуска редактора: xdg-open по умолчанию, $EDITOR при наличии.

### Deferred Ideas (OUT OF SCOPE)

- Смешанный текст (посимвольная инверсия) — отдельный backlog-план (06-UAT Deferred).
- Показ наблюдаемой идентичности в status — backlog (06-UAT Deferred); blocklist делает
  его менее критичным.
- Терминальный класс, OSD, GUI настроек — как раньше (v2).

### Технические заметки CONTEXT (дано)

- DBusMenu-канал и супервизор троя готовы (quick 261001-fg3, self-heal) — расширение
  меню идёт поверх существующего com.canonical.dbusmenu.
- Полевой role-гейт и 25 мс контур не трогаются (контракт фазы 6).
- Регресс: матрица v4 строки autocorrect-* перезакрепить на blocklist-семантику;
  silence-matrix юнит-корпус переработать; observeFixtureApp-шов для white-list не нужен
  (blocklist-кейсам нужна известная идентичность — sentinel остаётся для негативов).
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SPEC §11 | spec-delta: ревизия решения по автокоррекции (blocklist) | Текст §11 прочитан (docs/SPEC.md:194-210): датированный append с HTML-комментарием-аудит-трейлом — форма для второй ревизии 2026-10-04 готова |
| MACR-ACL | Blocklist-ревизия D-53: apps → apps_blocklist (regex) | Схема/валидация/гейт/корпус/доки — Q4-Q6; слаги актора процитированы дословно (actor.go:92-102); все затрагиваемые ячейки перечислены |
| CORR-01..09 | Регресс ручной коррекции | Ручные пути (Double/фраза/выделение) гейт blocklist НЕ трогают (CONTEXT locked) — регресс = существующий корпус + матрица; manual-double round кейса fires сохраняется |
| SWCH-01..04 | Регресс переключения | flipTo не трогается; меню EN/RU идёт через тот же публичный шов (ToggleMode-образец, actor.go:902-906); матричные switch-строки остаются |
</phase_requirements>

## Project Constraints (from AGENTS.md / CONVENTIONS.md — CLAUDE.md отсутствует)

1. **Strict TDD** — red → green → refactor на каждое поведение; verify-блоки гоняют тесты на каждом шаге.
2. **Зелёная итерация** — `go build ./...`, `go test -race ./...`, `golangci-lint run` чистые на каждой итерации.
3. **mise, не make** — задачи сборки/тестов/e2e в `mise.toml` `[tasks]` (существующие: `ci`, `e2e-*`; `[VERIFIED: mise.toml:19-55]`).
4. **golangci-lint максимальный** с Фазы 1 (`.golangci.yml` v2, `linters.default: all`, depguard deny `unsafe`).
5. **Зависимости**: stdlib + godbus/dbus v5 + yaml.v3 + fsnotify только — фаза НЕ добавляет ни одной.
6. **Приватность** (D-20/D-21): слово не логируется; blocklist-паттерны — имена приложений, не слово (в статус не обязателен).
7. **Русские пользовательские доки** (CONFIG.md/README), spec-delta до кода, ADR-007 amendment (датированный append, не новый ADR).
8. **Git workflow**: работа в фазовой ветке (`gsd/phase-07-…`), изменения через PR.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Меню v2 (radio EN/RU, тумблер, макро-инфо, настройки, о программе) | Daemon (`internal/indicator` — Menu) | `internal/session` (состояние через швы), `cmd/goswitchd` (wiring) | Расширение существующего com.canonical.dbusmenu-объекта; состояние — через Callbacks-расширение + push-шов (Item.ModeChanged-прецедент) |
| Динамические обновления меню | Daemon (`internal/indicator` — Emit сигналов) | — | Emitter-шов уже есть (emitNewIcon); сигналы dbusmenu v3 отправляются той же коннекцией |
| Persist автокоррекции (тумблер → config.yaml) | Daemon (`internal/config` — writer) | `cmd/goswitchd` (resolve пути, apply) | Пакет config владеет схемой; writer — чистый файловый код + Node-раундтрип, TDD без D-Bus |
| Adopt дефолтного конфига при пустом -config | Daemon (`cmd/goswitchd` loadConfig + watcher) | — | Решение о том, ЧТО загружать — wiring-уровень; контракты Load/Watcher не меняются |
| Blocklist-схема (regex-валидация на Load) | Daemon (`internal/config`) | — | Strict decode D-33 + новые сентинелы по образцу существующих |
| Гейт срабатывания (arm/confirm) | Daemon (`internal/session/actor.go`) | — | autoCorrectBoundary/autoConfirm/acConfirmRefusals — точечная переработка существующих функций |
| e2e-кейсы (fires упрощается, новый blocklist-негатив) | e2e stand (`test/e2e/case_autocorrect.go`) | goswitchctl status (счётчики-оракул) | observeFixtureApp остаётся только негативам; оракулы без изменений |
| Spec-delta §11 + ADR-007 amendment + доки | docs | — | D-55-дисциплина: до кода |

## Research Findings

### Q1. DBusMenu v3 на этом столе: что рендерер реально умеет `[VERIFIED: исходники расширения + живые пробы]`

**Источник рендеринга** — `/usr/share/gnome-shell/extensions/ubuntu-appindicators@ubuntu.com/`
(GNOME 46, dbusMenu.js 2024-04-18, appIndicator.js 2024-09-10; расширение активно на этом
столе — трей-иконка goswitch рендерится именно им). Словарь свойств, которые расширение
знает и использует (dbusMenu.js:74-90, `PropertyStore.MandatedTypes`, дословно):

> `visible`(b), `enabled`(b), `label`(s), `type`(s), `children-display`(s),
> `icon-name`(s), `icon-data`(ay), `toggle-type`(s), `toggle-state`(i)

Дефолты: visible=true, enabled=true, label='', type='standard'. **Всё, что нужно меню v2,
покрыто; ничего сверх словаря расширение не прочтёт** (неизвестные свойства —
`Util.Logger.debug('Unhandled property change')`).

Покапабилити по пунктам (все — чтение исходника этой сессией):

| Возможность | Как выражается на проводе | Что делает рендерер | Проверка |
|---|---|---|---|
| Радио/чекмарк отметка | `toggle-type: "radio"\|"checkmark"` (s) + `toggle-state: 1\|0` (i) | `_updateOrnament` (dbusMenu.js:731-739): radio+state → `Ornament.DOT`, checkmark+state → `Ornament.CHECK`, иначе NONE | `[VERIFIED: dbusMenu.js:731-739]` |
| Серый (disabled) пункт | `enabled: false` | `_updateSensitive` → `setSensitive(false)` | `[VERIFIED: dbusMenu.js:787-789]` — уже используется для «Перечитать конфиг» без -config (живой дамп Q1b) |
| Скрытый пункт | `visible: false` | `_updateVisible` → `this.visible` | `[VERIFIED: dbusMenu.js:782-784]` |
| Смена метки на живую | `label` + сигнал | `_onPropertyChanged('label')` → `_updateLabel` | `[VERIFIED: dbusMenu.js:754-764]` |
| Подменю | `children-display: "submenu"` + дети в GetLayout | `PopupSubMenuMenuItem`; вложенность поддержана (NEED_NESTED_SUBMENU_FIX-ветка) | `[VERIFIED: dbusMenu.js:628-632, 699-707]` |
| Разделитель | `type: "separator"` | `PopupSeparatorMenuItem` | `[VERIFIED: dbusMenu.js:633-634]` |
| Иконка пункта | `icon-name`/`icon-data` | `_updateImage` | `[VERIFIED: dbusMenu.js:755-779]` |
| Инкрементальное обновление | сигнал `ItemsPropertiesUpdated(a(ia{sv}) a(ias))` | `_onPropertiesUpdated`: propertySet → widget живо обновляется; удаление свойства = propertySet(name, null) | `[VERIFIED: dbusMenu.js:486-505 + interfaces-xml/DBusMenu.xml]` |
| Полная перечитка макета | сигнал `LayoutUpdated(u revision, i parent)` | `_requestLayoutUpdate` → **всегда полный** `GetLayout(0, -1, ['type','children-display'])`; **номер ревизии не сравнивается** — любой сигнал = перечитка; после неё props каждого пункта перезапрашиваются `GetGroupProperties` | `[VERIFIED: dbusMenu.js:349-351, 479-484, 421-440]` |
| Обновление при закрытом меню | сигналы флагаются (`_flagLayoutUpdateRequired`/`_flagItemsUpdateRequired`), проигрываются при открытии (`set active`) | меню на открытии всегда свежее, если демон эмитил сигналы об изменениях | `[VERIFIED: dbusMenu.js:492-527]` |
| AboutToShow | вызывается на КАЖДОЕ открытие корня (`_onMenuOpened → sendAboutToShow`) и подменю; `true` → полная перечитка; после UNKNOWN_METHOD/FAILED выключается навсегда | текущий демон отвечает false — расширение продолжает спрашивать (отключение только по ошибке метода) | `[VERIFIED: dbusMenu.js:504-527, 876, 940-955]` |
| Клик | `Event(id, "clicked", v?i 0, u timestamp)` | `_onActivate` → EventAsync; открытие/закрытие подменю шлют события `opened`/`closed` | `[VERIFIED: dbusMenu.js:709-722, 530-543]` |

**Чего расширение НЕ вызывает** (реализовывать не нужно): `GetProperty`, `EventGroup`,
`AboutToShowGroup` — grep по dbusMenu.js даёт ноль вызовов; на живом демоне оба метода
отдают «Unknown / invalid method» без последствий (проба этой сессии). Сигналы
подписываются штатным Gio.DBusProxy (`DO_NOT_CONNECT_SIGNALS` не выставлен —
`[VERIFIED: dbusProxy.js:23,38]`), фильтр — по владельцу нашего well-known имени.

**Ловушка-мнемоники:** `_updateLabel` делает `label.replace(/_([^_])/, '$1')`
`[VERIFIED: dbusMenu.js:746]` — одиночное `_` в метке СЪЕДАЕТСЯ (GTK-мнемоника). Метка
«Слово: shift_r» отрисуется «Слово: shiftr». Все метки с именами клавиш обязаны удваивать
подчёркивания (`shift__r`) или транслировать имена. См. Pitfall 1.

### Q1b. Живые пробы работающего демона (эта сессия) `[VERIFIED: busctl, эта сессия]`

Демон запущен `/home/nil/.local/share/goswitch/goswitchd` — **БЕЗ `-config`**
(`ps aux`, pid 1007599), файла `~/.config/goswitch/config.yaml` НЕТ. Это делает вопрос
adopt-семантики дефолтного пути (Q3d) не теоретическим.

- `GetLayout(iias 0 10 2 …)` → `u(ia{sv}av) 1 0 0 3 …` — ревизия 1, корень с детьми
  1/2/3, каждый с тройкой (label, type, enabled); reload-пункт честно `enabled: false`
  (демон без -config) — совпадает байт-в-байт с кодом menu.go.
- `GetGroupProperties(aias 1 2 3 …)` → три `(ia{sv})` пары — форма ответа и порядок
  подтверждены.
- `AboutToShow(0)` → `b false`.
- `Event(999, "clicked", …)` → rc 0 (неизвестный id игнорируется — шов диспетчера жив).
- `EventGroup`/`GetProperty` → «Unknown / invalid method» (не экспортированы — и не нужны).

**Активация меню без ydotool/мыши** — `busctl/gdbus call … Event` — уже работает на
живом демоне и является оракулом для e2e-кейсов меню v2 (клик пункта = Event на id
пункта; режим/конфиг проверяются логом/статусом/файлом).

### Q2. Меню v2: состав и проводка состояния

Состав (CONTEXT locked + discretion):

| Пункт | Тип | Состояние/действие | Данные |
|---|---|---|---|
| EN / RU | два пункта, `toggle-type: "radio"` | `toggle-state` = текущий режим; клик = `flipTo(конкретный режим)` | режим — ModeDisplay-шов (Item.ModeChanged уже пушится из flipTo/syncMode, actor.go:785-799) |
| Автокоррекция | `toggle-type: "checkmark"` | `toggle-state` = `autocorrect.enabled`; клик = персист в config.yaml + apply | состояние — из снапшота конфига; обновление — на reload и на клик |
| Макросы (3–5 инфо-пунктов или подменю) | `enabled: false` (серые) | живые метки: «1× tap_key — язык», «2× — слово», «3× — фраза», «word_layout_combo», «mode_switch_chord» | из применённого конфига; обновление на reload |
| «Настройки…» | стандартный | клик = открыть config.yaml в редакторе (Q7) | путь конфига |
| «О программе» | стандартный или серый | клик = нотификация версии (notifyStatus-образец уже несёт Version) или серая строка «goswitch vX.Y.Z» | `actor.StatusSnapshot().Version` (D-37; `[VERIFIED: main.go:149-159]`) |
| Статус / Перечитать конфиг | без изменений | как сегодня | — |

Рекомендация discretion «инфо-пункты vs подменю»: **плоские серые инфо-пункты** —
владелец просил «показывать сразу»; подменю добавляет клик и `children-display`-цикл
событий, а плоский список из ~5 пунктов эргономичен и тривиально обновляем. Старый пункт
«Переключить раскладку» становится избыточным рядом с парой EN/RU — кандидат на вывод
(планировщик пинирует; CONTEXT прямого приказа не даёт).

**Проводка состояния (архитектура, по прецедентам in-repo):**

1. **Чтение** — свойства пунктов вычисляются в момент ответа (`GetLayout`/
   `GetGroupProperties`) из малого мьютекс-охраняемого снимка состояния меню (mode,
   autocorrect enabled, имена клавиш, версия) — прецедент `Item.properties()` под
   `it.mu`. Снимок наполняется: режим — из ModeChanged-пуша; конфиг — из applySnapshot;
   версия — при attach.
2. **Push** — на каждое изменение снимка демоn эмитит `ItemsPropertiesUpdated` со
   свежими `(id, {label?, toggle-state?})` затронутых пунктов (приемник флагает при
   закрытом меню и переигрывает при открытии — Q1). Emitter-шов уже есть
   (`connEmitter.Emit`, indicator.go:178-184; прецедент emitNewIcon). `LayoutUpdated` +
   бамп ревизии — только при изменении НАБОРА пунктов (его у v2 нет после первого
   построения; константа-ревизия menuRevision уходит в счётчик).
3. **Точка вызова push** — там же, где Item.ModeChanged вызывается сегодня (flipTo/
   syncMode — «observer-last»), плюс applySnapshot (reload конфига) и обработчик клика
   тумблера. Нужен второй наблюдательский слот рядом с ModeDisplay (или обобщение
   ModeDisplay до «display+menu» — планировщик пинирует; швы SetSwitcher/SetModeDisplay
   — образец).
4. **Клик EN/RU** — новый публичный метод актора `SwitchMode(target)` (обёртка
   `flipTo(target)` с его same-target no-op guard'ом — actor.go:1645; ToggleMode
   actor.go:902-906 — образец публичной обёртки).
5. **Recover-шим** `recoverMenuCall` уже окружает Event — расширяется на новые пункты
   бесплатно.

### Q3. Persist-тумблер: writer, echo, adopt-семантика

**Q3a. Формат правки — yaml.v3 Node round-trip (рекомендация).** Проба этой сессии
`[VERIFIED: /tmp/gsy-yamlprobe, go1.27.1 + yaml.v3 v3.0.1]`:

```text
D_DATA-q3a-START (вывод пробы: правка enabled: false→true в документе с комментариями)
# goswitch config
hotkeys:
  tap_key: shift_r # single tap switches layout
…
autocorrect:
  enabled: true # the toggle target — comments must survive
  min_word_len: 4
D_DATA-q3a-END
```

Заголовочный и строчные комментарии пережили round-trip; значение встало. Две
верифицированные детали: (1) `Encoder.SetIndent(2)` обязателен — дефолт yaml.v3 = 4
пробела, а примеры CONFIG.md/доков 2-пробельные; (2) обход обязан начинаться с
**DocumentNode → Content[0]** (у документа нет key/value-пар — первый обход моей пробы
мимо него прошёл и ничего не поменял). Пустая секция autocorrect / её отсутствие →
вставка целого маппинга-узла с 5 ключами (enabled + пороги 4/2.0/1.0 — см. Q4c).

**Q3b. Атомарность = temp + rename в том же каталоге.** Смотритель смотрит КАТАЛОГ
(watch.go:126-131), фильтрует по базовому имени и Write|Create|Remove|Rename. Проба
fsnotify этой сессии `[VERIFIED: /tmp/gsy-research-echo]`:

```text
D_DATA-q3b-START (temp+rename)
EVENT op=CREATE name=.config.yaml.tmp
EVENT op=WRITE    name=.config.yaml.tmp
EVENT op=RENAME   name=.config.yaml.tmp
EVENT op=CREATE   name=config.yaml
EVENT op=WRITE    name=config.yaml
(после паузы, in-place: EVENT op=WRITE name=config.yaml)
D_DATA-q3b-END
```

События .tmp отфильтруются по имени; CREATE/WRITE config.yaml свернутся дебаунсом 200 мс
в **ровно один** повторный парс только что записанного (валидного) документа → INFO
«config reloaded» + applySnapshot с теми же значениями. **Зацикливания нет** (запись
бывает только по клику), эхо безвредно. Рекомендация: после записи тумблер применяет
новое значение к актору СИНХРОННО (при -config — через `watcher.Reload()`, он же
синхронизирует с echo под reloadMu; watch.go:167-172), не дожидаясь эха.

**Q3c. Путь записи.** Явный `-config` → он. Без `-config` → дефолт
`~/.config/goswitch/config.yaml` (os.UserConfigDir; путь уже пинится selfcheck'ом —
04-02). Файла нет (сегодняшнее состояние живой машины) → первый тумблер СОЗДАЁТ документ.
Минимальный валидный документ (правило 03-02: Load ничего не доводит дефолтами) обязан
нести hotkeys (3 ключа — иначе ParseBinding("") валит Validate), timeouts (2 ключа —
нули вне диапазонов), correction (3 ключа — backspace_cap 0 вне [1,500]) и явную секцию
autocorrect с порогами (Q4c); macr можно опустить (zero = off валиден). Рекомендация:
генерировать ПОЛНЫЙ документ Defaults()-эквивалент с enabled в тумблерном значении —
самодостаточен, читается как эталон в CONFIG.md.

**Q3d. Adopt-семантика — критичное следствие (планировщик пинирует, см. Open Questions
1).** Юнит-инсталляция стартует демон БЕЗ `-config` (живой факт). Если демон по-прежнему
читает только явный `-config`, записанный тумблером файл никогда не прочитается ни при
перезапуске, ни при перезагрузке — locked требование «переживает перезапуск» не
выполняется. Рекомендация: `loadConfig("")` резолвит дефолтный путь; файл существует →
Load + watcher (hot reload); файла нет → Defaults() + watcher всё равно стартует на
дефолтном пути (тогда первый тумблер подхватывается эхом). Контракт «явный -config обязан
загрузиться» не меняется; «отсутствует = зелёные дефолты» (04-02) сохраняется для
отсутствия. Файл есть, но битый → отказ старта (та же D-33-видимость, что и явный
-config; тише нельзя — расхождение с диском невидимо).

### Q4. Схема: apps → apps_blocklist

**Q4a. Замена поля** (config.go:113-119): `Apps []string \`yaml:"apps"\`` →
`AppsBlocklist []string \`yaml:"apps_blocklist"\``; Options актора —
`AutoCorrectApps []string` → `AutoCorrectBlocklist []string` (actor.go:279, дословно
сейчас: `AutoCorrectEnabled bool / AutoCorrectApps []string / AutoCorrectMinWordLen int /
AutoCorrectMargin float64 / AutoCorrectFloor float64`). Strict decode (KnownFields,
load.go:28) сделает старые документы с `apps:` жирной ошибкой декодирования — это
правильное поведение (D-33), миграционная заметка в CONFIG.md обязательна («переименуйте
ключ; ключIntent другой — white-list не переносится в blocklist механически»). Живых
конфигов с `apps:` на этой машине нет (файла нет вовсе) — миграционный риск нулевой,
но у владельца документировать.

**Q4b. Валидация (новый код в Validate):**
- компиляция каждого паттерна `regexp.Compile` → сентинел
  `errAutocorrectBlocklistRegex = errors.New("must be a valid regular expression")`,
  обёрнутый с полем и индексом (D-33 «каждая ошибка называет поле»); компиляция в Load —
  единственное место, где битый regex виден до рантайма `[VERIFIED: regexp.Compile
  возвращает ошибку — проба этой сессии]`;
- пустой/whitespace-паттерн → отдельный отказ: пустая строка — валидный RE2 и
  **совпадает со всем** (`regexp.Compile("").MatchString("org.gnome.Zenity") == true` —
  проба этой сессии): молчаливый «блокируй везде» — это ловушка, а не фича;
- потолок списка 64 — переиспользование `maxAutocorrectApps` (переименовать в
  `maxAutocorrectBlocklist`); НЕусловно (DoS-потолок не зависит от enabled — прецедент
  T-06-04-01, config.go:291-294);
- семантика матчинга — RE2 `MatchString` = неподанкоренная подстрока
  (`"chrom"` матчит `org.chromium.Chromium`; `^org\.gnome\.` — префикс;
  проба этой сессии). RE2 линеен — катастрофический бэктрекинг невозможен по
  построению `[CITED: golang.org/s/re2syntax]`. Регистр чувствителен — имена
  bridge-namespace чувствительны и сами.

**Q4c. Активная секция меняет условие — каскад.** Сейчас пороги проверяются только при
`enabled && len(Apps)>0` (config.go:295-297: «enabled+empty-list валиден» — пустой
white-list никуда не стреляет). Под blocklist **пустой список = стреляет везде** —
активное состояние есть просто `enabled: true`. Следствия:
- условие в `Autocorrect.validate()` → `if !a.Enabled { return nil }`; нули порогов вне
  диапазонов [2,16]/positive → громкий отказ при enabled без порогов (в духе
  complete-document 03-02);
- writer тумблера (Q3c) и e2e-шаблоны обязаны нести пороги явно;
- юнит-ячейки config-корпуса «enabled+empty-list валиден с нулевыми порогами» (пин
  06-04) — перезакрепить: enabled без порогов теперь отказ, enabled с порогами и пустым
  blocklist — валидно и активно.
- Defaults() не меняется: enabled:false, blocklist nil, пороги 4/2.0/1.0 (D-54).

### Q5. Гейт актора: реворк arm/confirm

Точки (actor.go, всё прочитано): `autoCorrectBoundary` (1883-1941), `autoConfirm`
(2015-2070), `acConfirmRefusals` (2087-2107), слаги (92-102):

```go
// actor.go:92-102 — дословно
acReasonNoApps          = "no-apps"
acReasonNoCaps          = "no-caps"
acReasonAppUnknown      = "app-unknown"
acReasonAppNotListed    = "app-not-listed"
acReasonRoleUnknown     = "role-unknown"
acReasonRoleForbidden   = "role-forbidden"
acReasonRoleTimeout     = "role-timeout"
acReasonRecheckDisabled = "recheck-disabled"
acReasonPayloadStale    = "payload-stale"
acReasonAppChanged      = "app-changed"
```

**Arm-time (дешёвая половина, под мьютексом):** enabled → caps → идентичность
(`acFocusedApp`) → **blocklist** → детектор. Отличия от сегодня: (1) отказа
`no-apps`/`app-not-listed` больше нет — их место занимает новый слаг (рекомендация
`app-blocked`) при МАТЧЕ известной идентичности с любым паттерном; (2) «неизвестно»
(нет источника/нет фокус-события/ошибка) — **не отказ**: идентичность опускается
(payload.app = ""), детектор выполняется, payload вооружается; role-гейт на confirm
остаётся фильтром безопасности (CONTEXT locked). WARN-дисциплина `warnACApp`
(«app identity unavailable») в новой семантике спамит на легальном unknown —
рекомендация: WARN только на ОШИБКУ источника (источник есть, ответ сломался),
отсутствие идентичности — тихая норма или debug; планировщик пинирует.

**Confirm-time (acConfirmRefusals):** payload-stale остаётся как есть (CR-01 — буферная
геометрия). Проверка идентичности перерабатывается: **текущая** (на момент confirm)
идентичность разрешается; известно И матчит blocklist → отказ (`app-blocked`);
неизвестно → пропуск. Равенство `app != payload.app` (слаг `app-changed`, тест
actor_test.go:5393-5419) — рекомендация УДАЛИТЬ: его исходная задача (WR-01 — конъюнкция
для одного объекта в один миг) при live-role-запросе покрывается тем, что blocklist
оценивается на confirm-идентичности, а перенос фокуса ловится payload-stale
(FocusOut → HardReset; фаза-6 research Q5). Условное сохранение равенства сломало бы
locked «unknown не запрет» (armed-unknown + выученная к confirm идентичность →
мнимый mismatch). Альтернативная консервативная форма — в Open Questions 2.

**Ячейки юнит-корпуса (полный реестр):**

| Ячейка silenceMatrixCells (actor_test.go:5077-5127, 9 ячеек) | Судьба |
|---|---|
| role 40 password → role-forbidden (async) | без изменений |
| role 60 terminal → role-forbidden (async) | без изменений |
| role error → role-unknown (async) | без изменений |
| role deadline → role-timeout (async) | без изменений |
| **missing identity source → app-unknown** | **ПЕРЕВОРАЧИВАЕТСЯ**: становится fire-ячейкой (unknown пропускает, role ok → fired) |
| **unlisted app → app-not-listed** | **ЗАМЕНЯЕТСЯ**: ячейка blocklist-match → app-blocked (известное приложение с матчащим паттерном) |
| no surrounding-text cap → no-caps | без изменений |
| detector unsure → trigram-unsure | без изменений |
| short word → abstain-short | без изменений |

Смежные тесты: `TestAutoCorrect_FailClosedNoFailOpen` (5163-5174: «missing identity
source = TOTAL silence») — **заменяется инверсией** «unknown identity is not a
prohibition» (идентичность — единственный сегмент конъюнкции, чей unknown стал
пропуском; роль/капсы/детектор остаются fail-closed — это надо пинить ЯВНО, чтобы
направление не расползлось); `app-changed`-тест (5393-5419) — удаляется/заменяется
confirm-blocklist-ячейкой (armed → identity появился и матчит к confirm → отказ);
новые позитивные ячейки: blocklist не матчит → fires; unknown-at-arm + role-ok → fires;
known-not-blocked + role-ok → fires. `acOptions()` (4644-4652) — поле
AutoCorrectApps → AutoCorrectBlocklist с паттерном-не-матчем для fire-путей и
матчащим для негативов.

**Что НЕ меняется:** off-состояние (enabled:false → ноль счётчиков/записей, D-54;
проверено actor.go:1884-1886 — гейт стоит до всякого счёта), ручные пути
(Double/фраза/выделение — blocklist их не касается, CONTEXT locked), конвейер замены
(startRangeCorrection), 25 мс контур, warn-once по ролям, счётчики/статус-форма
(ctlsvc.go:156-162) — словарь skip-reasons расширяется картой автоматически.

### Q6. e2e: репины и новый кейс

- **autocorrect-fires** (case_autocorrect.go:453-499): `observeFixtureApp` +
  `acStartObserverRound` + fresh-fixture-хореография были нужны только ради
  white-list-гейта — теперь НЕ нужны: конфиг `enabled: true` + `apps_blocklist: []`,
  инжекция сразу. Идентичность фикстуры (deterministic `org.gtk.application.gs_e2e_password`,
  case_autocorrect.go:40-41) остаётся полезной только негативу. Manual-double round
  (CORR-01-регресс) сохраняется дословно.
- **autocorrect-password-silent** (539-598): ролевой негатив НЕ меняется
  (роль 40 не зависит от списков); оракул `ac_skip_role_forbidden=1` остаётся; конфиг
  упрощается до enabled+пустой blocklist. Тот же метод удаления хореографии.
- **autocorrect-terminal-silent** (684-716): sentinel-гуард `org.gnome.GoswitchE2eAbsent`
  (белый список «против всех») теряет смысл — конфиг: enabled + пустой blocklist;
  молчание терминала держится на no-caps (контекст wezterm 0x9 без
  surrounding-text-бита) и отсутствии роли — оба fail-closed-механизма не тронуты.
  Оракулы (fired=0, ноль correction-записей, статус) без изменений.
- **НОВЫЙ кейс `autocorrect-blocklist-silent`**: GTK4-фикстура, observed identity →
  `apps_blocklist: ["<identity>"]` (или анкерованная форма), инжекция слова с
  границей → поле дословно, fired=0, ровно одна абстенция `ac_skip_app_blocked`
  (имя слага пинирует планировщик), witness до=после. `observeFixtureApp` живёт РАДИ
  этого кейса (идентичность — цель blocklist, не догадка).
- **Матрица v4** (cases/matrix-v4.yaml:401-425): `autocorrect-fires` — config_base
  теряет `apps` (matrix.go:311-338 `matrixACConfigYAML`: строка
  `apps: ["org.gnome.Zenity"]` → `apps_blocklist: []`); шаги «x »+fresh можно
  упростить (идентичность больше не гейт) — решает планировщик, шаги безвредны;
  `autocorrect-off`/`autocorrect-oov` — без изменений (off — zero-поведение; oov —
  детекторная ветка). Регресс остальных 30+ строк — механический (config_base общий).
- **Новый живой proof меню** (ydotool-независимый, канал уже запробован живьём — Q1b):
  кейс/шаг шлёт `Event(id_EN, "clicked", …)` по шине → лог несёт mode-запись →
  `Event(id_toggle_ac)` → файл конфига изменён + ctl status показывает
  `autocorrect_enabled=true` → рестарт демона → включено пережило перезапуск (locked
  «переживает перезапуск» — автоматизированное доказательство). Это закрывает и
  validate-work без ручных кликов по трею.
- **«Настройки…»/редактор** — живой кейс по желанию (оракул: запуск процесса —
  pgrep/лог); разумный минимум — юнит-корпус лаунчера с фейк-раннером.

### Q7. «Настройки…»: запуск редактора

- `xdg-open` установлен (`/usr/bin/xdg-open 1.1.3` — `[VERIFIED: command -v]`);
  пользовательский `$EDITOR` пуст (вershе) и в systemd user unit не попадает — ориентир
  только xdg-open (он маршрутизирует text/* на gedit/gnome-text-editor GNOME 46).
  Discretion «$EDITOR при наличии» рекомендую НЕ реализовывать: демон без
  контролирующего tty не может нести TUI-редактор, а GUI-редактор через $EDITOR — редкий
  кейс; xdg-open покрывает. (Планировщик волен добавить опору $VISUAL/$ENV — кода на
  один if, но тестировать живьём нечем.)
- **Форма запуска — no-pipes (fork-shaped), прецедент in-repo дословно:**
  clipboard.go:141-150: «wl-copy is the fork-shaped exception: … a piped stdout/stderr
  would make Run() wait for that grandchild's pipe write-ends to close — a deadlock
  (live finding …)». xdg-open/редактор — та же морфология (графический потомок переживает
  родителя): `cmd := exec.CommandContext(ctx, "xdg-open", path)` с nil Stdout/Stderr,
  без piped Run. Таймаут-контекст ОБЯЗАН НЕ рвать уже запущенный редактор — рекомендация:
  `cmd.Start()` + `go cmd.Wait()` (reap) без дедлайна, ошибка Start → WARN (одна, по
  эпизоду).
- Файла ещё нет → сначала ensure-документ (Q3c), потом редактор — иначе редактор откроет
  пустой буфер, сохранение которого не пройдёт strict Load.
- Нужен файловый дескриптор путём: writer знает путь; лаунчер принимает путь — тот же
  резолвер Q3d.

## Standard Stack

### Core

| Library/Tool | Version | Purpose | Why Standard |
|--------------|---------|---------|--------------|
| godbus/dbus/v5 (уже в go.mod) | v5.2.2 | Меню-сигналы (ItemsPropertiesUpdated/LayoutUpdated) через существующий Emitter; Event-диспетчер | Сигнатуры сигналов по interfaces-xml/DBusMenu.xml `[VERIFIED: файл расширения]`; godbus.Emit считает сигнатуру рефлексией — корпус пинит `[VERIFIED: menu.go:41-49 прецедент]` |
| gopkg.in/yaml.v3 (уже в go.mod) | v3.0.1 | Node round-trip persist-письменного редактирования | Комментарии переживают round-trip `[VERIFIED: проба этой сессии]`; zero-dep |
| regexp (stdlib) | — | apps_blocklist паттерны | RE2: линеен, подстрока MatchString `[VERIFIED: проба]`; компиляция на Load |
| os/exec (stdlib) | — | xdg-open лаунчер | no-pipes fork-shaped прецедент clipboard.go |
| fsnotify (уже в go.mod) | v1.10.1 | эхо-релоад персистной записи | поведение на temp+rename пробовано живьём `[VERIFIED]` |

**Новых зависимостей НЕТ.** Инсталляция — ничего.

## Package Legitimacy Audit

> Новых внешних пакетов фаза не добавляет (все — уже в go.mod или stdlib). Seam
> `package-legitimacy` npm/pypi/crates-only — Go-верификация не требуется сверх того.

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| godbus/dbus/v5 | proxy.golang.org | v5.2.2 2025-12-29 | n/a | github.com/godbus/dbus | OK | Approved (уже в go.mod) |
| gopkg.in/yaml.v3 | proxy.golang.org | v3.0.1 2022-05-27 | n/a | github.com/go-yaml/yaml | OK | Approved (уже в go.mod) |
| fsnotify | proxy.golang.org | v1.10.1 2026-05-04 | n/a | github.com/fsnotify/fsnotify | OK | Approved (уже в go.mod) |
| xdg-open | Ubuntu 24.04 archive | 1.1.3 | distro | freedesktop.org xdg-utils | OK | Approved (установлен `[VERIFIED]`) |

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### System Architecture Diagram

```text
ТРЕЙ (gnome-shell AppIndicator agent)          ДЕМОН goswitchd
  открытие меню ──AboutToShow/GetLayout──────▶ Menu (internal/indicator)
  GetGroupProperties ◀──(id,{label,type,         │ снимок состояния (mutex):
                          enabled,toggle-*)      │  mode, ac_enabled, имена клавиш, версия
                          …})                    │
  клик пункта ──Event(id,"clicked")──────────▶ Callbacks:
                                                  SwitchMode(EN|RU) → actor.flipTo (ADR-006)
                                                  ToggleAutocorrect → config.Writer
                                                    → temp+rename config.yaml
                                                    → watcher.Reload()/applySnapshot
                                                    → echo-reload через 200 мс (безвреден)
                                                  Settings → ensure-doc → xdg-open (no-pipes)
  ◀──ItemsPropertiesUpdated(a(ia{sv})a(ias))── на каждое изменение снимка
      (закрыто → флаг, открыто → живой апдейт)

НАБОР НА ГРАНИЦЕ СЛОВА (не изменилось, кроме одного гейта)
  feedKey → PushFeed → autoCorrectBoundary:
    enabled? → caps? → идентичность (unknown ⇒ ПРОПУСК, было: молчание)
             → blocklist-матч? (матч ⇒ app-blocked) → детектор
    → arm → autoConfirm (вне мьютекса, 25 мс):
    live GetRole → text-класс? → payload-stale? → blocklist(текущая идент.)?
    → fired → startRangeCorrection (единственный конвейер, без изменений)
```

### Recommended Project Structure (additions only)

```
internal/config/writer.go          # persist: Node round-trip, temp+rename, ensure-document
internal/config/writer_test.go     # корпус: создать/вставить секцию/флип/комментарии/эхо-события
internal/indicator/menu.go         # +динамика: снимок состояния, сигналы, новые пункты/типы
internal/indicator/menu_test.go    # +корпус сигналов (fake Emitter, байт-формы)
internal/session/actor.go          # гейт: blocklist, unknown-пропуск, SwitchMode
internal/session/actor_test.go     # реворк silence-matrix + новые ячейки
internal/config/config.go          # apps_blocklist, валидация regex/пустых/потолок, ACTIVE-условие
cmd/goswitchd/main.go              # adopt дефолтного пути, wiring меню/тумблера/настроек
docs/SPEC.md §11                   # датированный append (третья ревизия)
docs/adr/ADR-007-*.md              # amendment-секция (не новый ADR)
docs/CONFIG.md / README.md         # apps_blocklist, меню, migration note
test/e2e/case_autocorrect.go       # упрощение fires/silents + autocorrect-blocklist-silent
test/e2e/case_menu.go              # (новый) Event-драйв меню: EN/RU, тумблер, персист-переживает-рестарт
```

### Pattern 1: Динамический пункт DBusMenu (обновление свойства)
**What:** пункт меняет toggle-state/label; демон эмитит ItemsPropertiesUpdated.
**When to use:** каждое изменение режима/конфига/клика тумблера.
**Example:**
```go
// Source: interfaces-xml/DBusMenu.xml + dbusMenu.js:_onPropertiesUpdated (прочитаны)
// сигнал: ItemsPropertiesUpdated(a(ia{sv}) updatedProps, a(ias) removedProps)
type acItemDelta struct {
	ID    int32
	Props map[string]dbus.Variant
}
type removedProps struct {
	ID    int32
	Names []string
}
// emit: e.Emit(menuPath, menuIface, "ItemsPropertiesUpdated",
//         []acItemDelta{{ID: menuIDACToggle,
//             Props: map[string]dbus.Variant{"toggle-state": dbus.MakeVariant(int32(1))}}},
//         []removedProps{})
// пустой removedProps обязан быть ПУСТЫМ срезом (не nil) — сигнатура a(ias) с нулевой длиной
```

### Pattern 2: Persist-правка с сохранением комментариев
**What:** флип одного скаляра в живом пользовательском YAML.
**Example:**
```go
// Source: проба этой сессии (/tmp/gsy-yamlprobe); порядок: DocumentNode → Content[0]
var root yaml.Node
if err := yaml.Unmarshal(raw, &root); err != nil { return err }
if len(root.Content) > 0 && !setEnabled(root.Content[0], "autocorrect", "enabled", on) {
	if err := insertAutocorrectSection(&root, on); err != nil { return err } // секции не было
}
var out bytes.Buffer
enc := yaml.NewEncoder(&out)
enc.SetIndent(2) // дефолт yaml.v3 = 4; доки и примеры 2-пробельные
if err := enc.Encode(&root); err != nil { return err }
_ = enc.Close()
// атомарно: temp в том же каталоге → chmod 0o600 (прецедент configFilePerm case_select.go:96)
// → os.Rename → watcher-эхо (ровно один дебаунснутый reload, безвреден — Q3b)
```

### Anti-Patterns to Avoid

- **Пипед stdout/stderr для xdg-open/редактора** — fork-shaped потомок держит write-ends
  и вешает Run (живой прецедент wl-copy 03-03). Только no-pipes или Start+Wait.
- **Полагаться на номер ревизии LayoutUpdated** — расширение его не сравнивает; но и
  эмитить ревизию неверно нельзя (другие реализации сравнивают) — бампить честно.
- **Гнать метки клавиш с одиночным `_`** — мнемоника вырезается рендерером; удваивать.
- **Тумблер = только запись файла без apply** — без -config echo-релоада не будет
  (watcher отсутствует... до Q3d-adopt), а с adopt'ом задержка 200 мс ощущаема; применять
  синхронно.
- **Оставлять в схеме `apps`** — strict decode превратит старый документ в отказ и без
  того; полу-поддержка (silent ignore) нарушает D-33.
- **Логировать паттерн blocklist в связи с конкретным словом** — паттерны можно, слово
  нельзя (D-20/D-21); слаг + при необходимости паттерн — норма.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Меню-виджеты/отметки | свой рендеринг, иконки отметок | toggle-type/toggle-state + enabled словаря dbusmenu | рендерер расширения умеет всё это (Q1) — на проводе только свойства |
| Динамическое обновление меню | пересоздание item/пере-регистрация SNI | ItemsPropertiesUpdated / LayoutUpdated сигналы | инкрементальный протокол уже в v3; пере-регистрация мигает иконкой |
| Правка YAML | строковые sed-подмены, регэксп по тексту | yaml.v3 Node round-trip | sed ломает комментарии/кавычки/отступы; Node верифицирован (Q3a) |
| Атомарная запись | самодельные lock-файлы | temp + os.Rename (тот же каталог) | атомарность rename в POSIX; эхо-события уже понятны смотрителю |
| Валидация regex | свой парсер/ограничитель | regexp.Compile на Load | RE2 линеен — ReDoS невозможен; ошибка называется полем |
| Запуск GUI-редактора | свой разбор ~/.desktop, gio-вызовы | xdg-open | задача xdg-utils; установлен; fork-shaped ноу-пайпс |
| Клик-оракул e2e | ydotool/мышь по трею | busctl/gdbus `Event` | канал запробован живьём (Q1b), детерминирован, headless-совместим |

**Key insight:** и меню, и персист, и blocklist собираются из уже пробитых железок
пунктиром — единственная новая логика это state-снимок меню, Node-писатель и перевернутый
сегмент гейта.

## Runtime State Inventory

> Не rename/refactor-фаза в строгом смысле, но схема конфига МЕНЯЕТСЯ (ключ apps →
> apps_blocklist) и появляется новый рантайм-артефакт (персистный конфиг-файл).

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | `~/.config/goswitch/config.yaml` — сегодня ОТСУТСТВУЕТ на машине (`[VERIFIED: ls]`); тумблер начнёт его создавать | none (создание — само поведение фазы) |
| Live service config | Демон pid 1007599 запущен БЕЗ -config (`[VERIFIED: ps]`); e2e-кейсы пишут свои -config в tmpDir (прецедент) | adopt-семантика Q3d решает чтение; рестарты кейсов без изменений |
| OS-registered state | systemd user unit goswitchd.service — argv без -config; юнит фаза не меняет (если только не решат добавить -config — НЕ рекомендуется, см. Q3d) | none |
| Secrets/env vars | none — новых нет; blocklist-паттерны — не секреты (имена приложений) | none |
| Build artifacts | none — новых генераторов нет; бинарник пересобирается обычным поездом; релиз v1.1.0 — тег/goreleaser (04-07 прецедент, .goreleaser.yaml готов) | тег в релизном плане фазы |

## Common Pitfalls

### Pitfall 1: Мнемоники съедают подчёркивания в метках клавиш
**What goes wrong:** метка «Слово: shift_r» отображается «Слово: shiftr»; пользователь
видит неверное имя клавиши.
**Why:** `_updateLabel` рендерера делает `label.replace(/_([^_])/, '$1')`
`[VERIFIED: dbusMenu.js:746]` — GTK-мнемоника.
**How to avoid:** при формировании меток удваивать `_` → `__` (транформация
`strings.ReplaceAll(name, "_", "__")` перед вставкой) или транслировать имена в
человекочитаемые («Right Shift»). Юнит-тест меток с `shift_r` обязателен.
**Warning signs:** скриншот меню от владельца; e2e не поймает (метки не читаются со
стенда).

### Pitfall 2: Активная секция без порогов — тихие нули или громкий отказ?
**What goes wrong:** `enabled: true` без порогов (старая валидация пропускала — список
пуст «нигде не стреляет»; новая — стреляет везде с min_word_len=0).
**Why:** условие ACTIVE меняется с «enabled && non-empty apps» на «enabled» (Q4c).
**How to avoid:** валидация при enabled требует пороги в диапазонах — нули вне [2,16]/
positive → громкий отказ с именем поля; writer/шаблоны всегда пишут пороги.
**Warning signs:** config-тест «enabled+empty-list валиден» без порогов красный — это
ожидаемый RED, не регресс.

### Pitfall 3: Пустой паттерн blocklist матчит всё
**What goes wrong:** `apps_blocklist: [""]` молча запрещает автокоррекцию везде.
**Why:** пустая строка — подстрока любой строки (проба этой сессии: true).
**How to avoid:** Validate отвергает пустые/whitespace-паттерны отдельным сентинелом;
CONFIG.md документирует.
**Warning signs:** юнит-тест пустого паттерна.

### Pitfall 4: Unknown-идентичность случайно остаётся запретом
**What goes wrong:** под новой семантикой коррекция «не стреляет» в приложениях без
AT-SPI — та же боль, что 06-UAT, только переименованная.
**Why:** half-переворот: слаги no-apps/app-not-listed удалены, но ветка
`acFocusedApp !ok → abstain` осталась.
**How to avoid:** RED-тест «identity unknown → fired при role-ok» пинит пропуск на arm;
confirm-пропуск пинится отдельной ячейкой. Направление остальных сегментов (роль/капсы/
детектор — fail-closed) пинится сохранёнными ячейками.
**Warning signs:** wezterm-кейс вдруг считает role-unknown-абстенции — механизм не
тронут (no-caps/роль отсутствуют), но прогонять его после реворка обязательно.

### Pitfall 5: Эхо-релоад гонит applySnapshot в момент чужой правки
**What goes wrong:** пользователь сохраняет свою правку конфига в ту же 200-мс окно, что
клик тумблера — две записи race'ят.
**Why:** тумблер пишет весь документ (Node-раундтрип читает диск на момент клика).
**How to avoid:** принять как задокументированное ограничение (last-writer-wins; окна
миллисекундные); НЕ строить merge-логику. Дебаунс уже сводит двойной парс к одному.
**Warning signs:** нет — риск принять и забыть (документировать в CONFIG.md).

### Pitfall 6: yaml Node-обход начинается не с документа
**What goes wrong:** правка «не применяется», комментарии сохраняются — тест зелёный на
флип без правки, если тест не читает результат.
**Why:** DocumentNode не имеет key/value-пар; обход `i += 2` мимо Content[0] молчит
(поймано в пробе этой сессии — первый вариант не менял файл).
**How to avoid:** walk от root.Content[0]; каждый тест писателя читает результат ОБРАТНО
через Load и утверждает значение.
**Warning signs:** writer-тест без обратного чтения.

### Pitfall 7: Signal-payload нил vs пустой срез
**What goes wrong:** `ItemsPropertiesUpdated` с nil-removed ломает сигнатуру a(ias) у
godbus-рефлексии (прецедент zero-Variant дисциплины menu.go:69-72).
**How to avoid:** пустые срезы-литералы в эмитах; корпус фейкового Emitter пинит байт-форму
обоих аргументов.
**Warning signs:** паника сигнатуры на первом эмите (только в живой шине, не в корпусе —
потому корпус обязателен).

## Code Examples

### Динамическое состояние пунктов (форма, по прецеденту menuProps)

```go
// Source: internal/indicator/menu.go (прочитан; расширение — дизайн по его формам)
// Снимок состояния меню: пишется wiring'ом/пушами, читается ответами GetLayout/Props.
type menuState struct {
	mu            sync.Mutex
	mode          string // "EN"/"RU" — из ModeChanged-пуша
	acEnabled     bool   // из applySnapshot/тумблера
	tapKey        string // из применённого конфига (для меток)
	wordCombo     string
	modeChord     string
	version       string // D-37, при attach
	revision      uint32 // бамп на любое изменение
}
```

### Событие клика (провод уже живой — Q1b)

```go
// busctl/gdbus оракул (headless, без ydotool):
// gdbus call --session -d org.djarvur.goswitch -o /Menu \
//   -m com.canonical.dbusmenu.Event 2 clicked "<0>" 0    # → ()
// Обе пробы этой сессии ответили rc 0; ItemActivationRequested демону не нужен.
```

### Гейт blocklist (форма, не реализация)

```go
// Source: actor.go:1883-1941 (прочитан) — точечная замена white-list-сегмента
app, appKnown := a.acFocusedApp() // unknown больше НЕ отказ
if appKnown && a.acBlocklistMatch(app) { // compiled []*regexp.Regexp, fold-кэш
	a.recordACAbstain(acReasonAppBlocked) // новый слаг вместо app-not-listed/no-apps
	return acPayload{}, false
}
// unknown ⇒ детектор выполняется, payload.app = "" — role-гейт на confirm держит безопасность
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Белый список (opt-in по приложениям, unknown→молчание) | Чёрный список (opt-out по паттернам, unknown→пропуск) | решение владельца 2026-10-04 (D-53 ревизия) | автокоррекция стреляет «везде, где безопасно»; UX-пробел 06-UAT закрыт |
| Статичное меню (ревизия-константа, AboutToShow=false) | Динамическое (снимок + сигналы) | эта фаза | меню отражает режим/конфиг вживую; ревизия — счётчик |
| Конфиг = только ручной файл | Конфиг = ручной файл + персист из трея | эта фаза | появляется писатель; adopt дефолтного пути |
| `no-apps`/`app-not-listed` слаги | `app-blocked` слаг | эта фаза | словарь skip-reasons меняется; e2e-оракулы перезакрепляются |

**Deprecated/outdated:**
- «autocorrect.apps» — удаляется из схемы/кодов/доков; strict decode сам отвергнет старые
  документы.
- «unknown → silence для идентичности» — остаётся верным ТОЛЬКО для роли/капсов/детектора.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | xdg-open маршрутизирует YAML на подходящий GUI-редактор на этом столе (сам xdg-open и его наличие верифицированы; целевое приложение редактора — нет) | Q7 | «Настройки…» откроет не тот редактор/архиватор — owner pixel-check при verify; замена на явный fallback тривиальна |
| A2 | Расширение ubuntu-appindicators остаётся рендерером трея к релизу (сегодня активен, файлы датированы 2024) | Q1 | Обновление расширения изменит поведение — но словарь dbusmenu v3 стабилен годами; корпус на wire-формах держит контракт демона |
| A3 | FocusOut → HardReset гарантированно ловит перенос фокуса в ролевом окне confirm (обоснование удаления app-changed) | Q5 | Если найдётся путь смены фокуса без FocusOut, fired уйдёт «в чужое поле» — но verify-ранг конвейера замены всё равно проверит поле перед Delete (слой защиты остаётся); консервативная альтернатива в OQ2 |
| A4 | Метки макросов строятся из имён конфига без локализации ключей (тексты меток — RU, имена клавиш — как в CONFIG.md) | Q2 | Владелец захочет другие тексты — правка строк, не архитектуры |
| A5 | Оракул «рестарт пережил включение» в e2e стабилен (рестарт демона с -config — обкатанный путь кейсов) | Q6 | Флейки рестарта — лечатся существующими маркерами ожидания (configLoadedMark и пр.) |

## Open Questions

1. **Adopt дефолтного конфига при пустом -config (Q3d) — ПИНИРОВАТЬ ПЕРВЫМ.**
   - Что мы знаем: юнит стартует без -config (живой факт); locked «переживает перезапуск»
     невозможен, если демон не читает дефолтный путь; контракт 04-02 (инсталлятор НЕ
     генерирует конфиг; отсутствие = зелёные дефолты) остаётся совместимым.
   - Что неясно: принимает ли владелец (а) adopt+watch дефолтного пути с отказом старта
     на битом файле (рекомендация), (б) adopt без watch, (в) toggle только при -config
     (пункт серый без файла — не выполняет locked).
   - Рекомендация: (а); если владелец против — план меняет поведение тумблера, не гейта.
2. **Confirm-идентичность: строгая форма (рекомендация) vs консервативная.**
   - Что мы знаем: CONTEXT зафиксировал только «blocklist-match ⇒ отказ; unknown ⇒
     пропуск»; равенство payload.app (app-changed) не упомянуто.
   - Что неясно: держать ли дополнительно равенство «armed app == confirm app» (когда
     оба известны) — консервативная форма сохраняет WR-01-дословность, но ломает сценарий
     «unknown на arm → выучили к confirm» (мнимый mismatch ⇒ отказ вопреки locked).
   - Рекомендация: строгая форма (blocklist на текущей идентичности, равенство удалить);
     conservative-вариант с условием «оба известны И различны» допустим, если владелец
     захочет буквальный WR-01.
3. **Состав меню: судьба старого пункта «Переключить раскладку» и вид «О программе»**
   (кликабельный пункт с нотификацией vs серая строка с версией).
   - Рекомендация: toggle-пункт вывести (радиопара его заменяет), «О программе» —
     кликабельный с нотификацией (переиспользует notifyStatus с Version); оба решения
     косметические, планировщик пинирует по эргономике.
4. **Слаг нового отказа**: `app-blocked` (рекомендация, симметрично role-forbidden) —
   имя пинируется в плане и в e2e-оракуле.
5. **Релиз v1.1.0**: ROADMAP-цель фазы включает выпуск. Поезд готов (goreleaser, семвер
   теги, 04-07 прецедент «тег = последний план»). Рекомендация: финальный план фазы —
   релизный (версия-бамп, доки, тег, acceptance по 04-07-форме).

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| ubuntu-appindicators extension | рендеринг меню | ✓ активен | GNOME 46 (файлы 2024) | SNI icon-only деградация (уже встроена) |
| com.canonical.dbusmenu канал | всё меню | ✓ живой (quick 261001-fg3) | dbusmenu v3 | — |
| xdg-open | «Настройки…» | ✓ | 1.1.3 | $VISUAL/$EDITOR (не рекомендуется, Q7) |
| yaml.v3 Node API | persist | ✓ v3.0.1 в go.mod | v3.0.1 | перегенерация документа (теряет комментарии — не рекомендуется) |
| fsnotify | эхо-релоад | ✓ | v1.10.1 | polling (STACK-альтернатива) |
| gdbus/busctl | e2e-оракулы меню | ✓ (часть ОС) | systemd 255 | — |
| GTK4-фикстура + a11y | e2e-кейсы автокоррекции | ✓ (фаза 6) | 4.14.5 / at-spi 2.52.0 | — |

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** none required.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + golden/wire-формы; strict TDD (`tdd_mode: true`) |
| Config file | `mise.toml` + `.golangci.yml` |
| Quick run command | `mise run ci` (build+vet+lint+test -race) |
| Full suite command | `mise run ci` + live-кейсы (`e2e-matrix-v4`, `e2e-autocorrect-*`, новые меню-кейсы) на столе владельца |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SPEC §11 | ревизия blocklist в спеке/ADR до кода | doc-gate (grep) | grep-гейты плана A: §11 датированный append, ADR-007 amendment-секция, CONFIG.md apps_blocklist | ❌ Wave 0 (план A) |
| MACR-ACL | схема: apps_blocklist, regex-валидация, пустой паттерн, потолок, ACTIVE-условие | unit | `go test -race ./internal/config/ -run Autocorrect` | ✓ (расширить; RED-стаб по прецеденту 06-04) |
| MACR-ACL | writer: создать/вставить секцию/флип/комментарии/0o600/обратное чтение | unit | `go test -race ./internal/config/ -run Writer` | ❌ Wave 0 |
| MACR-ACL | гейт: blocklist-матч отказ; unknown-пропуск; confirm-форма; off-нулевое поведение | unit | `go test -race ./internal/session/ -run AutoCorrect` | ✓ (реворк ячеек Q5) |
| MACR-ACL | блок-негатив живьём (identity-цель, fired=0, app-blocked=1, witness) | e2e (live) | `mise run e2e-autocorrect-blocklist-silent` (новый) | ❌ Wave 0 |
| MACR-ACL | регресс: fires упрощённый + password/terminal silent | e2e (live) | `mise run e2e-autocorrect-fires` (+ -password-silent, -terminal-silent) | ✓ (репины конфигов) |
| CORR-01..09 | ручные пути не ограничены blocklist; матрица | unit+matrix | `go test -race ./internal/session/` + `mise run e2e-matrix-v4` | ✓ |
| SWCH-01..04 | EN/RU пункты через flipTo; радиометки; матрица switch | unit+e2e | `go test -race ./internal/indicator/` + новый `e2e-menu-v2` (Event-драйв: клик EN → mode-запись; клик тумблера → файл+статус; рестарт → пережил) | ❌ Wave 0 |
| Меню v2 | wire-формы пунктов (radio/checkmark/disabled/метки с `__`), сигналы | unit (fake Emitter/экспортер) | `go test -race ./internal/indicator/` | ✓ (расширить корпус menu.go) |
| Меню v2 | «Настройки…» лаунчер (no-pipes, ensure-doc) | unit (fake runner) | `go test -race ./cmd/goswitchd/ ./internal/config/` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `mise run ci`
- **Per wave merge:** `mise run ci` + соответствующие live-кейсы волны
- **Phase gate:** полный зелёный набор + `e2e-matrix-v4` двойной свежесессионный прогон
  (D-48-форма) + verify-work UAT владельца (скриншот меню — Pitfall 1)

### Wave 0 Gaps
- [ ] `internal/config/writer.go` + корпус (Node round-trip, echo-события в синтетике)
- [ ] `internal/indicator/menu.go` динамика + корпус сигналов
- [ ] реворк silence-matrix + новые ячейки (до гейта — RED)
- [ ] `test/e2e/case_menu.go` + mise-задачи `e2e-menu-v2`, `e2e-autocorrect-blocklist-silent`
- [ ] doc-gates плана A (спека/ADR/CONFIG.md/README)

## Security Domain

`security_enforcement: true`, ASVS L1 (`.planning/config.json`).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | нет auth-поверхностей |
| V3 Session Management | no | — |
| V4 Access Control | yes | безопасность срабатывания остаётся на конъюнкции: роль (живой GetRole) + caps + детектор — fail-closed; blocklist — UX-опт-АУТ, НЕ замена ролевого гейта; default off (D-54) неизменен |
| V5 Input Validation | yes | regex-компиляция на Load (отказ с именем поля); потолок 64 паттерна (условно); пустые паттерны отвергаются; writer читает результат обратно через Load |
| V6 Cryptography | no | — |
| V14 Config | yes | strict decode D-33 (старый `apps` → громкий отказ — документированная миграция); last-good reload D-32; файл 0o600 (прецедент configFilePerm) |

### Known Threat Patterns for меню + blocklist

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| ReDoS через blocklist-паттерн | DoS | RE2 линеен по построению `[CITED: golang.org/s/re2syntax]`; потолок количества; компиляция на Load |
| Тумблер «прощает» парольные поля (пользователь включает и ждёт безопасного везде) | Elevation/Deception | role-гейт первичен (пароли 40/серые GTK3 — конъюнкция); CONFIG.md/README проговаривают явно |
| Пункт меню запускает произвольный редактор | Tampering | только xdg-open на ПИНned путь конфига; no-pipes; аргумент — путь, не shell |
| Слова пользователя в логах через новые записи | Information disclosure | новые слаги/записи не несут слова (прецедент recordACAbstain); тест на приватность продолжается |
| Гонка двух писателей конфига (тумблер vs пользователь) | Tampering | last-writer-wins задокументирован; атомарный rename не даёт битых файлов |

## Implications for Planning

**Рекомендуемая нарезка** (spec-delta первым — D-55/06-01-прецедент):
1. **План A (docs):** SPEC §11 датированный append (третья ревизия) + ADR-007 amendment
   (white-list → blocklist, unknown-семантика идентичности, confirm-форма) + CONFIG.md
   (apps_blocklist, миграция, меню) + README. SPEC-DELTA-OK гейт до кодовых планов.
2. **План B (config):** schema swap (RED-стаб), валидация (regex/пустые/потолок/
   ACTIVE-условие), writer (Node round-trip + temp+rename + ensure-document) + adopt
   дефолтного пути в loadConfig/watcher (OQ1 — сначала checkpoint владельца).
3. **План C (menu):** динамическое меню (state-снимок, сигналы, радиопара/чекмарк/инфо/
   настройки/о-программе), SwitchMode-шов актора, wiring.
4. **План D (gate):** реворк arm/confirm (blocklist, unknown-пропуск, слаг), реворк
   silence-matrix + новые ячейки, Options-поле.
5. **План E (e2e):** упрощение fires/silents, новый blocklist-silent, menu-v2 кейс
   (Event-драйв: EN/RU, тумблер-персист, пережил рестарт), mise-задачи, матричные репины.
6. **План F (release):** версия/доки/тег v1.1.0 по поезду 04-07 (acceptance-форма).

**Риски:** R1 adopt-семантика требует решения владельца (OQ1 — поставить checkpoint в
план B); R2 метки-мнемоники (Pitfall 1 — юнит-тест+UAT-скриншот); R3 confirm-форма (OQ2
— пин в плане D); R4 старые документы с `apps` перестают грузиться (миграция задокументирована,
живых нет).

**Прецеденты для задач:** RED-стаб схемы (06-03/06-04), warn-once эпизоды (actor.go),
wire-корпус menu.go, Emitter fake (indicator_test), Event-оракул (эта сессия, Q1b),
startAutocorrectConfigDaemon (case_autocorrect.go:156-175), релизный поезд (04-07).

## Sources

### Primary (HIGH confidence — чтение/пробы этой сессии)
- `/usr/share/gnome-shell/extensions/ubuntu-appindicators@ubuntu.com/dbusMenu.js` (прочитан:
  словарь свойств 74-90; орнаменты 731-739; сигналы 479-527; layout 349-351/370-440;
  AboutToShow 504-527/876/940-955; метки 746; подменю/разделители 628-634/699-707) +
  `interfaces-xml/DBusMenu.xml` (сигнатуры методов/сигналов) + `dbusProxy.js` (подписка
  сигналов 23/38) — рендерер этого стола
- Живые пробы: busctl GetLayout/GetGroupProperties/AboutToShow/Event на
  org.djarvur.goswitch (дампы процитированы в Q1b); EventGroup/GetProperty — не
  экспортированы; `ps` (демон без -config), `ls ~/.config/goswitch/` (файла нет)
- Проба fsnotify `/tmp/gsy-research-echo` (temp+rename vs in-place события — Q3b)
- Проба yaml.v3 Node round-trip `/tmp/gsy-yamlprobe` (комментарии, флип, DocumentNode,
  SetIndent(2); regex: пустой паттерн/подстрока/анкер/Compile-ошибка — Q3a/Q4b)
- In-repo (прочитаны этой сессией): internal/indicator/{menu,indicator,supervisor}.go,
  internal/config/{config,load,watch}.go, internal/session/actor.go (85-110, 255-360,
  1839-2124) + actor_test.go (4644-4652, 5073-5213, 5393-5419), internal/clipboard/
  clipboard.go (85-171), cmd/goswitchd/{main,version}.go, internal/ctlsvc/ctlsvc.go
  (статус-токены), test/e2e/case_autocorrect.go, test/e2e/matrix.go (296-347),
  test/e2e/cases/matrix-v4.yaml, test/e2e/case_select.go:96, docs/SPEC.md §11,
  docs/adr/ADR-007, docs/CONFIG.md, mise.toml, .planning/{config.json,STATE.md,
  phases/06-*/06-RESEARCH.md,06-UAT.md}

### Secondary (MEDIUM confidence)
- [freedesktop DBusMenu spec](https://developer.ubuntu.com/api/apps/qml/sdk-15.04/DBusMenu/) —
  протокольный контекст (первичен здесь — исходник рендерера)
- [Go regexp/syntax (RE2)](https://golang.org/s/re2syntax) — линейность/подстрока
  (свойство подтверждено и пробой)

### Tertiary (LOW confidence)
- Целевое приложение xdg-open для text/x-yaml на этом столе (A1) — пинировать при
  реализации/UAT

## Metadata

**Confidence breakdown:**
- DBusMenu capabilities: HIGH — исходник рендерера этого стола прочитан + живые дампы
- Persist/echo/regex-механики: HIGH — пробы этой сессии
- Blocklist-реворк гейта/корпуса: HIGH — весь код и тесты прочитаны, ячейки перечислены
- Editor launch: MEDIUM-HIGH — xdg-open верифицирован, целевой редактор — A1
- Adopt-семантика: рекомендация HIGH, решение — владельца (OQ1)

**Research date:** 2026-10-02
**Valid until:** 2026-11-01 (стабильный домен; расширение GNOME — по факту обновления)
