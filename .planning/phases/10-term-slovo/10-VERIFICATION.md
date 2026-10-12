---
phase: 10-term-slovo
verified: 2026-10-12T22:08:34Z
status: human_needed
score: 5/5 must-haves verified
covered_files:
  - .planning/REQUIREMENTS.md
  - .planning/phases/10-term-slovo/10-01-PLAN.md
  - .planning/phases/10-term-slovo/10-01-SUMMARY.md
  - .planning/phases/10-term-slovo/10-02-PLAN.md
  - .planning/phases/10-term-slovo/10-02-SUMMARY.md
  - .planning/phases/10-term-slovo/10-03-PLAN.md
  - .planning/phases/10-term-slovo/10-03-SUMMARY.md
  - .planning/phases/10-term-slovo/10-04-PLAN.md
  - .planning/phases/10-term-slovo/10-04-SUMMARY.md
  - .planning/phases/10-term-slovo/10-05-PLAN.md
  - .planning/phases/10-term-slovo/10-05-SUMMARY.md
  - .planning/phases/10-term-slovo/10-06-PLAN.md
  - .planning/phases/10-term-slovo/10-06-SUMMARY.md
  - docs/CONFIG.md
  - docs/adr/ADR-008-terminal-transport.md
  - docs/adr/adr-008-spike-log.md
  - internal/appid/appid.go
  - internal/appid/appid_registry_test.go
  - internal/appid/export_test.go
  - internal/clipboard/clipboard.go
  - internal/clipboard/clipboard_test.go
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/correct/stretch.go
  - internal/correct/stretch_test.go
  - internal/session/actor.go
  - internal/session/actor_test.go
  - mise.toml
  - skills/goswitch-config/SKILL.md
  - test/e2e/case_autocorrect.go
  - test/e2e/case_spike_caps.go
  - test/e2e/case_spike_wezterm.go
  - test/e2e/case_switch.go
  - test/e2e/case_term_negative.go
  - test/e2e/case_term_word_vte.go
  - test/e2e/case_term_word_wezterm.go
  - test/e2e/fixtures/gs-e2e-wezterm-stand.lua
  - test/e2e/main.go
  - test/e2e/preflight.go
  - test/e2e/spike_journal_test.go
covered_digest: "v1:sha256:4a9184786834c5fc91911dd912f400426e7f611ee29d516131ec08f2948ebe31"
behavior_unverified: 0
overrides_applied: 0
human_verification:
  - test: "Owner human-check (owner's live desktop, from 10-06-PLAN Task 3 <human-check>): (1) wezterm: type ghbdtn, double Right Shift → привет; (2) same in GNOME Terminal; (3) mid-line word corrected by the gesture; (4) «ytn?» + gesture → «нет,»; (5) autocorrect stayed silent in the terminal during the manual trials. Answer «approved» or describe deviations. Rollback if asked: ~/.local/share/goswitch/releases/v1.1.1/goswitchctl install"
    expected: "All 5 points green on the owner's own desktop (host build v1.1.1-94-ge62a5e4 installed and active); the machine venue (VM stand) already shows byte-exact green rows in both terminals"
    why_human: "The owner's daily desktop (wezterm config, host kernel 7.2.9-2-t2 with its background-window keyboard defect, personal input habits) is outside the VM stand venue; only the owner can sign off criterion 1's acceptance"
  - test: "Formal classification of the TERM-02 edge coverage (flagged assumption carried unresolved by all six plans, never auto-resolvable): review the edge evidence — 8 golden stretch behaviors (ytn?/ghbdtn␠/(ytn/ghbdtn2026/mix/ytn é/YTN?/replace-run) + live rows bash/zsh/mid-line/clip-frozen/witnesses/wezterm ×2 — and give the owner's classification verdict"
    expected: "Owner accepts the edge coverage as expressed in concrete acceptance forms (the plans' own resolution condition — a human_needed verdict — applies only if the edge were NOT so expressed)"
    why_human: "The spec-less probe left the edge `unclassified` by design; the verdict is reserved to the owner, the verifier may only account for the row, never quietly pass it"
---

# Phase 10: TERM-слово Verification Report

**Phase Goal:** Пользователь исправляет слово, набранное не в той раскладке, одним жестом в wezterm и GNOME Terminal — через собственный канал ввода демона (IBus), без внешних инжекторов
**Verified:** 2026-10-12T22:08:34Z
**Status:** human_needed (machine verification complete and green; two owner signatures pending — see Human Verification Required)
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Двойной правый Shift исправляет слово в wezterm (bash/zsh, кириллица, мид-лайн): ghbdtn → привет на живом столе | ✓ VERIFIED | Code present+wired: `test/e2e/case_term_word_wezterm.go` (508 LOC, registered main.go:338 + usage), terminal branch `internal/session/actor.go:3047-3216`. Behavior pinned by passing unit tests (RoutesChunkedNoPending, VTERoleUnifiedChunks, Level2PacesWire). Live machine evidence: word-a/word-b 20:34:29/31 + independent repeat 21:17:16/18 — `pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a`, verifier-decoded byte-exact «привет\n»; mid-line 21:17:20 decodes «echo привет\n». Read-only stand check: goswitchd active, wezterm-nightly 20261010-083911-e2a98e1b installed. Owner's own-desktop signature pending → Human Verification item 1 |
| 2 | То же в GNOME Terminal (VTE); чтение — собственный буфер демона; primary-лег — по необходимости (вердикт владельца OQ 2 «contract») | ✓ VERIFIED | word-bash/word-zsh 19:17:39/41 pass — pty_hex decodes «привет\n» exactly; reading channel wired: Buffer fed by ProcessKeyEvent → `ConvertibleStretch()` consumed at `actor.go:2889-2891`. Primary clause: owner verdict OQ 2 «contract» (2026-10-12, recorded in ADR-008 amendment) supersedes — `internal/clipboard.Primary` exists, tested (pinned argv `--primary --no-newline`), deliberately without consumer; leg-probe (OQ 3) honestly red 20:34:47/21:17:23; D7 boundary documented docs/CONFIG.md:43-51. Codebase honors the owner's authoritative interpretation |
| 3 | N×BackSpace-реплей + commit_text через IBus; CLIPBOARD байт-нетронут; commit-over-selection и paste-чорды не используются | ✓ VERIFIED | `executeLevel2` (actor.go:3164-3216): only ForwardKeyEvent(KeyBackSpace) chunks + ONE CommitText — zero DeleteSurroundingText, zero selection commits. clip-frozen rows: VTE 19:17:46 + wezterm 20:34:45/21:17:22 — clip_before=clip_after=GS-SPIKE-SENTINEL. Paste chord unreachable from terminal branch: settleAfter's noRung branch (actor.go:1668-1675) returns before the rung condition; unit-pinned by TestTerminalBranch_VerifyAfterNoRungOnMismatch (green, including ClipboardRung=true config). No uinput/wl-copy in daemon path (grep clean) |
| 4 | До первого BackSpace отмена бесплатна; после — verify-сверка логирует и останавливается, никогда не чинит; перечек фокуса между чанками ≤8 | ✓ VERIFIED | Behavior-dependent — exercised by passing named tests: TestTerminalBranch_FreeAbortBeforeBurst (all 4 cells: empty-buffer/convert-failed/no-engine/backspace-cap — zero forwards), TestTerminalBranch_FocusLossBetweenChunksStops (burst stops at chunk boundary, no commit, INFO рубеж), TestTerminalBranch_VerifyAfterNoRungOnMismatch + VerifyDeadQuietTimeout. Code: aborts at actor.go:3011-3043 (all before burst); mismatch log+stop at 1667-1675; chunks `terminalChunkKeys=8` with ctxEpoch re-check (`chunkBoundaryMoved`, mutex released one quantum) → «correction outcome=context-changed», no commit, no достройка (3196-3201) |
| 5 | Двухразрешительная модель Фазы 9 живая: замороженные негативные ряды зелёные — автокоррекция молчит при работающей ручной (vim-insert/htop-свидетели) | ✓ VERIFIED | Veto wired at the ARMING point BEFORE the detector: actor.go:2458 (`classifyTerminal` → recordACAbstain(acReasonTerminalVeto)) precedes detectVerdict at 2464; fail-closed via verify-dead source (2508). Unit corpus green: TestAutoCorrect_TerminalVeto{BeforeDetector,PatternMatch,IdentityUnobservable,GTKUnchanged,SingleSlug} + TestTerminalClass_* lifecycle. Live frozen rows: witness-vim 19:29:28 (fired=0, veto Δ=1, buffer_hex=67686264746e200a — verifier-decoded «ghbdtn␠» verbatim), witness-htop 19:29:31, positive-control 19:29:34 («привет\n» byte-exact — silence is the veto, not deafness), autocorrect-terminal-silent-repin 21:13:40 pass (full counters oracle). Counter chain verified end-to-end: recordACAbstain → AutoCorrectSkipReasons → renderStatus `ac_skip_`+dash-flattened slug (ctlsvc.go:195-197) → case regex `ac_skip_terminal_veto` |

**Score:** 5/5 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/correct/stretch.go` | ConvertibleStretch/StretchTail/ReplaceStretch + tablePosition predicate | ✓ VERIFIED | 93 LOC, substantive; consumed at actor.go:2889-2891 (wired); 8-test golden corpus green |
| `internal/config/config.go` (Correction.Terminals) | schema + ceiling 64 + sentinel errors + Defaults() excludes | ✓ VERIFIED | Lines 32-74, 113, 385-397; strict decoder, zero-value=off; docs/CONFIG.md:77 + SKILL.md:151 rows in sync |
| `internal/clipboard/clipboard.go` (Primary) | flagPrimary + Primary(ctx) with Runner seam | ✓ VERIFIED | Lines 36, 108-109; tested; primitive without consumer by owner verdict (documented, not orphaned-by-omission) |
| `internal/session/actor.go` (terminal machinery) | classifyTerminal + verify-dead + veto + terminal branch + chunks + noRung + changed + pacing | ✓ VERIFIED | All symbols present and wired: 61 (chunk=8), 71 (20ms), 129 (slug), 208/281/542 (epoch/dead/noRung), 1736 (armAfterVerifyTerminal), 2458 (veto), 2507 (seam), 3047 (routing), 3164 (branch), 3224 (re-check) |
| `internal/appid/appid.go` | a11y observer registers RegisterEvent interest | ✓ VERIFIED | Lines 49, 326; TestRegisterEventInterest_PinsRegistryCall green; live probe 0→153 StateChanged in journal mechanics row |
| `test/e2e/case_term_word_vte.go` | term-word-vte case: 4 rows, pty oracle, CLIPBOARD sentinel | ✓ VERIFIED | 417 LOC; rows word-bash/word-zsh/mid-line/clip-frozen; registered registry+usage |
| `test/e2e/case_term_negative.go` | term-negative-witnesses: vim/htop + positive control | ✓ VERIFIED | 301 LOC; counters oracle vetoCountRe; rows match journal |
| `test/e2e/case_term_word_wezterm.go` + fixture | term-word-wezterm: caps re-pin, ×2 rounds, delta oracle, leg-probe | ✓ VERIFIED | 508 LOC + `fixtures/gs-e2e-wezterm-stand.lua` (stand-owned config, top-level --config-file form) |
| `mise.toml` | three e2e tasks NOT in ci | ✓ VERIFIED | e2e-term-word-vte (178), e2e-term-negative-witnesses (182), e2e-term-word-wezterm (186) — all with «NOT in ci» |
| `docs/adr/ADR-008-terminal-transport.md` | Status Accepted, Evidence D1–D7 from journal rows, amendment | ✓ VERIFIED | Accepted (2026-10-12); 7/7 D-rows cite dated journal rows; zero web links (grep http = 0); amendment quotes OQ 1 unified / OQ 2 contract, caps re-pin, venue verdict, leg-probe red |
| `docs/adr/adr-008-spike-log.md` | dated venue + live rows + witnesses + caps re-pin + host-install | ✓ VERIFIED | All rows read; hex fragments independently decoded byte-exact («привет», «echo привет», «ghbdtn␠», «ghbdtn»); venue verdict vm-primary cites both probe rounds |
| `docs/CONFIG.md` + `skills/goswitch-config/SKILL.md` | terminals row + D7 contract boundary | ✓ VERIFIED | Row identical in both (line 77/151); D7 boundary paragraph at CONFIG.md:43-51 |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | -- | --- | ------ | ------- |
| internal/correct/stretch.go | internal/session/actor.go startCorrection | ConvertibleStretch consumed for correctionRange | ✓ WIRED | actor.go:2889-2891; autocorrect payload still consumes Token() (D-13/D-14 untouched, existing corpus green) |
| config Correction.Terminals | actor classifyTerminal | refreshACTerminals compile cache | ✓ WIRED | actor.go:577, 1425, 1446, 2577-2583 → matchACTerminal:2532; live-fold on document edit |
| actor autoCorrectBoundary | actor classifyTerminal | veto BEFORE detect.Check | ✓ WIRED | Line 2458 precedes 2464; spy-detector test pins zero detector calls |
| actor startRangeCorrection | actor classifyTerminal | terminal routing BEFORE caps fork | ✓ WIRED | Line 3047; VTERoleUnifiedChunks pins role-60 → same branch |
| actor executeLevel2 | engine Emitter | ForwardKeyEvent(BackSpace) chunks + one CommitText | ✓ WIRED | Lines 3184, 3204; paced 20ms (Level2PacesWire green) |
| actor settleAfter | terminal verify-after | noRung round — mismatch logs and stops | ✓ WIRED | Lines 1668-1675, armAfterVerifyTerminal:1736; noRung returns before rung condition |
| case_term_word_vte.go | actor terminal branch | done + level-2 grep oracle + pty hex | ✓ WIRED | Case asserts done+level-2 marks (lines 218-242) pinning the TERMINAL branch, then pty-hex oracle |
| case_term_negative.go | actor terminal-veto | ac_skip_terminal_veto counters delta | ✓ WIRED | vetoCountRe parses renderStatus output; slug produced dynamically from acReasonTerminalVeto |
| spike-log rows | ADR-008 Evidence D1–D7 | cells cite dated journal rows only | ✓ WIRED | 7/7 rows, zero web links |
| case files | main.go + mise.toml | registry + usage + tasks | ✓ WIRED | main.go:129-130, 276-277, 337-339; mise.toml:178-188 |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| terminal write branch | correction target | daemon's own keystroke Buffer (ProcessKeyEvent feed) → ConvertibleStretch | yes — live rows byte-exact | ✓ FLOWING |
| live e2e oracles | pty hex / delta | real VTE/wezterm pane pseudo-terminals on the VM stand | yes — verifier-decoded «привет\n» | ✓ FLOWING |
| negative witnesses | fired / veto counters | real daemon counters through goswitchctl status renderStatus | yes — Δ=1 rows | ✓ FLOWING |
| verify-after | expected suffix | converted+tail from the same correction round | yes (unit-pinned) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Full CI gate (build+vet+golangci-lint+test -race) | `mise run ci` | exit 0; 21 packages ok | ✓ PASS |
| Chunked routing + unified + free abort + focus stop + no-rung + pacing | `go test ./internal/session/ -run 'TestTerminalBranch_(...)' -race -v` | 7/7 PASS (incl. 4 free-abort subtests) | ✓ PASS |
| Veto before detector + verify-dead lifecycle | `go test ./internal/session/ -run 'TestAutoCorrect_TerminalVeto|TestTerminalClass_' -race -v` | 11/11 PASS | ✓ PASS |
| ConvertibleStretch golden corpus | `go test ./internal/correct/ -run 'TestConvertibleStretch|TestStretch' -race -v` | 8/8 PASS | ✓ PASS |
| Journal hex byte-exactness (independent decode) | python3 bytes.fromhex on d0bfd180…/6563686f20…/67686264746e200a | «привет» / «echo привет» / «ghbdtn␠» | ✓ PASS |
| VM stand alive with internals (read-only) | ssh stand: `is-active goswitchd; command -v ydotool wl-paste gnome-terminal; wezterm --version` | active; all tools present; wezterm 20261010-083911-e2a98e1b | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
| ----- | ------- | ------ | ------ |
| scripts/*/tests/probe-*.sh (conventional discovery) | find scripts -path '*/tests/probe-*.sh' | none exist | N/A — no conventional probes |
| e2e-term-word-vte (phase-declared live verification) | `go run ./test/e2e -case term-word-vte` (inside VM) | executed live during plan 10-05; dated rows 19:17:39-46 in journal; NOT re-run by verifier | SKIP (re-run) — mutating live case (restarts stand daemon, spawns windows, overwrites CLIPBOARD/gsettings); verifier basis: dated rows + independent hex decode + stand reachability |
| e2e-term-negative-witnesses (phase-declared) | `go run ./test/e2e -case term-negative-witnesses` (inside VM) | executed live during plan 10-05; rows 19:29:28-34; not re-run | SKIP (re-run) — same mutation rationale |
| e2e-term-word-wezterm (phase-declared) | `go run ./test/e2e -case term-word-wezterm` (inside VM) | executed live during plan 10-06 twice (20:34, 21:17); rows in journal; not re-run | SKIP (re-run) — same mutation rationale |

Rationale: the three live e2e cases are the phase's own deliverable verification, already executed on the designated venue (vm-primary venue verdict, two-consecutive-greens discipline) with dated, internally consistent, hex-verified rows; re-execution mutates the stand state (daemon restarts, window spawns, clipboard/sources takeover), which the behavioral spot-check constraints forbid. The mise tasks keep them machine-reproducible.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| TERM-02 | 10-01..10-06 (all six) | Пользователь может исправить слово не в той раскладке жестом (двойной правый Shift) в wezterm и GNOME Terminal; чтение — буфер демона / primary (по необходимости); запись — N×BackSpace + commit_text через IBus; безопасная точка отмены; verify-сверка | ✓ SATISFIED | Truths 1-5 above; REQUIREMENTS.md checkbox [x], traceability «TERM-02 \| Phase 10 \| Complete»; live rows both terminals byte-exact; safety invariants unit-pinned |
| — (orphaned) | — | REQUIREMENTS.md maps only TERM-02 to Phase 10 | ✓ NONE | No orphaned requirements |

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
| --------- | ---------- | ------ | ------- | -------- | --------------- | ------- |
| internal/correct/stretch_test.go | TERM-02 | 8 | 0 | no (fixed golden constants) | Value | OK |
| internal/session/actor_test.go (terminal corpus) | TERM-02 | 20 | 0 | no (ops-log series, exact hex) | Behavioral | OK |
| internal/config/config_test.go | TERM-02 | corpus | 0 | no | Value | OK |
| internal/clipboard/clipboard_test.go | TERM-02 | corpus | 0 | no (pinned argv) | Value | OK |
| internal/appid/appid_registry_test.go | TERM-02 | 1 | 0 | no | Value | OK |
| test/e2e/spike_journal_test.go | TERM-02 | sections | 0 | no | Existence (section contract) | OK |

**Disabled tests on requirements:** 0. **Circular patterns:** 0 — journal hex evidence independently decoded by the verifier, not generated by the system under test. **Insufficient assertions:** 0 — the D-24 changed=false half has no behavioral test (WINDOWS ledger #16, defensive branch, ConvertRuns provably never returns changed=false); changed=true is behaviorally pinned; recorded as documented limit, not a gap.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (worktree, uncommitted) | — | untracked `test/e2e/e2e` binary build artifact in git status | ℹ️ Info | Not committed, not part of phase evidence; local `go build` residue |
| none in phase files | — | TBD/FIXME/XXX/TODO/PLACEHOLDER scan across all 27 changed files | — | clean |

### Human Verification Required

### 1. Owner human-check — live-desktop acceptance of criterion 1 (5-point checklist)

**Test:** On the owner's own desktop (host build v1.1.1-94-ge62a5e4 installed and active — journal row host-install 00:19): (1) wezterm: type `ghbdtn`, double Right Shift → see «привет»; (2) same in GNOME Terminal; (3) a word in the middle of a phrase corrected by the gesture; (4) `ytn?` + gesture → «нет,»; (5) autocorrect stayed silent in the terminal during the manual trials. Answer «approved» or describe deviations. Rollback if asked: `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install`.
**Expected:** All five points green; the machine venue (VM stand) already shows byte-exact green rows in both terminals, twice, independently.
**Why human:** The owner's daily desktop (personal wezterm config, host kernel 7.2.9-2-t2 background-window keyboard defect, real workloads) is outside the VM venue; only the owner can sign the acceptance. Declared in 10-06-PLAN Task 3 `<human-check>`; pending end-of-phase harvest.

### 2. Formal classification of the TERM-02 edge coverage (flagged assumption)

**Test:** Review the edge evidence — 8 golden stretch behaviors (punctuation head/tail, space stop, digits, mixed ride-as-typed, é whole-stretch refusal, registers, replace-run) plus the live rows (bash/zsh, mid-line, clip-frozen, vim/htop witnesses, wezterm ×2) — and give the owner's classification verdict on the edge.
**Expected:** The owner accepts the edge as expressed in concrete acceptance forms. The plans' own resolution condition — a forced human_needed — applies only if the edge were NOT so expressed; it is expressed, but the formal verdict stays the owner's (the flagged assumption is never auto-resolved).
**Why human:** The spec-less probe classified the edge `unclassified` by design; every plan demands the verifier account for the row, never a quiet pass.

## Gaps Summary

No gaps. All five ROADMAP success criteria are machine-verified: the terminal write branch (chunks ≤8, focus re-check, free abort, verify-after without clipboard rank, D-24 discriminator) is implemented and unit-pinned in `internal/session/actor.go`; the conversion geometry (`internal/correct/stretch.go`) fixes the punctuation edge with an 8-test golden corpus; the D-55 arming veto sits before the detector with a verified counters chain; ADR-008 is Accepted on dated, hex-verified live evidence; the frozen negative rows pin criterion 5 on the designated VM venue. `mise run ci` is green at HEAD (exit 0, 21 packages under -race; golangci-lint strict).

The phase status is human_needed solely because two owner signatures are open: the live-desktop human-check (declared in 10-06-PLAN, machine evidence already green) and the flagged-assumption edge classification. One non-blocking owner gate is already tracked in ADR-008 Consequences and needs no UAT item: the keep-or-remove decision on the debug `probe-forward` verb (OQ 5, confirmed open precondition, fail-closed without -debug).

---

_Verified: 2026-10-12T22:08:34Z_
_Verifier: Claude (gsd-verifier)_
