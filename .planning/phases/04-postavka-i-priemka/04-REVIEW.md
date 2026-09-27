---
phase: 04-postavka-i-priemka
reviewed: 2026-09-16T12:00:00Z
depth: standard
files_reviewed: 37
files_reviewed_list:
  - .github/workflows/e2e-matrix.yml
  - .github/workflows/release.yml
  - .gitignore
  - .goreleaser.yaml
  - README.md
  - README.ru.md
  - cmd/goswitchctl/main.go
  - cmd/goswitchd/main.go
  - cmd/goswitchd/main_test.go
  - cmd/goswitchd/version.go
  - docs/ACCEPTANCE.md
  - docs/ci-runner.md
  - internal/ctlsvc/ctlsvc.go
  - internal/ctlsvc/ctlsvc_test.go
  - internal/install/install.go
  - internal/install/install_internal_test.go
  - internal/install/install_test.go
  - internal/install/selfcheck.go
  - internal/install/selfcheck_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - test/e2e/case_d01.go
  - test/e2e/case_install.go
  - test/e2e/case_m1.go
  - test/e2e/case_resilience.go
  - test/e2e/case_surface.go
  - test/e2e/cases/matrix-v3.yaml
  - test/e2e/focus_helper.py
  - test/e2e/main.go
  - test/e2e/matrix.go
  - test/e2e/perf.go
  - test/e2e/perf_test.go
  - test/e2e/preflight.go
  - test/e2e/surface.go
  - test/e2e/watchdog_test.go
findings:
  critical: 1
  warning: 4
  info: 6
  total: 11
status: findings
---

# Phase 04: Code Review Report

**Reviewed:** 2026-09-16T12:00:00Z
**Depth:** standard
**Files Reviewed:** 37
**Status:** findings

## Summary

Reviewed all 37 phase-04 files at standard depth: the install/uninstall/selfcheck
lifecycle (`internal/install`), the ctl client surface (`cmd/goswitchctl`),
the version stamping (`cmd/goswitchd`), the D-Bus control service
(`internal/ctlsvc`), the session actor, the release pipeline (goreleaser +
tag workflow), the D-48 double-run CI gate, the perf harness, matrix v3,
and both READMEs. Verification gates are green on this tree: `go build ./...`,
`go vet ./...`, `go test -race -count=1 ./...` (13 packages, 0 failures).

The phase is overall solid: the installer's atomic-write + read-back guard,
the env-carrying write-cache discipline, the ASVS V5 shape-checked restore,
and the WR-05 verify-after debounce are all correctly implemented and pinned
by tests. The workflows honor the least-privilege split (contents:write only
in release.yml), the dispatch input is injection-safe (env, not shell
interpolation), and the double-run gate has no cleanup between runs.

One Critical defect was found in the uninstall path: on any machine where
the install-state file is absent (never installed, or already uninstalled),
`goswitchctl uninstall` silently overwrites the user's GNOME input sources
with `[('xkb', 'us')]` — a destructive no-op that contradicts the package's
own "green no-op" contract and is invisible to the corpus because
`TestUninstall_Idempotent` asserts only file artifacts, never the gsettings
calls of the second uninstall.

## Critical Issues

### CR-01: Uninstall without install state clobbers the user's input sources

**File:** `internal/install/install.go:616-636` (reached from `Uninstall`, `internal/install/install.go:343`)

**Issue:** `Uninstall` calls `restoreSources` unconditionally. When the state
file `~/.local/share/goswitch/install-state.json` is absent,
`savedSources` (`install.go:753-768`) returns `("", false)`, and the
not-trusted branch sets `value = fallbackSources` — then executes
`gsettings set org.gnome.desktop.input-sources sources "[('xkb', 'us')]"`
anyway. Two real scenarios destroy user configuration:

1. `goswitchctl uninstall` on a machine where goswitch was **never
   installed** — the user's current source list (e.g. `us + ru`, or anything
   else) is overwritten with US-only, plus an unnecessary `ibus write-cache`
   + `ibus restart` on a clean desktop.
2. A **second** `uninstall` (the package doc at `install.go:314-315` promises
   "uninstalling an already removed installation is a green no-op") — the
   sources restored by the first uninstall, including any layouts the user
   re-added in between, are overwritten with US-only.

This is silent, unrecoverable desktop-config data loss behind a documented
no-op. The unit corpus misses it: `TestUninstall_Idempotent`
(`internal/install/install_test.go:724-745`) asserts only that file
artifacts stay absent, not that the second uninstall's `gsettings set` never
fires. The e2e install-cycle case only ever uninstalls once, so it passes.

**Fix:** make the restore conditional on the state file actually existing —
only a present state file may rewrite `sources` (the malformed-value
fallback stays as-is; absence means nothing to restore):

```go
func (i *Installer) restoreSources(ctx context.Context) ([]string, error) {
	statePath := i.path(stateDirRel, stateFile)
	if _, err := os.Stat(statePath); errors.Is(err, os.ErrNotExist) {
		// Never installed, or already uninstalled: a true no-op — the
		// desktop's sources were never taken over by (or already restored
		// from) this deployment. Touch nothing.
		return []string{"sources: no install state — desktop sources untouched"}, nil
	}
	value, trusted := savedSources(statePath)
	// ... unchanged from here
}
```

Add a corpus pin: after the second `Uninstall` in `TestUninstall_Idempotent`,
assert that no `gsettings set` call is recorded for the uninstall suffix,
and add a never-installed-uninstall test asserting the same from a clean
`$HOME`.

## Warnings

### WR-01: Uninstall discards stop/disable errors — `unitGone` classification is dead code

**File:** `internal/install/install.go:322-323` (with `unitStep` at 602-609 and `unitGone` at 793-799)

**Issue:** `Uninstall` ignores both unit-step errors wholesale (`_ =
i.unitStep(...)`). `unitStep` carefully tolerates only the
already-gone-unit verdicts via `unitGone`, but the caller discards even the
errors `unitStep` deliberately surfaces — so a genuine failure (wedged
systemd user manager, bus hiccup) is swallowed. The consequence: the unit
file is then removed and `daemon-reload` runs while the daemon may still be
alive — a zombie goswitchd keeps holding the IBus registration and the
`org.djarvur.goswitch` control name after an "successful" uninstall.

**Fix:** propagate the non-gone errors, exactly as every other uninstall
step does:

```go
if err := i.unitStep(ctx, "stop"); err != nil {
	return nil, fmt.Errorf("systemctl --user stop %s: %w", daemonBinary, err)
}
if err := i.unitStep(ctx, "disable"); err != nil {
	return nil, fmt.Errorf("systemctl --user disable %s: %w", daemonBinary, err)
}
```

### WR-02: Per-app MACR check runs a synchronous a11y D-Bus call under the actor mutex in the keystroke path

**File:** `internal/session/actor.go:1054` (`macrTargetActive`, called from `macrIntercept` → `feedKey` → `HandleKey`)

**Issue:** whenever `macr.apps` is non-empty, every configured-letter press
blocked in `feedKey` calls `a.appid.FocusedApp()` — a live AT-SPI D-Bus
round trip — while holding `a.mu`. The file's own WR-01 discipline states
the keystroke path "must never stall" behind slow IPC (the clipboard rung
was moved off-mutex for exactly this reason), and `docs/ci-runner.md`
documents a hung a11y bridge as a real failure mode ("обходы AT-SPI виснут
до таймаута"). A wedged a11y bus therefore freezes ALL keystroke processing
of the desktop (the engine's `ProcessKeyEvent` serializes on this mutex)
until the D-Bus call times out — an input-plane availability bug, not a
style issue. The degradation WARN can also fire from inside the key path.

**Fix:** mirror the `runClipboardRung` discipline: snapshot `a.appid`
under the mutex, resolve the identity outside it (the decision can be
carried back into a short re-entry), or cache the focused-app identity on a
bounded background refresh keyed on focus events. At minimum, bound the
call with a short timeout context (e.g. 50 ms) so a dead bus cannot park the
keyboard.

### WR-03: chromium-x11 witness recovery poke ignores the char filter its own gate depends on

**File:** `test/e2e/surface.go:517` (`awaitPidInput`); contrast `test/e2e/matrix.go:1356` (`waitMatrixInputPid`, which passes the filter)

**Issue:** `awaitPidInput` — the shared witness loop for GTE, gedit AND
chromium-x11 — pokes `grab-input-pid <pid>` without the `want-chars`
argument. The helper's own documentation
(`test/e2e/focus_helper.py:36-46`) and the matrix-path code both state that
without the filter, Chromium's omnibox (the app's first focusable ENTRY,
~60 chars) "would win every grab" and shadow the page input. On the x11
surface, whose document already shows ambient omnibox focus steals
(matrix-v3.yaml header), the recovery poke can itself move focus to the
omnibox — after which the char-exact gate (`TEXT:chars=0`) can never pass
and the case fails after `surfaceFocusWait`. This affects test reliability
on the exact surface v3 was added to cover.

**Fix:** pass the gate's own expectation to the poke, as the matrix path
already does:

```go
_, _ = runCmd(ctx, "/usr/bin/python3", s.cfg.helper,
	"grab-input-pid", pidStr, strconv.Itoa(wantChars))
```

### WR-04: `closeZenity` can run `exec.Cmd.Wait` concurrently with its own wait goroutine

**File:** `test/e2e/case_m1.go:254-277` (`closeZenity`) with `reapZenity` at 280-290

**Issue:** `closeZenity` spawns a goroutine calling `s.zenity.Wait()`; if
the Enter-then-exit wait times out (`killGrace`), the timeout branch calls
`s.reapZenity()`, which calls `s.zenity.Wait()` **again, concurrently** with
the still-running goroutine (`os/exec` documents `Wait` as a
call-at-most-once API; concurrent calls race on the Cmd's internal error
state). A zenity that ignores Enter (modal state, blocked stdout) hits this
path. Latent data race in the stand's teardown, on the path that is also
the delivery oracle of several cases.

**Fix:** make one Wait owner. Simplest: set `s.zenity = nil` before spawning
the wait goroutine and let `reapZenity` be the only other `Wait` caller, or
guard with a `sync.Once` shared by both call sites:

```go
var waitOnce sync.Once
doWait := func() { waitOnce.Do(func() { _ = s.zenity.Wait() }) }
go func() { doWait(); close(done) }()
select {
case <-done:
case <-time.After(killGrace):
	s.reapZenityWith(doWait) // reap kills; the once-guarded Wait reaps
}
```

## Info

### IN-01: Corrupted dangling doc comments in the e2e stand

**File:** `test/e2e/main.go:101-102` and `test/e2e/main.go:367-369`

**Issue:** editing artifacts left two comments broken: `caseListUsage`'s doc
comment begins with a stray fragment of `run`'s ("run wires the whole
stand: registry, setup, preflight, the case and the"), and `startDaemon`
carries an orphaned sentence ("it only AFTER the desktop state is restored
— a CommandContext kill would fire…") whose subject is missing. Misleading
for the next reader of a safety-critical teardown path.

**Fix:** delete the stray fragments; both functions need one-line doc
comments only.

### IN-02: Component XML version is hardcoded "0.1.0" while the release stamps v1.0.0

**File:** `internal/install/install.go:692`

**Issue:** `renderComponentXML` pins `Version: "0.1.0"`, so the installed
IBus component's version never matches the stamped release version the rest
of D-37 reports (`goswitchd -version`, `status version=`). Diagnostics that
compare component vs daemon version will disagree from v1.0.0 onward.

**Fix:** thread the stamped version through (an `-ldflags -X` receiver in
the install package, or accept it as a parameter from the CLI), with "dev"
as the unstamped fallback — the same honest-dev discipline as
`cmd/goswitchd/version.go`.

### IN-03: xkb engine-name derivation duplicated between product and stand

**File:** `internal/install/install.go:773-790` (`derivedEngine`) vs `test/e2e/main.go:563-575` (`fallbackEngine`)

**Issue:** the same "first `('xkb', layout)` tuple → `xkb:<layout>::eng`"
string surgery lives twice with slightly different guards (the stand's
version can index `rest[start+1:]` at `-1` when `start == -1` guarded only
by the later `end` check — benign in practice because `start >= 0` is
checked after slicing, but Go evaluates the slice first). Divergence risk
for a restore-critical derivation.

**Fix:** keep one implementation (a small exported helper or a
shared internal package); have the stand import the install package's
derivation or vice versa.

### IN-04: `goswitchctl uninstall` silently ignores unexpected positional arguments

**File:** `cmd/goswitchctl/main.go:146-162`

**Issue:** after `fs.Parse(args)`, leftover positionals are never checked —
`goswitchctl uninstall --purge extra` or `goswitchctl uninstall foobar`
execute a normal uninstall instead of failing the usage contract the other
branches enforce (`errUsage`).

**Fix:** `if fs.NArg() > 0 { return fmt.Errorf("uninstall: unexpected argument %q: %w", fs.Arg(0), errUsage) }` (same for `status`).

### IN-05: golangci-lint pinned as floating major "2" while every other tool is exact

**File:** `mise.toml:5`

**Issue:** `golangci-lint = "2"` resolves to the latest v2.x at install
time, so local and CI can drift across weeks (new linters, new defaults)
while `go = "1.23"` and `goreleaser = "2.18.1"` are exact. The
"зелёная итерация" gate (directive 2/4) depends on this tool being the same
everywhere; a new v2 minor can turn a green tree red without any code
change.

**Fix:** pin the exact minor (e.g. `golangci-lint = "2.x.y"` matching the
version currently green in CI), or accept and document the floating pin as
an owner decision.

### IN-06: Install preflight accepts a non-executable goswitchd

**File:** `internal/install/install.go:375-389` (`resolveDaemon`)

**Issue:** the ExecStart preflight is `os.Stat` only — a goswitchd file
without the executable bit passes, the unit is installed and enabled, and
the failure surfaces later as a systemd crash-loop (`Restart=on-failure`)
that `waitRegistration` reports as the generic `errNotRegistered` after 20 s
instead of a named preflight refusal before anything touched the desktop.

**Fix:** check `info.Mode().Perm()&0o111 != 0` (or attempt `syscall.Access`
with `X_OK`) in `resolveDaemon` and fail with the same `errDaemonMissing`
family naming plus an actionable hint.

---

_Positive notes (no action required): the atomic-write + byte-exact
read-back guard (`writeVerified`), the env-carrying write-cache with the
user+system component path, the shape-checked source restore with reported
fallback, the watchdog-limit regression fix pinned headless
(`watchdog_test.go`), the injection-safe dispatch input in `e2e-matrix.yml`,
and the least-privilege workflow split are all correctly implemented and
well tested._

_Reviewed: 2026-09-16T12:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
