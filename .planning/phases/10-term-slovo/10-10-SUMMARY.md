---
phase: 10-term-slovo
plan: 10
subsystem: terminal-answer-only-routing
tags: [answer-only-signature, stale-cache-match-branch, wr01-correction-busy, wr02-verify-expiry-epoch, host-journal-replay, tdd, gap-closure-round2]

requires:
  - phase: 10-term-slovo plans 01-09
    provides: чанкированный терминальный level-2 с verify-after (10-04), флап-стойкая классификация + identity-якорь + role-диагностика (10-07), defer-until-confirmed + stale-cache mismatch-маршрут (10-08), host-ревалидация с REJECTED вердиктом 2026-10-14 09:39 (10-09), WR-01/WR-02 из 10-REVIEW
provides:
  - Маршрут answer-only клиента (round-2 G-10-1, routing-половина): предикат pendingStaleCache применён на ОБЕИХ ветках вердикта HandleSurroundingText — word-раунд, армированный над ПУСТЫМ push-кэшем, маршрутизируется в терминальную ветку и по свежему совпадающему ответу (match-ветка, хост-ряды 09:39), и по протухшему несовпадению (mismatch-ветка 10-08) — один предикат, байт-та же форма (resolvePending → setVerifyDead(true) → executeLevel2(..., true)); класс учится существующим setVerifyDead, вето D-55 следует за тем же швом classifyTerminal автоматически
  - WR-02 закрыт: pendingFix.epoch + монотонный счётчик pendingEpoch; обе arm-точки через armVerifyDeadline захватывают тег в замыкание таймера; VerifyExpiry(epoch) при чужой эпохе — тихий no-op (форма afterExpiry) — устаревший колбэк больше не съедает следующий раунд и не учит класс контексту, никогда не пропускавшему бюджет
  - WR-01 закрыт: level2Busy checked-and-set до BuildPlan/первого форварда executeLevel2, очистка на всех трёх выходах (backspace-cap, context-changed, успех); повторный вход — skipCorrection("correction-busy") + settleCombo, второй деструктивный BackSpace-бурст по не-заменённому буферу невозможен
  - Реплей-корпус хост-рядов 2026-10-14 09:39 (ReplayHost0939*): красный до правок (evidence RED_EVIDENCE_OK), зелёный после; пины честно-пушащего контраста, подписи отказов (блок 09:39, не 16:20) и единства шва D-55
  - Амендмент ADR-008 2026-10-14: answer-only сигнатура — расширение owner-approved stale-cache семантики 2026-10-13 с содержимого ответа на канал ответа; Evidence — датированные журнальные ряды, без веб-ссылок
affects: [10-11 (delivery-половина round-2: спонтанно-пушащие терминальные клиенты), 10-12 (re-acceptance владельца — закрывающий оракул G-10-1), phase-verification (TERM-02 round-2)]

actuals:
  tokens: 13987        # chars/4 over the realized diff 418c3c4..HEAD (55946 chars, 4 files + evidence-JSON)
  tasks: 2
  commits: 3           # MEASURED: git rev-list --count gsd-plan-head-before-10-10..HEAD

tech-stack:
  added: []
  patterns:
    - "Реплей журнального ряда на юнит-швах: фикстура пинит ДОСЛОВНУЮ подпись блока (role-unknown через ошибающийся UseRole-double, appid unknown, nil-cache канал, свежий match-ответ) и ОТРИЦАТЕЛЬНЫЕ контрасты (пушащий клиент не маршрутизируется; окно чисто от подписей отказа) — латентности из журнала живут только в комментах происхождения, тайминги не синтезируются"
    - "Гонка таймера с раундом на детерминированных швах: arm-точка оборачивает колбэк с тегом эпохи за таймер-швом (UseVerifyTimer, форма flipAfter); корpus удерживает «уже выстреливший» колбэк в recorder'е с no-op-стопом (точная форма продакшн-гонки Stop-после-fire) и стреляет им после армирования следующего раунда"
    - "Гард повторного входа через мьютекс-окно: блокировка ресурса (burst) паркуется сном-швом ПОД мьютексом, второй вход стартуется заранее и ждёт >1 мс (порог starvation-режима sync.Mutex) — handoff на chunk-boundary детерминирован, оба порядка завершения дают один наблюдаемый (ровно один бурст)"

key-files:
  created:
    - ".planning/phases/10-term-slovo/evidence-20261014-10-10-red-task1.json (RED_EVIDENCE_OK Task 1: 4 not ok целевых, 1 защитный пин зелёный)"
  modified:
    - "internal/session/actor.go (match-ветка pendingStaleCache-маршрута; pendingFix.epoch+stop, pendingEpoch, verifyAfter-шов + UseVerifyTimer + armVerifyDeadline, гард VerifyExpiry(epoch); level2Busy с тремя очистками)"
    - "internal/session/actor_test.go (ReplayHost0939* корпус 4 теста + RouteLearnsVetoSeam + VerifyExpiryStaleRound + CorrectionBusy; honestPush-миграция 17 nil-cache пинов уровня-1; teachVerifyDead на UseVerifyTimer)"
    - "internal/ctlsvc/ctlsvc_test.go (миграция TestSvc_CorrectNow на honest-push мир)"
    - "docs/adr/ADR-008-terminal-transport.md (амендмент 2026-10-14: answer-only сигнатура; Reversibility дополнен)"

key-decisions:
  - "Правило владельца 2026-10-13 переносится с СОДЕРЖИМОГО ответа на КАНАЛ ответа: предикат pendingStaleCache не меняется (sel==nil && !acBoundary && cacheNil) — меняется число точек применения (один предикат на обе ветки вердикта); содержимое (свежее/протухшее) не различает, канал различает; selection- и boundary-раунды исключены как и в 10-08"
  - "Тег эпохи раунда — в замыкании arm-точки, а не в сигнатуре потребителя: armVerifyDeadline инкапсулирует counter++/захват/шов в одной критической секции с армированием — ни одно чередование не может спарить колбэк с чужим раундом; живой раунд с совпавшей эпохой ведёт себя байт-как-сегодня (все teachVerifyDead-пины зелёные без правок)"
  - "Гард WR-01 ставится ДО BuildPlan и первого форварда и очищается напрямую на каждом из трёх выходов (без defer — мьютекс отпускается внутри chunkBoundaryMoved); повторный вход честен: D-20-слаг correction-busy + settleCombo, ни одной эмиссии"
  - "Миграция 17+1 nil-cache пинов уровня-1 на honest-push мир (honestPush-идиома, прецедент миграции 10-08) — намерение каждого пина сохранено: кэш-свидетельство при арме, тот же вердикт, тот же op-лог; ни одно ожидание не ослаблено"
  - "RED-коммит несёт инертный скаффолдинг (прецедент 10-07): поля epoch/stop, счётчик, шов UseVerifyTimer, форма armVerifyDeadline и сигнатура VerifyExpiry(epoch) без гарда — целевые тесты падали на УТВЕРЖДЕНИЯХ, не на компиляции"

patterns-established:
  - "answer-only terminal route: подпись клиента (канал без спонтанных пушей) как классификационный вход при мёртвой идентичности — реплей реального журнального ряда красный-до/зелёный-после (урок 2026-10-12 исполнен дословно)"
  - "TAP-evidence конвертер с node-формой summary (# tests/# pass/# fail) — gsd check tdd-red-evidence принимает механический перевод go test -v 1:1"

requirements-completed: [TERM-02]  # shared-ID gate: уже Complete (открыт 10-06); TERM-02 разделяется с 10-11/10-12 (SUMMARY отсутствуют) — expected not-ready, mark-complete не требуется

coverage:
  - id: D1
    description: "Маршрут answer-only клиента на обеих ветках вердикта: nil-cache word-раунд маршрутизируется в терминальную ветку и по свежему match-ответу (ряды 09:39), и по протухшему mismatch (10-08); честно-пушащие клиенты, selection- и boundary-раунды не тронуты;defer-until-confirmed армится на маршруте"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestReplayHost0939_MatchAnswerRoutesTerminalBranch (RED: evidence-20261014-10-10-red-task1.json, RED_EVIDENCE_OK; ops-лог require+burst+require без delete, ноль mode/switch до подтверждения), TestReplayHost0939_RejectionSignaturesAbsent (окно чисто от verify-timeout/verify-mismatch/флап-стираний — блок 09:39, не 16:20; ac_role_errors растёт), TestReplayHost0939_SpontaneousPushClientNotRouted (cacheNil=false держит байт-прежний уровень-1), TestTerminalBranch_StaleCacheAnswerRoutesTerminal/StaleCacheLearnsForNextGesture/HonestPushStillFailClosed/StaleRouteFlipStillDeferred — зелёные БЕЗ правок ожиданий; полный session-корпус зелёный -race"
        status: pass
    human_judgment: false
  - id: D2
    description: "Единство шва classifyTerminal: выученный маршрутом класс автоматически включает вето D-55 — автокоррекция молчит в каждом контексте, записываемом терминальной веткой; оба потребителя (армирующее вето и конвейер записи) читают одно и то же"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestReplayHost0939_RouteLearnsVetoSeam (после маршрута boundary в том же контексте даёт acReasonTerminalVeto, детектор не достигнут); пины TestAutoCorrect_TerminalVeto* корпуса зелёные"
        status: pass
    human_judgment: false
  - id: D3
    description: "WR-02 закрыт: устаревший колбэк VerifyExpiry (выстреливший дедлайн раунда 1, доживший до армирования раунда 2) — тихий no-op; раунд 2 живёт, класс не учится контексту, никогда не пропускавшему бюджет; ложный класс больше не может армировать и машину отложенного флипа"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestVerifyExpiryStaleRound_DoesNotEatNextRound (RED: тот же evidence, RED_EVIDENCE_OK; сценарий 10-REVIEW WR-02 на timerRecorder — no-op-стоп = продакшн Stop-после-fire); teachVerifyDead-зависимый флап/таймаут корпус зелёный без правок (живая эпоха совпадает)"
        status: pass
    human_judgment: false
  - id: D4
    description: "WR-01 закрыт: второй коррекционный вход через мьютекс-окно чанк-границы завершается skipCorrection(\"correction-busy\") + settleCombo — ровно один CommitText, ровно один бурст, ровно один done; флаг не пересекает раунды (очистка на всех трёх выходах)"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestCorrectionBusy_SecondEntrySkipsNotSecondBurst (RED: тот же evidence, RED_EVIDENCE_OK; burst 20 рун через растянутый sleep-шов, CorrectNow через starvation-handoff; RED давал 40 forwards/2 commits/2 done — ровно сценарий второго деструктивного бурста 10-REVIEW WR-01)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Flagged assumption плана: формальная классификация края TERM-02 остаётся unresolved (SPEC без ## Edge Coverage; прецедент 10-04..10-09); routing-половина round-2 закрыта behavior-строками и реплеем реальных рядов, но хост-ряды 09:39 не различают INFO-журналом answer-only vs спонтанно-пушащую модель терминалов — план покрывает ОБЕ маршрутизацией (эта) и доставкой (10-11); закрывающий оракул — re-acceptance владельца в 10-12, не этот план"
    requirement: TERM-02
    verification:
      - kind: other
        ref: ".planning/phases/10-term-slovo/10-10-PLAN.md — Flagged Assumptions + Prohibition probe (10 surfaced = 3 authored + 7 precision-dropped + 1 flagged); {"requirement_id":"TERM-02","category":"unclassified","status":"unresolved","probe":"unclassified"}"
        status: unknown
    human_judgment: true
    rationale: "Flagged assumption никогда не авто-разрешается: verifier обязан классифицировать край явно, не тихий pass; закрывающий оракул G-10-1 — владелец (10-12)"

duration: 48min
completed: 2026-10-14
status: complete
---

# Phase 10 Plan 10: answer-only маршрут терминального контекста + WR-01/WR-02 closed Summary

**Routing-половина round-2 отказа G-10-1 закрыта в коде: предикат pendingStaleCache поднят с mismatch-ветки на ОБЕ ветки вердикта HandleSurroundingText — клиент, который за жизнь контекста ни разу не пушил состояние поля (answer-only канал), маршрутизируется в терминальную ветку и по свежему совпадающему ответу (точная подпись хост-рядов 09:39, где оба жеста владельца уходили уровнем-1 за 1,9–2,4 мс с мгновенным флипом и «привет, и он удаляется»), и по протухшему несовпадению (амендмент 10-08); класс учится существующим setVerifyDead — шов classifyTerminal един, вето D-55 следует автоматически; в зоне, которую маршрут делает горячей, закрыты оба prior-warning ревью (WR-02 epoch-тег VerifyExpiry, WR-01 level2Busy-гард), реплей-корпус реальных журнальных рядов красный до правок и зелёный после — урок 2026-10-12 исполнен дословно.**

## Performance

- **Duration:** 48 min (07:49–08:37 UTC)
- **Started:** 2026-10-14T07:49:19Z
- **Completed:** 2026-10-14T08:37Z
- **Tasks:** 2/2 (Task 1 tracer TDD — реплей-корпус 09:39 + RED-пины WR-01/WR-02, tracer-гейт перезапущен end-to-end: ⚡ Tracer verified end-to-end — expanding; Task 2 auto TDD — GREEN маршрут + epoch + guard + миграции + амендмент ADR-008)
- **Files:** 5 репозиторных (4 из frontmatter + ctlsvc_test.go) + 1 evidence-JSON в .planning; measured diff 418c3c4..HEAD: 55946 chars

## Accomplishments

- **Task 1 (tracer, TDD):** реплей-корпус хост-блока 09:39 на существующих швах — ошибающийся role-источник (UseRole-double; ac_role_errors растёт, машинно видно через StatusSnapshot), appid unknown, клиент ни разу не пушит (cacheNil при арме), Require-ответ свежий и содержащий набранное (pendingVerdict match). Четыре целевых теста красные по точным причинам («терминальная ветка не получила управления» — в ops-логе delete+commit уровня-1 вместо burst; «stale колбэк съел раунд»; «второй бурст прошёл» — 40 forwards/2 commits), защитный пин честно-пушащего контраста зелёный. Пин подписи отказов: окно реплея чисто от verify-timeout/verify-mismatch и флап-стираний класса — воспроизводит ИМЕННО блок 09:39, не 16:20. RED_EVIDENCE_OK (RED-фаза зафиксирована, фикстура таймингов не синтезирует — латентности 1,9–2,4 мс только в комментах происхождения). Инертный скаффолдинг RED-коммита (поля/шов/сигнатура без гарда) — целевые тесты падали на утверждениях, существующий корпус зелёный.
- **Task 2 (TDD GREEN):** match-ветка HandleSurroundingText оборачивает pendingStaleCache байт-той же формой mismatch-ветки 10-08 (resolvePending → setVerifyDead(true) → executeLevel2(..., true); executeCorrection на этом пути не вызывается); коммент-док предиката переформулирован на answer-only сигнатуру (канал, не содержимое). WR-02: pendingFix.epoch + монотонный pendingEpoch, обе arm-точки через armVerifyDeadline (тег + захват в одной критической секции), VerifyExpiry(epoch) — no-op по чужой эпохе. WR-01: level2Busy check-and-set до BuildPlan/первого форварда, очистка на всех трёх выходах, повторный вход — correction-busy + settleCombo. Вето-следствие спинен (RouteLearnsVetoSeam: boundary после маршрута даёт acReasonTerminalVeto). Пины 10-07/10-08 зелёные без правок ожиданий (проверено поштучно: StaleCacheAnswerRoutesTerminal, HonestPushStillFailClosed, VerifyDeadSurvivesFlapPair, FlipDeferredUntilConfirmed, StaleRouteFlipStillDeferred, StaleCacheLearnsForNextGesture).
- **Амендмент ADR-008 2026-10-14:** answer-only сигнатура — расширение одобренной владельцем семантики 2026-10-13 на match-ветку; Evidence — дословные ряды 09:39:41/09:40:28 + ac_role_errors=47 из DEBUG-term-host-live-failures.md, ноль веб-ссылок, D1–D7 байт-прежние, Status Accepted, Reversibility дополнен.
- Строгий TDD (RED 95345c4 → GREEN 7e2d88e, RED_EVIDENCE_OK), зелёная итерация (mise run ci: build+vet+golangci-lint 0 issues+test -race, exit 0) на каждом коммите; go.mod/go.sum нетронуты (git diff --exit-code чист).

## Task Commits

1. **Task 1 (tracer): реплей-корпус 09:39 + RED-пины** — `95345c4` (test: RED, 4 целевых красных + 1 защитный пин + инертный скаффолдинг + evidence RED_EVIDENCE_OK)
2. **Task 2: GREEN маршрут + WR-02 + WR-01 + миграции** — `7e2d88e` (feat: match-ветка маршрута, epoch-гард, level2Busy, honestPush-миграция 18 пинов)
3. **Task 2: амендмент ADR-008** — `689f7f5` (docs: секция 2026-10-14 на журнальных Evidence)

**Plan metadata:** финальный docs-коммит (SUMMARY+STATE+ROADMAP+REQUIREMENTS)

## Files Created/Modified

- `internal/session/actor.go` — match-ветка pendingStaleCache-маршрута (форма 10-08 дословно), pendingStaleCache doc на answer-only канал; pendingFix.epoch+stop, pendingEpoch, verifyAfter+UseVerifyTimer+armVerifyDeadline, VerifyExpiry(epoch) гарда; level2Busy с тремя очистками и слагом correction-busy
- `internal/session/actor_test.go` — ReplayHost0939* (4), RouteLearnsVetoSeam, VerifyExpiryStaleRound_DoesNotEatNextRound (timerRecorder double), CorrectionBusy_SecondEntrySkipsNotSecondBurst; honestPush-идиома + миграция 17 пинов; teachVerifyDead на UseVerifyTimer
- `internal/ctlsvc/ctlsvc_test.go` — TestSvc_CorrectNow мигрирован на honest-push мир (ctlMismatchLine)
- `docs/adr/ADR-008-terminal-transport.md` — амендмент 2026-10-14 (answer-only сигнатура, Evidence ряды 09:39, WR-01/WR-02 закрытия)
- `.planning/phases/10-term-slovo/evidence-20261014-10-10-red-task1.json` — RED_EVIDENCE_OK (TAP-перевод 1:1)

## Decisions Made

См. key-decisions. Ядро: правило владельца переносится с содержимого ответа на канал ответа (один предикат, две точки применения); тег эпохи живёт в замыкании arm-точки за таймер-швом; гард busy ставится до первой эмиссии и очищается на каждом выходе; nil-cache пины уровня-1 мигрированы на honest-push мир без ослабления ожиданий.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Инертный скаффолдинг в RED-коммите (файл вне files_modified Task 1)**
- **Found during:** Task 1, RED-фаза
- **Issue:** RED-тесты WR-02 требуют новой сигнатуры VerifyExpiry(epoch) и шва достоверного выстрела «уже стартовавшего» колбэка — без поверхностей тесты не компилировались бы (INVALID_RED по форме); files_modified Task 1 называл только actor_test.go
- **Fix:** в RED-коммит вошли инертные поверхности (pendingFix.epoch/stop, pendingEpoch, verifyAfter+UseVerifyTimer, armVerifyDeadline, сигнатура VerifyExpiry(epoch) БЕЗ гарда — поведение байт-прежнее); гард лег в GREEN 7e2d88e
- **Files modified:** internal/session/actor.go (сопровождает actor_test.go)
- **Verification:** check tdd-red-evidence → RED_EVIDENCE_OK; существующий корпус зелёный при RED-коммите
- **Commit:** 95345c4

**2. [Rule 3 - Blocking] Миграция 18 nil-cache пинов уровня-1 (session + ctlsvc_test — файл вне files_modified)**
- **Found during:** Task 2, GREEN-прогон полного корпуса
- **Issue:** расширение маршрута на match-ветку делает nil-cache+match терминальной сигнатурой — 20 тестов, армящих раунд над пустым кэшем с уровневым-1 settlement (DoubleTapCorrects, verifyAfterSettled, settleWordCorrection/settledWordCorrection хелперы, RU/phrase/mixed/combo корпусы, TestSvc_CorrectNow), словили терминальный burst вместо delete+commit
- **Fix:** honestPush-идиома (один спонтанный пуш до жеста = кэш-свидетельство; прецедент миграции 10-08 «на honest-push мир») — намерение каждого пина сохранено байт-прежним, ни одно ожидание не ослаблено; пере-пины 10-07/10-08 не тронуты (проверены поштучно)
- **Files modified:** internal/session/actor_test.go, internal/ctlsvc/ctlsvc_test.go
- **Verification:** полный session-корпус + ctlsvc зелёные -race; mise run ci exit 0
- **Commit:** 7e2d88e

---

**Total deviations:** 2 (все Rule 3 auto-fixed, оба — форма исполнения: TDD-скаффолдинг и локализация миграции; семантика задач не менялась). **Impact on plan:** нулевой на контракт задач — обе задачи исполнены, критерии приёмки закрыты.

## Issues Encountered

- funlen-потолок (40 стейтментов) на TestReplayHost0939_MatchAnswerRoutesTerminalBranch после добавления вето-пина — пин выделен в отдельный именованный тест RouteLearnsVetoSeam (лучше и по форме «один behavior — один тест»).
- gsd check tdd-red-evidence требует node-TAP-форму summary (# tests/# pass/# fail) — конвертер evidence дополнен summary-строками (прецедент 10-02..10-08 продолжен).
- Хост-демон владельца не тронут (юнит-тесты + код + docs only; дерево чистое, чужие untracked-файлы не тронуты).
- Честная оговорка по must_haves.truths: хост-ряды 09:39:41 и 09:40:28 несут ОДНУ сигнатуру (action n=2, nil-cache канал, свежий match-ответ) — реплей-корпус пинит сигнатуру одним тестом; «оба жеста» покрыты идентичностью подписи, не двумя независимыми фикстурами.

## User Setup Required

None — no external service configuration required. (Закрывающий оракул G-10-1 — re-acceptance владельца на живом столе, план 10-12, после 10-11.)

## Next Phase Readiness

- Готово для 10-11 (delivery-половина round-2): answer-only клиенты маршрутизируются терминально (defer-until-confirmed + чанкированный level-2 стоят на пути), спонтанно-пушащие терминальные клиенты (cacheNil=false) остаются на уровне-1 — их доставка чинится в 10-11.
- Открытые хвосты для верификатора: D5 (flagged assumption TERM-02, human_judgment — как в 10-04..10-09); T-10-10-01 (честный, но медленный GTK-клиент с первым пушем позже арма маршрутизируется терминально один раз — направление отказа безопасно, задокументировано в амендменте ADR-008); WR-03/WR-04 ревью 10-REVIEW вне объёма этого плана.
- Блокеров нет; TERM-02 остаётся Complete (shared-ID gate открыт 10-06; 10-11/10-12 ещё не закрыты — expected not-ready).

---
*Phase: 10-term-slovo*
*Completed: 2026-10-14*

## Self-Check: PASSED

SUMMARY exists; все 3 коммита от базы плана 418c3c4 верифицированы в git log (95345c4/7e2d88e/689f7f5; ledger gsd-plan-head-before-10-10 на диске, rev-list count = 3); файлы созданы (evidence-20261014-10-10-red-task1.json, SUMMARY) и модифицированы (actor.go, actor_test.go, ctlsvc_test.go, ADR-008) — все FOUND; TDD-гейт: test(10-10) 95345c4 + feat(10-10) 7e2d88e в логе (check tdd-red-evidence → RED_EVIDENCE_OK по TestReplayHost0939_MatchAnswerRoutesTerminalBranch); ReplayHost0939*/VerifyExpiryStaleRound/CorrectionBusy зелёные под -race; mise run ci зелёный на HEAD (build+vet+golangci-lint 0 issues+test -race, exit 0); go.mod/go.sum нетронуты (git diff --exit-code); grep-гейты: correction-busy в skipCorrection executeLevel2 (:3594), pending.epoch в обеих arm-точках через armVerifyDeadline (:3394/:3502); ADR-008 секция 2026-10-14 с рядами 09:39, ноль http(s)-упоминаний, D1–D7 байт-прежние (diff только аддитивный +58/−1); пины 10-07/10-08 зелёные без правок ожиданий; stub-скан чист (TODO/FIXME/placeholder — ноль); deletions в diff плана — ноль; хост-демон владельца не тронут.
