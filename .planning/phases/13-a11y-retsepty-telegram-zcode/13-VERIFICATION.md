---
phase: 13-a11y-retsepty-telegram-zcode
verified: 2026-11-06T15:30:52Z
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
covered_digest: "v4:sha256:5f2d9d4b1739faadfbf14d4ce1d997bc6da6c38d715b384e0bad3ef24a0e941b"
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
---

# Phase 13: A11Y — рецепты Telegram/ZCode Verification Report

**Phase Goal:** Автокоррекция достижима в ZCode и Telegram через включение доступности в самих приложениях; демон не меняется (решение владельца 2026-10-08) — рецепты, доки, self-check
**Verified:** 2026-11-06T15:30:52Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

Фаза целиком опирается на живые ВМ-пробы (директива 6: повторный живой прогон с хоста верификатору запрещён). Доказательная база — 21 датированный ряд секции «Фаза 13» журнала ADR-008 (включая восстановленный 2026-11-06 ряд snap r2), пин секции в `TestSpikeJournalSectionsExist` (прогнан зелёным при верификации) и решения владельца, принятые в живых чекпойнтах: вердикт «Задокументировать» по A11Y-02 (2026-11-03, записан дословно в ряд «a11y-проба развязка») и приёмка A11Y-01 по живому дереву на ВМ (делегирование D-05). Премиса цели («автокоррекция достижима через включение доступности в приложениях») опровергнута машиной для ОБЕИХ целей — ZCode (мост не эмитит identity-события) и Telegram (стаб роли 75 даже unconfined); владелец закрыл оба требования задокументированными границами, рецепты/доки/self-check доставлены, демон не тронут (0 строк в internal/ и cmd/ от базы фазы 8aa3882 до HEAD). Две буквальные клаузы, опровергнутые пробой, проведены как overrides с полным обоснованием (см. frontmatter).

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | (SC1a) ZCode после user-local .desktop override с --force-renderer-accessibility и полного рестарта присутствует в живом AT-SPI-дереве — приёмка по живому дереву (exit 0 + номер роли текст-класса) | ✓ VERIFIED | Журнал ADR-008 ряды «a11y-проба zcode r1/r2 red\|green» (2026-11-03): RED ×2 (plain-стаб 2 ноды, exit 1) → GREEN ×2 (43+ нод, exit 0, роль 79 [entry], сфокусированная нода 79) — пара RED→GREEN ×2, два зелёных подряд (SPEC §7.2); приёмка делегирована ВМ (D-05) и закрыта владельцем |
| 2 | (SC1b) Автокоррекция работает в поле ввода ZCode | ✓ PASSED (override) | Опровергнута пробой ×2 с позитив-контролем (zenity fired +1 той же сессией); мост не эмитит identity-события → демон fail-closed skip role-unknown. Граница задокументирована: README.md:263-266, README.en.md:257-261, docs/CONFIG.md:449-456 (residual-risk PRIMARY). Override — владелец принял границу; демон не менялся |
| 3 | (SC2) Решение владельца по лестнице Telegram принято и исполнено | ✓ VERIFIED | Развилка исполнена дословно: все рунги красные по слоям 2-3 → checkpoint:decision gate=blocking-human ветка (c), ветка «правка демона» среди вариантов отсутствует; ВЕРДИКТ ВЛАДЕЛЬЦА «Задокументировать» (вариант 2) записан дословно в ряд «a11y-проба развязка 2026-11-03»; исполнено: граница в docs/CONFIG.md:458-491 + README×2, A11Y-02 закрыт в REQUIREMENTS.md; проба 3 форм сборки ×2 (snap 7.3.3 rev 7175 побайтовая реплика, flatpak GNOME 51, официальный tarball unconfined) — D-04 опровергнут машиной (стаб идентичен без confinement) |
| 4 | (SC3) Рецепты включают побочные эффекты и границы: плашка диктора, «флаги Electron применяются только на запуске», restart + self-check; README несёт residual-risk PRIMARY | ✓ VERIFIED | docs/CONFIG.md:485-491 (плашка ровно как наблюдена: flatpak/tarball живьём 2026-11-03, snap — не наблюдена), :409-413 («флаги Electron применяются только при запуске» + трей), шаг 2 каждого рецепта — полный выход/рестарт; PRIMARY-residual: README.md:276-279, README.en.md:273-277; grep-гейты зелёные |
| 5 | (SC4) К каждому рецепту приложен user-runnable self-check одной командой | ✓ VERIFIED | Полный вид `/usr/bin/python3 scripts/a11y-selfcheck.py <app>` в каждом рецепте: docs/CONFIG.md:417 (zcode), :482 (telegram); README.md:267,272,367; README.en.md:262,268,362; коды 0/1/2 расшифрованы во всех трёх файлах (включая ветки ошибки вызова/сломанного окружения — IN-01); коллизия с «goswitchctl selfcheck» отсутствует (счётчики упоминаний 5/5/0 идентичны до/после фазы, все — контекст валидации конфига) |
| 6 | (Plan 13-02 t2) Дисциплина рестарта доказуема: pid ДО ≠ pid ПОСЛЕ + флаг в cmdline нового процесса | ✓ PASSED (override) | pid-часть исполнена ×2 (34606→35923, 40414→41055, множества не пересекаются); cmdline-клауза структурно невыполнима на Electron 44 (ground-truth: явный флаг → grep всё равно 1); замена: экранный дифференциал дерева 2→43+ нод + прямой запуск с флагом → exit 0 (ряды r2 red/r2 green); отклонение в WINDOWS.md row 28 (open) |
| 7 | (D-06) Скрипт-зонд: роли по НОМЕРАМ против канона демона, контракт 0/1/2, живость /proc без subprocess | ✓ VERIFIED | scripts/a11y-selfcheck.py: TEXT_ROLES = {61,79,94} и ROLE_PASSWORD_TEXT = 40 побайтово совпадают с internal/appid/appid.go:88-96 (40/60/61/79/94); имена ролей — только печать; exit-матрица доказана живьём (ряд «a11y-selfcheck контракт»: rc0=0/rc1=1/rc2=2/rcu=2); живость — чтение /proc (process_pids), ноль subprocess/os.system (grep подтверждает — только docstring-упоминания); фиксы ревью WR-01/02/03 в исходнике (gi-import guard exit 2; applications() guard exit 2; поштучный try/except детей); py_compile OK, бит x установлен |
| 8 | Граница фазы: демон не меняется (решение владельца 2026-10-08) | ✓ VERIFIED | `git diff --stat 8aa3882..HEAD -- internal/ cmd/` — пусто (0 строк); в рабочем дереве `git diff --exit-code -- internal/ cmd/` чисто; подтверждено и 13-REVIEW.md («zero changes under internal/ and cmd/»), и disposition-гейтом |

**Score:** 8/8 truths verified (0 present, behavior-unverified; 2 via owner-accepted overrides)

### Overrides Applied (2)

| # | Must-have | Why acceptable | Accepted |
|---|-----------|----------------|----------|
| 1 | Автокоррекция работает в поле ZCode (SC1b) | Опровергнуто пробой ×2 с живым позитив-контролем; residual-risk задокументирован в 3 файлах; демон нетронут — граница фазы соблюдена; приёмка A11Y-01 закрыта владельцем по живому дереву | owner, live checkpoint, 2026-11-06 |
| 2 | cmdline- клауза рестарта (plan 13-02 t2) | Электроны Electron 44 переписывают argv — буквальная форма невыполнима даже с флагом; замена (pid-дифференциал + дифференциал дерева + ground-truth запуск) доказывает то же строгоче; WINDOWS.md row 28 фиксирует отклонение | owner (в составе приёмки A11Y-01), 2026-11-06; row 28 pending formal waive |

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `scripts/a11y-selfcheck.py` | Внешний AT-SPI-зонд D-06: роли по номерам, exit 0/1/2, /proc-живость | ✓ VERIFIED | 244 строки, shebang /usr/bin/python3, бит x, py_compile OK; все элементы контракта в исходнике; используется всеми доками (5+ вызовов полного вида) и журналом |
| `docs/adr/adr-008-spike-log.md` (секция Фазы 13) | Датированные ряды пробы: подъём, калибровка, рунги ×2, лестница, итог, развязка, контракт, парковка | ✓ VERIFIED | Строки 721-764: все ряды планов 13-01/13-02 присутствуют, включая восстановленный 2026-11-06 ряд snap r2 с честными квантификаторами (WR-04); 21 строка с тегами a11y-* |
| `test/e2e/spike_journal_test.go` | Пин secA11yProbe в spikeSections() | ✓ VERIFIED | Константа (строка 45) побайтово = заголовок журнала; тест прогнан зелёным при верификации |
| `docs/CONFIG.md` | Полный ZCode-рецепт + Telegram-граница + побочки + self-check | ✓ VERIFIED | Раздел «Ручной рецепт» :373-512: валидация 2026-11-03, роль 79, роль 40 API-ключа, chmod 4755 chrome-sandbox, self-rewrite override, argv-перезапись, плашка диктора, предупреждения :394-405 сохранены + flatpak-пункт |
| `README.md` | Краткие рецепты + PRIMARY + «что магия НЕ делает» | ✓ VERIFIED | :257-279 (рецепты + residual-risk), :361-370 («Что магия НЕ делает» + self-check) |
| `README.en.md` | Байт-параллельное зеркало | ✓ VERIFIED | :251-277, :355-365 — структурно параллельные блоки, те же утверждения и one-liner-ы |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| a11y-selfcheck.py (вердикт по номеру роли) | internal/appid/appid.go:88-96 (канон 40/60/61/79/94) | Один канон номеров — скрипт печатает номера, которыми гейтит демон | ✓ WIRED | Сравнение исходников: TEXT_ROLES {61,79,94} + 40 + 60 совпадают побайтово; имена только в describe_role для человека |
| Тексты рецептов (README×2, CONFIG.md) | Датированные ряды секции Фазы 13 (ADR-008) | probe-sourced-docs: каждое утверждение имеет ряд-первоисточник | ✓ WIRED | Перекрёстная сверка: роль 79 / стаб роль 75 / плашка flatpak+tarball не snap / даты 2026-11-03 / «2 ноды → 43+ нод» — доки и журнал согласованы; в рецепте Telegram команд установки НЕТ (ветка = вердикту итог-ряда) |
| Строка запуска в рецептах | scripts/a11y-selfcheck.py (контракт 0/1/2) | selfcheck-one-liner полного вида | ✓ WIRED | Полный путь `/usr/bin/python3 scripts/a11y-selfcheck.py <app>` во всех 3 файлах; ни «goswitchctl selfcheck», ни голого a11y-selfcheck в a11y-контексте |
| secA11yProbe (test) | Заголовок секции журнала | Пин TestSpikeJournalSectionsExist | ✓ WIRED | Строка 45 = журнальная строка 721; `go test ./test/e2e/ -run TestSpikeJournalSectionsExist` → ok |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| scripts/a11y-selfcheck.py | role_counts / focused_roles | Живой обход AT-SPI-дерева (walk → node.get_role()) | Да — вердикт вычисляется из пробы дерева, ноль захардкоженных вердиктов | ✓ FLOWING |
| scripts/a11y-selfcheck.py | pids | Чтение /proc/<pid>/cmdline | Да | ✓ FLOWING |
| docs (рецепты) | Факты (роль 79, стаб 75, плашка, даты) | Датированные ряды журнала ADR-008 | Да — каждый факт сверен с рядом | ✓ FLOWING |
| Журнал (ряды слоёв 3) | Дельты счётчиков | goswitchctl status до/после сентинела на ВМ | Да — машиночитаемые дельты в рядах (fired/abstained/ac_skip_role_unknown/ac_role_errors) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Пин секции журнала жив | `go test ./test/e2e/ -run 'TestSpikeJournalSectionsExist' -count=1` | ok (0.008s) | ✓ PASS |
| Скрипт компилируется системным python3 | `/usr/bin/python3 -m py_compile scripts/a11y-selfcheck.py` | exit 0 | ✓ PASS |
| Скрипт исполним, shebang дословный | `test -x` + `head -1` | `#!/usr/bin/python3` | ✓ PASS |
| Гейты A11Y-03 (позитив) | grep: selfcheck полного вида ×3 файла, force-renderer-accessibility ×2 README, PRIMARY ×2, диктор/трей/«применяются только при запуске»/«a11y-проба zcode» в CONFIG.md | Все найдены | ✓ PASS |
| Гейты A11Y-03 (негатив) | grep: gail:atk-bridge (0 ×3), snap connect (0 ×3), счётчики «goswitchctl selfcheck» pre/post фазы идентичны (5/5/0 — предсуществующие, контекст валидации конфига) | Чисто | ✓ PASS |
| Граница демона | `git diff --stat 8aa3882..HEAD -- internal/ cmd/` | Пусто (0 изменений) | ✓ PASS |
| Живой прогон зонда на хосте | — | НЕ исполнялся осознанно: директива 6 запрещает a11y-зонд на хосте в любой форме | ✓ SKIP (по запрету) |

### Probe Execution

| Probe | Command | Result | Status |
|-------|---------|--------|--------|
| scripts/*/tests/probe-*.sh | find | Не существуют — для этой фазы не заявлены | N/A |
| Живые ВМ-пробы (snap/flatpak/tarball/zcode, 21 ряд) | Повтор с хоста запрещён директивой 6; верификация read-only | Не перезапускались — доказательная база: датированные ряды журнала + пин-тест + закрытые чекпойнты владельца | N/A (evidence = journal rows + owner checkpoints) |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| A11Y-01 | 13-02 | ZCode: автокоррекция; приёмка по живому AT-SPI-дереву | ✓ SATISFIED | Живое дерево: exit 0 + роль 79 ×2, RED→GREEN ×2; слой 3 — опровергнутая клауза проведена owner-override (residual-risk документирован); REQUIREMENTS.md: [x] Complete |
| A11Y-02 | 13-01 | Telegram: решение владельца (unconfined vs демон-fallback) | ✓ SATISFIED | Вердикт «Задокументировать» (2026-11-03, дословно в ряду развязки) — третья, владельческая развязка всех-красных рунгов; граница задокументирована в CONFIG/README; REQUIREMENTS.md: [x] Complete |
| A11Y-03 | 13-03 | Рецепты в README/CONFIG.md с побочками и границами | ✓ SATISFIED | Все гейты зелёные (см. Spot-Checks); рецепты probe-sourced; REQUIREMENTS.md: [x] Complete |

Orphaned requirements: none — REQUIREMENTS.md привязывает к Фазе 13 ровно A11Y-01/02/03, все три заявлены в планах (13-02/13-01/13-03 соответственно).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER: ноль во всех 6 файлах фазы | — | — |
| — | — | Код-ревью 2026-11-06: 0 critical / 4 warning / 2 info — все шесть fixed и dispositioned (96a3705, bc1b29e, a615ddc, 8fbaf62, 4f37110, 2677fa5); фиксы подтверждены в исходнике | ℹ️ Info | Закрыто до верификации |

### Human Verification Required

None — живые чекпойнты владельца уже закрыли человеческие пункты фазы (вердикт A11Y-02 «Задокументировать» 2026-11-03; приёмка A11Y-01 по живому дереву на ВМ, делегирование D-05). Перенесены закрытыми; не переоткрыты. Новых пунктов, требующих человека, верификация не выявила (артефакты на ВМ сохранены для UAT-демонстрации — ряд «a11y-стенд итог»).

### Gaps Summary

Gaps отсутствуют. Примечания для прозрачности:

1. **WINDOWS.md row 28** (deviation, phase 13, статус open): cmdline-доказательство заменено на pid-дифференциал + дифференциал дерева — проведено как override #2; формальный waive строки остаётся владельцу.
2. **REQUIREMENTS.md A11Y-02** буквальный текст («автокоррекция достижима») расходится с фактическим исходом (недостижима ни в одной форме, закрыта вердиктом владельца) — расходование зафиксировано вердиктом в журнале и override-логикой; текст требования в REQUIREMENTS.md не переписывался (вне полномочий верификатора).
3. **QR-честность Telegram**: слой 3 измерен на поле авторизации; поле сообщения недостижимо без QR-логина владельца (DC с ВМ недоступны) — честно записано в итог-ряду и чекпойнте.

---

_Verified: 2026-11-06T15:30:52Z_
_Verifier: Claude (gsd-verifier)_
