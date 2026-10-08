---
phase: 01-adr-paket-i-karkas-ibus-dvizhka
plan: "01"
subsystem: infra
tags: [ibus, dbus, godbus, go, mise, golangci-lint, slog, input-method]

requires:
  - phase: 00-skeleton
    provides: empty repo skeleton with .gitkeep placeholders
provides:
  - toolchain skeleton — mise.toml (single source of tool versions) and strict .golangci.yml (D-09/D-10)
  - Go module github.com/Djarvur/goswitch (go 1.23, godbus v5.2.2 only)
  - engine/ IBus wire adapter — programmatic RegisterComponent, factory, engine object on 3 interfaces with recover shim, address discovery with session-type suffix filter, reconnect loop
  - internal/logging slog JSON setup with -debug-gated key trace
  - cmd/goswitchd binary (thin main + run(ctx, cfg) error)
  - TDD corpus pinning wire field order, observer-false, decode, recover, log gating, address precedence
affects: [01-02 layouts, 01-03 e2e-skeleton, 01-04 d01-experiment, 01-05 ci-readme, phase-2-correction]

actuals:
  tokens: 14457   # chars/4 over the realized diff 3db4b5a..HEAD (estimate was 48000)
  tasks: 2
  commits: 3       # MEASURED: git rev-list --count 3db4b5a..HEAD (ledger base)

tech-stack:
  added:
    - github.com/godbus/dbus/v5 v5.2.2 (only external dependency)
    - mise 2025.9.4 (user-local install, official installer, no root)
    - go 1.23.12 (per-project via mise [tools])
    - golangci-lint 2.13.2 (per-project via mise [tools])
  patterns:
    - wire structs marshal by declaration order — field order is the contract, pinned by reflect tests
    - recover shim via named returns + defer recoverHandler on every D-Bus handler
    - address discovery never cached; re-discovered per connection cycle with session-type suffix filter
    - observer mode: ProcessKeyEvent returns literal false for everything (transit by construction)
    - key trace only behind -debug (default off); no buffer contents at INFO

key-files:
  created:
    - mise.toml
    - .golangci.yml
    - .gitignore
    - go.mod
    - go.sum
    - cmd/goswitchd/main.go
    - internal/logging/logging.go
    - internal/logging/logging_test.go
    - engine/address.go
    - engine/address_test.go
    - engine/types.go
    - engine/keys.go
    - engine/conn.go
    - engine/factory.go
    - engine/engine.go
    - engine/engine_test.go
    - engine/wire_test.go
  modified: []

key-decisions:
  - "RequestName targets the component name org.freedesktop.IBus.goswitch, not org.freedesktop.IBus — live probe proved the latter is reserved by ibus-daemon ('Can not acquire the service name ... reserved by IBus'); single-instance semantics preserved (PrimaryOwner check, T-01-03)"
  - "Auth set is EXTERNAL-only: godbus v5.2.2 ships no DBUS_COOKIE_SHA1 on Linux (its Unix default is AuthExternal alone) and the IBus socket authenticates via SO_PEERCRED"
  - "EngineDesc wire struct enumerated with 18 fields (magic + attachments + 8 strings + rank + 7 strings) — the research table's '17 fields' plus its trailing '→ Textdomain' entry; verified against /usr/include/ibus-1.0/ibusenginedesc.h and live ListActiveEngines; Layout@9 and Symbol@12 exactly as pinned"
  - "Live-gate verification adapted to build facts: programmatic registrations never appear in `ibus list-engine` (it reads the XML registry only on IBus 1.5.29) — daemon-side ListActiveEngines proves both engines; key assertion pinned on layout-independent keycodes (34/35/48) instead of latin keyval 0x67, because keyvals are computed by the focused client's XKB group and the live session's focused window sat on the RU group"
  - "nonamedreturns disabled in .golangci.yml (justified in-file): the recover shim requires named returns to guarantee zero-value replies after a contained panic — must-have truth INTEG-05 outranks the style linter"
  - "mise 2025.9.4 installed user-local via the official installer (mise.run) — no root (T-01-05)"

patterns-established:
  - "Triple green gate per iteration: mise run ci = build + vet + golangci-lint + go test -race (D-08)"
  - "TDD on pure adapter units: RED committed with gsd check tdd-red-evidence verdict RED_EVIDENCE_OK before any product edit (D-07); live D-Bus handshake proven by the live gate instead"
  - "External test packages (engine_test, logging_test) with vanilla testing, TestF_suffixCamelCase names"

requirements-completed: [INTEG-01, INTEG-02, INTEG-03, INTEG-04, INTEG-05, INST-04]

coverage:
  - id: D1
    description: "Toolchain skeleton: mise.toml (tools + tasks) and strict .golangci.yml, first code commit"
    requirement: INTEG-03
    verification:
      - kind: other
        ref: "mise exec -- golangci-lint config verify (exit 0)"
        status: pass
      - kind: other
        ref: "mise run ci — build + vet + lint + test -race all green"
        status: pass
    human_judgment: false
  - id: D2
    description: "goswitchd registers a component with two engines programmatically on the private IBus bus (no XML, no root) and exits cleanly on SIGTERM"
    requirement: INTEG-01
    verification:
      - kind: e2e
        ref: "live gate: 'component registered' within 10 s; ListActiveEngines contains goswitch-en and goswitch-ru with correct descs (layout us/ru, symbol en/ru); SIGTERM exit code 0"
        status: pass
    human_judgment: false
  - id: D3
    description: "Observer engine: being the active input source it receives ProcessKeyEvent for physically injected keys and returns false for everything"
    requirement: INTEG-02
    verification:
      - kind: unit
        ref: "engine/engine_test.go#TestEngine_ProcessKeyEventReturnsFalse (Shift_R press/release, letters, modifier combos)"
        status: pass
      - kind: e2e
        ref: "live gate: ydotool type ghbdtn -> 12 key records in the JSON log (press+release pairs, keycodes 34/35/48)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Recover shim on every exported D-Bus handler — an injected panic never escapes, safe replies, ERROR-level containment"
    requirement: INTEG-05
    verification:
      - kind: unit
        ref: "engine/engine_test.go#TestEngine_RecoverContainsPanic (all 21 handlers, panic-injecting EventHandler)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Structured JSON logging with level contract and privacy-gated key trace"
    requirement: INST-04
    verification:
      - kind: unit
        ref: "internal/logging/logging_test.go#TestLogging_JSONShape, TestLogging_LevelFiltering, TestLogging_KeyTraceGated"
        status: pass
    human_judgment: false
  - id: D6
    description: "Address discovery precedence contract: env → session-type suffix filter → freshest mtime; named error on empty/missing dir; never cached, re-discovered every reconnect cycle"
    requirement: INTEG-04
    verification:
      - kind: unit
        ref: "engine/address_test.go#TestAddress_EnvPrecedence, TestAddress_PicksFreshest, TestAddress_EmptyDirError (RED then GREEN)"
        status: pass
    human_judgment: false
  - id: D7
    description: "Reconnect/re-register cycle (fresh Discover per Dial, Factory re-export, RegisterComponent, 1-2 s jittered backoff) shipped in code as a day-one element"
    requirement: INTEG-04
    verification: []
    human_judgment: true
    rationale: "The bus-loss cycle is only observable against a live ibus-daemon restart — plan 01-03's ibus-restart e2e case is the designated proof; headless tests cannot exercise a real bus generation."

duration: 35 min
completed: 2026-09-10
status: complete
plan_head_before: 3db4b5a32e957cdf236100b2b3bf2c09be75d020
---

# Phase 1 Plan 01: ADR-пакет и каркас IBus-движка — Tracer Summary

**Walking skeleton: goswitchd registers two engines on the private IBus bus via pure-Go godbus (no XML, no root), observes every keystroke behind a privacy-gated -debug trace, survives handler panics by recover shim, and exits cleanly on SIGTERM — with the mise/golangci toolchain gate green from the first commit.**

## Performance

- **Duration:** 35 min
- **Started:** 2026-09-10T18:22:41Z
- **Completed:** 2026-09-10T18:58:17Z
- **Tasks:** 2 (tracer + TDD corpus)
- **Files modified:** 19 (17 created, engine/.gitkeep removed, .golangci.yml)

## Accomplishments

- Toolchain skeleton (D-09/D-10): mise 2025.9.4 user-local (official installer, no root), mise.toml pinning go 1.23.12 + golangci-lint 2.13.2 with tasks build/test/lint/vet/tidy-diff/ci; strict .golangci.yml (default: all, depguard deny unsafe, generated: strict, goimports local-prefixes) — `mise run ci` green from the first commit (D-08)
- engine/ wire adapter (~700 LOC): programmatic RegisterComponent with wire-faithful Component/EngineDesc structs (order verified live via ListActiveEngines — every field incl. Layout@9/Symbol@12 parsed correctly), factory minting monotonic engine paths, engine object exported on 3 interfaces with recover shim on all 21 D-Bus handlers, address discovery with session-type suffix filter, reconnect loop with 1-2 s jittered backoff
- cmd/goswitchd thin main + run(ctx, cfg) error with -debug flag; internal/logging slog JSON Setup with the level contract (key trace only behind -debug — INST-04/T-01-01)
- TDD corpus (D-07): RED committed with RED_EVIDENCE_OK (suffix filter + missing-dir named error failed on assertions), then GREEN; pins for wire order, observer-false, decode, keyval constants, recover, log gating — all green under -race
- Live gate on this machine's GNOME Wayland session: component registered → both engines listed daemon-side → focus_in → 12 key events for injected ghbdtn → engine restored → SIGTERM exit 0

## Task Commits

1. **Task 1: Tracer — goswitchd registers on IBus and sees keys** — `7654296` (feat)
2. **Task 2: TDD corpus RED** — `d7387e6` (test)
3. **Task 2: TDD corpus GREEN** — `f610436` (feat)

**Plan metadata:** (see final docs commit)

_Note: Task 2 is a TDD task with separate test/feat commits; no REFACTOR commit — no behavior-neutral cleanup was warranted._

## Files Created/Modified

- `mise.toml` — tool pins + mise tasks (the only task runner; no Makefile, D-09)
- `.golangci.yml` — strict v2 lint config adapted from the owner reference
- `go.mod` / `go.sum` — module github.com/Djarvur/goswitch, go 1.23, godbus v5.2.2 only
- `.gitignore` — skill baseline minus dist/ (tracked packaging sources) plus .mise.local.toml
- `engine/types.go` — wire structs + goswitch constructors; field order = wire contract
- `engine/keys.go` — verified keyvals, state masks, capability bits
- `engine/address.go` — Discover() precedence: env → suffix-filtered freshest; ErrNoAddress named error
- `engine/conn.go` — Run/serve cycle: Dial/Auth/Hello/RequestName guard/Factory export/RegisterComponent/bus-loss wait/backoff
- `engine/factory.go` — CreateEngine minting /org/freedesktop/IBus/Engine/goswitch/<n>
- `engine/engine.go` — engine object, EventHandler seam, recover shim, CommitText emitter
- `engine/{address,wire,engine}_test.go`, `internal/logging/logging_test.go` — the corpus

## Decisions Made

See key-decisions in frontmatter. Highlights: RequestName retargeted to the component name after the live reserved-name probe; EXTERNAL-only auth (godbus has no Linux CookieSha1); live-gate assertions adapted to ListActiveEngines + layout-independent keycodes (both documented as deviations below).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] RequestName target is reserved by the daemon**
- **Found during:** Task 1 (pre-code live probe, per the plan's own flagged assumption)
- **Issue:** The plan pinned `RequestName("org.freedesktop.IBus", NameFlagReplaceExisting)`; a live busctl probe returned "Can not acquire the service name 'org.freedesktop.IBus', it is reserved by IBus" — the literal implementation could never connect.
- **Fix:** Request `cfg.Component.ComponentName` ("org.freedesktop.IBus.goswitch") with the same PrimaryOwner check — the owner's Python prototype's proven pattern (it requested org.freedesktop.IBus.TestShift). Single-instance semantics intact (T-01-03).
- **Files modified:** engine/conn.go
- **Verification:** live gate — daemon becomes PrimaryOwner, registers, serves; second-instance guard path unit-level simple.
- **Committed in:** 7654296

**2. [Rule 3 - Blocking] godbus v5.2.2 API differs from the plan's sketched calls**
- **Found during:** Task 1 (first build)
- **Issue:** `Hello()` returns only error; `CallWithContext` returns only *Call; `Signal()` registers a caller-supplied channel (closed by godbus on connection loss — which is exactly the bus-loss detector); `AuthExternal` takes a string uid; **no AuthCookieSha1 exists on Linux** in v5.2.2 (Unix default auth set is EXTERNAL alone).
- **Fix:** Rewrote conn.go against the actual API; auth set = AuthExternal(decimal uid); bus loss = close of the registered signal channel.
- **Files modified:** engine/conn.go
- **Verification:** build/vet/lint green; live reconnect-signal semantics from godbus source (default_handler.go Terminate closes registered channels).
- **Committed in:** 7654296

**3. [Rule 1 - Verification] Live-gate assertions adapted to build/dispatch facts**
- **Found during:** Task 1 verify (live gate)
- **Issue:** (a) `ibus list-engine | grep goswitch` can never pass — the CLI reads only the XML registry on IBus 1.5.29-rc2, while programmatic registrations live daemon-side. (b) The keyval-0x67 assertion assumes the focused client translates through a latin XKB group; the live session's focused window sat on the RU group (injected ghbdtn arrived as Cyrillic keyvals — correct behavior for an IME), and GNOME's engine-activation path deliberately does not re-apply XKB (the D-01 finding).
- **Fix:** (a) Assert via daemon-side `ListActiveEngines` (both engines, correct descs). (b) Assert on layout-independent keycodes (34/35/48 = G/H/B) for the injected burst. Both preserve the plan's intent (INTEG-01 registration proof; M1-in-miniature key observation).
- **Files modified:** none (verification procedure only)
- **Verification:** LIVE-GATE-PASS (REG=0 FOC=0 FAIL=0 DEXIT=0 KEYS=12), re-run green on the post-GREEN binary.
- **Committed in:** n/a (procedure)

**4. [Rule 2 - Missing critical / lint coexistence] Strict-lint config deltas**
- **Found during:** Task 1 (first lint iteration, D-10 gate)
- **Issue:** default:all fired on code the plan's must_haves require: nonamedreturns vs the recover shim's named returns (INTEG-05); hugeParam vs the plan-pinned Run(ctx, cfg Config) signature; gosec G404 (math/rand/v2 backoff jitter — timing, not secrets); gosec G703 (paths enumerated from the fixed IBus config dir); paralleltest in tests using t.Setenv/global slog.
- **Fix:** nonamedreturns disabled with one-line justification; gocritic hugeParam sizeThreshold 512 (commented); two targeted gosec exclusions and a paralleltest test-path exclusion, each with in-file justification. All other default:all linters pass on the product code.
- **Files modified:** .golangci.yml
- **Verification:** mise run lint — 0 issues.
- **Committed in:** 7654296 (exclusion for tests in d7387e6)

**5. [Rule 1 - Documentation] EngineDesc field count is 18, not 17**
- **Found during:** Task 1 (types.go authoring)
- **Issue:** The plan/research say "17 fields" but the research's own indexed list ends "16 Version → then Textdomain" and the signature line (s, a{sv}, s×8, u, s×7) sums to 18.
- **Fix:** Enumerated 18 fields; verified against /usr/include/ibus-1.0/ibusenginedesc.h property order and live ListActiveEngines output (every field parsed correctly). Layout@9/Symbol@12 exactly as the acceptance criteria pin. The wire test pins all 18 names.
- **Files modified:** engine/types.go (vs the plan's stated count only)
- **Verification:** TestWire_FieldOrder green; live ListActiveEngines desc dump.
- **Committed in:** 7654296

---

**Total deviations:** 5 auto-fixed (2 bug, 1 blocking, 1 missing-critical/lint, 1 documentation) + 1 verification-procedure adaptation
**Impact on plan:** All fixes preserve plan intent; the two wire-level fixes (RequestName target, field count) were anticipated by the plan's own flagged assumptions and were settled by live probes exactly as the plan instructed ("сверка при отклонении").

## Issues Encountered

- The live gate's key-delivery leg is sensitive to whatever window the live session has focused (twice the injected burst produced zero events — the focused surface was not routing through the engine; once an Escape dismissed the shell overview and keys flowed). This is precisely the gap plan 01-03's TEST-03 (AT-SPI grabFocus + gnome-text-editor target) closes; recorded here as an input for that plan.
- A Super+Space probe to flip the session's XKB group opened the GNOME overview (dismissed with Escape; gsettings `current` unchanged throughout — no persistent desktop state mutation). Alt+Shift toggles did not fire through the injection path. Layout control stays with the D-01 experiment (plan 01-04).

## Authentication Gates

None — no authenticated external services involved.

## User Setup Required

None — mise installed user-locally by the executor (official installer, T-01-05); no external service configuration.

## Known Stubs

None. The CommitText emitter is functional (unused until Phase 2 by design); no TODO/FIXME/placeholder markers in product code.

## Next Phase Readiness

- Ready for 01-02 (layouts generator): module builds under the strict lint gate; generated-file path is pre-wired via exclusions.generated: strict + DO-NOT-EDIT marker convention.
- Ready for 01-03 (e2e skeleton): the daemon's JSON log is the assertion surface and the lifecycle sequence connected → component registered → engine created → focus_in is live-proven; the focus-control gap observed above is that plan's core work.
- Ready for 01-04 (D-01): both engines registered from day one as required; the XKB-group observation (engine activation does not re-apply layout; keyvals follow the focused client's group) is direct experimental input.
- EngineDesc/Layout fields verified against mutter's expectations remain to be exercised through GNOME's own input-source path (01-03+).

---
*Phase: 01-adr-paket-i-karkas-ibus-dvizhka*
*Completed: 2026-09-10*

## Self-Check: PASSED

All 17 key files exist on disk; all four plan commits (7654296, d7387e6, f610436, 1d26021) found in history. Final gates re-verified post-GREEN: mise run ci green, live gate green (KEYS=12, DEXIT=0).
