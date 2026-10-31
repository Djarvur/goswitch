---
phase: 11-menu-punkt-v-chernyy-spisok
plan: 03
subsystem: config
tags: [yaml, atomic-write, strict-validation, blocklist, tdd, yaml-node-roundtrip]

# Dependency graph
requires:
  - phase: 07-svet-i-tumbler-menu
    provides: persist-writer (Node round-trip, temp+rename 0600), T-07-03-01/02 атомарная посадка
  - phase: 06-autocorrect-golden
    provides: схема autocorrect.apps_blocklist, Autocorrect.validate (потолок/regex/пустой паттерн), строгий декодер KnownFields
provides:
  - AppendAutocorrectBlocklist(path, pattern) — атомарный append ГОТОВОГО regex-паттерна: read → absent-ветки → Node round-trip → дедуп → потолок → strict pre-validation → rewriteAtomically
  - MaxAutocorrectBlocklist (экспорт 64) — единый источник потолка для writer'а, актора (11-02) и меню (11-04)
  - decodeValid(r io.Reader) в load.go — семантика Load (KnownFields + Validate) над произвольным ридером; Load и pre-write гейт writer'а делят один декодер
  - golden/idempotent/atomic/strict/absent/ceiling/метасимволы/not-seq корпус append-метода (writer_test.go, config_test.go)
affects: [11-02 (actor full-состояние по MaxAutocorrectBlocklist), 11-04 (клик-композиция QuoteMeta → AppendAutocorrectBlocklist), 11-05 (e2e menu-blocklist)]

# Actuals (#2632)
actuals:
  tokens: 9183
  tasks: 2
  commits: 5

# Tech tracking
tech-stack:
  added: [] # ноль новых модулей — stdlib + yaml.v3, go.mod/go.sum нетронуты (tidy-diff зелёный)
  patterns:
    - "append-на-готовом-persist-path: новый метод садится на setDocumentToggle-каркас (read → ErrNotExist-ветка → Node → rewriteAtomically), свой writer не строится (S3/S6)"
    - "strict-pre-validation: результат кодируется в байты и прогоняется через Load-семантику ДО rename; отказ — оригинал байт-нетронут (S5, D-33)"
    - "single-ceil-source: экспортированный MaxAutocorrectBlocklist — единственный источник 64 для validate/writer/actor/menu"

key-files:
  created: []
  modified:
    - internal/config/writer.go
    - internal/config/load.go
    - internal/config/config.go
    - internal/config/writer_test.go
    - internal/config/config_test.go

key-decisions:
  - "Pre-write валидация одноисточна: decodeValid извлечён в load.go (Load поверх ридера) — writer-гейт и Load делят ровно одну строгую семантику, дубликат декодера не заводится (отклонение: load.go сверх files_modified плана)"
  - "Потолок-гейт стоит ПОСЛЕ дедупа и ДО append: отказ sentinel'ом errAutocorrectBlocklistOverCeil называет ТЕКУЩИЙ размер (64) — over-ceiling документ никогда не конструируется, файл байт-нетронут"
  - "Отказ формы (apps_blocklist — не последовательность) реализован уже в GREEN Задачи 1 (writer-уровень, errWriterBlocklistNotSequence): append в скаляр портит дерево — корректность раньше пина; тест Задачи 2 садится на готовое поведение"
  - "buildBlocklistSection — отдельный строитель absent-секции append-пути (enabled остаётся D-54 false: блокировка приложения никогда не включает слой), buildSectionNode не переиспользован (bool-специфичная сигнатура — Anti-Pattern RESEARCH)"

patterns-established:
  - "appendRefusedUntouched — общий отказ-оракул corpus'а: append на битом входе возвращает ошибку И файл байт-равен до/после"
  - "goldenAddedLines — golden-дифф упорядоченной вставки: единственная новая строка, ноль сдвинутых/удалённых существующих"

requirements-completed: [MENU-01]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Append растёт ровно на одну строку списка, комментарии и все прочие секции байт-целы (golden-дифф), результат принимается строгим Load"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_AppendGoldenDiff"
        status: pass
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_AppendRoundTrip"
        status: pass
    human_judgment: false
  - id: D2
    description: "Absent-ветки полны: несуществующий файл → полный дефолтный документ с паттерном (ensure-ветка); absent-key / явный null / absent-section → последовательность создана, паттерн на месте, слой остаётся D-54 off"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_AppendAbsentFile"
        status: pass
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_AppendAbsentKeyAndNull"
        status: pass
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_AppendInsertsMissingSection"
        status: pass
    human_judgment: false
  - id: D3
    description: "Дедуп: паттерн уже в списке → идемпотентный success без записи, байты до/после идентичны (механическая защита D-01 «клик по отмеченному безопасен»)"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_AppendDedupNoOp"
        status: pass
    human_judgment: false
  - id: D4
    description: "Атомарная дисциплина: после записи нет .tmp-артефактов, режим 0600 (T-07-03-01/02 не тронуты)"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_AppendAtomicWrite"
        status: pass
    human_judgment: false
  - id: D5
    description: "Строгая pre-write валидация (семантика Load: KnownFields + Validate) по байтам до rename: битый результат / пустой паттерн → ошибка, оригинал байт-нетронут"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_AppendBrokenSourceUntouched"
        status: pass
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_AppendEmptyPatternRefused"
        status: pass
    human_judgment: false
  - id: D6
    description: "Потолок 64 с sentinel-отказом до append (текущий размер назван, файл байт-нетронут); 63 → success + строгий Load; MaxAutocorrectBlocklist — экспортированный единый источник (старое unexported имя в репозитории отсутствует)"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_AppendBlocklistCeil"
        status: pass
      - kind: unit
        ref: "internal/config/config_test.go#TestMaxAutocorrectBlocklistExport"
        status: pass
      - kind: other
        ref: "grep -rn maxAutocorrectBlocklist --include='*.go' . → ноль совпадений"
        status: pass
    human_judgment: false
  - id: D7
    description: "Метасимволы: QuoteMeta-экранированный литерал садится в список дословно и переживает строгий Load; сырой валидный RE2 writer'ом НЕ отвергается (валидация = regexp.Compile, не политика экранирования)"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_AppendMetacharPatterns"
        status: pass
    human_judgment: false
  - id: D8
    description: "append к документу с apps_blocklist-скаляром → отказ формы (errWriterBlocklistNotSequence), оригинал байт-нетронут"
    requirement: MENU-01
    verification:
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_AppendNonSequenceRefused"
        status: pass
    human_judgment: false
  - id: D9
    description: "Зелёная итерация целиком (директива 2): build + vet + golangci-lint 0 issues + test -race ./... (22 пакета) + go mod tidy без дрейфа go.mod/go.sum"
    verification:
      - kind: other
        ref: "go build ./... && go vet ./... && golangci-lint run (0 issues) && go test -race -count=1 ./... (exit 0) && git diff --exit-code -- go.mod go.sum"
        status: pass
    human_judgment: false

# Metrics
duration: 34min
completed: 2026-10-31
status: complete
---

# Phase 11 Plan 03: MENU persist-примитив — AppendAutocorrectBlocklist Summary

**AppendAutocorrectBlocklist — атомарный append готового regex-паттерна в autocorrect.apps_blocklist: golden-дифф «+1 строка списка», дедуп no-op, потолок 64 sentinel'ом ДО записи, строгая pre-write валидация по Load-семантике и MaxAutocorrectBlocklist как единый источник потолка (MENU-01 persist-слой).**

## Performance

- **Duration:** 34 min
- **Started:** 2026-10-31T20:44:00Z
- **Completed:** 2026-10-31T21:18:07Z
- **Tasks:** 2 (полный RED→GREEN→REFACTOR на каждую)
- **Files modified:** 5

## Accomplishments

- AppendAutocorrectBlocklist(path, pattern string) error на каркасе setDocumentToggle: read → ErrNotExist-ветка (ensure: полный дефолтный документ с паттерном, слой D-54 off) → yaml.Node round-trip → поиск sectionAutocorrect/keyAppsBlocklist (absent-ключ, явный null, absent-секция — все ветки) → дедуп (идемпотентный byte-level no-op) → потолок → append scalarString (Style 0, yaml.v3 сам закавычивает спец-символы) → strict pre-write валидация → rewriteAtomically (temp same-dir + rename + 0600, T-07-03-01/02)
- Потолок 64: гейт ПОСЛЕ дедупа и ДО append — sentinel errAutocorrectBlocklistOverCeil называет ТЕКУЩИЙ размер; maxAutocorrectBlocklist → экспортированный MaxAutocorrectBlocklist, ноль старого имени по репозиторию; для 11-02 (full-состояние актора) и 11-04 (метка заполненности меню) источник готов
- Критерии 2 и 4 роадмапа спиннуты юнит-корпусом: golden-дифф «изменилась только новая строка, комментарии целы» и «имя с метасимволами не ломает конфиг» (QuoteMeta-литерал дословно + сырой валидный RE2 не отвергается)
- decodeValid извлечён в load.go — Load и writer-гейт делят одну строгую семантику (KnownFields + Validate); никакого дублированного декодера
- Корпус: 12 новых тест-функций (writer_test.go) + пин экспорта (config_test.go); писатель по-прежнему ничего не логирует (D-20/D-21)

## Task Commits

Каждая задача — полный TDD-цикл с атомарными коммитами:

1. **Task 1 RED: append-corpus** - `97a4255` (test) — 9 тест-функций + compile-stub API; RED evidence `RED_EVIDENCE_OK` (TestWriter_AppendRoundTrip упал на planned-behavior-assertion)
2. **Task 1 GREEN: реализация** - `fd16b81` (feat) — append/absent-ветки/дедуп/strict pre-validation + decodeValid в load.go
3. **Task 1 REFACTOR** - `a4be13f` (refactor) — encodeDocument: один encode path для persist и pre-write гейта
4. **Task 2 RED: ceiling-corpus + пины** - `2d470d0` (test) — TestWriter_AppendBlocklistCeil (RED target), метасимволы, not-seq; RED evidence `RED_EVIDENCE_OK`
5. **Task 2 GREEN: потолок + экспорт** - `3f7b5b5` (feat) — MaxAutocorrectBlocklist, sentinel до append, пин экспорта, corpus на экспортной константе

## Files Created/Modified

- `internal/config/writer.go` — AppendAutocorrectBlocklist + appendBlocklistAbsentFile + landBlocklistPattern/landBlocklistKey + buildBlocklistSection + blocklistSeq + validateNode/encodeDocument + sentinel errWriterBlocklistNotSequence; потолок-гейт
- `internal/config/load.go` — decodeValid(r io.Reader): семантика Load поверх произвольного ридера; Load(refactor) = open + decodeValid
- `internal/config/config.go` — MaxAutocorrectBlocklist (экспорт 64, док-комментарий «единый источник»), использования validate/комментариев обновлены
- `internal/config/writer_test.go` — append-corpus: 12 тест-функций, 5 новых фикстур-документов, goldenAddedLines/assertAppendCommentsSurvive/appendRefusedUntouched хелперы
- `internal/config/config_test.go` — TestMaxAutocorrectBlocklistExport

## Decisions Made

- **decodeValid в load.go (единая Load-семантика).** План предписывал pre-write валидацию «семантикой Load» — извлечение decodeValid(r io.Reader) даёт это дословно без копии декодера; Load свёлся к open + decodeValid. Файл сверх files_modified плана (отклонение, см. ниже).
- **Потолок до append, sentinel с текущим размером.** Отказ именует «has 64 entries» (текущий список), а не сконструированные «65» — различает writer-сентинел и strict-validation отказ машинно.
- **Форма-отказ раньше пина.** errWriterBlocklistNotSequence реализован в GREEN Задачи 1: append ребёнка в скаляр портит дерево — это корректность (Rule 2), тест Задачи 2 садится на готовое поведение (пин, не RED).
- **buildBlocklistSection вместо buildSectionNode.** Сигнатура buildSectionNode bool-специфична (Anti-Pattern RESEARCH); отдельный строитель append-пути: enabled остаётся false (D-54) — блокировка приложения никогда не включает слой.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] mise run ci в сессии исполнителя падает на toolchain-разъёме (2-е проявление, прецедент 11-01)**
- **Found during:** Task 1 (verify-шаг)
- **Issue:** env сессии несёт GOROOT mise-инсталла go1.23.12 при `go`-бинарнике системы 1.27.3 (usr/local/go/bin стоит выше mise-путей в PATH задачи): «compile: version go1.23.12 does not match go tool version go1.27.3»; `mise exec -- go build` работает (mise перехватывает прямой вызов), но shell-задачи `mise run` резолвят системный go
- **Fix:** те же гейты исполнялись напрямую: go build ./..., go vet ./..., golangci-lint 2.15.0 (бинарник из mise-инсталла, GOROOT сброшен), go test -race -count=1 ./... (22 пакета, exit 0), go mod tidy + git diff --exit-code — согласовано со STACK.md («build with current 1.27.x», язык модуля go 1.23). Записано в WINDOWS.md (#23, kind unrun-verify)
- **Files modified:** нет (только метод прогона)
- **Verification:** все гейты зелёные на финальном HEAD 3f7b5b5
- **Committed in:** — (методика, не код)

**2. [Rule 2 - Missing critical] decodeValid извлечён в load.go (файл сверх files_modified)**
- **Found during:** Task 1 (GREEN)
- **Issue:** pre-write валидация обязана быть «семантикой Load» (KnownFields + Validate); без общего кода появился бы дублирующий строгий декодер в writer.go
- **Fix:** load.go получил decodeValid(r io.Reader), Load = open + decodeValid; writer-гейт зовёт decodeValid по байтам из encodeDocument
- **Files modified:** internal/config/load.go
- **Verification:** полный корпус config-пакета зелёный (включая все прежние Load-тесты — поведение Load байт-неизменное)
- **Committed in:** fd16b81

---

**Total deviations:** 2 auto-fixed (1 blocking-методика, 1 missing-critical-композиция). **Impact on plan:** оба отклонения необходимы (зелёная итерация директивы 2; одноисточность D-33). Скоуп-крипа нет: go.mod/go.sum нетронуты, файлы плана не расширены сверх load.go.

## Issues Encountered

- **Фикстуры RED-корпуса Задачи 1 изначально были не Load-валидны** (минимальные документы без hotkeys/timeouts/correction): строгая pre-write валидация честно отвергала их результат. Демон никогда не принимал частичные документы (D-33), поэтому фикстуры приведены к полным Load-валидным документам — это итерация тест-авторства внутри RED/GREEN, не изменение поведения.
- **До-существующий флак watch-тестов под -race:** по одному одиночному флаку TestWatch_A11yUnknownKeyKeepsLastGood и TestWatch_BrokenAutocorrectBlockKeepsLastGood в разных прогонах (mtime-poll тайминги при нагрузке), повторные прогоны зелёные; код плана эти файлы не трогает — известный deferred-класс v1.0.0 «config/ctlsvc/appid флаки под -race» (прецедент 10-18: одиночный флак, два повторных зелёных). Не чинилось (вне скоупа).
- **Пины, зелёные до GREEN:** TestWriter_AppendDedupNoOp/TestWriter_AppendAtomicWrite (Задача 1) и метасимволы/not-seq (Задача 2) проходят уже на stub'e/готовом коде — они пины корпуса; RED-гейт закрывается целевым тестом каждого таска (AppendRoundTrip; AppendBlocklistCeil).

## TDD Gate Compliance

| Task | RED | GREEN | REFACTOR | Status |
|------|-----|-------|----------|--------|
| 1 | 97a4255 (test) | fd16b81 (feat) | a4be13f | Pass |
| 2 | 2d470d0 (test) | 3f7b5b5 (feat) | — (рефакторинг не потребовался сверх a4be13f) | Pass |

- RED evidence обеих задач проверено машиной: `gsd_run check tdd-red-evidence` → RED_EVIDENCE_OK (target_test_failed) в обоих случаях; raw-выводы сохранены (/tmp/gsd-11-03-red-raw.txt, /tmp/gsd-11-03-t2-red-raw.txt).
- Compile-stub API в RED-коммите Задачи 1 — необходимость Go (корпус именует API до реализации; без символа пакет не собирается = INVALID_RED); GREEN заменил тело. Stub не содержит поведения (return nil) — RED-провал честный, на assertion.
- Порядок gate-коммитов проверен: test(...) предшествует feat(...) для обоих тасков.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- 11-02 (пункт меню/freshness/D-04) и 11-04 (клик-композиция/проводка) получают: готовый persist-примитив (писатель получает ГОТОВЫЙ паттерн — QuoteMeta в композиции), экспортированный MaxAutocorrectBlocklist для full-состояния/метки заполненности, sentinel-семантику отказа потолка (отказ виден состоянием пункта, S9).
- 11-05 (e2e) — oracle «документ вырос ровно на паттерн + строгий Load» уже спиннут юнит-корпусом (golden-дифф).
- Блокеров нет; go.mod/go.sum заморожены; miscele: literal `mise run ci` на этой машине остаётся за direct-гейтами до починки сессийного PATH/GOROOT разъёма (WINDOWS #23, открыт).

## Self-Check: PASSED

- Все 5 файлов плана существуют на диске (writer.go, load.go, config.go, writer_test.go, config_test.go) — FOUND.
- Все 5 коммитов плана присутствуют в истории (97a4255, fd16b81, a4be13f, 2d470d0, 3f7b5b5) — FOUND.
- commits измерены от ledger-базы 407b617f: git rev-list --count = 5.
- Acceptance criteria обеих задач перепроверены на HEAD 3f7b5b5: 8 behavior-строк корпусом покрыты; golden-дифф пинен; strict pre-validation пинен; потолок 64/63 пинен; grep старого имени — ноль; go.mod/go.sum — diff --exit-code чист.

---
*Phase: 11-menu-punkt-v-chernyy-spisok*
*Completed: 2026-10-31*
