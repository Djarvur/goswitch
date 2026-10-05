---
phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
reviewed: 2026-10-05T03:47:19Z
depth: standard
reviewer: gsd-code-reviewer
diff_base: 1a64aca..HEAD
files_reviewed: 23
findings:
  critical: 1
  warning: 3
  info: 6
  total: 10
status: issues_found
---

# Phase 7: Code Review Report

**Reviewed:** 2026-10-05T03:47:19Z
**Depth:** standard (per-file analysis, Go-specific checks, cross-file tracing of the WR-01 removal, the persist/watcher interaction, the menu state machine and the sound path)
**Commits:** 1a64aca..HEAD (branch gsd/phase-07-menyu-v2-i-chernyy-spisok-avtokorrektsii)
**Files Reviewed:** 23 (list below)
**Status:** issues_found — 1 Critical, 3 Warnings, 6 Info
**Gates re-run during review:** `go build ./...`, `go vet ./...`, `golangci-lint run` (0 issues), `go test -race -count=1` on internal/{session,config,sound,indicator,ctlsvc} + cmd/goswitchd — all green.

## Summary

Phase 7 ships menu v2 (DBusMenu state snapshot, EN/RU radio pair, two persisted toggles, settings launcher), the blocklist revision of the autocorrect gate (WR-01 equality removed), the persist writer (atomic Node round-trip), the sound engine (no-pipes subprocess sink), and the adopt+watch/startup-fold wiring. The engineering quality of the bulk remains high: the persist writer is atomic, comment-preserving and strict-complete; the menu's pinned IDs survive supervisor re-attach by re-exporting the same instance; the sound engine never spawns pipes, reaps every child and keeps its warn-once episodes; the WR-01 removal is mechanically sound (details below); the 06-review fixes (CR-01/CR-02/WR-01 mechanics) all survive; D-20/D-21 hold on every new log/status/error site.

The one Critical finding is not a code bug but a provable gap between the phase's own stated security invariant and what the code can deliver: SPEC §11 (phase-7 revision) and 07-CONTEXT claim the role gate makes autocorrect never fire in passwords, while the project's own ADR-007 live fact (GTK3 password fields report AT-SPI role 61 — indistinguishable from text) says the role gate does NOT catch GTK3 passwords. Phase 6's safety argument for that hole was the default-empty white-list; phase 7 deleted the white-list, made unknown-identity a pass, and put the ON switch one menu click away with an empty blocklist. The residue must be either documented as an owner-acknowledged risk or re-guarded — today the docs assert an invariant the implementation cannot keep.

## Critical Issues

### CR-01: The unknown⇒pass inversion makes the GTK3-password hole (role 61) default-reachable one menu click after enabling — SPEC §11's «пароли … — никогда» is not delivered for GTK3

**Files:** `internal/session/actor.go:2101-2117` (arm gate: unknown identity passes), `:2303-2304` (role gate: 61 allowed), `internal/config/writer.go:145-159` (the toggle's created document: `enabled: true` + empty blocklist), `docs/SPEC.md:219`, `docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md:132-137` vs `:224`

**Issue:** the role gate itself is untouched and correct (numeric 61/79/94 allowed, 40/60 and everything else refused, errors fail closed, warn-once). The finding is the conjunction around it:

1. ADR-007's phase-6 Consequences record the live fact (lines 132-137): a GTK3 password field (`zenity --password`, `GtkEntry visibility=off`) reports role **61 "text box"** — «Ролевый гейт GTK3-пароли НЕ ловит; безопасность держится на конъюнкции D-53 в целом: **white-list приложений (default пуст = выключено везде)** + роль + …». That safety argument is exactly what phase 7 deleted: `apps_blocklist` empty forbids nothing (`config.go:363-370`, empty is the validated default), an unknown identity passes both gates (`actor.go:2101-2117`, `:2357-2367`), and the menu toggle (07-05/07-06) turns the layer ON — creating `enabled: true` with an **empty** blocklist (`writer.go:145-159`) — in one click, no YAML hand-editing required as in phase 6.
2. The phase's own locked rule (07-CONTEXT «Locked semantics», SPEC §11 line 219: «пропускает (текстовый ввод; **пароли и терминалы — никогда**)») therefore does not hold on GTK3: enabled + non-blocklisted GTK3 app + password field (role 61) + surrounding-text caps + a ≥4-letter confidently wrong-layout "word" (i.e. a password) ⇒ the confirm fires and the ladder deletes/commits inside a password field. The e2e negative is GTK4 (role 40) and stays green — the GTK3 class is untested by design of the fixture.
3. The ADR-007 amendment (line 224) asserts «полевой role-гейт остаётся фильтром безопасности» — directly contradicted by the same document's live-fact paragraph three sections earlier. The amendment voided the white-list leg of the safety argument without re-deriving it.

**Fix:** owner decision, one of:
- (a) Honest residue: amend ADR-007 + SPEC §11 + README with the dated caveat «GTK3-пароль (роль 61) проходит гейт; blocklist-ревизия сделала этот класс достижимым при включённом слое — риск принят владельцем», mirroring the WINDOWS-#12 honesty form. Cheapest, matches the owner's variant-1 choice, but must be explicit — the current texts claim the opposite of the evidence.
- (b) Code mitigation: refuse to fire when the observed identity is UNKNOWN **and** the role is the ambiguous 61 (the one role value empirically shared by text and GTK3-password objects) — a one-line tightening of `acConfirmRefusals`/the confirm role switch that keeps the locked «unknown passes» scenario for roles 79/94 and the known-identity pass, and closes only the genuinely ambiguous cell.
- (c) Menu-toggle mitigation: when the autocorrect toggle CREATES the document, seed `apps_blocklist` with the password-class hint from docs/CONFIG.md instead of `[]`.

## Warnings

### WR-01: Unsynchronized `menu` local in the OnConn wiring — data race between the assignment and the toggle push closures

**File:** `cmd/goswitchd/main.go:141,160-161,173-174` (write at `menu = item.Menu()`), `:220-224,231-235` (push closures read `*menu`)

**Issue:** `newMenuToggles(actor, &menu, …)` captures `&menu` before `indicator.Attach`; the push closures read `*menu` on godbus dispatch goroutines. `Attach` exports the menu object (making `Event` dispatch possible) BEFORE `menu = item.Menu()` executes — a click arriving in that window races the plain pointer write (a Go data race, UB; `-race` cannot see it because the corpus drives the menu directly). The `*menu != nil` guard shows the early-click case was considered but not synchronized. Practically hard to hit (clicks are user-paced), zero cost to fix.

**Fix:**
```go
var menuSlot atomic.Pointer[indicator.Menu]
// push: if m := menuSlot.Load(); m != nil { m.SetSoundEnabled(on) }
// after Attach: menuSlot.Store(item.Menu())
```
or capture the `*indicator.Item` and read `item.Menu()` (immutable field, safe after construction).

### WR-02: `sound.autocorrect_event` is frozen at startup while the docs sell it as a hot-reloadable key

**Files:** `cmd/goswitchd/main.go:187` (event wired once from the startup document), `docs/CONFIG.md:65` (documented as a normal key), `docs/CONFIG.md:172-175` (blanket hot-reload contract), `07-CONTEXT.md:46` («Hot reload, strict decode — автоматически»)

**Issue:** only `sound.enabled` folds live (applySnapshot); the Player's `acEvent` is pinned at `sound.New(...)` for the daemon's lifetime (documented in 07-08-SUMMARY as a deliberate interface decision — but the user-facing contract was never updated). The «Настройки…» flow exists precisely so the user edits config.yaml in an editor: changing `autocorrect_event` there silently does nothing until a daemon restart, with no WARN — the D-32 contract («a reload never lies») is visually broken for this key.

**Fix:** either add `SetAutocorrectEvent(string)` to the sink and push `snap.Sound.EffectiveAutocorrectEvent()` from `pushMenuSync` on change (small: the Player guards `acEvent` with its existing mutex), or document the freeze explicitly in the CONFIG.md row («applied at daemon start; changes require a restart»).

### WR-03: `configToggle.flip` has no empty-path guard — the no-HOME environment writes `./config.yaml` into the daemon's CWD and permanently desyncs the menu from the status

**Files:** `cmd/goswitchd/main.go:334-348` (`flip`), `internal/config/writer.go:326-350` (`writeAtomically` with `path == ""` → `filepath.Dir("") == "."`)

**Issue:** when `config.DefaultPath()` fails (no HOME/XDG — the degenerate env `loadConfig` explicitly supports, `main.go:438-440`), `cfgPath` is `""` and the toggle's `write` becomes `SetSoundEnabled("", on)`: `os.ReadFile("")` → ENOENT → the ensure branch → `CreateTemp(".", ".config.yaml.*.tmp")` → rename to `./config.yaml` **in the daemon's working directory**, a file nothing will ever watch or read. The menu push then shows the new value while no fold can ever run (no watcher) — `StatusSnapshot.SoundEnabled` and the menu stay permanently divergent, and the next click computes from the stale status. The reload-nil leg is guarded (`main.go:342-346`); the path leg is not.

**Fix:**
```go
func (t *configToggle) flip() {
    if t.path == "" {
        slog.Warn("config toggle unavailable", "component", "tray indicator", "reason", "no config path")
        return
    }
    ...
}
```

## Info

### IN-01: `acPayload.app` is dead data after the WR-01 equality removal

**File:** `internal/session/actor.go:398,2145` — written at arm, never read; the struct comment (lines 380-385) still narrates confirm-time consumption («app is the bridge-namespace identity the blocklist verdict was evaluated for»), which is now false — the confirm consults the live identity only.

**Fix:** drop the field and the `app: app` literal, or reduce the comment to «reserved/unused since the 07-04 inversion».

### IN-02: The fired tone plays before the pipeline can refuse — an audible false positive

**File:** `internal/session/actor.go:2314-2323` — `s.AutoCorrect()` fires before `startRangeCorrection`, whose `empty-buffer`/`convert-failed`/`no-engine` refusals then land with a tone already played (the detector ran at arm, but ConvertRuns can still refuse — e.g. letters missing from the tables). The 06-REVIEW IN-01 counter-semantics class, now user-audible.

**Fix:** move the tone into the fired branch of the pipeline's settled execution, or accept and document (frequency is low).

### IN-03: `setScalar` mutates the value node in place — a non-scalar `enabled` yields a franken-node instead of a replacement

**File:** `internal/config/writer.go:290-294` — for a hand-broken `enabled: [x]` (edited between start and click), the node keeps `Kind: SequenceNode` with its old `Content` while `Tag`/`Value` claim bool; the write produces a document whose toggle is unchanged and whose reload the strict decoder rejects (fail-safe but the toggle silently no-ops with only a WARN).

**Fix:** replace wholesale — `sectionNode.Content[i+1] = scalarBool(on)` — which is also simpler than the three-field mutation.

### IN-04: `Status.SoundEnabled` lies in the no-watcher environment

**File:** `internal/session/actor.go:651` (reads `a.soundEnabled`), `:1263-1264` (set only inside `pushMenuSync`, which only the fold calls) — with `cfgSrc == nil` (no HOME) `opts.SoundEnabled` is ON (defaults via `SetOptions`, `main.go:484`) and tones play, while the status token reads `sound_enabled=false`. Degenerate-env only; the same env as WR-03.

**Fix:** set `a.soundEnabled = o.SoundEnabled` in `SetOptions` next to the fold, mirroring the two surfaces.

### IN-05: The focus-move coverage claim of `acConfirmRefusals` has no direct test pin

**File:** `internal/session/actor.go:2343-2347` (comment: «a focus move inside the window surfaces as payload-stale (FocusOut hard-resets the buffer…)») — the mechanism is pinned for the interleaved-keystroke variant (`TestAutoConfirm_RevalidatesArmedPayload`) and FocusOut's reset is pinned separately (`TestBuffer_ResetByFocusOut`), but no test drives FocusOut **inside the role-RTT window** of an armed confirm.

**Fix:** one corpus cell: arm → block the role double → `HandleLifecycle(LifecycleFocusOut)` → release → assert `payload-stale`, fired=0.

### IN-06: Carry-over (06-REVIEW IN-04, unfixed): concurrent `exec.Cmd.Wait` in the fixture timeout path

**Files:** `test/e2e/case_autocorrect.go:311-324` vs `:222-233` — on the timeout branch `reapPasswordFixture` calls `Wait` while the goroutine's `Wait` is still blocked (`exec.Cmd.Wait` is not safe for concurrent callers). Test-only, already-failing path; the file was heavily rewritten this phase without touching it.

**Fix:** kill in the timeout branch, let the single goroutine `Wait` own the reap (`close(done)` after its Wait; reap = Kill only).

## Checked dimensions and verdicts

- **WR-01 removal correctness (question 1):** SOUND mechanically. The confirm's fired threshold re-validates buffer content (`actor.go:2350-2356`) — a focus loss hard-resets the buffer (`HandleLifecycle`, `actor.go:743`), so the focus-move case lands as `payload-stale`; the always-verify rule for boundary payloads survives (`actor.go:2602`, the 06 CR-02 fix), so even the a11y-vs-engine ordering race (observer learns the new app before the engine's FocusOut reaches the actor) ends fail-closed: the delete/commit always target the ARMING context's engine, and that field's fresh Require answer must still end with token+tail. The unknown⇒pass polarity is consistent between arm and confirm; the role gate is byte-untouched (61/79/94 vs the world) — the residue it leaves is CR-01 above. No half-removed leftovers: `no-apps`/`app-not-listed`/`app-unknown`/`app-changed` appear nowhere in code or corpus; the old WR-01 test is replaced by `TestAutoCorrect_ConfirmBlocklistRefuses`/`TestAutoCorrect_ConfirmUnknownPasses`; the 06 CR-01/CR-02 regression tests survive. Only `acPayload.app` dangles (IN-01).
- **Persist writer (question 2):** CLEAN — same-directory dot-prefixed temp (the watcher's base-name filter ignores it), `os.Rename` atomicity, 0600 (`writer.go:326-350`), comment-preserving Node round-trip from the document's root mapping (`:128-138`), ensure branch builds the full Defaults document (strict-decodable, thresholds present when enabled — `:166-184`), toggle-on-absent-file creates the complete document with the sibling at default (`:145-159`). Adopt+watch: `MkdirAll` 0700 before the directory watch (`main.go:408-415`), missing file = defaults, corrupt adopted file = loud start refusal (`loadConfig`, T-07-03-03), broken mid-run edit = rejected reload with last-good serving. No reload loops: the rename produces one Create for the watched name → one debounced reload; the toggle's synchronous reload + the echo re-apply identical values (idempotent fold).
- **Menu v2 (question 3):** CLEAN — pinned IDs are compile-time constants and the exported menu is the item's own instance (`indicator.go:277-280`), so supervisor self-heal re-exports without resetting the seeded snapshot; `revision` bumps only on item-set changes (`menu.go:270-274`), property dynamics ride `ItemsPropertiesUpdated` with the empty-literal `a(ias)` tail (the nil-panic guard) and correct per-key deltas (state for toggles, label for text rows, LayoutUpdated for visibility flips); `SetKeys`' presence/label change matrix is complete (empty⇄non-empty on either binding → revision; label-only → deltas); underscores doubled for every rendered key name (`doubleUnderscores`); xdg-open is `exec.Command` with nil std descriptors, `Start` + reaped `Wait` — never pipes, never blocks the click handler (`main.go:294-313`). No lock cycle: `Event`/`dispatchClick` hold no menu mutex while calling into the actor; the actor's pushes take menu.mu under actor.mu one-way only.
- **Sound engine (question 4):** CLEAN — no-pipes Start + goroutine-reaped Wait (no zombies), backend decided once and cached, warn-once per reason per episode with healthy-launch reopen, deterministic stdlib RIFF/PCM synthesis (byte-exact header, size-matched cache reuse, 0700/0600), zero new dependencies. The synchronous `Flip()` under the actor mutex is the documented 07-08 discipline (fire-and-forget: the sync cost is one fork, never a wait); the G-6-1 lock-free `AttachEngine` contract is untouched. Both tones sit behind the single `soundSinkFor` gate; same-target flips are silent; `syncMode` deliberately silent.
- **FoldAppliedConfig + startup fold (question 5):** CLEAN — `newActor` order is SetOptions → AttachConfig → FoldAppliedConfig, so the document truth (including sound and the menu pushes) is in force before the first event; the toggle composition is snapshot → write → push → syncReload → fold, the fold re-reading the just-stored snapshot (correct ordering); the actor and the watcher never lock in opposite order (actor.mu → watcher snapshot read only).
- **Privacy D-20/D-21 (question 6):** CLEAN — every new log/status/error site carries slugs, counts, roles, paths, binary names or transport errors only; the writer logs nothing; menu labels carry config literals and the version; the unit corpus re-asserts word non-leakage. The only content-bearing records remain the pre-existing `-debug` ones.
- **Tests:** the new corpus is deterministic (channel-gated role doubles, no sleeps in unit paths); e2e menu-v2 oracles are independent (file + status + journal) with polling budgets; matrix v4 repin keeps the frozen v3 regression path.

## Files Reviewed

- internal/session/actor.go
- internal/session/actor_test.go (confirm/fold/sound corpus; survey of the remainder)
- internal/config/config.go
- internal/config/writer.go
- internal/config/config_test.go, load_test.go, watch_test.go, writer_test.go (gates green; survey)
- internal/ctlsvc/ctlsvc.go
- internal/indicator/indicator.go
- internal/indicator/menu.go
- internal/indicator/menu_test.go (gates green; survey)
- internal/sound/sound.go
- cmd/goswitchd/main.go
- cmd/goswitchd/main_test.go (gates green; survey)
- test/e2e/case_autocorrect.go
- test/e2e/case_menu.go
- test/e2e/case_menu_test.go (gates green; survey)
- test/e2e/main.go
- test/e2e/matrix.go
- test/e2e/matrix_test.go
- docs/SPEC.md (§11 revision, claim check)
- docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md (amendment, claim check)
- docs/CONFIG.md (sound/blocklist rows, hot-reload section)
- .planning/phases/07-*/07-CONTEXT.md, 07-0[1-8]-SUMMARY.md (context; not reviewed as code)
- 06-REVIEW.md / 06-REVIEW-FIX.md (fix-survival baseline)

## Fix Log (code-review --fix, 2026-10-05)

Fixer run against this review — the Critical and the three Warnings fixed RED-first (each finding: a failing corpus commit, then the minimal GREEN), the six Info findings stay documented above with no action (contract scope). Work was committed on the isolated review-fix branch (`gsd-reviewfix/07-1895025`) and fast-forwarded onto `gsd/phase-07-menyu-v2-i-chernyy-spisok-avtokorrektsii`; all gates ran in the isolated worktree before the merge.

| Finding | Mechanism | RED | GREEN | Re-test verdict |
|---|---|---|---|---|
| CR-01 | `acConfirmRefusals(payload, role)`: the ambiguous role 61 (the GTK3 "text box" passwords report, ADR-007 live fact) with an UNKNOWN identity refuses fail-closed under the new closed slug `role-ambiguous`; known identity + 61 unchanged (the blocklist decides); 79/94 + unknown keep firing; role-unknown/role-timeout disciplines untouched | b708020 | 80c99f2 | `TestAutoConfirm_RoleTextAmbiguity` (refuse + fire pair) green; the three unknown-identity fire cells re-pinned to the unambiguous entry 79; session package fully green under `-race` |
| WR-01 | the `menu` local became an `atomic.Pointer[indicator.Menu]` slot: `menuSlot.Store(item.Menu())` right after Attach, the toggle pushes read via a `Load()` nil guard | 713d25a | a23a47c | `TestMenuTogglePushRacesAttachStore`: RED tripped the detector exactly at the push-closure read (main.go:221) vs the store write; GREEN clean over 500 store/push interleavings under `-race` |
| WR-02 | resolved HOT (the fold path exists): `SoundSink.SetAutocorrectEvent` added; `sound.Player` setter with the mutex-guarded `acEvent` read; `pushMenuSync` pushes the effective event on change; a late `SetSoundSink` self-syncs (the SetMenuSync install-push precedent). The CONFIG.md row and the §Hot reload blanket are now truthful AS WRITTEN — no doc edit required | d10671a | 2504bf6 | `TestNewActor_SoundEventHotReload` (fold push + late-install self-sync) and `TestPlayer_SetAutocorrectEvent` (argv pin) green |
| WR-03 | `configToggle.flip` empty-path guard first: refuse with one WARN `config toggle unavailable` before any write/push/reload — the no-HOME environment never writes `./config.yaml` into the CWD and the menu never desyncs from the status | c70d620 | d727ef4 | `TestToggleEmptyPathRefused`: RED showed write+push+reload on the empty path; GREEN refuses and WARNs |

**Gates at close:** `mise run ci` (build + vet + golangci-lint + `go test -race -count=1 ./...`) fully green in the worktree at 5df3583. The known flakes needed no re-run: `TestRun_OnConnHookCalledOnce` and `TestWatch_BrokenBlocklistPatternKeepsLastGood` both passed on the first try (re-verified individually afterwards). A lint-conformance pass (5df3583 — intrange/lll/unused on the new corpus) closes the set; all review-fix commits are conventional and atomic per finding (RED and GREEN separate).

**Residue honestly noted (not actionable in this run):**
- CR-01's doc leg: SPEC §11 (line 219) and ADR-007 (line 224) still narrate the role gate as the safety filter without the role-61 caveat. After the (b) tightening the «пароли — никогда» invariant holds for UNKNOWN identities; a KNOWN, non-blocklisted identity + role 61 (a GTK3 password in an a11y-observable app) still fires — the owner-sanctioned narrow exception. The dated SPEC/ADR/README caveat in the review's variant-(a) form remains open for the owner.
- The live e2e fire case (`runAutocorrectFires`) is analyzed safe under the tightening: the confirm's identity is KNOWN there by construction (the daemon's observer stores the focus pair before the role-witness wait, the same pair the role query reads). Live e2e was not re-run — it drives the owner's GNOME session and is deliberately outside `mise run ci`.
- IN-01…IN-06 left as documented above (scope: Critical + Warnings only).

---

_Reviewed: 2026-10-05T03:47:19Z_
_Reviewer: Claude (gsd-code-reviewer), go-ultimate strict discipline_
_Depth: standard_
