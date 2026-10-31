---
phase: "11"
slug: "menu-punkt-v-chernyy-spisok"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-31"
---

# Phase 11 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (toolchain через mise; гонки через -race) |
| **Config file** | mise.toml |
| **Quick run command** | `mise exec -- go test ./internal/indicator/ ./internal/config/ ./internal/appid/ ./internal/session/ -race -count=1` |
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
| (заполняется из планов после планировщика) | 01 | 1 | MENU-01 | — | N/A | unit+golden | `mise exec -- go test ./internal/... -race -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Пункт виден в живом трей-меню GNOME с точным именем приложения; клик добавляет запись, автокоррекция замолкает | MENU-01 | трей-меню рендерит только живой стол (DBusMenu); e2e-ряд на ВМ покрывает машину, by-ear приёмка — владельца | e2e-ряд menu-blocklist на ВМ-стенде; владелец открывает меню, кликает пункт, проверяет тишину автокоррекции |
| Неактивный пункт при неизвестной личности (wezterm/snap) — без имени, без пояснений, клик ничего не пишет | MENU-01 (D-03 CONTEXT) | визуальное состояние DBusMenu на живом столе | сфокусировать wezterm, открыть меню: пункт серый без имени; клик не меняет config.yaml |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
