# Architecture Research — goswitch v1.2 «Терминалы и видимость приложений»

**Domain:** IBus engine daemon (Go) — integration research for terminal correction, dynamic tray-menu blocklist item, a11y recipes
**Researched:** 2026-10-08
**Confidence:** HIGH for all claims grounded in the existing codebase (read directly, live-verified by v1.0.0); MEDIUM for VTE/wezterm transport behavior (web findings cross-checked against the owner's 2026-10-08 live spike); LOW for raw web-only claims (flagged inline)

This file OVERWRITES the v1.0.0 architecture research. The v1.0.0 architecture it supersedes is shipped and documented in the code itself (`internal/session/actor.go` carries the authoritative commentary).

---

## Standard Architecture (the system this milestone extends)

### System Overview

```
┌────────────────────────────────────────────────────────────────────────┐
│  cmd/goswitchd (wiring only — composes, never decides)                 │
│    engineConfig(actor) · startCtl · newMenuToggles · configEditor      │
└──────┬──────────────────────┬───────────────────────┬──────────────────┘
       │                      │                       │
┌──────▼───────┐   ┌──────────▼─────────┐   ┌─────────▼────────┐
│ internal/    │   │ internal/ctlsvc    │   │ internal/        │
│ engine       │   │ + goswitchctl      │   │ indicator        │
│ IBus socket  │   │ session-bus service│   │ SNI item +       │
│ wire adapter │   │ status/reload/     │   │ DBusMenu v3      │
└──────┬───────┘   │ correct            │   └─────────┬────────┘
       │           └──────────┬─────────┘             │
┌──────▼──────────────────────▼───────────────────────▼──────────────┐
│ internal/session — Actor (one mutex; the daemon's only brain)       │
│   hotkey FSM → buffer → correct pipeline → emitter bursts           │
│   seams (nil-safe, install-push): MenuSync · ModeDisplay ·          │
│   SoundSink · A11ySink · Switcher · AppidSource · Clipboard         │
│   + NEW v1.2: terminalOp stepper · MenuSync.SetBlocklistTarget      │
└──┬────────┬───────────┬──────────┬───────────┬───────────┬──────────┘
   │        │           │          │           │           │
┌──▼───┐ ┌──▼────┐ ┌────▼────┐ ┌───▼────┐ ┌────▼───┐ ┌─────▼─────┐
│correct│ │hotkey │ │internal/│ │internal│ │internal│ │internal/  │
│(PURE) │ │(PURE) │ │clipboard│ │/appid  │ │/a11y   │ │config     │
│runs,  │ │tap FSM│ │wl-clip  │ │a11y-bus│ │gsettings│ │strict YAML│
│plan,  │ │bindgs │ │Runner   │ │identity│ │+IsEnabled│ │watcher +  │
│verify │ │       │ │         │ │+role   │ │enabler │ │writer     │
└──────┘ └───────┘ └─────────┘ └────────┘ └────────┘ └───────────┘
```

### Component Responsibilities (v1.2 deltas marked)

| Component | Responsibility | v1.2 change |
|-----------|----------------|-------------|
| `internal/session` (Actor) | Serializes all events through one mutex; owns every decision and every emitter burst; off-mutex rungs for subprocess work (`rungMu`) | **Modified — main work.** New terminal-correction rung family + `terminalOp` stepper; terminal gate in the Double/Triple dispatch; `MenuSync` gains `SetBlocklistTarget` |
| `internal/correct` | Pure conversion/planning/verification (no D-Bus, no time) | **Modified — small.** Pure terminal helpers: read-back sanity gate, select-chord geometry; reuses `ConvertRuns`/`BuildPlan` unchanged |
| `internal/clipboard` | wl-clipboard subprocess client: `Runner` seam, 1500 ms deadline, wl-copy fork/no-pipes, stdin-only content | **Modified — small.** Add `ReadPrimary` (`wl-paste --primary`) + primary save/restore; zero new dependencies |
| `internal/indicator` | SNI item + DBusMenu v3; mutex-guarded snapshot; `Callbacks` func fields are the daemon's interactive surface | **Modified.** New dynamic item «Добавить ⟨app⟩ в чёрный список» (id 16); `Callbacks` gains `BlocklistTarget`/`AddToBlocklist`; refresh at `AboutToShow` |
| `internal/config` | Strict YAML, 2 s stat-poll watcher (last-good), Node round-trip writer with temp+rename | **Modified — small.** Add `AppendBlocklistEntry(path, app)` beside `SetAutocorrectEnabled` |
| `internal/appid` | Lazy a11y-bus observer: bridge-namespace identity + live `GetRole` | **Modified — trigger only.** Also started when the menu needs identity (today: only when `macr.apps` non-empty) |
| `internal/engine` | IBus wire adapter: `ProcessKeyEvent` in, `CommitText`/`ForwardKeyEvent`/`DeleteSurroundingText`/`RequireSurroundingText` out; generation seam | **Modified — only if the wezterm defect fix needs it** (spike decides; most likely fixes land in session) |
| `internal/a11y` | `toolkit-accessibility` + `IsEnabled` reconciler with snapshot/restore | **Unchanged (daemon side).** v1.2's A11Y item is recipes/docs + install-side helpers per owner decision 2026-10-08 |
| `internal/hotkey`, `internal/detect`, `internal/layouts`, `internal/sound`, `internal/ctlsvc`, `internal/install`, `internal/activate` | Unchanged responsibilities | Unchanged (ctlsvc `Status` render grows terminal counters; install grows A11Y recipe bits if any) |

---

## The Six Integration Questions

### 1. Where the terminal-correction flow lives: actor orchestration, not a new planner

**Verdict: orchestration in `internal/session`, planning stays pure in `internal/correct`. HIGH confidence.**

The repo already answered this shape twice:

- The one-correction-pipeline rule (`startRangeCorrection` is "the ONE correction pipeline of the daemon, parameterized by its range") lives in the actor; `internal/correct` only computes.
- The clipboard rung (`runClipboardRung`) is the sanctioned pattern for multi-step, subprocess-bearing, time-taking work: **runs OFF the actor mutex** (WR-01 — up to three 1.5 s subprocesses must never stall the keystroke path), serialized by `rungMu.TryLock()` (busy → skip, never block), touches no FSM/buffer state, takes `a.mu` only for counter updates.

The terminal flow is the same animal (forward keys → wait → subprocess read → compute → forward keys → commit), so it is a **terminal rung family in `internal/session`**, gated at the existing dispatch:

- `ExpiryAt` → `hotkey.Double` → **terminal gate** → either today's `startCorrection()` or new `startTerminalCorrection(termWord)`.
- `hotkey.Triple` → terminal gate → `startTerminalCorrection(termLine)` (feasibility-gated feature).

What goes into `internal/correct` (pure, headless-TDD-able, zero new imports):

- read-back sanity gate: the text read from the primary selection must be non-empty, bounded length, and convertible — a pure `TerminalPayload(text []rune) (payload, reason, ok)` wrapping `ConvertRuns` plus terminal-specific refusals (shell prompt characters, empty read).
- replay arithmetic: `correct.BuildPlan(token, nil, converted, 0, backspaceCap)` already computes exactly "N backspaces + one commit of converted" — the level-2 ladder math is **reused verbatim** for the terminal replay (caps forced to 0, so `BuildPlan` always yields the backspace level).
- select-chord geometry table (word → `Shift+Ctrl+Left`, line → `Shift+Home`) as data, so the corpus pins it.

Dependency direction is preserved: `correct` never imports `session`/`engine`/`clipboard` (pure packages never import D-Bus-bound ones — the repo's standing rule).

### 2. Clipboard read/write: extend `internal/clipboard`, never a second runner

**Verdict: `ReadPrimary` + primary save/restore on the existing `Clipboard`; all subprocess discipline is already solved there. HIGH confidence (code-read).**

The existing package already pins everything the terminal flow needs:

- **`Runner` seam** (`func(ctx, name, args, stdin) ([]byte, error)`) — the unit corpus drives a recording fake; production is `exec.CommandContext` under the package deadline (`cmdTimeout` = 1500 ms, T-03-03-05).
- **wl-copy fork/no-pipes precedent** (`execRunner`): wl-copy gets NO stdout/stderr pipes (the forked grandchild serving the clipboard deadlocks a piped `Run` — live finding 2026-09-15); wl-paste captures stdout byte-exactly. `wl-paste --primary` is the same binary shape — same runner, one more flag.
- **Empty-vs-error semantics** (`Save`): genuine non-zero wl-paste exit = empty selection (`had=false`, no error); signal-kill or done-context = real error. `ReadPrimary` must reuse this classification verbatim — misreading "killed" as "empty" would clear/replace a selection the subprocess merely failed to report.
- **Content hygiene**: bytes ride stdin only, never argv; content never logged at any level (D-20/D-21 extension).

Additions (all in `internal/clipboard`, no new package):

| New method | Wire shape | Notes |
|---|---|---|
| `ReadPrimary(ctx)` | `wl-paste --primary --no-newline` | same empty/error split as `Save` |
| `SetPrimary(ctx, b)` | `wl-copy --primary --trim-newline` (stdin) | restore path only |
| `ClearPrimary(ctx)` | `wl-copy --primary --clear` | restore of an originally-empty primary |

**Save/restore symmetry (D-29 carried over):** the forwarded select burst itself overwrites the Wayland primary selection (keyboard selections update primary — Arch Wiki terminal-copy conventions; LOW-to-MEDIUM web confidence, live-confirm in the spike). The rung therefore saves the primary **before** the select burst and restores it best-effort after the commit — a failed restore is a WARN, never an operation error. Note: unlike the selection-correction rung, the terminal flow needs **no `wl-copy` write of the payload at all** — the write-back is backspace-replay + `commit_text`, not a Ctrl+V paste (plain Ctrl+C is SIGINT in a shell; the copy chord `Ctrl+Shift+C` is exactly the trap the primary-read design avoids).

### 3. Selection-key forwarding vs the existing `ForwardKeyEvent` machinery

**Verdict: new four-event forward helpers beside `forwardCtrlV`; time-stepped coordination with epoch guards; do NOT reuse `pendingFix`/`pendingAfter`. HIGH confidence (code-read).**

Proven mechanics to reuse:

- `ForwardKeyEvent` is the daemon's only clean injection channel: the engine emits, ibus-daemon relays on the client's input context, the client IM module synthesizes the key into the app (owner's validated prototype four-event Ctrl+V: `punto_engine.py:301-304`; MACR's Ctrl+letter bursts live-proven). The owner's 2026-10-08 spike confirmed this channel works in wezterm where the external injector fails — «демонский канал свободен от гонки за фокусом».
- **The MACR lesson is binding:** a burst forwarded while a physical modifier is still held arrives polluted (client reads Ctrl+Super — live finding). The terminal rung must forward only from **clean modifier state**; the triggering gestures (Double/Triple Shift) decide at window expiry *after* Shift release, and the rung should assert clean mods at arm.
- New helpers, same shape as `forwardCtrlV` (modifier press bare → key under mask → key release → modifier release):
  - `forwardShiftCtrlLeft(eng)` — word select (4 events);
  - `forwardShiftHome(eng)` — line select (4 events);
  - replay burst: `plan.Backspaces × ForwardKeyEvent(KeyBackSpace, 14, 0)` — byte-identical to `executeLevel2`, which is built but never live-driven: **v1.2 is its first live client**.

What does NOT carry over: `pendingFix`/`pendingAfter` model an **event-answered verify round** (client pushes surrounding text). Terminals never answer: VTE's `im_retrieve_surrounding()` is a stub returning false behind a false-advertised `IBUS_CAP_SURROUNDING_TEXT` (Red Hat Bugzilla 1847347 → GNOME vte#214, unfixed; MEDIUM confidence — corroborated by the repo's own observation that clients push spontaneously and VTE never does), and wezterm honestly reports caps `0x9` (no surrounding bit — owner-observed). The terminal flow is therefore **time-stepped**, with the read-back as its own verification:

```
arm (under a.mu, epoch++)
  → forward select burst          (fire-and-forget)
  → settle delay (AfterFunc; config knob, spike-pinned ~50–100 ms)
  → ReadPrimary (off-mutex goroutine)
  → pure payload gate + ConvertRuns + BuildPlan
  → replay Backspaces ×N + CommitText   ← DESTRUCTIVE POINT, one way
  → (optional verify-after) re-select + re-read primary, suffix-compare
```

**Verify-after in terminals:** reuse the *pattern* of `armAfterVerify` (epoch tag, deadline timer, mismatch ⇒ INFO + never auto-repair — «abort, не мусорить») but as the terminal stepper's last phase, not the `pendingAfter` type. A mismatched re-read logs `terminal-verify-mismatch` and stops; a clipboard-rung repair fallback is explicitly out of v1.2 scope.

### 4. State-machine additions for the multi-step flow (failure/rollback)

**Verdict: a dedicated `terminalOp` state with an explicit safe point; aborts are free only before the destructive point. HIGH confidence (design follows the actor's own invariants).**

```go
// internal/session — one open terminal op; nil = idle.
type terminalOp struct {
    kind      termScope        // termWord | termLine
    phase     termPhase        // armed→selected→read→planned→replayed→verified
    epoch     uint64           // stale-timer guard (verifyEpoch precedent)
    converted []rune           // the payload (memory-only; never logged)
    plan      correct.Plan
    saved     []byte; hadPrimary bool
    deadline  *time.Timer      // per-phase budget (per-phase slugs below)
}
```

Rules the spec-delta must pin (each maps to an existing actor invariant):

1. **Safe point.** Everything before the first `ForwardKeyEvent(BackSpace)` mutates nothing in the field — abort is free. Failure slugs: `terminal-busy` (TryLock-style re-entry guard), `terminal-read-empty`, `terminal-refused-<correct reason>`, `terminal-interrupted` (printable key / FocusOut / Reset arrived mid-flow — abort, restore primary), `terminal-appid-unavailable` (if the gate needs identity and it is unknown — policy per spec).
2. **Past the safe point there is no rollback.** The replay burst is wire-synchronous (same stance as `executeLevel2`: «the burst→commit order is the wire contract»); a commit failure after replay is one WARN + counter, never a second destructive pass. This is the level-2 degradation doctrine applied to a new transport.
3. **Focus/context death.** `HandleLifecycle` (FocusOut/Reset) cancels the op like `resolvePending`+`clearAfter` — but if phase ≥ replayed, let the burst finish; the context that died receives at most the correction it asked for. The flipCredit machinery (own-flip synthetic lifecycle) is untouched — the terminal flow does not flip engines.
4. **Re-entry.** A second Double while an op is open ⇒ `terminal-busy` skip (the `rungMu.TryLock` precedent), never interleaving.
5. **Settlement effects copied from the word path:** `buf.HardReset()` after commit (the selection-path precedent: the buffer can no longer mirror the field), `settleCorrectionFlip(scriptOf(converted))` (owner rule: mode follows the corrected script), `settleCombo()` strictly after the settled record (D-36 order).
6. **Counters/status:** `corrDone`/`skipReasons` extend with the terminal slugs; `Status` grows a terminal breakdown (counts only — D-20/D-21). `goswitchctl status` and the menu's Status row render them.
7. **Latency budget honesty:** SPEC §5's <50 ms is the *reaction* budget (hotkey → first wire action). The terminal flow's select→read leg is subprocess-bound (~150–400 ms realistic). The spec-delta must state: reaction <50 ms (select burst leaves inside it), completion best-effort with the per-phase deadlines as wedge guards — otherwise the acceptance matrix will «fail» a healthy design.

**Gate design (which path decides «this is a terminal»):** caps are untrustworthy (VTE lies; wezterm's honest `0x9` differs). The workable discriminators, in order of preference:

1. **Identity/role** via `internal/appid`: `Role(ctx) == RoleTerminal (60)` or the focused identity in a configured terminal-apps list (wezterm, `org.gnome.Terminal`). Fail-closed default when unknown: standard path (today's behavior — verify-timeout skip — is already the safe outcome).
2. **Fallback escalation** (spec decision): a Double whose standard verify round times out *and* `correction.terminal: on` may re-arm as a terminal op instead of silently skipping. This turns today's dead gesture in terminals into the feature — at the cost of ~one verify budget of latency and a new failure surface. Recommend ON for word, gated by spike evidence for line.

Note the autocorrect role gate stays as-is: role 60 never autocorrects in terminals (owner: автокоррекция в терминалах НЕ делается) — only the manual gesture path bypasses the gate.

**wezterm layout-switch defect:** treat as a spike-first defect with three candidate hypotheses, each with a journal-observable probe (owner's `-debug` trace + `ibus engine`/journal): (h1) commits/keys race mutter's `zwp_text_input_v3` arming at focus (wezterm#1772 class; the GNOME GitLab «IME dead after unlock» report is the same family — LOW web confidence, live-confirm); (h2) engine re-mint churn on flip loses the armed state (the flipCredit machinery's cousin); (h3) RU-mode consumption happens but `commit_text` doesn't land on the bridge. Candidate fixes range from a wezterm-scoped retry/re-arm in session to documentation-plus-config workaround. **No fix is designed before the spike's verdict table** — the defect phase exists to produce it.

### 5. Menu «Добавить ⟨app⟩ в чёрный список»: identity in, config file append, hot-reload out

**Verdict: a third `Callbacks`-driven toggle-composition; refresh the label at `AboutToShow`; write through a new `config.AppendBlocklistEntry` on the existing atomic-rename writer. HIGH confidence (all precedents in-repo).**

Data flow:

```
[menu open] indicator.Menu.AboutToShow(id)
    → cb.BlocklistTarget()  (actor: appid.FocusedApp(); "" ⇒ item disabled)
    → m.SetBlocklistTarget(app)   (label «Добавить ⟨app⟩ в чёрный список»; LayoutUpdated if visibility changed)
[click] Menu.Event → dispatchClick(menuIDBlocklist)
    → cb.AddToBlocklist(app)      (daemon wiring composition, own mutex)
        1. empty-path guard (WR-03 precedent)          → one WARN, nothing changes
        2. EnsureDocument(path)
        3. config.AppendBlocklistEntry(path, app):
           load Node → autocorrect.apps_blocklist → append regexp.QuoteMeta(app)
           (identities contain dots — raw dots are regex any-char; QuoteMeta IS the
            owner's "substring regex" semantics applied to a literal identity),
           dedupe by exact string, cap 64 (maxAutocorrectBlocklist), validate
           → rewriteAtomically: encode 2-space → temp ".config.*.tmp" (dot-prefix
             ignored by the watcher) → chmod 0600 → os.Rename  (POSIX-atomic)
        4. push menu state; synchronous reload + actor.FoldAppliedConfig()
           (the configToggle.flip order — the apply is in force in the same click;
            the 200 ms stat-poll echo re-applies the same values harmlessly)
        5. the fold recompiles acBlocklist (refreshACBlocklist) — the gate is live
```

Component specifics:

- **`internal/indicator`:** new item id `menuIDBlocklist = 16` (after Reload); `Callbacks` gains `BlocklistTarget func() string` and `AddToBlocklist func(app string)` — nil-safe no-ops, the established Callbacks discipline (the menu never imports appid/session; callbacks run on godbus dispatch goroutines and own their synchronization). The dynamic label is a property update (`ItemsPropertiesUpdated`); hidden↔visible flips bump the revision and emit `LayoutUpdated` — the existing machinery already covers both.
- **Identity freshness:** the actor's appid cache holds the **last-focused** bridge-namespace identity (`FocusedApp()`), which is exactly right for a tray action («the app I was just in»). `AboutToShow` is the natural refresh point — the shell calls it on every menu open. This requires appid to be started for menu reasons, not only MACR: change the lazy-start trigger to «menu attached OR macr.apps non-empty» (start-at-menu-attach is simplest; the a11y bus is normally up because `a11y.enabled` defaults ON). When appid is unavailable, `BlocklistTarget` answers "" and the item renders disabled — the autocorrect gate's behavior never changes with the menu.
- **`internal/config`:** `AppendBlocklistEntry` sits beside `SetAutocorrectEnabled`/`SetSoundEnabled` in `writer.go`, reusing `EnsureDocument`, the Node round-trip (comments preserved), `rewriteAtomically`, the validation ceiling (64) and the regex/empty checks from `validate`. Concurrent clicks serialize on a small mutex in the wiring composition (toggles are idempotent and skipped this; an append+rename is worth one mutex).
- **Hot-reload fold:** nothing new — the watcher's stat-poll sees the rename; `FoldAppliedConfig`/`applySnapshot` recompile the blocklist (existing `refreshACBlocklist` + `compileACBlocklist`). The one new fold interaction: after a blocklist change the actor re-pushes `SetBlocklistTarget` so the menu can disable the item for the now-blocked app.

### 6. Suggested build order (spec-delta first, spike first)

**Verdict: five phases; the wezterm spike is Phase 1 and its evidence gates the terminal-correction design. HIGH confidence.**

D-55 is absolute: **spec-delta (SPEC §4/§6/§9 amendments: TERM scope — word mandatory, line feasibility-gated, автокоррекция в терминалах off; MENU item; A11Y recipes; latency-budget completion clause) and a new ADR-008 (terminal transport) precede all code.** ADR-008 starts Proposed, is driven to Accepted by the Phase 1 spike's evidence table.

| Phase | Content | Depends on | Key risk retired |
|---|---|---|---|
| **1. Спайк + дефект wezterm** | Live spike per terminal (wezterm, gnome-terminal): pin caps, `ProcessKeyEvent` arrival, `commit_text` landing, `ForwardKeyEvent` select-burst → primary read-back, a11y identity presence; verdict table → ADR-008 Accepted → **fix layout switch in wezterm** (+ e2e rows for wezterm) | spec-delta | The owner's daily defect; whether `commit_text` or per-key forwarded replay is the terminal transport; the gate's discriminators |
| **2. TERM-слово (обязательный)** | `clipboard.ReadPrimary` + primary save/restore; pure terminal helpers in `correct`; `terminalOp` stepper + failure vocabulary; gate (identity-first, verify-timeout fallback); config knobs (`correction.terminal`, per-phase `timeouts`); e2e matrix rows (wezterm + VTE) | Phase 1 evidence | The never-live-driven level-2 replay gets its first real client under full instrumentation |
| **3. MENU — пункт чёрного списка** | `config.AppendBlocklistEntry` (QuoteMeta, dedupe, cap, atomic rename); indicator item 16 + callbacks; `MenuSync.SetBlocklistTarget`; appid start-on-menu-attach; e2e via DBusMenu Event | independent of 2 | Config-file mutation from a click without breaking hot-reload |
| **4. TERM-строка (если возможно)** | `forwardShiftHome` line select on the proven word rung; feasibility gate from Phase 1 read-back quality (prompt glyphs etc.); possibly scoped out by the owner at the phase gate | Phase 2 | Multi-word terminal corrections |
| **5. A11Y — рецепты Telegram/ZCode** | Docs/recipes + install-side helpers; daemon unchanged (owner decision 2026-10-08); check-in gate: «если Telegram не поддастся — вернуться к вопросу настройки демона» | independent | Can run in parallel with 2–4 |

Ordering rationale: the defect phase is first because it is the owner's daily pain **and** produces the transport evidence every later terminal phase consumes; MENU (3) is small, fully precedented, and independent — it slots after the mandatory TERM-word or swaps with it if terminal evidence slips; A11Y (5) is daemon-untouched and parallelizable.

---

## Recommended Project Structure (deltas only)

```
internal/
├── correct/
│   ├── terminal.go          # NEW pure: TerminalPayload gate, select-chord geometry table
│   └── terminal_test.go
├── clipboard/
│   └── clipboard.go         # MODIFIED: ReadPrimary/SetPrimary/ClearPrimary (+ tests)
├── session/
│   ├── terminal.go          # NEW: terminalOp stepper, forwardShiftHome/ShiftCtrlLeft,
│   │                        #      gate, slugs, counters (off-mutex rung discipline)
│   ├── terminal_test.go
│   └── actor.go             # MODIFIED: gate in ExpiryAt dispatch; MenuSync +=
│                            #      SetBlocklistTarget; Status += terminal counters
├── config/
│   └── writer.go            # MODIFIED: AppendBlocklistEntry (+ tests)
├── indicator/
│   ├── menu.go              # MODIFIED: menuIDBlocklist=16, dynamic label, AboutToShow refresh
│   └── indicator.go         # MODIFIED: Callbacks.BlocklistTarget/AddToBlocklist
├── appid/                   # code unchanged — the start-trigger change lives in session/cmd
└── engine/                  # touch ONLY if the Phase 1 verdict demands it
```

**Structure rationale:** every v1.2 change lands either in an existing file beside its precedent or in a new file inside the package that owns the concern; no new packages; the pure/D-Bus boundary (`correct`, `hotkey` pure — everything they serve stays out) is untouched, which keeps the golangci depguard architecture and the headless corpus green from task one.

## Architectural Patterns (the three this milestone leans on)

### Pattern 1: The off-mutex rung (WR-01) — extended to a time-stepped stepper

**What:** long, subprocess-bearing work runs outside `a.mu`, serialized by its own mutex, communicating with actor state only through brief re-acquisitions and atomic snapshots.
**When:** any flow with >1 ms of blocking I/O in the correction path.
**Trade-offs:** state must be snapshotted at arm (`eng`, `clip`, payload); cancellation needs epoch-tagged timers.
**Example:** `runClipboardRung` (actor.go:1659) is the template; the terminal stepper adds per-phase `AfterFunc` deadlines with the `verifyEpoch` stale-guard idiom.

### Pattern 2: The config-toggle composition — generalized to any persisted menu action

**What:** click = guard (empty path) → pure writer function on the Node tree → atomic rename → menu push → synchronous reload + `FoldAppliedConfig`.
**When:** any menu item that must outlive the click in the user's config.
**Trade-offs:** the whole document re-encodes (comments preserved, formatting normalized); a mutex is needed once actions are non-idempotent (append).
**Example:** `configToggle.flip` (cmd/goswitchd/main.go:351) + `writeAtomically` (config/writer.go:326).

### Pattern 3: Callbacks-in, pushes-out (the seam discipline)

**What:** leaf D-Bus packages (`indicator`) receive func-field callbacks and expose `SetX` push methods; `cmd` wiring composes; the actor holds nil-safe seams with install-push self-sync.
**When:** adding any interactive surface.
**Trade-offs:** one more interface method per datum (`MenuSync` grows); never an import from leaf to brain.
**Example:** `MenuSync`/`SetMenuSync` install-push (actor.go:1160, 914); the blocklist item adds exactly one datum (target label) and one callback pair.

## Data Flow

### Terminal-correction flow (new)

```
Double Shift (in wezterm/VTE)
    ↓ ExpiryAt → gate (appid role/list; else verify-timeout fallback)
startTerminalCorrection(termWord)   [a.mu: op armed, epoch++]
    ↓ off-mutex stepper
ForwardKeyEvent Shift+Ctrl+Left ×4   → terminal selects the word (primary updates)
    ↓ settle delay (AfterFunc, config)
clip.ReadPrimary()                   → wl-paste --primary (1500 ms budget)
    ↓ pure gate: correct.TerminalPayload → ConvertRuns → BuildPlan(caps=0)
ForwardKeyEvent BackSpace ×N; CommitText(converted)   ← destructive, one-way
    ↓ restore primary (best-effort WARN) → optional re-select+re-read verify-after
buf.HardReset · settleCorrectionFlip · counters/menu pushes
```

### Blocklist flow (new)

```
menu open → AboutToShow → BlocklistTarget() → label push
click → AddToBlocklist(app) → AppendBlocklistEntry (QuoteMeta, atomic rename)
      → reload + FoldAppliedConfig → acBlocklist recompiled → item disables
```

## Scaling Considerations

| Scale | Adjustment |
|-------|-----------|
| 2 terminals (v1.2) | Hard-coded select-chord geometry + app list in config; nothing more |
| More terminals (kitty, alacritty…) | The app list + per-app quirk table grows — keep it config data, never code branches; the Ghostty ForwardKeyEvent segfault (ghostty#2392, LOW web confidence) says per-terminal verification stays mandatory |
| Other layout pairs (LAYOUT-01) | `correct`'s tables and the terminal payload gate are pair-agnostic already — no v1.2 debt |

**First bottleneck (predictable):** primary-selection timing on slow terminals (settle-delay tuning) — mitigate with the per-phase deadline + verify-after, not with longer blind sleeps.

## Anti-Patterns

### Anti-Pattern 1: Trusting capability bits in terminals
**What people do:** branch the ladder on `CapSurroundingText`.
**Why it's wrong:** VTE advertises surrounding text and stubs it out (RHBZ 1847347) — a level-1 ladder would `DeleteSurroundingText` (no-op) + commit ⇒ duplicated line; a stale cross-app surrounding leak (the bug's Firefox example) could even suffix-match by coincidence.
**Do this instead:** identity/role gate first; never execute the delete+commit ladder in a terminal-gated context; treat the read-back as the only terminal truth.

### Anti-Pattern 2: A second clipboard runner / a second config writer
**What people do:** shell out to `wl-paste` ad hoc in session code, or write the config with `os.WriteFile`.
**Why it's wrong:** loses the deadline, the empty-vs-error split, the no-pipes wl-copy exception, the atomic-rename/watcher interplay — each a live-learned fix.
**Do this instead:** `clipboard.ReadPrimary` on the existing `Runner`; `AppendBlocklistEntry` on the existing Node+rename writer.

### Anti-Pattern 3: Reusing verify-round types for a time-stepped flow
**What people do:** shove the terminal flow into `pendingFix`/`pendingAfter`.
**Why it's wrong:** those types model event-answered verification; terminals never answer — the flow would wedge on timers that only settle on pushes that never come.
**Do this instead:** the dedicated `terminalOp` stepper with its own phases, deadlines and safe-point rule.

### Anti-Pattern 4: Designing the wezterm fix from the issue tracker
**What people do:** pick a fix from wezterm#1772/GNOME GitLab threads.
**Why it's wrong:** the failure is a mutter↔wezterm↔ibus bridge race observed on one desktop; upstream threads describe other manifestations (unlock, commit-string bugs) — LOW-confidence material.
**Do this instead:** the Phase 1 spike produces a verdict table on the owner's desktop; the fix (if in goswitch at all) cites it; upstream-linkable findings go back to the trackers.

## Integration Points

### External Services / System Surfaces

| Surface | Integration pattern | Gotchas |
|---|---|---|
| Wayland primary selection | `wl-paste --primary` / `wl-copy --primary` via existing Runner | select keys overwrite it — save/restore; asynchronous update after select (settle delay) |
| Terminal key semantics | Forwarded chords only via engine `ForwardKeyEvent`; Ctrl+C = SIGINT — never forward it as «copy» | VTE crash class with IM backspace sequences (gnome-terminal SIGSEGV under ibus-keyman, Launchpad 2024) — the replay burst is plain BackSpace, the safest shape |
| wezterm IME path | mutter `zwp_text_input_v3` bridge, caps `0x9` | commit/arming races are bridge state, not engine state; spike before fix |
| Config file (user-owned) | Node round-trip + dot-prefix temp + rename; watcher stat-poll + synchronous fold | QuoteMeta identities; cap 64; concurrent clicks need one mutex |
| a11y bus (appid) | existing observer; new start trigger (menu attach) | silent a11y (wezterm may not be in the tree) ⇒ identity "" ⇒ disabled item / fail-closed gate |

### Internal Boundaries

| Boundary | Communication | Notes |
|---|---|---|
| session ↔ correct | pure function calls only | terminal helpers stay pure; no time/D-Bus imports (depguard) |
| session ↔ clipboard | `Save/Set/Clear/Restore` + new `ReadPrimary/SetPrimary/ClearPrimary` | Runner seam for tests; content never logged |
| indicator ↔ session/cmd | `Callbacks` in, `SetX` pushes out | new: `BlocklistTarget`/`AddToBlocklist`/`SetBlocklistTarget` |
| cmd ↔ config | pure writer funcs + reload/fold composition | the configToggle order is the contract |

## Sources

- **Codebase (HIGH — read directly, live-proven in v1.0.0):** `internal/session/actor.go` (rung discipline WR-01, forwardCtrlV, MACR clean-modifier lesson, pending*/verify-epoch, flipCredit, gate sites, Status), `internal/clipboard/clipboard.go` (Runner, 1500 ms, no-pipes wl-copy, empty-vs-error), `internal/config/writer.go` (rewriteAtomically/writeAtomically, temp dot-prefix), `cmd/goswitchd/main.go` (configToggle.flip order, Callbacks composition, editor no-pipes precedent), `internal/indicator/menu.go` + `indicator.go` (item ids, revision/ItemsPropertiesUpdated, AboutToShow, Callbacks), `internal/appid/appid.go` (bridge identity, RoleTerminal=60, lazy start), `internal/correct/plan.go`/`verify.go` (BuildPlan level-2 math), `internal/config/config.go` (blocklist validation, cap 64)
- **Owner live spike 2026-10-08 (HIGH for this desktop):** external injector fails, daemon ForwardKeyEvent channel works in wezterm; wezterm caps `0x9` (PROJECT.md milestone record)
- Red Hat Bugzilla [1847347 — VTE surrounding-text false advertising](https://bugzilla.redhat.com/show_bug.cgi?id=1847347) / GNOME vte#214 — **MEDIUM** (upstream bug, corroborated by repo observations)
- [wezterm use_ime docs](https://wezterm.org/config/lua/config/use_ime), [wezterm#1772 text-input-v3](https://github.com/wezterm/wezterm/issues/1772), GNOME GitLab IME-after-unlock report — **LOW-MEDIUM** (web; live-confirm in Phase 1)
- [ibus-bamboo mode ladder (Pre-edit/Surrounding/ForwardKeyEvent/XTest)](https://github.com/BambooEngine/ibus-bamboo/blob/master/README_EN.md), Arch BBS ForwardKeyEvent-in-GNOME-Wayland thread, [ghostty#2392 ForwardKeyEvent segfault](https://github.com/ghostty-org/ghostty/issues/2392) — **LOW-MEDIUM** (production precedent; per-terminal verification mandatory)
- [Arch Wiki — copying text from a terminal](https://wiki.archlinux.org/title/Copying_text_from_a_terminal) (primary vs clipboard; Ctrl+Shift+C) — **LOW-MEDIUM**

---
*Architecture research for: goswitch v1.2 «Терминалы и видимость приложений»*
*Researched: 2026-10-08*
