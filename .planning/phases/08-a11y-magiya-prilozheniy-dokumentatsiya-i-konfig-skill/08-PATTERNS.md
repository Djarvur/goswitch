# Phase 8: a11y-магия приложений, документация и конфиг-skill - Pattern Map

**Mapped:** 2026-10-05
**Files analyzed:** 16 (7 stream-1 code files + 4 stream-2 docs files + 5 stream-3 files)
**Analogs found:** 15 / 16 (the only no-analog item is the SKILL.md — no *tracked* in-repo precedent exists; format comes from RESEARCH.md)

**Tracked-source gate:** every analog path below was verified with `git ls-files` (tracked). Exception: `.zcode/skills/gsd-ns-context/SKILL.md` is **untracked** (the whole `.zcode/` dir is `??` untracked; not gitignored) — it is cited as a *format reference only*, never a copy-from target. The new `skills/goswitch-config/SKILL.md` MUST be `git add`ed (no `.gitignore` change needed) or the D-8-10 CI sync gate would check an invisible file.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `docs/SPEC.md` (modify — spec-delta §4, D-55) | docs/spec | content | `docs/SPEC.md` §4 itself + CONFIG.md contract prose | self (content work) |
| `internal/config/config.go` (modify — new `a11y` section) | model (config schema) | transform | itself — the `Autocorrect` section (06-04/07-02 canon) | exact |
| `internal/config/config_test.go` (modify — a11y corpus) | test | transform | itself — `TestValidate_Autocorrect*` family | exact |
| `internal/a11y/` (NEW package — reconciler) | service | event-driven + subprocess | `internal/sound/sound.go` (degradation sink) + `internal/clipboard/clipboard.go` (Runner seam) | exact (role+dataflow split across two) |
| `internal/session/actor.go` (modify — fold field + trigger) | service (actor) | event-driven | itself — `SoundSink` seam + `applySnapshot` fold + `refreshACBlocklist` diff gate | exact |
| `cmd/goswitchd/main.go` (modify — wiring) | controller (wiring) | event-driven | itself — `SetSoundSink` install + `FoldAppliedConfig` startup fold | exact |
| `internal/install/install.go` (modify — snapshot field + revert step) | service (installer) | file-I/O + subprocess | itself — `installState`/`saveState`/`restoreSwitchBinding` | exact |
| `internal/install/install_test.go` (modify) | test | file-I/O | itself — `TestInstall_SecondInstallKeepsOriginalBackup` et al. | exact |
| `README.md` (modify — becomes Russian, D-8-7) | docs | content | current `README.ru.md` language register + `README.md` heading map | self (content work) |
| `README.en.md` (NEW — English translation) | docs | content | current `README.md` (this file's content moves there) | self |
| `README.ru.md` (absorbed into README.md / removed) | docs | content | — | rename surface (research «Docs-file rename surface») |
| `docs/CONFIG.md` (modify — russian-first + a11y rows, «six»→«seven») | docs | content | itself — the key table (also the stream-3 generator source) | self |
| `cmd/skillgen/` or `internal/skillgen/` (NEW — CONFIG.md → SKILL.md region) | utility (dev tool) | transform | `layouts/dictgen/main.go` (deterministic committed-golden generator) | role-match |
| `skills/goswitch-config/SKILL.md` (NEW — hand frame + generated region) | docs (skill) | content | `.zcode/skills/gsd-ns-context/SKILL.md` — UNTRACKED, format reference only | no tracked analog |
| `mise.toml` (modify — skill-gen task) | config (build tasks) | batch | itself — `[tasks.dictgen-regen]` (dev-only regeneration precedent) | exact |
| skillgen golden test (NEW) | test | transform | `layouts/dictgen/main_test.go` (byte-determinism + header-marker checks) | role-match |

---

## Pattern Assignments

### `internal/config/config.go` — the `a11y` section (model, transform)

**Analog:** itself — the `Autocorrect` section is the pinned pattern for every new section since 06-04/07-02 (research Pattern 1 names `config.go:117-123, 356-385` verbatim).

**Section struct + zero-value-off doc** (lines 110-123):

```go
// Autocorrect is the automatic wrong-layout correction layer's schema
// section (D-54, opt-in): the global switch, the per-app blocklist of
// regex patterns (the D-53 revision — a match FORBIDS the correction;
// substring RE2 matching, explicit ^…$ anchoring, order carries no
// meaning) and the detector thresholds. The zero value is the OFF
// state: a document without the section decodes disabled, never
// activated (default off everywhere, D-54).
type Autocorrect struct {
	Enabled       bool     `yaml:"enabled"`
	AppsBlocklist []string `yaml:"apps_blocklist"`
	MinWordLen    int      `yaml:"min_word_len"`
	TrigramMargin float64  `yaml:"trigram_margin"`
	TrigramFloor  float64  `yaml:"trigram_floor"`
}
```

The `a11y` analog: same shape — `Enabled bool` + `Apps []string` (RE2 substring regexes, D-8-5). The magic set is FIXED by D-8-6, not per-app configurable — no other keys.

**Ceilings const block** (lines 21-31) — add the a11y ceiling beside these:

```go
const (
	maxTapWindowMs  = 2000
	maxVerifyWaitMs = 2000
	minBackspaceCap = 1
	maxBackspaceCap = 500
	maxMACRApps     = 64

	maxAutocorrectBlocklist = 64
	minAutocorrectWordLen   = 2
	maxAutocorrectWordLen   = 16
)
```

**Error sentinels — package-level, err113 discipline** (lines 61-66):

```go
errAutocorrectBlocklistOverCeil = errors.New("entries, at most 64 allowed")
errAutocorrectBlocklistRegex    = errors.New("must be a valid regular expression")
errAutocorrectBlocklistEmpty    = errors.New("must not be empty or blank — an empty pattern matches everything")
```

**Validate dispatch + section validate with blank-refusal BEFORE Compile, field+index errors, ACTIVE-conditional gating** (lines 244-259, 356-385):

```go
// in Config.Validate():  return c.Autocorrect.validate()  // + a11y.validate()
func (a Autocorrect) validate() error {
	if len(a.AppsBlocklist) > maxAutocorrectBlocklist {
		return fmt.Errorf(
			"autocorrect.apps_blocklist has %d %w",
			len(a.AppsBlocklist), errAutocorrectBlocklistOverCeil,
		)
	}
	for i, pattern := range a.AppsBlocklist {
		if strings.TrimSpace(pattern) == "" {
			return fmt.Errorf("autocorrect.apps_blocklist[%d] %q: %w", i, pattern, errAutocorrectBlocklistEmpty)
		}
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("autocorrect.apps_blocklist[%d] = %q: %w", i, pattern, errAutocorrectBlocklistRegex)
		}
	}
	// ... ACTIVE-conditional checks only when a.Enabled ...
}
```

**Strict decode comes free** — `internal/config/load.go` lines 27-28:

```go
dec := yaml.NewDecoder(f)
dec.KnownFields(true) // D-33: a typo'd key invalidates the whole config.
```

No load.go/watch.go changes are needed for the new section (research: «sections' strict decode and last-good reload need ZERO new code»). Hot reload reaches the section through `Watcher.Snapshot()` (`watch.go:141-143`).

---

### `internal/a11y/` — NEW reconciler package (service, event-driven + subprocess)

**Analog A — degradation sink discipline:** `internal/sound/sound.go`. The package doc is the contract template (lines 1-9):

```go
// Package sound plays the daemon's acoustic feedback ... played by a
// best-effort subprocess ... that can never block, error or
// otherwise alter the gesture that triggered it: every failure is ONE WARN
// per episode and a quiet no-op. The tones know nothing about text — no
// word content ever reaches this package or its logs (D-20/D-21).
```

**Warn-once-per-episode with closed reason vocabulary** (sound.go lines 33-37, 331-345):

```go
const (
	reasonPlayerMissing = "player-missing" // closed vocabulary — reasons only, never user context
	...
)
func (p *Player) warn(reason string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.warned[reason] {
		return
	}
	p.warned[reason] = true
	if err == nil {
		slog.Warn("sound unavailable", "reason", reason)
		return
	}
	slog.Warn("sound unavailable", "reason", reason, "error", err)
}
```

A healthy apply reopens the budget — `closeEpisode` (sound.go 320-325, `clear(p.warned)`).

**Analog B — the two subprocess shapes.** `gsettings get` (read-verify) = the deadline-bounded Runner with captured output; `gsettings set` = the no-pipes fire-and-forget form. Clipboard's Runner seam is the read template (clipboard.go lines 46, 134-139, 163-170):

```go
// Runner executes one wl-clipboard subprocess: the seam the unit corpus
// drives with a recording fake (argv + stdin are the observable surface),
// backed in production by os/exec.CommandContext.
type Runner func(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error)

func (c *Clipboard) call(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()
	return c.run(ctx, name, args, stdin)
}
// execRunner: deadline-bounded ctx kills a wedged process (CommandContext);
// stdout captured byte-exactly; stderr folded into the error.
```

The no-pipes write form is sound.go's `newPlayProc` (lines 84-93):

```go
// newPlayProc builds one player child: the pinned binary with NOTHING
// attached to Stdin/Stdout/Stderr — the no-pipes fork shape (the wl-copy
// precedent, clipboard.go:141-150): a piped descriptor would make a reaper
// wait on a fork-shaped grandchild's write-ends (T-07-08-03).
func newPlayProc(name string, args []string) playProc {
	return exec.Command(name, args...) // Stdin/Stdout/Stderr stay nil — no pipes
}
```

Reaped-goroutine form for fire-and-forget: `go func() { _ = cmd.Wait() }()` (sound.go:192).

**Fixed argv discipline** — the app list NEVER becomes an argument; install.go lines 120-124 are the canon comment:

```go
// The gsettings argument is rendered ONLY from these
// literals and the values parsed by activate.ParseSourceTuples — the raw
// user line never reaches `gsettings set` (ASVS V5 / T-05-02-01, the
// GVariant-injection class).
```

For the reconciler: `gsettings` argv is the fixed schema/key/literal (`org.gnome.desktop.interface`, `toolkit-accessibility`, `true`); `a11y.apps` selects WHETHER the fixed action runs, never an argument.

**Read-verify-then-set** (anti-dconf-churn): mirror install.go's `readSources` trim discipline (lines 482-489):

```go
out, err := i.call(ctx, binGSettings, "get", gsettingsSchema, gsettingsKey)
...
return strings.TrimSpace(string(out)), nil
```

`gsettings get` answers `"true\n"` — trim, compare, set only on diff.

---

### `internal/session/actor.go` — fold addition (service, event-driven)

**Analog:** itself. The fold is `applySnapshot` (lines 1199-1277); the a11y addition mirrors the autocorrect fold (lines 1258-1270):

```go
// The autocorrect section folds live (plan 06-06, the MACR-fold
// precedent): the D-54 defaults are off, so a document without the
// section keeps the boundary byte-as-today; ...
a.opts.AutoCorrectEnabled = snap.Autocorrect.Enabled
a.opts.AutoCorrectBlocklist = snap.Autocorrect.AppsBlocklist
...
a.refreshACBlocklist(snap.Autocorrect.AppsBlocklist)
```

**Diff-gate pattern** — the reconciler trigger must be diff-gated exactly like `refreshACBlocklist` (lines 2188-2198):

```go
// refreshACBlocklist rebuilds the compiled blocklist cache only when the
// raw pattern LIST changed (the macrLettersName parse-cache form) — one
// compile per document edit or SetOptions install, never per boundary.
// The caller holds the mutex.
func (a *Actor) refreshACBlocklist(names []string) {
	if slices.Equal(names, a.acBlocklistNames) {
		return
	}
	a.acBlocklistNames = names
	a.acBlocklist = compileACBlocklist(names)
}
```

**The sink seam** — the a11y sink follows `SoundSink` (lines 882-911) exactly:

```go
// SoundSink is the acoustic-feedback seam (plan 07-08, ...). Defined at
// the point of use; the interface travels with the consumer (the
// AppidSource precedent). The implementation is fire-and-forget by
// contract — it must never block the actor's hot path and must never
// panic; every playback failure is the sink's own best-effort episode
// (one WARN), never an actor error.
type SoundSink interface {
	Flip()
	AutoCorrect()
	SetAutocorrectEvent(event string)
}

// SetSoundSink installs the sound seam — the SetMenuSync mirror (plan
// 07-08). nil = no sink: every tone stays a silent no-op ... A non-nil
// install self-syncs the last folded autocorrect event ...
func (a *Actor) SetSoundSink(s SoundSink) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.soundSink = s
	if s != nil && a.soundACEvent != "" {
		s.SetAutocorrectEvent(a.soundACEvent)
	}
}
```

Lock-ordering rule (pushMenuSync comment, lines 1293-1296): «the sink's setter is a field write behind its own mutex, and the actor.mu → sink.mu ordering is one-way (the sink never calls back)». The reconciler apply runs OUTSIDE the actor mutex (WR-01 — research Anti-Pattern 5).

---

### `cmd/goswitchd/main.go` — wiring (controller, event-driven)

**Analog:** itself. The a11y sink installs beside `SetSoundSink` in the `OnConn` wiring (lines 184-190):

```go
// The sound sink (plan 07-08): the flip tone rides the schema
// constant, the autocorrect tone starts at the startup
// document's effective event — the actor's fold re-pushes every
// changed event to the installed sink (WR-02: the key folds
// live), so no reload branch exists here either (the D-32
// contour — the wiring installs once).
actor.SetSoundSink(sound.New(config.DefaultSoundFlipEvent, cfg.Sound.EffectiveAutocorrectEvent()))
```

The startup apply rides the startup fold (lines 488-496 in `newActor`):

```go
if watcher != nil {
	actor.AttachConfig(watcher)
	// The startup document is in force AT STARTUP (plan 07-06, the
	// restart-survival pin): the fold the first key event would run
	// runs now — the status and the menu report the document truth
	// from the first read, never a stale off ...
	actor.FoldAppliedConfig()
}
```

This is the D-8-3 reconcile-at-start hook: the fold's a11y trigger fires here AND on every debounced reload — no separate wiring branch. Startup wiring reads the startup document (`cfg`) for the initial sink state, exactly as `cfg.Sound.EffectiveAutocorrectEvent()` does above.

---

### `internal/install/install.go` — snapshot field + uninstall revert (service, file-I/O + subprocess)

**Analog:** itself. The state contract (lines 45-50, 56-66, 213-222):

```go
permPrivate = 0o600 // install-state.json — the restore key of the desktop
...
stateDirRel     = ".local/share/goswitch"
stateFile       = "install-state.json"

// installState is the on-disk restore contract between install and
// uninstall: the pre-install gsettings sources string AND the pre-install
// switch chords ..., all verbatim.
type installState struct {
	Sources             string `json:"sources"`
	SwitchInputSource   string `json:"switch_input_source"`
	SwitchInputSourceBw string `json:"switch_input_source_backward"`
}
```

The a11y field joins this struct (research Pattern 3): a `ToolkitAccessibility` pre-install value read in `saveState` via the `gsettings get` call form (lines 520-529), serialized in the `json.Marshal(installState{...})` at lines 537-541, written `writeAtomic(path, data, permPrivate)` (line 545).

**Pitfall-5 guard (must-copy discipline):** `saveState`'s idempotent-backup rule (lines 491-497, 532-533):

```go
// unless a state file already exists, in which case it is LEFT UNTOUCHED:
// the FIRST install's backup is sacred (an install-over-install must never
// save the post-takeover goswitch-only desktop ...)
if _, err := os.Stat(path); err == nil {
	return prior, nil // idempotent backup: the original state stays
}
```

**Uninstall restore — only-if-present/only-if-trusted, fallback REPORTED** (restoreSources, lines 938-963):

```go
// restoreSources puts the saved sources back. The saved value is
// shape-validated BEFORE it reaches gsettings (ASVS V5/T-04-01-02: a
// corrupt or injected state file must never brick the keyboard) — ...
// the substitution is REPORTED, never silent.
func (i *Installer) restoreSources(ctx context.Context) ([]string, error) {
	value, trusted := savedSources(i.path(stateDirRel, stateFile))
	if !trusted {
		value = fallbackSources
	}
	...
}
```

For the a11y key the restore has NO safe fallback (research Pitfall 5): a missing field = WARN + skip («nothing to revert»), never «restore false». Restore runs BEFORE the state file is removed (restoreSwitchBinding comment, lines 979-983).

---

### Test files — TDD corpus conventions

- **Config schema:** `internal/config/config_test.go` — `TestValidate_AutocorrectRanges` (369), `TestValidate_AutocorrectRangeBoundaries` (431), `TestValidate_AutocorrectDormantShapesValid` (452), `TestValidate_EnabledRequiresThresholds` (474). The a11y corpus mirrors this family: dormant-shapes-valid, ceiling, blank refusal, regex compile, field+index error text.
- **Reconciler:** `internal/sound/sound_test.go` names are the template — `TestPlayer_NoPipesForm` (168), `TestPlayer_StartFailureWarnOnce` (369), `TestPlayer_BothMissingSilent` (346): assert argv shapes, warn-once counts, and quiet no-ops through the seam fakes.
- **Install:** `internal/install/install_test.go` — `TestInstall_SecondInstallKeepsOriginalBackup` (588) pins the idempotent-snapshot rule; the a11y revert test pairs with it.
- **Skillgen:** `layouts/dictgen/main_test.go` (~lines 195-217) — the golden test asserts byte-determinism AND header markers:

```go
if !bytes.Equal(first, second) {
	t.Error("two emissions of the same input differ byte-wise")
}
for _, marker := range []string{"DO NOT EDIT", "//go:generate go run ./dictgen", "package layouts", "Lebedev"} {
	if !bytes.Contains(first, []byte(marker)) {
		t.Errorf("emitted source misses %q in the header", marker)
	}
}
```

skillgen's equivalent markers: `goswitch-config:generated BEGIN/END` region sentinels (research Code Examples, SKILL.md shape).

---

### `cmd/skillgen/` (or `internal/skillgen/`) — NEW dev tool (utility, transform)

**Analog:** `layouts/dictgen/main.go`. The doc comment is the discipline statement (lines 1-10):

```go
// Command dictgen bakes the hunspell ru/en dictionaries into committed golden
// Go sources ... It is a dev-side tool (CORR-08): `go generate ./layouts`
// runs it, the generated files are committed, and CI only runs the golden
// tests over the committed data — no hunspell dependency in CI. Parsing
// follows the golden generator discipline ...: hard errors on data
// anomalies, never silent skips, and regeneration is byte-for-byte
// deterministic.
```

Differences for skillgen: input is `docs/CONFIG.md`'s key table (the strictly formatted markdown table, CONFIG.md lines 45-65 — parse THAT, it is the D-8-10 single source), output is the marked region of `skills/goswitch-config/SKILL.md`. Hard-error on malformed table rows, never skip. Gate = a Go test that regenerates and compares against the committed file (research Pattern 5, option (a) — needs no workflow change; pr-sanity already runs `mise run test`).

---

### `skills/goswitch-config/SKILL.md` — NEW (docs/skill)

**No tracked analog.** `.zcode/skills/gsd-ns-context/SKILL.md` (untracked) shows only the frontmatter shape (lines 1-7):

```yaml
---
name: gsd-ns-context
description: "codebase intel | map graphify docs learnings mempalace"
allowed-tools:
  - Read
  - Skill
---
```

The required shape comes from RESEARCH.md (agentskills.io): frontmatter `name` (lowercase, ≤64 chars) + `description`; body = hand-written procedure/diagnostics frame + the generated region between `<!-- goswitch-config:generated BEGIN -->` / `END` sentinels. Language: russian-first (D-8-9/D-8-10, Open Question 4). **Must be git-tracked** (see gate note at top).

---

### `mise.toml` — skill-gen task (config, batch)

**Analog:** itself — the dev-only regeneration task (lines 211-216):

```toml
# Dictionary regeneration (06-02): a dev-only task — it reads the system
# hunspell dictionaries, which exist on the dev machine only. CI never runs
# it: CI гоняет только golden-тесты по закоммиченным данным (CORR-08).
[tasks.dictgen-regen]
description = "regen: dictgen — пересобрать golden-словари/триграммы из системных hunspell (dev-only; CI гоняет только goldens)"
run = "go generate ./layouts"
```

The `skillgen-regen` task follows this form (`go run ./cmd/skillgen` or equivalent). CI coverage rides the golden test inside `mise run test` (pr-sanity.yml runs `mise run build/vet/lint/test/tidy-diff` — verified; no workflow edit needed). Directive 3: mise tasks, never a Makefile.

---

### Stream-2 docs files (`README.md`, `README.en.md`, `docs/CONFIG.md`, `docs/SPEC.md`)

**No code analog — content work.** Concrete anchors for the planner:

- **README heading maps** (verified): `README.md` (410 lines, EN, the v1.1.0 truth): Requirements(17) / Install(24) / Verify(120) / Usage(135) / Autocorrect v1.1(160, incl. «Where autocorrect stays silent» 222, «Switch sounds» 249, GTK3 limitation 268) / Configuration(289) / Performance(323) / Privacy(348) / Troubleshooting(362) / Uninstall(391). `README.ru.md` (217 lines) ends at «Использование»(84)/«Конфигурация»(109) — the whole Autocorrect block and «Input sources and the mode indicator»(README.md:67) exist only in English; that is the D-8-8 gap to close while making README.md Russian.
- **CONFIG.md edit anchors:** line 14-15 «exactly six sections» → seven + the a11y rows in the key table (45-65) + example YAML blocks (91-170); the Privacy section's «enable `org.gnome.desktop.interface toolkit-accessibility` globally (then restart the application)» remedy (lines 224-237) survives the sweep — add the restart-semantics emphasis (research Pitfall 1).
- **SPEC.md:** §4 is the config contract; a section-count/contract change is spec-delta territory (D-55 — spec lands before code). grep-verified: SPEC has no section-count statement.
- **Language register:** russian prose «сочетание клавиш» (never «чорда»), no jargon in owner-facing text; ACCEPTANCE.md / ci-runner.md / SECURITY.md stay EN (D-8-9).

---

## Shared Patterns

### Strict decode + validate-at-load (D-33)
**Source:** `internal/config/load.go:27-28` + `config.go:244-259`
**Apply to:** the a11y section — free via `KnownFields(true)`; all validation in `Config.Validate()`/section validate. Zero-value = off (D-54); old configs load unchanged (research Pitfall 7).

### Regex list discipline (D-53/V5)
**Source:** `internal/config/config.go:356-370`
**Apply to:** `a11y.apps` — ceiling 64, blank refused before `regexp.Compile`, errors name field AND index (`"a11y.apps[%d] = %q: %w"`), RE2 substring with explicit anchoring semantics.

### Fold-diff under mutex, side effects outside it (WR-01/CONF-02)
**Source:** `internal/session/actor.go:1199-1277` (fold), `:2188-2198` (diff gate), `:891-911` (sink seam)
**Apply to:** the a11y trigger — one `Snapshot()` read per fold; `slices.Equal` diff against last-applied desired state; the apply chain (gsettings/IsEnabled) runs on the sink's own serialization, never under `actor.mu`.

### Subprocess shapes
**Source:** `internal/clipboard/clipboard.go:46,134-171` (deadline-bounded read with output) + `internal/sound/sound.go:84-93,192` (no-pipes fire-and-forget write, reaped goroutine)
**Apply to:** `gsettings get` (read-verify — capture output, deadline); `gsettings set` (no-pipes). Fixed argv only; the `a11y.apps` list never reaches argv (T-05-02-01, `install.go:120-124`).

### Warn-once-per-episode degradation
**Source:** `internal/sound/sound.go:33-37,320-345`
**Apply to:** every reconciler failure path — closed reason vocabulary, one WARN per episode, healthy apply reopens the budget, never an error to the caller.

### Snapshot verbatim → restore only-if-present/trusted → substitution REPORTED (ASVS V5, T-04-01-02)
**Source:** `internal/install/install.go:213-222,491-550,938-963`
**Apply to:** the toolkit-accessibility pre-install snapshot + uninstall revert. Missing a11y field in an old install-state.json = WARN + skip (research Pitfall 5 — never fabricate a default).

### Deterministic dev-tool generation with committed goldens
**Source:** `layouts/dictgen/main.go:1-10` + `layouts/dictgen/main_test.go:195-217` + `mise.toml:214-216`
**Apply to:** skillgen — byte-deterministic output, hard errors on anomalies, committed SKILL.md, golden test in `mise run test`, dev-only regen mise task.

### Log discipline (D-20/D-21)
**Source:** `internal/sound/sound.go:1-9,327-345`; `docs/CONFIG.md:212-222`
**Apply to:** all new logs — slog structured key-values (`"reason"`, `"error"`, `"component"`); app names and config list contents never enter logs; config-related logs carry path/applied-value only.

## Anti-Patterns (from RESEARCH.md, binding for plans)

- No `gsettings set` per fold/reload — read-verify-then-set, act only on effective diff (dconf churn).
- Never revert the key when an app leaves the list (D-8-4 — manual vs goswitch-enabled indistinguishable).
- Never compose command lines from config list entries (fixed templates only).
- Never run the reconciler apply under the actor mutex.
- Never treat a missing install-state a11y field as «restore false».

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `skills/goswitch-config/SKILL.md` | docs (skill) | content | No *tracked* in-repo SKILL.md exists (`.zcode/` is fully untracked). Format source: RESEARCH.md (agentskills.io frontmatter: `name` + `description`); untracked `.zcode/skills/gsd-ns-context/SKILL.md` is a shape reference only. Planner must add the file to git explicitly. |
| `README.md` / `README.en.md` restructure | docs | content | Pure bilingual content work; anchors are the heading maps above, not code. |

## Metadata

**Analog search scope:** `internal/` (config, session, sound, clipboard, install), `cmd/goswitchd`, `layouts/dictgen`, `docs/`, repo root (README, mise.toml, .zcode/skills)
**Files scanned:** ~25 (14 read in full or targeted, plus grep-level structural scans of test files, workflows, README headings)
**Pattern extraction date:** 2026-10-05
**All cited analog paths verified git-tracked** via `git ls-files`, except the explicitly flagged untracked SKILL.md reference.
