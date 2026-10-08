---
phase: 261007-0yg-move-engine-and-layouts-under-internal-o
verified: 2026-10-07T02:20:00Z
status: passed
score: 5/5 must-haves verified
covered_files: [".planning/quick/261007-0yg-move-engine-and-layouts-under-internal-o/261007-0yg-PLAN.md", ".planning/quick/261007-0yg-move-engine-and-layouts-under-internal-o/261007-0yg-SUMMARY.md", ".golangci.yml", "cmd/goswitchd/main.go", "cmd/goswitchd/main_test.go", "docs/LICENSE-data.md", "docs/SPEC.md", "internal/correct/buffer.go", "internal/correct/convert.go", "internal/correct/runs.go", "internal/ctlsvc/ctlsvc_test.go", "internal/detect/detect_test.go", "internal/engine/address.go", "internal/engine/address_test.go", "internal/engine/conn.go", "internal/engine/conn_switcher_test.go", "internal/engine/conn_sync_test.go", "internal/engine/emitter_wire_test.go", "internal/engine/engine.go", "internal/engine/engine_test.go", "internal/engine/factory.go", "internal/engine/keys.go", "internal/engine/types.go", "internal/engine/wire_test.go", "internal/install/install.go", "internal/install/install_test.go", "internal/layouts/dict_en.go", "internal/layouts/dict_ru.go", "internal/layouts/dict_test.go", "internal/layouts/dictgen/main.go", "internal/layouts/dictgen/main_test.go", "internal/layouts/dictgen/testdata/en_fixture.dic", "internal/layouts/dictgen/testdata/ru_bad_count.dic", "internal/layouts/dictgen/testdata/ru_fixture.dic", "internal/layouts/generator/main.go", "internal/layouts/tables.go", "internal/layouts/tables_test.go", "internal/layouts/trigrams.go", "internal/session/actor.go", "internal/session/actor_test.go", "internal/sound/sound_test.go", "mise.toml", "test/e2e/case_macr.go", "test/e2e/case_switch.go", "test/e2e/matrix.go", "test/e2e/preflight.go"]
covered_digest: "v1:sha256:8a415a518394a5eeaab79bc8d574a45367e6729cda84716f3eb20e411afa56af"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 4/5
  gaps_closed:
    - "Behavior byte-preserved / full battery green — the pre-existing -race flake (TestRun_OnConnHookCalledOnce in internal/ctlsvc; same read-once class in TestPlayer_ConnectFailedWarnOnceThenReopen in internal/sound) fixed in 08b598a with deadline-polled waits; 9 consecutive green full-suite -race runs post-fix plus 15x/15x targeted stress"
  gaps_remaining: []
  regressions: []
---

# Quick Task 261007-0yg: Move engine/ and layouts/ under internal/ — Verification Report

**Task Goal:** Move `engine/` and `layouts/` under `internal/` (owner decision 2026-10-07, revising SPEC §8): spec-delta committed BEFORE the move (D-55), mechanical move with zero logic edits and zero stale tool-consumed path references, generator parity byte-identical, gates green.
**Verified:** 2026-10-07T02:20:00Z
**Status:** passed
**Re-verification:** Yes — after gap closure (fix commit 08b598a)

## Gap Closure (re-verification)

**Carried gap (initial verification, status gaps_found):** the project's `-race` battery was not reliably green — `TestRun_OnConnHookCalledOnce` (internal/ctlsvc) failed in ~25% of full-suite parallel runs (read-once of the OnConn call counter with no happens-before edge to the hook's invocation), plus the same read-once class observed in `TestPlayer_ConnectFailedWarnOnceThenReopen` (internal/sound). Pre-existing (reproduced on pre-move parent a0d72c9), not move-induced.

**Fix verified in the tree (commit 08b598a, test-only, 2 files, +41/−13):**
- `internal/ctlsvc/ctlsvc_test.go` (+27/−6): new `waitHookCalls` helper — a genuine deadline-polled wait-for-condition (5 s deadline, 2 ms poll tick; the `calls() == want` condition terminates the wait, the sleep only spaces polls — not a fixed-sleep hack), replacing the one-shot counter read in `TestRun_OnConnHookCalledOnce`. Exact-count semantics preserved (over-calls spin to deadline then fail with the real count).
- `internal/sound/sound_test.go` (+14/−7): both read-once log-count assertions in `TestPlayer_ConnectFailedWarnOnceThenReopen` converted to the file's pre-existing `poll(...)` helper.
- Test-only confirmed: both files are `_test.go`; zero production code changed; the move's mechanical truths untouched.

**Independent stability evidence (this verification, not the executor's runs):**
- 8 consecutive green full-suite `go test -race -count=1 ./...` runs + 1 green `mise run ci` = 9 consecutive full-battery greens (under the pre-fix ~25% per-run flake rate, P(9 greens) ≈ 7% — the flake is gone).
- Targeted stress: `TestRun_OnConnHookCalledOnce` ×15 -race ok; `TestPlayer_ConnectFailedWarnOnceThenReopen` ×15 -race ok; both packages ×5 -race ok.
- `mise run tidy-diff` rc=0; generator parity re-confirmed (`go generate ./internal/layouts` + `git diff --exit-code` clean); tracked tree clean; stale-import grep still 0.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SPEC §8 dated amendment, old placement preserved verbatim as audit trail + owner rationale; committed BEFORE any code (D-55), zero .go files in that commit | ✓ VERIFIED | Commit 649c00f: subject exactly `docs(261007-0yg): spec-delta — SPEC §8: engine/ и layouts/ переносятся под internal/ (D-55)`, changed-file list exactly `docs/SPEC.md`; history order a0d72c9 → 649c00f → 5427d7a → 08b598a. SPEC.md:358-368: dated HTML comment (2026-10-07, решение владельца, compiler-enforced encapsulation rationale) with the original `engine/`/`layouts/` bullets verbatim INSIDE the comment (indented, not live), live bullets `- internal/engine/` / `- internal/layouts/` below; other §8 bullets untouched |
| 2 | Packages at internal/engine and internal/layouts; EVERY import reads …/internal/engine\|layouts; stale-path grep == 0; package names unchanged | ✓ VERIFIED | `test ! -e engine && test ! -e layouts` OK; `internal/engine` (12 .go files) and `internal/layouts` (root + generator/ + dictgen/) present; stale quoted-import grep across tracked .go files = **0** (re-checked post-fix); 20 files carry the new import paths (pre-move census: 20 files / 21 import lines); package clauses `package engine` / `package layouts` unchanged; dictgen/main_test.go:209 `"package layouts"` marker stays valid |
| 3 | Behavior byte-preserved — mechanical restructure, NO logic edits beyond import lines and path references; full battery green (mise run ci incl. test -race) and tidy-diff green at every task boundary | ✓ VERIFIED (after gap closure) | Mechanical half proven in initial verification: all 24 renames R098–R100; every diff hunk in all 19 changed .go files an import-line pair; config/doc deltas exactly the 6 specified lines; numstat +31/−31. Battery half now stable: 9 consecutive green full-suite -race runs (8× `go test -race -count=1 ./...` + `mise run ci`) post-fix 08b598a, `mise run tidy-diff` rc=0, targeted stress 15×/15×/5× green — see Gap Closure |
| 4 | Regenerating layout tables from the new location reproduces committed files byte-identically; relative go:generate directives move unchanged | ✓ VERIFIED | `go generate ./internal/layouts` then `git diff --exit-code -- internal/layouts` clean (re-run post-fix), tracked tree clean; relative directives `./generator` (tables.go:3) and `./dictgen` (dict_ru/dict_en/trigrams.go:3) moved unchanged; mise.toml:216 dictgen-regen runs `go generate ./internal/layouts` |
| 5 | All non-Go path references point at the new locations (.golangci.yml:102 comment + :139/:144 gosec regexes; mise.toml dictgen-regen; docs/LICENSE-data.md pointers + generate line) | ✓ VERIFIED | .golangci.yml:102 `internal/layouts/tables.go`, :139 `internal/engine/conn\.go`, :144 `internal/engine/address\.go`; mise.toml:216 `go generate ./internal/layouts`; docs/LICENSE-data.md:5,7,8,9,12,93 all `internal/layouts/…` forms. Negated stale sweeps clean (rc=1, zero matches): `(^|[^/])layouts/\|generate \./layouts` on mise.toml and docs/LICENSE-data.md; `(^|[^/])(engine\|layouts)/` on .golangci.yml |

**Score:** 5/5 truths verified (0 present-but-behavior-unverified)

### Intentional Residue (verified present and untouched, per plan)

| Residue | Location | Status |
|---------|----------|--------|
| Generator identity marker `goswitch/layouts/generator` | internal/layouts/tables.go:1 (template: generator/main.go:418) | Present, byte-identical — correct (editing would break parity) |
| Generator identity marker `goswitch/layouts/dictgen` | internal/layouts/dict_ru.go:1, dict_en.go:1, trigrams.go:1 (template: dictgen/main.go:217) | Present, byte-identical — correct |
| Dev-tool doc comments naming old root generate path | internal/layouts/generator/main.go:4, internal/layouts/dictgen/main.go:5 | Present, untouched — correct (import-lines-only rule) |

These are outside the quoted-import grep gate by design and outside the non-Go sweeps (.go files). Not gaps.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| docs/SPEC.md §8 | Dated spec-delta, docs-only commit | ✓ VERIFIED | 649c00f, exactly one file changed |
| internal/engine/ | 12 .go files moved from engine/ | ✓ VERIFIED | All 12 renames R098–R100, history preserved |
| internal/layouts/ | root + generator/ + dictgen/ moved | ✓ VERIFIED | All 12 renames R098–R100 (incl. 3 testdata .dic files) |
| 15 external importer files rewritten | + 5 own-package tests = 20 files | ✓ VERIFIED | 20 files carry new imports; grep gate 0 stale |
| .golangci.yml, mise.toml, docs/LICENSE-data.md | Path references updated | ✓ VERIFIED | Exactly the 6 specified lines; sweeps clean |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| Import rewrite | Old-path grep gate | `…goswitch/\(engine\|layouts\)` count 20 files → 0 | ✓ WIRED | grep = 0 tracked .go files (re-checked post-fix); 20 files on new paths |
| .golangci.yml gosec exclusions | Moved file locations | `internal/engine/conn\.go` / `internal/engine/address\.go` regexes | ✓ WIRED | Lines 139/144 re-pointed; golangci-lint 0 issues in battery (exclusions still match) |
| mise.toml dictgen-regen | go generate ./internal/layouts | Relative directives survive | ✓ WIRED | Parity check reproduced golden files byte-identically (re-run post-fix) |
| SPEC §8 amendment | Tree reality | Skeleton bullets vs directories | ✓ WIRED | Bullets name internal/engine, internal/layouts — both exist; old dirs gone |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Stale import paths | `grep -rn '"github.com/Djarvur/goswitch/\(engine\|layouts\)' --include='*.go'` | 0 matches | ✓ PASS |
| Directory move | `test ! -e engine && test ! -e layouts && test -d internal/engine && test -d internal/layouts` | all OK | ✓ PASS |
| Generator parity | `go generate ./internal/layouts && git diff --exit-code -- internal/layouts` | clean, tree clean | ✓ PASS |
| Full battery | `mise run ci` | rc=0, lint 0 issues, 21 packages ok | ✓ PASS |
| Dependency hygiene | `mise run tidy-diff` | rc=0, go.mod/go.sum untouched | ✓ PASS |
| SKILL.md golden gate | `go test -race -count=1 ./cmd/skillgen -run TestSkillgen_CommittedFileInSync` | ok (also green in battery) | ✓ PASS |
| Battery stability (post-fix) | `go test -race -count=1 ./...` ×8 + `mise run ci` ×1 | 9/9 green, zero FAIL, zero DATA RACE (pre-fix flake rate was ~25%/run) | ✓ PASS |
| Fixed tests under stress | `go test -race -count=15 -run TestRun_OnConnHookCalledOnce ./internal/ctlsvc`; same ×15 for `TestPlayer_ConnectFailedWarnOnceThenReopen` ./internal/sound; both packages ×5 | all ok | ✓ PASS |
| Fix pattern audit | `git show 08b598a` diff review | deadline-polled wait-for-condition in both files (ctlsvc: new `waitHookCalls` 5 s/2 ms tick; sound: pre-existing `poll` helper); test-only `_test.go` files; no sleeps-as-sync, no production code | ✓ PASS |

### Probe Execution

Step 7c: SKIPPED — no probes exist (`find scripts -path '*/tests/probe-*.sh'` → 0 files; none declared in PLAN/SUMMARY).

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| SPEC-8 | 261007-0yg-PLAN | §8 skeleton lists internal/engine, internal/layouts | ✓ SATISFIED | SPEC.md:367-368 live bullets; audit trail in comment |
| OWNER-DECISION-2026-10-07 | 261007-0yg-PLAN | engine/ and layouts/ under internal/ | ✓ SATISFIED | Tree reality + commit 5427d7a |
| D-55 | 261007-0yg-PLAN | Spec before code | ✓ SATISFIED | 649c00f (docs-only) precedes 5427d7a (move) |

No orphaned requirements (quick task; all three plan-declared IDs covered).

### Anti-Patterns Found

None. No TBD/FIXME/XXX/PLACEHOLDER markers in any moved, modified, or newly fixed file; no stubs; no debt introduced. The fix commit itself is clean (test-only, polled waits, commented rationale). Untracked planning artifacts (.gsd/, .planning/*) are house-flow files, not phase output.

### gofmt Import-Sort Deviation (SUMMARY flag — verified acceptable)

The executor's declared deviation is import-lines-only: every hunk in the 7 gofmt-re-sorted files (cmd/goswitchd/main.go, cmd/goswitchd/main_test.go, internal/ctlsvc/ctlsvc_test.go, internal/install/install.go, internal/session/actor.go, internal/session/actor_test.go, test/e2e/case_macr.go) touches import lines exclusively. Commit 5427d7a numstat +31/−31 confirms no non-import lines moved anywhere; the fix commit 08b598a (+41/−13) is test-logic only.

### Byte-Preservation of Moved Files (rename similarity)

All 24 renames in 5427d7a detected at R098–R100. R100: all non-test sources, dictgen, generator, testdata. R098/R099: exactly the 5 own-package test files whose single-line import delta accounts for the ~1–2% difference. History preserved.

### Human Verification Required

None. The task is fully automatable; all automatable checks were executed (per task instructions, the owner's acceptance of this mechanical move is the merge itself).

### Gaps Summary

None remaining. The single gap from the initial verification — the pre-existing `-race` battery flake (read-once counter reads racing the awaited event in internal/ctlsvc and internal/sound) — is closed by commit 08b598a with the correct deadline-polled pattern, test-only, verified in the diff. Independent post-fix evidence: 9 consecutive green full-battery `-race` runs (pre-fix flake rate ~25% per run makes 9 greens ≈ 7% likely by chance), 15×/15× targeted test stress, plus green `mise run ci` / `tidy-diff` / generator-parity re-checks. No regressions: the move's mechanical truths all re-verified against the post-fix tree (dirs, grep gates, parity, residue, SPEC §8, commit order).

---

_Verified: 2026-10-07T02:20:00Z_
_Verifier: Claude (gsd-verifier)_
