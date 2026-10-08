---
phase: "01"
plan: "03"
subsystem: e2e-stand
tags: [e2e, ydotool, at-spi, ibus, resilience, session-actor]
requires:
  phase: "01"
  provides: "engine seam (plan 01-01), hotkey FSM (plan 01-02), layout tables (plan 01-02)"
provides:
  - "internal/session.Actor — EventHandler implementation serializing the FSM (mutex + AfterFunc window), logging {\"msg\":\"action\",\"n\":N} at expiry"
  - "test/e2e CLI runner — preflight fail-fast, daemon subprocess with -debug, wait-for-condition log asserts, snapshot/restore desktop state, exit != 0 on FAIL"
  - "Three live cases: m1-gate (M1 gate), ibus-restart (INTEG-04), kill9-survive (INTEG-05)"
  - "test/e2e/focus_helper.py — AT-SPI witness / text readback / best-effort grabFocus on /usr/bin/python3"
  - "mise tasks e2e-m1, e2e-ibus-restart, e2e-kill9-survive (live session; NOT in ci)"
  - "test/e2e/README.md — run book, state policy, keyd manual checklist (INTEG-02), no-root rights (INTEG-03)"
affects:
  - cmd/goswitchd/main.go
  - mise.toml
  - .golangci.yml
actuals:
  tokens: 17654
  tasks: 3
  commits: 4
  plan_head_before: a18a92f87ddec2a204fd99068b1cacb3873ed75e
tech-stack:
  added: []
  patterns:
    - "witness-gated injection: never type before the AT-SPI witness proves the stand's surface owns keyboard focus"
    - "dual-path input surface: zenity --entry (self-focuses on map, stdout readback) primary; shell PASSWORD_TEXT entry (overview search / lock prompt) fallback for locked sessions"
    - "desktop-state restore contract: gsettings both keys only-if-changed, global engine re-asserted AFTER daemon death (ibus unsets it on component disconnect)"
key-files:
  created:
    - test/e2e/main.go
    - test/e2e/preflight.go
    - test/e2e/case_m1.go
    - test/e2e/case_resilience.go
    - test/e2e/focus_helper.py
    - test/e2e/README.md
  modified:
    - internal/session/actor.go
    - internal/session/actor_test.go
    - cmd/goswitchd/main.go
    - mise.toml
    - .golangci.yml
key-decisions:
  - "Activation via IBus SetGlobalEngine (`ibus engine`), not gsettings sources+current: the GNOME 46 shell runtime-ignores the keys write (live-verified); both keys stay snapshotted and restored only-if-changed"
  - "Input surface = self-focused zenity entry (primary) / shell PASSWORD_TEXT entry (locked fallback): AT-SPI grabFocus is refused under Wayland focus-stealing prevention (GTK4 errors, GTK3 returns false)"
  - "Teardown restores the global engine AFTER daemon death and never to a goswitch engine: ibus-daemon unsets the global engine when an engine component's connection closes (live-verified)"
  - "kill9 delivery oracle is length-based: the desktop's active XKB group (RU at test time) maps ghbdtn→привет, so content comparison would be layout-dependent"
  - "Engine-registration preflight via ListActiveEngines on the private bus: `ibus list-engine` reads only the XML registry and cannot see programmatic registrations (01-01 live finding)"
patterns-established:
  - "wait-for-condition on the daemon JSON log (waitForLog/waitForNew with count baselines) — the M1 assertion surface for all later phases"
  - "AT-SPI witness as the stand's focus oracle (input-surface roles preferred over window frames; walker tolerates defunct nodes)"
  - "Resilience case shape: prove desktop-side survival with the daemon dead, then respawn and await re-registration in the same log"
requirements-completed: [TEST-01, TEST-02, TEST-03, INTEG-02, INTEG-03, INTEG-04, INTEG-05]
coverage:
  - id: session-actor
    description: "Actor serializes the FSM, logs action n=1|2|3 at window expiry, FocusOut disarms; wired into the daemon with one line"
    requirement: TEST-01
    verification:
      - kind: unit
        ref: "mise exec -- go test ./internal/session/ -race -count=1 (RED 4278895 → GREEN 74a26d4)"
        status: pass
    human_judgment: false
  - id: e2e-m1-gate
    description: "Physically injected keys visible in the goswitchd log with goswitch-en active; FSM decision on double Right Shift; lifecycle order registered → focus_in"
    requirement: [TEST-02, TEST-03, INTEG-02, INTEG-03]
    verification:
      - kind: e2e
        ref: "mise run e2e-m1 → PASS (live GNOME Wayland session, repeated after interim FAILs)"
        status: pass
    human_judgment: false
  - id: surface-activation
    description: "Stand activates its own input surface before every case and gates injection on the witness (zenity self-focus; shell entry fallback)"
    requirement: TEST-03
    verification:
      - kind: e2e
        ref: "witness-gated openEntrySurface in every case; mise run e2e-* PASS"
        status: pass
    human_judgment: false
  - id: ibus-restart-case
    description: "ibus restart → re-registered on the new socket → keys flow again"
    requirement: INTEG-04
    verification:
      - kind: e2e
        ref: "mise run e2e-ibus-restart → PASS (log shows WARN reconnect → connected to new socket → re-registered generation=1)"
        status: pass
    human_judgment: false
  - id: kill9-survive-case
    description: "kill -9 daemon → desktop typing alive through plain engine (AT-SPI/stdout readback) → runner respawn → fresh registration"
    requirement: INTEG-05
    verification:
      - kind: e2e
        ref: "mise run e2e-kill9-survive → PASS"
        status: pass
    human_judgment: false
  - id: keyd-compatibility-docs
    description: "INTEG-02: architectural proof (no evdev capture, transit by construction, proven by cases) + documented manual check for a live keyd session"
    requirement: INTEG-02
    verification:
      - kind: manual_procedural
        ref: "test/e2e/README.md § INTEG-02 (4-step owner checklist; keyd.service is failed/disabled on this machine, start requires root)"
        status: pass
    human_judgment: true
    rationale: "The live keyd-session check cannot be automated here (keyd needs root and is disabled); the owner runs the documented checklist. The architectural claim itself is proven by the stand's cases (keys transit, nothing consumed)."
duration: 80 min
completed: 2026-09-11T00:30:00+03:00
status: complete
---

# Phase 01 Plan 03: ADR-пакет и каркас IBus-движка — e2e-стенд Summary

Session actor wiring the tap FSM into the daemon, plus a live-session e2e stand (ydotool injection, zenity/shell surface activation, AT-SPI witness) that greens the M1 gate and both resilience cases — ibus restart re-registers on the new socket, kill -9 leaves desktop input alive with runner respawn.

## Performance

- Duration: ~80 min wall (23:03–00:15 in-commit work; executed across two agents — the first was killed by a harness timeout mid-task-2, the continuation finished, verified and closed out)
- Started: 2026-09-10T23:03 (first plan commit)
- Completed: 2026-09-11T00:30
- Tasks: 3/3
- Files: 13 changed (+1875/−1)

## Accomplishments

- **Task 1 (TDD, D-07):** `internal/session` actor — mutex-serialized `Feed`, `time.AfterFunc(window)` re-arming, `{"msg":"action","n":N}` logged at expiry, `FocusOut` disarms; daemon wiring is exactly one line (`session.NewActor(hotkey.DefaultWindow)` in `cmd/goswitchd/main.go`). RED against a stub → GREEN minimal implementation, corpus green under `-race`.
- **Task 2:** e2e runner (`test/e2e`, CLI not go-test): six named fail-fast preflights (injection self-test, ibus address, /dev/uinput writable, /usr/bin/python3 gi, engine registration via ListActiveEngines, log heartbeat); daemon as subprocess with `-debug`; snapshot/restore of BOTH gsettings keys plus the global engine; `waitForLog`/`waitForNew` polling asserts; m1-gate case proving keys + FSM decision + lifecycle order in the log.
- **Task 3:** resilience cases green live — `ibus-restart` (reconnect loop reaches the new socket, `re-registered` generation=1, keys flow after re-activation) and `kill9-survive` (SIGKILL, desktop typing alive via readback, runner respawn, second `component registered` awaited); stand README (run book, state policy, keyd manual checklist, no-root rights); two more mise tasks.
- **Full canonical verification green in one sequence:** `mise run ci && mise run e2e-m1 && mise run e2e-ibus-restart && mise run e2e-kill9-survive` — exit 0, desktop state fully restored after every run (engine `xkb:us::eng`, sources untouched).

## Task Commits

| Task | Type | Hash | Subject |
|------|------|------|---------|
| 1 | test (RED) | 4278895 | session-actor corpus against the stub (D-07) |
| 1 | feat (GREEN) | 74a26d4 | session actor serializes the FSM and logs decisions (D-07) |
| 2 | feat | 60cc0bb | e2e stand — preflight, shell surface activation, ydotool injection, m1-gate case |
| 3 | feat | a23ce33 | resilience cases + stand surface generalization + INTEG-02/03 docs |

No refactor commit: the GREEN implementation needed no restructuring (REFACTOR is conditional on an obvious improvement).

## Files Created/Modified

Created: `test/e2e/main.go`, `test/e2e/preflight.go`, `test/e2e/case_m1.go`, `test/e2e/case_resilience.go`, `test/e2e/focus_helper.py`, `test/e2e/README.md`.
Modified: `internal/session/actor.go`, `internal/session/actor_test.go`, `cmd/goswitchd/main.go`, `mise.toml`, `.golangci.yml`.
Removed: `test/e2e/.gitkeep` (placeholder, plan-prescribed).

## Decisions Made

- Activation is IBus `SetGlobalEngine`, not the planned gsettings write: the GNOME 46 shell runtime-ignores `sources`/`current` writes (live-verified — list updates, no focus_in, no key routing). Both keys remain snapshotted and are restored only-if-changed.
- The stand's input surface is a spawned `zenity --entry` (freshly mapped windows take focus without a grab) with the shell's own PASSWORD_TEXT entry (overview search / lock-screen prompt) as the locked-session fallback. The planned AT-SPI `Component.grabFocus` is refused under Wayland focus-stealing prevention (GTK4 errors, GTK3 returns false).
- Injection is witness-gated: text is typed only after the AT-SPI witness proves the surface owns keyboard focus — injected text never reaches the owner's windows.
- Teardown re-asserts the global engine AFTER daemon death (ibus-daemon unsets it when an engine component disconnects — live-verified twice) and never restores to a goswitch engine (xkb fallback derived from the snapshot).
- kill9's delivery oracle is length-based: the desktop's active XKB group (RU during the runs) maps `ghbdtn`→`привет`, so readback compares character count, not content.

## Deviations from Plan

**1. [Rule 1 - Bug] Activation mechanism replaced (gsettings → SetGlobalEngine)**
- Found during: Task 2 (previous executor, re-verified by continuation)
- Issue: `gsettings set org.gnome.desktop.input-sources sources/current` is runtime-ignored by the GNOME 46 shell — no focus_in, no key routing to the engine.
- Fix: activation via `ibus engine goswitch-en` (SetGlobalEngine); snapshot/restore of both keys kept per the safety prohibition.
- Files: `test/e2e/case_m1.go` (activateGoswitch), `test/e2e/README.md`
- Verification: all three live cases PASS.
- Commit: 60cc0bb

**2. [Rule 1 - Bug] Teardown engine restore reordered + gsettings writes made conditional**
- Found during: Task 2 (continuation, live runs)
- Issue: (a) value-identical `gsettings set` triggers GNOME's input-source re-evaluation and live-unsets the global engine; (b) ibus-daemon unsets the global engine when the engine component's connection closes, undoing a pre-death restore.
- Fix: gsettings restored only when the value differs; engine re-asserted after `stopDaemon`; goswitch engines never restore targets (xkb fallback).
- Files: `test/e2e/main.go`
- Verification: `ibus engine` = `xkb:us::eng` after every run (checked after each live case).
- Commit: 60cc0bb

**3. [Rule 3 - Blocker] Surface activation mechanism replaced (grabFocus → zenity self-focus / shell entry)**
- Found during: Task 2
- Issue: AT-SPI grabFocus is refused under Wayland (GTK4 errors, GTK3 returns false); the overview search entry is not AT-SPI-exposed as focused on an unlocked session; the lock-screen prompt is (PASSWORD_TEXT) — one oracle cannot serve both states.
- Fix: dual-path `openEntrySurface` — zenity entry primary (self-focuses on map, stdout readback), shell entry fallback for locked sessions (super wakes, Enter reveals the prompt); injection gated on the witness.
- Files: `test/e2e/case_m1.go`, `test/e2e/focus_helper.py`, `test/e2e/case_resilience.go`
- Verification: m1-gate PASSED on both paths (locked 23:55, unlocked 00:12+).
- Commits: 60cc0bb, a23ce33

**4. [Rule 1 - Bug] ibus-daemon name-table corruption after a SIGKILLed daemon**
- Found during: continuation start (standoff: `org.freedesktop.IBus.goswitch` listed with no owner, RequestName → InQueue)
- Issue: the previous executor's harness-kill left the component name stuck in ibus-daemon's name table; every daemon start failed the single-instance guard.
- Fix (environment, not code): `ibus restart` clears the table. The kill9 case's clean SIGKILL+respawn did NOT reproduce the corruption (second RequestName succeeded) — noted as a rare ibus 1.5.29 race for Phase 2 observation.
- Files: none (live-environment remediation; recorded in README's state policy context and here)
- Verification: registration green across all subsequent runs.
- Commit: n/a

**5. [Rule 2 - Missing critical] Preflight check 2 uses ListActiveEngines, not `ibus list-engine`**
- Found during: Task 2 (01-01 live finding carried forward)
- Issue: `ibus list-engine` reads only the XML registry — programmatic registrations are invisible to it.
- Fix: preflight waits for the `component registered` log line, then queries `ListActiveEngines` on the private bus and walks the variant tree for `goswitch-en`.
- Files: `test/e2e/preflight.go`
- Verification: preflight green on every run.
- Commit: 60cc0bb

**6. [Rule 1 - Bug] focus_helper text readback against atspi 2.52 API**
- Found during: Task 3
- Issue: `Atspi.Text.get_text(start, end)` has a broken introspected signature ("takes exactly 1 argument").
- Fix: `get_text_at_offset(0, LINE_START)` → TextRange.content, with the legacy call as fallback.
- Files: `test/e2e/focus_helper.py`
- Verification: probe readback worked; zenity stdout remains the primary oracle.
- Commit: a23ce33

Totals: 6 deviations (3× Rule 1, 1× Rule 2, 1× Rule 3, 1 environment remediation). None architectural (Rule 4); daemon/engine code untouched by tasks 2–3.

## Issues Encountered

- The returning owner physically typed on the desktop during two live runs (~23:58): cases failed safely (witness-gate refused to inject into the owner's window); re-runs after idle (verified via org.gnome.Mutter.IdleMonitor) passed. Lesson recorded in the README run book implicitly via the witness gate.
- The AT-SPI tree races window churn ("accessible does not exist") — the helper's walker treats every node access as best-effort.
- `ibus engine` (no args) exits non-zero when no global engine is set — the snapshot treats that as "no engine" rather than a stand failure.

## User Setup Required

- INTEG-02 live keyd check (owner, manual): the 4-step checklist in `test/e2e/README.md` — keyd.service is failed/disabled on this machine and needs root; until run, keyd compatibility stays "architecturally proven, not live-confirmed".
- None otherwise: the stand self-activates its surfaces and restores desktop state.

## Next Phase Readiness

- M1 gate green: key events, FSM decisions (`single/double/triple` at INFO as `{"msg":"action","n":N}`) and lifecycle order are observable in the daemon's JSON log with goswitch-en active — the assertion surface Phase 2 builds on.
- The AT-SPI readback primitive (witness char counts, text subcommand) and the layout-mapping observation (ghbdtn→привет under the RU group) are direct inputs to the Phase 2 correction matrix (TEST-04).
- The stale-name race (deviation 4) and the keycode/keyval semantics on the live IM path (key trace captured in -debug logs) are recorded for Phase 2 planning (D-01 experiment in plan 04).

## Self-Check: PASSED

- Commits 4278895, 74a26d4, 60cc0bb, a23ce33 present on gsd/phase-1-adr-paket-i-karkas-ibus-dvizhka (git log verified).
- All six created files exist under test/e2e/ (checked); actor test files present in internal/session/.
- Full canonical verification (ci + 3 live cases) exit 0 on 2026-09-11T00:2x; desktop state restored (ibus engine = xkb:us::eng, sources/current unchanged).
