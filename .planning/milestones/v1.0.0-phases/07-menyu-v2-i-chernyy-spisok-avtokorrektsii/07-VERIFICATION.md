---
phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
verified: "2026-10-07T23:22:00Z"
status: passed
score: 5/5 must-haves verified
covered_files:
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-01-PLAN.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-02-PLAN.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-03-PLAN.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-04-PLAN.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-05-PLAN.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-06-PLAN.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-07-PLAN.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-08-PLAN.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-01-SUMMARY.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-02-SUMMARY.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-03-SUMMARY.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-04-SUMMARY.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-05-SUMMARY.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-06-SUMMARY.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-07-SUMMARY.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-08-SUMMARY.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-CONTEXT.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-REVIEW.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-RELEASE-READINESS.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-SECURITY.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-UAT.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-VALIDATION.md
  - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/deferred-items.md
  - .planning/REQUIREMENTS.md
  - docs/SPEC.md
  - docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md
  - docs/CONFIG.md
  - README.md
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/config/load_test.go
  - internal/config/watch_test.go
  - internal/config/writer.go
  - internal/config/writer_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - internal/indicator/menu.go
  - internal/indicator/menu_test.go
  - internal/indicator/indicator.go
  - internal/sound/sound.go
  - internal/sound/sound_test.go
  - internal/ctlsvc/ctlsvc.go
  - cmd/goswitchd/main.go
  - cmd/goswitchd/main_test.go
  - test/e2e/case_menu.go
  - test/e2e/case_menu_test.go
  - test/e2e/case_autocorrect.go
  - test/e2e/main.go
  - test/e2e/matrix.go
  - test/e2e/matrix_test.go
  - test/e2e/cases/matrix-v4.yaml
  - mise.toml
covered_digest: "v1:sha256:566481df94ef83d02f7d1b0ef90ee8c9b5609e870026119094b968cabc26b282"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 5/5
  gaps_closed: []
  gaps_remaining: []
  regressions: []
deferred:
  - truth: "word-mixed / phrase-mixed matrix rows (mixed text across a flip boundary)"
    addressed_in: "FULLY RESOLVED by quick 261008-00m (fix 4a0c887, ADR-004 amendment 0781a5f): the G-2-3 residual — own flip resetting the correction buffer mid-word — is fixed via the flipCredit synthetic-lifecycle credit; live matrix v1 on the fix tree 16/16 with word-mixed → «паиghbdtn» (orchestrator-recorded, 261008-00m-VERIFICATION.md); phases 2/5 verification ledgers absorbed the delta"
    evidence: "internal/session/actor.go flipCredit (L263, armed flipTo L2017, consumed HandleLifecycle L788-790); TestActor_OwnFlipMixedWordCorrectsWhole GREEN -race; live 16/16 recorded (261008-00m-VERIFICATION.md:97)"
  - truth: "Sound latency (~150-300 ms player subprocess spawn)"
    addressed_in: "RESOLVED by quick 261006-squ: persistent PulseAudio stream + XDG sound-theme resolver (the prescribed remedy — постоянный плеер + предзагрузка темы); residual owner by-ear acceptance gates live in the 261006-squ task's own verification ledger (human_needed there, 7/7 automatable must-haves verified)"
    evidence: ".planning/quick/261006-squ-sound-persistent-pulseaudio-stream-gnome/261006-squ-VERIFICATION.md; internal/sound/{pulse,theme,pcm}.go"
  - truth: "Silent apps (Electron/Chromium without accessibility mode: ZCode, Telegram snap) reliability tuning"
    addressed_in: "Owner-deferred backlog"
    evidence: "Remedies documented (toolkit-accessibility gsettings / --force-renderer-accessibility, README.md «Where autocorrect stays silent», README.en.md:228, CONFIG.md); owner confirmed autocorrect works in ZCode after the remedy"
  - truth: "Formal fresh-session ≤30-min D-48 matrix gate"
    addressed_in: "Nightly pipeline / owner (WINDOWS #13 phase-6 precedent)"
    evidence: "07-UAT test 4 note: performed ×2 on an aged session with honest recording (STATE.md:268); formal gate delegated"
  - truth: "Pre-existing flake TestWatch_BrokenBlocklistPatternKeepsLastGood (~1/15)"
    addressed_in: "deferred-items.md (phase 07 ledger, open) — reparse ordering structurally unchanged (watch.go:242-245 stores snapshot before clearing lastErr); re-run-on-occurrence remedy stands"
    evidence: "internal/config/watch.go reparse; deferred-items.md row"
  - truth: "Review Infos IN-01..IN-06 (dead acPayload.app, fired-tone-before-refusal, non-scalar franken-node, focus-move pin gap, concurrent Wait in fixture)"
    addressed_in: "07-REVIEW.md Info section (review contract scope: Critical+Warnings only)"
    evidence: "Documented with fixes; test-only or degenerate-env surface; none gates the phase goal"
---

# Phase 7: Меню v2 и чёрный список автокоррекции — Verification Report

**Phase Goal:** Довести автокоррекцию до пользовательского качества и выпустить v1.1.0: трей-меню v2 (EN/RU двумя пунктами с отметкой текущего, тумблер автокоррекции с записью в конфиг, живые горячие клавиши, «Настройки…», версия, тумблер «Звук»), чёрный список apps_blocklist (regex) с ревизией D-53; spec-delta §11 + ADR-007 amendment до кода.
**Verified:** 2026-10-07T23:22:00Z (HEAD 39a5d82)
**Status:** passed
**Re-verification:** Yes — second fingerprint-staleness refresh of the milestone-close chain. Since the previous pass (HEAD e76f6b0, 2026-10-07T21:34:08Z) quick task 261008-00m landed: the G-2-3 own-flip fix (4a0c887 — actor.go `flipCredit`, ADR-004 amendment 0781a5f), the G-5-5 auto-wrap fix (2fa8b53 — internal/install/*), matrix-stand hermeticity (4dc4877 — test/e2e/matrix.go), and docs-only spec-deltas (0781a5f: ADR-004/ADR-006/SPEC §4.3/§4.4). All 51 covered_files re-derived against the CURRENT tree — all resolve; the set is unchanged from the prior pass (internal/install/* is NOT phase-7 covered: zero references to the install package in any phase-7 plan/summary — all mentions are English prose; G-5-5 is owned by the phase-2/5 ledgers per commit 9cb032c). Digest recomputed over the refreshed contents (HEAD moved to 39a5d82 mid-pass: only 05-VERIFICATION.md, not in this phase's covered set — digest unaffected).

## Goal Achievement

Second staleness refresh: the previous verdict was passed 5/5 with zero open gaps. This pass regression-checked exactly the seams quick 261008-00m touched inside this phase's covered set — the actor (which hosts the menu sync fold, the sound sink wiring, and the blocklist gate), and the e2e matrix stand — with named -race tests; everything else (indicator/, config/, sound/, ctlsvc/, cmd/goswitchd/) is byte-unchanged since e76f6b0 (`git diff e76f6b0..HEAD` over those paths is empty), so prior full evidence stands. No regressions found.

### Observable Truths (roadmap success criteria)

| # | Truth | Status | Evidence (current tree, HEAD 39a5d82) |
|---|-------|--------|---------------------------------------|
| 1 | Меню v2 живьём: EN/RU пункты переключают с отметкой текущего; тумблер автокоррекции persist-ит в конфиг (hot reload подхватывает); макро-инфопункты живые; «Настройки…» открывает config.yaml; версия видна; тумблер «Звук» | ✓ VERIFIED | internal/indicator/ + cmd/goswitchd/ byte-unchanged since e76f6b0 — prior full evidence (owner UAT 5/5 + TestMenu corpus + e2e-menu-v2) stands. The actor-side seams this truth runs through re-verified on the post-4a0c887 tree: SwitchMode (:1105, menu EN/RU = ONE MORE GESTURE on the single flipTo, ADR-006) holds a.mu at :1106 before flipTo — the new flipCredit increment (:2017) is therefore a.mu-only, race-clean under -race; menu sync fold seam intact (SetMenuSync :906, SetModeDisplay :896); TestActor_MenuSyncApplySnapshotPushesSound and TestMenuTogglePushRacesAttachStore both GREEN -race on the current tree. The flipTo modification is additive (arms one credit AFTER a successful round trip; the flip path, the D-36 order and the display-observer-last discipline untouched) |
| 2 | Blocklist: `apps_blocklist` (regex, подстрока, анкеровка явная) — совпадение запрещает; пустой = ничего не запрещено; unknown-идентичность НЕ запрет; hot reload; strict decode; CONFIG.md дополнен | ✓ VERIFIED | Schema and gate untouched by 261008-00m (internal/config/ has zero diff since e76f6b0): `AppsBlocklist []string yaml:"apps_blocklist"` (config.go:119), ceiling + per-pattern validation (:406–417), writer key :30/:231. Actor gate re-confirmed on the current tree: acReasonRoleAmbiguous :99, cache refresh on fold (:510/:1354), refreshACBlocklist :2335 / compileACBlocklist :2349, role-ambiguous refusal :2556. TestAutoCorrect_IdentityUnknownNotProhibition GREEN -race on the post-fix tree. The blocklist/autocorrect layer fired LIVE on the fix tree: matrix v1 16/16 with `autocorrect fired` records (orchestrator-recorded, 261008-00m-VERIFICATION.md:97) |
| 3 | Белый список `autocorrect.apps` удалён из схемы/кода/корпуса/доков; D-53-молчание заменено «blocklist + полевой role-гейт»; регресс перезакреплён | ✓ VERIFIED | The Autocorrect struct carries NO `apps` field (config.go:117–123, unchanged); the only `yaml:"apps"` is MACR.Apps (:106, the pre-existing phase-3 feature — adversarial check stands). The docs-only delta 0781a5f did NOT clobber the audit trail: SPEC §11 present at its new offset (:409, shifted by the §4.3/§4.4 half-state amendments) with the blocklist semantics (:436–449, D-53 conjunction + unknown-identity pass); ADR-007 Amendment (2026-10-04, D-53, :193) + dated role-61 Amendment note (:246+) both intact. CR-01 slug in code (:99) unchanged |
| 4 | Регресс фаз 5-6 зелёный: `mise run ci`, матрица v4 (перезакреплена), flip-keystroke 24/24 | ✓ VERIFIED | Focused per instruction (no full suite): every seam 261008-00m touched inside this phase's covered set re-run GREEN under `-race` on the current tree — actor corpus: TestActor_MenuSyncApplySnapshotPushesSound + TestActor_SoundFoldOffStopsSink + TestAutoCorrect_IdentityUnknownNotProhibition + the three new flipCredit pins (TestActor_OwnFlipMixedWordCorrectsWhole, TestActor_OwnFlipCreditConsumedOnce, TestActor_OwnFlipFailedSwitcherArmsNothing); cmd/goswitchd: TestMenuTogglePushRacesAttachStore; test/e2e: TestMatrixCaseNeedsBaseEstablishment (the new hermeticity predicate). LIVE evidence on the fix tree recorded by the orchestrator: matrix v1 16/16 with the G-2-3 word-mixed row settling to «паиghbdtn», matrix v2 20/21 with the single FAIL the documented environmental focus-stealing class (261008-00m-VERIFICATION.md:97–98) — not re-run by verifier. matrix-v4.yaml expectations untouched by the G-2-3/G-5-5 commits (e2e expectations diff empty; matrix.go change is stand-hermeticity plumbing only, behavior-pinned). Prior full `mise run ci` evidence covers the unchanged remainder |
| 5 | Релиз v1.1.0: тег → goreleaser → ассеты; README/CONFIG.md соответствуют v1.1.0 | ✓ VERIFIED | `git tag --list "v1.1*"` → v1.1.0 (unchanged; tag integrity verified in the prior pass against the fully-verified tree). README.md and docs/CONFIG.md have ZERO diff since e76f6b0 — the v1.1.0 docs correspondence stands. The post-release 261008-00m fixes are milestone-close quick work with spec-delta audit trail (0781a5f) and their own verified ledger (6/6) — they do not reopen the release truth |

**Score:** 5/5 truths verified (0 present-behavior-unverified)

### Cross-Reference: G-2-3 flipCredit ownership and this phase's seams

Quick 261008-00m's actor.go change (flipCredit, ADR-004 amendment) lives inside a file this phase covers, so its interaction with phase-7 seams was checked, not assumed: (a) the flip path it arms on IS the menu's EN/RU SwitchMode→flipTo path — verified additive and mutex-clean (SwitchMode holds a.mu at :1106; the comment's serialization claim holds); (b) the HandleLifecycle FocusOut branch it modifies skips ONLY the buffer hard-reset — every other side effect (FSM reset, surrounding-cache clear, pending retirement) preserved, so no phase-7 behavior (menu sync, sound fold, blocklist) depends on the skipped reset; (c) the sound fold and menu sync seams are byte-identical and re-exercised green by named tests. The fix's OWN behavioral truth (word across own flip corrects whole) is pinned by three new tests, all green under -race, and witnessed live (matrix v1 16/16). The G-5-5 install change (2fa8b53) is outside this phase's covered set (phases 2/5 ledgers own it).

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| docs/SPEC.md | §11 blocklist revision + role-61 addition surviving the NEW §4.3/§4.4 deltas (0781a5f) | ✓ VERIFIED | §11 at :409 (offset shifted again by the half-state amendments), blocklist conjunction :436–449, audit-trail pattern (dated HTML-comment amendments) consistent with the house style |
| docs/adr/ADR-007-….md | Amendment (2026-10-04, D-53) + role-61 amendment note | ✓ VERIFIED | :193 and :246+ both present on the current tree; untouched by 0781a5f (which amended ADR-004/ADR-006) |
| internal/config/{config,writer}.go | AppsBlocklist schema + sentinels + ceiling; persist setters | ✓ VERIFIED | Zero diff since e76f6b0: config.go:119/:406–417, writer.go:30/:231 |
| internal/session/actor.go | app-blocked gate, unknown⇒pass, role-ambiguous refusal, SwitchMode, SoundSink seam + fold, menu sync; NOW also flipCredit (G-2-3) | ✓ VERIFIED | :99 slug, :2335–2349 blocklist cache, :1105 SwitchMode (a.mu discipline), :934–937 SoundSink + SetAutocorrectEvent, :906 SetMenuSync; flipCredit :263/:2017/:788–790 behavior-pinned green under -race |
| internal/indicator/{menu,indicator}.go | Menu v2 snapshot, IDs pinned, greyed info rows, trimmed statusLabel | ✓ VERIFIED | Byte-unchanged since e76f6b0 (`git diff` empty); prior full evidence stands |
| cmd/goswitchd/main.go | wiring: menuSlot, toggles, xdg-open WR-03 guard, sound sink, gsettings watcher | ✓ VERIFIED | Byte-unchanged since e76f6b0; TestMenuTogglePushRacesAttachStore re-run green |
| test/e2e/matrix.go + matrix_test.go | matrix driver; NOW hermeticity predicate (every case pins the base config) | ✓ VERIFIED | caseNeedsBaseEstablishment/caseHasReloadSteps added (4dc4877); TestMatrixCaseNeedsBaseEstablishment pins all three rows, GREEN -race; expectations in matrix-v4.yaml untouched |
| internal/install/* | (NOT phase-7 covered) | N/A | G-5-5 auto-wrap (2fa8b53) is owned by the phase-2/5 verification ledgers (9cb032c); zero phase-7 plan/summary references the install package — covered set correctly excludes it |
| README.md / docs/CONFIG.md | v1.1.0 docs, silent-apps remedies, role-61 notes | ✓ VERIFIED | Zero diff since e76f6b0 — correspondence stands |
| 07-UAT.md | owner UAT session, 5/5 complete | ✓ VERIFIED | status: complete — unchanged |

### Key Link Verification

All previously ✓ WIRED links hold on the current tree; the only rewiring risk (flipTo's new arming branch) was traced by hand: menu→SwitchMode→flipTo (menu.go dispatch → actor.go:1105, a.mu held :1106), toggle→persist→reload→fold (main.go → writer.go → Watcher → actor fold — main.go/writer.go byte-unchanged), flipTo→SoundSink.Flip / fired→SoundSink.AutoCorrect (seam :934–937 byte-unchanged, fold test green), e2e oracles→pinned IDs (menu.go byte-unchanged), blocklist gate→compiled patterns (:2335–2349 byte-identical), config keys↔README/CONFIG.md (docs byte-unchanged), SPEC §11↔ADR-007↔code slugs (:99). New link from the fix (successful round trip → one credit → FocusOut skips exactly HardReset) verified by TestActor_OwnFlipMixedWordCorrectsWhole / TestActor_OwnFlipCreditConsumedOnce / TestActor_OwnFlipFailedSwitcherArmsNothing, all green under -race.

### Data-Flow Trace (Level 4)

No rendering surfaces in this phase's artifact set (daemon + engine + e2e stand). The dynamic-value chains (config fold → blocklist cache, config fold → sound event, menu snapshot push) are byte-unchanged and were traced in the prior pass; the one new state variable (flipCredit) flows switcher round-trip result → credit → FocusOut branch and is value-pinned by the three named tests above.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Menu sync fold + sound sink wiring (touched package) | `go test -race -count=1 -run 'TestActor_MenuSyncApplySnapshotPushesSound\|TestActor_SoundFoldOffStopsSink\|TestAutoCorrect_IdentityUnknownNotProhibition\|TestActor_OwnFlipMixedWordCorrectsWhole' ./internal/session/` | ok 1.071s | ✓ PASS |
| flipCredit edge pins (over-suppression + failed arming) | `go test -race -count=1 -run 'TestActor_OwnFlipCreditConsumedOnce\|TestActor_OwnFlipFailedSwitcherArmsNothing' ./internal/session/` | ok 1.029s | ✓ PASS |
| Menu toggle push/attach race (as prior pass) | `go test -race -count=1 -run 'TestMenuTogglePushRacesAttachStore' ./cmd/goswitchd/` | ok 1.027s | ✓ PASS |
| Matrix hermeticity predicate (touched file) | `go test -race -count=1 -run 'TestMatrixCaseNeedsBaseEstablishment' ./test/e2e/` | ok 1.010s | ✓ PASS |
| Release tag integrity | `git tag --list "v1.1*"` | v1.1.0 present | ✓ PASS |
| Untouched-seam proof | `git diff e76f6b0..HEAD --name-only` over indicator/config/sound/ctlsvc/cmd | empty | ✓ PASS |
| Debt markers on changed covered files | TBD/FIXME/XXX/PLACEHOLDER scan over actor.go, actor_test.go, matrix.go, matrix_test.go | 0 hits | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| mise live e2e tasks (e2e-matrix, e2e-menu-v2, e2e-autocorrect-blocklist-silent) | not re-run by verifier | live GNOME-desktop only; the fix-tree live evidence is RECORDED by the orchestrator (261008-00m-VERIFICATION.md:97–98: matrix v1 16/16, matrix v2 20/21 with the single FAIL the documented environmental focus-stealing class, `autocorrect fired` records present) — per instruction, not re-run | ? SKIP — recorded live evidence on the fix tree is the resolution record |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| SPEC §11 (spec-delta) | 07-01, 07-02, 07-08 | blocklist revision + sound decision recorded in spec BEFORE code (D-55) | ✓ SATISFIED | §11 + ADR-007 audit trail survive BOTH delta waves (§4.2/§8 earlier; §4.3/§4.4 + ADR-004/ADR-006 from 0781a5f) |
| MACR-ACL (blocklist-ревизия) | 07-02..07-07 | white-list → blocklist revision end to end | ✓ SATISFIED | schema+gate+docs byte-unchanged or re-confirmed; identity pin green; blocklist layer fired live on the fix tree (matrix v1 16/16) |
| CORR-01..09 (регресс) | 07-04, 07-06, 07-07 | manual-correction regression pins | ✓ SATISFIED | manual-double alive; the G-2-3 fix STRENGTHENS the correction contract (word across own flip now corrects whole) and is behavior-pinned green |
| SWCH-01..04 (регресс) | 07-05, 07-06, 07-07 | switch-path regression through the single flipTo | ✓ SATISFIED | EN/RU menu → SwitchMode → flipTo unchanged in discipline (a.mu, D-36 order, display-last); flipCredit is additive post-round-trip state; TestActor_OwnFlip* + TestActor_FlipInvokesDisplay corpus green |

Orphaned requirements: none — REQUIREMENTS.md unchanged in the e76f6b0..HEAD diffset.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | TBD/FIXME/XXX/placeholder/empty-impl scan across the covered files changed since the prior pass (actor.go, actor_test.go, matrix.go, matrix_test.go) | clean | — |

Info (non-blocking, not a phase-7 gap): 261008-00m's ledger noted one flaky `internal/ctlsvc` TestSvc_ReloadInvalidKeepsLastGood under full parallel `-race` (inotify instance exhaustion; passes in isolation) — environmental, ctlsvc is byte-unchanged since the prior phase-7 pass and belongs to the phase-5/6 surfaces.

### Advisory (New Scope, Unevidenced)

New-scope findings from Step 7 with no deterministic evidence — reported, not blocking.

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| (none) | — | — | — |

### Human Verification Required

None at phase level. All five former phase-7 human items remain closed by the owner's UAT verdicts (07-UAT.md: complete, 5/5). The flipCredit behavior is unit-pinned under -race at the canonical seams and witnessed live (matrix v1 16/16 on the fix tree); the matrix v2 20/21 single FAIL is the documented environmental focus-stealing class owned by the nightly/phase-3 ledger, not a phase-7 must-have. Live by-ear sound gates remain tracked at the 261006-squ task ledger.

### Gaps Summary

None. The second fingerprint staleness is resolved: all 5 must-haves hold on the current tree (HEAD 39a5d82). The 261008-00m diffset inside this phase's covered set (actor.go flipCredit, actor_test.go corpus, matrix.go hermeticity) is regression-checked with named -race tests — all green; the untouched seams (indicator, config, sound, ctlsvc, cmd) are proven byte-unchanged; the docs-only deltas preserved the SPEC §11 / ADR-007 audit trail; internal/install is correctly outside this phase's covered set. The prior deferred G-2-3 item is now FULLY RESOLVED (fix landed + live 16/16 witness); the remaining deferred items are unchanged and none gates the phase goal.

---

_Verified: 2026-10-07T23:22:00Z_
_Verifier: Claude (gsd-verifier)_
