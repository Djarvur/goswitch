# Feature Research

**Domain:** Keyboard layout switcher / mistyped-text corrector (Punto Switcher analog) for GNOME Wayland, implemented as an IBus input method engine daemon in Go
**Researched:** 2026-09-10
**Confidence:** MEDIUM (cross-checked across 10+ competitor products and primary READMEs; single-source claims marked inline)

## Feature Landscape

Scope anchor: goswitch's owner decisions constrain this landscape — **no autocorrect in v1** (SPEC §11), **EN↔RU only** (§10), **no GUI settings** (§10). The table stakes below are therefore filtered through "what does a *hotkey-based* corrector need", not "what does Punto Switcher bundle".

### Table Stakes (Users Expect These)

Every serious competitor (Punto, Caramba, Mahou, SimpleSwitcher, WaylandSwitcher, langSwitcher) has all of the top items. Missing them = the tool feels broken to anyone migrating from Punto-likes.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Correct last word by hotkey | The single defining gesture of the category (Punto: Pause; Caramba: double-Shift; Mahou: Pause; WaylandSwitcher: PAUSE) | MEDIUM | Requires invisible typed-text buffer + word-boundary detection (back to nearest separator). Spec: double right shift (§4.1) |
| Correct whole phrase/line by hotkey | Present in Punto, Mahou (Shift+Pause), WaylandSwitcher (Shift+PAUSE), langSwitcher ("to line start") | MEDIUM | Requires buffer-reset semantics to define phrase boundaries. Spec: triple right shift (§4.1) |
| Correct selected text | Punto (selection has priority over last word), Mahou (Scroll Lock), WaylandSwitcher, mrmakko KDE switcher | MEDIUM | Reads selection via clipboard round-trip (WaylandSwitcher approach) or IBus surrounding-text API (spec Q2). Spec: double right shift with active selection (§4.1) |
| Switch layout by hotkey | A "switcher" that cannot switch is not a switcher; Caramba: single Shift; Mahou: CapsLock | LOW | Proxy to GNOME Super+Space via `org.gnome.desktop.input-sources`. Spec: right shift (§4.1) |
| Case preservation per character | `GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет`, `ghbdtn`→`привет` is the canonical demo of every converter since Punto | LOW | Standard technique everywhere: lowercase layout map + per-character case restore; mixed shapes (`gHBdtn`→`пРивеТ`) fall out for free. Spec §4.3 |
| Full layout tables incl. punctuation | `[ ] ; ' , . /` and digits differ between ЙЦУКЕН and QWERTY; converters that skip punctuation feel broken mid-sentence | LOW | Pure table data + go:generate (spec §8 `layouts/`). Punctuation/digits pass through mapping, not "untouched" |
| Auto direction detection | Users do not want to think about which direction to convert; all converters detect by buffer composition | LOW | EN/RU scripts are disjoint → majority-of-letters rule suffices (Punto's statistical dictionaries are for *auto* mode, not needed here). Spec §4.2 |
| Mixed-text handling | Real text contains both scripts; users expect only the wrong-layout portion to be converted | HIGH | Hardest table-stakes item: segmentation + "relative to last-used layout" semantics (spec §4.2 open question). langSwitcher precedent: find where wrong layout begins, convert that span |
| Buffer reset rules | Stale buffers produce absurd corrections; Punto clears on click, spec adds Enter/Tab/Escape/focus change | MEDIUM | Requires focus/change events from IBus + mouse-click detection (click is the tricky one on Wayland). Spec §4.3 |
| Configurable hotkeys | Every competitor offers this; hardcoded keys fail laptop/keyboard variations (T2 MacBook Pro lacks some keys) | LOW | YAML config; hot reload is the differentiator, configurability itself is table stakes. Spec §4.1 |
| No root / easy install | WaylandSwitcher's root requirement is a documented liability | LOW | systemd user unit + `go install`. Spec §5 |
| Replace exactly the wrong range (no visual jump) | Users abandoned uinput tools because Backspace×N retyping visibly mangles text | HIGH | Core architectural bet; IBus `commit_text` / surrounding-text (spec Q2). This is what makes the IBus approach worth building |

### Differentiators (Competitive Advantage)

What no working GNOME Wayland competitor currently offers. These align with Core Value: "fix wrong-layout text in any GNOME Wayland input field via IBus, without root, without keyd/xremap conflicts."

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| IBus-engine level, no keyboard grab | WaylandSwitcher needs root + EVIOCGRAB (monopolizes keyboard, incompatible with keyd/xremap); goswitch coexists by construction | HIGH | The founding architectural decision (PROJECT.md Key Decisions). Verified: no other Wayland corrector works this way today |
| keyd/xremap compatibility | Owner's machine runs remappers; every existing Wayland switcher conflicts with them | MEDIUM | Follows from IBus level; must be validated in e2e matrix (spec §4.4) |
| Automated e2e test matrix (ydotool + AT-SPI, YAML cases, PASS/FAIL) | No competitor in the entire category has automated full-cycle tests; this is the project's quality moat and the acceptance gate | HIGH | Spec §7. Unique in category; also the metric for "done" |
| Switch-and-correct combo hotkey | One gesture instead of two; Caramba users cite single-Shift+double-Shift ergonomics as the reason they pay | LOW | Spec: Shift+RightCtrl (§4.1); flagged `[Q]` for muscle-memory conflict — keep, but confirm default |
| Multi-tap hotkeys (double/triple right shift) | Caramba's signature ergonomics (single Shift = switch, double = fix word) — familiar to the target audience, no extra keys consumed | MEDIUM | Spec §4.1. See dependency notes: single-vs-double on the *same key* needs explicit semantics; no OSS Linux tool ships triple-tap |
| YAML config hot reload (no restart) | Competitors use GUI dialogs or interactive configurators requiring restart | LOW | fsnotify + atomic swap; table stakes for servers, differentiator for desktop tools |
| Terminal fallback via wl-clipboard paste | Terminals are the known IBus weak spot (wezterm `use_ime` issues, kitty needs `GLFW_IM_MODULE=ibus`); WaylandSwitcher proves the clipboard path works on Wayland | MEDIUM | Spec Q3. Detection of "this app is a terminal" needed — IBus focus events carry app identity |
| Super+Letter → Ctrl+Letter per-app macros | No layout switcher does this; solves the terminal Ctrl+C conflict problem uniformly | MEDIUM | Spec §4.4 (second). Needs per-app context from IBus engine focus events. Risk of scope creep — keep per-app rule list in YAML |
| Single static Go binary, MIT, zero telemetry | Punto's diary/keylogger and Yandex integration are the most-cited trust complaints in the ecosystem | LOW | Positioning: "on-device, auditable, 2 dependencies" |
| < 50 ms hotkey reaction | Rekey markets "<200 ms" as fast; goswitch targets 4× better with a guaranteed bound | LOW | Follows from in-process IBus path; measure in e2e |
| Generalized layout-table interface | Other pairs (DE↔EN, etc.) become data, not code | LOW | v1 ships EN↔RU data only (§10), but the table API stays generic — cheap insurance |

### Anti-Features (Commonly Requested, Often Problematic)

Punto's bundle is a catalogue of scope traps. The spec already rejects the top four; ecosystem evidence confirms each rejection.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Autocorrect (silent, no hotkey) | Punto/Caramba/Rekey flagship mode; "why press anything?" | Owner-rejected deliberately (SPEC §11): passwords, code, no reliable per-app exclusions on Wayland. Ecosystem validates: Punto's most-cited complaint is *false switches*; Rekey/ Caramba need terminal/editor exclusion lists to make autocorrect tolerable — exactly what Wayland cannot provide reliably | Hotkey-only correction (v1 model). If ever revisited: separate option, default off, window-class white/black lists (SPEC §11 wording) |
| Exception heuristics (URL/e-mail/path/hex) | "Don't mangle my URLs" | Spec defers to v2 (§4.3) — heuristics are a correctness rabbit hole (hex vs words, IDs vs emails) and unnecessary while correction is user-triggered: the user can see the target before pressing | Defer to v2 as configured rules; document as known limitation in v1 |
| Other layout pairs in v1 | International appeal | EN↔RU is the owner's actual need; each pair needs tables + test corpus | Generic table interface, EN↔RU data only (§10) |
| GUI settings dialog | "Every tool has one" | Duplicate config surface; YAML is diffable, versionable, and already needs hot reload for tests | YAML only (§10); `goswitchctl status/reload` for feedback |
| Typed-text diary / keylogger | Punto shipped it; "search what I typed Tuesday" | Privacy catastrophe; single biggest trust killer in competitor reviews | Structured logs with key tracing only in explicit debug mode (§5) |
| Clipboard manager | Punto bundles 30-entry history | Different product; clipboard ownership rules on Wayland are hostile to it | Out of scope; wl-clipboard used only as transport |
| Transliteration (Привет→Privet) | Punto feature; looks like "the same table" | Different mapping semantics (phonetic vs positional); corrupts the core tables' simplicity | Rejected: positional layout mapping only |
| Typo correction (xneur-style) | xneur bundles it | Requires dictionaries per language; false positives; orthogonal to layout errors | None — layout correctness only |
| Snippets / auto-replacement / number-to-words | Punto extras | Each is its own product; inflate config surface and test matrix | None |
| Cloud / search integration on selection | Punto's Yandex/Wikipedia hooks | Telemetry concerns; network dependency in an input-path daemon | None; user's own tools |
| Layout state indicator / tray icon | Punto shows flags; users orient by it | GNOME Shell already shows the input-sources indicator; duplicating it in IBus engine is redundant work | Rely on GNOME's native indicator (spec reads layout from input-sources anyway) |

## Feature Dependencies

```
[IBus engine registration + key event interception]
    └──requires──> [Typed-text buffer tracking]
                        ├──requires──> [EN↔RU layout tables + case mapping]
                        │                  └──enables──> [Correct last word] ──requires──> [Direction auto-detect]
                        │                                                                      │
                        ├──requires──> [Buffer reset rules (Enter/Tab/click/Escape/focus)]        │
                        │                  └──enables──> [Correct whole phrase] <────────────────┤
                        │                                                                      │
                        └──enables──> [Mixed-text correction] <──requires── [Direction auto-detect]
                                                                                               │
[Multi-tap detector (timing layer)] ──triggers──> [Correct word / phrase]                       │
        │                                                                                       │
        └──conflicts──> [Switch layout <50ms]  (single vs double on same key — needs semantics) │
                                                                                                │
[GNOME input-sources integration] ──enables──> [Switch layout] ──enables──> [Switch+correct combo]
                                                │
[Selection acquisition (IBus surrounding text or clipboard)] ──enables──> [Correct selection]
        │
        └──shared with──> [Terminal fallback via wl-clipboard] ──requires──> [Per-app context from IBus focus events]
                                                                  │
                                                                  └──also enables──> [Super→Ctrl per-app macros]

[YAML config + hot reload] ──parameterizes──> (every hotkey feature)
[E2E test matrix] ──verifies──> (every feature above)
```

### Dependency Notes

- **Every correction feature requires the IBus engine + buffer:** there is no correction without event interception (spec Q1 gates everything; M1 exists to prove it).
- **Buffer reset rules gate phrase correction:** "the phrase" is only well-defined if Enter/click/focus marked its start; word correction can ship first (M2 before M3) exactly along this seam.
- **Selection correction and terminal fallback share the clipboard path:** if Q2 (IBus range replacement) resolves positively, selection uses the IBus API and clipboard remains only the terminal fallback; if negatively, both ride wl-clipboard (WaylandSwitcher precedent). Build the clipboard transport once either way.
- **Multi-tap detector conflicts with the <50ms switch latency:** single right shift = switch, double right shift = correct word, on the *same key*. Waiting ~300–500ms (industry-standard double-tap window; tools cluster there) to disambiguate breaks the <50ms bound for the single-tap switch. Two viable semantics: (a) fire switch immediately on first tap; a second tap within the window triggers correction of the buffer typed *before* the switch — direction auto-detect makes the result coherent; (b) fire-on-release with short window for the single-tap action only. Option (a) matches Caramba's observable behavior. **This must be pinned in requirements (spec phase 2), not discovered in code.**
- **Mixed-text correction is the highest-complexity table-stakes item:** requires per-word direction classification plus "last-used layout" state (spec §4.2 open question). Recommend explicit ADR with golden-test corpus before M3.
- **Per-app context (from IBus focus events) unlocks two features:** terminal detection (fallback routing) and Super→Ctrl macros. One mechanism, two consumers — build it once, in the engine adapter.
- **E2E matrix enhances everything:** not a runtime feature, but every feature's acceptance test; feature order in roadmap should track what the matrix can verify at each step (spec §7.2 already sequences this: M1 event-in-log → M2 word matrix → M3 full matrix).

## MVP Definition

### Launch With (v1)

Aligned with spec milestones M1–M4 (§9).

- [ ] IBus engine registration + hotkey event interception (M1) — proves the architecture; nothing else exists without it
- [ ] Typed-text buffer with reset rules (Enter, Tab, mouse click, Escape, focus change) — the correctness foundation
- [ ] Correct last word, EN↔RU, full tables incl. punctuation + case preservation (M2) — the single most-used gesture; validates Q2 (exact-range replacement)
- [ ] Direction auto-detect by buffer composition — without it users must think, and they won't
- [ ] Correct whole phrase (M3) — second-most-used gesture
- [ ] Correct selected text (M3) — clipboard path acceptable if Q2 unresolved
- [ ] Layout switch hotkey (proxy to GNOME Super+Space) — cheap once input-sources integration exists
- [ ] Multi-tap detection: double + triple right shift (M3) — Caramba ergonomics; semantics per dependency note above
- [ ] Switch+correct combo (Shift+RightCtrl) (M3) — trivial once both halves exist
- [ ] YAML config, hot reload, all hotkeys remappable (M3)
- [ ] Mixed-text correction, defined semantics (M3) — hardest item; if it slips, ship v1 with "whole buffer converts by majority" and document
- [ ] Install without root: systemd user unit, `go install`, IBus registration (M4)
- [ ] E2E matrix green (gnome-text-editor, gedit, Chrome minimum; spec §7.2) — the acceptance criterion itself

### Add After Validation (v1.x)

- [ ] Terminal fallback via wl-clipboard paste — after Q3 investigation; enabled per-app (terminals only)
- [ ] Super+Letter → Ctrl+Letter per-app macros (spec §4.4 second) — independent of correction path; ships when per-app context lands
- [ ] Exception rules (URL/e-mail/path/hex heuristics) — spec's v2 marker; only after real mistype corpus exists from dogfooding
- [ ] Performance telemetry: assert <50ms/<50MB in e2e report — turn the NFR into a tested property

### Future Consideration (v2+)

- [ ] Additional layout pairs — data-only once table interface proves itself (spec keeps interface generic for exactly this)
- [ ] Autocorrect, default off, per window-class lists — only if owner's stance changes (SPEC §11; recorded, not debated)
- [ ] Non-GNOME compositors (KDE/Sway) — architecturally allowed, untested (§10)

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Correct last word (tables + case) | HIGH | MEDIUM | P1 |
| Direction auto-detect | HIGH | LOW | P1 |
| Buffer + reset rules | HIGH | MEDIUM | P1 |
| IBus engine interception (no grab) | HIGH | HIGH | P1 |
| Exact-range replacement, no jump | HIGH | HIGH | P1 |
| Correct phrase | HIGH | MEDIUM | P1 |
| Correct selection | MEDIUM | MEDIUM | P1 |
| Layout switch hotkey | HIGH | LOW | P1 |
| Multi-tap (double/triple shift) | HIGH | MEDIUM | P1 |
| Switch+correct combo | MEDIUM | LOW | P1 |
| YAML config + hot reload | HIGH | LOW | P1 |
| Mixed-text correction | MEDIUM | HIGH | P1 (semantics ADR) / P2 (full) |
| E2E test matrix | HIGH (owner's gate) | HIGH | P1 |
| Terminal clipboard fallback | MEDIUM | MEDIUM | P2 |
| Super→Ctrl per-app macros | MEDIUM (owner-specific) | MEDIUM | P2 |
| Exception heuristics | LOW (hotkey model) | MEDIUM | P3 |
| Additional layout pairs | LOW (owner) | LOW | P3 |
| Autocorrect | — (rejected) | — | out of scope |

## Competitor Feature Analysis

| Feature | Punto Switcher (Win/macOS) | Caramba Switcher (macOS) | WaylandSwitcher (Nim, Wayland) | xneur / Mahou / SimpleSwitcher (Win/X11 OSS) | goswitch (plan) |
|---------|---------------------------|--------------------------|-------------------------------|-----------------------------------------------|-----------------|
| Correct trigger | Auto + Pause (manual) | Auto + double-Shift (manual) | PAUSE (manual only) | xneur: auto; Mahou/SimpleSwitcher: Pause/ScrollLock (manual or auto) | **Hotkey only** (double/triple right shift) — autocorrect rejected |
| Correct scope | Last word / selection (priority) / phrase | Last typed word | Word / phrase (Shift+PAUSE) / selection | Mahou: word, line, selection | Word / phrase / selection (spec §4.1) |
| Layout switch | Shift+Break; tray | Single Shift | None (uses system key) | Mahou: CapsLock; SimpleSwitcher: OS cycle or per-language key | Right shift → GNOME Super+Space proxy |
| Direction detect | Statistical + dictionaries | Auto (per-word) | Manual per hotkey | Majority heuristics | Majority-of-buffer (scripts disjoint) + mixed-text split |
| Case handling | Preserved; Alt+Break cycles case | Preserved | Case *change* via Shift+PAUSE | Preserved (SimpleSwitcher converts case too) | Preserved per character (§4.3) |
| Insertion mechanism | Backspace×N + retype | Native (macOS CGEvent) | uinput retype; selection via wl-clipboard | SendInput retype | **IBus commit_text exact range** (Q2) + terminal clipboard fallback (Q3) |
| Privileges | User | User | **Root + EVIOCGRAB** | User | User, systemd user unit |
| Remapper coexistence (keyd/xremap) | n/a | n/a | **No — keyboard monopoly** | n/a | **Yes — IBus layer** |
| Per-app rules | Exceptions tab | Exclusion lists (games, IDEs) | None | SimpleSwitcher/Mahou: app-specific | Per-app only for Super→Ctrl macros + terminal fallback routing |
| Config | GUI dialogs | GUI | Interactive configurator, no file | GUI / config file | **YAML + hot reload, no GUI** |
| Tests | None public | None public | None | None | **E2E matrix = acceptance gate** |
| Trust | Diary keylogger + Yandex cloud cited as concerns | Subscription | MIT, but binary-first distribution | OSS | MIT, on-device, auditable, 2 deps |

## Sources

- [Balans097/WaylandSwitcher README](https://github.com/Balans097/WaylandSwitcher) — primary source, feature list fetched directly (HIGH confidence for that column)
- [Grokipedia: Punto Switcher](https://grokipedia.com/page/Punto_Switcher) — full feature inventory and history, incl. Caramba lineage (Moskalyov left Yandex 2017, released Caramba 2018)
- [Rekey comparison of Punto alternatives for macOS](https://trishchuk.com/rekey/alternatives/punto-switcher/) — Caramba/Boomkey/MLSwitcher feature matrix (promotional source; facts cross-checked)
- [Caramba Switcher on App Store](https://apps.apple.com/us/app/caramba-switcher-autocorrect/id1565826179) and [AutoHotkey forum discussion of Caramba hotkeys](https://www.autohotkey.com/boards/viewtopic.php?t=131938) — single-Shift/double-Shift semantics
- [xneur man page (Ubuntu)](https://manpages.ubuntu.com/manpages/jammy/man1/xneur.1.html) — auto/manual modes
- [Mahou (BladeMight)](https://github.com/iamkarlson/Mahou), [SimpleSwitcher (Aegel5)](https://github.com/Aegel5/SimpleSwitcher) — Windows OSS hotkey conventions
- [reg2005/langSwitcher](https://github.com/reg2005/langSwitcher), [rashn/RuSwitcher](https://github.com/rashn/RuSwitcher), [dspinellis/kbd-layout-fix](https://github.com/dspinellis/kbd-layout-fix) — conversion semantics, mixed-phrase handling
- [autokbisw](https://github.com/ohueter/autokbisw) — per-app layout memory precedent
- [wezterm use_ime docs](https://wezterm.org/config/lua/config/use_ime.html), [wezterm issue #5125](https://github.com/wezterm/wezterm/issues/5125), [kitty GLFW_IM_MODULE workaround (Debian)](https://lists.debian.org/debian-input-method/2021/06/msg00048.html) — terminal IME support facts (Q3)
- [keyd issue #219](https://github.com/rvaiya/keyd/issues/219), [Karabiner double-tap tutorial](https://agileadam.com/2024/11/double-tap-modifier-hotkeys-in-any-application/), [Keyboard Maestro double-tap](https://forum.keyboardmaestro.com/t/double-tap-control/202) — multi-tap detection windows and latency tradeoffs
- [Stack Overflow: Latin↔Cyrillic symbol translation](https://stackoverflow.com/questions/78010107/how-to-translate-symbols-from-latin-to-cyrillic) — per-character case restoration technique
- [grafov/shift-shift](https://github.com/grafov/shift-shift), [mrmakko/Keyboard-layout-switcher-Linux-KDE](https://github.com/mrmakko/Keyboard-layout-switcher-Linux-KDE), [PolterType](https://poltertype.com/blog/wrong-layout-typing-on-wayland/) — adjacent Wayland tools

---
*Feature research for: GNOME Wayland layout switcher/corrector (IBus engine, Go)*
*Researched: 2026-09-10*
