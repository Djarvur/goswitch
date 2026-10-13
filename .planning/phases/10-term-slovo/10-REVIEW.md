---
phase: 10-term-slovo
reviewed: 2026-10-13T22:15:24Z
depth: standard
files_reviewed: 16
files_reviewed_list:
  - internal/session/actor.go
  - internal/session/actor_test.go
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/config/load.go
  - internal/ctlsvc/ctlsvc.go
  - internal/ctlsvc/ctlsvc_test.go
  - test/e2e/case_term_word_vte.go
  - test/e2e/case_term_word_vte_flip.go
  - test/e2e/main.go
  - test/e2e/wezterm.go
  - mise.toml
  - docs/CONFIG.md
  - docs/adr/ADR-008-terminal-transport.md
  - docs/adr/adr-008-spike-log.md
  - skills/goswitch-config/SKILL.md
findings:
  critical: 0
  warning: 4
  info: 6
  total: 10
status: issues_found
---

# Phase 10: Code Review Report (incremental re-review, delta 10-07/10-08)

**Reviewed:** 2026-10-13T22:15:24Z
**Depth:** standard
**Files Reviewed:** 16
**Status:** issues_found

## Summary

Incremental re-review of the gap-closure delta since 59eb02f (plans 10-07/10-08):
the flap-tolerant verifyDead erasure, the default terminal identity anchor,
loud/sticky role diagnostics, the deferred terminal flip (defer-until-confirmed),
the stale-cache terminal route, the term-word-vte-flip e2e case and the machine
focus nudge. Verification gates run green on this tree: `go build ./...`,
`go vet`, `go test -race` over internal/session, internal/config, internal/ctlsvc.
The xrandr probe and the ydotool CLI forms were verified empirically against the
target system (ydotool 0.1.8, real xrandr output).

The delta itself is well built. The defer/flip state machine follows the epoch
discipline at every timer site (flipEpoch guards the grace deadline, flapEpoch
guards the wipe deadline) with all state under the actor mutex; the owner-pinned
invariants hold — FocusOut touches no flip slot, flip_after_correction=false
arms nothing (byte-identical ops paths, pinned by test), the stale-cache route
stays fail-closed for cache evidence (selection and autocorrect-boundary rounds
excluded, pinned by test), and the level-1 GTK path sees only no-op
releaseDeferredFlip calls. The config nil-vs-empty normalization is pinned by
Load-level tests; the watcher reload path routes through the same Load; docs,
SKILL.md and mise task are consistent with the code; the spike journal is honest
(FAIL/blocked rows retained).

Four warnings survive. Two are the prior review's race-shaped findings in the
terminal write branch, still unresolved — and WR-02's blast radius grew with
this delta (a falsely taught verifyDead now also arms the deferred-flip machine
for a GTK context). Two are new: the flap-window FocusIn cancellation cannot
distinguish a same-engine flap pair from a real window switch (the actor
receives no context identity, internal/engine/engine.go:466 — a wezterm-learned
class can leak into the next focused context), and the focus nudge's xrandr
geometry probe is provably dead (verified against real xrandr output on the
target).

## Narrative Findings (AI reviewer)

## Warnings

### WR-01: No re-entrancy guard on the terminal write branch — a second correction can interleave a second destructive burst (prior, unresolved)

**File:** `internal/session/actor.go:3502-3565` (executeLevel2, chunk loop 3518-3540), `3573-3579` (chunkBoundaryMoved)
**Issue:** carried over from the 2026-10-12 review unchanged — the delta did not
touch the guard surface. executeLevel2 releases the actor mutex at every chunk
boundary (one Gosched) so a real lifecycle callback can be observed; the
ctxEpoch re-check stops the round ONLY on FocusOut/Reset. A second correction
decision can still enter through the same unlocked window: a tap-series expiry
(Expiry → Double → startCorrection), a fired autocorrect confirm re-acquiring
the mutex mid-burst (autoConfirm:3026 → startRangeCorrection → classifyTerminal
→ executeLevel2 — the buffer is still pre-replace until the burst's commit, so
acConfirmRefusals passes), CorrectNow over ctl, or — new with 10-08 — a second
surrounding push racing the stale-cache route's own burst. Both bursts compute
BuildPlan from the same un-replaced buffer and both replay Backspaces over the
same field range; in a terminal that erases user text beyond intent. Exposure
with the default cap is ~1 s (20 ms per key).
**Fix:** arm a write-branch generation/flag under the mutex at executeLevel2
entry and refuse a second entry while one is in flight — the rungMu TryLock
precedent (skip with a D-20 reason, e.g. `correction-busy`, via skipCorrection +
settleCombo). Set before the first forward, clear on every return path
(including the context-changed branch).

### WR-02: VerifyExpiry carries no per-round tag — a stale deadline can consume the next pending round and falsely teach verify-dead (prior, unresolved; impact raised by 10-08)

**File:** `internal/session/actor.go:1336-1356` (VerifyExpiry), arm sites `3308`, `3416`; contrast the epoch form `2013-2023` (afterExpiry)
**Issue:** still open. pendingFix rounds all share the bare `a.VerifyExpiry`
callback with no round tag, while every 10-07/10-08 timer (flap wipe, flip
grace, verify-after) got the epoch discipline this delta built. The unguarded
window is unchanged: round 1's deadline fires and blocks on a.mu; the main path
resolves round 1 (Stop is a no-op — already fired) and arms round 2; the stale
callback then sees round 2's pending, drops it, records verify-timeout — and
`setVerifyDead(true)` teaches the behavioral terminal class for a context that
never missed a budget. The 10-08 delta RAISED the consequence: a falsely taught
verifyDead now (a) persists across flap pairs (10-07), (b) misroutes every
correction of the context to the destructive write branch, and (c) arms the
deferred-flip machine for a GTK context (mode flips deferred to the next
keystroke or 300 ms). The corpus pins only the settled-round case
(pending == nil), not the new-round case.
**Fix:** tag pendingFix with a monotonic epoch (the verifyEpoch field already
exists): capture `epoch := a.verifyEpoch` in the AfterFunc closure, pass it to
VerifyExpiry (or a wrapped callback), and no-op unless
`a.pending != nil && a.pending.epoch == epoch` — the afterExpiry form, byte
for byte.

### WR-03: The flap-window FocusIn cancellation has no context identity — a real window switch inside 10 ms leaks the learned terminal class into the next context

**File:** `internal/session/actor.go:1074-1090` (FocusIn branch, cancel at 1080), `1415-1451` (arm/cancel/expiry), `73` (focusFlapWindow); evidence `internal/engine/engine.go:250-273,466`
**Issue:** the actor's verifyDead is documented as "the behavioral terminal
class of the LIVE input context" (actor.go:304-312), but the wipe-deferral
mechanism cannot scope it to a context. HandleLifecycle receives only the kind
— every engine object (one per input context) forwards to the same actor
handler with no identity (engine.go:466; the engine logs its own `e.name` but
the handler signature drops it). So a REAL window switch — old context's
FocusOut arming the 10 ms wipe, then the new context's FocusIn arriving within
the window (the mint/switch round trip between the two D-Bus calls is
plausibly sub-10 ms on the local socket, the same magnitude as the 1–3 ms flap
pairs the journal shows) — cancels the wipe exactly like a flap pair. The
learned class (e.g. wezterm's verifyDead) then survives into the freshly
focused GTK context: every correction there routes to the terminal write
branch — no ADR-004 pre-verification, 20 ms/key pacing, and the 10-08
deferred-flip machine armed — for the whole focus episode (until that context
loses focus with no FocusIn within 10 ms). The unit corpus cannot catch this:
TestTerminalClass_VerifyDeadSurvivesFlapPair exercises the same
FocusOut→FocusIn sequence a real switch produces, and the 10 ms justification
("a real window change costs orders of magnitude more", actor.go:63-73) is an
assumption the cited journal does not cover (it measured same-engine flap
pairs only). Not corruption — the write branch corrects a GTK field correctly —
but it silently downgrades honest contexts to the legacy degradation path and
breaks the stated live-context invariant.
**Fix:** deliver the engine identity with lifecycle events (the engine already
holds `e.name`/`e.path`; the journal prints it) and cancel the armed wipe only
when the FocusIn's identity matches the one whose FocusOut armed it — or
compare emitter snapshots (the engine object that sent FocusIn vs the one that
armed the window). Different identity ⇒ let the wipe expire.

### WR-04: The focus nudge's xrandr geometry probe is provably dead — the click always rides the hardcoded 1280x800 fallback

**File:** `test/e2e/wezterm.go:345` (probe), `354-364` (nudgeFocusClick), `335-343` (constants)
**Issue:** `nudgeXrandrProbe` greps xrandr output for `[0-9]*x[0-9]* screen`,
but no real xrandr line has that shape — the header prints `current 2048 x 1280`
with spaces around the x and no "screen" suffix, mode lines print `2048x1280`
with no "screen". Verified on the target system: the grep exits 1 with no
output on the stand's real Xwayland display. The probe therefore NEVER answers,
`nudgeFocusClick` always falls back to (640,400), and the comments' claims —
"reads the live geometry of the stand's Xwayland display" and "the documented
1280x800 otherwise" — are false (2048x1280 is what the stand actually runs;
the fallback center is off-center at 31%/31%). On this stand's geometry the
point still lands inside a centered default-size terminal window, so the rows
pass by luck; on any other resolution/window size the nudge clicks outside the
fixture — focusing another window fires the daemon focus_in the gate awaits,
and the row types into the wrong surface (caught only downstream by the pty
oracle as a confusing failure).
**Fix:** parse the real header format, e.g.
`xrandr --current | grep -oP 'current \K[0-9]+ x [0-9]+'` split on `" x "`, or
drop the probe and read the geometry once via `xdpyinfo`/`xrandr --query` with
an honest error if unparseable; assert the clicked point inside the expected
window geometry if precision matters.

## Info

### IN-01: caseListUsage still omits term-negative-witnesses while claiming "every registry name" (prior, refreshed)

**File:** `test/e2e/main.go:117-131` (caseListUsage) vs `341` (registry), `275-278` (error text)
**Issue:** the 10-08 delta added term-word-vte-flip to the usage string, but
term-negative-witnesses — in the registry since 10-05, named by the
unknown-case error — is still missing from the `-case` help. The drift the
prior review flagged has grown by one entry.
**Fix:** add `" | term-negative-witnesses"`, or derive the usage from
caseRegistry() so the two cannot drift.

### IN-02: Save/Primary discard the wl-paste stderr diagnostic on the empty-classification path (prior, carried; file outside the current delta list)

**File:** `internal/clipboard/clipboard.go:85-97` (Save), `108-120` (Primary)
**Issue:** carried over from the 2026-10-12 review, still unresolved (the file
was not part of the 10-07/10-08 delta). wl-paste exit 1 with a non-empty
stderr can be a genuine connection/service failure, and the
empty-classification branch drops the whole error — had=false with no trace of
why. Diagnosability cost only; the CR-03 signal-kill guard is correct.
**Fix:** on the empty-classification branch, emit a Debug record with the
swallowed error (it carries exit status + stderr, never clipboard content).

### IN-03: boundRowText truncates by bytes and can split a UTF-8 rune (prior, unresolved; blast radius grew)

**File:** `test/e2e/case_term_word_vte.go:72-78`; now also used by `test/e2e/case_term_word_vte_flip.go:293` (the flip case's honest-abort row)
**Issue:** the 200-byte cap slices at a byte offset, so a multi-byte Cyrillic
rune at the boundary lands as an invalid sequence in the journal cell. The
10-08 flip case's abort path reuses the helper, so the artifact can now also
land in the flip section's rows.
**Fix:** back the cut off to a rune boundary, e.g. walk back while
`!utf8.RuneStart(text[termRowTextBound])`, or pass the result through
`strings.ToValidUTF8`.

### IN-04: The flip case's settlement gate accepts mismatch/timeout, but the comment and pass verdict claim "released by delivery confirmation"

**File:** `test/e2e/case_term_word_vte_flip.go:39-60` (isSettleRecord accepts match/timeout/mismatch), `137-141` + `218-222` (comment and verdict text)
**Issue:** the case comment defines the settlement рубеж as «the client's
confirming push (match) or the quiet budget (timeout)» — mismatch is not
listed — yet isSettleRecord accepts the mismatch conclusion too. That mirrors
the daemon's actual release sites (settleAfter releases the flip on mismatch
as well), so the ORDER oracle is faithful to the implementation; but when the
settlement was a timeout/mismatch, delivery was by definition NOT confirmed,
and the row's pass verdict «флип освободился подтверждением доставки»
overstates what happened. The pty-hex oracle still independently proves the
word landed, so the row's evidential value holds; this is journal-honesty
drift between the documented рубеж, the gate, and the verdict text.
**Fix:** record the settlement form in the observables cell
(`settlement=match|timeout|mismatch`) and soften the verdict text to name the
actual form, or restrict the gate to match/timeout to match the comment.

### IN-05: The E3 hazard is knowingly kept on the combo path of the terminal branch

**File:** `internal/session/actor.go:1580-1597` (deferCorrectionFlip's comboPending guard), `3545-3557` (executeLevel2 flip fork)
**Issue:** a combo-launched terminal correction never defers its flip: the
comboPending guard makes deferCorrectionFlip a no-op and settleCombo fires
flipTo synchronously right after the commit — the exact SetGlobalEngine-after-
commit ordering the E3 verdict forbids on the keystroke path. This is
deliberate and pinned (TestTerminalBranch_ComboFlipImmediateNotDeferred;
ADR-008 amendment: «combo-путь и путь flip=false байт-прежние»), and a
Shift+Control_R combo inside a terminal is a rare gesture. Recorded here so
the accepted residual is visible to the next revision of the flip machine.
**Fix:** none required now; if E3 ever reproduces on the combo path, the
deferred slot can serve it by letting settleCombo arm instead of execute when
the round was terminal.

### IN-06: newActor's SetOptions wiring omits Terminals — the identity anchor reaches the actor only through the watcher fold

**File:** `cmd/goswitchd/main.go:521-534` (SetOptions call; file outside the current delta list, cited as the consumer of the in-scope config surface), `internal/config/config.go:266-317` (Defaults claims), `internal/session/actor.go:1707-1740` (applySnapshot fold)
**Issue:** config.Defaults() documents that "an absent document and no document
classify identically" — but the startup SetOptions passes BackspaceCap,
ClipboardRung, FlipAfterCorrection, ModeSwitchChord and SoundEnabled and NOT
Correction.Terminals (nor the autocorrect fields). Every real desktop path is
covered because a watcher is always attached and the startup FoldAppliedConfig
folds the snapshot (including the anchor) — the degenerate no-HOME path
(`config.DefaultPath() == ""` → watcher nil) is the only one where the anchor
documented in Defaults() never reaches the actor, silently leaving
classification to the behavioral/role sources only. Moot without a desktop
session, hence Info — but the wiring contradicts the Defaults() comment and is
one new Options field away from being uniform.
**Fix:** add `Terminals: cfg.Correction.Terminals` (and the autocorrect
fields, for the same uniformity) to newActor's SetOptions literal.

---

_Considered, no finding_: the defer/flip epoch discipline is applied uniformly
(flipGraceExpiry and verifyDeadWipeExpiry both re-check under the mutex; a
second correction supersedes a pending flip via the same slot, pinned by test);
the stale-cache route settles the combo on every executeLevel2 exit including
the backspace-cap refusal; phrase rounds ride the same stale-cache semantics
correctly; flip_after_correction=false leaves the ops paths byte-identical
(pinned); the GTK level-1 path and combo flips see only no-op release calls;
the ydotool `click --delay N 1` / `mousemove --delay N x y` forms are valid
0.1.8 syntax (verified against the installed binary's --help — the stand rows
at 21:32 confirm the nudge path live); the ac_role_errors token insertion
moves no old token and is pinned in ctlsvc_test; the terminals decode corpus
pins nil→anchor, `[]`→opt-out and the strict list; watch.go routes every
reload through Load, so hot reload honors the normalization; adoptLoad serves
Defaults for the absent adopted file (anchor in force); the ADR-008 amendment
and spike-log sections are dated, journal-anchored, and keep the FAIL/blocked
rows (honest evidence); the D7 contract text in docs/CONFIG.md is preserved.

_Reviewed: 2026-10-13T22:15:24Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
