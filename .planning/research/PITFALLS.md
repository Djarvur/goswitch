# Pitfalls Research

**Domain:** goswitch v1.2 «Терминалы и видимость приложений» — terminal correction (TERM), tray-menu blocklist item (MENU), Telegram-snap/ZCode a11y enabling (A11Y), added to a working IBus-engine daemon on GNOME 46 Wayland (Ubuntu 24.04, T2 MacBook)
**Researched:** 2026-10-08
**Confidence:** MEDIUM (web findings are LOW unverified → MEDIUM where cross-checked across ≥2 independent sources per the classify-confidence seam; facts from the project's own ledgers/live spikes — wezterm caps 0x9, IME arming race, daemon-channel spike 2026-10-08 — are HIGH, target-runtime evidence)

**Scope discipline:** pitfalls already catalogued in v1.0 research and the WINDOWS ledger are referenced, not re-derived: GTK4-Wayland forwarded-events waiver (WINDOWS #4), session drift after mode flips (WINDOWS #13), wl-copy fork deadlock on pipes, ydotool 0.1.8 traps, inotify exhaustion → 2 s stat-poll hot reload, GNOME focus-stealing prevention (WINDOWS #7 closure). This file covers what is *new* in v1.2.

**Phase vocabulary:** v1.2 phase IDs are not yet assigned (REQ-IDs pending per PROJECT.md); this file uses the PROJECT.md feature areas — **TERM**, **MENU**, **A11Y** — with sub-stages (`-0` = spike/ADR before feature code). Roadmap should read "Phase to address" as feature-area + order hint; a recommended sequencing is in the mapping table at the end.

---

## Critical Pitfalls

### Pitfall 1: The selection-capture key (Shift+Home) does not create a selection in VTE — per-terminal key semantics are unverified assumptions

**What goes wrong:**
The whole-line capture design forwards Shift+Home through the IBus input context and then reads PRIMARY via wl-clipboard. But terminal emulators have **no native keyboard text selection**: GNOME Terminal/VTE has an open feature request since 2002 (GNOME GitLab #588; VTE #648) — Shift+Home/Shift+arrows are *passed through to the application inside*, not intercepted by the terminal. In bash/readline Shift+Home is typically unbound → a no-op (or a cursor jump in configs that bind it). Result: no PRIMARY update, wl-paste --primary returns the *user's stale selection* (possibly another window's text), and the daemon "corrects" against garbage — deleting N characters that were never selected.

**Why it happens:**
GUI-editor semantics (selection keys mutate a selection object) do not exist at the terminal layer; the corrector's authors transfer GtkEntry habits to the pty world. Each terminal family (VTE, wezterm key tables, foot/kitty with kitty-keyboard protocol) resolves the same key differently, and the daemon cannot distinguish them from caps 0x9 alone (no surrounding text, no app identity inside the terminal).

**How to avoid:**
- TERM-0 spike, per target terminal, before any line-correction code: forward Shift+Home into bash with a known full line, read `wl-paste --primary -n`, assert it equals the expected prefix. Record a per-terminal capability verdict (selection / cursor-jump / no-op) as data, not folklore.
- Treat the PRIMARY readback as the *only* ground truth: capture → compare against the daemon's own typed buffer → **abort the line path on any mismatch** (word path remains available). Never derive deletion counts from the buffer alone when the readback disagrees.
- Decide the fallback now: if VTE cannot select, the GNOME Terminal line path either (a) ships word-only with line path documented as wezterm-specific, or (b) uses a different capture primitive (e.g. readline `unix-line-discard` Ctrl+U — but the kill ring is *not* externally readable, so this only works combined with an immediate re-yank/paste strategy — a different design that needs its own spike). Do not let requirements freeze "Shift+Home" as the universal mechanism before the spike verdict.
- Note the GTK3 vs GTK4 asymmetry already on the ledger: GNOME Terminal is GTK3/VTE — IBus-forwarded events reach it (the WINDOWS #4 waiver is GTK4-Wayland only); wezterm's daemon channel was proven live 2026-10-08. The open question is *what the key does*, not whether it arrives.

**Warning signs:**
- Spike: `wl-paste --primary -n` output unchanged after forwarding Shift+Home, or contains text selected hours ago in another app.
- e2e: line-correction "works" only when the user happens to have just selected text manually.

**Phase to address:**
TERM-0 (spike, kill-criterion: if no per-terminal selection primitive exists, line path scope shrinks to wezterm or redesigns capture) — before requirements freeze the mechanism.

---

### Pitfall 2: Backspace-replay in a pty has three different erase semantics; Cyrillic multi-byte is where they diverge (IUTF8)

**What goes wrong:**
N forwarded BackSpaces mean different things depending on what the foreground program is:
1. **bash/zsh with readline/zle** — tty in raw mode; the *program* erases, rune-correct: one BackSpace = one Cyrillic character. (Correct target.)
2. **Canonical-mode programs** (`cat`, `read`, python `input()`, some prompts) — the kernel line discipline erases, and it is **byte-based unless the IUTF8 termios flag is set**: one Cyrillic char (2 bytes) then needs 2 BackSpaces; with a rune-count of N the daemon under-deletes, leaving half-characters and mojibake (documented symptom class: "erases more than was typed", UTF-8 decode errors — SlackWiki, LKML 2004, Manjaro/FreeBSD reports).
3. **Raw-mode TUIs** (vim, htop, less) — BackSpace is an application *command* (see Pitfall 4), deletion semantics are whatever the app decided.

The engine has no way to know which of the three is active: same client, caps 0x9, no surrounding text, no cursor info.

**Why it happens:**
"Backspace deletes the last character" is a GUI belief. In the pty world deletion is a negotiated behavior across termios flags, line-discipline code, and application input loops — the daemon sees only that a key event was forwarded successfully.

**How to avoid:**
- Scope the supported surface explicitly: shells with readline/zle are the contract (they are also where correction makes sense); document canonical-mode and TUI programs as out of contract.
- Always close the loop: after replay+commit, verify (where a primitive exists — selection readback, or a sentinel echo in e2e) and surface "verification failed" instead of silent corruption.
- e2e corpus must include Cyrillic wrong-layout words in bash **and** one canonical-mode program, with golden output, so the byte/rune divergence is pinned rather than discovered by the owner.
- Do not try to read termios of the foreground pty from the daemon (it is not the pty owner; root-free constraint); rely on verification instead.

**Warning signs:**
- Correction leaves one stray half-rendered character or `?` glyph, only for Cyrillic text, only in some programs.
- Length math is right in unit tests (runes) but wrong in one terminal integration case (bytes).

**Phase to address:**
TERM-1 (unit corpus with the three-semantics model), TERM-2 (e2e Cyrillic case per target shell). The IUTF8 fact must be in the design doc so "count in runes" is stated as a readline-only assumption.

---

### Pitfall 3: Cursor mid-line makes count-based replay silently corrupt the line

**What goes wrong:**
BackSpace deletes left of the *current* cursor. The daemon's count derives from its typed buffer, which assumes the caret sits at end-of-line (the common case). If the user pressed Left/arrowed/half-edited before hitting the correction hotkey, N BackSpaces delete the wrong span, and commit_text then inserts the corrected text at the wrong place — the surviving right part of the line ends up spliced into the middle of the correction. Silent, plausible-looking corruption of a command line the user is about to press Enter on.

**Why it happens:**
caps 0x9 means no cursor-position reports from terminals (no set_cursor_location truth), and GUI selection semantics (select-then-replace) don't exist in the pty — the design converts a positional edit into a counting edit without knowing the position.

**How to avoid:**
- For the line path, normalize the caret before counting: forward End (readline `end-of-line`) so the count assumption holds; record the normalization as a documented side effect.
- Verify-then-act: after Shift+Home, compare the PRIMARY readback against the expected prefix (buffer tail). Selected-length < buffer-length → caret was mid-line → take the defined branch (abort line path, or correct only the selected prefix *with the caret branch handled*), never guess.
- For the word path (primary requirement), restrict the replay to the trailing word and re-verify after commit.
- e2e case: type a line, Left Left, correct — assert the suffix survived intact.

**Warning signs:**
- Integration tests pass with typed-then-correct sequences (caret at EOL) but a manual "arrow back then correct" scramble produces mangled lines.
- User report: "it fixed the word and ate the end of my command."

**Phase to address:**
TERM-1 (design: normalization + mismatch branches), TERM-2 (e2e mid-line case in the matrix).

---

### Pitfall 4: TUI / raw-mode apps — replayed BackSpace is a command, not a deletion; the engine cannot see whether bash or vim is in front

**What goes wrong:**
In vim, htop, less, fzf, and every raw-mode TUI, BackSpace/H-key is an application binding (page up, parent dir, scroll, filter-char). The daemon's key log sees keys typed into vim's *insert* mode and cannot tell that the user then pressed Escape — wait, Escape *is* a key event and already a buffer boundary per spec (good) — but vim's *normal* mode consumes plain letters as commands which still flow through the engine into the buffer, and htop's keys never meant text at all. A correction replayed there mutates application state (moved selection, deleted wrong buffer content, re-ordered UI) and garbles the display; in the worst case a command sequence is replayed into an app where keys *do* things (e.g. `d` in htop = mark for kill).

**Why it happens:**
The IBus engine is below the terminal; app mode is invisible. The spec already bans autocorrect in terminals — but manual correction (double-Shift etc.) is user-triggered, and the user cannot always know their caret is "inside" a TUI state where the replay is destructive.

**How to avoid:**
- Keep correction in terminals strictly manual (spec decision) and buffer-gated: replay only the span the daemon itself recorded as typed *since the last boundary* (Enter/Tab/Escape/focus/reset — all already observable). Escape-as-boundary + focus_out-as-boundary covers vim mode switches and most TUI launches (launched from the shell after Enter).
- Hard length cap and mismatch abort from Pitfall 1/3 readbacks; refuse line path when the daemon's buffer is older than N seconds (stale buffer ≠ current line).
- One deliberate e2e witness per risk class: vim insert-mode correction (should work: erase is rune-correct, replayed text is insertable), vim normal-mode after Escape (must no-op: buffer cleared), htop (word path fired deliberately → document observed behavior, define the honest warning in docs).
- Consider an opt-in terminal allowlist in config (inverse of the blocklist) so cautious users can confine the feature to wezterm+GNOME Terminal only.

**Warning signs:**
- Garbage repaints or unexpected app state after correction in any TUI; e2e htop/vim cases flipping.
- User says "it typed `dd` into vim" — a replay leaked past a boundary that should have cleared the buffer.

**Phase to address:**
TERM-1 (boundary discipline + allowlist option in spec-delta), TERM-2 (vim/htop matrix rows), docs phase (honest limitation table).

---

### Pitfall 5: Focus change mid-sequence converts a correction into cross-app corruption; verify rounds need a strict state machine

**What goes wrong:**
A line correction is a *sequence*: capture keys → read clipboard → compute → K × ForwardKeyEvent(BackSpace) → commit_text. If focus leaves the terminal mid-sequence (Alt-Tab, notification, trackpad accident on the T2), the remaining forwarded BackSpaces and the commit land in the **new** focused input context — deleted characters in, say, a password field or a chat draft. The IBus signal channel *does* deliver FocusOut/reset to the engine (observable), so this is preventable — but only if the sequence is abort-aware instead of a fire-and-forget loop.

**Why it happens:**
Developers write the replay as a tight synchronous loop; D-Bus delivery is async and the engine's context object outlives the focus that motivated it. The known session-drift ledger entry (WINDOWS #13) is the same class of lifecycle churn seen from the other side.

**How to avoid:**
- Token-per-sequence: capture a context token at sequence start; every forwarded key checks "same context, still focused, no reset observed" and aborts the remainder on mismatch. Abort must also cancel pending commit_text.
- Chunk bursts (e.g. ≤8 keys per batch) with a context re-check between batches — bounded worst case even on a missed event.
- Verify rounds for capture (Pitfall 1) get the same treatment: bounded retries (2–3), inter-round context check, total deadline; "give up" is a defined, logged outcome with a tray/status hint — not a partial correction.
- e2e case: start correction, force focus flip mid-burst, assert the second window's text is untouched and the terminal is left in a *defined* state (documented: possibly half-corrected, never cross-contaminated).

**Warning signs:**
- Any log line where commit_text or BackSpace is emitted after a FocusOut for that context.
- e2e cross-window case red, or flaky only under load (racing window).

**Phase to address:**
TERM-1 (state machine in design + unit corpus for abort transitions), TERM-2 (focus-flip e2e row). Do not defer to hardening: this is a data-destruction class.

---

### Pitfall 6: wezterm specifics — IME arming race, forwarded shifted keys can lose Shift, caps 0x9 blindness stacks

**What goes wrong:**
wezterm is the owner's daily terminal and the primary target, and it stacks three known behaviors: (a) typing races IME arming — keys right after focus/engine activation can be dropped or misrouted (owner-verified; two-consecutive-greens oracle already mandated); (b) wezterm's Wayland IM frontend "forward key event" path has a documented modifier hazard — shifted keys registering as **unshifted** (wezterm#4615), so a forwarded Shift+Home may arrive as bare Home (cursor jump, no selection) or be consumed differently; (c) caps 0x9 (preedit+focus only, no surrounding text) removes every verification primitive except the clipboard channel. Any one of them makes the line path flaky; combined, a naive implementation looks "randomly broken".

**Why it happens:**
IME activation is a lazy, racy handshake on the client side; the forwarding path is the least-travelled code in wezterm's IME frontend; and the daemon chose (correctly) the daemon channel over external injectors — which concentrates all verification burden on readback rounds that wezterm's caps make expensive.

**How to avoid:**
- Mandate the two-consecutive-greens rule for *every* wezterm oracle row (already the owner's call for phase-6-style oracles — extend it to TERM matrix).
- Pre-arm on focus_in: as soon as a wezterm context gains focus, do the cheapest activation exchange so the arming race window shrinks before a human-scale hotkey arrives.
- Modifier sanity check in the spike: forward Shift+Home, verify via PRIMARY whether a selection appeared; if the Shift-lost signature repeats (cursor jumps instead), pin a wezterm-specific capture alternative (e.g. wezterm key-table binding that produces a selection, or accept cursor-jump + End-normalize + count semantics) — as a *wezterm profile*, not a global behavior.
- Budget verify rounds: each wl-paste spawn is a fork (~tens of ms); 2–3 rounds max, deadline-capped; cache the seat/selection readiness between corrections.

**Warning signs:**
- First correction after focus change fails, second succeeds (arming race).
- Selection readback empty but the line got cursor-jumped to column 0 (Shift-lost signature).
- Matrix v-term rows that pass once and fail the rerun (single-green trap).

**Phase to address:**
TERM-0 (wezterm profile spike), TERM-2 (double-green wezterm rows in the matrix), release gate (two full green matrix runs — house rule since v1.0).

---

### Pitfall 7: Selection capture weaponizes the clipboard channel — newline suffix, PRIMARY clobber, manager history, and secrets

**What goes wrong:**
The capture path reads (and potentially restores) selections that other parts of the desktop watch and value:
- **wl-paste appends a trailing newline** to text by default — any comparison or length computation against naive `wl-paste` output is off-by-one (must use `-n/--no-newline`).
- **Shift+Home overwrites the user's PRIMARY selection** — their previously selected text (for middle-click paste) is destroyed by the capture; on Wayland you cannot even "clear" PRIMARY except by setting something (and setting requires owning a selection source).
- **Clipboard managers** (Clipboard Indicator/History — 2M+ installs, GPaste, CopyQ) record CLIPBOARD and often PRIMARY changes into persistent history: the wrong-layout fragment *and* the corrected text both land in searchable history.
- **Secrets:** terminal lines can be passwords (sudo prompts, `read -s`). The daemon's buffer already holds those keystrokes (accepted v1.0 risk, never logged); but the new capture step would *actively copy* such a line into PRIMARY — a new, real exfiltration surface. Hidden-input (echo off) is invisible to the engine, so it cannot be cleanly detected.

**Why it happens:**
The clipboard is treated as a private scratch buffer between our own steps, while it is actually a desktop-wide, monitored, secret-bearing channel with two selections (PRIMARY vs CLIPBOARD) and async consumers.

**How to avoid:**
- `-n` on every wl-paste invocation; golden-test the capture parser against newline-suffixed and empty readbacks (empty is a defined state, not an error).
- Policy decision (owner, spec-delta): whether user PRIMARY is restored after capture. If restore is attempted, treat it as best-effort with a documented race (restore may race paste consumers); simplest honest v1.2: capture from PRIMARY without touching CLIPBOARD at all (write-back is backspace+commit — no paste), so CLIPBOARD stays pristine and only PRIMARY is clobbered; document the clobber.
- Secrets: never route captured content through CLIPBOARD; never log captured content (lengths only — extend the v1.0 logging rule to the capture path); note the residual risk ("correction hotkey on a hidden-input line copies it to PRIMARY") in README, and consider clearing PRIMARY after use by setting an empty string selection (verify this is even observable in the spike).
- e2e: assert CLIPBOARD unchanged after a full line correction; run one matrix with a clipboard-manager extension active on the stand to observe history noise (documented, not necessarily fixed).

**Warning signs:**
- One-off corrections in tests (no trailing content) pass; corrections of lines ending exactly at buffer width fail (newline off-by-one).
- User report: "my middle-click paste broke" / a password surfaced in CopyQ history.

**Phase to address:**
TERM-1 (clipboard policy in spec-delta: PRIMARY-only, `-n`, restore-or-document), TERM-2 (CLIPBOARD-unchanged e2e row), docs phase (residual-risk note).

---

### Pitfall 8: Opening terminals for *manual* correction re-opens the gate the role-gate deliberately closed (regression risk)

**What goes wrong:**
Today terminals are hard-blocked («пароли и терминалы — никогда», role-gate fail-closed, D-53 blocklist semantics: unknown identity never blocks). v1.2 must allow **manual** correction in terminals while still forbidding **autocorrect** there. The natural implementation — weakening the role-gate for terminals — can accidentally admit the autocorrect layer (whose firing condition is "role-gate passes AND not blocklisted"), reintroducing exactly the behavior the spec bans, in the least safe surface. Conversely, implementing it as a blocklist entry ("terminals in autocorrect blocklist") collides with D-53's "unknown identity is not a block" and with the fact that terminal identity is exactly the weak/unknown case.

**Why it happens:**
Two gates (manual vs auto) were one gate (manual-only everywhere) for the whole project's life; the split is new state space, and the wrong-layout fix is tempting to express as a blocklist hack.

**How to avoid:**
- Spec-delta first (D-55 process): explicit two-permission model — `manual_correction(terminals)=on` (new) vs `autocorrect(terminals)=never` (invariant, not config). The autocorrect arming condition gains a terminal-role veto that is *not* expressible via apps_blocklist.
- Matrix rows proving both sides: autocorrect-in-terminal must stay silent (negative rows, carried from v4 matrix), manual word/line correction must fire (new rows). The negative rows are the regression insurance — keep them frozen.
- Decide identity-free terminal detection (role-based, as today) so the veto does not depend on the flaky terminal app-identity story (Pitfall 9).

**Warning signs:**
- Any code path where the autocorrect layer consults the *manual* gate; any spec text where the terminal veto lives in apps_blocklist.
- A matrix autocorrect row firing inside a terminal after the TERM phase lands.

**Phase to address:**
TERM-0/1 (spec-delta with the two-permission model before code), TERM-2 (frozen negative rows in the matrix).

---

### Pitfall 9: Menu identity via the AT-SPI bridge — namespace unknown, generic, recycled, or absent exactly for the Electron apps the item targets

**What goes wrong:**
AT-SPI application identity is the a11y bus's **unique D-Bus name** (`:1.NN` — recycled across app restarts) plus `org.a11y.atspi.Application` attributes the app *itself* sets. For Chromium/Electron, the a11y tree (and often a meaningful name) appears **only after accessibility is enabled** — so the tray item "Добавить ИмяПриложения в чёрный список" shows "unknown"/generic for precisely ZCode-class apps that most need blocking, until the A11Y feature lands. Additional failure modes: stale identity (identity snapshotted for window A, menu clicked for window B), and the menu click itself moving focus to gnome-shell (the "current app" at click time is the shell, not the target).

**Why it happens:**
Wayland hides app identity from the engine; AT-SPI is the only witness, and its notion of "application" is the toolkit bridge's registration, not the user's notion of "the app I'm typing in". The observer must bridge those notions and keep the snapshot protocol airtight.

**How to avoid:**
- Snapshot identity **at menu-open** (when the user's mental "current app" is defined), not at click time; freeze it in the menu item label («Добавить *wezterm* …») so the user sees exactly what will be added.
- Store all identity forms observed for the focused window (Application name attribute, package/toolkit attributes, bus unique name as last resort) and match the blocklist against all of them (matching is substring RE2, case-sensitive — keep generated patterns consistent with that semantics, Pitfall 10).
- When identity is unknown/anonymous: disable the menu item with an explanatory label rather than adding a junk pattern — honest UX beats a blocklist entry that silently never matches.
- Sequence dependency: A11Y enabling (Pitfall 12) is what makes Electron identity *exist* — order the phases so the A11Y spike verdicts feed the MENU identity design; a cheap A11Y spike early de-risks MENU.

**Warning signs:**
- Menu item reads «Добавить (неизвестное приложение)» in ZCode; blocklist gains entries like `:1.93` that match nothing after the next restart.
- Adding an app from the menu while a different window is focused (racer in manual testing).

**Phase to address:**
MENU-0 (snapshot protocol + identity-forms design; consumes A11Y spike output), MENU-1 (implementation), A11Y-0 spike runs before MENU-0 (cheap, parallel-friendly).

---

### Pitfall 10: Regex injection from app names into the RE2 blocklist — and the strict decoder turns a bad write into a config rejection

**What goes wrong:**
The menu assembles a blocklist pattern from an AT-SPI app name and appends it to `autocorrect.apps_blocklist`. App names are attacker-ish strings full of regex metacharacters: `Visual Studio Code` (fine), `C++ IDE (beta)` (broken RE2 or unintended group), `Node.js` (dot matches everything → `NodeXjs` also blocked). The project's own config contract makes this worse: a broken regex in `apps_blocklist` makes the strict decoder reject the *whole config with the key named* — one bad menu click and the daemon is refusing its config. Also: duplicates on repeated clicks, and the 64-pattern ceiling with no user feedback.

**Why it happens:**
String interpolation into pattern-land without QuoteMeta; the happy path (simple names) never shows the failure, and the failure surfaces far from the cause (config reload error, not the menu).

**How to avoid:**
- `regexp.QuoteMeta(appName)` unconditionally — user-typed patterns in the YAML keep hand-written regex power, but *menu-generated* patterns are always literal, anchored only if that matches the documented substring semantics (plain QuoteMeta substring is consistent with «chromium находит org.chromium.Chromium»).
- Dedupe before write; check the 64-cap and surface a tray notification on refusal; validate the assembled YAML by round-trip decode *before* the atomic rename (never write an unvalidated config).
- e2e/unit: QuoteMeta corpus with hostile names (`(`, `*`, `.`, `+`, Cyrillic names, empty string, 300-char name) asserting both compile and match behavior.

**Warning signs:**
- `goswitchctl status`/reload reporting apps_blocklist decode errors right after menu use.
- Blocklist entry `Node.js` blocking unrelated identities in the matching unit tests.

**Phase to address:**
MENU-1 (QuoteMeta + dedupe + cap + pre-write validation as one task; unit corpus same task).

---

### Pitfall 11: Daemon-driven config mutation — yaml.v3 round-trip destroys comments; self-write reload loop; read-modify-write races

**What goes wrong:**
The menu item must *edit* the user's YAML. The obvious implementation (unmarshal to struct, append, marshal back) **destroys every comment and reorders keys** — and this project's config is comment-rich by contract (CONFIG.md ↔ SKILL.md golden-sync, owner-authored annotations). Secondary traps: the daemon's own write is picked up by the 2 s stat-poll hot reload (fine — intended) but a naive "reload → rewrite normalized config" creates a write/reload echo loop; and a user editing the file in an editor concurrently with a menu click loses one side's changes (read-modify-write race).

**Why it happens:**
yaml.v3 preserves comments only through the `yaml.Node` API (HeadComment/LineComment/FootComment); struct marshal is lossy. The hot-reload poller cannot distinguish "user saved" from "I saved" without a fingerprint.

**How to avoid:**
- Node-tree surgical edit: decode into `yaml.Node`, locate/extend the `autocorrect.apps_blocklist` sequence, encode — nothing else in the file is touched. If the sequence/section is absent, create it at a defined position with the canonical comment from CONFIG.md.
- Atomic write (temp file + rename, same directory), then record a post-write fingerprint (content hash or mtime+size) that the reload poller treats as "self-write, expected" — no rewrite in response, no loop.
- Single-writer discipline: menu, `goswitchctl`, and any future writer share one mutation code path with a lock; document that hand edits concurrent with menu clicks are last-writer-wins (acceptable, stated).

**Warning signs:**
- `git diff` of the config after one menu click shows hundreds of deleted lines (comments gone, keys reordered).
- Daemon CPU spin or repeated reload log lines after a menu write (echo loop).
- User's manual edit reverted by a menu click moments later.

**Phase to address:**
MENU-1 (Node-edit + atomic write + fingerprint as one task; golden test: click → diff contains *only* the new blocklist line), docs phase (CONFIG.md note on daemon-written entries).

---

### Pitfall 12: `org.a11y.Status.IsEnabled=true` is not the signal Chromium/Electron watch — recipes need ScreenReaderEnabled semantics or per-launch flags, and they only apply at app launch

**What goes wrong:**
The daemon's phase-8 a11y magic sets `org.a11y.Status.IsEnabled=true` plus toolkit-accessibility. That suffices for GTK/Qt bridges, but **Chromium/Electron enable renderer accessibility off `ScreenReaderEnabled`** (the property the session ties to screen readers; freedesktop AT-SPI2 docs; corroborating 2026 write-up) — which is why ZCode/Telegram sit outside the a11y tree today "по дизайну". The v1.2 plan says "daemon unchanged — recipes/docs/installer", so the recipes must supply the missing trigger: launch flags (`--force-renderer-accessibility`, the documented Electron switch; Chromium equivalents) or env (`ACCESSIBILITY_ENABLED=1`), or an in-app setting. All of these apply **only at process start**: enabling after the app is up changes nothing until restart — the #1 "I did the recipe, it doesn't work" report.

**Why it happens:**
Two same-named-looking boolean properties (`IsEnabled`, `ScreenReaderEnabled`) with different consumers; Chromium's conservative AT detection (with known bugs, e.g. crbug/1449855) plus Electron's dependence on Chromium's detection.

**How to avoid:**
- A11Y-0 spike matrix before writing any recipe: {Telegram snap, ZCode} × {IsEnabled only (current state), + ScreenReaderEnabled=true, + launch flag, + env} on the real GNOME 46 Wayland session; verify tree visibility via AT-SPI introspection and record the minimal working trigger per app.
- Every recipe ends with "fully quit and relaunch the app" **and** a verification step (a one-liner users can run — e.g. a `busctl`/ATSPI introspection command, or a future `goswitchctl a11y check`) — a recipe without a self-check generates support noise.
- Desktop-entry flags: patch a *copy* in `~/.local/share/applications` (never the vendor file — snap desktop files live in `/var/lib/snapd/desktop` and are overwritten by refreshes); document that app updates may invalidate the override.
- Owner decision point: if the minimal trigger turns out to be `ScreenReaderEnabled=true`, that is a session-wide semantic change (apps may enter screen-reader behaviors — see Telegram's warning bar, Pitfall 13) — decide with eyes open, not as a silent installer side effect.

**Warning signs:**
- Recipe followed exactly, AT-SPI tree still lacks the app → wrong trigger or no restart.
- Works when launched from a terminal with the env var, fails when launched from the dock (env did not reach the launched process).

**Phase to address:**
A11Y-0 (spike matrix; verdict feeds recipes and the MENU identity design), A11Y-1 (recipes/docs/installer-script changes only).

---

### Pitfall 13: Snap confinement and Qt bridge conditions — plus visible side effects (Telegram's "Screen Reader mode" warning bar)

**What goes wrong:**
Telegram is a **snap**: strictly confined builds have a history of AT-SPI integration failures (missing plugs to the a11y bus socket; Ubuntu's own docs note snap sandboxing blocking AT-SPI for Qt apps). The Qt side has its own gate: the AT-SPI bridge activates on a11y detection, with env fallbacks (`QT_ACCESSIBILITY`, `QT_LINUX_ACCESSIBILITY_ALWAYS_BELOW` for older Qt) that must be set **before** launch and must actually reach the snap-launched process (env propagation into snaps is its own trap). And the feature has **in-app side effects**: Telegram Desktop shows a yellow "Screen Reader mode" warning bar whenever it detects an active a11y environment (false-positive complaint, tdesktop issue Apr 2026) — the owner gets a visible banner in exchange for autocorrect. If the snap refuses to cooperate, the honest answer may be "use the deb/flatpak build" — a user-facing recommendation the project has not made before.

**Why it happens:**
Snaps sandbox the very bus sockets a11y needs; Qt's bridge predates snap conventions; and apps increasingly *react* to a11y being on with UI changes meant for real screen-reader users.

**How to avoid:**
- The A11Y-0 spike matrix (Pitfall 12) must run against **the snap build the owner actually has** (`snap list telegram-desktop`; pin the channel in the report), not against upstream docs for deb builds.
- Test env propagation explicitly: set the variable via the mechanism the recipe recommends (environment.d → systemd user manager → .desktop launch) and verify inside the app's environment (`/proc/<pid>/environ` of the snap process) — not just in the launching shell.
- Present the side effects honestly in the recipe: Telegram's warning bar (and any ZCode perf/a11y overhead observed in the spike) stated as known consequences with the owner's acceptance recorded in the spec-delta — this is the same UAT-tail pattern as WINDOWS #5.
- Fallback ladder pre-agreed with the owner: snap+env → snap+desktop-override flags → recommend deb/flatpak variant → (last, and only if recipes fail) revisit the "daemon unchanged" decision per PROJECT.md's own escape hatch.

**Warning signs:**
- Recipe verified on a developer's deb install, fails on the owner's snap (channel/build mismatch).
- Telegram launches but Orca-less a11y tree is empty and no warning bar appears (bridge never came up — confinement or env missed).
- User acceptance surprises: "why is there an orange bar in Telegram?"

**Phase to address:**
A11Y-0 (snap-pinned spike + env-propagation proof), A11Y-1 (recipes with side-effect disclosures + fallback ladder in docs).

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Counting backspaces from the daemon buffer without PRIMARY/other readback verification | Line path works in the demo (caret at EOL, bash) | Silent mid-line corruption; unfixable reports "it ate my command" | Never for the line path; tolerable for word path only with post-commit verification |
| Expressing "no autocorrect in terminals" as a default apps_blocklist entry | No gate changes needed | Collides with D-53 fail-open-for-unknown; terminal identity is the flakiest; one config reset resurrects autocorrect in terminals | Never — invariant belongs in the arming condition, not config data |
| Naive struct→Marshal config edit from the menu | Item works day 1 | User comments destroyed; golden CONFIG.md sync breaks; trust loss | Never — yaml.Node surgical edit is barely more code |
| Forwarded-key sequences as fire-and-forget loops | Fast to write | Cross-app data destruction on focus races; debugging = user reports | Never — token-per-sequence + chunking is cheap |
| Hardcoding Shift+Home as the universal capture key | One code path | VTE verdict (no selection) discovered post-merge; wezterm Shift-lost flake | Only behind a per-terminal capability table from the TERM-0 spike |
| Rewriting app-specific recipes from upstream docs instead of snap-pinned spike | Recipes exist sooner | Owner's snap behaves differently; support noise | Never — spike output is the source of truth |
| Logging captured line content at debug "temporarily" | Easy debugging of capture | Secrets from hidden-input lines in the journal | Never — lengths only, per v1.0 rule |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| VTE / GNOME Terminal | Assuming forwarded Shift+Home creates a selection | No keyboard selection exists (gnome-terminal #588 since 2002); probe per terminal; scope line path by capability verdict |
| wezterm | Assuming forwarded modifiers survive the IM frontend | #4615 shifted-keys hazard + IME arming race; pre-arm on focus_in; double-green oracle; wezterm capture profile |
| pty erase semantics | One BackSpace = one character | readline (rune), canonical n_tty (byte unless IUTF8), raw TUI (app command) — three models; verification over assumption |
| wl-clipboard | `wl-paste` output == selection content | Trailing newline appended (use `-n`); empty selection is a state; `--primary` vs default CLIPBOARD; Wayland primary can't be "unset" trivially |
| Clipboard managers | Treating PRIMARY/CLIPBOARD writes as private | Managers (Clipboard Indicator 2M+, GPaste, CopyQ) record selections; PRIMARY-only policy; secrets never through CLIPBOARD; document history noise |
| AT-SPI identity | Treating app name as stable and always present | Identity = recycled unique bus name + app-set attributes; absent/generic for Electron pre-a11y; snapshot at menu-open; disable item when unknown |
| RE2 blocklist | Interpolating app names into patterns | `regexp.QuoteMeta` for menu-generated patterns; round-trip decode before atomic write (strict decoder rejects whole config on bad regex) |
| yaml.v3 | Struct marshal round-trip for config edits | Comments/order survive only via `yaml.Node`; atomic temp+rename; self-write fingerprint for the 2 s poll reload |
| Chromium/Electron a11y | Setting IsEnabled and expecting Electron trees | ScreenReaderEnabled / `--force-renderer-accessibility` / env; applies at launch only; recipes end with restart + self-check |
| Telegram snap | Deb-docs recipes; env assumed to reach the snap | Spike on the pinned snap channel; verify `/proc/<pid>/environ`; expect the Screen Reader warning bar; fallback ladder to deb/flatpak |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Verify rounds spawning `wl-paste` per attempt (fork each time) | Correction feels sluggish; p95 spikes on loaded sessions | Cap rounds (≤3), deadline the whole capture phase, reuse warmed processes where legal; measure p95 in e2e | Every line correction; worst on first-after-focus (arming + cold fork) |
| BackSpace burst as one D-Bus message per key with no batching | Long lines = long bursts; focus race window grows with burst length | Chunk bursts (≤8) with context re-check; the check costs microseconds | Lines > ~40 chars; Alt-Tab during correction |
| AT-SPI identity observer implemented as polling | Idle CPU burn, battery hit on a laptop daemon (memory budget is fine, CPU duty cycle is not) | Event-driven: subscribe to a11y focus/window events via godbus on the a11y bus; poll only as fallback with backoff | Immediately on every session; noticeable in battery stats |
| Menu-open identity resolution doing a synchronous a11y tree walk | Tray menu opens visibly late | Snapshot cached from the observer's continuous state; tree walk only on cache miss, async | Every menu open in Electron-heavy sessions |
| Re-decoding YAML on every menu action | Config churn latency; reload storms with the 2 s poller | One mutation path; fingerprinted self-write; debounced reload (already exists) | Rapid repeated menu clicks |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Captured line content (possible password from `sudo`/`read -s`) written to CLIPBOARD or logs | Secret exfiltration via clipboard managers/history | PRIMARY-only capture; CLIPBOARD untouched; log lengths only; document residual risk of manual correction on hidden-input lines |
| Menu-generated blocklist patterns from unquoted app names | Over-matching blocks correction in unrelated apps (availability attack on the tool itself); broken regex bricks config load | QuoteMeta + pre-write round-trip validation + refusal-with-notification at the 64 cap |
| A11Y recipes advising global ScreenReaderEnabled=true without side-effect audit | Session-wide screen-reader behaviors activate (app UI changes, perf overhead in Chromium apps) | Spike records per-app minimal trigger; owner accepts side effects in spec-delta; never bundle into silent installer default without the A11Y-0 verdicts |
| Config mutation path writable by anything except the single owner code path | Concurrent writers corrupt YAML (half-written file picked up by strict decoder) | Single-writer module + file lock + atomic rename; validate-before-rename |
| Desktop-entry flag overrides left stale after app updates | Flags silently dropped (a11y "mysteriously" off) or pointing at changed binaries | Recipes include a re-verify step; document refresh behavior of snap desktop files |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Menu item «Добавить (неизвестное приложение)» enabled and adding junk patterns | Blocklist pollution; user believes blocking works | Disable with explanatory label when identity unknown; show the exact name to be added in the label |
| No feedback after "add to blocklist" click | User unsure it worked; clicks repeatedly (duplicates) | Tray notification «wezterm добавлен в чёрный список»; dedupe silently |
| Correction in terminal moves the caret (End-normalization) as a side effect | Surprise for users who correct mid-edit | Document; consider restoring caret position after commit where the primitive exists; word path avoids normalization |
| Telegram's Screen Reader warning bar appears after following the recipe | "Your recipe broke Telegram" | Disclose in the recipe with a screenshot; state it is the cost of autocorrect there |
| Line correction silently aborts (mismatch/focus) with no signal | "Feature is broken" reports | Defined abort outcome: brief tray/status hint + journal entry; honest failure beats partial corruption |

## "Looks Done But Isn't" Checklist

- [ ] **TERM capability table:** per-terminal verdicts (selection / cursor-jump / no-op for Shift+Home) recorded from the live spike — not assumed — for wezterm AND GNOME Terminal
- [ ] **wezterm double-green:** every wezterm oracle row green twice in one matrix run (arming race), plus the release-gate double full-matrix run
- [ ] **Cyrillic replay:** wrong-layout Cyrillic word corrected in bash and zsh; no mojibake (readline rune semantics verified, IUTF8 assumption documented)
- [ ] **Mid-line case:** type → Left Left → correct line → suffix intact (e2e row exists and is green)
- [ ] **Focus-flip case:** correction aborted mid-burst leaves the second window untouched (e2e row exists and is green)
- [ ] **TUI witnesses:** vim insert (works), post-Escape vim (no-op via boundary), htop (defined, documented behavior) — all three in the matrix
- [ ] **Clipboard hygiene:** CLIPBOARD byte-identical before/after a full line correction; `wl-paste -n` used everywhere; PRIMARY policy implemented as decided
- [ ] **Autocorrect veto:** frozen negative rows (autocorrect never fires in terminals) still green after the gate change; manual rows fire
- [ ] **Menu identity:** item label shows the concrete app name; disabled for unknown identity; click during a focus change adds the *snapshot*, not the shell
- [ ] **Pattern safety:** hostile-name corpus (metacharacters, Cyrillic, 300 chars) passes compile+match; config round-trip validates before write; 64-cap refuses with a notification
- [ ] **Config integrity:** after a menu click, `git diff` of the config shows only the new blocklist line (comments/key order intact); no reload echo loop
- [ ] **A11Y spike verdicts:** per-app minimal trigger recorded for the *snap-pinned* Telegram and for ZCode; env verified inside `/proc/<pid>/environ`; recipes include restart + self-check; side effects (warning bar) accepted in spec-delta
- [ ] **e2e stand self-activation:** terminal cases activate their own window before the case (house rule) — terminal focus is exactly where the mutter focus-stealing waiver bites

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Wrong-length replay corrupted a shell line (2, 3) | LOW | User: Ctrl+C fresh line (readline: Ctrl+_ undo exists); daemon side: tighten verification branch, add the missed case to the corpus |
| Replay leaked into a TUI (4) | LOW–MEDIUM | vim `u` / reopen file; add allowlist/boundary fix; add the TUI witness row that was missing |
| Cross-app deletion after focus race (5) | MEDIUM | Affected app undo where it exists; make abort-on-focus tokenized and chunked; add focus-flip e2e row before shipping further TERM work |
| Capture read stale PRIMARY → wrong correction (1, 7) | MEDIUM | Abort path already bounded the damage (verify-before-replay); add mismatch signature to logs; capability table updated |
| Bad menu pattern rejected by config decoder (10) | LOW | Daemon keeps last-good config (existing contract); remove the line via goswitchctl/menu; QuoteMeta fix + hostile corpus lands same task |
| Comments destroyed by config edit (11) | LOW–MEDIUM | Restore from git/backup; rewrite mutation via yaml.Node; golden diff test added before re-enabling the menu item |
| Recipe didn't enable a11y (12, 13) | LOW | Run the recipe's self-check; walk the fallback ladder; if snap-confined, switch channel or deb/flatpak per the pre-agreed ladder |
| Autocorrect fired in terminal (8) | LOW | Turn autocorrect off globally (menu), fix the veto (it must not live in apps_blocklist), unfreeze negative rows |

## Pitfall-to-Phase Mapping

Recommended v1.2 sequencing implied below: **TERM-0 spike → TERM-1/2 (implementation+matrix) with A11Y-0 spike run early in parallel (cheap, de-risks both A11Y recipes and MENU identity) → MENU-0/1 → A11Y-1 (recipes/docs) → milestone gates.**

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| 1. Shift+Home selection assumptions | TERM-0 spike | Per-terminal capability table committed; line-path scope matches verdicts |
| 2. Erase semantics / IUTF8 | TERM-1 design, TERM-2 | Cyrillic bash+zsh e2e rows green; canonical-mode row documented |
| 3. Mid-line cursor | TERM-1, TERM-2 | Mid-line e2e row green; normalization documented |
| 4. TUI replay | TERM-1, TERM-2, docs | vim/htop witness rows; limitation table in README |
| 5. Focus race / verify rounds | TERM-1, TERM-2 | Focus-flip e2e row green; token-abort unit corpus |
| 6. wezterm arming/modifier/caps stack | TERM-0, TERM-2, release gate | Double-green wezterm rows; two full green matrices |
| 7. Clipboard hazards | TERM-1, TERM-2, docs | CLIPBOARD-unchanged e2e row; `-n` audit; policy in spec-delta |
| 8. Terminal autocorrect veto | TERM-0 spec-delta, TERM-2 | Frozen negative rows green post-change; veto not in blocklist |
| 9. AT-SPI identity | A11Y-0 (early), MENU-0/1 | Snapshot protocol tests; unknown-identity UX; Electron identity visible post-A11Y |
| 10. Regex injection / config brick | MENU-1 | Hostile-name corpus green; pre-write validation unit-tested |
| 11. Config mutation integrity | MENU-1 | Golden diff test (only the new line changes); no echo loop in soak |
| 12. Electron a11y trigger | A11Y-0, A11Y-1 | Spike verdicts recorded; recipe self-check passes on ZCode |
| 13. Snap confinement / side effects | A11Y-0, A11Y-1 | Snap-pinned spike report; env-in-`/proc` proof; side effects accepted in spec-delta |

## Sources

- GNOME GitLab gnome-terminal#588 — keyboard selection open since 2002; VTE#648 (Shift+arrows request); SuperUser "why can't I highlight with keyboard in terminal" — VTE passes Shift+Home through to the app — **MEDIUM** (multi-source, matches project's live observations)
- wezterm: [use_ime docs](https://wezterm.org/config/lua/config/use_ime.html), [discussion #2929 — IME intermittently stops](https://github.com/wezterm/wezterm/discussions/2929), [#4615 — Wayland IM frontend forward-key-event, shifted keys register unshifted](https://github.com/wezterm/wezterm/issues/4615), [#1772 — text-input-v3 open](https://github.com/wezterm/wezterm/issues/1772) — **LOW→MEDIUM** (single issues; corroborated by the owner's live spike 2026-10-08 → treat the local evidence as HIGH)
- Kernel/pty erase: [SlackWiki UTF-8 console — IUTF8 symptom](https://www.slackwiki.com), [LKML 2004 — stty utf8 / line editor erase](https://lkml.rescloud.iu.edu), [man7 utf-8(7)](https://man7.org/linux/man-pages/man7/utf-8.7.html), [Manjaro forum — backspace deletes part of a unicode char](https://forum.manjaro.org), [kbd-project delete-and-backspace](http://kbd-project.org/docs/howtos) — **MEDIUM** (multi-source)
- Readline raw mode: [CS3214 raw-mode notes — bash readline enters raw mode](https://courses.cs.vt.edu/cs3214/spring2024/questions/rawmode), [kilo tutorial — entering raw mode](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html), [unix.SE 151068 — backspace in raw reads](https://unix.stackexchange.com/questions/151068) — **MEDIUM**
- Bracketed paste & history expansion: Vim Tips Wiki "Toggle auto-indenting for code paste"; [zsh 5.1 bracketed paste — newlines not accept-line](https://archive.zhimingwang.org/blog/2015/09/06/zsh-5.1-and-bracketed-paste/); [gemini-cli #16348 — "!" in pasted text mishandled without bracketed paste](https://github.com/google-gemini/gemini-cli/issues/16348); unix.SE histexpand references — **MEDIUM**
- wl-clipboard: [wl-paste(1)](https://man.archlinux.org/man/wl-paste.1.en) / [Debian manpages](https://manpages.debian.org) — `-n/--no-newline`, `--primary`, primary-settable-only-while-selection-exists — **HIGH** (man pages)
- Clipboard managers: [PrivacyGuides discuss — GNOME keeps no persistent history by default](https://discuss.privacyguides.net), [askubuntu — CopyQ security settings/secret detection](https://askubuntu.com), extensions.gnome.org Clipboard Indicator/History listings; GPaste primary-capture reports — **MEDIUM**
- Chromium/Electron a11y: [freedesktop AT-SPI2 docs — org.a11y.Status ScreenReaderEnabled](https://www.freedesktop.org/wiki/Accessibility/AT-SPI2/), [Chromium accessibility design doc](https://www.chromium.org/developers/design-documents/accessibility/), 2026 a11y.rs write-up (ScreenReaderEnabled premise), [crbug/1449855 — detection bug](https://issues.chromium.org), [electron/electron#11544 — --force-renderer-accessibility for Linux a11y](https://github.com/electron/electron/issues/11544), [Electron command-line switches docs](https://www.electronjs.org/docs/latest/api/command-line-switches), [ArchWiki accessibility](https://wiki.archlinux.org/title/Accessibility) — **MEDIUM** (cross-verified)
- Telegram snap / Qt: [Ubuntu desktop docs — snap sandboxing can block AT-SPI](https://ubuntu.com/desktop/docs), [tdesktop issue — Screen Reader warning bar false positive (Apr 2026)](https://github.com/telegramdesktop/tdesktop/issues), [Qt 6 accessibility docs — QT_ACCESSIBILITY env](https://doc.qt.io/qt-6/accessible.html), tdesktop-accessible project — **MEDIUM** (multi-source; exact snap channel behavior to be pinned by the A11Y-0 spike)
- AT-SPI identity: [at-spi2-core dev docs — org.a11y.atspi.Application/Socket](https://gnome.pages.gitlab.gnome.org/at-spi2-core/), [at-spi-bus-launcher](https://github.com/GNOME/at-spi2-core) — unique bus names assigned by the bus; Socket.Embed registration; Application attributes — **MEDIUM**
- yaml.v3 round-trip: go-yaml#132 (comments only on yaml.Node), goccy/go-yaml docs (key-order preservation) — **MEDIUM**
- Project-local (HIGH — target runtime): .planning/WINDOWS.md (#4 GTK4-Wayland waiver, #5 latency-window, #7/#13 focus & drift), .planning/PROJECT.md (v1.2 feature notes: daemon-channel spike 2026-10-08, wezterm caps 0x9, two-consecutive-greens oracle, a11y.enabled default-ON decision, D-53 blocklist semantics), docs/SPEC.md §4 (a11y mechanism: IsEnabled + toolkit-accessibility, strict decoder rejects broken regex with key name), docs/CONFIG.md (apps_blocklist substring/case/64-cap semantics; strict decode errors)

---
*Pitfalls research for: goswitch v1.2 — terminal correction, tray-menu blocklist item, Telegram/ZCode a11y enabling*
*Researched: 2026-10-08*
