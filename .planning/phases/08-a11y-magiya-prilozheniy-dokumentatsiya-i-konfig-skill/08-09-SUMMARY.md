---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
plan: 09
subsystem: config
tags: [a11y, config-schema, yaml-strict-decode, pointer-bool, spec-delta, skillgen, documentation]

# Dependency graph
requires:
  - phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill (08-01..08-08)
    provides: the executed phase tree — SPEC §4 a11y revision 2026-10-05, the A11y schema section, the daemon reconciler, the install snapshot/revert, the actor a11y fold, CONFIG.md, SKILL.md + skillgen
provides:
  - SPEC §4 dated revision 2026-10-06 (add-only): the owner's verbatim decision, a11y collapses to the single a11y.enabled key (bool, default ON), app matching annulled (D-8-5/REV), loud strict-decode migration, uninstall-revert untouched (D-8-4)
  - internal/config: A11y{Enabled *bool} + EffectiveEnabled() (nil = ON, the Sound pointer-bool precedent) as the section's ONLY method; list field, 64-ceiling, three list sentinels and A11y.validate removed; Defaults() ships ON
  - rewritten config/load/watch corpus: absent section → ON, absent key → ON, explicit false → OFF, removed a11y.apps key rejected whole (D-33 pin), unknown-key hot-reload last-good pin
  - actor fold reading EffectiveEnabled (compile bridge in T2 GREEN, corpus migrated in T3): defaults document → exactly one Apply(true), explicit false → [true false], diff gate preserved
  - internal/a11y doc comments without application-list mentions (series/belt/episodes behaviorally untouched); internal/install untouched
  - docs/CONFIG.md + skills/goswitch-config/SKILL.md in bool semantics with a migration note; skillgen region regenerated, golden boundary 21 → 20 rows
  - 08-VALIDATION.md rows 08-09 T1..T4
affects: [phase-8-verification (final), verify-work UAT, future config-section revisions]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
actuals:
  tokens: 18161          # chars/4 over the realized diff (72645 diff bytes), estimate scale
  tasks: 4
  commits: 6             # MEASURED: git rev-list --count gsd-plan-head-before-08-09..HEAD

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "pointer-bool default-ON section (Enabled *bool + EffectiveEnabled, nil = ON) — the Sound precedent reused for a second section, never a second mechanism"
    - "corpus-migration TDD: behavior delivered by the sanctioned compile bridge in the schema GREEN, the consumer corpus rewritten in the follow-up TDD task (plan-pinned structural red)"

key-files:
  created: []
  modified:
    - docs/SPEC.md (§4 revision 2026-10-06, +39/−0 add-only)
    - docs/CONFIG.md (one a11y row default true, bool examples, rewritten apply/revert section, migration note)
    - skills/goswitch-config/SKILL.md (a11y procedure in bool semantics; generated region rebuilt)
    - cmd/skillgen/main_test.go (golden boundary 21 → 20)
    - internal/config/config.go (A11y pointer-bool, list machinery removed, Defaults ON)
    - internal/config/config_test.go, load_test.go, watch_test.go (rewritten a11y corpus)
    - internal/session/actor.go (EffectiveEnabled fold + doc comments)
    - internal/session/actor_test.go (a11y-sink corpus without the list)
    - internal/a11y/a11y.go, a11y_test.go (doc comments only)
    - .planning/phases/08-.../08-VALIDATION.md (rows 08-09 T1..T4)

key-decisions:
  - "Default-ON mechanism = the Sound pointer-bool precedent (Enabled *bool + EffectiveEnabled, nil = ON); no custom UnmarshalYAML, no Defaults() overlay in Load (load.go/watch.go untouched)"
  - "Loud migration: a document carrying the removed a11y.apps key is rejected WHOLE by strict decode with the field named (D-33) — pinned by a test that passes only while the field is gone"
  - "T2 compile bridge: pushA11y switched to EffectiveEnabled inside the schema GREEN commit so production compiles; the session corpus rewrite deferred to T3 as the plan designs (full ci structurally red only between T2 and T3)"
  - "The yaml strict-decode error names `apps ... in type config.A11y` (no dotted path) — the migration pin asserts both substrings, the 07-02 legacy-apps precedent's shape"

patterns-established:
  - "Spec-delta revision AFTER execution: add-only dated block beside the earlier revision blocks, audit trail greps + numstat deletion-0 gate"
  - "Golden-boundary discipline: when the generator's data source loses a row, the corpus boundary moves in the same commit (21 → 20)"

requirements-completed: ["D-8-1/REV", "D-8-5/REV", "D-8-6/REV"]

coverage:
  - id: D1
    description: "SPEC §4 carries the dated revision 2026-10-06 in the 08-01 form: verbatim owner quote, single enabled key default ON, list/matching annulled, loud migration, uninstall-revert unchanged; all earlier revision blocks intact (add-only)"
    requirement: "D-8-1/REV"
    verification:
      - kind: other
        ref: "grep gates: 2026-10-06 + «список не нужен» + toolkit-accessibility; audit markers 2026-10-05/2026-09-27/2026-10-04/D-53; numstat 39 additions 0 deletions"
        status: pass
    human_judgment: false
  - id: D2
    description: "docs/CONFIG.md and skills/goswitch-config/SKILL.md in bool semantics: exactly one a11y row with default true, migration note, no list-key mention; generator region byte-synced, skillgen golden boundary 20"
    requirement: "D-8-1/REV"
    verification:
      - kind: unit
        ref: "tests/cmd/skillgen/main_test.go#TestSkillgen_ParseRealConfigTable (≥20 rows) + TestSkillgen_CommittedFileInSync"
        status: pass
      - kind: other
        ref: "grep gates: no a11y.apps in CONFIG.md/SKILL.md; `| `a11y.enabled` | bool | `true` |` present; exactly one `| `a11y` row; go run ./cmd/skillgen -check green"
        status: pass
    human_judgment: false
  - id: D3
    description: "Schema: A11y = single Enabled *bool with EffectiveEnabled (nil = ON) as the only method; list field/ceiling/sentinels/validate removed; Defaults() ON; loud whole-file rejection of the removed list key; load.go/watch.go untouched"
    requirement: "D-8-6/REV"
    verification:
      - kind: unit
        ref: "tests/internal/config: TestLoad_A11yAbsentSectionMeansDefaultOn, TestLoad_A11yAbsentKeyMeansDefaultOn, TestLoad_A11yExplicitFalseDisables, TestLoad_A11yRemovedListKeyRejectedWhole, TestLoad_A11yStrictDecodeUnknownKey, TestDefaults_A11yDefaultOn, TestWatch_A11yUnknownKeyKeepsLastGood (go test ./internal/config/ -race)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Actor fold delivers EffectiveEnabled: defaults document → exactly one Apply(true), explicit false → [true false], repeat-fold diff gate preserved; reconciler doc comments free of list mentions with zero behavioral edits; internal/install untouched"
    requirement: "D-8-1/REV"
    verification:
      - kind: unit
        ref: "tests/internal/session: TestActor_A11yDefaultsDocumentPushesTrue, TestActor_A11yFoldTriggersSinkOnActivation, TestActor_A11yExplicitFalsePushesFalse, TestActor_A11ySinkInstallSelfSyncs, TestActor_A11yNilSinkNoOp (go test ./internal/session/ ./internal/a11y/ -race)"
        status: pass
      - kind: other
        ref: "diff gates: A11Y-COMMENTS-ONLY (comments-only diff in internal/a11y/a11y.go), INSTALL-UNTOUCHED (no internal/install paths in plan diff)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Full green iteration on the final tree (mise run ci: build + vet + golangci-lint + test -race) and 08-VALIDATION.md rows 08-09 T1..T4 in the Per-Task Verification Map"
    requirement: "D-8-5/REV"
    verification:
      - kind: other
        ref: "command: mise run ci (exit 0, run twice on the final tree); grep `^| 08-09 T1` 08-VALIDATION.md"
        status: pass
    human_judgment: false

# Metrics
duration: 34min
completed: 2026-10-06
status: complete
---

# Phase 8 Plan 09: Дельта-ревизия владельца (a11y: bool вместо списка) Summary

**Спецификация, схема, фолд, доки и skill переведены на единственный ключ `a11y.enabled` (bool, default ON) по дословному решению владельца 2026-10-06 — список приложений удалён из контракта, миграция громкая, uninstall-revert и reconciler-контракты нетронуты.**

## Performance

- **Duration:** 34 min
- **Started:** 2026-10-06T00:59:50Z
- **Completed:** 2026-10-06T01:34:03Z
- **Tasks:** 4
- **Files modified:** 13 (11 repo files + 08-VALIDATION.md + red-evidence records)

## Accomplishments

- SPEC §4 несёт датированную ревизию 2026-10-06 в форме 08-01: дословная цитата владельца «раз настройка глобальная, то список не нужен, а нужен bool параметр, по дефолту настройка включен» (D-8-1/REV), секция с единственным ключом `a11y.enabled` (bool, default ON), сопоставление аннулировано (D-8-5/REV), громкая миграция strict decode (D-33), uninstall-revert без изменений (D-8-4); прежние блоки ревизий целы (numstat: 39 добавлений, 0 удалений)
- Схема: `A11y{Enabled *bool}` + `EffectiveEnabled()` (nil = ON) — единственный метод секции, по прецеденту Sound (07-08); поле списка, потолок 64, три сентинела и `A11y.validate` удалены; `Defaults()` шлёт секцию включённой; кастомного `UnmarshalYAML` нет, load.go/watch.go не тронуты
- Фолд актора читает `EffectiveEnabled`: defaults-документ толкает ровно один `Apply(true)`, явное `enabled: false` толкает `[true false]`, дифф-гейт сохранён; reconciler — только doc-комменты; internal/install — ноль правок
- CONFIG.md/SKILL.md в bool-семантике с миграционной заметкой; регион перегенерирован, `go run ./cmd/skillgen -check` зелёный, golden-граница корпуса 21 → 20
- Полная зелёная итерация `mise run ci` на выходе T1, T3 и T4 (каждый прогон зелёный с первой попытки — ни один из известных флейков ctlsvc/watch не воспроизвёлся)

## Task Commits

Each task was committed atomically:

1. **Task 1: SPEC ревизия 2026-10-06 + CONFIG.md + SKILL.md + golden-граница** - `d9c2909` (docs)
2. **Task 2 (RED): failing a11y default-ON corpus** - `d71bbf6` (test)
3. **Task 2 (GREEN): a11y bool schema — EffectiveEnabled, list machinery removed** - `48c5311` (feat)
4. **Task 3 (RED): a11y-sink fold corpus rewrite** - `cea26b2` (test)
5. **Task 3 (GREEN): fold truth EffectiveEnabled — reconciler doc comments** - `87c815c` (feat)
6. **Task 4: phase regression + VALIDATION rows** - `3af05f3` (docs)

_Note: Tasks 2–3 are `type="tdd"` — RED (`test`) and GREEN (`feat`) commits per task; no separate REFACTOR commit was needed (final-form doc comments were written within the GREEN commits and tests stayed green)._

## Files Created/Modified

- `docs/SPEC.md` — §4 ревизия 2026-10-06 (add-only, +39/−0)
- `docs/CONFIG.md` — одна строка `a11y.enabled` (default true), bool-примеры, переписанный раздел «применение и откат», миграционная заметка, диагностика без ключа-списка; рецепт per-app оверрайда сохранён
- `skills/goswitch-config/SKILL.md` — процедура a11y-магии в bool-семантике; сгенерированный регион пересобран из CONFIG.md
- `cmd/skillgen/main_test.go` — golden-граница корпуса 20 строк
- `internal/config/config.go` — A11y pointer-bool, EffectiveEnabled, удаление машинерии списка, Defaults ON
- `internal/config/config_test.go`, `load_test.go`, `watch_test.go` — переписанный корпус (absent → ON, явное false → OFF, прежний ключ → отказ всего файла, hot-reload last-good на неизвестном ключе)
- `internal/session/actor.go` — фолд на EffectiveEnabled + doc-комменты
- `internal/session/actor_test.go` — корпус фолда без списка (хелпер a11yDoc(enabled), новые пины)
- `internal/a11y/a11y.go`, `a11y_test.go` — doc-комменты без упоминаний списка (поведение нетронуто)
- `.planning/phases/08-.../08-VALIDATION.md` — строки 08-09 T1..T4
- `.planning/phases/08-.../red-evidence/08-09-task2-red.json`, `08-09-task3-red.json` — TDD-свидетельства

## Decisions Made

- Default-ON механика — pointer-bool по прецеденту Sound (`Enabled *bool` + `EffectiveEnabled`, nil = ON): Load никогда не накладывает Defaults (запиненный контракт пакета), поэтому оверлей через Defaults() потребовал бы правки load.go (запрещена), а custom UnmarshalYAML дублировал бы существующий ответ пакета на «ключ отсутствует → default ON». Второй механизм не введён (grep UnmarshalYAML = 0).
- Мост компиляции в GREEN задачи 2: `pushA11y` переключён на `snap.A11y.EffectiveEnabled()` в том же коммите, что удаляет поле/метод, — production-код компилируется после каждой задачи; полный `mise run ci` встаёт на T3/T4, как план и фиксирует (truth 7).
- Пин громкой миграции требует, чтобы ошибка называла ключ: yaml.v3 называет `apps ... in type config.A11y` (без dotted-пути) — пин проверяет обе подстроки (форма прецедента 07-02 `TestLoad_StrictDecodeRejectsLegacyAppsKey`).
- RED-заготовка задачи 2 — форма 08-02 (RED-stub): поле переключено на `*bool`, `Active()` механически адаптирован, stub `EffectiveEnabled()` держал прежнюю семантику Active(), `Defaults()` оставался OFF — пять целевых тестов упали на запланированных утверждениях.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Пин миграции ссылался на несуществующую форму ошибки yaml**
- **Found during:** Task 2 (GREEN)
- **Issue:** `TestLoad_A11yRemovedListKeyRejectedWhole` ожидал в ошибке literal `a11y.apps`; yaml.v3 KnownFields называет поле как `field apps not found in type config.A11y` — dotted-путь не входит в сообщение
- **Fix:** утверждение переведено на проверку обеих подстрок `apps` и `config.A11y` (форма прецедента 07-02)
- **Files modified:** internal/config/load_test.go
- **Verification:** корпус зелёный под -race; пин по-прежнему проходит только при отсутствии поля
- **Committed in:** 48c5311 (part of task commit)

**2. [Rule 3 - Blocking] Три строки корпуса превышали потолок lll (120 символов)**
- **Found during:** Task 2 (GREEN, golangci-lint run ./internal/config/...)
- **Issue:** свеженаписанные сообщения тестов длиннее лимита .golangci.yml
- **Fix:** переносы строк в трёх t.Errorf/t.Fatalf
- **Files modified:** internal/config/config_test.go, load_test.go, watch_test.go
- **Verification:** golangci-lint run ./internal/config/... — 0 issues
- **Committed in:** 48c5311 (part of task commit)

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking lint). **Impact on plan:** оба исправления — внутризадачные доводки до зелёной итерации; объём и границы плана не изменены.

## TDD Gate Compliance

- **Task 2:** полный RED→GREEN. RED: 5 целевых тестов падают на запланированных утверждениях (92 строки сырья /tmp/08-09-t2-red-raw.txt, TAP-транскрипция в red-evidence/08-09-task2-red.json) — `gsd_run check tdd-red-evidence` вернул **RED_EVIDENCE_OK** (target_test_failed). GREEN: корпус зелёный. Коммиты `test(08-09)` d71bbf6 → `feat(08-09)` 48c5311.
- **Task 3:** план закрепляет структурный RED: legacy-корпус 08-05 не компилируется против новой схемы (5 build-ошибок, /tmp/08-09-t3-red-raw.txt, red-evidence/08-09-task3-red.json), а поведение фолда уже доставлено мостом компиляции задачи 2 — therefore assertion-failing RED для миграции корпуса не существует by design (behavior list плана прямо называет переписанные кейсы «зелёными после переписывания хелпера»). Честная фиксация: `gsd_run check tdd-red-evidence` на record задачи 3 вернул **INVALID_RED / zero_tests_discovered** (build-ред не открывает тестов) — свидетельство сохранено как audit trail, GREEN не «авторизован» им, а следует из плана (truth 7 + action задачи 3). Коммиты `test(08-09)` cea26b2 → `feat(08-09)` 87c815c.
- REFACTOR-коммитов нет: doc-комменты писались в финальной форме внутри GREEN, тесты оставались зелёными (правило tdd.md «commit only if changes made»).

## Issues Encountered

- Порядок внутри verify-цепочки задачи 1: `TestSkillgen_CommittedFileInSync` сравнивает закоммиченный SKILL.md со свежим рендером, поэтому `mise run skillgen-regen` выполнен до цепочки (regen идемпотентен, цепочка затем зелёная целиком: тесты → regen → -check → SKILL-REGENT-OK). Это порядок применения правки, а не отступление от контракта.
- Известные флейки из deferred-items.md (ctlsvc `TestRun_OnConnHookCalledOnce`, config watch debounce) не воспроизвелись: все четыре прогона `mise run ci` зелёные с первой попытки.

## Known Stubs

None — ни один файл плана не содержит заглушек: все данные wired (схема → фолд → reconciler), доки/регенерированный регион синхронны со схемой.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- План закрывает WR-02 (матчинга нет — его и не будет) и todo a11y-apps-matching-semantics; фаза 8 полностью исполнена (08-01..08-09 с SUMMARY)
- Готово к `/gsd:verify-work 8`: дерево зелёное (`mise run ci`), skill-регион байт-синхронен, аудит-трейл SPEC цел
- Миграционная заметка для владельцев старых конфигов — в docs/CONFIG.md (раздел «Секция a11y»); SPEC сохраняет прежний ключ-список только в исторических блоках (аудит-трейл)

---
*Phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill*
*Completed: 2026-10-06*

## Self-Check: PASSED

- All 9 key files verified on disk (`[ -f ]`); all 7 commits (d9c2909, d71bbf6, 48c5311, cea26b2, 87c815c, 3af05f3, 1e1cc6b) present in git log; working tree clean of tracked modifications after the SUMMARY commit
- Commit count re-measured from the ledger: `git rev-list --count gsd-plan-head-before-08-09..HEAD` = 6 task commits + 1 SUMMARY commit
- All task acceptance criteria re-run before each commit; final plan gates (`mise run ci`, `go run ./cmd/skillgen -check`, FINAL-REV-OK, AUDIT-TRAIL-STILL-OK) green on the committed tree
