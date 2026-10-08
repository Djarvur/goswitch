---
phase: 06-avtokorrekcija-opcionalno
date: 2026-10-02T21:30:00Z
depth: standard
reviewer: gsd-code-reviewer
diff_base: e13af69..HEAD
files_reviewed: 24
findings:
  critical: 2
  warning: 1
  info: 4
  total: 7
status: findings
---

# Phase 6: Code Review Report

**Reviewed:** 2026-10-02T21:30:00Z
**Depth:** standard (per-file analysis, Go-specific checks, cross-file call-chain tracing of the D-53 contour)
**Commits:** e13af69..HEAD (branch gsd/phase-06-avtokorrekcija-opcionalno)
**Files Reviewed:** 24 (list below)
**Status:** findings — 2 Critical, 1 Warning, 4 Info
**Gates re-run during review:** `go build ./...`, `go vet ./...`, `go test -race -count=1` on internal/{detect,correct,config,session,ctlsvc,appid} + layouts — all green.

## Summary

Phase 6 adds the opt-in autocorrect layer (detector 06-05, appid Role seam 06-03, config section 06-04, actor integration 06-06, live e2e 06-07, acceptance 06-08). The bulk of the submission is high quality: `internal/detect` is a genuinely pure package (no D-Bus/engine imports, closed reason vocabulary, token never enters a Verdict — D-20/D-21 hold on every new log/status site; the only content-bearing log lines are the pre-existing, documented `-debug` records of phase 3). The D-53 cheap-half gates are correctly fail-closed and inverted from the MACR fail-open direction; the off state is zero-behavior/zero-counter; `Buffer.PushFeed` is an honest additive flag; config defaults are off with strict validation; the generated dictionaries carry license headers, sortedness invariants, and deterministic regeneration; the e2e negatives use three independent oracles with missing-witness = FAIL.

The two Critical findings are both in the same new structural element: the **asynchronous confirm** (`autoConfirm`) executes a delete+commit pipeline against actor state that is NOT re-validated at execution time. The generation counter — the only staleness guard — advances exclusively at boundary arming, so every non-boundary mutation of the field/buffer between the boundary keystroke and the confirm (an interleaved keystroke, Backspace, the reset key's own client-side effect) leaves the payload "fresh" by its reckoning. The manual correction pipeline never had this exposure because a human-paced gesture cannot land inside a keystroke's in-flight window; the async confirm can, by construction. Fixes are small and local (revalidation under the mutex; skipping the cached-push shortcut for boundary-armed payloads), and the 06-07 pre-boundary-answer machinery (cbfa46c) already provides most of the needed mechanism.

## Critical Issues

### CR-01: Async autocorrect confirm executes against moved-on field/buffer state — wrong-range delete can corrupt the typed text — FIXED (see Fix Log)

**File:** `internal/session/actor.go:1789-1817` (`armAutoCorrect`), `:1935-2002` (`autoConfirm`), `:2200-2203` (cached-push fast path), `:2232` (`p.rng.replace`)

**Issue:** `acGeneration` is bumped only in `armAutoCorrect` (line 1794), i.e. only at boundary arming. Between the boundary keystroke and the confirm's final mutex section (role RTT, budget 25 ms), any of the following mutates state WITHOUT advancing the generation: a token-capable keystroke (`feedKey` → `pushAndArm`), a Backspace (line 1724), a HardReset (line 1739), a FocusOut lifecycle (`HandleLifecycle` clears `a.surr` but does not touch `acGeneration`). The confirm then runs `startRangeCorrection(payload.rng)` on the stale payload:

- **Wrong-range delete (field corruption).** Interleaving: separator armed at t0; the confirm's Role RTT is in flight; a token-capable key `x` is processed by the daemon at t0+δ (δ < RTT) and transits to the client; the confirm acquires the mutex at t0+RTT while the client's `SetSurroundingText` push for `x` is still in flight. `a.surr` still ends with `"ghbdtn "`, so the cached-push fast path (`startRangeCorrection`, line 2200) matches `match = token+tail` and `executeCorrection` immediately issues `DeleteSurroundingText(-(6+1), 7)` against a client cursor that has already moved past `x`. The delete consumes the wrong 7 runes (the tail of the word plus `x`), the commit then inserts `"привет "` — and the verify-after suffix check **passes**, because the commit landed after the wrong delete. The field is silently garbled; nothing counts it. The manual Double path cannot hit this window (a gesture and a keystroke cannot interleave at human pace); the async confirm is the first construct in the daemon that can.
- **Buffer-mirror corruption.** Even when the verify round refuses the field edit, `executeCorrection:2232` calls `p.rng.replace(p.converted)` — `a.buf.ReplaceToken` captured at arm time but executed at confirm time against the buffer's **current** `activeRange`. After an interleaved keystroke the current token is the NEW word (`x`), which gets overwritten with `"привет"`; the mirror no longer matches the field, so subsequent manual toggles on that word verify-fail (fail-safe, but the manual override feature is dead for the word) until the next HardReset/recompute.

**Fix:** re-validate the armed payload against live state under the mutex in `autoConfirm`'s final section, before the fired branch — the data is already in the payload:

```go
a.mu.Lock()
defer a.mu.Unlock()
if payload.gen != a.acGeneration {
    return
}
// The field must still hold exactly what was armed: a keystroke, a
// backspace, a reset or a focus change since the boundary invalidates
// the payload (T-06-06-06 extended to non-boundary mutations).
if !slices.Equal(a.buf.Token(), payload.word) ||
    !slices.Equal(a.buf.Tail(), payload.rng.tail) {
    a.recordACAbstain(acReasonPayloadStale) // new closed slug
    return
}
```

(Equivalently: bump `acGeneration` on every buffer mutation — Backspace, HardReset, and token-capable `pushAndArm` — so the existing stale check covers the whole class.)

### CR-02: Reset-boundary arming (Enter/Tab/Escape) races the reset key's own client-side effect — the correction can land in the wrong input context — FIXED (see Fix Log)

**File:** `internal/session/actor.go:1727-1741` (arming site), `:2152-2207` (`startRangeCorrection`), `:2200-2203` (cached-push fast path)

**Issue:** the reset boundary arms BEFORE `HardReset` (line 1738) with `rng.tail = nil` and `match = token`. The arming keystroke is itself an in-flight field-mutating/focus-moving key: for Tab, GTK moves focus to the NEXT entry; for Enter/Escape, single-line fields submit/clear. The async confirm fires ~one Role-RTT later. Two failure shapes:

1. The cached-push fast path (line 2200) compares against the PRE-reset push (`surr` ends with the raw word — the last text push before the reset), matches, and fires the delete+commit without any Require round. The delete/commit signals travel daemon → ibus-daemon → client and are routed to the input context ibus-daemon considers focused **at forwarding time** — which races the client's own focus-change notification (a different connection; ordering is not guaranteed, and the 06-07 terminal re-pin 1dc76af already documents that IME delivery races on this desktop are real). When the routing lands post-Tab, `DeleteSurroundingText(-n, n)` executes against the NEW field's cursor — deleting its existing text — and `CommitText` inserts the converted word into it. A wrong-context write, the exact failure class the D-53 care exists to prevent.
2. Even when routing stays in the old context, in single-line fields the Enter/Escape has already cleared/submitted the entry, so the commit resurrects the word as ghost text in a consumed dialog.

The phase's own records concede this surface was never validated live: 06-06 pins only that the reset boundary ARMS ("живой эффект не несущая часть приёмки, план-допущение"), and the e2e matrix exercises only the space boundary. For the space boundary the race is benign — the separator is a same-field text edit and the wire ordering `[separator transit][delete][commit]` keeps the geometry right. For Tab/Enter/Escape it is not.

**Fix:** never take the cached-push shortcut for a boundary-armed payload — the field is mid-keystroke by construction, which is exactly what the pre-correction verify exists for (ADR-004):

```go
// in startRangeCorrection, before the cached-push fast path:
if rng.acBoundary || !correct.MatchesSuffix(a.surr, a.pending.match) {
    a.pending.deadline = time.AfterFunc(a.verifyWait, a.VerifyExpiry)
    a.eng.RequireSurroundingText()

    return
}
```

The 06-07 machinery (cbfa46c: pre-boundary early answer keeps the round open for the post-boundary push) already handles the space case through this route, and a post-Tab Require answer names the NEW field → mismatch → fail-closed skip. Consider additionally excluding Tab/Escape from the boundary set until the reset-boundary behavior is proven live (owner decision — the plan accepted the risk, this review flags that the acceptance never covered it).

## Warnings

### WR-01: D-53 conjunction is evaluated non-atomically — white-list app identity is arm-time, role is confirm-time, no focus-change guard — FIXED (see Fix Log)

**File:** `internal/session/actor.go:1826-1836` (app gate at boundary), `:1946-1952` (role query at confirm against the observer's current focus pair)

**Issue:** condition 1 (app in white-list) is evaluated at boundary arming via `acFocusedApp()`; condition 2 (live role) is evaluated ≤25 ms later at confirm time, querying `GetRole` on the observer's **current** `(sender, path)` pair — which a focus change inside the window re-points to a different object of a different app. A focus switch inside the window therefore pairs app A's white-list verdict with app B's role: D-53's "конъюнкция В МОМЕНТ решения" is not evaluated for one object at one instant, and no counter or generation advances on focus change. Practical damage is bounded (the verify round fails in the new field → silence), which is why this is a Warning and not a Critical — but the policy gap is provable from the code, and the fix is one cached read.

**Fix:** store the resolved app in `acPayload` and re-check under the mutex in `autoConfirm`'s final section: `if app, ok := a.acFocusedApp(); !ok || app != payload.app { a.recordACAbstain(acReasonAppUnknown); return }` — or fold into CR-01's revalidation block.

## Info

### IN-01: `acFired` counted before the pipeline can refuse

**File:** `internal/session/actor.go:1976-1980`

**Issue:** `autoConfirm` increments `acFired` and logs `"reason":"fired"` before `startRangeCorrection`, whose own refusal paths (`empty-buffer`, `convert-failed`, `no-engine`) then leave `autocorrect_fired` overstated alongside a manual-skip record. The detector gates exclude the first two on this path; `no-engine` is reachable only if the engine detaches between the role gate and the confirm. Counter semantics only — no behavioral effect.

**Fix:** move the fired counting/log into `executeCorrection` under `rng.acBoundary`, or have `startRangeCorrection` report refusal and un-count.

### IN-02: Per-word INFO journal records while the layer is enabled

**File:** `internal/session/actor.go:1910-1918` (`recordACAbstain`)

**Issue:** with autocorrect enabled, EVERY word boundary outside a white-listed app records `slog.Info("autocorrect skipped", ...)` (slug `app-not-listed` — the common case). A typist at 300+ words/minute writes hundreds of INFO journal lines per minute for a feature that, for them, does nothing. Sanctioned by plan 06-06 (the skipCorrection form), and off-state is clean — but the operational noise scales with typing, not with events.

**Fix:** demote the per-boundary abstention records to `slog.Debug`; keep INFO for `fired` and for the one-WARN degradation episodes.

### IN-03: Godbus pending-call accumulation when the a11y bus wedges

**File:** `internal/appid/appid.go:222-232` (`Role`), `:216-227` (`Start` role closure)

**Issue:** `CallWithContext` under the 25 ms deadline returns on `ctx.Done()`, but godbus keeps the abandoned call in its pending map until a reply arrives or the connection dies. The wedge scenario the deadline exists for (a stuck a11y object) plus active typing therefore accumulates in-flight call entries for the session's duration. Bounded and small (hundreds of bytes each), and there is no goroutine leak — the dispatch goroutine is shared. Out of v1 scope (resource growth), recorded for the follow-up.

**Fix:** a cheap circuit breaker in the actor — after K consecutive `role-timeout` abstentions, stop querying until the next focus gain (the warn-once episode machinery is the natural home).

### IN-04: `waitFixtureExit` timeout path races concurrent `exec.Cmd.Wait`

**File:** `test/e2e/case_autocorrect.go:291-314`

**Issue:** the happy path joins the goroutine's `Wait` via `done`; the timeout path calls `reapPasswordFixture()` (which calls `Wait` again) while the goroutine's `Wait` is still blocked — `exec.Cmd.Wait` is not safe for concurrent callers. Test-only, already-failing path (the case FAILS either way); no e2e verdict is affected.

**Fix:** guard the reap with the same `done` channel (kill only, let the single `Wait` own the reap), or have the goroutine be the only `Wait` caller.

## Checked dimensions and verdicts

- **D-53 fail-closed conjunction (cheap half):** CLEAN — off = zero behavior/zero counters (byte-as-today boundary); empty list, missing caps bit, unknown app, unlisted app (exact equality, no prefix merging), unsure/short detector verdicts each refuse with exactly one closed slug; silence matrix + fail-closed-no-fail-open corpus pins it.
- **D-53 role contour (expensive half):** two findings — CR-01/CR-02 (stale-state execution class) and WR-01 (conjunction split). The role values themselves are correct (numeric comparison only, 40/60 refused, 61/79/94 allowed, errors/timeouts fail closed, warn-once per episode with healthy-answer reset).
- **Privacy D-20/D-21:** CLEAN — grep over the full phase diff: every new log/status/error site carries slugs, counts, roles, or transport errors only; `detect.Verdict` is structurally token-free; the unit corpus asserts non-leakage of the word; the only content-bearing records (`slog.Debug("surrounding push", "text", ...)`, `logCorrectionDone` DEBUG source/result) are pre-phase-6 and covered by the documented `-debug` carve-out.
- **`internal/detect` purity and hot path:** CLEAN — imports are `sort`/`strings`/`unicode` + `internal/correct` only; no clock, no goroutines, no D-Bus; `PushFeed` is allocation-free; `Token()/Tail()` clone only on armed boundaries; dictionary lookups are bounded binary searches; veto ordering (both-hit before cur-hit) is safe.
- **Concurrency:** CLEAN apart from CR-01 — every goroutine exits (ctx-honoring Role, `defer cancel()` timer hygiene, generation guard); role call strictly off-mutex (fakeRole started/release double proves HandleKey liveness); engine emits are fire-and-forget so no blocking under the mutex; no new shared state without synchronization (`fakeRole`, observer pair copies under its mutex).
- **Wire discipline of the GetRole seam (06-03):** CLEAN — `GetRole` as the full wire literal, uint32 `Store`, `ErrRoleUnknown` sentinel vs `%w`-wrapped transport errors, `errors.Is(err, context.DeadlineExceeded)` routing to `role-timeout` survives the wrap chain, ctx propagated end-to-end, focus-loss keeps the stale pair whose cross-app collision is impossible (sender-scoped addressing) and whose dead-object case fails closed.
- **Config (D-54/D-31..D-33):** CLEAN — defaults off; strict decode; ceiling 64 unconditional; threshold ranges bite only on the activatable shape, so a dormant section cannot smuggle garbage into an `enabled: true` toggle (whole-document re-validation blocks it); `Load` decodes into the zero Config, and the validation gate keeps partial enabled sections out.
- **Status/ctl surface:** CLEAN — map cloned at snapshot (mutation-safe), nil-map safe render, sorted `ac_skip_*` slugs before the config block, `config_error` last; CLI wiring untouched.
- **Golden data + dictgen:** CLEAN — loud hard errors (count mismatch, unsorted, empty), byte-deterministic emission (alphabetical pre-sort + stable count sort + sorted map keys), license notices in headers, golden invariants pin sortedness/disjoint scripts/no-yo/envelopes/non-positive trigrams.
- **e2e drivers:** CLEAN apart from IN-04 — three independent oracles per negative, missing witness = FAIL not skip, observed (never assumed) bridge identity, positive control in the terminal case, honest matrix v4 superset with v3 frozen.
- **Mise/workflow/nightly script:** CLEAN — v4 default in both D-48 forks with v1/v2/v3 regression paths preserved, dispatch updated consistently, expression-injection-safe env indirection preserved.

## Files Reviewed

- internal/session/actor.go
- internal/detect/detect.go
- internal/correct/buffer.go
- internal/appid/appid.go
- internal/config/config.go
- internal/ctlsvc/ctlsvc.go
- layouts/dictgen/main.go
- layouts/dict_ru.go (generated; header/invariant spot-check)
- layouts/dict_en.go (generated; header/invariant spot-check)
- layouts/trigrams.go (generated; header/invariant spot-check)
- internal/session/actor_test.go (autocorrect corpus + fakeRole double)
- internal/detect/detect_test.go (structure)
- internal/correct/buffer_test.go
- internal/config/config_test.go, load_test.go, watch_test.go (coverage survey)
- layouts/dict_test.go (invariants)
- test/e2e/case_autocorrect.go
- test/e2e/main.go
- test/e2e/matrix.go
- test/e2e/perf.go
- test/e2e/case_ctl.go
- test/e2e/fixtures/password_entry.py
- test/e2e/cases/matrix-v4.yaml (autocorrect block)
- mise.toml, .github/workflows/e2e-matrix.yml, scripts/d48-nightly-dispatch.sh

## Fix Log

_Iteration 1 — gsd-code-fixer, 2026-10-02T23:05:00Z. Scope: Critical + Warning (the 4 Info findings stay documented above, unfixed). Full report: `06-REVIEW-FIX.md`. All gates ran in the isolated review-fix worktree and were fast-forwarded to this branch (tip `48013f2`); reproducible there._

| ID | Fix mechanism | RED commit | GREEN commit | Re-test verdict |
|---|---|---|---|---|
| CR-01 | Content re-validation under the mutex at confirm: `acPayload.expectToken/expectTail` snapshot the buffer as the arming event leaves it (reset branch captures pre-`HardReset`, launches post-reset); the fired branch refuses on `!slices.Equal(buf.Token(), expectToken) || !slices.Equal(buf.Tail(), expectTail)` with the new closed slug `payload-stale`. One mechanism (content), not generation bumping. | `3e64e95` `test(06): RED — the async confirm fires the armed payload against a moved-on buffer — CR-01` | `d5f11e4` `fix(06): re-validate the armed payload under the mutex at confirm — CR-01` | RED demonstrated (cached fast path fired delete+commit against the moved-on buffer, `fired=1`); GREEN: `TestAutoConfirm_RevalidatesArmedPayload` passes under `-race`, full session+correct corpus green; counter semantics of the existing `role-timeout` corpus preserved (the check sits inside the allowed-role case). Logic flagged for human verification per fixer contract (async race window). |
| CR-02 | Always-verify: `startRangeCorrection` never takes the cached-push shortcut for `rng.acBoundary` payloads (`if rng.acBoundary || !MatchesSuffix(...)` → Require + verify budget). The cbfa46c early-answer machinery decides the space case; a post-reset answer names the field as the reset key left it → fail-closed `verify-mismatch`. Tab/Escape stay in the boundary set (owner decision, untouched). Manual paths byte-unchanged. | `92991c8` `test(06): RED — the reset-boundary payload shortcuts the verify round on the pre-reset cache — CR-02` | `2d9ea85` `fix(06): never take the cached-push shortcut for boundary-armed payloads — CR-02` | RED demonstrated (2 correction ops with no verify round); GREEN: `TestAutoConfirm_ResetBoundaryAlwaysVerifies` passes under `-race` (Require goes out, post-reset push concludes verify-mismatch, zero field edits); full corpus green. |
| WR-01 | `acPayload.app` stores the bridge-namespace identity the white-list verdict was taken for; the confirm's fired threshold re-checks the live focused app under the mutex via `acFocusedApp()` (cached observer read, mutex-safe): `!ok` → `app-unknown`, `app != payload.app` → new closed slug `app-changed` — both fail-closed silence + skip counter. | `802de00` `test(06): RED — the confirm pairs the armed app's white-list verdict with another app's role — WR-01` | `7c8d3cc` `fix(06): re-check the focused app identity under the mutex at confirm — WR-01` | RED demonstrated (confirm fired with gedit's verdict + Zenity's role); GREEN: `TestAutoConfirm_RechecksFocusedApp` passes under `-race`, full corpus green. Deviation (documented in 06-REVIEW-FIX.md): the mismatch counts the honest `app-changed` slug instead of overloading `app-unknown` — same fail-closed semantics. |

Follow-up `48013f2` `refactor(06)`: lint-gate cleanup after all three fixes — the two confirm re-checks extracted into `acConfirmRefusals` (`funlen` 44 > 40) plus one corpus line wrap (`lll`). Behavior-neutral; the three review-fix tests re-run green after it.

**Gates at iteration 1 close:** `mise exec -- go test ./internal/session/ ./internal/correct/ -race -count=1` green; full `mise run ci` (build + vet + golangci-lint v2 + `go test -race -count=1 ./...`) green — the known flake `TestRun_OnConnHookCalledOnce` did not fire.

---

_Reviewed: 2026-10-02T21:30:00Z_
_Reviewer: Claude (gsd-code-reviewer), go-ultimate strict discipline_
_Depth: standard_
