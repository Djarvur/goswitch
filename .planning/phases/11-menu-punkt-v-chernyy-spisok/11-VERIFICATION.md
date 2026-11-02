---
phase: 11-menu-punkt-v-chernyy-spisok
verified: 2026-11-02T19:56:09Z
status: passed
score: 15/15 must-haves verified
covered_files: [".planning/REQUIREMENTS.md", ".planning/debug/orca-terminal-second-shift.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-01-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-01-SUMMARY.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-02-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-02-SUMMARY.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-03-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-03-SUMMARY.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-04-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-04-SUMMARY.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-05-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-05-SUMMARY.md", "README.en.md", "README.md", "cmd/goswitchd/main.go", "cmd/goswitchd/main_test.go", "docs/CONFIG.md", "docs/SPEC.md", "docs/adr/adr-008-spike-log.md", "internal/appid/appid.go", "internal/appid/appid_test.go", "internal/appid/export_test.go", "internal/config/config.go", "internal/config/config_test.go", "internal/config/load.go", "internal/config/writer.go", "internal/config/writer_test.go", "internal/indicator/indicator.go", "internal/indicator/menu.go", "internal/indicator/menu_test.go", "internal/session/actor.go", "internal/session/actor_test.go", "mise.toml", "skills/goswitch-config/SKILL.md", "test/e2e/case_menu.go", "test/e2e/case_menu_blocklist.go", "test/e2e/case_menu_blocklist_test.go", "test/e2e/main.go"]
covered_digest: "v4:sha256:ef77990feb185e6897f20eb12bfc5c45db44aa78ed183be43b83edf99dc651da"
behavior_unverified: 0
overrides_applied: 0
behavior_unverified_items: []
re_verification:
  previous_status: passed
  previous_score: 15/15
  stale_reason: "Прошлый отчёт (verified 2026-11-01T20:55:00Z, passed 15/15) устарел ТОЛЬКО сменой covered-файлов после прогона: PR #27 (merge 6086cdb, коммит 5a259ad «fix(session): verifyDead class rides app memory», Orca first-gesture fix) изменил internal/session/actor.go и internal/session/actor_test.go. Пробелов в прошлом отчёте не было; это stale-refresh по текущему дереву, включающему PR #27"
  gaps_closed: []
  gaps_remaining: []
  regressions: []
human_verification:
  - test: "By-ear приёмка владельца пункта «Добавить ИмяПриложения в чёрный список» в реальном трее (UAT end-of-phase, прецедент G-10-1): пункт виден в группе автокоррекции под тумблером; с точным именем наблюдаемого приложения; клик мышью добавляет приложение в блок-лист; галочка на уже заблокированном; пункт гаснет (без имени и пояснений) при уходе в ненаблюдаемое приложение; super+space строка читается «Super+Space — язык»"
    expected: "Пункт живёт от свежей личности, клик работает с одного нажатия, состояния D-01/D-03 визуально корректны; автокоррекция в заблокированном приложении замолкает"
    why_human: "VALIDATION.md Manual-Only: визуальный рендер, эргономика клика и by-ear зачёт — машинные оракулы кейса (файл/Load/тишина/no-op) их не заменяют; живой цикл машиной уже доказан на ВМ (adr-008, PASS 00:27 UTC)"
    closed: "2026-11-01: владелец делегировал живую приёмку стенду («Тестируй сам в виртуалке»); закрыто живым дампом меню и кликом на сборке ветки фазы 11 (v1.2.0-alpha.2-76-g903c846): пункт id 16 под «Автокоррекцией» с точным именем org.gnome.Zenity, клик Event(16) → apps_blocklist: [org\\.gnome\\.Zenity] (QuoteMeta), toggle-state=1; фокус wezterm → enabled=false, лейбл без имени; id 10 = «super+space — язык» — UAT Тест 1 pass"
  - test: "Round 2 пассивная journal-верификация на хост-установке (владелец next desktop session): оракул сформулирован в .planning/debug/orca-terminal-second-shift.md §Round 2 journal verify"
    expected: "Следующая коррекция с ОДНОГО нажатия после «коррекция → флип → клиентский Reset»; INFO-ряд lifecycle aftermath=true; ноль регресс-маркеров verify_dead=false"
    why_human: "Требует живой сессии владельца (Wayland-нативное окно Orca); инъекция в живую сессию запрещена (T-09-02-03); принадлежит приёмке бага Orca, маршрутизирована в UAT end-of-phase вместе с by-ear пунктом"
    closed: "2026-11-01: владелец делегировал стенду («Тестируй сам в виртуалке»); закрыто живым ВМ-рядом на сборке фикса g2c00c10 — UAT Тест 2 pass, см. строку 15 и debug-док §Round 2 VM-попытка"
coincidental_reliance_items:
  - truth: "MaxAutocorrectBlocklist — единый источник потолка (writer, actor, validate)"
    reason: undeclared-precondition
    harden: "Метка labelBlocklistFull «Чёрный список заполнен — 64/64» (internal/indicator/menu.go:94) дублирует потолок литералом зеркально, без импорта config (IN-01 ревью): дрейф потолка в config оставит метку ложной. Свести метку из константы (fmt.Sprintf при инициализации) или добавить drift-пин «64» в TestMaxAutocorrectBlocklistExport/layout-тест"
---

# Phase 11: MENU — пункт «в чёрный список» Verification Report

**Phase Goal:** Пользователь добавляет текущее приложение в чёрный список автокоррекции одним кликом в трей-меню — без ручного редактирования YAML
**Verified:** 2026-11-02T19:56:09Z
**Status:** passed (15/15 — stale-refresh после PR #27; обе human-приёмки закрыты 2026-11-01 и перенесены как закрытые)
**Re-verification:** Yes — stale-refresh: прошлый PASS (15/15, 2026-11-01) устарел сменой covered-файлов (actor.go/actor_test.go, PR #27 Orca first-gesture fix); пробелов в нём не было

## Goal Achievement

Фаза достигнута и остаётся достигнутой на текущем дереве (HEAD 05e5ed4, включает PR #27 / merge 6086cdb). Вся цепочка MENU-01 существует, проводна и доказана поведенчески: пункт id 16 (menu.go), fresh-машина (actor.go), writer (11-03), клик-композиция с гвардами (main.go), живой e2e на ВМ (adr-008, датированные PASS-ряды). Причина устаревания — PR #27 — довёл до actor.go/actor_test.go только app-memory расширение behavioral-класса (orca-first-gesture-dead, вердикт владельца 2026-11-02): ни один блок-лист-механизм фазы 11 не изменён (проверено полным диффом 5a259ad по actor.go и перепрогоном корпусов на текущем дереве). Обе human-приёмки закрыты 2026-11-01 живыми ВМ-рядами по делегации владельца («Тестируй сам в виртуалке») — перенесены как закрытые, не переоткрыты.

### Observable Truths

| #   | Truth   | Status     | Evidence       |
| --- | ------- | ---------- | -------------- |
| 1   | SC1: пункт «Добавить ИмяПриложения в чёрный список» (id 16) виден; клик пишет regexp.QuoteMeta в apps_blocklist (дедуп, потолок 64) и автокоррекция замолкает без рестарта | ✓ VERIFIED | `menu.go:76` menuIDBlocklistApp=16 (реестр :501, rootKids после ACToggle); композиция `blocklistToggle.add` (main.go:436-484): empty-path WARN → гварды known/covered/full → `regexp.QuoteMeta(app)` main.go:469 → `t.write` = config.AppendAutocorrectBlocklist (main.go:419) → `t.reload` = syncReload+FoldAppliedConfig (main.go:254); корпус `TestBlocklist*` (трассер+гварды+порядок) — ok 1.158s на этом прогоне; живой ряд adr-008:716 (menu-blocklist PASS 2026-11-01 00:27 UTC) |
| 2   | SC2: запись атомарна — golden-дифф «только новая строка», комментарии целы, pre-write валидация, atomic rename | ✓ VERIFIED | `writer.go:92` AppendAutocorrectBlocklist → rewriteAtomically (:121) + decodeValid pre-write гейт (:274, семантика Load); корпус `TestWriter_Append*\|TestMaxAutocorrectBlocklistExport` — ok 1.703s на этом прогоне |
| 3   | SC3 (в сужении D-03): неизвестная/несвежая личность — пункт неактивен со статической меткой без имени и пояснений; демон не угадывает, :1.NN не становится ни меткой, ни паттерном | ✓ VERIFIED | Двойной барьер: actor `blocklistStateLocked` (actor.go:3133-3145) отдаёт ПУСТОЕ имя при !fresh; menu `blocklistProps` (menu.go:578-585) — `!fresh \|\| app==""` → labelBlocklistApp disabled; bridgeApp отказывает namespace реестра (appid.go:211, registryPathPrefix); `TestBridgeApp_RefusesRegistryNamespace` + `TestBlocklistGuardMatrixSilent` GREEN |
| 4   | SC4: имя с метасимволами не ломает конфиг (строгий декодер принимает); отказ потолка 64 виден состоянием пункта в трее | ✓ VERIFIED | TestWriter_AppendMetacharPatterns GREEN (QuoteMeta-литерал переживает строгий Load); labelBlocklistFull «Чёрный список заполнен — 64/64» (menu.go:94) disabled; sentinel errAutocorrectBlocklistOverCeil (config.go:72) ДО записи, файл байт-нетронут. «Уведомление из трея» = состояние пункта — осознанное решение (десктоп-уведомления невидимы, UAT фазы 7; зафиксировано в CONTEXT и SPEC-ревизии) |
| 5   | D-04: macroChordWhat « — язык»; метка «super+space — язык»; README синхронны | ✓ VERIFIED | `menu.go:102` macroChordWhat = " — язык"; grep « — язык»: README.md:100,102 и README.en.md:97,99; пин в TestMenuLayout GREEN |
| 6   | D-01: covered — галочка (toggle-state 1); клик по отмеченному — no-op | ✓ VERIFIED | TestMenuBlocklistStateMatrixMarks (menu_test.go:325) GREEN; TestBlocklistGuardMatrixSilent/covered (main_test.go:1054) + TestWriter_AppendDedupNoOp (writer_test.go:783) GREEN; живой ряд: повторный клик — документ БАЙТ-ИДЕНТИЧЕН (adr-008:716) |
| 7   | D-02: props пункта считаются AT REQUEST TIME из снапшота; живых a11y-запросов при открытии меню нет | ✓ VERIFIED | `blocklistProps` (menu.go:578) читает только поля снапшота blApp/blFresh/blCovered/blFull; TestMenuBlocklistItemGroupProps (menu_test.go:200) GREEN |
| 8   | Freshness-машина: gain ставит fresh=true и пушит; деарм — blur отслеживаемой поверхности (loss-репорт наблюдателя, PushFocusedApp("")); engine-шум/synthetic свежесть не трогают | ✓ VERIFIED (перепроверено на пост-PR27 дереве) | `PushFocusedApp` (actor.go:1624-1645): name=="" → blApp=""/blFresh=false+пуш, gain → fresh=true+пуш; `HandleLifecycle` (actor.go:1414-1422) — свежесть НАМЕРЕННО не тронута engine-churn'ом (живая находка 11-05, дизайн зачётен прошлой верификацией и UAT: фокус wezterm → enabled=false); поля blApp/blFresh под a.mu (actor.go:420-433); TestActor_BlocklistItemFreshPushes (:5168) / FreshLosses (:5210) — ok 6.305s на этом прогоне (корпус TestActor_Blocklist\|TestTerminalClass_VerifyDead) |
| 9   | Проводка OnConn: Callbacks.BlockCurrentApp, install-пуш начального состояния, EnsureAppid (безусловный старт наблюдателя) | ✓ VERIFIED (перепроверено на пост-PR27 дереве) | main.go:267 (`BlockCurrentApp: toggles.block.add`), :275 installMenuState (→ :329-334, читает BlocklistAppState), :284 actor.EnsureAppid(); indicator.go:107-111 поле Callbacks.BlockCurrentApp; menu.go:468-471 dispatchClick → cb.BlockCurrentApp с nil-гвардом; actor.go:1647-1661 EnsureAppid (безусловный once-per-actor, стартап-флаг), :3117 ensureAppid (ленивый), общее тело startAppidOnce (:3193); проводка хука actor.go:3170 `appid.WithGainHook(a.PushFocusedApp)`; TestActor_EnsureAppidUnconditional + TestObserver_GainHook/LossHook GREEN |
| 10  | Потолок 64 — единый источник: config.MaxAutocorrectBlocklist экспортирован; actor читает его; старые имена отсутствуют | ✓ VERIFIED (перепроверено на пост-PR27 дереве) | config.go:42 `const MaxAutocorrectBlocklist = 64`; actor.go:3134 `len(...) >= config.MaxAutocorrectBlocklist`; grep `maxAutocorrectBlocklist` по *.go — 0 совпадений; TestMaxAutocorrectBlocklistExport GREEN. (См. coincidental_reliance: метка «64/64» — зеркальный литерал, IN-01) |
| 11  | Гейты D-53 байт-неизменны (autoCorrectBoundary/acFocusedApp/терминальное вето не тронуты ФАЗОЙ) | ✓ VERIFIED (с пост-фазовой оговоркой) | Фаза гейты не тронута — сверено байт-уровнем против до-планового 34059d8 прошлой верификацией. Пост-фазовой работой PR #27 (владелец-одобренный фикс orca-first-gesture-dead, вне объёма фазы 11) classifyTerminal получил app-memory консульт (actor.go:3779 `rememberedVerifyDeadApp`), а autoCorrectBoundary/acFocusedApp НЕ тронуты (полный дифф 5a259ad: единственная изменённая строка с acFocusedApp — НОВЫЙ вызов чтения внутри terminalAppKey, не правка гейта); корпус гейтов `TestAutoCorrect\|TestAutoConfirm` — ok 1.487s на текущем дереве; семантические строки D-53 в SPEC целы (grep ×2) |
| 12  | SPEC §4.3 несёт датированную ревизию меню v2 (2026-10-31); CONFIG.md дополнен; семантические строки D-53 целы | ✓ VERIFIED | grep-гейты зелёные на этом прогоне: «2026-10-31» ×5, « — язык» ×1, «Неизвестная идентичность приложения НЕ запрет» ×2 в docs/SPEC.md; «пункт меню» ×2, «QuoteMeta» ×1 в docs/CONFIG.md |
| 13  | Round 2: клиентский Reset окна aftermath не стирает выученный класс, настоящий Reset вне окна стирает (fail-closed); PR #25 → main; сборка установлена | ✓ VERIFIED (перепроверено на пост-PR27 дереве) | `verifyDeadOnContextDeath` (actor.go:1990-2032): Reset при flipAftermathStop != nil → cancel+class survives, Reset вне окна → немедленная стирка, real FocusOut завершает aftermath; PR #27 дисциплину окна НЕ тронул (карве-аут кода идентичен, память едет РЯДОМ с классом); TestTerminalClass_VerifyDeadSurvivesClientResetAfterOwnFlip (:7962) + ClearedByResetOutsideAftermath (:8029) — ok на этом прогоне; `git merge-base --is-ancestor 2c00c10 HEAD` — да (PR #25 в main); хост-юнит active, ExecStart = releases/v1.2.0-alpha.2-83-gda4ced8 (сборка PR #27-приёмки, несёт и Round 2-фикс; двинулась вперёд с g2c00c10 легитимно — пост-фазовой фиксой владельца); строки «Round 2 install»/«Round 2 journal verify» в debug-доке на месте |
| 14  | e2e-кейс menu-blocklist зарегистрирован, mise-задача на месте; живой цикл + re-pin menu-v2 — зелёные на ВМ | ✓ VERIFIED | test/e2e/case_menu_blocklist.go (454 строки; пин id 16, Event-канал, оракулы файла/Load/тишины/no-op, teardown) + case_menu_blocklist_test.go (74); «menu-blocklist» в реестре main.go:422; [tasks.e2e-menu-blocklist] mise.toml:279; датированные ряды adr-008:716-718 (menu-blocklist PASS 00:27 UTC; menu-v2 re-pin PASS 00:30 UTC, id 1/2/4/5 не сдвинуты) |
| 15  | «Round 2 journal verify»: живой ряд фиксирует следующую коррекцию с одного нажатия | ✓ VERIFIED (ВМ-стенд, 2026-11-01 20:26–20:29 UTC — закрыто, перенесено) | Закрыто 2026-11-01 по делегации владельца («Тестируй сам в виртуалке»): живой ряд на сборке фикса g2c00c10 — один штатный verify-раунд обучения, далее 4 подряд коррекции done changed=true с ОДНОГО double-shift (включая 20:27:45 — воспроизведение исходного бага со 2-го нажатия), каждая с собственным флипом и ре-минтом; verify_dead=false — 0 строк за сессию; оговорка честности (эхо-Reset не пришёл — глотать было нечего, карве-аут pinned юнит-парой) — в UAT Тест 2 (pass). Хвост: PR #27 (2026-11-02) закрывает следующую итерацию того же бага (first-gesture после churn) — отдельная сессия, отдельная приёмка, фазу 11 не ретроспективно не инвалидирует |

**Score:** 15/15 truths verified (0 present-behavior-unverified)

**Верификатором независимо перепрогнано на текущем дереве (HEAD 05e5ed4):** go build ./... OK; go vet ./... OK; корпусы -race — internal/session `TestActor_Blocklist\|TestTerminalClass_VerifyDead` ok 6.305s, `TestAutoCorrect\|TestAutoConfirm` (гейты D-53) ok 1.487s, `TestFlapWindowContextIdentity\|SameAppFreshContextFirstPress` (PR #27) ok 1.473s; internal/indicator `TestMenuLayout\|TestMenuBlocklist` ok 1.050s; internal/config `TestWriter_Append\|TestMaxAutocorrectBlocklistExport` ok 1.703s; cmd/goswitchd `TestBlocklist` ok 1.158s; internal/appid `TestObserver_GainHook\|TestObserver_LossHook\|TestBridgeApp_RefusesRegistryNamespace` ok 5.095s; grep-гейты доков/README/mise/e2e-реестра — все зелёные; debt-marker скан изменённых файлов (actor.go, actor_test.go) — чист.

### Deferred Items

Не обнаружены (как и в прошлом отчёте): хвост «удаление из блок-листа через меню» — deferred-идея CONTEXT вне объёма MENU-01; фазы 12-13 его не покрывают и не обязаны.

### Advisory (New Scope, Unevidenced)

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| — | None | — | новых unevidenced находок Step 7 на этом прогоне нет; PR #27 — задокументированная, протестированная, владельцем-принятая пост-фазовая работа (см. truth 11/13), не находка |

Примечания по известным находкам ревью (11-REVIEW.md WR-01..WR-04, IN-01, IN-02 — документированные advisories, перенесены из прошлого отчёта): ни одна не инвалидирует must-have. WR-03 частично разрешён PR #27 (фальсифицированная посылка «different object inside the window = real switch» снята — reset keys by app identity, TestFlapWindowContextIdentity_DifferentAppWipesClass GREEN).

Flagged assumption из SUMMARY (spec-less probe fallback: `{"requirement_id":"MENU-01","category":"unclassified","status":"unresolved"}`) — учтена (перенесено из прошлого отчёта): край «неизвестная личность» фактически покрыт 7-строчной матрицей состояний, fresh-машиной, гвард-матрицей и guard'ом реестра — все GREEN на текущем дереве; строка остаётся формально unresolved как артефакт классификации проба, не как пробел реализации.

### Required Artifacts

| Artifact | Expected    | Status | Details |
| -------- | ----------- | ------ | ------- |
| `internal/indicator/menu.go` | id 16, матрица состояний, сеттер, dispatchClick, D-04 | ✓ VERIFIED | :76/:87/:94/:102/:310/:416/:468/:501/:524/:578 — все элементы на месте (пост-PR27 дереве) |
| `internal/indicator/indicator.go` | Callbacks.BlockCurrentApp | ✓ VERIFIED | :107-111 |
| `internal/indicator/menu_test.go` | layout/матрица/дельты/клик пины | ✓ VERIFIED | TestMenuBlocklistItemGroupProps :200, TestMenuBlocklistStateMatrixMarks :325 — корпус GREEN |
| `internal/appid/appid.go` | WithGainHook вне o.mu; guard реестра | ✓ VERIFIED | :116-123/:205-211/:340/:367/:385; хук-корпус GREEN |
| `internal/session/actor.go` | PushFocusedApp/fresh, BlocklistAppState, EnsureAppid, fold-пуш, Round 2 aftermath | ✓ VERIFIED (фокус stale-refresh) | :420-433 (blApp/blFresh под a.mu), :1414-1422 (свежесть не в HandleLifecycle), :1606/:2467-2477 (install/fold пуши), :1624-1645/:1647-1661/:1667-1672, :3108-3159, :3170, :1990-2032 (aftermath карве-аут не тронут PR #27) |
| `internal/session/actor_test.go` | freshness/fold/accessor/ensure + Round 2 пины | ✓ VERIFIED (фокус stale-refresh) | :5168/:5210/:5452/:7962/:8029 — все фазовые тест-функции живы на пост-PR27 файле и GREEN |
| `internal/config/writer.go` | AppendAutocorrectBlocklist | ✓ VERIFIED | :92-121; pre-write decodeValid :274; корпус GREEN |
| `internal/config/load.go` | decodeValid (единая Load-семантика) | ✓ VERIFIED | используется pre-write гейтом writer.go:274 |
| `internal/config/config.go` | MaxAutocorrectBlocklist=64 + sentinel | ✓ VERIFIED | :42/:72/:479-482; 0 использований старого имени |
| `cmd/goswitchd/main.go` | click-композиция + проводка OnConn | ✓ VERIFIED | :254/:267/:275/:284/:329-334/:418-419/:436-484 |
| `cmd/goswitchd/main_test.go` | трассер + гвард-матрица | ✓ VERIFIED | TestBlocklistClickTracer :827, TestBlocklistGuardMatrixSilent :1054 — GREEN |
| `test/e2e/case_menu_blocklist.go` | живой кейс с оракулами | ✓ VERIFIED | 454 строки + тест-файл 74 строки |
| `test/e2e/main.go` + `mise.toml` | регистрация + задача | ✓ VERIFIED | main.go:422, mise.toml:279 |
| `docs/SPEC.md` / `docs/CONFIG.md` | датированная ревизия + контракт | ✓ VERIFIED | все grep-гейты зелёные |
| `README.md` / `README.en.md` | D-04 строки | ✓ VERIFIED | :100/:102 и :97/:99 « — язык» |
| `.planning/debug/orca-terminal-second-shift.md` | ряды Round 2 | ✓ VERIFIED | «Round 2 install» + «Round 2 journal verify» на месте |

### Key Link Verification

Все 13 линков планов перепроверены вручную на текущем дереве (инструмент verify.key-links не парсит аннотированные from/to — как и в прошлом прогоне):

| From | To  | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| menu.go dispatchClick menuIDBlocklistApp | main.go композиция | Callbacks.BlockCurrentApp | ✓ WIRED | menu.go:468-471 → main.go:267 → blocklistToggle.add :456 |
| main.go композиция | actor.BlocklistAppState + FoldAppliedConfig | аксессор + toggleReload | ✓ WIRED | main.go:418 state:, :254 reload = syncReload+Fold |
| main.go композиция | config.AppendAutocorrectBlocklist | единственный persist path, QuoteMeta до вызова | ✓ WIRED | main.go:419, :469 |
| writer.go append | rewriteAtomically | temp same-dir + rename 0600 | ✓ WIRED | writer.go:92→121 |
| writer.go pre-write | load.go decodeValid | отказ до rename | ✓ WIRED | writer.go:274 |
| config.MaxAutocorrectBlocklist | actor full-состояние | единый источник | ✓ WIRED | actor.go:3134 |
| appid.go run() хук | actor.PushFocusedApp | WithGainHook вне o.mu | ✓ WIRED | actor.go:3170; appid.go:367/:385 (после снятия мьютекса) |
| Identity-layer деарм (loss/blur) | fresh=false + пуш | свежесть живёт в слое identity | ✓ WIRED | actor.go:1628-1641; TestActor_BlocklistItemFreshLosses GREEN (перепрочено: HandleLifecycle :1414-1422 свежесть намеренно не трогает — дизайн 11-05, зачтён прежней верификацией) |
| actor pushBlocklistAppLocked/fold | menu.SetBlocklistApp | MenuSync-шов, дельты | ✓ WIRED | actor.go:3153-3159 → menu.go:310-321 |
| SPEC-ревизия меню v2 | internal/indicator/menu.go | спека прежде кода | ✓ WIRED | ревизия SPEC ↔ реализация (id 16, D-01/D-03, D-04) |
| debug-док Round 2 гипотеза | actor.go verifyDeadOnContextDeath | aftermath-карве-аут | ✓ WIRED | :1990-2032; тест-пара GREEN на пост-PR27 дереве |
| e2e case Event id 16 | живой демон dbusmenu.Event | канал case_menu.go | ✓ WIRED | регистрация main.go:422; живой PASS-ряд adr-008:716 |
| e2e silence oracle | case_autocorrect.go blocklist-silent машина | app-blocked абстейнция | ✓ WIRED | переиспользован; fired=0, одна app-blocked в живом ряду |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| Метка пункта меню | m.blApp/blFresh | a11y-gain → WithGainHook → PushFocusedApp → SetBlocklistApp | да (живой наблюдатель; живой дамп UAT: org.gnome.Zenity) | ✓ FLOWING |
| covered/галочка | blCovered | fold applySnapshot реального конфиг-документа (re.MatchString) | да | ✓ FLOWING |
| Конфиг-документ | apps_blocklist | клик → QuoteMeta → AppendAutocorrectBlocklist → YAML-файл | да (живой ряд: [] → [org.gtk.application.gs_e2e_password]) | ✓ FLOWING |
| Гвард-тишина | ac_app_blocked | блок-гейт автокоррекции по документу | да (живой ряд: fired=0, одна app-blocked) | ✓ FLOWING |

### Behavioral Spot-Checks

Все команды исполнены верификатором на текущем дереве (HEAD 05e5ed4, пост-PR27) 2026-11-02:

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Fresh-машина + fold + аксессор + Round 2 цикл + fail-closed пины (фокус stale-refresh) | `go test ./internal/session/ -race -count=1 -run 'TestActor_Blocklist\|TestTerminalClass_VerifyDead'` | ok 6.305s | ✓ PASS |
| Гейты D-53 (не сломаны PR #27) | `go test ./internal/session/ -race -count=1 -run 'TestAutoCorrect\|TestAutoConfirm'` | ok 1.487s | ✓ PASS |
| PR #27 собственные контракты (не влияют на фазу 11) | `go test ./internal/session/ -race -count=1 -run 'TestFlapWindowContextIdentity\|SameAppFreshContextFirstPress'` | ok 1.473s | ✓ PASS |
| Layout/матрица/дельты/клик меню | `go test ./internal/indicator/ -race -count=1 -run 'TestMenuLayout\|TestMenuBlocklist'` | ok 1.050s | ✓ PASS |
| Writer: golden/дедуп/потолок/метасимволы | `go test ./internal/config/ -race -count=1 -run 'TestWriter_Append\|TestMaxAutocorrectBlocklistExport'` | ok 1.703s | ✓ PASS |
| Клик-композиция: трассер + гвард-матрица | `go test ./cmd/goswitchd/ -race -count=1 -run 'TestBlocklist'` | ok 1.158s | ✓ PASS |
| Gain/loss хуки + guard реестра | `go test ./internal/appid/ -race -count=1 -run 'TestObserver_GainHook\|TestObserver_LossHook\|TestBridgeApp_RefusesRegistryNamespace'` | ok 5.095s | ✓ PASS |
| Зелёная итерация (сборка) | `go build ./... && go vet ./...` | OK / OK | ✓ PASS |
| Ancestry фиксов в main | `git merge-base --is-ancestor 2c00c10 HEAD && git merge-base --is-ancestor 6086cdb HEAD` | оба в HEAD | ✓ PASS |
| Хост-юнит | `systemctl --user is-active goswitchd` + ExecStart | active; releases/v1.2.0-alpha.2-83-gda4ced8 (сборка PR #27-приёмки, несёт Round 2-фикс; см. truth 13) | ✓ PASS |
| Doc-гейты | grep SPEC/CONFIG/README/mise/e2e-реестр | все зелёные (значения в truths 5/12/14) | ✓ PASS |

### Probe Execution

Конвенциональных `scripts/*/tests/probe-*.sh` в проекте нет. Фазовый probe — живой e2e-прогон 11-05 (ssh-ритуал ВМ); перепрогон из этой среды невозможен (директива владельца de702f6: проверки продукта — только на ВМ goswitch-stand). Вердикты зафиксированы датированными рядами репозитория, произведёнными ровно теми командами, что пинит план (перенесено; ряды не изменились):

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| menu-blocklist (заезд 8) | `go run ./test/e2e -case menu-blocklist` (ВМ, HEAD 489aa5f) | adr-008:716 — 2026-11-01 00:27 UTC, CASE_EXIT=0, все оракулы зелёные | ✓ PASS (датированный ряд) |
| menu-v2 re-pin (заезд 2) | `go run ./test/e2e -case menu-v2` (ВМ, HEAD cf9c857) | adr-008:718 — 00:30 UTC, CASE_EXIT=0, id 1/2/4/5 не сдвинуты | ✓ PASS (датированный ряд) |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| MENU-01 | 11-01..11-05 (все пять) | Динамический пункт «Добавить ИмяПриложения в чёрный список»: QuoteMeta, без дублей, потолок 64, атомарная запись + живая перечитка; при неизвестной личности пункт неактивен, демон не угадывает | ✓ SATISFIED | Полная цепочка перепроверена на текущем дереве: пункт id 16 → клик-композиция с гвардами → writer → живой e2e ВМ (датированные ряды); закрытая UAT-приёмка живым дампом (id 16, QuoteMeta, галочка, wezterm-гашение). REQUIREMENTS.md:19 [x], таблица :53 «Complete» |

Orphaned requirements: нет — только MENU-01 маппится на Phase 11, объявлен всеми пятью планами.

### Decision Coverage

All trackable CONTEXT.md decisions are honored by shipped artifacts: 4/4 (D-01 галочка/no-op; D-02 кэш без живых a11y-запросов; D-03 неактивен без имени и пояснений; D-04 « — язык»). Discretion (позиция под тумблером) исполнена (menu.go:501, rootKids после ACToggle).

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| internal/indicator/menu_test.go (blocklist-корпус) | MENU-01 | да | 0 | нет | Value/Behavioral (матрица 7 строк, дельты, dispatch) | VALID |
| internal/session/actor_test.go (freshness/fold/accessor + Round 2; расширен PR #27 app-memory тестами) | MENU-01 | да | 0 | нет | Behavioral (полные циклы переходов) | VALID |
| internal/config/writer_test.go + config_test.go (append-корпус) | MENU-01 | да | 0 | нет | Value (golden-дифф, байтовые сравнения, потолок) | VALID |
| cmd/goswitchd/main_test.go (трассер + гвард-матрица) | MENU-01 | да | 0 | нет | Behavioral (write→reload порядок, счётчик writer) | VALID |
| internal/appid/appid_test.go (GainHook/LossHook/registry guard) | MENU-01 | да | 0 | нет | Behavioral | VALID |
| test/e2e/case_menu_blocklist_test.go (parse/swap пины) | MENU-01 | да | 0 | нет | Value | VALID |

Disabled tests on requirements: 0. Circular patterns: 0. Insufficient assertions: 0.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| — | — | TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER в изменённых (actor.go, actor_test.go) и фазовых файлах | — | не обнаружено (debt-marker gate чист; rc=1 на обоих grep) |
| — | — | placeholder/coming soon/not-yet-implemented | — | не обнаружено |
| — | — | пустые реализации / console-only | — | не обнаружено |
| go.mod/go.sum | — | новые модули | — | ноль (T-11-SC; PR #27 тоже без новых модулей — diff не трогает go.mod/go.sum) |

Evidence gate (#3304): изменённые с прошлого прогона файлы (actor.go, actor_test.go) проверены полным диффом + перепрогоном всех фазовых корпусов — детерминированных находок нет; unevidenced новых находок нет (advisory пуст).

### Human Verification Required

Обе приёмки закрыты 2026-11-01 живыми ВМ-рядами по делегации владельца («Тестируй сам в виртуалке», UAT 11-UAT.md: 2/2 pass) — перенесены в frontmatter как закрытые (с `closed:`-строками), не переоткрываются. Новых human-пунктов этот stale-refresh не породил (behavior_unverified: 0).

### Gaps Summary

Gaps нет. Stale-refresh подтверждает прошлый PASS на текущем дереве: единственное пост-фазовое изменение covered-файлов (PR #27, Orca first-gesture fix) не тронуло ни один механизм фазы 11 — блок-лист-состояние (blApp/blFresh под a.mu), fresh-машина (gain/loss через PushFocusedApp, engine-churn не трогает), EnsureAppid, BlocklistAppState, fold-пуш, проводка Callbacks.BlockCurrentApp, writer, потолок-единый-источник и Round 2 aftermath-карве-аут — всё на месте и зелёное (-race корпусы перепрогнаны верификатором 2026-11-02). PR #27 расширил classifyTerminal app-memory консультом (владельцем одобрено, спиннуто собственными тестами, корпус гейтов D-53 зелёный) — пост-фазовая эволюция, не регрессия фазы. Обе human-приёмки закрыты и перенесены. covered_digest пересчитан CLI (`verification.fingerprint`, вывод скопирован дословно): v4:sha256:ef77990feb185e6897f20eb12bfc5c45db44aa78ed183be43b83edf99dc651da — версия алгоритма поднялась v1→v4 вместе с gsd-tools, набор covered_files тот же 38-файловый.

---

_Verified: 2026-11-02T19:56:09Z_
_Verifier: Claude (gsd-verifier)_
