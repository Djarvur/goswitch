---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
verified: 2026-10-07T21:32:42Z
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
covered_digest: "v1:sha256:31dac8ee9594c0f047ac57ee9b7880efb7a57e5120f90635e96366ef86cab0d5"
behavior_unverified: 0
overrides_applied: 0
coincidental_reliance_items: []
re_verification:
  previous_status: passed
  previous_score: 44/44
  gaps_closed: []
  gaps_remaining: []
  regressions: []
deferred: []
advisory: []
---

# Phase 8: a11y-магия приложений, документация и конфиг-skill — Verification Report (RE-VERIFICATION: fingerprint refresh at milestone close — passed)

**Phase Goal:** (1) a11y-магия — конфиг-секция с ЕДИНСТВЕННЫМ ключом `a11y.enabled` (bool, default ON, решение владельца D-8-1/REV 2026-10-06); демон применяет сам (reconcile старт/reload; uninstall-revert; список приложений АННУЛИРОВАН). (2) Документация — сверка с фактическим поведением, русский-первый. (3) Skill для AI — SKILL.md генерируется из docs/CONFIG.md, синхрон через golden-тест.
**Verified:** 2026-10-07T21:32:42Z
**Status:** passed
**Re-verification:** Yes — четвёртый проход: fingerprint-staleness refresh на закрытии milestone (все 8 фаз, v1.1.0 выпущена). История: 44/46 human_needed (2026-10-05T23:55Z) → 43/44 gaps_found (2026-10-06T01:55:43Z) → 44/44 passed (2026-10-06T05:53:08Z) → **44/44 passed, дайджест обновлён** (этот отчёт).

## Re-Verification Scope (2026-10-07)

Предыдущий отчёт не имел открытых гэпов (`gaps_remaining: []`), поэтому этот проход — регрессионная проверка passed-истин против ТЕКУЩЕГО дерева после трёх quick-workstream'ов, приземлившихся с момента верификации 2026-10-06T05:53Z (BASE-коммит 3ea1731):

- **Дрейф охарактеризован:** `git diff --name-only 3ea1731..HEAD` = 84 файла — 261006-vqw (посимвольная инверсия смешанного текста), 261006-squ (постоянный PulseAudio-стрим + звуковая тема), 261007-0yg (internalize engine/+layouts/ под internal/, SPEC §8 D-55), транскрипция UAT (2171fae) и planning-метаданные. Ни один a11y-контрактный файл не изменился ФУНКЦИОНАЛЬНО.
- **Затронутые covered-файлы фазы 8 (11 из 43) — изменения проверены попарно:**
  - `internal/session/actor.go` / `actor_test.go` — рефакторинги звука/смешанного текста; a11y-фолд СОХРАНЁН: `a.pushA11y(snap)` (actor.go:1327, «the a11y-magic fold rides the same per-fold push chain»), `pushA11y` читает только `snap.A11y.EffectiveEnabled()` (:1361); корпус TestActor_A11y* (5 кейсов) — **ok** в прогоне верификатора.
  - `cmd/goswitchd/main.go` — импорты internalize + sound-sink wiring; a11y-wiring цел: main.go:210 `actor.SetA11ySink(a11y.New(a11y.NewExecRunner(), a11y.NewDBusStatusSetter()))`.
  - `internal/install/install.go` / `install_test.go` — ТОЛЬКО переименование импорта `engine` → `internal/engine` (2 строки); контракты 08-04 нетронуты.
  - `mise.toml` — путь `dictgen-regen` (`./layouts` → `./internal/layouts`); задача `skillgen-regen` нетронута.
  - `docs/CONFIG.md` — изменена ТОЛЬКО строка `sound.autocorrect_event` (261006-squ); a11y-строка `| a11y.enabled | bool | true | …` — байт-целa.
  - `skills/goswitch-config/SKILL.md` — перегенерированная область зеркалит изменение звуковой строки; a11y-раздел нетронут; **golden-синхрон подтверждён** (см. гейты).
  - `README.md` — хунки вне a11y-разделов (проза смешанного текста :155–160 и звука :262–275); a11y-раздел (:301–340) нетронут, на bool-контракте.
  - `docs/SPEC.md` — добавлены позднейшие D-55-ревизии (звук, mixed-text, internalize §8); ревизия a11y 2026-10-06 (§4, :250–260+) цела, прежние блоки на месте.
  - `08-UAT.md` — транскрибированы вердикты владельца (коммит 2171fae): `status: complete`, 6/6 pass.
- **Пути:** предупреждение о stale-путях не подтвердилось — прежний covered_files НИКОГДА не содержал путей engine//layouts/; все 43 пути проверены на существование в текущем дереве (43/43 OK). engine/ и layouts/ отсутствуют в корне — только internal/engine/, internal/layouts/.
- **Go-схема в текущем дереве:** config.go:184–186 `A11y{Enabled *bool}` — единственное поле; `EffectiveEnabled()` (:195–197, nil = ON) — единственный метод секции; maxA11yApps/errA11yApps = 0; единственное упоминание `a11y.apps` в config.go:169 — doc-коммент об УДАЛЕНИИ машинерии (корректно). Контракт ревизии владельца — действующий.

## Goal Achievement

Дерево после трёх quick-workstream'ов не тронуло ни одну истин фазы 8: a11y-схема и internal/a11y — без изменений с 2026-10-06; фолд и wiring пережили рефакторинги (тесты зелёные в прогоне верификатора); доки/skill синхронны после регенерации звуковой строки (golden-гейт зелёный); README-пара честна в обоих языках; UAT-вердикты владельца транскрибированы в дерево.

**Score:** 44/44 in-force truths verified (7 former truths superseded-and-honored by the 2026-10-06 revision — excluded from the denominator; 0 FAILED; 0 behavior-unverified — все human-пункты разрешены записанными вердиктами владельца, транскрибированными в 08-UAT.md).

### Observable Truths

Статусы перенесены из предыдущего прохода (passed); для 11 затронутых дрейфом файлов доказательства обновлены в текущем дереве. Полные таблицы по планам 08-01..08-10 — в git-истории предыдущего отчёта; ниже — сводная матрица с точечными обновлениями.

#### Plan 08-01 — SPEC §4 spec-delta (6 truths)

| # | Truth | Status | Evidence (текущее дерево) |
|---|-------|--------|--------------------------|
| 1 | SPEC §4 несёт датированную ревизию 2026-10-05 ДО кода; аудит-трейл цел | ✓ VERIFIED | Блок 2026-10-05 нетронут; ревизия 2026-10-06 рядом (SPEC.md:250+); позднейшие D-55-блоки — add-only другими воркстримами |
| 2 | Ревизия фиксирует секцию a11y (enabled/apps, потолок 64, zero-off) | ⚠️ SUPERSEDED-AND-HONORED | Аннулировано D-8-1/REV; замена верифицирована (08-09 #1) |
| 3 | Ревизия фиксирует daemon-reconcile, отказ от снятия, uninstall-revert | ✓ VERIFIED | Ревизия 2026-10-06 цела в дереве; internal/install — import-only дифф |
| 4 | Ревизия фиксирует restart-семантику | ✓ VERIFIED | SPEC.md ревизия 2026-10-06 цела; README.en.md:319 «at their next start» |
| 5 | Пер-апп оверрайды — документируемый рецепт + argv-дисциплина | ✓ VERIFIED | Рецепт сохранён (CONFIG.md); argv-дисциплина в ревизии («по построению»); reconciler без изменений с 2026-10-06 |
| 6 | Spec-less fallback пропущен видимо | ✓ VERIFIED | REQUIREMENTS.md не маппит фазу 8; трассировка через 08-CONTEXT D-8-* |

#### Plan 08-02 — Схема секции a11y (5 truths)

| # | Truth | Status | Evidence (текущее дерево) |
|---|-------|--------|--------------------------|
| 1 | Секция A11y (yaml a11y: enabled/apps), zero-value off | ⚠️ SUPERSEDED-AND-HONORED | Замена в дереве: config.go:184–186 `A11y{Enabled *bool}`, nil = ON |
| 2 | Валидация: потолок 64, blank-отказ, сентинелы | ⚠️ SUPERSEDED-AND-HONORED | Машинерия удалена (maxA11yApps/errA11yApps = 0); config.go:169 — коммент об удалении |
| 3 | Active() + Defaults() off | ⚠️ SUPERSEDED-AND-HONORED | Замена: EffectiveEnabled() — единственный метод (config.go:195–197); Defaults ON |
| 4 | Strict decode D-33 и last-good; load.go/watch.go — ноль правок | ✓ VERIFIED | internal/config вне дрейфа 3ea1731..HEAD; корпус TestLoad_A11y*/TestWatch_A11y* — **ok** (прогон верификатора 2026-10-07) |
| 5 | Hot reload пин | ⚠️ SUPERSEDED-AND-HONORED | Замена: TestWatch_A11yUnknownKeyKeepsLastGood — в корпусе **ok** |

#### Plan 08-03 — Reconciler internal/a11y (6 truths)

Все 6 ✓ VERIFIED — **internal/a11y ОТСУТСТВУЕТ в дрейфе 3ea1731..HEAD** (ни one файла не менялся с предыдущей верификации): Apply fire-and-forget, read-verify-then-set, пояс IsEnabled, argv-дисциплина, degradation, -race. Wiring демона подтверждён заново (main.go:210).

#### Plan 08-04 — Installer snapshot/revert (5 truths)

Все 5 ✓ VERIFIED — дифф install.go/install_test.go с 3ea1731 — ТОЛЬКО импорт `engine` → `internal/engine` (проверено попарным diff'ом); контракты snapshot/revert/only-if-present/argv/chain нетронуты; `go build ./...` exit 0 (прогон верификатора).

#### Plan 08-05 — Фолд актора + wiring (4 truths)

| # | Truth | Status | Evidence (текущее дерево) |
|---|-------|--------|--------------------------|
| 1 | Фолд на EffectiveEnabled с дифф-гейтом | ⚠️ SUPERSEDED-AND-HONORED→✓ | Замена верифицирована в ТЕКУЩЕМ дереве: actor.go:1327/1361; корпус TestActor_A11y{DefaultsDocumentPushesTrue, FoldTriggersSinkOnActivation, ExplicitFalsePushesFalse} — **ok** после рефакторингов звука/смешанного текста |
| 2 | A11ySink seam + self-sync; nil no-op | ✓ VERIFIED | Seam пережил дифф (TestActor_A11ySinkInstallSelfSyncs / NilSinkNoOp — **ok**) |
| 3 | Wiring демона на production-адаптерах | ✓ VERIFIED | main.go:210 `SetA11ySink(a11y.New(NewExecRunner, NewDBusStatusSetter))` |
| 4 | Двойная идемпотентность | ✓ VERIFIED | Дифф-гейты в коде; корпус **ok** |

#### Plan 08-06 — CONFIG.md русский-первый + SPEC-сверка (7 truths)

| # | Truth | Status | Evidence (текущее дерево) |
|---|-------|--------|--------------------------|
| 1 | CONFIG.md русский-первый | ✓ VERIFIED | Язык сохранён; дифф дрейфа — одна строка звуковой таблицы |
| 2 | Построчная сверка таблицы со схемой | ✓ VERIFIED | a11y-строка `a11y.enabled/bool/true` байт-цела; golden 20 строк синхронен (гейт ниже) |
| 3 | SPEC-сверка 08-06, аудит-трейл цел | ✓ VERIFIED | Блоки 08-06/2026-10-05/2026-10-06 целы; позднейшие D-55 — add-only |
| 4 | «Ровно семь секций» + ровно 2 a11y-строки | ⚠️ SUPERSEDED-AND-HONORED | Семь секций подтверждены (config.go doc-коммент «exactly the seven sections»); a11y-строка ровно 1 |
| 5 | Поведение магии задокументировано честно | ✓ VERIFIED | CONFIG.md a11y-строка и раздел :227–262 нетронуты дрейфом |
| 6 | Per-app .desktop-оверрайды — ручной рецепт | ✓ VERIFIED | Рецепт сохранён |
| 7 | Формат таблицы байт-строгий | ✓ VERIFIED | skillgen -check exit 0; TestSkillgen_CommittedFileInSync ok — ПОСЛЕ регенерации звуковой строки |

#### Plan 08-07 — README русский-первый (5 truths)

| # | Truth | Status | Evidence (текущее дерево) |
|---|-------|--------|--------------------------|
| 1 | README.md — русская главная + a11y-раздел | ✓ VERIFIED | A11y-раздел :301–340 вне хунков дрейфа (:155–160, :262–275); grep a11y.enabled — :313,:324,:340 |
| 2 | README.en.md — EN канон, паритет | ✓ VERIFIED | Паритет заголовков 22/22 — ПОСЧИТАН верификатором в текущем дереве |
| 3 | README.ru.md удалён, дублей нет | ✓ VERIFIED | В корне только README.md + README.en.md; git ls-files без README.ru.md |
| 4 | a11y-раздел на bool-контракте | ✓ VERIFIED | grep `a11y\.apps` = 0 в обоих; якоря: RU :313/:324/:340, EN :302/:309/:314/:316/:319/:324 (toolkit-accessibility ×3, enabled: false, next start, uninstall) |
| 5 | Входящие ссылки перепoint'нуты | ✓ VERIFIED | docs/ACCEPTANCE.md вне дрейфа — цел |

#### Plan 08-08 — Skill goswitch-config (6 truths)

| # | Truth | Status | Evidence (текущее дерево) |
|---|-------|--------|--------------------------|
| 1 | SKILL.md frontmatter + RU description + каркас | ✓ VERIFIED | Файл цел; a11y-каркас вне диффа регенерации |
| 2 | Single source: генератор парсит таблицу | ✓ VERIFIED | cmd/skillgen вне дрейфа |
| 3 | Синхрон-гейт: -check + golden-тест | ✓ VERIFIED | `go run ./cmd/skillgen -check` → **exit 0**; TestSkillgen_CommittedFileInSync → **ok** (прогон верификатора 2026-10-07, после регенерации) |
| 4 | Dev-only mise-задача skillgen-regen | ✓ VERIFIED | mise.toml дифф — только dictgen-regen путь; skillgen-regen нетронута |
| 5 | TRACKING-гейт: SKILL.md в git-индексе | ✓ VERIFIED | git-tracked (дифф регенерации закоммичен воркстримом squ) |
| 6 | Регенерация идемпотентна | ✓ VERIFIED | -check exit 0 означает: закоммиченный файл байт-идентичен выводу генератора |

#### Plan 08-09 — Дельта-ревизия владельца (7 truths)

| # | Truth | Status | Evidence (текущее дерево) |
|---|-------|--------|--------------------------|
| 1 | SPEC §4 ревизия 2026-10-06, дословная цитата, аудит-цел | ✓ VERIFIED | SPEC.md:250+ прочитано в дереве; цитата «…список не нужен, а нужен bool параметр, по дефолту настройка включен» на месте |
| 2 | Схема без Apps-машинерии; единственный EffectiveEnabled; Validate без a11y | ✓ VERIFIED | config.go:184–186/195–197; maxA11yApps/errA11yApps = 0; :169 — коммент удаления |
| 3 | Default ON запинен; Load не накладывает Defaults | ✓ VERIFIED | Тесты на месте; TestLoad_A11y{AbsentSectionMeansDefaultOn, AbsentKeyMeansDefaultOn, ExplicitFalseDisables} + TestDefaults_A11yDefaultOn — **ok** |
| 4 | Громкая миграция D-33 | ✓ VERIFIED | TestLoad_A11yRemovedListKeyRejectedWhole — в корпусе **ok** |
| 5 | Фолд на EffectiveEnabled; контракты 08-03/08-04 не меняются | ✓ VERIFIED | actor.go:1327/1361; internal/a11y и install вне функционального дрейфа |
| 6 | Доки/skill согласованы | ✓ VERIFIED | grep a11y.apps = 0 в CONFIG.md/SKILL.md/README×2; одна a11y-строка bool/true; -check exit 0 |
| 7 | Зелёная итерация | ✓ VERIFIED | Точечные прогоны верификатора: skillgen ok, config A11y ok, session A11y ok, `go build ./...` exit 0 |

#### Plan 08-10 — Гэп-закрытие README (3 truths)

Все 3 ✓ VERIFIED — README.md/README.en.md a11y-разделы вне хунков дрейфа; grep-гейты (негатив a11y.apps = 0, позитивные якоря, паритет 22/22) повторены верификатором в текущем дереве — PASS.

### Superseded-and-Honored Truths (плановая цепочка)

Семь истин планов 08-01..08-06 (a11y.apps-машинерия) остаются замещёнными ревизией D-8-1/REV через план 08-09 — цепочка и верификации замен не изменились с предыдущего прохода (см. git-историю предыдущего отчёта); все семь замен переподтверждены в текущем дереве (таблицы 08-02/08-05/08-06 выше).

### UAT Dispositions (owner verdicts 2026-10-06 — транскрибированы в дерево коммитом 2171fae)

08-UAT.md в текущем дереве: `status: complete`, 6/6 pass (живой эффект a11y-магии, живой uninstall-revert, русская проза/пара README, полезность SKILL.md, WR-02 RESOLVED BY OWNER, человеческий проход CR-01 — все ACCEPTED ON EVIDENCE с записанными обоснованиями). Все 6 human-пунктов остаются разрешёнными; новых human-пунктов этот проход не породил.

### Key Link Verification

| From | To | Via | Status | Details (текущее дерево) |
|------|----|----|--------|-------------------------|
| docs/SPEC.md §4 (ревизия 2026-10-06) | internal/config (A11y) | контракт секции | ✓ WIRED | Схема = ревизии: config.go:184–197 |
| internal/config A11y.EffectiveEnabled | internal/session/actor.go (pushA11y) | snap.A11y.EffectiveEnabled() | ✓ WIRED | actor.go:1361 (вызов фолда :1327); пережил рефакторинги |
| internal/session/actor.go | internal/a11y Reconciler.Apply | A11ySink seam | ✓ WIRED | seam цел; main.go:210 |
| cmd/goswitchd/main.go | internal/a11y production-адаптеры | a11y.New(NewExecRunner, NewDBusStatusSetter) | ✓ WIRED | main.go:210 |
| internal/install restoreToolkitAccessibility | install-state.json | toolkit_accessibility | ✓ WIRED | Import-only дифф; контракты 08-04 целы |
| docs/CONFIG.md таблица | cmd/skillgen → SKILL.md | parseKeyTable + regen | ✓ WIRED | 20 строк; -check exit 0 после регенерации звуковой строки |
| docs/SPEC.md ревизия 2026-10-06 | README.md/README.en.md a11y-разделы | doc-сверка D-8-8 | ✓ WIRED | Оба README вне хунков дрейфа в a11y-разделах; grep-гейты PASS |

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
| Skillgen байт-синхрон | `go run ./cmd/skillgen -check` | CHECK-EXIT-0 (2026-10-07, после регенерации) | ✓ PASS |
| Golden-синхрон (single named) | `go test -run TestSkillgen_CommittedFileInSync ./cmd/skillgen` | ok 0.002s | ✓ PASS |
| Config a11y-пины (single named) | `go test -run A11y ./internal/config` | ok 0.038s (7 пинов) | ✓ PASS |
| Фолд актора после рефакторингов (single named) | `go test -run A11y ./internal/session` | ok 0.004s (5 кейсов) | ✓ PASS |
| Сборка дерева после internalize | `go build ./...` | exit 0 | ✓ PASS |
| Ноль ключа-списка | `grep -c "a11y\.apps" README.md README.en.md docs/CONFIG.md SKILL.md` | 0/0/0/0 (config.go:169 — коммент удаления) | ✓ PASS |
| Позитивные якоря bool-контракта | grep a11y.enabled/toolkit-accessibility/enabled: false/next start/uninstall в README×2 | все якоря на месте в обоих языках | ✓ PASS |
| Паритет заголовков | `grep -cE '^##+ '` | 22/22 в обоих файлах | ✓ PASS |
| Существование пинов (enumeration) | `go test -list` | 7 config + 5 session + 1 skillgen a11y/golden пинов присутствуют | ✓ PASS |
| Существование covered-файлов | existence loop | 43/43 OK | ✓ PASS |
| Debt-маркеры | grep TBD/FIXME/XXX/HACK/PLACEHOLDER по 12 файлам фазы | чисто | ✓ PASS |

### Probe Execution

Не применимо — REQUIREMENTS.md не маппит фазу 8 (spec-less fallback, подтверждено предыдущими проходами); scripts/*/tests/probe-*.sh в репо отсутствуют.

### Requirements Coverage

Трассировка по 08-CONTEXT.md (backlog-derived D-8-1..D-8-11 + Owner Revision 2026-10-06) — все статусы предыдущего прохода подтверждены в текущем дереве:

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| D-8-1 → D-8-1/REV | 08-01/02/05/09 | Секция (пересмотрена: единственный bool-ключ default ON) | ✓ SATISFIED | config.go:184–197 + фолд actor.go:1327/1361 + SPEC-ревизия |
| D-8-2 | 08-01 | Механика исследована при планировании | ✓ SATISFIED | 08-RESEARCH; резолюции в ревизиях §4 |
| D-8-3 | 08-03/05 | Демон-reconcile при старте и hot reload | ✓ SATISFIED (unit + UAT-accepted) | Фолд/wiring/тесты ok в текущем дереве |
| D-8-4 | 08-03/04 | Откат только uninstall; снятие запрещено | ✓ SATISFIED (unit + UAT-accepted) | Apply(false) ноль-касаний; install import-only дифф |
| D-8-5 → D-8-5/REV | 08-02 → 08-09 | Regex-подстрока RE2 — АННУЛИРОВАНА | ✓ SUPERSEDED-AND-HONORED | Машинерия удалена; config.go:169 фиксирует удаление |
| D-8-6 → D-8-6/REV | 08-01/03/09 | Минимальный набор магии + default ON | ✓ SATISFIED | toolkit-accessibility + IsEnabled-пояс нетронуты; nil=ON |
| D-8-7 | 08-07 | README русский-первый | ✓ SATISFIED | README.md(RU)/README.en.md(EN); паритет 22/22 в текущем дереве |
| D-8-8 | 08-06/07/10 | Сверка доков с поведением | ✓ SATISFIED | SPEC/CONFIG/SKILL/README честны; golden-гейт зелёный после регенерации |
| D-8-9 | 08-06 | CONFIG.md русский-первый | ✓ SATISFIED | CONFIG.md по-русски |
| D-8-10 | 08-08 | Single source: генерация из CONFIG.md | ✓ SATISFIED | skillgen + golden 20 + -check exit 0 |
| D-8-11 | 08-08 | skills/goswitch-config/ | ✓ SATISFIED | git-tracked |

Orphaned requirements: нет. Отложенных к поздним фазам нет — фаза 8 последняя в milestone.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | TBD/FIXME/XXX/HACK/PLACEHOLDER в файлах фазы (12 файлов, включая 11 затронутых дрейфом) | none | Скан чист |
| — | — | Стабы/empty-implementation в a11y-коде | none | Нет |
| .planning/todos/completed/ | — | (прежний ℹ️ Info/advisory: три свёрнутых todo в pending) — РЕШЁН: все четыре файла в completed/, pending пуст | resolved | Process hygiene закрыта |

### Human Verification Required

Нет — все 6 прежних human-пунктов разрешены записанными вердиктами владельца и транскрибированы в дерево (08-UAT.md, `status: complete`, коммит 2171fae). Изменения дерева с прошлого прохода (quick-workstreams) покрыты автоматическими гейтами этого прохода; новых human-пунктов не возникло.

### Final Verdict

Fingerprint-staleness re-verification на закрытии milestone: дрейф 3ea1731..HEAD (84 файла — mixed-text inversion, звуковой стрим, internalize engine/layouts, UAT-транскрипция) не задел ни одну истин фазы 8 функционально. A11y-схема (`a11y.enabled` bool default ON) и reconciler — без изменений; фолд и wiring пережили рефакторинги (тесты ok); installer — import-only; CONFIG.md/SKILL.md синхронны после регенерации (golden + -check зелёные в прогоне верификатора); README-пара нетронута в a11y-разделах (grep-гейты, паритет 22/22); SPEC-ревизия 2026-10-06 цела; UAT 6/6 транскрибирован; дерево собирается (`go build ./...` exit 0). Advisory предыдущего прохода (todo-hygiene) закрыт. 44/44 in-force истин verified, 0 гэпов, 0 behavior-unverified, 0 регрессий. Фаза 8 готова к ship вместе с milestone.

---

_Verified: 2026-10-07T21:32:42Z_
_Verifier: Claude (gsd-verifier)_
_History: 44/46 human_needed (2026-10-05T23:55:00Z, d61f26e) → 43/44 gaps_found (2026-10-06T01:55:43Z, cbe4160) → 44/44 passed (2026-10-06T05:53:08Z, 3ea1731) → 44/44 passed — fingerprint refresh (этот отчёт, e76f6b0). Предыдущие отчёты preserved in git history._
