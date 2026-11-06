---
phase: 09-spayk-i-defekt-wezterm
plan: 01
subsystem: docs
tags: [spec-delta, adr, terminal-transport, wezterm, ibus, term-04, d-55]

# Dependency graph
requires:
  - phase: 08-spisok-dokumentatsiya-i-nav (v1.1.0 close)
    provides: "SPEC.md с конвенцией аудит-трейл-ревизий (rev G-2-3, rev G-5-5, D-53), ADR-006/007 как формат-образцы"
provides:
  - "SPEC.md — датированная ревизия rev D-55 в пяти секциях (§4.3/§4.4/§5/§7.2/§11): двухразрешительная модель, вето TERM-04 в точке армирования (роль AT-SPI 60, не blocklist), завершение бюджета (< 50 мс = первое действие на проводе; терминальное завершение best-effort), two-greens-дисциплина wezterm-рядов, feasibility-гейт строки"
  - "ADR-008 «Терминальный транспорт» в Proposed: решения D1–D7 с вариантами и пустыми Evidence-ячейками-плейсхолдерами журнала adr-008-spike-log.md; ни одной веб-ссылки"
  - "Именованные якоря для wave-2 планов: журнал adr-008-spike-log.md (создаёт 09-02), перевод ADR-008 в Accepted (ставит 09-03)"
affects: [09-02 live spike, 09-03 wezterm fix, phase-10 word chord, phase-12 line, phase-11 menu, phase-13 a11y]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
# Same estimateTokens scale (chars/4 over the realized diff), never a harness token count.
actuals:
  tokens: 7020
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Спек-ревизия по конвенции аудит-трейла: HTML-комментарий с прежним текстом дословно + новый текст с маркером (rev D-55, …) — четвёртый прецедент после rev G-2-3/G-5-5/D-53"
    - "ADR-пустышка: Proposed с плейсхолдерами Evidence-ячеек на строки живого журнала спайка — решения фиксируются до кода, свидетельства добавляет только спайк"

key-files:
  created:
    - docs/adr/ADR-008-terminal-transport.md
  modified:
    - docs/SPEC.md

key-decisions:
  - "Вето автокоррекции в терминалах живёт в точке армирования (роль AT-SPI 60 / терминальная identity, проверка ДО запуска детектора) — никогда через autocorrect.apps_blocklist (коллизия с D-53 «blocklist не может исключить невидимое» названа явно)"
  - "Двухразрешительная модель SPEC §11: ручная коррекция в терминалах разрешена, автокоррекция — никогда (решение владельца 2026-10-08, терминал = команды)"
  - "Бюджет §5 завершён: реакция < 50 мс = хоткей → первое действие на проводе; завершение коррекции в терминале — best-effort с фазовыми дедлайнами (нога select→read ~150–400 мс)"
  - "Транспорт терминалов §4.4 — собственный канал демона (ForwardKeyEvent + commit_text); внешние инжекторы и paste-чорды в демоне отвергнуты"
  - "ADR-008 D1–D7 зафиксированы как Proposed с пустыми Evidence-ячейками — ни одна строка не считается решённой до живых свидетельств (Anti-Pattern «фикс из issue-tracker'а»: апстрим-факты без ссылок)"

patterns-established:
  - "Spec-delta rev D-55: датированная ревизия с дословным аудит-трейлом в HTML-комментарии — паттерн для всех будущих spec-правок вехи v1.2"
  - "ADR-008 skeleton: таблица «Решение | Варианты | Evidence (живой спайк)» с плейсхолдерами «(spike: строка … журнала adr-008-spike-log.md, терминал …)» — заполняется только живым спайком"

requirements-completed: [TERM-01, TERM-04]

coverage:
  - id: D1
    description: "SPEC §11 — двухразрешительная модель и вето TERM-04 в точке армирования (роль AT-SPI 60), blocklist упомянут только как то, через что вето НЕ выражается"
    requirement: TERM-04
    verification:
      - kind: other
        ref: "grep -q 'никогда не армируется в контексте терминала' docs/SPEC.md && grep -q 'точке армирования' docs/SPEC.md && grep -q 'в терминалах разрешена' docs/SPEC.md"
        status: pass
    human_judgment: true
    rationale: "Словесная формулировка контракта проходит ручной гейт владельца на verify-work (Test Map фазы: TERM-04 spec-delta = docs review, manual gate); grep-гейты подтверждают механику, не адекватность формулировки"
  - id: D2
    description: "SPEC §4.3/§4.4/§5/§7.2 — терминальный пункт коррекции (слово обязателен, строка feasibility-gated), транспортный контракт (ForwardKeyEvent + commit_text), завершение бюджета (< 50 мс = первое действие на проводе; best-effort), two-greens-дисциплина; прежний текст §5 сохранён дословно в аудит-трейле"
    requirement: TERM-01
    verification:
      - kind: other
        ref: "test $(grep -c 'rev D-55' docs/SPEC.md) -ge 5 (маркер в каждом из пяти разделов: 4.3×2, 4.4×2, 5×2, 7.2×3, 11×1); grep 'первое действие на проводе' + 'best-effort' + 'два подряд' + 'замороженн' + 'feasibility' — все PASS"
        status: pass
      - kind: other
        ref: "git diff -U0 — хунки только в пяти целевых секциях (@@145/160/334/374/508); прежний текст пункта §5 встречается дословно 2 раза (аудит-комментарий + ревизованный пункт)"
        status: pass
    human_judgment: true
    rationale: "Содержательная адекватность ревизий (полнота рамки фикса TERM-01, честность feasibility-гейта) — ревью владельца на verify-work"
  - id: D3
    description: "ADR-008 «Терминальный транспорт» в Proposed: Status с указанием переводящего плана 09-03, Context (гипотеза XIM + h1/h2/h3 + P0–P5), семь решений D1–D7 с вариантами, каждая Evidence-ячейка — плейсхолдер журнала adr-008-spike-log.md, ни одной веб-ссылки"
    requirement: TERM-01
    verification:
      - kind: other
        ref: "test -f docs/adr/ADR-008-terminal-transport.md && grep '^## Status' + 'Proposed'"
        status: pass
      - kind: other
        ref: "grep -oE '^\\| D[1-7] ' | sort -u | wc -l = 7; grep -c 'adr-008-spike-log' = 10 (>= 7); ! grep 'https\\?://' (веб-ссылок нет)"
        status: pass
    human_judgment: false

# Metrics
duration: 10 min
completed: 2026-10-08
status: complete
---

# Phase 9 Plan 1: SPEC spec-delta rev D-55 + ADR-008 Summary

**Документационный фундамент вехи v1.2 до всякого кода (D-55): SPEC несёт rev D-55 в пяти секциях — двухразрешительная модель с вето TERM-04 в точке армирования (роль AT-SPI 60, не blocklist), транспорт терминалов через собственный канал демона, завершение бюджета, two-greens-дисциплина; ADR-008 «Терминальный транспорт» создан в Proposed с решениями D1–D7 и пустыми Evidence-ячейками под живой спайк**

## Performance

- **Duration:** 10 min
- **Started:** 2026-10-08T19:40:41Z
- **Completed:** 2026-10-08T19:50:53Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- SPEC rev D-55: пять датированных ревизий (§4.3 терминальный пункт коррекции, §4.4 транспортный контракт, §5 завершение бюджета, §7.2 two-greens + замороженные негативы, §11 двухразрешительная модель и дом вето) — каждая по конвенции аудит-трейла, прежний текст §5 сохранён дословно; diff ограничен пятью секциями
- Вето TERM-04 закреплено spec-ом: «никогда не армируется в контексте терминала (роль AT-SPI 60…)», проверка в точке армирования ДО запуска детектора, `autocorrect.apps_blocklist` упомянут только как то, через что вето НЕ выражается; коллизия с D-53 названа явно
- ADR-008 Proposed: Context с помеченной гипотезой (HIGH уверенности, не вердикт), классами разрыва h1/h2/h3 и лестницей P0–P5; таблица D1–D7 «Решение | Варианты | Evidence (живой спайк)» с пустыми плейсхолдерами журнала adr-008-spike-log.md; ноль веб-ссылок
- Зелёная итерация подтверждена: `go build ./...`, `go vet ./...`, `go test -race -count=1 ./...` (все пакеты ok), `golangci-lint run` (0 issues) — docs-правки дерево не трогали

## Task Commits

Each task was committed atomically:

1. **Task 1: SPEC spec-delta rev D-55 — пять секций с дословным аудит-трейлом** - `ac676c4` (docs)
2. **Task 2: ADR-008 «Терминальный транспорт» — Proposed, решения D1–D7, пустые Evidence-ячейки** - `2a6e67a` (docs)

## Files Created/Modified
- `docs/SPEC.md` — ревизии rev D-55 в §4.3/§4.4/§5/§7.2/§11 (+80/−1, пять хунков)
- `docs/adr/ADR-008-terminal-transport.md` — новый ADR, Proposed (129 строк)

## Decisions Made
- Для аддитивных секций (§4.3/§4.4/§7.2/§11) аудит-трейл-комментарий фиксирует «ревизия только добавляет, прежние пункты не менялись» — по внутрифайловому прецеденту ревизии a11y 2026-10-05 (§4.5); дословное сохранение прежнего текста применено к единственному переписанному пункту (§5 производительность)
- Инвариант TERM-04 внесён в §11 дословно из 09-RESEARCH как цитата-блоклок, с разбором коллизии с D-53 рядом (вне цитаты) — цитата не разбавлялась
- Апстрим-ловушки wezterm (#5125/#1772/#4615) упомянуты в ADR-008 без ссылок как чекпоинт-ряды спайка — формулировки без «http» прошли grep-гейт Anti-Pattern 2/4

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Wave 2 разблокирована: 09-02 (живой спайк) создаёт `docs/adr/adr-008-spike-log.md` и заполняет строки журнала; 09-03 заполняет Evidence-ячейки и переводит ADR-008 в Accepted; 09-04 (A11Y-зонд) независим
- Контракт Фаз 10/12 зафиксирован: каналы D1/D3/D4, kill-критерий строки, клип-политика D5, аборта D6 — переоткрытие только новой ревизией
- Требования TERM-01/TERM-04 остались Pending в REQUIREMENTS.md намеренно: shared-ID gate — 09-02 (TERM-01, TERM-04) и 09-03 (TERM-01) ещё не завершены; последний завершивший план отметит их
- Спайк-гейт честности: все Evidence-ячейки пусты, ни одно решение D1–D7 не предрешено — веб-источники не цитируются

---
*Phase: 09-spayk-i-defekt-wezterm*
*Completed: 2026-10-08*

## Self-Check: PASSED

- docs/SPEC.md, docs/adr/ADR-008-terminal-transport.md, 09-01-SUMMARY.md — найдены на диске
- Коммиты ac676c4 (Task 1), 2a6e67a (Task 2) — найдены в git log
- Все десять план-гейтов верификации — PASS (5 grep-гейтов SPEC, 4 гейта ADR-008, диф-скоуп пяти секций)
- Зелёная итерация: go build / go vet / go test -race ./... / golangci-lint run — чисто
