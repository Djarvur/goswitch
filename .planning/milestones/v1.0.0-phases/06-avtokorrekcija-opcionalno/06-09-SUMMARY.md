---
phase: 06-avtokorrekcija-opcionalno
plan: 09
subsystem: session-actor
tags: [golang, atomic, sync-pointer, ibus, dbus, deadlock, gap-closure, tdd]
requires:
  - phase: 06-08
    provides: the autocorrect corpus this plan must not disturb; the D-54 counters and the single correction pipeline
  - phase: 05-03
    provides: the flipTo-through-seam corpus and the deadline-under-mutex pin this plan preserves and re-confirms
provides:
  - lock-free Actor.AttachEngine — one atomic Store of the emitter slot, no a.mu, answering the factory's re-entrant CreateEngine during a flip (G-6-1 closed mechanically)
  - engSlot atomic.Pointer[emitterSlot] + the emitter() snapshot accessor; all 25 former a.eng reads on snapshots, nil semantics preserved
  - reentrant G-6-1 unit corpus: TestActor_ReentrantAttachDuringFlip + TestActor_AttachEngineWhileFlipInFlight
  - fake-bus reattach mode + two wire witnesses (TestFactoryReentrantCreateEngineAnswersDuringAwait / ...BlockingHandlerIsOnlyGate)
  - ADR-006 Amendment 2026-10-02 documenting the G-6-1 mechanism with the pin verdict
affects: [06-10 (live wired proof), phase-06-verification, flip e2e matrix]
tech-stack:
  added: []
  patterns:
    - "atomic.Pointer[wrapper] for an interface-valued shared slot (Go atomics store pointers, not interfaces)"
    - "snapshot-accessor discipline: one atomic Load per computation point, nil-guards byte-preserved"
    - "reattach hook on the fake bus: a controlled synchronous re-entrant call in the dispatch goroutine before the reply (the setFail precedent)"
key-files:
  created:
    - .planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-09-task1-red.json
  modified:
    - internal/session/actor.go
    - internal/session/actor_test.go
    - engine/conn_switcher_test.go
    - docs/adr/ADR-006-two-engine-revision.md
key-decisions:
  - "Mechanism (b) of the fix direction: AttachEngine made lock-free instead of moving the SetGlobalEngine await off the mutex — the deadline-under-mutex pin (05-03) is kept and re-confirmed, not replaced"
  - "engSlot atomic.Pointer[emitterSlot]: the interface rides in a one-field wrapper because Go atomics store pointers, not interfaces"
  - "Out-of-order stores between concurrent CreateEngine calls accepted (pre-fix equivalent ordering) — no ordering machinery added (T-06-09-02)"
  - "Witness B's blocked-mint collection is proven by a fresh flip answering with the hook still armed — the collected state itself serves the live path"
requirements-completed: [SWCH-01, SWCH-02]
coverage:
  - id: D1
    description: "Lock-free AttachEngine: the engine factory's re-entrant CreateEngine completes during a flip without touching a.mu (atomic emitter slot)"
    requirement: SWCH-01
    verification:
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_ReentrantAttachDuringFlip"
        status: pass
      - kind: unit
        ref: "tests/internal/session/actor_test.go#TestActor_AttachEngineWhileFlipInFlight"
        status: pass
      - kind: other
        ref: "grep gate: AttachEngine body contains engSlot.Store and no a.mu.Lock"
        status: pass
    human_judgment: false
  - id: D2
    description: "Flip contract unchanged: D-36 record order (mode → switch_engine → UpdateModeSymbol → display), 150 ms wedge guard, WARN-not-fatal, nil seam, autocorrect contour untouched — the whole pre-existing corpus green under -race without weakening"
    requirement: SWCH-02
    verification:
      - kind: unit
        ref: "mise run ci (build + vet + golangci-lint + go test -race ./...)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Sealed wire witness: the fake bus's SetGlobalEngine synchronously re-enters factory.CreateEngine before the reply; instant handler answers inside the await, blocking handler is the only gate (DeadlineExceeded), production engine/ code untouched"
    requirement: SWCH-01
    verification:
      - kind: unit
        ref: "tests/engine/conn_switcher_test.go#TestFactoryReentrantCreateEngineAnswersDuringAwait"
        status: pass
      - kind: unit
        ref: "tests/engine/conn_switcher_test.go#TestFactoryReentrantCreateEngineBlockingHandlerIsOnlyGate"
        status: pass
    human_judgment: false
  - id: D4
    description: "ADR-006 Amendment 2026-10-02 — G-6-1: diagnosis (journal 00:05:04-00:05:17 + probes), the property, the atomic-slot decision, what does NOT change (the pin), evidence"
    requirement: SWCH-02
    verification:
      - kind: other
        ref: "grep gate: 2026-10-02 + G-6-1 + AttachEngine present in docs/adr/ADR-006-two-engine-revision.md"
        status: pass
    human_judgment: true
    rationale: "The amendment's honesty — that the pin is confirmed rather than silently contradicted, and that the diagnosis faithfully reflects the UAT journal — is a documentation judgment; automated gates check only structural presence"
coverage-note: "Live wired proof of the closed gap (first keystroke after a flip reaches the field) is plan 06-10's deliverable, deliberately out of this plan's scope"
duration: 25min
completed: 2026-10-03
status: complete
actuals:
  tokens: 9858
  tasks: 2
  commits: 4
  plan_head_before: b600ce4d48bc4e08c6c33a39631f4c4812a17117
---

# Phase 06 Plan 09: Gap Closure G-6-1 (Flip Factory Self-Block) Summary

**Lock-free AttachEngine (atomic emitter slot) ends the flip's factory self-deadlock: the engine factory now answers CreateEngine during a flip's SetGlobalEngine await, with the deadline-under-mutex pin re-confirmed and sealed by a reentrant unit corpus plus fake-bus wire witnesses.**

## Performance

- **Duration:** 25 min
- **Started:** 2026-10-03T19:12:53Z
- **Completed:** 2026-10-03T19:37:30Z
- **Tasks:** 2 (TDD: RED→GREEN on Task 1)
- **Files modified:** 5

## Accomplishments

- G-6-1 closed mechanically: `AttachEngine` is one atomic `engSlot.Store(...)` — the re-entrant CreateEngine path (ibus-daemon answers a flip by minting the target engine through the daemon's factory) no longer waits on `a.mu`, so the flip's 150 ms deadline is no longer burned by the daemon's own factory
- All 25 former `a.eng` reads moved to `emitter()` snapshots at the same computation points; nil-before-first-mint semantics byte-preserved; the flip contract (D-36 record order, wedge-guard deadline, WARN-not-fatal, nil seam) and the autocorrect contour untouched
- Strict TDD: RED committed first with two budget-fatal failing tests + validated red-evidence JSON (`RED_EVIDENCE_OK`), then the GREEN implementation
- Fake bus gained the reattach mode (the exact live G-6-1 shape) and two witnesses: instant handler answered inside the await; blocking handler is the only gate — production `engine/` code unchanged
- ADR-006 honestly amended with the dated 2026-10-02 section: diagnosis, property, decision, what does NOT change, evidence

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): reentrant AttachEngine corpus** - `bfaf590` (test)
2. **Task 1 (GREEN): lock-free AttachEngine** - `0a17fea` (feat)
3. **Lint residue of Task 1** - `0ceddc3` (style)
4. **Task 2: reattach witnesses + ADR-006 amendment** - `04f5dee` (test)

_Note: TDD plan — Task 1 produced the RED (`test`) and GREEN (`feat`) commits; the RED evidence JSON rode the RED commit._

## Files Created/Modified

- `internal/session/actor.go` — `emitterSlot` type + `engSlot atomic.Pointer[emitterSlot]`, lock-free `AttachEngine`, `emitter()` snapshot accessor, all reads on snapshots, G-6-1 paragraphs in the `switchTimeout` and `flipTo` why-comments
- `internal/session/actor_test.go` — `TestActor_ReentrantAttachDuringFlip`, `TestActor_AttachEngineWhileFlipInFlight`, `attachBudget` const
- `engine/conn_switcher_test.go` — fake-bus `reattach` hook + `setReattach`, `dialFakeBus`, `reattachHandler` double, `flipDeadline`/`mintSeq` consts, the two `TestFactoryReentrantCreateEngine*` witnesses
- `docs/adr/ADR-006-two-engine-revision.md` — Amended status line + the "Amendment 2026-10-02 — G-6-1" section
- `.planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-09-task1-red.json` — the RED run record (exit 1, both target tests failing on the planned assertions)

## Decisions Made

- Mechanism (b) chosen per plan: free the factory path from the mutex instead of moving the await — the async WR-01-handoff alternative stays rejected; ADR-006 records the pin verdict
- `atomic.Pointer[emitterSlot]` wrapper because Go atomics do not store interfaces
- No ordering machinery for concurrent CreateEngine stores (T-06-09-02 accepts the pre-fix equivalent ordering)
- `//nolint:ireturn` on `emitter()` with justification: the plan mandates the interface-returning accessor; the config's point-disable-with-justification discipline applied inline (no .golangci.yml relaxation added)

## Deviations from Plan

None — plan executed exactly as written. (The `0ceddc3` style commit is Task 1's own lint line-wrap surfacing in the shared CI gate; it touches only the Task 1 test file.)

## Issues Encountered

- The red-evidence checker parses node-test TAP summaries, so the go-test output was transcribed to TAP in the evidence JSON — the same transcription precedent as 05-03
- golangci-lint (funcorder/ireturn/lll) required three mechanical adjustments during the green iteration: `emitter()` moved after the last exported Actor method, the justified `//nolint:ireturn`, and two long-line wraps

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- The airtight half of G-6-1 is done; the LIVE wired proof is plan 06-10 (immediate-keystroke case on the owner's desktop) — the matrix flip rows stay red until that lands
- UAT test 1's issue stays open until 06-10 provides the live evidence; ADR-007's acceptance-evidence row 5 is unaffected by this plan

## Self-Check: PASSED

- All 6 key files exist on disk (4 code/docs artifacts + red-evidence JSON + this SUMMARY)
- All 4 commit hashes present in git log (bfaf590, 0a17fea, 0ceddc3, 04f5dee)
- Claim spot-check: `engSlot.Store` present in the AttachEngine body; `mise run ci` green on the final tree; both reentrant actor tests and both reattach witnesses green under `-race`

---
*Phase: 06-avtokorrekcija-opcionalno*
*Completed: 2026-10-03*
