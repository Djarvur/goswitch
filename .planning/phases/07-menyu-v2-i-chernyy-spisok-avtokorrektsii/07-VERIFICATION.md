---
phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
verified: "2026-10-07T21:34:08Z"
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
covered_digest: "v1:sha256:e077ad22b015e8108dd3ba85fcc19282eb459bdb20ba17cf71536645327c7153"
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
    addressed_in: "RESOLVED by quick 261006-vqw (owner verdict 2026-10-06): rows re-pinned to per-character layout inversion (f9ed497, matrix v1..v4); residual G-2-3 word-mixed witness (own flip resets correction buffer mid-word) is owned by the milestone-close batch 261008-00m (e76f6b0)"
    evidence: "STATE.md:316 (261006-vqw row); matrix-v4.yaml:28-30 header (WINDOWS-12 freeze rejected, per-character re-pin); e76f6b0 G-2-3 witness"
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
**Verified:** 2026-10-07T21:34:08Z (HEAD e76f6b0)
**Status:** passed
**Re-verification:** Yes — fingerprint-staleness refresh at milestone close. Since the previous pass (HEAD da4d694, 2026-10-05) the tree gained: quick 261006-squ (persistent PulseAudio stream — internal/sound/ rewritten), quick 261006-vqw (per-character mixed-text inversion, matrix rows re-pinned), quick 261007-0yg (engine/ + layouts/ moved under internal/, commit 5427d7a; ctlsvc OnConn wait stabilized), and phase-8 documentation/schema work. All 50 covered_files paths re-derived against the CURRENT tree — all resolve (the previous list already used internal/ paths; the moved engine/layouts packages were never in this phase's covered set). The digest was recomputed over the refreshed contents.

## Goal Achievement

Fingerprint-staleness re-verification at milestone close: the previous verdict was passed 5/5 with zero open gaps. This pass re-verified every must-have against the CURRENT tree (HEAD e76f6b0) — quick regression where prior evidence + unchanged paths suffice (indicator/ package is byte-unchanged since da4d694), full re-checks where paths/wiring changed (sound backend rewrite, actor sound fold, main.go wiring, e2e matrix re-pin). No regressions found.

### Observable Truths (roadmap success criteria)

| # | Truth | Status | Evidence (current tree, HEAD e76f6b0) |
|---|-------|--------|---------------------------------------|
| 1 | Меню v2 живьём: EN/RU пункты переключают с отметкой текущего; тумблер автокоррекции persist-ит в конфиг (hot reload подхватывает); макро-инфопункты живые; «Настройки…» открывает config.yaml; версия видна; тумблер «Звук» | ✓ VERIFIED | internal/indicator/ (menu.go, indicator.go) byte-unchanged since da4d694 — prior full evidence (owner UAT 5/5 + TestMenu corpus + e2e-menu-v2) stands unchanged; fresh greps re-confirm the load-bearing surface: IDs pinned (menuIDAbout=13, menuIDStatus=14, menu.go:69–70), labelDelta dispatch rows (:217/:271/:324), greyed Status/About info rows (:384), trimmed «<MODE> · испр. <N>» statusLabel (:670–676), notify path absent (0 hits in non-test code). cmd wiring re-checked after the +36 diff: menuSlot atomic.Pointer (main.go:145), toggles persist via config.SetAutocorrectEnabled/SetSoundEnabled (:248/:259), menu re-sync from the folded config (:181–182, the Sound toggle now reading cfg.Sound.EffectiveEnabled() — the phase-8 nil=ON bool contract, same wiring line evolved compatibly), «Настройки…» → EnsureDocument + xdg-open no-pipes with WR-03 guard (:311), gsettings watcher from daemon start (:192–198). TestMenuTogglePushRacesAttachStore green on the current tree. EN/RU → SwitchMode (actor.go:1071) → flipTo single flip path intact |
| 2 | Blocklist: `apps_blocklist` (regex, подстрока, анкеровка явная) — совпадение запрещает; пустой = ничего не запрещено; unknown-идентичность НЕ запрет; hot reload; strict decode; CONFIG.md дополнен | ✓ VERIFIED | Schema intact on the current tree: `AppsBlocklist []string yaml:"apps_blocklist"` (config.go:119), ceiling 64 with OverCeil sentinel (:406–409); actor gate fully present — compiled blocklist cache (:228, refreshAC/compileACBlocklist :2279–2302), live-identity consult on confirm (:2379–2380, blocklist-only confirm gate). TestAutoCorrect_IdentityUnknownNotProhibition (actor_test.go:5543) re-run GREEN by this verifier on the current tree. CONFIG.md rows present (:71 blocklist semantics, :129 default, :186 example). WR-02 hot sound event preserved (SetAutocorrectEvent in the SoundSink interface, actor.go:900–905) |
| 3 | Белый список `autocorrect.apps` удалён из схемы/кода/корпуса/доков; D-53-молчание заменено «blocklist + полевой role-гейт»; регресс перезакреплён | ✓ VERIFIED | The Autocorrect struct carries NO `apps` field (config.go:117–123) — the only `yaml:"apps"` in config.go is MACR.Apps (:106), the pre-existing phase-3 Super→Ctrl per-app allow list, a different feature (adversarial check: NOT a white-list regression). SPEC §11 present at its new offset (:391, shifted by the later §4.2/§8 spec-deltas) with the blocklist conjunction (:418); ADR-007 Amendment (2026-10-04, D-53, :193) + dated role-61 Amendment note (2026-10-05, :248) both intact; CR-01 tightening in code: acReasonRoleAmbiguous (actor.go:99), refusal path, role-61 fail-closed (:2389) |
| 4 | Регресс фаз 5-6 зелёный: `mise run ci`, матрица v4 (перезакреплена), flip-keystroke 24/24 | ✓ VERIFIED | Quick regression per re-verification scope (full suite not re-run by instruction): `go build ./...` GREEN on the current tree — proves the engine/+layouts/→internal/ move left no dangling imports across all 16 internal packages; named behavioral tests GREEN: TestActor_SoundFoldOffStopsSink (actor_test.go:6433 — the changed sound fold wiring) and TestAutoCorrect_IdentityUnknownNotProhibition; test/e2e package compiles clean (case_menu/case_autocorrect/matrix + tests). Matrix v4 re-pin verified in the file: mixed rows re-pinned to per-character inversion with the decision header (matrix-v4.yaml:28–30, rows :110/:161); the double-run 31/34 acceptance stands as recorded (STATE.md:268, owner-acknowledged — design rows re-pinned by the owner's 2026-10-06 mixed-text verdict; NOT re-opened). Every commit since da4d694 landed under the green-iteration gate (build+test -race+lint per directives); the prior pass's full ci 19/19 evidence covers the unchanged remainder |
| 5 | Релиз v1.1.0: тег → goreleaser → ассеты; README/CONFIG.md соответствуют v1.1.0 | ✓ VERIFIED | The release has since HAPPENED: `git tag --list "v1.1*"` → v1.1.0, pointing at 42cbe72 (merge of PR #12, the phase-7 line) with da4d694 — the previously fully-verified tree — as its ancestor (tag integrity checked). README.md carries the blocklist semantics + «Where autocorrect stays silent — and how to fix it» remedies; the phase-8 README a11y rewrites (bool contract) preserve the section in both README.md and README.en.md:228 (+ :300 cross-ref); CONFIG.md blocklist/sound/a11y rows current (:71/:129/:186). The previous pass's no-premature-tag prohibition is superseded by the owner's release action at milestone close |

**Score:** 5/5 truths verified (0 present-behavior-unverified)

### Cross-Reference: sound backend rewrite ownership

The quick task 261006-squ (landed AFTER the phase-7 UAT verdicts) replaced the canberra→paplay spawn player with a persistent PulseAudio stream + XDG sound-theme resolver behind the SAME SoundSink seam (actor.go:900–905: Flip/AutoCorrect/SetAutocorrectEvent, plus the optional SoundStopper mute lifecycle). Phase 7's must-haves assert the seam, the menu toggle, persist and hot reload — all re-verified above. The rewritten backend's inherently-manual by-ear acceptance gates (latency feel, theme retune, muted-socket silence) are owned and tracked by that task's own verification ledger (261006-squ-VERIFICATION.md: status human_needed, 7/7 automatable must-haves verified) — not duplicated here; they do not attach to any phase-7 must-have. The ON→OFF fold-stops-sink behavior added by the rewrite is behaviorally pinned and was re-run green (TestActor_SoundFoldOffStopsSink).

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| docs/SPEC.md | §11 blocklist revision + role-61 addition; later §4.2/§8 spec-deltas must not have clobbered it | ✓ VERIFIED | §11 at :391 (offset shifted by later deltas), blocklist conjunction :418, role-61 caveat intact; audit trail preserved |
| docs/adr/ADR-007-….md | Amendment (2026-10-04, D-53) + role-61 amendment note | ✓ VERIFIED | :193 and :248 both present on the current tree |
| internal/config/{config,writer}.go | AppsBlocklist schema + sentinels + ceiling; persist setters, EnsureDocument/DefaultPath | ✓ VERIFIED | config.go:119/:406; writer.go:62/:69/:76/:99 all present (phase-8 added the a11y section and the Sound nil=ON pointer — additive, compatible) |
| internal/session/actor.go | app-blocked gate, unknown⇒pass, role-ambiguous refusal, SwitchMode, SoundSink seam + fold | ✓ VERIFIED | :99 slug, :228/:2279–2302 blocklist cache, :2379–2380 confirm gate, :1071 SwitchMode, :900–905 SoundSink + SoundStopper; fold-stop behavior test green |
| internal/indicator/{menu,indicator}.go | Menu v2 snapshot, IDs pinned, greyed info rows, trimmed statusLabel, no notify path | ✓ VERIFIED | Byte-unchanged since da4d694; fresh greps confirm :69–70/:217/:384/:670–676 |
| cmd/goswitchd/main.go | wiring: menuSlot, toggles, xdg-open WR-03 guard, sound sink, gsettings watcher | ✓ VERIFIED | :145/:181–182/:248/:259/:311/:192–198 re-checked after the +36 diff |
| internal/sound/sound.go | (was: canberra→paplay player; NOW: persistent-stream Player per quick 261006-squ) | ✓ VERIFIED | Rewritten by 261006-squ behind the same seam: pulse.go/theme.go/pcm.go added; own verification 7/7 automatable; named fold test green |
| test/e2e/* (cases, matrix) | menu-v2 + blocklist-silent cases, matrix re-pinned to per-character mixed semantics | ✓ VERIFIED | matrix-v4.yaml re-pin header :28–30 + rows :110/:161; package compiles clean |
| README.md / docs/CONFIG.md | v1.1.0 docs, silent-apps remedies, role-61 notes | ✓ VERIFIED | README silent-apps section survives the phase-8 a11y rewrite (README.en.md:228/:300 mirror); CONFIG.md :71/:129/:186 |
| 07-UAT.md | owner UAT session, 5/5 complete | ✓ VERIFIED | status: complete, passed: 5 / pending: 0 — unchanged |

### Key Link Verification

All previously ✓ WIRED links re-confirmed on the current tree via the truth-table greps: menu→SwitchMode→flipTo (menu.go dispatch → actor.go:1071), toggle→persist→reload→fold (main.go:248/:259 → writer.go:62/:69 → Watcher → actor fold), flipTo→SoundSink.Flip and fired→SoundSink.AutoCorrect (actor.go seam, fold-stop test green), fold→EffectiveEnabled/EffectiveAutocorrectEvent (main.go:182/:198 — now through the nil=ON pointer contract), e2e oracles→pinned IDs (menu.go:69–70), blocklist gate→compiled patterns (actor.go:2279–2302), config keys↔README/CONFIG.md (:71/:129/:186), SPEC §11↔ADR-007↔code slugs (:99). New links from the rewrite (daemon starts the sink, ON→OFF fold stops it) verified by TestActor_SoundFoldOffStopsSink green.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Whole-tree build after the engine/layouts move | `go build ./...` | exit 0, no output | ✓ PASS |
| ON→OFF sound fold stops the sink (changed wiring) | `go test -run 'TestActor_SoundFoldOffStopsSink' -count=1 ./internal/session/` | ok 0.005s | ✓ PASS |
| Blocklist unknown-identity pin (changed package) | `go test -run 'TestAutoCorrect_IdentityUnknownNotProhibition' -count=1 ./internal/session/` | ok 0.031s | ✓ PASS |
| Menu toggle push/attach race (changed main.go) | `go test -run 'TestMenuTogglePushRacesAttachStore' -count=1 ./cmd/goswitchd/` | ok 0.006s | ✓ PASS |
| e2e oracle package compiles | `go test -run 'ZZZNoSuchTest' -count=1 ./test/e2e/` | ok (compiles incl. tests) | ✓ PASS |
| Release tag integrity | `git tag --list "v1.1*"` + merge-base --is-ancestor da4d694 v1.1.0 | v1.1.0 exists; verified tree is its ancestor | ✓ PASS |
| Debt markers on changed files | TBD/FIXME/XXX scan over sound/ actor.go main.go config.go e2e | 0 hits | ✓ PASS |
| Disabled-test scan on changed tests | skip-scan over actor_test.go / sound_test.go / menu_test.go | no real skips | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| mise live e2e tasks (e2e-matrix-v4, e2e-flip-keystroke, e2e-menu-v2, e2e-autocorrect-blocklist-silent) | not re-run by verifier | live GNOME-desktop only (>10 s budget, spawns daemons); the matrix double-run 31/34 acceptance is the recorded milestone state (STATE.md:268, owner-acknowledged, design rows re-pinned by the 2026-10-06 verdict) — not re-opened per milestone-close instruction | ? SKIP — recorded acceptance state + owner verdicts are the resolution record |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| SPEC §11 (spec-delta) | 07-01, 07-02, 07-08 | blocklist revision + sound decision recorded in spec BEFORE code (D-55) | ✓ SATISFIED | §11 revisions + role-61 addition survive the later §4.2/§8 deltas; ADR-007 amendment + note intact |
| MACR-ACL (blocklist-ревизия) | 07-02..07-07 | white-list → blocklist revision end to end | ✓ SATISFIED | schema+validation+gate+confirm+e2e negative+docs all re-confirmed on the current tree; identity pin test green |
| CORR-01..09 (регресс) | 07-04, 07-06, 07-07 | manual-correction regression pins | ✓ SATISFIED | manual-double alive; mixed semantics re-pinned per-character (owner verdict); build + named tests green |
| SWCH-01..04 (регресс) | 07-05, 07-06, 07-07 | switch-path regression through the single flipTo | ✓ SATISFIED | EN/RU menu → SwitchMode → flipTo unchanged; flip-keystroke acceptance recorded |

Orphaned requirements: none — REQUIREMENTS.md unchanged in the da4d694..e76f6b0 diffset; all v1 REQ-ID mappings stand.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | TBD/FIXME/XXX/placeholder/empty-impl scan across all files changed since the prior pass (sound/*, actor.go, main.go, config.go, e2e/*, matrix-v4.yaml) | clean | — |

Note: the word "todo" appears in sound package doc comments only as provenance citations to the owner's backlog item «Звук» that AUTHORIZED and was CLOSED by those very commits (fa6c81f "todo closed") — not unresolved-work markers.

### Human Verification Required

None at phase level. All five former phase-7 human items remain closed by the owner's UAT verdicts (07-UAT.md: complete, 5/5). The only live by-ear gates on the current tree (sound latency feel, theme retune, muted-socket silence) attach to the quick task 261006-squ's own verification ledger (human_needed there, 7/7 automatable verified) — tracked at task level, not a phase-7 must-have.

### Gaps Summary

None. The fingerprint staleness is resolved: all 5 must-haves hold on the current tree (HEAD e76f6b0). The menu v2 surface is byte-unchanged since the owner-approved pass; the blocklist revision, SPEC/ADR audit trail, and release docs are intact through the later spec-deltas; the engine/+layouts/→internal/ move builds clean tree-wide; the rewritten sound backend sits behind the same verified seam with its behavioral fold pin green; the v1.1.0 tag exists and contains the fully-verified phase-7 tree. The matrix acceptance stands as recorded. Remaining items are accurately deferred (mixed-text G-2-3 witness → milestone-close batch; by-ear sound gates → 261006-squ ledger; silent-apps tuning, formal fresh-session gate, the ~1/15 watch flake, review Infos) — none gates the phase goal.

---

_Verified: 2026-10-07T21:34:08Z_
_Verifier: Claude (gsd-verifier)_
