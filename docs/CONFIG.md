# goswitch configuration reference

`goswitchd` reads a single YAML file. Pass it with the `-config` flag:

```sh
goswitchd -config ~/.config/goswitch/goswitch.yaml
```

Without `-config` the daemon runs on built-in defaults (documented below) —
the behavior is identical to a default-valued file.

## Schema

The file has exactly four sections — `hotkeys`, `timeouts`, `correction`,
`macr`. Every key of every section is listed here; the document must be
complete (missing keys are validation errors, not silently defaulted) and
**strict**: a typo'd key invalidates the whole file (D-33), so a setting can
never appear applied while it actually is not.

| Key | Type | Default | Range / vocabulary | Meaning |
|-----|------|---------|--------------------|---------|
| `hotkeys.tap_key` | string | `shift_r` | closed name table (below) | the key whose tap series (single/double/triple) drives switch/correct-word/correct-phrase |
| `hotkeys.word_layout_combo` | string | `shift+ctrl_r` | closed name table (below) | the combo «correct the last word, then switch the layout» (D-36) |
| `timeouts.tap_window_ms` | int | `300` | (0, 2000] | tap disambiguation window; all decisions fire at window expiry (ADR-002) |
| `timeouts.verify_wait_ms` | int | `100` | (0, 2000] | wait for a fresh surrounding-text push when verifying the correction (ADR-004) |
| `correction.backspace_cap` | int | `50` | [1, 500] | cap of the Backspace ladder series (D-27); over-cap corrections are refused silently, not half-deleted |
| `correction.clipboard_rung` | bool | `false` | — | opt-in clipboard replacement rung for selections (D-28); OFF by default |
| `macr.enabled` | bool | `false` | — | global switch of the Super→Ctrl remapping layer (ADR-005) |
| `macr.letters` | string | `""` | comma-separated single letters `a`–`z` | the remapped letter set; required (non-empty) when `macr.enabled` is true |
| `macr.apps` | list | `[]` | at most 64 entries | per-app allow list (ADR-005); empty list = the rule set applies everywhere |
| `macr.alt_modifier` | string | `""` | `""` \| `ctrl_l` \| `ctrl_r` | alternative modifier for apps that reject Ctrl+letter; empty = do not introduce one (ADR-005 b.3) |

### Binding names

Binding values are name strings from a closed table — a name that is not
in the table rejects the whole config. Combos are `"+"`-joined with the
**key last**, the rest are held modifiers:

- keys: `shift_l`, `shift_r`, `ctrl_l`, `ctrl_r`, `alt_l`, `alt_r`,
  `super_l`, `super_r` (keyvals from `ibuskeysyms.h`, verified verbatim)
- modifiers: `shift`, `ctrl`, `alt`, `super`

Examples: `shift_r` (a tap of the right Shift), `shift+ctrl_r` (hold
Shift, press right Ctrl — the D-36 default combo).

## Example

A complete file (the documented defaults):

```yaml
hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
```

An activated MACR layer over Chromium only, with the alternative modifier:

```yaml
hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
macr:
  enabled: true
  letters: "c,v,t"
  apps: ["google-chrome"]
  alt_modifier: "ctrl_r"
```

## Hot reload (D-32)

Once started with `-config`, the daemon watches the config file's
**directory** and re-reads the file after every change (debounced). Editors
that save via write-temporary-then-rename (vim, gedit, kwrite) are covered
by design.

- A valid edit takes effect **without a restart**.
- An invalid edit (typo, out-of-range value, unknown key) is **rejected as
  a whole**: the daemon keeps running on the last valid configuration
  (last-good), logs a WARN `config reload rejected`, and the rejection is
  visible later in `goswitchctl status` (plan 03-06).
- A missing or empty file at startup is a **start error** — an explicit
  `-config` must yield a working configuration, never silent defaults.

Note on the tap window: a reloaded `timeouts.tap_window_ms` applies to tap
series started after the reload; an already-armed window timer lives out
its old value (the FSM's stale-timer invariant).

## Caramba correspondence table (CONF-03, wish-level)

goswitch deliberately does **not** clone Caramba's key names or its
behavioral model. Caramba's schema is bound to its «switch on the first
tap» model; goswitch follows ADR-002 — the classic-with-waiting scheme
where every decision fires at the expiry of the disambiguation window, so
single/double/triple taps can coexist on one key. This table maps our keys
to the Caramba settings that serve the same purpose:

| goswitch key | Caramba counterpart (purpose) | Notes |
|--------------|-------------------------------|-------|
| `hotkeys.tap_key` | переключение раскладки / горячая клавиша (their one-tap switch) | ours: one key carries 1/2/3-tap actions at window expiry (ADR-002); theirs: the first tap switches immediately |
| `hotkeys.word_layout_combo` | «исправить и переключить»-класс комбинаций | ours: correct the word FIRST, then switch (D-36 order, fixed by the SPEC action name) |
| `timeouts.tap_window_ms` | their tap/series timing interval | theirs tunes a first-tap switch delay; ours the multi-tap discrimination window (default 300 ms) |
| `correction.backspace_cap` | (no direct counterpart) | D-27 bound on the destructive Backspace ladder series |
| `correction.clipboard_rung` | (no direct counterpart) | opt-in selection replacement through the clipboard (D-28) |
| `macr.enabled`, `macr.letters`, `macr.apps`, `macr.alt_modifier` | their macro/app rules | ADR-005 mechanism: Super+letter → Ctrl+letter, per-app lists |

## Privacy

The config's contents never enter the logs. The daemon logs the config
**path**, the applied tap window and validity status — nothing else (the
`macr.apps` names and letters stay out of every log level).
