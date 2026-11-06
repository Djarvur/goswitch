# Фаза 12: TERM-строка — Карта паттернов

**Составлено:** 2026-10-31
**Файлов проанализировано:** 12 (4 правки, 4 новых, 4 re-read-only потребителя)
**Аналоги найдены:** 12 / 12
**База кода:** post-Phase-10/11 main (b45ed4a); все номера строк пинены по фактическому состоянию файлов 2026-10-31

## Классификация файлов

| Файл (новый/правимый) | Роль | Поток данных | Ближайший аналог | Качество |
|------------------------|------|--------------|------------------|----------|
| `internal/session/actor.go` (правка) | service (актор/FSM) | event-driven | сам: терминальный путь Фазы 10 (`startPhraseCorrection`/`startRangeCorrection`/`executeLevel2`) | exact |
| `internal/session/*_test.go` (новые: TestGestureDict_, TestLineAnchor_, TestLineDelivery_) | test | event-driven | `internal/session/actor_test.go` (TestActor_TripleTap*) | exact |
| `test/e2e/case_term_line_probe.go` (новый) | test (e2e live-ряды) | event-driven (probe) | `test/e2e/case_term_word_wezterm.go` (wezLegProbeRow) + `case_spike_wezterm.go` (recipeChordRow) | exact |
| `test/e2e/case_term_line_wezterm.go` (новый) | test (e2e live-ряды) | event-driven | `test/e2e/case_term_word_wezterm.go` (весь) | exact |
| `test/e2e/fixtures/gs-e2e-wezterm-xim-line.lua` (новый) | config (стенд-фикстура) | — | `gs-e2e-wezterm-xim.lua` (тело) + `gs-e2e-wezterm-wayland.lua` (рецепт) | exact |
| `test/e2e/main.go` (правка) | config (диспетчер кейсов) | — | сам: реестр кейсов :402–433 | exact |
| `mise.toml` (правка) | config (задачи) | — | сам: секция e2e-term-* :175–204 | exact |
| `docs/CONFIG.md` (правка) | docs | — | сам: строка `correction.terminals` :84 + терминальная клауза | role-match |
| `internal/config/config.go` (правка-возможна, OQ4) | config | — | сам: :24–25, :87–105 | role-match (дефолт — без правок) |
| `internal/clipboard/clipboard.go` (БЕЗ правок) | utility (subprocess) | file-I/O (subprocess) | сам: `Primary` :99–120 — готов, ждёт первого потребителя | n/a (готов) |
| `internal/correct/*` (БЕЗ правок) | utility (чистая логика) | transform | `plan.go:57–85`, `runs.go:78–107`, `buffer.go:127–148` | n/a (готово) |
| `internal/engine/keys.go` (БЕЗ правок) | model (константы) | — | сам: :5–12 (маски), :58–63 (KeyHome/KeyEnd/KeyLeft/KeyRight) | n/a (готово) |

**Неприкасаемые по построению:** `executeLevel2`, `replayBurst`, `drainSettleWindow`, `deferCorrectionFlip`, `armAfterVerifyTerminal`, `BuildPlan`, `ConvertRuns`, `Buffer` — write-машина строки обязана остаться байт-прежней (зелёные ряды Фазы 10); правка = регрессионный риск без причины. Новое — только вход строки и словарь D-02.

## Назначения паттернов

### `internal/session/actor.go` (service, event-driven) — правка

**Аналог:** сам файл, терминальный путь Фазы 10. Точка входа строки = существующий `startPhraseCorrection`, внутри которого появляется словарь D-02.

**Диспетчер жеста — Triple уже различается FSM, правок до словаря ноль** (:1621–1639):
```go
func (a *Actor) ExpiryAt(now time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.applySnapshot() // one config read per event (CONF-02, Pattern 2)
	a.timer = nil
	for _, action := range a.fsm.Feed(hotkey.TimerExpired{}, now) {
		slog.Info("action", "n", int(action))
		switch action {
		case hotkey.Double:
			a.startCorrection()
		case hotkey.Single:
			a.flipTo(oppositeMode(a.mode))
		case hotkey.Triple:
			a.startPhraseCorrection()
		}
	}
}
```

**Вход фразы — сюда встаёт словарь D-02** (:3893–3907):
```go
func (a *Actor) startPhraseCorrection() {
	// The ladder flush of plan 10-11 extended to the phrase gesture: …
	a.flushLevel1Confirm()
	a.startRangeCorrection(correctionRange{
		token:   a.buf.Phrase(),
		replace: a.buf.ReplacePhrase,
	})
}
```
Форма словаря: `terminal := a.classifyTerminal()` + identity-якорь (VTE-паттерн `gnome-terminal` из `a.acTerminals` → тихий отказ `triple-disabled-vte`-класса; терминал без identity → ветка строки; не-терминал → фраза байт-как сегодня). Паттерн отказ-записи — `skipCorrection` (:2113–2121):
```go
func (a *Actor) skipCorrection(reason string) {
	a.corrSkipped++
	if a.skipReasons == nil {
		a.skipReasons = make(map[string]int)
	}
	a.skipReasons[reason]++
	slog.Info("correction skipped", "reason", reason)
}
```

**Классификация терминала — единый шов, reused как есть** (:3407–3427):
```go
func (a *Actor) classifyTerminal() bool {
	if a.verifyDead {
		return true
	}
	if app, known := a.acFocusedApp(); known && a.matchACTerminal(app) {
		return true
	}
	if rs := a.role; rs != nil {
		ctx, cancel := context.WithTimeout(context.Background(), autoRoleTimeout)
		defer cancel()
		role, err := rs.Role(ctx)
		if err == nil && role == acRoleTerminal {
			return true
		}
		if err != nil {
			a.warnACRole(err)
		}
	}
	return false
}
```
Идентичность: `matchACTerminal` (:3432–3440) — substring RE2 по `a.acTerminals`; для D-02 различения VTE/wezterm тот же якорь (`config.DefaultTerminals()` = `["gnome-terminal", "wezterm"]`, `internal/config/config.go:101–105`). wezterm off-a11y-bus → identity неизвестна → строка; `gnome-terminal` матчится → молчание.

**Терминальный маршруто записи — строка едет ТЕМ ЖЕ швом, без правок** (:3964–3982, внутри `startRangeCorrection`):
```go
	a.resolvePending() // a second Double/Triple supersedes the stale round
	a.clearAfter()     // …and the previous round's verify-after with it
	// TERM-02 routing (plan 10-04, owner verdict OQ 1 «unified»): …
	terminal := a.classifyTerminal()
	if a.caps&correct.CapSurroundingText == 0 || terminal {
		a.executeLevel2(rng, converted, armed, terminal)
		return
	}
```

**WR-01 guard + cap-отказ — форма входа строки (reused)** (:4131–4155, `executeLevel2`):
```go
func (a *Actor) executeLevel2(rng correctionRange, converted []rune, armed time.Time, terminal bool) {
	if a.level2Busy {
		a.skipCorrection("correction-busy")
		a.settleCombo()
		return
	}
	a.machineArm()
	eng := a.emitter()
	plan := correct.BuildPlan(rng.token, rng.tail, converted, a.caps&^correct.CapSurroundingText, a.backspaceCap())
	if plan.Level == correct.LevelNone {
		a.machineSettle()
		a.skipCorrection("backspace-cap")
		a.settleCombo()
		return
	}
```

**Verify-then-act терминала — noRung, пост-гейт строки остаётся этим** (:2569–2580):
```go
func (a *Actor) armAfterVerifyTerminal(converted, tail []rune) {
	a.verifyEpoch++
	a.clearAfter() // never two open rounds
	a.after = &pendingAfter{
		expected: concatRunes(converted, tail),
		epoch:    a.verifyEpoch,
		noRung:   true,
	}
	epoch := a.verifyEpoch
	a.after.deadline = time.AfterFunc(a.verifyWait, func() { a.afterExpiry(epoch) })
	a.emitter().RequireSurroundingText()
}
```

**НОВЫЙ код — сел-лег: форма off-mutex ступени.** Аналог дисциплины — `runClipboardRung` (:2624–2681): снапшоты под коротким локом → subprocess/чтение БЕЗ мьютекса → короткая секция под мьютексом для вердикта; `rungMu.TryLock()` = busy-отказ, никогда блокировка:
```go
	a.mu.Lock()
	clip := a.clip
	a.mu.Unlock()
	eng := a.emitter()
	if eng == nil || clip == nil {
		return
	}
	if !a.rungMu.TryLock() {
		a.recordSkip("clipboard-rung-busy")
		return
	}
	defer a.rungMu.Unlock()
	ctx := context.Background()
	saved, had, err := clip.Save(ctx)
	// … subprocess work strictly off-mutex …
```
Чтение PRIMARY — готовый примитив, `internal/clipboard/clipboard.go:99–120`:
```go
// Primary reads the PRIMARY selection byte-exactly (wl-paste --primary
// --no-newline) — the read half of the wezterm CopyMode leg (ADR-008 D5: …).
// The call is OFF-mutex by contract — the caller reads it
// outside the actor's mutex, never under it (the runClipboardRung
// precedent, actor.go) — and bounded by the same cmdTimeout.
func (c *Clipboard) Primary(ctx context.Context) (saved []byte, had bool, err error) {
	out, err := c.call(ctx, binPaste, []string{flagPrimary, flagNoNewline}, nil)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() >= 0 && ctx.Err() == nil {
			return nil, false, nil // wl-paste's genuine empty-selection exit
		}
		return nil, false, fmt.Errorf("clipboard primary: %w", err)
	}
	return out, true, nil
}
```

**Форвард чорда выделения — парная дисциплина (амендмент (7) ADR-008)**, форма `replayBurst` :4292–4301 и `ProbeForward` (`internal/session/probe.go:179–187`):
```go
			sleep(a.level2ForwardGap)
			if emitted == 0 && a.freeAbortBeforeFirstForward(gen) {
				return true
			}
			eng.ForwardKeyEvent(engine.KeyBackSpace, backSpaceKeycode, 0)
			sleep(a.level2ForwardGap)
			eng.ForwardKeyEvent(engine.KeyBackSpace, backSpaceKeycode, engine.MaskRelease)
```
```go
	eng.ForwardKeyEvent(keyval, keycode, mods)
	slog.Debug("probe forward",
		"keyval", fmt.Sprintf("0x%x", keyval), "keycode", keycode,
		"mods", fmt.Sprintf("0x%x", mods), "release", false)
	eng.ForwardKeyEvent(keyval, keycode, mods|engine.MaskRelease)
```
Keyvals End/Home/Left/Right и MaskShift — `internal/engine/keys.go:5–12, 58–63` (`KeyHome = 0xff50`, `KeyEnd = 0xff57`, `MaskShift = 1 << 0`, `MaskRelease = 1 << 30`). Константы пейcинга: `terminalChunkKeys = 8`, `forwardEmissionsPerKey = 2` (actor.go:57–71).

**Буферный якорь SC3** — данные: `a.buf.Phrase()` (`internal/correct/buffer.go:131–133`), геометрия: `correct.BuildPlan` по буферу (N = len(буфера), НЕ len(выделения)), конвертация ТОЛЬКО буферной части через `correct.ConvertRuns` (`internal/correct/runs.go:78–107`); кап-граница — `DefaultBackspaceCap = 50` (`internal/correct/plan.go:26–30`), конфиг-потолок 500 (`internal/config/config.go:24–25`). Расхождение хвоста выделения с буфером → `skipCorrection(reason)` и return до первого форварда (D6).

### `internal/session/*_test.go` (test) — новые юнит-корпуса

**Аналог:** `internal/session/actor_test.go:726–766` — harness `wiredActor()` + `sink` + `captureLogs`, тапы через `tapShift(a)` + `a.ExpiryAt(expiryAfterWindow)`, оракулы по `sink.requireCount()/deleteCalls()/commitTexts()` и INFO-формам в логе:
```go
func TestActor_TripleTapCorrectsPhrase(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	typeWord(a, wordEN)
	pressSpace(a)
	typeWord(a, wordEN)
	honestPush(a)
	tapShift(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	if got := sink.requireCount(); got != 1 { … }
	if calls := sink.deleteCalls(); len(calls) != 0 { … }
	if texts := sink.commitTexts(); len(texts) != 0 { … }
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))
	…
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done"`) { … }
}
```
Рефьюзал-форма: `TestActor_TripleTapEmptyBuffer` (:768–772+) — «skips with the reason and makes zero calls of any kind on the sink». Матрица словаря D-02: {VTE-identity, терминал-без-identity, не-терминал} × {Single, Double, Triple} — ноль вызовов sink в VTE-Triple ячейке, Double-слово живо, GTK-Triple = фраза.

### `test/e2e/case_term_line_probe.go` (test, e2e live) — сел-проба канала чтения (Wave 0, гейт фазы)

**Аналоги:** `case_spike_wezterm.go:774–796` (recipeChordRow — механика «сентинел → чорд → settle → wl-paste --primary») и `case_term_word_wezterm.go:393–435` (wezLegProbeRow — форварднутый чорд через probe-forward):
```go
func recipeChordRow(ctx context.Context, s *stand, caseName string) error {
	if err := primarySentinel(ctx); err != nil {
		return err
	}
	if err := s.injectKeys(ctx, "shift+home"); err != nil {
		return fmt.Errorf("%s: CopyMode chord: %w", caseName, err)
	}
	if err := sleepCtx(ctx, chordSettle); err != nil {
		return err
	}
	primary := primaryReadNoNewline(ctx)
	chordVerdict := "Verdict: CopyMode-рецепт выделения не дал (sentinel не сменился)"
	if primary != "" && primary != primarySentinelMark {
		chordVerdict = "Verdict: CopyMode-рецепт выделяет (PRIMARY сменился)"
	}
	return appendSpikeRow(secRecipe, "select-chord CopyMode "+time.Now().Format("15:04:05"),
		"ydotool shift+home (рецепт CopyMode в фикстуре)",
		"sentinel="+primarySentinelMark+"; primary_after="+primary, chordVerdict)
}
```
Форварднутая ячейка (c) — форма wezLegProbeRow: `runCtl(ctx, ctlBin, "probe-forward", "shift+home")` + ожидание `probeForwardMark` записей + `chordSettle` (400 мс, `wezterm.go:198`) + `primaryReadNoNewline`; keyspec `shift+home` УЖЕ валиден грамматикой (`internal/session/probe.go:14–20, 25–39`: `ErrProbeSpecFormat = … "[mod+]{Home|End|Left|Right|BackSpace} or one key"`, `probeModifiers["shift"] = engine.MaskShift`). Вердикт-развилка датированными рядами в журнал — никогда error-code вместо ряда; красная ячейка (c) → эскалация владельцу (Rule 4), не фолбэк.

**Общая инфраструктура кейса** (все хелперы готовы, копировать вызовами, не reimplement):
- журнал: `appendSpikeRow` (`case_spike_wezterm.go:246–251`), секция = новая константа «Фаза 12 …» (контракт секций пинет `spikeJournalTest`);
- фикстура: `openWezFixture` (`case_term_word_wezterm.go:110–133`) — spawn `--config-file` TOP-LEVEL, `waitPaneReady` (`wezterm.go:234`), focus-gate `waitFocusInFrom`;
- арминг: `wezArmGate` (:188–205) — key-trace лестница ×3 по `keyTraceMark = "msg":"key"` (`wezterm.go:62`);
- клип-хелперы: `primarySentinel`/`primaryReadNoNewline` (`wezterm.go:565, 578`), `clipCopy`/`clipReadNoNewline` (`case_term_word_vte.go:367, 378`), `primarySentinelMark = "GS-SPIKE-SENTINEL"` (`wezterm.go:202`);
- корпус: `corpusLatin = "ghbdtn"`, `corpusHexRU` (`wezterm.go:56–58`), промпт-класс — префикс `echo ` (`termMidPrefix`, `case_term_word_vte.go:42`);
- venue: `venueMarker(ctx)` (`case_term_word_vte.go:85`); teardown: `killTermPane`/`capturePane` (:118, :149).

### `test/e2e/case_term_line_wezterm.go` (test, e2e live) — живые ряды строки

**Аналог:** `case_term_word_wezterm.go` целиком (1–508). Структурный скелет ряда (:318–341):
```go
func (s *stand) wezWordRow(
	ctx context.Context, caseName string, wf *wezFixtureRound, rowTag, text, wantHex string,
) error {
	offset := captureSize(wf.ptyPath)
	learned, err := s.wezWordRound(ctx, caseName, text)
	if err != nil {
		return err
	}
	hexr, cerr := waitForCaptureDelta(ctx, wf.ptyPath, offset, termCaptureWait)
	observed := fmt.Sprintf("venue: %s; learning=%v; pty_delta_hex=%s", venueMarker(ctx), learned, hexr)
	cmd := "armGate + type " + text + " + double Shift_R (learning handshake) + pty-delta oracle"
	if cerr == nil && hexr == wantHex {
		return wezRow(rowTag, cmd, observed,
			"Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона)")
	}
	…
	if rerr := wezRow(rowTag, cmd, observed, "Verdict: FAIL (pty-дельта не байт-точна)"); rerr != nil {
		return rerr
	}
	return cerr
}
```
Для строки отличия: жест `injectKeys(ctx, "Shift_R", "Shift_R", "Shift_R")` (тройной), дельта-оракул = N×`0x7f` строго прежде байт замены (форма drain-кейса `case_term_word_drain.go`), handshейк-ожидания по `termLevel2Mark = "msg":"correction","level":2` (`case_term_word_vte.go:36`) и `correctionDoneMa` (`case_autocorrect.go:89`), метки `flip-armed`/`flip-released` (delivery-машина 10-17). Обязательные ряды по RESEARCH Pattern 5: ×2 два зелёных, промпт-класс (`echo ` + корпус), mid-line, clip-frozen (PRIMARY затёрт выделением — задокументировать пред-стейт, D5), отказ-ряды (синтетический recall: набор + Up-arrow → mismatch → refusal, ноль форвардов), cap-ряд (строка > 50 → backspace-cap), VTE-негатив (тройной = НОЛЬ emitted записей; двойной-слово зелёный). Обёртка honest-abort — форма `runTermWordWezterm` (:442–457).

### `test/e2e/fixtures/gs-e2e-wezterm-xim-line.lua` (config fixture) — новый

**Аналоги:** тело — `gs-e2e-wezterm-xim.lua:1–7` (`enable_wayland = false`, минимум, без ключей владельца); рецепт — `gs-e2e-wezterm-wayland.lua:10–18`:
```lua
table.insert(keys, {
  key = 'Home', mods = 'SHIFT',
  action = wezterm.action.Multiple{
    wezterm.action.ActivateCopyMode,
    wezterm.action.CopyMode { SetSelectionMode = 'Cell' },
    wezterm.action.CopyMode 'MoveToStartOfLineContent',
    wezterm.action.CopyMode 'Close',
  },
})
return { enable_wayland = true, keys = keys }
```
Новый файл = xim-тело (`enable_wayland = false`) + рецепт из wayland-фикстуры. Stand-owned: `~/.wezterm.lua` владельца НЕ трогается (Pitfall 7); деплой рецепта в ежедневный конфиг — отдельный чекпойнт-ритуал с backup (план C).

### `test/e2e/main.go` (config) — регистрация кейсов

**Аналог:** реестр :402–433; форма терминальных live-кейсов (:420–421):
```go
		"term-word-vte":                {fn: runTermWordVTE, noSurfacePreflight: true, watchdog: recipeWatchdog},
		"term-word-wezterm":            {fn: runTermWordWezterm, noSurfacePreflight: true, watchdog: recipeWatchdog},
```
Новые: `"term-line-probe"` и `"term-line-wezterm"` с `noSurfacePreflight: true, watchdog: recipeWatchdog`; плюс usage-строки :167–171, :355–359.

### `mise.toml` (config) — задачи

**Аналог:** :178–204, форма (:186–188):
```toml
[tasks.e2e-term-word-wezterm]
description = "e2e: 10-06 term-word-wezterm — live wezterm-nightly word rows on the VM stand (criterion 1): … rows to adr-008-spike-log.md (live GNOME session; NOT in ci)"
run = "go run ./test/e2e -case term-word-wezterm"
```
Новые: `e2e-term-line-probe`, `e2e-term-line-wezterm` — НЕ in ci (комментарий-заголовок секции :175–177 повторить).

### `docs/CONFIG.md` (docs) — документация договора строки

**Аналог:** строка `correction.terminals` :84 (табличная форма: ключ | тип | дефолт | ограничения | семантика). Рядом с терминальной клаузой добавить: End-нормализация и курсор в mid-line флоу (SC1), граница D7 (интерактивные bash/zsh внутри; wrapped/PS2 — вне, отказ по якорю), PRIMARY затирается выделением жеста (D5-оговорка), cap-граница 50.

### `internal/config/config.go` (config) — вероятно БЕЗ правок (OQ4)

**Аналог:** якорь :87–105 (`DefaultTerminals()` = `["gnome-terminal", "wezterm"]`, отсутствующий ключ = якорь) и кап-граница :24–25 (`minBackspaceCap = 1`, `maxBackspaceCap = 500`). Рекомендация RESEARCH OQ4: не добавлять ключ без необходимости — словарь D-02 решается на шве classifyTerminal + identity, zero-config. Новое значение (если решит планировщик) — только по образцу `Terminals []string \`yaml:"terminals"\`` (:139) с валидацией формы :416–425.

## Общие паттерны

### D-20 refusal-запись
**Источник:** `internal/session/actor.go:2113–2121` (`skipCorrection`), применяемая форма в `executeLevel2` (:4139 `correction-busy`, :4151 `backspace-cap`)
**Применять ко:** всем новым отказам строки (`anchor-mismatch`, `selection-empty`, `line-disabled-vte`, `line-cap`, `sel-leg-unavailable`) — INFO, slug-причина, ноль содержимого.

### Контент никогда не в логах (D-20/D-21, ASVS V16)
**Источник:** `internal/session/probe.go:180–186` (keyval hex, не текст), `internal/clipboard/clipboard.go:1–8` (package doc)
**Применять ко:** сел-легу и e2e-рядам — длины/hex-дельты/булевы/латентности; содержимое выделения и буфера только в pty-оракуле стенда (tmp-only), никогда в journal/логах демона.

### Пары press+release на каждом форварде (амендмент (7))
**Источник:** `internal/session/actor.go:4292–4301`, `internal/session/probe.go:179–187`
**Применять ко:** End-форварду и чорду shift+Home строки (оба члена чорда — пары; press-only взводит автоповтор клиента — «появилось и исчезло»).

### OFF-mutex subprocess-ступень
**Источник:** `internal/session/actor.go:2624–2681` (`runClipboardRung`), контракт `internal/clipboard/clipboard.go:104–107`
**Применять ко:** сел-легу — `clip.Primary(ctx)` строго вне `a.mu`; emitter-вызовы wire-safe off-mutex; результат/вердикт — короткая секция под мьютексом.

### Verify-then-act fail-closed (SC3/D6)
**Источник:** `internal/session/actor.go:2569–2580` (`armAfterVerifyTerminal` noRung), ветка :3964–3982
**Применять ко:** буферному якорю — любое расхождение хвоста выделения с буфером = отказ до первого разрушающего действия; никогда «частично совпало — обрежем».

### Живые ряды: appendSpikeRow + ×2 + арминг-гейт + venue
**Источник:** `test/e2e/case_spike_wezterm.go:246–251`; `case_term_word_wezterm.go:81–83, 188–205, 463–508`
**Применять ко:** обоим новым кейсам — датированный ряд на каждое наблюдение, honest-abort обёртка, два подряд зелёных раунда на вердиктный ряд, ретрай-лестница арминга = 3.

### TDD-harness актора
**Источник:** `internal/session/actor_test.go:726–766` (wiredActor/sink/tapShift/ExpiryAt + log-формы)
**Применять ко:** TestGestureDict_, TestLineAnchor_, TestLineDelivery_ — строгий red→green→refactor (директива 1), zero sink calls в отказных ячейках.

## Нет аналога

| Файл | Роль | Поток данных | Причина |
|------|------|--------------|---------|
| — | — | — | Нет: каждый файл фазы имеет точного или ролевого аналога в post-Phase-10/11 коде. САМЫЙ слабый канал — цепочка «чорд→селекция→PRIMARY» — не имеет аналога в КОДЕ, но имеет точный процессный аналог (recipeChordRow/leg-probe) и гейтится сел-пробой (Pattern 1 RESEARCH) до кода машины; планировщик закладывает вердикт-развилку на владельца при красной ячейке (c). |

## Метаданные

**Область поиска аналогов:** `internal/session`, `internal/clipboard`, `internal/correct`, `internal/engine`, `internal/hotkey`, `internal/config`, `internal/uinject`, `test/e2e`, `test/e2e/fixtures`, `mise.toml`, `docs/`
**Просмотрено файлов:** ~30 (из них прочитано дословно: 18)
**Tracked-source gate:** все названные аналоги проверены `git ls-files` — tracked, зеркал нет
**Координация:** 11-01 Task 1 (Orca Round 2) правит ТОТ ЖЕ `internal/session/actor.go` (HandleLifecycle ~:1282–1360); фаза 12 работает в диапазонах ~:1621–1639, ~:2113, ~:2569–2580, ~:3407–3440, ~:3893–4008, ~:4107–4316. Integration-ordering: Round 2 PR в main ДО фазового PR, ветка фазы ребейзится, пины строк пересчитываются.
**Дата извлечения:** 2026-10-31
