---
phase: 10-term-slovo
reviewed: 2026-10-12T21:56:44Z
depth: standard
files_reviewed: 27
files_reviewed_list:
  - docs/adr/ADR-008-terminal-transport.md
  - docs/adr/adr-008-spike-log.md
  - docs/CONFIG.md
  - internal/appid/appid.go
  - internal/appid/appid_registry_test.go
  - internal/appid/export_test.go
  - internal/clipboard/clipboard.go
  - internal/clipboard/clipboard_test.go
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/correct/stretch.go
  - internal/correct/stretch_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - skills/goswitch-config/SKILL.md
  - test/e2e/case_autocorrect.go
  - test/e2e/case_spike_caps.go
  - test/e2e/case_spike_wezterm.go
  - test/e2e/case_switch.go
  - test/e2e/case_term_negative.go
  - test/e2e/case_term_word_vte.go
  - test/e2e/case_term_word_wezterm.go
  - test/e2e/fixtures/gs-e2e-wezterm-stand.lua
  - test/e2e/main.go
  - test/e2e/preflight.go
  - test/e2e/spike_journal_test.go
findings:
  critical: 0
  warning: 2
  info: 3
  total: 5
status: issues_found
---

# Phase 10: Code Review Report

**Reviewed:** 2026-10-12T21:56:44Z
**Depth:** standard
**Files Reviewed:** 27
**Status:** issues_found

## Summary

Scope: the whole phase-10 change set — the convertible stretch geometry
(internal/correct), the terminal-class seam and terminal write branch
(internal/session/actor.go), the correction.terminals schema key
(internal/config), the PRIMARY clipboard read (internal/clipboard), the
a11y registry interest fix (internal/appid), the new e2e live cases
(VTE/wezterm/negative-witnesses) with their stand plumbing, and the
contract/ADR docs.

High-level assessment: the phase is in good shape. The stretch geometry
keeps the converter untouched and the replace half covers the whole run;
the terminal write branch forbids the clipboard rung by construction
(noRung carries no payload at all); the config key follows the D-54/D-53
precedents with the ceiling and the empty-pattern trap pinned by tests;
the free-abort criterion holds (empty-buffer / convert-failed / no-engine
/ backspace-cap all refuse before the first destructive key). Test
reliability is strong — the corpus pins the stale-expiry no-op, the
word-free class marker, chunk-boundary focus loss, noRung-on-mismatch and
the paced wire.

Two warnings survive adversarial review, both in the new terminal write
branch of actor.go, both race-shaped: the chunk-boundary mutex release
opens a re-entrancy window that the ctxEpoch guard does not cover (a
second correction decision can interleave a second destructive burst),
and VerifyExpiry — unlike afterExpiry — carries no per-round tag, so a
late deadline callback can consume the next pending round and falsely
teach verify-dead. Neither reproduces in the happy path; both need
unusual timing (a decision landing inside a burst window; a deadline
callback blocked on the mutex across a re-arm).

Considered and NOT flagged (documented, deliberate trade-offs): the
mutex-held 20 ms pacing sleeps and the resulting ProcessKeyEvent reply
latency for combo-launched bursts (plan 10-05's accepted wire-packing
price); the live GetRole under the actor mutex bounded at 25 ms
(T-10-03-03 wedge guard); DEBUG-level field/word logging as the opt-in
-diagnostic mode (D-20/D-21 discipline held at INFO); compileACPatterns
failing open to the role gate on a compile error impossible from a
validated document (documented defensive branch); the e2e re-pin row
recording a blocked verdict without failing its case (the plan's split).

## Narrative Findings (AI reviewer)

## Warnings

### WR-01: No re-entrancy guard on the terminal write branch — a second correction can interleave a second destructive burst

**File:** `internal/session/actor.go:3180-3202` (chunk loop), `3224-3230` (chunkBoundaryMoved)
**Issue:** executeLevel2 releases the actor mutex at every chunk boundary
(one Gosched) so a real lifecycle callback can be observed — that is the
ctxEpoch re-check, which stops the round ONLY on FocusOut/Reset. Nothing
stops a second correction decision from entering through the same
unlocked window: a previously armed tap-series timer expiring mid-burst
(Expiry → Double → startCorrection), the combo press path
(feedKey → startCorrection at line 2306), a fired autocorrect confirm
(line 2764 — acConfirmRefusals revalidates buffer content, which is still
pre-replace until the burst's commit), or CorrectNow over ctl. Both
bursts compute their BuildPlan from the same un-replaced buffer, so both
replay Backspaces over the same field range and both commit — the second
deletes runes the first already replaced, corrupting the field (in a
terminal: erased user text beyond intent). Reachable without any focus
change; the guard the phase built (ctxEpoch) does not see this class.
Burst windows are long enough to matter: 20 ms per key with the default
cap of 50 gives a ~1 s exposure.
**Fix:** arm a write-branch generation/flag in executeLevel2 under the
mutex at entry and refuse a second entry while one is in flight — the
rungMu TryLock precedent (skip with a D-20 reason, e.g.
`correction-busy`, via skipCorrection + settleCombo). The flag must be
set before the first forward and cleared on every return path
(including the context-changed branch).

### WR-02: VerifyExpiry carries no per-round tag — a stale deadline can consume the next pending round and falsely teach verify-dead

**File:** `internal/session/actor.go:1171-1191` (VerifyExpiry), arm sites `2971`, `3078`; contrast `1689-1705` (afterExpiry)
**Issue:** pendingFix rounds all share the bare `a.VerifyExpiry` callback
with no round tag, while the verify-after rounds solved exactly this race
with the epoch discipline (afterExpiry re-checks `a.after.epoch`). The
unguarded window: round 1's deadline fires and its callback blocks on
a.mu; the main path then resolves round 1 (Stop is a no-op — already
fired) and arms round 2; when the mutex frees, the stale VerifyExpiry
proceeds, sees `a.pending != nil` (round 2), drops it via resolvePending,
records a verify-timeout skip — and `setVerifyDead(true)` teaches the
behavioral terminal class for a context that never missed a budget. The
correction is lost and the context is misrouted to the terminal write
branch until the next FocusOut. The corpus pins the settled-round case
(TestTerminalClass_StaleVerifyExpiryLearnsNothing — pending == nil) but
not the new-round case. Narrow window, real consequence, and the
codebase's own epoch pattern shows the race class is treated as real.
**Fix:** tag pendingFix with a monotonic epoch (the verifyEpoch
precedent): capture `epoch := a.pendingEpoch` in the AfterFunc closure,
pass it to VerifyExpiry (or a wrapped callback), and no-op unless
`a.pending != nil && a.pending.epoch == epoch` — the afterExpiry form,
byte for byte.

## Info

### IN-01: caseListUsage omits term-negative-witnesses while claiming "every registry name"

**File:** `test/e2e/main.go:113-128` (caseListUsage) vs `339` (registry), `276` (error text)
**Issue:** the `-case` help text lists term-word-vte and
term-word-wezterm but not term-negative-witnesses, although the registry
runs it and the unknown-case error names it. A stand operator reading
only the usage string misses the negative-witness case.
**Fix:** add `" | term-negative-witnesses"` to caseListUsage (or derive
the usage from caseRegistry() so the two cannot drift).

### IN-02: Save/Primary discard the wl-paste stderr diagnostic on the empty-classification path

**File:** `internal/clipboard/clipboard.go:85-97` (Save), `108-120` (Primary)
**Issue:** a genuine wl-paste exit 1 (empty clipboard/selection) is a
state, not an error — correct — but wl-paste also exits 1 with a stderr
message for genuine failures short of a signal kill (connection or
service trouble). execRunner embeds the stderr in the wrapped error, and
the empty-classification branch then drops the whole error: had=false
with no trace of why. Diagnosability cost only (content loss requires
the paired wl-copy to succeed where wl-paste failed — not a realistic
chain), and the CR-03 signal-kill guard is correctly pinned.
**Fix:** on the empty-classification branch, emit a Debug record with the
swallowed error (it carries exit status + stderr text, never clipboard
content).

### IN-03: boundRowText truncates by bytes and can split a UTF-8 rune in the journal row

**File:** `test/e2e/case_term_word_vte.go:66-72`
**Issue:** the 200-byte cap slices the error text at a byte offset, so a
multi-byte Cyrillic rune at the boundary lands as an invalid sequence in
the journal's observables cell (cosmetic; the row remains readable up to
the cut).
**Fix:** back off the cut to a rune boundary, e.g.
`for len(text) > termRowTextBound && !utf8.RuneStart(text[termRowTextBound]) { ... }`
or `strings.ToValidUTF8(text[:termRowTextBound+4], "")`.

---

_Considered, no finding_: stretch.go's range walk is guarded and clone-safe;
config.go's terminals grammar mirrors the blocklist precedents with the
ceiling inclusive and the empty-pattern trap refused before compile;
appid.go's registry interest fails Start honestly (a blind observer is
worse than the degradation WARN) and its run-loop cleanup runs exactly
once; the e2e cases keep the honest-abort exit contract with bounded
retries only where the plan prescribes a ladder (wezArmGateTries);
main.go's snapshot-before-reap reordering and the unit-restore name wait
close real stand races; the D7 contract boundary is documented in
docs/CONFIG.md in the owner's verbatim terms.

_Reviewed: 2026-10-12T21:56:44Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
