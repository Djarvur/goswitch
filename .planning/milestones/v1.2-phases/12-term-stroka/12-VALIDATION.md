---
phase: "12"
slug: "term-stroka"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-31"
---

# Phase 12 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (toolchain через mise; гонки через -race) |
| **Config file** | mise.toml |
| **Quick run command** | `mise exec -- go test ./internal/session/ ./internal/clipboard/ ./internal/hotkey/ -race -count=1` |
| **Full suite command** | `mise run ci` |
| **Estimated runtime** | ~60 секунд |

---

## Sampling Rate

- **After every task commit:** Run quick run command (пакеты фазы)
- **After every plan wave:** Run `mise run ci`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| (заполняется из планов после планировщика) | 01 | 1 | TERM-03 | — | N/A | unit+golden | `mise exec -- go test ./internal/... -race -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Тройной Shift исправляет набранную строку в живом wezterm (курсор/суффикс по SC1) | TERM-03 | живой стол + XIM-транспорт владельца; e2e-ряд на ВМ покрывает машину, by-ear зачёт — владельца | набрать строку не в той раскладке, тройной правый Shift, сверить курсор/суффикс |
| VTE: тройной Shift молчит (D-02), слово двойным работает | TERM-03 / D-02 | визуальное отсутствие реакции | gnome-terminal: тройной Shift — ничего; двойной — слово исправлено |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
