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
| 08-09 T1 | 08-09 | 6 | D-8-1/REV/D-8-5/REV/D-8-6/REV | T-08-09-01/02/05 | SPEC-ревизия 2026-10-06 add-only (цитата владельца, аудит-трейл цел), CONFIG/SKILL в bool-семантике без ключа-списка, миграционная заметка, golden-граница 20 | unit (grep-гейты) + golden | `cd /home/nil/DiskD/W/Djarvur/goswitch && grep -q "2026-10-06" docs/SPEC.md && grep -q "список не нужен" docs/SPEC.md && grep -q '\| `a11y.enabled` \| bool \| `true` \|' docs/CONFIG.md && mise exec -- go test ./cmd/skillgen/ -race -count=1 && mise exec -- go run ./cmd/skillgen -check && echo T1-REV-OK` | ✅ | ✅ green |
| 08-09 T2 | 08-09 | 6 | D-8-1/REV/D-8-6/REV | T-08-09-02 | pointer-bool nil = ON (прецедент Sound), прежний ключ-список отвергается strict decode целиком (громкая миграция D-33), load.go/watch.go не тронуты | unit (RED→GREEN) | `cd /home/nil/DiskD/W/Djarvur/goswitch && mise exec -- go test ./internal/config/ -race -count=1 && grep -q "func (a A11y) EffectiveEnabled" internal/config/config.go && ! grep -qE "maxA11yApps\|errA11yApps" internal/config/config.go && echo SCHEMA-BOOL-OK` | ✅ | ✅ green |
| 08-09 T3 | 08-09 | 6 | D-8-1/REV/D-8-6/REV | T-08-09-03/04 | фолд читает EffectiveEnabled: defaults-документ → push true, явное false → push false, дифф-гейт сохранён; reconciler — только doc-комменты; install не тронут | unit + full | `cd /home/nil/DiskD/W/Djarvur/goswitch && mise exec -- go test ./internal/session/ ./internal/a11y/ -race -count=1 && mise run ci && echo FOLD-CI-OK` | ✅ | ✅ green |
| 08-09 T4 | 08-09 | 6 | D-8-1/REV/D-8-5/REV/D-8-6/REV | T-08-09-01/02/05 | полная зелёная итерация фазы, skillgen -check байт-в-байт, пользовательские доки/skill чисты от ключа-списка, схема чиста от валидации списка | full | `cd /home/nil/DiskD/W/Djarvur/goswitch && mise run ci && mise exec -- go run ./cmd/skillgen -check && ! grep -q "a11y\.apps" docs/CONFIG.md && ! grep -q "a11y\.apps" skills/goswitch-config/SKILL.md && echo FINAL-REV-OK` | ✅ | ✅ green |
| 08-10 T1 | 08-10 | 7 | D-8-8/D-8-7 | T-08-10-01 | README.md a11y-раздел на bool-контракте ревизии 2026-10-06 (зеркало docs/CONFIG.md «Секция a11y: применение и откат»): единственный ключ a11y.enabled (default ON — отсутствующая секция/ключ читаются ВКЛ; явное enabled: false — единственный OFF, D-8-4), глобальная механика (toolkit-accessibility + пояс IsEnabled), идемпотентный reconcile при старте и hot reload, restart-семантика, uninstall-revert, громкая миграция; ноль ключа-списка и ноль прозы сопоставления | unit (grep-гейты) | `cd /home/nil/DiskD/W/Djarvur/goswitch && grep -q "a11y\.apps" README.md && echo RU-STALE; ! grep -qE "перечисляет приложения\|добавления его в список\|приложения из списка" README.md && grep -q "a11y.enabled" README.md && grep -q "toolkit-accessibility" README.md && grep -q "IsEnabled" README.md && grep -q "enabled: false" README.md && grep -q "goswitchctl uninstall" README.md && grep -qE "следующ(ем\|его) запуск" README.md && grep -qF "единственный ключ \`a11y.enabled\`" README.md && grep -qF "## A11y-магия приложений" README.md && echo RU-A11Y-BOOL-OK` | ✅ | ✅ green |
| 08-10 T2 | 08-10 | 7 | D-8-8/D-8-7 | T-08-10-02 | README.en.md — тот же bool-контракт по-английски, секция-в-секцию параллельно README.md (те же четыре блока в том же порядке); паритет заголовков 22/22 обоих README; ноль ключа-списка в обоих | unit (grep-гейты) | `cd /home/nil/DiskD/W/Djarvur/goswitch && grep -q "a11y\.apps" README.en.md && echo EN-STALE; ! grep -qE "lists the applications\|adding it to the list\|application from the list" README.en.md && grep -q "a11y.enabled" README.en.md && grep -q "toolkit-accessibility" README.en.md && grep -q "IsEnabled" README.en.md && grep -q "enabled: false" README.en.md && grep -q "goswitchctl uninstall" README.en.md && grep -qE "next (start\|launch)" README.en.md && grep -qF "the single key \`a11y.enabled\`" README.en.md && grep -qF "## App accessibility magic (a11y)" README.en.md && test "$(grep -cE '^##+ ' README.md)" -eq 22 && test "$(grep -cE '^##+ ' README.en.md)" -eq 22 && echo EN-A11Y-BOOL-OK` | ✅ | ✅ green |
| 08-10 T3 | 08-10 | 7 | D-8-8 | T-08-10-03/SC | скоуп-дифф README целиком внутри прежних границ a11y-разделов (RU старые 301–320, EN старые 296–314; awk-гейт исполнен в mawk-совместимой форме — split по «,» без пробельного разделителя, [Rule 3], см. 08-10-SUMMARY), кросс-ссылки шапок целы, skillgen -check exit 0 (генератор читает только docs/CONFIG.md — cmd/skillgen/main.go:31 — README не подвластны), mise run ci зелёная, строки 08-10 T1..T3 в этой карте | full + scope-diff | `cd /home/nil/DiskD/W/Djarvur/goswitch && BASE=$(git rev-parse --verify -q gsd-plan-head-before-08-10) && git diff -U0 "$BASE..HEAD" -- README.md \| awk '/^@@/{r=$2; sub(/^-/,"",r); split(r,a,","); s=a[1]+0; c=a[2]+0; if (s<301 \|\| s+c-1>320) bad=1} END{exit bad}' && git diff -U0 "$BASE..HEAD" -- README.en.md \| awk '/^@@/{r=$2; sub(/^-/,"",r); split(r,a,","); s=a[1]+0; c=a[2]+0; if (s<296 \|\| s+c-1>314) bad=1} END{exit bad}' && mise exec -- go run ./cmd/skillgen -check && ! grep -q "a11y\.apps" README.md && ! grep -q "a11y\.apps" README.en.md && test "$(grep -cE '^##+ ' README.md)" -eq 22 && test "$(grep -cE '^##+ ' README.en.md)" -eq 22 && mise run ci && echo GAP-README-OK` | ✅ | ✅ green |

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
