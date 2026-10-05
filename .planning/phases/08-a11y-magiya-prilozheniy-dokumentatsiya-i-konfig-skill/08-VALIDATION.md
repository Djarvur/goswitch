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
| 08-01 T1 | 08-01 | 1 | D-8-1/2/3/4/6 | T-08-01-01/02 | ревизия §4 только добавлениями, полный контракт a11y | unit (grep-гейты) | `cd /home/nil/DiskD/W/Djarvur/goswitch && grep -c "2026-10-05" docs/SPEC.md \| grep -qx "[1-9]" && grep -q "toolkit-accessibility" docs/SPEC.md && echo S4-OK` | ✅ (docs/SPEC.md) | ⬜ pending |
| 08-01 T2 | 08-01 | 1 | D-8-* | T-08-01-SC | зелёная итерация docs-only | full suite | `cd /home/nil/DiskD/W/Djarvur/goswitch && mise run ci` | ✅ | ⬜ pending |
| 08-02 T1 | 08-02 | 2 | D-8-1/D-8-5 | T-08-02-01/02/03 | потолок безусловен, blank-отказ, strict decode | unit (RED→GREEN) | `cd /home/nil/DiskD/W/Djarvur/goswitch && mise exec -- go test ./internal/config/ -race -count=1` | ❌ → создаётся задачей (корпус) | ⬜ pending |
| 08-02 T2 | 08-02 | 2 | D-8-5 | T-08-02-01 | hot-reload last-good на битом паттерне | unit + full | `cd /home/nil/DiskD/W/Djarvur/goswitch && mise exec -- go test ./internal/config/ -race -count=1 -run TestWatch_BrokenA11yPatternKeepsLastGood -v` | ❌ → Wave 0 задача 08-02 T2 | ⬜ pending |
| 08-03 T1 | 08-03 | 2 | D-8-3/D-8-6 | T-08-03-01/02/03/05 | литеральный argv, read-verify-then-set, снятие запрещено, no-pipes | unit (Runner seam) | `cd /home/nil/DiskD/W/Djarvur/goswitch && mise exec -- go test ./internal/a11y/ -race -count=1` | ❌ → Wave 0 (новый пакет) | ⬜ pending |
| 08-03 T2 | 08-03 | 2 | D-8-6 | T-08-03-01/04 | пояс IsEnabled, warn-once эпизоды, budget reopen | unit (StatusSetter seam) | `cd /home/nil/DiskD/W/Djarvur/goswitch && grep -q "org.a11y.Bus" internal/a11y/a11y.go && grep -q "IsEnabled" internal/a11y/a11y.go && echo BELT-OK` | ❌ → Wave 0 задача 08-03 T1 | ⬜ pending |
| 08-04 T1 | 08-04 | 2 | D-8-4 | T-08-04-04 | снапшот verbatim, отказ → пусто+WARN, idempotent-backup цел | unit (fakeRunner) | `cd /home/nil/DiskD/W/Djarvur/goswitch && mise exec -- go test ./internal/install/ -race -count=1` | ❌ → Wave 0 корпус | ⬜ pending |
| 08-04 T2 | 08-04 | 2 | D-8-4 | T-08-04-01/02 | restore only-if-present, shape true\|false, никогда не «restore false» | unit + full | `cd /home/nil/DiskD/W/Djarvur/goswitch && grep -q "restoreToolkitAccessibility" internal/install/install.go && grep -q "nothing to revert" internal/install/install.go && echo RESTORE-OK` | ❌ → Wave 0 задача 08-04 T2 | ⬜ pending |
| 08-05 T1 | 08-05 | 3 | D-8-1/D-8-3 | T-08-05-01/02/03 | неблокирующий sink, дифф-гейт, self-sync, nil no-op | unit (fake sink) | `cd /home/nil/DiskD/W/Djarvur/goswitch && mise exec -- go test ./internal/session/ -race -count=1` | ❌ → Wave 0 корпус | ⬜ pending |
| 08-05 T2 | 08-05 | 3 | D-8-3 | T-08-05-01 | wiring без новых watcher-веток | build+vet+full | `cd /home/nil/DiskD/W/Djarvur/goswitch && mise exec -- go build ./... && mise exec -- go vet ./... && echo BUILD-VET-OK` | ✅ | ⬜ pending |
| 08-06 T1 | 08-06 | 4 | D-8-8/D-8-9 | T-08-06-01 | русский-первый, все 19 ключей на месте, операционные доки EN | unit (grep-гейты) | `cd /home/nil/DiskD/W/Djarvur/goswitch && for k in hotkeys.tap_key sound.autocorrect_event; do grep -q "\`$k\`" docs/CONFIG.md \|\| exit 1; done && echo ALL-KEYS-PRESENT` | ✅ | ⬜ pending |
| 08-06 T2 | 08-06 | 4 | D-8-6/D-8-1 | T-08-06-02/03/04 | семь секций, 2 a11y-строки, restart-семантика, рецепт | unit (grep-гейты) | `cd /home/nil/DiskD/W/Djarvur/goswitch && test "$(grep -c '^| `' docs/CONFIG.md)" -ge 21 && echo ROWS-SEVEN-SECTIONS` | ✅ | ⬜ pending |
| 08-07 T1 | 08-07 | 5 | D-8-7/D-8-8 | T-08-07-01 | языковой swap, все разделы, README.ru.md удалён | unit (grep-гейты) | `cd /home/nil/DiskD/W/Djarvur/goswitch && head -40 README.md \| grep -qP "[А-Яа-яЁё]{3,}" && git ls-files -- README.ru.md \| grep -q .; test $? -eq 1 && echo LANG-SWAP-DELETE-OK` | ✅ | ⬜ pending |
| 08-07 T2 | 08-07 | 5 | D-8-7 | T-08-07-02 | нет stale-ссылок на README.ru.md | unit (grep-гейты) | `cd /home/nil/DiskD/W/Djarvur/goswitch && grep -rn "README\.ru" README.md README.en.md docs/ 2>/dev/null \| grep -q .; test $? -eq 1 && echo NO-STALE-RU` | ✅ | ⬜ pending |
| 08-08 T1 | 08-08 | 5 | D-8-10 | T-08-08-01/02 | hard errors, детерминизм, сентинелы, -check | unit (golden) | `cd /home/nil/DiskD/W/Djarvur/goswitch && mise exec -- go test ./cmd/skillgen/ -race -count=1` | ❌ → Wave 0 (новый пакет) | ⬜ pending |
| 08-08 T2 | 08-08 | 5 | D-8-10/D-8-11 | T-08-08-03 | SKILL.md tracked, regen идемпотентен, check зелёный | unit + full | `cd /home/nil/DiskD/W/Djarvur/goswitch && git ls-files -- skills/goswitch-config/SKILL.md \| grep -q . && mise run skillgen-regen && mise exec -- go run ./cmd/skillgen -check && echo CHECK-GREEN` | ❌ → создаётся задачей | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

**Wave 0 note:** строки с ❌ — кейсы, создаваемые самими задачами планов (RED→GREEN внутри задачи, TDD-дисциплина проекта); отдельный Wave 0-план не нужен — ни один `<automated>` не ссылается на несуществующий файл ДРУГОЙ задачи без `depends_on`/внутрипланового порядка (прецедент 06-02 dictgen).

---

## Wave 0 Requirements

- [x] Существующая инфраструктура покрывает go test/lint/mise — новый фреймворк не нужен
- [x] RED-корпуса новых пакетов (internal/a11y, cmd/skillgen) и новых секций (config a11y, install a11y, session фолд) создаются внутри самих задач планов 08-02..08-05/08-08 как первые RED→GREEN шаги (TDD-директива проекта) — отдельный Wave 0-план не нужен

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
