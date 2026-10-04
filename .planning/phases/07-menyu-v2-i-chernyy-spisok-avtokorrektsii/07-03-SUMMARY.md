---
phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
plan: 03
subsystem: config
tags: [yaml-persist, yaml-node-roundtrip, atomic-write, adopt-watch, hot-reload, fsnotify, tdd, tracer]

# Dependency graph
requires:
  - phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii (plan 02)
    provides: autocorrect/apps_blocklist + sound schema, strict Load, Defaults(), EffectiveEnabled, ACTIVE-condition validation
  - phase: 03 (plan 04)
    provides: Watcher with debounce + WithLoader seam, D-32 last-good reload contract
provides:
  - config.SetAutocorrectEnabled(path, on) — yaml.v3 Node round-trip persist of the autocorrect toggle; comments and document structure survive; atomic same-dir temp+rename, 0600
  - config.SetSoundEnabled(path, on) — the «Звук» toggle on the same persist mechanism (owner decision)
  - config.EnsureDocument(path) — full defaults document creation when absent, byte-level no-op when present («Настройки…» prerequisite)
  - config.DefaultPath() — the canonical ~/.config/goswitch/config.yaml resolver (XDG_CONFIG_HOME-aware)
  - goswitchd adopt+watch — daemon without -config loads the default path when present (broken file = loud refusal), stays on green defaults when absent (no file generation at start), watcher always runs on the factual path; INFO «config loaded»/«config defaults» contract
affects: [07-05 menu toggle and Settings wiring, 07-06 menu e2e (restart survival), docs/CONFIG.md persist semantics]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
actuals:
  tokens: 10825   # 43302 chars over the realized internal/+cmd/ diff
  tasks: 3
  commits: 6      # MEASURED: git rev-list --count gsd-plan-head-before-07-03..HEAD

# Tech tracking
tech-stack:
  added: []       # zero new dependencies — yaml.v3/stdlib only, per the phase constraint
  patterns:
    - "yaml.v3 Node round-trip persist: DocumentNode → Content[0] walk, in-place scalar flip, SetIndent(2), same-dir temp + os.Rename at 0600"
    - "ensure-branch: absent file → complete Defaults()-equivalent document built by marshaling the Config struct itself (strict-decodable by construction)"
    - "adopt+watch: WithLoader seam feeds a defaults-fallback loader so the watcher constructs before the document exists"

key-files:
  created:
    - internal/config/writer.go
    - internal/config/writer_test.go
  modified:
    - cmd/goswitchd/main.go
    - cmd/goswitchd/main_test.go

key-decisions:
  - "RED-stub precedent (06-03) for all three tasks: each RED commit carries the signature shape with a loud stub refusal, so RED fails on real assertions — the tracer, ensure and adopt corpora each landed test(07-03) → feat(07-03)"
  - "[Rule 3] startConfigWatcher creates the adopted config DIRECTORY (0700) before watching — a fresh install has neither file nor dir and the watcher watches the directory; the FILE itself is still never generated at start (04-02 contract intact)"
  - "Section insertion/replace shares one upsertToggle path (in-place flip, missing-key insert, null-section replace, whole-section append) — never a duplicate section, which yaml.v3 refuses at decode"
  - "buildFullDocument round-trips the marshaled Config struct (not hand-rolled key lists) — the created document is strict-decodable by construction and thresholds come from Defaults(), one source of truth"

patterns-established:
  - "Persist writer discipline: writer logs NOTHING (paths/outcomes live with the caller) — D-20/D-21 extended; every writer value test reads back through strict Load (Pitfall 6)"
  - "adoptLoad watcher-loader pattern: ErrNotExist → Defaults, everything else propagates — absence stays a green state under hot reload too"

requirements-completed: ["MACR-ACL (blocklist-ревизия)"]

coverage:
  - id: D1
    description: "Persist flip: SetAutocorrectEnabled flips one scalar in a live commented YAML via Node round-trip, comments survive, atomic temp+rename, 0600, broken source untouched, idempotent"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_FlipRoundTrip"
        status: pass
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_CommentsSurvive"
        status: pass
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_AtomicWrite"
        status: pass
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_IdempotentFlip"
        status: pass
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_BrokenSourceUntouched"
        status: pass
    human_judgment: false
  - id: D2
    description: "Ensure semantics: first toggle on an absent file creates the complete Defaults()-equivalent document (strict Load passes); EnsureDocument creates + byte-no-ops; missing autocorrect/sound sections insert whole with the rest preserved; SetSoundEnabled round-trips through EffectiveEnabled"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_CreatesFullDocumentOnAbsent"
        status: pass
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_EnsureDocumentCreatesAndNoOps"
        status: pass
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_InsertsMissingSection"
        status: pass
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_SoundFlipReverseRead"
        status: pass
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_SoundInsertMissingSection"
        status: pass
      - kind: unit
        ref: "internal/config/writer_test.go#TestWriter_FullDocumentStrictLoad"
        status: pass
    human_judgment: false
  - id: D3
    description: "Adopt+watch: daemon without -config loads an existing default-path file, stays on green defaults (creating nothing) when absent, refuses the start loudly on a broken one, explicit -config stays priority, watcher always on the factual path (non-nil on an absent file)"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: unit
        ref: "cmd/goswitchd/main_test.go#TestLoadConfigAdoptExisting"
        status: pass
      - kind: unit
        ref: "cmd/goswitchd/main_test.go#TestLoadConfigAdoptAbsent"
        status: pass
      - kind: unit
        ref: "cmd/goswitchd/main_test.go#TestLoadConfigAdoptBroken"
        status: pass
      - kind: unit
        ref: "cmd/goswitchd/main_test.go#TestLoadConfigExplicitPriority"
        status: pass
      - kind: unit
        ref: "cmd/goswitchd/main_test.go#TestAdoptWatcherServesDefaultsOnAbsentFile"
        status: pass
    human_judgment: false
  - id: D4
    description: "Live adopt+echo on the owner's desktop: a real daemon start without -config picks up a toggle-created file via the 200 ms debounced echo (research Q3b predicted exactly one harmless reload); INFO lines visible in the journal"
    verification: []
    human_judgment: true
    rationale: "Requires the live GNOME session and the menu toggle of plan 07-05 — the restart-survival e2e case of plan 07-06 is the sanctioned live oracle (research A5/Q6); unit corpus cannot drive the desktop."

# Metrics
duration: 32 min
completed: 2026-10-04
status: complete
---

# Phase 7 Plan 3: Persist writer + adopt/watch Summary

**yaml.v3 Node round-trip persist writer (flip one scalar, comments survive, atomic temp+rename 0600, full-document ensure) plus daemon adopt+watch of ~/.config/goswitch/config.yaml with loud refusal on broken files**

## Performance

- **Duration:** 32 min
- **Started:** 2026-10-04T21:24:41Z
- **Completed:** 2026-10-04T21:56:44Z
- **Tasks:** 3 (tracer + tdd + auto, each RED → GREEN)
- **Files modified:** 4 (2 created, 2 modified)

## Accomplishments

- The persist path exists end to end: `config.SetAutocorrectEnabled` / `config.SetSoundEnabled` flip the single target scalar in a live commented document — the walk starts at `DocumentNode → Content[0]` (Pitfall 6), `SetIndent(2)` matches the docs, the write lands through a same-directory temp + `os.Rename` at 0600, and any failure leaves the original byte-untouched.
- Ensure semantics per the owner's adopt+watch decision: the first toggle on a machine without a config CREATES the complete Defaults()-equivalent document (strict-decodable by construction — it round-trips the marshaled `Config` struct), missing sections insert whole (never a duplicate), `EnsureDocument` gives «Настройки…» a valid buffer and is a byte-level no-op when the file exists.
- The daemon adopts the default path: existing file loads with the explicit-contract (broken file = loud start refusal, T-07-03-03), absent file stays the green defaults start with NO file generation (04-02), the watcher runs unconditionally on the factual path — the file the first toggle creates is picked up by the same hot-reload loop. INFO contract distinguishes «config loaded» from «config defaults» (path + tap window only, never section content).

## Task Commits

Each task was committed atomically (TDD: RED → GREEN):

1. **Task 1 (tracer): persist flip path** — `395382a` (test RED) → `852033d` (feat GREEN)
2. **Task 2 (tdd): ensure-semantics** — `cae12a7` (test RED) → `0fa029c` (feat GREEN)
3. **Task 3 (auto): adopt+watch** — `d1c5f52` (test RED) → `1afaac0` (feat GREEN)

_Note: REFACTOR commits were not needed — cleanup (assertion helper extraction, error wrapping) happened within the GREEN commits and stayed green._

## Files Created/Modified

- `internal/config/writer.go` (created) — SetAutocorrectEnabled / SetSoundEnabled / EnsureDocument / DefaultPath; internal setDocumentToggle → upsertToggle → rewriteAtomically → writeAtomically chain; zero log statements
- `internal/config/writer_test.go` (created) — tracer corpus (flip round-trip via Load, comment survival, idempotence, atomicity, 0600, broken-source) + ensure corpus (full-document creation, byte-no-op, section insertion, sound flip/insert, strict decode)
- `cmd/goswitchd/main.go` (modified) — loadedConfig result type, adopt branch in loadConfig, adoptLoad watcher loader, unconditional startConfigWatcher, INFO contract, updated menu-reload comment
- `cmd/goswitchd/main_test.go` (modified) — four TestLoadConfig adopt cases + TestAdoptWatcherServesDefaultsOnAbsentFile

## Decisions Made

- RED-stub shape for every task (06-03 precedent): stubs refuse loudly so RED evidence is a real target-test assertion failure; the tracer's RED evidence was validated (`RED_EVIDENCE_OK`, target `TestWriter_CreatesFullDocumentOnAbsent`, 6/6 failing) via `check tdd-red-evidence` with a mechanical TAP projection of the `go test -v` run (the verb parses node --test TAP; the projection preserves the exact tests and outcomes).
- `[Rule 3]` The adopt watcher creates the config DIRECTORY (0700): a fresh install has neither file nor directory, the watcher watches the directory (watch.go untouched), and refusing to start over a missing empty directory would break the 04-02 green-defaults contract. The FILE is still generated only by a user action.
- One upsert path instead of plan-literal section-missing-only insertion: a section present without its key gets the key inserted (no duplicate-section decode refusals), a null section is replaced wholesale — small correctness extensions of the pinned cases.
- `TestLoadConfigExplicitPriority` landed green-by-design (the continuity pin of the byte-as-today explicit contract — the 06-04 continuity-pin precedent); the other four adopt tests were genuinely RED.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Adopt watcher refused to start on a fresh machine (no config directory)**
- **Found during:** Task 3 GREEN (TestAdoptWatcherServesDefaultsOnAbsentFile)
- **Issue:** the watcher watches the config's DIRECTORY; on a fresh install `~/.config/goswitch/` does not exist, `fsnotify Add` fails, and the daemon would refuse to start — breaking the plan's own must-have (absent file = green defaults start with the watcher running)
- **Fix:** the adopt branch runs `os.MkdirAll(filepath.Dir(lc.path), 0o700)` before constructing the watcher; the explicit -config path keeps its byte-as-today refusal behavior
- **Files modified:** cmd/goswitchd/main.go
- **Verification:** TestAdoptWatcherServesDefaultsOnAbsentFile green (watcher non-nil, defaults served, directory created, file still absent); TestLoadConfigAdoptAbsent re-pins that the FILE is never created
- **Committed in:** 1afaac0 (Task 3 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** the fix is required for the plan's adopt+watch contract to hold on the machines it targets (the owner's live machine has no config directory); no scope creep.

## Issues Encountered

- Pre-existing flake (07-02, NOT caused by this plan): `TestWatch_BrokenBlocklistPatternKeepsLastGood` fails occasionally (~1/15 package runs) — `Watcher.reparse` stores the new snapshot before clearing `lastErr`, and the test reads `LastError` immediately after observing the snapshot update, so it can see the stale rejection. Writer code is not in that path. Logged to `deferred-items.md` per the scope boundary; affected runs were re-run green (same policy as the declared TestRun_OnConnHookCalledOnce flake).
- The `check tdd-red-evidence` verb parses node --test TAP, not `go test -v` output; the RED evidence record therefore carries a mechanical TAP projection of the real run (same tests, same outcomes; raw log at /tmp/red-07-03-task2.log during the session). Verdict `RED_EVIDENCE_OK`.

## User Setup Required

None - no external service configuration required.

## TDD Gate Compliance

| Task | RED commit | GREEN commit | Evidence |
|------|-----------|--------------|----------|
| Task 1 (tracer) | 395382a test(07-03) | 852033d feat(07-03) | 4 target tests failing on stub refusal (assertion RED) |
| Task 2 (tdd) | cae12a7 test(07-03) | 0fa029c feat(07-03) | `check tdd-red-evidence` → RED_EVIDENCE_OK (target_test_failed, 6/6) |
| Task 3 (auto) | d1c5f52 test(07-03) | 1afaac0 feat(07-03) | 4 adopt tests RED on assertions; explicit-priority green-by-design (continuity pin) |

## Next Phase Readiness

- Plan 07-05 (menu v2) has its complete persist surface: SetAutocorrectEnabled/SetSoundEnabled for the two toggles, EnsureDocument before the editor launch, DefaultPath as the shared resolver; the echo-reload is expected and harmless (research Q3b) — the toggle applies its value synchronously after the write.
- Plan 07-06 (menu e2e) inherits the restart-survival oracle: the daemon now reads the default path without -config, so a persisted toggle survives restarts by mechanism.
- `docs/CONFIG.md` persist-semantics note (last-writer-wins window, Pitfall 5) is assigned to the docs-carrying plans per the plan's assumption log; not re-done here.

---
*Phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii*
*Completed: 2026-10-04*

## Self-Check: PASSED

- All 4 key files exist on disk (writer.go, writer_test.go created; main.go, main_test.go modified).
- All 6 task commits present in git log (395382a, 852033d, cae12a7, 0fa029c, d1c5f52, 1afaac0); measured plan commits: 6 from ledger base 79bc287.
- Plan-level verification re-run green: `go test ./internal/config/ ./cmd/goswitchd/ -race -count=1` ok; `mise run ci` ok (build+vet+lint 0 issues+test -race); tracer verify gates TRACER-OK, PASS count 8 in [5-9]; adopt count exactly 4 `^--- PASS`.
- All writer value assertions read back through strict config.Load (Pitfall 6); writer.go carries zero log statements (prohibition 3 by construction).
