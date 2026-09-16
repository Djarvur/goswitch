---
phase: 04-postavka-i-priemka
plan: 04
subsystem: testing
tags: [e2e, perf, atspi, ydotool, percentile, vmhwm, budget-gate, witness]

# Dependency graph
requires:
  - phase: 04-postavka-i-priemka (04-01, 04-02, 04-05)
    provides: e2e stand surfaces + install-cycle, combo case (03-04) injection forms, daemon INFO log contract
provides:
  - resident AT-SPI witness (focus_helper.py witness-events, event-driven object:text-changed, one line per event with arrival timestamp; witness-poll 5 ms fallback)
  - test/e2e/perf.go — percentile (sort+index pin), /proc VmHWM/VmRSS parser, budgetGate, perfWitness client (one spawn, line protocol, pid shutdown), perf case (N=40 combo repeats, t0/t1 ydotool→AT-SPI window, report + gate)
  - mise task e2e-perf; perf-report.txt artifact format (methodology block + p50/p95/p99 + VmRSS/VmHWM + budget verdict)
  - per-case watchdog override in the stand (caseSpec.watchdog — raised for perf, never removed)
  - DIAGNOSTIC VERDICT: the combo fires at press (no discrimination window) and the daemon-internal reaction is 0.2–1.8 ms
  - MEASURED RESULT: under the prescribed D-44 window the latency budget FAILS (p95 190–205 ms ≥ 50 ms) — the window is dominated by ydotool 0.1.8 per-event injection overhead (~80–125 ms/event); memory budget HOLDS (VmHWM 12.4 MB < 50 MB)
affects: [04-07 (README perf table — blocked on the owner's latency-verdict decision), phase verify-work (INST-03 disposition)]

# Actuals (#2632)
actuals:
  tokens: 7900    # chars/4 over the realized diff (8 files, +702/-12)
  tasks: 2
  commits: 3      # MEASURED: git rev-list --count 41884ce..HEAD

# Tech tracking
tech-stack:
  added: []       # zero new dependencies — Atspi.EventListener via python3-gi (installed), /proc via stdlib
  patterns:
    - "resident witness line protocol: <RFC3339 UTC> <app> <pid> <text>, timestamp stamped at observation arrival (readback latency excluded from the measured window)"
    - "per-case watchdog override (raised, never removed) for batch-measurement cases"
    - "count-based waitForNew after any in-case daemon restart — plain waitForLog would match the old daemon's records"

key-files:
  created:
    - test/e2e/perf.go
    - test/e2e/perf_test.go
  modified:
    - test/e2e/focus_helper.py
    - test/e2e/main.go
    - test/e2e/matrix.go
    - mise.toml
    - .gitignore

key-decisions:
  - "Event-driven witness pinned (A2 positive): the Task-1 live smoke delivered zenity lines well under 2 s — witness-events is the run mode; witness-poll (named 5 ms quantum) is the documented fallback"
  - "Combo semantics pinned live (assumption flag resolved): Shift+RightCtrl fires at the Control_R press with the correction completing 0.2–1.8 ms later (cached surrounding-push path) — no discrimination window, no verify_wait stall; budget semantics untouched"
  - "Fresh zenity surface per repeat: the closing Enter doubles as the ADR-004 buffer reset, so no repeat inherits the previous one's buffer state (a persistent surface + ctrl+a clear provably corrupts the buffer into a mixed-token correction)"
  - "Prod-form daemon for the acceptance numbers: no -config (built-in defaults), no -debug (T-04-04-03); INFO records (component registered/focus_in/combo/correction/mode) suffice for gating"
  - "The budget gate verdict stands as a REAL result (per plan prohibition: the run must not pass silently over budget): p95 190–205 ms with the pinned ydotool injector; escalation to the owner rather than a silent methodology change"

patterns-established:
  - "Perf report format: methodology block (witness mechanism + quantum, window, N, surface, daemon form) → samples → p50/p95/p99 (ms, 1 decimal) → vmrss_start_kb/vmrss_end_kb → vm_hwm_kb → budget verdict; field CONTENT never printed (T-04-04-01)"
  - "Count-based activation/flip gates after in-case daemon restarts (activateGoswitchFresh, flip-back enBase+1)"

requirements-completed: [INST-03]  # copied verbatim from the plan frontmatter; SEE coverage D4 + Issues: the LATENCY half of INST-03 is measured-and-FAILED under the prescribed window — deliberately NOT marked complete in REQUIREMENTS.md

coverage:
  - id: D1
    description: "Pure perf mathematics pinned: percentile exact ranks (formula s[int(float64(len(s)-1)*p)]), /proc VmHWM/VmRSS parser with the named missing-line error, budget gate on both boundaries"
    requirement: INST-03
    verification:
      - kind: unit
        ref: "test/e2e/perf_test.go#TestPercentile_Exact"
        status: pass
      - kind: unit
        ref: "test/e2e/perf_test.go#TestPercentile_SingleSample"
        status: pass
      - kind: unit
        ref: "test/e2e/perf_test.go#TestReadProcStatus"
        status: pass
      - kind: unit
        ref: "test/e2e/perf_test.go#TestPerfBudgetGate"
        status: pass
    human_judgment: false
  - id: D2
    description: "Resident event witness in focus_helper.py (witness-events object:text-changed, one line per event with arrival timestamp) + resident poller fallback (witness-poll, named 5 ms quantum); exit codes and /usr/bin/python3 conventions preserved"
    requirement: INST-03
    verification:
      - kind: manual_procedural
        ref: "live smoke 2026-09-16: spawn witness-events → zenity → ydotool 'a' → line '2026-09-16T13:24:24.781466+00:00 zenity 1660676 a' well under 2 s; shutdown by pid"
        status: pass
      - kind: manual_procedural
        ref: "live smoke 2026-09-16: witness-poll <pid> → initial-state line + change line on typing, same protocol"
        status: pass
      - kind: unit
        ref: "mise run ci (build + vet + golangci-lint + go test -race) green on every commit"
        status: pass
    human_judgment: false
  - id: D3
    description: "Perf case end to end: prod-form daemon, fresh zenity per repeat with ADR-004 reset, quiesce, t0 immediately before the chord, t1 = witness arrival, flip-back + stdout readback per repeat, 40 homogeneous samples, report with methodology + memory checkpoints, budget gate exit≠0, mise task e2e-perf"
    requirement: INST-03
    verification:
      - kind: e2e
        ref: "mise run e2e-perf — two full live runs 2026-09-16: 40/40 repeats completed, perf-report.txt written, gate exit≠0 on the latency breach (the gate's designed behavior)"
        status: pass
      - kind: unit
        ref: "test/e2e/perf_test.go#TestPerfBudgetGate (gate semantics both sides)"
        status: pass
    human_judgment: false
  - id: D4
    description: "INST-03 latency half PROVEN (p95 < 50 ms in the ydotool→AT-SPI window): NOT achieved — measured p95 190–205 ms over two full runs; decomposition attributes ~170 ms to ydotool 0.1.8 per-event injection overhead (spawn alone is 3 ms; the daemon reacts in 0.2–1.8 ms). Requires an owner decision: sanction the STACK.md Go-uinput-injector fallback (~150 LOC, stand-only) vs reinterpret the window vs accept the failure"
    requirement: INST-03
    verification:
      - kind: e2e
        ref: "perf-report.txt runs 1–2 (2026-09-16): p95 204.4 ms / 190.2 ms — budget FAIL verdicts, gate exit≠0"
        status: fail
    human_judgment: true
    rationale: "The measured window is dominated by the stand's injector (a falsified research assumption: 'spawn ydotool < 10 ms' held only for --help, not per key event). Whether the acceptance window may exclude the injector's own serialization cost — or the injection mechanism may change to the researched uinput fallback — is a methodology decision the plan forbids making silently (budget is a gate, never redefined by the executor)."
  - id: D5
    description: "INST-03 memory half: VmHWM < 50 MB (D-45 oracle, one read at run end by pid) + VmRSS checkpoints in the report"
    requirement: INST-03
    verification:
      - kind: e2e
        ref: "perf-report.txt: vm_hwm_kb 12364 < 51200, vmrss_start_kb 7808 → vmrss_end_kb 12364"
        status: pass
    human_judgment: false

# Metrics
duration: 35min
completed: 2026-09-16
status: complete
---

# Phase 4 Plan 4: Perf-приёмка INST-03 Summary

**Резидентный событийный свидетель AT-SPI + perf-прогон N=40 с бюджет-гейтом: память держится (VmHWM 12.4 МБ < 50 МБ), латентность по предписанному окну ydotool→AT-SPI ЧЕСТНО ПРОВАЛИВАЕТ гейт (p95 190–205 мс) — окно доминирует оверхедом инжектора ydotool 0.1.8 (~80–125 мс/событие), реакция самого демона 0.2–1.8 мс; эскалация методики владельцу**

## Performance

- **Duration:** 35 min
- **Started:** 2026-09-16T13:13:51Z
- **Completed:** 2026-09-16T13:49:30Z
- **Tasks:** 2
- **Files modified:** 8 (+702/−12)

## Accomplishments
- Резидентный свидетель (focus_helper.py `witness-events`): Atspi.EventListener на object:text-changed, строка-на-событие `<RFC3339> <app> <pid> <text>` с таймстемпом прибытия — кванта опроса НЕТ, Падение 3 (квантование свидетеля ≥ бюджета) закрыто конструктивно; резервный `witness-poll` с именованным квантом 5 мс — тот же протокол, живой смок обеих режимов зелёный.
- Чистая математика perf под юнит-корпусом (RED→GREEN): точные перцентили по формуле `s[int(float64(len(s)-1)*p)]`, парсер VmHWM/VmRSS с именованной ошибкой отсутствующей строки, бюджет-гейт по обеим сторонам обеих границ (50 мс / 50·1024 кБ).
- Perf-кейс в реестре стенда: прод-форма демона (без -config/-debug), свежий zenity на повтор (Enter-закрытие = ADR-004 сброс буфера), quiesce до t0, t0 непосредственно перед инжекцией аккорда, t1 = прибытие события свидетеля, flip-back и readback-подтверждение повтора вне окна, отчёт stdout + perf-report.txt с блоком методики, VmRSS-контрольные, VmHWM и вердиктом бюджета; mise-задача `e2e-perf`.
- Диагностический вердикт (первый живой шаг, -debug-трейс): комбо Shift+RightCtrl возбуждается НЕМЕДЛЕННО на нажатии Control_R (press→combo 0.03 мс), коррекция завершается через 0.10–0.48 мс (путь кэша спонтанных surrounding-пушей), флип сразу после — окна различения нет, verify_wait-сталла нет. SWCH-02-семантика подтверждена, бюджет не «спасается» внутренними числами.
- Два полных живых прогона 40/40 повторов: воркер, свидетель, отчёт и гейт работают как спроектировано; гейты сработали на превышение латентности (дважды, воспроизводимо), teardown вернул стол владельцу (verifyRestored зелёный).

## Task Commits

1. **Task 1 (tdd): Свидетель + чистая математика perf** — RED `6d9ce4d` (test), GREEN `2028f08` (feat); REFACTOR не требовался
2. **Task 2 (auto): Perf-кейс + отчёт + гейт + mise** — `1b443eb` (feat)

## Files Created/Modified
- `test/e2e/perf.go` — percentile/parseProcStatus/readProcStatus/budgetGate, perfWitness-клиент (один spawn, line-протокол, pid-shutdown, ctx-bound), perf-кейс, кнобы прогона, отчёт
- `test/e2e/perf_test.go` — юнит-корпус математики (4 теста)
- `test/e2e/focus_helper.py` — witness-events (событийный свидетель) + witness-poll (резерв, квант 5 мс), docstring-таблица дополнена
- `test/e2e/main.go` — caseSpec.watchdog (пер-кейсный оверрайд), startDaemonPlain (прод-форма), реестр + usage + error-строка
- `test/e2e/matrix.go` — сигнатура runCaseWatchdog (лимит параметром, матрица передаёт 0 = дефолт)
- `mise.toml` — задача e2e-perf
- `.gitignore` — perf-report.txt (генерируемый артефакт)

## Decisions Made
- Событийный режим свидетеля запинен смоком (A2 закрыт положительно); опросник — документированный резерв с квант-константой.
- Свежая поверхность на каждый повтор — единственный корректный способ очистки: ctrl+a-очистка на живой поверхности доказуемо портит буфер («приветghbdtn» → миксованная коррекция), а Enter-закрытие = штатный CORR-09/ADR-004 сброс.
- Flip-back тап гейтится count-based waitForNew — после пробы в логе уже есть её «to":"en», и обычный waitForLog проходит по устаревшей записи, закрывая поверхность с открытым окном тапа.
- Прод-форма демона для чисел приёмки (T-04-04-03): без -debug; гейтинг регистрации/фокус_In — по счётчику, не по наличию записи.
- Вердикт бюджета оставлен ЧЕСТНЫМ провалом (гейт — приёмка, не метрика): методика не подгонялась под бюджет, эскалация — владельцу (см. Issues).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Арифметическая неувязка пина перцентилей в плане**
- **Found during:** Task 1 (GREEN)
- **Issue:** behavior-блок плана пинит p95 == p99 == 100 мс на выборке n=10, но нормативная формула плана `s[int(float64(len(s)-1)*p)]` даёт индекс 8 (9·0.95 = 8.55, 9·0.99 = 8.91 → int 8) → 90 мс; {50,100,100} не даёт ни одна непротиворечивая формула
- **Fix:** тест пинит фактический вывод формулы (p50=50, p95=p99=90) с комментарием; формула сохранена дословно (греп-пин приёмки) — она же закреплена в RESEARCH Don't Hand-Roll
- **Files modified:** test/e2e/perf_test.go
- **Verification:** юнит-корпус зелёный; grep-пин формулы в perf.go (2 вхождения)
- **Committed in:** 2028f08

**2. [Rule 1 - Bug] Гонка устаревшей записи в flip-back гейте (первая реализация)**
- **Found during:** Task 2 (первый живой прогон)
- **Issue:** flip-back ждал `"to":"en"` через waitForLog — запись flip-back'а ПРОБЫ уже в логе, wait проходил мгновенно, поверхность закрывалась с открытым окном тапа
- **Fix:** count-based waitForNew (enBase+1) вокруг тапа; тот же класс, что уже лечился для регистрации/фокус_In (activateGoswitchFresh)
- **Files modified:** test/e2e/perf.go
- **Verification:** повторный полный прогон — 40/40 повторов с корректным порядком combo→done→ru→tap→en→close
- **Committed in:** 1b443eb

**3. [Rule 3 - Blocking] .gitignore для perf-report.txt**
- **Found during:** Task 2 (перед коммитом)
- **Issue:** perf-report.txt — генерируемый артефакт прогона (verify требует его существование в корне), не входил в files_modified; оставлять незаигнорированным — незачищаемый мусор
- **Fix:** строка в .gitignore рядом с e2e-report.txt
- **Files modified:** .gitignore
- **Verification:** git status чист после прогонов
- **Committed in:** 1b443eb

**Total deviations:** 3 auto-fixed (2 × Rule 1, 1 × Rule 3)
**Impact on plan:** Правки не меняют методику и объём; все три необходимы для корректности измерения и чистоты итерации.

## Issues Encountered

**ГЛАВНОЕ — вердикт бюджета (реальный результат, не подгонять):**
- Два полных прогона e2e-perf: p95 = 204.4 мс / 190.2 мс (p50 171.9–175.2, p99 191.3–222.4) — плотное, воспроизводимое распределение; гейт exit ≠ 0 отработал как спроектировано.
- Декомпозиция (диагностика этой же сессии): реакция демона 0.10–0.48 мс (combo→done); ydotool 0.1.8: `--help` 3 мс ( assumption «spawn < 10 мс» верна только для спавна), `key Return` ≈ 250 мс, аккорд `SHIFT_R+CTRL_R` ≈ 352 мс — **~80–125 мс накладных на СОБЫТИЕ** (жизненный цикл uinput внутри вызова). Решающее событие аккорда — второе (Ctrl_R press под Shift) → ~176 мс ≈ измеренному p50. Реакция продукта «аккорд дошёл → поле исправлено» ≈ 10–40 мс — в бюджет укладывается, но предписанное окно D-44 «ydotool→AT-SPI» включает инжектор, и с пиннным инжектором окно математически не может быть < 50 мс.
- ФЛАГИРОВАННОЕ допущение плана/исследования («spawn ydotool < 10 мс — остаётся в окне как честная верхняя оценка») ОПРОВЕРГНУТО измерением на уровне события. Эскалация владельцу (Rule 4-класс — смена механизма/методики молча запрещена): (а) санкционировать фолбэк STACK.md — собственный uinput-инжектор ~150 LOC stdlib ioctls, только для стенда («симулирует человека»), окно честно сужается до < 10 мс инжекции; (б) переопределить окно методики (исключить серийную стоимость инжектора, зафиксировать t0 на событие инжекции, а не на вызов); (в) принять результат как есть. README-таблица D-46 (план 04-07) публикации чисел ДОЖИДАЕТСЯ этого решения.

**Прочее (среда/чужой код):**
- combo-word-layout дважды упал на ФИНАЛЬНОМ утверждении (post-reload одиночный тап после closeZenity — нет сфокусированного input-context на столе в текущем состоянии сессии; кейс исторически зелёный). Чужой для плана код — записано в deferred-items; perf-кейс спроектирован без этого класса (flip-back ДО закрытия поверхности).
- Диагностику дважды загрязнял забытый zenity от смока (pid-путаница `$!` subshell) — убран, чистота стола проверяется перед прогонами.

## TDD Gate Compliance

| Task | RED | GREEN | REFACTOR | Status |
|------|-----|-------|----------|--------|
| 1 (witness + math) | ✓ 6d9ce4d (4 corpus tests fail on assertions vs stubs; RED_EVIDENCE_OK by gate) | ✓ 2028f08 | — (not needed) | Pass |
| 2 (perf case) | n/a (type="auto") | ✓ 1b443eb | — | Pass |

Evidence record: `.planning/phases/04-postavka-i-priemka/red-evidence/04-04-task1-red.json` (TAP-transcribed per repo precedent, verified RED_EVIDENCE_OK).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Инфраструктура INST-03 полностью готова и воспроизводима: `mise run e2e-perf` — один прогон даёт отчёт + вердикт гейта; артефакт формата README-таблицы (D-46) пишется каждый прогон.
- Памятная половина INST-03 ЗАКРЫТА измерением (VmHWM 12.4 МБ; в 4 раза лучше бюджета).
- БЛОКЕР для 04-07 (README-таблица чисел): публикация латентности ждёт решения владельца по методике окна (см. Issues). До решения — числа из perf-report.txt публиковать НЕЛЬЗЯ (гейт красный).
- WINDOWS ledger: дополнен записью unmet-truth по латентному вердикту (best-effort).

---
*Phase: 04-postavka-i-priemka*
*Completed: 2026-09-16*

## Self-Check: PASSED

- Files on disk: perf.go, perf_test.go, focus_helper.py (witness modes), mise e2e-perf task, perf registry entry, red-evidence record, SUMMARY — all FOUND
- Commits: 6d9ce4d (RED), 2028f08 (GREEN), 1b443eb (feat perf case), 93ef281 (docs) — all present in git log
- Working tree clean; desktop left healthy (teardown restore verified by every live run)
