---
phase: 01
plan: 05
subsystem: ci-and-dev-experience
tags: [ci, github-actions, dependabot, govulncheck, mise, systemd, readme, agents-md]
requires:
  phase: 01-adr-paket-i-karkas-ibus-dvizhka
  provides: [mise.toml task set from 01-01, green Go module from 01-02..01-04, owner-rewritten CONVENTIONS.md]
provides:
  - pr-sanity workflow (build/vet/golangci-lint/test -race/tidy-diff + separate govulncheck job, tools from mise.toml)
  - weekly dependabot gomod updates gated by full pr-sanity
  - weekly scheduled govulncheck with non-zero exit on vulnerabilities
  - AGENTS.md Conventions section carrying the five owner engineering directives
  - dev systemd user unit dist/systemd/user/goswitchd.service
  - README debug-mode/privacy warning + no-root dev launch instructions
affects: []
actuals:
  tokens: 4200
  tasks: 2
  commits: 2
plan_head_before: 8d2d69c71d6d19f0cd65b80fd0f09c53ad60f2e0
tech-stack:
  added: []
  patterns:
    - "CI installs tools via `mise install` from mise.toml (single version source) instead of setup-go/golangci-lint-action — CI and local `mise run ci` are the same path"
    - "Periodic dependency/security checks as PR-generating mechanisms (dependabot) or red-job scanners (scheduled govulncheck), never report-only"
key-files:
  created:
    - .github/workflows/pr-sanity.yml
    - .github/workflows/security-scheduled.yml
    - .github/dependabot.yml
    - dist/systemd/user/goswitchd.service
  modified:
    - AGENTS.md
    - .planning/codebase/CONVENTIONS.md
    - README.md
key-decisions:
  - "mise install over golangci-lint-action in CI: the action's version input would duplicate the mise.toml pin and drift; mise gives exact parity with local mise run ci (D-09/D-11a)"
  - "dependabot over a report-only scheduled `go list -m -u all` job: dependabot PRs automatically run the full pr-sanity gate, so updates cannot merge un-gated (D-11b)"
  - "CONVENTIONS.md directives reformatted as single-line `- ` bullets (ordinals kept inside the bold prefix) — the generate-claude-md summarizer carries only headings/bullets/tables and silently dropped the numbered-list directives"
patterns-established:
  - D-11 CI triplet (pr-sanity + weekly dependabot + weekly scheduled govulncheck) created from Phase 1 and kept current
  - GSD-managed AGENTS.md sections regenerate from .planning/codebase/ sources; directive content must be transport-safe (bullets, not numbered lists)
requirements-completed: [TEST-01, INST-04, INTEG-03]
coverage:
  - id: cov-pr-sanity
    description: "pr-sanity.yml: mise install, gates build/vet/lint/test -race/tidy-diff, separate govulncheck job, permissions, concurrency, no compat matrix"
    requirement: TEST-01
    verification:
      - kind: unit
        ref: "grep mise install/golangci-lint/tidy-diff/govulncheck/'contents: read' .github/workflows/pr-sanity.yml → WORKFLOWS-OK"
        status: pass
      - kind: other
        ref: "yaml.v3 parse of pr-sanity.yml → YAML-OK"
        status: pass
      - kind: unit
        ref: "mise run ci && mise run tidy-diff (same gates CI runs, minus network-side govulncheck) → LOCAL-GATES green"
        status: pass
    human_judgment: true
    rationale: "GitHub-runner behavior (mise.run installer availability, mise install of pinned tools) only proves out on the first real push; local gates prove the gate commands themselves pass."
  - id: cov-security-scheduled
    description: "security-scheduled.yml: weekly cron + workflow_dispatch, govulncheck with non-zero exit on vulnerabilities"
    verification:
      - kind: unit
        ref: "grep schedule/govulncheck .github/workflows/security-scheduled.yml → pass"
        status: pass
      - kind: other
        ref: "yaml.v3 parse of security-scheduled.yml → YAML-OK"
        status: pass
    human_judgment: true
    rationale: "First scheduled run on GitHub is the only proof the cron fires and govulncheck exits non-zero on a real vulnerability; govulncheck is network-side by design and not mirrored locally."
  - id: cov-dependabot
    description: "dependabot.yml: gomod weekly with in-file rationale against the report-only alternative"
    verification:
      - kind: unit
        ref: "grep gomod .github/dependabot.yml → pass; rationale comment present"
        status: pass
      - kind: other
        ref: "yaml.v3 parse of dependabot.yml → YAML-OK"
        status: pass
    human_judgment: true
    rationale: "Dependabot activation (first weekly PR, full pr-sanity on it) is GitHub-side and not locally provable."
  - id: cov-agents-md
    description: "AGENTS.md Conventions carries the five owner directives after regeneration from CONVENTIONS.md"
    verification:
      - kind: unit
        ref: "grep 'Строгий TDD'/mise/golangci-lint AGENTS.md → CONVENTIONS-CARRIED"
        status: pass
    human_judgment: false
  - id: cov-systemd-unit
    description: "dev unit: PartOf=graphical-session.target, After=IBus GNOME user unit, Restart=on-failure, WantedBy"
    requirement: INTEG-03
    verification:
      - kind: unit
        ref: "grep all four directives → UNIT-OK"
        status: pass
      - kind: other
        ref: "systemd-analyze verify dist/systemd/user/goswitchd.service → exit 0 (dev binary absent is a runtime condition, installed via README go install)"
        status: pass
    human_judgment: false
  - id: cov-readme-privacy
    description: "README: «Режим отладки и конфиденциальность» (-debug traces keystrokes, passwords warning) + no-root dev launch via user unit"
    requirement: INST-04
    verification:
      - kind: unit
        ref: "grep -c -debug ≥ 1, grep goswitchd.service, grep -i парол → README-OK"
        status: pass
    human_judgment: false
duration: 8 min
completed: 2026-09-11
status: complete
---

# Phase 01 Plan 05: CI and Dev-Experience Package Summary

pr-sanity (build/vet/golangci-lint/test -race/tidy-diff + govulncheck job) with mise-pinned toolchain, weekly dependabot gomod and weekly scheduled govulncheck, owner directives transported into AGENTS.md, dev systemd user unit and privacy-documented -debug key tracing.

## Performance

- Duration: 8 min (estimate band: 2 tasks, low confidence — actual 2 tasks, 8 min)
- Commits: 2 production (measured `git rev-list --count 8d2d69c..HEAD` at SUMMARY time)
- Files touched: 7 (4 created, 3 modified)
- Actual tokens: ~4,200 (16,717 diff chars / 4) vs 28,000 estimated — plan over-estimated; YAML/docs work is compact

## Accomplishments

- All three D-11 mechanisms created from Phase 1: `.github/workflows/pr-sanity.yml` (push/PR gates as mise tasks — build, vet, golangci-lint, test -race, tidy-diff — plus a separate govulncheck job), `.github/dependabot.yml` (weekly gomod, choice against report-only fixed in a comment), `.github/workflows/security-scheduled.yml` (weekly cron govulncheck, red job on vulnerabilities).
- Tools come only from `mise install` reading mise.toml — no setup-go, no golangci-lint-action (rejected: its version input would duplicate the mise.toml pin and drift). CI and local `mise run ci` are literally the same tasks.
- e2e stays local: no `mise run e2e-*` anywhere in pr-sanity (live GNOME stand cannot run headless).
- AGENTS.md Conventions section regenerated from the owner's CONVENTIONS.md and now carries the five engineering directives (strict TDD, green iteration, mise-not-make, strict golangci-lint from Phase 1, current GitHub Actions); the old "golangci-lint deferred" rejection is gone (cancelled by D-10).
- `dist/systemd/user/goswitchd.service`: PartOf/After graphical session chain, Restart=on-failure (covers kill -9), ExecStart=%h/go/bin/goswitchd; installed by hand per README, not by a phase script.
- README gained «Режим отладки и конфиденциальность» (-debug traces every keystroke — logs may contain passwords; stderr to journal in production) and the no-root dev launch instructions (INTEG-03/INST-04).

## Task Commits

| Task | Commit | Subject |
|------|--------|---------|
| 1 — GHA D-11 package + AGENTS.md regeneration | 8959105 | ci(01-05): pr-sanity + dependabot + scheduled govulncheck; AGENTS.md directives |
| 2 — dev systemd unit + README privacy | cb93253 | feat(01-05): dev systemd user unit + README debug/privacy section |

## Files Created/Modified

Created: `.github/workflows/pr-sanity.yml`, `.github/workflows/security-scheduled.yml`, `.github/dependabot.yml`, `dist/systemd/user/goswitchd.service`.
Modified: `AGENTS.md` (Conventions section), `.planning/codebase/CONVENTIONS.md` (directive formatting — see Deviations), `README.md`.

## Decisions Made

- mise install (not setup-go / golangci-lint-action) as the CI tool source — single version pin in mise.toml, exact parity with `mise run ci` (D-09/D-10/D-11a).
- dependabot (not a report-only scheduled `go list -m -u all` job) for dependency updates — every dependabot PR runs the full pr-sanity gate, so nothing merges un-gated (D-11b).
- govulncheck stays network-side (`go run ...@latest` in CI only): tool + vuln DB downloads belong to the ephemeral runner; if @latest ever needs a newer toolchain, the go pin rises in mise.toml together with go.mod — one place.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] generate-claude-md dropped the numbered-list directives; default output path was not AGENTS.md**

- **Found during:** Task 1 (AGENTS.md regeneration step)
- **Issue:** Two transport problems. (a) `generate-claude-md`'s conventions summarizer carries only headings, `- `/`* ` bullets and tables — the owner's five directives are a numbered list with indented continuation lines, so the regenerated Conventions section came out with headings only and zero directives (verify `grep 'Строгий TDD' AGENTS.md` would fail). (b) The tool wrote to `.claude/CLAUDE.md` (config `claude_md_path` default), a file that does not exist in git and is not this project's instruction file.
- **Fix:** (a) Reformatted the directives in the SOURCE `.planning/codebase/CONVENTIONS.md` as single-line `- ` bullets with ordinals preserved inside the bold prefix (`- **1. Строгий TDD.** …`) — wording verbatim, only list formatting changed; added a NOTE in the source explaining the transport constraint. (b) Re-ran with `--output AGENTS.md` and deleted the accidentally created untracked `.claude/CLAUDE.md`. AGENTS.md itself was not hand-edited between markers — content still flows from the source via the tool.
- **Files modified:** `.planning/codebase/CONVENTIONS.md`, `AGENTS.md` (regenerated)
- **Verification:** `grep 'Строгий TDD'/mise/golangci-lint AGENTS.md` → CONVENTIONS-CARRIED; `git diff AGENTS.md` shows only the Conventions section changed
- **Commit:** 8959105

Totals: 1 deviation, 0 unfixed. Impact: none on plan goals — directives now survive every future regeneration.

## Issues Encountered

- `actionlint` and `yq` are not installed on this machine; YAML syntax validated instead via yaml.v3 (cached module) — all three files parse. Workflow-action semantics (action versions, expressions) rest on the go-ultimate skill template plus the flagged human-judgment coverage entries.

## User Setup Required

- None for this plan. (The dev systemd unit is installed by hand per README when the owner wants to dogfood the daemon — that is the documented dev flow, not setup for this plan.)
- Worth watching after first push: the three GitHub-side activations (pr-sanity run, dependabot first PR, first scheduled govulncheck) — flagged as human_judgment in coverage.

## Next Phase Readiness

- Phase 1 wave complete: D-11 CI triplet live, AGENTS.md carries binding directives, dev unit + privacy docs in place.
- Phase 2 (word correction EN↔RU) inherits: green headless gate on every push (including the layout-table work), dependabot keeping godbus current under the gate, and the mise task set as the single entry point.

## Self-Check: PASSED

- Files exist: pr-sanity.yml, security-scheduled.yml, dependabot.yml, goswitchd.service, AGENTS.md, README.md, CONVENTIONS.md — all present (verified via git diff --name-only against plan ledger 8d2d69c).
- Commits exist: 8959105, cb93253 both in `git log`.
- All plan verification commands re-run green at close-out.
