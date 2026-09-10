<!-- GSD:project-start source:PROJECT.md -->

## Project

**goswitch**

`goswitch` — корректор раскладки и введённого текста для GNOME Wayland (Ubuntu 24.04, T2 MacBook Pro): исправление текста, набранного не в той раскладке (EN↔RU), и переключение раскладки по горячим клавишам. Реализуется как собственный демон на Go, работающий в роли **IBus input method engine** через D-Bus. Полное техническое задание: `docs/SPEC.md` (SDD фаза 1, владелец Daniel Podolsky).

**Core Value:** По горячей клавише исправить текст, набранный не в той раскладке (EN↔RU), в любом поле ввода GNOME Wayland — через IBus engine, без root и без конфликтов с ремапперами (keyd/xremap).

### Constraints

- **Tech stack**: Go ≥ 1.23, только stdlib + `godbus/dbus` + минимальные зависимости — простота поставки и аудита
- **Architecture**: Инъекция текста через IBus engine (`commit_text`), НЕ через uinput — не конфликтует с evdev-ремапперами, работает в любом поле ввода GTK/Qt/Chromium/Electron
- **Architecture**: Один демон `goswitchd` — IBus engine + D-Bus сервис + наблюдение за раскладкой
- **Permissions**: Демон без root — членство в группе доступа IBus-сокета; systemd user unit
- **Performance**: Реакция на горячую клавишу < 50 мс; память < 50 МБ
- **Process**: SDD (spec-driven development), трёхфазная модель; изменения требований через spec-delta
- **Quality**: Автоматизированное тестирование полного цикла без участия человека; e2e требует сам активировать целевое окно перед кейсом
- **Licensing**: MIT, репозиторий `Djarvur/goswitch` public; conventional commits, основные ветки защищены, изменения через PR

<!-- GSD:project-end -->

<!-- GSD:stack-start source:research/STACK.md -->

## Technology Stack

## What This Stack Must Satisfy

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

### GNOME Integration (layout read/switch)

| Mechanism | Status on target | Use |
|---|---|---|
| `gsettings get/set org.gnome.desktop.input-sources current` | **verified writable** (`gsettings writable ... current` → `true`; `sources` = `[('xkb','us'), ('xkb','ru')]`) | Primary layout switch/read: subprocess `gsettings` via `os/exec` — zero new deps, robust against dconf backend details. The AskUbuntu lore that `current` is read-only is outdated for GNOME 46 — verified live. |
| `gsettings monitor org.gnome.desktop.input-sources` | std subprocess | Reactive layout-change observation: parse its line output. (The alternative — direct `ca.desrt.dconf` D-Bus — is fragile and buys nothing.) |
| `ibus engine <name>` / `org.freedesktop.IBus.SetGlobalEngine` | available | Fallback/activation of the goswitch engine itself. |
| GNOME keybindings | `switch-input-source` = `['<Alt>Shift_L','<Super>space']` (verified) | Reference only; the engine intercepts its own hotkeys inside `ProcessKeyEvent`, no keybinding registration needed. |

### Supporting Libraries

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

# Go dependencies — the complete external dependency set

# e2e stand packages (Ubuntu 24.04, once per machine)

# ~/.config/systemd/user/goswitchd.service

# ibus-daemon may not be up yet, or may restart out from under us:

# daemon internal loop: connect to ibus socket -> request name ->

# _RegisterComponent -> (re)activate engine; retry until ibus-daemon answers

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

<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->

## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->

## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->

## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, `.github/skills/`, or `.codex/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->

## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:

- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->

<!-- GSD:profile-start -->

## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
