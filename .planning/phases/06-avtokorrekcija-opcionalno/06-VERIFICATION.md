---
phase: 06-avtokorrekcija-opcionalno
verified: 2026-10-04T16:45:00Z
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
  - engine/conn_switcher_test.go
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
  - test/e2e/case_switch.go
  - test/e2e/cases/matrix-v4.yaml
  - test/e2e/fixtures/input.html
  - test/e2e/fixtures/password_entry.py
  - test/e2e/main.go
  - test/e2e/matrix.go
  - test/e2e/perf.go
covered_digest: "v1:sha256:5f8ca4edc428bfb0b2a5b90d3ff3a190b4faf99adb2890914df5a87842838bcc"
overrides_applied: 0
gaps: []  # no gaps — gap G-6-1 resolved (06-09/06-10), UAT 5/5 complete with owner verdicts, both former human gates closed
re_verification:
  previous_status: human_needed
  previous_score: 42/44
  gaps_closed:
    - "06-08 T2 — полный двойной зелёный matrix-v4: корень красных строк найден UAT-измерением (самоблокировка фабрики на флипе, G-6-1), закрыт планами 06-09/06-10; после фикса матрица 32/35 ×2 подряд идентично, все 9 дрейф-строк зелёные; вердикт владельца 2026-10-04 (06-UAT.md тест 1)"
    - "06-08 T3 — ADR-007 Proposed → Accepted: владелец принял на verify-work 2026-10-04; Status: Accepted на дереве (commit 87d5650)"
    - "G-6-1 — первая буква после флипа: lock-free AttachEngine (atomic emitterSlot), живое доказательство 24/24 немедленных букв, RTT 6.5–9.2 мс, ноль deadline-WARN"
  gaps_remaining: []
  regressions: []
---

# Phase 6: Система автокоррекции — Verification Report (Re-verification)

**Phase Goal:** Опциональная (default off) автокоррекция — молчаливое исправление слова, набранного не в той раскладке, без горячей клавиши. Безопасность гибридным детектором (словарь + ремап + словарь; триграммный fallback для OOV) и политикой app×role на живых AT-SPI-данных (unknown → молчание). SPEC §10/§11 правится spec-delta внутри фазы (D-51: milestone v1.1.0).
**Verified:** 2026-10-04T16:45:00Z
**Status:** passed
**Re-verification:** Yes — after gap closure (06-09/06-10) + UAT complete (bea10b9)

## Свежесть (что изменилось с прошлой верификации)

Предыдущий отчёт (2026-10-02T20:15:00Z, human_needed, 42/44) предшествовал gap-планам G-6-1 и UAT. Эта ре-верификация перепроверила must-haves на ТЕКУЩЕМ дереве (HEAD 9dd0de8) first-hand:

- **G-6-1 закрыт:** `AttachEngine` — один атомарный `a.engSlot.Store(&emitterSlot{...})` без `a.mu` (internal/session/actor.go:770, слот :161, снапшот-аксессор :916); ADR-006 несёт «Amendment 2026-10-02 — G-6-1» (:14, :227); e2e-кейс flip-keystroke зарегистрирован (test/e2e/main.go:271, mise.toml:145); живое доказательство 24/24, RTT 6.5–9.2 мс (06-10-SUMMARY).
- **UAT 5/5, статус complete** (06-UAT.md, bea10b9): все 5 тестов — с вердиктами владельца 2026-10-04.
- **ADR-007: «Accepted — владелец, verify-work фазы 6, 2026-10-04»** на дереве (87d5650).
- **Nyquist:** 06-VALIDATION.md `status: validated, nyquist_compliant: true` (9b25bd7).
- **Security:** 06-SECURITY.md `status: verified, threats_open: 0, asvs_level: 1` (9dd0de8); D-20/D-21 перепроверены на каждом новом эммиссионном сайте.
- **Review:** 06-REVIEW-FIX.md `status: all_fixed` (CR-01/CR-02/WR-01, RED→GREEN).
- **super-space-alive выведена из матрицы** владельцем (1eda53d; SANCTIONED RETIREMENT 2026-10-04 в matrix-v4.yaml:5) — одноисточниковая модель, владение переключением у демона доказано e2e-flip-keystroke.

## Goal Achievement

### Success Criteria (ROADMAP) — вердикты

| # | Критерий | Вердикт | Evidence (свежий) |
|---|----------|---------|----------|
| 1 | Гибридный детектор: точность на корпусе, FP в бюджете | ✓ VERIFIED | `go test ./internal/detect/ -race` ok (прогнан мной 2026-10-04); golden-корпус: 190 уверенных, 404 легитимных слова — 0 FP, OOV никогда не уверены (провёрено в прошлой верификации, пакет не менялся с тех пор — fingerprint-файлы detect не в gap-дифе) |
| 2 | Политика app×role, unknown → молчание | ✓ VERIFIED | `go test ./internal/session/ -race` ok (прогнан мной); silence-matrix 10 ячеек fail-closed; живые негативы — UAT тест 4: password-silent (witness 0→7, fired=0), terminal-silent (fired=0 при обеих доставках IME) — зелёные 2026-10-04 |
| 3 | Default off, hot reload, счётчики в status, лог без содержимого | ✓ VERIFIED | Тесты зелёные в полном CI; renderStatus-токены на дереве (ctlsvc.go:160 `autocorrect_enabled=`, ac_skip_*); privacy D-20/D-21 — 0 открытых угроз (06-SECURITY.md, перепроверка на HEAD) |
| 4 | Двойной Shift — ручное переопределение | ✓ VERIFIED | TestActor_ManualOverrideAfterAutocorrect в зелёном пакете; UAT тест 4: fires — исправление без хоткея + ручной Double конвертирует обратно, зелёный 2026-10-04 на починенном дереве |
| 5 | e2e-матрица расширена автокоррекционными кейсами, ночная двойная зелёная | ✓ VERIFIED (вердикт владельца) | Расширение — на дереве (matrix-v4.yaml, ночной конвейер). Двойная зелень: корень красных строк был G-6-1 (UAT-измерение: потеря буквы внутри окна переключения, WARN deadline на каждом флипе) — закрыт 06-09/06-10; после фикса **32/35 ×2 подряд идентично, все 9 дрейф-строк зелёные**; остаток — word/phrase-mixed (замороженные WINDOWS #12 строки) с вердиктом владельца: целевая семантика = посимвольная инверсия раскладки, бэклог v1.1.x; super-space-alive выведена владельцем (1eda53d). Вердикт: «Владелец 2026-10-04: принял» (06-UAT.md тест 1) — human-гейт закрыт владельцем |

### Gap-closure G-6-1 — must-haves планов 06-09/06-10 (7/7, новое)

| # | Truth | Status | Evidence (прогнано/прочитано мной на дереве 2026-10-04) |
|---|-------|--------|----------|
| 1 | 06-09 D1: lock-free AttachEngine — реентерабельный CreateEngine фабрики завершается во время флипа без a.mu | ✓ VERIFIED | Код: `AttachEngine` = один `engSlot.Store` (actor.go:770), ноль Lock в теле; 25 чтений на снапшотах через `emitter()`. Поведенческие тесты **прогнаны мной под -race: TestActor_ReentrantAttachDuringFlip, TestActor_AttachEngineWhileFlipInFlight — PASS** |
| 2 | 06-09 D2: контракт флипа не тронут (D-36 порядок, 150 мс wedge-guard, WARN-not-fatal, nil seam, автокоррекционный контур) | ✓ VERIFIED | Полный `mise run ci` **прогнан мной 2026-10-04: зелёный** (build+vet+lint strict+test -race, все пакеты ok — включая internal/session) |
| 3 | 06-09 D3: wire-свидетели фаб-шины — синхронный реентерабельный CreateEngine внутри await; блокирующий хендлер — единственный гейт | ✓ VERIFIED | **TestFactoryReentrantCreateEngineAnswersDuringAwait + …BlockingHandlerIsOnlyGate — прогнаны мной под -race: PASS**; production engine/ не тронут (git diff gap-коммитов: только тест-файл) |
| 4 | 06-09 D4: ADR-006 Amendment 2026-10-02 (диагноз, свойство, решение, что НЕ меняется, улики) | ✓ VERIFIED | Amendment на дереве (ADR-006 :14, секция :227); диагноз подтверждён журналом UAT (WARN на каждом флипе, engine created +1 мс после аборта); живая улика дописана 06-10 (24/24) — честность правки подтверждена владельцем через UAT тест 1 |
| 5 | 06-10 D1: e2e-кейс flip-keystroke зарегистрирован, mise-задача вне [tasks.ci], оракул немедленной буквы | ✓ VERIFIED | main.go:271 (`"flip-keystroke": {fn: runFlipKeystroke, standalone: true}`), main.go:113/286, mise.toml:145; runFlipKeystroke в case_switch.go (grep = 2); оракул не ослаблен (T-06-10-01: паузы только ПОСЛЕ буквы) |
| 6 | 06-10 D2: живое доказательство double-green + регрессия пары + журнальный аудит | ✓ VERIFIED | 24/24 немедленных буквы через 4 прогона (6/6 ×4; pre-fix база — 5/5 потерь), two-source-flip зелёный, аудит: 8/8 switch_engine INFO, RTT 6.50–9.16 мс (медиана 8.32), ноль WARN (06-10-SUMMARY Live Results; запуск e2e — живой стол, вне компетенции верификатора — числа приняты по SUMMARY+UAT с коррелирующим вердиктом владельца) |
| 7 | 06-10 D3: fresh-session вердикт WINDOWS #13 — гейт владельца | ✓ VERIFIED (закрыт владельцем) | UAT тест 1: «Владелец 2026-10-04: принял. Матрица 32/35×2 стабильно» — вердикт записан в 06-UAT.md, гейт больше не открыт |

### Observable Truths исходных 8 планов (44) — регрессионная перепроверка

42 истины, верифицированные 2026-10-02, — quick-regression (существование + санити) на текущем дереве: все артефакты на месте (matrix-v4.yaml, case_autocorrect.go, fixtures, appid/detect/config/buffer/actor/ctlsvc — ARTIFACTS-OK), греп-гейты SPEC-DELTA-OK и AUDIT-TRAIL-PRESENT **перезапущены мной — зелёные**, статусные токены на месте (ctlsvc.go:160). Пакеты internal/session и internal/detect прогнаны под -race — ok. Два бывших непокрытых must-have:

| # | Truth | Прежний статус | Новый статус | Evidence |
|---|-------|----------------|--------------|----------|
| 43 | 06-08 T2: полный регресс — matrix-v4 зелёный дважды подряд | ⚠️ PRESENT_BEHAVIOR_UNVERIFIED | ✓ VERIFIED | Корень найден и закрыт (G-6-1 → 06-09/06-10); после фикса 32/35 ×2 идентично, дрейф-строки зелёные; вердикт владельца 2026-10-04 в 06-UAT.md тест 1 (детали — критерий 5) |
| 44 | 06-08 T3: ADR-007 переведён в Accepted | ⚠️ HUMAN GATE | ✓ VERIFIED | ADR-007:5 «**Accepted — владелец, verify-work фазы 6, 2026-10-04**»; commit 87d5650; строка 5 Acceptance Evidence закрыта фиксом G-6-1 |

**Score: 51/51 truths verified** (44 исходных + 7 gap-closure; 0 present-behavior-unverified; 0 FAILED; 0 overrides)

## Resolved Human Verification Ledger (все 5 прежних пунктов)

| # | Бывший human-пункт | Резолюция | Evidence |
|---|--------------------|-----------|----------|
| 1 | WINDOWS #13 / двойной зелёный + вердикт по дрейф-строкам | RESOLVED — владелец принял | 06-UAT.md тест 1: pass; матрица 32/35×2 после G-6-1; mixed-semantics → v1.1.x; super-space-alive retired (1eda53d) |
| 2 | ADR-007 Proposed → Accepted | RESOLVED — владелец принял | 06-UAT.md тест 2: pass; ADR-007:5 Accepted с датой (87d5650) |
| 3 | CR-01 асинхронное окно подтверждения — живая проверка | RESOLVED — три инструмента | 06-UAT.md тест 3: pass; юнит-корпус -race зелёный + живая проба 18/18 (0..100 мс, ноль порчи, честные abstain) + fires-кейс зелёный на починенном дереве |
| 4 | Живой UAT трёх автокоррекционных кейсов | RESOLVED — все три зелёные | 06-UAT.md тест 4: pass; fires (ghbdtn→привет без хоткея + Double-регресс), password-silent (witness 0→7, fired=0, FINAL-verbatim), terminal-silent (fired=0 при обеих доставках) |
| 5 | Judgment-запреты (4 позиции) | RESOLVED — владелец подтвердил | 06-UAT.md тест 5: pass; «да — все четыре подтверждены» (audit-trail дословность, GPL-чистота dictgen, cache-not-basis D-53, фикстура-без-утечки D-20/D-21) |

Все пять UAT-тестов — с вердиктами владельца 2026-10-04; ни один не остаётся human-gated. Frontmatter `unverified-prohibitions` прежнего отчёта снят: 4/4 подтверждены владельцем (тест 5), privacy-запреты перепроверены 06-SECURITY.md на HEAD.

## Required Artifacts (прирост с прошлой верификации)

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| internal/session/actor.go (gap-правка) | lock-free AttachEngine + engSlot | ✓ VERIFIED | actor.go:118-121/:154-161/:770/:916-917; ноль Lock в теле AttachEngine |
| engine/conn_switcher_test.go (gap-правка) | reattach-хук + 2 wire-свидетеля | ✓ VERIFIED | Тесты прогнаны мной под -race: PASS |
| docs/adr/ADR-006-two-engine-revision.md | Amendment 2026-10-02 G-6-1 | ✓ VERIFIED | :14 + секция :227 |
| test/e2e/case_switch.go | runFlipKeystroke + tail-оракул | ✓ VERIFIED | runFlipKeystroke ×2 (определение+регистрация) |
| test/e2e/main.go / mise.toml | регистрация кейса | ✓ VERIFIED | main.go:113/:271/:286; mise.toml:145 |
| test/e2e/cases/matrix-v4.yaml | retirement super-space-alive | ✓ VERIFIED | SANCTIONED RETIREMENT 2026-10-04 (yaml:5, commit 1eda53d) |
| 06-UAT.md / 06-VALIDATION.md / 06-SECURITY.md / 06-REVIEW-FIX.md | сводные артефакты приёмки | ✓ VERIFIED | complete 5/5 / validated+nyquist / verified+0 threats / all_fixed |

Остальные 13 артефактов прежнего отчёта — quick-regression: все на месте (ARTIFACTS-OK), не менялись gap-коммитами (кроме перечисленных).

## Behavioral Spot-Checks (свежие прогоны этой верификации)

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Полный CI на HEAD | `mise run ci` (2026-10-04) | build+vet+lint+test -race — все пакеты ok, 0 FAIL | ✓ PASS |
| G-6-1 реентераб-корпус | `go test ./internal/session/ -race -run 'TestActor_ReentrantAttachDuringFlip\|TestActor_AttachEngineWhileFlipInFlight' -v` | 2/2 PASS | ✓ PASS |
| Wire-свидетели фабрики | `go test ./engine/ -race -run TestFactoryReentrantCreateEngine -v` | 2/2 PASS | ✓ PASS |
| Spec-delta гейты | grep v1.1.0+2026-09-27+ADR-007, !«явно ВНЕ объёма»; grep «отвергнута владельцем осознанно» | SPEC-DELTA-OK, AUDIT-TRAIL-PRESENT | ✓ PASS |
| Детектор/сессия -race | `go test ./internal/session/ ./internal/detect/ -race -count=1` | оба ok | ✓ PASS |
| Статусные токены | grep ctlsvc.go | autocorrect_enabled (:160) + ac_skip_* | ✓ PASS |

Пропущено: живые e2e-прогоны (flip-keystroke, матрица) — требуют живой GNOME-стол; живые числа покрыты парой «06-10 SUMMARY + коррелирующий вердикт владельца в 06-UAT.md» и не оспариваются.

## Probe Execution

Не применимо — проект не объявляет scripts/*/tests/probe-*.sh; планы фазы (вкл. 06-09/06-10) не декларируют пробники. Носители — mise-задачи и go test (выше).

## Requirements Coverage

| Requirement | Source Plan(s) | Status | Evidence |
| ----------- | -------------- | ------ | -------- |
| SPEC §10/§11 (spec-delta), D-51 | 06-01..08 | ✓ SATISFIED | SPEC-DELTA-OK/AUDIT-TRAIL перезапущены; ADR-007 Accepted владельцем |
| CORR-01..09 | 06-02..08 | ✓ SATISFIED | Полный CI зелёный на HEAD; UAT тест 4 (fires/password/terminal) — живые подтверждения; прежние носители не регрессировали |
| MACR-ACL (app×role) | 06-01,03,04,06,07 | ✓ SATISFIED | silence-matrix в зелёном пакете; UAT тест 4: password-silent (одна role-forbidden абстенция) и terminal-silent живьём |
| SWCH-01/SWCH-02 (G-6-1) | 06-09, 06-10 | ✓ SATISFIED | lock-free фабрика + reentrant-корпус + wire-свидетели (прогнаны мной); живое доказательство 24/24; two-source-flip регрессия зелёная |

Orphaned requirements: НЕТ (как и прежде — REQUIREMENTS.md заморожен до complete-milestone по D-51; формализация REQ-ID — при new-milestone v1.1.0).

## Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| — | — | Debt markers (TBD/FIXME/XXX/HACK) в gap-файлах (actor.go, case_switch.go, conn_switcher_test.go, matrix-v4.yaml, ADR-006, ADR-007) | none | скан 2026-10-04 — 0 совпадений |
| — | — | Placeholder / stub / empty impl | none | 0; полный CI зелёный |
| — | — | go.mod churn | none | tidy CLEAN в CI |

Info: ROADMAP.md:293 всё ещё показывает `- [ ]` для 06-10-PLAN.md при полном 06-10-SUMMARY (status: complete) и feat-коммите 9f78017 в истории — устаревшая галочка буккипинга, на цель фазы не влияет (рекомендация: отметить при complete-milestone). Преждевременный ручной флип `status: human_needed → passed` в незакоммиченном 06-VERIFICATION.md (без улик) заменён настоящим отчётом с full-уликами.

## Deferred Follow-Ups (бэклог владельца v1.1.x — не гэпы фазы)

Оба пункта записаны владельцем в 06-UAT.md § Deferred Follow-Ups (canonical ledger); поздних фаз текущего милстоуна нет (фаза 6 — последняя), в гэпы не включаются.

1. **Семантика смешанного текста** (из UAT теста 1, решение владельца 2026-10-03): конвертация инверсией раскладки посимвольно — целевая семантика для word-mixed/phrase-mixed; WINDOWS #12-фриз отклонён; отдельный план, строки матрицы перезакрепить. Бэклог v1.1.x.
2. **UX white-list identity-discovery** (из UAT теста 3): goswitchctl status дополнить токеном последней наблюдаемой bridge-идентичности приложения, чтобы список заполнялся без кейс-харнесса (18×ac_skip_app_not_listed при документационной догадке). Бэклог v1.1.x.

## Decision Coverage

Прежняя проверка gsd-tools check.decision-coverage-verify: 5/5 honored (запись 2026-10-02). Повторный вызов в этой сессии скипается вербом («CONTEXT.md missing» — проектный лэйаут); новых trackable CONTEXT-решений gap-планы не вводили (их key-decisions — планового уровня, записаны в 06-09/06-10 SUMMARY). Переносится как 5/5, 0 not honored.

## Gaps Summary

Гэпов нет. Обе прежние «незакрываемые» истины закрыты: 06-08 T2 — фиксацией корня (G-6-1: самоблокировка фабрики флипа) планами 06-09/06-10 с живым доказательством 24/24 и вердиктом владельца (32/35×2 после фикса); 06-08 T3 — прямым вердиктом владельца (ADR-007 Accepted, 2026-10-04). Все 5 human-пунктов разрешены (06-UAT.md, 5/5, owner verdicts). 4 judgment-запрета подтверждены владельцем. 0 FAILED, 0 STUB, 0 ORPHANED, 0 NOT_WIRED, 0 behavior-unverified. Цель фазы достигнута на текущем дереве: полный `mise run ci` зелёный (прогнан верификатором 2026-10-04).

---

_Verified: 2026-10-04T16:45:00Z_
_Verifier: Claude (gsd-verifier)_
