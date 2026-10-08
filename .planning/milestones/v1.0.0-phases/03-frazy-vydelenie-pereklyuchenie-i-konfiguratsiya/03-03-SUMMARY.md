---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
plan: 03
subsystem: input-correction
tags: [ibus, selection, anchor-pos, clipboard-rung, wl-clipboard, tdd, e2e, spike]

# Dependency graph
requires:
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 01
    provides: ConvertRuns (the one conversion entry the selection range reuses, D-23), correctionRange seam, Buffer.Phrase/ReplacePhrase
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 02
    provides: config.Correction{BackspaceCap, ClipboardRung} (the rung is opt-in via config, D-28), -config flag
provides:
  - engine.EventHandler.HandleSurroundingText(text, cursorPos, anchorPos) — the seam widened, the anchor reaches the actor (Pitfall 1); engine.KeyV = 0x076
  - internal/correct.SelectionRange(cursor, anchor) (start, end, active) + VerifyRangeAt(text, start, want) — the D-30 geometry and the Pitfall 6 range verify
  - internal/session selection branch — Double under an active selection corrects exactly the selected range in either geometric direction; session.Options{BackspaceCap, ClipboardRung} + SetOptions
  - internal/clipboard — Save/Set/Clear/Restore over wl-paste --no-newline / wl-copy --trim-newline / wl-copy --clear, Runner seam, stdin-only content, per-call deadlines
  - live e2e cases select-smoke (spike), select-correct (GTE), select-clipboard (unreachable-primary mechanics); stand startDaemonArgs/restartDaemonWithArgs
affects: [03-04-config, 03-07-matrix-v2]

actuals:
  tokens: 28547   # chars/4 over the realized diff (git diff c7e54f4..HEAD = 114189 chars)
  tasks: 3
  commits: 5       # measured: git rev-list --count c7e54f4..HEAD
  plan_head_before: c7e54f47078383a0a8b5971b738b2e55e5b37b18

tech-stack:
  added: []
  patterns:
    - "no-pipes subprocess shape for forking binaries: wl-copy's background grandchild holds piped stdout write-ends open, deadlocking cmd.Run — fork-shaped binaries get no pipes, only stdin"
    - "range-anchored verify: VerifyRangeAt checks the converted text AT the selection position — MatchesSuffix only sees the before-cursor prefix (Pitfall 6)"
    - "interleaved sequence log across two guarded recorders (sink forward hook + subprocess fake hook) pins cross-component call order"

key-files:
  created:
    - internal/clipboard/clipboard.go
    - internal/clipboard/clipboard_test.go
    - test/e2e/case_select.go
    - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/red-evidence/03-03-task2-red.json
    - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/red-evidence/03-03-task3-red.json
  modified:
    - engine/engine.go
    - engine/engine_test.go
    - engine/keys.go
    - internal/correct/verify.go
    - internal/correct/verify_test.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - cmd/goswitchd/main.go
    - test/e2e/main.go
    - mise.toml

key-decisions:
  - "Selection takes precedence over the word path at the Double decision, and the buffer HardResets after a selection replacement: a selection may span text the buffer never mirrored, so it can no longer stand in for the field"
  - "Selection ladder geometry is computed directly (offset = start-cursor signed by the cursor's side, nchars = exactly the range) — BuildPlan's token+tail arithmetic cannot express a range right of the cursor"
  - "Spike verdicts (A1/A2/A5 closed live): zenity never pushes an anchor with a selection (D-30 degradation, word path keeps working); GTE pushes cursor=0/anchor=len under ctrl+a (the positive-offset shape); chromium pushes cursor=len/anchor=0 and a commit REPLACES the active selection"
  - "wl-copy gets NO pipes in the exec runner: its forked background grandchild holds piped stdout write-ends open and deadlocks cmd.Run for the full watchdog (live finding of the first select-clipboard run)"
  - "select-clipboard runs the plan's unreachable-primary fallback: every anchor surface on this desktop applies the primary rung under a selection, so no live correction reaches verify-after mismatch — the case pins the -config spawn wiring and the round-trip mechanics through the daemon's own client"

requirements-completed: [CORR-03]

coverage:
  - id: D1
    description: "The anchor survives the seam (Pitfall 1): SetSurroundingText forwards both wire positions to EventHandler.HandleSurroundingText(text, cursorPos, anchorPos); all implementors updated in one task; the DEBUG trace keeps cursor_pos and anchor_pos"
    verification:
      - kind: unit
        ref: engine/engine_test.go#TestEngine_SurroundingTextForward
        status: pass
      - kind: command
        ref: "grep -n anchorPos engine/engine.go → signature + wire call + debug trace (3 hits)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Spike table per surface (A1/A2/A5 closed live, input for matrix v2 / 03-07): anchor push behavior, working select-all name, actual replacement rung under selection"
    verification:
      - kind: e2e
        ref: mise run e2e-select-smoke (PASS with per-surface verdicts, see Spike Table below)
        status: pass
    human_judgment: false
  - id: D3
    description: "Double Right Shift under an active selection corrects exactly the selected range (CORR-03, D-30): geometry in both directions, tail untouched, mixed selections convert by runs (D-23), no selection keeps the Phase-2 word path"
    verification:
      - kind: unit
        ref: internal/correct/verify_test.go#TestSelectionRange
        status: pass
      - kind: unit
        ref: internal/correct/verify_test.go#TestVerifyRangeAt
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_DoubleTapSelectionCorrects
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_SelectionLeftToRight
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_SelectionMixedConverts
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_NoSelectionStillWord
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_SelectionNoLetters
        status: pass
      - kind: e2e
        ref: mise run e2e-select-correct (PASS — GTE, mixed field → привет привет)
        status: pass
    human_judgment: false
  - id: D4
    description: "Clipboard rung opt-in with best-effort restore (D-28/D-29): zero subprocesses at the default config, rung-C order after verify-after mismatch, stdin-only content, WARN-only restore failure, INFO records content-free"
    verification:
      - kind: unit
        ref: internal/clipboard/clipboard_test.go#TestClipboard_SaveSetRestore
        status: pass
      - kind: unit
        ref: internal/clipboard/clipboard_test.go#TestClipboard_Timeouts
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_ClipboardRungDisabledByDefault
        status: pass
      - kind: unit
        ref: internal/session/actor_test.go#TestActor_ClipboardRungAfterMismatch
        status: pass
      - kind: command
        ref: "grep cmd.Stdin → STDIN-ONLY; grep 'clipboard restore failed' actor.go → LOG-SHAPES-OK"
        status: pass
      - kind: e2e
        ref: mise run e2e-select-clipboard (PASS — -config spawn + live round-trip)
        status: pass
    human_judgment: false
  - id: D5
    description: "D-30 continuity: without a selection the Double decision runs the Phase-2 word path unchanged — the whole Phase-2 corpus stays green without a single expectation edit"
    verification:
      - kind: e2e
        ref: mise run e2e-word (PASS word-en-ru)
        status: pass
      - kind: e2e
        ref: mise run e2e-matrix (16/16 PASS on the final tree)
        status: pass
    human_judgment: false

duration: 45 min
completed: 2026-09-15
status: complete
---

# Phase 3 Plan 3: Выделение — anchorPos-seam, ветка выделения, clipboard-ступень Summary

**Двойной Right Shift при активном выделении исправляет ровно выделенный диапазон в обеих направлениях геометрии (якорь доходит до актора через расширенный seam), спайк запинил фактическое поведение поверхностей живьём, и opt-in clipboard-ступень (wl-copy/wl-paste round-trip с best-effort восстановлением) закрыла rung C лестницы ADR-003.**

## Performance

- **Duration:** 45 min (09:35–10:20 UTC)
- **Tasks:** 3 (1 auto spike + 2 tdd: RED→GREEN, RED_EVIDENCE_OK ×2)
- **Files modified:** 15 (13 code/config + 2 red-evidence records)
- **Commits:** 5 (measured from plan_head_before c7e54f4)

## Spike Table (Task 1 — the plan's A1/A2/A5 closure, input for matrix v2 / 03-07)

| Surface | Anchor pushed? | Working select name | Actual rung under selection (Task-1 word path probing) |
|---|---|---|---|
| zenity (GTK4 entry) | NO — no surrounding push with anchor ≠ cursor ever arrives (the selection key lands; nothing observable follows) | ctrl+a selects silently; "Control+a" types a literal "ca" through the first-letter fallback (Pitfall 5 class) | n/a — selection unobservable → D-30 documented degradation: the word path keeps handling the Double |
| gnome-text-editor 46.3 | YES — ctrl+a pushes cursor=0, anchor=len (LEFT-TO-RIGHT selection) | ctrl+a | pre-branch: nothing applied (before-cursor empty at cursor=0 → verify skip); with the Task-2 branch: delete+commit applies cleanly over the range (positive offset — the select-correct PASS) |
| chromium (google-chrome 153) | YES — ctrl+a pushes cursor=len, anchor=0 (RIGHT-TO-LEFT selection) | ctrl+a | CommitText REPLACES the active selection (typing-over-selection semantics): the field settled at exactly «привет» |

Canonical ydotool name pinned live: **ctrl+a** (case-insensitive letter works; the "Control" modifier token does not exist in ydotool 0.1.8's table and falls back to typing literal characters).

## Accomplishments

- Seam widened (Pitfall 1): `HandleSurroundingText(text string, cursorPos, anchorPos uint32)` — the anchor reaches the actor on every surface; the actor caches the full push (`selectionState{full, cursor, anchor}`); the DEBUG trace keeps both positions; all implementors (Actor + 3 test doubles) updated in the one Task-1 iteration.
- Selection branch (CORR-03, D-30, Pitfall 6): an active selection in the last push takes precedence over the word path; the range `[min(cursor,anchor), max(cursor,anchor))` is clamped to the reported text (T-03-03-03); conversion runs through the same `ConvertRuns` (D-23); the ladder deletes with the offset SIGNED by the cursor's side (`start-cursor` — negative left of the cursor, zero-or-positive right of it) and `nchars` exactly the range — never a rune more; the verify-after round is range-anchored (`VerifyRangeAt` — `MatchesSuffix` cannot see a selection right of the cursor).
- Clipboard rung (D-28/D-29, Pitfall 3, ADR-003 rung C): `internal/clipboard` with the pinned flags (`wl-paste --no-newline` byte-exact save, `wl-copy --trim-newline` set with content through `cmd.Stdin` ONLY — the T-03-03-01 argv prohibition, `wl-copy --clear` for the empty original); the rung fires only on verify-after mismatch and only when `Options.ClipboardRung` armed it (zero subprocesses at the default), in the order Save → Set(converted) → the 4-event Ctrl+V forward burst (the owner prototype's wire shape: `KeyV = 0x076`, evdev keycodes 29/47) → best-effort Restore; a restore failure is the WARN `clipboard restore failed`, never an operation error; clipboard content never enters any log record at any level.
- Daemon wiring: `goswitchd` feeds `session.Options{BackspaceCap, ClipboardRung}` from `config.Load` at startup (both BuildPlan sites consume the Options-fed cap; the zero Options falls back to `DefaultBackspaceCap`). The snapshot consumption on hot reload is plan 03-04.
- Live gates: `e2e-select-smoke` (per-surface verdicts, reproducible across two runs), `e2e-select-correct` (GTE, positive-offset geometry, mixed field `ghbdtn привет` → `привет привет`), `e2e-select-clipboard` (-config spawn + live round-trip through the daemon's own client, owner bytes restored); regressions `e2e-word` PASS and the full matrix v1 16/16 PASS on the final tree.

## Task Commits

Each task committed atomically (TDD: RED before GREEN):

1. **Task 1: Seam anchorPos + живой спайк select-smoke** — `1050e39` (feat)
2. **Task 2: Ветка выделения — геометрия в обе стороны** — `780db03` (test, RED) + `1181111` (feat, GREEN)
3. **Task 3: Clipboard-ступень opt-in** — `88a37d1` (test, RED) + `dbb5816` (feat, GREEN)

**Plan metadata:** this commit (docs: complete plan)

## TDD Gate Compliance

Both tdd tasks followed RED → GREEN with machine-validated red evidence:

| Task | RED commit | Evidence record | Verdict | GREEN commit |
|------|-----------|-----------------|---------|--------------|
| 2 | `780db03` test(03-03) | red-evidence/03-03-task2-red.json | RED_EVIDENCE_OK (6 top-level / 10 with subtests failing on assertions, 131 green; TestActor_NoSelectionStillWord green immediately by design — the Phase-2 continuity pin) | `1181111` feat(03-03) |
| 3 | `88a37d1` test(03-03) | red-evidence/03-03-task3-red.json | RED_EVIDENCE_OK (4 top-level / 10 with subtests failing on assertions, 53 green; TestActor_ClipboardRungDisabledByDefault green immediately by design — the D-28 default pin) | `dbb5816` feat(03-03) |

No REFACTOR commit — the GREEN implementations landed lint-clean under the strict v2 config (all findings fixed inside the GREEN iterations per D-08).

## Decisions Made

- Selection precedence + buffer death: the Double decision checks the selection first (an empty buffer does not block a selection correction — the range comes from the client push); after a selection replacement the buffer HardResets instead of being edited (a selection may span text the buffer never mirrored).
- Selection geometry computed in the actor (signed offset), not through `BuildPlan`: the token+tail plan cannot express a range right of the cursor; level 1 is the only reachable level for selections (they exist only because a surrounding push carried them).
- The rung's Restore follows the plan's literal order (immediately after the forward burst): the client processes the forwarded Ctrl+V asynchronously, so the restore can race the paste — the owner-accepted best-effort qualification of D-29 covers exactly this; documented here for the matrix v2 decision (03-07).
- `wl-copy --trim-newline` on Restore may strip one trailing newline of the original (the plan-pinned flags): the live case compares whitespace-trimmed — a documented nuance of the pinned flag set, not a deviation.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] wl-copy piped-stdout deadlock in the exec runner**
- **Found during:** Task 3 GREEN (first live select-clipboard run)
- **Issue:** `wl-copy` forks a background grandchild that serves the clipboard indefinitely; with `cmd.Stdout`/`cmd.Stderr` set to buffers, `cmd.Run()` waits for the pipe write-ends to close — i.e. forever. The case hung the full 3-minute watchdog and left a truncated probe in the clipboard.
- **Fix:** `execRunner` gives wl-copy NO pipes (it is silent on success; the error carries the exit status alone); `wl-paste` keeps captured stdout. The rung then completed in normal time and the owner's clipboard restored byte-faithfully.
- **Files modified:** internal/clipboard/clipboard.go
- **Verification:** `mise run e2e-select-clipboard` PASS ×2; unit corpus green; ci green
- **Committed in:** `dbb5816`

**2. [Rule 1 - Plan-literal] The plan's spike text `injectText("ghbdtn привет")` is not injectable**
- **Found during:** Task 1 (spike authoring)
- **Issue:** ydotool cannot type Cyrillic through the EN XKB group (the 03-01 live finding); the plan's literal injection would produce mojibake, not the mixed field.
- **Fix:** The select-correct case assembles the mixed field through both real printing branches (EN transit + flip + RU engine commits) — the `assembleMixedField` helper; the spike itself uses the EN-only probe field (the rung verdict does not depend on the field's script mix).
- **Files modified:** test/e2e/case_select.go
- **Verification:** e2e-select-smoke and e2e-select-correct PASS
- **Committed in:** `1050e39`, `1181111`

**3. [Rule 3 - Blocking] RED-phase fixture crash in TestClipboard_NewProductionRunner**
- **Found during:** Task 3 RED (first run)
- **Issue:** `t.Setenv` inside a `t.Parallel()` test panics ("cannot set environment variables in parallel tests") — a fixture crash that killed the whole package run before the target assertions could report (INVALID_RED material, #3770).
- **Fix:** The test runs serially (no `t.Parallel()`); the RED re-run then failed on assertions as required.
- **Files modified:** internal/clipboard/clipboard_test.go
- **Verification:** RED_EVIDENCE_OK for 03-03-task3-red.json
- **Committed in:** `88a37d1`

**4. [Rule 3 - Blocking] Strict-lint findings inside the GREEN iterations (D-08)**
- **Found during:** Tasks 1–3 (mise run ci)
- **Issue:** dupl/gochecknoglobals/gofmt/lll/perfsprint (T1); cyclop/funcorder/goconst/gosec/gofmt/lll (T2); cyclop ×2/err113/errorlint/gochecknoglobals/gocognit/goconst/gocritic/gofmt ×3/lll ×4/mnd/nolintlint/wrapcheck + the noctx directive migration to startDaemonArgs (T3).
- **Fix:** Driver-struct dedup of the spike probes; candidates as a function; extraction of subtest bodies/helpers (roundTripRestoresOwnerBytes, rungFullSequenceInOrder, clipboardRoundTripLive, assertNoClipContentInINFO); sentinels + %w wrapping; named constants; #nosec G115 with the input-field bound argument; the //nolint:noctx directive follows the exec.Command into startDaemonArgs. All inside the same iterations.
- **Verification:** `mise run ci` green after each fix (lint 0 issues)
- **Committed in:** all five commits

---

**Total deviations:** 4 auto-fixed (1 live bug, 1 plan-literal resolution, 2 blocking tooling/fixture)
**Impact on plan:** All resolved toward the locked must-haves; no scope creep. The spike table feeds 03-07 exactly as the plan scheduled.

## Issues Encountered

None beyond the deviations above. All live gates green on the target desktop: e2e-select-smoke (×2, verdicts reproducible), e2e-select-correct, e2e-select-clipboard (×2), e2e-word, e2e-matrix 16/16; `mise run ci` green at every task boundary.

## Known Limitations (tracked for 03-07)

- The daemon-side rung has no live driver on this desktop (every anchor surface applies the primary rung): select-clipboard pins the -config wiring and the subprocess mechanics — the plan's prescribed unreachable-primary fallback; whether a rung case enters matrix v2 is the 03-07 decision.
- The restore can race the client's asynchronous paste processing (plan-literal order; D-29 best-effort qualification).
- `wl-copy --trim-newline` may strip one trailing newline of the restored original (pinned-flag nuance).
- zenity-class surfaces without an anchor push degrade to the word path by design (D-30).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `session.Options` + `SetOptions` is the startup half of the 03-04 snapshot consumption (Watcher.Snapshot → SetOptions on reload; the stale-timer window rule rides with it).
- The spike table is the direct input for the matrix v2 `expect_sel_*` fields (03-07): zenity = no anchor (degradation row), GTE = positive-offset selection row, chromium = commit-replaces-selection row.
- `startDaemonArgs`/`restartDaemonWithArgs` lets matrix v2 drive per-case `-config` documents (the config-reload cases of 03-07).
- `clipboard.Client`-level Runner seam is ready for any future rung-level status surface (03-06 goswitchctl).

## Self-Check: PASSED

All key-files exist on disk; all five task commits found in history; plan ledger measured 5 commits from plan_head_before c7e54f4 (matches `commits:` in frontmatter); `mise run ci` green at the final tree.
