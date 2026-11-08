---
phase: quick-261108-qwa
plan: 01
subsystem: release
tags: [goreleaser, arm64, release-binaries, readme]
requires:
  - goreleaser 2.18.1 (mise pin, Phase 4)
  - release.yml CI workflow (unchanged)
provides:
  - релизный канал linux/arm64 (оба бинарника в каждом пер-архитектурном архиве)
  - README RU/EN честно называют обе архитектуры и оба имени архива
affects:
  - будущие релизные теги (следующий тег публикует 2 архива вместо 1)
tech-stack:
  added: []
  patterns:
    - конфиг .goreleaser.yaml остаётся единственным источником архитектур (CI-матрицы нет)
key-files:
  created: []
  modified:
    - .goreleaser.yaml
    - README.md
    - README.en.md
decisions:
  - arm64 добавлен ровно по запросу владельца 2026-11-08 (A7 «+ nothing else»); goos остаётся [linux], никаких darwin/windows/386/armv7
  - каждый пер-архитектурный архив несёт ОБА бинарника (default name_template semantics не менялись)
metrics:
  duration: 11 min
  completed: 2026-11-08
  tasks: 2
  commits: 2
status: complete
actuals:
  tokens: 985
  tasks: 2
  commits: 2
plan_head_before: 563e58db7d7a4f7a00e7b1ddc68128370524c6eb
---

# Quick Task 261108-qwa: linux/arm64 в релизных бинарниках — Summary

Релизный конвейер goswitch собирает и публикует linux/arm64 рядом с linux/amd64: `goarch: [amd64, arm64]` на обеих сборках (goswitchd, goswitchctl), комментарии A7 переведены в исполненное состояние с датой запроса, README RU/EN называют обе архитектуры в требованиях, оба имени архива и универсальную маску tar.

## What Was Built

**Task 1 — .goreleaser.yaml (commit 9240d28):**
- `goarch: [amd64, arm64]` в билдах goswitchd и goswitchctl; `goos: [linux]` нетронут; новых ключей нет (A7 «+ nothing else» соблюдён дословно).
- Заголовочный discretion-блок и инлайн-комментарий у goarch первого билда фиксируют исполнение: arm64 добавлен по запросу владельца 2026-11-08, за пределами двух архитектур расширения нет.
- Verify: `goreleaser check` зелёный; `release --snapshot --clean` собирает 4 бинарника (2 билда × 2 арх) и 2 архива `goswitch_1.2.1-SNAPSHOT-563e58d_linux_{amd64,arm64}.tar.gz`; arm64-архив содержит goswitchd + goswitchctl + LICENSE + оба README; arm64-бинарники — настоящие статические ELF aarch64; artifacts.json фиксирует `"goarch":"arm64"` в 3 записях (оба бинарника + архив); checksums.txt покрывает ОБА архива (T-quick-qwa-01).

**Task 2 — README.md + README.en.md (commit b485d2f):**
- Требования: `linux/amd64 или linux/arm64` / `linux/amd64 or linux/arm64` с квалификацией во множественном числе, пара на одной строке.
- Канал А, шаг 1: выбор архива по архитектуре процессора, ОБА имени файла названы, каждое целиком на одной строке; хвост про содержимое архива нетронут.
- Код-блок шага 2: универсальная маска `goswitch_*_linux_*.tar.gz`; старая маска одной архитектуры исчезла из обоих файлов.
- Все grep-гейты плана зелёные; `mise run ci` (build + vet + lint + test -race) зелёный.

## Verification Results

| Gate | Result |
|------|--------|
| `mise exec -- goreleaser check` | PASS |
| `goreleaser release --snapshot --clean` | PASS (2 архива, 4 бинарника, 3s) |
| arm64-архив содержит оба бинарника | PASS (LICENSE, README.en.md, README.md, goswitchctl, goswitchd) |
| artifacts.json ≥3 записей goarch=arm64 | PASS (ровно 3: 2 бинарника + архив) |
| checksums.txt покрывает оба архива | PASS |
| grep-гейты README RU/EN (10 проверок) | PASS |
| `mise run ci` | PASS (35.65s) |
| diff строго в 3 файлах плана | PASS (.goreleaser.yaml, README.md, README.en.md) |

Живая публикация arm64 произойдёт ближайшим релизным тегом через release.yml (`mise exec -- goreleaser release --clean` без матрицы — конфиг единственный источник); тег в скоуп задачи не входит.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking env] Хостовый конфликт toolchain: профиль экспортирует GOROOT mise-go-1.23.12 при /usr/local/go/bin (go1.27.3) раньше mise-бинов в PATH**
- **Found during:** Task 1 verify
- **Issue:** `mise exec -- goreleaser release --snapshot --clean` падает «compile: version "go1.23.12" does not match go tool version "go1.27.3"»: в песочнице этого агента PATH не содержит mise-шеймов, goreleaser резолвит `go` в /usr/local/go/bin (драйвер 1.27.3), а экспортированный GOROOT указывает на дерево 1.23.12. Подтверждено strace (execve /usr/local/go/bin/go ×6 при корректном PATH процесса; mise-шим перестраивает PATH для порождённых инструментов). В CI (release.yml) конфликт невозможен: нет системного go и экспортированного GOROOT; в активированной оболочке владельца шимы mise стоят первыми.
- **Fix:** для прогона verify перед PATH поставлен сквозной индирект-каталог с исполняемым `go`, тождественно exec'ающим mise-пинованный `/home/nil/.local/share/mise/installs/go/1.23.12/bin/go` (тот же бинарник, что даёт `mise exec` в активированной оболочке). Семантика команды плана сохранена: goreleaser 2.18.1 (mise), toolchain go 1.23.12 (mise-пин). То же решение применено к `mise run ci` (без него падает той же ошибкой). Файлы проекта не менялись.
- **Files modified:** none (env-only)
- **Commit:** n/a

**2. [Rule 1 - Plan bug] Литеральная форма artifacts-гейта считала строки, а не вхождения**
- **Found during:** Task 1 verify
- **Issue:** гейт `grep -c '"goarch":"arm64"' artifacts.json` ждёт ≥3, но artifacts.json goreleaser — однострочный compact JSON: `grep -c` даёт 1 при любых 3+ вхождениях.
- **Fix:** гейт исполнен в форме подсчёта вхождений (`grep -o | wc -l` = 3 ≥ 3) — намерение плана («обоих бинарников + архив») выполнено и подтверждено.
- **Files modified:** none
- **Commit:** n/a

**3. [Rule 1 - Plan bug] Предложенная EN-формулировка противоречила собственному негативному гейту плана**
- **Found during:** Task 2
- **Issue:** план предлагал «the release binaries' architectures», но гейт `! grep -qF "the release binaries' architecture"` запрещен длины: множественное число содержит единственное как префикс — буквальная формулировка не может пройти гейт.
- **Fix:** использована эквивалентная квалификация множественного числа «(architectures of the release binaries)» — смысл сохранён, гейт зелёный.
- **Files modified:** README.en.md
- **Commit:** b485d2f

## Auth Gates

None.

## Known Stubs

None.

## Self-Check: PASSED

- Commits verified: 9240d28 (Task 1), b485d2f (Task 2) on quick/261108-qwa-arm64.
- Files verified: .goreleaser.yaml, README.md, README.en.md modified; no other files touched (git diff 563e58d..HEAD = 3 files).
- No deletions, no untracked files left behind.
- Plan diff scope exact: 16 insertions, 10 deletions across the 3 planned files.
