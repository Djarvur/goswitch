---
phase: 06-avtokorrekcija-opcionalno
plan: 02
subsystem: layouts
tags: [dictgen, hunspell, golden-generator, trigrams, sorted-slice, rodata, licensing, tdd]

# Dependency graph
requires:
  - phase: 06-avtokorrekcija-opcionalno (06-01)
    provides: ADR-007 (Proposed) — лицензионная позиция доноров и kill-таблица; SPEC §10/§11 — решение D-51/D-55
  - phase: 01 (layouts/generator)
    provides: золотой паттерн генератора (committed golden, CI гоняет только goldens — CORR-08), форма заголовков и эмиссии
provides:
  - layouts/dictgen — dev-генератор: парсер hunspell .dic (снять флаги, len≥2, дедуп, ё→е, громкие отказы), тренировка триграмм (top-4096/язык, log10, детерминированный tie-break), байт-детерминированная эмиссия через go/format
  - Закоммиченные golden-данные: DictRU 138 914 слов / DictEN 78 951 (отсортированные []string-литералы, rodata) + TriRU/TriEN (map[string]float64 log-вероятности, ≤4096/язык)
  - Golden-инварианты layouts/dict_test.go: сортированность, дизъюнктность сценариев, отсутствие ё, конверты счётчиков, потолок триграмм — пинят данные независимо от генератора
  - docs/LICENSE-data.md — дословные нотисы Lebedev (custom-bsd-4-clauses) и SCOWL/Kevin Atkinson + происхождение/дата запекания
  - [tasks.dictgen-regen] в mise.toml (dev-only, вне [tasks.ci])
affects: [06-05 detect (потребитель DictRU/DictEN/TriRU/TriEN), 06-08 acceptance (полный регресс), будущие регенерации словарей при обновлении пакета hunspell]

# Actuals (#2632)
actuals:
  tokens: 1274105        # chars/4 over the realized diff (5 096 422 bytes) — оценка плана (55 000) не учитывала запекание мегабайт словарных литералов
  tasks: 2
  commits: 4              # MEASURED: git rev-list --count gsd-plan-head-before-06-02..HEAD

# Tech tracking
tech-stack:
  added: []               # ноль новых Go-зависимостей — только stdlib (constraint соблюдён, tidy-diff зелёный)
  patterns:
    - "Золотой генератор словарных данных: hunspell → committed golden, CI goldens only (расширение CORR-08 на данные детектора)"
    - "Отсортированные []string-литералы в rodata вместо рантайм-map (−10–12 МБ RSS, 06-RESEARCH Q3); поиск — sort.SearchStrings (план 06-05)"
    - "Независимый пин потолка в тесте: константа maxTrigramsPerLang продублирована в dict_test.go — потеря потолка в генераторе громко валит тест"

key-files:
  created:
    - layouts/dictgen/main.go
    - layouts/dictgen/main_test.go
    - layouts/dictgen/testdata/ru_fixture.dic
    - layouts/dictgen/testdata/en_fixture.dic
    - layouts/dictgen/testdata/ru_bad_count.dic
    - layouts/dict_ru.go
    - layouts/dict_en.go
    - layouts/trigrams.go
    - layouts/dict_test.go
    - docs/LICENSE-data.md
  modified:
    - mise.toml

key-decisions:
  - "Конверт ru-счётчика скорректирован по факту (оговорка плана «уточнить границы по факту Task 1»): 146 261 из research Q3 измерен ДО ё→е-фолдинга; фолдинг (locked, Pitfall 3/R3) сливает 7347 е/ё-пар лемм → 138 914; границы [137000, 140500] + комментарий с полной декомпозицией"
  - "Тесты генератора — белый ящик package main (прецедент engine/conn_*, internal/indicator): внешний package main_test невозможен против неэкспортированного пайплайна (main-пакет неимпортируем, проверено эмпирически на go1.23)"
  - "CI-независимость от hunspell доказана живьём: bwrap с пустым tmpfs на /usr/share/hunspell → все тесты зелёные, TestGenerateLive скипается по os.Stat-гейту (золотой контракт CORR-08)"
  - "Первичная генерация требует ручного bootstrap (cd layouts && go run ./dictgen), т.к. //go:generate-директивы живут в самих генерируемых файлах; далее go generate ./layouts самоподдерживается"

patterns-established:
  - "dictgen-заголовок: DO NOT EDIT + //go:generate go run ./dictgen + лицензионный комментарий per-language; в dict_ru.go ноль ё-глифов (греп-гейт)"
  - "Громкие отказы на аномалиях данных: errBadCountLine/errCountMismatch/errEmptyDict/errUnsortedWords/errUnknownName — ни одного тихого пропуска (T-06-02-04)"

requirements-completed: ["CORR-08", "SPEC §10/§11 (spec-delta)"]  # CORR-08 — уже Complete с Фазы 1 (xkb-таблицы); план расширяет его golden-контракт на словарные данные; второй элемент — свободная метка фазы, формальная фиксация при new-milestone (D-51)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "dictgen — генератор: parseDIC (флаги/len≥2/дедуп/ё→е/громкие отказы), trainTrigrams (окна по 3 руны, log10, top-K с алфавитным tie-break), emit (gofmt-чистые DO NOT EDIT-исходники)"
    requirement: "CORR-08"
    verification:
      - kind: unit
        ref: "layouts/dictgen/main_test.go#TestParseDIC (+TestTrigramTrain, TestEmitDeterministic, TestEmitRefusals) — corpus RED→GREEN"
        status: pass
    human_judgment: false
  - id: D2
    description: "DictRU/DictEN запечены как отсортированные []string-литералы (не map), ё→е свёрнут, сценарии дизъюнктны, счётчики 138 914/78 951 в конвертах"
    requirement: "CORR-08"
    verification:
      - kind: unit
        ref: "layouts/dict_test.go#TestGolden_DictSorted + TestGolden_DictDisjoint + TestGolden_DictNoYo + TestGolden_DictEnvelope"
        status: pass
    human_judgment: false
  - id: D3
    description: "TriRU/TriEN — map[string]float64 log-вероятности, непустые, ≤4096/язык, все значения ≤0; потолок пинится независимо от генератора"
    requirement: "SPEC §10/§11 (spec-delta)"
    verification:
      - kind: unit
        ref: "layouts/dict_test.go#TestGolden_TrigramsBounded (+mutation RED: потолок 4096→8192 + регенерация → тест красный, RED_EVIDENCE_OK)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Регенерация байт-в-байт детерминирована: go generate ./layouts не меняет закоммиченные golden-файлы"
    requirement: "CORR-08"
    verification:
      - kind: other
        ref: "mise run dictgen-regen && git diff --exit-code -- layouts/ → REGEN-DETERMINISTIC"
        status: pass
    human_judgment: false
  - id: D5
    description: "Лицензионная атрибуция: docs/LICENSE-data.md с дословными нотисами (Lebedev custom-bsd-4-clauses, SCOWL/Kevin Atkinson, PD-компоненты) + лицензионные заголовки в сгенерированных файлах"
    requirement: "SPEC §10/§11 (spec-delta)"
    verification:
      - kind: other
        ref: "grep Lebedev/custom-bsd + SCOWL docs/LICENSE-data.md; GENERATED-OK гейт плана — pass"
        status: pass
    human_judgment: true
    rationale: "Достаточность атрибуции как юридическое суждение — владелец на гейте verify-work фазы (позиция ADR-007, перевод в Accepted — план 06-08)"
  - id: D6
    description: "[tasks.dictgen-regen] существует и НЕ входит в mise run ci (золотой контракт: CI не требует hunspell)"
    requirement: "CORR-08"
    verification:
      - kind: other
        ref: "grep-гейт MISE-TASK-OK: tasks.dictgen-regen в mise.toml, отсутствует в блоке [tasks.ci]"
        status: pass
    human_judgment: false
  - id: D7
    description: "CI-независимость от hunspell доказана исполнением: словари скрыты (пустой tmpfs) → все тесты зелёные, живой тест скипается"
    requirement: "CORR-08"
    verification:
      - kind: other
        ref: "bwrap --dev-bind / / --tmpfs /usr/share/hunspell — go test ./layouts/... ok/ok, 1 SKIP (TestGenerateLive)"
        status: pass
    human_judgment: false
  - id: D8
    description: "Регресс существующих таблиц нетронут: TestGolden_SpecExamples / TestGolden_TableSize зелёные без правок (CORR-08 xkb-контракт)"
    requirement: "CORR-08"
    verification:
      - kind: unit
        ref: "layouts/tables_test.go#TestGolden_SpecExamples + TestGolden_TableSize — green under -race"
        status: pass
    human_judgment: false

# Metrics
duration: 38min
completed: 2026-10-01
status: complete
---

# Phase 6 Plan 2: dictgen — golden-словари и триграммы Summary

**Dev-генератор layouts/dictgen запекает полные hunspell-словари (DictRU 138 914 / DictEN 78 951, отсортированные []string в rodata, ё→е свёрнут) и триграммные таблицы TriRU/TriEN (log10, ≤4096/язык) в закоммиченные golden-файлы с дословными лицензионными нотисами; CI доказанно не требует hunspell.**

## Performance

- **Duration:** 38 min
- **Started:** 2026-10-01T20:19:32Z
- **Completed:** 2026-10-01T20:57:40Z
- **Tasks:** 2 (каждая — полный RED→GREEN; REFACTOR не потребовался)
- **Files modified:** 11

## Accomplishments
- layouts/dictgen: парсер hunspell .dic со громкими отказами на всех аномалиях (несогласованный счётчик, битая count-строка, пустой результат, нечитаемый файл), снятие affix-флагов, фильтр len≥2, дедупликация, ё→е-фолдинг, сортировка; тренировка триграмм (окна по 3 руны, log10(count/total), top-4096 по частоте с алфавитным tie-break); эмиссия gofmt-чистых DO NOT EDIT-исходников — отсортированные []string-литералы (rodata), НЕ map
- Golden-данные закоммичены: dict_ru.go (3.6 МБ, 138 914 слов, ноль ё-глифов во всём файле), dict_en.go (1.0 МБ, 78 951), trigrams.go (244 КБ, 2×4096 записей); docs/LICENSE-data.md с дословными нотисами обоих источников
- Golden-инварианты пинят данные независимо от генератора: сортированность, дизъюнкт сценариев, отсутствие ё, конверты счётчиков, потолок 4096; детерминизм регенерации проверен (mise run dictgen-regen && git diff --exit-code — пусто)
- CI-независимость от hunspell доказана живьём (bwrap, пустой /usr/share/hunspell → зелёные тесты + SKIP живого теста); mise run ci зелёная на каждом закрытии задачи (D-08)

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: corpus против заглушек** - `186652e` (test) — RED_EVIDENCE_OK (TestParseDIC упал на утверждении, 5/5 корпуса красные)
2. **Task 1 GREEN: dictgen + golden-данные + LICENSE-data** - `d7e403a` (feat)
3. **Task 2 RED: golden-инварианты** - `5258857` (test) — корпус зелёный на валидных данных (ожидаемо планом); честный RED пинен mutation-прогоном (потолок 4096→8192 + регенерация → TestGolden_TrigramsBounded красный, RED_EVIDENCE_OK), мутация откачена байт-чисто
4. **Task 2 GREEN: mise dictgen-regen** - `6455c96` (feat)

**Plan metadata:** (см. финальный docs-коммит)

_Note: TDD-задачи; REFACTOR-коммиты не создавались — после GREEN правок, стоящих коммита, не было (lint 0 issues)._

## TDD Gate Compliance

| Plan | RED | GREEN | REFACTOR | Status |
|------|-----|-------|----------|--------|
| 06-02 Task 1 | ✓ `186652e` | ✓ `d7e403a` | — (не требовался) | Pass |
| 06-02 Task 2 | ✓ `5258857` (mutation-pinned, планом предписано) | ✓ `6455c96` | — (не требовался) | Pass |

Оба RED-корпуса прошли `check tdd-red-evidence` с вердиктом RED_EVIDENCE_OK (целевой тест падал на утверждении запланированного поведения, не на компиляции/фикстуре).

## Files Created/Modified
- `layouts/dictgen/main.go` — команда dictgen: parseDIC, trainTrigrams, emitSortedSlice/emitTrigrams( File), maxTrigramsPerLang=4096, входы /usr/share/hunspell/{ru_RU,en_US}.dic
- `layouts/dictgen/main_test.go` — белый ящик корпус: фикстуры, громкие отказы, детерминизм, живой конверт-тест с os.Stat-скипом
- `layouts/dictgen/testdata/{ru_fixture,en_fixture,ru_bad_count}.dic` — фикстуры парсера (флаги, дубли, 1-буквенные, ё-формы, несогласованный счётчик)
- `layouts/dict_ru.go` / `layouts/dict_en.go` / `layouts/trigrams.go` — сгенерированные golden-данные (committed)
- `layouts/dict_test.go` — golden-инварианты данных
- `docs/LICENSE-data.md` — атрибуция данных
- `mise.toml` — [tasks.dictgen-regen] (вне ci)

## Decisions Made
- Конверт ru-счётчика исправлен по факту с полной декомпозицией в комментарии теста: 146 269 записей − 8 коротких = 146 261 уникальных (число research Q3 — БЕЗ фолдинга), ё→е сливает 7347 е/ё-пар лемм → **138 914**; границы [137000, 140500]; en — 78 951 (совпало с research точно). Фолдинг — locked-решение плана (Pitfall 3/R3, 06-RESEARCH Open Question 3 «РЕШЕНО: фолдинг ё→е в данных dictgen»), оговорка плана «уточнить границы по факту Task 1, зафиксировать в комментарии» исполнена
- Нюанс prohibition-гейта: в исходнике dictgen есть map[string]struct{} — это dedup-множество ПАРСЕРА (рабочее состояние dev-инструмента), а не представление словаря; сгенерированные данные — только []string-литералы, golden-тест пинит slices.IsSorted. Запрет рантайм-map соблюдён по существу
- Вес триграмм (A7): trigrams.go 244 КБ (≈122 КБ/язык) — выше догадки «десятки КБ», но потолок 4096 удержан и вес ничтожен против +4.6 МБ словарей и бюджета SPEC §5

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] package main_test невозможен — тесты белого ящика package main**
- **Found during:** Task 1 (RED)
- **Issue:** план предписывал package main_test, но внешний тест-пакет не может достучаться до неэкспортированного пайплайна main-пакета (main неимпортируем — проверено эмпирически на go1.23), а белый ящик падал бы на линтере testpackage
- **Fix:** corpus в package main (прецедент репозитория: engine/conn_*, internal/indicator; testpackage не флагирует package main — проверено lint'ом), идентификаторы плана parseDIC/trainTrigrams/emit сохранены дословно
- **Files modified:** layouts/dictgen/main_test.go
- **Verification:** mise run ci зелёная (lint 0 issues)
- **Committed in:** 186652e

**2. [Rule 1 - Bug] Конверт счётчика DictRU 146 261 недостижим при обязательном фолдинге**
- **Found during:** Task 1 (GREEN, живой прогон)
- **Issue:** 146 261 из research Q3 измерен без ё→е-фолдинга; locked-решение плана требует фолдинг → факт 138 914; план внутренне противоречив (truth-текст «конверт 146 261» vs обязательный фолдинг)
- **Fix:** границы конверта скорректированы по оговорке самого плана («уточнить границы по факту Task 1, зафиксировать в комментарии»): [137000, 140500] + декомпозиция в комментариях TestGenerateLive и TestGolden_DictEnvelope
- **Files modified:** layouts/dictgen/main_test.go, layouts/dict_test.go
- **Verification:** живой прогон 138 914 в конверте; golden-тесты зелёные; дрейф пакета по-прежнему валит тест громко
- **Committed in:** d7e403a, 5258857

**3. [Rule 2 - Missing critical] Третья фикстура + temp-ветки громких отказов**
- **Found during:** Task 1 (RED)
- **Issue:** план перечислял 2 фикстуры, но ветка несогласованного счётчика требует собственной фикстуры; ветки битой count-строки, пустого словаря и нечитаемого файла покрыты temp-фикстурами — иначе критерий «фикстуры покрывают несогласованный счётчик» и дисциплина T-06-02-04 были бы неполны
- **Fix:** testdata/ru_bad_count.dic + table-driven refusals-сабтест (bad-count-line, empty, missing-file)
- **Files modified:** layouts/dictgen/testdata/ru_bad_count.dic, layouts/dictgen/main_test.go
- **Verification:** TestParseDIC/refusals зелёный; каждый sentinel (errCountMismatch/errBadCountLine/errEmptyDict) пинен
- **Committed in:** 186652e

**4. [Rule 3 - Blocking] [tasks.dictgen-regen] перенесён в конец mise.toml**
- **Found during:** Task 2 (GREEN)
- **Issue:** размещение сразу после [tasks.ci] делало literal verify-гейт плана (sed-диапазон до следующей `[`-секции) ложноположительным — заголовок самой задачи попадал в диапазон
- **Fix:** задача перенесена в конец файла; гейт MISE-TASK-OK исполнен дословно, семантика (вне ci) не изменилась
- **Files modified:** mise.toml
- **Verification:** MISE-TASK-OK; mise run ci не содержит dictgen-regen
- **Committed in:** 6455c96

---

**Total deviations:** 4 auto-fixed (1 Rule 1, 1 Rule 2, 2 Rule 3)
**Impact on plan:** все отклонения — либо оговорённые планом ходы (конверт-факт, mutation-RED), либо механические блокеры (импорт-семантика Go, lint-гейт). Скоуп не расширен; ноль новых зависимостей.

## Issues Encountered
- Bootstrap go:generate: первая генерация требует ручного `cd layouts && go run ./dictgen`, т.к. //go:generate-директивы dictgen живут в самих генерируемых файлах (паттерн generator/tables.go). После коммита goldens `go generate ./layouts` самоподдерживается (задача работает, директивы в трёх файлах — три идемпотентных прогона)
- Черновой прогон линтера на RED-скаффолде нашёл 7 замечаний (funlen/gocognit/goconst/lln/nlreturn/gocheckcompilerdirectives/nolintlint) — все исправлены до RED-коммита, lint 0 issues на каждом коммите
- unshare недоступен в сессии (Operation not permitted) — CI-доказательство выполнено через bwrap (см. coverage D7)

## Known Stubs

None — реализация полная, заглушек нет (RED-скаффолд Task 1 заменён реализацией в том же плане).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Данные детектора готовы: DictRU/DictEN (sort.SearchStrings-контракт) + TriRU/TriEN — входы плана 06-05 (internal/detect); golden-корпус детектора пинит ё-слова (ёлка/ещё — теперь е-формы) на этих данных
- Регенерация при обновлении hunspell-пакета: mise run dictgen-regen; дрейф счётчиков за конверты громко валит TestGolden_DictEnvelope — пере-пин deliberatively
- Полный регресс — план 06-08; ADR-007 лицензионная позиция подкреплена docs/LICENSE-data.md

---
*Phase: 06-avtokorrekcija-opcionalno*
*Completed: 2026-10-01*

## Self-Check: PASSED

- Файлы: все 10 созданных + mise.toml существуют на диске (11/11 FOUND)
- Коммиты: 186652e (Task 1 RED), d7e403a (Task 1 GREEN), 5258857 (Task 2 RED), 6455c96 (Task 2 GREEN) — присутствуют в git log; measured commits от ledger: 4
- План-level verification: go test ./layouts/... -race — зелёный; mise run dictgen-regen && git diff --exit-code -- layouts/ — REGEN-DETERMINISTIC; mise run ci — зелёная (D-08)
- Оба RED-корпуса: RED_EVIDENCE_OK по check tdd-red-evidence (Task 2 — планом предписанный mutation-прогон, откачен байт-чисто)

