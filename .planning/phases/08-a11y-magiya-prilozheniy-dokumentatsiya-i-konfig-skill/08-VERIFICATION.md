---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
verified: 2026-10-06T05:53:08Z
status: passed
score: 44/44 must-haves verified
covered_files:
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-01-PLAN.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-01-SUMMARY.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-02-PLAN.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-02-SUMMARY.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-03-PLAN.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-03-SUMMARY.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-04-PLAN.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-04-SUMMARY.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-05-PLAN.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-05-SUMMARY.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-06-PLAN.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-06-SUMMARY.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-07-PLAN.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-07-SUMMARY.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-08-PLAN.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-08-SUMMARY.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-09-PLAN.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-09-SUMMARY.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-10-PLAN.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-10-SUMMARY.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-CONTEXT.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-UAT.md
  - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-VALIDATION.md
  - README.md
  - README.en.md
  - cmd/goswitchd/main.go
  - cmd/skillgen/main.go
  - cmd/skillgen/main_test.go
  - docs/ACCEPTANCE.md
  - docs/CONFIG.md
  - docs/SPEC.md
  - internal/a11y/a11y.go
  - internal/a11y/a11y_test.go
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/config/load_test.go
  - internal/config/watch_test.go
  - internal/install/install.go
  - internal/install/install_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - skills/goswitch-config/SKILL.md
covered_digest: "v1:sha256:73560cca1bbb9c8f144cfe94556ac91d905516e5c6da9781fe9213892f474342"
behavior_unverified: 0
overrides_applied: 0
coincidental_reliance_items: []
re_verification:
  previous_status: gaps_found
  previous_score: 43/44
  gaps_closed:
    - "08-07 #4 (README.md:303–320 / README.en.md:300–313 documented the annulled a11y.apps contract) — RESOLVED by gap plan 08-10 (commits d1e44a4 RU, 8df7162 EN, 9174109 gates+VALIDATION; SUMMARY 3ea1731): both a11y sections rewritten to the bool contract (single key a11y.enabled, default ON, global mechanism, restart semantics, uninstall-revert, loud migration); zero a11y.apps references in either README (verifier grep); wording mirrors docs/CONFIG.md:227–262; parity 22/22 (verifier-counted); diff confined to the old section bounds (RU 306–319 ⊂ 301–320, EN 301–313 ⊂ 296–314, verifier awk-gate on BASE 9660977); skillgen -check exit 0 and mise run ci exit 0 (both verifier runs)"
  gaps_remaining: []
  regressions: []
deferred: []
advisory:
  - finding: "Housekeeping — the three folded source todos (a11y-magic-apps-list.md, config-skill-for-ai.md, docs-accuracy-and-russian-first.md) remain in .planning/todos/pending/ although 08-CONTEXT records them as folded into this phase (a11y-apps-matching-semantics.md is gone as the UAT records). Not a phase artifact; no effect on the verdict."
    category: other
    reason: "Process hygiene only; flagged so the todo board matches the phase records."
    evidence_status: "ls .planning/todos/pending/ (re-checked 2026-10-06T05:53Z — still present)"
---

# Phase 8: a11y-магия приложений, документация и конфиг-skill — Verification Report (FINAL — гэп закрыт, фаза пройдена)

**Phase Goal:** Разобранный беклог (3 todo от 2026-10-05), ПЕРЕСМОТРЕННЫЙ решением владельца 2026-10-06 (D-8-1/REV, вербатим: «раз настройка глобальная, то список не нужен, а нужен bool параметр, по дефолту настройка включен»): (1) a11y-магия — конфиг-секция с ЕДИНСТВЕННЫМ ключом `a11y.enabled` (bool, **default ON**); демон применяет сам (reconcile старт/reload; uninstall-revert; список приложений и сопоставление АННУЛИРОВАНЫ — D-8-5/REV). (2) Документация — сверка + русский-первый. (3) Skill для AI — SKILL.md генерируется из docs/CONFIG.md.
**Verified:** 2026-10-06T05:53:08Z
**Status:** passed
**Re-verification:** Yes — третий проход: после гэп-плана 08-10 (README a11y-разделы → bool-контракт); история: 44/46 human_needed (2026-10-05T23:55Z) → 43/44 gaps_found (2026-10-06T01:55:43Z, commit cbe4160) → **44/44 passed** (этот отчёт)

## Gap Resolution (2026-10-06)

Единственный гэп пост-дельта реверификации (README.md:303–320 / README.en.md:300–313 — аннулированный контракт `a11y.apps`) **закрыт в дереве и верифицирован прогонами этого верификатора**, не по CLAIM'ам SUMMARY:

- **Ноль `a11y.apps`** в README.md и README.en.md (`grep -c` = 0 в обоих; заодно ноль прозы списка: «перечисляет приложения», «добавления его в список», «adding it to the list», «Removing an application» — единственный матч «из списка» в README.md:412 относится к чапту troubleshooting «исчез из списка источников ввода» и к a11y не относится).
- **Bool-контракт в обоих разделах:** README.md:301–335 и README.en.md:296–331 — единственный глобальный переключатель `a11y.enabled` (bool, default ON: отсутствующая секция/ключ читаются ВКЛ; явное `enabled: false` — единственный OFF, D-8-4); глобальность обоснована общим для стола `toolkit-accessibility` («сопоставлений приложений в секции нет»); идемпотентный reconcile при старте и на каждой горячей перезагрузке (ключ, уже равный `true`, не переписывается — ноль записей в dconf); restart-семантика («при СВОЁМ следующем запуске» / «at their next start»); запрет снятия демоном; uninstall-revert из снапшота установки (`goswitchctl uninstall`); громкая миграция строки списка (конфиг целиком отвергается — удалите её).
- **Зеркало docs/CONFIG.md:227–262** по порядку фактов и дословным формулировкам («отсутствующая секция или ключ читаются включёнными, явное `enabled: false` — единственный способ выключить», «конфиг целиком отвергается при загрузке (строгий декодер называет ключ ошибкой)», «ключ, уже равный `true`, не переписывается — ноль записей в dconf», «Снятие ключа демоном не предусматривается», «единственный путь отката — `goswitchctl uninstall`»); завершающая ссылка обоих языков называет «единственный ключ `a11y.enabled`» и ведёт в docs/CONFIG.md.
- **Паритет и кросс-ссылки целы:** 22/22 заголовка в обоих файлах (посчитано верификатором), структура уровней идентична; README.md:3 `[English](README.en.md)`, README.en.md:3 `[Russian](README.md)`.
- **Скоуп без дрейфа:** `git diff d9c2909..HEAD -- README.md README.en.md` — хунки только внутри прежних границ a11y-разделов (RU старые 306..311 и 314..319 ⊂ 301–320; EN старый 301..313 ⊂ 296–314; awk-гейт верификатора на BASE=gsd-plan-head-before-08-10=9660977: RU-SCOPE-OK, EN-SCOPE-OK). С момента верификации 43/44 (cbe4160) менялись ТОЛЬКО README.md, README.en.md и planning-метаданные (ROADMAP, STATE, 08-10-PLAN/SUMMARY, 08-VALIDATION) — `git diff --name-only cbe4160..HEAD`; изменения internal/ в диапазоне d9c2909..HEAD — это коммиты TDD 08-09 (d71bbf6→87c815c), ПРЕДШЕСТВУЮЩИЕ cbe4160 и верифицированные предыдущим отчётом.
- **Гейты зелёные в прогоне верификатора:** `go run ./cmd/skillgen -check` exit 0; `mise run ci` exit 0 (build + vet + golangci-lint + test -race, все пакеты ok); строки 08-10 T1..T3 присутствуют в Per-Task Verification Map 08-VALIDATION.md (:63–65).

UAT-вердикты владельца (08-UAT.md, 2026-10-06) подтверждены на месте — все 6 human-пунктов остаются разрешёнными (раздел «UAT Dispositions»), новых human-пунктов не возникло.

## Goal Achievement

Дельта 08-09 (ревизия владельца) верифицирована в дереве предыдущим проходом; её единственный недочёт — непропаганированная правка README — устранён гэп-планом 08-10 и верифицирован этим проходом. Спека add-only с дословной цитатой, pointer-bool nil=ON по прецеденту Sound, громкая миграция запинена тестом, фолд на EffectiveEnabled, reconciler (comments-only diff) и installer нетронуты, skillgen синхронен и идемпотентен, `mise run ci` зелёный, README-пара честна в обоих языках.

**Score:** 44/44 in-force truths verified (7 former truths superseded-and-honored by the revision — recorded below, excluded from the denominator; 0 FAILED; 0 behavior-unverified — все human-пункты разрешены UAT-вердиктами владельца).

### Observable Truths

#### Plan 08-01 — SPEC §4 spec-delta (6 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SPEC §4 несёт датированную ревизию 2026-10-05 ДО кода; аудит-трейл цел | ✓ VERIFIED | Блок 2026-10-05 нетронут в дереве; аудит-маркеры после дельты: 2026-10-05 ×10, 2026-09-27 ×9, 2026-10-04 ×5, D-53 ×5 |
| 2 | Ревизия фиксирует секцию a11y (enabled/apps, потолок 64, zero-off) | ⚠️ SUPERSEDED-AND-HONORED | Аннулировано D-8-1/REV владельцем; блок 2026-10-05 сохранён в SPEC как аудит-трейл (намеренно); действующий контракт — ревизия 2026-10-06 (замена верифицирована в 08-09 #1) |
| 3 | Ревизия фиксирует daemon-reconcile (старт+reload), отказ от динамического снятия, uninstall-revert only-if-present | ✓ VERIFIED | Семантика подтверждена ревизией 2026-10-06 (SPEC.md:247–253: «динамическое снятие по-прежнему запрещено — D-8-4», «контракт плана 08-04 не меняется»); internal/install вне дифа 08-09 |
| 4 | Ревизия фиксирует restart-семантику | ✓ VERIFIED | SPEC.md:254–257 «Restart-семантика … сохраняются без изменений» |
| 5 | Пер-апп оверрайды — документируемый рецепт + argv-дисциплина | ✓ VERIFIED | Рецепт сохранён (CONFIG.md:335,359); argv-дисциплина в ревизии 2026-10-06 («список не может попасть в argv по построению») |
| 6 | Spec-less fallback пропущен видимо | ✓ VERIFIED | grep REQUIREMENTS.md: ноль строк фазы 8; трассировка через 08-CONTEXT D-8-1..D-8-11 + /REV |

#### Plan 08-02 — Схема секции a11y (5 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Секция A11y (yaml a11y: enabled/apps), zero-value off | ⚠️ SUPERSEDED-AND-HONORED | apps удалён владельцем; замена верифицирована: config.go:184–187 `A11y{Enabled *bool}`, nil = ON (08-09 #3) |
| 2 | Валидация на Load: потолок 64, blank-отказ, errA11yApps сентинелы | ⚠️ SUPERSEDED-AND-HONORED | D-8-5/REV аннулирована; grep maxA11yApps/errA11yApps = 0 в internal/config; диспетч Validate без a11y-вызова с поясняющим комментарием (D-8-5/REV annulled) |
| 3 | Active() = enabled && непустой список; Defaults() выключенная секция | ⚠️ SUPERSEDED-AND-HONORED | Замена: `EffectiveEnabled()` (config.go:195, nil = ON) — единственный метод секции (`func (a A11y)` ровно 1); `Defaults()`: `A11y{Enabled: boolPtr(true)}` (config.go:266–268); TestDefaults_A11yDefaultOn PASS |
| 4 | Strict decode D-33 и last-good автоматически; load.go/watch.go — ноль правок | ✓ VERIFIED | load.go/watch.go отсутствуют в дифе 08-09 (`git diff gsd-plan-head-before-08-09..HEAD --name-only`); TestLoad_A11yStrictDecodeUnknownKey + TestWatch_A11yUnknownKeyKeepsLastGood PASS |
| 5 | Hot reload пин: битый паттерн → reload отклонён, last-good держится | ⚠️ SUPERSEDED-AND-HONORED | Триггер заменён владельцем (паттернов больше нет); свойство D-32/D-33 запинено заново: TestWatch_A11yUnknownKeyKeepsLastGood (неизвестный ключ в a11y → отказ, last-good держится) — PASS |

#### Plan 08-03 — Reconciler internal/a11y (6 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Apply(active bool) fire-and-forget, сериализация вне акторного мьютекса | ✓ VERIFIED | Диф a11y.go в 08-09 — ТОЛЬКО комментарии (гейт A11Y-COMMENTS-ONLY вычислен: код не менялся); регресс-корпус CR-01 зелёный в mise run ci |
| 2 | Read-verify-then-set; Apply(false) — ноль подпроцессов | ✓ VERIFIED | Код нетронут; пакет internal/a11y ok под -race |
| 3 | Пояс org.a11y.Status.IsEnabled при каждой активации | ✓ VERIFIED | Код нетронут; тесты зелёные |
| 4 | Argv-дисциплина: argv только из package-литералов | ✓ VERIFIED | Код нетронут; ревизия 2026-10-06 подтверждает («по построению») |
| 5 | Degradation: закрытый reason-словарь, WARN once, переоткрытие бюджета | ✓ VERIFIED | Код нетронут; тесты зелёные |
| 6 | Сериализация/сходимость под -race | ✓ VERIFIED | internal/a11y ok под -race в полном прогоне |

#### Plan 08-04 — Installer snapshot/revert (5 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | installState.ToolkitAccessibility, verbatim; отказ чтения → пусто+WARN | ✓ VERIFIED | internal/install ОТСУТСТВУЕТ в дифе 08-09 (prohibition «install untouched» вычислен); regression: пакет ok |
| 2 | Idempotent-backup не тронут | ✓ VERIFIED | Файл не менялся с прежней верификации |
| 3 | Restore only-if-present + shape-валидация, до удаления state-файла | ✓ VERIFIED | Файл не менялся; тесты зелёные |
| 4 | Argv-дисциплина restore: shape-валидированный литерал | ✓ VERIFIED | Файл не менялся |
| 5 | Место в цепочке uninstall | ✓ VERIFIED | Файл не менялся |

#### Plan 08-05 — Фолд актора + wiring (4 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Фолд через snap.A11y.Active() с дифф-гейтом | ⚠️ SUPERSEDED-AND-HONORED | Замена верифицирована: pushA11y читает `snap.A11y.EffectiveEnabled()` (actor.go:1335), дифф-гейт a11yActive сохранён в коде; корпус: TestActor_A11yDefaultsDocumentPushesTrue / FoldTriggersSinkOnActivation / ExplicitFalsePushesFalse — PASS |
| 2 | A11ySink seam + SetA11ySink self-sync; nil no-op | ✓ VERIFIED | actor.go seam нетронут (08-09 правил только doc-комменты и точку чтения); тесты SinkInstallSelfSyncs / NilSinkNoOp PASS |
| 3 | Wiring демона на production-адаптерах | ✓ VERIFIED | main.go:201 `actor.SetA11ySink(a11y.New(a11y.NewExecRunner(), a11y.NewDBusStatusSetter()))` |
| 4 | Двойная идемпотентность (дифф-гейт актора + reconciler) | ✓ VERIFIED | Оба гейта в коде; тесты зелёные |

#### Plan 08-06 — CONFIG.md русский-первый + SPEC-сверка (7 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | CONFIG.md русский-первый | ✓ VERIFIED | Весь файл по-русски; дельта сохранила язык |
| 2 | Построчная сверка таблицы со схемой | ✓ VERIFIED | Таблица синхронна новой схеме: ровно одна a11y-строка (bool/true) — TestSkillgen_ParseRealConfigTable (20 строк) PASS; сверка-правки 08-06 на месте |
| 3 | SPEC-сверка в форме 08-01, аудит-трейл цел | ✓ VERIFIED | Сверка-блок 08-06 нетронут (прочитан в дереве); ревизия 2026-10-06 добавлена рядом; numstat SPEC: 39/0 (только добавления) |
| 4 | «Ровно семь секций» + ровно 2 a11y-строки | ⚠️ SUPERSEDED-AND-HONORED | Семь секций — правда (CONFIG.md:14, схема Config: 7 полей); строк a11y теперь ровно 1 (решение владельца); замена в 08-09 #6 |
| 5 | Поведение магии задокументировано честно | ✓ VERIFIED (WR-02 RESOLVED) | CONFIG.md:227–262: «Сопоставления приложений в секции нет: ключ toolkit-accessibility общий для стола», default-ON семантика, громкая миграция (:245–251), restart-семантика, uninstall-revert; строка :77 совпадает с рантаймом |
| 6 | Per-app .desktop-оверрайды — ручной рецепт без кода | ✓ VERIFIED | CONFIG.md:335,359 сохранены |
| 7 | Формат таблицы байт-строгий (источник генератора) | ✓ VERIFIED | skillgen -check exit 0; TestSkillgen_CommittedFileInSync PASS |

#### Plan 08-07 — README русский-первый (5 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | README.md — русская главная + a11y-раздел | ✓ VERIFIED | Русский; a11y-раздел переписан гэп-планом 08-10 (d1e44a4, README.md:301–335) на bool-контракт — см. #4 ✓ |
| 2 | README.en.md — английский канон, паритет 22/22 | ✓ VERIFIED | Английское зеркало раздела (8df7162, README.en.md:296–331); паритет заголовков 22/22 — ПОСЧИТАН верификатором после правки |
| 3 | README.ru.md удалён, дублей/stale-ссылок нет | ✓ VERIFIED | git ls-files пуст; grep README\.ru — ноль |
| 4 | a11y-раздел в README с restart-семантикой и uninstall-откатом, без дубля справочника | ✓ VERIFIED (RESOLVED 2026-10-06 планом 08-10) | Раздел описывает действующий контракт ревизии 2026-10-06: единственный ключ `a11y.enabled` (bool, default ON — отсутствующая секция/ключ = ВКЛ, явное `enabled: false` = единственный OFF, D-8-4), глобальная механика (toolkit-accessibility + пояс IsEnabled, «сопоставлений приложений в секции нет»), идемпотентный reconcile старт/reload, restart-семантика, uninstall-revert, громкая миграция (README.md:331–333 / README.en.md:327–329). Ноль `a11y.apps` в обоих файлах (grep верификатора); зеркало docs/CONFIG.md:227–262 по порядку фактов; паритет 22/22; кросс-ссылки шапок целы. Прежний провал (аннулированный контракт + config-bricking инструкция) устранён; см. Gap Resolution |
| 5 | Входящие ссылки перепoint'нуты | ✓ VERIFIED | docs/ACCEPTANCE.md ссылки целы (файл не менялся дельтой) |

#### Plan 08-08 — Skill goswitch-config (6 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SKILL.md frontmatter + RU description + ручной каркас | ✓ VERIFIED | Frontmatter name goswitch-config, RU description; каркас переехал в bool-семантику (:77–100) |
| 2 | Single source: генератор парсит таблицу, verbatim, hard errors | ✓ VERIFIED | cmd/skillgen/main.go не менялся; TestSkillgen_MalformedRowHardError PASS |
| 3 | Синхрон-гейт: -check байт-в-байт + golden-тест | ✓ VERIFIED | `go run ./cmd/skillgen -check` → CHECK-EXIT-0 (прогон верификатора, повторён 2026-10-06T05:53Z) |
| 4 | Dev-only mise-задача skillgen-regen; CI не перегенерирует | ✓ VERIFIED | mise.toml не менялся дельтой |
| 5 | TRACKING-гейт: SKILL.md в git-индексе | ✓ VERIFIED | git ls-files непуст |
| 6 | Регенерация идемпотентна | ✓ VERIFIED | regen → файл байт-идентичен (REGEN-IDEMPOTENT, проверено верификатором cmp) |

#### Plan 08-09 — Дельта-ревизия владельца (7 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SPEC §4 датированная ревизия 2026-10-06 в форме 08-01 с дословной цитатой; прежние блоки целы | ✓ VERIFIED | SPEC.md:226–247 (HTML-пометка + датированный абзац); цитата :235 «…список не нужен, а нужен bool параметр, по дефолту настройка включен»; аудит-маркеры 2026-10-05/2026-09-27/2026-10-04/D-53 целы; numstat 39/0 — add-only |
| 2 | Схема без Apps-машинерии; единственный метод EffectiveEnabled; Validate без a11y-вызова | ✓ VERIFIED | config.go:184–187 (`Enabled *bool` — единственное поле); `func (a A11y)` ровно 1; maxA11yApps/errA11yApps = 0; диспетч Validate: явный комментарий «a11y section carries no validation» |
| 3 | Default ON запинен (pointer-bool по прецеденту Sound; НЕ D-54); Load НИКОГДА не накладывает Defaults | ✓ VERIFIED | config.go:189–202 (doc-коммент называет D-8-6/REV и Sound-прецедент); TestLoad_A11yAbsentSectionMeansDefaultOn / AbsentKeyMeansDefaultOn / ExplicitFalseDisables / TestDefaults_A11yDefaultOn — PASS; UnmarshalYAML = 0; load.go/watch.go вне дифа |
| 4 | Миграция громкая (D-33): прежний ключ-список отвергается ЦЕЛИКОМ с именем ключа; миграционная заметка в CONFIG.md | ✓ VERIFIED | TestLoad_A11yRemovedListKeyRejectedWhole — PASS (проходит только при отсутствии поля); CONFIG.md:245–251 («конфиг целиком отвергается… Удалите ключ») |
| 5 | Фолд на EffectiveEnabled (defaults → TRUE, явное false → false); контракты reconciler 08-03 не меняются; снапшот/revert 08-04 не тронут | ✓ VERIFIED | actor.go:1335; корпус фолда PASS (defaults-документ → [true], явное false → false, дифф-гейт сохранён); A11Y-COMMENTS-ONLY гейт; install вне дифа |
| 6 | Доки/skill согласованы (без ключа-списка, строка default true, regen идемпотентен, golden 20) | ✓ VERIFIED | grep a11y\.apps в CONFIG.md/SKILL.md/README.md/README.en.md = 0 (README дотянут планом 08-10); ровно одна строка `| \`a11y.enabled\| bool \| \`true\``; -check exit 0; regen идемпотентен; golden-граница 20 (main_test.go:24). Согласованность доков теперь в объёме ВСЕЙ фазы |
| 7 | mise run ci зелёный | ✓ VERIFIED | Прогон верификатора: exit 0 — build + vet + golangci-lint (0 issues) + test -race, 22 пакета ok (повторён финально 2026-10-06T05:53Z, ci-exit=0) |

#### Plan 08-10 — Гэп-закрытие README (3 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 (D1) | README.md a11y-раздел на bool-контракте ревизии 2026-10-06 (зеркало CONFIG.md), ноль ключа-списка и прозы сопоставления | ✓ VERIFIED | Прочитано в дереве: README.md:301–335; grep-гейты верификатора (негатив + якоря a11y.enabled/toolkit-accessibility/IsEnabled/enabled: false/goswitchctl uninstall/следующ(ем\|его) запуск/«единственный ключ») — PASS |
| 2 (D2) | README.en.md — тот же контракт по-английски, секция-в-секцию параллельно; паритет 22/22; кросс-ссылки целы | ✓ VERIFIED | README.en.md:296–331 — те же четыре блока в том же порядке; 22/22 заголовка в обоих (grep -cE '^##+ '); README.md:3 / README.en.md:3 кросс-ссылки — PASS |
| 3 (D3) | Скоуп-дифф внутри прежних границ разделов; skillgen -check exit 0; mise run ci зелёная; VALIDATION-строки 08-10 T1..T3 | ✓ VERIFIED | awk-гейт верификатора (BASE=9660977): RU-SCOPE-OK, EN-SCOPE-OK; skillgen -check exit 0; mise run ci exit 0 (прогон верификатора); 08-VALIDATION.md:63–65 строки на месте |

### Superseded-and-Honored Truths (плановая цепочка)

Семь истин планов 08-01..08-06 ссылались на a11y.apps-машинерию, аннулированную владельцем 2026-10-06. Они НЕ провалены — они замещены ревизией через план 08-09 (цепочка записана: 08-CONTEXT «Owner Revision», 08-09-PLAN depends_on + must_haves, ревизия SPEC 2026-10-06), и каждая замена верифицирована:

| Superseded truth | Replacement (08-09) | Verified |
|------------------|--------------------|----------|
| 08-01 #2: секция enabled/apps, потолок 64, zero-off | Ревизия SPEC 2026-10-06: единственный ключ, default ON, D-8-5/REV annulled | ✓ (08-09 #1) |
| 08-02 #1: yaml enabled/apps, zero-value off | `A11y{Enabled *bool}`, nil = ON | ✓ (08-09 #2/#3) |
| 08-02 #2: потолок/blank/сентинелы | Машинерия удалена; Validate без a11y-вызова | ✓ (08-09 #2) |
| 08-02 #3: Active() + Defaults off | EffectiveEnabled (единственный метод) + Defaults ON | ✓ (08-09 #2/#3) |
| 08-02 #5: hot-reload пин на «битый паттерн» | Пин на неизвестный ключ (то же свойство D-32/D-33) | ✓ (TestWatch_A11yUnknownKeyKeepsLastGood) |
| 08-05 #1: фолд через Active() | Фолд через EffectiveEnabled, дифф-гейт сохранён | ✓ (08-09 #5) |
| 08-06 #4: ровно 2 a11y-строки | Ровно 1 a11y-строка (bool/true) | ✓ (08-09 #6) |

### UAT Dispositions (owner verdicts 2026-10-06 — human_needed items disposed; подтверждены на месте 2026-10-06T05:53Z)

Все 6 human-пунктов прежних верификаций разрешены записанными вердиктами владельца (08-UAT.md, 2026-10-06 — Summary total: 6, passed: 6, pending: 0; раздел «Gaps» с вербатимом решения) — повторно НЕ поднимаются:

| # | Item | Verdict | Basis recorded |
|---|------|---------|----------------|
| 1 | Живой эффект a11y-магии | ACCEPTED ON EVIDENCE | Прецедент WINDOWS #5; корпуса/гейты зелёные, механика ключей live-проверена research на этом столе; остаточный риск (gsettings из контекста демона) осознанно принят |
| 2 | Живой uninstall-revert | ACCEPTED ON EVIDENCE | Тот же вердикт |
| 3 | Русская проза / пара README | ACCEPTED ON EVIDENCE | Детальное чтение — за владельцем на досуге |
| 4 | SKILL.md для AI | ACCEPTED ON EVIDENCE | Живое демо возможно в любой момент |
| 5 | WR-02 (семантика a11y.apps) | RESOLVED BY OWNER | Дословное решение D-8-1/REV → дельта-план 08-09 (исполнен и верифицирован) |
| 6 | Человеческий проход CR-01 | ACCEPTED ON EVIDENCE | Регресс-корпус воспроизводил заморозку на старом коде; оба теста зелёные под -race |

behavior_unverified: 0 — human-верификация завершена вердиктами владельца; новых human-пунктов третий проход не породил (README-правка — контентная истина, верифицированная чтением дерева и grep-гейтами).

### Review-Fix Carryover (08-REVIEW → 08-REVIEW-FIX)

| Finding | Status post-delta |
|---------|-------------------|
| CR-01 (reconciler mutex) | ✓ VERIFIED — код не менялся дельтой (comments-only), regression-корпус зелёный; UAT item 6 принят владельцем |
| WR-01 (SKILL.md switch-input-source) | ✓ VERIFIED — исправление сохранилось (SKILL.md:127 «лечится секцией a11y» без ключа-списка) |
| WR-02 (matching semantics) | ✓ RESOLVED — решение владельца исполнено дельтой; todo закрыт |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| docs/SPEC.md §4 (ревизия 2026-10-06) | internal/config (A11y) | контракт секции | ✓ WIRED | Схема = ревизии: один ключ, default ON, D-8-5/REV |
| internal/config A11y.EffectiveEnabled | internal/session/actor.go (pushA11y) | snap.A11y.EffectiveEnabled() | ✓ WIRED | actor.go:1335 — единственная точка чтения семантики |
| internal/session/actor.go | internal/a11y Reconciler.Apply | A11ySink seam | ✓ WIRED | actor.go:930–947; main.go:201 |
| cmd/goswitchd/main.go | internal/a11y production-адаптеры | a11y.New(NewExecRunner, NewDBusStatusSetter) | ✓ WIRED | main.go:201 (вне дифа, цело) |
| internal/install restoreToolkitAccessibility | install-state.json | toolkit_accessibility | ✓ WIRED | install вне дифа 08-09 — контракты 08-04 целы |
| docs/CONFIG.md таблица | cmd/skillgen → SKILL.md | parseKeyTable + regen | ✓ WIRED | 20 строк; -check exit 0; regen идемпотентен |
| docs/SPEC.md ревизия 2026-10-06 | README.md/README.en.md a11y-раздел | doc-сверка D-8-8 | ✓ WIRED (RESOLVED 2026-10-06) | Оба README переведены на ревизию планом 08-10 — контент сверен с CONFIG.md/SPEC чтением дерева (см. Gap Resolution) |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| pushA11y fold | active bool | snap.A11y.EffectiveEnabled() (загруженный конфиг) | Yes | ✓ FLOWING |
| Reconciler series | desired bool | sink.Apply | Yes | ✓ FLOWING |
| Defaults() | A11y ON | boolPtr(true) в схеме | Yes | ✓ FLOWING |
| skillgen region | table rows | docs/CONFIG.md (единственный источник) | Yes | ✓ FLOWING |
| restoreToolkitAccessibility | value | install-state.json | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Полная зелёная итерация (финал) | `mise run ci` (build+vet+lint+test -race) | exit 0, прогон верификатора 2026-10-06T05:53Z; все пакеты ok | ✓ PASS |
| Skillgen байт-синхрон (финал) | `go run ./cmd/skillgen -check` | CHECK-EXIT-0, прогон верификатора | ✓ PASS |
| Ноль ключа-списка/стейл-прозы в README | `grep -n -iE "перечисляет приложения\|добавления его в список\|adding it to the list\|Removing an application\|a11y\.apps" README.md README.en.md` | единственный матч README.md:412 — troubleshooting «исчез из списка источников ввода» (не a11y); a11y-стейла ноль | ✓ PASS |
| Позитивные якоря bool-контракта в README | grep a11y.enabled/toolkit-accessibility/IsEnabled/enabled: false/goswitchctl uninstall/следующ(ем\|его) запуск/next start/«единственный ключ» | все якоря присутствуют в обоих языках (×3/×4/×1/×1/×3/…) | ✓ PASS |
| Паритет заголовков | `grep -cE '^##+ '` + diff уровней | 22/22 в обоих файлах; структура уровней идентична | ✓ PASS |
| Скоуп-гейт хунков | `git diff -U0 9660977..HEAD` + awk | RU-SCOPE-OK (306..311, 314..319 ⊂ 301–320); EN-SCOPE-OK (301..313 ⊂ 296–314) | ✓ PASS |
| Отсутствие дрейфа с 43/44 | `git diff --name-only cbe4160..HEAD` | только README.md, README.en.md + planning-метаданные; ноль Go/CI/CONFIG/SKILL файлов | ✓ PASS |
| Существование пинов (enumeration) | `go test -list` | 7 config-пинов + 5 session-кейсов присутствуют (TestLoad_A11yAbsentSectionMeansDefaultOn, …RemovedListKeyRejectedWhole, …, TestActor_A11yDefaultsDocumentPushesTrue, …) | ✓ PASS |

### Probe Execution

Не применимо — requirement-ID для probe-падения нет (spec-less fallback; REQUIREMENTS.md не маппит фазу 8 — подтверждено grep'ом), scripts/*/tests/probe-*.sh в репо отсутствуют.

### Requirements Coverage

Трассировка по 08-CONTEXT.md (backlog-derived D-8-1..D-8-11 + Owner Revision 2026-10-06):

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| D-8-1 → D-8-1/REV | 08-01/02/05/09 | Секция в конфиге (пересмотрена: единственный bool-ключ default ON) | ✓ SATISFIED | Схема + фолд + ревизия SPEC; grep-гейты |
| D-8-2 | 08-01 | Механика исследована при планировании | ✓ SATISFIED | 08-RESEARCH; резолюция в ревизиях §4 |
| D-8-3 | 08-03/05 | Демон-reconcile при старте и hot reload | ✓ SATISFIED (unit + UAT-accepted) | Фолд/wiring/тесты; живой эффект принят владельцем по доказательствам |
| D-8-4 | 08-03/04 | Откат только uninstall; снятие запрещено | ✓ SATISFIED (unit + UAT-accepted) | Apply(false) ноль-касаний; install вне дифа; живой revert принят владельцем |
| D-8-5 → D-8-5/REV | 08-02 → 08-09 | Regex-подстрока RE2 — АННУЛИРОВАНА владельцем | ✓ SUPERSEDED-AND-HONORED | Машинерия удалена из схемы; ревизия 2026-10-06 фиксирует аннулирование |
| D-8-6 → D-8-6/REV | 08-01/03/09 | Минимальный набор магии + default ON (осознанное решение) | ✓ SATISFIED | toolkit-accessibility + IsEnabled-пояс нетронуты; pointer-bool default ON запинен |
| D-8-7 | 08-07 | README.ru как главная, README как перевод | ✓ SATISFIED | Структура README.md(RU)/README.en.md(EN) цела; паритет 22/22 подтверждён после правки 08-10 |
| D-8-8 | 08-06/07/10 | Сверка доков с поведением; расхождения в ДОКАХ правятся | ✓ SATISFIED (RESOLVED 2026-10-06) | SPEC/CONFIG/SKILL честны (WR-02 закрыт); README-пара переведена на ревизию планом 08-10 и сверена с CONFIG.md чтением дерева — дельта-недочёт устранён |
| D-8-9 | 08-06 | CONFIG.md русский-первый | ✓ SATISFIED | CONFIG.md по-русски |
| D-8-10 | 08-08 | Single source: генерация из CONFIG.md | ✓ SATISFIED | skillgen + golden 20 + -check зелёный |
| D-8-11 | 08-08 | skills/goswitch-config/ (не .zcode/) | ✓ SATISFIED | git-tracked |

Orphaned requirements: нет (REQUIREMENTS.md не маппит фазу 8). Отложенных к поздним фазам нет — фаза 8 последняя в milestone.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| README.md | 301–335 | (прежний 🛑 Blocker: аннулированный контракт) — РЕШЁН планом 08-10 (d1e44a4): раздел на bool-контракте, ноль a11y.apps | resolved | Пользователь по README получает валидный конфиг |
| README.en.md | 296–331 | (прежний 🛑 Blocker, EN) — РЕШЁН планом 08-10 (8df7162) | resolved | То же |
| — | — | TBD/FIXME/XXX/HACK/PLACEHOLDER в дельта-файлах (config.go, actor.go, a11y.go, main_test.go, CONFIG.md, SKILL.md, README×2) | none | Скан чист |
| — | — | Стабы/empty-implementation в новом коде | none | Нет |
| .planning/todos/pending/ | — | Три свёрнутых в фазу todo не перемещены из pending (housekeeping) | ℹ️ Info | Не артефакт фазы; см. advisory |

### Human Verification Required

Нет — все прежние human-пункты разрешены записанными вердиктами владельца (08-UAT.md, 2026-10-06; раздел «UAT Dispositions»), подтверждены на месте при этом проходе. Остаточный риск, осознанно принятый владельцем (gsettings из контекста демона не проверен живьём), зафиксирован в вердикте и не reopened.

### Final Verdict

44/44 in-force истин верифицированы; 7 superseded-and-honored исключены из знаменателя с верифицированными заменами; 0 behavior-unverified; 0 гэпов; 0 блокеров. Единственный гэп предыдущего прохода (README-пара на аннулированном контракте) закрыт гэп-планом 08-10, скоуп правок подтверждён конфайндом хунков, дрейфа с 43/44 нет, все гейты (`mise run ci`, `skillgen -check`, grep-гейты, паритет 22/22) зелёные в прогоне верификатора. Цель фазы достигнута во всех трёх потоках: a11y-секция (bool-контракт ревизии владельца), документация (сверка честна во всех пользовательских доках, русский-первый), конфиг-skill (генерация из CONFIG.md байт-синхронна). Фаза 8 — последняя в milestone — готова к ship.

---

_Verified: 2026-10-06T05:53:08Z_
_Verifier: Claude (gsd-verifier)_
_History: 44/46 human_needed (2026-10-05T23:55:00Z, commit d61f26e) → 43/44 gaps_found (2026-10-06T01:55:43Z, commit cbe4160) → 44/44 passed (этот отчёт). Предыдущие отчёты preserved in git history._
