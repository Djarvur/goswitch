---
phase: 13-a11y-retsepty-telegram-zcode
plan: "03"
subsystem: docs
tags: [a11y, docs, recipes, zcode, telegram, selfcheck, residual-risk, a11y-03]

requires:
  - phase: 13-a11y-retsepty-telegram-zcode (13-01: ряды пробы трёх сборок Telegram, вердикт владельца «Задокументировать», плашка диктора A5; 13-02: ряды «a11y-проба zcode r1/r2 red|green» — роль 79, RED→GREEN ×2, три находки установки, контракт selfcheck 0/1/2)
  - scripts/a11y-selfcheck.py (контракт D-06 из 13-01, живая матрица из 13-02)
provides:
  - A11Y-03 закрыт: рецепты ZCode/Telegram в README.md + README.en.md (кратко) и docs/CONFIG.md (полностью) — по датированным рядам пробы, без опережения машины
  - строка самопроверки полного вида /usr/bin/python3 scripts/a11y-selfcheck.py <app> в каждом рецепте (D-06/SC4) с расшифровкой кодов 0/1/2
  - residual-risk заметка «ручная коррекция на скрытых полях копирует в PRIMARY» в обоих README (D-07/SC3)
  - уточнение «что магия НЕ делает» в разделах «A11y-магия приложений» обоих README
affects: [phase-verification, /gsd:ship docs gate]

actuals:
  tokens: 4006        # chars/4 over the realized diff 45acde7..HEAD (171 insertions, 3 doc files)
  tasks: 2
  commits: 2          # MEASURED: git rev-list --count gsd-plan-head-before-13-03..HEAD
plan_head_before: 45acde7836e81088cfd22185ccfb961c0a71fd03

tech-stack:
  added: []
  patterns:
    - "probe-sourced-docs: каждое фактическое утверждение рецепта ссылается на ряд-первоисточник журнала ADR-008 (роль — числом из зонда, дата пробы, форма плашки ровно как наблюдена)"
    - "selfcheck-one-liner: применил → перезапустил → /usr/bin/python3 scripts/a11y-selfcheck.py <app> → код 0/1/2; никогда «goswitchctl selfcheck» и никогда голый a11y-selfcheck"

key-files:
  created:
    - ".planning/phases/13-a11y-retsepty-telegram-zcode/13-03-SUMMARY.md"
  modified:
    - "docs/CONFIG.md (+98/-1: раздел «Ручной рецепт» расширен подразделами ZCode (валидация 2026-11-03) и Telegram (задокументированная граница); предупреждения дополнены flatpak-пунктом)"
    - "README.md (+34: краткие рецепты + residual-risk PRIMARY в «Где автокоррекция молчит»; «что магия НЕ делает» в «A11y-магии приложений»)"
    - "README.en.md (+39: байт-параллельное зеркало тех же абзацев)"

key-decisions:
  - "Telegram-рецепт написан по ФАКТИЧЕСКОЙ ветке итог-ряда — вердикт владельца «Задокументировать» (вариант 2): команд установки flatpak в рецептах НЕТ (плановая ветка D-04 опровергнута пробой); рецепт = граница: переключение зелёное во всех сборках, автокоррекция недостижима ни в одной (стаб даже unconfined — потолок Qt6-моста билда), смена формата не лечит"
  - "ZCode-рецепт несёт все три обязательных факта ряда «a11y-механика zcode install» (chmod 4755 chrome-sandbox при AppArmor-blocked userns, самоперезапись override при каждом старте с консервативной формой «проверяйте после запусков/обновлений», перезапись argv/environ → подтверждать флаг самопроверкой, не /proc) и честную границу слоя 3: дерево живо, автокоррекция молчит (мост не эмитит identity-события → skip role-unknown)"
  - "Плашка диктора задокументирована ровно в наблюдённой форме: «Telegram is working in Screen Reader mode» + кнопка Disable, живьём на flatpak и tarball (2026-11-03), на snap не наблюдена — ожидаемый побочный эффект, не поломка (A5 закрыта)"
  - "Уточнение «что магия НЕ делает» сформулировано probe-честно: глобальный ключ не заменяет флаг конкретного Electron-приложения и не открывает дерево, которое сборка не экспонирует (Telegram — во всех формах, даже unconfined); формулировка плана «не снимает confinement» уточнена по вердикту пробы (confinement — не потолок)"

requirements-completed: [A11Y-03]

coverage:
  - id: D1
    description: "docs/CONFIG.md несёт полный ZCode-рецепт с живой валидацией (override-шаблон :385-392, полный выход — трей жив, «флаги Electron применяются только при запуске», selfcheck полного вида с кодами 0/1/2, факт валидации датирован 2026-11-03, роль поля — числом 79) и Telegram-рецепт по вердикту итог-ряда"
    requirement: A11Y-03
    verification:
      - kind: unit
        ref: "verify задачи 1: grep-гейты (selfcheck-строка, диктор, «применяются только при запуске», трей, «a11y-проба zcode», отсутствие gail:atk-bridge) + git diff --exit-code -- internal/ cmd/ + mise run ci — все зелёные"
        status: pass
    human_judgment: false
  - id: D2
    description: "Оба README несут краткие рецепты ZCode (force-renderer-accessibility + полный перезапуск + граница) и Telegram (по вердикту), строку self-check полного вида с именем zcode/telegram, residual-risk PRIMARY-заметку и «что магия НЕ делает» — байт-параллельно в одном коммите"
    requirement: A11Y-03
    verification:
      - kind: unit
        ref: "verify задачи 2: grep-гейты по обоим README (флаг, полный путь selfcheck, PRIMARY) + отсутствие gail:atk-bridge + mise run ci + git diff --exit-code -- go.mod go.sum — все зелёные; counts «goswitchctl selfcheck» pre/post идентичны (см. Deviations)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Плашка диктора покрыта темой ровно в наблюдённой форме (flatpak/tarball — живьём, snap — не наблюдена); границы только-на-запуске/трей/маскировка обновлений присутствуют; существующие предупреждения :394-405 сохранены и расширены flatpak-пунктом"
    requirement: A11Y-03
    verification:
      - kind: unit
        ref: "grep 'диктор' docs/CONFIG.md; текст ссылается на скриншотные строки рядов flatpak r1/tarball r1 и snap-механику (ADR-008); git show 2940ef7 — предупреждения на месте"
        status: pass
    human_judgment: false

duration: 12min
completed: 2026-11-03
status: complete
---

# Phase 13 Plan 03: Рецепты Telegram/ZCode в доках (D-07, A11Y-03) Summary

**Рецепты перенесены в доки ровно по датированным рядам пробы: CONFIG.md — полный ZCode-рецепт (валидация 2026-11-03, роль поля 79, RED→GREEN ×2, три факта установки, честная граница «мост не эмитит identity-события») и Telegram-граница по вердикту владельца «Задокументировать» (без выдуманного зелёного — команд установки flatpak нет); оба README — краткие рецепты + residual-risk PRIMARY + «что магия НЕ делает», байт-параллельно; self-check полного вида в каждом рецепте; демон и go.mod нетронуты, ci зелёный.**

## Performance

- **Duration:** 12 мин (18:14-18:26 UTC, включая два прогона mise run ci)
- **Tasks:** 2/2 (docs-only, без ВМ — по плану)
- **Files:** 3 репозиторных (+171/-1): docs/CONFIG.md, README.md, README.en.md

## Accomplishments

- **Задача 1 (docs/CONFIG.md):** раздел «Ручной рецепт: per-app запуск с a11y-флагами» расширен подразделами: ZCode — проверенная форма (3 шага: override по шаблону :385-392 → полный выход из процесса с пояснением «трей жив, флаги Electron применяются только при запуске» → самопроверка `/usr/bin/python3 scripts/a11y-selfcheck.py zcode` с кодами 0/1/2), факт валидации (2026-11-03, оба раунда: 2 ноды/exit 1 без флага → 43+ нод/exit 0/роль 79 с флагом; поле API-ключа — роль 40), две особенности сборки (самоперезапись override 2/2 запусков; перезапись argv/environ), честная граница слоя 3 со ссылкой на клаузу «Где слой молчит»; Telegram — задокументированная граница по итог-ряду и вердикту владельца (переключение зелёное во всех трёх сборках; автокоррекция недостижима — стаб роли 75 даже unconfined; смена формата не открывает; самопроверка ожидаемо код 1), плашка диктора ровно как наблюдена (flatpak/tarball живьём, snap — нет); предупреждения :394-405 сохранены, добавлен соседний пункт про flatpak-dbus-прокси.
- **Задача 2 (README.md + README.en.md):** в «Где автокоррекция молчит»/«Where autocorrect stays silent» — проверенные краткие рецепты ZCode и Telegram (суть + selfcheck-строка полного вида с именем zcode/telegram + ссылка на docs/CONFIG.md) и отдельный абзац residual-risk: ручная коррекция на скрытых полях кладёт исправленный текст в PRIMARY-выделение (D-07/SC3); в «A11y-магия приложений»/«App accessibility magic» — уточнение «что магия НЕ делает» (глобальный ключ ≠ флаг конкретного Electron-приложения; не открывает дерево, которое сборка не экспонирует) с self-check одной командой; оба файла — одними и теми же по структуре абзацами в одном коммите, застарелые расхождения зеркала (canberra :261-262 RU/EN) не тронуты.

## Task Commits

1. **Задача 1: CONFIG.md — ZCode-рецепт + Telegram-граница + побочки/границы + self-check** — `2940ef7` (docs)
2. **Задача 2: README ×2 — краткие рецепты + PRIMARY + «что магия НЕ делает»** — `bce06c5` (docs)

**Plan metadata:** финальный docs-коммит (SUMMARY+STATE+ROADMAP+REQUIREMENTS) — далее.

## Files Created/Modified

- `docs/CONFIG.md` — раздел «Ручной рецепт» +98/-1 (подразделы ZCode/Telegram, flatpak-предупреждение)
- `README.md` — +34 (рецепты + residual-risk + «что магия НЕ делает»)
- `README.en.md` — +39 (зеркало, те же абзацы по-английски)

## Decisions Made

См. key-decisions. Ядро: Telegram-рецепт = граница (не установка flatpak — проба опровергла D-04); ZCode-рецепт = проверенная форма со всеми фактами ряда установки; плашка диктора — в наблюдённой форме; все самопроверки — полным путём.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Check-form] Литеральный grep-гейт «goswitchctl selfcheck» задачи 2 невыполним на предсодержимом репо**
- **Found during:** Задача 2 (verify)
- **Issue:** гейт `! grep -q 'goswitchctl selfcheck' README.md README.en.md docs/CONFIG.md` падает на ПРЕСУЩЕСТВУЮЩИХ упоминаниях: README.md ×5, README.en.md ×5 — все о валидации конфига/IBus-аудите (реальная субкоманда `goswitchctl selfcheck`), ни одно не в контексте a11y-зонда; у плана собственная клауза верификации запрещает именно «в контексте a11y-зонда»
- **Fix:** гейт исполнен в форме плана-намерения: counts pre-edit vs post-edit идентичны (5/5/0 — ни одного нового упоминания), все новые строки самопроверки — только полный вид `/usr/bin/python3 scripts/a11y-selfcheck.py <app>`; предсуществующие валидационные упоминания — вне фазы (scope boundary), удалены не были (они корректны)
- **Files modified:** нет (только форма проверки)
- **Commit:** bce06c5 (проверка при коммите)

**Total deviations:** 1 автофикс (форма проверки, не продукта). **Impact:** контракт запретов соблюдён в намеренной форме; ни одно утверждение рецептов не изменено.

## Issues Encountered

- None сверх отклонения выше.

## Known Stubs

None — рецепты несут только факты, доказанные рядами пробы; Telegram-граница — задокументированное ограничение по вердикту владельца, не стаб.

## User Setup Required

None — доки самодостаточны; стенд не нужен и не тронут (запаркован по 13-02).

## Next Phase Readiness

- Фаза 13 укомплектована: A11Y-01 (13-02), A11Y-02 (13-01, вердикт владельца), A11Y-03 (этот план) — все три требования фазы закрыты; фазовая верификация имеет: grep-гейты зелёные по трём файлам, ci зелёный, демон/go.mod нетронуты (git diff пуст по internal/ cmd/ go.mod go.sum от базы фазы)
- Flagged assumption плана (A11Y-03 unclassified/unresolved — SPEC без ## Edge Coverage): обязательные пункты выражены behavior-строками и grep-гейтами этого плана (selfcheck полного вида, PRIMARY, диктор, только-на-запуске/трей, restart); формальная классификация — за фазовой верификацией, строки — не тихий pass
- Для верификатора: плашка диктора A5 закрыта живьём (flatpak/tarball — наблюдена, snap — честное «не наблюдала»); ветка Telegram-рецепта = ветка итог-ряда (никаких команд установки)

## Self-Check: PASSED

SUMMARY существует; оба коммита от базы плана 45acde7 верифицированы в git log (2940ef7 / bce06c5; ledger gsd-plan-head-before-13-03 на диске, rev-list count = 2); файлы модифицированы (docs/CONFIG.md, README.md, README.en.md) — все FOUND; grep-гейты A11Y-03 зелёные по трём файлам (selfcheck полного пути, PRIMARY ×2, диктор/трей/«применяются только при запуске»/«a11y-проба zcode» в CONFIG.md, force-renderer-accessibility ×2, gail:atk-bridge отсутствует); mise run ci зелёный дважды (после каждой задачи); демон и go.mod/go.sum нетронуты (git diff пуст); stub-скан чист (TODO/FIXME/placeholder — ноль в добавленных строках).

---
*Phase: 13-a11y-retsepty-telegram-zcode*
*Completed: 2026-11-03*
