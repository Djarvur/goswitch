---
phase: 10-term-slovo
verified: 2026-10-13T22:30:04Z
status: human_needed
score: 16/17 must-haves verified
covered_files:
  - .planning/REQUIREMENTS.md
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
  - .planning/phases/10-term-slovo/10-UAT.md
  - docs/CONFIG.md
  - docs/adr/ADR-008-terminal-transport.md
  - docs/adr/adr-008-spike-log.md
  - internal/appid/appid.go
  - internal/appid/appid_registry_test.go
  - internal/clipboard/clipboard.go
  - internal/clipboard/clipboard_test.go
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/config/load.go
  - internal/ctlsvc/ctlsvc.go
  - internal/ctlsvc/ctlsvc_test.go
  - internal/correct/stretch.go
  - internal/correct/stretch_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - skills/goswitch-config/SKILL.md
  - test/e2e/case_term_negative.go
  - test/e2e/case_term_word_vte.go
  - test/e2e/case_term_word_vte_flip.go
  - test/e2e/case_term_word_wezterm.go
  - test/e2e/fixtures/gs-e2e-wezterm-stand.lua
  - test/e2e/main.go
  - test/e2e/wezterm.go
covered_digest: "v1:sha256:055056a1642d2ffa7e611270a7b274bbfecc52464cdcb437f33b4dc952e6503d"
behavior_unverified: 0
overrides_applied: 0
unverified_prohibitions: 11
unverified_prohibitions_note: "unverified-prohibition — human review recommended: 11 judgment-tier prohibitions (10-07 ×4, 10-08 ×4, 10-09 ×3) carry a non-authoritative LLM-judge verdict in the Prohibition Review section; plans declare no verification tier — human review folds into the owner re-acceptance round"
re_verification:
  previous_status: human_needed
  previous_score: 5/5
  gaps_closed:
    - "G-10-1 причина A: флап-стойкая классификация терминала (10-07) — код + behavior-тесты зелёные + стенд re-pin байт-точный"
    - "G-10-1 причина B: defer-until-confirmed доставка commit + stale-cache амендмент (10-08) — код + behavior-тесты зелёные + живой flip-ряд с оракулом порядка"
    - "G-10-1 недостающий элемент 5: gap-сборка v1.1.1-124-g7ace3fd на хосте, машинные предпроверки зелёные (10-09, верифицировано live read-only)"
  gaps_remaining: []
  regressions: []
human_verification:
  - test: "Owner re-acceptance (закрывающий оракул G-10-1; gap-сборка v1.1.1-124-g7ace3fd установлена и активна): (1) wezterm: «ghbdtn» + двойной правый Shift → «привет»; (2) GNOME Terminal: то же; (3) флап-бурст — быстро пощёлкать фокус между окнами и СРАЗУ жест в терминале — исправление срабатывает; (4) слово в середине фразы исправляется; (5) «ytn?» + жест → «нет,»; (6) автокоррекция молчала во время всех проб (goswitchctl status — autocorrect_fired не вырос). Ответ «approved» или описание отклонений; при approved — ряд «gap closed by owner re-acceptance» в Closure validation; при rejected — СТОП без правок кода. Откат по запросу: ~/.local/share/goswitch/releases/v1.1.1/goswitchctl install"
    expected: "Все шесть пунктов зелёные на живом столе владельца; журнальные ряды приёмки: нет skip-рядов verify-timeout/verify-mismatch, done changed=true не потерян по E3 (следующий push трекает исправленное слово), флип-рубежи строго после settlement-рубежей"
    why_human: "Стенд принципиально не воспроизводит хостовую флап-тряску ядра 7.2.9-2-t2 (E1/E2b доказали); владелец — единственный валидный оракул закрытия G-10-1 (план 10-09 Task 2 human-check, T-10-09-02); вердикт честно ОЖИДАНИЕ в Closure validation, не синтезируется"
  - test: "Формальная классификация края TERM-02 (флагнутое допущение, UAT Test 2 — [pending]): рассмотреть край-свидетельства — 8 золотых поведений пробега + живые ряды (bash/zsh, мид-лайн, clip-frozen, свидетели vim/htop, wezterm ×2 раунда 10-06 и re-pin 10-08 с learning=true по stale-cache амендменту, flip-ряд) — и вынести вердикт классификации края"
    expected: "Владелец принимает край в том виде, в каком он выражен конкретными проверками; human_needed-вердикт требовался бы, только если бы край НЕ был так выражен"
    why_human: "Spec-less проба классифицировала край unclassified по построению; флагнутое допущение никогда не снимается автоматически — формальный вердикт за владельцем"
---

# Phase 10: TERM-слово Verification Report (re-verification после закрытия G-10-1)

**Phase Goal:** Пользователь исправляет слово, набранное не в той раскладке, одним жестом в wezterm и GNOME Terminal — через собственный канал ввода демона (IBus), без внешних инжекторов
**Verified:** 2026-10-13T22:30:04Z
**Status:** human_needed (машинная верификация полностью зелёная, включая обе причины разрыва G-10-1; открыта ревизия владельца — см. Human Verification Required)
**Re-verification:** Yes — после закрытия разрыва G-10-1 (планы 10-07, 10-08, 10-09)

## Goal Achievement

### Observable Truths

Truths 1–5 — контракт ROADMAP (предыдущий раунд, регрессионная проверка). Truths G1–G12 — must_haves планов закрытия 10-07/10-08/10-09 (полная трёхуровневая проверка).

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Двойной правый Shift исправляет слово в wezterm (bash/zsh, кириллица, мид-лайн): ghbdtn → привет на живом столе | ✓ VERIFIED | Регрессия: символы целы (`case_term_word_wezterm.go`, терминальная ветка actor.go), корпус зелёный. Живые ряды: 20:34/21:17 (10-06) байт-точные; re-pin 10-08 — word-a 21:31:32 learning=true (stale-cache амендмент виден живым рядом), word-b/mid-line байт-точные. Хост-сборка v1.1.1-124-g7ace3fd активна. Приёмка владельца — Human Verification 1 |
| 2 | То же в GNOME Terminal (VTE); чтение — собственный буфер демона; primary-лег — по необходимости (вердикт OQ 2 «contract») | ✓ VERIFIED | Регрессия: re-pin ряды word-bash/word-zsh/mid-line 21:30:49-53 — pty_hex=d0bfd180d0b8d0b2d0b5d1820a дословно рядов 10-05 (верификатором декодировано «привет»); reading channel: ProcessKeyEvent → Buffer → ConvertibleStretch → actor.go:2889-2891 нетронут; leg-probe red честен (OQ 2 «contract») |
| 3 | N×BackSpace-реплей + commit_text через IBus; CLIPBOARD байт-нетронут; commit-over-selection и paste-чорды не используются | ✓ VERIFIED | executeLevel2: только ForwardKeyEvent(KeyBackSpace) чанки + один CommitText; clip-сентинел удержан во ВСЕХ записанных прогонах (clip_before=clip_after=GS-SPIKE-SENTINEL: 18:58:55, 21:30:56, 21:31:38, flip 21:32:57); paste-чорд недостижим (noRung-ветка, пин зелёный); uinput/wl-copy в daemon-пути нет |
| 4 | До первого BackSpace отмена бесплатна; после — verify-сверка логирует и останавливается, никогда не чинит; перечек фокуса между чанками ≤8 | ✓ VERIFIED | Поведенческие тесты зелёные (верификатором, -race): TestTerminalBranch_FreeAbortBeforeBurst, FocusLossBetweenChunksStops, VerifyAfterNoRungOnMismatch, FlapPairBetweenChunksStops (ctxEpoch-остановка переживает флап); код: aborts/forks actor.go:3384-3488+, chunks=8 (:61), ctxEpoch++ на каждом FocusOut (:1043) |
| 5 | Двухразрешительная модель Фазы 9 живая: автокоррекция молчит в терминале при работающей ручной (vim-insert/htop-свидетели) | ✓ VERIFIED | Вето на точке ARMING перед детектором: actor.go:2782-2788 (classifyTerminal → recordACAbstain(acReasonTerminalVeto) → detectVerdict); TestAutoCorrect_TerminalVeto* зелёные (верификатором); re-pin свидетели 21:32:13/27 — fired=0, veto Δ=1, буферы дословно; positive-control 21:32:29 байт-точный |
| G1 | Флап-пара FocusOut/FocusIn (1–3 мс) не стирает выученный терминальный класс: verifyDead переживает пару; стирают только истечение окна, Reset, синтетика флипа (10-07, причина A) | ✓ VERIFIED | Код: focusFlapWindow=10 мс (actor.go:73), FocusOut-ветка armVerifyDeadWipe / Reset+синтетика setVerifyDead(false) (:1047-1062), FocusIn cancelVerifyDeadWipe (:1080), один таймер с epoch-гардом (:1411-1444). Behavior-тест зелёный (верификатором): TestTerminalClass_VerifyDeadSurvivesFlapPair; RED честен (evidence-20261013-10-07-red-task1.json, «not ok 153»); пины ClearedBy*/Rearms/FlapPairBetweenChunksStops зелёные |
| G2 | Отсутствующий ключ correction.terminals даёт дефолт-якорь [gnome-terminal, wezterm]; явный [] — opt-out; docs фиксируют обе формы (10-07) | ✓ VERIFIED | config.go:91-128 (DefaultTerminals, константы с происхождением G-10-1), load.go:46 (нормализация nil→якорь после декода), Defaults() :285; docs/CONFIG.md:43,84 + SKILL.md:151 (регенерирован, golden зелёный). Тесты зелёные (верификатором): TestCorrection_TerminalsDecode (без ключа → якорь; [] → opt-out; список → как написано), TestDefaults_CorrectionTerminalsAnchor |
| G3 | Недоступность a11y-роли перестала быть тихой: каждая ошибка растит счётчик в status, WARN переармивается (10-07) | ✓ VERIFIED | actor.go:116-124 (acRoleWarnRearm 30 мин), :353-364 (acRoleErrors), :875-923 (Status.AutoCorrectRoleErrors), :3066 (здоровый ответ обнуляет); ctlsvc.go:196 токен ac_role_errors= (пин ctlsvc_test.go:816 зелёный, старые токены байт-прежние). Тесты зелёные (верификатором): RoleErrorsCounted, RoleWarnRearmsAfterQuietWindow, RoleHealthyClosesEpisodeAndZeroes. ЖИВОЕ доказательство: host status сейчас показывает ac_role_errors=3 — диагностика машинно видна на реальном столе |
| G4 | Семантика остановки чанков (ctxEpoch) и жёсткий сброс на реальном FocusOut не тронуты — защищённое поведение 10-04 байт-прежнее (10-07) | ✓ VERIFIED | Код: ctxEpoch++ ДО флап-ветки (:1043) — растёт на каждом FocusOut включая флап; HardReset/resolvePending/clearAfter следуют как прежде. Пин TestTerminalBranch_FlapPairBetweenChunksStops зелёный (верификатором); полный session-корпус зелёный в mise run ci (exit 0) |
| G5 | Стенд не регрессирует: при flip=false машина флипа не армится вовсе; term-word-vte, term-word-wezterm, term-negative-witnesses re-pin зелёные (10-08) | ✓ VERIFIED | Пин «flip=false → ops байт-идентичен» зелёный (саммари+CI); журнальные ряды 21:30-21:32 (свежая сессия boot 21:27:35, четыре кейса подряд): VTE 4/4 рядов байт-точных, свидетели fired=0/veto Δ=1, wezterm word-a/b/mid-line байт-точные. ОГОВОРКА (warning): wezterm clip-frozen 21:31:38 — ряд записан Verdict: FAIL (в pty-дельте префиксный мусор 16×0x6e и нет хвостового 0a), при этом инвариант ряда — CLIPBOARD-сентинел — удержан (clip_before=clip_after), «привет» в дельте присутствует, кейс exit 0 (форма: wezClipFrozenRow возвращает cerr захвата, а не результат сравнения — FAIL-ряд не валит exit). Не блокер: SC3-инвариант удержан, ряд записан честно, наблюдение самоисполнителем помечено на разбор фазовой верификацией (разобрано здесь) |
| G6 | Пост-коррекционный флип не осиротяет in-flight commit: SetGlobalEngine только после подтверждения доставки (push / клавиша / grace 300 мс) (10-08, причина B) | ✓ VERIFIED | Код: flipConfirmGrace=300 мс (actor.go:75-82), deferCorrectionFlip (:1580-1596), releaseDeferredFlip (:1599-1620), flipGraceExpiry с epoch-гардом (:1637-1644); release-сайты: match settleAfter :1970, mismatch :1984, тихий таймаут :2022, вход feedKey :2617 (СТРОГО до обработки ключа); НИ ОДНОГО release на FocusOut/Reset (проверено grep+чтением HandleLifecycle). Behavior-тесты зелёные (верификатором): FlipDeferredUntilConfirmed, FlipReleasedByNextKey, FlipGraceDeadlineFires, FlipSurvivesFocusOutGraceReleases, SecondCorrectionReleasesPendingFlip; RED честен (10-08-red-task1, «not ok 182»). Живой ряд flip-word-bash 21:32:54 — delivery_order=done<settlement<mode |
| G7 | Правило владельца «режим следует за скриптом результата» сохранено: флип отложен, не отменён; ровно один flipTo на коррекцию; D-36 порядок байт-прежний (10-08) | ✓ VERIFIED | releaseDeferredFlip переиспользует flipTo целиком (D-36 не дублируется); слот очищается при release; grace-дедлайн гарантирует освобождение. Тесты FlipGraceDeadlineFires (ровно один, stale-fire no-op) и SecondCorrectionReleasesPendingFlip зелёные (верификатором) |
| G8 | Новый ряд term-word-vte-flip (flip=true) зелёный: pty байт-точный, журнальный порядок done → settlement → mode/switch (10-08) | ✓ VERIFIED | case_term_word_vte_flip.go (собственный конфиг-документ flip_after_correction: true), зарегистрирован main.go:131/278/340, mise-задача e2e-term-word-vte-flip (mise.toml:190-192, NOT in ci). Живые ряды: flip-word-bash 21:32:54 pty_hex байт-точен + delivery_order=done<settlement<mode; flip-clip-frozen 21:32:57 — сентинел удержан при активной машине флипа |
| G9 | GTK level-1 путь и combo-флипы не тронуты — немедленный settleCorrectionFlip; mismatch verify-after только логирует и останавливает (10-08) | ✓ VERIFIED | executeCorrection level-1: немедленные флипы :3450/:3481; терминальный форк :3553-3555 (deferCorrectionFlip только в терминальной ветке); combo-гард в deferCorrectionFlip; пины TestTerminalBranch_ComboFlipImmediateNotDeferred и мигрированные VerifyPaths/RefusalsNeverFlip зелёные (верификатором, включая HonestPushStillFailClosed — fail-closed для кэш-свидетельств сохранён) |
| G10 | Хост несёт gap-сборку (оба фикса), unit active, версия/коммит записаны, откат-путь назван; машинная предпроверка зелёная (10-09) | ✓ VERIFIED | Верифицировано LIVE read-only этим верификатором: `systemctl --user is-active goswitchd` = active, ExecStart=/home/nil/.local/share/goswitch/releases/v1.1.1-124-g7ace3fd/goswitchd (старт 2026-10-14 00:46:57 MSK); фикс. путь ctl существует; status парсится: config_valid=true, config_error пуст, ac_role_errors виден (сейчас 3), autocorrect_fired=0. Ряд host-install-gap 2026-10-14 00:46 в spike-log с маркером «gap G-10-1 closure build» и откатом ДО приёмки |
| G11 | Секция Closure validation (G-10-1) в debug-файле: машинные предпроверки, шестипунктовая инструкция re-acceptance, критерии журнальных рядов (10-09) | ✓ VERIFIED | .planning/debug/DEBUG-term-host-live-failures.md:259-338 — предпроверки 00:46–00:49 MSK с датами; инструкция 6 пунктов (5 исходных + флап-бурст + молчание автокоррекции с машинной проверкой); критерии: нет verify-timeout/mismatch, done changed=true не потерян по E3, флипы строго после settlement; откат в инструкции; вердикт — честный ОЖИДАНИЕ (не синтезирован — проверено чтением) |
| G12 | Разрыв G-10-1 закрывается ТОЛЬКО ответом владельца: re-acceptance на живом столе — «approved» (10-09 Task 2) | → HUMAN | By design: закрывающий оракул — жест владельца; машинные проверки его не заменяют (T-10-09-02). Вердикт честно ОЖИДАНИЕ в Closure validation → Human Verification item 1. Единственная не-машинная истина фазы; единственная причина статуса human_needed |

**Score:** 16/17 истин verified (G12 — by-design человеческий оракул, честно ОЖИДАНИЕ; 0 present-behavior-unverified)

**Behavior-доказательства:** все behavior-зависимые истины (G1 флап-переход, G6/G7 машина флипа, G9 fail-closed, G4 ctxEpoch, 3/5 вето/abort) закреплены ИМЕНОВАННЫМИ тестами, исполненными верификатором с -race; RED-фазы обоих TDD-планов подтверждены по evidence-JSON («not ok» целевых тестов при exit 1).

### Deferred Items

Нет. Пункты, не покрытые кодом (вердикт владельца, классификация края), — человеческие оракулы, а не отложенные машинные работы; поздние фазы милстоуна (MENU-01, TERM-03, A11Y-*) их не покрывают.

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/session/actor.go` (флап-механика) | focusFlapWindow + armVerifyDeadWipe/cancelVerifyDeadWipe/verifyDeadWipeExpiry, UseFlapTimer | ✓ VERIFIED | :63-73, :1047-1090, :1411-1444; один таймер, epoch-гард, переармывание; wired в HandleLifecycle |
| `internal/session/actor.go` (машина флипа) | deferCorrectionFlip/releaseDeferredFlip/flipGraceExpiry, flipConfirmGrace, pendingStaleCache | ✓ VERIFIED | :75-82, :340-346, :1569-1644, :1912-1924, :3403; release-сайты :1970/:1984/:2022/:2617; арм :3553 |
| `internal/session/actor.go` (role-диагностика) | acRoleErrors + acRoleWarnRearm + roleClock + Status-поле | ✓ VERIFIED | :116-124, :353-364, :800-807, :875-923, :3066 |
| `internal/config/config.go` + `load.go` | DefaultTerminals + якорь в Defaults + нормализация nil→якорь | ✓ VERIFIED | config.go:91-128, :285; load.go:46; docs/CONFIG.md:43,84 и SKILL.md:151 синхронны (golden зелёный) |
| `internal/ctlsvc/ctlsvc.go` | токен ac_role_errors= в ac-блоке | ✓ VERIFIED | :196; пин ctlsvc_test.go:816; старые ключи байт-прежние; ЖИВО на хосте (ac_role_errors=3) |
| `test/e2e/case_term_word_vte_flip.go` | кейс term-word-vte-flip: оракул порядка + CLIPBOARD-сентинел | ✓ VERIFIED | 12 885 байт; главный real-data оракул; зарегистрирован main.go:340 + usage :131 + mise.toml:190 |
| `docs/adr/ADR-008-terminal-transport.md` | амендмент 2026-10-13: delivery-order safety + клиент-дрейф | ✓ VERIFIED | Секция :159+ (defer-until-confirmed, flipConfirmGrace 300 мс, E3-вердикт как Evidence); Status Accepted; D1–D7 не переписаны; веб-ссылок нет |
| `docs/adr/adr-008-spike-log.md` | ряды re-pin 10-07, flip 10-08, дрейф/амендрмент, host-install-gap 10-09 | ✓ VERIFIED | Секции :288+ (re-pin 18:58-19:08), flip :21:32:54/57, дрейф+checkpoint владельца (вариант 1 одобрен дословно), host-install-gap 2026-10-14 00:46; FAIL/aborted/blocked-ряды сохранены честно |
| `.planning/debug/DEBUG-term-host-live-failures.md` | секция Closure validation (G-10-1) | ✓ VERIFIED | :259-338; вердикт ОЖИДАНИЕ не синтезирован |
| `.planning/phases/10-term-slovo/evidence-20261013-10-0{7,8}-red-task*.json` | 5 честных RED-свидетельств | ✓ VERIFIED | Все 5 содержат «not ok N - <целевой тест>» при exit 1 (проверено парсингом) |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | -- | --- | ------ | ------- |
| actor.go HandleLifecycle FocusOut (:900-зона) | setVerifyDead / flap-механика | отложенное стирание focusFlapWindow; FocusIn отменяет | ✓ WIRED | :1047-1090; Reset/синтетика — немедленно; ctxEpoch++ вне ветки |
| config Load/Defaults (Correction.Terminals) | actor classifyTerminal | дефолт-якорь через существующий applySnapshot-фолд | ✓ WIRED | load.go:46; якорь-тест classifyTerminal зелёный; watch.go гонит reload через Load (ревью подтвердило) |
| actor warnACRole (acRoleErrors) | ctlsvc StatusSnapshot | счётчик виден машинно | ✓ WIRED | actor.go:875-923 → ctlsvc.go:196; ЖИВО: host status ac_role_errors=3 |
| actor executeLevel2 (коммит) | settleAfter / afterExpiry / feedKey | арм отложенного флипа; release на первом подтверждении | ✓ WIRED | :3553 арм; release :1970/:1984/:2022/:2617; ни одного release на lifecycle |
| case_term_word_vte_flip.go | машина отложенного флипа | конфиг flip=true армит; журнальный порядок оракулятся | ✓ WIRED | живой ряд delivery_order=done<settlement<mode |
| host unit goswitchd (gap-сборка) | Closure validation debug-секция | журнальные ряды жестов владельца (после приёмки) | ✓ WIRED (форма готова) | unit на v1.1.1-124-g7ace3fd верифицирован live; критерии рядов сформулированы; сами ряды — после жестов владельца |
| case files | main.go + mise.toml | реестр + usage + задачи | ✓ WIRED | term-word-vte-flip: main.go:131/278/340; mise.toml:190 (NOT in ci) |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| машина отложенного флипа | flip target | scriptOf(converted) из реального конвертируемого пробега раунда | да — живой ряд 21:32:54 | ✓ FLOWING |
| stale-cache маршрут | cacheNil | снапшот push-кэша в момент арма реального раунда (:3403) | да — learning=true word-a 21:31:32 | ✓ FLOWING |
| ac_role_errors | счётчик ошибок эпизода | реальные ошибки role-источника живого демона | да — хост показывает 3 | ✓ FLOWING |
| flip-кейс оракулы | pty_hex / delivery_order | реальные VTE pane + журнал демона стенда | да — байт-точные ряды | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Флап-пара сохраняет класс | `go test ./internal/session/ -race -run TestTerminalClass_VerifyDeadSurvivesFlapPair -v` | PASS | ✓ PASS |
| Машина отложенного флипа (defer/release/keys/grace) | `go test ./internal/session/ -race -run 'Flip(DeferredUntilConfirmed\|ReleasedByNextKey\|GraceDeadlineFires\|SurvivesFocusOutGraceReleases)' -v` | 4/4 PASS | ✓ PASS |
| Вторая коррекция/выживание/комбо/fail-closed пины | `go test ./internal/session/ -race -run '(SecondCorrectionReleasesPendingFlip\|ComboFlipImmediateNotDeferred\|HonestPushStillFailClosed\|FlapPairBetweenChunksStops)' -v` | 4/4 PASS | ✓ PASS |
| Stale-cache маршрут | `go test ./internal/session/ -race -run TestTerminalBranch_StaleCacheAnswerRoutesTerminal -v` | PASS | ✓ PASS |
| Регрессия 10-04/10-03 (abort/focus-stop/noRung/veto/class-lifecycle) | `go test ./internal/session/ -race -run '(FreeAbortBeforeBurst\|FocusLossBetweenChunksStops\|VerifyAfterNoRungOnMismatch\|TestAutoCorrect_TerminalVeto\|TestTerminalClass_)'` | ok | ✓ PASS |
| Геометрия пробега (8 golden) | `go test ./internal/correct/ -race -run 'TestConvertibleStretch\|TestStretch'` | ok | ✓ PASS |
| Якорь конфига + токен status | `go test ./internal/config/ ./internal/ctlsvc/ -race -run '(TerminalsDecode\|CorrectionTerminalsAnchor\|AutocorrectTokens)' -v` | все PASS (subtests: без ключа→якорь; []→opt-out) | ✓ PASS |
| Полный CI-гейт | `mise run ci` | exit 0; lint 0 issues; все пакеты ok под -race | ✓ PASS |
| Хост: gap-сборка активна, диагностика видна (read-only) | `systemctl --user is-active goswitchd` + show ExecStart + фикс.ctl status | active; ExecStart=…v1.1.1-124-g7ace3fd/goswitchd; config_valid=true; ac_role_errors=3; autocorrect_fired=0 | ✓ PASS |
| Журнальные ряды стенда — независимый hex-декод | python3 bytes.fromhex(d0bfd180…0a / 6563686f20… / 67686264746e0a) | «привет\n» / «echo привет\n» / «ghbdtn\n» | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| scripts/*/tests/probe-*.sh | find scripts -path '*/tests/probe-*.sh' | не существуют | N/A |
| Живые e2e-кейсы (term-word-vte/-wezterm/-negative-witnesses/-vte-flip) | `go run ./test/e2e -case …` (внутри ВМ) | исполнены исполнителем 2026-10-13 (ряды 18:58-19:08, 21:30-21:32); верификатором НЕ перезапускались | SKIP (re-run) — мутация живого стенда (рестарт демона, окна, CLIPBOARD/gsettings); базис верификатора: датированные ряды + независимый hex-декод + именованные юнит-тесты + live read-only хост-чеки |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| TERM-02 | 10-01…10-09 (все девять) | Слово не в той раскладке исправляется жестом в wezterm и GNOME Terminal; чтение — буфер демона / primary (по необходимости); запись — N×BackSpace + commit_text через IBus; безопасная точка отмены; verify-сверка | ✓ SATISFIED (машинно; финальная приёмка — за владельцем) | Истины 1-5, G1-G11; REQUIREMENTS.md checkbox [x], трассировка «TERM-02 \| Phase 10 \| Complete»; обе причины UAT-разрыва G-10-1 закрыты в коде с behavior-доказательствами и живыми рядами; host re-acceptance — Human Verification 1 |
| — (orphaned) | — | REQUIREMENTS.md отображает на Phase 10 только TERM-02 | ✓ NONE | Потерянных требований нет |

### Prohibition Review (non-authoritative LLM-judge — human review recommended)

Планы 10-07/10-08/10-09 декларируют 11 prohibitions категории safety/values/privacy без tier `verification:` — по ADR-550 D4 записывается НЕАВТОРИТЕТНЫЙ вердикт; человеческое ревью складывается в раунд owner re-acceptance.

| # | Plan | Prohibition (кратко) | LLM-judge verdict | Basis |
| - | ---- | -------------------- | ----------------- | ----- |
| 1 | 10-07 | Флап-толерантность не ослабляет защитные остановки (ctxEpoch/HardReset/resolvePending) | judged PASS (non-authoritative) | Код :1043-1070; пин FlapPairBetweenChunksStops зелёный |
| 2 | 10-07 | Дефолт-якорь только усиливает вето D-55 | judged PASS (non-authoritative) | Якорь — identity-источник; вето-код :2782-2788 нетронут; TestAutoCorrect_TerminalVetoDefaultAnchor зелёный |
| 3 | 10-07 | Инвариант D-55 не пересматривается | judged PASS (non-authoritative) | Свидетели fired=0/veto Δ=1 (18:59, 21:32) |
| 4 | 10-07 | Новая диагностика без содержимого (D-20/D-21) | judged PASS (non-authoritative) | Токен — только счётчик; WARN-текст байт-прежний (WarnOncePerEpisode зелёный без правок) |
| 5 | 10-08 | Замороженные ряды не мутируются (конфиги/оракулы байт-прежние; новый ряд — новый кейс) | judged PASS with NOTE (non-authoritative) | Конфиг-документы и оракулы нетронуты; case_term_word_vte.go получила только арминг-лестницу termArmGateTries=3 (форма гейта, не пин; ряды вышли байт-точными) |
| 6 | 10-08 | Мисматч verify-after только логирует и останавливает | judged PASS (non-authoritative) | release :1984 добавлен В mismatch-ветку, но семантика log+stop сохранена; пины зелёные |
| 7 | 10-08 | Правило владельца сохранено (отложено, не отменено) | judged PASS (non-authoritative) | G7 |
| 8 | 10-08 | Журнальные рубежи без содержимого; grace — не конфиг | judged PASS (non-authoritative) | flipConfirmGrace — const :82 |
| 9 | 10-09 | Хост read-only вне установки | judged PASS (non-authoritative) | Ряд установки + предпроверки 00:46-00:49; независимая пост-проверка верификатором read-only; полная гарантируемость — за ревизией владельца |
| 10 | 10-09 | Разрыв закрывается только вердиктом владельца | judged PASS (non-authoritative) | Closure validation: вердикт ОЖИДАНИЕ не синтезирован (проверено чтением) |
| 11 | 10-09 | Откат-путь назван ДО приёмки | judged PASS (non-authoritative) | Ряд host-install-gap + инструкция Closure validation |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| test/e2e/case_term_word_wezterm.go | 355-380 | wezClipFrozenRow пишет ряд Verdict: FAIL, но возвращает cerr (ошибку захвата), а не результат сравнения hexr — FAIL-ряд не валит exit кейса (продемонстрировано рядом 21:31:38: FAIL при exit 0) | ⚠️ Warning | Тест-форма: строгость клип-ряда не доезжает до вердикта кейса; сам инвариант CLIPBOARD удержан; кандидат на точечный фикс (row-ошибка → return error) вне закрытия G-10-1 |
| docs/adr/adr-008-spike-log.md | :338 | wezterm clip-frozen re-pin ряд записан Verdict: FAIL (pty-дельта: префиксный мусор 16×0x6e + нет хвостового 0a) при пройденной сути (сентинел удержан, «привет» в дельте) | ⚠️ Warning | Заявление 10-08 «пере-pin'нуты байт-точными зелёными» преувеличено для этого ряда; записано честно, разбор исполень пометил на фазовую верификацию — анализ здесь: capture-артефакт дрейфовавшего клиента (0x29 stale-cache контекст), не регрессия безопасности |
| test/e2e/main.go | 117-131 | caseListUsage не называет term-negative-witnesses (IN-01 ревью, подтверждено) | ℹ️ Info | Косметика usage-строки |
| test/e2e/case_term_word_vte.go | 72-78 | boundRowText режет по байтам, может разбить UTF-8 руну (IN-03 ревью) | ℹ️ Info | Диагностика журнала только |
| internal/session/actor.go и др. | — | TBD/FIXME/XXX/TODO/PLACEHOLDER по всем изменённым файлам фазы | — | clean (0) |
| 10-REVIEW.md WR-01…WR-04 | — | Гоночная форма write-ветки, VerifyExpiry без тега раунда, identity-слепота флап-окна, мёртвый xrandr-проб | ℹ️ Info (advisory) | 0 critical; известные предупреждения ревью — консультативные, верификацию не блокируют (контекст раунда) |

### Test Quality Audit (delta)

| Test File | Linked Req | Verdict |
| --------- | ---------- | ------- |
| actor_test.go: флап-корпус, якорный корпус, role-корпус, deferred-flip корпус, stale-cache корпус | TERM-02 | OK — 5/5 RED-свидетельств честны («not ok» целевых тестов, exit 1); behavior-пины исполнены верификатором зелёными |
| config_test.go: двухформенный decode-корпус | TERM-02 | OK — subtests покрывают nil/[]/list точно по контракту |
| ctlsvc_test.go: пин ac_role_errors= | TERM-02 | OK — токен-пин со значением 7, старые токены целы |
| case_term_word_vte_flip.go: оракул порядка доставки | TERM-02 | OK с NOTE — settlement-гейт принимает match/timeout/mismatch (IN-04 ревью: вердикт-текст ряда может преувеличивать «подтверждение доставки» при timeout/mismatch; pty-оракул независим) |

Disabled tests on requirements: 0. Circular patterns: 0 — журнальные hex-ряды независимо декодированы верификатором.

### Human Verification Required

### 1. Owner re-acceptance — закрывающий оракул G-10-1 (6 пунктов, gap-сборка активна)

**Test:** На живом столе владельца (установлена и активна gap-сборка v1.1.1-124-g7ace3fd, старт 00:46:57 MSK — верифицировано live): (1) wezterm: «ghbdtn» + двойной правый Shift → «привет»; (2) GNOME Terminal: то же; (3) флап-бурст — быстро пощёлкать фокус между окнами и СРАЗУ жест в терминале; (4) слово в середине фразы; (5) «ytn?» + жест → «нет,»; (6) автокоррекция молчала (goswitchctl status — autocorrect_fired не вырос). После жестов снять журнальные ряды за окно приёмки: НЕТ skip-рядов verify-timeout/verify-mismatch; done changed=true не потерян по E3 (следующий surrounding-push трекает исправленное слово); флип-рубежи строго после settlement-рубежей. Ответ «approved» (дописать ряд «gap closed by owner re-acceptance» в Closure validation) или описание отклонений (rejected → СТОП без правок кода). Откат по запросу: `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install`.
**Expected:** Все шесть пунктов зелёные в среде, которую стенд принципиально не воспроизводит (флап-тряска ядра 7.2.9-2-t2).
**Why human:** Владелец — единственный валидный оракул закрытия G-10-1 (план 10-09 Task 2 human-check; T-10-09-02); машинные проверки видимость диагностики фиксируют, но приёмку не заменяют; вердикт честно ОЖИДАНИЕ в Closure validation.

### 2. Формальная классификация края TERM-02 (флагнутое допущение; UAT Test 2 — pending)

**Test:** Рассмотреть край-свидетельства — 8 золотых поведений пробега + живые ряды (bash/zsh, мид-лайн, clip-frozen всех прогонов, vim/htop-свидетели, wezterm 10-06 ×2 и re-pin 10-08 с learning=true по stale-cache амендменту, flip-ряд с оракулом порядка) — и вынести вердикт классификации края.
**Expected:** Край принят в том виде, в каком он выражен конкретными проверками; human_needed требовался бы, только если бы край НЕ был так выражен.
**Why human:** Spec-less проба классифицировала край unclassified по построению; флагнутое допущение никогда не снимается автоматически.

### Gaps Summary

Машинных разрывов нет. Обе причины UAT-разрыва G-10-1 закрыты в коде и подкреплены behavior-доказательствами: причина A (10-07) — verifyDead переживает флап-пары (тест зелёный верификатором), классификация получила дефолтный identity-якорь с opt-out и обновлённой документацией, role-диагностика машинно видна (на реальном хосте прямо сейчас ac_role_errors=3 против немого эпизода 11:00:16); причина B (10-08) — SetGlobalEngine физически не следует за commit (release-сайты push/клавиша/grace, ни одного на FocusOut — E3 не воспроизводим по построению), stale-cache амендмент одобрен владельцем (Rule 4) и виден живым рядом learning=true; хост несёт сборку обоих фиксов (верифицировано live read-only), `mise run ci` зелёный на HEAD.

Статус human_needed исключительно из-за честного ожидания владельца: re-acceptance критерия 1 (единственная by-design человеческая истина G12) и формальная классификация края. Два предупреждения зафиксированы консультативно: (1) wezterm clip-frozen re-pin-ряд записан FAIL при удержанном CLIPBOARD-инварианте и exit 0 кейса — форма теста (FAIL-ряд не валит exit) плюс capture-артефакт дрейфовавшего клиента; (2) известные advisory-находки ревью WR-01…WR-04 (0 critical). Ни одно не противоречит критериям успеха ROADMAP и не блокирует достижение цели; оба видны владельцу в этом отчёте до приёмки.

---

_Verified: 2026-10-13T22:30:04Z_
_Verifier: Claude (gsd-verifier)_
