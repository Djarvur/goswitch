---
phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
plan: 07
subsystem: docs
tags: [release-readiness, readme, blocklist, apps-blocklist, menu-v2, sound, regression, d48, matrix-v4, flip-keystroke, dbusmenu, v1.1.0]

# Dependency graph
requires:
  - phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii (07-01..07-06, 07-08)
    provides: spec-delta §11 + ADR-007 amendment, blocklist schema+writer+adopt+watch, gate semantics, menu v2, sound engine, e2e репины + menu-v2/blocklist-silent кейсы, matrix v4
  - phase: 04-postavka-i-priemka
    provides: релизный поезд (тег → release.yml → goreleaser → ассеты), D-48 форма двойного прогона, docs/ACCEPTANCE.md
provides:
  - README v1.1.0: секция Autocorrect на blocklist-семантике (regex-подстрока, явная анкеровка, пустой ничего не запрещает, unknown-идентичность НЕ запрет, роль — фильтр безопасности, ручные жесты вне списка), меню v2 по каноническому макету (EN/RU, тумблеры «Автокоррекция»/«Звук», живые макро-подсказки, Настройки…, О программе, Статус, Перечитать конфиг), абзац звуков (тон любого флипа + отдельный тон автокоррекции, default ON, canberra/paplay, WARN-best-effort), adopt-следствие (-config — опция), одна миграционная фраза
  - .planning/phases/07-.../07-RELEASE-READINESS.md: доказательства (команда/дата/exit/HEAD), чек-лист точности доков (5/5 PASS), сверка 17/17 probe-строк с носителями, известные ограничения (7), post-merge шаги владельца (НЕ исполняются планом), статус RELEASE-READY
  - Доказанный регресс-набор: mise run ci зелёный ×2, flip-keystroke 24/24 (4 зелёных прогона), e2e-menu-v2 + e2e-autocorrect-blocklist-silent зелёные на HEAD плана, матрица v4 дважды подряд (31/34 в обеих — класс-строчный вердикт: 2 дизайн-строки WINDOWS #12 байт-равны актуалу реестра + по 1 мигрирующему средовому транзиенту)
  - Свидетельства: e2e-report-07-07-run1.txt / -run2.txt в каталоге фазы
affects: [verify-work phase 07 (UAT владельца: WINDOWS #12, fresh-session D-48, звук-UAT, Pitfall-1 скриншот, A1 редактор), new-milestone v1.1.0, post-merge релизный поезд владельца]

# Actuals (#2632) — same estimateTokens scale (chars/4 over the realized diff).
actuals:
  tokens: 8902        # chars/4 over the realized diff (35610 chars, 4 files, +382/−34)
  tasks: 3
  commits: 3          # MEASURED: git rev-list --count gsd-plan-head-before-07-07..HEAD (718d9d0..7fbe4ec)

# Tech tracking
tech-stack:
  added: []           # ноль новых зависимостей — план документально-регрессный
  patterns:
    - "Класс-строчный вердикт двойного прогона (метод 06-08): дизайн-замороженные строки / средовые транзиенты / несущие строки считаются раздельно; слово «зелёный» к полному exit-коду не применяется"
    - "Греп-гейт согласованности доков: множества ключей autocorrect.*/sound.* в README и CONFIG.md байт-равны (diff сорт-множеств)"

key-files:
  created:
    - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-RELEASE-READINESS.md
    - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/e2e-report-07-07-run1.txt
    - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/e2e-report-07-07-run2.txt
  modified:
    - README.md

key-decisions:
  - "Матрица v4 двойки — 31/34 в обеих: два красных — замороженные WINDOWS #12 строки (readback байт-в-байт равен документированному актуалу реестра), по одному мигрирующему средовому транзиенту SetGlobalEngine «Operation was cancelled» на прогон (разные кейсы, зелёные в соседнем прогоне); красные НЕ перегонялись в фальшивую зелень — зафиксированы как есть"
  - "Предусловие «свежесессия ≤30 мин» оценено НЕ выполненным (возраст 3477+ мин, записан); двойка исполнена по прямому указанию оркестратора и прецеденту 06-08/WINDOWS #13 — формальный fresh_session-гейт честно оставлен машинным инструментом владельца на verify-work, в доказательства НЕ заявлен"
  - "Канал B README переведён @v1.0.0 → @latest: точен и до, и после тега v1.1.0 — README не анонсирует несуществующий тег (prohibits) и не заставляет пользователей ставить устаревшую версию после релиза"
  - "requirements mark-complete — корректный no-op: строки фронматтера («CORR-01..09 (регресс)» и т.п.) — свободные формулировки, не REQ-ID REQUIREMENTS.md; MACR-ACL — фазовая метка; таблица трассировки обновляется на verify-work (прецедент 06-08/07-01)"

patterns-established:
  - "RELEASE-READINESS артефакт: доказательства (команда/дата/exit/HEAD + сохранённые отчёты) + чек-лист доков + реестр probe-строк с носителями + ограничения + post-merge шаги владельца — форма для следующих релизных планов"
  - "Сохранение отчётов матрицы с номером прогона в каталог фазы (e2e-report-07-07-runN.txt) — свидетельства переживают перезапись e2e-report.txt"

requirements-completed: ["CORR-01..09 (регресс)", "SWCH-01..04 (регресс)", "MACR-ACL (blocklist-ревизия)"]  # дословно из фронматтера плана; mark-complete no-op (не REQ-ID — см. key-decisions)

coverage:
  - id: D1
    description: "README соответствует v1.1.0 по машинным гейтам: apps_blocklist/меню/v1.1 присутствуют; множества autocorrect.* и sound.* байт-равны CONFIG.md; «Звук»+canberra описаны; тегов нет"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: other
        ref: "grep-гейты Task 1: README-OK / KEYS-MATCH (diff сорт-множеств autocorrect.*) / SOUND-DOC-OK (diff sound.* + Звук + canberra) / NO-TAG (git tag --list v1.1* пуст) — все PASS, повторены при Task 3"
        status: pass
    human_judgment: false
  - id: D2
    description: "Читательская адекватность README v1.1.0 (blocklist-семантика понятна пользователю; меню v2 описано по каноническому макету; видимый рендер меню и меток в трее)"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification: []
    human_judgment: true
    rationale: "Документная читабельность и видимый рендер меню (Pitfall 1 — метки, скриншот UAT; A1 — целевой редактор «Настроек…») не автоматизируются — маршрутизируется в UAT владельца на verify-work"
  - id: D3
    description: "Регресс-набор зелёный: mise run ci; flip-keystroke 24/24 (4 зелёных прогона, ноль deadline-abort); e2e-menu-v2 и e2e-autocorrect-blocklist-silent PASS на HEAD плана"
    requirement: "CORR-01..09 (регресс)"
    verification:
      - kind: e2e
        ref: "mise run ci → exit 0 ×2 (2026-10-05, HEAD a264a54, 19 пакетов -race; /tmp/07-07-ci.log)"
        status: pass
      - kind: e2e
        ref: "mise run e2e-flip-keystroke → PASS exit 0 ×4 (= 24/24 букв по счёту G-6-1, 6 букв × 4 прогона; /tmp/07-07-flip.log, /tmp/07-07-flip-run4-retry.log)"
        status: pass
      - kind: e2e
        ref: "mise run e2e-menu-v2 → PASS exit 0 (mode-записи EN/RU, оба тумблера persist+apply, пережил рестарт; /tmp/07-07-menu.log)"
        status: pass
      - kind: e2e
        ref: "mise run e2e-autocorrect-blocklist-silent → PASS exit 0 (поле дословно 0→7, fired=0, одна ac_skip_app_blocked; /tmp/07-07-blocklist.log)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Матрица v4 — двойной прогон D-48-формы подряд с полным префлайтом, отчёты сохранены; класс-строчный вердикт: 2 дизайн-строки WINDOWS #12 байт-равны актуалу, по 1 средовому транзиенту на прогон, несущие строки зелёные ×2; формальный свежесессионный гейт — НЕ заявлен"
    requirement: "SWCH-01..04 (регресс)"
    verification:
      - kind: e2e
        ref: "mise run e2e-matrix-v4 ×2 подряд → 31/34 exit 1 в обеих (штампы 05:59:07/06:13:35+03:00; e2e-report-07-07-run{1,2}.txt в каталоге фазы; diff прогонов = только штамп + позиция транзиента)"
        status: pass
      - kind: other
        ref: "байт-сверка readback красных строк с WINDOWS ledger #12: word-mixed → gfbghbdtn, phrase-mixed → ghbdtn ghbdtn — идентичны в обоих прогонах"
        status: pass
    human_judgment: true
    rationale: "Вердикт о транзиентах (средовые, не продукт) опирается на прецедент 06-08 и миграцию между кейсами, но формальный вердикт владельца остаётся за WINDOWS #12 и за machine-checkable свежесессионным гейтом fresh_session=true на verify-work — артефакт честно маршрутизирует оба туда"
  - id: D5
    description: "07-RELEASE-READINESS.md полный: Доказательства (команда/дата/exit/HEAD), чек-лист точности доков 5/5 PASS, реестр 17/17 probe-строк с носителями, 7 известных ограничений, post-merge шаги владельца (не исполняются планом), статус RELEASE-READY"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: other
        ref: "грейты Task 3: READINESS-OK (Доказательства + Post-merge + 17/probe) / REGISTRY-COVERED (CORR-0x|SWCH-0x ≥9 строк) / STATUS-OK / NOT-EXECUTED-OK"
        status: pass
    human_judgment: true
    rationale: "Артефакт — основание решения «теговать»: полноту и честность доказательств финально оценивает владелец перед пост-merge шагами (T-07-07-02 mitigate)"

# Metrics
duration: 55 min
completed: 2026-10-05
status: complete
---

# Phase 7 Plan 7: Релизная готовность v1.1.0 Summary

**README на blocklist-семантике с меню v2 и звуками, полный регресс-набор доказан живьём (ci, flip-keystroke 24/24, матрица v4 дважды с класс-строчным вердиктом, новые кейсы зелёные), RELEASE-READINESS с 17/17 probe-строками и post-merge шагами владельца — фаза в состоянии «мержь — и тегуй».**

## Performance

- **Duration:** 55 min (02:28–03:23 UTC; живые прогоны ~1h50m wall-time внутри)
- **Started:** 2026-10-05T02:28:09Z
- **Completed:** 2026-10-05T03:23Z
- **Tasks:** 3
- **Files modified:** 4 (+382/−34)

## Accomplishments

- README.md переведён на v1.1.0: секция Autocorrect описывает чёрный список `apps_blocklist` (regex-подстрока, явная анкеровка `^…$`, пустой список ничего не запрещает, неизвестная идентичность НЕ запрет — фильтр безопасности это ролевый гейт, ручные жесты вне списка), меню v2 по каноническому макету (EN/RU с отметкой, тумблеры «Автокоррекция» и «Звук» с персистом, живые макро-подсказки, «Настройки…», «О программе», «Статус», «Перечитать конфиг»), звуки (тон любого флипа + отдельный тон автокоррекции, default ON, canberra-gtk-play с paplay-фолбэком, WARN-best-effort), adopt-следствие (демон работает без `-config`; первый тумблер создаёт полный конфиг; битый файл — громкий отказ старта; явный `-config` помечен опцией); ровно одна миграционная фраза про удалённый белый список.
- Греп-гейты согласованности: множества ключей `autocorrect.*` и `sound.*` в README и docs/CONFIG.md байт-равны (KEYS-MATCH / SOUND-DOC-OK), `git tag --list 'v1.1*'` пуст (NO-TAG).
- Регресс доказан: `mise run ci` зелёный ×2 (19 пакетов, -race); flip-keystroke 24/24 буквы за 4 зелёных прогона, ноль deadline-abort WARN; e2e-menu-v2 и e2e-autocorrect-blocklist-silent перепрогнаны зелёными на HEAD плана (дерево ушло от 07-06 — по условию плана перепрогон, не ссылка).
- Матрица v4 — двойной прогон D-48-формы (подряд, префлайт 9/9 перед каждым): 31/34 в обеих; класс-строчный вердикт — 2 дизайн-красные WINDOWS #12 строки (readback байт-равен актуалу реестра в обоих прогонах) + по 1 мигрирующему средовому транзиенту активации (macr-super-letter в прогоне 1 / autocorrect-fires в прогоне 2, одна GDBus-сигнатура, зелёные в соседнем прогоне); отчёты сохранены как свидетельства фазы.
- 07-RELEASE-READINESS.md: Доказательства (команда/дата/exit/HEAD на каждый пункт), чек-лист точности доков 5/5 PASS (включая сверку меню README == menu.go по 13 меткам/15 ID), реестр 17/17 probe-строк (SPEC-S11×3, MACR-ACL, CORR-01..09, SWCH-01..04) с носителями, 7 известных ограничений, post-merge шаги владельца по поезду 04-07 (явно НЕ исполняются планом), статус RELEASE-READY.

## Task Commits

Each task was committed atomically:

1. **Task 1: README v1.1.0 — blocklist, меню v2, adopt** - `a264a54` (docs)
2. **Task 2: полный регресс-набор — доказательства** - `d21d6d6` (test)
3. **Task 3: RELEASE-READINESS — чек-лист, реестр, шаги владельца** - `7fbe4ec` (docs)

## Files Created/Modified

- `README.md` — v1.1.0: blocklist-семантика, меню v2 («Звук» включительно), звуки, adopt/-config-опция, миграционная фраза, канал B @latest
- `.planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-RELEASE-READINESS.md` (new) — доказательства, чек-лист доков, реестр 17 probe-строк, ограничения, post-merge шаги владельца, статус
- `.planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/e2e-report-07-07-run1.txt` / `-run2.txt` (new) — свидетельства двойного прогона матрицы

## Decisions Made

- Красные строки матрицы НЕ перегонялись в фальшивую зелень: двойка зафиксирована как есть (31/34 ×2), вердикт — класс-строчным методом 06-08 с байт-сверкой дизайн-строк против WINDOWS #12; транзиенты классифицированы по прецеденту (миграция между кейсами при одной сигнатуре, зелёные в соседнем прогоне двойки).
- Предусловие «свежесессия ≤30 мин» оценено по loginctl честно (НЕ выполнено, 3477+ мин) и зафиксировано в Доказательствах как ограничение: двойка — лучшее локальное свидетельство по прямому указанию оркестратора; формальный fresh_session-гейт — машинный инструмент владельца на verify-work.
- Канал B README: `@latest` вместо устаревающего `@v1.0.0` — одновременно выполняет prohibits (не анонсировать несуществующий тег) и не оставляет пользователей на старой версии после релиза.
- Перепрогон menu-v2/blocklist-silent вместо ссылки на 07-06: дерево ушло (docs-коммит плана) — условие плана «если дерево ушло — перепрогнать».

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Предусловие Task 2 «свежесессионный стол» не выполнено в форме ≤30 мин**
- **Found during:** Task 2 (префлайт-оценка, loginctl)
- **Issue:** сессия стола №264 стартовала 2026-10-02 19:32:39 MSK — возраст на момент двойного прогона 3477+ мин; перелогин недоступен исполнителю
- **Fix:** прогон исполнен по прямому указанию оркестратора (ожидаемый исход «32/34, 2 дизайн-строки — задокументировать, не подделывать») и прецеденту 06-08/WINDOWS #13; возраст сессии записан в Доказательствах; формальный fresh_session-гейт явно НЕ заявлен доказанным и маршрутизирован владельцу
- **Files modified:** 07-RELEASE-READINESS.md (разделы «Доказательства» п.4, «Известные ограничения» п.2)
- **Verification:** запись возраста + статус-формулировка критерия 4 «доказан в локальной форме двойки»
- **Committed in:** d21d6d6, 7fbe4ec

**2. [Rule 3 - Blocking] Литеральный гейт Task 2 «exit 0 обоих прогонов матрицы» несовместим с замороженными дизайн-строками**
- **Found during:** Task 2 (прогоны)
- **Issue:** драйвер матрицы выходит 1 при любом FAIL (matrix.go exitCode); строки word-mixed/phrase-mixed красные ПО ДИЗАЙНУ (WINDOWS #12, CASES-FROZEN) до вердикта владельца — литеральный гейт никогда бы не позеленел без подделки
- **Fix:** исход принят и задокументирован по санкции оркестратора (ожидаемый исход плана) и методу 06-08: класс-строчный вердикт с байт-сверкой против реестра; несущие строки зелёные ×2; фальшивая зелень запрещена guard-rail'ом
- **Files modified:** 07-RELEASE-READINESS.md (класс-строчный вердикт), e2e-report-07-07-run{1,2}.txt (сырые свидетельства)
- **Verification:** diff прогонов = штамп + позиция транзиента; обе дизайн-строки байт-равны актуалу реестра
- **Committed in:** d21d6d6

**3. [Rule 3 - Blocking] Канал B README нёс устаревающий пин @v1.0.0**
- **Found during:** Task 1 (сверка README с v1.1.0)
- **Issue:** после тега v1.1.0 инструкция `go install …@v1.0.0` ставила бы пользователям устаревшую версию; пин @v1.1.0 до мержа анонсировал бы несуществующий тег (prohibits)
- **Fix:** `@latest` — точен до и после тега
- **Files modified:** README.md
- **Verification:** KEYS-MATCH/SOUND-DOC-OK/README-OK не задеты; текст канала B согласован с пост-merge шагами RELEASE-READINESS
- **Committed in:** a264a54

---

**Total deviations:** 3 auto-fixed (3 × Rule 3 blocking: 1 precondition-deferral, 1 exit-semantics acceptance, 1 doc-accuracy)
**Impact on plan:** ни одного отступления от сути оракулов; все три — честная фиксация средовых/релизных реальностей, санкционированные прецедентами фаз 4-6 и прямым указанием оркестратора.

## Issues Encountered

- **flip-keystroke прогон 4 — watchdog 3m0s** (цикл замедлился до ~5.4 c/rep; буквы долетали — surrounding push «вd», «вdв»; zero deadline-abort WARN): средовый класс, применено документированное средство `ibus restart`, один санкционированный перепрогон — зелёный; счёт 24/24 по 4 зелёным прогонам сохранён.
- **Матрица: по одному транзиенту SetGlobalEngine «Operation was cancelled» на прогон** (разные кейсы, одна GDBus-сигнатура, 05:54:36 / 06:12:07): класс «ibus-activation transient» 06-08; оба кейса зелёные в соседнем прогоне двойки; отказ — на активации движка СТЕНДОМ до поведенческих ступеней.
- **requirements mark-complete — no-op:** строки фронматтера плана — свободные формулировки, не REQ-ID; таблица трассировки REQUIREMENTS.md обновляется на verify-work (прецеденты 06-08/07-01).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- **Фаза 7 закрыта кодово и документально (8/8 планов):** критерии 1-3 доказаны планами 07-01..07-06/07-08, критерий 4 — доказан в локальной форме двойки (этот план), критерий 5 — README/CONFIG точны, релиз = тег владельца post-merge.
- **К verify-work фазы 7 маршрутизируется (всё собрано в RELEASE-READINESS):** WINDOWS #12 (вердикт владельца по word-mixed/phrase-mixed), формальный свежесессионный D-48-гейт (fresh_session=true, docs/ci-runner.md), UAT звука (слышимость тонами), UAT-скриншот меню (Pitfall 1 — метки/мнемоники), A1 pixel-check целевого редактора «Настроек…».
- **Пост-merge владельца:** мерж PR → `git tag v1.1.0` → push → release.yml/goreleaser → сверка ассетов (поезд 04-07); план ни один шаг не исполнял, тегов на ветке нет.

## Self-Check: PASSED

- Created files exist on disk: 07-RELEASE-READINESS.md (21279 B), e2e-report-07-07-run1.txt (1502 B), e2e-report-07-07-run2.txt (1444 B) — FOUND; README.md modified (17053 B) — FOUND.
- All 3 plan commits exist: a264a54, d21d6d6, 7fbe4ec (FOUND via git log); measured from ledger 718d9d0..HEAD = 3 commits (matches frontmatter `commits: 3`).
- Acceptance criteria re-run on the final tree: Task 1 — README-OK / KEYS-MATCH / SOUND-DOC-OK / NO-TAG PASS, white-list остаток только в миграционной фразе; Task 2 — mise run ci exit 0 (повторно), flip-keystroke 24/24, menu-v2 + blocklist-silent PASS, матрица ×2 исполнена с сохранёнными отчётами; Task 3 — READINESS-OK / REGISTRY-COVERED / STATUS-OK / NOT-EXECUTED-OK PASS.
- План-level verification: греп-гейты README/CONFIG/RELEASE-READINESS зелёные; `git tag --list 'v1.1*'` пуст; состояние стола восстановлено (unit active, engine reactivated, goswitch-en).

---
*Phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii*
*Completed: 2026-10-05*
