---
phase: "01"
slug: "adr-paket-i-karkas-ibus-dvizhka"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-09-11"
---

# Phase 01 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.
> State B reconstruction from PLAN `<threat_model>` blocks; L1 (grep-depth) mitigation
> verification per ASVS level 1 short-circuit — no open threats at plan-authored
> register time.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| ibus-daemon → goswitchd (private D-Bus) | daemon invokes exported engine methods with variant args — the single external-data entry point | keystroke events (sensitive: typed text) |
| goswitchd → stderr/journal | structured log read by owner and e2e stand | buffer lengths (INFO) / key trace (DEBUG, opt-in) |
| системные xkb/keysym-файлы → генератор (dev-side) | read-only inputs; not used at runtime or in CI | xkb layout definitions |
| e2e-раннер → живой рабочий стол владельца | stand injects keypresses and mutates gsettings — reversible impact on the live session | session input-source state |
| ydotool → /dev/uinput (kernel) | physical injection path — stand tool, unreachable from product code | synthetic key events |
| workflow-файлы → среда выполнения GitHub Actions | CI executes repo code on every push/PR; workflow defines the rights boundary | repo contents, module proxy downloads |
| CI-раннер → module proxy и инсталлер mise | tools are downloaded on every run | golang.org/x/vuln, mise.run installer |
| dependabot → репо | bot opens dependency-bump PRs; every PR runs full pr-sanity | go.mod updates |
| журнал эксперимента → ADR-001 → будущий код Фаз 2-3 | evidence chain of decisions | observed verdicts |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-01-01 | Information disclosure | internal/logging + engine.ProcessKeyEvent | high | mitigate | key trace only behind `-debug` (default off), INFO logs buffer lengths only; unit test gates it (`internal/logging/logging.go:21`, logging_test.go) | closed |
| T-01-02 | Denial of service | engine/engine.go handlers | high | mitigate | recover shim on every exported handler + comma-ok variant assertions; TestEngine_RecoverContainsPanic (engine/engine.go:102) | closed |
| T-01-03 | Spoofing | engine/conn.go RequestName | medium | mitigate | reply checked; not PrimaryOwner → errNotPrimaryOwner fatal guard against a duplicate engine (engine/conn.go:93) | closed |
| T-01-04 | Tampering | component registration | low | accept | Phase 1 uses runtime-only registration (no XML / system files) — nothing to subvert; XML install is Phase 4 | closed |
| T-01-05 | Tampering | mise installation (official installer) | medium | mitigate | official mise.run installer to ~/.local/bin without root; version pinned in mise.toml; same installer in ephemeral CI runner with contents:read | closed |
| T-01-SC | Tampering (supply chain) | go get godbus/dbus/v5 | high | mitigate | package legitimacy audited in RESEARCH: godbus v5.2.2 verified via proxy.golang.org; no packages beyond the stack table | closed |
| T-02-01 | Tampering | layouts/generator (reads system files) | low | accept | dev-side generator; output committed and pinned by golden tests — substitution cannot pass the diff review | closed |
| T-02-02 | Denial of service | internal/hotkey FSM | low | mitigate | pure deterministic FSM; burst-of-ten corpus pins it against cycling/panic (TestFSM_BurstOfTen, fsm_test.go:274) | closed |
| T-03-01 | Tampering (session state) | gsettings sources/current | high | mitigate | snapshot BOTH keys before, restore in defer-teardown on every outcome; restore mismatch → FAIL (test/e2e case_m1.go, case_d01.go) | closed |
| T-03-02 | Denial of service (desktop) | kill9-survive case | medium | mitigate | case first switches the source to a plain xkb engine and proves live input; injection limited to short strings (case_resilience.go) | closed |
| T-03-03 | Information disclosure | goswitchd -debug log in temp file | medium | mitigate | temp file in $TMPDIR, removed in teardown; stand never runs in CI | closed |
| T-03-04 | Tampering | /dev/uinput from the stand | low | accept | injection is the stand's purpose (simulating the human); 0660 input group is the stock mechanism; product code separated by the structural gate of plan 01 | closed |
| T-04-01 | Elevation of privilege / Tampering | Shell.Eval probe, unsafe mode | high | mitigate | unsafe mode never enabled; Eval refusal recorded as verdict unavailable; micro-extension probe user-local only, removed after | closed |
| T-04-02 | Tampering (session state) | gsettings sources/current in probes | high | mitigate | stand snapshot/restore reused; teardown compares both keys with the snapshot, mismatch → FAIL; independently checked by Task 1 verify (RESTORED line) | closed |
| T-04-03 | Repudiation (evidence base) | d01-experiment-log.md | medium | mitigate | journal written case-by-case (command + observation + verdict); ADR-001 must reference the journal; no two-engine claim without an observation (prohibition) | closed |
| T-04-04 | Tampering | ~/.local/share/gnome-shell/extensions/ | medium | mitigate | conditional probe, user-local dir (no root), removed after the probe — Task 1 acceptance criterion | closed |
| T-05-01 | Tampering | govulncheck via go run @latest | medium | mitigate | golang.org/x/vuln is the Go team's first-party module; ephemeral runner, contents:read, no secrets; pin on first tool update | closed |
| T-05-02 | Elevation of privilege | workflow token rights | low | mitigate | `permissions: contents: read` at workflow level in both pr-sanity.yml and security-scheduled.yml (verified) | closed |
| T-05-03 | Tampering | mise install in CI (curl installer) | medium | mitigate | official mise.run domain, ephemeral runner, contents:read, no secrets; reproducible locally (T-01-05) | closed |
| T-05-04 | Tampering | dependabot PR on godbus bump | low | mitigate | every dependabot PR automatically runs full pr-sanity (dependabot.yml documents the mechanism); update cannot reach main without the green gate | closed |
| T-05-SC | Tampering (supply chain) | govulncheck@latest + mise install in CI | low | mitigate | first-party Go team module + official jdx/mise installer, delivered via module proxy with checksum verification; no new runtime deps for the daemon | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on (high) count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-01-01 | T-01-04 | Runtime-only component registration in Phase 1 — no delivery file exists to tamper with; XML install + directory permissions deferred to Phase 4 | Owner (plan 01-01 disposition) | 2026-09-10 |
| AR-01-02 | T-02-01 | Generator is a dev-side tool; committed output pinned by golden tests makes subversion visible in diff review | Owner (plan 01-02 disposition) | 2026-09-10 |
| AR-01-03 | T-03-04 | uinput injection is the e2e stand's designed purpose (simulating the human); product code structurally separated from the stand | Owner (plan 01-03 disposition) | 2026-09-10 |

*Accepted risks do not resurface in future audit runs.*

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-09-11 | 21 | 21 | 0 | gsd verify-work (L1 grep-depth, ASVS 1 short-circuit) |

L1 evidence spot-checks (grep-level): debug gating `internal/logging/logging.go:21`; recover shim `engine/engine.go:102` + TestEngine_RecoverContainsPanic; RequestName guard `engine/conn.go:93`; TestFSM_BurstOfTen `internal/hotkey/fsm_test.go:274`; snapshot/restore `test/e2e/case_m1.go:82`, `case_d01.go:279`; `permissions: contents: read` in both workflows; dependabot-through-gate documented in `.github/dependabot.yml`.

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-09-11
