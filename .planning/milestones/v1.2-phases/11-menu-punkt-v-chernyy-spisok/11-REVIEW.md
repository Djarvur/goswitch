---
phase: 11-menu-punkt-v-chernyy-spisok
reviewed: 2026-11-01T00:57:41Z
depth: deep
files_reviewed: 21
files_reviewed_list:
  - cmd/goswitchd/main.go
  - cmd/goswitchd/main_test.go
  - internal/appid/appid.go
  - internal/appid/appid_test.go
  - internal/appid/export_test.go
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/config/load.go
  - internal/config/writer.go
  - internal/config/writer_test.go
  - internal/indicator/indicator.go
  - internal/indicator/menu.go
  - internal/indicator/menu_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - skills/goswitch-config/SKILL.md
  - test/e2e/case_menu.go
  - test/e2e/case_menu_blocklist.go
  - test/e2e/case_menu_blocklist_test.go
  - test/e2e/main.go
findings:
  critical: 0
  warning: 4
  info: 2
  total: 6
status: issues_found
---

# Phase 11: Code Review Report

**Reviewed:** 2026-11-01T00:57:41Z
**Depth:** deep
**Files Reviewed:** 21
**Status:** issues_found

## Summary

Ревью поверхностных изменений фазы 11 (диапазон `fac7c89..43346ea`, +3415/−99): Round 2 карве-аут `flipAftermathWindow` в `internal/session`, persist-примитив `AppendAutocorrectBlocklist` в `internal/config`, машина свежести и пункт меню id 16 в `internal/indicator`/`internal/appid`/`internal/session`, клик-композиция `blocklistToggle.add` в `cmd/goswitchd`, e2e-кейс `menu-blocklist` в `test/e2e`.

Методология: полный трассинг диффа попроизводственно (не по SUMMARY), межмодульный анализ цепочки клика (dbusmenu Event → dispatchClick → композиция → writer → fold → сеттер меню), анализ порядка захвата мьютексов (actor.mu → menu.mu против обратного пути через колбэк клика — инверсии нет: `Event`/`dispatchClick` зовут колбэки БЕЗ `m.mu`), контраст `focusGain`/`focusLoss`, сверка с решениями D-01..D-04 из 11-CONTEXT.md и контрактами SPEC/CONFIG. Дополнительно ревьюером независимо перепрогнаны `go build ./...` и `go test -race -count=1` по всем шести затронутым пакетам — зелёные (подтверждение заявлений SUMMARY).

Контракты D-01..D-04 выдержаны: D-01 (галочка covered, клик no-op — гвард + дедуп писателя, двойной пояс), D-02 (props из снапшота AT REQUEST TIME, живых a11y-запросов нет), D-03 (unknown — статичная метка disabled без имени и пояснений; двойной барьер `blocklistStateLocked`/гвард композиции), D-04 (« — язык» в `macroChordWhat`, обоих README, пин в `TestMenuLayout`; grep по репозиторию старой формулировки — ноль). Потолок 64 сведён на `config.MaxAutocorrectBlocklist`, локальная константа 11-02 удалена. Писатель соблюдает D-20/D-21 и дисциплину «отказ = байт-нетронутый файл» на всех ветках (dedup-выход до валидации — корректен: записи нет). Раскладка id 1–15 не тронута, id 16 добавлен в диапазон `GetGroupProperties` (Pitfall 5 закрыт) и в `rootKidsLocked` сразу после ACToggle — по дискреции владельца.

Критических дефектов не найдено. Четыре предупреждения — все класса «узкая гонка/гранулярность атрибуции» в свежепосаженной freshness-машине и aftermath-карве-ауте, плюс один риск вакуумного прохода e2e-оракула D-01. Оба Info — сопровождаемость. Блокеров для фазовой верификации нет; WR-01/WR-02/WR-03 рекомендуется зафиксировать как известные хвосты или закрыть точечными правками в следующей итерации.

## Warnings

### WR-01: Aftermath-карве-аут глотает РЕАЛЬНЫЙ Reset при аномальной смерти контекста в окне 2 с после флипа

**File:** `internal/session/actor.go:1938-1963` (ветка 1 `verifyDeadOnContextDeath`), `internal/session/actor.go:1975-1988` (`armFlipAftermath`)
**Issue:** Карве-аут атрибутирует эхо по единственному признаку — «Reset прибыл, пока `flipAftermathStop != nil`», т.е. в течение 2 с после любого успешного флипа. Дизайн предполагает, что настоящая смена контекста всегда начинается с реального FocusOut (ветка default гасит окно). Но при аномальной смерти контекста БЕЗ FocusOut (SIGKILL приложения, выброс ibus-ом мёртвого контекста) реальный Reset нового контекста попадает в ветку 1: класс переживает его и наследуется ЧУЖИМ контекстом — следующая коррекция в другом приложении может ехать на чужой verify-класс (fail-open). Дополнительно окно однолюбецкое: первый echo-Reset тратит его (`cancelFlipAftermath`, строка 1950), поэтому клиент, шлющий ДВА Reset на один ре-минт (disconnect + reconnect), получает стирку на втором — intermittent-рецидив именно того бага, что чинит Round 2. На реальном клиенте (журнал green106) эхо одиночное, но контракт «один Reset на ре-минт» нигде не закреплён.
**Fix:** Сузить атрибуцию эха тем же приёмом, что у flap-окна (план 10-11): в момент флипа снапшотить эмиттер (`a.emitter()` — форма `flapEng`) и тратить aftermath только на Reset с того же пере-минченного семейства объектов/соединения; либо разрешить окну поглотить серию Reset до истечения дедлайна (не гасить слот в ветке 1, а лишь отмечать потраченное эхо до expiry). Минимум — задокументировать оба хвоста (аномальная смерть в окне; второй Reset) как принятые рядом с `flipAftermathWindow`.

### WR-02: Деарм свежести атрибутирует blur по имени приложения, а не по отслеживаемой поверхности

**File:** `internal/appid/appid.go:382-390` (loss-ветка `run`), `internal/appid/appid.go:183-198` (`focusLoss`)
**Issue:** `mine := lossApp == o.app` сравнивает только identity-строку, хотя наблюдатель уже кэширует точный адрес поверхности последнего gain (`o.focusSender`, `o.focusPath`, строки 136-137). Blur ЛЮБОЙ поверхности того же приложения (второе окно, другой виджет a11y-дерева) совпадает по имени: если такой blur приходит после gain той же identity (порядок a11y-событий не гарантирован для разных отправителей — ровно та гонка, которую 11-05 фиксировала на engine-слое), freshness деармится у живой поверхности. Пункт падает в unknown (D-03-направление — fail-closed, мусор не пишется), но состояние ложно до следующего gain. Тесты (`testLossOwnBlur`) гранулярность не пинят.
**Fix:** Сравнивать `sig.Path == o.focusPath && sig.Sender == o.focusSender` (оба уже под `o.mu` в этой ветке) вместо сравнения имён; identity-сравнение оставить запасным для форм, где адрес недоступен. Пин на «blur другого окна того же приложения не деармит» добавить в `TestObserver_LossHook`.

### WR-03: `EnsureAppid` расходует одноразовый старт наблюдателя на каждом старте демона — неудачный ранний дайл убивает пункт (и ленивый старт MACR/autocorrect) до рестарта

**File:** `internal/session/actor.go:3141-3153` (`startAppidOnce` — флаг ставится ДО попытки), `cmd/goswitchd/main.go:284` (безусловный вызов на каждом OnConn)
**Issue:** Одноразовая деградация «один WARN, без retry» — до-фазовый контракт 07-04, но фаза 11 расширила её радиус: раньше старт-попытка случалась только когда per-app-фича реально понадобилась (ленивое условие в `ensureAppid`), теперь `EnsureAppid` зовётся безусловно на каждом подключении, в том числе на D-54-дефолтной установке и на раннем этапе сессии. Если a11y-шина в этот момент ещё не отвечает (холодный старт сессии, dbus-активация at-spi2-core не успела/отсутствует), `appidStarted` защёлкивается навсегда: пункт «в чёрный список» мёртв до рестарта демона, и — главное расширение — ПОЗДНЕЙШИЙ ленивый старт для MACR/autocorrect, который раньше произошёл бы при реальной нужде, тоже заблокирован тем же флагом. Вероятность на целевой системе низкая (at-spi dbus-активируем), но отказ тихо превращает фичу в тыкву на весь lifetime процесса.
**Fix:** При неудаче старта не защёлкивать `appidStarted` (оставить false) — предупреждение уже однократно через `appidWarned`, повторные попытки естественным образом идут по fold-пути `ensureAppid` и не образуют «retry loop»; либо защёлкнуть, но переоткрыть флаг при следующем успешном `OnConn`. Минимум — покрыть хвост тестом и записать в известные хвосты фазы.

### WR-04: e2e-оракул D-01 не отличает covered-no-op от отказа по unknown-гварду — проверка может пройти вакуумно

**File:** `test/e2e/case_menu_blocklist.go:397-415` (`blocklistNoopRound`)
**Issue:** Оракул «повторный клик байт-идентичен» выполняется для ОБЕХ причин отказа: covered (писатель, дедуп) и `!known` (тихий гвард композиции, WR-02-класс деарма). Если между первым кликом и вторым freshness самаостыла (спуриозный blur, уход фокуса в ходе silence-раунда), клик отклоняется гвардом `!known`, файл остаётся байт-идентичным — и D-01 «клик по отмеченному безопасен» считается доказанным, хотя путь covered-дедупа не исполнялся. Юнит-матрица (`TestBlocklistGuardMatrixSilent`) этот путь пинит, но живое доказательство D-1 в coverage фазы ссылается именно на этот e2e-оракул.
**Fix:** Перед вторым кликом пере-проверить arm/состояние тем же `GetGroupProperties`-полл (метка несёт имя + состояние 1 = covered) и зафейлить шаг, если пункт уже не covered, — тогда байт-идентичность доказывает именно covered-no-op.

## Info

### IN-01: Метка full дублирует потолок литералом

**File:** `internal/indicator/menu.go:94` (`labelBlocklistFull = "Чёрный список заполнен — 64/64"`)
**Issue:** «64/64» захардкожено при том, что фаза целенаправленно свела потолок на `config.MaxAutocorrectBlocklist` (писатель, валидатор, актор — единый источник). `internal/indicator` сознательно не импортирует `config`, зеркальность зафиксирована комментарием — но смена потолка в config молча оставит пункт с ложной меткой (актор при этом покажет full по новому потолку — расхождение состояния и метки).
**Fix:** Строить метку при инициализации пакета из константы (`fmt.Sprintf`, var вместо const) — направление импорта indicator→config ациклично; или добавить в `TestMaxAutocorrectBlocklistExport`/layout-тест grep-пин «64» в `labelBlocklistFull`, чтобы дрейф рвал тест.

### IN-02: Содержательная чистота WARN композиции держится на случайном инварианте, а не на контракте

**File:** `cmd/goswitchd/main.go:470` (`slog.Warn("blocklist write failed", ..., "error", err)`)
**Issue:** WARN прокидывает текст ошибки писателя как есть. Сегодня все досягаемые из композиции отказы контент-свободны (decode/ceiling/not-sequence называют путь/размер), а ошибки валидации, встраивающие паттерн (`autocorrect.apps_blocklist[%d] %q` в `Autocorrect.validate`), недостижимы, потому что `regexp.QuoteMeta` всегда компилируется, а пустая identity отсечена гвардом. Инвариант нигде не закреплён — будущая правка композиции или писателя может тихо начать логировать identity в журнал (граница D-20/D-21). Тест `assertGuardJournal` пинит только текущую матрицу.
**Fix:** На барьере шва опускать `error` до классифицированной причины (errors.Is по сентинелам писателя → короткая формулировка, путь уже в поле `path`), либо добавить пин-тест «никакой WARN клика не содержит QuoteMeta(app)» на произвольной ошибке writer-двойника.

---

_Reviewed: 2026-11-01T00:57:41Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: deep_
