---
phase: 12-term-stroka
verified: 2026-11-06T16:15:12Z
status: passed
score: 9/9 must-haves verified
covered_files:
  - .planning/REQUIREMENTS.md
  - .planning/ROADMAP.md
  - .planning/STATE.md
  - .planning/phases/12-term-stroka/12-01-PLAN.md
  - .planning/phases/12-term-stroka/12-01-SUMMARY.md
  - .planning/phases/12-term-stroka/12-02-PLAN.md
  - .planning/phases/12-term-stroka/12-03-PLAN.md
  - .planning/phases/12-term-stroka/12-04-PLAN.md
  - .planning/phases/12-term-stroka/12-05-PLAN.md
  - docs/adr/adr-008-spike-log.md
  - mise.toml
  - test/e2e/case_term_line_probe.go
  - test/e2e/fixtures/gs-e2e-wezterm-xim-line.lua
  - test/e2e/main.go
  - test/e2e/spike_journal_test.go
covered_digest: "v4:sha256:b81b67f618cc3d26ffee01c08bf5d2c9f6559d058ca20208ad1ef68af6c3c5bf"
behavior_unverified: 0
overrides_applied: 1
overrides:
  - must_have: "SC1 (механизм): тройной правый Shift исправляет всю набранную строку в wezterm — курсор возвращается в исходную позицию, суффюкс справа от курсора цел (End-нормализация задокументирована)"
    reason: "Плановый механизм (машина строки чтением выделения) на живой сел-пробе 12-01 мёртв на корне — все три ячейки RED (физика-wayland / физика-xim / форвард, adr-008 ряды 23:16:57–23:17:04), сентинел-раундтрип доказал: красное — свойство рецепта CopyMode→PRIMARY, не инструментария. Владелец закрыл гейт вердиктом 2026-11-03 (дословно, adr-008 ряд «сел-проба закрытие»): «коррекция всей строки в wezterm у меня работает, а коррекция выделения мне не нужна» — цель SC1 (пользователь исправляет всю строку) достигнута существующей машиной слова Фазы 10 на длинных пробегах (журнал хоста green106 2026-10-31…11-03: 160 done / 0 skipped / 0 rejected, одиночные жесты до 47 символов, forwards:94 = 47 пар BackSpace+commit, машина round-4). Гейт-развилка плана исполнена дословно: checkpoint:decision gate=blocking-human, фолбэк «буфер вместо выделения» не предлагался и не исполнялся (D-01)"
    accepted_by: owner (Daniel Podolsky)
    accepted_at: 2026-11-03
re_verification:
  previous_status: passed
  previous_score: 9/9
  gaps_closed: []
  gaps_remaining: []
  regressions: []
advisory:
  - finding: "Ссылка оркестратора на «WINDOWS.md:28 — deviation note фазы 12, ожидавшая формального waive владельца» не локализуется: в .planning/WINDOWS.md (текущее слитое дерево, ветка Фазы 12 на дед4c1f и коммит 8656321) НЕТ ни одной строки фазы 12; строка 28 — это строка №11 фазы 05 (indicator verdict unknown), уже закрытая владельцем 2026-10-08 (status: fixed)"
    category: other
    reason: "Пере-проверена вся таблица WINDOWS.md поиском по фазе 12/12-01/term-line — ноль попаданий. Единственный кандидат (строка 28) формально разрешён до предыдущей верификации. Если имелась в виду иная запись — владельцу нужно назвать точный файл/строку; на статус фазы 12 это не влияет"
    evidence_status: "searched current tree + git history (8656321, ded4c1f^2) — no phase-12 row exists"
---

# Phase 12: TERM-строка Verification Report

**Phase Goal:** Пользователь исправляет всю набранную строку/фразу в терминале — там и только там, где capability-таблица Фазы 9 подтверждает примитив захвата строки
**Verified:** 2026-11-06T16:15:12Z
**Status:** passed
**Re-verification:** Yes — STALE-REFRESH on the merged tree (merge ded4c1f: ветка Фазы 12 слита в ветку Фазы 13 после исполнения Фазы 13)

**Причина пере-верификации:** предыдущий отчёт (2026-11-03T14:12:54Z, passed 9/9, digest v4 49f8b3…) устарел по трекинг-файлам — Фаза 13 выросила секцию в adr-008 (+36 строк), пин-контракты test/e2e/spike_journal_test.go (+secA11yProbe), REQUIREMENTS/ROADMAP/STATE/state.json. Субстанция Фазы 12 не менялась: git log по четырём файлам реализации (case_term_line_probe.go, xim-line фикстура, main.go-регистрации, mise.toml) с момента прошлого отчёта — 0 коммитов; изменения spike_journal_test.go чисто аддитивные (пин Фазы 13 добавлен, пин secLineProbe нетронут — проверено diff-ом 8656321..HEAD).

**Форма закрытия фазы:** honest closure by owner decision (без изменений с прошлого раунда). План 12-01 (сел-проба, probe-first гейт) исполнен полностью; designed stop сработал — ячейка (c) RED → checkpoint:decision gate=blocking-human → вердикт владельца 2026-11-03 (дословно в adr-008 ряд «сел-проба закрытие», подтверждён этим раундом). Планы 12-02…12-05 retired superseded (frontmatter `status: superseded, retired: 2026-11-03` во всех четырёх — перечитаны; ROADMAP-аннотации SUPERSEDED «не исполнялся» на месте).

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | [Цель/SC1-исход] Пользователь исправляет всю набранную строку/фразу в терминале | ✓ VERIFIED | Quick-regression (прошлый раунд VERIFIED): машина на месте — `internal/session/actor.go` (round-4: drain-settle :571-575, deferCorrectionFlip :2256, коррекции level-2); именованный тест `TestActor_FlipAfterLevel2Correction` PASS этим раундом (0.266s); `go build ./...` + `go vet ./test/e2e/` чистые на слитом дереве. Поведенческое свидетельство не тронуто слиянием: adr-008 ряд «сел-проба закрытие 2026-11-03» — журнал хоста green106 (160 done / 0 skipped / 0 rejected, жесты до 47 символов, forwards:94). Человеческий оракул исполнен и закрыт 2026-11-03 — это назначенная форма приёмки проекта (владелец — закрывающий оракул) |
| 2 | [SC1-механизм] Тройной правой Shift-машина строки с End-нормализацией и сохранением суффикса | ✓ PASSED (override) | Механизм не строился (гейт 12-01 закрыл до 12-02; планы 12-02…12-05 superseded, «не исполнялся» в ROADMAP). Override принят владельцем 2026-11-03, перенесён дословно (frontmatter overrides); развитие механизма владельцем не запрошено |
| 3 | [SC2] Строка работает только при подтверждённом capability-таблицей примитиве; kill-критерий → честное сужение, вердикт за владельцем | ✓ VERIFIED | Quick-regression: кандидат capability-таблицы (CopyMode→PRIMARY) опровергнут живой пробой и помечен неактуальным (ряд «сел-проба закрытие»: «рецепт CopyMode из capability-таблицы помечается неактуальным для этого применения» — перечитан этим раундом); сужение двумя датированными вердиктами владельца (D-01 2026-10-31 в 12-CONTEXT.md :17; закрытие 2026-11-03); ROADMAP Note несёт вердикт дословно |
| 4 | [SC3] Verify-then-act ветка + замороженные негативные ряды автокоррекции зелёные | ✓ VERIFIED | Quick-regression: первая половина mooted by design (машина выделения не строилась — расхождение длин структурно недостижимо, якорь снят вердиктом D-01/2026-11-03); вторая половина: build/vet зелёные на слитом дереве, именованные тесты PASS (TestSpikeJournalSectionsExist 0.002s, TestActor_FlipAfterLevel2Correction 0.266s); TERM-04-инвариантные ряды живут в прогоняемых пакетах |
| 5 | [T1] Сел-проба имеет датированные ряды трёх ячеек (a)/(b)/(c) в adr-008 | ✓ VERIFIED | Полный перечит секции «Фаза 12 сел-проба канала чтения (term-line-probe)» (:721-750): pending 22:58 → (a) 23:09:36 + abort 23:10:08 (заезд 1) → (a) 23:16:57 / (b) 23:17:01 / (c) 23:17:04 (заезд 2) → итог → mechanics → закрытие. Все грепы verify-блока плана зелёные этим раундом |
| 6 | [T2] Ячейка (c) — гейт: красная → СТОП + checkpoint:decision gate=blocking-human; фолбэк запрещён (D-01) | ✓ VERIFIED | Итог-ряд :748 несёт развилку дословно (СТОП фазы, checkpoint владельцу, D-01-формулировка); 12-02 не стартовал (superseded frontmatter + ROADMAP «не исполнялся»); вердикт владельца recorded дословно с датой (:750); вариант «буфер вместо выделения» отсутствует и в чекпойнте, и в закрытии |
| 7 | [T3] Механика пробы — формы recipeChordRow/wezLegProbeRow; вердикт только из наблюдаемого | ✓ VERIFIED | Код не менялся с прошлой верификации (0 коммитов по файлу): вердикт вычисляется из наблюдаемого `changed` (primaryReadNoNewline после chordSettle), обе ветви green/red честно закодированы (:172-175 физические ячейки, :237-240 форвард); строки журнала совпадают с шаблонами кода; sentinel-раундтрип в mechanics-ряду доказал инструмент до вердикта канала |
| 8 | [T4] Кейс term-line-probe + mise-задача зарегистрированы (main.go реестр + usage ×2, NOT in ci); секция запинена | ✓ VERIFIED | main.go :453 реестр (`noSurfacePreflight: true, watchdog: recipeWatchdog`) + usage :176/:362; mise.toml :209-211 задача e2e-term-line-probe, вне `[tasks.ci]` depends (:35-37 = build/vet/lint/test); spike_journal_test.go :36-39 пин secLineProbe пережил рост Фазы 13 (diff 8656321..HEAD: только добавление secA11yProbe, пин Фазы 12 нетронут); именованный тест TestSpikeJournalSectionsExist PASS на слитом дереве |
| 9 | [T5] Сужение D-01 цитируется вердикт-развилкой дословно; ADR-008 D3 не переоткрывается | ✓ VERIFIED | Код :257-264 (red-ветвь итог-ряда) и :196-203 (комментарий гейта) несут D-01-формулировку дословно; журнал итог-ряд совпадает; в закрытии VTE-примитив не реоткрыт — отвергнут и не предлагался |

**Score:** 9/9 truths verified (8 VERIFIED + 1 PASSED (override); 0 present-but-behavior-unverified)

### Advisory (New Scope, Unevidenced)

Re-verification ran — section included per contract. One carried-forward reference checked, not blocking:

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| 1 | Ссылка на «WINDOWS.md:28 deviation note фазы 12, ожидавшая waive» не локализуется: в .planning/WINDOWS.md нет НИ ОДНОЙ строки фазы 12 (поиск по «фаза 12 / phase 12 / 12-01 / term-line» в текущем дереве, на ветке Фазы 12 при слиянии ded4c1f^2 и в коммите 8656321 — ноль попаданий); строка 28 = строка №11 фазы 05, закрытая владельцем 2026-10-08 (fixed) | other | Координатор просил перенести как advisory, не блок. Формально открытого предмета нет: единственный кандидат уже разрешён до прошлой верификации; если имелась в виду иная запись — требуется точная ссылка от владельца |

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `test/e2e/case_term_line_probe.go` | кейс term-line-probe: три ячейки, honest-abort, privacy | ✓ VERIFIED | 319 строк, неизменен с прошлой верификации; wired: реестр main.go :453 → runTermLineProbe; vet зелёный |
| `test/e2e/fixtures/gs-e2e-wezterm-xim-line.lua` | xim-тело + рецепт CopyMode байт-тот же с wayland | ✓ VERIFIED | 21 строка; diff блока table.insert с wayland-фикстурой: различие ТОЛЬКО в `enable_wayland = false` vs `true` (запланированное тело), рецепт Multiple из 4 CopyMode-действий байт-идентичен; используется кейсом (weztermXimLineFixture :44, ячейки b/c) |
| `mise.toml` | задача e2e-term-line-probe, NOT in ci | ✓ VERIFIED | :209-211; вне ci-задачи — состав ci (:35-37) без e2e |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| case_term_line_probe.go (ячейки a/b) | internal/clipboard (Primary, wl-paste --primary --no-newline) | тот же флаг-набор, что готовый примитив | ✓ WIRED | Кейс вызывает primaryReadNoNewline в каждой ячейке (:168/:233); семантика та же, что Clipboard.Primary — неизменно с прошлой верификации |
| case_term_line_probe.go (ячейка c) | internal/session/probe.go (ProbeForward, keyspec End + shift+home) | форварднутый чорд = последовательность будущего жеста | ✓ WIRED | ProbeForward probe.go :171 на месте; кейс шлёт End-пару прежде shift+home-пары (:224-228) — код не менялся |
| docs/adr/adr-008 (секция «Фаза 12 сел-проба») | 12-02-PLAN.md (машина строки) | вердикт (c) — гейт волны; красный = эскалация, не код | ✓ WIRED | Гейт сработал и остался закрытым: (c) red → СТОП + checkpoint; 12-02 superseded, код машины не писался; секция и ряды пережили слияние без правок |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| adr-008 ряды a/b/c | primary_changed / primary_len / latency | живой прогон на стенде vm-primary (заезд 2, exit 0) | Да — значения (latency 1.005s/874ms/757ms, primary_len=17) наблюдаемы только из исполнения; строки перечитаны этим раундом в слитом дереве | ✓ FLOWING |
| Сводный итог + закрытие | verdict-цепочка | наблюдаемые ячеек + вердикт владельца + журнал хоста green106 | Да — каждое звено датировано и закоммичено (231d443, 8656321); коммиты существуют в слитом дереве | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Пин секции журнала (контракт, пережил слияние) | `mise exec -- go test ./test/e2e -run 'TestSpikeJournalSectionsExist' -count=1` | ok 0.002s | ✓ PASS |
| Машина коррекции level-2 (носитель цели) | `mise exec -- go test ./internal/session -run 'TestActor_FlipAfterLevel2Correction' -count=1` | ok 0.266s | ✓ PASS |
| Компиляция + vet слитого дерева | `mise exec -- go vet ./test/e2e/ && mise exec -- go build ./...` | чисто, exit 0 | ✓ PASS |
| Грепы verify-блока плана (ряды a/b/c/итог/закрытие) | `grep -q 'сел-проба a (физика-wayland)' … b … c … итог … закрытие 2026-11-03 docs/adr/adr-008-spike-log.md` | все 5 присутствуют (8 вхождений «сел-проба») | ✓ PASS |
| Рецепт фикстуры байт-идентичен wayland | diff блока `table.insert` обеих фикстур | различие только `enable_wayland` (запланировано) | ✓ PASS |
| go.mod/go.sum нетронуты | `git log 911773c..HEAD --oneline -- go.mod go.sum` | 0 коммитов | ✓ PASS |
| Коммиты плана в слитом дереве | `git show 12c49fe / 231d443 / 8656321` (наследованы слиянием ded4c1f) | feat/docs-коммиты на месте | ✓ PASS |

Полный `mise run ci` в этом раунде сознательно не повторён: re-verification quick-regression для прошедших истин (build + vet + два именованных теста зелёные; полный ci был зелёным в прошлом раунде, а код фазы с тех пор не менялся — 0 коммитов по файлам реализации).

### Probe Execution

| Probe | Command | Result | Status |
|-------|---------|--------|--------|
| Сел-проба канала (mise-задача `e2e-term-line-probe`) | повторный запуск не выполнялся | — | SKIPPED (live-прогон требует GNOME-сессию стенда vm-primary; запуск на хосте запрещён директивой владельца №6 — прецедент UAF-краха сеанса 2026-10-31. Доказательная форма пробы — датированные ряды журнала, это дизайн фазы: вердикт «никогда не синтезируется из кода, только из наблюдаемого ряда». Все 9 рядов секции перечитаны в слитом дереве, внесены коммитами 231d443/8656321, значения наблюдаемы только из живого исполнения) |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| TERM-03 | 12-01 (frontmatter `requirements: [TERM-03]`); 12-02…12-05 superseded | Пользователь может исправить всю набранную строку/фразу в терминале (растяжка); feasibility-gated, вердикт о сужении за владельцем | ✓ SATISFIED | REQUIREMENTS.md :14 `[x]` с припиской («достигнуто существующей машиной… вердикт владельца 2026-11-03»); трассировка :54 «TERM-03 \| Phase 12 \| Complete»; :67 last-updated. Orphaned-требований нет: на фазу 12 отображён только TERM-03 |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | Debt-маркеры (TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER) в файлах фазы | не найдено | grep по 5 файлам фазы — 0 совпадений |
| — | — | Хост-окружение: голый `mise run` в неактивированной оболочке (PATH-деградация, finding Фазы 13) | ℹ️ Info | Окружение хоста, не дефект репо; этот раунд исполнял гейты через `mise exec --` — все зелёные |

### Prohibition Review

Единственная prohibitions-запись плана (privacy: содержимое выделения никогда не попадает в журнал/логи) — **VERIFIED повторно**: эмиттер `lineProbeFacts` (:127-133) форматрует только сентинел-марку/факт смены/длину/латентность; все ряды фазы в слитом adr-008 перечитаны — только sentinel=GS-SPIKE-SENTINEL / primary_changed / primary_len / clip_frozen / latency; primary_changed=false во всех ячейках (реальное выделение никогда не поднималось). Код не менялся с прошлого раунда.

### Human Verification Required

Нет pending-предметов. Единственный человеческий оракул фазы — checkpoint:decision gate=blocking-human — исполнен и закрыт владельцем 2026-11-03 (вердикт записан дословно с датой в adr-008 ряд «сел-проба закрытие», STATE Decisions, ROADMAP Note, REQUIREMENTS приписка; всё перечитано этим раундом в слитом дереве). Повторное предъявление decided-вопроса не требуется.

### Re-verification Summary (STALE-REFRESH)

Слияние ded4c1f не затронуло субстанцию Фазы 12: (1) файлы реализации — 0 коммитов с момента прошлого отчёта; (2) spike_journal_test.go — чисто аддитивный рост Фазы 13, пин secLineProbe нетронут (named test PASS); (3) adr-008 — секция Фазы 12 байт-нетронута, выросла только секция Фазы 13 ниже; (4) трекинг-файлы (REQUIREMENTS/ROADMAP/STATE) несут закрытие TERM-03 и Фазы 13 без взаимных перезаписей; (5) override владельца перенесён дословно. Регрессий нет.

### Gaps Summary

Gaps отсутствуют. Перенесённая реальность фазы подтверждена на слитом дереве: probe-first гейт исполнен, вердикт-развилка сработала дословно (три RED-ячейки → checkpoint → вердикт владельца), цель TERM-03 достигнута существующей машиной слова Фазы 10 (owner-accepted, журнальные ряды хоста recorded), планы 12-02…12-05 retired superseded без тихих выбрасываний, зелёная база подтверждена точечными проверками этого раунда.

---

_Verified: 2026-11-06T16:15:12Z_
_Verifier: Claude (gsd-verifier) — STALE-REFRESH re-verification on merged tree ded4c1f_
