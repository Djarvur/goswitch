---
phase: "3"
slug: "frazy-vydelenie-pereklyuchenie-i-konfiguratsiya"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-15"
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Audit filled 2026-09-15 (verify-work post-hook): every task's pinned tests exist in the tree,
> the full `-race` suite is green at HEAD 54ea460 (re-run twice this day), and the live matrix v2
> (21 cases) is green at HEAD on the owner's desktop.

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib, -race per D-08) |
| **Config file** | mise.toml (D-09 — mise is the only scenario interface) |
| **Quick run command** | `mise exec -- go test ./... -race -count=1` |
| **Full suite command** | `mise run ci` |
| **Live acceptance** | `mise run e2e-matrix-v2` (21 cases, fresh daemon per case, live GNOME session) |

## Sampling Rate

- **After every task commit:** Run `mise exec -- go test ./... -race -count=1`
- **After every plan wave:** Run `mise run ci`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 03-01 phrase conversion + mixed runs + backspace cap | 03-01 | 1 | CORR-02, CORR-06 | T-03-01-01..04 | verify-before-delete, D-27 cap, atomic run conversion, D-20/D-21 logs | unit (golden + actor) | `go test -race ./internal/correct ./internal/session -run 'Phrase|TripleTap|ConvertRuns|MixedWord|BackspaceCap'` | ✅ TestActor_TripleTapCorrectsPhrase, TestActor_PhraseVerifyMismatch, TestConvertRuns_GoldenCorpus, TestBuildPlan_BackspaceCap, TestActor_MixedWord* | ✅ green |
| 03-02 config model + watcher | 03-02 | 1 | CONF-01 | T-03-02-01..04, SC | strict decode, ceilings, last-good | unit | `go test -race ./internal/config -run 'Load|Validate|Watch'` | ✅ TestLoad_UnknownKeyRejected, TestLoad_RangeViolationRejected, TestValidate_TimeoutRanges, TestValidate_CorrectionRanges, TestWatch_InvalidRewriteKeepsLastGood, TestWatch_DebounceCoalesces | ✅ green |
| 03-03 selection correction + clipboard rung | 03-03 | 2 | CORR-03 | T-03-03-01..05 | range clamp, stdin-only clipboard, deadline-bounded subprocesses | unit | `go test -race ./internal/session ./internal/clipboard -run 'Selection|Clipboard'` | ✅ TestActor_DoubleTapSelectionCorrects, TestActor_SelectionLeftToRight, TestActor_SelectionMixedConverts, TestClipboard_SaveSetRestore, TestClipboard_Timeouts, TestClipboard_KilledNotEmpty | ✅ green |
| 03-04 hotkeys/timeouts wiring + hot reload consumption | 03-04 | 2 | SWCH-02, SWCH-04, CONF-02 | T-03-04-01..03 | atomic snapshot, precise combo binding, timer no-re-arm | unit | `go test -race ./internal/session ./internal/hotkey -run 'Combo|HotReload|FlipOnSingle|ConfigurableTapKey|ModifierUse'` | ✅ TestActor_ComboWordThenFlip, TestActor_ComboDoesNotFeedBuffer, TestActor_HotReloadWindowNewSeries, TestActor_HotReloadTapKey, TestActor_HotReloadVerifyWait, TestFSM_ConfigurableTapKey, TestActor_FlipOnSingle | ✅ green |
| 03-05 MACR interception + per-app identity | 03-05 | 3 | MACR-01 | T-03-05-01..04 | configured-set-only interception, degraded appid, no content in logs | unit | `go test -race ./internal/session ./internal/appid -run 'MACR|Appid'` | ✅ TestActor_MACRIntercepts, TestActor_MACRConsumedUpstream, TestActor_MACRDisabledByDefault, TestActor_MACRPerAppMatch, TestActor_MACRAppidDegradation, TestAppid_FocusedAppByFakeBus | ✅ green |
| 03-06 goswitchctl + ctlsvc | 03-06 | 3 | INST-02 | T-03-06-01..04 | primary-owner guard, recover shim, text-free status | unit (over a real private dbus-daemon) | `go test -race ./internal/ctlsvc -run 'Svc'` | ✅ TestSvc_Status, TestSvc_ReloadInvalidKeepsLastGood, TestSvc_CorrectNow, TestSvc_NameGuard, TestSvc_RecoverShim | ✅ green |
| 03-07 matrix v2 acceptance base | 03-07 | 4 | CORR-02/03/06, SWCH-01/02/03, CONF-02, MACR-01 (e2e) | T-03-07-01..03 | canonical key names, strict case decode, fresh-daemon isolation | unit (decoder) + live e2e | `go test -race ./test/e2e -run 'Matrix'` ; `mise run e2e-matrix-v2` | ✅ TestMatrixDecode, TestMatrixDecode_RejectsNew, TestMatrixKeyNames_Canonical; live 21/21 PASS at HEAD 54ea460 (2026-09-15, exit 0) | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

## Wave 0 Requirements

- [x] Existing infrastructure covers all phase requirements (go test + mise + test/e2e stand from Phases 1–2; no new infrastructure was required).

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| (none — live-desktop behaviors go through the e2e stand per project quality constraint; owner UAT items resolved 4/4 in 03-UAT.md) | | | |

*All phase behaviors have automated verification.*

## Validation Audit 2026-09-15

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

Notes: plan 03-02 cited shorthand test names (TestValidate_Ranges, TestWatch_IgnoresOtherFiles) — implemented
as TestValidate_TimeoutRanges/CorrectionRanges and TestWatch_IgnoresIrrelevantEvents; same behaviors, actual
names recorded in the map. Requirements coverage: all 12 phase-3 IDs (CORR-02/03/06, SWCH-01..04, CONF-01..03,
MACR-01, INST-02) map to green automated tests above (CONF-03 is a documentation deliverable — docs/CONFIG.md
Caramba table, verified by the phase verifier — with its behavior exercised live by the 03-07 reload rows).

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** validated 2026-09-15 (verify-work post-hook audit; full -race suite green at HEAD 54ea460)
