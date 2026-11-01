---
phase: 11-menu-punkt-v-chernyy-spisok
verified: 2026-11-01T20:55:00Z
status: passed
score: 15/15 must-haves verified
covered_files: [".planning/REQUIREMENTS.md", ".planning/debug/orca-terminal-second-shift.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-01-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-01-SUMMARY.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-02-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-02-SUMMARY.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-03-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-03-SUMMARY.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-04-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-04-SUMMARY.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-05-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-05-SUMMARY.md", "README.en.md", "README.md", "cmd/goswitchd/main.go", "cmd/goswitchd/main_test.go", "docs/CONFIG.md", "docs/SPEC.md", "docs/adr/adr-008-spike-log.md", "internal/appid/appid.go", "internal/appid/appid_test.go", "internal/appid/export_test.go", "internal/config/config.go", "internal/config/config_test.go", "internal/config/load.go", "internal/config/writer.go", "internal/config/writer_test.go", "internal/indicator/indicator.go", "internal/indicator/menu.go", "internal/indicator/menu_test.go", "internal/session/actor.go", "internal/session/actor_test.go", "mise.toml", "skills/goswitch-config/SKILL.md", "test/e2e/case_menu.go", "test/e2e/case_menu_blocklist.go", "test/e2e/case_menu_blocklist_test.go", "test/e2e/main.go"]
covered_digest: "v1:sha256:57d8a8196d3778a47fa8e699216babddd8ead7b2b9aca659f4a04d287224e8b0"
behavior_unverified: 0
overrides_applied: 0
behavior_unverified_items: []
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
**Verified:** 2026-11-01T01:12:06Z
**Status:** passed (15/15 — живой ряд закрыл truth 15 на ВМ-стенде 2026-11-01; см. строку 15 и UAT)
**Re-verification:** No — initial verification

## Goal Achievement

Фаза достигнута: вся цепочка MENU-01 существует в коде, проводена и доказана поведенчески — юнит-корпусом (трассер клика, гвард-матрица, writer-корпус, fresh-машина) и живым e2e на ВМ-стенде с датированными PASS-рядами в adr-008-spike-log.md. Остались только человеческие приёмки (by-ear владельца + Round 2 passive-ряд на живой сессии), обе маршрутированы в UAT end-of-phase самими планами.

### Observable Truths

| #   | Truth   | Status     | Evidence       |
| --- | ------- | ---------- | -------------- |
| 1   | SC1: пункт «Добавить ИмяПриложения в чёрный список» (id 16) виден; клик пишет regexp.QuoteMeta в apps_blocklist (дедуп, потолок 64) и автокоррекция замолкает без рестарта | ✓ VERIFIED | `menu.go:76` menuIDBlocklistApp=16, rootKidsLocked сразу после ACToggle; композиция `blocklistToggle.add` (main.go:459-484): гварды → QuoteMeta → AppendAutocorrectBlocklist → syncReload+FoldAppliedConfig в том же клике; юнит-трассер TestBlocklistClickTracer GREEN; живой ряд adr-008: menu-blocklist 2026-11-01 00:27 UTC PASS (клик id 16 → apps_blocklist [] → [org.gtk.application.gs_e2e_password], witness chars=0→7, fired=0, одна app-blocked абстейнция) |
| 2   | SC2: запись атомарна — golden-дифф «только новая строка», комментарии целы, pre-write валидация, atomic rename | ✓ VERIFIED | `writer.go:92` AppendAutocorrectBlocklist → rewriteAtomically + decodeValid (строгая pre-write семантика Load); корпус 12 тестов GREEN на этом прогоне: AppendGoldenDiff, AppendRoundTrip, AppendDedupNoOp, AppendAtomicWrite (0600, без .tmp), AppendBrokenSourceUntouched |
| 3   | SC3 (в сужении D-03): неизвестная/несвежая личность — пункт неактивен со статической меткой без имени и пояснений; демон не угадывает, :1.NN не становится ни меткой, ни паттерном | ✓ VERIFIED | `blocklistProps`: `!blFresh \|\| blApp == ""` → labelBlocklistApp disabled; `blocklistStateLocked` отдаёт ПУСТОЕ имя при !fresh (двойной барьер D-03); гвард-матрица TestBlocklistGuardMatrixSilent (unknown/stale — writer-счётчик 0, файл байт-нетронут); bridgeApp отказывает namespace реестра (registryPathPrefix, TestBridgeApp_RefusesRegistryNamespace GREEN) — мусорная identity «org» невозможна; сужение «без объяснительного текста» задокументировано решением владельца 2026-10-31 в SPEC-ревизии, текст роадмапа не переписан |
| 4   | SC4: имя с метасимволами не ломает конфиг (строгий декодер принимает); отказ потолка 64 виден состоянием пункта в трее | ✓ VERIFIED | TestWriter_AppendMetacharPatterns GREEN (QuoteMeta-литерал переживает строгий Load); TestBlocklistClickTracer/metacharacter_identity_lands_as_a_literal GREEN (`org\.example\.App\[1\]`); labelBlocklistFull «Чёрный список заполнен — 64/64» disabled; sentinel errAutocorrectBlocklistOverCeil ДО записи, файл байт-нетронут. Примечание: «уведомление из трея» реализовано состоянием пункта меню — осознанное решение (десктоп-уведомления на этом столе невидимы, вердикт UAT фазы 7; зафиксировано в CONTEXT и SPEC-ревизии) |
| 5   | D-04: macroChordWhat « — язык»; метка «super+space — язык»; README.md:102 и README.en.md:99 синхронны | ✓ VERIFIED | `menu.go:102` macroChordWhat = " — язык"; grep « — язык» в README.md и README.en.md зелёный; пин в TestMenuLayout GREEN |
| 6   | D-01: covered — галочка (toggle-state 1); клик по отмеченному — no-op | ✓ VERIFIED | Матрица TestMenuBlocklistStateMatrixMarks GREEN (covered → checkmark); TestBlocklistGuardMatrixSilent/covered GREEN (writer не вызван, файл байт-идентичен) + дедуп писателя TestWriter_AppendDedupNoOp — двойной пояс; живой ряд: повторный клик — документ БАЙТ-ИДЕНТИЧЕН |
| 7   | D-02: props пункта считаются AT REQUEST TIME из снапшота; живых a11y-запросов при открытии меню нет | ✓ VERIFIED | `blocklistProps` читает только поля снапшота blApp/blFresh/blCovered/blFull (контракт в doc-комментарии); TestMenuBlocklistItemGroupProps GREEN; контрольный факт SPEC-ревизии |
| 8   | Freshness-машина: gain ставит fresh=true и пушит; деарм — blur отслеживаемой поверхности; engine-шум/synthetic свежесть не трогают | ✓ VERIFIED | Поведенческие тесты GREEN на этом прогоне: TestActor_BlocklistItemFreshPushes/Losses (blur деармит, engine churn не трогает, unknown no-op), TestObserver_GainHook/LossHook (own/foreign/relayed blur) |
| 9   | Проводка OnConn: Callbacks.BlockCurrentApp, install-пуш начального состояния, EnsureAppid (безусловный старт наблюдателя) | ✓ VERIFIED | main.go:267 (`BlockCurrentApp: toggles.block.add`), :275 installMenuState, :284 actor.EnsureAppid(); menu.go:468-471 dispatchClick → cb.BlockCurrentApp с nil-гвардом; appid.go:3118 `appid.Start(ctx, appid.WithGainHook(a.PushFocusedApp))`; компоненты спиннуты (TestActor_EnsureAppidUnconditional, install-пуш); живая сборка доказана e2e: arm-гейт читает GetGroupProperties живого демона (PASS 00:27 UTC) |
| 10  | Потолок 64 — единый источник: config.MaxAutocorrectBlocklist экспортирован; actor читает его; старые имена отсутствуют | ✓ VERIFIED | config.go:42 `const MaxAutocorrectBlocklist = 64`; actor.go:3082 `len(...) >= config.MaxAutocorrectBlocklist`; grep autocorrectBlocklistCeil и maxAutocorrectBlocklist по репозиторию — 0 совпадений; TestMaxAutocorrectBlocklistExport GREEN. (См. coincidental_reliance: метка «64/64» — зеркальный литерал, IN-01) |
| 11  | Гейты D-53 байт-неизменны (autoCorrectBoundary/acFocusedApp/терминальное вето не тронуты) | ✓ VERIFIED | 24 gate-теста (TestAutoCorrect|TestAutoConfirm) GREEN на этом прогоне под -race; ревьюер независимо сверил тела функций с до-плановым 34059d8 — BYTE-IDENTICAL; fresh-признак — надстройка только для меню, семантика кэша для гейтов не менялась |
| 12  | SPEC §4.3 несёт датированную ревизию меню v2 (2026-10-31); CONFIG.md дополнен; семантические строки D-53 целы | ✓ VERIFIED | grep-гейты: «2026-10-31» ×5, « — язык» ×1, «Неизвестная идентичность приложения НЕ запрет» ×2 в docs/SPEC.md; «пункт меню» ×2, «QuoteMeta» ×1 в docs/CONFIG.md; ревизия покрывает все шесть пунктов плана (пункт+позиция, состояния D-01/D-03/потолок, QuoteMeta-действие, freshness, D-04, принятый риск a11y-шины); SKILL.md-регион синхронен (TestSkillgen_CommittedFileInSync GREEN) |
| 13  | Round 2: клиентский Reset окна aftermath не стирает выученный класс, настоящий Reset вне окна стирает (fail-closed); PR #25 → main; сборка установлена | ✓ VERIFIED | TestTerminalClass_VerifyDeadSurvivesClientResetAfterOwnFlip + TestTerminalClass_VerifyDeadClearedByResetOutsideAftermath + пины Round 1 — GREEN на этом прогоне; RED f1f0019 / GREEN 2c00c10 в origin/main через merge 58b07cc (PR #25); host unit active, ExecStart = releases/v1.2.0-alpha.2-15-g2c00c10 (проверено верификатором); строка «Round 2 install» датирована, откат назван |
| 14  | e2e-кейс menu-blocklist зарегистрирован, mise-задача на месте; живой цикл + re-pin menu-v2 — зелёные на ВМ | ✓ VERIFIED | test/e2e/case_menu_blocklist.go (454 строки; пин id 16, Event-канал, оракулы файла/Load/тишины/no-op, teardown-снапшот); «menu-blocklist» в реестре main.go:422; [tasks.e2e-menu-blocklist] mise.toml:279; датированные ряды adr-008: menu-blocklist PASS 00:27 UTC (HEAD 489aa5f), menu-v2 re-pin PASS 00:30 UTC (HEAD cf9c857, id 1/2/4/5 не сдвинуты); TestParseBlocklistDoc/TestAssertOnlyBlocklistLine GREEN |
| 15  | «Round 2 journal verify»: живой ряд фиксирует следующую коррекцию с одного нажатия | ✓ VERIFIED (ВМ-стенд, 2026-11-01 20:26–20:29 UTC) | Владелец делегировал живую приёмку стенду («Тестируй сам в виртуалке», 2026-11-01). Живой ряд на сборке фикса g2c00c10 (Orca Wayland-нативно; драйв virsh send-key — композитор-доверенная клавиатура ВМ, канал найден с 4-й попытки, см. debug-док §Round 2 VM-попытка): один штатный verify-раунд обучения нового контекста (20:27:19), затем 4 подряд коррекции done changed=true с ОДНОГО double-shift (20:27:45 со 2-го нажатия — воспроизведение исходного бага; 20:28:05/20:28:40/20:28:45 — каждая с одного), каждая с собственным флипом (mode to ru/en) и ре-минтом движка (/3→/4); регресс-маркер verify_dead=false — 0 строк за сессию. Оговорка честности: lifecycle aftermath=true не наблюдался — эхо-Reset от ВМ-клиента не пришло (глотать нечего, класс выжил и без карве-аута; само проглатывание pinned юнит-парой truth 13). Хостовой пассивный ряд продолжает накапливаться естественно. UAT Тест 2 закрыт этим рядом |

**Score:** 15/15 truths verified (truth 15 закрыт живым ВМ-рядом на сборке фикса 2026-11-01: одна штатная инициализация verify-раунда, далее каждая коррекция — с одного double-shift, включая следующие за собственным флипом демона; verify_dead=false — 0 строк)

**Верификатором независимо перепрогнано:** go build ./..., go vet ./... (включая ./test/e2e/), golangci-lint 2.15.0 — 0 issues на всех фазовых пакетах (после сброса GOROOT — задокументированный toolchain-разъём WINDOWS #23-26, не дефект кода), целевые -race корпусы пяти пакетов, TestSkillgen_CommittedFileInSync, host-юнит active на Round 2-релизе, grep-гейты доков.

### Deferred Items

Не обнаружены: хвост «удаление из блок-листа через меню» — deferred-идея CONTEXT вне объёма MENU-01 (кандидат после v1.2), требованием фазы никогда не был; фазы 12-13 его не покрывают и не обязаны. Round 2 journal-verify не деферится в фазу — он маршрутирован в UAT end-of-phase (см. human_verification).

### Advisory (New Scope, Unevidenced)

Не применимо — initial verification (is_re_verification = false).

Примечания по известным находкам ревью (11-REVIEW.md WR-01..WR-04, IN-01, IN-02 — документированные advisories): ни одна не инвалидирует must-have. WR-01 (aftermath глотает реальный Reset при аномальной смерти контекста в окне 2с; одноразовое окно) — fail-closed вне окна запинен, хвост задокументирован; WR-02 (деарм blur по имени, не по поверхности) — ложный деарм fail-closed-направления (unknown, мусор не пишется); WR-03 (одноразовый старт наблюдателя защёлкивается при раннем неудачном дайле) — маловероятный хвост, задокументирован; WR-04 (e2e-оракул D-01 теоретически вакуумный) — путь covered-no-op запинен юнит-матрицей. Дублирующая строка WINDOWS #27 (= #26, 9 секунд спустя) — шум bookkeeping, не дефект.

Flagged assumption из SUMMARY (spec-less probe fallback: `{"requirement_id":"MENU-01","category":"unclassified","status":"unresolved"}`) — учтена: край «неизвестная личность» фактически покрыт (7-строчная матрица состояний, fresh-машина, гвард-матрица, guard реестра — все GREEN); строка остаётся формально unresolved как артефакт классификации проба, не как пробел реализации.

### Required Artifacts

| Artifact | Expected    | Status | Details |
| -------- | ----------- | ------ | ------- |
| `internal/indicator/menu.go` | id 16, матрица состояний, сеттер, dispatchClick, D-04 | ✓ VERIFIED | все элементы найдены; verify.artifacts 4/4 |
| `internal/indicator/indicator.go` | Callbacks.BlockCurrentApp | ✓ VERIFIED | поле + проводка main.go:267 |
| `internal/indicator/menu_test.go` | layout/матрица/дельты/клик пины | ✓ VERIFIED | 3+ корпус GREEN |
| `internal/appid/appid.go` | WithGainHook вне o.mu; guard реестра | ✓ VERIFIED | :116-123, :205-211; TestObserver_GainHook/LossHook, TestBridgeApp_RefusesRegistryNamespace GREEN |
| `internal/session/actor.go` | PushFocusedApp/fresh, BlocklistAppState, EnsureAppid, fold-пуш, Round 2 aftermath | ✓ VERIFIED | :1572/:1601/:1615/:3072-3110/:1911-1988 |
| `internal/session/actor_test.go` | freshness/fold/accessor/ensure + Round 2 пины | ✓ VERIFIED | 9+ тест-функций GREEN |
| `internal/config/writer.go` | AppendAutocorrectBlocklist | ✓ VERIFIED | :92-121; 12-тест корпус GREEN |
| `internal/config/load.go` | decodeValid (единая Load-семантика) | ✓ VERIFIED | используется pre-write гейтом writer.go:274 |
| `internal/config/config.go` | MaxAutocorrectBlocklist=64 + sentinel | ✓ VERIFIED | :42, :72; TestMaxAutocorrectBlocklistExport GREEN |
| `cmd/goswitchd/main.go` | click-композиция + проводка OnConn | ✓ VERIFIED | :267/:275/:284/:425-484 |
| `cmd/goswitchd/main_test.go` | трассер + гвард-матрица | ✓ VERIFIED | TestBlocklistClickTracer + GuardMatrix/Refusals/Order GREEN |
| `test/e2e/case_menu_blocklist.go` | живой кейс с оракулами | ✓ VERIFIED | 454 строки; пин id 16, teardown |
| `test/e2e/main.go` + `mise.toml` | регистрация + задача | ✓ VERIFIED | main.go:422, mise.toml:279 |
| `docs/SPEC.md` / `docs/CONFIG.md` | датированная ревизия + контракт | ✓ VERIFIED | все grep-гейты зелёные |
| `README.md` / `README.en.md` | D-04 строки | ✓ VERIFIED | :102 / :99 « — язык» |
| `.planning/debug/orca-terminal-second-shift.md` | ряды Round 2 | ✓ VERIFIED | install (23:11-23:12 MSK, release-dir, откат) + journal verify (честный ряд-ожидание) |

Примечание: verify.artifacts для 11-02 показал 4/5 — «README.md / README.en.md File not found» это артефакт парсинга frontmatter (комбинированный путь «README.md / README.en.md» в одном поле), оба файла существуют и пины несут. Ложное срабатывание инструмента, не пробел.

### Key Link Verification

Инструмент verify.key-links не парсит аннотированные from/to («file (context)» → «Source file not found» для всех 13 линков всех планов) — все линки проверены вручную:

| From | To  | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| menu.go dispatchClick menuIDBlocklistApp | main.go композиция | Callbacks.BlockCurrentApp | ✓ WIRED | menu.go:468-471 → main.go:267 |
| main.go композиция | actor.BlocklistAppState + FoldAppliedConfig | аксессор + toggleReload | ✓ WIRED | main.go:417 state:, :248-255 reload = syncReload+Fold |
| main.go композиция | config.AppendAutocorrectBlocklist | единственный persist path, QuoteMeta до вызова | ✓ WIRED | main.go:419, :469 |
| writer.go append | rewriteAtomically | temp same-dir + rename 0600 | ✓ WIRED | writer.go:92→121 |
| writer.go pre-write | load.go decodeValid (KnownFields+Validate) | отказ до rename | ✓ WIRED | writer.go:274 |
| config.MaxAutocorrectBlocklist | actor full-состояние | единый источник | ✓ WIRED | actor.go:3082 |
| appid.go run() хук | actor.PushFocusedApp | WithGainHook вне o.mu | ✓ WIRED | actor.go:3118 |
| HandleLifecycle/blur | fresh=false + пуш | деарм слоя identity | ✓ WIRED | actor.go:1572-1595; TestActor_BlocklistItemFreshLosses |
| actor pushMenuSync/fold | menu.SetBlocklistApp | MenuSync-шов, дельты | ✓ WIRED | actor.go:3100-3110 → menu.go:320-321 |
| SPEC-ревизия меню v2 | internal/indicator/menu.go | спека прежде кода | ✓ WIRED | ревизия docs/SPEC.md:588-651 ↔ реализация (id 16, D-01/D-03, D-04) |
| debug-док Round 2 гипотеза | actor.go HandleLifecycle | aftermath-карве-аут | ✓ WIRED | flipAftermathWindow/verifyDeadOnContextDeath :124/:1938; тест GREEN |
| e2e case Event id 16 | живой демон dbusmenu.Event | канал case_menu.go | ✓ WIRED | case:40-43/:402; живой PASS-ряд adr-008 00:27 UTC |
| e2e silence oracle | case_autocorrect.go blocklist-silent машина | app-blocked абстейнция | ✓ WIRED | acReasonAppBlocked переиспользован; fired=0, одна app-blocked в живом ряду |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| Метка пункта меню | m.blApp/blFresh | a11y-gain → WithGainHook → PushFocusedApp → SetBlocklistApp | да (живой observер) | ✓ FLOWING |
| covered/галочка | blCovered | fold applySnapshot реального конфиг-документа (re.MatchString) | да | ✓ FLOWING |
| Конфиг-документ | apps_blocklist | клик → QuoteMeta → AppendAutocorrectBlocklist → YAML-файл | да (живой ряд: [] → [org.gtk.application.gs_e2e_password]) | ✓ FLOWING |
| Гвард-тишина | ac_app_blocked | блок-гейт автокоррекции по документу | да (живой ряд: fired=0, одна app-blocked) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Layout/матрица/дельты/клик меню | `go test ./internal/indicator/ -race -run 'TestMenuLayout\|TestMenuBlocklist'` | ok 1.0s | ✓ PASS |
| Fresh-машина + fold + аксессор + Round 2 полный цикл + fail-closed пины | `go test ./internal/session/ -race -run 'TestActor_Blocklist\|TestTerminalClass_VerifyDead'` | 16/16 PASS (вкл. VerifyDeadSurvivesClientResetAfterOwnFlip, ClearedByResetOutsideAftermath) | ✓ PASS |
| Writer: golden/дедуп/потолок/метасимволы/atomic/strict | `go test ./internal/config/ -race -run 'TestWriter_Append\|TestMaxAutocorrectBlocklistExport'` | 13/13 PASS | ✓ PASS |
| Клик-композиция: трассер + гвард-матрица + порядок + QuoteMeta | `go test ./cmd/goswitchd/ -race -run 'TestBlocklist'` | 10 subtest PASS | ✓ PASS |
| Gain/loss хуки + guard реестра | `go test ./internal/appid/ -race -run 'TestObserver_GainHook\|TestObserver_LossHook\|TestBridgeApp_RefusesRegistryNamespace'` | PASS | ✓ PASS |
| Гейты D-53 | `go test ./internal/session/ -race -run 'TestAutoCorrect\|TestAutoConfirm'` | ok | ✓ PASS |
| Зелёная итерация | `go build ./... && go vet ./... && golangci-lint run` (GOROOT reset — WINDOWS #23-26 seam) | OK; 0 issues | ✓ PASS |
| Хост-установка Round 2 | `systemctl --user is-active goswitchd` + ExecStart | active; releases/v1.2.0-alpha.2-15-g2c00c10 | ✓ PASS |
| SKILL.md-регион синхронен | `go test ./cmd/skillgen/ -run TestSkillgen_CommittedFileInSync` | ok | ✓ PASS |

### Probe Execution

Конвенциональных `scripts/*/tests/probe-*.sh` в проекте нет. Фазовый probe — живой e2e-прогон 11-05 (ssh-ритуал ВМ). Перепрогон верификатором невозможен из этой среды (живая GNOME-сессия ВМ; директива владельца de702f6 от 2026-11-01: проверки продукта — только на ВМ) — вердикт зафиксирован датированными рядами репозитория, произведёнными ровно теми командами, что пинит план:

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| menu-blocklist (заезд 8) | `go run ./test/e2e -case menu-blocklist` (ВМ, сессионный env, HEAD 489aa5f) | adr-008:2026-11-01 00:27 UTC — CASE_EXIT=0, все оракулы зелёные | ✓ PASS (датированный ряд) |
| menu-v2 re-pin (заезд 2) | `go run ./test/e2e -case menu-v2` (ВМ, HEAD cf9c857) | adr-008:2026-11-01 00:30 UTC — CASE_EXIT=0, id 1/2/4/5 не сдвинуты | ✓ PASS (датированный ряд) |
| menu-v2 re-pin заезд 1 | тот же, HEAD 489aa5f | рестарт-survival упал (среда активации, не сдвиг пинов); фикс cf9c857 | ℹ️ честный FAIL среды, закрыт фиксом |
| Лестница отказов заездов 1-7 | — | 7 setup/oracle-отказов с причинами, фиксы отдельными коммитами, оракулы не полировались | ℹ️ зафиксирована в adr-008 |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| MENU-01 | 11-01..11-05 (все пять) | Динамический пункт «Добавить ИмяПриложения в чёрный список»: QuoteMeta, без дублей, потолок 64, атомарная запись + живая перечитка; при неизвестной личности пункт неактивен, демон не угадывает | ✓ SATISFIED | Полная цепочка: пункт id 16 (11-02) → клик-композиция с гвардами (11-04) → writer (11-03) → живой e2e на ВМ (11-05); SPEC/CONFIG-ревизия (11-01). REQUIREMENTS.md:19 отмечен [x], таблица :53 «Complete» — трассируемость подтверждена |

Orphaned requirements: нет — только MENU-01 маппится на Phase 11, объявлен всеми пятью планами.

### Decision Coverage

All trackable CONTEXT.md decisions are honored by shipped artifacts: 4/4 (D-01 галочка/no-op; D-02 кэш без живых a11y-запросов; D-03 неактивен без имени и пояснений; D-04 « — язык»). Discretion (позиция под тумблером) исполнена. (gsd-tools check.decision-coverage-verify: honored 4, not_honored 0.)

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| internal/indicator/menu_test.go (blocklist-корпус) | MENU-01 | да | 0 | нет | Value/Behavioral (матрица 7 строк, дельты, dispatch) | VALID |
| internal/session/actor_test.go (freshness/fold/accessor + Round 2) | MENU-01 | да | 0 | нет | Behavioral (полные циклы переходов) | VALID |
| internal/config/writer_test.go + config_test.go (append-корпус) | MENU-01 | да | 0 | нет | Value (golden-дифф, байт-сравнения, потолок) | VALID |
| cmd/goswitchd/main_test.go (трассер + гвард-матрица) | MENU-01 | да | 0 | нет | Behavioral (write→reload порядок, счётчик writer) | VALID |
| internal/appid/appid_test.go (GainHook/LossHook/registry guard) | MENU-01 | да | 0 | нет | Behavioral | VALID |
| test/e2e/case_menu_blocklist_test.go (parse/swap пины) | MENU-01 | да | 0 | нет | Value | VALID |

Disabled tests on requirements: 0. Circular patterns: 0. Insufficient assertions: 0.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| — | — | TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER во всех фазовых файлах | — | не обнаружено (debt-marker gate чист) |
| — | — | placeholder/coming soon/not-yet-implemented | — | не обнаружено |
| — | — | пустые реализации / console-only | — | не обнаружено |
| go.mod/go.sum | — | новые модули | — | ноль (T-11-SC) |

### Human Verification Required

### 1. By-ear приёмка пункта меню владельцем (UAT end-of-phase, прецедент G-10-1)

**Test:** В реальном трее живой сессии: открыть меню — пункт «Добавить ИмяПриложения в чёрный список» в группе автокоррекции под тумблером; проверить точное имя наблюдаемого приложения; кликнуть — приложение блокируется; повторный клик по отмеченному — ничего не меняется; уйти в ненаблюдаемое приложение (snap/wezterm) — пункт гаснет без имени и пояснений; сверить строку «Super+Space — язык».
**Expected:** Все состояния D-01/D-03 визуально корректны; автокоррекция в заблокированном приложении молчит; формулировка D-04 читается «Super+Space — язык».
**Why human:** Визуальный рендер, эргономика клика и by-ear зачёт — VALIDATION.md Manual-Only; машинные оракулы (файл/Load/тишина/no-op, всё зелёное на ВМ) эстетику и удобство не заменяют.

### 2. Round 2 passive journal-верификация (владелец, следующая десктоп-сессия)

**Test:** В окне Orca прогнать полный цикл «коррекция → флип → клиентский Reset → следующая коррекция»; сверить журнал демона с оракулом из .planning/debug/orca-terminal-second-shift.md §Round 2 journal verify.
**Expected:** Следующая коррекция с ОДНОГО нажатия; INFO-ряд lifecycle aftermath=true; ноль строк terminal class verify_dead=false между flip-released и следующей коррекцией.
**Why human:** Окно Orca Wayland-нативно, вне достижимости исполнительного тулинга; инъекция в живую сессию запрещена (T-09-02-03); сам переход класса уже спиннут юнит-тестом (зелёный), не верифицирован только live-host прогон.

### Gaps Summary

Gaps нет. Все артефакты фазы существуют, существенны и проводены; все 13 key-links живы; данные текут от реальных источников; живой цикл MENU-01 доказан дважды (in-process трассер + датированный e2e PASS на ВМ); гейты D-53 и семантика кэша не тронуты; зелёная итерация подтверждена верификатором независимо (build/vet/lint 0 issues/тесты -race). Статус human_needed обусловлен только двумя human-приёмками, обе маршрутированы в UAT end-of-phase самими планами (не синтезированные вердикты): by-ear владельца (G-10-1) и Round 2 passive-ряд на следующей сессии. Advisory-хвосты ревью WR-01..WR-04/IN-01/IN-02 задокументированы, must-have не инвалидируют.

---

_Verified: 2026-11-01T01:12:06Z_
_Verifier: Claude (gsd-verifier)_
