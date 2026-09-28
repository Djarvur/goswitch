# Phase 5: Интеграция с GNOME: индикация и двухисточниковое переключение - Research

**Researched:** 2026-09-28
**Domain:** GNOME 46 input-source integration / IBus SetGlobalEngine switching / install sources-takeover rewrite
**Confidence:** MEDIUM (mechanism analysis is code-verified HIGH, but the central design expectation — indicator follows an external SetGlobalEngine — is contradicted by the GNOME Shell 46 source and MUST be settled by a live tracer spike before dependent code; see Critical Finding)

## Summary

Phase 5 turns ADR-001 Option B (internal flip, invisible to GNOME) into Option A (two goswitch input sources, flip via `org.freedesktop.IBus.SetGlobalEngine` on the daemon's own IBus connection). Three locked owner decisions govern the phase (D-52 two-source scheme, D-53 sources must be `('ibus',…)`, D-54 install wraps the user's chosen pair instead of hardcoding). The mechanical surface is small and already exists: both `goswitch-en`/`goswitch-ru` engines are registered on the wire today (`ibus list-engine` shows both on this desktop, live-checked 2026-09-28), `EngineDesc.Layout` ("us"/"ru") and `Symbol` ("en"/"ru") are the exact fields GNOME Shell consumes, and `activate.IfOwned` already reactivates the engine at the current sources index — the generalization D-52 needs is 90% present.

The load-bearing finding of this research is negative and must shape the plan's first task: **GNOME Shell 46 does not follow an external `SetGlobalEngine`.** Verified from the upstream 46.0 sources of `js/misc/ibusManager.js` and `js/ui/status/keyboard.js` (fetched from gitlab.gnome.org, tag 46.0): the shell's `InputSourceManager` activates a source (applies its XKB layout, updates the panel indicator, updates MRU) ONLY from user gestures, per-window focus (off by default), password-field reloads, or a gsettings `sources` rewrite — and the last one re-activates the shell's internal MRU[0], not a daemon-chosen source. `ibusManager._engineChanged` merely records the engine name for panel-property attribution. This independently explains ADR-001's live kill-criterion (XKB group did not follow `SetGlobalEngine`) and extends it: the panel indicator does not follow either. The plan must open with a live two-source spike that observes the flip with owner eyes, and carry an explicit fallback ladder if the indicator stays stale (below, "Mechanism resolution").

Everything else is conventional, in-repo work with zero new dependencies: the wrap-takeover in `internal/install` (strict parse of the user's sources, closed-enum mapping, ASVS verbatim restore untouched), the switcher seam from `engine.Config` down to `flipScript`/`settleCorrectionFlip`/`modeSwitchChord`, a sync listener (`GlobalEngineChanged` + FocusIn engine name) instead of the unmaintained `current` gsettings key, selfcheck's two-source assertion, and e2e switch cases that keep the `"msg":"mode"` journal contract and add the `SetGlobalEngine` record.

**Primary recommendation:** Plan Wave 1 as a live tracer spike (two-source install + manual `ibus engine goswitch-ru` flip + journal/indicator/XKB observation) whose verdict picks between "D-52 as designed" and the fallback ladder; only then commit the daemon-side flip seam, and keep the RU-commit machinery (modeRU branch) regardless of the outcome — it is the engine-side safety net that makes the input script correct even when the XKB group does not follow.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-52 (2026-09-28): двухисточниковая схема.** sources = два goswitch-движка, флип = `SetGlobalEngine` на шине IBus (прямой D-Bus call, не subprocess), индикатор нативный. Это ревизия ADR-001 (Option B → Option A) — оформить как ADR-006 с живыми находками как обоснованием.
- **D-53 (2026-09-28): источники обязаны быть `('ibus',...)`.** xkb-источники обрабатываются mutter/XKB МИМО ibus-движков — демон не видит ни одной клавиши. Доказано живьём: при отвале движка GNOME переписал sources в xkb и ввод молча шёл мимо goswitch.
- **D-54 (2026-09-28): install НЕ хардкодит пару en/ru.** Он читает текущие sources ПОЛЬЗОВАТЕЛЯ и оборачивает каждый источник пары в goswitch-движок его раскладки: `('xkb','us')→('ibus','goswitch-en')`, `('xkb','ru')→('ibus','goswitch-ru')`. XKB-layout лежит внутри EngineDesc каждого движка (механизм уже есть). «Раскладку выбирает пользователь» — install лишь оборачивает выбор.
- v1 оборачивает ТОЛЬКО пару us↔ru (таблицы конвертации только ЙЦУКЕН↔QWERTY): другая пара — честный отказ/предупреждение. Произвольные пары из xkb-данных — веха v1.1 (Phase 7+/milestone v1.1 вместе с автокоррекцией).
- Пользователь может сам добавить третий xkb-источник — переключение на него честно выводит из-под goswitch (ожидаемое поведение, задокументировать).
- Правило флипа после коррекции (2026-09-28, уже в коде): режим = скрипт результата; в двухисточниковой схеме это «активным становится движок скрипта результата».

### Claude's Discretion
(Not separated in CONTEXT.md — the technical notes section «Технические заметки для планировщика» is advisory, not locked.)

### Deferred Ideas (OUT OF SCOPE)
- Произвольные пары раскладок (v1.1, из xkb-данных) — Phase 7+/milestone v1.1.
- Автокоррекция — Phase 6 (зависит от этой фазы: политика жестов/режима стабильна).
- GUI настроек.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SWCH-01 | Right Shift переключает раскладку; механизм фиксируется ADR фазы (теперь two-engine) | SetGlobalEngine wire verified live; flip seam design below; ADR-006 replaces the ADR-001 Option B clause |
| SWCH-02 | Комбо «исправить слово и переключить» | `settleCombo` already routes through `flipScript` — one seam change covers all flip sites |
| SWCH-03 | Родное Super+Space и MRU; индикатор GNOME актуален | Super+Space is the goswitch `mode_switch_chord` (already in code); indicator accuracy is the Critical Finding — spike + fallback ladder |
| SWCH-04 | Тайминги single/double/triple разрешены ADR (свитч на первый тап) | unchanged; SetGlobalEngine fires at window expiry, same timing contract |
| INTEG-01 | IBus engine, commit_text | unchanged; both engines already registered (`wireEngines`, verified live via `ibus list-engine`) |
| INTEG-02 | Совместимость с keyd/xremap, транзит | unchanged — no new grab, no uinput |
| INTEG-03 | Без root поверх дефолтного IM-стека | unchanged; takeover is gsettings-only |
| INTEG-04 | Перерегистрация после рестарта ibus-daemon | `activate.IfOwned` already engine-name-based; defect found in its `current`-index input (Pitfall 3) — fix in this phase |
| INTEG-05 | Паника не роняет движок | recover shim already wraps every exported handler; the new switcher seam must stay inside handler paths or be panic-safe by contract |
| INST-01 | Установка без root | wrap-takeover replaces `ownerSourcesSet`; ASVS discipline (snapshot/shape-check/verbatim restore) carried over; selfcheck step 6 updated to the two-source assertion |

Note: REQUIREMENTS.md marks these IDs Complete from Phases 1–4; Phase 5 re-opens them under the revised architecture (the traceability table should be updated at verify time, and ADR-006 records the revision).
</phase_requirements>

## Project Constraints (from AGENTS.md / CONVENTIONS.md)

- Go ≥ 1.23 in go.mod (local toolchain go1.27.1, live-verified), stdlib + `godbus/dbus` + minimal deps — **this phase adds ZERO new packages**.
- Строгий TDD: red → green → refactor per behavior task; verify blocks run tests at every step.
- Зелёная итерация: `go build ./...`, `go test -race ./...`, `golangci-lint run` — all green, no deferred lint.
- mise, не Makefile; tools via `mise.toml [tools]` (go 1.23, golangci-lint 2, goreleaser 2.18.1).
- golangci-lint strict config from Phase 1 (`.golangci.yml`, `linters.default: all`, depguard denies `unsafe`).
- GitHub Actions stay current (pr-sanity on push/PR; dependabot weekly; scheduled govulncheck).
- Go work follows the go-ultimate skill (project skill `.zcode/skills/go-ultimate/`): no `pkg/`, interfaces at point of use, `errors.Is/As`, comments explain why, `-race` always.
- GSD workflow: all edits through GSD entry points; D-20/D-21 logging discipline (counts and states only, never user text).

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Flip execution (`SetGlobalEngine`) | daemon `engine` package (IBus socket conn) | `session.Actor` (decision sites) | Only the daemon's ibus connection can call it directly (D-52: не subprocess); the actor decides WHEN, the engine layer owns HOW |
| Mode state (EN/RU) | daemon `session.Actor` | engine lifecycle (FocusIn name) | Single-writer discipline; mode must equal the actually-active engine (sync listener corrects drift) |
| Sources takeover (wrap pair) | `internal/install` (CLI, gsettings subprocess) | — | ASVS V5/V14 restore discipline already lives there; install-time only, never daemon-side |
| Panel indicator rendering | GNOME Shell (NOT daemon-controllable) | — | Shell-only state (`_currentSource`); the daemon can influence it only via the sources list or user gestures — see Critical Finding |
| XKB group selection | GNOME Shell `activateInputSource` → mutter backend | `EngineDesc.Layout` (input) | Shell applies `is.xkbId` which for an ibus source IS the engine desc layout (code-verified) |
| Post-restart engine restore | `internal/activate` (IfOwned) | shell's own MRU[0] activation after ibus restart | ibus#2910 loses the global engine; both sides re-assert — must not fight |
| Sync observation | daemon: `GlobalEngineChanged` on ibus conn (+ FocusIn engine name) | `gsettings monitor` mru-sources (user gestures only) | ibus-daemon is the truth holder; `current`/dconf are not maintained by GNOME 46 |
| e2e switch observation | `test/e2e` stand (journal grep) | — | `"msg":"mode"` records stay the oracle; add the `SetGlobalEngine` journal record |

## Critical Finding: GNOME Shell 46 does not follow an external SetGlobalEngine

This is the phase's central risk and MUST be resolved by a live spike before dependent code is written. Evidence chain (all three legs independent):

1. **Upstream source (authoritative for GNOME 46):** fetched `js/misc/ibusManager.js` and `js/ui/status/keyboard.js` at tag 46.0 from gitlab.gnome.org [CITED: https://gitlab.gnome.org/GNOME/gnome-shell/-/raw/46.0/js/misc/ibusManager.js and .../js/ui/status/keyboard.js]:
   - `ibusManager.js:79` — `this._ibus.connect('global-engine-changed', this._engineChanged.bind(this));` and `ibusManager.js:246-268` — `_engineChanged(bus, engineName)` only sets `this._currentEngineName` and waits for the next panel `register-properties`; it emits nothing that `keyboard.js` listens to for activation. The only ibusManager signals keyboard.js consumes are `'ready'`, `'properties-registered'`, `'property-updated'`, `'set-content-type'` (keyboard.js:377-380). **No path from an engine change to `InputSource.activate()`.**
   - `keyboard.js:502-530` — `activateInputSource(is, …)` is the ONLY place that applies the XKB layout (`this._keyboardManager.apply(is.xkbId)`), tells ibus to switch (`this._ibusManager.setEngine(engine …)` → `set_global_engine_async`), and drives `_currentInputSourceChanged` → the indicator. It is reached only from: menu click / switcher popup / `_modifiersSwitcher` (user gestures), `_inputSourcesChanged` (gsettings sources rewrite → `this._mruSources[0].activate(false)`, keyboard.js:662-664), password-field `reload()` (keyboard.js:680-703), per-window focus with `perWindow` on (off by default), and post-ibus-restart reload.
   - `keyboard.js:71-77` — `InputSource._getXkbId()` returns the ENGINE desc's layout (`engineDesc.layout` + variant) — this is why `EngineDesc.Layout` "us"/"ru" is the XKB carrier.
   - `keyboard.js:644-649` — `_makeEngineShortName(engineDesc)` returns `engineDesc.get_symbol()` — the indicator label for an ibus source IS the EngineDesc Symbol ("en"/"ru" already set).
   - `keyboard.js:941-973` — `_currentSourceChanged`: indicator hidden when `nVisibleSources < 2 && !newSource.properties` — the owner's live finding "индикатор виден только при 2+ источниках"; the visible label is `_indicatorLabels[index]` = the believed-current source's shortName.
2. **Live experiment on this exact desktop (Phase 1):** ADR-001 kill-table row, verbatim [VERIFIED: docs/adr/ADR-001-layout-switching-mechanism.md:59]: «IBus `SetGlobalEngine` (`ibus engine goswitch-ru`) | **Решающая находка:** пара фокусов движка исполняется (focus_out(en)+1, focus_in(ru)+1), но XKB-группа за внешним SetGlobalEngine НЕ следует — последующий ввод `ghbdtn` остаётся латиницей. keyboard.js 46 применяет раскладку только из собственного `InputSource.activate()`.» The engine-switch half works; the desktop-state half does not.
3. **Ubuntu delta check:** the noble patch queue for gnome-shell 46.0-0ubuntu6 contains no engine-follow patch — the keyboard-prefixed patches touch the on-screen keyboard (`js/ui/keyboard.js`, a different file), plus `status-keyboard-Preserve-MRU-order-while-IBus-is-disabled.patch` (MRU preservation during password-field IBus disable, not engine sync) [CITED: https://git.launchpad.net/ubuntu/+source/gnome-shell/tree/debian/patches?h=applied/ubuntu/noble-updates]. The shipped JS is ELF-embedded on Ubuntu (not loose files), so byte-level local verification was not possible — the upstream-46.0 + patch-queue basis is tagged [ASSUMED] where it matters (A1).

**Consequences for the design (all code-derived, MEDIUM-HIGH):**
- A daemon `SetGlobalEngine(goswitch-ru)` flips the ENGINE (keys flow to the ru engine) but: the indicator keeps the old label, the XKB group stays `us` (keyvals arrive Latin), the shell's believed-current source and MRU are untouched.
- The shell re-asserts its believed source (MRU[0]) on: ibus-daemon restart, password-field focus+leave (`reload()`), and any sources rewrite — each of these REVERTS a daemon-made flip the daemon does not re-assert.
- With the XKB group stuck on `us`, the ru engine receives LATIN keyvals — the existing `modeRU` commit branch (actor.go:1386-1401) is what makes the typed script correct. **Keep that machinery in either outcome**; under the two-engine scheme it remains the script producer unless the XKB group provably follows.

**Mechanism resolution (the Wave-1 spike):** on the live desk, in a reversible window: snapshot sources → write the two-source list (the new takeover value) → observe the shell's own MRU[0] activation → flip via the daemon's seam (or manually `ibus engine goswitch-ru` for the probe) → observe (a) journal focus pairs, (b) indicator under owner eyes, (c) XKB truth (type `ghbdtn` into a field: Cyrillic arriving natively = XKB followed; Latin = it did not), (d) `ibus engine` readback. Restore the snapshot afterwards (the e2e stand's snapshot/restore discipline already models this — test/e2e/main.go:469-544).

**Fallback ladder if the indicator does not follow (planner brings the verdict to the owner; possibly a small spec-delta / ADR-006 wording adjustment):**
1. **Sources-rewrite flip:** flip = write gsettings `sources` re-ordered with the target engine tuple first (a value change fires `_inputSourcesChanged` → shell activates MRU[0] → indicator+XKB+engine all follow). Risk (code-verified): the shell activates its internal **MRU[0]**, not index 0 — after the user has interactively switched both ways (MRU length ≥ 2), MRU order sticks and the rewrite activates the OLD source. Test exactly this in the spike; the live 01-03 finding (test/e2e/main.go:489-494, «A value-identical `gsettings set` still triggers GNOME's input-source re-evaluation, which live-unsets the global engine») makes even same-value writes disruptive.
2. **Engine-truth flip (D-52 literal) + accepted indicator semantics:** the flip is engine-level only; the indicator reflects the shell's believed source and changes on user indicator clicks (which the daemon FOLLOWS via GlobalEngineChanged). Criterion 2's «меняющийся при каждом переключении» is then satisfied only for user-driven switches — needs the owner's explicit re-wording.
3. **Login-restore seeding:** `mru-sources` gsettings IS read by the shell when its internal MRU is empty (session start, keyboard.js `_updateMruSources`) — the daemon/install can seed `[target, other]` there so the SESSION STARTS on the last active engine. This complements (never replaces) the live flip; races with the shell's own `_updateMruSettings` writes on interactive switches.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/godbus/dbus/v5` | v5.2.2 (already in go.mod) | `SetGlobalEngine` call, `GlobalEngineChanged` subscription on the daemon's ibus connection | Project constraint; API verified from the module cache: `func (conn *Conn) AddMatchSignal(options ...MatchOption) error` (conn.go:648), `AddMatchSignalContext` (conn.go:653), `func (conn *Conn) Signal(ch chan<- *Signal)` (conn.go:688), `WithMatchInterface` (match.go:41), `WithMatchObjectPath` (match.go:51) [VERIFIED: module cache ~/go/pkg/mod/github.com/godbus/dbus/v5@v5.2.2/conn.go:646-694, match.go:35-53] |
| `os/exec` gsettings subprocess (existing) | stdlib | install-time sources read/wrap/write, uninstall restore | STACK-sanctioned; robust against dconf backend details |
| In-repo `engine` / `internal/install` / `internal/activate` / `internal/session` | current tree | all phase work | no new imports anywhere |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `gsettings monitor` subprocess | stdlib | optional user-gesture observer (mru-sources) | only if the spike shows GlobalEngineChanged misses indicator clicks — note clicks DO change the engine, so the signal suffices |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `SetGlobalEngine` D-Bus call on daemon conn | `ibus engine` subprocess (as install/activate do) | rejected by D-52 — subprocess in the <50 ms flip path; the CLI also lacks a deadline knob |
| `GlobalEngineChanged` signal listener | polling `GetGlobalEngine` | polling hides latency and wastes the bus; the signal is the event truth (and the shell itself consumes it the same way) |
| sources-rewrite flip (fallback 1) | uinput chord injection | forbidden by SPEC §3.1 axiom (no uinput in the daemon); ydotool cannot inject Super+Space anyway (03-04 live finding) |

**Installation:**
```bash
# Nothing to install — zero new dependencies. go.mod/go.sum must not change (tidy-diff gate).
```

**Version verification:** no new packages; existing module set verified by `mise run tidy-diff` in CI.

## Package Legitimacy Audit

**No new packages are installed by this phase.** All work extends in-repo packages over the existing dependency set (godbus v5.2.2, yaml.v3, fsnotify). The audit table is therefore empty by construction.

**Packages removed due to SLOP verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### System Architecture Diagram

```
            INSTALL TIME (goswitchctl install)                RUNTIME (goswitchd)
 ┌───────────────────────────────────────────┐   ┌─────────────────────────────────────────────┐
 │ gsettings get sources ──► wrap pair       │   │ tap/chord/combo/correction                   │
 │  ('xkb','us') → ('ibus','goswitch-en')    │   │      │ (actor: flipScript / settleCombo /     │
 │  ('xkb','ru') → ('ibus','goswitch-ru')    │   │      ▼  settleCorrectionFlip / chord)       │
 │ no us/ru pair ─► honest refusal           │   │ switcher seam (engine.Config.BindSwitcher)  │
 │ snapshot originals ► install-state.json   │   │      │                                      │
 │ gsettings set sources (wrapped list) ─────┼─► │      ▼ SetGlobalEngine("goswitch-ru")       │
 │ clear wm.keybindings chords (already code)│   │ ibus-daemon ──► engine switch               │
 └───────────────────────────────────────────┘   │   ├─ FocusOut(en)/FocusIn(ru) → actor sync  │
                                                  │   └─ GlobalEngineChanged(s) → actor sync    │
                                                  │ GNOME Shell (NOT daemon-controlled):        │
                                                  │   indicator label = EngineDesc.Symbol of    │
                                                  │     the SHELL's believed source             │
                                                  │   XKB group = EngineDesc.Layout, applied    │
                                                  │     only via InputSource.activate()         │
                                                  └─────────────────────────────────────────────┘
```

Data flow of one flip: actor decision → switcher seam → ibus-daemon → (engine switch; focus pair) → daemon sync listener pulls the mode back under the actual engine. The GNOME Shell box is drawn outside the daemon's control boundary deliberately — that boundary is the phase's central fact.

### Recommended Project Structure
```
engine/            conn.go: Config.BindSwitcher seam; serve hands a generation-scoped
                   switcher closure; optional GlobalEngineChanged subscription
internal/session/  actor.go: flip sites call the switcher seam; mode sync from
                   lifecycle engine identity; Status carries the active engine
internal/install/  install.go: takeoverSources → wrapSources (pair wrap, refusal);
                   selfcheck.go: two-source assertion
internal/activate/ activate.go: engine-name derivation fix (GetGlobalEngine over `current`)
cmd/goswitchd/     main.go: wires BindSwitcher → actor
docs/adr/          ADR-006-two-engine-revision.md (format per ADR-001)
test/e2e/          case_combo.go / matrix: switch cases + SetGlobalEngine journal record
```

### Pattern 1: The switcher seam (BindSwitcher)
**What:** the daemon's ibus connection lives inside `serve` (engine/conn.go:74-137) and is re-created every generation; the actor must call `SetGlobalEngine` without owning the connection.
**When to use:** every flip site (flipScript, settleCombo path, settleCorrectionFlip, modeSwitchChord) and install-style activation checks.
**Example:**
```go
// engine/conn.go — Config gains the seam beside PostRegister
type Config struct {
    Component Component
    Engines   []EngineDesc
    Handler   EventHandler
    PostRegister func(ctx context.Context, generation int)
    // BindSwitcher receives, per connection generation, a closure that calls
    // org.freedesktop.IBus.SetGlobalEngine with the caller's deadline. nil = no-op.
    BindSwitcher func(switch func(ctx context.Context, engineName string) error)
}
// serve, after RequestName succeeds:
if cfg.BindSwitcher != nil {
    ibus := conn.Object(ibusService, ibusPath)
    cfg.BindSwitcher(func(ctx context.Context, name string) error {
        return ibus.CallWithContext(ctx, ibusService+".SetGlobalEngine", 0, name).Err
    })
}
```
Wire signature verified live: `SetGlobalEngine(in s engine_name)` [VERIFIED: live gdbus introspection of org.freedesktop.IBus at /org/freedesktop/IBus, 2026-09-28]. The existing call shape `ibus.CallWithContext(ctx, ibusService+".RegisterComponent", 0, …)` (engine/conn.go:122-126) is the in-repo precedent.

### Pattern 2: Wrap takeover with closed-enum mapping (ASVS V5)
**What:** the user's `sources` value is untrusted external input; the wrap builds the output ONLY from literals selected by a validated enum, never by interpolating the raw input.
**When to use:** `takeoverSources` replacement.
**Example:**
```go
// wrapSources maps the user's xkb pair to goswitch engines, preserving positions
// and every other source. ('xkb',"us")→('ibus',"goswitch-en"),
// ('xkb',"ru")→('ibus',"goswitch-ru"); anything else in the pair slots: refusal.
func wrapSources(raw string) (string, error) {
    // strict parse first — the activate.parseSources discipline (tupleRe +
    // full-remainder check), then per-tuple closed-enum mapping, then render
    // from Go values via fmt — the gsettings argument is never raw user text.
}
```
In-repo strict-parse precedent to reuse (do not duplicate): `parseSources` in internal/activate/activate.go:153-175 with `tupleRe = regexp.MustCompile(`\('([^']*)',\s*'([^']*)'\)`)` [VERIFIED: internal/activate/activate.go:64-67,153-175].

### Pattern 3: Mode = active engine (sync, not trust)
**What:** the daemon's mode is derived from the actually-active engine, corrected by observation, never trusted from gsettings.
**When to use:** FocusIn of an engine object (carries the engine name — engine.FocusIn already logs it: `slog.Info("focus_in", "engine", e.name)` [VERIFIED: engine/engine.go:247-253]) and the GlobalEngineChanged subscription.
**Example:**
```go
// subscription on the daemon's ibus connection (per generation):
conn.AddMatchSignal(dbus.WithMatchObjectPath(ibusPath), dbus.WithMatchInterface(ibusService))
signals := make(chan *dbus.Signal, signalBufferSize)
conn.Signal(signals)
// dispatch GlobalEngineChanged(engineName) → actor.SyncEngine(engineName)
```
AddMatch works on the ibus socket because the shell itself receives this signal as an ordinary socket client via GLib subscription (ibusManager.js:74-79 with `set_watch_ibus_signal(true)`, ibusbus.h:1125-1135) — inference tagged [ASSUMED] (A2) until the spike proves delivery to the daemon's own connection. Note `waitBusLoss` currently DRAINS every signal (engine/conn.go:139-156): the subscription must coexist with (or replace) that drain loop — one serve-loop select can do both.

### Anti-Patterns to Avoid
- **Reading `current` (or any input-sources key) as runtime mode truth** — GNOME 46 never writes `current` (no `settings.current` assignment exists in keyboard.js; live value stays `uint32 0` even after engine changes). ibus `GetGlobalEngine`/`GlobalEngineChanged` is the only maintained truth.
- **Subprocess in the flip path** — `ibus engine` per flip violates D-52 and adds fork latency to the <50 ms budget.
- **Calling SetGlobalEngine synchronously under the actor mutex without a deadline** — a wedged ibus-daemon would stall every keystroke (the WR-01 clipboard-rung precedent: subprocess/bus calls run off the hot path or with a hard deadline).
- **String-building the gsettings sources argument from raw input** — GVariant injection class; map through the enum, render from literals.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| IBus engine switching | own XML/CLI plumbing per flip | `SetGlobalEngine` D-Bus method on the daemon's connection | one call, deadline-capable, the exact method GNOME itself uses (keyboard.js `_setEngine` → `set_global_engine_async`) |
| Strict GVariant sources parsing | ad-hoc Split/Trim | the `tupleRe` + remainder-check discipline already in internal/activate | proven against real gsettings output; keep one canonical parser (extract/share, don't fork it) |
| Engine re-derivation after restart | new gsettings-current logic | extend `activate.IfOwned` (already engine-name-based) | the sy8 quick task built exactly this; only its input derivation needs the fix below |
| Journal contract for switch cases | new log shapes | keep `slog.Info("mode", "to", …)`; add one `SetGlobalEngine` record | the e2e stand greps the exact `"msg":"mode","to":"ru"` forms (case_combo.go:297-320) — changing them rewrites the matrix |

**Key insight:** every wire-level primitive the phase needs already exists in the tree (two engines registered, Layout/Symbol fields, IfOwned, panel properties, switch chord). The new engineering is (1) one seam from the ibus connection to the actor, (2) the wrap in install, (3) sync listeners, (4) proof of the GNOME-side behavior.

## Runtime State Inventory

> This phase migrates live desktop state (sources takeover shape) — inventory completed 2026-09-28.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | `~/.local/share/goswitch/install-state.json` — schema `{Sources, SwitchInputSource, SwitchInputSourceBw}` [VERIFIED: internal/install/install.go:176-180]; existing installs hold the pre-single-owner backup (owner's original `[('xkb','us'), ('xkb','ru')]` per STACK) | none — schema unchanged; first-backup-is-sacred rule must survive the new takeover (install-over-install reads the saved original as wrap input when live sources are already goswitch-owned) |
| Live service config | gsettings `org.gnome.desktop.input-sources` `sources` = `[('ibus', 'goswitch-en')]` (live-read 2026-09-28), `current` = `uint32 0`, `mru-sources` = `@a(ss) []`; wm.keybindings switch chords cleared by install (restore values saved) | code edit (wrap takeover) + live migration via `goswitchctl install` re-run on the owner's desk; `mru-sources` optional seeding (design question Q2) |
| OS-registered state | systemd user unit `goswitchd.service` (running — `goswitch-en` global engine live-read); ibus registry cache carries both engines (component XML already lists goswitch-en AND goswitch-ru: `wireEngines()` [VERIFIED: internal/install/install.go:886-891]) | none — registry/unit surfaces unchanged; only the gsettings sources value changes shape |
| Secrets/env vars | none — no secrets or env names touched by this phase | none |
| Build artifacts | none — no renames; go.mod/go.sum must be byte-stable (tidy-diff gate) | none |

**Nothing found in category:** Secrets/env vars — verified by reading every file the phase touches (engine, install, activate, session, cmd/goswitchd); no new key names introduced.

## Common Pitfalls

### Pitfall 1: Assuming the shell follows the flip
**What goes wrong:** indicator stays on the old label; XKB group stays `us` (Latin keyvals under goswitch-ru); shell later re-activates its believed source (ibus restart, password-field reload, sources rewrite) and silently reverts the flip.
**Why it happens:** GNOME Shell 46 has no path from GlobalEngineChanged to InputSource activation (Critical Finding).
**How to avoid:** the Wave-1 spike decides; the fallback ladder is pre-agreed with the owner; the modeRU commit branch stays regardless.
**Warning signs:** e2e ru-mode cases pass but the owner reports a stale indicator; journal shows GlobalEngineChanged storms (daemon and shell fighting).

### Pitfall 2: MRU stickiness defeats sources-rewrite flips
**What goes wrong:** rewriting the sources list activates the shell's internal MRU[0] (keyboard.js:662-664 `this._mruSources[0].activate(false)`), which after user interaction is NOT index 0 of the written list — the "flip" re-activates the old source.
**Why it happens:** `_updateMruSources` preserves internal MRU order; mru-sources gsettings is only read when the internal list is empty (session start).
**How to avoid:** if fallback 1 is chosen, the spike must include "user clicked both sources first, then daemon flips"; consider mru-sources seeding only for login restore.
**Warning signs:** flip works on a fresh session, inverts after the owner clicks the indicator once.

### Pitfall 3: `current` gsettings is dead on GNOME 46 — `IfOwned`'s input is wrong for the ru case
**What goes wrong:** with sources=[goswitch-en, goswitch-ru] and the ru engine active, `current` still reads 0 (shell never writes it — no `settings.current` in keyboard.js 46), so `IfOwned` re-activates goswitch-en after a daemon restart, reverting the user's mode.
**Why it happens:** [VERIFIED: internal/activate/activate.go:114-141] — IfOwned derives the engine from `readKey(ctx, run, keyCurrent)` → `names[index]`.
**How to avoid:** reactivation must prefer the actual engine: `GetGlobalEngine` on the daemon's ibus connection (it owns one) when the daemon still sees a live bus, falling back to the current-index derivation only on a cold bus (where the shell's own post-restart MRU[0] activation lands on index 0 anyway today).
**Warning signs:** daemon restart flips ru→en in the journal.

### Pitfall 4: D-Bus call under the actor mutex stalls keystrokes
**What goes wrong:** `flipScript` runs under `a.mu` (timer callback path); a blocking `SetGlobalEngine` on a slow/wedged bus freezes every `ProcessKeyEvent`.
**Why it happens:** godbus `CallWithContext` blocks until reply or deadline.
**How to avoid:** hard deadline on the call (well under the 50 ms budget, e.g. 40 ms context), and/or flip via a bounded async hand-off mirroring WR-01 (clipboard rung runs off-mutex); the WARN on failure is already required by criterion 3.
**Warning signs:** key latency spikes in the perf case exactly at flips.

### Pitfall 5: The flip's own FocusOut/FocusIn resets mid-flight correction state
**What goes wrong:** SetGlobalEngine triggers FocusOut(old)/FocusIn(new) on the input contexts; the actor's FocusOut path hard-resets buffer/FSM/pending rounds (actor.go:499-519) — a flip fired from `settleCombo`/`settleCorrectionFlip` can kill the just-armed verify-after round.
**Why it happens:** engine switch ≡ context switch by construction (ADR-001 live evidence: the focus pair fires).
**How to avoid:** this is the DESIRED "переключение сбрасывает контекст" semantics (CONTEXT) — but ordering matters: the flip must happen strictly AFTER the settled records (the D-36 order already enforced for the mode record must extend to the engine call), and the sync listener must treat the flip's OWN focus pair as confirmation, not as a foreign switch (self-echo suppression or epoch tag).
**Warning signs:** verify-after timeouts exactly on combo cases in the journal.

### Pitfall 6: Value-identical gsettings writes are not no-ops
**What goes wrong:** any sources write — even with the identical value — triggers GNOME's input-source re-evaluation and live-unsets the global engine (Phase 01-03 live finding, recorded in the stand: test/e2e/main.go:489-494).
**Why it happens:** dconf/GNOME re-evaluates on write, not on diff.
**How to avoid:** the wrap takeover must be followed by the existing explicit activation step (install already ends with `activateEngine`); never use a "refresh" write as a sync mechanism.
**Warning signs:** engine unsets right after install's takeover step.

### Pitfall 7: Upgrade path from the current single-source install
**What goes wrong:** the live sources are already `[('ibus','goswitch-en')]` — a naive wrap finds no xkb pair and refuses on the owner's own desktop.
**Why it happens:** the phase-4 takeover is in force right now (live-read 2026-09-28).
**How to avoid:** wrap input resolution order: live xkb pair → saved original from install-state.json (already-goswitch-owned case, idempotent upgrade) → honest refusal. Keep the first-backup-is-sacred rule untouched.
**Warning signs:** `goswitchctl install` fails on the owner's machine after merge.

### Pitfall 8: ydotool cannot inject Super+Space (e2e)
**What goes wrong:** new switch cases try to inject the chord; ydotool 0.1.8 falls back to the first physical letter for every Super+Space spelling (03-04 live finding).
**How to avoid:** keep the D-34 form — bare Super probe + name-owner + post-chord correction; drive the two-source flip cases through injected Shift_R taps and the daemon seam, or through the indicator only in owner-manual gates.
**Warning signs:** super-space case flakes with a stray 's' typed into the field.

## Code Examples

### Current flip sites (all must route through the new seam)
```go
// Source: internal/session/actor.go:1291-1302 [VERIFIED: read this session]
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
Callers today: `ExpiryAt` Single decision (actor.go:598), `settleCombo` (actor.go:647), `settleCorrectionFlip`→`setScriptMode` (actor.go:667-689, the "mode = result script" rule: `switch scriptOf(converted) { case modeEN: a.setScriptMode(a.mode == modeRU); case modeRU: a.setScriptMode(a.mode == modeEN) }`), `modeSwitchChord` (actor.go:1433). In the two-source scheme the target engine is derived, not toggled: modeEN ↔ `"goswitch-en"`, modeRU ↔ `"goswitch-ru"` — `settleCorrectionFlip`'s SET-semantics already computes the target script; only the execution changes.

### The RU commit branch that must survive (engine-side script production)
```go
// Source: internal/session/actor.go:1386-1401 [VERIFIED: read this session]
	case a.mode == modeRU:
		r := rune(ev.Keyval)
		if ru, ok := layouts.ENToRU[r]; ok && ru != r && a.eng != nil {
			a.eng.CommitText(engine.NewIBusText(string(ru)))
			a.buf.Push(ru)
			return true
		}
```
Rationale: until the XKB group provably follows the engine (and on any desktop where it does not), the ru engine receives Latin keyvals and must commit Cyrillic itself.

### EngineDesc fields GNOME consumes
```go
// Source: engine/types.go:32-52 (excerpt) [VERIFIED: read this session]
	EngineName    string  // "goswitch-en" / "goswitch-ru".
	...
	Language      string  // "en" / "ru".
	...
	Layout        string  // "us" / "ru".
	...
	Symbol        string  // "en" / "ru".
```
`Layout` → shell `InputSource._getXkbId()` → XKB group; `Symbol` → `_makeEngineShortName` → panel label. Field order is the wire contract (do not reorder — types.go:5-8).

### The single-owner value being replaced
```go
// Source: internal/install/install.go:106-108 [VERIFIED: read this session]
const (
	ownerSourcesSet        = "[('ibus', 'goswitch-en')]"
	fallbackSources        = "[('xkb', 'us')]"
```
The takeover write site: `takeoverSources` (install.go:631-639) calls `gsettings set org.gnome.desktop.input-sources sources ownerSourcesSet`. The phase replaces the constant with the computed wrap; `fallbackSources` and the verbatim-restore path stay unchanged.

### Selfcheck assertion to widen
```go
// Source: internal/install/selfcheck.go:176-186 [VERIFIED: read this session]
	out, err := i.call(ctx, binGSettings, "get", gsettingsSchema, gsettingsKey)
	...
	if !strings.Contains(string(out), "('ibus', '"+engineEN+"')") {
		return "", errSourceNotOwner
	}
```
Two-source check: both `('ibus', 'goswitch-en')` and `('ibus', 'goswitch-ru')` present; the verdict message should state the two-source expectation (criterion 5).

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| ADR-001 Option B: internal flip, single source, XKB always 'us', indicator unteachable | Option A two-source scheme (D-52/D-53/D-54), ADR-006 to record the revision with the 2026-09-27/28 live findings | 2026-09-28 (owner decisions) | flip becomes a bus call; indicator question becomes the phase's gate; modeRU commit machinery retained as the script producer |
| Panel properties (RegisterProperties/UpdateProperty, 260927-way) as indicator hope | stays in code, declared harmless/no-render on this shell (CONTEXT: «не трогаем») | 2026-09-28 | no new panel-property work in this phase |
| `current` gsettings as activation lever (STACK-era hypothesis) | dead key on GNOME 46 — shell never writes it | verified this session from keyboard.js 46.0 | all runtime state sync moves to ibus (GetGlobalEngine/GlobalEngineChanged/FocusIn name) |

**Deprecated/outdated:**
- `gsettings set … current` at runtime: ignored by the shell runtime (Phase 1 live finding, confirmed by code — no reader in keyboard.js 46).
- AskUbuntu lore about `current` being read-only: irrelevant either way — the key is simply unmaintained.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Ubuntu's gnome-shell 46.0-0ubuntu6 behaves as upstream 46.0 for keyboard.js/ibusManager.js (patch queue checked — no engine-follow patch; shipped JS is ELF-embedded, not byte-verified locally) | Critical Finding | if Ubuntu carries an unseen behavior change, the spike falsifies it cheaply — spike stays mandatory |
| A2 | `org.freedesktop.DBus.AddMatch` works on the ibus socket for the daemon's connection (inferred: the shell receives GlobalEngineChanged as an ordinary socket client) | Pattern 3 | if not, fall back to draining signals in the serve loop (ibus-daemon may broadcast); one-line spike check |
| A3 | ibus-daemon emits GlobalEngineChanged also for switches the daemon itself initiated (self-echo) | Pattern 3, Pitfall 5 | if NOT emitted to the initiator, sync is simpler (no suppression needed); if emitted, an epoch/echo tag prevents flip loops — spike check |
| A4 | The panel renders shortName "en"/"ru" from EngineDesc.Symbol for ibus sources (code-derived; live render unverified) | Critical Finding, criteria 2 | if Symbol is ignored, Symbol/longname tuning is trivial; owner's live check at the gate |
| A5 | The spike's outcome (indicator follows or not) is unknown at research time — the fallback ladder is ready but unpicked | whole plan | the planner MUST sequence the spike before the flip-seam task commits to one mechanism |
| A6 | Engine switch latency (SetGlobalEngine roundtrip + focus pair) fits the <50 ms budget on the local socket (STACK: local D-Bus RTT 1–10 ms; the 04-07 perf window was injector-dominated) | criterion 3 | if slower, the perf table/honest-budget statement covers it; measure in the e2e switch case |

## Open Questions

1. **Does the indicator change on an external SetGlobalEngine in the two-source configuration?**
   - What we know: upstream code says no; ADR-001's live experiment falsified the XKB half on this machine; Ubuntu carries no compensating patch.
   - What's unclear: nothing code-level — but the locked D-52 expects "индикатор нативный", so the LIVE verdict on the exact target desktop decides the plan's shape.
   - Recommendation: Wave-1 tracer spike with reversible desktop mutation; pick D-52-as-is or a fallback WITH the owner before Wave 2.
2. **mru-sources seeding policy (login restore)**
   - What we know: the shell reads mru-sources only when its internal MRU is empty (session start); it also writes the key on interactive switches (keyboard.js `_updateMruSettings`).
   - What's unclear: whether the daemon should persist its last engine into mru-sources at flip time (races the shell's writes) or leave login state alone.
   - Recommendation: leave alone in v1 unless the spike shows post-restart mode loss matters to the owner; document as a known behavior.
3. **Lifecycle seam shape for engine identity**
   - What we know: `HandleLifecycle(kind engine.LifecycleKind)` carries NO engine name [VERIFIED: engine/engine.go:83-97 — `HandleLifecycle(kind LifecycleKind)`]; FocusIn logs the name but drops it at the seam.
   - What's unclear: extend the lifecycle event with the engine name (interface change rippling to test doubles) vs a separate engine-identity callback.
   - Recommendation: smallest ripple — a dedicated sync entry point fed from the subscription and from FocusIn's name, keeping EventHandler's existing signatures.
4. **Wrap input resolution on already-wrapped desktops (Pitfall 7)**
   - What we know: live sources are goswitch-owned today; install-state.json holds the original pair.
   - What's unclear: exact precedence (live pair → saved original → refuse) and whether a half-wrapped list (one goswitch + one xkb) should be re-wrapped or refused.
   - Recommendation: planner pins the table of cases as unit corpus; refusal messages name the failing tuple type (D-20-safe: types only).
5. **ADR-006 wording under a fallback outcome**
   - What we know: ADR-006 must record Option B→A with live findings; the format follows ADR-001 (Status/Context/Decision/Consequences/Reversibility).
   - What's unclear: if the spike forces fallback 2 (engine-truth flip, indicator reflects user-driven switches only), criterion 2's wording needs an owner amendment (small spec-delta).
   - Recommendation: draft ADR-006 AFTER the spike verdict, in the same plan that lands the chosen mechanism.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| ibus-daemon + ibus CLI | engine registration, SetGlobalEngine, spike | ✓ | 1.5.29-2 (dpkg; `ibus address` live-answered 2026-09-28) | — |
| gsettings | takeover/selfcheck/sync probes | ✓ | ships with GNOME 46 (live reads succeeded) | — |
| gdbus | spike introspection/verification | ✓ | ships with glib; live introspection used in this research | busctl --user |
| gnome-shell | indicator, sources handling | ✓ | 46.0-0ubuntu6~24.04.15 (dpkg) | — |
| go toolchain | all code work | ✓ | go1.27.1 linux/amd64 (module pinned go 1.23 via mise) | — |
| running goswitchd | spike baseline | ✓ | live: global engine `goswitch-en`, both engines in `ibus list-engine` | — |
| ydotool 0.1.8 | e2e injection | ✓ | distro package (03 findings apply — no Super+Space) | own uinput injector (~150 LOC, stand-only) |

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** none.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (+`-race`), project skill go-ultimate conventions (`package xxx_test`, `TestF_suffixCamelCase`) |
| Config file | `.golangci.yml` (strict v2), `mise.toml` (tasks + tools) |
| Quick run command | `mise run test` (go test -race -count=1 ./...) |
| Full suite command | `mise run ci` (build + vet + lint + test) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SWCH-01 | single Shift flips via SetGlobalEngine seam; mode record preserved | unit (seam double) + e2e layout-single | `go test -race ./internal/session/ -run TestActor_Flip` / live `mise run e2e-layout-single`-class case | ✅ actor corpus exists; seam tests Wave 0 |
| SWCH-02 | combo word+flip through the seam, D-36 order | unit | `go test -race ./internal/session/` (combo corpus) | ✅ extend |
| SWCH-03 | indicator/GNOME-side accuracy | manual (owner) + spike | live spike + owner gate | ❌ Wave 0 (spike script) |
| SWCH-04 | tap timings unchanged | unit (FSM corpus) | `go test -race ./internal/hotkey/` | ✅ |
| INTEG-01..03 | unchanged engine contracts | existing corpus | `go test -race ./engine/` | ✅ |
| INTEG-04 | reactivation after ibus restart under two sources | e2e ibus-restart + unit (activate) | `mise run e2e-ibus-restart`; `go test -race ./internal/activate/` | ✅ extend (engine-name derivation fix) |
| INTEG-05 | recover shim on any new exported path | unit | `go test -race ./engine/` | ✅ extend |
| INST-01 | wrap takeover, refusal, verbatim restore, selfcheck two-source | unit (install corpus with fake Runner) + live install-cycle | `go test -race ./internal/install/`; live install/uninstall cycle | ✅ extend (wrap corpus Wave 0) |

### Sampling Rate
- **Per task commit:** `mise run ci`
- **Per wave merge:** `mise run ci` + live e2e switch cases on the desk
- **Phase gate:** full matrix green twice in the two-source configuration (D-48 nightly gate form) before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] Spike script/checklist (two-source window + flip observation + restore) — SWCH-03/criterion 2 evidence vehicle
- [ ] `internal/install` wrap corpus: pair wrap, positional preservation, third-source passthrough, refusal table, already-wrapped upgrade, verbatim restore unchanged
- [ ] `internal/activate` engine-name derivation corpus (GetGlobalEngine path + cold-bus fallback)
- [ ] e2e: switch-case extension grepping the new `SetGlobalEngine` journal record (the `"msg":"mode"` marks stay byte-stable)
- [ ] ADR-006 skeleton (docs/adr format per ADR-001)

## Security Domain

ASVS L1 (config `security_enforcement: true`, `security_asvs_level: 1`, `security_block_on: high`).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | none added (no new surfaces) |
| V3 Session Management | no | none added |
| V4 Access Control | no | daemon bus access unchanged (ibus socket peer-cred EXTERNAL auth, engine/conn.go:177-179) |
| V5 Input Validation | **yes** | the user's gsettings `sources` value is untrusted external input: strict tuple parse (tupleRe + remainder check), closed-enum mapping ('us'/'ru' only), render from literals — never interpolate raw input into a `gsettings set` argument; refusal over guessing |
| V6 Cryptography | no | none |
| V14 Config/Files | **yes** (carried) | install-state.json unchanged discipline: 0600, atomic write, read-back verify, shape-checked restore with reported fallback |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| GVariant injection via crafted sources value (quotes/brackets in layout names) | Tampering | closed-enum wrap (Pattern 2); unparsable or non-us/ru tuples are a visible refusal, never a passthrough |
| Tampered install-state.json steering the restore | Tampering | existing savedSources/savedSwitchBindings shape checks (install.go:703-814) — unchanged |
| Flip-loop DoS (daemon vs shell re-assert war) | DoS | self-echo suppression/epoch tag (Pitfall 5); WARN-bounded re-assert attempts; no unbounded retry |
| Log disclosure of typed content | Information disclosure | D-20/D-21 discipline: the new `SetGlobalEngine` record logs the engine NAME (config-level literal, "goswitch-en"/"goswitch-ru") only — never field content |

## Sources

### Primary (HIGH confidence)
- Live introspection of the target ibus bus (2026-09-28): `SetGlobalEngine(in s engine_name)`, `GetGlobalEngine(out v desc)`, signal `GlobalEngineChanged(s engine_name)`, `ListEngines`/`ListActiveEngines`/`CurrentInputContext` — via `gdbus introspect --address $(ibus address)`
- Live desktop state reads (2026-09-28): `sources=[('ibus', 'goswitch-en')]`, `current=uint32 0`, `mru-sources=@a(ss) []`, global engine `goswitch-en`, both goswitch engines in `ibus list-engine`
- In-repo sources read this session (cited with line ranges throughout): engine/conn.go, engine/types.go, engine/engine.go, engine/factory.go, internal/session/actor.go, internal/install/install.go, internal/install/selfcheck.go, internal/activate/activate.go, cmd/goswitchd/main.go, test/e2e/main.go, test/e2e/case_combo.go, mise.toml
- godbus v5.2.2 module cache: conn.go:646-694, match.go:35-53 (AddMatchSignal/Signal/MatchOption signatures)
- /usr/include/ibus-1.0/ibusbus.h:1064-1135 (set_global_engine, set_watch_ibus_signal)
- docs/adr/ADR-001 + d01-experiment-log.md — the Phase-1 live falsification of XKB-follow

### Secondary (MEDIUM confidence)
- GNOME Shell 46.0 sources fetched from gitlab.gnome.org (js/misc/ibusManager.js:60-82,244-268,287-323; js/ui/status/keyboard.js:43-77,377-380,410-500,590-680,680-740,900-975; js/misc/keyboardManager.js entire) — upstream tag 46.0, cross-checked against the Ubuntu noble patch queue (no engine-follow patch; launchpad.net / git.launchpad.net ubuntu/+source/gnome-shell debian/patches)
- .planning/research/STACK.md — gsettings/current writability, monitor approach, SetGlobalEngine availability (reused, not re-derived)
- 260927-way / 260927-sy8 / 260927-vu8 quick-task artifacts (.planning/quick/) — panel properties, self-reactivation, flip-after-correction provenance

### Tertiary (LOW confidence)
- Web search summary claiming ibusManager "syncs the active input source" on global-engine-changed — CONTRADICTED by the primary source read; discarded (recorded here deliberately as a refuted claim)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new packages; all APIs verified from module cache and live bus
- Architecture: MEDIUM — seams are clean and in-repo, but the central mechanism (shell follow) is unsettled pending the spike
- Pitfalls: HIGH — each backed by code lines, prior live findings, or both

**Research date:** 2026-09-28
**Valid until:** 2026-10-28 (stable domain — GNOME 46 is frozen on the target; revisit only on a GNOME upgrade or an ibus-daemon update)
