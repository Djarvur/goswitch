# Phase 9: Спайк и дефект wezterm — Research

**Researched:** 2026-10-08
**Domain:** live terminal-transport spike (wezterm flip defect + terminal capability table) + spec-delta/ADR-008 docs, on a shipped IBus-engine daemon (GNOME 46 Wayland, Ubuntu 24.04, T2 MacBook)
**Confidence:** MEDIUM overall — HIGH for everything read from the repo or probed live this session; the wezterm root-cause and all capability-table verdicts are **deliberately unproven until the spike runs** (this phase *is* the research — the committed milestone research explicitly says web transport sources must be probed, not re-researched)

## Summary

Phase 9 is an **evidence phase with one embedded defect fix**. Five artifacts must exist when it closes: (1) layout switching working in wezterm on the live desktop, proven by two consecutive green e2e oracles; (2) a dated SPEC spec-delta (D-55 — before any product code) carrying the two-permission model and the TERM-04 invariant; (3) ADR-008 (terminal transport) at **Accepted**, citing live spike evidence, not web sources; (4) the terminal capability verdict table that gates Phase 10 (word chord) and Phase 12 (line kill-criterion); (5) the A11Y probe matrix verdicts consumed by Phase 11 (MENU identity degradation) and Phase 13 (Telegram decision). Zero new Go modules; all spike tooling is already installed on the target (verified live this session).

The wezterm defect has a live-verified structural picture from the milestone research: the owner's `~/.wezterm.lua` sets `enable_wayland = false` [VERIFIED: ~/.wezterm.lua:77-80 — `local config = { enable_wayland = false, keys = keys, }`], making wezterm 20240203-110809-5046fc22 [VERIFIED: `wezterm --version`, this session] an XWayland/XIM client; under ADR-006 both GNOME sources are goswitch engines (`sources = [('ibus', 'goswitch-en'), ('ibus', 'goswitch-ru')]`, live-probed this session), so a flip swaps engines but the X keymap an XIM client resolves characters from never changes — the flip is structurally invisible. This is a *hypothesis with HIGH local confidence*, not a verdict: ARCHITECTURE.md names three competing break points (arming race h1, engine re-mint churn h2, commit-not-landing h3), and the crucial in-repo fact is that the daemon **already has the software keymap** a daemon-side XIM fallback would need: in RU mode `feedKey` consumes the printable Latin key and commits its Cyrillic counterpart — `case a.mode == modeRU: … eng.CommitText(engine.NewIBusText(string(ru))); … return true` [VERIFIED: internal/session/actor.go:2194-2204]. So the spike's job is not to invent a mechanism but to pin **which link in the chain breaks** for wezterm specifically: transport identity → key delivery → mode flip → consume+commit → commit landing.

The spike must be run through the repo's established probe pattern (`mise run e2e-d01` appends per-run verdict lines to `docs/adr/d01-experiment-log.md` [VERIFIED: docs/adr/d01-experiment-log.md:14-17 — «Прогоны: строки дописывает кейс `mise run e2e-d01` … каждый прогон добавляет построчные вердикты»]), with one new probe capability the daemon does not offer today: a debug-gated ctl verb that forwards an arbitrary key event through the live engine (needed for the select-chord → PRIMARY rows; ForwardKeyEvent is a signal only the engine's own connection can emit — no external tool can). That verb is spike tooling: explicitly scoped, and its keep-or-remove is an owner gate decision.

**Primary recommendation:** four plans in two waves — (1) spec-delta + ADR-008 skeleton (docs, D-55 gate); (2) spike: wezterm root-cause probe ladder + capability table + A11Y matrix (three probe tracks, one e2e case family, verdict artifacts); (3) the wezterm fix + two-greens oracle (branch chosen by plan 2's verdict: config recipe first, daemon-side fallback second); (4) A11Y matrix may run parallel to (2) as its own small plan or fold into (2). Nothing from Phases 10–13 (terminalOp stepper, `ReadPrimary`, `AppendBlocklistEntry`, menu item, product recipes) is built here.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| TERM-01 | Layout switching works in wezterm; live spike confirms root cause (hypothesis: `enable_wayland=false` → XWayland/XIM client, flip never changes its X keymap); fix = native-Wayland config recipe first, daemon workaround fallback; acceptance = two consecutive green e2e oracles | Falsifiable probe ladder P0–P5 (Architecture Patterns) with evidence channels and explicit verdict-flip rules; native-Wayland recipe test incl. 20240203 hazards (#5125/#1772); pty-readback oracle design; XIM fallback shapes pre-derived from the existing `modeRU` consume+commit code path |
| TERM-04 | Autocorrect never fires in terminals (invariant); veto fixed by spec-delta, NOT via apps_blocklist; frozen negative matrix rows | Spec-delta target sections and exact invariant wording (Spec-Delta section); SPEC §11 already carries «пароли и терминалы — никогда» in the D-53 revision [VERIFIED: docs/SPEC.md:441-445]; frozen negative row precedent = `autocorrect-terminal-silent` case (wezterm fixture) [VERIFIED: test/e2e/case_autocorrect.go:778-806] |
</phase_requirements>

## Project Constraints (from workspace AGENTS.md / conventions)

No `./CLAUDE.md` exists; the workspace `AGENTS.md` conventions bind all phases:

- **1. Strict TDD** — every behavior task red → green → refactor; verify blocks run tests at each step. (Phase 9 is spike/docs-heavy; the rule binds the probe-case code and the fix plan.)
- **2. Green iteration** — every plan closes only with `go build ./...`, `go test -race ./...`, `golangci-lint run` clean.
- **3. mise, not make** — spike cases/tasks are mise tasks (`mise.toml [tasks]`), following the `e2e-d01` precedent.
- **4. golangci-lint strict from Phase 1** — no relaxations for probe code.
- **5. GitHub Actions current** — no CI changes needed by this phase (e2e stays manual-dispatch/live).
- **Process: D-55 absolute** — spec-delta before code; changes via PR to protected `main`; conventional commits.
- **Tech stack: Go ≥ 1.23, stdlib + godbus + minimal deps** — zero new modules (confirmed: the spike needs none).
- **Architecture axiom:** text injection through the IBus engine channel, never uinput in the daemon.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| wezterm flip root-cause evidence | Live desktop (probes) | e2e stand (probe case driver) | The defect lives in the mutter↔wezterm↔ibus bridge; only live wire evidence decides |
| wezterm flip FIX (recipe branch) | User config (`~/.wezterm.lua` / stand-owned test config) | — | `enable_wayland=true` is a wezterm-side setting; the daemon is unchanged |
| wezterm flip FIX (daemon fallback branch) | `internal/session` / `internal/engine` (minimal, spike-decided) | — | Only if the recipe fails; the software keymap already exists in `feedKey` modeRU |
| Terminal capability verdict table | e2e stand (probe case) + live probes | daemon `-debug` trace, new debug-gated forward verb | Verdicts must be produced by repeatable automated probes, recorded as artifacts |
| Spec-delta (SPEC revision) | `docs/SPEC.md` | — | D-55 process; dated revision with audit-trail comments (repo convention) |
| ADR-008 | `docs/adr/ADR-008-*.md` | — | Proposed → Accepted on spike evidence |
| A11Y probe matrix | Live desktop probes | e2e stand helper (`focus_helper.py`) | Per-app minimal a11y trigger; env verified in `/proc/<pid>/environ` |
| Two-greens wezterm oracle | e2e stand (`test/e2e`) | ydotool, pty read-back fixture | House discipline: wezterm IME arming race makes single-green meaningless |

## Standard Stack

### Core (spike tooling — all verified present this session; zero new Go modules)

| Tool | Version (live) | Purpose | Notes |
|------|----------------|---------|-------|
| `wezterm` CLI | 20240203-110809-5046fc22 (dpkg) | Spawn fixture windows; `--config-file`/`--config` overrides; `start -- PROG` runs a program in the pane | `--config-file <CONFIG_FILE>` and `--config name=value` present in `wezterm --help` [VERIFIED: live `wezterm --help` this session]; `wezterm start [OPTIONS] [PROG]...` with `--always-new-process` verified [VERIFIED: live `wezterm start --help`] |
| `ydotool` | installed (`/usr/bin/ydotool`) | Physical-path key injection (flip gesture, typing, chords) | 0.1.8 opens `/dev/uinput` directly, no daemon (v1.0 research, live-verified) |
| `wl-paste`/`wl-copy` | 2.2.1 (dpkg) | PRIMARY read-back probes: `wl-paste --primary --no-newline` | `-n` mandatory (trailing-newline pitfall) |
| `dbus-monitor` / `busctl --user` | installed | Wire evidence: engine CommitText signals, ibus name state, a11y bus probes | Note: dbus-monitor shows committed text — probe discipline below |
| `xlsclients`, `/proc/<pid>/environ` | OS | Transport identity proof (XWayland client vs native Wayland) | `WAYLAND_DISPLAY`/`DISPLAY` presence decides |
| `gsettings` / `ibus engine` CLI | GNOME 46 / IBus 1.5.29 | Flip drivers for probes; engine state | `gsettings get org.gnome.desktop.input-sources sources` → `[('ibus', 'goswitch-en'), ('ibus', 'goswitch-ru')]`, `current = 0` [VERIFIED: live probe this session] |
| `test/e2e/focus_helper.py` | in-repo | AT-SPI witness/a11y-tree probes (`witness`, `focused-app`, `focused-input-pid`, `text`, `witness-events`) [VERIFIED: test/e2e/focus_helper.py:620-637 — command dispatch verbatim] | python3-gi + gir1.2-atspi-2.0 installed (v1.0 research) |
| Daemon `-debug` flag | in-repo | Key trace evidence channel: `-debug` «enable key tracing (logs every keystroke…)» [VERIFIED: cmd/goswitchd/main.go:38] | The spike's primary evidence channel |

**Installation:** nothing to install. `go.mod` unchanged; e2e stand packages already present. The only *possible* machine change is the Phase-13-gated `flatpak install org.telegram.desktop` for the A11Y reachability probe — owner decision first (Open Questions). `flatpak` binary exists (`/usr/bin/flatpak`), **no flatpak apps installed** [VERIFIED: live probes this session].

**Version verification:** done live this session (dpkg + `--version` + `--help`). The committed milestone research pinned wezterm 20240203 as the **last tagged release ever** (upstream nightlies-only since Feb 2024; #6341 tracks the next release) [CITED: .planning/research/STACK.md] — no bump path without an owner decision.

## Package Legitimacy Audit

**Not applicable — this phase installs no external packages.** `go.mod` is untouched (zero new modules, per milestone STACK research and confirmed by the spike design: every probe is an OS tool or in-repo code). The one candidate machine-level install (Telegram flatpak) is a Phase-13 decision item gated on owner input, not a Phase 9 dependency.

**Packages removed due to [SLOP] verdict:** none.
**Packages flagged as suspicious [SUS]:** none.

## Recommended Plan Decomposition

**Four plans, two waves** (config: `parallelization: true`, `granularity: fine`):

| Plan | Content | Depends on | Why separate |
|------|---------|------------|--------------|
| **9-A: spec-delta + ADR-008 skeleton** (docs) | Dated SPEC revision (targets below); ADR-008 created **Proposed** with decision rows D1–D7 pre-formatted and evidence cells empty | nothing | D-55 is absolute and the delta's content (two-permission model, word mandatory, line feasibility-gated) is owner-locked already — it does not wait for spike results |
| **9-B: live spike — verdict tables** | Probe-case code (e2e case family + debug-gated forward verb + wezterm fixture extensions); run P0–P5 ladder → wezterm root-cause verdict; run capability-table probes (wezterm + VTE) → table artifact; run A11Y probe matrix → verdict artifact. ADR-008 evidence cells filled | 9-A | The spike is one coherent probe program; verdicts must be recorded before any fix is chosen |
| **9-C: wezterm flip fix + two-greens oracle** | The chosen branch: (a) recipe (owner config change + stand-owned native-Wayland test config + oracle case green ×2), or (b) daemon-side fallback (minimal session/engine change per 9-B verdict). ADR-008 → Accepted | 9-B | The fix branch is conditional; isolating it keeps 9-B pure-evidence even if the recipe fails |
| **9-D: A11Y probe matrix** (optional split from 9-B) | ZCode minimal-trigger test + Telegram snap re-pin + flatpak reachability + verdict artifacts | 9-A (parallel with 9-B/9-C) | Success criterion 5 demands it run «рано/параллельно»; it shares no code with 9-B and can go first if the terminal spike stalls |

Wave 1 = 9-A. Wave 2 = 9-B ∥ 9-D. Wave 3 = 9-C. If the owner prefers fewer plans, 9-D folds into 9-B as a second case family — the recommended split exists only to keep the Telegram machine-state question (flatpak install) from blocking the terminal probes.

## Architecture Patterns

### System Architecture Diagram — the spike's evidence flow

```
[ydotool (uinput, physical path)]          [debug-gated ctl probe verb]
        │  flip gesture / typing / chords         │  ForwardKeyEvent burst (daemon channel)
        ▼                                          ▼
┌──────────────────────── GNOME 46 Wayland session ───────────────────────┐
│  mutter ── zwp_text_input_v3 bridge ── ibus-daemon ── goswitch engines  │
│     │                                        │               (daemon    │
│     │ XIM bridge (XWayland)                  │                -debug:   │
│     ▼                                        ▼               key trace)│
│  wezterm (transport under test:     goswitchd journal: SetGlobalEngine │
│  XWayland/XIM or native Wayland)    focus pair, consume, commit events │
│        │                                                                     │
│        ▼ pty (ground truth)                     ┌──────────────────────┐    │
│  wezterm start -- bash -c 'cat > file'          │ wl-paste --primary -n│    │
└────────┼────────────────────────────────────────┴──────────┬───────────┘    │
         ▼                                                    ▼                │
  [verdict artifacts: docs/adr/ probe logs + ADR-008 evidence + table]         │
```

Entry points: two injection channels (physical ydotool = what a human does; daemon ForwardKeyEvent = what the product does). Evidence sinks: daemon `-debug` journal, ibus journal, pty byte file, PRIMARY read-back, `/proc` environ. Every probe ends in a recorded verdict line (d01-log precedent).

### Pattern 1: Falsifiable wezterm root-cause protocol (P0–P5)

**What:** a probe ladder where each step has a named command, a named observable, and an explicit rule for what evidence flips the root-cause verdict. No step's conclusion may rest on a web source (Anti-Pattern 4 from ARCHITECTURE.md).

| # | Probe | Command shape | Observable (evidence channel) | Verdict rule — what flips what |
|---|-------|---------------|-------------------------------|-------------------------------|
| P0 | Transport identity | `xlsclients` while wezterm focused; `tr '\0' '\n' < /proc/<wezterm-pid>/environ \| grep -E 'WAYLAND_DISPLAY\|DISPLAY'`; wezterm's own log/stderr on spawn | `WAYLAND_DISPLAY` absent + `DISPLAY` present + wezterm listed in xlsclients = XWayland client | **Confirms** the XIM hypothesis's precondition. If `WAYLAND_DISPLAY` IS set (native Wayland already), the whole XIM story is refuted → jump to h1/h2/h3 probes under the Wayland transport |
| P1 | Key delivery into the engine | Fresh wezterm (`start --always-new-process`), focus proven, `ydotool` type `ghbdtn`, daemon `-debug` key trace | Count of `"msg":"key"` records above base | Keys present = engine sees wezterm typing (arming race may need settle/пре-arm retries — both modes recorded, per the 2026-10-02 live finding). Keys absent after settle = arming/dead-context break (h1 class) — and **also** kills any daemon-side XIM fallback that needs ProcessKeyEvent |
| P2 | Mode flip observed by the daemon | Flip (`ydotool` single Right-Shift tap or `ibus engine goswitch-ru`) with wezterm focused; daemon journal | `focus_out`(goswitch-en)/`focus_in`(goswitch-ru) pair; internal mode RU in records | Pair present = flip reaches the daemon even for the XIM context. Pair absent = the wezterm context never re-arms onto the new engine (h2 class) |
| P3 | Consume+commit happens | With P1+P2 green: type `ghbdtn` in RU mode; `dbus-monitor` on the IBus connection (or daemon commit counters/log records with lengths, never content) | CommitText signals with Cyrillic-length payloads | Commits emitted = daemon side works end-to-end; break is client delivery (h3). No commits while keys visible = engine never consumed (mode/modeRU path not reached) |
| P4 | Commit landing (pty ground truth) | `wezterm start --always-new-process -- bash -c 'cat > /tmp/gs-spike-in'`; type; kill; `cat /tmp/gs-spike-in` (hexlify to compare bytes) | File bytes == `привет` (fix works) vs `ghbdtn` (defect reproduced) vs empty/mojibake | This is the **defect signature** and later the **oracle**. Latin bytes despite P3 commits emitted = XIM commit delivery broken — the exact fallback-branch discriminator |
| P5 | Raw-keymap path | With the IM context deliberately NOT armed (fresh spawn, immediate typing before arming — the race window): type `ghbdtn` | pty bytes Latin regardless of engine | Proves the fall-through path: XIM client resolves via the X keymap when the context is dead → the «structurally invisible» mechanism demonstrated, not just argued |

**Root-cause verdict matrix (ADR-008 evidence):**

| P0 | P1 | P2 | P3 | P4 | Verdict |
|----|----|----|----|----|---------|
| XIM | keys arrive | flip pair OK | commits emitted | Latin in pty | XIM commit-delivery break (h3) — daemon fallback = make commit land (see Pattern 3), or recipe |
| XIM | keys absent (after settle) | — | — | Latin in pty | XIM context arming dead (h1/dead-context) — daemon fallback likely impossible without the context; recipe is the only honest branch |
| XIM | keys arrive | no flip pair | — | Latin | Engine re-mint/arming churn for the XIM context (h2) — pre-arm on focus_in candidate |
| native Wayland already | any | any | any | ? | XIM hypothesis refuted; run h1/h2/h3 probes under zwp_text_input_v3; #5125/#1772 hazard list applies |

### Pattern 2: Native-Wayland candidate test (the recipe branch)

**What:** prove/refute the first fix candidate **without touching the owner's config**: `wezterm --config-file <stand-owned-test-config> start --always-new-process -- bash -c 'cat > /tmp/…'` with `enable_wayland = true` in the test config. (`--config-file` override verified present on this binary [VERIFIED: live `wezterm --help`]; alternatively `--config enable_wayland=true`.)
**Checks, in order:** (1) P0 re-run under the test config — `WAYLAND_DISPLAY` must now be set in the wezterm process environ; (2) full P1–P4 under native Wayland: flip → type → pty must read `привет`; (3) select-chord probes under the new transport (below — #4615 hazard lives exactly there); (4) stability: two consecutive rounds green in one session.
**Known 20240203 hazards to record as checkpoint rows, not assumptions:** wezterm#5125 «IME Not Working with IBus» affected exactly this build (fixed in nightly) [CITED: .planning/research/FEATURES.md]; wezterm#1772 text-input-v3 quirks [CITED: same]; gnome-shell «IME in wezterm dies after screen unlock» [CITED: same, LOW]. If native-Wayland IME is broken on this build, the branch decision (build bump vs daemon fallback) is an **owner gate** — 20240203 is the last tagged release.
**Only after the stand-owned config proves green** does the recipe move into `~/.wezterm.lua` (owner's daily config — mutation discipline in Runtime State Inventory).

### Pattern 3: Daemon-side XIM fallback branch shape (derived, not invented)

The fallback is NOT new mechanism — the software keymap already exists: `feedKey` in RU mode consumes the Latin keyval and commits the Cyrillic rune [VERIFIED: internal/session/actor.go:2194-2204]. What the spike decides is which link to repair:

- **P3 green / P4 Latin (commit doesn't land):** candidates: pre-arm the context on focus_in (cheapest activation exchange so the arming race window closes before human-scale keys — ARCHITECTURE.md's own suggestion); or, if XIM commit is structurally dead, forward the *Cyrillic* keyval instead of committing (client-side keysym resolution — risky: the X keymap is `us`-only, may not render; spike-probe before designing).
- **P1 red (keys never arrive):** no daemon fallback exists without the input context — the honest branch is the recipe, escalated to the owner.
- **P2 red (flip pair never arrives for the XIM context):** pre-arm/re-arm on focus_in in `internal/session` (the flipCredit machinery's neighborhood, actor.go:766-789 documents the re-mint behavior).

Any daemon-side fallback lands in `internal/session` (orchestration) and touches `internal/engine` only if a new emitter shape is required — matches ARCHITECTURE.md's component table («engine — touch ONLY if the Phase 1 verdict demands it»).

### Pattern 4: Verdict-table artifact (the d01-log precedent)

**What:** every probe run appends a dated verdict line to a committed docs file; the phase closes with tables, not chat. Format verbatim from the precedent: `| Проба | Команда | Наблюдаемое | Вердикт |` rows appended by the e2e case [VERIFIED: docs/adr/d01-experiment-log.md:16-19]. Recommended artifacts: `docs/adr/adr-008-spike-log.md` (wezterm ladder + capability table) and `docs/adr/d09-a11y-probe-log.md` (A11Y matrix). The e2e case owns the appends (mise task, `NOT in ci`), exactly like `e2e-d01`.

## The Terminal Capability Verdict Table (what Phase 9 must produce)

Rows are the success-criterion-4 columns; probes are per terminal. **wezterm** (owner's daily) and **GNOME Terminal/VTE** (second target). Every cell = a recorded verdict from a live probe, never a web claim.

| Row | Probe (wezterm) | Probe (GNOME Terminal/VTE) | Maps to Phase 10/12 decision |
|-----|------------------|----------------------------|------------------------------|
| caps reporting | daemon log `SetCapabilities` for the wezterm context (owner-observed `0x9` — preedit+focus, no surrounding bit [CITED: .planning/research/PITFALLS.md #6, live 2026-10-08]) | same for the VTE context — record the actual value; VTE false-advertises surrounding text (RHBZ 1847347 [CITED: .planning/research/ARCHITECTURE.md]) | Gate design: caps untrustworthy → identity/role-first gate; level-1 ladder never executed in terminals |
| ProcessKeyEvent delivery | ydotool typing → `-debug` key trace count (arming race: pin both delivery modes, positive control = zenity keystroke — `acControlKeystroke` precedent [VERIFIED: test/e2e/case_autocorrect.go:840-853]) | same | Word-correction arming: does the daemon buffer even fill in this terminal? |
| commit_text lands in pty | P4 probe (pty bytes) | Same probe with gnome-terminal spawned with a pty-capturing program; **plus the cheap shortcut check**: flip→RU→type→pty already `привет`? If yes, VTE word correction is nearly free (buffer+commit, no read channel) | Write-back transport viability per terminal; VTE shortcut verdict |
| select-chord → PRIMARY read-back | Two channels: (a) ydotool physical Shift+Home / Ctrl+Shift+Left → `wl-paste --primary -n` (tests the *terminal's* selection semantics — the kill-criterion question); (b) daemon-channel forward via the new debug-gated verb (tests ForwardKeyEvent delivery incl. the #4615 Shift-lost signature: selection made vs cursor jumped). wezterm additionally with the CopyMode recipe (config-recipe selection from milestone STACK [CITED: .planning/research/STACK.md]) | (a) only — VTE has no keyboard selection at all (open request since 2002 [CITED: .planning/research/PITFALLS.md #1]); expected verdict: no selection → **kill criterion fires**, line scope shrinks to wezterm | **Kill criterion for Phase 12** (VTE line = dead unless a capture primitive is redesigned); word chord for Phase 10 (wezterm: recipe vs Ctrl+Shift+Left vs #4615-cursed forwarded chords) |
| a11y identity visibility | wezterm expected **absent** from the a11y bus (already pinned by the terminal-silent case [VERIFIED: test/e2e/case_autocorrect.go:866-867 — «wezterm is off the a11y bus»]); re-confirm + record | gnome-terminal IS a GTK3 a11y citizen: probe `focused-app` witness + `appid` role for the VTE surface (record role/name forms observed) | Phase 11 MENU identity degradation (wezterm = disabled item); Phase 10 gate discriminator (role 60 list) |

**Verdict-to-phase mapping (the table's whole purpose):** Phase 10 consumes rows 1–5 for the word chord + read channel per terminal; Phase 12 consumes the select-chord row as its **kill criterion** — «в VTE нет выделения с клавиатуры… область честно сужается до wezterm; вердикт о сужении — за владельцем» [VERIFIED: .planning/REQUIREMENTS.md:14, TERM-03]. Phase 11 consumes the a11y row. Phase 13 consumes the A11Y matrix (below), not this table.

## Spec-Delta Targets (D-55 — dated revision, before code)

Repo convention: each revision keeps the old wording verbatim as an audit-trail HTML comment, then states the new rule with date and owner decision (three in-file precedents: SPEC §4.3 rev G-2-3 [VERIFIED: docs/SPEC.md:135-144], §4.4 rev G-5-5 [VERIFIED: docs/SPEC.md:148-157], §4.5/§11 D-53 [VERIFIED: docs/SPEC.md:424-460]).

| SPEC section | Revision content |
|--------------|------------------|
| **§4.3 Поведение коррекции** | New terminal clause: manual correction in terminals — yes (new permission); word scope mandatory, whole-line feasibility-gated by the Phase 9 capability table; buffer/verify semantics for terminals defer to ADR-008 |
| **§4.4 Интеграция с окружением** | Terminal transport statement: correction and flip delivery in terminals ride the daemon's own IBus channel (ForwardKeyEvent + commit_text); wezterm supported in native-Wayland mode (config recipe) with the daemon-side fallback per ADR-008 if the spike dictates |
| **§5 Нефункциональные требования** | Latency-budget completion clause: reaction < 50 ms (hotkey → first wire action; current §5 text: «Производительность: действия без ожидания — < 50 мс» [VERIFIED: docs/SPEC.md:334]); terminal completion best-effort with per-phase deadlines (select→read leg is subprocess-bound ~150–400 ms) — otherwise the acceptance matrix «fails» a healthy design |
| **§7.2 e2e** | Terminal matrix discipline: wezterm rows demand two consecutive greens (arming race); frozen autocorrect-negative terminal rows carried into the new matrix version |
| **§11 Решение по автокоррекции** | Amendment: the two-permission model and the veto's home. §11 already pins «полевой role-гейт пропускает (текстовый ввод; **пароли и терминалы — никогда**)`» [VERIFIED: docs/SPEC.md:441-445]; the revision states explicitly that the terminal veto lives **in the autocorrect arming condition** (role/terminal check at arm time), never in `apps_blocklist` — because D-53's own rule «Неизвестная идентичность приложения НЕ запрет» [VERIFIED: docs/SPEC.md:448-449] makes a blocklist entry unreliable exactly where terminal identity is weakest |

**TERM-04 invariant wording (for the planner to draft into the delta):** «Автокоррекция никогда не армируется в контексте терминала (роль AT-SPI 60 или сконфигурированный терминальный движок/identity). Вето проверяется в точке армирования автокоррекции ДО запуска детектора; оно не выражается через `autocorrect.apps_blocklist` и не зависит от наблюдаемости identity. Ручная коррекция (двойной/тройной Shift) в терминалах разрешена. Контракт закреплён замороженными негативными рядами e2e-матрицы.» — role value 60 verbatim from `RoleTerminal uint32 = 60 // terminal — never corrected` [VERIFIED: internal/appid/appid.go:73-79].

## ADR-008 Skeleton (decisions it must settle; evidence each row needs)

Created **Proposed** in plan 9-A with empty evidence cells; driven to **Accepted** in 9-C citing `adr-008-spike-log.md` rows by name. Format follows ADR-006/ADR-007 (Status with dated verdicts, Context, Decision, Ammendments, Evidence table).

| # | Decision | Options | Evidence required from the spike |
|---|----------|---------|----------------------------------|
| D1 | Transport per terminal (read/write channels) | daemon ForwardKeyEvent + commit_text (only real candidate; external injectors refuted live 2026-10-08 [CITED: .planning/PROJECT.md milestone record]) | P1/P3/P4 rows for both terminals |
| D2 | wezterm flip fix branch | (a) native-Wayland recipe; (b) build bump (owner); (c) daemon-side fallback | P0–P5 under both configs; the verdict matrix |
| D3 | Read channel per terminal | wezterm: CopyMode-recipe selection → `wl-paste --primary -n`; VTE: daemon keystroke buffer only (no read channel) | select-chord row (both channels), VTE buffer+commit shortcut check |
| D4 | Write-back primitive | N×BackSpace replay + `commit_text` (never commit-over-selection — terminal selections are visual-only [CITED: .planning/research/STACK.md §a.1]; never a paste chord — Ctrl+C is SIGINT) | commit-landing row; #4615 modifier sanity on forwarded bursts |
| D5 | Clipboard policy | PRIMARY-only; CLIPBOARD byte-untouched; `-n` everywhere; save/restore-or-document the PRIMARY clobber | probe records which selections each chord overwrites (copy_on_select behavior) |
| D6 | Abort / safe point | free abort before the first forwarded BackSpace; past it one-way replay + commit; verify-after logs and stops, never auto-repairs | design row — pinned by D1–D4 evidence, no extra probe |
| D7 | Scope contract | shells (bash/zsh) in contract; canonical-mode and TUI apps out of contract (three erase semantics; BackSpace-as-command) [CITED: .planning/research/PITFALLS.md #2/#4] | witnessed by the shortcut check + documented as contract |

## E2E Oracle: «two consecutive greens» in wezterm

**Precedent to extend:** `autocorrect-terminal-silent` already owns the wezterm fixture — `startWezterm` spawns `wezterm start --always-new-process` (never delegates to an owner instance — the chromium-profile isolation lesson) [VERIFIED: test/e2e/case_autocorrect.go:751-763], focus is proven negatively via the a11y witness («wezterm is off the a11y bus») [VERIFIED: test/e2e/case_autocorrect.go:855-881], and the case already survived the live discovery that wezterm's typing **races the IME arming** — one run delivered all 7 keystrokes, another zero [VERIFIED: test/e2e/case_autocorrect.go:784-802].

**The flip-fix oracle (new case, `mise run e2e-wezterm-flip`):**
1. Stand-owned daemon config (EN mode, autocorrect off); `-debug` on; positive control keystroke visible in the key trace (deafness counter-evidence).
2. Spawn the wezterm fixture **with a pty-capturing program**: `wezterm start --always-new-process --config-file <stand-config> -- bash -c 'cat > /tmp/<case>-in'` (PROG passing verified [VERIFIED: live `wezterm start --help`]; stand-config pins `enable_wayland=true` on the recipe branch — the *stand* config, not the owner's).
3. Focus proof: witness stops naming the previous surface (existing helper), **plus** an arming check: one dummy keystroke must appear in the key trace before the real round (the arming race made a formal gate; retry window bounded).
4. Drive: ydotool flip gesture (single Right Shift — the daemon's own switch) → type `ghbdtn` → bounded settle.
5. Assert: pty file bytes == `привет` (hexlify before compare — encoding discipline); daemon journal shows the flip pair; **CLIPBOARD untouched is not yet asserted here** (no clipboard involvement on this path — that row belongs to Phase 10's matrix).
6. **Two consecutive greens = the whole round (steps 2–5) runs twice back-to-back in the same case invocation, both rounds green.** A single green is void by house rule (arming race + WINDOWS #13 fresh-session drift precedent: each row typing right after a flip can lose the first keys on fresh sessions [VERIFIED: .planning/WINDOWS.md row 13]).

**Drift/flake hazards handled:** fresh spawn per round (no state); arming gate with bounded retries; settle budgets; the case never runs in headless CI (live-session mise task, `NOT in ci` family); the D-48 fresh-session night gate stays the owner's formal instrument.

## A11Y Probe Matrix (criterion 5 — early/parallel)

| Probe | Steps | Verdict artifact | Consumed by |
|-------|-------|------------------|-------------|
| ZCode minimal a11y trigger | Current state (daemon already sets `toolkit-accessibility`, 'zcode' visible in the tree — live-probed 2026-10-08 [CITED: .planning/research/STACK.md §c]); then the `.desktop` override mechanism test: create user-local override with `--force-renderer-accessibility`, restart ZCode, assert tree presence via `focus_helper.py` (`focused-app`/`text` commands); **env checked in `/proc/<zcode-pid>/environ`** | minimal-trigger row per app (IsEnabled vs ScreenReaderEnabled vs flag vs env) | Phase 13 recipe; Phase 11 (Electron identity exists only post-a11y — PITFALLS #12) |
| Telegram snap AppArmor wall (re-pin) | snap 7.2.9 rev 7150 [VERIFIED: live `snap list` this session]; re-run the denial probe (`Accessible.GetRole` → expect "Access denied"); verify `QT_ACCESSIBILITY` in `/proc/<telegram-pid>/environ` (present per 2026-10-08 probe [CITED: .planning/research/STACK.md §c]) | wall confirmed/denied + env proof | Phase 13 owner decision (unconfined build vs daemon fallback) |
| Flatpak reachability (org.telegram.desktop) | **Owner gate first** (installs a second Telegram on the daily machine): `flatpak install` → launch → a11y-tree probe + `/proc` env check | flatpak-a11y reachable: yes/no | Phase 13 ladder rung 2 |
| ZCode override residue | Decide keep-or-rollback of the probe-created override (it is Phase 13 product scope to ship it properly) | recorded decision | Phase 13 |

Artifacts land in `docs/adr/d09-a11y-probe-log.md` (d01 format). Phases 11/13 consume the log rows, never this research file.

## Code Seams the Spike Touches — and what NOT to build

**Touched (spike-grade only):**
- `test/e2e/` — new probe/oracle case family (fixture extension of `startWezterm` with `--config-file` + PROG; pty-file oracle helper; flip-round runner) + `mise.toml` tasks (`NOT in ci` family, `e2e-d01` precedent).
- `cmd/goswitchctl` — one **debug-gated** probe verb (e.g. `probe-forward <keyspec>`) wired to the engine emitter seam, so the daemon-channel select-chord rows are drivable. ForwardKeyEvent is a signal only the engine's own connection can emit — no external tool substitutes. Keep-or-remove at the phase gate = owner input. (`internal/engine.Emitter` interface and the `forwardCtrlV` four-event shape at internal/session/actor.go:1710-1713 are the pattern to mirror.)
- `internal/engine/keys.go` — keyval constants needed by the probe verb, in the file's established verified-verbatim pattern (`KeyV = 0x076 // … paste half of the clipboard rung's Ctrl+V forward burst` [VERIFIED: internal/engine/keys.go:49-51]): `KeyHome`, `KeyEnd`, `KeyLeft`, `KeyRight` — values to be transcribed verbatim from `/usr/include/ibus-1.0/ibuskeysyms.h` in-task (0xff50/0xff57/0xff51/0xff53 are the expected X keysyms — **not verified this session, never from memory**: the 02-05 class trap is documented in keys.go itself [VERIFIED: internal/engine/keys.go:36-41]).
- `~/.wezterm.lua` — only in plan 9-C, only after the stand-owned config proves green, with backup/restore (Runtime State Inventory).

**NOT built (Phases 10–13 scope — hard guards):**
- `terminalOp` stepper, select-chord helpers (`forwardShiftCtrlLeft`/`forwardShiftHome` as product code), failure vocabulary, config knobs (`correction.terminal`) — Phase 10.
- `clipboard.ReadPrimary/SetPrimary/ClearPrimary` as package API — Phase 10 (the probe shells out to `wl-paste` directly; the package seam is designed there).
- `config.AppendBlocklistEntry`, indicator item 16, `MenuSync.SetBlocklistTarget` — Phase 11.
- Line-capture design, End-normalization — Phase 12.
- `.desktop` override as shipped product/recipe/installer work — Phase 13 (Phase 9 only probes and records).
- No SPEC §8 поставочные-артефакты changes, no new packages, no go.mod changes.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Key injection into the live desktop | a custom injector/uinput tool | `ydotool` (installed, physical path, e2e-proven) | 0.1.8 works daemonless; the daemon-channel forwarding already exists in-repo |
| AT-SPI tree/witness probes | new introspection scripts | `test/e2e/focus_helper.py` command set [VERIFIED: test/e2e/focus_helper.py:620-637] | Already handles pid-keyed focus, roles, text read-back |
| Verdict record-keeping | ad-hoc notes in plans/PRs | d01-log append pattern via the e2e case [VERIFIED: docs/adr/d01-experiment-log.md:14-19] | Repeatable, dated, diffable; the ADR cites it |
| wezterm config variations | editing the owner config per experiment | `--config-file` / `--config` CLI overrides with a stand-owned file | Owner's daily config must not churn during the spike |
| pty ground truth | terminal-screen scraping/OCR | spawn the pane with `cat > file` and diff bytes | Byte-exact, scriptable, no a11y dependency (wezterm is off the bus anyway) |
| D-Bus wire evidence | a hand-rolled monitor daemon | `dbus-monitor`/`busctl --user` one-liners inside the probe case | Tools ship with the OS; content-handling discipline below |

**Key insight:** the spike's only genuinely new code is the probe verb and the oracle case — everything observational already exists as installed tools or in-repo precedents. A spike that grows a framework has leaked Phase 10 scope.

## Runtime State Inventory

Not a rename/refactor phase, but the spike **mutates live desktop state** — the inventory the planner needs:

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Owner config | `~/.wezterm.lua` sets `enable_wayland = false` [VERIFIED: ~/.wezterm.lua:77-80] | Plan 9-C: backup → apply recipe (or daemon fallback) → owner verifies daily flow; rollback path documented in the plan |
| Session config | `sources = [('ibus', 'goswitch-en'), ('ibus', 'goswitch-ru')]`, `current = 0` [VERIFIED: live gsettings] | None — probes must not rewrite sources; assert-only |
| User-local desktop entries | `~/.local/share/applications/` already carries `org.telegram.desktop._3e485da34fc040f9218e3891ecde1e6c.desktop` [VERIFIED: live ls] — an override mechanism precedent, do not clobber | A11Y probe adds `zcode.desktop` override; keep-or-rollback decided in-probe |
| Snap state | telegram-desktop 7.2.9 rev 7150, chromium 153 [VERIFIED: live snap list] | Read-only probes (GetRole denial re-pin) |
| Flatpak | binary installed, zero apps [VERIFIED: live] | `flatpak install org.telegram.desktop` only after owner gate (Open Questions) |
| Test artifacts | `/tmp/gs-spike-in` pty files, probe logs | Tmp-only; never commit pty captures (secret-hygiene class) |

## Common Pitfalls

### Pitfall 1: The single-green wezterm oracle
**What goes wrong:** flip+type round passes once, declared fixed; the next morning the owner's terminal is broken again.
**Why:** wezterm's IME arming races injection — one run delivered all 7 keystrokes, another zero (live finding pinned in the terminal-silent case) [VERIFIED: test/e2e/case_autocorrect.go:784-802]; plus WINDOWS #13 first-keys-after-flip drift on fresh sessions [VERIFIED: .planning/WINDOWS.md row 13].
**How to avoid:** the oracle runs its full round twice back-to-back; arming gate with a dummy keystroke + bounded retries before each round; fresh spawn per round.
**Warning signs:** a fix plan whose acceptance says «e2e green» without the ×2 discipline.

### Pitfall 2: Designing the fix from the issue tracker
**What goes wrong:** the plan picks #5125/#1772 workarounds as the fix.
**Why:** those threads describe other manifestations on other builds; ARCHITECTURE.md bans this as Anti-Pattern 4.
**How to avoid:** ADR-008 evidence cells cite spike-log rows, never URLs; upstream links go into the log as context only.
**Warning signs:** any ADR-008 row whose evidence is a web source.

### Pitfall 3: Trusting caps in terminals
**What goes wrong:** the spike «proves» VTE supports surrounding text (its caps advertise it) and the ladder branches on it.
**Why:** VTE false-advertises surrounding text (RHBZ 1847347) and stubs it [CITED: .planning/research/ARCHITECTURE.md]; wezterm's honest `0x9` differs.
**How to avoid:** the capability table records observed caps but no decision keys on them; identity/role is the gate discriminator.
**Warning signs:** any verdict-table row that says «caps say so» without a behavioral probe.

### Pitfall 4: Spike code becoming product code
**What goes wrong:** the probe verb grows into the Phase 10 stepper; ReadPrimary lands in `internal/clipboard` "since we're here".
**Why:** momentum; the seams are adjacent.
**How to avoid:** the NOT-built list above is a plan-level guard; probe verb is debug-gated and its fate is an owner gate.
**Warning signs:** a 9-B plan containing `correction.terminal` config knobs or terminalOp types.

### Pitfall 5: Stale/garbage PRIMARY read-back
**What goes wrong:** the select-chord probe reads the user's hours-old selection and the table records a false "selection works" verdict.
**Why:** no selection was actually made (VTE: none exists; wezterm #4615 Shift-lost); `wl-paste --primary` happily returns stale data; trailing newline off-by-one without `-n`.
**How to avoid:** pre-clear/set a sentinel PRIMARY before the chord, assert the sentinel changed to the expected prefix; `-n` everywhere; empty read is a defined state.
**Warning signs:** verdict table rows without a sentinel step.

### Pitfall 6: Probe evidence leaking secrets
**What goes wrong:** `dbus-monitor` captures or the pty file holds real user text; `-debug` daemon logs passwords.
**Why:** monitoring the IBus connection shows committed content by design; the daemon's own docs call `-debug` "passwords become visible" [VERIFIED: cmd/goswitchd/main.go:38].
**How to avoid:** probes type only synthetic corpus (`ghbdtn`); pty captures to tmp, hexlified, never committed; probe runs on the stand's own windows, not owner apps.
**Warning signs:** any committed artifact containing pty bytes.

### Pitfall 7: Owner-config churn during the spike
**What goes wrong:** experiments edit `~/.wezterm.lua` in place; the owner's daily tool breaks mid-phase.
**Why:** the fix target IS the owner config.
**How to avoid:** stand-owned config via `--config-file` for every experiment; the owner's file changes once, in 9-C, after green proof, with a backup.

## Code Examples

### P0 — transport identity probe
```bash
# wezterm focused; find the gui process
pid=$(pgrep -f wezterm-gui | head -1)
tr '\0' '\n' < /proc/$pid/environ | grep -E '^(WAYLAND_DISPLAY|DISPLAY|XMODIFIERS)='
xlsclients                      # wezterm listed => XWayland client
```

### P4 — pty ground-truth oracle (fixture shape)
```bash
wezterm start --always-new-process --config-file /tmp/gs-e2e-wezterm.lua -- \
  bash -c 'cat > /tmp/gs-spike-in'
# drive: ydotool <flip>; ydotool type ghbdtn
# after teardown: compare bytes (hexlify, never trust locale)
xxd /tmp/gs-spike-in     # expect: d0 bf d1 80 d0 b8 d0 b2 d0 b5 d1 82 ("привет")
```

### Native-Wayland recipe (stand-owned test config)
```lua
-- /tmp/gs-e2e-wezterm.lua — minimal; the owner's keys/plugins are NOT loaded
local wezterm = require 'wezterm'
return { enable_wayland = true }
```

### Select-chord sentinel probe (PRIMARY read-back)
```bash
printf 'SENTINEL' | wl-copy --primary           # known pre-state
# forward/inject Shift+Home over a typed 'ghbdtn' line
sleep 0.2                        # settle; knob spike-pinned ~50-100 ms (SUMMARY gap)
wl-paste --primary --no-newline  # expect 'ghbdtn' prefix, NOT 'SENTINEL'
```

### wezterm CopyMode select recipe (only if the chord row fails) [CITED: .planning/research/STACK.md]
```lua
table.insert(keys, {
  key = 'Home', mods = 'SHIFT',
  action = wezterm.action.Multiple{
    wezterm.action.ActivateCopyMode,
    wezterm.action.CopyMode { SetSelectionMode = 'Cell' },
    wezterm.action.CopyMode 'MoveToStartOfLineContent',
    wezterm.action.CopyMode 'Close',   -- Close LEAVES the selection active
  },
})
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| wezterm tagged releases | nightlies-only upstream; 20240203 = **last tag ever** (#6341 tracks next) | Feb 2024 | Any build-bump branch = nightly (owner risk decision); recipes must target the installed build |
| wezterm X11-only | `zwp_text_input_v3` native Wayland supported since 20220807 | 2022 | The recipe branch is a supported path on this build [CITED: .planning/research/FEATURES.md] |
| goswitch single-engine internal flip (ADR-001 Option B) | two-engine `SetGlobalEngine` flip (ADR-006, amended ×3) | 2026-09-30 | The defect exists under the two-engine model; probes must assert the flip pair, not dconf |
| GUI-style select-replace correction | terminals: backspace-replay + commit (ADR-003 level 2 goes live in Phase 10) | v1.2 design | Phase 9's D4/D6 rows exist to freeze this before Phase 10 code |

**Deprecated/outdated:** AskUbuntu-era claims that `gsettings current` is read-only (dead on GNOME 46 — the key is simply not written by the shell; the daemon reads the factual global engine instead [VERIFIED: internal/activate/activate.go:137-149]); `xclip`/`xsel` (X11-only); GTK4-Wayland forwarded-events as a channel (WINDOWS #4 waiver — gnome-terminal is GTK3 and unaffected).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The XIM root-cause story (`enable_wayland=false` → XWayland/XIM → flip invisible) — HIGH local confidence from committed research, **not yet proven on the wire** | Summary, Pattern 1 | The whole P-ladder is built around falsifying it; if P0 refutes, plan 9-B pivots to h1/h2/h3 under native Wayland (the ladder covers this branch) |
| A2 | commit_text landing in a terminal reaches the pty byte-exactly (the oracle's premise) | Pattern 1 P4, Oracle | If an IM commits via preedit/other path and the pty misses bytes, the oracle needs the pty program swapped (e.g. `script -q` or a readline reader) — cheap, first-task verification |
| A3 | `--config-file` loads the full config path (keys/plugins resolution) for the spawned GUI, not just some settings | Pattern 2, Oracle | If partial, fall back to `--config enable_wayland=true` overrides or `WEZTERM_CONFIG_FILE` env — verify in the first 9-B task |
| A4 | Forwarded Cyrillic keyvals (fallback candidate under h3) render client-side in an XIM client with a `us`-only keymap | Pattern 3 | If unrenderable, the h3 fallback narrows to pre-arm/arming repair — spike-probe before design (stated in-plan) |
| A5 | VTE actually reports a caps value distinguishable from wezterm's `0x9` (expected: includes the surrounding-text bit) | Capability table row 1 | None material — the row records the observed value; no decision keys on caps |
| A6 | ydotool can drive the flip gesture and typing into the wezterm fixture window for the oracle | Oracle | Already precedented: the terminal-silent case injects into wezterm today [VERIFIED: test/e2e/case_autocorrect.go:825-833]; focus proof helper exists |
| A7 | ZCode's a11y trigger today is `toolkit-accessibility` at startup (daemon-set), and the override test adds the deterministic belt | A11Y matrix | If ZCode needs `ScreenReaderEnabled` instead, Phase 13's recipe changes and the session-wide-semantic owner acceptance fires (PITFALLS #12) — the probe exists precisely to decide this |

## Open Questions (owner input points) (RESOLVED)

> **Resolved-by-plan (2026-10-08, планирование Фазы 9):** все шесть вопросов закреплены за чекпойнт-гейтами планов и НЕ решаются исполнителем молча: OQ 1 (правка `~/.wezterm.lua`) → checkpoint Task 1 плана `09-03`; OQ 4 (ветка фикса при сломанном нативном IME 20240203) → checkpoint Task 1 плана `09-03`; OQ 6 (пред-вердикт сужения строки VTE) → checkpoint Task 1 плана `09-03`; OQ 3 (судьба probe-глагола) → checkpoint Task 3 плана `09-03`; OQ 2 (flatpak-согласие) → checkpoint Task 2 плана `09-04`; OQ 5 (ScreenReaderEnabled-семантика) → checkpoint Task 2 плана `09-04` (условный — открывается только при сессионно-семантическом триггере). Формулировки вопросов ниже оставлены дословно как исходник; вердикты владельца фиксируются строками журналов спайка (`adr-008-spike-log.md`, `d09-a11y-probe-log.md`).

1. **May plan 9-C edit `~/.wezterm.lua` directly** (backup + apply recipe) once the stand-owned config proves green ×2, or does the owner apply the final recipe manually and confirm the daily flow? (The owner's daily-driver config; mutation discipline favors explicit consent.)
2. **Flatpak probe consent:** may Phase 9 `flatpak install org.telegram.desktop` on this machine for the A11Y reachability row (~100+ MB, a second Telegram), or is that row deferred to Phase 13 planning?
3. **Probe-verb fate:** keep the debug-gated `goswitchctl probe-forward` after the phase (debug harness for future terminal work) or remove at phase close?
4. **If the 20240203 build's native-Wayland IME is broken (#5125 class):** bump to a nightly (abandon the last-tagged pin) vs pursue the daemon-side fallback vs hybrid (nightly only for the owner's machine)? This decision gate exists only if the recipe probe fails.
5. **If ZCode's minimal trigger turns out to be `ScreenReaderEnabled=true`:** session-wide screen-reader semantics — accept the side effects (Telegram warning bar class) or constrain to launch-flag-only recipes?
6. **VTE line-scope pre-verdict:** if the kill criterion fires as expected (no VTE selection primitive), is «line scope = wezterm-only» acceptable to record now, or does the owner want the Ctrl+U/kill-ring redesign spike queued for v1.3 before Phase 12 planning?

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| wezterm (fixture + target) | flip spike, capability table | ✓ | 20240203-110809-5046fc22 (dpkg) | — |
| gnome-terminal / VTE | capability table | ✓ | 3.52.0 (dpkg) | — |
| ydotool + `/dev/uinput` + `input` group | all typing/gesture injection | ✓ (v1.0-verified stand) | 0.1.8 | own ~150-LOC uinput injector (parked) |
| wl-clipboard | PRIMARY probes | ✓ | 2.2.1 (dpkg) | — |
| dbus-monitor / busctl / xlsclients | wire + transport evidence | ✓ | OS/brew | — |
| python3-gi + gir1.2-atspi-2.0 | focus_helper.py, a11y tree | ✓ (v1.0-verified) | 3.48.2 / 2.52.0 | raw busctl on the a11y bus |
| live GNOME Wayland session | every probe/oracle | ✓ (owner desktop; e2e = manual mise tasks) | GNOME 46 | none — this phase is live-only |
| flatpak | Telegram reachability row | ✓ binary, 0 apps | — | defer row to Phase 13 (owner gate) |
| goswitchd `-debug` + ctlsvc | evidence channel | ✓ in-repo | — | — |

**Missing dependencies with no fallback:** none — every probe runs on the existing stand.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + race detector; live e2e stand `test/e2e` (manual mise tasks) |
| Config file | `mise.toml` (tasks), strict `.golangci.yml` |
| Quick run command | `mise run build && mise run test` |
| Full suite command | `mise run ci` (build+vet+lint+test); live: `mise run e2e-matrix-v4` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TERM-01 | wezterm flip works: type `ghbdtn` after flip → pty reads `привет`, twice | e2e (live) | `mise run e2e-wezterm-flip` (new; two rounds internal) | ❌ Wave 0 (plan 9-B/9-C) |
| TERM-01 | root-cause probes produce verdict rows | e2e probe case (live) | `mise run e2e-spike-wezterm` (new; appends to `docs/adr/adr-008-spike-log.md`) | ❌ Wave 0 (plan 9-B) |
| TERM-01 | capability table rows (both terminals) | e2e probe case (live) | `mise run e2e-spike-caps` (new) | ❌ Wave 0 (plan 9-B) |
| TERM-04 | autocorrect never fires in terminals (frozen negative re-pin under the new matrix) | e2e (live) | `mise run e2e-autocorrect-terminal-silent` (exists [VERIFIED: mise.toml:194-196]) | ✅ |
| TERM-04 | spec-delta carries the invariant; veto not in apps_blocklist | docs review (manual gate) | owner review at verify-work | n/a |
| — (A11Y) | probe matrix verdicts recorded | live probes + log | `mise run e2e-spike-a11y` (new) | ❌ Wave 0 (plan 9-B/9-D) |

### Sampling Rate
- **Per task commit:** `mise run ci` (headless gates — green-iteration directive).
- **Per wave merge:** full headless suite; live probe tasks run at task completion on the owner session.
- **Phase gate:** two consecutive green `e2e-wezterm-flip` rounds + matrix v4 still green + ADR-008 Accepted before `/gsd:verify-work`.

### Wave 0 Gaps
- [ ] `test/e2e/case_spike_wezterm.go` (probe ladder + flip oracle) — TERM-01
- [ ] `test/e2e/case_spike_caps.go` (capability table probes) — TERM-01
- [ ] `test/e2e/case_spike_a11y.go` (A11Y matrix) — criterion 5
- [ ] debug-gated probe verb in `cmd/goswitchctl` + unit corpus (TDD)
- [ ] `internal/engine/keys.go` new keyval constants (verbatim-from-header rule)
- [ ] mise tasks for the three new cases (`NOT in ci` family)

## Security Domain

`security_enforcement: true`, ASVS level 1 (config). This phase is probes + docs + one config fix; the applicable surface:

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | n/a (no auth surface) |
| V3 Session Management | no | n/a |
| V4 Access Control | yes (probe etiquette) | probes never touch owner windows/data; stand-owned surfaces only (house rule since v1.0) |
| V5 Input Validation | yes | probe verb validates its keyspec strictly (reject unknown); config-file fixture is stand-owned |
| V6 Cryptography | no | n/a |
| V14 Config | yes | owner config edited once, backup + rollback documented; no `%h`/shell interpolation in any new unit/launcher content |
| V16 Data Logging | yes | `-debug` key trace exposes keystrokes («passwords become visible» [VERIFIED: cmd/goswitchd/main.go:38]); probes type synthetic corpus only; pty captures tmp-only, never committed; PRIMARY sentinel uses non-secret tokens |

### Known Threat Patterns for this phase

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Probe artifacts carrying user text (pty files, dbus-monitor captures) | Information Disclosure | synthetic corpus only; hexlified tmp files; artifact commit gate in plan review |
| PRIMARY sentinel/sentiment leaking to clipboard managers | Information Disclosure | sentinel tokens are non-secret; CLIPBOARD untouched by policy (D5) |
| debug probe verb as a misuse surface (any process can call ctl) | Tampering/Elevation | verb is debug-gated (daemon already runs `-debug` only in e2e), validates keyspecs, forwards keys only — same power as the existing MACR/clipboard bursts; session-bus access already implies local trust |
| Flatpak/snap probes altering the owner's app set | Denial of Service (UX) | owner gate before any install; probes read-only otherwise |

## Sources

### Primary (HIGH confidence — read or probed this session)
- `~/.wezterm.lua:77-80` — `enable_wayland = false` (owner config state)
- Live system probes (2026-10-08): `wezterm --version` / `wezterm --help` (`--config-file`, `--config`, `start -- PROG`), `wezterm start --help`, dpkg (wezterm, gnome-terminal 3.52.0), `gsettings` input-sources (two goswitch engines, current=0), `snap list` (telegram-desktop 7.2.9 rev 7150, chromium 153), `flatpak` binary present/no apps, tool paths (wl-clipboard 2.2.1, ydotool, xlsclients, dbus-monitor, busctl)
- `internal/session/actor.go:2144-2218` — `feedKey`: the RU-mode consume+commit software keymap (fallback-branch foundation); `actor.go:1692-1728` `forwardCtrlV` burst shape; `actor.go:766-789` flipCredit/re-mint context
- `internal/engine/keys.go` — keyval/mask constants and the verified-verbatim transcription rule
- `internal/install/install.go:1258-1263` — `wireEngines()`: goswitch-en layout "us", goswitch-ru layout "ru" (the flip swaps engines, and the declared XKB layouts with them — the exact fact the XIM hypothesis turns on)
- `internal/appid/appid.go:73-79` — role constants: `RolePasswordText=40`, `RoleTerminal=60 // terminal — never corrected`, `RoleText=61`, `RoleEntry=79`, `RoleDocumentText=94`
- `internal/activate/activate.go:137-149` — factual global-engine reactivation; `gsettings current` dead on GNOME 46
- `cmd/goswitchd/main.go:38` — `-debug` flag semantics
- `test/e2e/case_autocorrect.go:751-906` — the wezterm fixture: `--always-new-process` spawn, negative focus witness, arming-race live finding (2026-10-02), positive control, counters oracle
- `test/e2e/focus_helper.py:620-637` — probe command vocabulary
- `docs/SPEC.md` §4.3/§4.4/§5/§7.2/§11 — spec-delta targets and revision conventions (audit-trail comments, dated owner decisions)
- `docs/adr/d01-experiment-log.md` — the probe-verdict artifact precedent (format + mise-task appends)
- `docs/adr/ADR-006-two-engine-revision.md` — flip mechanism, amendment format, D-52/D-53/D-54 context
- `.planning/WINDOWS.md` rows 4, 13 — GTK4-Wayland waiver; fresh-session flip drift
- `.planning/ROADMAP.md`, `.planning/REQUIREMENTS.md`, `.planning/PROJECT.md` — phase scope, TERM-01/03/04 texts, milestone record (daemon-channel spike 2026-10-08)
- `test/e2e/cases/matrix-v4.yaml` (header) — 34 frozen cases; negative-row discipline

### Secondary (MEDIUM — committed milestone research, web-derived)
- `.planning/research/FEATURES.md` / `STACK.md` / `ARCHITECTURE.md` / `PITFALLS.md` / `SUMMARY.md` — wezterm #1772/#5125/#4615/#6341, zwp_text_input_v3 since 20220807, 20240203 last-tagged, VTE source-verified findings, RHBZ 1847347, gnome-terminal#588, three erase semantics, Telegram AppArmor wall (live-probed 2026-10-08), A11Y trigger matrix design
- wezterm official docs (keyboard concepts, use_ime) — as cited by the committed research

### Tertiary (LOW — deliberately not researched further)
- Web transport lore (unlock-IME bug threads, ghostty#2392) — the milestone research gates all of it behind this live spike; probed, not cited

## Metadata

**Confidence breakdown:**
- Spike protocol & oracle design: **HIGH** — every component rides a live-verified in-repo precedent (d01 log, wezterm fixture, focus helper) or an OS tool verified this session
- wezterm root cause: **LOW until P0–P5 run** — by design; the ladder is falsifiable and covers the refutation branch
- Spec-delta/ADR targets: **HIGH** — section texts read this session; revision conventions have three in-file precedents
- A11Y matrix feasibility: **MEDIUM-HIGH** — mechanisms live-probed 2026-10-08; per-app minimal triggers remain empirical

**Research date:** 2026-10-08
**Valid until:** ~2026-11-07 (stable: pinned local stack; the only churn risk is the owner changing `~/.wezterm.lua` mid-phase)
