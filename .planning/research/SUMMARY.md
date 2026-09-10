# Project Research Summary

**Project:** goswitch (Djarvur/goswitch)
**Domain:** IBus input-method-engine daemon for GNOME Wayland — keyboard-layout switcher / mistyped-text corrector (Punto Switcher analog), written in Go
**Researched:** 2026-09-10
**Confidence:** HIGH for core viability and architecture; MEDIUM for per-app behavioral details (see Gaps)

## Executive Summary

goswitch is a single Go daemon (`goswitchd`) operating as an IBus input-method engine over D-Bus: it sees keystrokes after XKB translation and after any keyd/xremap remap, buffers what the user types, and on a hotkey replaces exactly the wrong-layout range via IBus `CommitText`. The category is mature on Windows/macOS (Punto Switcher, Caramba, Mahou) and empty on GNOME Wayland precisely because every existing tool works at the evdev layer (root, EVIOCGRAB keyboard monopoly, remapper conflicts). The IBus-engine level is the only layer that coexists with remappers by construction, and research validated it end-to-end on the actual target machine: the IBus 1.5.29 daemon was introspected live over its private socket, an engine was registered and activated during research, and the owner's two Python prototypes prove the full key-receive/commit cycle on this hardware. Spec question Q1 (Go viability) is retired: godbus speaks the IBus socket protocol in production today (ibus-bamboo, Fedora-packaged 2026); the only real engineering is an in-repo `engine/` protocol adapter (~800–1,200 LOC), required solely because the existing Go bindings (`sarim/goibus` / `BambooEngine/goibus`) ship **no license file** and cannot be imported into an MIT project.

The recommended shape: a three-dependency stack (`godbus/dbus/v5` v5.2.2, `gopkg.in/yaml.v3` v3.0.1, `fsnotify` v1.10.1; `log/slog` for logging, `gsettings` subprocess for GNOME integration), a strict wire-adapter boundary (`engine/` is the only package touching godbus for IBus), a single session-actor goroutine owning all mutable state, and pure, headless-CI-testable packages beneath it (`layouts/` generated tables, `hotkey/` tap FSM, `correct/` scope+direction logic). Replacement follows a capability ladder per app class (surrounding-text delete where supported, Backspace×N fallback in runes elsewhere, clipboard round-trip for selections/terminals). Q6 (autostart) is retired: systemd user unit + programmatic `_RegisterComponent` + a reconnect/re-register loop, because IBus 1.5.29 loses engine registrations on daemon restart by design (ibus#2910, closed not-planned).

Three risks dominate everything else. **First — the #1 architectural decision:** ARCHITECTURE recommends a *two-engine registration pattern* (`goswitch-en` layout=us + `goswitch-ru` layout=ru as `('ibus',…)` input sources; verified against gnome-shell `keyboard.js` source), while PITFALLS documents the owner's *working prototype* doing the opposite — a single engine with an internal en/ru mode flip — on the grounds that writing `gsettings … input-sources current` is ignored at runtime by GNOME Shell. These must be reconciled via ADR + an M1 kill-criterion spike before any buffer/correction code is written (details below). **Second:** single right shift (= switch, <50 ms NFR) and double/triple right shift (= correct) share one key; the standard 300–500 ms double-tap window cannot fit inside a <50 ms budget. Adopt Caramba-style semantics — fire the switch immediately on first tap; a second tap within the window corrects the pre-switch buffer — and pin it in requirements, not in code. **Third:** a panic escaping a D-Bus handler makes ibus-daemon kill itself, silencing typing desktop-wide; a `recover()` shim belongs in the first commit of the D-Bus layer.

## Key Findings

### Recommended Stack

Minimal by design — three external dependencies total, everything else stdlib. Verified against the Go module proxy (authoritative) and the live target system (Ubuntu 24.04, GNOME 46, IBus 1.5.29-rc2, Go 1.27.1).

**Core technologies:**
- **Go ≥ 1.23 (module), build with 1.27.x** — project constraint; `log/slog` covers structured logging with zero deps.
- **`github.com/godbus/dbus/v5` v5.2.2** — all D-Bus, both the IBus private socket (service side: `Export`, signal channels) and the session bus (control service); the only maintained native Go D-Bus impl; proven against the IBus socket by ibus-bamboo in production.
- **In-repo `engine/` adapter (~1k LOC)** — the IBus wire protocol (address discovery, `_RegisterComponent`, Factory, engine objects, Commit/Forward/DeleteSurrounding signals). Mandatory: both goibus repos have **no LICENSE** (GitHub-API-verified) — legally unusable as an import; legitimate as read-only behavioral reference. Truth sources installed locally: `/usr/include/ibus-1.0/*.h` + the owner's validated Python prototypes. This is the single real piece of engineering in the stack.
- **`gopkg.in/yaml.v3` v3.0.1** — config parsing; frozen/stable is a feature under the audit constraint. (`goccy/go-yaml` is a mechanical swap if strict typing is later wanted.)
- **`fsnotify` v1.10.1** — config hot reload, with the directory-watch + debounce + `atomic.Pointer[Config]` pattern (editors save via atomic rename, swapping the inode).
- **systemd user unit** (off `graphical-session.target`, after the distro's IBus unit, `Restart=on-failure`) — autostart/supervision, no root anywhere; programmatic registration means zero files outside `$HOME`.

**Critical version facts:** godbus must be the v5 module path (v4 is EOL); ydotool must be the distro 0.1.8 (opens `/dev/uinput` directly — no ydotoold daemon; upstream 1.0.x has a different architecture and is not in 24.04); IBus target pinned at 1.5.x (1.5.29 on target). Do **not** use: goibus (no license), Viper (~10+ transitive deps), cgo/libibus (breaks static binary + audit story), xdotool (X11-only), uinput anywhere in the product (e2e stand only).

### Expected Features

**Must have (table stakes — every competitor has them):**
- Correct last word by hotkey — the defining gesture of the category; the most-used feature
- Correct whole phrase; correct selected text — standard second/third gestures
- Switch layout by hotkey — a switcher that cannot switch is not a switcher
- Case preservation per character (`Ghbdtn`→`Привет`); full tables incl. punctuation/digits; direction auto-detect by script majority
- Mixed-text handling (convert only the wrong-layout span) — **highest-complexity table-stakes item**; needs explicit semantics ADR + golden corpus
- Buffer reset rules; configurable hotkeys; no-root install; exact-range replacement with no visible jump — the core architectural bet that justifies IBus over uinput

**Should have (differentiators — no working GNOME Wayland competitor offers them):**
- IBus-engine level with no keyboard grab → keyd/xremap coexistence by construction (the founding decision)
- Automated e2e matrix (ydotool + AT-SPI, YAML cases) — unique in the entire category; the project's quality moat *and* acceptance gate
- Multi-tap hotkeys (double/triple right shift — Caramba ergonomics); switch+correct combo (Shift+RightCtrl); YAML hot reload without restart
- Terminal fallback via wl-clipboard paste (v1.x, after Q3); Super+Letter→Ctrl+Letter per-app macros (v1.x, needs app identity); single static MIT binary, zero telemetry; <50 ms measured hotkey reaction

**Defer (v2+) / rejected:**
- Exception heuristics (URL/e-mail/hex) — v2, unnecessary while correction is user-triggered
- Additional layout pairs — data-only later; table interface stays generic
- Autocorrect (owner-rejected: passwords, code, no reliable per-app exclusions on Wayland), GUI settings, typed-text diary, clipboard manager, transliteration, typo dictionaries, cloud hooks, tray indicator (GNOME Shell already shows one)

**Feature dependency insights:** every correction feature hangs off engine interception + typed-text buffer; buffer-reset rules gate phrase correction (word can ship first — the natural M2/M3 seam); selection correction and terminal fallback share one clipboard transport (build it once either way); per-app context (from IBus focus events) unlocks both terminal detection and macros — one mechanism, two consumers; the e2e matrix grows with every milestone and is itself the acceptance criterion.

### Architecture Approach

A single daemon split into a thin D-Bus wire adapter and a pure, testable core; all state funneled into one actor goroutine. Layering verified live: engine sees keys *post-XKB and post-remap*, injects text via signals routed back through ibus-daemon — remapper coexistence holds by construction, never by effort.

**Major components:**
1. **`engine/` wire adapter** — owns the IBus private-socket connection (address discovery → Dial/Auth/Hello → RequestName single-instance guard → Factory export → `_RegisterComponent` → per-input-context engine objects); translates D-Bus ↔ typed events; the only godbus-for-IBus package. Includes the reconnect + re-register loop with backoff.
2. **`internal/session` actor** — one goroutine owning ALL mutable state: per-IC keystroke buffers, tap FSM state, capability bitmaps; timers re-enter as synthetic events. Zero mutexes on the hot path.
3. **`internal/hotkey`** — pure tap/combo state machine (single/double/triple RShift, Shift+RCtrl, modifier-use discrimination, configurable windows), unit-testable against synthetic event streams.
4. **`internal/correct` + `layouts/`** — scope resolution (word/phrase/selection), direction detection, case-preserving conversion over generated positional tables (`go:generate` from `/usr/share/X11/xkb/symbols/{us,ru}` — kills the hand-transcription bug class the owner's prototype exhibited).
5. **`internal/layoutset` / `internal/config` / `internal/ctlsvc`** — gsettings subprocess wrapper for input-sources; YAML schema + fsnotify hot reload publishing immutable snapshots; session-bus service for `goswitchctl`.

**Key patterns:** reply to `ProcessKeyEvent` immediately (it is a synchronous daemon call — debounce waits happen *after* replying, via timers); capability-aware replacement strategy (SURROUNDING_TEXT cap → `DeleteSurroundingText`+`CommitText`, else `ForwardKeyEvent(BackSpace)×n`+`CommitText`; `DeleteSurroundingText` on 1.5.29 is a fire-and-forget **signal with no ack**); per-input-context buffers keyed by engine object path, reset on FocusOut; switch layouts via `gsettings` write, **never** by forwarding a synthetic Super+Space (compositor grabs sit above the IME path).

### Critical Pitfalls

1. **Engine crash = desktop-wide input death** — ibus-daemon suicides when an engine dies abnormally. Prevention: a `recover()` shim wrapping every exported D-Bus handler in the first commit of M1; guard every godbus variant type assertion (the #1 panic source); e2e case "kill -9 the engine, then type."
2. **The engine sees keys ONLY as the active input source** — the single most dangerous assumption. Both viable designs require goswitch to *be* the input source (see Decision #1 below); the rejected Option (keep plain xkb sources and snoop) is impossible via IBus.
3. **`gsettings set … current` may be a one-way mailbox at runtime** — Shell's in-process InputSourceManager holds live state; dconf writes change the value but not necessarily the layout (STACK verified the key is *writable* on GNOME 46; writability ≠ runtime effect). This is the crux of Decision #1.
4. **Every key arrives twice (press + release, bit 30)** and keycode ≠ keyval — filter releases from buffer paths (else doubled letters); key the buffer on physical `keycode` + own tables; never consume bare modifier presses (breaks Shift+letter); tolerate ibus#2600-style modifier glitches without corrupting data.
5. **Exact-range replacement is a capability ladder, not a given** — Chromium/Electron don't implement `DeleteSurroundingText` (ibus#2354); `commit_text` does not replace an active selection (prototype hit this → clipboard round-trip). Count deletions in runes, validate against surrounding text when available, and **abort rather than send a wrong-length Backspace burst**. Plus: registry-cache staleness and the root-owned component dir (no-root installer must dance around `IBUS_COMPONENT_PATH` / `ibus write-cache`); e2e stand rot (ydotool provisioning, Chromium a11y flags, wait-for-condition not sleeps, monthly canary).

## Implications for Roadmap

Suggested phase structure mirrors the spec's milestones (M0–M4). M0 is a *decision* milestone, not a build milestone — fold it into requirements definition / first-phase planning. Four build phases follow, ordered by the dependency chain: pure packages → engine adapter + minimal e2e → correction → switching/config → install/acceptance.

### Decision Pack (spec M0 — settle during requirements/plan, before any code)

**Rationale:** Five semantics are contested or unobservable-in-v1 and each caused a documented failure elsewhere; discovering them in code is the most expensive possible path.
**Delivers:** ADRs for: (1) layout-switching mechanism — **Decision #1**, see below; (2) multi-tap semantics vs the <50 ms NFR — recommend Caramba-style "fire switch on first tap; second tap within window corrects the pre-switch buffer; triple cancels pending double and corrects the phrase," with release-based decision points; (3) replacement capability ladder per app class (Q2); (4) buffer-reset triggers spec-delta — mouse clicks are NOT observable at the IME layer; redefine as Enter/Tab/Escape/focus-out/`Reset()`/caret-jump-inferred; (5) per-app macro scope — global with kill-switch for v1, `FocusInfo` seam for later.
**Avoids:** Pitfalls 2, 3, 7, 8, 9, 10, 12 — all of which map "decide at M0" as their prevention phase.

**Decision #1 — layout-switching mechanism (THE architectural fork):**
- **Option A (ARCHITECTURE, recommended primary):** two engines in one component (`goswitch-en` layout=us, `goswitch-ru` layout=ru) as `('ibus',…)` sources. Verified from gnome-shell source: activating a source applies the engine's declared XKB layout *and* routes keys to goswitchd. Common path is pure passthrough; layout observation is free (FocusIn/FocusOut between our engines); native indicator; MRU cycling preserved. **Open sub-question:** "switch" then requires a *programmatic* source change — `gsettings set current N` — whose runtime efficacy on GNOME 46 is exactly what PITFALLS disputes (fallbacks: `org.gnome.Shell.Eval`, a tiny shell extension, or fold back to internal flip).
- **Option B (PITFALLS/owner's prototype — proven working today):** single engine declaring layout=us with an internal en/ru mode flip. Instant switching with no GNOME round-trip; but per-keystroke `commit_text` in RU mode (ARCHITECTURE's documented anti-pattern: latency on 100% of typing to serve the correction path, broken app composition) and the indicator shows the engine symbol, not per-layout state.
- **Resolution:** ADR at M0 + kill-criterion spike at the start of M1 — test two-engine activation via manual Super+Space, then empirically test programmatic switching (gsettings write first). If programmatic switching is unreliable, adopt Option B without architectural damage (buffer/correction logic is identical; isolate the choice behind the `layoutset` interface). Note the conflicting evidence on `gsettings current` (STACK: writable on GNOME 46, lore outdated; PITFALLS/SE-316998: ignored at runtime, pre-GNOME-46 era) — only a live test on the target settles it.

### Phase 1: Engine Skeleton & Architecture Validation (spec M1)

**Rationale:** Everything depends on the engine seeing keys as the active source; this is where Q1/Q4 are retired or the fallback is adopted cheaply. Pure packages first (zero integration risk, CI-green day one), then the wire adapter + minimal e2e.
**Delivers:** `layouts/` generated tables + golden tests; `internal/hotkey` FSM with unit corpus; `engine/` adapter with address discovery, registration, Factory, reconnect/re-register loop, **recover() shim from the first commit**; two-engine registration spike (Decision #1 kill-criterion); systemd user unit (dev form); e2e skeleton (ydotool → engine log) with self-provisioning and fail-fast diagnostics.
**Gate:** hotkey events visible in the log *while goswitch is the GNOME-active source* (not only standalone); `ibus restart` and engine kill -9 both leave desktop typing alive.
**Addresses:** engine interception MVP item; e2e matrix foundation.
**Avoids:** pitfalls 1, 2, 4, 5 (dev loop), 11 (skeleton), 12.

### Phase 2: Word Correction EN↔RU (spec M2)

**Rationale:** The single most-used gesture and the validation of the core architectural bet (exact-range replacement). Buffer/reset semantics gate the phrase feature, so word ships first along this natural seam.
**Delivers:** per-IC typed-text buffer keyed on keycode with release filtering; case-preserving conversion + direction auto-detect; replacement capability ladder tiers 1–2 (`DeleteSurroundingText`+`CommitText`, BackSpace×N fallback with rune counting and abort-not-garbage); single/double-tap wiring with tuned windows; `internal/correct` + session actor integration; e2e matrix v1 (gnome-text-editor + one Chromium case) with AT-SPI readback, per-case state reset, and failure output of observed text.
**Addresses:** correct-last-word, tables, case, direction, exact-range replacement, buffer reset rules (empirical per-client check becomes an e2e case).
**Avoids:** pitfalls 6, 7 (tiers 1–2), 9 (tuning), 10 (empirical click/reset verification).

### Phase 3: Phrase, Selection, Config & Switching (spec M3)

**Rationale:** Second-tier gestures plus the configurability layer; switching integration lands after correction because it is orthogonal to the core value and depends on Decision #1's outcome.
**Delivers:** triple-tap phrase correction; selection correction (surrounding-text anchor where caps allow, clipboard transport otherwise — built once, shared with the future terminal fallback); mixed-text semantics per ADR with golden corpus; YAML config + hot reload + `goswitchctl` + `ctlsvc`; layout switching integrated per the M0 ADR (two-engine gsettings path or internal flip); switch+correct combo.
**Addresses:** phrase/selection/multi-tap/combo/config/switching MVP items; mixed text (P1 semantics / P2 full — may slip to "whole buffer converts by majority" with documentation if needed).
**Avoids:** pitfalls 3 (mechanism behind interface), 8 (`FocusInfo` seam), hot-reload traps (subprocess-per-event, parse-per-event).

### Phase 4: Install, Docs & Acceptance (spec M4)

**Rationale:** Hardening and the no-root story, once behavior is stable and there is something worth installing.
**Delivers:** no-root installer (component XML under `$HOME`, `IBUS_COMPONENT_PATH` via `environment.d`, `ibus write-cache` + verify `ibus list-engine`, self-check, uninstaller); final systemd user unit(s) per the supervision ADR; full e2e matrix green **twice in a row** on an untouched session (gnome-text-editor, gedit, Chrome; per-capability-tier coverage incl. Chromium windowing modes); performance assertions (p95 <50 ms via ydotool→AT-SPI timestamps, RSS <50 MB); docs, README, monthly canary job.
**Addresses:** no-root install MVP item; the e2e acceptance gate itself.
**Avoids:** pitfalls 5 (installer dance), 11 (rot/canary), 12 (final units, double-spawn lock).

### Post-v1 track (v1.x/v2 — do not schedule inside v1)

Terminal fallback via wl-clipboard (after Q3 investigation, per-app enablement); Super→Ctrl per-app macros (needs the app-identity spike; candidate for spec-delta to v2); exception heuristics; performance telemetry in the e2e report; additional layout pairs.

### Phase Ordering Rationale

- **Dependency chain:** pure packages (layouts, hotkey FSM) unblock everything with zero integration risk; engine adapter + e2e skeleton is the M1 spike that retires the architecture questions while the fallback is still cheap; correction before switching (core value vs orthogonal integration); config after behavior (reload semantics stabilize once something is worth reloading); install last.
- **Architecture grouping:** each phase completes a horizontal slice of the wire-adapter → actor → pure-logic stack, so the headless-CI fraction is maximal from day one and every phase's e2e growth matches spec §7.2 sequencing (M1 event-in-log → M2 word matrix → M3 full matrix).
- **Pitfall avoidance:** the M0 decision pack exists because pitfalls 2/3/7/8/10/12 all name "M0 ADR" as their prevention phase; the recover() shim and reconnect loop are day-one M1 items, not retrofits; the kill-criterion spike makes the godbus dead-end (pitfall 4) cheap now, expensive later.

### Research Flags

Phases likely needing deeper research during planning (`/gsd:plan-phase --research-phase`):
- **Phase 1 (M1):** the engine adapter is niche-domain territory — private-bus handshake details, `UpdatePreeditText` arity on installed 1.5.29-rc2, `DeleteSurroundingText` signal acceptance, two-engine activation specifics; monitor traffic with dbus-monitor on the IBus socket.
- **Phase 2 (M2):** per-app capability matrix (GTK4 vs Firefox vs Chromium/Electron windowing modes) decides fallback prevalence; empirical `reset()`-on-click behavior per client class.
- **Phase 3 (M3) — partial:** mixed-text semantics corpus design and (if macros are kept in scope) the app-identity spike; the rest of M3 is standard patterns.

Phases with standard patterns (skip research-phase):
- **Phase 4 (M4):** installer/units/documentation — PITFALLS.md itself is the checklist; mechanisms are documented and verified.
- Pure packages inside Phase 1 (layouts generation, hotkey FSM) and config hot reload inside Phase 3 — conventional, well-documented Go patterns.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Verified against the Go module proxy and the live target system; versions pinned; license blocker on goibus confirmed via GitHub API |
| Features | MEDIUM | Cross-checked across 10+ competitors; owner is the sole user (dogfooding), so risk is semantics, not demand; single-source claims marked in FEATURES.md |
| Architecture | HIGH | Live D-Bus introspection of IBus 1.5.29 on target, end-to-end engine activation performed during research, gnome-shell + ibus upstream sources read directly |
| Pitfalls | MEDIUM-HIGH | Upstream issues + empirical inspection of the owner's machine and prototype; several claims version-dependent (flagged inline) |

**Overall confidence:** HIGH for the core bet (Go IBus engine viable, stack minimal and pinned, architecture verified live); MEDIUM for per-app behavioral details and the GNOME-switching mechanism, both of which have explicit in-phase validation planned.

### Gaps to Address

- **Layout-switching mechanism (Decision #1):** conflicting evidence between two-engine pattern (gnome-shell source) and single-engine internal flip (working prototype), hinging on whether `gsettings set current` switches sources at runtime on GNOME 46. Handle: M0 ADR + M1 kill-criterion spike; keep both options implementable behind `internal/layoutset`.
- **Multi-tap vs <50 ms:** semantics must be pinned in requirements (recommend fire-on-first-tap) before Phase 2 wires the FSM to actions.
- **goibus characterization discrepancy:** STACK/ARCHITECTURE treat it as pure-godbus reference; PITFALLS calls it cgo. Immaterial to implementation (unlicensed either way, in-repo adapter required regardless; the live Python prototypes + godbus-in-production settle Q1 viability independently).
- **Chromium/Electron surrounding-text support:** version-dependent (ibus#2354); resolve empirically in the Phase 2 capability matrix; one e2e case per claimed windowing mode.
- **Terminals (Q3):** deferred to v1.x; detection mechanism (per-app context from focus events) lands with Phase 3's `FocusInfo` seam.
- **Signal-arity details on the installed rc:** `UpdatePreeditText` 4-arg form and `DeleteSurroundingText` no-ack behavior to be confirmed with dbus-monitor during Phase 1.
- **ydotool provisioning on other machines:** target's distro 0.1.8 needs no daemon (locally verified, HIGH); PITFALLS' ydotoold/socket-perms guidance applies to upstream 1.0.x — the e2e stand should auto-detect and fail fast either way; CI runner needs a self-hosted GNOME session.
- **Mouse-click buffer reset:** unobservable at the IME layer; spec-delta at M0, approximate via caret-jump inference, verify per client in Phase 2.

## Sources

### Primary (HIGH confidence)
- Live target-system verification (Ubuntu 24.04 / GNOME 46 / IBus 1.5.29-rc2, 2026-09-10): D-Bus introspection of ibus-daemon over its private socket, end-to-end engine registration + activation, `gsettings` probes, systemd user units, component XMLs, `/usr/include/ibus-1.0/` headers, dpkg versions (ydotool 0.1.8, python3-gi 3.48.2, at-spi 2.52.0)
- Owner prototypes: `/home/nil/.local/share/punto-switcher/punto_engine.py` (full working Python IBus engine), `/home/nil/.local/share/ibus-test/test_shift.py` — prior art for handshake, tap timing, replacement fallbacks
- Go module proxy — godbus v5.2.2, fsnotify v1.10.1, yaml.v3 v3.0.1; GitHub API license check (goibus: none)
- Upstream source read directly: ibus/ibus (`ibusengine.c`, `ibustypes.h`), gnome-shell (`js/ui/status/keyboard.js`), sarim/goibus

### Secondary (MEDIUM confidence)
- ibus issues: #2910 (registrations lost on daemon restart), #2354 (Chromium DeleteSurroundingText broken), #2054 (Firefox surrounding text), #2600 (fast-shift modifier loss), #2423 (surrounding-text staleness)
- Competitor primary READMEs/docs: WaylandSwitcher, Mahou, SimpleSwitcher, langSwitcher, autokbisw; Caramba (App Store + AutoHotkey forum); Grokipedia Punto Switcher; wezterm `use_ime` docs + issues #1772/#5125
- TUX IM dev log (2026) — engine crashes killing ibus-daemon; press/release doubling
- ydotool version/daemon facts (RH bugzilla 2250692, Debian manpage) — superseded by local dpkg where they conflict
- systemd.io desktop integration; graphical-session.target design materials; godbus signal-channel-close semantics

### Tertiary (LOW confidence)
- AskUbuntu / unix.stackexchange lore on `gsettings current` (read-only / runtime-ignored) — pre-GNOME-46 era, partially superseded by live writability verification; runtime effect still unverified (Gap #1)
- DeepWiki ibus pages, venam.net input-stack overview, wayland-input blog, Fedora Magazine AT-SPI automation — corroborating context only

---
*Research completed: 2026-09-10*
*Ready for roadmap: yes*
