---
phase: 06-avtokorrekcija-opcionalno
plan: 07
subsystem: testing
tags: [autocorrect, e2e, gtk4, atspi, password-entry, witness, matrix-v4, nightly, d-48, oracle, live-desktop]

# Dependency graph
requires:
  - phase: 06-avtokorrekcija-opcionalno (06-06)
    provides: счётчики-оракул D-54 (autocorrect_fired/abstained + ac_skip_*), fired-INFO-класс журнала, config-секция autocorrect с живым фолдом, verify-round конвейер актора
  - phase: 06-avtokorrekcija-opcionalno (06-03)
    provides: appid.Observer.Role + enum 40/60/61/79/94 — ролевая политика ACL
  - phase: 06-avtokorrekcija-opcionalno (06-04)
    provides: схема config-секции autocorrect (enabled/apps/min_word_len/пороги) для autocorrectConfigTmpl
  - phase: 05-avtokorrekcija (05-05, поверхность) / 04-04 (матрица)
    provides: surface-дисциплина (startZenity/witness/pid-точность), count-agnostic оракул матричного драйвера, snapshotDesktop
provides:
  - test/e2e/case_autocorrect.go — три живых кейса: runAutocorrectFires (fired-запись + correction done + pid-чтение поля + D-54-счётчики + ручной Double-регресс CORR-01 + PLAIN/FINAL-оракул), runAutocorrectPasswordSilent (witness PASSWORD_TEXT до/после с наблюдаемым числом символов, FINAL-verbatim, fired=0, ровно одна роль-отказная абстиненция; отсутствие witness = FAIL, не skip — T-06-07-01), runAutocorrectTerminalSilent (wezterm вне a11y: fired=0 + ноль коррекционных записей + app-not-listed абстиненция; наш zenity — положительный контроль фокуса)
  - test/e2e/fixtures/password_entry.py — GTK4-фикстура (GtkWindow + GtkEntry + GtkPasswordEntry), timeout-argv, FINAL:/PLAIN: stdout-оракулы; input.html — password-поле рядом с нетронутым text-input
  - test/e2e/cases/matrix-v4.yaml — superset: v3 встроена БАЙТ-ДОСЛОВНО (смежный блок, проверено программно) + autocorrect-блок (fires/off/oov, 35 кейсов); честное исключение password-строки (нет password-шага у матричного драйвера — GTK4-кейс несёт ролевой негатив автономно)
  - Ночной конвейер на v4: e2e-matrix.yml — v4 default в ОБОИХ D-48-развилках (v1/v2/v3 регресс-пути сохранены), d48-nightly-dispatch.sh → matrix-v4.yaml, docs/ci-runner.md — секция v4 (линия матриц, правила superset, замороженные красные строки)
  - Продуктовый фикс, найденный живьём: acBoundary-вооружённый verify-раунд распознаёт pre-boundary ответ (токен без хвоста) как ранний in-flight-separator ответ и остаётся открыт в рамках verify-бюджета (55dd248 RED + cbfa46c GREEN)
  - mise: [tasks.e2e-autocorrect-fires/-password-silent/-terminal-silent] + [tasks.e2e-matrix-v4] — вне [tasks.ci]
affects: [06-08 (формальный D-48 fresh-session двойной прогон на v4, ADR-007 Accept, perf p95), verify-work фазы 06]

# Actuals (#2632)
actuals:
  tokens: 25783       # chars/4 over the realized diff (103 132 bytes, 15 files) — оценка плана 65 000 не достигнута по объёму, но план — живой стенд: львиную долю оценки съели живые прогоны и чекпоинт, не диф
  tasks: 3
  commits: 6          # MEASURED: git rev-list --count gsd-plan-head-before-06-07..HEAD (8b90924..2c41b83)

# Tech tracking
tech-stack:
  added: []           # ноль новых зависимостей (go.mod не тронут)
  patterns:
    - "Три независимых оракула негатива (witness role+chars, FINAL-stdout, счётчик fired=0) — отсутствие witness = FAIL кейса, никогда skip-тишина (T-06-07-01)"
    - "Идентичность поверхности наблюдается живьём (macr-per-app дисциплина): case-side appid.Observer + фикстура пинит свой prgname → словарь org.gtk.application.gs_e2e_password, никогда python3"
    - "Наблюдатель стартует лениво от ПЕРВОГО нажатия (applySnapshot/ensureAppid едут по per-KEY-EVENT чтению снапшота) — identity-поверхность мапится только после первого keystroke (acStartObserverRound, урок surface-C 03-05)"
    - "Оракул негатива пинит свойство БЕЗОПАСНОСТИ, а не одну руку гонки доставки (терминальный re-pin 1dc76af): fired=0 + ноль коррекционных записей при любой доставке keystrokes"

key-files:
  created:
    - test/e2e/case_autocorrect.go
    - test/e2e/fixtures/password_entry.py
    - test/e2e/cases/matrix-v4.yaml
  modified:
    - test/e2e/fixtures/input.html
    - test/e2e/main.go
    - test/e2e/matrix.go
    - test/e2e/matrix_test.go
    - test/e2e/case_ctl.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - mise.toml
    - .github/workflows/e2e-matrix.yml
    - scripts/d48-nightly-dispatch.sh
    - docs/ci-runner.md
    - .planning/phases/06-avtokorrekcija-opcionalno/deferred-items.md

key-decisions:
  - "A2 спинен живьём (google-chrome 153 Wayland, 2026-10-01): chromium отдаёт <input type=password> как AT-SPI PASSWORD_TEXT (40), один Tab от autofocus ложится на него (witness PASSWORD_TEXT:chars=0) — chromium-негатив юридически выразим; носитель доказательства — GTK4-фикстура"
  - "Identity-поверхность мапится только после первого keystroke (живая находка Task 2: enabled=false до первого нажатия — applySnapshot/ensureAppid едут по per-KEY-EVENT снапшоту); acStartObserverRound переносит урок macr-per-app surface-C"
  - "Продуктовый фикс verify-раунда (живая фальсификация первого fires-прогона): RequireSurroundingText обслуживается ДО того как сепаратор доходит до клиента — первый ответ обязан быть pre-boundary; раунд остается открыт (fired=1 abstained=0, ноль операций), пост-boundary пуш закрывает; manual-пути (acBoundary=false) байт-как-сегодня"
  - "Терминальный оракул пере-пинен на свойство безопасности: wezterm РЕГИСТРИРУЕТ IBus input context (caps 0x9, нет surrounding-text бита) и его typing ГОНИТ гонку с armed-ингом IME — ноль-записей был ставкой на одну руку; 1dc76af: fired=0 + ноль коррекционных записей + active-layer status, пришедшая граница отказывается fail-closed (D-53, условие 3)"
  - "Password-строка ЧЕСТНО исключена из matrix-v4 (у матричного драйвера нет password-шага поверхности) — GTK4-кейс несёт ролевой негатив автономно; санкционировано действием плана («без ослабления валидатора»)"
  - "Session-drift деферрал: полный зелёный matrix-v4 отложен в формальный D-48 fresh-session двойной прогон плана 06-08 — потеря первого нажатия после флипа воспроизведена БАЙТ-ИДЕНТИЧНО на до-фазовом дереве d96dcb2 (/tmp/matrix-v3-oldtree.log, worktree /tmp/goswitch-0505) — диф фазы 6 exonered, дрейф средовой"

patterns-established:
  - "Живой негатив на счётчиках: когда поверхность вне a11y (wezterm), оракул = goswitchctl status (fired=0) + отрицание коррекционных записей в -debug-журнале (waitForNew/countSub-отрицание) + положительный контроль фокуса (наш zenity)"
  - "Фикстура-носитель stdout-оракула: FINAL: печатает содержимое ТОЛЬКО для проверки неизменности (ожидается неизменённое) — канал наблюдения не становится каналом утечки (T-06-07-03)"

requirements-completed: ["CORR-01", "MACR-ACL (app×role политика)", "SPEC §10/§11 (spec-delta)"]  # провизорные метки плана: CORR-01 разделяется с 06-08 без SUMMARY — шлагбаум #2388 (REQUIREMENTS.md не трогается этим планом, прецедент 06-06); свободные метки не размечаются verb'ом; формальная фиксация REQ-ID — при new-milestone v1.1.0 (D-51)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Фикстуры Task 1: GTK4 password_entry.py (GtkWindow+GtkEntry+GtkPasswordEntry, timeout-argv, FINAL:/PLAIN: stdout-оракулы), input.html password-поле рядом с нетронутым text-input; Chromium-роль спинена живьём (PASSWORD_TEXT=40) и задокументирована"
    verification:
      - kind: other
        ref: "Task 1 verify: py_compile + грепы GtkPasswordEntry/FINAL:/type=password → FIXTURES-OK (commit a1326cd)"
        status: pass
      - kind: e2e
        ref: "фикстура — живой носитель mise run e2e-autocorrect-password-silent (PASS)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Три живых кейса case_autocorrect.go с честными оракулами (witness+FINAL+счётчики), зарегистрированы во всех трёх местах (регистр/caseListUsage/error-список), mise-задачи вне ci"
    verification:
      - kind: other
        ref: "Task 2 verify: go build/vet ./test/e2e/ + REGISTERED-OK + MISE-OK грепы (commit 46e336e)"
        status: pass
      - kind: e2e
        ref: "mise run e2e-autocorrect-fires / e2e-autocorrect-password-silent / e2e-autocorrect-terminal-silent — зелёные живьём (см. таблицу живых результатов)"
        status: pass
    human_judgment: false
  - id: D3
    description: "matrix-v4.yaml — дословный superset замороженной v3 (байт-верификация смежного блока) + autocorrect-блок (fires/off/oov); ночной конвейер на v4 (workflow default в обеих развилках, dispatch-скрипт, docs/ci-runner.md); matrix.go валидатор расширен явно (config_base, fresh-старт)"
    verification:
      - kind: other
        ref: "Task 3 verify: V3-FROZEN (git diff --exit-code на matrix-v3.yaml) + NIGHTLY-V4-OK грепы; v3-вложение проверено программно (commit 2c41b83); пере-проверено на close-out"
        status: pass
    human_judgment: false
  - id: D4
    description: "Живые доказательства на реальном столе: autocorrect-fires (исправление без хоткея + ручной Double-регресс), autocorrect-password-silent (пароль нетронут, роль 40), autocorrect-terminal-silent (молчание под обеими режимами доставки IME)"
    verification:
      - kind: e2e
        ref: "mise run e2e-autocorrect-fires → PASS; e2e-autocorrect-password-silent → PASS; e2e-autocorrect-terminal-silent → PASS (два подряд зелёных после re-pin 1dc76af)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Полный зелёный прогон mise run e2e-matrix-v4 (гейт Task 3 verify №4): НЕ достигнут на этой сессии — autocorrect-строки зелёные ×2, полный прогон красный на до-существующих v3-строках из-за средового session-drift"
    requirement: "SPEC §10/§11 (spec-delta)"
    verification:
      - kind: e2e
        ref: "mise run e2e-matrix-v4 — autocorrect rows (fires/off/oov) green ×2; FULL red: first-key-after-flip loss на pre-existing v3 rows"
        status: fail
    human_judgment: true
    rationale: "Контрольный эксперимент: до-фазовое-6 дерево d96dcb2 (последнее живо-зелёное 05-05) падает БАЙТ-ИДЕНТИЧНО (/tmp/matrix-v3-oldtree.log, worktree /tmp/goswitch-0505) — диф фазы 6 exonered, дрейф rides сессию, не код. FLAGGED-допущение самого плана уже маршрутизировало формальный гейт (двойной свежесессионный прогон D-48) в план 06-08; чекпоинт закрыт решением accept & defer (2026-10-02). 06-08 и verify-work судят на свежей сессии; до тех пор D5 не может быть auto-pass."

# Metrics
duration: 19h span (3 executor continuations: задачи 1-2, задача 3 + чекпоинт, close-out)
completed: 2026-10-02
status: complete
---

# Phase 6 Plan 7: Живые доказательства автокоррекции + матрица v4 Summary

**Три живых e2e-кейса доказали на реальном столе, что автокоррекция исправляет текст без хоткея и МОЛЧИТ в GTK4-пароле (роль 40, три независимых оракула) и в wezterm (оракул по счётчикам, устойчив к гонке доставки IME); матрица v4 — байт-дословный superset замороженной v3 с autocorrect-блоком, ночной конвейер переключён на v4; попутно найден живьём и закрыт продуктовый баг verify-раунда (pre-boundary ответ убивал каждую fired-коррекцию).**

## Performance

- **Duration:** ~19h wall-clock span (2026-10-01T23:01Z → 2026-10-02T18:05Z) через 3 executor-континуации: (1) задачи 1-2, (2) задача 3 + живые прогоны + блокирующий чекпоинт, (3) close-out после решения accept & defer. Полный зелёный matrix-v4 на этой сессии не достигнут — см. Issues/Deviations.
- **Started:** 2026-10-01T23:01:52Z (первый коммит a1326cd)
- **Completed:** 2026-10-02 (close-out, mise run ci зелёный на финальном дереве)
- **Tasks:** 3
- **Files modified:** 15 (измерено git diff --numstat 8b90924..2c41b83)

## Accomplishments

- **Три живых кейса зелёные** (см. таблицу): fires исправил ghbdtn→привет БЕЗ хоткея (pid-ключенное чтение поля + D-54-счётчики + ручной Double конвертировал обратно — CORR-01 регресс на той же поверхности); password-silent — witness PASSWORD_TEXT до/после, символы нетронуты, FINAL-verbatim, fired=0, ровно одна роль-отказная абстиненция; terminal-silent — молчание под ОБЕИМИ руками гонки доставки (после re-pin 1dc76af — два подряд зелёных)
- **Продуктовый фикс найден живьём (TDD RED→GREEN):** граница срабатывает НА сепараторе, RequireSurroundingText обслуживается клиентом ДО его доставки — первый verify-ответ обязан быть pre-boundary; 06-06-корпус моделировал первый ответ пост-boundary пушем (форма, недостижимая живьём). Фикс cbfa46c: acBoundary-вооружённый раунд распознаёт ранний ответ и остаётся открыт в рамках существующего verify-бюджета; чужой ответ по-прежнему verify-mismatch (guard-тест); manual-пути байт-как-сегодня. БЕЗ этого фикса каждая fired-коррекция умирала на verify-mismatch на таких клиентах (zenity выигрывал только таймингом)
- **matrix-v4 — честный superset:** все v3-строки встроены байт-дословно (смежный блок, программная верификация; гейт git diff --exit-code на v3 зелёный), autocorrect-блок (fires/off/oov, 35 кейсов), password-строка честно исключена (нет password-шага у драйвера матрицы); валидатор matrix.go расширен ЯВНО (config_base-словарь, fresh-старт, индексированный цикл) с корпусом matrix_test.go
- **Ночной конвейер на v4:** e2e-matrix.yml — v4 default входа в ОБОИХ D-48-развилках (v1/v2/v3 регресс-пути сохранены); d48-nightly-dispatch.sh → matrix-v4.yaml fresh_session=true; docs/ci-runner.md — секция v4 (линия матриц, правила superset, замороженные красные строки)
- **A2 спинен живьём:** chromium отдаёт password-поле как AT-SPI PASSWORD_TEXT (40), один Tab от autofocus ложится на него — chromium-негатив юридически выразим; GTK4-фикстура — носитель доказательства
- **mise run ci зелёный на финальном дереве** (known flake TestRun_OnConnHookCalledOnce воспроизвёлся 1× на close-out прогоне, изолированный ре-ран и полный ре-ран зелёные — протокол одного ре-рана; root-cause уже в deferred-items)
- **Угрозы threat-модели закрыты:** T-06-07-01 (три независимых оракула, отсутствие witness = FAIL — реализовано в password-silent); T-06-07-02 (v3 заморожена, гейт зелёный); T-06-07-03 (FINAL печатает только для проверки неизменности, настоящий секрет не вводится); T-06-07-04 (autocorrect-строки зелёные ×2 ДО переключения default; полный гейт честно красный — деферрал задокументирован, тихого ослабления нет)

### Живые результаты (4 гейта)

| Гейт (mise) | Результат | Оракулы/доказательство |
|---|---|---|
| e2e-autocorrect-fires | **PASS** | fired-запись (`"reason":"fired"`) + correction done + pid-ключенное AT-SPI-чтение (ghbdtn → привет) + status `autocorrect_fired=1`/`enabled=true` + ручной Double конвертировал обратно |
| e2e-autocorrect-password-silent | **PASS** | witness role=PASSWORD_TEXT(40) до и после с наблюдаемым числом символов, chars неизменны, stdout фикстуры FINAL == неизменённое, `autocorrect_fired=0`, ровно одна role-forbidden абстиненция |
| e2e-autocorrect-terminal-silent | **PASS** (после re-pin 1dc76af, 2 подряд зелёных) | `autocorrect_fired=0` + ноль коррекционных записей + app-not-listed абстиненция; пришедшая граница отказана fail-closed (D-53 условие 3); наш zenity — положительный контроль фокуса |
| e2e-matrix-v4 | **ЧАСТИЧНО на этой сессии** | autocorrect-строки (fires/off/oov) зелёные ×2 в обоих матричных прогонах; ПОЛНЫЙ прогон красный на до-существующих v3-строках — средовой session-drift (см. ниже); формальный D-48 fresh-session двойной прогон → план 06-08 |

## Task Commits

Each task was committed atomically:

1. **Task 1: Фикстуры — GTK4 password_entry.py + input.html password-поле (A2 спинен живьём)** — `a1326cd` (test)
2. **Task 2: case_autocorrect.go — три кейса + регистрация в трёх местах + mise-задачи** — `46e336e` (feat)
3. **Task 3: матрица v4 superset + ночной конвейер на v4** — `2c41b83` (feat; статические деливераблы + deferred-items записи)
4. **Отклонения живых прогонов:** `55dd248` (test, RED — pre-boundary ранний ответ) → `cbfa46c` (feat, GREEN — раунд ждёт пост-boundary пуш); `1dc76af` (fix — re-pin терминального оракула на свойство безопасности)

**Plan metadata:** _docs(06-07) — этот и следующий коммит (прецедент 06-06: SUMMARY и STATE/ROADMAP отдельными коммитами)_

## Files Created/Modified

- `test/e2e/case_autocorrect.go` — три кейса, autocorrectConfigTmpl (полный YAML-документ), startAutocorrectConfigDaemon (лестница case_macr.go), acStartObserverRound (наблюдатель от первого keystroke), case-side appid.Observer (наблюдаемая идентичность)
- `test/e2e/fixtures/password_entry.py` — GTK4-фикстура (исследовательски верифицированный прототип): GtkEntry + GtkPasswordEntry, timeout-argv, FINAL:/PLAIN: stdout-оракулы
- `test/e2e/fixtures/input.html` — `<input type="password">` рядом с нетронутым text-input
- `test/e2e/main.go` — регистрация трёх кейсов в регистре/caseListUsage/error-списке; stand-структура несёт фикстуру/wezterm-процессы с teardown
- `test/e2e/matrix.go` + `matrix_test.go` — config_base-словарь, fresh-старт, индексированный цикл кейсов + корпус
- `test/e2e/cases/matrix-v4.yaml` — 35 кейсов: v3 байт-дословно + autocorrect-блок
- `test/e2e/case_ctl.go` — configLoadedMark в shared const (goconst)
- `internal/session/actor.go` + `actor_test.go` — продуктовый фикс verify-раунда (55dd248+cbfa46c): ранний pre-boundary ответ держит раунд открытым
- `mise.toml` — [tasks.e2e-autocorrect-*] ×3 + [tasks.e2e-matrix-v4], все вне [tasks.ci]
- `.github/workflows/e2e-matrix.yml` — v4 default в обеих D-48-развилках; `scripts/d48-nightly-dispatch.sh` → matrix-v4.yaml; `docs/ci-runner.md` — секция v4
- `.planning/.../deferred-items.md` — root-cause флака OnConnHookCalledOnce + session-drift находка с контрольным экспериментом

## Decisions Made

См. key-decisions: A2 спинен живьём (chromium PASSWORD_TEXT=40); identity-поверхность мапится после первого keystroke; verify-раунд ждёт pre-boundary ответ; терминальный оракул = свойство безопасности; password-строка честно исключена из v4; session-drift деферрал в 06-08.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking lint] configLoadedMark в shared const**
- **Found during:** Task 2 (GREEN — mise run ci)
- **Issue:** литерал `"msg":"config loaded"` дублировался между case_macr.go и case_autocorrect.go — goconst строгого линта не пропускает
- **Fix:** константа вынесена в shared (case_ctl.go), оба потребителя ссылаются
- **Files modified:** test/e2e/case_ctl.go, test/e2e/case_autocorrect.go
- **Verification:** mise run ci зелёный
- **Committed in:** 46e336e

**2. [Rule 1 - Bug] Identity-поверхность мапится только после первого keystroke**
- **Found during:** Task 2 (живой прогон fires)
- **Issue:** applySnapshot/ensureAppid едут по per-KEY-EVENT чтению снапшота — до первого нажатия enabled=false и наблюдатель ленив; identity-маппинг поверхности ДО инжекции не работал бы (живая находка ручной репликой — урок macr-per-app surface-C)
- **Fix:** acStartObserverRound — identity-несущая поверхность мапится после первого keystroke, стартующего ленивого наблюдателя
- **Files modified:** test/e2e/case_autocorrect.go
- **Verification:** живой fires-кейс зелёный (fired-запись в журнале, счётчики в status)
- **Committed in:** 46e336e

**3. [Rule 1 - Bug] Продуктовый фикс: pre-boundary ранний ответ убивал fired-коррекции**
- **Found during:** Task 3 (первый живой прогон fires на GTK4-фикстуре)
- **Issue:** граница срабатывает НА сепараторе; RequireSurroundingText обслуживается клиентом ДО доставки сепаратора — первый verify-ответ necessarily pre-boundary (токен без хвоста); раунд заключал verify-mismatch и коррекция умирала; 06-06-корпус моделировал недостижимую живьём пост-boundary форму (zenity выигрывал только таймингом кэшированного пуша)
- **Fix:** TDD RED `55dd248` (новый тест гоняет живой порядок: boundary → Require → pre-boundary пуш → раунд открыт, ноль операций, fired=1 abstained=0 → пост-boundary пуш закрывает) → GREEN `cbfa46c` (acBoundary-вооружённый раунд трактует пуш «токен без хвоста» как in-flight-separator ранний ответ, остаётся открыт в рамках существующего verify-бюджета; VerifyExpiry держит fail-closed границу; чужой ответ — по-прежнему mismatch (guard-тест); manual-пути байт-как-сегодня)
- **Files modified:** internal/session/actor.go, internal/session/actor_test.go
- **Verification:** RED падал на утверждении, GREEN — весь корпус зелёный под -race; живой fires-кейс зелёный на GTK4-фикстуре (ранее умиравшей поверхности)
- **Committed in:** 55dd248 (RED), cbfa46c (GREEN)

**4. [Rule 1 - Bug] Терминальный оракул re-pin: гонка доставки IME**
- **Found during:** Task 3 (живые прогоны 3-5 terminal-silent)
- **Issue:** план-оракул «ноль engine-записей от терминального typing» держался только в части прогонов: wezterm РЕГИСТРИРУЕТ IBus input context при фокусе (caps 0x9, нет surrounding-text бита) и его typing ГОНИТ гонку с armed-ингом IME — один прогон доставил все 7 keystrokes + границу через демон (abstained fail-closed), другой — ноль записей; оракул был ставкой на одну руку гонки
- **Fix:** 1dc76af — оракул пере-пинен на само свойство безопасности: fired=0 + ноль коррекционных записей + active-layer status; пришедшая граница отказывается fail-closed (D-53 условие 3, замкнутые слаги, юнит-спиненные); положительный контроль фокуса (наш zenity) сохранён
- **Files modified:** test/e2e/case_autocorrect.go
- **Verification:** два подряд зелёных живых прогона после re-pin
- **Committed in:** 1dc76af

**5. [Rule 3 - Blocking environmental] Полный зелёный matrix-v4 не достигнут на этой сессии — деферрал в 06-08**
- **Found during:** Task 3 (полные матричные прогоны)
- **Issue:** средовой session-drift свежеперезалогиненной сессии (2026-10-02): ПЕРВОЕ key-событие после флипа режима (flipTo = SetGlobalEngine) теряется на композиторном/ibus-слое — каждая строка, чьи шаги ставят typing/tapping сразу после флипа, падает с ровно одним потерянным ведущим символом или одной записью режима
- **Fix (решение чекпоинта, accept & defer 2026-10-02):** контрольный эксперимент — до-фазовое-6 дерево d96dcb2 (последнее живо-зелёное, 05-05) падает БАЙТ-ИДЕНТИЧНО (/tmp/matrix-v3-oldtree.log, worktree /tmp/goswitch-0505) — диф фазы 6 exonered, дрейф rides сессию; формальный гейт (D-48 fresh-session двойной прогон) маршрутизируется в план 06-08 FLAGGED-допущением самого плана; находка записана в deferred-items.md и WINDOWS-ледгер
- **Files modified:** .planning/phases/06-avtokorrekcija-opcionalno/deferred-items.md (запись), .planning/WINDOWS.md (ледгер-запись)
- **Verification:** контрольный эксперимент байт-идентичен; autocorrect-строки v4 зелёные ×2; три живых кейса зелёные
- **Committed in:** 2c41b83 (deferred-items), close-out (WINDOWS-ледгер)

---

**Total deviations:** 5 auto-fixed (3 Rule 1, 2 Rule 3). **Impact on plan:** № 3 — продуктовый баг, найденный и закрытый живьём (усилил 06-06-контур, manual-пути нетронуты); № 4 — честный re-pin оракула на свойство безопасности (не ослабление: старый оракул был ставкой на одну руку гонки); № 5 — средовое, с контрольным экспериментом, exonерирующим диф фазы; скоуп-крипа нет.

## Issues Encountered

- **Session-drift (главная):** см. отклонение № 5 и deferred-items.md — потеря первого нажатия после флипа; byte-identical на d96dcb2; также `super-space-alive` теряет пост-overview шаг фокуса (witness "(none)"). Кандидаты: стабилизация сессии/релогин владельца перед формальным ночным гейтом; stand-side bounded post-flip settle (прецедент 03-04 renegotiation-wait) как санкционированный quick-план; D-48 fresh-session прогон 06-08 может просто потребовать устоявшейся input-стека сессии
- **Флак TestRun_OnConnHookCalledOnce** воспроизвёлся 1× на close-out `mise run ci` (OnConn called 0 times) — изолированный ре-ран зелёный, полный ре-ран зелёный (0 FAIL). Root-cause (ассерт читает `calls` ДО `waitHookContext`) уже в deferred-items.md; 06-07 трогает только test/e2e — вне скоупа по boundary-правилу
- **Первые живые прогоны fires умирали на verify-mismatch** — не средовое, а настоящий продуктовый баг (отклонение № 3): фикс cbfa46c сделал fires-кейс зелёным на поверхности, где он не проходил никогда

## Known Stubs

None — все поверхности живые: фикстура, кейсы, матрица, ночной конвейер и mise-задачи wired; заглушек «not available/TODO» нет. Password-строка в v4 отсутствует по честной причине (нет password-шага у матричного драйвера), задокументировано в docs/ci-runner.md и решениях.

## User Setup Required

None - no external service configuration required. (Живые кейсы требуют живого GNOME-стола владельца с ydotool/input-группой — уже настроено фазами 1-5; goswitchd.service владельца восстановлен в active, проверено на close-out: `systemctl --user is-active goswitchd.service` → active.)

## Next Phase Readiness

- **06-08** получает: matrix-v4 как default ночного гейта (двойной свежесессионный D-48 прогон — формальный гейт критерия 5 по FLAGGED-допущению); счётчики-оракул и фикс verify-раунда в акторе; три живых кейса как автономные носители негативов. 06-08 судит session-drift на свежей сессии (verdict-владелец: 06-08 + verify-work)
- **Открытые пункты для 06-08/verify-work:** (1) session-drift — релогин/стабилизация сессии перед формальным прогоном или stand-side post-flip settle (WINDOWS-запись); (2) WINDOWS #12 (word-mixed/phrase-mixed vs flip-reset-семантика) остаётся открытым — v4 несёт строки дословно; (3) ADR-007 Accept — за 06-08 (этим планом НЕ отмечен)
- **Границы контракта:** ручной Double — по-прежнему переопределение (fires-кейс пинит на той же поверхности); off-состояние = ноль счётчиков; терминальный негатив = свойство безопасности при любой доставке

---
*Phase: 06-avtokorrekcija-opcionalno*
*Completed: 2026-10-02*

## Self-Check: PASSED

- Файлы: case_autocorrect.go, password_entry.py, input.html, matrix-v4.yaml, main.go, matrix.go, matrix_test.go, mise.toml, e2e-matrix.yml, d48-nightly-dispatch.sh, ci-runner.md, actor.go/actor_test.go, case_ctl.go — существуют на диске (14/14 FOUND)
- Коммиты: a1326cd, 46e336e, 55dd248, cbfa46c, 1dc76af, 2c41b83 — присутствуют в git log; measured commits от ledger gsd-plan-head-before-06-07: 6
- Гейт заморозки v3: `git diff --exit-code -- test/e2e/cases/matrix-v3.yaml` → V3-FROZEN-OK (пере-проверено на close-out)
- Зелёная итерация D-08: `mise run ci` на финальном дереве — первый прогон поймал известный флак TestRun_OnConnHookCalledOnce (изолированный ре-ран ok), полный ре-ран exit 0, 0 FAIL (/tmp/06-07-ci-final2.log)
- goswitchd.service владельца: active (восстановлен, не тронут этим планом)
