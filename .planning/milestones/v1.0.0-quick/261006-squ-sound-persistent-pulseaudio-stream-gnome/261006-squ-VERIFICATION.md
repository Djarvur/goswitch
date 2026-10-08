---
phase: 261006-squ-sound-persistent-pulseaudio-stream-gnome
verified: 2026-10-06T19:46:16Z
status: human_needed
score: 7/7 must-haves verified
covered_files:
  - ".planning/REQUIREMENTS.md"
  - ".planning/quick/261006-squ-sound-persistent-pulseaudio-stream-gnome/261006-squ-PLAN.md"
  - ".planning/quick/261006-squ-sound-persistent-pulseaudio-stream-gnome/261006-squ-SUMMARY.md"
  - "README.md"
  - "cmd/goswitchd/main.go"
  - "docs/CONFIG.md"
  - "docs/SPEC.md"
  - "go.mod"
  - "go.sum"
  - "internal/session/actor.go"
  - "internal/session/actor_test.go"
  - "internal/sound/pcm.go"
  - "internal/sound/pcm_test.go"
  - "internal/sound/pulse.go"
  - "internal/sound/pulse_test.go"
  - "internal/sound/sound.go"
  - "internal/sound/sound_test.go"
  - "internal/sound/theme.go"
  - "internal/sound/theme_test.go"
covered_digest: "v1:sha256:628f594af4394e2930dca37ad66396d1f944246dc17701ba42ee758f4ca2e8c1"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 6/7
  gaps_closed:
    - "With sound muted there is NO connection to the sound server — the same close applies when the desktop's system-wide event-sounds key turns off (plan decision PD-1): the Player now holds a muted-suppression state (atomic.Bool) wired from BOTH watcher read paths; foldTone drops arriving tones WITHOUT a redial; corpus cells pin dialCount stays-put under the mute in both the steady-state and the startup arm"
  gaps_remaining: []
  regressions: []
---

# Quick Task 261006-squ Verification Report: persistent PulseAudio stream + GNOME sound theme

**Task Goal:** Sound: persistent PulseAudio stream + GNOME sound-theme playback to eliminate tone latency. Owner's three hard requirements: (a) GNOME sound theme as the primary path; (b) theme settings read at start, tracked async — playback never waits; (c) sound disabled = no sound-server connection. Plus: spec-delta before code (D-55); persistent jfreymuth/pulse stream; PCM cache; instant-fail degradation; synthesized tone as never-wait fallback; SoundSink seam unchanged.

**Verified:** 2026-10-06T19:46:16Z
**Status:** human_needed
**Re-verification:** Yes — after gap closure (fix commits b02da6a RED + 38f892a fix on gsd/sound-persistent-pulse-stream)
**Branch:** gsd/sound-persistent-pulse-stream at 38f892a (10 commits ae9ca73..38f892a, all present)

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Theme tone through a PERSISTENT native connection; playback never blocks the gesture (drop-if-busy) | ✓ VERIFIED | `internal/sound/pulse.go` (one lazily dialed `pulse.Client`, ONE int16-LE stream, `toneReader` select/default silence so `Start` always returns); `sound.go` `enqueue` = muted early-drop + select/default on a buffered-1 channel; `TestPlayer_LazyConnectDialsOncePerConnectedPeriod`, `TestPlayer_FlipNonBlockingUnderWedgedWorker` green. End-to-end beyond fakes: `TestDecodeOggVorbis_RealThemeFile` PASSED against the live `/usr/share/sounds/Yaru/stereo/bell.oga` (44.1 kHz stereo), and a verifier-run live-server smoke (production `dialPulse` → `openStream` → silence `play` → `drain` → `close`, temp test since deleted) PASSED against the machine's running PulseAudio socket. The ~30 ms audibility figure is the documented owner by-ear gate (Human Verification item 1). |
| 2 | Playback never waits for a settings read or decode; theme read at start + supervised monitor; async rebuild + atomic swap; miss → synth immediately | ✓ VERIFIED | `sound.go` `watchSettings`/`reread`/`scanMonitor` (full two-key read at every (re)start = PD-2; 5 s supervision pause, no ticker); `applyTheme` → `go p.cache.load` → `atomic.Pointer` swap strictly off the play path; `foldTone` miss → `playSynth` at the stream's format with no load on the path. Corpus: `TestPlayer_WatcherThemeCatchUp`, `TestPCMCache_AtomicSwapMidRebuild`, `TestPCMCache_MissingPairAnswersMiss`, `TestSynthTone_*` — all green. |
| 3 | Muted = NO sound-server connection; ON→OFF fold closes; lazy redial; same close AND steady-state suppression on event-sounds off (PD-1) | ✓ VERIFIED | **Config/tray arm** (verified in the initial round, regression-checked): `foldSound` ON→OFF → `SoundStopper.Stop()` pinned once (`TestActor_SoundFoldOffStopsSink`); `TestPlayer_StopMuteLifecycle` unchanged and green. **PD-1 arm — gap closed and re-verified at all three levels:** (1) EXISTS — `Player.muted atomic.Bool` (sound.go:167) + `setEventSounds(on, observed)`; (2) WIRED from BOTH watcher read paths — `reread()` calls `p.setEventSounds(on, false)` (startup/catch-up applies silently, no INFO), `scanMonitor()` calls `p.setEventSounds(on, true)` for every event-sounds line (off: `muted.Swap(true)` + ONE INFO per observed transition + `p.Stop()` closes the live connection; on: `muted.Store(false)` — the lazy redial rides the next tone); `enqueue` early-drops under the mute, and the worker-authoritative `foldTone` re-check `p.muted.Load()` drops WITHOUT a redial (closing the enqueue/unmute race); (3) BEHAVIOR PINNED — `TestPlayer_EventSoundsMuteDropsTonesWithoutRedial` (steady state: tone under the mute → dialCount stays 1, writeCount stays 1; unmute line clears the gate; next tone dials a FRESH conn, count 2) and `TestPlayer_EventSoundsMutedAtStartDialsNothing` (startup: read `sounds="false"` → dialCount stays 0, writeCount stays 0; unmute → dial 1, write 1). Both cells green ×5 under -race. Pre-existing `TestPlayer_WatcherEventSoundsMute`/`TestPlayer_StopMuteLifecycle` unchanged (diff 795c8dc..38f892a is purely additive to the test file). |
| 4 | Server unavailable = instant clean failure; one WARN per episode, quiet after, lazy retry; never an error on the gesture path | ✓ VERIFIED | `dialPulse` returns the dial's own fast error, no retry loop; warn-once per reason with `closeEpisode` reopen. Corpus: `TestPlayer_ConnectFailedWarnOnceThenReopen`, `TestPlayer_StreamFailureOwnEpisode`, `TestPlayer_GSettingsMissingFallsBack` — green. |
| 5 | SoundSink seam byte-unchanged; existing actor sound-gating corpus green untouched | ✓ VERIFIED | `SoundSink` interface block unchanged in the diff (context only); `SoundStopper` a pure addition; pre-existing corpus passes unmodified (`internal/session` ok, 3.574 s, -race at the fix HEAD). |
| 6 | Synth survives as the never-wait fallback; NO player subprocesses; NO cache files; retired machinery gone (gsettings watcher remains per PD-2/PD-3) | ✓ VERIFIED | Grep gates clean (no canberra/paplay/aplay/pw-play/UserCacheDir/toneWAV/wavHeader); only pinned-argv `gsettings get`/`monitor` subprocesses; no file writes — in-memory atomic snapshot only. Synth corpus green. |
| 7 | Exactly two DIRECT deps (pulse v0.1.3 + oggvorbis v1.0.5, + vorbis v1.0.2 via oggvorbis) recorded in SPEC BEFORE code (D-55); tidy-diff clean | ✓ VERIFIED | go.mod graph exact; `9bebd67` docs-only commit precedes all code; SPEC §5 dated amendment + §11 dated addendum with the 2026-10-04 verdict verbatim; `go mod tidy` → no drift (re-checked at the fix HEAD). |

**Score:** 7/7 truths verified (0 present, behavior-unverified)

### Re-verification: Gap Closure Evidence (Level 1–3 on the carried gap)

| Level | Check | Result |
| ----- | ----- | ------ |
| Exists | `muted atomic.Bool` + `setEventSounds(on, observed)` in `internal/sound/sound.go` | ✓ substantive, documented (PD-1 effective-play gate) |
| Wired | `reread()` → `setEventSounds(on, false)`; `scanMonitor()` → `setEventSounds(on, true)` on every event-sounds line; `enqueue` early-drop; `foldTone` worker-authoritative drop; `setEventSounds(false)` → `p.Stop()` closes the live connection | ✓ both read paths converge on one applier; both edges handled |
| Behavior | `TestPlayer_EventSoundsMuteDropsTonesWithoutRedial` + `TestPlayer_EventSoundsMutedAtStartDialsNothing` assert dialCount/writeCount stay pinned under the mute and the fresh-dial lazy reconnect after unmute; green ×5 under -race | ✓ PASS |
| Regression | `TestPlayer_WatcherEventSoundsMute`, `TestPlayer_StopMuteLifecycle` unchanged (additive diff only) and green; full suite at fix HEAD green (sound 1.532 s, session 3.574 s) | ✓ no regressions |

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `docs/SPEC.md` §5+§11 | Dated dependency amendment + dated redesign block, audit trail preserved | ✓ VERIFIED | Unchanged since the initial round; the fix touched no docs (the §11 "НЕТ вовсе" wording now matches the code) |
| `docs/CONFIG.md` / `README.md` | Playback rewording, keys byte-identical | ✓ VERIFIED | Unchanged since the initial round |
| `go.mod` / `go.sum` | Two pinned direct modules + vorbis, no drift | ✓ VERIFIED | Exact measured graph; tidy re-run clean at the fix HEAD |
| `internal/sound/theme.go` | gsettings at start + supervised monitor + XDG resolver | ✓ VERIFIED (wired) | Unchanged since the initial round |
| `internal/sound/pcm.go` | (theme,event) cache with atomic swap + deterministic synth | ✓ VERIFIED (wired) | Unchanged since the initial round |
| `internal/sound/pulse.go` | Persistent client+stream: lazy dial, silence reader, format-follows, idempotent close | ✓ VERIFIED (wired) | Unchanged since the initial round; live-server smoke passed |
| `internal/sound/sound.go` | Player: seam methods, drop-if-busy, single worker, warn-once, Start/Stop, PD-1 muted gate | ✓ VERIFIED (wired) | 60-line fix: `muted` flag, `setEventSounds`, `enqueue` early arm, `foldTone` drop |
| `internal/session/actor.go` | SoundStopper seam + ON→OFF fold closing the sink | ✓ VERIFIED (wired) | Unchanged since the initial round |
| `cmd/goswitchd/main.go` | Sink constructed once, Start BEFORE SetSoundSink, install shape unchanged | ✓ VERIFIED (wired) | Unchanged since the initial round |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | -- | --- | ------ | ------- |
| actor applySnapshot ON→OFF | SoundStopper.Stop → pulse client+stream closed | `foldSound` type assertion → `Player.Stop` → worker `pb.close()` | ✓ WIRED | Pinned; unchanged by the fix |
| Flip/AutoCorrect | muted early-drop → non-blocking enqueue → worker `foldTone` re-check → cache lookup; miss → synth | `muted.Load()` + buffered-1 `toneCh` select/default | ✓ WIRED | The muted arm now gates BEFORE dialing; drop-if-busy unchanged |
| theme watcher | resolver → oggvorbis decode → atomic swap, off the play path | `applyTheme` → `go cache.load` → `snap.Store` | ✓ WIRED | Unchanged by the fix |
| watcher event-sounds (both arms) | `setEventSounds` → `muted` flag (+INFO once per observed transition, +Stop on off) | `reread` (silent) and `scanMonitor` (observed) | ✓ WIRED | The re-verified gap's link; both read paths and both edges pinned |
| stream format | follows decoded PCM; change recreates stream; reader serves silence | `ensureStream` drain+close+recreate; `toneReader` | ✓ WIRED | Unchanged; live smoke passed |
| config keys | `sound.enabled` / `sound.autocorrect_event` schema unchanged | no config changes in the diff | ✓ WIRED | `internal/config` absent from the branch diff |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| `themeResolver` | theme file bytes | XDG data dirs (`/usr/share/sounds/...`) | Yes — live `bell.oga` decoded 44.1 kHz stereo in-test | ✓ FLOWING |
| `pcmCache` | decoded PCM | `decodeOggVorbis` over resolved theme files | Yes | ✓ FLOWING |
| `pulseConn`/`pulseStream` | int16-LE samples | cache hit or `synthTone` | Yes — live-server smoke wrote silence through the real socket | ✓ FLOWING |
| watcher → `currentTheme` / `muted` | theme name; event-sounds state | `gsettings get`/`monitor org.gnome.desktop.sound` | Yes (pinned-argv subprocess; schema constants only) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Full module suite, -race, at fix HEAD | `go test -race -count=1 ./...` | sound ok (1.532 s), session ok (3.574 s), all others ok; 1 failure: the known pre-existing `TestRun_OnConnHookCalledOnce` flake (internal/ctlsvc — untouched by the branch, passes isolated; adjudicated in the initial round) | ✓ PASS |
| New mute cells + pre-existing mute cells, ×5 stress, -race | `go test -race -count=5 -run 'TestPlayer_EventSoundsMuteDropsTonesWithoutRedial\|TestPlayer_EventSoundsMutedAtStartDialsNothing\|TestPlayer_WatcherEventSoundsMute\|TestPlayer_StopMuteLifecycle' ./internal/sound/` | ok (3.152 s) — no flakes | ✓ PASS |
| Real decoder on live theme PCM | `go test -run TestDecodeOggVorbis_RealThemeFile ./internal/sound/` (initial round) | PASS against live Yaru PCM | ✓ PASS |
| Live-server dial/stream/play/close (production path, silence only) | temp `TestVerifyRealDialSmoke` (initial round, temp test deleted after) | PASS in 1.02 s against `/run/user/1001/pulse/native` | ✓ PASS |
| Lint | `mise run lint` (golangci-lint, fix HEAD) | 0 issues | ✓ PASS |
| Tidy gate | `go mod tidy && git diff -- go.mod go.sum` (fix HEAD) | no drift | ✓ PASS |
| By-ear latency / live theme retune / muted-socket check | owner on the GNOME session | not automatable | ? SKIP → Human Verification |

### Probe Execution

No `scripts/*/tests/probe-*.sh` probes declared or conventional for this task; the mise gates (build/vet/lint/test/tidy-diff) are the task's declared runnable checks and were run above at the fix HEAD (lint, tidy, full -race suite, targeted stress).

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ----------- | ----------- | ------ | -------- |
| QUICK-261006-squ-sound-persistent-pulse-stream | 261006-squ-PLAN.md (inline; no REQUIREMENTS.md entry — quick task) | Persistent PulseAudio stream + GNOME sound-theme playback, latency eliminated, three owner requirements, D-55 spec-delta | ✓ SATISFIED | All three owner requirements verified at code+behavior level: (a) theme primary path with real decode; (b) never-waits with async watcher; (c) no connection while muted — both the config/tray arm and the PD-1 event-sounds arm, in transition AND in steady state. Remaining: the inherently-manual by-ear acceptance. |

No orphaned requirements: `.planning/REQUIREMENTS.md` maps no additional IDs to this quick task.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | Zero debt markers in all changed non-test files (re-checked at the fix HEAD for sound.go); no stubs; logs carry reason slugs and library errors only | — | — |

**Info (pre-existing, out of scope):** `TestRun_OnConnHookCalledOnce` (internal/ctlsvc) failed once again in this round's single full -race run (same test-isolation quirk as the initial round); passes isolated; `internal/ctlsvc` has an empty diff across the entire branch including the fix commits. Not a gap of this task.

### Human Verification Required

The owner's by-ear acceptance (plan Task 4 human-check; documented manual gate — the only remaining items):

### 1. Flip → tone immediate

**Test:** On the GNOME machine, flip layouts repeatedly and listen.
**Expected:** The theme tone follows the gesture immediately (~30 ms prototype class vs ~150 ms per-tone subprocess today); the switch is never delayed or blocked.
**Why human:** Perceptual latency in the live session is not automatable.

### 2. Live theme retune

**Test:** Switch the GNOME sound theme in Settings while the daemon runs.
**Expected:** The next tone reflects the new theme without a daemon restart (supervised watcher + async rebuild).
**Why human:** Requires the live desktop session and hearing the change.

### 3. Muted = no sound-server connection

**Test:** Toggle «Звук» off in the tray/config; run `ss -x | grep pulse` and observe no goswitch connection; toggle back on and flip. Optionally repeat with the system's event-sounds key (GNOME Settings → Sound → alert/system sounds off): flipping must stay silent with no new pulse connection, and re-enabling restores the tone on the next flip.
**Expected:** No pulse socket connection while muted (config arm verified; the event-sounds steady-state arm is now corpus-proven — this listens for what the corpus cannot hear).
**Why human:** Requires the live session's socket state and hearing.

### Gaps Summary

The single gap from the initial verification — the PD-1 event-sounds arm leaking in steady state (tones arriving while the system event-sounds key is off redialed and played) — is **resolved** and verified at all three levels in commits b02da6a (RED corpus) + 38f892a (fix): a `muted atomic.Bool` on the Player, applied from BOTH watcher read paths through one `setEventSounds(on, observed)` applier (silent on re-reads, ONE INFO per observed transition), closing the live connection via `Stop()` on the off-edge, gating `enqueue` early, and dropping arriving tones in `foldTone` without a redial; the unmute edge clears the gate so the lazy redial rides the next tone. The new corpus pins dialCount/writeCount stays-put under the mute in both the steady-state and startup arms, green ×5 under -race; the pre-existing mute cells are unchanged and green; lint 0 issues, tidy clean, and the full -race suite at the fix HEAD is green modulo the adjudicated pre-existing ctlsvc flake. No regressions.

**Status: human_needed** — every automatable must-have now verifies (7/7); the only remaining items are the owner's three inherently-manual by-ear gates listed above.

---

_Verified: 2026-10-06T19:46:16Z_
_Verifier: Claude (gsd-verifier)_
