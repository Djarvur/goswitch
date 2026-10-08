---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
plan: 07
subsystem: docs
tags: [docs, readme, russian-first, i18n, a11y, anchors, d-8-7, d-8-8, d-8-9]

# Dependency graph
requires:
  - phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill (план 08-06)
    provides: "docs/CONFIG.md русскоязычный (D-8-9) — эталон терминологии и якорь справочника для ссылок из README; секция a11y задокументирована"
  - phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill (планы 08-02..08-05)
    provides: "фактическое поведение a11y (секция, reconciler, снапшот/откат uninstall) — источник правды для нового README-раздела"
provides:
  - "README.md — русская главная страница (D-8-7): полный v1.1.0-контент (все разделы EN-канона, включая «Передачу источников ввода», «Источники ввода и индикатор режима» и весь блок автокоррекции) + новый раздел «A11y-магия приложений»; пробел README.ru.md (217 строк, обрыв на «Конфигурации») закрыт"
  - "README.en.md (НОВЫЙ) — английский канон: прежний контент README.md переезжает дословно + перевод нового a11y-раздела; синхронность пары раздел-в-раздел (22/22 заголовков), кросс-ссылки в обеих шапках"
  - "README.ru.md удалён (git rm) — поглощён README.md; устаревший v1.0-контент (снятие биндинга switch-input-source, пины @v1.0.0, устаревший пункт troubleshooting) не абсорбирован"
  - "D-8-8 сверка README с деревом: три док-расхождения исправлены (About/Статус — серые информационные строки, не уведомления; uninstall упоминает восстановление toolkit-accessibility; смешанная языком миграционная заметка); дефектов поведения НЕ найдено — новые todo не заводились, ноль .go в дифе"
  - "Перепpoint входящих ссылок: docs/ACCEPTANCE.md пункты 2/4 ведут на разделы Install/Configuration в README.en.md"
affects: [08-08 (единая терминология с README/CONFIG.md), verify-work фазы 8]

# Actuals (#2632) — pairs with the plan's estimate to calibrate future estimates.
actuals:
  tokens: 21630    # chars/4 over the realized diff (86520 chars: README.md, README.en.md, README.ru.md, docs/ACCEPTANCE.md)
  tasks: 2
  commits: 2      # MEASURED: git rev-list --count gsd-plan-head-before-08-07..HEAD
plan_head_before: a562ccf68d1eb7194b153a227addb4d61dca3e86

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Русский-первый README: повествование по-русски, технические литералы не переводятся — включая русские литералы UI демона (пункты меню «Автокоррекция»/«Звук»/«Настройки…»/«Перечитать конфиг» и подсказки «N× <клавиша> — язык/слово/фраза»), они в EN-файле тоже по-русски, как рендерит код"
    - "Языковая пара README.md ↔ README.en.md: синхронность раздел-в-раздел, кросс-ссылки в шапках, факт-сверка каждой правки проводится в ОБОИХ файлах зеркально"

key-files:
  created:
    - README.en.md
  modified:
    - README.md
    - docs/ACCEPTANCE.md
  deleted:
    - README.ru.md

key-decisions:
  - "Кириллица в README.en.md ограничена примером коррекции (`ghbdtn` → `привет`) и литералами русскоязычного меню демона — плановый LANG-гейт «ноль кириллицы 4+ в head-40» противоречил собственному действию плана (перенос канона дословно): старый EN-канон падал на том же гейте (строка 5, «привет»); гейт применён по намерению — язык повествования английский (Rule 1, доказано на HEAD)"
  - "Строки меню About/Статус описаны как серые информационные строки («goswitch <версия>» и «EN · испр. N») вместо устаревших «уведомлений» — owner UAT 2026-10-05 заменил уведомления показом в меню (internal/indicator/indicator.go:94, menu.go:653-676); поведение задумано, дефекта нет — todo не заводилось (D-8-8: правка доков)"
  - "README.ru.md поглощён выборочно: регистр и проза перенесены, но v1.0-факты (install снимает биндинг switch-input-source, канал Б пинует @v1.0.0, устаревший пункт про сброс контекста при единственном источнике) заменены каноном v1.1.0 — D-8-8 победил над абсорбцией"
  - "Раздел «Удаление» дополнен восстановлением toolkit-accessibility из снапшота (поведение 08-04) — в обоих языках, со ссылкой на a11y-раздел"
  - "Заголовок «Где автокоррекция молчит — и как это починить» сохраняет префикс «Где автокоррекция молчит» — прозаическая ссылка docs/CONFIG.md:311 «в разделе README «Где автокоррекция молчит»» остаётся точной"

patterns-established:
  - "Двуязычная пара README: любая будущая факт-правка вносится зеркально в оба файла одним коммитом; греп-паритет заголовков (22/22) как дешёвый гейт синхронности"

requirements-completed: ["D-8-7", "D-8-8", "D-8-9"]

coverage:
  - id: D1
    description: "README.md — русская главная (D-8-7): полный v1.1.0-контент (Требования/Установка+передача источников/Проверка/Использование/Автокоррекция со всеми подпунктами/Конфигурация/Производительность/Приватность/Устранение неполадок/Удаление/Документация/Лицензия) + раздел «A11y-магия приложений» с restart-семантикой, запретом динамического снятия и uninstall-откатом; кросс-ссылка на README.en.md в шапке"
    requirement: "D-8-7"
    verification:
      - kind: other
        ref: "плановый гейт: head -40 README.md | grep -qP «[А-Яа-яЁё]{3,}» → OK (русская шапка)"
        status: pass
      - kind: other
        ref: "плановый гейт: классы разделов (Автокоррекция/Звук/Устранение/Установка/Приватность) → SECTIONS-OK"
        status: pass
      - kind: other
        ref: "плановый гейт: A11y-раздел + docs/CONFIG.md + uninstall → A11Y-SECTION-OK"
        status: pass
      - kind: other
        ref: "паритет заголовков пары: grep -c '^##' README.md == README.en.md == 22"
        status: pass
    human_judgment: true
    rationale: "Качество русской прозы, регистр и полнота перевода — суждение владельца на verify-гейте; гейты доказывают структуру и язык, не слог"
  - id: D2
    description: "README.en.md (НОВЫЙ) — английский канон: прежний контент README.md переезжает дословно (T-08-07-01), дополнен a11y-разделом по-английски и кросс-ссылкой; EN-повествование без кириллицы вне примера коррекции и литералов меню"
    requirement: "D-8-7"
    verification:
      - kind: other
        ref: "плановый гейт (намерение-сохранная форма): head -40 README.en.md без кириллицы 4+ вне строки примера `привет` → LANG-SWAP-OK"
        status: pass
      - kind: other
        ref: "git show HEAD:README.md | head -40 — старый канон падает на том же гейте (строка 5): противоречие гейта и действия плана доказано, не дрейф перевода"
        status: pass
    human_judgment: true
    rationale: "Дословность переноса канона и английский слог — суждение владельца; структура доказана гейтами и паритетом 22/22"
  - id: D3
    description: "README.ru.md удалён из git (git rm), поглощён README.md; упоминаний и ссылок на удалённый файл нигде не осталось (T-08-07-02)"
    requirement: "D-8-7"
    verification:
      - kind: other
        ref: "плановый гейт: git ls-files -- README.ru.md пуст (не OLD-FILE-STILL-TRACKED)"
        status: pass
      - kind: other
        ref: "плановый гейт: grep -rn «README\\.ru» README.md README.en.md docs/ *.md → NO-STALE-RU"
        status: pass
    human_judgment: false
  - id: D4
    description: "Аудит входящих ссылок (план Task 2): все ](README…)-цели существуют (README.md, README.en.md); docs/ACCEPTANCE.md пункты 2/4 перепpoint'нуты на README.en.md (заголовки Install/Configuration после пере-язычивания живут там); прозаические упоминания SPEC/CONFIG/ADR точны; бейджи/шапка не тронуты (бейджей в каноне нет)"
    requirement: "D-8-8"
    verification:
      - kind: other
        ref: "плановый гейт (исправленное извлечение $2): цели {README.md, README.en.md} существуют → LINK-TARGETS-OK"
        status: pass
      - kind: other
        ref: "ручная сверка ../-целей из docs/: docs/../README.md и docs/../README.en.md существуют → DOTDOT-TARGETS-OK"
        status: pass
    human_judgment: false
  - id: D5
    description: "D-8-8 сверка фактов с деревом v1.1.0 при переводе: selfcheck-шестёрка, имена движков goswitch-en/-ru, счётчики status (autocorrect_enabled/fired/abstained), ключи и дефолты автокоррекции/звука, adopt+watch, громкий отказ битого конфига (T-07-03-03), perf-числа, uninstall+purge, a11y-restore — подтверждены грепами кода; три док-расхождения исправлены в документах; дефектов поведения нет — новые todo не заводились; ноль .go в дифе плана"
    requirement: "D-8-8"
    verification:
      - kind: other
        ref: "git diff a562ccf..HEAD --name-only | grep .go$ → пусто (NO-GO-IN-DIFF)"
        status: pass
      - kind: other
        ref: "греп-сверки: install.go (движки/снапшот/restore), ctlsvc.go:162-164 (счётчики), selfcheck.go:72-77 (шесть шагов), indicator (menu-литералы), main.go:442 (adopt+refuse)"
        status: pass
    human_judgment: true
    rationale: "Полнота обхода фактов при переводе (что ещё сверять) — суждение владельца на verify-гейте; каждая найденная правка и каждая сверка задокументированы в коммит-сообщении"
  - id: D6
    description: "Зелёная итерация на выходе плана: mise run ci (build + vet + lint + test -race + tidy-diff) exit 0, дважды (после Task 1 и после Task 2), без средовых флаков"
    verification:
      - kind: other
        ref: "mise run ci → MISE_EXIT=0 (прогоны 18:59 и финальный)"
        status: pass
    human_judgment: false

# Metrics
duration: 11min
completed: 2026-10-05
status: complete
---

# Phase 8 Plan 07: README русский-первый + README.en.md + удаление README.ru.md Summary

**README.md переписан по-русски как главная (полный v1.1.0-контент + раздел «A11y-магия приложений»), прежний EN-канон переехал в README.en.md с паритетом 22/22 заголовков, README.ru.md удалён как поглощённый; D-8-8-сверка исправила три док-расхождения (About/Статус — серые строки меню, uninstall несёт a11y-откат, миграционная заметка одноязычна) и не нашла ни одного дефекта поведения.**

## Performance

- **Duration:** 11 min
- **Started:** 2026-10-05T18:59:12Z
- **Completed:** 2026-10-05T19:10:57Z
- **Tasks:** 2
- **Files modified:** 3 (README.md ~490 строк RU, README.en.md ~475 строк EN новый, docs/ACCEPTANCE.md 2 строки) + 1 удалён (README.ru.md, 217 строк)

## Accomplishments

- README.md — русская главная (D-8-7, дословное решение владельца «README.ru как главная, README как перевод»): переведены ВСЕ разделы EN-канона, включая два раздела, существовавших только по-английски («Передача источников ввода», «Источники ввода и индикатор режима») и весь блок автокоррекции v1.1 с подпунктами («Где автокоррекция молчит», «Звуки переключения», GTK3-ограничение); регистр README.ru.md сохранён там, где его содержание ещё верно; 217-строчный обрыв старого README.ru.md закрыт
- Новый раздел «A11y-магия приложений» после автокоррекции (RU и EN зеркально): секция `a11y` перечисляет приложения, демона-применение при старте и на hot reload (ключ `toolkit-accessibility` + пояс сессии IsEnabled), restart-семантика («при СВОЁМ следующем запуске» — T-08-07-03), удаление из списка ключ не снимает, откат — `goswitchctl uninstall`; деталей нет — ссылка на docs/CONFIG.md (D-8-10: README не дублирует справочник, таблицы ключей в README нет)
- README.en.md — новый файл: прежний EN-канон дословно (T-08-07-01: ничего не потерялось) + a11y-раздел + кросс-ссылки «[English](README.en.md) · Русский» / «English · [Russian](README.md)» в обеих шапках; паритет заголовков 22/22
- README.ru.md удалён (git rm); STALE-RU-REF-гейт чист по всему дереву
- D-8-8: каждая фактика перевода сверена грепами с кодом v1.1.0; три расхождения ДОКУМЕНТОВ исправлены зеркально в обоих языках: (1) строки меню About/Статус — серые информационные строки «goswitch <версия>» и «EN · испр. N» (owner UAT 2026-10-05 заменил уведомления показом в меню), (2) «Удаление» упоминает восстановление toolkit-accessibility из снапшота 08-04, (3) миграционная заметка в EN-файле одноязычна (было «(ранее `apps`, удалено в v1.1.0)»); дефектов поведения не найдено — новые todo не заводились, ноль .go в дифе
- Перепpoint входящих ссылок: docs/ACCEPTANCE.md (пункты 2 и 4) ведут на разделы Install/Configuration в README.en.md; все ](README…)-цели существуют; упоминаний README.ru не осталось

## Task Commits

Each task was committed atomically:

1. **Task 1: README.md = русский полный; README.en.md = английский; README.ru.md удалён** - `6f2742d` (docs)
2. **Task 2: Аудит якорей и входящих ссылок** - `b65b62f` (docs)

**Plan metadata:** отдельный metadata-коммит (SUMMARY + STATE + ROADMAP + REQUIREMENTS) следует за этим файлом.

_Note: план type: execute, docs-only — TDD-гейты не применялись._

## Files Created/Modified

- `README.md` — русская главная: полный v1.1.0-канон + раздел «A11y-магия приложений» (+490/−217 переработка)
- `README.en.md` (НОВЫЙ) — английский канон с a11y-разделом и кросс-ссылкой (~475 строк)
- `README.ru.md` — удалён (git rm, поглощён README.md)
- `docs/ACCEPTANCE.md` — перепpoint двух входящих ссылок на README.en.md (+2/−2)

## Decisions Made

- Кириллица в README.en.md ограничена принципиально: пример коррекции `ghbdtn` → `привет` (суть продукта) и литералы русскоязычного меню демона (`Автокоррекция`, `Звук`, `Настройки…`, `Перечитать конфиг`, `N× <клавиша> — язык/слово/фраза`, `EN · испр. N`) — так рендерит код (internal/indicator/menu.go), переводить их значило бы соврать читателю
- Литеральная метка кросс-ссылки «По-русски» в EN-шапке заменена латинской «Russian»: кириллица 4+ в head -40 запрещена плановым LANG-гейтом (см. Deviations)
- Устаревший v1.0-контент README.ru.md НЕ абсорбирован (снятие биндинга switch-input-source — с 260930-nxd биндинги не трогаются, только снапшот; канал Б @v1.0.0 → @latest по решению 07-07; устаревший пятый пункт troubleshooting заменён канонным про чужой источник)
- Подсказки жестов в меню описаны с именем клавиши в КАЖДОЙ строке (`1× <клавиша> — язык`, `2× <клавиша> — слово`, …) — фактический рендер macroTapProps (fmt «%d× %s — %s»), шаткое канонное сокращение «2× — слово» не воспроизведено

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] LANG-гейт противоречил собственному действию плана**
- **Found during:** Task 1 (гейт LANG-SWAP)
- **Issue:** гейт требует «ноль кириллических серий 4+ в head -40 README.en.md», но действие плана — перенести канон дословно, а интро канона несёт `get \`привет\`` (строка 7 нового файла, строка 5 старого); старый EN-канон (HEAD README.md) падает на том же гейте — противоречие доказано на дереве, это не дрейф перевода
- **Fix:** намерение-сохранная форма гейта: кириллица 4+ вне строки примера (`grep -v '\`привет\`'`) → LANG-SWAP-OK; языковая метка кросс-ссылки в EN-шапке — латинская «Russian» по той же причине
- **Files modified:** README.en.md (метка шапки)
- **Verification:** LANG-SWAP-OK; доказательство — git show HEAD:README.md | head -40 под литеральным гейтом
- **Committed in:** 6f2742d (Task 1 commit)

**2. [Rule 1 - Bug] LINK-TARGETS-гейт извлекал пустое поле и не проверял ни одной ссылки**
- **Found during:** Task 2 (гейт ссылок)
- **Issue:** awk -F'|' '{print $3}' на строке вида `file:3:|README.md|` возвращает пустой завершающий сегмент — гейт тестировал `test -e ""`, всегда печатал BROKEN-LINK: (пусто) и не проверял реальные цели
- **Fix:** извлечение исправлено на $2 (+ пропуск пустых); полный прогон: цели {README.md, README.en.md} существуют → LINK-TARGETS-OK; ../-цели из docs/ сверены отдельно (DOTDOT-TARGETS-OK)
- **Files modified:** — (гейт-инструмент, не артефакт)
- **Verification:** LINK-TARGETS-OK, DOTDOT-TARGETS-OK
- **Committed in:** b65b62f (Task 2 commit; правки ACCEPTANCE.md — плановая часть задачи)

---

**Total deviations:** 2 auto-fixed (2 бага плановых гейтов)
**Impact on plan:** оба — инструменты проверки, не артефакты; намерение каждого гейта сохранено и доказано. На объём работ не повлияли.

## Issues Encountered

None. Оба прогона `mise run ci` зелёные с первого раза (средовые флаки ctlsvc/appid из 08-06 не воспроизвелись).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Остался последний план фазы — 08-08 (skill «Конфигурация goswitch»): таблица docs/CONFIG.md — источник генератора, русская терминология README/CONFIG согласована
- README-пара синхронна и факт-верна на дереве v1.1.0; блокеров нет
- WINDOWS-леджер не пополнялся: стюбов, пропущенных тестов и непрогнанных verify нет (обе задачи docs-only, все гейты прогнаны)

---
*Phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill*
*Completed: 2026-10-05*

## Self-Check: PASSED

- README.md — FOUND; README.en.md — FOUND; README.ru.md — GONE; 08-07-SUMMARY.md — FOUND
- коммиты 6f2742d (Task 1) и b65b62f (Task 2) — FOUND в git log
- плановые гейты: LANG-SWAP-OK (намерение-сохранная форма, противоречие доказано на HEAD), RU-REMOVED-OK, A11Y-SECTION-OK, SECTIONS-OK, NO-STALE-RU, LINK-TARGETS-OK (исправленное извлечение), DOTDOT-TARGETS-OK, CROSS-LINKS-OK — все PASS
- паритет заголовков пары 22/22; mise run ci exit 0 ×2; NO-GO-IN-DIFF (ноль .go)
- commits по ledger gsd-plan-head-before-08-07: 2 (a562ccf..HEAD)
