---
status: diagnosed
trigger: "Round-6 stand experiment: does wezterm-nightly (the client on the goswitch stand VM) APPLY DeleteSurroundingText — do the deleted characters actually disappear from the terminal's input line? Matrix: warm-context DST apply on wezterm, same on VTE, the full DST+commit choreography on wezterm (pty oracle: final line «привет»), mutter survival per run. Verdict = the round-6 routing recommendation (DST+commit vs forwards+Require-drain fallback)."
created: 2026-10-19T00:00:00Z
updated: 2026-10-19T18:10:00Z
symptoms_prefilled: true
goal: find_root_cause_only
---

## Current Focus

hypothesis: CONFIRMED (experiment verdict) — wezterm-nightly 20261010 IGNORES server-emitted DeleteSurroundingText in both context classes (pushing-class: Require-reflection unchanged cursor=6 after warm DST; 0x9-class: no erase bytes ever reach the pty); VTE warm DST is a mutter 46.2 session-killer even at +3.2 s focus-age on a Require-ANSWERING context (outside the bisection's virgin window)
test: complete matrix executed — 8 e2e case runs + light manual drives (the bisection's manual-row form); RAW pty captures + Require-reflection + survival witnesses
expecting: (measured) see Evidence + the round-6 verdict block in the spike-log
next_action: none — the experiment is delivered; the round-6 routing recommendation is recorded (DST route dead; round-4 fallback stands)

## Symptoms

expected: the round-6 fix candidate replaces the BackSpace key-forwards (which sit in wezterm's queue ~seconds while CommitText renders instantly — «появилось и исчезло», host defect 19:21/19:36) with DeleteSurroundingText — same text-input channel as commit, order guaranteed by protocol
actual: MEASURED — wezterm ignores DST; VTE DST crashes the shell on mutter 46.2; the routing question is answered
errors: one gnome-shell crash (VTE cell, warm +3.2 s — logged as a data point; session recovered)
reproduction: stand VM goswitch-daemon -debug, wezterm-nightly fixture, warm-gated probe-delete -6 6, Require-reflection + raw-pty oracles (rows in the spike-log round-6 section)
started: 2026-10-19
bug_class: Bohrbug (deterministic client capability + stack defect)

## Eliminated

- hypothesis: (pre-pinned, 2026-10-16 bisection) DST crashes the shell regardless of context state
  evidence: warm-context wezterm DST survived in every run today (5+ emissions); the ONLY crash was VTE at +3.2 s
  timestamp: refined 2026-10-19 (the survive claim was wezterm-only; VTE warm crashes)
- hypothesis: DST is honored by wezterm (the round-6 candidate premise)
  evidence: pushing-class reflection cursor 6 → 6 after DST (manual apply-v2); 0x9-class pty never grows erase bytes (case-run-2, clean-test)
  timestamp: 2026-10-19 17:57
- hypothesis: the warm gate must be a Require-push witness
  evidence: the 0x9-class nightly never answers (run 1 honest refusal); the gate rides the bisection's time boundary + the typed token instead (c27b96d)
  timestamp: 2026-10-19 16:53
- hypothesis: case-run arming failures are harness placement races alone
  evidence: A/B/C isolation (17:18-17:23) — standing daemon delta=2, plain restart delta=0, late ibus-engine poke delta=2: the case's sub-second post-restart activation lands on a cold bus and never wires dispatch (589aef0)
  timestamp: 2026-10-19 17:23

## Evidence

(Ключевые — полный список в спайк-логе, секция round-6; ошибки/крахи честно в рядах.)

- push-класс ночника: typed token lands in the IM surrounding cache (push text=ghbdtn cursor=6 — FRESH, refines the 10-08 stale-cache finding) and is NOT echoed to the pty; 0x9-class: keys pass through to the pty, no pushes. Two classes on one session (17:52 vs 17:57 mints) — the class drift pinned since 10-08.
- wezterm apply (pushing class): baseline push cursor=6 → DST(-6,6) at ≥10 s age → shell alive → reflection push cursor=6 (buffer unshortened) — DST=ignored.
- wezterm apply (0x9 class): pty holds the token; DST → zero erase bytes, no growth (case-run-2, clean-test) — DST=ignored.
- VTE apply: warm ask answered (stale cursor=1) → DST at +3.2 s → shell died ~2.6 s later (case-run-3 vte-apply) — mutter 46.2 session-killer outside the virgin window; upstream guards only from 48.0 (the 10-16 pin).
- Choreography: the commit leg lands byte-exact in the pty (choreo-a DELTA_HEX=d0bfd180…d182 even from a degenerate no-token round; case-run-2's 8-byte capture confirms pty transparency). The token-in-play choreography was starved by the click-placement race; its outcome is determined: DST ignored ⇒ DST+commit ≡ commit-only.
- Mutter survival: every warm wezterm emission survived (5+); the single crash was VTE +3.2 s.
- Harness findings (commits b0b1702..589aef0): the retest-gate supersession (warm-discipline mandate), the gate v1→v2 rework (time boundary + recorded-not-gating Require), wait-free apply reads, pre-typing capture offset, VTE cell last, the 10-08 focus-nudge recipe, and the cold-bus activation cure (late `ibus engine goswitch-en` re-poke per cell). A foreign uncommitted WIP (actor.go press+release pairs — someone's parallel round-6 work) rode the mirror builds; it is inert for the probe path (the verbs bypass executeCorrection).

## Resolution

root_cause: the round-6 DST route is dead on this stack for TWO independent reasons — (1) the client: wezterm-nightly 20261010 ignores server-emitted DeleteSurroundingText entirely (both context classes: the pushing-class IM buffer is unshortened by DST per Require-reflection; the 0x9-class pty never sees erase bytes), and (2) the compositor: mutter 46.2 (noble, no upstream guards until 48.0) crashes gnome-shell on a warm — not only virgin — context (VTE at +3.2 s focus-age on a Require-answering context, session down ~2.6 s later). The host defect's «появилось и исчезло» therefore stays with the round-4 machine: forwards + pty-order drain proofs.
fix: none in goswitch (diagnose-only). Deliverables: the term-dst-apply case + harness fixes (commits b0b1702, c27b96d, 3ebb083, 630974c, 4ba5539, ea270ff, 589aef0), the round-6 section with 11 dated rows + the verdict block in docs/adr/adr-008-spike-log.md, this session file.
verification: matrix rows carry outcomes/counters/survival per emission; the VTE crash recovered (session rebooted, unit+engine+sources verified); the stand left restored (tmp clean, no windows, engine goswitch-en, default kernel).
files_changed:
  - test/e2e/case_term_dst_apply.go (+_test.go) — the case
  - test/e2e/spike_journal_test.go, test/e2e/main.go, mise.toml — registration
  - docs/adr/adr-008-spike-log.md — the round-6 section (rows + verdict)
  - .planning/debug/2026-10-19-stand-dst-apply.md, .planning/debug/knowledge-base.md
