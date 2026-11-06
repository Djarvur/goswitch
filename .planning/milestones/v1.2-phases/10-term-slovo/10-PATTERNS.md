# Phase 10: TERM-слово - Pattern Map

**Mapped:** 2026-10-12
**Files analyzed:** 9 (2 новых, 7 модифицируемых)
**Analogs found:** 9 / 9

Все указанные пути аналогов — git-TRACKED исходники (проверено `git ls-files` 2026-10-12); gitignored-зеркал в списке нет. Фаза — расширение зрелой кодовой базы: почти все файлы **расширяют сами себя** (лучший возможный аналог), по-настоящему новый код один — семья live e2e-кейсов `test/e2e/case_term_word*.go` (роль-match на три прецедентных кейса) и, возможно, новый файл в `internal/correct` (прецедент формы — соседние файлы пакета).

**Ключевая рамка из RESEARCH:** весь терминальный путь — ОРКЕСТРАЦИЯ поверх уже написанных механизмов (буфер, конвертер, level-2, verify-after, ProbeForward, pty-оракул). Нового транспортного кода нет; риск фазы — семантика (классификация контекста, границы пробега, безопасность середины строки), не механизмы.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/correct/stretch.go` (новый файл; альтернатива — новый метод в `buffer.go`) — конвертируемый пробег слова | model (pure domain logic) | transform | `internal/correct/buffer.go` (`activeRange`/`Token`/`Tail`, `letterCapable`) + `runs.go` (моно/микс) | exact (тот же пакет, та же чистая функция) |
| `internal/correct/stretch_test.go` или расширение `buffer_test.go` | test (golden-корпус) | — | `internal/correct/buffer_test.go` | exact |
| `internal/session/actor.go` (mod) — классификационный шов, арминг-вето, чанкированный level-2, verify-after без ранга, D-24-дискриминатор | controller (оркестратор сессии) | event-driven | сам себя: конвейер `startRangeCorrection` → `executeLevel2`, вето-точка `autoCorrectBoundary`, verify-машина `settleAfter`/`armAfterVerify` | exact (in-place extension) |
| `internal/session/actor_test.go` (mod) | test (фейковый эмиттер) | — | сам себя: `fakeSink` | exact |
| `internal/config/config.go` (mod) — ключи терминального класса (identity, бюджеты чанков) | config (схема + строгий декодер) | file-I/O | сам себя: секции `Correction`/`Autocorrect` + потолки + sentinel-ошибки + `Defaults()` | exact |
| `docs/CONFIG.md` (mod) | docs (схема конфига) | — | сам себя: табличный формат + секция «Миграция»; связка с skillgen golden | exact |
| `internal/clipboard/clipboard.go` (mod, УСЛОВНО — только если OQ 3 в объёме) — флаг `--primary` для чтения PRIMARY | utility (subprocess-клиент) | file-I/O (subprocess) | сам себя: `Save()` + блок пиннутых флагов + Runner-шов | exact |
| `test/e2e/case_term_word_vte.go`, `case_term_word_wezterm.go` (новые; возможен один файл) | test (live e2e-кейсы) | event-driven + subprocess | `case_spike_caps.go` (VTE-фикстура, pty-оракул), `case_autocorrect.go` (wezterm-фикстура, counters-оракул негатива), `case_word.go` (анатомия кейса коррекции) | role-match (новая семья, механизмы фикстур готовые) |
| `mise.toml` (mod) — задачи `e2e-term-*` (NOT in ci) | config (task runner) | — | сам себя: семья `[tasks.e2e-*]` | exact |

## Pattern Assignments

### `internal/correct` — конвертируемый пробег слова (TERM-02, дефект пунктуации «ytn?»→«нет?»)

**Analog:** `internal/correct/buffer.go` + `internal/correct/runs.go`

**Рекомендация RESEARCH (Pattern 3):** НОВЫЙ метод рядом с `Token`/`Tail`, НЕ переписывать `Token()` — семантика D-13/D-14 автокоррекционного payload'а и десятки golden'ов остаются байт-в-байт. Правка — геометрия ЧТО скармливается `ConvertRuns`, не конвертер.

**Геометрия, которую расширяет новый метод** (`buffer.go:89-109`, `154-163` — источник диапазона; пробег строится от `activeRange` в сторону хвоста и/или головы):
```go
func (b *Buffer) Token() []rune {
	start, end, ok := b.activeRange()
	if !ok {
		return nil
	}

	return slices.Clone(b.runes[start:end])
}

func (b *Buffer) Tail() []rune {
	_, end, ok := b.activeRange()
	if !ok || end == len(b.runes) {
		return nil
	}

	return slices.Clone(b.runes[end:])
}

func (b *Buffer) activeRange() (start, end int, ok bool) {
	if b.curStart < len(b.runes) {
		return b.curStart, len(b.runes), true
	}
	if b.lastLen > 0 {
		return b.lastStart, b.lastStart + b.lastLen, true
	}

	return 0, 0, false
}
```

**Предикат членства, который расширяется до «табличной позиции»** (`buffer.go:189-205` — сегодня членство токена = буква/цифра; `?` имеет табличную позицию `ENToRU['?']=','` (`internal/layouts/tables.go:44`), но её образ не буква → `?` валится в tail — корень дефекта):
```go
func letterCapable(r rune) bool {
	if unicode.IsLetter(r) {
		return true
	}
	if m, ok := layouts.ENToRU[r]; ok && unicode.IsLetter(m) {
		return true
	}
	m, ok := layouts.RUToEN[r]

	return ok && unicode.IsLetter(m)
}

func tokenCapable(r rune) bool {
	return letterCapable(r) || unicode.IsDigit(r)
}
```
Новый предикат «конвертируемого пробега» = «руна имеет табличную позицию хотя бы в одной таблице» — грубо: `maps ENToRU/RUToEN`-позиция без ограничения «образ — буква». Граница пробега останавливается на руне БЕЗ позиции (пробел — D-13 живёт), но НЕ на руне с позицией.

**Потребитель пробега — семантика, которую нельзя ломать** (`runs.go:78-107`; моно-диапазон конвертирует ЦЕЛИКОМ с пунктуацией — конвертер уже умеет нужное; микс: пунктуация едет как набрана, `runs.go:52-62` и `118-140`):
```go
	switch {
	case !hasLatin && !hasCyrillic:
		return nil, false, false // no letters — D-20 refusal
	case hasLatin && hasCyrillic:
		return invertPerChar(text)
	default:
		converted, cok := Convert(text, dirOf(script))
		if !cok {
			return nil, false, false // unmapped letter — wholesale refusal
		}

		return converted, true, true
	}
```
Инварианты (RESEARCH Pattern 3): D-13 «ghbdtn »→«привет » (пробел вне таблицы — в tail); микс-семантика spec-delta 2026-10-06 не меняется; T-03-01-03 — расширенный пробег расширяет отказ честно (golden на «ytn é»).

**Стиль golden-корпуса** (`buffer_test.go:11-23` — именованные константы + хелпер, `package correct_test`; таблица-тесты — форма `TestBuffer_TwoTokens`, `buffer_test.go:63-89`):
```go
const (
	wordEN       = "ghbdtn"
	wordRU       = "привет"
	wordENDigits = "ghbdtn2026"
	wordENComma  = "ghbdtn,"
)

func push(b *correct.Buffer, s string) {
	for _, r := range s {
		b.Push(r)
	}
}
```

---

### `internal/session/actor.go` — классификационный шов + арминг-вето (10-A, ядро)

**Analog:** сам себя. Четыре точки входа, все прочитаны с HEAD.

**(1) Арминг-вето TERM-04 — точка вставки** (`autoCorrectBoundary`, `actor.go:2272-2329`): вето добавляется ПЕРЕД `detect.Check` (строка 2300), после blocklist-сегмента, по классификационному шву; confirm-стадия (`autoConfirm`, `actor.go:2447-2513`, ролевой switch 2487-2512) остаётся вторым поясом как есть. Форма существующих вето-веток — копировать дословно:
```go
	if a.caps&correct.CapSurroundingText == 0 {
		// D-53 condition 3: without the bit the verify rung is not
		// applicable — the word is never touched.
		a.recordACAbstain(acReasonNoCaps)

		return acPayload{}, false
	}
```
Закрытый словарь воздержаний — новый slug `terminal-veto` дописывается в const-блок (`actor.go:93-105`), контракты: closed, word-free (D-20/D-21):
```go
const (
	acReasonAppBlocked = "app-blocked"
	acReasonNoCaps     = "no-caps"
	acReasonRoleAmbiguous   = "role-ambiguous"
	acReasonRoleUnknown     = "role-unknown"
	acReasonRoleForbidden   = "role-forbidden"
	acReasonRoleTimeout     = "role-timeout"
	acReasonRecheckDisabled = "recheck-disabled"
	acReasonPayloadStale    = "payload-stale"
)
```
Источник роли — живой `RoleSource` (`UseRole`, `actor.go:580`; константы `internal/appid/appid.go:73-79`: `RoleTerminal uint32 = 60 // terminal — never corrected`); для wezterm источника нет — поведенческий класс verify-dead (RESEARCH Pattern 1), снимаемый в `HandleLifecycle`.

**(2) Классификационный шов — точка маршрутизации** (`startRangeCorrection`, `actor.go:2745-2811`; caps-развилка 2779-2787 — классификация встаёт ДО неё):
```go
	a.resolvePending() // a second Double/Triple supersedes the stale round
	a.clearAfter()     // …and the previous round's verify-after with it
	if a.caps&correct.CapSurroundingText == 0 {
		// Ladder level 2 (ADR-003): no surrounding text, no verification —
		// the documented ADR-004 degradation.
		a.executeLevel2(rng, converted, armed)

		return
	}
	a.pending = &pendingFix{
		rng:       rng,
		converted: converted,
		match:     concatRunes(rng.token, rng.tail),
		armed:     armed,
	}
	if rng.acBoundary || !correct.MatchesSuffix(a.surr, a.pending.match) {
		a.pending.deadline = time.AfterFunc(a.verifyWait, a.VerifyExpiry)
		eng.RequireSurroundingText()

		return
	}
	a.executeCorrection() // cached push is the freshest report
```
Диапазон собирается в `startCorrection` (`actor.go:2629-2633`): `token: a.buf.Token(), tail: a.buf.Tail(), replace: a.buf.ReplaceToken` — терминальная ветка потребляет ТУ ЖЕ структуру `correctionRange` (`actor.go:392-404`).

**(3) Чанкированный level-2 — точка расширения** (`executeLevel2`, `actor.go:2881-2900`; сегодня один слепой бурст):
```go
func (a *Actor) executeLevel2(rng correctionRange, converted []rune, armed time.Time) {
	eng := a.emitter() // one atomic snapshot for the whole burst
	plan := correct.BuildPlan(rng.token, rng.tail, converted, a.caps, a.backspaceCap())
	if plan.Level == correct.LevelNone {
		// D-27: the Backspace series would exceed the cap and this client
		// has no DeleteSurroundingText — refuse silently, not one deletion.
		a.skipCorrection("backspace-cap")
		a.settleCombo()

		return
	}
	for range plan.Backspaces {
		eng.ForwardKeyEvent(engine.KeyBackSpace, backSpaceKeycode, 0)
	}
	eng.CommitText(engine.NewIBusText(string(plan.Commit)))
	rng.replace(converted)
	a.logCorrectionDone(plan.Level, rng.token, converted, time.Since(armed))
	a.settleCorrectionFlip(converted)
	a.settleCombo()
}
```
Дельта фазы: чанки ≤8 с перечеком фокуса между чанками. Перечек НЕ изобретает механизм — контекст/поколение проверяется на границе чанка; жизненный цикл, который гасит раунд при смене контекста, уже есть (`HandleLifecycle`, `actor.go:770-824`: FocusOut/Reset → `resolvePending()` + `clearAfter()` + `buf.HardReset()`). Free-abort до первого BackSpace структурный — все отказы (empty-buffer 2748, convert 2755, D-24 2761, no-engine 2771, backspace-cap 2884) происходят до бурста.

**(4) Verify-after без ранга клипборда** — существующая машина (`settleAfter`, `actor.go:1528-1549`; `armAfterVerify`, `actor.go:1586-1597`); обязательная дельта — терминальный вариант НЕ заполняет `paste` и НЕ проверяет ранг НИ ПРИ КАКОМ конфиге:
```go
	a.clearAfter()
	slog.Info("correction verify", "outcome", "mismatch")

	return round.paste, a.opts.ClipboardRung && a.clip != nil && eng != nil  // actor.go:1546-1548
```
Терминальная ветка копирует эпоху/WR-05-дебаунс/тихий таймаут (`afterExpiry`, `actor.go:1567-1576`), но mismatch → только `slog.Info` + стоп. Контракт Pitfall 3: юнит-тест пинит «mismatch в терминале не порождает Ctrl+V ни при каком конфиге».

**(5) D-24-дискриминатор** — обе done-ветки пишут ИДЕНТИЧНУЮ INFO-строку (`actor.go:2761-2769` vs `logCorrectionDone`, `actor.go:2907-2916`); дискриминатор `changed` добавляется в INFO (без содержимого — D-20/D-21):
```go
func (a *Actor) logCorrectionDone(level correct.Level, token, converted []rune, latency time.Duration) {
	a.corrDone++
	slog.Info("correction", "outcome", "done")
	slog.Debug("correction",
		"level", int(level),
		"runes", len(token),
		"source", string(token),
		"result", string(converted),
		"latency_ms", latency.Milliseconds())
}
```
Внимание: e2e-греп-формы (`"msg":"correction","outcome":"done"` — `case_word.go:66`) должны остаться зелёными — дискриминатор добавляется атрибутом, не заменой msg.

---

### `internal/session/actor_test.go` — юнит-корпус фейкового эмиттера

**Analog:** сам себя — `fakeSink` (`actor_test.go:112-171`), под мьютексом, единый op-лог пинит ПОРЯДОК вызовов (форма бурста — ровно то, что нужно чанк-тесту):
```go
type fakeSink struct {
	mu          sync.Mutex
	requires    int
	deletes     []deleteCall
	commits     []string
	forwards    []forwardCall
	modes       []string
	ops         []string
	forwardHook func(forwardCall)
	modeHook    func(symbol string)
}

func (f *fakeSink) ForwardKeyEvent(keyval, keycode, state uint32) {
	f.mu.Lock()
	call := forwardCall{keyval: keyval, keycode: keycode, state: state}
	f.forwards = append(f.forwards, call)
	f.ops = append(f.ops, fmt.Sprintf("forward(%d,%d,%d)", keyval, keycode, state))
	hook := f.forwardHook
	f.mu.Unlock()

	if hook != nil {
		hook(call)
	}
}
```
Готовые приёмы корпуса: детерминированный expiry через `ExpiryAt` (не реальные таймеры — `farWindow`, `actor_test.go:27-36`), `captureLogs` (guarded буфер + JSON handler, `actor_test.go:48-57`), `forwardHook` для интерлива. Тесты в `package session_test`, pubkey-API актора.

---

### `internal/config/config.go` — новые ключи терминального класса

**Analog:** сам себя. Прецедент секции со списком паттернов + потолками — `Autocorrect` (`config.go:117-123`), простой секции с бюджетами — `Correction` (`config.go:94-98`):
```go
type Correction struct {
	BackspaceCap        int  `yaml:"backspace_cap"`
	ClipboardRung       bool `yaml:"clipboard_rung"`
	FlipAfterCorrection bool `yaml:"flip_after_correction"`
}
```
Обязательные сопутствующие элементы (Pitfall 5 — схема+валидация+docs/CONFIG.md одним PR, zero-value = старое поведение, прецедент D-54):
- потолок в const-блоке (`config.go:21-31`): `maxAutocorrectBlocklist = 64` — форма для потолка списка identity; `minBackspaceCap/maxBackspaceCap = 1/500` — форма для коридора бюджета чанков;
- sentinel-ошибки (`config.go:51-66`, дисциплина err113): каждая ошибка называет поле и значение;
- дефолт в `Defaults()` (`config.go:224-260`) — zero-value совместим со старым поведением;
- identity-паттерны — RE2 с явной анкеровкой (прецедент D-53 в `Autocorrect.AppsBlocklist`); валидация диапазонов — DoS-охрана (нулевые/отрицательные бюджеты чанков — громкий отказ).

---

### `docs/CONFIG.md` — синхронизация схемы

**Analog:** сам себя. Формат ключа — строка таблицы `| Key | Type | Default | Range / vocabulary | Meaning |` (раздел «Схема», строка 12+); прецедент миграционной заметки при ломающем изменении — «Миграция (ревизия D-53, 2026-10-04)» (строки ~41-50). **Скрытая связка:** `docs/CONFIG.md` — источник региона SKILL.md; правка требует `go run ./cmd/skillgen` (mise-задача `skillgen-regen`, `mise.toml:247-249`) и обновления golden-теста `CommittedFileInSync` (D-8-10) — иначе зелёная итерация падает на реген-тесте.

---

### `internal/clipboard/clipboard.go` — флаг `--primary` (УСЛОВНО: OQ 3)

**Analog:** сам себя. Дельта минимальна: константа в блок пиннутых флагов (`clipboard.go:30-34`), чтение в форме `Save` (`clipboard.go:82-94`):
```go
const (
	flagNoNewline   = "--no-newline"
	flagTrimNewline = "--trim-newline"
	flagClear       = "--clear"
)

func (c *Clipboard) Save(ctx context.Context) (saved []byte, had bool, err error) {
	out, err := c.call(ctx, binPaste, []string{flagNoNewline}, nil)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() >= 0 && ctx.Err() == nil {
			return nil, false, nil // wl-paste's genuine empty-clipboard exit
		}

		return nil, false, fmt.Errorf("clipboard save: %w", err)
	}

	return out, true, nil
}
```
Контракты, которые сохраняются: контент ТОЛЬКО stdin/pipe, аргументы — литеральные флаги (T-03-03-01, V14); timeout 1500 мс (`cmdTimeout`, `clipboard.go:25`); Runner-шов `WithRunner` (`clipboard.go:65-69`) — юнит-корпус гоняет recording-fake по argv; пустое состояние — `ExitCode() >= 0 && ctx.Err() == nil`, не ошибка. OFF-mutex вызова (вне мьютекса актора) — обязанность вызывающей стороны (прецедент runClipboardRung, `actor.go:1641-1649`).

---

### `test/e2e/case_term_word_*.go` — новые live-кейсы (10-B: VTE; 10-C: wezterm + негативные ряды)

**Analogs:** три прецедента, по одному на каждый тип ряда.

**(a) VTE-фикстура + pty-оракул** (`case_spike_caps.go:128-136` — спавн с pty-захватом; байт-точная земля без a11y-зависимости):
```go
func startVTECapture(_ context.Context, readyPath, ptyPath string) (*vteCapture, error) {
	cmd := exec.Command(vteBin, "--window", "--",
		"bash", "-c", "echo ready > "+readyPath+"; cat > "+ptyPath)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s capture: %w", vteBin, err)
	}

	return &vteCapture{cmd: cmd}, nil
}
```
Сопутствующая дисциплина: teardown через EOF (`sendVTETeof`, `case_spike_caps.go:153-159`, ydotool `ctrl+d`), hex-чтение pty (`readPtyHexArmed`), pid-точечный reap без name-wide kills, сентинел PRIMARY (`primarySentinel` + `primaryReadNoNewline`, форма `vteChordRow`, `case_spike_caps.go:164-175`).

**(b) wezterm-фикстура + негативный counters-оракул** (`case_autocorrect.go:755-776` — спавн/тиддаун; `runAutocorrectTerminalSilent`, `case_autocorrect.go:806-838` — анатомия негатива с положительным контролем):
```go
func (s *stand) startWezterm(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, weztermBin, "start", "--always-new-process")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", weztermBin, err)
	}
	s.wezterm = cmd

	return nil
}
```
Для wezterm-ряда — обязательная дисциплина ×2 (два подряд полных раунда в одном запуске кейса, SPEC §7.2 rev D-55, Pitfall 1) + арминг-гейт по key-trace демона (НЕ по a11y-свидетелю — Pitfall 7) + caps re-pin строкой журнала (Pitfall 4).

**(c) Анатомия кейса коррекции** (`case_word.go:40-82`): активация → инжекция текста → гейт доставки (`waitForNew` key-trace) → жест (`injectKeys "Shift_R", "Shift_R"`) → грейп решения (`"msg":"action","n":2`) → грейп исхода (`"msg":"correction","outcome":"done"`) → контентный оракул. Для терминалов оракул — pty/hex вместо AT-SPI/zenity.

**(d) Журнальная дисциплина вердиктов** (`appendSpikeRow`; контракт секции — `case_spike_caps.go:26-31`: «HEADER TEXT IS THE CONTRACT», датированные строки, честные aborted-ветки; прецедент защиты секции тестом — `spike_journal_test.go`).

**Сентинел CLIPBOARD (критерий 3)** — замороженный ряд по форме `vteChordRow`: `printf 'CLIP-SENTINEL' | wl-copy` → коррекция → `wl-paste --no-newline` байт-нетронут; синтетический корпус only (ghbdtn), pty-артефакты tmp-only, hexlify, не коммитятся (V16).

---

### `mise.toml` — новые задачи `e2e-term-*`

**Analog:** сам себя — семья live-задач NOT in ci. Форма-прецедент (`mise.toml:167-169`):
```toml
[tasks.e2e-term04-negative]
description = "e2e: 09-02 term04-negative — the TERM-04 frozen negative under the unified layer (…), rows to adr-008-spike-log.md (live GNOME session; NOT in ci)"
run = "go run ./test/e2e -case term04-negative"
```
Новые задачи (`e2e-term-word-vte`, `e2e-term-word-wezterm`, `e2e-term-negative-witnesses`) копируют форму: описание несёт фазу/план, оракул и «NOT in ci»; re-pin `e2e-autocorrect-terminal-silent` (`mise.toml:220-222`) остаётся как есть.

## Shared Patterns

### Закрытый словарь отказов (closed slugs)
**Source:** `internal/session/actor.go:93-105` + `recordACAbstain` (`actor.go:2412`) + `skipCorrection` (`actor.go:1256`) + `refusalReason` (`actor.go:2990`)
**Apply to:** арминг-вето (новый slug `terminal-veto`), все новые отказы терминальной ветки (verify-dead-маршрутизация, чанк-аборт). Слоги — closed, word-free (D-20/D-21); каждый отказ считает ровно один slug.

### Лог-дисциплина D-20/D-21
**Source:** `logCorrectionDone`, `internal/session/actor.go:2907-2916`
**Apply to:** все новые INFO-записи исходов (D-24-дискриминатор, verify-after терминала). INFO — исход без содержимого; содержимое (token/result) — только DEBUG; слово НИКОГДА не в логах. Grep-формы стенда (`"msg":"correction","outcome":"done"`, `"msg":"action","n":2`) — замороженный контракт, расширять атрибутами, не переименовывать.

### Таймеры + эпохи (stale-deadline гашение)
**Source:** `pendingAfter.epoch` (`actor.go:469-477`), `afterExpiry` (`actor.go:1567-1576`), `VerifyExpiry` (`actor.go:1069-1082`), `rearmAfter` (`actor.go:1555-1561`)
**Apply to:** любые новые ждущие состояния терминальной ветки. Ожидание живёт в `time.AfterFunc`, никогда в хендлере (Pitfall 4 ADR-004); таймер-колбэк перепроверяет тег эпохи под мьютексом; `applySnapshot()` — одна конфиг-читка на событие.

### Мьютекс-дисциплина актора
**Source:** все хендлеры `actor.go` (паттерн «caller holds the mutex» + doc-комментарий), off-mutex прецеденты: `autoConfirm` (`actor.go:2447-2464` — живой GetRole вне мьютекса под своим дедлайном), `HandleSurroundingText` (`actor.go:841-845` — rung-полезная нагрузка исполняется снаружи)
**Apply to:** чанк-бурст (под мьютексом — это один эмиттер-_snapshot), перечек фокуса (состояние, уже гасимое `HandleLifecycle`), primary-read лег (subprocess — ТОЛЬКО вне мьютекса).

### Keycode-дисциплина (transcribed, never from memory)
**Source:** `backSpaceKeycode = 14` (`actor.go:50-53`), `ctrlVKeycode = 47` (`actor.go:1635-1639`), таблица `probeLetterKeycodes` (`internal/session/probe.go:49-69`)
**Apply to:** любые новые физические коды (чанк-бурст уже имеет BackSpace=0xff08/14 — новых кодов фаза не требует; select-chord — через существующий `ProbeForward`, `internal/session/probe.go:164-179`).

### Строгий конфиг: схема+валидация+док одним PR
**Source:** `internal/config/config.go:21-66, 94-123, 224-260` + `docs/CONFIG.md` (раздел «Схема», миграция D-53)
**Apply to:** новые ключи терминального класса. KnownFields-декодер отвергает незнакомый ключ ЦЕЛИКОМ (Pitfall 5); zero-value = старое поведение; правка CONFIG.md требует `go run ./cmd/skillgen` + golden `CommittedFileInSync`.

### Сентинелы и синтетический корпус (V16/D5)
**Source:** `primarySentinel`/`primaryReadNoNewline` (`test/e2e/case_spike_caps.go:164-175`), класс GS-SPIKE-SENTINEL
**Apply to:** CLIPBOARD-ряд (критерий 3), PRIMARY-лег wezterm. Только синтетика; восстановление пред-стейта или документирование затирания; pty-захваты tmp-only + hexlify.

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| Конфиг wezterm CopyMode-рецепта (фикстура стенда `--config-file`) | внешняя фиксация | — | Сторона wezterm, не демона; собственного конфига wezterm в репо нет. Фикстура стенда свободна; `~/.wezterm.lua` — чекпойнт владельца с backup (прецедент 09-03) |
| Wave-0 проба VM-стенда goswitch-stand (клавиатурность фоновых окон) | инфраструктурная задача | — | Внешняя среда (libvirt), не код; гейт выбора площадки живых рядов |

Для обоих: planner берёт механику из RESEARCH.md (Environment Availability, Runtime State Inventory), в кодовой базе прецедентов нет.

## Metadata

**Analog search scope:** `internal/{correct,session,config,clipboard,engine,appid}`, `test/e2e`, `mise.toml`, `docs/`
**Files scanned:** ~20 (из них прочитаны детально: buffer.go, runs.go, plan.go, verify.go, clipboard.go, config.go, probe.go, appid.go, engine.go:30-90, actor.go — 9 целевых диапазонов, actor_test.go:1-215, case_spike_caps.go:1-175, case_autocorrect.go:720-839, case_word.go:1-130, mise.toml, buffer_test.go:1-100, docs/CONFIG.md)
**Контроль зеркал:** все пути аналогов проверены `git ls-files` (2026-10-12) — tracked
**Pattern extraction date:** 2026-10-12
