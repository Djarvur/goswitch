---
phase: 06-avtokorrekcija-opcionalno
plan: 08
subsystem: acceptance
tags: [autocorrect, perf, pitfall-5, matrix-v4, d-48, session-drift, adr-007, readme, windows-13, live-desktop]

# Dependency graph
requires:
  - phase: 06-avtokorrekcija-opcionalno (06-07)
    provides: matrix-v4 + ночной конвейер, счётчики-оракул, фикс verify-раунда (cbfa46c), WINDOWS #13 (session-drift деферрал с контрольным экспериментом d96dcb2)
  - phase: 06-avtokorrekcija-opcionalno (06-05/06-06)
    provides: числа golden-корпуса (190/0/34) и silence-matrix — вход в таблицу доказательств ADR-007
  - phase: 04-avtokorrekcija... (04-04/04-07)
    provides: perf-методика (N=40, окно ydotool→AT-SPI, VmHWM), базовая линия README, прецедент «budget-fail — санкционированный честный исход» (WINDOWS #5)
provides:
  - test/e2e/perf.go: runPerfAutocorrect — перф-вариант Pitfall-5 (тот же 40-повторный комбо-кейс, -config без -debug, единственный диф от дефолтов — autocorrect ON для живьём-наблюдаемой идентичности поверхности; общий perfMeasure-корпус); a completed measurement exits 0 — абсолютный бюджет-вердикт едет в отчёте
  - mise.toml: [tasks.e2e-perf-autocorrect] (вне ci) — воспроизводимая пара прогонов для дельты
  - Измеренная дельта p95: −3.3 мс (baseline 180.2 → autocorrect ON 176.9 мс, одна сессия) — гейт ≤ +5 мс ПРОЙДЕН; обе пары чисел в ADR-007
  - README: раздел «Autocorrect (v1.1, optional — default off)» — включение, white-list модель, приватность (слово никогда в логах/статусе), GTK3-оговорка, ручные жесты = переопределение; perf-таблица 04-07 нетронута
  - ADR-007: полная Acceptance Evidence таблица (6 строк по критериям фазы); Status честно ОСТАЁТСЯ Proposed — вердикт владельца на verify-work (двойная зелень v4 не получена — средовой дрейф)
  - WINDOWS #13 — формальный вердикт: дрейф CONFIRMED REAL и STICKY (двойной прогон 23/35 и 24/35 на сессии 2ч03м/2ч08м; переживает ibus restart — 24/35 тем же набором); диф фазы exonерирован трижды
affects: [verify-work фазы 06 (вердикты: ADR-007 Accept, WINDOWS #12/#13, criterion-5), new-milestone v1.1.0 (REQ-IDs), возможный quick-план «stand-side post-flip settle»]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
# Same estimateTokens scale (chars/4 over the realized diff), never a harness token count.
actuals:
  tokens: 6548       # chars/4 over the realized diff (26 190 bytes, 2 task-коммита) — оценка плана 48 000 сильно выше дифа: львиная доля стоимости плана — живые прогоны (перф ×2 по 40 повторов, матрица ×3), не диф
  tasks: 2
  commits: 2         # MEASURED: git rev-list --count gsd-plan-head-before-06-08..HEAD (c004134..e27b9bd); docs-коммиты SUMMARY/STATE — следующим за измерением

# Tech tracking
tech-stack:
  added: []          # ноль новых зависимостей (go.mod не тронут; tidy-diff зелёный)
  patterns:
    - "Честное перф-сравнение фичи: общий perfMeasure-корпус, единственная переменная — состояние фичи в -config (байт-зеркало дефолтов, flip_after_correction как в базовой линии); идентичность поверхности наблюдается живьём (case-side appid.Observer + пробная zenity-итерация), никогда не хардкодится"
    - "Exit-семантика перф-варианта: завершённое измерение = exit 0 (абсолютный бюджет-вердикт едет в отчёте — санкционированный исход 04-07); сентиел errPerfBudget отделяет бюджет-фейл от упавшего прогона"
    - "Пост-гейт диагностика отдельно от формального гейта: двойной прогон D-48 без вмешательства, ЗАТЕМ документированное средство (ibus restart) одним маркированным прогоном — вердикт «переживает средство», не подмена гейта"

key-files:
  created:
    - .planning/phases/06-avtokorrekcija-opcionalno/06-08-SUMMARY.md
  modified:
    - test/e2e/perf.go
    - test/e2e/main.go
    - mise.toml
    - README.md
    - docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md
    - .planning/WINDOWS.md
    - .planning/phases/06-avtokorrekcija-opcionalno/deferred-items.md

key-decisions:
  - "ADR-007 ОСТАЁТСЯ Proposed: T-06-08-01 («двойная зелень v4 — обязательное условие Accepted; красный прогон не пропускается») не выполнен — исполнитель НЕ решает вопрос гейта владельца (прецедент WINDOWS #5); вместо флипа — полная Acceptance Evidence таблица и честный Status с маршрутизацией вердикта на verify-work"
  - "Session-drift вердикт 06-08: CONFIRMED REAL и STICKY — двойной прогон на сессии 2ч+ (не «молодая сессия») дал те же 9 строк дрейфа байт-схожими формами; диагностика документированным средством (ibus restart) дрейф НЕ сняла — состояние живёт глубже ibus-демона в логине; формальный ≤30-мин fresh_session гейт остаётся машинным инструментом владельца"
  - "Дельта-методика Pitfall 5: базовая линия снята ЗАНОВО в той же сессии (mise run e2e-perf, exit 1 санкционирован), пара чисел baseline/ac в ADR-007; дельта p95 −3.3 мс — в пределах шума прогона, гейт ≤ +5 мс пройден; README-таблица 04-07 не тронута (план прямо запрещает)"
  - "Двойной регресс судится по классам строк: autocorrect-строки и словесно/фразовые носители ручных контрактов зелёные ×2; красные — 9 строк дрейфа (exonерированы) + 2 дизайн-красные строки WINDOWS #12 (замороженная семантика сохранена) + 1 транзиент ibus-активации; слово «зелёный» к прогону 23/35 не применяется НИКОГДА"

patterns-established:
  - "Перф-пара одной сессии: baseline и вариант снимаются подряд (~4 мин) — шума сессии между прогонами нет; свежая baseline обязательна, если README-числа из другой сессии"
  - "Класс-строчный вердикт матрицы: drift-строки / замороженные строки / несущие строки считаются раздельно — полный exit-код не единственный оракул при открытых WINDOWS-вердиктах владельца"

requirements-completed: ["CORR-01", "CORR-02", "CORR-03", "CORR-04", "CORR-05", "CORR-06", "CORR-07", "CORR-08", "CORR-09", "SPEC §10/§11 (spec-delta)"]  # копия frontmatter плана дословно; формальная фиксация REQ-ID в REQUIREMENTS.md — при new-milestone v1.1.0 (D-51-дисциплина фазы, прецедент 06-06/06-07: REQUIREMENTS.md планом не трогается, verify-гейт заморозки это пинит)

coverage:
  - id: D1
    description: "Perf-гейт Pitfall 5: перф-вариант с включённой автокоррекцией измерен живьём (та же сессия, N=40, жест не менялся); дельта p95 −3.3 мс ≤ +5 мс; обе пары p50/p95/p99 и вердикт в ADR-007; mise-задача e2e-perf-autocorrect существует и вне ci"
    verification:
      - kind: e2e
        ref: "mise run e2e-perf → exit 1 (санкционированный бюджет-гейт), отчёт p50/p95/p99 169.3/180.2/180.4 мс; mise run e2e-perf-autocorrect → PASS exit 0, 169.6/176.9/177.5 мс; VmHWM 19.1 МБ < 50 МБ (/tmp/06-08-perf-baseline-report.txt, /tmp/06-08-perf-ac-report.txt)"
        status: pass
      - kind: other
        ref: "verify-гейты: go build/vet ./test/e2e/ + grep tasks.e2e-perf-autocorrect mise.toml + grep p95 ADR-007 → PERF-TASK-OK/PERF-RECORDED"
        status: pass
    human_judgment: false
  - id: D2
    description: "Формальный двойной зелёный регресс matrix-v4 (критерий 5, D-48-форма) — НЕ ЗАКРЫТ: 23/35 и 24/35 на сессии 2ч03м/2ч08м; 9 строк session-drift (байт-схожи с 06-07 и с контролем d96dcb2) + 2 замороженные WINDOWS #12 строки + 1 ibus-транзиент; autocorrect-строки и словесно/фразовые носители ручных контрактов зелёные ×2; диагностика: дрейф переживает ibus restart"
    requirement: "SPEC §10/§11 (spec-delta)"
    verification:
      - kind: e2e
        ref: "mise run e2e-matrix-v4 ×2 подряд (/tmp/06-08-matrix-double.log: RUN 1 exit=1 23/35, RUN 2 exit=1 24/35) + диагностика после ibus restart (24/35, /tmp/06-08-matrix-post-ibus-restart.log)"
        status: fail
    human_judgment: true
    rationale: "Средовой дрейф стола — не код фазы (три независимые exonерации: до-фазовое дерево d96dcb2 байт-идентично, тот же набор при зелёных autocorrect-строках в тех же прогонах, переживает ibus restart); вердикт «принять с зафиксированным средовым риском / чинить стол / гнать формальный ≤30-мин ночной гейт» — владельца на verify-work (план прямо предвидел этот исход: «не fake green, не blind-click»)"
  - id: D3
    description: "README: раздел автокоррекции — default off, как включить (секция autocorrect + ссылка docs/CONFIG.md), white-list точных имён, приватность (слово никогда в логах/статусе), GTK3-оговорка с рекомендацией не вносить недоверенные приложения, ручные жесты = переопределение; perf-таблица 04-07 без изменений"
    verification:
      - kind: other
        ref: "греп-гейты verify: 'default off' + gtk3 + autocorrect в README → FINALIZE-GREPS-OK; git diff показывает раздел единственным README-изменением"
        status: pass
    human_judgment: true
    rationale: "Честность пользовательской документации (полнота оговорок, ясность white-list модели) — суждение владельца на UAT/verify-work; грепы пинят наличие, не качество"
  - id: D4
    description: "ADR-007: Acceptance Evidence таблица (корпус 190/0/34, живые кейсы 06-07, счётчики D-54, носители ручных контрактов ×2, perf-пара, строка 5 — честный красный с разбивкой); Status остаётся Proposed с маршрутизацией вердикта владельцу"
    requirement: "SPEC §10/§11 (spec-delta)"
    verification:
      - kind: other
        ref: "греп-гейт 'matrix-v4' в ADR-007 → FINALIZE-GREPS-OK; Status-секция читается 'Proposed' с датой 06-08 и маршрутом verify-work"
        status: pass
    human_judgment: true
    rationale: "Перевод в Accepted — сам гейт владельца (T-06-08-01); исполнитель собирает доказательства и НЕ флипает статус — это и есть санкционированный исход плана при красном двойном прогоне"
  - id: D5
    description: "WINDOWS #13 закрыт вердиктом (статус open — за владельцем): формальный двойной прогон исполнен, дрейф подтверждён реальным и живучим, exonерации задокументированы, диагностический прогон после ibus restart добавлен; deferred-items.md дополнен 06-08-записью с кандидатами лечения"
    verification:
      - kind: other
        ref: ".planning/WINDOWS.md запись #13 (таблица + JSON) и deferred-items.md секция 06-08 — обновлены этим планом"
        status: pass
    human_judgment: true
    rationale: "Закрытие записи (fixed/waive) — решение владельца на verify-work: средовой дефект стола вне досягаемости репозитория"
  - id: D6
    description: "Милстоун-дисциплина D-51: git diff пуст на .planning/PROJECT.md/.planning/STATE.md/.planning/REQUIREMENTS.md/release.yml/.goreleaser.yaml на закрытии задач; матрицы v3/v4 и оракулы не ослаблены"
    verification:
      - kind: other
        ref: "git diff --exit-code (все пять путей) → MILESTONE-FROZEN; git diff --exit-code на matrix-v3.yaml/v4.yaml → CASES-FROZEN-OK"
        status: pass
    human_judgment: false
  - id: D7
    description: "Зелёная итерация D-08: mise run ci (build+vet+lint strict+test -race) зелёный на закрытии каждой задачи (5 прогонов за план, включая финальный; известный флак TestRun_OnConnHookCalledOnce не воспроизводился)"
    verification:
      - kind: other
        ref: "mise run ci exit 0 ×5 (последний — финальное дерево, 2026-10-02T18:5xZ)"
        status: pass
    human_judgment: false

# Metrics
duration: 41min
completed: 2026-10-02
status: complete
---

# Phase 6 Plan 8: Приёмка фазы — перф-гейт, двойной регресс, ADR/README Summary

**Perf-гейт Pitfall 5 пройден живьём (дельта p95 −3.3 мс ≤ +5 мс); формальный двойной прогон matrix-v4 честно красный на подтверждённом живучем средовом дрейфе стола (переживает ibus restart; диф фазы exonерирован трижды) — ADR-007 получил полную таблицу доказательств, но Status остаётся Proposed: флип — вердикт владельца на verify-work; README получил честный раздел автокоррекции (default off, white-list, приватность, GTK3-оговорка).**

## Performance

- **Duration:** 41 min (18:18–18:59 UTC)
- **Started:** 2026-10-02T18:18:04Z
- **Completed:** 2026-10-02T18:59Z
- **Tasks:** 2
- **Files modified:** 7 (+1 SUMMARY)

## Accomplishments

- **Perf-гейт (Pitfall 5) — пройден:** новый `runPerfAutocorrect` измеряет тот же 40-повторный комбо-кейс на -config-демоне, чей единственный диф от встроенных дефолтов — секция autocorrect ON для живьём-наблюдаемой идентичности поверхности (`org.gnome.Zenity`, case-side наблюдатель + пробная zenity-итерация). Одна сессия, прогоны подряд: baseline p50/p95/p99 169.3/180.2/180.4 мс → autocorrect ON 169.6/176.9/177.5 мс; **дельта p95 −3.3 мс, гейт ≤ +5 мс ПРОЙДЕН**; VmHWM 19.1 МБ. Абсолютный бюджет-фейл окна — санкционированный честный исход 04-07 (WINDOWS #5). Обе пары чисел — в ADR-007 Consequences.
- **Двойной регресс v4 — исполнен по форме D-48, вердикт честный:** два последовательных `mise run e2e-matrix-v4` без вмешательства — 23/35 и 24/35. Разбивка по классам: **autocorrect-строки (fires/off/oov) зелёные ×2**; **словесно/фразовые носители ручных контрактов зелёные ×2** (flip-after-word-correction, combo-word-layout, gedit/x11-word-*, регистровые, фразовые); красные — 9 строк session-drift (первая клавиша после флипа теряется; байт-схожие формы с 06-07 и с контролем до-фазового дерева d96dcb2) + 2 замороженные строки WINDOWS #12 (дизайн-красные, семантика сохранена) + 1 транзиент ibus-активации в прогоне 1 (reload-window, зелёный в прогоне 2).
- **Session-drift — формальный вердикт 06-08: CONFIRMED REAL и STICKY.** Дрейф persisted 2ч03м/2ч08м после перелогина (гипотеза «молодой сессии» фальсифицирована) и **переживает документированное средство `ibus restart`** (маркированный пост-гейт диагностический прогон: 24/35, идентичный набор строк). Третья независимая exonерация дифа фазы: в тех же прогонах зелёные autocorrect-строки и носители ручных контрактов. Полная запись — WINDOWS #13 + deferred-items + ADR-007 строка 5.
- **ADR-007 — полная Acceptance Evidence таблица** (6 строк по критериям фазы: корпус 190/0/34, живые кейсы, счётчики, ручные контракты ×2, perf-пара, строка 5 честно красная); **Status остаётся Proposed** — исполнитель не флипает гейт владельца (T-06-08-01, прецедент WINDOWS #5): вердикт «Accepted с зафиксированным средовым риском / лечение стола / формальный ночной гейт» — на verify-work.
- **README — раздел «Autocorrect (v1.1, optional — default off)»:** что делает, как включить (секция autocorrect → docs/CONFIG.md), white-list точных имён («не в списке — никогда»), приватность (слово никогда в логах/статусе — счётчики и причины; -debug-трейс оговорён отдельно), GTK3-оговорка (роль неотличима; не добавляйте недоверенные приложения), ручные жесты = переопределение. Perf-таблица 04-07 не тронута.
- **D-51-дисциплина:** MILESTONE-FROZEN (git diff пуст на PROJECT/STATE/REQUIREMENTS/release.yml/.goreleaser.yaml на закрытии задач); CASES-FROZEN-OK (v3 и v4 не тронуты); оракулы не ослаблены; mise run ci зелёный ×5 за план (флак OnConnHookCalledOnce не воспроизводился).

### Живые результаты

| Гейт | Результат | Доказательство |
|---|---|---|
| e2e-perf (baseline, свежая той же сессии) | exit 1 — санкционированный бюджет-гейт | p50/p95/p99 169.3/180.2/180.4 мс, VmHWM 19.8 МБ (/tmp/06-08-perf-baseline-report.txt) |
| e2e-perf-autocorrect | **PASS** exit 0 | 169.6/176.9/177.5 мс, VmHWM 19.1 МБ; идентичность наблюдена живьём (`org.gnome.Zenity`); дельта p95 **−3.3 мс** → гейт ПРОЙДЕН (/tmp/06-08-perf-ac-report.txt) |
| e2e-matrix-v4 ×2 (форма D-48) | **честно красный** — 23/35 и 24/35 | autocorrect-строки и носители ручных контрактов зелёные ×2; 9 drift-строк + 2 замороженные #12 + 1 ibus-транзиент (/tmp/06-08-matrix-double.log, /tmp/06-08-run1-rows.txt, run2) |
| диагностика: matrix-v4 после `ibus restart` | 24/35, тот же набор строк | дрейф переживает документированное средство — живёт в состоянии логина глубже ibus (/tmp/06-08-matrix-post-ibus-restart.log) |
| mise run ci | **зелёный ×5** | финальный — на закрытии плана |

## Task Commits

Each task was committed atomically:

1. **Task 1: Perf-гейт — p95 с включённой автокоррекцией против базовой линии** — `a1c8392` (feat: perf-autocorrect вариант + mise-задача + ADR-007 perf-запись; живые прогоны до коммита)
2. **Task 2: Двойной регресс v4 + README + ADR-007 evidence** — `e27b9bd` (docs: README-раздел + Acceptance Evidence таблица; живые прогоны до коммита)

**Plan metadata:** _docs(06-08) — SUMMARY и STATE/ROADMAP/WINDOWS — следующими коммитами (прецедент 06-06/06-07)_

## Files Created/Modified

- `test/e2e/perf.go` — runPerfAutocorrect + observePerfSurfaceIdentity (живое наблюдение идентичности) + perfMeasure (общий корпус) + perfAutocorrectConfigTmpl (байт-зеркало дефолтов) + errPerfBudget сентиел; writePerfReport несёт строку daemon-формы
- `test/e2e/main.go` — startDaemonPlainArgs (prod-форма + -config, без -debug), регистрация/usage perf-autocorrect
- `mise.toml` — [tasks.e2e-perf-autocorrect] (вне [tasks.ci])
- `README.md` — раздел «Autocorrect (v1.1, optional — default off)»
- `docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md` — perf-измерение в Consequences; Acceptance Evidence таблица; Status-маршрутизация на verify-work
- `.planning/WINDOWS.md` — #13 дополнен вердиктом 06-08 (статус open — за владельцем)
- `.planning/phases/06-avtokorrekcija-opcionalno/deferred-items.md` — 06-08-запись: вердикт о дрейфе + кандидаты лечения; флак-нота

## Decisions Made

См. key-decisions: ADR остаётся Proposed (гейт владельца), drift-вердикт REAL и STICKY (переживает ibus restart), свежая baseline в той же сессии, класс-строчный суд матрицы, слово «зелёный» к красному прогону не применяется.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking lint] nolintlint/noctx перестановка при рефакторинге стартов демона**
- **Found during:** Task 1 (GREEN — mise run ci)
- **Issue:** startDaemonPlain стал обёрткой над startDaemonPlainArgs — его //nolint:noctx стал неиспользуемым (nolintlint), а noctx-диагностика переехала в новую функцию с exec.Command
- **Fix:** директива перенесена на startDaemonPlainArgs; обёртка — без директивы
- **Files modified:** test/e2e/main.go
- **Verification:** mise run ci зелёный
- **Committed in:** a1c8392

**2. [Rule 1 - Bug] Неточная формулировка «все носители ручных контрактов зелёные» в ADR-строке 4**
- **Found during:** Task 2 (составление таблицы доказательств)
- **Issue:** select-строки — тоже носители CORR-03, и они красные (с drift-подписью); заявление «все» было бы неточным
- **Fix:** строка 4 уточнена: словесно/фразовые носители зелёные ×2; select-носители — в зоне дрейфа строки 5 с той же подписью (потерянная первая клавиша, не select-механика); юнит-носители CORR-03 зелёные
- **Files modified:** docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md
- **Verification:** формулировка сверена с /tmp/06-08-run1-rows.txt / run2
- **Committed in:** e27b9bd

**3. [Rule 3 - Blocking environmental] Формальная двойная зелень v4 не достигнута — честный вердикт вместо флипа ADR**
- **Found during:** Task 2 (живые прогоны)
- **Issue:** verify-гейт «оба exit 0» невыполним: 23/35 и 24/35 на средовом дрейфе (подтверждён живучим); must_haves truth №2 не получена; флип ADR в Accepted при красном прогоне прямо запрещён планом (T-06-08-01)
- **Fix:** план предвидел этот исход («report it as a REAL finding with logs — do not fake green, do not blind-click»): полный вердикт по классам строк в ADR-007 строке 5 и WINDOWS #13; Status остаётся Proposed с маршрутом verify-work; диагностика (ibus restart) выполнена ОТДЕЛЬНО от формального гейта маркированным прогоном
- **Files modified:** docs/adr/ADR-007..., .planning/WINDOWS.md, .planning/.../deferred-items.md
- **Verification:** /tmp/06-08-matrix-double.log (два прогона), /tmp/06-08-matrix-post-ibus-restart.log (диагностика); логи сохранены
- **Committed in:** e27b9bd (ADR), metadata-коммит (WINDOWS/deferred-items)

---

**Total deviations:** 3 auto-fixed (1 Rule 1, 2 Rule 3). **Impact on plan:** № 3 — санкционированный планом честный исход (средовой, с тремя exonерациями и диагностикой); скоуп-крипа нет; гейт-грепы README/ADR/заморозок зелёные.

## Issues Encountered

- **Session drift (главная, подтверждена):** см. отклонение № 3, WINDOWS #13, deferred-items. Ключевое новое знание против 06-07: дрейф НЕ «молодо-сессионный» (2ч+ после перелогина) и НЕ лечится `ibus restart` — состояние глубже ibus-демона в логине. Кандидаты владельцу: формальный ≤30-мин fresh_session ночной гейт на verify-work; stand-side bounded post-flip settle (03-04 прецедент) как quick-план; расследование изменений десктоп-стека с зелёных прогонов фазы 5.
- Один новый транзиент в прогоне 1: `reload-window` — `SetGlobalEngine: Operation was cancelled` (ibus-уровень); прогон 2 зелёный по строке. Не воспроизводился далее.
- Стенд-дисциплина соблюдена: goswitchd.service останавливался на живые прогоны и восстановлен active (дважды — после перфа и после матрицы); осиротевшего IBus-имени не возникало (префлайты 7/7 зелёные на всех прогонах).

## Known Stubs

None — все артефакты живые: перф-вариант прогоняется живьём, mise-задача wired, README/ADR полные, WINDOWS/deferred-items обновлены. Заглушек «not available/TODO» нет.

## User Setup Required

None - no external service configuration required. (goswitchd.service владельца восстановлен в active после живых прогонов: `systemctl --user is-active goswitchd.service` → active, движок goswitch-en активен.)

## Next Phase Readiness

- **verify-work фазы 6** получает полный пакет: perf-гейт пройден и записан; критерий 1 (корпус) закрыт 06-05; критерий 2/3 (политика/безопасность) — 06-06/06-07 живьём; критерий 4 (ручное переопределение) — зелёные носители ×2 + юнит-пины; критерий 5 — ЧЕСТНО красный по средовой причине с тремя exonерациями и диагностикой. Вердикты владельца: ADR-007 Accept (строка 5 — принять риск/лечить стол/гнать формальный ночной гейт), WINDOWS #12 (две замороженные строки), WINDOWS #13 (дрейф), WINDOWS #11 (индикатор — из фазы 5).
- **Формальный ≤30-мин гейт** (docs/ci-runner.md, `fresh_session=true`, d48-nightly-dispatch.sh) — машинный инструмент; 06-08 его НЕ подменял (сессия 2ч+ — префлайт отказал бы честно).
- **Кандидат quick-плана:** stand-side bounded post-flip settle перед следующим шагом строки (прецедент 03-04 renegotiation-wait) — лечил бы симптом предсказуемо; решение владельца.
- **REQ-IDs** (CORR-01..09, SPEC §10/§11) — формальная фиксация в REQUIREMENTS.md при new-milestone v1.1.0 (D-51, прецедент 06-06/06-07).

---
*Phase: 06-avtokorrekcija-opcionalno*
*Completed: 2026-10-02*

## Self-Check: PASSED

- Файлы: test/e2e/perf.go, test/e2e/main.go, mise.toml, README.md, docs/adr/ADR-007..., .planning/WINDOWS.md, deferred-items.md — существуют на диске (7/7 FOUND)
- Коммиты: a1c8392 (feat), e27b9bd (docs) — присутствуют в git log; measured commits от ledger gsd-plan-head-before-06-08 (c004134): 2 (docs-коммиты SUMMARY/STATE — следующим за измерением, прецедент 06-06/06-07)
- Verify-гейты плана: PERF-TASK-OK; PERF-LIVE-DONE (e2e-perf-autocorrect PASS exit 0); PERF-RECORDED (p95 в ADR); FINALIZE-GREPS-OK (README default off/GTK3/autocorrect + ADR matrix-v4); MILESTONE-FROZEN; CASES-FROZEN-OK; mise run ci exit 0 ×5
- Гейт «оба exit 0» двойного прогона: ЧЕСТНО НЕ ВЫПОЛНЕН (23/35, 24/35) — задокументирован как санкционированный средовой исход с полным разбором; слово «зелёный» к прогонам не применяется
- Живой стол восстановлен: goswitchd.service active, движок goswitch-en
