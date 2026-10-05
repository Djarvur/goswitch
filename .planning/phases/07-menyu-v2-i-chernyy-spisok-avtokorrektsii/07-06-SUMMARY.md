---
phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
plan: 06
subsystem: testing
tags: [e2e, dbusmenu, blocklist, autocorrect, ydotool-free, live-desktop, goswitchctl]

# Dependency graph
requires:
  - phase: 07-02/07-03/07-04/07-05/07-08
    provides: blocklist schema (apps_blocklist), persist writer + adopt+watch, actor blocklist gate (app-blocked slug), menu v2 (pinned IDs EN=1/RU=2/AC=4/Sound=5), sound engine
  - phase: 06-avtokorrekcija-opcionalno
    provides: the autocorrect live-case shapes, the GTK4 password fixture, matrix v4, startAutocorrectConfigDaemon ladder
provides:
  - runAutocorrectBlocklistSilent — the live blocklist negative on the OBSERVED fixture identity (fired=0, one ac_skip_app_blocked, field verbatim)
  - runMenuV2 — the ydotool-free menu proof over com.canonical.dbusmenu.Event: EN/RU mode records, both persisted toggles file+status in the same click, restart survival of BOTH enabled values
  - Actor.FoldAppliedConfig — the eager fold seam; the toggles' apply and the STARTUP document are in force without waiting for a key event
  - startAutocorrectConfigDaemon(document) — the case owns its config document
  - matrixACConfigYAML on apps_blocklist: [] (CASES-FROZEN held: matrix-v4.yaml byte-untouched)
  - mise tasks e2e-autocorrect-blocklist-silent, e2e-menu-v2 (live-only, NOT in ci)
affects: [07-07 (release readiness — the live acceptance run), verify-work phase 07]

# Actuals (#2632) — same estimateTokens scale (chars/4 over the realized diff).
actuals:
  tokens: 42800
  tasks: 3
  commits: 9

# Tech tracking
tech-stack:
  added: [] # zero new dependencies — gdbus (OS) and the existing yaml.v3 only
  patterns:
    - "DBusMenu Event-drive e2e: click = gdbus call on the daemon's own /Menu, oracles = journal marks + file + ctl status — never ydotool/mouse"
    - "Eager fold seam: the ONE fold path (applySnapshot) invoked at the two pinned moments — toggle click and startup — the 03-04 per-event succession unchanged for every other consumer"
    - "Witness-form oracles only: counts and fact-matches, never the word (D-20/D-21 on the stand surface)"

key-files:
  created:
    - test/e2e/case_menu.go
    - test/e2e/case_menu_test.go
  modified:
    - test/e2e/case_autocorrect.go
    - test/e2e/main.go
    - test/e2e/matrix.go
    - test/e2e/matrix_test.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - cmd/goswitchd/main.go
    - cmd/goswitchd/main_test.go
    - mise.toml

key-decisions:
  - "The observer-START round stays in fires/password-silent (LIVE finding of the first repinned run): the confirm's live GetRole reads the a11y observer's stored (sender, path) pair — without the round the layer abstains role-unknown. Only observeFixtureApp (the identity string in the config) is gone"
  - "The 07-05 «synchronous apply» pin is now true end to end: the toggle's reload folds the actor (FoldAppliedConfig) — the RED was found live (status lagged one key event)"
  - "The startup document is in force AT STARTUP (newActor folds after AttachConfig) — the restart-survival oracle demanded it; the 03-04 per-event succession is unchanged for every other consumer"
  - "The blocklist negative pins the plain exact-string form of the observed identity (the anchored ^…$ alternative documented as the author's choice)"
  - "menu-v2 clicks RU first: the daemon starts in EN and flipTo's same-target guard would no-op an EN click first — both radio halves still proven by their mode records"

patterns-established:
  - "Event-click helper (eventClick): synchronous over D-Bus — when it returns, the daemon-side dispatch (flip, persist, apply) has completed"
  - "Per-click full-state oracle (menuToggleStep): document AND status read exactly the wanted pair after every click, intermediate states included"

requirements-completed: ["MACR-ACL (blocklist-ревизия)", "CORR-01..09 (регресс)", "SWCH-01..04 (регресс)"]

coverage:
  - id: D1
    description: "Fires/password-silent/terminal-silent repinned to the blocklist schema and green live (manual-double CORR-01 regression alive)"
    requirement: "CORR-01..09 (регресс)"
    verification:
      - kind: e2e
        ref: "mise run e2e-autocorrect-fires # PASS (fired + manual double converted back)"
        status: pass
      - kind: e2e
        ref: "mise run e2e-autocorrect-password-silent # PASS (one role-forbidden abstention)"
        status: pass
      - kind: e2e
        ref: "mise run e2e-autocorrect-terminal-silent # PASS (fired=0, zero correction records)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Live blocklist negative: the observed fixture identity in apps_blocklist forbids the correction (field verbatim, fired=0, exactly one ac_skip_app_blocked)"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: e2e
        ref: "mise run e2e-autocorrect-blocklist-silent # PASS (witness TEXT:chars=0→7, PLAIN oracle byte-exact)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Matrix v4 regression on the new config_base; matrix-v4.yaml byte-untouched (CASES-FROZEN)"
    requirement: "CORR-01..09 (регресс)"
    verification:
      - kind: e2e
        ref: "mise run e2e-matrix-v4 # 32/34 PASS ×3 — the two reds are the standing WINDOWS #12 rows (word-mixed/phrase-mixed, no config_base, byte-identical documented actuals; acceptance point 07-07)"
        status: pass
      - kind: unit
        ref: "test/e2e/matrix_test.go#TestMatrixCaseBaseConfig"
        status: pass
      - kind: other
        ref: "git diff --exit-code -- test/e2e/cases/matrix-v4.yaml"
        status: pass
    human_judgment: false
  - id: D4
    description: "Menu v2 live proof without ydotool: EN/RU mode records, both toggles persist+apply synchronously, both enabled values survive a daemon restart"
    requirement: "SWCH-01..04 (регресс)"
    verification:
      - kind: e2e
        ref: "mise run e2e-menu-v2 # PASS (radio records, synchronous toggles, restart survival, off restore)"
        status: pass
    human_judgment: false
  - id: D5
    description: "The toggle's synchronous apply and the startup document truth (the product seams the live oracles exposed and fixed)"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_FoldAppliedConfig"
        status: pass
      - kind: unit
        ref: "cmd/goswitchd/main_test.go#TestNewActor_StartupFoldInForce"
        status: pass
    human_judgment: false

# Metrics
duration: 86 min
completed: 2026-10-05
status: complete
---

# Phase 7 Plan 6: e2e-репины blocklist-семантики и Event-драйв меню Summary

**Живые кейсы автокоррекции на blocklist-семантике (fires без white-list-хореографии, новый негатив на наблюдаемой идентичности), ydotool-независимый кейс меню v2 с автоматическим доказательством «переживает перезапуск», матричная база на apps_blocklist: [] при байт-целом YAML — плюс два найденных живьём продукта-пробела синхронного применения, закрытые через RED→GREEN.**

## Performance

- **Duration:** 86 min
- **Started:** 2026-10-05T00:48:10Z
- **Completed:** 2026-10-05T05:14Z (live runs) / close-out 05:20Z
- **Tasks:** 3 (+2 TDD RED commits on the exposed product seams)
- **Files modified:** 11

## Accomplishments

- Три перезакреплённых автокоррекционных кейса зелёные живьём на новой семантике: fires стреляет без white-list-наблюдения (manual-double CORR-01 жив), password-silent держит одну role-forbidden абстенцию, terminal-silent молчит на no-caps/роли.
- Новый живой негатив autocorrect-blocklist-silent: наблюдаемая идентичность фикстуры (org.gtk.application.gs_e2e_password) в apps_blocklist → поле дословно (witness 0→7 рун + PLAIN-оракул байт-в-байт), fired=0, ровно одна ac_skip_app_blocked.
- Новый кейс menu-v2 через com.canonical.dbusmenu.Event (никакого ydotool/мыши): клики EN/RU дают журнальные mode-записи, оба тумблера (автокоррекция id 4, «Звук» id 5) пишут файл И переворачивают goswitchctl status тем же кликом, оба включения переживают рестарт демона на том же -config, обратные клики восстанавливают off.
- Матрица v4: config_base на apps_blocklist: [] — 32/34 PASS ×3 (два красных — стоящие открытые WINDOWS #12 строки, не этого плана; см. Deviations). matrix-v4.yaml байт-нетронут (git-diff гейт зелёный все три прогона).
- Два продукта-пробела, найденные живыми оракулами и закрытые TDD: синхронное применение тумблера (Actor.FoldAppliedConfig + wiring) и истинность стартового документа при старте/рестарте (newActor startup fold).

## Task Commits

Each task was committed atomically (TDD sub-commits included):

1. **Task 1: репины case_autocorrect.go** - `b54b3af` (feat)
2. **Task 2: blocklist-негатив + матрица + mise** - `3b31ae0` (test RED), `4d8f6f0` (feat)
3. **Task 3: case_menu.go Event-драйв + mise e2e-menu-v2** - `cbe12f7` (test RED), `57fee91` (test RED — fold seam), `b1cac39` (feat — fold), `3f79bd3` (test RED — startup fold), `9676466` (feat — startup fold), `1bd75b3` (feat)

## Files Created/Modified

- `test/e2e/case_autocorrect.go` — blocklist-документы (enabled + пустой apps_blocklist), startAutocorrectConfigDaemon(документ), упрощённые fires/password-silent/terminal-silent, новый runAutocorrectBlocklistSilent + acBlocklistOracle
- `test/e2e/case_menu.go` (new) — runMenuV2, eventClick, menuToggleStep/menuRadioRound, parseMenuToggles, waitMenuToggles/menuStatusHas; пин ID 1/2/4/5
- `test/e2e/case_menu_test.go` (new) — TestParseMenuToggles (файловый оракул)
- `test/e2e/main.go` — регистрация autocorrect-blocklist-silent и menu-v2 (три списка)
- `test/e2e/matrix.go` — matrixACConfigYAML: apps_blocklist: [] + комментарий
- `test/e2e/matrix_test.go` — TestMatrixCaseBaseConfig перезакреплён (RED→GREEN)
- `internal/session/actor.go` + actor_test.go — Actor.FoldAppliedConfig (RED→GREEN)
- `cmd/goswitchd/main.go` + main_test.go — toggle-reload fold + startup fold (RED→GREEN)
- `mise.toml` — [tasks.e2e-autocorrect-blocklist-silent], [tasks.e2e-menu-v2] (live-only)

## Decisions Made

- Наблюдение идентичности (observeFixtureApp) удалено из fires/password-silent; acStartObserverRound ОСТАВЛЕН во всех армящих кейсах — живой GetRole читает (sender, path) парy наблюдателя (см. Deviations 1).
- Blocklist-негатив пинит точную строку моста (не анкерованную форму) — выбор задокументирован в комментарии кейса.
- Стартовый документ меню: явная секция sound.enabled: false (форма «существующей секции» — оракул инверсии однозначен; nil-means-on исключён из оракула).
- Продуктовые фиксы — минимальные швы: fold только на тумблерах и при старте; CLI reload и пункт «Перечитать конфиг» сохраняют пер-эвент succession (03-04/03-07 пины не тронуты).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] acStartObserverRound остаётся в fires/password-silent (план требовал удалить всю хореографию)**
- **Found during:** Task 1 (первый живой прогон перезакреплённого fires)
- **Issue:** живой confirm-GetRole читает сохранённую (sender, path) пару a11y-наблюдателя демона; без раунда наблюдения фикстура мапится до старта наблюдателя, пара пуста, Role()==unknown — слой абстентен role-unknown, fires красный («fired record: timeout»)
- **Fix:** acStartObserverRound возвращён в fires и password-silent (и нужен в blocklist-негативе); удалена только observeFixtureApp (строка идентичности в конфиге больше не нужна)
- **Files modified:** test/e2e/case_autocorrect.go
- **Verification:** fires/password-silent/blocklist-silent зелёные живьём (повторно на финальном дереве)
- **Committed in:** b54b3af

**2. [Rule 1 - Bug] «Синхронное применение» тумблера (пин 07-05) не держалось конец-в-конец**
- **Found during:** Task 3 (живой прогон menu-v2, шаг в)
- **Issue:** flip писал файл и обновлял снапшот watcher'а, но опции актора (токены статуса autocorrect_enabled/sound_enabled и фактическое значение) фолдились только по следующему клавиатурному событию — статус лгал после клика
- **Fix:** Actor.FoldAppliedConfig (тот же applySnapshot под мьютексом — единственный путь применения, вызван ажорно) + замыкание toggle-reload в daemon wiring; CLI reload и «Перечитать конфиг» не тронуты (их пины 03-04/03-07)
- **Files modified:** internal/session/actor.go, internal/session/actor_test.go, cmd/goswitchd/main.go
- **Verification:** TestActor_FoldAppliedConfig (RED 57fee91 → GREEN b1cac39); menu-v2 живой зелёный
- **Committed in:** 57fee91, b1cac39

**3. [Rule 1 - Bug] Стартовый документ не был в силе при старте/рестарте**
- **Found during:** Task 3 (живой прогон menu-v2, шаг г — рестарт-переживание)
- **Issue:** после рестарта на переключённом документе статус отвечал autocorrect_enabled=false sound_enabled=false против true/true на диске: опции автокоррекции вообще не попадали в стартовый SetOptions, а статусное поле звука фолд-мейнетт (читает a.soundEnabled, заполняемый только фолдом); изолированный репро подтверждает
- **Fix:** newActor фолдит присоединённый документ сразу после AttachConfig — статус и меню говорят правду документа с первого чтения
- **Files modified:** cmd/goswitchd/main.go, cmd/goswitchd/main_test.go
- **Verification:** TestNewActor_StartupFoldInForce (RED 3f79bd3 → GREEN 9676466); шаг г menu-v2 зелёный
- **Committed in:** 3f79bd3, 9676466

**4. [Rule 3 - Blocking] Временный lint-разрыв между задачами 1 и 2 (санкционирован планом)**
- **Found during:** Task 1
- **Issue:** observeFixtureApp/acStartObserverRound временно без живого потребителя → unused-линт
- **Fix:** порядок задач плана 1→2 закрыл (потребитель — blocklist-негатив); lint зелёный с задачи 2 и до конца
- **Committed in:** 4d8f6f0

**5. [Rule 3 - Blocking] Оракулы кейса menu-v2 уточнены живыми прогонами (тело плана vs механика)**
- **Found during:** Task 3
- **Issue:** (а) план вел кликать EN первым — демон стартует в EN, same-target guard flipTo молчит, записи не будет → первый клик RU; (б) промежуточное состояние после первого обратного клика false/TRUE (план не называл); (в) eventClick без параметра stand (unparam); (г) полный литерал метода com.canonical.dbusmenu.Event вынесен константой для греп-гейта
- **Fix:** все четыре формы исправлены в кейсе; суть оракулов плана (mode-записи, файл+статус, рестарт) сохранена дословно
- **Committed in:** 1bd75b3

---

**Total deviations:** 5 auto-fixed (3 × Rule 1 bug — два из них продуктовые TDD-пары, 2 × Rule 3 blocking)
**Impact on plan:** все отклонения — уточнение механики, найденной живыми оракулами (назначение e2e-плана); locked-оракулы плана (fired=0/app_blocked=1/witness/переживает рестарт) выполнены дословно. Продуктовые фиксы минимальны и законтрактованы юнит-корпусом.

## Issues Encountered

- Матрица v4: 32/34 на финальном дереве (×3 прогона, один после ibus restart). Два красных — word-mixed/phrase-mixed: стоящий открытый WINDOWS #12 (с 05-05: ожидание коррекции ЧЕРЕЗ границу флипа против спроектированной flip-resets-context семантики ADR-006). Строки без config_base — диффа этого плана механически не касается; readback байт-в-байт равен задокументированному в реестре актуалу. Приёмочная точка — план 07-07 (formal fresh-session D-48). Не фиксилось этим планом (scope boundary).
- Два средовых сбоя живых прогонов (таймаут регистрации компонента): (1) гонка сразу после ibus restart, (2) сиротский IBus-именной держатель — собственный диагностированный probe-демон этой сессии; вылечено документированным ibus-restart + снятием процесса. goswitchd.service остановлен на прогоны и восстановлен (активен, engine reactivated).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Готово к 07-07 (release readiness): полный живой набор зелёный на финальном дереве (4 автокоррекционных кейса + menu-v2), матрица v4 32/34 с двумя задокументированными WINDOWS #12 строками, CASES-FROZEN цел, mise-задачи live-only.
- Открытые пункты для verify-work/07-07: WINDOWS #12 (верdict владельца) и формальный свежесессионный D-48 двойной прогон — приёмочная точка плана 07-07; оба не блокируют кодовое дерево.

---
*Phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii*
*Completed: 2026-10-05*

## Self-Check: PASSED

- Created files exist on disk: test/e2e/case_menu.go, test/e2e/case_menu_test.go (FOUND); modified files verified via git log.
- All 10 plan commits exist: b54b3af, 3b31ae0, 4d8f6f0, cbe12f7, 57fee91, b1cac39, 3f79bd3, 9676466, 1bd75b3, ef0b46e (FOUND).
- Acceptance re-runs on the final tree: e2e-autocorrect-{fires,password-silent,terminal-silent,blocklist-silent} PASS; e2e-menu-v2 PASS; e2e-matrix-v4 32/34 ×3 (two standing WINDOWS #12 rows); mise run ci green; CASES-FROZEN git-diff gate green; new live tasks absent from [tasks.ci].
