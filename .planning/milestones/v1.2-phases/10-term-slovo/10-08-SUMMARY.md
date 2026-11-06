---
phase: 10-term-slovo
plan: 08
subsystem: terminal-delivery-order
tags: [deferred-flip, defer-until-confirmed, stale-cache-amendment, client-drift, e2e-flip-case, stand-repin, tdd, gap-closure, rule4-owner-approved]

requires:
  - phase: 10-term-slovo plans 01-07
    provides: чанкированный терминальный level-2 с verify-after без ранга (10-04), живые ряды term-word-vte/-wezterm/-negative-witnesses (10-05/10-06), флап-стойкая классификация + identity-якорь + role-диагностика (10-07), стенд goswitch-stand (/tmp/stand-env.sh, rsync-рецепт, unit -debug), UAT-разрыв G-10-1 причина B (вердикт E3)
provides:
  - Машина отложенного флипа терминальной ветки (defer-until-confirmed, E3-контракт): SetGlobalEngine не следует за commit — флип армится и освобождается подтверждающим push / следующей клавишей / grace-дедлайном flipConfirmGrace 300 мс (плюс mismatch/quiet-timeout verify-after); ни один release-сайт не сидит на FocusOut — флап-пара причины A физически не воспроизводит E3; правило владельца «режим следует за скриптом результата» сохранено, D-36 порядок записи не дублируется (release переиспользует flipTo целиком)
  - Клиент-дрейф амендмент (owner-approved Rule 4, 2026-10-13): nil-cache word-раунд, чей единственный ответ не содержит набранного пробега, — терминальный контекст; класс учится (setVerifyDead) и раунд исполняет терминальную ветку в этом же ходе; fail-closed для контекстов с кэш-свидетельством (честный GTK push-трекинг) сохранён байт-прежним
  - Живой кейс term-word-vte-flip: механизм ряда отложенного флипа с журнальным оракулом ПОРЯДКА доставки (done < settlement < mode/switch; SetGlobalEngine между commit и settlement = E3-сигнатура = FAIL) + CLIPBOARD-сентинел под активной машиной флипа; регистрация main.go (реестр+usage+pickCase) и mise-задача e2e-term-word-vte-flip (NOT in ci)
  - Машинная активация стенда: гейт waitFocusInFrom провоцирует пару ydotool-кликом по центру после тихих секунд (manual-activation шаг автоматизирован; сигнал продолжения — focus_in демона, честность гейтов не ослаблена); арминг-гейт VTE несёт ограниченную ретрай-лестницу wezterm-формы (SPEC §7.2)
  - Re-pin трёх замороженных рядов байт-точными зелёными при коде амендмента + flip-кейс зелёный на одной свежей сессии (boot 21:27:35); дрейф 0x9→0x29 задокументирован датированными рядами (стабилен across ibus restart и reboot ВМ)
affects: [phase-verification (причины A и B закрыты в коде; host-ревалидация — за владельцем/10-09), 10-09 (host-ревалидация и owner-приёмка)]

actuals:
  tokens: 58000        # chars/4 over the realized diff 13e0a70..4b65c17 (13 файлов, +1352/−17 код-дерево + журнальные/плановые записи)
  tasks: 2
  commits: 10          # MEASURED: git rev-list --count gsd-plan-head-before-10-08..HEAD

tech-stack:
  added: []
  patterns:
    - "Отложенное действие с тремя release-триггерами и grace-дедлайном: слот {target, epoch} + one-window таймер (форма flapAfter/level2Sleep), освобождение переиспует существующий исполнитель целиком (flipTo) — D-36 порядок не дублируется, ровно одно исполнение на раунд"
    - "Stale-cache сигнатура клиента: снапшот состояния push-кэша в момент армирования раунда (cacheNil) + вердикт ответа против собственного буфера демона — обучение поведенческого класса из ДОСТАВКИ ответа, а не только из тишины (расширение verifyDead-семантики 10-03)"
    - "Машинная активация стенда: провокация focus-пары ydotool-кликом внутри гейта ожидания (manual-activation шаг автоматизирован), сигнал продолжения — прежний (focus_in демона над базлайном); честность оракула не делегируется клику"

key-files:
  created:
    - "test/e2e/case_term_word_vte_flip.go (кейс term-word-vte-flip: оракул порядка доставки + CLIPBOARD-сентинел + собственный конфиг-документ flip=true)"
    - ".planning/phases/10-term-slovo/evidence-20261013-10-08-red-task1.json (RED_EVIDENCE_OK Task 1, 6 целевых тестов)"
    - ".planning/phases/10-term-slovo/evidence-20261013-10-08-red-task2.json (RED_EVIDENCE_OK Task 2-амендрмент, 3 целевых теста)"
  modified:
    - "internal/session/actor.go (flipPending/flipEpoch/flipStop/flipAfter + deferCorrectionFlip/releaseDeferredFlip/flipGraceExpiry + pendingFix.cacheNil + pendingStaleCache + 5 release-сайтов)"
    - "internal/session/actor_test.go (deferred-flip корпус: 6+4 тестов, flipClock double; миграция трёх GTK-mismatch пинов на honest-push мир)"
    - "test/e2e/case_term_word_vte.go (арминг-лестница termArmGateTries=3, SPEC §7.2)"
    - "test/e2e/wezterm.go (waitFocusInFrom + nudgeFocusClick: машинная провокация фокус-пары)"
    - "test/e2e/main.go (реестр+usage+pickCase: term-word-vte-flip)"
    - "mise.toml (задача e2e-term-word-vte-flip, NOT in ci)"
    - "docs/adr/ADR-008-terminal-transport.md (амендмент 2026-10-13: delivery-order safety + клиент-дрейф семантики)"
    - "docs/adr/adr-008-spike-log.md (секции: flip-ряды, дрейф+амендмент, re-pin механика вечера)"
    - ".planning/phases/10-term-slovo/10-08-PLAN.md (дельта-пометка owner-approved Rule 4)"

key-decisions:
  - "SetGlobalEngine терминальной коррекции вызывается только после подтверждения доставки: подтверждающий push (тот же критерий match, что у verify-after) / первая клавиша (строго до её обработки — ввод продолжается в режиме результата) / flipConfirmGrace 300 мс (16 мс здорового E3-push × 20 запаса; внутренняя константа, не конфиг); mismatch и тихий таймаут verify-after тоже освобождают — исход verify не владеет правилом режима"
  - "Ни один release-сайт не сидит на FocusOut/Reset — отложенный флип переживает контекстную смерть (правило режима глобально), флап-пара причины A не может вызвать SetGlobalEngine и воспроизвести E3"
  - "Владелец одобрил семантический амендмент по контрольной точке (Rule 4; варианты 2/3/4 отвергнуты дословно): nil-cache + ответ-без-набранного-пробега = терминальный контекст; выбор исполнять коррекцию В ТОМ ЖЕ раунде (а не только учить класс) — прямой смысл директивы «терминальная ветка должна маршрутизировать в терминальный путь»"
  - "Fail-closed инвариант сохранён скоупом: cacheNil-снапшот distinguishes never-pushing клиентов (wezterm-класс) от push-трекинга (честный GTK); selection- и acBoundary-раунды исключены; три GTK-mismatch пина мигрированы на honest-push мир — их nil-cache форма и есть амендированная сигнатура"
  - "Settlement-рубеж flip-кейса — «push match / verify» по букве плана: стендовая VTE отвечает на Require протухшим кэшем (live finding первого прогона), честный ряд принимает оба заключения verify-after; WR-05 stale-retry — промежуточное, не settlement"

patterns-established:
  - "RED-корпус с инертным швом (UseFlipTimer без поведения) — целевые тесты падают на утверждениях (10-07 форма продолжена)"
  - "Журнальный оракул ПОРЯДКА для асинхронных машин: индексы записей между рубежами (done/settlement/mode) вместо точного совпадения — ловит E3-класс перестановок, не ломаясь о промежуточные записи (stale-retry)"

requirements-completed: [TERM-02]  # shared-ID gate: уже Complete (открыт 10-06) — 0/1 ready, mark-complete не требуется

coverage:
  - id: D1
    description: "Отложенный флип терминальной ветки: SetGlobalEngine не следует за commit ни в какой фазе доставки — release на подтверждающий push / следующую клавишу / grace 300 мс / mismatch / тихий таймаут; ровно один флип на коррекцию; D-36 порядок (mode record → switcher → panel symbol) переиспользован из flipTo; FocusOut/Reset не трогают слот"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestTerminalBranch_FlipDeferredUntilConfirmed (RED: evidence-20261013-10-08-red-task1.json, RED_EVIDENCE_OK; ops-лог burst+commit+require без mode), FlipReleasedByNextKey (release строго до обработки: [... mode(ru) commit(ч)]), FlipGraceDeadlineFires (ровно один, stale-fire no-op, 300 мс пин), FlipReleasedOnMismatchAndTimeout, SecondCorrectionReleasesPendingFlip (ops-эквивалентность с release-опом между раундами), FlipSurvivesFocusOutGraceReleases; полный session-корпус зелёный -race"
        status: pass
    human_judgment: false
  - id: D2
    description: "Клиент-дрейф амендмент (owner-approved Rule 4): nil-cache word-раунд с ответом без набранного пробега учитывает класс и маршрутизирует в терминальную ветку в том же раунде; fail-closed для кэш-свидетельств; learning живёт на следующий жест"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestTerminalBranch_StaleCacheAnswerRoutesTerminal (RED: evidence-20261013-10-08-red-task2.json, RED_EVIDENCE_OK; маркер terminal class + done + level-2 + отсутствие verify-mismatch), StaleCacheLearnsForNextGesture (решение на decision), HonestPushStillFailClosed (защитный пин fail-closed), StaleRouteFlipStillDeferred (E3-контракт на новом пути); мигрированные пины VerifyPaths/PhraseVerifyMismatch/RefusalsNeverFlip зелёные"
        status: pass
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md секция wezterm: word-a 21:31:32 learning=true (амендмент виден живым рядом — первый жест 0x29-контекста скорректирован, pty_delta байт-точен), word-b learning=false"
        status: pass
    human_judgment: false
  - id: D3
    description: "Живой механизм-ряд term-word-vte-flip: flip_after_correction=true (собственный конфиг-документ), pty «привет» байт-точен, журнальный порядок done < settlement < mode/switch подтверждён греп-гейтом кейса; SetGlobalEngine между done и settlement = FAIL; CLIPBOARD-сентинел байт-нетронут под активной машиной флипа"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md секция flip: flip-word-bash 21:32:54 (pty_hex=d0bfd180d0b8d0b2d0b5d1820a; delivery_order=done<settlement<mode), flip-clip-frozen 21:32:57 (clip_before=clip_after=GS-SPIKE-SENTINEL; pty_hex=67686264746e0a — набор в перевёрнутом режиме, коррекция вернула латиницу); exit кейса 0; кейс зарегистрирован в main.go (реестр+usage+pickCase) и mise.toml (e2e-term-word-vte-flip, NOT in ci)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Re-pin трёх замороженных рядов байт-точными зелёными при коде амендмента (машина отложенного флипа при их flip=false пинах не армится; амендмент маршрутизирует 0x29-ночник терминально) — на одной свежей сессии (boot 21:27:35)"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md: term-word-vte rows 21:30:49-56 (pty_hex дословно рядов 10-05) PASS exit 0; term-word-wezterm rows 21:31:29-40 PASS exit 0; term-negative-witnesses rows 21:32:13-33 (fired=0, veto Δ=1, буферы дословно) PASS exit 0; заезды вечера и механика — ряды repin-mechanics/wezterm-abort/drift-probe секции дрейфа"
        status: pass
    human_judgment: false
  - id: D5
    description: "Flagged assumption плана: формальная классификация края TERM-02 остаётся unresolved (SPEC без ## Edge Coverage); владелец УЖЕ вынес вердикт по семантике амендмента (контрольная точка 10-08, вариант 1 одобрен) — формальная классификация края остаётся за фазовой верификацией; host-ревалидация (missing #5 UAT) — за 10-09/владельцем"
    requirement: TERM-02
    verification:
      - kind: other
        ref: ".planning/phases/10-term-slovo/10-08-PLAN.md — Flagged Assumptions + дельта-пометка owner-approved Rule 4 2026-10-13"
        status: unknown
    human_judgment: true
    rationale: "Flagged assumption никогда не авто-разрешается: verifier обязан классифицировать край явно, не тихий pass"

duration: 141min
completed: 2026-10-13
status: complete
---

# Phase 10 Plan 08: закрытие причины B — delivery-order safety и клиент-дрейф амендмент Summary

**Причина B разрыва G-10-1 закрыта в коде: пост-коррекционный флип терминальной ветки больше не осиротяет in-flight commit (defer-until-confirmed: SetGlobalEngine только после подтверждающего push / клавиши / grace 300 мс; release-сайты не сидят на FocusOut — флап-пара причины A физически не воспроизводит E3), и владелец одобренным Rule 4-амендрментом сделал рабочий инструмент из ночной дрейф-подписи клиента (0x9→0x29): nil-cache раунд с ответом-без-набранного-слова теперь маршрутизируется в терминальную ветку с обучением класса; жёсткое ограничение разрыва удержано — три замороженных ряда re-pin'нуты байт-точными зелёными, новый механизм-ряд term-word-vte-flip зелёный с журнальным оракулом порядка доставки на той же свежей сессии.**

## Performance

- **Duration:** 141 min (19:15-21:36 UTC; включая контрольную точку Task 2 с вердиктом владельца)
- **Started:** 2026-10-13T19:15:02Z
- **Completed:** 2026-10-13T21:36:27Z
- **Tasks:** 2/2 (Task 1 tracer TDD — машина отложенного флипа, tracer-гейт перезапущен end-to-end зелёным; Task 2 auto — стенд: re-pin ×3 + новый flip-кейс + амендменты, с owner-approved Rule 4 дельтой)
- **Files:** 13 репозиторных (7 из frontmatter + main.go/wezterm.go/case_term_word_vte.go/10-08-PLAN.md/adr-008-spike-log.md) + 3 evidence/плановых записи в .planning

## Accomplishments

- **Task 1 (tracer, TDD):** машина отложенного флипа — executeLevel2 терминальной ветки АРМИТ флип (те же гарды: comboPending, FlipAfterCorrection, target=scriptOf(converted), target≠mode) вместо немедленного settleCorrectionFlip; освобождение — подтверждающий push (match-ветка settleAfter), первая клавиша (вход feedKey, строго до обработки), grace 300 мс, mismatch-ветка и afterExpiry; слот очищается при release (никогда два открытых), epoch-гард устаревшего грейса (форма afterExpiry), таймер-шов UseFlipTimer (форма UseFlapTimer). Юнит-корпус: 6 целевых тестов (RED_EVIDENCE_OK) + 3 защитных пина (flip=false байт-идентичен, combo немедленный, level-1 пины зелёные без правок). Tracer-гейт: verify перезапущен end-to-end зелёным (auto_advance) — ⚡ Tracer verified end-to-end — expanding.
- **Контрольная точка Task 2:** re-pin term-word-wezterm вскрыл дрейф стенда — nightly минтит caps 0x29 (бит анонсируется) и отвечает на Require протухшим кэшем; стабильность доказана (ibus restart + reboot ВМ; пакеты тех же дат, что зелёные ряды 10-06). Возвращена контрольная точка (Rule 4); владелец одобрил вариант 1 (семантический амендмент), 2/3/4 отвергнуты дословно.
- **Task 2 (TDD амендмент + стенд):** pendingStaleCache — nil-cache word-раунд (клиент ни разу не пушил состояние поля) с ответом-без-набранного-пробега учитывает класс и исполняет терминальную ветку в том же раунде; fail-closed сохранён для кэш-свидетельств (3 GTK-пина мигрированы на honest-push мир). Стенд: re-sync, четыре кейса зелёными на одной свежей сессии — vte (байт-точные ряды 21:30), wezterm (word-a learning=true — амендмент виден живым рядом), witnesses (fired=0, вето держит), flip (порядок done<settlement<mode, pty байт-точен, CLIPBOARD нетронут).
- Строгий TDD на обеих кодовых задачах (RED_EVIDENCE_OK ×2), зелёная итерация (mise run ci: build+vet+golangci-lint 0 issues+test -race) на каждом коммите; go.mod/go.sum нетронуты.

## Task Commits

1. **Task 1 (tracer): машина отложенного флипа** — `5a47555` (test: RED, 6 целевых + 3 пина + evidence), `bfdeaf1` (feat: defer-until-confirmed машина)
2. **Контрольная точка** — `12632f8`, `6cd09f7` (docs: state-записи ожидания вердикта владельца)
3. **Task 2: амендмент + стенд** — `b53b44e` (test: RED stale-cache маршрута, 3 целевых + защитный пин + миграция пинов + evidence), `442f69d` (feat: pendingStaleCache маршрут), `9cc620e` (feat: flip-кейс + машинный nudge + регистрация + mise), `87f3012` (docs: амендмент ADR-008 + дельта плана), `0bcf0fd` (docs: ряды стенда + арминг-лестница VTE), `4b65c17` (fix: settlement-гейт flip-кейса принимает тихий таймаут)

**Plan metadata:** финальный docs-коммит (SUMMARY+STATE+ROADMAP+REQUIREMENTS)

## Files Created/Modified

- `internal/session/actor.go` — flipPending/flipEpoch/flipStop/flipAfter, deferCorrectionFlip/releaseDeferredFlip/cancelFlipGrace/flipGraceExpiry, flipConfirmGrace (300 мс), release-сайты (settleAfter match/mismatch, afterExpiry, feedKey), pendingFix.cacheNil, pendingStaleCache
- `internal/session/actor_test.go` — deferred-flip корпус (flipClock double, 6 тестов), stale-cache корпус (4 теста), миграция VerifyPaths/PhraseVerifyMismatch/RefusalsNeverFlip на honest-push мир
- `test/e2e/case_term_word_vte_flip.go` — новый кейс: оракул порядка доставки (assertFlipDeliveryOrder), settlement = match|timeout (WR-05 stale-retry — промежуточное), clip-ряд в перевёрнутом режиме
- `test/e2e/wezterm.go` — waitFocusInFrom + nudgeFocusClick (ydotool центр-клик, machine-активация)
- `test/e2e/case_term_word_vte.go` — termArmGateTries=3 ретрай-лестница
- `test/e2e/main.go` + `mise.toml` — регистрация term-word-vte-flip
- `docs/adr/ADR-008-terminal-transport.md` — амендмент 2026-10-13 (delivery-order + клиент-дрейф), Reversibility дополнен
- `docs/adr/adr-008-spike-log.md` — секция flip (2 ряда), секция дрейфа/амендмента (5 рядов), 14 re-pin рядов вечера, 2 mechanics-ряда
- `.planning/phases/10-term-slovo/10-08-PLAN.md` — дельта-пометка owner-approved Rule 4

## Decisions Made

См. key-decisions. Ядро: флип АРМИТСЯ, а не исполняется — SetGlobalEngine после подтверждения доставки (три триггера + grace); ни один release не сидит на lifecycle; амендмент семантики — по вердикту владельца (Rule 4, «рабочий инструмент»), fail-closed сохранён скоупом cacheNil; settlement-рубеж — «push match / verify» по букве плана (стендовая VTE отвечает протухшим кэшем — тихий таймаут честная форма).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] settlement-рубеж flip-кейса: стендовая VTE не подтверждает push'ем**
- **Found during:** Task 2, первый прогон flip-кейса (21:10:45)
- **Issue:** VTE отвечает на RequireSurroundingText протухшим кэшем («x» арминг-пробы) — verify-after заключался WR-05 stale-retry → тихий таймаут; мой первичный гейт ждал только «match»
- **Fix:** settlement = match | timeout по букве плана («push match / verify»); WR-05 stale-retry помечен промежуточным; оракул порядка не ослаблен (mode/switch между done и settlement = FAIL)
- **Files modified:** test/e2e/case_term_word_vte_flip.go
- **Verification:** flip-кейс PASS (ряды 21:32:54/21:32:57)
- **Commit:** 4b65c17

**2. [Rule 3 - Blocking] Файлы за пределами files_modified: main.go, wezterm.go, case_term_word_vte.go**
- **Found during:** Task 2
- **Issue:** регистрация кейса требует правок main.go (реестр+usage+pickCase); свежая загрузка стенда не отдаёт map-focus VTE-фикстурам и молча съедает первые ydotool-вызовы — гейты адаптированы машинно (nudgeFocusClick; арминг-лестница wezterm-формы)
- **Fix:** waitFocusInFrom провоцирует пару кликом после тихих секунд (сигнал продолжения — focus_in демона); termArmGate несёт termArmGateTries=3 (SPEC §7.2)
- **Files modified:** test/e2e/wezterm.go, test/e2e/case_term_word_vte.go, test/e2e/main.go
- **Verification:** все гейты четырёх кейсов зелёные на свежей сессии; nudge срабатываний tonight — 0 (гейты прошли сами), лестница — 0 (доставка восстановилась reboot'ом)
- **Commit:** 9cc620e, 0bcf0fd

**3. [Rule 4 - Architectural → владелец] Семантический амендмент маршрутизации (дрейф клиента)**
- **Found during:** Task 2, re-pin term-word-wezterm
- **Issue:** дрейф 0x9→0x29 + Require-ответы протухшим кэшем убивали обучение класса — замороженный ряд не мог исправляться
- **Fix:** возвращена контрольная точка (Rule 4); владелец одобрил вариант 1 — амендмент исполнен TDD (RED b53b44e / GREEN 442f69d), план несёт дельта-пометку, ADR-008 — амендмент
- **Verification:** 3 целевых теста RED_EVIDENCE_OK; word-a learning=true живым рядом
- **Commit:** b53b44e, 442f69d, 87f3012

---

**Total deviations:** 3 (2 Rule 3 auto-fixed; 1 Rule 4 — решён владельцем в рамках исполнения). **Impact on plan:** нулевой на контракт задач — обе задачи исполнены, критерии приёмки закрыты.

## Issues Encountered

- Вечерняя деградация uinput-доставки стенда: 5 честных FAIL-заездов term-word-vte (жест/тип теряли вызовы) после batch-1 greens — curing reboot ВМ (boot 21:27:35), после которого все четыре кейса зелёные с первой попытки на одной сессии; документировано рядами repin-mechanics.
- Мой промежуточный rsync затер ряды batch-1 в зеркальном журнале (прецедент 10-06; восстановлены дословно из сессионного вывода), и мой batch-2 поллинг читал сталые exit-файлы — три кейса ушли параллельно и разошлись за имя демона (честные aborted-ряды гонки затёрты следующим rsync; задокументировано в repin-mechanics).
- clip-frozen wezterm-ряда (21:31:38) нёс 16 байт 0x6e в префиксе pty-потока при пройденном contains-оракуле — задокументировано в ряде и mechanics-строке; вердикт ряда не опирается на чистоту префикса.
- Хост-демон владельца не тронут (юнит-тесты + код + docs + стенд-ВМ только).

## User Setup Required

None — no external service configuration required. (Host-ревалидация причин A/B на живом столе владельца — UAT missing #5, план 10-09.)

## Next Phase Readiness

- Готово для 10-09: причины A (10-07) и B (этот план) закрыты в коде и доказаны юнит-корпусами + живыми рядами; host-ревалидация и owner-приёмка критерия 1 — оставшийся UAT-хвост.
- Открытые хвосты для верификатора: D5 (flagged assumption TERM-02, human_judgment — при том что семантический вердикт по амендменту владелец УЖЕ вынес); префикс-мусор wezterm clip-ряда (наблюдение задокументировано, кандидат в разбор 10-09/фазовой верификации); встроенная форма terminal-silent-repin (blocked-ряд 21:32:33 — до-существующий артефакт общего журнала, standalone форма зелёная).
- Блокеров нет; TERM-02 остаётся Complete (shared-ID gate открыт 10-06).

---
*Phase: 10-term-slovo*
*Completed: 2026-10-13*

## Self-Check: PASSED

SUMMARY exists; все 10 коммитов от базы плана 13e0a70 верифицированы в git log (8 production/docs + 2 state-записи контрольной точки; ledger gsd-plan-head-before-10-08 на диске, rev-list count = 10); файлы созданы (case_term_word_vte_flip.go, 2 evidence-JSON, SUMMARY) и модифицированы (actor.go, actor_test.go, wezterm.go, case_term_word_vte.go, main.go, mise.toml, ADR-008, spike-log, 10-08-PLAN.md) — все FOUND; TDD-гейт: test(10-08) 5a47555/b53b44e + feat(10-08) bfdeaf1/442f69d/9cc620e в логе (RED_EVIDENCE_OK ×2, check tdd-red-evidence); mise run ci зелёный на HEAD (build+vet+golangci-lint 0 issues+test -race, exit 0); go.mod/go.sum нетронуты (git diff --exit-code); журнальные ряды вечера в host-журнале (21:30:49/21:31:32/21:32:54 — 4 вхождения); stub-скан чист (TODO/FIXME/placeholder — ноль); tracked-дерево чистое (только метаданные .planning); хост-демон владельца не тронут.

