---
phase: "7"
slug: "menyu-v2-i-chernyy-spisok-avtokorrektsii"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-10-04"
validated: "2026-10-05"
---

# Phase 7 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib `testing`), race detector on; golden/`-v` count conventions per CONVENTIONS.md |
| **Config file** | `mise.toml` (`[tasks]` ci/lint/test/e2e-*) + `.golangci.yml` (v2, default: all) |
| **Quick run command** | `mise run ci` (build + vet + golangci-lint + `go test -race -count=1 ./...` + tidy-diff) — measured 5.4 s at HEAD 9534888 |
| **Full suite command** | `mise run ci` + live-desktop e2e (`mise run e2e-matrix-v4`, `e2e-flip-keystroke`, `e2e-menu-v2`, `e2e-autocorrect-blocklist-silent` — live GNOME session only, never in ci) |
| **Estimated runtime** | ~7 seconds (ci); e2e tasks minutes each, live-only |

---

## Sampling Rate

- **After every task commit:** Run `mise run ci`
- **After every plan wave:** Run `mise run ci` (+ live e2e at waves 5–6 per plans 07-06/07-07)
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 7 seconds

---

## Requirements Coverage (carrier map)

| Requirement | Carrier (named, verified green this audit) | Type | Command / Evidence |
|-------------|--------------------------------------------|------|--------------------|
| SPEC §11 (spec-delta) — blocklist revision + sound decision recorded BEFORE code (D-55) | `docs/SPEC.md` §11 third dated revision (3× "2026-10-04", audit trail: «отвергнута владельцем осознанно» + 7× "2026-09-27" byte-present); ADR-007 `## Amendment (2026-10-04…D-53)` :193; no ADR-008 (`ls docs/adr` max ADR-007); gatests S11-OK / SEMANTICS-OK / SOUND-REV-OK (07-01 coverage) | doc grep | `grep -c 2026-10-04 docs/SPEC.md` → 3; `grep -n "Amendment (2026-10-04" docs/adr/ADR-007*.md` → :193 |
| SPEC §11 — sound decision implemented | `internal/sound` corpus (10 tests: argv pins canberra→paplay, RIFF magic, warn-once) + `TestPlayer_SetAutocorrectEvent` (WR-02 hot path) + `TestNewActor_SoundEventHotReload` | unit | `go test ./internal/sound/ -race -count=1`; `go test ./cmd/goswitchd/ -run TestNewActor_SoundEventHotReload` |
| MACR-ACL — blocklist schema/regex/ceiling/empty-pattern/ACTIVE | `TestValidate_BlocklistRegexCompile`, `TestValidate_BlocklistEmptyPattern`, `TestValidate_BlocklistCeilUnconditional`, `TestValidate_EnabledEmptyBlocklistWithThresholdsValid`, watch last-good pins | unit | `go test ./internal/config/ -race -count=1` |
| MACR-ACL — legacy white-list removed (schema strict decode) | `TestLoad_StrictDecodeRejectsLegacyAppsKey` (load_test.go:352); slug negative grep: `no-apps`/`app-not-listed`/`app-unknown`/`app-changed` = 0 in `internal/ cmd/ test/` (only the amendment's own deletion narrative in ADR-007) | unit + grep | `go test ./internal/config/ -run TestLoad_StrictDecodeRejectsLegacyAppsKey -v` |
| MACR-ACL — gate semantics (unknown⇒pass, confirm blocklist-only, WR-01 equality removed) | `TestAutoCorrect_IdentityUnknownNotProhibition` (direction pin, 4 fail-closed cells), `TestAutoCorrect_ConfirmBlocklistRefuses`, `TestAutoCorrect_ConfirmUnknownPasses`, `TestAutoConfirm_RoleTextAmbiguity` (CR-01 fail-closed pair: unknown+role61 refuses / known+role61 fires) | unit | `go test ./internal/session/ -run 'IdentityUnknownNotProhibition|ConfirmBlocklistRefuses|ConfirmUnknownPasses|RoleTextAmbiguity' -v` |
| MACR-ACL — persist + hot reload | writer corpus (11 tests: `TestWriter_FlipRoundTrip`, `TestWriter_CommentsSurvive`, `TestWriter_AtomicWrite`, `TestWriter_CreatesFullDocumentOnAbsent`, `TestWriter_SoundFlipReverseRead`, …), `TestToggleEmptyPathRefused` (WR-03), `TestMenuTogglePushRacesAttachStore` (WR-01, 500 interleavings -race) | unit | `go test ./internal/config/ ./cmd/goswitchd/ -race -count=1` |
| MACR-ACL — live negative (blocklist silences in a real field) | e2e `autocorrect-blocklist-silent` PASS @a264a54 (fired=0, exactly one `ac_skip_app_blocked`, field byte-verbatim) | integration (live e2e) | `mise run e2e-autocorrect-blocklist-silent` (live-only; run recorded in 07-RELEASE-READINESS §3) |
| CORR-01..09 (регресс) | matrix v4 ×2: 31/34 both runs (saved reports `e2e-report-07-07-run{1,2}.txt` — fails = 2 designed WINDOWS #12 rows word-mixed/phrase-mixed byte-equal to WINDOWS.md:29 registry actual, owner-deferred per 06-UAT; + 1 migrating ibus SetGlobalEngine transient per run, green in the sibling run); `autocorrect-fires` + `password-silent` + `terminal-silent` simplified-and-green; matrix config_base repinned on `apps_blocklist: []` (matrix-v4.yaml byte-untouched, CASES-FROZEN) | matrix (live e2e ×2) | `mise run e2e-matrix-v4` ×2 (reports verified against claims this audit); `mise run ci` |
| SWCH-01..04 (регресс) | flip-keystroke 24/24 (4 green runs); matrix switch rows green; menu EN/RU click → `cb.Switch` → `SwitchMode` → single `flipTo` path (menu.go:390 → indicator.go:103 → main.go:171 → actor.go:1017→1031, no SetGlobalEngine bypass in menu.go); 13 `TestMenu*` pins | smoke (live e2e) + unit | `mise run e2e-flip-keystroke` (recorded); `go test ./internal/indicator/ -run TestMenu -v \| grep -c "^--- PASS"` → 13 |
| Критерий 1 — menu v2 live | e2e `menu-v2` PASS @a264a54 (EN/RU mode records, both toggles persist+status same click, BOTH survive daemon restart, reverse clicks); menu IDs pinned `menuIDEN=1…menuIDReload=15` | integration (live e2e) + unit | `mise run e2e-menu-v2` (recorded in 07-RELEASE-READINESS §3) |
| Критерий 4 — regression gate | `mise run ci` exit 0 re-run by this auditor at HEAD 9534888 (19 packages -race, 5.4 s); WR-01/02/03 fix tests green at tip | full suite | `mise run ci` |
| Критерий 5 — release readiness | 07-RELEASE-READINESS.md (evidence, doc checklist 5/5 PASS incl. README↔CONFIG.md key-set diff greps, probe registry 17/17, owner post-merge steps); `git tag --list "v1.1*"` empty (no premature tag) | doc grep + git | `git tag --list "v1.1*" \| wc -l` → 0 |

### WINDOWS #12 rows (owner-deferred follow-up, NOT a gap)

`word-mixed` / `phrase-mixed` matrix rows fail identically in both runs with readback byte-equal to the documented registry actual (WINDOWS.md:29). These are the phase-5 residue rows the owner explicitly deferred to the v1.1.x backlog (06-UAT Deferred Follow-Ups); they are designed-frozen (CASES-FROZEN) and recorded honestly per the class-line verdict method (06-08). Not phase-7 coverage gaps.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 7-01-01 | 01 | 1 | SPEC §11 | — | spec-delta recorded before code; audit trail append-only | doc grep | `grep -c 2026-10-04 docs/SPEC.md` → 3; prior verdicts byte-present | ✅ | ✅ green |
| 7-01-02 | 01 | 1 | SPEC §11 | — | ADR-007 amendment, no new ADR file; cross-greps | doc grep | `grep -n "Amendment (2026-10-04" docs/adr/ADR-007*`; `ls docs/adr` | ✅ | ✅ green |
| 7-02-01 | 02 | 1 | MACR-ACL | T-7-02 | regex compile gate, empty-pattern refusal, ceiling 64, field+index errors | unit | `go test ./internal/config/ -run TestValidate_Blocklist -race -count=1` | ✅ | ✅ green |
| 7-02-02 | 02 | 1 | MACR-ACL | T-7-02 | ACTIVE condition: enabled requires thresholds; last-good reload on broken pattern | unit | `go test ./internal/config/ -race -count=1` | ✅ | ✅ green |
| 7-02-03 | 02 | 1 | MACR-ACL | T-7-02 | consumer boundary polarity + pattern cache; legacy key rejected whole-document | unit | `go test ./internal/session/ ./internal/config/ -race -count=1` | ✅ | ✅ green |
| 7-03-01 | 03 | 1 | MACR-ACL | T-7-03 | persist round-trip: comments survive, atomic temp+rename, 0600 | unit | `go test ./internal/config/ -run TestWriter -race -count=1` | ✅ | ✅ green |
| 7-03-02 | 03 | 1 | MACR-ACL | T-7-03 | ensure semantics: full strict-decodable defaults document on absent file | unit | `go test ./internal/config/ -run 'TestWriter_CreatesFullDocumentOnAbsent\|TestWriter_EnsureDocumentCreatesAndNoOps' -race -count=1` | ✅ | ✅ green |
| 7-03-03 | 03 | 1 | MACR-ACL | T-7-03 | adopt+watch default path; corrupt adopted file = loud start refusal | unit | `go test ./cmd/goswitchd/ -race -count=1` | ✅ | ✅ green |
| 7-04-01 | 04 | 1 | MACR-ACL, CORR | T-7-04 | RED corpus inversion: unknown-pass, confirm cells, fail-closed direction | unit | `go test ./internal/session/ -run TestAutoCorrect_IdentityUnknownNotProhibition -v` | ✅ | ✅ green |
| 7-04-02 | 04 | 1 | MACR-ACL | T-7-04 | arm gate: unknown ⇒ pass (only identity segment); ensureAppid by enabled | unit | `go test ./internal/session/ -run 'UnknownIdentityFires\|KnownNotBlockedFires' -race -count=1` | ✅ | ✅ green |
| 7-04-03 | 04 | 1 | MACR-ACL | T-7-04 | confirm blocklist-only on CURRENT identity; WR-01 equality removed no-remainder | unit | `go test ./internal/session/ -run 'TestAutoCorrect_ConfirmBlocklistRefuses\|TestAutoCorrect_ConfirmUnknownPasses' -v` | ✅ | ✅ green |
| 7-05-01 | 05 | 4 | SWCH-01..04 | T-7-05 | menu v2 snapshot: pinned IDs, ornaments, mnemonic doubling | unit | `go test ./internal/indicator/ -run TestMenu -v` (13 PASS) | ✅ | ✅ green |
| 7-05-02 | 05 | 4 | SWCH-01..04 | T-7-05 | SwitchMode→flipTo single path; MenuSync pushes observer-last; dedupe | unit | `go test ./internal/session/ -race -count=1` | ✅ | ✅ green |
| 7-05-03 | 05 | 4 | SWCH-01..04 | T-7-05 | wiring: toggle persist+reload composition; xdg-open no-pipes launcher | unit | `go test ./cmd/goswitchd/ -race -count=1` | ✅ | ✅ green |
| 7-06-01 | 06 | 5 | MACR-ACL, CORR | — | fires/password-silent/terminal-silent repinned on `apps_blocklist: []` | integration (live e2e) | `mise run e2e-matrix-v4` (rows; recorded in 07-RELEASE-READINESS) | ✅ | ✅ green |
| 7-06-02 | 06 | 5 | MACR-ACL | — | live blocklist negative: fired=0, one `ac_skip_app_blocked`, field verbatim | integration (live e2e) | `mise run e2e-autocorrect-blocklist-silent` (PASS @a264a54) | ✅ | ✅ green |
| 7-06-03 | 06 | 5 | SWCH-01..04 | — | menu-v2 Event-drive: toggles persist + survive daemon restart | integration (live e2e) | `mise run e2e-menu-v2` (PASS @a264a54) | ✅ | ✅ green |
| 7-07-01 | 07 | 6 | MACR-ACL | — | README v1.1.0 doc greps (README-OK / KEYS-MATCH / SOUND-DOC-OK / NO-TAG) | doc grep | RELEASE-READINESS checklist 5/5 PASS | ✅ | ✅ green |
| 7-07-02 | 07 | 6 | CORR, SWCH | — | full regression: ci ×2 green, matrix ×2 31/34 (honest class verdict), flip 24/24 | full suite + live e2e | `mise run ci`; saved reports e2e-report-07-07-run{1,2}.txt | ✅ | ✅ green |
| 7-07-03 | 07 | 6 | CORR, SWCH | — | RELEASE-READINESS artifact; no premature v1.1 tag | doc grep + git | `git tag --list "v1.1*"` → empty | ✅ | ✅ green |
| 7-08-01 | 08 | 5 | SPEC §11 | T-7-08 | flip tone on single flipTo path, observer-last, same-target silent | unit | `go test ./internal/session/ -run Sound -race -count=1` | ✅ | ✅ green |
| 7-08-02 | 08 | 5 | SPEC §11 | T-7-08 | Player: paplay fallback + RIFF tones, warn-once episodes, no-pipes | unit | `go test ./internal/sound/ -race -count=1` | ✅ | ✅ green |
| 7-08-03 | 08 | 5 | SPEC §11 | T-7-08 | fired-tone hook + sound.enabled gating via fold + hot event (WR-02) | unit | `go test ./cmd/goswitchd/ -run TestNewActor_SoundEventHotReload -v` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements (go test + mise tasks + golangci-lint pre-dated the phase; no stubs or framework install were needed).

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Menu v2 pixel-check (radio/checkmark ornaments, «Звук» placement, macro rows, no `_r` mnemonic swallow, old item gone) | Критерий 1 | GTK/Shell pixel rendering invisible to grep and to state-level e2e oracles | Open tray menu; verify ornaments, placement, live macro bindings vs config.yaml |
| Live sound hearing (flip tone; distinct autocorrect tone; «Звук» off = silence; player failure = WARN only, never blocks) | SPEC §11 sound | Audibility unrecordable on the stand by design (unit seam owns engine invocation) | Flip with Right Shift, fire autocorrection, toggle «Звук» off, repeat both |
| «Настройки…» editor-open (absent file → full defaults document created and opened; present → comments intact; broken save refuses loudly) | Критерий 1 | xdg-open per-desktop routing and visible editor window not observable programmatically | Click «Настройки…» without, then with, ~/.config/goswitch/config.yaml |
| Fresh-session D-48 formal gate + live re-run on post-review-fix tree (b708020..3c315b8 postdate saved live evidence @a264a54; unit delta covered, ci green at tip) | Критерий 4 | Requires owner's live GNOME session, fresh login ≤30 min | At verify-work: matrix v4 ×2 + flip-keystroke + menu-v2 + blocklist-silent on a fresh session |
| Doc-leg role-61 caveat owner decision (SPEC §11 :219 / ADR-007 :224 vs GTK3 role-61 live fact) | MACR-ACL residue (CR-01) | Owner-semantics wording decision on safety narration; code mitigation (role-ambiguous refusal) shipped and green | Add dated caveat (review variant-a) to SPEC §11/ADR-007/README, or accept as-is |

All five are routed to UAT / verify-work (07-VERIFICATION.md `human_verification`, 07-UAT.md); none is automatable on this stand.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references (none — existing infrastructure)
- [x] No watch-mode flags
- [x] Feedback latency < 7s (`mise run ci` measured 5.4 s)
- [x] `nyquist_compliant: true` set in frontmatter

Auditor verification (2026-10-05, HEAD 9534888): `mise run ci` re-run green (19 packages -race, 5.44 s); 13 `TestMenu*` pins; role-ambiguous fail-closed pair green; blocklist gate direction pins green; config schema/legacy-key and sound/writer corpora green; WR-01/02/03 fix tests green; saved matrix reports match claims (31/34 ×2 = 2 designed WINDOWS #12 + 1 migrating ibus transient each); WR-01-removal negative greps clean; SPEC §11 / ADR-007 amendment / menu-ID carriers present; no premature v1.1 tag. Zero coverage gaps found; the 5 human items are owner-facing and properly routed. **Nyquist verdict: compliant.**

**Approval:** approved 2026-10-05
