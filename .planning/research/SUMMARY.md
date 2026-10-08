# Project Research Summary

**Project:** goswitch — milestone v1.2 «Терминалы и видимость приложений»
**Domain:** GNOME Wayland IBus input-method daemon (Go) — terminal text correction, dynamic tray-menu blocklist item, per-app accessibility enablement
**Researched:** 2026-10-08
**Confidence:** MEDIUM overall (HIGH for target-runtime live probes and in-repo code-read design; MEDIUM for cross-checked web findings on VTE/wezterm transport; LOW items are single-source upstream issues — all deliberately spike-gated)

## Executive Summary

goswitch v1.2 extends a shipped, live-proven IBus-engine daemon with three features: **TERM** — manual layout correction in terminals (word mandatory, whole line feasibility-gated; autocorrect in terminals is permanently off) plus a fix for the wezterm layout-switch defect in the owner's daily terminal; **MENU** — a dynamic tray item «Добавить ИмяПриложения в чёрный список»; **A11Y** — recipes enabling accessibility inside Telegram and ZCode so autocorrect reaches them, with the daemon itself unchanged (owner decision 2026-10-08). Terminal correction at the daemon level is genuinely novel on Linux — layout_corrector refuses terminals outright, xneur is dormant X11-only, shell plugins are per-shell/per-machine — so TERM is the milestone's differentiator, and Punto Switcher's 20-year select→convert→paste UX is the muscle memory the flow reproduces through the IBus engine channel.

The recommended approach is **spike-first with zero new Go dependencies**. Every mechanism rides existing in-repo packages: `internal/engine` ForwardKeyEvent bursts (the daemon's own injection channel, live-verified race-free on 2026-10-08), `internal/clipboard` wl-clipboard wrapper (adds a `--primary` flag), `internal/indicator` DBusMenu machinery, `internal/appid`, `internal/a11y`, and the atomic-rename config writer. The wezterm defect is live-diagnosed to its root: `enable_wayland = false` in the owner's config makes wezterm an **XWayland/XIM client**, and since ADR-006 makes both GNOME input sources goswitch engines, an engine flip never changes the X keymap the XIM client renders characters from — the flip is structurally invisible. Primary fix candidate is a config recipe (native Wayland, `zwp_text_input_v3`), with a daemon-side XIM workaround as fallback. The terminal **read channel is per-terminal**: VTE has no keyboard selection at all (VTE source-verified), so VTE word correction rides the daemon's own keystroke buffer + `commit_text` (no read channel needed), while wezterm gets a config-recipe selection (CopyMode) read back via `wl-paste --primary`. Write-back is always **N×BackSpace replay + `commit_text`** — never commit-over-selection (terminal selections are visual-only; VTE merely deselects), never a Ctrl+V paste (CLIPBOARD stays untouched; PRIMARY-only policy).

Key risks, in order: (1) per-terminal key semantics are unverified assumptions until the Phase 1 spike runs — VTE passes Shift+Home through to the shell, wezterm has a documented shifted-keys-lost hazard (#4615); (2) the Telegram **snap is structurally AppArmor-blocked** for inbound AT-SPI calls (live-probed: `GetRole` → "Access denied" despite correct env) — no env var or snap interface fixes this, so an owner decision (unconfined build vs daemon-side fallback) is required before that phase can plan; (3) the replay/focus race is a data-destruction class — tokenized abort, chunked bursts, and verify-before-the-destructive-point are design requirements, not hardening; (4) menu identity is only as good as the a11y tree (Electron identity exists only after a11y is enabled; snaps and wezterm are invisible) — the item must degrade to disabled, never guess. Mitigations are structural: D-55 spec-delta + ADR-008 before any code, spike verdict tables, two-consecutive-greens wezterm oracles, frozen autocorrect-negative matrix rows.

## Key Findings

### Recommended Stack

Zero new modules for v1.2 — `go.mod` is unchanged; everything is stdlib + godbus + existing in-repo packages (full detail in STACK.md).

**Core technologies (delta from v1.1):**
- `github.com/godbus/dbus/v5` v5.2.2 — unchanged, still latest; engine service side, DBusMenu, a11y belt, appid.
- In-repo `internal/engine` ForwardKeyEvent bursts — the injection channel for select chords and the BackSpace replay; adds keyval constants (Home/End/arrows) to `keys.go`; the level-2 replay emitter is wire-tested and becomes live in v1.2.
- In-repo `internal/clipboard` + `--primary` — `ReadPrimary`/`SetPrimary`/`ClearPrimary` on the existing Runner (wl-clipboard 2.2.1, dpkg-verified); same empty-vs-error and no-pipes discipline.
- In-repo `internal/indicator` (SNI + DBusMenu v3) + `internal/appid` — dynamic menu item id 16 with interpolated label via the existing `ItemsPropertiesUpdated` deltas.
- In-repo `internal/a11y` (gsettings `toolkit-accessibility`) — ZCode/Electron a11y already works via this key (live-probed); the milestone adds a belt, not daemon code.
- wezterm config recipe (owner's `wezterm.lua`) — SHIFT+Home → CopyMode `Multiple` sequence; version pinned at 20240203 (the **last tagged release ever**; upstream is nightlies-only).
- User-local `~/.local/share/applications/zcode.desktop` override with `--force-renderer-accessibility` — deterministic Electron a11y belt, no root.

**Explicitly NOT:** any new Go module; `QT_LINUX_ACCESSIBILITY_ALWAYS_ON` as primary (perf cost, redundant); `GTK_MODULES=gail:atk-bridge` (GTK2-era); xclip/xsel (X11-only); snap-interface workarounds for Telegram (none exists); OSC-52 tricks; app-side wezterm Lua layout remaps (fights ADR-006 single-source).

### Expected Features

**Must have (table stakes — missing = milestone failed):**
- **TERM-0** — layout flip works in wezterm (defect fix or verified recipe); prerequisite for all TERM work there.
- **TERM-1** — manual word correction at the prompt in wezterm + GNOME Terminal/VTE; autocorrect stays off in terminals (invariant, not config).
- **MENU-1** — dynamic «Добавить ИмяПриложения в чёрный список» with correct hide/disable-when-unknown behavior.
- **A11Y-1** — recipes/docs/installer steps for Telegram + ZCode, verified against live AT-SPI trees; Telegram warning bar documented.

**Should have (stretch):**
- **TERM-2** — whole-line correction with cursor restoration (feasibility-gated by the Phase 1 capability table).
- XWayland identity fallback for MENU (wezterm in the menu) — P3, owner-annoyance-triggered.

**Anti-features (explicitly not building):**
- **Autocorrect in terminals — never.** Hidden-input contexts (sudo/ssh passwords) are indistinguishable at key level; the veto lives in the autocorrect arming condition, never in `apps_blocklist` (D-53 collision).
- Commit-over-selection (does nothing to the pty), fire-and-forget key bursts, per-app wezterm Lua layout switching, un-blacklist menu counterpart, selection preservation and «не трогать» URL/path heuristics (v2+/CORR-10).

### Architecture Approach

Orchestration lives in `internal/session` as a **`terminalOp` stepper** — a time-stepped, off-mutex rung following the `runClipboardRung` discipline (WR-01), because terminals never answer verify rounds (VTE falsely advertises surrounding text, RHBZ 1847347; wezterm caps `0x9`); pure helpers stay in `internal/correct` (`TerminalPayload` gate, select-chord geometry table, `BuildPlan` reused verbatim with caps forced to 0 — the level-2 backspace-replay math). No new packages; the pure/D-Bus boundary and depguard architecture are untouched. The flow has an explicit **safe point**: everything before the first forwarded BackSpace mutates nothing in the field, so abort is free; past it, one-way replay + commit with epoch-guarded per-phase deadlines and a verify-after suffix-compare that logs and stops, never auto-repairs. The gate is identity/role first (appid `RoleTerminal=60`), with verify-timeout fallback escalation (recommended ON for word). MENU is a third Callbacks-driven toggle composition: snapshot identity at menu-open, `config.AppendBlocklistEntry` (yaml.Node round-trip, `regexp.QuoteMeta`, dedupe, cap 64, pre-write validation, atomic rename), synchronous reload + fold.

### Critical Pitfalls (top 5 of 13 — full list in PITFALLS.md)

1. **Selection-capture key assumptions.** VTE has no keyboard selection at all (open request since 2002); wezterm #4615 can drop Shift from forwarded keys (cursor-jump signature). Avoid: per-terminal capability table from the TERM-0 spike; the PRIMARY read-back is the only ground truth; abort on any mismatch with the daemon's buffer.
2. **Focus change mid-burst = cross-app data destruction.** Remaining BackSpaces and the commit can land in the newly focused context. Avoid: token-per-sequence, context re-check between chunks (≤8 keys), abort cancels pending commit; dedicated e2e focus-flip row.
3. **Triple pty erase semantics + Cyrillic IUTF8.** readline erases runes; canonical mode erases bytes unless IUTF8; TUI apps treat BackSpace as a command. Avoid: scope to shells as the contract; buffer-gate the replay span (boundaries: Enter/Tab/Escape/focus); verify after; Cyrillic bash+zsh e2e corpus; vim/htop witness rows.
4. **Telegram snap AppArmor wall + Electron launch-time triggers.** The snap's profile denies inbound AT-SPI regardless of env (live-probed); Chromium/Electron a11y triggers apply only at process launch and key off the right property/flag. Avoid: snap-pinned A11Y-0 spike matrix (per-app minimal trigger, env verified in `/proc/<pid>/environ`); pre-agreed fallback ladder; restart + self-check in every recipe.
5. **Regex injection and config integrity from menu writes.** App names are regex metacharacter bombs; the strict decoder rejects the whole config on a bad pattern; struct-marshal destroys comments. Avoid: unconditional `QuoteMeta` for menu-generated patterns, dedupe, cap 64 with refusal notification, round-trip decode before atomic rename, yaml.Node surgical edit, golden diff test (only the new line changes).

Plus the regression trap: opening terminals for *manual* correction must not re-admit *autocorrect* — the two-permission model (manual=on, auto=never) goes into the spec-delta, with frozen negative matrix rows as insurance.

## Implications for Roadmap

**Resolved cross-file tensions** (the roadmapper should carry these as decisions, not open questions):

1. **Read channel is per-terminal, not universal select-via-keys.** The original "forward selection keys everywhere" design is dead for VTE: no keyboard selection exists (FEATURES/PITFALLS agree; VTE source-verified). Final shape: VTE word = daemon keystroke buffer + `commit_text` (no read channel); VTE line = gated by the **TERM-0 kill criterion** (no selection primitive ⇒ line scope shrinks to wezterm-only or the capture primitive is redesigned — the readline kill ring is not externally readable, so Ctrl+U capture is a different design needing its own spike); wezterm = config-recipe CopyMode selection + `wl-paste --primary`, with the #4615 Shift-lost signature as an explicit spike checkpoint. The per-terminal capability table from Phase 1 is the gating artifact.
2. **Telegram snap a11y is structurally blocked; the owner's "enable accessibility in the apps" decision needs one more input.** STACK live-probed the wall: env is already correct (`QT_ACCESSIBILITY=1` in the process), the app registers on the a11y bus, and AppArmor still denies all inbound AT-SPI calls; no snap interface exists. The A11Y phase must not plan until the owner picks: (a) an unconfined Telegram build — official tarball or flatpak `org.telegram.desktop` (flatpak a11y reachability needs a one-off live check first), or (b) the owner's reserved daemon-side fallback. Ladder: snap → deb/flatpak → daemon fallback; the snap has no rung.
3. **Write-back is backspace-replay + `commit_text`, never commit-over-selection, never a clipboard paste.** Terminal selections are visual-only (VTE deselects on incoming content; the shell never learns), and a forwarded paste chord would need a CLIPBOARD/PRIMARY write of possibly-secret content plus bracketed-paste uncertainty (plain Ctrl+C is SIGINT). FEATURES' anti-feature objection to backspace-replay ("blind N backspaces") is answered by design: the count comes from the verified read-back/buffer span, the safe-point rule makes pre-replay abort free, and verify-after catches mismatch — the replay is never blind. Paste-over remains a parked fallback only if a spike shows `commit_text` not landing in some terminal.
4. **A11Y-0 spike runs early even though A11Y execution is last/parallel.** ARCHITECTURE places the A11Y phase fifth; PITFALLS correctly notes Electron identity (needed by MENU) exists only after a11y is enabled. Resolution: run the cheap A11Y-0 spike matrix during the Phase 1 window; its verdicts feed the Phase 3 identity design and the Telegram decision; Phase 5 execution stays parallelizable.

### Phase 1: Спайк + дефект wezterm (spec-delta, ADR-008, flip fix)
**Rationale:** the owner's daily defect, and the only phase that produces the transport evidence every later terminal phase consumes. D-55 is absolute: the spec-delta (TERM scope — word mandatory, line feasibility-gated, автокоррекция в терминалах off; two-permission model; latency-budget completion clause: reaction <50 ms, completion best-effort with per-phase deadlines) and ADR-008 (terminal transport, Proposed) precede all code; ADR-008 is driven to Accepted by the spike's evidence.
**Delivers:** per-terminal capability verdict table (caps, `ProcessKeyEvent` arrival, `commit_text` landing, select chord → PRIMARY read-back, a11y identity presence); wezterm flip fixed — config recipe `enable_wayland = true` first candidate, upstream build bump if #5125 bites on 20240203, daemon-side XIM workaround as the fallback branch; ADR-008 Accepted; e2e wezterm rows (two-consecutive-greens oracle); the TERM-2 kill-criterion verdict; the A11Y-0 spike matrix run (see tension 4).
**Addresses:** TERM-0; unblocks TERM-1/TERM-2; feeds MENU identity design and the Telegram decision.
**Avoids:** Pitfalls 1, 6, 8 (mechanism frozen only after the spike; two-permission model in spec first); Anti-Pattern 4 (no fix designed from the issue tracker).

### Phase 2: TERM-слово (обязательный)
**Rationale:** the MVP core; the select→read→rewrite machinery is built once here and reused by the line phase with a chord swap.
**Delivers:** `clipboard.ReadPrimary` + primary save/restore (PRIMARY-only policy: CLIPBOARD byte-untouched, `-n` everywhere, clobber documented); pure terminal helpers in `correct`; `terminalOp` stepper + failure vocabulary (`terminal-busy`, `terminal-read-empty`, `terminal-refused-*`, `terminal-interrupted`); gate (identity/role first, verify-timeout fallback for word); config knobs (`correction.terminal`, per-phase timeouts, settle delay); e2e matrix — wezterm + VTE word, Cyrillic bash+zsh, mid-line, focus-flip, vim-insert/vim-normal/htop witnesses, frozen autocorrect-negative rows, CLIPBOARD-unchanged row.
**Uses:** ForwardKeyEvent bursts (`forwardShiftCtrlLeft`, BackSpace replay), `BuildPlan(caps=0)` level-2 math (first live client), `wl-paste --primary`, MACR clean-modifier rule.
**Avoids:** Pitfalls 1–8.

### Phase 3: MENU — пункт чёрного списка (independent; may swap with Phase 2)
**Rationale:** small, fully precedented in-repo, independent of terminal evidence; consumes the early A11Y-0 verdicts (Electron identity exists only post-a11y — sequencing matters for testing, not for code).
**Delivers:** `config.AppendBlocklistEntry` (QuoteMeta, dedupe, cap 64, Node round-trip, pre-write validation, atomic rename with dot-prefix temp); indicator item 16 + `Callbacks.BlocklistTarget`/`AddToBlocklist`; `MenuSync.SetBlocklistTarget`; appid start-on-menu-attach; snapshot-at-menu-open protocol (label shows the exact name; disabled-with-explanation when unknown; never a junk `:1.NN` pattern); tray notification on success/refusal; e2e via DBusMenu Event; golden diff test (click changes only the new blocklist line).
**Avoids:** Pitfalls 9, 10, 11.

### Phase 4: TERM-строка (если возможно — feasibility-gated)
**Rationale:** enhances the proven word rung (`End` + `forwardShiftHome` instead of Ctrl+Shift+Left; cursor-restore replay `ArrowLeft` × len(right side)); explicitly gated by the Phase 1 capability table — the VTE kill criterion can shrink scope to wezterm-only or force a capture redesign; the owner may scope it out at the phase gate without failing the milestone.
**Delivers:** line select + read on the surviving scope; End-normalization with documented side effect; verify-then-act mismatch branches (selected-length < buffer-length ⇒ defined branch, never guess); e2e line rows for the surviving scope, including the mid-line suffix-intact case.
**Avoids:** Pitfalls 1, 2, 3.

### Phase 5: A11Y — рецепты Telegram/ZCode (parallelizable with 2–4)
**Rationale:** daemon unchanged (owner decision); recipes/docs/installer-side only; blocked only on the A11Y-0 spike verdicts and the Telegram build decision (tension 2). Includes the check-in gate «если Telegram не поддастся — вернуться к вопросу настройки демона».
**Delivers:** ZCode user-local `.desktop` override (`--force-renderer-accessibility`) + "restart the app" docs + a user-runnable self-check one-liner; Telegram per the owner's ladder verdict (unconfined tarball/flatpak with a one-off flatpak a11y reachability check, or the daemon-side fallback); the yellow «Screen Reader mode» bar disclosed as a known consequence with owner acceptance recorded in the spec-delta; README residual-risk note (manual correction on hidden-input lines copies to PRIMARY).
**Avoids:** Pitfalls 12, 13.

### Phase Ordering Rationale

- **Spec-delta + ADR-008 before code** (D-55 absolute); the spike is Phase 1 because it retires the most uncertainty per unit of work and is the evidence source ADR-008 needs to reach Accepted.
- **Word → line:** same machinery, chord swap; line additionally gated by the kill criterion, so it must follow both the spike and the working word rung.
- **MENU is independent** and fully precedented — it slots after TERM-word or swaps with it if terminal evidence slips; it needs only the early A11Y-0 spike output for its identity-degradation design.
- **A11Y execution is last/parallel** because it is daemon-untouched; its real dependencies (spike verdicts, owner decision) are cheap and run early.
- The ordering keeps the two data-destruction surfaces (terminal replay, config mutation) each behind their gating artifact: the capability table for TERM, spike verdicts + QuoteMeta/Node discipline for MENU.

### Research Flags

Phases likely needing deeper research during planning (`/gsd:plan-phase --research-phase`):
- **Phase 2 (TERM-слово):** the most novel surface (time-stepped stepper, per-terminal quirks, gate escalation). Architecture research is already unusually detailed (code-read HIGH) — run research-phase only if Phase 1 verdicts introduce surprises (e.g., `commit_text` not landing in VTE, settle-delay behavior).
- **Phase 4 (TERM-строка):** plan only after Phase 2 is green and the kill criterion is known; if any VTE-scope line path survives, its capture primitive needs its own spike before planning.
- **Phase 5 (A11Y):** planning is blocked on the A11Y-0 spike matrix results and the Telegram owner decision; the recipe work itself is docs-level once verdicts exist.

Phases with standard patterns (skip research-phase):
- **Phase 1:** it *is* the research — a live spike by design; web sources on wezterm transport are LOW confidence and should not be researched further, just probed. Planning should encode the kill criteria and the verdict-table format.
- **Phase 3 (MENU):** every precedent is in-repo and live-proven (`configToggle.flip`, `writer.go` atomic rename, `menu.go` item ids/AboutToShow, appid) — HIGH-confidence code-read; standard patterns apply.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Zero new dependencies; every component live-verified on the target (dpkg, gsettings, a11y probes, owner configs) and in-repo code is wire-tested and production-driven since v1.0/v1.1 |
| Features | MEDIUM–HIGH | Table stakes are owner-defined and unambiguous; TERM-2 feasibility and Telegram autocorrect usefulness are empirical, spike-gated |
| Architecture | HIGH in-repo / MEDIUM transport | Component design read directly from shipped code (live-proven v1.0.0); VTE/wezterm wire behavior cross-checked but not yet live-spiked — exactly what Phase 1 retires |
| Pitfalls | MEDIUM | Project-ledger and live-probe facts HIGH; web findings LOW→MEDIUM (multi-source cross-checked); residual LOWs are all assigned to spike gates |

**Overall confidence:** MEDIUM — high wherever the target runtime and the codebase are the source; deliberately unproven wherever only a live spike can decide (wezterm transport, VTE select behavior, Telegram confinement workarounds, minimal a11y triggers).

### Gaps to Address

- **wezterm flip root cause/fix:** FEATURES' live-verified XIM explanation (`enable_wayland=false`) vs ARCHITECTURE's h1–h3 bridge-race hypotheses — the Phase 1 verdict table reconciles them; LOW confidence until run. Note the coupling: if wezterm moves to native Wayland for the flip fix, forwarded select chords travel the Wayland IM frontend where #4615 (Shift-lost) lives — the spike must test both under the new transport.
- **VTE line capture:** the kill criterion may scope the line path to wezterm-only; Ctrl+U kill-ring capture is not externally readable and would be a different design needing its own spike.
- **Telegram build decision (owner, blocking Phase 5 planning):** unconfined tarball vs flatpak (a11y reachability unverified) vs daemon-side fallback; no snap interface exists.
- **Minimal a11y trigger per app:** IsEnabled vs ScreenReaderEnabled vs launch flag vs env — the A11Y-0 matrix decides; if `ScreenReaderEnabled=true` is the trigger, that is a session-wide semantic change requiring explicit owner acceptance. All triggers apply only at launch.
- **wezterm word select chord:** whether Ctrl+Shift+Left selects by default in wezterm was not established in research (only Shift+Home absence was) — the capability table pins the word chord too.
- **PRIMARY settle timing:** selection update after injected keys is asynchronous; settle-delay knob spike-pinned (start ~50–100 ms), per-phase deadlines as wedge guards.
- **Buffer+commit in VTE:** "may already be close to working" is untested — a cheap Phase 1 spike check that could simplify the VTE word path.
- **Telegram a11y quality:** even unconfined, upstream a11y is poor; verify the input field is actually exposed in the tree before promising autocorrect there.

## Sources

### Primary (HIGH confidence)
- Live target-system probes, 2026-10-08 (all four research files): owner `~/.wezterm.lua` (`enable_wayland=false`), wezterm 20240203 + WM_CLASS on XWayland, gsettings input-sources (both sources goswitch engines), dpkg versions (wl-clipboard 2.2.1, VTE 0.76.0, gnome-terminal 3.52.0), telegram-desktop snap 7.2.9 process environ, a11y-bus AppArmor denial probes (`GetRole` denied on telegram/chromium snaps; 'zcode' visible), daemon-channel race-free spike
- Codebase (read directly; live-proven in v1.0.0): `internal/session/actor.go` (WR-01 rungs, forwardCtrlV, MACR lesson, flipCredit), `internal/clipboard/clipboard.go` (Runner, empty-vs-error, no-pipes), `internal/config/writer.go` (atomic rename, Node round-trip), `internal/indicator/menu.go`, `internal/appid/appid.go` (RoleTerminal=60), `internal/correct/plan.go` (BuildPlan level-2), `cmd/goswitchd/main.go`
- VTE 0.76.0 source read (GitLab tag): `src/vte.cc` widget_key_press (no keyboard selection), deselect-only incoming-contents, `src/keymap.cc` modifier encoding
- Project ledgers: `.planning/WINDOWS.md`, `.planning/PROJECT.md` (spike record, caps 0x9, D-53, two-greens oracle), `docs/SPEC.md`, `docs/CONFIG.md`
- wl-clipboard man pages (`--primary`, `-n`) + local 2.2.1; `/usr/include/ibus-1.0/` headers

### Secondary (MEDIUM confidence)
- RHBZ 1847347 / GNOME vte#214 — VTE surrounding-text false advertising
- wezterm: #1772 (text-input-v3), #5125 (IME broken with IBus on exactly 20240203, fixed in nightly), #4615 (forwarded shifted keys register unshifted), default-keys docs, releases (20240203 = last tagged), #6341, #4358
- gnome-terminal#588 — keyboard selection requested since 2002
- Chromium/Electron/Qt accessibility docs — startup detection, `--force-renderer-accessibility`, `app.setAccessibilitySupportEnabled`, `QT_ACCESSIBILITY`; freedesktop AT-SPI2 ScreenReaderEnabled docs; crbug/1449855
- pty erase semantics / IUTF8 — SlackWiki, LKML, man7 utf-8(7), Manjaro reports; raw-mode/readline sources; bracketed paste references
- Clipboard managers (Clipboard Indicator, GPaste, CopyQ) recording PRIMARY/CLIPBOARD
- go-yaml#132 — comments survive only via yaml.Node
- Telegram: tdesktop Screen Reader warning-bar issue (Apr 2026), Ubuntu docs on snap AT-SPI sandboxing, Zendalona fork
- Competitors: layout_corrector README («terminal not supported», archived 2022), xneur (X11-only, dormant), Punto Switcher Shift+Break/Alt+Break mechanics

### Tertiary (LOW confidence — all spike-gated)
- GNOME GitLab «IME in wezterm dies after screen unlock» — same family as the defect; live-confirm in Phase 1
- ghostty#2392 ForwardKeyEvent segfault — motivates mandatory per-terminal verification
- Snapcraft forum AT-SPI-in-snaps threads — superseded for the verdict by the live AppArmor probe
- AskUbuntu gsettings lore — superseded by local `gsettings writable` verification

---
*Research completed: 2026-10-08*
*Ready for roadmap: yes*
