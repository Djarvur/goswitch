---
phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
plan: 02
subsystem: config
tags: [yaml-schema, blocklist, regex-re2, validation, sound, autocorrect, strict-decode, tdd]

# Dependency graph
requires:
  - phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii/01
    provides: "SPEC §11 third revision + ADR-007 amendment — the blocklist+sound contract this plan implements"
  - phase: 06-...
    provides: "Autocorrect config section (06-04), autocorrect actor boundary + silence corpus (06-06/06-07)"
provides:
  - "Autocorrect.AppsBlocklist (yaml apps_blocklist) — the white-list key apps is GONE; strict decode refuses old documents whole (D-33)"
  - "Load-time blocklist validation: unconditional 64 ceiling, empty/blank-pattern refusal, regexp.Compile gate — every error names field + index"
  - "ACTIVE condition: enabled alone is active; thresholds mandatory when enabled (zero-threshold enabled refuses loudly)"
  - "Sound section (owner decision): Enabled *bool (nil = ON), AutocorrectEvent, EffectiveEnabled/EffectiveAutocorrectEvent, DefaultSoundFlipEvent=bell / DefaultSoundAutocorrectEvent=message"
  - "Consumer blocklist polarity: a known identity matching a compiled pattern refuses with slug app-blocked; unknown identity still conservatively silent (07-04 inverts)"
  - "Compiled-pattern cache in the actor (rebuild on list change, parse-cache form), refreshed from applySnapshot AND SetOptions"
  - "docs/CONFIG.md: six sections, migration note, ACTIVE/regex semantics, sound rows"
affects: [07-03-persist-writer, 07-04-gate-rework, 07-05-menu, 07-06-e2e-repins, 07-08-sound-engine]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
actuals:
  tokens: 15667    # chars/4 over the realized diff (62671 chars, 7 files, +767/-196)
  tasks: 3
  commits: 5

# Tech tracking
tech-stack:
  added: []        # zero new dependencies — regexp is stdlib (Package Legitimacy: no new packages)
  patterns:
    - "Pointer-bool default-ON accessor: Sound.Enabled *bool because Load overlays no defaults — nil reads ON (EffectiveEnabled)"
    - "Sentinel-per-validation-class with field+index wrapping (errAutocorrectBlocklistRegex/Empty/OverCeil, D-33 discipline)"
    - "Compiled-pattern parse-cache keyed by the raw list (macrLettersName form), one derivation point shared by applySnapshot and SetOptions"

key-files:
  created: []
  modified:
    - internal/config/config.go
    - internal/config/config_test.go
    - internal/config/load_test.go
    - internal/config/watch_test.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - docs/CONFIG.md

key-decisions:
  - "RED-stub precedent (06-04) reused: the RED commit carries the schema shape only (AppsBlocklist alongside Apps + full Sound shape) so the final-form corpus compiles and fails on assertions, never on compile errors"
  - "Pattern cache derivation is shared: applySnapshot (config path) AND SetOptions (no-config/test surface) both call refreshACBlocklist — a cache built only in applySnapshot left the SetOptions surface matching nothing (caught by the corpus)"
  - "Slug registry cleanup moved up for two slugs: acReasonNoApps/acReasonAppNotListed deleted at 07-02 (their branches end here; unused constants fail the strict lint) — the rest of the 07-04 vocabulary work (app-unknown inversion, app-changed removal) untouched"
  - "requirements MACR-ACL left open per the shared-ID gate: 07-03/04/06/07 also declare it — the LAST declaring plan marks it complete"

patterns-established:
  - "Blocklist polarity at the boundary: known identity + any regex match ⇒ abstain app-blocked; empty blocklist forbids nothing"
  - "Load-time regex compile as the single pre-runtime visibility point for broken patterns; hot reload inherits via D-32 with zero watcher code"

requirements-completed: ["MACR-ACL (blocklist-ревизия)"]  # copied verbatim; MARKING deferred to the last declaring plan (shared-ID gate: 07-03/04/06/07 pending)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Schema: apps_blocklist replaces apps; strict decode refuses the legacy key whole; compile/empty/ceiling/order validation at Load with field+index errors"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: unit
        ref: "tests/internal/config/load_test.go#TestLoad_StrictDecodeRejectsLegacyAppsKey"
        status: pass
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestValidate_BlocklistRegexCompile"
        status: pass
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestValidate_BlocklistEmptyPattern"
        status: pass
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestValidate_BlocklistCeilUnconditional"
        status: pass
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestValidate_BlocklistOrderIrrelevant"
        status: pass
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestDefaults_AutocorrectBlocklistNil"
        status: pass
    human_judgment: false
  - id: D2
    description: "ACTIVE condition: enabled requires thresholds (per-field refusals); enabled+empty blocklist+thresholds valid; disabled+zeros valid; broken blocklist pattern holds last-good through hot reload"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestValidate_EnabledRequiresThresholds"
        status: pass
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestValidate_EnabledEmptyBlocklistWithThresholdsValid"
        status: pass
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestValidate_DisabledZeroThresholdsValid"
        status: pass
      - kind: integration
        ref: "tests/internal/config/watch_test.go#TestWatch_BrokenBlocklistPatternKeepsLastGood"
        status: pass
    human_judgment: false
  - id: D3
    description: "Sound section: absent section means ON (pointer-bool), explicit off and custom event decode, unknown sound key rejects the file, Defaults carry ON + bell/message constants"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: unit
        ref: "tests/internal/config/load_test.go#TestSound_AbsentSectionMeansOn"
        status: pass
      - kind: unit
        ref: "tests/internal/config/load_test.go#TestSound_ExplicitOffAndCustomEvent"
        status: pass
      - kind: unit
        ref: "tests/internal/config/load_test.go#TestSound_StrictDecodeUnknownKey"
        status: pass
      - kind: unit
        ref: "tests/internal/config/config_test.go#TestDefaults_SoundOn"
        status: pass
    human_judgment: false
  - id: D4
    description: "Consumer: blocklist polarity (known app + matching pattern ⇒ app-blocked abstain), compiled-pattern cache rebuilt on list change, fire paths green with a non-matching pattern"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestAutoCorrect_SilenceMatrix/blocklist_match_forbids_a_known_app"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestApplySnapshot_AutocorrectFold"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_AutocorrectBoundaryArms"
        status: pass
    human_judgment: false
  - id: D5
    description: "Conservative intermediate: an UNKNOWN identity still abstains (app-unknown) at this plan's output — the inversion to pass-through is 07-04's; no unknown⇒fired cell appears"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestAutoCorrect_SilenceMatrix/missing_identity_source_is_unknown"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestAutoCorrect_FailClosedNoFailOpen"
        status: pass
    human_judgment: false
  - id: D6
    description: "docs/CONFIG.md: six sections, apps_blocklist row byte-matched to the schema, D-53 migration note, ACTIVE + regex semantics, sound rows (canberra/paplay, WARN-only), tray-toggle persist mention"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: other
        ref: "grep gates: apps_blocklist / sound.enabled / sound.autocorrect_event / canberra / автокоррекц present; key names == yaml tags"
        status: pass
    human_judgment: true
    rationale: "Documentation reader-adequacy (does the migration note actually warn the right user, is the regex semantics paragraph clear) is owner judgment; only the key-name agreement is machine-checked"

# Metrics
duration: 28min
completed: 2026-10-05
status: complete
---

# Phase 7 Plan 2: Config schema — apps_blocklist + ACTIVE condition + sound section Summary

**Config schema D-53 revision landed: `autocorrect.apps_blocklist` (RE2, validated at Load) replaces the white-list key with a loud strict-decode migration, `enabled` alone is the ACTIVE state demanding thresholds, and the new `sound` section ships default-ON via a pointer-bool; the actor refuses on a blocklist match (`app-blocked`) with a compiled-pattern cache while unknown identities stay conservatively silent until 07-04.**

## Performance

- **Duration:** 28 min
- **Started:** 2026-10-04T20:50:49Z
- **Completed:** 2026-10-05T00:38:35Z (5 commits, 3 tasks)
- **Tasks:** 3 (TDD strict: 2× RED→GREEN + 1 auto)
- **Files modified:** 7

## RED→GREEN evidence (TDD gate)

| Task | RED commit (target failure) | GREEN commit |
|------|------------------------------|--------------|
| 1. Schema | `8985e50` — RED_EVIDENCE_OK on `TestLoad_StrictDecodeRejectsLegacyAppsKey` (old key still decoded; 4 target failures: legacy key, compile gate, empty pattern, ceiling) | `a7dac34` |
| 2. ACTIVE condition | `e7a0cb0` — RED_EVIDENCE_OK on `TestValidate_EnabledRequiresThresholds` (old enabled+non-empty-list condition passed zero thresholds; + the 06-04 dormant cell re-pinned) | `6d0cc5d` |
| 3. Consumer + docs | auto (no RED gate) | `92380f6` |

The -run RUN-count gate for task 2 returned exactly 5 (`TestValidate_Enabled*` parents+subtests + the watch pin).

## Accomplishments

- Schema: `Autocorrect.AppsBlocklist []string \`yaml:"apps_blocklist"\`` — the legacy `apps` key is gone from the struct (exactly one app-list yaml tag remains, MACR's own); strict decode refuses old documents whole (D-33, no half-support)
- Validation at Load: unconditional 64 ceiling (`maxAutocorrectBlocklist`, dormant included — T-07-02-02), empty/blank-pattern sentinel (an empty pattern matches everything — T-07-02-03), `regexp.Compile` gate with field+index errors (T-07-02-01); order carries no meaning
- ACTIVE condition: `if !a.Enabled { return nil }` — thresholds mandatory whenever enabled (T-07-02-05); enabled + empty blocklist + thresholds is valid AND active; Defaults stay off (D-54)
- Sound section (owner decision «Звуки при переключении»): `Enabled *bool` (nil = ON — Load overlays no defaults), `AutocorrectEvent` (empty = default), `EffectiveEnabled()`/`EffectiveAutocorrectEvent()`, `DefaultSoundFlipEvent="bell"` / `DefaultSoundAutocorrectEvent="message"` (both verified in the Yaru theme); no section validation — strict decode covers unknown keys automatically
- Consumer: blocklist polarity at `autoCorrectBoundary` — a known identity matching any compiled pattern abstains with the new `app-blocked` slug; empty blocklist gates nothing; the unknown-identity branch stays conservatively silent (07-04 inverts); compiled-pattern cache rebuilt only on list change and shared by `applySnapshot` and `SetOptions`
- docs/CONFIG.md: six sections, apps_blocklist table row (substring/anchor/case/order/≤64), migration note (rename; white lists are never mechanically transferred — opposite intents), ACTIVE semantics, hot-reload bullet now names broken regex patterns, sound rows with canberra-gtk-play/paplay WARN-only playback, tray-toggle persist mention, examples byte-matched to the schema
- `mise run ci` green on the final tree; `load.go`/`watch.go` untouched (zero diff — D-32/D-33 propagate); no live e2e runs (templates transitively stale until 07-06, per the plan's own assumption)

## Task Commits

1. **Task 1 RED: failing blocklist/sound corpus** — `8985e50` (test)
2. **Task 1 GREEN: schema — apps_blocklist, validation, sound** — `a7dac34` (feat; carries the precondition-sanctioned mechanical consumer rename to keep `go build ./...` green in one pass)
3. **Task 2 RED: ACTIVE-condition corpus + watch pin** — `e7a0cb0` (test)
4. **Task 2 GREEN: ACTIVE condition** — `6d0cc5d` (feat)
5. **Task 3: consumer polarity + cache + docs/CONFIG.md** — `92380f6` (feat)

## Files Created/Modified

- `internal/config/config.go` — blocklist field + sentinels + ceiling + compile/empty gates; ACTIVE condition; Sound type with pointer-bool accessors and default-event constants; Defaults updated
- `internal/config/config_test.go` — blocklist validation corpus, sound defaults pin, ACTIVE-condition corpus, 06-04 dormant cell re-pinned
- `internal/config/load_test.go` — legacy-key strict rejection, sound decode corpus, decode-default assertions on the new field
- `internal/config/watch_test.go` — corpus key swap; `TestWatch_BrokenBlocklistPatternKeepsLastGood` (last-good through a broken pattern)
- `internal/session/actor.go` — Options field rename, fold, compiled-pattern cache (+shared refresh), boundary polarity with `app-blocked`, slug registry update
- `internal/session/actor_test.go` — corpus re-pinned: non-matching fire pattern, blocklist-match cell (substring pin), fold re-pin, untouched unknown-silence cells
- `docs/CONFIG.md` — six sections, migration note, ACTIVE/regex/sound documentation

## Decisions Made

- RED commits carry the schema SHAPE only (the 06-04 RED-stub precedent) so the final-form corpus compiles and fails on assertions; the interim both-keys state exists only inside the RED commit — the plan output carries exactly one app-list key
- The compiled-pattern cache has ONE derivation point (`refreshACBlocklist`) called from both `applySnapshot` and `SetOptions`: a cache rebuilt only in applySnapshot left the SetOptions/no-config surface with an empty cache (the corpus caught it — the blocked cell abstained 0 times)
- `acReasonNoApps`/`acReasonAppNotListed` deleted from the slug registry at this plan (their branches end here; unused constants fail the strict lint) — the remaining 07-04 vocabulary work is untouched
- MACR-ACL is NOT marked complete in REQUIREMENTS.md: plans 07-03/04/06/07 declare the same ID (shared-ID gate #2388) — the last declaring plan marks it

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] watch_test.go corpus renamed inside the Task 1 GREEN commit**
- **Found during:** Task 1 GREEN
- **Issue:** `autocorrectDocYAML` carried the removed `apps:` key — the schema rename invalidates it at the same commit (strict decode would fail the watch corpus)
- **Fix:** key swap in the watch corpus template; the file sits in Task 2's list but the rename part belongs to the schema commit
- **Files modified:** internal/config/watch_test.go
- **Verification:** package tests green at the GREEN commit
- **Committed in:** a7dac34

**2. [Rule 3 - Blocking] Slug constants acReasonNoApps/acReasonAppNotListed removed at Task 3, not 07-04**
- **Found during:** Task 3
- **Issue:** the plan schedules "removal of the no-apps/not-listed slug constants" for 07-04, but Task 3's polarity replaces both branches — unused constants fail the strict lint (unused)
- **Fix:** deleted the two constants with their branches; the 07-04 work (app-unknown inversion, app-changed removal, confirm rework) remains intact
- **Files modified:** internal/session/actor.go
- **Verification:** lint green; no test referenced the literals outside the replaced cell
- **Committed in:** 92380f6

**3. [Rule 3 - Blocking] Pattern cache refreshed from SetOptions as well as applySnapshot**
- **Found during:** Task 3
- **Issue:** the plan pins the cache rebuild in applySnapshot (parse-cache form); SetOptions — the no-config/test surface — writes opts directly and would leave the cache empty, so a blocklist armed via SetOptions matched nothing (corpus failure: abstained 0)
- **Fix:** one `refreshACBlocklist` derivation point called from both surfaces (change-detect + compile preserved)
- **Files modified:** internal/session/actor.go
- **Verification:** blocked cell and fold test green; full session corpus green
- **Committed in:** 92380f6

---

**Total deviations:** 3 auto-fixed (3× Rule 3 blocking, all mechanical necessities of the rename/polarity)
**Impact on plan:** All three are inside the plan's own files and sanctioned directions (precondition explicitly blends tasks 1–3 into one green working pass). No scope creep.

## Issues Encountered

- `TestRun_OnConnHookCalledOnce` (internal/ctlsvc) flaked twice during full-suite runs ("OnConn called 0 times, want exactly 1") — pre-existing timing flake under parallel load, unrelated to this plan's diff (config/session only; passes 4/5 in isolation and on package re-runs). Logged to the phase deferred-items ledger, not fixed here (scope boundary).
- The requirements.ready-ids verb split the space-containing ID into two words (cosmetic); the shared-ID decision was applied by reading the sibling plans directly.

## TDD Gate Compliance

RED and GREEN commits present in order for both TDD tasks (`test(07-02)` 8985e50 → `feat(07-02)` a7dac34; `test(07-02)` e7a0cb0 → `feat(07-02)` 6d0cc5d); both RED runs verified via `gsd_run check tdd-red-evidence` with verdict RED_EVIDENCE_OK (target test failed on the planned assertion; green-by-design pins documented in the commits). No REFACTOR-only commits were needed (REFACTOR happened within the GREEN passes).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Ready for 07-03 (persist writer + adopt): the schema this writer must produce is final — `apps_blocklist` + `sound`, and the ACTIVE condition requires the writer to emit explicit thresholds with `enabled: true`
- 07-04 (gate rework) owns the unknown-identity inversion, the app-changed removal and the confirm rework — the boundary's unknown branch and the acPayload comment here mark exactly that seam
- Transitively stale until 07-06: live e2e templates (`case_autocorrect.go`, `matrixACConfigYAML`) still carry the old key — live runs before 07-06 would be refused by Load (unit CI stays green)

---
*Phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii*
*Completed: 2026-10-05*

## Self-Check: PASSED

- All 7 modified key-files exist on disk; SUMMARY.md exists.
- All 6 commits verified in git log (8985e50, a7dac34, e7a0cb0, 6d0cc5d, 92380f6, 22067ad).
- Commits measured from the plan ledger: `git rev-list --count b28862a..HEAD` = 6 (5 task commits + 1 SUMMARY commit), matching the Task Commits section.
- Plan verification re-run on the committed tree: `mise run ci` exit 0; grep gates (SCHEMA-OK / EMPTY-PAT-OK / SOUND-SCHEMA-OK / CONSUMER-DOC-OK / SOUND-DOC-OK) all pass; `git diff` for load.go/watch.go is empty.
