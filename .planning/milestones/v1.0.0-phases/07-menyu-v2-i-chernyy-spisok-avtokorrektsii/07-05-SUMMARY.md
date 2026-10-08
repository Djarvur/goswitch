---
phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
plan: 05
subsystem: tray-menu
tags: [dbusmenu, sni, gnome-shell, godbus, items-properties-updated, tray, menu-v2]

# Dependency graph
requires:
  - phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii (07-03)
    provides: persist writer SetAutocorrectEnabled/SetSoundEnabled/EnsureDocument/DefaultPath + adopt+watch
  - phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii (07-04)
    provides: final blocklist gate semantics (StatusSnapshot autocorrect truth)
  - phase: quick 261001-fg3
    provides: DBusMenu channel (com.canonical.dbusmenu at /Menu), SNI item, Callbacks, recoverMenuCall, supervisor
provides:
  - Menu v2 serving the canonical layout (287999c) from a mutex-guarded state snapshot: EN/RU radio pair, «Автокоррекция»/«Звук» checkmark toggles, five grey macro rows (empty binding = row disappears), separators, Настройки/О программе/Статус/Перечитать конфиг — IDs pinned menuIDEN=1..menuIDReload=15
  - ItemsPropertiesUpdated dynamics (empty non-nil removed slice, byte-form SignatureOf pins) + LayoutUpdated + revision counter on set changes only
  - Actor seam: SwitchMode(target) (the flipTo wrapper — ADR-006 single flip path), MenuSync point-of-use interface + SetMenuSync install push, applySnapshot folds pushing applied config truth, Status.SoundEnabled + the sound_enabled= status token
  - Daemon wiring: toggle compositions (snapshot → invert → write → menu push → sync reload), no-pipes xdg-open settings launcher, menu seeding from the startup config
affects: [07-06 (e2e menu cases assert the pinned IDs and the status token), 07-07 (release docs — the README menu list must match menu.go)]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
actuals:
  tokens: 24674    # chars/4 over the realized diff (98695 chars, 9 files, +1826/-212)
  tasks: 3
  commits: 6

# Tech tracking
tech-stack:
  added: []        # zero new dependencies (godbus/dbus only, already in go.mod)
  patterns:
    - "State-snapshot menu: properties computed at request time under a mutex; every snapshot change emits exactly one ItemsPropertiesUpdated with the affected deltas and an EMPTY non-nil removed slice (Pitfall 7)"
    - "Menu-side dedupe: the actor pushes per flip/fold, the menu drops identical values — the per-keystroke fold never spams the bus"
    - "procStarter func-var seam: exec.Command with nil pipes + Start + deadline-free Wait reaper (the wl-copy fork-shaped precedent)"
    - "configToggle composition: snapshot → invert → write → push → reload, one type for both persisted toggles"

key-files:
  created: []
  modified:
    - internal/indicator/menu.go (v2: constants, snapshot, setters, layout, dispatch, emitters)
    - internal/indicator/menu_test.go (v2 corpus: 13 TestMenu* pins)
    - internal/indicator/indicator.go (Callbacks +5 nil-safe fields; Item.menu; Menu(); re-attach survival)
    - internal/session/actor.go (SwitchMode; MenuSync + SetMenuSync; flipTo/syncMode/applySnapshot push points; soundEnabled)
    - internal/session/actor_test.go (menu-sync corpus: fakeMenuSync + 5 tests)
    - cmd/goswitchd/main.go (startCtl wiring; configToggle; configEditor; newEditorProc seam)
    - cmd/goswitchd/main_test.go (wiring corpus: composition + launcher)
    - internal/ctlsvc/ctlsvc.go (sound_enabled status token) [Rule 3]
    - internal/ctlsvc/ctlsvc_test.go (token pin) [Rule 3]

key-decisions:
  - "Menu dedupes identical pushes: the actor pushes SetMode/SetAutocorrectEnabled/SetSoundEnabled/SetKeys on every flip and every applySnapshot fold (per keystroke); the menu-side dedupe keeps unchanged pushes a cheap no-op — the bus is never spammed"
  - "The wiring seeds the menu's initial state from the STARTUP config (version, both toggles, raw key names) right after Attach; SetMenuSync's install push shows the current mode — the actor re-folds the same truth per event, idempotently"
  - "A macro-row binding entering/leaving is the ONE legal revision bump (LayoutUpdated, full renderer re-read); label-only changes ride ItemsPropertiesUpdated — the item set is otherwise static after first construction"
  - "The launcher seam is exec.Command (no context): the reaper goroutine has NO deadline — nothing, including daemon shutdown, kills a running editor (the plan's 'таймаут-контекст НЕ рвёт запущенный редактор')"
  - "Item owns its Menu instance: the supervisor's re-registration re-exports the same live object, so the seeded snapshot survives every self-heal re-attach"

patterns-established:
  - "Point-of-use observer interface #3: MenuSync mirrors AppidSource (definition at the consumer) and ModeDisplay (install push of the current state, observer-last ordering)"
  - "RED-stub corpus shape: schema constants/fields/stub bodies compile the corpus, which fails on behavior assertions (06-04 precedent); TAP-transcribed tdd-red-evidence records per the 04-04 precedent"

requirements-completed: ["SWCH-01..04 (регресс)"]

coverage:
  - id: D1
    description: "Menu v2 serves the canonical layout (radio pair marked with the current mode, both checkmark toggles adjacent, five grey macro rows with doubled-mnemonic labels, two separators, four standard items) from the state snapshot; IDs pinned by constant"
    requirement: "SWCH-01..04 (регресс)"
    verification:
      - kind: unit
        ref: "tests/internal/indicator/menu_test.go#TestMenuLayout"
        status: pass
      - kind: unit
        ref: "tests/internal/indicator/menu_test.go#TestMenuGetGroupProperties"
        status: pass
      - kind: unit
        ref: "tests/internal/indicator/menu_test.go#TestMenuMacroVisibility"
        status: pass
      - kind: unit
        ref: "tests/internal/indicator/menu_test.go#TestMenuUnderscoreDoubling"
        status: pass
    human_judgment: false
  - id: D2
    description: "Dynamic channel: every snapshot change emits exactly one ItemsPropertiesUpdated (affected deltas, empty non-nil removed slice, a(ias)/a(ia{sv}) signatures pinned); revision bumps only on set changes (LayoutUpdated); nil-emitter degradation"
    verification:
      - kind: unit
        ref: "tests/internal/indicator/menu_test.go#TestMenuSignalFormPins"
        status: pass
      - kind: unit
        ref: "tests/internal/indicator/menu_test.go#TestMenuToggleAndVersionDeltas"
        status: pass
      - kind: unit
        ref: "tests/internal/indicator/menu_test.go#TestMenuNilEmitterDegradation"
        status: pass
    human_judgment: false
  - id: D3
    description: "EN/RU clicks dispatch nil-safe through Callbacks.Switch; the actor's SwitchMode flips through the SINGLE flipTo path with the same-target no-op guard; the mode push rides observer-last (after the display) on flip and drift; SWCH-01..04 mechanisms untouched (D-36 order pins green)"
    requirement: "SWCH-01..04 (регресс)"
    verification:
      - kind: unit
        ref: "tests/internal/indicator/menu_test.go#TestMenuEventDispatch"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_SwitchModeFlipsToTarget"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_MenuSyncSetModeOnFlip"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_ToggleModeFlips (D-36 order intact)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Persisted toggles: both compositions pin snapshot → invert → write (SetAutocorrectEnabled/SetSoundEnabled on the factual path) → menu push → synchronous reload; write failure stops everything; nil-reload keeps the write and WARNs; sound truth (EffectiveEnabled, absent section reads ON) reaches StatusSnapshot and the sound_enabled= token"
    verification:
      - kind: unit
        ref: "tests/cmd/goswitchd/main_test.go#TestConfigToggleComposition"
        status: pass
      - kind: unit
        ref: "tests/cmd/goswitchd/main_test.go#TestToggleNilReloadKeepsWrite"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_MenuSyncApplySnapshotPushesSound"
        status: pass
      - kind: unit
        ref: "tests/internal/ctlsvc/ctlsvc_test.go#TestRenderStatus_AutocorrectTokens (sound_enabled pin)"
        status: pass
    human_judgment: false
  - id: D5
    description: "«Настройки…»: EnsureDocument then xdg-open on the pinned path — argv [xdg-open, path], nil pipes, one Start, deadline-free Wait reaper, one WARN per failed-start episode, ensure failure skips the launch"
    verification:
      - kind: unit
        ref: "tests/cmd/goswitchd/main_test.go#TestOpenConfigEditorLauncher"
        status: pass
      - kind: unit
        ref: "tests/cmd/goswitchd/main_test.go#TestOpenConfigEditorStartErrorWarnsOnce"
        status: pass
      - kind: unit
        ref: "tests/cmd/goswitchd/main_test.go#TestOpenConfigEditorEnsureFailureSkipsLaunch"
        status: pass
    human_judgment: false
  - id: D6
    description: "Live tray menu renders correctly on the owner's desktop: radio dots/checkmarks drawn by the extension, greyed macro rows, doubled underscores actually displaying single ones, close-open replay of flagged updates"
    verification: []
    human_judgment: true
    rationale: "The renderer is gnome-shell JS (dbusMenu.js) — the unit corpus pins the daemon's wire bytes, but the pixel truth (ornaments, label rendering after mnemonic stripping, live updates behind a closed menu) is only observable in the owner's session; the plan explicitly reserves the UAT screenshot (Pitfall 1) and research A1/A2 carry the FLAGGED renderer assumptions"

# Metrics
duration: 62min
completed: 2026-10-04
status: complete
---

# Phase 7 Plan 05: Menu v2 (canonical layout, state snapshot, actor seam, daemon wiring) Summary

**Трей-меню v2 по каноническому макету 287999c: радиопара EN/RU и тумблеры «Автокоррекция»/«Звук» из мьютекс-снимка состояния с ItemsPropertiesUpdated-динамикой, серые макро-строки с исчезновением при пустом биндинге, «Настройки…» (ensure + xdg-open no-pipes), «О программе»; шов актора SwitchMode→flipTo (единственный путь флипа) + MenuSync-пуши + токен sound_enabled.**

## Performance

- **Duration:** 62 min (2026-10-04T22:42:58Z → 2026-10-04T23:44:34Z)
- **Tasks:** 3 (все; каждая RED → GREEN)
- **Files modified:** 9 (+1826/−212)
- **Commits:** 6 (3× test RED, 3× feat GREEN)

## Accomplishments

- Menu v2 обслуживает полный канонический набор пунктов из мьютекс-снимка состояния в момент запроса: радиопара EN/RU с отметкой текущего режима, чекмарк-тумблеры «Автокоррекция» и «Звук» подряд, пять серых макро-инфострок с живыми именами клавиш (удвоение подчёркиваний; пустой биндинг — строка исчезает из layout), два сепаратора, «Настройки…»/«О программе»/«Статус»/«Перечитать конфиг» — ID пинены константами menuIDEN=1..menuIDReload=15 (старый пункт «Переключить раскладку» выведен).
- Каждое изменение снимка эмитит ровно одну ItemsPropertiesUpdated с дельтами затронутых пунктов и ПУСТЫМ не-nil removed-срезом (Pitfall 7, SignatureOf-пины a(ias)/a(ia{sv})); появление/исчезновение макро-строки — единственный законный бамп ревизии (LayoutUpdated); идентичные пуши дедупятся.
- Шов актора: SwitchMode(target) — публичная обёртка flipTo (same-target guard, без дублирования проверки; SWCH-регресс — единственный путь флипа не разветвлён, D-36-пины зелёные); MenuSync (point-of-use, форма AppidSource) + SetMenuSync в форме SetModeDisplay (инсталл-пуш текущего режима); push-точки flipTo/syncMode (observer-last, после display) и applySnapshot (SetAutocorrectEnabled + SetSoundEnabled из EffectiveEnabled — absent-секция читается «включён» + SetKeys сырыми именами); StatusSnapshot.SoundEnabled + токен sound_enabled= в ctl-статусе.
- Wiring демона: клики EN/RU/обоих тумблеров/Настроек/О программе через пять новых nil-safe полей Callbacks; обе тумблер-композиции (снимок → инверт → config.SetAutocorrectEnabled/SetSoundEnabled на фактическом пути → пуш меню → синхронный reloadConfig); «Настройки…» — EnsureDocument + xdg-open no-pipes (argv [xdg-open, путь], Start + дедлайн-фри Wait-reaper, одна WARN на эпизод неудачного старта); меню засеивается версией/тумблерами/ключами из стартового конфига и ставится актору через SetMenuSync.

## TDD Gate Compliance

- Task 1 (tracer): RED 077e19b (8 целевых тестов падают на утверждениях: «root layout = 0 kids, want exactly 15»; «Menu без состояния») → GREEN 5b7671f.
- Task 2 (tdd="true"): RED f237090 → `check tdd-red-evidence` = **RED_EVIDENCE_OK** (target_test_failed; 4 из 5 падают, nil-пин зелёный by design — прецедент continuity-pin) → GREEN 5473dce. Гейт соблюдён: feat не предшествует test в рамках плана.
- Task 3 (auto; строгий TDD директивой 1 CLAUDE.md): RED 9207af2 → RED_EVIDENCE_OK → GREEN 34aecb7.
- REFACTOR-коммитов нет — рефакторинг (хелперы dispatchClick/pushMenuSync, сплит тяжёлых тестов) вошёл в GREEN-коммиты под давлением lint-гейтов; тесты зелёные на каждом шаге.
- Записи RED-свидетельств: `.planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/red-evidence/07-05-task{2,3}-red.json` (TAP-транскрипция по прецеденту 04-04 — SDK-верб парсит только node:test TAP).

## Task Commits

1. **Task 1 (tracer): радиопара EN/RU end-to-end** — `077e19b` (test RED) + `5b7671f` (feat GREEN)
2. **Task 2: шов актора SwitchMode + SetMenuSync** — `f237090` (test RED) + `5473dce` (feat GREEN)
3. **Task 3: wiring демона** — `9207af2` (test RED) + `34aecb7` (feat GREEN)

## Files Created/Modified

- `internal/indicator/menu.go` — menu v2: константы канонического макета, снимок состояния, сеттеры с дедупом и эмитами, layout/dispatch/props-билдеры, удвоение мнемоник
- `internal/indicator/menu_test.go` — корпус v2: 13 TestMenu* (layout, сигналы, видимость, мнемоники, диспетчер, nil-деградации, wire-имена)
- `internal/indicator/indicator.go` — Callbacks +Switch/ToggleAutocorrect/ToggleSound/Settings/About (nil-safe); Item.menu + Menu(); меню переживает re-attach супервизора
- `internal/session/actor.go` — SwitchMode; MenuSync/SetMenuSync; пуши flipTo/syncMode/applySnapshot (pushMenuSync); soundEnabled; константы symbolEN/symbolRU
- `internal/session/actor_test.go` — fakeMenuSync + 5 шов-кейсов
- `cmd/goswitchd/main.go` — startCtl(cfg, cfgPath); configToggle/configEditor/newEditorProc-seam; OnConn v2: пиненые композиции, лаунчер, засев меню, SetMenuSync
- `cmd/goswitchd/main_test.go` — wiring-корпус: композиции обоих тумблеров, nil-reload, лаунчер (argv/pipes/reap/WARN-once/ensure-skip)
- `internal/ctlsvc/ctlsvc.go` — токен `sound_enabled=` (форма autocorrect_enabled, до конфиг-блока) [Rule 3]
- `internal/ctlsvc/ctlsvc_test.go` — пин токена [Rule 3]

## Decisions Made

- **Дедуп на стороне меню**: актор пушит на каждый флип/фолд (фолд — на каждое нажатие клавиши); сеттеры меню молчат на идентичных значениях — шина не спамится, двойное применение (клик + эхо + фолд) безвредно по построению.
- **Засев меню из стартового конфига** в OnConn сразу после Attach (version, оба тумблера, сырые имена ключей) — план требует «начальное состояние» при attach; истина та же, что актор пере-фолдит по событиям.
- **Item владеет своим Menu**: супервизор ре-экспортирует тот же живой объект — засеянный снимок переживает каждый self-heal re-attach (Rule 2: без этого re-attach сбрасывал бы состояние меню).
- **Лаунчер без контекста** (exec.Command, не CommandContext): reaper без дедлайна — ничто, включая остановку демона, не убивает запущенный редактор (требование плана; T-07-05-02 держится на nil-пайпах и Start+Wait).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Токен sound_enabled в ctl-статусе**
- **Found during:** Task 2 (GREEN)
- **Issue:** must_haves требуют `sound_enabled` токен ctl-оракулом e2e 07-06 («форма autocorrect_enabled»), но internal/ctlsvc отсутствует в files_modified — без токена новое поле Status ничего не сериализует
- **Fix:** renderStatus дополнен `"sound_enabled="+FormatBool(st.SoundEnabled)` перед ac_skip-блоком; пин в TestRenderStatus_AutocorrectTokens
- **Files modified:** internal/ctlsvc/ctlsvc.go, internal/ctlsvc/ctlsvc_test.go
- **Verification:** корпус ctlsvc зелёный; токен стоит до конфиг-блока, config_error остаётся последним
- **Committed in:** 5473dce

**2. [Rule 2 - Missing Critical] Меню переживает re-attach супервизора**
- **Found during:** Task 1 (GREEN, при проводке эмиттера)
- **Issue:** registerItem создавал НОВЫЙ Menu при каждой регистрации; супервизор (quick 261001-fg3) перезапускает её на каждом self-heal — засеянное состояние (режим, тумблеры, ключи, версия) сбрасывалось бы при первом re-attach
- **Fix:** Item владеет `menu *Menu` (создаётся в attach с эмиттером), registerItem экспортирует it.menu; геттер Menu() для wiring
- **Files modified:** internal/indicator/indicator.go
- **Verification:** существующий супервизорный корпус зелёный; снапшот-состояние живёт на Menu, не на регистрации
- **Committed in:** 5b7671f

**3. [CLAUDE.md Directive 1] Task 3 выполнен RED → GREEN**
- **Found during:** Task 3
- **Issue:** план типирует Task 3 как auto (один feat-коммит), но директива 1 CONVENTIONS.md требует строгий TDD на каждую задачу поведения — директива старше плана
- **Fix:** corpus (TestConfigToggleComposition/TestToggleNilReloadKeepsWrite/TestOpenConfigEditor*) с RED-стабами → test(07-05) 9207af2 → реализация → feat(07-05) 34aecb7; RED_EVIDENCE_OK
- **Files modified:** cmd/goswitchd/main.go, cmd/goswitchd/main_test.go
- **Committed in:** 9207af2 + 34aecb7

---

**Total deviations:** 3 auto-handled (1 blocking-missing-file, 1 missing-critical, 1 CLAUDE.md-процедурное). **Impact on plan:** все три необходимы для корректности/контракта фазы; scope creep нет.

## Issues Encountered

- SDK-верб `check tdd-red-evidence` парсит только node:test TAP-вывод — `go test -v` даёт `zero_tests_discovered` как артефакт формата парсера; записи транскрибированы в TAP по прецеденту 04-04 (04-04-SUMMARY.md:214), вербификация вернула RED_EVIDENCE_OK на обоих tdd-гейтах.
- Известные флейки (TestRun_OnConnHookCalledOnce, TestWatch_BrokenBlocklistPatternKeepsLastGood) не стреляли — оба корпуса зелёные с первого прогона в этой сессии.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- 07-06 получает всё заявленное в artifacts: пиненые ID (menuIDEN=1..menuIDReload=15) для Event-драйв e2e-кейса, токен `sound_enabled` для ctl-оракула, SwitchMode/SetMenuSync-швы для наблюдения, живой канал ItemsPropertiesUpdated (юнит-пины байт-форм).
- Отложено на verify-work/UAT владельца: пиксельная проверка живого меню (D6 — скриншот: орнаменты, серые строки, shift__r → shift_r после вырезания мнемоник; research A1 — целевой редактор xdg-open; A2 — стабильность расширения-рендерера).

---
*Phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii*
*Completed: 2026-10-04*

## Self-Check: PASSED

- 8 production/test files + SUMMARY + 2 red-evidence records present on disk (`[ -f ]` all FOUND)
- All 6 task commits present in history: 077e19b, 5b7671f, f237090, 5473dce, 9207af2, 34aecb7
- Plan-level verification re-run green: `go test ./internal/indicator/ ./internal/session/ ./cmd/goswitchd/ -race -count=1` + `mise run ci` (build/vet/lint/test)
- Греп-гейты плана: MENU-V2-OK, SEAM-OK, WIRING-OK; счётчик TestMenu* PASS = 13 (гейт 8-19); счётчик TestConfig|TestOpenConfig|TestToggle PASS = 5 (гейт 3-6)
