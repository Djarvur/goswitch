# Pitfalls Research

**Domain:** IBus input-method-engine daemon on GNOME Wayland (keyboard-layout corrector, Punto Switcher analog) written in Go
**Researched:** 2026-09-10
**Confidence:** MEDIUM-HIGH (key claims cross-verified against upstream issues/docs + empirical inspection of the owner's machine running IBus 1.5.29-rc2 on Ubuntu 24.04 GNOME Wayland, including the owner's own Python prototype in `~/.local/share/punto-switcher/punto_engine.py`)

**Phase vocabulary in this file** follows the spec's milestones (docs/SPEC.md §9): M0 spec frozen, M1 daemon skeleton + IBus registration + hotkey, M2 word correction e2e, M3 phrase/selection/case/config, M4 install/acceptance.

---

## Critical Pitfalls

### Pitfall 1: An engine crash kills input for the whole desktop (ibus-daemon suicide)

**What goes wrong:**
When an engine process exits abnormally, `ibus-daemon` kills itself by default to avoid dirty state. Since ibus-daemon is the input path for the entire session, one unhandled `nil` dereference / panic in `goswitchd` silences typing in every application until the daemon is restarted. The TUX IM dev log (June 2026) counted 8 such pitfalls in ~2 weeks of building a Python IBus engine — nearly all triggered daemon suicide.

**Why it happens:**
Developers assume engine bugs are contained to the engine process. The lifecycle coupling is undocumented in the API reference and only discoverable by killing the daemon repeatedly. In Go, a panic in a D-Bus method handler goroutine is especially easy to cause (bad variant type assertion from hand-written introspection) and it aborts the whole process.

**How to avoid:**
- Wrap the top of every exported D-Bus method/signal handler in a `recover()` shim that logs and returns a safe error reply instead of letting the process die. Treat "no panic may escape a handler" as an invariant enforced by one wrapper, not per-handler discipline (the TUX lesson: filter/handle at the single entry point).
- Guard every `godbus` variant type assertion (`Store`, `.([]interface{})`, custom `StoreAt` implementations) — these are the #1 panic source.
- Engine must handle SIGHUP/SIGTERM gracefully, and never `os.Exit` from a library path.
- Also design for the reverse: `ibus restart` (or any `ibus-daemon` death) kills *your* engine mid-flight. All in-memory buffer state must be reconstructable/clearable; config must be re-readable at startup (hot-reload already planned helps).

**Warning signs:**
- During manual testing, typing freezes desktop-wide right after a goswitch action → almost certainly a panic escaping a handler.
- Dmesg/journal shows `ibus-daemon` respawning while you debug the engine.

**Phase to address:**
M1 (process skeleton + registration) — the recover-shim belongs in the first commit of the D-Bus layer, before any feature logic. e2e in M2 must include a "kill -9 the engine, then type" case asserting desktop input survives.

---

### Pitfall 2: The engine only receives key events while it IS the active input source

**What goes wrong:**
The single most dangerous architectural assumption. IBus routes `ProcessKeyEvent` only to the engine of the currently active input source. On the owner's machine right now: `gsettings get org.gnome.desktop.input-sources sources` = `[('xkb', 'us'), ('xkb', 'ru')]` — plain XKB sources, no IBus engine active. In that state `goswitchd` sees **zero** key events: no double-shift hotkey, no buffer, nothing. A team that discovers this after M2 has to re-architect.

**Why it happens:**
The spec's mental model ("IBus engine sees keyboard events before applications") is true only for the active engine. GNOME Shell drives plain `xkb:` sources itself (GNOME Bugzilla 676102 lineage); an engine that is merely *registered* is dormant until the user selects it as the input source.

**How to avoid:**
Decide at M0 (this is the real content of open questions Q1/Q4):
- **Option A (recommended, matches the owner's prototype):** `goswitch` is the *only* input source. The engine declares `<layout>us</layout>` in its component XML, receives physical keys, and produces both EN and RU output itself via its own key-position tables (exactly what `punto_engine.py` line 82 `self.mode = 'ru'|'en'` already does). "Switching layout" becomes an internal mode flip — instant, no GNOME round-trip. Consequence: the GNOME layout indicator shows goswitch's `symbol` (e.g. "EN/RU"), not the system one, and the requirement "переключение раскладки проксируется в GNOME Super+Space" must be re-specified as "goswitch switches its own mode and additionally syncs/announces state" (see Pitfall 3).
- **Option B (rejected direction):** keep `xkb us` + `xkb ru` as sources and expect the engine to snoop — impossible via IBus; would require the EVIOCGRAB/uinput layer the project explicitly avoids.
- If hybrid per-app behavior is ever wanted, know that switching source away from goswitch (e.g. via Super+Space if the user still has other sources) takes the hotkeys with it — v1 should replace the source list or clearly warn.

**Warning signs:**
- M1 gate "hotkey event visible in log" fails with *no events at all* → the source isn't active, not a D-Bus bug.
- Works in `gnome-text-editor` but stops after user switches layout by habit.

**Phase to address:**
M0 (ADR: goswitch-as-the-layout-engine), M1 (verify event flow while active source). This decision gates everything downstream.

---

### Pitfall 3: `gsettings set org.gnome.desktop.input-sources current` does not switch layout at runtime

**What goes wrong:**
Naive implementation of "чтение/установка текущей раскладки через org.gnome.desktop.input-sources": writing `current` changes the dconf value but the layout stays; `mru-sources` reads stale. Live state lives in GNOME Shell's `InputSourceManager` (in-process JS), not in dconf (unix.stackexchange 316998, multiple corroborating threads). Reading `current` also drifts from the truth.

**Why it happens:**
The key is a one-way mailbox in practice: Shell writes it for observers; it does not subscribe to external writes. People trust the schema doc that says settable.

**How to avoid:**
- With Pitfall 2 Option A this requirement mostly dissolves: layout state is goswitch's own. Read `sources`/`current` only for initial state and diagnostics.
- If true GNOME-driven switching is still needed, the viable mechanisms are: (1) `org.gnome.Shell.Eval` calling `imports.ui.status.keyboard.getInputSourceManager().inputSources[n].activate()` — requires enabling unsafe mode on modern GNOME; (2) ship a tiny GNOME Shell extension exposing a D-Bus method (precedent: `kevinhwang91/gnome-shell-ibus-switcher`); (3) synthesize a real Super+Space keystroke via ydotool (needs the uinput daemon from the e2e stand — conflicts with "no root, no grab" purity but it's the same physical path). Note: `forward_key_event(Super+Space)` from inside the engine does **not** work — compositor keybindings are consumed by mutter above the IME path; the engine can only deliver keys *into the focused app*.
- Spec delta at M0: redefine Q-item "переключение раскладки проксируется в GNOME Super+Space" to the chosen mechanism; an ADR is mandatory.

**Warning signs:**
- e2e asserts gsettings value changed but the typed text proves the layout didn't.
- Tests pass on X11 fallback sessions but fail on the real GNOME Wayland session.

**Phase to address:**
M0 (ADR), M1 (implement chosen mechanism behind an interface so it can be swapped).

---

### Pitfall 4: No pure-Go IBus engine has ever shipped — the private-bus handshake is unproven (Q1)

**What goes wrong:**
The team treats "we speak D-Bus, ibus speaks D-Bus, therefore godbus works" as settled. Reality: engines do **not** talk to dbus-daemon. `ibus-daemon` runs its own private socket with a mini-bus implementation (`BusDBusImpl`: own Hello, NameOwnerChanged, match-rule engine), reachable via the address file `~/.config/ibus/bus/<machine-id>-unix-wayland-0` (empirically confirmed on this machine) or `IBUS_ADDRESS`. The only Go precedent (`sarim/goibus`, powering ibus-avro and ibus-bamboo) is **cgo** bindings to libibus — deliberately excluded by the project's stdlib+godbus constraint. A godbus engine must hand-roll: address discovery, `RegisterComponent`, `/org/freedesktop/IBus/Factory` + engine objects, the `org.freedesktop.IBus.Engine` method/signal surface, capability negotiation, and hand-written introspection XML matching libibus serialization quirks (variant ordering, `IBusText` structs-as-variants).

**Why it happens:**
Official docs describe the C API; the wire reality is in `src/ibusengine.c` / `bus/` sources. Everyone who skipped this step ends up either "fixing" it with cgo or abandoning the engine (the owner's 2024-era Python prototype used GI bindings precisely to avoid it).

**How to avoid:**
- M0/M1 spike: a minimal Go binary that connects to the address file, registers a component, gets spawned (or connects standalone), receives `process_key_event`, and logs keyvals. This is exactly the spec's M1 gate — do it before writing any buffer/correction code.
- Keep `/tmp` reference material: the daemon spawns engines with the component `exec` line; `ibus list-engine` must show the engine; use `ibus-engine-simple` and `test-shift` (already on this machine) as behavioral oracles.
- Pin the target: IBus 1.5.29 (Ubuntu 24.04). Version drift: older distros' ibus differ in portal/engine-set APIs; declare 1.5.x compatibility policy in the ADR.
- Write a thin `engine/ibus` package boundary so the handshake can be fixed without touching feature logic.

**Warning signs:**
- `godbus` connects but `RegisterComponent` returns "unknown method" → you are on dbus-daemon, not the ibus socket.
- Engine spawns then dies instantly with no daemon log — introspection/signature mismatch (see Pitfall 1 for the blast radius).

**Phase to address:**
M1 entirely. This is the highest-risk item; timebox a kill-criterion spike at the start ("if the handshake can't be proven in N days, fall back to evaluating goibus/cgo despite the constraint — owner decision").

---

### Pitfall 5: IBus registry cache staleness + root-only component dir vs the no-root install constraint

**What goes wrong:**
- New/reinstalled engines don't appear in `ibus list-engine` or GNOME Settings even after `ibus restart`, because the registry cache (`~/.cache/ibus/bus/` + system cache) is timestamp-based and frequently stale (mozc #500, Guix #22707, Fedora Silverblue #9). Sometimes a logout/login is required for GNOME Settings.
- The canonical component dir is `/usr/share/ibus/component/` (root-owned — the owner's own prototypes were dropped there with sudo). The spec demands install **without root** (`go install`, GitHub releases binary).
- Ghost artifacts accumulate: this machine has 15+ stale `~/.cache/ibus/dbus-*` sockets from previous daemon generations; old component XMLs pointing to deleted exec paths make ibus-daemon log spawn errors forever.

**Why it happens:**
ibus-daemon scans only standard dirs plus `IBUS_COMPONENT_PATH`; there is no blessed per-user component directory on stock ibus. The cache is an optimization that invalidates on mtime, which tooling (copying files, containers, Nix) defeats.

**How to avoid:**
- Installer (M4) must: install the component XML to `~/.local/share/ibus/component/` **and** ensure `IBUS_COMPONENT_PATH` includes it — which means the *ibus-daemon process* needs that env (set via `~/.config/environment.d/*.conf`, which GNOME's session absorbs, then restart ibus-daemon). Verify empirically at M4; fallback: a one-time polkit-free helper or documented `sudo` line for `/usr/share/ibus/component/`.
- The installer should run `ibus write-cache` + `ibus restart` and verify `ibus list-engine | grep goswitch`, failing loudly.
- Ship an uninstaller that removes the XML and refreshes the cache; never leave dangling `exec` paths.
- e2e stand must reset registry state between matrix runs (delete `~/.cache/ibus/bus/registry*`, restart daemon) to avoid "works after reinstall" false negatives.

**Warning signs:**
- Engine listed yesterday, gone today after a rebuild changed the binary mtime only.
- GNOME Settings > Keyboard shows no goswitch source while `ibus list-engine` shows the engine (cache/Settings refresh skew — needs logout or `write-cache`).

**Phase to address:**
M1 (basic XML + `write-cache` dance for dev loop), M4 (final no-root installer + ADR on IBUS_COMPONENT_PATH).

---

### Pitfall 6: Every key arrives twice; modifiers race; consuming the wrong event breaks typing

**What goes wrong:**
- IBus delivers **both** press and release for every key (`state & IBUS_RELEASE_MASK`, bit 30). Release events fed into the buffer double every letter ("n" becomes "nn") — TUX IM's very first engine bug.
- Returning `TRUE` (consumed) for a modifier press breaks Shift+letter, Ctrl+X etc. in the app; returning `TRUE` for the wrong printable swallows user input.
- Upstream ibus#2600 (closed not-planned): near-simultaneous shift-down+key (~1 ms, QMK/fast typists) can lose the Shift modifier through the ibus path entirely — clients see unshifted keys. You cannot fully fix this from the engine; you must not *corrupt* data when you see a modifier glitch.
- The engine's declared `<layout>` governs the keyvals you receive: goswitch (layout=us) never sees Cyrillic keyvals. Physical-key identity comes from `keycode`, not `keyval`.

**Why it happens:**
process_key_event semantics feel like X11 event handlers but the return contract and release bit are easy to miss; keyval-vs-keycode is conflated because on a US layout they look equivalent.

**How to avoid:**
- One top-of-dispatch filter: drop all release events for buffer-building paths; keep them only for the tap-timing state machine (which needs right-shift press AND release).
- Buffer logic keyed on `keycode` (physical) + shift state → char via goswitch's own tables; keyval used only as cross-check. This makes layout mode explicit and immune to keyval surprises.
- Never consume bare modifier presses (return FALSE); decide single/double/triple right-shift on the *release* timeline (see Pitfall 9).
- Unit-test the dispatcher with event streams (YAML cases in spec §7.1) including modifier glitches: shift-down, `2` down/up, shift-up with 1 ms spacing must not produce a spurious correction.

**Warning signs:**
- Doubled letters in e2e golden files.
- "Typing feels dead" reports on Shift+key combos after a hotkey feature lands.

**Phase to address:**
M1 (dispatcher + release filter), M2 (buffer keyed on keycode; glitch cases in unit corpus).

---

### Pitfall 7: "Replace exactly the wrong-text range" is not reliably expressible — surrounding text is a client-dependent luxury

**What goes wrong:**
Spec §4.3 requires `commit_text` to replace *exactly* the wrong range (Q2). The IBus mechanisms are: `set_surrounding_text` (client pushes context — only if client advertises `IBUS_CAP_SURROUNDING_TEXT`; engine should call `get_surrounding_text()` in `enable`), and `delete_surrounding_text(offset, n_chars)` (engine asks client to delete). Support is uneven even among GTK apps (stale-cache bug ibus#2423), poor in Chromium/XIM clients, absent in terminals. The owner's own prototype empirically hit this: `delete_surrounding_text(-n, n)` returns False in many apps → fallback `forward_key_event(BackSpace)×N`; and `commit_text` does **not** replace an active selection → the prototype fell back to `wl-copy`/`wl-paste` + forwarded Ctrl+V for selection correction.

**Why it happens:**
IME APIs were designed for preedit+commit composition, not retro-editing committed text. "Delete N chars then commit" has three independent failure modes: no surrounding text (can't verify), delete unsupported (must forward Backspaces), selection not owned (must paste).

**How to avoid:**
- M0 ADR for Q2 defining a capability ladder per app class: (1) surrounding text verified → `delete_surrounding_text` + `commit`; (2) no surrounding text → `Backspace×N` forwarded, N counted in **runes** (grapheme clusters for safety), then commit; (3) selection → clipboard replace via wl-copy/wl-paste + Ctrl+V forward (prototype-proven), restoring clipboard after; (4) terminals → v1 refusal or wl-clipboard paste per Q3.
- Track "what goswitch last committed" per context to compute N from its own buffer (prototype `last_committed`) — never from assumed app state when surrounding text is absent.
- Count deletion length in runes, and validate against surrounding text when available; abort (log + no-op) rather than send 50 Backspaces when uncertain — a wrong-length Backspace burst destroys user text.
- e2e matrix must include one app per capability tier (gnome-text-editor = tier 1/2, Chrome = tier 2/3, wezterm = tier 4).

**Warning signs:**
- Corrections eat one char too many/few intermittently (surrounding text lying or absent).
- Selection correction pastes but leaves original text (client didn't honor commit-over-selection).

**Phase to address:**
M0 (capability-ladder ADR), M2 (word path, tier 1–2), M3 (selection path tier 3 + terminals decision).

---

### Pitfall 8: "Per-application" behavior is not really available to an IBus engine

**What goes wrong:**
Two spec features assume app awareness: Super+Letter→Ctrl+Letter macros "в определённых приложениях" (§4.4b), and any future autocorrect exclusions. IBus gives the engine **no first-class app identity** — no app-id/window-class API on the engine interface. The owner already cited this exact limitation when rejecting autocorrect ("невозможность исключений по приложениям на Wayland") but the macro feature re-introduces it.

**Why it happens:**
Wayland deliberately hides global window state; IBus predates the need. The daemon knows the client's unique bus name, but mapping that to an application (short of heuristics over /proc or AT-SPI queries) is fragile and racy on focus change.

**How to avoid:**
- M0 spec delta: scope the macro feature as (a) global for v1 with an explicit kill-switch, or (b) per-app via best-effort identity: match the focused client's bus name via the ibus connection + `org.freedesktop.DBus.GetNameOwner`-style introspection, or shell out to AT-SPI for the focused accessible's application name — accepted as best-effort with documented false negatives.
- Design the rule engine around a `FocusInfo` interface now (even if v1 returns only "unknown app") so per-app rules can land later without re-architecting the hotkey dispatcher.

**Warning signs:**
- Macro fires in the terminal where it was supposed to be suppressed; flaky focus attribution in logs.

**Phase to address:**
M0 (scope decision), M3 (macro feature implementation with the FocusInfo seam).

---

### Pitfall 9: Double/triple right-shift detection is a latency-vs-correctness tradeoff, and the GNOME Super+Space proxy is unspecified

**What goes wrong:**
- The engine cannot know a right-shift press is "single" until either (a) another key arrives, or (b) a timeout expires. Fire on press → breaks every Shift+letter; fire only on a timeout → adds the timeout to every single-switch latency (budget is <50 ms total). Punto-class tools resolve single-tap on *release with no intervening key*, double-tap within a ~300–500 ms window — meaning "fix last word" fires at the *second* release, and a triple window must cancel a pending double.
- If the engine consumes right-shift events, apps lose legitimate shifted keys; if it passes them through (required), a "confirmed hotkey" has already leaked a bare Shift to the app — mostly harmless, but it mutates XKB latched state some apps react to.
- Proxying to GNOME's Super+Space (Pitfall 3) has no working mechanism via the engine API itself.

**Why it happens:**
Tap detection on modifiers is inherently deferred; developers prototype with generous timeouts (800 ms) that feel fine solo and fail the 50 ms requirement, or with tight ones that break fast typists.

**How to avoid:**
- Implement an explicit small state machine (states: idle, saw-RS1, saw-RS2-pending, …) with configurable `tap_window` (default ~350 ms) and `triple_window`; drive it from press+release pairs; make the *decision points* release-based.
- Budget honestly: single-shift switch = commit by release + mechanism latency (Option A internal flip ≈ instant; GNOME proxy adds round-trip). Document measured p95 in e2e output.
- Add e2e timing assertions with ydotool-paced taps (ydotool can emit with precise inter-key delays) — flaky-by-design tests must use generous CI margins but the unit corpus should pin the state machine transitions exactly.
- For "switch layout" v1, prefer internal mode flip (Pitfall 2 Option A); keep the GNOME-proxy mechanism behind an interface for later (Pitfall 3).

**Warning signs:**
- Users report "switch fires while I'm still typing shifted capitals".
- Latency assertions in e2e fail only on loaded machines (timeouts competing with scheduling).

**Phase to address:**
M1 (state machine skeleton + logging), M2 (wire to word correction; tune windows), M3 (triple-tap + configurable bindings via YAML).

---

### Pitfall 10: Buffer-clearing triggers that IBus cannot observe (mouse click, caret moves)

**What goes wrong:**
Spec §4.3: buffer clears on Enter, Tab, **mouse click**, focus change, Escape. An IBus engine does not see pointer events at all. Clicks surface only indirectly (many GTK widgets call `im_context_reset()` on button press; surrounding-text/cursor-location updates sometimes follow), and Chromium/XIM clients differ.

**Why it happens:**
The spec author listed natural "session boundary" events without checking which layer sees them. focus_in/focus_out and `reset()` are engine-visible; clicks are not first-class.

**How to avoid:**
- M0 spec delta: redefine triggers as "Enter/Tab/Escape (key events), `focus_out`, `reset()` callback, caret-position change inferred from `set_cursor_location`/surrounding-text updates" — and verify empirically which clients actually deliver `reset()` on click (gnome-text-editor, Chrome) at M2; note gaps in the ADR.
- If click-clearing proves essential for a class of apps, the honest mechanisms are surrounding-text polling (ugly) or accepting the limitation for v1.

**Warning signs:**
- e2e: type wrong word, click elsewhere in the same field, tap correction → it edits the wrong span (buffer not cleared).

**Phase to address:**
M0 (spec delta), M2 (empirical per-client check becomes an e2e case).

---

### Pitfall 11: The e2e stand rots silently — focus stealing, ydotool perms, AT-SPI flags, leaked IBus state, no Wayland CI

**What goes wrong:**
- GitHub-hosted runners have no GNOME Wayland session; e2e silently degrades to "manual dispatch" that nobody runs, then bitrots.
- ydotool: default socket perms are root-only 600 (RH bugzilla 2250692; upstream issues #73/#198/#207) — needs `ydotoold` provisioned, a dedicated group, `UMask=0077`, and `YDOTOOL_SOCKET` exported; breakage looks like "injection silently does nothing".
- AT-SPI + Chromium: no tree until an AT connects or Chrome launches with `ACCESSIBILITY_ENABLED=1` + `--force-renderer-accessibility`; the tree materializes asynchronously (poll/wait needed); the flag has regressed across Chrome versions. Reading text too early yields empty/stale nodes.
- The known trap already in the spec (§7.3): the test MUST activate the target window itself; notifications/other windows steal focus nondeterministically.
- IBus state leaks between cases: previous case's mode/selection/register lingers; a red case is actually the previous case's fault.
- Focus verification: asserting via AT-SPI focus events is flaky; asserting nothing yields ghost-failures.

**Why it happens:**
Every dependency sits outside the app's control and degrades independently; each failure mode masquerades as an engine bug, burning days.

**How to avoid:**
- Stand = code, not wiki: a `test/e2e/` runner that (1) provisions/checks ydotoold + socket + group and fails fast with a diagnostic; (2) launches Chrome with the flags; (3) activates windows via AT-SPI/WM before every case (spec requirement — encode as a hard precondition assert); (4) resets goswitch state (via `goswitchctl` reset) AND restarts/clears engine between cases; (5) waits-for-condition (not sleep) on AT-SPI text values with timeout+retry.
- CI policy from M1: unit/golden on push (headless, no Wayland needed — pure Go + fake D-Bus connection), e2e on manual dispatch on the owner's machine, with a scheduled monthly canary run to prevent rot.
- Make every e2e case print the observed AT-SPI text on failure (golden-diff style) — the #1 debugging shortcut.
- Keep the ydotool pacing configurable so timing-sensitive tap tests can be slowed on slow machines without weakening production timing.

**Warning signs:**
- e2e passes interactively, fails under dispatch (focus/permissions).
- Random ~5% failure rate concentrated in Chromium cases (tree timing).

**Phase to address:**
M1 (stand skeleton + ydotool + window activation), M2 (matrix v1 + reset discipline), M4 (runner docs + scheduled canary).

---

### Pitfall 12: Two supervision domains fight — systemd user unit vs ibus-daemon spawning; godbus never reconnects

**What goes wrong:**
- `Type=dbus` in a systemd user unit only tracks names on the **session bus** systemd watches; the goswitch engine lives on ibus-daemon's **private** socket, and ibus-daemon launches engines itself via the component `exec`. Making systemd the supervisor of the same binary produces double-spawning (ibus spawns one, systemd another), port/path conflicts, or systemd "restarting" a process ibus thinks it owns. Also: user units need `PartOf=graphical-session.target` + `After=`, and `WAYLAND_DISPLAY` imported into the user manager; `Requires=` creates dependency-failure loops.
- godbus has no auto-reconnect: when the bus connection drops, channels from `Signal()`/`Eavesdrop()` close and pending calls return `ErrClosed`. Unhandled, the daemon becomes a zombie that logs nothing.

**Why it happens:**
Standard service-management instincts (systemd owns daemons) collide with IBus's 2008-era design (ibus-daemon owns engines). The team copies a typical `Type=dbus` unit from another project.

**How to avoid:**
- M0 ADR for Q6: ibus-daemon is the engine's supervisor (it spawns `goswitchd --engine` per component exec). The systemd **user** unit, if kept at all, supervises only the control/monitor sidecar (`goswitchctl` D-Bus API on the session bus, config-watcher) — that one may legitimately use `Type=dbus` with `BusName=` on the session bus, `PartOf=graphical-session.target`.
- One process, two roles, one flag: `goswitchd` runs engine mode when spawned by ibus (component exec), control mode under systemd. Shared state via a small file/socket or by the control sidecar talking to the engine over the session bus.
- Wrap every godbus connection in a reconnect loop keyed off the closed signal channel; re-dial, re-export objects, re-add match rules; treat engine state as disposable across restarts (buffer loss on ibus restart is acceptable — document it).
- Expect ibus-daemon restart to kill the engine ( Pitfall 1 corollary): config hot-reload must survive; nothing may require a warm cache to function.

**Warning signs:**
- Two `goswitchd` processes in `ps` after an install script runs.
- Daemon alive but silent after an `ibus restart` (connection zombie).

**Phase to address:**
M0 (ADR), M1 (process model + reconnect wrapper), M4 (final units + installer).

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Backspace×N fallback everywhere, skip surrounding text | Correction works in M2 quickly | Data loss when count is wrong; unfixable in terminals | Acceptable as tier-2 path behind the ladder (Pitfall 7), never as the only path |
| Layout tables copy-pasted from transliteration sites | Tables exist day 1 | Wrong punctuation mapping (translit ≠ key-position; e.g. `/`↔`.`, `?`↔`,`, `` ` ``↔`ё`, `@`↔`"`, `#`↔`№`) | Never — generate from xkb symbols files (`xkbcomp`/`setxkbmap -print`, `go:generate`) and golden-test |
| Hand-rolled introspection XML "good enough for GTK" | godbus engine works sooner | Chromium/XIM clients reject or misparse signatures later | MVP only, with an e2e case per client class from M2 |
| e2e only on the owner's machine, no provisioning script | Fast start | Stand rot, unreproducible reds | Only if the provisioning script lands by M4 (never acceptable after) |
| Sleeps instead of wait-for-condition in e2e | Stand "works" | Random CI reds, ignored failures | Never — condition helpers from day one |
| Logging keyvals unconditionally in debug | Debugging is easy at M1 | Passwords/typed secrets leak into logs | Only behind explicit opt-in flag with big warning; default off in shipped binaries |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| ibus-daemon (private bus) | Connecting to the session bus; assuming dbus-daemon routing | Read address from `IBUS_ADDRESS` or `~/.config/ibus/bus/<machine-id>-unix-wayland-0`; speak to the mini-bus; re-read on reconnect |
| ibus-daemon lifecycle | Assuming engine survives daemon restart | Persist nothing critical; re-register on spawn; expect SIGTERM on `ibus restart`; tolerate `ibus write-cache` dance |
| GNOME input-sources | Writing `current` to switch; trusting `mru-sources` | Own the layout mode (engine-internal); use Shell Eval/extension/ydotool only if GNOME-driven switch is a hard requirement |
| GTK apps | Assuming commit replaces selection; assuming reset() on click | Capability ladder (Pitfall 7); test per widget class (GtkEntry vs GtkTextView vs libadwaita) |
| Chromium/Electron | Testing only default-mode Chrome; assuming IME works in native Wayland mode | Pin windowing mode + flags (`--enable-wayland-ime --wayland-text-input-version=3` for native; XWayland uses XIM and has fast-typing reorder bugs); one e2e case per mode you claim to support |
| wezterm/terminals | Expecting preedit/surrounding text; expecting commit to land | XIM-only on Wayland (`XMODIFIERS=@im=ibus`), text-input-v3 still open (wezterm#1772); v1 = detect & refuse or wl-clipboard paste fallback (Q3 ADR) |
| ydotool (e2e) | Running as plain user with default socket perms | Provision ydotoold + group + UMask + `YDOTOOL_SOCKET`; fail fast with diagnostics in the stand |
| AT-SPI (e2e) | Querying immediately after launch/focus | Wait-for-condition with timeout; Chromium needs `ACCESSIBILITY_ENABLED=1` + `--force-renderer-accessibility`; expect async tree |
| systemd user units | `Type=dbus` for the engine; `Requires=graphical-session.target` | ibus-daemon supervises the engine; systemd unit (if any) for the session-bus sidecar with `PartOf=`; import WAYLAND_DISPLAY |
| keyd/xremap coexistence | Trying to grab or re-map keys yourself | Pure IBus layer already coexists; never open evdev/uinput in the product (e2e stand's ydotool is test-only) |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Subprocess in the hot path (prototype shells out to `wl-paste`/`wl-copy` on selection correction) | 100–300 ms spikes, budget is <50 ms | Pre-warm, or avoid clipboard path in word/phrase flows; measure p95 in e2e | Every selection correction on loaded machines |
| ibus path adds inherent latency (community reports input lag in recent ibus releases) | p95 near/over budget though goswitch code is fast | Keep handler work O(1); commit synchronously, log async; measure end-to-end with ydotool→AT-SPI timestamps, not internal timers | Fast typists; dual-tap windows |
| Per-key allocations in the buffer (string concat per keystroke) | GC pauses visible as missed taps | Preallocate rune buffer, recycle; benchmark in CI (`go test -bench`) | Long typing sessions |
| Hot-reload re-reading/parsing YAML per event | Config change stalls typing | Watch + swap atomic pointer; reload only on file event | M3+ |
| Timeout windows tuned on an idle machine | Tap detection flaky under load | Make windows configurable; e2e pacing tolerant; document measured latencies | Always (design for margin) |

Memory budget (<50 MB) is a non-issue for Go + small tables unless dictionary features arrive later (TUX IM's trie dictionaries are the cautionary tale — keep v1 dictionary-free per spec).

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Debug key-trace logging on by default | Passwords/secrets typed into any field end up in logs/journal | Debug trace behind explicit flag; redact or default-off in release builds; document in README |
| Treating the buffer as non-sensitive (it is a keystroke recorder) | Buffer dumps in crash logs/logs expose secrets | Never log buffer contents at info level; crash reports contain lengths only |
| Component XML `exec` pointing at world-writable path | Another user/process can swap the binary → code exec as you on every ibus spawn | Installer checks binary dir perms (0700 under `~/.local`); warn on unsafe paths |
| Clipboard round-trip for selection correction leaves converted text in clipboard | Sensitive text left in CLIPBOARD after correction | Restore previous clipboard content after paste; make the mechanism documented |
| e2e stand disables screen lock / runs with accessibility wide open on the owner's session | Whole-session exposure if stand is left running | Stand tears down flags; runner warns if left enabled; canary job auto-cleanups |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|------------------|
| Visible "jump" when correcting (delete-then-commit flicker) | Feels broken vs Punto's seamless replace | Prefer surrounding-text delete + single atomic commit; e2e should assert final text AND that no intermediate state is observable via AT-SPI polling (bounded) |
| Correction fires on ambiguous words (e.g. `aaaa`, `ror`) | Trust destruction — user stops using the tool | Auto-detect direction by letter-set majority but never *silently* correct (spec already rejects autocorrect); on ambiguity keep explicit hotkey semantics deterministic |
| Hotkey latency = tap window + action | "Sluggish" feel | Release-based decisions; internal layout flip (no GNOME round-trip); measure and publish p95 |
| Layout indicator disagrees with goswitch mode (engine symbol static "RU") | User unsure which layout is active | Update engine `symbol`/property on mode flip (IBus property mechanism) so GNOME Shell indicator follows |
| Buffer survives focus change by accident | Correction edits text in the wrong window/app | Clear buffer on `focus_out` unconditionally; e2e cross-window case |

## "Looks Done But Isn't" Checklist

- [ ] **IBus registration:** component visible in `ibus list-engine` AND spawnable — verify daemon actually launches the binary (journal shows exec) after `ibus write-cache` + restart
- [ ] **Hotkey:** works when goswitch is the active source in GNOME (not only when launched standalone for testing) — the M1 gate must run through the real source switch
- [ ] **Release filtering:** typed text has no doubled letters in e2e golden files (press+release both delivered)
- [ ] **Correction length:** deletion count validated against surrounding text when available; wrong-count → abort, never guess
- [ ] **Case preservation:** `GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет`, `ghbdtn`→`привет` AND punctuation variants (`}`→`Ъ`, `?`→`,`, `` ~ ``→`Ё`) in golden corpus
- [ ] **Mixed text:** only the wrong-layout run is converted; spec's exact semantics implemented and tested (Q at §4.2)
- [ ] **Dead keys / Compose:** sequences (`Compose`+key, dead-key accents) still work — forwarded, not swallowed
- [ ] **Terminals:** explicit behavior in wezterm (works via XIM / refuses gracefully per Q3 ADR), not accidental
- [ ] **Chromium:** correction verified in BOTH the default windowing mode and any mode claimed supported, with flags pinned by the stand
- [ ] **Restart survival:** `ibus restart` and engine `kill -9` — typing continues, goswitch re-registers, no ghost sockets/XMLs block re-spawn
- [ ] **No-root install:** fresh user, `go install` + installer script only — engine registered and spawned without a single sudo (except documented fallback)
- [ ] **e2e matrix green twice in a row** on an untouched owner session (catches state leakage and focus assumptions)

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Engine crash kills desktop input (1) | LOW | Add/fix recover shim; the daemon self-heals on restart — but audit *why* it reached production: add the missing handler-level test |
| Wrong architecture re: active source (2) | HIGH | Migrate to engine-as-layout (Option A): move layout tables into commit path, add mode state; buffer logic survives; GNOME-proxy requirements re-scoped via spec delta |
| GNOME switch mechanism wrong (3) | MEDIUM | Swap mechanism behind the interface (Eval→extension→ydotool); no buffer changes |
| godbus handshake dead-end (4) | HIGH | Fall back to goibus/cgo (owner decision, constraint change) or ibus XML `exec` wrapping a Python-free helper; M1 kill-criterion exists precisely to make this cheap *now*, expensive later |
| Stale registry / engine not found (5) | LOW | `ibus write-cache` (+ `--system`), `ibus restart`, verify `ibus list-engine`; clean `~/.cache/ibus/bus/registry*`; last resort logout/login |
| Deletion wrong length (7) | MEDIUM | Add tier-1 path where missing; tighten rune/grapheme counting; add golden cases; restore from buffer journal (`last_committed`) |
| Tap window mistuned (9) | LOW | Config-only change (YAML); ship measured defaults |
| e2e rot (11) | MEDIUM | Re-run provisioning script; fix flag drift (Chrome regressions); canary job surfaces rot monthly |
| Double-spawned daemon (12) | LOW | Kill stray process; fix unit/ADR; make binary refuse to start engine mode twice (lock file) |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| 1. Crash → daemon suicide | M1 | Code review of recover shim + e2e case "kill engine, keep typing" |
| 2. Active-source visibility | M0 ADR / M1 | M1 gate: hotkey logged while goswitch is the GNOME-active source |
| 3. GNOME layout switch runtime-ignored | M0 ADR / M1 | e2e: switch then type, assert Cyrillic output |
| 4. Pure-Go private-bus handshake (Q1) | M1 spike (kill-criterion) | Minimal engine registers, spawns, receives keys via godbus |
| 5. Registry cache + no-root install | M1 (dev loop) / M4 (installer) | `ibus list-engine` after install on clean user; installer self-check |
| 6. Double delivery / modifier races / keycode-vs-keyval | M1–M2 | Unit corpus with raw event streams incl. ibus#2600-style glitch |
| 7. Range replacement / surrounding text (Q2) | M0 ADR / M2–M3 | e2e per capability tier; abort-not-garbage assertion |
| 8. Per-app rules unavailable | M0 scope / M3 | FocusInfo seam present; macro behind kill-switch |
| 9. Tap timing + Super+Space proxy | M1 / M2–M3 | State-machine unit tests + ydotool-paced e2e with p95 report |
| 10. Unobservable clear triggers (click) | M0 spec delta / M2 | e2e click-then-correct case per app class |
| 11. e2e stand rot | M1 / M2 / M4 | Stand provisions itself; scheduled canary; double-green rule |
| 12. Supervision + reconnect | M0 ADR / M1 / M4 | `ibus restart` test; single-process invariant; reconnect loop unit test |

## Sources

- [TUX IM Dev Log Part 2 — engine crashes killing ibus-daemon, release-mask doubling (tux.fan, 2026-06-25)](https://tux.fan/2026/06/25/tux-im-dev-log-part-2-·-the-first-scream-of-the-ibus-engine-how-i-crashed-the-daemon-troubleshooting/) + [tux-dot-fan/tux-im repo](https://github.com/tux-dot-fan/tux-im)
- [ibus/ibus D-Bus communication — private mini-bus architecture (DeepWiki)](https://deepwiki.com/ibus/ibus/2.2-d-bus-communication), [Input Method Engines — process_key_event contract](https://deepwiki.com/ibus/ibus/6-input-method-engines), [IBusEngine 1.5 reference](http://ibus.github.io/docs/ibus-1.5/IBusEngine.html), [IBusInputContext — delete-surrounding-text](https://ibus.github.io/docs/ibus-1.5/IBusInputContext.html)
- [ibus#2600 — fast shift sequences lose modifier (closed not-planned)](https://github.com/ibus/ibus/issues/2600); [ibus#2423 — surrounding text staleness](https://github.com/ibus/ibus/issues/2423); [ibus#2930/#2931 — socket/registry lifecycle](https://github.com/ibus/ibus/issues/2930); [ibus#2322 — xkb layouts as simple.xml engines](https://github.com/ibus/ibus/issues/2322)
- [GNOME Bugzilla 676102 — XKB layouts vs IBus integration](https://bugzilla.gnome.org/show_bug.cgi?id=676102); [unix.stackexchange 316998 — `current` gsetting ignored at runtime](https://unix.stackexchange.com/questions/316998/how-to-change-keyboard-layout-in-gnome-3-from-command-line); [kevinhwang91/gnome-shell-ibus-switcher — extension-based switching precedent](https://github.com/kevinhwang91/gnome-shell-ibus-switcher)
- [wezterm#1772 — text-input-v3 open](https://github.com/wezterm/wezterm/issues/1772), [#5125 IBus not working](https://github.com/wezterm/wezterm/issues/5125), [#3411 preedit](https://github.com/wezterm/wezterm/issues/3411), [#6925 IME reconnect perf](https://github.com/wezterm/wezterm/issues/6925), [use_ime docs](https://github.com/wezterm/wezterm/blob/main/docs/config/lua/config/use_ime.md)
- [Electron#33662 — gtk-version=4 crash / Wayland IME](https://github.com/electron/electron/issues/33662); [VS Code#120084 — IM modules not loaded under Wayland ozone](https://github.com/microsoft/vscode/issues/120084); [Fast-typing reordering under XWayland (blog.040304.xyz)](https://blog.040304.xyz/posts/wayland-input/); [Fcitx5-on-Wayland environment guidance](https://fcitx-im.org/wiki/Using_Fcitx_5_on_Wayland)
- [goibus (cgo Go IBus bindings)](https://github.com/sarim/goibus), [ibus-avro](https://github.com/sarim/ibus-avro), [ibus-bamboo](https://pkg.go.dev/github.com/bambooengine/ibus-bamboo)
- [ydotool socket permissions — bugzilla 2250692](https://bugzilla.redhat.com/show_bug.cgi?id=2250692), [issues #73](https://github.com/ReimuNotMoe/ydotool/issues/73)/[#198](https://github.com/ReimuNotMoe/ydotool/issues/198)/[#207](https://github.com/ReimuNotMoe/ydotool/issues/207)
- [Chromium accessibility — force-renderer-accessibility (chromium.org)](https://www.chromium.org/developers/design-documents/accessibility/), [AT-SPI for Chrome (SO 26152972)](https://stackoverflow.com/questions/26152972/at-spi-for-google-chrome-in-linux)
- [godbus docs — signal channel close semantics](https://pkg.go.dev/github.com/godbus/dbus), [godbus#109](https://github.com/godbus/dbus/issues/109)
- [systemd.io desktop integration](https://systemd.io/DESKTOP_ENVIRONMENTS/), [graphical-session.target design slides](https://people.debian.org/~mpitt/systemd.conf-2016-graphical-session.pdf), [sway#7862](https://github.com/swaywm/sway/issues/7862)
- [ibus registry cache tips (desktopi18n)](https://desktopi18n.wordpress.com/2015/03/10/ibus-tips/), [mozc#500 write-cache](https://github.com/google/mozc/issues/500), [Guix#22707](https://issues.guix.gnu.org/22707)
- Go standard library case-mapping semantics (`go doc strings.ToUpper`, `unicode.SpecialCase` on Go 1.27.1)
- Empirical inspection (2026-09-10) of the target machine: IBus 1.5.29-rc2, address files under `~/.config/ibus/bus/`, stale `~/.cache/ibus/dbus-*` sockets, registered prototypes (`/usr/share/ibus/component/punto-switcher.xml`, `test-shift.xml`), owner's prototype source `~/.local/share/punto-switcher/punto_engine.py` (surrounding-text failure, Backspace fallback, clipboard-paste selection hack, engine-internal layout mode)

---
*Pitfalls research for: IBus-engine layout switcher/corrector on GNOME Wayland in Go*
*Researched: 2026-09-10*
