# Stack Research — goswitch v1.2 «Терминалы и видимость приложений»

**Domain:** GNOME Wayland IBus input-method daemon (Go) — terminal correction, dynamic tray-menu item, per-app accessibility enablement
**Researched:** 2026-10-08 (live probes on the target: Ubuntu 24.04, GNOME 46, wezterm 20240203, Telegram snap 7.2.9 rev 7150, ZCode /opt/ZCode Electron)
**Confidence:** MEDIUM–HIGH (local target-runtime probes = HIGH; cross-checked web = MEDIUM; single-source web = LOW)

## Headline Finding

**Zero new Go dependencies.** All three v1.2 features ride the existing stack: godbus v5.2.2 (unchanged, still latest on the proxy), the in-repo `internal/engine` ForwardKeyEvent emitter (wire-tested AND already live-driven by the clipboard rung's Ctrl+V forward burst, D-28), the in-repo `internal/clipboard` wl-clipboard wrapper (needs only a `--primary` flag), the in-repo `internal/indicator` DBusMenu label-delta machinery, and the in-repo `internal/a11y` + `internal/appid` packages. The only "new stack" is: a few in-repo keyval constants, external-tool recipes (wezterm config, .desktop override, docs), and one structural discovery that reshapes the Telegram plan (AppArmor wall, below).

## Recommended Stack

### Core Technologies (delta from v1.1 — everything else unchanged)

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| In-repo `internal/engine` — ForwardKeyEvent bursts | existing (go.mod `go 1.23`) | Deliver injected keys (BackSpace, Home, modifiers) into the focused terminal through the daemon's OWN IBus input context — the channel the 2026-10-08 spike validated as race-free | The emitter exists and its wire contract is pinned by `emitter_wire_test.go`; the Ctrl+V forward burst of the clipboard rung (D-28) already drives this exact path in production. v1.2 adds keyval constants to `internal/engine/keys.go` from `/usr/include/ibus-1.0/ibuskeysyms.h` in the file's established verified-verbatim pattern: `KeyHome = 0xff50`, `KeyEnd = 0xff57`, `KeyLeft/Up/Right/Down = 0xff51–0xff54`. `MaskShift = 1<<0` already exists; `CapSurroundingText = 1<<5` (0x20) is the bit terminals lack (wezterm caps 0x9 = Preedit|Focus). |
| `wl-clipboard` (`wl-paste`/`wl-copy`) via in-repo `internal/clipboard` | 2.2.1 (dpkg-verified on target) | Read the selection produced by injected keys: `wl-paste --primary --no-newline` | GNOME/mutter implements the primary-selection protocol; `--primary` exists in wl-clipboard 2.x. The existing wrapper already has the security discipline (content only via stdin pipe, flag-only argv, 1.5 s timeout) — adding `--primary` is a flag-level extension. Write-back to the clipboard is NOT needed: correction text enters via `commit_text`, and restoring the user's clipboard is the existing save/restore rung (D-28/D-29). |
| `gsettings org.gnome.desktop.interface toolkit-accessibility` via in-repo `internal/a11y` | GNOME 46 (existing Phase 8 code) | Enable a11y in ZCode (Electron/Chromium reads this key **at startup**) and trigger the Telegram snap's own `QT_ACCESSIBILITY=1` export | The daemon already sets it (default ON) plus the `org.a11y.Status.IsEnabled` belt. Live probe: ZCode is visible and nameable in the a11y tree today ('zcode' in the desktop children list). Chromium honors the key only for apps launched after it is set — the milestone work is relaunch guidance/docs, not new daemon code. |
| In-repo `internal/indicator` (SNI + DBusMenu v3) + `internal/appid` | existing | Dynamic «Добавить {ИмяПриложения} в чёрный список» menu item | `menu.go` already pushes dynamic labels through `ItemsPropertiesUpdated` (`SetMode`/`SetKeys`/`SetCorrections`/`SetVersion` precedents) and routes `Event` clicks; a new item id + interpolated label builder + callback rides the existing machinery. `appid` supplies the focused app's bridge-namespace identity over the a11y bus. No Go library would do this better — the DBusMenu v3 subset is already implemented and live-verified. |
| wezterm config recipe (owner's `~/.config/wezterm/wezterm.lua`) | wezterm 20240203-110809-5046fc22 | Enable keyboard select-to-line-start in wezterm | wezterm 20240203's default key tables have **no Shift+Home/End bindings at all** (only Shift+PageUp/PageDown scroll; Ctrl+Shift+X ActivateCopyMode). Keyboard selection exists only via CopyMode sequences — and the owner's config is already built on `wezterm.gui.default_keys()` + CopyMode `Multiple` sequences (the Cmd+A recipe). The enabling entry: `SHIFT+Home` → `Multiple{ ActivateCopyMode, CopyMode{SetSelectionMode='Cell'}, CopyMode'MoveToStartOfLineContent', CopyMode'Close' }` (Close leaves the selection active). Note: 20240203 is the **last tagged release ever** (upstream ships nightlies only since Feb 2024; issue #6341 tracks the next release) — the recipe targets the installed version. |
| User-local `.desktop` override for ZCode | `~/.local/share/applications/zcode.desktop` (new file; `/usr/share/applications/zcode.desktop` is root/dpkg-owned, `Exec=/opt/ZCode/zcode %U`) | Deterministic Electron a11y belt: `--force-renderer-accessibility` | Removes the launch-order dependency on the gsettings key. User-local entries shadow system ones — the same mechanism the Telegram snap entry already uses on this machine. Electron/Chromium documents automatic a11y on AT detection; the flag forces the renderer tree regardless of startup timing. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/godbus/dbus/v5` | v5.2.2 (still latest — proxy verified 2025-12-29) | Engine service side, a11y belt, DBusMenu, appid | Unchanged; no bump. |
| `gopkg.in/yaml.v3` | v3.0.1 | Config (blocklist regex list already there) | Unchanged. |
| `github.com/fsnotify/fsnotify` | v1.10.1 (latest 2026-05-04) | Config hot reload | Unchanged. |
| `github.com/jfreymuth/pulse` (+ oggvorbis) | v0.1.3 / v1.0.5 (both latest) | Sounds | Unchanged. |
| VTE 0.76.0 / gnome-terminal 3.52.0 | dpkg-verified on target | Target terminal #2 — read-only integration facts below | No package to add; e2e stand covers it. |
| telegram-desktop snap | 7.2.9 rev 7150 | Target app whose a11y is **structurally blocked** (see Telegram finding) | No package change inside the milestone; decision needed (Alternatives). |
| `ibus` CLI (`ibus engine`, `ibus address`, `ibus restart`) | 1.5.29 | Debug/verification of engine state, wezterm-defect diagnosis | Already used. |

### Development Tools (e2e stand — additions only)

| Tool | Purpose | Notes |
|------|---------|-------|
| a11y-tree probe (existing `python3-gi` 3.48.2 + `gir1.2-atspi-2.0`, or raw `busctl --address=$(org.a11y.Bus GetAddress)`) | Assert app visibility in e2e: list desktop children, `Accessible.GetRole` per app; also the tool that produced the AppArmor finding | Already installed on target; no new packages. |
| `gdbus`/`busctl --user` introspection of `com.canonical.dbusmenu` | Verify dynamic-label deltas land (GetLayout/GetGroupProperties before/after) | Ships with OS. |
| `wezterm show-keys` (20240203 has it) | Assert the Shift+Home recipe binds, in CI | No install; binary present. |

## Installation

```bash
# go.mod: NO CHANGES — zero new modules for v1.2.
# (existing: godbus/dbus v5.2.2, yaml.v3 v3.0.1, fsnotify v1.10.1, jfreymuth/pulse v0.1.3)

# System packages (already installed on target; formalize as install-deps):
sudo apt install wl-clipboard   # 2.2.1 — already required since the v1.1 clipboard rung

# No new e2e stand packages: python3-gi + gir1.2-atspi-2.0 + ydotool already present.
```

```lua
-- wezterm recipe (owner's ~/.config/wezterm/wezterm.lua, appended to the keys table):
table.insert(keys, {
  key = 'Home', mods = 'SHIFT',
  action = wezterm.action.Multiple{
    wezterm.action.ActivateCopyMode,
    wezterm.action.CopyMode { SetSelectionMode = 'Cell' },
    wezterm.action.CopyMode 'MoveToStartOfLineContent',
    wezterm.action.CopyMode 'Close',   -- Close LEAVES the selection active
  },
})
-- wezterm copy_on_select defaults to ClipboardAndPrimarySelection: the
-- selection lands in BOTH buffers -> wl-paste --primary can read it.
```

```desktop
# ~/.local/share/applications/zcode.desktop (user-local override; installer drop)
[Desktop Entry]
Name=ZCode
Exec=/opt/ZCode/zcode --force-renderer-accessibility %U
Terminal=false
Type=Application
Icon=zcode
StartupWMClass=ZCode
MimeType=x-scheme-handler/zcode;
Categories=Development;
```

## Load-Bearing Findings (what the stack must account for)

### (a) Terminal correction

1. **Terminal selections are visual only — `commit_text` does NOT erase the selected pty text.** VTE 0.76's incoming-contents handler only *deselects* when contents change (source-verified, `src/vte.cc` ~line 3937); the `widget_key_press` path (src/vte.cc:4972) has zero selection logic. Consequence: the write-back in terminals is exactly the built-but-never-live-driven **backspace-replay** (ADR-003 level 2): `N × ForwardKeyEvent(KeyBackSpace)` then `commit_text(corrected)`; N comes from the daemon keystroke buffer or the read-back length. v1.2 takes level-2 live — its wire behavior is already pinned by tests.
2. **Select-to-line-start via injected Shift+Home does NOT work out of the box in either target terminal.**
   - **VTE 0.76 / gnome-terminal 3.52: keyboard selection does not exist.** `Terminal::widget_key_press` goes IM-filter → context menu → keymap; Shift+Home falls to `_vte_keymap` and is sent to the child app as `CSI 1;2H` (xterm modifier-parameter encoding; keymap.cc source-verified). No user-configurable binding can change this — gnome-terminal has no "select" action to bind.
   - **wezterm 20240203: no default Shift+Home binding** (default-keys docs); keyboard selection only via CopyMode, enabled by the config recipe above.
   - Consequence for scope: the injected-keys + clipboard read channel is **wezterm-only, and conditional on a user config recipe**. For VTE, the whole-line read must come from the daemon's own keystroke buffer (already exists for phrase correction; resets on Enter/Tab/focus/Escape); word correction in VTE needs no read channel at all. Do not gate VTE acceptance on a clipboard read-back.
3. **The ForwardKeyEvent channel itself is production-proven** — the clipboard rung already forwards a Ctrl+V burst through it (D-28); the 2026-10-08 spike additionally confirmed the daemon's IBus channel has no focus race (external injectors could not manage window focus on the live desktop).
4. **wezterm layout-switch defect (flip via SetGlobalEngine does not take effect):** no stack addition fixes it — it is in-repo engine work plus a diagnosis spike. Verified context: wezterm on GNOME Wayland is a text-input-v3 client bridged to IBus by mutter (no direct IBus client on Wayland); the fragility class is documented (GNOME: "IME does not work in wezterm after unlock screen"; wezterm #4358 "keeps switching keyboard layout" on focus; nightlies reworked winit keyboard handling while 20240203 is pinned). Candidate mechanisms for the spike: stale client xkb keymap after `wl_keyboard.keymap` change vs input context not re-armed after `SetGlobalEngine`. LOW→MEDIUM confidence until the spike runs.
5. **Clipboard side effect worth a pitfalls entry:** wezterm's `copy_on_select` default (`ClipboardAndPrimarySelection`) clobbers the user's CLIPBOARD on every selection — interacts with the existing clipboard save/restore rung in the terminal path.

### (b) Dynamic blocklist menu item

- Zero new dependencies: `menu.go` label deltas + `appid` observer cover it. The item is one more `menuID` with a label built from the focused app's bridge-namespace name; `ItemsPropertiesUpdated` pushes the change (the renderer refreshes an open menu).
- **Identity reach = a11y visibility.** `appid` sees only apps on the a11y bus: GTK, unconfined Electron/Qt. It can never see: wezterm (no a11y at all — irrelevant: terminals are not autocorrected), and **any strictly confined snap** (Telegram, Chromium — see below). The menu item must degrade gracefully (hidden, or a generic «Неизвестное приложение») when the name is unavailable; the D-53 fail-closed precedent applies.

### (c) Accessibility in Telegram (Qt snap) and ZCode (Electron)

- **ZCode — works today, add a belt.** Live probe 2026-10-08: 'zcode' is in the a11y desktop children and nameable. Mechanism: Chromium/Electron reads `toolkit-accessibility` **at startup** (chromium.org Linux Accessibility docs) — the daemon already sets it with default ON (Phase 8), so every fresh ZCode launch is covered; the owner's D-8-6 manual precedent validated the lever. The `.desktop` override with `--force-renderer-accessibility` removes the only failure mode (app launched before the key was set). Milestone work = recipe/installer drop + relaunch docs.
- **Telegram — structurally blocked by snap confinement; NO environment recipe can fix it.** Live-probed 2026-10-08: the snap wrapper already exports `QT_ACCESSIBILITY=1` (sourced from `toolkit-accessibility` by the desktop-launch helper — verified in the running process's environ); the process **is registered on the a11y bus** (it appears as a live bus destination), but the snap's AppArmor profile denies **all inbound AT-SPI method calls** from unconfined clients: `org.a11y.atspi.Cache.GetItems` → policy-denied; raw `org.a11y.atspi.Accessible.GetRole` on `/org/a11y/atspi/accessible/root` → **"Access denied"**. The Chromium snap shows the same denial class. Consequence: the daemon's role-gate can never see Telegram-snap, by confinement — this is the "если Telegram не поддастся" branch materializing at the stack level. Options, in order: (1) an unconfined Telegram build — the official tarball (telegram.org) or flatpak `org.telegram.desktop` — where Qt reads the same gsettings key and the a11y tree becomes readable (flatpak's a11y reachability needs a one-off live check before committing); (2) the owner's stated daemon-side fallback. Telegram is Qt6: auto-activation via `org.a11y.Status.IsEnabled` (which the existing belt sets) plus `QT_ACCESSIBILITY=1` covers the toolkit side once confinement is out of the way.
- **`QT_LINUX_ACCESSIBILITY_ALWAYS_ON=1`** forces the Qt at-spi bridge regardless of detection — works (qaccessible.cpp), but carries a measurable Qt performance cost (community-reported) and is redundant while the belt + wrapper path functions. Docs-level last resort only.
- **`GTK_MODULES=gail:atk-bridge`** is GTK2-era; GTK3+ auto-loads the bridge from `toolkit-accessibility`. Not needed for any v1.2 target.

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| wl-clipboard subprocess (+ `--primary`) | Pure-Go Wayland client implementing `wp_primary_selection_unstable_v1` (e.g. `rajveermalviya/go-wayland`) | Never for v1.2 — a heavyweight new dep against the audit constraint; the subprocess path is validated in production since v1.1. |
| Daemon ForwardKeyEvent bursts | External injectors driven by the daemon (ydotool/wtype) or uinput in-process | Never: the 2026-10-08 spike showed the external path loses the focus race on a live desktop, and uinput in the daemon violates the architecture axiom (keyd/xremap coexistence). |
| Daemon keystroke buffer as the VTE line-read channel | OSC 52 emission from a shell hook into the clipboard | If a future milestone needs whole-line in VTE and the buffer proves insufficient — requires shell cooperation (precmd hook) and VTE does not even support OSC 52 clipboard writes; heavy user setup. |
| `appid` bridge-namespace identity | /proc/wm-based window identity | Never on Wayland — no root-free stable identity source outside the a11y bus; silent degradation is the designed behavior. |
| User-local `.desktop` override for ZCode | Editing `/usr/share/applications/zcode.desktop` | Never — dpkg-owned, clobbered on app updates. |
| Unconfined Telegram (tarball/flatpak) | `snap connect`/interface tweaks for the Telegram snap | Flatpak/tarball only; there is **no** user-fixable interface for the AppArmor AT-SPI denial (the rule lives in snapd's generated profile). |
| gsettings + relaunch for Electron a11y | `--force-renderer-accessibility` only, or in-app `app.setAccessibilitySupportEnabled(true)` | In-app API requires patching ZCode; flag-only works but keeps the gsettings dependency implicit — belt-and-suspenders (both) is the cheap deterministic choice. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Any new Go module for v1.2 (clipboard libs, Wayland client libs, menu/DBusMenu libs, AT-SPI bindings) | Every v1.2 capability is already covered by stdlib + godbus + subprocess + in-repo packages; additions dilute the "auditable in an afternoon" property | Existing `internal/clipboard`, `internal/engine`, `internal/indicator`, `internal/appid`, `internal/a11y` |
| `m31labs.dev/fluffyui` atspi package | Part of a heavy UI-automation module, not a maintained standalone AT-SPI client (carried from v1.0 research) | godbus calls per `internal/appid` |
| `QT_LINUX_ACCESSIBILITY_ALWAYS_ON=1` as the primary recipe | Measurable Qt perf cost; redundant with toolkit-accessibility + IsEnabled belt | gsettings key (existing Phase 8 series) + relaunch guidance; ALWAYS_ON documented as last resort |
| `GTK_MODULES=gail:atk-bridge` in recipes | GTK2-era; GTK3+ auto-loads; meaningless for Qt/Electron | Nothing — omit from docs |
| `xclip`/`xsel` | X11-only, dead on Wayland | wl-clipboard |
| Expecting `commit_text` over a terminal selection to replace it | Terminal selections are visual-only (VTE source-verified); the shell never learns about the erase | Backspace-replay (level 2) then commit |
| Building the VTE whole-line read on injected selection keys | VTE has no keyboard selection at all (source-verified) | Daemon keystroke buffer; wezterm gets the config-recipe channel |
| Snap-confinement workarounds for Telegram a11y | Denial lives in snapd's generated AppArmor profile; no interface/slot exists | Unconfined Telegram build or the daemon-side fallback decision |

## Stack Patterns by Variant

**If target = wezterm (owner's daily):**
- Read: config-recipe CopyMode selection → `wl-paste --primary --no-newline` (in-repo clipboard pkg + `--primary` flag).
- Write-back: N×BackSpace replay + `commit_text` (selections are visual; never rely on commit-over-selection).
- Layout-switch defect: in-repo engine work + diagnosis spike; ForwardKeyEvent hotkeys are the unaffected path.

**If target = VTE / gnome-terminal:**
- Read: daemon keystroke buffer only. Word correction needs no read channel.
- No clipboard read-back; no keyboard-selection recipes possible. Acceptance criteria must not reference clipboard reads for VTE.

**If app = unconfined Electron (ZCode):**
- gsettings toolkit-accessibility (existing) covers fresh launches; `.desktop` override with `--force-renderer-accessibility` as the deterministic belt; docs say "restart the app after enabling".

**If app = Qt snap (Telegram):**
- Stack-blocked for a11y visibility regardless of env. Decide: unconfined build (tarball/flatpak — flatpak needs a one-off a11y reachability check) or daemon-side fallback (owner's stated contingency).

**If app is a11y-invisible (any reason):**
- Role-gate fail-closed stays silent (validated design); manual correction chords still work wherever the input context delivers keys.

## Version Compatibility

| Component | Version | Compatible With | Notes |
|-----------|---------|-----------------|-------|
| `github.com/godbus/dbus/v5` | v5.2.2 (latest, 2025-12-29) | Go 1.17+; IBus 1.5.x | Unchanged; ForwardKeyEvent wire contract pinned by in-repo tests. |
| `gopkg.in/yaml.v3` / `github.com/fsnotify/fsnotify` | v3.0.1 / v1.10.1 (both latest) | — | Unchanged. |
| `github.com/jfreymuth/pulse` | v0.1.3 (latest, 2026-08-17) | — | Unchanged. |
| wl-clipboard | 2.2.1 (target dpkg) | Wayland primary-selection protocol via mutter | `--primary` supported; flags pinned in `internal/clipboard`. |
| wezterm | 20240203-110809-5046fc22 (installed = latest tagged release; nightlies-only upstream) | CopyMode `SetSelectionMode`/`MoveToStartOfLineContent`/`Close` exist in this version (owner config already uses them) | Re-verify key tables if the owner ever moves to a nightly. |
| VTE | 0.76.0 (gnome-terminal 3.52.0, target dpkg) | Keyboard-path facts source-verified at tag 0.76.0 | Re-verify if Ubuntu backports keyboard selection (none announced). |
| Electron/Chromium (ZCode) | `/opt/ZCode` current | Reads `toolkit-accessibility` at startup; `--force-renderer-accessibility` honored | Flag is stable Chromium surface. |
| Telegram desktop snap | 7.2.9 rev 7150 | AppArmor denies inbound AT-SPI (live-probed) | Re-test confinement if the snap ships changed profiles. |
| IBus | 1.5.29 | engine registration, SetGlobalEngine flip | Unchanged from v1.0 research. |

## Direct Answers Feeding the Roadmap

- **(a) Terminal correction stack:** no new deps. Read = injected keys (in-repo ForwardKeyEvent; add Home/End/arrows keyvals to keys.go) + `wl-paste --primary` (in-repo clipboard pkg + `--primary`); write-back = N×BackSpace replay + commit (level-2 goes live — the built replay emitter is the mechanism; terminal selections are visual-only so commit-over-selection is a dead end, source-verified). wezterm needs a user config recipe for Shift+Home (version pinned at 20240203); VTE cannot do keyboard selection at all — whole-line there rides the keystroke buffer. The wezterm layout-switch defect is in-repo engine work, not a stack addition; plan a diagnosis spike (two named candidate mechanisms, LOW confidence until probed).
- **(b) Dynamic blocklist menu item:** zero deps — one menuID + interpolated label through the existing `ItemsPropertiesUpdated` delta machinery; appid supplies the name. Design for invisible apps (snaps, wezterm): degrade, never guess.
- **(c) Telegram/ZCode a11y:** ZCode — already working via toolkit-accessibility (live-probed); add the `--force-renderer-accessibility` desktop override as belt + relaunch docs. Telegram snap — **structurally blocked by AppArmor** (env already correct; GetRole denied live): the milestone decision is unconfined Telegram build vs the owner's daemon-side fallback; no dependency or env var changes this.
- **What NOT to add:** any new Go module; `QT_LINUX_ACCESSIBILITY_ALWAYS_ON` as primary; `GTK_MODULES`; xclip/xsel; snap-interface workarounds; OSC-52 tricks.

## Sources

- Live target-system probes (2026-10-08, HIGH — the target runtime itself): dpkg versions (wezterm 20240203, wl-clipboard 2.2.1, VTE 0.76.0, gnome-terminal 3.52.0, python3-gi 3.48.2); snap list + connections (telegram-desktop 7.2.9 rev 7150: desktop/desktop-legacy/gsettings plugged); Telegram process environ (`QT_ACCESSIBILITY=1` present); a11y-bus probes (`org.a11y.Bus.GetAddress`; AppArmor denial of `Cache.GetItems` and `Accessible.GetRole` into telegram-desktop/chromium snap; 'zcode' visible+nameable in the desktop children); owner configs (`~/.config/wezterm/wezterm.lua` CopyMode patterns, `/usr/share/applications/zcode.desktop`, user-local Telegram desktop entry).
- VTE 0.76.0 source (GitLab tag 0.76.0, HIGH — primary source read): `src/vte.cc` `Terminal::widget_key_press` (no keyboard selection), incoming-contents deselect-only behavior, `src/keymap.cc` modifier-parameter encoding; `src/widget.cc` thin GTK dispatch.
- [WezTerm default key assignments](https://wezterm.org/config/default-keys.html) — no Shift+Home default, Ctrl+Shift+X CopyMode — **MEDIUM** (docs, live-config cross-check)
- [WezTerm CopyTo destinations](https://wezterm.org) + copy_on_select default `ClipboardAndPrimarySelection` — **MEDIUM** (docs via search; mechanism cross-checked against owner config)
- [WezTerm releases](https://github.com/wezterm/wezterm/releases), [#6341 next-release checklist](https://github.com/wezterm/wezterm/issues/6341), [#4358 layout switching on focus](https://github.com/wezterm/wezterm/issues/4358) — release cadence, layout-bug class — **MEDIUM**
- [Chromium Linux Accessibility](https://www.chromium.org) — startup toolkit-accessibility detection — **MEDIUM** (official docs, owner-precedent cross-check)
- [Electron accessibility tutorial](https://www.electronjs.org/docs/latest/tutorial/accessibility) — AT auto-detection, `app.setAccessibilitySupportEnabled` — **MEDIUM**
- Qt accessibility: `QT_ACCESSIBILITY` / `QT_LINUX_ACCESSIBILITY_ALWAYS_ON` (qtatspi, qaccessible.cpp via codebrowser; KDE UserBase; Manjaro forum perf report) — **MEDIUM**
- [wl-clipboard manual](https://man.freebsd.org/cgi/man.cgi?query=wl-clipboard) (`--primary`), wp_primary_selection_unstable_v1 context — **MEDIUM**, dpkg 2.2.1 local cross-check
- Snap a11y confinement: [Snapcraft forum — AT-SPI in snaps](https://forum.snapcraft.io), AskUbuntu socket-denial reports — **LOW/MEDIUM**, superseded for the verdict by the live AppArmor probe (HIGH)

---
*Stack research for: goswitch v1.2 — terminals, tray-menu dynamism, per-app accessibility*
*Researched: 2026-10-08*
