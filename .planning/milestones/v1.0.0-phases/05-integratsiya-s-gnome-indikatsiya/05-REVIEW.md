---
phase: 05-integratsiya-s-gnome-indikatsiya
reviewed: 2026-09-30T00:45:00Z
depth: standard
files_reviewed: 28
files_reviewed_list:
  - cmd/goswitchd/main.go
  - cmd/goswitchd/main_test.go
  - engine/conn.go
  - engine/conn_switcher_test.go
  - engine/conn_sync_test.go
  - engine/engine.go
  - engine/factory.go
  - engine/types.go
  - internal/activate/activate.go
  - internal/activate/activate_test.go
  - internal/install/install.go
  - internal/install/install_internal_test.go
  - internal/install/selfcheck.go
  - internal/install/selfcheck_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - test/e2e/case_combo.go
  - test/e2e/case_install.go
  - test/e2e/case_resilience.go
  - test/e2e/case_switch.go
  - test/e2e/case_switch_test.go
  - test/e2e/main.go
  - test/e2e/matrix.go
  - .golangci.yml
  - mise.toml
  - README.md
  - test/e2e/README.md
status: findings
critical_count: 0
warning_count: 5
info_count: 4
---

# Phase 05 Code Review

**Reviewed:** 2026-09-30T00:45:00Z
**Depth:** standard (concurrency/deadlock axes traced at deep level — this phase is the daemon's flip/sync core)
**Files Reviewed:** 28
**Status:** findings

## Summary

Phase 05 delivers the ADR-006 two-engine revision: the D-52 `SetGlobalEngine` flip
(`BindSwitcher` seam, `actor.flipTo`), the sync listener (`GlobalEngineChanged` +
`FocusIn` → `actor.SyncEngine`, echo suppression, storm-safe signal handler), the
factual-engine `GetGlobalEngine` reader wired into `activate.IfOwned`, the D-54
two-source installer wrap, and the 05-05 live-proof e2e cases. The build and vet are
clean; the corpus is unusually strong (in-process fake ibus bus with a hand-rolled
SASL server, closed-enum sync pins, pure flip-marks oracles). The core state machines
hold up under adversarial tracing: the `SyncEngine` closed list genuinely closes
(own-flip echo = same-mode no-op, drift = record+WARN, foreign = WARN+untouched,
zero switcher calls from any sync input), the storm handler is bounded and preserves
the `errBusClosed` verdict, and per-generation seam rebinding is correctly ordered
(bind before PostRegister in the same serve goroutine — no race on `readGlobalEngine`).

No Critical findings. Five Warnings, led by a mutex-coupling hazard on the new flip
path that best explains ADR-006's own A6 live anomaly (every live flip journaling a
failure WARN at exactly the 40 ms deadline while the switch lands ~1 ms later), and a
documented-but-nonexistent timeout bound on the `PostRegister` reactivation path.
The `.golangci.yml` wrapcheck relaxation was scrutinized as instructed: it is scoped
to `.*_test\.go` only with a stated rationale — properly contained, no finding.

## Warnings

### WR-01: flipTo holds the actor mutex across the SetGlobalEngine round trip that can trigger CreateEngine → AttachEngine → the same mutex

**File:** `internal/session/actor.go:1362-1383` (flipTo), `engine/factory.go:47` (AttachEngine), `engine/conn.go:238-250` (newSwitcher)
**Issue:** `flipTo` runs the switcher closure synchronously **under `a.mu`** (documented
as the deliberate choice over the async handoff). The closure issues
`org.freedesktop.IBus.SetGlobalEngine` on the bus. Setting the global engine makes
ibus-daemon instantiate the target engine for the focused input context **through our
own factory**: the incoming `CreateEngine` call dispatches on its own godbus goroutine
and calls `f.handler.AttachEngine(eng)` → `actor.AttachEngine` → `a.mu.Lock()` — the
mutex `flipTo` is holding while waiting for the `SetGlobalEngine` reply. If ibus-daemon
serializes the `SetGlobalEngine` reply behind engine creation (or any slow
create→attach path), the daemon deadlocks against itself until the flip's own 40 ms
`switchTimeout` cancels the call and releases `a.mu` — after which `CreateEngine`
completes and the switch lands anyway.

This matches ADR-006's A6 live finding precisely: the `switch_engine` WARN fires at
~40.4 ms on **every** live flip (deterministic ×2), and the `engine created` record
trails by only ~1.2–1.4 ms — i.e. the bus "completes" ~1 ms after our own deadline
releases the mutex. Pure bus latency would not deterministically track our internal
deadline across runs; a lock-release point would. Consequences today: every live flip
journals a **failure** record for a successful switch (misleading journal, WARN
noise on the phase's primary mechanism), the bus reply is discarded as a cancelled
call, and the flip costs the full 40 ms budget instead of the real (much smaller)
RTT. Note the ADR's candidate mitigation (raise the deadline to 45–48 ms) does **not**
fix this if the coupling theory is right — it only moves the release point, and the
WARN still fires at the new deadline on every flip.

**Fix:** First discriminate: log a timestamp at `CreateEngine` entry and compare it
against the flip's call instant in a live run (if `CreateEngine` entry sits inside the
flip's 40 ms window, the coupling is confirmed). If confirmed, break the coupling —
options: (a) release `a.mu` around the switcher call with a dedicated flip mutex
serializing flips, keeping the pinned record order by emitting the panel symbol after
re-acquisition; or (b) make `AttachEngine` a non-blocking handoff (enqueue the emitter,
consume under the mutex at the next event/timer entry) so `CreateEngine` never blocks
on `a.mu`. Alternative mitigation that avoids the creation path entirely: have
`serve`/the factory proactively trigger creation of both engines right after
registration (ibus does this on first focus anyway), so a flip never races a
`CreateEngine`.

### WR-02: PostRegister reactivation's GetGlobalEngine call is unbounded, contradicting the Config contract that claims a bound

**File:** `engine/conn.go:40-44` (contract), `engine/conn.go:378-395` (reader), `cmd/goswitchd/main.go:184-186`, `internal/activate/activate.go:150-168`
**Issue:** `Config.PostRegister`'s doc promises: "serve runs it inline before
waitBusLoss, so a hang here stalls the generation (the activate package's 10 s
per-call timeout bounds it)." That is false for the new factual-engine path:
`IfOwned` → `globalEngine(ctx)` → `newGlobalEngineReader` issues
`ibus.CallWithContext(ctx, ...GetGlobalEngine)` with **the raw daemon context — no
deadline**. The activate package's 10 s timeout bounds only its **subprocess** calls
(`NewExecRunner`); it never touches the godbus call. A wedged/half-dead ibus-daemon
that accepts but never replies stalls the serve generation indefinitely: no signal
dispatch, no reconnect, and daemon shutdown (ctx cancel) cannot interrupt it — the
same failure class Pitfall 4 deemed worth a hard 40 ms bound on the flip path.
Related and unattributed in the same cycle: `dialIbus` dials with no context at all
(`dbus.Dial(addr)`, `conn.go:189` — pre-existing shape, re-landed inside the new
`dialIbus`), and `RegisterComponent`/`AddMatchSignalContext` run with the deadline-free
daemon ctx (`conn.go:143`, `conn.go:215`).

**Fix:** Bound the whole seam window in `serve` — e.g. wrap the ctx passed to
`PostRegister`/`BindSwitcher`/`AddMatchSignalContext` in
`context.WithTimeout(ctx, postRegisterBudget)` (a few seconds covers the 3×500 ms
subprocess retries plus one D-Bus RTT), and give `dbus.Dial` a dial deadline. Then the
"the activate package's 10 s per-call timeout bounds it" sentence becomes true for the
subprocess leg and the doc should name the D-Bus bound explicitly.

### WR-03: Install always activates goswitch-en, overriding a ru-first desktop's active layout — and uninstall does not restore `current`

**File:** `internal/install/install.go:375` (`activateEngine(ctx, engineEN)`), report line `install.go:894`
**Issue:** D-54 (and the pinned corpus, e.g. `TestWrapSourcesPairPreserved`
"the ru slot must wrap first") commits to wrapping *the user's own pair with positions
preserved*. But the final step of `Install` hardcodes activation of `engineEN`. On a
desktop whose pair is ru-first (or whose current source was the ru layout), the shell
activates the wrapped source at the user's current index after the sources write — and
`Install` then forces goswitch-en, silently switching the session's active layout to
English. The state file saves `sources` and the two switch chords but **not**
`current`, so `Uninstall` restores sources yet still leaves the user on EN — an
install→uninstall round trip that does not return the desktop to its pre-install
active-layout state. The report line ("engine: activated goswitch-en") documents the
act but not the deviation from D-54's "the layout choice stays yours".

**Fix:** Derive the activation target from the computed wrapper (activate the engine at
the user's pre-install/current index — the pair source is already resolved in
`takeoverSources`), falling back to `engineEN` only when no derivation is possible.
Alternatively record `current` in `installState` and restore it in `Uninstall`.

### WR-04: README names the wrong gsettings schema for the cleared layout-switch binding

**File:** `README.md:58-59`
**Issue:** The README (a paragraph reworded this phase) says install clears
"`org.gnome.desktop.input-sources` `switch-input-source`". The code clears
`org.gnome.desktop.wm.keybindings` `switch-input-source[-backward]`
(`internal/install/install.go:91-93`) — and `install.go`'s own comment records the live
finding that the `desktop.input-sources` schema "carries no switch key at all
(a gsettings get answers 'No such key')". A user trying to inspect or manually restore
the binding per the README hits exactly that "No such key" error.

**Fix:** Change the README to `org.gnome.desktop.wm.keybindings` for both
`switch-input-source` and `switch-input-source-backward`.

### WR-05: e2e standalone cases register the unit-restore defer only after the stop succeeds — a failed stop can strand the owner's daemon in the stopped state

**File:** `test/e2e/case_switch.go:252-262` (runSwitchSpike), `test/e2e/case_switch.go:1044-1049` (runTwoSourceFlip), `test/e2e/case_switch.go:1139-1144` (runExternalFlipSync)
**Issue:** All three cases do:

```go
if unit == unitActiveState {
    if _, err := runCmd(ctx, "systemctl", "--user", "stop", "goswitchd"); err != nil {
        return fmt.Errorf(...)   // <- returns BEFORE the restore defer is registered
    }
}
defer func() { s.restoreSpikeWindow(ctx, unit) }()
```

If the stop command errors after the stop was actually initiated (e.g. the 15 s
`cmdTimeout` fires while systemd is mid-stop, or the unit enters a wedged state), the
case returns with **no restore registered**: the owner's unit daemon stays stopped,
and nothing in the stand's teardown brings it back (teardown restores gsettings and
the stand daemon, not systemd units). Every other owner-desktop mutation in this stand
is defer-guarded before the mutation; the stop is the one exception.

**Fix:** Register `defer func() { s.restoreSpikeWindow(ctx, unit) }()` **before** the
stop call (the restore is idempotent and a no-op when the unit was already inactive),
or restore the unit in the stop-error branch before returning.

## Info

### IN-01: Stale comment contradicts the case's own two-source precondition

**File:** `test/e2e/case_resilience.go:66-68`
**Issue:** `runIbusRestart` still carries "GNOME re-activates the plain xkb sources
after a bus restart — goswitch is not in gsettings sources", but the same case now
**requires** the two-source goswitch desktop (`pinRestartEngineRU` →
`requireTwoSourceDesktop`, lines 139-141). The comment describes the pre-05 world and
will mislead the next reader of the restart oracle.
**Fix:** Update the comment to the two-source reality (the re-activation after the
restart re-asserts the goswitch engine the shell brought back).

### IN-02: Engine.caps is write-only dead state

**File:** `engine/engine.go:110, 242`
**Issue:** `Engine.caps` is assigned in `SetCapabilities` and never read anywhere
(the actor keeps its own `a.caps`). Pre-existing, but the file was touched this phase
and the field invites the assumption that it does something.
**Fix:** Drop the field (or use it in the nil-handler observer path it was presumably
meant for).

### IN-03: 05-03 SUMMARY misstates the fake-bus stand size

**File:** `.planning/phases/05-integratsiya-s-gnome-indikatsiya/05-03-SUMMARY.md:198` ("~150 test-only LOC") vs `engine/conn_switcher_test.go` (620 lines)
**Issue:** The deviation note claims the fake ibus bus stand cost ~150 LOC; the
delivered corpus (bus + SASL server + dispatch loop + seam tests) is 620 lines. A
facts-only mismatch in the planning record; the code itself is fine and well-built.
**Fix:** Correct the SUMMARY actual, or note the growth was sanctioned.

### IN-04: writeAtomic performs no fsync before rename — install-state.json durability

**File:** `internal/install/install.go:1234-1264`
**Issue:** The state file is "the restore key of the desktop" (0600), yet the atomic
write renames without fsync of file or directory; a power loss right after install can
leave a truncated/empty state file. Mitigated: `savedSources`' shape check refuses a
corrupt value and uninstall falls back to `('xkb','us')` (reported, never silent), so
the failure mode is "user's original pair lost, safe default applied" — degradation,
not bricking.
**Fix:** `tmp.Sync()` before close and a directory fsync after rename for the state
file (the XML/unit can stay as-is; only the state file is irreplaceable).

## Verified clean (adversarial probes that did NOT stick)

- **Echo suppression (05-04's mandated obligation):** `SyncEngine`'s closed list
  (`engine.NameEN`/`NameRU`/default) genuinely closes — own echo is the same-mode
  no-op branch, drift records+WARNs, foreign warns untouched, and no sync input can
  reach the switcher (`TestActor_SyncEngineNeverSwitches`). A flip that times out but
  lands late is absorbed as a silent echo; a genuine failure self-heals on the next
  `FocusIn` sync.
- **stormHandler:** drop-on-full under one mutex, single bounded WARN, `Terminate`
  closes channels exactly once (`chans = nil`), `RemoveSignal` after `Terminate` is a
  safe no-op; memory bounded by `signalBufferSize`; the dispatcher's synchronous
  `OnGlobalEngine` can block at most one flip deadline, and the drop policy covers the
  overflow (missed sync self-heals on FocusIn).
- **`readGlobalEngine` rebinding:** bind and read are sequential in the same serve
  goroutine (bind before PostRegister) — no data race; a dead generation's closure
  fails its call and is replaced per generation (pinned).
- **GVariant injection:** `wrapSources` renders only from the closed enum over
  regex-parsed tuples (no quotes can survive the parser); the raw user line never
  reaches `gsettings set`. Refusals name the tuple type, never the content.
- **Double-close / goroutine leaks** across ibus restarts: `conn.Close` → `Terminate`
  is idempotent; serve's defer closes each generation's connection; no channel or
  goroutine outlives its generation.
- **wrapcheck relaxation:** scoped to `.*_test\.go` only, rationale recorded — no
  production error-surface loss.

---

_Reviewed: 2026-09-30T00:45:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
