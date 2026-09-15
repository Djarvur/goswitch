---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
plan: 04
subsystem: input-correction
tags: [ibus, hotkey-combo, fsm, config-hot-reload, tdd, e2e, layout-flip]

# Dependency graph
requires:
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 02
    provides: config.Watcher.Snapshot() value-copy contract, ParseBinding closed name tables, -config flag, last-good watcher (D-32)
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 03
    provides: selection branch, session.Options + SetOptions, startDaemonArgs/restartDaemonWithArgs, zenity/GTE live drivers
provides:
  - internal/hotkey: NewFSM(window, tapKeyval) + FSM.SetWindow — the configurable series key and the reload window seam (D-35 mechanics, corpus untouched semantically)
  - internal/hotkey: FamilyMask(keyval) — the press-side wire truth (a press carries only pre-held modifiers; the key's own family bit rides the release)
  - internal/session: Actor.AttachConfig(interface{ Snapshot() config.Config }) — one snapshot read per event, value-folded (Pattern 2); window/options/combo-binding consumption with Pitfall-8 no-re-arm
  - internal/session: the D-36 combo branch — press of the bound key under its held modifiers above every mode branch, deliberate series Reset (Pitfall 4), comboPending settled by EVERY pipeline outcome with the flip strictly after the settled record; INFO "combo" kind=word-layout grep shape
  - engine/keys.go: KeyControlR = 0xffe4 (ibuskeysyms.h:191, provenance pinned)
  - cmd/goswitchd: actor.AttachConfig(watcher) wiring on -config
  - live e2e cases combo-word-layout (with the live tap_window_ms 500→200 reload section), layout-single (mode-record oracle, zero gsettings), super-space-alive (name-owner + post-chord correction)
  - mise tasks e2e-combo, e2e-layout-single, e2e-super-space
affects: [03-05-macr, 03-06-ctl, 03-07-matrix-v2]

actuals:
  tokens: 21800   # chars/4 over the realized diff (git diff c57abb5..HEAD = 87197 chars)
  tasks: 2
  commits: 4      # measured: git rev-list --count c57abb5..HEAD
  plan_head_before: c57abb5e873cb71c7ad976fdd31e76c1dc192eb5

tech-stack:
  added: []
  patterns:
    - "press-side binding match: Binding.ModMask with the key's own family bit cleared (hotkey.FamilyMask) — the IBus state word of a PRESS carries only the modifiers held before the key, the key's own bit appears on its release (live trace 2026-09-15)"
    - "one snapshot per event, value-folded: applySnapshot at the head of every acting entry point; the config source interface defined at the point of use (interface{ Snapshot() config.Config })"
    - "combo settlement as a terminal-state fan-out: settleCombo called at EVERY pipeline outcome (done, D-24 no-change, each refusal, verify-mismatch, verify-timeout) — the D-36 order holds on all of them; FocusOut clears without flip (the gesture died with its context)"

key-files:
  created:
    - test/e2e/case_combo.go
    - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/red-evidence/03-04-task1-red.json
    - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/red-evidence/03-04-task2-red.json
  modified:
    - internal/hotkey/fsm.go
    - internal/hotkey/fsm_test.go
    - internal/hotkey/names.go
    - internal/hotkey/names_test.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - internal/config/watch.go
    - engine/keys.go
    - cmd/goswitchd/main.go
    - test/e2e/main.go
    - mise.toml

key-decisions:
  - "Press-side combo predicate: the plan's literal mods&ModMask==ModMask is unsatisfiable on the wire (a Control_R press under Shift arrives with mods Shift|NumLock — 0x11; the Control bit rides only the release — 0x15). The match compares Binding.ModMask with the key's own family bit cleared via hotkey.FamilyMask; 03-02's summary had explicitly deferred the matching predicate to this plan"
  - "SetOptions → snapshot succession: SetOptions remains the surface of the no-config path and the unit corpus; once a source is attached, applySnapshot overwrites the options and the combo binding on every event — the attached source has priority (documented in actor.go and main.go)"
  - "The combo press TRANSITS (consume=false): a bare modifier chord puts no rune in the field, and the transit keeps the client's press/release pairing intact (HandleKey never consumes releases)"
  - "A combo retired by FocusOut/Reset is cleared WITHOUT the flip: the word half never settled because the gesture's input context died — the flip would be a surprise in the new context (the plan names done/skipped/timeout as flip outcomes; context death is the documented fourth)"
  - "Super+Space is NOT injectable on ydotool 0.1.8 (live finding: every spelling — space/SPACE/KEY_SPACE/spc/numeric — falls back to its first physical letter); super-space-alive pins the same SWCH-03 interpretation through the bare Super key (the shell mechanism the switcher rides on) + name-owner + post-chord correction. The alt+Shift_L alternate binding was tried and REMOVED: it can genuinely switch the session's input source — a destructive side effect on the owner's desktop"

requirements-completed: [SWCH-01, SWCH-02, SWCH-03, CONF-02]

coverage:
  - id: D1
    description: "The D-36 combo gesture: press of the bound key under its held modifiers (default Control_R under Shift) corrects the last word through the Double pipeline and THEN flips the mode — order pinned by the sink sequence and the log-record order; zero series actions (Pitfall-4 Reset); renavigable binding (alt+ctrl_l via ParseBinding); combo keys never feed the buffer"
    requirement: SWCH-02
    verification:
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_ComboWordThenFlip
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_ComboEmptyBufferStillFlips
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_ComboConfigurableBinding
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_ComboDoesNotFeedBuffer
        status: pass
      - kind: e2e
        ref: mise run e2e-combo (PASS — SHIFT_R+CTRL_R canonicalized by the live probe, done→mode log order, stdout привет)
        status: pass
    human_judgment: false
  - id: D2
    description: "D-35 configurability mechanics: NewFSM(window, tapKeyval) decides series of the constructor's key — corpus edits are constructor-argument-only (zero expectation changes, diff-verified against the plan base), TestFSM_ModifierUse/TestFSM_DefaultWindow green without edits, KeyControlR=0xffe4 pinned with provenance"
    requirement: SWCH-04
    verification:
      - kind: unit
        ref: internal/hotkey/fsm_test.go#TestFSM_ConfigurableTapKey
        status: pass
      - kind: command
        ref: "go test -run 'TestFSM_ModifierUse|TestFSM_DefaultWindow' -v | grep -c '^--- PASS' → 2"
        status: pass
      - kind: unit
        ref: internal/hotkey/names_test.go#TestFamilyMask
        status: pass
      - kind: command
        ref: "grep KeyControlR engine/keys.go → 0xffe4 with ibuskeysyms.h:191 provenance"
        status: pass
    human_judgment: false
  - id: D3
    description: "CONF-02 live consumption: one Snapshot() read per event; the window of a changed snapshot arms NEW series (real-timer pinned) while an armed timer survives the reload with exactly one decision (Pitfall 8); cap/rung/combo binding picked up by the next event; last-good keeps behavior unchanged; the daemon wires actor.AttachConfig(watcher) on -config and the live case pins the tap_window_ms 500→200 reload by the reload record plus a deciding post-reload tap"
    requirement: CONF-02
    verification:
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_HotReloadWindowNewSeries
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_HotReloadOptionsAndCombo
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_HotReloadInvalidKeepsLastGood
        status: pass
      - kind: e2e
        ref: mise run e2e-combo (PASS — the -config PASS part + the live reload section)
        status: pass
      - kind: command
        ref: "grep AttachConfig cmd/goswitchd/main.go → watcher handed to the actor"
        status: pass
    human_judgment: false
  - id: D4
    description: "SWCH-01 live: a single Right Shift flips the internal mode EN→RU→EN — the daemon's own mode records are the only oracle (D-34), zero gsettings reads in the case"
    requirement: SWCH-01
    verification:
      - kind: e2e
        ref: mise run e2e-layout-single (PASS)
        status: pass
      - kind: command
        ref: "grep -c gsettings test/e2e/case_combo.go → 0"
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_FlipOnSingle
        status: pass
    human_judgment: false
  - id: D5
    description: "SWCH-03 working interpretation (D-34): the shell's native Super mechanism is not broken by goswitch's presence — after the Super key fires (the overview toggle the Super+Space switcher rides on), the daemon still owns its IBus name (goswitchNameOwner) and a full word correction still runs"
    requirement: SWCH-03
    verification:
      - kind: e2e
        ref: mise run e2e-super-space (PASS ×2)
        status: pass
    human_judgment: true
    rationale: "The Super+Space chord itself is not injectable through ydotool 0.1.8 on this desktop (live finding — every name spelling falls back to its first physical letter); the case pins the mechanism through the injectable Super key plus name-owner and correction oracles. The owner judges the adequacy of this interpretation at the verify gate (the plan's FLAGGED assumption names exactly this resolution risk)."

duration: 63 min
completed: 2026-09-15
status: complete
---

# Phase 3 Plan 4: Комбо, конфиг-снапшоты и переключение Summary

**Комбо Shift+RightCtrl исправляет слово и ЗАТЕМ переключает режим (порядок D-36 пинен последовательностью sink и порядком лог-записей на обеих развязках конвейера), FSM получил конфигурируемую ключевую клавишу без единой правки ожиданий корпуса (D-35), актор потребляет конфиг-снапшоты живьём — одно чтение на событие, окно новых серий, Pitfall-8 без перевзвода — и живые кейсы закрыли SWCH-01 (mode-оракул D-34) и SWCH-03 (имя владельца + коррекция после системного Super).**

## Performance

- **Duration:** 63 min (10:32–11:36 UTC)
- **Tasks:** 2 (both tdd: RED→GREEN, RED_EVIDENCE_OK ×2)
- **Files modified:** 14 (11 code/config + 2 red-evidence records + this summary's plan artifacts)
- **Commits:** 4 (measured from plan_head_before c57abb5)

## Accomplishments

- Combo branch (SWCH-02/D-36): a press of the bound key under its HELD modifiers — recognized in feedKey ABOVE every mode branch — kills the tap series with a deliberate Reset (Pitfall 4: the pre-combo FSM silently swallowed the gesture as modifier use), launches the word pipeline of the Double semantics with `comboPending`, and `settleCombo` fires the flip at EVERY pipeline terminal (done, D-24 no-change, every refusal, verify-mismatch, verify-timeout) strictly AFTER the settled record; FocusOut/Reset clears without the flip (context death). New grep-stable INFO shape `"combo" kind=word-layout`; the `"action"`/`"mode"` shapes untouched (matrix.go greps them).
- Press-side wire truth (Rule-1 fix of the plan-literal predicate): a key PRESS's IBus state word carries only the modifiers held BEFORE the key — the key's own family bit appears on its RELEASE (live trace: Control_R press mods 0x11, release 0x15). `hotkey.FamilyMask` clears the own bit from `Binding.ModMask` for the press-side match; the 03-02 producer contract is untouched (that summary had deferred the predicate to 03-04).
- D-35 mechanics: `NewFSM(window, tapKeyval)` + `FSM.SetWindow`; the fsm_test.go diff against the plan base contains ONLY constructor-argument changes plus the purely additive new test — zero expectation edits; KeyvalShiftR stays the documented default; corpus 12/12 green.
- CONF-02 consumption: `Actor.AttachConfig(interface{ Snapshot() config.Config })`; `applySnapshot` at the head of HandleKey/ExpiryAt/HandleSurroundingText/VerifyExpiry reads the snapshot ONCE and value-folds it — the window reaches `armTimer` and `FSM.SetWindow` for NEW series (an armed timer keeps its own deadline — the reload never re-arms or cancels, pinned by the exactly-one-decision subtest), the options overwrite the SetOptions feed (documented succession), and the combo binding re-resolves only when the document's name changed (last-good on a parse failure). goswitchd hands the watcher to the actor on -config; the watcher logs `config reloaded` with the applied window (the reload-application oracle).
- Live cases: **combo-word-layout** (probe canonicalizes the injection to SHIFT_R+CTRL_R — the X-table spelling; the round pins combo→done→mode as NEW records with the timestamp order done<mode, readback+stdout привет; then the -config daemon at 500 ms, the round again, the 500→200 rewrite on the RUNNING daemon gated by the reload record, and a post-reload single tap deciding) — PASS ×4 total; **layout-single** (mode records to-ru then to-en, zero gsettings in the case file) — PASS ×3; **super-space-alive** (native Super mechanism fires, goswitchNameOwner still held, full correction after) — PASS ×4.
- Regressions: `mise run ci` green at every task boundary (0 lint issues under the strict v2 config); e2e-word, e2e-phrase PASS; the full matrix v1 16/16 PASS on the final tree.

## Task Commits

Each task committed atomically (TDD: RED before GREEN):

1. **Task 1: Комбо-распознавание до FSM + конфигурируемый tap_key + порядок слово→флип** — `a208d7a` (test, RED) + `019baf5` (feat, GREEN)
2. **Task 2: Живое потребление конфиг-снапшотов + layout-single / super-space-alive** — `6e3e96f` (test, RED) + `20b4f74` (feat, GREEN)

**Plan metadata:** this commit (docs: complete plan)

## TDD Gate Compliance

Both tdd tasks followed RED → GREEN with machine-validated red evidence:

| Task | RED commit | Evidence record | Verdict | GREEN commit |
|------|-----------|-----------------|---------|--------------|
| 1 | `a208d7a` test(03-04) | red-evidence/03-04-task1-red.json | RED_EVIDENCE_OK (4 new top-level tests failing on assertions, 8 with subtests, 102 green; the FSM stub stores tapKeyval without using it, so TestFSM_ConfigurableTapKey failed on assertions — no compile-RED needed) | `019baf5` feat(03-04) |
| 2 | `6e3e96f` test(03-04) | red-evidence/03-04-task2-red.json | RED_EVIDENCE_OK (3 new top-level tests failing on assertions, 7 with subtests: the stub AttachConfig stores the source without reading it) | `20b4f74` feat(03-04) |

No REFACTOR commit — the GREEN implementations landed lint-clean under the strict v2 config (all findings fixed inside the GREEN iterations per D-08).

## Decisions Made

- The five key-decisions of the frontmatter, in brief: press-side held-subset predicate (the live wire truth); SetOptions→snapshot succession with the source priority; the combo press transits (press/release pairing stays client-intact); context death clears a pending combo without the flip; the super-space case's injectable form.
- `tap_key` configurability ships as the constructor seam exactly as the plan scoped it ("только механика NewFSM(window, tapKeyval)"): the actor and daemon keep the documented shift_r default; the config schema validates tap_key (03-02), and wiring a non-default tap key through to the actor is a future surface the seam already accepts.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Plan-literal vs wire truth] Combo press predicate unsatisfiable as written**
- **Found during:** Task 1 GREEN (first live combo run)
- **Issue:** The plan's predicate `ev.Mods&combo.ModMask == combo.ModMask` with ModMask=Shift|Control (the 03-02 Binding) can never match a PRESS: the press's state word carries only the modifiers held before the key (live trace: Control_R press mods 0x11 = Shift|NumLock; the Control bit appears on the release 0x15).
- **Fix:** `hotkey.FamilyMask(keyval)` returns the key's own family bit; the actor matches `ModMask &^ FamilyMask(Keyval)` — the HELD subset — latch-tolerant through &. Test literals updated to the wire truth (press carries Shift+latch, no own bit).
- **Files modified:** internal/hotkey/names.go, internal/hotkey/names_test.go (TestFamilyMask), internal/session/actor.go, internal/session/actor_test.go
- **Verification:** combo corpus green under -race; e2e-combo PASS live
- **Committed in:** `019baf5`

**2. [Rule 3 - Blocking] Reload-application log record missing**
- **Found during:** Task 2 (case authoring)
- **Issue:** The plan's live-reload oracle ("waitForLog ... применение ... пин факт применения reload-записью лога") requires a log record on a VALID reload; the 03-02 watcher logged only the REJECTED path.
- **Fix:** `watch.go` reload() logs `slog.Info("config reloaded", "tap_window_ms", ...)` on the valid path — content-free (the window only), mirroring the startup `config loaded` shape.
- **Files modified:** internal/config/watch.go
- **Verification:** config corpus green; the live case gates on the record
- **Committed in:** `20b4f74`

**3. [Rule 1 - Plan-literal uninjectable] Super+Space chord cannot be injected on this stack**
- **Found during:** Task 2 (the case's live probe — exactly the plan's A5/Q-Марker resolution risk)
- **Issue:** ydotool 0.1.8 resolves EVERY unknown key name by first-letter fallback: "space", "SPACE", "KEY_SPACE", "spc" landed on the physical S/K keys, a numeric keycode on "5" (all live-traced). The D-01 experiment log's own verdict for `ydotool key super+space` was "not-switched" — no chord delivery was ever proven.
- **Fix:** super-space-alive pins the SAME D-34 working interpretation through injectable forms: the bare Super key (the shell's Super mechanism — the overview toggle, closed with esc) → goswitchNameOwner still held → full word correction after. The alt+Shift_L alternate binding was tried live and REMOVED: it can genuinely switch the session's input source — a destructive side effect on the owner's desktop the stand must not cause.
- **Files modified:** test/e2e/case_combo.go
- **Verification:** e2e-super-space PASS ×4 (reproduced)
- **Committed in:** `20b4f74`

**4. [Rule 3 - Test authoring] Three RED-authored expectation bugs fixed inside the GREEN iterations**
- **Found during:** Tasks 1–2 GREEN
- **Issue:** (a) the new FSM subtest expired the series mid-test (a second machine fixes the sequencing); (b) the combo-rebind expectation counted 2 requires where the first correction's verify-after round spends a third; (c) the e2e round captured the key-visibility baseline AFTER typing and waited for records that could never arrive.
- **Fix:** all three corrected in place; the product behavior was right in each case.
- **Files modified:** internal/hotkey/fsm_test.go, internal/session/actor_test.go, test/e2e/case_combo.go
- **Committed in:** `019baf5`, `20b4f74`

**5. [Rule 1 - Live-desktop race hardening] Post-overview focus/rebind race in super-space-alive**
- **Found during:** Task 2 (two failing runs with distinct signatures)
- **Issue:** typing right after the overview episode + engine re-activation races the IM renegotiation — early keystrokes bypass the engine (one run lost letters, one flipped the field mid-word).
- **Fix:** `waitZenityFocused` settle gate after the re-activation, plus latin-word settle gates before the taps in both the super-space and combo rounds — the race now fails early and diagnosably instead of corrupting the correction oracle.
- **Files modified:** test/e2e/case_combo.go
- **Verification:** e2e-super-space PASS ×4 after the gate
- **Committed in:** `20b4f74`

---

**Total deviations:** 5 auto-fixed (2 plan-literal vs wire truth, 1 blocking observability, 1 plan-literal uninjectable, 1 test-authoring bundle)
**Impact on plan:** All resolved toward the locked must-haves; no scope creep. The two live findings (press-side mods, ydotool name fallback) are exactly the class the plan's FLAGGED assumptions scheduled probes for.

## Issues Encountered

Environmental, all recovered without code impact: (a) one zero-delivery ydotool invocation (combo run 2 — single occurrence, four subsequent PASSes); (b) the AT-SPI bridge hung after the failed super-space probe runs (the known 02-06 live trap — Super+letter chords opened shell overlays that stole focus; fixed by dismissing the overlay and restarting the a11y registry); (c) one GTE-preflight focus flake before the bridge restart.

## Known Limitations (tracked for 03-07)

- The Super+Space chord is not e2e-injectable on ydotool 0.1.8 (see Deviations 3) — the matrix v2 super-space row, if any, inherits the Super-mechanism form of the case.
- A config reload that LENGTHENS the window mid-series can leave the armed (shorter-window) expiry a stale no-op — the FSM invariant absorbs it by design (the pinned reload tests cover the shortening direction, matching the plan's oracle).
- The mid-word mode-flake signature seen once before the settle gates remains attributed to the focus/rebind race family; the gates convert it to an early failure rather than masking it.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The combo branch, AttachConfig seam and the `"combo"` log shape are ready for the 03-07 matrix v2 step schema (a combo step greps the same record family as action/mode).
- `hotkey.FamilyMask` is the canonical press-side matching rule for any future binding consumer (MACR's Super+letter interception in 03-05 faces the same wire truth: the letter press carries Mod4, its own family, plus latches).
- The watcher's `config reloaded` record is a third grep-stable shape for reload-driving matrix cases.

## Self-Check: PASSED

All key-files exist on disk; all four task commits found in history; plan ledger measured 4 commits from plan_head_before c57abb5 (matches `commits:` in frontmatter); `mise run ci` green (exit 0, 0 lint issues) at the final tree; e2e-combo / e2e-layout-single / e2e-super-space PASS on the final tree; matrix v1 16/16.
