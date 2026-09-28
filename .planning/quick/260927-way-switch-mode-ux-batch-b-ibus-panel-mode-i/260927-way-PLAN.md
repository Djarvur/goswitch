---
phase: 260927-way-switch-mode-ux-batch-b
plan: 01
type: execute
wave: 1
depends_on: ["260927-vu8"]
files_modified:
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
  - docs/CONFIG.md
  - docs/SPEC.md
  - README.md
  - README.ru.md
autonomous: true
requirements: [SPEC-4.1, OWNER-DEC-1, OWNER-DEC-2, OWNER-DEC-3]
estimate:
  tokens: 85000
  raw_tokens: 53000
  tasks: 3
  confidence: low
must_haves:
  truths:
    - The GNOME input indicator reflects the daemon's current script mode (EN/RU) while the goswitch engine is active — driven by an engine panel property (RegisterProperties at engine creation, UpdateProperty on every flip), with the engine descs carrying icon_prop_key so the panel icon goes dynamic — and ADR-001 stands untouched: no engine switch, XKB layout stays 'us', keyvals stay Latin (owner decision 1)
    - Pressing Super+Space flips the script mode immediately, consumes the press (no space lands in the field), kills any pending tap series, and emits the panel symbol update — the MACR layer is unaffected (no interception, no false consumed-upstream accounting) (owner decision 2)
    - The mode-switch chord is configurable as hotkeys.mode_switch_chord (closed hotkey name grammar, default super+space, empty = disabled); hot reload applies document changes live (CONF-02)
    - goswitchctl install clears GNOME's org.gnome.desktop.input-sources switch-input-source and records the previous value verbatim in install-state.json; uninstall restores it (first-install backup stays sacred; an unreadable/malformed restore falls back to the distro default, REPORTED) (owner decision 3)
    - README (both languages) documents that Alt+Shift/Super+Space no longer switch via GNOME after install, and lists the goswitch chords: single Shift tap, double Shift, Shift+right Ctrl, Super+Space
  artifacts:
    - engine/types.go (IBusProperty/IBusPropList wire structs, NewModeProperty, EngineDesc.IconPropKey)
    - engine/engine.go (UpdateModeSymbol emit) + engine/factory.go (RegisterProperties at CreateEngine)
    - internal/session/actor.go (Emitter grows UpdateModeSymbol; flipScript emits; FocusIn re-asserts; chord branch in feedKey; Options.ModeSwitchChord + applySnapshot fold)
    - internal/hotkey/names.go ("space" key name, KeyvalSpace 0x020) + internal/config/config.go (Hotkeys.ModeSwitchChord, default super+space)
    - internal/install/install.go (installState.switch_input_source, clearSwitchBinding, restore with shape check + fallback)
    - cmd/goswitchd/main.go (chord through SetOptions) + docs/CONFIG.md, docs/SPEC.md §4.1, README.md, README.ru.md
  key_links:
    - actor.flipScript ↔ engine Emitter.UpdateModeSymbol ↔ ifaceEngine UpdateProperty signal ↔ panel property keyed modePropKey ↔ EngineDesc.IconPropKey (one key string end to end)
    - feedKey chord branch ↔ hotkey.ParseBinding("super+space") press-side rule (held Mod4, FamilyMask(space)=0) ↔ macrIntercept ordering (MACR first, chord second, witness flag set)
    - config.Hotkeys.ModeSwitchChord ↔ applySnapshot fold (empty document value = disabled, not last-good) ↔ cmd/goswitchd SetOptions default
    - install saveState ↔ installState.switch_input_source ↔ restoreSwitchBinding shape check ↔ gsettings set/clear of org.gnome.desktop.input-sources switch-input-source
---

<objective>
Switch-mode UX batch B, three owner-locked behaviors (2026-09-27):

1. INDICATOR: show the current script mode (EN/RU) through an IBus engine panel property so the GNOME input indicator reflects it — RegisterProperties at engine creation, UpdateProperty emit on every flip, EngineDesc.icon_prop_key set so the panel icon goes dynamic. Do NOT switch engines; ADR-001 (internal flip, XKB 'us', Latin keyvals) stands (owner decision 1).
2. EXTRA SWITCHER: Super+Space becomes a configurable mode-switch chord owned by goswitch (YAML `hotkeys.mode_switch_chord`, default `super+space`), recognized in the actor's key path after the MACR interception and the word-layout combo, with the MACR hold explicitly witnessed so the layer is unaffected (owner decision 2; live trace 2026-09-27 20:31 proves the chord reaches the daemon once GNOME's binding is cleared).
3. INSTALL HANDOVER: `goswitchctl install` clears GNOME's `org.gnome.desktop.input-sources switch-input-source` (currently `['<Alt>Shift_L', '<Super>space']`), records the previous value in install-state.json and restores it on uninstall; README (both languages) documents the handover and the goswitch chords (owner decision 3).

Purpose: with a single input source GNOME's own switch binding only churns/disables the engine context (live finding: the 20:31 disable event followed a Super+Space press; keys then silently bypassed goswitch), and the user has no visible mode feedback.
Output: panel mode indicator end to end, configurable chord, install handover with restore, documented schema/gestures, green `mise run ci`.

STRICT TDD (directive 1): red → green → refactor, tests run at each step. Every task ends with `mise run ci` green (directive 2). Conventional commits on `gsd/quick-correction-ux`.
</objective>

<execution_context>
@/home/nil/DiskD/W/Djarvur/goswitch/.zcode/gsd-core/workflows/execute-plan.md
@/home/nil/DiskD/W/Djarvur/goswitch/.zcode/gsd-core/templates/summary.md
</execution_context>

<context>
@/home/nil/DiskD/W/Djarvur/goswitch/.planning/STATE.md
@/home/nil/DiskD/W/Djarvur/goswitch/engine/engine.go
@/home/nil/DiskD/W/Djarvur/goswitch/engine/types.go
@/home/nil/DiskD/W/Djarvur/goswitch/engine/factory.go
@/home/nil/DiskD/W/Djarvur/goswitch/engine/emitter_wire_test.go
@/home/nil/DiskD/W/Djarvur/goswitch/internal/session/actor.go
@/home/nil/DiskD/W/Djarvur/goswitch/internal/hotkey/names.go
@/home/nil/DiskD/W/Djarvur/goswitch/internal/config/config.go
@/home/nil/DiskD/W/Djarvur/goswitch/internal/install/install.go
@/home/nil/DiskD/W/Djarvur/goswitch/docs/CONFIG.md

Ground truth verified on this tree (2026-09-27):
- Property stubs: engine/engine.go:323-343 (PropertyActivate/PropertyShow/PropertyHide); the emit methods live at engine.go:380-431 with the conn-nil quiet guard. Signal interface: `ifaceEngine` (engine.go:19). Wire-test seam: `emitRecorder` over `dbus.WithOutgoingInterceptor` (emitter_wire_test.go — signals observable without a bus); field-order pins live in wire_test.go (TestWire_FieldOrder).
- Wire formats, locally pinned: `/usr/include/ibus-1.0/ibusproperty.h:159-167` (ctor order key, type, label, icon, tooltip, sensitive, visible, state, prop_list), PropType/PropState enums at :79-83/:109-111, `ibus_property_get_symbol` at :215; `/usr/include/ibus-1.0/ibusenginedesc.h:327-335` (`ibus_engine_desc_get_icon_prop_key` — the appended LAST EngineDesc field); `/usr/include/ibus-1.0/ibuskeysyms.h:376` (`IBUS_KEY_space 0x020`). The BambooEngine/goibus `property.go` (the STACK-sanctioned read-only wire reference, fetched 2026-09-27) serializes Property as Name, Attachments, Key, Type, Label, Icon, Tooltip, Sensitive, Visible, State, SubProps, Symbol — Symbol appended LAST, the appended-field convention. The owner's own working component XMLs (`/usr/share/ibus/component/punto-switcher.xml`, `test-shift.xml`) pin the XML tag spelling `icon_prop_key` (underscore).
- The dynamic-panel mechanism: EngineDesc.icon_prop_key names the engine property whose SYMBOL becomes the panel icon ("The key of IBusProperty for the dynamic panel icon" — libibus 1.5.29 string, verified via `strings /usr/lib/x86_64-linux-gnu/libibus-1.0.so.5`).
- Actor: feedKey (actor.go:1225) order is macrIntercept → word-layout combo (the D-36 press-side held-mask match at :1230-1241 with FamilyMask) → mode branches. macrIntercept (:978) owns Super+LETTERS only; space under Mod4 transits it, but leaves the Super hold unwitnessed → the Super release would WARN and count ConsumedUpstream (the macrKeyRelease logic at :1017). flipScript (:1190) is the single flip point (single tap, settleCombo, and — after 260927-vu8 — settleCorrectionFlip) and its INFO `mode` record is the e2e sequencing contract. HandleKey (:450) routes releases to fsm+macrKeyRelease and presses to feedKey. Options (actor.go:151) and the applySnapshot name-cache fold (actor.go:659, comboName precedent at :690) are the config seams.
- Config: Config/Hotkeys (config.go:62-97), Defaults() (:104), Hotkeys.validate via hotkey.ParseBinding (:157), defTapKey/defCombo (:129-130). The closed key table (names.go:56) has NO "space" yet; FamilyMask default 0 covers non-modifier keys.
- Install: installState{Sources} (install.go:153), saveState first-sacred idempotent rule (:395), takeoverSources/ownerSourcesSet (:564, :53), restoreSources shape-check + reported fallback (:616, savedSources :753), Install/Uninstall flows (:273/:316), fakeRunner call-log test idiom (install_test.go:60, TestInstall_Sequence :207, TestUninstall_CorruptStateFallsBack :511, TestInstall_ComponentXMLMirrorsWireIdentity :274), xmlEngine (install.go:172).
- goswitchd wiring: newActor SetOptions literal (cmd/goswitchd/main.go:128) — an attached watcher has per-event priority (03-04 succession).
- Sibling composition (260927-vu8 lands FIRST on this branch): correction.flip_after_correction, Options.FlipAfterCorrection, settleCorrectionFlip at the two changed-success sites. Its tests pin "the flip itself made zero sink calls" (TestActor_FlipOnSingle, TestActor_FlipAfterWordCorrection) — Task 1 changes that truth deliberately; reconcile, do not preserve.
</context>

<precondition>
Working tree is on branch `gsd/quick-correction-ux` with the 260927-vu8 sibling already landed (`grep -q FlipAfterCorrection internal/config/config.go` succeeds) and `mise run ci` green before Task 1 (assert with `git branch --show-current`, the grep, and `mise run ci`; halt if unmet — fixing vu8 is out of this plan's scope).
</precondition>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Panel mode indicator — property wire structs, engine emits, actor emits on flip and focus (owner decision 1)</name>
  <files>engine/types.go, engine/engine.go, engine/factory.go, engine/wire_test.go, engine/emitter_wire_test.go, engine/engine_test.go, internal/session/actor.go, internal/session/actor_test.go, internal/install/install.go, internal/install/install_test.go</files>
  <behavior>
    RED first, in this order (run `go test -race ./engine/` and watch the new cases fail):
    - engine/wire_test.go: extend TestWire_FieldOrder with the two new wire structs (compute signatures from the Go structs, the file's existing idiom): Property fields in EXACTLY this order — Name, Attachments, Key, Type, Label, Icon, Tooltip, Sensitive, Visible, State, SubProps, Symbol (the installed ctor order ibusproperty.h:159-167 plus the appended Symbol, per the goibus reference and the appended-field convention); PropList — Name, Attachments, PropertyList.
    - engine/emitter_wire_test.go: TestEmitters_ModeProperty with the emitRecorder seam (follow TestEmitters_DeleteAndRequire): an engine bound to the intercepted conn; UpdateModeSymbol("ru") emits exactly ONE signal — member UpdateProperty on ifaceEngine — whose body is one variant wrapping the Property struct with Key == modePropKey, Type == PropTypeNormal, State == PropStateUnchecked, Sensitive/Visible true, SubProps an empty PropList, and Symbol's inner IBusText.Text == "ru"; a second call UpdateModeSymbol("en") yields the same shape with Text "en". New factory pin: construct factory{conn: rec.conn} in-package, CreateEngine("goswitch-en") emits exactly one RegisterProperties signal (ifaceEngine) carrying a PropList of exactly one Property with the initial symbol "en" (the daemon's start mode, ADR-001).
    - engine/engine_test.go: extend TestEmitters_DetachedQuiet — a detached engine's UpdateModeSymbol is a silent no-op (the conn-nil guard).
    - internal/session/actor_test.go: fakeSink gains UpdateModeSymbol(symbol string) recording an op with the argument (the guarded-sink style). Corpus, following the wiredActor/tapShift/ExpiryAt idioms: single Shift tap → ExpiryAt flips → fakeSink recorded exactly one UpdateModeSymbol("ru") AFTER the mode log record (strings.Index, the D-36 order idiom); flip back → "en". FocusIn re-assert: after a flip to RU, HandleLifecycle(LifecycleFocusIn) re-emits UpdateModeSymbol("ru") (a newly minted engine context must not resurrect a stale EN registration). RECONCILE deliberately, decision 1 cited: the vu8-era zero-sink-call pins on flip (TestActor_FlipOnSingle; TestActor_FlipAfterWordCorrection's zero-call assertion) become exactly-one-UpdateModeSymbol pins — CommitText/DeleteSurroundingText/ForwardKeyEvent counts stay zero on a bare flip.
    GREEN:
    - engine/types.go: constants PropTypeNormal = 0 and PropStateUnchecked = 0 (ibusproperty.h:79/:109, verified verbatim comment); const modePropKey = "InputMode" and const initialModeSymbol = "en"; wire structs Property and PropList in the pinned field order (Label/Tooltip/SubProps/Symbol are dbus.Variant wrapping IBusText/PropList values, the IBusText.AttrList precedent); NewModeProperty(symbol string) building the full payload — Key modePropKey, Type PropTypeNormal (a plain label, NOT a toggle: PropertyActivate stays a stub, nothing may invite clicking), Label AND Symbol both variant-wrapped NewIBusText(symbol), Icon "", Sensitive/Visible true, State PropStateUnchecked, SubProps an empty PropList; extend EngineDesc with IconPropKey string as the LAST field (19th, after Textdomain — ibusenginedesc.h:335 appended-field order) and set it to modePropKey inside NewEngineDesc (no signature change — both prod callers and the wire/XML identity keep flowing).
    - engine/engine.go: UpdateModeSymbol(symbol string) on Engine — the conn-nil quiet guard, then Emit ifaceEngine+".UpdateProperty" with dbus.MakeVariant(NewModeProperty(symbol)), error-logged like every emitter.
    - engine/factory.go: at the end of CreateEngine, after the export loop (and with the export-failure early return untouched), emit ifaceEngine+".RegisterProperties" with dbus.MakeVariant(PropList{...Properties: []dbus.Variant{dbus.MakeVariant(NewModeProperty(initialModeSymbol))}}) on the minted path — the owner decision's "RegisterProperty at engine creation"; emission before the CreateEngine REPLY is the proven engine pattern (every Python/Go engine registers properties inside CreateEngine).
    - engine Emitter interface gains UpdateModeSymbol(symbol string) (the fifth outgoing primitive; fakeSink already updated in RED).
    - internal/session/actor.go: private modeSymbol() string ("ru"/"en" by mode); flipScript calls a.eng.UpdateModeSymbol(a.modeSymbol()) after the INFO record when a.eng != nil (the caller holds the mutex — fire-and-forget, the CommitText precedent); HandleLifecycle's FocusIn branch re-asserts the same emit (self-heals every re-minted input context: mode is actor state, engine objects are per-context).
    - internal/install/install.go + install_test.go: xmlEngine gains IconPropKey string `xml:"icon_prop_key"` (the tag spelling pinned by the owner's live component XMLs) rendered as modePropKey, keeping the component XML the wire identity's on-disk shape; extend TestInstall_ComponentXMLMirrorsWireIdentity accordingly.
    REFACTOR — whole engine + session corpora green; no existing assertion weakened without the decision cited in a comment.
  </behavior>
  <action>
    Implement exactly the pinned shapes — no generic property API surface beyond the two structs and the two emit points, no second flip path. If TestWire_FieldOrder or the mirror test surfaces an ordering mismatch with what the interceptor captures, the header + goibus order above wins; the LIVE rendering verdict belongs to the orchestrator's deferred check, not to this task.
  </action>
  <verify>
    <automated>cd /home/nil/DiskD/W/Djarvur/goswitch && go test -race -count=1 ./engine/ ./internal/session/ ./internal/install/ && mise run ci</automated>
  </verify>
  <done>
    Wire pins hold: Property/PropList field order pinned, UpdateProperty carries Key=InputMode with Symbol "ru"/"en", RegisterProperties fires once per engine creation with "en", detached engines stay quiet, every flip and every FocusIn emits exactly one symbol update in the corpus, the vu8 zero-call pins reconciled to one-call, EngineDesc carries IconPropKey end to end (wire + component XML), engine/session/install corpora and mise run ci green. Commit: feat(engine): panel mode property — EN/RU indicator through the IBus panel.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: super+space mode-switch chord — hotkey table, config key, actor branch, MACR witness, docs (owner decision 2)</name>
  <files>internal/hotkey/names.go, internal/hotkey/names_test.go, internal/config/config.go, internal/config/config_test.go, internal/session/actor.go, internal/session/actor_test.go, cmd/goswitchd/main.go, docs/CONFIG.md, docs/SPEC.md</files>
  <behavior>
    RED first, in this order:
    - internal/hotkey/names_test.go: ParseBinding("super+space") resolves to Binding{Keyval: 0x020, ModMask: MaskMod4} — space carries NO family bit, so the press-side held mask is exactly Mod4; the key-table closure still rejects non-table keys.
    - internal/config/config_test.go: TestDefaults pins Hotkeys.ModeSwitchChord == "super+space"; decode coverage (the clipboard_rung idioms): a document with `mode_switch_chord: ctrl+space` decodes and validates; an EXPLICIT empty string decodes "" and VALIDATES ok (chord disabled — the missing-key compatibility rule of the vu8 batch and the macr.alt_modifier empty precedent, so every pre-existing document loads unchanged); a garbage value refuses the whole document (D-33).
    - internal/session/actor_test.go (the live trace 2026-09-27 20:31 is the wire truth: Super+Space arrives as the space keyval with mods 0x50 — Mod4|NumLock):
      - TestActor_SuperSpaceChordFlipsMode: SetOptions{ModeSwitchChord: ParseBinding("super+space")}; HandleKey press space mods 0x50 → consume TRUE, exactly one mode record "to":"ru" IMMEDIATELY (no ExpiryAt — the chord is not a tap), one UpdateModeSymbol("ru") (Task 1 through flipScript), zero CommitText; the buffer stays clean — typing "abc" afterwards yields token "abc" with no leading space.
      - FSM reset pin: tap Shift once (series armed), chord press inside the window → after window expiry exactly ONE mode record total (the chord's; the pending tap decision died — the Pitfall 4 Reset).
      - MACR witness pins (decision 2: space is not a letter — ordering must not disturb MACR): with MACR enabled (letters "b"), press super_l bare, then the chord press, then super_l release → SuperIntercepted stays 0 (space is not a MACR letter), ConsumedUpstream stays 0, and no super-combo skip WARN is logged (the chord marked the hold as witnessed); the existing MACR corpus (super+b intercepted, burst on release) passes UNMODIFIED.
      - Gate-off: zero-value Options → the chord press transits (consume false, no mode record, no sink call); Mod2-only space (mods 0x10, no Mod4) never matches.
      - Combo precedence: ModeSwitchChord equal to the word-layout combo binding → the combo branch wins (chord branch sits after it), no double flip.
      - Hot reload: the TestActor_HotReloadOptionsAndCombo double serving ModeSwitchChord — document "super+space" → chord works; document "" → chord DISABLED (transit), not last-good; document back → chord restored (applySnapshot fold, CONF-02).
    GREEN:
    - internal/hotkey/names.go: const KeyvalSpace = 0x020 (ibuskeysyms.h:376, verified verbatim comment — never from memory); keyNameValues gains "space": KeyvalSpace; keyFamilyMasks deliberately gains NO entry (space is not a modifier — FamilyMask's default 0 is the correct press-side behavior; note it in the comment).
    - internal/config/config.go: Hotkeys.ModeSwitchChord string `yaml:"mode_switch_chord"`; defModeSwitchChord = "super+space"; Defaults() sets it; Hotkeys.validate — after the existing tap/combo checks, an empty chord returns nil (disabled), a non-empty chord must ParseBinding (the existing errTapKeyBare-style wrapping with the field name).
    - internal/session/actor.go: Options.ModeSwitchChord hotkey.Binding (zero = disabled, the WordLayoutCombo zero-value precedent); actor field chordName string cache; applySnapshot folds with the EMPTY-DISABLES twist: name change → if "" set the zero Binding, else ParseBinding-success sets it, failure keeps last-good; feedKey branch AFTER the word-layout combo block and BEFORE the mode branches: chord non-zero && keyval match && ev.Mods&held == held (held = ModMask &^ hotkey.FamilyMask(keyval), the D-36 press-side rule) → set a.macrSawLetter = true (the hold WAS witnessed by the IME — goswitch itself consumed the chord, nothing went upstream), a.fsm.Feed(hotkey.Reset{}, a.elapsed()), slog.Info("combo", "kind", "mode-switch-chord"), flipScript(), return true. Update the feedKey doc comment (the branch order is its contract) and the Options doc block.
    - cmd/goswitchd/main.go: newActor's SetOptions literal gains ModeSwitchChord from hotkey.ParseBinding(cfg.Hotkeys.ModeSwitchChord) with a warn-and-skip on error (validated documents never error; Defaults always parses).
    - docs/CONFIG.md: schema row `hotkeys.mode_switch_chord | string | super+space | closed name table (below); empty = disabled | flips the script mode directly — goswitch owns Super+Space (install clears the GNOME switch-input-source binding)`; the key-name list gains `space` (0x020); BOTH example documents gain the key; the Caramba correspondence table gains a row. State the tradeoff: documents written before this key load with the chord DISABLED (missing = off, not defaulted — the completeness discipline).
    - docs/SPEC.md §4.1: gestures table gains the row `| Переключить раскладку (сочетание) | Super+Space (goswitch перехватывает биндинг GNOME switch-input-source при установке) |` — deliberate edit of the pinned table, owner decision 2026-09-27 cited in a comment; the stale "(проксируется в GNOME Super+Space)" parenthetical on the Right-shift row is updated to match. Match the file's Russian style.
    REFACTOR — whole hotkey/config/session corpora green unmodified except the deliberate pins above.
  </behavior>
  <action>
    Implement exactly what the behavior block pins, in the layer order given. The chord branch is the entire actor-side logic — no timers, no new state beyond the Options field and the name cache. Do not touch macrIntercept or macrKeyRelease: the witness flag set in the chord branch is the whole MACR interaction.
  </action>
  <verify>
    <automated>cd /home/nil/DiskD/W/Djarvur/goswitch && go test -race -count=1 ./internal/hotkey/ ./internal/config/ ./internal/session/ && mise run ci</automated>
  </verify>
  <done>
    super+space parses (Keyval 0x020, held Mod4), defaults carry it, empty disables, garbage refuses; the chord flips immediately, consumes the press, keeps the buffer clean, kills the pending tap series, leaves MACR counters at zero with the hold witnessed, transits when off, loses to the combo on collision, hot-reloads live; CONFIG.md and SPEC §4.1 document it; corpora and mise run ci green. Commit: feat(session): configurable super+space mode-switch chord (mode_switch_chord).
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: install hands switch-input-source over, uninstall restores it; README handover docs (owner decision 3)</name>
  <files>internal/install/install.go, internal/install/install_test.go, README.md, README.ru.md</files>
  <behavior>
    RED first (the fakeRunner call-log idiom of install_test.go):
    - TestInstall_Sequence extension: install reads `gsettings get org.gnome.desktop.input-sources switch-input-source` during saveState and later sets it to `[]` (a new step after takeoverSources); the report gains a cleared line naming the key; the state JSON written carries the read value verbatim in the new field alongside sources (the current live value is `['<Alt>Shift_L', '<Super>space']`).
    - TestInstall_SecondInstallKeepsOriginalBackup extension: the existing state file is left untouched (the first-install-sacred rule covers the new field — an over-install must never save the post-handover cleared value).
    - TestUninstall_FullRollback extension: uninstall sets the saved binding back via gsettings BEFORE the state file is removed; the report names the restored value.
    - TestUninstall_CorruptStateFallsBack extension: a state file whose new field is absent (a pre-batch JSON) or malformed restores the distro default `['<Alt>Shift_L', '<Super>space']` with the substitution REPORTED (the restoreSources discipline, ASVS V5/T-04-01-02: a forged state file must never brick the keyboard bindings silently).
    GREEN:
    - internal/install/install.go: const gsettingsKeySwitch = "switch-input-source", fallbackSwitchBindings = "['<Alt>Shift_L', '<Super>space']" (live-verified 2026-09-27, comment says so), clearedSwitchBindings = "[]"; installState gains SwitchInputSource string `json:"switch_input_source"`; saveState reads both keys and stores both (same idempotent rule); a clearSwitchBinding(ctx) step in Install right after takeoverSources (set the cleared value unconditionally — idempotent; report line "switch-input-source: cleared (previous value saved)"); restoreSwitchBinding(ctx) beside restoreSources in Uninstall — read the saved field, shape-check (trimmed, starts '[' ends ']' — the savedSources discipline), untrusted → the fallback constant + a reported fallback line; failure degrades the report only where restoreSources degrades.
    - README.md and README.ru.md: the gestures tables gain the Super+Space row (switch layout EN ↔ RU); a short handover note in BOTH languages: goswitchctl install clears GNOME's switch-input-source binding — Alt+Shift and Super+Space no longer switch the GNOME input source (with the single goswitch source they only disabled the engine context — the live finding), goswitchctl uninstall restores the previous binding; the goswitch chords are single Shift tap, double Shift, Shift+right Ctrl, Super+Space. Keep each file's existing tone and structure.
    REFACTOR — whole install corpus green.
  </behavior>
  <action>
    Implement exactly the state extension, the two gsettings steps and the reported fallback — no selfcheck change (the six-step D-41 audit is a pinned Phase-4 contract; extending it is not part of this decision), no new CLI surface (goswitchctl install/uninstall already flow through). The README edits are documentation of decision 3, not marketing.
  </action>
  <verify>
    <automated>cd /home/nil/DiskD/W/Djarvur/goswitch && go test -race -count=1 ./internal/install/ && mise run ci</automated>
  </verify>
  <done>
    Install snapshots + clears the binding with a report line; uninstall restores the saved value (or the reported distro default when untrusted/absent) before the state file dies; first-install backup stays sacred; README (both languages) documents the handover and the chord list; install corpus and mise run ci green. Commit: feat(install): hand the GNOME switch-input-source binding over to goswitch (restore on uninstall).
  </done>
  <reversibility rating="reversible">The handover is uninstall-restorable by construction (state file + fallback), and the chord/indicator are config/behavior deltas with no schema migration — every piece reverts by code revert or uninstall.</reversibility>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| config document → chord binding | the YAML value crosses into key-path behavior on load and every hot reload |
| install-state.json → gsettings set | the saved binding string is untrusted input at uninstall time (a forged/corrupt file must not brick the keyboard) |
| IBus wire payloads → ibus-daemon/panel | the Property/PropList variants cross to the daemon that owns desktop input (a malformed variant must not destabilize registration) |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-WAY-01 | Tampering | install-state.json switch_input_source | medium | mitigate | shape check (trimmed, '['…']') before any gsettings write + reported fallback to the distro default — the savedSources/T-04-01-02 discipline extended to the new field |
| T-WAY-02 | DoS | mode_switch_chord parse on hot reload | low | accept | one ParseBinding over the closed name tables per document change, the existing validate+fold path; no new allocation class |
| T-WAY-03 | Denial of Service | property variants to ibus-daemon 1.5.29 | medium | mitigate | wire order pinned from installed headers + the goibus reference and pinned by interceptor tests; registration acceptance is explicitly in the orchestrator's live gate — a daemon-rejecting variant is a loud, revertible failure (registration dies visibly), never silent input loss |
| T-WAY-04 | Information Disclosure | mode/chord log records | low | accept | records carry only the mode symbol and config binding names (D-20: config names are not user data; keystroke content never enters the log) |
| T-WAY-SC | Tampering | package installs | n/a | accept | no npm/pip/cargo installs — stdlib + godbus only, package legitimacy gate not triggered |
</threat_model>

<verification>
- `mise run ci` green after every task (build + vet + golangci-lint + `go test -race ./...`).
- engine: Property/PropList field-order pins, UpdateProperty/RegisterProperties shapes via the interceptor seam, detached quiet, KeySpace constant provenance.
- session: flip emits exactly one symbol update (and after refusals/D-24, the vu8 pins still hold where they should); chord flips immediately + consumes + cleans the FSM; MACR unaffected (counters zero, corpus unmodified); gate-off, combo precedence, hot reload, empty-disables.
- config: defaults, decode round-trip, empty = valid-disabled, garbage refuses.
- install: snapshot + clear + restore + reported fallback + sacred first backup.
- Docs: CONFIG.md schema + examples + name table, SPEC §4.1 rows, README both languages.
- DEFERRED LIVE VERIFICATION — mark these for the orchestrator (the house liveness convention; the panel and ibus-daemon acceptance are not observable headlessly):
  1. Registration survives the 19-field EngineDesc: after a daemon restart on the live desktop, `ibus list-engine | grep goswitch` still shows both engines and `journalctl --user -u goswitchd` shows no variant/registration errors.
  2. Indicator: with the goswitch engine active, `dbus-monitor --session "interface='org.freedesktop.IBus.Engine'"` shows RegisterProperties at engine creation and UpdateProperty with Symbol "ru"/"en" on each flip; the GNOME top-bar indicator visibly flips EN ↔ RU.
  3. Chord live: press Super+Space in a real field — mode flips (log `mode to=ru`), no space inserted, indicator updates; GNOME's cleared binding produces no engine-disable event (the 20:31 defect shape stays dead).
  4. Handover live: `goswitchctl install` then `gsettings get org.gnome.desktop.input-sources switch-input-source` → `[]`; `goswitchctl uninstall` restores `['<Alt>Shift_L', '<Super>space']`.
</verification>

<success_criteria>
- The GNOME input indicator follows the goswitch script mode while ADR-001 stands (owner decision 1).
- Super+Space flips the mode instantly as a goswitch-owned chord, configurable via hotkeys.mode_switch_chord, MACR untouched (owner decision 2).
- Install owns the GNOME switch-input-source binding with a faithful, restore-guaranteed handover; README documents the new gesture reality in both languages (owner decision 3).
- All pre-existing tests pass unmodified or were updated deliberately with the decision cited; mise run ci green; conventional commits on gsd/quick-correction-ux.
</success_criteria>

<output>
Create `.planning/quick/260927-way-switch-mode-ux-batch-b-ibus-panel-mode-i/260927-way-SUMMARY.md` when done
</output>
