---
phase: 11-menu-punkt-v-chernyy-spisok
plan: 04
subsystem: daemon-menu
tags: [blocklist, click-composition, quotemeta, guard-matrix, fail-closed, onconn-wiring, ceiling-reconciliation, tdd]

# Dependency graph
requires:
  - phase: 11-menu via 11-02
    provides: пункт id 16 + Callbacks.BlockCurrentApp + case dispatchClick, actor.BlocklistAppState (единственный канал состояния), actor.EnsureAppid, menu.SetBlocklistApp (сеттер с dedupe), fresh-машина blApp/blFresh
  - phase: 11-menu via 11-03
    provides: AppendAutocorrectBlocklist(path, pattern) — persist на готовом паттерне; MaxAutocorrectBlocklist (экспорт 64) — цель сведения
provides:
  - blocklistToggle.add — клик-композиция id 16 (форма configToggle.flip): пустой путь → WARN-отказ (WR-03), unknown/несвежая личность → тихий no-op (D-03, fail-closed второй пояс), covered → тихий no-op (D-01), full → тихий no-op (отказ виден состоянием пункта, S9), pattern := regexp.QuoteMeta(app) ДО writer'а (S8), запись через AppendAutocorrectBlocklist (S3 — единственный persist path), syncReload + FoldAppliedConfig в том же клике (SC1 «замолкает без рестарта»)
  - Проводка OnConn: Callbacks.BlockCurrentApp = toggles.block.add; installMenuState — первая покраска меню (07-06 install-пуши + начальное состояние blocklist-пункта из BlocklistAppState, до SetMenuSync); actor.EnsureAppid рядом с SetMenuSync (безусловный старт наблюдателя, Pitfall-2-лечение, RESEARCH Open Q1)
  - Гвард-матрица main_test.go: silent-матрица (writer-счётчик = 0, документ байт-нетронут, журнал без WARN и без identity — D-20/D-21), громкие отказы (WR-03 — гвард раньше identity-чтения; writer-error — WARN исход+путь), порядок гвард + QuoteMeta-точность (метасимвол org\.example\.App\[1\])
  - Потолок — единый источник: session-константа autocorrectBlocklistCeil=64 (временная мера 11-02) удалена, blocklistStateLocked читает config.MaxAutocorrectBlocklist
  - Трассер in-process всей цепочки: клик id 16 → файл вырос ровно на QuoteMeta-паттерн → строгий Load зелёный → fold прислал covered=true дельтой без второго клика → гейт автокоррекции воздержался app-blocked (SC1/SC4)
affects: [11-05 (e2e menu-blocklist: клик id 16 живым com.canonical.dbusmenu.Event на ВМ)]

# Actuals (#2632)
actuals:
  tokens: 6300 # chars/4 по фактическому диффу плана (610 строк changed)
  tasks: 2
  commits: 6 # MEASURED: git rev-list --count 38d6c445..HEAD; из них 5 коммитов плана + 1 сторонний docs(conventions) de702f6 владельца, попавший в диапазон

# Tech tracking
tech-stack:
  added: [] # ноль новых модулей — go.mod/go.sum нетронуты (tidy-diff зелёный, T-11-SC accept)
  patterns:
    - "click-composition-guard-belt: композиция клика = громкий путь-гвард (WR-03) → тихие identity-гварды (D-03/D-01/потолок) → QuoteMeta → writer → sync reload; recording-writer + fake-аксессор + счётчик вызовов — инъекция зависимостей по образцу configToggle-полей, тестируемый шов вместо глобальных вызовов"
    - "install-state-extraction: первая покраска меню — одна функция installMenuState (install-пуши 07-06 + начальное состояние пункта), OnConn держит порядок Attach → Store → paint → SetMenuSync → EnsureAppid"

key-files:
  created:
    - ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-04-SUMMARY.md"
  modified:
    - "cmd/goswitchd/main.go (blocklistToggle.add — гвард-композиция клика; newMenuToggles.block; OnConn: BlockCurrentApp + installMenuState + EnsureAppid)"
    - "cmd/goswitchd/main_test.go (трассер TestBlocklistClickTracer → 3 subtest-funcs; гвард-матрица TestBlocklistGuardMatrixSilent/Refusals/Order; assertGuardJournal)"
    - "internal/session/actor.go (autocorrectBlocklistCeil удалён; blocklistStateLocked читает config.MaxAutocorrectBlocklist)"
    - "internal/indicator/menu.go (комментарий labelBlocklistFull синхронизирован со сведённой константой — comment-only)"

key-decisions:
  - "Гвард-матрица — пин-корпус на готовой композиции: трассер (Task 1) по дизайну плана построил композицию production-качества СО ВСЕМИ гвардами (5 шагов в его action); матрица Task 2 закрепляет 8 behavior-строк поверх — RED-гейт плана закрыт трассером (evidence RED_EVIDENCE_OK, test a2275a8 < feat c46e633);fail-fast «unexpected GREEN — investigate» отработан: фича существует по дизайну трассера"
  - "installMenuState — одна точка первой покраски меню: install-пуши 07-06 + начальное состояние blocklist-пункта ДО SetMenuSync (меню дедуплицирует identical install-пуш сеттера); попутно лечение funlen OnConn (41 statement)"
  - "Сведение потолка байт-безопасно: 64 == config.MaxAutocorrectBlocklist, корпус session 11-02 зелёный БЕЗ правок (actor_test.go нетронут); forward-ссылка в комментарии menu.go синхронизирована той же правкой"
  - "Journal-вердикт гварда — отдельный helper assertGuardJournal: каждая строка матрицы пинит D-20/D-21 (WARN несёт исход и путь конфига, никогда паттерн/имя; silent-гварды не логируют вовсе)"

patterns-established:
  - "guard-matrix-runner: table-runner с recording-writer (счётчик вызовов «writer не вызван»), fake-аксессором BlocklistAppState и байт-сравнением документа до/после отказного клика"
  - "subtest-funcs-split: t.Run(name, namedFunc) — gocognit-лечение крупного теста без потери подтест-структуры"

requirements-completed: [MENU-01]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Клик id 16 сквозь композицию (трассер): fresh-identity «gedit» → файл вырос ровно на паттерн, строгий Load зелёный, ровно один apply-reload строго ПОСЛЕ записи, fold прислал covered=true третьим пушем (install+gain+covered) без второго клика"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "cmd/goswitchd/main_test.go#TestBlocklistClickTracer/testBlocklistClickLandsFresh (RED a2275a8 RED_EVIDENCE_OK → GREEN c46e633)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Гвард-матрица fail-closed: unknown (пустой кэш) и fresh=false (реальный FocusOut был) — тихие no-op; covered (D-01) и full — тихие no-op; все четыре — writer-счётчик 0, документ байт-нетронут; журнал без WARN и без identity"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "cmd/goswitchd/main_test.go#TestBlocklistGuardMatrixSilent (4 строки, recording-writer + байт-сравнение)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Громкие отказы и порядок: пустой cfgPath → один WARN «no config path» ДО identity-чтения (WR-03 — состояние пункта значения не имеет); ошибка writer'а → WARN исход+путь без содержимого, файл байт-нетронут; порядок гвард → ровно один write с QuoteMeta-паттерном → reload тем же кликом"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "cmd/goswitchd/main_test.go#TestBlocklistGuardRefusals + #TestBlocklistGuardOrder (включая failure-path QuoteMeta-пин S8)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Метасимволы и гейт-тишина: «org.example.App[1]» литералом org\\.example\\.App\\[1\\], строгий Load принимает ровно один элемент; с blocklist из клика feedKey-раунд воздерживается app-blocked, разделитель транзитом (без правок существующих гейт-тестов)"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "cmd/goswitchd/main_test.go#TestBlocklistClickTracer/testBlocklistClickMetacharLiteral + /testBlocklistClickGateSilent (SC4)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Сведение потолка на единый источник: config.MaxAutocorrectBlocklist в blocklistStateLocked, локальная константа удалена (grep по репозиторию — ноль вхождений autocorrectBlocklistCeil), поведение full байт-то же (session-корпус 11-02 зелёный без правок, -race)"
    requirement: MENU-01
    verification:
      - kind: other
        ref: "grep -q 'config.MaxAutocorrectBlocklist' internal/session/actor.go — OK; grep -rn autocorrectBlocklistCeil — ноль; go test ./internal/session/ -race -count=1 — ok"
        status: pass
    human_judgment: false
  - id: D6
    description: "Проводка OnConn (Callbacks.BlockCurrentApp, install-пуш начального состояния через installMenuState, EnsureAppid рядом с SetMenuSync): каждый компонент спиннут юнит-корпусом (композиция — cmd, install-пуш/EnsureAppid — session 11-02), но сама сборка OnConn требует живой session bus"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "компоненты: TestBlocklistClickTracer (композиция через Callbacks), TestActor_EnsureAppidUnconditional + install-пуш SetMenuSync (session, 11-02)"
        status: pass
    human_judgment: true
    rationale: "Сборка OnConn на живом D-Bus — e2e-кейс 11-05 на ВМ goswitch-stand (директива 6 от 2026-11-01: проверки продукта — только на ВМ); хост-юнит-корпус доказывает каждый компонент шва, живую сборку доказывает 11-05"
  - id: D7
    description: "Зелёная итерация целиком (директива 2): go build ./..., go vet ./..., golangci-lint 2.15.0 0 issues (весь репозиторий), go test -race -count=1 ./... exit 0 (22 пакета), go mod tidy + git diff --exit-code -- go.mod go.sum чисто"
    requirement: MENU-01
    verification:
      - kind: other
        ref: "прямые гейты на HEAD e8f7f82 (literal mise run ci — toolchain-разъём сессии, WINDOWS #23/#24/#25, прецедент 11-01/11-02/11-03)"
        status: pass
    human_judgment: false

# Metrics
duration: 30min (двухсеансная дуга: Task 1 предыдущей сессией 22:29-22:34 UTC, Task 2 продолжением 22:35-22:53 UTC)
completed: 2026-11-01
status: complete
---

# Phase 11 Plan 04: MENU — клик-композиция «в чёрный список» + проводка OnConn Summary

**Последний сантиметр MENU-01: клик по пункту id 16 пишет regexp.QuoteMeta(свежей identity) в autocorrect.apps_blocklist через writer 11-03 и применяет документ в том же клике (syncReload + FoldAppliedConfig — автокоррекция замолкает без рестарта); гварды fail-closed D-01/D-03/потолка спиннуты матрицей 8 строк; проводка OnConn (callback + install-пуш + EnsureAppid) собирает пункт в живого демона; потолок сведён на config.MaxAutocorrectBlocklist.**

## Performance

- **Duration:** 30 min (двухсеансная дуга: Task 1 — предыдущая сессия исполнителя, Task 2 — сессия-продолжение)
- **Started:** 2026-10-31T22:29Z (RED-коммит трассера; работа над корпусом ранее в той же сессии)
- **Completed:** 2026-10-31T22:53 UTC / 2026-11-01 01:53 +0300
- **Tasks:** 2 (Task 1 tracer: полный RED→GREEN; Task 2: пин-корпус → lint-сплит → сведение потолка)
- **Files modified:** 4 (main.go, main_test.go, actor.go, menu.go comment)

## Accomplishments

- **Трассер id 16 сквозь композицию** (Task 1, RED a2275a8 → GREEN c46e633): in-process стенд — реальный writer на temp-документе, fake-меню (SetBlocklistApp-рекордер), fold через value-source; клик через `Callbacks.BlockCurrentApp` (цель dispatchClick(menuIDBlocklistApp)) пишет QuoteMeta-паттерн, применяет документ тем же кликом (write строго до reload — Reload-колбэк читает файл ДО перечитки), fold пересчитывает covered и пушит галочку третьей дельтой БЕЗ второго клика; метасимвол `org.example.App[1]` садится литералом `org\.example\.App\[1\]` и переживает строгий Load (SC4); с blocklist из клика существующая блок-машина гейта воздерживается app-blocked, разделитель транзитом. Композиция `blocklistToggle.add` — форма configToggle.flip: путь-гвард (WR-03, WARN) → identity-гварды (тихие no-op D-03/D-01/потолок) → QuoteMeta (S8) → AppendAutocorrectBlocklist (S3) → sync reload + FoldAppliedConfig. Проводка OnConn: `Callbacks.BlockCurrentApp = toggles.block.add`, install-пуш начального состояния пункта из `BlocklistAppState()` до `SetMenuSync`, `actor.EnsureAppid()` рядом (безусловный старт наблюдателя — Pitfall 2, Open Q1)
- **Гвард-матрица** (Task 2, test 3bcdaaa): 8 behavior-строк таблицей — silent-матрица unknown/fresh=false/covered/full (recording-writer: счётчик 0, документ байт-нетронут; assertGuardJournal: ни WARN, ни identity в журнале — D-20/D-21), громкие отказы empty-path (WR-03: гвард раньше identity-чтения — известное covered/full-состояние всё равно WARN'ит) и writer-error (WARN исход+путь; failure-path QuoteMeta-пин S8), порядок гвард с метасимвольной строкой
- **Lint-сплит зелёной итерации** (Task 2, refactor 56eba96): TestBlocklistClickTracer → 3 именованные subtest-funcs (gocognit 35>30 на GREEN-коде — lint не прогонялся в предыдущей сессии), `installMenuState` извлечена из OnConn (funlen startCtl 41>40) — одна точка первой покраски меню
- **Сведение потолка** (Task 2, refactor e8f7f82): `autocorrectBlocklistCeil` удалён, `blocklistStateLocked` читает `config.MaxAutocorrectBlocklist` (пин 11-03, grep-гейт плана зелёный, по репозиторию ноль вхождений старого имени); поведение full байт-то же — session-корпус 11-02 зелёный без единой правки; комментарий labelBlocklistFull в menu.go синхронизирован
- **Вертикаль и весь репозиторий зелёные под -race**; go.mod/go.sum нетронуты (T-11-SC accept: ноль новых модулей)

## Task Commits

1. **Task 1 RED: tracer corpus** - `a2275a8` (test) — корпус трассера; RED evidence `RED_EVIDENCE_OK` (target_test_failed, 4/4 целевых упали на planned-behavior assertions; evidence-20261101-11-04-red-task1.json)
2. *(сторонний коммит владельца в диапазоне)* - `de702f6` (docs) — директива 2026-11-01 «проверки — только на ВМ»; не работа плана, попал в измеренный диапазон ledger
3. **Task 1 GREEN: композиция + проводка** - `c46e633` (feat) — blocklistToggle.add с гвардами, newMenuToggles.block, OnConn-проводка (BlockCurrentApp, install-пуш, EnsureAppid)
4. **Task 2 corpus: гвард-матрица** - `3bcdaaa` (test) — 8 behavior-строк, recording-writer, assertGuardJournal (D-20/D-21), failure-path QuoteMeta-пин
5. **Task 2 lint-сплит** - `56eba96` (refactor) — трассер 3 subtest-funcs (gocognit), installMenuState из OnConn (funlen)
6. **Task 2 сведение потолка** - `e8f7f82` (refactor) — config.MaxAutocorrectBlocklist единый источник, константа удалена, menu.go-комментарий синхронизирован

## Tracer Feedback Gate

Task 1 — tracer. Авто-режим (auto_advance=true), verify только automated, human_verify_mode=end-of-phase. Verify перезапущен end-to-end на каждой ступени продолжения и на финальном HEAD: вертикаль (cmd/goswitchd, session, config, indicator) зелёная под -race, затем весь репозиторий `go test -race -count=1 ./...` exit 0. `⚡ Tracer verified end-to-end — expanding` — Task 2 исполнялся без checkpoint.

## Files Created/Modified

- `cmd/goswitchd/main.go` — blocklistToggle (path/state/write/reload поля — инъекция по образцу configToggle) + add(): путь-гвард WARN → identity-гварды тихим no-op → QuoteMeta → writer → sync reload; newMenuToggles.block; OnConn: BlockCurrentApp + installMenuState + EnsureAppid
- `cmd/goswitchd/main_test.go` — трассер (3 subtest-funcs: lands-fresh/metachar/gate-silent; blocklistFixture, clickBlocklistApp, fakeMenuSync/fakeAppID/fakeEngineEmitter); гвард-матрица (blocklistGuardCase, runBlocklistGuard, assertGuardJournal; TestBlocklistGuardMatrixSilent/Refusals/Order)
- `internal/session/actor.go` — сведение потолка: константа удалена, config.MaxAutocorrectBlocklist в blocklistStateLocked
- `internal/indicator/menu.go` — комментарий labelBlocklistFull синхронизирован (comment-only)

## Decisions Made

- **Гвард-матрица как пин-корпус.** План назначил Task 2 tdd=true, но по собственному дизайну трассер (Task 1) построил композицию production-качества со всеми гвардами — матрице нечего RED-ить: fail-fast-правило «unexpected GREEN — investigate» дало ответ «фича существует по дизайну трассера». Пины committed test'ом; план-уровневый RED-гейт закрыт трассером (RED_EVIDENCE_OK, test a2275a8 < feat c46e633 — проверено машиной и на продолжении)
- **installMenuState — одна точка первой покраски.** Extract funlen-лечения совпал с семантикой: все install-пуши меню (07-06 + blocklist-состояние) — одна функция, порядок Attach → Store → paint → SetMenuSync → EnsureAppid читается в OnConn без шума
- **Failure-path — тоже S8.** Ошибка writer'а не отменяет того, что паттерн дошёл до writer'а заквоченным: wantPattern пинит QuoteMeta и на отказной строке матрицы
- **Порядок гвард доказан состоянием.** Empty-path строка матрицы подаёт known+covered+full одновременно: WARN «no config path» возможен только если путь-гвард стоит ДО identity-гвардов (WR-03-форма)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Runner гвард-матрицы не реализовывал собственный контракт pattern-assertion**
- **Found during:** Task 2 (RED-итерация корпуса)
- **Issue:** runBlocklistGuard ассертировал паттерн на каждой строке безусловно; writer-failure строка оставляла wantPattern пустым («no pattern assertion» по контракту поля) — честный прогон падал на `write pattern = "gedit", want ""` — отказ корпуса, не поведения
- **Fix:** ассертика условна по контракту (`tc.wantPattern != "" && ...`) + фактический пин wantPattern: blocklistTestApp на failure-строке (S8 держится и на failure path)
- **Files modified:** cmd/goswitchd/main_test.go
- **Verification:** корпус зелёный; строка writer-failure пинит QuoteMeta
- **Committed in:** 3bcdaaa

**2. [Rule 2 - Missing critical] Lint не был зелёный на GREEN-коде Task 1 (директива 2)**
- **Found during:** Task 2 (verify-шаг)
- **Issue:** golangci-lint (строгий v2, default: all) поднимал 3 замечания на код плана: gocognit 35>30 TestBlocklistClickTracer, funlen startCtl 41>40 (оба — GREEN c46e633; lint в предыдущей сессии не прогонялся — toolchain-разъём, см. #3), cyclop 16>15 runBlocklistGuard (незафиксированный корпус)
- **Fix:** трассер → 3 именованные subtest-funcs; installMenuState извлечена из OnConn; journal-вердикт вынесен в assertGuardJournal
- **Files modified:** cmd/goswitchd/main.go, cmd/goswitchd/main_test.go
- **Verification:** golangci-lint 2.15.0 — 0 issues (весь репозиторий), корпус зелёный, поведение байт-то же
- **Committed in:** 56eba96 (+ assertGuardJournal в 3bcdaaa)

**3. [Rule 3 - Blocking] literal `mise run ci` в сессии исполнителя падает на toolchain-разъёме (4-е проявление, прецедент 11-01/11-02/11-03)**
- **Found during:** Task 2 (verify-шаг)
- **Issue:** shell-задачи mise резолвят системный go 1.27.3 при GOROOT mise-инсталла 1.23.12 («compile: version go1.23.12 does not match go tool version go1.27.3»); `mise exec -- go build/test` работает, `mise run` — нет
- **Fix:** те же гейты исполнены напрямую на HEAD e8f7f82: go build ./..., go vet ./..., golangci-lint 2.15.0 (бинарник mise-инсталла, GOROOT сброшен) — 0 issues, go test -race -count=1 ./... (exit 0), go mod tidy + git diff --exit-code (чисто); записано в WINDOWS.md (#25, kind unrun-verify)
- **Files modified:** нет (методика, не код)
- **Verification:** все гейты зелёные на финальном HEAD
- **Committed in:** — (методика, не код)

**4. [Rule 1 - Bug] Forward-ссылка комментария menu.go на сведённую константу**
- **Found during:** Task 2 (GREEN сведения)
- **Issue:** комментарий labelBlocklistFull ссылался на «session-side autocorrectBlocklistCeil» и «the reconciliation rides plan 11-04's point edit» — после сведения ссылка висячая (файл сверх files_modified плана)
- **Fix:** комментарий синхронизирован (comment-only: константа 11-04 сведена, единый источник назван)
- **Files modified:** internal/indicator/menu.go
- **Verification:** grep autocorrectBlocklistCeil по репозиторию — ноль
- **Committed in:** e8f7f82

---

**Total deviations:** 4 auto-fixed (2 bug, 1 missing-critical-директива 2, 1 blocking-окружение). **Impact on plan:** все отклонения — обслуживание зелёной итерации и честности корпуса; объём плана не менялся, go.mod/go.sum нетронуты, новых семантик композиции нет (гварды — из action Task 1 дословно).

## Issues Encountered

- **Директива владельца 2026-11-01 (de702f6, bind all phases): «проверки — только на ВМ».** Юнит/build/lint-гейты — in-process, десктоп не трогают (прогон завершён до директивы и повторён после — сеанс жив, a11y/ibus/gsettings/e2e-пробы на хосте не запускались); живая верификация пункта (клик в реальном меню, e2e 11-05) — только на ВМ goswitch-stand. OnConn-проводка поэтому покрыта юнит-корпусом покомпонентно, живая сборка — за 11-05 (coverage D6, human_judgment)
- literal `mise run ci` остаётся недоступен в сессии исполнителя (WINDOWS #23/#24 → #25) — прямые гейты эквивалентны и зелёные; починка сессийного PATH/GOROOT разъёма — отдельная задача вне фазы
- До-существующий untracked-хвост репозитория (.gsd/, .zcode/, draft.txt, goswitchd, test/e2e/e2e и пр.) не создавался этим планом и не трогался

## TDD Gate Compliance

| Task | RED | GREEN | REFACTOR | Status |
|------|-----|-------|----------|--------|
| 1 (tracer) | a2275a8 (test) | c46e633 (feat) | 56eba96 (lint-сплит корпуса трассера, попутно main.go) | Pass |
| 2 | 3bcdaaa (test, пин-корпус) | e8f7f82 (refactor сведения; feat-поведения нет — гварды уже в GREEN Task 1) | 56eba96 (lint-сплит) | Pass (pins) |

- RED evidence Task 1 проверено машиной дважды (предыдущей сессией и продолжением): `gsd_run check tdd-red-evidence evidence-20261101-11-04-red-task1.json` → `RED_EVIDENCE_OK` (target_test_failed, TestBlocklistClickTracer)
- Порядок gate-коммитов плана: test(11-04) a2275a8 < feat(11-04) c46e633 — план-гейт закрыт; дополнительные коммиты Task 2 порядок не нарушают
- Task 2 честно помечен «pins»: его 8 behavior-строк не имели RED-able поведения (композиция со всеми гвардами построена GREEN Task 1 по дизайну плана — tracer = production-quality), единственная production-правка (сведение) — байт-та же семантика, потому refactor, не feat; machine-check на Task 2 не синтезировался (честность выше галочки)
- Compile-stub не потребовался: corpus Task 1 компилировался против готовых символов 11-02/11-03 (writer, BlocklistAppState, EnsureAppid), RED-провалы — на planned-behavior assertions

## Flagged Assumptions

План не добавлял новых flagged assumptions. Открытая фазовая строка из 11-02 (`{"requirement_id":"MENU-01","category":"unclassified","status":"unresolved"}` — край «неизвестная личность») остаётся unresolved до фазовой верификации; Verifier обязан учитывать.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- **11-05 (e2e)** получает: живую композицию id 16 (клик → запись → перечитка → галочка → тишина гейта — вся цепочка спиннута юнит-трассером, e2e-драйвер повторяет её живым Event'ом), гвард-матрицу как спецификацию ожиданий (covered-клик no-op, метасимвол литералом), oracle «документ вырос ровно на паттерн + строгий Load». ВНИМАНИЕ: по директиве 6 e2e — только на ВМ goswitch-stand, хост-пробы запрещены
- MENU-01 остаётся shared-ID гейтом (11-05 объявляет его же) — requirements.ready-ids не закроет требование до 11-05 SUMMARY
- Открытые хвосты фазы: Round 2 journal-verify (11-01, ряд на сессии владельца); WINDOWS #23/#24/#25 (mise run ci, средовой)
- Блокеров для следующего плана нет

## Self-Check: PASSED

- Файлы: cmd/goswitchd/main.go, cmd/goswitchd/main_test.go, internal/session/actor.go, internal/indicator/menu.go, 11-04-SUMMARY.md — FOUND на диске
- Коммиты: a2275a8, c46e633, 3bcdaaa, 56eba96, e8f7f82 (план) + de702f6 (сторонний в диапазоне) — все FOUND в git log; TDD-порядок test-before-feat плана выдержан
- commits измерены от ledger-базы 38d6c445: git rev-list --count = 6 (5 план + 1 сторонний)
- Acceptance criteria перепроверены на HEAD e8f7f82: трассер 5 behavior-строк; метасимвол литералом + строгий Load; write строго до reload; гейт-тест без правок существующих гейт-тестов; 8 строк матрицы, «writer не вызван» счётчиком; covered байт-идентичен; grep config.MaxAutocorrectBlocklist зелёный; session-тесты 11-02 без диффа; go.mod/go.sum diff --exit-code чист; весь репозиторий -race зелёный; golangci-lint 0 issues
- Блокеров для следующего плана нет

---
*Phase: 11-menu-punkt-v-chernyy-spisok*
*Completed: 2026-11-01*
