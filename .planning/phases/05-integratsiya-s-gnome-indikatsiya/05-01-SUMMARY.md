---
phase: 05-integratsiya-s-gnome-indikatsiya
plan: 01
subsystem: e2e-spike-and-architecture-records
tags: [ibus, gnome-shell, set-global-engine, two-source, adr, e2e-spike, d52]

requires:
  - phase: 04-postavka-i-priyomka
    provides: living daemon/engines on the owner's desk (goswitch-en/goswitch-ru registered), e2e stand snapshot/restore discipline (test/e2e/main.go), mise live-case convention
provides:
  - live spike verdicts P1/P2/P4 (two byte-identical PASS runs) + P3 unresolved row with routing
  - ADR-006-two-engine-revision.md (ADR-001 Option B → Option A, Status Proposed, d52-literal)
  - resolved D-52 checkpoint: criterion-2 spec-delta wording; rejected alternatives recorded (sources-rewrite, mru-seed→v1.1)
  - obligations for 05-04 (sync listener + self-echo suppression, P4-proven) and 05-05 (live mechanism proofs incl. P3 re-open, README, UAT)
affects: [05-03, 05-04, 05-05, verify-work]

actuals:
  tokens: 16197   # chars/4 over the 05-01-scoped diffs (64788 chars, 4 commits)
  tasks: 3
  commits: 4      # 05-01-scoped; ledger window bd71533..HEAD holds 10 incl. 6 interleaved 05-02 commits (see Issues)
plan_head_before: bd71533a13ff184cfe10bfd74ad96702c1fd2796

tech-stack:
  added: []
  patterns:
    - "Live-desktop spike window: snapshotDesktop → reversible two-source write (write-only-on-diff) → named probes with per-probe baseline restore → defer-restore verified machine-side"
    - "Closed-set verdict rendering (follows / not-follows / error) — probe outcomes are machine classes, never free text (D-20/D-21)"
    - "Journal-pinned mode pinning: daemon mode is pinned by \"msg\":\"mode\" records, not by assumptions about defaults"

key-files:
  created:
    - docs/adr/ADR-006-two-engine-revision.md
    - test/e2e/case_switch.go
    - test/e2e/case_switch_test.go
    - .planning/phases/05-integratsiya-s-gnome-indikatsiya/red-evidence/05-01-task1-red.json
  modified:
    - test/e2e/main.go
    - test/e2e/case_install.go
    - mise.toml
    - .gitignore

key-decisions:
  - "D-52 resolved d52-literal (owner, 2026-09-30): flip = SetGlobalEngine engine-truth on the daemon's ibus connection; criterion 2 reworded via spec-delta — «индикатор отражает активный источник при пользовательских жестах; daemon-флип меняет ввод немедленно, ярлык — при следующем пользовательском жесте»"
  - "ADR-006 Status Proposed; Accepted lands with plan 05-05 after the live mechanism proof"
  - "05-04 obligations pinned by P4 ×2: sync listener consumes GlobalEngineChanged; self-echo suppression mandatory"
  - "sources-rewrite rejected (dconf weight per tap, MRU inversion race, live engine-unset per write, contradicts D-52's letter); d52-plus-mru-seed deferred to v1.1 candidate (research OQ2)"
  - "P3 XKB-truth left explicitly unresolved (environmental focus-refusal ×2); re-opens at 05-05 live proofs; indicator-verdict unknown routes to UAT"
  - "ADR-006 Status line carries the ACTUAL verdict date 2026-09-30 (plan placeholder said 2026-09-28 — the checkpoint resolved in the night session 2026-09-29→30)"

patterns-established:
  - "Spike verdict report as gitignored artifact with verbatim citation into the ADR (report never carries typed text — engine names/classes/verdicts only)"
  - "ADR evidence table rows carry the probe's verbatim verdict line + run multiplicity (×2) + explicit unresolved routing"

requirements-completed: [SWCH-03]

coverage:
  - id: D1
    description: "Кейс switch-spike: обратимое двухисточниковое окно, матрица проб P1–P4, вердикт-отчёт + реестр main.go + mise-задача e2e-switch-spike"
    requirement: SWCH-03
    verification:
      - kind: unit
        ref: "mise exec -- go test ./test/e2e/ -run 'TestRenderSwitchVerdict|TestSwitchSpikeProbesOrdered' -race -count=1 -v"
        status: pass
      - kind: e2e
        ref: "mise run e2e-switch-spike (2026-09-29T21:08–21:11Z и 2026-09-29T21:15–21:16Z; оба PASS exit 0, вердикт-строки байт-идентичны, verifyRestored байт-в-байт)"
        status: pass
      - kind: integration
        ref: "mise run ci (build + vet + golangci-lint 0 issues + go test -race ./...) — 2026-09-29T21:25Z"
        status: pass
    human_judgment: false
  - id: D2
    description: "P1 shell-activation — шелл принимает/активирует goswitch-движок из двухисточникового списка"
    verification:
      - kind: e2e
        ref: "switch-spike-report.txt: 'P1 shell-activation | readback=goswitch-en focus_in(en)=+1 focus_in(ru)=+0 (sources already two-source, write skipped) | follows' ×2"
        status: pass
    human_judgment: false
  - id: D3
    description: "P2 external-flip-signal — GlobalEngineChanged доставляется на подписанное godbus-соединение (research A2 подтверждено живьём)"
    verification:
      - kind: e2e
        ref: "switch-spike-report.txt: 'P2 external-flip-signal | flip=goswitch-ru readback=goswitch-ru signal=GlobalEngineChanged(goswitch-ru) | follows' ×2"
        status: pass
    human_judgment: false
  - id: D4
    description: "P4 self-echo — инициатор получает собственный SetGlobalEngine как сигнал (research A3 разрешено: подавление эха обязательно в 05-04)"
    verification:
      - kind: e2e
        ref: "switch-spike-report.txt: 'P4 self-echo | flip=goswitch-en initiator-signal=GlobalEngineChanged(goswitch-en) | follows' ×2"
        status: pass
    human_judgment: false
  - id: D5
    description: "ADR-006-two-engine-revision.md: ревизия ADR-001 Option B → Option A, d52-literal, таблица доказательств с цитатами спайка, spec-delta критерия 2, обязательства 05-04, отклонённые альтернативы"
    verification:
      - kind: other
        ref: "grep-gate ADR-SKELETON-OK: пять секций формата ADR-001 + D-52 + SetGlobalEngine; acceptance-токены: D-53/D-54 blockquotes, 2× ссылки на switch-spike-report.txt, en_pinned/RU-пинный/d52-literal"
        status: pass
    human_judgment: false
  - id: D6
    description: "P3 xkb-truth — XKB-истина за внешним SetGlobalEngine (EN-пинный readback)"
    verification: []
    human_judgment: true
    rationale: "Проба завершилась error ×2 экологически: entry-поверхность спайка не взяла фокус (witness «gnome-shell:WINDOW:chars=-1», повтор идентичен на простаивающем столе и на авторизованном перезапуске P3); строка НЕРАЗРЕШЕНО зафиксирована в таблице ADR-006 с маршрутом — переоткрывается живыми доказательствами механизма (план 05-05); диагностика отказа фокуса — отложенная follow-up работа"
  - id: D7
    description: "Индикатор (критерий 2, SWCH-03): меняется ли ярлык при daemon-флипе — глаза владельца в 5-секундных паузах P3"
    verification: []
    human_judgment: true
    rationale: "Владелец «не разглядел» (indicator-verdict: unknown); вынесено на приёмку фазы (UAT/verify-work); семантика критерия 2 зафиксирована spec-delta независимо от наблюдения"
  - id: D8
    description: "Разрешение чекпойнта Task 2: выбор варианта механизма владельцем"
    verification: []
    human_judgment: true
    rationale: "Решение вынесено и записано владельцем 2026-09-30 (d52-literal + вердикты по индикатору/XKB); машинная верификация неприменима — факт решения зафиксирован в ADR-006, STATE decision log и настоящем SUMMARY"

duration: 134min
completed: 2026-09-30
status: complete
---

# Phase 05 Plan 01: switch-spike tracer — вердикт двухисточникового окна + ADR-006 Summary

**Живой спайк (два PASS-прогона, байт-идентичные вердикты) развязал D-52 в d52-literal; ADR-006 ревизует ADR-001 Option B → Option A с таблицей живых доказательств; 05-04 обязан подавлять self-echo (P4), XKB-истина (P3) переоткрывается в 05-05.**

## Performance

- **Duration:** 134 min (metric записан на закрытие; два раунда — см. ниже)
- **Started:** 2026-09-29T19:12:45Z (раунд 1; agent-history sequential-05-01)
- **Completed:** 2026-09-30 (раунд 2, UTC ~21:35 / MSK ~00:35; живые прогоны раунда 1: 21:08–21:16Z; чекпойнт Task 2 разрешён владельцем в ночной сессии 2026-09-29→30)
- **Tasks:** 3 (Task 1 RED+GREEN, Task 2 live runs + чекпойнт, Task 3 ADR-006)
- **Files modified:** 8 (05-01-scoped; отчёт спайка gitignored)

## Accomplishments

- Кейс `switch-spike` в реестре стенда и mise (`e2e-switch-spike`): обратимое двухисточниковое окно, пробы P1–P4, вердикт-отчёт из закрытого множества — стол после кейса байт-в-байт равен снимку в обоих прогонах.
- Центральный вопрос фазы решён живьём ДО зависимого кода: шелл принимает/активирует goswitch-движок из двух источников (P1 ×2), GlobalEngineChanged доставляется подписчику (P2 ×2 — sync-listener 05-04 viable), инициатор получает собственный сигнал (P4 ×2 — подавление эха обязательно); XKB-истина (P3) — явно неразрешённая строка с маршрутом в 05-05.
- ADR-006-two-engine-revision.md создан в формате ADR-001: Status Proposed (d52-literal, вердикт владельца 2026-09-30; Accepted — план 05-05), цитаты D-52/D-53/D-54 дословно, таблица доказательств со ссылками на вердикты спайка, spec-delta критерия 2, отклонённые альтернативы с причинами.
- Обязательства зависимых планов зафиксированы: 05-03 — шов флипа через SetGlobalEngine; 05-04 — sync-listener + эхо-подавление; 05-05 — живые доказательства механизма (включая P3 и re-assert сценарии), README, UAT.

## Task Commits

1. **Task 1 (RED): корпус рендерера/порядка проб** - `682666c` (test; red-evidence RED_EVIDENCE_OK)
2. **Task 1 (GREEN): живой кейс switch-spike P1–P4 + реестр + mise** - `da6519b` (feat)
3. **Task 2: живой прогон спайка + gitignore отчёта** - `d8839bb` (chore; PASS exit 0 21:08–21:11Z; авторизованный перезапуск P3 21:15–21:16Z — байт-идентично, без коммитов)
4. **Task 3: ADR-006 двухдвижковая ревизия ADR-001** - `e973ab1` (docs)

**Plan metadata:** this commit (docs(05-01): complete switch-spike plan)

## Files Created/Modified

- `docs/adr/ADR-006-two-engine-revision.md` - ADR ревизии ADR-001: Status Proposed, таблица доказательств (7 строк, ×2 прогоны), spec-delta, обязательства 05-04/05-05, отклонённые альтернативы
- `test/e2e/case_switch.go` - кейс switch-spike: twoSources-окно, пробы P1–P4, renderVerdictTable, собственное godbus-соединение (P2/P4)
- `test/e2e/case_switch_test.go` - TestRenderSwitchVerdict, TestSwitchSpikeProbesOrdered
- `test/e2e/main.go` - реестр pickCase + caseListUsage («switch-spike», одна правка)
- `test/e2e/case_install.go` - сопутствующая правка реестра GREEN-шага
- `mise.toml` - задача e2e-switch-spike (live GNOME session; NOT in ci)
- `.gitignore` - switch-spike-report.txt (вердикты процитированы в ADR дословно)
- `.planning/phases/05-integratsiya-s-gnome-indikatsiya/red-evidence/05-01-task1-red.json` - RED-доказательство (RED_EVIDENCE_OK)

## Decisions Made

- **d52-literal (владелец, 2026-09-30)** — флип = SetGlobalEngine (engine-truth); критерий 2 переформулирован spec-delta: «индикатор отражает активный источник при пользовательских жестах; daemon-флип меняет ввод немедленно, ярлык — при следующем пользовательском жесте». Отклонены: sources-rewrite (dconf-тяжесть на каждый тап, MRU-инверсия, живой сброс движка на каждой записи, противоречие букве D-52), d52-plus-mru-seed (отложена кандидатом v1.1 — гонка с `_updateMruSettings`, OQ2).
- **ADR-006 Status Proposed** — Accepted ставит план 05-05 после живого доказательства механизма; в Status записана фактическая дата вердикта 2026-09-30 (плановый шаблон предполагал 2026-09-28 — чекпойнт разрешился в ночную сессию).
- **P3 зафиксирован как неразрешённая строка ADR, а не блокер** — экологическая причина (отказ фокуса entry-поверхности) повторилась идентично дважды; владелец авторизовал перезапуск P3 (21:15–21:16Z), результат байт-идентичен; вопрос маршрутизирован в 05-05, диагностика — отложенная follow-up работа.

## Deviations from Plan

None — plan executed exactly as written (дата вердикта в Status ADR — фактическая 2026-09-30 вместо планового плейсхолдера 2026-09-28; см. Decisions Made).

**Total deviations:** 0
**Impact on plan:** none

## Issues Encountered

- **P3 environmental focus-refusal (два прогона, идентичный witness «gnome-shell:WINDOW:chars=-1»):** entry-поверхность спайка не взяла фокус — XKB-истина осталась неразрешённой; RU-пинный readback (safety-net) нечитаем по той же причине. Диагностика отложена; вопрос переоткрывается в 05-05. Записано в WINDOWS (unmet-truth).
- **indicator-verdict: unknown** — владелец не разглядел индикатор в паузах P3; маршрутизировано на UAT/verify-work. Записано в WINDOWS (unmet-truth).
- **Живой спайк не перезапускался в этом раунде** — оба авторизованных прогона PASS с байт-идентичными вердиктами; повторный запуск мутировал бы стол владельца без новой информации. Плановая верификация `mise run e2e-switch-spike` считается покрытой записанными прогонами.
- **Переплетение коммитов в окне ledger** (`gsd-plan-head-before-05-01` = bd71533): пока 05-01 стоял на чекпойнте Task 2, план 05-02 исполнился на той же ветке — окно bd71533..HEAD содержит 10 коммитов, из которых 6 принадлежат 05-02 (свой SUMMARY их уже учёл). Собственные коммиты 05-01: 4 (682666c, da6519b, d8839bb, e973ab1) — учтены в actuals.commits.
- **Кросс-ссылка на фазовый гейт:** интерпретация владельцем half-wrapped-refusal остаётся открытой в WINDOWS (запись 05-02, #9) — к разрешению на verify-work фазы; планом 05-01 не разрешается.

## User Setup Required

None - no external service configuration required. (Живые кейсы требуют GNOME-сессию владельца — прогон дважды авторизован и выполнен.)

## Next Phase Readiness

- **05-03 (шов флипа) разблокирован:** механизм выбран (d52-literal), порядок записей mode→switch_engine и дедлайн из Pitfall 4 зафиксированы в ADR-006.
- **05-04 (sync) получил пиннинг из живых вердиктов:** подписка GlobalEngineChanged viable (P2 ×2); подавление self-echo обязательно (P4 ×2).
- **05-05 должен закрыть:** P3 XKB-истина (переоткрыта), re-assert сценарии (ibus-restart/password-field/sources-rewrite), README-документацию третьего xkb-источника и известного поведения индикатора, повышение ADR-006 в Accepted, UAT-пункт по индикатору.
- **verify-work фазы:** пункты human_judgment этого плана (D6 XKB, D7 индикатор, D8 решение) + WINDOWS-записи #9 (05-02 install-cycle) и двух новых unmet-truth записей 05-01.

---
*Phase: 05-integratsiya-s-gnome-indikatsiya*
*Completed: 2026-09-30*

## Self-Check: PASSED

- Files found: docs/adr/ADR-006-two-engine-revision.md, test/e2e/case_switch.go, test/e2e/case_switch_test.go, test/e2e/main.go, mise.toml, 05-01-SUMMARY.md
- Commits found: 682666c (RED), da6519b (GREEN), d8839bb (chore/gitignore), e973ab1 (ADR-006), db25751 (WINDOWS ledger) + this metadata commit (docs(05-01): complete switch-spike tracer plan)
- Plan <verification> re-run this round: unit corpus `TestRenderSwitchVerdict|TestSwitchSpikeProbesOrdered` -race PASS; `mise run ci` PASS (build/vet/golangci-lint 0 issues/test -race, 2026-09-29T21:25Z); ADR-006 grep-gate ADR-SKELETON-OK + acceptance tokens OK; live `mise run e2e-switch-spike` covered by the two recorded authorized PASS runs (byte-identical verdicts, desktop restored machine-verified)
- Task acceptance_criteria: Task 1 — report carries P1–P4 verdict lines + indicator-verdict, no dconf-as-oracle, restoration machine-checked, no typed text in logs/report (engine names/classes only); Task 2 — owner verdict recorded (d52-literal, 2026-09-30); Task 3 — five ADR-001-format sections, owner verdict in Status, evidence rows P1–P4 with report references, EN-pinned readback cited for XKB with RU-pinned as a separate safety-net row, D-52/D-53/D-54 blockquotes, SetGlobalEngine present
