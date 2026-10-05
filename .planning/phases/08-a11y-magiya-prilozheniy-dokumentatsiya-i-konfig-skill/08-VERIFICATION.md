---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
verified: 2026-10-05T23:55:00Z
status: human_needed
score: 44/46 must-haves verified
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
  - cmd/goswitchd/main.go
  - cmd/skillgen/main.go
  - cmd/skillgen/main_test.go
  - skills/goswitch-config/SKILL.md
  - mise.toml
  - docs/SPEC.md
  - docs/CONFIG.md
  - README.md
  - README.en.md
  - docs/ACCEPTANCE.md
covered_digest: "v1:sha256:5fbcf61a0638c73d59b134f7761fd2e953a88a169fcd51e2f3a11824bd340004"
behavior_unverified: 2
behavior_unverified_items:
  - truth: "Live desktop: daemon at start/hot reload really sets org.gnome.desktop.interface toolkit-accessibility via the real gsettings and the real session bus accepts the org.a11y.Status.IsEnabled Properties.Set (08-03 D7 / 08-05 D6 coverage)"
    test: "On the owner's GNOME session: add an app to a11y.apps, enable a11y.enabled, restart goswitchd; verify `gsettings get org.gnome.desktop.interface toolkit-accessibility` returns true and `busctl --user get-property org.a11y.Bus /org/a11y/bus org.a11y.Status IsEnabled` is true; then restart a listed Electron app and check its a11y tree is exposed (08-VALIDATION Manual-Only row 1)"
    expected: "Key and belt set to true; listed app exposes the a11y/AT-SPI tree after ITS OWN restart (one-shot Chromium/Electron start check)"
    why_human: "Requires a real GNOME session, real gsettings/dconf and a real session bus; unit corpus proves the contract through Runner/StatusSetter seams, not the production adapters against a live desktop"
  - truth: "Live uninstall: goswitchctl uninstall restores the pre-install toolkit-accessibility value from install-state.json and never resets a manually-enabled key (08-04 D7 coverage)"
    test: "Install with a11y active (or with the key off), then run goswitchctl uninstall and check `gsettings get org.gnome.desktop.interface toolkit-accessibility` equals the pre-install value (08-VALIDATION Manual-Only row for D-8-4)"
    expected: "Key restored verbatim from the snapshot; on an old install whose state file lacks the field the key is left as-is (WARN, zero set calls)"
    why_human: "Touches the owner's real session (daemon stop, sources restore); unit corpus proves it through the fakeRunner seam"
overrides_applied: 0
coincidental_reliance_items: []
advisory:
  - finding: "Operational docs (docs/ACCEPTANCE.md, docs/ci-runner.md, SECURITY.md) are already in Russian on this tree, while plan 08-06's D-8-9 assumption described them as EN ('остаются EN'). Phase 8 did not translate them — the only touch was a 2-line link repoint in docs/ACCEPTANCE.md (commit b65b62f), which is the correct D-8-9-compliant behavior for this phase."
    category: other
    reason: "Pre-existing tree state, not a phase-8 action; noted so the 'остаются EN' wording in plans is not read as a claim about current file language. No phase action required."
    evidence_status: "git log -1 -- docs/ACCEPTANCE.md = b65b62f (+2/−2 link repoint only)"
---

# Phase 8: a11y-магия приложений, документация и конфиг-skill — Verification Report

**Phase Goal:** Разобранный беклог (3 todo от 2026-10-05): (1) a11y-магия — конфиг-секция приложений; демон применяет её сам (reconcile при старте/reload; revert при uninstall; глобальный toolkit-accessibility + D-Bus IsEnabled «пояс»). (2) Документация — сверка README/CONFIG.md/SPEC с фактическим поведением v1.1.0; реструктура «русский-первый». (3) Skill «Конфигурация goswitch» для AI — SKILL.md в skills/goswitch-config/, конфиг-раздел генерируется из docs/CONFIG.md (golden-гейт в mise run test).
**Verified:** 2026-10-05T23:55:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

All three streams are implemented, wired and green in the tree. 44 of 46 must-have truths verified (plan-frontmatter truths all verified; 2 goal-level live-desktop behaviors present+wired but not exercised by any automated test — Manual-Only per 08-VALIDATION). One known deferred owner decision (WR-02) is honestly escalated as a todo, not silently dropped.

### Observable Truths

#### Plan 08-01 — SPEC §4 spec-delta (5 truths + spec-less-fallback note)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SPEC §4 несёт датированную ревизию 2026-10-05 ДО кода; ноль .go в дифе; аудит-трейл цел | ✓ VERIFIED | Ревизия-блок docs/SPEC.md:117–175; git show 5293daa --numstat = `62 0 docs/SPEC.md`; прежние маркеры на месте (2026-09-27 ×8, 2026-10-04 ×4, D-53 ×4) |
| 2 | Ревизия фиксирует секцию a11y (enabled/apps, потолок 64, zero-off) и две ручки магии | ✓ VERIFIED | SPEC.md:128–141 (`a11y.enabled`, `a11y.apps`, RE2, потолок 64, zero-off; toolkit-accessibility + org.a11y.Status.IsEnabled через Properties.Set) |
| 3 | Ревизия фиксирует daemon-reconcile (старт+hot reload), отказ от динамического снятия, uninstall-revert only-if-present | ✓ VERIFIED | SPEC.md:147–169 («Момент применения», «Откат», install-state.json, «никогда не восстановить выключенное») |
| 4 | Ревизия фиксирует restart-семантику (одноразовая стартовая проверка Chromium/Electron) | ✓ VERIFIED | SPEC.md:142–146 («при следующем запуске приложения … не подписываются на изменения») |
| 5 | Ревизия фиксирует резолюцию per-app оверрайдов (документируется, вне v1-дифа) и argv-дисциплину | ✓ VERIFIED | SPEC.md:169–175 (ASVS V5 argv-абзац; «механизм документируется, вне v1-дифа»); рецепт реализован в CONFIG.md:344 |
| 6 | Spec-less fallback пропущен ВИДИМО | ✓ VERIFIED | Записано в 08-01-PLAN must_haves и 08-01-SUMMARY; REQUIREMENTS.md не содержит строк фазы 8 (backlog-derived) — подтверждено grep'ом |

#### Plan 08-02 — Схема секции a11y (5 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Секция A11y (yaml a11y: enabled/apps), zero-value off, doc-коммент «семь секций» | ✓ VERIFIED | config.go:183–207 (`type A11y`, `A11y \`yaml:"a11y"\``, «exactly the seven sections» :197) |
| 2 | Валидация на Load: безусловный потолок 64, blank-отказ до Compile, errA11yApps* сентинелы, ошибки «поле+индекс» | ✓ VERIFIED | config.go:32,69–71,440–448; тесты TestValidate_A11yCeilingUnconditional / A11yBlankPattern / A11yRegexCompile — PASS |
| 3 | Active() = enabled && непустой список; Defaults() выключенная секция | ✓ VERIFIED | config.go:193–195; TestA11yActive (4 сабкейса) + TestDefaults_A11yOff — PASS |
| 4 | Strict decode D-33 и last-good распространяются автоматически; load.go/watch.go — ноль правок | ✓ VERIFIED | TestLoad_A11yStrictDecodeUnknownKey — PASS; git diff по фазе: load.go/watch.go отсутствуют в списке изменённых (только их тесты) |
| 5 | Hot reload пин: битый паттерн → reload отклонён, last-good держится | ✓ VERIFIED | TestWatch_BrokenA11yPatternKeepsLastGood в watch_test.go (код присутствует; пакет internal/config зелёный в mise run ci) |

#### Plan 08-03 — Reconciler internal/a11y (6 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Apply(active bool) fire-and-forget, серия на собственной сериализации ВНЕ акторного мьютекса | ✓ VERIFIED | a11y.go:99 (`seriesMu` — единственный лок через I/O), :124–137 (Apply касается только state-lock + spawn); **усилено CR-01-фиксом** (1927bf9): TestA11y_ApplyDoesNotWaitForInFlightSeries — PASS под -race (0.10s) |
| 2 | Read-verify-then-set; Apply(false) — ноль подпроцессов | ✓ VERIFIED | TestA11y_ActivateSetsKeyWhenFalse / AlreadyTrueNoSet / DeactivateNeverTouchesKey — PASS |
| 3 | Пояс org.a11y.Status.IsEnabled при каждой активации через StatusSetter; production-адаптер godbus | ✓ VERIFIED | a11y.go:57–68,337 (org.a11y.Bus /org/a11y/bus, IsEnabled, dbus.MakeVariant(true)); TestA11y_BeltSetOnEveryActivation / BeltNotCalledOnDeactivation — PASS |
| 4 | Argv-дисциплина: argv только из package-литералов, список приложений не входит в seam | ✓ VERIFIED | a11y.go:38–39 литералы; TestA11y_FixedArgvOnlyLiterals / TestA11yWireLiterals — PASS |
| 5 | Degradation: закрытый reason-словарь, один WARN за эпизод, успешная активация переоткрывает бюджет | ✓ VERIFIED | a11y.go:50 (reasonBelt) и словарь; TestA11y_BeltFailureWarnOnce / ReadFailureWarnOnceAndFailsTowardDesired / SuccessfulEpisodeReopensBudget — PASS |
| 6 | Сериализация/сходимость последовательности Apply под -race | ✓ VERIFIED | TestA11y_ApplyDoesNotBlockAndConverges + TestA11y_MidEpisodeDesiredChangeHonoredByFollowUp — PASS под -race |

#### Plan 08-04 — Installer snapshot/revert (5 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | installState.ToolkitAccessibility (json toolkit_accessibility), verbatim; отказ чтения → пусто+WARN | ✓ VERIFIED | install.go:242,565; TestInstall_A11ySnapshotCaptured / SnapshotTrueValue / SnapshotReadFailureLeavesEmpty — PASS |
| 2 | Idempotent-backup не тронут | ✓ VERIFIED | TestInstall_SecondInstallKeepsOriginalBackup — PASS (запущен именованно) |
| 3 | Restore only-if-present + shape-валидация true\|false, никогда «restore false», ДО удаления state-файла | ✓ VERIFIED | install.go:460 (вызов после restoreSwitchBinding, перед os.Remove:456-460 блок); savedA11yState трёхсостоянийный :1117; TestUninstall_A11yRestoreTrue/False/MissingFieldSkipsSilently/CorruptValueSkips/RestoreFailureReported — PASS |
| 4 | Argv-дисциплина restore: только shape-валидированный литерал | ✓ VERIFIED | install.go:1101 (set с `value` после shape-гейта); TestUninstall_A11yCorruptValueSkips — PASS |
| 5 | Место в цепочке uninstall: после restoreSwitchBinding, до purge | ✓ VERIFIED | install.go:452–466 (restoreSources → restoreSwitchBinding → restoreToolkitAccessibility → os.Remove(statePath) → purgeDirs) |

#### Plan 08-05 — Фолд актора + wiring (4 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Фолд через snap.A11y.Active() с дифф-гейтом, вызов sink неблокирующий, ВНЕ actor.mu | ✓ VERIFIED | actor.go:1332–1335 (pushA11y: только Active(), дифф-гейт по a11yActive); TestActor_A11yFoldDiffGated — PASS; контендед-инвариант подкреплён TestA11y_ApplyDoesNotWaitForInFlightSeries (CR-01-фикс) |
| 2 | A11ySink seam + SetA11ySink с self-sync; nil — no-op | ✓ VERIFIED | actor.go:930,942–947; TestActor_A11ySinkInstallSelfSyncs / NilSinkNoOp — PASS |
| 3 | Wiring демона: reconciler на production-адаптерах рядом с SetSoundSink; без новых watcher-веток | ✓ VERIFIED | main.go:201 `actor.SetA11ySink(a11y.New(a11y.NewExecRunner(), a11y.NewDBusStatusSetter()))`; в фазовом дифе main.go нет fsnotify/NewWatcher |
| 4 | Двойная идемпотентность (актор diff-gate + reconciler diff-gate) | ✓ VERIFIED | Оба гейта кодом + тестами (см. 08-03 #2/#6, 08-05 #1) |

#### Plan 08-06 — CONFIG.md русский-первый + SPEC-сверка (7 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | CONFIG.md русский-первый; операционные доки не переводились этой фазой | ✓ VERIFIED | head docs/CONFIG.md — русский; ACCEPTANCE.md тронут только перепoint'ом ссылок (b65b62f, +2/−2) — см. advisory |
| 2 | Построчная сверка таблицы со схемой | ✓ VERIFIED | 29 строк `\| \``; a11y-строки байт-согласованы (false / потолок 64); две сверка-правки (tap_key bare, autocorrect_event) отражены в SUMMARY |
| 3 | SPEC-сверка в форме 08-01, аудит-трейл цел | ✓ VERIFIED | Сверка-блок SPEC.md:189+; единственная ин-плейс правка (install не пишет конфиг) задокументирована в сверка-блоке; ca30cbe numstat 49/2 |
| 4 | «Ровно семь секций» + ровно 2 a11y-строки | ✓ VERIFIED | CONFIG.md:14,72–73 (`grep -c '^| \`a11y'` = 2) |
| 5 | Поведение магии задокументировано честно (старт+reload, restart-семантика, снятие запрещено, uninstall-revert) | ✓ VERIFIED | CONFIG.md:240–252 (uninstall-откат), пример + Privacy; оговорка: строка a11y.apps:73 описывает per-app матчинг — см. WR-02 WARNING ниже |
| 6 | Per-app .desktop-оверрайды — ручной рецепт без кода | ✓ VERIFIED | CONFIG.md:344 (Exec=env ACCESSIBILITY_ENABLED=1 … --force-renderer-accessibility) + предупреждения |
| 7 | Формат таблицы байт-строгий (источник генератора) | ✓ VERIFIED | TestSkillgen_ParseRealConfigTable (21 строк, 5 колонок) + TestSkillgen_CommittedFileInSync — PASS |

#### Plan 08-07 — README русский-первый (5 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | README.md — русская главная с полным v1.1.0-контентом + a11y-раздел | ✓ VERIFIED | README.md русский (шапка, 22 `##`), раздел «A11y-магия приложений» :301–319; качество прозы → human |
| 2 | README.en.md — новый английский канон, паритет раздел-в-раздел | ✓ VERIFIED | README.en.md английский, паритет заголовков 22/22, кросс-ссылки в обеих шапках |
| 3 | README.ru.md удалён из git, дублей нет, stale-ссылок нет | ✓ VERIFIED | `git ls-files -- README.ru.md` пуст; grep README\.ru по README/docs/*.md — ноль |
| 4 | a11y-раздел в README с restart-семантикой и uninstall-откатом, без дубля справочника | ✓ VERIFIED | README.md:301+ (restart/uninstall), ноль строк `\| \`` в README (таблица не дублируется), ссылка на docs/CONFIG.md есть |
| 5 | Входящие ссылки перепpoint'нуты | ✓ VERIFIED | docs/ACCEPTANCE.md:38,60 → README.en.md#установка/…; все ](README…) цели существуют |

#### Plan 08-08 — Skill goswitch-config (6 truths)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | skills/goswitch-config/SKILL.md: frontmatter name goswitch-config + RU description; ручной каркас + регион между сентинелами | ✓ VERIFIED | SKILL.md:1–5 frontmatter; сентинелы BEGIN/END (2 вхождения); процедуры/диагностика по-русски |
| 2 | Single source: генератор парсит таблицу CONFIG.md, verbatim-копирование, hard errors | ✓ VERIFIED | cmd/skillgen/main.go:31,36–37,129 (parseKeyTable, 5 колонок, file:line ошибки); TestSkillgen_MalformedRowHardError — PASS |
| 3 | Синхрон-гейт: -check байт-в-байт + golden-тест в mise run test | ✓ VERIFIED | main.go:122 (bytes.Equal); `go run ./cmd/skillgen -check` → exit 0; TestSkillgen_CommittedFileInSync — PASS в полном прогоне |
| 4 | Dev-only mise-задача skillgen-regen; CI не перегенерирует | ✓ VERIFIED | mise.toml:221–223; workflow-файлы не тронуты (в фазовом дифе отсутствуют) |
| 5 | TRACKING-гейт: SKILL.md в git-индексе | ✓ VERIFIED | `git ls-files -- skills/goswitch-config/SKILL.md` непуст |
| 6 | Регенерация идемпотентна | ✓ VERIFIED | CommittedFileInSync байт-в-байт после regen (CHECK-EXIT-0); регион содержит 21 строку ключей |

**Score:** 44/46 (44 verified; 2 goal-level live-desktop behaviors present+wired but not exercised — behavior_unverified: 2). Полный прогон `mise run ci` (build + vet + golangci-lint strict + test -race ./... + tidy-diff): **exit 0**, все 24 пакета, включая internal/a11y, internal/config, internal/install, internal/session, cmd/skillgen.

### Review-Fix Verification (08-REVIEW → 08-REVIEW-FIX)

| Finding | Claimed disposition | Verified in tree |
|---------|--------------------|------------------|
| CR-01 (reconciler mutex across I/O) | fixed: commits 3db244d (RED) + 1927bf9 (GREEN) | ✓ VERIFIED — a11y.go:99 `seriesMu` split, state-lock никогда не держится через I/O; regression-тесты TestA11y_ApplyDoesNotWaitForInFlightSeries + TestA11y_MidEpisodeDesiredChangeHonoredByFollowUp присутствуют и PASS под -race. Ревью помечало «requires human verification» — тест-корпус пинит контракт; человеческий проход записан в Human Verification (advisory) |
| WR-01 (SKILL.md ложное утверждение про switch-input-source) | fixed: commit 8d68ea4 | ✓ VERIFIED — SKILL.md:61,140 и CONFIG.md:55,292 теперь «биндинг GNOME switch-input-source install не трогает»; регион перегенерирован (skillgen -check exit 0) |
| WR-02 (a11y.apps matching semantics) | skipped — owner decision deferred as todo | ✓ HONESTLY DEFERRED — todo .planning/todos/pending/a11y-apps-matching-semantics.md существует (commit bcd2670 в HEAD), REVIEW-FIX документирует skip; a11y-строки CONFIG.md намеренно не тронуты fixer'ом |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| internal/config A11y.Active() | internal/session/actor.go (pushA11y) | `snap.A11y.Active()` | ✓ WIRED | actor.go:1333; ACTIVE-семантика в одном месте (нет inline enabled/len) |
| internal/session/actor.go | internal/a11y Reconciler.Apply | A11ySink seam | ✓ WIRED | actor.go:930–947; main.go:201 собирает reconciler и ставит sink |
| cmd/goswitchd/main.go | internal/a11y production-адаптеры | a11y.New(NewExecRunner, NewDBusStatusSetter) | ✓ WIRED | main.go:201; имена конструкторов совпадают с запиненными 08-03 |
| internal/install restoreToolkitAccessibility | install-state.json snapshot | toolkit_accessibility поле | ✓ WIRED | install.go:242,565,1087–1135; порядок restore-до-удаления подтверждён |
| docs/CONFIG.md таблица | cmd/skillgen | parseKeyTable | ✓ WIRED | 21 строка парсится; CommittedFileInSync байт-синхронен; -check exit 0 |
| docs/SPEC.md §4 ревизия | кодовые планы 08-02..08-05 | контракт a11y | ✓ WIRED | Реализация совпадает с ревизией (секция/две ручки/reconcile/restart/uninstall-revert/argv) |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| pushA11y fold | active bool | snap.A11y.Active() из загруженного конфига | Yes (config.Load → FoldAppliedConfig/applySnapshot) | ✓ FLOWING |
| Reconciler series | desired bool | sink.Apply от фолда | Yes | ✓ FLOWING |
| gsettings argv | literals only | package-константы (по дизайну ASVS V5) | n/a (список приложений намеренно не достигает argv) | ✓ FLOWING |
| skillgen region | table rows | docs/CONFIG.md (единственный источник) | Yes (verbatim, golden-гейт) | ✓ FLOWING |
| restoreToolkitAccessibility | value | install-state.json (shape-валидированный литерал) | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| CR-01 contended Apply не ждёт серию | `go test ./internal/a11y/ -race -run 'TestA11y_ApplyDoesNotWaitForInFlightSeries\|TestA11y_MidEpisodeDesiredChangeHonoredByFollowUp' -v` | 2/2 PASS (0.10s each) | ✓ PASS |
| Конфиг-валидация a11y (потолок/blank/regex/Active/strict-decode) | `go test ./internal/config/ -race -run 'TestValidate_A11y…\|TestA11yActive\|TestLoad_A11yStrictDecodeUnknownKey' -v` | 5/5 PASS | ✓ PASS |
| Снапшот/restore установщика | `go test ./internal/install/ -race -run 'TestInstall_A11y\|TestUninstall_A11y\|TestInstall_SecondInstallKeepsOriginalBackup' -v` | 9/9 PASS | ✓ PASS |
| Фолд актора (6 кейсов) | `go test ./internal/session/ -race -run 'TestActor_A11y' -v` | 6/6 PASS | ✓ PASS |
| Skillgen golden-гейт + -check | `go test ./cmd/skillgen/ -race -count=1` + `go run ./cmd/skillgen -check` | 5/5 PASS; CHECK-EXIT-0 | ✓ PASS |
| Полная зелёная итерация | `mise run ci` | exit 0, 24 пакета ok | ✓ PASS |

### Probe Execution

Не применимо: у фазы нет requirement-ID для probe-падения (spec-less fallback пропущен видимо в must_haves всех планов; источник требований — D-8-1..D-8-11 из 08-CONTEXT.md). Скриптов scripts/*/tests/probe-*.sh в репо нет.

### Requirements Coverage

REQUIREMENTS.md не содержит строк для фазы 8 (подтверждено grep'ом) — трассировка по решениям 08-CONTEXT.md, как объявлено в ROADMAP:

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| D-8-1 | 08-01/02/05 | Секция в конфиге (не перечень в доках) | ✓ SATISFIED | Схема + фолд + доки; a11y.enabled/a11y.apps канон |
| D-8-2 | 08-01 | Механика исследована при планировании | ✓ SATISFIED | 08-RESEARCH live-пробы; резолюция записана в §4 ревизии |
| D-8-3 | 08-03/05 | Демон-reconcile при старте и hot reload | ✓ SATISFIED (unit) | Фолд + wiring + тесты; живой эффект → human |
| D-8-4 | 08-03/04 | Откат только uninstall; снятие запрещено | ✓ SATISFIED (unit) | Apply(false) ноль-касаний; snapshot/restore only-if-present; живой revert → human |
| D-8-5 | 08-02 | Regex-подстрока RE2, валидация на Load | ✓ SATISFIED (валидация) | Компиляция на Load с именованными ошибками; **runtime-матчинг не существует — WR-02, отложено владельцу** |
| D-8-6 | 08-01/03 | Минимальный набор магии (глобальный ключ + пояс), без root | ✓ SATISFIED | toolkit-accessibility read-verify-then-set + IsEnabled belt через godbus session bus |
| D-8-7 | 08-07 | README.ru как главная, README как перевод | ✓ SATISFIED | README.md (RU) / README.en.md (EN) / README.ru.md удалён; качество → human |
| D-8-8 | 08-06/07 | Сверка доков с поведением; дефект → todo | ✓ SATISFIED | Сверка-ревизия SPEC + 3 док-правки README; дефектов поведения не найдено (кроме WR-02 — todo заведён) |
| D-8-9 | 08-06 | CONFIG.md русский-первый | ✓ SATISFIED | CONFIG.md по-русски; операционные доки фазой не переводились |
| D-8-10 | 08-08 | Single source: генерация из CONFIG.md | ✓ SATISFIED | cmd/skillgen + golden-тест в mise run test + -check; дрейф = красный CI |
| D-8-11 | 08-08 | skills/goswitch-config/ (не машинная .zcode/) | ✓ SATISFIED | SKILL.md в skills/goswitch-config/, git-tracked |

Orphaned requirements: нет (REQUIREMENTS.md не маппит фазу 8 — расхождений план↔REQUIREMENTS быть не может).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | TBD/FIXME/XXX/HACK/PLACEHOLDER в файлах фазы | none | Скан чист (internal/a11y, config.go, cmd/skillgen, SKILL.md) |
| — | — | Стабы/empty-implementation в новом коде | none | `return nil` в a11y.go:354 и skillgen — легитимные success-пути, проверены чтением |
| docs/CONFIG.md | 73 | Строка a11y.apps описывает per-app regex-матчинг, которого в рантайме нет (только non-emptiness гейт Active()) | ⚠️ Warning | Известный WR-02 — отложено владельцу как todo (a11y-apps-matching-semantics.md); docs advertising > daemon behavior, решение за владельцем |

### Human Verification Required

### 1. Живой эффект a11y-магии (Manual-Only, 08-VALIDATION)

**Test:** Включить a11y.enabled + добавить приложение в a11y.apps, перезапустить goswitchd; проверить `gsettings get org.gnome.desktop.interface toolkit-accessibility` → true и `busctl --user get-property org.a11y.Bus /org/a11y/bus org.a11y.Status IsEnabled` → true; перезапустить приложение из списка и убедиться, что a11y/AT-SPI-дерево появилось.
**Expected:** Обе ручки true; приложение из списка exposes a11y-дерево после СВОЕГО перезапуска.
**Why human:** Нужны реальная GNOME-сессия, реальный gsettings/dconf и реальная сессионная шина; юнит-корпус доказывает контракт через seam-фейки, не production-адаптеры на живом столе.

### 2. Живой uninstall-revert ключа (08-04 D7)

**Test:** На столе владельца выполнить goswitchctl uninstall и сверить `gsettings get org.gnome.desktop.interface toolkit-accessibility` с пред-установочным значением; отдельно — на старом инсталле без a11y-поля в state-файле убедиться, что ключ не сброшен.
**Expected:** Ключ восстановлен verbatim из снапшота; старый инсталл переживает uninstall без сброса вручную-включённого ключа.
**Why human:** Трогает реальную сессию владельца (остановка демона, восстановление sources).

### 3. Качество русской прозы README/CONFIG/SKILL и синхронность пары README (08-06/08-07 human_judgment)

**Test:** Прочитать README.md (RU) / README.en.md, сверить раздел-в-раздел; просмотреть CONFIG.md и SKILL.md.
**Expected:** Читаемый русский, синхронная пара, никакой потери контента EN-канона.
**Why human:** Языковое качество и слог — владенческая оценка; автоматика доказала структуру (22/22 заголовков, греп-гейты).

### 4. Полезность SKILL.md для AI-ассистента (08-VALIDATION Manual-Only, поток 3)

**Test:** Дать AI-ассистенту skills/goswitch-config/SKILL.md и попросить изменить настройку (например, тумблер звука) без чтения кода.
**Expected:** Ассистент корректно правит config.yaml по справочнику (ключи/дефолты/процедуры верны).
**Why human:** Потребитель — AI-сессия; оценивает владелец.

### 5. WR-02: решение владельца по семантике a11y.apps (deferred, не silently dropped)

**Test:** Прочитать todo .planning/todos/pending/a11y-apps-matching-semantics.md и выбрать: (а) честные доки — переформулировать a11y.apps как «выключатель магии + декларация целевых приложений» + mise run skillgen-regen, или (б) пер-апп механизм — отдельная фаза.
**Expected:** Доки и код согласованы после решения; сейчас docs/CONFIG.md:73 и SKILL.md описывают матчинг, которого в рантайме нет.
**Why human:** Меняет пользовательский контракт (D-8-8: дефект/расхождение решает владелец); сознательно выведено из scope fixer'а.

### 6. Человеческий проход по CR-01-фиксу (advisory — review flag)

**Test:** Просмотреть диф 1927bf9 (split seriesMu/mu, post-episode re-check) на соответствие comment-инвариантам.
**Expected:** Разделение локов корректно; regression-корпус пинит контракт (оба теста зелёные под -race).
**Why human:** Ревью пометило фикс «requires human verification» (concurrency-semantics change); автоматическое доказательство — 2 regression-теста — уже зелёное.

### Gaps Summary

Блокирующих гэпов нет. Все 44 plan-frontmatter must-have истины верифицированы в дереве: код присутствует, содержателен, подключён (фолд→sink→reconciler→gsettings/belt; install→snapshot→restore; skillgen→SKILL.md→golden-гейт), полный `mise run ci` зелёный (exit 0). Два goal-level поведения (живой эффект магии; живой uninstall-revert) присутствуют и подключены, но не упражнены автоматикой — они сознательно спроектированы Manual-Only через seam'ы и уходят в UAT (behavior_unverified: 2). WR-02 — единственное известное док↔код расхождение — честно отложено владельцу через todo (коммит bcd2670), не замолчено; REVIEW-FIX-CR-01/WR-01 верифицированы в дереве с regression-тестами. Pre-existing флаки тестов (ctlsvc/appid, live-демон на session bus) задокументированы в deferred-items.md фазы и не относятся к дифу фазы (повторно не воспроизвелись в верификационном прогоне).

---

_Verified: 2026-10-05T23:55:00Z_
_Verifier: Claude (gsd-verifier)_
