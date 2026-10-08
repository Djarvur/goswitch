---
phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
plan: 08
subsystem: session
tags: [sound, canberra-gtk-play, paplay, subprocess, exec, riff, pcm, fire-and-forget, warn-once]

# Dependency graph
requires:
  - phase: 07-02
    provides: config Sound schema (Enabled *bool nil=ON, EffectiveEnabled/EffectiveAutocorrectEvent, bell/message defaults)
  - phase: 07-03/07-05
    provides: SetSoundEnabled persist, menu «Звук» toggle, SetMenuSync seam form, actor fold/status plumbing
provides:
  - internal/sound Player: canberra-gtk-play preferred, paplay fallback over runtime-synthesized bundled tones (UserCacheDir cache 0700/0600), warn-once episodes, no-pipes Start+reaped Wait
  - actor SoundSink seam (Flip/AutoCorrect) + SetSoundSink; flip tone from the single flipTo path observer-last; fired tone from the autoConfirm fired point; Options.SoundEnabled gate folded from snap.Sound.EffectiveEnabled()
  - daemon wiring: SetSoundSink(sound.New(...)) next to SetMenuSync; startup effective switch feed
affects: [07-06 (e2e file-oracle), README/CONFIG docs 07-07, owner UAT (live audibility)]

# Actuals (#2632)
actuals:
  tokens: 11867   # chars/4 over the realized diff (47 468 chars)
  tasks: 3
  commits: 6

# Tech tracking
tech-stack:
  added: [] # zero new dependencies — stdlib os/exec, encoding/binary, math only
  patterns:
    - "no-pipes fork-shaped subprocess: nil Stdin/Stdout/Stderr + cmd.Start() + go cmd.Wait() reap, never Run (the wl-copy clipboard.go:141-150 precedent)"
    - "backend decision cached at the first tone (never a LookPath per flip)"
    - "warn-once per reason per episode; a healthy launch reopens the budget (the appidWarned form)"
    - "point-of-use consumer interface (SoundSink) + nil-tolerant installer (the AppidSource/SetMenuSync form)"
    - "RED-stub TDD: schema-shape stubs in the test commit, target tests failing on assertions (the 06-03 precedent), RED_EVIDENCE_OK via check tdd-red-evidence"

key-files:
  created:
    - internal/sound/sound.go
    - internal/sound/sound_test.go
  modified:
    - internal/session/actor.go
    - internal/session/actor_test.go
    - cmd/goswitchd/main.go
    - cmd/goswitchd/main_test.go

key-decisions:
  - "Bundled fallback tones are RUNTIME-SYNTHESIZED deterministic sines (880/1320 Hz, 60 ms, 16-bit PCM mono RIFF via encoding/binary+math), cached in UserCacheDir/goswitch — no binary assets in git, no go:generate, zero dependencies (the plan's flagged design discretion)"
  - "syncMode (external drift correction) is deliberately SILENT — the single sound point is flipTo per the plan's prohibition (no per-call-site wiring); a no-op same-target flip never sounds"
  - "The backend decision (canberra/paplay/none) is cached at the first tone and never re-decided (the plan's permissive «допустимо пересмотреть при новом эпизоде») — installing a player binary at runtime takes a daemon restart"
  - "The autocorrect EVENT name is wired from the STARTUP document (the pinned SoundSink interface is exactly Flip/AutoCorrect); only the enabled switch is hot-reloadable through the fold"

patterns-established:
  - "SoundSink discipline: synchronous call under the actor mutex, fire-and-forget implementation (Start, not Wait), never panics, never blocks the hot path"
  - "Tone cache validity = size match (deterministic synthesis; a size-matching file is reused, never rewritten)"

requirements-completed: ["SPEC §11 (spec-delta)"]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Every script flip (Shift taps, menu EN/RU via SwitchMode, combo settle, flip_after_correction — all through the single flipTo) sounds the flip tone exactly once, after the same-target guard, observer-last (D-36 intact)"
    requirement: "SPEC §11 (spec-delta)"
    verification:
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_FlipSoundsSink"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_SoundSinkAfterModeRecord"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_SameTargetFlipSilent"
        status: pass
    human_judgment: false
  - id: D2
    description: "A fired autocorrect sounds the distinct AutoCorrect tone from the fired point; abstentions (app-blocked and every other slug) and the off state never reach the sink; manual paths carry no tone"
    requirement: "SPEC §11 (spec-delta)"
    verification:
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_AutoCorrectFireSoundsSink"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_AutoCorrectAbstainSilent"
        status: pass
    human_judgment: false
  - id: D3
    description: "Playback is a no-pipes fork-shaped subprocess: [canberra-gtk-play -i <event>] preferred, [paplay <cache>/goswitch/<tone>.wav] fallback over deterministic bundled tones (RIFF/PCM, 0600, distinct pitches), Start+reaped Wait, never Run"
    requirement: "SPEC §11 (spec-delta)"
    verification:
      - kind: unit
        ref: "tests/internal/sound/sound_test.go#TestPlayer_FlipCanberraArgv"
        status: pass
      - kind: unit
        ref: "tests/internal/sound/sound_test.go#TestPlayer_CanberraPreferred"
        status: pass
      - kind: unit
        ref: "tests/internal/sound/sound_test.go#TestPlayer_PaplayFallback"
        status: pass
      - kind: unit
        ref: "tests/internal/sound/sound_test.go#TestPlayer_NoPipesForm"
        status: pass
      - kind: unit
        ref: "tests/internal/sound/sound_test.go#TestPlayer_NonBlocking"
        status: pass
      - kind: unit
        ref: "tests/internal/sound/sound_test.go#TestPlayer_WavDeterministic"
        status: pass
    human_judgment: false
  - id: D4
    description: "Best-effort failure discipline: missing binaries, failed Start, failed tone write — each ONE WARN per episode and a quiet no-op; attempts continue (no permanent disable); the child's own non-zero exit stays silent; no blocking, no error to the gesture"
    requirement: "SPEC §11 (spec-delta)"
    verification:
      - kind: unit
        ref: "tests/internal/sound/sound_test.go#TestPlayer_BothMissingSilent"
        status: pass
      - kind: unit
        ref: "tests/internal/sound/sound_test.go#TestPlayer_StartFailureWarnOnce"
        status: pass
    human_judgment: false
  - id: D5
    description: "Options.SoundEnabled gates both tones ahead of the sink (a muted daemon never consults it); the switch folds from snap.Sound.EffectiveEnabled() per applySnapshot (default ON, hot reload without restart); the daemon wires SetSoundSink(next to SetMenuSync) and feeds the startup effective switch"
    requirement: "SPEC §11 (spec-delta)"
    verification:
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_SoundDisabledNoSink"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_SoundFoldGatesFromSnapshot"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_NilSoundSinkNoOp"
        status: pass
      - kind: unit
        ref: "tests/cmd/goswitchd/main_test.go#TestNewActor_SoundGatingFromStartupConfig"
        status: pass
      - kind: unit
        ref: "tests/cmd/goswitchd/main_test.go#TestNewActor_SoundHotReloadGating"
        status: pass
    human_judgment: false
  - id: D6
    description: "Live audibility of both tones on the owner's desktop (canberra path; bell/message in the Yaru theme)"
    requirement: "SPEC §11 (spec-delta)"
    verification: []
    human_judgment: true
    rationale: "Sound is not readable from the e2e stand (the plan sanctions: live playback is deliberately NOT asserted; the 07-06 e2e carries the toggle's file-persistence oracle only) — the owner hears the tones at UAT"

# Metrics
duration: 36min
completed: 2026-10-05
status: complete
---

# Phase 7 Plan 8: Sound Engine Summary

**Flip and autocorrect tones from the single flipTo/fired points through a best-effort no-pipes subprocess sink — canberra preferred, runtime-synthesized paplay fallback, one-WARN episodes, hot-reloadable sound.enabled gating.**

## Performance

- **Duration:** 36 min
- **Started:** 2026-10-04T23:58:21Z
- **Completed:** 2026-10-05T00:32:22Z
- **Tasks:** 3 (tracer + tdd + auto, each RED-first)
- **Files modified:** 6

## Accomplishments
- internal/sound Player: canberra-gtk-play preferred ([canberra-gtk-play -i <event>], the event ONE argv element, no shell), paplay fallback over bundled deterministic tones (pure-stdlib RIFF/PCM synthesis cached in UserCacheDir/goswitch, 0700/0600), backend decision cached at the first tone, full no-pipes discipline (nil descriptors, Start + reaped Wait, no Run anywhere)
- Actor SoundSink seam: the flip tone fires from the SINGLE flipTo path (after the same-target guard, observer-last — D-36 record order intact), the distinct autocorrect tone from the fired point; abstentions, off state, no-op flips and manual paths never reach the sink; nil sink = silent degradation
- sound.enabled gating: Options.SoundEnabled folded from snap.Sound.EffectiveEnabled() in applySnapshot (default ON, hot reload mutes/re-arms without restart, the gate sits ahead of the sink); daemon wiring installs the Player next to SetMenuSync and feeds the startup effective switch
- Failure discipline: one WARN per episode per reason (player-missing / start-failed / tone-write-failed), attempts continue, the child's non-zero exit is silent — sound never blocks, waits or errors a flip/correction

## Task Commits

Each task was committed atomically (strict TDD — RED before GREEN):

1. **Task 1 (tracer): флип → звук** — `6698763` (test RED), `4191a12` (feat tracer)
2. **Task 2: paplay фолбэк + эпизоды WARN** — `b09e911` (test RED), `5512c7e` (feat player)
3. **Task 3: fire-хук + gating + wiring** — `4c79bb4` (test RED), `01785c9` (feat wiring)

## Files Created/Modified
- `internal/sound/sound.go` — Player: backend cache, no-pipes fork, bundled tone synthesis, warn-once episodes
- `internal/sound/sound_test.go` — fake-runner corpus: argv pins, fallback, RIFF/0600, episodes, non-blocking, determinism
- `internal/session/actor.go` — SoundSink seam + SetSoundSink + soundSinkFor gate; flipTo/autoConfirm pushes; Options.SoundEnabled + applySnapshot fold
- `internal/session/actor_test.go` — sound-sink corpus: flip/same-target/order, fire/abstain, disabled/nil, snapshot fold
- `cmd/goswitchd/main.go` — startup switch feed in newActor; SetSoundSink(sound.New(...)) next to SetMenuSync
- `cmd/goswitchd/main_test.go` — wiring corpus: startup default ON / explicit mute, hot-reload gating contour

## Decisions Made
- Bundled fallback tones are runtime-synthesized deterministic sines (880/1320 Hz, 60 ms) — the plan's flagged design discretion; no binary assets in git, no go:generate, the auditability constraint intact
- syncMode (external drift correction) is deliberately silent: flipTo is the single sound point (the plan's prohibition); a no-op same-target flip never sounds
- The backend decision is cached at the first tone and never re-decided — installing a player at runtime takes a daemon restart (the plan's permissive reading; no state beyond episodes)
- The autocorrect event name is wired from the STARTUP document — the pinned SoundSink interface is exactly {Flip, AutoCorrect}; only the enabled switch is hot

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered
- Two transient CI flakes of the known pair (TestRun_OnConnHookCalledOnce / TestWatch_BrokenBlocklistPatternKeepsLastGood class): the first `mise run ci` invocation after Task 3 exited non-zero, the two immediate re-runs were green with zero FAIL lines — sanctioned one-rerun-each treatment, both subsequent full runs clean
- TestPlayer_NonBlocking's RED run exposed a test-ordering bug (the reap count was polled while the fake child was still blocked) — fixed inside the RED commit before any production change

## User Setup Required

None — no external service configuration required. (Live audibility is the owner's UAT item: flip the language and listen; the tones ride the desktop's own sound theme via canberra.)

## Next Phase Readiness
- The sound surface the remaining phase plans consume is complete: 07-06's e2e file-oracle for the «Звук» toggle, 07-07's README/CONFIG documentation (sound section), and the owner UAT (live tones + menu toggle)
- mise run ci green on the final tree; `go vet`/lint zero issues; internal/config untouched (the sound schema belongs to 07-02/07-03)

---
*Phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii*
*Completed: 2026-10-05*

## Self-Check: PASSED

All 6 created/modified files exist on disk; all 6 task commits verified in git log
(6698763, 4191a12, b09e911, 5512c7e, 4c79bb4, 01785c9); the TDD gate sequence
test→feat is present for all three tasks; plan-level verification re-run on the
final tree: the three target packages green under -race, mise run ci green
(two consecutive clean runs), all grep gates pass, internal/config diff empty.
