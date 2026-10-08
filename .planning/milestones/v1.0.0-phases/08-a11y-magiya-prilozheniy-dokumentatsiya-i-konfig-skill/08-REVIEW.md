---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
reviewed: 2026-10-05T20:19:42Z
depth: standard
files_reviewed: 15
files_reviewed_list:
  - internal/a11y/a11y.go
  - internal/a11y/a11y_test.go
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/config/load_test.go
  - internal/config/watch_test.go
  - internal/install/install.go
  - internal/install/install_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - cmd/goswitchd/main.go
  - cmd/skillgen/main.go
  - cmd/skillgen/main_test.go
  - mise.toml
  - skills/goswitch-config/SKILL.md
findings:
  critical: 1
  warning: 2
  info: 3
  total: 6
status: issues_found
---

# Phase 8: Code Review Report

**Reviewed:** 2026-10-05T20:19:42Z
**Depth:** standard
**Files Reviewed:** 15
**Status:** issues_found

## Summary

Phase 8 ships the a11y reconciler (`internal/a11y`), the config `a11y` section, the actor fold + sink seam, the installer snapshot/restore of `toolkit-accessibility`, the skillgen generator with its CI sync gate, and the generated `SKILL.md`. The subprocess argv discipline (config app list never reaches argv), the install snapshot three-state restore, the strict-decode/ceiling/compile-gate schema work, and the skillgen determinism are all solid — verified against the 08-CONTEXT/08-PATTERNS contracts. Gates run during review: `go build ./...`, `go vet ./...`, `go test -race -count=1` on all five touched packages (green), `golangci-lint run` via mise (0 issues), and `go run ./cmd/skillgen -check` (byte-in-sync).

The one critical defect is a concurrency contract violation inside the new reconciler: the apply series holds the reconciler's serialization mutex across its subprocess and D-Bus I/O, so `Apply` — which the actor calls synchronously **under the actor mutex** — blocks until the in-flight series finishes. On a degraded desktop that is up to ~3 s of keystroke-path stall, the exact outcome the phase's WR-01 invariant exists to prevent. The finding was proven empirically against a verbatim copy of the file (see CR-01). Two warnings cover shipped documentation that contradicts or over-promises relative to the reviewed code; three infos are minor.

## Critical Issues

### CR-01: Reconciler holds its mutex across subprocess/D-Bus I/O — `Apply` blocks the actor's keystroke hot path for the whole in-flight series (WR-01 violation)

**File:** `internal/a11y/a11y.go:117-128, 136-149, 159-179, 186-197, 239-250` (callers: `internal/session/actor.go:1332-1339`, `internal/session/actor.go:942-950`)

**Issue:** `run()` takes `r.mu` with `defer r.mu.Unlock()` and holds it across the entire `episode()` — the `gsettings get` probe (`read`, deadline `readTimeout = 1500ms`, a11y.go:55), the optional `gsettings set`, and the belt D-Bus round trip (`belt`, deadline `beltTimeout = 1500ms`, a11y.go:73); `warn()` also logs (I/O) under the same lock. `Apply()` (a11y.go:117-128) takes the same `r.mu` to update `desired`. The actor calls `Apply` **synchronously while holding `actor.mu`** — from `pushA11y` inside `applySnapshot` (the per-keystroke fold, `HandleKey` → `applySnapshot` → `pushA11y`) and from `SetA11ySink`'s self-sync. Therefore any `Apply` issued while a series is in flight blocks the fold, and with it **every keystroke in every application**, for the remaining duration of the in-flight series: typically 10-50 ms (healthy gsettings + session-bus dial), up to ~3 s on exactly the wedged-desktop scenario the timeouts exist to bound.

**Proof:** reproduced against a byte-verbatim copy of `a11y.go` (scratch module outside the repo, since `internal/` blocks external imports): with the first series wedged inside its `get` probe, a second `Apply` did not return within a 300 ms observation window (unbounded — it returned only when the probe was released). The existing test corpus pins only the *spawning* `Apply`'s non-blocking (`TestA11y_ApplyDoesNotBlockAndConverges`, a11y_test.go:298-321) — the contended case is untested.

This contradicts three of the phase's own binding statements: the package doc ("the series runs serialized on the reconciler's OWN mutex, outside the actor mutex (WR-01 — no subprocess ever rides the keystroke hot path)", a11y.go:4-8), the `Apply` doc ("fire-and-forget by contract: the fold's caller never waits for a subprocess (WR-01)", a11y.go:111-113), and the actor's legality argument for calling the sink under `actor.mu` ("legal because the sink contract is fire-and-forget (a field write plus its own spawn)", actor.go:1325-1330). The sink contract is broken by the reconciler itself: `Apply` is not "a field write plus its own spawn" — it is a lock acquisition that contends with a mutex held across I/O.

**Fix:** split serialization from state. Keep `r.mu` a short state lock (never held across I/O or logging) and give the series its own `seriesMu` for the one-at-a-time serialization; have `warn` take `r.mu` for the map check and log outside the critical section:

```go
type Reconciler struct {
	runner   Runner
	status   StatusSetter

	mu       sync.Mutex      // guards desired/lastApplied/warned — never held across I/O
	seriesMu sync.Mutex      // serializes apply series — held across the episode's I/O only
	desired     bool
	lastApplied bool
	warned      map[string]bool
}

func (r *Reconciler) Apply(active bool) {
	r.mu.Lock()
	if active == r.desired {
		r.mu.Unlock()
		return
	}
	r.desired = active
	r.mu.Unlock()

	go func() {
		r.seriesMu.Lock()
		defer r.seriesMu.Unlock()
		r.runSeries(active)
	}()
}

func (r *Reconciler) runSeries(desired bool) {
	r.mu.Lock()
	supersededOrNoop := desired != r.desired || desired == r.lastApplied
	r.mu.Unlock()
	if supersededOrNoop {
		return // the stale-series gate, unchanged semantics
	}
	if !desired {
		r.mu.Lock()
		if !r.desired {
			r.lastApplied = false // D-8-4: deactivation touches nothing
		}
		r.mu.Unlock()
		return
	}
	healthy := r.episode() // no state-mutex held across the subprocess/D-Bus legs

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.desired == desired && r.lastApplied != desired {
		r.lastApplied = desired
		if healthy {
			clear(r.warned)
		}
	}
}
```

(`episode` keeps its shape; `warn`/`closeEpisode` move to explicit `r.mu` lock/unlock around the map access only, with the `slog.Warn` call outside it.) Add a contended-case corpus test: `Apply(false)` while the first series' probe is blocked must return within a few milliseconds. The stale-series/convergence semantics of `TestA11y_ApplyDoesNotBlockAndConverges` are preserved by the post-episode `desired` re-check.

## Warnings

### WR-01: SKILL.md manual frame claims `goswitchctl install` intercepts the GNOME `switch-input-source` binding — contradicts install.go's pinned "bindings left untouched" contract

**File:** `skills/goswitch-config/SKILL.md:57-59`

**Issue:** the manual frame says: «Дефолт `super+space` работает потому, что `goswitchctl install` перехватывает биндинг GNOME switch-input-source.» The reviewed installer says the opposite, pinned by its own comment and report: "Install leaves both bindings untouched regardless of the sources count (quick 260930-nxd; the single-source model, owner decision 2026-09-30): install touches ONLY the sources key" (`internal/install/install.go:100-112`, report line `switch-input-source: left untouched (saved for uninstall)` at install.go:965). For a document whose entire audience is AI assistants configuring goswitch, a false mechanical claim about the installer can steer wrong troubleshooting (e.g. "rebind GNOME keys" or "uninstall will remove a binding"). The same wording also feeds the generated row (`install перехватывает биндинг…`, SKILL.md:138) from `docs/CONFIG.md:55,292` — CONFIG.md is outside this review's file list, but the SKILL.md copies are in scope.

**Fix:** reword the frame to the actual mechanism, e.g.: «Дефолт `super+space` работает потому, что goswitch владеет переключением источника ввода (install оборачивает sources в движки goswitch), а демон следует за внешней сменой движка; сам биндинг GNOME install не трогает (снапшотится для отката).» Then align the `mode_switch_chord` rows in `docs/CONFIG.md` and run `mise run skillgen-regen`.

### WR-02: `a11y.apps` regex patterns are validated but never matched at runtime — the documented matching semantics do not exist in the code

**File:** `internal/config/config.go:183-195, 428-453` (docs: `skills/goswitch-config/SKILL.md:156`, source `docs/CONFIG.md:73`)

**Issue:** phase decision D-8-5 (08-CONTEXT) prescribed regex-substring matching for the app list, consistent with `autocorrect.apps_blocklist`, and the shipped docs advertise exactly that («приложения, которым goswitch включает магию: совпадение подстрокой (`chromium` находит `org.chromium.Chromium`)…»). In the implementation the patterns have **no runtime consumer**: the only read of the section is `snap.A11y.Active()` (actor.go:1333 — `Enabled && len(Apps) > 0`), and the a11y package deliberately never sees app names (a11y.go:11-15). The list's entire runtime effect is non-emptiness of a boolean gate: `a11y.apps: ["x"]` and a 64-entry anchored list behave identically — the global key is enabled at daemon start regardless of whether any listed app ever runs. A user (or AI) following SKILL.md will expect the magic scoped to the listed applications. The prose section `docs/CONFIG.md` §«Секция a11y: применение и откат» describes the global mechanism honestly, so the docs contradict each other too.

**Fix (owner decision required, per D-8-8):** either implement the documented matching (e.g. arm the reconcile on a listed app's focus, like the MACR scope — noting D-8-4 makes per-app *dis*abling impossible for this global lever), or — more honest given the lever's global nature — state plainly in the CONFIG.md row and SKILL.md that the list is an on/off gate (non-empty = the fixed global magic is applied at start/reload) and the regex grammar is validated but not matched, then regen the region. Silence here ships a doc that over-promises scoping the daemon cannot do.

## Info

### IN-01: Degraded activation marks `lastApplied = true` — a failed set is never retried while the config stays unchanged

**File:** `internal/a11y/a11y.go:159-178`

**Issue:** `episode()` sets `r.lastApplied = true` unconditionally, including when the `gsettings set` failed to start (`reasonKeySet`) — the comment documents the intent ("lastApplied lands true whatever the degradation"). Consequence: if the one activation episode fails (e.g. `gsettings` momentarily unavailable at daemon start), the desired state is recorded as converged and never re-applied until the a11y active state toggles via a config edit. The warn-once budget keeps the failure visible exactly once. Deliberate per the no-retry design, but worth a follow-up todo if retry-on-recovery is ever wanted.

**Fix:** none required now; if auto-recovery is desired, schedule one bounded re-episode when an episode closed unhealthy (a timer, not a loop), keeping the diff gate.

### IN-02: skillgen rewrites SKILL.md even when the rendered region is byte-identical

**File:** `cmd/skillgen/main.go:111-117`

**Issue:** without `-check`, `run()` always `os.WriteFile`s the skill file, even when `updated` equals the committed bytes (mtime churn; `git status` unaffected). The byte-equality data is already in hand in `-check` mode — the write path just never compares.

**Fix:** read the current file first and skip the write when `bytes.Equal`:

```go
current, err := os.ReadFile(skillPath)
if err != nil || !bytes.Equal(current, updated) {
	if err := os.WriteFile(skillPath, updated, filePerm); err != nil { ... }
}
```

### IN-03: install-over-install performs the a11y-key `gsettings get` whose result is always discarded

**File:** `internal/install/install.go:565-571` (stat guard at 573-578)

**Issue:** `saveState` reads the a11y key (and the two bindings) before checking whether the state file exists; on an install-over-install the early return then discards all three reads. The a11y read follows the pre-existing bindings shape, so this is consistency, not a regression — but the new read adds a third wasted subprocess on the upgrade path.

**Fix:** move the three snapshot reads below the state-file stat guard (the `return prior, nil` idempotent-backup branch), preserving the refusal gate before any mutation.

---

_Positive notes: the argv discipline end-to-end (config list → boolean gate only, `TestA11y_FixedArgvOnlyLiterals`), the install three-state restore with no fabricated fallback (`savedA11yState`, install.go:1117-1135) and its corpus, the strict-decode/ceiling/blank/compile schema work with field+index errors, and the skillgen hard-error determinism with the committed-golden CI gate all match the phase contracts exactly._

_Reviewed: 2026-10-05T20:19:42Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
