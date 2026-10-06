---
phase: 261006-squ-sound-persistent-pulseaudio-stream-gnome
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - docs/SPEC.md
  - docs/CONFIG.md
  - README.md
  - go.mod
  - go.sum
  - internal/sound/sound.go
  - internal/sound/sound_test.go
  - internal/sound/theme.go
  - internal/sound/theme_test.go
  - internal/sound/pcm.go
  - internal/sound/pcm_test.go
  - internal/sound/pulse.go
  - internal/sound/pulse_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - cmd/goswitchd/main.go
autonomous: true
requirements:
  - QUICK-261006-squ-sound-persistent-pulse-stream
estimate:
  tokens: 84000
  raw_tokens: 56000
  tasks: 4
  confidence: low
must_haves:
  truths:
    - A flip or autocorrect tone on a healthy desktop comes from the GNOME sound theme (Yaru: decoded .oga PCM) through a PERSISTENT native sound-server connection — the tone is audible ~30 ms after the gesture (prototype-measured vs ~150 ms today), and the playback call never blocks the gesture (non-blocking enqueue; a busy worker drops the tone) — sound-latency requirement 1.
    - The playback path never waits for a settings read or a decode — sound-latency requirement 2: the theme name is read once at daemon start and tracked by a supervised `gsettings monitor`; a theme switch rebuilds the PCM cache asynchronously with an atomic swap; any cache miss (daemon start, theme switch, unknown event) plays the synthesized tone immediately.
    - With sound muted there is NO connection to the sound server — sound-latency requirement 3: the actor fold's ON→OFF transition closes the client+stream; re-enabling dials lazily on the first tone; the same close applies when the desktop's system-wide event-sounds key turns off (plan decision PD-1).
    - Sound-server unavailability is an instant clean failure (fast dial error, no hangs, no retries on the play path): one WARN per episode, quiet no-ops after, lazy retry on a later tone — best-effort, never an error on the gesture path.
    - The SoundSink seam (Flip / AutoCorrect / SetAutocorrectEvent) is byte-unchanged and all existing actor sound-gating tests stay green untouched.
    - The synthesized tone survives as the fallback (in-memory sine through the same persistent stream at the stream's rate/channels); internal/sound runs NO player subprocesses and writes NO cache files — the retired players and their WAV machinery are gone (the supervised gsettings watcher remains per PD-2/PD-3: settings infrastructure, not a player).
    - The module gains exactly two DIRECT dependencies — jfreymuth/pulse v0.1.3 + jfreymuth/oggvorbis v1.0.5 (MIT, pure Go) — recorded in SPEC §5/§11 BEFORE any code (D-55, the 06-01/07-01 discipline); the measured module graph: oggvorbis v1.0.5 itself requires jfreymuth/vorbis v1.0.2 (same author, MIT — its decoder core) and pulse's require block is empty; `mise run tidy-diff` stays clean.
  artifacts:
    - docs/SPEC.md §5 (dated dependency-set amendment) + §11 (dated 2026-10-05 sound redesign revision, audit trail preserved) + docs/CONFIG.md sound rows + README.md Switch sounds section
    - go.mod / go.sum with the two pinned direct modules plus oggvorbis's required jfreymuth/vorbis and no other drift
    - internal/sound/theme.go — gsettings get at start + supervised `gsettings monitor org.gnome.desktop.sound` + the XDG Sound Theme resolver (index.theme Directories/Inherits → XDG data dirs → <event>.oga|.ogg|.wav)
    - internal/sound/pcm.go — (theme, event) PCM cache with atomic swap + deterministic synthesized-tone generator at arbitrary rate/channels
    - internal/sound/pulse.go — persistent client+stream wrapper: lazy dial, silence-serving reader, format-follows-PCM, idempotent close
    - internal/sound/sound.go — Player over the new components: seam methods, drop-if-busy enqueue, single playback worker, warn-once episodes, Start/Stop lifecycle
    - internal/session/actor.go — SoundStopper optional seam + the ON→OFF fold transition closing the sink's connection
    - cmd/goswitchd/main.go — the sink constructed once, started at daemon start, installed via the unchanged SetSoundSink
  key_links:
    - actor applySnapshot ON→OFF transition → SoundStopper.Stop → pulse client+stream closed (requirement 3; lazy redial on the next tone)
    - Flip/AutoCorrect → non-blocking enqueue → playback worker → cache lookup (theme from the watcher, event from the seam) — miss → synthesized tone at the stream format (never a wait)
    - theme watcher (get at start + supervised monitor) → resolver → oggvorbis decode → atomic cache swap — strictly off the play path (requirement 2)
    - stream format follows the decoded theme PCM; a format change recreates the stream; the reader serves silence between tones so the library's Start never deadlocks (the prototype's SIGQUIT lesson)
    - config keys unchanged: sound.enabled / sound.autocorrect_event keep their schema, defaults and tray-toggle persistence (canberra retired; the theme events bell / autocorrect_event stay)
---

<objective>
feat(sound): persistent PulseAudio stream + GNOME sound-theme playback — eliminate the ~150 ms tone latency behind every flip.

Purpose: every tone today spawns a fresh `canberra-gtk-play` subprocess AFTER the flip (GTK init ~220 ms, warm start 100–330 ms; total bell 0.44–0.51 s) — the owner hears the switch before its confirmation. The owner approved the redesign («делай», 2026-10-05) with a LOCKED design recorded in `.planning/todos/pending/sound-latency.md`: per-tone experiments proved only a persistent native-protocol connection delivers the tone in ~30 ms; the chosen pure-Go modules are jfreymuth/pulse v0.1.3 + jfreymuth/oggvorbis v1.0.5 (both MIT, prototype-verified on the owner's machine; oggvorbis itself requires its same-author MIT decoder module jfreymuth/vorbis).

Output: SPEC/CONFIG/README spec-delta (D-55: docs before code), internal/sound rewritten (theme watcher + XDG resolver + PCM cache + persistent stream + synthesized fallback), the actor's mute gate closing the connection, daemon wiring. Strict TDD per task (red → green → refactor); every task ends green on `mise run ci`.
</objective>

<execution_context>
@/home/nil/DiskD/W/Djarvur/goswitch/.zcode/gsd-core/workflows/execute-plan.md
@/home/nil/DiskD/W/Djarvur/goswitch/.zcode/gsd-core/templates/summary.md
</execution_context>

<context>
@.planning/todos/pending/sound-latency.md
@.planning/STATE.md
@internal/sound/sound.go
@internal/session/actor.go
@cmd/goswitchd/main.go
@docs/SPEC.md
@docs/CONFIG.md
@go.mod
@mise.toml

Load the go-ultimate project skill (.zcode/skills/go-ultimate/) before writing Go — conventions and review checklist come from there. Engineering directives bind every task: strict TDD (red → green → refactor), green iteration (`go build ./...`, `go test -race ./...`, `golangci-lint run` via mise), conventional commits, changes land via PR on this branch (gsd/sound-persistent-pulse-stream).

Wiring facts (verified on the current tree; read before the tasks):
- internal/sound/sound.go (346 lines): Player with two pinned player binaries resolved via LookPath at the first tone, no-pipes child spawn, synthesized WAV tones materialized in the user cache dir for the second player, warn-once episodes per reason (warned map + closeEpisode on a healthy launch), fire-and-forget Flip/AutoCorrect, SetAutocorrectEvent under the mutex. The WHOLE file and its corpus are superseded by this plan — Tasks 2–3 replace it.
- internal/session/actor.go: SoundSink interface (~882) — Flip() / AutoCorrect() / SetAutocorrectEvent(string), fire-and-forget by contract, defined at the point of use. SetSoundSink (~903) installs it (nil = silent). soundSinkFor (~1059) is the SINGLE play gate: returns nil when a.opts.SoundEnabled is false. The gate's callers: flipTo (~1904) and autoConfirm (~2363) — both call the sink under a.mu. applySnapshot (~1283) folds `a.opts.SoundEnabled = snap.Sound.EffectiveEnabled()` live (hot reload); SetOptions seeds it at startup. a.mu → sink.mu is a ONE-WAY ordering (the sink never calls back).
- cmd/goswitchd/main.go:190 — `actor.SetSoundSink(sound.New(config.DefaultSoundFlipEvent, cfg.Sound.EffectiveAutocorrectEvent()))` — the only production construction site.
- internal/config/config.go: DefaultSoundFlipEvent = "bell" (schema constant, not configurable), DefaultSoundAutocorrectEvent = "message"; Sound section: Enabled (*bool, absent = ON), AutocorrectEvent. Schema keys do NOT change in this plan.
- mise tasks (repo root): build / vet / lint / test / tidy-diff / ci. `mise run tidy-diff` fails on any go.mod/go.sum drift.
- Sound-theme facts from the locked design (owner machine, verified): theme files live at /usr/share/sounds/Yaru/stereo/<event>.oga (index.theme: `[Sound Theme]` with `Directories=stereo`); Yaru PCM is 44.1 kHz stereo Ogg Vorbis; prototype numbers — persistent stream: connect 3 ms, tone send 133 µs, audible ~30 ms stable; per-tone connect+stream: 111–161 ms; server down = instant dial failure.
- The prototype knowledge lives ONLY in the todo file (/tmp/audtest is temporary) — the silence-serving reader (fixes the library's blocking Start) and the drop-if-busy shape must be re-derived in code and pinned by the corpus.

## Plan decisions (settled here, binding for the executor)

- **PD-1 — `org.gnome.desktop.sound event-sounds` is treated like the master mute.** goswitch tones ARE event sounds (bell / message theme events); a user who turned system event sounds off opted out of exactly this class, so honoring it is the consistent reading of sound-latency requirement 3. The key rides the SAME monitored gsettings schema (zero extra processes). Transition to off = close the connection + ONE INFO («event sounds disabled system-wide» — an observed preference, not a degradation, so INFO not WARN); back on = lazy redial on the next tone. Effective play = config `sound.enabled` (the actor gate) AND `event-sounds` (the package watcher). Unreadable key = assume true (the GNOME default), one WARN in the theme-unavailable episode.
- **PD-2 — monitor reliability = supervision, not a ticker.** `gsettings monitor` runs supervised: an exited monitor restarts, and every (re)start performs a FULL re-read of both keys (catch-up for anything missed while down). No independent periodic re-read — the dead-monitor window is bounded by supervision and the restart re-read closes it. If the whole gsettings path is unavailable (binary/schema missing), the theme name degrades to the XDG spec's fallback theme with ONE WARN episode and the synthesized tones carry playback.
- **PD-3 — the theme watcher runs from daemon start regardless of `sound.enabled`.** It is settings infrastructure (a gsettings subprocess), not a sound-server connection; requirement 3 forbids the sound-server connection specifically, and a pre-warmed theme keeps the first tone after an enable instant. The pulse client/stream stays down while muted — that is the requirement's letter.
</context>

<tasks>

<task type="auto">
  <name>Task 1: spec-delta — SPEC §5/§11 + CONFIG.md + README, docs-only, BEFORE any code (D-55)</name>
  <files>docs/SPEC.md, docs/CONFIG.md, README.md</files>
  <action>
Docs first, zero .go changes in this task's diff (the 06-01/07-01 D-55 precedent: the spec records the decision before the code exists; the stack is frozen in SPEC §5, so new dependencies legally enter only through a dated spec revision).

docs/SPEC.md:
- §5, the dependency line («только stdlib + godbus/dbus + минимальные зависимости»): append a dated amendment comment in the file's established style (the §4.1/§11 dated-comment forms) recording the owner's 2026-10-05 approval («делай») of exactly two additional DIRECT pure-Go MIT dependencies for sound playback — the native PulseAudio client (jfreymuth/pulse) and the Ogg Vorbis decoder (jfreymuth/oggvorbis, which itself requires its same-author MIT decoder module jfreymuth/vorbis — name it in the amendment so the recorded set matches the measured go.mod graph) — and reword the line to name the grown closed set (stdlib + godbus + those two + минимальные). Keep the old line readable as audit trail (the §11 revision pattern: old verdict stays, dated addition follows).
- §11, after the 2026-10-04 «Звуки при переключении» block: a dated «Дополнение владельца (2026-10-05, …)» block recording the approved redesign: (1) tones play from the desktop's standard sound theme, resolved per the XDG Sound Theme and Naming Spec (theme-name from org.gnome.desktop.sound, read at daemon start, changes tracked — fully async from playback); (2) playback through the daemon's PERSISTENT native sound-server connection (pure Go), not a per-tone subprocess — the retired mechanism of the 2026-10-04 block is superseded (its text stays verbatim as audit trail); (3) the synthesized tones remain the fallback whenever theme PCM is not ready (start, theme switch) — the play path never waits; (4) muted sound (config switch, tray toggle, or the system-wide event-sounds key off — PD-1) means NO sound-server connection at all: lazy connect on enable, close on disable; (5) an unavailable sound server is an instant clean failure — one WARN, quiet after, best-effort; latency target: the tone audible immediately after the flip (~30 ms prototype vs ~150 ms today).

docs/CONFIG.md: the `sound.autocorrect_event` row and the schema-intro sound sentences — replace the playback wording (the player-subprocess description) with the persistent-stream + theme-resolution description; keys, defaults, optionality and tray persistence stay byte-identical (no schema change). State the fallback: a not-yet-decoded event plays the synthesized tone.

README.md «Switch sounds» section: same rewording (theme sounds through the persistent connection, synthesized fallback, no subprocess, never blocks the switch).

Commit: docs(261006-squ) conventional form.
  </action>
  <verify>
    <automated>test -z "$(git status --porcelain -- '*.go')" && { [ -n "$(git status --porcelain)" ] || test -z "$(git show --name-only --format='' HEAD -- '*.go')"; } && [ "$(grep -c jfreymuth docs/SPEC.md)" -ge 1 ] && mise run ci</automated>
  </verify>
  <done>SPEC §5 names the grown dependency set with the dated owner approval; §11 carries the dated redesign block with the 2026-10-04 verdict preserved verbatim above it; CONFIG.md and README describe the new mechanism with unchanged schema; the diff touches no .go file; mise gates green on the docs-only tree (the 07-08 docs-only precedent).</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: deps + XDG sound-theme resolver + PCM cache + synthesized fallback (TDD)</name>
  <files>go.mod, go.sum, internal/sound/theme.go, internal/sound/theme_test.go, internal/sound/pcm.go, internal/sound/pcm_test.go</files>
  <action>
Dependencies (the locked design pins exact versions; owner-approved, licenses verified in the todo — no further gate): `go get github.com/jfreymuth/pulse@v0.1.3 github.com/jfreymuth/oggvorbis@v1.0.5`, then confirm go.sum's new module set matches the MEASURED graph (proxy.golang.org): oggvorbis v1.0.5 requires jfreymuth/vorbis v1.0.2 (same author, MIT) and pulse v0.1.3's require block is empty — expect exactly those three modules added, nothing else; `mise run tidy-diff` is clean.

RED first (new test files; the old sound.go corpus keeps compiling — this task is additive), commit test(261006-squ):

theme_test.go — the resolver corpus over a fake-filesystem seam (an opener func(path) (io.ReadCloser, error); no real disk in unit tests):
- theme discovery order per the XDG Sound Theme spec: $XDG_DATA_HOME/sounds (default ~/.local/share/sounds) BEFORE $XDG_DATA_DIRS/sounds (default /usr/local/share:/usr/share); the first index.theme wins.
- index.theme parsing: `[Sound Theme]` section gives Directories (support the single-entry Yaru form `Directories=stereo`; accept comma- and semicolon-separated lists defensively) and Inherits (a chain). Within a theme's directory the event file resolves as <event>.oga, then .ogg, then .wav (the locked extension priority).
- the Inherits chain: theme → its inherits → … terminating at the spec's fallback theme; a cycle in the chain terminates via a visited set; the first hit across the whole chain wins; total miss = a resolvable error.
- empty/absent Directories in an index.theme = that theme yields nothing for the directory search (moves on down the chain, never panics).

pcm_test.go — cache + synth corpus:
- cache keyed by (theme, event): a decoded entry is returned byte-identically on repeat lookups; the loader decodes each pair ONCE (a counting fake decoder proves single decode).
- atomic swap: a theme-change rebuild publishes a new snapshot under a new theme key while lookups of the old theme keep answering from the old snapshot mid-rebuild (the never-waits contract — a lookup never blocks on a load in flight); a missing pair answers miss (the caller synthesizes).
- decode via oggvorbis: float32 interleaved → int16 little-endian PCM carrying rate + channels; REAL-decoder test against a system theme file (e.g. the Yaru stereo directory) gated on file presence with t.Skip when absent (the 06-02 CI-independence SKIP precedent) — pins rate/channels extraction and the conversion on real input where the desktop provides it.
- synthesized tones: deterministic pure sine at the STREAM's rate/channels (no resampling) — flip pitch 880 Hz, autocorrect pitch 1320 Hz, 60 ms, half amplitude, int16 LE; byte-stable across runs (no clocks, no rand — the old tone determinism pin carried over); a nil/zero format argument is refused loudly in tests (the synth always targets a concrete stream format).

GREEN: internal/sound/theme.go (data-dir discovery, index.theme parse via a small INI-section reader — stdlib only, no ini dependency; the XDG resolver returning an opened theme file) and internal/sound/pcm.go (the (theme,event) cache with atomic.Pointer snapshot swap; the oggvorbis decode adapter with the int16 conversion; the synth generator ported from the old sine math, minus the WAV container). Name identifiers in the package's established voice; errors wrapped with %w carrying reasons only — never any user content.

REFACTOR: hoist shared little-endian/int16 helpers; keep both files inside the complexity ceilings (cyclop 15 / gocyclo 20 / gocognit 30).

Old code untouched this task — internal/sound must build and lint green with the old player still present alongside the new files.
  </action>
  <verify>
    <automated>go test -race -count=1 ./internal/sound/ && mise run tidy-diff && mise run ci</automated>
  </verify>
  <done>Resolver corpus green (discovery order, index.theme Directories/Inherits, extension priority, cycle-safe fallback chain); cache decodes once per (theme, event), swaps atomically, never blocks a lookup; synth byte-deterministic at the stream's format; real-decoder test green where system theme files exist and skips cleanly in headless CI; exactly the two approved direct modules in go.mod plus oggvorbis's required jfreymuth/vorbis v1.0.2 and nothing else, tidy-diff clean; mise gates green.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: persistent stream playback + mute lifecycle + gsettings watcher — internal/sound rewrite (TDD)</name>
  <files>internal/sound/sound.go, internal/sound/sound_test.go, internal/sound/pulse.go, internal/sound/pulse_test.go</files>
  <action>
RED first — pulse_test.go (a fake dialer seam models the library: connect returns a client whose stream records writes/format/closes; the fake makes Start block until a first buffer is delivered, mirroring the real library's contract that killed the prototype with a SIGQUIT), commit test(261006-squ):

pulse.go — the persistent connection wrapper over jfreymuth/pulse:
- lazy dial on the first play request after enable (the prototype's 3 ms connect lives OFF the gesture path — the worker dials, never the seam caller); server down = the dial's own fast error, no retry loop, no deadline machinery on the play path.
- ONE persistent stream; its format (rate, channels, int16 LE) follows the first decoded theme PCM served; a later format change drains+closes+recreates the stream (the locked «формат стрима следует за декодированным PCM» rule).
- the reader between tones serves SILENCE via select/default — Start must return without a pending tone (pin as the corpus's blocking contract); between tones the reader blocks on the tone channel with silence as the default — no busy spin (a wakeup-count assertion in the corpus).
- play(pcm) = one write (the prototype's 133 µs); close() closes stream+client, idempotent, safe on a never-dialed wrapper.

sound_test.go — the Player corpus (the old subprocess tests are DELETED with the machinery they pin; the warn-once episode discipline, the non-blocking play pin and the event re-pin test are PORTED to the new shapes):
- Flip/AutoCorrect enqueue without blocking: a wedged worker (a fake dialer that never completes) must not stall the seam call — the drop-if-busy select/default answers immediately (the T-07-08-02 non-blocking pin, carried over).
- lazy connect: construction + Start dial NOTHING; the first tone dials exactly once; a burst of tones during one connected period reuses the client (dial count stays 1).
- Stop() (the mute lifecycle): closes client+stream exactly once; repeated Stop is a no-op; after Stop, tones are dropped WITHOUT a redial until the muted state ends — the next tone after re-enable dials again (requirement 3's lazy connect on enable, close on disable).
- degradation: dial failure = ONE warn-once WARN per episode (reason connect-failed), quiet afterwards, a later successful tone closes the episode (a subsequent failure warns again — the old closeEpisode discipline); a stream write failure = its own episode reason; neither ever surfaces an error to the caller.
- the gsettings watcher (a fake command runner seam): Start reads theme-name AND event-sounds once (get), then runs the supervised monitor; a monitor process exit restarts it and the restart performs a FULL re-read (PD-2 — pin the catch-up); a theme change triggers an async rebuild and the atomic swap of Task 2's cache (the play path observable: the NEXT tone uses the new theme without any synchronous wait); gsettings missing/schema absent = fallback theme + ONE warn-once WARN; the event-sounds=false transition closes the connection with ONE INFO and re-enable redials lazily (PD-1); the watcher runs from Start regardless of any muted state (PD-3).
- SetAutocorrectEvent re-pins the event and prefetches the new (theme, event) pair asynchronously — a tone arriving before the prefetch lands plays the synthesized tone, never waits (requirement 2).

GREEN: rewrite sound.go as the Player over theme.go/pcm.go/pulse.go: New(flipEvent, autocorrectEvent string) keeps the wiring signature; Start() (idempotent) launches the playback worker + the watcher; the seam methods do a mutex read + non-blocking enqueue only; Stop() signals the worker via a buffered control channel and returns without waiting for a mid-tone write; the worker owns connection lifecycle, cache lookup (miss → synth at the current stream format), episode bookkeeping and the closed-reason vocabulary (connect-failed / stream-failed / theme-unavailable — names and library errors in logs only, never user content). The package ends subprocess-free and file-cache-free: the retired players, the WAV writer and the user-cache machinery are deleted with their tests.

REFACTOR: keep the worker loop under the complexity ceilings; hoist the episode/warn bookkeeping into the old warn/closeEpisode shape.

Every behavior task step ends green: `go build ./... && go test -race ./internal/sound/ && golangci-lint run` (mise tasks) — no «fix lint later».
  </action>
  <verify>
    <automated>go test -race -count=1 ./internal/sound/ && ! grep -rniE 'canberra|paplay|aplay|pw-play|UserCacheDir|toneWAV|wavHeader' internal/sound/ --include='*.go' && mise run ci && mise run tidy-diff</automated>
  </verify>
  <done>Player corpus green: lazy single dial, drop-if-busy non-blocking seam calls, Stop-then-redial mute lifecycle, warn-once episodes with episode reopen, supervised watcher with catch-up re-read, async theme swap, event-sounds mute per PD-1, synthesized fallback on every cache miss; the package runs no player subprocesses (retired player binaries and the WAV cache machinery gone — the supervised gsettings watcher remains per PD-2/PD-3) and keeps no cache files; -race clean; mise gates green.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 4: wiring — actor mute close + daemon Start + README polish + full green iteration</name>
  <files>internal/session/actor.go, internal/session/actor_test.go, cmd/goswitchd/main.go</files>
  <action>
RED first (internal/session/actor_test.go), commit test(261006-squ):
- the fake sound sink gains a stop counter; new tests: a fold that transitions sound ON→OFF (TestActor_SoundFoldGatesFromSnapshot's shape: reloadSource + src.set + flipMode) calls the installed sink's stop exactly once; a fold that stays OFF stops nothing extra; OFF→ON stops nothing; a sink without the stop capability folds silently (type-assertion miss = no-op); a nil sink folds silently.

GREEN:
- internal/session/actor.go: define SoundStopper (interface { Stop() }) next to SoundSink — same point-of-use style, the SoundSink interface itself stays BYTE-IDENTICAL (pin: the three-method declaration and both call sites in flipTo/autoConfirm unchanged). In applySnapshot, where the sound switch folds: when the effective value transitions ON→OFF relative to a.opts.SoundEnabled, type-assert the installed sink to SoundStopper and Stop() it (the sink's own mutex makes it quick and re-entrant-safe under a.mu; the actor.mu → sink.mu one-way ordering already documented holds). The gate in soundSinkFor keeps its exact current nil form — the fold transition is the close trigger, the gate remains the single play gate.
- cmd/goswitchd/main.go (~190): construct the sink once into a variable, call its Start() BEFORE actor.SetSoundSink (the watcher must run from daemon start per PD-3, and the fold's stopper assertion needs the sink installed before any fold can observe it); the seam install call itself keeps its exact shape and arguments.
- README.md «Switch sounds»: confirm the Task 1 wording matches the shipped behavior (muted = no connection, lazy redial, synthesized fallback); adjust only if the implementation's exact semantics differ from the Task 1 text.

REFACTOR + full green iteration (directive 2): `go test -race ./...` across the WHOLE module — the pre-existing actor sound-gating corpus (TestActor_FlipSoundsSink, TestActor_SoundSinkAfterModeRecord, TestActor_AutoCorrectFireSoundsSink, TestActor_SoundDisabledNoSink, TestActor_NilSoundSinkNoOp, TestActor_SoundFoldGatesFromSnapshot) must pass UNTOUCHED (the seam contract held). Then the full gate.

Owner by-ear (the final acceptance, by arrangement — not automatable): after merge/deploy the owner flips layouts and listens — the tone must follow the gesture immediately (prototype expectation ~30 ms vs ~150 ms today), a GNOME Settings theme switch changes the tone without a daemon restart, and toggling «Звук» off leaves no sound-server connection (observable via ss -x for the pulse socket while muted).
  </action>
  <verify>
    <automated>go test -race -count=1 ./internal/session/ ./internal/sound/ ./cmd/... && mise run ci && mise run tidy-diff</automated>
    <human-check>Owner, on the GNOME machine: flip → tone immediate by ear; GNOME theme switch retunes the tone live; «Звук» off → no pulse socket connection, on → tone returns without restart.</human-check>
  </verify>
  <done>The ON→OFF fold closes the sink once (corpus-pinned), OFF folds and re-enables stay clean; SoundSink byte-unchanged with the whole pre-existing actor corpus green untouched; the daemon starts the watcher and installs the sink at startup; whole-module -race green; mise ci + tidy-diff green; owner by-ear check noted as the final gate.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| XDG data dirs → decoder | theme files (index.theme, .oga/.ogg/.wav) read from system/user data directories and parsed+decoded by in-process code |
| daemon → sound server | a persistent native-protocol socket connection (unix/tcp) carrying PCM |
| daemon → gsettings | a long-lived monitor subprocess over the project's pinned-argv shape |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-SQU-01 | Tampering | theme.go decoder input (index.theme / .oga from data dirs) | medium | mitigate | decode errors are contained values, never panics: a malformed/hostile theme file resolves to a warn-once episode and the synthesized fallback; the resolver caps chain depth via a visited set; theme paths derive from XDG constants + pinned event names only |
| T-SQU-02 | Information Disclosure | pcm.go cache, all logs | high | mitigate | no user text exists anywhere in the sound path — logs carry reason slugs and library errors only (D-20/D-21 unchanged); the package is grep-gated free of any text-content plumbing |
| T-SQU-03 | Denial of Service | pulse.go worker, seam callers | high | mitigate | the gesture path never touches the connection: drop-if-busy enqueue, single worker, dial errors are instant; a wedged server drops tones instead of stalling flips; silence-serving reader prevents the blocking-Start wedge the prototype hit |
| T-SQU-04 | Elevation of Privilege | theme.go gsettings calls | low | mitigate | argv carries only the pinned schema/key constants — no shell, no interpolated user data (the T-05-02-01 install-package shape); the monitor reads, never writes |
| T-SQU-SC | Tampering | go.mod supply chain (two new direct modules) | medium | mitigate | owner-approved in the locked design with verified MIT licenses and exact versions via the module proxy; prototype ran on the owner's machine; module set measured from the registry: direct = pulse v0.1.3 + oggvorbis v1.0.5, oggvorbis requires jfreymuth/vorbis v1.0.2 (same author, MIT), pulse's require block is empty; versions pinned and go.sum-committed; tidy-diff gate keeps the set closed |
</threat_model>

<verification>
- `mise run ci` green after every task (build + vet + strict golangci-lint + test -race).
- `mise run tidy-diff` green after every dependency-touching task (no go.mod/go.sum drift).
- Task 1: zero .go files in the diff; SPEC §11 audit trail intact (the 2026-10-04 verdict verbatim above the dated addition).
- Task 3 grep gate: no player machinery left in internal/sound (retired player binaries and the WAV cache machinery absent from the package; the supervised gsettings watcher stays per PD-2/PD-3).
- Task 4: the pre-existing actor sound-gating corpus passes UNTOUCHED (the seam contract held byte-for-byte).
</verification>

<success_criteria>
- Tones on a healthy desktop come from the GNOME sound theme through a persistent connection, audible immediately after the flip (~30 ms prototype-measured class), never blocking or erroring the gesture — sound-latency requirements 1.
- Theme settings are read at start and tracked asynchronously (supervised monitor with catch-up re-read); the play path never waits for a settings read or a decode; any cache miss plays the synthesized tone at once — requirement 2.
- Muted sound (config, tray toggle, or system event-sounds off per PD-1) leaves zero sound-server connections; enable reconnects lazily — requirement 3.
- Server unavailability = instant clean failure, one WARN per episode, quiet after, best-effort forever.
- Exactly the two owner-approved direct modules (plus oggvorbis's required jfreymuth/vorbis) entered go.mod, recorded in SPEC before code; schema keys unchanged; SoundSink seam byte-unchanged; every task's iteration green under the full mise gate.
</success_criteria>

<output>
Create `.planning/quick/261006-squ-sound-persistent-pulseaudio-stream-gnome/261006-squ-SUMMARY.md` when done
</output>
