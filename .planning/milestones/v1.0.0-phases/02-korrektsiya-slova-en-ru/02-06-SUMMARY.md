---
phase: 02-korrektsiya-slova-en-ru
plan: 06
subsystem: testing
tags: [e2e, yaml-matrix, ibus, at-spi, case-isolation, test-04, d-18]

requires:
  - phase: 02-korrektsiya-slova-en-ru
    provides: word/flip/ladder/reset live cases (02-03..02-05) and the chromium/gnome-text-editor surface drivers (02-02)
provides:
  - YAML case matrix v1 (16 D-18 cases on zenity+chromium+gnome-text-editor) driven by the phase-1 stand with per-case fresh-daemon isolation
  - Strict matrix loader (yaml.v3 KnownFields) + PASS/FAIL report with exit≠0 contract and e2e-report.txt artifact
  - mise task e2e-matrix (same interface for local runs and the 02-07 CI runner)
affects: [02-07-ci-runner, verify-work, phase-3-regression-loop]

actuals:
  tokens: 16444   # chars/4 over the realized diff (12 files, +1494/-24)
  tasks: 2
  commits: 3      # measured: git rev-list --count 5ae70edb..HEAD
  plan_head_before: 5ae70edbb7e203c8737ec10f2e2e98c7840d50dc

tech-stack:
  added: ["gopkg.in/yaml.v3 v3.0.1 (the only new dependency — Approved in the research audit)"]
  patterns:
    - "per-case isolation as an explicit run-loop step: setupStand → watchdog(case) → teardown per case — daemon-carried state (scriptMode, buffer) never leaks, file order cannot matter"
    - "strict declarative schema with closed vocabularies (surface/mode/tap/key/focus) — a typo is a decode error with the case number and name, never a silent skip"

key-files:
  created:
    - test/e2e/matrix.go
    - test/e2e/matrix_test.go
    - test/e2e/cases/matrix-v1.yaml
  modified:
    - test/e2e/main.go
    - test/e2e/focus_helper.py
    - test/e2e/README.md
    - mise.toml
    - go.mod
    - go.sum
    - .gitignore

key-decisions:
  - "expect_level pinned to ACTUAL live levels (1 on every matrix surface of this desktop): the plan's expect_level:2 wording for chromium was falsified live by 02-05 (google-chrome 153 reports caps 0x29 and APPLIES DeleteSurroundingText); the level-2 contract stays in the unit corpus — no level-2 witness exists on this target"
  - "verify-after match is NOT a matrix oracle: the daemon's verify round settles on the first surrounding-text push and clients can push the intermediate post-delete pre-commit state — false mismatch on a correct field, nondeterministic across surfaces (hit GTE then zenity on consecutive runs); content-exact readback is the ground truth, the race filed as a deferred item"
  - "the RU-mode flip runs AFTER the surface opens (the proven 02-04 order): a pre-surface flip tap reaches no engine input context at all (0 key records live)"
  - "reset-tab needs an explicit {focus: chromium} step after Tab: focus lands on the browser omnibox whose context delivers no keys to the engine, and the omnibox (the URL's ~60 chars) is chrome's first focusable input — grab-input-pid gained a want-chars filter to aim the poke at the page input"
  - "matrix preflight refuses to start against a live goswitchd process or a stale ibus-side name registration (leftover of an interrupted run; GetNameOwner probe — ibus words the free-name answer 'no such name'); a11y quiesce gate between cases waits out the registry's asynchronous reaping of killed surfaces"

patterns-established:
  - "quiesce-then-open: witness probe with a 4 s budget before each case's polls — wedged AT-SPI walks cost seconds, not the 15 s command deadline, and transient wedges are waited out instead of failing healthy cases"
  - "focused-inputs enumeration beats the single witness wherever two stand surfaces coexist: a stale AT-SPI FOCUSED bit on the surface being left must not mask the fresh focus"

requirements-completed: [TEST-04]

coverage:
  - id: D1
    description: "Strict YAML matrix loader and report/exit-code contract (TEST-04 headless half)"
    requirement: TEST-04
    verification:
      - kind: unit
        ref: "test/e2e/matrix_test.go#TestMatrixDecode_Valid"
        status: pass
      - kind: unit
        ref: "test/e2e/matrix_test.go#TestMatrixDecode_UnknownFieldRejected"
        status: pass
      - kind: unit
        ref: "test/e2e/matrix_test.go#TestMatrixReport_ExitCode"
        status: pass
    human_judgment: false
  - id: D2
    description: "Full D-18 word corpus green on a live session — 16 cases, three surfaces, three registers on gnome-text-editor (criterion 1 verbatim)"
    requirement: TEST-04
    verification:
      - kind: e2e
        ref: "mise run e2e-matrix — 16/16 PASS (ordered run, e2e-report.txt)"
        status: pass
      - kind: e2e
        ref: "command: go run ./test/e2e -matrix /tmp/matrix-shuffled.yaml — 16/16 PASS in shuffled order"
        status: pass
    human_judgment: false
  - id: D3
    description: "Per-case isolation — fresh daemon per case, EN start, empty buffer, order-independence"
    requirement: TEST-04
    verification:
      - kind: e2e
        ref: "command: shuffled-order full run 16/16 PASS; word-ru-en/word-mixed (RU-mode leavers) pass before en-cases in both orders"
        status: pass
    human_judgment: false
  - id: D4
    description: "FAIL reporting and exit≠0 on a broken expectation"
    requirement: TEST-04
    verification:
      - kind: e2e
        ref: "command: corrupted-expectation copy — FAIL word-en-ru (readback oracle), exit code 1, e2e-report.txt carries the line"
        status: pass
    human_judgment: false
  - id: D5
    description: "mise task e2e-matrix and the stand README matrix section (run book: schema, isolation why, expect_level, FAIL behavior, single-stand rule)"
    verification:
      - kind: other
        ref: "mise.toml [tasks.e2e-matrix]; test/e2e/README.md «Матрица v1»"
        status: pass
    human_judgment: false

duration: 109 min
completed: 2026-09-15
status: complete
---

# Phase 2 Plan 06: YAML-матрица e2e-кейсов v1 Summary

**16 декларативных кейсов D-18 на zenity+chromium+gnome-text-editor с изоляцией «свежий демон на кейс», зелёных живьём в прямом и перемешанном порядке; строгий yaml.v3-декодер, отчёт PASS/FAIL с кодом выхода ≠ 0**

## Performance

- **Duration:** 109 min (6 полных живых прогонов + разбор двух живых деградаций стола)
- **Started:** 2026-09-14T22:31:59Z
- **Completed:** 2026-09-15T00:21:24Z
- **Tasks:** 2
- **Files modified:** 12 (+1494/−24)

## Accomplishments

- Раннер матрицы: строгий мультидок-декодер (KnownFields, закрытые словари surface/mode/tap/key/focus), шаговый прогон type/key/tap/focus над примитивами стенда, сверка expect_text руна в руну (AT-SPI readback с settle-поллингом; stdout-оракул для закрытых диалогов), expect_level — фактический уровень лестницы из -debug-лога (ADR-003)
- Изоляция кейса — явный шаг цикла: setupStand → кейс (watchdog 180 с) → teardown на КАЖДЫЙ кейс; перемешанный порядок даёт те же 16/16 PASS
- Полный словесный набор D-18: оба направления, три регистра (на gnome-text-editor — критерий 1 дословно), слово после пробела, пунктуация/цифры, смешанное слово, сбросы Escape/Enter/Tab/фокус
- Живые находки и их проводка: флип после поверхности; омнибокс-ловушка Tab (расширение grab-input-pid want-chars); устаревший бит FOCUSED у уходящей поверхности (гейт по перечислению); префлайт против живого процесса/осиротевшего имени goswitch; a11y-quiesce между кейсами

## Task Commits

1. **Task 1 RED: корпус декодера и exit-кода** — `61adf95` (test)
2. **Task 1 GREEN: раннер матрицы** — `3b7051d` (feat)
3. **Task 2: полный набор D-18 + README + живые фикс-находки** — `bdcb069` (feat)

## TDD Gate Compliance

RED `61adf95` (test) строго предшествует GREEN `3b7051d` (feat); red-evidence
`red-evidence/02-06-task1-red.json` верифицирован гейтом (`RED_EVIDENCE_OK`,
target `TestMatrixDecode_Valid` — падение на утверждении, не на компиляции).
После GREEN корпус зелёный: `go test ./test/e2e/ -race -count=1` (6 тестов).

## Files Created/Modified

- `test/e2e/matrix.go` — загрузчик, валидация, изолированный прогон, шаги, сверка, отчёт+e2e-report.txt, префлайт (процесс/имя/quiesce)
- `test/e2e/matrix_test.go` — пины декодера и exit-кода (6 тестов, пакет стенда)
- `test/e2e/cases/matrix-v1.yaml` — 16 кейсов D-18 с решениями в комментариях
- `test/e2e/main.go` — флаг -matrix (ветка до общего цикла стенда; реестр -case нетронут)
- `test/e2e/focus_helper.py` — grab-input-pid <pid> [want-chars] (прицельный grab)
- `test/e2e/README.md` — раздел «Матрица v1»
- `mise.toml` — [tasks.e2e-matrix]; `.gitignore` — e2e-report.txt
- `go.mod`/`go.sum` — единственная новая зависимость gopkg.in/yaml.v3 v3.0.1

## Decisions Made

- expect_level по факту, не по предположению (все поверхности стола = уровень 1; уровень 2 — юнит-контракт; план-формулировка «fallback-уровень 2» опровергнута живьём ещё в 02-05)
- verify-match не оракул матрицы (гонка устаревшего пуша; отложенный пункт)
- Ключ step {key: space} идёт typing-путём, Escape→esc, Enter→enter, Tab→tab (+{focus: chromium} в кейсе) — имена ydotool 0.1.8 закреплены в одной таблице

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Порядок флипа: flip до поверхности не достигает движка**
- **Found during:** Task 2 (word-ru-en FAIL в первом полном прогоне)
- **Issue:** План ставил mode-шаг ДО открытия поверхности; тап уходил в контекст без фокуса стенда — 0 записей key/mode в логе
- **Fix:** Флип после openMatrixSurface (проверенный порядок 02-04); инъекция всегда в собственную поверхность
- **Files modified:** test/e2e/matrix.go
- **Verification:** word-ru-en PASS (subset + оба полных прогона)
- **Committed in:** bdcb069

**2. [Rule 1 - Bug] reset-tab: омнибокс глотает Tab и не доставляет клавиши движку**
- **Found during:** Task 2 (первый полный прогон)
- **Issue:** После Tab фокус в омнибоксе (ENTRY, ~63 символа URL); его контекст не доставляет клавиши движку — решение FSM не наблюдаемо; первый focusable-узел chrome = омнибокс, poke бил мимо поля
- **Fix:** Кейс получил {focus: chromium} после Tab; helper — прицельный grab по количеству символов
- **Files modified:** test/e2e/cases/matrix-v1.yaml, test/e2e/focus_helper.py, test/e2e/matrix.go
- **Verification:** reset-tab PASS (subset + оба полных прогона)
- **Committed in:** bdcb069

**3. [Rule 1 - Bug] reset-focus: устаревший бит FOCUSED уходящей поверхности маскировал флип**
- **Found during:** Task 2 (первый полный прогон)
- **Issue:** Единичный witness продолжал называть input chrome (бит FOCUSED), пока zenity уже держал реальный фокус
- **Fix:** Гейт spawn-zenity по перечислению focused-inputs (строка zenity: среди кандидатов)
- **Files modified:** test/e2e/matrix.go
- **Verification:** reset-focus PASS (subset + оба полных прогона)
- **Committed in:** bdcb069

**4. [Rule 2 - Missing Critical] Префлайт матрицы: живой процесс goswitchd / осиротевшее имя IBus**
- **Found during:** Task 2 (полный прогон упал 16× на регистрации после осиротевшего стенда)
- **Issue:** Оборванный прогон оставляет имя без владельца (ibus-daemon 1.5.29 держит запись); все 16 кейсов падали криптически, а два параллельных стенда запрещены планом
- **Fix:** Проверки goswitch-process-absent (pgrep) и goswitch-name-free (GetNameOwner; ibus отвечает «no such name»); у ошибки регистрации — точная причина
- **Files modified:** test/e2e/matrix.go
- **Verification:** Префлайт ловит оба состояния живьём; после `ibus restart` — зелёный прогон
- **Committed in:** bdcb069

**5. [Rule 2 - Missing Critical] a11y-quiesce между кейсами**
- **Found during:** Task 2 (перемешанный прогон: 4 FAIL «signal: killed» на открытиях поверхностей)
- **Issue:** Регистр AT-SPI репаит убитые teardown-ом поверхности асинхронно; обход в этом окне блокируется до полного cmdTimeout — падают здоровые кейсы
- **Fix:** Проба witness с бюджетом 4 с и ожиданием затишья (до 30 с) перед опросами кейса
- **Files modified:** test/e2e/matrix.go
- **Verification:** Перемешанный прогон после лечения стола — 16/16 без «signal: killed»
- **Committed in:** bdcb069

**6. [Rule 1 - Plan assumption falsified] verify-match убран из оракулов матрицы**
- **Found during:** Task 1 (дым-прогоны)
- **Issue:** Раунд сверки демона закрывается по первому пушу; клиент может запушить промежуточное post-delete pre-commit состояние — ложный mismatch при верном поле; недетерминировано по поверхностям (GTE, затем zenity на соседних прогонах)
- **Fix:** Оракулы матрицы: expect_level + content-exact readback; матч остаётся в юнит-корпусе и ladder-chromium; гонка — отложенный пункт
- **Files modified:** test/e2e/matrix.go, .planning/.../deferred-items.md
- **Verification:** 16/16×2 прогонов стабильно зелёные
- **Committed in:** 3b7051d (снятие пина), bdcb069 (документация)

---

**Total deviations:** 6 auto-fixed (4 bug, 2 missing critical)
**Impact on plan:** Все — следствия живых прогонов (ровно то, зачем существует матрица); формат кейсов, объём набора и изоляция — без изменений. План-слово «expect_level:2» для chromium замещено фактом (02-05), зафиксировано в YAML-комментарии и здесь.

## Issues Encountered

- **Осиротевший стенд + осиротевшее имя IBus (самоиндуцировано):** короткоживущий python-хук убил только `go run`, e2e-бинарник выжил и гнал матрицу параллельно с моим полным прогоном — два стенда на одном столе (запрет плана), взаимная блокировка имени, а ibus-daemon навсегда удержал запись оборванного подключения. Вылечено `ibus restart` (прецедент — собственный кейс стенда); префлайт-проверки (отклонение 4) теперь ловят оба состояния до касания стола
- **Стойкое зависание дерева AT-SPI (окружение стола):** после серий SIGKILL поверхностей мост gnome-shell перестал отвечать реестру — все обходы висли до полного таймаута; перезапуск одного at-spi2-registryd НЕ лечит, перезапуск всей шины a11y (at-spi-bus-launcher + её dbus-daemon; стек dbus-активируется) пересобрал мосты, прогулки вернулись к ~60 мс. Стойкая форма — окружение, не кейс: задокументировано в deferred-items и README (для раннера 02-07)

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Матрица v1 зелёная и стабильная (прямой и перемешанный порядок) — регрессионный контур Фазы 2 готов стать CI-задачей раннера (02-07) без изменений формата: mise run e2e-matrix + e2e-report.txt как артефакт
- Для раннера 02-07 в README уже описаны: правило одного стенда, лечение осиротевшего имени (`ibus restart`) и лечения зависшего дерева a11y (перезапуск шины)
- Открытые пункты (deferred-items): гонка устаревшего пуша в verify-after (демон), стойкое зависание моста gnome-shell (апстрим/раннер)

---
*Phase: 02-korrektsiya-slova-en-ru*
*Completed: 2026-09-15*

## Self-Check: PASSED

All created files exist on disk; all four commits (61adf95, 3b7051d, bdcb069, c3fe36a) present in git log.
