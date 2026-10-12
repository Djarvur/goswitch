---
phase: 10-term-slovo
plan: 04
subsystem: session-terminal-write
tags: [terminal-branch, chunked-level2, focus-recheck, verify-after, no-rung, d24, contract-d7, tdd]

# Dependency graph
requires:
  - phase: 10-term-slovo plan 03
    provides: classifyTerminal — классификационный шов терминального класса (роль 60 / correction.terminals / verify-dead), вызываемый в точке маршрутизации
  - phase: 10-term-slovo plan 02
    provides: ConvertibleStretch/StretchTail геометрия пробега (таргет терминальной ветки) + internal/clipboard.Primary (не потреблён — вердикт OQ 2 «contract»)
  - phase: 02-word-correction
    provides: correctionRange/BuildPlan/D-27 cap, executeLevel2, verify-after машина (settleAfter/armAfterVerify/epoch-дисциплина)
provides:
  - Терминальная ветка записи в startRangeCorrection: маршрутизация ПО caps-развилке по classifyTerminal (вердикт OQ 1 «unified» — весь терминальный класс), мимо pending-машины и апп-верификации ADR-004
  - Чанкированный executeLevel2: чанки ForwardKeyEvent(BackSpace) ≤8 (terminalChunkKeys), геометрия level-2 форсируется маскировкой анонсированного бита, BuildPlan не тронут (D-13 хвост едёт как набран)
  - Перечек контекста между чанками: ctxEpoch-поколение, гасимое HandleLifecycle FocusOut/Reset; мьютекс отпускается на один scheduling-квант на границе чанка — смена поколения = INFO-рубеж «correction outcome=context-changed», без commit и без достройки
  - Бесплатный аборт: empty-buffer / convert-failed / no-engine / D-27 cap — ноль forwards, до первого разрушающего действия (критерий 4 ROADMAP)
  - Терминальный verify-after без ранга клипборда: armAfterVerifyTerminal (раунд без paste-нагрузки), settleAfter noRung-ветка — mismatch только логирует и останавливает; Ctrl+V в шелл невозможен ни при каком конфиге (ADR-008 D4 по построению)
  - D-24-дискриминатор changed=true|false на всех трёх done-INFO (grep-формы стенда «msg»:«correction»,«outcome»:«done» сохранены — атрибут, не переименование)
  - Контрактная граница D7 в docs/CONFIG.md (рядом с correction.terminals): буфер-доверие только для потока «набор сразу в позицию курсора»
affects: [10-05 (живые ряды VTE/wezterm гоняют эту ветку; ops-формы 8+8+4 и context-changed — оракулы), 10-06 (амендмент ADR-008 закрепляет OQ 1 unified + OQ 2 contract D7)]

# Actuals (#2632) — pairs with the plan's `estimate` (55000) to calibrate future estimates.
actuals:
  tokens: 15955        # chars/4 over the realized diff 1a7d636..HEAD (63821 chars, 5 files)
  tasks: 3
  commits: 4           # MEASURED: git rev-list --count gsd-plan-head-before-10-04..HEAD (2 RED + 2 GREEN)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Чанк-граница с отпуском мьютекса: реальный lifecycle-колбэк сериализуется на том же мьютексе и без окна был бы наблюдаем только ПОСЛЕ бурста; executeLevel2 отпускает мьютекс на runtime.Gosched() между чанками, тег поколения (ctxEpoch, verifyEpoch-дисциплина) ловит смену контекста на этой или следующей границе"
    - "Запрет по построению, не конфигом: терминальный verify-after-раунд не несёт paste-нагрузки (noRung) — settleAfter физически не может армить ранг, юнит пиннит «ни при каком конфиге»"

key-files:
  created: []
  modified:
    - "internal/session/actor.go (терминальная маршрутизация в startRangeClassification перед caps-развилкой, чанкированный executeLevel2 + chunkBoundaryMoved, ctxEpoch в HandleLifecycle, armAfterVerifyTerminal + pendingAfter.noRung + settleAfter noRung-ветка, changed-дискриминатор на трёх done-INFO)"
    - "internal/session/actor_test.go (терминальный корпус Task 2: 6 тестов; verify-after корпус Task 3: 3 теста; миграция второй половины TestTerminalClass_VerifyDeadLearnedOnVerifyTimeout)"
    - "docs/CONFIG.md (контрактная граница D7 рядом с correction.terminals)"

key-decisions:
  - "ВЕРДИКТ ВЛАДЕЛЬЦА OQ 1 = «unified» (2026-10-12): единый терминальный канал level-2 для ОБОИХ терминалов (VTE и wezterm) — буква TERM-02 и черновика ADR-008 D3/D4, один кодовый путь, единообразная тест-матрица. Позволено: жертвууется работающий level-1 путь VTE с апп-верификацией — середина строки VTE защищается на уровне буфера, не поля. Роль 60/паттерн/verify-dead ведут в одну ветку (TestTerminalBranch_VTERoleUnifiedChunks пиннит)"
  - "ВЕРДИКТ ВЛАДЕЛЬЦА OQ 2 = «contract» (2026-10-12): контрактная граница БЕЗ механизма (D7) — буфер-доверие только для флоу «набор сразу в позицию курсора»; сценарий «стрелки → набор» в wezterm остаётся вне защиты, честно задокументированная граница; закрепляется амендментом ADR-008 в плане 10-06. Никакого ядовитого флага, никакого primary-read механизма в ветке записи (internal/clipboard.Primary остаётся примитивом без потребителя)"
  - "Чанк-граница отпускает мьютекс на один квант (chunkBoundaryMoved): буквальное «бурст под мьютексом» делало бы остановку ненаблюдаемой — реальный FocusOut блокируется на том же мьютексе и садится только после бурста; окно границы и есть механизм критерия 4. Незамеченный в первом окне колбэк ловится на следующей границе (бурст не достраивается в любом случае)"
  - "Легаси no-caps путь едет через тот же чанкированный executeLevel2, но НЕ армит verify-after (terminal=false) — ADR-004-деградация байт-как-сегодня; терминальный класс армит всегда (безбитный терминал соберёт тихий таймаут)"
  - "Аборт по фокусу вызывает settleCombo: раунд завершился (остановкой) — D-36-флип комбо-жеста стреляет, comboPending не зависает"

patterns-established:
  - "TDD-ред с TAP-свидетельством: механический 1:1 перевод go test -v в evidence-JSON, check tdd-red-evidence → RED_EVIDENCE_OK перед каждым GREEN (прецедент 10-02/10-03 продолжен)"
  - "Миграция ожидания при переносе семантики с комментарием о миграции (прецедент 10-03): вторая половина теста обучения verify-dead переведена на «повторный verify-раунд недостижим под unified»"

requirements-completed: [TERM-02]  # declared by ALL six 10-xx plans — shared-ID gate keeps it open until 10-06 (ready-ids: 0/1)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Терминальная ветка маршрутизации: классификация ПО caps-развилке, весь терминальный класс (verify-dead с анонсированным битом, роль 60) идёт в чанкированный level-2 — ноль delete, ноль pending-машинных require, форма бурста 8+8+4 + один commit, GTK-путь байт-в-байт (полный корпус зелёный без правок)"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestTerminalBranch_RoutesChunkedNoPending (ops-лог: 20 forwards 8+8+4 + commit + verify-after-arm), TestTerminalBranch_VTERoleUnifiedChunks (роль 60 → та же ветка)"
        status: pass
      - kind: unit
        ref: "существующий корпус GTK — TestActor_DoubleTapCorrects/VerifyPaths и весь пакет зелёные (mise run ci): pending-машина ADR-004 не тронута"
        status: pass
    human_judgment: false
  - id: D2
    description: "Критерий 4: бесплатный аборт до первого BackSpace (empty-buffer/convert-failed/no-engine/D-27 — ноль forwards) и остановка без достройки при потере фокуса между чанками (forwards обрываются на границе чанка, commit отсутствует, INFO-рубеж context-changed)"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestTerminalBranch_FreeAbortBeforeBurst (4 ячейки), TestTerminalBranch_FocusLossBetweenChunksStops (forwardHook → HandleLifecycle FocusOut между чанками)"
        status: pass
      - kind: unit
        ref: "TestTerminalBranch_TailRidesReplay — D-13 на реплее: «ghbdtn » → «привет » через чанки, BuildPlan не тронут"
        status: pass
    human_judgment: false
  - id: D3
    description: "Терминальный verify-after без ранга: mismatch при clipboard_rung=true и живом клиенте не порождает ни одного Ctrl+V-форварда и ни одного subprocess'а; verify-dead таймаут тих (без paste, без повторов, epoch-дисциплина); честный пуш суффиксом подтверждается тихо; легаси-путь не армит ничего"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestTerminalBranch_VerifyAfterNoRungOnMismatch, TestTerminalBranch_VerifyDeadQuietTimeout, TestTerminalBranch_VerifyAfterQuietConfirm; TestActor_Level2NoCaps (легаси, ноль requires) зелёный без правок"
        status: pass
      - kind: other
        ref: "mise run ci — build+vet+golangci-lint (v2 strict, 0 issues)+test -race; go.mod/go.sum нетронуты (git diff --exit-code); 16 grep-форм e2e «msg»:«correction» substring-совместимы с changed-атрибутом"
        status: pass
    human_judgment: false
  - id: D4
    description: "D-24-дискриминатор: исполненная коррекция пишет INFO «correction outcome=done, changed=true» (юнит-пин); instant-done-ветка несёт changed=false той же формой, но поведенчески недостижима — ConvertRuns по построению никогда не возвращает changed=false (runs.go доктрина, WINDOWS-2 owner-зачёт), ветка остаётся защитной"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestTerminalBranch_RoutesChunkedNoPending (grep «outcome\":\"done\",\"changed\":true); обе frozen grep-формы матчатся substring'ом"
        status: pass
    human_judgment: true
    rationale: "Половина дискриминатора (changed=false) не имеет поведенческого теста: вход в D-24-ветку невозможен без нового шва в ConvertRuns — вне объёма плана; зафиксировано в WINDOWS ledger (kind: deviation)"
  - id: D5
    description: "Семантика середины строки ровно по вердикту OQ 2 «contract»: poison-механизм НЕ поставлен (ноль кода), primary-read НЕ подключён, граница D7 зафиксирована док-абзацем docs/CONFIG.md рядом с correction.terminals"
    requirement: TERM-02
    verification:
      - kind: other
        ref: "grep poison|Primary( internal/session/actor.go — ноль вхождений в ветке записи; docs/CONFIG.md:45-49 — контрактный абзац «набор сразу в позицию курсора»; skillgen-золотой в синхроне"
        status: pass
    human_judgment: false
  - id: D6
    description: "Flagged assumption плана: формальная классификация края TERM-02 остаётся unresolved (SPEC без ## Edge Coverage) — край реплея выражен behavior-строками Tasks 2-3 и покрыт корпусом, но классификационный вердикт — за фазовой верификацией"
    requirement: TERM-02
    verification:
      - kind: other
        ref: ".planning/phases/10-term-slovo/10-04-PLAN.md — Flagged Assumptions: {\"requirement_id\":\"TERM-02\",\"category\":\"unclassified\",\"status\":\"unresolved\"}"
        status: unknown
    human_judgment: true
    rationale: "Flagged assumption плана никогда не авто-разрешается: verifier обязан классифицировать край явно, а не тихим pass"

# Metrics
duration: 39min
completed: 2026-10-12
status: complete
---

# Phase 10 Plan 04: терминальная ветка записи — чанкированный level-2 + verify-after без ранга Summary

**Терминальный класс корректируется собственной веткой записи (вердикт OQ 1 «unified» — оба терминала в один канал): чанки BackSpace ≤8 с перечеком контекста между чанками через отпускаемый на квант мьютекс, бесплатный аборт до первого разрушающего действия, verify-after без paste-нагрузки — mismatch в терминале может только логировать и останавливать, Ctrl+V в шелл невозможен по построению при любом конфиге; середина строки — контрактная граница D7 по вердикту OQ 2 «contract».**

## Performance

- **Duration:** 39 min
- **Started:** 2026-10-12T17:23:08Z
- **Completed:** 2026-10-12T18:02:14Z
- **Tasks:** 3/3 (Task 1 — checkpoint:decision, разрешён вердиктами владельца без коммита; Tasks 2-3 — tracer/auto, каждый RED→GREEN)
- **Files modified:** 3 production (actor.go, actor_test.go, docs/CONFIG.md) + 2 RED-evidence JSON в .planning

## Accomplishments

- Вердикты владельца из checkpoint'а применены дословно и процитированы в key-decisions: **OQ 1 = oq1-unified** — единый терминальный канал level-2 для VTE и wezterm (роль 60, паттерн correction.terminals и verify-dead ведут в одну ветку через classifyTerminal — тот же шов, что у вето D-55, разъезд невозможен по построению); **OQ 2 = oq2-contract** — граница D7 без механизма: буфер-доверие только для «набор сразу в позицию курсора», poison-флаг и primary-read не ставились (Primary остаётся примитивом без потребителя — лег 10-06 снят).
- Критерий 4 ROADMAP исполнен: маршрутизация ПО caps-развилке (терминал никогда не входит в pending-машину — ноль Require до бурста); чанки ≤8 с перечеком поколения контекста (ctxEpoch гасится существующим HandleLifecycle); потеря фокуса между чанками останавливает раунд INFO-рубежом без commit и достройки; все четыре отказа (empty-buffer, convert-failed, no-engine, D-27) происходят до первого BackSpace.
- Ключевой механизм перечека: реальный FocusOut-колбэк сериализуется на том же мьютексе, что и бурст — без окна он был бы наблюдаем только ПОСЛЕ всего бурста (дефект, который чинит критерий 4). executeLevel2 отпускает мьютекс на один scheduling-квант на каждой границе чанка (chunkBoundaryMoved); колбэк, пропущенный в одном окне, ловится на следующей — достройка в любом случае невозможна.
- Терминальный verify-after: та же эпоха/WR-05-дебаунс/тихий таймаут, но раунд не несёт paste-нагрузки — settleAfter's noRung-ветка возвращает до вычисления условия ранга. Юнит пиннит запрет при clipboard_rung=true и живом wl-clipboard-клиенте: ноль Ctrl+V-форвардов, ноль subprocess'ов.
- D-24-дискриминатор changed=true|false добавлен атрибутом ко всем трём done-INFO — 16 grep-форм стенда («msg»:«correction»,«outcome»:«done») остались зелёными substring'ом.
- Легаси no-caps путь и GTK pending-машина байт-в-байт: весь существующий корпус зелёный без правок ожиданий (единственная миграция — вторая половина теста обучения verify-dead, см. Deviations).
- Новых модулей ноль; go.mod/go.sum нетронуты; строгий golangci-lint v2 — 0 issues; mise run ci зелёный на обоих GREEN.

## Task Commits

Each task was committed atomically (TDD: test → feat):

1. **Task 1: checkpoint:decision — два вердикта владельца (OQ 1, OQ 2)** — разрешён владельцем 2026-10-12 (без коммита — решение; вердикты записаны в этом SUMMARY и STATE.md, читаются планами 10-05/10-06)
2. **Task 2: Трассер — маршрутизация + чанки ≤8 + перечек фокуса + бесплатный аборт + D-24** - `630c274` (test, RED: 6 целевых падений, RED_EVIDENCE_OK) + `b9464a4` (feat, GREEN)
3. **Task 3: Verify-after без ранга клипборда + семантика середины строки (OQ 2 contract)** - `8453a09` (test, RED: 3 целевых падения, RED_EVIDENCE_OK) + `193778e` (feat, GREEN: armAfterVerifyTerminal + noRung + CONFIG.md D7)

**Plan metadata:** финальный docs-коммит (SUMMARY+STATE+ROADMAP+WINDOWS)

_Note: TDD tasks have two commits each (test RED → feat GREEN); no refactor-only commits (чистка линта уехала в GREEN)_

## Files Created/Modified

- `internal/session/actor.go` — terminalChunkKeys=8; ctxEpoch-поколение (гасится HandleLifecycle FocusOut/Reset); маршрутизация в startRangeCorrection ПО caps-развилке (classifyTerminal, вердикт unified); чанкированный executeLevel2 (маскировка анонсированного бита — BuildPlan не тронут; chunkBoundaryMoved — отпуск мьютекса на квант; INFO-рубеж «correction outcome=context-changed»); armAfterVerifyTerminal + pendingAfter.noRung + settleAfter noRung-ветка; changed-дискриминатор на трёх done-INFO
- `internal/session/actor_test.go` — терминальный корпус Task 2 (6 тестов: маршрутизация/унификация/фокус-стоп/free-abort×4/легаси-чанки/D-13-хвост), verify-after корпус Task 3 (3 теста: no-rung mismatch/тихий таймаут/тихое подтверждение); миграция второй половины TestTerminalClass_VerifyDeadLearnedOnVerifyTimeout
- `docs/CONFIG.md` — контрактная граница D7: абзац рядом с correction.terminals (буфер-доверие только для «набор сразу в позицию курсора»; «стрелки → набор» вне защиты; отдельного выключателя нет)
- `.planning/phases/10-term-slovo/evidence-20261012-10-04-red-task{2,3}.json` — TDD RED-свидетельства (RED_EVIDENCE_OK для каждой задачи)

## Decisions Made

См. key-decisions. Ядро: оба вердикта владельца применены без домысливания (unified-маршрутизация пиннута отдельным тестом роли 60; contract — ноль механизмов, только док-граница); чанк-граница с отпуском мьютекса — единственный способ сделать критерий 4 наблюдаемым в реальном процессе; терминальный класс армит verify-after всегда, легаси — никогда; аборт по фокусу закрывает и comboPending (settleCombo — D-36 без зависших флипов).

## Deviations from Plan

**[Rule 1 - Bug] Вторая половина TestTerminalClass_VerifyDeadLearnedOnVerifyTimeout (10-03) кодировала до-unified семантику**
- **Found during:** Task 2, GREEN
- **Issue:** тест ждал второй verify-timeout на уже выученном контексте — под вердиктом OQ 1 «unified» выученный verify-dead контекст маршрутизирует следующий Double в терминальную ветку, verify-раунд больше никогда не армится, второй expiry недостижим (обнаружено падением существующего теста)
- **Fix:** ожидание мигрировано (verify-timeout остаётся 1, маркер класса остаётся 1) с комментарием о миграции в GREEN-коммите; дисциплина «нет перехода — нет маркера» сохранена; сценарий «повторное обучение после FocusOut/Reset» в соседних тестах 10-03 не тронут и зелёный
- **Files modified:** internal/session/actor_test.go
- **Verification:** полный корпус зелёный; recorded in .planning/WINDOWS.md (kind: deviation)
- **Commit:** b9464a4

**[Rule 3 - Blocking] Дизайн перечека: «бурст под мьютексом» в буквальной форме делает остановку ненаблюдаемой**
- **Found during:** Task 2, реализация
- **Issue:** реальный HandleLifecycle(FocusOut) сериализуется на том же мьютексе актора — при бурсте, держащем мьютекс, колбэк садится только ПОСЛЕ всего бурста (и юнит с forwardHook→HandleLifecycle взаимно блокировался бы): критерий 4 «фокус-переключение перечекается» был бы мёртвым кодом
- **Fix:** ctxEpoch-тег (verifyEpoch-дисциплина), гасимый существующей веткой HandleLifecycle; executeLevel2 отпускает мьютекс на один scheduling-квант между чанками (chunkBoundaryMoved) и сравнивает поколение; «один эмиттер-снимок на бурст» сохранён; пропущенный колбэк ловится на следующей границе
- **Files modified:** internal/session/actor.go
- **Verification:** TestTerminalBranch_FocusLossBetweenChunksStops — forwards обрываются на границе чанка, commit отсутствует, INFO-рубеж записан
- **Commit:** b9464a4

**[Rule 3 - Blocking] Строгий golangci-lint на новом коде (funlen/lll/gochecknoglobals/unparam/gofmt)**
- **Found during:** Tasks 2-3, итерационный gate
- **Issue:** 64>60 строк в таблице free-abort; 4 длинные строки (lll >120); 2 глобальные var; неиспользуемый параметр terminal в GREEN Task 2 (арминг — Task 3); gofmt-выравнивание после вставки поля
- **Fix:** ячейки no-engine/backspace-cap вынесены в хелперы; строки перенесены; константы termWord20/termWord20Converted собраны const-выражением; параметр введён только в GREEN Task 3 (минимальный GREEN каждой задачи)
- **Verification:** mise run ci зелёный, golangci-lint 0 issues
- **Commit:** b9464a4, 193778e

**Эволюция ожиданий Task 2 при GREEN Task 3 (по плану, не отклонение):** «ноль requires» тестов Task 2 отражало промежуточное состояние — Task 3 добавляет ровно один require (verify-after arm, строго после commit по ops-логу); ожидания обновлены с комментарием «Task 3 semantics», пиннитcя отсутствие PENDING-машины (require не ДО бурста).

**Total deviations:** 3 auto-fixed (1 Rule 1 — миграция устаревшего ожидания, предписанная вердиктом владельца; 2 Rule 3 — наблюдаемость перечека и дисциплина строгого линта) + 1 документированный предел (changed=false без поведенческого теста — защитная ветка, WINDOWS ledger). **Impact:** нулевое на контракт задач — критерии 1/3/4 ROADMAP исполнены, GTK и легаси пути байт-в-байт, оба вердикта реализованы ровно по одному выбранному варианту.

## Issues Encountered

- Первый прогон mise run ci после GREEN Task 2 упал без видимого FAIL-пакета (выход отличен от нуля при зелёных всех пакетах в выводе); два подряд полных прогона (test, ci) после этого зелёные — классифицировано как средовый флак, воспроизведений нет.
- Юнит формы бурста не различает чанки в линейном ops-логе (8+8+4 forwards неотличимы от слитных 20) — форма пиннится связкой: точная серия ops + фокус-стоп на границе чанка (граница наблюдаема только через остановку). Это соответствует формулировке приёмки плана («N forwards по ≤8 + один commit»).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- План 10-05 (живые ряды) получает готовую ветку и её журнальные оракулы: «correction outcome=context-changed» (фокус-стоп), «correction outcome=done,changed=true», slug'и отказов unchanged; чанки 8+8+4 — ожидаемая pty-земля для VTE-ряда; A2-вопрос потери клавиш в очереди IME решает живой ряд.
- План 10-06 (амендмент ADR-008) закрепляет: OQ 1 = unified (D3/D4 без двухклассовой записи), OQ 2 = contract (D7 — граница без механизма, текст в docs/CONFIG.md уже есть); OQ 3/primary-read лег снят — internal/clipboard.Primary остаётся примитивом без потребителя.
- Замороженные негативные ряды 10-05 пиннят вето 10-03 — терминальная ветка не меняет арминг-вето (veto стоит в autoCorrectBoundary и не зависит от ветки записи).
- Flagged assumption плана (край TERM-02 unclassified, D6 в coverage) остаётся unresolved до фазовой верификации — verifier обязан учесть строку, не тихий pass.
- Блокеров нет; TERM-02 остаётся открытой до 10-06 (shared-ID, шесть планов; ready-ids: 0/1).

---
*Phase: 10-term-slovo*
*Completed: 2026-10-12*

## TDD Gate Compliance

- RED: `test(10-04)` commits 630c274 (Task 2: 6 целевых падений), 8453a09 (Task 3: 3 целевых падения) — каждый с проверенным свидетельством RED_EVIDENCE_OK (gsd check tdd-red-evidence, TAP 1:1 из go test -v)
- GREEN: `feat(10-04)` commits b9464a4, 193778e — целевые тесты зелёные, пакет зелёный под -race, mise run ci зелёный
- REFACTOR: отдельных refactor-коммитов нет (REFACTOR опционален; чистка линта уехала в GREEN-коммиты)
- Tracer gate (Task 2): verify перезапущен end-to-end после GREEN — зелёный; расширение на Task 3 autorizовано (⚡ Tracer verified end-to-end — expanding)

## Self-Check: PASSED

SUMMARY exists; commits 630c274 / b9464a4 / 8453a09 / 193778e verified in git log; ledger recount: 4 commits from plan base 1a7d636 (gsd-plan-head-before-10-04); production file list matches plan frontmatter 2/2 (+docs/CONFIG.md по вердикту contract — предписано планом); no stubs (TODO/FIXME/placeholder-скан чист); go.mod/go.sum нетронуты.
