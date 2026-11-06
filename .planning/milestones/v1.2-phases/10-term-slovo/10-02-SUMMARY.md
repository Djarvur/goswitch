---
phase: 10-term-slovo
plan: 02
subsystem: correct-core
tags: [convertible-stretch, punctuation-fix, config-schema, terminals, primary-selection, tdd]

# Dependency graph
requires:
  - phase: 02-word-correction
    provides: Buffer/Token/Tail/ReplaceToken geometry (D-13/D-14/D-15), ConvertRuns mono/mixed semantics, BuildPlan ladder, actor startCorrection pipeline
  - phase: 06-autocorrect
    provides: autocorrect payload consuming Token() boundaries — byte-identical through this plan
  - phase: 10-term-slovo plan 01
    provides: Wave-0 stand verdict (vm-primary) — not needed by this plan (headless work), recorded for the wave
provides:
  - ConvertibleStretch/StretchTail/ReplaceStretch — the run geometry (token + adjacent table-position runes, both sides) in internal/correct
  - «ytn?» → «нет,» through the whole manual Double pipeline (unit-proven: buffer corpus + actor cached-push test, delete (-4,4), one commit)
  - correction.terminals schema key: strict decode, RE2 validation with empty-pattern trap + 64 ceiling + sentinel errors, docs/CONFIG.md row, skillgen golden region in sync
  - internal/clipboard Primary(ctx) — wl-paste --primary --no-newline over the Runner seam, empty-state/non-error and CR-03 kill-shape pinned
affects: [10-03 (classification seam consumes correction.terminals), 10-04 (terminal branch consumes stretch geometry), 10-06 (wezterm leg consumes Primary)]

# Actuals (#2632) — pairs with the plan's `estimate` (60000) to calibrate future estimates.
actuals:
  tokens: 13112        # chars/4 over the realized diff 72ffc65..HEAD (52450 chars, 13 files)
  tasks: 3
  commits: 6           # MEASURED: git rev-list --count gsd-plan-head-before-10-02..HEAD (3 RED + 3 GREEN)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "ConvertibleStretch: геометрия подачи меняется, конвертер — нет; membership = ключ в ENToRU/RUToEN без ограничения «образ — буква»; автокоррекционный payload живёт на Token()"
    - "Строгий ключ конфига одним PR: схема+валидация+sentinel-ошибки+docs/CONFIG.md+skillgen-золотой (прецедент D-53/D-54); zero-value = прежнее поведение"

key-files:
  created:
    - "internal/correct/stretch.go"
    - "internal/correct/stretch_test.go"
  modified:
    - "internal/session/actor.go (startCorrection feeds ConvertibleStretch)"
    - "internal/session/actor_test.go (TestActor_DoubleTapPunctuationStretch, additive)"
    - "internal/config/config.go (Correction.Terminals + maxCorrectionTerminals + sentinels + validate)"
    - "internal/config/config_test.go (terminals corpus, additive)"
    - "internal/clipboard/clipboard.go (flagPrimary + Primary)"
    - "internal/clipboard/clipboard_test.go (Primary corpus, additive)"
    - "docs/CONFIG.md (schema row + zero-value note)"
    - "skills/goswitch-config/SKILL.md (regenerated golden region)"

key-decisions:
  - "Дефект «ytn?»→«нет?» чинится геометрией подачи, не конвертером: ConvertibleStretch строится от activeRange в обе стороны по предикату табличной позиции (ENToRU/RUToEN key membership); runs.go не тронут — микс-семантика spec-delta 2026-10-06 байт-в-байт"
  - "Автокоррекционный payload продолжает потреблять Token() (D-13/D-14 байт-в-байт): существующие detect/buffer/actor золотые зелёные без правок ожиданий; пробег — диапазон только ручного пути слова"
  - "correction.terminals: ключ в секции correction (обслуживает и терминальный путь, и вето), zero-value = off (прецедент D-54), потолок 64 + sentinel-ошибки с полем и индексом (прецедент D-53); Defaults() ключ не включает"
  - "Primary-чтение — тестируемый примитив без потребителя (лег 10-06 по OQ 2/3): пиннутый argv [--primary, --no-newline], OFF-mutex контракт в док-комментарии (прецедент runClipboardRung)"

patterns-established:
  - "Пробежная геометрия: граница пробега останавливается ТОЛЬКО на руне без табличной позиции (пробел D-13 живёт); головная сторона («(ytn») и хвостовая («ytn?») симметричны"
  - "TDD-дисциплина строгого линта: RED-скелеты для компиляции новых API (stretch.go, Terminals field, Primary stub) едут в test-коммите, GREEN — минимальная реализация"

requirements-completed: [TERM-02]  # declared by ALL six 10-xx plans — shared-ID gate keeps it open until 10-06

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Конвертируемый пробег: «ytn?» конвертируется целиком в «нет,» сквозь конвейер correct→session; D-13 (пробел), головная сторона, цифры, микс, é-отказ, регистры спиннуты золотыми"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/correct/stretch_test.go — TestConvertibleStretch_* (8 тестов: WordWithPunctuation, SpaceStopsRun, HeadSidePunctuation, DigitsRemainMembers, MixedPunctRidesAsTyped, UnmappedRefusesWholeStretch, RegisterPreserved, ReplaceStretchCoversRun)"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go — TestActor_DoubleTapPunctuationStretch (GTK cached-push: delete (-4,4), один commit «нет,»)"
        status: pass
      - kind: other
        ref: "mise run ci — build+vet+golangci-lint (v2 strict)+test -race, вся итерация зелёная; git diff --exit-code -- go.mod go.sum"
        status: pass
    human_judgment: false
  - id: D2
    description: "Ключ correction.terminals: строгий decode, валидация (regexp/empty/потолок 64, sentinel-ошибки с полем и индексом), zero-value = прежнее поведение, docs/CONFIG.md и золотой регион SKILL.md синхронны"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/config/config_test.go — TestCorrection_TerminalsDecode, TestValidate_CorrectionTerminals(Accepted), TestDefaults_CorrectionTerminalsNil"
        status: pass
      - kind: unit
        ref: "cmd/skillgen/main_test.go — TestSkillgen_CommittedFileInSync; skillgen-regen идемпотентен (hash-object совпал на повторном прогоне); grep terminals в docs/CONFIG.md и skills/goswitch-config/SKILL.md"
        status: pass
    human_judgment: false
  - id: D3
    description: "PRIMARY-чтение: Primary(ctx) с пиннутым argv [--primary, --no-newline], empty = (nil,false,nil), signal-kill = обёрнутая ошибка, дедлайн на каждом вызове, Runner-шов"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/clipboard/clipboard_test.go — TestClipboard_Primary (argv+empty), TestClipboard_PrimaryKilledNotEmpty, TestClipboard_PrimaryDeadline; пакет зелёный под -race"
        status: pass
    human_judgment: false
  - id: D4
    description: "Край слова TERM-02: формальная классификация края (SPEC без ## Edge Coverage) остаётся unresolved — план выражает край задачей 1 (8 золотых строк пробега), но классификационный вердикт — за фазовой верификацией"
    requirement: TERM-02
    verification:
      - kind: other
        ref: ".planning/phases/10-term-slovo/10-02-PLAN.md — Flagged Assumptions: {\"requirement_id\":\"TERM-02\",\"category\":\"unclassified\",\"status\":\"unresolved\"}"
        status: unknown
    human_judgment: true
    rationale: "Flagged assumption плана никогда не авто-разрешается: verifier обязан классифицировать край слова явно, а не тихим pass"

# Metrics
duration: 28min
completed: 2026-10-12
status: complete
---

# Phase 10 Plan 02: correct-ядро — конвертируемый пробег + терминальная схема + PRIMARY Summary

**«ytn?» конвертируется целиком в «нет,» сквозь весь конвейер (пробежная геометрия от activeRange по табличной позиции, конвертер и автокоррекционный payload не тронуты), ключ correction.terminals в строгой схеме с docs+golden, PRIMARY-чтение готово как тестируемый примитив.**

## Performance

- **Duration:** 28 min
- **Started:** 2026-10-12T14:27:19Z
- **Completed:** 2026-10-12T14:54:58Z (metadata commit ~14:56)
- **Tasks:** 3/3
- **Files modified:** 13 (10 production по списку плана + 3 RED-evidence JSON в .planning)

## Accomplishments

- Дефект владельца 2026-10-12 («ytn?» → «нет?» — пунктуация ехала в tail «как набрана» во ВСЕХ приложениях) закрыт в ядре: ConvertibleStretch собирает монораскладочный пробег (токен + примыкающие руны с табличной позицией, обе стороны), startCorrection скармливает его конвейеру — юнит-доказано на двух уровнях (буферный корпус и актор с фейковым sink: delete (-4,4), один commit «нет,»).
- Инварианты спиннуты золотыми: D-13 (пробел без позиции останавливает пробег, «ghbdtn » → «привет » байт-в-байт), головная сторона («(ytn» → «(нет»), цифры (tokenCapable-наследие, в т.ч. сквозь продлённую границу «ytn?2026»), микс (пунктуация едет как набрана — runs.go не тронут), é-отказ всего пробега без частичного вывода, регистры («YTN?» → «НЕТ,»), toggle через ReplaceStretch.
- correction.terminals: строгий KnownFields-decode, валидация с sentinel-ошибками (битый regexp / пустой-бланк паттерн / потолок 65-й записи — каждая ошибка называет поле и индекс), zero-value = прежнее поведение (D-54), docs/CONFIG.md (строка схемы + смысловая заметка) и золотой регион SKILL.md синхронны, skillgen-regen идемпотентен.
- Primary-чтение: wl-paste --primary --no-newline с пиннутым argv, empty-selection = состояние а не ошибка, signal-kill = обёрнутая ошибка (CR-03), дедлайн на каждом вызове, Runner-шов — лег для wezterm-ряда 10-06.

## Task Commits

Each task was committed atomically (TDD: test → feat):

1. **Task 1: Трассер — конвертируемый пробег сквозь correct→session** - `93eab16` (test, RED: 7 целевых тестов на живых ассертах) + `50761f3` (feat, GREEN)
2. **Task 2: Конфиг терминального класса — correction.terminals** - `3a40806` (test, RED) + `ea79e4d` (feat, GREEN: схема+валидация+docs+золотой)
3. **Task 3: PRIMARY-чтение в internal/clipboard** - `bbd4af7` (test, RED) + `0411f91` (feat, GREEN)

**Plan metadata:** финальный docs-коммит (SUMMARY+STATE+ROADMAP+REQUIREMENTS)

## Files Created/Modified

- `internal/correct/stretch.go` (new) — tablePosition, stretchRange (двусторонний обход от activeRange), ConvertibleStretch/StretchTail/ReplaceStretch
- `internal/correct/stretch_test.go` (new) — золотой корпус пробега: 8 тестов behavior-блока + toggle
- `internal/session/actor.go` — startCorrection потребляет пробег (token/tail/replace), док-комментарий
- `internal/session/actor_test.go` — TestActor_DoubleTapPunctuationStretch (только добавление)
- `internal/config/config.go` — Correction.Terminals, maxCorrectionTerminals, 3 sentinel-ошибки, validate-цикл
- `internal/config/config_test.go` — decode/validate/defaults корпус terminals (только добавления)
- `internal/clipboard/clipboard.go` — flagPrimary, Primary(ctx)
- `internal/clipboard/clipboard_test.go` — Primary корпус (argv/empty/kill/deadline)
- `docs/CONFIG.md` — строка схемы correction.terminals + zero-value заметка
- `skills/goswitch-config/SKILL.md` — реген золотого региона из docs/CONFIG.md
- `.planning/phases/10-term-slovo/evidence-20261012-10-02-red{,-task2,-task3}.json` — TDD RED-свидетельства (RED_EVIDENCE_OK для каждой задачи)

## Decisions Made

See key-decisions. Core: геометрия подачи вместо конвертера; Token() остаётся payload'ом автокоррекции; zero-value терминального ключа = off; Primary — примитив без потребителя.

## Deviations from Plan

**[Rule 1 - Bug] Золотое ожидание микс-кейса в новом корпусе было неверным (кириллическая «а» вместо латинской)**
- **Found during:** Task 1, GREEN
- **Issue:** константа stretchMixOut содержала «паи?» с кириллической «а»; RUToEN['ф'] = 'a' — образ латинский, реализация вернулась «пaи?»
- **Fix:** константа теста исправлена на таблично-верный образ (комментарий называет «a» латинским образом «ф»); реализация не менялась
- **Files modified:** internal/correct/stretch_test.go
- **Verification:** корпус зелёный; ConvertRuns смешанного пробега поинтен по spec-delta 2026-10-06
- **Commit:** 50761f3

**[Rule 3 - Blocking] Строгий golangci-lint (lll ≤120, funlen ≤60) на новых тестах**
- **Found during:** Tasks 1–3, итерационный gate
- **Issue:** 9 длинных строк (lll) и 2 функции сверх потолка (funlen) в новых тестовых файлах
- **Fix:** перенос аргументов t.Fatalf на следующие строки; разделение TestValidate_CorrectionTerminals на Accepted/refusal-половины; вынос Primary kill/deadline кейсов в TestClipboard_PrimaryKilledNotEmpty / TestClipboard_PrimaryDeadline
- **Verification:** mise run ci зелёный (build+vet+lint+test -race)
- **Commit:** 50761f3, ea79e4d, 0411f91

**Total deviations:** 2 auto-fixed (1 Rule 1 — ошибка ожидания в собственном новом тесте, 1 Rule 3 — дисциплина строгого линта). **Impact:** нулевое на контракт задач — обе находки внутри новых файлов плана, ни один существующий тест и ни одна существующая семантика не изменены (диффы actor_test/config_test/clipboard_test строго аддитивны).

## Issues Encountered

- Boundary-подтест «64 entries validate» сначала кормил 64 пустые строки — валидатор справедливо отказал по ловушке пустого паттерна раньше потолка; тест исправлен на валидные паттерны (реализация верна). Зафиксировано в GREEN Task 2.
- Пробный прогон TDD-ред-свидетельства требует TAP-формы: записи evidence-JSON строятся механическим 1:1 преобразованием реального вывода `go test -v` (--- PASS/FAIL → ok/not ok + счётчики), каждая строка соответствует реальному результату прогона.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- План 10-03 получает оба лега: correction.terminals (источник «сконфигурированной identity» инварианта D-55) и пробежную геометрию correctionRange; потребитель ключа — классификационный шов в точке армирования.
- План 10-04 (терминальная ветка) строится на ConvertibleStretch-геометрии correctionRange без изменений механизмов.
- План 10-06 (wezterm) получает Primary-примитив; подключение — по вердиктам OQ 2/3.
- Flagged assumption плана (край TERM-02 unclassified, D4 в coverage) остаётся unresolved до фазовой верификации — verifier обязан учесть строку, не тихий pass.
- Отсутствуют блокеры; TERM-02 остаётся открытой до 10-06 (shared-ID, шесть планов).

---
*Phase: 10-term-slovo*
*Completed: 2026-10-12*

## Self-Check: PASSED

SUMMARY exists; commits 93eab16 / 50761f3 / 3a40806 / ea79e4d / bbd4af7 / 0411f91 verified in git log; ledger recount: 6 commits from plan base 72ffc65 (gsd-plan-head-before-10-02); production file list matches plan frontmatter 10/10; no stubs (scaffold-комментарии удалены в GREEN; TODO/FIXME/placeholder-скан чист).
