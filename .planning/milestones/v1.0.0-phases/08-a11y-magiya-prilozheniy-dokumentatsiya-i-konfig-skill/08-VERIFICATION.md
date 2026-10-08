---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
verified: 2026-10-07T23:22:23Z
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
covered_digest: "v1:sha256:b3b6f5fd74c71afada2b30e038e45553735b30798e726411f6e0b3ace35f80ea"
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

# Phase 8: a11y-магия приложений, документация и конфиг-skill — Verification Report (RE-VERIFICATION: staleness refresh #2 at milestone close — passed)

**Phase Goal:** (1) a11y-магия — конфиг-секция с ЕДИНСТВЕННЫМ ключом `a11y.enabled` (bool, default ON, решение владельца D-8-1/REV 2026-10-06); демон применяет сам (reconcile старт/reload; uninstall-revert; список приложений АННУЛИРОВАН). (2) Документация — сверка с фактическим поведением, русский-первый. (3) Skill для AI — SKILL.md генерируется из docs/CONFIG.md, синхрон через golden-тест.
**Verified:** 2026-10-07T23:22:23Z
**Status:** passed
**Re-verification:** Yes — пятый проход: второй fingerprint-staleness refresh цепочки закрытия milestone. Предыдущий отчёт (verified 2026-10-07T21:32:42Z, закоммичен ae54bf4) инвалидован дайджестом: после него приземлился quick task 261008-00m, изменивший tree-код, возможно покрытый этой фазой. История: 44/46 human_needed (2026-10-05T23:55Z) → 43/44 gaps_found (2026-10-06T01:55:43Z) → 44/44 passed (2026-10-06T05:53:08Z) → 44/44 passed, refresh #1 (2026-10-07T21:32:42Z) → **44/44 passed, refresh #2** (этот отчёт).
**Branch/HEAD observed:** `gsd/phase-07-menyu-v2-i-chernyy-spisok-avtokorrektsii` @ 39a5d82.

## Re-Verification Scope (2026-10-07, refresh #2)

Предыдущий проход не имел открытых гэпов (`gaps_remaining: []`), поэтому этот проход — регрессионная проверка passed-истин против ТЕКУЩЕГО дерева после quick task 261008-00m (26 коммитов e76f6b0..39a5d82). Затронутые tree-файлы, попадающие в покрытие фазы 8:

- **`internal/session/actor.go` — flipCredit (G-2-3, ADR-004 amendment 2026-10-07, 4a0c887):** новое поле `flipCredit int` (:253–260), армирование кредита при УСПЕШНОМ SetGlobalEngine round trip в `flipTo` (:2008–2015), потребление ровно одного кредита в ветке FocusOut `HandleLifecycle` (:773–782) с пропуском ТОЛЬКО buffer hard-reset. **A11y-фолд СОХРАНЁН:** изменение конфайнится к lifecycle/flip-путям — фолд же управляется снапшотами (start/reload), не lifecycle-событиями; `a.pushA11y(snap)` (:1361), `pushA11y` читает только `snap.A11y.EffectiveEnabled()` (:1396), дифф-гейт `a11yActive` цел. Корпус TestActor_A11y* (5 кейсов) — **ok с -race** в прогоне верификатора (1.013s).
- **`internal/install/*` — G-5-5 auto-wrap (2fa8b53):** `wrapSources` теперь ДОЗАВЕРШАЕТ half-state (one goswitch tuple + один wrappable xkb = канонический owned-список, дедуп повторов) вместо отказа; selfcheck лечит completable half-state (wrap → один gsettings set → read-back verify). Это НОВОЕ owner-декретированное поведение (spec-delta ADR-006 amendment), верифицированное под собственным отчётом quick'а (261008-00m-VERIFICATION.md). **Контракты фазы 8 не задеты:** 08-04 пинит ТОЛЬКО toolkit_accessibility snapshot/restore — grep 08-04-PLAN.md по wrapSources/errMixedSources/half = 0 совпадений; `restoreToolkitAccessibility` на месте (:460 вызов, :1096–1110 определение), state-ключ `toolkit_accessibility` (:242). Корпус фазы TestInstall_A11y* (3) + TestUninstall_A11y* (5) = 8/8 — **ok с -race**. Отказ-гейт прежней дисциплины (refusal ДО любой мутации) сохранён в новом heal-коде selfcheck (:212–216: wrap-refusal до gsettings set).
- **`internal/install/install_internal_test.go` / `selfcheck.go` / `selfcheck_test.go`** — созданы фазой 4 / изменены quick'ом; в covered_files фазы 8 НЕ входят (и не входили), их контракты — зона quick-отчёта; поверхностная проверка debt-маркеров — чисто.
- **docs: ADR-004/ADR-006 amendments + SPEC §4.3/§4.4 deltas (0781a5f):** docs-only. SPEC-хунок confайнится к buffer-reset прозе (~:123): прежний пункт сохранён ДОСЛОВНО в audit-trail HTML-комменте (та дисциплина, что фаза 8 установила), новая строка несёт маркер «rev G-2-3, ADR-004 amendment 2026-10-07». **A11y-ревизия 2026-10-06 цела** (:268–277, вербатим-цитата владельца «…список не нужен, а нужен bool параметр, по дефолту…» на :277). ADR-004/ADR-006 — чистые добавления (+36 строк каждый, ноль удалений).
- **Golden skillgen (CONFIG.md ↔ SKILL.md):** `git diff e76f6b0..HEAD -- docs/CONFIG.md skills/goswitch-config/SKILL.md` — ПУСТО (quick task конфиг не трогал, подтверждено); a11y-строка CONFIG.md:77 байт-цела. TestSkillgen_CommittedFileInSync — **ok с -race**; `skillgen -check` — **exit 0**.
- **Не задеты дрейфом (пустые diff'ы e76f6b0..HEAD):** internal/a11y/, internal/config/, cmd/goswitchd/main.go, cmd/skillgen/, mise.toml, README.md, README.en.md, docs/ACCEPTANCE.md, docs/CONFIG.md, SKILL.md, 08-UAT.md — a11y-ядро фазы нетронуто вовсе.
- **Покрытие:** все 43 covered-файла существуют в текущем дереве (43/43 OK); дайджест пересчитан по текущему дереву через verification.fingerprint (см. frontmatter; прежний 31dac8ee… заменён на b3b6f5fd…).

## Goal Achievement

Второй дрейф цепочки закрытия milestone (flipCredit, G-5-5 auto-wrap, spec-delta-доки) не задел ни одну истин фазы 8 функционально: a11y-схема и internal/a11y — без изменений; фолд и wiring пережили flipCredit-рефакторинг (5/5 -race ok); installer a11y-snapshot/restore пережил G-5-5 (8/8 -race ok; grep по 08-04 PLAN — нулевая связка с wrap-семантикой); доки/skill синхронны (CONFIG/SKILL нетронуты, golden + -check зелёные); SPEC-ревизия a11y 2026-10-06 цела, новые дельты — audit-trail-дисциплиной; README-пара нетронута вовсе.

**Score:** 44/44 in-force truths verified (7 former truths superseded-and-honored by the 2026-10-06 revision — excluded from the denominator; 0 FAILED; 0 behavior-unverified — все human-пункты разрешены записанными вердиктами владельца, транскрибированными в 08-UAT.md; 08-UAT.md вне дрейфа).

### Observable Truths

Статусы перенесены из предыдущего прохода (passed); для 4 затронутых дрейфом covered-файлов (actor.go, actor_test.go, install.go, SPEC.md) доказательства обновлены в текущем дереве. Полные таблицы по планам 08-01..08-10 — в git-истории предыдущих отчётов; ниже — сводная матрица с точечными обновлениями.

#### Plan 08-01 — SPEC §4 spec-delta (6 truths)

| # | Truth | Status | Evidence (текущее дерево) |
|---|-------|--------|--------------------------|
| 1 | SPEC §4 несёт датированную ревизию 2026-10-05 ДО кода; аудит-трейл цел | ✓ VERIFIED | Блоки 2026-10-05/2026-10-06 целы (:268–277); новейший хунок — rev G-2-3 с audit-trail-комментом (прежний текст дословно сохранён) |
| 2 | Ревизия фиксирует секцию a11y (enabled/apps, потолок 64, zero-off) | ⚠️ SUPERSEDED-AND-HONORED | Аннулировано D-8-1/REV; замена верифицирована (08-09 #1) |
| 3 | Ревизия фиксирует daemon-reconcile, отказ от снятия, uninstall-revert | ✓ VERIFIED | Ревизия 2026-10-06 цела; restoreToolkitAccessibility :460/:1096 цел после G-5-5 |
| 4 | Ревизия фиксирует restart-семантику | ✓ VERIFIED | SPEC.md ревизия 2026-10-06 цела; README.en.md «at their next start» (README вне дрейфа) |
| 5 | Пер-апп оверрайды — документируемый рецепт + argv-дисциплина | ✓ VERIFIED | Рецепт сохранён (CONFIG.md:77 a11y-строка байт-цела); argv-дисциплина в ревизии; новый selfcheck-heal соблюдает refusal-before-mutation |
| 6 | Spec-less fallback пропущен видимо | ✓ VERIFIED | REQUIREMENTS.md не маппит фазу 8; трассировка через 08-CONTEXT D-8-* |

#### Plan 08-02 — Схема секции a11y (5 truths)

| # | Truth | Status | Evidence (текущее дерево) |
|---|-------|--------|--------------------------|
| 1 | Секция A11y (yaml a11y: enabled/apps), zero-value off | ⚠️ SUPERSEDED-AND-HONORED | Замена в дереве: config.go:184–186 `A11y{Enabled *bool}`, nil = ON (internal/config вне дрейфа) |
| 2 | Валидация: потолок 64, blank-отказ, сентинелы | ⚠️ SUPERSEDED-AND-HONORED | Машинерия удалена (maxA11yApps/errA11yApps = 0); config.go:169 — коммент об удалении |
| 3 | Active() + Defaults() off | ⚠️ SUPERSEDED-AND-HONORED | Замена: EffectiveEnabled() (:195–197); Defaults ON |
| 4 | Strict decode D-33 и last-good; load.go/watch.go — ноль правок | ✓ VERIFIED | internal/config вне дрейфа e76f6b0..HEAD; корпус A11y — **ok с -race** (прогон верификатора) |
| 5 | Hot reload пин | ⚠️ SUPERSEDED-AND-HONORED | Замена: TestWatch_A11yUnknownKeyKeepsLastGood — в корпусе **ok** |

#### Plan 08-03 — Reconciler internal/a11y (6 truths)

Все 6 ✓ VERIFIED — **internal/a11y ОТСУТСТВУЕТ в дрейфе e76f6b0..HEAD**: Apply fire-and-forget, read-verify-then-set, пояс IsEnabled, argv-дисциплина, degradation, -race. Wiring демона переподтверждён (main.go:210 `SetA11ySink(a11y.New(a11y.NewExecRunner(), a11y.NewDBusStatusSetter()))`).

#### Plan 08-04 — Installer snapshot/revert (5 truths)

Все 5 ✓ VERIFIED — G-5-5-дифф install.go не пересекается с a11y-контрактами: grep 08-04-PLAN.md по wrapSources/errMixedSources/half = **0**; toolkit_accessibility (:242), restoreToolkitAccessibility (:460, :1096–1110) на месте; корпус TestInstall_A11y* + TestUninstall_A11y* — 8/8 **ok с -race**; `go build ./...` exit 0.

#### Plan 08-05 — Фолд актора + wiring (4 truths)

| # | Truth | Status | Evidence (текущее дерево) |
|---|-------|--------|--------------------------|
| 1 | Фолд на EffectiveEnabled с дифф-гейтом | ⚠️ SUPERSEDED-AND-HONORED→✓ | Замена верифицирована в ТЕКУЩЕМ дереве ПОСЛЕ flipCredit: actor.go:1361/1396; TestActor_A11y{DefaultsDocumentPushesTrue, FoldTriggersSinkOnActivation, ExplicitFalsePushesFalse} — **ok -race** |
| 2 | A11ySink seam + self-sync; nil no-op | ✓ VERIFIED | Seam пережил flipCredit-дифф (TestActor_A11ySinkInstallSelfSyncs / NilSinkNoOp — **ok -race**) |
| 3 | Wiring демона на production-адаптерах | ✓ VERIFIED | main.go:210 — main.go вне дрейфа e76f6b0..HEAD |
| 4 | Двойная идемпотентность | ✓ VERIFIED | Дифф-гейты в коде; корпус **ok -race** |

#### Plan 08-06 — CONFIG.md русский-первый + SPEC-сверка (7 truths)

| # | Truth | Status | Evidence (текущее дерево) |
|---|-------|--------|--------------------------|
| 1 | CONFIG.md русский-первый | ✓ VERIFIED | CONFIG.md вне дрейфа e76f6b0..HEAD (пустой diff) |
| 2 | Построчная сверка таблицы со схемой | ✓ VERIFIED | a11y-строка :77 байт-цела; golden синхронен (гейты ниже) |
| 3 | SPEC-сверка 08-06, аудит-трейл цел | ✓ VERIFIED | Блоки 08-06/2026-10-05/2026-10-06 целы; новейшая дельта G-2-3 — audit-trail-коммент с дословным сохранением прежнего текста |
| 4 | «Ровно семь секций» + ровно 2 a11y-строки | ⚠️ SUPERSEDED-AND-HONORED | Семь секций подтверждены (config.go doc-коммент); a11y-строка ровно 1 |
| 5 | Поведение магии задокументировано честно | ✓ VERIFIED | CONFIG.md a11y-строка и раздел нетронуты дрейфом |
| 6 | Per-app .desktop-оверрайды — ручной рецепт | ✓ VERIFIED | Рецепт сохранён |
| 7 | Формат таблицы байт-строгий | ✓ VERIFIED | `skillgen -check` exit 0; TestSkillgen_CommittedFileInSync ok -race — на нетронутом golden'е |

#### Plan 08-07 — README русский-первый (5 truths)

Все 5 ✓ VERIFIED — README.md/README.en.md вне дрейфа e76f6b0..HEAD (пустые diff'ы); grep-гейты (негатив a11y.apps = 0/0, позитивные якоря 7/7 в обоих, паритет заголовков 22/22) повторены верификатором в текущем дереве — PASS.

#### Plan 08-08 — Skill goswitch-config (6 truths)

| # | Truth | Status | Evidence (текущее дерево) |
|---|-------|--------|--------------------------|
| 1 | SKILL.md frontmatter + RU description + каркас | ✓ VERIFIED | Файл вне дрейфа — цел |
| 2 | Single source: генератор парсит таблицу | ✓ VERIFIED | cmd/skillgen вне дрейфа |
| 3 | Синхрон-гейт: -check + golden-тест | ✓ VERIFIED | `go run ./cmd/skillgen -check` → **exit 0**; TestSkillgen_CommittedFileInSync → **ok -race** (прогон верификатора, CONFIG.md нетронут quick'ом — подтверждено пустым diff'ом) |
| 4 | Dev-only mise-задача skillgen-regen | ✓ VERIFIED | mise.toml вне дрейфа |
| 5 | TRACKING-гейт: SKILL.md в git-индексе | ✓ VERIFIED | git-tracked |
| 6 | Регенерация идемпотентна | ✓ VERIFIED | -check exit 0: закоммиченный файл байт-идентичен выводу генератора |

#### Plan 08-09 — Дельта-ревизия владельца (7 truths)

| # | Truth | Status | Evidence (текущее дерево) |
|---|-------|--------|--------------------------|
| 1 | SPEC §4 ревизия 2026-10-06, дословная цитата, аудит-цел | ✓ VERIFIED | SPEC.md:268–277 прочитано в дереве; цитата «…список не нужен, а нужен bool параметр, по дефолту…» на :277; новейший хунок (:123) её не касается |
| 2 | Схема без Apps-машинерии; единственный EffectiveEnabled; Validate без a11y | ✓ VERIFIED | config.go:184–197; maxA11yApps/errA11yApps = 0 (internal/config вне дрейфа) |
| 3 | Default ON запинен; Load не накладывает Defaults | ✓ VERIFIED | Корпус TestLoad_A11y* + TestDefaults_A11yDefaultOn — **ok -race** |
| 4 | Громкая миграция D-33 | ✓ VERIFIED | TestLoad_A11yRemovedListKeyRejectedWhole — в корпусе **ok -race** |
| 5 | Фолд на EffectiveEnabled; контракты 08-03/08-04 не меняются | ✓ VERIFIED | actor.go:1361/1396 целы после flipCredit; internal/a11y вне дрейфа; 08-04 PLAN grep-coupling к wrap = 0 |
| 6 | Доки/skill согласованы | ✓ VERIFIED | grep a11y.apps = 0 в CONFIG.md/SKILL.md/README×2; одна a11y-строка bool/true; -check exit 0 |
| 7 | Зелёная итерация | ✓ VERIFIED | Точечные -race прогоны верификатора: skillgen ok, config A11y ok, session A11y ok (5), install A11y ok (8), `go build ./...` exit 0 |

#### Plan 08-10 — Гэп-закрытие README (3 truths)

Все 3 ✓ VERIFIED — README-пара вне дрейфа; grep-гейты (негатив a11y.apps = 0, позитивные якоря 7/7, паритет 22/22) повторены верификатором в текущем дереве — PASS.

### Superseded-and-Honored Truths (плановая цепочка)

Семь истин планов 08-01..08-06 (a11y.apps-машинерия) остаются замещёнными ревизией D-8-1/REV через план 08-09 — цепочка и верификации замен не изменились; все семь замен переподтверждены в текущем дереве (таблицы 08-02/08-05/08-06 выше).

### UAT Dispositions (owner verdicts 2026-10-06 — транскрибированы в дерево коммитом 2171fae)

08-UAT.md вне дрейфа e76f6b0..HEAD: `status: complete`, 6/6 pass. Все 6 human-пунктов остаются разрешёнными; новых human-пунктов этот проход не породил.

### Key Link Verification

| From | To | Via | Status | Details (текущее дерево) |
|------|----|----|--------|-------------------------|
| docs/SPEC.md §4 (ревизия 2026-10-06) | internal/config (A11y) | контракт секции | ✓ WIRED | Схема = ревизии: config.go:184–197; config вне дрейфа |
| internal/config A11y.EffectiveEnabled | internal/session/actor.go (pushA11y) | snap.A11y.EffectiveEnabled() | ✓ WIRED | actor.go:1396 (вызов фолда :1361); пережил flipCredit-дифф; тесты -race ok |
| internal/session/actor.go | internal/a11y Reconciler.Apply | A11ySink seam | ✓ WIRED | seam цел (5/5 -race ok); main.go:210 |
| cmd/goswitchd/main.go | internal/a11y production-адаптеры | a11y.New(NewExecRunner, NewDBusStatusSetter) | ✓ WIRED | main.go:210; main.go вне дрейфа |
| internal/install restoreToolkitAccessibility | install-state.json | toolkit_accessibility | ✓ WIRED | :242/:460/:1096–1110 целы после G-5-5; корпус 8/8 -race ok |
| docs/CONFIG.md таблица | cmd/skillgen → SKILL.md | parseKeyTable + regen | ✓ WIRED | a11y-строка :77 байт-цела; -check exit 0 |
| docs/SPEC.md ревизия 2026-10-06 | README.md/README.en.md a11y-разделы | doc-сверка D-8-8 | ✓ WIRED | Оба README вне дрейфа e76f6b0..HEAD; grep-гейты PASS |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| pushA11y fold | active bool | snap.A11y.EffectiveEnabled() (загруженный конфиг) | Yes | ✓ FLOWING |
| Reconciler series | desired bool | sink.Apply | Yes | ✓ FLOWING |
| Defaults() | A11y ON | boolPtr(true) в схеме | Yes | ✓ FLOWING |
| skillgen region | table rows | docs/CONFIG.md (единственный источник) | Yes | ✓ FLOWING |
| restoreToolkitAccessibility | value | install-state.json | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

Все прогоны верификатора 2026-10-07T23:1xZ, -race, single-named (полный suite не запускался).

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Сборка дерева после flipCredit+G-5-5 | `go build ./...` | exit 0 | ✓ PASS |
| Фолд актора после flipCredit (named -race) | `go test -race -run 'A11y' ./internal/session` | ok 1.013s — 5/5 PASS (Defaults/FoldTriggers/ExplicitFalse/SinkSelfSyncs/NilSinkNoOp) | ✓ PASS |
| Config a11y-пины (named -race) | `go test -race -run 'A11y' ./internal/config` | ok 1.042s | ✓ PASS |
| Install a11y-корпус после G-5-5 (named -race) | `go test -race -run 'A11y' ./internal/install` | ok 1.028s — 8/8 PASS (3× Install_A11y + 5× Uninstall_A11y) | ✓ PASS |
| Golden-синхрон (single named -race) | `go test -race -run TestSkillgen_CommittedFileInSync ./cmd/skillgen` | ok 1.010s — PASS | ✓ PASS |
| Skillgen байт-синхрон | `go run ./cmd/skillgen -check` | CHECK-EXIT-0 | ✓ PASS |
| Ноль ключа-списка | `grep -c "a11y\.apps" README.md README.en.md docs/CONFIG.md skills/goswitch-config/SKILL.md` | 0/0/0/0 | ✓ PASS |
| Позитивные якоря bool-контракта | grep a11y.enabled/toolkit-accessibility в README×2 | 7/7 в обоих языках | ✓ PASS |
| Паритет заголовков | `grep -cE '^##+ '` | 22/22 в обоих файлах | ✓ PASS |
| CONFIG.md нетронут quick'ом | `git diff e76f6b0..HEAD -- docs/CONFIG.md skills/goswitch-config/SKILL.md` | пусто | ✓ PASS |
| A11y-ядро вне дрейфа | `git diff --stat e76f6b0..HEAD -- internal/a11y/` | пусто | ✓ PASS |
| SPEC a11y-ревизия цела | grep 2026-10-06 + вербатим-цитата | :268–277, цитата на :277 | ✓ PASS |
| 08-04 PLAN не связан с wrap-семантикой | grep -icE "wrapSources\|errMixedSources\|half" 08-04-PLAN.md | 0 | ✓ PASS |
| Debt-маркеры в затронутых файлах | grep TBD/FIXME/XXX/HACK/PLACEHOLDER по 11 tree-файлам дрейфа | чисто (exit 1) | ✓ PASS |
| Существование covered-файлов | existence loop | 43/43 OK | ✓ PASS |

### Probe Execution

Не применимо — REQUIREMENTS.md не маппит фазу 8 (spec-less fallback, подтверждено предыдущими проходами); scripts/*/tests/probe-*.sh в репо отсутствуют. Живой e2e не запускался (по скоупу refresh'а).

### Requirements Coverage

Трассировка по 08-CONTEXT.md (backlog-derived D-8-1..D-8-11 + Owner Revision 2026-10-06) — все статусы предыдущего прохода подтверждены в текущем дереве:

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| D-8-1 → D-8-1/REV | 08-01/02/05/09 | Секция (пересмотрена: единственный bool-ключ default ON) | ✓ SATISFIED | config.go:184–197 + фолд actor.go:1361/1396 (после flipCredit) + SPEC-ревизия :268–277 |
| D-8-2 | 08-01 | Механика исследована при планировании | ✓ SATISFIED | 08-RESEARCH; резолюции в ревизиях §4 |
| D-8-3 | 08-03/05 | Демон-reconcile при старте и hot reload | ✓ SATISFIED (unit + UAT-accepted) | Фолд/wiring/тесты -race ok в текущем дереве |
| D-8-4 | 08-03/04 | Откат только uninstall; снятие запрещено | ✓ SATISFIED (unit + UAT-accepted) | Apply(false) ноль-касаний; a11y-корпус 8/8 -race ok после G-5-5 |
| D-8-5 → D-8-5/REV | 08-02 → 08-09 | Regex-подстрока RE2 — АННУЛИРОВАНА | ✓ SUPERSEDED-AND-HONORED | Машинерия удалена; config.go:169 фиксирует удаление |
| D-8-6 → D-8-6/REV | 08-01/03/09 | Минимальный набор магии + default ON | ✓ SATISFIED | toolkit-accessibility + IsEnabled-пояс нетронуты; nil=ON |
| D-8-7 | 08-07 | README русский-первый | ✓ SATISFIED | README.md(RU)/README.en.md(EN); паритет 22/22; вне дрейфа |
| D-8-8 | 08-06/07/10 | Сверка доков с поведением | ✓ SATISFIED | SPEC/CONFIG/SKILL/README честны; golden-гейт зелёный на нетронутом golden'е |
| D-8-9 | 08-06 | CONFIG.md русский-первый | ✓ SATISFIED | CONFIG.md по-русски; вне дрейфа |
| D-8-10 | 08-08 | Single source: генерация из CONFIG.md | ✓ SATISFIED | skillgen + golden + -check exit 0 |
| D-8-11 | 08-08 | skills/goswitch-config/ | ✓ SATISFIED | git-tracked |

Orphaned requirements: нет. Отложенных к поздним фазам нет — фаза 8 последняя в milestone.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | TBD/FIXME/XXX/HACK/PLACEHOLDER в затронутых дрейфом файлах (11 tree-файлов flipCredit/G-5-5/spec-delta) | none | Скан чист |
| — | — | Стабы/empty-implementation в a11y-коде и новом heal-коде selfcheck | none | Heal-код самодостаточен (wrap → set → read-back verify → verdict) |

### Advisory (New Scope, Unevidenced)

None — re-verification ran; new-scope findings отсутствуют: затронутые файлы просканированы, debt-маркеров нет, новые selfcheck/install пути покрыты -race тестами quick'а и не пересекаются с контрактами фазы 8.

### Human Verification Required

Нет — все 6 прежних human-пунктов разрешены записанными вердиктами владельца и транскрибированы в дерево (08-UAT.md, `status: complete`, коммит 2171fae; файл вне дрейфа e76f6b0..HEAD). Изменения дерева с прошлого прохода (quick 261008-00m) покрыты автоматическими гейтами этого прохода (именные -race тесты всех затронутых швов); flipCredit-поведение (G-2-3 — own flip сохраняет буфер) и G-5-5 auto-wrap — контракты других фаз/quick'а, верифицированные их собственными отчётами, и не входят в цели фазы 8. Новых human-пунктов не возникло.

### Final Verdict

Staleness refresh #2 на закрытии milestone: дрейф e76f6b0..39a5d82 (26 коммитов quick 261008-00m — flipCredit/G-2-3, G-5-5 auto-wrap + selfcheck heal, ADR-004/ADR-006 spec-delta) не задел ни одну истин фазы 8 функционально. A11y-схема (`a11y.enabled` bool default ON) и reconciler — без изменений (internal/a11y, internal/config, main.go вне дрейфа); фолд actor.go:1361/1396 пережил flipCredit-дифф (5/5 -race ok); installer a11y-snapshot/restore пережил G-5-5 (8/8 -race ok; 08-04 PLAN grep-coupling к wrap = 0); SPEC a11y-ревизия 2026-10-06 цела, новые дельты — audit-trail-дисциплиной; CONFIG.md/SKILL.md/README-пара нетронуты вовсе (пустые diff'ы); golden-гейт зелёный (TestSkillgen_CommittedFileInSync ok -race + -check exit 0); дерево собирается (`go build ./...` exit 0). 44/44 in-force истин verified, 0 гэпов, 0 behavior-unverified, 0 регрессий, 0 advisory. Фаза 8 готова к ship вместе с milestone.

---

_Verified: 2026-10-07T23:22:23Z_
_Verifier: Claude (gsd-verifier)_
_History: 44/46 human_needed (2026-10-05T23:55:00Z, d61f26e) → 43/44 gaps_found (2026-10-06T01:55:43Z, cbe4160) → 44/44 passed (2026-10-06T05:53:08Z, 2ca9187) → 44/44 passed — refresh #1 (2026-10-07T21:32:42Z, ae54bf4) → 44/44 passed — refresh #2, после quick 261008-00m (этот отчёт). Предыдущие отчёты preserved in git history._
