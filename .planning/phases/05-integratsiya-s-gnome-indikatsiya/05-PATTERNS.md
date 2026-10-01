# Phase 5: Интеграция с GNOME: индикация и двухисточниковое переключение - Pattern Map

**Mapped:** 2026-09-28
**Files analyzed:** 16 (4 new, 12 modified)
**Analogs found:** 16 / 16 (a mature 4-phase repo: nearly every touched file is its own best analog — the planner needs the *in-file precedent* for each new seam, named below with line anchors)

All analog paths verified git-tracked (`git ls-files` non-empty for every path cited below). No mirror/untracked paths are referenced. All Go work follows the project skill `.zcode/skills/go-ultimate/SKILL.md` (no `pkg/`, interfaces at point of use, `errors.Is/As`, `fmt.Errorf("...: %w")`, `-race` always, `package xxx_test` tests named `TestF_suffixCamelCase`).

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `engine/conn.go` (modify) | service (D-Bus protocol adapter) | event-driven + request-response | itself — `Config.PostRegister` seam, `serve`, `waitBusLoss` | exact |
| `engine/engine.go` (modify, maybe) | adapter (per-context engine object) | event-driven | itself — `FocusIn` name logging, `lifecycle` forwarding | exact |
| `engine/` seam tests (new file, e.g. `engine/conn_switcher_test.go`) | test | request-response | `engine/wire_test.go` + `internal/session/actor_test.go` fakeSink | role-match |
| `internal/session/actor.go` (modify) | service (event consumer / state machine) | event-driven | itself — `flipScript`, `settleCorrectionFlip`, `modeSwitchChord`, `HandleLifecycle` | exact |
| `internal/session/actor_test.go` (modify) | test | event-driven | itself — `fakeSink`, `TestActor_FlipOnSingle`, `TestActor_FlipEmitsPanelSymbol` | exact |
| `internal/install/install.go` (modify) | service (installer) | batch + transform (sources wrap) | itself — `takeoverSources`, `saveState`, `savedSources`, `derivedEngine` | exact |
| `internal/install/install_test.go` (modify) | test | batch | itself — `fakeRunner`, `TestInstall_Sequence`, `TestInstall_SecondInstallKeepsOriginalBackup` | exact |
| `internal/install/selfcheck.go` (modify) | service (audit probe) | request-response | itself — `checkInputSource` | exact |
| `internal/activate/activate.go` (modify) | service (restart recovery) | request-response | itself — `IfOwned`, `parseSources`, `reactivate` | exact |
| `internal/activate/activate_test.go` (modify) | test | request-response | itself — `fakeRunner`, `TestIfOwned*` corpus | exact |
| `cmd/goswitchd/main.go` (modify) | config (composition root) | wiring (n/a) | itself — `engineConfig` PostRegister closure, `newActor` | exact |
| `docs/adr/ADR-006-two-engine-revision.md` (new) | documentation (ADR) | n/a | `docs/adr/ADR-001-layout-switching-mechanism.md` | exact (format) |
| `test/e2e/` switch-case extension (`case_combo.go` and/or new `case_switch.go`) | e2e test | batch (inject + journal grep) | `test/e2e/case_combo.go` `runLayoutSingle` | exact |
| `test/e2e/main.go` (modify) | e2e harness | batch | itself — case registry, `snapshotDesktop`/`restoreGsettings`/`verifyRestored` | exact |
| `test/e2e/matrix.go` (modify) | e2e config (YAML matrix) | batch | itself — `matrixCase` validation, `Mode` oracle field | exact |
| Wave-1 spike script/checklist (new, `.planning` quick-task or e2e case) | spike / e2e probe | batch (reversible desktop mutation) | `test/e2e/case_d01.go` + `docs/adr/d01-experiment-log.md` | exact (behavioral) |

## Pattern Assignments

### `engine/conn.go` (service, event-driven + request-response) — BindSwitcher seam + GlobalEngineChanged subscription

**Analog:** the file itself. Three in-file precedents combine:

**Config-seam precedent** — `PostRegister` is the exact shape `BindSwitcher` copies (engine/conn.go:31-44):
```go
type Config struct {
	Component Component
	Engines   []EngineDesc
	Handler   EventHandler // may be nil — observer-only engine.

	// PostRegister is the optional self-reactivation seam (SY8): invoked
	// synchronously after EVERY successful RegisterComponent — generation 0
	// and every reconnect generation. nil = no-op. It must be cheap and
	// non-fatal: serve runs it inline before waitBusLoss, so a hang here
	// stalls the generation (the activate package's 10 s per-call timeout
	// bounds it).
	PostRegister func(ctx context.Context, generation int)
}
```
BindSwitcher adds: `BindSwitcher func(switch func(ctx context.Context, engineName string) error)` — invoked per connection **generation** (the conn is re-created every `serve` cycle, conn.go:74-137; a switcher captured from a dead generation must never survive — rebind inside `serve` after `RequestName` succeeds, beside the `PostRegister` call at conn.go:132-134). nil = no-op, same contract.

**D-Bus call precedent** — the only in-repo method call TO ibus-daemon; SetGlobalEngine copies this shape verbatim (engine/conn.go:122-126):
```go
	ibus := conn.Object(ibusService, ibusPath)
	call := ibus.CallWithContext(ctx, ibusService+".RegisterComponent", 0, dbus.MakeVariant(component))
	if call.Err != nil {
		return fmt.Errorf("register component: %w", call.Err)
	}
```
Wire signature live-verified 2026-09-28: `SetGlobalEngine(in s engine_name)`. The switcher closure is `ibus.CallWithContext(ctx, ibusService+".SetGlobalEngine", 0, name).Err`. Constants `ibusService`/`ibusPath` already exist (conn.go:17-20).

**Signal-channel precedent** — the subscription must coexist with (or replace) this drain loop; note the comment says signals "carry no meaning for the tracer" — Phase 5 changes that, so the drain becomes a dispatch select (engine/conn.go:139-156):
```go
func waitBusLoss(ctx context.Context, conn *dbus.Conn) error {
	signals := make(chan *dbus.Signal, signalBufferSize)
	conn.Signal(signals)
	for {
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-signals:
			if !ok {
				return errBusClosed
			}
		}
	}
}
```
Add `conn.AddMatchSignal(dbus.WithMatchObjectPath(ibusPath), dbus.WithMatchInterface(ibusService))` before `conn.Signal(signals)`; keep `signalBufferSize` (conn.go:46-48). Channel-close = bus-loss verdict must survive the refactor.

**Imports pattern** (conn.go:3-14): `context`, `errors`, `fmt`, `log/slog`, ... `"github.com/godbus/dbus/v5"` — no new imports needed for the seam.

**Error handling precedent** (conn.go:56-69, Run loop): transient errors → `slog.Warn` + backoff; `errNotPrimaryOwner` fatal. The switcher's failures are NEVER fatal — WARN only (the flip-loss is observable in the journal; criterion 3).

---

### `engine/engine.go` (adapter, event-driven) — optional: engine name at the lifecycle seam

**Analog:** itself. `FocusIn` already logs the engine name but drops it at the seam (engine/engine.go:246-253):
```go
func (e *Engine) FocusIn() (err *dbus.Error) {
	defer recoverHandler("FocusIn", &err)
	slog.Info("focus_in", "engine", e.name)
	e.lifecycle(LifecycleFocusIn)

	return nil
}
```
And `lifecycle` forwards kind only (engine/engine.go:449-454):
```go
func (e *Engine) lifecycle(kind LifecycleKind) {
	if e.handler != nil {
		e.handler.HandleLifecycle(kind)
	}
}
```
Research Open Question 3 recommends the smallest ripple: a dedicated sync entry point (new seam) fed from the subscription AND from FocusIn's name, KEEPING `EventHandler.HandleLifecycle(kind LifecycleKind)` unchanged (engine/engine.go:83-97) — extending the signature ripples through every test double. `recoverHandler` (engine/engine.go:127-135) is the panic-safety contract any new exported handler path must carry (INTEG-05).

**Panic-shim pattern** (engine/engine.go:127-135) — copy verbatim onto any new exported method:
```go
func recoverHandler(method string, dbusErr **dbus.Error) {
	if r := recover(); r != nil {
		slog.Error("handler panic contained",
			"method", method,
			"panic", fmt.Sprint(r),
			"stack", string(debug.Stack()))
		*dbusErr = nil
	}
}
```

---

### `engine/` seam tests (new test file) (test, request-response)

**Analog:** `engine/wire_test.go` for package/style (`package engine_test`, plain stdlib `testing`, helper extract to stay under funlen — wire_test.go:7-25) and `internal/session/actor_test.go` fakeSink for the recording-double style (actor_test.go:115-178: struct with `sync.Mutex`, recorded ops slice, hook fired OUTSIDE the lock). godbus `Signal`/`AddMatchSignal` signatures verified in module cache v5.2.2 (conn.go:646-694, match.go:35-53 per research). The existing `waitBusLoss` behavior (ctx-done → nil, channel-close → `errBusClosed`) must stay pinned when the drain becomes a dispatch.

---

### `internal/session/actor.go` (service, event-driven) — flip sites route through the switcher seam; mode sync; Status active engine

**Analog:** itself. All four flip sites and the sync surface:

**Core flip** — the single choke point every caller already funnels through (actor.go:1291-1302):
```go
func (a *Actor) flipScript() {
	if a.mode == modeEN {
		a.mode = modeRU
		slog.Info("mode", "to", "ru")
	} else {
		a.mode = modeEN
		slog.Info("mode", "to", "en")
	}
	if a.eng != nil {
		a.eng.UpdateModeSymbol(a.modeSymbol())
	}
}
```
Callers today: `ExpiryAt` Single decision (actor.go:598), `settleCombo` (actor.go:647), `setScriptMode` (actor.go:687), `modeSwitchChord` (actor.go:1433). The seam change lands INSIDE `flipScript` (or a `flipTo(mode scriptMode)` generalization): derive the target engine from the mode (`modeEN ↔ "goswitch-en"`, `modeRU ↔ "goswitch-ru"`), call the switcher. The `"msg":"mode","to":"ru"` record is the e2e oracle — byte-stable contract, do not change its shape.

**SET-semantics precedent** — `settleCorrectionFlip` already computes the target script (the "режим = скрипт результата" rule); only the execution changes (actor.go:667-689):
```go
func (a *Actor) settleCorrectionFlip(converted []rune) {
	if a.comboPending || !a.opts.FlipAfterCorrection {
		return
	}
	// The mode is SET to the converted text's script, never toggled (owner
	// rule 2026-09-28): ...
	switch scriptOf(converted) {
	case modeEN:
		a.setScriptMode(a.mode == modeRU)
	case modeRU:
		a.setScriptMode(a.mode == modeEN)
	}
}
```

**Mode-sync entry** — the FocusIn branch of `HandleLifecycle` is where engine identity already re-asserts state (the UpdateModeSymbol self-heal precedent, actor.go:520-529):
```go
	case engine.LifecycleFocusIn, engine.LifecycleEnable, engine.LifecycleDisable:
		slog.Debug("lifecycle", "kind", kind.String())
		if kind == engine.LifecycleFocusIn && a.eng != nil {
			// The panel indicator self-heals on every focus gain ...
			a.eng.UpdateModeSymbol(a.modeSymbol())
		}
```
The new `SyncEngine(name string)`-style entry point follows this shape: under `a.mu`, correct `a.mode` to the named engine's script, WARN (not INFO) when it drifted — the daemon-follows-the-bus rule (single writer: goswitch flips, mutter's indicator clicks are followed, never fought).

**FocusOut hard-reset** — the desired "переключение сбрасывает контекст" semantics already exist; the flip's OWN focus pair will re-enter here (self-echo handling, Pitfall 5) (actor.go:499-519):
```go
	case engine.LifecycleFocusOut, engine.LifecycleReset:
		a.fsm.Feed(hotkey.Reset{}, a.elapsed())
		a.buf.HardReset()
		// the caches belong to the input context that just left
		a.surr = nil
		a.sel = selectionState{}
		a.resolvePending()
		a.clearAfter()
		...
		a.comboPending = false
```

**Status surface** — `StatusSnapshot` is where the active-engine field lands (actor.go:401-427; struct at 374-385). D-20/D-21: counts and engine names only, never typed text.

**Mutex discipline** — Pitfall 4: the D-Bus switcher call must NOT block under `a.mu`. In-repo precedent for off-mutex work: `rungMu` (actor.go:85, WR-01) and `HandleSurroundingText`'s armed-payload handoff (actor.go:548-552). Either a hard deadline ctx (≤40 ms) inside the call, or a bounded async hand-off mirroring the clipboard rung.

**The RU commit branch that must survive** (actor.go:1386-1395) — unchanged; it stays the script producer regardless of the spike verdict:
```go
	case a.mode == modeRU:
		// #nosec G115 -- printableKeyval bounds the keyval below
		// 0xFE00, so the uint32→rune conversion cannot overflow.
		r := rune(ev.Keyval)
		if ru, ok := layouts.ENToRU[r]; ok && ru != r && a.eng != nil {
			a.eng.CommitText(engine.NewIBusText(string(ru)))
			a.buf.Push(ru) // the committed rune is what the field now holds

			return true
		}
```

---

### `internal/session/actor_test.go` (test, event-driven)

**Analog:** itself. `fakeSink` (actor_test.go:115-178) gains switcher-recording in the same style (mutex-guarded `switches []string` + ops entry + hook); `TestActor_FlipOnSingle` (actor_test.go:1112) and `TestActor_FlipEmitsPanelSymbol` (actor_test.go:1151, which pins the emit-after-log order via `modeHook`) are the templates for the seam-order test: mode record → engine-name switch call → UpdateModeSymbol, strictly ordered.

---

### `internal/install/install.go` (service, batch + transform) — takeoverSources → wrapSources

**Analog:** itself, four precedents compose into the wrap:

**The value being replaced** (install.go:106-108):
```go
const (
	ownerSourcesSet        = "[('ibus', 'goswitch-en')]"
	fallbackSources        = "[('xkb', 'us')]"
```
`fallbackSources` and the whole verbatim-restore path stay unchanged.

**The takeover write site** (install.go:631-639) — becomes the computed wrap:
```go
// takeoverSources writes the D-40 single-owner value.
func (i *Installer) takeoverSources(ctx context.Context) error {
	_, err := i.call(ctx, binGSettings, "set", gsettingsSchema, gsettingsKey, ownerSourcesSet)
	if err != nil {
		return fmt.Errorf("set input sources to %s: %w", ownerSourcesSet, err)
	}

	return nil
}
```

**Strict-parse discipline to reuse, not fork** (internal/activate/activate.go:153-175, `tupleRe` at :66) — the closed-enum wrap parses the user's sources with `tupleRe = regexp.MustCompile(`\('([^']*)',\s*'([^']*)'\)`)` + full-remainder check, maps `('xkb','us')→('ibus','goswitch-en')`, `('xkb','ru')→('ibus','goswitch-ru')`, renders from Go values via fmt — the gsettings argument is never raw user text (ASVS V5, GVariant-injection class). Research recommends extracting/sharing the parser (one canonical parser, "Don't Hand-Roll" table) — the daemon must not import `internal/install`, the `activate` precedent of a local `cmdTimeout` (activate.go:45-48) shows how the packages already mirror rather than import when needed; the planner picks extract-to-shared vs mirror.

**First-backup-is-sacred** — the wrap input resolution (live xkb pair → saved original → refusal, Pitfall 7) hooks here (install.go:434-476):
```go
	path := i.path(stateDirRel, stateFile)
	if _, err := os.Stat(path); err == nil {
		return prior, nil // idempotent backup: the original state stays
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat state file: %w", err)
	}
```

**Shape-checked restore** (install.go:703-723 + savedSources 918-933) — untouched; `derivedEngine` (install.go:938-955) is the positional-preservation/best-effort-activation precedent for the wrap's index handling.

**Wire identity mirror** — `wireEngines` (install.go:886-891) and `cmd/goswitchd`'s `engineConfig` (main.go:162-165) must keep byte-mirroring; `TestInstall_ComponentXMLMirrorsWireIdentity` (install_test.go:335) enforces it:
```go
func wireEngines() []engine.EngineDesc {
	return []engine.EngineDesc{
		engine.NewEngineDesc("goswitch-en", "goswitch English (US)", "en", "us", "en"),
		engine.NewEngineDesc("goswitch-ru", "goswitch Русская", "ru", "ru", "ru"),
	}
}
```
`Layout` ("us"/"ru") and `Symbol` ("en"/"ru") are exactly the fields GNOME consumes (types.go:32-52; field order is the wire contract — types.go:5-8).

**Report + install sequence** — `Install` (install.go:302-349): the wrap replaces the constant inside `takeoverSources`, everything else (saveState → ... → takeoverSources → clearSwitchBinding → activateEngine(engineEN)) keeps its order; the explicit `activateEngine` after the write is mandatory (Pitfall 6 — value-identical writes unset the global engine). `report` (install.go:672-685) line "sources: single owner ..." becomes the two-source verdict.

**Static errors** (install.go:118-133) — new refusals follow the err113 style: name the failing step and the fix; refusal messages name the failing tuple TYPE only (D-20-safe).

---

### `internal/install/install_test.go` (test, batch)

**Analog:** itself. `fakeRunner` (install_test.go:76) drives the wrap corpus (pair wrap, positional preservation, third-source passthrough, refusal table, already-wrapped upgrade, verbatim-restore unchanged — the Wave 0 gap list). `TestInstall_Sequence` (install_test.go:232) pins the step order; `TestInstall_SecondInstallKeepsOriginalBackup` (install_test.go:482) is the direct template for the Pitfall-7 upgrade case; `TestInstall_OverInstallIdempotent` (install_test.go:731) for the re-wrap idempotence.

---

### `internal/install/selfcheck.go` (service, request-response) — two-source assertion

**Analog:** itself. `checkInputSource` (selfcheck.go:174-186):
```go
func (i *Installer) checkInputSource(ctx context.Context) (string, error) {
	out, err := i.call(ctx, binGSettings, "get", gsettingsSchema, gsettingsKey)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errSourceNotOwner, err)
	}
	if !strings.Contains(string(out), "('ibus', '"+engineEN+"')") {
		return "", errSourceNotOwner
	}

	return "input-source", nil
}
```
Widen: both `('ibus', 'goswitch-en')` and `('ibus', 'goswitch-ru')` present; reword `errSourceNotOwner` (selfcheck.go:43-44) to state the two-source expectation (red verdict names the fix — the D-41 contract, selfcheck.go:26-45). The step loop (selfcheck.go:65-91) and its fail-fast/one-verdict-per-line rendering stay as-is.

---

### `internal/activate/activate.go` (service, request-response) — engine-name derivation fix

**Analog:** itself. `IfOwned` (activate.go:101-142) — the defect is its input derivation (reads dead `current` key, Pitfall 3):
```go
	names := parseSources(sources)
	index, err := parseCurrent(current)
	...
	if !strings.HasPrefix(names[index], enginePrefix) {
		slog.Debug("engine reactivation skipped", "reason", "current input source is foreign",
			"engine", names[index])

		return
	}

	reactivate(ctx, run, names[index])
```
Fix per research: prefer the actual engine (`GetGlobalEngine` on the daemon's ibus connection when the bus is live) over the current-index derivation; keep the index path only as the cold-bus fallback. The void-contract (never returns error, never panics, DEBUG/WARN degradation) and the retry loop `reactivate` (activate.go:202-228) with its test-knob globals (activate.go:58-62, `//nolint:gochecknoglobals // test seam` comments) stay. Package boundary note (activate.go:45-48): the daemon-side `GetGlobalEngine` must reach the engine connection — same generation-scoped seam problem as BindSwitcher; consider one seam serving both (planner's call).

---

### `internal/activate/activate_test.go` (test, request-response)

**Analog:** itself. `fakeRunner` (activate_test.go:39) + the `TestIfOwned*` corpus (activate_test.go:163-292: activates-owned, foreign-skip, malformed-skip, gsettings-failure-skip, retry, cancelled-ctx). The derivation corpus adds: GetGlobalEngine path live-engine wins, cold-bus falls back to index, ru-active-not-reverted (Pitfall 3 warning sign).

---

### `cmd/goswitchd/main.go` (config, wiring) — BindSwitcher → actor

**Analog:** itself. `engineConfig`'s PostRegister closure is the exact wiring template (main.go:161-175):
```go
func engineConfig(actor *session.Actor) engine.Config {
	engines := []engine.EngineDesc{
		engine.NewEngineDesc("goswitch-en", "goswitch English (US)", "en", "us", "en"),
		engine.NewEngineDesc("goswitch-ru", "goswitch Русская", "ru", "ru", "ru"),
	}

	return engine.Config{
		Component: engine.NewComponent(engines),
		Engines:   engines,
		Handler:   actor, // decides consumption (RU script mode) at the key; tap decisions at window expiry.
		PostRegister: func(ctx context.Context, _ int) {
			activate.IfOwned(ctx, activate.NewExecRunner())
		},
	}
}
```
`BindSwitcher` lands beside `PostRegister` here: the closure receives the generation-scoped switcher and hands it to the actor (a `SetSwitcher`-style actor setter mirroring `AttachEngine` at actor.go:565-570, or a field set before `engine.Run` at main.go:76). `newActor` (main.go:125-148) is the WARN-not-fatal wiring precedent (the chord-parse degradation, main.go:133-136) — a nil switcher degrades to internal flip with a WARN, never a start failure.

---

### `docs/adr/ADR-006-two-engine-revision.md` (documentation, n/a) — NEW

**Analog:** `docs/adr/ADR-001-layout-switching-mechanism.md` (102 lines). Format to copy: `## Status` (owner-gate state + date) → `## Context` (verbatim decision quotes as blockquotes — D-52/D-53/D-54 from 05-CONTEXT.md) → `## Decision` (bold verdict line + evidence table: candidate | observed outcome, ADR-001:52-61 kill-table shape) → `## Consequences` (bulleted user-contract items incl. e2e-oracle implications, ADR-001:65-89) → `## Reversibility` (reversible/costly quotes, ADR-001:94-102). ADR-006 records Option B → Option A revision with the 2026-09-27/28 live findings; draft AFTER the spike verdict (research Open Question 5), in the plan that lands the chosen mechanism. Evidence-vehicle companion: `docs/adr/d01-experiment-log.md` style if the spike appends a journal.

---

### `test/e2e/` switch-case extension (e2e test, batch) — modify `case_combo.go` / new case file

**Analog:** `runLayoutSingle` (test/e2e/case_combo.go:284-320) — the journal-oracle template:
```go
	for _, tc := range []struct {
		tap  string
		mark string
	}{
		{"first tap flips EN→RU", `"msg":"mode","to":"ru"`},
		{"second tap flips RU→EN", `"msg":"mode","to":"en"`},
	} {
		base := s.countSub(tc.mark)
		if err := s.injectKeys(ctx, "Shift_R"); err != nil {
			return err
		}
		if err := s.waitForNew(ctx, tc.mark, base+1, decisionWait); err != nil {
			return fmt.Errorf("layout-single %s: %w", tc.tap, err)
		}
	}
```
Contract: the `"msg":"mode"` marks stay byte-stable; the new `SetGlobalEngine` journal record (engine NAME only — "goswitch-en"/"goswitch-ru", D-20) becomes an additional mark. `runSuperSpaceAlive` (case_combo.go:347-379) keeps the D-34 bare-Super form — never inject Super+Space (Pitfall 8). Drive two-source flips through injected Shift_R taps (as here) or the daemon seam; the indicator-click sync case is owner-manual unless a probe is found. Registration: the case registry map (test/e2e/main.go:240-264) + the usage string (main.go:108, 267-272) + a mise task (`[tasks.e2e-layout-single]` at mise.toml:125-127 is the template; live cases are deliberately NOT in `mise run ci`).

---

### `test/e2e/main.go` (e2e harness, batch) — registry + spike's snapshot/restore model

**Analog:** itself. `snapshotDesktop`/`restoreGsettings`/`verifyRestored`/`restoreEngine` (main.go:473-558) already model the spike's reversible window — including the write-only-on-diff rule (main.go:507-509, the value-identical-write pitfall) and the never-restore-to-goswitch guard (main.go:545-551). New cases register at main.go:240-264.

---

### `test/e2e/matrix.go` (e2e config, batch)

**Analog:** itself. `matrixCase` carries the switch oracle today (`Mode` field + the mode-record-after-correction-done ordering note, matrix.go:259-284); validation vocabulary checks (matrix.go:339-341, 378-381) are the extension points for two-source case rows.

---

### Wave-1 spike script/checklist (spike, batch) — NEW

**Analog:** `test/e2e/case_d01.go`. Copy its probe-matrix skeleton (case_d01.go:52-77): reset-to-baseline before/after each probe, per-probe verdict appended to a journal, observable-only criteria (journal focus pairs — the focus-record constants at case_d01.go:30-34 — plus typed-script truth `ghbdtn`→Cyrillic, NEVER dconf values). It even carries the two-source list constant already (case_d01.go:22: `goswitchSourcesPrefix = "[('ibus', 'goswitch-en'), ('ibus', 'goswitch-ru')]"`). The spike additionally observes the indicator under owner eyes (manual gate, SWCH-03) and restores via the stand's snapshot discipline.

## Shared Patterns

### Generation-scoped seams on the ibus connection
**Source:** `engine/conn.go:31-44` (PostRegister contract), `conn.go:74-137` (serve cycle)
**Apply to:** BindSwitcher, any GetGlobalEngine/GlobalEngineChanged work in engine.
The connection dies with every ibus-daemon restart; anything handed upward must be rebound per generation inside `serve`, after `RequestName` succeeds. nil = no-op. Cheap and non-fatal.

### Logging discipline (D-20/D-21)
**Source:** `internal/session/actor.go:369-373` (Status doc), `engine/engine.go:161-165` (DEBUG key trace), selfcheck.go:63-64
**Apply to:** every new record — the SetGlobalEngine record logs the engine NAME (a config-level literal) only; counts and states, never typed/corrected text; one-line verdicts flattened (selfcheck.go:79-83).

### Static errors with fix hints (err113)
**Source:** `internal/install/install.go:118-133`, `internal/install/selfcheck.go:29-45`
**Apply to:** new wrap refusals and selfcheck verdicts — name the failing step, the failing tuple type, and the fix.

### Subprocess/bus calls carry a deadline; nothing blocking under the actor mutex
**Source:** `internal/activate/activate.go:48` (cmdTimeout), `internal/install/install.go:38-42`, actor.go:85 (rungMu, WR-01), actor.go:548-552 (armed-payload handoff)
**Apply to:** the SetGlobalEngine call (Pitfall 4) — deadline ctx well under 50 ms, or bounded off-mutex hand-off.

### Journal records as the e2e oracle
**Source:** `internal/session/actor.go:1294,1297` (`slog.Info("mode", "to", ...)`), `test/e2e/case_combo.go:303-317` (countSub/waitForNew)
**Apply to:** all switch cases — byte-stable `"msg":"mode","to":"ru|en"` marks; new records additive only.

### Runner/fake seams at the point of use
**Source:** `internal/activate/activate.go:69-93` (Runner + NewExecRunner), `internal/install/install.go:135-156`, actor_test.go:115-178 (fakeSink), install_test.go:76, activate_test.go:39
**Apply to:** all new unit corpora — record argv/args/name as the observable surface; test knobs as package vars with `//nolint:gochecknoglobals // test seam` + t.Cleanup restore.

### TDD + green-iteration gates
**Source:** CONVENTIONS.md directives 1-3; go-ultimate SKILL.md (verification block)
**Apply to:** every task — red → green → refactor; `mise run ci` (build + vet + golangci-lint + test -race) green per commit; zero new dependencies (tidy-diff gate).

## No Analog Found

No file lacks a same-role analog. Two mechanisms are nonetheless novel in-repo (planner: treat as spike-gated, assumption-tagged work, not copy-paste):

| Mechanism | Novelty | Nearest precedent (partial) |
|-----------|---------|---------------------------|
| `AddMatchSignal` match-rule subscription + signal dispatch in `engine` | only the blind drain loop exists today (conn.go:143-156); nobody dispatches signals by name | `waitBusLoss` channel registration; assumption A2 (AddMatch works on the ibus socket) must be spike-proven |
| In-daemon `SetGlobalEngine` caller | no daemon-side ibus method call exists except RegisterComponent | `ibus.CallWithContext` shape at conn.go:122-126; subprocess callers `activate.reactivate` (activate.go:210) and `install.activateEngine` (install.go:662-668) are the rejected-by-D-52 contrast |

## Metadata

**Analog search scope:** entire module tree (`engine/`, `internal/`, `cmd/`, `test/e2e/`, `docs/adr/`, `mise.toml`) — 62 tracked .go files; `.zcode/skills/go-ultimate` read for conventions only (untracked, never an analog)
**Tracked-source gate:** every analog path verified via `git ls-files -- <path>` (all non-empty); `.gsd/`, `.zcode/`, `dist/` excluded
**Files deep-read:** 11 (conn.go, types.go, engine.go, activate.go, actor.go targeted, install.go targeted, selfcheck.go, main.go goswitchd, ADR-001, case_combo.go targeted, case_d01.go targeted) + targeted greps in actor_test.go, install_test.go, activate_test.go, wire_test.go, matrix.go, test/e2e/main.go, mise.toml
**Pattern extraction date:** 2026-09-28
