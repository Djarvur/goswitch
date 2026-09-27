---
phase: 02-korrektsiya-slova-en-ru
plan: 07
subsystem: infra
tags: [ci, self-hosted-runner, github-actions, gnome, e2e-matrix, d-19, test-04, systemd-user-unit]

requires:
  - phase: 02-korrektsiya-slova-en-ru
    provides: mise task e2e-matrix (16-case D-18 YAML matrix, PASS/FAIL, exit≠0, e2e-report.txt) from plan 02-06
provides:
  - Workflow e2e-matrix.yml (workflow_dispatch-only, runs-on [self-hosted, gnome], concurrency e2e-gnome, contents: read) registered on the default branch
  - Live self-hosted runner (green106, label gnome, actions-runner 2.337.0) in the owner's graphical session via systemd user unit — first dispatch run green 16/16
  - docs/ci-runner.md — reproducible runner book (install/unit/env/bootstrap-registration/update/remove/desktop remedies)
  - SECURITY.md — phase-2 STRIDE threat register incl. the CI-runner block
affects: [verify-work, phase-3-regression-loop, phase-4-acceptance]

actuals:
  tokens: 6726    # chars/4 over the realized diff (4 files, +357)
  tasks: 2
  commits: 2      # measured: git rev-list --count 413dbc9..HEAD
  plan_head_before: 413dbc91db335de35c14196a66d0f299d77e62cd

tech-stack:
  added: ["actions-runner v2.337.0 (linux-x64, external CI infrastructure — not a Go dependency)"]
  patterns:
    - "declared first-dispatch sequence verified live: GitHub resolves workflow_dispatch workflows from the DEFAULT branch (pre-registration dispatch → HTTP 404 'not found on the default branch'); bootstrap = chore-PR with strictly the single workflow file, then dispatch --ref <phase branch> (definition from default branch, checkout from ref)"
    - "runner in the graphical session as systemd USER unit after graphical-session.target (pattern of dist/systemd/user/goswitchd.service); session env (WAYLAND_DISPLAY/XDG_RUNTIME_DIR/DBUS_SESSION_BUS_ADDRESS) inherited from the user manager — GDM imports it at login, verified via /proc/<pid>/environ"

key-files:
  created:
    - .github/workflows/e2e-matrix.yml
    - docs/ci-runner.md
    - SECURITY.md
  modified:
    - test/e2e/README.md
    - docs/ci-runner.md

key-decisions:
  - "registration/remove-token mint requires gh api --method POST — plain gh api does GET and answers HTTP 404 (live finding, both commands fixed in the book)"
  - "runner environment import: GDM login import into the systemd user manager was sufficient — no EnvironmentFile needed; the unit inherits WAYLAND_DISPLAY/XDG_RUNTIME_DIR/DBUS_SESSION_BUS_ADDRESS directly"
  - "default branch is main, not master as the plan text said — identical bootstrap mechanism, terminology only"
  - "runner name green106 (hostname-derived), one runner with label gnome exactly — the verify gate pins count == 1"

patterns-established:
  - "CI dispatch bootstrap: minimal chore-PR (single workflow file, T-02-07-04) → merge → dispatch --ref feature branch; reusable for any future dispatch-only workflow"

requirements-completed: [TEST-04]

coverage:
  - id: D1
    description: "Workflow e2e-matrix.yml — dispatch-only trigger, self-hosted gnome label, concurrency e2e-gnome without cancel-in-progress, contents: read, mise run e2e-matrix step, e2e-report.txt artifact with if: always()"
    requirement: TEST-04
    verification:
      - kind: other
        ref: "command: typed yaml.v3 structural decode — name/on/permissions/concurrency/runs-on/steps asserted (STRUCT-OK)"
        status: pass
      - kind: e2e
        ref: "GitHub run 34913812782 — workflow registered, dispatched, concluded success"
        status: pass
    human_judgment: false
  - id: D2
    description: "docs/ci-runner.md — runner book: machine requirements, install with one-time token, systemd user unit with session env, enable/check/update/remove, bootstrap-registration of the workflow, desktop remedies, security section"
    verification:
      - kind: other
        ref: "command: plan grep gate — graphical-session/WAYLAND_DISPLAY/config.sh/bootstrap present (DOCS-OK)"
        status: pass
      - kind: manual_procedural
        ref: "the book was EXECUTED live for the actual bring-up (runner online, first run green) — POST fix folded back into the doc"
        status: pass
    human_judgment: false
  - id: D3
    description: "SECURITY.md — phase-2 STRIDE register from threat models of plans 02-01..02-07 with dispositions, dedicated CI-runner block (dispatch-only, permissions, concurrency, live-desktop printing accepted by owner D-19), accepted risks log"
    verification:
      - kind: other
        ref: "command: plan grep gate — registration/token present (DOCS-OK gate covers SECURITY.md)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Live runner in the owner's graphical session + first green dispatch run of the full matrix per the declared sequence (D-19, phase gate)"
    requirement: TEST-04
    verification:
      - kind: e2e
        ref: "command: systemctl --user is-active goswitch-ci-runner → active; gh api runners → exactly 1 gnome-labeled, online (RUNNER-ONLINE)"
        status: pass
      - kind: e2e
        ref: "command: gh run list/view — run 34913812782, event workflow_dispatch, headBranch gsd/phase-02-korrektsiya-slova-en-ru, conclusion success; artifact e2e-report downloaded to /tmp/e2e-ci-artifact/e2e-report.txt (16/16 PASS)"
        status: pass
      - kind: other
        ref: "command: mise run ci — triple gate green (after Task 1 and after Task 2)"
        status: pass
    human_judgment: false

duration: 13 min
completed: 2026-09-15
status: complete
---

# Phase 2 Plan 07: Self-hosted GitHub Actions runner на живой GNOME-машине Summary

**Живой CI-контур D-19: workflow e2e-matrix.yml (dispatch-only, [self-hosted, gnome], concurrency e2e-gnome) + раннер actions-runner 2.337.0 в графической сессии владельца (systemd user unit) — первый dispatch-прогон матрицы зелёный 16/16 по заявленной bootstrap-последовательности**

## Performance

- **Duration:** 13 min (2026-09-15T00:24:56Z → 00:38:10Z; сам CI-прогон на раннере — ~1 мин, матрица 16 кейсов за 51 с)
- **Started:** 2026-09-15T00:24:56Z
- **Completed:** 2026-09-15T00:38:10Z
- **Tasks:** 2
- **Files modified:** 4 (+357; вся инфраструктура раннера — вне git, в $HOME)

## Accomplishments

- Workflow `.github/workflows/e2e-matrix.yml`: единственный триггер `workflow_dispatch` (раннер печатает на живом столе — fork-события запустить не могут), `permissions: contents: read`, `concurrency: e2e-gnome` без cancel-in-progress (один прогон за раз), шаги checkout → mise install → `mise run e2e-matrix` (та же задача, что локально, D-09) → артефакт `e2e-report.txt` с `if: always()`
- Раннер поднят по книге: actions-runner v2.337.0 в `~/actions-runner`, зарегистрирован ярлыком `gnome` (имя green106), systemd user unit `~/.config/systemd/user/goswitch-ci-runner.service` после `graphical-session.target`; окружение сессии унаследовано от user-менеджера (GDM-импорт; проверено по `/proc/<pid>/environ`: WAYLAND_DISPLAY=wayland-0, XDG_RUNTIME_DIR, DBUS_SESSION_BUS_ADDRESS)
- Заявленная последовательность первого запуска исполнена и подтверждена живьём: проверка регистрации (e2e-matrix отсутствует в реестре) → диспатч до регистрации → фактическое поведение зафиксировано (HTTP 404 «workflow e2e-matrix.yml not found on the default branch») → bootstrap chore-PR №4 (строго один файл workflow, squash `d815e72` на main) → диспатч `--ref gsd/phase-02-korrektsiya-slova-en-ru` → run 34913812782 conclusion success, 16/16 PASS, артефакт скачан
- `docs/ci-runner.md` — воспроизводимая книга раннера (требования из STACK, установка, unit, окружение, включение/проверка, bootstrap-регистрация, обновление, снятие, диагностика живого стола — лечения из deferred-items 02-06: осиротевшее имя IBus, зависание моста a11y); `SECURITY.md` — реестр угроз фазы 2 (02-01..02-07) с блоком раннера и принятыми рисками

## Task Commits

1. **Task 1: workflow + книга раннера + SECURITY.md + README CI-подраздел** — `419799d` (ci)
2. **Task 2: живой раннер + первый зелёный dispatch (repo-сторона: POST-фикс книги)** — `23a439e` (docs)

Вне ветки плана: bootstrap squash-коммит `d815e72` на `main` (PR №4, chore-ветка удалена).

**Plan metadata:** см. финальный docs-коммит ниже.

## Files Created/Modified

- `.github/workflows/e2e-matrix.yml` — dispatch-only workflow матрицы на self-hosted gnome-раннере
- `docs/ci-runner.md` — инструкция раннера с bootstrap-регистрацией воркфлоуса
- `SECURITY.md` — реестр угроз фазы 2 + блок CI-раннера (создан; реестр фазы 1 вёл `.planning/phases/01-.../01-SECURITY.md`, в дереве не дублировался)
- `test/e2e/README.md` — подраздел «CI-прогон» (одна mise-задача локально и в CI)
- Вне git: `~/actions-runner/` (v2.337.0, credentials вне репозитория), `~/.config/systemd/user/goswitch-ci-runner.service`

## Decisions Made

- Токены минта (`registration-token`, `remove-token`) требуют `gh api --method POST` — без него gh делает GET и получает 404; обе команды в книге исправлены по живой находке
- Импорт окружения: хватило GDM-импорта в user-менеджер при входе — EnvironmentFile не понадобился (юнит наследует переменные сессии напрямую)
- Формулировка плана «origin/master» ↔ фактическая default branch `main` — механизм идентичен, использован `origin/main`
- Версия раннера зафиксирована: 2.337.0 (последний релиз на 2026-09-15)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Команды минта токена в книге не работали без POST**
- **Found during:** Task 2 (живая установка раннера)
- **Issue:** `gh api repos/.../registration-token --jq .token` (формулировка плана, скопированная в книгу) делает GET — endpoint отвечает HTTP 404; установка была заблокирована
- **Fix:** `gh api --method POST ...` для registration-token и remove-token; заметка о находке добавлена в docs/ci-runner.md
- **Files modified:** docs/ci-runner.md
- **Verification:** registration succeeded, раннер online, зелёный dispatch-прогон
- **Committed in:** 23a439e

**2. [Rule 3 - Blocking] Default branch — main, не master**
- **Found during:** Task 2 (bootstrap-последовательность)
- **Issue:** план буквально требует «свежую ветку от origin/master»; в репозитории default branch — `main` (origin/HEAD → origin/main)
- **Fix:** bootstrap исполнен от `origin/main` — механизм без изменений
- **Files modified:** none (терминология плана)
- **Verification:** chore-ветка создана, PR №4 смержен, регистрация прошла
- **Committed in:** n/a (off-repo действие по плану)

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking)
**Impact on plan:** Обе — живые факты против формулировок плана; заявленная последовательность первого запуска исполнена в точности.

## Issues Encountered

None - обе задачи прошли по заявленной последовательности; единственные живые сюрпризы (404 на GET-минте, main-vs-master) закрыты на месте и задокументированы как отклонения.

## User Setup Required

None - no external service configuration required. (Раннер уже поставлен исполнителем в рамках Task 2; повторная установка — по docs/ci-runner.md.)

## Next Phase Readiness

- D-19 закрыт: self-hosted GNOME-раннер поднят в Фазе 2, матрица гоняется CI-контуром; phase gate фазы достигнут — `mise run ci` зелёный, e2e-matrix зелёная локально (02-06) И на раннере (16/16, run 34913812782)
- Фаза 2 завершена (7/7 планов) — готова к `/gsd:verify-work`
- Регрессионный контур Фазы 3+: `gh workflow run e2e-matrix.yml --ref <ветка>` + артефакт e2e-report на каждом прогоне

---
*Phase: 02-korrektsiya-slova-en-ru*
*Completed: 2026-09-15*

## Self-Check: PASSED

Все key-files существуют на диске; оба коммита задачи (419799d, 23a439e) в git log; bootstrap-коммит d815e72 на origin/main; все verify-гейты плана (WORKFLOW-OK, DOCS-OK, RUNNER-ONLINE, E2E-CI-GREEN, mise run ci ×2) исполнены с PASS.
