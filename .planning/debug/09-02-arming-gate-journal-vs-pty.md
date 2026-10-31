---
status: awaiting_human_verify
trigger: "Live validation of the pty-arming fix: mechanically the round ended FAIL (probe never landed in the fixture pty) but the NEW evidence supersedes the R1 layer-3 story — the engine journal HAS full key records for every probe (0x31–0x34, five tries), and the probes LANDED IN THE OWNER'S OWN TERMINAL (prompt collected 123422123422…). Engine receives the keys and delivers/forwards them to a STALE IM context instead of the freshly focused wezterm-XIM fixture — that IS the TERM-01 defect mechanism, honestly reproduced by the smoke. Verdict semantics must change: engine-traced-but-pty-absent is a reproduced verdict, not a failure; only absent-from-BOTH is a harness FAIL."
created: 2026-10-09T00:00:00+03:00
updated: 2026-10-09T13:00:00+03:00
---

## Current Focus

hypothesis: RESOLVED (2026-10-14, live hang). The term04 zenity-control marker never fired despite the daemon's focus pairs: waitFocusIn captured its baseline AFTER the 3 s settle — the baseline sat ABOVE the dialog's self-focus pairs (T+1 s) and the owner's click pairs (T+2-3 s), forever. Fix: the detector is waitFocusInFrom with a caller-captured PRE-SPAWN baseline (focusBaseline) — all four surfaces rewired (wezterm spawn, zenity control, negative wezterm, VTE); no settle inside the detector.
test: REPLAY-PROVEN on real journals (the hard gate): the verbatim focus records from gs-t04.log / gs-flip2.log / gs-smoke3.log / gs-caps2.log replayed through the detector — marker FIRES in both arrival modes (pre-landed self-focus; mid-wait click) for every case, and the negative (no fresh pairs) never fires (TestReplayMarkers_OnRealRecords, TestReplayMarker_Negative_NoFreshPairsDoesNotFire); detector semantics pinned (early pairs fire, new pairs fire, stale-baseline waits).
next_action: committed (e2ce145). Owner Ctrl+C's the hung run and re-runs the identical command.

## Resolution

root_cause: two-layer. (1) Semantics: the arming gate treated engine-traced-but-pty-absent as failure, but that state IS the TERM-01 defect (engine→stale-context IM focus desync for XIM clients) — the smoke had measured the mechanism and called it a harness crash. (2) Mechanics: the Enter/Ctrl+D keystrokes used ydotool 1.0 colon syntax on a 0.1.8 binary, typing "22"/"2332" instead (first-character fallback), so no real Enter ever flushed a canonical tty and the teardown EOF never closed windows (R1's unclosed window explained).
fix: armGate returns armReport{outcome, engineDelta, tries} via classifyArm (pty wins → armWindow; engine probe-keyval trace + empty pty → armStaleContext; neither → armDead FAIL); smoke records "Verdict: reproduced: engine→stale-context desync (XIM)" and PASSES; flip rounds record the row and continue (roundDatum.stale, no injections in the stale branch); recipe/caps record the branch honestly; armProbeEnter = KEY_ENTER and close/sendVTETeof = ctrl+d in 0.1.8 named-sequence form. Only absent-from-BOTH fails.
verification: TestClassifyArm RED→GREEN; full suite + mise run ci green (build/vet/golangci-lint 0 issues/test -race). Owner's confirming run pending.
files_changed: [test/e2e/wezterm.go, test/e2e/wezterm_test.go, test/e2e/case_spike_wezterm.go, test/e2e/case_spike_caps.go]

## Symptoms

expected: after the owner clicks the fixture, the arming gate passes and the case drives flip → type → pty oracle to a verdict row.
actual: `FAIL spike-smoke: arming gate: no key-trace record after 5 modifier taps` although the owner click was confirmed (daemon focus pairs) and the injected input was VISIBLE in the fixture window.
errors: "arming gate: no key-trace record after 5 modifier taps" (wezterm.go armGate).
reproduction: run spike-smoke / wezterm-flip with -owner-launch on the drifted session; owner launches the runbook wezterm fixture and clicks it.
started: R1 of the attach batch, 2026-10-11 19:04:46 (docs/adr/adr-008-spike-log.md, /tmp/gs-batch3.log).

## Eliminated

- hypothesis: the daemon journal was lost/deleted so the gate could not read it (ops explanation).
  evidence: the gate reads s.logPath (a live temp file) via os.ReadFile each poll; a missing file counts 0 — but the R1 journal EXISTED (batch3 log excerpt shows records) and simply held no `"msg":"key"` records for the fixture taps.
  timestamp: 2026-10-09

## Evidence

- timestamp: 2026-10-09
  checked: test/e2e/wezterm.go armGate (lines 375-395), main.go countSub/waitForNew (655-701)
  found: gate = 5×(tap Shift_L → wait 2 s for one new `"msg":"key"` in the daemon log); error message matches R1 verbatim.
  implication: the gate's observable lives in the daemon journal.
- timestamp: 2026-10-09
  checked: internal/engine/engine.go:162-177 ProcessKeyEvent
  found: `"msg":"key"` is logged ONLY inside Engine.ProcessKeyEvent — the IBus engine D-Bus handler.
  implication: a record exists only when a key traverses the engine; window-delivered keys leave no trace.
- timestamp: 2026-10-09
  checked: docs/adr/adr-008-spike-log.md R1 rows + /tmp/gs-batch3.log daemon excerpt
  found: fixture focus pairs at 19:07:04/39/42 with ZERO key records after; owner's own window typing at 19:05:55 DID produce key records (that window routes through the engine); ADR quotes the owner: injection visible in the fixture.
  implication: engine-bypass delivery for the XIM fixture is real on this session; the gate tests the wrong channel.
- timestamp: 2026-10-09
  checked: runbookCmd/openFixture/ownerCapture (wezterm.go), case_spike_wezterm.go oracle reads
  found: every fixture pane runs `cat > <pty>`; oracle reads are full-file hex with exact equality (smoke "78", flip/recipe corpusHexRU/Latin); armGate has 4 call sites (smoke, flip round, recipe round, VTE caps).
  implication: an arming probe must land in the pty (Enter required for the canonical-mode cat) and oracle reads must strip it; retries may double the probe → strip by line, not prefix-once.
- timestamp: 2026-10-09
  checked: ydotool 0.1.8 `key --help` output + binary key-name table (strings) + the owner's prompt content 123422…/2332
  found: 0.1.8's key takes NAMED sequences ("modifiers and keys, separated by plus"; table: KEY_ENTER, SHIFT_R, …); the 1.0 colon form "28:1"/"29:1" typed each arg's FIRST CHARACTER (22 per arming try; 2332 at teardown) — no real Enter, no real Ctrl+D ever fired.
  implication: the Enter flush and the window-closing EOF were silently broken; the named form (KEY_ENTER, ctrl+d) is load-bearing for classification soundness.
- timestamp: 2026-10-09
  checked: the owner's live validation evidence (engine journal + prompt + pty absence)
  found: engine traced every probe (0x31–0x34 ×5 tries) while the fixture pty stayed empty — engine→stale-context desync; R1's layer-3 story (engine deaf) superseded for owner-launched fixtures.
  implication: the desync IS the TERM-01 measurement; the gate must classify it as a verdict, not a failure.
