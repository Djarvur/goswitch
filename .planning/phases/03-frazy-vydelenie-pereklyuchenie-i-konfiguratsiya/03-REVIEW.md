---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
reviewed: 2026-09-15T15:12:16Z
depth: standard
files_reviewed: 51
files_reviewed_list:
  - cmd/goswitchctl/main.go
  - cmd/goswitchd/main.go
  - docs/CONFIG.md
  - engine/engine.go
  - engine/engine_test.go
  - engine/keys.go
  - .github/workflows/e2e-matrix.yml
  - go.mod
  - go.sum
  - internal/appid/appid.go
  - internal/appid/appid_test.go
  - internal/clipboard/clipboard.go
  - internal/clipboard/clipboard_test.go
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/config/load.go
  - internal/config/load_test.go
  - internal/config/watch.go
  - internal/config/watch_test.go
  - internal/correct/buffer.go
  - internal/correct/buffer_test.go
  - internal/correct/plan.go
  - internal/correct/plan_test.go
  - internal/correct/runs.go
  - internal/correct/runs_test.go
  - internal/correct/verify.go
  - internal/correct/verify_test.go
  - internal/ctlsvc/ctlsvc.go
  - internal/ctlsvc/ctlsvc_test.go
  - internal/hotkey/fsm.go
  - internal/hotkey/fsm_test.go
  - internal/hotkey/names.go
  - internal/hotkey/names_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - test/e2e/case_combo.go
  - test/e2e/case_ctl.go
  - test/e2e/case_m1.go
  - test/e2e/case_macr.go
  - test/e2e/case_phrase.go
  - test/e2e/case_resilience.go
  - test/e2e/case_select.go
  - test/e2e/case_word.go
  - test/e2e/cases/matrix-v1.yaml
  - test/e2e/cases/matrix-v2.yaml
  - test/e2e/main.go
  - test/e2e/matrix.go
  - test/e2e/matrix_test.go
  - test/e2e/preflight.go
  - test/e2e/README.md
findings:
  critical: 3
  warning: 5
  info: 6
  total: 14
status: issues_found
---

# Phase 3: Code Review Report

**Reviewed:** 2026-09-15T15:12:16Z
**Depth:** standard
**Files Reviewed:** 51
**Status:** issues_found

## Summary

Full standard-depth pass over the phase-3 scope: config schema/hot-reload,
correction pipeline (word/phrase/selection), clipboard rung, MACR layer,
ctlsvc D-Bus surface, hotkey FSM/binding tables, both binaries, and the e2e
stand + YAML matrices.

The pure core (correct/*, hotkey/*) is solid — rune arithmetic, run
segmentation, and the FSM are correct and well-tested. The dominant defects
are in the wiring seams, exactly where the linter cannot look:

1. **Two documented config knobs are validated and reload "applied" but never
   consumed by the daemon** (`hotkeys.tap_key`, `timeouts.verify_wait_ms`) —
   the precise "appears applied while it is not" failure D-33 was written to
   prevent (CR-01/CR-02). The e2e matrix even carries a case that "proves" a
   `verify_wait_ms` reload with a value identical to the base document.
2. **The clipboard rung can destroy the user's clipboard**: a wl-paste killed
   by the rung's own 1500 ms deadline is misread as the *empty* clipboard, so
   the restore path clears a non-empty clipboard (CR-03). The unit test
   asserts the intended behavior but its fake returns a non-`ExitError` for
   cancellation, masking the real `os/exec` shape.
3. The same rung runs up to three 1.5 s subprocesses **while holding the
   actor mutex**, stalling every keystroke for up to ~4.5 s (WR-01), and a
   self-documented verify-after race can trigger the rung into a duplicate
   paste (WR-05).

## Critical Issues

### CR-01: `hotkeys.tap_key` is validated, documented, and silently ignored — the FSM always listens to Shift_R

**File:** `internal/session/actor.go:233-247`, `internal/session/actor.go:648-674`, `internal/session/actor.go:421`, `cmd/goswitchd/main.go:114-126`
**Issue:** `docs/CONFIG.md:22` documents `hotkeys.tap_key` as "the key whose
tap series (single/double/triple) drives switch/correct-word/correct-phrase".
`config.Hotkeys.validate` resolves it through `hotkey.ParseBinding`, so
`tap_key: ctrl_l` (or any bindable name, see WR-02) passes validation, loads
at startup, and reloads with "applied". But no consumer exists:

- `NewActor` hardcodes `hotkey.NewFSM(window, hotkey.KeyvalShiftR)` (actor.go:235);
- `applySnapshot` (actor.go:648-674) folds in the tap window, cap, rung,
  combo, and all MACR fields — but never `snap.Hotkeys.TapKey`;
- `HandleKey` re-hardcodes `hotkey.KeyvalShiftR` for the timer arming
  condition (actor.go:421);
- `cmd/goswitchd/newActor` passes only `TapWindowMs`, `BackspaceCap`,
  `ClipboardRung`.

Verified: `grep -rn "\.TapKey" cmd internal/session engine` (non-test) has
zero hits. The FSM itself supports the knob (`NewFSM(window, tapKeyval)`,
pinned by `TestFSM_ConfigurableTapKey`) — only the wiring is missing. Every
e2e config uses `tap_key: shift_r`, so no test can catch it. A user who
remaps the tap key gets a daemon that "applies" the document while still
deciding taps of right Shift — the exact failure mode D-33's strict decoding
was supposed to make impossible.

**Fix:** thread the configured key through, mirroring the `comboName` cache:

```go
// in applySnapshot (caller holds mutex):
if name := snap.Hotkeys.TapKey; name != a.tapKeyName {
    if binding, err := hotkey.ParseBinding(name); err == nil {
        a.fsm = hotkey.NewFSM(a.window, binding.Keyval) // or add FSM.SetTapKey
        a.tapKeyName = name
    }
}
// in NewActor: parse cfg.Hotkeys.TapKey (fallback shift_r) before NewFSM;
// in HandleKey: compare against the configured tap keyval, not the constant.
```

### CR-02: `timeouts.verify_wait_ms` is validated, documented, and silently ignored — the actor hardcodes 100 ms

**File:** `internal/session/actor.go:29`, `internal/session/actor.go:730`, `internal/session/actor.go:750`, `internal/session/actor.go:1239`, `internal/session/actor.go:1323`
**Issue:** Same class as CR-01. `docs/CONFIG.md:25` documents
`timeouts.verify_wait_ms` as the ADR-004 verify budget; `Timeouts.validate`
enforces `(0, 2000]`. The actor, however, uses the package const
`verifyWait = 100 * time.Millisecond` for every `time.AfterFunc(verifyWait, …)`
(the pending-fix deadline at actor.go:1239/1323 and both verify-after
deadlines at actor.go:730/750). `applySnapshot` never reads
`snap.Timeouts.VerifyWaitMs` and `Options` has no field for it. Setting
`verify_wait_ms: 500` reloads "applied" and changes nothing — on a slow
client the correction still times out after 100 ms. The e2e
`ctl-status-reload` matrix case (matrix-v2.yaml:240-249) reloads
`verify_wait_ms: 100` — the same value as the base document — so it asserts
the reload plumbing while giving false confidence the knob is live.

**Fix:** make the budget actor state:

```go
// Actor field: verifyWait time.Duration (default 100 ms from NewActor)
// applySnapshot:
if w := time.Duration(snap.Timeouts.VerifyWaitMs) * time.Millisecond; w != a.verifyWait {
    a.verifyWait = w
}
// replace all three time.AfterFunc(verifyWait, …) call sites with a.verifyWait.
```

### CR-03: A wl-paste killed by the rung's own deadline is misread as the empty clipboard — the restore then CLEARS the user's clipboard (data loss)

**File:** `internal/clipboard/clipboard.go:76-88`, `internal/clipboard/clipboard.go:145-165`
**Issue:** `Save` classifies *any* `*exec.ExitError` as "clipboard is empty":

```go
if err != nil {
    var exitErr *exec.ExitError
    if errors.As(err, &exitErr) {
        return nil, false, nil // wl-paste's empty-clipboard exit
    }
    ...
}
```

`execRunner` runs wl-paste through `exec.CommandContext` with the package
`cmdTimeout` (1500 ms). When the deadline fires, `Cancel` SIGKILLs the
process and `cmd.Run` returns a `*exec.ExitError` ("signal: killed") — which
`errors.As` happily matches. `Save` then reports `had=false, err=nil`, and
`runClipboardRung` → `Restore(saved, false)` → `Clear()` runs
`wl-copy --clear`, **wiping a non-empty clipboard** because a transient
wl-paste hang looked like "no selection". This is precisely the outcome the
D-28/D-29 save/restore design exists to prevent.

`TestClipboard_Timeouts` ("cancelled context is an error, not an empty read")
pins the correct intent, but its fake returns `fmt.Errorf("subprocess
canceled: …")` for cancellation — a plain error — so the real `os/exec` kill
shape is never exercised and the misclassification is invisible to the
corpus.

**Fix:** distinguish the empty-clipboard exit from a signal kill before
treating `ExitError` as "empty":

```go
if err != nil {
    var exitErr *exec.ExitError
    if errors.As(err, &exitErr) && exitErr.ExitCode() >= 0 && ctx.Err() == nil {
        return nil, false, nil // genuine wl-paste non-zero exit: empty clipboard
    }
    return nil, false, fmt.Errorf("clipboard save: %w", err)
}
```

(`exitErr.ExitCode()` is `-1` for a signal kill; also short-circuit on
`ctx.Err() != nil`.) Add a corpus row whose fake returns a signal-shaped
`*exec.ExitError` (`ProcessState` via a real killed subprocess, or
constructing the exit error from `os.ProcessState`) and assert `Save`
returns an error, not `had=false`.

## Warnings

### WR-01: The clipboard rung runs up to three 1.5 s subprocesses while holding the actor mutex — every keystroke stalls behind it

**File:** `internal/session/actor.go:785-809`
**Issue:** `runClipboardRung` is called from `HandleSurroundingText` with
`a.mu` held and synchronously executes `Save` + `Set` + `Restore` + the
forward burst — three subprocesses at 1500 ms each ≈ up to 4.5 s during which
every `ProcessKeyEvent`, `SetSurroundingText`, and lifecycle call blocks on
the mutex. The daemon's contract is <50 ms key reaction; a wedged
wl-clipboard pair freezes all input through the engine for the full budget.
This is input-path robustness, not perf tuning.

**Fix:** deliver the rung off the event path — capture `paste`, snapshot the
engine emitter, and run the subprocess sequence on a dedicated goroutine (or
shrink the per-call budget for the rung and run Save concurrently with Set
preparation). At minimum, drop the mutex before the subprocesses and
re-acquire only to update counters, since the rung touches no FSM/buffer
state.

### WR-02: `tap_key` validation accepts modifier-bearing bindings the tap FSM cannot honor

**File:** `internal/config/config.go:151-160`, `internal/hotkey/names.go:123-146`
**Issue:** `Hotkeys.validate` parses `tap_key` through the full
`ParseBinding` grammar, so `tap_key: ctrl+shift_r` or even
`tap_key: shift+ctrl_r` (which collides with the default combo) validates.
Tap semantics are key-only — the FSM tracks a single keyval and no modifier
state — so even after CR-01 is wired, such a document would "apply" while
its modifier tokens are ignored. The combo is the only binding that uses
modifiers.

**Fix:** require `tap_key` to be a bare key — `strings.Contains(h.TapKey, "+")`
rejects, or assert `binding.ModMask == FamilyMask(binding.Keyval)` after
parsing.

### WR-03: `goswitchNameOwner` authenticates EXTERNAL with the PID instead of the UID

**File:** `test/e2e/matrix.go:616`
**Issue:** `dbus.AuthExternal(strconv.Itoa(os.Getpid()))` — the EXTERNAL
mechanism's initial response must be the decimal UID of the socket peer.
Every other dial site does it correctly (`engine/conn.go:167`,
`internal/appid/appid.go:135`, `test/e2e/preflight.go:192` — all
`os.Getuid()`). It only works today because ibus-daemon authorizes on
SO_PEERCRED and tolerates the claimed identity; a stricter bus (or a
regression in ibus-daemon) turns every matrix preflight into
`dbus auth` failures that name nothing.

**Fix:** `dbus.AuthExternal(strconv.Itoa(os.Getuid()))`.

### WR-04: Workflow interpolates `inputs.matrix` directly into shell — expression-injection pattern

**File:** `.github/workflows/e2e-matrix.yml:68`
**Issue:** `if [ "${{ inputs.matrix }}" = "test/e2e/cases/matrix-v1.yaml" ]`
splices a dispatch input into `run:` shell verbatim. The trigger is
`workflow_dispatch` only (write-role users, as the header notes), so
exploitation requires privilege — but this is the canonical GitHub
script-injection anti-pattern, and the runner is a self-hosted agent living
inside the owner's graphical session where a `$(…)` payload would execute.
Defense-in-depth is cheap here.

**Fix:**

```yaml
- name: e2e matrix (${{ inputs.matrix }})
  env:
    MATRIX: ${{ inputs.matrix }}
  run: |
    if [ "$MATRIX" = "test/e2e/cases/matrix-v1.yaml" ]; then
      mise run e2e-matrix
    else
      mise run e2e-matrix-v2
    fi
```

### WR-05: The verify-after stale-answer race is documented as known and unfixed — with `clipboard_rung` on it pastes a duplicate word into the field

**File:** `internal/session/actor.go:506-518` (after-branch), `test/e2e/matrix.go:1306-1319` (the documented LIVE FINDING)
**Issue:** matrix.go's own comment records that a client may push the
INTERMEDIATE post-delete/pre-commit state, so the verify-after round can
compute a **false mismatch** while the field actually ends up correct, and
calls the race "a deferred item". With `correction.clipboard_rung: true`
(opt-in, D-28), the false mismatch routes straight into `runClipboardRung`,
which pastes `converted` over whatever the client holds — a duplicate of the
just-applied correction in the user's field. Without the rung the damage is
an INFO record plus a wrong skip counter; with it, it is user-visible text
corruption.

**Fix:** debounce the after-verdict — on mismatch, issue one more
`RequireSurroundingText` and re-compare before concluding (budgeted by the
verify wait), or have `afterVerdict` accept a push whose cursor position
proves it predates the commit (the intermediate state has the pre-commit
cursor). At minimum, gate the rung on two consecutive mismatches.

## Info

### IN-01: `logCorrectionDone` writes the source word and converted result to the log at DEBUG

**File:** `internal/session/actor.go:1409-1418`
**Issue:** `"source", string(token), "result", string(converted)` puts typed
text into the log stream. The `-debug` flag help honestly warns that
keystrokes become visible, and the status surface (T-03-06-03) stays clean,
but nothing in the e2e stand greps `source=`/`result=` (only
`"msg":"correction","level":N`), so the two fields buy no oracle coverage
while weakening the "counts and states only" discipline the surrounding
comments assert.
**Fix:** drop the two fields (keep `level`/`runes`/`latency_ms`), or document
in the comment that the DEBUG record deliberately carries the pair so the
boundary is at least stated where D-20/D-21 is cited.

### IN-02: `Engine.caps` is write-only state

**File:** `engine/engine.go:103-109`, `engine/engine.go:234-243`
**Issue:** `SetCapabilities` stores `e.caps` but nothing ever reads it (the
actor keeps its own copy via `HandleCapabilities`). Dead field; because it
is written from godbus dispatch goroutines without synchronization, the
first future reader inherits a data race.
**Fix:** delete the field (handler-only forwarding, like the other no-op
methods) or route it exclusively through the handler.

### IN-03: e2e README `-case` table is stale

**File:** `test/e2e/README.md`
**Issue:** The `-case` help table lists only the phase-1/2 cases; the
registry (and `caseListUsage`) also carries `select-smoke`,
`select-correct`, `select-clipboard`, `combo-word-layout`, `layout-single`,
`super-space-alive`, `macr-probe`, `macr-super-letter`, `macr-per-app`,
`ctl-smoke`.
**Fix:** regenerate the table from `caseListUsage()` (or point the reader at
`go run ./test/e2e -help`).

### IN-04: `ParseBinding` accepts degenerate bindings

**File:** `internal/hotkey/names.go:123-146`
**Issue:** `"shift+shift+ctrl_r"` (duplicate modifier),
`"ctrl_l+ctrl_r"` (a modifier token that is also a key name) parse
successfully; `"shift+ctrl_l"` masks the key's own family bit ambiguously
for combo matching. Harmless today because the vocabulary is small, but the
grammar has no canonical-form check.
**Fix:** reject duplicate modifier tokens and key names used as modifiers.

### IN-05: `ExpiryAt` clears the timer reference it no longer owns

**File:** `internal/session/actor.go:552-570`
**Issue:** An `Expiry` that blocked on the mutex while `armTimer` re-armed a
newer timer still executes `a.timer = nil`, orphaning the live timer's
reference (a later `HandleLifecycle` then Stops nothing). Benign today — the
FSM absorbs the stale expiry and `time.Timer` fires at most once — but the
wipe invites future double-arming.
**Fix:** tag the timer (`a.timerSeq++` captured by the callback) and clear
only when the sequence matches, as `pendingAfter.epoch` already does for the
verify rounds.

### IN-06: mise pins Go 1.23.x while the stack doc mandates building with the current 1.27.x

**File:** `mise.toml:4`
**Issue:** STACK/AGENTS state the policy "go.mod stays 1.23, build with the
current 1.27.x toolchain"; mise — the declared single source of tool
versions (D-09/D-11) — installs `go = "1.23"`, so CI builds with a 1.23.x
toolchain while the owner builds with 1.27.1. Divergent vet/lint/toolchain
behavior between local and CI runs is exactly what the single-source rule
exists to prevent.
**Fix:** `go = "1.27"` (or whatever the owner's current toolchain is) — the
`go 1.23` directive in go.mod already keeps the language version compatible.

---

_Reviewed: 2026-09-15T15:12:16Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
