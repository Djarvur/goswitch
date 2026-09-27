---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
plan: 05
subsystem: input-correction
tags: [ibus, macr, super-letter, forward-key-event, at-spi, appid, a11y, tdd, e2e, adr-005]

# Dependency graph
requires:
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 02
    provides: macr.* config schema (enabled/letters/apps/alt_modifier), Snapshot() value-copy contract, -config flag
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 04
    provides: feedKey combo branch placement, applySnapshot one-read-per-event, ForwardKeyEvent burst discipline, live case scaffolding (startDaemonArgs/restartDaemonWithArgs, zenity drivers, refocus pokes)
provides:
  - internal/session: the MACR-01 branch — Super+letter consumed above combo and mode branches, the Ctrl+letter burst emitted on the Super RELEASE, consumed-upstream detect (WARN "super combo skipped" reason=consumed-upstream) with counters MACRStats{SuperIntercepted, ConsumedUpstream} (the 03-06 goswitchctl surface)
  - internal/appid: Observer — live godbus a11y-bus focus observer (FocusedApp() (string, error), ErrBusClosed, errors never panics), lazy-started only under a non-empty macr.apps, degradation to the global rule with WARN "app identity unavailable"
  - internal/session: per-app scoping + the ADR-005 degradation ladder (start failure and runtime error both fall back to the global rule, once-per-episode WARN)
  - engine/keys.go: KeySuperL=0xffeb / KeySuperR=0xffec (ibuskeysyms.h provenance)
  - test/e2e: cases macr-probe (A7 spike), macr-super-letter (wire-relay oracle), macr-per-app (live A4 observer + matched-focus interception); mise tasks e2e-macr-probe / e2e-macr-super-letter / e2e-macr-per-app
affects: [03-06-ctl, 03-07-matrix-v2]

actuals:
  tokens: 24108   # chars/4 over the realized diff (git diff ef29e5a..HEAD = 96434 chars)
  tasks: 3
  commits: 5      # measured: git rev-list --count ef29e5a..HEAD
  plan_head_before: ef29e5af9632e507f3d0aacd39c50e5cc74b0bfb

tech-stack:
  added: []
  patterns:
    - "remap-on-release: a ForwardKeyEvent burst emitted while the physical Super is still held reaches the client as a Ctrl+Super chord (the client tracks the held modifier) and the binding never matches — the remap rides the Super release, the same clean-modifier discipline as the clipboard rung's Ctrl+V (live finding 2026-09-15)"
    - "wire-relay oracle: when the client stack does not apply forwarded events, the dbus-monitor capture of the ibus bus (engine emission + ibus-daemon relay to the focused InputContext) is the provable boundary — pin everything up to it and pin the client non-action as the documented actual"
    - "shell-bound Super letters read from gsettings, never probed: toggle-application-view=<Super>a here opens the app grid mid-probe and steals the keyboard; bound letters (a/h/n/o/s/v on this desktop) become verdict rows without injection"
    - "a11y identity = the bridge path namespace: /org/gnome/Zenity/a11y/<uuid> → org.gnome.Zenity from the StateChanged(focused) signal — no tree walk, no cache round trip; the observer only learns from focus GAINS (fresh windows)"

key-files:
  created:
    - internal/appid/appid.go
    - internal/appid/appid_test.go
    - test/e2e/case_macr.go
    - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/red-evidence/03-05-task2-red.json
    - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/red-evidence/03-05-task3-red.json
  modified:
    - internal/session/actor.go
    - internal/session/actor_test.go
    - engine/keys.go
    - test/e2e/main.go
    - mise.toml

key-decisions:
  - "Super+letter delivery CONFIRMED live (A7): the letter press arrives at ProcessKeyEvent with Mod4 (b: keyval 0x62, mods 0x50 = Mod4|NumLock) and its release too — but only for letters the shell has NOT bound; Super+a is toggle-application-view on this desktop (press reaches the IME, then the app-grid overlay steals the keyboard), so the probe reads the bound set from gsettings and never injects it, and the Task-2 live letter is a free one ('x'; 'b' proven by the spike scan)"
  - "The Ctrl+letter burst fires on the SUPER RELEASE, not on the letter press: a burst forwarded while the physical Super is held arrives client-side as Ctrl+Super and no binding matches (first live implementation failed exactly there; the plan-literal press-time burst was corrected to the wire truth)"
  - "The live interception oracle is the IBus WIRE: the engine's four-event burst and ibus-daemon's relay to the focused InputContext are captured by a dbus-monitor sidecar — because the GTK4-Wayland client stack never APPLIES forwarded key events (no cut under an active selection, no cursor move, no paste; the relay demonstrably lands). The unchanged field is pinned as the documented actual; the same limitation covers the never-live-driven ADR-003 level-2 replay and the D-28 Ctrl+V burst"
  - "The per-app identity vocabulary is the bridge path namespace (org.gnome.Zenity), reported from the StateChanged(focused) signal's path — no AT-SPI tree walk; A4 closed POSITIVE: the godbus observer works live (identity within the focus event of a fresh window)"
  - "The consumed-upstream WARN fires on a Super release with NO letter press seen during the hold (not 'no interception'): a transited unconfigured letter means the IME saw everything — nothing was consumed upstream; the detect is armed only while the layer is enabled, so a MACR-off desktop never warns on its native Super use (the overview toggle)"

requirements-completed: [MACR-01]

coverage:
  - id: D1
    description: "The A7 delivery spike ran BEFORE the interception logic and pinned the wire truth in verdict lines: the bound-letter table from gsettings (Super+a = toggle-application-view here: press delivered, focus stolen), a free letter delivered with Mod4 (super+b: press+release both, entry keeps focus), the identical trace under the internal RU mode, and the bare Super press/release pair both visible — consumed-upstream detect viable"
    requirement: MACR-01
    verification:
      - kind: e2e
        ref: mise run e2e-macr-probe (PASS ×4 — stdout verdict lines "letter-delivered: yes (letter \"b\" keyval=0x62 mods=0x50)" and "consumed-upstream-detect: true" grep-stable)
        status: pass
    human_judgment: false
  - id: D2
    description: "The MACR-01 interception: a configured letter press carrying Mod4 is consumed ABOVE the combo and mode branches (RU pin: Latin burst, no Cyrillic commit), the buffer is never fed, the Ctrl+letter burst rides the Super release (4-event prototype sequence, alt_modifier swaps the modifier key only, empty default = Control_L per b.3), and the consumed-upstream detect WARNs with its counter while SuperIntercepted grows separately; live: the interception record plus the on-bus relay of both letter halves under -config"
    requirement: MACR-01
    verification:
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_MACRIntercepts
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_MACRConsumedUpstream
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_MACRRULayoutStillIntercepts
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_MACRAltModifier
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_MACRDisabledByDefault
        status: pass
      - kind: e2e
        ref: mise run e2e-macr-super-letter (PASS ×3 — INFO super-intercept record + dbus-monitor wire relay of both letter halves + the unchanged-field actual)
        status: pass
    human_judgment: false
  - id: D3
    description: "Per-app scoping with the ADR-005 ladder: a non-empty macr.apps list scopes the interception to the focused app (in-list consumed, out-of-list transits, empty-apps = global with ZERO observer starts — the lazy start fires exactly once); an unavailable or erroring identity source degrades to the GLOBAL rule with the WARN \"app identity unavailable\" (once per episode, never a panic); live: the case-side observer reads the identity from a fresh window, the observed name becomes the macr.apps entry, and the daemon's own lazy observer feeds the interception under the matched focus"
    requirement: MACR-01
    verification:
      - kind: unit
        ref: internal/appid/appid_test.go#TestAppid_FocusedAppByFakeBus
        status: pass
      - kind: unit
        ref: internal/appid/appid_test.go#TestAppid_StartBusUnavailable
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_MACRPerAppMatch
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_MACRAppidDegradation
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_MACRGlobalWhenNoApps
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_MACRAppidLazyStart
        status: pass
      - kind: e2e
        ref: mise run e2e-macr-per-app (PASS ×3 — per-app-observer: focused identity "org.gnome.Zenity"; interception fired under the matched focus)
        status: pass
    human_judgment: false

duration: 99 min
completed: 2026-09-15
status: complete
---

# Phase 3 Plan 5: Макросы Super→Ctrl Summary

**Перехват MACR-01 работает по ADR-005: спайк доказал живую доставку буквы с Mod4 (только для не-забинженных оболочкой букв — Super+a занят toggle-application-view), бёрст Ctrl+буква выезжает на ОТпускании Super (клиент с зажатым Super читает синтетическую комбинацию как Ctrl+Super), consumed-upstream WARN со счётчиками наблюдаем, per-app списки живут на godbus-наблюдателе a11y-шины с деградацией на глобальные правила; живой оракул перехвата — dbus-monitor захват релея ibus-daemon, потому что GTK4-Wayland стек НЕ применяет форварднутые события (задокументированное фактическое, тот же класс, что level-2 ADR-003).**

## Performance

- **Duration:** 99 min (14:40–16:19 UTC)
- **Tasks:** 3 (1 auto spike + 2 tdd: RED→GREEN, RED_EVIDENCE_OK ×2)
- **Files modified:** 10 (8 code/config + 2 red-evidence records)
- **Commits:** 5 (measured from plan_head_before ef29e5a)

## Accomplishments

- **Spike (A7, Task 1):** the macr-probe case pins the ProcessKeyEvent trace per gesture — the delivery table below. Two live findings shaped everything after: Super+a is shell-bound here (the press reaches the IME, the app grid then steals the keyboard — releases lost, the follow-up flip died), and a FREE letter's full four-event sequence arrives with Mod4 on every letter event. The bound-letter set is read from gsettings (never injected — each probe of a bound chord wedges the desktop's focus chain); the consumed-upstream raw material (bare Super press AND release visible) is proven.
- **Interception (Task 2):** the feedKey branch sits above the combo and every mode branch — a configured letter press with Mod4 is consumed in ANY internal mode (the RU pin: Latin burst, no Cyrillic commit), never feeds the buffer, prints nothing, and logs the INFO `super intercept` record naming the config letter only (D-20). The remap burst (the owner prototype's four-event Ctrl+letter shape with exact evdev keycodes) is deferred to the Super RELEASE — the first implementation burst at the press and the live cut never landed: the client still tracks the held Super and reads the synthetic chord as Ctrl+Super. consumed-upstream (b.2): a Super release with no letter press seen during the hold WARNs `super combo skipped` reason=consumed-upstream and counts; the detect is armed only while enabled (a MACR-off desktop never spams on native Super use). Counters `MACRStats{SuperIntercepted, ConsumedUpstream}` are the fixed 03-06 surface; alt_modifier mechanics per b.3 (empty default stays Control_L; ctrl_r swaps the modifier key only).
- **Live interception oracle:** macr-super-letter drives a -config daemon (layer on, letters "x"), intercepts super+x and proves the remap ON THE WIRE — a dbus-monitor sidecar captures the engine's burst AND ibus-daemon's relay of both letter halves to the focused InputContext. The field's unchanged content is pinned as the actual: on this GTK4-Wayland desktop the client stack does not apply forwarded key events at all (see Known Limitations).
- **Per-app (Task 3):** internal/appid.Observer — a pure-godbus observer on the a11y bus (session bus GetAddress → socket dial EXTERNAL → StateChanged match); identities come from the bridge-namespaced event paths (org.gnome.Zenity), no tree walk. The actor scopes a non-empty macr.apps list by the focused app; an unavailable/erroring source degrades to the GLOBAL rule with the WARN `app identity unavailable` (once per episode); the observer starts lazily, exactly once, only under a non-empty list (zero a11y connections otherwise). Live (macr-per-app): the case-side observer reads the identity from a fresh window's focus gain (A4 POSITIVE), the observed name becomes the macr.apps entry, and the daemon's own lazy observer feeds the interception under the matched focus.
- **Regressions:** `mise run ci` green at every task boundary (0 lint issues under the strict v2 config); e2e-word, e2e-combo, e2e-macr-probe, e2e-macr-super-letter, e2e-macr-per-app all PASS on the final tree.

## The A7 delivery table (spike verdicts)

| Probe | Injection | Observed in ProcessKeyEvent (keyval, mods) | Verdict |
|---|---|---|---|
| (a) super+a | `ydotool key super+a` | press 0xffeb (Super) mods 0x10; press 0x61 mods 0x50 — then a focus_out/in churn (~8 ms later): the releases NEVER arrive | press delivered with Mod4, but the chord is shell-bound (toggle-application-view): the app grid steals the keyboard — NOT a usable interception letter |
| (a) super+b | `ydotool key super+b` | press 0xffeb mods 0x10; press 0x62 mods 0x50; release 0x62 mods 0x50; release 0xffeb mods 0x50 — the full four-event sequence, the entry keeps focus | letter-delivered: yes (the Task-2 live corpus letter) |
| (b) super+b in RU mode | single Right-Shift flip, then the same chord | IDENTICAL trace (press 0x62 mods 0x50, …) | the interception input is layout-independent — the branch must sit above the mode branches (pinned) |
| (c) super solo | `ydotool key super` | press 0xffeb mods 0x10; release 0xffeb mods 0x50 — both halves as separate records | consumed-upstream-detect: viable (the detect's raw material is fully observable) |

Bound Super+letters on this desktop (gsettings, never probed): a=toggle-application-view, h=minimize, n=focus-active-notification, o=rotate-video-lock-static, s=toggle-quick-settings, v=toggle-message-tray.

## Task Commits

Each task committed atomically (TDD: RED before GREEN):

1. **Task 1: Спайк доставки Super+Буква (A7, Pitfall 7)** — `9737fd1` (feat)
2. **Task 2: MACR-ветка перехвата + consumed-upstream** — `7fd0aec` (test, RED) + `8e7f7de` (feat, GREEN)
3. **Task 3: Per-app списки + internal/appid + лестница деградации** — `324b1c0` (test, RED) + `c49310f` (feat, GREEN)

**Plan metadata:** this commit (docs: complete plan)

## TDD Gate Compliance

Both tdd tasks followed RED → GREEN with machine-validated red evidence:

| Task | RED commit | Evidence record | Verdict | GREEN commit |
|------|-----------|-----------------|---------|--------------|
| 2 | `7fd0aec` test(03-05) | red-evidence/03-05-task2-red.json | RED_EVIDENCE_OK (3 new top-level tests failing on assertions — interception, consumed-upstream, RU order; DisabledByDefault/BufferNotFed green immediately by design, the 03-04 continuity-pin precedent; compile stubs only: Options MACR fields, MACRCounters zero stub, KeySuperL/KeySuperR) | `8e7f7de` feat(03-05) |
| 3 | `324b1c0` test(03-05) | red-evidence/03-05-task3-red.json | RED_EVIDENCE_OK (5 new top-level tests failing on assertions — the appid observer over a fake feed, Start's no-panic error, per-app match/degradation/lazy-start; 54 green) | `c49310f` feat(03-05) |

No REFACTOR commit — the GREEN implementations landed lint-clean under the strict v2 config (all findings fixed inside the GREEN iterations per D-08, including the test-corpus splits for funlen/goconst).

## Decisions Made

- The five key-decisions of the frontmatter, in brief: bound letters are read from gsettings and never injected; the remap rides the Super release (held-Super pollution); the wire relay is the live oracle because GTK4-Wayland does not apply forwarded events; the a11y identity is the bridge path namespace (A4 positive); the consumed-upstream WARN keys on no-letter-seen, armed only while enabled.
- The observer's vocabulary (org.gnome.Zenity-style) is self-consistent: macr.apps matches what the observer reports, and goswitchctl (03-06) will display the same names — the python-helper's tree names ("zenity") are a different vocabulary and are NOT the config's.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Plan-literal vs live truth] The probe letter 'a' is shell-bound on this desktop**
- **Found during:** Task 1 (the first probe runs)
- **Issue:** The plan's probe (a) assumes super+a is an unbound GNOME combo; live, Super+a IS toggle-application-view: the letter press reaches the IME with Mod4, then the app grid opens and steals the keyboard — the releases vanish and every follow-up injection lands in the shell surface (the RU-flip round died there).
- **Fix:** The probe reads the shell-bound Super+letter set from the three GNOME keybinding schemas (bound letters become verdict rows without injection) and scans only free letters; the Task-2/3 live cases use 'x' (free; the spike scan proved 'b' — the scan stops at the first delivered letter).
- **Files modified:** test/e2e/case_macr.go
- **Verification:** e2e-macr-probe PASS ×4; the bound-letter table in stdout
- **Committed in:** `9737fd1`

**2. [Rule 1 - Plan-literal vs wire truth] The burst cannot fire at the letter press**
- **Found during:** Task 2 (the first live interception runs)
- **Issue:** The plan's burst-on-press sequence is delivered while the physical Super is still held: the client's own modifier tracking composes the synthetic Ctrl+letter into Ctrl+Super+x and no binding matches — the live cut never happened (the widget semantics were separately proven correct with a manual ctrl+a/ctrl+x round).
- **Fix:** The remap is deferred to the Super RELEASE (macrPendingKeyval; FocusOut/Reset drop it with the dying context); the unit corpus pins no-burst-before-release plus the burst-at-release.
- **Files modified:** internal/session/actor.go, internal/session/actor_test.go
- **Verification:** MACR corpus green under -race; the live relay captured on the wire
- **Committed in:** `8e7f7de`

**3. [Rule 1 - Live-desktop limitation] GTK4-Wayland does not apply IBus-forwarded key events**
- **Found during:** Task 2 (the live oracle chase)
- **Issue:** Even with the burst correctly on the release, the client never acts: a dbus-monitor capture proves the engine's four events AND ibus-daemon's relay to the focused InputContext (mutter's IM client), and the widget never applies them — no cut under an active selection, no cursor move. The same never-live-driven family as the ADR-003 level-2 Backspace replay (02-05: "no live level-2 witness on this desktop") and the D-28 Ctrl+V burst (03-03: "the daemon-side rung has no live driver").
- **Fix:** The live case's oracle is the on-bus relay (a dbus-monitor sidecar in the stand) plus the PINNED unchanged field — the actual, per the project's pin-the-actual discipline; the daemon-side behavioral contract stays in the unit corpus. Escalated to the owner at the verify gate (Known Limitations + the WINDOWS ledger).
- **Files modified:** test/e2e/case_macr.go
- **Verification:** e2e-macr-super-letter PASS ×3 (the relay of both letter halves captured)
- **Committed in:** `8e7f7de`

**4. [Rule 1 - Semantics refinement] The consumed-upstream WARN keys on no-letter-seen, not no-interception**
- **Found during:** Task 2 (the detect's design)
- **Issue:** The plan's literal "no interception during the hold" would WARN on a super+x chord whose letter simply is not in the configured set (it transited — the IME saw everything, nothing was consumed upstream): a false positive.
- **Fix:** The WARN fires when NO letter press (intercepted or transited) arrived during the hold — the honest "consumed before the IME" signal; armed only while the layer is enabled.
- **Files modified:** internal/session/actor.go
- **Verification:** TestActor_MACRConsumedUpstream (the transited-letter subtest) green
- **Committed in:** `8e7f7de`

**5. [Rule 3 - Environmental] The already-focused poke emits no focus gain**
- **Found during:** Task 3 (the first per-app run)
- **Issue:** The grabFocus poke on an ALREADY-focused window produces no StateChanged(focused=1) — the observer's cache stayed empty through pokes (both the case's probe and the daemon's lazy observer).
- **Fix:** Every observer boundary in the case is crossed by a FRESH window (the map-focus event is the gain); the case documents the lazy observer's one-focus-cycle latency property.
- **Files modified:** test/e2e/case_macr.go
- **Verification:** e2e-macr-per-app PASS ×3
- **Committed in:** `c49310f`

---

**Total deviations:** 5 auto-fixed (2 plan-literal vs wire truth, 1 live-desktop limitation made the oracle, 1 semantics refinement, 1 environmental probe lesson)
**Impact on plan:** All resolved toward the locked must-haves. Deviations 2 and 3 are exactly the FLAGGED A7-adjacent unknowns the plan scheduled live probes for; no scope creep.

## Issues Encountered

Environmental, all recovered without product impact: (a) the AT-SPI bridge wedged twice during Task 1 (the known 02-06/03-04 trap — Super chords opening shell overlays; fixed by dismissing the overlay and restarting the a11y registry; one GTE-preflight flake was collateral); (b) one buffered-stdout diagnostic confusion and one gsettings GVariant-bracket parse bug in the probe's bound-letter reader (found by the injection itself re-opening the app grid — the exact failure the reader exists to prevent).

## Known Limitations (tracked for the verify gate / 03-06/03-07)

- **GTK4-Wayland does not apply IBus-forwarded key events on this desktop** (GNOME 46): the relay demonstrably reaches the focused InputContext; the widget never acts. Consequences: the MACR Ctrl+letter remap reaches the client boundary but does not move GTK4 apps; the ADR-003 level-2 Backspace replay and the D-28 clipboard rung's Ctrl+V share the path and were never live-driven either. Recorded in the WINDOWS ledger; the owner decides at the verify gate (document as a known GNOME limitation vs investigate XWayland/native-IBus surfaces).
- **The lazy per-app observer needs one focus cycle after the daemon start**: its start lands on the first key event (which postdates the current surface's focus gain), so the first chord after a restart may transit until focus moves once — documented in the case; a cache-priming Start (the registry's GetItems walk) is a possible future improvement.
- **The consumed-upstream WARN fires on every bare Super press/release while MACR is enabled** — including the plain overview toggle (that IS a mutter-consumed Super combination, but it is also normal desktop use). Per plan/ADR b.2 as written; the owner may want a per-combination scoping later.
- **The observer only learns from focus GAINS on bridge-marked paths** (GTK4 /a11y/ namespaces; older atk /root trees report no identity) — unmarked apps leave the cache at the last known value.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `MACRCounters() MACRStats{SuperIntercepted, ConsumedUpstream}` is the fixed goswitchctl surface for 03-06; the new grep-stable log shapes are `"msg":"super intercept"` (INFO, config letter only), `"msg":"super combo skipped","reason":"consumed-upstream"` (WARN) and `"msg":"app identity unavailable"` (WARN).
- The per-app vocabulary is the bridge path namespace (org.gnome.Zenity) — goswitchctl status should surface the observed focused app so users can copy it into macr.apps.
- For the 03-07 matrix: a macr step can gate on the `super intercept` record like action/mode; the macr-probe case is the template for any future live-delivery question (bound-letter table + verdict lines).

## Self-Check: PASSED

All key-files exist on disk; all five task commits found in history (9737fd1, 7fd0aec, 8e7f7de, 324b1c0, c49310f); the plan ledger measured 5 commits from plan_head_before ef29e5a (matches `commits:` in frontmatter); `mise run ci` green (exit 0, 0 lint issues) at the final tree; e2e-macr-probe / e2e-macr-super-letter / e2e-macr-per-app / e2e-word / e2e-combo PASS on the final tree.
