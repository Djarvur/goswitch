---
phase: 260927-way-switch-mode-ux-batch-b
verified: 2026-09-27T22:15:46Z
status: human_needed
score: 5/5 must-haves verified
covered_files:
  - .planning/quick/260927-way-switch-mode-ux-batch-b-ibus-panel-mode-i/260927-way-PLAN.md
  - .planning/quick/260927-way-switch-mode-ux-batch-b-ibus-panel-mode-i/260927-way-SUMMARY.md
  - README.md
  - README.ru.md
  - cmd/goswitchd/main.go
  - docs/CONFIG.md
  - docs/SPEC.md
  - engine/emitter_wire_test.go
  - engine/engine.go
  - engine/engine_test.go
  - engine/factory.go
  - engine/types.go
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/ctlsvc/ctlsvc_test.go
  - internal/hotkey/names.go
  - internal/hotkey/names_test.go
  - internal/install/install.go
  - internal/install/install_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
covered_digest: "v1:sha256:6d8a11dbf198507f62f5d7057772284d29f77a366504a6af0f13d17b31f2f46e"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: none
  previous_score: n/a
  gaps_closed: []
  gaps_remaining: []
  regressions: []
human_verification:
  - test: "Registration survives the 19-field EngineDesc: after a daemon restart on the live desktop, run `ibus list-engine | grep goswitch` (both engines still listed) and `journalctl --user -u goswitchd` (no variant/registration errors)."
    expected: "Both goswitch engines remain registered; no D-Bus variant/serialization errors in the journal (T-WAY-03: a daemon-rejecting Property/EngineDesc variant would die loudly here)."
    why_human: "Requires the live ibus-daemon accepting the wire payloads on the owner desktop — not observable headlessly."
  - test: "Indicator: with the goswitch engine active, run `dbus-monitor --session \"interface='org.freedesktop.IBus.Engine'\"`, flip the mode (single Shift tap), watch the GNOME top-bar indicator."
    expected: "RegisterProperties at engine creation (Symbol \"en\") and UpdateProperty with Symbol \"ru\"/\"en\" on each flip; the GNOME panel icon visibly flips EN <-> RU."
    why_human: "Panel rendering and the dynamic icon (icon_prop_key resolution) are GNOME-shell visual behavior."
  - test: "Chord live: press Super+Space in a real input field; check the daemon log for the `mode to=ru` record and the field content."
    expected: "Mode flips immediately, no space inserted in the field, indicator updates; GNOME's cleared binding produces no engine-disable event (the 2026-09-27 20:31 defect shape stays dead)."
    why_human: "Real keyboard chord through the live GNOME/IBus stack — unit tests pin the logic, not the desktop wiring."
  - test: "Handover live: run `goswitchctl install`, then `gsettings get org.gnome.desktop.input-sources switch-input-source`; then `goswitchctl uninstall` and read the key again."
    expected: "After install the key reads `[]` (and the state file carries the previous value verbatim); after uninstall it reads `['<Alt>Shift_L', '<Super>space']` again."
    why_human: "Real gsettings/dconf round-trip on the owner desktop; the fakeRunner corpus pins the call sequences, not the live store."
---

# Quick Task 260927-way: Switch-mode UX batch B — Verification Report

**Phase Goal:** Switch-mode UX batch B: IBus panel mode indicator (EN/RU), configurable Super+Space mode-switch chord, install hands the GNOME switch-input-source binding over to goswitch.
**Verified:** 2026-09-27T22:15:46Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

All five must-have truths are implemented, wired end to end, and behaviorally proven by the unit corpus (interceptor-seam wire pins, fakeRunner call-sequence pins, actor state-transition pins). Every behavior-dependent truth (chord state transition, tap-series kill, MACR witness invariant, restore-before-state-removal ordering) has a passing named test — `go test -race -count=1` over the six touched packages and a full `mise run ci` ran GREEN in this verifier's own process. What remains is exactly the plan's own deferred-live list: four checks that need the owner desktop (real ibus-daemon, GNOME panel, real keyboard, real gsettings store) — recorded below as pending human/orchestrator verification per the plan's liveness convention, NOT as failures.

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | GNOME input indicator reflects script mode via engine panel property (RegisterProperties at creation, UpdateProperty per flip, icon_prop_key on descs); ADR-001 stands | ✓ VERIFIED | `engine/factory.go:57-70` emits RegisterProperties (one property, initial "en") inside CreateEngine before the reply; `engine/engine.go:434-447` UpdateModeSymbol emits ifaceEngine+".UpdateProperty" with `NewModeProperty(symbol)`; `engine/types.go:32-52,74-132,203` — EngineDesc.IconPropKey as 19th (appended-LAST) field = modePropKey "InputMode", Property/PropList in the pinned ctor+appended order; `internal/session/actor.go:1252-1263` flipScript emits exactly one symbol update strictly after the INFO mode record, `:520-529` FocusIn re-asserts the CURRENT mode; `internal/install/install.go:201` xmlEngine renders `<icon_prop_key>InputMode</icon_prop_key>`. Tests green: TestEmitters_ModeProperty (shape pins: Key=InputMode, PropTypeNormal, PropStateUnchecked, sensitive/visible, empty SubProps, Symbol "ru"/"en"; factory registers exactly once with "en"), TestActor_FlipEmitsPanelSymbol, TestActor_FocusInReassertsModeSymbol, TestWire_FieldOrder (Property/PropList order), TestInstall_ComponentXMLMirrorsWireIdentity. ADR-001: flipScript touches only `a.mode` — no engine-switch call exists in the actor; goswitch-en keeps Layout "us" (`cmd/goswitchd/main.go:163`). |
| 2 | Super+Space flips mode immediately, consumes the press (no space lands), kills the pending tap series, emits the symbol update; MACR layer unaffected | ✓ VERIFIED | `internal/session/actor.go:1383-1397` modeSwitchChord: zero-binding disable, keyval match, held-mask rule (`ModMask &^ FamilyMask(keyval)` — space has no family bit, so held is exactly Mod4), sets `macrSawLetter = true` (witness), `fsm.Feed(Reset)` (kills the tap series, Pitfall 4), logs `combo kind=mode-switch-chord`, calls flipScript (which emits), returns true (consume). Branch order in feedKey (`:1308-1328`): macrIntercept FIRST, word-layout combo second, chord third, mode branches last. Tests green: TestActor_SuperSpaceChordFlipsMode (consume true, immediate mode record, zero CommitText, buffer clean — "abc" typed after yields token "abc"), TestActor_SuperSpaceChordKillsPendingTap, TestActor_SuperSpaceChordWitnessesSuperHold (SuperIntercepted/ConsumedUpstream stay 0, no WARN; existing MACR corpus unmodified), TestActor_SuperSpaceChordGateOff (zero Options transits; Mod2-only space never matches), TestActor_SuperSpaceChordLosesToCombo. |
| 3 | Chord configurable as hotkeys.mode_switch_chord (closed grammar, default super+space, empty = disabled); hot reload applies live | ✓ VERIFIED | `internal/config/config.go:65` yaml tag, `:113,137` default "super+space", `:174-183` Validate: empty = valid-disabled (missing-key compatibility), non-empty must ParseBinding with the field-named error (D-33); `internal/hotkey/names.go:40` KeyvalSpace 0x020 (ibuskeysyms.h:376 provenance in comment), `:67` closed key table gains "space", `:84-87` deliberate NO family-mask entry with rationale; `internal/session/actor.go:742-755` applySnapshot fold with the empty-disables twist ("" clears the binding, never last-good; parse failure keeps last-good); `cmd/goswitchd/main.go:129-141` chord parsed at startup and fed through SetOptions with warn-and-skip. Load semantics coherent: `internal/config/load.go:20-42` decodes into zero Config (missing key = "" = disabled — matching docs/CONFIG.md "missing = off"), and the watcher exists only when -config is given (`main.go:66`), so the no-config defaults path can never race the fold. Tests green: TestParseBinding_SuperSpace (Keyval 0x020, ModMask Mod4), TestDefaults, TestLoad_ModeSwitchChord (decode/empty-valid/garbage-refuses), TestActor_HotReloadModeSwitchChord (works -> disabled on "" -> restored). |
| 4 | install clears switch-input-source, records previous value verbatim in install-state.json; uninstall restores it (first-install backup sacred; malformed restore falls back to distro default, REPORTED) | ✓ VERIFIED | `internal/install/install.go:87` gsettingsKeySwitch, `:104-105` clearedSwitchBindings "[]" / fallbackSwitchBindings "['<Alt>Shift_L', '<Super>space']" (live-verified comment), `:169` installState.SwitchInputSource json:"switch_input_source", `:417-444` saveState reads BOTH keys, stores both, existing state file left untouched (first-install sacred — an over-install can never save the post-handover cleared value), `:324` clearSwitchBinding in Install right after takeoverSources (unconditional set, idempotent) with report line `:645` "switch-input-source: cleared (previous value saved)", `:369` restoreSwitchBinding in Uninstall BEFORE the state file is removed (`:373`), `:711-747` shape check (trimmed, '[' ... ']'; empty `[]` legitimate) with untrusted -> fallback + REPORTED line `:723`. Tests green: TestInstall_Sequence (get/set call order, state field, report line), TestInstall_SecondInstallKeepsOriginalBackup, TestUninstall_FullRollback (restore before state removal, report names the value), TestUninstall_CorruptStateFallsBack (absent/malformed field -> distro default, reported). |
| 5 | README (both languages) documents the handover and the goswitch chords | ✓ VERIFIED | README.md "Input-source handover" section (lines 56-65): install clears `org.gnome.desktop.input-sources` `switch-input-source`, remembers the previous value verbatim, explains the single-source engine-context disable (the live finding), uninstall restores; chord list: single Shift tap, double Shift, Shift + right Ctrl, Super + Space. README.ru.md "Передача переключения раскладки" (lines 58-67): same content in Russian, same four chords. Both keep their files' existing tone/structure. |

**Score:** 5/5 truths verified (0 present, behavior-unverified)

### Advisory (New Scope, Unevidenced)

None — initial verification (the re-verification evidence gate did not run).

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `engine/types.go` | Property/PropList wire structs, NewModeProperty, EngineDesc.IconPropKey | ✓ VERIFIED | All present in the pinned field orders (ctor order + appended-LAST Symbol/IconPropKey); constants PropTypeNormal/PropStateUnchecked = 0 with header provenance; modePropKey "InputMode", initialModeSymbol "en" |
| `engine/engine.go` | UpdateModeSymbol emit | ✓ VERIFIED | `:434-447` — conn-nil quiet guard, Emit ifaceEngine+".UpdateProperty" with variant-wrapped NewModeProperty, error-logged; Emitter interface `:70-76` gains the fifth primitive |
| `engine/factory.go` | RegisterProperties at CreateEngine | ✓ VERIFIED | `:57-70` — after the export loop, export-failure early return untouched, one-property PropList with initial "en", emitted before the CreateEngine reply |
| `internal/session/actor.go` | Emitter.UpdateModeSymbol; flipScript emits; FocusIn re-asserts; chord branch; Options.ModeSwitchChord + applySnapshot fold | ✓ VERIFIED | All five pieces present and wired (see truths 1-3); Options field documented zero-disabled `:163-173` |
| `internal/hotkey/names.go` + `internal/config/config.go` | "space" key name, KeyvalSpace 0x020; Hotkeys.ModeSwitchChord default super+space | ✓ VERIFIED | See truth 3; family-mask absence deliberate and commented |
| `internal/install/install.go` | installState.switch_input_source, clearSwitchBinding, restore with shape check + fallback | ✓ VERIFIED | See truth 4 |
| `cmd/goswitchd/main.go` + docs (CONFIG.md, SPEC.md §4.1, README.md, README.ru.md) | chord through SetOptions; schema/docs updated | ✓ VERIFIED | main.go:129-141; CONFIG.md schema row (line 27), `space` in the key-name list (line 45), BOTH example documents carry the key (66, 87), Caramba correspondence row (134), missing-key tradeoff stated (53); SPEC §4.1 gains the Super+Space row with the owner-decision 2026-09-27 comment and the stale "проксируется" parenthetical removed; README both languages (truth 5) |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| actor.flipScript | engine Emitter.UpdateModeSymbol -> UpdateProperty signal -> panel property keyed modePropKey -> EngineDesc.IconPropKey | one key string end to end | ✓ WIRED | modePropKey const (types.go:105) is the Key of every NewModeProperty payload and the IconPropKey of every NewEngineDesc; flipScript/modeSymbol feed the symbol; the mirror test pins `<icon_prop_key>InputMode</icon_prop_key>` in the component XML; wire shapes pinned by interceptor tests |
| feedKey chord branch | hotkey.ParseBinding("super+space") | press-side held rule + MACR ordering | ✓ WIRED | Binding{Keyval 0x020, ModMask Mod4} (space: no family bit -> held = Mod4 exactly); branch sits AFTER macrIntercept and AFTER the word-layout combo (collision test green), sets the witness flag before flip |
| config.Hotkeys.ModeSwitchChord | applySnapshot fold -> cmd/goswitchd SetOptions default | empty = disabled, not last-good | ✓ WIRED | Fold at actor.go:742-755; startup path main.go:133-141 warns-and-skips on unparseable; hot-reload test pins works/disabled/restored |
| install saveState | installState.switch_input_source -> restoreSwitchBinding shape check -> gsettings set/clear | verbatim round-trip | ✓ WIRED | saveState reads `gsettings get ... switch-input-source` and stores verbatim; clearSwitchBinding sets `[]`; restore re-reads, shape-checks, sets saved or reported fallback — call sequences pinned by the fakeRunner corpus |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| Engine.UpdateModeSymbol | symbol | actor's live script-mode state machine (`a.mode`) | yes — flips with every mode change | ✓ FLOWING |
| modeSwitchChord | c (Binding) | config document / built-in defaults via ParseBinding | yes — hot-reload folds live | ✓ FLOWING |
| clearSwitchBinding / restoreSwitchBinding | value | production `gsettings get/set` subprocess (fakeRunner in corpus) | yes — saved pre-install value verbatim | ✓ FLOWING |
| NewEngineDesc.IconPropKey | modePropKey | package constant (identity key, same string wire + XML by design) | yes — one key end to end | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Full phase corpus (engine, session, hotkey, config, install, ctlsvc) incl. TestEmitters_ModeProperty, TestActor_SuperSpaceChord* (5), TestActor_FlipEmitsPanelSymbol, TestActor_FocusInReassertsModeSymbol, TestParseBinding_SuperSpace, TestDefaults, TestLoad_ModeSwitchChord, TestInstall_Sequence, TestInstall_SecondInstallKeepsOriginalBackup, TestUninstall_FullRollback, TestUninstall_CorruptStateFallsBack | `go test -race -count=1 ./engine/ ./internal/session/ ./internal/hotkey/ ./internal/config/ ./internal/install/ ./internal/ctlsvc/` | all 6 packages ok | ✓ PASS |
| Phase completion gate (build + vet + golangci-lint strict + `go test -race -count=1 ./...`) | `mise run ci` | build ok, vet ok, lint "0 issues", all 16 packages ok | ✓ PASS |
| Test existence (enumeration, no suite filtering) | `go test -list` over session/hotkey/config | all 23 chord/flip/reload session tests + 3 hotkey/config tests present | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| `mise run ci` (the plan's per-task and HEAD gate) | `mise run ci` (verifier's own process) | exit 0; lint 0 issues; full -race suite green | PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| SPEC-4.1 | 260927-way-PLAN | SPEC §4.1 hotkeys table | ✓ SATISFIED | docs/SPEC.md §4.1: new row "Переключить раскладку (сочетание) | Super+Space (goswitch перехватывает биндинг GNOME switch-input-source при установке)" with the owner-decision comment; stale parenthetical removed |
| OWNER-DEC-1 | 260927-way-PLAN | Panel mode indicator without breaking ADR-001 | ✓ SATISFIED | Truth 1 |
| OWNER-DEC-2 | 260927-way-PLAN | Super+Space goswitch-owned chord | ✓ SATISFIED | Truths 2-3 |
| OWNER-DEC-3 | 260927-way-PLAN | Install handover of switch-input-source + README docs | ✓ SATISFIED | Truths 4-5 |
| CONF-01 / CONF-02 (REQUIREMENTS.md) | context | YAML hotkeys + hot reload | ✓ SATISFIED | mode_switch_chord in the schema with hot-reload fold (already-Complete requirements reinforced, no regression) |

No orphaned requirements: REQUIREMENTS.md maps no additional IDs to this quick task, and the plan's four declared IDs are all satisfied.

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|-----------|-----------|--------|---------|----------|-----------------|---------|
| engine/emitter_wire_test.go | OWNER-DEC-1 | 3 subtests + helpers | 0 | no | Behavioral (exact signal shape/count/order via interceptor) | VALID |
| engine/wire_test.go | OWNER-DEC-1 | TestWire_FieldOrder ext | 0 | no | Value (reflect field-order pins vs installed headers) | VALID |
| internal/session/actor_test.go | OWNER-DEC-1/2 | 23 chord/flip/reload tests | 0 | no | Behavioral (consume verdicts, mode records, counters, cross-stream emit order) | VALID |
| internal/hotkey/names_test.go + internal/config/config_test.go | OWNER-DEC-2 / CONF-01/02 | TestParseBinding_SuperSpace, TestDefaults, TestLoad_ModeSwitchChord | 0 | no | Value (keyval/mask constants vs ibuskeysyms.h; decode/validate matrix) | VALID |
| internal/install/install_test.go | OWNER-DEC-3 | 4 extended/new tests | 0 | no | Behavioral (full gsettings call sequences, state-file bytes, report lines) | VALID |

**Disabled tests on requirements:** 0. **Circular patterns:** 0 (expected values derive from installed IBus headers and live-verified desktop constants, never from the system under test). **Insufficient assertions:** 0.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | Debt-marker scan (TBD/FIXME/XXX/TODO/HACK/placeholder) across all 8 production files touched: clean; skip-scan across all 7 test files: clean (matches were the `SkipReasons` status field, not test skips) | — | — |

Note: `engine.Engine.PropertyActivate/PropertyShow/PropertyHide` (engine.go:324-344) have empty bodies — this is the plan-pinned protocol no-op ("PropertyActivate stays a stub, nothing may invite clicking" — the property is a plain label, not a toggle), not an incomplete implementation. Info only.

### Human Verification Required

The four live-desktop checks in the frontmatter `human_verification:` list are the plan's own DEFERRED LIVE VERIFICATION, owned by the orchestrator per the house liveness convention: (1) ibus-daemon acceptance of the 19-field EngineDesc after restart, (2) panel rendering of RegisterProperties/UpdateProperty and the visible EN/RU indicator flip, (3) live Super+Space chord behavior with no engine-disable event, (4) the live gsettings round-trip on install/uninstall. All are external-integration behaviors (real ibus-daemon, GNOME shell, keyboard, dconf) that no headless check can observe.

### Gaps Summary

None. All five must-haves are implemented, wired, documented, and behaviorally proven at the code level; `mise run ci` is green in this verifier's own process. The task's remaining uncertainty is entirely the four deferred live-desktop checks, recorded as human/orchestrator verification items — per the plan's own verification block, not as failures.

---

_Verified: 2026-09-27T22:15:46Z_
_Verifier: Claude (gsd-verifier)_
