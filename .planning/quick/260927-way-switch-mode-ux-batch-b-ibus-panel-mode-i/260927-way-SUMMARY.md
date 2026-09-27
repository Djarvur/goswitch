---
phase: 260927-way-switch-mode-ux-batch-b
plan: 01
subsystem: engine + session + hotkey + config + install + docs
tags: [ibus-panel, mode-indicator, super-space, mode-switch-chord, macr-witness, install-handover, owner-decision, adr-001]
requires:
  - engine Emitter seam + emitRecorder interceptor tests (02-03)
  - actor flipScript (the single flip point) + feedKey branch order (MACR → combo → mode branches)
  - internal/config strict decode + Defaults() + hotkey.ParseBinding closed tables (03-02)
  - internal/install installState/saveState/restoreSources discipline (04-01)
provides:
  - engine.Property/PropList wire structs + NewModeProperty + EngineDesc.IconPropKey (19 fields, appended LAST)
  - Engine.UpdateModeSymbol (UpdateProperty) + factory RegisterProperties at CreateEngine ("en")
  - session.Options.ModeSwitchChord + modeSwitchChord branch (immediate flip, consume, FSM Reset, MACR witness) + FocusIn panel re-assert + flipScript emits after the mode record
  - hotkey.KeyvalSpace ("space", 0x020, no family bit) + config hotkeys.mode_switch_chord (default super+space, empty = disabled, hot-reload live with the empty-disables fold)
  - install switch-input-source handover (snapshot verbatim → clear → restore on uninstall; reported distro-default fallback)
affects: [GNOME panel indicator (live gate), install-state.json schema (+1 field, backward-compatible read)]
tech-stack:
  added: []
  patterns: [appended-LAST wire field convention, empty-disables snapshot fold, witness-flag MACR interaction, shape-checked gsettings restore]
key-files:
  created: []
  modified:
    - engine/types.go
    - engine/engine.go
    - engine/factory.go
    - engine/wire_test.go
    - engine/emitter_wire_test.go
    - engine/engine_test.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - internal/hotkey/names.go
    - internal/hotkey/names_test.go
    - internal/config/config.go
    - internal/config/config_test.go
    - cmd/goswitchd/main.go
    - internal/install/install.go
    - internal/install/install_test.go
    - internal/ctlsvc/ctlsvc_test.go
    - docs/CONFIG.md
    - docs/SPEC.md
    - README.md
    - README.ru.md
decisions:
  - "REFACTOR commits were folded into the GREEN commits: every task's commit gate is a green `mise run ci`, so the lint-driven extractions (funlen/gocognit/lll/cyclop) landed with the feat commits — the same shape as the 260927-vu8 sibling (RED commit + feat commit per task, no empty refactor commits)"
  - "Task 1 RED extended the engine.Emitter interface, which forced a one-stub UpdateModeSymbol on internal/ctlsvc/ctlsvc_test.go's ctlSink (outside the task's closed file list) — Rule 3 auto-fix, one no-op method, no corpus assertion touched"
  - "The install-state verbatim pin is asserted on the DECODED field value, not raw bytes: json.Marshal HTML-escapes the '<'/'>' of `['<Alt>Shift_L', '<Super>space']`; the restore path reads the field through the same decode, so the value round-trips verbatim"
  - "EngineDesc.IconPropKey assertion in the install mirror test was split RED (XML string pin) / GREEN (wire pin) so the RED commit compiled without the field — compile failures are INVALID_RED, not RED evidence"
  - "FocusIn re-asserts the panel symbol unconditionally (even EN at start): a freshly minted engine context always re-asserts the CURRENT mode — idempotent and self-healing, the factory's initial 'en' registration can never go stale"
metrics:
  duration: 34 min
  completed: 2026-09-27
status: complete
actuals:
  tokens: 22794   # chars/4 over the realized diff (91177 chars, 20 files, +1228/−114)
  tasks: 3
  commits: 6      # MEASURED: git rev-list --count f679b95..HEAD
plan_head_before: f679b9544614526513dae8722ee4b1d9c7604a9d
---

# Quick Task 260927-way: Switch-mode UX batch B (IBus panel mode indicator + Super+Space chord + install handover) Summary

The GNOME input indicator now follows the goswitch script mode (EN/RU) through an engine panel property — `RegisterProperties` at engine creation, `UpdateProperty` on every flip, `icon_prop_key` on the 19-field EngineDesc — while ADR-001 stands untouched; Super+Space is a goswitch-owned, configurable mode-switch chord (`hotkeys.mode_switch_chord`, default `super+space`, empty = disabled) that flips immediately, consumes the press and witnesses the MACR hold; and `goswitchctl install` hands GNOME's `switch-input-source` binding over (snapshot verbatim → clear, restore on uninstall with a reported distro-default fallback). README (both languages) documents the handover.

## Tasks

| # | Task | Type | Commits |
|---|------|------|---------|
| 1 | Panel mode indicator — property wire structs, engine emits, actor emits on flip and focus | auto, TDD | bf48faf (RED) → b33678d (GREEN, refactor folded) |
| 2 | super+space mode-switch chord — hotkey table, config key, actor branch, MACR witness, docs | auto, TDD | bfb7dd2 (RED) → af9a583 (GREEN, refactor folded) |
| 3 | install hands switch-input-source over, uninstall restores it; README handover docs | auto, TDD | 0d036d7 (RED) → 6459762 (GREEN, refactor folded) |

## TDD Gate Compliance (STRICT TDD, directive 1)

Each task followed red → green → refactor with tests run at every step. RED commits carry tests plus the
minimal compile-enabling scaffolding (the vu8 precedent: fields/structs exist unwired so the targets fail on
assertions, never the build). RED evidence below — the canonical `check tdd-red-evidence` verb parses only
node:test TAP, so the go-test RED evidence is recorded here per the orchestrator's instruction.

### Task 1 — RED evidence (`go test -race -count=1 ./engine/ ./internal/session/ ./internal/install/`, exit 1, commit bf48faf)

- `TestEmitters_ModeProperty/update_mode_symbol_ru` / `..._en`: "UpdateModeSymbol(ru) emitted 0 signals, want exactly 1" (the stub emitted nothing)
- `TestEmitters_ModeProperty/factory_registers_the_initial_property`: "CreateEngine emitted 0 RegisterProperties signals, want exactly 1"
- `TestActor_FlipOnSingle`: "panel symbols = [], want exactly [ru en] (one per flip, owner decision 1)"
- `TestActor_FlipEmitsPanelSymbol`: "panel symbols after the first flip = [], want exactly [ru]"
- `TestActor_FocusInReassertsModeSymbol`: "FocusIn symbols at start = [], want exactly [en]"
- `TestActor_FlipAfterWordCorrection`: "panel symbols = [], want exactly [ru]"
- `TestInstall_ComponentXMLMirrorsWireIdentity`: "component XML misses the panel icon key <icon_prop_key>InputMode</icon_prop_key>"
- All pre-existing pins green; scaffolding: Property/PropList structs + NewModeProperty + constants, `Engine.UpdateModeSymbol` stub, Emitter interface extension (+ ctlSink stub)

### Task 2 — RED evidence (`go test -race -count=1 ./internal/hotkey/ ./internal/config/ ./internal/session/`, exit 1, commit bfb7dd2)

- `TestParseBinding_SuperSpace`: `key "space" in binding "super+space": not a name of the closed binding tables`
- `TestDefaults`: `hotkeys.mode_switch_chord = "", want "super+space"`
- `TestLoad_ModeSwitchChord/garbage_refuses_the_whole_document`: "Load(garbage chord) = nil error, want a whole-document rejection (D-33)"
- `TestActor_SuperSpaceChordFlipsMode`: chord press transited (consume false, no mode record)
- `TestActor_SuperSpaceChordKillsPendingTap`: the armed tap decision fired at expiry (want 0 actions)
- `TestActor_SuperSpaceChordWitnessesSuperHold`: ConsumedUpstream = 1 + the consumed-upstream WARN (want zeros, no WARN)
- `TestActor_HotReloadModeSwitchChord`: chord press transited under the default document
- Guards green in RED (by design): gate-off, Mod2-only, combo precedence; scaffolding: Hotkeys.ModeSwitchChord yaml field + Options.ModeSwitchChord field unwired, literal `superSpaceChord()`

### Task 3 — RED evidence (`go test -race -count=1 ./internal/install/`, exit 1, commit 0d036d7)

- `TestInstall_Sequence`: missing `get switch-input-source` and `set switch-input-source []` calls (call-sequence mismatch), missing state field, missing "switch-input-source: cleared" report line
- `TestUninstall_FullRollback`: missing `set ... switch-input-source ['<Alt>Shift_L', '<Super>space']` before state removal, missing restored-value report line
- `TestUninstall_CorruptStateFallsBack` (all 4 subtests): no gsettings set recorded for switch-input-source
- Guard green in RED: `TestInstall_SecondInstallKeepsOriginalBackup` (saveState already skips existing state)

## Verification

- Per-task verify blocks: targeted `go test -race -count=1` green, then `mise run ci` green (build + vet + golangci-lint strict + `go test -race ./...`) after every task; final combined run green (engine / session / install / hotkey / config).
- Engine: Property/PropList/EngineDesc field-order pins (reflect), UpdateProperty shape (Key=InputMode, PropTypeNormal, PropStateUnchecked, sensitive/visible, empty sub-list, Symbol "ru"/"en"), RegisterProperties exactly once at CreateEngine with "en", detached engines quiet, all through the `WithOutgoingInterceptor` seam — no bus needed.
- Session: flip emits exactly one symbol update strictly after the INFO mode record (hook-based cross-stream order pin); FocusIn re-asserts; chord flips immediately + consumes + clean buffer ("abc" token) + kills the pending tap + MACR counters zero with the hold witnessed + corpus unmodified; gate-off / Mod2-only / combo collision / hot-reload empty-disables all pinned.
- Config: defaults, decode round-trip, missing-key loads unchanged (disabled), explicit empty valid, garbage refuses with the field named.
- Install: snapshot (verbatim) + clear + restore + reported distro-default fallback + sacred first backup (byte-intact marker) + full call sequences pinned.
- `mise run ci` green at HEAD 6459762.

## Deviations from Plan

**1. [Rule 3 - Blocking fix] ctlSink stub for the extended Emitter interface**
- **Found during:** Task 1 RED
- **Issue:** adding `UpdateModeSymbol` to `engine.Emitter` broke the compile of `internal/ctlsvc/ctlsvc_test.go`'s `ctlSink` (a file outside Task 1's closed list)
- **Fix:** one no-op `UpdateModeSymbol(string) {}` method in the ctlSink style
- **Files modified:** internal/ctlsvc/ctlsvc_test.go
- **Commit:** bf48faf

**2. [Documented] REFACTOR commits folded into GREEN commits** — the commit gate is a green `mise run ci`, so lint-driven refactors (funlen: `wireFieldWants`/`corruptSwitchCases` extractions; gocognit: `wantModeUpdate`/`wantFactoryRegistration`/`modeSwitchChord` extractions; cyclop; lll rewraps) landed inside the feat commits. Same shape as the 260927-vu8 sibling; no assertion weakened anywhere.

**3. [Documented] Verbatim state assertion reads the decoded value** — `json.Marshal` HTML-escapes `<`/`>`; the test pins the round-tripped field value (the property the restore path actually consumes) instead of raw bytes. Production code unchanged from the plan's pinned shape.

**4. [Cosmetic] config.go committed in RED with pre-gofmt field alignment, corrected in the GREEN commit** (bfb7dd2 → af9a583); caught by the lint gate, no behavior change.

## Known Stubs

None. Every emit path, config knob and install step is fully wired; the stub scan (TODO/FIXME/placeholder/hardcoded-empty flows) came back empty.

## Deferred Live Verification (deferred to orchestrator — the house liveness convention)

1. **Registration survives the 19-field EngineDesc:** after a daemon restart on the live desktop, `ibus list-engine | grep goswitch` still shows both engines and `journalctl --user -u goswitchd` shows no variant/registration errors (T-WAY-03's mitigated disposition — a daemon-rejecting variant would die loudly here).
2. **Indicator:** with the goswitch engine active, `dbus-monitor --session "interface='org.freedesktop.IBus.Engine'"` shows RegisterProperties at engine creation and UpdateProperty with Symbol "ru"/"en" on each flip; the GNOME top-bar indicator visibly flips EN ↔ RU.
3. **Chord live:** press Super+Space in a real field — mode flips (log `mode to=ru`), no space inserted, indicator updates; GNOME's cleared binding produces no engine-disable event (the 20:31 defect shape stays dead).
4. **Handover live:** `goswitchctl install` then `gsettings get org.gnome.desktop.input-sources switch-input-source` → `[]`; `goswitchctl uninstall` restores `['<Alt>Shift_L', '<Super>space']`.

## Self-Check: PASSED

- Files exist: all 20 modified files present at HEAD (verified via `git diff --name-only f679b95..HEAD`).
- Commits exist: bf48faf, b33678d, bfb7dd2, af9a583, 0d036d7, 6459762 — all on `gsd/quick-correction-ux` (verified via `git log`).
- Commits measured from the ledger: 6 (`gsd-plan-head-before-260927-way` = f679b95).
- No tracked-file deletions in the range (`git diff --diff-filter=D` empty).
- `mise run ci` green at HEAD.
