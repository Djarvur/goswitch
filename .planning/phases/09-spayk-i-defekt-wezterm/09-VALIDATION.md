---
phase: "9"
slug: "spayk-i-defekt-wezterm"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-08"
---

# Phase 9 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` + race detector (unit-able code); live e2e stand `test/e2e` (Go CLI runner, `PASS <case>` / `FAIL`, exit ≠ 0 на провал) — НЕ `go test` |
| **Config file** | `mise.toml` ([tasks]; головные гейты — `ci` = build+vet+lint+test), `.golangci.yml` (strict v2) |
| **Quick run command** | `mise exec -- go test ./internal/session/ ./internal/ctlsvc/ ./cmd/... -race -count=1` (пакеты фазы) |
| **Full suite command** | `mise run ci` (headless); live-ряды — `mise run e2e-spike-*` / `mise run e2e-wezterm-flip` / `mise run e2e-autocorrect-terminal-silent` / `mise run e2e-matrix-v4` (live GNOME session; NOT in ci — ручные ряды стенда) |
| **Estimated runtime** | ~90 секунд (headless ci); live-ряды — минуты, машинно-зависимые |

---

## Sampling Rate

- **After every task commit:** `mise run ci` (директива 2 «зелёная итерация» — инвариант конца каждой задачи, не финальный гейт)
- **After every plan wave:** полный headless (`mise run ci`) + живые ряды задач плана на сессии владельца (правило стенда: юнит `goswitchd` остановлен перед живым прогоном)
- **Before `/gsd:verify-work`:** полный набор зелёный + оракул `wezterm-flip` = `verdict: fixed` (два раунда в одном запуске) + `e2e-autocorrect-terminal-silent` + `e2e-matrix-v4` + ADR-008 Accepted
- **Max feedback latency:** ~90 секунд

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 09-01-01 | 01 | 1 | TERM-01, TERM-04 | T-09-01-01/02 | вето в точке армирования; аудит-трейл дословен | docs content (grep) | `test "$(grep -c 'rev D-55' docs/SPEC.md)" -ge 5 && grep -q 'никогда не армируется в контексте терминала' docs/SPEC.md` | ✅ | ⬜ pending |
| 09-01-02 | 01 | 1 | TERM-01, TERM-04 | T-09-01-03 | Evidence только из журнала спайка; без веб-ссылок | docs content (grep) | `grep -oE '^\| D[1-7] ' docs/adr/ADR-008-terminal-transport.md \| sort -u \| wc -l` = 7; `! grep -q 'https\?://' docs/adr/ADR-008-terminal-transport.md` | ❌ создаст 09-01-02 | ⬜ pending |
| 09-02-01 | 02 | 2 | TERM-01 | T-09-02-01/02 | keyspec строго валидирован; nil-dep fail-closed; эмиссия логируется | unit (TDD red→green) | `mise exec -- go test ./internal/session/ ./internal/ctlsvc/ ./cmd/... -race -count=1` | ❌ W0 этой задачи (probe_test.go/ctlsvc_test.go-расширение создаются red-шагом задачи) | ⬜ pending |
| 09-02-01 live | 02 | 2 | TERM-01 | T-09-02-03 | только поверхности стенда; фокус-пруф | live e2e | `go run ./test/e2e -case spike-smoke` | ❌ W0 (кейс+реестр+mise создаются задачей) | ⬜ pending |
| 09-02-02 | 02 | 2 | TERM-01 | T-09-02-02/05 | hex-гигиена свидетельств; вердикты только живые | live e2e | `go run ./test/e2e -case spike-wezterm` | ❌ W0 (расширение кейса задачи) | ⬜ pending |
| 09-02-03 | 02 | 2 | TERM-01, TERM-04 | T-09-02-02/03/05 | сентинелы PRIMARY; оракул честный RED | live e2e | `go run ./test/e2e -case spike-caps`; `go run ./test/e2e -case wezterm-flip \| grep -q 'verdict: defect-reproduced'`; `mise run e2e-autocorrect-terminal-silent` | ❌ W0 (кейсы+задачи создаются задачей) | ⬜ pending |
| 09-03-01 | 03 | 3 | TERM-01 | T-09-03-03/04 | ветка фиксируется только решением владельца | checkpoint gate (журнал) | `grep -qi 'ветка фикса' docs/adr/adr-008-spike-log.md` | ❌ W0 (журнал создаёт 09-02) | ⬜ pending |
| 09-03-02 | 03 | 3 | TERM-01 | T-09-03-01/02 | оракул two-greens на целевом транспорте; ADR без веб-ссылок | live e2e + grep | ВЕТКА (a): `go run ./test/e2e -case wezterm-flip -wezterm-config test/e2e/fixtures/gs-e2e-wezterm-wayland.lua \| grep -q 'verdict: fixed'` + контроль `grep -q 'verdict: defect-reproduced'` на дефолте; ВЕТКА (b): `go run ./test/e2e -case wezterm-flip \| grep -q 'verdict: fixed'`; ВЕТКА (c): `! wezterm --version \| grep -q '20240203' && go run ./test/e2e -case wezterm-flip -wezterm-config test/e2e/fixtures/gs-e2e-wezterm-wayland.lua \| grep -q 'verdict: fixed'`; `grep -q 'Accepted' docs/adr/ADR-008-terminal-transport.md`; + human-check владельца | ⚠️ частично (кейс/журнал — из 09-02; wayland-фикстура — из 09-02; остальное создаёт задача) | ⬜ pending |
| 09-03-03 | 03 | 3 | TERM-01, TERM-04 | T-09-03-06 | замороженные негативы зелёные на закрытии | checkpoint gate + live e2e | `mise run e2e-autocorrect-terminal-silent && mise run e2e-matrix-v4 && mise run ci` | ✅ (существующие ряды; при ветке удаления W0 — см. ниже) | ⬜ pending |
| 09-04-01 | 04 | 3 | A11Y-02 (spike) | T-09-04-01/04/05 | env только точечными grep-ами; приёмка по живому дереву | live e2e | `go run ./test/e2e -case spike-a11y` (+гейт остатка: absent OR `grep -q 'override-kept' docs/adr/d09-a11y-probe-log.md`) | ❌ W0 (кейс+реестр+mise создаются задачей) | ⬜ pending |
| 09-04-02 | 04 | 3 | A11Y-02 (spike) | T-09-04-02 | flatpak только по согласию | checkpoint gate (журнал) | `grep -qi 'flatpak' docs/adr/d09-a11y-probe-log.md && grep -qiE 'согласие\|consent\|deferred\|отлож' docs/adr/d09-a11y-probe-log.md` | ⚠️ частично (журнал — из 09-04-01) | ⬜ pending |
| 09-04-03 | 04 | 3 | A11Y-02 (spike) | T-09-04-02/05 | deferred ⇒ пакет отсутствует (exit ≠ 0 на нарушении) | live e2e | `go run ./test/e2e -case spike-a11y`; гейт согласия: `if grep -qi 'deferred-to-phase-13' docs/adr/d09-a11y-probe-log.md; then ! flatpak list 2>/dev/null \| grep -qi telegram; fi` | ⚠️ частично (кейс — из 09-04-01) | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

Все `<automated>`-команды взяты дословно из verify-блоков планов 09-01..09-04; каждая несёт `<fails_when>` с наблюдаемым сигналом отказа.

---

## Wave 0 Requirements

- [ ] `internal/session/probe.go` + `internal/session/probe_test.go`, расширения `internal/ctlsvc/ctlsvc_test.go`, `cmd/goswitchctl/main_test.go` — красный корпус до кода (09-02-01, strict TDD)
- [ ] `test/e2e/wezterm.go` + `test/e2e/fixtures/gs-e2e-wezterm-{xim,wayland}.lua` + `test/e2e/case_spike_wezterm.go` + `test/e2e/case_spike_caps.go` + регистрация в `test/e2e/main.go` + mise-задачи `e2e-spike-smoke` / `e2e-spike-wezterm` / `e2e-spike-caps` / `e2e-wezterm-flip` (09-02)
- [ ] `test/e2e/case_spike_a11y.go` + регистрация + mise-задача `e2e-spike-a11y` (09-04)
- [ ] `docs/adr/adr-008-spike-log.md`, `docs/adr/d09-a11y-probe-log.md` (шапки + таблицы — создаются задачами 09-02-01 / 09-04-01 до первых живых прогонов)

*Ни одна automated-команда планов не ссылается на файл, который не создаётся более ранней задачей той же цепочки (волновой порядок 09-01 → 09-02 → 09-03; 09-01 → 09-04). Отдельного Wave-0-плана не требуется.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Шесть входных пунктов владельца (OQ 1/4/6 — ветка фикса, право правки `~/.wezterm.lua`, пред-вердикт сужения строки) | TERM-01 / TERM-03-рамка | решения владельца, не автоматы — checkpoint Task 1 плана 09-03 | Предъявить сводку `adr-008-spike-log.md`; зафиксировать ответы строкой журнала |
| Судьба зондового глагола probe-forward (keep/remove) | — (фазовый гейт) | решение владельца — checkpoint Task 3 плана 09-03 | Ответ «verb-keep/verb-remove»; удаление — только после завершения 09-04 |
| Flatpak-согласие + ScreenReaderEnabled-семантика | A11Y-02 (spike) | установка второй копии Telegram и сессионная семантика — решения владельца — checkpoint Task 2 плана 09-04 | Ответы строками в `d09-a11y-probe-log.md`; без согласия ряд deferred |
| Повседневный флоу wezterm после фикса | TERM-01 | живой ежедневный инструмент владельца, не стенд | Флип по горячей клавише + набор в обеих раскладках в окнах владельца; подтверждение перед done 09-03-02 |
| SPEC/ADR содержательный ревью-гейт (вето TERM-04 сформулировано честно) | TERM-04 | качество формулировки, не греп | Владелец ревьюит rev D-55 на `/gsd:verify-work` |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90 s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
