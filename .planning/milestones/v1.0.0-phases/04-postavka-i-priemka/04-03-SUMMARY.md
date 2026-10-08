---
phase: 04-postavka-i-priemka
plan: 03
subsystem: release
tags: [goreleaser, github-actions, ci, release-engineering, version-stamping, mise, checksums]
requires:
  - phase: 04-postavka-i-priemka (04-02)
    provides: cmd/goswitchd vars version/commit/date (dev fallback) — the ldflags stamping receivers (D-37)
  - phase: 01 (01-05)
    provides: mise.toml [tools] convention and the mise-install-in-CI precedent (D-09/D-11)
provides:
  - .goreleaser.yaml — two builds (goswitchd, goswitchctl) linux/amd64, ONE tar.gz with both binaries, checksums.txt, changelog filters; version stamping rides goreleaser DEFAULT ldflags (no custom ldflags field)
  - .github/workflows/release.yml — tag v* → mise install → goreleaser release --clean; permissions contents: write scoped to THIS workflow only; concurrency group release without cancel-in-progress
  - goreleaser = "2.18.1" pinned in mise.toml [tools] — single version source for local and CI
  - Snapshot proof on this tree: goreleaser check green, build+release dry-run produce both binaries, goswitchd -version prints 0.0.0-SNAPSHOT-<sha> (≠ dev)
affects: [04-06 (double matrix gate), 04-07 (first v1.0.0 tag, README install docs), verify-work UAT]
tech-stack:
  added: ["goreleaser 2.18.1 (mise [tools] pin — CI/build tool, NEVER a Go dependency; go.mod/go.sum untouched, D-38)"]
  patterns:
    - "release tool pinned in mise.toml [tools], CI installs via mise install — one version source (D-09/D-11, Pitfall 10)"
    - "goreleaser dist redirected away from ./dist when ./dist holds tracked sources — .goreleaser-dist/ gitignored"
    - "write-scope containment: contents: write on the release workflow ONLY, grep-gated against sibling workflows (T-04-03-02)"
key-files:
  created:
    - .goreleaser.yaml
    - .github/workflows/release.yml
  modified:
    - mise.toml
    - .gitignore
key-decisions:
  - "Discretion A7 fixed in config comments: v1 is linux/amd64-only (Ubuntu 24.04 target; +arm64 on owner request), tag scheme semver with first release v1.0.0 (tag pushed by 04-07), assets = one tar.gz with BOTH binaries + sha256 checksums.txt"
  - "No custom ldflags anywhere: goreleaser defaults inject -X main.version/main.commit/main.date -X main.builtBy into the 04-02 vars; go install builds honestly report dev (D-37/D-38)"
  - "goreleaser archive `formats: [tar.gz]` (v2 current schema) instead of the research quote's `format:` — goreleaser check 2.18.1 validates with zero deprecation warnings"
  - "checkout fetch-depth: 0 in release.yml — goreleaser builds the changelog from commit history and stamps main.commit from git; a shallow checkout would publish a broken changelog/identity"
patterns-established:
  - "Tag-gated release: nothing publishes from branch pushes; the release workflow resolves from the default branch only (02-07 precedent) — first tag is 04-07's job after green gates"
requirements-completed: []  # INST-01 stays open across sibling plans; see requirements step
duration: 7 min
completed: 2026-09-16
actuals:
  tokens: 1853
  tasks: 2
  commits: 2
plan_head_before: d69415e53106bdb3c310c6bbad4b5e64b7cdcdf1
commits: 2
status: complete
---

# Phase 4 Plan 03: Релизный конвейер Summary

**goreleaser 2.18.1 (mise-pinned) собирает оба бинарника с дефолтной прошивкой версии в vars из 04-02 — snapshot-пруф `goswitchd 0.0.0-SNAPSHOT-<sha>`, один tar.gz + checksums.txt — и release-workflow на тег v* держит contents: write только в себе.**

## Performance

- **Duration:** 7 min
- **Started:** 2026-09-16T13:55:42Z
- **Completed:** 2026-09-16T14:02:22Z
- **Tasks:** 2
- **Files modified:** 4 (2 created, 2 modified)

## Accomplishments

- **D-38 исполнен без единой новой зависимости демона**: `.goreleaser.yaml` — version 2, builds goswitchd (./cmd/goswitchd) + goswitchctl (./cmd/goswitchctl), linux/amd64, ОДИН tar.gz с обоими бинарниками (пара рядом — install резолвит демона через dirname(os.Executable)), checksums.txt (sha256), changelog asc с фильтрами ^docs:/^test:/^chore:. `go mod tidy`-диф пуст (TIDY-CLEAN) — goreleaser остался CI-инструментом.
- **D-37 прошивается дефолтами goreleaser**: в конфиге НЕТ поля ldflags — дефолтные `-X main.version/main.commit/main.date -X main.builtBy` дошли до vars из 04-02: snapshot-бинарник печатает `goswitchd 0.0.0-SNAPSHOT-d69415e <full-sha> <date>` (≠ dev); `go install`-канал собирается как есть и честно показывает dev.
- **Полномасштабный snapshot-пруф**: не только `build --snapshot` (оба бинарника в dist), но и `release --snapshot --clean` — архив `goswitch_0.0.0-SNAPSHOT-d69415e_linux_amd64.tar.gz` содержит goswitchd + goswitchctl + LICENSE + README, checksums.txt рядом.
- **Release-workflow**: тег v* → checkout (fetch-depth: 0) → mise install (goreleaser из пина mise.toml — локально и CI одна версия, D-09/D-11) → `mise exec -- goreleaser release --clean` под GITHUB_TOKEN; permissions contents: write ТОЛЬКО здесь — SCOPE-OK: pr-sanity.yml (3×) и e2e-matrix.yml (2×) остались contents: read без contents: write; concurrency group `release` без cancel-in-progress.

## TDD Gate Compliance

Plan 04-03 is `type: execute` with both tasks `type="auto"` config/CI tasks (no `tdd="true"`, no `<behavior>` blocks, no Go source files) — the RED/GREEN/REFACTOR gate does not apply per tdd.md `when_to_use_tdd` (configuration changes skip TDD). Green iteration D-08 enforced instead: `mise run ci` green on both task commits.

| Task | RED | GREEN | REFACTOR | Status |
|------|-----|-------|----------|--------|
| 1 (.goreleaser.yaml + mise pin) | n/a (config) | ✓ adc4303 | — | n/a |
| 2 (release.yml) | n/a (CI config) | ✓ 8372348 | — | n/a |

## Task Commits

Each task was committed atomically:

1. **Task 1: .goreleaser.yaml (две сборки, дефолтная прошивка) + goreleaser-пин в mise + snapshot-доказательство** - `adc4303` (feat)
2. **Task 2: .github/workflows/release.yml — тег v* → goreleaser release, contents:write только здесь** - `8372348` (feat)

**Plan metadata:** the docs commit following this SUMMARY.

## Files Created/Modified

- `.goreleaser.yaml` — release pipeline config: two builds, linux/amd64 (A7 comment), default-ldflags stamping (D-37), one tar.gz, checksums.txt, changelog filters; `dist: .goreleaser-dist` with the WHY comment
- `.github/workflows/release.yml` — tag v* workflow: contents: write scoped here, mise install, goreleaser release --clean, GITHUB_TOKEN only; header cites D-38/Падение 10/A7/02-07
- `mise.toml` — `[tools] goreleaser = "2.18.1"` with the not-a-daemon-dependency comment
- `.gitignore` — `.goreleaser-dist/` entry (goreleaser build output; tracked `dist/` packaging sources protected)

## Decisions Made

- **Discretion A7 зафиксирован в комментариях конфига**: v1 — amd64-only (целевая платформа Ubuntu 24.04 amd64; +arm64 по запросу владельца), теги semver, первый релиз v1.0.0 (тег ставит 04-07), ассеты tar.gz+checksums.
- **`formats: [tar.gz]`** — текущий ключ схемы v2 вместо дословного `format:` из research-цитаты; `goreleaser check` 2.18.1 валиден без deprecation-предупреждений.
- **`fetch-depth: 0`** в checkout release-workflow — changelog и штамп main.commit требуют полной истории (стандартное требование goreleaser).
- **Workflow-резолюция от default branch** (прецедент 02-07): до мержа фазы тег не публикуется — заметка в шапке release.yml; это же требование плана (FLAGGED assumption).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] goreleaser dist redirected to `.goreleaser-dist/`**
- **Found during:** Task 1 (before running the snapshot command)
- **Issue:** the plan's verify literally runs `goreleaser build --snapshot --clean && find dist …`, but this repo's `./dist` holds TRACKED packaging sources (`dist/systemd/user/goswitchd.service`, plan 01-05; `.gitignore` documents the intent) — a `--clean` run wipes the whole dist directory and would delete a tracked file.
- **Fix:** `dist: .goreleaser-dist` in `.goreleaser.yaml` + `.goreleaser-dist/` in `.gitignore`; verify `find` retargeted to the configured dir; `git status -- dist/` verified clean after both snapshot runs.
- **Files modified:** .goreleaser.yaml, .gitignore
- **Verification:** snapshot build + release dry-run green; `git status --short -- dist/` empty; no deletions in either task commit
- **Committed in:** adc4303 (Task 1 commit)

**2. [Rule 2 - Missing Critical] `fetch-depth: 0` on the release checkout**
- **Found during:** Task 2
- **Issue:** the plan's step list copied pr-sanity's bare `actions/checkout@v4`; goreleaser builds the changelog from git history and stamps `main.commit`/`main.date` (D-37) from the git state — a shallow checkout would publish releases with a truncated changelog and broken commit identity (T-04-03-01 adjacent: trust in the release channel).
- **Fix:** `with: fetch-depth: 0` plus the rationale comment in release.yml.
- **Files modified:** .github/workflows/release.yml
- **Verification:** structural greps RELEASE-WF-OK; standard goreleaser release requirement
- **Committed in:** 8372348 (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 blocking, 1 missing-critical).
**Impact on plan:** both fixes protect tracked sources and the release channel's integrity; no scope creep. The plan's semantic acceptance criteria are all met as written.

## Issues Encountered

None beyond the deviations above. The python-yaml bonus sanity check was unavailable on this box (no module) — the plan's own verify gates (greps + `mise run ci`) are green, and the file mirrors pr-sanity.yml's proven structure.

## Verification Results

- `mise exec -- goreleaser check` — PASS (1 configuration file(s) validated)
- `mise exec -- goreleaser build --snapshot --clean` — PASS (both binaries under .goreleaser-dist/); `find .goreleaser-dist -name goswitchd -type f` non-empty
- Snapshot stamping: `.goreleaser-dist/goswitchd_linux_amd64_v1/goswitchd -version` → `goswitchd 0.0.0-SNAPSHOT-d69415e d69415e53… 2026-09-16T14:00:06Z`, exit 0
- `mise exec -- goreleaser release --snapshot --clean` — PASS: one tar.gz with goswitchd+goswitchctl, checksums.txt (sha256)
- `mise exec -- go mod tidy && git diff --exit-code -- go.mod go.sum` — TIDY-CLEAN (D-38)
- RELEASE-WF-OK grep gate — PASS (tags v*, contents: write, goreleaser release --clean, mise install)
- SCOPE-OK grep gate — PASS (pr-sanity.yml / e2e-matrix.yml keep contents: read, no contents: write)
- `mise run ci` — PASS on both task commits (build + vet + golangci-lint 0 issues + test -race)

## User Setup Required

None - no external service configuration required. (The real publish needs only a `v*` tag — plan 04-07 after green gates; GITHUB_TOKEN is implicit in Actions.)

## Next Phase Readiness

- Both INST-01 channels are pipeline-ready: GitHub Releases (tar.gz + checksums via goreleaser) and `go install github.com/Djarvur/goswitch/cmd/goswitchd@vX.Y.Z` (module unchanged).
- 04-06 (double matrix gate) and 04-07 (first v1.0.0 tag, README) remain in the phase; release.yml resolves from the default branch only after the phase merge — sequencing held by design.
- Perf budget note from 04-04 stands (injector-dominated latency, owner decision pending) — orthogonal to this plan, tracked in 04-04-SUMMARY.

---
*Phase: 04-postavka-i-priemka*
*Completed: 2026-09-16*

## Self-Check: PASSED

- Files exist: .goreleaser.yaml, .github/workflows/release.yml, mise.toml (modified), .gitignore (modified) — all FOUND
- Commits: adc4303, 8372348 — both FOUND on gsd/phase-04-postavka-i-priemka (rev-list from ledger gsd-plan-head-before-04-03 = 2)
- Acceptance criteria re-run: goreleaser check green; snapshot both binaries + non-dev -version; tidy-diff empty; RELEASE-WF-OK + SCOPE-OK greps; mise run ci green
