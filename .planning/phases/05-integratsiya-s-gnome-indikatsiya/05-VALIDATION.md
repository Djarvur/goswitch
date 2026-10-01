---
phase: "5"
slug: "integratsiya-s-gnome-indikatsiya"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-28"
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Seeded from `05-RESEARCH.md` § Validation Architecture (2026-09-28).

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` (+`-race`), go-ultimate conventions (`package xxx_test`, `TestF_suffixCamelCase`) |
| **Config file** | `.golangci.yml` (strict v2), `mise.toml` (tasks + tools) |
| **Quick run command** | `mise run test` (go test -race -count=1 ./...) |
| **Full suite command** | `mise run ci` (build + vet + lint + test) |
| **Estimated runtime** | ~60 seconds (unit corpus); e2e cases are live-desk, minutes each |

---

## Sampling Rate

- **After every task commit:** Run `mise run test`
- **After every plan wave:** Run `mise run ci` + live e2e switch cases on the desk
- **Before `/gsd:verify-work`:** Full matrix green twice in the two-source configuration (D-48 nightly gate form)
- **Max feedback latency:** 120 seconds (unit corpus)

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| *(populated by validate-phase §6 after PLAN.md files exist)* | | | | | | | | | |

Requirement → test-type baseline from research (planner must keep every row automated or explicitly manual):

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SWCH-01 | single Shift flips via SetGlobalEngine seam; mode record preserved | unit (seam double) + e2e layout-single | `go test -race ./internal/session/ -run TestActor_Flip` / live layout-single case | ✅ actor corpus exists; seam tests Wave 0 |
| SWCH-02 | combo word+flip through the seam, D-36 order | unit | `go test -race ./internal/session/` (combo corpus) | ✅ extend |
| SWCH-03 | indicator/GNOME-side accuracy | manual (owner) + spike | live spike + owner gate | ❌ Wave 0 (spike script) |
| SWCH-04 | tap timings unchanged | unit (FSM corpus) | `go test -race ./internal/hotkey/` | ✅ |
| INTEG-01..03 | unchanged engine contracts | existing corpus | `go test -race ./engine/` | ✅ |
| INTEG-04 | reactivation after ibus restart under two sources | e2e ibus-restart + unit (activate) | `mise run e2e-ibus-restart`; `go test -race ./internal/activate/` | ✅ extend (engine-name derivation fix) |
| INTEG-05 | recover shim on any new exported path | unit | `go test -race ./engine/` | ✅ extend |
| INST-01 | wrap takeover, refusal, verbatim restore, selfcheck two-source | unit (install corpus with fake Runner) + live install-cycle | `go test -race ./internal/install/`; live install/uninstall cycle | ✅ extend (wrap corpus Wave 0) |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] Spike script/checklist (two-source window + flip observation + restore) — SWCH-03/criterion 2 evidence vehicle
- [ ] `internal/install` wrap corpus: pair wrap, positional preservation, third-source passthrough, refusal table, already-wrapped upgrade, verbatim restore unchanged
- [ ] `internal/activate` engine-name derivation corpus (GetGlobalEngine path + cold-bus fallback)
- [ ] e2e: switch-case extension grepping the new `SetGlobalEngine` journal record (the `"msg":"mode"` marks stay byte-stable)
- [ ] ADR-006 skeleton (docs/adr format per ADR-001)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Panel indicator shows active layout (en/ru), changes on every flip | SWCH-03 / criterion 2 | Not observable without eyes — no AT-SPI surface for the GNOME panel indicator | Owner watches the top panel across flips driven by the live spike (single Shift, Super+Space, flip-after-correction); confirms label change each time |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
