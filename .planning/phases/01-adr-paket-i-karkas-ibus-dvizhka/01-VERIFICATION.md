---
phase: 01-adr-paket-i-karkas-ibus-dvizhka
verified: 2026-09-16T19:20:46Z
status: passed
score: 20/20 must-haves verified
covered_files:
  - .github/dependabot.yml
  - .github/workflows/pr-sanity.yml
  - .github/workflows/security-scheduled.yml
  - .gitignore
  - .golangci.yml
  - .planning/codebase/CONVENTIONS.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-01-PLAN.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-01-SUMMARY.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-02-PLAN.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-02-SUMMARY.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-03-PLAN.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-03-SUMMARY.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-04-PLAN.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-04-SUMMARY.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-05-PLAN.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-05-SUMMARY.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-REVIEW.md
  - .planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-VALIDATION.md
  - AGENTS.md
  - README.md
  - cmd/goswitchd/main.go
  - dist/systemd/user/goswitchd.service
  - docs/SPEC.md
  - docs/adr/ADR-001-layout-switching-mechanism.md
  - docs/adr/ADR-002-tap-semantics.md
  - docs/adr/ADR-003-replacement-capability-ladder.md
  - docs/adr/ADR-004-buffer-reset-triggers.md
  - docs/adr/ADR-005-macr-01-owner-checkpoint.md
  - docs/adr/d01-experiment-log.md
  - engine/address.go
  - engine/address_test.go
  - engine/conn.go
  - engine/engine.go
  - engine/engine_test.go
  - engine/factory.go
  - engine/keys.go
  - engine/types.go
  - engine/wire_test.go
  - go.mod
  - go.sum
  - internal/hotkey/fsm.go
  - internal/hotkey/fsm_test.go
  - internal/logging/logging.go
  - internal/logging/logging_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - layouts/generator/main.go
  - layouts/tables.go
  - layouts/tables_test.go
  - mise.toml
  - test/e2e/README.md
  - test/e2e/case_d01.go
  - test/e2e/case_m1.go
  - test/e2e/case_resilience.go
  - test/e2e/focus_helper.py
  - test/e2e/main.go
  - test/e2e/preflight.go
covered_digest: "v1:sha256:1078838d233c7a6cdcfa2959ef01b06214f202a1d811203ac8e6270127c801f7"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 20/20
  gaps_closed: []
  gaps_remaining: []
  regressions: []
gaps: []
deferred:
  - truth: "Plan 01-05 key link: README.md documents the unit placement (cp dist/systemd/user/goswitchd.service → ~/.config/systemd/user/)"
    addressed_in: "Phase 4"
    evidence: "Phase 4 goal «Продукт ставится без root по документированной инструкции» + SC-1; Phase 4-01 INST-01 implemented goswitchctl install, which writes the same goswitchd.service into .config/systemd/user (internal/install/install.go:58-59, read-back verified); Phase 4-07 rewrote README to document that installer. The Phase 1 SC-level contract (user unit, no root, truth #12) is untouched; only the plan-level grep pattern (goswitchd\\.service in README) no longer matches — verify.key-links reports 13/14."
---

# Phase 01: ADR-пакет и каркас IBus-движка — Verification Report

**Phase Goal (ROADMAP.md):** «Все архитектурные развилки закрыты утверждёнными ADR, а `goswitchd` работает как IBus engine: видит клавиатурные события как активный input source, логирует их, переживает рестарт ibus-daemon и собственные паники, не конфликтуя с keyd/xremap — фундамент, на котором безопасно строить коррекцию и переключение.»

**Verified:** 2026-09-16T19:20:46Z
**Status:** passed
**Re-verification:** Yes — fingerprint refresh after Phase 4 modified shared covered files

## Why this re-verification ran

Phase 1 passed verification 2026-09-11 (20/20), UAT 4/4 (owner-accepted
2026-09-11, 01-UAT.md — final, not re-opened), and was re-verified twice
since (2026-09-15T03:55:00Z after Phase 2; 2026-09-15T17:57:51Z after
Phase 3, both 20/20). Phase 4 execution (54 commits d597a88..42d10a1 —
install/uninstall lifecycle, perf case, selfcheck, matrix v3, release
pipeline, bilingual README) has since legitimately modified covered files
again. Git classification of every change since the prior `verified:`
timestamp (2026-09-15T17:57:51Z), per file of the covered set:

- **Untouched — regression-check only** (0 commits): the entire engine/
  package (conn, factory, address, types, keys, engine.go + all four test
  files), internal/hotkey/*, internal/logging/*, layouts/*, docs/SPEC.md,
  docs/adr/* (ADR-001..005 + journal + 01-CONTEXT.md), dist/systemd/user/
  goswitchd.service, .github/workflows/{pr-sanity,security-scheduled}.yml,
  .github/dependabot.yml, .golangci.yml, AGENTS.md, test/e2e/README.md,
  all five Phase 1 PLAN/SUMMARY pairs, 01-REVIEW.md, 01-VALIDATION.md.
- **Phase 4 rework — deep re-check** (full 3-level + behavioral at HEAD
  42d10a1): internal/session/{actor.go,actor_test.go} + cmd/goswitchd/
  main.go (D-37 version identity: additive `version` field, `SetVersion`,
  `-version` flag exiting before `run()` — the FSM event path, the mutex
  serialization, the expiry `{"msg":"action","n":N}` logging and the
  `engine.Config.Handler` wiring are byte-for-byte preserved),
  test/e2e/{main.go,preflight.go,case_d01.go,case_m1.go,
  case_resilience.go,focus_helper.py} (additive case-spec registry,
  standalone-mode preflight split keeping all six Phase 1 named checks,
  additive surfaceKind switch branches, additive witness subcommands —
  every Phase 1 assertion intact), mise.toml (additive goreleaser pin +
  6 new tasks; Phase 1 [tools] pins and 6 core tasks unchanged),
  go.mod/go.sum (tidy clean, same 3 direct deps), README.md (bilingual
  rewrite — see Designed successions §9), .gitignore (additive).

New Phase 4 packages (internal/install, perf/selfcheck surfaces,
.goreleaser.yml, release workflow, matrix-v3 case file) carry no primary
Phase 1 contract weight and stay outside covered_files — conservative by
design, same policy as the Phase 3 pass.

The prior report had no `gaps:` section; its 20 truths, evidence and
live-gate policy carried over as the map.

## Designed successions (documented, not gaps)

Plan-level Phase 1 details superseded by later roadmap-sanctioned work.
The SC-level contracts all still hold on the current tree (items 1–8
carried from the prior pass and re-confirmed; item 9 is new this pass):

1. **ProcessKeyEvent consumption** delegated to the EventHandler
   (engine.go:148-171); MACR Super+letter consumption and the
   configurable-tap path transit the same handler verdict. Nil handler =
   pure observer by construction; declining handler transits everything
   (`TestEngine_ProcessKeyEventReturnsFalse`, re-run green this pass);
   EN mode is explicitly «Phase 1 semantics — transit»
   (`TestActor_ENTransitUnchanged`, re-run green).
2. **go.mod dependency set**: 3 direct deps — godbus v5.2.2, yaml.v3
   v3.0.1, fsnotify v1.10.1 — every one a core technology in the approved
   stack; +1 indirect (golang.org/x/sys). `go mod tidy` +
   `git diff --exit-code` zero drift this pass. Minimal-dep constraint
   holds (goreleaser is a mise `[tools]` pin, NOT a go.mod dependency).
3. **Actor construction wiring**: actor built from startup config
   (main.go:114-126), no-config default equals `hotkey.DefaultWindow`
   (300 ms, ADR-002, pinned by config.TestDefaults); actor remains
   `engine.Config.Handler` (main.go:152); D-37 adds only
   `actor.SetVersion(version)` (main.go:126).
4. **FSM constructor** `NewFSM(window, tapKeyval)` + hot-reload seams;
   purity preserved; Phase 1 corpus expectations untouched and green.
5. **HandleSurroundingText seam** with anchorPos; decode path and recover
   shim unchanged.
6. **MACR-01**: implemented by Phase 3 per Accepted ADR-005 (the
   roadmap-sanctioned next step of the ADR-005 decision).
7. **engine/keys.go** additive keyval constants, verified verbatim
   against ibuskeysyms.h.
8. **mise.toml** ~26 tasks total; Phase 1 contracts ([tools] go 1.23 +
   golangci-lint 2, tasks build/test/lint/vet/tidy-diff/ci, no Makefile)
   unchanged.
9. **NEW this pass — README install section**: plan 01-05's key link
   «README → dist/systemd/user/goswitchd.service» (dev-инструкция
   `cp` юнита + daemon-reload + start) no longer matches mechanically:
   the Phase 4-07 bilingual README rewrite documents the product install
   path `goswitchctl install` instead of the manual dev copy. The
   succession is roadmap-sanctioned — Phase 4's goal is literally
   «Продукт ставится без root по документированной инструкции», and
   Phase 4-01 (INST-01) implemented the installer that writes the SAME
   `goswitchd.service` into `.config/systemd/user`
   (internal/install/install.go:58-59, with read-back verification).
   The unit file itself is byte-untouched; README still states
   «no root anywhere: a systemd user unit plus a user-level IBus
   component» (README.md:14) and documents unit override via
   `systemctl --user edit goswitchd`. Verdict: SC-level truth #12
   intact; the plan-level link is recorded in frontmatter `deferred`
   (mechanical check: verify.key-links 13/14). If later re-verifications
   should keep this from re-tripping, an `overrides:` entry accepting the
   deviation is the appropriate mechanism — suggested, not applied
   (nothing failed at the truth level).

## Goal Achievement

### Observable Truths

| # | Truth | Status | Re-verification evidence (current tree, HEAD 42d10a1) |
|---|-------|--------|------------------------------------------|
| 1 | SC-1: ADR-001..005 exist in unified format (Status/Context/Decision/Consequences/Reversibility), CONTEXT decisions quoted | ✓ VERIFIED (regression) | Untouched since prior pass; re-checked: each of the 5 files carries 5/5 `## Status/Context/Decision/Consequences/Reversibility` headers, Status = «Accepted — гейт M0 … пройден 2026-09-10» |
| 2 | SC-1: D-01 journal — ≥4 verdict rows, decision derivable | ✓ VERIFIED (regression) | Untouched; `grep -c "Verdict\|Вердикт"` = **15** |
| 3 | SC-1: ADR-001 written last from journal — winner/Option B + kill-criteria, losers' observed failures, D-03 consequences | ✓ VERIFIED (regression) | File untouched (git: 0 commits since prior pass) |
| 4 | SC-1: M0 owner gate passed; MACR-01 fixed in ADR-005 (D-12), spec-deltas applied; no MACR-01 code in Phase 1 | ✓ VERIFIED (regression + succession) | All 5 ADRs Accepted, SPEC.md untouched — the owner-gate contract intact. Phase 3 shipped MACR code per Accepted ADR-005 (designed succession §6); Phase 4 added none; the truth held at phase close and its decision-record contract still holds |
| 5 | SC-2: programmatic registration on the private IBus bus (no XML, no root), two engines goswitch-en/-ru (INTEG-01) | ✓ VERIFIED | engine/conn.go + factory.go byte-unchanged since their live proof; **fresh machine-generated live evidence**: e2e-matrix CI run 35108412175 (2026-09-16T14:25:22Z, self-hosted GNOME, head 9f2fd75 — an ancestor of HEAD) — **31/31 PASS** (incl. the D-48 double-run gate) with a fresh daemon per case; registration + activation re-proven on the Phase 4 lineage; TestAddress_* green in this pass's suite |
| 6 | SC-2: as active source, engine receives ProcessKeyEvent, logs keys; normal typing transits | ✓ VERIFIED (deep) | engine.go untouched this cycle: decode → slog.Debug("key",…) (engine.go:160, still DEBUG-only) → handler verdict, nil → false (engine.go:156-171); TestEngine_ProcessKeyEventReturnsFalse, TestActor_ENTransitUnchanged, TestActor_RUTransitUnmapped all green at HEAD in this pass's `-race` run |
| 7 | SC-2: e2e skeleton — ydotool injection, self-activation, fail-fast preflight, exit-code contract (TEST-02/03) | ✓ VERIFIED (deep) | main.go: `os.Exit(run())` (line 98), FAIL paths print + non-zero; preflight keeps all 6 Phase 1 named checks (injection-selftest, ibus-address, uinput-writable, python-gi always; engine-registered, daemon-log-heartbeat + surface checks for stand mode — the new standalone split affects only the install-cycle case); focus_helper.py `focus` contract intact with additive witness subcommands (/usr/bin/python3 + AT-SPI grabFocus unchanged); m1 assertions intact (`"msg":"key"` ≥ min case_m1.go:63, `"msg":"action","n":2` case_m1.go:70, registered-before-focus_in by timestamps) |
| 8 | SC-3: ibus restart → re-registration on the new socket, keys flow again (INTEG-04) | ✓ VERIFIED | conn.go (reconnect implementation) byte-unchanged since the committed live PASS; case_resilience.go re-registered wait intact (line 34) + respawn `registrations+1` (lines 86, 129); mise task e2e-ibus-restart present (mise.toml:45) |
| 9 | SC-3: recover shim contains handler panics (INTEG-05) | ✓ VERIFIED (deep) | Re-counted at HEAD: `defer recoverHandler` on **21/21** exported Engine D-Bus handlers (engine.go — still 25 exported methods = 21 handlers + 4 outgoing emitters; none added) + CreateEngine (factory.go:30); **TestEngine_RecoverContainsPanic re-run green under -race at HEAD this pass** |
| 10 | SC-3: kill -9 → desktop input alive, daemon respawns and re-registers | ✓ VERIFIED | case_resilience.go kill9 path intact (SIGKILL → plain-engine switch → inject → desktop oracle → respawnAndWait `component registered` count+1, lines 60-130; additive surfaceKind branches only); dist unit `Restart=on-failure` unchanged; committed live PASS stands |
| 11 | SC-4: no EVIOCGRAB / evdev / uinput capture in product code | ✓ VERIFIED (deep, widened) | Structural grep over engine/ cmd/ internal/ layouts/ — including Phase 3's and Phase 4's internal/{appid,clipboard,config,correct,ctlsvc,install} and cmd/goswitchctl — **CLEAN**: no EVIOCGRAB, no /dev/uinput, no /dev/input, no evdev access, no OpenFile/ioctl in product code; the only "evdev" matches are comments describing physical keycode semantics (actor.go:874,1148,1159) |
| 12 | SC-4: works over default Ubuntu 24.04 GNOME Wayland IM stack without root (INTEG-03) | ✓ VERIFIED (regression) | systemd **user** unit byte-unchanged (PartOf=graphical-session.target, Restart=on-failure — no root anywhere); conn.go EXTERNAL/SO_PEERCRED auth unchanged; README:14 still documents no-root; mise user-local |
| 13 | SC-4: keyd/xremap non-conflict (INTEG-02) | ✓ VERIFIED | No-capture gate (truth 11) + transit (truth 6) + e2e README INTEG-02 owner checklist intact (test/e2e/README.md — untouched); live keyd session was owner-UAT-accepted 2026-09-11 |
| 14 | SC-5: golden tests of generated tables incl. `[ ] ; ' , . /` + asymmetric pairs (CORR-08) | ✓ VERIFIED (regression) | layouts/ untouched; TestGolden_* green in this pass's suite at HEAD |
| 15 | SC-5: tables generated by go:generate, deterministic, CI xkb-free | ✓ VERIFIED (re-run) | **Re-run this pass: `go generate ./layouts && git diff --exit-code` → byte-identical**; DO-NOT-EDIT marker + go:generate header intact |
| 16 | SC-5: FSM unit corpus on synthetic streams (TEST-01) | ✓ VERIFIED (deep) | fsm.go untouched this cycle — purity re-checked clean (no goroutines/chans/time.Now/Sleep); TestFSM_* corpus green under -race at HEAD |
| 17 | SC-5: structured logs with levels; key trace only behind -debug (INST-04) | ✓ VERIFIED | logging.go untouched (LevelVar INFO/DEBUG); key trace still slog.Debug-only (engine.go:160; surrounding_text/capabilities DEBUG engine.go:192,237; lifecycle INFO engine.go:248,257); TestLogging_* green; -debug password warning in main.go:26 (untouched); README privacy section in the bilingual pair |
| 18 | SC-5: headless CI gates green (build/vet/lint/test -race/tidy-diff) | ✓ VERIFIED (re-run) | **Verifier re-ran all gates this pass at HEAD 42d10a1**: `go build ./...` OK; `go vet ./...` OK; `go test -race -count=1 ./...` — 13 test-bearing packages ok, 0 failures; `mise run lint` → 0 issues; `go mod tidy` + `git diff --exit-code` → zero drift |
| 19 | Plan 01-05: CI triplet (pr-sanity + dependabot gomod weekly + scheduled govulncheck), e2e excluded from CI | ✓ VERIFIED (prior caveat CLOSED) | All three workflow/config files untouched; **pr-sanity now GREEN on this content lineage**: run 35130001841 (push to main 80d8534, 2026-09-16T17:44Z) + PR runs 35129888432/34914315850 — the prior pass's «no pr-sanity run yet» caveat is resolved; security-scheduled cron green 2026-09-14 (run 34838604697, weekly cadence); dependabot dynamic runs green; e2e remains CI-excluded except the dispatch-only e2e-matrix workflow (newest run 35108412175 success) |
| 20 | Plan 01-03: session actor serializes FSM, logs `{"msg":"action","n":N}` at expiry, wired into daemon | ✓ VERIFIED (deep) | Wiring intact through the Phase 4 D-37 addition: `session.NewActor(window)` at cmd/goswitchd/main.go:125, actor = `engine.Config.Handler` at main.go:152; HandleKey holds `a.mu` across the whole event (unchanged); ExpiryAt still emits `slog.Info("action", "n", int(action))`; the only actor.go delta is the additive version field/SetVersion/Status.Version (D-37); **TestActor_Serialization + TestActor_ENTransitUnchanged re-run green under -race at HEAD** |

**Score:** 20/20 truths verified (0 present-but-behavior-unverified)

Behavior-dependent truths note: panic containment (#9), FSM window
semantics (#16), actor expiry/timer/serialization (#20), transit contract
(#6) are all covered by behavioral tests that ran green in this pass's
`go test -race ./...` at HEAD (the named tests additionally re-run
individually). Live-bus truths (#5, #8, #10) rest on committed live-run
evidence per the phase's binding evidence policy — #5 and the whole
key-flow/correction chain additionally re-proven live by the 31/31
e2e-matrix CI run 35108412175 (2026-09-16, head 9f2fd75, an ancestor of
HEAD). The prior pass's recency caveat is now CLOSED: the only code file
changed between 9f2fd75 and HEAD is cmd/goswitchctl/main_test.go
(test-only corpus, commit 100af04) — every product-code path at HEAD is
byte-identical to the live-proven run head.

### Advisory (New Scope, Unevidenced)

Re-verification ran; the Step 7 scan over all covered files found no new
blocking pattern. Advisory entries: none.

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| 1 | None — only the informational notes below (lint tooling notice; doc wording; decision-coverage tool note) | — | — |

### Required Artifacts

`gsd-tools verify artifacts` per plan, re-run this pass: **31/31 passed,
0 issues** — 01-01 (8/8), 01-02 (5/5), 01-03 (6/6), 01-04 (7/7), 01-05
(5/5).

### Key Link Verification

`gsd-tools verify key-links` per plan, re-run this pass: **13/14
verified** across all five plans — including main.go→engine.Run
(cmd/goswitchd/main.go:74), e2e→daemon subprocess, ADR-001→journal,
ADR-004→SPEC delta, pr-sanity→mise.toml, AGENTS.md→CONVENTIONS.md
regeneration chain.

The one non-verified link is **README.md → dist/systemd/user/
goswitchd.service** (plan 01-05; pattern `goswitchd\.service` no longer
present in README.md): a designed succession — see Designed successions
§9 and frontmatter `deferred`. The unit file is untouched; the README's
documented installer writes the same unit; the SC-level no-root contract
(truth #12) holds. Not a truth failure; recorded honestly rather than
absorbed silently.

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| engine/conn.go (unchanged) | addr | `Discover()` per cycle: $IBUS_ADDRESS → suffix-filtered freshest ~/.config/ibus/bus file | ✓ real socket discovery (TestAddress_* green; live-registered in the 31/31 matrix run) | ✓ FLOWING |
| engine/engine.go (unchanged this cycle) | ev / consume | decodeEvent of ProcessKeyEvent args; verdict from EventHandler | ✓ matrix cases prove keys flow and corrections/MACR decisions act on the live bus | ✓ FLOWING |
| internal/session/actor.go (D-37 additive) | action n / corrections / snapshot / version | FSM Feed at timer expiry; one config Snapshot per event; version pinned at construction | ✓ TestActor_* green at HEAD (now incl. TestStatusCarriesVersion) + live matrix corrections | ✓ FLOWING |
| layouts/tables.go (unchanged) | ENToRU/RUToEN | generated from system xkb | ✓ regeneration byte-identical this pass | ✓ FLOWING |
| docs/adr/d01-experiment-log.md (unchanged) | verdict rows | appended by d01-probe live runs | ✓ 15 rows; append path intact in case_d01.go (additive branches only) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Build | `go build ./...` | exit 0 | ✓ PASS |
| Vet | `go vet ./...` | exit 0 | ✓ PASS |
| Full race suite at HEAD (single run) | `go test -race -count=1 ./...` | 13 test-bearing packages ok, 0 failures (incl. new internal/install, test/e2e) | ✓ PASS |
| Named: panic containment | `go test -race -run TestEngine_RecoverContainsPanic ./engine/` | PASS | ✓ PASS |
| Named: transit + serialization | `go test -race -run 'TestActor_ENTransitUnchanged\|TestActor_Serialization\|TestActor_RUTransitUnmapped' ./internal/session/` | all PASS | ✓ PASS |
| Named: FSM corpus + golden tables + logging gates | `go test -race -run 'TestFSM' ./internal/hotkey/`, `TestGolden` ./layouts/, `TestLogging` ./internal/logging/ | all PASS | ✓ PASS |
| Lint (strict v2 config) | `mise run lint` | 0 issues | ✓ PASS |
| Tidy drift | `go mod tidy` + `git diff --exit-code -- go.mod go.sum` | zero drift | ✓ PASS |
| Generation determinism | `go generate ./layouts && git diff --exit-code -- layouts/` | byte-identical | ✓ PASS |
| Recover-shim count | `grep -c "defer recoverHandler" engine/engine.go` (21) + factory.go (1) | 21 handlers + CreateEngine; 25 exported methods, 0 new since prior pass | ✓ PASS |
| No device capture (widened) | `grep -rniE "EVIOCGRAB\|/dev/uinput\|/dev/input\|evdev" engine/ cmd/ internal/ layouts/` | comment-only matches (keycode semantics); no access | ✓ PASS |
| FSM purity | `grep -nE "go func\|chan \|time.Now\|Sleep" internal/hotkey/fsm.go` | no matches | ✓ PASS |
| Journal verdict count | `grep -c "Verdict\|Вердикт" docs/adr/d01-experiment-log.md` | 15 (≥4) | ✓ PASS |
| Daemon wiring | grep of cmd/goswitchd/main.go | engine.Run (74), session.NewActor (125), Handler: actor (152) — plus additive SetVersion (126) | ✓ PASS |
| CI green | `gh run list` | e2e-matrix 35108412175 success (head 9f2fd75, 2026-09-16T14:25Z); pr-sanity 35130001841 success (main, 2026-09-16T17:44Z); security-scheduled 34838604697 success (cron 2026-09-14); release 35130309174 success | ✓ PASS |
| Live e2e matrix artifact | `gh run download 35108412175 -n e2e-report` | **31/31 PASS** summary (incl. gedit/x11 families; D-48 double-run gate), 2026-09-16 | ✓ PASS |
| `-version` flag behavior | `go run ./cmd/goswitchd -version` | prints `goswitchd dev`, exits before run() | ✓ PASS |
| Live m1/ibus-restart/kill9/d01 mise tasks | not re-run (drive owner's desktop) | committed Phase 1 evidence + intact assertions + unchanged conn.go + fresh 31/31 matrix coverage of the same daemon paths | ? SKIP (by policy) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` conventions in this repo. The runnable probes
are the mise tasks: headless CI gates re-run this pass (PASS, table above);
live-session e2e tasks excluded by the binding evidence policy — their
machine-generated CI counterpart (e2e-matrix v3 with the D-48 double-run
gate, dispatch-only) was fetched and is green 31/31 at head 9f2fd75
(2026-09-16).

### Requirements Coverage

All 10 Phase 1 requirement IDs remain mapped to Phase 1 and Complete in
REQUIREMENTS.md traceability (re-checked this pass); no orphans.

| Requirement | Source Plan | Status | Evidence |
|-------------|------------|--------|----------|
| CORR-08 | 01-02 | ✓ SATISFIED | truths #14-15 |
| INTEG-01 | 01-01 | ✓ SATISFIED | truth #5; emitters live-driven (31/31 matrix run 2026-09-16) |
| INTEG-02 | 01-01, 01-03 | ✓ SATISFIED | truths #6, #11, #13; live keyd owner-UAT-accepted |
| INTEG-03 | 01-01, 01-03, 01-05 | ✓ SATISFIED | truths #11-12; CI live on GitHub runners + pr-sanity green on main |
| INTEG-04 | 01-01, 01-03, 01-04 | ✓ SATISFIED | truth #8 |
| INTEG-05 | 01-01, 01-03 | ✓ SATISFIED | truths #9-10 |
| TEST-01 | 01-02, 01-03, 01-05 | ✓ SATISFIED | truths #16-18 |
| TEST-02 | 01-03, 01-04 | ✓ SATISFIED | truth #7; stand re-proven live by the 31/31 matrix run |
| TEST-03 | 01-03 | ✓ SATISFIED | truth #7 (focus_helper.py focus contract intact) |
| INST-04 | 01-01, 01-05 | ✓ SATISFIED | truths #17, #19 |

### Decision Coverage

The `check decision-coverage-verify` gate re-run reports `skipped:
CONTEXT.md missing` under the current gsd-tools resolution (warning-only,
non-blocking). The phase's 01-CONTEXT.md is byte-untouched since the
prior pass's 12/12 honored verdict — the underlying decision set has not
changed; only the tool's lookup path differs in this version.

### Anti-Patterns Found

Debt-marker gate re-run over all covered files touched since the prior
pass: **ZERO TBD/FIXME/XXX, zero TODO/HACK/PLACEHOLDER, zero stub
patterns**. Carry-forward warnings from the prior pass (files unchanged,
still non-blocking quality debt): boot-time retry race semantics
(conn.go:119-123), NameFlagReplaceExisting inertness (conn.go:88-101),
length-only oracle limitation in locked-session d01 re-runs.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| engine/conn.go | 93, 122 | carry-forward boot-time retry race + rapid-respawn race (unchanged since prior passes) | ⚠️ Warning | Canonical paths demonstrably green (31/31 live matrix 2026-09-16); documented quality debt for hardening |
| .golangci.yml → lint output | — | golangci-lint reports `exhaustruct` deprecated since v2.13.0 (replaced by exhaustruct_v5) — informational tooling notice, config still applies, 0 issues | ℹ️ Info | Cosmetic future config rename; no code impact |
| test/e2e/README.md | 201 | INTEG-02 section still carries the Phase 1 wording «`ProcessKeyEvent`, возвращающий `false` (наблюдатель)» — predates the consumption succession; the architectural claim (no evdev layer, keyd below IBus) remains true | ℹ️ Info | Documentation nuance only; a wording refresh matching the conditional-consumption model would help future readers |
| README.md | install section | documents `goswitchctl install` instead of the plan-01-05 manual unit-copy dev instruction | ℹ️ Info | Designed succession (see §9 + `deferred`); SC-level no-root contract intact |

### Human Verification Required

None new. The four environment-gated items from the original pass were
resolved by UAT (01-UAT.md: 4/4 pass, owner-accepted 2026-09-11) and are
not re-opened. No ⚠️ PRESENT_BEHAVIOR_UNVERIFIED truths exist on this pass.

### Gaps Summary

**No gaps.** All 20 must-have truths verified on the current
(post-Phase-4) tree at HEAD 42d10a1; 31/31 artifacts pass; 13/14 key
links wired (the 14th is a roadmap-sanctioned designed succession,
recorded in `deferred` — the SC-level contract it served holds); 
requirements 10/10 with zero orphans; zero debt markers; all headless
gates re-run green by the verifier at HEAD (build, vet, 13-package race
suite, strict lint, tidy, generation determinism); fresh
machine-generated live evidence (e2e-matrix 31/31 on the self-hosted
GNOME runner, 2026-09-16, head 9f2fd75 — with only a test file changed
between that head and HEAD) covers the registration/key-flow/transit
chain on the Phase 4 lineage, and the prior pass's pr-sanity caveat is
closed by green runs on main and PRs. Phase 4's additive rework of shared
files (actor version identity, daemon -version flag, e2e case registry,
preflight split, witness helpers, bilingual README) preserved every
SC-level Phase 1 contract — recover-shim coverage, no-device-capture over
the widened package set, DEBUG-gated key trace, actor serialization and
action logging, deterministic tables, the CI triplet — with nine designed
successions documented above.

---

_Verified: 2026-09-16T19:20:46Z_
_Verifier: Claude (gsd-verifier)_
