---
phase: 261006-vqw-mixed-text-per-character-layout-inversio
verified: 2026-10-06T00:00:00Z
status: human_needed
score: 5/5 must-haves verified
covered_files:
  - ".planning/quick/261006-vqw-mixed-text-per-character-layout-inversio/261006-vqw-PLAN.md"
  - ".planning/quick/261006-vqw-mixed-text-per-character-layout-inversio/261006-vqw-SUMMARY.md"
  - "README.md"
  - "docs/SPEC.md"
  - "internal/correct/runs.go"
  - "internal/correct/runs_test.go"
  - "internal/session/actor.go"
  - "internal/session/actor_test.go"
  - "test/e2e/case_word.go"
  - "test/e2e/cases/matrix-v1.yaml"
  - "test/e2e/cases/matrix-v2.yaml"
  - "test/e2e/cases/matrix-v3.yaml"
  - "test/e2e/cases/matrix-v4.yaml"
covered_digest: "v1:sha256:ef231206671dac5e6d4c47114bfe9dc3cd315bac4323ebac735fcdf7c32f6af4"
behavior_unverified: 0
overrides_applied: 0
human_verification:
  - test: "Run the live word/phrase matrix on the GNOME stand: mise run e2e-matrix"
    expected: "The re-pinned word-mixed row passes — after the double tap the field reads \"паиghbdtn\" (every letter inverted), homogeneous rows unchanged"
    why_human: "Drives the owner's live GNOME desktop (IBus + real surfaces); cannot run headless or in this verification process"
  - test: "Run the live matrix v2: mise run e2e-matrix-v2"
    expected: "phrase-mixed reads \"привет ghbdtn\", word-mixed \"паиghbdtn\", select-all-gte/chromium \"привет ghbdtn\"; select-all-zenity (D-30 degradation) and partial/reverse-selection rows unchanged"
    why_human: "Live-desktop gate; the runner types into real GTK/chromium surfaces"
  - test: "Run the live matrix v3: mise run e2e-matrix-v3"
    expected: "Same 4 mixed rows at the new inversion outputs; degradation/partial/homogeneous rows untouched"
    why_human: "Live-desktop gate"
  - test: "Run the live matrix v4: mise run e2e-matrix-v4"
    expected: "Same 4 mixed rows at the new inversion outputs; the header now records the WINDOWS-12 freeze as REJECTED; all carried breadth rows green"
    why_human: "Live-desktop gate"
  - test: "Hand-typed mixed-text check in gedit/GTK4: type a Latin word, flip, type a Cyrillic word (e.g. gfb + привет), hit the correction hotkey"
    expected: "Every letter inverts (паиghbdtn); digits/space/punctuation in the mixed range ride unchanged; no visual jump; correction-done in the log"
    why_human: "Real-time keyboard/IME interaction and visual result — not reproducible programmatically"
---

# Quick Task 261006-vqw: Mixed-Text Per-Character Layout Inversion — Verification Report

**Task Goal:** Mixed text: per-character layout inversion as THE correction semantics for mixed-script text (owner verdict, UAT phase 6 — replaces convert-only-foreign-runs D-16..D-23; WINDOWS-12 freeze rejected). Hard requirements (a)–(f): spec-delta before code, RED-first corpus re-pin, inversion in the run pipeline, e2e matrix re-pins v1..v4, non-regression of homogeneous/refusal paths, UX contracts preserved.

**Verified:** 2026-10-06
**Status:** human_needed (all automatable checks PASS; the 5 remaining items are the plan's declared inherently-manual gates)
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Mixed text inverts EVERY letter per character, no anchor, no last-letter rule ("ghbdtn"+"привет"→"привет"+"ghbdtn"; "gfb"+привет→"паиghbdtn") | ✓ VERIFIED | `invertPerChar` (internal/correct/runs.go:118-140) wired into the mixed branch at runs.go:97-98; `TestConvertRuns_MixedInvertsPerChar` pins both cases; `TestActor_MixedWordInvertsPerChar` / `TestActor_PhraseMixedCorrects` / `TestActor_SelectionMixedConverts` pin the same outputs through the real pipeline. Behaviorally proven: independently re-run at the RED commit d7a8a5e in a throwaway worktree — exactly the 4 re-pinned tests FAIL against the old pipeline; at HEAD all PASS under `-race` (mise run ci green). |
| 2 | Neutral characters pass through (digits, space); mixed-range punctuation rides as typed; stated in SPEC §4.2, not implicit | ✓ VERIFIED | SPEC.md:106-111 states the rule explicitly with the both-tables ',' argument; `invertPerChar` neutral branch (runs.go:127-131); unit pins "digits ride", "space rides", "bracket rides in a mixed range". Table facts confirmed live: layouts/tables.go:25 ENToRU[',']='б' vs :120 RUToEN[',']='?'; '5'→'5' at :34 and :129. |
| 3 | Non-mixed behavior preserved byte-for-byte (wholesale conversion incl. punctuation-with-token; homogeneous/refusal corpora green untouched) | ✓ VERIFIED | Byte-diff vs pre-plan tree 0bec17f: `TestConvertRuns_Refusals` and `TestActor_DoubleTapSelectionCorrects` BYTE-IDENTICAL; `TestConvertRuns_HomogeneousWholesale` differs only by the 3 deliberately moved mixed-case lines (the planned Task 2b move); wholesale pins intact ("ghbdtn,"→"приветб" at runs_test.go:107, "ghbdtn["→"приветх" at :113, "привет["→"ghbdtn[" at :116); convert.go/direction.go/buffer.go/plan.go/verify.go NOT in the 0bec17f..HEAD diff; homogeneous + degradation (select-all-zenity) + partial/reverse matrix rows untouched (0 changed lines). |
| 4 | Correction UX contracts compose unchanged (per-rune case preservation, exact-range geometry, direction auto-detect, D-20 silent refusals) | ✓ VERIFIED | Case: "register per rune" pin "Gfb"+wordRU→"Паиghbdtn"; geometry: actor pins unchanged (-9,9 word / -13,13 phrase / selection deletes nothing) — only expected commit texts changed; direction: direction.go untouched, homogeneous branch still `Convert(text, dirOf(script))` (runs.go:100); refusals: TestConvertRuns_Refusals byte-identical and green — and its é case input "ghé"+wordRU is a MIXED range, so it exercises the NEW `invertPerChar` whole-range refusal (runs.go:132-135) asserting nil output (no partial leak, T-03-01-03). |
| 5 | SPEC §4.2 dated amendment with old bullet verbatim as audit trail + owner-verdict trace, committed BEFORE any code (D-55); README matches | ✓ VERIFIED | Commit 47d8c71 touches exactly docs/SPEC.md + README.md and precedes d7a8a5e/cb0e968/871ef9e (verified via git log order + `git show --name-only`). Old bullet "Смешанный текст: корректируется только часть, набранной не той раскладкой…" preserved verbatim at SPEC.md:92-93 (byte-compared against 0bec17f — identical, including the `[Q:]` marker). Dated HTML comment (SPEC.md:95-100) carries the full trace: owner decision 2026-10-05 todo, UAT verdict 2026-10-03, commit 7dd46e9, D-22/D-23 superseded, WINDOWS-12 REJECTED. README.md:149-151 states the inversion semantics. |

**Score:** 5/5 truths verified (0 present, behavior-unverified — every truth is backed by a passing behavioral test)

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `docs/SPEC.md` §4.2 | Dated spec-delta, audit trail, neutral rule | ✓ VERIFIED | Lines 92-117; old bullet verbatim; verdict trace; all four clauses (a)-(d) of plan Task 1 present |
| `README.md` | Mixed-script sentence = inversion semantics | ✓ VERIFIED | Lines 149-151 |
| `internal/correct/runs.go` | Mixed branch inverts per char; anchor removed; homogeneous byte-identical | ✓ VERIFIED | `convertForeignRuns` deleted; `invertPerChar` at :118-140; zero "anchor" tokens remain in the file (grep — even the supersession cite uses "last-letter semantics") |
| `internal/correct/runs_test.go` | Golden mixed corpus re-pinned RED-first; homogeneous/refusal byte-identical | ✓ VERIFIED | `TestConvertRuns_MixedInvertsPerChar` 7 inversion cases; RED state independently reproduced at d7a8a5e |
| `internal/session/actor_test.go` | 3 pipeline tests re-pinned; DoubleTap untouched | ✓ VERIFIED | texts[6]/texts[0] = "паиghbdtn", phrase = "привет ghbdtn"; geometry pins unchanged; DoubleTap BYTE-IDENTICAL |
| `test/e2e/cases/matrix-v1..v4.yaml` | 13 mixed-output rows re-pinned; degradation/partial/homogeneous untouched | ✓ VERIFIED | Exact counts: "паиghbdtn" ×4 (v1:88, v2:103, v3:135, v4:166), "привет ghbdtn" ×9 (phrase ×3, select-all-gte ×3, select-all-chromium ×3); v4 header records the REJECTED freeze; all 23 remaining "привет привет"-family pins are legitimate homogeneous rows; select-all-zenity rows: 0 changed lines |

**Deviation audit (case_word.go):** SUMMARY declares re-pinning `test/e2e/case_word.go` beyond the plan's `files_modified`. Verified legitimate: `wordMixedConverted = "паи" + wordProbeEN` (= "паиghbdtn") is the live standalone word-mixed oracle — the same pin class as the 13 matrix rows; leaving it stale would have guaranteed a false failure at the owner's next live run. Included in commit 871ef9e; comment updated to the inversion semantics. Accepted.

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | -- | --- | ------ | ------- |
| runs.go mixed branch | layouts.ENToRU / layouts.RUToEN | `invertPerChar` table lookups (runs.go:124,126) | WIRED | Both tables imported and consulted per rune; the both-tables punctuation fact (tables.go:25/:120) grounds the neutral rule |
| runs_test.go golden corpus | ConvertRuns (out, changed, ok) contract | test assertions | WIRED | Every case asserts ok=true, changed=true, exact output; refusals assert ok=false, out=nil |
| e2e matrix mixed rows | ConvertRuns outputs | expect_text values | WIRED | The 13 re-pinned expect_text values are byte-identical to the unit-pinned outputs for the same inputs ("паиghbdtn", "привет ghbdtn") |
| SPEC §4.2 amendment | runs.go doc comment | same semantics statement | WIRED | Both state per-character inversion + explicit neutral rule + the verdict trace (runs.go:52-65 ↔ SPEC.md:102-117); runs.go adds the T-03-01-03 no-partial-leak clause |

### Data-Flow Trace (Level 4)

Not applicable — no rendered UI values in this change surface. The data-flow equivalent (commit_text payloads flow from real table lookups, not literals) is covered by the unit corpus: every expected output is computed by the implementation from `layouts` tables and asserted against independent pins.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| TDD RED before implementation (req. b) | temp worktree at d7a8a5e: `go test -race -count=1 ./internal/correct/ ./internal/session/` | exit 1; FAIL set = exactly TestConvertRuns_MixedInvertsPerChar, TestActor_MixedWordInvertsPerChar, TestActor_PhraseMixedCorrects, TestActor_SelectionMixedConverts | ✓ PASS (RED shape exact) |
| Preserved tests green at RED commit | worktree d7a8a5e: `-run 'TestConvertRuns_Refusals\|TestConvertRuns_HomogeneousWholesale'` and `-run TestActor_DoubleTapSelectionCorrects` | both ok | ✓ PASS |
| Whole module green (req. e, directive 2) | `mise run ci` (build + vet + golangci-lint + `go test -race -count=1 ./...`) | build ok, vet ok, lint 0 issues, all 19 packages ok | ✓ PASS |
| Matrix re-pin counts (req. d) | grep counts across v1..v4 | "паиghbdtn"=4, "привет ghbdtn"=9, REJECTED in v4 header=1, stale mixed pins=0 | ✓ PASS |
| Mixed-branch no-partial-leak refusal | TestConvertRuns_Refusals "unmapped letter fails wholesale" (input "ghé"+wordRU — mixed range, exercises invertPerChar's unmapped path) | ok — asserts ok=false, out=nil | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| test/e2e live matrix v1 | `mise run e2e-matrix` | not run — drives the owner's live GNOME desktop; plan Task 4 declares it the owner's MANUAL GATE | ? SKIP → human verification |
| test/e2e live matrix v2 | `mise run e2e-matrix-v2` | not run — same | ? SKIP → human verification |
| test/e2e live matrix v3 | `mise run e2e-matrix-v3` | not run — same | ? SKIP → human verification |
| test/e2e live matrix v4 | `mise run e2e-matrix-v4` | not run — same | ? SKIP → human verification |

The probe YAMLs and the standalone oracle (case_word.go) are statically verified to encode the new semantics and to build (`go build ./...` inside mise run ci covers test/e2e). Their live PASS is the owner's acceptance act.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| SPEC-4.2 | 261006-vqw-PLAN | Mixed-text correction semantics (§4.2) | ✓ SATISFIED | SPEC.md §4.2 amended per the owner verdict; semantics implemented and test-pinned end to end |
| OWNER-VERDICT-UAT-2026-10-03 | 261006-vqw-PLAN | Owner's locked verdict: per-character inversion; WINDOWS-12 rejected | ✓ SATISFIED | Trace present in SPEC amendment (todo 2026-10-05, verdict 2026-10-03, commit 7dd46e9), runs.go doc comment, matrix-v4 header, and test comments |

No orphaned requirements — this is a todo-driven quick task; the plan's `requirements` field is the authoritative claim set and both are satisfied. (`.planning/REQUIREMENTS.md` carries no phase-mapped IDs for this quick task; nothing orphaned.)

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | — | — | No TBD/FIXME/XXX/HACK/PLACEHOLDER/stub patterns in any of the 11 modified files; no empty implementations; no stale old-semantics pins anywhere in the mixed surface |

### Hard Requirements Audit (task goal a–f)

| Req | Statement | Status |
| --- | --------- | ------ |
| a | Spec-delta to SPEC §4.2 committed BEFORE any code (D-55); old semantics preserved verbatim; neutral behavior explicit | ✓ VERIFIED (47d8c71 docs-only, first of four; audit trail byte-identical; SPEC.md:106-111) |
| b | Golden mixed unit corpus re-pinned to inversion, TDD RED before implementation | ✓ VERIFIED (d7a8a5e RED independently reproduced: exact 4-test failing set; closed by cb0e968) |
| c | Per-character inversion implemented in the run pipeline | ✓ VERIFIED (runs.go invertPerChar; all three range paths funnel through ConvertRuns) |
| d | word-mixed/phrase-mixed e2e matrix rows (v1..v4) re-pinned | ✓ VERIFIED (13 rows + v4 header + case_word.go oracle; exact counts) |
| e | Non-regression: pure-EN/pure-RU, refusals, homogeneous corpora untouched and green | ✓ VERIFIED (byte-identity diffs + full suite green under -race) |
| f | UX contracts preserved (direction auto-detect, case preservation, exact-range replacement) | ✓ VERIFIED (direction.go untouched; case pin; geometry pins unchanged) |

### Human Verification Required

The 5 items in the frontmatter `human_verification` list — the four live matrices (`mise run e2e-matrix`, `-v2`, `-v3`, `-v4`) and the hand-typed mixed-text check in gedit/GTK4. These are the plan Task 4's declared MANUAL GATE: they drive the owner's live GNOME desktop and cannot be executed by an automated verifier. Everything automatable about them has been verified statically (values pinned, code compiles, unit+session behavior proven).

### Gaps Summary

No gaps. All six hard requirements are met with codebase and behavioral evidence; the SUMMARY's claims survived adversarial re-verification in every particular, including the declared case_word.go deviation (legitimate, verified) and the RED→GREEN TDD pair (independently reproduced via a throwaway worktree at d7a8a5e — the failing set was exactly the four re-pinned tests, and the preserved tests passed byte-identically against the old pipeline). The status is human_needed solely because the final acceptance gates are inherently manual live-desktop acts reserved to the owner.

---

_Verified: 2026-10-06_
_Verifier: Claude (gsd-verifier)_
