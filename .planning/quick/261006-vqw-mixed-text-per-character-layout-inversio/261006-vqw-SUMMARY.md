---
phase: 261006-vqw-mixed-text-per-character-layout-inversio
plan: 01
subsystem: correct
tags: [correction, mixed-text, layout-inversion, spec-delta, tdd]
requires:
  - SPEC-4.2 (mixed-text correction semantics — amended by this task)
  - layouts.ENToRU / layouts.RUToEN key-position tables (layouts/tables.go)
  - correct.ConvertRuns (internal/correct/runs.go) — the ONE conversion entry
provides:
  - per-character layout inversion as THE mixed-text semantics (spec-delta 2026-10-06)
  - correct.invertPerChar — the mixed-branch implementation
  - re-pinned golden mixed corpus (unit + session pipeline + e2e matrices v1..v4)
affects:
  - word/phrase/selection correction of mixed EN+RU ranges (actor.go pipeline, unchanged code)
  - e2e live oracles (matrix rows + case_word.go wordMixedConverted)
tech-stack:
  added: []
  patterns:
    - per-rune script-classified table inversion (no anchor, no run segmentation)
    - spec-delta BEFORE code (D-55) with dated HTML-comment audit trail (precedent 9bebd67)
key-files:
  created: []
  modified:
    - docs/SPEC.md
    - README.md
    - internal/correct/runs.go
    - internal/correct/runs_test.go
    - internal/session/actor.go
    - internal/session/actor_test.go
    - test/e2e/cases/matrix-v1.yaml
    - test/e2e/cases/matrix-v2.yaml
    - test/e2e/cases/matrix-v3.yaml
    - test/e2e/cases/matrix-v4.yaml
    - test/e2e/case_word.go
decisions:
  - "Mixed-text semantics = per-character layout inversion (owner verdict, UAT of phase 6, 2026-10-03, commit 7dd46e9; todo 2026-10-05); D-22/D-23 anchor superseded, WINDOWS-12 readback freeze rejected — recorded in SPEC §4.2 as a dated amendment BEFORE any code (D-55)"
  - "Neutral rule (SPEC + code): self-inverse runes (digits) and keyless runes (space) pass through; mixed-range punctuation rides as typed — a per-character direction is undefined (',' is a source key of BOTH tables); wholesale single-script conversion unchanged byte-for-byte"
  - "Mixed range always converts (ok=true, changed=true): every mapped letter inverts into the other script; the D-24 changed=false branch stays in actor.go as the defensive no-op only"
  - "RED commit deliberately failing (d7a8a5e) closed by the GREEN commit (cb0e968) — strict TDD pair; the branch is not pushed until the full green (Task 4)"
metrics:
  duration: 19 min
  completed: 2026-10-06
status: complete
actuals:
  tokens: 10678   # chars/4 over the realized diff (42714 chars, 11 files, +229/-208)
  tasks: 4
  commits: 4      # MEASURED: git rev-list --count 0bec17f..HEAD
plan_head_before: 0bec17fa3fa5330bd4e1a28e23ee805aaf4fb13f
---

# Quick Task 261006-vqw: Mixed-Text Per-Character Layout Inversion Summary

**One-liner:** Mixed-text correction now inverts EVERY letter per character through its script's key-position table ("gfb"+"привет" → "паиghbdtn"), replacing the D-22/D-23 last-letter anchor — spec-delta committed before code, golden corpus re-pinned RED→GREEN, all four e2e matrices re-pinned, whole module green.

## What Was Done

- **Task 1 — spec-delta (docs-only, D-55), commit 47d8c71:** SPEC §4.2 gained a dated 2026-10-06 amendment (house style, precedent 9bebd67) defining per-character inversion as THE mixed-text semantics with the owner-verdict trace (UAT of phase 6, 2026-10-03, commit 7dd46e9; WINDOWS-12 REJECTED). The old anchor-era bullet stays verbatim above as audit trail. The neutral rule, the é whole-range refusal (D-20), and the untouched single-script wholesale conversion are stated explicitly. README lines 149-151 reworded to match.
- **Task 2 — RED corpus, commit d7a8a5e:** `TestConvertRuns_GoldenCorpus` renamed `TestConvertRuns_MixedInvertsPerChar` with seven mixed cases pinned to inversion outputs (wordEN+wordRU → wordRU+wordEN; gfb → паиghbdtn; digits/space/bracket ride; register per rune "Паиghbdtn"; plus a new space-riding case). Session tests re-pinned: `TestActor_MixedWordInvertsPerChar` (texts[6]="паиghbdtn"), `TestActor_PhraseMixedCorrects` (texts[6]="привет ghbdtn", D-26 under inversion), `TestActor_SelectionMixedConverts` (texts[0]="паиghbdtn") — geometry pins unchanged. Verified RED: the failing set was exactly the four re-pinned tests; `TestConvertRuns_Refusals` and `TestConvertRuns_HomogeneousWholesale` (minus the moved mixed case) passed byte-identically against the old pipeline.
- **Task 3 — GREEN implementation, commit cb0e968:** `convertForeignRuns` (the anchor pass) removed; the mixed branch of `ConvertRuns` now calls `invertPerChar` — one pass, Latin through `layouts.ENToRU`, Cyrillic through `layouts.RUToEN`, neutrals appended unchanged, an unmapped letter fails the WHOLE range with nil output before any rune is emitted (T-03-01-03 carry-over). The homogeneous wholesale branch (`Convert(text, dirOf(script))`) is behavior-identical. The doc comment cites the decision IDs and the delta date without re-describing the old mechanism (verified by grep: the only "anchor" mention is the supersession citation). actor.go: comment-only updates (startRangeCorrection/startSelectionCorrection docs, both D-24 comments, correctionRange doc, refusalReason doc) — zero behavior change. Test files carry the lint-shape fixes from the green iteration (funlen/goconst/lll).
- **Task 4 — e2e re-pins, commit 871ef9e:** 13 mixed-output rows re-pinned across matrices v1..v4 with comments naming the inversion semantics: word-mixed ×4 → `паиghbdtn`; phrase-mixed ×3, select-all-gte ×3, select-all-chromium ×3 → `привет ghbdtn`. The matrix-v4 header now records the WINDOWS #12 freeze as REJECTED and the rows as re-pinned to spec-delta 2026-10-06.

## Verification Results

- `mise run ci` (build + vet + golangci-lint v2 strict + `go test -race -count=1 ./...`) — GREEN on the final tree; `mise run tidy-diff` — GREEN (go.mod/go.sum untouched).
- `go test -race -count=1 ./internal/correct/ ./internal/session/` — GREEN with the re-pinned corpus.
- Matrix grep pins (exact per plan): `expect_text: "паиghbdtn"` = 4 across v1..v4; `expect_text: "привет ghbdtn"` = 9; `REJECTED` present in matrix-v4 header.
- Byte-preservation (diffed against the pre-plan tree 0bec17f): `TestConvertRuns_Refusals` and `TestActor_DoubleTapSelectionCorrects` BYTE-IDENTICAL; `TestConvertRuns_HomogeneousWholesale` case table identical minus the deliberately moved mixed case; convert.go/direction.go/buffer.go/plan.go/verify.go untouched; matrix degradation (select-all-zenity, D-30), partial/reverse-selection (homogeneous head [0,6)), and every homogeneous row untouched (diff shows exactly the 13 rows + v4 header, nothing else).
- No tracked file deletions; no new untracked files.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical re-pin] Re-pinned the live word-mixed oracle in test/e2e/case_word.go**
- **Found during:** Task 4 pre-flight scan (grep for remaining old-semantics pins)
- **Issue:** The plan's `files_modified` and task list omit `test/e2e/case_word.go`, whose `wordMixedConverted = "паи" + wordResultRU` ("паипривет") is the live standalone word-mixed oracle (runWordMixed, e2e driver) — the same class of pin as the 13 matrix rows. Left stale, the owner's next live word-mixed run would fail against a CORRECT daemon.
- **Fix:** `wordMixedConverted = "паи" + wordProbeEN` ("паиghbdtn"); the three related comments reworded to the inversion semantics. Included in the Task 4 commit; go build/vet/lint green.
- **Files modified:** test/e2e/case_word.go
- **Commit:** 871ef9e

No other deviations — the plan executed as written otherwise.

## MANUAL GATE for the owner (final acceptance items — not automatable here)

1. Live matrices on the GNOME stand: `mise run e2e-matrix`, `mise run e2e-matrix-v2`, `mise run e2e-matrix-v3`, `mise run e2e-matrix-v4` — the 13 re-pinned mixed rows must PASS against the re-built daemon.
2. Hand check: type mixed text in gedit/GTK4 (e.g. latin word, flip, cyrillic word), hit the correction hotkey, and confirm every letter inverted with digits/space/punctuation intact and no visual jump.

## Threat Model Mitigations (all applied)

- **T-VQW-01** (whole-range refusal preserved): `invertPerChar` returns `(nil, false, false)` on an unmapped letter BEFORE any rune is emitted; the é refusal case stayed green byte-identically through RED and GREEN.
- **T-VQW-02** (homogeneous regression): the mixed branch is the only changed code path; the homogeneous corpus, refusals, DoubleTapSelection, and matrix v1 homogeneous rows gate every verify — byte-identical, green.
- **T-VQW-03** (DoS): accepted per plan — single O(n) map-lookup pass, same complexity as the removed anchor pass.
- **T-VQW-04** (SPEC ↔ code drift): spec-delta committed BEFORE code (47d8c71 precedes all code commits); runs.go doc comment cites the same verdict; this SUMMARY records the manual e2e gate.

## Self-Check: PASSED

- Commit 47d8c71 present (docs-only, exactly docs/SPEC.md + README.md) — verified via `git show --name-only`.
- Commit d7a8a5e present (RED corpus; verified failing set exactly the four re-pinned tests at execution time).
- Commit cb0e968 present (GREEN implementation; `mise run ci` green).
- Commit 871ef9e present (e2e re-pins; matrix counts 4+9 verified; tidy-diff green).
- All 11 modified files exist in the working tree; SUMMARY written to the plan directory; no docs artifacts committed by the executor (orchestrator handles the docs commit).
