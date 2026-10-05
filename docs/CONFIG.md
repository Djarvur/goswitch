# goswitch configuration reference

`goswitchd` reads a single YAML file. Pass it with the `-config` flag:

```sh
goswitchd -config ~/.config/goswitch/goswitch.yaml
```

Without `-config` the daemon runs on built-in defaults (documented below) —
the behavior is identical to a default-valued file.

## Schema

The file has exactly six sections — `hotkeys`, `timeouts`, `correction`,
`macr`, `autocorrect`, `sound`. Every key of every section is listed here;
the document must be complete (missing keys are validation errors, not
silently defaulted) and **strict**: a typo'd key invalidates the whole file
(D-33), so a setting can never appear applied while it actually is not.
Documents written before `correction.flip_after_correction` existed must add
the key: a missing key decodes as `false` (off, not the built-in default)
instead of refusing to start — the completeness discipline is on the
document's author. The `autocorrect` section follows the same completeness
rule with the opposite outcome guarded: absent keys decode into the zero
values, and the zero value is the **off** state (D-54) — the layer is
active exactly when `autocorrect.enabled: true`, and an enabled section
must carry explicit thresholds (enabled with zero thresholds is a loud
refusal, never a degenerate detector), while an empty blocklist forbids
nothing. The `sound` section is the one optional-shape exception: both of
its keys are optional, and an absent section reads as «sounds on» (the
owner's default). Both tray-menu toggles persist into this file — тумблер
автокоррекции («Автокоррекция») writes `autocorrect.enabled`, «Звук»
writes `sound.enabled` — so the choice survives a restart and hot reload
picks it up.

### Migration (D-53 revision, 2026-10-04)

The white-list key `autocorrect.apps` is **removed**. A document carrying
it no longer loads: the strict decoder rejects the whole file with an
unknown-field error — rename the key. The old white list is **not**
transferred to the blocklist mechanically: `apps` used to mean «correct
ONLY in these apps», `apps_blocklist` means «never correct in these apps» —
opposite intents, so carrying the list over by hand would silently invert
your policy.

| Key | Type | Default | Range / vocabulary | Meaning |
|-----|-----|---------|--------------------|---------|
| `hotkeys.tap_key` | string | `shift_r` | closed name table (below) | the key whose tap series (single/double/triple) drives switch/correct-word/correct-phrase |
| `hotkeys.word_layout_combo` | string | `shift+ctrl_r` | closed name table (below) | the combo «correct the last word, then switch the layout» (D-36) |
| `hotkeys.mode_switch_chord` | string | `super+space` | closed name table (below); empty = disabled | flips the script mode directly — goswitch owns Super+Space (install clears the GNOME switch-input-source binding) |
| `timeouts.tap_window_ms` | int | `300` | (0, 2000] | tap disambiguation window; all decisions fire at window expiry (ADR-002) |
| `timeouts.verify_wait_ms` | int | `100` | (0, 2000] | wait for a fresh surrounding-text push when verifying the correction (ADR-004) |
| `correction.backspace_cap` | int | `50` | [1, 500] | cap of the Backspace ladder series (D-27); over-cap corrections are refused silently, not half-deleted |
| `correction.clipboard_rung` | bool | `false` | — | opt-in clipboard replacement rung for selections (D-28); OFF by default |
| `correction.flip_after_correction` | bool | `true` | — | flip the script mode after a successful correction that changed the text (word, phrase, and selection paths); ON by default |
| `macr.enabled` | bool | `false` | — | global switch of the Super→Ctrl remapping layer (ADR-005) |
| `macr.letters` | string | `""` | comma-separated single letters `a`–`z` | the remapped letter set; required (non-empty) when `macr.enabled` is true |
| `macr.apps` | list | `[]` | at most 64 entries | per-app allow list (ADR-005); empty list = the rule set applies everywhere |
| `macr.alt_modifier` | string | `""` | `""` \| `ctrl_l` \| `ctrl_r` | alternative modifier for apps that reject Ctrl+letter; empty = do not introduce one (ADR-005 b.3) |
| `autocorrect.enabled` | bool | `false` | — | global switch of the automatic wrong-layout correction (D-54); OFF by default — the feature never activates on its own; `enabled: true` IS the active state and requires explicit thresholds |
| `autocorrect.apps_blocklist` | list | `[]` | at most 64 regex patterns (RE2) | per-app block list (D-53) — a pattern matching the focused app's identity forbids the correction; matching is by substring (`chrom` matches `org.chromium.Chromium`), `^…$` anchoring is explicit, matching is case-sensitive, order carries no meaning; an empty list means nothing is forbidden |
| `autocorrect.min_word_len` | int | `4` | [2, 16] | minimum word length the detector considers; shorter words are never touched; mandatory while the layer is enabled |
| `autocorrect.trigram_margin` | float | `2.0` | > 0, ≥ `trigram_floor` | required trigram-score margin of the other layout over the current one; checked while the layer is enabled |
| `autocorrect.trigram_floor` | float | `1.0` | > 0 | absolute floor of the other layout's trigram score (the confidence fallback's plausibility demand); mandatory while the layer is enabled |
| `sound.enabled` | bool | `true` (on) | — | the switch sounds: a tone accompanies every layout flip and every autocorrect firing; an absent section or key reads as ON (default on — owner decision); the tray menu's «Звук» toggle persists here |
| `sound.autocorrect_event` | string | `message` | sound-theme event name | the DISTINCT tone for autocorrect firings — the event name played through `canberra-gtk-play` (with `paplay` as the fallback); the flip tone is the fixed `bell` event and is not configurable; an unavailable or failing player is a WARN in the journal and never blocks the switch |

### Binding names

Binding values are name strings from a closed table — a name that is not
in the table rejects the whole config. Combos are `"+"`-joined with the
**key last**, the rest are held modifiers:

- keys: `shift_l`, `shift_r`, `ctrl_l`, `ctrl_r`, `alt_l`, `alt_r`,
  `super_l`, `super_r`, `space` (keyvals from `ibuskeysyms.h`, verified
  verbatim)
- modifiers: `shift`, `ctrl`, `alt`, `super`

Examples: `shift_r` (a tap of the right Shift), `shift+ctrl_r` (hold
Shift, press right Ctrl — the D-36 default combo), `super+space` (the
default mode-switch chord).

Tradeoff to know: documents written before `hotkeys.mode_switch_chord`
existed load with the chord **disabled** (a missing key decodes as empty
= off, not defaulted on) — the completeness discipline is on the
document's author, the same rule as every other key.

## Example

A complete file (the documented defaults):

```yaml
hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
  mode_switch_chord: super+space
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
autocorrect:
  enabled: false
  apps_blocklist: []
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
sound:
  enabled: true
  autocorrect_event: message
```

An activated MACR layer over Chromium only, with the alternative modifier:

```yaml
hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
  mode_switch_chord: super+space
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
macr:
  enabled: true
  letters: "c,v,t"
  apps: ["google-chrome"]
  alt_modifier: "ctrl_r"
```

An enabled autocorrect layer with a blocklist entry (corrections fire
everywhere except the GNOME Terminal windows), with the documented start
thresholds (plan 06-05's detector corpus re-pins the values):

```yaml
hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
  mode_switch_chord: super+space
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
autocorrect:
  enabled: true
  apps_blocklist: ["^org\\.gnome\\.Terminal"]
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
sound:
  enabled: true
  autocorrect_event: message
```

## Hot reload (D-32)

Once started with `-config`, the daemon watches the config file's
**directory** and re-reads the file after every change (debounced). Editors
that save via write-temporary-then-rename (vim, gedit, kwrite) are covered
by design.

- A valid edit takes effect **without a restart**.
- An invalid edit (typo, out-of-range value, unknown key, a broken
  `apps_blocklist` regex pattern) is **rejected as a whole**: the daemon
  keeps running on the last valid configuration (last-good), logs a WARN
  `config reload rejected`, and the rejection is visible later in
  `goswitchctl status` (plan 03-06).
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
| `hotkeys.mode_switch_chord` | their Super+Space-style direct switch binding | ours: flips the script mode immediately, no window; install hands GNOME's `switch-input-source` binding over to goswitch |
| `timeouts.tap_window_ms` | their tap/series timing interval | theirs tunes a first-tap switch delay; ours the multi-tap discrimination window (default 300 ms) |
| `correction.backspace_cap` | (no direct counterpart) | D-27 bound on the destructive Backspace ladder series |
| `correction.clipboard_rung` | (no direct counterpart) | opt-in selection replacement through the clipboard (D-28) |
| `correction.flip_after_correction` | «исправить и переключить»-класс поведения (switch after correction) | owner decision 2026-09-27: a successful correction that changed the text flips the script mode (word and phrase paths); the done-without-change outcome never flips |
| `macr.enabled`, `macr.letters`, `macr.apps`, `macr.alt_modifier` | their macro/app rules | ADR-005 mechanism: Super+letter → Ctrl+letter, per-app lists |

## Privacy

The config's contents never enter the logs. The daemon logs the config
**path**, the applied tap window and validity status — nothing else (the
`macr.apps` names and letters stay out of every log level).

Autocorrect is stricter still: the typed word and the corrected word never
(никогда) enter ANY log level or the `goswitchctl status` — the autocorrect
surface reports counters and reason slugs only (D-54, stricter than
D-20/D-21: an autocorrection fires without a user command, so the word may
be a password).

Where the layer stays silent (the same semantics the README's «Where
autocorrect stays silent» documents, with remedies): an application
without an accessibility tree — Electron/Chromium without the
accessibility mode (e.g. ZCode, the Telegram snap) — gives the daemon no
field role to verify, so nothing is ever corrected there; enable
`org.gnome.desktop.interface toolkit-accessibility` globally (then
restart the application) or launch that one application with
`--force-renderer-accessibility` to fix it. A GTK4 password field
(password-text, role 40) and a widget with no defined role are never
touched. A GTK3 password field reports the ambiguous text-box role (61):
in a known, non-blocklisted application it may still be corrected — an
AT-SPI limitation accepted by the owner (UAT, 2026-10-05); an unknown
application identity combined with role 61 stays silent (fail-closed,
review CR-01).
