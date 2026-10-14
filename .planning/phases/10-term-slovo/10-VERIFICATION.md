---
phase: 10-term-slovo
verified: 2026-10-14T13:10:58Z
status: human_needed
score: 31/33 must-haves verified
covered_files:
  - .planning/REQUIREMENTS.md
  - .planning/ROADMAP.md
  - .planning/debug/DEBUG-term-host-live-failures.md
  - .planning/phases/10-term-slovo/10-UAT.md
  - .planning/phases/10-term-slovo/10-REVIEW.md
  - .planning/phases/10-term-slovo/10-01-PLAN.md
  - .planning/phases/10-term-slovo/10-01-SUMMARY.md
  - .planning/phases/10-term-slovo/10-02-PLAN.md
  - .planning/phases/10-term-slovo/10-02-SUMMARY.md
  - .planning/phases/10-term-slovo/10-03-PLAN.md
  - .planning/phases/10-term-slovo/10-03-SUMMARY.md
  - .planning/phases/10-term-slovo/10-04-PLAN.md
  - .planning/phases/10-term-slovo/10-04-SUMMARY.md
  - .planning/phases/10-term-slovo/10-05-PLAN.md
  - .planning/phases/10-term-slovo/10-05-SUMMARY.md
  - .planning/phases/10-term-slovo/10-06-PLAN.md
  - .planning/phases/10-term-slovo/10-06-SUMMARY.md
  - .planning/phases/10-term-slovo/10-07-PLAN.md
  - .planning/phases/10-term-slovo/10-07-SUMMARY.md
  - .planning/phases/10-term-slovo/10-08-PLAN.md
  - .planning/phases/10-term-slovo/10-08-SUMMARY.md
  - .planning/phases/10-term-slovo/10-09-PLAN.md
  - .planning/phases/10-term-slovo/10-09-SUMMARY.md
  - .planning/phases/10-term-slovo/10-10-PLAN.md
  - .planning/phases/10-term-slovo/10-10-SUMMARY.md
  - .planning/phases/10-term-slovo/10-11-PLAN.md
  - .planning/phases/10-term-slovo/10-11-SUMMARY.md
  - .planning/phases/10-term-slovo/10-12-PLAN.md
  - .planning/phases/10-term-slovo/10-12-SUMMARY.md
  - internal/appid/appid.go
  - internal/appid/appid_registry_test.go
  - internal/clipboard/clipboard.go
  - internal/clipboard/clipboard_test.go
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/config/load.go
  - internal/correct/stretch.go
  - internal/correct/stretch_test.go
  - internal/ctlsvc/ctlsvc.go
  - internal/ctlsvc/ctlsvc_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - docs/CONFIG.md
  - docs/adr/ADR-008-terminal-transport.md
  - docs/adr/adr-008-spike-log.md
  - skills/goswitch-config/SKILL.md
  - mise.toml
  - test/e2e/case_term_negative.go
  - test/e2e/case_term_word_vte.go
  - test/e2e/case_term_word_vte_flip.go
  - test/e2e/case_term_word_wezterm.go
  - test/e2e/fixtures/gs-e2e-wezterm-stand.lua
  - test/e2e/main.go
  - test/e2e/wezterm.go
covered_digest: "v1:sha256:69ba33a5875bbd892d71ae60fe6b80678335afcc02acd7c6d1535ed5b9b1030a"
behavior_unverified: 0
overrides_applied: 0
unverified_prohibitions: 21
unverified_prohibitions_note: "unverified-prohibition — human review recommended: 21 judgment-tier prohibitions (10-07 ×4, 10-08 ×4, 10-09 ×3, 10-10 ×3, 10-11 ×4, 10-12 ×3) carry non-authoritative LLM-judge verdicts in the Prohibition Review section; plans declare no verification tier — human review folds into the owner re-acceptance round-2"
re_verification:
  previous_status: human_needed
  previous_score: 16/17
  gaps_closed:
    - "G-10-1 round-2 причина routing: классификация терминала на хосте не включалась (identity недоступна, verifyDead не выучен) — оба жеста 09:39 уходили на уровень-1; 10-10 поднял предикат pendingStaleCache на ОБЕ ветки вердикта (match + mismatch): answer-only клиент маршрутизируется в терминальную ветку (код :2123-2141/:2149-2162, behavior-тесты ReplayHost0939* зелёные верификатором, RED честен)"
    - "G-10-1 round-2 причина delivery: «вижу привет, и он удаляется» = инверсия delete/commit на уровне-1; 10-11 построил двухфазный уровень-1 (delete → доказательство применения → commit, фолбэк по бюджету 60 мс), отложенный флип уровня-1 на машине 10-08 и WR-03 same-context флап-дисциплину (код :3737-3760/:1673-1691/:1216, behavior-тесты TwoPhaseLevel1*/FlapWindowContextIdentity*/Level1ConfirmFlush* зелёные верификатором, RED честен)"
    - "WR-01 (второй деструктивный бурст) и WR-02 (stale VerifyExpiry съедает чужой раунд) закрыты в зоне, которую маршрут делает горячей: level2Busy с тремя очистками (:3814-3865), pendingFix.epoch + armVerifyDeadline (:1559-1568), VerifyExpiry(epoch) no-op (:1484-1486) — behavior-тесты зелёные"
    - "Реплей-гейт реальных рядов 09:39 исполнен ДО стенда и установки (ряд repin-round2 в spike-log); хост несёт round-2 сборку v1.1.1-142-g2fbfee1 — верифицировано LIVE read-only этим верификатором (active, ExecStart, config_valid=true, счётчики на нуле, откат-каталог v1.1.1 нетронут)"
  gaps_remaining: []
  regressions: []
coincidental_reliance_items:
  - truth: "Ни один выход ладдера уровня-1 не оставляет поле без замены (окно никогда не умирает некоммитом) — ветка повторного армирования armLevel1Confirm при открытом окне останавливает бюджет вытесненного окна, не завершив ладдер (10-REVIEW IN-07); недостижимость ветки держится анализом графа вызовов, не по построению"
    reason: undeclared-precondition
    harden: "Одна строка из IN-07: в armLevel1Confirm перед a.l1.stop() звать a.flushLevel1Confirm() (идемпотентен по nil-check) — инвариант станет конструктивным; кандидат в следующий gap-цикл вместе с WR-05"
advisory:
  - finding: "WR-05 (10-REVIEW, warning): двухфазное окно уровня-1 отсрочило rng.replace, на которую опирался CR-01 тест устарелости — autoConfirm (единственный не-flush'ащий вход в startRangeCorrection) может армировать pending-раунд поверх незавершённой ладдера в RTT-окне AC-подтверждения; исход B — ложное обучение verifyDead честного GTK-контекста (класс последствий WR-02 через новый интерлив)"
    category: architectural
    reason: "Плин человека improbable (двойная тапка < 25 мс); плин машины реален только через программный ctl CorrectNow; коррупции текста нет ни на одной ветке (match требует реального token+tail, повторное исполнение идемпотентно) — new-scope, детерминированного отказа нет; fix задокументирован в 10-REVIEW (двухстрочный refusal в autoConfirm)"
    evidence_status: "none provided"
  - finding: "TestCorrectionBusy_SecondEntrySkipsNotSecondBurst (пин WR-01) разово FAIL в -race-батче верификатора под нагрузкой, затем 6/6 зелёных (изолированно ×5, повторный батч, полный mise run ci); механика теста — starvation-handoff >1 мс под мьютексом, чувствительна к планировщику"
    category: other
    reason: "Прод-гард level2Busy — обычный check-and-set под мьютексом, от таймингов не зависит; чувствительна только детерминированность ФИКСТУРЫ; candidate: увеличить порог ожидания второго входа или флейк-рetry в тесте"
    evidence_status: "один воспроизведённый верификатором FAIL при 6 последующих зелёных"
  - finding: "IN-08 (10-REVIEW, info): flapEng сравнивает interface-значения через == — паника на несравнимом динамическом типе; безопасно, пока все реализации engine.Emitter — указатели (проверено ревью grep'ом). IN-09 (info): flush раздвинул до-существующее окно устаревшего verify-after раунда поверх тихой смены движка — открыт только при opt-in ClipboardRung и тихой смерти без lifecycle; WR-04: xrandr-проба нуджа e2e-харнесса мертва (факт харнесса, на round-2 не повлияла)"
    category: other
    reason: "Консультативные находки ревью round-2 дельты (0 critical, 1 warning, 3 info), задокументированы в закоммиченном 10-REVIEW.md; верификацию не блокируют"
    evidence_status: "анализ ревью; живых отказов нет"
human_verification:
  - test: "Owner re-acceptance round-2 (закрывающий оракул G-10-1; сборка v1.1.1-142-g2fbfee1 установлена и активна с 12:55:53 MSK — верифицировано live): (1) wezterm: «ghbdtn» + двойной правый Shift → «привет», слово ОСТАЁТСЯ (не «появляется и удаляется» — дословный отказ 09:39); (2) GNOME Terminal: то же; (3) флап-бурст — быстро пощёлкать фокус между окнами и СРАЗУ жест в терминале (в 09:39 владелец пункт пропустил — при новом отказе от пункта записать skipped-by-owner, не failed); (4) слово в середине фразы (рецепт: «ghbdtn » + хвостовое слово текущей раскладкой, курсор в конце); (5) «ytn?» + жест → «нет,»; (6) автокоррекция молчала (машинно: /home/nil/.local/share/goswitch/releases/v1.1.1-142-g2fbfee1/goswitchctl status — autocorrect_fired не вырос; база чистая — счётчики на нуле). После жестов снять journalctl --user -u goswitchd за окно приёмки и сверить с критериями round-2 (DEBUG-term-host-live-failures.md): НЕТ verify-timeout/verify-mismatch; done changed=true не потерян; НЕТ инверсии A2 — первый surrounding-push после коррекции трекает исправленное слово; флип-рубежи строго после settlement-рубежей; мгновенный «mode +0,03 мс» формы 09:39 более невозможен. Ответ «approved» (дописать ряд «gap closed by owner re-acceptance round-2, дата» в round-2 блок) или описание отклонений (rejected → СТОП без правок кода вне нового цикла планирования; следующий цикл обязан реплеем НОВЫХ рядов доказать фикс до owner-запуска). Откат по запросу: ~/.local/share/goswitch/releases/v1.1.1/goswitchctl install"
    expected: "Все шесть пунктов зелёные на живом столе владельца; оба классифицированных отказа round-1 (маршрут не включался; инверсия доставки) не воспроизводятся — по построению закрыты 10-10/10-11 и доказаны реплеем рядов 09:39"
    why_human: "Стенд принципиально не воспроизводит хостовую среду (флап-тряска ядра 7.2.9-2-t2, реальный порядок клиентской обработки — round-1 доказал: полный машинный зелёный коридор дважды не предсказал хост); владелец — единственный валидный оракул закрытия G-10-1 (10-12 Task 3 human-check, T-10-12-01/03); вердикт честно ОЖИДАНИЕ в round-2 блоке Closure validation, не синтезируется"
  - test: "Формальная классификация края TERM-02 (флагнутое допущение, UAT Test 2 — [pending], продлён раундом 2): рассмотреть край-свидетельства — 8 золотых поведений пробега + живые ряды всех трёх раундов (bash/zsh, мид-лайн, clip-frozen, свидетели vim/htop, wezterm ×3 раунда re-pin, flip-ряды с оракулом порядка, STANDALONE-негатив) + round-2 дополнения (answer-only маршрут, two-phase доставка, инвертирующий потребитель) — и вынести вердикт классификации края"
    expected: "Владелец принимает край в том виде, в каком он выражен конкретными проверками; human_needed-вердикт требовался бы, только если бы край НЕ был так выражен"
    why_human: "Spec-less проба классифицировала край unclassified по построению (SPEC без ## Edge Coverage, все 12 планов); флагнутое допущение никогда не снимается автоматически — формальный вердикт за владельцем"
---

# Phase 10: TERM-слово Verification Report (re-verification round 3 — после gap-цикла 2)

**Phase Goal:** Пользователь исправляет слово, набранное не в той раскладке, одним жестом в wezterm и GNOME Terminal — через собственный канал ввода демона (IBus), без внешних инжекторов
**Verified:** 2026-10-14T13:10:58Z
**Status:** human_needed (машинный коридор round-2 полностью зелёный — обе причины отказа 09:39 закрыты в коде с behavior-доказательствами, реплей-гейт исполнен, хост несёт round-2 сборку; открыт вердикт владельца — см. Human Verification Required)
**Re-verification:** Yes — раунд 3, после gap-цикла 2 (планы 10-10/10-11/10-12; предыстория: раунд 1 pre-gap, раунд 2 post-цикла-1 с owner REJECTED 09:39)

## Goal Achievement

### Observable Truths

Truths 1–5 — контракт ROADMAP (регрессия). G1–G12 — must_haves планов закрытия раунда 1 (регрессия). R1–R5 — must_haves 10-10. D1–D6 — must_haves 10-11. H1–H5 — must_haves 10-12 (полная трёхуровневая проверка нового; регрессионная — ранее пройденного).

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Двойной правый Shift исправляет слово в wezterm (bash/zsh, кириллица, мид-лайн): ghbdtn → привет на живом столе | ✓ VERIFIED (регрессия) | Полный CI зелёный на HEAD (верификатором); регрессионный батч 27 пинов зелёный -race; ряды round-2 wezterm word-a/word-b 09:48:38/40 pty-дельта байт-точная; хост на v1.1.1-142-g2fbfee1. Приёмка владельца — Human Verification 1 |
| 2 | То же в GNOME Terminal (VTE); чтение — собственный буфер демона; primary-лег — по необходимости (OQ 2 «contract») | ✓ VERIFIED (регрессия) | Ряды round-2 VTE word-bash/zsh 09:45:44/47 — pty_hex=d0bfd180d0b8d0b2d0b5d1820a, верификатором независимо декодировано «привет\n»; буферный канал ProcessKeyEvent → ConvertibleStretch нетронут (полный корпус зелёный) |
| 3 | N×BackSpace-реплей + commit_text через IBus; CLIPBOARD байт-нетронут; commit-over-selection и paste-чорды не используются | ✓ VERIFIED (регрессия) | executeLevel2: только ForwardKeyEvent(BackSpace) чанки + один CommitText (:3837-3862); clip-frozen round-2 ряды 09:45:52 (VTE), 09:48:45 (wezterm), 09:46:12 (flip) — clip_before=clip_after=GS-SPIKE-SENTINEL во всех; paste-чорд недостижим (noRung-ветка, HonestPushStillFailClosed зелёный) |
| 4 | До первого BackSpace отмена бесплатна; после — verify-сверка логирует и останавливается, никогда не чинит; перечек фокуса между чанками ≤8 | ✓ VERIFIED (регрессия) | Пины FreeAbortBeforeBurst/FocusLossBetweenChunksStops/VerifyAfterNoRungOnMismatch зелёные (регрессионный батч); chunks=terminalChunkKeys (:3838); WR-01-гард добавил честный отказ второго входа (correction-busy) — усиление, не ослабление |
| 5 | Двухразрешительная модель Фазы 9 живая: автокоррекция молчит в терминале при работающей ручной (vim-insert/htop-свидетели) | ✓ VERIFIED (регрессия) | Вето на точке ARMING, шов classifyTerminal един (RouteLearnsVetoSeam зелёный); round-2 свидетели 09:49:13/16 — fired=0, veto Δ=1, буфер не изменён; STANDALONE autocorrect-terminal-silent ~09:51 PASS (fired=0, zero correction records) |
| G1 | Флап-пара FocusOut/FocusIn (1–3 мс) не стирает выученный терминальный класс (10-07) | ✓ VERIFIED (регрессия) | TestTerminalClass_VerifyDeadSurvivesFlapPair зелёный БЕЗ правок ожиданий (регрессионный батч); WR-03 добавил same-context различение — same-engine пара по-прежнему переживает (SameEnginePairSurvives зелёный) |
| G2 | Отсутствующий ключ correction.terminals даёт дефолт-якорь; явный [] — opt-out (10-07) | ✓ VERIFIED (регрессия) | config-корпус зелёный в полном CI; код нетронут round-2 (config.go/load.go вне дельты) |
| G3 | Недоступность a11y-роли растит ac_role_errors в status (10-07) | ✓ VERIFIED (регрессия) | Токен ac_role_errors= жив (round-1 хост показывал 3, блок 09:39 — 47); round-2 сборка: счётчик 0 на чистой базе — парсится; контроль ReplayHost0939_RejectionSignaturesAbsent пинит рост |
| G4 | Семантика остановки чанков (ctxEpoch) и жёсткий сброс на реальном FocusOut не тронуты (10-07) | ✓ VERIFIED (регрессия) | ctxEpoch-логика вне round-2 дельты; FlapPairBetweenChunksStops зелёный (регрессионный батч); полный session-корпус зелёный |
| G5 | Стенд не регрессирует: flip=false машина флипа не армится; кейсы re-pin зелёные (10-08) | ✓ VERIFIED (регрессия) | Ряды round-2: flip=false VTE-кейсы байт-точные 09:45:44-52; flip=true ряд delivery_order=done<settlement<mode; HotReloadFlipOn + FlipFalseArmsNothing зелёные |
| G6 | Пост-коррекционный флип не осиротяет in-flight commit: SetGlobalEngine только после подтверждения доставки (10-08) | ✓ VERIFIED (регрессия) | deferCorrectionFlip/releaseDeferredFlip нетронуты (код :1809-1860 прочитан верификатором); round-2 расширил машину на уровень-1 (flush :1688) — release-сайты те же; Flip* пины зелёные (регрессионный батч + полный CI) |
| G7 | Правило владельца «режим следует за скриптом результата» (10-08) | ✓ VERIFIED (регрессия) | releaseDeferredFlip переиспользует flipTo целиком; deferCorrectionFlip — гары comboPending/!FlipAfterCorrection/target==mode (:1821-1830); на уровне-1 правило теперь тоже отложено (D3) |
| G8 | Ряд term-word-vte-flip зелёный: pty байт-точный, порядок done → settlement → mode (10-08) | ✓ VERIFIED (регрессия) | Round-2 ряд flip-word-bash 09:46:09: pty_hex байт-точен + delivery_order=done<settlement<mode; flip-clip-frozen 09:46:12 — сентинел удержан |
| G9 | GTK level-1 путь и combo-флипы не тронуты (10-08) | ✓ VERIFIED (регрессия) | selection-путь байт-неизменен (executeSelectionCorrection :3771-3791, settleCorrectionFlip :3788); ComboFlipImmediateNotDeferred зелёный; честно-пушащие клиенты — байт-прежний мир (миграция honestPush — 28 пинов, ни одно ожидание не ослаблено; полный CI зелёный) |
| G10 | Хост несёт gap-сборку, unit active, откат-путь назван (10-09 → обновлено round-2) | ✓ VERIFIED | LIVE read-only верификатором 2026-10-14: unit active, ExecStart=…/releases/v1.1.1-142-g2fbfee1/goswitchd (старт 12:55:53 MSK), status парсится config_valid=true, счётчики на нуле; каталоги v1.1.1 (откат, Oct 8) и v1.1.1-124-g7ace3fd нетронуты (ls mtime); ряд host-install-gap-round2 12:55 MSK в spike-log |
| G11 | Секция Closure validation в debug-файле (10-09) | ✓ VERIFIED (регрессия) | Блок 09:39 REJECTED сохранен дословно (:343-395); round-2 блок дописан после него (:397+) |
| G12 | Разрыв G-10-1 закрывается ТОЛЬКО ответом владельца (10-09) | → HUMAN | Раунд-1 форма оракула исполнена: вердикт REJECTED 09:39 записан дословно, СТОП соблюдён (правки round-2 — отдельный санкционированный gap-цикл). Открытое закрытие переносится на round-2 форму оракула — truth H5 |
| R1 | Маршрутизация answer-only клиента достигает терминальной ветки на ОБЕИХ ветках вердикта; класс учится setVerifyDead; вето D-55 следует за тем же швом (10-10) | ✓ VERIFIED | Код: match-ветка :2123-2141 и mismatch-ветка :2149-2162 — байт-та же форма (resolvePending → setVerifyDead(true) → executeLevel2(..., true)); предикат :2205-2207 (sel==nil && !acBoundary && cacheNil). Behavior-тесты зелёные верификатором (-race): TestReplayHost0939_MatchAnswerRoutesTerminalBranch (ops: burst, ноль delete+commit уровня-1, ноль mode до подтверждения), TestReplayHost0939_RouteLearnsVetoSeam (boundary после маршрута → acReasonTerminalVeto) |
| R2 | Клиенты с живым пуш-каналом (cacheNil=false) маршрутизация не меняет; selection/boundary исключены (10-10) | ✓ VERIFIED | TestReplayHost0939_SpontaneousPushClientNotRouted зелёный (был зелёным и в RED — защитный пин); TestTerminalBranch_HonestPushStillFailClosed зелёный; исключения внутри предиката (:2206); миграция 18 nil-cache пинов сохранила ожидания (полный CI зелёный) |
| R3 | WR-02: VerifyExpiry не съедает чужой раунд — epoch-тег, устаревший колбэк no-op (10-10) | ✓ VERIFIED | Код: pendingFix.epoch (:247-251), armVerifyDeadline — тег+захват в одной критической секции (:1559-1568), VerifyExpiry(epoch) no-op при чужой эпохе (:1484-1486). Behavior-тест TestVerifyExpiryStaleRound_DoesNotEatNextRound зелёный верификатором (сценарий 10-REVIEW WR-02 на timerRecorder, no-op-stop = продакшн Stop-после-fire); teachVerifyDead-корпус зелёный без правок (живая эпоха совпадает) |
| R4 | WR-01: вторая коррекция не вклинивает второй деструктивный бурст — level2Busy до первого форварда, очистка на каждом выходе, повторный вход — skipCorrection("correction-busy") + settleCombo (10-10) | ✓ VERIFIED | Код: check-and-set до BuildPlan (:3814-3820), очистки на всех трёх выходах (:3826/:3855/:3865). Behavior-тест TestCorrectionBusy_SecondEntrySkipsNotSecondBurst зелёный (RED: 40 forwards/2 commits — точный сценарий второго бурста); ОГОВОРКА (advisory): разовый FAIL под нагрузкой -race с 6/6 последующими зелёными — чувствительность ФИКСТУРЫ (starvation-handoff >1 мс), прод-гард мьютексный; см. Advisory |
| R5 | Replay-тест реальных рядов 09:39 КРАСНЫЙ до правок и ЗЕЛЁНЫЙ после (10-10) | ✓ VERIFIED | evidence-20261014-10-10-red-task1.json: «not ok» ×4 целевых (MatchAnswer/RejectionSignatures/VerifyExpiryStale/CorrectionBusy) при защитном пине SpontaneousPush зелёном — проверено верификатором парсингом; на HEAD все зелёные (прогон верификатором). Реплей пинит подпись блока 09:39 (role-unknown, appid unknown, nil-cache, свежий match; НЕТ verify-timeout/мismatch и флап-стираний — блок 09:39, не 16:20) |
| D1 | Level-1 безопасна по порядку: CommitText не идёт в провод до подтверждения удаления; инверсия A2 невоспроизводима по построению (10-11) | ✓ VERIFIED | Код: executeCorrection — только delete + armLevel1Confirm (:3737-3760), коммит существует ТОЛЬКО в flushLevel1Confirm (:1683) через захваченный w.eng. Behavior-тесты зелёные верификатором: TestTwoPhaseLevel1_InvertingConsumerKeepsReplacement (финальное состояние поля держит «привет» при произвольном порядке клиентской обработки; RED-причина «CommitText shares the delete's arrival batch» — точный класс A2), TestReplayHost0939_* (модель answer-only), модель B пушащего — реплей стендового мира в полном CI |
| D2 | Молчание клиента в пределах бюджета → деградация к нынешней паре с громким INFO-слагом; verify-after армится на каждом выходе (10-11) | ✓ VERIFIED | Код: level1ConfirmExpiry — slog.Info("correction confirm","outcome","timeout") + flush (:1650-1659); level1ConfirmBudget = 60 мс const :87 с коммент-происхождением (НЕ конфиг); flush армит armAfterVerify (:1690). TestTwoPhaseLevel1_SilentClientDegradesLoud зелёный (фолбэк-пара + слаг + verify-after) |
| D3 | Флип уровня-1 отложен, не отменён (машина 10-08); combo-путь и flip=false байт-прежние; selection-путь не тронут (10-11) | ✓ VERIFIED | Код: flush зовёт deferCorrectionFlip (:1688) — гары comboPending/!FlipAfterCorrection/target==mode (:1821-1824); settleCorrectionFlip остался на selection (:3788) и замороженном legacy level-2 (:3876, НЕ ТРОГАТЬ); deferCorrectionFlip — единственный флип flush-пути. Behavior-тесты: TestTwoPhaseLevel1_FlipDeferredUntilConfirmingPush (RED-причина «mode records right after the delete» — форма 09:39), TestTwoPhaseLevel1_FlipReleasedByNextKey (delete < commit < mode < commit(хвост)) — зелёные; ComboFlipImmediateNotDeferred + HotReloadFlipOn зелёные |
| D4 | Flush использует ЗАХВАЧЕННЫЙ снапшот эмиттера; flush после перечеканки не вливает текст в новое приложение (10-11) | ✓ VERIFIED | Код: w.eng.CommitText (:1683) — никогда текущий emitter(); 6 flush-сайтов верифицированы grep+чтением (HandleLifecycle :1160, фолбэк :1658, confirm-проверка :2114, feedKey :2901, startCorrection :3514, startPhraseCorrection :3616). TestLevel1ConfirmFlush_FocusOutUsesCapturedSnapshot зелёный (AttachEngine нового объекта между жестом и FocusOut → коммит на старом снапшоте, новый sink ноль вызовов) |
| D5 | WR-03: флап-окно различает контексты — FocusIn отменяет стирание ТОЛЬКО same-engine; реальный переключ окна внутри 10 мс стирает класс (10-11) | ✓ VERIFIED | Код: flapEng захват в armVerifyDeadWipe (:1579), гард в FocusIn (:1216 `a.flapEng != nil && a.flapEng == eng`), очистка в cancel/expiry (:1599/:1615). Behavior-тесты зелёные: TestFlapWindowContextIdentity_RealSwitchWipesClass (подмена эмиттера → окно не отменено, класс стёрт), _SameEnginePairSurvives (защитный); TestTerminalClass_VerifyDeadSurvivesFlapPair зелёный БЕЗ правок — 10-07 не откачен |
| D6 | Replay инвертирующего потребителя КРАСНЫЙ до правок и ЗЕЛЁНЫЙ после; реплей 09:39 (пушащий VTE-класс) зелёный (10-11) | ✓ VERIFIED | evidence-20261014-10-11-red-task1.json: «not ok» ×6 целевых при зелёном SameEnginePairSurvives — проверено верификатором; на HEAD все зелёные (прогон верификатором). Оговорка честности (SUMMARY): фикстура пинит ОДНУ подпись каждого класса — «оба жеста» покрыты идентичностью подписи, инвертирующий потребитель моделирует ОБА варианта механики одним инвариантом финального состояния |
| H1 | Стенд не регрессирует на коде round-2: четыре кейса байт-точные зелёные; mise run ci зелёный (10-12) | ✓ VERIFIED | Ряды round-2 spike-log (:380-412, прочитаны верификатором): VTE 09:45:44-52 + flip 09:46:09/12 + wezterm word-a/b 09:48:38/40 + свидетели 09:49:13-23 + STANDALONE ~09:51 — все pass с байт-точными hex (декод верификатором: «привет\n»/«echo привет\n»/«ghbdtn\n»); честные отказо-ряды заездов-1 (setxkbmap-env, stray-x) записаны дословно с повторными зелёными заездами; mise run ci зелёный на HEAD (верификатором, exit 0) |
| H2 | Replay-гейт исполнен ДО всякого owner-запуска (10-12) | ✓ VERIFIED | Ряд repin-round2 (:410): «реплей-гейт ReplayHost0939*/TwoPhaseLevel1* на HEAD (ok, -race) → re-sync зеркала → кейсы»; порядок в ряду и SUMMARY Task 1 (реплей первым шагом); верификатором реплей-корпус зелёный на HEAD; код устанавливаемой сборки = HEAD-код без .go-дрифта (git diff 2fbfee1..HEAD -- '*.go' пуст, 2fbfee1 — предок HEAD) |
| H3 | Хост несёт gap-сборку round-2 (новый release-dir, откат назван ДО приёмки), предпроверки read-only зелёные (10-12) | ✓ VERIFIED | LIVE read-only верификатором: active; ExecStart=…/v1.1.1-142-g2fbfee1/goswitchd; старт 12:55:53 MSK; status по фикс. пути: config_valid=true, config_error= (пуст), autocorrect_fired=0, corrections_done=0, ac_role_errors=0 (чистая база); откат-каталог v1.1.1 нетронут (mtime Oct 8); ряд host-install-gap-round2 с откат-путём в spike-log |
| H4 | Round-2 блок Closure validation: критерии с anti-инверсионным пунктом + шестипунктовая инструкция (10-12) | ✓ VERIFIED | DEBUG-term-host-live-failures.md:397-520 прочитано верификатором: сборка/фикс. путь/откат (:404-420), предпроверки 12:55-12:58 (:422-435), критерии с «НЕТ инверсии A2 — первый push трекает исправленное слово» и «мгновенный mode +0,03 мс более невозможен» (:437-461), инструкция 6 пунктов (:463-485), доставка 13:00 MSK (:501+), вердикт — честное ОЖИДАНИЕ (:489-499, не синтезирован — проверено чтением) |
| H5 | Вердикт владельца — единственный оракул закрытия G-10-1: approved → ряд закрытия; rejected → честные ряды, СТОП (10-12 Task 3) | → HUMAN | By design: закрывающий оракул — жест владельца на живом столе; машинные проверки его не заменяют (T-10-12-01/03; human_judgment: true в coverage D4 10-12). Вердикт честно ОЖИДАНИЕ с 13:00 MSK. Единственная не-машинная истина round-2; та же единственная причина статуса human_needed, что и G12 round-1 (один оракул в двух формулировках) |

**Score:** 31/33 истин verified (G12 и H5 — by-design человеческий оракул G-10-1 в двух раундовых формулировках, честно ОЖИДАНИЕ; 0 present-behavior-unverified)

**Behavior-доказательства round-2:** все behavior-зависимые истины (R1 маршрут-переход на обеих ветках, R3 stale-переход эпохи, R4 гонка второго входа, D1 двухфазная доставка под инвертирующим потребителем, D3 отложенный флип, D4 flush-снапшот, D5 WR-03 переход) закреплены ИМЕНОВАННЫМИ тестами, исполненными верификатором с -race; RED-фазы обоих кодовых планов подтверждены по evidence-JSON («not ok» целевых тестов при зелёных защитных пинах).

### Deferred Items

Нет. Пункты, не покрытые кодом (вердикт владельца round-2, формальная классификация края), — человеческие оракулы, а не отложенные машинные работы; поздние фазы милстоуна (MENU-01, TERM-03, A11Y-*) их не покрывают.

### Advisory (New Scope, Unevidenced)

Re-verification ran — секция обязательна.

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| 1 | WR-05 (10-REVIEW warning): autoConfirm может армировать AC-pending поверх незавершённой ладдера уровня-1 в RTT-окне подтверждения — ложное обучение verifyDead (класс последствий WR-02) через новый интерлив; машина-плин только через программный CorrectNow | architectural | new-scope, детерминированного отказа нет (плин человека improbable <25 мс; коррупции текста нет ни на одной ветке — match идемпотентен); fix задокументирован ревью, кандидат в следующий цикл; виден владельцу до приёмки |
| 2 | Пин WR-01 (TestCorrectionBusy) разово FAIL в -race-батче верификатора под нагрузкой; 6/6 последующих зелёных + полный CI зелёный | other | Чувствительность ФИКСТУРЫ (starvation-handoff >1 мс), не прод-гарда (мьютексный check-and-set); свидетельство one-off, воспроизводимость не показана |
| 3 | IN-07/IN-08/IN-09 (10-REVIEW info) + WR-04 (мертвая xrandr-проба харнесса) | other | Консультативные находки ревью (0 critical), закоммичены в 10-REVIEW.md; верификацию не блокируют |

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/session/actor.go` (маршрут answer-only) | pendingStaleCache на обеих ветках HandleSurroundingText | ✓ VERIFIED | :2123-2141 (match), :2149-2162 (mismatch), предикат :2205-2207; wired через handleSurroundingLocked |
| `internal/session/actor.go` (WR-02 epoch) | pendingFix.epoch + armVerifyDeadline + VerifyExpiry(epoch) no-op | ✓ VERIFIED | :247-251, :1559-1568, :1478-1498; обе arm-точки (:3603/:3716) через шов |
| `internal/session/actor.go` (WR-01 guard) | level2Busy check-and-set + три очистки + correction-busy | ✓ VERIFIED | :3814-3865; wired до BuildPlan/первого форварда |
| `internal/session/actor.go` (машина level1Confirm) | budget const, arm/flush/expiry, 6 flush-сайтов, снапшот эмиттера | ✓ VERIFIED | :75-87 (const+происхождение), :1628-1691; сайты :1160/:1658/:2114/:2901/:3514/:3616; executeCorrection delete→arm :3737-3760 |
| `internal/session/actor.go` (WR-03) | flapEng захват/гард/очистка | ✓ VERIFIED | :1579 (захват), :1216 (гард FocusIn), :1599/:1615 (очистка) |
| `internal/session/actor_test.go` | ReplayHost0939×4 + RouteLearnsVetoSeam + VerifyExpiryStaleRound + CorrectionBusy + TwoPhase×4 + FlapWindowIdentity×2 + Level1ConfirmFlush | ✓ VERIFIED | 27 именованных тестов перечислены -list и исполнены верификатором зелёными -race |
| `docs/adr/ADR-008-terminal-transport.md` | два амендмента 2026-10-14 (10-10 answer-only; 10-11 three-part) | ✓ VERIFIED | :225+ и :280+; Evidence — ряды 09:39 и класс A2; ноль http-упоминаний (grep); D1–D7 не переписаны (аддитивные diff по саммари и ревью) |
| `docs/adr/adr-008-spike-log.md` | секция re-pin round-2 + ряд host-install-gap-round2 | ✓ VERIFIED | :380-412; 20 датированных рядов; честные FAIL/aborted сохранены |
| `.planning/debug/DEBUG-term-host-live-failures.md` | round-2 блок Closure validation | ✓ VERIFIED | :397-520; вердикт ОЖИДАНИЕ не синтезирован |
| `.planning/phases/10-term-slovo/evidence-20261014-10-1{0,1}-red-task1.json` | 10 честных RED-свидетельств (4+6) | ✓ VERIFIED | Парсинг верификатором: «not ok» целевых тестов, защитные пины зелёные в обоих |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | -- | --- | ------ | ------- |
| HandleSurroundingText pending-switch (обе ветки вердикта) | executeLevel2 + setVerifyDead(true) | pendingStaleCache — один предикат гейтит обе ветки | ✓ WIRED | :2123/:2149; форма байт-та же; RouteLearnsVetoSeam пинит единство шва D-55 |
| VerifyExpiry(epoch) | pendingFix.epoch + arm-точки | armVerifyDeadline — тег+захват в одной критической секции | ✓ WIRED | :1559-1568; VerifyExpiryStaleRound пинит no-op чужой эпохи |
| executeCorrection (delete) | level1Confirm + handleSurroundingLocked (confirm-проверка) | Require-доказательство/фолбэк/flush-сайты | ✓ WIRED | :3750-3759 arm; confirm :2111-2118; 6 flush-сайтов; flushLevel1Confirm — единственная точка завершения |
| executeCorrection settleCorrectionFlip (старый сайт) | deferCorrectionFlip (машина 10-08) | flush — второй потребитель, ноль новых release-сайтов | ✓ WIRED | :1688; release-сайты 10-08 нетронуты; FlipReleasedByNextKey на новом пути зелёный |
| HandleLifecycle FocusIn | flapEng-снапшот FocusOut | cancel ТОЛЬКО при равенстве объектов движка | ✓ WIRED | :1216; RealSwitchWipesClass/SameEnginePairSurvives пинят обе стороны |
| Реплей-корпус 10-10/10-11 | код устанавливаемой сборки | зелёный прогон на HEAD до установки | ✓ WIRED | Ряд repin-round2 (реплей первым шагом); .go-дрифта 2fbfee1..HEAD нет; верификатором зелёный |
| Хост unit goswitchd (round-2 сборка) | round-2 блок Closure validation | журнальные ряды жестов владельца (после приёмки) | ✓ WIRED (форма готова) | Unit на v1.1.1-142-g2fbfee1 верифицирован live; критерии рядов с anti-инверсией сформулированы; сами ряды — после жестов владельца |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| answer-only маршрут | cacheNil-предикат | реальный канал пуш/ответ живого клиента | да — реплей реальных рядов 09:39 зелёный | ✓ FLOWING |
| машина level1Confirm | match = concatRunes(token, tail) | реальный конвертируемый пробег раунда | да — стендовые ряды round-2 байт-точные | ✓ FLOWING |
| flapEng | снапшот emitter() момента FocusOut | живой объект движка контекста | да — RealSwitchWipesClass на подмене AttachEngine | ✓ FLOWING |
| ac_role_errors | счётчик ошибок role-источника | реальный демона хоста (round-1: 3, блок 09:39: 47) | да | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Реплей 09:39 (маршрут + подписи отказов + вето-шов) | `go test ./internal/session/ -race -run 'ReplayHost0939' -v` | 4/4 PASS | ✓ PASS |
| Двухфазный уровень-1 (инвертор/фолбэк/флип) + WR-03 + flush | `go test ./internal/session/ -race -run 'TwoPhaseLevel1\|FlapWindowContextIdentity\|Level1ConfirmFlush' -v` | 7/7 PASS | ✓ PASS |
| WR-02 stale-эпоха | `go test ./internal/session/ -race -run TestVerifyExpiryStaleRound -v` | PASS | ✓ PASS |
| WR-01 второй вход | `go test ./internal/session/ -race -run TestCorrectionBusy -v` | PASS (после разового флака под нагрузкой — см. Advisory; 6/6 зелёных) | ✓ PASS |
| Регрессия 10-07/10-08 (класс/флип/вето/fail-closed/abort) | `go test ./internal/session/ -race -run 'TestTerminalClass_\|TestTerminalBranch_StaleCache\|HonestPush\|ComboFlip\|TerminalVeto\|FreeAbort\|FocusLoss'` | 27/27 PASS | ✓ PASS |
| Полный CI-гейт | `mise run ci` | exit 0; lint 0 issues; все пакеты ok под -race (session 10.8s) | ✓ PASS |
| Заморозка зависимостей | `git diff --exit-code -- go.mod go.sum` | чисто | ✓ PASS |
| Хост: round-2 сборка активна, база чистая (read-only) | `systemctl --user is-active goswitchd` + ExecStart + фикс.ctl status | active; v1.1.1-142-g2fbfee1; старт 12:55:53 MSK; config_valid=true; все счётчики 0 | ✓ PASS |
| Ряды стенда round-2 — независимый hex-декод | python3 bytes.fromhex(d0bfd180…0a / 6563686f20… / 67686264746e0a) | «привет\n» / «echo привет\n» / «ghbdtn\n» | ✓ PASS |
| RED-свидетельства round-2 | парсинг evidence-20261014-10-1{0,1}-red-task1.json | «not ok» ×4 + ×6 целевых; защитные пины зелёные | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| scripts/*/tests/probe-*.sh | find scripts -path '*/tests/probe-*.sh' | не существуют | N/A |
| Живые e2e-кейсы round-2 (term-word-vte/-flip/-wezterm/-negative-witnesses + STANDALONE silent) | `go run ./test/e2e -case …` (внутри ВМ) | исполнены исполнителем 2026-10-14 09:44-09:52 UTC (ряды в spike-log); верификатором НЕ перезапускались | SKIP (re-run) — мутация живого стенда (рестарт демона, окна, CLIPBOARD/gsettings); базис верификатора: датированные ряды + независимый hex-декод + именованные юнит-тесты + live read-only хост-чеки |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| TERM-02 | 10-01…10-12 (все двенадцать; 10-10/10-11/10-12 frontmatter requirements: [TERM-02]) | Слово не в той раскладке исправляется жестом в wezterm и GNOME Terminal; чтение — буфер демона / primary (по необходимости); запись — N×BackSpace + commit_text через IBus; безопасная точка отмены; verify-сверка | ✓ SATISFIED (машинно; финальная приёмка round-2 — за владельцем) | Истины 1-5, G1-G11, R1-R5, D1-D6, H1-H4; REQUIREMENTS.md checkbox [x], трассировка «TERM-02 \| Phase 10 \| Complete»; оба класса отказа 09:39 (маршрут, инверсия доставки) закрыты в коде с behavior-доказательствами, реплеем реальных рядов и байт-точным стендом; host re-acceptance round-2 — Human Verification 1 |
| — (orphaned) | — | REQUIREMENTS.md отображает на Phase 10 только TERM-02 | ✓ NONE | Потерянных требований нет |

### Prohibition Review (non-authoritative LLM-judge — human review recommended)

Планы round-2 декларируют 10 новых prohibitions категории safety/values/privacy без tier `verification:` (10-10 ×3, 10-11 ×4, 10-12 ×3); суммарно с раундом 1 — 21. По ADR-550 D4 записываются НЕАВТОРИТЕТНЫЕ вердикты; человеческое ревью складывается в раунд owner re-acceptance round-2. Раунд-1 таблица (11 пунктов) переносится из предыдущего отчёта без изменений.

| # | Plan | Prohibition (кратко) | LLM-judge verdict | Basis |
| - | ---- | -------------------- | ----------------- | ----- |
| 12 | 10-10 | Маршрут не расходится с вето D-55: класс учится единственным швом setVerifyDead, оба потребителя classifyTerminal читают одно | judged PASS (non-authoritative) | RouteLearnsVetoSeam зелёный (верификатором); ревью подтвердило: arming-veto :3072-зона и write-branch читают один шов |
| 13 | 10-10 | Fail-closed мир кэш-свидетельств сохранён байт-в-байт; маршрут не трогает selection/boundary | judged PASS (non-authoritative) | HonestPushStillFailClosed + SpontaneousPushClientNotRouted зелёные (защитный пин был зелёным и в RED); исключения внутри предиката |
| 14 | 10-10 | Новые журнальные следы без содержимого; константы/флаги не конфиг | judged PASS (non-authoritative) | Слаг correction-busy — из закрытого словаря; epoch/busy — поля, не конфиг; go.mod нетронут |
| 15 | 10-11 | Ни один выход ладдера не оставляет поле без замены и без проверки | judged PASS with NOTE (non-authoritative) | Тесты SilentClientDegradesLoud/InvertingConsumer зелёные; NOTE: ветка повторного армирования armLevel1Confirm держит инвариант анализом графа вызовов (IN-07) — совпадающая coincidental_reliance запись |
| 16 | 10-11 | Flush-коммит ТОЛЬКО в захваченный снапшот, никогда в текущий emitter() | judged PASS (non-authoritative) | Код w.eng :1683; Level1ConfirmFlush_FocusOutUsesCapturedSnapshot зелёный (юнит-пин обязателен и стоит) |
| 17 | 10-11 | Правило владельца сохранено на уровне-1 (отложено, не отменено); combo/flip=false байт-прежние; 10-07 не откачен | judged PASS (non-authoritative) | deferCorrectionFlip-гары в коде; ComboFlipImmediateNotDeferred + HotReloadFlipOn + VerifyDeadSurvivesFlapPair зелёные без правок |
| 18 | 10-11 | Новые журнальные рубежи без содержимого; budget — не конфиг | judged PASS (non-authoritative) | level1ConfirmBudget const :87 с коммент-происхождением; слаги outcome-only (timeout/confirmed) |
| 19 | 10-12 | Хост read-only вне санкционированной установки | judged PASS (non-authoritative) | Ряд host-install-gap-round2 + предпроверки 12:55-12:58; независимая пост-проверка верификатора read-only (13:0x MSK); полная гарантируемость — за ревизией владельца |
| 20 | 10-12 | Откат-путь назван в ряду установки ДО приёмки | judged PASS (non-authoritative) | Ряд + инструкция + фикс. путь; каталог v1.1.1 нетронут (mtime Oct 8 — проверено верификатором) |
| 21 | 10-12 | Вердикт владельца не синтезируется | judged PASS (non-authoritative) | Round-2 блок: вердикт ОЖИДАНИЕ не синтезирован (проверено чтением); оговорка о grep-гейте записана честно (прецедент 10-09 D4) |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| internal/session/actor_test.go (TestCorrectionBusy_SecondEntrySkipsNotSecondBurst) | фикс-механика starvation-handoff | Разовый FAIL в -race-батче под нагрузкой (верификатор, 1 из 7 прогонов); 6/6 зелёных + полный CI зелёный | ⚠️ Warning | Тест-форма: детерминированность фикстуры зависит от планировщика; прод-гард мьютексный и от таймингов не зависит; candidate — увеличить порог/flye-ретрай |
| internal/session/actor.go | :1628-1631 | armLevel1Confirm повторного армирования останавливает бюджет вытесненного окна без завершения ладдера (IN-07) — ветка недостижима анализом графа вызовов, не по построению | ℹ️ Info (advisory) | Однострочный фикс из ревью (flush перед stop) сделает инвариант конструктивным; кандидат в следующий цикл с WR-05 |
| internal/session/actor.go | :1216 | flapEng interface-сравнение через == паникует на несравнимом динамическом типе (IN-08) — все живые реализации указатели | ℹ️ Info | Оборонительная заметка для будущего третьего Emitter |
| internal/session/actor.go | :1683-1690 | Flush раздвинул окно устаревшего verify-after поверх тихой смены движка (IN-09) — pre-existing класс, opt-in ClipboardRung + тихая смерть без lifecycle | ℹ️ Info | Записано для следующей ревизии машины флипа |
| docs/adr/adr-008-spike-log.md | :395 | Честные отказо-ряды заездов-1 round-2 (setxkbmap-env aborted; stray-x aborted с инъекционным мусором) — записаны дословно, вердикт по повторным зелёным заездам | ℹ️ Info | Честность ведения журнала; класс префиксного мусора repin-mechanics 10-08, не регрессия round-2 |
| internal/session/actor.go и все изменённые round-2 файлы | — | TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER | — | clean (0) |

### Test Quality Audit (delta round-2)

| Test File | Linked Req | Verdict |
| --------- | ---------- | ------- |
| actor_test.go: ReplayHost0939* (4) + RouteLearnsVetoSeam | TERM-02 | OK — RED честен (4 not ok при зелёном защитном пине); реплей пинит подпись блока 09:39 (не 16:20), тайминги журнала только в origin-комментах |
| actor_test.go: VerifyExpiryStaleRound + CorrectionBusy | TERM-02 | OK — сценарии 10-REVIEW WR-02/WR-01; no-op-стоп = продакшн Stop-после-fire; starvation-handoff чувствителен к нагрузке (см. Advisory) |
| actor_test.go: TwoPhaseLevel1* (4) + FlapWindowContextIdentity* (2) + Level1ConfirmFlush | TERM-02 | OK — RED честен (6 not ok); инвертирующий потребитель — оракул финального СОСТОЯНИЯ, не порядка эмиссии; подмена эмиттера — юнит-пин flush-безопасности |
| actor_test.go: миграция honestPush (28 пинов) | TERM-02 | OK — механически честна (ревью + полный CI): байт-семантика level-1 пинов сохранена добавлением пуша-свидетельства, не ослаблением ассертов |
| ctlsvc_test.go: TestSvc_CorrectNow (подтверждающий пуш) | TERM-02 | OK — смысл пина сохранён |

Disabled tests on requirements: 0. Circular patterns: 0 — стендовые hex-ряды независимо декодированы верификатором; реплеи пинят решения и порядок ops, не тайминги.

### Human Verification Required

### 1. Owner re-acceptance round-2 — закрывающий оракул G-10-1 (6 пунктов, сборка v1.1.1-142-g2fbfee1 активна)

**Test:** На живом столе владельца (round-2 сборка v1.1.1-142-g2fbfee1, старт 12:55:53 MSK — верифицировано live, счётчики на нуле): (1) wezterm: «ghbdtn» + двойной правый Shift → «привет», слово ОСТАЁТСЯ; (2) GNOME Terminal: то же; (3) флап-бурст — быстро пощёлкать фокус между окнами и СРАЗУ жест в терминале (в 09:39 владелец пункт пропустил — при новом отказе записать skipped-by-owner, не failed); (4) слово в середине фразы (рецепт: «ghbdtn » + хвостовое слово, курсор в конце); (5) «ytn?» + жест → «нет,»; (6) автокоррекция молчала (машинно: `/home/nil/.local/share/goswitch/releases/v1.1.1-142-g2fbfee1/goswitchctl status` — autocorrect_fired не вырос). После жестов снять `journalctl --user -u goswitchd` за окно приёмки и сверить с критериями round-2: НЕТ verify-timeout/verify-mismatch; done changed=true не потерян; НЕТ инверсии A2 (первый surrounding-push после коррекции трекает исправленное слово — не пустоту E3 и не старое слово); флип-рубежи строго после settlement-рубежей; мгновенный «mode +0,03 мс» формы 09:39 более невозможен. Ответ «approved» (дописать ряд «gap closed by owner re-acceptance round-2, дата») или описание отклонений (rejected → СТОП без правок кода вне нового цикла; следующий цикл обязан реплеем НОВЫХ рядов доказать фикс до owner-запуска). Откат по запросу: `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install`.
**Expected:** Все шесть пунктов зелёные; оба машинно диагностированных отказа round-1 (маршрут не включался — 10-10; «вижу привет, и он удаляется» — 10-11) не воспроизводятся.
**Why human:** Стенд принципиально не воспроизводит хостовую среду (флап-тряска ядра 7.2.9-2-t2; реальный порядок клиентской обработки — round-1 доказал, что полный машинный зелёный коридор дважды не предсказал хост); владелец — единственный валидный оракул (10-12 Task 3 human-check, T-10-12-01/03); вердикт честно ОЖИДАНИЕ, не синтезируется.

### 2. Формальная классификация края TERM-02 (флагнутое допущение; UAT Test 2 — pending)

**Test:** Рассмотреть край-свидетельства — 8 золотых поведений пробега + живые ряды всех трёх раундов (bash/zsh, мид-лайн, clip-frozen, свидетели vim/htop, wezterm ×3 re-pin, flip-ряды с оракулом порядка, STANDALONE-негатив) + round-2 дополнения (answer-only маршрут на обеих ветках, two-phase доставка, инвертирующий потребитель) — и вынести вердикт классификации края.
**Expected:** Край принят в том виде, в каком он выражен конкретными проверками; human_needed требовался бы, только если бы край НЕ был так выражен.
**Why human:** Spec-less проба классифицировала край unclassified по построению (SPEC без ## Edge Coverage, все 12 планов); флагнутое допущение никогда не снимается автоматически.

### Gaps Summary

Машинных разрывов нет. Обе причины owner-отказа 09:39 закрыты в коде round-2 и подкреплены behavior-доказательствами: **routing-половина (10-10)** — предикат pendingStaleCache поднят на обе ветки вердикта HandleSurroundingText: answer-only клиент (пустой кэш при арме — терминалы не пушат спонтанно) маршрутизируется в терминальную ветку и по свежему совпадающему ответу (точная подпись рядов 09:39), и по протухшему (10-08); класс учится единственным швом setVerifyDead, вето D-55 следует автоматически; WR-01/WR-02 закрыты в зоне, которую маршрут сделал горячей. **Delivery-половина (10-11)** — уровень-1 двухфазен: коммит не существует вне flushLevel1Confirm и уходит в провод только после доказательства применения удаления (пуш/ответ без token+tail на хвосте) либо по бюджету 60 мс с громким INFO-слагом (никогда «удаление без замены»); инверсия A2 («вижу привет, и он удаляется») невоспроизводима по построению — доказано реплеем инвертирующего потребителя; флип уровня-1 отложен машиной 10-08, flush живёт на захваченном снапшоте эмиттера, WR-03 различает контексты флап-пары.

Урок 2026-10-12/09:39 исполнен: реплей-гейт реальных рядов исполнен ДО стенда и установки (ряд repin-round2), стенд байт-точный зелёный ×4 + STANDALONE-негатив, хост несёт round-2 сборку v1.1.1-142-g2fbfee1 (верифицировано live read-only), `mise run ci` зелёный на HEAD, go.mod/go.sum нетронуты, .go-дрифта между установленной сборкой и HEAD нет.

Статус human_needed исключительно из-за честного ожидания владельца: re-acceptance round-2 (единственная by-design человеческая истина — G12/H5, один оракул в двух формулировках) и формальная классификация края. Advisory-хвосты (WR-05, IN-07…IN-09, разовый флак пина WR-01) видны владельцу в этом отчёте до приёмки; ни один не противоречит критериям успеха ROADMAP и не блокирует достижение цели.

---

_Verified: 2026-10-14T13:10:58Z_
_Verifier: Claude (gsd-verifier)_
