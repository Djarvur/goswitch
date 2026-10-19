---
phase: 10
plan: 17
subsystem: session/terminal-delivery
tags: [round4, drain-settle, flip-off-window, replay-gate, delivery-marks, g-10-1, term-02]
requires:
  - "Ряды 19:21:26/19:36:00 (DEBUG-term-host-live-failures.md, round-2 блок + дополнения 19:28/~19:45)"
  - "Бисекция 2026-10-16 (verdict (a)/(c): DST мёртв, форварды — единственный write-примитив)"
  - "Машина отложенного флипа 10-08 (deferCorrectionFlip/releaseDeferredFlip) и executeLevel2 10-04/10-05"
provides:
  - "Машина round-4 в executeLevel2: drain-settle 150 мс перед коммитом (мьютекс освобождён, WR-01 накрывает окно), флип вне окна (без grace и verify-релизов), delivery-метки INFO"
  - "Abort/commit-дисциплину смены контекста: свободный стоп только до первого форварда; после — коммит на захваченном снапшоте за меткой context-changed-midflow"
  - "Debug-gated pacing-оверрайд (GOSWITCH_DEBUG_LEVEL2_FORWARD_GAP_MS / _TERMINAL_DRAIN_SETTLE_MS) — fail-closed без -debug"
  - "Стенд-кейс term-word-drain + mise-задача e2e-term-word-drain (NOT in ci) — pty-drain оракул под явной матрицей pacing (живой ряд — 10-18)"
  - "Реплей-корпус ReplayWezterm1921*/ReplayHost1936* — зонированная матрица на инвертирующем потребителе + RED-evidence файл"
affects:
  - "10-18 (стенд + установка + owner-приёмка): устанавливаемая сборка несёт машину round-4, оверрайд и кейс; сужение release-сайтов флипа валидирует владелец"
tech-stack:
  added: []
  patterns:
    - "bounded delivery window с журнальным происхождением константы (terminalDrainSettle 150 мс — ряд 19:21:26 + E3 push)"
    - "константы → поля актора с -debug env-оверрайдом (fail-closed, прецедент debug-вербов 10-13)"
    - "replay-gated fix на инвертирующем потребителе с машинными pacing-тиками (форма TwoPhaseLevel1_InvertingConsumer)"
key-files:
  created:
    - "test/e2e/case_term_word_drain.go"
    - ".planning/phases/10-term-slovo/evidence-20261016-10-17-red-task1.json"
  modified:
    - "internal/session/actor.go"
    - "internal/session/actor_test.go"
    - "cmd/goswitchd/main.go"
    - "cmd/goswitchd/main_test.go"
    - "test/e2e/main.go"
    - "test/e2e/spike_journal_test.go"
    - "mise.toml"
    - "docs/adr/ADR-008-terminal-transport.md"
    - "docs/adr/adr-008-spike-log.md"
key-decisions:
  - "terminalDrainSettle = 150 мс: ряд 19:21:26 — done +144 мс при пакете 6×20 мс (клиентская задержка хоста превышает весь коррекционный пакет), здоровый push E3 +16 мс — окно класса наблюдённой хостовой задержки с запасом; константа с журнальным происхождением, НЕ конфиг (план делегирует значение исполнителю)"
  - "Окно drain-settle исполняется при ОСВОБОЖДЁННОМ мьютексе: второй вход коррекции отказывает correction-busy (WR-01 накрывает всё окно — пин T-10-17-01 в реплей-корпусе), lifecycle-колбэки садятся свободно — точнее планового chunkBoundaryMoved-Gosched-минимума и прямо закрывает T-10-17-05 (неограниченное удержание мьютекса)"
  - "Свободный стоп смены контекста перенесён ДО первого форварда (per-iteration перечек в первом чанке): план требовал сохранения context-changed-стопа «до первого форварда», а на HEAD такой стоп физически недостижим для ≤8-форвардных раундов (перечек был только между чанками) — теперь семантика буквальна и запинена"
  - "Тайм-аут-ячейка реплей-матрицы водит СВОЙ раунд с уже подложенным бюджетом 30 мс (форма quiet-timeout-корпуса): тихий verify-ряд — DEBUG-уровень, общий INFO-захват его не видит"
patterns-established:
  - "Зонированная реплей-матрица: в пределах окна — поле держит слово при любой release-причине; за пределами — громкость без утверждений персистентности"
  - "Кейс-секция журнала спайка открывается датированным pending-рядом в коммите кейса (контракт TestSpikeJournalSectionsExist)"
requirements-completed:
  - "requirements: [TERM-02] (см. assumptions плана: край TERM-02 — объём слова = пробег у курсора + примыкающие руны — принят в формулировке живых проверок, формальная классификация за владельцем, UAT pending; plans 10-17/10-18 разделяют ID)"
coverage:
  human_judgment: false
  rationale: "Все verify-блоки трёх задач исполнены и зелёные на финальном HEAD: RED-гейт на неизменённом actor.go (обе подписи FAIL отдельно; RED_EVIDENCE_OK), регресс-рампа (ReplayHost0939/TwoPhaseLevel1/FlapWindowContextIdentity/Level1ConfirmFlush/TestDstGuard) зелёная с -race, полный suite зелёный, mise run ci зелёный целиком, grep-гейты (mise-задача, кейс, амендрмент, env-имена) зелёные, go.mod/go.sum заморожены. Отложенная часть — валидация сужения release-сайтов владельцем — по плану лежит в 10-18 (human_judgment там)."
actuals:
  tokens: 31812
  tasks: 3
  commits: 5
  plan_head_before: 65975fb4b2ddfac32785008369ef471e4aec2d00
duration: 54min
completed: 2026-10-16
status: complete
---

# Phase 10 Plan 17: Машина round-4 — drain-settle, флип вне окна, delivery-метки Summary

**Слово, которое видно и ОСТАЁТСЯ: терминальная ветка коммитит замену только после bounded drain-settle (150 мс, мьютекс освобождён, WR-01 накрывает окно), отложенный флип лишился grace- и verify-релизов (SetGlobalEngine между коммитом и причиной невозможен по построению), смена контекста после первого форварда больше не обрывает раунд без коммита — всё доказано реплеем дословных рядов хоста 19:21:26/19:36:00 (RED на неизменённом HEAD → GREEN) и готово к pty-drain пробе стенда и жесту владельца в 10-18.**

## Performance

| Метрика | План | Факт |
|---|---|---|
| Токены (chars/4 по диффу 127247 байт) | 75000 | ~31812 |
| Задачи | 3 | 3 |
| Коммиты | — | 5 (измерено: rev-list 65975fb..HEAD; Task 3 развёрнут в 3 коммита по плану) |
| Длительность | — | 54 мин |

## Accomplishments

- **Реплей-гейт реальных рядов (Task 1, RED, 9631b52):** ReplayWezterm1921_* + ReplayHost1936_SecondSampleKeepsWord — инвертирующий потребитель wezterm-класса (коммит синхронен, форварды с управляемой задержкой по машинным pacing-тикам, release-причины драйвимы швами), зонированная матрица «задержка форвардов × release-причина флипа × момент смены контекста». На неизменённом HEAD: 6 функций КРАСНЫЕ (обе подписи отдельно в evidence-файле, RED_EVIDENCE_OK) — поле «ghприве» при lag 60 мс (дословный класс ряда владельца «вижу привет — и он удаляется»), тишина вместо delivery-меток, grace/verify-релизы флипа внутри окна, criterion-4-стоп без коммита, WR-01 не накрывал окно доставки.
- **Машина round-4 (Task 2, GREEN, cb0d93c):** replayBurst (чанки ≤8 + свободный стоп ДО первого форварда + mid-flow метки) → drainSettleWindow (terminalDrainSettle 150 мс при освобождённом мьютексе; epoch-перечек после окна — метка, никогда не аборт) → commit на захваченном снапшоте (:3821-форма) → deferCorrectionFlip(converted, true) БЕЗ grace и БЕЗ verify-релизов. Delivery-метки INFO (chunks-sent/commit-sent/flip-armed/flip-released: причина+латентность) без содержимого (D-20/D-21); grep-формы done не переименованы. Уровень-1, legacy no-caps, selection, гвард 10-16, шов classifyTerminal, вето D-55 — байт-нетронуты.
- **Инструментарий стенда (Task 3):** debug-gated оверрайд pacing-констант (деее512 — поля актора с журнальными дефолтами; env строго под -debug, мусор — громкий отказ старта, юнит-пины fail-closed); стенд-кейс term-word-drain (3d6ad07 — RAW pty-захват, матрица default/tight/loose через оверрайд, оракулы: pty-порядок 0x7f-прежде-слова, порядок delivery-меток, честная причина flip-released, CLIPBOARD-сентинел); амендрмент ADR-008 2026-10-16 с решениями 1-6, Evidence без веб-ссылок и Reversibility-строкой (c877711).
- **Регресс-рампа:** ReplayHost0939*/TwoPhaseLevel1*/FlapWindowContextIdentity*/Level1ConfirmFlush*/TestDstGuard* зелёные; пины flip=false и комбо зелёные; полный suite с -race зелёный; mise run ci зелёный целиком; go.mod/go.sum заморожены.

## Task Commits

| Task | Name | Commit |
| ---- | ---- | ------ |
| 1 | Реплей-корпус 19:21:26/19:36:00 — КРАСНЫЙ на неизменённом HEAD (TDD RED, RED_EVIDENCE_OK) | 9631b52 |
| 2 | Машина round-4 — drain-settle + флип вне окна + delivery-метки (TDD GREEN) | cb0d93c |
| 3a | Debug-gated pacing-оверрайд — fail-closed без -debug | deee512 |
| 3b | Стенд-кейс term-word-drain + mise-задача + секция журнала | 3d6ad07 |
| 3c | Амендрмент ADR-008 (решения 1-6 + Reversibility) + секция round-4 в spike-log | c877711 |

## Files

- internal/session/actor.go — terminalDrainSettle (дефолт-константа), поля level2ForwardGap/terminalDrainSettle + сеттеры; replayBurst/freeAbortBeforeFirstForward/drainSettleWindow; deferCorrectionFlip(converted, terminal) без grace для терминала; releaseDeferredFlip(reason) + releaseDeferredFlipOnVerifyOutcome (ветвленные verify-сайты); delivery-метки
- internal/session/actor_test.go — replayWezterm (потребитель), ReplayWezterm1921_* (матрица, flip=false-пин, обе ячейки смены контекста, busy-пин), ReplayHost1936*; переписанные пины старой семантики (FlapPairBetweenChunksCommits, FocusLossAfterFirstForwardCommits, FlipArmsNoGrace, FlipImmuneToVerifyMismatch/Timeout, FlipSurvivesFocusOutGraceReleases, StaleRouteFlipStillDeferred)
- cmd/goswitchd/main.go (+main_test.go) — debugPacing/parsePacingMs, env-имена, отказ старта до всякого bus-dialing'а, INFO pacing override
- test/e2e/case_term_word_drain.go (+main.go регистрация, +spike_journal_test.go секция) — RAW-pty кейс-матрица
- mise.toml — e2e-term-word-drain (NOT in ci)
- docs/adr/ADR-008-terminal-transport.md — амендрмент 2026-10-16 (план 10-17) + Reversibility; docs/adr/adr-008-spike-log.md — секции «round-4 pty-drain (term-word-drain)» и «round-4 машина доставки»
- .planning/phases/10-term-slovo/evidence-20261016-10-17-red-task1.json — RED-свидетельство (RED_EVIDENCE_OK, 232 tests / 6 fail)

## Decisions Made

- **terminalDrainSettle = 150 мс** (план делегировал значение исполнителю с журнальным происхождением): ряд 19:21:26 — done +144 мс при пакете 6×20 мс, клиентская задержка хоста превышает весь пакет; здоровый push E3 +16 мс; окно = класс наблюдённой хостовой задержки с запасом. Комментарий константы несёт происхождение дословно.
- **Окно при освобождённом мьютексе** (точнее планового «chunkBoundaryMoved-форма»): планировщик коррекции держит мьютекс через весь раунд, поэтому ожидание с удержанием мьютекса делало бы refusal второй коррекции недостижимым до конца окна; освобождение делает WR-01 реально накрывающим окно (пин busy-ячейкой реплей-корпуса) и честно закрывает T-10-17-05.
- **Свободный стоп ДО первого форварда стал буквальным**: на HEAD context-changed-стоп существовал только между чанками (для ≤8-форвардных рядов — никогда); план требовал сохранения аборта «до первого форварда» — перечек перенесён в первый чанк и запинен отдельной ячейкой.
- **Тайм-аут-ячейка матрицы** водит собственный раунд с подложенным бюджетом 30 мс (тихий verify-ряд — DEBUG-уровень, невидимый INFO-захвату общего раунда) — форма quiet-timeout-корпуса 10-08.
- **Секция журнала кейса открыта в коммите кейса** (3d6ad07), а не в док-коммите: контракт TestSpikeJournalSectionsExist требует существования заголовка к моменту появления константы — R1-регрессия тихой потери рядов.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Свободный стоп смены контекста перенесён до первого форварда**
- **Found during:** Task 2 (трассировка достижимых путей: на HEAD chunkBoundaryMoved звался только между чанками — для ≤8-форвардных раундов abort был мёртвым кодом, а план требует живой пин «смена ДО первого форварда остаётся существующим абортом»)
- **Fix:** перечек внутри первого чанка перед каждым форвардом, пока emitted==0 (freeAbortBeforeFirstForward); после первого форварда — только громкие метки
- **Files modified:** internal/session/actor.go, internal/session/actor_test.go
- **Commit:** cb0d93c

**2. [Rule 1 - Bug] Пины старой семантики переписаны под заменённую (в границах плана)**
- **Found during:** Task 2 (полный suite: TestTerminalBranch_FlipGraceDeadlineFires / _FlipReleasedOnMismatchAndTimeout / _FocusLossBetweenChunksStops / _FlapPairBetweenChunksStops / _FlipSurvivesFocusOutGraceReleases / _StaleRouteFlipStillDeferred пиннули заменённое планом поведение — grace-релиз, verify-релизы, criterion-4-стоп)
- **Fix:** каждый переписан на новую семантику с сохранением исходного инварианта (правило режима не теряется; коммит следует за удалением; lifecycle-нейтральность), см. Task Commits cb0d93c
- **Files modified:** internal/session/actor_test.go
- **Commit:** cb0d93c

**3. [Rule 3 - Blocking] Мелкие конструктивные правки под зелёную итерацию**
- **Found during:** Tasks 1-3 (funlen/cyclop/gocognit/goconst/gocritic/contextcheck — строгий lint-конфиг)
- **Fix:** извлечены helper'ы (replayBurst, drainSettleWindow, freeAbortBeforeFirstForward, replayRowEnv, replayRunTimeoutCause, drainRoundOracles/drainRowVerdict, generationBumpHook); строковые теги и матрицы названы; тайм-аут-ячейка получает собственную проводку; err113 — статическая ошибка; финальный `mise run lint` — 0 issues
- **Files modified:** internal/session/actor.go, internal/session/actor_test.go, cmd/goswitchd/main.go, test/e2e/case_term_word_drain.go
- **Commit:** cb0d93c / deee512 (amended) / 3d6ad07

### Plan-Delta (конструкция RED-сценариев, в границах плана)

- Ячейка «смена контекста после первого форварда» построена на 20-форвардной форме (termWord20): на HEAD criterion-4-аборт физически достижим только на многочленных раундах — для 7-форвардного ряда 19:21:26abort-ветка не вызывалась вовсе. Наблюдаемый отказ RED тот же, что предписан планом (раунд обрывается без коммита при форвардах на проводе).
- Ячейка «busy в окне доставки» на HEAD паркуется на финальном pacing-gap (окна ещё нет): отказ RED — «второй вход не отказал» (correction-busy отсутствует), сообщение ячейки называет отсутствие bounded-окна честно.
- Обе дельты задокументированы в комментариях корпуса (файл evidence содержит фактические наблюдаемые).

**Всего отклонений: 3 авто-фикса + 2 plan-delta конструкции тестов; на приёмку не влияют (все acceptance-критерии задач исполнены дословно).**

## Issues Encountered

- Тихий verify-ряд (timeout) — DEBUG-уровень: общий INFO-захват реплей-ячейки его не видит; ячейка водит собственный раунд с DEBUG-захватом и подложенным бюджетом 30 мс.
- gocognit/gocyclo-потолки строгого lint-конфига потребовали декомпозиции матрицы и кейса на helper'ы (директива 2 — зелёная итерация соблюдена: lint 0 issues на каждом коммите).
- `mise run lint` в цепочке коммита 3a был проглочен пайпом (tail съел exit-код) — найден на следующем прогоне, исправлен amend'ом коммита (deee512); финальный CI целиком зелёный.

## User Setup Required

- Ничего; хост не мутирован (демон владельца v1.1.1-142 продолжает работу; машина round-4 поедет на стол владельца установкой плана 10-18).

## TDD Gate Compliance

- Task 1: RED-коммит `test(10-17)` 9631b52 на неизменённом actor.go (git diff --exit-code чист); `gsd_run check tdd-red-evidence` → **RED_EVIDENCE_OK (target_test_failed)**; evidence-файл закоммичен; обе подписи (ReplayWezterm1921, ReplayHost1936) доказаны отдельно.
- Task 2: GREEN-коммит `feat(10-17)` cb0d93c после RED_EVIDENCE_OK; реплеи зелёные по зонированной матрице, регресс-рампа зелёная.
- Task 3: коммиты deee512/3d6ad07/c877711; mise run ci зелёный целиком (build/vet/lint/test).
- REFACTOR проходил внутри Tasks 2-3 (декомпозиция под lint-потолки) без смены поведения.

## Next Phase Readiness

- **10-18 (стенд + установка + owner-приёмка):** устанавливаемая сборка несёт машину round-4 (drain-settle, флип вне окна, delivery-метки), fail-closed оверрайд и кейс term-word-drain; живой ряд кейса — pty-drain доказательство механизма; журнальные delivery-метки — хостовая наблюдаемость приёмки; сужение release-сайтов флипа и правило «режим следует за скриптом результата» (отложено-не-отменено) валидирует владелец — human_judgment 10-18.
- **Гипотеза round-2** (флип гонит очередь клиента) machine-unreachable внутри окна: SetGlobalEngine между коммитом и причиной невозможен по построению и запинен матрицей; host-подпись остаётся за реплей-гейтом и финальным жестом владельца.
- **Откат** задокументирован в Reversibility ADR-008: финальный sleep 20 мс + полные release-сайты 10-08/10-11 + criterion-4-стоп.
- G-10-1 остаётся открытым до вердикта владельца (TERM-02, UAT end-of-phase).

## Self-Check: PASSED

Все заявленные файлы существуют (case_term_word_drain.go, evidence-20261016-10-17-red-task1.json, mise-задача, амендрмент); все пять коммитов (9631b52, cb0d93c, deee512, 3d6ad07, c877711) найдены в git log; commits=5 измерено `git rev-list --count 65975fb..HEAD`; регресс-рампа и mise run ci зелёные на финальном HEAD; go.mod/go.sum нетронуты; пост-коммитная проверка удалений — пусто на всех пяти коммитах.

---

## Round-6 addendum (2026-10-19): парные форварды — press+release на каждый BackSpace

Sourcedig-вердикт (`.planning/debug/2026-10-19-wezterm-repeat-sourcedig.md`, 0c2fd6e) закрыл
корень «появилось и исчезло»: press-only форварды взводили клиентский автоповтор wezterm
(window.rs:1194-1206), стреляющий через GNOME repeat-delay 500 мс. Амендмент (7) ADR-008.

### Коммиты (round-6)

- `fed2a98` test(10-17): RED-пин парной дисциплины + evidence-20261019-10-17-red-pairs.json (RED_EVIDENCE_OK)
- `33f32c3` feat(10-17): replayBurst пары press+release; весь путь записи (терминальная ветка, legacy ADR-003, диверсия DST-гарда 10-16); корпус парный; mise run ci зелёный
- `8bd4edc` docs(10-17): амендмент (7) ADR-008 + стенд re-pin секция round-6 (4 кейса зелёные, честные aborted-ряды интерференции зонда 10-13 сохранены)
- `a6dacba` docs(10-17): установка v1.1.1-226-g8bd4edc на хост + selftype-гейт (DEBUG Closure validation round-6 + 10-UAT Current Test №6)

### Что изменилось в коде

- actor.go replayBurst: press → sleep(level2ForwardGap) → release (MaskRelease) на каждый ключ;
  межпарный пейсинг сохранён; terminalChunkKeys считает ПАРЫ; forwardEmissionsPerKey=2 —
  delivery-метки (forwards=2×keys + новый атрибут keys), pacing_ms, бюджет awaitDeliverySettle.
- Корпус: burstOps/termChunkOps парные; replayWezterm-модель игнорирует release (пара = один
  backspace в pty — оракул drain-кейса не менялся); парки спящего шва на 2N+1-й тик (окно
  drain-settle, мьютекс освобождён — mid-burst парк держал бы лок, найден DEADLOCK'ом в GREEN);
  счётчики 2N по всему реплей-корпусу; deliverymarks — forwards=2×keys.

### Аудит форвард-сайтов (итог)

- actor.go:4252 replayBurst — БЫЛ press-only (дефект) → ИСПРАВЛЕН парами (единственный нарушитель).
- actor.go:2669-2672 forwardCtrlV — уже пары (эталон).
- actor.go:2862-2865 терминальная ctrl-ветка — уже пары.
- actor.go:4459 reemitTransit (CR-02) — правомочно одиночный: переизлучает РЕАЛЬНОЕ событие
  с его собственным release-битом, не синтезирует press.
- probe.go:179-183 ProbeForward — уже пары (эталон дисциплины).

### Стенд и хост

- Стенд re-pin: term-word-vte PASS, term-word-vte-flip PASS, term-word-wezterm PASS (после
  честных aborted-заездов интерференции параллельного DST-зонда 10-13 — его панели/чурн, ничего
  не убивалось; зонд завершился вердиктом 4f940b0, независимо подтвердившим DST-ветку мёртвой),
  term-negative-witnesses PASS. pty байт-точность: пара = один 0x7f на символ.
- Хост: v1.1.1-226-g8bd4edc установлен (unit active, /proc/2096528/exe идентичен, счётчики нули,
  откат назван); selftype-гейт за владельцем: cd /home/nil/DiskD/W/Djarvur/goswitch && go run ./test/e2e -selftype — одно слово: ОСТАЛОСЬ/ИСЧЕЗЛО.

### Deviations from Plan (round-6)

- Corpus-обновление (work item 3) вошло в GREEN-коммит 33f32c3 вместе с реализацией: раздельные
  коммиты оставили бы дерево с красным корпусом на RED→GREEN границе — противоречие с
  директивой зелёной итерации; работа едина (ожидания корпуса суть контракт пар).
- Два честных DEADLOCK-фикса тест-инфраструктуры в GREEN (Rule 1): парки спящего шва
  (parkRoundInDrainWindow, BusyGuardCoversDeliveryWindow) считали тики по старой арифметике —
  парк попадал в mid-burst sleep с зажатым мьютексом; порог 2N+1.
- Стенд-интерференция зонда 10-13 (4 честных aborted/FAIL-заезда) — не дефикс: уровень-2 машина
  отработала байт-точно в каждом окне (17:58:47 done level=2 ghbdtn→привет 273 мс); после
  завершения зонда + ibus re-assert оба кейса зелёные с первого повтора.

## Self-Check round-6: PASSED

Файлы существуют (evidence-20261019-10-17-red-pairs.json; release-dir v1.1.1-226-g8bd4edc с goswitchd+goswitchctl); коммиты fed2a98/33f32c3/8bd4edc/a6dacba в git log; mise run ci зелёный на 33f32c3; пандус реплея зелёный -race; go.mod/go.sum нетронуты (tidy-diff чист); пост-коммитных удалений нет.
