---
phase: 06-avtokorrekcija-opcionalno
verified: 2026-10-02T20:15:00Z
status: human_needed
score: 42/44 must-haves verified
behavior_unverified: 1
covered_files:
  - .github/workflows/e2e-matrix.yml
  - .planning/phases/06-avtokorrekcija-opcionalno/06-01-PLAN.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-01-SUMMARY.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-02-PLAN.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-02-SUMMARY.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-03-PLAN.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-03-SUMMARY.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-04-PLAN.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-04-SUMMARY.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-05-PLAN.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-05-SUMMARY.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-06-PLAN.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-06-SUMMARY.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-07-PLAN.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-07-SUMMARY.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-08-PLAN.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-08-SUMMARY.md
  - README.md
  - docs/CONFIG.md
  - docs/LICENSE-data.md
  - docs/SPEC.md
  - docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md
  - docs/ci-runner.md
  - internal/appid/appid.go
  - internal/appid/appid_test.go
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/config/load_test.go
  - internal/config/watch_test.go
  - internal/correct/buffer.go
  - internal/correct/buffer_test.go
  - internal/ctlsvc/ctlsvc.go
  - internal/ctlsvc/ctlsvc_test.go
  - internal/detect/detect.go
  - internal/detect/detect_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - layouts/dict_en.go
  - layouts/dict_ru.go
  - layouts/dict_test.go
  - layouts/dictgen/main.go
  - layouts/trigrams.go
  - mise.toml
  - scripts/d48-nightly-dispatch.sh
  - test/e2e/case_autocorrect.go
  - test/e2e/cases/matrix-v4.yaml
  - test/e2e/fixtures/input.html
  - test/e2e/fixtures/password_entry.py
  - test/e2e/main.go
  - test/e2e/matrix.go
  - test/e2e/perf.go
covered_digest: "v1:sha256:82f7fe91d107f798943d2feea0a1917d09e634c207cfb2d2e8ad72d0dc63f7ff"
overrides_applied: 0
gaps: []  # no gaps: every artifact the 8 plans promised exists, is substantive, and is wired
behavior_unverified_items:
  - truth: "Полный регресс: mise run e2e-matrix-v4 зелёный ДВАЖДЫ ПОДРЯД на живом столе (06-08 must-have T2; критерий 5 «ночная двойная зелёная»)"
    test: "Run the formal fresh-session gate: docs/ci-runner.md dispatch (fresh_session=true, ≤30-min session) or two back-to-back `mise run e2e-matrix-v4` on a settled session"
    expected: "Both runs exit 0 (35/35) — or the owner adjudicates the 9 known drift rows (WINDOWS #13) as accept-with-risk / desktop-fix"
    why_human: "Artifacts are present and wired (matrix-v4, nightly pipeline, mise tasks — all verified on the tree; autocorrect rows PASS ×2 first-hand in /tmp/06-08-matrix-double.log), but the full double-green was honestly red twice (23/35, 24/35) on environmental session-drift that survives `ibus restart`, is byte-identical on the pre-phase-6 tree d96dcb2, and the formal gate is the owner's machine-checked instrument at verify-work. A verifier re-run on this desktop cannot distinguish code from session state."
unverified-prohibitions:
  - statement: "06-01 P1: spec-delta сохраняет старый вердикт §11 дословно (audit-trail)"
    verdict: "LLM-judge: PASS — grep 'отвергнута владельцем осознанно' present; §11 read: old verdict verbatim + dated HTML-comment decision (audit-trail style §4.1). Non-authoritative."
  - statement: "06-02 P2: dictgen НЕ копирует GPL-код xneur/Easy Switcher"
    verdict: "LLM-judge: PASS — no GPL sources in any read_first; dictgen written from stdlib primitives; license notes verbatim from local copyright files in docs/LICENSE-data.md. Non-authoritative."
  - statement: "06-03 P2: сохранённая пара (sender,path) — только адресат запроса, не основание срабатывания"
    verdict: "LLM-judge: PASS — code read of appid.Observer.Role (snapshot + live roleCall; no role cache) and actor autoConfirm (live GetRole at confirm; cached pair never decides). Non-authoritative."
  - statement: "06-07 P4: фикстура не печатает пароль-подобный текст вне stdout-оракула"
    verdict: "LLM-judge: PASS — password_entry.py prints only FINAL:/PLAIN: of the (expected-unchanged) field; the case injects no real secret. Non-authoritative."
decision_coverage:
  honored: 5
  total: 5
  not_honored: []
human_verification:
  - test: "WINDOWS #13 / критерий 5 — формальный двойной зелёный прогон matrix-v4 на свежей сессии (fresh_session=true, docs/ci-runner.md dispatch) или два прогона подряд на устоявшейся сессии; вердикт по 9 дрейф-строкам: принять-с-риском / чинить стол / ночной гейт"
    expected: "Оба прогона exit 0 (35/35), либо владелец фиксирует вердикт в WINDOWS #13 (fixed/waive). Autocorrect-строки (fires/off/oov) и словесно/фразовые носители ручных контрактов уже зелёные ×2 — предмет решения только дрейф-строки + 2 замороженные WINDOWS #12 строки"
    why_human: "Средовой дрейф стола (переживает ibus restart, байт-идентичен на до-фазовом дереве d96dcb2) неразличим автоматикой от кодовой регрессии без свежей сессии; вердикт — владельца по ADR-007 Acceptance Evidence строке 5"
  - test: "ADR-007: перевод Proposed → Accepted владельцем на гейте verify-work по таблице Acceptance Evidence (6 строк) внутри ADR"
    expected: "Status section of docs/adr/ADR-007 flips to Accepted with date — only if the owner accepts row 5's environmental risk (or after the formal gate greens)"
    why_human: "План 06-08 прямо запретил исполнителю флипать гейт владельца (T-06-08-01: красный прогон не пропускается); решение — владельца, а не верификатора"
  - test: "CR-01 (async confirm race window) — живая проверка фикса: включить autocorrect для white-list приложения, набрать слово + разделитель и НЕМЕДЛЕННО продолжить набор (перехлёст с role-RTT окном ≤25 мс) на живом столе"
    expected: "Поле не портится: коррекция либо применяется к правильному диапазону, либо молча абстейнется (счётчик ac_skip_payload_stale в status); никакого delete по чужим рунам"
    why_human: "Фикс закрыт детерминированным юнит-тестом (TestAutoConfirm_RevalidatesArmedPayload, RED→GREEN), но живой перехлёст keystroke×RTT по контракту фиксера требует человеческой верификации — юнит не воспроизводит реальный тайминг wire-доставки"
  - test: "Живой UAT трёх автокоррекционных кейсов на здоровой сессии владельца: mise run e2e-autocorrect-fires / e2e-autocorrect-password-silent / e2e-autocorrect-terminal-silent"
    expected: "fires: ghbdtn→привет без хоткея + ручной Double конвертирует обратно; password-silent: witness PASSWORD_TEXT(40), символы нетронуты, FINAL-verbatim, fired=0; terminal-silent: fired=0 + ноль коррекционных записей при любой доставке IME (wezterm-гонка — свойство молчания)"
    why_human: "Живые прогоны требуют реального GNOME-стола и инжекции в сессию владельца; верификатор не запускает e2e на живом столе; юнит-политика (silence-matrix 10 ячеек) доказана, живое поведение — нет"
  - test: "Review judgment-tier prohibitions (4 позиции выше) — human review на гейте verify-work вместе с ADR-007 Accept"
    expected: "Владелец подтверждает 4 LLM-judge вердикта (audit-trail, GPL-clean, cache-not-basis, fixture no-leak) или опровергает"
    why_human: "Judgment-tier запрещения не имеют детерминированной проверки; автоматика дала неавторитетные вердикты с уликами — молчаливый pass запрещён"
coincidental_reliance_items: []
advisory: []
---

# Phase 6: Система автокоррекции — Verification Report

**Phase Goal:** Опциональная (default off) автокоррекция — молчаливое исправление слова, набранного не в той раскладке, без горячей клавиши. Безопасность гибридным детектором (словарь + ремап + словарь; триграммный fallback для OOV) и политикой app×role на живых AT-SPI-данных (unknown → молчание). SPEC §10/§11 правится spec-delta внутри фазы (D-51: milestone v1.1.0).
**Verified:** 2026-10-02T20:15:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

Верификация goal-backward, first-hand: все греп-гейты 8 планов перезапущены на дереве, ключевые именованные тесты прогнаны под `-race`, полный `mise run ci` прогнан один раз (зелёный, 17 пакетов ok), живые логи 06-07/06-08 в /tmp прочитаны first-hand (матрица-двойной-прогон, perf-пары). Ни один SUMMARY-клэйм не принят на слово без проверки на дереве.

### Success Criteria (ROADMAP) — вердикты

| # | Критерий | Вердикт | Evidence |
|---|----------|---------|----------|
| 1 | Гибридный детектор: точность на корпусе, FP в бюджете | ✓ VERIFIED | `internal/detect` (314 строк, чистый headless — без D-Bus/часов/горутин); golden-корпус на реальных hunspell-данных прогнан мной: TestCorpus_WrongLayoutBothDirections (190 уверенных), TestCorpus_LegitWordsZeroFalsePositives (**404 слова, 0 ложных — бюджет = ровно 0**), TestCorpus_OOVNames (34, ни одного Confident), TestCorpus_ReasonVocabularyClosed — все PASS под -race; словарь+ремап+словарь = словарный путь, OOV = триграммный fallback (порогово-мёртвый на дефолтах — санкционированный консервативный старт, см. Honest Outcomes) |
| 2 | Политика app×role, unknown → молчание | ✓ VERIFIED (unit) / живой негатив → UAT | `internal/appid`: Role(ctx) — живой GetRole enum по (sender,path) последнего gain, ErrRoleUnknown fail-closed, GetRoleName=0; актор: конъюнкция D-53 (enabled → white-list точным равенством → caps → детектор → живая роль 61/79/94 вне мьютекса под 25 мс → CR-01/WR-01 re-checks) — TestAutoCorrect_SilenceMatrix (10 ячеек: role 40/60/ошибка/timeout/app-unknown/app-not-listed/no-caps/unsure/короткое) PASS, TestAutoCorrect_FailClosedNoFailOpen PASS. Живые негативы (GTK4-пароль, wezterm) — 06-07 live PASS + item 4 human_verification |
| 3 | Default off, hot reload, счётчики в status, лог без содержимого | ✓ VERIFIED | TestDefaults_AutocorrectOff PASS (zero-value = off); TestWatch_BrokenAutocorrectBlockKeepsLastGood PASS + load.go/watch.go не тронуты фазой (git log: последние правки — фаза 3); TestApplySnapshot_AutocorrectFold PASS (hot reload живьём); renderStatus tokens (ctlsvc.go:160-165: autocorrect_enabled/fired/abstained + ac_skip_*) ; TestStatus_AutocorrectFields PASS; приватность: slog в detect.go = 0, Reason — замкнутый набор 8 констант, actor пишет только счётчик/слаг (`"msg":"autocorrect","reason":"fired"`), ни токена ни конвертации |
| 4 | Двойной Shift — ручное переопределение поверх автокоррекции | ✓ VERIFIED | TestActor_ManualOverrideAfterAutocorrect PASS (авто ghbdtn→привет, затем Double конвертирует обратно — ручной путь без детектора); manual-пути байт-нетронуты (acBoundary=false сохраняет cached-push шорткат); матричные носители (flip-after-word-correction, combo-word-layout, x11-word-*, регистровые, фразовые) зелёные ×2 — first-hand в /tmp/06-08-matrix-double.log; fires-кейс 06-07 пинит CORR-01 на той же поверхности |
| 5 | e2e-матрица расширена автокоррекционными кейсами, ночная двойная зелёная | ⚠️ PARTIAL — расширение VERIFIED, двойная зелень → human gate | matrix-v4.yaml (433 строки): v3 байт-дословно (git diff --exit-code на v3 пуст — V3-FROZEN) + autocorrect-блок; ночной конвейер на v4 (e2e-matrix.yml default в обеих D-48-развилках, d48-nightly-dispatch.sh → matrix-v4.yaml, docs/ci-runner.md секция v4); autocorrect-строки PASS ×2 first-hand в логе двойного прогона. **Полный двойной зелёный НЕ достигнут: 23/35 и 24/35** — средовой session-drift (WINDOWS #13: REAL и STICKY, переживает ibus restart, диф фазы exonерирован трижды) → human_verification item 1 |

### Observable Truths (must_haves по 8 планам)

**06-01 (spec-delta + ADR-007) — 5/5**

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SPEC §2 п.3 — автокоррекция = v1.1.0, не «ВНЕ объёма» | ✓ VERIFIED | SPEC-DELTA-OK (перезапущен: v1.1.0 + 2026-09-27 + ADR-007 в SPEC.md; «явно ВНЕ объёма» отсутствует) |
| 2 | §10 без запрещающего буллета, 3 остальных нетронуты | ✓ VERIFIED | AUDIT-TRAIL-OK: « Wayland-композиторы кроме GNOME» ровно 1, GUI настроек на месте; буллет §10:189 заменён датированным указателем |
| 3 | §11 — старый вердикт дословно + датированное решение D-51 | ✓ VERIFIED | §11 прочитана: вердикт «отвергнута владельцем осознанно…» дословно + HTML-комментарий 2026-09-27 + абзац решения со ссылкой на ADR-007 |
| 4 | ADR-007 существует, 5 секций, kill-таблица, лицензии, GTK3-факт | ✓ VERIFIED | ADR-SKELETON-OK + ADR-CONTENT-OK (перезапущены); Status Proposed; Acceptance Evidence таблица 6 строк (06-08) |
| 5 | Спека ДО кода; милстоун-переключатели не тронуты | ✓ VERIFIED | git diff пуст на ADR-006/PROJECT.md/REQUIREMENTS.md; STATE.md не в дифе фазовых коммитов кода |

**06-02 (dictgen, golden-данные) — 5/5**

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | dictgen — генератор в золотом паттерне, CI goldens only | ✓ VERIFIED | GENERATED-OK; dictgen/main.go субстантивен; opencorpora=0; mise dictgen-regen вне [tasks.ci]; TestGenerateLive скипается без hunspell (CI-независимость) |
| 2 | DictRU/DictEN — отсортированные []string литералы, НЕ map | ✓ VERIFIED | layouts/dict_ru.go / dict_en.go []string-литералы; TestGolden_DictSorted PASS |
| 3 | TriRU/TriEN map[string]float64, ≤4096/язык, детерминизм | ✓ VERIFIED | TestGolden_TrigramsBounded PASS; trigrams.go 2×4096 |
| 4 | Golden-инварианты зелёные (сорт/дизъюнкт/без ё/конверты) | ✓ VERIFIED | TestGolden_{DictSorted,DictDisjoint,DictNoYo,DictEnvelope,TrigramsBounded} все PASS под -race (прогнаны мной) |
| 5 | Лицензионные нотиси сохранены | ✓ VERIFIED | docs/LICENSE-data.md: Lebedev custom-bsd + SCOWL/Atkinson; заголовки в сгенерированных файлах |

**06-03 (appid Role) — 5/5**

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | (focusSender, focusPath) cache-keep под тем же мьютексом | ✓ VERIFIED | appid.go run(): запись трёх значений одной критической секцией; TestObserver_{StoresFocusPair,LossKeepsPair,FreshObserverUnknown} в зелёном пакете |
| 2 | Role(ctx) — живой GetRole в момент решения, пустой путь → сентинел | ✓ VERIFIED | Код прочитан: снапшот под мьютексом, path==""/roleCall==nil → ErrRoleUnknown; один CallWithContext; пакет ok под -race |
| 3 | Роль только по enum (40/60/61/79/94), GetRoleName нигде | ✓ VERIFIED | ROLE-SEAM-OK; grep -c GetRoleName = 0; TestRole_EnumConstants |
| 4 | Ошибки — error return, шов инжектируем, nil-шов → «неизвестно» | ✓ VERIFIED | TestRole_{NilSeamUnknown,LiveCallShape,ConcurrentWithGain}; пакет ok |
| 5 | Role не пишет в лог | ✓ VERIFIED | Код Role() — ноль slog; grep |

**06-04 (config autocorrect) — 5/5**

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Секция в strict-decode + hot reload без нового кода load/watch | ✓ VERIFIED | SCHEMA-OK; TestLoad_StrictDecodeRejectsTypo PASS; TestWatch_BrokenAutocorrectBlockKeepsLastGood PASS; load.go/watch.go — ноль правок фазой (git log) |
| 2 | Default off везде, нулевое значение = off | ✓ VERIFIED | TestDefaults_AutocorrectOff PASS; Config «exactly the five sections» |
| 3 | Валидация: потолок 64, диапазоны, поименованные ошибки | ✓ VERIFIED | TestValidate_Autocorrect{AppsCeil,Ranges,RangeBoundaries} в зелёном пакете; сентинелы в config.go |
| 4 | White-list — плоский список, точное совпадение, порядок незначим; пустой = глухо | ✓ VERIFIED | TestValidate_AppsOrderIrrelevant, TestLoad_EmptyAppsEnabledIsValid + акторный гейт len(Apps)>0 (actor.go) — конъюнкция требует непустой список |
| 5 | docs/CONFIG.md — секция autocorrect + приватность | ✓ VERIFIED | CONFIG-DOC-OK (five sections, min_word_len, trigram_margin, «никогда») |

**06-05 (detect) — 6/6**

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Verdict/Check — детерминированная чистая функция | ✓ VERIFIED | Код прочитан (Check :129 — полная ветка словарь→вето→fallback→abstain); DICT-PATH-OK; slog=0 |
| 2 | Словарный путь D-52(а): вето cur-hit/both-hit, confident при miss+hit | ✓ VERIFIED | TestCheck_{DictWrongLayout,DictCurHitVetoes,DictBothHitVetoes,DirMatchesCorrect} в зелёном пакете; both-hit классифицируется первым (fixed-порядок) |
| 3 | Триграммный fallback только при обоюдном miss; margin И floor | ✓ VERIFIED | TestTrigram_{WrongLayout,UnsureBelowMargin,UnsureBelowFloor,NeutralTrigramsUnknown,NormalizedByLength,DictHitSkipsScoring} PASS |
| 4 | Границы abstain (короткий/без-букв/mixed/1-рунный) | ✓ VERIFIED | TestCheck_{AbstainShort,AbstainNoLetters,AbstainConvertFail,UnknownModeAbstains} PASS |
| 5 | Golden-корпус: FP = ровно 0 на легитимных; OOV никогда не уверены | ✓ VERIFIED | TestCorpus_{WrongLayoutBothDirections(190),LegitWordsZeroFalsePositives(404,0),OOVNames(34)} — прогнаны мной, PASS |
| 6 | Приватность: Verdict без токена, замкнутый словарь причин | ✓ VERIFIED | TestCorpus_ReasonVocabularyClosed PASS (8 констант, рефлексия 4 полей); slog=0 |

**06-06 (акторная интеграция) — 7/7**

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Граница слова: PushFeed на трёх сайтах, reset-граница ДО HardReset, consume нетронут | ✓ VERIFIED | HOOK-OK; TestBuffer_PushFeed, TestActor_{ResetKeyBoundary,BackspaceNoBoundary,FocusOutNoBoundary} в зелёных пакетах |
| 2 | Конъюнкция D-53 — все шесть условий, fail-closed, warn-once | ✓ VERIFIED | TestAutoCorrect_SilenceMatrix (10 ячеек) + TestAutoCorrect_FailClosedNoFailOpen + TestAutoCorrect_WarnOncePerEpisode — прогнаны мной, PASS |
| 3 | GetRole вне мьютекса под дедлайном ≤25 мс; единственный путь замены | ✓ VERIFIED | TestAutoCorrect_RoleDeadlineBounded PASS; autoRoleTimeout=25ms; startRangeCorrection — единственный вход (CommitText count = 5, греп-пин неизменен); SINGLE-PATH |
| 4 | Счётчики AutoCorrectStats; дебаунс по вето dict-cur-hit | ✓ VERIFIED | TestAutoCorrect_CountersAndReasons PASS; AutoCorrectCounters в actor.go |
| 5 | Status/ctl: 4 поля + renderStatus токены до config-блока; CLI без правок | ✓ VERIFIED | TestStatus_AutocorrectFields + TestApplySnapshot_AutocorrectFold + TestRenderStatus_AutocorrectTokens PASS; cmd mains — git diff пуст |
| 6 | Двойной Shift — переопределение поверх автокоррекции (критерий 4) | ✓ VERIFIED | TestActor_ManualOverrideAfterAutocorrect PASS (прогнан мной) |
| 7 | Лог: только счётчик/причина, слово никогда (вкл. DEBUG) | ✓ VERIFIED | Код autoConfirm — slog без токена; ревью подтвердило D-20/D-21 на всех новых сайтах; корпус-пин журнала |

**06-07 (живые e2e + матрица v4) — 6/6**

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | autocorrect-fires (живой кейс) | ✓ VERIFIED (артефакт+политика) | case_autocorrect.go зарегистрирован в 3 местах; строка PASS ×2 first-hand в /tmp/06-08-matrix-double.log (:522, :1020); живой огонь → UAT item 4 |
| 2 | autocorrect-password-silent (живой негатив, 3 оракула) | ✓ VERIFIED (артефакт+политика) | Кейс с witness-гейтами (FAIL-not-skip: errFixtureOracle/witness gate в коде); role-40-запрет доказан ячейкой silence-matrix; live PASS 06-07 → UAT item 4 |
| 3 | autocorrect-terminal-silent (оракул по счётчикам) | ✓ VERIFIED (артефакт+политика) | app-not-listed ячейка матрицы PASS; live PASS ×2 (после re-pin 1dc76af на свойство безопасности); терминальный оракул = silence property → UAT item 4 |
| 4 | password_entry.py — GTK4-фикстура, FINAL-оракул | ✓ VERIFIED | FIXTURES-OK (py_compile + GtkPasswordEntry + FINAL: + type="password" в input.html); 74 строки |
| 5 | Матрица v4 — superset v3 + autocorrect-блок; ночной конвейер на v4 | ✓ VERIFIED | V3-FROZEN + NIGHTLY-V4-OK (перезапущены); matrix-v4.yaml 433 строки; документированное честное исключение password-строки (нет password-шага у драйвера) |
| 6 | Chromium-роль пароля пинится живьём | ✓ VERIFIED | A2 спинен живьём (google-chrome 153 → PASSWORD_TEXT=40, комментарий в case_autocorrect.go:28-41) |

**06-08 (приёмка) — 3/5 verified, 1 behavior-unverified, 1 owner-gate**

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Perf-гейт Pitfall 5: p95 дельта ≤ +5 мс, числа в ADR-007 | ✓ VERIFIED | Perf-пары прочитаны first-hand (/tmp/06-08-perf-{baseline,ac}-report.txt): baseline p95 180.2 → ON 176.9, **дельта −3.3 мс — гейт ПРОЙДЕН**; ADR-007 Consequences несёт обе пары + VmHWM 19.1 МБ < 50 МБ; [tasks.e2e-perf-autocorrect] вне ci |
| 2 | Полный регресс: matrix-v4 зелёный ДВАЖДЫ ПОДРЯД | ⚠️ PRESENT_BEHAVIOR_UNVERIFIED | Артефакты wired; автокоррект-строки и носители зелёные ×2, но полный прогон 23/35 и 24/35 — средовой дрейф (WINDOWS #13, exonерирован трижды); формальный гейт — владельца → behavior_unverified_items + human_verification |
| 3 | ADR-007 переведён в Accepted | ⚠️ HUMAN GATE (by design) | Status честно ОСТАЁТСЯ Proposed — план прямо запретил флип при красном прогоне (T-06-08-01); полная Acceptance Evidence таблица в ADR (6 строк, все улики верифицированы мной); флип — вердикт владельца → human_verification item 2 |
| 4 | README — раздел автокоррекции с честными оговорками | ✓ VERIFIED | README-OK (default off + GTK3 + autocorrect прочитаны); white-list модель, приватность, ручные жесты = переопределение; perf-таблица 04-07 нетронута |
| 5 | Милстоун-дисциплина D-51 | ✓ VERIFIED | git diff пуст на PROJECT.md/STATE.md/REQUIREMENTS.md/release.yml/.goreleaser.yaml (перезапущено); CASES-FROZEN (v3/v4) |

**Score: 42/44 truths verified** (1 present-behavior-unverified: 06-08 T2; 1 owner-decision pending: 06-08 T3; 0 FAILED; 0 overrides)

### Review fixes CR-01/CR-02/WR-01 — присутствие верифицировано

| Fix | Коммиты | Код на дереве | Тест (прогнан мной) |
|-----|---------|----------------|---------------------|
| CR-01 re-validation | 3e64e95/d5f11e4 | acPayload.expectToken/expectTail (:344-345, :1815-1816), acConfirmRefusals payload-stale (:2037-2040) | TestAutoConfirm_RevalidatesArmedPayload **PASS** |
| CR-02 always-verify | 92991c8/2d9ea85 | `rng.acBoundary \|\| !MatchesSuffix` — кэш-шорткат никогда для boundary (:2287) | TestAutoConfirm_ResetBoundaryAlwaysVerifies **PASS** |
| WR-01 atomic conjunction | 802de00/7c8d3cc (+48013f2 lint) | payload.app + re-check acFocusedApp под мьютексом, app-changed слаг (:2043-2050) | TestAutoConfirm_RechecksFocusedApp **PASS** |

Все три теста прогнаны вместе под -race — PASS. CR-01 помечен фиксером как требующий живой верификации (async race window) → human_verification item 3.

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| docs/SPEC.md (§2/§10/§11) | spec-delta с аудит-трейлом | ✓ VERIFIED | гейт-грепы + чтение §11 |
| docs/adr/ADR-007-…md | 5 секций, kill-таблица, evidence | ✓ VERIFIED | 187 строк; Proposed + таблица 6 строк |
| layouts/dictgen/ + dict_ru/en.go + trigrams.go | golden-данные | ✓ VERIFIED | инварианты PASS; dict_ru 138 914 / dict_en 78 951 |
| internal/appid | Role(ctx) живой GetRole | ✓ VERIFIED | 295 строк, пакет ok -race |
| internal/config | секция autocorrect | ✓ VERIFIED | SCHEMA-OK; пакет ok |
| internal/detect | гибридный детектор | ✓ VERIFIED | 314 строк, корпус 628+ subtests PASS |
| internal/correct/buffer.go | PushFeed | ✓ VERIFIED | Push — обёртка; корпус зелёный |
| internal/session/actor.go | хук/конъюнкция/счётчики | ✓ VERIFIED | D-53 contour + 3 review-fix, CommitText=5 |
| internal/ctlsvc/ctlsvc.go | renderStatus токены | ✓ VERIFIED | :160-165 + ac_skip_ |
| test/e2e/case_autocorrect.go + fixtures | 3 живых кейса | ✓ VERIFIED | 784+74 строк; REGISTERED-OK |
| test/e2e/cases/matrix-v4.yaml | superset v3 + autocorrect | ✓ VERIFIED | 433 строки; v3 frozen |
| test/e2e/perf.go + mise задачи | perf-вариант | ✓ VERIFIED | PERF-RECORDED; задачи вне ci |
| README.md / docs/CONFIG.md / docs/ci-runner.md / LICENSE-data.md | документация | ✓ VERIFIED | все греп-гейты |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| internal/session/actor.go | internal/detect | detect.Check(Data{layouts.DictRU/EN}, Trigrams{TriRU/EN}) | ✓ WIRED | actor.go:1863-1865 |
| internal/session/actor.go | internal/appid | RoleSource тайп-ассерт + rs.Role(ctx) | ✓ WIRED | actor.go:479-493, :1479, :1982 |
| internal/session/actor.go | startRangeCorrection | единственный путь замены | ✓ WIRED | actor.go:2016; CommitText=5 (пин) |
| internal/correct/buffer.go | actor feedKey | PushFeed на трёх сайтах Push | ✓ WIRED | HOOK-OK |
| internal/ctlsvc/ctlsvc.go | goswitchctl | generic key=value токены | ✓ WIRED | cmd mains — git diff пуст |
| docs/SPEC.md §11 | ADR-007 | перекрёстные ссылки | ✓ WIRED | §11 → ADR-007; ADR цитирует §11 |
| mise.toml / workflow / dispatch | matrix-v4.yaml | ночной конвейер | ✓ WIRED | NIGHTLY-V4-OK |
| internal/config | actor Options | applySnapshot-фолд 5 полей | ✓ WIRED | TestApplySnapshot_AutocorrectFold PASS |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| detect.Check | токен слова | буфер актора (PushFeed/Token) | да | ✓ FLOWING |
| детекторные словари | Data{RU,EN} | layouts.DictRU/DictEN (запечённые hunspell golden) | да (138 914/78 951) | ✓ FLOWING |
| роль | rs.Role(ctx) | живой GetRole по (sender,path) фокус-события | да (a11y-шина; шов инжектируем для тестов) | ✓ FLOWING |
| статус-счётчики | AutoCorrectStats | acFired/acAbstained/acReasons актора | да | ✓ FLOWING |
| e2e оракулы | autocorrect_fired / ac_skip_* | goswitchctl status реального демона | да (логи 06-08) | ✓ FLOWING |

Hollow-пейзажей нет: ни один вердикт/счётчик не питается статикой.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Полный CI (build+vet+lint strict+test -race) | `mise run ci` | 17 пакетов ok, 0 FAIL | ✓ PASS |
| Silence matrix (10 ячеек fail-closed) | `go test ./internal/session/ -run TestAutoCorrect_SilenceMatrix -race -v` | 10/10 subtests PASS | ✓ PASS |
| CR-01/CR-02/WR-01 фиксы | `go test -run 'TestAutoConfirm_{RevalidatesArmedPayload,ResetBoundaryAlwaysVerifies,RechecksFocusedApp}' -race` | 3/3 PASS | ✓ PASS |
| Ручное переопределение (критерий 4) | `go test -run TestActor_ManualOverrideAfterAutocorrect -race` | PASS | ✓ PASS |
| Zero-FP корпус (критерий 1) | `go test ./internal/detect/ -run 'TestCorpus_' -race` | 4/4 PASS | ✓ PASS |
| Golden-инварианты данных | `go test ./layouts/ -run TestGolden -race` | 8/8 PASS | ✓ PASS |
| Hot reload + default off + фолд | `go test ./internal/config/ ./internal/session/ -run 'TestWatch_Broken…\|TestLoad_Strict…\|TestApplySnapshot_AutocorrectFold\|TestDefaults_AutocorrectOff\|TestStatus_AutocorrectFields' -race` | 5/5 PASS | ✓ PASS |
| Autocorrect-строки матрицы живьём ×2 | /tmp/06-08-matrix-double.log (:522-524, :1020-1022) | PASS ×2 | ✓ PASS |
| Автокоррект-строки матрицы (полный прогон) | `mise run e2e-matrix-v4` ×2 (06-08, формальный D-48) | 23/35, 24/35 — средовой дрейф | ⚠️ → human gate |

Пропущено: перезапуск трёх живых кейсов (требуют живой GNOME-стол и инжекцию в сессию владельца — не делает верификатор) → human_verification item 4.

### Probe Execution

Не применимо — проект не объявляет scripts/*/tests/probe-*.sh (скан пуст); планы фазы не декларируют пробники. Носители доказательств — mise-задачи e2e (см. выше).

### Requirements Coverage

| Requirement | Source Plan(s) | Description | Status | Evidence |
| ----------- | -------------- | ----------- | ------ | -------- |
| SPEC §10/§11 (spec-delta) | 06-01,02,04,05,07,08 | возврат автокоррекции по D-51 | ✓ SATISFIED | SPEC.md §2/§10/§11 + ADR-007 + доказательства (см. SC-таблицу) |
| CORR-01 | 06-06,07,08 (non-regression) | двойной Shift не сломан, остаётся переопределением | ✓ SATISFIED | юнит-пин + матричные носители ×2 + fires-кейс |
| CORR-02 | 06-06,08 (non-regression) | Triple-путь не тронут хуком | ✓ SATISFIED | фразовые носители ×2 (x11-phrase PASS в логе); session/correct корпус зелёный |
| CORR-03 | 06-08 (non-regression) | select-контракт | ✓ SATISFIED (unit) / live → WINDOWS #13 | юнит-носители зелёные; select-строки матрицы красные ×2 — с drift-подписью (потерянная первая клавиша до жеста), exonерированы контролем d96dcb2; вердикт — владелец в item 1 |
| CORR-04 | 06-05,08 | направление по составу буфера | ✓ SATISFIED | детектор потребляет correct.Dir без правок; обе directions в корпусе 190 + матрица |
| CORR-05 | 06-05,08 | регистры | ✓ SATISFIED | 190 кейсов трёх регистров; регистровые строки v4 зелёные ×2 |
| CORR-06 | 06-05,08 | смешанный текст → молчание в автокоррекции; ручной частичный путь нетронут | ✓ SATISFIED | abstain-convert-fail пин + unit CORR-06 корпус зелёный; word/phrase-mixed строки v4 — замороженные WINDOWS #12 строки (pre-existing фаза 5, вердикт владельца — вне скоупа фазы 6) |
| CORR-07 | 06-06,08 | замена по диапазону той же лестницей | ✓ SATISFIED | единственный конвейер (CommitText=5); ladder-носители; CR-01/CR-02 усилили диапазонную безопасность |
| CORR-08 | 06-02 | golden-генераторный контракт расширен на словарные данные | ✓ SATISFIED | dictgen; CI goldens only (bwrap-доказательство 06-02); все TestGolden зелёные |
| CORR-09 | 06-06,08 | сбросы буфера (Enter/Tab/Escape) не изменены | ✓ SATISFIED | reset-граница ДО HardReset + транзит пин; CR-02 делает reset-confirm always-verify (усиление) |
| MACR-ACL (app×role) | 06-01,03,04,06,07 | app×role политика, unknown → молчание | ✓ SATISFIED | appid Role + конъюнкция актора + white-list конфига + silence-matrix + живые негативы 06-07 |

Orphaned requirements: НЕТ. REQUIREMENTS.md не отображает фазу 6 (D-51 freeze — формальная фиксация REQ-ID при new-milestone v1.1.0, санкционировано 06-CONTEXT/планами); строка REQUIREMENTS.md:76 «Автокоррекция… отвергнута» относится к v1.0.0 и superseded-ена spec-delta в docs/SPEC.md — REQUIREMENTS.md заморожен до complete-milestone по D-51.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| — | — | Debt markers (TBD/FIXME/XXX) | none | скан всех фазовых файлов — 0 |
| — | — | Placeholder / not-yet-implemented | none | 0 по всем фазовым файлам |
| — | — | Empty implementations / stubs | none | Known Stubs: none во всех 8 SUMMARY, подтверждено чтением кода |
| — | — | go.mod churn | none | tidy CLEAN; только 3 санкционированные зависимости |

Info-уровень находки ревью IN-01..04 (counter semantics, per-word INFO, godbus pending-call accumulation, fixture Wait race) — задокументированы в 06-REVIEW.md, оставлены по скоуп-инструкции оркестратора; ни одна не блокирует цель фазы. Известный флак TestRun_OnConnHookCalledOnce — предсуществующий, root-caused в deferred-items.md (ассерт читает calls до waitHookContext), вне скоупа фазы.

### Honest Outcomes (не гэпы — взвешенные санкционированные исходы)

1. **Триграммный fallback порогово-мёртв на дефолтах** (floor 1.0 выше log10-шкалы таблиц, все значения ≤ 0 → trigram-unsure всегда). Санкционировано research A4 и планом 06-05 («fallback отключается порогами — санкционированный исход»); конфиг-калибруется парой мест (config Defaults + DefaultParams, кросс-комментарии); зафиксировано в комментарии DefaultParams. Согласуется с критерием 1: OOV не дают уверенной коррекции.
2. **Терминальный оракул — свойство молчания** (wezterm регистрирует IBus context и гонит доставку IME; re-pin 1dc76af: fired=0 + ноль коррекционных записей при любой доставке — не ставка на одну руку гонки). Два подряд зелёных живых прогона.
3. **Полный двойной зелёный v4 — WINDOWS #13** (открыт): дрейф REAL и STICKY, переживает ibus restart, диф фазы exonерирован трижды; формальный ≤30-мин fresh_session гейт — машинный инструмент владельца.
4. **Перф-окно бюджет-FAIL** (p95 ~177 мс ≥ 50 мс) — санкционированный исход 04-07/WINDOWS #5 (ydotool-инжектор доминирует в окне; реакция демона суб-миллисекундная); дельта фичи −3.3 мс — в бюджете.

### Decision Coverage

All trackable CONTEXT.md decisions are honored by shipped artifacts. (gsd-tools check.decision-coverage-verify: 5/5 honored, 0 not honored.)

### Test Quality Audit

| Corpus | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|--------|-----------|--------|---------|----------|-----------------|---------|
| internal/detect (TestCheck/Trigram/Corpus, 628+ subtests) | CORR-04/05/06, критерий 1 | да | только TestGenerateLive (hunspell-гейт, санкционирован) | нет (golden на данных, определения → ожидания) | Value/Behavioral | PASS |
| internal/session (silence-matrix, TDD RED-корпуса 06-06 + review-fix) | MACR-ACL, CORR-01/07/09 | да | нет | нет | Behavioral (ноль операций + ровно один слаг на ячейку) | PASS |
| internal/config (strict-decode/hot-reload) | D-32/D-33/D-54 | да | нет | нет | Behavioral (last-good, whole-file rejection) | PASS |
| layouts golden | CORR-08 | да | TestGenerateLive вне CI | нет | Value (конверты/инварианты) | PASS |
| e2e stand units (matrix/perf/watchdog/quiesce) | критерий 5 | да | нет | нет | Status/Value | PASS |

Disabled tests on requirements: 0 блокирующих (единственный скип — TestGenerateLive по отсутствию системного hunspell в CI, санкционирован золотым контрактом CORR-08). Circular patterns: 0. Insufficient assertions: 0.

### Human Verification Required

### 1. WINDOWS #13 — формальный двойной зелёный matrix-v4 (критерий 5)

**Test:** Прогнать формальный fresh-session гейт (docs/ci-runner.md dispatch, fresh_session=true, сессия ≤30 мин) или два `mise run e2e-matrix-v4` подряд на устоявшейся сессии; вынести вердикт по 9 дрейф-строкам (accept-with-risk / чинить стол / ночной гейт) и по 2 замороженным WINDOWS #12 строкам.
**Expected:** Оба прогона exit 0 (35/35) — либо вердикт владельца зафиксирован в WINDOWS #13 (fixed/waive). Autocorrect-строки и словесно/фразовые носители уже зелёные ×2.
**Why human:** Средовой дрейф стола неразличим автоматикой от кодовой регрессии; формальный гейт — машинный инструмент владельца; ADR-007 evidence-строка 5 несёт полную запись.

### 2. ADR-007: Proposed → Accepted (гейт владельца)

**Test:** На verify-work прочитать Acceptance Evidence таблицу (6 строк) в docs/adr/ADR-007 и флипнуть Status в Accepted (с датой) — только если строка 5 (двойная зелень) принимается с зафиксированным средовым риском или закрыта item 1.
**Expected:** Status: Accepted + дата; формулировка «принято владельцем» (прецедент ADR-006).
**Why human:** План 06-08 прямо запретил исполнителю флипать гейт владельца при красном прогоне (T-06-08-01); решение — владельца.

### 3. CR-01 async race window — живая проверка фикса

**Test:** Включить autocorrect для white-list приложения; на живом столе набрать слово + разделитель и немедленно продолжить набор (перехлёст с role-RTT ≤25 мс).
**Expected:** Поле не портится: коррекция либо ложится ровно в вооружённый диапазон, либо молча абстейнтся (ac_skip_payload_stale в status).
**Why human:** Фикс доказан детерминированным юнит-тестом, но живой тайминг wire-доставки юнитом не воспроизводится; фиксер явно пометил логику «flagged for human verification per fixer contract».

### 4. Живой UAT трёх автокоррекционных кейсов

**Test:** `mise run e2e-autocorrect-fires`, `mise run e2e-autocorrect-password-silent`, `mise run e2e-autocorrect-terminal-silent` на здоровой сессии владельца.
**Expected:** fires — ghbdtn→привет без хоткея, ручной Double конвертирует обратно, fired=1; password-silent — witness PASSWORD_TEXT(40) до/после, символы нетронуты, FINAL-verbatim, fired=0, ровно одна role-forbidden абстиненция; terminal-silent — fired=0 + ноль коррекционных записей при любой доставке.
**Why human:** Живые прогоны инжектируют в реальную сессию владельца; верификатор подтверждает артефакты/политику/логи, но не глаз владельца на живом столе.

### 5. Judgment-tier prohibitions (4 позиции)

**Test:** Подтвердить/опровергнуть 4 неавторитетных LLM-judge вердикта из frontmatter `unverified-prohibitions` (audit-trail дословность, GPL-clean dictgen, кэш-не-основание-срабатывания, фикстура-без-утечки) — вместе с гейтом ADR-007 Accept.
**Expected:** 4 подтверждения (или опровержения с уликами) на гейте verify-work.
**Why human:** Judgment-tier запрещения не имеют детерминированной проверки; молчаливый pass запрещён (ADR-550 D4).

### Gaps Summary

Гэпов нет: каждый артефакт, обещанный 8 планами, существует на дереве, субстантивен и wired (0 MISSING, 0 STUB, 0 ORPHANED, 0 NOT_WIRED). Единственные незакрытые must-have-истины (06-08 T2 «двойная зелень ×2» и 06-08 T3 «ADR-007 Accepted») — намеренные гейты владельца по дизайну фазы: T2 заблокирован подтверждённым средовым дрейфом стола (exonерирован трижды, WINDOWS #13), T3 прямо зарезервирован планом за владельцем при красном T2. Обе истины маршрутизированы в human_verification, не в gaps: дерево содержит всё необходимое для их закрытия (формальный fresh-session гейт, полная evidence-таблица в ADR).

---

_Verified: 2026-10-02T20:15:00Z_
_Verifier: Claude (gsd-verifier)_
