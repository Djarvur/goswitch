# Architecture Research

**Domain:** IBus input-method-engine daemon for GNOME Wayland (keyboard layout corrector / switcher, Punto Switcher analog)
**Researched:** 2026-09-10
**Confidence:** HIGH (live D-Bus introspection on the actual target system, owner's prior prototypes read in full, upstream C/JS sources read directly; web-sourced claims tagged individually)

> Method note: this research was performed **on the target machine** (Ubuntu 24.04, GNOME Wayland, IBus 1.5.29-rc2, layouts `us`+`ru`, keyd installed but disabled). The IBus daemon was introspected live over its private socket; an engine was registered and activated end-to-end to verify the lifecycle; the owner's two Python prototypes (`~/.local/share/punto-switcher/punto_engine.py`, `~/.local/share/ibus-test/test_shift.py`) were read as prior art. Upstream sources (`ibus/ibus` C code, `gnome-shell` keyboard.js, `sarim/goibus`) were fetched and read. Claims below carry provenance.

---

## Standard Architecture

### System Overview — where an IBus engine sits in the Linux input stack

```
┌─────────────────────────────────────────────────────────────────────────┐
│ KERNEL LAYER                                                            │
│   physical keyboard → evdev (/dev/input/event*) → [keyd / xremap]       │
│                                        │ (optional remap, uinject out)  │
├────────────────────────────────────────┼────────────────────────────────┤
│ COMPOSITOR LAYER (mutter / gnome-shell, Wayland)                        │
│   libinput → XKB translation (per current layout) → keyval stream       │
│   input-method-v1/v2 + text-input protocols → routes keys into IME      │
├─────────────────────────────────────────────────────────────────────────┤
│ IME LAYER (IBus)                                                        │
│   GTK/Qt/Electron IM module (in-app, creates InputContext)              │
│        ↕ D-Bus (private ibus socket, session-user)                      │
│   ibus-daemon (registry, engine lifecycle, key routing)                 │
│        ↕ D-Bus (Factory.CreateEngine → engine object)                   │
│   goswitchd  ← THIS PROJECT (engine: sees keys, commits text)           │
├─────────────────────────────────────────────────────────────────────────┤
│ GNOME SESSION LAYER                                                     │
│   org.gnome.desktop.input-sources (sources / current / mru-sources)     │
│   gnome-shell activates engine + applies its declared XKB layout        │
│   goswitchd control service (session bus) ← goswitchctl CLI             │
└─────────────────────────────────────────────────────────────────────────┘
```

Key structural fact (verified by live traffic): **an IBus engine receives key events after XKB layout translation and after any keyd/xremap remap**, and injects text via D-Bus signals that ibus-daemon routes back into the focused application's IM module. It never touches evdev, so it cannot conflict with remappers (spec axiom satisfied by construction).

### goswitchd process architecture (single daemon, per spec §3.2)

```
┌──────────────────────────── goswitchd (one process) ─────────────────────────────┐
│                                                                                   │
│  engine/ (wire adapter)                    internal/* (pure logic, no D-Bus)      │
│  ┌──────────────────────────┐              ┌───────────────────────────────────┐  │
│  │ IBus conn (private sock) │  EngineEvent │ session actor (1 goroutine)       │  │
│  │  Factory.CreateEngine    ├─────────────►│  per-IC keystroke buffers         │  │
│  │  Engine objects (per IC) │◄─────────────┤  hotkey FSM (tap/debounce)        │  │
│  │  - ProcessKeyEvent→bool  │  ActionCmd   │  correction decisions             │  │
│  │  - FocusIn/Out,Reset     │              └───────┬───────────────────────────┘  │
│  │  - SetSurroundingText    │                      │ uses                          │
│  │  emits: CommitText,      │              ┌───────▼───────────┐  ┌────────────┐  │
│  │  ForwardKeyEvent,        │              │ correct/          │  │ layouts/   │  │
│  │  DeleteSurroundingText   │              │ direction+detect  │  │ EN↔RU map  │  │
│  └──────────────────────────┘              └───────────────────┘  └────────────┘  │
│  ┌──────────────────────────┐              ┌───────────────────────────────────┐  │
│  │ ctlsvc (session bus)     │◄─────────────│ config watcher (fsnotify→atomic)  │  │
│  │ org.djarvur.goswitch.Svc │  status/reload│ layoutset (gsettings subprocess)  │  │
│  └──────────────────────────┘              └───────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────────────────────────┘
        ▲ goswitchctl (separate binary, session bus client)
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| `engine/` wire adapter | Own the IBus private-socket connection; discover address; register component; export Factory + per-IC engine objects; translate D-Bus ↔ typed Go events; emit Commit/Forward/Delete signals | godbus `Dial/Auth/Hello/Export/Emit` — pattern proven by sarim/goibus (pure-Go path, read in source) |
| Session actor | Single goroutine owning ALL mutable state: per-input-context buffers, hotkey tap FSM, debounce timers, correction triggering | funnel channel; godbus dispatches handlers on arbitrary goroutines, so all events are serialized into one actor |
| `internal/hotkey` | Pure tap/combo state machine: single/double/triple RShift, Shift+RCtrl, modifier-use discrimination, configurable debounce windows | plain struct + `Feed(event, now) → []Action`; fully unit-testable against synthetic streams |
| `internal/correct` | Decide replacement scope (word/phrase/selection), direction (EN→RU / RU→EN / mixed), produce replacement strategy per client caps | pure functions over `layouts/` tables |
| `layouts/` | Positional key mapping tables (QWERTY↔ЙЦУКЕН incl. punctuation, case, digits row variants) | `go:generate` from `/usr/share/X11/xkb/symbols/{us,ru}` (see Build Order) |
| `internal/layoutset` | Read/write/observe `org.gnome.desktop.input-sources` (current, sources, mru-sources) | subprocess `gsettings get/set/monitor` (no GLib in Go; dconf D-Bus API is not a stable contract) |
| `internal/config` | YAML parse + validation + hot reload | `fsnotify` (or stdlib mtime poll to keep dep budget) → publish immutable `*Config` snapshot |
| `internal/ctlsvc` | D-Bus service on the **session bus** for goswitchctl: Status, ReloadConfig, CorrectNow | godbus, second connection in same process |
| Logging | Structured, levels, key-trace debug mode | `log/slog` (stdlib, Go ≥1.21) |

---

## The IBus Engine Lifecycle (verified on target system)

Empirically validated by registering and activating an engine during this research; all signatures below are from **live introspection of IBus 1.5.29-rc2** on this machine plus owner prototype logs.

1. **Address discovery** — read `$IBUS_ADDRESS`; else parse `~/.config/ibus/bus/<machine-id>-<display>` files (select by `WAYLAND_DISPLAY`/`DISPLAY` suffix, newest mtime). Content: `IBUS_ADDRESS=unix:path=/home/<u>/.cache/ibus/dbus-<rand>,guid=<guid>`.
2. **Connect** — `dbus.Dial(addr)` → `conn.Auth(External)` → `conn.Hello()`.
3. **Own the name** — `RequestName("org.freedesktop.IBus.goswitch", ReplaceExisting)`; result also serves as single-instance guard (AlreadyOwner → another instance lives).
4. **Export Factory** at `/org/freedesktop/IBus/Factory`, interface `org.freedesktop.IBus.Factory`:
   - `CreateEngine(s engine_name) → (o object_path)` — daemon calls this when an input context switches to one of our engines. We mint a fresh path (e.g. `/org/freedesktop/IBus/Engine/goswitch/<n>`), export the engine object there, return the path. (Verified live: factory interface introspected; prototype log shows `engine __init__` → `focus_in` after engine selection.)
5. **RegisterComponent** — `org.freedesktop.IBus.RegisterComponent(v component)` where the variant wraps the marshalled component struct (`'IBusComponent'`, attachments dict, name, description, version, license, author, homepage, exec, textdomain, engines[] of EngineDesc structs — 17-field wire form confirmed in the daemon's `Engines` property).
6. **Per-input-context engine object** — daemon then calls on it:
   - `ProcessKeyEvent(u keyval, u keycode, u state) → b handled` — keyval is XKB-translated; keycode is hardware evdev code; `state` packs modifier masks + `IBUS_RELEASE_MASK = 1<<30` for key-up (values verified in `ibustypes.h`).
   - `FocusIn() / FocusOut()`, `Enable() / Disable()`, `Reset()`
   - `SetCapabilities(u caps)` — per-client capability bitmap: PREEDIT=1, AUX=2, LOOKUP=4, FOCUS=8, PROPERTY=16, **SURROUNDING_TEXT=32**, OSK=64, **SYNC_PROCESS_KEY=128** (Chromium historically sets sync mode).
   - `SetSurroundingText(v text, u cursor_pos, u anchor_pos)` — arrives only when caps include SURROUNDING_TEXT; `anchor_pos != cursor_pos` means active selection (this is how "correct selection" should be detected — no clipboard hacks).
7. **Engine → daemon signals** (emit on own object path):
   - `CommitText(v IBusText)` — THE text injection primitive.
   - `ForwardKeyEvent(u keyval, u keycode, u state)` — inject a synthetic key into the app (used for the BackSpace fallback).
   - `DeleteSurroundingText(i offset_from_cursor, u nchars)` — **verified in ibusengine.c (main): this is now a SIGNAL with no boolean ack**; the engine is expected to update its own surrounding-text cache optimistically. Treat as fire-and-forget; older GI bindings' bool return is meaningless on 1.5.29.
   - `UpdatePreeditText(v, u, b, u mode)` (4-arg form on 1.5.29; verify arity against installed daemon in M1), `UpdateProperty(v)` for panel state.
8. **Teardown / restart** — daemon may call `org.freedesktop.IBus.Service.Destroy()`; on ibus-daemon restart the connection drops and **runtime registrations are lost** → goswitchd needs a reconnect+re-register loop with backoff (spec Q6). Component XML in `/usr/share/ibus/component/` remains valuable for GNOME discovery and as respawn hint; guard against double-spawn via the RequestName check (step 3).

Timing requirement (critical): `ProcessKeyEvent` is a synchronous call from ibus-daemon; the reply must return in single-digit milliseconds. **Debounce waits (450–550 ms tap windows) must happen after replying** — reply immediately, decide later via timers posting events back into the actor.

## The Key-Event Pipeline

```
app IM module → ibus-daemon → ProcessKeyEvent(keyval, keycode, state)
     ↓ (engine/, decode: release bit, modifier masks)
 [passthrough gate] not a tracked hotkey key? → return false  (~99.9% of keys, zero work)
     ↓ tracked key (e.g. Shift_R press/release)
 hotkey FSM: tap counting, modifier-use discrimination
     (any other key while held ⇒ it was a modifier chord, not a tap)
     ↓ single / double / triple decided by timer expiry
 buffer scope resolution (word | phrase | selection)
     word/phrase: engine's own keystroke buffer (per input context)
                 + SetSurroundingText tail as ground truth / fallback
     selection: anchor_pos≠cursor_pos from surrounding text
     ↓
 direction decision (layouts/: char-class vote, mixed-text partial)
     ↓
 replacement execution (capability-aware):
   caps & SURROUNDING_TEXT  → emit DeleteSurroundingText(-n, n); emit CommitText(fixed)
   else                     → emit ForwardKeyEvent(BackSpace)×n; emit CommitText(fixed)
     ↓
 optional layout switch (combo action): layoutset.Switch → gsettings set current
 buffer reset (Enter/Tab/Escape, FocusOut, Reset, surrounding-cursor jump)
```

Design decisions baked into this pipeline:

- **Buffer per input context.** The daemon creates one engine object per input context (per window). Buffers, tap FSM, and caps must be keyed by engine object path; `FocusOut` resets all.
- **Surrounding text is best-effort.** Chromium/Electron do not implement `delete_surrounding_text` (ibus issue #2354) and Firefox had cap regressions (ibus#2054). The replacement strategy object must be selected per client caps, with the BackSpace fallback always available. `commit_text` itself works everywhere IBus works.
- **Mouse-click buffer reset is NOT observable at the IBus layer** (engines see no pointer events). Approximate via: focus-out, Escape/Enter/Tab, and surrounding-text cursor jumps (a click that moves the caret shows as cursor_pos delta). Flag for spec-delta: exact click semantics cannot be guaranteed from the IME layer (potential AT-SPI subscription later — costs a11y bus dependency; v2 decision).

## Layout Observation and Switching — the two-engine pattern (RECOMMENDED)

Verified from gnome-shell source (`js/ui/status/keyboard.js`, main):

```js
_getXkbId() {
    const engineDesc = IBusManager.getIBusManager().getEngineDesc(this.id);
    ...
    return `${engineDesc.layout}+${engineDesc.variant}`;   // (or plain layout)
}
activateInputSource(is, interactive) {
    this._keyboardManager.apply(is.xkbId);    // 1) set XKB layout FROM ENGINE DESC
    ...
    this._ibusManager.setEngine(engine);      // 2) activate the IBus engine
}
```

Therefore: **register TWO engines in one component** — `goswitch-en` (EngineDesc.layout=`us`) and `goswitch-ru` (layout=`ru`) — and set GNOME sources to `[('ibus','goswitch-en'), ('ibus','goswitch-ru')]`. Consequences:

- Activating either source makes gnome-shell apply the correct XKB layout **and** route all keys to goswitchd. Keyvals arrive already layout-translated; the common path stays pure passthrough (`return false`).
- "Proxy Right Shift → Super+Space" becomes: consume the tap, then `gsettings set org.gnome.desktop.input-sources current N` (or use GNOME's switch-input-source keybinding semantics via the settings key). Forwarding a synthetic Super+Space to the app would NOT trigger the compositor grab — gsettings write is the correct mechanism.
- **Layout observation needs no polling**: a user pressing Super+Space manifests as FocusOut on one of our engines + FocusIn on the other — the daemon derives current layout from which engine is focused. `mru-sources` is maintained by GNOME itself; Super+Space MRU cycling keeps working.
- Selection conversion uses surrounding-text anchor instead of the prototype's `wl-paste/wl-copy` hack.

Fallback (proven by owner's prototype, keep as plan B): single engine with internal mode, translating every keystroke via per-key `commit_text`. Works, but per-key overhead, broken composing, indicator mismatch — only if the two-engine pattern hits an unexpected wall in M1 (it is the primary thing M1 must validate).

Non-Latin/Cyrillic third layouts (spec Q5): engines are only `us`/`ru`; if the user adds e.g. `de`, switching there moves focus away from our engines — goswitch simply goes dormant (keys unseen, no correction). Graceful by construction.

## Coexistence with keyd / xremap

- keyd/xremap remap at the kernel evdev/uinput layer; mutter sees remapped events; IBus receives post-remap keyvals. goswitch never opens evdev and never grabs (no EVIOCGRAB) — spec axiom holds by construction.
- Consequence for configuration: goswitch hotkeys are expressed in **post-remap** keys. If the user remaps Right-Shift in keyd, they must rebind goswitch in YAML (document this).
- e2e interaction: ydotool injects via its own uinput device; a running keyd may grab it and form feedback loops (documented class of problem, see Mouseless/keyd notes). e2e runner must ensure keyd exclusion for the ydotool device (or run the matrix with keyd disabled first, keyd-enabled variant later).

## Recommended Project Structure

Spec §8 fixes the top-level skeleton; everything else goes under `internal/`:

```
goswitch/
├── cmd/goswitchd/            # main: flag/env config path, wiring, graceful stop
├── cmd/goswitchctl/          # CLI: status | reload | correct <scope> | version
├── engine/                   # IBus wire adapter (spec-fixed location)
│   ├── address.go            # IBUS_ADDRESS / bus-file discovery + validation
│   ├── conn.go               # Dial/Auth/Hello, reconnect+reregister loop
│   ├── types.go              # Component/EngineDesc marshalling (+ XML emit for install)
│   ├── factory.go            # CreateEngine dispatch → session.NewSession
│   ├── engine.go             # per-IC object: exported methods, event funnel
│   └── keys.go               # keyval consts (Shift_R=0xFFE1…), state masks, release bit
├── internal/session/         # actor goroutine: owns buffers, FSM, caps, per-IC registry
├── internal/hotkey/          # pure tap/combo state machine
├── internal/correct/         # scope resolution + direction + replacement strategy
├── internal/layoutset/       # gsettings subprocess wrapper + monitor parsing
├── internal/config/          # YAML schema, validation, fsnotify hot reload
├── internal/ctlsvc/          # session-bus control service (org.djarvur.goswitch)
├── internal/logging/         # slog setup, key-trace mode
├── layouts/                  # generated tables + generator (go:generate; spec-fixed)
├── test/e2e/                 # ydotool + AT-SPI harness, YAML case matrix (spec-fixed)
└── docs/                     # spec, ADRs, install guide
```

### Structure Rationale

- **`engine/` is the only package allowed to import godbus for IBus purposes** — a wire adapter translating D-Bus into typed events (`EngineEvent`) and commands (`EngineAction`). Everything below it is pure Go testable headless in CI.
- **`internal/`** hides non-contract packages from external importers (Go convention; also keeps the audit surface small per the "minimal deps" constraint).
- **`layouts/` is pure data + generator**: `go:generate ./...` parses `/usr/share/X11/xkb/symbols/us` and `symbols/ru` to emit `tables.go`. This kills the hand-transcription bug class (the owner's prototype hand-wrote winkeys-style digit-row maps while the session actually uses plain `ru` — exactly the kind of mismatch generation avoids; parameterize by variant).
- **`test/e2e/` is a Go program + shell runner**, not `go test` — it needs a live GNOME session and its own exit-code contract.

## Architectural Patterns

### Pattern 1: Wire adapter over state machine

**What:** `engine/` translates D-Bus ↔ typed events and owns NOTHING stateful; `internal/session` is a single-actor state machine with no D-Bus knowledge.
**When to use:** always here — it is what makes the CI-testable fraction maximal (spec §7.1 demands headless unit/golden tests of "buffer logic, decision fixing").
**Trade-offs:** one extra indirection layer; pays for itself the moment hotkey semantics are unit-tested against recorded event streams.

### Pattern 2: Single session actor (funnel)

**What:** all engine objects push `EngineEvent` into one channel; one goroutine owns buffers/FSM/config snapshot; timers (`time.AfterFunc`) re-enter as synthetic events; replies to `ProcessKeyEvent` are computed synchronously but only ever "handled/not-handled" (cheap).
**When to use:** godbus dispatches each method call on its own goroutine — without a funnel you need locks around every buffer; with it, zero mutexes on the hot path.
**Trade-offs:** actor is a serialization point; irrelevant at human typing rates (~10–20 events/s).

```go
// engine.go (wire adapter sketch — signatures match live introspection)
func (e *Engine) ProcessKeyEvent(keyval, keycode, state uint32) (bool, *dbus.Error) {
    ev := KeyEvent{Keyval: keyval, Keycode: keycode,
        Release: state&(1<<30) != 0, Mods: state & 0x5f001fff}
    return e.session.Feed(ev), nil   // Feed replies in O(1); debounce happens inside via timers
}
```

### Pattern 3: Capability-aware replacement strategy

**What:** per input context, remember `SetCapabilities`; at correction time choose `DeleteSurroundingText+CommitText` (caps&32) or `ForwardKeyEvent(BackSpace)×n+CommitText`.
**When to use:** Chromium/Electron break surrounding-text deletion; GTK is fine. The strategy object makes the e2e matrix's per-app expectations explicit.
**Trade-offs:** BackSpace fallback visibly walks the cursor (no "jump-free" guarantee there) — acceptable, matches spec's known limitation wording.

### Pattern 4: Immutable config snapshots

**What:** watcher loads YAML, validates, publishes `atomic.Pointer[Config]`; every decision reads `cfg.Load()` once per event.
**When to use:** hot reload is a hard requirement; immutable snapshots make reload race-free without locks.

### Pattern 5: Self-healing registration

**What:** monitor connection loss (ibus-daemon restart) → reconnect with backoff → re-`RegisterComponent` → re-export factory. Combined with the RequestName single-instance guard and the installed component XML (whose exec points at the same binary) this closes spec Q6.

## Data Flow

### Correction round trip (hotkey to replaced text)

```
user: ghbdtn <tap-tap RShift>
apps' IM module ─(keys)→ ibus-daemon ─→ engine.ProcessKeyEvent (×7)
   replies: false ×5 (pass-through), true ×2 (RShift consumed)
actor: tap FSM → DoubleTap action
actor: buffer "ghbdtn" → correct.Detect → EN→RU, case-preserving → "привет"
actor: caps(surrounding)? ─ y → engine emits DeleteSurroundingText(-6, 6)
                            └ n → engine emits ForwardKeyEvent(BS)×6
engine emits CommitText("привет") → ibus-daemon → IM module → widget text
(actor updates surrounding cache; clears scope buffer)
```

### Layout switch round trip

```
actor (single RShift tap) → layoutset.Switch(next)
   → gsettings set org.gnome.desktop.input-sources current N
   → gnome-shell activateInputSource: keyboardManager.apply(us|ru) + ibus setEngine
   → daemon: our engine-A FocusOut; engine-B FocusIn   ← daemon observes = layout known
   (user-initiated Super+Space produces the identical FocusIn/FocusOut signature)
```

### Control flow

```
goswitchctl ─(session bus)→ ctlsvc: Status()/ReloadConfig()/CorrectNow(scope)
ctlsvc ─(direct call)→ session actor (same process)
```

## Performance Envelope (replaces "scaling" for a single-user daemon)

| Concern | Budget (spec: <50 ms reaction, <50 MB RSS) | Where it goes |
|---------|--------------------------------------------|---------------|
| Key delivery | ~1–5 ms | kernel→mutter→IM module→ibus-daemon→engine D-Bus hop |
| Decision | <1 ms | actor + table lookups (map[rune]rune) |
| Replacement | 5–20 ms | DeleteSurroundingText+CommitText round trip; BackSpace fallback ≈ n×2–5 ms |
| Memory | ~1–2 MB tables + small buffers | generated tables; well under budget |
| Footprint risk | none structural | single process, two D-Bus connections |

## Anti-Patterns

### Anti-Pattern 1: Blocking inside ProcessKeyEvent

**What people do:** wait out the debounce window before replying to decide single vs double tap.
**Why it's wrong:** the daemon's call is synchronous with a timeout — stalling it freezes key delivery for the whole session.
**Do this instead:** reply immediately; decide via timers that re-enter the actor (Pattern 2/3 above). The owner's prototype already worked this way (GLib timeouts).

### Anti-Pattern 2: Per-keystroke commit_text in the common path

**What people do:** translate every key in-engine and commit (prototype's ru-mode path).
**Why it's wrong:** adds latency and D-Bus traffic to 100% of typing to serve 0.1% corrections; breaks apps' own composition.
**Do this instead:** two-engine layout split; engine is a passthrough except on tracked hotkeys.

### Anti-Pattern 3: Trusting surrounding text universally

**What people do:** assume DeleteSurroundingText works because GTK apps are fine.
**Why it's wrong:** Chromium/Electron don't implement it (ibus#2354); no ack even where supported.
**Do this instead:** caps-gated strategy + BackSpace fallback + optimistic cache reconciliation from later SetSurroundingText.

### Anti-Pattern 4: Sharing one buffer across input contexts

**What people do:** module-level buffer. **Why it's wrong:** engines are per input context; cross-window bleed corrupts corrections. **Do this instead:** per-engine state keyed by object path, reset on FocusOut.

### Anti-Pattern 5: Triggering GNOME switching via forwarded Super+Space

**What people do:** `ForwardKeyEvent(Super, Space)`. **Why it's wrong:** forwarded events target the app's IM context, not compositor grabs — the switcher never fires. **Do this instead:** gsettings `current` write (or IBus GlobalShortcut APIs in later versions).

### Anti-Pattern 6: uinput / evdev anywhere in the correction path

Spec axiom — restate: injection is `commit_text` only; e2e tooling may use ydotool, the product never does.

## Integration Points

### External Services

| Service | Integration Pattern | Notes / Gotchas |
|---------|---------------------|-----------------|
| ibus-daemon (private socket) | godbus Dial+Auth+Hello; RegisterComponent; Factory+Engine exports | registrations lost on daemon restart → re-register loop; address from `~/.config/ibus/bus/` |
| GNOME input-sources (gsettings/dconf) | subprocess `gsettings get/set/monitor org.gnome.desktop.input-sources` | dconf's D-Bus write API is not a stable contract; subprocess is the pragmatic, dependency-free route; FocusIn-based observation is primary |
| session bus (control) | own name `org.djarvur.goswitch` | goswitchctl client; second godbus connection |
| systemd user unit | simple unit, `Wants=graphical-session.target` + `PartOf` | ordering after ibus-daemon; restart on failure; no root anywhere |
| keyd/xremap | none (layer separation) | hotkeys documented as post-remap; e2e must handle ydotool-device exclusion when keyd enabled |
| AT-SPI (e2e only) | python3-gi script (or Go via a11y bus) for focus + text readback | `Component.grabFocus` to activate target window per case (spec §7.3 trap) |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| engine/ ↔ session | typed `EngineEvent` chan in, `EngineAction` returns/calls out | only godbus-touching package on this side |
| session ↔ correct/layouts | direct function calls (pure) | golden-tested without any bus |
| config → all | `atomic.Pointer[Config]` snapshots | reload never mutates in place |
| ctlsvc ↔ session | in-process method calls | serialized through the actor's mailbox for mutations |
| layoutset ↔ session | async results channel | gsettings subprocess latency (10–50 ms) off the key path |

## Build Order Implications (for roadmap phasing)

Dependency-ordered foundation → derived:

```
layouts/ (pure, golden tests)              ← foundation A (no deps)
hotkey/ FSM (pure, stream tests)           ← foundation B (no deps)
        ↓
engine/ wire adapter (godbus)              ← integrates A? no — B only later;
test/e2e MINIMAL skeleton (ydotool→log)      │  M1 gate = "hotkey visible in log"
        ↓                                     │  needs engine/ + a log line
correct/ (uses layouts/)                   ← M2: word correction in an editor
        ↓                                   │  e2e grows AT-SPI readback + matrix v1
session actor integration (buffer+FSM+caps) ↗ (evolves through M1–M3)
        ↓
layoutset/ two-engine switching            ← M2/M3 (phrase/selection/regs + switch combos)
config/ + ctlsvc/ + goswitchctl             ← M3 (hot reload gate)
macros (Super→Ctrl per-app, §4.4b)          ← last functional layer: needs app-identity
                                              (FocusInId client string / AT-SPI) — riskiest,
                                              schedule behind core or descope to v2 via spec-delta
install/unit/docs/ADRs                      ← M4
```

Rationale for ordering:

1. **Pure packages first** (layouts, hotkey FSM): zero integration risk, unblock everything, CI-green from day one.
2. **engine/ + minimal e2e next**: this is the M1 gate and the spike that retires spec Q1/Q4 (engine viability, hotkey interception at engine level). If the two-engine layout pattern misbehaves, it surfaces here while the fallback (single-engine internal mode) is still cheap to adopt.
3. **Correction before switching**: word correction is the core value; switching integration (gsettings, MRU) is orthogonal and can lag.
4. **Config/ctlsvc after behavior**: reload semantics stabilize once there is something worth reloading.
5. **Macro layer last or out**: per-app Super→Ctrl remap needs focused-app identity, which the engine layer exposes only weakly (FocusInId's client string on 1.5.29) — flag as research item, candidate for spec-delta to v2.

## Research Flags for Phases

- **Phase engine-adapter (M1):** verify two-engine source activation end-to-end on target; confirm UpdatePreeditText arity and DeleteSurroundingText signal acceptance on the installed 1.5.29-rc2 (monitor traffic with dbus-monitor on the ibus socket).
- **Phase correction (M2):** per-app caps matrix (GTK4 editor, Firefox, Chromium/Electron) — decides fallback prevalence; terminals (wezterm) are spec Q3 with wl-clipboard fallback or documented refusal.
- **Phase macros (if kept):** app-identity mechanism needs its own spike.

## Confidence Assessment

| Area | Level | Basis |
|------|-------|-------|
| IBus lifecycle & wire surface | HIGH | live introspection on target + end-to-end engine activation performed during research |
| Engine method/signal signatures | HIGH | introspected; cross-checked with ibus main source and goibus |
| DeleteSurroundingText no-ack semantics | MEDIUM | read in ibus main source; installed 1.5.29-rc2 is an rc — verify on target in M1 |
| Two-engine layout switching | HIGH | gnome-shell keyboard.js source read directly (activateInputSource applies engine layout then sets engine) |
| Chromium surrounding-text breakage | MEDIUM | upstream issue #2354 (primary repo) — behavior version-dependent |
| Go/godbus viability (spec Q1) | HIGH | sarim/goibus exists as working pure-godbus engine implementation; patterns reproduced above |
| gsettings-subprocess integration | MEDIUM | standard practice; dconf-direct alternative undocumented (web tier per seam) |
| Mouse-click buffer reset | LOW (feasibility) | engines cannot see pointer events; approximation via surrounding-cursor jumps only |
| Per-app macro layer (§4.4b) | LOW | app identity at engine layer is weakly exposed; needs spike |

## Sources

Primary (read directly during this research):
- Live D-Bus introspection of ibus-daemon 1.5.29-rc2 on the target machine (org.freedesktop.IBus, InputContext, Engine, Factory interfaces), 2026-09-10
- Owner prototypes: `/home/nil/.local/share/punto-switcher/punto_engine.py` (438 lines, full IBus engine in Python GI), `/home/nil/.local/share/ibus-test/test_shift.py`
- ibus upstream source (main): `src/ibusengine.c` (DeleteSurroundingText/ForwardKeyEvent emit paths), `src/ibustypes.h` (RELEASE_MASK=1<<30, CapabilityFlags) — raw.githubusercontent.com/ibus/ibus
- gnome-shell (main): `js/ui/status/keyboard.js` (`_getXkbId`, `activateInputSource`, MRU handling) — gitlab.gnome.org/GNOME/gnome-shell
- sarim/goibus (GitHub): bus/factory/engine/component/engineDesc/common.go — pure-godbus libibus implementation

Secondary (web, seam tier LOW):
- [ibus issue #2354 — delete surrounding text broken in Chromium-based apps](https://github.com/ibus/ibus/issues/2354)
- [ibus issue #2054 — get_surrounding_text not working in Firefox](https://github.com/ibus/ibus/issues/2054)
- [keyd — kernel-level remapping daemon](https://github.com/rvaiya/keyd) and [Mouseless app-conflicts notes](https://mouseless.click/docs/app_conflicts.html) (uinput feedback-loop caveat)
- [The Input Stack on Linux — venam.net](https://venam.net/blog/unix/2025/11/27/input_devices_linux.html) (layering)
- [splondkie/wayland-accessibility-notes](https://github.com/splondkie/wayland-accessibility-notes/blob/main/README.md), [Fedora Magazine: Automation through accessibility](https://fedoramagazine.org/automation-through-accessibility/) (AT-SPI read/actuate pattern for e2e)

---
*Architecture research for: goswitch — IBus-engine layout corrector for GNOME Wayland*
*Researched: 2026-09-10*
