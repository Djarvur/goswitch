---
phase: 11-menu-punkt-v-chernyy-spisok
verified: 2026-11-06T15:54:17Z
status: passed
score: 15/15 must-haves verified
covered_files: [".planning/REQUIREMENTS.md", ".planning/debug/orca-terminal-second-shift.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-01-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-01-SUMMARY.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-02-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-02-SUMMARY.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-03-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-03-SUMMARY.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-04-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-04-SUMMARY.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-05-PLAN.md", ".planning/phases/11-menu-punkt-v-chernyy-spisok/11-05-SUMMARY.md", "README.en.md", "README.md", "cmd/goswitchd/main.go", "cmd/goswitchd/main_test.go", "docs/CONFIG.md", "docs/SPEC.md", "docs/adr/adr-008-spike-log.md", "internal/appid/appid.go", "internal/appid/appid_test.go", "internal/appid/export_test.go", "internal/config/config.go", "internal/config/config_test.go", "internal/config/load.go", "internal/config/writer.go", "internal/config/writer_test.go", "internal/indicator/indicator.go", "internal/indicator/menu.go", "internal/indicator/menu_test.go", "internal/session/actor.go", "internal/session/actor_test.go", "mise.toml", "skills/goswitch-config/SKILL.md", "test/e2e/case_menu.go", "test/e2e/case_menu_blocklist.go", "test/e2e/case_menu_blocklist_test.go", "test/e2e/main.go"]
covered_digest: "v4:sha256:f036cc3a179116df4738328c6df3b897923084d34863b425ca8a75c4214f8a16"
behavior_unverified: 0
overrides_applied: 0
behavior_unverified_items: []
re_verification:
  previous_status: passed
  previous_score: 15/15
  stale_reason: "Прошлый отчёт (verified 2026-11-02T19:56:09Z, passed 15/15, digest v4:ef7799…, коммит dda1e0b) устарел ТОЛЬКО записями в covered-файлах после прогона: закрытие фаз 12/13 и merge c54a388 (HEAD) добавили контент в 7 covered-входов — REQUIREMENTS.md (трекинг статусов 12/13), README.md/README.en.md, docs/CONFIG.md, docs/adr/adr-008-spike-log.md, mise.toml, test/e2e/main.go (регистрация term-line-probe). Всё добавленное — аддитивная работа фаз 12/13; ни одна строка фазы 11 не изменена: git diff 6086cdb..HEAD по internal/session/, internal/indicator/, internal/config/, internal/appid/, cmd/ — ПУСТ. Пробелов в прошлом отчёте не было; это stale-refresh по текущему дереву"
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
**Verified:** 2026-11-06T15:54:17Z
**Status:** passed (15/15 — stale-refresh после закрытия фаз 12/13 и merge c54a388; обе human-приёмки закрыты 2026-11-01 и перенесены как закрытые)
**Re-verification:** Yes — stale-refresh: прошлый PASS (15/15, 2026-11-02, digest ef7799…) устарел записями фаз 12/13 в covered-входах (трекинг/доки/задачи/реестр e2e); исходный код фазы 11 не тронут (пустой дифф 6086cdb..HEAD по всем фазовым пакетам); пробелов в прошлом отчёте не было

## Goal Achievement

Фаза достигнута и остаётся достигнутой на текущем дереве (HEAD c54a388 — merge ветки Фазы 12 в ветку Фазы 13, «веха v1.2 все фазы закрыты»). Вся цепочка MENU-01 существует, проводна и доказана поведенчески: пункт id 16 (menu.go), fresh-машина (actor.go), writer, клик-композиция с гвардами (main.go), живой e2e на ВМ (adr-008, датированные PASS-ряды). Причина устаревания — пост-прогонные записи фаз 12/13 в covered-входах — не тронула ни одного механизма фазы 11: дифф 6086cdb..HEAD по internal/session/, internal/indicator/, internal/config/, internal/appid/, cmd/ ПУСТ (проверено верификатором этого прогона); изменения ограничились аддитивным контентом фаз 12/13 в REQUIREMENTS.md (трекинг статусов), README ×2 и docs/CONFIG.md (рецепры A11Y), adr-008 (новые ряды 12/13), mise.toml (задача e2e-term-line-probe) и test/e2e/main.go (регистрация term-line-probe, +10 строк вокруг — фазовая строка «menu-blocklist» цела, см. Behavioral Spot-Checks). Обе human-приёмки закрыты 2026-11-01 живыми ВМ-рядами по делегации владельца («Тестируй сам в виртуалке») — перенесены как закрытые, не переоткрыты.

### Observable Truths

Все 15 истин прошлого отчёта перепроверены на HEAD c54a388 (quick-regression по политике re-verification: существование + здравость + перепрогон корпусов; исходники фазовых пакетов байт-идентичны состоянию прошлого прогона, поэтому номера строк сдвинулись только в test/e2e/main.go и mise.toml — обновлены в таблицах ниже).

| #   | Truth   | Status     | Evidence       |
| --- | ------- | ---------- | -------------- |
| 1   | SC1: пункт «Добавить ИмяПриложения в чёрный список» (id 16) виден; клик пишет regexp.QuoteMeta в apps_blocklist (дедуп, потолок 64) и автокоррекция замолкает без рестарта | ✓ VERIFIED (перепроверено на HEAD c54a388) | menu.go:76 menuIDBlocklistApp=16 (реестр :501, rootKids после ACToggle); композиция `blocklistToggle.add` (main.go:436-484): empty-path WARN → гварды known/covered/full → `regexp.QuoteMeta(app)` main.go:469 → `t.write` = config.AppendAutocorrectBlocklist (main.go:419) → `t.reload` = syncReload+FoldAppliedConfig (main.go:254); корпус `TestBlocklist*` — ok 1.039s на этом прогоне; живой ряд adr-008:716 (menu-blocklist PASS 2026-11-01 00:27 UTC) |
| 2   | SC2: запись атомарна — golden-дифф «только новая строка», комментарии целы, pre-write валидация, atomic rename | ✓ VERIFIED (перепроверено) | writer.go:92 AppendAutocorrectBlocklist → rewriteAtomically (:121) + decodeValid pre-write гейт (:274, семантика Load); корпус `TestWriter_Append*\|TestMaxAutocorrectBlocklistExport` — ok 1.213s на этом прогоне |
| 3   | SC3 (в сужении D-03): неизвестная/несвежая личность — пункт неактивен со статической меткой без имени и пояснений; демон не угадывает, :1.NN не становится ни меткой, ни паттерном | ✓ VERIFIED (перепроверено) | Двойной барьер: actor `blocklistStateLocked` (actor.go:3133+) отдаёт ПУСТОЕ имя при !fresh; menu `blocklistProps` — `!fresh \|\| app==""` → labelBlocklistApp disabled (menu.go:87 статическая метка); bridgeApp отказывает namespace реестра (appid.go, registryPathPrefix); `TestBridgeApp_RefusesRegistryNamespace` + `TestBlocklistGuardMatrixSilent` GREEN (корпуса этого прогона) |
| 4   | SC4: имя с метасимволами не ломает конфиг (строгий декодер принимает); отказ потолка 64 виден состоянием пункта в трее | ✓ VERIFIED (перепроверено) | TestWriter_AppendMetacharPatterns GREEN (корпус этого прогона); labelBlocklistFull «Чёрный список заполнен — 64/64» (menu.go:94) disabled; sentinel errAutocorrectBlocklistOverCeil (config.go:72) ДО записи, файл байт-нетронут. «Уведомление из трея» = состояние пункта — осознанное решение (десктоп-уведомления невидимы, UAT фазы 7; зафиксировано в CONTEXT и SPEC-ревизии) |
| 5   | D-04: macroChordWhat « — язык»; метка «super+space — язык»; README синхронны | ✓ VERIFIED (перепроверено) | menu.go:102 macroChordWhat = " — язык"; grep « — язык»: README.md:100,102 и README.en.md:97,99 (стрелки не сдвинуты README-дополнениями фаз 12/13); SPEC « — язык» ×1; пин в TestMenuLayout GREEN |
| 6   | D-01: covered — галочка (toggle-state 1); клик по отмеченному — no-op | ✓ VERIFIED (перепроверено) | TestMenuBlocklistStateMatrixMarks GREEN; TestBlocklistGuardMatrixSilent/covered + TestWriter_AppendDedupNoOp GREEN (корпуса этого прогона); живой ряд: повторный клик — документ БАЙТ-ИДЕНТИЧЕН (adr-008:716) |
| 7   | D-02: props пункта считаются AT REQUEST TIME из снапшота; живых a11y-запросов при открытии меню нет | ✓ VERIFIED (перепроверено) | blocklistProps читает только поля снапшота blApp/blFresh/blCovered/blFull; TestMenuBlocklistItemGroupProps GREEN (корпус этого прогона) |
| 8   | Freshness-машина: gain ставит fresh=true и пушит; деарм — blur отслеживаемой поверхности (loss-репорт наблюдателя, PushFocusedApp("")); engine-шум/synthetic свежесть не трогают | ✓ VERIFIED (перепроверено на HEAD c54a388; actor.go байт-идентичен прошлому прогону) | `PushFocusedApp` (actor.go:1624): name=="" → blApp=""/blFresh=false+пуш, gain → fresh=true+пуш; `HandleLifecycle` (actor.go:1414-1422) — свежесть НАМЕРЕННО не тронута engine-churn'ом (живая находка 11-05, дизайн зачтён прежней верификацией и UAT); поля blApp/blFresh под a.mu; TestActor_BlocklistItemFreshPushes/FreshLosses — ok 6.239s на этом прогоне (корпус TestActor_Blocklist\|TestTerminalClass_VerifyDead) |
| 9   | Проводка OnConn: Callbacks.BlockCurrentApp, install-пуш начального состояния, EnsureAppid (безусловный старт наблюдателя) | ✓ VERIFIED (перепроверено) | main.go:267 (`BlockCurrentApp: toggles.block.add`), :275 installMenuState, :284 actor.EnsureAppid(); indicator.go:107-111 поле Callbacks.BlockCurrentApp; menu.go:468 dispatchClick → cb.BlockCurrentApp с nil-гвардом; EnsureAppid (безусловный once-per-actor) + ленивый ensureAppid + общее тело startAppidOnce; проводка хука appid.WithGainHook(a.PushFocusedApp); TestActor_EnsureAppidUnconditional + TestObserver_GainHook/LossHook GREEN (корпус этого прогона) |
| 10  | Потолок 64 — единый источник: config.MaxAutocorrectBlocklist экспортирован; actor читает его; старые имена отсутствуют | ✓ VERIFIED (перепроверено) | config.go:42 `const MaxAutocorrectBlocklist = 64`; actor.go:3134 `len(...) >= config.MaxAutocorrectBlocklist`; grep `maxAutocorrectBlocklist` по *.go — 0 совпадений (проверено этим прогоном); TestMaxAutocorrectBlocklistExport GREEN. (См. coincidental_reliance: метка «64/64» — зеркальный литерал, IN-01) |
| 11  | Гейты D-53 байт-неизменны (autoCorrectBoundary/acFocusedApp/терминальное вето не тронуты ФАЗОЙ) | ✓ VERIFIED (перепроверено: actor.go не менялся с прошлого прогона — пустой дифф 6086cdb..HEAD) | Фаза гейты не тронуты (байт-уровень против до-планового 34059d8 — прошлый отчёт); PR #27 app-memory консульт в classifyTerminal не тронул autoCorrectBoundary/acFocusedApp (полный дифф 5a259ad — прошлый отчёт); после того actor.go не менялся ВООБЩЕ (пустой дифф); корпус гейтов `TestAutoCorrect\|TestAutoConfirm` — ok 1.420s на этом прогоне; семантические строки D-53 в SPEC целы (grep ×2 — этот прогон) |
| 12  | SPEC §4.3 несёт датированную ревизию меню v2 (2026-10-31); CONFIG.md дополнен; семантические строки D-53 целы | ✓ VERIFIED (перепроверено) | grep-гейты зелёные на этом прогоне: «2026-10-31» ×5, « — язык» ×1, «Неизвестная идентичность приложения НЕ запрет» ×2 в docs/SPEC.md; «пункт меню» ×2, «QuoteMeta» ×1 в docs/CONFIG.md (дополнения фаз 12/13 в CONFIG.md фазовых строк не тронули) |
| 13  | Round 2: клиентский Reset окна aftermath не стирает выученный класс, настоящий Reset вне окна стирает (fail-closed); PR #25 → main; сборка установлена | ✓ VERIFIED (перепроверено: actor.go байт-идентичен прошлому прогону) | verifyDeadOnContextDeath (actor.go:1990-2032): Reset при flipAftermathStop != nil → cancel+class survives, Reset вне окна → немедленная стирка; TestTerminalClass_VerifyDeadSurvivesClientResetAfterOwnFlip + ClearedByResetOutsideAftermath — ok 6.239s на этом прогоне; `git merge-base --is-ancestor 2c00c10 HEAD` — да (проверено этим прогоном); `git merge-base --is-ancestor 6086cdb HEAD` — да; хост-юнит active, ExecStart = releases/v1.2.0-alpha.2-83-gda4ced8; строки «Round 2 install»/«Round 2 journal verify» в debug-доке на месте (grep ×2 — этот прогон) |
| 14  | e2e-кейс menu-blocklist зарегистрирован, mise-задача на месте; живой цикл + re-pin menu-v2 — зелёные на ВМ | ✓ VERIFIED (перепроверено) | test/e2e/case_menu_blocklist.go + case_menu_blocklist_test.go на месте (байт-нетронуты — пустой дифф ветки covered); «menu-blocklist» в реестре test/e2e/main.go:424 (стрелка 422→424: +2 строки аддитивной регистрации term-line-probe фазы 12 над ней; факт регистрации перепроверен grep'ом :169/:358/:424); [tasks.e2e-menu-blocklist] mise.toml:286 (было :279 — сдвиг от задач фаз 12/13, проверено grep'ом); датированные ряды adr-008:716-718 целы (grep — этот прогон) |
| 15  | «Round 2 journal verify»: живой ряд фиксирует следующую коррекцию с одного нажатия | ✓ VERIFIED (ВМ-стенд, 2026-11-01 20:26–20:29 UTC — закрыто, перенесено) | Закрыто 2026-11-01 по делегации владельца («Тестируй сам в виртуалке»): живой ряд на сборке фикса g2c00c10 — один штатный verify-раунд обучения, далее 4 подряд коррекции done changed=true с ОДНОГО double-shift, каждая с собственным флипом и ре-минтом; verify_dead=false — 0 строк за сессию; оговорка честности — в UAT Тест 2 (pass). Хвост PR #27 (2026-11-02, first-gesture после churn) — отдельная сессия, отдельная приёмка, фазу 11 не инвалидирует |

**Score:** 15/15 truths verified (0 present-behavior-unverified)

**Верификатором независимо перепрогнано на текущем дереве (HEAD c54a388, merged tree фаз 12/13):** go build ./... OK; корпусы -race — internal/session `TestActor_Blocklist\|TestTerminalClass_VerifyDead` ok 6.239s, `TestAutoCorrect\|TestAutoConfirm` (гейты D-53) ok 1.420s; internal/indicator `TestMenuLayout\|TestMenuBlocklist` ok 1.016s; internal/config `TestWriter_Append\|TestMaxAutocorrectBlocklistExport` ok 1.213s; cmd/goswitchd `TestBlocklist` ok 1.039s; internal/appid `TestObserver_GainHook\|TestObserver_LossHook\|TestBridgeApp_RefusesRegistryNamespace` ok 5.070s; grep-гейты доков/README/mise/e2e-реестра — все зелёные; debt-marker скан единственного изменённого covered-кода (test/e2e/main.go) — чист (rc=1 на обоих grep).

### Deferred Items

Не обнаружены (как и в прошлом отчёте): хвост «удаление из блок-листа через меню» — deferred-идея CONTEXT вне объёма MENU-01; фазы 12-13 его не покрывают и не обязаны.

### Advisory (New Scope, Unevidenced)

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| — | None | — | новых unevidenced находок Step 7 на этом прогоне нет; записи фаз 12/13 в covered-входах — задокументированная, протестированная, владельцем-принятая пост-фазовая работа (рецепты A11Y, задача/кейс term-line-probe), не находка |

Примечания по известным находкам ревью (11-REVIEW.md WR-01..WR-04, IN-01, IN-02 — документированные advisories, перенесены из прошлого отчёта): ни одна не инвалидирует must-have. WR-03 частично разрешён PR #27 (снят фальшивой посылкой — reset keys by app identity, TestFlapWindowContextIdentity_DifferentAppWipesClass GREEN — прошлый отчёт).

Flagged assumption из SUMMARY (spec-less probe fallback: `{"requirement_id":"MENU-01","category":"unclassified","status":"unresolved"}`) — учтена (перенесено из прошлого отчёта): край «неизвестная личность» фактически покрыт матрицей состояний, fresh-машиной, гвард-матрицей и guard'ом реестра — все GREEN на текущем дереве; строка остаётся формально unresolved как артефакт классификации проба, не как пробел реализации.

### Required Artifacts

Все фазовые артефакты перепроверены существованием + здравостью (grep-пины этого прогона); исходников, изменившихся с прошлого прогона, среди них НЕТ (пустой дифф 6086cdb..HEAD по всем пакетам):

| Artifact | Expected    | Status | Details |
| -------- | ----------- | ------ | ------- |
| `internal/indicator/menu.go` | id 16, матрица состояний, сеттер, dispatchClick, D-04 | ✓ VERIFIED | :76/:87/:94/:102/:320-321/:416/:468/:501 — все элементы на месте (grep-пины этого прогона) |
| `internal/indicator/indicator.go` | Callbacks.BlockCurrentApp | ✓ VERIFIED | :107-111 |
| `internal/indicator/menu_test.go` | layout/матрица/дельты/клик пины | ✓ VERIFIED | корпус `TestMenuLayout\|TestMenuBlocklist` ok 1.016s на этом прогоне |
| `internal/appid/appid.go` | WithGainHook вне o.mu; guard реестра | ✓ VERIFIED | хук-корпус ok 5.070s на этом прогоне |
| `internal/session/actor.go` | PushFocusedApp/fresh, BlocklistAppState, EnsureAppid, fold-пуш, Round 2 aftermath | ✓ VERIFIED (байт-нетронут с прошлого прогона) | :420-433, :1414-1422, :1612/:1624/:1653/:1667/:3124-3164, :1990-2032 — grep-пины этого прогона |
| `internal/session/actor_test.go` | freshness/fold/accessor/ensure + Round 2 пины | ✓ VERIFIED (байт-нетронут) | корпус ok 6.239s на этом прогоне |
| `internal/config/writer.go` | AppendAutocorrectBlocklist | ✓ VERIFIED | :92/:121/:274; корпус ok 1.213s |
| `internal/config/load.go` | decodeValid (единая Load-семантика) | ✓ VERIFIED | используется pre-write гейтом writer.go:274 |
| `internal/config/config.go` | MaxAutocorrectBlocklist=64 + sentinel | ✓ VERIFIED | :42/:72/:479; 0 использований старого имени (grep этого прогона) |
| `cmd/goswitchd/main.go` | click-композиция + проводка OnConn | ✓ VERIFIED | :254/:267/:275/:284/:329/:419/:469 — grep-пины этого прогона |
| `cmd/goswitchd/main_test.go` | трассер + гвард-матрица | ✓ VERIFIED | корпус `TestBlocklist` ok 1.039s |
| `test/e2e/case_menu_blocklist.go` (+ тест) | живой кейс с оракулами | ✓ VERIFIED | файлы на месте, байт-нетронуты (пустой дифф) |
| `test/e2e/main.go` + `mise.toml` | регистрация + задача | ✓ VERIFIED | main.go:424 (сдвиг +2 от регистрации term-line-probe — фазовая строка цела), mise.toml:286 |
| `docs/SPEC.md` / `docs/CONFIG.md` | датированная ревизия + контракт | ✓ VERIFIED | все grep-гейты зелёные (этот прогон) |
| `README.md` / `README.en.md` | D-04 строки | ✓ VERIFIED | :100/:102 и :97/:99 « — язык» (стрелки не сдвинуты) |
| `.planning/debug/orca-terminal-second-shift.md` | ряды Round 2 | ✓ VERIFIED | «Round 2 install» + «Round 2 journal verify» на месте |
| `docs/adr/adr-008-spike-log.md` | датированные ряды фазы 11 | ✓ VERIFIED | :716-718 целы (дополнения фаз 12/13 — новые секции ниже, фазовые ряды нетронуты) |

### Key Link Verification

Все 13 линков планов перепроверены grep-пинами на текущем дереве (исходники байт-идентичны прошлому прогону; инструмент verify.key-links не парсит аннотированные from/to — как и прежде):

| From | To  | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| menu.go dispatchClick menuIDBlocklistApp | main.go композиция | Callbacks.BlockCurrentApp | ✓ WIRED | menu.go:468 → main.go:267 → blocklistToggle.add |
| main.go композиция | actor.BlocklistAppState + FoldAppliedConfig | аксессор + toggleReload | ✓ WIRED | main.go:419 state-источник, :254 reload = syncReload+Fold |
| main.go композиция | config.AppendAutocorrectBlocklist | единственный persist path, QuoteMeta до вызова | ✓ WIRED | main.go:419, :469 |
| writer.go append | rewriteAtomically | temp same-dir + rename 0600 | ✓ WIRED | writer.go:92→121 |
| writer.go pre-write | load.go decodeValid | отказ до rename | ✓ WIRED | writer.go:274 |
| config.MaxAutocorrectBlocklist | actor full-состояние | единый источник | ✓ WIRED | actor.go:3134 |
| appid.go run() хук | actor.PushFocusedApp | WithGainHook вне o.mu | ✓ WIRED | actor.go (проводка в startAppid); корпус хуков GREEN |
| Identity-layer деарм (loss/blur) | fresh=false + пуш | свежесть живёт в слое identity | ✓ WIRED | actor.go:1624; корпус FreshLosses GREEN |
| actor pushBlocklistAppLocked/fold | menu.SetBlocklistApp | MenuSync-шов, дельты | ✓ WIRED | actor.go:3157 → menu.go:320-321 (дельты covered) |
| SPEC-ревизия меню v2 | internal/indicator/menu.go | спека прежде кода | ✓ WIRED | ревизия SPEC ↔ реализация (id 16, D-01/D-03, D-04) — гейты зелёные |
| debug-док Round 2 гипотеза | actor.go verifyDeadOnContextDeath | aftermath-карве-аут | ✓ WIRED | :1990-2032; тест-пара GREEN на этом прогоне |
| e2e case Event id 16 | живой демон dbusmenu.Event | канал case_menu.go | ✓ WIRED | регистрация main.go:424; живой PASS-ряд adr-008:716 |
| e2e silence oracle | case_autocorrect.go blocklist-silent машина | app-blocked абстейнция | ✓ WIRED | переиспользован; fired=0, одна app-blocked в живом ряду |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| Метка пункта меню | m.blApp/blFresh | a11y-gain → WithGainHook → PushFocusedApp → SetBlocklistApp | да (живой наблюдатель; живой дамп UAT: org.gnome.Zenity) | ✓ FLOWING |
| covered/галочка | blCovered | fold applySnapshot реального конфиг-документа (re.MatchString) | да | ✓ FLOWING |
| Конфиг-документ | apps_blocklist | клик → QuoteMeta → AppendAutocorrectBlocklist → YAML-файл | да (живой ряд: [] → [org.gtk.application.gs_e2e_password]) | ✓ FLOWING |
| Гвард-тишина | ac_app_blocked | блок-гейт автокоррекции по документу | да (живой ряд: fired=0, одна app-blocked) | ✓ FLOWING |

### Behavioral Spot-Checks

Все команды исполнены верификатором на текущем дереве (HEAD c54a388, merged tree фаз 12/13) 2026-11-06:

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Fresh-машина + fold + аксессор + Round 2 цикл + fail-closed пины | `go test ./internal/session/ -race -count=1 -run 'TestActor_Blocklist\|TestTerminalClass_VerifyDead'` | ok 6.239s | ✓ PASS |
| Гейты D-53 | `go test ./internal/session/ -race -count=1 -run 'TestAutoCorrect\|TestAutoConfirm'` | ok 1.420s | ✓ PASS |
| Gain/loss хуки + guard реестра | `go test ./internal/appid/ -race -count=1 -run 'TestObserver_GainHook\|TestObserver_LossHook\|TestBridgeApp_RefusesRegistryNamespace'` | ok 5.070s | ✓ PASS |
| Layout/матрица/дельты/клик меню | `go test ./internal/indicator/ -race -count=1 -run 'TestMenuLayout\|TestMenuBlocklist'` | ok 1.016s | ✓ PASS |
| Writer: golden/дедуп/потолок/метасимволы | `go test ./internal/config/ -race -count=1 -run 'TestWriter_Append\|TestMaxAutocorrectBlocklistExport'` | ok 1.213s | ✓ PASS |
| Клик-композиция: трассер + гвард-матрица | `go test ./cmd/goswitchd/ -race -count=1 -run 'TestBlocklist'` | ok 1.039s | ✓ PASS |
| Зелёная итерация (сборка) | `go build ./...` | OK | ✓ PASS |
| Ancestry фиксов в HEAD | `git merge-base --is-ancestor 2c00c10 HEAD && git merge-base --is-ancestor 6086cdb HEAD` | оба в HEAD | ✓ PASS |
| Исходники фазы 11 нетронуты с прошлого прогона | `git diff --stat 6086cdb..HEAD -- internal/session/ internal/indicator/ internal/config/ internal/appid/ cmd/` | пусто (0 файлов) | ✓ PASS |
| Фазовая строка в e2e-реестре (после регистрации term-line-probe фазы 12) | `grep -n "menu-blocklist" test/e2e/main.go` | :169/:358/:424 — регистрация цела | ✓ PASS |
| Хост-юнит | `systemctl --user is-active goswitchd` + ExecStart | active; releases/v1.2.0-alpha.2-83-gda4ced8 | ✓ PASS |
| Doc-гейты | grep SPEC/CONFIG/README/mise/e2e-реестр/debug-док/adr-008 | все зелёные (значения в truths 5/12/13/14) | ✓ PASS |
| Debt-маркеры изменённого covered-кода | `grep -E "TBD\|FIXME\|XXX\|TODO\|HACK\|PLACEHOLDER" test/e2e/main.go` | 0 совпадений (rc=1) | ✓ PASS |

### Probe Execution

Конвенциональных `scripts/*/tests/probe-*.sh` в проекте нет. Фазовый probe — живой e2e-прогон 11-05 (ssh-ритуал ВМ); перепрогон из этой среды невозможен (директива владельца de702f6: проверки продукта — только на ВМ goswitch-stand). Вердикты зафиксированы датированными рядами репозитория, произведёнными ровно теми командами, что пинит план (перенесено; ряды байт-нетронуты — adr-008:716-718 проверены grep'ом этого прогона):

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| menu-blocklist (заезд 8) | `go run ./test/e2e -case menu-blocklist` (ВМ, HEAD 489aa5f) | adr-008:716 — 2026-11-01 00:27 UTC, CASE_EXIT=0, все оракулы зелёные | ✓ PASS (датированный ряд) |
| menu-v2 re-pin (заезд 2) | `go run ./test/e2e -case menu-v2` (ВМ, HEAD cf9c857) | adr-008:718 — 00:30 UTC, CASE_EXIT=0, id 1/2/4/5 не сдвинуты | ✓ PASS (датированный ряд) |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| MENU-01 | 11-01..11-05 (все пять) | Динамический пункт «Добавить ИмяПриложения в чёрный список»: QuoteMeta, без дублей, потолок 64, атомарная запись + живая перечитка; при неизвестной личности пункт неактивен, демон не угадывает | ✓ SATISFIED | Полная цепочка перепроверена на HEAD c54a388: пункт id 16 → клик-композиция с гвардами → writer → живой e2e ВМ (датированные ряды целы); закрытая UAT-приёмка живым дампом. REQUIREMENTS.md:19 [x], таблица :53 «Complete» (проверено этим прогоном — правки статусов 12/13 строку MENU-01 не тронули) |

Orphaned requirements: нет — только MENU-01 маппится на Phase 11, объявлен всеми пятью планами.

### Decision Coverage

All trackable CONTEXT.md decisions are honored by shipped artifacts: 4/4 (D-01 галочка/no-op; D-02 кэш без живых a11y-запросов; D-03 неактивен без имени и пояснений; D-04 « — язык»). Discretion (позиция под тумблером) исполнена (menu.go:501, rootKids после ACToggle).

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| internal/indicator/menu_test.go (blocklist-корпус) | MENU-01 | да | 0 | нет | Value/Behavioral (матрица 7 строк, дельты, dispatch) | VALID |
| internal/session/actor_test.go (freshness/fold/accessor + Round 2) | MENU-01 | да | 0 | нет | Behavioral (полные циклы переходов) | VALID |
| internal/config/writer_test.go + config_test.go (append-корпус) | MENU-01 | да | 0 | нет | Value (golden-дифф, байтовые сравнения, потолок) | VALID |
| cmd/goswitchd/main_test.go (трассер + гвард-матрица) | MENU-01 | да | 0 | нет | Behavioral (write→reload порядок, счётчик writer) | VALID |
| internal/appid/appid_test.go (GainHook/LossHook/registry guard) | MENU-01 | да | 0 | нет | Behavioral | VALID |
| test/e2e/case_menu_blocklist_test.go (parse/swap пины) | MENU-01 | да | 0 | нет | Value | VALID |

Disabled tests on requirements: 0. Circular patterns: 0. Insufficient assertions: 0.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| — | — | TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER в единственном изменённом covered-коде (test/e2e/main.go) и фазовых файлах | — | не обнаружено (rc=1 на обоих grep; прочие covered-файлы байт-нетронуты) |
| — | — | placeholder/coming soon/not-yet-implemented | — | не обнаружено |
| — | — | пустые реализации / console-only | — | не обнаружено |
| go.mod/go.sum | — | новые модули | — | ноль (дифф 6086cdb..HEAD go.mod/go.sum не трогает) |

Evidence gate (#3304): с прошлого прогона среди covered-входов изменились только 7 файлов (REQUIREMENTS.md, README ×2, CONFIG.md, adr-008, mise.toml, test/e2e/main.go) — все проверены: изменения аддитивные (фазы 12/13), фазовые строки/регистрации/ряды целы (grep-гейты зелёные), debt-маркеров нет; детерминированных находок нет, unevidenced новых находок нет (advisory пуст).

### Human Verification Required

Обе приёмки закрыты 2026-11-01 живыми ВМ-рядами по делегации владельца («Тестируй сам в виртуалке», UAT 11-UAT.md: 2/2 pass) — перенесены в frontmatter как закрытые (с `closed:`-строками), не переоткрываются. Новых human-пунктов этот stale-refresh не породил (behavior_unverified: 0).

### Gaps Summary

Gaps нет. Stale-refresh подтверждает прошлый PASS на текущем дереве (HEAD c54a388): единственная причина устаревания — аддитивные записи фаз 12/13 в covered-входы (трекинг REQUIREMENTS, доки README/CONFIG/adr-008, mise-задача и e2e-регистрация term-line-probe) — не тронула ни один механизм фазы 11: исходники всех фазовых пакетов (internal/session, internal/indicator, internal/config, internal/appid, cmd/goswitchd) байт-идентичны состоянию прошлого прогона (пустой дифф 6086cdb..HEAD), фазовая регистрация «menu-blocklist» и задача mise целы (grep-пины), датированные ряды adr-008 нетронуты. Все корпуса перепрогнаны верификатором на merged tree 2026-11-06 — зелёные. Обе human-приёмки закрыты и перенесены. covered_digest пересчитан CLI (`verification.fingerprint`, вывод скопирован дословно): v4:sha256:f036cc3a179116df4738328c6df3b897923084d34863b425ca8a75c4214f8a16 — изменился относительно прошлого (ef7799…) легитимно: 7 covered-входов получили контент фаз 12/13; набор covered_files тот же 38-файловый.

---

_Verified: 2026-11-06T15:54:17Z_
_Verifier: Claude (gsd-verifier)_
