---
phase: 10-term-slovo
plan: 03
subsystem: session-classification
tags: [terminal-class, verify-dead, arming-veto, d55, classify-seam, tdd]

# Dependency graph
requires:
  - phase: 10-term-slovo plan 02
    provides: Correction.Terminals schema (RE2-validated, ceiling 64, zero-value=off) — the classification seam's second source
  - phase: 10-term-slovo plan 01
    provides: vm-primary venue verdict — not needed by this plan (headless unit work), recorded for the wave
  - phase: 06-autocorrect
    provides: autoCorrectBoundary/armAutoCorrect/autoConfirm machinery, recordACAbstain closed-slug discipline, RoleSource/UseRole seam
provides:
  - classifyTerminal — единый шов терминального класса контекста из трёх источников (живая AT-SPI роль 60, паттерны correction.terminals, поведенческий класс verify-dead), один шов на обоих потребителей (вето сейчас, конвейер записи 10-04)
  - verify-dead жизненный цикл: обучение в VerifyExpiry (истёкший verify-бюджет при живом контексте), снятие в HandleLifecycle FocusOut/Reset, один source-neutral INFO-маркер на переход (A5, D-20/D-21)
  - terminal-veto в закрытом словаре воздержаний + вето-ветка в autoCorrectBoundary после blocklist-сегмента и ДО detect.Check (инвариант D-55 дословно; fail-closed при невидимой identity; confirm-стадия — второй пояс без изменений)
  - detectorFunc/UseDetector — спай-шов детектора границы (ноль вызовов при терминальном классе пиннут тестом)
affects: [10-04 (terminal write pipeline consumes classifyTerminal), 10-05 (frozen negative rows pin the veto live), 10-06 (wezterm leg rides the verify-dead class)]

# Actuals (#2632) — pairs with the plan's `estimate` (45000) to calibrate future estimates.
actuals:
  tokens: 15109        # chars/4 over the realized diff 2c88cc8..HEAD (60436 chars, 4 files)
  tasks: 2
  commits: 4           # MEASURED: git rev-list --count gsd-plan-head-before-10-03..HEAD (2 RED + 2 GREEN)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Один шов на два потребителя: classifyTerminal — булев предикат контекста; источники cheap-to-expensive (флаг → regex по кэшу → живой GetRole под autoRoleTimeout); план 10-04 вызывает тот же метод"
    - "Поведенческий класс поверх нечестных сигналов: verify-dead учится на факте (verify-бюджет истёк при живом контексте), не на caps-битах (VTE лгал 0x29, ночник сменил подпись); переходы класса — один INFO-маркер без содержимого"

key-files:
  created: []
  modified:
    - "internal/session/actor.go (verifyDead + setVerifyDead, VerifyExpiry/HandleLifecycle wiring, Options.Terminals + refreshACTerminals + compileACPatterns, classifyTerminal/matchACTerminal, terminal-veto ветка в autoCorrectBoundary, detectorFunc/UseDetector/detectVerdict, applySnapshotBindings extract)"
    - "internal/session/actor_test.go (terminal-class corpus 6 тестов, veto corpus 5 тестов, fakeDetector, teachVerifyDead, silence-matrix role-60 ячейка мигрирована на арминг-ожидание)"

key-decisions:
  - "Вето живёт в точке армирования (autoCorrectBoundary, после blocklist, до detect.Check) — дословное исполнение инварианта D-55; blocklist-сегмент остаётся первым (known app-blocked приоритетнее), confirm-стадия — второй пояс без правок"
  - "Роль 60 запрашивается в момент решения под мьютексом с дедлайном autoRoleTimeout — трактовка T-10-03-03: дедлайн и есть гарантия «не блокирующих вызовов» (клин a11y стоит тот же бюджет, что уже платит confirm); error роли = «не терминал» на границе — role-unknown дисциплина confirm остаётся авторитетом"
  - "verify-dead ставится в VerifyExpiry ПОСЛЕ pending-nil guard (stale-таймер ничему не учит), снимается FocusOut/Reset; класс — истина контекста, не потребителя (ставится и при выключенном autocorrect)"
  - "Шов назван classifyTerminal и вызывается границей напрямую (identity резолвится внутри — двойной acFocusedApp это мьютекс-кэш-читка, поведенчески невидим); отдельная At-форма убрана — неоплодотворённый unexported метод умирает на строгом unused-линте"
  - "Спай-шов детектора (план утверждал «существующий seam» — его не было): detectorFunc named type + UseDetector installer, nil = настоящий detect.Check; скелет по прецеденту 10-02 едет в RED-коммите"

patterns-established:
  - "TDD-ред с TAP-свидетельством: механический 1:1 перевод go test -v в evidence-JSON (ok/not ok + счётчики), check tdd-red-evidence → RED_EVIDENCE_OK перед каждым GREEN"
  - "Миграция ожидания при переносе семантики: старая ячейка матрицы правится в RED-коммите с комментарием о миграции (роль-60: role-forbidden/async → terminal-veto/sync) — намеренный RED, не тихая правка"

requirements-completed: [TERM-02]  # declared by ALL six 10-xx plans — shared-ID gate keeps it open until 10-06 (ready-ids: 0/1)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Классификационный шов classifyTerminal отвечает из трёх источников: роль 60 (живой GetRole в момент решения), паттерн correction.terminals (substring RE2 против наблюдаемой identity), verify-dead (поведенческий класс); GTK-контекст (61, без паттернов, живой verify) → false"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestAutoCorrect_TerminalVetoBeforeDetector (роль 60), TestAutoCorrect_TerminalVetoPatternMatch (паттерн при роли 61), TestAutoCorrect_TerminalVetoGTKUnchanged (GTK → детектор вызван 1 раз, конвейер стреляет)"
        status: pass
      - kind: unit
        ref: "existing corpus re-pin — TestAutoCorrect_SilenceMatrix (9 ячеек), TestAutoCorrect_IdentityUnknownNotProhibition, TestAutoConfirm_RoleTextAmbiguity зелёные: источники роли/ошибок не сдвинуты"
        status: pass
    human_judgment: false
  - id: D2
    description: "verify-dead жизненный цикл: обучение на истёкшем verify-бюджете при живом контексте, снятие FocusOut'ом И Reset'ом (два отдельных теста), stale-таймер не учит, переходы — один INFO-маркер без содержимого (D-20/D-21), класс переживает второй expiry без нового маркера"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestTerminalClass_VerifyDeadLearnedOnVerifyTimeout, TestTerminalClass_VerifyDeadClearedByFocusOut, TestTerminalClass_VerifyDeadClearedByReset, TestTerminalClass_NeverDeadContextClearsSilently, TestTerminalClass_StaleVerifyExpiryLearnsNothing, TestTerminalClass_MarkerWordFree"
        status: pass
    human_judgment: false
  - id: D3
    description: "Арминг-вето D-55: терминальный класс воздерживается ровно один раз с slug'ом terminal-veto ДО detect.Check (spy: 0 вызовов), confirm не назначается (requires не растут), fail-closed при ненаблюдаемой identity (verify-dead без identity ветоит), GTK-путь и ролевой switch confirm нетронуты"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestAutoCorrect_TerminalVetoBeforeDetector, TestAutoCorrect_TerminalVetoSingleSlug, TestAutoCorrect_TerminalVetoIdentityUnobservable; существующий корпус autoConfirm/acConfirmRefusals зелёный без правок (git diff: единственные удалённые строки тестов — 2 строки роли-60 ячейки)"
        status: pass
      - kind: other
        ref: "mise run ci — build+vet+golangci-lint (v2 strict, 0 issues)+test -race; go.mod/go.sum нетронуты (git diff --exit-code)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Flagged assumption плана: формальная классификация края TERM-02 остаётся unresolved (SPEC без ## Edge Coverage) — границы класса терминальности выражены behavior-строками Task 1 и покрыты корпусом, но классификационный вердикт — за фазовой верификацией"
    requirement: TERM-02
    verification:
      - kind: other
        ref: ".planning/phases/10-term-slovo/10-03-PLAN.md — Flagged Assumptions: {\"requirement_id\":\"TERM-02\",\"category\":\"unclassified\",\"status\":\"unresolved\"}"
        status: unknown
    human_judgment: true
    rationale: "Flagged assumption плана никогда не авто-разрешается: verifier обязан классифицировать край явно, а не тихим pass"

# Metrics
duration: 31min
completed: 2026-10-12
status: complete
---

# Phase 10 Plan 03: классификационный шов терминального контекста + арминг-вето D-55 Summary

**Один шов classifyTerminal отвечает «этот контекст — терминал?» из трёх источников (роль AT-SPI 60 / паттерны correction.terminals / поведенческий класс verify-dead), и вето автокоррекции terminal-veto стоит в точке армирования autoCorrectBoundary до детектора — инвариант D-55 исполнен дословно, fail-closed при невидимой identity (wezterm).**

## Performance

- **Duration:** 31 min
- **Started:** 2026-10-12T14:59:28Z
- **Completed:** 2026-10-12T15:30:50Z
- **Tasks:** 2/2
- **Files modified:** 2 production (actor.go, actor_test.go) + 2 RED-evidence JSON в .planning

## Accomplishments

- Инвариант D-55 (критерий 5 ROADMAP) исполнен буквально: вето автокоррекции перенесено из де-факто confirm-стадии в точку армирования — терминальный класс воздерживается один раз с закрытым slug'ом `terminal-veto` ДО `detect.Check`, детектор не запускается (spy пиннит ноль вызовов), confirm не назначается (ничего не армится), при этом blocklist-семантика D-53 и ролевой switch confirm (61/79/94 — второй пояс) байт-в-байт нетронуты.
- Классификационный шов `classifyTerminal` — одно место на обоих потребителей (вето сейчас, терминальный конвейер записи — план 10-04): три источника cheap-to-expensive — флаг verify-dead (бесплатно), скомпилированные паттерны correction.terminals против наблюдаемой identity (кэш, один compile на смену списка), живой GetRole роли 60 в момент решения под дедлайном autoRoleTimeout.
- Поведенческий класс verify-dead — единственный рабочий признак для wezterm (вне a11y-шины, бит surrounding есть, пушей нет): учится в VerifyExpiry на истёкшем бюджете при живом контексте (stale-таймер защищён pending-nil guard'ом), умирает с контекстом (FocusOut/Reset), каждый переход пишет один source-neutral INFO-маркер без содержимого (A5-страховка против мигающего класса, D-20/D-21).
- Fail-closed ядро инварианта проверено юнитом: невидимая identity (нет a11y-источника, роль ошибается) + verify-dead класс → вето всё равно стреляет.
- Новых модулей ноль; go.mod/go.sum нетронуты; строгий golangci-lint v2 — 0 issues; mise run ci зелёный на обоих GREEN.

## Task Commits

Each task was committed atomically (TDD: test → feat):

1. **Task 1: Классификационный шов — роль 60 | конфиг-паттерн | verify-dead** - `1a900ce` (test, RED: 3 целевых падения, RED_EVIDENCE_OK) + `5fc9317` (feat, GREEN: verify-dead жизненный цикл + маркер + fold correction.terminals)
2. **Task 2: Арминг-вето в autoCorrectBoundary — terminal-veto до детектора** - `5af74f1` (test, RED: 5 целевых падений + миграция ячейки матрицы, RED_EVIDENCE_OK) + `c7fcd51` (feat, GREEN: classifyTerminal + veto-ветка + detectorFunc/UseDetector)

**Plan metadata:** финальный docs-коммит (SUMMARY+STATE+ROADMAP+REQUIREMENTS)

_Note: TDD tasks have two commits each (test RED → feat GREEN); no refactor-only commits (REFACTOR gate optional, чистка уехала в GREEN)_

## Files Created/Modified

- `internal/session/actor.go` — verifyDead флаг + setVerifyDead (INFO-маркер переходов), обучение в VerifyExpiry / снятие в HandleLifecycle; Options.Terminals + refreshACTerminals + compileACPatterns (параметризованный compile blocklist/terminal-class); classifyTerminal + matchACTerminal (шов из трёх источников); terminal-veto ветка в autoCorrectBoundary (после blocklist, до детектора) + acReasonTerminalVeto + acRoleTerminal в закрытых словарях; detectorFunc/UseDetector/detectVerdict (спай-шов детектора); applySnapshotBindings extract (funlen-дисциплина)
- `internal/session/actor_test.go` — terminal-class corpus (6 тестов verify-dead жизненного цикла), veto corpus (5 тестов + fakeDetector spy + teachVerifyDead хелпер), silence-matrix role-60 ячейка мигрирована на арминг-ожидание (terminal-veto, sync)
- `.planning/phases/10-term-slovo/evidence-20261012-10-03-red-task{1,2}.json` — TDD RED-свидетельства (RED_EVIDENCE_OK для каждой задачи)

## Decisions Made

See key-decisions. Core: вето в точке армирования до детектора (D-55 дословно); запрос роли под мьютексом с дедлайном autoRoleTimeout как трактовка T-10-03-03 («не блокирующих» = дедлайн-ограниченных); verify-dead — истина контекста, не потребителя; шов одним методом (At-форма убрана ради strict unused-линта); спай-шов детектора добавлен по Rule 3 (план ссылался на несуществующий seam).

## Deviations from Plan

**[Rule 1 - Bug] Ячейка silence-matrix «role 60 terminal is forbidden» кодировала до-D55 семантику**
- **Found during:** Task 2, RED
- **Issue:** ячейка ждала async role-forbidden от confirm'а — ровно то размещение вето, которое план переносит в точку армирования (must_haves: «терминальный класс → воздержание terminal-veto, детектор не запускается, confirm не назначается»)
- **Fix:** ячейка мигрирована на арминг-ожидание (terminal-veto, sync) с комментарием о миграции в RED-коммите; роль-40 и остальные ячейки не тронуты
- **Files modified:** internal/session/actor_test.go (silenceMatrixCells)
- **Verification:** корпус зелёный; git diff по тестам = 2 удалённые строки (только эта ячейка); confirm-корпус (TestAutoConfirm_*, acConfirmRefusals) зелёный без правок
- **Commit:** 5af74f1, c7fcd51; recorded in .planning/WINDOWS.md (kind: deviation)

**[Rule 3 - Blocking] Строгий golangci-lint на новом коде (funlen/funcorder/lll/unused)**
- **Found during:** Tasks 1–2, итерационный gate
- **Issue:** applySnapshot 42 statement'а (>40) после fold'а terminals; UseDetector среди unexported методов (funcorder, 39 срабатываний); длинная сигнатура seam'а и док-строки (lll ×3); classifyTerminal без вызывающего в Task 1 (unused)
- **Fix:** extract applySnapshotBindings (пере-резолв комбины/чорда — поведение байт-в-байт); UseDetector перенесён к экспортируемым installer'ам (рядом с UseRole); detectorFunc named type + перенос сигнатур; classifyTerminalAt слит в classifyTerminal (двойная резолюция identity — мьютекс-кэш-читка, поведенчески невидима)
- **Verification:** mise exec -- golangci-lint run → 0 issues; mise run ci зелёный
- **Commit:** 5fc9317, c7fcd51

**Структурная заметка (не отклонение поведения):** функция шва classifyTerminal физически едет в GREEN Task 2, а не Task 1 — в Task 1 у неё нет вызывающего, и строгий `unused`-линт отверг неоплодотворённый unexported метод; шов приезжает вместе со своим потребителем (вето). done-критерий плана («classifyTerminal отвечает из трёх источников») выполнен в объёме плана.

**Total deviations:** 2 auto-fixed (1 Rule 1 — миграция устаревшего ожидания, предписанная самим планом; 1 Rule 3 — дисциплина строгого линта). **Impact:** нулевое на контракт задач — существующий корпус зелёный (единственное изменение ожидания — сама семантика D-55), поведение GTK-пути и confirm-стадии байт-в-байт.

## Issues Encountered

- Первый черновик fail-closed теста учил verify-dead без разделителя после обучающего раунда: буфер склеивал «ghbdtn»+«cnhjrf» в один токен, детектор честно отвечал trigram-unsure вместо армирования — тестовая форма исправлена (разделитель при ещё выключенном слое), реализация не менялась.
- План Task 2 ссылался на «существующий seam детектора» — на HEAD detect.Check вызывался напрямую; спай-шов (detectorFunc + UseDetector + detectVerdict) добавлен по Rule 3, скелет в RED-коммите по прецеденту 10-02.
- Блокированные role-двойники (release-канал) в существующих тестах теперь встречают роль-запрос и на границе: тесты остались зелёными (дедлайн autoRoleTimeout ограничивает +25ms на границу), семантика воздержаний не сдвинулась.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- План 10-04 (терминальная ветка записи) получает готовый шов: classifyTerminal вызывается в точке маршрутизации startRangeCorrection (caps-развилка) тем же методом, что и вето — разъезд вето и конвейера невозможен по построению.
- Замороженные негативные ряды 10-05 пинят вето вживую (counters-оракул: fired=0 при зелёной ручной); slug terminal-veto уже в закрытом словаре воздержаний и виден в goswitchctl status.
- План 10-06 (wezterm): verify-dead класс — механизм, на который ложится терминальная ветка для невидимого контекста; живое обучение класса уже покрыто юнитами, живой ряд подтвердит на стенде.
- Flagged assumption плана (край TERM-02 unclassified, D4 в coverage) остаётся unresolved до фазовой верификации — verifier обязан учесть строку, не тихий pass.
- Блокеров нет; TERM-02 остаётся открытой до 10-06 (shared-ID, шесть планов; ready-ids: 0/1).

---
*Phase: 10-term-slovo*
*Completed: 2026-10-12*

## TDD Gate Compliance

- RED: `test(10-03)` commits 1a900ce (Task 1), 5af74f1 (Task 2) — каждый с проверенным свидетельством RED_EVIDENCE_OK (gsd check tdd-red-evidence, TAP 1:1 из go test -v)
- GREEN: `feat(10-03)` commits 5fc9317, c7fcd51 — целевые тесты зелёные, пакет зелёный под -race
- REFACTOR: отдельных refactor-коммитов нет (REFACTOR опционален; чистка линта уехала в GREEN-коммиты)

## Self-Check: PASSED

SUMMARY exists; commits 1a900ce / 5fc9317 / 5af74f1 / c7fcd51 verified in git log; ledger recount: 4 commits from plan base 2c88cc8 (gsd-plan-head-before-10-03); production file list matches plan frontmatter 2/2; no stubs (TODO/FIXME/placeholder-скан чист); go.mod/go.sum нетронуты.
