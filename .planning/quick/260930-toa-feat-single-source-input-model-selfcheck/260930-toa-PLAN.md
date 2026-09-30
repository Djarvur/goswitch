---
phase: 260930-toa-feat-single-source-input-model-selfcheck
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/install/install.go
  - internal/install/install_internal_test.go
  - internal/install/install_test.go
  - internal/install/selfcheck.go
  - internal/install/selfcheck_test.go
  - test/e2e/case_install.go
  - test/e2e/case_switch.go
  - README.md
  - docs/adr/ADR-006-two-engine-revision.md
autonomous: true
requirements:
  - QUICK-260930-toa-single-source-input-model
estimate:
  tokens: 52000
  raw_tokens: 26000
  tasks: 4
  confidence: low
must_haves:
  truths:
    - goswitchctl install accepts a desktop carrying a SINGLE xkb 'us' or single xkb 'ru' source and wraps it into exactly its goswitch engine ([('xkb','us')] → [('ibus','goswitch-en')], [('xkb','ru')] → [('ibus','goswitch-ru')]); the us+ru pair still wraps to both engines with positions preserved; already-goswitch lists keep the skip-if-equal takeover (owner decision 2026-09-30: one source hides the dead native GNOME indicator).
    - A wrap is refused whenever it would leave a foreign source beside goswitch engines (xkb 'us'+'fr', us+ru+fr) or when nothing wrappable exists — the existing named refusals fire (errMixedSources / errUnsupportedPair), mixed goswitch+foreign inputs and unsupported kinds keep today's refusals, and the atomic refusal gate still fails BEFORE any mutation and without creating the state file.
    - goswitchctl selfcheck's input-source step is green when the sources carry AT LEAST ONE goswitch engine and zero foreign entries, and the verdict names the engines found ("ok input-source goswitch-en"); zero goswitch engines and goswitch-beside-foreign are distinct red verdicts, each naming its fix (D-53: the daemon sees no keys through an xkb source).
    - The e2e install-cycle oracle recomputes the ≥1 wrap for single-source snapshots, and the ibus-restart extension's desktop precondition accepts any goswitch-owned sources count; the runtime flip oracles (journal pair + readback in busFlipRound) stay untouched.
    - README documents the single-source user model (one goswitch source recommended, native indicator hidden by design, goswitch's own SNI tray icon is the indicator, GNOME switch chords left untouched, switching = the daemon's gestures) and does NOT document tray-click flipping; ADR-006 carries an amendment section with evidence pointers; SPEC.md is untouched (owner spec-delta route, deferred note recorded).
  artifacts:
    - internal/install/install.go (wrapSources ≥1 gate + foreign-residue refusal, carriesWrappableXKB, reworded static errors, wrappedEngines recorded for the report)
    - internal/install/install_internal_test.go and install_test.go (single-source success corpus, updated refusal table, single-source install sequence, report pins)
    - internal/install/selfcheck.go and selfcheck_test.go (≥1+no-foreign predicate, named-engines verdict, reworded errSourceNotOwner, mixed-red and single-green tests)
    - test/e2e/case_install.go and test/e2e/case_switch.go (oracle wraps ≥1 source; goswitch-count-agnostic desktop precondition)
    - README.md (rewritten input-source handover and indicator model) and docs/adr/ADR-006-two-engine-revision.md (Amendment 2026-09-30 section + status pointer)
  key_links:
    - resolveWrapInput → carriesWrappableXKB → wrapSources: the atomic refusal gate admits single-source desktops before any mutation (refusal ⇒ no state file, zero mutating calls)
    - takeoverSources → wrapSources → Installer.wrappedEngines → report(): the install report names the engines actually wrapped (the rendered line reads sources wrapped goswitch-en)
    - selfcheck checkInputSource → activate.ParseSourceTuples + countOwnedForeign: the SAME sources grammar the installer uses — one parser on both sides, never substring guessing
    - e2e wrapForCheck/carriesOracleSource ↔ production wrapSources: the oracle restates (never imports) the ≥1 rule, so a wrap bug still cannot hide
    - README/ADR-006 amendment ↔ shipped behavior: the documented model (one source, tray icon, chords untouched) matches the code exactly
---

<objective>
feat: single-source input model — installer and selfcheck accept 1..N goswitch-wrapped sources.

Purpose: owner decision 2026-09-30 evening — «в гномовском переключателе остается один источник». GNOME 46 hides its native input indicator with fewer than 2 sources, and on this desktop that indicator's switching handler is dead anyway (proven exhaustively); the visible indicator is goswitchd's own SNI tray icon (quick 260930-pf6, live). This amends ADR-006's two-source scheme: the daemon-side flip (SetGlobalEngine), the 05-04 sync listener and the tray icon are UNAFFECTED — engines are registered by the component on the IBus bus, not by the sources list, and SetGlobalEngine works for any registered engine (v1.0.0 ran a single [('ibus','goswitch-en')] source with internal flips).

Output: wrapSources/resolveWrapInput accept at least one wrappable xkb source (single 'us' → goswitch-en, single 'ru' → goswitch-ru, pair → both, unchanged); selfcheck's input-source step is green on ≥1 goswitch engine with zero foreign entries and names what it found; the e2e oracle and the ibus-restart precondition follow; README + ADR-006 document the new model. Strict TDD per task (RED commit → GREEN commit), every iteration green on `mise run ci`, atomic conventional commits.
</objective>

<execution_context>
@/home/nil/DiskD/W/Djarvur/goswitch/.zcode/gsd-core/workflows/execute-plan.md
@/home/nil/DiskD/W/Djarvur/goswitch/.zcode/gsd-core/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@internal/install/install.go
@internal/install/selfcheck.go
@internal/install/install_internal_test.go
@internal/install/install_test.go
@internal/install/selfcheck_test.go
@test/e2e/case_install.go
@README.md
@docs/adr/ADR-006-two-engine-revision.md

Load the go-ultimate project skill (.zcode/skills/go-ultimate/) before writing Go — conventions and review checklist come from there.

Wiring facts (line anchors on the current tree):
- install.go: consts kindXKB..pairTuples ~125-134; static errors ~137-162; Install doc comment ~326-333; saveState gate comment ~487-491; carriesXKBPair ~697-714; countOwnedForeign ~718-728; renderWrapped ~734-754; wrapSources ~766-789; resolveWrapInput ~798-825; takeoverSources ~833-854; report ~869-882.
- selfcheck.go: errSourceNotOwner ~43-45; checkInputSource ~175-191.
- install_internal_test.go: wrap corpus consts ~43-55; TestWrapSourcesPairPreserved ~60; TestWrapSourcesThirdSourcePassthrough ~85; TestWrapSourcesRefusalTable ~100; TestWrapSourcesAlreadyOwnedUpgrade ~134.
- install_test.go: sources consts ~20-27; TestInstall_Sequence ~243-276; assertSequenceState ~284-322 (report pin at ~317-321); TestInstall_RefusalBeforeAnyWrite ~1094-1121 (plants sourcesUSFR).
- selfcheck_test.go: stubs ~59-101; assertSelfcheckLines ~124-139; wrappedSourcesStub ~400; input-source tests ~408-460.
- test/e2e/case_install.go: oracle consts ~127-133; carriesOraclePair ~137-153; wrapForCheck ~182-206; installSourcesWrapped ~215-237; installSelfcheck ~241-260 (greps "ok input-source" — stays matching).
- test/e2e/case_switch.go: requireTwoSourceDesktop ~1216-1234; sole caller pinRestartEngineRU in case_resilience.go ~140.
- mise tasks: build / vet / lint / test / tidy-diff / ci — run from repo root; no Makefile (D-09). go.mod must not change (no new deps; activate.ParseSourceTuples is already imported by install.go and case_install.go).
- untouchable by scope: internal/session, internal/indicator, internal/ctlsvc, SPEC.md, and the uninstall restore path (saveState/restoreSwitchBinding/restoreSources/derivedEngine logic and fallbackSources stay verbatim).
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: installer wraps at least one wrappable xkb source; foreign residue refused; report names the engines</name>
  <files>internal/install/install.go, internal/install/install_internal_test.go, internal/install/install_test.go</files>
  <action>
RED first: update/extend the unit corpus, run `go test ./internal/install/`, watch the new expectations fail, commit test(260930-toa).

install_internal_test.go:
- Add consts: wrapSingleRU = `[('xkb', 'ru')]`, wrapSingleUSWrapped = `[('ibus', 'goswitch-en')]`, wrapSingleRUWrapped = `[('ibus', 'goswitch-ru')]`, wrapSingleFR = `[('xkb', 'fr')]`.
- New TestWrapSourcesSingleSource: wrapSources(wrapSingleUS) → wrapSingleUSWrapped; wrapSources(wrapSingleRU) → wrapSingleRUWrapped; nil errors.
- TestWrapSourcesRefusalTable rows become: us+fr (wrapUSFR) → errMixedSources; single fr (wrapSingleFR) → errUnsupportedPair; unsupported kind (wrapForeignKind) → errUnsupportedPair; half-wrapped (wrapHalfWrapped) → errMixedSources; already-owned single (wrapOwnedSingle) → errUnsupportedPair (nothing wrappable — today's behavior, now pinned). Drop the old single-us refusal row.
- TestWrapSourcesThirdSourcePassthrough FLIPS into TestWrapSourcesForeignResidueRefused: wrapTriple ([us, ru, fr]) → errMixedSources; delete the now-unused wrapTripleWrapped const.

install_test.go:
- assertSequenceState: the report pin becomes "sources: wrapped (goswitch-en, goswitch-ru)" for the pair desktop.
- New TestInstall_SequenceSingleSource mirroring TestInstall_Sequence but with the fake desktop answering a single `[('xkb', 'us')]`: exactly one `gsettings set sources` carrying `[('ibus', 'goswitch-en')]`, the `ibus engine goswitch-en` activation kept, switch bindings still untouched, and the report naming "sources: wrapped (goswitch-en)".
- TestInstall_RefusalBeforeAnyWrite keeps its planted us+fr desktop and ALL its assertions (refusal before any mutation, no state file) — only its comment wording moves from "without the us/ru pair" to the foreign-residue reason.

GREEN (install.go), minimal until the corpus holds, then commit feat(260930-toa):
- Rename carriesXKBPair → carriesWrappableXKB: true when the parsed list carries AT LEAST ONE xkb 'us' or 'ru' entry; update resolveWrapInput's call and both functions' comments (the gate admits a single wrappable source).
- Replace the pairTuples const with minWrapped = 1.
- wrapSources gate order: (1) input mixed goswitch+foreign → errMixedSources (unchanged); (2) unsupportedKind from renderWrapped → errUnsupportedPair (unchanged); (3) wrapped < minWrapped → errUnsupportedPair reworded: no xkb 'us'/'ru' layout goswitch wraps found — goswitch wraps a single 'us' or 'ru' source, or the pair — the layout choice stays yours — fix: set a layout to xkb 'us' or 'ru', then re-run goswitchctl install; (4) foreign residue → errMixedSources reworded: goswitch wraps only a list it can own entirely — a foreign source beside goswitch engines is one the daemon sees no keys through (D-53) — fix: keep only xkb 'us'/'ru' (and goswitch) sources, then re-run goswitchctl install; (5) render.
- Foreign residue, computed exactly: after renderWrapped, transited = len(parts) - wrapped - owned. Derivation: parts holds goswitch tuples (owned), wrapped xkb entries (wrapped) and verbatim transits — unsupported kinds were already refused at (2), so every remaining non-goswitch non-wrapped part IS a transit. Order matters: check (3) BEFORE (4) so a pure-foreign desktop ([xkb 'fr']) stays the nothing-wrappable refusal, while us+fr and us+ru+fr become the residue refusal. This residue refusal is DELIBERATE, not scope creep: the existing corpus pins us+fr as a refusal (TestInstall_RefusalBeforeAnyWrite), and selfcheck's new rule (Task 2) makes a goswitch-beside-foreign list red — install must never write a list its own audit rejects (D-53).
- Reword errUnsupportedPair and errSavedOriginalUnusable fix hints from "set both layouts" to the single-source reality ("set a layout to xkb 'us' or 'ru'"); keep every error static (err113), naming reason + fix, never the raw user line.
- resolveWrapInput's default-branch comment: the verdict is the wrap refusal, never a forced pair.
- Report honesty: add an unexported field wrappedEngines []string to Installer; takeoverSources records the engine names of the COMPUTED wrapped value (parse it with activate.ParseSourceTuples, collect isGoswitchTuple IDs in order) BEFORE the skip-if-equal early return so both the write and the skip paths report truthfully; report() renders "sources: wrapped (" + strings.Join(i.wrappedEngines, ", ") + ")".
- Comment hygiene: Install's doc comment ("computed two-source sources wrap") and saveState's gate comment ("wrappable pair") move to the ≥1 wording.
- DO NOT touch: saveState/restoreSwitchBinding/restoreSources logic, fallbackSources, Uninstall, renderWrapped's body (its verbatim/unsupported branches stay — the refusal lives in wrapSources), waitRegistration, activateEngine.
  </action>
  <verify>
    <automated>go test -race -count=1 ./internal/install/ && mise run ci</automated>
  </verify>
  <done>Single-us and single-ru desktops wrap to their one engine; us+fr and us+ru+fr refuse with the residue verdict; pure-foreign and unsupported kinds refuse as today; the atomic refusal test stays green unchanged; the sequence corpus pins the single-source set and the named report line; mise gates green; RED and GREEN commits exist.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: selfcheck input-source green on ≥1 goswitch engine, verdict names the engines, mixed stays red</name>
  <files>internal/install/selfcheck.go, internal/install/selfcheck_test.go</files>
  <action>
RED first: failing expectations in selfcheck_test.go, run, commit test(260930-toa).

- assertSelfcheckLines: the last green line becomes "ok input-source goswitch-en, goswitch-ru" (the wrappedSources stub desktop).
- TestSelfcheck_InputSourceTwoEngines additionally pins the verdict content ("goswitch-en, goswitch-ru" on the ok line).
- TestSelfcheck_InputSourceSingleRejected FLIPS into TestSelfcheck_InputSourceSingleGreen: the goswitchSources stub (only goswitch-en) is GREEN, the verdict line carries goswitch-en and NOT goswitch-ru.
- TestSelfcheck_InputSourceForeignRejected stays red with the hintInstall pin; its content pins move from the two-engine wording to "no goswitch engine".
- New TestSelfcheck_InputSourceMixedRejected: a stub answering `[('ibus', 'goswitch-en'), ('xkb', 'us')]` is RED with the mixed-source verdict (the D-53 rationale) and the install hint.

GREEN (selfcheck.go), then commit feat(260930-toa):
- checkInputSource: read sources (unchanged call); parse with activate.ParseSourceTuples (add the import — the same grammar the installer uses; a parse failure wraps errSourceNotOwner with the reason, D-20 allows parse reasons); split with countOwnedForeign; owned == 0 → errSourceNotOwner, reworded: the input sources carry no goswitch engine (goswitch-en or goswitch-ru) — fix: goswitchctl install; foreign > 0 → errMixedSources (the Task-1 reworded static error — the D-53 rationale and fix hint come with it); otherwise collect the goswitch engine IDs in list order and return "input-source " + strings.Join(names, ", ").
- Keep the step name "input-source" and the fail-fast render unchanged; the verdict carries engine names only — config literals, never the raw gsettings line (D-20/D-21).
- DO NOT touch: the other five steps, checkEngineRegistered (both engines are always REGISTERED by the component regardless of the sources list), the repair path, probeCtlStatus.
  </action>
  <verify>
    <automated>go test -race -count=1 ./internal/install/ && mise run ci</automated>
  </verify>
  <done>Two-engine, single-engine (green, named), zero-goswitch (red + install hint) and mixed (red, D-53 rationale) input-source desktops all pinned; the full six-line green transcript updated; -race clean; RED and GREEN commits exist.</done>
</task>

<task type="auto">
  <name>Task 3: e2e corpus — install-cycle oracle wraps ≥1 source; ibus-restart precondition count-agnostic</name>
  <files>test/e2e/case_install.go, test/e2e/case_switch.go</files>
  <action>
The production code is already green from Tasks 1-2; this task aligns the oracle corpus that deliberately restates (never imports) the wrap rule. One commit test(260930-toa) after the gates pass.

case_install.go:
- Rename carriesOraclePair → carriesOracleSource: true when the parsed line carries AT LEAST ONE wrappable xkb us/ru entry (mirror of the new carriesWrappableXKB); update both call sites (installSourcesWrapped ~217, runUninstallAndVerify ~361) and the comment.
- wrapForCheck: the success condition becomes at-least-one — track hasUS/hasRU as today but refuse only when NEITHER is present (error text: "no wrappable xkb us/ru source"); the mapping and verbatim transit stay as-is; comment updated to the ≥1 restatement.
- installSourcesWrapped/runUninstallAndVerify comments: "the desk's pair" → the desk's wrappable source(s); the snapshot-else-saved-original resolution is unchanged and now also serves single-source snapshots.

case_switch.go (requireTwoSourceDesktop literally pins both engines and would fail-fast the resilience case on the owner's post-migration desktop):
- Rename to requireGoswitchDesktop: green when the live sources carry at least one goswitch engine (substring for either engine literal, matching the current style); failure message: sources carry no goswitch engine — install first. Update pinRestartEngineRU's call and both functions' comments: the ibus-restart extension presupposes a goswitch-owned desktop of ANY count — with one source the flip mechanism still works (SetGlobalEngine targets any registered engine; the component registers both engines regardless of the sources list; v1.0.0 single-source precedent).

Verify-and-leave (do not edit unless a literal two-source pin shows up): the runtime flip oracles — busFlipRound's journal pair + readback, external-flip-sync, flip-after-done, two-source-flip rows, matrix v3 (including the frozen word-mixed/phrase-mixed rows, WINDOWS #12) — are sources-count-agnostic. Confirm by scanning the e2e tree for remaining both-engines sources pins; record the scan result in the summary. installSelfcheck's Contains("ok input-source") keeps matching the named verdict — leave it.
  </action>
  <verify>
    <automated>mise run ci && ! grep -rn "func (s \*stand) requireTwoSourceDesktop" test/e2e/</automated>
  </verify>
  <done>The oracle computes the ≥1 wrapper for single-source snapshots; the ibus-restart precondition accepts one goswitch source; flip oracles untouched and count-agnostic (scan recorded in the summary); the whole e2e package compiles and its unit tests pass under mise run ci.</done>
</task>

<task type="auto">
  <name>Task 4: README user model + ADR-006 amendment + deferred notes + live end-state</name>
  <files>README.md, docs/adr/ADR-006-two-engine-revision.md</files>
  <action>
Documentation task (no TDD); one commit docs(260930-toa); `mise run ci` still green before committing.

README.md — rewrite the user-model sections honestly and short:
1. The "Input-source handover" section still carries the cleared-era claim (install clears the native switch binding) — stale since quick 260930-nxd: rewrite it to the truth — install touches ONLY the sources key; the GNOME switch chords in the wm.keybindings schema are left untouched (snapshotted verbatim for the uninstall restore); switching is goswitch's own gestures (single right-Shift tap; Super+Space as a goswitch gesture per the defaults table).
2. The "Two engines, one indicator (ADR-006)" section → retitle (e.g. "Input sources and the mode indicator") and rewrite: install reads YOUR sources and wraps at least one xkb 'us'/'ru' entry into its goswitch engine — a single 'us' → goswitch-en, a single 'ru' → goswitch-ru, both → both engines; ONE goswitch source is the recommended shape: GNOME's native panel indicator only appears with 2+ sources, so with one it is hidden BY DESIGN — the mode indicator is goswitch's own tray icon (StatusNotifierItem), which follows the daemon's mode immediately; the daemon's flip changes typing at once; a foreign source beside goswitch engines cannot come from the installer (refused) and is flagged red by selfcheck — the daemon sees no keys through an xkb source — while the daemon itself never fights the desktop (follows the factual engine); ibus-restart semantics paragraph stays (shell gesture survives, daemon follows the factual engine).
3. Do NOT document tray-click flipping (260930-pf6 follow-up, not implemented).
4. Troubleshooting: the "source outside the goswitch pair" bullet — reword to a non-goswitch source taking over (drop the legacy single-source confusion; mention selfcheck's mixed verdict as the audit's answer); keep the remedy (switch back to a goswitch source, or restart the unit).

docs/adr/ADR-006-two-engine-revision.md — append, do not rewrite history:
1. Under ## Status add one pointer line: Amended 2026-09-30 — single-source model per owner decision (see the amendment section below).
2. Append "## Amendment 2026-09-30 — single-source model (решение владельца)" in the document's language and style: the owner's verbatim intent («в гномовском переключателе остается один источник»); WHY the two-source rationale is obsolete — the native GNOME indicator is dead on the owner's desktop (proven exhaustively) and was replaced by goswitch's own SNI tray icon (live since quick 260930-pf6); WHAT changes — install wraps ≥1 wrappable xkb source (single 'us' → goswitch-en, single 'ru' → goswitch-ru, pair → both), selfcheck is green on ≥1 goswitch engine with zero foreign entries and names the engines found, the refusal table loses only the pair requirement (mixed/unsupported kinds and the new foreign-residue case keep named refusals, D-53); WHAT IS UNAFFECTED — the daemon-side flip (SetGlobalEngine), the 05-04 sync listener, the tray icon, XKB safety-net (engines are registered by the component on the IBus bus, not by the sources list; SetGlobalEngine works for any registered engine — v1.0.0 ran [('ibus','goswitch-en')] with internal flips); evidence pointers — .planning/quick/260930-pf6-feat-tray-indicator-statusnotifieritem-a/260930-pf6-VERIFICATION.md and the owner's 2026-09-30 evening desktop probes; SPEC.md is NOT edited here — the specification change routes through the owner's spec-delta process in the v1.1 pass.

Deferred notes (record in the SUMMARY's deferred section when writing it): (a) SPEC.md still documents the two-source takeover — owner routes via spec-delta in v1.1; (b) known interaction, owner-accepted: after the desktop migrates to one source, a later `goswitchctl install` re-wraps from the SAVED original in the state file (the first backup is sacred, Pitfall 7), so a reinstall re-imposes the saved pair — v1.1 spec-delta candidate.
  </action>
  <verify>
    <automated>mise run ci && ! grep -n "clears GNOME" README.md && ! grep -n "Both engines sit" README.md && grep -qn "Amendment 2026-09-30" docs/adr/ADR-006-two-engine-revision.md</automated>
    <human-check>Live end-state, executed by the orchestrator on the owner's desktop AFTER merge with freshly built binaries: set the sources to one goswitch engine (`gsettings set org.gnome.desktop.input-sources sources "[('ibus', 'goswitch-en')]"`), `systemctl --user restart goswitchd`, then: `goswitchctl selfcheck` prints all six ok lines with "ok input-source goswitch-en"; GNOME's native input indicator is GONE from the top bar (one source); the goswitch tray icon is present and reads en; Super+Space flips typing (tray icon follows to ru); `goswitchctl status` reports mode= matching the factual engine.</human-check>
  </verify>
  <done>README documents the single-source model with no stale cleared-era or both-engines claims; ADR-006 carries the amendment with evidence pointers and an untouched history; deferred notes recorded; live end-state steps listed for the orchestrator.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| user-controlled gsettings sources line → installer / selfcheck | The raw desktop string crosses into parsing, the `gsettings set` argv and the audit verdicts — the GVariant-injection class (T-05-02-01) and the D-20/D-21 no-user-content rule |
| state file → resolveWrapInput upgrade path | The saved original drives the wrap on an already-goswitch desktop (unchanged in this task, re-validated below) |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-TOA-01 | Tampering | wrapSources → gsettings argv | medium | mitigate | Unchanged closed-enum render: the argv is built ONLY from parsed tuples over the kindXKB/kindIBus/layoutUS/layoutRU literals; the relaxation widens the ACCEPTED inputs, not the render path; the raw user line never reaches argv (ASVS V5 / T-05-02-01) — pinned by the existing refusal-table no-raw-line assertion |
| T-TOA-02 | Information Disclosure | checkInputSource verdict line | low | mitigate | The new named-engines verdict carries only the config literals goswitch-en/goswitch-ru — never the raw sources line or user content (D-20/D-21); tests pin the exact verdict shape |
| T-TOA-03 | Tampering | forged install-state.json driving the upgrade wrap | low | accept | Existing ASVS V5 discipline unchanged: savedSources shape-checks the value before use and the uninstall restore falls back to the safe xkb default; this plan does not touch the state-file path |
| T-TOA-04 | Denial of Service | pathological sources strings (deep/oversized lists) | low | accept | Parse-refuse on shape errors, no loops, bounded list sizes from a gsettings reply; behavior identical to today for refused shapes |
</threat_model>

<verification>
- `mise run ci` green after every task (build + vet + golangci-lint strict + go test -race -count=1 + tidy-diff); no go.mod change.
- internal/install corpus: single-source wraps, residue/mixed/unsupported refusals, atomic refusal before any mutation, single-source install sequence with named report, selfcheck green/red matrix for the input-source step — all pinned by tests.
- test/e2e compiles and its unit tests pass; the oracle restates the ≥1 rule; the ibus-restart precondition is count-agnostic; flip oracles verified untouched (scan recorded in the summary).
- README and ADR-006 match the shipped behavior; grep gates prove the stale wording is gone and the amendment exists.
- Live end-state (human-check in Task 4) executed by the orchestrator post-merge: single source on the owner's desktop, selfcheck green, native indicator hidden, tray icon live, Super+Space flips.
</verification>

<success_criteria>
A desktop carrying one xkb 'us' (or 'ru') source installs into exactly [('ibus','goswitch-en')] (or goswitch-ru); the pair still installs as before; goswitchctl selfcheck is green on any goswitch-only sources list of count ≥1 and names the engines it found; foreign-beside-goswitch lists are refused by the installer and flagged red by the audit (D-53); the e2e corpus and the ibus-restart precondition accept the single-source desktop; README and ADR-006 document the amended model with the tray icon as the indicator; SPEC.md, internal/session, internal/indicator, internal/ctlsvc and the uninstall restore path are untouched.
</success_criteria>

<output>
Create `.planning/quick/260930-toa-feat-single-source-input-model-selfcheck/260930-toa-SUMMARY.md` when done (include the deferred notes from Task 4 and the e2e scan result from Task 3)
</output>
