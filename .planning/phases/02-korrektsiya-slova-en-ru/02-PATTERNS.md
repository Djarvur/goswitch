# Phase 2: Коррекция слова EN↔RU - Pattern Map

**Mapped:** 2026-09-11
**Files analyzed:** 22 (12 new, 10 modified)
**Analogs found:** 20 / 22 (2 data-only files have no in-repo analog)

All analog paths below are git-TRACKED sources (verified via `git ls-files`); no gitignored mirrors are referenced.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/correct/buffer.go` (new) | service (pure domain logic) | transform (runes in → token out) | `internal/hotkey/fsm.go` | exact (pure pkg, no clock/goroutines) |
| `internal/correct/direction.go` (new) | service | transform | `internal/hotkey/fsm.go` | exact |
| `internal/correct/convert.go` (new) | service | transform | `layouts/tables_test.go` (`translateStrict`) + `layouts/tables.go` | exact (consumer of golden tables) |
| `internal/correct/verify.go` (new) | service | transform | `internal/hotkey/fsm.go` | role-match (pure, deterministic) |
| `internal/correct/plan.go` (new) | service | transform | `internal/hotkey/fsm.go` | role-match |
| `internal/correct/*_test.go` (new) | test | batch (golden corpora) | `internal/hotkey/fsm_test.go`, `layouts/tables_test.go` | exact |
| `internal/session/actor.go` (mod) | service (event consumer) | event-driven | itself + timer pattern `internal/session/actor.go:84-108` | exact (in-place extension) |
| `internal/session/actor_test.go` (mod) | test | event-driven | itself (log-capture + injected time) | exact |
| `engine/engine.go` (mod) | adapter (D-Bus transport) | event-driven | itself: `CommitText` emitter `engine/engine.go:310-317` | exact (symmetric emitter) |
| `engine/engine_test.go` (mod) | test | — | itself + `engine/wire_test.go` | exact |
| `engine/factory.go` (possible mod) | adapter | event-driven | itself: `CreateEngine` binds conn/path `engine/factory.go:38-40` | exact (wiring point for emitter sink) |
| `cmd/goswitchd/main.go` (mod) | entrypoint (wiring) | — | itself: `engineConfig` `cmd/goswitchd/main.go:48-60` | exact |
| `test/e2e/matrix.go` (new) | test harness (CLI) | batch (YAML cases → report) | `test/e2e/main.go` | exact (same stand) |
| `test/e2e/surface.go` (new) | test harness (surface drivers) | request-response (subprocess) | `test/e2e/case_m1.go` (zenity driver) | exact |
| `test/e2e/main.go` (mod) | test harness | — | itself (`-case` flag surface) | exact |
| `test/e2e/preflight.go` (mod) | test harness | — | itself (checks table) | exact |
| `test/e2e/cases/*.yaml` (new) | config/data | batch | **none** (RESEARCH Pattern 5) | none |
| `test/e2e/fixtures/input.html` (new) | config/data | — | **none** (trivial static page) | none |
| `mise.toml` (mod) | config | — | itself: `[tasks.e2e-*]` `mise.toml:37-51` | exact |
| `go.mod` (mod) | config | — | itself (godbus require line) | exact |
| `.github/workflows/e2e-matrix.yml` (new) | config (CI workflow) | batch | `.github/workflows/pr-sanity.yml` | exact (same skeleton, self-hosted runner) |
| `docs/` runner instruction (new) | docs | — | `test/e2e/README.md` | role-match |

## Pattern Assignments

### `internal/correct/` package (buffer.go, direction.go, convert.go, verify.go, plan.go + tests)

**Analog:** `internal/hotkey/fsm.go` (155 lines — read whole; this is the project's one existing pure-domain package and the template to copy)

**Package shape pattern** (`internal/hotkey/fsm.go:1-11`) — doc comment cites the decision IDs; constants carry decision references:
```go
// Package hotkey implements the Right Shift tap state machine: the
// classic-with-waiting scheme with a 300 ms disambiguation window whose
// decisions fire only when the window expires after the last tap.
package hotkey

// DefaultWindow is the tap disambiguation window (D-05): ...
const DefaultWindow = 300 * time.Millisecond
```
For `internal/correct`: package doc cites D-13/D-14/D-15/D-16, ADR-003/ADR-004; e.g. `verify.go` cites ADR-004 (обязательная сверка), `plan.go` cites ADR-003 (ladder).

**Purity contract pattern** (`internal/hotkey/fsm.go:48-63`) — no goroutines, no real clock, no channels, constructor takes config:
```go
// FSM is the Right Shift tap state machine: pure and deterministic — no
// goroutines, no real clock, no channels. The adapter feeds key events with
// injected timestamps and re-enters TimerExpired when its window timer fires.
type FSM struct { ... }

func NewFSM(window time.Duration) *FSM {
	return &FSM{window: window}
}
```
`correct.Buffer` follows this exactly: `runes []rune` + token-index state, methods `Push/Backspace/HardReset/Token()`, zero D-Bus imports (RESEARCH "Pattern 2"). All counting in `[]rune`, never bytes (ADR-003, Pitfall 2).

**Method-with-explicit-result pattern** (`internal/hotkey/fsm.go:68-81`) — one entry point returning a value, switch on event kind. `direction.go`/`verify.go` are plain pure functions (RESEARCH Code Examples give the exact bodies: `Direction(token []rune) (dir Direction, ok bool)` and `MatchesSuffix` using `slices.Equal`).

**Conversion pattern** (`layouts/tables_test.go:9-19`) — per-rune table lookup is already proven in-repo; `convert.go` is this plus the `ok`-check:
```go
func translateStrict(s string, table map[rune]rune) string {
	mapped := make([]rune, 0, len(s))
	for _, r := range s {
		mapped = append(mapped, table[r])
	}
	return string(mapped)
}
```
Tables to consume: `layouts.ENToRU` / `layouts.RUToEN` (`layouts/tables.go:13` and `:112`) — both shift levels merged, position-joined, golden-pinned by `layouts/tables_test.go:26-125`. Register preservation is constructive (`'G':'П'` and `'g':'п'` both present — CORR-05).

**Test corpus pattern** (`internal/hotkey/fsm_test.go`) — external test package, named corpus constants, helpers, table-driven subtests:
```go
package hotkey_test   // fsm_test.go:1

const (               // fsm_test.go:12-20 — every case states its contract in the name
	gapJustInside  = 299 * time.Millisecond
	gapAtEdge      = 300 * time.Millisecond
	gapJustOutside = 301 * time.Millisecond
)

func tap(f *hotkey.FSM, t time.Duration) []hotkey.Action { ... } // :23
```
Golden word corpus lives ready-made in `layouts/tables_test.go:58-70` (`ghbdtn`→`привет` in all three registers, both directions) — the `correct` corpora extend it with digits/punctuation-in-token (D-14/D-15), mixed-script rejection (D-16), reset triggers (CORR-09), and tail arithmetic (Pitfall 1). All tests `t.Parallel()`, vanilla stdlib `testing` (no testify — CONVENTIONS).

---

### `internal/session/actor.go` (mod — flip mode, correction wiring, two-phase verify)

**Analog:** itself — the Phase 1 actor is the file being extended; its established patterns must be preserved.

**Mutex-serialization pattern** (`internal/session/actor.go:43-61`) — every handler takes the mutex; godbus dispatches per-call goroutines:
```go
func (a *Actor) HandleKey(ev engine.EngineEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()
	...
}
```

**Timer re-entry pattern** (`internal/session/actor.go:84-108`) — THE pattern for the two-phase verification (Pitfall 4: never block `ProcessKeyEvent`; ~100 ms timeout lives in an `time.AfterFunc`):
```go
// ExpiryAt feeds the FSM a window expiry at the given logical time and logs
// every decision as {"msg":"action","n":N} — the e2e stand greps this exact
// shape. Tests inject the logical time directly (deterministic expiry)...
func (a *Actor) ExpiryAt(now time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.timer = nil
	for _, action := range a.fsm.Feed(hotkey.TimerExpired{}, now) {
		slog.Info("action", "n", int(action))
	}
}

func (a *Actor) armTimer() {           // caller holds the mutex
	if a.timer != nil {
		a.timer.Stop()
	}
	a.timer = time.AfterFunc(a.window, a.Expiry)
}
```
Phase 2 hooks `Double` → start correction pipeline, `Single` → script flip (Open Q1) at exactly this decision site (`actor.go:97-99`); the ~100 ms SetSurroundingText wait reuses `armTimer`-style `time.AfterFunc` with a second deadline.

**Lifecycle reset pattern** (`internal/session/actor.go:66-80`) — FocusOut/Reset disarm and stop the timer; Phase 2 adds `correct.Buffer.HardReset()` to the same cases (CORR-09).

**Wiring-in point** (`cmd/goswitchd/main.go:48-60`) — the actor is constructed here and passed as `engine.Config.Handler`; if the actor now needs an emitter sink, this constructor and `engine.Config` (`engine/conn.go:32-36`) are the seam to change.

**Test pattern for the actor** (`internal/session/actor_test.go:18-69`) — copy for the new integration tests (fake emitter + correction):
```go
const farWindow = time.Hour                // :18 — expiry injected, never fires
const expiryAfterWindow = 2 * farWindow    // :23

type syncBuffer struct { mu sync.Mutex; buf bytes.Buffer } // :27 — guarded log sink
func captureLogs(t *testing.T) *syncBuffer {               // :52
	t.Helper()
	buf := &syncBuffer{}
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, nil)))
	return buf
}
func tapShift(a *session.Actor) { ... }                   // :61
func countActions(buf *syncBuffer) int {                  // :67 — greps the JSON
	return strings.Count(buf.String(), `"msg":"action"`)
}
```
The fake-emitter test double for Phase 2 extends this style: assert emitted calls (CommitText/DeleteSurroundingText args) exactly like `countActions` greps records. Real-timer variant: `TestActor_WindowTimerFiresAutomatically` (`actor_test.go:180-199`) is the template for the async verify phase.

---

### `engine/engine.go` (mod — 3 new emitters + RU-mode consumption)

**Analog:** itself — `CommitText` is the explicit template ("the adapter ships the emitter from day one; correction logic wires it in Phase 2", `engine/engine.go:307-309`).

**Emitter pattern to copy verbatim** (`engine/engine.go:310-317`):
```go
func (e *Engine) CommitText(text IBusText) {
	if e.conn == nil {
		return
	}
	if err := e.conn.Emit(e.path, ifaceEngine+".CommitText", dbus.MakeVariant(text)); err != nil {
		slog.Error("commit text emit failed", "error", err)
	}
}
```
New emitters, same body shape (signatures from RESEARCH, verified against IBus 1.5.29 introspection):
- `func (e *Engine) DeleteSurroundingText(offset int32, nchars uint32)` — `ifaceEngine+".DeleteSurroundingText"`, args `(offset, nchars)`
- `func (e *Engine) ForwardKeyEvent(keyval, keycode, state uint32)` — `ifaceEngine+".ForwardKeyEvent"`
- `func (e *Engine) RequireSurroundingText()` — `ifaceEngine+".RequireSurroundingText"`

**RU-mode consumption point** (`engine/engine.go:128-142`) — `ProcessKeyEvent` currently always returns `false` (observer, INTEG-02); the flip branch (RESEARCH Pattern 3) goes inside this method, keeping the `defer recoverHandler(...)` shim and the DEBUG trace:
```go
func (e *Engine) ProcessKeyEvent(keyval, keycode, state uint32) (handled bool, err *dbus.Error) {
	defer recoverHandler("ProcessKeyEvent", &err)
	ev := decodeEvent(keyval, keycode, state)
	slog.Debug("key", ...)
	if e.handler != nil {
		e.handler.HandleKey(ev)
	}
	return false, nil   // ← becomes conditional in RU mode (commit + return true)
}
```

**Handler-seam extension point** (`engine/engine.go:66-71` + `engine/factory.go:38-40`) — the actor has no engine reference today; engines are minted per input context in `CreateEngine`:
```go
// factory.go:38-40
eng := NewEngine(f.handler, name)
eng.conn = f.conn
eng.path = path
```
If the decision-flow needs an emitter sink in the actor, extend `EventHandler` (same nil-tolerance as `engine.go:137-139` and the `lifecycle` funnel `engine.go:319-324`) or register the sink at engine creation — planner decides; both precedent shapes are in this file pair.

**Constants already present** (`engine/keys.go:24-33` — `KeyBackSpace = 0xff08`, `KeyTab/KeyReturn/KeyEscape` for CORR-09 resets; `:16-20` — `CapSurroundingText = 1 << 5` for ladder level selection; `:5-12` — `MaskShift` etc. for the flip branch's `mods&^MaskShift == 0` guard).

**Emitter test pattern** — `engine/wire_test.go` (external `package engine_test`, pins wire contracts by exact-value assertions, `t.Parallel()`) is the named analog for new emitter wire tests. Constraint to plan around: `e.conn`/`e.path` are unexported and bound only in `factory.CreateEngine`; a detached engine skips emission (`engine.go:311-313`), so emission tests need an in-package seam or a localhost godbus connection (precedent for dialing: `test/e2e/preflight.go:132-159`). Existing handler tests to keep green: `engine/engine_test.go:54-107` (ProcessKeyEvent returns false — this contract CHANGES for RU mode; the table gains flip-mode cases) and `:225-280` (panic containment).

---

### `test/e2e/matrix.go` (new — YAML matrix runner, TEST-04)

**Analog:** `test/e2e/main.go` (the stand it plugs into) + `test/e2e/case_m1.go` (case-body shape)

**Case-registry pattern** (`test/e2e/main.go:165-179`) — matrix entry joins this registry or gets its own `-matrix` flag (flag surface at `main.go:88-93`):
```go
func pickCase(name string) (func(context.Context, *stand) error, error) {
	registry := map[string]func(context.Context, *stand) error{
		"m1-gate":       runM1Gate,
		...
	}
```

**Exit-code + teardown contract** (`test/e2e/main.go:86-135`) — `run()` returns int, deferred teardown ALWAYS runs and `verifyRestored` (`:384-403`) can downgrade PASS→FAIL. The matrix keeps this contract: per-case PASS/FAIL lines, non-zero exit on any FAIL (TEST-04).

**Watchdog pattern** (`test/e2e/main.go:141-161`) — every matrix case runs under `runCaseWatchdog` (`caseTimeout = 180 s`, `main.go:35`): a live-desktop stand must never strand the owner's machine (Pitfall 8).

**Bounded-subprocess pattern** (`test/e2e/main.go:447-460`) — every external call (ydotool, python3 helper, gsettings) goes through `runCmd` with `cmdTimeout`:
```go
func runCmd(ctx context.Context, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	...
}
```

**Wait-for-condition pattern** (`test/e2e/main.go:486-508`) — `waitForLog`/`waitForNew` poll the daemon log (`logPollInterval = 25 ms`); never fixed sleeps. This is exactly the `{tap}` → `waitForLog(`"msg":"action","n":K`)` sequencing (Pitfall 5) and the ladder-level observation from `-debug` log (D-21/ADR-003).

**Injection pattern** (`test/e2e/main.go:559-594`) — `injectText` (ydotool `type --key-delay <pacing>`), `injectKeys` (pacing×3 so multi-taps fit the 300 ms window), `pressKey` (single key: Escape, Enter). Matrix steps `{type}`/`{key}`/`{tap}` map 1:1 onto these.

**Case-body shape** (`test/e2e/case_m1.go:49-74` — `runM1Gate`): activate engine → open surface (defer close) → inject → assert on log. A matrix case is the same skeleton parameterized by YAML steps, with the assertion being text readback instead of log-count.

**YAML loading** — no in-repo analog (`yaml.v3` is a new dependency; `go.mod` gains one require line next to `github.com/godbus/dbus/v5 v5.2.2`, `go.mod:5`). Use strict decode per RESEARCH: `dec := yaml.NewDecoder(r); dec.KnownFields(true)`.

---

### `test/e2e/surface.go` (new — zenity / chromium / gnome-text-editor drivers)

**Analog:** `test/e2e/case_m1.go` zenity driver — move/generalize these functions:

- `surfaceKind` enum (`case_m1.go:36-43`) → grows `surfaceChromium`
- `startZenity` (`:175-186`) — `exec.CommandContext("zenity", "--entry", ...)`, stdout captured to `bytes.Buffer` = delivery oracle
- `waitZenityEntry` (`:190-207`) — polls the AT-SPI witness until the entry owns focus BEFORE any injection (Pitfall 8 gate)
- `closeZenity` (`:211-234`) — Enter + bounded wait, returns stdout (the oracle zenity-close cases need, Pitfall 6)
- `reapZenity` (`:238-247`) — force-kill error path
- `focusWitness` (`:253-260`) — `/usr/bin/python3 <helper> witness` (NEVER PATH python3 — it is linuxbrew without gi)
- Text readback: `focus_helper.py text <app>` (`test/e2e/focus_helper.py:122-136`, `cmd_text`) — already documented as "bridge to Phase 2 TEST-04" (`:11-12`); note the atspi 2.52 `get_text_at_offset` quirk (`:102-119`)

Chromium driver copies the same four-function shape (start with fixture URL `file://.../fixtures/input.html`, wait for focus via witness, read text via `cmd_text("chrome")`, close/reap) — differences are only in process spawn and close semantics.

### `test/e2e/preflight.go` (mod — chromium preflight, mouse self-test)

**Analog:** itself — append to the checks table (`test/e2e/preflight.go:29-49`):
```go
checks := []struct {
	name string
	run  func(context.Context, *stand) error
}{
	{"injection-selftest", checkInjectionSelfTest},
	...
}
for _, check := range checks {
	if err := check.run(ctx, s); err != nil {
		return fmt.Errorf("preflight %s: %w", check.name, err)
	}
	fmt.Printf("preflight %s: ok\n", check.name)
}
```
Each check carries a one-line human-actable diagnostic (style of `checkUinputWritable`, `:80-86`). New checks: chromium/google-chrome launch, ydotool mouse self-test (A8).

---

### `.github/workflows/e2e-matrix.yml` (new — self-hosted GNOME runner, D-19)

**Analog:** `.github/workflows/pr-sanity.yml` — copy the skeleton, change the runner/trigger:
```yaml
name: pr-sanity
on:
  push: { branches: [main] }
  pull_request:
  workflow_dispatch:
permissions:
  contents: read
concurrency:
  group: pr-sanity-${{ github.event_name }}-${{ github.ref }}
  cancel-in-progress: true
jobs:
  sanity:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install mise
        run: curl https://mise.run | sh
      - name: Add mise to PATH
        run: echo "$HOME/.local/bin" >> "$GITHUB_PATH"
      - name: Install tools from mise.toml
        run: mise install
      - name: build
        run: mise run build
```
Deltas per RESEARCH Pattern 6: `runs-on: [self-hosted, gnome]`; trigger `workflow_dispatch` ONLY (fork-PRs must not print on the owner's desktop — Security Domain); `concurrency: {group: e2e-gnome}` (no cancel-in-progress — one run at a time); steps checkout → `mise install` → `mise run e2e-matrix` → upload report artifact. Keep the header comment block style (`pr-sanity.yml:1-13`) explaining why.

### `mise.toml` (mod — e2e-matrix task)

**Analog:** itself, `[tasks.e2e-*]` block (`mise.toml:37-51`) — copy the exact convention:
```toml
[tasks.e2e-m1]
description = "e2e: M1 gate case (live GNOME session; NOT in ci)"
run = "go run ./test/e2e -case m1-gate"
```
New: `[tasks.e2e-matrix]` with the same "(live GNOME session; NOT in ci)" description discipline — local run and CI run the SAME task (D-09).

### `docs/` runner instruction (new)

**Analog:** `test/e2e/README.md` — Russian-language operational doc: canonical mise invocations in a console block, flags table, state-policy section explaining snapshot/restore reasoning. The runner doc covers: actions-runner install in-session, systemd **user** unit `After=graphical-session.target`, env import (`WAYLAND_DISPLAY`, `XDG_RUNTIME_DIR`, `DBUS_SESSION_BUS_ADDRESS`), label `gnome`, runner-group restriction.

## Shared Patterns

### Panic containment (INTEG-05)
**Source:** `engine/engine.go:101-109`
**Apply to:** every exported D-Bus handler touched in Phase 2 (`ProcessKeyEvent` flip branch especially).
```go
func recoverHandler(method string, dbusErr **dbus.Error) {
	if r := recover(); r != nil {
		slog.Error("handler panic contained", "method", method, "panic", fmt.Sprint(r),
			"stack", string(debug.Stack()))
		*dbusErr = nil
	}
}
```
Used as `defer recoverHandler("MethodName", &err)` — see every handler in `engine/engine.go:128-305`.

### Log privacy contract (D-20/D-21)
**Source:** `internal/logging/logging.go:1-11` (level contract doc) + `engine/engine.go:152-160` (`SetSurroundingText` logs cursor positions only, "contents never recorded at INFO")
**Apply to:** all new INFO sites — correction-refusal reasons WITHOUT word contents (D-20); counters at INFO; source→result/ladder-level/latency only under `-debug` (D-21). The default logger is JSON (`logging.Setup`, `logging.go:21-29`) — the e2e stand greps `"msg":"..."` JSON shapes, so new records must keep that shape (precedent: `"msg":"action","n":2`, `internal/session/actor.go:98`).

### Structured error wrapping
**Source:** `test/e2e/main.go:447-460`, `test/e2e/preflight.go:41-46`
**Apply to:** matrix runner and preflight additions — `fmt.Errorf("preflight %s: %w", ...)` / command line + stderr in the message; `errors.Is/As` for sentinel checks (go-ultimate).

### External test packages + vanilla testing
**Source:** every `*_test.go` in the repo (`package hotkey_test`, `package session_test`, `package engine_test`, `package layouts_test`)
**Apply to:** all new test files — `package correct_test`, extensions stay external; no testify; `t.Parallel()` on every test and subtest; helpers as unexported funcs with doc comments.

### Golden corpora pin decisions
**Source:** `layouts/tables_test.go:26-125`
**Apply to:** `internal/correct` corpora — each D-13..D-16 behavior becomes a named table case with the decision ID in a comment (style of `fsm_test.go:12-20` where constants cite D-05).

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `test/e2e/cases/*.yaml` | config/data | batch | No YAML anywhere in the repo (mise.toml is TOML); use RESEARCH Pattern 5 schema verbatim (`name/surface/mode/steps/expect_text`, `---`-separated docs) with `yaml.v3` `KnownFields(true)` |
| `test/e2e/fixtures/input.html` | config/data | — | Trivial static `<input>` page; nothing to mirror — keep it minimal and committed next to the cases |

## Metadata

**Analog search scope:** all git-tracked sources — `engine/`, `internal/{hotkey,session,logging}/`, `layouts/`, `cmd/goswitchd/`, `test/e2e/` (Go + Python + README), `.github/workflows/`, `mise.toml`, `go.mod`
**Files scanned:** 23 tracked source files (all read in full; largest `test/e2e/main.go` at 594 lines)
**Tracked-source gate:** every analog path passed `git ls-files -- <path>` (whole-repo listing inspected); branch `gsd/phase-02-korrektsiya-slova-en-ru`
**Pattern extraction date:** 2026-09-11
