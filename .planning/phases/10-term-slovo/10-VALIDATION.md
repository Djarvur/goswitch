---
phase: "10"
slug: "term-slovo"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-10-12"
---

# Phase 10 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Post-execution audit (validate-phase 2026-10-12): все задачи фазы имеют автоматизированную верификацию; пропусков покрытия (gaps) не обнаружено.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib testing, project convention — testify запрещён) |
| **Config file** | mise.toml (`[tasks]` ci / test / lint + e2e-term-* задачи, NOT in ci) |
| **Quick run command** | `mise exec -- go test ./internal/... -race -count=1` |
| **Full suite command** | `mise run ci` (build + vet + golangci-lint v2 strict + test -race) |
| **Estimated runtime** | ~15–60 с (измерено на HEAD фазы: полный ci ≈ 10–15 с) |

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
| 10-01-01 | 01 | 1 | TERM-02 | — | N/A | e2e (стенд) | ssh-пробы стенда; строка `venue-ready` журнала ADR-008 | ✅ | ✅ green |
| 10-01-02 | 01 | 1 | TERM-02 | — | N/A | e2e | venue-probe r1+r2 — ряды журнала ADR-008 (key-trace + pty hex, флип-контроль ×2) | ✅ | ✅ green |
| 10-01-03 | 01 | 1 | TERM-02 | — | N/A | other | `grep venue-verdict docs/adr/adr-008-spike-log.md` | ✅ | ✅ green |
| 10-02-01 | 02 | 1 | TERM-02 | — | N/A | unit | `go test -race ./internal/correct ./internal/session -run 'TestConvertibleStretch\|TestActor_DoubleTapPunctuationStretch' -count=1` | ✅ | ✅ green |
| 10-02-02 | 02 | 1 | TERM-02 | — | N/A | unit | `go test -race ./internal/config ./cmd/skillgen -run 'Terminals\|CommittedFileInSync' -count=1` | ✅ | ✅ green |
| 10-02-03 | 02 | 1 | TERM-02 | — | N/A | unit | `go test -race ./internal/clipboard -run TestClipboard_Primary -count=1` | ✅ | ✅ green |
| 10-03-01 | 03 | 2 | TERM-02 | — | N/A | unit | `go test -race ./internal/session -run TestTerminalClass_ -count=1` (6 тестов + silence-matrix) | ✅ | ✅ green |
| 10-03-02 | 03 | 2 | TERM-02 | — | N/A | unit | `go test -race ./internal/session -run TestAutoCorrect_TerminalVeto -count=1` (5 тестов, D-55) | ✅ | ✅ green |
| 10-04-01 | 04 | 3 | TERM-02 | — | N/A | decision | checkpoint-вердикты владельца OQ1/OQ2 — пиннуты downstream-тестами (см. 10-04-02/03) + docs/CONFIG.md D7 | ✅ | ✅ green |
| 10-04-02 | 04 | 3 | TERM-02 | — | N/A | unit | `go test -race ./internal/session -run 'TestTerminalBranch_(RoutesChunkedNoPending\|VTERoleUnifiedChunks\|FocusLossBetweenChunksStops\|FreeAbortBeforeBurst\|LegacyNoCapsChunkedToo\|TailRidesReplay)' -count=1` | ✅ | ✅ green |
| 10-04-03 | 04 | 3 | TERM-02 | — | N/A | unit | `go test -race ./internal/session -run 'TestTerminalBranch_(VerifyAfterNoRungOnMismatch\|VerifyDeadQuietTimeout\|VerifyAfterQuietConfirm)' -count=1` | ✅ | ✅ green |
| 10-05-01 | 05 | 4 | TERM-02 | — | N/A | e2e | `mise run e2e-term-word-vte` (NOT in ci; живые ряды word-bash/zsh/mid-line/clip-frozen — журнал ADR-008) + `mise run ci` | ✅ | ✅ green |
| 10-05-02 | 05 | 4 | TERM-02 | — | N/A | unit+e2e | `TestTerminalBranch_Level2PacesWire` + `TestRegisterEventInterest_PinsRegistryCall` + ряды слова/клипа после pacing-фикса (байт-точные) | ✅ | ✅ green |
| 10-05-03 | 05 | 4 | TERM-02 | — | N/A | e2e | `mise run e2e-term-negative-witnesses` (NOT in ci; witness-vim/htop fired=0 + positive-control) | ✅ | ✅ green |
| 10-06-01 | 06 | 5 | TERM-02 | — | N/A | e2e | `mise run e2e-term-word-wezterm` (NOT in ci; caps re-pin, ×2 двухподряд, дельта-оракул) | ✅ | ✅ green |
| 10-06-02 | 06 | 5 | TERM-02 | — | N/A | e2e+other | ряды ×2/mid-line/clip/leg-probe/re-pin + `grep Accepted docs/adr/ADR-008-terminal-transport.md` (Evidence D1–D7) | ✅ | ✅ green |
| 10-06-03 | 06 | 5 | TERM-02 | — | N/A | other+manual | ряд `host-install` журнала + owner-зачёт (Manual-Only, ниже) | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky — полное состояние на HEAD: `mise run ci` зелёный (build + vet + golangci-lint 0 issues + test -race, все пакеты).*

---

## Wave 0 Requirements

- [x] юнит-стабы для TERM-02 (internal/correct + internal/session — геометрия пробега с пунктуацией, терминальный класс контекста) — закрыты 10-02/10-03 (8 × TestConvertibleStretch_*, TestActor_DoubleTapPunctuationStretch, corpus TestTerminalClass_*/TestAutoCorrect_TerminalVeto*)
- [x] живые ряды: проверка клавиатурности фоновых окон внутри VM-стенда goswitch-stand — закрыто 10-01 (venue-probe r1+r2, вердикт vm-primary)

*Существующая инфраструктура покрывает все требования фазы; новые тестовые каркасы не требовались.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| ghbdtn → привет на живом столе владельца (критерий 1, финальная приёмка) | TERM-02 | жест владельца — human UAT на harvest фазы; само поведение уже доказано автоматикой (живые ряды wezterm ×2 и VTE, байт-точный pty-оракул) | 5-пунктовый чек-лист из 10-06-SUMMARY («User Setup Required»): ghbdtn+двойной правый Shift в wezterm и GNOME Terminal, середина строки, «ytn?»→«нет,», молчание автокоррекции |
| Классификация края слова TERM-02 (flagged assumption всех шести планов) | TERM-02 | процессная обязанность фазовой верификации (SPEC без ## Edge Coverage) — вердикт человека, не поведение кода; все поведенческие края покрыты корпусом и живыми рядами | verifier фиксирует классификационный вердикт края в фазовой верификации (не тихий pass) |
| D-24 ветка `changed=false` (instant-done) | TERM-02 | защитная ветка, недостижимая по построению (ConvertRuns никогда не возвращает changed=false); поведенческий тест требует нового шва в production-коде — вне объёма фазы (WINDOWS ledger, kind: deviation) | защитная запись; пересмотреть при любом изменении контракта ConvertRuns |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 120s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** validated 2026-10-12 (validate-phase; gaps: 0)

---

## Validation Audit 2026-10-12

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

Аудит Nyquist-покрытия по шести SUMMARY (State A): требование TERM-02 (единственное требование фазы) покрыто автоматизированной верификацией на каждой задаче — юнит-корпус (internal/correct, internal/session, internal/config, internal/clipboard, internal/appid, cmd/skillgen), автоматизированные e2e-кейсы (test/e2e/case_term_*.go через mise-задачи, NOT in ci) и журнальные оракулы ADR-008. Новые тесты не генерировались — пропуск не обнаружен. Критерии 1–5 ROADMAP: 1 — живые ряды wezterm ×2 (10-06) + owner-зачёт (manual), 2 — VTE-ряды (10-05), 3 — clip-frozen обоих терминалов + noRung-юниты, 4 — free-abort/чанки/перечек фокуса (10-04 юниты), 5 — негативные свидетели (10-05) + вето D-55 (10-03 юниты).
