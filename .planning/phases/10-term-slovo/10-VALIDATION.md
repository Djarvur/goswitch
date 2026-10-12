---
phase: "10"
slug: "term-slovo"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-12"
---

# Phase 10 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib testing, project convention — testify запрещён) |
| **Config file** | mise.toml (`[tasks]` ci / test / lint) |
| **Quick run command** | `mise exec -- go test ./internal/... -race -count=1` |
| **Full suite command** | `mise run ci` |
| **Estimated runtime** | ~60–120 seconds |

---

## Sampling Rate

- **After every task commit:** Run `mise exec -- go test ./internal/... -race -count=1`
- **After every plan wave:** Run `mise run ci`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| (заполняется планами фазы) | — | — | TERM-02 | — | N/A | unit | `mise exec -- go test ./internal/... -race -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] юнит-стабы для TERM-02 (internal/correct + internal/session — геометрия пробега с пунктуацией, терминальный класс контекста)
- [ ] живые ряды: проверка клавиатурности фоновых окон внутри VM-стенда goswitch-stand (libvirt, стоковое ядро) — пре-реквизит livе-верификации фазы

*Если инфраструктура покрывает — отметить в планах явно.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| ghbdtn → привет на живом столе владельца (критерий 1) | TERM-02 | финальное подтверждение жестом владельца (до VM-стенда) | набрать ghbdtn в wezterm/VTE, двойной правый Shift, увидеть «привет» |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
