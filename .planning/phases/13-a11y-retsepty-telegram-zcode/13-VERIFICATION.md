---
phase: 13-a11y-retsepty-telegram-zcode
verified: 2026-11-06T16:14:52Z
status: passed
score: 8/8 must-haves verified
covered_files:
  - .planning/REQUIREMENTS.md
  - .planning/phases/13-a11y-retsepty-telegram-zcode/13-01-PLAN.md
  - .planning/phases/13-a11y-retsepty-telegram-zcode/13-01-SUMMARY.md
  - .planning/phases/13-a11y-retsepty-telegram-zcode/13-02-PLAN.md
  - .planning/phases/13-a11y-retsepty-telegram-zcode/13-02-SUMMARY.md
  - .planning/phases/13-a11y-retsepty-telegram-zcode/13-03-PLAN.md
  - .planning/phases/13-a11y-retsepty-telegram-zcode/13-03-SUMMARY.md
  - .planning/phases/13-a11y-retsepty-telegram-zcode/13-REVIEW-DISPOSITION.md
  - .planning/phases/13-a11y-retsepty-telegram-zcode/13-REVIEW.md
  - README.en.md
  - README.md
  - docs/CONFIG.md
  - docs/adr/adr-008-spike-log.md
  - scripts/a11y-selfcheck.py
  - test/e2e/spike_journal_test.go
covered_digest: "v4:sha256:f7ad816329ff336cd1feba1695b8bb47c87164b3f2870a37e5b2f0a9d25b4638"
behavior_unverified: 0
overrides_applied: 2
overrides:
  - must_have: "Автокоррекция работает в поле ввода ZCode (ROADMAP SC1, вторая клауза; REQUIREMENTS A11Y-01 «автокоррекция работает»)"
    reason: "Живая проба ×2 опровергла достижимость на этой сборке: мост Chromium отвечает внешнему обходу (зонд видит поле, роль 79, exit 0), но не эмитит identity/state-события (dbus-monitor 0 state-changed; 0 zcode-identity в журнале демона) → демон честно уходит в skip role-unknown (fail-closed). Позитив-контроль zenity той же сессии fired +1 — красный специфичен мосту ZCode, оракул жив. Граница задокументирована как residual-risk в README.md/README.en.md/CONFIG.md; демон не менялся (граница фазы 2026-10-08). Приёмка A11Y-01 выполнена по живому дереву на ВМ (делегирование D-05) и закрыта владельцем в живом чекпойнте."
    accepted_by: "owner (Daniel Podolsky) — live checkpoint, закрыто при сдаче фазы (per phase outcome context)"
    accepted_at: "2026-11-06T00:00:00Z"
  - must_have: "Дисциплина рестарта доказуема: pid ДО ≠ pid ПОСЛЕ и /proc/<pid>/cmdline нового процесса содержит --force-renderer-accessibility (plan 13-02, truth 2)"
    reason: "Буквальная cmdline-клауза структурно невыполнима на Electron 44: процесс перезаписывает argv на «ZCode» ≤1 c после exec — ground-truth запуск с ЯВНЫМ флагом всё равно даёт grep=1 (ряд «a11y-механика zcode install», 2026-11-03). Замена доказательства достигает того же намерения строгоче: непересекающиеся pid-множества ДО/ПОСЛЕ ×2 + экранный дифференциал дерева (тот же welcome-экран: 2 ноды plain против 43+ нод с флагом) + прямой ground-truth запуск с флагом → exit 0. Отклонение честно записано в WINDOWS.md (row 28, статус open — формальный waive за владельцем)."
    accepted_by: "owner (Daniel Podolsky) — в составе закрытой приёмки A11Y-01; WINDOWS.md row 28 pending formal waive"
    accepted_at: "2026-11-06T00:00:00Z"
re_verification:
  previous_status: passed
  previous_score: 8/8
  stale_refresh: true
  trigger: "merge ded4c1f (ветка Фазы 12 в ветку Фазы 13) после исходной верификации 2026-11-06T15:30:52Z"
  gaps_closed: []
  gaps_remaining: []
  regressions: []
advisory:
  - finding: "WINDOWS.md row 28 (deviation, phase 13, статус open): cmdline-доказательство рестарта заменено на pid-дифференциал + дифференциал дерева — override #2 принят владельцем по существу, формальный waive строки остаётся незакрытым"
    category: other
    reason: "Владелец принял отклонение в живом чекпойнте при сдаче фазы; формальное закрытие строки реестра WINDOWS.md (смена статуса open → waived) — за владельцем; пере-верификация не блокируется"
    evidence_status: "owner-accepted in substance; formal waive pending"
---

# Phase 13: A11Y — рецепты Telegram/ZCode Verification Report (STALE-REFRESH)

**Phase Goal:** Автокоррекция достижима в ZCode и Telegram через включение доступности в самих приложениях; демон не меняется (решение владельца 2026-10-08) — рецепты, доки, self-check
**Verified:** 2026-11-06T16:14:52Z
**Status:** passed
**Re-verification:** Yes — STALE-REFRESH после merge ded4c1f (предыдущий отчёт 2026-11-06T15:30:52Z, passed 8/8, закрывался до слияния)

## Goal Achievement

Повторная верификация на слитом дереве (HEAD = ded4c1f). Влияние слияния проверено пофайлово и не задело субстанцию фазы 13:

- **Журнал**: секция «Фаза 13 A11Y-проба сборок» сдвинута 721 → 757 ровно на 36 строк вставленной ДО неё секции Фазы 12 «сел-проба канала чтения»; все 21 строка данных фазы 13 байт-идентичны (включая восстановленный 2026-11-06 ряд snap r2 с честными квантификаторами WR-04).
- **Пин-тест**: `secA11yProbe` сохранён (строка 49 = заголовок журнала строка 757, побайтово); слияние добавило только `secLineProbe` Фазы 12 рядом; `go test ./test/e2e/ -run 'TestSpikeJournalSectionsExist' -count=1` → ok (0.003s).
- **REQUIREMENTS.md**: строки A11Y-01/02/03 не тронуты (все `[x] … Phase 13 Complete`); слияние закрыло только TERM-03.
- **Граница демона**: `git diff --stat 8aa3882..HEAD -- internal/ cmd/` — пусто и на слитом дереве; рабочее дерево чисто.
- **.gitignore** +2 паттерна (`.planning/tmp/`, `.planning/phases/*/evidence-*.log`) — ни один покрытый путь из покрытого набора не удалён из дерева; все 15 файлов covered-набора отслеживаются в HEAD.
- **Доки и скрипт**: `git diff 5544d7e..HEAD -- scripts/a11y-selfcheck.py docs/CONFIG.md README.md README.en.md` — пусто (байт-идентичны до/после слияния); 13-REVIEW.md/13-REVIEW-DISPOSITION.md впервые закоммичены самим merge-коммитом и присутствуют в HEAD.

Доказательная база фазы — живые ВМ-пробы (директива 6: повторный живой прогон с хоста верификатору запрещён): 21 датированный ряд секции «Фаза 13» журнала ADR-008, пин-тест (прогнан зелёным на слитом дереве) и решения владельца из живых чекпойнтов: вердикт «Задокументировать» по A11Y-02 (2026-11-03, дословно в ряду «a11y-проба развязка») и приёмка A11Y-01 по живому дереву на ВМ (делегирование D-05). Оба owner-closed пункта перенесены закрытыми. Премиса цели («автокоррекция достижима через включение доступности в приложениях») опровергнута машиной для обеих целей — ZCode (мост не эмитит identity-события) и Telegram (стаб роли 75 даже unconfined); владелец закрыл оба требования задокументированными границами, рецепты/доки/self-check доставлены, демон не тронут.

### Observable Truths

| # | Truth | Status | Evidence (слитое дерево, HEAD ded4c1f) |
|---|-------|--------|----------|
| 1 | (SC1a) ZCode после user-local .desktop override с --force-renderer-accessibility и полного рестарта присутствует в живом AT-SPI-дереве — приёмка по живому дереву (exit 0 + номер роли текст-класса) | ✓ VERIFIED | Журнал ADR-008 строки 795-798 (ряды «a11y-проба zcode r1/r2 red\|green», 2026-11-03): RED ×2 (стаб 2 ноды, exit 1) → GREEN ×2 (43+ нод, exit 0, роль 79 [entry], сфокусированная нода 79) — пара RED→GREEN ×2, два зелёных подряд (SPEC §7.2); приёмка делегирована ВМ (D-05) и закрыта владельцем — переносится закрытой |
| 2 | (SC1b) Автокоррекция работает в поле ввода ZCode | ✓ PASSED (override) | Опровергнута пробой ×2 с позитив-контролем (zenity fired +1 той же сессией); мост не эмитит identity-события → демон fail-closed skip role-unknown. Граница задокументирована: README.md:277-279, README.en.md:273-277, docs/CONFIG.md:449-456 (residual-risk PRIMARY; перенос на слитом дереве подтверждён grep'ом). Override — владелец принял границу; демон не менялся |
| 3 | (SC2) Решение владельца по лестнице Telegram принято и исполнено | ✓ VERIFIED | Вердикт «Задокументировать» записан дословно в ряду «a11y-проба развязка 2026-11-03» (журнал, строка 793); развилка исполнена дословно: все рунги красные по слоям 2-3 → ветка (c) checkpoint:decision gate=blocking-human, ветка «правка демона» среди вариантов отсутствует; исполнено: граница в docs/CONFIG.md:458-491 + README×2, A11Y-02 `[x] Complete` в REQUIREMENTS.md (строки не тронуты слиянием — проверено diff'ом); проба 3 форм сборки ×2 (snap 7.3.3 rev 7175 побайтовая реплика, flatpak GNOME 51, официальный tarball unconfined) — D-04 опровергнут машиной (стаб идентичен без confinement) |
| 4 | (SC3) Рецепты включают побочные эффекты и границы: плашка диктора, «флаги Electron применяются только на запуске», restart + self-check; README несёт residual-risk PRIMARY | ✓ VERIFIED | docs/CONFIG.md:485-487 (плашка ровно как наблюдена: flatpak/tarball живьём 2026-11-03, snap — не наблюдена), :409-410 («флаги Electron применяются только при запуске» + трей), шаг 2 каждого рецепта — полный выход/рестарт; PRIMARY-residual: README.md:277-279, README.en.md:273-277; grep-гейты зелёные на слитом дереве |
| 5 | (SC4) К каждому рецепту приложен user-runnable self-check одной командой | ✓ VERIFIED | Полный вид `/usr/bin/python3 scripts/a11y-selfcheck.py <app>` в каждом рецепте: CONFIG.md ×2 (строки ~417, ~482), README.md ×3, README.en.md ×3 (grep-счётчики 2/3/3); коды 0/1/2 расшифрованы во всех трёх файлах, включая ветки IN-01 (ошибка вызова / сломанное окружение — CONFIG.md:419-424 проверено); упоминания «goswitchctl selfcheck» (5/5 в README RU/EN, 0 в CONFIG.md) — все в контексте валидации конфигурации (строки 116/134/384/454/473), не a11y-зонд |
| 6 | (Plan 13-02 t2) Дисциплина рестарта доказуема: pid ДО ≠ pid ПОСЛЕ + флаг в cmdline нового процесса | ✓ PASSED (override) | pid-часть исполнена ×2 (34606→35923, 40414→41055, множества не пересекаются); cmdline-клауза структурно невыполнима на Electron 44 (ground-truth: явный флаг → grep всё равно 1, ряд install строка 794); замена: экранный дифференциал дерева 2→43+ нод + прямой запуск с флагом → exit 0 (ряды r2 red/r2 green, строки 797-798); отклонение в WINDOWS.md row 28 (open) — advisory, не блокирует |
| 7 | (D-06) Скрипт-зонд: роли по НОМЕРАМ против канона демона, контракт 0/1/2, живость /proc без subprocess | ✓ VERIFIED | scripts/a11y-selfcheck.py (243 строки, байт-идентичен до/после слияния): TEXT_ROLES = {61,79,94} (строка 51) и ROLE_PASSWORD_TEXT = 40 (строка 52) побайтово совпадают с internal/appid/appid.go:88-96 (40/60/61/79/94); имена ролей — только печать; exit-матрица в исходнике (sys.exit(2) guard'ы строк 46/63; return 2/2/1/0 строк 229/234/238/239); живость — чтение /proc (process_pids), ноль subprocess/os.system (только docstring-упоминания); фиксы ревью WR-01/02/03 + IN-02 в исходнике (gi-guard exit 2; applications() guard exit 2; поштучный try/except детей walk(); f-префикс снят); py_compile OK, бит x установлен, shebang `#!/usr/bin/python3` |
| 8 | Граница фазы: демон не меняется (решение владельца 2026-10-08) | ✓ VERIFIED | На СЛИТОМ дереве: `git diff --stat 8aa3882..HEAD -- internal/ cmd/` — пусто (0 строк; слияние Фазы 12 демона тоже не тронуло); рабочее дерево `git diff --exit-code -- internal/ cmd/` чисто; подтверждено 13-REVIEW.md и disposition-гейтом |

**Score:** 8/8 truths verified (0 present, behavior-unverified; 2 via owner-accepted overrides, перенесены закрытыми)

### Overrides Applied (2 — перенесены из исходной верификации без изменений)

| # | Must-have | Why acceptable | Accepted |
|---|-----------|----------------|----------|
| 1 | Автокоррекция работает в поле ZCode (SC1b) | Опровергнуто пробой ×2 с живым позитив-контролем; residual-risk задокументирован в 3 файлах (перепроверен на слитом дереве); демон нетронут; приёмка A11Y-01 закрыта владельцем по живому дереву | owner, live checkpoint, 2026-11-06 |
| 2 | cmdline-клауза рестарта (plan 13-02 t2) | Electron 44 переписывает argv — буквальная форма невыполнима даже с флагом; замена (pid-дифференциал + дифференциал дерева + ground-truth запуск) доказывает то же строгоче; WINDOWS.md row 28 фиксирует отклонение (open, formal waive pending) | owner (в составе приёмки A11Y-01), 2026-11-06 |

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `scripts/a11y-selfcheck.py` | Внешний AT-SPI-зонд D-06: роли по номерам, exit 0/1/2, /proc-живость | ✓ VERIFIED | 243 строки, байт-идентичен до/после слияния; shebang /usr/bin/python3, бит x, py_compile OK; все элементы контракта в исходнике; используется всеми доками (8 вызовов полного вида) и журналом |
| `docs/adr/adr-008-spike-log.md` (секция Фазы 13) | Датированные ряды пробы: подъём, калибровка, рунги ×2, лестница, итог, развязка, контракт, парковка | ✓ VERIFIED | Секция на строках 757-800 (сдвиг +36 от вставки секции Фазы 12 ДО неё, содержимое не тронуто): все ряды планов 13-01/13-02 присутствуют, включая восстановленный 2026-11-06 ряд snap r2; 21 строка данных |
| `test/e2e/spike_journal_test.go` | Пин secA11yProbe в spikeSections() | ✓ VERIFIED | Константа (строка 49) побайтово = заголовок журнала (строка 757); слияние добавило только secLineProbe Фазы 12 — пин фазы 13 сохранён; тест прогнан зелёным на слитом дереве (ok, 0.003s) |
| `docs/CONFIG.md` | Полный ZCode-рецепт + Telegram-граница + побочки + self-check | ✓ VERIFIED | Байт-идентичен до/после слияния (git diff пуст): раздел «Ручной рецепт»: валидация 2026-11-03, роль 79, роль 40 API-ключа, chmod 4755 chrome-sandbox, self-rewrite override, argv-перезапись, плашка диктора, предупреждения :394-405 + flatpak-пункт; честные квантификаторы WR-04 (:461-463) |
| `README.md` | Краткие рецепты + PRIMARY + «что магия НЕ делает» | ✓ VERIFIED | Байт-идентичен до/после слияния: :276-279 (residual-risk PRIMARY), :361-370 («Что магия НЕ делает» + self-check) |
| `README.en.md` | Байт-параллельное зеркало | ✓ VERIFIED | Байт-идентичен до/после слияния: :273-277, :355-365 — структурно параллельные блоки, те же утверждения и one-liner-ы |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| a11y-selfcheck.py (вердикт по номеру роли) | internal/appid/appid.go:88-96 (канон 40/60/61/79/94) | Один канон номеров — скрипт печатает номера, которыми гейтит демон | ✓ WIRED | Сравнение исходников на слитом дереве: TEXT_ROLES {61,79,94} + 40 совпадают побайтово; имена только в describe_role для человека |
| Тексты рецептов (README×2, CONFIG.md) | Датированные ряды секции Фазы 13 (ADR-008) | probe-sourced-docs: каждое утверждение имеет ряд-первоисточник | ✓ WIRED | Перекрёстная сверка на слитом дереве: роль 79 / стаб роль 75 / плашка flatpak+tarball не snap / даты 2026-11-03 / «2 ноды → 43+ нод» — доки и журнал согласованы; в рецепте Telegram команд установки НЕТ (ветка = вердикту итог-ряда) |
| Строка запуска в рецептах | scripts/a11y-selfcheck.py (контракт 0/1/2) | selfcheck-one-liner полного вида | ✓ WIRED | Полный путь `/usr/bin/python3 scripts/a11y-selfcheck.py <app>` во всех 3 файлах (счётчики 2/3/3); ни «goswitchctl selfcheck», ни голого a11y-selfcheck в a11y-контексте |
| secA11yProbe (test) | Заголовок секции журнала | Пин TestSpikeJournalSectionsExist | ✓ WIRED | Строка 45 в spikeSections() → константа строка 49 = журнальная строка 757; `go test ./test/e2e/ -run TestSpikeJournalSectionsExist` → ok на слитом дереве |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| scripts/a11y-selfcheck.py | role_counts / focused_roles | Живой обход AT-SPI-дерева (walk → node.get_role()) | Да — вердикт вычисляется из пробы дерева, ноль захардкоженных вердиктов | ✓ FLOWING |
| scripts/a11y-selfcheck.py | pids | Чтение /proc/<pid>/cmdline | Да | ✓ FLOWING |
| docs (рецепты) | Факты (роль 79, стаб 75, плашка, даты) | Датированные ряды журнала ADR-008 | Да — каждый факт сверен с рядом на слитом дереве | ✓ FLOWING |
| Журнал (ряды слоёв 3) | Дельты счётчиков | goswitchctl status до/после сентинела на ВМ | Да — машиночитаемые дельты в рядах (fired/abstained/ac_skip_role_unknown/ac_role_errors) | ✓ FLOWING |

### Behavioral Spot-Checks (слитое дерево, HEAD ded4c1f)

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Пин секции журнала жив после слияния | `go test ./test/e2e/ -run 'TestSpikeJournalSectionsExist' -count=1` | ok (0.003s) — компиляция пакета с новыми файлами Фазы 12 + пины всех 15 секций зелёные | ✓ PASS |
| Скрипт компилируется системным python3 | `/usr/bin/python3 -m py_compile scripts/a11y-selfcheck.py` | exit 0 | ✓ PASS |
| Скрипт исполним, shebang дословный | `test -x` + `head -1` | `#!/usr/bin/python3` | ✓ PASS |
| Гейты A11Y-03 (позитив) | grep: selfcheck полного вида (2/3/3), force-renderer-accessibility ×2 в каждом из 3 файлов, PRIMARY в README×2, диктор/Screen Reader + «только при запуске»/трей в CONFIG.md | Все найдены | ✓ PASS |
| Гейты A11Y-03 (негатив) | grep: gail:atk-bridge (0 ×3 файла), snap connect (0 ×3); «goswitchctl selfcheck» — 5/5/0, все в контексте валидации конфигурации (README:116/134/384/454/473) | Чисто | ✓ PASS |
| Граница демона на слитом дереве | `git diff --stat 8aa3882..HEAD -- internal/ cmd/` + `git diff --exit-code -- internal/ cmd/` | Пусто / чисто (0 изменений; Фаза 12 демона тоже не тронула) | ✓ PASS |
| Слияние не задело субстанцию фазы 13 | `git diff 5544d7e..HEAD -- scripts/a11y-selfcheck.py docs/CONFIG.md README.md README.en.md` | Пусто (байт-идентично); журнал/тест изменены ТОЛЬКО вставкой контента Фазы 12 (сдвиг секции 721→757, +secLineProbe) | ✓ PASS |
| Живой прогон зонда на хосте | — | НЕ исполнялся осознанно: директива 6 запрещает a11y-зонд на хосте в любой форме | ✓ SKIP (по запрету) |

### Probe Execution

| Probe | Command | Result | Status |
|-------|---------|--------|--------|
| scripts/*/tests/probe-*.sh | find | Не существуют — для этой фазы не заявлены | N/A |
| Живые ВМ-пробы (snap/flatpak/tarball/zcode, 21 ряд) | Повтор с хоста запрещён директивой 6; STALE-REFRESH верификация read-only | Не перезапускались — доказательная база: датированные ряды журнала (проверены на слитом дереве) + пин-тест (прогнан зелёным) + закрытые чекпойнты владельца | N/A (evidence = journal rows + owner checkpoints) |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| A11Y-01 | 13-02 | ZCode: автокоррекция; приёмка по живому AT-SPI-дереву | ✓ SATISFIED | Живое дерево: exit 0 + роль 79 ×2, RED→GREEN ×2 (журнал:795-798); слой 3 — опровергнутая клауза проведена owner-override (residual-risk документирован); REQUIREMENTS.md: `[x] Complete` (строка не тронута слиянием) |
| A11Y-02 | 13-01 | Telegram: решение владельца (unconfined vs демон-fallback) | ✓ SATISFIED | Вердикт «Задокументировать» (2026-11-03, дословно в ряду развязки, журнал:793); граница задокументирована в CONFIG/README; REQUIREMENTS.md: `[x] Complete` (строка не тронута слиянием) |
| A11Y-03 | 13-03 | Рецепты в README/CONFIG.md с побочками и границами | ✓ SATISFIED | Все гейты зелёные на слитом дереве (см. Spot-Checks); рецепты probe-sourced; REQUIREMENTS.md: `[x] Complete` |

Orphaned requirements: none — REQUIREMENTS.md привязывает к Фазе 13 ровно A11Y-01/02/03, все три заявлены в планах (13-02/13-01/13-03 соответственно).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER: ноль во всех 6 файлах фазы (скан на слитом дереве) | — | — |
| — | — | Код-ревью 2026-11-06: 0 critical / 4 warning / 2 info — все шесть fixed и dispositioned (96a3705, bc1b29e, a615ddc, 8fbaf62, 4f37110; коммиты в истории, фиксы переподтверждены в исходнике на слитом дереве) | ℹ️ Info | Закрыто до исходной верификации; слияние не откатало |

### Advisory (New Scope, Unevidenced)

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| 1 | WINDOWS.md row 28 (deviation, phase 13, open): cmdline-доказательство заменено на pid-дифференциал + дифференциал дерева | other | Отклонение принято владельцем по существу (override #2); формальный waive строки реестра — незакрытый владелец-экшн, не верификаторский; не блокирует |

### Human Verification Required

None — живые чекпойнты владельца уже закрыли человеческие пункты фазы (вердикт A11Y-02 «Задокументировать» 2026-11-03; приёмка A11Y-01 по живому дереву на ВМ, делегирование D-05). Перенесены закрытыми; не переоткрыты. STALE-REFRESH не выявил новых пунктов, требующих человека (артефакты на ВМ сохранены для UAT-демонстрации — ряд «a11y-стенд итог»).

### Gaps Summary

Gaps отсутствуют. STALE-REFRESH на слитом дереве подтверждает исходный вердикт без изменений: слияние ded4c1f затронуло только tracking-файлы (ROADMAP/STATE/REQUIREMENTS/state.json — A11Y-строки не тронуты), .gitignore (+2 паттерна, ни один покрытый файл не удалён) и впервые закоммитило файлы ревью фазы 13; субстанция фазы (скрипт, доки, секция журнала, пин-тест, фиксы ревью) байт-идентична или эквивалентна (журнал/тест — с точностью до вставки контента Фазы 12 рядом). Примечания для прозрачности:

1. **WINDOWS.md row 28** (deviation, phase 13, статус open): cmdline-доказательство заменено на pid-дифференциал + дифференциал дерева — проведено как override #2; формальный waive строки остаётся владельцу (advisory).
2. **REQUIREMENTS.md A11Y-02** буквальный текст («автокоррекция достижима») расходится с фактическим исходом (недостижима ни в одной форме, закрыта вердиктом владельца) — расходование зафиксировано вердиктом в журнале и override-логикой; текст требования не переписывался (вне полномочий верификатора).
3. **QR-честность Telegram**: слой 3 измерен на поле авторизации; поле сообщения недостижимо без QR-логина владельца (DC с ВМ недоступны) — честно записано в итог-ряду и чекпойнте.

---

_Verified: 2026-11-06T16:14:52Z_
_Verifier: Claude (gsd-verifier) — STALE-REFRESH after merge ded4c1f_
