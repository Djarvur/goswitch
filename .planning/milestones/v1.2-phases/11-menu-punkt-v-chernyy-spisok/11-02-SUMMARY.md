---
phase: 11-menu-punkt-v-chernyy-spisok
plan: 02
subsystem: session-indicator
tags: [menu-item, blocklist, freshness-machine, gain-hook, dbusmenu, d04-rename, tdd]

# Dependency graph
requires:
  - phase: 11-menu via 11-01
    provides: post-fix main (Round 2 LifecycleReset переписан — RE-PIN актора исполнен по факту), SPEC-ревизия меню v2 (макет пункта до кода)
  - phase: 11-menu via 11-03
    provides: AppendAutocorrectBlocklist + MaxAutocorrectBlocklist (экспорт 64 — единый источник потолка, читается тестами full-состояния)
provides:
  - Пункт меню id 16 «Добавить <имя> в чёрный список» — живёт от свежей личности: gain → активен с именем (doubleUnderscores), реальный FocusOut → статичная метка disabled (D-03), covered → галочка (D-01), full → «Чёрный список заполнен — 64/64» disabled (S9); covered бьёт full; props AT REQUEST TIME из снапшота (D-02)
  - appid.WithGainHook(func(name string)) — Option-инъекция; хук зовётся ПОСЛЕ снятия o.mu (Pitfall 9); Start принял variadic Option
  - actor.PushFocusedApp + fresh-машина (blApp/blFresh): сброс ТОЛЬКО на реальном FocusOut (synthetic-флип G-2-3 и Reset свежесть не трогают); push-моменты — gain, реальный FocusOut, fold applySnapshot, install-пуш SetMenuSync
  - MenuSync.SetBlocklistApp(app, fresh, covered, full) + indicator.Menu.SetBlocklistApp (dedupe, labelDelta+stateDelta, []removedProps{}, revision заморожен) + Callbacks.BlockCurrentApp + case dispatchClick — полный шов для клик-композиции 11-04
  - actor.BlocklistAppState() (app, known, covered, full) — единственный канал состояния композиции (known требует fresh); actor.EnsureAppid() — безусловный старт наблюдателя (Pitfall-2-лечение, once-per-actor через startAppidOnce)
  - D-04: macroChordWhat « — смена режима» → « — язык»; метка «super+space — язык»; README.md:102 и README.en.md:99 синхронны
affects: [11-04 (клик-композиция QuoteMeta → AppendAutocorrectBlocklist, гварды по BlocklistAppState, проводка EnsureAppid в OnConn), 11-05 (e2e menu-blocklist: id 16 + канал ItemsPropertiesUpdated готовы)]

# Actuals (#2632)
actuals:
  tokens: 21035
  tasks: 3
  commits: 7

# Tech tracking
tech-stack:
  added: [] # ноль новых модулей — go.mod/go.sum нетронуты (tidy-diff зелёный)
  patterns:
    - "freshness-surface: кэш наблюдателя держит «последнюю известную» личность, НОВОЕ поле fresh (gain-хук вверх, реальный FocusOut вниз) — UI читает freshness, не кэш; семантика кэша для гейтов D-53 байт-неизменна"
    - "hook-outside-mutex: колбэк наблюдателя вызывается после Unlock — инверсия actor.mu → appid.mu невозможна (Pitfall 9, -race тест с чтением наблюдателя из хука)"
    - "порядок лока сохраняется односторонним: appid-горутина (хук) → actor.mu → menuSync; push-дисциплина pushMenuSync (синхронно, быстро, дедуп на принимающей стороне)"

key-files:
  created:
    - ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-02-SUMMARY.md"
  modified:
    - "internal/indicator/menu.go (menuIDBlocklistApp=16, rootKidsLocked после ACToggle, blocklistProps-матрица, blocklistRow, фильтр диапазона, SetBlocklistApp, dispatchClick case, macroChordWhat « — язык», labelBlocklistFull)"
    - "internal/indicator/menu_test.go (wantKids id 16, TestMenuBlocklistItemGroupProps, матрица Unknown/Marks 7 строк, TestMenuBlocklistSetterDeltas, TestMenuBlocklistClickDispatch, пин D-04)"
    - "internal/indicator/indicator.go (Callbacks.BlockCurrentApp)"
    - "internal/appid/appid.go (WithGainHook, gainHook-поле, вызов хука в run() вне o.mu, Start variadic opts)"
    - "internal/appid/appid_test.go (TestObserver_GainHook — 3 subtest, Pitfall 9 под -race)"
    - "internal/session/actor.go (PushFocusedApp, blApp/blFresh, HandleLifecycle real-FocusOut ветка, pushBlocklistAppLocked/blocklistStateLocked, autocorrectBlocklistCeil=64, EnsureAppid/BlocklistAppState/startAppidOnce/startAppidDefault, MenuSync+SetBlocklistApp, install-пуш SetMenuSync, fold-пуш applySnapshot)"
    - "internal/session/actor_test.go (fakeMenuSync-рекордер, freshness/fold/accessor/ensure-корпус — 9 тест-функций, blTestApp)"
    - "README.md («<сочетание> — язык» :102)"
    - "README.en.md («<chord> — язык» :99)"

key-decisions:
  - "Freshness = actor-snapshot, не меню и не кэш: blApp/blFresh под a.mu; пустой/несвежий Identity никогда не доходит до меню и композиции (аксессор отдаёт пустое имя при !known — двойной барьер против D-03-мусора)"
  - "EnsureAppid — обёртка над startAppidOnce, а НЕ над ensureAppid: поведение плана («условие ленивого старта не мешает пункту») требует обхода условия при сохранении once-per-actor — буквально-«обёртка» оставила бы Pitfall 2 в силе (поймано GREEN-прогоном)"
  - "Потолок в session — пакетная константа autocorrectBlocklistCeil=64 с комментарием-ссылкой на config.MaxAutocorrectBlocklist (сведение — точечная правка Task 2 плана 11-04); тесты full-состояния читают экспорт config.MaxAutocorrectBlocklist напрямую (единый источник уже сейчас)"
  - "Метка full «Чёрный список заполнен — 64/64» — дискреция A5 в indicator-константе с тем же комментарием-ссылкой; пункт всегда checkmark-элемент (toggleProps жёстко ставит enabled — blocklistRow со свободной формой)"
  - "Install-пуш SetBlocklistApp в SetMenuSync (Rule 2, пуш-момент RESEARCH Pattern 2): меню, перестроенное после рестарта IBus, получает ТЕКУЩЕЕ состояние свежести, а не дефолт-unknown до ближайшего gain"

patterns-established:
  - "freshness-hook-Option: наблюдатель с колбэком после store — форма WithRoleCall (Option-инъекция, вызов вне критсекции)"
  - "state-matrix table-runner: пара Unknown/Marks-тестов с общим раннером (funlen/gochecknoglobals-совместимая форма табличной матрицы)"

requirements-completed: [MENU-01]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "id 16 живёт в layout: rootKidsLocked сразу после menuIDACToggle (дискреция владельца), существующие id 0-15 не сдвинуты, props через расширенный фильтр GetGroupProperties"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/indicator/menu_test.go#TestMenuLayout (wantKids id 16 после ACToggle; RED dd2ca8e → GREEN 8b999e6)"
        status: pass
      - kind: unit
        ref: "internal/indicator/menu_test.go#TestMenuBlocklistItemGroupProps (фильтр диапазона Pitfall 5 + рендер AT REQUEST TIME)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Freshness-машина: gain ставит fresh=true и пушит имя; реальный FocusOut гасит и пушит unknown; synthetic FocusOut (G-2-3) молчит; install-пуш unknown"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_BlocklistItemFreshPushes + #TestActor_BlocklistItemFreshLosses (RED dd2ca8e → GREEN 8b999e6)"
        status: pass
      - kind: unit
        ref: "internal/appid/appid_test.go#TestObserver_GainHook (gain→хук, loss мимо, чтение наблюдателя из хука без дедлока — Pitfall 9 под -race)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Матрица состояний 7 строк: unknown (пустой/протухший) статичная метка disabled без имени/пояснений (D-03), active, covered=галочка (D-01), full «64/64» disabled (S9/A5), covered бьёт full, имя с «_» через doubleUnderscores"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/indicator/menu_test.go#TestMenuBlocklistStateMatrixUnknown + #TestMenuBlocklistStateMatrixMarks (RED 9afc6c9 → GREEN 80d5f92)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Сеттер и клик-шов: dedupe identical (нет эмиссии), эмиссия []removedProps{} + label/state дельты, revision заморожен (Pitfall 10); клик по 16 → Callbacks.BlockCurrentApp, nil — тихо"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/indicator/menu_test.go#TestMenuBlocklistSetterDeltas + #TestMenuBlocklistClickDispatch"
        status: pass
    human_judgment: false
  - id: D5
    description: "D-04 « — язык»: macroChordWhat в коде, пин TestMenuLayout, README.md:102 и README.en.md:99 синхронны"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/indicator/menu_test.go#TestMenuLayout (wantKids «super+space — язык»; RED-ступень 9afc6c9, GREEN 80d5f92)"
        status: pass
      - kind: other
        ref: "grep -q ' — язык' README.md && grep -q ' — язык' README.en.md && grep -q 'macroChordWhat = \" — язык\"' internal/indicator/menu.go — все зелёные (verify Task 2)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Fold-пересчёт: документ с паттерном identity → covered=true без клика; потолок-документ → full; свежесть бьёт кэш на fold-е (real FocusOut → fold пушит unknown даже при паттерне в документе)"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_BlocklistFoldCovered + #TestActor_BlocklistFoldFull + #TestActor_BlocklistFoldFreshnessBeatsCache (RED 41db210 → GREEN 1228b1b)"
        status: pass
    human_judgment: false
  - id: D7
    description: "Аксессор композиции и безусловный старт: unknown / known-требует-fresh / known-fresh / covered / full; EnsureAppid стартует наблюдателя при AutoCorrectEnabled=false + MACR пуст, идемпотентно (Pitfall 2)"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_BlocklistAppState + #TestActor_BlocklistAppFoldedState + #TestActor_EnsureAppidUnconditional"
        status: pass
    human_judgment: false
  - id: D8
    description: "Гейты D-53 байт-неизменны: autoCorrectBoundary/acFocusedApp сверены с до-плановым 34059d8 — BYTE-IDENTICAL; 24 gate-теста зелёны без правок ожиданий; семантика кэша appid для гейтов не менялась (focusGain не тронут)"
    requirement: MENU-01
    verification:
      - kind: other
        ref: "awk-сверка тел функций 34059d8 vs HEAD — BYTE-IDENTICAL обе; go test -run 'TestAutoCorrect|TestAutoConfirm' -race — 24/24 PASS"
        status: pass
    human_judgment: false
  - id: D9
    description: "Зелёная итерация целиком (директива 2): build + vet + golangci-lint 2.15.0 0 issues + test -race ./... (22 пакета, exit 0) + go mod tidy без дрейфа go.mod/go.sum"
    requirement: MENU-01
    verification:
      - kind: other
        ref: "прямые гейты на HEAD 9ab1b1c (literal mise run ci — WINDOWS #23/#24, toolchain seam, прецедент 11-01/11-03)"
        status: pass
    human_judgment: false

# Metrics
duration: 46min
completed: 2026-11-01
status: complete
---

# Phase 11 Plan 02: MENU — пункт «в чёрный список» (ядро: freshness + матрица + D-04) Summary

**Пункт трей-меню id 16 «Добавить <имя> в чёрный список» живёт от свежей личности приложения — gain-хук appid → PushFocusedApp-машина актора → DBusMenu-props; полная 7-строчная матрица состояний D-01/D-03/потолок, fold-пересчёт галочки/полноты без клика, безусловный старт наблюдателя и переименование D-04 « — язык» в коде и README**

## Performance

- **Duration:** 46 min
- **Started:** 2026-10-31T21:25:52Z
- **Completed:** 2026-11-01T00:12Z
- **Tasks:** 3 (полный RED→GREEN на каждую; REFACTOR — lint-сплит корпуса, 9ab1b1c)
- **Files modified:** 9 (кода 7 + README 2)

## Accomplishments

- **Трассер identity-цепочки** (Task 1): `appid.WithGainHook` — хук вызывается ПОСЛЕ снятия o.mu (Pitfall 9; -race тест с чтением наблюдателя прямо из хука); `Start` принял variadic Option; `actor.PushFocusedApp` — fresh-машина под a.mu; в HandleLifecycle ТОЛЬКО не-synthetic ветка FocusOut (без flipCredit, G-2-3; Reset не трогается) ставит fresh=false и пушит; default starter несёт `appid.WithGainHook(a.PushFocusedApp)` — main.go не тронут; в menu.go id 16 в реестре и layout сразу после «Автокоррекции», фильтр GetGroupProperties расширен (Pitfall 5), сеттер по форме SetCorrections
- **Машина состояний пункта** (Task 2): 7-строчная матрица — unknown (пустой кэш ИЛИ fresh=false) → статичная метка disabled БЕЗ имени и БЕЗ пояснений (D-03); fresh → «Добавить <имя> в чёрный список» enabled (doubleUnderscores); covered → галочка enabled (D-01); full → «Чёрный список заполнен — 64/64» disabled (S9/A5); covered бьёт full; props AT REQUEST TIME из снапшота (D-02 — живых a11y-запросов при открытии меню нет); дельты с []removedProps{} литералом, revision заморожен (Pitfall 10); `Callbacks.BlockCurrentApp` + case dispatchClick (nil — тихо); **D-04**: macroChordWhat « — смена режима» → « — язык», метка «super+space — язык», README.md:102/README.en.md:99 синхронны
- **Актор: fold + аксессор + старт** (Task 3): fold applySnapshot после refreshACBlocklist пересчитывает covered (ТА ЖЕ дисциплина матча re.MatchString, что гейт) и full (потолок) и пушит SetBlocklistApp БЕЗ клика — документ меняет галочку/полноту; `BlocklistAppState()` — единственный канал состояния для click-композиции 11-04, known требует fresh (протухшее имя не доходит до композиции); `EnsureAppid()` — безусловный старт (обход ленивого условия, once-per-actor) — Pitfall-2-лечение: выключенная автокоррекция (D-54 дефолт) не оставляет пункт мёртвым
- **Гейты D-53 байт-неизменны**: autoCorrectBoundary/acFocusedApp сверены с до-плановым 34059d8 — BYTE-IDENTICAL; 24 gate-теста зелёны байт-теми же; семантика кэша appid (focusGain, cache-keep) не менялась — fresh-признак надстройка только для меню
- Полный corpus: 3 RED → 3 GREEN (+1 lint-refactor), 9 новых тест-функций в session, 3 в indicator, 1 в appid; всё под -race

## Task Commits

Каждая задача — полный TDD-цикл с атомарными коммитами:

1. **Task 1 (tracer) RED** - `dd2ca8e` (test) — корпус трассера + compile-stub API; RED evidence RED_EVIDENCE_OK (4/4 целевых, 293 pass)
2. **Task 1 GREEN** - `8b999e6` (feat) — WithGainHook, PushFocusedApp/fresh-машина, id 16 в layout+фильтр+сеттер
3. **Task 2 RED** - `9afc6c9` (test) — матрица, клик-шов, пин D-04 (ожидаемая RED-ступень); RED evidence RED_EVIDENCE_OK (3 целевых, 27 pass)
4. **Task 2 GREEN** - `80d5f92` (feat) — полная матрица blocklistProps, dispatchClick case, D-04 в коде+README
5. **Task 3 RED** - `41db210` (test) — fold/accessor/ensure-корпус; RED evidence RED_EVIDENCE_OK (3 целевых)
6. **Task 3 GREEN** - `1228b1b` (feat) — fold-пуш, BlocklistAppState, EnsureAppid/startAppidOnce, wrapcheck/funcorder-чистка
7. **REFACTOR** - `9ab1b1c` (refactor) — сплит матрицы funlen/gocognit (итоговая зелёная итерация)

**Plan metadata:** final docs commit (SUMMARY + STATE + ROADMAP + REQUIREMENTS)

## Tracer Feedback Gate

Task 1 — tracer. Авто-режим (auto_advance=true), verify только automated, human_verify_mode=end-of-phase → verify перезапущен end-to-end на GREEN-коммите: три пакета зелёные под -race. `⚡ Tracer verified end-to-end — expanding` — Задачи 2-3 исполнялись дальше без checkpoint.

## Files Created/Modified

- `internal/appid/appid.go` — gainHook-поле + WithGainHook (форма WithRoleCall); вызов хука в run() после снятия o.mu; Start(ctx, opts ...Option)
- `internal/appid/appid_test.go` — TestObserver_GainHook: gain доставляет имя, loss/немаркированные мимо хука, чтение наблюдателя из хука без дедлока (Pitfall 9)
- `internal/session/actor.go` — blApp/blFresh; PushFocusedApp; HandleLifecycle real-FocusOut ветка (synthetic/Reset мимо); blocklistStateLocked/pushBlocklistAppLocked; autocorrectBlocklistCeil=64; EnsureAppid + startAppidOnce (извлечение из ensureAppid); BlocklistAppState; startAppidDefault (named method, wrapcheck, ireturn-nolint); MenuSync.SetBlocklistApp; install-пуш SetMenuSync; fold-пуш applySnapshot
- `internal/session/actor_test.go` — fakeMenuSync.SetBlocklistApp/menuBlocklist; blTestApp; TestActor_BlocklistItemFreshPushes/Losses; TestActor_BlocklistFoldCovered/Full/FreshnessBeatsCache; TestActor_BlocklistAppState/BlocklistAppFoldedState; TestActor_EnsureAppidUnconditional
- `internal/indicator/menu.go` — menuIDBlocklistApp=16 (после menuIDReload, без перенумерации); labelBlocklistApp/labelBlocklistFull; rootKidsLocked после ACToggle; blocklistProps-матрица + blocklistRow; фильтр диапазона до menuIDBlocklistApp; SetBlocklistApp (dedupe, labelDelta+stateDelta, []removedProps{}); dispatchClick case; macroChordWhat « — язык»
- `internal/indicator/menu_test.go` — wantKids id 16 (checkmark-элемент, пин D-04 «super+space — язык»); TestMenuBlocklistItemGroupProps; TestMenuBlocklistStateMatrixUnknown/Marks (7 строк, runner checkBlocklistRow); TestMenuBlocklistSetterDeltas; TestMenuBlocklistClickDispatch
- `internal/indicator/indicator.go` — Callbacks.BlockCurrentApp
- `README.md` — :102 «сочетание смены режима (`<сочетание> — язык`)»
- `README.en.md` — :99 «mode-switch combination (`<chord> — язык`)»

## Decisions Made

- **EnsureAppid обходит условие, а не зовёт его.** Буквальная «обёртка над ensureAppid» из action-текста оставила ленивое условие в силе — GREEN-прогон теста TestActor_EnsureAppidUnconditional поймал это (start=0 при выключенной автокоррекции). Поведение из behavior-блока плана («условие ленивого старта не мешает пункту») и must_haves (Pitfall 2) делают обход условия обязательным: ensureAppid разложен на условие + startAppidOnce; публичная форма зовёт startAppidOnce с проверкой флага
- **Аксессор отдаёт пустое имя при !known** — двойной барьер D-03: композиция 11-04 не увидит протухшее имя, даже проигнорировав флаг known (запрет «sender-имена никогда не становятся паттерном»)
- **Пункт всегда checkmark-элемент** (toggle-type=checkmark во всех состояниях) — галочка D-01 появляется сменой toggle-state, без LayoutUpdated
- **Install-пуш в SetMenuSync** (Rule 2, пуш-момент из PATTERNS раздел 2): меню после рестарта IBus получает текущее состояние пункта
- **Тесты потолка читают config.MaxAutocorrectBlocklist** (экспорт 11-03) — единый источник уже в корпусе; session-константа autocorrectBlocklistCeil=64 и indicator-метка «64/64» несут комментарий-ссылку для точечного сведения в 11-04

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] mise run ci в сессии исполнителя падает на toolchain-разъёме (3-е проявление, прецедент 11-01/11-03)**
- **Found during:** Task 3 (verify-шаг)
- **Issue:** env-задачи mise резолвит системный go 1.27.3 при GOROOT mise-инсталла 1.23.12: «compile: version go1.23.12 does not match go tool version go1.27.3»; `mise exec -- go build/test` работает (прямой вызов перехватывается), shell-задачи `mise run` — нет
- **Fix:** те же гейты исполнены напрямую: go build ./..., go vet ./..., golangci-lint 2.15.0 (бинарник mise-инсталла, GOROOT сброшен) — 0 issues, go test -race -count=1 ./... (22 пакета, exit 0), go mod tidy + git diff --exit-code; записано в WINDOWS.md (#24, kind unrun-verify)
- **Files modified:** нет (методика, не код)
- **Verification:** все гейты зелёные на HEAD 9ab1b1c; D-53 гейты байт-неизменны
- **Committed in:** — (методика, не код)

**2. [Rule 2 - Missing critical] Install-пуш SetBlocklistApp в SetMenuSync**
- **Found during:** Task 1 (GREEN)
- **Issue:** без install-пуша меню, пересобранное после рестарта IBus (OnConn заново), показывало unknown до ближайшего gain — пункт «мертвел» на живое состояние после реконнекта
- **Fix:** SetMenuSync установочный блок пушит текущее состояние (форма SetCorrections install-push; пуш-момент RESEARCH Pattern 2 «install-пуш при attach»)
- **Files modified:** internal/session/actor.go
- **Verification:** TestActor_BlocklistItemFreshPushes/install — install-пуш ровно один, zero-state
- **Committed in:** 8b999e6

**3. [Rule 1 - Bug] EnsureAppid как буквально-«обёртка» сохранял Pitfall 2**
- **Found during:** Task 3 (GREEN-прогон)
- **Issue:** a.EnsureAppid(){a.ensureAppid()} — ленивое условие (MACR пуст И автокоррекция выкл) отказывало старт; тест behavior-строки «условие ленивого старта не мешает пункту» красный
- **Fix:** ensureAppid разложен на ensureAppid (условие) + startAppidOnce (тело); EnsureAppid проверяет флаг и зовёт startAppidOnce — безусловный старт с сохранением once-per-actor и failed-start WARN-деградации
- **Files modified:** internal/session/actor.go
- **Verification:** TestActor_EnsureAppidUnconditional зелёный (start=1, идемпотентно)
- **Committed in:** 1228b1b

**4. [Rule 3 - Blocking] Замыкание в собственном инициализаторе не компилируется**
- **Found during:** Task 1 (GREEN-сборка)
- **Issue:** `a := &Actor{startAppid: func() ... a.PushFocusedApp ...}` — область видимости `a` начинается ПОСЛЕ объявления; «undefined: a»
- **Fix:** starter вынесен из композитного литерала; в Task 3 GREEN поднят до именованного метода startAppidDefault (метод-value после конструкции) — попутно закрыл contextcheck-флаг на NewActor$1
- **Files modified:** internal/session/actor.go
- **Verification:** go build зелёный; contextcheck в lint-корпусе 0
- **Committed in:** 8b999e6, 1228b1b

---

**Total deviations:** 4 auto-fixed (1 bug, 1 missing-critical, 2 blocking-окружение/компиляция). **Impact on plan:** все отклонения — обслуживание зелёной итерации (директива 2) и полноты пуш-моментов; объём плана не менялся, go.mod/go.sum нетронуты, main.go не тронут.

## Issues Encountered

- Литеральный `mise run ci` остаётся недоступен в сессии исполнителя (WINDOWS #23 открыт с 11-03; текущий ряд записан #24) — прямые гейты эквивалентны и зелёные; починка сессийного PATH/GOROOT разъёма — отдельная задача вне фазы
- Линтер (строгий v2-конфиг, default: all) поднял 20 замечаний на новый код — все закрыты в GREEN/REFACTOR коммитах (funcorder-перестановки, funlen/cyclop-сплиты, goconst, lll, nlreturn, wrapcheck-обёртка, ireturn-nolint с прецедентом emitter()); контекстных исключений из конфига не потребовалось
- До-существующий untracked-хвост репозитория (.gsd/, .zcode/, draft.txt, test/e2e/e2e и пр.) не создавался этим планом и не трогался

## TDD Gate Compliance

| Task | RED | GREEN | REFACTOR | Status |
|------|-----|-------|----------|--------|
| 1 (tracer) | dd2ca8e (test) | 8b999e6 (feat) | — | Pass |
| 2 | 9afc6c9 (test) | 80d5f92 (feat) | — | Pass |
| 3 | 41db210 (test) | 1228b1b (feat) | 9ab1b1c (lint-сплит корпуса indicator) | Pass |

- RED evidence всех трёх задач проверено машиной: `gsd_run check tdd-red-evidence` → RED_EVIDENCE_OK (target_test_failed); записи-прецеденты TAP-проекции `go test -v` 1:1 сохранены в фаз-директории (evidence-20261031-11-02-red-task{1,2,3}.json; конвертер — прецедент 10-10)
- Compile-stub API в RED-коммитах — необходимость Go (11-03 прецедент): WithGainHook хранит поле без вызова, PushFocusedApp/EnsureAppid/BlocklistAppState/SetBlocklistApp — no-op/zero; RED-провалы честные, на assertion целевого поведения
- Порядок gate-коммитов проверен: test(11-02) предшествует feat(11-02) для всех трёх задач (git log выше)
- Пины, зелёные до GREEN (TestMenuBlocklistSetterDeltas, часть subtest-ов на zero-стабах) — пины корпуса; RED-гейт каждой задачи закрывается целевым тестом (TestActor_BlocklistItemFreshMachine; TestMenuBlocklistStateMatrix; TestActor_BlocklistFoldRecompute)

## Flagged Assumptions

План несёт surfaced flagged assumption (spec-less probe fallback, никогда не авто-разрешается): `{"requirement_id":"MENU-01","category":"unclassified","status":"unresolved"}` — край «неизвестная личность» выражен матрицей состояний и fresh-машиной (задачи 1-2), формальная классификация края остаётся unresolved до фазовой верификации. Verifier обязан учитывать строку.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- 11-04 получает: полный клик-шов (Callbacks.BlockCurrentApp, case dispatchClick), аксессор BlocklistAppState (единственный канал состояния), EnsureAppid для проводки в OnConn, сеттер меню и все push-моменты; писатель AppendAutocorrectBlocklist + QuoteMeta — композиция из RESEARCH Pattern 4; сведение потолка — точечная правка (autocorrectBlocklistCeil, метка «64/64» → config.MaxAutocorrectBlocklist)
- 11-05 (e2e): id 16 в layout, канал ItemsPropertiesUpdated и машина fresh/covered/full спиннуты юнит-корпусом
- Открытые хвосты фазы: Round 2 journal-verify (11-01, ряд на следующей сессии владельца); WINDOWS #23/#24 (mise run ci, средовой)
- MENU-01 остаётся shared-ID гейтом (11-04/11-05 объявляют его же) — requirements.ready-ids не даст закрыть требование до их SUMMARY

## Self-Check: PASSED

- Файлы: internal/appid/appid.go, internal/appid/appid_test.go, internal/session/actor.go, internal/session/actor_test.go, internal/indicator/menu.go, internal/indicator/menu_test.go, internal/indicator/indicator.go, README.md, README.en.md — FOUND на диске
- Коммиты: dd2ca8e, 8b999e6, 9afc6c9, 80d5f92, 41db210, 1228b1b, 9ab1b1c — все FOUND в git log (TDD-порядок test-before-feat на каждую задачу)
- commits измерены от ledger-базы 34059d84: git rev-list --count = 7
- Acceptance criteria перепроверены на HEAD 9ab1b1c: три пакета зелёные под -race; wantKids id 16 после ACToggle; фильтр диапазона включает 16; матрица 7 строк спиннута; дедуп/эмиссия/revision пины; dispatchClick + nil; doubleUnderscores; D-04 в коде и обоих README (grep зелёный); fold/accessor/ensure корпус; гейты D-53 BYTE-IDENTICAL vs 34059d8, 24 gate-теста PASS; go.mod/go.sum diff --exit-code чист
- Блокеров для следующего плана нет

---
*Phase: 11-menu-punkt-v-chernyy-spisok*
*Completed: 2026-11-01*
