# Phase 1: ADR-пакет и каркас IBus-движка - Pattern Map

**Mapped:** 2026-09-10
**Files analyzed:** 22 (new; repo contains zero Go files — all `.gitkeep` placeholders, no `go.mod`)
**Analogs found:** 0 / 22 in-repo code analogs — every file is mapped to a named reference pattern instead (see "Reference Sources" below)

> **Greenfield disclaimer (tracked-source gate).** `git ls-files` shows no `.go` files, no `go.mod`, no CI, no `docs/adr/`. The only git-TRACKED sources with copyable content are `01-RESEARCH.md` (this phase dir) and root `AGENTS.md` (stack decisions). All other references below are machine-local or external, are marked **[ext]**, and are read-only: never edit them, never vendor them into the repo. `sarim/goibus` has NO LICENSE — copy nothing from its repo; the wire field orders you need are already reproduced verbatim inside `01-RESEARCH.md` (use those).

## Reference Sources (the "analogs" for this phase)

| Tag | Source | Tracked? | Use for |
|-----|--------|----------|---------|
| **R** | `.planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-RESEARCH.md` | yes (git) | Wire-struct field orders, connection sequence, key constants, FSM design, generator pipeline, e2e skeleton, logging design, ADR composition — all with ready-to-copy Go excerpts |
| **A** | `AGENTS.md` (repo root) | yes (git) | Stack constraints, What NOT to Use (lines 101-113), daemon loop sketch (lines 82-86), gsettings/ydotool/AT-SPI decisions |
| **P1** | `/home/nil/.local/share/ibus-test/test_shift.py` (77 lines) | **[ext]** owner prototype, on this machine | Minimal IBus registration lifecycle + observer `ProcessKeyEvent` returning `False` — the exact behavioral model of the Phase 1 skeleton |
| **P2** | `/home/nil/.local/share/punto-switcher/punto_engine.py` (438 lines) | **[ext]** owner prototype, on this machine | Tap-FSM prior art (`_handle_shift_release`, `rshift_saw_key`) + one flagged anti-pattern (line 355) |
| **S** | `.zcode/skills/go-ultimate/` (project skill) | **[ext]** project-local skill, untracked | Go conventions: thin `main`/`run(ctx, cfg) error`, test naming, `.gitignore` baseline, CI workflow shape, error/context rules |

Line numbers for **R** refer to `01-RESEARCH.md` as tracked today.

## File Classification

| New/Modified File | Role | Data Flow | Reference Pattern | Match Quality |
|-------------------|------|-----------|-------------------|---------------|
| `go.mod` | config | — | S `references/engineering-policy.md` (apps: latest Go; A: Go ≥ 1.23, godbus only) | none — reference-mapped |
| `.gitignore` | config | — | S `references/repo-and-ci.md` § Ignore baselines | none — reference-mapped |
| `.github/workflows/test.yml` | config (CI) | batch | S `assets/ci/test.yml` (vet+build+test-race shape) | none — reference-mapped |
| `cmd/goswitchd/main.go` | entrypoint (thin main) | event-driven (daemon lifecycle) | S `assets/cli-simple/main.go` lines 17-45; lifecycle from P1 lines 36-74 | none — reference-mapped |
| `engine/address.go` | utility | file-I/O | R § Pattern 1 "Address discovery" (lines 183-188) | none — reference-mapped |
| `engine/conn.go` | service | event-driven (reconnect loop) | R § Pattern 1 "Connection" (190-197) + § Pattern 5 (341-344) | none — reference-mapped |
| `engine/types.go` | model | transform (wire marshalling) | R § Pattern 1 field-order tables (199-261) + § Code Examples (465-486) | none — reference-mapped |
| `engine/factory.go` | controller | request-response | R § Pattern 1 "Factory" (236) | none — reference-mapped |
| `engine/engine.go` | controller | request-response (+signal emission) | R § Pattern 1 engine-method table (238-260) + § Pattern 5 recover-shim (344); P1 lines 25-34; P2 anti-pattern 355 | none — reference-mapped |
| `engine/keys.go` | model (constants) | — | R § Pattern 2 (265-289) | none — reference-mapped |
| `engine/recover_test.go` | test | — | S `references/testing.md` (naming, `package xxx_test`); fake-conn per R Validation table (603) | none — reference-mapped |
| `internal/hotkey/fsm.go` | service (pure logic) | event-driven (state machine) | R § Pattern 3 (293-308) + § Code Examples FSM feed (489-511); P2 prior art 72-75, 219-237, 357-359 | none — reference-mapped |
| `internal/hotkey/fsm_test.go` | test (corpus, fake clock) | — | R § Pattern 3 corpus dimensions (308) | none — reference-mapped |
| `internal/logging/logging.go` | utility | streaming (log stream) | R § Pattern 7 (361-371) + § Code Examples (514-518) | none — reference-mapped |
| `internal/logging/logging_test.go` | test | — | R Validation table row INST-04 (602) | none — reference-mapped |
| `layouts/generator/main.go` | utility (codegen CLI) | file-I/O → transform | R § Pattern 4 pipeline (312-317) | none — reference-mapped |
| `layouts/tables.go` | model (GENERATED, committed) | — | R § Pattern 4 effective-mapping table (319-337) as review baseline | none — reference-mapped |
| `layouts/tables_test.go` | test (golden) | — | R § Pattern 4 golden corpus (337) | none — reference-mapped |
| `test/e2e/main.go` (+ case files) | test harness (CLI runner) | batch orchestration | R § Pattern 6 (348-357) + § Code Examples e2e shell sketch (521-533) | none — reference-mapped |
| `test/e2e/focus_helper.py` | test utility | request-response (AT-SPI) | R § Pattern 6 step 3 (351); A line 69 (python3-gi split, `/usr/bin/python3` mandatory) | none — reference-mapped |
| `docs/adr/ADR-001..005-*.md` | documentation | — | R § Pattern 8 (375-380); CONTEXT D-01..D-06 quoted verbatim | none — reference-mapped |
| systemd dev unit `goswitchd.service` (location at planner's discretion, e.g. `dist/systemd/user/`) | config | — | A lines 80-86; R Runtime State Inventory systemd row (414) | none — reference-mapped |

`cmd/goswitchctl/` stays `.gitkeep` — not in Phase 1 scope (R's recommended structure omits it).

## Pattern Assignments

### `go.mod`, `.gitignore`, `.github/workflows/test.yml` (Wave 0 scaffolding)

**Reference:** S + A. `A` lines 26-27 pin the constraint: Go ≥ 1.23, stdlib + godbus + minimal deps. R line 162: `go 1.23` in go.mod, godbus only at first; R line 86: do NOT add fsnotify in Phase 1.

**`.gitignore` baseline** — copy from S `references/repo-and-ci.md` § Ignore baselines (4 categories: binaries/build output, test/coverage output, tool caches, editors/OS noise):
```gitignore
# Binaries and build output
bin/
dist/
*.exe
*.test

# Test and coverage output
*.coverprofile
coverage.out
coverage.html

# Tool caches
.cache/
.gocache/
.gomodcache/
```

**CI workflow shape** — copy from S `assets/ci/test.yml` (verbatim structure: permissions `contents: read`, concurrency group with event name, `go-version-file: go.mod`, then `go build ./...` / `go vet ./...` / `go test -race -cover ./...`, tidy check). R line 579 overrides the lint part: golangci-lint NOT installed → vet + test -race suffice (research explicitly allows this staging). Drop the `compat` matrix job (application, not library); keep govulncheck only if it doesn't complicate the skeleton.

---

### `cmd/goswitchd/main.go` (entrypoint, event-driven)

**Reference:** S `assets/cli-simple/main.go` lines 17-45 — the go-ultimate thin-main pattern (R line 163 names it explicitly). Copy this shape exactly; swap the greeting body for daemon wiring:

```go
func main() {
	// flags: -debug (key trace, default OFF — security, R line 369)
	flag.Parse()
	// No defer here: os.Exit skips defers, so signals live inside run().
	if err := run(context.Background(), cfg); err != nil {
		fmt.Fprintf(os.Stderr, "goswitchd: %v\n", err)
		os.Exit(1)
	}
}

// run holds all testable logic: signal context, logging init, engine loop.
func run(ctx context.Context, cfg Config) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	...
}
```

**Daemon lifecycle** to hang off `run` — port from P1 `test_shift.py` lines 36-74 (the owner-proven registration sequence this phase reproduces in Go):
```
request_name → factory + add_engine → component + engineDesc → register_component → main loop
```
R line 289 pins the observable order for e2e: `bus connected → component registered → engine __init__ → focus_in`.

---

### `engine/types.go` (model, transform — the load-bearing file)

**Reference:** R § Pattern 1 (lines 199-261) + § Code Examples (465-486). The field order IS the wire contract (godbus marshals exported fields in declaration order). R's adapter skeleton, lines 466-481:

```go
// Source: goibus source read verbatim (component.go/engineDesc.go/bus.go/engine.go);
// field order is the wire contract — do not reorder.
type Component struct {
	Name          string                  // "IBusComponent"
	Attachments   map[string]dbus.Variant // {}
	ComponentName string                  // "org.freedesktop.IBus.goswitch"
	Description   string
	Version       string
	License       string  // "MIT"
	Author        string
	Homepage      string
	Exec          string  // "" — programmatic registration only
	Textdomain    string
	ObservedPaths []dbus.Variant // {}
	EngineList    []dbus.Variant // MakeVariant of each EngineDesc (dereferenced struct value)
}
```

Plus `EngineDesc` (17 fields, order at R lines 216-234 — note `Layout "us"/"ru"` at index 9 drives mutter XKB, `Symbol` at 12 is the indicator label) and `IBusText` (R line 261): `(s "IBusText", a{sv}, s text, v AttrList)` with empty `au` attributes for plain commits. Registration call shape (R lines 483-485):

```go
obj := conn.Object("org.freedesktop.IBus", "/org/freedesktop/IBus")
obj.Go("org.freedesktop.IBus.RegisterComponent", 0, ch, dbus.MakeVariant(component))
```

Method name is `RegisterComponent` (no underscore — verified in libibus strings, R line 56).

---

### `engine/conn.go` (service, event-driven — connect + reconnect/re-register loop)

**Reference:** R lines 190-197 (connection sequence) and § Pattern 5 lines 341-344 (loop). Connection:

```go
conn, err := dbus.Dial(ibusAddress)          // IBUS_ADDRESS=unix:path=...
conn.Auth(GetUserAuth())                      // AuthExternal(uid), AuthCookieSha1 fallback
conn.Hello()
conn.RequestName(IBUS_SERVICE_IBUS /*"org.freedesktop.IBus"*/, dbus.NameFlagReplaceExisting)
```

`RequestName` result doubles as the single-instance guard (AlreadyOwner → exit policy, R line 197). Loop contract (R 343): detect closed `Signal()` channel / `ErrClosed` → **re-run address discovery from scratch** (fresh socket path each daemon generation — 15 stale sockets observed; never cache) → Dial/Auth/Hello → RequestName → re-export Factory → RegisterComponent → log. Backoff ~1-2 s with jitter.

---

### `engine/engine.go` (controller, request-response — per-input-context object)

**Reference:** R engine-method table (lines 238-260): export on THREE interfaces (`org.freedesktop.IBus.Engine`, `...IBus.Service`, `org.freedesktop.DBus.Properties`). Phase 1 behavior: `ProcessKeyEvent` logs and **returns `false` immediately** (observer mode, INTEG-02; anti-pattern "replying late", R line 383). Behavioral model P1 lines 25-34:

```python
def do_process_key_event(self, keyval, keycode, state):
    ...
    is_release = bool(state & (1 << 30))
    log(f"keyval={kname} keycode={keycode} state=0x{state:x} mods=... release={is_release}")
    return False
```

**Recover-shim (INTEG-05)** — wrap EVERY exported handler (R line 344): `defer func(){ if r:=recover(); r!=nil { log error; return safe reply } }()`. godbus dispatches each call on its own goroutine; one panic = process death = desktop-wide input loss.

**Explicit anti-pattern NOT to copy:** P2 `punto_engine.py` line 355 `return True  # consume shift always` — the prototype got away with it; Phase 1 must return `false` for every event (R line 384).

---

### `engine/keys.go` (model, constants)

**Reference:** R § Pattern 2 (lines 265-289) — copy the verified values verbatim: state masks (`IBUS_SHIFT_MASK 1<<0`, `IBUS_CONTROL_MASK 1<<2`, `IBUS_MOD1_MASK 1<<3`, `IBUS_MOD4_MASK 1<<6`, `IBUS_MOD5_MASK 1<<7`, `IBUS_RELEASE_MASK 1<<30`), capability bits, keyvals (`BackSpace 0xff08`, `Shift_R 0xffe2`, ...). The adapter decodes once into `EngineEvent{Keyval, Keycode, Release: state&(1<<30)!=0, Mods: state&0xff}` (R line 289) — the FSM never re-parses raw state.

---

### `engine/factory.go` (controller, request-response)

**Reference:** R line 236: export at `/org/freedesktop/IBus/Factory`, interface `org.freedesktop.IBus.Factory`, method `CreateEngine(s engine_name) → (o object_path)` — mint `/org/freedesktop/IBus/Engine/goswitch/<n>`, export the engine object, return the path. Register both `goswitch-en` and `goswitch-ru` from day one (the D-01 experiment needs both, R line 151-155).

---

### `engine/address.go` (utility, file-I/O)

**Reference:** R lines 183-188: precedence `$IBUS_ADDRESS` env → newest matching bus file `~/.config/ibus/bus/<machine-id>-unix-wayland-0` (select by `WAYLAND_DISPLAY`/`DISPLAY` suffix), parse `IBUS_ADDRESS=unix:path=...,guid=...`. Re-read on every reconnect.

---

### `internal/hotkey/fsm.go` (service, event-driven — pure tap FSM)

**Reference:** R § Pattern 3 (293-308) + FSM feed code example (489-511). Contract: pure function-object, no goroutines, no real clock — `Feed(ev Event, now time.Duration) (consumed bool, actions []Action)`; timers synthetic (`time.AfterFunc(window)` in the adapter re-enters `Feed(TimerExpired{})`). R's skeleton:

```go
type Event interface{} // KeyPress, KeyRelease, TimerExpired, Reset
type Action int        // Single, Double, Triple

func (f *FSM) Feed(ev Event, now time.Duration) []Action {
	switch e := ev.(type) {
	case KeyPress:
		if e.Keyval == KeyvalShiftR { f.held = true; f.sawKey = false }
	case KeyRelease:
		if e.Keyval == KeyvalShiftR && f.held {
			f.held = false
			if f.sawKey { f.taps = 0; return nil } // modifier use — not a tap
			if f.taps < 3 { f.taps++ }              // 4th tap stays at 3
			f.deadline = now + f.window             // adapter arms AfterFunc(window) → TimerExpired
		}
	case TimerExpired:
		if f.taps > 0 { a := Action(f.taps); f.taps = 0; return []Action{a} }
	case Reset: f.taps, f.held, f.sawKey = 0, false, false
	}
	return nil
}
```

**Prior art (modifier-use discrimination)** — P2 `punto_engine.py` lines 219-224:
```python
def _handle_shift_release(self):
    if self.rshift_saw_key:
        # was used as modifier, not a tap
        self.rshift_saw_key = False
        return
```
and the saw-key marking at P2 lines 357-359 (`if not is_release and self.rshift_down_at is not None: self.rshift_saw_key = True`). Prototype constants (P2 73-74, `DOUBLE_TAP_MS=450`, `SINGLE_TAP_MS=550`) are prior art only — D-05 pins 300 ms default. Deltas vs prototype (owner decisions, must hold): all actions fire at window expiry (no instant-single), tap cap at 3 (prototype had single/double only), any intervening key cancels, `FocusOut`/`Reset` disarm (P2 lines 331-338 show the focus-out reset precedent).

---

### `internal/hotkey/fsm_test.go` (test, fake clock)

**Reference:** R § Pattern 3 corpus dimensions (line 308): single/double/triple at window edges (299/300/301 ms), modifier-use, intervening key between taps, 4th tap, FocusOut mid-burst, release-without-press, press held across Reset, ibus#2600-style 1 ms shift glitch, burst of 10 taps. Deterministic: injected timestamps, no wall-clock sleeps (R line 401).

---

### `internal/logging/logging.go` (utility, streaming)

**Reference:** R § Pattern 7 (361-371) — copy verbatim:

```go
level := new(slog.LevelVar)          // runtime-adjustable later via goswitchctl
level.Set(slog.LevelInfo)
h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
slog.SetDefault(slog.New(h))
```

Level contract: ERROR = adapter failures / re-register exhaustion; WARN = reconnects, caps anomalies; INFO = lifecycle (connected, registered, engine created, focus in/out, re-registered); DEBUG = key trace. Key trace behind explicit `-debug` flag, default OFF (R line 369); never log buffer contents at INFO (lengths only). Debug shape (R 516-518): `slog.Debug("key", "keyval", fmt.Sprintf("0x%x", ev.Keyval), "keycode", ev.Keycode, "release", ev.Release, "mods", fmt.Sprintf("0x%x", ev.Mods))`.

---

### `layouts/generator/main.go` + `layouts/tables.go` + `layouts/tables_test.go` (codegen, model, golden test)

**Reference:** R § Pattern 4 (310-339). Pipeline: `go:generate` → `go run ./layouts/generator` → parse `/usr/share/X11/xkb/symbols/us` § basic + `symbols/ru` (default resolves to winkeys on Ubuntu 24.04 — include-recursive, skip `kpdl(comma)`) → join on xkb key name (TLDE, AE01-AB10, BKSL) → keysym values from `/usr/include/ibus-1.0/ibuskeysyms.h` → rune via Latin-1 direct + embedded Cyrillic keysym→Unicode table → emit `layouts/tables.go` (committed; CI tests the committed file, never regenerates). Golden corpus (R line 337): `ghbdtn`→`привет`, `GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет`, plus all punctuation pairs incl. asymmetric `&`↔`?`. R lines 319-335 hold the full effective-mapping review baseline.

---

### `test/e2e/` runner + `focus_helper.py` (test harness)

**Reference:** R § Pattern 6 (348-357) + shell sketch (521-533). Go program (not `go test` — live session, own exit-code contract): (1) preflight fail-fast (address file, `ibus list-engine` contains goswitch, `/dev/uinput` writable, `/usr/bin/python3 -c "import gi"`, log-file heartbeat); (2) sources setup with snapshot/restore of `gsettings ... sources` AND `current` (R line 412, A5); (3) focus via AT-SPI `Component.grabFocus` helper — `/usr/bin/python3` absolute path mandatory (PATH python3 is linuxbrew, no gi); (4) `ydotool type --key-delay` / `ydotool key --key-delay` with configurable pacing; (5) assert on goswitchd JSON log (`wait_for_log '"msg":"action","n":2' 5s` — wait-for-condition, never sleep); (6) resilience cases `ibus-restart`, `kill9-survive`; (7) per-case PASS/FAIL + log excerpt, exit ≠ 0 on FAIL. Cases incl. `-case d01-probe` (switch-matrix probes → evidence for ADR-001).

---

### `docs/adr/ADR-001..005-*.md` (documentation)

**Reference:** R § Pattern 8 (373-380). Format per CONTEXT discretion: `docs/adr/ADR-00N-<slug>.md`, sections Status/Context/Decision/Consequences/Reversibility, quoting CONTEXT decision wording (D-01..D-06). Composition: ADR-001 switching mechanism (written LAST, from D-01 experiment log, incl. kill criteria and losing candidates with observed failures); ADR-002 tap semantics (classic-with-waiting, 300 ms, decision-at-expiry, modifier-use, 4th-tap rule); ADR-003 capability ladder (DeleteSurroundingText → Backspace×N → clipboard); ADR-004 buffer reset triggers (D-06 verbatim + spec-delta §4.3); ADR-005 MACR-01/OPEN-01 owner checkpoint — executor recommendation is input NOT a decision, no code either way.

---

### systemd dev unit `goswitchd.service` (config)

**Reference:** A lines 80-86 (unit at `~/.config/systemd/user/goswitchd.service`; daemon loop "connect → request name → RegisterComponent → retry") + R line 414: `PartOf=graphical-session.target`, `After=org.freedesktop.IBus.session.GNOME.service`; `Restart=on-failure` covers uncatchable `kill -9`.

## Shared Patterns

### Observer-mode return contract (INTEG-02)
**Source:** R lines 44, 62, 384; behavioral proof P1 line 34.
**Apply to:** `engine/engine.go` every method, FSM wiring.
Phase 1 returns `false` from `ProcessKeyEvent` for everything — transit by construction. Never consume bare modifier presses (P2 line 355 is the counter-example).

### Recover-shim + comma-ok (INTEG-05)
**Source:** R line 344; S SKILL.md non-negotiable ("comma-ok on every type assertion").
**Apply to:** all `engine/` exported handlers, every variant assertion anywhere.
`defer func(){ if r:=recover(); r!=nil { slog.Error(...); return safe reply } }()` on each handler; comma-ok (`v, ok := x.(dbus.Variant)`) on every assertion — the #1 panic source.

### Error handling
**Source:** S SKILL.md principle 4 (`fmt.Errorf("...: %w", err)`, `errors.Is/As`, context-first signatures).
**Apply to:** all new Go files. `func F(ctx context.Context, ...) (..., error)`; every goroutine has an exit condition (reconnect loop honors ctx).

### Structured logging contract (INST-04)
**Source:** R § Pattern 7.
**Apply to:** engine, hotkey wiring, e2e assertions (JSON log is the M1 assertion surface).
Levels as assigned in R; key trace opt-in `-debug`; no buffer contents at INFO.

### Decode-once event shape
**Source:** R line 289.
**Apply to:** engine → hotkey boundary.
Adapter decodes release bit and mods once into `EngineEvent`; FSM/buffers never touch raw `state`.

### Test conventions
**Source:** S `references/testing.md` (lines 15, 59-63): vanilla `testing`, external `package xxx_test`, names `TestF_suffixCamelCase`; `-race` mandatory. No testify (A line 61).
**Apply to:** all `*_test.go`. FSM uses injected timestamps; engine recover tests use an in-memory fake connection (R line 603).

## No Analog Found

All 22 files have no in-repo analog (greenfield). Substitutes per the table above: R (tracked research, primary), A (tracked root AGENTS.md), P1/P2 [ext] owner prototypes, S [ext] go-ultimate skill. Two hard rules travel with the substitutes:

| File | Role | Reason no analog / substitute caveat |
|------|------|--------------------------------------|
| all `engine/*.go` | adapter | Only reference is goibus (no LICENSE) — never vendor or copy files from it; use the wire tables already inside 01-RESEARCH.md |
| `test/e2e/focus_helper.py` | test utility | No Python code in repo; ~30-LOC helper per R § Pattern 6 step 3, `/usr/bin/python3` shebang, gir1.2-atspi |

## Metadata

**Analog search scope:** `git ls-files` (58 entries — .planning docs, AGENTS.md, LICENSE, README.md, docs/SPEC.md, five `.gitkeep`); module cache `/home/nil/go/pkg/mod` (no goibus/godbus present); owner prototype dirs `/home/nil/.local/share/{ibus-test,punto-switcher}`; project skill `.zcode/skills/go-ultimate/`.
**Files scanned:** 58 tracked (14 non-cache), 2 prototypes read at cited ranges, 4 skill files/assets read.
**Pattern extraction date:** 2026-09-10
