---
phase: 260930-nxd-fix-switchtimeout-40-150ms-actor-go-inst
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/session/actor.go
  - internal/session/actor_test.go
  - internal/install/install.go
  - internal/install/install_test.go
autonomous: true
must_haves:
  truths:
    - The flip's SetGlobalEngine call carries a hard deadline (150 ms) that sits ABOVE the measured live round trip (~41-45 ms) — every healthy flip completes inside the budget and no flip logs a spurious deadline-exceeded WARN; a wedged ibus-daemon still costs at most the deadline, never an unbounded keystroke stall.
    - goswitchctl install writes nothing to org.gnome.desktop.wm.keybindings switch-input-source / switch-input-source-backward — both GNOME switch chords stay live and cycle between the two goswitch engines (ADR-006 two-source; the daemon follows external flips via the 05-04 sync listener).
    - Install-state still snapshots both pre-install bindings verbatim; uninstall still restores them (restoreSwitchBinding and the fallback constants untouched).
  artifacts:
    - internal/session/actor.go (switchTimeout 150 ms, truthful doc comments at ~51-55 and ~1346-1357)
    - internal/session/actor_test.go (flipBudget mirror 150 ms + floor assertion above the measured live RTT)
    - internal/install/install.go (no clearSwitchBinding, no clearedSwitchBindings, ADR-006 rationale comments, updated report line)
    - internal/install/install_test.go (corpus asserts binding preservation, not clearing)
  key_links:
    - flipBudget (test mirror) ↔ switchTimeout (production) — the seam corpus pins the deadline from the outside
    - saveState reads both switch bindings ↔ restoreSwitchBinding restores them verbatim — the save/restore round trip survives the clearing removal
    - report() verdict line ↔ assertSequenceState report assertion — wording stays parsed-consistent
---

<objective>
Two live-evidenced fixes from the owner's desktop (2026-09-30 root-cause session):

1. **internal/session/actor.go** — `switchTimeout` 40 ms → 150 ms. Every real SetGlobalEngine round trip measures ~41-45 ms (05-05 live proofs + 2026-09-30 journal), so the 40 ms deadline fires FIRST: every flip logs `switch_engine: context deadline exceeded` + an `engine switch failed` WARN while the switch actually lands. The WARN lies on every flip. Raise the deadline so it is a real wedge guard on the failure path, not a lie; do NOT touch the <50 ms SPEC reaction budget (the deadline only bounds the failure path).

2. **internal/install/install.go** — install stops clearing GNOME `switch-input-source` / `switch-input-source-backward`. The clearing rationale ("owner decision 3": with a single goswitch source the chord only churned/disables the engine context) is OBSOLETE under ADR-006 two-source: sources are now a pair of goswitch engines, GNOME's chord cycles between them as a first-class VISIBLE switch, and the daemon FOLLOWS external flips via the 05-04 sync listener (GlobalEngineChanged + FocusIn engine-name feed, live proof "external-flip-sync", commit 1caf5b4). Clearing now disables the user's own visible switch. Install must leave both bindings untouched (saveState still snapshots them verbatim for uninstall restore — unchanged), and the uninstall restore path stays as-is.

Purpose: the flip WARN currently reports failure on every successful flip (operator-visible lie), and every installed desktop loses its GNOME layout-switch chord for no reason under the two-source model.

Output: two atomic conventional commits, one per task, each ending green on the mise gates.

**Scope guard (binding):** two files + their tests only — internal/session/actor.go (+actor_test.go), internal/install/install.go (+install_test.go). No config schema changes, no ADR file edits (ADR-006 stays Accepted as-is), no e2e changes, no README/docs edits (the README wording about no-longer-switching GNOME chords dates from the cleared era and will need a separate doc pass — deliberately out of scope here per the scope guard).

**Process requirements (AGENTS.md, binding):** strict TDD — each task writes the failing test first, then minimal implementation, then refactor; every iteration ends green on `mise run build`, `mise run vet`, `mise run lint`, `mise run test` (equivalently `mise run ci`, run from repo root). Atomic conventional commits per task.
</objective>

<execution_context>
@/home/nil/DiskD/W/Djarvur/goswitch/.zcode/gsd-core/workflows/execute-plan.md
@/home/nil/DiskD/W/Djarvur/goswitch/.zcode/gsd-core/templates/summary.md
</execution_context>

<context>
@/home/nil/DiskD/W/Djarvur/goswitch/.planning/STATE.md
@/home/nil/DiskD/W/Djarvur/goswitch/internal/session/actor.go (lines 51-55: switchTimeout const + doc comment; lines 1335-1383: flipTo + the switcher-deadline comment block)
@/home/nil/DiskD/W/Djarvur/goswitch/internal/session/actor_test.go (lines 3945-3990: seam-corpus header, errSwitchInjected, flipBudget, switchProbe; lines 4127-4190: TestActor_SwitcherDeadlineBounded)
@/home/nil/DiskD/W/Djarvur/goswitch/internal/install/install.go (lines 80-118: schema/key/fallback constants + comments; lines 352-380: Install tail — takeover → clearSwitchBinding → activateEngine; lines 852-896: clearSwitchBinding + report; lines 950-985: restoreSwitchBinding)
@/home/nil/DiskD/W/Djarvur/goswitch/internal/install/install_test.go (lines 236-314: TestInstall_Sequence + assertSequenceState; lines 495-568: TestInstall_SecondInstallKeepsOriginalBackup + assertWrapperFromSavedOriginal; lines 574-625: TestUninstall_FullRollback)
</context>

<precondition>Repo at HEAD with the full gate green before any change: `mise run ci` (build + vet + lint + test -race) passes from the repo root.</precondition>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: switchTimeout 40→150 ms — the deadline becomes a real wedge guard above the measured live RTT</name>
  <files>internal/session/actor.go, internal/session/actor_test.go</files>
  <behavior>
    - TestActor_SwitcherDeadlineBounded (actor_test.go:4133) keeps pinning the MECHANISM: the switcher seam receives a context with a hard deadline; a seam stuck until the deadline returns ctx.Err(); the actor keeps serving (StatusSnapshot through the stalled flip completes); the internal mode stands switched.
    - NEW floor contract in the same test: the deadline budget measured at seam entry must be STRICTLY GREATER than the measured live SetGlobalEngine round trip — add a test constant liveSwitchRTTMax of 45 * time.Millisecond (live evidence: every real round trip is ~41-45 ms, 05-05 proofs + 2026-09-30 journal; the old 40 ms deadline fired before completion on every flip) and assert the entry budget is in (liveSwitchRTTMax, flipBudget]. With the production constant still at 40 ms this assertion FAILS — that is the RED.
    - flipBudget (actor_test.go:3959) moves to 150 * time.Millisecond with its doc comment reworded: it mirrors switchTimeout and sits above the measured live round trip, so the deadline guards the wedge path without firing on healthy flips.
    - The doc comments at actor.go:51-55 and inside the flipTo comment block (~1346-1357) state the NEW truth: the deadline is the failure-path wedge guard (Pitfall 4 / T-05-03-01), not a slice of the <50 ms SPEC reaction budget — the measured success RTT (~41-45 ms) must complete inside it; 150 ms gives headroom above the measured maximum while a wedged bus still costs at most the deadline.
  </behavior>
  <action>RED first (test-only commit): in internal/session/actor_test.go set the flipBudget constant at line 3959 to 150 * time.Millisecond, reword its doc comment per <behavior>, reword the TestActor_SwitcherDeadlineBounded doc comment at lines 4127-4132 (it currently says the deadline is at most 40 ms), and extend the entry-budget assertion with the liveSwitchRTTMax floor constant plus its live-evidence comment. Run `go test -race -count=1 ./internal/session/` — TestActor_SwitcherDeadlineBounded must FAIL on the floor assertion (40 ms is not above the 45 ms measured maximum). Commit the RED. GREEN (minimal implementation): in internal/session/actor.go line 55 change the switchTimeout constant to 150 * time.Millisecond; rewrite its doc comment (lines 51-54) so no sentence claims the deadline is a slice of the <50 ms reaction budget — it bounds the failure path only; touch the flipTo comment block (~1346-1357) only where needed to stay truthful about the same fact (the "wedged ibus-daemon costs at most switchTimeout" sentence stays correct as written). Do NOT touch any SPEC reaction-budget text elsewhere, do NOT change flipTo's logic — the context.WithTimeout call at line 1370 already consumes the constant. Run `go test -race -count=1 ./internal/session/` — green, including TestActor_FlipRoutesThroughSwitcher and the rest of the seam corpus (the fake-bus corpus answers fast; only the stall test gets ~110 ms slower — acceptable). REFACTOR if any comment now contradicts another; then end the iteration green on `mise run ci` from the repo root. Commit: conventional fix scope session, message stating the deadline now sits above the measured live SetGlobalEngine round trip.</action>
  <verify>
    <automated>cd /home/nil/DiskD/W/Djarvur/goswitch && go test -race -count=1 -run 'TestActor_SwitcherDeadlineBounded|TestActor_FlipRoutesThroughSwitcher' ./internal/session/ && mise run ci</automated>
  </verify>
  <done>switchTimeout is 150 ms with truthful comments; the seam corpus pins the deadline both above the measured 45 ms live-RTT maximum and at most the 150 ms mirror; the full gate (build, vet, lint, test -race) is green; RED and GREEN commits exist.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: install leaves GNOME switch-input-source bindings untouched (ADR-006 two-source makes clearing obsolete)</name>
  <files>internal/install/install.go, internal/install/install_test.go</files>
  <behavior>
    - TestInstall_Sequence (install_test.go:236) asserts install records NO `gsettings set` call against gsettingsKeybindingsSchema at all — neither switch-input-source nor switch-input-source-backward is written during install; the expected call sequence otherwise stays identical (takeover writes only the desktop.input-sources sources key; the next recorded set-capable call is `ibus engine goswitch-en` activation).
    - assertSequenceState (install_test.go:276) keeps asserting the state file carries both pre-install bindings VERBATIM (ownerSwitchBindings / ownerSwitchBindingsBackward — saveState is unchanged) and flips its report assertion: the report names the preserved verdict whose exact line is "switch-input-source: left untouched (saved for uninstall)" instead of the cleared verdict.
    - TestUninstall_FullRollback (install_test.go:574) stays UNCHANGED and green: the saved switch bindings still go back before the state file dies, and the report still names both restored values — the restore path is untouched.
    - TestInstall_SecondInstallKeepsOriginalBackup (install_test.go:503) keeps its marker byte-intact; its stub returning the empty-array value for the switch-binding gets stays (now it proves a junk live value never overwrites the sacred backup, rather than simulating the post-handover desktop); its doc comment and the TestInstall_Sequence doc comment (lines 230-235) drop the owner-decision-3 clearing language.
  </behavior>
  <action>RED first (test-only commit): in internal/install/install_test.go remove the two expected `set ... []` sequence entries at lines 260-262 and add the absence assertion (no recorded gsettings set touches gsettingsKeybindingsSchema during install); flip the report assertion at lines 304-308 to the preserved wording pinned in <behavior>; reword the doc comments named above. Run `go test -race -count=1 ./internal/install/` — TestInstall_Sequence and TestInstall_SecondInstallKeepsOriginalBackup assertions FAIL against the current clearing code. Commit the RED. GREEN (minimal implementation) in internal/install/install.go: delete the clearSwitchBinding call from Install (lines 372-374 — takeover flows straight into activateEngine), delete the clearSwitchBinding function (lines 852-868) and the clearedSwitchBindings constant (line 111, now orphaned — strict lint flags it); rewrite the comment blocks at lines 84-90 and 96-105 to the ADR-006 two-source rationale: sources are a pair of goswitch engines, GNOME's switch chord cycles between them as a first-class visible switch, the daemon follows external flips via the 05-04 sync listener (GlobalEngineChanged + FocusIn engine-name feed, live proof external-flip-sync, commit 1caf5b4), so install leaves both bindings untouched while saveState still snapshots them verbatim for uninstall restore; update the report() line at 893 to the preserved wording pinned in <behavior>. Keep gsettingsKeybindingsSchema, both key constants, both fallback constants, restoreSwitchBinding, and the uninstall flow byte-identical. Run `go test -race -count=1 ./internal/install/` — green including TestUninstall_FullRollback. REFACTOR comments if any residue of the clearing rationale survives; end the iteration green on `mise run ci`. Commit: conventional fix scope install, message stating the GNOME switch chords stay live under ADR-006 two-source.</action>
  <verify>
    <automated>cd /home/nil/DiskD/W/Djarvur/goswitch && go test -race -count=1 -run 'TestInstall_|TestUninstall_' ./internal/install/ && mise run ci</automated>
  </verify>
  <done>Install writes nothing to the wm.keybindings schema and reports the preserved verdict; the state file still snapshots both bindings verbatim; uninstall restore is unchanged and green; clearedSwitchBindings and clearSwitchBinding are gone; the full gate is green; RED and GREEN commits exist.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| install → gsettings | install/uninstall write GVariant values to the user's dconf via subprocess; restore values come from the saved state file (untrusted if corrupted) |
| daemon → ibus-daemon | the flip's SetGlobalEngine round trip crosses the IBus socket; a wedged server is the failure mode the deadline bounds |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-Q-01 | Denial of Service | actor.go flipTo | low | mitigate | Task 1 keeps the hard-deadline mechanism (existing TestActor_SwitcherDeadlineBounded) and raises the budget to 150 ms — a wedged ibus-daemon still costs at most the deadline, never an unbounded keystroke stall |
| T-Q-02 | Tampering | install.go restoreSwitchBinding | low | accept | Unchanged by this plan: restore values remain shape-validated against the ASVS V5 fallbacks (existing tests green); leaving the bindings untouched at install strictly REDUCES the install-time write surface |
| T-Q-03 | Denial of Service | install.go Install | low | accept | Removing the clearing step removes one external mutation; the GNOME chord regains its distro-default function, so no new failure path is introduced |
</threat_model>

<verification>
- `go test -race -count=1 ./internal/session/` green — the seam corpus pins deadline ∈ (45 ms, 150 ms] at seam entry.
- `go test -race -count=1 ./internal/install/` green — install sequence contains no wm.keybindings write; uninstall rollback sequence unchanged.
- `mise run ci` green after each task (build, vet, golangci-lint strict, test -race) — AGENTS.md directive 2.
- grep confirms no surviving references: clearedSwitchBindings and clearSwitchBinding absent from internal/install/install.go; no `40 * time.Millisecond` remains in internal/session/.
- Two conventional commits, one per task, each carrying its RED (test) and GREEN (implementation) as separate commits or a paired sequence per the strict-TDD directive.
</verification>

<success_criteria>
- A flip on the live desktop completes inside the deadline: no `switch_engine: context deadline exceeded` WARN on healthy flips; a wedged bus still cannot stall the keystroke path beyond 150 ms.
- After `goswitchctl install`, `gsettings get org.gnome.desktop.wm.keybindings switch-input-source` still returns the user's pre-install binding; Super+Space (or the user's chord) cycles the two goswitch engines with the panel indicator following.
- `goswitchctl uninstall` restores both chords verbatim exactly as before this plan.
</success_criteria>

<output>
Create `.planning/quick/260930-nxd-fix-switchtimeout-40-150ms-actor-go-inst/260930-nxd-SUMMARY.md` when done
</output>
