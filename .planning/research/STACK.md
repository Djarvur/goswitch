# Stack Research

**Domain:** IBus input-method-engine daemon (layout switcher / mistyped-text corrector) for GNOME Wayland, written in Go
**Researched:** 2026-09-10
**Confidence:** HIGH for core stack (verified against the Go module proxy — the authoritative version registry — and against the live target system: Ubuntu 24.04, GNOME 46, IBus 1.5.29, ydotool 0.1.8, Go 1.27.1, where two validated Python IBus-engine prototypes exist and one was the *active* engine during research)

## What This Stack Must Satisfy

Fixed constraints from `PROJECT.md` / `SPEC.md` §5: Go ≥ 1.23, stdlib + `godbus/dbus` + minimal deps, no root for install/operation, systemd user unit, hot-reloadable YAML config, e2e testing via physical input path (uinput) + AT-SPI reading. Spec open questions directly addressed by this file: **Q1** (IBus engine viability in Go) and **Q6** (autostart scheme); stack-relevant notes for Q2/Q3/Q4 at the end.

---

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | ≥ 1.23 (go.mod), build with current 1.27.x | Language/toolchain | Project constraint. Local toolchain is go1.27.1. Keep `go 1.23` as the module language version so prebuilt release binaries and `go install` remain compatible with older distro toolchains. `log/slog` (stdlib since 1.21) covers the structured/leveled logging requirement of SPEC §5 with zero deps. |
| `github.com/godbus/dbus/v5` | **v5.2.2** (2025-12-29; verified via `proxy.golang.org`) | All D-Bus: IBus engine socket (service side), session bus, a11y bus if needed | The only maintained native Go D-Bus implementation (BSD-2-Clause, ~1.2k stars). Needed in both client *and service* mode: `Export` publishes the engine/factory object paths on the IBus connection, signal channels receive `NameOwnerChanged` for daemon-restart detection. Proven against the IBus socket in production by ibus-bamboo (built on godbus v5.1.0, actively maintained 2026, packaged in Fedora by Red Hat's IBus maintainer). API stable across v5.x. |
| In-repo `engine/` package (IBus protocol adapter) — **no external dependency** | n/a (approx. 800–1,200 LOC) | Implements `org.freedesktop.IBus` engine side: `_RegisterComponent`, factory `CreateEngine`, engine object with `ProcessKeyEvent`, signals `CommitText` / `ForwardKeyEvent` / `UpdatePreeditText` / `DeleteSurroundingText` / `RequireSurroundingText` | The only existing Go bindings (`sarim/goibus` 2018, fork `BambooEngine/goibus` 2024) have **NO LICENSE FILE** in either repo (verified via GitHub API) — legally unusable inside an MIT project. The protocol surface needed is small (goibus is 832 LOC total). Primary sources for the wire format are installed on the target system: `/usr/include/ibus-1.0/*.h` (ibusregistry.h, ibuscomponent.h, ibusengine.h, ibustext.h) plus the owner's two validated Python prototypes (see below). This is the single real piece of engineering in the stack; everything else is off-the-shelf. |
| `gopkg.in/yaml.v3` | **v3.0.1** (2022-05-27; frozen but stable) | YAML config parsing | Ecosystem-default YAML 1.2 parser, zero external dependencies, known-DoS issues fixed in v3.0.1. Frozen status is a *feature* under the minimal-dep/audit constraint: no churn, tiny audit surface. Hot reload is orthogonal (fsnotify's job). Alternative: `github.com/goccy/go-yaml` v1.19.2 (2026-01-08, active, also zero-dep) if stricter typing or better error messages become desirable — swap is mechanical. |
| `github.com/fsnotify/fsnotify` | **v1.10.1** (2026-05-04; verified via `proxy.golang.org`) | Config hot reload (SPEC §4.1 "without restart") | The standard Go inotify wrapper. Must be used with the directory-watch + debounce pattern (see Patterns section) because editors save config files via atomic rename, which swaps the inode. |
| systemd **user** units | systemd 255 (Ubuntu 24.04) | Autostart / supervision (Q6) | The distro itself starts ibus-daemon this way: `org.freedesktop.IBus.session.GNOME.service` (`Type=dbus`, `BusName=org.freedesktop.IBus`, `WantedBy=gnome-session.target`, `After=gnome-session-initialized.target`, `Restart=on-abnormal` — read from `/usr/lib/systemd/user/` on the target). `goswitchd.service` (user) hooks into the same chain. No root required anywhere. |

### The IBus Protocol Layer (the load-bearing detail)

Verified facts that shape the implementation, all confirmed on the target system (IBus 1.5.29) unless noted:

1. **Connection handoff:** ibus-daemon does *not* serve the engine API on the session bus (probed: `gdbus call ... /org/freedesktop/IBus` on the session bus → "Object does not exist"). It embeds its own D-Bus server on a dedicated unix socket whose address is published in `~/.config/ibus/bus/<machine-id>-unix[-wayland]-<display>` files and via the `ibus address` CLI (returns `unix:path=/home/nil/.cache/ibus/dbus-XXXX,guid=...`). `goswitchd` reads that file (or `IBUS_ADDRESS` env override) and `godbus` dials it directly — exactly what `IBus.Bus()` does in Python and what goibus/ibus-bamboo do from Go.
2. **No introspection:** the IBus connection publishes no standard introspection data (probed: empty node). Method names, signatures, and the GVariant serialization of IBus objects (`IBusText` = struct with string + attribute array, etc.) must be hand-coded. Truth sources, in order: `/usr/include/ibus-1.0/` headers (installed), `ibus.github.io/docs/ibus-1.5` reference manual, the owner's validated prototypes at `/home/nil/.local/share/punto-switcher/punto_engine.py` and `/home/nil/.local/share/ibus-test/test_shift.py` (both register, receive `process_key_event` with keyval/keycode/state, and commit text successfully on this machine — `ibus engine` returned `test-shift`, i.e. a prototype was the live engine), and goibus source as a cross-check only.
3. **Registration, two paths:** (a) component XML in a scanned directory — how the distro engines do it (`/usr/share/ibus/component/*.xml` verified: `<component>` with name/exec/version/license + nested `<engines><engine>` with name/language/layout/symbol/rank); `IBUS_COMPONENT_PATH` (colon-separated) adds custom dirs, documented in the installed `ibusregistry.h` — but it must be set in *ibus-daemon's* environment, i.e. a user-unit override, still root-free; (b) **programmatic `_RegisterComponent` at daemon startup** — what both Python prototypes do (connect → request component name → `_RegisterComponent`) and the recommended primary path for goswitchd: zero files outside `$HOME`, immediately active, matches the no-root constraint.
4. **Key delivery:** the engine receives `ProcessKeyEvent(keyval, keycode, state) → bool` (return `true` = consumed). Keyvals are X11 keysyms — a static Go table generated by `go:generate` (SPEC §8 `layouts/`), no runtime library needed. Modifier/release bit (0x40000000) semantics demonstrated in the prototypes.
5. **Text replacement surface (Q2-relevant):** engine signals `CommitText(text)` and `DeleteSurroundingText(offset_from_cursor, nchars)`; the client pushes `SetSurroundingText(text, cursor, anchor)` when the app's IM module supports it and the engine called `RequireSurroundingText`. The punto prototype's `delete_surrounding_text(-n, n)` + fallback to forwarding Backspace is the validated pattern; actual per-app precision (GTK vs Chromium vs terminals) is phase-research territory, not a stack dependency.

### GNOME Integration (layout read/switch)

| Mechanism | Status on target | Use |
|---|---|---|
| `gsettings get/set org.gnome.desktop.input-sources current` | **verified writable** (`gsettings writable ... current` → `true`; `sources` = `[('xkb','us'), ('xkb','ru')]`) | Primary layout switch/read: subprocess `gsettings` via `os/exec` — zero new deps, robust against dconf backend details. The AskUbuntu lore that `current` is read-only is outdated for GNOME 46 — verified live. |
| `gsettings monitor org.gnome.desktop.input-sources` | std subprocess | Reactive layout-change observation: parse its line output. (The alternative — direct `ca.desrt.dconf` D-Bus — is fragile and buys nothing.) |
| `ibus engine <name>` / `org.freedesktop.IBus.SetGlobalEngine` | available | Fallback/activation of the goswitch engine itself. |
| GNOME keybindings | `switch-input-source` = `['<Alt>Shift_L','<Super>space']` (verified) | Reference only; the engine intercepts its own hotkeys inside `ProcessKeyEvent`, no keybinding registration needed. |

### Supporting Libraries

None beyond the two above — that is the point of the constraint. Deliberate non-additions:

| Need | Resolution without a new dependency |
|---|---|
| Structured logging (SPEC §5) | stdlib `log/slog` |
| CLI for `goswitchctl` | stdlib `flag` (or `os.Args` dispatch); a CLI framework would violate the constraint |
| Layout tables EN↔RU | static data + `go:generate` generator (SPEC §8); no xkb parser needed for ЙЦУКЕН↔QWERTY |
| Asserts in unit tests | stdlib `testing` + golden files (SPEC §7.1); add `stretchr/testify` only if the owner relaxes the constraint — not recommended now |
| Terminal fallback (Q3, v1-optional) | `wl-copy`/`wl-paste` subprocess — no Go Wayland client dep |

### Development Tools (e2e stand, SPEC §7.2–7.3)

| Tool | Version on target | Purpose | Notes |
|------|-------------------|---------|-------|
| ydotool | **0.1.8-3build1** (Ubuntu 24.04 archive) | Physical-path key injection | Distro 0.1.8 opens `/dev/uinput` directly — no `ydotoold` daemon needed (verified: no socket/service on target, yet the binary is installed and `/dev/uinput` exists root:input 0660). Test user is already in the `input` group (verified). Do **not** chase upstream 1.0.4: it requires the ydotoold daemon + socket setup and is not packaged in 24.04. Fallback if injection semantics prove insufficient: a ~150-LOC Go uinput injector via stdlib `syscall` ioctls. |
| AT-SPI reader | `python3-gi` 3.48.2, `gir1.2-atspi-2.0` 2.52.0, `at-spi2-core` 2.52.0 — all installed (verified) | Reading editor content + focusing windows (`org.a11y.atspi.Text.getText`, `Component.grabFocus`) | SPEC §7.3 already prescribes `at-spi` via python3-gi. There is **no established standalone Go AT-SPI library** (searched; only `m31labs.dev/fluffyui`'s atspi package, part of a heavier UI-automation module — LOW confidence, avoid). Pragmatic split: e2e orchestrator and YAML case matrix in Go; text-reading/focusing step as a small `/usr/bin/python3` helper script using gi. Note: unlike IBus, **AT-SPI2 is fully introspectable D-Bus** (`org.a11y.Bus` on the session bus hands out the a11y bus address), so a pure-Go godbus reader remains a clean later option. |
| `gdbus` / `busctl --user` / `gsettings` | ships with the OS | Manual verification of every D-Bus claim during development | Used to produce the verifications in this file. |
| `ibus engine`, `ibus list-engine`, `ibus address`, `ibus restart` | IBus 1.5.29 CLI | Debugging engine state, simulating daemon restarts | Used extensively in this research. |
| GitHub Actions | unit on push (headless), e2e on manual dispatch with a self-hosted GNOME runner | CI per SPEC §8 | e2e runner needs: GNOME session, user in `input` group, python3-gi + at-spi gir packages. |

## Installation

```bash
# Go dependencies — the complete external dependency set
go get github.com/godbus/dbus/v5@v5.2.2
go get gopkg.in/yaml.v3@v3.0.1
go get github.com/fsnotify/fsnotify@v1.10.1

# e2e stand packages (Ubuntu 24.04, once per machine)
sudo apt install ydotool python3-gi gir1.2-atspi-2.0 at-spi2-core
sudo usermod -aG input "$USER"   # uinput access for injection (already true on target)
```

Autostart (Q6) — user units, no root:

```ini
# ~/.config/systemd/user/goswitchd.service
[Unit]
Description=goswitch IBus engine daemon
PartOf=graphical-session.target
After=graphical-session.target org.freedesktop.IBus.session.GNOME.service
# ibus-daemon may not be up yet, or may restart out from under us:
Restart=on-failure
RestartSec=2

[Service]
ExecStart=%h/go/bin/goswitchd
# daemon internal loop: connect to ibus socket -> request name ->
# _RegisterComponent -> (re)activate engine; retry until ibus-daemon answers

[Install]
WantedBy=graphical-session.target
```

`systemctl --user enable --now goswitchd` completes the install. Component XML in `/usr/share/ibus/component/` is *not* required (programmatic registration); it becomes relevant only for later distro packaging, or set `IBUS_COMPONENT_PATH=%h/.local/share/ibus/component` via a user-unit override on `org.freedesktop.IBus.session.GNOME.service` if persistent XML-based registration is ever wanted.

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| godbus/dbus v5.2.2 | `linuxdeepin/go-dbus` (2014), `carnegierobotics/go-dbus` (stale godbus v4 fork) | Never — both abandonware (module proxy: 2014 / dead). |
| In-repo `engine/` adapter | `BambooEngine/goibus` v0.0.0-20240724 | Never as an import (no LICENSE — legal blocker for MIT code). Legitimate as a *read-only behavioral reference* for the GVariant wire format. |
| In-repo `engine/` adapter | cgo bindings to `libibus` (`gir1.2-ibus-1.0` + gotk3 stack) | Never for this project — cgo breaks the static-binary/audit story and drags in GLib. |
| `gopkg.in/yaml.v3` | `github.com/goccy/go-yaml` v1.19.2 | If config validation needs its strict decoder, better line/col errors, or comment round-tripping (comment-preserving config rewrites in v2+). Both are zero-dep; migration is mechanical. |
| `gsettings` subprocess | direct `ca.desrt.dconf.Writer` D-Bus calls via godbus | Only if subprocess overhead in the hot path ever measurably matters — unlikely (switch actions are user-paced, <50 ms budget is generous). |
| ydotool 0.1.8 (distro) | own Go uinput injector | If ydotool's timing/behavior proves flaky in CI; ~150 LOC of stdlib ioctls is a contained fallback. Upstream ydotool 1.0.4 + ydotoold is a third option only on distros that package it. |
| python3-gi reader helper | pure-Go AT-SPI client on godbus | If the e2e stand must shed the python dependency; AT-SPI is introspectable so godbus generated/hand stubs are viable. Keep python for v1 — it is installed, spec-prescribed, and the least code. |
| fsnotify v1.10.1 | polling `os.Stat` mtime every 500 ms | Genuinely acceptable for a single config file (simpler than handling rename semantics); keep fsnotify only because it is already tiny and standard. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `sarim/goibus` / `BambooEngine/goibus` as a dependency | **No LICENSE file in either repo** (verified via GitHub API) — default "all rights reserved"; cannot ship inside MIT-licensed goswitch. Also pins godbus v5.1.0 and is idle since 2024-07. | In-repo `engine/` package; read their source only as protocol reference. |
| Viper (spf13/viper) for config | ~10+ transitive deps (including its own YAML/FS abstraction layers) — direct violation of the "stdlib + godbus + minimal deps" constraint for a config file with 3 concerns (keys, mappings, hot reload). | yaml.v3 + fsnotify (~50 LOC of glue). |
| cgo / libibus / GLib bindings | Loses static cross-compilable binary, complicates `go install` distribution, audit surface balloons. | godbus pure-Go wire protocol. |
| godbus/dbus **v4** (2017) or un-versioned import path | EOL, missing v5 fixes (split of introspect/propgen, `dbus.Variant` typing). | `github.com/godbus/dbus/v5` module path. |
| xdotool for e2e | X11-only — dead on Wayland (would only "work" via XWayland apps, invalidating the test premise). | ydotool (uinput = compositor-independent physical path). |
| uinput injection **inside goswitchd** for text correction | Architecture axiom (SPEC §3.1): uinput conflicts with keyd/xremap at the evdev layer and breaks the no-monopoly requirement. IBus `commit_text` is the injection mechanism. | IBus engine signals; uinput stays in the e2e stand only, where it *simulates the human*. |
| GUI settings frameworks / GTK | Out of scope v1 (YAML suffices, SPEC §10); any GUI dep would dwarf the daemon itself. | YAML + `goswitchctl`. |
| `stretchr/testify` (optional stance) | Test-only sugar not worth breaking the "auditable in an afternoon" property; golden-file tests are more valuable for layout tables anyway. | stdlib `testing` + golden files. |

## Stack Patterns by Variant

**Config hot reload (atomic-rename trap):**
Watch the config file's *parent directory* with fsnotify, match events against the file name, debounce 150–300 ms, then re-parse fully and swap atomically (`atomic.Pointer[Config]`). Re-adding the watch on rename is not needed when watching the directory. Rationale: vim/kwrite/gedit save via `write temp + rename`, which replaces the inode a file-path watch points at.

**ibus-daemon restart resilience (Q6 core risk):**
Registrations die with ibus-daemon and the global engine reverts to default (IBus 1.5.29, issue ibus/ibus#2910, closed as not planned — no upstream fix coming). Pattern: `goswitchd` watches the IBus connection (godbus signal `NameOwnerChanged`, plus socket-file replacement detection); on loss it re-enters the connect → `_RegisterComponent` → reactivate loop with ~2 s retries, mirroring the workaround validated in that issue.

**Terminal applications (Q3):**
If IBus preedit/commit misbehaves in specific terminals (wezterm), the fallback is `wl-copy`/`wl-paste` subprocess + `ForwardKeyEvent` paste emulation — the exact pattern in the punto prototype's `_convert_selection`. No Go Wayland dependency required.

**Hotkey layer (Q4):**
The engine *is* the hotkey layer: while goswitch is the active input source, all keys traverse `ProcessKeyEvent` with full modifier state (proven by the shift-capture prototype logging every event). Cost: goswitch must remain the active engine across layouts, which the "engine as the only input source, layouts handled inside the engine" model (used by the punto prototype) provides without any extra dependency.

**If distro packaging arrives later (beyond v1):**
Install component XML to `/usr/share/ibus/component/goswitch.xml` (schema verified from live distro files), run `ibus write-cache`, ship a system package — the runtime code does not change.

## Version Compatibility

| Component | Version | Compatible With | Notes |
|-----------|---------|-----------------|-------|
| `github.com/godbus/dbus/v5` | v5.2.2 | Go 1.17+; IBus 1.5.x wire protocol | go.mod `go` directive satisfies 1.23 constraint. ibus-bamboo runs godbus v5.1.0 in production 2026 — v5.x API stable; protocol unchanged 2018→2026 (goibus worked across that span unmodified). |
| `gopkg.in/yaml.v3` | v3.0.1 | Go 1.11+ | Last release 2022 (maintenance mode); no known open CVEs post-v3.0.1; zero transitive deps. |
| `github.com/fsnotify/fsnotify` | v1.10.1 | Go 1.17+, Linux/inotify | Linux-only usage; no macOS/Windows concerns for this daemon. |
| IBus protocol (target) | 1.5.29 (Ubuntu 24.04) | engines written against any 1.5.x | `IBUS_COMPONENT_PATH` documented in installed `ibusregistry.h`; `_RegisterComponent` takes a variant-wrapped serialized component (arg type `v` confirmed in daemon strings). |
| ydotool | 0.1.8 (24.04 archive) | kernel uinput; needs `input` group | Upstream 1.0.4 has a different daemon architecture — do not mix docs/versions. |

## Direct Answers Feeding the Roadmap

- **Q1 — IBus engine on Go: VIABLE, HIGH confidence.** Evidence chain: (1) goibus/ibus-bamboo proves Go engines on modern distros (Fedora-packaged 2025–2026); (2) godbus speaks the IBus socket protocol in production today; (3) on this exact machine, Python prototypes registered via pure D-Bus calls and received/committed keys — the only Go-specific work is the in-repo protocol adapter (~1k LOC) because the existing bindings are unlicensed. Risks are behavioral (per-app surrounding-text support), not stack-level.
- **Q6 — Autostart: systemd user unit + programmatic re-registration, HIGH confidence.** Unit hangs off `graphical-session.target` after the distro's IBus unit; registration is a runtime D-Bus call (no root, no files outside `$HOME`); daemon-restart resilience is mandatory because 1.5.29 loses registrations by design (#2910).

## Sources

- Go module proxy (authoritative version registry) — godbus v5.2.2, fsnotify v1.10.1, yaml.v3 v3.0.1, goccy v1.19.2, BambooEngine/goibus v0.0.0-20240724 — **HIGH** (primary registry)
- Local target-system verification (Ubuntu 24.04 / GNOME 46 / IBus 1.5.29): component XMLs in `/usr/share/ibus/component/`, systemd units in `/usr/lib/systemd/user/`, `ibus address` output, `gsettings writable ... current`, `dpkg` versions (ydotool 0.1.8, python3-gi 3.48.2, gir1.2-atspi-2.0 2.52.0), `/usr/include/ibus-1.0/ibusregistry.h`, D-Bus probes of the IBus connection, live prototypes `/home/nil/.local/share/punto-switcher/punto_engine.py` and `/home/nil/.local/share/ibus-test/test_shift.py` — **HIGH** (the target runtime itself)
- [github.com/godbus/dbus](https://github.com/godbus/dbus) + [releases](https://github.com/godbus/dbus/releases) — cross-check of v5.2.2 — **MEDIUM** (web, cross-verified)
- [sarim/goibus](https://github.com/sarim/goibus), [BambooEngine/ibus-bamboo](https://github.com/BambooEngine/ibus-bamboo), [BambooEngine org](https://github.com/BambooEngine), [Fedora packaging](https://src.fedoraproject.org/rpms/ibus-bamboo) — production-Go-engine evidence; GitHub API license check (both goibus repos: license = None) — **MEDIUM→HIGH** (API-verified)
- [IBus 1.5 reference manual](http://ibus.github.io/docs/ibus-1.5/) ([IBusInputContext](http://ibus.github.io/docs/ibus-1.5/IBusInputContext.html)), [IBus.Engine class docs](https://lazka.github.io/pgi-docs/IBus-1.0/classes/Engine.html) — method/signal names — **MEDIUM** (cross-verified against headers + prototypes)
- [ibus/ibus#2910 — global engine not persisted across daemon restarts](https://github.com/ibus/ibus/issues/2910) — Q6 re-registration requirement — **MEDIUM** (single issue, closed as not planned; behavior locally plausible, re-test in M1)
- [ydotool on Ubuntu/Fedora Wayland](https://github.com/ReimuNotMoe/ydotool/issues/285), [Debian ydotool 1.0.4-3 manpage](https://manpages.debian.org/testing/ydotool/ydotool.1.en.html) — version/daemon differences; resolved authoritatively by local `dpkg` (0.1.8) — **MEDIUM**
- [AT-SPI2 project](https://www.freedesktop.org/wiki/Accessibility/AT-SPI2/), [Ubuntu AT-SPI D-Bus interface docs](https://ubuntu.com/desktop/docs/en/latest/reference/accessibility/dbus/), [fluffyui atspi package](https://pkg.go.dev/m31labs.dev/fluffyui) — no established Go AT-SPI lib; introspectable D-Bus route — **MEDIUM** (negative claim cross-checked across searches + local gir packages)
- [AskUbuntu — switch layout by CLI](https://askubuntu.com/questions/1176703/switch-between-keyboard-languages-by-cli-in-gnome) — superseded by local `gsettings writable` verification (true on GNOME 46) — **LOW for the lore, HIGH for the locally verified fact**

---
*Stack research for: GNOME Wayland IBus layout-switcher daemon in Go*
*Researched: 2026-09-10*
