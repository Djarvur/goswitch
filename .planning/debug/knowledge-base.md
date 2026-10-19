# GSD Debug Knowledge Base

Resolved debug sessions. Used by `gsd-debugger` to surface known-pattern hypotheses at the start of new investigations.

---

## 2026-10-19-stand-dst-apply — wezterm игнорирует DeleteSurroundingText; тёплый VTE-DST валит mutter 46.2
- **Date:** 2026-10-19
- **Error patterns:** probe-delete no erase bytes in raw pty, reflection cursor unchanged after DeleteSurroundingText, gnome-shell died after warm DeleteSurroundingText on VTE, "ibus connection lost" after probe-delete, arming gate key-trace floor unmet after daemon restart
- **Root cause(s):** wezterm-nightly 20261010 does not apply server-emitted DeleteSurroundingText in either context class (pushing-class IM buffer unshortened per Require-reflection cursor 6→6; 0x9-class pty never sees erase bytes); independently, mutter 46.2 noble (no NULL/bounds guards upstream until 48.0) crashes gnome-shell on a warm-context VTE DST at +3.2 s focus-age — outside the 10-16 bisection's virgin window
- **Fix:** none in goswitch (diagnose-only round-6 experiment); the terminal write stays on the round-4 machine (forwards + pty-order drain proofs); the 10-16 DST guard and retest-gate remain in force for the terminal DST primitive
- **Files changed:** test/e2e/case_term_dst_apply.go, test/e2e/case_term_dst_apply_test.go, test/e2e/spike_journal_test.go, test/e2e/main.go, mise.toml, docs/adr/adr-008-spike-log.md
- **Why not caught:** no gate existed for this class — client DST application was never measured before (the 10-16 bisection measured survival, not application); the e2e suite exercises forwards, never DST
- **Recurrence guard:** spike-log round-6 section rows (dated apply/reflection evidence); the 10-16 retest-gate (DST emissions forbidden on the stack until a pinned upstream mutter fix); the term-dst-apply case remains registered for future retests
---
## 2026-10-19-stand-dst-apply (harness) — рестарт демона на холодной шине рвёт диспетчеризацию ключей
- **Date:** 2026-10-19
- **Error patterns:** arming gate key-trace floor unmet, zero "msg":"key" records after daemon restart, focus pairs present but no key records, SetGlobalEngine rc=0 no effect
- **Root cause(s):** a daemon restart sub-second after spawn («global engine unreadable; falling back to the current index» — cold bus) leaves key dispatch unwired: ibus reports the engine active, focus events flow, but ProcessKeyEvent never fires; the late `ibus engine goswitch-en` CLI poke (seconds later) re-wires it (A/B/C isolation: standing delta=2 → restart delta=0 → poke delta=2)
- **Fix:** dstReassertEngine — the per-cell late re-poke in the term-dst-apply case (commit 589aef0); the 10-08 focus-nudge recipe (dstFocusNudge) rides every arming gate
- **Files changed:** test/e2e/case_term_dst_apply.go
- **Why not caught:** the case start ladder (activateGoswitchRetry) checks rc and focus marks, not actual dispatch; no prior case restarted the daemon on a cold fresh-boot bus
- **Recurrence guard:** the round-6 section rows (case-runs-5..8 + the A/B/C isolation row); any new stand case that restarts the daemon must re-poke the engine AFTER its fixture's focus gate
---
