# Phase 3: Фразы, выделение, переключение и конфигурация - Pattern Map

**Mapped:** 2026-09-15
**Files analyzed:** 18 (8 new, 10 modified)
**Analogs found:** 16 / 18 (fsnotify watcher internals and the AT-SPI appid observer have no in-repo analog — RESEARCH/STACK patterns apply)

All analog paths below are git-TRACKED sources (verified via `git ls-files`); no gitignored mirrors are referenced. Note: `.zcode/skills/go-ultimate/assets/*` Go files exist on disk but are NOT tracked — never use them as analogs.

Phase 3 is an extension of a mature two-phase codebase, so most files are **in-place extensions of themselves** (best possible analog). The genuinely new code with in-repo precedents: `internal/config` (strict-decode precedent lives in the e2e matrix), `internal/ctlsvc` (D-Bus service precedent lives in `engine/conn.go` + `engine/factory.go`), `cmd/goswitchctl` (CLI + D-Bus client precedent lives in `cmd/goswitchd` + `matrix.go:goswitchNameOwner`).

**Important scoping note (layoutset):** RESEARCH "State of the Art" pins «layoutset» = the interface OVER the existing `scriptMode` in `internal/session/actor.go` — SWCH-01 (single tap → flip) is ALREADY IMPLEMENTED (plan 02-04). No new `internal/layoutset` package is required; the analog for the layout toggle is `actor.go:47-53` (scriptMode) + `actor.go:340-349` (flipScript). If the planner chooses to extract an interface, extract it FROM there.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/correct/runs.go` + `runs_test.go` (new) | service (pure domain logic) | transform (range in → converted range out) | `internal/correct/direction.go` (`Detect`) + `convert.go` | exact (same package, same pure-function shape) |
| `internal/correct/buffer.go` (mod) | model (typed-phrase state) | CRUD/transform | itself: `ReplaceToken` `buffer.go:100-111` | exact (in-place extension) |
| `internal/correct/plan.go` (mod) | service | transform | itself: `BuildPlan` `plan.go:43-61` | exact |
| `internal/config/config.go`,`load.go`,`watch.go` + tests (new) | config (schema + strict decode + watcher) | file-I/O + event-driven | `test/e2e/matrix.go:113-136` (strict decode) + `engine/conn.go:132-145` (ctx-driven event loop) | role-match (no config package exists) |
| `internal/hotkey/fsm.go` + `fsm_test.go` (mod) | state machine | event-driven | itself: key routing `fsm.go:84-111` | exact (11 existing tests stay green, D-35) |
| `internal/session/actor.go` + `actor_test.go` (mod) | controller (orchestrator) | event-driven | itself: `ExpiryAt` dispatch `actor.go:253-269`, `startCorrection` `actor.go:413-464` | exact |
| `engine/engine.go` + `engine_test.go` (mod) | adapter (D-Bus transport) | event-driven | itself: `SetSurroundingText` `engine.go:187-197` | exact (seam widening, one line) |
| `engine/keys.go` (mod) | constants | — | itself: keyval block `keys.go:24-37` | exact |
| `internal/ctlsvc/` + tests (new) | service (D-Bus on session bus) | request-response | `engine/conn.go` (RequestName guard `:93-101`) + `engine/factory.go` (Export `:47-56`) + `engine/engine.go:125-133` (recoverHandler) | role-match (second godbus connection) |
| `cmd/goswitchctl/main.go` (new) | CLI client | request-response | `cmd/goswitchd/main.go` (flag/run split) + `test/e2e/matrix.go:388-423` (D-Bus client probe) | role-match |
| `cmd/goswitchd/main.go` (mod) | entrypoint (wiring) | — | itself: `engineConfig` `main.go:48-60` | exact |
| clipboard round-trip helper (`internal/clipboard/` or session helper, new) | utility (subprocess) | subprocess I/O | `test/e2e/main.go:475-488` (`runCmd`) | partial (test-only context; daemon needs ctx + stdin pipe) |
| `internal/appid/` (new, optional — spike-gated per ADR-005) | service (AT-SPI observer on a11y bus) | event-driven | `engine/address.go` (bus address discovery) | partial (no a11y client in repo) |
| `test/e2e/matrix.go` + `matrix_test.go` (mod) | test harness | batch (YAML cases → report) | itself: step schema `matrix.go:90-107`, step dispatch `:611-624` | exact |
| `test/e2e/cases/matrix-v2.yaml` + smoke/spike cases (new) | config/data | batch | `test/e2e/cases/matrix-v1.yaml` + corpus style `matrix_test.go:16-34` | exact |
| `test/e2e/surface.go`,`main.go` (mod) | test harness | — | itself (`injectKeys`/`pressKey`/`waitForLog`) | exact |
| `mise.toml` (mod) | config (task runner) | — | itself: `[tasks.e2e-*]` rows `mise.toml:37-88` | exact |
| `go.mod` (mod) | config | — | itself (godbus/yaml.v3 require lines) | exact |

## Pattern Assignments

### `internal/correct/runs.go` (new — CORR-06 run segmentation, D-22/D-23/D-24)

**Analog:** `internal/correct/direction.go` and `convert.go` — the package's own pure-function idioms. Copy their shape exactly; the package purity contract is written in the package doc (`buffer.go:1-9`): "no D-Bus, no clock, no goroutines — the whole corpus runs under -race without a live session".

**Script-classification pattern** (`direction.go:22-42` — the per-rune scan to copy for run segmentation):
```go
func Detect(token []rune) (Dir, bool) {
	var en, ru int
	for _, r := range token {
		switch {
		case unicode.Is(unicode.Latin, r):
			en++
		case unicode.Is(unicode.Cyrillic, r):
			ru++
		}
	}
	...
}
```
`runs.go` generalizes this scan from token-wide counters to per-rune run boundaries: each maximal run of one script is a segment; digits/punctuation are neutral (D-15 continuation — reuse the `tokenCapable`/`letterCapable` split at `buffer.go:152-168` to decide which runes are "letters" for anchor purposes). The anchor is the script of the LAST letter (D-22); homogeneous ranges convert wholesale (Q1 recommendation / A6 — matrix v1 stays green); no foreign runs after segmentation = success-no-change (D-24), NOT a refusal.

**Conversion reuse** (`convert.go:15-40`): `Convert(token, dir)` is reused AS IS per foreign run — it already passes digits/punctuation through identically and fails the whole conversion on an unmapped letter. Do not fork it.

**Golden-corpus test pattern** (`internal/correct/buffer_test.go:11-23` — named constants + tiny helper, tests in `package correct_test`):
```go
const (
	wordEN       = "ghbdtn"
	wordRU       = "привет"
	...
)
func push(b *correct.Buffer, s string) { for _, r := range s { b.Push(r) } }
```
`runs_test.go` pins the CONTEXT specifics corpus verbatim: `ghbdtnпривет`→`приветпривет`, `gfbпривет`→`паипривет`, `приветghbdtn`→`ghbdtnghbdtn`, `ghbdtn2026привет`→`привет2026привет`, plus D-24 no-op and D-20 empty-input refusal cases. Table-test shape: `buffer_test.go:142-166` (`cases := []struct{name, in, want string}` + `t.Run(tc.name, ...)` with `t.Parallel()`).

---

### `internal/correct/buffer.go` (mod — phrase API, CORR-02/D-25)

**Analog:** itself. The buffer ALREADY stores the whole phrase; the phrase is `b.runes` and only an access/replacement method is missing.

**Range-replacement pattern to copy** (`buffer.go:100-111` — `ReplaceToken`; the phrase variant replaces `[0, len)` instead of `activeRange()`):
```go
func (b *Buffer) ReplaceToken(repl []rune) {
	start, end, ok := b.activeRange()
	if !ok {
		return
	}
	next := make([]rune, 0, len(b.runes)-(end-start)+len(repl))
	next = append(next, b.runes[:start]...)
	next = append(next, repl...)
	next = append(next, b.runes[end:]...)
	b.runes = next
	b.recompute()
}
```
Add alongside it a whole-phrase accessor cloning `b.runes` (the `slices.Clone` idiom of `Token()` at `buffer.go:75-82`). After any destructive mutation call `b.recompute()` (`buffer.go:132-146`) so token coordinates re-derive — this invariant is what keeps the buffer mirroring the field.

---

### `internal/correct/plan.go` (mod — phrase geometry + Backspace cap, D-27)

**Analog:** itself — `BuildPlan` (`plan.go:43-61`). The phrase call site passes the whole phrase as "token" with an empty tail; the ONE behavioral delta is the D-27 cap: `Plan.Backspaces` for a phrase is bounded by the configurable cap (default ~50, from config), and when `n > cap` at level 2 the plan is not built — the actor takes the D-20 silent refusal with a reason.

**Rune-counting discipline** (`plan.go:44`): `n := len(token) + len(tail)` — every count is runes, never bytes (ADR-003; Cyrillic runes are 2 UTF-8 bytes). The existing `#nosec G115` comments carry the phrase-length bound argument — re-evaluate, do not copy blindly, for the phrase case (Pitfall 9: surrounding-text windows are client-limited).

---

### `internal/config/` (new — CONF-01..03, D-31..D-33)

**Analog A — strict YAML decode:** `test/e2e/matrix.go:113-136` is the repo's only yaml.v3 consumer and the exact D-33 pattern:
```go
func loadMatrixCases(data []byte) ([]matrixCase, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)            // unknown key = loud failure, never silent
	var cases []matrixCase
	for i := 1; ; i++ {
		var c matrixCase
		err := dec.Decode(&c)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode case #%d: %w", i, err)
		}
		if err := c.validate(); err != nil {
			return nil, fmt.Errorf("case #%d: %w", i, err)
		}
		...
```
`load.go` = this shape over a single `Config` struct + `cfg.Validate()`.

**Schema/validation pattern** (`matrix.go:39-51` closed vocabularies + `matrix.go:139-167` validate): expose vocabularies as FUNCTIONS not vars ("package-level vars would be mutable globals under the strict lint" — the comment at `matrix.go:37-39` says exactly this); validate mandatory fields, closed value sets, and ranges (window > 0, cap within sane max — ASVS V5 DoS ceilings) with the field named in the error. YAML tags snake_case (`tagliatelle` lint — see matrixStep `yaml:"type"` at `matrix.go:91-94`).

**Analog B — the watcher's ctx-driven event loop:** `engine/conn.go:132-145` (`waitBusLoss`) is the repo's loop-over-channel-with-ctx-cancel idiom to copy in `watch.go`:
```go
func waitBusLoss(ctx context.Context, conn *dbus.Conn) error {
	signals := make(chan *dbus.Signal, signalBufferSize)
	conn.Signal(signals)
	for {
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-signals:
			if !ok {
				return errBusClosed
			}
		}
	}
}
```
Replace the signals channel with `w.Events`/`w.Errors` (fsnotify), filter by `filepath.Base(ev.Name)`, debounce via `time.AfterFunc` (the arm/stop timer discipline of `actor.go:529-535` `armTimer`), then `Load` → ok: publish through `atomic.Pointer[Config]`; err: `slog.Warn("config reload rejected", "error", err)` and keep last-good (D-32). The fsnotify API surface itself has NO in-repo analog — use RESEARCH §3 verbatim (directory-watch, not file-watch: atomic-rename editors swap the inode).

**Dependency gate:** `go.mod` gains `github.com/fsnotify/fsnotify v1.10.1` — the ONLY new dependency of the phase (RESEARCH Installation). Check `.golangci.yml` gomodguard_v2 disable comment still holds (RESEARCH constraints note).

**Test pattern:** decoder tests copy `matrix_test.go:108-165` (`TestMatrixDecode_UnknownFieldRejected` / `MissingRequiredRejected` — typo-key MUST fail loudly). Watcher tests: put the fsnotify dependency behind a small interface so synthetic events drive the debounce logic headlessly (RESEARCH Wave 0 gap note).

---

### `internal/hotkey/fsm.go` (mod — configurable bindings + SWCH-02 combo, D-35/D-36)

**Analog:** itself. Two extensions, both must keep the existing 11 tests of `fsm_test.go` green WITHOUT semantic edits (D-35: the corpus is the executed ADR-002).

**1. Configurable window/binding:** `NewFSM(window)` (`fsm.go:60-63`) already takes the window — configurability flows through the constructor from config; `DefaultWindow` (`fsm.go:8-11`) stays as the documented default (pin `TestFSM_DefaultWindow`, `fsm_test.go:296-302`).

**2. Combo recognition (Pitfall 4):** today a press of Control_R while Shift_R is held hits `keyPress` (`fsm.go:85-98`), sets `sawKey = true` and silently kills the series — the combo is invisible. The recognition seam is exactly here:
```go
func (f *FSM) keyPress(e KeyPress) {
	if e.Keyval == KeyvalShiftR {
		f.held = true
		f.sawKey = false
		return
	}
	if f.held {
		f.sawKey = true    // ← a Control_R press lands here today
		return
	}
	f.taps = 0
}
```
Pattern to follow: extend the event vocabulary (`fsm.go:32-46` — the `KeyPress/KeyRelease/TimerExpired/Reset` sum type) and/or return a new Action kind from `Feed` (`fsm.go:68-81` returns `[]Action`) rather than bolting side effects in — the FSM stays pure and deterministic (its struct doc, `fsm.go:48-51`). The actor-side alternative (recognize combo before FSM, RESEARCH Pitfall 4 "How to avoid") is equally valid — planner picks ONE; either way modifier discrimination for every OTHER key must survive (`TestFSM_ModifierUse`, `fsm_test.go:139-153`, stays green).

New keyval `KeyControlR = 0xffe4` goes in `engine/keys.go` next to `KeyControlL = 0xffe3` (`keys.go:35`) with the header provenance comment style of `keys.go:22-23` (/usr/include/ibus-1.0/ibuskeysyms.h:191 — verified in RESEARCH §6). NEVER hardcode from memory (02-05 class trap).

---

### `internal/session/actor.go` (mod — Triple/selection/combo/MACR/phrase/anchorPos)

**Analog:** itself — the phase's main wiring point; four in-place extensions.

**Action dispatch** (`actor.go:253-269` — where Triple and the combo cut in):
```go
for _, action := range a.fsm.Feed(hotkey.TimerExpired{}, now) {
	slog.Info("action", "n", int(action))
	switch action {
	case hotkey.Double:
		a.startCorrection()
	case hotkey.Single:
		a.flipScript()
	case hotkey.Triple:
		slog.Debug("action deferred", "n", int(action)) // 02-05 phrase  ← CORR-02 cut-in
	}
}
```
NOTE: `slog.Info("action", "n", int(action))` is the EXACT log shape the e2e stand greps (`matrix.go:663`: `fmt.Sprintf(`"msg":"action","n":%d`, n)`) — do not change it. Combo (D-36) slots into the same router as word-then-`flipScript()` (order fixed by SPEC name: correct FIRST, flip SECOND).

**Correction pipeline** (`actor.go:413-464` `startCorrection`): the selection branch (D-30) cuts in at the top — active anchor (anchor ≠ cursor from the last push) → range = selection; else → token as today. Every refusal path copies the one-liner idiom:
```go
slog.Info("correction skipped", "reason", "empty-buffer")   // actor.go:417-418
```
reasons-only at INFO, contents never (D-20/D-21 — extends to clipboard: clipboard content is NEVER logged).

**Ladder execution** (`actor.go:473-485` `executeCorrection`, `:493-501` `executeLevel2`): the clipboard rung (D-28) appends AFTER commit/delete per ADR-003 rung-C ordering — only when config-enabled AND verify-after mismatched. `executeLevel2`'s Backspace loop (`:495-497`) is where the D-27 cap check lands.

**MACR branch (ADR-005, Pattern 6):** goes in `feedKey` (`actor.go:356-398`) ABOVE the mode branches — `ev.Mods&engine.MaskMod4 != 0` (mask constants `keys.go:5-12`; `comboMask` at `actor.go:35` already excludes Mod4 from the buffer) → configured letter set → `consume=true` + `ForwardKeyEvent` Ctrl+letter sequence. The Ctrl+V forward sequence for clipboard paste is validated in the owner prototype (RESEARCH §4).

**Invariant to preserve verbatim** (actor struct doc, `actor.go:69-70`): "The ProcessKeyEvent answer never waits" — all waits live in `time.AfterFunc` timers re-entering the actor under the mutex (`armTimer` `actor.go:529-535`, epoch-tagged stale-timer guards `actor.go:293-302`). Config hot-reload follows the same rule (Pitfall 8: a new window affects series armed AFTER reload; never re-arm a live timer).

**Timer/config snapshot rule:** `NewActor(window)` (`actor.go:116-123`) gains config-snapshot consumption; read `cfg.Load()` once per event, never hold a pointer across the mutex boundary (RESEARCH Pattern 2).

---

### `engine/engine.go` (mod — seam widening for anchorPos, Pitfall 1)

**Analog:** itself. The wire ALREADY decodes anchor; the seam drops it. One-line change plus implementors:

Wire side (`engine.go:187-197`):
```go
func (e *Engine) SetSurroundingText(text dbus.Variant, cursorPos, anchorPos uint32) (err *dbus.Error) {
	defer recoverHandler("SetSurroundingText", &err)
	slog.Debug("surrounding_text", "cursor_pos", cursorPos, "anchor_pos", anchorPos)
	if e.handler != nil {
		if t, ok := decodeIBusText(text); ok {
			e.handler.HandleSurroundingText(t.Text, cursorPos) // ← anchorPos lost; widen to (text, cursorPos, anchorPos)
		}
	}
	return nil
}
```
Seam (`engine.go:82-95`, `EventHandler` interface): widen `HandleSurroundingText(text string, cursorPos uint32)` → `HandleSurroundingText(text string, cursorPos, anchorPos uint32)`. The interface is small (Actor + test doubles) — update all implementors in the SAME task (RESEARCH Pitfall 1 "How to avoid"). The actor's `surr` cache (`actor.go:82`) gains the anchor; `beforeCursor` (`actor.go:565-574`) shows the clamping idiom for the new selection-range math (Pitfall 6: signed offset by anchor/cursor order).

**Recover-shim invariant** (`engine.go:125-133`): every exported D-Bus method starts with `defer recoverHandler(<name>, &err)` — copy for any new exported method (INTEG-05: one panic would kill desktop-wide input).

---

### `internal/ctlsvc/` (new — INST-02, session-bus control service)

**Analog A — name request + single-instance guard:** `engine/conn.go:93-101`:
```go
reply, err := conn.RequestName(cfg.Component.ComponentName, dbus.NameFlagReplaceExisting)
if err != nil {
	return fmt.Errorf("request name %s: %w", cfg.Component.ComponentName, err)
}
if reply != dbus.RequestNameReplyPrimaryOwner {
	slog.Error("another goswitchd instance already registered", "reply", int(reply))
	return fmt.Errorf("%w: got reply %d", errNotPrimaryOwner, int(reply))
}
```
ctlsvc does the same on `dbus.ConnectSessionBus()` with name `org.djarvur.goswitch` (ARCHITECTURE-pinned). Sentinel-error idiom: package-level `var errNotPrimaryOwner = errors.New(...)` (`conn.go:24`) checked with `errors.Is` in the serve loop (`conn.go:47-61`).

**Analog B — method export:** `engine/factory.go:47-56`:
```go
for _, iface := range []string{ifaceEngine, ifaceService, ifaceProps} {
	if exportErr := f.conn.Export(eng, path, iface); exportErr != nil {
		slog.Error("engine export failed", "path", string(path), "error", exportErr)
		return "", &dbus.Error{Name: "org.freedesktop.DBus.Error.Failed", Body: []any{exportErr.Error()}}
	}
}
```
Export the Svc struct once on `org.djarvur.goswitch` at `/org/djarvur/goswitch`; each method (Status/ReloadConfig/CorrectNow) opens with the recover-shim pattern (Analog above; note godbus service methods use `(err *dbus.Error)` named returns, not raw errors — mirror `engine.go`'s handler signatures). In-process calls into the actor go through its public methods under the actor mutex — never a second FSM.

**Wiring point:** `cmd/goswitchd/main.go:48-60` `engineConfig()` is where the ctlsvc joins (it needs the actor + config pointer); the daemon then owns TWO godbus connections (IBus socket via `engine.Run` + session bus via ctlsvc) — start/stop ctlsvc on its own ctx, mirroring `run()`'s signal-context lifecycle (`main.go:33-44`).

---

### `cmd/goswitchctl/main.go` (new — INST-02 CLI client)

**Analog A — CLI skeleton:** `cmd/goswitchd/main.go:20-44` (the repo's only binary; stdlib `flag` only — a CLI framework violates the constraint):
```go
func main() {
	debug := flag.Bool("debug", false, "...")
	flag.Parse()
	// No defer here: os.Exit skips defers, so signal handling lives in run().
	if err := run(context.Background(), *debug); err != nil {
		fmt.Fprintf(os.Stderr, "goswitchd: %v\n", err)
		os.Exit(1)
	}
}
```
Copy: subcommand dispatch on `os.Args` (status|reload|correct), `--json` optional flag, all testable logic in `run()`, exit-code contract non-zero on failure.

**Analog B — D-Bus client call:** `test/e2e/matrix.go:388-423` `goswitchNameOwner` is the in-repo godbus client idiom (connect → object → CallWithContext → Store):
```go
call := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").
	CallWithContext(ctx, "org.freedesktop.DBus.GetNameOwner", 0, goswitchComponentName)
...
if err := call.Store(&owner); err != nil { ... }
```
goswitchctl uses `dbus.ConnectSessionBus()` (simpler than the IBus socket dial — no manual Auth/Hello needed on the session bus) and calls `org.djarvur.goswitch.Status` etc. Error wording when the name has no owner must be human-helpful ("daemon not running?").

---

### Clipboard round-trip helper (new — D-28/D-29)

**Analog:** `test/e2e/main.go:475-488` `runCmd` — the repo's subprocess idiom (context deadline, captured streams, named diagnostics):
```go
func runCmd(ctx context.Context, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String()), nil
}
```
ADAPT for the daemon (this is a test-context analog, not copy-paste): (a) content flows through `cmd.Stdin` — NEVER argv (argv leaks into `ps` and parses as flags; RESEARCH anti-pattern + security table); (b) `wl-paste --no-newline` byte-exact save, empty-buffer = non-zero exit → restore via `wl-copy --clear` (Pitfall 3); (c) `wl-copy` forks and serves in background — its `Run()` returning is normal; (d) restore failure = WARN only (D-29), and clipboard content never enters the log (D-20 extension). Exact flags and the Ctrl+V ForwardKeyEvent sequence: RESEARCH §4 (validated prototype).

---

### `internal/appid/` (new, optional — MACR per-app identity, ADR-005)

**Analog:** `engine/address.go` (bus address discovery — the a11y bus address arrives the same way, via `org.a11y.Bus` on the session bus; godbus dial follows). SPIKE-GATED (RESEARCH A4 + Open Question 3): build only after the Super+letter delivery probe; degradation is WARN "app-identity unavailable" + global rules only. No in-repo event-subscription analog besides `conn.go:132-145` (loop pattern, see config section).

---

### `test/e2e/` (mod+new — matrix v2, selection/combo/reload steps, MACR cases)

**Analog:** itself — the whole stand is the asset (CONTEXT code_context: "матрица v2 надстраивается над этим каркасом").

**Step-schema extension** (`matrix.go:90-95` matrixStep + `:171-203` step validate + `:611-624` runMatrixStep): new step kinds (select/combo/config-reload) each add one field to `matrixStep`, one vocabulary entry, one case in the `runMatrixStep` switch, and one arm in `stepSummary` (`:562-575`) — the "exactly one step kind" counter at `:173-178` must grow with the field list. Key names for selection injection (`ctrl+a`, `shift+Right`) go through the `matrixKeyNames()` canonicalization table (`:67-73`) AFTER a live smoke probe pins which ydotool names resolve correctly (Pitfall 5 — `space`→physical-S class traps).

**Tap gate oracle** (`matrix.go:648-673` `tapMatrix`): gates on `"msg":"action","n":N` and `"msg":"mode"` log shapes — the SWCH-02 combo case gates the same way on whatever new action label the planner picks (keep it a grep-stable slog shape). Per D-34 the layout oracle is the daemon's `mode` log record, NEVER gsettings.

**Case isolation + report contract** (`matrix.go:490-530` `runMatrixCaseIsolated`, `:276-316` `runMatrixFile`): every v2 case inherits the full setupStand→case→teardown cycle and the PASS/FAIL report with exit code — no new machinery needed, just cases. `expect_level`-style per-surface pinning (`:820-829`) is the model for pinning the ACTUAL selection-replacement step per surface (Open Question 2 spike → matrix fields).

**Injection/witness primitives** (`main.go:587-622` `injectText`/`injectKeys`/`pressKey`, `:514-536` `waitForLog`/`waitForNew`): reuse as-is. New smoke/spike cases (select-injection, Super+letter delivery, clipboard round-trip) register in `pickCase` (`main.go:182-205`) exactly like `ladder-chromium` did, and get a mise task row (`mise.toml:77-83` pattern: description notes "live GNOME session; NOT in ci").

**Decoder tests** (`matrix_test.go`): extend the rejection table (`:128-165`) and vocab table (`:208-244`) for every new step kind/field — strict-schema tests are cheap and mandatory (T-02-06-01 lineage).

## Shared Patterns

### Log privacy contract (D-20/D-21 → D-24/D-27/D-28/D-30)
**Source:** `internal/session/actor.go:507-515` (`logCorrectionDone`) + `internal/logging/logging.go:1-11` (level contract doc).
**Apply to:** ALL new paths — phrase, selection, clipboard, MACR, config reload.
```go
slog.Info("correction", "outcome", "done")          // INFO: counters/outcomes only
slog.Debug("correction", "level", int(level), "runes", len(token), "source", string(token), ...)
```
INFO records carry reasons/counters, never content; clipboard content is never logged at ANY level. New log shapes that e2e greps (`action`, `mode`) are contracts — extend, don't rename.

### Silent refusal "abort, не мусорить" (D-20/D-24/D-27/D-30)
**Source:** `internal/session/actor.go:416-437` (every refusal in `startCorrection`).
**Apply to:** phrase/selection/clipboard/MACR refusal paths.
```go
slog.Info("correction skipped", "reason", "verify-mismatch")  // one line, reason slug, nothing touched
```
Exception: D-24 "nothing to convert" is a SUCCESS without changes — outcome `done`, not `skipped`.

### Pure-package purity + dependency direction
**Source:** `internal/correct/plan.go:3-7` ("the pure package must not import the D-Bus-bound engine package").
**Apply to:** `internal/correct/runs.go`, `internal/config` (no D-Bus), `internal/hotkey` (stays clock/goroutine-free — `fsm.go:48-51`).
Engine imports pure packages; never the reverse. `internal/ctlsvc` may import session/config; session never imports ctlsvc.

### Recover shim on every exported D-Bus method (INTEG-05)
**Source:** `engine/engine.go:125-133` (`recoverHandler`).
**Apply to:** every `internal/ctlsvc` exported method. One unrecovered panic kills desktop-wide input.

### Single-instance guard via RequestName (V4 Spoofing)
**Source:** `engine/conn.go:93-101`.
**Apply to:** `internal/ctlsvc` on the session bus — check `RequestNameReplyPrimaryOwner`, sentinel error, `errors.Is` in the caller.

### Structured error wrapping
**Source:** `engine/conn.go:69-117` (every error `fmt.Errorf("verb noun: %w", err)`).
**Apply to:** `internal/config` load/watch errors, clipboard helper, goswitchctl output.

### Strict YAML decode (D-33)
**Source:** `test/e2e/matrix.go:113-116` (`dec.KnownFields(true)`).
**Apply to:** `internal/config/load.go` — plus a rejection unit test per `matrix_test.go:108-121`.

### Mutex-serialized actor + timer re-entry (never block a handler)
**Source:** `internal/session/actor.go:69-72` (struct doc), `:529-535` (`armTimer`), `:293-302` (epoch-tagged stale guards).
**Apply to:** config reload consumption, combo handling, MACR branch, CorrectNow from ctlsvc. The ProcessKeyEvent answer never waits.

### TDD corpus style (D-07)
**Source:** `internal/hotkey/fsm_test.go` (named timing constants `:12-20`, helpers `:23-46`), `internal/correct/buffer_test.go` (named words `:11-16`), `test/e2e/matrix_test.go` (rejection tables).
**Apply to:** ALL new tests — `package xxx_test`, `t.Parallel()`, vanilla `testing` (no testify, convention), table tests with named cases, golden corpora pin decision IDs in comments.

### Green iteration + mise tasks (D-08/D-09)
**Source:** `mise.toml:11-33` (`ci` = build+vet+lint+test), `:37-88` (live e2e tasks pattern).
**Apply to:** every new e2e case gets a mise task row with the "live GNOME session; NOT in ci" note; no Makefile ever.

## No Analog Found

Files/mechanics with no close in-repo match (planner uses RESEARCH.md patterns instead):

| File/Mechanic | Role | Data Flow | Reason |
|---------------|------|-----------|--------|
| `internal/config/watch.go` fsnotify integration | config watcher | event-driven | No fsnotify usage exists in-repo (new dep v1.10.1); RESEARCH §3 gives the directory-watch+debounce skeleton; STACK prescribes the pattern |
| `internal/appid/` AT-SPI observer | service | event-driven | No a11y-bus client in the daemon (python helper is stand-only); spike-gated per ADR-005/RESEARCH A4 |
| Daemon-side subprocess with stdin pipe | utility | file-I/O | `runCmd` analog is test-context; the stdin-pipe + background-fork semantics of wl-copy are new (RESEARCH §4) |
| `atomic.Pointer[Config]` snapshot publication | concurrency | — | No prior hot-reload state in-repo; RESEARCH Pattern 2 defines it |

## Metadata

**Analog search scope:** `engine/`, `internal/{correct,hotkey,session,logging}`, `cmd/goswitchd`, `test/e2e/`, `layouts/`, `mise.toml`, `go.mod`, `.golangci.yml` (tree of 47 tracked files; all named analogs verified via `git ls-files`)
**Files read whole this session:** fsm.go, buffer.go, convert.go, direction.go, plan.go, verify.go, actor.go, engine.go, keys.go, conn.go, factory.go, cmd/goswitchd/main.go, fsm_test.go, buffer_test.go, matrix.go, matrix_test.go, surface.go, test/e2e/main.go, logging.go, mise.toml
**Pattern extraction date:** 2026-09-15
