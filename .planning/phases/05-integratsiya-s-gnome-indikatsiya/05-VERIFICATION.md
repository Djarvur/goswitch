---
phase: 05-integratsiya-s-gnome-indikatsiya
verified: 2026-10-07T20:52:42Z
status: passed
score: 6/6 must-haves verified
covered_files:
  - .gitignore
  - .golangci.yml
  - .planning/REQUIREMENTS.md
  - .planning/WINDOWS.md
  - .planning/phases/05-integratsiya-s-gnome-indikatsiya/05-01-PLAN.md
  - .planning/phases/05-integratsiya-s-gnome-indikatsiya/05-01-SUMMARY.md
  - .planning/phases/05-integratsiya-s-gnome-indikatsiya/05-02-PLAN.md
  - .planning/phases/05-integratsiya-s-gnome-indikatsiya/05-02-SUMMARY.md
  - .planning/phases/05-integratsiya-s-gnome-indikatsiya/05-03-PLAN.md
  - .planning/phases/05-integratsiya-s-gnome-indikatsiya/05-03-SUMMARY.md
  - .planning/phases/05-integratsiya-s-gnome-indikatsiya/05-04-PLAN.md
  - .planning/phases/05-integratsiya-s-gnome-indikatsiya/05-04-SUMMARY.md
  - .planning/phases/05-integratsiya-s-gnome-indikatsiya/05-05-PLAN.md
  - .planning/phases/05-integratsiya-s-gnome-indikatsiya/05-05-SUMMARY.md
  - .planning/phases/05-integratsiya-s-gnome-indikatsiya/05-REVIEW.md
  - README.md
  - cmd/goswitchd/main.go
  - cmd/goswitchd/main_test.go
  - docs/adr/ADR-006-two-engine-revision.md
  - internal/activate/activate.go
  - internal/activate/activate_test.go
  - internal/engine/conn.go
  - internal/engine/conn_switcher_test.go
  - internal/engine/conn_sync_test.go
  - internal/engine/engine.go
  - internal/engine/factory.go
  - internal/engine/types.go
  - internal/install/install.go
  - internal/install/install_internal_test.go
  - internal/install/install_test.go
  - internal/install/selfcheck.go
  - internal/install/selfcheck_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - test/e2e/README.md
  - test/e2e/case_combo.go
  - test/e2e/case_install.go
  - test/e2e/case_resilience.go
  - test/e2e/case_switch.go
  - test/e2e/case_switch_test.go
  - test/e2e/main.go
  - test/e2e/matrix.go
covered_digest: "v1:sha256:15e3bf4febd3787f359f066cc9dfc6b722397b1f546d21e6d21cb874cfa5edbe"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: human_needed
  previous_score: 4/6
  gaps_closed: []
  gaps_remaining: []
  regressions: []
---

# Phase 5: Интеграция с GNOME — Verification Report (re-verification, milestone close)

**Phase Goal:** goswitch переключает раскладку как полноценный гражданин GNOME: верхняя панель показывает индикатор активной раскладки (en/ru), флип (одиночный Shift, Super+Space, флип после коррекции) переключает источники через SetGlobalEngine на шине IBus — видимым для GNOME образом. Ревизия ADR-001 → ADR-006 фиксируется. (Исходная формулировка «ДВА goswitch-источника» ревизована владельцем 2026-09-30 вечером — ADR-006 Amendment: single-source model, ≥1 goswitch-источник; собственный индикатор — SNI goswitch в трее.)
**Verified:** 2026-10-07T20:52:42Z (HEAD 2171fae, branch gsd/internalize-engine-layouts — последний код-коммит 5427d7a/08b598a)
**Status:** passed
**Re-verification:** Yes — fingerprint-staleness refresh at milestone close (previous: human_needed 4/6, 2026-09-30T15:10:58Z, HEAD bbbbd04)

**Scope of this pass.** Предыдущий отчёт признал 4/6 истин и вынес 6 human-гейтов. С тех пор дерево прошло 306 коммитов (фазы 6–8, PR #10, quick-задачи 260930-toa / 261001-fg3 / 261006-squ / 261006-vqw / 261007-0yg). Проверяются те же 6 must-have на ТЕКУЩЕМ дереве: полные проверки там, где пути/проводка менялись (engine/ → internal/engine/, actor.go/install.go/main.go/case_switch.go тронуты поздними фазами), точечные регрессии — где доказательства переносятся. Человеко-принимаемые измерения закрыты вердиктами владельца (в UAT фаз 6/8, записанных на диске, и вердиктами 2026-10-07, переданными координатором milestone-UAT-сессии); human-пункты здесь не возобновляются — по правилам этого прогона они возобновляются только при отсутствующем/несвязанном артефакте (таких нет).

**Path-staleness note:** все пути `engine/*.go` из прошлого отчёта устарели (коммит 5427d7a перенёс engine/ и layouts/ под internal/, SPEC §8 rev D-55). Все пути ниже re-выведены против текущего дерева; covered_digest пересчитан вербой verification.fingerprint.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | **SC-1** — install оборачивает источники пользователя в goswitch-движки их раскладки; непригодный ввод — честный отказ; uninstall восстанавливает оригинал дословно | ✓ VERIFIED | Текущее дерево, модель по ADR-006 Amendment 2026-09-30 (одновложенная ревизия: «в гномовском переключателе остается один источник» — дословная цитата владельца в ADR): `wrapSources` (install.go:830) оборачивает КАЖДЫЙ пригодный источник, **минимум один** (TestWrapSourcesSingleSource); отказы `errUnsupportedPair` (:168) и `errMixedSources` (:172, foreignResidueHint :858) ДО первой записи; `resolveWrapInput` → `savedSources` state-файла (:864-894); `TestUninstall_FullRollback` зелёный. Именованные прогоны этого верификатора на HEAD 2171fae (-race): TestWrapSourcesPairPreserved/SingleSource/ForeignResidueRefused/RefusalTable/AlreadyOwnedUpgrade, TestUninstall_FullRollback — все PASS. Полный живой install-cycle принят владельцем по доказательствам (2026-10-07, вердикт координатора) |
| 2 | **SC-2 (спек-дельта)** — индикатор отражает активный источник при пользовательских жестах; видим нативным для GNOME образом | ✓ VERIFIED (закрыт вердиктом владельца) | Предыдущий ⚠️ PRESENT_BEHAVIOR_UNVERIFIED закрыт: **владелец подтвердил 2026-10-07 ежедневным использованием, что индикатор следует переключениям источника** (вердикт передан координатором milestone-UAT; ранее — пиксельное подтверждение SNI-иконки 2026-09-30, quick 260930-pf6). Артефакты на текущем дереве wired: `UpdateModeSymbol` эмитится на каждом флипе (actor.go:1973, :2035; TestActor_FlipEmitsPanelSymbol PASS), display-шов `ModeChanged` (:1980) и трей-меню `menuSync.SetMode` (:1987) — тот же символ; интерактивный трей (261001-fg3): SNI Activate + DBusMenu, супервизор re-attach (main.go:166 `Toggle: actor.ToggleMode`). ADR-006 Amendment фиксирует дисплей-путь (нативный индикатор скрыт при одном источнике — SNI goswitch его замещает) |
| 3 | **SC-3** — все жесты через SetGlobalEngine на ibus-соединении демона (не subprocess), <50 мс; состояние демона не расходится с активным источником (sync при FocusIn; WARN при неудаче) | ✓ VERIFIED | Воронка едина и расширена: `flipTo` (actor.go:1954) — 4 классических сайта + **новый пятый жест** ToggleMode (:1050-1057, трей-меню/Activate) через ТОТ ЖЕ путь (TestActor_ToggleModeFlips PASS); `SetSwitcher`-замыкание (actor.go:840) → `CallWithContext(ibusService+".SetGlobalEngine")` (conn.go:240) с записью `switch_engine` + WARN при сбое (:242-246); `switchTimeout = 150ms` (actor.go:67) — клин-гард над бюджетом, не бюджет. Wiring в main.go (:554-569: OnGlobalEngine→SyncEngine, BindSwitcher→SetSwitcher, BindGlobalEngine→IfOwned) перепривязан на internal/engine — сборка зелёная, TestSwitcherCallsSetGlobalEngine (wire-акт через fake bus) PASS. Sync: TestActor_SyncEngineNeverSwitches PASS. Латентность: живой замер 05-05 ≈41 мс < 50 мс; **G-6-1 (ADR-006 Amendment 2026-10-02)** устранил самоблокировку фабрики — `AttachEngine` lock-free (`atomic.Pointer[emitterSlot]`, actor.go:171/:832-834): 24/24 немедленных буквы, RTT 6.5–9.2 мс, матрица 32/35×2 — бывший симптом «съеденной буквы» закрыт изменением механизма при неизменном контракте. Именованные прогоны: TestActor_FlipRoutesThroughSwitcher/FlipOrderPinned/SwitcherDeadlineBounded — PASS |
| 4 | **SC-4** — режим после коррекции = скрипт результата (lat→cyr оставляет goswitch-ru, наоборот — goswitch-en) | ✓ VERIFIED | `settleCorrectionFlip` (actor.go:1190) → `flipTo(scriptOf(converted))`, три сайта вызова (:2787, :2818, :2846) нетронуты ревизией mixed-семантики (261006-vqw переписал конвейер инверсии, НЕ флип-семантику). TestActor_CorrectionFlipSetsResultEngine — PASS на текущем дереве (оба направления) |
| 5 | **SC-5** — GNOME-переключение по клавишам остаётся у goswitch; selfcheck отражает состояние источников | ✓ VERIFIED (модель ревизована владельцем) | Selfcheck по ADR-006 Amendment: `checkInputSource` требует **≥1** goswitch-движок и НОЛЬ чужих (selfcheck.go:176-186, красный вердикт называет движки и фикс); корпус TestSelfcheck_InputSourceTwoEngines/SingleGreen/ForeignRejected/MixedRejected — PASS. Чорды GNOME не трогаются (суперсе́ссия 260930-nxd стоит): `clearSwitchBinding`/`ownerSourcesSet`/`clearedSwitchBindings` отсутствуют во всех *.go (absence-пин этого прогона). **WR-04 закрыт**: README переписан (v1.1.0, 6f2742e) — схема теперь верная `org.gnome.desktop.wm.keybindings` `switch-input-source` … «не трогаются» (README.md:65-66), абзац очищенной эры удалён; раздел «Источники ввода и индикатор режима» документирует single-source + SNI-трей (README.md:72-85). super-space-alive выведен из matrix-v4 вердиктом владельца 2026-10-04 (daemon владеет переключением, доказано e2e-flip-keystroke) |
| 6 | **SC-6** — e2e-матрица зелёная дважды на живом столе, включая кейсы переключения; акт переключения наблюдаем в журнале | ✓ VERIFIED (вердикты владельца на диске) | WINDOWS #12 разрешён веткой «семантика пересматривается»: mixed-семантика ревизована в посимвольную инверсию (спека-дельта D-55, 2026-10-06 — записанный владельцем follow-up из 06-UAT), 13 строк матриц re-pinned (f9ed497, matrix-v1..v4), поведение доказано юнит-корпусом и верифицировано фазой 6 (51/51, 06-VERIFICATION). Двойная зелень — вердикт владельца 2026-10-04 (06-UAT тест 1: PASS, «Матрица 32/35×2 стабильно» после G-6-1-фикса); формальный свежесессионный гейт D-48/D-49 принят владельцем по доказательствам (2026-10-07, координатор); чек-лист docs/ACCEPTANCE.md (D-49) — рабочая поверхность milestone-UAT-сессии координатора, не гэп фазы. Акт переключения наблюдаем: оракул пары mode→switch_engine TestFlipMarksPaired/TestModeFollowedOnly — PASS на текущем дереве; кейсы two-source-flip/external-flip-sync в реестре (test/e2e/main.go:269-271); mise-задачи на месте (mise.toml:137-169) |

**Score:** 6/6 truths verified (0 present-behavior-unverified; 0 FAILED; 0 overrides)

Обе «человеческие» истины прошлого прогона (SC-2 индикатор, SC-6 матрица) закрыты записанными вердиктами владельца — источник каждого вердикта указан в строке таблицы. Это апгрейд статуса, не пере-классификация по присутствию: машинные поверхности (артефакты, проводка, юнит-оракулы) дополнительно перепроверены на текущем дереве.

### Deferred Items

Нет. Ни один открытый пункт не покрывается поздней фазой (фазы 6–8 завершены, milestone закрывается). Пункт deferred-items.md (preflight selftest-тап как реальный флип) остаётся задокументированным advisory — безобидным для всех текущих кейсов, как записано в 05-05.

### Advisory (New Scope, Unevidenced)

Re-verification прогон выполнен; новых new-scope находок нет: debt-marker скан всех covered-файлов — ноль TBD/FIXME/XXX; стаб-паттерны — ноль. Три review-ворнинга 05-REVIEW переносятся из прошлого отчёта (carried, не new-scope) — см. Anti-Patterns.

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/install/install.go` | wrapSources (≥1 источник), отказы, resolveWrapInput, skip-if-equal, report | ✓ VERIFIED | все символы на месте (:818-894, :411, :942-966); a11y-расширение фазы 8 (snapshot/restore toolkit-accessibility) не тронуло ASVS-дисциплину (TestUninstall_FullRollback PASS) |
| `internal/activate/activate.go` | канонический ParseSourceTuples/SourceTuple | ✓ VERIFIED | единственное место разбора; install и selfcheck потребляют (:137, selfcheck.go:14) |
| `internal/install/selfcheck.go` | checkInputSource ≥1 goswitch-движок, ноль чужих | ✓ VERIFIED | :176-204; вердикт называет движки (260930-toa bfebd28) |
| `internal/session/actor.go` | flipTo/SetSwitcher/ToggleMode/SyncEngine/settleCorrectionFlip/switchTimeout | ✓ VERIFIED | 2946+ строк после фаз 6-8; все швы и пины порядка на месте; G-6-1 lock-free emitter-слот |
| `internal/engine/conn.go` | BindSwitcher/OnGlobalEngine/BindGlobalEngine, newSwitcher → SetGlobalEngine, диспетчер, switch_engine-запись | ✓ VERIFIED | переехал из engine/ (5427d7a); wire-контракты нетронуты (TestSwitcherCallsSetGlobalEngine PASS) |
| `internal/engine/engine.go`/`factory.go`/`types.go` | FocusIn→sync-вход; фабрика; NameEN/NameRU | ✓ VERIFIED | AttachEngine→actor lock-free (G-6-1); сборка+импорты internal/engine зелёные |
| `cmd/goswitchd/main.go` | wiring: BindSwitcher/OnGlobalEngine→SyncEngine/BindGlobalEngine→IfOwned; трей Toggle | ✓ VERIFIED | :554-569, :121, :166; расширения фаз 7-8 (звук, a11y) рядом, не поверх |
| `test/e2e/case_switch.go` (+test) | two-source-flip, external-flip-sync, flip-marks оракул | ✓ VERIFIED | TestFlipMarksPaired/TestModeFollowedOnly PASS на HEAD |
| `test/e2e/case_combo.go`/`case_resilience.go` | аддитивные switch_engine-ожидания; ibus-restart (precondition count-agnostic с 260930-toa) | ✓ VERIFIED | ec2883e оракул ≥1 источника |
| `test/e2e/matrix.go` | switch_engine-оракул | ✓ VERIFIED | 13 mixed-строк re-pinned (f9ed497); YAML-ожидания = юнит-доказанной семантике |
| `docs/adr/ADR-006-two-engine-revision.md` | Accepted + 2 amendment + таблица живых доказательств | ✓ VERIFIED | Status: Accepted (2026-09-30); Amendment 2026-09-30 (single-source model, дословное решение владельца); Amendment 2026-10-02 (G-6-1, lock-free фабрика) |
| `README.md` (+test/e2e/README.md) | single-source модель пользователя, SNI-индикатор | ✓ VERIFIED (WR-04 закрыт) | README.md:65-66 (wm.keybindings не трогаются), :72-85 (один источник goswitch + SNI-трей) |
| `mise.toml` | e2e-two-source-flip/e2e-external-flip-sync/e2e-switch-spike/e2e-install-cycle/e2e-matrix* | ✓ VERIFIED | :97-169 |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| cmd/goswitchd/main.go | internal/session/actor.go | BindSwitcher → SetSwitcher до engine.Run | ✓ WIRED | main.go:558-565; TestMain_WiringSyncAndReader корпус зелёный (фаза 6 re-verify) |
| cmd/goswitchd/main.go | internal/engine | internal/engine import, BindGlobalEngine→IfOwned | ✓ WIRED | main.go:29; сборка зелёная после переезда 5427d7a |
| actor.go flipTo | internal/engine/conn.go | flipTo → SetSwitcher-замыкание → SetGlobalEngine | ✓ WIRED | conn.go:240; TestSwitcherCallsSetGlobalEngine (wire-акт) + TestActor_FlipRoutesThroughSwitcher PASS |
| conn.go | actor.go | GlobalEngineChanged/FocusIn → OnGlobalEngine → SyncEngine | ✓ WIRED | main.go:554; TestActor_SyncEngineNeverSwitches PASS |
| install.go | activate.go | activate.ParseSourceTuples (единственный парсер) | ✓ WIRED | install.go:137, activate.go:114 |
| install.go | gsettings set sources | wrapSources-вывод, skip-if-equal | ✓ WIRED | install.go:913-917; TestInstall_Sequence/SequenceSingleSource PASS |
| install state-файл | wrapSources вход | resolveWrapInput → savedSources | ✓ WIRED | install.go:864-894; TestUninstall_FullRollback |
| test/e2e/main.go | case_switch.go | реестр pickCase (switch-spike, two-source-flip, external-flip-sync) | ✓ WIRED | main.go:269-271 |
| case_switch.go / matrix.go | mise.toml | задачи e2e-* | ✓ WIRED | mise.toml:97-169 |
| трей (main.go) | actor.flipTo | Toggle: actor.ToggleMode → тот же flipTo | ✓ WIRED | main.go:121/:166; TestActor_ToggleModeFlips PASS (261001-fg3) |
| actor | SNI-индикатор | UpdateModeSymbol/ModeChanged/menuSync.SetMode на каждом флипе | ✓ WIRED | actor.go:1973/:1980/:1987; TestActor_FlipEmitsPanelSymbol PASS |
| matrix.go | cases/matrix-v*.yaml | switch_engine-оракул + посимвольная инверсия ожиданий | ✓ WIRED | f9ed497 re-pin; CASES-FROZEN дисциплина соблюдена спека-дельтой |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| StatusSnapshot.Engine | engineNameOf(a.mode) | внутренний режим демона (single writer) | да | ✓ FLOWING |
| switch_engine journal record | engineNameOf(target) | замыкание newSwitcher → wire-ответ ibus | да (живой RTT 6.5–9.2 мс после G-6-1) | ✓ FLOWING |
| IfOwned reactivation | globalEngine reader | GetGlobalEngine desc-вариант (conn.go:369-390) | да | ✓ FLOWING |
| install takeover value | wrapSources(resolveWrapInput) | живые sources gsettings / state-файл | да | ✓ FLOWING |
| трей-меню/иконка | modeSymbol | a.mode после флипа | да — реальное состояние, не хардкод | ✓ FLOWING |

Статических подстановок, хардкодов данных или HOLLOW-звеньев не найдено.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Сборка всех пакетов на HEAD 2171fae | `go build ./...` | exit 0 | ✓ PASS |
| Флип через шов, порядок, SET-семантика, sync, дедлайн, трей-жест, панель-символ | `go test -race ./internal/session/ -run 'TestActor_FlipRoutesThroughSwitcher\|FlipOrderPinned\|CorrectionFlipSetsResultEngine\|SwitcherDeadlineBounded\|SyncEngineNeverSwitches\|ToggleModeFlips\|FlipEmitsPanelSymbol'` | 7/7 PASS | ✓ PASS |
| Обёртка/отказы/откат/selfcheck (single-source модель) | `go test -race ./internal/install/ -run 'TestWrapSources\|TestUninstall_FullRollback\|TestSelfcheck_InputSource'` | 10/10 PASS | ✓ PASS |
| Wire-акт SetGlobalEngine через fake bus | `go test -race ./internal/engine/ -run TestSwitcherCallsSetGlobalEngine` | PASS | ✓ PASS |
| Оракул пары марок / sync-follow в e2e-корпусе | `go test -race ./test/e2e/ -run 'TestFlipMarksPaired\|TestModeFollowedOnly'` | 2/2 PASS | ✓ PASS |

Полный сюит НЕ прогонялся (правило этого прогона); зелёная итерация на HEAD corroborated фазовыми верификациями 6–8 (06: 51/51; 08: финальная верификация + UAT 6/6) и сборкой этого прогона.

### Probe Execution

Нет `scripts/*/tests/probe-*.sh` (конвенция проекта — mise-задачи живых кейсов, неизменна). Живые сценарии фазы покрыты записанными прогонами и владелец-вердиктами на диске (06-UAT, 08-UAT); новых probe-обязательностей у фазы 5 не возникло.

### Requirements Coverage

| Requirement | Source Plan(s) | Description | Status | Evidence |
| ----------- | ------------- | ----------- | ------ | -------- |
| SWCH-01 | 05-03 | Right Shift переключает раскладку | ✓ SATISFIED | flipTo-воронка + TestActor_FlipRoutesThroughSwitcher; живой two-source-flip (05-05) |
| SWCH-02 | 05-03 | Комбо «исправить и переключить» | ✓ SATISFIED | settleCombo через flipTo; TestActor_FlipOrderPinned |
| SWCH-03 | 05-01, 05-05 | Родное GNOME-переключение и индикатор актуальны | ✓ SATISFIED | SNI-индикатор следует режиму (владелец 2026-10-07); трей-жест в flipTo; чорды GNOME не тронуты |
| SWCH-04 | 05-03 | Тайминги single/double/triple разрешены ADR | ✓ SATISFIED | FSM-корпус нетронут; TestActor_FlipOnSingle корпус зелёный |
| INTEG-01 | 05-03 | goswitchd как IBus engine, commit_text | ✓ SATISFIED | wire-контракты нетронуты переездом internal/ (TestSwitcherCallsSetGlobalEngine) |
| INTEG-02 | 05-05 | Совместимость с keyd/xremap: транзит набора | ✓ SATISFIED | шов — исходящий D-Bus call; G-6-1 устранил глотание буквы (24/24) |
| INTEG-03 | 05-02 | Дефолтный IM-стек без root | ✓ SATISFIED | gsettings-only; install-корпус зелёный |
| INTEG-04 | 05-04 | Перерегистрация после рестарта ibus-daemon | ✓ SATISFIED | BindGlobalEngine per-generation (conn.go:66-74); IfOwned; живой ibus-restart (05-05) |
| INTEG-05 | 05-03 | Паника обработчика не роняет движок | ✓ SATISFIED | switcher-контур вне D-Bus-обработчиков; WARN-не-фатально пин |
| INST-01 | 05-02 | Установка без root: unit, регистрация, поставка | ✓ SATISFIED | SC-1; принят владельцем (install-cycle по доказательствам, 2026-10-07) |

**Orphaned requirements:** нет — все 10 ID ROADMAP фазы 5 заявлены фронтматтером планов (пере-проверено этим прогоном).

### Decision Coverage

All trackable CONTEXT.md decisions are honored by shipped artifacts. (check.decision-coverage-verify: total 3, honored 3, not_honored 0.)

### Test Quality Audit

Disabled/skipped тесты в requirement-linked корпусах: ноль (grep skip-паттернов по internal/, test/e2e — пусто). Циклических оракулов нет: ожидания матриц фиксируются спека-дельтой (D-55) и юнит-корпусом с независимым golden-файлом; flip-оракул читает журнал демона, не генерирует ожидания из системы под тестом. Сила утверждений — value/behavioral (пары марок в порядке, readback движка, вердикты отказов).

### Anti-Patterns Found

Debt-marker scan всех covered-файлов на HEAD 2171fae: **ноль** TBD/FIXME/XXX/PLACEHOLDER; стабов нет. Статус переносённых review-ворнингов 05-REVIEW на текущем дереве:

| File | Finding | Severity | Status on HEAD |
| ---- | ------- | -------- | -------------- |
| internal/session/actor.go flipTo | WR-01: мьютекс через SetGlobalEngine round trip | Warning | **ЗАКРЫТ** G-6-1 (ADR-006 Amendment 2026-10-02): фабрика освобождена от мьютекса (lock-free `atomic.Pointer[emitterSlot]`, actor.go:832-834); ауэт SetGlobalEngine под мьютексом с дедлайном 150 мс остаётся осознанным клин-гардом (пин 05-03 стоит) |
| internal/engine/conn.go reader | WR-02: GetGlobalEngine реактивации без дедлайна | Warning | ОТКРЫТ (reader использует ctx звонящего, WithTimeout в conn.go нет) — перенос из прошлого отчёта, не на keystroke-траектории, сам-хилится реконнектом; не инвалидирует must-have |
| internal/install/install.go:411 | WR-03: install активирует goswitch-en безусловно | Warning | ОТКРЫТ (`activateEngine(ctx, engineEN)` на месте) — перенос; при single-source модели менее остр (владелец сам выбрал активную запись); кандидат в fix-очередь |
| test/e2e/case_switch.go:253/1045/1140 | WR-05: unit-restore defer регистрируется ПОСЛЕ systemctl stop | Warning | ОТКРЫТ (проверено на текущем дереве) — перенос; отказ stop может оставить юнит-демон остановленным; e2e-only поверхность |
| README.md | WR-04: абзац очищенной эры с неверной схемой | Warning | **ЗАКРЫТ** — README v1.1.0 (6f2742e): верная схема wm.keybindings, «не трогаются» (README.md:65-66) |

Все три открытых ворнинга — carried-forward наблюдения прошлого отчёта (в его таблице антипаттернов), ни один не является new-scope находкой этого прогона и ни один не инвалидирует must-have как написано (вердикт прошлого прогона подтверждён).

### Human Verification Required

Нет (для этого прогона). Все 6 human-гейтов прошлого отчёта закрыты записанными вердиктами:

1. Индикаторная UAT (WINDOWS #11) — **владелец подтвердил 2026-10-07** ежедневным использованием (координатор milestone-UAT).
2. WINDOWS #12 (два красных ряда) — **разрешено владельцем**: mixed-семантика ревизована в посимвольную инверсию (спека-дельта D-55, 2026-10-06; записанный follow-up 06-UAT); строки re-pinned, доказаны юнит-корпусом, верифицированы фазой 6 (51/51).
3. Полный install-cycle (WINDOWS #9) — **принят владельцем по доказательствам** (2026-10-07); каждая часть доказана живьём в 05-02/05-05.
4. Формальный D-48/D-49 двойной свежесессионный прогон — **принят владельцем по доказательствам** (2026-10-07); чек-лист docs/ACCEPTANCE.md — рабочая поверхность milestone-UAT сессии координатора.
5. Трактовка полуобёрнутого отказа — **суперсе́днута решением владельца** (single-source model, ADR-006 Amendment: отказ errMixedSources на подмес — владельческий контракт, корпус зелёный).
6. Judgment-tier prohibitions — перепинены на текущем дереве: (b) приватность журнала (TestFlipMarksPaired PASS), (c) SyncEngine не воюет (TestActor_SyncEngineNeverSwitches PASS), (d) отказной корпус install (PASS), (e) чужой источник честно отклонён (TestSelfcheck_InputSourceForeignRejected/TestWrapSourcesForeignResidueRefused PASS); (a) WR-05 остаётся открытым ворнингом e2e-поверхности; «тихий журнал после 150 мс» — закрыт сильнее исходного вопроса: G-6-1 убрал сам узел (RTT 6.5–9.2 мс, 24/24 немедленных буквы, WARN более не норма на каждом флипе).

Артефакты всех пунктов присутствуют и wired — по правилам этого прогона human-пункты не возобновляются.

### Gaps Summary

Гэпов нет. Все 6 must-have держатся на текущем дереве (HEAD 2171fae): код присутствует и wired по пере-выведенным путям (internal/engine/, internal/install/, internal/session/), поведенческие истины переподтверждены 20 именованными -race прогонами этого верификатора, две прежде-человеческие истины закрыты записанными вердиктами владельца (на диске: 06-UAT/08-VERIFICATION; переданы координатором: 2026-10-07). Три ревизии после фазы 5 (single-source model, G-6-1, per-char inversion + internalize) — санкционированные владельцем эволюции контракта, задокументированные в ADR-006/SPEC, ни одна не сломала носитель фазы; две из четырёх переносившихся review-строк (WR-01, WR-04) закрыты этими же ревизиями. 0 FAILED, 0 STUB, 0 ORPHANED, 0 NOT_WIRED, 0 behavior-unverified, 0 overrides. Цель фазы достигнута.

Advisory: deferred-items.md (preflight selftest-тап) — безобидный, задокументированный; открытые WR-02/WR-03/WR-05 — fix-очередь владельца, на milestone-гейт не влияют.

---

_Verified: 2026-10-07T20:52:42Z_
_Verifier: Claude (gsd-verifier)_
