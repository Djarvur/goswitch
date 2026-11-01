---
phase: 11-menu-punkt-v-chernyy-spisok
plan: 05
subsystem: e2e-menu-blocklist
tags: [blocklist, dbusmenu-event, live-stand, vm-venue, quotemeta, freshness, appid, repin, tdd]

# Dependency graph
requires:
  - phase: 11-menu via 11-04
    provides: живую клик-композицию id 16 (blocklistToggle.add — гварды, QuoteMeta, sync reload), Callbacks.BlockCurrentApp, SetBlocklistApp, BlocklistAppState, EnsureAppid
  - phase: 11-menu via 11-03
    provides: AppendAutocorrectBlocklist — persist на готовом паттерне (одна строка свопа, строгий pre-write Load)
  - phase: 07-menu via 07-06
    provides: Event-канал (case_menu.go пины), blocklist-silent oracle-машину (case_autocorrect.go), menu-v2 кейс для re-pin
provides:
  - e2e-кейс menu-blocklist: Event-клик id 16 по живому демону — наблюдение fixture-identity (appid-обозреватель + machine-activation нуджи 10-08), arm-гейт по GetGroupProperties, оракулы файла (ровно одна строка свопа на QuoteMeta-литерал), строгого config.Load, тишины автокоррекции (машина blocklist-silent дословно) и D-01 no-op повторного клика; teardown — temp-config state-policy case_menu.go
  - регистрация в реестре/usage/pickCase test/e2e/main.go; mise-задача e2e-menu-blocklist (live, NOT in ci)
  - unit-пины хоста: TestParseBlocklistDoc + TestAssertOnlyBlocklistLine (форма TestParseMenuToggles; байтовый след свопа проверен против 11-03 энкодера)
  - дефект-фиксы живого прогона: bridgeApp отказывает namespace реестра at-spi2 (мусорная identity «org» на relay-путях, c980d2e); деарм freshness перенесён в слой identity — blur отслеживаемой поверхности через hook(""), engine-шум вне freshness (a943656)
  - menu-v2 VM-хардненинг: noSurfacePreflight (меню-семейство) + activateGoswitchRetry в spawnMenuDaemon (гонка одношотовой активации, находка 2026-10-12)
affects: [фазовая верификация MENU-01 (живой цикл доказан на ВМ), UAT end-of-phase]

# Actuals (#2632)
actuals:
  tokens: 15700 # chars/4 по фактическому диффу плана (62508 chars, 889+/53-)
  tasks: 2
  commits: 10 # MEASURED: git rev-list --count 03a73fe..HEAD (1 feat + 8 fix + 1 docs)

# Tech tracking
tech-stack:
  added: [] # ноль новых модулей — go.mod/go.sum нетронуты (tidy-diff зелёный, T-11-SC accept)
  patterns:
    - "menu-case-vm-form: observer-round и fixture-witness на machine-activation нуджах (waitFocusInFrom/блоклистFixtureFocus с ydotool center-click), сигнал продолжения — daemon focus_in / AT-SPI witness — честность гейтов не делегирована клику"
    - "arm-gate-by-label: арминг пункта читается из живого GetGroupProperties (тот же снапшот, что читает гвард клика, S1); литерал props непустой (gdbus не типизирует пустой []); матч по mnemonic-удвоенному имени (Pitfall 4)"

key-files:
  created:
    - "test/e2e/case_menu_blocklist.go"
    - "test/e2e/case_menu_blocklist_test.go"
    - ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-05-SUMMARY.md"
  modified:
    - "test/e2e/main.go (регистрация menu-blocklist; noSurfacePreflight menu-v2/menu-blocklist)"
    - "test/e2e/case_menu.go (spawnMenuDaemon → activateGoswitchRetry)"
    - "mise.toml (задача e2e-menu-blocklist, live/NOT in ci)"
    - "internal/appid/appid.go (registry-namespace guard в bridgeApp; focusLoss + loss-ветка run-цикла)"
    - "internal/appid/appid_test.go (+TestBridgeApp_RefusesRegistryNamespace, +TestObserver_LossHook)"
    - "internal/appid/export_test.go (+BridgeApp alias)"
    - "internal/session/actor.go (PushFocusedApp(\"\") — деарм; HandleLifecycle без blFresh-сброса)"
    - "internal/session/actor_test.go (freshness-пины на blur-семантику)"
    - "docs/adr/adr-008-spike-log.md (датированные ряды + механика menu-семейства на ВМ)"

key-decisions:
  - "Деарм freshness перенесён из engine-слоя в слой identity: клик, дающий окну фокус, парой несёт РЕАЛЬНЫЙ focus_out вслед за a11y-gain (порядок воспроизведён в diag16) — сброс на реальном FocusOut (11-02) структурно проигрывал гонку; деарм теперь = blur НАБЛЮДАЕМОЙ поверхности (focused=0 с bridge-атрибуцией == кэш), relay-события реестра никогда не деармят, engine вне бизнеса freshness. Семантика D-02/D-03 для владельца сохранена (деарм = «ушёл с поверхности», в т.ч. на ненаблюдаемое приложение); известный хвост — SIGKILL поверхности без blur оставляет арм до следующего gain (самозаживление на следующем взаимодействии)"
  - "bridgeApp отказывает namespace самого реестра (/org/a11y/): relay реестра ретранслирует события под своей адресацией /org/a11y/atspi/accessible/<id>, naive-скан «/a11y/» усекал префикс до identity «org» — пункт армился мусором, клик записал бы QuoteMeta(«org»); guard = дисциплина немаркированных (cache-keep, never guess)"
  - "Живая лестница отказов заездов 1-7 записана честно в spike-log (прецедент Фазы 10): каждый FAIL — setup/oracle-отказ с причиной, фиксы — отдельные коммиты; PASS-ряды несут HEAD сборки и машинные оракулы"
  - "Стендовая мутация venue-mechanics-прецедентом: gnome-shell-extension-appindicator 58 из архива + virsh reboot — без StatusNotifierWatcher экспорт /Menu не происходит (indicator.registerItem: probe → export → register) и Event-канал меню мёртв на корню"
  - "menu-v2 re-pin тем же прогоном стенда: существующие id 1/2/4/5 не сдвинуты (Pitfall 5 «ломается громко by design» — не сработал, сдвига нет); радио-клики прошли зелёными даже в FAIL-заезде"

patterns-established:
  - "blocklistFixtureFocus: witness-гейт фикстуры с nudge-лестницей (AT-SPI witness — единственный сигнал продолжения, нудж — провокация)"
  - "lossRecorder: synthetic-feed форма тестов loss-ветки наблюдателя (own/foreign/relayed blur)"

requirements-completed: [MENU-01]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Живой цикл MENU-01 на ВМ: Event-клик id 16 → документ вырос ровно на одну строку (QuoteMeta-литерал наблюдаемой identity) → строгий config.Load зелёный → тишина автокоррекции в fixture (witness 0→7, поле дословно, fired=0, одна app-blocked) → повторный клик байт-идентичен (D-01)"
    requirement: MENU-01
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md секция Фаза 11: menu-blocklist 2026-11-01 00:27 UTC (заезд 8, HEAD 489aa5f) — PASS, CASE_EXIT=0"
      - kind: unit
        ref: "test/e2e/case_menu_blocklist_test.go#TestParseBlocklistDoc + #TestAssertOnlyBlocklistLine (байтовый след против 11-03 энкодера)"
    status: pass
    human_judgment: false
  - id: D2
    description: "Re-pin menu-v2 тем же прогоном стенда: id 1/2/4/5 не сдвинуты (радио-записи, оба тумблера persist+apply синхронно, рестарт-survival, обратные клики)"
    requirement: MENU-01
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md секция Фаза 11: menu-v2 re-pin 2026-11-01 00:30 UTC (заезд 2, HEAD cf9c857) — PASS, CASE_EXIT=0"
    status: pass
    human_judgment: false
  - id: D3
    description: "Кейс зарегистрирован (реестр/usage/pickCase), mise-задача e2e-menu-blocklist на месте (live, NOT in ci)"
    requirement: MENU-01
    verification:
      - kind: other
        ref: "grep -q 'menu-blocklist' test/e2e/main.go — OK; grep -q 'e2e-menu-blocklist' mise.toml — OK"
    status: pass
    human_judgment: false
  - id: D4
    description: "Хостовые гейты зелёной итерации (директива 2): go build/vet ./... чисто, golangci-lint 2.15.0 0 issues (весь репозиторий), go test -race -count=1 ./... exit 0, go mod tidy + git diff --exit-code -- go.mod go.sum чисто; literal 'mise run ci' — известный toolchain-разъём сессии (WINDOWS #26), гейты исполнены напрямую"
    requirement: MENU-01
    verification:
      - kind: other
        ref: "прямые гейты на HEAD bc5a506 (прецедент 11-01..11-04, WINDOWS #23-#25)"
    status: pass
    human_judgment: false
  - id: D5
    description: "bridgeApp отказывает namespace реестра at-spi2: unit-пин на живых формах захвата ВМ (relay /org/a11y/atspi/accessible/2720 → \"\", прямые legacy-формы → полная identity); corpus appid/session/cmd зелёный под -race"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/appid/appid_test.go#TestBridgeApp_RefusesRegistryNamespace"
    status: pass
    human_judgment: false
  - id: D6
    description: "Деарм freshness из слоя identity: observer loss-ветка (own blur → hook(\"\"), foreign/relayed blur → тишина) + actor деарм-пуш; freshness-пины переписаны на blur-семантику; corpus session/appid зелёный под -race; живое подтверждение — заезд 8 (арм после нуджа, деарм-гонка исчезла)"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/appid/appid_test.go#TestObserver_LossHook + internal/session/actor_test.go#TestActor_BlocklistItemFreshLosses"
    status: pass
    human_judgment: false
  - id: D7
    description: "By-ear приёмка владельца пункта меню в реальном трее (рендер, клик мышью, деарм при уходе в ненаблюдаемое приложение) — машинные оракулы кейса её не заменяют"
    requirement: MENU-01
    verification: []
    human_judgment: true
    rationale: "VALIDATION.md Manual-Only + UAT end-of-phase (конфиг human_verify_mode: end-of-phase); живой цикл доказан машиной на ВМ (D1), визуальная/эргономическая приёмка — за владельцем"

# Metrics
duration: 85min (заезды стенда с диагностикой дефектов живой цепочки: 23:07-00:31 UTC)
completed: 2026-11-01
status: complete
---

# Phase 11 Plan 05: MENU — живой e2e menu-blocklist + re-pin menu-v2 Summary

**Живое доказательство MENU-01 на ВМ-стенде: один Event-клик id 16 по живому демону пишет QuoteMeta(identity) в apps_blocklist (байтовый след — ровно одна строка свопа), строгий Load зелёный, автокоррекция замолкает в fixture (fired=0, одна app-blocked, поле дословно), повторный клик по отмеченному байт-идентичен (D-01); menu-v2 re-pin тем же прогоном — id 1/2/4/5 живы. Попутно пойманы и закрыты два дефекта живой цепочки (мусорная identity «org» на relay-путях реестра; деарм freshness, проигрывающий гонку engine-шуму) и venue-блокер (экспорт /Menu условен StatusNotifierWatcher).**

## Performance

- **Duration:** 85 min (23:07–00:31 UTC; 8 заездов кейса с живой диагностикой)
- **Started:** 2026-10-31T23:07Z
- **Completed:** 2026-11-01T00:31 UTC / 2026-11-01 03:31 +0300
- **Tasks:** 2 (кейс+mise-задача; живой прогон+ряды)
- **Commits:** 10 (MEASURED от ledger 03a73fe: 1 feat + 8 fix + 1 docs)
- **Files:** 11 (3 создано, 8 модифицировано)

## Accomplishments

- **Кейс menu-blocklist** (Task 1, 6afd1e4): Event-драйвер клика id 16 по живому демону (канал case_menu.go), наблюдение fixture-identity кейс-обозревателем appid, arm-гейт по живому GetGroupProperties (тот же снапшот, что читает гвард клика), оракулы: файл вырос ровно на одну строку (QuoteMeta-литерал; парсинг partial-декодером + байтовый своп против 11-03 энкодера), строгий config.Load, статус autocorrect_enabled=true нетронут; тишина — машина blocklist-silent 07-06 дословно (абстейнция app-blocked, fired=0, witness-счёт, дословный readback, PLAIN stdout oracle); D-01 no-op повторного клика (байт-идентичность); teardown — temp-config state-policy case_menu.go. Unit-пины хоста: TestParseBlocklistDoc, TestAssertOnlyBlocklistLine. Регистрация в main.go (реестр/usage/pickCase) + mise-задача e2e-menu-blocklist (live, NOT in ci)
- **Живой прогон** (Task 2): menu-blocklist PASS (00:27 UTC, HEAD 489aa5f) — весь цикл зелёный на живом столе ВМ; menu-v2 re-pin PASS (00:30 UTC, HEAD cf9c857) — существующие пины id 1/2/4/5 не сдвинуты; датированные ряды обоих кейсов + честная лестница отказов заездов 1-7 + пять находок механики — в adr-008-spike-log.md (bc5a506)
- **Два дефекта живой цепочки закрыты** (Rule 1, в коде фазы 11):
  - **bridgeApp vs namespace реестра** (c980d2e): relay реестра at-spi2 ретранслирует события под /org/a11y/atspi/accessible/<id>; naive-скан «/a11y/» давал identity «org» — пункт армился мусором, клик записал бы QuoteMeta(«org») в конфиг. Фикс: namespace реестра refused, прямые legacy-формы извлекаются как прежде; unit-пин на живых формах захвата
  - **Деарм freshness** (a943656): клик, дающий окну фокус, парой несёт РЕАЛЬНЫЙ focus_out вслед за a11y-gain (порядок воспроизведён журналом) — сброс blFresh на реальном FocusOut (11-02) структурно проигрывал гонку; деарм перенесён в слой identity: blur НАБЛЮДАЕМОЙ поверхности (focused=0 с bridge-атрибуцией == кэш) через hook(""), relay-события не деармят, engine вне freshness. Корпус: TestObserver_LossHook (own/foreign/relayed), freshness-пины на blur-семантику; D-02/D-03-семантика для владельца сохранена
- **Хостовые гейты**: build/vet чисто, golangci-lint 2.15.0 0 issues, go test -race ./... exit 0, tidy-diff чист (go.mod/go.sum нетронуты — T-11-SC accept)

## Task Commits

1. **Task 1: кейс+пины+регистрация+mise** - `6afd1e4` (feat) — case_menu_blocklist.go + unit-пины + реестр + e2e-menu-blocklist
2. **Fix: noSurfacePreflight menu-семейства** - `b67d1cd` — VM-стенд без chromium (заезд 1; философия 10-05/10-06)
3. **Fix: observer-round machine-activation** - `19fc9a8` — вместо 30-минутного host-бюджета ручного клика (заезд 2)
4. **Fix: witness nudge-лестница** - `5fa8781` — map-focus не раздаётся свежим окнам (заезд 3; находка 10-08)
5. **Fix: непустой props-литерал** - `e78789c` — gdbus не типизирует пустой [] (заезд 4)
6. **Fix: bridgeApp vs реестр** - `c980d2e` — мусорная identity «org» (заезд 5; Rule 1, продукт)
7. **Fix: деарм freshness из identity** - `a943656` — engine-шум вне freshness (заезд 6; Rule 1, продукт)
8. **Fix: arm-гейт mnemonic-удвоение** - `489aa5f` — gs__e2e__password в метке (заезд 7)
9. **Task 2: ряды spike-log** - `bc5a506` (docs) — датированные PASS-ряды + лестница отказов + механика
10. **Fix: menu-v2 activateGoswitchRetry** - `cf9c857` — встроен между 8 и 9 по ходу прогона (re-pin заезд 1 FAIL → заезд 2 PASS)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] VM-стенд без chromium — default preflight валит setup**
- **Found during:** Task 2 (заезд 1)
- **Issue:** preflight chromium-launch требует google-chrome/gte/gedit — поверхности, которых menu-кейсы не используют; на ВМ их нет
- **Fix:** noSurfacePreflight для menu-v2/menu-blocklist (b67d1cd) — дословная философия 10-05/10-06 (autocorrect-terminal-silent, term-семейство)
- **Files modified:** test/e2e/main.go
- **Committed in:** b67d1cd

**2. [Rule 3 - Blocking] Observer-round кейса нёс 30-минутный host-бюджет ручного клика**
- **Found during:** Task 2 (заезд 2 — watchdog 3m0s истёк в waitZenityEntry)
- **Issue:** acStartObserverRound ждёт человеческий клик (manual-activation протокол 2026-10-09/10) — на безлюдной ВМ бюджет не расходуется
- **Fix:** blocklistObserverRound — zenity + waitFocusInFrom с ydotool-нуджем (10-08 машина); сигнал продолжения прежний — daemon focus_in (19fc9a8); acStartObserverRound не тронут (host-кейсы 07-06)
- **Files modified:** test/e2e/case_menu_blocklist.go
- **Committed in:** 19fc9a8

**3. [Rule 3 - Blocking] AT-SPI-witness фикстуры не садился — сессия ВМ не раздаёт map-focus**
- **Found during:** Task 2 (заезд 3 — witness "(none)" 15 s)
- **Issue:** fixture в a11y-дереве есть (gs-e2e-password, pid процесса), но фокус свежему окну не выдаётся (находка 10-08 для gnome-terminal, теперь и для GTK4-фикстуры); grab-полки host-эры не грантуют
- **Fix:** blocklistFixtureFocus — nudge-лестница ydotool center-click, witness остаётся единственным сигналом продолжения (5fa8781)
- **Files modified:** test/e2e/case_menu_blocklist.go
- **Committed in:** 5fa8781

**4. [Rule 1 - Bug] bridgeApp усекал relay-пути реестра до мусорной identity «org»**
- **Found during:** Task 2 (заезд 5 — пункт армился «org»; клик записал бы QuoteMeta(«org») — ровно те «мусорные паттерны», которые D-03 запрещает)
- **Issue:** реестр at-spi2 ретранслирует события под /org/a11y/atspi/accessible/<id>; naive-скан "/a11y/" усекал префикс; relay-события гоняют прямые legacy-эмиссии — последний выигранный кэш был мусорным
- **Fix:** guard registryPathPrefix в bridgeApp (c980d2e) + TestBridgeApp_RefusesRegistryNamespace на живых формах захвата
- **Files modified:** internal/appid/appid.go, appid_test.go, export_test.go
- **Committed in:** c980d2e

**5. [Rule 1 - Bug] Деарм freshness на реальном FocusOut проигрывал гонку engine-шуму**
- **Found during:** Task 2 (заезд 6 — пункт не армился вовсе; журнал diag16: gain → push → focus_out)
- **Issue:** клик, грантующий фокус, парой несёт РЕАЛЬНЫЙ focus_out вслед за a11y-gain — сброс 11-02 стирал свежесть живой поверхности; live-цикл MENU-01 не сходился
- **Fix:** деарм из слоя identity — blur отслеживаемой поверхности (focused=0 с bridge-атрибуцией == кэш) через hook(""); PushFocusedApp("") деарм-пуш; HandleLifecycle без blFresh-сброса (a943656); корпус переписан (TestObserver_LossHook, freshness-пины). Семантика D-02/D-03 владельца сохранена; хвост: SIGKILL без blur оставляет арм до следующего gain (самозаживление)
- **Files modified:** internal/appid/appid.go, internal/session/actor.go, оба тест-файла
- **Committed in:** a943656

**6. [Rule 3 - Blocking] gdbus не типизирует пустой [] в GetGroupProperties**
- **Found during:** Task 2 (заезд 4 — InvalidArgs)
- **Fix:** непустой props-литерал ["label"] — демон игнорирует фильтр, гейт по label валиден (e78789c)
- **Files modified:** test/e2e/case_menu_blocklist.go
- **Committed in:** e78789c

**7. [Rule 1 - Bug] Arm-гейт искал сырое имя в метке с mnemonic-удвоением**
- **Found during:** Task 2 (заезд 7 — label нёс gs__e2e__password при вооружённом пункте; цепочка уже зелёная)
- **Fix:** гейт матчит strings.ReplaceAll(app, "_", "__") — дисциплина doubleUnderscores renderer'а (489aa5f)
- **Files modified:** test/e2e/case_menu_blocklist.go
- **Committed in:** 489aa5f

**8. [Rule 3 - Blocking] menu-v2 spawn ladder — одношотовая активация против свежего демона**
- **Found during:** Task 2 (re-pin заезд 1 — «Cannot find engine goswitch-en»; радио-пины 1/2 при этом зелёные)
- **Fix:** spawnMenuDaemon → activateGoswitchRetry — форма term-negative-witnesses 2026-10-12, которую несут все VM-лестницы (cf9c857)
- **Files modified:** test/e2e/case_menu.go
- **Committed in:** cf9c857

**9. [Rule 3 - Blocking] literal `mise run ci` — toolchain-разъём сессии (5-е проявление, WINDOWS #23/#24/#25)**
- **Found during:** Task 1 verify
- **Issue:** GOROOT mise-инсталла 1.23.12 vs системный go 1.27.3 («compile: version does not match»)
- **Fix:** те же гейты напрямую: go build/vet, golangci-lint 2.15.0 (0 issues), go test -race ./... (exit 0), go mod tidy + git diff --exit-code (чисто); записано в WINDOWS.md (#26, kind unrun-verify)
- **Files modified:** нет (методика)
- **Committed in:** — (WINDOWS.md — .planning, вне код-коммитов)

---

**Total deviations:** 9 auto-fixed (2 Rule 1 — дефекты продукта, пойманные живым прогоном; 6 Rule 3 — средовые/форменные блокеры заездов; 1 методика гейтов) + 1 стендовая мутация venue-mechanics-прецедентом (gnome-shell-extension-appindicator из архива + virsh reboot — вне репо). **Impact on plan:** объём плана не менялся (оба таска исполнены дословно); живой прогон потребовал починки двух дефектов фазового кода 11-02/11-04 — именно ради этого живой прогон и существовал; go.mod/go.sum нетронуты.

## Issues Encountered

- **Экспорт /Menu условен StatusNotifierWatcher** (indicator.registerItem: probe → export → register; отказ = errNoWatcher до всякого экспорта): на stock GNOME-сессии ВМ без трей-стека Event-канал меню мёртв на корню. Доставлен gnome-shell-extension-appindicator 58-1ubuntu24.04.1 + reboot ВМ (venue-mechanics-прецедент: zsh, wl-clipboard, GTK_MODULES). Для будущих menu-кейсов на ВМ — трей-стек обязателен
- literal `mise run ci` остаётся недоступен в сессии исполнителя (WINDOWS #23-#26) — прямые гейты эквивалентны и зелёные; починка сессийного PATH/GOROOT разъёма — отдельная задача вне фазы
- До-существующий untracked-хвост репозитория (.gsd/, .zcode/, draft.txt, goswitchd, test/e2e/e2e и пр.) не создавался этим планом и не трогался
- config-флак TestWatch_BrokenAutocorrectBlockKeepsLastGood под -race параллельной нагрузкой — задокументированный класс (WINDOWS 07/08); соло и повторный полный прогон зелёные

## Known Stubs

None - кейс полностью проводит живой цикл; заглушек нет (stub-скан чист).

## TDD Gate Compliance

План не содержит tdd="true" задач (TDD_MODE гейт не активен для этих тасков — test-only артефакты + live-прогон). Директива 1 применена в доступной форме: unit-пины хоста (TestParseBlocklistDoc, TestAssertOnlyBlocklistLine, TestBridgeApp_RefusesRegistryNamespace, TestObserver_LossHook) написаны ДО/вместе с поведением и гоняются в каждом прогоне; e2e-кейс сам является тестом — RED-форма его оракулов доказана честной лестницей отказов заездов 1-7 (каждый FAIL — на planned-behavior assertion, ни один оракул не полировался под отказ).

## Flagged Assumptions

План не добавлял новых flagged assumptions. Открытая фазовая строка из 11-02 (`{"requirement_id":"MENU-01","category":"unclassified","status":"unresolved"}` — край «неизвестная личность») остаётся unresolved до фазовой верификации; Verifier обязан учитывать. Живой прогон дал дополнительное свидетельство: после c980d2e/a943656 демона никогда не наблюдается арм мусорной identity на ВМ.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Фаза 11 (5 из 5 планов) исполнена: живой цикл MENU-01 доказан на ВМ, re-pin menu-v2 зелёный; by-ear приёмка владельца — UAT end-of-phase (VALIDATION.md Manual-Only, D7 coverage)
- requirements.ready-ids закроет MENU-01 (объявлен 11-04 и 11-05; обе SUMMARY существуют)
- Хвосты: WINDOWS #23-#26 (mise run ci toolchain-разъём, средовой); Orca Round 2 (11-01, journal-verify); стендовая зависимость menu-кейсов от трей-стека (appindicator) — задокументирована в spike-log
- Блокеров для фазовой верификации нет

## Self-Check: PASSED

- Файлы: test/e2e/case_menu_blocklist.go, case_menu_blocklist_test.go, main.go, case_menu.go, mise.toml, internal/appid/{appid.go,appid_test.go,export_test.go}, internal/session/{actor.go,actor_test.go}, docs/adr/adr-008-spike-log.md, 11-05-SUMMARY.md — FOUND на диске
- Коммиты: 6afd1e4, b67d1cd, 19fc9a8, 5fa8781, e78789c, c980d2e, a943656, 489aa5f, cf9c857, bc5a506 — все FOUND в git log (10 от ledger 03a73fe, rev-list count = 10)
- Acceptance criteria перепроверены: пин id 16 + Event-канал; оракулы (одна строка свопа = QuoteMeta(identity), строгий Load, silence, no-op D-01); teardown state-policy; menu-v2 re-pin зелёный (id 1/2/4/5); датированные ряды с HEAD в spike-log; tidy-diff чист; хостовые гейты зелёные
- Живые ряды: menu-blocklist PASS 00:27 UTC, menu-v2 PASS 00:30 UTC (одна сессия стенда)
- Блокеров для фазовой верификации нет

---
*Phase: 11-menu-punkt-v-chernyy-spisok*
*Completed: 2026-11-01*
