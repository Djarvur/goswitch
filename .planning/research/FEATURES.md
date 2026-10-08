# Feature Research — goswitch v1.2 «Терминалы и видимость приложений»

**Domain:** terminal text-correction + per-app visibility for an IBus input-method daemon (GNOME Wayland)
**Researched:** 2026-10-08
**Confidence:** MEDIUM overall (live target-system verification HIGH; web ecosystem claims MEDIUM — cross-checked, no single-source claims presented as authoritative)

## Feature Landscape (v1.2 scope)

The milestone has three features: **TERM** (correction in terminals + wezterm switching defect), **MENU** (dynamic tray «Добавить ИмяПриложения в чёрный список»), **A11Y** (Telegram/ZCode autocorrect via enabling accessibility inside those apps). Existing v1.0/v1.1 capabilities are NOT re-researched (word/phrase/selection correction, mode flip, tray+menu, autocorrect blocklist, a11y enabler, sounds, YAML hot reload).

### Table Stakes (owner expects these — missing = milestone failed)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **TERM-0: layout flip works in wezterm** | Owner's daily terminal is wezterm (2026-10-08); flip works everywhere else today. This is a defect, and it is a **prerequisite** for any correction in wezterm. | MEDIUM–HIGH (needs a live spike) | **Live-verified root-cause picture:** owner's `~/.wezterm.lua` sets `enable_wayland = false` → wezterm 20240203-110809-5046fc22 runs as an **XWayland/XIM client** (`XMODIFIERS=@im=ibus`, window confirmed on the X server via WM_CLASS). With ADR-006 both GNOME input sources are goswitch engines (`sources = [('ibus','goswitch-en'),('ibus','goswitch-ru')]`, verified live) — an engine flip never changes the XWayland xkb keymap, and XIM clients derive characters from the keymap, so the flip is invisible in wezterm. GTK/Qt apps follow the IBus engine via IM modules instead. First fix candidate is a **config recipe**: run wezterm native Wayland (`enable_wayland = true`) where the engine channel is `zwp_text_input_v3` (supported since wezterm 20220807) and engine commits/forwards behave like GTK. Known risk history: wezterm#1772 (text-input-v3 quirks), wezterm#5125 «IME Not Working with IBus» (fixed-in-nightly; affected exactly the 20240203 build), gnome-shell bug (IME in wezterm dies after screen unlock until refocus). If native Wayland misbehaves, a daemon-side XIM workaround is the fallback. A phase-1 spike on the live machine is mandatory — this cannot be settled from docs. |
| **TERM-1: word correction in terminals (mandatory)** | Core value «в любом поле ввода» stops being true in the owner's most-used app; SPEC anticipated it (Q3: wl-clipboard + paste-emulation fallback). | HIGH | Terminals lack surrounding-text, so v1.x GUI paths do not transfer mechanically. Two sub-paths: (a) **own keystroke buffer + `commit_text`** — the daemon sees every key since focus-in (both sources are goswitch engines), so in GNOME Terminal/VTE (GTK IM module — commit works there) word correction may already be close to working; needs a spike «does buffer+commit correction already work in VTE?». (b) **universal select→read→rewrite path** (the milestone plan): inject selection keys (Ctrl+Shift+Left) via the daemon's IBus input-context channel (`ForwardKeyEvent` — the live 2026-10-08 spike showed the daemon channel is race-free, unlike an external injector), read the freshly-made selection from the Wayland **primary selection** (`wl-paste -p`; wl-clipboard 2.2.1 installed — reading primary avoids clobbering the user's clipboard), compute inversion, rewrite via shell-edit keys (Backspace ×N, then paste). NEVER replace by «typing over the selection»: a terminal selection is an emulator-layer concept the shell cannot see or delete. |
| **TERM-2: whole typed line correction (if feasible)** | Phrase exists in GUI; users type whole commands wrong. | HIGH | Realistic terminal semantics = **whole current line**, not multi-line phrase: the correction buffer clears on Enter (ADR-004), and bracketed paste means multi-line write-back is safe-but-awkward (shows PS2 continuations, not executed until Enter). Flow: `End` → read line (Shift+Home + primary) → `Ctrl+U` (kill line in bash/zsh defaults; in fish, after `End` the kill-to-start effect is equivalent) → paste corrected. Cursor was mid-line? Point lands at end-of-line after rewrite; restoring the original position requires replaying `ArrowLeft` ×(len_right) — computable from buffer+clipboard lengths, cheap, should be done or mid-line users lose position. |
| **MENU-1: dynamic tray item «Добавить ИмяПриложения в чёрный список»** | Owner-requested replacement for showing the app in `goswitchctl status`; one-click exclusion is the expected UX of any autocorrect tool (Punto ships a per-app exceptions GUI). | LOW–MEDIUM | All building blocks exist: own StatusNotifierItem+DBusMenu (v1.1 tray), `internal/appid.FocusedApp()` (AT-SPI bridge-namespace identity), `internal/config/writer.go` + hot reload. Work = menu-layout revision on focus change (DBusMenu `AboutToShow`/`LayoutUpdate`) + regex append + reload. **Degradation rule required:** apps without an AT-SPI application bridge (notably wezterm/winit on XWayland — invisible in the a11y tree) never refresh `FocusedApp` → item must hide/grey out, never show a stale app. |
| **A11Y-1: autocorrect in Telegram (Qt snap) and ZCode (Electron)** | Owner decision 2026-10-08: daemon unchanged; enable accessibility inside the apps via recipes/docs/installer. | LOW (code) / MEDIUM (verification) | **Electron/ZCode:** renderer a11y is exposed only when launched with `--force-renderer-accessibility` (external forcing requires restart by design; `app.setAccessibilitySupportEnabled(true)` is callable only from inside the app) → recipe = **user-level `.desktop` override** in `~/.local/share/applications/` adding the flag (no root). Tooling reports (cua.ai; oh-my-pi issue) say AT-SPI settings alone are insufficient for Chromium/Electron. **Qt/Telegram:** Qt exposes AT-SPI when the a11y bus is active (`QT_LINUX_ACCESSIBILITY_ALWAYS_ON=1` or desktop toolkit-accessibility — the daemon's enabler already sets the latter); Telegram is a snap → override via a user-level copy of the snap desktop file with env, or a documented `snap run` alias. **Known cosmetic side effect:** Telegram shows a false-positive yellow «Screen Reader mode» bar on Linux when a11y looks enabled (GitHub issue 2026-04) — document it in README. Telegram a11y quality itself is poor upstream (screen readers read nothing; Zendalona fork exists) — verify empirically that the a11y tree exposes the input field before promising autocorrect there; if Telegram does not submit, owner reserved the right to revisit daemon settings. |

### Differentiators (competitive advantage)

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Terminal correction at all, on Wayland, daemon-level | **No existing Linux tool does it.** layout_corrector (Linux Punto clone) README: «using inside terminal emulator is not supported»; xneur is X11-only keysym remapping with per-app ignore lists (dormant); Punto Switcher is Windows/macOS. A dedicated terminal flow is the milestone's headline. | — | The table-stakes rows ARE the differentiator vs the Linux field; entries below are differentiators *within* the design. |
| Correction works over ssh and across shells uniformly | Key-level channel is client-agnostic: select+read+rewrite works the same in an ssh session inside a local terminal (selection and injection are local terminal-layer operations). Shell plugins (bash `READLINE_LINE`, zsh ZLE `BUFFER`, fish `commandline`) give perfect UX but are per-shell, per-machine installs and cover nothing outside the line editor. | — | Document as a feature; do not chase shell-plugin parity (theirs is install-bound, ours is system-bound). |
| Whole-line rewrite with cursor restoration | Even Punto's Alt+Break only converts «last typed text». Rewriting the full line and replaying cursor position is beyond all compared tools. | MEDIUM | Feasibility-flagged (TERM-2); word first. |
| One-click blocklist from the tray | Punto/xneur require digging into a settings GUI; goswitch adds the focused app to `autocorrect.apps_blocklist` in one click with instant hot-reload. | — | MENU-1 as shipped covers this. |

### Anti-Features (explicitly NOT building — with reasons)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **Autocorrect in terminals** | Punto does autocorrect everywhere | Owner decision (no). Hidden-input contexts (sudo/ssh password prompts) are indistinguishable from a prompt at the key level; auto-rewriting an invisible password line is a corruption hazard; terminals fail the role-gate by design (fail-closed silence). | Manual hotkey correction only; document that autocorrect stays globally disabled in terminals. |
| **Backspace-replay for pty-level deletion as the default write-back** | Already built (never driven); looks like the «clean» deletion primitive | N blind Backspaces bet on shell keybindings, vi-mode, and TUI apps; latency grows with length; position assumptions break under bracketed paste and multi-byte edits. | Paste-over rewrite via shell-edit keys (Ctrl+U / Backspace×N for the word tail only) + bracketed-paste insert; keep backspace-replay parked. |
| **Per-app layout switching hacked inside wezterm (Lua key tables / app-side remaps)** | wezterm is scriptable enough to fake layout switching | Duplicates layout state in a second place; fights the daemon's single-source model (ADR-006); unmaintainable. | Fix the transport (native Wayland text-input-v3 or daemon-side XIM handling) so wezterm follows the global engine like every other app. |
| **Un-blacklist counterpart in the tray (remove app from blocklist)** | Symmetry | Menu bloat; removal is rare and config-edit + hot reload already exists and is documented. | Edit `autocorrect.apps_blocklist` in YAML. |
| **Selection preservation after terminal correction** | Nice touch | Explicitly out of scope for v1.2 (owner: rollback via Right Ctrl+Shift already works in GUI contexts). | Deferred; note in WINDOWS ledger. |
| **«Не трогать» heuristics for URLs/paths in terminal correction** | Terminals are full of paths/URLs | v2 scope (CORR-10, owner roadmap). | Not in v1.2. |

## How terminal text-correction tools actually work (ecosystem answer)

Five approach classes exist; goswitch's planned flow is the fifth (daemon-level, channel-aware):

1. **Keyboard+clipboard emulation, app-agnostic (Punto Switcher canon).** Hotkeys: `Shift+Break` = convert selected text, `Alt+Break` = convert last typed text. Mechanism: select → cut to clipboard → per-char convert → paste back. Works in Windows consoles precisely because it touches only keys+clipboard. This is exactly the canonical flow goswitch plans for terminals — 20 years of Punto UX expectations stand behind it (the owner's muscle memory is this UX).
2. **X-level keysym remapping (xneur, Linux).** Sits under the X server (XRecord), swaps keysyms before apps see them, per-app ignore lists. On-the-fly remap technically «works» in terminals, but after-the-fact correction has the same selection problems; X11-only → dead on GNOME Wayland; project dormant. Not a usable precedent — only a UX reference.
3. **Refuse terminals (layout_corrector, coffebar — archived 2022-10).** X11 tool (xsel/xdotool/xvkbd); README: terminal emulators not supported. Direct evidence that a generic GUI-correction flow breaks in terminals and that a **dedicated** terminal flow is genuinely novel.
4. **Correct inside the shell line editor (shell plugins).** bash: `bind -x` + `READLINE_LINE`/`READLINE_POINT`; zsh: ZLE widget editing `BUFFER`/`CURSOR`; fish: `bind` + `commandline`. Gold-standard UX (prompt-safe, cursor-correct) but per-shell, per-machine, and blind outside the line editor (no REPL/TUI/other-terminal coverage). No dominant RU-layout plugin exists (oh-my-zsh ships only Colemak/vi-mode remaps) — the niche is open.
5. **IM-engine channel with terminal adaptation (goswitch plan).** Engine sees keys; where surrounding-text/commit works (VTE) the v1.x paths already nearly apply; where they do not, select→primary→rewrite through the daemon's own `ForwardKeyEvent` channel. Uniquely covers ssh-in-terminal and every shell at once.

## Expected user-visible behavior in a terminal (UX contract the roadmap must honor)

- **Prompt:** never touched. All rewrite keys are line-editor keys (`Home/End/Backspace/Ctrl+U` + printable paste); the shell redraws the line exactly as if typed very fast. No screen escape sequences, no `printf` into the tty — the daemon has no tty and must not want one.
- **Cursor mid-line:** word correction = select word left of point (Ctrl+Shift+Left), rewrite in place → point ends right after the corrected word (natural). Whole-line rewrite = `End` → `Ctrl+U` → paste → point at end-of-line; restore the original offset by replaying `ArrowLeft` × len(right-side text) — lengths are known (buffer + clipboard), deterministic.
- **Readline/ZLE specifics:** line editors own the buffer; external programs have no replace-API (`rl_replace_line` exists only in-process) — synthesized keystrokes are the only honest channel. Ctrl+U kills the buffer only (prompt intact) in bash/zsh defaults.
- **Multi-line / bracketed paste:** bash ≥ 8.1 (readline) enables bracketed paste by default; zsh since 5.1; fish native. Pasted multi-line text is wrapped, shows PS2 continuations, and is NOT executed until an explicit Enter — safe write-back. Because the daemon's buffer clears on Enter (ADR-004), «phrase» in a terminal degenerates to the current logical line; a phrase spanning continuation lines is out of reach and should be declared out of scope.
- **Known failure modes to design against:**
  - Terminal selection is **emulator-layer**: the shell cannot delete a selection; «type over selection» does nothing to the shell buffer. Replacement must always be shell-edit keys + insert.
  - **zsh-autosuggestions** render text beyond the logical buffer to the right of the cursor; `Shift+End` would capture suggestion tail into the selection. Leftward selection (Shift+Home, Ctrl+Shift+Left) is immune — select left, never right.
  - **Screen-wrapped long lines:** selection geometry is screen rows, buffer is logical; anchor selection at cursor/prompt (`End`+`Shift+Home`), never compute from screen geometry.
  - **Tabs / wide chars:** screen column ≠ buffer length when the line contains tabs (Cyrillic is single-width — safe). Never trust screen-derived lengths over buffer lengths.
  - **TUI apps (vim, less, fzf):** injected Ctrl+U/Backspace mean different things (vim Ctrl-key semantics). v1.2 scope = «works at the shell prompt»; document that behavior inside full-screen TUI apps is undefined. Do not attempt TUI detection in v1.2 (no reliable signal at key level).
  - **Empty/no-selection, nothing typed:** no-op with existing sounds as feedback (reuse v1.1 sound grammar); never paste a converted empty string.
  - **Password prompts (sudo/ssh):** keys arrive but are not echoed; a *manual* correction there is user-invoked and allowed, but the daemon cannot distinguish; keep buffer-clear-on-Enter and the fail-closed role-gate intact so nothing automatic ever fires.

## wezterm per-app layout switching: approaches and failure modes

| Approach | Status | Verdict |
|----------|--------|---------|
| **Native Wayland + `zwp_text_input_v3`** (`enable_wayland = true`, upstream default) | Supported since wezterm 20220807; mutter is the IM frontend; engine `commit_text`/`ForwardKeyEvent` flow to wezterm like to GTK apps | **Primary fix candidate.** Risks: #1772 (empty-preedit/commit semantics), #5125 (IME broken with IBus on exactly the 20240203 build — fixed in nightlies, so also check upstream fix status / suggest build bump), gnome-shell «IME dies in wezterm after unlock» (compositor-side, refocus restores). Must be spike-tested live on this machine before committing the phase plan. |
| **XWayland + XIM** (owner's current config) | XMODIFIERS=@im=ibus; ibus XIM bridge forwards core key events | **The defect state.** Forwarded keys are physical keycodes; the client resolves characters via the X keymap, which never flips (both sources are ibus engines). Layout switching is structurally invisible here; XIM is also the least-maintained IBus client path. Not fully fixable from goswitch alone — requires either moving wezterm to Wayland or daemon-side workarounds of unknown robustness. |
| **App-side emulation** (wezterm Lua key tables, IM emulation in config) | Possible but anti-pattern | Duplicates layout state; fights ADR-006 single-source; rejected (see Anti-Features). |
| **Upstream build bump** (wezterm nightly fixes for #5125) | Unknown on 24.04 for this install | Secondary: if the spike shows the 20240203 build's Wayland IME is broken but fixed later, recommend the build change together with the recipe. |

Live-verified environment facts (HIGH confidence, target machine, 2026-10-08): wezterm 20240203-110809-5046fc22, `enable_wayland = false` in `~/.wezterm.lua`, wezterm window present on XWayland (WM_CLASS `org.wezfurlong.wezterm`); gnome-terminal 3.52.0 (VTE); wl-clipboard 2.2.1 (`wl-paste -p` available); GNOME Wayland session (`ubuntu:GNOME`); ibus 1.5.x running; `sources = [('ibus','goswitch-en'),('ibus','goswitch-ru')]`, `current = 0`.

## Feature Dependencies

```
[TERM-0 wezterm flip defect] ──prerequisite──> [TERM-1 word correction IN WEZTERM]
        │                                            │
        │ (VTE path independent)                     └──requires──> [ForwardKeyEvent channel + wl-clipboard/primary read]
[TERM-1 word correction] ──enhances──> [TERM-2 whole-line correction]
        └──conflicts──> (autocorrect in terminals — prohibited by owner; role-gate stays fail-closed)

[TERM-2] ──requires──> [selection read path + shell-edit rewrite keys + cursor restore]
[TERM-2 phrase-across-Enter] ──conflicts──> [ADR-004 buffer-clears-on-Enter] → scope to current line only

[MENU-1 dynamic blocklist item] ──requires──> [internal/appid.FocusedApp (exists)]
                               ──requires──> [internal/config writer + hot reload (exists)]
                               ──requires──> [own SNI/DBusMenu (exists)]
[MENU-1] ──degrades──> (no AT-SPI bridge, e.g. wezterm/XWayland → hide item; XWayland identity fallback = later)

[A11Y-1 Telegram/ZCode] ──requires──> [a11y enabler (exists) + AT-SPI role-gate (exists)]
        ──requires──> [user-level .desktop overrides / snap recipe; restart of target app]
        ──constrains──> Telegram warning-bar side effect must be documented
```

### Dependency Notes

- **TERM-0 before TERM-1-in-wezterm:** if the engine channel does not reach wezterm (XIM state today), neither selection injection nor commit/paste write-back can work there. The defect fix is the enabling phase, whichever form it takes (config recipe vs daemon code). VTE word correction may be independently spikable — keep the two tracks separable in planning.
- **TERM-2 enhances TERM-1:** the select→read→rewrite machinery is built once for the word and reused for the line (`End`+`Shift+Home` instead of Ctrl+Shift+Left; `Ctrl+U` instead of Backspace×N; cursor-restore replay). Sequencing word→line is natural.
- **MENU-1 needs no new infrastructure** — it consumes three existing capabilities; its only real design decision is the no-identity degradation rule.
- **A11Y-1 conflicts with nothing** in the daemon (unchanged by owner decision), but its success criteria live inside the target apps (does the AT-SPI tree actually expose Telegram's message field? does ZCode expose its editor?) — e2e assertions must be written against live trees, not documentation.

## MVP Definition (v1.2)

### Launch With (v1.2 mandatory)

- [ ] **TERM-0** wezterm layout flip works (root cause fixed or recipe shipped + verified live) — without it the owner's daily terminal stays broken and TERM there is unreachable
- [ ] **TERM-1** word correction at the prompt in wezterm + GNOME Terminal/VTE (mandatory per milestone; autocorrect stays off in terminals)
- [ ] **MENU-1** dynamic «Добавить ИмяПриложения в чёрный список» with correct hide-when-unknown behavior
- [ ] **A11Y-1** recipes/docs/installer steps enabling a11y in Telegram + ZCode, verified against live AT-SPI trees (Telegram warning bar documented)
- [ ] e2e: matrix extensions for both terminals (word case at minimum), self-activating target window per project quality rule

### Add After Validation (v1.2 stretch / v1.2.x)

- [ ] **TERM-2** whole-line correction with cursor restoration — trigger: word correction green in both terminals and the select-read-rewrite machinery stable
- [ ] XWayland fallback identity for MENU-1 (wezterm in the menu) — trigger: owner annoyance; likely via an XWayland WM_CLASS probe as a new identity source behind the existing `appid` seam
- [ ] Daemon-side XIM workaround if native-Wayland wezterm proves unusable — trigger: TERM-0 spike verdict

### Future Consideration (v2+)

- [ ] Selection preservation after correction (explicitly rolled back from v1.2)
- [ ] CORR-10 «не трогать» heuristics (URL/e-mail/paths) — terminals raise its priority but scope is unchanged
- [ ] Phrase spanning continuation lines in terminals (blocked by ADR-004 buffer semantics; revisit only with a spec-delta)

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| TERM-0 wezterm flip fix | HIGH (daily-driver defect) | MEDIUM–HIGH (spike first) | P1 |
| TERM-1 word correction in terminals | HIGH (core promise in the owner's main app) | HIGH | P1 |
| MENU-1 dynamic blocklist item | MEDIUM–HIGH (owner-requested UX) | LOW–MEDIUM | P1 |
| A11Y-1 Telegram/ZCode recipes | MEDIUM (re-enables autocorrect in 2 daily apps) | LOW code / MEDIUM verify | P1–P2 |
| TERM-2 whole-line correction | MEDIUM–HIGH | HIGH | P2 |
| XWayland identity fallback for menu | LOW–MEDIUM | MEDIUM | P3 |
| Selection preservation | LOW (rollback exists) | MEDIUM | P3 (deferred, owner) |

**Priority key:** P1 = must have for the v1.2 release; P2 = should have, planned stretch; P3 = nice to have, future consideration.

## Competitor Feature Analysis

| Feature | Punto Switcher (Win/mac) | xneur (Linux X11) | layout_corrector (Linux, archived) | Shell plugins (bash/zsh/fish) | goswitch v1.2 plan |
|---------|--------------------------|-------------------|-------------------------------------|-------------------------------|--------------------|
| Correction in terminals | Yes — select/cut/convert/paste emulation | On-the-fly keysym remap; after-the-fact flaky; per-app ignore lists | **Refused** (README) | Perfect in-buffer, per-shell only | Dedicated flow via engine channel |
| Wayland support | n/a (different OS) | No (X11 only, dormant) | No (X11 only) | Yes (shell-level, compositor-agnostic) | Yes (IBus engine + Wayland clipboard) |
| Last-typed convert hotkey | Alt+Break | Yes (X-level) | ScrollLock script | ZLE/READLINE widgets | Exists since v1.0 (double/triple Shift) |
| Whole-line convert | Selected text only | Limited | Limited | BUFFER rewrite (zsh/bash) | TERM-2 (planned, with cursor restore) |
| Per-app exceptions | Settings GUI | Ignore lists | — | — | YAML `apps_blocklist` + **tray one-click** (MENU-1) |
| Works over ssh | Yes (keys are local) | No (needs X) | No | Only where installed | Yes (local key+selection channel) |
| Autocorrect with guard rails | Yes, everywhere | Yes, X-level | No | No | Yes v1.1, **never in terminals** (fail-closed role-gate) |

## Sources

- **Live target-system verification (2026-10-08, HIGH):** `~/.wezterm.lua` (`enable_wayland = false`), wezterm 20240203 binary + WM_CLASS on XWayland; `gsettings get org.gnome.desktop.input-sources` → both sources goswitch engines; wl-clipboard 2.2.1 installed; gnome-terminal 3.52.0; `XDG_SESSION_TYPE=wayland`; ibus-daemon running; repo code `internal/appid/appid.go` (FocusedApp = AT-SPI bridge namespace), `internal/config/writer.go`, `docs/CONFIG.md` (blocklist semantics), `docs/SPEC.md` (terminal fallback wl-clipboard + paste emulation anticipated)
- [wezterm Keyboard Concepts](https://wezterm.org/config/keyboard-concepts.html), [use_ime docs](https://wezterm.org/config/lua/config/use_ime.html) — IME paths, defaults, build history — **MEDIUM** (official docs, cross-checked)
- [wezterm#1772 — wayland text-input-v3 support](https://github.com/wezterm/wezterm/issues/1772), [wezterm#5125 — IME Not Working with IBus](https://github.com/wezterm/wezterm/issues/5125) (20240203 build, fixed-in-nightly), gnome-shell issue — IME dies in wezterm after unlock — **MEDIUM** (issue trackers)
- [coffebar/layout_corrector](https://github.com/coffebar/layout_corrector) — Linux Punto analog, X11-only, «terminal not supported», archived 2022 — **MEDIUM→HIGH** (README read directly)
- Punto Switcher mechanics — [4PDA](https://4pda.to), [kopirajter.by](https://kopirajter.by), [chipset.kiev.ua](https://chipset.kiev.ua), [official page](https://yandex.ru) — Shift+Break/Alt+Break select-cut-convert-paste — **MEDIUM** (multiple consistent secondary sources)
- [Julia Evans — Entering text in the terminal is complicated](https://jvns.ca/blog/2024%2F07%2F08/readline), Unix.SE — changing line contents programmatically — readline buffer/point/redisplay model — **MEDIUM**
- [cirw.in — bracketed paste](https://cirw.in/blog/bracketed-paste), [Red Hat Bugzilla #2079078 — readline 8.1 default](https://bugzilla.redhat.com/show_bug.cgi?id=2079078) — **MEDIUM**
- [Electron Accessibility docs (app.setAccessibilitySupportEnabled)](https://electronjs.org/docs/latest/tutorial/accessibility), [ArchWiki Accessibility](https://wiki.archlinux.org/title/Accessibility), [cua.ai — Inside Linux computer-use](https://cua.ai/blog/inside-linux-computer-use) — Electron/Chromium a11y exposure requirements — **MEDIUM**
- Telegram Desktop a11y — [bugs.telegram.org report](https://bugs.telegram.org), [Zendalona Accessible Telegram](https://zendalona.com), 2026 false-positive «Screen Reader mode» warning-bar issue — **MEDIUM/LOW** (community reports; verify live in-phase)
- [wl-paste man page](https://man.archlinux.org/man/wl-paste.1.en), [primary-selection-unstable-v1 protocol](https://wayland.app/primary-selection-unstable-v1) — clipboard vs primary, `-p` flag — **MEDIUM** (man pages; local binary verified)

---
*Feature research for: goswitch v1.2 «Терминалы и видимость приложений»*
*Researched: 2026-10-08*
