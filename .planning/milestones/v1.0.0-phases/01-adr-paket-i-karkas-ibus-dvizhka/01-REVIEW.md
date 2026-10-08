---
phase: 01-adr-paket-i-karkas-ibus-dvizhka
reviewed: 2026-09-11T08:35:43Z
depth: standard
files_reviewed: 32
files_reviewed_list:
  - cmd/goswitchd/main.go
  - engine/address.go
  - engine/conn.go
  - engine/engine.go
  - engine/factory.go
  - engine/keys.go
  - engine/types.go
  - engine/address_test.go
  - engine/engine_test.go
  - engine/wire_test.go
  - internal/hotkey/fsm.go
  - internal/hotkey/fsm_test.go
  - internal/logging/logging.go
  - internal/logging/logging_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - layouts/generator/main.go
  - layouts/tables.go
  - layouts/tables_test.go
  - test/e2e/main.go
  - test/e2e/preflight.go
  - test/e2e/case_m1.go
  - test/e2e/case_resilience.go
  - test/e2e/case_d01.go
  - test/e2e/focus_helper.py
  - .golangci.yml
  - mise.toml
  - go.mod
  - .github/workflows/pr-sanity.yml
  - .github/workflows/security-scheduled.yml
  - .github/dependabot.yml
  - dist/systemd/user/goswitchd.service
findings:
  critical: 0
  warning: 3
  info: 11
  total: 14
status: issues_found
---

# Phase 01: Code Review Report

**Reviewed:** 2026-09-11T08:35:43Z
**Depth:** standard
**Files Reviewed:** 32
**Status:** issues_found

## Summary

Full standard-depth pass over the Phase 1 deliverable: daemon entry point, IBus
wire adapter (address discovery, connection lifecycle, factory, engine object),
hotkey FSM, actor serialization, logging, generated layout tables, the live
e2e stand, and the repo/CI scaffolding. Every listed file was read in full;
`go vet`, `go build` and `go test -race ./...` were executed and are green.

Verified strengths (checked, not assumed): the FSM invariant holds (no
goroutines, channels or `time.Now` in `internal/hotkey/fsm.go`); the FSM
logic itself survived a systematic edge-case trace (window boundaries, cap at
`MaxTaps`, modifier-use, glitches, stray releases, reset-while-held) with no
defect found; the wire field-order contract is pinned by reflection tests; the
key-trace privacy gate is correct (DEBUG only, `-debug` flag carries an explicit
password warning, e2e deletes its log); no secrets, no injection surfaces, no
eval/dynamic-exec use; `focus_helper.py`'s `ATSPI_ROLE_` regex was empirically
validated against the installed pygobject (`str(Role.PASSWORD_TEXT)` yields
`"<enum ATSPI_ROLE_PASSWORD_TEXT of type Atspi.Role>"`); godbus v5.2.2 does
close `Signal()` channels on connection termination, so the bus-loss detector
is sound.

Three defects degrade correctness in supported-but-secondary paths: the
reconnect loop's log contract is wrong after a failed first attempt, the
single-instance guard's documented zombie-retake mechanism does not actually
exist in D-Bus semantics, and the D-01 experiment can derive a wrong verdict
on locked sessions. None are security issues or data-loss risks, so none are
classified Critical.

## Narrative Findings (AI reviewer)

### Warnings

### WR-01: First registration after a failed attempt logs `re-registered` and breaks the `component registered` observability contract

**File:** `engine/conn.go:48-49` (loop), `engine/conn.go:119-123` (log site); contract documented in `internal/logging/logging.go:5-8`; consumers in `test/e2e/preflight.go:104` and `test/e2e/case_m1.go:286-291`

**Issue:** `Run` increments `generation` on every loop iteration, i.e. per
connection **attempt**, but `serve` uses `generation == 0` to choose between
`"component registered"` and `"re-registered"`. If attempt 0 fails — `Discover`
error or `Dial` failure because ibus-daemon has not created its socket yet —
the first *successful* registration (attempt 1+) is logged as
`"re-registered"`, a state that has never occurred. The daemon's own systemd
unit comment (`dist/systemd/user/goswitchd.service:7-9`) declares retry-before-
ibus-is-ready a supported mode ("registers with ibus-daemon and retries until
it answers"), so in the primary deployment the mislabel is expected, not
exceptional. Impact: (a) the documented INFO lifecycle contract is violated;
(b) `checkEngineRegistered` waits up to 10 s for the literal substring
`component registered` and `assertRegisteredBeforeFocusIn` resolves it via
`firstRecordTime` — both fail spuriously against such a log; (c) an operator
reading "re-registered, generation=3" infers a daemon-restart history that
never happened.

**Fix:** Count successful registrations, not attempts — e.g. a Run-scoped flag
threaded through `serve`:

```go
func Run(ctx context.Context, cfg Config) error {
	registeredOnce := false
	for generation := 0; ; generation++ {
		if err := serve(ctx, &cfg, generation, &registeredOnce); err != nil {
			...
		}
	}
}

func serve(ctx context.Context, cfg *Config, attempt int, registeredOnce *bool) error {
	...
	if *registeredOnce {
		slog.Info("re-registered", "generation", attempt, "engines", len(cfg.Engines))
	} else {
		slog.Info("component registered", "engines", len(cfg.Engines))
		*registeredOnce = true
	}
	...
}
```

### WR-02: Single-instance guard's claimed zombie-retake does not work; rapid respawn after SIGKILL can fatal-exit on `RequestNameReplyExists`

**File:** `engine/conn.go:88-101`; respawn paths racing it: `test/e2e/case_resilience.go:125-133` (`respawnAndWait` right after `killDaemon9`) and `dist/systemd/user/goswitchd.service:20` (`Restart=on-failure`, default `RestartSec=100ms`)

**Issue:** The comment at `engine/conn.go:88-92` claims "ReplaceExisting lets
a reconnecting generation retake its own name from a zombie connection". That
is not D-Bus semantics: replacement requires the *incumbent* to have acquired
the name with `AllowReplacement` (verified: godbus v5.2.2 `export.go:457-459`
defines `NameFlagAllowReplacement = 1<<iota; NameFlagReplaceExisting` — the
code passes only `NameFlagReplaceExisting`). Since no goswitchd connection
ever sets `AllowReplacement`, `ReplaceExisting` here is inert: against any
live owner the reply is `Exists`, never primary. Consequence: after a SIGKILL,
if ibus-daemon has not yet reaped the dead connection and released the name
when the replacement process reaches `RequestName` — exactly the race the
comment claims to cover — the new goswitchd receives `Exists`, returns
`errNotPrimaryOwner`, and exits fatally even though no other instance is
running. systemd then restarts it again (masked failure, or a failed e2e run
when `respawnAndWait`'s 15 s `waitForNew` times out). The window is narrow
(the kernel closes the socket at SIGKILL; the replacement must still boot,
`Discover`, `Dial`, `Auth`, `Hello` first) but it is real, and the documented
mitigation does not exist. Note the same-process reconnect path is safe by
construction (the old `Conn` is `Close`d in `serve`'s defer before the next
generation, and an ibus-daemon restart means a fresh bus with fresh name
state).

**Fix:** Keep the flags exactly as they are (adding `AllowReplacement` would
let a *second* goswitchd steal the name from a healthy first instance and
break T-01-03), but treat a non-primary reply as transient for a bounded
period before declaring the guard tripped:

```go
var reply dbus.RequestNameReply
for range 3 {
	r, err := conn.RequestName(cfg.Component.ComponentName, dbus.NameFlagReplaceExisting)
	if err != nil {
		return fmt.Errorf("request name %s: %w", cfg.Component.ComponentName, err)
	}
	if r == dbus.RequestNameReplyPrimaryOwner {
		reply = r
		break
	}
	// Exists may be a stale zombie connection the bus has not reaped yet.
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second):
	}
}
if reply != dbus.RequestNameReplyPrimaryOwner { ... return errNotPrimaryOwner ... }
```

Optionally add `RestartSec=1s` to `dist/systemd/user/goswitchd.service` to
widen the margin.

### WR-03: D-01 experiment can never report `switched` on locked sessions and may derive a wrong DECISION

**File:** `test/e2e/case_d01.go:309-311` (`switched`), `test/e2e/case_d01.go:384-401` (`readBackText`), `test/e2e/case_d01.go:94-100` (`decisionFor`); the locked-session surface is opened by `test/e2e/case_m1.go:101-124`

**Issue:** The stand explicitly supports locked sessions by falling back to
the shell's own entry (`openEntrySurface` → `openShellEntry`). On that surface
`readBackText` returns the length-only note
`"(length-only oracle: chars=N, err=...)"`, so `hasCyrillic(o.text)` is false
by construction and `switched()` can never be true even when the focus pair
(`outEN >= 1 && inRU >= 1`) did fire. Every probe is then journaled
"not-switched" and `decisionFor(false)` appends "DECISION: Option B —
internal flip" — a wrong experimental conclusion derived from an oracle
limitation, silently (the verdict row does carry the note, but the DECISION
row does not). This defeats the case's stated purpose ("a decision is
derivable" from observations) on a supported input surface.

**Fix:** Exclude oracle-limited rows from the decision, or require only the
focus pair when the text oracle is unavailable:

```go
func (o d01Obs) switched() bool {
	if o.oracleLimited { // set by readBackText on the shell fallback
		return o.outEN >= 1 && o.inRU >= 1
	}
	return o.outEN >= 1 && o.inRU >= 1 && hasCyrillic(o.text)
}
```

…and make `decisionFor` state when any contributing probe was oracle-limited,
or fail the case outright when no full-script oracle was available for the
whole matrix.

### Info

### IN-01: `serve` rebuilds `EngineList` that `NewComponent` already built

**File:** `engine/conn.go:109-113` vs `engine/types.go:84-104`
**Issue:** `engineConfig()` passes `Component: engine.NewComponent(engines)`,
which already populates `EngineList` with exactly the same variants; `serve`
then overwrites it with an identical copy. Dead duplication with a divergence
risk (if one site drifts, the registered payload silently differs from the
advertised constructor contract).
**Fix:** Delete the rebuild in `serve` (use `cfg.Component` as-is), or stop
populating `EngineList` in `NewComponent` and construct it only in `serve`.

### IN-02: `Engine.caps` is a dead store and a latent race

**File:** `engine/engine.go:83` (field), `engine/engine.go:164-170` (`SetCapabilities` writes it)
**Issue:** `caps` is written and never read. When Phase 2 starts reading it
(e.g. to gate surrounding-text requests on `CapSurroundingText`), the field is
unsynchronized — godbus dispatches each method call on its own goroutine, so a
read in one handler races the write in another. Nothing catches this today
because the read does not exist.
**Fix:** Remove the field until it has a consumer; when reintroducing, guard
it (e.g. route capability state through the actor's mutex like key events).

### IN-03: `modsMask` silently drops SUPER/HYPER/META state bits

**File:** `engine/keys.go:35-37`, consumed at `engine/engine.go:114-121`
**Issue:** The state word's modifier set is not limited to the low byte:
`ibustypes.h` also defines `IBUS_SUPER_MASK` (1<<26), `IBUS_HYPER_MASK`
(1<<27), `IBUS_META_MASK` (1<<28) and button masks; the comment "the rest is
reserved" mischaracterizes the header. Phase 1 (Shift_R keyval only) is
unaffected, but Phase 2/3 hotkey work involving Super (e.g. Super+space
detection, per the D-01 matrix) will silently lose those bits with no
compile-time signal.
**Fix:** Either widen the mask to the bits the project will need and document
the exact kept/dropped set, or add named constants for the dropped bits with a
comment that Phase 2 must revisit.

### IN-04: `busFileSuffix` produces non-matching suffixes for screen-suffixed or remote DISPLAY values

**File:** `engine/address.go:74-83`
**Issue:** Live-verified on this machine: the X11 bus file is named
`<machine-id>-unix-0` with `DISPLAY=:0`. For `DISPLAY=:0.0` the code yields
suffix `-unix-0.0`, and for a remote `DISPLAY=host:10.0` it yields
`-unix-host:10.0`; neither matches ibus's `<machine-id>-unix-0` naming (an
absolute-path `WAYLAND_DISPLAY` has the same problem). Discover then falls
through to "no usable ibus address" despite a live bus. Fallback-path only
(the target is Wayland with a bare `wayland-0`), so Info.
**Fix:** Normalize before joining: strip any host prefix, then cut at the
first `.` (`strings.SplitN(d, ".", 2)[0]` after removing `host:`), and use
`filepath.Base` for `WAYLAND_DISPLAY`.

### IN-05: `ExpiryAt` clobbers the timer handle even for stale no-op expiries

**File:** `internal/session/actor.go:92-100` (`a.timer = nil` at line 96)
**Issue:** A stale `AfterFunc` callback that lost the race with a newer tap
sets `a.timer = nil` unconditionally at entry, orphaning the still-armed newer
timer: a later `armTimer`/`HandleLifecycle` cannot `Stop` it, and a stray
`TimerExpired` can fire after FocusOut. Functionally benign today only because
the FSM no-ops stale expiry (`fsm.go:140-148`) — the safety rests on that
invariant with nothing linking the two.
**Fix:** Clear the handle only when it identifies the fired timer (capture the
timer in the callback closure and compare), or only clear when the FSM actually
consumed a series.

### IN-06: `withGoswitchSources` produces malformed GVariant for an empty sources list

**File:** `test/e2e/case_d01.go:284-293`
**Issue:** `gsettings get` prints an empty array as `@a(ss) []`. The
strip-brackets logic turns that into body `@a(ss) [`, and the result
`[('ibus', 'goswitch-en'), ('ibus', 'goswitch-ru'), @a(ss) []]` is rejected by
`gsettings set`, so probe 1 records "action failed" and measures nothing
instead of genuinely probing. Never triggers on the owner's populated desktop.
**Fix:** Detect the `@a(ss)` empty form first and return `goswitchSourcesPrefix`
unchanged in that case.

### IN-07: `activateGoswitch` can be satisfied by a pre-restart `focus_in` record

**File:** `test/e2e/case_m1.go:83-89`
**Issue:** `waitForLog(ctx, "focus_in", ...)` scans the whole log, so in
`runIbusRestart` the post-restart re-activation (via `activateGoswitchRetry`)
returns as soon as the *old* record is seen — before the engine is actually
active again — and typing can race activation. The downstream
`waitForNew('"msg":"key"', ...)` floor usually absorbs it, so this is a
reliability nit, not a wrong assertion. `resetToGoswitchEN`
(`case_d01.go:270-276`) already does it correctly.
**Fix:** Snapshot `countSub("focus_in")` before the `ibus engine` call and use
`waitForNew(..., before+1, focusWait)`.

### IN-08: `assertFieldOrder` panics instead of failing when a wire field is removed

**File:** `engine/wire_test.go:12-24`
**Issue:** After reporting a field-count mismatch with `t.Errorf` (non-fatal),
the loop still indexes `typ.Field(i)` for every `want` entry; with
`NumField() < len(want)` this panics with index-out-of-range and crashes the
test binary, obscuring the wire-contract message the test exists to print.
**Fix:** `if got := typ.NumField(); got != len(want) { t.Errorf(...); return }`.

### IN-09: Security workflows install tooling via unpinned pipe-to-shell and `@latest`

**File:** `.github/workflows/pr-sanity.yml:42` and `:93`; `.github/workflows/security-scheduled.yml:24` and `:36`
**Issue:** `curl https://mise.run | sh` executes an unpinned remote script on
every CI run, and `govulncheck@latest` floats. Both choices are deliberate
and documented in comments, but the security gate itself accepts unpinned
supply-chain inputs — the class of risk the gate exists to reduce elsewhere.
**Fix:** Pin the mise installer (versioned script URL or checksum verify) and
a tagged govulncheck version, bumping both via dependabot.

### IN-10: dependabot lacks the `github-actions` ecosystem

**File:** `.github/dependabot.yml:10-14`
**Issue:** Only `gomod` is covered; `actions/checkout@v4` is tracked by a
mutable tag and receives neither automated updates nor integrity pinning.
**Fix:** Add a second updates entry with `package-ecosystem: github-actions`
(and optionally pin actions to commit SHAs so dependabot manages the pins).

### IN-11: Layout generator recursion has no cycle guard

**File:** `layouts/generator/main.go:287-307` (`expandLayout`), `:311-329` (`expandInclude`)
**Issue:** Same-file includes recurse depth-first with no visited-set; an
xkb_symbols section that (transitively) includes itself would stack-overflow
the generator. Dev-side tool with a loud failure mode and golden-pinned
output, so Info only.
**Fix:** Pass a `map[string]bool` visited set through `expandLayout` and error
on re-entry (`errCrossFileInclude`-style named sentinel).

---

_Reviewed: 2026-09-11T08:35:43Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
