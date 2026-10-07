---
phase: 05-integratsiya-s-gnome-indikatsiya
verified: 2026-10-07T23:09:59Z
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
  - docs/SPEC.md
  - docs/adr/ADR-004-buffer-reset-triggers.md
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
  - test/e2e/matrix_test.go
covered_digest: "v1:sha256:c28a700d3c4cfe91d4fe116c4f960eb69593e1212bb872bf3b50041319aff2ca"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 6/6
  gaps_closed: []
  gaps_remaining: []
  regressions: []
---

# Phase 5: Интеграция с GNOME — Verification Report (focused staleness refresh, milestone close)

**Phase Goal:** goswitch переключает раскладку как полноценный гражданин GNOME: верхняя панель показывает индикатор активной раскладки (en/ru), флип (одиночный Shift, Super+Space, флип после коррекции) переключает источники через SetGlobalEngine на шине IBus — видимым для GNOME образом. Ревизия ADR-001 → ADR-006 фиксируется. (Исходная формулировка «ДВА goswitch-источника» ревизована владельцем 2026-09-30 вечером — ADR-006 Amendment: single-source model, ≥1 goswitch-источник; собственный индикатор — SNI goswitch в трее.)
**Verified:** 2026-10-07T23:09:59Z (HEAD 8292dd4, branch gsd/phase-07-menyu-v2-i-chernyy-spisok-avtokorrektsii — последний код-коммит дельты 4dc4877, docs 8292dd4)
**Status:** passed
**Re-verification:** Yes — fingerprint-staleness refresh №2 (quick task 261008-00m приземлился ПОСЛЕ верификации 2026-10-07T20:52:42Z и изменил код дерева, покрытый этой фазой; предыдущий прогон: passed 6/6 на HEAD 2171fae)

**Scope of this pass (focused).** Прошлый прогон (passed 6/6) устарел по отпечатку: quick 261008-00m (коммиты 77c91b7..4dc4877, спек-дельта 0781a5f) изменил ровно три шва, покрытые фазой. Дельта проверена полностью, остальное — точечная регрессия:

1. **G-2-3, internal/session/actor.go** (4a0c887; ADR-004 amendment 2026-10-07): `flipTo` после УСПЕШНОГО раунд-трипа SetGlobalEngine армит ровно один `flipCredit` (:2017; поле :263); `HandleLifecycle` FocusOut поглощает ровно один кредит и пропускает ТОЛЬКО жёсткий сброс буфера (:788-793) — синтетическая пара focus_out/focus_in собственного флипа — не «смена фокуса окна»; Reset кредит не потребляет никогда; WARN-неудача/nil-шов кредита не армит. Это ПРЯМО касается SC-3/SC-4 фазы: воронка flipTo, settleCorrectionFlip и D-36 порядок записей регресс-проверены (ниже).
2. **G-5-5, internal/install/**\* (2fa8b53; ADR-006 amendment 2026-10-07 + SPEC §4.3/§4.4 дельты, дословное решение владельца «автоматически приводить конфигурацию к правильной, а если не получилось - отказ»): SC-1 перевырен к ДЕЙСТВУЮЩЕМУ контракту — полуобёрнутый список теперь ДОЗАВЁРТЫВАЕТСЯ (`wrapSources` :837 — единственный арбитр завершения: renderWrapped + дедупликация повторных движковых туплей по первой позиции); отказ — ТОЛЬКО чужой остаток вне us/ru (`errMixedSources` :860 при наличии goswitch-материала), неподдерживаемый вид, нечего заворачивать (`errUnsupportedPair` :168) — все ДО любой мутации; `resolveWrapInput` :892 маршрутизирует mixed-live через тот же арбитр; `checkInputSource` (selfcheck.go :196) ЛЕЧИТ завершаемое полусостояние (один gsettings set :218 → перечитка :221-226 → зелёный вердикт), остаток остаётся красным с нулём записей. Гэп G-5-5 из 05-UAT.md закрыт этим quick-заданием (запись `status: resolved, resolved_by: 261008-00m` на диске).
3. **test/e2e/matrix.go** (4dc4877): герметичность стенда ТОЛЬКО — предикат `caseNeedsBaseEstablishment` пинит -config-документ каждому кейс-демону (кроме reload-кейсов, чей первый reload-шаг и есть установление); оракулы и ожидания не тронуты.

Живое доказательство на диске (записано оркестратором на fix-дереве, не пересоздаётся этим прогоном): матрица v1 **16/16** (слово через флип → «паиghbdtn» — case, падавший до фикса, зелёный) и v2 **20/21** (ОБА пере-пиненных mixed-ряда зелёные; единственный FAIL super-space-alive — детерминированно атрибутированный отчётом стендовый класс кражи фокуса, не продукт) — `.planning/quick/261008-00m-milestone-close-owner-fixes-batch-1-g-2-/261008-00m-VERIFICATION.md`, `.planning/STATE.md` (строка 261008-00m, Verified), корроборация в 03-VERIFICATION.

Человеко-принимаемые измерения остаются закрытыми вердиктами владельца прошлого прогона (на диске: 06-UAT, 08-VERIFICATION, 05-UAT; переданы координатором milestone-UAT-сессии 2026-10-07): артефакты всех пунктов присутствуют и wired, новых человеческих поверхностей дельта не создала — оба новых инварианта (кредит, дозаворот) имеют поведенческие тесты (ниже).

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | **SC-1 (контракт G-5-5, ревизован владельцем 2026-10-07)** — install оборачивает источники пользователя в goswitch-движки их раскладки; завершаемое полусостояние дозаворачивается к канонической goswitch-конфигурации; отказ — только чужой остаток вне us/ru / неподдерживаемый вид / нечего заворачивать — всегда ДО любой мутации; uninstall восстанавливает оригинал дословно | ✓ VERIFIED | Текущее дерево по ADR-006 amendment 2026-10-07 (дословная цитата владельца в audit-trail, снятый контракт задокументирован) и SPEC §4.3/§4.4 (:139-148): `wrapSources` (install.go:837) — ЕДИНСТВЕННЫЙ арбитр завершения: каждый xkb us/ru заворачивается, минимум один (TestWrapSourcesSingleSource), повторные движковые тупли дедуплицируются по первой позиции — полусостояние «goswitch-en + xkb us» завершается (новый TestWrapSourcesHalfWrappedCompletes PASS); отказы `errMixedSources` (:860, только при наличии goswitch-материала рядом с чужим остатком) / неподдерживаемый вид / нечего заворачивать — ДО первой записи (saveState-гейт :530-546, «ноль записей gsettings на отказе» — дисциплина 260930-toa сохранена); `resolveWrapInput` (:892) маршрутизирует mixed-live через wrap-попытку (один арбитр); selfcheck ЛЕЧИТ завершаемое полусостояние (selfcheck.go:212-234: один set → перечитка → верификация владения до зелёного вердикта; остаток красный с D-53-рационал). Именованные прогоны этого верификатора на HEAD 8292dd4 (-race): TestWrapSourcesPairPreserved/SingleSource/ForeignResidueRefused/HalfWrappedCompletes/RefusalTable/AlreadyOwnedUpgrade — 6/6 PASS; TestSelfcheck_InputSourceTwoEngines/SingleGreen/ForeignRejected/MixedHeals/ResidueStaysRed — 5/5 PASS; TestUninstall_FullRollback переносится (uninstall-путь дельтой не тронут — diff 2fa8b53 не меняет restore). Полный живой install-cycle принят владельцем по доказательствам (2026-10-07, вердикт координатора) |
| 2 | **SC-2 (спек-дельта)** — индикатор отражает активный источник при пользовательских жестах; видим нативным для GNOME образом | ✓ VERIFIED (закрыт вердиктом владельца) | Вердикт владельца 2026-10-07 (ежедневное использование; координатор milestone-UAT) переносится — дельта его не касается, машинная поверхность перепроверена: эмиссии панели на месте в flipTo — `UpdateModeSymbol` (:2025, :2087), `ModeChanged` (:2032, :2092), трей-меню `menuSync.SetMode` (:2039, :2097); кредитный путь G-2-3 пропускает ТОЛЬКО buf.HardReset — эмиссии режима в кредитной ветке не страдают (код :788-793 прочитан; TestActor_FlipEmitsPanelSymbol — PASS прошлым прогоном, его код-путь эмиссий дельтой не тронут). SNI-трей + DBusMenu + супервизор re-attach (main.go Toggle: actor.ToggleMode) — файл main.go дельтой не тронут |
| 3 | **SC-3** — все жесты через SetGlobalEngine на ibus-соединении демона (не subprocess), <50 мс; состояние демона не расходится с активным источником (sync при FocusIn; WARN при неудаче) | ✓ VERIFIED (регресс-проверено против дельты G-2-3) | Воронка едина и после дельты: `flipTo` (actor.go:1996) — все сайты вызова на месте (:1056, :1095 ToggleMode, :1118, :1204, :2613) + settleCorrectionFlip (:1234); кредит-арминг (:2017) АДДИТИВЕН — сам вызов `sw(ctx, engineNameOf(target))`, WARN-запись и порядок D-36 не тронуты (diff 4a0c887 прочитан: только else-ветка с `a.flipCredit++`); `SetSwitcher`-замыкание (:874) → `SetSwitcher:`-шов → `CallWithContext(ibusService+".SetGlobalEngine")` (conn.go:240, файл не тронут) с записью `switch_engine` + WARN при сбое; `switchTimeout = 150ms` (:67); wiring main.go (:554-566: OnGlobalEngine→SyncEngine, BindSwitcher→SetSwitcher, BindGlobalEngine) — main.go не тронут. Именованные прогоны этого верификатора (-race, HEAD 8292dd4): **TestSwitcherCallsSetGlobalEngine** (wire-акт через fake bus) PASS, **TestActor_SyncEngineNeverSwitches** PASS, TestActor_FlipRoutesThroughSwitcher/FlipOrderPinned/SwitcherDeadlineBounded/ToggleModeFlips — PASS прошлым прогоном (flipTo-тело вокруг арминга не менялось; D-36-пин :1950-1952 стоит дословно). Латентность: живой замер 05-05 ≈41 мс < 50 мс; G-6-1 lock-free фабрика не тронута |
| 4 | **SC-4** — режим после коррекции = скрипт результата (lat→cyr оставляет goswitch-ru, наоборот — goswitch-en); слово, набранное ЧЕРЕЗ собственный флип демона, корректируется целиком (G-2-3) | ✓ VERIFIED | `settleCorrectionFlip` (actor.go:1224) → `flipTo(scriptOf(converted))`, три сайта вызова (:2839, :2870, :2898) — diff 4a0c887 их НЕ трогает. Именованные прогоны (-race): **TestActor_CorrectionFlipSetsResultEngine** PASS (оба направления). Новый инвариант G-2-3 поведенчески закреплён: **TestActor_OwnFlipMixedWordCorrectsWhole** PASS (слово через синтетический FocusOut корректируется целиком → «паиghbdtn»-семантика, живое подтверждение — матрица v1 16/16 на диске), **TestActor_OwnFlipCreditConsumedOnce** PASS (ровно один кредит; ВТОРОЙ FocusOut — реальная потеря с полным сбросом), **TestActor_OwnFlipFailedSwitcherArmsNothing** PASS (WARN/nil-шов кредита не армит), **TestBuffer_ResetByFocusOut** PASS (реальная потеря фокуса сохраняет жёсткий сброс — прежний CORR-09 контракт не деградировал) |
| 5 | **SC-5** — GNOME-переключение по клавишам остаётся у goswitch; selfcheck отражает состояние источников | ✓ VERIFIED (модель ревизована владельцем; selfcheck-часть регресс-проверена против дельты G-5-5) | Selfcheck по ADR-006 Amendment + G-5-5: `checkInputSource` (selfcheck.go:196) требует ≥1 goswitch-движок и НОЛЬ чужих, ЛЕЧИТ завершаемое полусостояние (TestSelfcheck_InputSourceMixedHeals PASS), остаток — красный с D-53-рационалом и НУЛЁМ мутаций (TestSelfcheck_InputSourceResidueStaysRed PASS); зелёный вердикт называет движки (TestSelfcheck_InputSourceTwoEngines/SingleGreen PASS); чужой источник честно отклонён (TestSelfcheck_InputSourceForeignRejected PASS). Чорды GNOME не трогаются (absence-пин этого прогона: `clearSwitchBinding`/`ownerSourcesSet`/`clearedSwitchBindings` — 0 вхождений во всех *.go). README v1.1.0 (wm.keybindings, «не трогаются», README.md:65-66; single-source + SNI :72-85) — файл не тронут дельтой |
| 6 | **SC-6** — e2e-матрица зелёная дважды на живом столе, включая кейсы переключения; акт переключения наблюдаем в журнале | ✓ VERIFIED (живые вердикты на диске) | Дельта matrix.go — ТОЛЬКО герметичность (4dc4877: каждый кейс-демон пинит -config; найдено живьём 2026-10-08 — daemon без -config адоптировал живой конфиг владельца машины), оракулы/ожидания не тронуты. Живое доказательство на fix-дереве ЗАПИСАНО: v1 16/16 + v2 20/21, оба mixed-ряда зелёные, single FAIL = стендовый environmental-класс (261008-00m-VERIFICATION, STATE.md, 03-VERIFICATION). Акт переключения наблюдаем: оракул пары mode→switch_engine (TestFlipMarksPaired/TestModeFollowedOnly — PASS прошлым прогоном, их код-пути дельтой не тронуты); кейсы two-source-flip/external-flip-sync в реестре (test/e2e/main.go:269-271); mise-задачи на месте (mise.toml:137-169). Двойная зелень/чек-лист D-48/D-49 — вердикты владельца прошлого прогона (на диске), переносятся |

**Score:** 6/6 truths verified (0 present-behavior-unverified; 0 FAILED; 0 overrides)

### Deferred Items

Нет. Ни один открытый пункт не покрывается поздней фазой (фазы 6–8 завершены, milestone закрывается). G-5-5 из 05-UAT.md закрыт quick 261008-00m (запись resolved на диске).

### Advisory (New Scope, Unevidenced)

Re-verification прогон выполнен; новых new-scope находок нет: debt-marker скан всех шести дельта-файлов (actor.go, actor_test.go, install.go, install_internal_test.go, selfcheck.go, selfcheck_test.go, matrix.go, matrix_test.go) — ноль TBD/FIXME/XXX; стаб-паттерны — ноль. Три review-ворнинга 05-REVIEW переносятся из прошлых отчётов (carried, не new-scope; WR-03 перепроверен на текущем дереве — activateEngine(ctx, engineEN) install.go:411 на месте) — см. Anti-Patterns.

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/install/install.go` | wrapSources — единственный арбитр завершения (G-5-5): дозаворот полусостояния, отказы до мутации; resolveWrapInput; skip-if-equal; report | ✓ VERIFIED | :837-884 (renderWrapped + first-occurrence dedupe), отказы :168/:172/:860, resolveWrapInput :892, saveState-гейт :530-546; TestWrapSources* 6/6 PASS -race |
| `internal/install/selfcheck.go` | checkInputSource ≥1 goswitch-движок, ноль чужих; G-5-5 heal завершаемого полусостояния | ✓ VERIFIED | :196-236: wrapSources-арбитр → один gsettings set (:218) → перечитка → владение; TestSelfcheck_InputSource* 5/5 PASS -race |
| `internal/activate/activate.go` | канонический ParseSourceTuples/SourceTuple | ✓ VERIFIED | :114; файл дельтой не тронут; потребляется install (:540) и selfcheck (один парсер на обе стороны) |
| `internal/session/actor.go` | flipTo/SetSwitcher/ToggleMode/SyncEngine/settleCorrectionFlip/switchTimeout/flipCredit | ✓ VERIFIED | 3000+ строк после фаз 6-8 + G-2-3; flipCredit :263/:788/:2017; все швы и пины порядка (D-36 :1950-1952) на месте |
| `internal/engine/conn.go` | BindSwitcher/OnGlobalEngine/BindGlobalEngine, newSwitcher → SetGlobalEngine, диспетчер, switch_engine-запись | ✓ VERIFIED | файл дельтой не тронут; :240 SetGlobalEngine call; TestSwitcherCallsSetGlobalEngine PASS -race на текущем дереве |
| `internal/engine/engine.go`/`factory.go`/`types.go` | FocusIn→sync-вход; фабрика; NameEN/NameRU | ✓ VERIFIED | не тронуты дельтой; G-6-1 lock-free emitter-слот на месте |
| `cmd/goswitchd/main.go` | wiring: BindSwitcher/OnGlobalEngine→SyncEngine/BindGlobalEngine→IfOwned; трей Toggle | ✓ VERIFIED | :554-566, :121, :166; не тронут дельтой |
| `test/e2e/case_switch.go` (+test) | two-source-flip, external-flip-sync, flip-marks оракул | ✓ VERIFIED | не тронут дельтой; WR-05 статус неизменен (см. Anti-Patterns) |
| `test/e2e/case_combo.go`/`case_resilience.go` | аддитивные switch_engine-ожидания | ✓ VERIFIED | не тронуты дельтой |
| `test/e2e/matrix.go` (+test) | switch_engine-оракул; герметичность: -config пин каждому кейс-демону | ✓ VERIFIED | caseNeedsBaseEstablishment/caseHasReloadSteps (:866-882), TestMatrixCaseNeedsBaseEstablishment (261008-00m); живое v1 16/16 + v2 20/21 на диске |
| `docs/adr/ADR-006-two-engine-revision.md` | Accepted + 3 amendment + таблица живых доказательств | ✓ VERIFIED | Amendment 2026-10-07 G-5-5 (:280-320): дословное решение владельца, снятый контракт в audit-trail, «что меняется/что НЕ меняется» |
| `docs/adr/ADR-004-buffer-reset-triggers.md` | Amendment 2026-10-07 G-2-3 | ✓ VERIFIED | :81-112: механизм (синтетическая пара от re-mint), решение, различение кредитом — добавлено в covered_files этого прогона (теперь evidence-bearing для SC-3/SC-4) |
| `docs/SPEC.md` | §4.3/§4.4 дельты G-5-5 | ✓ VERIFIED | :139-148; добавлен в covered_files (теперь mapped-requirement для SC-1) |
| `README.md` (+test/e2e/README.md) | single-source модель пользователя, SNI-индикатор | ✓ VERIFIED | не тронут дельтой |
| `mise.toml` | e2e-* задачи | ✓ VERIFIED | :97-169; не тронут дельтой |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| cmd/goswitchd/main.go | internal/session/actor.go | BindSwitcher → SetSwitcher до engine.Run | ✓ WIRED | main.go:565; main.go не тронут дельтой |
| cmd/goswitchd/main.go | internal/engine | import, BindGlobalEngine→IfOwned | ✓ WIRED | main.go:554/:566 |
| actor.go flipTo | internal/engine/conn.go | flipTo → SetSwitcher-замыкание → SetGlobalEngine | ✓ WIRED | conn.go:240; TestSwitcherCallsSetGlobalEngine (wire-акт) PASS -race этим прогоном; кредит-арминг :2017 аддитивен к той же ветке |
| conn.go | actor.go | GlobalEngineChanged/FocusIn → OnGlobalEngine → SyncEngine | ✓ WIRED | main.go:554; TestActor_SyncEngineNeverSwitches PASS -race этим прогоном |
| install.go | activate.go | activate.ParseSourceTuples (единственный парсер) | ✓ WIRED | install.go:540, selfcheck.go:181; activate.go не тронут |
| install.go | gsettings set sources | wrapSources-вывод (завершённый список), skip-if-equal | ✓ WIRED | завершённый/deduped список — то, что пишется; отказы до записи |
| install state-файл | wrapSources вход | resolveWrapInput → savedSources | ✓ WIRED | install.go:892-910; already-owned путь не тронут (TestWrapSourcesAlreadyOwnedUpgrade PASS) |
| selfcheck.go | install.go | heal через ТОТ ЖЕ wrapSources-арбитр | ✓ WIRED (новый шов G-5-5) | selfcheck.go:212; TestSelfcheck_InputSourceMixedHeals PASS |
| test/e2e/main.go | case_switch.go | реестр pickCase | ✓ WIRED | main.go:269-271 |
| case_switch.go / matrix.go | mise.toml | задачи e2e-* | ✓ WIRED | mise.toml:97-169 |
| трей (main.go) | actor.flipTo | Toggle: actor.ToggleMode → тот же flipTo | ✓ WIRED | main.go:121/:166; flipTo-тело едино после дельты |
| actor | SNI-индикатор | UpdateModeSymbol/ModeChanged/menuSync.SetMode на каждом флипе | ✓ WIRED | actor.go:2025/:2032/:2039 (и :2087/:2092/:2097) |
| matrix.go | cases/matrix-v*.yaml | switch_engine-оракул + герметичный -config-пин | ✓ WIRED | оракул не тронут; establishment 4dc4877 |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| StatusSnapshot.Engine | engineNameOf(a.mode) | внутренний режим демона (single writer) | да | ✓ FLOWING |
| switch_engine journal record | engineNameOf(target) | замыкание newSwitcher → wire-ответ ibus | да (живой RTT 6.5–9.2 мс, G-6-1) | ✓ FLOWING |
| IfOwned reactivation | globalEngine reader | GetGlobalEngine desc-вариант (conn.go:369+) | да | ✓ FLOWING |
| install takeover value | wrapSources(resolveWrapInput) | живые sources gsettings / state-файл (завершённый список G-5-5) | да | ✓ FLOWING |
| selfcheck heal value | wrapSources(живой список) | живой gsettings get → set → re-read | да | ✓ FLOWING |
| трей-меню/иконка | modeSymbol | a.mode после флипа | да — реальное состояние, не хардкод | ✓ FLOWING |

Статических подстановок, хардкодов данных или HOLLOW-звеньев не найдено.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Сборка всех пакетов на HEAD 8292dd4 | `go build ./...` | exit 0 | ✓ PASS |
| Wire-акт SetGlobalEngine через fake bus | `go test -race ./internal/engine/ -run 'TestSwitcherCallsSetGlobalEngine$' -count=1` | PASS | ✓ PASS |
| Коррекция-флип = скрипт результата; sync не воюет; слово через собственный флип корректируется целиком | `go test -race -v ./internal/session/ -run 'TestActor_CorrectionFlipSetsResultEngine$\|TestActor_SyncEngineNeverSwitches$\|TestActor_OwnFlipMixedWordCorrectsWhole$' -count=1` | 3/3 PASS | ✓ PASS |
| Инварианты кредита G-2-3: ровно один кредит; неудача не армит; реальная потеря фокуса сохраняет сброс | `go test -race -v ./internal/session/ -run 'TestActor_OwnFlipCreditConsumedOnce$\|TestActor_OwnFlipFailedSwitcherArmsNothing$\|TestBuffer_ResetByFocusOut$' -count=1` | 3/3 PASS | ✓ PASS |
| Дозаворот/отказы/откат-корпус (G-5-5 контракт) + selfcheck-корпус (включая heal и residue-red) | `go test -race -v ./internal/install/ -run 'TestWrapSources\|TestSelfcheck_InputSource' -count=1` | 11/11 PASS | ✓ PASS |

Полный сюит НЕ прогонялся (правило прогона: только именованные тесты); итог 18 именованных -race прогонов этого верфикатора — все зелёные. Зелёная итерация corroborated: `mise run ci` зелёный на fix-дереве (261008-00m), живые матрицы v1 16/16 / v2 20/21 на диске.

### Probe Execution

Нет `scripts/*/tests/probe-*.sh` (конвенция проекта — mise-задачи живых кейсов, неизменна). Живые прогоны этого refresh-цикла (матрицы v1/v2 на fix-дереве) записаны оркестратором на диске (261008-00m-VERIFICATION, STATE.md) — по правилам прогона живое e2e не пересоздаётся; новых probe-обязательностей дельта не создала.

### Requirements Coverage

| Requirement | Source Plan(s) | Description | Status | Evidence |
| ----------- | ------------- | ----------- | ------ | -------- |
| SWCH-01 | 05-03 | Right Shift переключает раскладку | ✓ SATISFIED | flipTo-воронка едина после G-2-3 (кредит аддитивен); TestSwitcherCallsSetGlobalEngine + TestActor_OwnFlipMixedWordCorrectsWhole PASS |
| SWCH-02 | 05-03 | Комбо «исправить и переключить» | ✓ SATISFIED | settleCombo → settleCorrectionFlip → flipTo; сайты :2839/:2870/:2898 не тронуты |
| SWCH-03 | 05-01, 05-05 | Родное GNOME-переключение и индикатор актуальны | ✓ SATISFIED | SNI-индикатор следует режиму (владелец 2026-10-07, переносится); эмиссии :2025/:2032/:2039 на месте |
| SWCH-04 | 05-03 | Тайминги single/double/triple разрешены ADR | ✓ SATISFIED | FSM-корпус нетронут; switchTimeout :67 |
| INTEG-01 | 05-03 | goswitchd как IBus engine, commit_text | ✓ SATISFIED | wire-контракт перепроверен этим прогоном (TestSwitcherCallsSetGlobalEngine PASS -race на HEAD 8292dd4) |
| INTEG-02 | 05-05 | Совместимость с keyd/xremap: транзит набора | ✓ SATISFIED | шов — исходящий D-Bus call; G-6-1 не тронут |
| INTEG-03 | 05-02 | Дефолтный IM-стек без root | ✓ SATISFIED | gsettings-only; install/selfcheck корпусы зелёные (11/11 этим прогоном) |
| INTEG-04 | 05-04 | Перерегистрация после рестарта ibus-daemon | ✓ SATISFIED | BindGlobalEngine per-generation; conn.go не тронут; живой ibus-restart (05-05) |
| INTEG-05 | 05-03 | Паника обработчика не роняет движок | ✓ SATISFIED | switcher-контур вне D-Bus-обработчиков; WARN-не-фатально пин |
| INST-01 | 05-02 | Установка без root: unit, регистрация, поставка | ✓ SATISFIED | SC-1 по ДЕЙСТВУЮЩЕМУ G-5-5 контракту (дозаворот, отказы до мутации); принят владельцем (install-cycle по доказательствам, 2026-10-07) |

**Orphaned requirements:** нет — все 10 ID ROADMAP фазы 5 заявлены фронтматтером планов (перенесено из прошлого прогона; план-фронтматтеры дельтой не менялись).

### Decision Coverage

All trackable CONTEXT.md decisions are honored by shipped artifacts. (check.decision-coverage-verify этим прогоном: total 3, honored 3, not_honored 0.)

### Test Quality Audit

Disabled/skipped тесты в requirement-linked корпусах: ноль. Новые тесты дельты доказывают ровно заявленное: TestActor_OwnFlipCreditConsumedOnce — state-переход «ровно один кредит» (второй FocusOut — полный сброс, behavioral-assertions); TestActor_OwnFlipFailedSwitcherArmsNothing — негативная ветка арминга; TestWrapSourcesHalfWrappedCompletes / TestSelfcheck_InputSourceMixedHeals / TestSelfcheck_InputSourceResidueStaysRed — завершение/лечение/остаток по отдельности (value-level). Циклических оракулов нет: ожидания матриц фиксируются спека-дельтой (D-55) и юнит-корпусом; живые прогоны — внешний оракул рабочего стола.

### Anti-Patterns Found

Debt-marker скан шести дельта-файлов на HEAD 8292dd4: **ноль** TBD/FIXME/XXX/PLACEHOLDER; стабов нет. Статус переносённых review-ворнингов 05-REVIEW на текущем дереве (все — carried-forward, ни один не new-scope):

| File | Finding | Severity | Status on HEAD |
| ---- | ------- | -------- | -------------- |
| internal/session/actor.go flipTo | WR-01: мьютекс через SetGlobalEngine round trip | Warning | ЗАКРЫТ G-6-1 (lock-free фабрика, не тронута дельтой); ауэт с дедлайном 150 мс — осознанный клин-гард (пин 05-03 стоит) |
| internal/engine/conn.go reader | WR-02: GetGlobalEngine реактивации без дедлайна | Warning | ОТКРЫТ (conn.go не тронут дельтой) — перенос; не на keystroke-траектории |
| internal/install/install.go:411 | WR-03: install активирует goswitch-en безусловно | Warning | ОТКРЫТ (перепроверено этим прогоном на текущем дереве) — перенос; кандидат в fix-очередь |
| test/e2e/case_switch.go:253 | WR-05: unit-restore defer регистрируется ПОСЛЕ systemctl stop | Warning | ОТКРЫТ (перепроверено этим прогоном) — перенос; e2e-only поверхность |
| README.md | WR-04: абзац очищенной эры | Warning | ЗАКРЫТ (README v1.1.0, прошлый прогон) |

### Human Verification Required

Нет (для этого прогона). Все human-гейты прошлого отчёта закрыты записанными вердиктами владельца (на диске: 06-UAT, 08-VERIFICATION, 05-UAT — включая G-5-5 `resolved_by: 261008-00m`; переданы координатором milestone-UAT 2026-10-07) и переносятся: дельта не создала новой человеческой поверхности — оба новых поведенческих инварианта (кредит G-2-3, дозаворот G-5-5) закреплены зелёными именованными -race тестами этого прогона, а живое слово-через-флип подтверждено записанной матрицей v1 16/16 на fix-дереве.

### Gaps Summary

Гэпов нет. Оба шва дельты quick 261008-00m держат цели фазы на текущем дереве (HEAD 8292dd4): **G-2-3** — кредит-механизм в flipTo аддитивен (воронка, D-36-порядок, switch_engine-запись не тронуты), все четыре грани инварианта поведенчески закреплены (корректируется-целиком / ровно-один-кредит / неудача-не-армит / реальная-потеря-сбрасывается) — 6/6 именованных actor-прогонов зелёные; **G-5-5** — wrapSources стал единственным арбитром завершения по дословному решению владельца (ADR-006 amendment + SPEC §4.3/§4.4), отказы атомарны и без записи, selfcheck лечит с перечиткой — 11/11 install-прогонов зелёные; **matrix.go** — герметичность без изменения оракулов, живые v1 16/16 / v2 20/21 записаны на диске. SC-1 переверен к действующему контракту (не к устаревшей outright-refusal формулировке). 0 FAILED, 0 STUB, 0 ORPHANED, 0 NOT_WIRED, 0 behavior-unverified, 0 overrides, 0 debt-markers. Цель фазы достигнута.

Advisory: открытые WR-02/WR-03/WR-05 — fix-очередь владельца (carried, ни один не инвалидирует must-have); deferred-items.md (preflight selftest-тап) — задокументированный, безобидный.

---

_Verified: 2026-10-07T23:09:59Z_
_Verifier: Claude (gsd-verifier)_
