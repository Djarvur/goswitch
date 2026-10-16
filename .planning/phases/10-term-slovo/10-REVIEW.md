---
phase: 10-term-slovo
reviewed: 2026-10-16T22:15:31Z
depth: standard
files_reviewed: 15
files_reviewed_list:
  - internal/session/actor.go
  - internal/session/actor_test.go
  - internal/session/probe.go
  - internal/session/probe_test.go
  - internal/ctlsvc/ctlsvc.go
  - internal/ctlsvc/ctlsvc_test.go
  - cmd/goswitchd/main.go
  - cmd/goswitchd/main_test.go
  - cmd/goswitchctl/main.go
  - cmd/goswitchctl/main_test.go
  - test/e2e/case_term_cap_delivery.go
  - test/e2e/case_term_word_drain.go
  - mise.toml
  - docs/adr/ADR-008-terminal-transport.md
  - docs/adr/adr-008-spike-log.md
findings:
  critical: 2
  warning: 3
  info: 2
  total: 7
status: issues_found
---

# Phase 10: Code Review Report (incremental re-review, round-3+round-4 delta ab24c4c..HEAD)

**Reviewed:** 2026-10-16T22:15:31Z
**Depth:** standard
**Files Reviewed:** 15
**Status:** issues_found

## Summary

Инкрементальный ревью round-3+round-4 дельты с `ab24c4c`: сплит-вербы
probe-delete/probe-require (10-13), stand-кейсы term-cap-delivery и
term-word-drain, DST-гард девственного окна с перегенерационным
свидетельством `surrSeen`/`surrEng` и диверсией в форварды (10-16),
round-4 машина доставки — replayBurst со свободным стопом до первого
форварда, drain-settle окно 150 мс с освобождённым мьютексом, суженные
release-сайты терминального флипа, delivery-метки, debug-gated оверрайд
pacing (10-17), install/acceptance-ряды (10-18).

**Гейты на этом дереве — зелёные (прогнано ревьюером):** `go build ./...`,
`go vet` (все пять пакетов дельты), `go test -race -count=1`
(internal/session 18.2 s, internal/ctlsvc, cmd/goswitchd, cmd/goswitchctl),
`golangci-lint run` — 0 issues. `git diff ab24c4c..HEAD -- go.mod go.sum`
пуст — заморозка зависимостей удержана.

**Приоритет 1 — DST-гард (проверен на watertightness — держит).**
Единственная продакшн-точка DeleteSurroundingText подтверждена grep'ом по
`internal/`+`cmd/` (без тестов): `executeCorrection` actor.go:3919; остальные
эмиттеры — только debug-вербы probe.go (сознательный exempt, fail-closed без
-debug, ретест-гейт в журнале спайка). Гвард `!a.surrSeen || a.surrEng != eng`
(actor.go:3892) стоит ПЕРЕД эмиссией, на том же снапшоте `emitter()`, что и
эмиссия; `AttachEngine` вне мьютекса не ломает пару — сравнение с
`surrEng` ловит churn-реминт (пин
`TestDstGuard_EvidenceDiesWithGeneration` гоняет именно churn без
lifecycle-пары). Свидетельство ставится в `handleSurroundingLocked` на КАЖДЫЙ
пуш (actor.go:2233-2234) и умирает ровно в FocusOut/Reset-ветке вместе с
кэшем (actor.go:1251-1258) — synthetic FocusOut собственного флипа тоже
чистит (комментарий :1225-1231 подтверждает ветку). Уровень-1 для
материализованных контекстов байт-нетронут — пин
`TestDstGuard_MaterializedContextByteUnchanged` + защита «гвард не стреляет,
счётчик нулевой». Selection-путь DST не эмитирует с 2026-09-28 и гвард не
проходит (CommitText в девственное окно безопасен по бисекции R1b/R2 — краш
специфичен чтению surrounding-состояния).

**Приоритет 2 — reentrancy drain-settle окна: WR-01-класс не закрыт
полностью.** Вторая коррекция отказывает (`level2Busy`, пины
TestCorrectionBusy_* и ReplayWezterm1921_BusyGuardCoversDeliveryWindow),
lifecycle садится свободно с громкой меткой midflow — но окна ПОД Input
interleavings не рассмотрены: жестовые флипы и клавиши обрабатываются
посреди окна. Это обе критические находки ниже (CR-01, CR-02) — общий корень:
пока `level2Busy` поднят, машина не координирует НИ одну из собственных
эмиссий ввода/флипов, кроме второй коррекции и lifecycle.

**Приоритеты 4-5 — mid-flow rule и честность корпуса (держатся).**
Amendment (6) реализован как заявлено: свободный стоп только до первого
форварда (`replayBurst`/`freeAbortBeforeFirstForward` actor.go:4109-4136),
после — громкий `context-changed-midflow` и коммит на захваченном снапшоте
(пины ReplayWezterm1921_ContextChange* в обе стороны). Replay-корпус честен:
evidence-20261016-10-16/17-red-task1.json существуют, показывают
target_test_failed при зелёных защитных пинах; живые ряды стенда
(repround4-stand 21:39-21:48 UTC, row-default/tight/loose) и установка
(host-install-gap-round4 00:51 MSK) записаны в журнал спайка ДО/ПОСЛЕ по
протоколу round-2, честные abort/FAIL-ряды первых заездов сохранены
дословно.

**Замороженные инварианты — удержаны:** flip_after_correction=false
байт-стабилен (TestReplayWezterm1921_FlipFalseByteStable,
TestTerminalBranch_FlipFalseArmsNothing); D-55 вето живёт в единственном
шве `classifyTerminal` (arming-veto actor.go:3213, write-branch actor.go:3825);
fail-closed non-terminal диверсии — diverted round едет legacy-формой без
verify-after и без deferred flip; debug-оверрайд pacing fail-closed без
-debug (TestDebugPacing_*; мусор — отказ старта, диапазон (0,5000]).

## Narrative Findings (AI reviewer)

## Critical Issues

### CR-01: Жестовый flipTo внутри drain-settle окна осиротяет pending коммит — SetGlobalEngine до коммита воспроизводит E3-потерю («удаление без замены»)

**File:** `internal/session/actor.go:3992-3995` (busy-отказ → `settleCombo` →
`flipTo`), `3625-3639` (`modeSwitchChord` → `flipTo`), `1539-1557`
(`ExpiryAt` → Single → `flipTo`), `1598-1603` (`ToggleMode`), окно
`4147-4154`; корень — `flipTo`/`settleCombo` не знают про `level2Busy`
(`flipTo` actor.go:2886-2940)
**Issue:** Пока машина стоит в drain-settle окне (мьютекс ОСВОБОЖДЁН, до
150 мс), `flipTo` остаётся достижим из четырёх независимых входов и
исполняет `SetGlobalEngine` НЕМЕДЛЕННО — между последним форвардом и ещё НЕ
отправленным коммитом:

1. **combo в окне** (наиболее вероятный путь): `feedKey` распознаёт
   Shift+RightCtrl → `comboPending=true` → `startCorrection` →
   `startRangeCorrection` → `executeLevel2` → busy-отказ
   `skipCorrection("correction-busy")` **+ `settleCombo()`** — а
   `comboPending` уже поднят самим жестом, поэтому `settleCombo` тут же
   зовёт `flipTo(oppositeMode)` (actor.go:1846-1852). Сам механизм
   busy-отказа стреляет флипом посреди чужой доставки.
2. **mode-switch chord в окне**: `feedKey` → `modeSwitchChord` →
   `flipTo` — до всякой busy-проверки (её в этом пути нет вообще).
3. **Single-tap expiry** при `tap_window_ms` ≤ ~150 (ExpiryAt → flipTo) и
   **трей-входы** ToggleMode/SwitchMode (клик в окне маловероятен, но
   достижим).

Последствие — точный E3-класс, ради которого round-4 строился: burst
форвардов УЖЕ стёр слово, `SetGlobalEngine` перечеканкивает объекты
движка, и коммит замены уходит на захваченном СТАРОМ снапшоте — GTK
отбрасывает сигнал протухшего движка (вердикт E3, амендмент 2026-10-13:
«commit DROPPED»). Итог — «удаление без замены», единственный исход,
который амендмент 2026-10-16 (план 10-17, пункт 6) объявил недопустимым.
Утверждение амендмента (пункт 3) «SetGlobalEngine между коммитом и его
release-причиной невозможен по построению» держит только для
САМОинициированных release-сайтов (grace/verify — они сужены честно) и
после коммита; окно ДО коммита не гвардится ничем. Корпус молчит: оба
busy-пина заводят второй вход через `CorrectNow` (без флипа), конфликт
«combo во время окна» не воспроизводит ни один тест.
**Fix:** пока `a.level2Busy`, `flipTo` не должен исполняться синхронно —
армить в отложенный слот (машина `deferCorrectionFlip` формы terminal:
release по следующей клавише/подтверждающему пушу уже после коммита).
Минимальная форма — один гейт в `flipTo` (все четыре входа сходятся в
него):

```go
func (a *Actor) flipTo(target scriptMode) {
	if a.level2Busy { // delivery in flight: SetGlobalEngine orphans the
		// pending commit — the mode rule rides the post-commit slot instead
		a.deferCorrectionFlip([]rune(modeSymbolOf(target)), true)
		return
	}
	...
}
```

(целевой скрипт нужно передавать как scriptMode, не через руну — набросок
показывает шов). Альтернатива уже существующим кодом: в busy-отказе
`executeLevel2` заменить `a.settleCombo()` на арминг слота, а `flipTo`
заглушить до выхода машины. Дополнить корпус пином «combo/chord внутри
drain-окна — ноль switcher-вызовов до commit-sent».

### CR-02: Клавиша, обработанная внутри drain-settle окна, ложится на провод РАНЬШЕ коммита — инверсия порядка, которой round-3 машина не имела

**File:** `internal/session/actor.go:4147-4154` (окно освобождает мьютекс),
`3034-3124` (`feedKey` — нет ни одной проверки `level2Busy`; транзит/коммит
клавиши посреди окна), guard-поле `3992-3998`
**Issue:** Round-3 машина держала мьютекс от первого форварда до коммита —
клавиша, пришедшая в процессе доставки, ждала на мьютексе и обрабатывалась
СТРОГО после коммита: порядок провода «бурст → коммит → клавиша»
гарантировался серилизацией. Round-4 освободил мьютекс на 150 мс, но
`HandleKey`/`feedKey` об окне не знают: клавиша, набранная в окне,
обрабатывается немедленно и ложится на провод МЕЖДУ форвардами и коммитом.
Поле: «ghbdtn » → 7×0x7f → клавиша `x` → коммит «привет » = **«xпривет »**
вместо «привет x» — инверсия порядка, детерминированная на каждый
interleave. Окно существует именно потому, что клиент класса 19:21:26
сливает очередь +144 мс — а интервал между клавишами быстрого набора
~100-150 мс, то есть набор, продолженный сразу после жеста, попадает в
окно регулярно, не экстраординарно. Живой стенд ряды этого не видят:
drain-кейс после жеста НЕ набирает (оракул ждёт), т.е. межпакетный набор
не покрыт ни юнит-корпусом, ни стендом. RU-ветка хуже: демон в окне ещё и
коммитит руну (`eng.CommitText` actor.go:3105) — прямой демонский эмиссии
посреди чужой доставки.
**Fix:** тот же корень, что CR-01 — дисциплина входов при `level2Busy`.
Обработка клавиш не может молча теряться (набор владельца священен), поэтому:
(а) парковать события клавиатуры на время окна — в `HandleKey` при
`a.level2Busy` ждать завершения машины (канал close-on-exit, ожидание
ограничено `terminalDrainSettle` + один pacing-квант — bounded, никогда не
вечность), затем обрабатывать клавишу штатно (release-сайты флипа при этом
сработают уже после коммита — «клавиша набирает в режиме результата»
сохраняется); либо (б) переносить drain-settle НАЗАД, до первого форварда —
теряется смысл «дать очереди стечь перед коммитом», вариант (а) честнее.
В обоих вариантах busy-флаг остаётся семантикой «машина в полёте», а не
только «вторая коррекция».

## Warnings

### WR-06: Небезопасное чтение a.terminalDrainSettle после Unlock в drainSettleWindow — нарушение собственной дисциплины поля (race-класс WR-01)

**File:** `internal/session/actor.go:4147-4150`
**Issue:** `sleep(a.terminalDrainSettle)` вычисляет аргумент ПОСЛЕ
`a.mu.Unlock()` (строка 4149 идёт после 4148) — чтение поля актора вне
мьютекса. Документация поля (actor.go:437-441) требует «Read under a.mu by
the machine»; единственный писатель `SetTerminalDrainSettle` берёт мьютекс.
Сегодня практически доброкачественно (stand-оверырайд срабатывает в `run`
до `engine.Run` — happens-before через старт горутин), но это живой
data-race для любого будущего runtime-вызова сеттера и точный класс
нарушения, который WR-01 чинил. `-race` в существующих тестах не ловит:
ни один тест не зовёт `SetTerminalDrainSettle` параллельно с живым раундом.
**Fix:** прочитать в локальную до unlock:

```go
func (a *Actor) drainSettleWindow(gen uint64, sleep func(time.Duration)) {
	settle := a.terminalDrainSettle // read under a.mu (the field's discipline)
	a.mu.Unlock()
	sleep(settle)
	a.mu.Lock()
	...
```

### WR-07: ProbeForward без nil-emitter проверки — единственный из пяти probe-вербов паникует вместо контрактного fail-closed

**File:** `internal/session/probe.go:168-183` (ProbeForward), контраст —
`199-207`, `216-223`, `230-237`, `245-252`
**Issue:** `ProbeDeleteSurrounding`/`ProbeDelete`/`ProbeRequire`/`ProbeCommit`
открываются гардом `if eng == nil { return ErrProbeNoEngine }` — заявленный
контракт семейства («Without an emitter the probe fails closed», пины
`TestProbe*_NilEngineFailClosed`). `ProbeForward` гарда не имеет: вызов
`eng.ForwardKeyEvent` на nil-интерфейсе паникует. Достижимо с продакшн-формы:
`goswitchctl probe-forward <spec>` против -debug демона, у которого ещё не
сминтился ни один input context (`actor.ProbeEmitter()` → nil,
cmd/goswitchd/main.go:321-323) — паника гасится recover-шимом ctlsvc, но
клиент получает «internal error in ProbeForward» + ERROR-ряд со стеком
вместо тихого контрактного отказа. Несимметрия внутри одного семейства
вербов, внесённая раундом 09-02 и не выровненная 10-13/10-16.
**Fix:** тот же гард, что у остальных четырёх:

```go
func ProbeForward(eng engine.Emitter, spec string) error {
	keyval, keycode, mods, err := ParseProbeKey(spec)
	if err != nil {
		return err
	}
	if eng == nil {
		return ErrProbeNoEngine
	}
	...
```

и пин `TestProbeForward_NilEngineFailClosed` в ряд с остальными.

### WR-08: Оракул cause-honest кейса term-word-drain ищет причину по ВСЕМУ журналу — причина чужого ряда может закрыть ряд текущего прогона

**File:** `test/e2e/case_term_word_drain.go:250-262` (`cause` на строке 256),
накопление журнала — `restartDaemonWithArgs` (test/e2e/main.go:511-516)
пишет все три комбо в ОДИН s.logFile без ротации
**Issue:** Матрица гоняет три комбо (default/tight/loose) в одном кейсе, и
журнал демона накапливается: каждый ряд комбо оставляет свои
`"reason":"push"/"keystroke"` метки. `drainMarkOrder` считает `cause` через
`strings.Contains` по всему тексту — если БУДУЩАЯ регрессия выпустит
flip-released с нечестной причиной (например, возрождённый grace), ряд
текущего комбо всё равно закроется причиной чужого, ПРЕДЫДУЩЕГО ряда
(позиционная половина оракула при этом чиста — LastIndex берёт текущие
метки). today пин не маскирует (release-причины пинятся юнит-корпусом
TestTerminalBranch_FlipArmsNoGrace/FlipImmuneToVerify*), но оракул кейса —
последняя линия обороны на стенде — слеп к классу регрессии, который сам
же заявляет своей целью («release с любой другой причиной был бы геометрией
round-2 — комментарий на строке 46-50»).
**Fix:** ограничить поиск хвостом текущего раунда — срезать лог по индексу
`drainFlipArmMark` (LastIndex) перед `Contains`, или снять снапшот длины
журнала до жеста и искать cause только в добавке:

```go
armed := strings.LastIndex(log, drainFlipArmMark)
tail := log
if armed >= 0 {
	tail = log[armed:]
}
cause := strings.Contains(tail, drainFlipRelPush) || strings.Contains(tail, drainFlipRelKey)
```

## Info

### IN-10: Терминальный отложенный флип без выпуска — правило режима не применяется, пока владелец не сделает следующий жест; состояние наблюдаемо, но в меню/панели отражается СТАРЫЙ режим

**File:** `internal/session/actor.go:1919-1946` (deferCorrectionFlip terminal:
без grace), `1984-1989` (verify-исходы игнорируются), `1219-1292`
(FocusOut не release)
**Issue:** Осознанное сужение plan 10-17 («правило владельца 2026-09-28
отложено, не отменено — валидирует владелец в 10-18»): после коррекции в
терминале панель/меню показывают прежний режим, пока не придёт клавиша или
подтверждающий пуш; push-less клиент без последующего набора держит слот
неопределённо долго. Наблюдаемость заявленная и живая: `flip-armed` /
`flip-released(reason,latency_ms)` INFO-ряды (actor.go:1936, 1970-1972), кейс
drain-оракул их заказывает. Записано как explicit-reminder для re-acceptance
владельца: пункт «флип отложен до жеста» должен быть в чек-листе приёмки,
иначе «mode rule потерян» — ложная трактовка корректного поведения.
**Fix:** ничего в коде; включить оба delivery-ряда в host re-acceptance
рецепт 10-18 (ряды уже в журнале — repin-round4-stand, cause-honest=true).

### IN-11: applyPacingEnv мутирует окружение процесса e2e без восстановления на выходе — изоляция сегодня случайна (один кейс = один процесс)

**File:** `test/e2e/case_term_word_drain.go:143-161` (os.Setenv без restore),
контракт изоляции — `main.go:173-215` (pickCase: ровно один кейс на процесс)
**Issue:** Матрица оставляет последнее комбо (loose: gap=50, settle=400)
в process-env после завершения кейса. Сегодня любой не-e2e наблюдатель
этого не видит (кейс — последний юнит работы процесса), но инвариант
держится на форме запуска, а не на коде: любой будущий сценарий «несколько
кейсов в одном процессе» (batch-режим матрицы) унаследует loose-pacing во
все последующие -debug демоны молча. На error-пути середины матрицы
env остаётся от последнего УСПЕШНОГО комбо — тоже без очистки.
**Fix:** `defer` восстановления прежних значений (LookupEnv до первого
Setenv) в `runTermWordDrain`, либо os.Unsetenv обеих переменных в конце
кейса — три строки, контракт изоляции перестанет зависеть от pickCase.

## Carry-forward (статус прошлых раундов, перепроверено на этой дельте)

- **WR-01/WR-02/WR-03 (закрыты, round 2) — остаются закрытыми.** Гвард
  `level2Busy` покрывает все выходы; epoch-теги VerifyExpiry/флип-грейс/l1
  держат форму afterExpiry; `flapEng` сравнение живо. ОДНАКО: настоящие
  CR-01/CR-02 показывают, что busy-гвард покрывает второй вход коррекции,
  а не «машину в полёте» как таковую — формулировка WR-01 «the rung runs
  off the actor mutex» расширена окном, и дисциплина входов вокруг окна
  не достроена (см. CR-01/CR-02).
- **WR-04 (известный, вне кода)** — stand-harness xrandr-нудж; не в дельте,
  статус без изменений.
- **WR-05 (advisory, отложен владельцем) — окно всё ещё открыто.**
  `autoConfirm` → `startRangeCorrection` (actor.go:3520 → 3783) по-прежнему
  единственный вход без `flushLevel1Confirm`; `armLevel1Confirm`
  (actor.go:1719-1732) по-прежнему снимает бюджет вытесненного окна без
  flush. Оба зафиксированы прошлым раундом (WR-05/IN-07), решение
  отложено; дельта round-4 их не усугубила (окно l1 не расширено), но и
  не закрыла.
- **IN-07/IN-08/IN-09 прошлых раундов** — переносятся как есть (окно l1
  без конструктивного flush; interface-`==` эмиттеров — все реализации
  указатели, комментарий поля держит; verify-after читает текущий emitter()
  на устаревшем раунде).

_Рассмотрено, без находки_: единственность продакшн-DST подтверждена
grep'ом (probe-вербы — debug-gated, exempt по ретест-гейту журнала);
`cmd/goswitchctl` probe-вербы — типизированные (i u) аргументы, refusals
argv-форм, таймауты; `ctlsvc` — fail-closed nil-deps всех пяти вербов
(пины TestCtlsvc_Probe*NilDepsFailClosed), recover-шим, name-guard;
`debugPacing`/`parsePacingMs` — fail-closed без -debug, мусор = отказ
старта, (0,5000] мс; term-cap-delivery — бюджетные честные never-ряды,
базовый ask якорит арифметику (baselineOK-гейт), CLIPBOARD-сентинелы;
`drainPtyOrder` — 0x7f строго прежде и ни одного после слова, честный hex
в ряд при провале; evdev-таблицы probe.go (буквы позиционно, KEY_X=45 —
ловушка +offset закрыта пином) и macrLetterKeycodes сверены с
input-event-codes.h; `renderStatus` — corr_dst_guarded добавлен без сдвига
старых ключей, config_error остаётся последним токеном; replay-корпус
честен (RED-evidence файлы существуют с exitCode 1, живые ряды стенда и
установки датированы в журнале, честные FAIL/abort первых заездов
сохранены дословно); `mise.toml` — golangci-lint v2, e2e NOT in ci,
tidy-diff гейт на месте.

---

_Reviewed: 2026-10-16T22:15:31Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
