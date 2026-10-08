---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
plan: 08
subsystem: docs
tags: [skillgen, codegen, golden-test, determinism, dictgen-discipline, agent-skills, config-reference, mise, tdd, dev-tool]

# Dependency graph
requires:
  - phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill (план 08-06)
    provides: "docs/CONFIG.md — русскоязычный справочник со строгой таблицей ключей (21 строка, ровно 5 колонок | Key | Type | Default | Range / vocabulary | Meaning |) — единственный источник генератора (D-8-10)"
provides:
  - "cmd/skillgen — dev-only генератор (диктген-дисциплина): parseKeyTable (канон-заголовок как якорь, 5 непустых ячеек, секция-квалифицированный ключ, hard errors file:line, never silent skips), renderRegion (байт-детерминированный регион: заголовок + подпись источника + canonical-таблица + verbatim-строки), replaceRegion (сплайс между сентинелами goswitch-config:generated BEGIN/END; отсутствие/дубль/реверс сентинелов — hard error), -check (регенерация в память + bytes.Equal, non-zero exit на дрейфе)"
  - "skills/goswitch-config/SKILL.md (NEW, git-tracked, D-8-11): frontmatter name goswitch-config + RU description (agentskills.io), ручной каркас «Процедуры» (автокоррекция, blocklist, хоткеи, звуки, a11y-магия с restart-семантикой и откатом uninstall) и «Диагностика» (goswitchctl status key=value/config_error последним, journalctl --user -u goswitchd, известные ограничения) + генерированный регион с 21 строкой ключей verbatim"
  - "Синхрон-гейт D-8-10 в обычном mise run test (golden-тест TestSkillgen_CommittedFileInSync — research Pattern 5a, ноль правок workflow) + dev-only mise-задача [tasks.skillgen-regen] по форме dictgen-regen; CI только сверяет, никогда не перегенерирует"
affects: [verify-work (Manual-Only верификация полезности SKILL.md для AI — план assumptions), будущие правки конфиг-ключей (вносятся только в docs/CONFIG.md, затем mise run skillgen-regen)]

# Actuals (#2632) — pairs with the plan's estimate to calibrate future estimates.
actuals:
  tokens: 8860    # chars/4 over the realized diff (35439 chars across cmd/skillgen + skills/ + mise.toml)
  tasks: 2
  commits: 3      # MEASURED: git rev-list --count gsd-plan-head-before-08-08..HEAD
plan_head_before: 785c1caf20f796bfa770c7ba5efb145143b363d7

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Диктген-дисциплина для док-генератора: канон-константа заголовка таблицы как якорь парсера (расхождение — громкий отказ), verbatim-копирование строк (ноль дрейфа по построению), byte-deterministic render (без часов/карт/окружения), заменa строго между сентинелами"
    - "Golden-гейт генерации внутри обычного тест-прогона (research Pattern 5a): test CompareCommitted vs render — CI-дисциплина tidy-diff без правки workflow; dev-регенерация — отдельная mise-задача"

key-files:
  created:
    - cmd/skillgen/main.go
    - cmd/skillgen/main_test.go
    - skills/goswitch-config/SKILL.md
    - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/red-evidence/08-08-task1-red.json
  modified:
    - mise.toml

key-decisions:
  - "RED-фаза по прецеденту 06-04/07-02/08-02 (RED-stub): RED-коммит несёт форму (консты путей/сентинелов/заголовка, тип tableRow, сигнатуры с нулевыми возвратами), corpus пал на целевых утверждениях — check tdd-red-evidence → RED_EVIDENCE_OK (5 целевых падений); компиляционная ошибка была бы INVALID_RED по #3770"
  - "TestSkillgen_CommittedFileInSync при отсутствии SKILL.md падает (не skip): скип сделал бы CI-гейт слепым к удалению файла (класс T-08-08-03); честный красный до создания артефакта — прямо предписан планом acceptance-критерием и помечен в Task Commits"
  - "Заголовок таблицы пиннут константой и валидируется парсером байт-в-байт: регион наследует заголовок источника; separator валидируется по форме (тире/двоеточия), эмитится канонический"
  - "Регион несёт подпись источника (D-8-10 + команды regen/check) — AI-ассистент у таблицы видит, где единственный источник и как пересобрать"

patterns-established:
  - "cmd/skillgen — канон doc-генератора из markdown-таблицы: parse/render/replace + -check, golden в mise run test (расширение прецедента layouts/dictgen на docs-артефакты)"

requirements-completed: ["D-8-10", "D-8-11"]

coverage:
  - id: D1
    description: "Генератор cmd/skillgen по диктген-дисциплине: парсит реальную таблицу docs/CONFIG.md (≥21 строка, ровно 5 колонок, секция-квалифицированные ключи), malformed-строка (4 колонки / пустая ячейка / ключ без точки) — hard error с file:line и без частичного результата; рендер байт-детерминирован; замена строго между сентинелами (отсутствие/дубль/реверс — ошибка); -check сверяет байт-в-байт с non-zero exit на дрейфе"
    requirement: "D-8-10"
    verification:
      - kind: unit
        ref: "cmd/skillgen/main_test.go#TestSkillgen_ParseRealConfigTable"
        status: pass
      - kind: unit
        ref: "cmd/skillgen/main_test.go#TestSkillgen_MalformedRowHardError (3 subtests)"
        status: pass
      - kind: unit
        ref: "cmd/skillgen/main_test.go#TestSkillgen_RenderDeterministic"
        status: pass
      - kind: unit
        ref: "cmd/skillgen/main_test.go#TestSkillgen_RegionSentinelsRequired (3 subtests)"
        status: pass
      - kind: other
        ref: "плановые гейты: mise exec -- go build ./... && go vet ./... → BUILD-VET-OK; grep сентинелы/docs/CONFIG.md/bytes.Equal в main.go → GEN-OK; греп ключевых литералов в main.go вне тестов → ноль"
        status: pass
    human_judgment: false
  - id: D2
    description: "Синхрон-гейт D-8-10 живёт в обычном mise run test (golden-тест CommittedFileInSync — рендер из текущего CONFIG.md == регион закоммиченного SKILL.md байт-в-байт) без правки .github/workflows; dev-only регенерация — mise-задача [tasks.skillgen-regen]; регенерация идемпотентна (байт-в-байт); mise run ci зелёный (exit 0)"
    requirement: "D-8-10"
    verification:
      - kind: unit
        ref: "cmd/skillgen/main_test.go#TestSkillgen_CommittedFileInSync (CI-гейт; ok в mise run ci)"
        status: pass
      - kind: other
        ref: "плановые гейты: mise run skillgen-regen && go run ./cmd/skillgen -check → CHECK-GREEN (exit 0); regen byte-idempotent (cmp до/после + git diff HEAD пуст → REGEN-IDEMPOTENT); git diff --name-only .github/workflows/ → 0 строк"
        status: pass
      - kind: other
        ref: "mise run ci → CI_EXIT=0 (build + vet + lint + test -race + tidy-diff, включая ok cmd/skillgen)"
        status: pass
    human_judgment: false
  - id: D3
    description: "skills/goswitch-config/SKILL.md (D-8-11) — git-отслеживаемый project skill: frontmatter (name: goswitch-config, русский description по agentskills.io), ручной каркас процедур (автокоррекция on/off + пороги + hot reload, blocklist D-53, хоткеи + закрытая таблица имён + «голая» клавиша tap_key, звуки + фиксированный bell, a11y-магия: Active-семантика, идемпотентное применение двух ручек, restart-семантика, снятие запрещено, откат uninstall only-if-present) и диагностики (goswitchctl status key=value с config_error последним, journalctl --user -u goswitchd, ограничения: GTK4-Wayland forwarded-события, терминалы fail-closed, поля паролей), регион между сентинелами с 21 строкой ключей verbatim + подпись источника"
    requirement: "D-8-11"
    verification:
      - kind: other
        ref: "плановые гейты: SKILL-SHAPE-OK (файл есть, name: goswitch-config, оба сентинела); TRACKED-OK (git ls-files непуст); 21 строка ключей в регионе (grep -c '^| `'); подпись docs/CONFIG.md в регионе"
        status: pass
      - kind: unit
        ref: "cmd/skillgen/main_test.go#TestSkillgen_CommittedFileInSync (регион байт-синхронен CONFIG.md)"
        status: pass
    human_judgment: true
    rationale: "Полнота и фактическая точность ручного каркаса процедур/диагностики, качество русской прозы и полезность скилла для AI-ассистента — суждение владельца (Manual-Only верификация, plan assumptions; Automated-часть доказана гейтами выше)"

# Metrics
duration: 16min
completed: 2026-10-05
status: complete
---

# Phase 8 Plan 08: Skill «Конфигурация goswitch» — cmd/skillgen + SKILL.md + golden-гейт Summary

**Dev-only генератор cmd/skillgen переписывает конфиг-регион skills/goswitch-config/SKILL.md из таблицы ключей docs/CONFIG.md verbatim (диктген-дисциплина: hard errors, байт-детерминизм), синхрон-гейт живёт golden-тестом в обычном mise run test, регенерация — mise-задачей skillgen-regen; SKILL.md — русскоязычный project skill с ручным каркасом процедур/диагностики и 21 строкой ключей, байт-синхронной с CONFIG.md.**

## Performance

- **Duration:** 16 min
- **Started:** 2026-10-05T19:34:02Z
- **Completed:** 2026-10-05T19:50:05Z
- **Tasks:** 2 (Task 1 tracer tdd — строгий RED→GREEN; Task 2 auto)
- **Files:** 4 создано, 1 изменено (+ red-evidence-запись)

## Accomplishments

- cmd/skillgen (диктген-дисциплина, package doc по канону): `parseKeyTable` — канон-заголовок таблицы как якорь, ровно 5 непустых ячеек, секция-квалифицированный ключ, markdown-экранированные `\|` остаются внутри ячейки, нарушение — громкая ошибка с file:line, никогда тихий пропуск; `renderRegion` — байт-детерминированный регион (заголовок + подпись источника с D-8-10 и командами regen/check + canonical-таблица + строки verbatim); `replaceRegion` — сплайс строго между сентинелами `goswitch-config:generated BEGIN/END` с сохранением ручного каркаса, отсутствие/дубль/реверс сентинелов — ошибка; `-check` — регенерация в память + `bytes.Equal`, non-zero exit на дрейфе
- skills/goswitch-config/SKILL.md (D-8-11, git-tracked): frontmatter по agentskills.io, ручной каркас «Процедуры» (5 процедур короткими шагами с именами ключей) и «Диагностика» (status/journalctl/ограничения), генерированный регион — 21 строка ключей verbatim из docs/CONFIG.md
- Синхрон-гейт (D-8-10): TestSkillgen_CommittedFileInSync гоняется обычным `mise run test` внутри pr-sanity — ноль правок workflow (research Pattern 5a); dev-only `[tasks.skillgen-regen]` в mise.toml по форме dictgen-regen; CI только сверяет golden
- TRACKED-гейт: SKILL.md добавлен явным `git add` (новый каталог), `git ls-files` непуст — гейт CI сверяет видимый файл (T-08-08-03 закрыт)

## Task Commits

Each task was committed atomically (TDD: RED → GREEN for Task 1):

1. **Task 1 (трейсер): генератор cmd/skillgen — parse CONFIG.md → render → check** — `3655226` (test, RED) → `db1c9ac` (feat, GREEN); REFACTOR не потребовался (реализация легла в форму канона без чистки — lint-фиксы nlreturn/lll/mnd/gocritic вошли в GREEN)
2. **Task 2: SKILL.md (каркас + регион), mise-задача, git-tracking, финальный гейт** — `b204342` (feat)

**Plan metadata:** отдельный docs-коммит (SUMMARY + STATE + ROADMAP + red-evidence) следует за этим файлом.

_TDD Gate Compliance: RED `test(08-08) 3655226` предшествует GREEN `feat(08-08) db1c9ac`; RED-свидетельство `check tdd-red-evidence` → `RED_EVIDENCE_OK` (5 целевых тестов пали на запланированных утверждениях против shape-стаба по прецеденту 06-04/08-02; raw-вывод /tmp/08-08-t1-red-raw.txt, запись red-evidence/08-08-task1-red.json). Коммит-паттерн плана «RED = parseKeyTable/renderRegion не существуют» исполнен как shape-стаб с нулевым поведением: компиляционная ошибка была бы INVALID_RED по #3770, а падение корпуса на утверждениях — требование RED-гейта. Известный плану честный красный: CommittedFileInSync в Task 1 падал по отсутствию SKILL.md (артефакт Task 2) — предписан acceptance-критерием плана, закрыт в Task 2 («ok cmd/skillgen» в mise run ci)._

## Files Created/Modified

- `cmd/skillgen/main.go` — генератор: parse/render/replace + -check dispatch, сентинелы и канон таблицы константами (создан)
- `cmd/skillgen/main_test.go` — корпус из 5 тестов (реальная таблица / hard errors / детерминизм / сентинелы / CI-гейт) (создан)
- `skills/goswitch-config/SKILL.md` — project skill: frontmatter, ручной каркас, генерированный регион (создан)
- `mise.toml` — [tasks.skillgen-regen] после формы dictgen-regen (изменён)
- `.planning/phases/08-.../red-evidence/08-08-task1-red.json` — RED-свидетельство Task 1 (RED_EVIDENCE_OK)

## Decisions Made

- RED-фаза — shape-стаб (консты + тип tableRow + сигнатуры с нулевыми возвратами) по прецеденту 06-04/07-02/08-02: план формулировал RED как «функции не существуют», но в Go это компиляционная ошибка (INVALID_RED по #3770); стаб даёт падение корпуса на целевых утверждениях — RED_EVIDENCE_OK
- CommittedFileInSync при отсутствии файла падает, а не скипается: skip сделал бы CI-гейт слепым к удалению SKILL.md; красный между Task 1 и Task 2 — план-пиннутый (acceptance-критерий), закрыт созданием артефакта
- Заголовок таблицы пиннут константой и валидируется парсером (расхождение источника — громкий отказ), separator валидируется по форме; регенерация идемпотентна байт-в-байт
- Подпись региона называет единственный источник и обе команды (regen/check) — AI-ассистент у таблицы сразу видит дисциплину правок

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- **Известный средовый флак `internal/ctlsvc` (out of scope, не чинился).** Первый прогон `mise run ci` упал на `TestRun_OnConnHookCalledOnce` («OnConn called 0 times») — задокументированная в deferred-items.md гонка владения `org.djarvur.goswitch` с ЖИВЫМ демоном владельца на реальной session bus (плановая заметка: «NOT yours»). Повторный прогон зелёный (CI_EXIT=0). Диф плана — cmd/skillgen + SKILL.md + mise.toml, ноль строк в internal/.
- **REGEN-IDEMPOTENT-гейт до коммита показал бы дифф против HEAD** (новый файл staged, но не в HEAD): идемпотентность доказана cmp-копией до/после regen + пустым `git diff` (index vs worktree); точная плановая команда (`git diff HEAD ...` пусто → REGEN-IDEMPOTENT) выполнена после коммита — прошла.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Все 8 планов фазы 08 исполнены; поток 3 закрыт — skill «Конфигурация goswitch» живёт в репо (skills/goswitch-config/SKILL.md, git-tracked) и не может разойтись с docs/CONFIG.md ни молча, ни заметно (расхождение = красный CI)
- Дисциплина будущих правок: ключи меняются ТОЛЬКО в docs/CONFIG.md → `mise run skillgen-regen` → коммит обоих файлов; ручной каркас SKILL.md правится руками, генератор его не трогает
- Готово к `/gsd:verify-work` (Manual-Only верификация полезности SKILL.md для AI — план assumptions); блокеров нет

---
*Phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill*
*Completed: 2026-10-05*

## Self-Check: PASSED

- cmd/skillgen/main.go, cmd/skillgen/main_test.go, skills/goswitch-config/SKILL.md — FOUND на диске
- коммиты 3655226 (RED), db1c9ac (GREEN), b204342 (Task 2) — FOUND в git log
- commits по ledger gsd-plan-head-before-08-08: 3 (rev-list --count 785c1ca..HEAD)
- плановые гейты: BUILD-VET-OK / GEN-OK / SKILL-SHAPE-OK / TRACKED-OK / MISE-TASK-OK / REGEN-IDEMPOTENT / CHECK-GREEN — все PASS; mise run ci exit 0 (ctlsvc-флак — известный средовый, зелёно на повторе); NO-DELETIONS; .github/workflows/ не тронуты; стабов/скипов/TODO в созданных файлах нет
