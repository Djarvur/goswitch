# Phase 4: Поставка и приёмка - Pattern Map

**Mapped:** 2026-09-15
**Files analyzed:** 19 (new + modified)
**Analogs found:** 16 / 19 (3 with no in-repo analog — RESEARCH.md covers them)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/install/install.go` (+ uninstall/selfcheck, NEW package) | service | file-I/O + subprocess orchestration | `internal/clipboard/clipboard.go` | role-match (Runner seam + subprocess discipline) |
| `internal/install/*_test.go` (NEW) | test | — | `internal/clipboard/clipboard_test.go` | exact (fakeRunner recording seam) |
| `cmd/goswitchctl/main.go` (MODIFY: + install/uninstall/selfcheck) | controller (CLI) | request-response | `cmd/goswitchctl/main.go` (itself: switch dispatch) | exact |
| `cmd/goswitchd/main.go` (MODIFY: `var version` + `--version`) | controller (CLI) | request-response | `cmd/goswitchd/main.go` (itself: flag pattern) | exact |
| `internal/ctlsvc/ctlsvc.go` (MODIFY: version token in Status) | service | request-response | `internal/ctlsvc/ctlsvc.go` (itself: renderStatus) | exact |
| `.goreleaser.yaml` (NEW) | config | — | none in repo → RESEARCH.md "Code Examples" | none |
| `.github/workflows/release.yml` (NEW) | CI config | event-driven (tag `v*`) | `.github/workflows/pr-sanity.yml` + `.github/workflows/e2e-matrix.yml` | role-match |
| `test/e2e/perf.go` (NEW) | e2e harness (perf mode) | batch measurement | `test/e2e/case_m1.go` + `test/e2e/main.go` | role-match |
| `test/e2e/focus_helper.py` (MODIFY: resident/event witness) | utility | event-driven | `test/e2e/focus_helper.py` (itself: subcommand dispatch) | exact (extension of) |
| `test/e2e/cases/matrix-v3.yaml` (NEW) | test data | — | `test/e2e/cases/matrix-v2.yaml` | exact (v1→v2 precedent) |
| `test/e2e/matrix.go` (MODIFY: surface vocabulary) | test runner | — | `test/e2e/matrix.go` (itself: validate) | exact |
| `test/e2e/surface.go` (MODIFY: gedit + chromium-x11 drivers) | test driver | — | `test/e2e/surface.go` (itself: startGTE/startChromium) | exact |
| `test/e2e/preflight.go` (MODIFY: + gedit check) | test preflight | — | `test/e2e/preflight.go` (itself: checkGnomeTextEditorLaunch) | exact |
| `test/e2e/main.go` (MODIFY: + perf/install-cycle case registry) | test harness | — | `test/e2e/main.go` (itself: pickCase) + `test/e2e/case_ctl.go` | exact |
| `mise.toml` (MODIFY: + e2e-perf / e2e-matrix-v3 / install-cycle / goreleaser pin) | config | — | `mise.toml` (itself) | exact |
| `.github/workflows/e2e-matrix.yml` (MODIFY: v3 + double run) | CI config | event-driven (dispatch) | `.github/workflows/e2e-matrix.yml` (itself) | exact |
| `README.md` (REWRITE) + `README.ru.md` (NEW) | docs | — | `README.md` (current) + `docs/CONFIG.md` (doc conventions) | role-match |
| `docs/ACCEPTANCE.md` (NEW, location at discretion) | docs (UAT checklist) | — | `.planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-UAT.md` | partial (format precedent) |
| `dist/systemd/user/goswitchd.service` (READ-ONLY reference; install renders its own copy) | config | — | itself | exact |

Housekeeping (not code, no analog needed): `.planning/WINDOWS.md` ledger reconciliation (RESEARCH Open Question 1).

## Pattern Assignments

### `internal/install/` (service, file-I/O + subprocess orchestration)

**Analog:** `internal/clipboard/clipboard.go` — the project's canonical subprocess package: pinned binary names, a Runner seam for TDD, a package-level deadline on every call, and an `execRunner` production backing. All four install subprocess families (`ibus write-cache`, `systemctl --user`, `gsettings`, `ibus restart`) follow this shape.

**Imports pattern** (`internal/clipboard/clipboard.go:11-19`):
```go
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)
```

**Runner seam pattern** (`internal/clipboard/clipboard.go:43-69`) — the unit corpus drives install through a recording fake; extend the signature with an `env []string` parameter (clipboard needs no env; install MUST set `IBUS_COMPONENT_PATH` on every write-cache — RESEARCH Pitfall 1):
```go
// Runner executes one wl-clipboard subprocess: the seam the unit corpus
// drives with a recording fake (argv + stdin are the observable surface),
// backed in production by os/exec.CommandContext.
type Runner func(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error)

func New(opts ...func(*Clipboard)) *Clipboard {
	c := &Clipboard{run: execRunner}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithRunner replaces the subprocess runner — the test seam.
func WithRunner(r Runner) func(*Clipboard) { ... }
```

**Deadline-per-call pattern** (`internal/clipboard/clipboard.go:134-139`) — a wedged `systemctl`/`ibus` must fail the named step, not hang the terminal (the e2e runCmd live finding, `test/e2e/main.go:513-518` comment):
```go
func (c *Clipboard) call(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()
	return c.run(ctx, name, args, stdin)
}
```

**execRunner production backing** (`internal/clipboard/clipboard.go:151-171`) — stdout/stderr captured, error carries argv + stderr so the CLI verdict names the failing step:
```go
func execRunner(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	...
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return out.Bytes(), nil
}
```

**Env-override helper precedent** (`test/e2e/surface.go:248-258`) — `withEnv` removes later duplicates so a stale assignment cannot resurrect in the child; reuse this exact discipline when the install runner appends `IBUS_COMPONENT_PATH`:
```go
func withEnv(key, value string) []string {
	prefix := key + "="
	env := make([]string, 0, len(os.Environ())+1)
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, prefix) {
			env = append(env, kv)
		}
	}
	return append(env, prefix+value)
}
```

**gsettings save/set/restore pattern** — copy the stand's proven trio from `test/e2e/main.go`:
- `snapshotDesktop` (`test/e2e/main.go:412-424`) — `gsettings get org.gnome.desktop.input-sources sources` (and `current`) via runCmd; constants at `:42-46`.
- `restoreGsettings` (`test/e2e/main.go:432-450`) — restores ONLY keys that differ (a value-identical set still triggers GNOME's input-source re-evaluation and live-unsets the global engine).
- `verifyRestored` (`test/e2e/main.go:456-475`) — machine-verifies the restore readback; D-42 uninstall tests must do the same for restored sources.
- `fallbackEngine` (`test/e2e/main.go:499-511`) — derives a safe xkb fallback from the saved tuple; pairs with the ASVS V5 rule (shape-validate before restore, fall back to `[('xkb','us')]`).

**File-write discipline** — repo convention is explicit perms via `os.WriteFile` with `0o600` (e.g. `test/e2e/case_select.go:95` `const configFilePerm = 0o600`). Install's XML/unit/state files use explicit 0644/0600 + atomic write-then-rename (RESEARCH ASVS V14); no in-repo atomic-write helper exists yet — write one in `internal/install`.

**State file location** (RESEARCH Pitfall 8, binding): `~/.local/share/goswitch/` — NEVER `~/.config/goswitch/` (D-42 preserves that dir on default uninstall; a backup there outlives the uninstall that needs it).

---

### `internal/install/*_test.go` (test)

**Analog:** `internal/clipboard/clipboard_test.go` — the recording-fake pattern for the Runner seam.

**fakeRunner pattern** (`internal/clipboard/clipboard_test.go:32-97`) — records argv (and now env) as the observable surface, answers with configured replies, and can emulate a wedged subprocess honoring context cancellation:
```go
type clipCall struct {
	name           string
	args           []string
	stdin          []byte
	ctxHasDeadline bool
}

type fakeRunner struct {
	mu      sync.Mutex
	calls   []clipCall
	reply   []byte
	exitErr error
	hangUntilCtx bool
}
```
For install tests: record `env` in the call struct and assert EVERY recorded `ibus write-cache` invocation carries `IBUS_COMPONENT_PATH` (Pitfall 1 is a unit-testable contract). Use `t.TempDir()` as the fake `$HOME` for XML/unit/state rendering tests (RESEARCH Validation Architecture).

---

### `cmd/goswitchctl/main.go` (controller, request-response)

**Analog:** itself — the new subcommands slot into the existing switch; update `errUsage` (`:39`) to the new grammar.

**Dispatch pattern** (`cmd/goswitchctl/main.go:57-84`):
```go
func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errUsage
	}
	ctx, cancel := context.WithTimeout(ctx, ctlCallTimeout)
	defer cancel()

	switch args[0] {
	case "status":
		return runStatus(ctx, args[1:])
	case "reload":
		reply, err := callMethod(ctx, "ReloadConfig")
		...
	default:
		return fmt.Errorf("unknown subcommand %q: %w", args[0], errUsage)
	}
	return nil
}
```
Notes for the planner:
- install/uninstall/selfcheck do NOT dial D-Bus — they must not inherit `ctlCallTimeout` (5 s is too small for `ibus restart` + unit start); give them their own deadline budget the way `runStatus` owns its FlagSet (`:95-100`).
- Sub-flag parsing precedent: `runStatus` (`:95-100`) — `flag.NewFlagSet("uninstall", ...)`, `fs.Bool("purge", false, ...)` (D-42).
- Keep the client thin: the literals-stay-here comment (`:22-24`) — install logic lives in `internal/install`, the CLI only wires args to it (same split as goswitchd's `run`/packages).
- Exit contract: non-nil error → exit 1 (`:46-53`); selfcheck prints per-check green/red with fix hint BEFORE failing (preflight pattern below).

---

### `cmd/goswitchd/main.go` (controller) — version stamping (D-37)

**Analog:** itself — flags block.

**Flag pattern** (`cmd/goswitchd/main.go:25-35`):
```go
func main() {
	debug := flag.Bool("debug", false, "enable key tracing (logs every keystroke — passwords become visible)")
	configPath := flag.String("config", "", "path to the YAML config file (empty = built-in defaults)")
	flag.Parse()
	...
}
```
Add package-level `var version = "dev"` (goreleaser's default ldflags stamp `main.version`/`main.commit`/`main.date` — RESEARCH Pattern 3; `go install` builds stay unstamped and honestly show "dev"). A `-version` flag prints and exits 0. Same `var version` in `cmd/goswitchctl/main.go` if its version is surfaced.

**Version wire-through decision (planner):** `engine.NewComponent` hardcodes `Version: "0.1.0"` (`engine/types.go:98`, and `:128` per EngineDesc) — either thread the stamped value through `engine.Config` or consciously leave the wire component on a scheme version (RESEARCH Pattern 3 leaves it open). Do NOT change wire-struct field order (`engine/types.go:5-8` — declaration order is the wire contract).

---

### `internal/ctlsvc/ctlsvc.go` (service) — version in Status

**Analog:** itself — `renderStatus` (`:132-157`) is the single-line key=value canon; a `version=` token joins the fixed-token list. The CLI's `parseStatusLine` (`cmd/goswitchctl/main.go:159-173`) already parses unknown keys generically, so no client change is needed beyond display.
```go
tokens := []string{
	"mode=" + st.Mode,
	"corrections_done=" + strconv.Itoa(st.CorrectionsDone),
	...
}
```
Constraints: token grammar is space-free pairs (`:125-131` comment); counts and states only, never user text (D-20/D-21). `session.Status` (`internal/session/actor.go`) gains the version field — the snapshot provider interface (`:49-51`) stays as-is.

---

### `test/e2e/perf.go` (e2e harness, batch measurement) — D-43/D-44/D-45

**Analogs:** `test/e2e/case_m1.go` (witness/injection primitives + timing knobs) and `test/e2e/main.go` (runCmd, watchdog, registry).

**Timing-knob convention** (`test/e2e/case_m1.go:13-24`) — named constants at file top, never sprinkled:
```go
const (
	minKeyEvents   = 6
	keyWait        = 10 * time.Second
	...
	witnessPoll    = 100 * time.Millisecond
)
```
**Hard constraint (RESEARCH Pitfall 3):** do NOT reuse `witnessPoll` (100 ms) or the per-poll python spawn (`focusWitness`, `test/e2e/case_m1.go:290-297` — spawns `/usr/bin/python3 focus_helper.py` per call, 60–80 ms each); the perf run needs the resident/event witness mode in `focus_helper.py` first. Perf's own poll/quantum constants belong in this knob block, and the methodology names them (D-44).

**Injection + t0** (`test/e2e/main.go:631-666`): `injectText`/`injectKeys`/`pressKey` are the injection primitives — t0 stamps immediately before the ydotool call (RESEARCH Pattern 4). `runCmd` (`test/e2e/main.go:519-532`) is the subprocess runner with argv+stderr in the error.

**Watchdog** (`test/e2e/main.go:172-192`): a perf run (N repeats) rides the same `runCaseWatchdog` deadline discipline — raise the budget via a perf-specific constant, never remove the watchdog (a live-session stand must never strand the owner's desktop).

**Memory oracle (D-45):** single `/proc/<pid>/status` VmHWM read at run end — use the parse sketch from RESEARCH Pattern 4 verbatim (`strings.CutPrefix(line, "VmHWM:")`, kB). No repo precedent needed; no dependency (sort + index for p50/p95/p99 per RESEARCH "Don't Hand-Roll").

**Daemon PID:** the stand owns `s.daemon` (`test/e2e/main.go:332-343`) — the VmHWM read targets `s.daemon.Process.Pid`; pid-based identity only, never pkill/pgrep `-f` (RESEARCH anti-pattern, live hit).

---

### `test/e2e/focus_helper.py` (utility, event-driven) — resident/event witness

**Analog:** itself — new subcommand joins the dispatch; the module docstring subcommand table (lines 1-57) is the contract documentation precedent (each mode documented with its live-verified caveats).

**Conventions to keep:**
- `INPUT_ROLES`, `applications()`, `walk()`, `char_count()` (`:66-110`) are the shared tree primitives a `text-changed` event listener reuses (`Atspi.EventListener.register` + main loop).
- Exit codes: 0 success, 1 named failure, 2 usage error (`:56`).
- Resident mode changes the process shape (start once, stream events) — the Go side reads its stdout line-protocol style; keep one line = one observation with a timestamp so t1 is the event arrival (RESEARCH Pattern 4).
- Every caller uses `/usr/bin/python3` (PATH python3 is linuxbrew without gi — `:53-54`).

---

### `test/e2e/cases/matrix-v3.yaml` (test data) — D-47

**Analog:** `test/e2e/cases/matrix-v2.yaml` — the v1→v2 precedent: v2 opened with a comment block naming the plan, the schema delta vs v1, and documented exclusions (`:1-27`), then multi-doc cases separated by `---`.

**Case shape** (`test/e2e/cases/matrix-v2.yaml:27-42`):
```yaml
---
name: phrase-en-ru
surface: zenity
mode: en
steps:
  - {type: "ghbdtn ghbdtn"}
  - {tap: triple}
expect_text: "привет привет"
```
v3 rules (RESEARCH Pattern 5): `matrix-v2.yaml` stays FROZEN (untouched, accepted Phase 3 baseline); v3 adds `gedit` and chromium-x11 cases with spike-pinned vocab only — a name outside the closed vocabulary fails case validation, never the live run (`test/e2e/matrix.go:310-316`).

---

### `test/e2e/matrix.go` (test runner) — vocabulary extension

**Analog:** itself — four touch points for a new surface:

1. **Surface constants + list** (`test/e2e/matrix.go:42-52`):
```go
matrixSurfaceZenity = "zenity"
matrixSurfaceGTE    = "gnome-text-editor"
surfaceChromiumName = "chromium"

func matrixSurfaces() []string {
	return []string{matrixSurfaceZenity, surfaceChromiumName, matrixSurfaceGTE}
}
```
2. **Closed-vocabulary validation** (`test/e2e/matrix.go:314-316`): `slices.Contains(surfaces, c.Surface)` — adding `gedit`/`chromium-x11` to `matrixSurfaces()` is the whole validation change.
3. **Surface opener switch** (`test/e2e/matrix.go:806-836` `openMatrixSurface`) — one case per driver.
4. **AT-SPI app-name map** (`test/e2e/matrix.go:1113-1121` `matrixSurfaceApp`) — the gedit spike must pin its actual AT-SPI app name before this entry is written ("pin the actual, never the assumption" — 02-05 precedent).

Also note `matrixSurfaceReportsAnchor` (`:921-923`) — gedit's anchor-reporting behavior needs the same spike pinning for selection cases.

---

### `test/e2e/surface.go` (test driver) — gedit + chromium-x11

**Analog:** itself — the GTE driver is the template for gedit; the Chromium driver is the template for the x11 variant.

**GTE isolation pattern** (`test/e2e/surface.go:117-139`) — the gedit spike starts here: single-instance avoidance + session-restore avoidance through XDG redirection into the stand's temp dir (gedit has no `--standalone` — RESEARCH Pitfall 6/A4; the spike decides the exact flag set):
```go
func (s *stand) startGTE(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, gteBin, "--standalone", "--ignore-session")
	cmd.Env = withEnv("XDG_DATA_HOME", filepath.Join(s.tmpDir, "gte-data"))
	if err := cmd.Start(); err != nil { ... }
	s.gte = cmd
	return nil
}
```
**Driver lifecycle trio** (copy for gedit and chromium-x11):
- focus gate with grab-recovery: `waitGTEInput` (`test/e2e/surface.go:170-202`) — grace period, then pid-keyed `grab-input-pid` pokes (`:192-196`);
- pid-scoped readback with bounded retries: `readGTEText` (`:210-228`, `gteReadAttempts`/`gteReadRetryDelay` `:152-155`);
- pid-based close doubling as reap: `closeGTE` (`:234-243`).

**Chromium spawn-flags pattern** (`test/e2e/surface.go:63-83`) — the x11 variant is a second spawn with `--ozone-platform=x11` added to exactly this flag set (fresh `--user-data-dir` in `s.tmpDir` keeps PID-based closing possible; `--force-renderer-accessibility` keeps the AT-SPI tree alive). Pin its AT-SPI app name as a new constant next to `chromiumAppName` (`:21-26` — the re-pin-both-together comment discipline).

**surfaceKind extension precedent** (`test/e2e/surface.go:44-52`): new kinds append as `surfaceKind = iota + N` so `case_m1.go` stays untouched; `closeEntrySurface` gains driver-managed no-op cases (`test/e2e/case_m1.go:128-147`).

---

### `test/e2e/preflight.go` (test preflight) — + gedit check, selfcheck model

**Analog:** itself — D-41 says selfcheck follows the e2e-preflight fail-fast style; the Go-side gedit preflight is this exact block.

**Checks-table pattern** (`test/e2e/preflight.go:30-52`) — named check + one actionable diagnostic line per failure:
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
**Binary-presence check** (`test/e2e/preflight.go:159-174` `checkGnomeTextEditorLaunch`): `exec.LookPath` + one fix-hint line ("install gedit" — RESEARCH Pitfall 6 says gedit is NOT installed; this check is the warning-sign detector) + spawn/witness/close happy path.

**Live engine-registration probe** (`test/e2e/preflight.go:178-205` `listActiveEnginesContain` + `:210-228` `bodyMentions`) — selfcheck's "движок реально зарегистрирован" step reuses this ListActiveEngines dial-over-`engine.Discover()` verbatim (it is package-level, not stand-bound). The "component visible in registry" step uses `ibus list-engine` output instead — and per RESEARCH Pitfall 2, cache-visible ≠ daemon-visible: selfcheck's repair path re-runs the env-carrying write-cache before failing (D-41).

---

### `test/e2e/main.go` (test harness) — registry + install-cycle case

**Analog:** itself + `test/e2e/case_ctl.go`.

**Case registry** (`test/e2e/main.go:196-233` `pickCase`): add `install-cycle` (and perf if it is a `-case` rather than a flag — planner call). Keep the registry map built per call, update `caseListUsage()` (`:87-93`) in the SAME change (the usage string is hand-maintained, twice — `:88-92` and `:225-229`).

**Install-cycle e2e (RESEARCH Pitfall 9):** the case drives the freshly built `goswitchctl` binary the way `case_ctl.go` drives the daemon+CLI: spawn the built binary, assert on exit codes and stdout; `case_ctl.go:19-25` shows the config-document template convention, and `case_ctl.go:117` (`os.WriteFile` + `configFilePerm = 0o600` from `case_select.go:95`) the file-writing convention. Wrap the whole case in the stand's gsettings snapshot/restore — the `-case` path already guarantees this via `run()`'s defer (`test/e2e/main.go:142-148`): teardown + `verifyRestored` downgrade PASS→FAIL on unrestored state. Install/uninstall e2e inherits that contract for free by being a case.

---

### `mise.toml` (config) — new tasks + goreleaser pin

**Analog:** itself.

**e2e-task template** (`mise.toml:97-99` — the v1→v2 precedent for v3, and `:37-39` for any new case task):
```toml
[tasks.e2e-matrix-v2]
description = "e2e: 03-07 YAML case matrix v2 — full phase-3 breadth (...), per-case fresh daemon, PASS/FAIL report, e2e-report.txt (live GNOME session; NOT in ci)"
run = "go run ./test/e2e -matrix test/e2e/cases/matrix-v2.yaml"
```
Keep the description convention: plan ref, one-line scope, the trailer "(live GNOME session; NOT in ci)". Add: `e2e-matrix-v3`, `e2e-perf` (D-43), `e2e-install-cycle` (or `-case install-cycle` per harness shape).

**Tool pin** (`mise.toml:3-5`): add `goreleaser = "2.18.1"` under `[tools]` (RESEARCH Installation: `aqua:goreleaser/goreleaser` registry entry) — one place for versions, CI installs via `mise install` (D-09/D-11).

---

### `.github/workflows/release.yml` (CI config, NEW) — D-38

**Analog:** `.github/workflows/pr-sanity.yml` (structure/mise/permissions) + `.github/workflows/e2e-matrix.yml` (non-push trigger + input + concurrency). No tag-trigger precedent exists in the repo — the trigger block and `permissions: contents: write` are NEW (Pitfall 10: write scope on THIS workflow only; pr-sanity/e2e-matrix stay `contents: read`).

**Header-comment convention** (both workflows open with a purpose/design comment block — see `e2e-matrix.yml:1-21`, in Russian, citing decisions).

**Workflow skeleton to copy** (`pr-sanity.yml:36-50`):
```yaml
      - uses: actions/checkout@v4
      - name: Install mise
        run: curl https://mise.run | sh
      - name: Add mise to PATH
        run: echo "$HOME/.local/bin" >> "$GITHUB_PATH"
      - name: Install tools from mise.toml
        run: mise install
```
**Trigger + concurrency shape** (`e2e-matrix.yml:24-38`): non-default triggers are declared with a rationale comment; concurrency group named per concern, `cancel-in-progress` chosen deliberately. Release: `on: push: tags: ['v*']`, concurrency group e.g. `release`, then `mise exec -- goreleaser release` (goreleaser pin flows from mise.toml — Pitfall 10).

---

### `.github/workflows/e2e-matrix.yml` (CI config, MODIFY) — v3 + double run (D-48)

**Analog:** itself — the matrix path input (`:26-30`) and the env-guarded task selection (`:69-77`, MATRIX via env, never shell-interpolated — WR-04) extend with the v3 path; the double-run gate is two consecutive `mise run` steps (exit ≠ 0 fails the job) under the same `e2e-gnome` concurrency group (`:36-38`, no cancel-in-progress). Session freshness (D-48 definition) is an open operational question (RESEARCH A8/Open Question 2) — planner decides the preflight/assertion shape.

---

### `README.md` (REWRITE) + `README.ru.md` (NEW) — D-50/D-46

**Analog:** current `README.md` (structure to replace: status line `:11`, dev-run block `:29-34` — the hand `cp` steps that `goswitchctl install` now replaces, per RESEARCH State of the Art) and `docs/CONFIG.md` (public contract doc conventions).

**Content obligations:** bilingual pair with EN primary (`README.md`) + RU (`README.ru.md`) synced per release cycle (D-50); perf table p50/p95/p99 + RSS peak updated per release (D-46); install section documents `goswitchctl install` as ONE command including the env-carrying write-cache (never a bare `ibus write-cache` step — RESEARCH Pitfall 1 anti-pattern for docs); keep the `-debug` privacy warning (`README.md:13-22`) — log privacy is a standing owner directive.

---

### `docs/ACCEPTANCE.md` (docs, NEW — location at discretion) — D-49

**Analog:** `.planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-UAT.md` — the checklist format that closed Phases 1–3: numbered items, each with `expected:` / `result:` / `note:` fields, closed by a Summary block (total/passed/issues/pending). The Phase 4 repo checklist walks: install from README on a live desktop → live gesture set → uninstall; pass/fail recorded by the owner in one session (SPEC §7.2 mechanic).

---

### `dist/systemd/user/goswitchd.service` (READ-ONLY reference)

**Analog:** itself — install renders its own copy into `~/.config/systemd/user/`; the tracked template is the semantic reference, not the bytes to copy: keep `PartOf=graphical-session.target` / `After=org.freedesktop.IBus.session.GNOME.service` / `Restart=on-failure` / `WantedBy=graphical-session.target` (`:10-23`), but replace the dev `ExecStart=%h/go/bin/goswitchd` with the RESOLVED ABSOLUTE path of the running binary (RESEARCH ASVS V14: no `$PATH` lookup at unit start, no shell interpolation). The header comment's "installed BY HAND" line becomes obsolete the moment install ships.

## Shared Patterns

### Subprocess seam + per-call deadline
**Source:** `internal/clipboard/clipboard.go:43-69, 134-139, 151-171`
**Apply to:** all of `internal/install` (write-cache/systemctl/gsettings/ibus-restart calls), selfcheck probes.
Every external command: pinned binary name, flags only (never user content in argv), context deadline, error carrying argv+stderr. Extend the Runner signature with env (IBUS_COMPONENT_PATH — Pitfall 1).

### Fail-fast check with fix hint
**Source:** `test/e2e/preflight.go:30-52` (checks table), `:59-66` (diagnostic with parenthesized fix hint)
**Apply to:** `goswitchctl selfcheck` (D-41) — every red item prints one actionable line; the "component visible" check repairs by re-running the env-cache before failing.

### Snapshot → mutate → restore → machine-verify
**Source:** `test/e2e/main.go:412-475` (`snapshotDesktop`/`restoreGsettings`/`verifyRestored`) + `run()` defer `:142-148`
**Apply to:** install's input-sources save (to `~/.local/share/goswitch/`, Pitfall 8), uninstall's restore (D-40/D-42), and the install-cycle e2e case (inherits the stand's teardown verification for free). Restore only differing keys; validate the saved shape before restoring (ASVS V5).

### Closed-vocabulary YAML validation
**Source:** `test/e2e/matrix.go:284-342` (strict decode `KnownFields(true)`, `slices.Contains` vocab checks, numbered-case errors)
**Apply to:** matrix v3 surface extension — vocabulary extension IS the validation change; gedit/x11 names enter only after spike pinning.

### Registry + usage string discipline
**Source:** `test/e2e/main.go:196-233` (`pickCase`) — registry map AND `caseListUsage()` updated together
**Apply to:** any new e2e case/perf mode; also `cmd/goswitchctl` `errUsage` (`:39`).

### Log/output privacy
**Source:** `internal/ctlsvc/ctlsvc.go:125-157` (renderStatus: counts and states only, never user text)
**Apply to:** install/selfcheck output (D-20/D-21 spread): paths and verdicts at INFO, detail behind debug.

### mise tasks, not scripts
**Source:** `mise.toml` (all tasks; `[tools]` as the single version source)
**Apply to:** `e2e-perf`, `e2e-matrix-v3`, install-cycle task, goreleaser pin. CI calls the same tasks (D-09 — the pr-sanity precedent).

### Workflow conventions
**Source:** `.github/workflows/pr-sanity.yml:22-50` (minimal permissions, mise install, comment header) + `e2e-matrix.yml:24-38` (non-push trigger + concurrency rationale)
**Apply to:** `release.yml` (NEW: tag trigger, `contents: write` scoped here only) and the e2e-matrix v3/double-run extension.

## No Analog Found

Files with no close in-repo match — the planner uses RESEARCH.md patterns instead:

| File | Role | Data Flow | Reason / RESEARCH fallback |
|------|------|-----------|---------------------------|
| `.goreleaser.yaml` | config | — | No release tooling in repo. Copy RESEARCH.md "Code Examples → Minimal .goreleaser.yaml" (two builds, `version: 2`, default ldflags stamping) |
| `.github/workflows/release.yml` (trigger/permissions part) | CI config | event-driven | Skeleton analogs exist (mise steps), but tag trigger + `contents: write` + goreleaser invocation are new — follow Pitfall 10 |
| `focus_helper.py` event-listener mode | utility | event-driven | Resident/event `object:text-changed` listener has no precedent (all current modes are per-invocation); RESEARCH Pattern 4 + A2; keep a resident-poller fallback at a named quantum |
| p50/p95/p99 percentile math | utility | transform | No stats precedent; RESEARCH "Don't Hand-Roll": stdlib `sort` + index `s[int(float64(len(s)-1)*p)]` |

## Metadata

**Analog search scope:** whole tracked tree (`cmd/`, `internal/`, `engine/`, `test/e2e/`, `dist/`, `.github/workflows/`, `mise.toml`, root docs) — 43 tracked Go/config files scanned by listing; 16 analog files read.
**Tracked-source gate:** every analog path named above verified against `git ls-files` (all tracked at repo root; no gitignored mirrors involved).
**Pattern extraction date:** 2026-09-15
