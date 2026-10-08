---
phase: "06"
slug: "avtokorrekcija-opcionalno"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-28"
validated: "2026-10-04"
---

# Phase 06 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Source: 06-RESEARCH.md § Validation Architecture (Test Framework, Phase Requirements → Test Map,
> Sampling Rate, Wave 0 Gaps). Таблицы ниже заполняются планировщиком/validate-phase.
> **Validated 2026-10-04 (nyquist audit):** every phase requirement maps to at least one named
> green carrier (unit / live e2e / matrix row / regression gate); the only non-carriers
> (word-mixed/phrase-mixed target semantics) are owner-deferred v1.1.x follow-ups, not phase gaps.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib testing + -race, mise tasks) |
| **Config file** | mise.toml |
| **Quick run command** | `mise exec -- go test ./... -race -count=1` |
| **Full suite command** | `mise run ci` |
| **Estimated runtime** | ~120 seconds |

---

## Sampling Rate

- **After every task commit:** Run `mise exec -- go test ./... -race -count=1`
- **After every plan wave:** Run `mise run ci`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | SPEC §10/§11 (spec-delta) | T-06-02a | audit-trail comments keep the §11 verdict visible | doc greps | `grep -q "v1.1.0" docs/SPEC.md && grep -q "2026-09-27" docs/SPEC.md && …` | ✅ docs/SPEC.md | ✅ green (SPEC-DELTA-OK + AUDIT-TRAIL-OK re-run 2026-10-04) |
| 06-01-02 | 01 | 1 | SPEC §10/§11, MACR-ACL | T-06-01a | GPL read-only position, fail-closed recorded | doc greps | `test -f docs/adr/ADR-007-*.md && grep -q "## Status" …` | ✅ docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md | ✅ green (ADR-007 **Accepted** by owner 2026-10-04, UAT test 2) |
| 06-02-01 | 02 | 2 | CORR-08 | T-06-02a/b | generator hard errors + golden invariants (sorted/disjoint/envelope) | unit+golden | `mise exec -- go test ./layouts/ -race -count=1` | ✅ layouts/dictgen/main_test.go, layouts/dict_test.go | ✅ green (layouts ok under -race, re-run 2026-10-04) |
| 06-02-02 | 02 | 2 | CORR-08 | T-06-02b | table size caps 4096/lang, deterministic | unit+golden | `mise exec -- go test ./layouts/ -race -count=1 && mise run dictgen-regen && git diff --exit-code -- layouts/` | ✅ layouts/dict_test.go | ✅ green (TestGolden_TrigramsBounded, REGEN-DETERMINISTIC) |
| 06-03-01 | 03 | 2 | MACR-ACL | T-06-03a | (sender,path) cache-keep matches FocusedApp semantics | unit | `mise exec -- go test ./internal/appid/ -race -count=1` | ✅ appid_test.go | ✅ green (appid ok under -race, re-run 2026-10-04) |
| 06-03-02 | 03 | 2 | MACR-ACL | T-06-03a/b | live GetRole, ErrRoleUnknown sentinel, never GetRoleName | unit | `mise exec -- go test ./internal/appid/ -race -count=1` | ✅ appid_test.go | ✅ green (TestRole_*, ROLE-SEAM-OK) |
| 06-04-01 | 04 | 2 | MACR-ACL | T-06-04a/b/c | default-off everywhere, ceilings, legacy compat | unit | `mise exec -- go test ./internal/config/ -race -count=1` | ✅ config_test.go | ✅ green (TestDefaults_AutocorrectOff, TestValidate_AutocorrectAppsCeil) |
| 06-04-02 | 04 | 2 | MACR-ACL | T-06-04a | hot-reload propagation + last-good | unit | `mise exec -- go test ./internal/config/ -race -count=1` + CONFIG-DOC greps | ✅ watch_test.go | ✅ green (TestWatch_BrokenAutocorrectBlockKeepsLastGood; load.go/watch.go zero-diff pin) |
| 06-05-01 | 05 | 3 | CORR-04/05/06 | T-06-05a/b | dictionary-hit veto, closed reason vocabulary | unit (golden corpus) | `mise exec -- go test ./internal/detect/ -race -count=1` | ✅ detect_test.go | ✅ green (TestCheck_DictCurHitVetoes/BothHitVetoes, TestCorpus_ReasonVocabularyClosed) |
| 06-05-02 | 05 | 3 | CORR-04 | T-06-05b | OOV fires via trigram; legit-word budget exactly 0 | unit (golden corpus) | `mise exec -- go test ./internal/detect/ -race -count=1` | ✅ detect_test.go | ✅ green (corpus 190/0/34: TestCorpus_WrongLayoutBothDirections 190, LegitWords 0-false on 404, OOVNames 34) |
| 06-05-03 | 05 | 3 | CORR-04/06 | T-06-05a/c | empty/tie/ordering boundary answers pinned | unit | `mise exec -- go test ./internal/detect/ -race -count=1 && mise run ci` | ✅ detect_test.go | ✅ green (TestCheck_Abstain*, TestCorpus_MixedTokens) |
| 06-06-01 | 06 | 4 | CORR-01/07/09, MACR-ACL | T-06-06a/c/d | D-53 conjunction fail-closed; silence matrix; existing pipeline only | unit (fakes) | `mise exec -- go test ./internal/correct/ ./internal/session/ -race -count=1` | ✅ actor_test.go | ✅ green (TestAutoCorrect_SilenceMatrix 10 cells + FailClosedNoFailOpen + SINGLE-PATH-OK; session ok under -race 2026-10-04) |
| 06-06-02 | 06 | 4 | CORR-01, MACR-ACL | T-06-06b | word-free records at all levels; status counters | unit | `mise exec -- go test ./internal/session/ ./internal/ctlsvc/ -race -count=1` | ✅ ctlsvc_test.go | ✅ green (TestAutoCorrect_CountersAndReasons, TestRenderStatus_AutocorrectTokens, TestCtlStatus_EndToEnd) |
| 06-06-03 | 06 | 4 | MACR-ACL | T-06-SC | wiring, zero new deps (tidy-diff) | gate | `mise run ci` | ✅ | ✅ green (`mise run ci` green, 17–18 packages, verified this session) |
| 06-07-01 | 07 | 5 | MACR-ACL | T-06-07a | GTK4 fixture + live Chromium role pin | py_compile + greps | `python3 -m py_compile test/e2e/fixtures/password_entry.py && …` | ✅ test/e2e/fixtures/password_entry.py | ✅ green (FIXTURES-OK; live witness role=PASSWORD_TEXT(40) 0→7 chars) |
| 06-07-02 | 07 | 5 | MACR-ACL, CORR-01 | T-06-07a/b/c | password-silent + terminal-silent (stale-cache) live | e2e (live) | `mise run e2e-autocorrect-fires` / `-password-silent` / `-terminal-silent` | ✅ test/e2e/case_autocorrect.go | ✅ green (all three PASS live 2026-10-04, UAT test 4) |
| 06-07-03 | 07 | 5 | SPEC §10/§11 | T-06-07d | matrix v4 rows; v3 frozen | e2e (live) | `mise run e2e-matrix-v4` + v3 diff gate | ✅ test/e2e/cases/matrix-v4.yaml | ✅ green (v3 byte-frozen gate green; v4 32/35 ×2 at owner UAT — see Requirements Coverage note) |
| 06-08-01 | 08 | 6 | SPEC §10/§11 | T-06-08b | perf p95 delta ≤ 5 ms (Pitfall 5) | e2e (live) | `mise run e2e-perf-autocorrect` + ADR greps | ✅ test/e2e/perf.go | ✅ green (delta p95 −3.3 ms: 180.2 → 176.9 ms, same session; VmHWM 19.1 МБ < 50) |
| 06-08-02 | 08 | 6 | CORR-01..09 (base regression) | T-06-08a/b | double-run gate; README safety caveats | e2e (live) | `mise run e2e-matrix-v4 ×2 && mise run e2e-autocorrect ×2` + README greps | ✅ README.md | ✅ green (double-run done; owner verdict 2026-10-04: 32/35 ×2 stable post G-6-1 fix, drift rows green; FINALIZE-GREPS-OK) |
| 06-09-01 | 09 | 7 | SWCH-01/SWCH-02 (gap G-6-1) | — | lock-free AttachEngine: factory answers CreateEngine during flip | unit | `mise exec -- go test ./internal/session/ ./engine/ -race -count=1` | ✅ actor_test.go, conn_switcher_test.go | ✅ green (TestActor_ReentrantAttachDuringFlip + TestFactoryReentrantCreateEngine* wire witnesses, -race) |
| 06-10-01 | 10 | 8 | SWCH-01/SWCH-02 (gap G-6-1) | — | first keystroke after daemon flip reaches the field | e2e (live) | `mise run e2e-flip-keystroke` | ✅ test/e2e/case_switch.go | ✅ green (24/24 immediate letters over 4 runs; RTT 6.5–9.2 ms; two-source-flip regression PASS) |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Requirements Coverage (Nyquist Map)

Validated 2026-10-04. Every phase requirement carries at least one named green carrier.
CORR-01..09 are regression-guarded by the v3-frozen rows embedded byte-verbatim in matrix-v4
(git diff gate on matrix-v3.yaml) plus the unit pins listed — per the flagged-assumptions
reading in the plans.

| Requirement | Green Carrier(s) | Type | Status |
|-------------|------------------|------|--------|
| SPEC §10/§11 (spec-delta) | 06-01 doc gates SPEC-DELTA-OK / AUDIT-TRAIL-OK (re-run 2026-10-04); ADR-007 **Accepted** by owner 2026-10-04 (UAT test 2) with full Acceptance Evidence table | doc grep + owner gate | green |
| CORR-01 (word correction, manual override) | unit: `TestActor_ManualOverrideAfterAutocorrect`, `TestAutoCorrect_FiresThroughPipeline` (internal/session); live: `mise run e2e-autocorrect-fires` PASS (manual Double regression on same surface); matrix-v4 rows: `gedit-word-en-ru`, `x11-word-en-ru`, `autocorrect-fires` (green ×2) | unit + live e2e + matrix | green |
| CORR-02 (phrase, triple tap) | unit: internal/correct phrase corpus; matrix-v4 rows: `phrase-en-ru`, `phrase-ru-en`, `gedit-phrase`, `gedit-phrase-registers`, `phrase-registers`, `phrase-chromium`, `phrase-gte`, `x11-phrase`, `long-phrase`, `phrase-after-space` — all green ×2 in the 06-08 double run and post G-6-1 fix (UAT test 1: all 9 drift rows green) | unit + matrix | green |
| CORR-03 (selection) | unit: internal/correct selection/anchor corpus; matrix-v4 rows: `select-all-zenity`, `select-all-gte`, `select-all-chromium`, `select-partial-gte`, `select-reverse-gte` — green post G-6-1 fix (were in the drift set, all 9 drift rows green at owner UAT) | unit + matrix | green |
| CORR-04 (auto direction) | unit: `TestCheck_DirMatchesCorrect`, `TestCheck_DictWrongLayout`, `TestCorpus_WrongLayoutBothDirections` (190, both directions); live: e2e-autocorrect-fires (direction picked automatically) | unit (golden corpus) + live e2e | green |
| CORR-05 (case preserved) | unit: `TestCorpus_WrongLayoutBothDirections` (three registers per word); matrix-v4 rows: `gedit-word-capital`, `gedit-word-upper`, `x11-word-capital`, `x11-word-upper` (green ×2) | unit (golden corpus) + matrix | green |
| CORR-06 (mixed text) | unit pins: `TestCorpus_MixedTokens` + `TestCheck_AbstainConvertFail` (mixed → abstain, never partial edit). Target per-char-inversion semantics: rows `word-mixed`/`phrase-mixed` — **owner-deferred v1.1.x follow-up** (verdict 2026-10-03: per-char layout inversion; WINDOWS #12 freeze rejected), NOT a phase gap | unit pin + deferred follow-up | green (phase scope) / follow-up recorded |
| CORR-07 (exact range, ladder) | unit: internal/correct ladder corpus + SINGLE-PATH-OK grep pin (CommitText sites == 5, no second mechanism); live: e2e-autocorrect-fires (pid-read verifies exact-range correction via existing pipeline); matrix-v4 rows: `x11-word-*` (fallback ladder on Chromium-X11) | unit + live e2e + matrix | green |
| CORR-08 (generated tables + dict data) | unit: `layouts/tables_test.go#TestGolden_SpecExamples/TableSize`, `layouts/dict_test.go#TestGolden_DictSorted/Disjoint/NoYo/Envelope/TrigramsBounded`, dictgen corpus; gate: `mise run dictgen-regen && git diff --exit-code` (REGEN-DETERMINISTIC); layouts green under -race re-run 2026-10-04 | unit+golden+gate | green |
| CORR-09 (buffer reset triggers) | unit: `TestActor_ResetKeyBoundary`, `TestActor_BackspaceNoBoundary`, `TestActor_FocusOutNoBoundary`, `TestBuffer_PushFeed`; matrix-v4 rows: `flip-after-word-correction`, `ctl-status-reload` (green ×2) | unit + matrix | green |
| MACR-ACL (app×role политика) | unit: internal/appid corpus (pair cache-keep, live GetRole shape, ErrRoleUnknown, enum pins 40/60/61/79/94, GetRoleName==0 grep); `TestAutoCorrect_SilenceMatrix` (10 fail-closed cells), `TestAutoCorrect_FailClosedNoFailOpen`, `TestAutoCorrect_RoleDeadlineBounded` (25 ms off-mutex), `TestWarnOncePerEpisode`; config default-off/ceilings (`TestDefaults_AutocorrectOff`, `TestValidate_AutocorrectAppsCeil`); live: `mise run e2e-autocorrect-password-silent` PASS (role 40, fired=0, witness 0→7 chars untouched) + `mise run e2e-autocorrect-terminal-silent` PASS (fail-closed app-not-listed) | unit + live e2e | green |
| SPEC criterion 2 (default off, counters, hot reload) | config corpus + `TestApplySnapshot_AutocorrectFold`, `TestStatus_AutocorrectFields`, `TestRenderStatus_AutocorrectTokens`; matrix-v4 row `autocorrect-off` (off = zero behavior, zero counters); `autocorrect-fires` row reads D-54 counters live | unit + matrix | green |
| SPEC criterion 4 (double Shift manual override) | `TestActor_ManualOverrideAfterAutocorrect`; e2e-autocorrect-fires pins manual Double converts привет back on the same surface; matrix row `flip-after-word-correction` green ×2 | unit + live e2e + matrix | green |
| SPEC criterion 5 (matrix v4 + nightly double) | `mise run e2e-matrix-v4` — 32/35 ×2 stable at owner UAT (2026-10-04, all 9 session-drift rows green after G-6-1 fix); `super-space-alive` retired by owner (1eda53d, 34 rows remain); nightly pipeline defaults to v4 (e2e-matrix.yml + d48-nightly-dispatch.sh). Remaining red = 2 designed rows (owner-deferred semantics follow-up) — accepted by owner verdict, UAT test 1 pass | live e2e + owner gate | green (with recorded follow-up) |
| SWCH-01/SWCH-02 (gap G-6-1 carriers) | unit: `TestActor_ReentrantAttachDuringFlip`, `TestActor_AttachEngineWhileFlipInFlight`, `TestFactoryReentrantCreateEngine*` (fake-bus wire witnesses, -race); live: `mise run e2e-flip-keystroke` 24/24 immediate letters, RTT 6.5–9.2 ms, 8/8 flips INFO-form; regression `mise run e2e-two-source-flip` PASS | unit + live e2e | green |

**Coverage note (WARNING-level caveats, not gaps):**
1. `word-mixed` / `phrase-mixed` matrix rows stay red by design until the v1.1.x follow-up plan
   re-pins them to the owner-decided per-char layout-inversion semantics (06-UAT.md Deferred
   Follow-Ups, 2026-10-03). The phase-scoped CORR-06 contract (mixed → abstain, never partial)
   is unit-pinned green.
2. `super-space-alive` was retired from matrix-v4 by owner verdict (commit 1eda53d,
   2026-10-04): two-source-era leftover; single-source desktop model per owner decision
   260930-toa; daemon-owned switching proven live by e2e-flip-keystroke.
3. Live matrix / live-case carriers require the owner's GNOME desktop (ydotool + input group);
   they are environment-gated, not headless-CI carriers — headless regression rides the unit
   corpora listed above.

---

## Wave 0 Requirements

From 06-RESEARCH.md § Wave 0 Gaps, assigned to plans:

- [x] `layouts/dictgen/` + golden dict/trigram data + layouts golden tests (06-02)
- [x] `internal/detect/` package + mandatory golden corpus — ё-words, OOV names, mixed, legit-word budget, privacy, boundary semantics (06-05)
- [x] `internal/appid` (sender,path) storage + Role + feed tests (06-03)
- [x] `internal/config` autocorrect section + validation + hot-reload propagation tests (06-04)
- [x] `internal/session` word-boundary hook + D-53 conjunction + counters (06-06)
- [x] `test/e2e/fixtures/password_entry.py` (prototype live-verified by research) (06-07)
- [x] `test/e2e/case_autocorrect.go` + registry names + mise tasks + matrix-v4 (06-07)

All Wave 0 gaps closed by their plans; corpus evidence re-verified 2026-10-04
(golden corpus 190/0/34; silence-matrix 10 cells + acConfirm + G-6-1 reentrancy under -race).

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| ADR-007 Proposed → Accepted | SPEC §10/§11 | owner gate (T-06-08-01) | **DONE 2026-10-04** — owner accepted on the Acceptance Evidence table (UAT test 2) |
| WINDOWS #13 fresh-session double-green verdict | SPEC criterion 5 | machine-checked only by D-48 fresh_session preflight on the owner's desktop | **DONE 2026-10-04** — owner accepted: 32/35 ×2 stable post G-6-1 fix, drift rows green (UAT test 1) |
| Judgment bans: audit-trail wording (§11), GPL purity of dict data, cache-not-basis (D-53), no fixture content leak | SPEC §10/§11, MACR-ACL | legal/documentation judgment | **DONE 2026-10-04** — all four confirmed by owner (UAT test 5) |
| README / docs/CONFIG.md user-documentation adequacy | SPEC §10/§11 | prose quality is not grep-checkable | README autocorrect section greps green (FINALIZE-GREPS-OK); owner UAT pass on phrasing |
| CR-01 async confirm-window live race | MACR-ACL, CORR-01 | sub-25 ms interleaving on real hardware | **DONE 2026-10-04** — 18/18 race probe clean (no field corruption, honest abstentions) + deterministic unit corpus under -race (UAT test 3) |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references (all Wave 0 items landed and green)
- [x] No watch-mode flags
- [x] Feedback latency < 120s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** validated 2026-10-04 (gsd-nyquist-auditor). `mise run ci` green (build+vet+
golangci-lint+test -race, 17–18 packages); every phase requirement mapped to a named green
carrier in the Requirements Coverage table; the only non-carriers are the two owner-deferred
word-mixed/phrase-mixed semantics rows, recorded as a v1.1.x follow-up — not phase gaps.
