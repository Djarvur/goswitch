---
phase: 04-postavka-i-priemka
plan: 09
subsystem: testing
tags: [ci, github-actions, e2e, d48, nightly, systemd, atspi, gap-closure]

# Dependency graph
requires:
  - phase: 04-06/04-08 (e2e-matrix D-48 gate + focus-first witness)
    provides: e2e-matrix.yml double-run gate with fresh-session preflight; focus_helper.py frames-before-descent traversal (focused_frame_witness shape reused by focused-app)
provides:
  - focused-app mode in test/e2e/focus_helper.py — idle-desktop preflight primitive: name of the app holding the focused frame or "(none)", frames-level only (no descent, no full walk)
  - e2e-matrix.yml nightly gate: schedule cron 01:10 UTC (fresh_session semantics implied), idle-desktop preflight for ALL triggers (fail-fast "desktop busy" on dispatch, soft mode on schedule), D48_SKIP gates on both runs and the artifact upload
  - scripts/d48-nightly-dispatch.sh — in-repo session-dispatcher script (gh guard, D48_REF:-main, matrix-v3 + fresh_session=true, --repo pinned)
  - docs/ci-runner.md «Автономный ночной гейт (D-48 v2)» — the three one-time machine-setup blocks (GDM autologin, root relogin timer, user units), security tradeoff note, D-48 evidence semantics
affects: [owner machine setup for the nightly gate, D-48 formal acceptance (docs/ACCEPTANCE.md item 1), post-merge scheduled runs]

# Actuals (#2632) — pairs with the plan's estimate to calibrate future estimates.
actuals:
  tokens: 9500   # chars/4 over the realized diff (production diff 28466 chars + SUMMARY/state)
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "preflight soft mode: a scheduled trigger downgrades a busy-desktop/stale-session preflight failure to ::warning:: + D48_SKIP=1 + exit 0, and downstream run/upload steps gate on env.D48_SKIP != '1' — the nightly stays neutral while the dispatch keeps the hard gate"
    - "trigger-aware shell steps: the event name reaches the step via env GITHUB_EVENT_NAME (WR-04 canon — no ${{ }} interpolation in shell), so one step serves dispatch and schedule arms"
    - "count-gated workflow literals: verify greps pin 'the matrix needs an idle desktop' to exactly the two executable arms and env.D48_SKIP != '1' to exactly three steps, keeping prose comments free of load-bearing literals"

key-files:
  created:
    - scripts/d48-nightly-dispatch.sh
  modified:
    - test/e2e/focus_helper.py
    - .github/workflows/e2e-matrix.yml
    - docs/ci-runner.md

key-decisions:
  - "focused-app answers at the frames level only — first focused frame names its app ('(unnamed)' when empty), '(none)' when none; shell/mutter surfaces are deliberately NOT classified, interpretation is the caller's (preflight's) job (plan's stated simplification)"
  - "idle-desktop preflight sits right after checkout, before mise install — python3-gi is machine-native, so no tooling is installed just to learn the desk is busy; it runs for ALL triggers (fail-fast busy verdict instead of 31 opaque witness failures)"
  - "fresh-session preflight arms on schedule too (fresh_session semantics implied); its three failure points fold into one local function — schedule gets ::warning:: + D48_SKIP + exit 0, dispatch keeps the byte-identical error texts"
  - "run failures after passed preflights stay red for every trigger — soft mode covers only preflights, a nightly red remains a regression signal"
  - "dispatcher script guards gh presence + auth with named stderr messages (unit output goes to journalctl) and pins --repo (unit starts with arbitrary cwd, gh does not resolve from $HOME)"

patterns-established:
  - "Nightly-gate shape: root timer restarts gdm -> autologin -> graphical-session.target raises user units -> oneshot dispatcher (sleep 120 guard) runs the in-repo script — the human is out of the acceptance loop, the machine-checked freshness stays"
  - "Soft/hard dual-gate preflight pattern reusable for any scheduled live-desktop workflow"

requirements-completed: [INST-01]

coverage:
  - id: D1
    description: "focused-app mode in focus_helper.py: app name of the focused frame or '(none)', frames-before-descent, docstring contract; live-proven"
    requirement: INST-01
    verification:
      - kind: other
        ref: "/usr/bin/python3 -m py_compile test/e2e/focus_helper.py"
        status: pass
      - kind: e2e
        ref: "live smoke: /usr/bin/python3 test/e2e/focus_helper.py focused-app twice — 'gnome-shell' both calls, stable, exactly one line (desktop currently held by a focused shell frame; '(none)' is the no-frame arm)"
        status: pass
      - kind: integration
        ref: "mise run ci (build, vet, golangci-lint strict, go test -race ./...)"
        status: pass
    human_judgment: true
    rationale: "The mode's python behavior is live-provable only (gi/Atspi needs a session — 04-04/04-08 house precedent, plan documents headless RED as impossible); the two-call stability + single-line contract was exercised live, but the '(none)' arm on a genuinely idle desk and the app-name arm on a non-shell window are desk-state-dependent observations the verifier/owner sees in nightly use"
  - id: D2
    description: "e2e-matrix.yml nightly gate: schedule cron 01:10 UTC with default-branch semantics comment, idle-desktop preflight for all triggers (busy: hard error on dispatch / soft warning + D48_SKIP on schedule), fresh-session preflight armed on schedule with folded soft/hard failure function, D48_SKIP gates on run #1/#2/upload"
    requirement: INST-01
    verification:
      - kind: other
        ref: "grep chain NIGHTLY-WIRED: cron literal, focused-app call, github.event_name == 'schedule', D48_SKIP=1, 'the matrix needs an idle desktop' exactly 2 (error+warning arms), env.D48_SKIP != '1' exactly 3 (run#1/run#2/upload), mise run e2e-matrix-v3 exactly 2 (double run intact)"
        status: pass
      - kind: integration
        ref: "mise run ci green at the task commit"
        status: pass
    human_judgment: true
    rationale: "Schedule-vs-dispatch runtime behavior is not headless-checkable (plan states this): the first real scheduled run after merge — and its neutral soft-mode outcome before the owner applies the machine setup — is observed in GitHub Actions, not locally"
  - id: D3
    description: "scripts/d48-nightly-dispatch.sh: exec-bit, bash -n clean, gh guard, D48_REF:-main, dispatches matrix-v3 + fresh_session=true via --repo Djarvur/goswitch"
    requirement: INST-01
    verification:
      - kind: other
        ref: "bash -n scripts/d48-nightly-dispatch.sh && test -x + grep gates (gh auth status, --repo Djarvur/goswitch, fresh_session=true, D48_REF:-main)"
        status: pass
    human_judgment: false
  - id: D4
    description: "docs/ci-runner.md «Автономный ночной гейт (D-48 v2)»: chain intro, autologin security tradeoff note, three copy-paste setup blocks (GDM custom.yaml, d48-nightly-relogin root timer, user units runner + dispatcher), post-merge cron redundancy, D-48 evidence semantics; stale trigger claims updated"
    requirement: INST-01
    verification:
      - kind: other
        ref: "grep gates D48-V2-WIRED: section title, AutomaticLoginEnable, d48-nightly-relogin, goswitch-d48-dispatch, D48_REF present in docs/ci-runner.md"
        status: pass
    human_judgment: true
    rationale: "The machine-side steps (custom.yaml, root timer, user units, gdm restart, live gh workflow run) are DOCUMENTED ONLY by plan boundary — the owner applies them from the doc and thereby proves the copy-paste blocks; the plan explicitly forbids executing them"

# Metrics
duration: 9 min
completed: 2026-09-17
status: complete
---

# Phase 04 Plan 09: G-4-2 gap closure — human-free nightly D-48 gate Summary

**Ночной автономный гейт D-48: idle-desktop preflight с мягким режимом для schedule в e2e-matrix, примитив focused-app, ин-репо диспатч-скрипт и документация машинной настройки — человек выведен из цикла приёмки, свежесть сессии остаётся машинным свидетелем**

## Performance

- **Duration:** 9 min
- **Started:** 2026-09-17T22:09:57Z
- **Completed:** 2026-09-17T22:18:45Z
- **Tasks:** 3
- **Files modified:** 4 (1 created, 3 modified)

## Accomplishments
- G-4-2 fix direction delivered in scope (all five points): schedule cron 01:10 UTC with implied fresh_session semantics; idle-desktop preflight for ALL triggers with fail-fast «desktop busy: <app> holds focus — the matrix needs an idle desktop» on dispatch and neutral soft mode on schedule; focused-app primitive live-proven; dispatcher script with gh guard; docs «Автономный ночной гейт (D-48 v2)» with three copy-paste machine-setup blocks
- The nightly D-48 gate no longer requires a human: after the owner applies the documented machine setup (autologin, root timer 03:50, user units), the fresh session dispatches itself with fresh_session=true; until then scheduled runs end neutrally (warning + exit 0, runs and artifact skipped) — no red nightly noise
- Real run failures after passed preflights stay red for every trigger; double run v3, matrix selection, concurrency e2e-gnome and permissions untouched (count-gated in verify)
- Green iteration (directive 2): mise run ci green at every task commit; no product code touched

## Task Commits

Each task was committed atomically:

1. **Task 1: focused-app idle-preflight primitive** - `cafd4ba` (feat)
2. **Task 2: nightly schedule + idle-desktop preflight with soft mode** - `d759c9f` (feat)
3. **Task 3: dispatcher script + D-48 v2 docs** - `99fd0b1` (feat)

**Plan metadata:** docs commit carrying this SUMMARY, STATE.md, ROADMAP.md and REQUIREMENTS.md (immediately after this file).

## Files Created/Modified
- `test/e2e/focus_helper.py` - cmd_focused_app(): frames-level busy/idle verdict, "(none)" arm; docstring table + usage extended; other modes untouched
- `.github/workflows/e2e-matrix.yml` - schedule trigger, idle-desktop preflight (all triggers), soft-mode fresh-session preflight, D48_SKIP gates on runs/upload; header comment updated
- `scripts/d48-nightly-dispatch.sh` - session-dispatcher script (exec-bit): gh guard, D48_REF:-main, matrix-v3 + fresh_session=true, --repo pinned
- `docs/ci-runner.md` - «Автономный ночной гейт (D-48 v2)» section; intro trigger line + «Безопасность» bullet updated

## Decisions Made
See key-decisions in frontmatter. The load-bearing interpretations: soft mode covers ONLY preflight failures (busy desk / stale session on schedule) — the runs themselves stay red for every trigger; and the fresh-session preflight's three failure messages keep their dispatch semantics byte-identical while routing through the soft/hard function.

## Live proof (Task 1)

`/usr/bin/python3 test/e2e/focus_helper.py focused-app` on the current desktop, two consecutive calls: **`gnome-shell`** / **`gnome-shell`** — stable, exactly one line, exit 0. Observation: the desk is right now held by a focused gnome-shell frame (shell stage), so the busy arm answered; the plan's fails_when explicitly allows either the app name or "(none)" as the valid live answer. The "(none)" arm was not live-reproducible in this session (the shell frame stayed focused throughout) — it is the same frames-level scan with an empty result, and the nightly preflight observes it in ordinary use.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Stale doc intro contradicted the new trigger surface**
- **Found during:** Task 3
- **Issue:** docs/ci-runner.md intro line said the workflow «запускается ТОЛЬКО вручную (workflow_dispatch)» — the same stale fact as the «Триггер — workflow_dispatch только» bullet the plan orders to update; left as-is it would directly contradict the new D-48 v2 section below it
- **Fix:** intro line now reads «запускается вручную (workflow_dispatch) и — после мержа фазы в main — по ночному schedule (план 04-09)» with a pointer to the new section
- **Files modified:** docs/ci-runner.md
- **Verification:** section and intro agree; D48-V2-WIRED gate green
- **Committed in:** 99fd0b1 (Task 3 commit)

---

**Total deviations:** 1 auto-fixed (1 bug-class doc consistency fix directly caused by the task's own change).
**Impact on plan:** none — one-line doc consistency fix in the same file the task already updates; no scope creep.

## Issues Encountered
None.

## User Setup Required

The plan's own deliverable is exactly this setup documentation — the owner applies it once per machine from [docs/ci-runner.md](../../../docs/ci-runner.md), section «Автономный ночной гейт (D-48 v2)»:
- Block 1: GDM autologin (`/etc/gdm3/custom.yaml`, root, by hand)
- Block 2: root relogin timer `d48-nightly-relogin.timer` at 03:50 machine-local
- Block 3: user units — `systemctl --user enable goswitch-ci-runner` and the new `goswitch-d48-dispatch.service` (ExecStart path must point at the actual working copy; optional `D48_REF` env file for pre-merge dispatches)
- Security tradeoff (autologin, T-04-09-01 accepted) is stated in the doc; nothing in the repository executes sudo

## Next Phase Readiness
- G-4-2 closed in the plan's scope; the formal D-48 evidence becomes the unattended green double-run on an autologin-fresh session — the owner applies the three documented blocks, then the nightly (or a manual `gh workflow run` per docs/ci-runner.md) delivers the proof
- The workflow schedule arms itself automatically after this phase merges to main (default-branch definition); pre-merge nightly dispatches ride D48_REF through the session dispatcher
- No product code touched; matrix v3 double run intact (verify count-gate)

---
*Phase: 04-postavka-i-priemka*
*Completed: 2026-09-17*

## Self-Check: PASSED

- Files: focus_helper.py, e2e-matrix.yml, scripts/d48-nightly-dispatch.sh, docs/ci-runner.md, this SUMMARY — all exist on disk
- Commits: cafd4ba, d759c9f, 99fd0b1 — all present in git log
- Plan-level verification re-run at self-check: NIGHTLY-WIRED + D48-V2-WIRED grep chains green; py_compile green; bash -n green; live focused-app smoke stable ('gnome-shell' twice); mise run ci green (build, vet, golangci-lint strict 0 issues, go test -race ./... all ok)
- Plan boundary honored: no machine-side step executed (no custom.yaml, no root timer, no user units, no gdm restart, no live gh workflow run)
- Commits measured from ledger gsd-plan-head-before-04-09 (3cf58e5): 3

