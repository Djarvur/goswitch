---
phase: "8"
slug: "a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-05"
---

# Phase 8 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib, -race), golden-файлы для генерируемых артефактов |
| **Config file** | none — инфраструктура существует (mise.toml, .golangci.yml) |
| **Quick run command** | `mise exec -- go test ./internal/config/ ./internal/a11y/ -race -count=1` (пакет a11y — по фактическому имени из плана) |
| **Full suite command** | `mise run ci` (build + vet + test -race + lint + tidy-diff + govulncheck) |
| **Estimated runtime** | ~60–90 seconds (quick), ~3–5 minutes (full) |

---

## Sampling Rate

- **After every task commit:** Run quick run command (пакеты фазы)
- **After every plan wave:** Run `mise run ci`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 90 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| *(заполняется планировщиком: по строке на задачу планов 08-NN с их `<automated>` командами и `<fails_when>`)* | | | | | | | | | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] *(при необходимости: RED-стабы новых пакетов — internal/a11y|reconcile, генератор SKILL.md — по решению планировщика)*
- [ ] Существующая инфраструктура покрывает go test/lint/mise — новый фреймворк не нужен

*Existing infrastructure covers phase needs; Wave 0 только при новых пакетах (прецедент 06-02 dictgen).*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| a11y-магия реально открывает a11y-дерево приложения из списка (ZCode/Electron) | D-8-1/D-8-6 (CONTEXT) | Эффект наблюдаем владельцем в живом приложении; AT-SPI-поддерево меняется после перезапуска приложения | Включить приложение в список a11y, перезапустить его, проверить a11y-дерево (прецедент 260930-pf6 owner pixel check) |
| README русский-первый читаем и синхронен | D-8-7 (CONTEXT) | Языковое качество и структура — владенческая оценка | Прочитать README.md (RU) / README.en.md, сверить разделы |
| SKILL.md полезен AI-ассистенту как документация конфигурации | поток 3 | Потребитель — AI-сессия; оценивает владелец | Дать ассистенту SKILL.md, попросить изменить настройку (например, тумблер звука) без чтения кода |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
