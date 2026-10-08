---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
plan: 10
subsystem: docs
tags: [readme, a11y, documentation, bool-contract, gap-closure, parity]

requires:
  - phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
    provides: "bool-контракт a11y-секции из дельты 08-09 (pointer-bool default ON, громкая миграция D-33) и канонические формулировки docs/CONFIG.md «Секция a11y: применение и откат» (:227–262)"
provides:
  - "README.md a11y-раздел в bool-семантике ревизии владельца 2026-10-06 (RU, зеркало docs/CONFIG.md)"
  - "README.en.md — секционно-параллельный английский раздел (D-8-7), паритет заголовков 22/22"
  - "строки 08-10 T1..T3 в Per-Task Verification Map 08-VALIDATION.md"
  - "grep-гейт ключа-списка a11y.apps для обоих README — живёт в гейтах T1/T2/T3 этого плана"
affects: ["re-verification фазы 8 (43/44 → ожидается 44/44)", "будущие правки README (негативный grep-гейт ключа-списка)"]

actuals:
  tokens: 3388
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "README-разделы зеркалят канонический docs/CONFIG.md по порядку фактов, без пересказа синонимами (D-8-8)"
    - "Скоуп-гейт README-правок: awk по хункам git diff -U0 от тега gsd-plan-head-before-08-10, границы прежних разделов (RU 301–320, EN 296–314)"

key-files:
  created: []
  modified:
    - README.md
    - README.en.md
    - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-VALIDATION.md

key-decisions:
  - "README-проза зеркалит docs/CONFIG.md «Секция a11y: применение и откат» по порядку фактов: один глобальный ключ → механика двух ручек + идемпотентность → active-семантика (default ON) → restart/отказ от снятия/uninstall-revert → громкая миграция; без второго ключа в завершающей ссылке"
  - "Маркер (D-8-4) оставлен в прозе README по дословному тексту действия плана, хотя README более нигде D-refs не использует"
  - "EN «next start» в нижнем регистре (вместо прежнего NEXT start) — продиктовано собственным case-sensitive grep-гейтом плана (next (start|launch))"
  - "Постоянный grep-гейт ключа-списка для будущих правок README живёт в гейтах этого плана; вынос в .github/workflows — вне скоупа гэп-плана (files_modified не включают CI-файлы) — зафиксировано как optional-hardening заметка"

requirements-completed: [D-8-8, D-8-7]

coverage:
  - id: D1
    description: "README.md a11y-раздел переписан на bool-контракт: единственный ключ a11y.enabled (default ON — отсутствующая секция/ключ читаются ВКЛ; явное enabled: false — единственный OFF, D-8-4), глобальная механика (toolkit-accessibility + пояс org.a11y.Status.IsEnabled), идемпотентный reconcile при старте и hot reload, restart-семантика, uninstall-revert, громкая миграция; ноль a11y.apps и ноль прозы списка/сопоставления"
    requirement: D-8-8
    verification:
      - kind: other
        ref: "command: T1 grep-гейт (негативные stale-фразы + якоря a11y.enabled/toolkit-accessibility/IsEnabled/enabled: false/goswitchctl uninstall/следующее запуск/«единственный ключ») → RU-A11Y-BOOL-OK"
        status: pass
    human_judgment: false
  - id: D2
    description: "README.en.md несёт тот же контракт по-английски, секция-в-секцию параллельно README.md; паритет заголовков 22/22 в обоих файлах; кросс-ссылки шапок ([English](README.en.md) / [Russian](README.md)) целы; ноль a11y.apps"
    requirement: D-8-7
    verification:
      - kind: other
        ref: "command: T2 grep-гейт → EN-A11Y-BOOL-OK (включая test \"$(grep -cE '^##+ ' …)\" -eq 22 в обоих файлах)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Гейты закрытия гэпа: скоуп-дифф обоих README целиком внутри прежних границ a11y-разделов (RU старые 301–320, EN старые 296–314), git diff --name-only = только README.md/README.en.md/08-VALIDATION.md, skillgen -check exit 0, mise run ci exit 0, строки 08-10 T1..T3 в 08-VALIDATION.md"
    requirement: D-8-8
    verification:
      - kind: other
        ref: "command: T3-гейт → GAP-README-OK; mise run ci exit 0 (build + vet + golangci-lint + test -race, все пакеты ok, первый прогон)"
        status: pass
    human_judgment: false

duration: 12 min
completed: 2026-10-06
status: complete
plan_head_before: 9660977af76da2acdfdc7777d20c03659dd3c2ab
commits: 3
---

# Phase 8 Plan 10: README a11y-разделы → bool-контракт (гэп-закрытие) Summary

**README.md и README.en.md a11y-разделы переведены на bool-контракт ревизии владельца 2026-10-06 (единственный ключ `a11y.enabled`, default ON, без ключа-списка) с сохранением паритета 22/22 — единственный гэп пост-дельта реверификации закрыт, D-8-8 снова SATISFIED.**

## Performance

- **Duration:** 12 min
- **Started:** 2026-10-06T05:28:01Z
- **Completed:** 2026-10-06T05:42:05Z
- **Tasks:** 3
- **Files modified:** 3 (README.md, README.en.md, 08-VALIDATION.md)

## Accomplishments
- a11y-раздел README.md описывает секцию честно по ревизии 2026-10-06: один глобальный переключатель `a11y.enabled` (bool, по умолчанию включён; отсутствующая секция/ключ читаются ВКЛ, явное `enabled: false` — единственный OFF, демон тогда ключ не трогает — D-8-4), глобальность обоснована общим для стола `toolkit-accessibility`, идемпотентный reconcile при старте и на каждой горячей перезагрузке (ключ, уже равный `true`, не переписывается — ноль записей в dconf), restart-семантика, запрет снятия, uninstall-revert из снапшота установки, громкая миграция строки списка
- README.en.md — секционно-параллельный английский зеркало-раздел (те же четыре блока в том же порядке, D-8-7); паритет заголовков 22/22 в обоих README, кросс-ссылки шапок целы
- Все упоминания ключа-списка `a11y.apps` и прозы сопоставления приложений удалены из обоих README (grep-гейты T1/T2/T3); завершающая ссылка в обоих языках называет «единственный ключ» и ведёт в docs/CONFIG.md
- Скоуп-гейт: весь дифф плана — 3 файла (README.md, README.en.md, 08-VALIDATION.md); хунки README целиком внутри прежних границ a11y-разделов (RU 301–320, EN 296–314)
- Генераторная цепочка не тронута и не требует регенерации: skillgen читает ТОЛЬКО docs/CONFIG.md (`configPath`, cmd/skillgen/main.go:31) — README генератору не подвластны; `go run ./cmd/skillgen -check` exit 0
- `mise run ci` exit 0 с первого прогона (build + vet + golangci-lint + test -race; все пакеты ok — известных флаков ctlsvc/watch-debounce/appid сегодня не наблюдалось)
- Строки 08-10 T1..T3 дописаны в Per-Task Verification Map 08-VALIDATION.md — гэп-задача прослеживаема

## Task Commits

Each task was committed atomically:

1. **Task 1: README.md a11y-раздел → bool-контракт (RU)** - `d1e44a4` (docs)
2. **Task 2: README.en.md → тот же контракт по-английски, паритет** - `8df7162` (docs)
3. **Task 3: Гейты закрытия гэпа + VALIDATION-строки 08-10 T1..T3** - `9174109` (docs)

**Plan metadata:** см. финальный docs-коммит этого SUMMARY.

## Files Created/Modified
- `README.md` — a11y-раздел (:301–334, старые 301–320) переписан на bool-контракт; остальное дословно
- `README.en.md` — a11y-раздел (:298–327, старые 298–314) — английское зеркало; остальное дословно
- `.planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/08-VALIDATION.md` — строки 08-10 T1..T3 в Per-Task Verification Map

## Decisions Made
- Проза зеркалит docs/CONFIG.md «Секция a11y: применение и откат» (:227–262) по порядку фактов, без изобретения синонимов (D-8-8); завершающая ссылка в обоих языках называет «единственный ключ `a11y.enabled`»
- Маркер (D-8-4) оставлен в прозе README по дословному действию плана
- EN «next start» в нижнем регистре — продиктовано case-sensitive grep-гейтом плана
- Optional hardening (постоянный grep-гейт ключа-списка в CI) остаётся в гейтах этого плана; вынос в .github/workflows — вне скоупа гэп-плана

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] awk-гейт скоуп-диффа исполнен в mawk-совместимой форме**
- **Found during:** Task 3 (скоуп-гейт хунков)
- **Issue:** awk плана предполагал, что `split(" -301,20", a, " ")` оставит `a[1]=""` и возьмёт старт из `a[2]`; mawk (системный awk этой машины, 1.3.4) при `split(..., " ")` схлопывает ведущий пробел → `a[1]=старт, a[2]=количество, a[3]=""` → гейт читал s=количество, c=0 и ложно падал на ЛЮБОМ хунке (проверено: `a[1]=[306] a[2]=[6] a[3]=[]`); семантика гейта при этом не нарушена — реальные хунки в границах (RU 306–311 и 314–319 ⊂ 301–320; EN 301–313 ⊂ 296–314)
- **Fix:** тот же гейт с диалект-независимым разбором диапазона: `r=$2; sub(/^-/,"",r); split(r,a,",")` — старые координаты хунков проверяются дословно по исходной семантике (s=старт, c=количество)
- **Files modified:** файлы репо не менялись (реализация гейта; команда в mawk-совместимой форме зафиксирована в строке 08-10 T3 карты 08-VALIDATION.md)
- **Verification:** RU-SCOPE-OK / EN-SCOPE-OK; полный T3-гейт → GAP-README-OK
- **Committed in:** 9174109 (в строке карты)

---

**Total deviations:** 1 auto-fixed (1 blocking — диалект awk в verify-команде плана)
**Impact on plan:** единственная правка — форма исполнения гейта; семантика всех гейтов и скоуп плана соблюдены дословно.

## Issues Encountered
- `mise run ci` зелёный с первого прогона — известные флаки (ctlsvc, watch-debounce, appid — см. deferred-items.md фазы) не воспроизводились; в deferred-items ничего не добавлено.
- Косметика: тег `gsd-plan-head-before-08-10` и одноимённый ledger-файл `.git/gsd-plan-head-before-08-10` (протокол #3968) дают stderr-предупреждение «refname is ambiguous» — оба указывают на один коммит 9660977, все пути разрешения (`git rev-parse --verify`, `cat` ledger, тег) возвращают один SHA; оба артефакта требуются протоколами, ничего не предпринималось.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Провалившаяся истинность 08-07 #4 восстановлена: a11y-раздел документирует раздел честно (restart-семантика + uninstall-откат), без противоречия схеме (единственный ключ `a11y.enabled`, default ON)
- D-8-8 снова SATISFIED: расхождение док↔схема устранено в обоих языках; пользователь, следующий README, получает валидный конфиг
- Скор реверификации ожидается 43/44 → 44/44 — фаза 8 (последняя в milestone v1.0.0) готова к повторному `/gsd:verify-work`
- Блокеров нет

---
*Phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill*
*Completed: 2026-10-06*

## Self-Check: PASSED

- Files exist: README.md, README.en.md, 08-VALIDATION.md — FOUND (все три в коммитах d1e44a4 / 8df7162 / 9174109)
- Commits exist: d1e44a4, 8df7162, 9174109 — FOUND in `git log --all`
- Acceptance re-run: T1-гейт RU-A11Y-BOOL-OK, T2-гейт EN-A11Y-BOOL-OK, T3-гейт GAP-README-OK, `mise exec -- go run ./cmd/skillgen -check` exit 0, `mise run ci` exit 0, паритет 22/22, скоуп-дифф 3 файла — все PASS
- `git rev-list --count gsd-plan-head-before-08-10..HEAD` = 3 (совпадает с фактическими коммитами задач)
