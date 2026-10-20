---
phase: 10-term-slovo
plan: 07
subsystem: terminal-classification
tags: [flap-tolerant-verifydead, default-identity-anchor, role-diagnostics, config-normalization, stand-repin, tdd, gap-closure]

requires:
  - phase: 10-term-slovo plans 01-06
    provides: терминальный класс (classifyTerminal: verifyDead/identity/role-60), чанкированный level-2 с ctxEpoch-остановкой (10-04), живые ряды term-word-vte и term-negative-witnesses (10-05/10-06), стенд goswitch-stand (/tmp/stand-env.sh, rsync-рецепт, unit -debug), UAT-разрыв G-10-1 с причинами A/B
provides:
  - Флап-стойкое стирание verifyDead: реальный FocusOut откладывает стирание на focusFlapWindow (10 мс, журнальное происхождение), FocusIn в окне отменяет, истечение/Reset/синтетика флипа стирают — мгновенные same-engine пары хост-ядра больше не убивают выученный терминальный класс (G-10-1 причина A)
  - Дефолтный identity-якорь correction.terminals [gnome-terminal, wezterm]: отсутствие ключа нормализуется в config.Load и встроен в Defaults(); явный пустой список — осознанный opt-out (D-54 жив как форма выключения); docs/CONFIG.md фиксирует обе формы, SKILL.md регенерирован
  - Громкая/липкая role-диагностика: счётчик ошибок эпизода в goswitchctl status (ac_role_errors=), WARN с окном переармивания 30 мин, здоровый ответ закрывает эпизод и обнуляет счётчик (:2741-семантика расширена)
  - Стенд re-pin: term-word-vte и term-negative-witnesses байт-точные зелёные при коде 10-07 (pty_hex/counters дословно рядов 10-05); autocorrect-terminal-silent (standalone) зелёный — канонический негатив чист
  - Три RED-свидетельства (RED_EVIDENCE_OK) — строгий TDD-цикл на каждой задаче
affects: [phase-verification (причина A закрыта в коде; причина B — за 10-08 delivery-order safety), phase-10-human-check (host-ревалидация остаётся за владельцем), 10-08/10-09 (re-pin-база)]

actuals:
  tokens: 23823        # chars/4 over the realized diff 60e8b5e..HEAD (95291 chars, 13 files)
  tasks: 3
  commits: 7           # MEASURED: git rev-list --count gsd-plan-head-before-10-07..HEAD

tech-stack:
  added: []
  patterns:
    - "Отложенное стирание через одно-оконный timer-шов (UseFlapTimer, форма level2Sleep): epoch-гард устаревшего срабатывания (форма afterExpiry), переармывание вместо плодения таймеров (T-10-07-04)"
    - "Двухформенная нормализация конфига: yaml.v3 различает nil (отсутствующий ключ → дефолт-якорь) и явный пустой список (opt-out) — различение в одной точке config.Load после декода"
    - "Счётчик+окно переармивания для липкого WARN: каждый подсчитан, строка — одна на окно (акустическая дисциплина warnAppid, расширенная временем); часовой шов UseRoleClock (форма flapAfter)"

key-files:
  created:
    - ".planning/phases/10-term-slovo/evidence-20261013-10-07-red-task1.json (RED_EVIDENCE_OK Task 1)"
    - ".planning/phases/10-term-slovo/evidence-20261013-10-07-red-task2.json (RED_EVIDENCE_OK Task 2)"
    - ".planning/phases/10-term-slovo/evidence-20261013-10-07-red-task3.json (RED_EVIDENCE_OK Task 3)"
  modified:
    - "internal/session/actor.go (focusFlapWindow + flap-механика стирания; acRoleErrors/acRoleWarnRearm/roleClock)"
    - "internal/session/actor_test.go (флап-корпус, якорный корпус, role-диагностика корпус; flapClock/fakeClock)"
    - "internal/config/config.go (DefaultTerminals, якорь в Defaults, doc Correction)"
    - "internal/config/load.go (нормализация nil→якорь после декода)"
    - "internal/config/config_test.go (двухформенный decode-корпус, Defaults-якорь)"
    - "internal/ctlsvc/ctlsvc.go (токен ac_role_errors=)"
    - "internal/ctlsvc/ctlsvc_test.go (пин нового токена, старые байт-прежние)"
    - "docs/CONFIG.md (абзац D-54 → две формы, строка таблицы terminals)"
    - "skills/goswitch-config/SKILL.md (regen, CommittedFileInSync зелёный)"
    - "docs/adr/adr-008-spike-log.md (секция re-pin 10-07: 9 рядов кейсов + standalone-негатив + механика)"

key-decisions:
  - "Флап-окно — внутренняя константа 10 мс с коммент-происхождением из журнала (пары 1–3 мс, четыре за 6 мс в 16:20:05), не конфиг-ключ: геометрия churn-дефекта ядра, не пользовательская настройка; немедленно стирают только Reset и синтетика флипа (re-mint — новое поколение контекста)"
  - "ctxEpoch-остановка чанков 10-04, HardReset реального FocusOut и потребление flipCredit не тронуты — юнит-пин FlapPairBetweenChunksStops доказывает остановку бурста флап-парой"
  - "Якорь едет И нормализацией Load (документ без ключа), И в Defaults() (no-file демон и adopted-путь) — обе машины классифицируют одинаково; доезжает существующим applySnapshot-фолдом, шов classifyTerminal не правился"
  - "Счётчик role-ошибок — счётчик ЭПИЗОДА (обнуляется здоровым ответом): восстановившийся источник читается нулём, а необслуженный день (11:00:16) виден растущим числом; WARN остаётся строкой-на-эпизод с окном 30 мин"
  - "Вердикт негатива terminal-silent несёт STANDALONE форма (свежий журнал кейса): встроенная sub-run форма считает correction-записи общего журнала и попадает на done+verify-after позитивного контроля свидетелей — до-существующий артефакт формы, задокументирован честным blocked-рядом"

patterns-established:
  - "Инертный scaffolding в RED-коммите (поля/шов/Status-поверхность без поведения) — целевые тесты падают на утверждениях, не на компиляции; легитимная форма Go-TDD для тестов, требующих нового API"
  - "TAP-перевод go test -v в evidence-JSON 1:1 механическим конвертером — продолжение прецедента 10-02/10-03/10-04"

requirements-completed: [TERM-02]  # shared-ID gate: уже Complete (открыт 10-06, последним объявившим планом) — 0/1 ready, mark-complete не требуется

coverage:
  - id: D1
    description: "Флап-стойкое стирание verifyDead (G-10-1 причина A, код): пара FocusOut+FocusIn в окне сохраняет выученный класс; истечение окна/Reset/синтетика стирают; ctxEpoch растёт на каждом FocusOut включая флап; журнальный сценарий 16:20:00.334→16:20:05.478 больше не стирает класс"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestTerminalClass_VerifyDeadSurvivesFlapPair (RED: evidence-20261013-10-07-red-task1.json, RED_EVIDENCE_OK), VerifyDeadWipesWhenWindowExpires-форма (ClearedByFocusOut re-pin), ClearedByReset, ClearedBySyntheticFocusOut, FlapWindowRearmsNotBreeds, TerminalBranch_FlapPairBetweenChunksStops; полный session-корпус зелёный -race"
        status: pass
    human_judgment: false
  - id: D2
    description: "Дефолтный identity-якорь correction.terminals: документ без ключа → [gnome-terminal, wezterm]; явно пустой список → классификация не настроена; явный список → как написано; Defaults() несёт якорь; classifyTerminal классифицирует по identity до role-запроса и детектора; docs/CONFIG.md называет обе формы и происхождение (G-10-1)"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/config/config_test.go — TestCorrection_TerminalsDecode (RED: evidence-20261013-10-07-red-task2.json, RED_EVIDENCE_OK), TestDefaults_CorrectionTerminalsAnchor; internal/session/actor_test.go — TestAutoCorrect_TerminalVetoDefaultAnchor, DefaultAnchorDoesNotInventIdentity; grep gnome-terminal+wezterm docs/CONFIG.md; skillgen CommittedFileInSync зелёный"
        status: pass
    human_judgment: false
  - id: D3
    description: "Громкая/липкая role-диагностика: каждая ошибка role-запроса растит счётчик эпизода, видимый в goswitchctl status (ac_role_errors=); WARN не спамит внутри эпизода и звучит СНОВА после тихого окна 30 мин; здоровый ответ закрывает эпизод и обнуляет счётчик; прежний текст WARN байт-прежний"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestAutoCorrect_RoleErrorsCounted, RoleWarnRearmsAfterQuietWindow, RoleHealthyClosesEpisodeAndZeroes (RED: evidence-20261013-10-07-red-task3.json, RED_EVIDENCE_OK), WarnOncePerEpisode зелёный без правок; internal/ctlsvc/ctlsvc_test.go — TestRenderStatus_AutocorrectTokens с ac_role_errors=7, старые токены целы"
        status: pass
    human_judgment: false
  - id: D4
    description: "Стенд re-pin (жёсткое ограничение разрыва): term-word-vte и term-negative-witnesses при коде 10-07 — exit 0, ряды байт-точные (pty_hex=d0bfd180d0b8d0b2d0b5d1820a и counters дословно рядов 10-05); autocorrect-terminal-silent (standalone) — fired=0, ноль correction-записей при обоих режимах доставки"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md, секция «Фаза 10 re-pin … 10-07»: word-bash 18:58:48, word-zsh 18:58:50, mid-line 18:58:52, clip-frozen 18:58:55, witness-vim 18:59:10, witness-htop 18:59:14, positive-control 18:59:16, term-negative-witnesses 18:59:21, autocorrect-terminal-silent-repin 19:07:31 (standalone pass), mechanics-ряд 18:57-19:08"
        status: pass
    human_judgment: false
  - id: D5
    description: "Flagged assumption плана: формальная классификация края TERM-02 остаётся unresolved (SPEC без ## Edge Coverage) — причина A закрыта behavior-строками Tasks 1-3 и живыми re-pin рядами; вердикт классификации — за фазовой верификацией; host-ревалидация (missing #5 UAT) — за владельцем/10-08+"
    requirement: TERM-02
    verification:
      - kind: other
        ref: ".planning/phases/10-term-slovo/10-07-PLAN.md — Flagged Assumptions: {\"requirement_id\":\"TERM-02\",\"category\":\"unclassified\",\"status\":\"unresolved\"}"
        status: unknown
    human_judgment: true
    rationale: "Flagged assumption никогда не авто-разрешается: verifier обязан классифицировать край явно, не тихий pass"

duration: 150min
completed: 2026-10-13
status: complete
---

# Phase 10 Plan 07: флап-стойкая классификация терминала, дефолтный identity-якорь, громкая role-диагностика Summary

**Причина A разрыва G-10-1 закрыта в коде: verifyDead переживает мгновенные same-engine флап-пары хост-ядра (отложенное стирание в окне 10 мс, отменяемое FocusIn; Reset/синтетика/истечение стирают как прежде, ctxEpoch-защита 10-04 байт-неизменна), классификация терминала получила дефолтный identity-якорь [gnome-terminal, wezterm] с opt-out пустым списком и обновлённым docs/CONFIG.md, недоступность a11y-роли стала машинно-видимой (счётчик ac_role_errors в status + переармивающийся WARN); жёсткое ограничение разрыва удержано — стенд-ряды term-word-vte и term-negative-witnesses байт-точные зелёные при коде 10-07, канонический негатив terminal-silent (standalone) чист.**

## Performance

- **Duration:** 150 min
- **Started:** 2026-10-13T16:36:46Z
- **Completed:** 2026-10-13T19:06:22Z
- **Tasks:** 3/3 (Task 1 tracer TDD — флап-стойкий verifyDead, гейт перезапущен end-to-end зелёным; Task 2 auto TDD — дефолтный якорь + docs; Task 3 auto TDD — role-диагностика + стенд re-pin)
- **Files modified:** 13 репозиторных (7 из frontmatter + load.go, ctlsvc_test.go, SKILL.md-regen, spike-log; 3 evidence-JSON в .planning) — measured diff 60e8b5e..HEAD: 13 files, +728/−41

## Accomplishments

- **Task 1 (tracer, TDD):** флап-стойкое стирание verifyDead — реальный FocusOut армит одно окно focusFlapWindow (10 мс), стирание исполняется только по истечении окна; FocusIn в окне отменяет; Reset и синтетический FocusOut (кредит флипа) стирают немедленно байт-как-прежде. Один таймер с переармыванием (никогда не второй), epoch-гард устаревшего срабатывания (форма afterExpiry), шов UseFlapTimer (форма level2Sleep). Журнальный сценарий хоста (learn 16:20:00.334 → wipe 16:20:05.478 при флап-бурсте) больше не теряет класс. Tracer-гейт: verify перезапущен end-to-end зелёным (auto_advance) — ⚡ Tracer verified end-to-end — expanding.
- **Task 2 (TDD):** дефолтный identity-якорь [gnome-terminal, wezterm] — нормализация nil→якорь в config.Load (yaml.v3 различает отсутствие ключа и явный `[]`), якорь в Defaults() (no-file демон классифицирует идентично), docs/CONFIG.md фиксирует обе формы и происхождение (имена дословно из вердикта UAT G-10-1), SKILL.md регенерирован (golden зелёный). Юнит: identity-матч классифицирует терминалом до всякого role-запроса и детектора; якорь молчит на невидимой identity (не изобретает класс).
- **Task 3 (TDD + стенд):** role-диагностика громкая и липкая — каждая ошибка role-запроса растит счётчик эпизода (ac_role_errors= в status, обнуляется здоровым ответом), WARN строка-на-эпизод с окном переармивания 30 мин (происхождение: эпизод 11:00:16 молчал весь день), прежний текст WARN байт-прежний (WarnOncePerEpisode зелёный без правок). Стенд re-pin: rsync-рецепт 10-01 → build → unit -debug restart → term-word-vte и term-negative-witnesses зелёные с первой попытки, ряды байт-точные; autocorrect-terminal-silent (standalone) зелёный; датированные ряды в adr-008-spike-log.md.
- Строгий TDD на всех трёх задачах: RED-свидетельства RED_EVIDENCE_OK (3 файла evidence), зелёная итерация (mise run ci: build+vet+golangci-lint 0 issues+test -race) на каждом коммите; go.mod/go.sum нетронуты (git diff --exit-code зелёный).

## Task Commits

Каждая задача — атомарные TDD-коммиты:

1. **Task 1 (tracer): флап-стойкий verifyDead** - `2d1e3f9` (test: failing flap-pair test + RED evidence), `e2c2cba` (feat: focusFlapWindow + deferred wipe + pins)
2. **Task 2: дефолтный identity-якорь** - `f8601cd` (test: anchor RED + opt-out pin), `c2849fa` (feat: Load normalization + Defaults + docs/CONFIG.md + SKILL.md regen)
3. **Task 3: role-диагностика + стенд re-pin** - `3d2b558` (test: counting/rearm/healthy RED + inert scaffolding), `eb17979` (feat: acRoleErrors + rearm window + ac_role_errors token), `8b514cc` (docs: stand re-pin rows)

**Plan metadata:** финальный docs-коммит (SUMMARY+STATE+ROADMAP+REQUIREMENTS)

_Note: refactor-коммитов нет — funcorder/gofmt-чистка вошла в соответствующие feat-коммиты до их создания (зелёная итерация гейтит каждый коммит)._

## Files Created/Modified

- `internal/session/actor.go` — focusFlapWindow (10 мс, журнальное происхождение), flap-механика отложенного стирания (armVerifyDeadWipe/cancelVerifyDeadWipe/verifyDeadWipeExpiry, epoch-гард), UseFlapTimer; acRoleWarnRearm (30 мин), warnACRole со счётчиком и окном, здоровый сброс в autoConfirm, UseRoleClock, Status.AutoCorrectRoleErrors
- `internal/session/actor_test.go` — флап-корпус (survive/expiry/synthetic/rearm/ctxEpoch-пин), flapClock; якорный корпус (veto до role-запроса, не-изобретение identity), fakeRole.callCount; role-диагностика корпус (counted/rearm/healthy), fakeClock
- `internal/config/config.go` — DefaultTerminals() + константы-имена (коммент-происхождение G-10-1), якорь в Defaults(), doc Correction (две формы)
- `internal/config/load.go` — нормализация nil→якорь после декода (точка различения nil/`[]`)
- `internal/config/config_test.go` — decode-корпус пере-pinned (без ключа → якорь; `[]` → opt-out; список → как написано), Defaults-якорь
- `internal/ctlsvc/ctlsvc.go` — токен ac_role_errors= в ac-блоке (старые ключи байт-прежние)
- `internal/ctlsvc/ctlsvc_test.go` — пин нового токена
- `docs/CONFIG.md` — абзац прецедента D-54 переписан на две формы + происхождение якоря; строка таблицы terminals
- `skills/goswitch-config/SKILL.md` — регенерация (mise run skillgen-regen)
- `docs/adr/adr-008-spike-log.md` — секция «Фаза 10 re-pin … 10-07»: 9 рядов кейсов дословно + standalone-негатив + механика re-pin
- `.planning/phases/10-term-slovo/evidence-20261013-10-07-red-task{1,2,3}.json` — TDD RED-свидетельства

## Decisions Made

См. key-decisions. Ядро: флап-окно — внутренняя константа с журнальным происхождением (не конфиг-ключ); немедленно стирают только Reset и синтетика флипа (re-mint — новое поколение контекста); якорь едет и нормализацией Load, и в Defaults(); счётчик role-ошибок — счётчик эпизода (восстановление читается нулём, необслуженный день — растущим числом); вердикт негатива terminal-silent несёт standalone форма.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] RED-тесты Task 3 требовали не существующего API — инертный scaffolding в RED-коммите**
- **Found during:** Task 3, RED-фаза
- **Issue:** юниты счётчика/переармивания читают Status.AutoCorrectRoleErrors и управляют временем через часовой шов — обоих поверхностей не существовало; тесты не компилировались бы (INVALID_RED по форме)
- **Fix:** в RED-коммит вошли инертные поверхности (Status-поле, acRoleErrors/acRoleWarnedAt без инкрементов, UseRoleClock pass-through) — целевые тесты упали на УТВЕРЖДЕНИЯХ (счётчик 0 ≠ 2), не на компиляции; поведение легло в GREEN `eb17979`
- **Files modified:** internal/session/actor.go, internal/session/actor_test.go
- **Verification:** check tdd-red-evidence → RED_EVIDENCE_OK (target TestAutoCorrect_RoleErrorsCounted failed as expected)
- **Commit:** 3d2b558

**2. Задача 2, RED-посылка плана частично не воспроизвилась: «пустой список = не настроено» уже держалось**
- **Found during:** Task 2, RED-фаза (эмпирическая проба yaml.v3: отсутствующий ключ → nil, явный `[]` → non-nil empty)
- **Issue:** план ожидал падения ОБЕИХ RED-единиц; «явный пустой список декодируется в пустой non-nil список» уже было истиной — RED дал бы unexpected GREEN
- **Fix:** пустой список зафиксирован защитным пином (holds today, must survive normalization), RED-мишень — юнит «документ без ключа даёт дефолт»; различение форм легло в GREEN-нормализацию
- **Files modified:** internal/config/config_test.go
- **Verification:** RED_EVIDENCE_OK по TestCorrection_TerminalsDecode; пин opt-out зелёный до и после
- **Commit:** f8601cd

**3. [Rule 3 - Blocking] Файлы за пределами files_modified: load.go, ctlsvc_test.go, SKILL.md**
- **Found during:** Tasks 2-3
- **Issue:** план называл «нормализация в config.Load», но files_modified перечислял только internal/config/config.go — Load живёт в load.go; pin нового status-токена требует правки ctlsvc_test.go; правка таблицы docs/CONFIG.md обязывает regen skills/goswitch-config/SKILL.md (golden CommittedFileInSync, паттерн «Строгий конфиг»)
- **Fix:** нормализация в load.go (константы и Defaults — в config.go, как в плане); токен-пин в ctlsvc_test.go; `mise run skillgen-regen` исполнен
- **Files modified:** internal/config/load.go, internal/ctlsvc/ctlsvc_test.go, skills/goswitch-config/SKILL.md
- **Verification:** mise run ci зелёный (включая CommittedFileInSync)
- **Commit:** c2849fa, eb17979

---

**Total deviations:** 3 auto-fixed (1 Rule 3 — RED-scaffolding; 1 уточнение RED-посылки без категории отклонения кода — защитный пин; 1 Rule 3 — файлы вне frontmatter). **Impact on plan:** нулевое на контракт задач — все три задачи исполнены, критерии приёмки закрыты; отклонения — форма исполнения (TDD-механика и локализация файлов), не семантика.

## Issues Encountered

- Стенд, заход 1 (18:58:20): setup-отказ до строк — ssh-прогон без XAUTHORITY падает в setxkbmap-refusal (документированная механика term-word-vte-mechanics; env дополнен — это форма прогона, не ретрай ряда); заходы 2-3 зелёные с первой попытки. Честный aborted-ряд 18:58:20 сохранён в журнале.
- Встроенная sub-run terminal-silent-repin внутри term-negative-witnesses записала blocked-ряд «2 correction records»: sub-run считает `"msg":"correction"` по ОБЩЕМУ журналу кейса, куда падают done+verify-after самого позитивного контроля свидетелей (создание журнала — одно на прогон, рестарты демона аппендят). До-существующий артефакт встроенной формы (прогон 10-05 блокировался раньше — на zenity-гейте, до этой проверки), НЕ регрессия правок 10-07: каноническая STANDALONE форма (свежий журнал, форма re-pin 10-06) зелёная тут же — fired=0, ноль correction-записей при обоих режимах доставки (ряд 19:07:31). Оба исхода записаны честными рядами; вердикт негатива несёт standalone форма.
- Хост-демон владельца не тронут (юнит-тесты + код + docs + стенд-ВМ только).

## User Setup Required

None — no external service configuration required. (Host-ревалидация причины A на живом столе владельца — отдельный пункт UAT missing #5, вне этого плана.)

## Next Phase Readiness

- Готово для 10-08 (delivery-order safety, причина B): флап-стойкая классификация удерживает терминальную ветку при churn — стратегия доставки (flip не осиротяет in-flight commit) строится поверх живого класса; re-pin-ряды 10-07 — база сравнения для 10-08 (ряды term-word-vte/-wezterm/-negative-witnesses зелёные при текущем коде).
- Причина A (wezterm verify-timeout после стирания класса) закрыта в коде и доказана юнит-корпусом + байт-точным re-pin; host-ревалидация (missing #5) — за владельцем после 10-08/10-09 (E3-вердикт: доставка commit — отдельная причина B).
- Открытые хвосты для верификатора: D5 (flagged assumption TERM-02, human_judgment) и встроенная форма terminal-silent-repin (артефакт общего журнала задокументирован в коде кейса и журнале — кандидат на выделение baseline в sub-run, вне объёма фазы).
- Блокеров нет; TERM-02 остаётся Complete (shared-ID gate открыт 10-06).

---
*Phase: 10-term-slovo*
*Completed: 2026-10-13*

## Self-Check: PASSED

SUMMARY exists; все 8 коммитов от базы плана 60e8b5e верифицированы в git log (7 production/docs + метаданные ce8d8ba; ledger gsd-plan-head-before-10-07 на диске); файлы созданы (3 evidence-JSON, SUMMARY) и модифицированы (actor.go, actor_test.go, config.go, load.go, config_test.go, ctlsvc.go, ctlsvc_test.go, docs/CONFIG.md, SKILL.md, spike-log) — все FOUND; TDD-гейт: test(10-07) 2d1e3f9/f8601cd/3d2b558 + feat(10-07) e2c2cba/c2849fa/eb17979 в логе (RED_EVIDENCE_OK ×3); mise run ci зелёный на HEAD (build+vet+golangci-lint 0 issues+test -race); go.mod/go.sum нетронуты (git diff --exit-code); стенд-ряды term-word-vte/term-negative-witnesses/autocorrect-terminal-silent байт-точные зелёные при коде 10-07; tracked-дерево чистое (0 modified); stub-скан чист (TODO/FIXME/placeholder — ноль); хост-демон владельца не тронут.
