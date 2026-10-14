---
phase: 10-term-slovo
reviewed: 2026-10-14T10:29:42Z
depth: standard
files_reviewed: 5
files_reviewed_list:
  - internal/session/actor.go
  - internal/session/actor_test.go
  - internal/ctlsvc/ctlsvc_test.go
  - docs/adr/ADR-008-terminal-transport.md
  - docs/adr/adr-008-spike-log.md
findings:
  critical: 0
  warning: 1
  info: 3
  total: 4
status: issues_found
---

# Phase 10: Code Review Report (incremental re-review, round-2 delta 10-10/10-11/10-12)

**Reviewed:** 2026-10-14T10:29:42Z
**Depth:** standard
**Files Reviewed:** 5
**Status:** issues_found

## Summary

Инкрементальный ревью round-2 дельты с `ec56c123` (планы 10-10/10-11/10-12):
answer-only маршрут терминального контекста на обеих ветках вердикта
`HandleSurroundingText`, фиксы WR-01 (`level2Busy`) и WR-02 (epoch-тег
`VerifyExpiry`), миграция 16+1 nil-cache level-1 пинов на honest-push мир,
двухфазный уровень-1 (delete → доказательство удаления → commit) с бюджетом
`level1ConfirmBudget` = 60 мс, отложенный флип уровня-1 на машине 10-08,
WR-03 same-context флап-дисциплина (`flapEng`), docs-ряды 10-12. Проверочные
гейты на этом дереве зелёные: `go build ./...`, `go vet`, `go test -race`
(internal/session 10.8 s, internal/ctlsvc), `golangci-lint run` — 0 issues.

**Верификация folded-находок прошлого раунда (все три подтверждены
закрытыми):**

- **WR-01 (закрыт).** `level2Busy` ставится до BuildPlan и первого
  деструктивного форварда, снимается на всех трёх выходах (backspace-cap,
  context-changed, успех; actor.go:3813-3865). Отказ второго входа идёт через
  `skipCorrection("correction-busy")` + `settleCombo`. Пин
  `TestCorrectionBusy_SecondEntrySkipsNotSecondBurst` воспроизводит реальную
  гонку (запаркованный sleeper, второй вход через chunk-boundary окно,
  5-секундный страж дедлока) — ровно один бурст, один коммит, один done.
- **WR-02 (закрыт).** `pendingFix.epoch` + `pendingEpoch` назначаются в
  `armVerifyDeadline` под мьютексом (actor.go:1559-1568), колбэк несёт тег,
  `VerifyExpiry` no-op при несовпадении (actor.go:1484-1486) — форма afterExpiry,
  байт-в-байт. Обе точки армирования (selection, word) переведены на шов.
  Пин `TestVerifyExpiryStaleRound_DoesNotEatNextRound` с no-op-stop двойником
  таймера честно моделирует окно «fired-and-blocked».
- **WR-03 (закрыт).** `flapEng` снимается реальным FocusOut при арминге окна,
  FocusIn отменяет стирание только при равенстве объектов движка
  (actor.go:1216), снимается вместе со слотом на cancel/expiry. Пара пинов:
  `TestFlapWindowContextIdentity_RealSwitchWipesClass` (подмена эмиттера между
  половинами пары — класс стёрт, новый контекст учится с нуля) и
  `..._SameEnginePairSurvives` (10-07 не регрессировал).
- **WR-04** — вне кодовой дельты (harness), статус известен; round-2 ряд
  журнала спайка подтверждает («xrandr-проба нуджа мертва — факт харнесса»).

**Замороженные инварианты round-2 — все удержаны:**

- `go.mod`/`go.sum` нетронуты (git diff с базы пуст — проверено).
- flip_after_correction=false: `deferCorrectionFlip`/`settleCorrectionFlip`
  выходят до всякого арминга — ops-пути байт-прежние (мигрированные пины
  зелёные без правок семантики).
- D-55 вето живёт в единственном шве `classifyTerminal`, оба потребителя
  (arming-veto actor.go:3072, write-branch actor.go:3684) читают один шов;
  `TestReplayHost0939_RouteLearnsVetoSeam` прибивает сопряжённость.
- Fail-closed мир клиентов с кэш-свидетельством сохранён на обеих ветках:
  selection-раунды (range-anchored) и autocorrect-boundary-раунды исключены из
  `pendingStaleCache`; честно-пушащий GTK держит байт-прежний verify-mismatch
  (`TestReplayHost0939_SpontaneousPushClientNotRouted`).
- Two-phase level-1: все пять flush-сайтов (confirm-проверка, бюджет,
  вход feedKey, FocusOut/Reset, вход нового жеста) и фолбэк-бюджет сходятся в
  единственную точку `flushLevel1Confirm`; коммит едет только в захваченный
  снапшот эмиттера («удаление без замены» невоспроизводимо при произвольном
  порядке клиентской обработки — реплей `TwoPhaseLevel1_InvertingConsumer…`
  с моделью-инвертором это доказывает).
- Replay-пины честны: evidence-файлы существуют и показывают ровно целевые
  падения при зелёных защитных пинах (10-10: fail 4 — MatchAnswer/
  RejectionSignatures/VerifyExpiryStale/CorrectionBusy при ok
  SpontaneousPushClientNotRouted; 10-11: fail 5 целевых TwoPhase/Flap/Flush
  при ok SameEnginePairSurvives); тайминги журнала (1,9–2,4 мс) живут только
  в origin-комментариях, не синтезируются; пины кодируют решения и порядок
  ops, не тайминги. Коммиты журнальных рядов резолвятся (2c1ae04, 2fbfee1 —
  предки HEAD).

**Одна новая находка (warning).** Двухфазное окно уровня-1 отсрочило замену
буфера, на которую опирался CR-01 тест устарелости autocorrect-подтверждения:
`autoConfirm` может армировать свежий pending-раунд поверх незавершённой
ладдера (единственный не-flush'ащий вход в `startRangeCorrection`),
воспроизводя класс последствий WR-02 — ложное обучение verifyDead — через
новый интерлив. Детали и фикс ниже (WR-05). Поле текста не страдает ни на
одной ветке; страдает классификация/диагностика.

## Narrative Findings (AI reviewer)

## Warnings

### WR-05: Двухфазное окно уровня-1 ломает посылку CR-01 у autocorrect-подтверждения — autoConfirm армирует pending-раунд поверх незавершённой ладдера, и её бюджет может ложного выучить verifyDead (новое, 10-11)

**File:** `internal/session/actor.go:3335-3379` (autoConfirm → startRangeCorrection), `3412-3418` (acConfirmRefusals — буферная посылка), `1684` (rng.replace, отсроченный в flush), `1628-1641` + `2111-2118` (окно l1 и его consume-all ветка)
**Issue:** До 10-11 `executeCorrection` исполнял delete+commit+`rng.replace`
синхронно на вердикте, поэтому к моменту любого более позднего решения буфер
уже отражал поле, и CR-01 проверка `acConfirmRefusals`
(`slices.Equal(a.buf.Token(), payload.expectToken)`) честно отвечала
«геометрия больше не та» → payload-stale. Дельта 10-11 перенесла
`rng.replace` в `flushLevel1Confirm` — теперь до 60 мс после вердикта поле и
буфер остаются в ДО-коррекционном состоянии, и confirm автокоррекции,
прилетевший в это окно, проходит все refusals и зовёт
`startRangeCorrection`. `autoConfirm` — единственный вход в
`startRangeCorrection`, который НЕ зовёт `flushLevel1Confirm` (жестовые пути
flush'ат; boundary-полезная нагрузка самого AC-раунда дальше всё равно едет
в fresh-answer ветку). Возникает интерлив:

1. «ghbdtn » → граница пробела армирует AC-payload (expectToken=ghbdtn);
2. Double-жест (или машинный `CorrectNow`) уходит в cached-match level-1:
   delete на проводе, окно l1 открыто, буфер ещё «ghbdtn »;
3. confirm AC (role RTT ≤ 25 мс) видит: generation текущ, буфер равен
   ожиданию (замена ещё не случилась) → `acFired++` → новый pending-раунд
   армируется ПОВЕРХ открытого окна l1;
4. пуш клиента попадает в l1-ветку `handleSurroundingLocked` (actor.go:2111)
   и ПОГЛОЩАЕТСЯ ею — settlement-трафик, которого ждал AC-раунд, окну достался;
5. исход A (честный клиент пушит после flush): AC-раунд умирает как
   `verify-mismatch` — отказ правильный по существу, но слаг лжив и счётчик
   verify-mismatch загрязняется;
6. исход B (клиент молчит под churn дольше бюджета): `VerifyExpiry` видит
   живой pending со своим тегом → `setVerifyDead(true)` на честном GTK
   контексте, который бюджет не пропускал — окно l1 съело его раунд.
   Это в точности класс последствий WR-02 («ложно выученный класс: вето
   автокоррекции на весь эпизод + buffer-trust level-2 + машина отложенного
   флипа»), восстановленный через новый интерлив, который фиксированная форма
   WR-02 (epoch-тег) не покрывает — тег здесь СВОЙ, раунд действительно
   истёк, но причиной истечения стало чужое окно.

Плина человека: нужна двойная тапка < 25 мс от границы слова —
неправдоподобно. Плина машины: реальна — `CorrectNow` через ctl
(программный/трей-вход) в момент RTT-окна AC-подтверждения; e2e-стенды
водят `CorrectNow` программно. Коррупции текста нет ни на одной ветке:
AC-раунд исполняется только при match, match требует реального
token+tail в поле, а это значит, что клиент отбросил и первый ладдер —
повторное исполнение геометрически идемпотентно (разобрано, см. ниже).
Страдают классификация (ложное вето D-55 на эпизод), маршрутизация и
диагностические счётчики — то, ради чего round-2 и затевался.
**Fix:** в `autoConfirm` под мьютексом, сразу после generation-проверки
(actor.go:3337-3339), отказаться от payload, пока ладдер в полёте — посылка
CR-01 в этом окне недействительна:

```go
if payload.gen != a.acGeneration {
    return // superseded mid-flight by a newer boundary — the stale verdict drops
}
if a.l1 != nil {
    a.recordACAbstain(acReasonPayloadStale) // two-phase window: the armed
    return                                  // geometry no longer certifies the field
}
```

(Альтернатива — `a.flushLevel1Confirm()` перед refusals: замена буфера
сделает проверку CR-01 честной; отказ-форма дешевле и не двигает момент
коммита.) Слаг из закрытого словаря; новый slug не нужен.

## Info

### IN-07: armLevel1Confirm снимает бюджет вытесненного окна, не завершив ладдер — инвариант держится анализом графа вызовов, а не по построению

**File:** `internal/session/actor.go:1628-1634`
**Issue:** Повторный арминг при открытом окне вызывает только `a.l1.stop()` —
старый ладдер (его delete уже на проводе) молча теряет коммит: ровно
«удаление без замены», против которого построен весь план 10-11. Сегодня
ветка недостижима: оба не-жестовых пути в `executeCorrection` невозможны при
открытом окне (l1-ветка `handleSurroundingLocked` поглощает пуши раньше
pending-ветки; `autoConfirm` армит только acBoundary-раунды, которые всегда
едут в fresh-answer ветку, мимо instant-пути), а оба жестовых пути зовут
`flushLevel1Confirm` на входе. Но сам факт, что WR-05 возник из эволюции
графа вызовов, показывает цену этой недостижимости. Фикс — одна строка рядом
с существующим кодом.
**Fix:** в `armLevel1Confirm` перед `a.l1.stop()` звать
`a.flushLevel1Confirm()` (идемпотентен по nil-check) — инвариант «окно
никогда не умирает некоммитом» станет конструктивным, а не аналитическим.

### IN-08: flapEng сравнивает interface-значения через == — безопасно, пока все реализации Emitter — указатели

**File:** `internal/session/actor.go:1216` (сравнение), `394` (поле)
**Issue:** `a.flapEng == eng` паникует в рантайме для динамического типа с
несравнимым типом (структура со слайсом/картой по значению). Все три живые
реализации — указатели (`*engine.Engine`, тестовые `*fakeSink`,
`*probeSink` — проверено grep'ом), и комментарий поля это закрепляет; корпус
под `-race` ловил бы панику мгновенно. Чисто оборонительная заметка для
будущего третьего реализации `engine.Emitter`.
**Fix:** ничего сейчас; при появлении value-типизированного эмиттера —
сравнивать через captured pointer/identity-поле, либо document+assert в
AttachEngine.

### IN-09: flush расширяет (до 60 мс) до-существующее окно устаревшего раунда поверх тихой смены движка

**File:** `internal/session/actor.go:1683-1690` (flush → deferCorrectionFlip + armAfterVerify на текущем emitter()), `2324-2326` (armAfterVerify читает текущий emitter())
**Issue:** Класс старше дельты: verify-after всегда читал текущий
`emitter()`, а paste-rung по двойному mismatch существовал до 10-11. Дельта
раздвинула окно: коммит + арминг verify-after теперь могут уехать на 60 мс
после delete, и рестарт ibus/тихая смерть клиента внутри окна оставляют
чужой контекст отвечать на Require устаревшего раунда: два mismatching
ответа при включённом (opt-in) ClipboardRung дают Ctrl+V в постороннее
приложение, отложенный флип стреляет по grace. FocusOut-путь покрыт (тот же
lifecycle делает clearAfter); открыты только тихие смерти без lifecycle.
Цепь узкая, направление фейла прежнее, поведение grace — задокументированное
(«правило режима не ждёт дальше своих release-сайтов»). Записано, чтобы
следующая ревизия машины флипа видела окно целиком.
**Fix:** ничего сейчас; кандидат на следующую ревизию — тегировать
`pendingAfter` идентичностью эмиттера на арме и глушить rung/флип при
несовпадении с ответчиком.

---

_Рассмотрено, без находки_: answer-only маршрут на match-ветке консистентен с
одобренной владельцем семантикой (ключ — КАНАЛ: cacheNil на арме; честный
первый пуш позже арминга = задокументированный T-10-10-01 остаток, доставка
уровень-2 корректна и в GTK); `pendingStaleCache` исключает selection- и
boundary-раунды (двойная защита: и по `p.sel`, и по `p.rng.acBoundary`);
BuildPlan с caps-битом всегда Level1 — LevelNone в `executeCorrection`
недостижим, нулевой план невозможен; `MatchesSuffix` с пустым match
невозможен (match = token+tail, token непуст по empty-buffer отказу);
verify-after с paste-нагрузкой не может пережить обучение класса в терминал
(класс учится только при открытом pending, а pending исключает открытый
after — clearAfter на каждом арминге); все реализации `engine.Emitter` —
указатели, интерфейсное `==` сравнимо; миграция honestPush (16 сайтов в
session + ctlsvc) механически честна — байт-семантика level-1 пинов
сохранена добавлением пуша-свидетельства, не ослаблением ассертов;
`level1Budget` в тестах — задокументированное зеркальце продовой константы
(прецедент termFlipGrace); амendez-секции ADR-008 (10-10, 10-11) датированы,
якорены только журнальными рядами, D1–D7 не переписаны, Status Accepted,
Reversibility дополнена обеими ревизиями; журнал спайка хранит честные
FAIL/aborted-ряды (word-a 09:47:22 FAIL с мусорным префиксом записан
дословно, заезд 2 зелёный); флаги `2c1ae04`/`2fbfee1` в ряды 10-12 резолвятся
предками HEAD; prior-round IN-01…IN-06 — файлы вне текущей дельты, не
перепроверялись, статус переносится как есть.

_Reviewed: 2026-10-14T10:29:42Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
