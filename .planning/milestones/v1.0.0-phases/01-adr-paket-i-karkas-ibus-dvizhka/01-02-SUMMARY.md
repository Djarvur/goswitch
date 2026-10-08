---
phase: 01-adr-paket-i-karkas-ibus-dvizhka
plan: "02"
subsystem: core
tags: [xkb, go:generate, codegen, golden-tests, fsm, tdd, cyrillic, key-position, headless]

requires:
  - phase: 01-adr-paket-i-karkas-ibus-dvizhka (plan 01)
    provides: Go module github.com/Djarvur/goswitch, mise triple-green gate (D-08), strict .golangci.yml with exclusions.generated: strict
provides:
  - layouts package — committed generated ENToRU/RUToEN maps (94 entries each: 47 us(basic) key positions × 2 shift levels, position join with asymmetric pairs &↔?, @↔", #↔№, $↔;, ^↔:, |↔/)
  - layouts/generator dev-side codegen CLI — parses symbols/us default (basic) and symbols/ru default (=winkeys on Ubuntu 24.04) with recursive same-file include expansion (ru(common)), keypad kpdl(comma) skipped, keysym resolution via ibuskeysyms.h defines + embedded Cyrillic name→rune table
  - internal/hotkey — pure Right Shift tap FSM (D-04/D-05 executable spec): NewFSM(window), Feed(ev, now), Event kinds KeyPress/KeyRelease/TimerExpired/Reset, Action Single/Double/Triple, DefaultWindow=300ms, MaxTaps=3
  - 11-case deterministic FSM corpus + 3-function golden corpus, all green under -race with zero D-Bus and zero wall-clock sleeps (TEST-01 headless part)
affects: [01-04 (ADR-002 cites the FSM corpus as the executable spec), phase-2-correction (layouts tables are the correction data), 01-03 e2e]

actuals:
  tokens: 10279   # chars/4 over the realized diff c160d15..HEAD (estimate was 33000)
  tasks: 2
  commits: 4      # MEASURED: git rev-list --count c160d15..HEAD (ledger base)

tech-stack:
  added: []   # generator and FSM are stdlib-only; no new dependencies
  patterns:
    - //go:generate directive lives inside the generated file; determinism is proven dev-side by `go generate ./layouts && git diff --exit-code` — CI never touches xkb inputs
    - strict-lint-clean codegen without a single nolint: static sentinel errors wrapped with %w (err113), lookup tables as functions (gochecknoglobals), blank lines before non-leading continues (nlreturn)
    - pure FSM with injected time — Feed(ev, now time.Duration); the adapter owns time.AfterFunc and re-enters TimerExpired; TimerExpired honors the last tap's deadline so stale timers are no-ops

key-files:
  created:
    - layouts/generator/main.go
    - layouts/tables.go
    - layouts/tables_test.go
    - internal/hotkey/fsm.go
    - internal/hotkey/fsm_test.go
    - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-02-red-evidence-layouts.json
    - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-02-red-evidence-fsm.json
  modified: []

key-decisions:
  - "Plan behavior bullet's per-char pairs ('h'→'и','b'→'б','d'→'д') contradicted its own string-level pin 'Ghbdtn'→'Привет'; the xkb files and the SPEC examples settle the position-true pairs g→п, h→р, b→и, d→в, t→е, n→т — corpus pins those, all string examples green"
  - "TimerExpired honors the last tap's deadline (a stale AfterFunc from an earlier tap is a no-op) — required by the InterveningKeyCancels pin (no action at the canceled series' window) and it makes the adapter's re-arm race-safe"
  - "The embedded Cyrillic table is keyed by keysym NAME (~67 entries) with values resolved live from ibuskeysyms.h defines: Latin-1 direct for 0x20–0xff, table for the rest; unknown names fail generation loudly, never runtime"
  - "KeyvalShiftR (0xffe2) is defined inside internal/hotkey, not imported from engine — keeps the pure package free of the D-Bus adapter; the future dependency direction is engine → hotkey"
  - "Generator passes the strict lint gate with zero exclusions per D-10's mandate — restructured with sentinel errors and function-form lookup tables instead of any //nolint"

patterns-established:
  - "TDD with committed stubs: the generated-file stub and the FSM type-surface stub make RED a genuine assertion failure (RED_EVIDENCE_OK both tasks), never a compile error (D-07)"
  - "Determinism proof as the only generation check: no Go test executes the generator (package main cannot be imported; CI must not depend on xkb) — `go generate && git diff --exit-code` is the dev-side gate"

requirements-completed: [CORR-08, TEST-01]  # TEST-01 is the headless part only; REQUIREMENTS.md marking gated on sibling plans 01-03/01-05 (shared ID)

coverage:
  - id: D1
    description: "Generated, committed EN↔RU layout tables from xkb key positions with golden corpus (CORR-08)"
    requirement: CORR-08
    verification:
      - kind: unit
        ref: "layouts/tables_test.go#TestGolden_SpecExamples, TestGolden_Punctuation, TestGolden_TableSize (three registers, 16 punctuation pairs both directions, size pins)"
        status: pass
      - kind: other
        ref: "mise exec -- go generate ./layouts && git diff --exit-code -- layouts/tables.go → DETERMINISTIC"
        status: pass
      - kind: other
        ref: "head -1 layouts/tables.go | grep -q 'Code generated' && mise run ci (build+vet+lint+test -race) — all green"
        status: pass
    human_judgment: false
  - id: D2
    description: "Pure Right Shift tap FSM with deterministic 11-case corpus (TEST-01 headless part; D-04/D-05 executable spec for ADR-002)"
    requirement: TEST-01
    verification:
      - kind: unit
        ref: "internal/hotkey/fsm_test.go#TestFSM_SingleAtWindowExpiry, DoubleAndTriple, WindowEdges, ModifierUse, InterveningKeyCancels, FourthTapStaysAtThree, FocusOutDisarms, ReleaseWithoutPress, ShiftGlitch, BurstOfTen, DefaultWindow — green under -race"
        status: pass
      - kind: other
        ref: "grep clean: no goroutines/channels/time.Now/time.Sleep in fsm.go; mise run ci green"
        status: pass
    human_judgment: false

duration: 20 min
completed: 2026-09-10
status: complete
plan_head_before: c160d157bd7956fa0d8872ad03af440717ccdb7e
---

# Phase 1 Plan 02: ADR-пакет и каркас IBus-движка — Layouts + FSM Summary

**Two headless pure packages: xkb-generated ЙЦУКЕН↔QWERTY tables (position join, both shift levels, asymmetric pairs pinned by a golden corpus) and a classic-with-waiting Right Shift tap FSM whose 300 ms decisions fire only at window expiry over injected time.**

## Performance

- **Duration:** 20 min
- **Started:** 2026-09-10T19:00:30Z
- **Completed:** 2026-09-10T19:21:00Z
- **Tasks:** 2 (both type: tdd — RED stub → GREEN, no REFACTOR needed)
- **Files modified:** 8 (7 created, layouts/.gitkeep removed)

## Accomplishments

- layouts/ generated from system xkb (CORR-08 closed): generator resolves us default (basic) and ru default (winkeys via the default directive, ru(common) include expanded, kpdl skipped), joins the 47 alphanumeric key positions at both shift levels, resolves keysyms through ibuskeysyms.h + an embedded Cyrillic table, and emits gofmt-clean sorted maps — 94 entries per direction, byte-for-byte deterministic on regeneration (empty diff)
- Golden corpus: SPEC examples in all three registers both directions (ghbdtn↔привет), all 16 punctuation pairs including every asymmetric one, table-size/inverse pins — green in CI with zero xkb dependency (CI tests the committed file; generation is dev-side)
- internal/hotkey FSM (D-04/D-05 as executable specification): press/release return nothing, the decision (Single/Double/Triple) fires exactly once when the window expires after the last tap; 299/300 ms gaps continue a series, 301 ms starts a new one; modifier use and the ibus#2600 1 ms glitch disqualify a release; intervening keys cancel silently; 4th+ taps stay capped at Triple; Reset/FocusOut disarms everything — 11 test functions green under -race with pure time injection
- Strict TDD (D-07): both tasks committed a genuine assertion-level RED against committed stubs (gsd check tdd-red-evidence → RED_EVIDENCE_OK twice, records committed) before their GREEN commits; every iteration ended with the triple-green mise run ci gate (D-08)

## Task Commits

Each task was committed atomically (TDD: test before feat):

1. **Task 1: Layout tables — RED golden corpus vs empty-map stub** — `c391333` (test)
2. **Task 1: GREEN xkb-driven generator + committed tables** — `7930912` (feat)
3. **Task 2: Tap FSM — RED 11-case corpus vs nil stub** — `17ed1d4` (test)
4. **Task 2: GREEN classic-with-waiting transitions** — `3911e4c` (feat)

**Plan metadata:** (see final docs commit)

## TDD Gate Compliance

| Task | RED commit | RED evidence | GREEN commit | REFACTOR |
|------|-----------|--------------|--------------|----------|
| 1 — layouts | c391333 | RED_EVIDENCE_OK (3/3 TestGolden_* fail on assertions) | 7930912 | not needed — generator restructured during GREEN for the strict gate; no further behavior-neutral cleanup warranted |
| 2 — FSM | 17ed1d4 | RED_EVIDENCE_OK (8 behavioral tests fail on assertions; 3 absence-pins vacuously green vs nil stub) | 3911e4c | not needed — implementation already minimal |

## Files Created/Modified

- `layouts/generator/main.go` — dev-side codegen CLI (stdlib only; strict-lint clean with zero exclusions)
- `layouts/tables.go` — GENERATED, committed: ENToRU/RUToEN map literals; first line the DO-NOT-EDIT marker, second the //go:generate directive (skipped by golangci via generated: strict — no path excludes)
- `layouts/tables_test.go` — golden corpus (package layouts_test)
- `internal/hotkey/fsm.go` — the pure FSM: states idle→taps1..3, decisions at expiry
- `internal/hotkey/fsm_test.go` — the deterministic corpus (package hotkey_test)
- `.planning/.../01-02-red-evidence-{layouts,fsm}.json` — RED evidence records
- `layouts/.gitkeep` — removed (package now has real files)

## Decisions Made

See key-decisions in frontmatter. Highlights: the plan-text pair correction (h→р, b→и, d→в — the bullet's own 'Ghbdtn'→'Привет' pin was internally unsatisfiable as written); the stale-timer deadline guard on TimerExpired; Cyrillic table keyed by keysym name with values from the header; KeyvalShiftR local to hotkey.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Plan-text bug] Behavior bullet's per-character pairs were internally inconsistent**
- **Found during:** Task 1 (RED authoring)
- **Issue:** The behavior bullet pinned 'h'→'и', 'b'→'б', 'd'→'д' while the same bullet (and must_haves, objective, RESEARCH baseline, SPEC) pin 'Ghbdtn'→'Привет' — no function can satisfy both: привет requires h→р, b→и, d→в (verified against the xkb files themselves: h shares AC06 with р, b shares AB05 with и, d shares AC03 with в).
- **Fix:** Corpus pins the position-true pairs (g→п, h→р, b→и, d→в, t→е, n→т) plus both uppercase variants; all string-level examples (ghbdtn/GHBDTN/Ghbdtn ↔ привет/ПРИВЕТ/Привет) green.
- **Files modified:** layouts/tables_test.go
- **Verification:** TestGolden_SpecExamples green; the generated table independently agrees (position join, not hand-rolled).
- **Committed in:** c391333

**2. [Rule 3 - Blocking] Generator tripped the strict lint gate (default: all)**
- **Found during:** Task 1 (GREEN, first mise run ci)
- **Issue:** err113 (10 dynamic fmt.Errorf without %w), gochecknoglobals (keyOrder/cyrillicRunes lookup vars), nlreturn (4 continues) — the plan mandates the generator pass strict lint with no exclusions.
- **Fix:** Static sentinel errors wrapped with %w; lookup tables converted to functions called once in run(); blank lines before non-leading continues. Zero //nolint, zero config changes.
- **Files modified:** layouts/generator/main.go
- **Verification:** golangci-lint run → 0 issues; full mise run ci green; regenerated tables byte-identical.
- **Committed in:** 7930912

---

**Total deviations:** 2 auto-fixed (1 plan-text bug, 1 blocking/lint)
**Impact on plan:** Both preserve plan intent exactly; the pair correction was forced by the plan's own string-level pins, and the lint restructure implements D-10's "generator passes strict lint without exceptions" mandate.

## Issues Encountered

- One compile slip during Task 2 GREEN (Duration has no .Sub — subtraction is direct); fixed inline before any commit, never left the working tree. The intervening-key pin initially contemplated a deadline-less TimerExpired; the plan's "на истечении окна действия нет" wording settled the guarded variant (see key-decisions).

## Authentication Gates

None — no authenticated external services involved.

## User Setup Required

None — the generator reads system xkb/keysym files that are present on the target machine (precondition verified read-only).

## Known Stubs

None. Both packages are fully functional; no TODO/FIXME/placeholder markers in the new code.

## Threat Flags

None — no security-relevant surface beyond the plan's threat model (T-02-01 accept / T-02-02 mitigate both unchanged; the burst-of-10 corpus is the T-02-02 mitigation).

## Next Phase Readiness

- 01-04 (ADR-002) can quote internal/hotkey/fsm_test.go as the executable specification of D-04/D-05 (classic-with-waiting, 300 ms, decision-at-expiry, modifier-use, 4th-tap rule).
- Phase 2 correction logic consumes layouts.ENToRU / layouts.RUToEN directly (registry-free, committed data).
- 01-03/01-05: both packages join the CI suite green; TEST-01 remains open for the e2e-side siblings per the shared-ID gate.

---
*Phase: 01-adr-paket-i-karkas-ibus-dvizhka*
*Completed: 2026-09-10*

## Self-Check: PASSED

All 7 key files exist on disk; layouts/.gitkeep confirmed removed; all four plan commits (c391333, 7930912, 17ed1d4, 3911e4c) found in history. Gates re-verified post-GREEN: both package suites green under -race, regeneration deterministic (empty diff), DO-NOT-EDIT marker first line, no Makefile in the repo.
