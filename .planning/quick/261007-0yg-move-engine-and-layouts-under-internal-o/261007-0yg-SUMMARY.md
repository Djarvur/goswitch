---
phase: 261007-0yg-move-engine-and-layouts-under-internal-o
plan: 01
subsystem: infra
tags: [internal-packages, encapsulation, git-mv, import-rewrite, spec-delta, D-55]

# Dependency graph
requires:
  - phase: phase-8 (tree a0d72c9)
    provides: engine/ and layouts/ packages at repo root; SPEC §8 skeleton; mise task surface (D-09)
provides:
  - engine and layouts packages live at internal/engine and internal/layouts — compiler-enforced encapsulation (external modules cannot import the IBus wire adapter)
  - SPEC §8 dated amendment rev 2026-10-07 with the old root placement preserved as audit trail (D-55: spec before code)
  - go-generate parity proof from the new location (regeneration byte-identical)
affects: [any future consumer of engine/layouts, release engineering, ADR-006 audit trail readers]

# Actuals (#2632)
actuals:
  tokens: 4471    # 17883 diff chars / 4 over a0d72c9..HEAD
  tasks: 3
  commits: 2

# Tech tracking
tech-stack:
  added: []       # zero new dependencies — go.mod/go.sum untouched (tidy-diff green at every boundary)
  patterns: [git-mv package internalization, quoted-import-path rewrite, spec-delta audit trail inside HTML comment]

key-files:
  created:
    - internal/engine/ (12 .go files, moved)
    - internal/layouts/ (root + generator/ + dictgen/, moved)
  modified:
    - docs/SPEC.md (§8 dated amendment)
    - cmd/goswitchd/main.go, cmd/goswitchd/main_test.go
    - internal/install/install.go, internal/install/install_test.go
    - internal/correct/runs.go, internal/correct/convert.go, internal/correct/buffer.go
    - internal/detect/detect_test.go
    - internal/session/actor.go, internal/session/actor_test.go
    - internal/ctlsvc/ctlsvc_test.go
    - test/e2e/case_switch.go, test/e2e/case_macr.go, test/e2e/matrix.go, test/e2e/preflight.go
    - internal/engine/{address,engine,wire}_test.go, internal/layouts/{dict,tables}_test.go (import lines)
    - .golangci.yml (line 102 comment + gosec G404/G703 path exclusions)
    - mise.toml (dictgen-regen task)
    - docs/LICENSE-data.md (dictgen/dict/trigrams pointers + generate line)

key-decisions:
  - "Owner decision 2026-10-07 realized verbatim: engine/ and layouts/ move under internal/ — compiler-enforced encapsulation; single module, no functional change (supersedes the 'перенос был бы косметикой' stance)."
  - "D-55 discipline: SPEC §8 spec-delta committed FIRST as a docs-only commit (649c00f), mechanical move second (5427d7a)."
  - "gofmt re-sorting of the rewritten import blocks allowed as part of the import-lines-only edit: the internal/engine import path sorts after the other github.com/Djarvur/goswitch/internal/* imports; net diff stays exactly the 8 rewritten import lines."
  - "Acknowledged residue left byte-identical by design: generator identity markers (goswitch/layouts/generator, goswitch/layouts/dictgen) baked into generated-file line 1 and the two dev-tool doc comments naming the old root generate path — byte-preservation wins over path freshness."

requirements-completed: [SPEC-8, OWNER-DECISION-2026-10-07, D-55]

status: complete
---

# Quick Task 261007-0yg: Move engine/ and layouts/ under internal/ Summary

Behavior-preserving internalization of the IBus wire adapter and layout tables: SPEC §8 amended first (D-55), then engine/ and layouts/ moved under internal/ with all 20 files' import paths rewritten, go-generate parity proven byte-identical from the new location.

## Accomplishments

- **Task 1 — spec-delta (docs-only, D-55):** SPEC §8 carries the dated amendment comment (2026-10-07, решение владельца: compiler-enforced encapsulation) with the original `- \`engine/\`` / `- \`layouts/\`` bullets preserved verbatim INSIDE the comment as audit trail (indented, so no live skeleton bullet), and the revised `- \`internal/engine/\`` / `- \`internal/layouts/\`` bullets as the live skeleton. Commit 649c00f changed exactly docs/SPEC.md; mise ci + tidy-diff green on the docs-only tree (precedent 07-08).
- **Task 2 — the move (one mechanical commit):** `git mv engine internal/engine`, `git mv layouts internal/layouts` (git detected all renames at 98–100% similarity — history preserved); import paths rewritten in all 20 tracked .go files (5 own-package tests + 15 external importers); .golangci.yml line-102 comment and the gosec G404/G703 exclusion regexes re-pointed at internal/engine/conn\.go and internal/engine/address\.go; mise.toml dictgen-regen now runs `go generate ./internal/layouts`; docs/LICENSE-data.md pointers updated (dictgen, dict_ru/dict_en/trigrams, generate line). Commit 5427d7a: 42 files, +31/−31 lines net, zero deletions, zero logic edits.
- **Task 3 — parity + final battery:** `go generate ./internal/layouts` reproduces the committed golden files byte-identically (`git diff --exit-code -- internal/layouts` clean) — the relative go:generate directives (./generator, ./dictgen) survived the move intact. Final `mise run ci` (build + vet + strict golangci-lint 0 issues + test -race -count=1, 21 packages ok) and `mise run tidy-diff` green; tracked working tree clean.

## Verification

- Stale full-path imports across tracked .go files: **0** (was 20 files / 21 import lines) — `! grep -rq '"github.com/Djarvur/goswitch/\(engine\|layouts\)' --include='*.go' .`
- Negated stale-path sweeps clean: mise.toml, docs/LICENSE-data.md (`(^|[^/])layouts/|generate \./layouts` → no match), .golangci.yml (`(^|[^/])(engine|layouts)/` → no match, includes the line-102 comment).
- `test ! -e engine && test ! -e layouts && test -d internal/engine && test -d internal/layouts` — old directories gone, new locations present.
- Package names unchanged (`engine`, `layouts`); dictgen/main_test.go:209's `"package layouts"` marker stays valid.
- Generator parity: regenerate → `git diff --exit-code -- internal/layouts` clean.
- Commit ordering: 649c00f (docs-only spec-delta, changed-file list exactly docs/SPEC.md) precedes 5427d7a (mechanical move). Change is exactly two commits on top of a0d72c9.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] gofmt re-sorted the rewritten import blocks (7 files)**
- **Found during:** Task 2 (first `mise run ci` after the rewrite: gofmt flagged 7 importers)
- **Issue:** The rewritten import `github.com/Djarvur/goswitch/internal/engine` no longer sat in sorted position within its import group (it previously sorted early as `engine`); gofmt requires sorted imports within a group.
- **Fix:** `gofmt -w` on the 7 flagged files. Verified the net diff of those files is exactly the 8 import path rewrites — repositioning is import-lines-only, no other edits.
- **Files modified:** cmd/goswitchd/main.go, cmd/goswitchd/main_test.go, internal/ctlsvc/ctlsvc_test.go, internal/install/install.go, internal/session/actor.go, internal/session/actor_test.go, test/e2e/case_macr.go
- **Commit:** 5427d7a

### Observations (not deviations)

- One transient test FAIL occurred during Task 2's first post-rewrite `mise run ci` (a `-race` run under parallel load; the failing package's output was not captured before the run ended). It did not reproduce in three consecutive subsequent full runs (`go test -race -count=1 ./...` twice + `mise run ci` once — 21 packages ok each, no DATA RACE). Treated as environmental, not a product regression; no code or test was altered in response.

## Known Stubs

None. No stubs, skipped tests, or unrun verify steps were introduced.

## Threat Flags

None. No new trust boundaries — the public-API surface shrank (internal/ is unimportable from outside the module); no new inputs, processes, network, or storage surfaces.

## Intentional Residue (acknowledged, out of scope per plan)

- Generator identity markers baked into generated-file line 1 via templates (`goswitch/layouts/generator` — internal/layouts/generator/main.go:418; `goswitch/layouts/dictgen` — internal/layouts/dictgen/main.go:217) remain stale on purpose: editing them would break the byte-parity gate.
- The two dev-tool doc comments naming the old root generate path (internal/layouts/generator/main.go:4, internal/layouts/dictgen/main.go:5) remain untouched: editing them would break the import-lines-only rule.

## Self-Check: PASSED

- Files: internal/engine/ and internal/layouts/ exist with all moved sources; old engine/ and layouts/ gone — verified live.
- Commits: 649c00f (spec-delta) and 5427d7a (move) both present in `git log` on gsd/internalize-engine-layouts — verified live.
- Battery: final `mise run ci` rc=0 and `mise run tidy-diff` rc=0 — verified live.
