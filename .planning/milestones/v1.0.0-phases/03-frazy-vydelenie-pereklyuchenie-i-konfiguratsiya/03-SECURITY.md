---
phase: "3"
slug: "frazy-vydelenie-pereklyuchenie-i-konfiguratsiya"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-09-15"
---

# Phase 3 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.
> Built from the `<threat_model>` blocks of all seven Phase 3 plans (register authored at plan time);
> verified at L1 grep depth per ASVS L1 with the `-race` suite green at HEAD 54ea460 (re-run twice
> 2026-09-15 by the Phase 1/2 re-verifiers) and the live matrix v2 21/21 PASS at HEAD.

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| клиент → engine (SetSurroundingText + anchorPos) | text field content and cursor/anchor — client-originated data; may be inconsistent | user text, coordinates |
| engine → клиент (Delete/Commit/ForwardKeyEvent) | outgoing signals mutate the user's field — the largest replacement ranges of the phase | replacement ranges, synthetic keys |
| конфиг-файл → демон (YAML в $HOME) | external input: typos, giant values, YAML parser exploits | config values |
| watcher → актор | config snapshots drive live input behavior (window, bindings, caps) | snapshot values |
| демон ↔ буфер обмена (wl-copy/wl-paste) | user clipboard content transits the daemon | user text |
| Super-комбинации GNOME → демон | interception of shell-bound gestures; a11y bus focus events (appid) | key chords, focus events |
| session bus → ctlsvc | incoming D-Bus Methods from any process of the same user | control commands |
| matrix-v2.yaml → стенд; стенд → живой стол | case definitions; injections drive the owner's real desktop | test data |

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-03-01-01 | Tampering | phrase pipeline (large range) | high | mitigate | ADR-004 verify over the whole range BEFORE any deletion; invalid head/tail → verify-mismatch, zero emits — TestActor_PhraseVerifyMismatch green | closed |
| T-03-01-02 | DoS | Backspace×N series for a long phrase | high | mitigate | D-27 cap (DefaultBackspaceCap = 50, plan.go:30); over-cap → silent backspace-cap refusal, zero deletions — TestBuildPlan_BackspaceCap green | closed |
| T-03-01-03 | Tampering | mixed-text run conversion | medium | mitigate | Convert not forked: an unmapped letter fails the whole range (ok=false → refuse); partial corruption impossible — TestConvertRuns_Refusals green | closed |
| T-03-01-04 | Information Disclosure | phrase-correction logs | high | mitigate | D-20/D-21: INFO carries outcome/reason only; source→result only under -debug — TestLogging_* corpus (Phase 1) + phase-3 log pins green | closed |
| T-03-02-01 | Tampering | config swapped/corrupted by a neighbor process | medium | accept | single-user session (ASVS L1); strict-parse limits impact, last-good prevents breakage — see Accepted Risks | closed |
| T-03-02-02 | DoS | giant values (window 10⁹, cap 10⁶) | high | mitigate | Validate() ceilings (window/verify_wait ≤ 2000 ms, caps bounded, apps ≤ 64) — TestValidate_TimeoutRanges / TestValidate_CorrectionRanges / TestLoad_RangeViolationRejected green | closed |
| T-03-02-03 | DoS | malicious YAML (billion-laughs) | medium | mitigate | yaml.v3 v3.0.1 (known-DoS fixed, STACK); strict schema narrows the surface | closed |
| T-03-02-04 | Information Disclosure | config contents in logs | low | mitigate | path/window/validity logged, never section contents; docs/CONFIG.md note | closed |
| T-03-02-SC | Tampering | supply chain: fsnotify dependency | high | mitigate | v1.10.1 pinned in go.mod; legitimacy audited in RESEARCH (proxy.golang.org verified 2026-05-04); no [SLOP] packages | closed |
| T-03-03-01 | Information Disclosure | clipboard round-trip | high | mitigate | content never logged; transfer via stdin pipe only (internal/clipboard/clipboard.go:154 cmd.Stdin) — argv invisible in ps; TestClipboard_* green | closed |
| T-03-03-02 | Tampering | clipboard swapped without restore | high | mitigate | byte-exact Save BEFORE swap, Restore AFTER (D-29); empty source → wl-copy --clear; failure → WARN (owner-sanctioned best-effort); CR-03 kill-vs-empty distinction — TestClipboard_KilledNotEmpty / DeadlineKillNotEmpty green | closed |
| T-03-03-03 | Tampering | inconsistent cursor/anchor from client | medium | mitigate | SelectionRange start=min/end=max + clamp to len(text); VerifyRangeAt before/after — TestSelectionRange, TestActor_Selection* green | closed |
| T-03-03-04 | Elevation of Privilege | subprocess execution of wl-copy/wl-paste | low | mitigate | fixed binary names from PATH; arguments are flags only, no user input; per-call context deadline | closed |
| T-03-03-05 | DoS | hung wl-clipboard | medium | mitigate | every subprocess context-deadline-bounded — TestClipboard_Timeouts green; post-timeout rung silently declines | closed |
| T-03-04-01 | DoS | reload path inside the key handler | high | mitigate | lock-free atomic snapshot read; timer re-arm forbidden (Pitfall 8); race-free by construction, pinned under -race — TestActor_HotReload* green | closed |
| T-03-04-02 | Tampering | combo branch weakens series discrimination | high | mitigate | combo resets the series only on the exact binding (FamilyMask held-subset predicate); FSM path unchanged for other keys — TestActor_Combo*, TestFSM_ModifierUse green | closed |
| T-03-04-03 | Information Disclosure | mode/combo logs | low | mitigate | kind/mode/counters only, no content — continuation of D-20/D-21 | closed |
| T-03-05-01 | Tampering (UX) | interception of mutter-bound Super+letters | high | mitigate | only the configured set intercepted; default letters empty (off); consumed-upstream WARN makes swallowing visible — TestActor_MACRDisabledByDefault / ConsumedUpstream green | closed |
| T-03-05-02 | DoS | a11y bus unavailable/hung | medium | mitigate | appid returns errors, never panics; degradation to global rules with WARN; context-canceled goroutine — TestActor_MACRAppidDegradation, TestAppid_FocusedAppByFakeBus green | closed |
| T-03-05-03 | Information Disclosure | MACR logs | low | mitigate | logs carry config letters and counters; keyboard content never written (D-20) | closed |
| T-03-05-04 | Spoofing | fake focus events on the a11y bus | low | accept | single-user session (ASVS L1); worst case — a wrong per-app verdict within the same user — see Accepted Risks | closed |
| T-03-06-01 | Spoofing | goswitch control name seized by a foreign process | medium | mitigate | RequestName ReplaceExisting-without-AllowReplacement → ErrNotPrimaryOwner guard (ctlsvc.go:40) + sentinel; SO_PEERCRED same-uid session bus — TestSvc_NameGuard green | closed |
| T-03-06-02 | DoS | a ctlsvc method panic kills the daemon | high | mitigate | recoverMethod on every exported method (×5) — panic → dbus.Error, daemon lives — TestSvc_RecoverShim green | closed |
| T-03-06-03 | Information Disclosure | status discloses user text | high | mitigate | StatusSnapshot carries mode/counters/config-status only; correction text never included (D-20) — TestSvc_Status, TestActor_StatusSnapshot green | closed |
| T-03-06-04 | Tampering | CorrectNow from any user process | low | accept | single-user session (ASVS L1): any same-uid process already owns the session; impact within uid — see Accepted Risks | closed |
| T-03-07-01 | DoS | unstable injection names drop cases | medium | mitigate | matrixKeyNames canonicalized per spike tables; decode rejects unknown names — TestMatrixKeyNames_Canonical, TestMatrixDecode green | closed |
| T-03-07-02 | Tampering | off-schema case field silently ignored | high | mitigate | strict KnownFields matrix decode (matrix.go) — TestMatrixDecode_RejectsNew green | closed |
| T-03-07-03 | DoS | orphaned IBus names/processes between cases | high | mitigate | fresh-daemon isolation + teardown checks (matrix.go); preflight catches the orphaned name, ibus restart heals — preflight 7/7 + 21/21 PASS at HEAD 54ea460 (2026-09-15) | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-03-1 | T-03-02-01 | Config file swap/corruption by a neighbor process of the same user: single-user session (ASVS L1), strict-parse limits impact, last-good keeps input behavior stable | Owner (plan 03-02, ASVS L1 policy) | 2026-09-15 |
| AR-03-2 | T-03-05-04 | Fake a11y focus events: same-user spoofing only; worst case is a wrong per-app rule verdict within the same uid | Owner (plan 03-05, ASVS L1 policy) | 2026-09-15 |
| AR-03-3 | T-03-06-04 | CorrectNow callable by any same-user process: same-uid processes already own the session; impact bounded to uid | Owner (plan 03-06, ASVS L1 policy) | 2026-09-15 |

*Accepted risks do not resurface in future audit runs.*

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-09-15 | 28 | 28 (25 mitigated + 3 accepted) | 0 | Claude (verify-work post-hook, L1 grep depth + green -race suite + live matrix 21/21 at HEAD) |

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-09-15
