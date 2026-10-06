---
phase: 261006-vqw-mixed-text-per-character-layout-inversio
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
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
autonomous: true
requirements: [SPEC-4.2, OWNER-VERDICT-UAT-2026-10-03]
estimate:
  tokens: 55000
  raw_tokens: 55000
  tasks: 4
  confidence: low
must_haves:
  truths:
    - Mixed-text correction inverts EVERY letter per character through its script's key-position table (Latin letters via layouts.ENToRU, Cyrillic letters via layouts.RUToEN) with no anchor and no last-letter rule — "ghbdtn"+"привет" corrects to "привет"+"ghbdtn", "gfb"+"привет" corrects to "паи"+"ghbdtn" (owner verdict, UAT of phase 6, 2026-10-03, commit 7dd46e9; WINDOWS-12 freeze explicitly REJECTED)
    - Neutral characters in a mixed range pass through unchanged — digits and space are the stated cases ('5' inverts to '5' in both tables; space has no key position); table-mapped punctuation in a MIXED range also rides as typed because a per-character direction is undefined for it (',' is a source key of BOTH tables: ENToRU ','→'б', RUToEN ','→'?'); this is stated in SPEC §4.2, not left implicit
    - Non-mixed behavior is preserved byte-for-byte — single-script ranges keep the wholesale conversion including punctuation-with-the-token ("ghbdtn,"→"приветб", "ghbdtn["→"приветх", "привет["→"ghbdtn[") and the entire homogeneous/refusal corpora stay green untouched
    - The correction UX contracts compose with inversion unchanged — per-rune case preservation ("Gfb"+"привет"→"Паи"+"ghbdtn"), exact-range replacement geometry (-9,9 word / -13,13 phrase), direction auto-detection on single-script text, and the D-20 silent refusals (no letters; an unmapped Latin letter like é fails the WHOLE range, no partial output)
    - SPEC §4.2 records the new semantics as a dated amendment with the old anchor bullet preserved verbatim as audit trail and the owner-verdict trace, committed BEFORE any code (D-55 discipline); README.md describes the same semantics
  artifacts:
    - docs/SPEC.md §4.2 — dated spec-delta: per-character inversion as THE mixed-text semantics, D-22/D-23 anchor semantics superseded with ADR trace
    - README.md — mixed-script sentence matches the inversion semantics
    - internal/correct/runs.go — mixed branch inverts each letter via its script's table; anchor/foreign-run pass removed; homogeneous branch byte-identical
    - internal/correct/runs_test.go — golden mixed corpus re-pinned to inversion (RED first), homogeneous + refusal tests byte-identical
    - internal/session/actor_test.go — TestActor_MixedWordConvertsForeignRuns, TestActor_PhraseMixedCorrects, TestActor_SelectionMixedConverts re-pinned
    - test/e2e/cases/matrix-v1..v4.yaml — 13 mixed-output rows re-pinned (word-mixed ×4, phrase-mixed ×3, select-all-gte ×3, select-all-chromium ×3); degradation/partial rows untouched
  key_links:
    - runs.go mixed branch ↔ layouts.ENToRU / layouts.RUToEN per-key tables (layouts/tables.go:13, :112 — both tables carry punctuation as source keys, which is WHY mixed-range non-letters ride)
    - runs_test.go golden corpus ↔ ConvertRuns return contract (out, changed, ok) — mixed always ok=true, changed=true
    - e2e matrix mixed rows ↔ ConvertRuns outputs — expect_text values are the live oracle of the new semantics
    - SPEC §4.2 amendment ↔ runs.go doc comment — the same semantics statement in both places
---

<objective>
Replace the mixed-text correction semantics with per-character layout inversion, per the owner's locked verdict (UAT of phase 6, 2026-10-03, commit 7dd46e9; todo `.planning/todos/pending/mixed-text-inversion.md`, 2026-10-05). Today `ConvertRuns` (internal/correct/runs.go:68) corrects ONLY the foreign runs relative to a last-letter anchor (D-22/D-23, 03-01); the target semantics inverts EVERY letter independently — Latin through ENToRU, Cyrillic through RUToEN — so "ghbdtn привет" corrects to "привет ghbdtn" instead of "привет привет". The owner explicitly rejected the WINDOWS-12 freeze proposal (freezing readback strings as registry current); this is a semantic change, spec-delta first.

Purpose: the anchor semantics answers the wrong question for mixed text — the user's intent when text is half-wrong is "flip every character", not "fix the foreign runs only".
Output: SPEC §4.2 spec-delta (docs-only commit first), re-pinned golden unit corpus, inverted mixed branch in runs.go, re-pinned e2e matrix rows v1..v4, whole-module green.
</objective>

<execution_context>
@/home/nil/DiskD/W/Djarvur/goswitch/.zcode/gsd-core/workflows/execute-plan.md
@/home/nil/DiskD/W/Djarvur/goswitch/.zcode/gsd-core/templates/summary.md
</execution_context>

<context>
@/home/nil/DiskD/W/Djarvur/goswitch/.planning/todos/pending/mixed-text-inversion.md
@/home/nil/DiskD/W/Djarvur/goswitch/docs/SPEC.md
@/home/nil/DiskD/W/Djarvur/goswitch/internal/correct/runs.go
@/home/nil/DiskD/W/Djarvur/goswitch/internal/correct/convert.go
@/home/nil/DiskD/W/Djarvur/goswitch/layouts/tables.go

Measured facts (verified live on branch gsd/mixed-text-per-char-inversion, 2026-10-06):
- The ONE conversion entry is `correct.ConvertRuns` (internal/correct/runs.go:68); all three range paths funnel through it: word/phrase via `startRangeCorrection` (internal/session/actor.go:2600, call at :2612), selection via `startSelectionCorrection` (actor.go:2529). Autocorrect (opt-in, detect.Check at actor.go:2158) only fires on detector-confident single-script tokens — out of scope, untouched.
- The mixed branch is `convertForeignRuns` (runs.go:102): anchor = script of the LAST letter, foreign runs convert, anchor runs stay. The homogeneous branch (runs.go:89-94) calls `Convert` (convert.go:20) wholesale — punctuation converts with the token there (ENToRU[',']='б', layouts/tables.go:25).
- The tables are bijective per layout, but punctuation is a source key of BOTH tables with DIFFERENT images (ENToRU[',']='б', tables.go:25; RUToEN[',']='?', tables.go:120) — a per-character direction for non-letters is undefined; digits map to themselves (ENToRU['5']='5'), space has no entry. This grounds the neutral rule.
- Unit pins to the OLD semantics: internal/correct/runs_test.go TestConvertRuns_GoldenCorpus (:11, 5 mixed cases) and the mixed case "register per rune on the foreign run" inside TestConvertRuns_HomogeneousWholesale (:104-106); internal/session/actor_test.go TestActor_MixedWordConvertsForeignRuns (:1404), TestActor_PhraseMixedCorrects (:1453), TestActor_SelectionMixedConverts (:1777). TestConvertRuns_Refusals (:136) and TestActor_DoubleTapSelectionCorrects (:1696, homogeneous selected head) must stay green byte-identical.
- e2e pins to the OLD semantics (live gates, NOT in CI): word-mixed at matrix-v1.yaml:88, v2:103, v3:135, v4:166 (all expect the anchor-semantics output); phrase-mixed at v2:53, v3:85, v4:116; select-all-gte at v2:128, v3:160, v4:191 (whole mixed field selected → converts); select-all-chromium at v2:141, v3:173, v4:204. Rows that must NOT change: select-all-zenity (D-30 degradation — daemon refuses, field settles client-side), select-partial-gte and select-reverse-gte (only the homogeneous head [0,6) is selected), every homogeneous row. matrix-v4.yaml:27-31 header comment pins the WINDOWS-12 freeze verdict that the owner has since REJECTED — it must be re-worded.
- Green gates are mise tasks (mise.toml): `mise run ci` = build + vet + lint + test -race; `go.mod` is untouched by this work, `mise run tidy-diff` trivially green.
- Spec-delta house style (precedent 9bebd67): dated HTML-comment amendment, old text kept verbatim as audit trail, new text below, docs-only commit `docs(<id>): spec-delta — … (D-55)`.
</context>

<tasks>

<task type="auto">
  <name>Task 1: Spec-delta — per-character inversion into SPEC §4.2 (docs-only, D-55)</name>
  <files>docs/SPEC.md, README.md</files>
  <precondition>On branch gsd/mixed-text-per-char-inversion with a clean tree; docs/SPEC.md §4.2 currently reads the anchor-era bullet at lines 92-93.</precondition>
  <action>
    Edit docs/SPEC.md §4.2. Keep the current mixed-text bullet ("Смешанный текст: корректируется только часть, набранная не той раскладкой, относительно последней использованной." including its `[Q: определить точную семантику]` marker) VERBATIM in place as audit trail, preceded by a dated HTML comment in the house style (precedent 9bebd67): "Правка закреплённого решения — решение владельца 2026-10-05 (todo «Смешанный текст: посимвольная инверсия раскладки», вердикт UAT фазы 6 от 2026-10-03, коммит 7dd46e9): семантика якоря последней буквы (D-22/D-23, CONTEXT 03-01) снимается; предложение WINDOWS-12 о заморозке readback-строк ОТКЛОНЕНО владельцем." Below it add the new bullet defining THE mixed-text semantics:
    (a) mixed text corrects by PER-CHARACTER layout inversion: every Latin letter maps through the ENToRU key-position table, every Cyrillic letter through RUToEN, independently of position and of any anchor; register is preserved per rune (constructive, same as the homogeneous path);
    (b) the neutral rule, stated explicitly: a character whose inversion equals itself passes through unchanged — digits ('5' inverts to '5' in both tables) — as does a character with no key position (space); table-mapped punctuation and symbols in a MIXED range also ride as typed, because a per-character direction is undefined for them (',' is a source key of both tables with different images), unlike a WHOLESALE single-script range where punctuation still converts with the token;
    (c) a letter missing from its script's table (é) fails the WHOLE range silently (D-20 continuation, no partial output);
    (d) single-script text keeps the existing wholesale conversion — direction auto-detection, punctuation-with-token, everything in §4.2 above unchanged; replacement geometry and case preservation (§4.3) compose with inversion unchanged.
    Then reword README.md lines 149-150 ("Words already in the target layout and mixed-script words are converted point-wise — only the foreign letters change.") to state that mixed-script text inverts per character: every letter flips to the other layout's key position, digits/space/punctuation ride unchanged.
    Commit ONLY these two files: `docs(261006-vqw): spec-delta — per-character layout inversion as THE mixed-text semantics (D-55)`. No code, no tests, no planning files in this commit.
  </action>
  <verify>
    <automated>git show --stat --format= HEAD > /tmp/vqw-stat.txt && test "$(grep -cE '^\s*(docs/SPEC\.md|README\.md)' /tmp/vqw-stat.txt)" = "2" && test "$(wc -l < /tmp/vqw-stat.txt)" = "2" && grep -c "посимвольная инверсия" docs/SPEC.md | grep -qx "[1-9]" && grep -c "корректируется только часть, набранная не той раскладкой" docs/SPEC.md | grep -qx "1" && grep -c "inverts per character" README.md | grep -qx "1"</automated>
  </verify>
  <done>HEAD is a docs-only commit touching exactly docs/SPEC.md and README.md; SPEC §4.2 carries the old bullet verbatim exactly once (audit trail) plus the new inversion semantics with the owner-verdict trace and the explicit neutral rule; README matches.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Re-pin the golden unit corpus to inversion semantics (TDD RED)</name>
  <files>internal/correct/runs_test.go, internal/session/actor_test.go</files>
  <precondition>Task 1's docs-only commit is HEAD; the working tree is otherwise clean, so every failure observed in this task's verify is attributable to the corpus re-pin.</precondition>
  <action>
    Tests first — pin the NEW semantics so the current pipeline fails (RED). In internal/correct/runs_test.go:
    (a) Rename TestConvertRuns_GoldenCorpus to TestConvertRuns_MixedInvertsPerChar and re-pin its five mixed cases to the inversion outputs, updating each case's name/comment/why to name the inversion semantics and the supersession of the D-22/D-23 anchor (spec-delta 2026-10-06): wordEN+wordRU → wordRU+wordEN; "gfb"+wordRU → "паи"+wordEN; wordRU+wordEN → wordEN+wordRU; wordEN+"2026"+wordRU → wordRU+"2026"+wordEN (digits ride, letters all invert); "gfb"+wordRU+"[" → "паи"+wordEN+"[" (the bracket rides in a mixed range). The fixture constants live in internal/correct/buffer_test.go:12-15.
    (b) Move the mixed case "register per rune on the foreign run" (currently TestConvertRuns_HomogeneousWholesale, runs_test.go:104-106) out of that test into the mixed corpus, re-pinned: "Gfb"+wordRU → "Паи"+wordEN. Add one space-riding mixed case (e.g. "gfb "+wordRU → "паи "+wordEN).
    (c) Leave TestConvertRuns_HomogeneousWholesale's remaining cases and TestConvertRuns_Refusals BYTE-IDENTICAL — they pin the preserved behavior (including the é whole-range refusal, which inversion keeps, and the wholesale punctuation rows "ghbdtn,"→"приветб", "ghbdtn["→"приветх", "привет["→"ghbdtn[").
    In internal/session/actor_test.go re-pin the three pipeline tests to the inversion outputs (geometry assertions unchanged — inversion never changes the deleted range): TestActor_MixedWordConvertsForeignRuns → rename TestActor_MixedWordInvertsPerChar, want commits texts[6] = "паи"+wordEN; TestActor_PhraseMixedCorrects, want texts[6] = wordRU+" "+wordEN, doc comment updated to inversion (D-26 under the new semantics); TestActor_SelectionMixedConverts, want texts[0] = "паи"+wordEN, comment updated. Do NOT touch TestActor_DoubleTapSelectionCorrects — its selected range [0,6) is homogeneous and stays pinned to a single-word commit.
    Commit: `test(261006-vqw): re-pin golden mixed-text corpus to per-character inversion (RED)`. This is the RED pole of the pair with Task 3 — a deliberately failing intermediate state; the branch is not pushed until Task 4's full green.
  </action>
  <verify>
    <automated>grep -c '"паи"+wordEN\|паиghbdtn' internal/correct/runs_test.go internal/session/actor_test.go | grep -qv ':0' && go test -race -count=1 ./internal/correct/ ./internal/session/ 2>&1 | tee /tmp/vqw-red.log | grep -qE '^(FAIL|--- FAIL)' && grep -qE '--- FAIL.*(TestConvertRuns_MixedInvertsPerChar|TestActor_MixedWordInvertsPerChar|TestActor_PhraseMixedCorrects|TestActor_SelectionMixedConverts)' /tmp/vqw-red.log && go test -race -count=1 -run 'TestConvertRuns_Refusals|TestConvertRuns_HomogeneousWholesale' ./internal/correct/</automated>
  </verify>
  <done>The corpus is pinned to inversion: go test over ./internal/correct/ and ./internal/session/ FAILS, and the failing set is exactly the re-pinned mixed tests (TestConvertRuns_MixedInvertsPerChar, TestActor_MixedWordInvertsPerChar, TestActor_PhraseMixedCorrects, TestActor_SelectionMixedConverts); TestConvertRuns_Refusals and TestConvertRuns_HomogeneousWholesale (now mixed-free after the case move) still PASS against the current code.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Implement per-character inversion in the mixed branch (TDD GREEN)</name>
  <files>internal/correct/runs.go, internal/session/actor.go</files>
  <precondition>Task 2's RED corpus commit is HEAD; the named mixed tests fail against the current convertForeignRuns implementation.</precondition>
  <action>
    Make the RED corpus green. In internal/correct/runs.go: replace the anchor pass `convertForeignRuns` (runs.go:102-149, the D-22/D-23 last-letter semantics) with a per-character inversion pass over the mixed range: iterate the runes once; a Latin letter (classifyRune == scriptLatin) maps through layouts.ENToRU, a Cyrillic letter (scriptCyrillic) through layouts.RUToEN; a letter with no entry in its script's table fails the WHOLE range with ok=false and a nil output before any rune is emitted (T-03-01-03 carry-over: no partial conversion ever leaks); a scriptNeutral rune appends unchanged. Keep intact: the no-letters refusal (runs.go:84-85), the homogeneous wholesale branch verbatim — `Convert(text, dirOf(...))` at runs.go:89-94 must not change behavior (wholesale punctuation conversion is the byte-preservation surface) — and the (out, changed, ok) contract, where a mixed range always returns ok=true, changed=true. Rewrite the ConvertRuns doc comment: the mixed bullet states per-character inversion with the spec-delta trace (owner verdict 2026-10-03 commit 7dd46e9, spec-delta 2026-10-06, D-22/D-23 superseded, WINDOWS-12 rejected) and the explicit neutral rule (digits and space pass because their inversion equals themselves / no key position; mixed-range punctuation rides because no per-character direction exists — cite the both-tables fact); reference the superseded decision by its IDs and the delta date, without re-describing the old mechanism anywhere in the file. In internal/session/actor.go update the now-stale comments only — the startRangeCorrection doc (~actor.go:2586) and any other comment still describing the superseded mixed mechanism — to the inversion wording; zero behavior change in actor.go. Do NOT touch convert.go, direction.go (Detect stays for autocorrect and homogeneous use), buffer.go, plan.go, verify.go.
    Commit: `feat(261006-vqw): per-character layout inversion in the mixed-run pipeline`.
  </action>
  <verify>
    <automated>go test -race -count=1 ./internal/correct/ ./internal/session/ && mise run ci</automated>
  </verify>
  <done>The full re-pinned corpus passes under -race; `mise run ci` (build + vet + golangci-lint + test -race) is green; runs.go no longer contains any description of the superseded mixed mechanism (doc comment cites the decision IDs and the delta date instead); homogeneous/refusal tests pass byte-identically.</done>
</task>

<task type="auto">
  <name>Task 4: Re-pin e2e matrices v1..v4 + whole-module green iteration</name>
  <files>test/e2e/cases/matrix-v1.yaml, test/e2e/cases/matrix-v2.yaml, test/e2e/cases/matrix-v3.yaml, test/e2e/cases/matrix-v4.yaml</files>
  <precondition>Task 3 is green; ConvertRuns already produces the inversion outputs the re-pinned rows encode.</precondition>
  <action>
    Re-pin the 13 mixed-output rows to the inversion semantics — each row's expect_text AND its trailing comment (comments currently name the anchor/D-23 semantics and must name the inversion semantics, spec-delta 2026-10-06):
    - word-mixed rows → the paи-form inversion output (matrix-v1.yaml:88, matrix-v2.yaml:103, matrix-v3.yaml:135, matrix-v4.yaml:166).
    - phrase-mixed rows → the swapped-words inversion output (matrix-v2.yaml:53, matrix-v3.yaml:85, matrix-v4.yaml:116).
    - select-all-gte rows → the swapped-words output (the ctrl+a selection covers the whole mixed field, so the whole field inverts; matrix-v2.yaml:128, matrix-v3.yaml:160, matrix-v4.yaml:191).
    - select-all-chromium rows → the same swapped-words output (matrix-v2.yaml:141, matrix-v3.yaml:173, matrix-v4.yaml:204).
    MUST NOT change: select-all-zenity (D-30 degradation — the daemon refuses, the client eats the commit), select-partial-gte and select-reverse-gte (only the homogeneous head [0,6) is selected — single-script wholesale path), and every homogeneous row. Update the matrix-v4.yaml header comment (lines 27-31) that records the WINDOWS-12 freeze as the rows' ground truth: the freeze was REJECTED by the owner (2026-10-05 todo; verdict 2026-10-03, commit 7dd46e9) and the rows are re-pinned to per-character inversion. Update the matrix-v2.yaml header line 3 ("mixed text") only if it names the anchor semantics.
    Then the green iteration: `mise run ci` over the whole module (strict-TDD directive 2 — no lint debt). Commit: `test(261006-vqw): re-pin e2e matrix mixed rows to per-character inversion (v1..v4)`.
    MANUAL GATE for the owner (not automatable here; list in the plan SUMMARY as the final acceptance item): run the live matrices on the GNOME stand — `mise run e2e-matrix`, `mise run e2e-matrix-v2`, `mise run e2e-matrix-v3`, `mise run e2e-matrix-v4` — then a hand check: type mixed text in gedit/GTK4 (e.g. latin word, flip, cyrillic word), hit the correction hotkey, and confirm every letter inverted with digits/space/punctuation intact and no visual jump.
  </action>
  <verify>
    <automated>test "$(grep -h -c 'expect_text: "паиghbdtn"' test/e2e/cases/matrix-v1.yaml test/e2e/cases/matrix-v2.yaml test/e2e/cases/matrix-v3.yaml test/e2e/cases/matrix-v4.yaml | paste -sd+ | bc)" = "4" && test "$(grep -h -c 'expect_text: "привет ghbdtn"' test/e2e/cases/matrix-v1.yaml test/e2e/cases/matrix-v2.yaml test/e2e/cases/matrix-v3.yaml test/e2e/cases/matrix-v4.yaml | paste -sd+ | bc)" = "9" && grep -c "REJECTED\|отклонён" test/e2e/cases/matrix-v4.yaml | grep -qx "[1-9]" && mise run ci</automated>
  </verify>
  <done>The four matrices carry exactly 4 word-mixed rows and 9 phrase-mixed/select-all rows pinned to the inversion outputs; the v4 header records the freeze rejection; degradation/partial/homogeneous rows are untouched; `mise run ci` is green across the module.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| client text → correction pipeline | Untrusted field content (surrounding text pushes) crosses into ConvertRuns; inversion adds no new input surface — same runes, same tables |
| e2e matrices → runner | YAML case files drive the live desktop; edited values are expected texts only |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-VQW-01 | Tampering | ConvertRuns mixed branch | medium | mitigate | Whole-range refusal on an unmapped letter is preserved (no partial output ever leaks — T-03-01-03 carry-over); the é refusal case stays pinned byte-identical in TestConvertRuns_Refusals |
| T-VQW-02 | Tampering | homogeneous path regression | high | mitigate | The mixed branch is the ONLY changed code path; homogeneous corpus + matrix v1 rows stay byte-identical and gate every verify (task 2/3/4); the wholesale punctuation goldens ("приветб", "приветх") remain pinned |
| T-VQW-03 | Denial of Service | inversion pass over a phrase-length range | low | accept | Single O(n) map lookups over a human-typed phrase, same complexity as the removed anchor pass; well inside the <50 ms budget |
| T-VQW-04 | Repudiation | SPEC ↔ code drift | medium | mitigate | Spec-delta committed BEFORE code (D-55, task 1); runs.go doc comment cites the same verdict; SUMMARY records the manual e2e gate |
</threat_model>

<verification>
- `mise run ci` green on the final commit (build + vet + golangci-lint + test -race -count=1 ./...).
- `go test -race -count=1 ./internal/correct/ ./internal/session/` green with the re-pinned corpus.
- The 13 matrix rows grep-pinned to the inversion outputs (4 + 9, exact counts in task 4 verify).
- Homogeneous preservation: TestConvertRuns_HomogeneousWholesale (minus the moved mixed case), TestConvertRuns_Refusals, TestActor_DoubleTapSelectionCorrects, and all matrix homogeneous/degradation/partial rows unchanged from the pre-plan tree (`git diff` over those regions shows comment-free byte identity).
</verification>

<success_criteria>
- SPEC §4.2 states per-character inversion as THE mixed-text semantics with the old bullet preserved verbatim and the owner-verdict trace, in a docs-only commit that precedes all code commits.
- Mixed correction inverts every letter: "ghbdtn привет"→"привет ghbdtn", "gfb привет"→"паи ghbdtn", digits/space/punctuation ride, register preserved per rune, unmapped letter refuses the whole range.
- Pure-EN and pure-RU correction behavior, corpora, and e2e rows are byte-for-byte unchanged and green.
- All four e2e matrices re-pinned; the owner's manual acceptance items (live matrices + hand-typed mixed check in gedit/GTK4) are listed as the final human gate.
</success_criteria>

<output>
Create `.planning/quick/261006-vqw-mixed-text-per-character-layout-inversio/261006-vqw-SUMMARY.md` when done (the orchestrator handles the todo state — do not move `.planning/todos/pending/mixed-text-inversion.md` from the executor).
</output>
