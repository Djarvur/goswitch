---
phase: 06-avtokorrekcija-opcionalno
verified: 2026-10-07T23:18:41Z
status: passed
score: 51/51 must-haves verified
behavior_unverified: 0
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
  - .planning/phases/06-avtokorrekcija-opcionalno/06-09-PLAN.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-09-SUMMARY.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-10-PLAN.md
  - .planning/phases/06-avtokorrekcija-opcionalno/06-10-SUMMARY.md
  - README.md
  - docs/CONFIG.md
  - docs/LICENSE-data.md
  - docs/SPEC.md
  - docs/adr/ADR-006-two-engine-revision.md
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
  - internal/correct/runs.go
  - internal/correct/runs_test.go
  - internal/ctlsvc/ctlsvc.go
  - internal/ctlsvc/ctlsvc_test.go
  - internal/detect/detect.go
  - internal/detect/detect_test.go
  - internal/engine/conn_switcher_test.go
  - internal/layouts/dict_en.go
  - internal/layouts/dict_ru.go
  - internal/layouts/dict_test.go
  - internal/layouts/dictgen/main.go
  - internal/layouts/trigrams.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - scripts/d48-nightly-dispatch.sh
  - test/e2e/case_autocorrect.go
  - test/e2e/case_switch.go
  - test/e2e/case_word.go
  - test/e2e/cases/matrix-v4.yaml
  - test/e2e/fixtures/input.html
  - test/e2e/fixtures/password_entry.py
  - test/e2e/main.go
  - test/e2e/matrix.go
  - test/e2e/perf.go
covered_digest: "v1:sha256:473ee185130fa3a3212d107342d8798a1a2cfad7b9c3ca3ee5b7e2f199420794"
overrides_applied: 0
gaps: []  # no gaps — second fingerprint-staleness refresh: quick 261008-00m landed after the 07:45:54Z verdict; all 51 must-haves re-proven on HEAD 39a5d82
re_verification:
  previous_status: passed
  previous_score: 51/51
  gaps_closed: []  # nothing to close — no open gaps carried in; this pass absorbs the 261008-00m tree delta (flipCredit, matrix hermeticity, docs deltas) into the phase verdict
  gaps_remaining: []
  regressions: []
---

# Phase 6: Система автокоррекции — Verification Report (Re-verification: second fingerprint-staleness refresh)

**Phase Goal:** Опциональная (default off) автокоррекция — молчаливое исправление слова, набранного не в той раскладке, без горячей клавиши. Безопасность гибридным детектором (словарь + ремап + словарь; триграммный fallback для OOV) и политикой app×role на живых AT-SPI-данных (unknown → молчание).
**Verified:** 2026-10-07T23:18:41Z
**Status:** passed
**Re-verification:** Yes — second staleness refresh at milestone close (HEAD 39a5d82; previous verdict 2026-10-07T07:45:54Z, passed 51/51, gaps: [])

## Свежесть (что изменилось на дереве с прошлой верификации)

Прошлый отчёт (passed 51/51, gaps: []) закрывал фазу на HEAD 1efc6fa. После него на дерево лёг quick-task **261008-00m** (milestone-close owner fixes), изменивший код, покрытый этой фазой. Эта ре-верификация — ФОКУСИРОВАННАЯ: структурный анализ дельты + точечные поведенческие прогоны по тронутым швам (все под -race, прогнаны мной на текущем дереве 39a5d82).

- **G-2-3 / flipCredit (4a0c887 + 77c91b7 RED):** `internal/session/actor.go` — новое поле `flipCredit` (a.mu-only, :263); `flipTo` (:2017) вооружает РОВНО ОДИН кредит только при УСПЕШНОМ round trip к switcher-шву; `HandleLifecycle` FocusOut (:787-795) расходует кредит и пропускает РОВНО ОДИН побочный эффект — `a.buf.HardReset()`. Собственный флип демона (SetGlobalEngine; синтетическая пара focus_out/focus_in от re-mint движков) — не «смена фокуса окна» (ADR-004 amendment 2026-10-07, SPEC-ревизия G-2-3 с аудит-трейлом). **Изоляция дельты проверена diff'ом 1efc6fa..HEAD:** АВТОКОРРЕКЦИОННЫЙ ПУТЬ НЕ ТРОНУТ — гейт :2497 → `acConfirmRefusals` :2542 → `acFired++` :2500 → `startRangeCorrection` :2509/:2745, абстенции `acAbstained++` :2413, счётчики :220-221/:615/:678-679 — вне диффа; `a.resolvePending()` (:799 — retirement открытого correction-раунда, основа confirm re-check) выполняется БЕЗУСЛОВНО, в т.ч. на синтетическом FocusOut.
- **Регрессия по автокоррекционным истинам на акторе** (fired gate, confirm re-check, абстенции, счётчики, silence-matrix, ручное переопределение, приватность, звук): 17 именованных тестов прогнаны мной под -race — все PASS (см. Behavioral Spot-Checks).
- **Собственные тесты инварианта flipCredit** (новые, actor_test.go): TestActor_OwnFlipMixedWordCorrectsWhole, TestActor_OwnFlipCreditConsumedOnce (кредит расходуется РОВНО ОДИН раз — второй FocusOut берёт полный сброс), TestActor_OwnFlipFailedSwitcherArmsNothing (НЕудачный round trip не вооружает кредит) — содержательные поведенческие тесты состояния, прогнаны мной под -race: PASS.
- **Не тронуты (нулевый diff 1efc6fa..HEAD):** internal/detect/, internal/correct/, internal/config/, internal/ctlsvc/, internal/appid/, internal/layouts/, internal/engine/, test/e2e/cases/, docs/CONFIG.md, ADR-007. Golden-детекторный корпус — спот-чек прогоном: PASS.
- **G-5-5 / install (2fa8b53):** internal/install/* — НЕ входит в covered_files фазы 6 (шов фаз 4/5; их VERIFICATION.md уже обновлены — 4220a43/68fd703). Регресс-чек фазы 6 не требуется; на пересечении (config-фолд, D-53) — изменений нет.
- **Hermeticity стоя (4dc4877):** test/e2e/matrix.go — `caseNeedsBaseEstablishment`: каждый кейс-демон пинится на -config документ; семантика демона не менялась. Комментарий к диффу несёт ЖИВОЕ СВИДЕТЕЛЬСТВО на fix-дереве: adopted desktop config'ом **автокоррекция выстрелила живьём** внутри word-after-space кейса («autocorrect fired … its flip_after_correction kept the correction buffer alive») — fired-путь работает вживую после fix.
- **Доки (0781a5f):** SPEC §4 (буфер: собственный флип не чистит, rev G-2-3, прежний пункт сохранён дословно как аудит-трейл), §4 (G-5-5 auto-wrap — вне скоупа фазы 6), ADR-006 Amendment G-5-5; NEW docs/adr/ADR-004-buffer-reset-triggers.md. Только доки, датированы, аудит-трейл.
- **Живые матрицы (orchestrator-recorded, 261008-00m-VERIFICATION.md):** v1 **16/16** (word-mixed ряд → «паиghbdtn»), v2 **20/21** (единственный FAIL — задокументированный environmental focus-stealing класс, не продукт) — слой автокоррекции работает вживую после fix.

## Goal Achievement

### Success Criteria (ROADMAP) — вердикты на текущем дереве (HEAD 39a5d82)

| # | Критерий | Вердикт | Evidence (свежий) |
|---|----------|---------|----------|
| 1 | Гибридный детектор: точность на корпусе, FP в бюджете | ✓ VERIFIED | detect.go/словари — нулевой diff с вердикта 07:45; golden-корпус спот-чек прогоном мной под -race: TestCorpus_WrongLayoutBothDirections, TestCorpus_LegitWordsZeroFalsePositives, TestCorpus_MixedTokens — PASS |
| 2 | Политика app×role, unknown → молчание | ✓ VERIFIED (контракт D-53) | **TestAutoCorrect_SilenceMatrix — PASS** под -race (прогнан мной 2026-10-07T23:xx): fail-closed сегменты целы; identity-unknown СТРЕЛЯЕТ (TestAutoCorrect_UnknownIdentityFires — PASS); confirm-политика: TestAutoCorrect_ConfirmBlocklistRefuses + TestAutoCorrect_ConfirmUnknownPasses — PASS. ADR-007 не менялся (нулевой diff) |
| 3 | Default off, hot reload, счётчики в status, лог без содержимого | ✓ VERIFIED | config.go/ctlsvc.go — нулевой diff (defaults :251, валидация :392-414; статус-токены :162-164/:168 на месте); **счётчики переподтверждены прогоном: TestAutoCorrect_CountersAndReasons — PASS**; приватность: TestActor_DebugCorrectionRecord — PASS; фолд конфига :1316-1322 не тронут |
| 4 | Двойной Shift — ручное переопределение | ✓ VERIFIED | TestActor_ManualOverrideAfterAutocorrect — PASS под -race на ДЕРЕВЕ С flipCredit (интеракция override × синтетический FocusOut не сломала переопределение) |
| 5 | e2e-матрица расширена автокоррекционными кейсами, ночная двойная зелёная | ✓ VERIFIED (живое свидетельство на fix-дереве) | Матрица не тронута суждениями фазы (matrix.go diff — только hermeticity пин стоя); **на fix-дереве автокоррекция выстрелила живьём** (лог «autocorrect fired» в word-after-space, комментарий matrix.go:888-905) — матрицы v1 16/16 / v2 20/21 (orchestrator-recorded, 261008-00m-VERIFICATION.md); юнит-базис mixed-семантики переподтверждён (TestActor_MixedWordInvertsPerChar, TestActor_PhraseMixedCorrects — PASS) |

### Gap-closure G-6-1 — must-haves планов 06-09/06-10 (7/7) — carried forward

Швы G-6-1 (lock-free AttachEngine, реентерабельный CreateEngine, wire-свидетели, ADR-006, e2e flip-keystroke, живые вердикты 06-UAT) — **нулевой diff с прошлого вердикта** (actor.go:832-834 слот, internal/engine/conn_switcher_test.go, mise.toml, main.go, 06-UAT.md не менялись в 1efc6fa..HEAD; единственный diff actor.go — flipCredit, слот-механика не тронута). Вердикты 7/7 ✓ VERIFIED переносятся; свежий живой witness дополнительно: матрица v1 16/16 на fix-дереве.

### Observable Truths исходных 8 планов (44) — carried forward + spot-regression

44 истины двух прежних проходов — носители на месте по покрытым путям (61/61 существуют на текущем дереве, ARTIFACTS-OK), т.ч. все истины, сидящие на акторе, переподтверждены точечными поведенческими прогонами этой сессии. 0 FAILED, 0 STUB, 0 ORPHANED, 0 NOT_WIRED, 0 behavior-unverified.

**Score: 51/51 truths verified** (44 исходных + 7 gap-closure; 0 present-behavior-unverified; 0 FAILED; 0 overrides)

## Required Artifacts (ключевые носители — существование перепроверено на 39a5d82)

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| internal/session/actor.go | D-53 конъюнкция, lock-free AttachEngine, config-фолд, flipCredit | ✓ VERIFIED | 2991 строк; дельта изолирована (:263/:787-795/:2017); автокоррекционный путь :2497-2509/:2542/:2745 не тронут; silence-matrix PASS |
| internal/detect/detect.go + internal/layouts/* | golden словари + детектор | ✓ VERIFIED | нулевой diff; corpus PASS |
| internal/config/config.go | секция autocorrect (default off) + apps_blocklist | ✓ VERIFIED | нулевой diff (:119, :251, :392-414) |
| internal/ctlsvc/ctlsvc.go | renderStatus-токены автокоррекции | ✓ VERIFIED | нулевой diff (:162-164, :168) |
| internal/correct/runs.go | посимвольная инверсия mixed-пайплайна | ✓ VERIFIED | нулевой diff; TestConvertRuns-семантика переподтверждена через actor-прогоны |
| test/e2e/matrix.go + cases/matrix-v4.yaml | матрица + hermeticity пин | ✓ VERIFIED | дифф = caseNeedsBaseEstablishment (только сто-пин); TestMatrixCaseNeedsBaseEstablishment PASS (261008-00m); живое свидетельство 16/16 |
| docs/SPEC.md / ADR-006 / ADR-004 (new) | ревизии с аудит-трейлом | ✓ VERIFIED | SPEC G-2-3 (собственный флип не чистит буфер; прежний пункт сохранён дословно); ADR-006 Amendment G-5-5 (вне скоупа фазы 6); ADR-004 amendment — датированы |
| 06-UAT.md | приёмка 5/5 с вердиктами владельца | ✓ VERIFIED | не менялась с прошлого вердикта |

## Behavioral Spot-Checks (прогоны этой верификации, 2026-10-07T23:xx, все -race -count=1)

| Behavior | Command (сокр. -run паттерн, pkg internal/session если не указано) | Result | Status |
| -------- | ------- | ------ | ------ |
| Fired gate сквозь пайплайн | `TestAutoCorrect_FiresThroughPipeline` | PASS | ✓ |
| Confirm re-check (ре-вход, blocklist-отказ, unknown-пропуск) | `TestAutoCorrect_ReEntryRechecks\|TestAutoCorrect_ConfirmBlocklistRefuses\|TestAutoCorrect_ConfirmUnknownPasses` | 3/3 PASS | ✓ |
| Счётчики и причины | `TestAutoCorrect_CountersAndReasons` | PASS | ✓ |
| Silence-matrix D-53 + identity-unknown | `TestAutoCorrect_SilenceMatrix\|TestAutoCorrect_UnknownIdentityFires` | PASS | ✓ |
| Абстенции беззвучны, fired звучит, лог без содержимого | `TestActor_AutoCorrectAbstainSilent\|TestActor_AutoCorrectFireSoundsSink\|TestActor_DebugCorrectionRecord` | 3/3 PASS | ✓ |
| Ручное переопределение (на дереве с flipCredit) | `TestActor_ManualOverrideAfterAutocorrect` | PASS | ✓ |
| Инвариант flipCredit (exactly-once, no-arm-on-fail, mixed через флип) | `TestActor_OwnFlipMixedWordCorrectsWhole\|TestActor_OwnFlipCreditConsumedOnce\|TestActor_OwnFlipFailedSwitcherArmsNothing` | 3/3 PASS | ✓ |
| Mixed-семантика (юнит-базис SC 5) | `TestActor_MixedWordInvertsPerChar\|TestActor_PhraseMixedCorrects` | 2/2 PASS | ✓ |
| Golden-детекторный корпус (спот-чек; нулевой diff) | `go test ./internal/detect/ -race -run 'TestCorpus_WrongLayoutBothDirections\|TestCorpus_LegitWordsZeroFalsePositives\|TestCorpus_MixedTokens'` | ok | ✓ |

Пропущено: живые e2e-прогоны (матрицы v1-v4, flip-keystroke) — live GNOME-стол; носители — orchestrator-recorded прогоны на fix-дереве (v1 16/16, v2 20/21; 261008-00m-VERIFICATION.md) + юнит-доказательства (выше).

## Probe Execution

Не применимо — проект не объявляет scripts/*/tests/probe-*.sh; ни один план фазы не декларирует пробники. Носители — mise-задачи и go test (выше).

## Requirements Coverage

| Requirement | Source Plan(s) | Status | Evidence |
| ----------- | -------------- | ------ | -------- |
| SPEC §10/§11 (spec-delta), D-51 | 06-01..08 | ✓ SATISFIED | SPEC §11 ревизии целы; новая §4-ревизия G-2-3 — датированный аудит-трейл, ни один автокоррекционный пункт не переформулирован |
| CORR-01..09 | 06-02..08 | ✓ SATISFIED | Golden-корпус PASS; silence-matrix PASS; fired/confirm/counter корпус PASS на дереве с flipCredit |
| MACR-ACL (app×role) | 06-01,03,04,06,07 | ✓ SATISFIED | blocklist D-53 контракт неизменен (ADR-007 нулевой diff); fail-closed сегменты PASS |
| SWCH-01/SWCH-02 (G-6-1) | 06-09, 06-10 | ✓ SATISFIED | Швы G-6-1 — нулевой diff; живое свидетельство v1 16/16 на fix-дереве |

Orphaned requirements: НЕТ (REQUIREMENTS.md заморожен до complete-milestone по D-51).

## Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| — | — | Debt markers (TBD/FIXME/XXX/HACK/PLACEHOLDER) в тронутых файлах фазы (actor.go, actor_test.go, matrix.go) | none | скан 2026-10-07T23:xx — 0 совпадений |
| — | — | Placeholder / stub / empty impl | none | 0; новые OwnFlip-тесты — содержательные (проверяют empty-buffer абстенцию и RequireSurroundingText-счётчик) |
| — | — | go.mod churn с прошлого вердикта | none | 0 (звуковые депы легли ДО вердикта 07:45) |

## Post-verification Evolution (информационное, не гэпы фазы)

1. **flipCredit (G-2-3)** — целевая семантика буфера amended владельцем (ADR-004, SPEC rev G-2-3);autocorrect-слой не пострадал: весь поведенческий корпус фазы PASS на новом дереве, живое свидетельство «autocorrect fired» на fix-дереве. Интеракция «autocorrect выстрелил → собственный флип → буфер выжил» — не автокоррекционная истина (коррекция уже завершена к моменту флипа); override-интеракция переподтверждена прогоном.
2. **Hermeticity стоя (4dc4877)** — стендовые демоны пинятся на config; это чинит сам стенд (adoption живого десктоп-конфига), не демон. Побочное свидетельство: adopted-config прогон доказал живую стрельбу автокоррекции.
3. **G-5-5 (install)** — вне covered_files фазы 6; гейт живёт в фазах 4/5 (уже обновлены).

## Decision Coverage

Все 5 trackable CONTEXT.md-решений перенесены прошлым проходом (check.decision-coverage-verify 5/5 honored); с тех пор CONTEXT.md фазы не менялся.

## Gaps Summary

Гэпов нет. Все 51 must-have держатся на текущем дереве (HEAD 39a5d82). Дельта 261008-00m (flipCredit + стендовая hermeticity + доки-ревизии) структурно изолирована от автокоррекционного пути (дифф-анализ: fired-гейт, confirm re-check, абстенции, счётчики — вне диффа; resolvePending безусловен), и все истины, сидящие на акторе, переподтверждены 17 именованными поведенческими прогонами под -race этой сессии. Неавтокоррекционные швы (detect, correct, config, ctlsvc, appid, layouts, engine, ADR-007) — нулевой diff. Живое свидетельство на fix-дереве: автокоррекция выстрелила в матричном прогоне (лог «autocorrect fired»), v1 16/16 / v2 20/21. 0 FAILED, 0 STUB, 0 ORPHANED, 0 NOT_WIRED, 0 behavior-unverified, 0 overrides. Цель фазы достигнута.

---

_Verified: 2026-10-07T23:18:41Z_
_Verifier: Claude (gsd-verifier)_
