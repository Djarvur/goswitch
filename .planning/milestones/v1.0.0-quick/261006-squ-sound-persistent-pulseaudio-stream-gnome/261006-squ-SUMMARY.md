---
status: complete
phase: 261006-squ-sound-persistent-pulseaudio-stream-gnome
plan: 01
subsystem: sound, session, cmd/goswitchd
tags: [sound-latency, pulseaudio, ogg-vorbis, xdg-sound-theme, gsettings-watcher, mute-lifecycle, tdd, spec-delta]
requires:
  - owner-locked design in .planning/todos/pending/sound-latency.md (prototype measurements, «делай» 2026-10-05)
  - D-55 spec-delta discipline (docs before code)
provides:
  - persistent native PulseAudio connection (jfreymuth/pulse v0.1.3) — the tone rides a ~133 µs write on an already-open stream (prototype class ~30 ms audible vs ~150 ms per-tone subprocess)
  - GNOME sound-theme playback per the XDG Sound Theme spec (theme from org.gnome.desktop.sound, supervised gsettings monitor with PD-2 catch-up re-reads, atomic PCM cache swap)
  - synthesized sine fallback at the stream's own format on every cache miss — the play path never waits for a settings read or a decode
  - mute lifecycle: ON→OFF fold closes the sound-server connection exactly once (SoundStopper), event-sounds=false folds to the same close (PD-1), lazy redial on the first tone after re-enable
  - exactly two new direct modules (jfreymuth/pulse v0.1.3, jfreymuth/oggvorbis v1.0.5 + its own jfreymuth/vorbis v1.0.2) recorded in SPEC §5/§11 BEFORE any code (D-55)
affects:
  - internal/session (SoundStopper optional capability, foldSound hoist)
  - cmd/goswitchd (sink started before install)
  - docs/SPEC.md §5/§11, docs/CONFIG.md, README.md
tech-stack:
  added:
    - github.com/jfreymuth/pulse v0.1.3 (MIT, pure Go — the native PulseAudio client)
    - github.com/jfreymuth/oggvorbis v1.0.5 (MIT, pure Go — the theme decoder; requires jfreymuth/vorbis v1.0.2)
  patterns:
    - silence-serving toneReader (select/default) — the library's blocking Start returns without a pending tone (the prototype's SIGQUIT lesson, corpus-pinned)
    - format-follows-PCM stream recreation; drop-if-busy enqueue; warn-once episodes with episode reopen (connect-failed / stream-failed / theme-unavailable)
    - fake-filesystem opener seam for the XDG resolver; fake dialer/runner seams for the worker and watcher corpora
key-files:
  created:
    - internal/sound/theme.go
    - internal/sound/pcm.go
    - internal/sound/pulse.go
    - internal/sound/theme_test.go
    - internal/sound/pcm_test.go
    - internal/sound/pulse_test.go
  modified:
    - internal/sound/sound.go (rewritten — subprocess machinery deleted)
    - internal/sound/sound_test.go (rewritten — old subprocess corpus deleted, disciplines ported)
    - internal/session/actor.go
    - internal/session/actor_test.go
    - cmd/goswitchd/main.go
    - docs/SPEC.md
    - docs/CONFIG.md
    - README.md
    - go.mod
    - go.sum
decisions:
  - stop-lifecycle reading: no explicit OFF→ON signal exists (the SoundSink interface stays byte-identical per the plan), so the muted period is simply connection-free (tones cannot arrive — the actor gate) and the first tone after re-enable IS the lazy redial; no dropped-tone state machine
  - the stream's initial format is the canonical synth mono shape (22050 Hz); the first decoded theme PCM recreates it at the theme's format (the locked «формат стрима следует за декодированным PCM» rule)
  - PD-2 supervision with a restart pause (5 s) for a monitor that cannot start — never a settings-read ticker; every (re)start performs the full two-key re-read
  - deps landed per-import (tidy drops unimported requires): oggvorbis+vorbis in Task 2's GREEN, pulse in Task 3's GREEN — the final graph is exactly the three measured modules
metrics:
  duration: 52 min
  completed: 2026-10-06
  commits: 8
  plan_head_before: ae9ca73
actuals:
  tasks: 4
  commits: 8
---

# Quick Task 261006-squ Summary: persistent PulseAudio stream + GNOME sound-theme playback

The ~150 ms tone latency behind every flip is gone by design: tones now come from the desktop's sound theme (Yaru .oga decoded once per pair) through ONE persistent native sound-server connection — the play write is the prototype's 133 µs handoff, the tone audible in the ~30 ms class — with the synthesized sines as the never-waits fallback, zero sound-server connections while muted, and best-effort degradation forever.

## Tasks Completed

| # | Task | Commits | Files |
|---|------|---------|-------|
| 1 | spec-delta — SPEC §5/§11 + CONFIG.md + README, docs-only (D-55) | 9bebd67 | docs/SPEC.md, docs/CONFIG.md, README.md |
| 2 | deps + XDG theme resolver + PCM cache + synthesized fallback (TDD) | d71a211 (RED), d4d9ee4 (GREEN), f7c2759 (refactor) | go.mod, go.sum, internal/sound/theme.go, pcm.go + tests |
| 3 | persistent stream playback + mute lifecycle + gsettings watcher (TDD) | daded27 (RED), cdad58b (GREEN) | internal/sound/sound.go, pulse.go + tests, go.mod, go.sum |
| 4 | wiring — actor ON→OFF mute close + daemon Start + full green iteration (TDD) | 6037265 (RED), 795c8dc (GREEN) | internal/session/actor.go, actor_test.go, cmd/goswitchd/main.go |

## Task 1 — spec-delta (D-55, docs-only before any code)

- SPEC §5: the dependency line kept verbatim as audit trail; a dated amendment comment (the §4.1/§11 form) records the owner's 2026-10-05 approval of exactly two direct pure-Go MIT modules, naming jfreymuth/vorbis v1.0.2 so the recorded set matches the measured go.mod graph; a new line names the grown closed set.
- SPEC §11: a dated «Дополнение владельца (2026-10-05, «делай» …)» block directly after the 2026-10-04 «Звуки при переключении» block records the redesign — XDG theme resolution fully async from playback, the persistent native connection superseding the per-tone subprocess (old text stays verbatim), the synthesized fallback, NO connection while muted (config, tray toggle, or system event-sounds per PD-1), instant clean failure on an unavailable server, the ~30 ms target.
- CONFIG.md: the `sound.autocorrect_event` row reworded to theme resolution + persistent connection + synthesized fallback; keys/defaults/tray persistence byte-identical. README «Switch sounds»: same rewording, no-subprocess and never-blocks stated.
- Verified: zero .go files in the diff (tracked-file check — see Deviations), `grep -c jfreymuth docs/SPEC.md` ≥ 1, mise run ci green on the docs-only tree.

## Task 2 — deps + resolver + cache + synth (RED d71a211 → GREEN d4d9ee4 → refactor f7c2759)

- **RED:** the resolver corpus over a fake-filesystem opener (XDG discovery order with data-home-first pinned byte-level, index.theme parse — Yaru single-entry/comma/semicolon Directories, Inherits chain, .oga→.ogg→.wav priority, cycle-safe walk terminating at `freedesktop`, empty-Directories themes yield nothing, total miss = sentinel); the cache/synth corpus (decode-once per (theme,event) with byte-identical repeats, atomic mid-rebuild swap where old-theme lookups answer while the new build is wedged, miss answers miss, deterministic synth at the stream's rate/channels with half amplitude and a loud zero-format refusal, real-decoder cell against /usr/share/sounds/Yaru/stereo/bell.oga skip-gated for CI).
- **GREEN:** theme.go (roots builder, first-index-wins, the small INI-section reader — stdlib only, no ini dependency) and pcm.go (oggvorbis adapter: float32 interleaved → clamped int16 LE carrying rate/channels; the atomic.Pointer snapshot cache, lock-free lookups; the synth ported from the old sine math minus the WAV container; shared putInt16LE hoisted).
- The real-decoder cell RAN against the live Yaru file on this machine (44.1 kHz stereo pinned).
- **Refactor:** golangci-lint from 21 issues to 0 (gochecknoglobals with justified nolints per the house pattern, goconst fixtures, named parse results, mnd/lll/errcheck/gofmt).

## Task 3 — persistent stream + mute lifecycle + watcher (RED daded27 → GREEN cdad58b)

- **RED:** the old subprocess corpus deleted with the machinery it pinned; the new Player corpus pins lazy single dial per connected period, drop-if-busy under a wedged dialer (the T-07-08-02 pin carried over), the Stop mute lifecycle (close exactly once, muted period dials nothing, re-enable redials), connect-failed warn-once with episode reopen, stream-failed as its own episode, the watcher (initial read, theme line → async swap, monitor exit → PD-2 full re-read, event-sounds=false → PD-1 close + ONE INFO, gsettings missing → fallback theme + one WARN), SetAutocorrectEvent fire-and-forget prefetch; the toneReader corpus pins silence-via-default (Start returns without a pending tone — the prototype's SIGQUIT lesson) and one-Read-one-wakeup.
- **GREEN:** pulse.go (the persistent wrapper — lazy dial off the gesture path, ONE int16-LE stream, format recreation on change, the 133 µs play handoff, idempotent close) and sound.go rewritten (seam methods = mutex read + non-blocking enqueue; the single worker owns connection lifecycle, cache lookup with synth-at-stream-format on miss, episode bookkeeping; the supervised watcher with the full two-key catch-up at every restart; the retired players, WAV writer and user-cache machinery are gone — the package runs no player subprocesses and keeps no cache files).
- Grep gate clean: no canberra/paplay/aplay/pw-play/UserCacheDir/toneWAV/wavHeader anywhere in internal/sound. lint 0 issues, -race green, tidy-diff clean.

## Task 4 — wiring (RED 6037265 → GREEN 795c8dc)

- **RED:** fakeSoundSink gains the SoundStopper capability; the ON→OFF fold cell failed exactly on the unpinned close; the OFF→OFF/nil cells passed green-by-design (the 06-03 continuity-pin precedent).
- **GREEN:** `SoundStopper` defined next to the byte-unchanged `SoundSink` (verified by the untouched pre-existing corpus: TestActor_FlipSoundsSink, TestActor_SoundSinkAfterModeRecord, TestActor_AutoCorrectFireSoundsSink, TestActor_SoundDisabledNoSink, TestActor_NilSoundSinkNoOp, TestActor_SoundFoldGatesFromSnapshot — all green with zero edits); the sound fold hoisted into `foldSound` (applySnapshot stayed under the cyclop ceiling); the daemon constructs the sink once, calls Start() BEFORE SetSoundSink (PD-3), install shape and arguments unchanged.
- README's Task 1 wording verified against the shipped semantics — no adjustment needed.
- Full green iteration: whole-module `go test -race ./...` green, `mise run ci` green (lint 0 issues), `mise run tidy-diff` green.

## Verification

- `mise run ci` green after every task; `mise run tidy-diff` green at every task end.
- Task 1: zero .go files in the diff; SPEC §11 audit trail intact (both prior verdicts verbatim).
- Task 3: grep gate clean (no player machinery left; the supervised gsettings watcher remains per PD-2/PD-3).
- Task 4: the pre-existing actor sound-gating corpus passes UNTOUCHED.
- Final graph: go.mod carries exactly jfreymuth/pulse v0.1.3 + jfreymuth/oggvorbis v1.0.5 direct (+ vorbis v1.0.2 via oggvorbis) and nothing else.

## Deviations from Plan

1. **[Rule 3 - Blocking] Dependencies land per-import, not all at Task 2.** `go mod tidy` (inside the tidy-diff gate) drops requires nothing imports yet — pinning all three modules at Task 2's RED was immediately undone by the gate. oggvorbis+vorbis stuck at Task 2's GREEN (pcm.go imports), pulse at Task 3's GREEN (pulse.go imports). The final go.mod graph is exactly the measured three-module set; no version drift anywhere.
2. **[Rule 1 - Bug, task-internal] TDD iterations fixed inside their own GREEN steps:** the «bare» resolver fixture needed an Inherits link for yaru to be reachable (resolver behavior was correct); the fake dialer returned (nil, nil) for a programmed-success slot (panicked the worker — the fake now falls through to a fresh conn); the episode-reopen cell routed the post-success failure through the mute close (a healthy conn never redials — the original expectation was wrong, not the worker); the toneReader's short-read at the tone's end is legal io.Reader form (test expectation fixed).
3. **[Note] Task 1's verify expression as literally written** (`git status --porcelain -- '*.go'`) trips on PRE-EXISTING untracked skill-asset .go files under .zcode/; the intent (zero .go files in the DIFF) was verified via tracked-file checks (`git diff -- '*.go'` empty; only the three docs files modified).

## Auth Gates

None.

## Known Stubs

None — the synthesized tone is the designed fallback path (SPEC §11 addendum), not placeholder functionality.

## Deferred Items

- **Owner by-ear acceptance (the final gate, by arrangement — not automatable):** after merge/deploy the owner flips layouts and listens — the tone must follow the gesture immediately (~30 ms class vs ~150 ms today), a GNOME Settings theme switch must retune the tone live without a daemon restart, and toggling «Звук» off must leave no sound-server connection (observable via `ss -x` for the pulse socket while muted).
- **Pre-existing, out of scope:** `TestRun_OnConnHookCalledOnce` (internal/ctlsvc) flaked ONCE at the baseline run (passed with -count=1 and in every subsequent full `mise run ci`, ~6 runs; fails only under repeated in-process `-count=5` runs — a test-isolation quirk, not touched by this task).

## Self-Check: PASSED

- All 8 commits present on gsd/sound-persistent-pulse-stream (ae9ca73..795c8dc, `git rev-list --count` = 8).
- All created files exist (theme.go, pcm.go, pulse.go + three test files; sound.go/sound_test.go rewritten in place).
- Working tree clean of tracked modifications; go.mod carries exactly the two approved direct modules.

## Gap Fix (post-verification, 2026-10-06)

Verification caught the PD-1 muted arm leaking: a tone arriving while `event-sounds=false` (config sound still on) redialed and played through a fresh connection. Fixed strict-TDD on the same branch: RED b02da6a (the two suppression cells failing on the redial), GREEN 38f892a (`Player.muted` atomic gate set by both watcher read paths; foldTone drops WITHOUT a redial; enqueue early arm; ONE INFO per observed transition; unmute clears the gate for the lazy redial). Gates: `mise run ci` + `mise run tidy-diff` green, whole-module `-race` green, mute cells green ×5. Corpus evidence: tone under the mute → dialCount unchanged, nothing played; unmute → a fresh connection (dial 2) carries the tone as its first write; startup-muted → nothing dials until the key turns true.
