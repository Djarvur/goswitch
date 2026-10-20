---
phase: 10-term-slovo
plan: 13
subsystem: delivery-capability-probe
tags: [gap-closure-round3, g-10-1, debug-verbs, stand-probe, blocked-stand, gnome-shell-segv, escalation-rule4, tdd, continuation-executor]

requires:
  - phase: 10-term-slovo plans 10-10/10-11
    provides: round-2 код доставки (answer-only маршрут, двухфазный уровень-1) — фон, на котором читаются пробы
  - phase: 10-term-slovo plan 09-02
    provides: прецедент debug-верба probe-forward (fail-closed nil-dep, DEBUG-journal) — форма, зеркалируемая новыми вербами
  - phase: 10-term-slovo plan 10-01
    provides: стенд goswitch-stand (rsync-рецепт, unit -debug, vm-primary) — площадка проб
provides:
  - Два debug-верба доставки на трёх уровнях (internal/session/probe.go ProbeDeleteSurrounding/ProbeCommit, ctlsvc deps+Svc-методы, CLI probe-delete-surrounding/probe-commit, -debug-проводка в cmd/goswitchd) — fail-closed без -debug, тесты зелёные под -race (прецедент ProbeForward дословно)
  - Стенд-кейс term-cap-delivery (-case term-cap-delivery, mise e2e-term-cap-delivery, NOT in ci): три пробы × два клиента, машинный фокус-гейт, бюджетные опросы DST(0,0)+Require (15×20 мс), честный never-ряд при исчерпании, CLIPBOARD-сентинел, честный abort-ряд при отказе
  - РЕЗУЛЬТАТ ПРОБЫ — эскалация владельцу (Rule 4): эмиссия DST+Require движком демона воспроизводимо валит gnome-shell стенда (4/4 signal 11 секундой после эмиссии; контроль-заезды без эмиссии переживают прогон; на обоих ядрах -147/-146); Require-доказуемая доставка (drain-proof P3, receipt-окно P2) НЕПРОВЕРЯЕМА как минимум на стенде; шесть поведенческих рядов НЕ существуют, вердикт-таблица маршрутизации НЕ выведена — 10-14 лишён журнального входа
affects: [plan-10-14 (routing input отсутствует — СТОП до эскалации), phase-verification (TERM-02)]

actuals:
  tokens: 11330        # chars/4 over the realized diff cbe5e3a..HEAD (17 файлов, +1130/−37 ≈ 45320 chars)
  tasks: 3
  commits: 5           # MEASURED: git rev-list --count gsd-plan-head-before-10-13..HEAD (4cd97d0, 0ec87bc, 037a7d9, 592aa5a — предшественник; 5284b63 — продолжение); +1 метаданные (этот коммит)

tech-stack:
  added: []
  patterns:
    - "Debug-верб-эмиттер (прецедент probe-forward 09-02): session-функция fail-closed nil-dep + DEBUG-journal без содержимого + ctlsvc deps/Svc-метод recoverMethod-формы + CLI bare errUsage-верб + -debug-проводка — пятая реплика той же формы"
    - "Бюджетированный опрос с честным never-рядом: внутренняя константа (не конфиг), исчерпание = измеренный исход (T-10-13-03), молчание — ряд, не отказ"
    - "Честный отказ плана: краш-корреляция документируется ПАРОЙ заездов (с эмиссией — крах; без эмиссии — пережил), гипотезы снимаются экспериментом (ядро), вердикт BLOCKED без синтеза рядов"

key-files:
  created:
    - "test/e2e/case_term_cap_delivery.go (стенд-кейс term-cap-delivery: три пробы × два клиента, бюджетные опросы, сентинел CLIPBOARD, abort-ряды)"
    - ".planning/phases/10-term-slovo/10-13-SUMMARY.md (этот файл)"
  modified:
    - "internal/session/probe.go + probe_test.go (ProbeDeleteSurrounding/ProbeCommit — эмиттеры проб, fail-closed, порядок эмисов пиннут)"
    - "internal/ctlsvc/ctlsvc.go + ctlsvc_test.go (deps.ProbeDeleteSurrounding/ProbeCommit + Svc-методы org.djarvur.goswitch.Probe*)"
    - "cmd/goswitchctl/main.go + main_test.go (CLI-вербы probe-delete-surrounding <offset> <nchars>, probe-commit <text>)"
    - "cmd/goswitchd/main.go (-debug-проводка новых депсов в newCtlDeps)"
    - "mise.toml (задача e2e-term-cap-delivery)"
    - "docs/adr/adr-008-spike-log.md (секция round-3: 7 вербатим-рядов кейса + 5 диагностических рядов + BLOCKED-VERDICT)"

key-decisions:
  - "Опрос Require едет тем же вербом probe-delete-surrounding 0 0 (DST(0,0) нет-op + Require) — решение исполнителя по read_first плана (прецедент-пин порядка эмисов; отдельный poll-верб не введён)"
  - "Молчание клиента — измеренный исход (never-ряд в пределах бюджета 15×20 мс, окно ожидания пуша 5 с), не отказ кейса (592aa5a) — T-10-13-03 дословно"
  - "Пробный прогон остановлен как BLOCKED после 4/4 воспроизводимых крахов gnome-shell в точке первой пробы (Rule 3 исчерпан: 2 reboot, idle-settle, ydotool-санити, grub-эксперимент ядра, повторные заезды) — никакой синтез рядов/таблицы; эскалация владельцу по Rule 4 (форма доставки требует редизайна, архитектурное решение)"
  - "Стенд возвращён лучше найденного состояния: мёртвая с 2026-10-14 18:44 сессия поднята свежим reboot'ом, юнит goswitchd active, дефолтное ядро -147"

requirements-completed: []  # TERM-02 НЕ закрыт этим планом: shared-ID gate (10-14/10-15 держат ID), и сам план BLOCKED — требований не завершено

coverage:
  - id: D1
    description: "Debug-вербы доставки на трёх уровнях, fail-closed без -debug, порядок эмисов DST→Require пиннут, DEBUG-ряды без содержимого поля (T-10-13-01/02)"
    requirement: TERM-02
    verification:
      - kind: test
        ref: "go test ./internal/session/ ./internal/ctlsvc/ ./cmd/... -race -count=1 → ok (4cd97d0 RED → 0ec87bc GREEN, предшественник); mise run ci на HEAD 5284b63 → exit 0 (build/vet/lint/test -race)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Стенд-кейс term-cap-delivery компилируется, зарегистрирован, mise-задача существует; кейс исполняет пробы с машинным фокус-гейтом, бюджетные опросы, сентинел CLIPBOARD (T-10-13-03/D5)"
    requirement: TERM-02
    verification:
      - kind: test
        ref: "go build ./... && go vet ./test/e2e/ && grep e2e-term-cap-delivery mise.toml → pass (037a7d9, предшественник); mise run ci включает test ./test/e2e → ok"
        status: pass
    human_judgment: false
  - id: D3
    description: "Живой прогон пробы: шесть рядов P1/P2/P3 × 2 клиента + вердикт-таблица маршрутизации + drain/receipt-бюджеты в секции round-3"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "mise run e2e-term-cap-delivery на vm-primary × 6 заездов 2026-10-16 17:01-17:36 UTC: НЕ ВЫПОЛНЕНО — 4/4 краха gnome-shell signal 11 секундой после эмиссии P1 (17:03:22/17:09:38/17:23:59/17:35:45), 2 miss-набора (17:21:04/17:32:35); секция round-3 содержит 7 вербатим-рядов кейса + 5 диагностических рядов + BLOCKED-VERDICT; гейт секции (≥6 vm-primary P[123]-рядов) ЧЕСТНО НЕ насыщен; вердикт-таблица/бюджеты не существуют — эскалация владельцу (Rule 4)"
        status: failed
    human_judgment: true
    rationale: "Триггерный компонент краха (ibus-daemon→shell-панель vs mutter IM vs клиент) журналами не различим; редизайн формы доставки (Require-свободные пробы с pty-оракулом / раздельная эмиссия DST и Require / разбор краха вне кейса) — архитектурное решение владельца; VTE-first-порядок не пробовался (крах в wezterm-точке до VTE, эмиссия клиент-агностична)"
  - id: D4
    description: "Flagged assumption плана: формальная классификация края TERM-02 остаётся unresolved (наследуется; $COVERAGE: {\"items\":[{\"requirement_id\":\"TERM-02\",\"category\":\"unclassified\",\"status\":\"unresolved\",\"probe\":\"unclassified — review manually\"}],\"coverage\":{\"applicable\":1,\"resolved\":0,\"unresolved\":1}}) — классификация края за владельцем (UAT Test 2 pending)"
    requirement: TERM-02
    verification:
      - kind: other
        ref: ".planning/phases/10-term-slovo/10-13-PLAN.md — Flagged Assumptions; верификатор обязан учитывать строку, не тихий pass"
        status: unknown
    human_judgment: true
    rationale: "Flagged assumption никогда не авто-разрешается: формальная классификация края — за фазовой верификацией (и за владельцем — UAT Test 2 pending)"

duration: 48min
completed: 2026-10-16
status: blocked  # Task 3 инфраструктурно заблокирован (4/4 краха gnome-shell на эмиссии пробы) — план НЕ завершён, 10-14 не маршрутизируется; эскалация владельцу
---

# Phase 10 Plan 13: round-3 caps-проба доставки G-10-1 — вербы+кейс доставлены, живой прогон BLOCKED крахом gnome-shell на эмиссии DST+Require Summary

**Два debug-верба доставки (probe-delete-surrounding/probe-commit, fail-closed без -debug, тесты -race зелёные) и стенд-кейс term-cap-delivery доставлены; живой прогон на vm-primary ЧЕСТНО ЗАБЛОКИРОВАН: 4/4 заездов с эмиссией пробы валит gnome-shell (signal 11 секундой после DST+Require, оба ядра -147/-146; контроль-заезды без эмиссии переживают), шесть поведенческих рядов P1/P2/P3 не существуют — вердикт-таблица маршрутизации и drain/receipt-бюджеты не выведены, 10-14 лишён журнального входа; эскалация владельцу (Rule 4): форма Require-доказуемой доставки требует редизайна или разбора краха.**

## Performance

- **Duration:** 48 min (16:56–17:44 UTC) — сессия продолжения (предшественник умер ~21:39 2026-10-14, Tasks 1-2 уже в git)
- **Started:** 2026-10-16T16:56:24Z
- **Completed:** 2026-10-16T17:44Z (status: blocked)
- **Tasks:** 3 — Task 1 (предшественник, 2 коммита), Task 2 (предшественник, 2 коммита), Task 3 (продолжение: прогон исполнен, исход BLOCKED, ряды+диагностика записаны)
- **Files:** 17 changed vs база плана cbe5e3a (+1130/−37); этот план не добавил зависимостей (go.mod/go.sum нетронуты, git diff --exit-code чист)
- **Стенд:** vm-primary, ядро 6.8.0-147 (эксперимент на 6.8.0-146 одноразовой grub-загрузкой, возврат reboot'ом); сессия поднята из мёртвого состояния 2026-10-14

## Accomplishments

- **Task 1 (предшественник, `4cd97d0` RED → `0ec87bc` GREEN):** ProbeDeleteSurrounding/ProbeCommit на трёх уровнях — internal/session (fail-closed nil-dep, порядок DST→Require пиннут, DEBUG-ряды: параметры/длина без содержимого — «passwords become visible»), ctlsvc (deps + Svc-методы recoverMethod-формы, fail-closed при nil-dep), CLI (probe-delete-surrounding <offset> <nchars>, probe-commit <text>, bare errUsage-форма), -debug-проводка newCtlDeps (T-09-02-01 дословно).
- **Task 2 (предшественник, `037a7d9` + `592aa5a`):** кейс term-cap-delivery — регистрация -case + mise e2e-term-cap-delivery; три пробы × два клиента (wezterm-nightly, gnome-terminal/VTE) на свежих фикстурах с машинным ydotool-фокус-гейтом; базовый опрос-якорь; бюджет опроса 15×20 мс + окно пуша 5 с (молчание = честный never, не отказ); CLIPBOARD-сентинел clip-frozen; честный abort-ряд при отказе кейса.
- **Task 3 (продолжение) — прогон и честный отказ:** зеркало re-sync rsync-рецептом 10-08 (вербы в дереве, BUILD_OK на стенде); сессия стенда поднята из мёртвого состояния (мёртвый ibus-адрес, pid 1417; virsh reboot по прецеденту); 6 заездов mise-формы (go run ./test/e2e -case term-cap-delivery) с сессионным env; **7 честных abort-рядов кейса** перенесены в секцию round-3 вербатим с зеркала + **5 диагностических рядов продолжения**: shell-crash-correlation (4/4 краха секундой после эмиссии; контроль-заезды без эмиссии пережили), kernel-experiment (-146 — крах той же сигнатуры, гипотеза снята), typed-miss-class (WR-04 хардкод-фолбэк клика на свежих сессиях), environment-window (unattended-upgrades Oct 15; предыдущий signal-6 крах Oct 14 18:44 на старом ядре без кейс-загруза), BLOCKED-VERDICT (эскалация владельцу, Rule 4).
- **Стенд оставлен лучше найденного:** живая Wayland-сессия (вместо мёртвой с Oct 14 18:44), goswitchd unit active, дефолтное ядро -147, окно фикстур закрыто teardown'ом кейса.

## Task Commits

1. **Task 1: Debug-вербы probe-delete-surrounding и probe-commit (red → green)** — `4cd97d0` (test: RED corpus) + `0ec87bc` (feat: вербы на трёх уровнях) — *предшественник*
2. **Task 2: Стенд-кейс term-cap-delivery** — `037a7d9` (feat: кейс + mise-задача) + `592aa5a` (fix: молчание = измеренный исход, never-ряд; бюджет ответа 5 с) — *предшественник*
3. **Task 3: Живой прогон пробы + вердикт-таблица** — `5284b63` (docs: 7 вербатим-рядов кейса + 5 диагностических рядов + BLOCKED-VERDICT) — *продолжение; Task 3 НЕ завершён: исход BLOCKED*

**Plan metadata:** финальный docs-коммит (SUMMARY+STATE+ROADMAP+REQUIREMENTS)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Мёртвая сессия стенда (найденное состояние) — поднята reboot'ом**
- **Found during:** Task 3, заезд-1
- **Issue:** сессия пользователя stand мертва с 2026-10-14 18:44 (gnome-shell signal 6 в простое): ibus-адрес пользователя указывал на мёртвый pid, заезд-1 упал в preflight «dial ibus bus: connection refused»
- **Fix:** virsh reboot goswitch-stand (прецедент 10-08/10-12); свежая Wayland-сессия подняла ibus
- **Files modified:** нет (стенд-операция)
- **Verification:** gnome-shell up, ibus address валиден, заезды пошли
- **Commit:** 5284b63 (ряд session-regression)

**2. [Rule 3 - Blocking] Сессионный env прогона: XAUTHORITY живёт в XDG_RUNTIME_DIR, не в $HOME**
- **Found during:** Task 3, заезд-2
- **Issue:** recipe 10-12 экспортировал XAUTHORITY=$HOME/.mutter-Xwaylandauth.* — на этой ВМ файла в $HOME нет; заезд-2 упал setxkbmap-refusal «Authorization required» (сигнатура находки 1 механики 10-06)
- **Fix:** XAUTHORITY=$(ls /run/user/$(id -u)/.mutter-Xwaylandauth.* | head -1) (файл найден по -auth аргументу Xwayland); заезд-3 активацию прошёл
- **Files modified:** нет (исполнение)
- **Verification:** активация goswitch-en зелёная в заездах 3-8
- **Commit:** 5284b63 (ряд session-regression)

**3. [Rule 3 - Blocking → Rule 4 эскалация] Крах gnome-shell на эмиссии пробы — 4/4, БЛОКИРУЕТ план**
- **Found during:** Task 3, заезды 3/4/6/8
- **Issue:** эмиссия probe-delete-surrounding (DST(0,0)+Require) движком демона убивает gnome-shell стенда (signal 11) секундой после эмиссии — 4/4 заездов с эмиссией на обоих ядрах; 2 контроль-заезда без эмиссии (miss-набор) переживают прогон; тяжёлая нагрузка того же класса 2026-10-14 (клики/набор/probe-forward, 8 минут) — без краха
- **Fix attempts (исчерпаны):** 2× reboot, idle-settle 4 мин (пережил простой — умер на пробе), ydotool-санити, одноразовая загрузка ядра 6.8.0-146 (крах той же сигнатуры — гипотеза ядра снята), повторные заезды
- **Outcome:** STOP по плану («при инфраструктурном отказе — честный ряд, не ряды»): шесть поведенческих рядов не существуют, вердикт-таблица/бюджеты не выведены; **эскалация владельцу (Rule 4)** — форма доставки (Require-доказуемая) требует архитектурного решения: (а) Require-свободные пробы с pty-оракулом для P1/P3 (вопрос P2/receipt без Require неформулируем), (б) раздельная эмиссия DST и Require для локализации триггера, (в) разбор краха на стенде вне кейса; VTE-first не пробовался
- **Files modified:** docs/adr/adr-008-spike-log.md (ряды shell-crash-correlation, kernel-experiment, BLOCKED-VERDICT)
- **Commit:** 5284b63

**Total deviations:** 3 (2 Rule 3 — исполнение стенд-прогонов; 1 Rule 3→Rule 4 — крах стенда, эскалация). **Impact on plan:** Task 3 не завершён по существу (rows+table), план BLOCKED; Tasks 1-2 закрыты полностью.

## Issues Encountered

- **Гейт секции round-3 (≥6 vm-primary P[123]-рядов) ЧЕСТНО НЕ насыщен** — verify Task 3 FAIL: подсчёт находит 0 рядов пробы (7 abort-рядов кейса не несут P[123]+vm-primary пары). Это честный исход:fabricация рядов исключена планом и мандатом. 10-14 НЕ имеет журнального входа маршрутизации — его планирование должно быть пересмотрено после эскалации.
- Предшественник умер mid-plan (~21:39 2026-10-14) до всякого живого прогона; секция round-3 существовала с пустой таблицей по контракту секций — заполнена честными abort-рядами и диагностикой.
- Свежая сессия стенда: placement-гонка окна фикстуры против хардкод-фолбэка клика (WR-04) дала 2 miss-заезда — повторный заезд лечит; зафиксировано рядом typed-miss-class.
- Хост-мутаций нет: владелец не затронут, production-юнит хоста не тронут (стенд — ВМ; все мутации стенда обратимы и задокументированы).

## User Setup Required

**Эскалация владельцу (Rule 4) — решение по форме доставки:**
1. Признать или отвергнуть находку: эмиссия DeleteSurroundingText+RequireSurroundingText движком goswitchd валит gnome-shell стенда (4/4) — на Убунте 24.04/GNOME 46.2/IBus 1.5.29. Если это свойство стека (а не стенда), хост с той же версией стека УЯЗВИМ тем же образом — проба может объяснять и хостовые «пропадания» (мандат: «вижу исправленное слово, после чего оно удаляется» — при крахе/перезапуске shell поле владельца исчезает вместе с окном).
2. Выбрать маршрут round-4: (а) Require-свободные пробы (P1/P3 по pty-оракулу; P2/receipt отменяется или переформулируется), (б) раздельная эмиссия DST и Require (локализовать триггер), (в) разбор краха вне кейса (минимальный репро: демо-движок + gdb/valgrind на shell).
3. 10-14 не стартовать: его вход (вердикт-таблица) не существует.

## Next Phase Readiness

- Блокер плана: живые ряды P1/P2/P3 недостижимы спроектированным инструментом — см. эскалацию выше. Все прочие активы плана в коде и зелёные (вербы, кейс, mise-задача, тесты -race, mise run ci на HEAD).
- Верификатору: секция «Фаза 10 round-3 caps-проба доставки» в adr-008-spike-log.md — 12 рядов (7 кейс + 5 диагностика), вербатим с зеркала стенда; гейт секции не насыщен — по замыслу.
- STATE: статус фазы не продвигается к закрытию этим планом; TERM-02 остаётся Complete (shared-ID, открыт 10-06) — закрывающий оракул G-10-1 по-прежнему вердикт владельца (10-12 ОЖИДАНИЕ + новая эскалация round-3).

---
*Phase: 10-term-slovo*
*Completed: 2026-10-16 (status: blocked)*

## Self-Check: PASSED

SUMMARY exists; все 5 коммитов от базы плана cbe5e3a верифицированы в git log (4cd97d0/0ec87bc/037a7d9/592aa5a — предшественник, 5284b63 — продолжение; ledger gsd-plan-head-before-10-13 на диске, rev-list count = 5, 6-й — метаданные); mise-задача e2e-term-cap-delivery в mise.toml FOUND; секция «Фаза 10 round-3 caps-проба доставки» в adr-008-spike-log.md FOUND (12 рядов: 7 кейс-вербатим + 5 диагностика); mise run ci на HEAD зелёный (exit 0); go.mod/go.sum нетронуты (git diff --exit-code); гейт секции round-3 (≥6 vm-primary P[123]-рядов) ЧЕСТНО НЕ насыщен — исход BLOCKED задокументирован вместо синтеза; чужие untracked-файлы не тронуты (.gsd/, .zcode/, draft.txt, test/e2e/e2e, .planning/milestone.lock, .planning/debug/*.md, evidence-*.log, 10-PATTERNS.md — вне коммитов); стенд возвращён лучше найденного состояния (живая сессия, unit active, дефолтное ядро -147).
