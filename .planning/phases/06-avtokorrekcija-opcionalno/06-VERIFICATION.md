---
phase: 06-avtokorrekcija-opcionalno
verified: 2026-10-07T07:45:54Z
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
covered_digest: "v1:sha256:17d1137a29caf967b8e89ab28c5b0f13a2677cf2fbf9e7fa51f101320a06fa2c"
overrides_applied: 0
gaps: []  # no gaps — fingerprint-staleness refresh: 6 stale paths re-derived (engine/, layouts/ → internal/), all must-haves hold on HEAD 1efc6fa
re_verification:
  previous_status: passed
  previous_score: 51/51
  gaps_closed: []  # nothing to close — no open gaps carried in; this pass re-derives paths and re-proves wiring after post-phase tree evolution (phases 7/8, quick tasks 261006-squ/-vqw, internalize 261007-0yg)
  gaps_remaining: []
  regressions: []
---

# Phase 6: Система автокоррекции — Verification Report (Re-verification: fingerprint refresh)

**Phase Goal:** Опциональная (default off) автокоррекция — молчаливое исправление слова, набранного не в той раскладке, без горячей клавиши. Безопасность гибридным детектором (словарь + ремап + словарь; триграммный fallback для OOV) и политикой app×role на живых AT-SPI-данных (unknown → молчание).
**Verified:** 2026-10-07T07:45:54Z
**Status:** passed
**Re-verification:** Yes — fingerprint-staleness refresh at milestone close (HEAD 1efc6fa; previous verdict 2026-10-04T16:45:00Z, passed 51/51, gaps: [])

## Свежесть (что изменилось на дереве с прошлой верификации)

Прошлый отчёт (passed 51/51, gaps: []) закрывал фазу на HEAD 9dd0de8. С тех пор дерево эволюционировало; эта ре-верификация перепроверила все must-haves на ТЕКУЩЕМ дереве (1efc6fa) first-hand:

- **Фаза 7 (PR #12, 42cbe72):** ревизия D-53 — **white-list `apps` → blocklist `apps_blocklist`** (regex-подстрока, потолок 64, пустые паттерны отвергаются; internal/config/config.go:119, валидация :392-414); SPEC §11 третья датированная ревизия (:411, :418); **ADR-007 Amendment 2026-10-04** (:193 «white-list → blocklist», таблица решений :223-226 — «unknown-идентичность при арме/подтверждении → пропуск», роль/капсы/детектор ОСТАЮТСЯ fail-closed). Формулировки белого списка в прежних must-haves УСТАРЕЛИ — проверяется ДЕЙСТВУЮЩИЙ контракт (см. SC 2).
- **Фаза 8:** a11y-магия (internal/a11y), та же цепочка фолда конфига — автокоррекционный фолд не тронут (actor.go:1316-1322, рядом foldSound :1325 и pushA11y :1326).
- **261006-squ (звук):** постоянный PulseAudio-поток заменяет canberra-сабпроцесс (internal/sound/*; go.mod + jfreymuth/pulse, oggvorbis — санкционированное владельцем spec-delta D-55). Взаимодействие с автокоррекцией перепроверено живыми тестами: TestActor_AutoCorrectFireSoundsSink, TestActor_AutoCorrectAbstainSilent — PASS (прогнан мной).
- **261006-vqw (смешанный текст):** **посимвольная инверсия раскладки = ЦЕЛЕВАЯ семантика mixed** (SPEC §4.2:102-116; D-55 спека-до-кода); internal/correct/runs.go переписан, golden-корпус перепереносён (RED→GREEN), **13 строк e2e-матриц перепереносены** (matrix-v1..v4), case_word.go оракул «паиghbdtn». Это реализация записанного владельцем 2026-10-04 follow-up (06-UAT.md Deferred Follow-Ups п.1) — не новый скоуп фазы.
- **261007-0yg (интернализация):** `engine/` и `layouts/` перенесены под `internal/` (5427d7a; SPEC §8 ревизия 2026-10-07, D-55:359-368). **6 устаревших путей covered_files пере-выведены:** engine/conn_switcher_test.go → internal/engine/conn_switcher_test.go; layouts/{dict_en,dict_ru,dict_test,trigrams}.go, layouts/dictgen/main.go → internal/layouts/….

## Goal Achievement

### Success Criteria (ROADMAP) — вердикты на текущем дереве

| # | Критерий | Вердикт | Evidence (свежий, HEAD 1efc6fa) |
|---|----------|---------|----------|
| 1 | Гибридный детектор: точность на корпусе, FP в бюджете | ✓ VERIFIED | Golden-корпус прогнан мной под -race на текущем дереве: TestCorpus_WrongLayoutBothDirections, **TestCorpus_LegitWordsZeroFalsePositives** — PASS; корпус дополнен mixed-токенами новой семантики (TestCorpus_MixedTokens — PASS). detect.go не менялся (314 строк); пути словарей internal/layouts/dict_{en,ru}.go на месте (78 965 / 138 929 строк) |
| 2 | Политика app×role, unknown → молчание | ✓ VERIFIED (контракт D-53) | **TestAutoCorrect_SilenceMatrix — 9/9 ячеек PASS** под -race (прогнан мной): все СЕКЬЮРИТИ-сегменты fail-closed (роль 40/60, ошибка/дедлайн роли, нет caps, unsure, короткое слово); identity-unknown СТРЕЛЯЕТ (инверсия 07-04, закреплена в silenceMatrixCells actor_test.go:5364), role-ambiguous 61+unknown отвергается (CR-01). Старая white-list формулировка снята ревизией D-53 — действующий контракт blocklist проверен |
| 3 | Default off, hot reload, счётчики в status, лог без содержимого | ✓ VERIFIED | Default off: config.go :251 (AppsBlocklist nil, D-54), actor-фолд :1316-1322 (enabled/blocklist/пороги) + refreshACBlocklist (:2283, перекомпиляция только при изменении паттернов = hot reload); токены status: ctlsvc.go:162 autocorrect_enabled=, :163 autocorrect_fired=, :164 autocorrect_abstained=, :168 ac_skip_*; приватность: TestActor_DebugCorrectionRecord PASS (прогнан мной; «the log leaked %q (D-20/D-21)» — нет утечки содержимого) |
| 4 | Двойной Shift — ручное переопределение | ✓ VERIFIED | TestActor_ManualOverrideAfterAutocorrect PASS под -race (прогнан мной 2026-10-07) |
| 5 | e2e-матрица расширена автокоррекционными кейсами, ночная двойная зелёная | ✓ VERIFIED (вердикт владельца + юнит-доказательство ре-пина) | Матрица расширена (matrix-v4.yaml, 34 кейса; phrase-mixed :110, word-mixed :161); ночной конвейер цел (d48-nightly-dispatch.sh:32-33 → matrix-v4 fresh_session=true; e2e-matrix.yml default v4). Двойная зелень — вердикт владельца 2026-10-04 (32/35×2 после G-6-1, 06-UAT.md тест 1). **13 строк перепереносены** на посимвольную инверсию (261006-vqw) — поведение доказано юнит-корпусом на текущем дереве (TestActor_MixedWordInvertsPerChar, TestActor_PhraseMixedCorrects, TestConvertRuns_MixedInvertsPerChar — PASS, прогнаны мной); живой прогон ре-пиннутых строк — задекларированный ручной гейт самого quick-task'а (261006-vqw-VERIFICATION.md, human_needed) — НЕ гэп фазы 6 (см. «Post-verification evolution») |

### Gap-closure G-6-1 — must-haves планов 06-09/06-10 (7/7) — регрессия на текущем дереве

| # | Truth | Status | Evidence (прогнано/прочитано мной на дереве 2026-10-07) |
|---|-------|--------|----------|
| 1 | 06-09 D1: lock-free AttachEngine — реентерабельный CreateEngine завершается во время флипа без a.mu | ✓ VERIFIED | actor.go:832-834 — тело = один `a.engSlot.Store(&emitterSlot{eng: eng})`, ноль Lock; слот :171, снапшот-аксессоры через emitter(). **TestActor_ReentrantAttachDuringFlip + TestActor_AttachEngineWhileFlipInFlight — прогнаны мной под -race на internal/session: PASS** |
| 2 | 06-09 D2: контракт флипа не тронут (D-36 порядок, 150 мс wedge-guard, WARN-not-fatal, nil seam) | ✓ VERIFIED | Флип-путь и ci-гейт целы (mise.toml:35-37 [tasks.ci]); целевые пакеты session/engine/detect/correct зелёные под -race (точечные прогоны этой сессии) |
| 3 | 06-09 D3: wire-свидетели фаб-шины — синхронный реентерабельный CreateEngine внутри await | ✓ VERIFIED | internal/engine/conn_switcher_test.go (837 строк, перенесён 5427d7a): **TestFactoryReentrantCreateEngineAnswersDuringAwait + …BlockingHandlerIsOnlyGate — прогнаны мной под -race: PASS** |
| 4 | 06-09 D4: ADR-006 Amendment 2026-10-02 (диагноз, свойство, решение) | ✓ VERIFIED | ADR-006:14 «Amended 2026-10-02 — G-6-1: the factory's self-deadlock on a flip» + секция :227 — на дереве |
| 5 | 06-10 D1: e2e-кейс flip-keystroke зарегистрирован, mise-задача вне [tasks.ci] | ✓ VERIFIED | test/e2e/main.go:272 `"flip-keystroke": {fn: runFlipKeystroke, standalone: true}` (:113 usage, :289 standalone-список); mise.toml:145-147 `[tasks.e2e-flip-keystroke]` |
| 6 | 06-10 D2: живое доказательство double-green (24/24 немедленных буквы) | ✓ VERIFIED (записанная улика) | 24/24, RTT 6.50-9.16 мс, ноль WARN — 06-10-SUMMARY Live Results + коррелирующий вердикт владельца (06-UAT.md тест 1); живой прогон — live-стол, вне компетенции верификатора, числа не оспариваются |
| 7 | 06-10 D3: fresh-session вердикт WINDOWS #13 — гейт владельца | ✓ VERIFIED (закрыт владельцем) | 06-UAT.md тест 1: «Владелец 2026-10-04: принял. Матрица 32/35×2 стабильно» — вердикт записан, гейт закрыт |

### Observable Truths исходных 8 планов (44) — quick-regression

44 истины, верифицированные в двух прежних проходах, — quick-regression (существование + санити + точечные поведенческие прогоны) на текущем дереве: артефакты на месте по ПЕРЕ-ВЫВЕДЕННЫМ путям (internal/engine/, internal/layouts/ — ARTIFACTS-OK, см. таблицу артефактов), appid/detect/config/buffer/actor/ctlsvc не регрессировали, статусные токены на месте, греп-гейты SPEC-дельт закрыты чтением (§4.2:102, §8:359, §11:411-418 — все три ревизии на дереве). Ни одна истина не потеряла носитель; 0 FAILED, 0 STUB, 0 ORPHANED, 0 NOT_WIRED, 0 behavior-unverified.

**Score: 51/51 truths verified** (44 исходных + 7 gap-closure; 0 present-behavior-unverified; 0 FAILED; 0 overrides)

## Required Artifacts (ключевые носители на пере-выведенных путях)

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| internal/engine/conn_switcher_test.go | reattach-хук + 2 wire-свидетеля (из engine/, 5427d7a) | ✓ VERIFIED | 837 строк; TestFactoryReentrantCreateEngine* PASS -race (прогнан мной) |
| internal/layouts/dict_en.go / dict_ru.go / trigrams.go | golden словари + триграммы (из layouts/) | ✓ VERIFIED | 78 965 / 138 929 / 8 213 строк; едят тесты detect-корпуса — PASS |
| internal/layouts/dict_test.go / dictgen/main.go | golden-тесты + генератор | ✓ VERIFIED | на месте (102 / 373 строки) |
| internal/correct/runs.go (+runs_test.go) | посимвольная инверсия mixed-пайплайна (261006-vqw) | ✓ VERIFIED | 140 строк; TestConvertRuns_MixedInvertsPerChar PASS -race |
| internal/session/actor.go | D-53 конъюнкция, lock-free AttachEngine, config-фолд | ✓ VERIFIED | 2946 строк; :832-834, :1316-1322, :2283; silence-matrix 9/9 |
| internal/config/config.go | секция autocorrect (default off) + apps_blocklist | ✓ VERIFIED | :119 AppsBlocklist, :251 default nil, :392-414 валидация (потолок 64, пустые отвергаются) |
| internal/ctlsvc/ctlsvc.go | renderStatus-токены автокоррекции | ✓ VERIFIED | :162-164, :168 |
| test/e2e/cases/matrix-v4.yaml | 13 mixed-строк на посимвольной инверсии | ✓ VERIFIED | phrase-mixed :110, word-mixed :161; значения = юнит-доказанной семантике |
| docs/SPEC.md / ADR-006 / ADR-007 | три ревизии (§4.2, §8, §11) + 2 amendment | ✓ VERIFIED | SPEC :102/:359/:411; ADR-006 :14/:227; ADR-007 Accepted (:5) + Amendment D-53 (:193) |
| 06-UAT.md | приёмка 5/5 с вердиктами владельца | ✓ VERIFIED | status: complete, 5/5 pass — не изменилась |

## Behavioral Spot-Checks (прогоны этой верификации, 2026-10-07)

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| G-6-1 реентераб-корпус | `go test ./internal/session/ -race -run 'TestActor_ReentrantAttachDuringFlip\|TestActor_AttachEngineWhileFlipInFlight' -count=1` | ok | ✓ PASS |
| Wire-свидетели фабрики | `go test ./internal/engine/ -race -run 'TestFactoryReentrantCreateEngine' -count=1` | ok | ✓ PASS |
| Golden-корпус 0-FP + mixed | `go test ./internal/detect/ -race -run 'TestCorpus_WrongLayoutBothDirections\|TestCorpus_LegitWordsZeroFalsePositives\|TestCorpus_MixedTokens' -count=1 -v` | все PASS | ✓ PASS |
| Silence-matrix D-53 | `go test ./internal/session/ -race -run 'TestAutoCorrect_SilenceMatrix' -count=1 -v` | 9/9 ячеек PASS | ✓ PASS |
| Посимвольная инверсия | `go test ./internal/session/ ./internal/correct/ -race -run '…MixedWordInvertsPerChar\|…PhraseMixedCorrects\|TestConvertRuns_MixedInvertsPerChar' -count=1 -v` | все PASS | ✓ PASS |
| Ручное переопределение | `go test ./internal/session/ -race -run 'TestActor_ManualOverrideAfterAutocorrect' -count=1` | ok | ✓ PASS |
| Приватность лога + звук | `go test ./internal/session/ -race -run 'TestActor_DebugCorrectionRecord\|TestActor_AutoCorrectFireSoundsSink\|TestActor_AutoCorrectAbstainSilent' -count=1 -v` | все PASS | ✓ PASS |

Пропущено: живые e2e-прогоны (матрицы v1-v4, flip-keystroke) — live GNOME-стол; носители — записанные вердикты владельца (06-UAT.md) + юнит-доказательства семантики (выше).

## Probe Execution

Не применимо — проект не объявляет scripts/*/tests/probe-*.sh; ни один план фазы (вкл. 06-09/06-10) не декларирует пробники. Носители — mise-задачи и go test (выше).

## Requirements Coverage

| Requirement | Source Plan(s) | Status | Evidence |
| ----------- | -------------- | ------ | -------- |
| SPEC §10/§11 (spec-delta), D-51 | 06-01..08 | ✓ SATISFIED | SPEC §11 несёт три датированные ревизии (:411 blocklist — актуальная); ADR-007 Accepted + Amendment D-53 |
| CORR-01..09 | 06-02..08 | ✓ SATISFIED | Golden-корпус 0-FP PASS на текущем дереве; silence-matrix 9/9; mixed-семантика обновлена spec-delta §4.2 и доказана юнит-корпусом |
| MACR-ACL (app×role) | 06-01,03,04,06,07 | ✓ SATISFIED | Действующий контракт — blocklist D-53 (адаптация прежней white-list формулировки, санкционированная владельцем); fail-closed сегменты security доказаны |
| SWCH-01/SWCH-02 (G-6-1) | 06-09, 06-10 | ✓ SATISFIED | lock-free фабрика + реентераб-корпус + wire-свидетели PASS на internal/ путях |

Orphaned requirements: НЕТ (REQUIREMENTS.md по-прежнему заморожен до complete-milestone по D-51 — формализация REQ-ID при new-milestone).

## Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| — | — | Debt markers (TBD/FIXME/XXX/HACK) в файлах фазы (пере-выведенные пути) | none | скан 2026-10-07 — 0 совпадений |
| — | — | Placeholder / stub / empty impl | none | 0 (два совпадения «unknown placeholder» в test/e2e/case_switch.go:79/:161 — константа формата e2e-отчёта, не заглушка: значение не течёт в прод-рендер) |
| — | — | Disabled tests (t.Skip) в requirement-linked тест-файлах | none | 0 (совпадения SkipReasons — словарь причин абстенций) |
| — | — | go.mod churn | none | +2 прямых депа звука (jfreymuth/pulse, oggvorbis) — санкционированное spec-delta 261006-squ, вне скоупа фазы 6 |

## Post-verification Evolution (информационное, не гэпы фазы)

1. **Живые гейты 261006-vqw.** Ре-пин 13 mixed-строк (261006-vqw, реализация записанного владельцем follow-up) несёт СОБСТВЕННЫЕ живые ручные гейты: 4 матрицы + ручной mixed-чек (261006-vqw-VERIFICATION.md, status: human_needed — все автоматизируемое зелёное). Это pending-приёмка quick-task'а, а не регрессия фазы 6: mixed-строки и раньше не были живо-зелёными (владелец принял их как drift с записанной целевой семантикой). На milestone close виден здесь — единый sink гейта остаётся в артефакте 261006-vqw.
2. **Former deferred follow-ups — оба разрешены:** (а) семантика смешанного текста — реализована 261006-vqw (SPEC §4.2, «Смешанный текст: посимвольная инверсия раскладки» — todo закрыт, VERIFICATION quick-task'а на дереве); (б) UX white-list identity-discovery — растворён ревизией D-53: blocklist не требует наблюдаемой идентичности (ADR-007:223 прямо ссылается на UX-пробел 06-UAT).
3. **Dep-базис:** go.mod вырос на звуковые депы (см. Anti-Patterns) — сознательное решение владельца в post-фазном quick-task; ограничение «stdlib + godbus + минимум» эволюционирует через spec-delta, к цели фазы 6 отношения не имеет.

## Decision Coverage

`gsd-tools check.decision-coverage-verify` (fresh, 2026-10-07): **5/5 honored, 0 not honored** — «All trackable CONTEXT.md decisions are honored by shipped artifacts.» (В прошлый раз гейт скипался вербом из-за проектного лэйаута; теперь отработал и подтвердил перенос 5/5.)

## Gaps Summary

Гэпов нет. Все 51 must-have (44 исходных + 7 gap-closure) держатся на текущем дереве (HEAD 1efc6fa): поведенческие истины переподтверждены точечными прогонами под -race (G-6-1 корпус, wire-свидетели, golden-корпус 0-FP, silence-matrix 9/9, посимвольная инверсия, ручное переопределение, приватность лога, звук-интеракция), артефакты на месте по пере-выведенным путям (internal/engine/, internal/layouts/), три post-фазные ревизии SPEC (§4.2 mixed-семантика, §8 internalize, §11 blocklist) не тронули ни одного носителя фазы. Ревизия D-53 white-list → blocklist — санкционированная владельцем эволюция контракта SC 2 (ADR-007 Amendment), действующая семантика доказана инвертированной silence-matrix. 0 FAILED, 0 STUB, 0 ORPHANED, 0 NOT_WIRED, 0 behavior-unverified, 0 overrides. Цель фазы достигнута.

---

_Verified: 2026-10-07T07:45:54Z_
_Verifier: Claude (gsd-verifier)_
