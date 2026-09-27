---
phase: 01-adr-paket-i-karkas-ibus-dvizhka
verified: 2026-09-20T14:32:52Z
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
covered_digest: "v1:sha256:96547403c207079795be546a310235b822e9bf577e56a1a1ba0e4cda5ca905b0"
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
    evidence: "Phase 4 goal «Продукт ставится без root по документированной инструкции» + SC-1; Phase 4-01 INST-01 implemented goswitchctl install, which writes the same goswitchd.service into .config/systemd/user (internal/install/install.go:58-59, read-back verified); Phase 4-07 rewrote README to document that installer. The Phase 1 SC-level contract (user unit, no root, truth #12) is untouched; only the plan-level grep pattern (goswitchd\\.service in README) no longer matches — verify.key-links re-run this pass: 13/14, same single designed succession."
---

# Phase 01: ADR-пакет и каркас IBus-движка — Verification Report

**Phase Goal (ROADMAP.md):** «Все архитектурные развилки закрыты утверждёнными ADR, а `goswitchd` работает как IBus engine: видит клавиатурные события как активный input source, логирует их, переживает рестарт ibus-daemon и собственные паники, не конфликтуя с keyd/xremap — фундамент, на котором безопасно строить коррекцию и переключение.»

**Verified:** 2026-09-20T14:32:52Z
**Status:** passed
**Re-verification:** Yes — fingerprint refresh after Phase 4 gap-closure (04-08/04-09) modified a shared covered file

## Why this re-verification ran

Phase 1 passed verification 2026-09-11 (20/20), UAT 4/4 (owner-accepted
2026-09-11, 01-UAT.md — final, not re-opened), and was re-verified three
times since (2026-09-15T03:55:00Z after Phase 2; 2026-09-15T17:57:51Z
after Phase 3; 2026-09-16T19:20:46Z after Phase 4 mainline — all 20/20).
Phase 4's gap-closure round (G-4-1 witness probe budget → plan 04-08;
G-4-2 human-free acceptance pipeline → plan 04-09; commits 994ee2a..
04d00ad, 2026-09-17..18) has since legitimately modified one covered file
again: **test/e2e/focus_helper.py** (commits 4bf01fc, cafd4ba).

Git classification of every change since the prior `verified:` timestamp
(2026-09-16T19:20:46Z), per file of the covered set:

- **Exactly one covered file changed — deep re-check at HEAD 04d00ad:**
  test/e2e/focus_helper.py (+143/−13). Fully additive: the pre-04-08
  `cmd_witness` body is preserved **verbatim** as `full_walk_witness()`
  (the locked-session PASSWORD_TEXT arm included) and retained as the
  fallback when the new focus-first pass finds nothing; a new
  `focused-app` subcommand (04-09 idle-desktop preflight primitive) is
  added at the dispatch level; the **`focus <app>` subcommand — the
  Phase 1 TEST-03 contract — is untouched** (`cmd_focus`, grabFocus on
  the app's newest window, lines 504-520; only the usage line gained the
  new mode name). Phase 1 semantics survive byte-identically.
- **Untouched — regression-check only** (0 commits since prior pass):
  the entire engine/ package, internal/hotkey/*, internal/logging/*,
  internal/session/*, layouts/*, cmd/goswitchd/main.go, docs/SPEC.md,
  docs/adr/*, dist/systemd/user/goswitchd.service,
  .github/workflows/{pr-sanity,security-scheduled}.yml,
  .github/dependabot.yml, .golangci.yml, AGENTS.md, README.md, go.mod,
  go.sum, mise.toml, .gitignore, test/e2e/{README.md,main.go,preflight.go,
  case_d01.go,case_m1.go,case_resilience.go}, all five Phase 1
  PLAN/SUMMARY pairs, 01-REVIEW.md, 01-VALIDATION.md, 01-UAT.md.

New Phase 4 gap-closure files (test/e2e/quiesce_test.go, test/e2e/matrix.go
edits, scripts/d48-nightly-dispatch.sh, .github/workflows/e2e-matrix.yml
nightly arms, docs/ci-runner.md) carry no primary Phase 1 contract weight
and stay outside covered_files — conservative by design, same policy as
the two prior passes.

**Live-bus evidence status this pass (checked, not assumed):** three
e2e-matrix dispatch runs after the prior pass FAILED (35142914380,
35188570260, 35271719601 — all on head 9ff56cb, 2026-09-16..17). Verified
not a Phase 1 regression: every one of the 31 cases failed in Phase 4's
own **desktop a11y quiesce preflight** («a11y witness not answering within
30s … signal: killed») — the harness hung before any matrix case started,
so the daemon's registration/key-flow paths were never exercised. Phase 4
itself root-caused and formalized this (commit 8fe5a08 «environmental a11y
wedge», 4ed89c5 «witness full-tree walk exceeds probe budget» → gaps
G-4-1/G-4-2 → closure plans 04-08/04-09 at HEAD). The committed live proof
of the Phase 1 chain stands: e2e-matrix run 35108412175 (2026-09-16T14:25Z,
head 9f2fd75, **31/31 PASS**) — and `git diff 9f2fd75..HEAD -- engine/
internal/ cmd/ layouts/` shows exactly ONE changed product file,
cmd/goswitchctl/main_test.go (test-only corpus): **every Phase 1 daemon
path is byte-identical to the live-proven head.** A post-fix green matrix
run is Phase 4's open D-48 acceptance item, not a Phase 1 truth.

The prior report had no `gaps:` section; its 20 truths, evidence and
live-gate policy carried over as the map.

## Designed successions (documented, not gaps)

Plan-level Phase 1 details superseded by later roadmap-sanctioned work.
The SC-level contracts all still hold on the current tree (items 1–9
carried from the prior pass, re-confirmed; nothing new this pass):

1. **ProcessKeyEvent consumption** delegated to the EventHandler
   (engine.go:148-171); MACR Super+letter consumption and the
   configurable-tap path transit the same handler verdict. Nil handler =
   pure observer by construction; declining handler transits everything
   (`TestEngine_ProcessKeyEventReturnsFalse`, green in this pass's suite);
   EN mode is explicitly «Phase 1 semantics — transit»
   (`TestActor_ENTransitUnchanged`, green).
2. **go.mod dependency set**: 3 direct deps — godbus v5.2.2, yaml.v3
   v3.0.1, fsnotify v1.10.1 — every one a core technology in the approved
   stack; +1 indirect (golang.org/x/sys). `go mod tidy` +
   `git diff --exit-code` zero drift this pass. Minimal-dep constraint
   holds (goreleaser is a mise `[tools]` pin, NOT a go.mod dependency).
3. **Actor construction wiring**: actor built from startup config
   (main.go:114-126), no-config default equals `hotkey.DefaultWindow`
   (300 ms, ADR-002, pinned by config.TestDefaults); actor remains
   `engine.Config.Handler` (main.go:152); D-37's `actor.SetVersion(version)`
   (main.go:126) unchanged since.
4. **FSM constructor** `NewFSM(window, tapKeyval)` + hot-reload seams;
   purity preserved (re-grepped clean this pass); Phase 1 corpus
   expectations green.
5. **HandleSurroundingText seam** with anchorPos; decode path and recover
   shim unchanged.
6. **MACR-01**: implemented by Phase 3 per Accepted ADR-005 (the
   roadmap-sanctioned next step of the ADR-005 decision).
7. **engine/keys.go** additive keyval constants, verified verbatim
   against ibuskeysyms.h.
8. **mise.toml**: Phase 1 contracts ([tools] go 1.23 + golangci-lint 2,
   tasks build/test/lint/vet/tidy-diff/ci, no Makefile) unchanged.
9. **README install section**: plan 01-05's key link «README →
   dist/systemd/user/goswitchd.service» still does not match mechanically
   (Phase 4-07 bilingual README documents `goswitchctl install` instead of
   the manual dev copy). Roadmap-sanctioned — Phase 4-01 (INST-01) writes
   the SAME unit into `.config/systemd/user`
   (internal/install/install.go:58-59, read-back verified); the unit file
   is byte-untouched; README:14 still states «no root anywhere». SC-level
   truth #12 intact; recorded in frontmatter `deferred` (mechanical check
   re-run this pass: verify.key-links 13/14).

## Goal Achievement

### Observable Truths

| # | Truth | Status | Re-verification evidence (current tree, HEAD 04d00ad) |
|---|-------|--------|------------------------------------------|
| 1 | SC-1: ADR-001..005 exist in unified format (Status/Context/Decision/Consequences/Reversibility), CONTEXT decisions quoted | ✓ VERIFIED (regression) | Untouched since prior pass; re-counted this pass: each of the 5 files carries 5/5 `## Status/Context/Decision/Consequences/Reversibility` headers; ADR-001 Status = «Accepted — гейт M0 … пройден 2026-09-10» |
| 2 | SC-1: D-01 journal — ≥4 verdict rows, decision derivable | ✓ VERIFIED (regression) | Untouched; `grep -c "Verdict\|Вердикт"` = **15** (re-run this pass) |
| 3 | SC-1: ADR-001 written last from journal — winner/Option B + kill-criteria, losers' observed failures, D-03 consequences | ✓ VERIFIED (regression) | File untouched (git: 0 commits since prior pass) |
| 4 | SC-1: M0 owner gate passed; MACR-01 fixed in ADR-005 (D-12), spec-deltas applied; no MACR-01 code in Phase 1 | ✓ VERIFIED (regression + succession) | All 5 ADRs Accepted, SPEC.md untouched — the owner-gate contract intact. Phase 3 shipped MACR code per Accepted ADR-005 (designed succession §6); the decision-record contract holds |
| 5 | SC-2: programmatic registration on the private IBus bus (no XML, no root), two engines goswitch-en/-ru (INTEG-01) | ✓ VERIFIED | engine/conn.go + factory.go byte-unchanged since their live proof; committed live evidence stands: e2e-matrix run 35108412175 (2026-09-16T14:25:22Z, self-hosted GNOME, head 9f2fd75) — **31/31 PASS** — and `git diff 9f2fd75..HEAD -- engine/ internal/ cmd/ layouts/` = only cmd/goswitchctl/main_test.go (test corpus), i.e. the daemon's registration path at HEAD is byte-identical to the live-proven head; TestAddress_* green in this pass's `-race` suite. The three post-pass matrix failures are Phase 4's quiesce-preflight harness gaps (G-4-1/G-4-2, closed by 04-08/04-09) — daemon never exercised in them (see Why-this-ran) |
| 6 | SC-2: as active source, engine receives ProcessKeyEvent, logs keys; normal typing transits | ✓ VERIFIED | engine.go untouched this cycle: decode → slog.Debug("key",…) (engine.go:160, still DEBUG-only) → handler verdict, nil → false (engine.go:156-171); TestEngine_ProcessKeyEventReturnsFalse, TestActor_ENTransitUnchanged, TestActor_RUTransitUnmapped all green at HEAD in this pass's `-race` run; key-flow chain also live-proven at the byte-identical product head (run 35108412175) |
| 7 | SC-2: e2e skeleton — ydotool injection, self-activation, fail-fast preflight, exit-code contract (TEST-02/03) | ✓ VERIFIED (deep) | main.go: `os.Exit(run())` (line 98), FAIL paths print + non-zero; preflight keeps all 6 Phase 1 named checks; focus_helper.py — the ONE covered file changed this pass — is additive (verbatim full-walk fallback + new focused-app mode; **`focus <app>` grabFocus contract untouched**, cmd_focus lines 504-520); m1 assertions intact (`"msg":"key"` ≥ min case_m1.go:63, `"msg":"action","n":2` case_m1.go:70, registered-before-focus_in by timestamps); test/e2e package (incl. new quiesce corpus) ok under -race at HEAD |
| 8 | SC-3: ibus restart → re-registration on the new socket, keys flow again (INTEG-04) | ✓ VERIFIED | conn.go (reconnect implementation) byte-unchanged since the committed live PASS; case_resilience.go re-registered wait intact (line 34) + respawn `registrations+1` (lines 86, 129); mise task e2e-ibus-restart present (mise.toml:45) |
| 9 | SC-3: recover shim contains handler panics (INTEG-05) | ✓ VERIFIED (deep) | Re-counted at HEAD this pass: `defer recoverHandler` on **21/21** exported Engine D-Bus handlers (engine.go — 25 exported methods = 21 handlers + 4 outgoing emitters; none added) + CreateEngine (factory.go:30); **TestEngine_RecoverContainsPanic green in this pass's -race suite** |
| 10 | SC-3: kill -9 → desktop input alive, daemon respawns and re-registers | ✓ VERIFIED | case_resilience.go kill9 path intact (SIGKILL → plain-engine switch → inject → desktop oracle → respawnAndWait `component registered` count+1, lines 60-130); dist unit `Restart=on-failure` unchanged; committed live PASS stands; product head byte-identical to proven head |
| 11 | SC-4: no EVIOCGRAB / evdev / uinput capture in product code | ✓ VERIFIED (deep, widened) | Structural grep re-run over engine/ cmd/ internal/ layouts/ — including internal/{appid,clipboard,config,correct,ctlsvc,install} and cmd/goswitchctl — **CLEAN**: no EVIOCGRAB, no /dev/uinput, no /dev/input, no evdev access, no OpenFile/ioctl in product code; the only "evdev" matches are comments describing physical keycode semantics (actor.go:874,1148,1159) |
| 12 | SC-4: works over default Ubuntu 24.04 GNOME Wayland IM stack without root (INTEG-03) | ✓ VERIFIED (regression) | systemd **user** unit byte-unchanged (PartOf=graphical-session.target, Restart=on-failure — re-grepped this pass; no root anywhere); conn.go EXTERNAL/SO_PEERCRED auth unchanged; README:14 still documents no-root; mise user-local |
| 13 | SC-4: keyd/xremap non-conflict (INTEG-02) | ✓ VERIFIED | No-capture gate (truth 11) + transit (truth 6) + e2e README INTEG-02 owner checklist intact (test/e2e/README.md — untouched); live keyd session was owner-UAT-accepted 2026-09-11 |
| 14 | SC-5: golden tests of generated tables incl. `[ ] ; ' , . /` + asymmetric pairs (CORR-08) | ✓ VERIFIED (regression) | layouts/ untouched; TestGolden_* green in this pass's suite at HEAD |
| 15 | SC-5: tables generated by go:generate, deterministic, CI xkb-free | ✓ VERIFIED (re-run) | **Re-run this pass: `go generate ./layouts && git diff --exit-code` → byte-identical**; DO-NOT-EDIT marker + go:generate header intact |
| 16 | SC-5: FSM unit corpus on synthetic streams (TEST-01) | ✓ VERIFIED (deep) | fsm.go untouched this cycle — purity re-grepped clean (no goroutines/chans/time.Now/Sleep); TestFSM_* corpus green under -race at HEAD |
| 17 | SC-5: structured logs with levels; key trace only behind -debug (INST-04) | ✓ VERIFIED | logging.go untouched (LevelVar INFO/DEBUG); key trace still slog.Debug-only (engine.go:160 re-grepped); TestLogging_* green; -debug password warning in main.go:26 (untouched); README privacy section in the bilingual pair |
| 18 | SC-5: headless CI gates green (build/vet/lint/test -race/tidy-diff) | ✓ VERIFIED (re-run) | **Verifier re-ran all gates this pass at HEAD 04d00ad**: `go build ./...` OK; `go vet ./...` OK; `go test -race -count=1 ./...` — 14 test-bearing packages ok, 0 failures (incl. test/e2e with the new quiesce corpus); `mise run lint` → 0 issues; `go mod tidy` + `git diff --exit-code` → zero drift |
| 19 | Plan 01-05: CI triplet (pr-sanity + dependabot gomod weekly + scheduled govulncheck), e2e excluded from CI | ✓ VERIFIED | All three workflow/config files untouched; pr-sanity green on main (run 35130001841, 2026-09-16T17:44Z) and PRs; security-scheduled cron green 2026-09-14 (run 34838604697, weekly cadence); dependabot dynamic runs green (latest 35406821794, 2026-09-18); e2e remains CI-excluded except the dispatch-only e2e-matrix workflow — whose three recent failures are Phase 4's harness gaps (G-4-1/G-4-2, closed at HEAD; Phase 4's D-48 acceptance to re-prove green post-fix) |
| 20 | Plan 01-03: session actor serializes FSM, logs `{"msg":"action","n":N}` at expiry, wired into daemon | ✓ VERIFIED (deep) | Wiring re-grepped at HEAD: `session.NewActor(window)` at cmd/goswitchd/main.go:125, actor = `engine.Config.Handler` at main.go:152, engine.Run at main.go:74, additive SetVersion at 126; HandleKey holds `a.mu` across the whole event; ExpiryAt still emits `slog.Info("action", "n", int(action))`; **TestActor_Serialization + TestActor_ENTransitUnchanged + TestActor_RUTransitUnmapped green under -race at HEAD** |

**Score:** 20/20 truths verified (0 present-but-behavior-unverified)

Behavior-dependent truths note: panic containment (#9), FSM window
semantics (#16), actor expiry/timer/serialization (#20), transit contract
(#6) are all covered by behavioral tests that ran green in this pass's
`go test -race ./...` at HEAD 04d00ad (single full-suite run; every named
test lives in a package reported ok). Live-bus truths (#5, #8, #10) rest
on committed live-run evidence per the phase's binding evidence policy —
and the product-code identity check (`git diff 9f2fd75..HEAD -- engine/
internal/ cmd/ layouts/` → only a test-corpus file) keeps run 35108412175
(31/31, 2026-09-16) valid as evidence for the CURRENT tree.

### Advisory (New Scope, Unevidenced)

Re-verification ran; the Step 7 scan over all covered files and the
changed Phase 4 files found no new blocking pattern (zero debt markers
in focus_helper.py, quiesce_test.go, matrix.go, d48-nightly-dispatch.sh).
Advisory entries: none.

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| 1 | None — only the informational notes below (lint tooling notice; doc wording; matrix re-run pending) | — | — |

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
absorbed silently. (Kept honoring the recorded deferral per the prior
passes' disposition — not re-opened.)

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| engine/conn.go (unchanged) | addr | `Discover()` per cycle: $IBUS_ADDRESS → suffix-filtered freshest ~/.config/ibus/bus file | ✓ real socket discovery (TestAddress_* green; live-registered in the 31/31 matrix run) | ✓ FLOWING |
| engine/engine.go (unchanged) | ev / consume | decodeEvent of ProcessKeyEvent args; verdict from EventHandler | ✓ matrix run 35108412175 proved keys flow at the byte-identical product head | ✓ FLOWING |
| internal/session/actor.go (unchanged this pass) | action n / corrections / snapshot / version | FSM Feed at timer expiry; one config Snapshot per event; version pinned at construction | ✓ TestActor_* green at HEAD (incl. TestStatusCarriesVersion) + live matrix corrections | ✓ FLOWING |
| layouts/tables.go (unchanged) | ENToRU/RUToEN | generated from system xkb | ✓ regeneration byte-identical this pass | ✓ FLOWING |
| docs/adr/d01-experiment-log.md (unchanged) | verdict rows | appended by d01-probe live runs | ✓ 15 rows; append path intact in case_d01.go | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Build | `go build ./...` | exit 0 | ✓ PASS |
| Vet | `go vet ./...` | exit 0 | ✓ PASS |
| Full race suite at HEAD (single run) | `go test -race -count=1 ./...` | 14 test-bearing packages ok, 0 failures (incl. test/e2e with new quiesce corpus) — covers TestEngine_RecoverContainsPanic, TestActor_*, TestFSM_*, TestGolden_*, TestLogging_* | ✓ PASS |
| Lint (strict v2 config) | `mise run lint` | 0 issues | ✓ PASS |
| Tidy drift | `go mod tidy` + `git diff --exit-code -- go.mod go.sum` | zero drift | ✓ PASS |
| Generation determinism | `go generate ./layouts && git diff --exit-code -- layouts/` | byte-identical | ✓ PASS |
| Recover-shim count | `grep -c "defer recoverHandler" engine/engine.go` (21) + factory.go (1) | 21 handlers + CreateEngine; 25 exported methods, 0 new since prior pass | ✓ PASS |
| No device capture (widened) | `grep -rniE "EVIOCGRAB\|/dev/uinput\|/dev/input\|evdev" engine/ cmd/ internal/ layouts/` | comment-only matches (keycode semantics); no access | ✓ PASS |
| FSM purity | `grep -nE "go func\|chan \|time.Now\|Sleep" internal/hotkey/fsm.go` | no matches | ✓ PASS |
| Journal verdict count | `grep -c "Verdict\|Вердикт" docs/adr/d01-experiment-log.md` | 15 (≥4) | ✓ PASS |
| Daemon wiring | grep of cmd/goswitchd/main.go | engine.Run (74), session.NewActor (125), Handler: actor (152), additive SetVersion (126) | ✓ PASS |
| Key trace level | `grep -n 'slog.Debug("key"' engine/engine.go` | line 160, DEBUG-only | ✓ PASS |
| focus_helper `focus` contract (TEST-03) | grep + diff vs 42d10a1 | cmd_focus + grabFocus untouched; witness rework additive with verbatim full-walk fallback | ✓ PASS |
| `-version` flag behavior | `go run ./cmd/goswitchd -version` | prints `goswitchd dev`, exits before run() | ✓ PASS |
| Product-code identity vs live-proven head | `git diff 9f2fd75..HEAD -- engine/ internal/ cmd/ layouts/` | only cmd/goswitchctl/main_test.go (test corpus) | ✓ PASS |
| CI green (headless triplet) | `gh run list` | pr-sanity 35130001841 success (main); dependabot 35406821794 success (2026-09-18); security-scheduled 34838604697 success; release 35130309174 success | ✓ PASS |
| e2e-matrix post-gap-closure re-run | `gh run list` — runs 35142914380/35188570260/35271719601 (head 9ff56cb, 2026-09-16..17) failed in the quiesce preflight only; no dispatch run yet after the 04-08/04-09 fixes (committed 2026-09-17T22:1xZ) | Phase 4's D-48 acceptance item — daemon paths unaffected (0/31 cases started; product head byte-identical to the 31/31-proven 9f2fd75) | ? SKIP (by policy — Phase 4 surface, owner's desktop) |
| Live m1/ibus-restart/kill9/d01 mise tasks | not re-run (drive owner's desktop) | committed Phase 1 evidence + intact assertions + unchanged conn.go + product-code identity with the 31/31 run head | ? SKIP (by policy) |

### Probe Execution

No `scripts/*/tests/probe-*.sh` conventions in this repo (the new
scripts/d48-nightly-dispatch.sh is a session dispatcher for the
self-hosted runner, not a repo probe). The runnable probes are the mise
tasks: headless CI gates re-run this pass (PASS, table above);
live-session e2e tasks excluded by the binding evidence policy — their
machine-generated CI counterpart's last green run (e2e-matrix 31/31,
head 9f2fd75, 2026-09-16) remains valid for this tree via the
product-code identity check; the pending post-fix green matrix run is
Phase 4's D-48 acceptance item.

### Requirements Coverage

All 10 Phase 1 requirement IDs remain mapped to Phase 1 and Complete in
REQUIREMENTS.md traceability (re-checked this pass); no orphans.

| Requirement | Source Plan | Status | Evidence |
|-------------|------------|--------|----------|
| CORR-08 | 01-02 | ✓ SATISFIED | truths #14-15 |
| INTEG-01 | 01-01 | ✓ SATISFIED | truth #5; live-proven at byte-identical product head (run 35108412175) |
| INTEG-02 | 01-01, 01-03 | ✓ SATISFIED | truths #6, #11, #13; live keyd owner-UAT-accepted |
| INTEG-03 | 01-01, 01-03, 01-05 | ✓ SATISFIED | truths #11-12; CI live on GitHub runners + pr-sanity green on main |
| INTEG-04 | 01-01, 01-03, 01-04 | ✓ SATISFIED | truth #8 |
| INTEG-05 | 01-01, 01-03 | ✓ SATISFIED | truths #9-10 |
| TEST-01 | 01-02, 01-03, 01-05 | ✓ SATISFIED | truths #16-18 |
| TEST-02 | 01-03, 01-04 | ✓ SATISFIED | truth #7; stand live-proven by the 31/31 matrix run |
| TEST-03 | 01-03 | ✓ SATISFIED | truth #7 (focus_helper.py `focus` contract verified untouched after the 04-08/04-09 additive rework) |
| INST-04 | 01-01, 01-05 | ✓ SATISFIED | truths #17, #19 |

### Decision Coverage

The `check decision-coverage-verify` gate reports `skipped: CONTEXT.md
missing` under the current gsd-tools resolution (warning-only,
non-blocking). The phase's 01-CONTEXT.md is byte-untouched since the
prior pass's 12/12 honored verdict — the underlying decision set has not
changed; only the tool's lookup path differs in this version.

### Anti-Patterns Found

Debt-marker scan re-run over all covered files and the files changed
since the prior pass: **ZERO TBD/FIXME/XXX, zero TODO/HACK/PLACEHOLDER,
zero stub patterns**. Carry-forward warnings from the prior passes
(files unchanged, still non-blocking quality debt): boot-time retry race
semantics (conn.go:119-123), NameFlagReplaceExisting inertness
(conn.go:88-101), length-only oracle limitation in locked-session d01
re-runs.

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| engine/conn.go | 93, 122 | carry-forward boot-time retry race + rapid-respawn race (unchanged since prior passes) | ⚠️ Warning | Canonical paths demonstrably green (31/31 live matrix 2026-09-16 at the byte-identical product head); documented quality debt for hardening |
| .golangci.yml → lint output | — | golangci-lint reports `exhaustruct` deprecated since v2.13.0 (replaced by exhaustruct_v5) — informational tooling notice, config still applies, 0 issues | ℹ️ Info | Cosmetic future config rename; no code impact |
| test/e2e/README.md | 201 | INTEG-02 section still carries the Phase 1 wording «`ProcessKeyEvent`, возвращающий `false` (наблюдатель)» — predates the consumption succession; the architectural claim (no evdev layer, keyd below IBus) remains true | ℹ️ Info | Documentation nuance only; a wording refresh matching the conditional-consumption model would help future readers |
| README.md | install section | documents `goswitchctl install` instead of the plan-01-05 manual unit-copy dev instruction | ℹ️ Info | Designed succession (see §9 + `deferred`); SC-level no-root contract intact |

### Human Verification Required

None new. The four environment-gated items from the original pass were
resolved by UAT (01-UAT.md: 4/4 pass, owner-accepted 2026-09-11) and are
not re-opened. No ⚠️ PRESENT_BEHAVIOR_UNVERIFIED truths exist on this
pass. (The post-04-09 green matrix dispatch run is Phase 4's D-48
acceptance item, tracked in Phase 4's own ledger — outside this phase's
must-haves.)

### Gaps Summary

**No gaps.** All 20 must-have truths verified on the current
(post-04-08/04-09) tree at HEAD 04d00ad; 31/31 artifacts pass; 13/14 key
links wired (the 14th is the roadmap-sanctioned designed succession
already recorded in `deferred` — honored, not re-opened; the SC-level
contract it served holds); requirements 10/10 with zero orphans; zero
debt markers; all headless gates re-run green by the verifier at HEAD
(build, vet, 14-package race suite, strict lint, tidy, generation
determinism). The one covered file touched since the prior pass
(test/e2e/focus_helper.py) is additive with the Phase 1 `focus` contract
and witness full-walk semantics preserved verbatim. The committed live
proof of the registration/key-flow chain (e2e-matrix 31/31, head
9f2fd75, 2026-09-16) remains valid for this tree because product code is
byte-identical to that head apart from one test-corpus file; the three
matrix failures after the prior pass are Phase 4's quiesce-preflight
harness gaps (G-4-1/G-4-2), formally closed by plans 04-08/04-09 at
HEAD, and never exercised the daemon. No Phase 1 regression exists.

---

_Verified: 2026-09-20T14:32:52Z_
_Verifier: Claude (gsd-verifier)_
