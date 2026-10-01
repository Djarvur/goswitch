# goswitch

**goswitch** fixes text typed in the wrong keyboard layout (EN ↔ RU) in any
input field on GNOME Wayland — and switches layouts by hotkey. Type
`ghbdtn`, press the hotkey, get `привет`.

It runs as an **IBus input method engine** in your user session:

- corrects the last word, the whole typed phrase, or the current selection;
- switches the layout on a single hotkey tap; double and triple taps carry
  their own actions, disambiguated on one key;
- commits the corrected text directly — it never grabs the keyboard at the
  evdev layer, so it **coexists with keyd / xremap**;
- **no root anywhere**: a systemd user unit plus a user-level IBus
  component, installed and removed with one command.

## Requirements

- Ubuntu 24.04 (or a similar distro) with GNOME on **Wayland** and IBus
  (the GNOME default input framework);
- a systemd user session (standard on Ubuntu);
- linux/amd64 (the release binaries' architecture).

## Install

Everything — the IBus component, the systemd user unit, the input-source
handover, and engine activation — is one command: `goswitchctl install`.

**Channel A — release archive (recommended):**

1. Take `goswitch_<version>_linux_amd64.tar.gz` from the
   [Releases](https://github.com/Djarvur/goswitch/releases) page — the
   archive carries both binaries and `checksums.txt`.
2. Unpack and install:

   ```sh
   tar -xzf goswitch_*_linux_amd64.tar.gz
   ./goswitchctl install
   ```

**Channel B — go install:**

```sh
go install github.com/Djarvur/goswitch/cmd/goswitchd@v1.0.0
go install github.com/Djarvur/goswitch/cmd/goswitchctl@v1.0.0
goswitchctl install
```

Binaries built through `go install` honestly report `dev` as their
version — only the release archive carries the stamped release version.

`goswitchctl install` is idempotent (safe to re-run), needs no root, and
records your previous input sources — `uninstall` restores them.

### Input-source handover

`goswitchctl install` touches ONLY the input-sources key
(`org.gnome.desktop.input-sources` `sources`): it wraps your source(s)
into goswitch engines and remembers the previous value verbatim.
GNOME's own layout-switch chords (`org.gnome.desktop.wm.keybindings`
`switch-input-source` / `switch-input-source-backward`) are left
untouched — their values are still snapshotted verbatim for the
uninstall restore. Switching is goswitch's own job: a single tap of the
right Shift, or Super + Space (one of goswitch's gestures — see the
table under Usage).

### Input sources and the mode indicator

`goswitchctl install` does not hardcode a layout pair: it reads YOUR
current input sources and wraps at least one `xkb` 'us'/'ru' entry into
a goswitch engine of the matching layout — a single `('xkb', 'us')`
source becomes `('ibus', 'goswitch-en')`, a single `('xkb', 'ru')`
becomes `('ibus', 'goswitch-ru')`, and the classic pair becomes both
engines. You choose the layouts — the installer only wraps your choice.

ONE goswitch source is the recommended shape: GNOME's native panel
indicator appears only with two or more input sources, so with one it is
hidden — by design. The mode indicator is goswitch's own tray icon
(StatusNotifierItem), which follows the daemon's mode immediately. The
daemon's flip changes the typing at once; like GNOME's own layout
switching, a flip resets the correction context — text typed before the
switch is not corrected by a gesture after it.

The icon is interactive: on GNOME it renders as a menu button, and a
click opens a menu — «Переключить раскладку» toggles the layout,
«Статус» shows a notification with the current mode and version,
«Перечитать конфиг» re-reads the YAML (greyed without `-config`). The
icon also re-registers itself when the shell's tray watcher appears late
at boot or silently drops the item — no daemon restart needed.

A foreign source beside goswitch engines cannot come from the installer
(it is refused) and is flagged red by `goswitchctl selfcheck` — the
daemon sees no keys through a plain `xkb` source. You stay in charge of
the source list: if you switch to a foreign source yourself, the daemon
follows the factual engine and never fights the desktop (no correction,
no hijack). Switch back to a goswitch source to resume correction and
hotkey switching.

Restart behavior: a daemon restart keeps everything (it re-activates the
engine that owns the current source — the `engine reactivated` journal
line). An ibus-daemon restart returns the session to the shell's own
current source; a layout reached through a shell gesture survives it,
and the daemon follows whichever engine the shell brings back.

## Verify

Run the built-in audit:

```sh
goswitchctl selfcheck
```

Six live checks print one verdict line each — daemon version and build,
IBus component visibility, the user unit's state, live engine
registration, config validity, input-source ownership. Every red verdict
names its fix, and if the IBus registry cache has lost the goswitch entry
(another IBus tool can evict it), selfcheck repairs the cache itself and
re-checks — no manual steps.

## Usage

Default gestures (reconfigurable — see
[docs/CONFIG.md](docs/CONFIG.md)):

| Gesture | Action |
|---------|--------|
| single tap of the right Shift | switch layout EN ↔ RU |
| Super + Space | switch layout EN ↔ RU |
| double tap | correct the last word (`ghbdtn` → `привет`, letter case preserved) |
| triple tap | correct the whole typed phrase |
| double tap with text selected | correct the selected range only |
| Shift + right Ctrl | correct the last word **and** switch the layout |

Words already in the target layout and mixed-script words are converted
point-wise — only the foreign letters change.

CLI:

```sh
goswitchctl status    # daemon state and counters (--json for scripts)
goswitchctl correct   # force a word correction right now
goswitchctl reload    # re-read the config file immediately
```

## Configuration

Out of the box the daemon runs on **built-in defaults** — the installer
does not create a config file. To customize, write a YAML document (every
key is documented in [docs/CONFIG.md](docs/CONFIG.md); the conventional
location `~/.config/goswitch/config.yaml` is the file
`goswitchctl selfcheck` validates when it exists) and attach it to the
service:

```sh
systemctl --user edit goswitchd
# in the editor, add:
#   [Service]
#   ExecStart=
#   ExecStart=/path/to/goswitchd -config /home/you/.config/goswitch/config.yaml
systemctl --user restart goswitchd
```

With a config attached, valid edits apply **without a restart** (hot
reload); an invalid edit is rejected as a whole and the daemon keeps the
last good configuration.

## Performance

Measured on the reference desktop (Ubuntu 24.04, GNOME 46, Wayland;
2026-09-16), 40 repeats of one homogeneous hot case — correct the word
and switch the layout in a single chord:

| Metric | Value |
|--------|-------|
| p50 latency | 166.9 ms |
| p95 latency | 184.4 ms |
| p99 latency | 186.5 ms |
| peak RSS | 12.4 MB (budget: < 50 MB) |

Methodology: the window runs from the test rig's key injection (ydotool →
uinput) to the corrected text observed over AT-SPI — an honest upper
bound that includes the rig's own per-chord injection overhead
(~170 ms with ydotool 0.1.8). The daemon's internal reaction to the chord
is below 2 ms; the memory budget holds with a wide margin. The table is
re-measured and refreshed with every release.

Budget status: honestly stated, this end-to-end window **exceeds** the
50 ms hotkey-reaction budget — the excess comes from the measurement
rig's own injection overhead, not from the daemon (reaction < 2 ms;
the memory budget is met). v1.0.0 ships with these as-measured numbers.

## Privacy

The daemon logs counters and decisions only — never the text you type,
never your config contents.

The `-debug` flag enables key tracing: **every** keystroke (keyval,
keycode, modifiers, press/release) is written to the log at DEBUG level.

> **Warning:** debug logs necessarily contain everything typed, including
> passwords and other secrets — input into password fields is traced too.
> In the normal mode the daemon's stderr goes to the systemd journal
> without any key tracing. Do not leave `-debug` enabled permanently;
> enable it only while debugging, and delete the logs afterwards.

## Troubleshooting

- **goswitch vanished from the input-source list (the engine is gone).**
  Run `goswitchctl selfcheck` — it detects a lost IBus registry entry,
  repairs it, and re-checks by itself.
- **Gestures do nothing in a GTK4 application** (for example,
  gnome-text-editor ignores the hotkeys). Known GNOME 46 GTK4-Wayland
  platform limitation: forwarded key events reach the application, but
  GTK4 widgets do not act on them. Chromium-family applications apply
  them normally.
- **A config edit "does not apply".** An invalid document is rejected as
  a whole — the daemon keeps running on the last good configuration, and
  `goswitchctl status` shows the parse reason in `config_error=`.
- **Correction now survives daemon restarts by itself.** After every
  (re)registration goswitchd re-activates its engine — as long as goswitch
  owns the current input source — so a daemon restart
  (`systemctl --user restart goswitchd`) or an ibus-daemon restart no
  longer silently disables correction. The journal line
  `engine reactivated` confirms it.
- **A non-goswitch source took over and correction stopped, yet the daemon
  is still alive.** If you switch to a foreign source (say, a plain `xkb`
  layout), the daemon follows the factual engine and stays out of the way —
  it never fights the desktop. `goswitchctl selfcheck` is the audit's
  answer: it flags a goswitch engine sitting beside a foreign source red
  (the daemon sees no keys through an xkb source). Remedy: switch back to
  a goswitch source (`ibus engine goswitch-en`), or restart the unit —
  self-reactivation then re-applies the owned source. Check
  `journalctl --user -u goswitchd` for the WARN/INFO lines.

## Uninstall

```sh
goswitchctl uninstall
```

stops and removes the systemd unit and the IBus component, refreshes the
IBus registry, and **restores your previous input sources**. Your
`~/.config/goswitch/` directory is kept; remove it too with
`goswitchctl uninstall --purge`.

## Documentation

- [docs/SPEC.md](docs/SPEC.md) — the full technical specification;
- [docs/CONFIG.md](docs/CONFIG.md) — the complete configuration reference;
- [docs/ACCEPTANCE.md](docs/ACCEPTANCE.md) — the release acceptance checklist.

## License

[MIT](LICENSE)
