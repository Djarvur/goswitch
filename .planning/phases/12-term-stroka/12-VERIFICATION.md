---
phase: 12-term-stroka
verified: 2026-11-03T14:12:54Z
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
covered_digest: "v4:sha256:49f8b34288226196b2d5e184471c8ecfff3184b0cb5da1633cec54a671c3ec4e"
behavior_unverified: 0
overrides_applied: 1
overrides:
  - must_have: "SC1 (механизм): тройной правый Shift исправляет всю набранную строку в wezterm — курсор возвращается в исходную позицию, суффюкс справа от курсора цел (End-нормализация задокументирована)"
    reason: "Плановый механизм (машина строки чтением выделения) на живой сел-пробе 12-01 мёртв на корне — все три ячейки RED (физика-wayland / физика-xim / форвард, adr-008 ряды 23:16:57–23:17:04), сентинел-раундтрип доказал: красное — свойство рецепта CopyMode→PRIMARY, не инструментария. Владелец закрыл гейт вердиктом 2026-11-03 (дословно, adr-008 ряд «сел-проба закрытие»): «коррекция всей строки в wezterm у меня работает, а коррекция выделения мне не нужна» — цель SC1 (пользователь исправляет всю строку) достигнута существующей машиной слова Фазы 10 на длинных пробегах (журнал хоста green106 2026-10-31…11-03: 160 done / 0 skipped / 0 rejected, одиночные жесты до 47 символов, forwards:94 = 47 пар BackSpace+commit, машина round-4). Гейт-развилка плана исполнена дословно: checkpoint:decision gate=blocking-human, фолбэк «буфер вместо выделения» не предлагался и не исполнялся (D-01)"
    accepted_by: owner (Daniel Podolsky)
    accepted_at: 2026-11-03
---

# Phase 12: TERM-строка Verification Report

**Phase Goal:** Пользователь исправляет всю набранную строку/фразу в терминале — там и только там, где capability-таблица Фазы 9 подтверждает приматив захвата строки
**Verified:** 2026-11-03T14:12:54Z
**Status:** passed
**Re-verification:** No — initial verification

**Форма закрытия фазы:** honest closure by owner decision. План 12-01 (сел-проба, probe-first гейт) исполнен полностью; его designed stop сработал — ячейка (c) RED → checkpoint:decision gate=blocking-human → вердикт владельца 2026-11-03. Планы 12-02…12-05 retired superseded (ROADMAP: strikethrough + SUPERSEDED-аннотации, файлы сохранены; прецедент-форма 10-14/10-15, фаза 10 верифицирована passed с той же формой). Верификация ниже проверяет ЭТУ реальность: цель достигнута существующей машиной, плановый механизм честно мёртв и документирован, ничего не выброшено тихо.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | [Цель/SC1-исход] Пользователь исправляет всю набранную строку/фразу в терминале | ✓ VERIFIED | Существующая машина на месте и подключена: `internal/session/actor.go` (терминальная ветка round-4: пары press+release, drain-settle, отложенный флип, `slog.Info("correction", "outcome", "done", ...)` :4301/:4391); полный `go test -race ./...` зелёный этим раундом. Поведенческое свидетельство записано: adr-008 ряд «сел-проба закрытие 2026-11-03» — журнал хоста green106 (юнит v1.2.0-alpha.2-83-gda4ced8, 2026-10-31…11-03): 160 коррекций done / 0 skipped / 0 rejected, одиночные жесты до 47 символов (forwards:94 = 47 пар BackSpace+release + commit). Человеческий оракул исполнен: вердикт владельца 2026-11-03 дословно в adr-008 + STATE Decisions :119 + ROADMAP Note. Это назначенная форма приёмки проекта (владелец — закрывающий оракул, прецедент G-10-1) |
| 2 | [SC1-механизм] Тройной правый Shift-машина строки с End-нормализацией и сохранением суффикса | ✓ PASSED (override) | Механизм НЕ строился — гейт 12-01 его закрыл до начала 12-02 (все три ячейки RED: плановый канал чтения мёртв на корне). Override принят владельцем 2026-11-03 (см. frontmatter overrides): цель SC1 достигнута существующей машиной; развитие механизма владельцем не запрошено. Развилка исполнена дословно: СТОП + checkpoint, фолбэк «буфер вместо выделения» не предлагался (D-01) |
| 3 | [SC2] Строка работает только при подтверждённом capability-таблицей примитиве; kill-критерий → честное сужение, вердикт за владельцем | ✓ VERIFIED | Кандидат capability-таблицы (CopyMode→PRIMARY) честно опровергнут живой пробой и помечен неактуальным (adr-008 закрытие-ряд: «рецепт CopyMode из capability-таблицы помечается неактуальным для этого применения»). Сужение — двумя вердиктами владельца: D-01 2026-10-31 (wezterm-only, VTE-примитив отвергнут) и закрытие 2026-11-03. Ни одного тихого решения: каждая развилка датирована и дословна |
| 4 | [SC3] Verify-then-act ветка расхождения + замороженные негативные ряды автокоррекции зелёные | ✓ VERIFIED | Первая половина mooted by design: машина выделения не строилась, расхождение длин структурно недостижимо (якорь SC3 снят тем же вердиктом D-01/2026-11-03 — заявлено в плане 12-01 truth 2 ДО пробы). Вторая половина напрямую: полный `mise run ci` (build+vet+lint+test) зелёный этим раундом в mise-активированной оболочке; TERM-04-инвариантные ряды живут в прогнанных пакетах (internal/session 31.9s ok, test/e2e ok) |
| 5 | [T1] Сел-проба имеет датированные ряды трёх ячеек (a)/(b)/(c) в adr-008 | ✓ VERIFIED | Секция «Фаза 12 сел-проба канала чтения (term-line-probe)»: pending 22:58 → (a) 23:09:36 (заезд 1) + honest-abort 23:10:08 → (a) 23:16:57 / (b) 23:17:01 / (c) 23:17:04 (заезд 2) → итог + mechanics + закрытие. Ячейка (a) наблюдалась дважды — одинаковая красная сигнатура. Коммит 231d443 добавил ровно эти ряды (13 строк), 8656321 — закрытие |
| 6 | [T2] Ячейка (c) — гейт: красная → СТОП + checkpoint:decision gate=blocking-human; фолбэк запрещён (D-01) | ✓ VERIFIED | Итог-ряд журнала: «(c) red: самотриггер мёртв — СТОП фазы, checkpoint владельцу (D-01: фолбэк … запрещён — отвергнутый VTE-примитив по духу, SC3 теряет якорь, ADR-008 D3 не переоткрывается)». 12-02 не стартовал (retired, «не исполнялся» в ROADMAP). Вердикт владельца recorded дословно с датой. Вариант «буфер вместо выделения» отсутствует и в чекпойнте, и в закрытии |
| 7 | [T3] Механика пробы — формы recipeChordRow/wezLegProbeRow; вердикт только из наблюдаемого | ✓ VERIFIED | `case_term_line_probe.go`: вердикт вычисляется из наблюдаемого `changed` (primaryReadNoNewline после chordSettle 400 мс), обе ветви green/red честно закодированы (:172-175, :237-240); строки журнала байт-совпадают с шаблонами кода. Код подтверждает: sentinel-раундтрип доказывал инструмент ДО вердикта канала (mechanics-ряд) |
| 8 | [T4] Кейс term-line-probe + mise-задача зарегистрированы (main.go реестр + usage ×2, NOT in ci); секция запинена | ✓ VERIFIED | main.go :453 реестр (`noSurfacePreflight: true, watchdog: recipeWatchdog`) + usage :176 и :362; mise.toml :209-211 задача `e2e-term-line-probe` («NOT in ci» в описании; `[tasks.ci] depends = ["build","vet","lint","test"]` — e2e-задачи в ci не входят); spike_journal_test.go :35-38 пин secLineProbe; именованный тест зелёный этим раундом |
| 9 | [T5] Сужение D-01 цитируется вердикт-развилкой дословно; ADR-008 D3 не переоткрывается | ✓ VERIFIED | Код :260-262 (итог-ряд red-ветви) и :200-203 (комментарий гейта) несут формулировку дословно; журнал итог-ряд совпадает. В закрытии VTE-примитив не реоткрыт — вариант отвергнут и не предлагался |

**Score:** 9/9 truths verified (8 VERIFIED + 1 PASSED (override); 0 present-but-behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `test/e2e/case_term_line_probe.go` | кейс term-line-probe: три ячейки, honest-abort, privacy | ✓ VERIFIED | 319 строк, все три ячейки + итог + honest-abort обёртка; использует готовые хелперы (не реимплементация); wired: реестр main.go → runTermLineProbe |
| `test/e2e/fixtures/gs-e2e-wezterm-xim-line.lua` | xim-тело + рецепт CopyMode байт-тот же с wayland | ✓ VERIFIED | 21 строка; рецепт (строки 12-20) diff-идентичен рецепту wayland-фикстуры (проверено diff-ом); `enable_wayland = false`; используется кейсом (weztermXimLineFixture :44, ячейки b/c) |
| `mise.toml` | задача e2e-term-line-probe, NOT in ci | ✓ VERIFIED | :209-211; вне `[tasks.ci]` depends — подтверждено составом ci-задачи |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| case_term_line_probe.go (ячейки a/b) | internal/clipboard (Primary, wl-paste --primary --no-newline) | тот же флаг-набор, что готовый примитив | ✓ WIRED | e2e `primaryReadNoNewline` (wezterm.go :578, `wl-paste --primary -n`) — та же семантика, что `Clipboard.Primary` (clipboard.go :108-109, `flagPrimary+flagNoNewline`); кейс вызывает хелпер в каждой ячейке |
| case_term_line_probe.go (ячейка c) | internal/session/probe.go (ProbeForward, keyspec End + shift+home) | форварднутый чорд = последовательность будущего жеста | ✓ WIRED | `ProbeForward` probe.go :171; keyspec-грамматика :23-38 поддерживает `end` + `shift`-модификатор → «End» и «shift+home» валидны; маршрут: runCtl → `goswitchctl probe-forward` (cmd/goswitchctl/main.go :107/:333); кейс шлёт End-пару прежде shift+home-пары (:224-228) |
| docs/adr/adr-008 (секция «Фаза 12 сел-проба») | 12-02-PLAN.md (машина строки) | вердикт (c) — гейт волны; красный = эскалация, не код | ✓ WIRED | Гейт сработал как задуман: (c) red → СТОП + checkpoint владельцу; 12-02 не стартовал, помечен SUPERSEDED «не исполнялся»; код машины не писался — развилка ушла владельцу (Rule-4 форма) |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| adr-008 ряды a/b/c | primary_changed / primary_len / latency | живой прогон на стенде vm-primary (wl-paste --primary чтения, замеры времени, журнал демона key-trace) | Да — значения (1.005s/874ms/757ms, primary_len=17, даты) наблюдаемы только из исполнения; sentinel-раундтрип доказал инструмент | ✓ FLOWING |
| Сводный итог + закрытие | verdict-цепочка | наблюдаемые ячеек + вердикт владельца + журнал хоста green106 | Да — каждое звено датировано и закоммичено (231d443, 8656321) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Пин секции журнала (контракт) | `mise exec -- go test ./test/e2e -run 'TestSpikeJournalSectionsExist' -count=1` | ok 0.007s | ✓ PASS |
| Зелёная итерация на HEAD | `bash -c 'eval "$(mise activate bash)" && mise run ci'` (build+vet+lint+test) | все ступени зелёные, test -race ok все пакеты | ✓ PASS |
| go.mod/go.sum нетронуты | `git log 911773c..HEAD -- go.mod go.sum` + `git diff --stat 911773c..HEAD -- go.mod go.sum` | 0 коммитов, 0 diff | ✓ PASS |
| Коммиты плана существуют | `git show 12c49fe / 231d443 / 8656321` | feat 6 файлов +384; docs журнал +13; docs закрытие +1 — соответствуют SUMMARY | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
|-------|---------|--------|--------|
| Сел-проба канала (mise-задача `e2e-term-line-probe`) | повторный запуск не выполнялся | — | SKIPPED (live-прогон требует GNOME-сессию стенда vm-primary; запуск на хосте запрещён директивой владельца №6 — прецедент UAF-краха сеанса 2026-10-31. Доказательная форма пробы — датированные ряды журнала, что является дизайном фазы: вердикт «никогда не синтезируется из кода, только из наблюдаемого ряда». Ряды presence+content сверены этим раундом: 7 рядов фазы в adr-008, внесены коммитами 231d443/8656321; значения наблюдаемы только из живого исполнения) |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| TERM-03 | 12-01 (frontmatter `requirements: [TERM-03]`); 12-02…12-05 retired | Пользователь может исправить всю набранную строку/фразу в терминале (растяжка); feasibility-gated, вердикт о сужении за владельцем | ✓ SATISFIED | REQUIREMENTS.md :14 checkbox `[x]` с припиской («достигнуто существующей машиной… вердикт владельца 2026-11-03»); трассировка :54 «TERM-03 \| Phase 12 \| Complete». Механизм сужения исполнен честно: кандидат capability-таблицы опровергнут живой пробой, вердикт владельца recorded дословно. Orphaned-требований нет: на фазу 12 отображён только TERM-03 |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| — | — | Debt-маркеры (TBD/FIXME/XXX/TODO/HACK) в файлах фазы | не найдено | grep по всем 5 файлам фазы — 0 совпадений |
| — | — | Хост-окружение: голый `mise run ci` в неактивированной оболочке даёт compile version mismatch (go1.23.12 vs go1.27.3) | ℹ️ Info | Задокументированное окружение хоста (SUMMARY deviation 4), не дефект репо: в mise-активированной оболочке ci полностью зелёный (проверено этим раундом); CI-раннеры без системного go не затронуты |

### Prohibition Review

Одна prohibitions-запись плана (privacy: содержимое выделения никогда не попадает в журнал/логи) — **VERIFIED с детерминированным свидетельством**: (1) эмиттер `lineProbeFacts` (:127-133) форматирует только сентинел-марку/факт смены/длину/латентность; (2) все 7 рядов фазы в adr-008 перечитаны — только sentinel=GS-SPIKE-SENTINEL / primary_changed / primary_len / clip_frozen / latency; (3) primary_changed=false во всех ячейках — реальное выделение никогда не поднималось, единственный PRIMARY-контент был синтетический сентинел. Структурный риск утечки в записанных рядах отсутствует; будущие пробы защищены кодом.

### Human Verification Required

Нет pending-предметов. Единственный человеческий оракул фазы — checkpoint:decision gate=blocking-human — **исполнен и закрыт владельцем 2026-11-03** (вердикт записан дословно с датой в adr-008 ряд «сел-проба закрытие», STATE Decisions :119, ROADMAP Note и REQUIREMENTS приписка). Повторное предъявление decided-вопроса не требуется.

### Gaps Summary

Gaps отсутствуют. Фаза закрыта честной формой: (1) запланированный probe-first гейт исполнен полностью и его вердикт-развилка сработала дословно — плановый механизм чтения строки мёртв на корне, что доказано живой пробой, а не предположением; (2) цель фазы достигнута существующей машиной слова Фазы 10 (owner-accepted, журнальные ряды хоста recorded); (3) отклонение планов 12-02…12-05 от механики SC1 оформлено override-записью владельца в этом отчёте; (4) ничего не выброшено тихо — каждый ряд, вердикт и retirement датированы и закоммичены (12c49fe → 75b07aa, 8 коммитов от базы 911773c); (5) зелёная итерация на HEAD подтверждена этим раундом.

---

_Verified: 2026-11-03T14:12:54Z_
_Verifier: Claude (gsd-verifier)_
