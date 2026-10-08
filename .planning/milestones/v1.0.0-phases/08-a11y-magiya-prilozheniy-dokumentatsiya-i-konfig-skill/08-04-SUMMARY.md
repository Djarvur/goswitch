---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
plan: 04
subsystem: install
tags: [a11y, installer, snapshot, restore, install-state-json, gsettings, toolkit-accessibility, asvs-v5, argv-discipline, only-if-present, tdd]

# Dependency graph
requires:
  - phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill (план 08-01)
    provides: "SPEC §4 ревизия 2026-10-05 — контракт uninstall-revert: снапшот verbatim, restore only-if-present, никогда не «restore false»"
  - phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill (план 08-03)
    provides: "internal/a11y Reconciler — потребитель честного снапшота: канонический порядок «install до daemon» делает снапшот пред-состоянием"
  - phase: 05-integratsiya-s-gnome-indikatsiya (и ранее internal/install)
    provides: "saveState idempotent-backup (:531-536), restoreSources only-if-trusted форма, restore-до-удаления дисциплина, Runner seam"
provides:
  - "installState.ToolkitAccessibility (json \"toolkit_accessibility\") — пред-установочное значение org.gnome.desktop.interface toolkit-accessibility в install-state.json, verbatim; пустое = не захвачено (старый state или отказ чтения)"
  - "saveState читает ключ через существующий i.call после чордов; отказ — пустое поле + ОДИН WARN (a11y key value not captured), инсталл зелёный, значение никогда не фабрикуется (T-08-04-04)"
  - "(*Installer).restoreToolkitAccessibility(ctx, lines) ([]string, error) — вызов в Uninstall после restoreSwitchBinding, ДО удаления state-файла; shape-валидация: только точные \"true\"/\"false\" достигают gsettings set (ASVS V5, T-08-04-01); missing/empty → WARN + ноль set («nothing to revert», T-08-04-02, Pitfall 5); отказ set → WARN при зелёном uninstall"
  - "Корпус 8 новых тестов (снапшот: captured/true/read-failure; restore: true/false/missing/corrupt/set-failure) + recordSink WARN-двойка; TestInstall_SecondInstallKeepsOriginalBackup зелёный БЕЗ правок — idempotent-backup нетронут"
affects: [08-05 (wiring демон-reconcile поверх честного снапшота), 08-06 (доки описывают uninstall-revert), verify-work (UAT uninstall на живом столе)]

# Actuals (#2632) — pairs with the plan's estimate to calibrate future estimates.
actuals:
  tokens: 8594    # chars/4 over the realized diff (34378 chars: install.go + install_test.go + 2 red-evidence JSON)
  tasks: 2
  commits: 4      # MEASURED: git rev-list --count gsd-plan-head-before-08-04..HEAD
plan_head_before: 6977b439cdb18f9d2100109a64397fe679ed0332

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Трёхсостоянийный saved-ридер savedA11yState (value, present, trusted): absent (нет файла/непарсится/пустое поле) ≠ untrusted (не литерал) ≠ trusted — «nothing to revert» и «shape refused» сообщаются РАЗНЫМИ WARN'ами, в отличие от двухсостоянийных savedSources/savedSwitchBindings (там безопасный fallback существует, здесь — нет)"
    - "Auxiliary-member снапшот: член state-файла, чей отказ чтения деградирует в пустое поле + WARN при успехе инсталла — расширение дисциплины verbatim-снапшота случаем «нулевой критичности»"

key-files:
  created:
    - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/red-evidence/08-04-task1-red.json
    - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/red-evidence/08-04-task2-red.json
  modified:
    - internal/install/install.go
    - internal/install/install_test.go

key-decisions:
  - "a11y-чтение в saveState — после чтения чордов, до idempotent-backup раннего возврата: ранний возврат (:531-536) не тронут ни строкой, на повторном install чтение выполняется, но возврат держит первый снапшот нетронутым — TestInstall_SecondInstallKeepsOriginalBackup зелёный без правок"
  - "Снапшот вспомогателен (T-08-04-04): отказ чтения — пустое поле + один WARN с закрытой формулировкой, инсталл продолжается; никогда не фабрикуется значение — пустое поле честно говорит «не захвачено» и restore пропускает"
  - "restoreToolkitAccessibility сохраняет сигнатуру (ctx, lines) ([]string, error) c nolint:unparam (объяснение в строке) — форма restore-шага симметрична restoreSources/restoreSwitchBinding, вызывающий код обрабатывает ошибку одинаково; сам контракт деградирует каждый отказ в WARN"
  - "Точные sequence-пины TestInstall_Sequence/SingleSource и installCallCount 12→13 обновлены под новый снапшот-get — прямое механическое следствие GREEN-действия плана (новое чтение в Installe меняет зафиксированную последовательность), не ослабление защищаемого поведения"

patterns-established:
  - "Трёхсостоянийный saved-ридер для полей БЕЗ безопасного fallback: absent/untrusted/trusted с разными REPORTED-вербатами"
  - "Auxiliary-snapshot член installState: чтение может отказать без провала инсталла — пустое поле как честный «не захвачено», restore обязан это пережить"

requirements-completed: ["D-8-4"]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Снапшот verbatim при install: пред-установочное toolkit-accessibility (\"false\" и вручную-включённое \"true\") попадает в install-state.json как \"toolkit_accessibility\" без артефактов ответа gsettings"
    requirement: "D-8-4"
    verification:
      - kind: unit
        ref: "internal/install/install_test.go#TestInstall_A11ySnapshotCaptured"
        status: pass
      - kind: unit
        ref: "internal/install/install_test.go#TestInstall_A11ySnapshotTrueValue"
        status: pass
    human_judgment: false
  - id: D2
    description: "Отказ чтения деградирует: инсталл зелёный, поле пустое, один WARN (a11y key value not captured) — значение никогда не фабрикуется (T-08-04-04)"
    requirement: "D-8-4"
    verification:
      - kind: unit
        ref: "internal/install/install_test.go#TestInstall_A11ySnapshotReadFailureLeavesEmpty"
        status: pass
    human_judgment: false
  - id: D3
    description: "Idempotent-backup нетронут: повторный install не перезаписывает первый снапшот — существующий кейс зелёный БЕЗ правок, ранний возврат saveState не изменён"
    requirement: "D-8-4"
    verification:
      - kind: unit
        ref: "internal/install/install_test.go#TestInstall_SecondInstallKeepsOriginalBackup"
        status: pass
    human_judgment: false
  - id: D4
    description: "Restore валидированного значения ДО удаления state-файла: saved \"true\"/\"false\" восстанавливаются verbatim через gsettings set org.gnome.desktop.interface toolkit-accessibility <литерал>, отчёт называет значение"
    requirement: "D-8-4"
    verification:
      - kind: unit
        ref: "internal/install/install_test.go#TestUninstall_A11yRestoreTrue"
        status: pass
      - kind: unit
        ref: "internal/install/install_test.go#TestUninstall_A11yRestoreFalse"
        status: pass
    human_judgment: false
  - id: D5
    description: "Only-if-present + shape-гейт (Pitfall 5 / ASVS V5, T-08-04-01/02): старый state без поля и коррупт-значение дают НОЛЬ set-вызовов с toolkit-accessibility в argv + REPORTED skip (nothing to revert / shape refused), uninstall зелёный"
    requirement: "D-8-4"
    verification:
      - kind: unit
        ref: "internal/install/install_test.go#TestUninstall_A11yMissingFieldSkipsSilently"
        status: pass
      - kind: unit
        ref: "internal/install/install_test.go#TestUninstall_A11yCorruptValueSkips"
        status: pass
    human_judgment: false
  - id: D6
    description: "Отказ set деградирует (T-08-04-04) и зелёная итерация плана: set-отказ — один WARN + строка отчёта при зелёном uninstall; mise run ci (build+vet+lint strict+test -race ./...+tidy-diff) exit 0"
    requirement: "D-8-4"
    verification:
      - kind: unit
        ref: "internal/install/install_test.go#TestUninstall_A11yRestoreFailureReported"
        status: pass
      - kind: other
        ref: "mise run ci → CI_EXIT=0 (/tmp/08-04-ci-run2.log)"
        status: pass
    human_judgment: false
  - id: D7
    description: "Живой стол: реальный gsettings get при install пишет честное пред-значение, реальный gsettings set при uninstall возвращает ключ, старый инсталл (state без поля) переживает uninstall без сброса вручную-включённого ключа"
    verification: []
    human_judgment: true
    rationale: "Юнит-корпус доказывает контракт через Runner-seam; живой uninstall на рабочем столе владельца трогает его реальную сессию (остановка демона, восстановление sources) — предмет verify-work/UAT после волны, не атозапуска"

# Metrics
duration: 17min
completed: 2026-10-05
status: complete
---

# Phase 8 Plan 04: Installer a11y snapshot/revert Summary

**Снапшот пред-установочного toolkit-accessibility в install-state.json (verbatim, отказ чтения — пусто+WARN) и uninstall-restore only-if-present с shape-гейтом true|false — никогда «restore false» (D-8-4, Pitfall 5).**

## Performance

- **Duration:** 17 min
- **Started:** 2026-10-05T17:38:26Z
- **Completed:** 2026-10-05T17:55:35Z
- **Tasks:** 2 (4 TDD-коммита: RED→GREEN ×2)
- **Files modified:** 2 кода (+507/−18) + 2 red-evidence JSON

## Accomplishments

- **Снапшот (Task 1, трейсер):** `installState.ToolkitAccessibility` (json `"toolkit_accessibility"`) — пред-установочное значение `org.gnome.desktop.interface toolkit-accessibility`, прочитанное в `saveState` через существующий `i.call` после чордов, TrimSpace, verbatim; отказ чтения — пустое поле + один WARN (`a11y key value not captured`) при успешном инсталле (T-08-04-04) — инсталл никогда не валится, значение никогда не фабрикуется
- **Idempotent-backup нетронут:** ранний возврат по существованию state-файла не изменён ни строкой — первый снапшот священ; `TestInstall_SecondInstallKeepsOriginalBackup` зелёный без правок (снапшот-get на повторном install читается, но возврат не даёт перезаписи)
- **Restore (Task 2):** `restoreToolkitAccessibility(ctx, lines)` в `Uninstall` после `restoreSwitchBinding`, ДО удаления state-файла (та же restore-до-удаления дисциплина); `savedA11yState` — трёхсостоянийный ридер: absent (нет файла/непарсится/пустое поле — старый инсталл) → WARN «nothing to revert» + НОЛЬ set-вызовов (T-08-04-02, Pitfall 5); untrusted (не точный `"true"`/`"false"`) → WARN «shape refused» + ноль set (ASVS V5, T-08-04-01); trusted → set с единственным источником аргумента — shape-валидированным литералом state-файла
- **Argv-дисциплина:** произвольная строка из state-файла никогда не достигает `gsettings set` — corrupt-тест пинит ноль вызовов; восстановление легитимного `"false"` (владелец держал ключ выключенным) отличимо от запрещённого фабрикованного «restore false» (исключительно missing-field вердикт)
- **Auxiliary degradation:** отказ set — один WARN + строка отчёта, uninstall продолжаетcя зелёным (T-08-04-04); skip-и REPORTED, никогда не молчат
- **Полный TDD-цикл обоих задач:** RED `4d55501` (3/3 пали на planned-behavior, RED_EVIDENCE_OK) → GREEN `c83ff9d`; RED `0b2497c` (5/5 пали, RED_EVIDENCE_OK) → GREEN `bb769c6`; `mise run ci` exit 0

## Task Commits

Each task was committed atomically:

1. **Task 1 (трейсер): снапшот пред-установочного ключа** — RED `4d55501` (test), GREEN `c83ff9d` (feat); REFACTOR не потребовался
2. **Task 2: uninstall restore — only-if-present, shape-валидация** — RED `0b2497c` (test), GREEN `bb769c6` (feat; включает lint-фиксы nlreturn ×3 + nolint:unparam по прецеденту 08-02/08-03)

**Plan metadata:** следует отдельным docs-коммитом (SUMMARY + STATE + ROADMAP).

_Note: TDD-гейты: test(08-04) `4d55501` → feat(08-04) `c83ff9d`; test(08-04) `0b2497c` → feat(08-04) `bb769c6`._

## Files Created/Modified

- `internal/install/install.go` — константы `gsettingsSchemaInterface`/`gsettingsKeyA11y`/`a11yEnabled`/`a11yDisabled`; поле `ToolkitAccessibility` в installState (doc: пустое = не захвачено, Pitfall 5); `saveState` — a11y-чтение с WARN-деградацией; `restoreToolkitAccessibility` + `savedA11yState`; вызов в Uninstall до `os.Remove(statePath)`
- `internal/install/install_test.go` — корпус 8 тестов (снапшот ×3 + restore ×5) поверх fakeRunner; константы a11y (schema/key/replies/values); `recordSink`+`captureWarns` WARN-двойка (прецедент a11y/watch_test); драйверы `countA11ySets`/`plantA11yState`; sequence-пины обновлены под снапшот-get
- `red-evidence/08-04-task1-red.json`, `red-evidence/08-04-task2-red.json` — RED-свидетельства (оба RED_EVIDENCE_OK)

## Decisions Made

- **Позиция a11y-чтения в saveState:** после чтения чордов, до раннего возврата — все desktop-чтения сгруппированы (стиль файла), ранний возврат не тронут, защищённый тест зелёный без правок; на повторном install лишний get безвреден
- **Трёхсостоянийный savedA11yState:** absent ≠ untrusted — «nothing to revert» и «shape refused» сообщаются разными WARN'ами, потому что у a11y НЕТ безопасного fallback (в отличие от sources/чордов): любая подстановка была бы догадкой о настройке доступности владельца
- **Сигнатура с error-возвратом + nolint:unparam:** форма restore-шага симметрична restoreSources/restoreSwitchBinding — вызывающий код в Uninstall одинаков; контракт деградирует отказы в WARN (T-08-04-04)
- **Sequence-пины обновлены:** TestInstall_Sequence/SingleSource (exact-call пины) и installCallCount 12→13 — новое чтение легитимно меняет зафиксированную последовательность первого install; не путать с защищённым behavior-тестом idempotent-backup (не тронут)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Точные sequence-пины существующих тестов устарели после снапшот-get**
- **Found during:** Task 1 GREEN
- **Issue:** план добавляет четвёртый `gsettings get` в saveState; пины TestInstall_Sequence / TestInstall_SequenceSingleSource (exact-call последовательности) и константа installCallCount=12 становятся красными — задача не может завершиться зелёной итерацией
- **Fix:** в оба пина добавлена строка `{gsettings get org.gnome.desktop.interface toolkit-accessibility}` после чордов; installCallCount 12→13 с уточнённым комментарием
- **Files modified:** internal/install/install_test.go
- **Verification:** полный корпус зелёный под -race; TestInstall_SecondInstallKeepsOriginalBackup — зелёный БЕЗ правок (граница плана соблюдена)
- **Committed in:** c83ff9d (коммит Task 1 GREEN)

---

**Total deviations:** 1 auto-fixed (1 blocking — механическая синхронизация exact-call пинов с запланированным планом новым чтением).
**Impact on plan:** ни один behavior-кейс и ни одно правило (idempotent-backup, restore-порядок, argv-дисциплина) не изменились; обновлены только зафиксированные факты о последовательности, легитимно расширенной планом.

## TDD Gate Compliance

- Task 1: RED `test(08-04) 4d55501` предшествует GREEN `feat(08-04) c83ff9d`; RED-свидетельство `RED_EVIDENCE_OK` (3/3 целевых теста пали на planned-behavior-утверждениях: поле отсутствует ×2, WARN не зафиксирован; raw /tmp/08-04-t1-red-verbose.txt, запись red-evidence/08-04-task1-red.json). REFACTOR — без изменений, коммита нет.
- Task 2: RED `test(08-04) 0b2497c` предшествует GREEN `feat(08-04) bb769c6`; RED-свидетельство `RED_EVIDENCE_OK` (5/5 целевых restore-тестов пали: ноль set-вызовов ×2, отсутствие вердиктов в отчёте ×2, ноль попыток ×1; raw /tmp/08-04-t2-red-verbose.txt, запись red-evidence/08-04-task2-red.json). REFACTOR — без отдельного коммита (lint-фиксы вошли в GREEN по прецеденту 08-02/08-03).

## Issues Encountered

- Первый полный `go test -race ./...` (зелёная итерация Task 1) поймал известный pre-existing флак `internal/ctlsvc` `TestRun_OnConnHookCalledOnce` («another goswitchd instance already registered reply=2» — живой демон владельца на сессионной шине; задокументирован в deferred-items.md фазы, вне зоны плана). Перезапуск пакета однократно — зелёный (по заметкам плана); финальный `mise run ci` зелёный с первого прогона (exit 0, /tmp/08-04-ci-run2.log).
- Строгий lint поймал 4 замечания в новом коде Task 2 (nlreturn ×3 — blank line перед return в двухстейтментных блоках, unparam — всегда-nil error возврат): исправлены в GREEN-коммите (blank lines + nolint:unparam с обоснованием), затем `mise run lint` 0 issues.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `internal/install` самодостаточен по D-8-4: снапшот при install + честный restore при uninstall; демон (08-03) применяет ключ ПОСЛЕ инсталла — канонический порядок делает снапшот честным пред-состоянием; wiring демон-reconcile — 08-05, строки a11y.* в docs — 08-06
- Старые инсталлы (state-файл без поля) переживают uninstall без сброса ключа — миграция не требуется; поле появится при следующем install на чистом state (переустановка)
- Живой uninstall на столе владельца (реальный gsettings set, сбережение вручную-включённого ключа) — предмет UAT (coverage D7, human_judgment)
- Открытых блокеров нет

## Self-Check: PASSED

- internal/install/install.go, internal/install/install_test.go, red-evidence/08-04-task1-red.json, red-evidence/08-04-task2-red.json — FOUND (в дифе плана, 4 files changed, 507 insertions, 18 deletions)
- Commits 4d55501, c83ff9d, 0b2497c, bb769c6 — FOUND in git log; ledger gsd-plan-head-before-08-04 (6977b43) → 4 коммита (MEASURED: git rev-list --count)
- Гейты: SNAPSHOT-OK (toolkit_accessibility / gsettingsSchemaInterface / toolkit-accessibility) и RESTORE-OK (restoreToolkitAccessibility / nothing to revert) — FOUND
- TestInstall_SecondInstallKeepsOriginalBackup — PASS без правок (запуск -v на финальном дереве)
- mise run ci — exit 0 (финальный прогон 2026-10-05T17:5xZ, /tmp/08-04-ci-run2.log); mise run lint — 0 issues

---
*Phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill*
*Completed: 2026-10-05*
