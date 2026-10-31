# Phase 11: MENU — пункт «в чёрный список» - Pattern Map

**Mapped:** 2026-10-31
**Files analyzed:** 13 (изменения: menu.go, menu_test.go, appid.go, actor.go, actor_test.go, writer.go, writer_test.go, main.go, main_test.go, mise.toml, README.md, README.en.md, docs/CONFIG.md, docs/SPEC.md; новый файл: e2e-кейс)
**Analogs found:** 13 / 13 — все аналоги in-repo, git-tracked, прочитаны и верифицированы в этом сеансе

## File Classification

| Файл (изменение/создание) | Role | Data Flow | Ближайший аналог | Качество матча |
|---------------------------|------|-----------|------------------|----------------|
| `internal/indicator/menu.go` (мод.) | component (DBusMenu-рендерер, state-snapshot) | event-driven (D-Bus-сигналы) | сам `internal/indicator/menu.go` — паттерны расширения: ID-реестр, props-ветка, сеттер | exact (расширение) |
| `internal/indicator/menu_test.go` (мод.) | test (unit, wire-пины) | — | `TestMenuLayout`/`TestMenuMacroVisibility` в том же файле | exact |
| `internal/appid/appid.go` (мод.) | service (a11y-наблюдатель) | event-driven (StateChanged-сигналы) | сам `internal/appid/appid.go` — Option-паттерн `WithRoleCall` → новый `WithGainHook` | exact (расширение) |
| `internal/session/actor.go` (мод.) | service (актор-автомат) | event-driven | сам `internal/session/actor.go` — `pushMenuSync`, fold `applySnapshot`, `HandleLifecycle` | exact (расширение) |
| `internal/session/actor_test.go` (мод.) | test (unit, fake-дубли) | — | `fakeMenuSync` (5006-5054), `fakeAppid` (3276-3287) того же файла | exact |
| `internal/config/writer.go` (мод.) | service (persist-писатель) | file-I/O (atomic) | сам `internal/config/writer.go` — `setDocumentToggle` → новый append-метод | exact (расширение) |
| `internal/config/writer_test.go` (мод.) | test (unit, golden/round-trip) | — | корпус `TestWriter_*` того же файла | exact |
| `cmd/goswitchd/main.go` (мод.) | controller (click-композиция, проводка) | request-response (меню-клик) | `configToggle`/`newMenuToggles`/`flip` того же файла | exact |
| `cmd/goswitchd/main_test.go` (мод.) | test (unit композиции) | — | существующие кейсы `main_test.go` | exact |
| `test/e2e/case_menu_blocklist.go` (новый) | test (e2e-драйвер) | event-driven (D-Bus Event) | `test/e2e/case_menu.go` | exact |
| `mise.toml` (мод.) | config (mise-задачи) | — | существующие задачи `e2e-*` в `mise.toml` | exact |
| `README.md` / `README.en.md` (мод.) | docs | — | пины D-04: README.md:102, README.en.md:99 | exact |
| `docs/CONFIG.md`, `docs/SPEC.md` (мод.) | docs (спека-ревизия) | — | дисциплина дословных ревизий D-53/D-55 (см. RESEARCH Open Q4) | role-match |

**Tracked-source gate:** все названные пути проверены `git ls-files` — tracked; gitignored-зеркал в списке нет.

## Pattern Assignments

### 1. `internal/indicator/menu.go` — пункт id 16 + D-04 (component, event-driven)

**Аналог:** сам файл; расширение по трём готовым паттернам.

**Паттерн A: реестр menuID — новый ID, НЕ перенумерация** (строки 56-84):

```go
menuIDRoot        = int32(0)
menuIDEN          = int32(1)
menuIDRU          = int32(2)
menuIDSep1        = int32(3)
menuIDACToggle    = int32(4)
...
menuIDReload      = int32(15)
...
macroComboWhat = " — слово и язык"
macroChordWhat = " — смена режима"
```

- Новый пункт: `menuIDBlocklistApp = int32(16)` — следующий свободный после 15. e2e-кейс `menu-v2` пинит существующие id (`case_menu.go:40-43`, «a layout change breaks this case loudly, by design»), юнит `TestMenuLayout` пинит порядок детей — сдвиг чисел запрещён.
- **D-04 (строка 84):** `macroChordWhat = " — смена режима"` → `" — язык"`. Метка собирается в `bindingProps(m.modeChord, macroChordWhat)` (строка 475). Пины к обновлению: `menu_test.go:160` («super+space — смена режима»), `README.md:102`, `README.en.md:99`.

**Паттерн B: позиция в layout и фильтр диапазона** (строки 444-457 и 364-366):

```go
func (m *Menu) rootKidsLocked() []int32 {
	ids := []int32{
		menuIDEN, menuIDRU, menuIDSep1, menuIDACToggle, menuIDSoundToggle,
		menuIDMacro1, menuIDMacro2, menuIDMacro3,
	}
	...
	return append(ids, menuIDSep2, menuIDSettings, menuIDAbout, menuIDStatus, menuIDReload)
}
```

```go
	for _, id := range ids {
		if id < menuIDEN || id > menuIDReload {
			continue
		}
```

- Вставка в `rootKidsLocked` сразу после `menuIDACToggle` (дискреция владельца, CONTEXT). **Обязательно расширить верхнюю границу фильтра `GetGroupProperties`** (`id > menuIDReload` → включить 16), иначе props id 16 молча выпадают из ответов renderer'а.

**Паттерн C: props-билдеры — галочка (D-01) и enable (D-03)** (строки 599-613, 617-619):

```go
func standardProps(label string, enabled bool) map[string]dbus.Variant {
	return baseProps(label, enabled, true)
}

func toggleProps(label, kind string, on bool) map[string]dbus.Variant {
	p := standardProps(label, true)
	p[menuPropToggleType] = dbus.MakeVariant(kind)
	p[menuPropToggleState] = dbus.MakeVariant(stateInt(on))

	return p
}
```

Готовый образец галочки — `checkProps` (строки 500-506):

```go
func (m *Menu) checkProps(id int32) map[string]dbus.Variant {
	if id == menuIDSoundToggle {
		return toggleProps(labelSoundToggle, menuToggleCheckmark, m.soundEnabled)
	}

	return toggleProps(labelACToggle, menuToggleCheckmark, m.acEnabled)
}
```

- Новый пункт: ветка в `itemPropsLocked` (строки 462-487) с `menuToggleCheckmark`: `on` = приложение уже покрыто (D-01), `enabled` = identity известна И свежа И список не полон (D-03, потолок 64). Label строится с `doubleUnderscores(name)` (строки 649-651 — GTK-мнемоники, Pitfall 4).
- **Pitfall 7 (строки 93-101):** каждая эмиссия несёт `[]removedProps{}` — пустой слайс-литерал, НИКОГДА nil (nil ломает сигнатуру `a(ias)` и паникует первую живую эмиссию).
- **Revision не трогается** (строки 42-45): пункт ВСЕГДА в layout (D-03 «пункт виден»), все его изменения — property-динамика `ItemsPropertiesUpdated`, бамп revision — только на смену НАБОРА пунктов.

### 2. `internal/indicator/menu.go` — новый сеттер (образец `SetCorrections`)

**Аналог:** `SetCorrections`, строки 262-273 — дисциплина dedupe + дельта:

```go
func (m *Menu) SetCorrections(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if n == m.corrections {
		return
	}
	m.corrections = n
	m.emitSignal(signalItemsPropsUpdated,
		[]menuItemProps{m.labelDelta(menuIDStatus)},
		[]removedProps{})
}
```

- Новый сеттер (рекомендация RESEARCH: один метод с полями, напр. `SetBlocklistApp(app string, fresh, covered, full bool)`) копирует форму: lock → dedupe (identical values — silent no-op) → store → `emitSignal(signalItemsPropsUpdated, deltas, []removedProps{})`.
- Дельты: label — `labelDelta(id)` (строки 563-567, считает из снапшота AT REQUEST TIME), toggle-state — `stateDelta(id, on)` (строки 554-558).
- Push-моменты (RESEARCH Pattern 2): gain-хук наблюдателя; реальный FocusOut (fresh=false); fold `applySnapshot` (covered/full по новому документу); install-пуш при attach (образец `menu.SetKeys(...)` в OnConn, main.go:275).

### 3. `internal/appid/appid.go` — хук свежести `WithGainHook` (service, event-driven)

**Аналог:** сам файл; Option-паттерн `WithRoleCall`, строки 96-105:

```go
type Option func(*Observer)

func WithRoleCall(fn RoleCall) Option {
	return func(o *Observer) { o.roleCall = fn }
}
```

Новый `WithGainHook(func(name string))` добавляется ровно так же (поле + Option).

**Место вызова хука — цикл `run`, строки 297-314 (корень Pitfall 1 — cache-keep):**

```go
	for {
		select {
		case <-ctx.Done():
			return
		case sig, ok := <-feed:
			if !ok {
				return
			}
			if name, gained := focusGain(sig); gained {
				o.mu.Lock()
				o.app = name
				o.focusSender = dbus.Sender(sig.Sender)
				o.focusPath = sig.Path
				o.mu.Unlock()
			}
		}
	}
```

- Хук вызывается **ПОСЛЕ снятия `o.mu`** (вне критсекции, после store) — Pitfall 9: вызов под `appid.mu` с заходом в `actor.mu` инвертирует односторонний порядок actor.mu → sink.mu (дисциплина задокументирована в `pushA11y`, actor.go:2250-2263) → дедлок.
- Семантика кэша НЕ меняется: `focusGain` (строки 141-156) по-прежнему отбрасывает loss-события и немаркированные пути, кэш держит «последнюю известную» — гейты D-53 продолжают читать как сегодня; fresh-признак — новая надстройка только для меню.

### 4. `internal/session/actor.go` — fresh-состояние, `PushFocusedApp`, MenuSync-расширение (service, event-driven)

**Аналог 1: MenuSync-шов** (строки 1889-1895):

```go
type MenuSync interface {
	SetMode(symbol string)
	SetAutocorrectEnabled(on bool)
	SetSoundEnabled(on bool)
	SetKeys(tap, combo, chord string)
	SetCorrections(n int)
}
```

- Расширяется методом `SetBlocklistApp(...)`; контракт пушей (комментарий 1885-1888): синхронно под мьютексом актора, быстро, никогда не паникует и не блокирует; меню само дедуплицирует.

**Аналог 2: push из fold** (`pushMenuSync`, строки 2282-2303) — дисциплина dedupe-ключа на акторе и одностороннего порядка блокировок:

```go
	if a.menuSync != nil {
		a.menuSync.SetAutocorrectEnabled(snap.Autocorrect.Enabled)
		a.menuSync.SetSoundEnabled(a.soundEnabled)
		a.menuSync.SetKeys(snap.Hotkeys.TapKey, snap.Hotkeys.WordLayoutCombo, snap.Hotkeys.ModeSwitchChord)
		a.menuSync.SetCorrections(a.corrDone)
	}
```

**Аналог 3: fold blocklist** (`applySnapshot`, строки 2188-2195) — точка пересчёта covered/full и вызова `ensureAppid`:

```go
	a.opts.AutoCorrectEnabled = snap.Autocorrect.Enabled
	a.opts.AutoCorrectBlocklist = snap.Autocorrect.AppsBlocklist
	...
	a.refreshACBlocklist(snap.Autocorrect.AppsBlocklist)
	a.refreshACTerminals(snap.Correction.Terminals)
	a.ensureAppid() // the per-app list or the enabled autocorrect layer may have appeared with this document
```

**Аналог 4: точное условие ленивого старта** (`ensureAppid`, строки 2835-2850) — Pitfall 2:

```go
func (a *Actor) ensureAppid() {
	if a.appidStarted || (len(a.opts.MACRApps) == 0 && !a.opts.AutoCorrectEnabled) {
		return
	}
	a.appidStarted = true
	src, err := a.startAppid()
	if err != nil {
		slog.Warn("app identity unavailable", "error", err)

		return
	}
```

- При выключенной автокоррекции (дефолт D-54) наблюдатель не стартует — пункт мёртв. Рекомендация RESEARCH (Open Q1): безусловный старт при проводке демона — новый публичный метод рядом с `actor.SetMenuSync(menu)` (main.go:276); fallback — документировать «активен после первого включения автокоррекции».

**Аналог 5: реальный FocusOut** (`HandleLifecycle`, строки 1282-1360) — место сброса fresh:

```go
		synthetic := false
		if kind == engine.LifecycleFocusOut && a.flipCredit > 0 {
			a.flipCredit--
			synthetic = true
		}
```

- **Только НЕ-synthetic ветка** (без flipCredit — G-2-3) ставит fresh=false; synthetic FocusOut (собственный flip демона) — не «смена фокуса окна» и свежесть не сбрасывает.

**Аналог 6: блок-гейт НЕ трогается** (`autoCorrectBoundary`, строки 3322-3349) — фаза не меняет ни строчки здесь:

```go
	app, appKnown := a.acFocusedApp()
	if appKnown {
		for _, re := range a.acBlocklist {
			if re.MatchString(app) {
				a.recordACAbstain(acReasonAppBlocked)

				return acPayload{}, false
			}
		}
	}
```

- Дисциплина матча для галочки D-01 — та же: «покрыто» = любой скомпилированный паттерн совпал с identity (`re.MatchString(app)`).

**Аналог 7: Status-поверхность для click-композиции** (строки 1140-1148, 1188-1195) — тогглы читают именно StatusSnapshot; идентичность доводится новым полем Status либо отдельным аксессором (выбор планировщика):

```go
	AutoCorrectEnabled     bool
	AutoCorrectFired       int
	AutoCorrectAbstained   int
	AutoCorrectSkipReasons map[string]int
```

### 5. `internal/config/writer.go` — `AppendAutocorrectBlocklist` (service, file-I/O)

**Аналог: `setDocumentToggle` — ЕДИНЫЙ persist path** (строки 116-139):

```go
func setDocumentToggle(path, section, key string, on bool) error {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return ensureFullDocument(path, section, on)
	}
	if err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	var root yaml.Node
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return fmt.Errorf("decode config %s: %w", path, err)
	}
	// The walk starts at the DOCUMENT's root mapping — the DocumentNode
	// itself carries no key/value pairs (Pitfall 6: a walk from the
	// document node silently changes nothing).
	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return fmt.Errorf("edit config %s: %w", path, errWriterNotMapping)
	}
	...
	return rewriteAtomically(path, &root)
}
```

- Новый метод повторяет каркас: read → absent-file ветка (`ensureFullDocument`-образец, строки 145-159: создать полный дефолтный документ с паттерном на месте через `buildFullDocument`) → Node round-trip → правка → **pre-write строгая валидация** → `rewriteAtomically`.
- **Absent-file ветка обязательна** (Assumption A4: у части пользователей секции/ключа нет).

**Правка списка — по константам и helper'ам пакета** (строки 25-35, 191-256):

```go
	sectionAutocorrect = "autocorrect"
	...
	keyAppsBlocklist    = "apps_blocklist"
```

```go
// scalarString builds a plain string scalar.
func scalarString(s string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s}
}

// emptySeq builds the empty flow sequence — `key: []` в документе.
func emptySeq() *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
}
```

- Дедуп: паттерн уже в списке → no-op success (idempotent, образец `TestWriter_IdempotentFlip`).
- Потолок: `len(list) >= maxAutocorrectBlocklist` → sentinel-error (дисциплина err113 — пакетные сентинелы вида `errAutocorrectBlocklistOverCeil` в config.go, обёртка `fmt.Errorf` с путём).
- `buildSectionNode(section, on)` (строки 224-246) **НЕ переиспользовать** — сигнатура bool-специфична (Anti-Pattern RESEARCH).
- **Pre-write валидация** (критерии 2 и 4): закодировать в байты, прогнать семантику `Load` (`KnownFields(true)` + `Validate()`, load.go:27-28 и load.go:48) по байтам в памяти; отказ → ошибка, оригинал байт-нетронут (дисциплина `TestWriter_BrokenSourceUntouched`).
- Стиль 0 (plain) в `scalarString` — yaml.v3 сам закавычит спец-символы; ручное квотирование не добавлять.
- Писатель получает ГОТОВЫЙ паттерн (`QuoteMeta` — в композиции main.go); писатель не знает про identity и **не логирует содержимое секций** (D-20/D-21, шапка writer.go:1-8).

**Атомарная посадка — готовая** (строки 303-350, пинена T-07-03-01/02):

```go
func writeAtomically(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	...
	if err := os.Chmod(tmpName, configPerm); err != nil { // 0600, configPerm (строка 44)
	...
	if err := os.Rename(tmpName, path); err != nil {
```

### 6. `cmd/goswitchd/main.go` — click-композиция + проводка (controller, request-response)

**Аналог 1: `configToggle.flip` — порядок клика** (строки 475-494):

```go
func (t *configToggle) flip() {
	if t.path == "" {
		slog.Warn("config toggle unavailable", "component", "tray indicator", "reason", "no config path")

		return
	}
	on := !t.read()
	if err := t.write(t.path, on); err != nil {
		slog.Warn("config toggle write failed", "component", "tray indicator", "error", err)

		return
	}
	t.push(on)
	if t.reload == nil {
		slog.Warn("config toggle apply deferred", "component", "tray indicator", "reason", "no config source")

		return
	}
	t.reload()
}
```

Новая композиция копирует порядок read → guards → write → push → reload, с гвардами RESEARCH Pattern 4:
- пустой путь конфига — WARN-отказ (образец WR-03 выше);
- identity неизвестна/несвежа — **молчаливый no-op** (fail-closed, второй пояс: renderer и так не диспатчит clicked по enabled:false);
- covered (галочка) — no-op по D-01;
- full (потолок 64) — no-op, отказ виден состоянием пункта (Pitfall 6; десктоп-уведомления на этом столе невидимы — вердикт UAT фазы 7);
- паттерн: `regexp.QuoteMeta(identity)` (Pitfall 7/метасимволы: `org.example.App[1]`, `wezterm (2)`).

**Аналог 2: синхронный reload** (строки 249-255) — «замолкает без рестарта» в том же клике:

```go
			toggleReload := syncReload
			if reload != nil {
				toggleReload = func() {
					syncReload()
					actor.FoldAppliedConfig()
				}
			}
```

**Аналог 3: проводка и install-пуш** (OnConn, строки 230-307) — Callbacks собираются рядом с `ToggleAutocorrect: toggles.ac.flip`; install-пуши в порядке `menu.SetVersion/SetAutocorrectEnabled/SetSoundEnabled/SetKeys` (272-275), затем `actor.SetMenuSync(menu)` (276). Новый пункт добавляет: `cb.BlockCurrentApp = ...` (поле в `indicator.Callbacks`), install-пуш начального состояния пункта, проводку gain-хука (`appid.WithGainHook` → публичный метод актора). Push на атомарный слот с nil-гвардом — образец `menu.Load()` в newMenuToggles (374-377):

```go
			push: func(on bool) {
				if m := menu.Load(); m != nil {
					m.SetAutocorrectEnabled(on)
				}
			},
```

**Аналог 4: поле Callbacks** — `internal/indicator/indicator.go:98-107` (структура `Callbacks`), новый callback рядом с `ToggleAutocorrect`/`ToggleSound`; dispatch-кейс в `dispatchClick` (menu.go:403-430) по образцу:

```go
	case menuIDACToggle:
		if m.cb.ToggleAutocorrect != nil {
			m.cb.ToggleAutocorrect()
		}
```

### 7. `test/e2e/case_menu_blocklist.go` — новый e2e-кейс (test, event-driven)

**Аналог:** `test/e2e/case_menu.go` (exact: тот же role + data flow — D-Bus Event-драйвер по живому демону, без ydotool/мыши).

**Пины канала и ID** (строки 31-48):

```go
const (
	menuBusName = "org.djarvur.goswitch"
	menuObjPath = "/Menu"
	menuIface   = "com.canonical.dbusmenu"

	menuEventMethod = "com.canonical.dbusmenu.Event"

	menuIDEN          = int32(1)
	menuIDRU          = int32(2)
	menuIDACToggle    = int32(4)
	menuIDSoundToggle = int32(5)
	...
	caseMenuV2 = "menu-v2"
)
```

- Новый кейс: клик `com.canonical.dbusmenu.Event` по id 16 → документ вырос ровно на QuoteMeta-паттерн → строгий `Load` → статус; клик по отмеченному пункту — no-op; автокоррекция молчит в заблокированном приложении (oracle-машина `case_autocorrect.go` blocklist-silent, fixture-identity — Assumption A6).
- Документ-конфиг кейса — образец `menuConfigDoc` (строки 66-89: явная секция `autocorrect` с `apps_blocklist: []`, никакого overlay).
- Регистрация кейса — реестр в `test/e2e/main.go`; mise-задача — по образцу существующих `e2e-*` в `mise.toml`.

### 8. Тестовые файлы — расширение по готовым корпусам

**`internal/indicator/menu_test.go`** — `TestMenuLayout` wantKids (строки 150-166) — ожидаемая RED-ступень TDD (правятся длина/порядок + пин D-04 на строке 160):

```go
	wantKids := []menuKidWant{
		{menuIDEN, labelEN, menuTypeStandard, true, true, menuToggleRadio, 0},
		...
		{menuIDMacroChord, "super+space — смена режима", menuTypeStandard, false, true, "", 0},
		...
	}
```

Дельты/сигнатуры — образцы `TestMenuMacroVisibility` (199-249: исчезновение строки, ревизия, visible:false) и `TestMenuSignalFormPins`.

**`internal/config/writer_test.go`** — четыре готовых паттерна для append-версий:
- golden-дифф/round-trip: `TestWriter_FlipRoundTrip` (101-132) — write → строгий `config.Load(path)` → assert (строки 107-113);
- комментарии целы: `TestWriter_CommentsSurvive` (137-156, маркеры строк 147-151);
- idempotent-дедуп: `TestWriter_IdempotentFlip` (160-179 — байтовое сравнение до/после);
- атомарность: `TestWriter_AtomicWrite` (184-206 — нет `.tmp`-артефактов, mode 0600, строки 194-205) и `TestWriter_BrokenSourceUntouched` (211+).

**`internal/session/actor_test.go`** — `fakeMenuSync` (строки 5006-5054): новое поле + метод-рекордер по образцу:

```go
// SetCorrections records the corrections-counter push (owner UAT
// 2026-10-05: the Status info row renders the live count).
func (f *fakeMenuSync) SetCorrections(n int) {
	f.mu.Lock()
	f.corrections = append(f.corrections, n)
	f.mu.Unlock()
}
```

Identity-двой — `fakeAppid` (строки 3276-3287): `UseAppid(fakeAppid{app: ...})`, сценарии fresh/unknown — новые кейсы поверх этой машины.

**`cmd/goswitchd/main_test.go`** — кейсы композиции по образцу существующих тоггл-тестов того же файла.

## Shared Patterns

### S1. Меню — «глупый» рендерер снапшота; props AT REQUEST TIME
**Источник:** `internal/indicator/menu.go:121-132` (док-комментарий Menu), `itemPropsLocked` (462-487).
**Применять к:** всему пункту id 16 — label/enabled/галочка всегда считаются из снапшота в момент запроса; живых a11y-запросов при открытии меню НЕТ (D-02).

### S2. Дельты — только ItemsPropertiesUpdated; пустой removedProps; revision заморожен
**Источник:** `internal/indicator/menu.go:93-101` (Pitfall 7), сеттеры 205-326.
**Применять к:** каждой эмиссии нового сеттера: `emitSignal(signalItemsPropsUpdated, deltas, []removedProps{})`; `[]removedProps{}` литерал, не nil. Revision не бампить никогда (пункт всегда в layout — D-03).

### S3. Один persist path: writer — единственная точка записи документа
**Источник:** `internal/config/writer.go:1-8, 116-139`; дедуп/потолок — в writer'е (same package, константа `maxAutocorrectBlocklist`, config.go:29).
**Применять к:** клик-композиция never bypass — никаких прямых правок YAML вне writer-метода; композиция не знает схему.

### S4. Fail-closed: неизвестное = молчание; сентинелы err113
**Источник:** `internal/appid/appid.go:52-74` (ErrBusClosed/ErrRoleUnknown — concrete-типы, errors.Is-friendly); гейт actor.go:3322-3349 («unknown identity PASSES» для гейта, но для UI D-03 — пункт неактивен); writer-сентинелы config.go (`errAutocorrectBlocklistOverCeil` и пр.).
**Применять к:** гварам композиции (unknown/fresh=false — silent no-op), потолку writer'а, decoдеру.

### S5. Строгая валидация до rename; отказ байт-нетронут
**Источник:** `internal/config/load.go:27-28` (`dec.KnownFields(true)`), `config.go:472-501` (потолок, пустой паттерн, regex-compile — каждое имя поля и индекс в ошибке, D-33); `TestWriter_BrokenSourceUntouched`.
**Применять к:** pre-write валидации writer-метода — семантика Load по байтам в памяти до `rewriteAtomically`.

### S6. Атомарная посадка конфига: same-dir temp + rename + 0600
**Источник:** `internal/config/writer.go:303-350` (пинен T-07-03-01/02).
**Применять к:** не строить свой writer — только `rewriteAtomically`.

### S7. Синхронный под-мьютексный push; односторонний порядок блокировок actor.mu → sink.mu
**Источник:** `internal/session/actor.go:1885-1888` (контракт MenuSync), `pushA11y` 2250-2271 (дисциплина), `pushMenuSync` 2282-2303.
**Применять к:** `PushFocusedApp`, вызову gain-хука ВНЕ `appid.mu` (Pitfall 9 — инверсия → дедлок).

### S8. Экранирование: QuoteMeta для regex, doubleUnderscores для label
**Источник:** `doubleUnderscores` menu.go:645-651; `regexp.QuoteMeta` (stdlib; RESEARCH Don't-Hand-Roll).
**Применять к:** имя приложения в двух представлениях: в label — `doubleUnderscores`, в паттерн blocklist — `QuoteMeta` (в композиции, до writer-вызова).

### S9. Уведомления владельцу — строками/состоянием меню, не десктоп-уведомлениями
**Источник:** вердикт UAT фазы 7 (CONTEXT Established Patterns); прецедент — grey-строки About/Status (menu.go:617-619 `macroProps`).
**Применять к:** отказ по потолку 64 — видимое состояние пункта (disabled + label с заполненностью, формулировка — дискреция).

### S10. D-20/D-21: в журнале — только факт, никогда содержимое секций
**Источник:** шапка `internal/config/writer.go:1-8` («The writer logs nothing»).
**Применять к:** новым лог-записям writer-пути и композиции (путь и исход, не паттерн/имя секции; имя приложения в label — данные конфиг-области, слову не тождественно — RESEARCH: D-20/D-21 не нарушает).

## Hot-reload: правильное имя механизма

**Источник:** `internal/config/watch.go:13-21`:

```go
const defaultPollInterval = 2 * time.Second
```

CONTEXT.md упоминает «fsnotify watcher» — **устарело**: с 2026-10-08 (D-32 REV) — stat-опрос mtime+size каждые 2 с (inotify выведен из эксплуатации). Для фазы разницы нет (атомарный rename писателя виден опросу; fold — на следующем клавиатурном событии; клик дополнительно делает синхронный reload). Планировать и документировать по коду: watcher = stat-poll.

## No Analog Found

| Файл | Role | Data Flow | Причина |
|------|------|-----------|---------|
| — | — | — | Все файлы фазы имеют точные in-repo аналоги. Fresh-механизм (gain-хук + сброс на реальный FocusOut) — новая ЛОГИКА, но собирается из готовых паттернов: Option-инъекция (appid.go:98-105), lifecycle-ветка (actor.go:1282-1360), push-дисциплина (S7). Для него — RESEARCH Pattern 3 как спецификация. |

## Metadata

**Analog search scope:** `internal/indicator/`, `internal/appid/`, `internal/config/`, `internal/session/`, `cmd/goswitchd/`, `test/e2e/`
**Files scanned:** ~20 (6 прочитано целиком, остальные — целевые диапазоны по пинам RESEARCH; все цитаты сверены с трекнутыми исходниками)
**Pattern extraction date:** 2026-10-31
**Верификация пинов D-04:** `internal/indicator/menu.go:84`, `internal/indicator/menu_test.go:160`, `README.md:102`, `README.en.md:99`; grep «смена режима» по `test/` — пуст (e2e-пинов хвоста нет).
