---
phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
verified: "2026-10-05T09:35:44Z"
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
covered_digest: "v1:sha256:9d7fc28dd99f8606d9e9cd9c6ccd9fb4188d358b425726d7535f55b535b8558f"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: human_needed
  previous_score: 5/5
  gaps_closed:
    - "Human item 1 (menu pixel-check) — RESOLVED: owner UAT pass (07-UAT test 1); UAT-driven menu changes landed per owner verdicts — greyed Status/About live info rows replace notifications (97b29f0), status row trimmed to «<MODE> · испр. N» (822dcce), notify path removed; re-checked on the current tree"
    - "Human item 2 (live sound hearing) — RESOLVED: owner heard the flip tone AND a DISTINCT autocorrect tone; «Звук»-off silence accepted; ~150-300 ms spawn latency recorded as post-release tuning (deferred-items.md, 07-UAT Deferred Follow-Ups)"
    - "Human item 3 (editor-open A1) — RESOLVED: owner confirmed «Настройки…» creates the full defaults config.yaml and opens it in the desktop editor (07-UAT test 3)"
    - "Human item 4 (fresh-session D-48 gate + live re-run on the final tree) — RESOLVED: matrix v4 ×2 on the final tree (c84709c/4b75680) 31/34 in both — 2 designed WINDOWS #12 rows byte-identical to the ledger + 1 ibus activation transient green in the neighboring run (honest record); the formal ≤30-min fresh-session gate is delegated to the nightly pipeline / owner per the WINDOWS #13 phase-6 precedent"
    - "Human item 5 (role-61 doc caveat, CR-01 doc leg) — RESOLVED: dated (2026-10-05) role-61 caveat landed in SPEC §11 (:249), ADR-007 amendment note (:248), README «Where autocorrect stays silent» (:222) and CONFIG.md Privacy per owner direction (5b2617f)"
  gaps_remaining: []
  regressions: []
deferred:
  - truth: "word-mixed / phrase-mixed matrix rows (mixed text across a flip boundary)"
    addressed_in: "WINDOWS #12 backlog (owner verdict, phase-5 residue)"
    evidence: "WINDOWS.md:29 rows CASES-FROZEN; owner owns re-shape-or-revisit; byte-identical to the saved reports in both final-tree runs"
  - truth: "Sound latency (~150-300 ms player subprocess spawn)"
    addressed_in: "Post-release tuning (owner UAT verdict)"
    evidence: "deferred-items.md post-release-tuning row; 07-UAT Deferred Follow-Ups: earlier hook in flipTo, pw-play/resident player, theme preload"
  - truth: "Silent apps (Electron/Chromium without accessibility mode: ZCode, Telegram snap) reliability tuning"
    addressed_in: "Owner-deferred backlog"
    evidence: "Remedies documented (toolkit-accessibility gsettings / --force-renderer-accessibility, README.md:222-240, CONFIG.md:229-236); owner confirmed autocorrect works in ZCode after the remedy"
  - truth: "Formal fresh-session ≤30-min D-48 matrix gate"
    addressed_in: "Nightly pipeline / owner (WINDOWS #13 phase-6 precedent)"
    evidence: "07-UAT test 4 note: performed ×2 on an aged session with honest recording; formal gate delegated"
  - truth: "Pre-existing flakes: TestRun_OnConnHookCalledOnce (ctlsvc), TestWatch_BrokenBlocklistPatternKeepsLastGood (~1/15)"
    addressed_in: "deferred-items.md (phase 07 ledger, open)"
    evidence: "Both green on package re-runs and in the fresh ci run; one ci attempt hit a transient consistent with them — immediate re-run green (re-run-on-occurrence remedy)"
  - truth: "Review Infos IN-01..IN-06 (dead acPayload.app, fired-tone-before-refusal, non-scalar franken-node, SoundEnabled in no-watcher env, focus-move pin gap, concurrent Wait in fixture)"
    addressed_in: "07-REVIEW.md Info section (review contract scope: Critical+Warnings only)"
    evidence: "Documented with fixes; test-only or degenerate-env surface; none gates the phase goal"
---

# Phase 7: Меню v2 и чёрный список автокоррекции — Verification Report

**Phase Goal:** Довести автокоррекцию до пользовательского качества и выпустить v1.1.0: трей-меню v2 (EN/RU двумя пунктами с отметкой текущего, тумблер автокоррекции с записью в конфиг, живые горячие клавиши, «Настройки…», версия, тумблер «Звук»), чёрный список apps_blocklist (regex) с ревизией D-53; spec-delta §11 + ADR-007 amendment до кода.
**Verified:** 2026-10-05T09:35:44Z (HEAD da4d694)
**Status:** passed
**Re-verification:** Yes — after review fixes (b708020..5df3583), UAT-driven menu/doc changes (97b29f0, 822dcce, 5b2617f) and UAT completion (4b75680, da4d694)

## Goal Achievement

Freshness re-verification: the previous verdict was human_needed with all 5 must-haves verified — the tree has since gained the review-fix corpus, the UAT-driven menu/doc changes, and the owner's UAT verdicts (07-UAT.md status complete, 5/5, updated 2026-10-05T09:26:31Z). This pass re-verified the must-haves against the CURRENT tree (HEAD da4d694) and closed the human-item ledger.

### Observable Truths (roadmap success criteria)

| # | Truth | Status | Evidence (current tree, HEAD da4d694) |
|---|-------|--------|---------------------------------------|
| 1 | Меню v2 живьём: EN/RU переключают с отметкой текущего; тумблер persist-ит в конфиг (hot reload подхватывает); макро-инфопункты живые; «Настройки…» открывает config.yaml; версия видна; тумблер «Звук» | ✓ VERIFIED | Menu v2 unchanged at its load-bearing surface and owner-approved on UAT (test 1 pass): IDs pinned (menuIDAbout=13, menuIDStatus=14, menu.go:69–70); EN/RU radio deltas and toggle checkmarks; dispatch → `Callbacks.Switch` → `main.go SwitchMode` → `actor.flipTo` (single flip path intact); persist via SetAutocorrectEnabled/SetSoundEnabled (writer.go) + synchronous reload + FoldAppliedConfig; EnsureDocument/DefaultPath + xdg-open no-pipes with WR-03 guard (main.go:333) — UAT test 3 pass (owner saw the editor open). UAT-driven changes verified fresh: Status/About are greyed live info rows, NOT in dispatch (menu.go:384, labelDelta :217/:271/:324), `statusLabel()` renders the trimmed «<MODE> · испр. <N>» row (menu.go:670–676), and the notify path is GONE (grep for notify/notifyStatus across main.go/indicator/actor: 0 hits in non-test code); version notification removed (menu.go:654 comment), version still visible in the About row. 13+ TestMenu corpus green; e2e-menu-v2 PASS on the live stand. Pixel acceptance = owner verdict (07-UAT test 1) |
| 2 | Blocklist: `apps_blocklist` (regex, подстрока, анкеровка явная); совпадение = запрет; пустой = ничего не запрещено; unknown-идентичность НЕ запрет; hot reload; strict decode; docs/CONFIG.md дополнен | ✓ VERIFIED | Unchanged since the prior pass and re-run green: schema `AppsBlocklist []string yaml:"apps_blocklist"` with ceiling 64 and Regex/Empty/OverCeil sentinels; legacy `autocorrect.apps` whole-document rejection test; unknown⇒не-запрет pinned (TestAutoCorrect_IdentityUnknownNotProhibition); hot reload incl. WR-02 hot sound event (SetAutocorrectEvent, TestPlayer_SetAutocorrectEvent, TestNewActor_SoundEventHotReload); CONFIG.md blocklist/sound rows + NOW the silent-apps/remedies note (:229–236). Live negative e2e-autocorrect-blocklist-silent PASS (fired=0, ровно одна ac_skip_app_blocked) |
| 3 | Белый список удалён из схемы/кода/корпуса/доков; D-53-молчание заменено «blocklist + полевой role-гейт»; регресс перезакреплён | ✓ VERIFIED | SPEC §11 third dated revision (2026-10-04) + NOW the dated (2026-10-05) owner role-61 addition (:249–255, fail-closed only unknown+61, slug role-ambiguous, append-only); ADR-007 Amendment (2026-10-04, D-53) + NOW the dated Amendment note (2026-10-05, UAT фазы 07 — role-61 оговорка, :248–261) — the CR-01 doc residue from the prior pass is CLOSED; no new ADR files. CR-01 code tightening re-verified fresh: `acReasonRoleAmbiguous = "role-ambiguous"` (actor.go:99), refusal path :2298–:2413, `TestAutoConfirm_RoleTextAmbiguity` (actor_test.go:5891) green. Old white-list slugs remain 0-occurrence; matrix-v4.yaml untouched since phase 6 (CASES-FROZEN) |
| 4 | Регресс фаз 5-6 зелёный: mise run ci; матрица v4 (перезакреплена); flip-keystroke 24/24 | ✓ VERIFIED | Fresh on the final tree: `mise exec -- go test` on all six phase packages (indicator/session/config/sound/ctlsvc/cmd) -race -count=1 — green; `mise run ci` re-run by this verifier on HEAD da4d694 — 19/19 packages ok, exit 0 (one earlier ci attempt hit a transient consistent with the two documented pre-existing flakes, deferred-items.md; immediate re-run green — the documented remedy). Matrix v4 ×2 on the final tree (c84709c/4b75680) per owner UAT honest record: 31/34 both = 2 designed WINDOWS #12 rows byte-identical to the ledger + 1 ibus activation transient green in the neighboring run — class-verdict honestly recorded, not applied to exit code; flip-keystroke 24/24 (4 green runs, one documented environment watchdog-flake rerun); e2e-menu-v2 and autocorrect-blocklist-silent PASS |
| 5 | Релиз v1.1.0: README/CONFIG.md соответствуют; тег → goreleaser → ассеты (действие владельца) | ✓ VERIFIED | README now carries the silent-apps section «Where autocorrect stays silent — and how to fix it» (:222) with the gsettings/--force-renderer-accessibility remedies (:236/:240) and the dated (UAT, 2026-10-05) owner-accept role-61 note (:278); CONFIG.md matches (:229–236); key-set consistency gates in RELEASE-READINESS 5/5 PASS. `git tag --list "v1.1*"` = 0 entries — the no-premature-tag prohibition holds; tag→goreleaser remains the owner's post-merge action by design. 07-SECURITY.md added (45 threats, 0 open, ASVS L1); 07-VALIDATION.md Nyquist-validated |

**Score:** 5/5 truths verified (0 present-behavior-unverified)

### Human Verification — Resolved (UAT ledger)

All five items from the previous verification were routed to owner UAT and are now RESOLVED (07-UAT.md: status complete, 5 passed / 0 pending, updated 2026-10-05T09:26:31Z):

| # | Former item | Owner verdict | Follow-through |
|---|-------------|---------------|----------------|
| 1 | Menu v2 pixel-check (ornaments, macro rows, mnemonics) | pass — menu accepted with changes | Greyed Status/About info rows (97b29f0), status row trimmed «<MODE> · испр. N» (822dcce), notify path removed; second pixel re-check accepted with the trim |
| 2 | Live sound hearing | pass — flip tone + DISTINCT autocorrect tone heard; toggle-off silence | Latency ~150-300 ms recorded as post-release tuning (deferred-items.md, 07-UAT) |
| 3 | «Настройки…» editor-open (A1) | pass — config created and opened in the desktop editor | — |
| 4 | Fresh-session D-48 formal gate + live re-run on final tree | pass (honest record) — matrix ×2 on final tree 31/34 both; session aged | Formal ≤30-min gate delegated to the nightly pipeline/owner (WINDOWS #13 phase-6 precedent) |
| 5 | Role-61 doc-leg owner decision | pass — docs landed per owner direction (5b2617f) | Dated caveats in SPEC §11 (:249), ADR-007 (:248), README (:222, :278), CONFIG.md (:235) |

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| docs/SPEC.md | §11 third dated revision + sound decision + dated role-61 addition, audit trail intact | ✓ VERIFIED | :212–247 revision; :249–255 addition (2026-10-05); prior verdicts byte-present |
| docs/adr/ADR-007-….md | Amendment (2026-10-04, D-53) + dated role-61 amendment note | ✓ VERIFIED | :193+ and :248–261; no ADR-008 |
| internal/config/{config,writer}.go | AppsBlocklist + Sound schema, sentinels, ceiling 64; persist setters, EnsureDocument/DefaultPath | ✓ VERIFIED | unchanged since prior pass; regression green |
| internal/session/actor.go | app-blocked gate, unknown⇒pass, role-ambiguous refusal, SwitchMode, MenuSync/SoundSink seams, fold | ✓ VERIFIED | :99 slug, :2298–2413 refusal, :2413 recordACAbstain; corrections counter feeds status row |
| internal/indicator/{menu,indicator}.go | Menu v2 snapshot, IDs pinned, greyed info rows, trimmed statusLabel, no notify path | ✓ VERIFIED | :69–70 IDs, :384 greyed-row contract, :670–676 statusLabel; notify grep clean |
| cmd/goswitchd/main.go | wiring: menuSlot atomic.Pointer, toggles, xdg-open, WR-03 guard, adopt+watch | ✓ VERIFIED | :144, :175–176, :333 guard; notify plumbing removed (−48/+16 diff) |
| internal/sound/sound.go | Player canberra→paplay, bundled tones, hot SetAutocorrectEvent, warn-once | ✓ VERIFIED | corpus green; WR-02 hot path re-run green |
| test/e2e/* (cases, matrix) | menu-v2 + blocklist-silent cases, repinned corpus, matrix config_base | ✓ VERIFIED | compiles in ci (test/e2e ok); mise tasks live-only |
| README.md / docs/CONFIG.md | v1.1.0 docs + silent-apps remedies + dated role-61 notes | ✓ VERIFIED | README:222–240, :278; CONFIG.md:229–236 |
| 07-UAT.md | owner UAT session, 5/5 complete | ✓ VERIFIED | status complete; summary 5 passed / 0 pending; updated 2026-10-05T09:26:31Z |
| 07-REVIEW.md | Fix Log for CR-01/WR-01/WR-02/WR-03 | ✓ VERIFIED | fix commits b708020/80c99f2, 713d25a/a23a47c, d10671a/2504bf6, c70d620/d727ef4, 5df3583 all present on the branch |

### Key Link Verification

Carried from the prior pass (unchanged surfaces) — all previously ✓ WIRED links re-confirmed present on the current tree via the greps in the truths table: menu→SwitchMode→flipTo, toggle→persist→reload→fold, flipTo→SoundSink.Flip, fired→SoundSink.AutoCorrect, fold→EffectiveEnabled/EffectiveAutocorrectEvent, e2e oracles→pinned IDs, blocklist gate→compiled patterns, config keys↔README/CONFIG.md, SPEC §11↔ADR-007↔CONTEXT. New/changed links re-verified: corrections counter (actor snapshot → statusLabel via ItemsPropertiesUpdated channel — menu.go:217/:481) and the removed notify path (absence confirmed).

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Phase packages under race | `mise exec -- go test ./internal/indicator/ ./internal/session/ ./internal/config/ ./internal/sound/ -race -count=1` | ok ×4 (1.08s/3.61s/1.31s/1.05s) | ✓ PASS |
| Remaining packages under race | `mise exec -- go test ./internal/ctlsvc/ ./cmd/goswitchd/ -race -count=1` | ok ×2 | ✓ PASS |
| Full iteration gate | `mise run ci` (build+vet+lint+test -race+tidy-diff) on HEAD da4d694 | 19/19 packages ok, exit 0, Finished in 6.52s (a prior attempt hit a documented-flake transient; immediate re-run green) | ✓ PASS |
| CR-01 tightening | `TestAutoConfirm_RoleTextAmbiguity` in suite (actor_test.go:5891) | green in session -race run | ✓ PASS |
| WR-01 fix intact | menuSlot atomic.Pointer (main.go:144) + TestMenuTogglePushRacesAttachStore | green in cmd -race run | ✓ PASS |
| No premature tag | `git tag --list "v1.1*"` | 0 entries | ✓ PASS |
| Debt markers on changed files | TBD/FIXME/XXX scan over README/main.go/indicator/actor (+tests) | 0 hits | ✓ PASS |
| Disabled-test scan on changed tests | t.Skip-style scan menu_test.go / actor_test.go | 0 real skips (3 matches are `SkipReasons` assertion text) | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| mise live e2e tasks (e2e-matrix-v4 ×2, e2e-flip-keystroke, e2e-menu-v2, e2e-autocorrect-blocklist-silent) | not re-run by verifier | live GNOME-desktop only (>10 s budget, spawns daemons); final-tree ×2 matrix + menu/sound/blocklist verdicts recorded honestly by the owner in 07-UAT.md test 4 and per-item results; interactive evidence @a264a54 in RELEASE-READINESS §2–4 with saved reports | ? SKIP — owner UAT verdicts are the resolution record |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| SPEC §11 (spec-delta) | 07-01, 07-02, 07-08 | blocklist revision + sound decision recorded in spec BEFORE code (D-55) | ✓ SATISFIED | §11 three dated revisions + 2026-10-05 owner addition; audit trail intact; ADR-007 amendment + note |
| MACR-ACL (blocklist-ревизия) | 07-02..07-07 | white-list → blocklist revision end to end | ✓ SATISFIED | schema+validation+gate+confirm+e2e negative+docs; CR-01 tightening shipped and green |
| CORR-01..09 (регресс) | 07-04, 07-06, 07-07 | manual-correction regression pins | ✓ SATISFIED | manual-double alive in fires; matrix + ci green on the final tree |
| SWCH-01..04 (регресс) | 07-05, 07-06, 07-07 | switch-path regression through the single flipTo | ✓ SATISFIED | EN/RU menu → SwitchMode → flipTo unchanged; matrix switch rows + flip-keystroke 24/24 |

Orphaned requirements: none — REQUIREMENTS.md unchanged since the prior pass (not in the 3c315b8..da4d694 diffset); all v1 REQ-ID mappings stand.

### Test Quality Audit

Carried from the prior pass (all phase test corpora SOUND: value/behavioral assertions, 0 skipped, 0 circular); the changed test files since (menu_test.go, actor_test.go) re-scanned: 0 disabled tests, 0 circular patterns, behavioral/value assertion levels unchanged. New UAT-driven assertions (greyed-row labels, trimmed statusLabel, corrections counter) are value-level pins in menu_test.go.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | TBD/FIXME/XXX/placeholder/empty-impl scan across phase-modified files (incl. all files changed since the prior pass) | clean | — |

### Decision Coverage

No trackable decisions in CONTEXT.md (gate result: `{skipped: true, total: 0, honored: 0, not_honored: []}` — owner decisions recorded in prose/locked-semantics form; all verified against SPEC/ADR/code).

### Gaps Summary

None. The phase goal is achieved on the current tree: menu v2 is owner-approved live (with the UAT-driven greyed-info-row redesign re-verified in code), the blocklist revision is complete end-to-end with the CR-01 fail-closed tightening and its dated doc caveats, phases 5–6 regression is green fresh on the final tree (ci 19/19 + six package -race runs; matrix class-verdict honestly recorded), and release readiness stands with the no-premature-tag prohibition intact. The five former human-verification items are closed by the owner's UAT verdicts (5/5). Remaining items are recorded as deferred follow-ups (sound latency, silent-apps reliability tuning, the formal fresh-session nightly gate, WINDOWS #12 backlog, pre-existing flakes, review Infos) — none gates the phase goal.

---

_Verified: 2026-10-05T09:35:44Z_
_Verifier: Claude (gsd-verifier)_
