---
status: diagnosed
trigger: "On stand VM, emitting IBus engine messages DeleteSurroundingText + RequireSurroundingText reproducibly crashes gnome-shell (signal 11, ~1s after probe). 4/4 with emission died; 2/2 controls without emission survived; kernel 6.8.0-146 one-off boot also crashed (kernel regression ruled out). Localize trigger by bisection, capture stack, deliver round-4 verdict."
created: 2026-10-16T17:40:00Z
updated: 2026-10-16T18:55:00Z
symptoms_prefilled: true
goal: find_root_cause_only
---

## Current Focus

hypothesis: CONFIRMED — mutter 46.2's meta_wayland_text_input_focus_delete_surrounding (meta-wayland-text-input.c:254) dereferences the virgin surrounding-text state (text NULL, cursor uninitialized garbage) when the engine emits DeleteSurroundingText (any params, Require not needed) within ~1.5 s of a fresh context generation; strlen on a wild pointer → gnome-shell SIGSEGV.
test: bisection matrix executed (6 valid runs); 3 independent cores with the identical frame
expecting: (confirmed) DST-only fast CRASH; pair fast CRASH; Require-only fast SURVIVE; warm-context DST/pair SURVIVE
next_action: none — root cause delivered (goal: find_root_cause_only); round-4 routing owns the fix

## Symptoms

expected: emitting DST+Require from the engine updates surrounding state; gnome-shell survives
actual: gnome-shell dies SIGSEGV ~2.3-2.5 s after a DST emission in the virgin window; 6/6 crash-family runs on the trigger shape; warm-context and Require-only controls survive
errors: GNOME Shell crashed with signal 11; strlen on wild pointer (rdi=0x50c) in meta_wayland_text_input_focus_delete_surrounding
reproduction: stand VM, restart goswitchd → spawn fresh wezterm → focus_in → type → probe-delete 0 0 within ~0.5 s → crash ~2.5 s (~/bisect/run3.sh on the stand)
started: 2026-10-16 (first DST emission ever exercised; green 2026-10-14 sessions had no DST)
bug_class: Bohrbug (deterministic in the virgin window; upstream mutter defect triggered by our emission)

## Eliminated

- hypothesis: kernel 6.8.0-147 regression on stand
  evidence: one-off boot into 6.8.0-146 crashed with same signature
  timestamp: 2026-10-16 (pre-session, orchestrator)
- hypothesis: any DST+Require emission crashes regardless of context state
  evidence: R1b manual run (emission onto warm answering context) — shell survived
  timestamp: 2026-10-16 18:06 UTC
- hypothesis: RequireSurroundingText is the killing half
  evidence: R5b Require-only in the exact crash shape (+0.44 s after focus_in) — shell survived
  timestamp: 2026-10-16 18:44 UTC
- hypothesis: DST parameters matter (real deletion vs no-op)
  evidence: DST(0,6) fast-survived at +1.5 s (R3) while DST(0,0) no-op crashed at +0.49 s (R7) — timing, not parameters, selects the window
  timestamp: 2026-10-16 18:47 UTC
- hypothesis: the pair (DST+Require) is required (AND-gate)
  evidence: R7 DST-only fast crashed — the AND-gate does NOT fire; DST alone suffices in the virgin window
  timestamp: 2026-10-16 18:47 UTC

## Evidence

- 2026-10-16 (pre-session): 4/4 e2e crashes correlate with the baseline DST(0,0)+Require ask; 2/2 no-emission controls survived.
- 18:09 (R2, faithful e2e repro under coredump): ask 18:09:24.562 (+0.48 s after focus_in), zero surrounding pushes, shell died 18:09:26.8; core pid 1292.
- 18:09-18:47: three cores (1292 pair-e2e, 1285 pair-fast, 1308 DST-only-fast), one identical frame — meta_wayland_text_input_focus_delete_surrounding at meta-wayland-text-input.c:254 (gdb + libmutter-14-0-dbgsym 46.2-1ubuntu0.24.04.18); strlen arg 0x50c = garbage cursor over NULL text.
- 18:36-18:47: bisection matrix (see §1 below) — full trigger isolation.
- 18:51: production-code flag — the in-tree terminal branch carries a DST primitive (actor.go CORR-07); safe only outside the virgin window; reconciled in the verdict.

## Resolution

root_cause: mutter 46.2 (libmutter-14, src/wayland/meta-wayland-text-input.c:254) segfaults gnome-shell when the engine emits DeleteSurroundingText (any parameters; Require unnecessary) toward the text-input-v3 bridge context inside the ~1.5 s virgin window after a context generation restart (daemon/ibus restart, fresh session): the handler computes strlen over surrounding.text==NULL plus an uninitialized garbage cursor (rdi=0x50c) → SIGSEGV, whole session down. Proven by 3 independent cores with the identical frame + the bisection matrix. goswitch's emission is only the trigger; the defect is upstream mutter.
fix: none in goswitch (not our defect). Instrumentation added: split-emission debug verbs probe-delete/probe-require (harness for this and future bisections, fail-closed under non-debug). Round-4 routing per the verdict §3: drop DST as the terminal write primitive on GNOME 46; forwards + pty-oracle proofs; Require stays GTK-verify-only.
verification: matrix rows in docs/adr/adr-008-spike-log.md round-3 section (2026-10-16 18:01-18:51); cores + gdb stacks in §2; stand restored (session up, unit active, engine goswitch-en, default kernel, no strays, disk clean)
files_changed:
  - internal/session/probe.go (ProbeDelete/ProbeRequire)
  - internal/session/probe_test.go
  - internal/ctlsvc/ctlsvc.go (Deps+methods, fail-closed)
  - internal/ctlsvc/ctlsvc_test.go
  - cmd/goswitchd/main.go (debug-gated wiring)
  - cmd/goswitchd/main_test.go
  - cmd/goswitchctl/main.go (verbs probe-delete/probe-require)
  - cmd/goswitchctl/main_test.go
  - docs/adr/adr-008-spike-log.md (bisection rows + verdict)

---

# DEBUG: stand gnome-shell DST crash — bisection, stack, round-4 verdict

- **Date:** 2026-10-16 (17:40-18:55 UTC)
- **Venue:** stand VM «goswitch-stand» (libvirt, stand@192.168.122.150), Ubuntu 24.04, GNOME Wayland, gnome-shell 46.0-0ubuntu6~24.04.15, mutter 46.2-1ubuntu0.24.04.18, ibus 1.5.29-2, wezterm-nightly 20261010-083911-e2a98e1b, kernel 6.8.0-147 (default)
- **Session:** this file (frontmatter above); spike-log rows in docs/adr/adr-008-spike-log.md, section «Фаза 10 round-3 caps-проба доставки (план 10-13)»
- **Trigger (verbatim):** stand daemon emitting IBus engine DeleteSurroundingText + RequireSurroundingText reproducibly crashes gnome-shell (4/4 e2e runs 2026-10-16 17:03-17:35; controls without emission survived; kernel regression ruled out by one-off 6.8.0-146 boot)

## 1. The trigger — bisected (6 valid matrix runs + e2e repro + 2 invalidated manual attempts)

| Run | Emission | Timing vs fresh focus_in | Client | Result |
|---|---|---|---|---|
| R2 (e2e, 18:09:24) | pair DST(0,0)+Require | +0.48 s | wezterm | **CRASH** (~2.3 s later), core pid 1292 |
| R1b (manual, 18:06:34) | pair DST(0,0)+Require | old answering context (no restart) | wezterm | survive |
| R3 (manual, 18:36:56) | DST(0,6) only, no Require | +1.5 s | wezterm | survive |
| R4 (manual, 18:40:01) | pair DST(0,0)+Require | +23 s (warm) | wezterm | survive |
| R5b (manual, 18:44:33) | Require only, no DST | +0.44 s | wezterm | survive |
| R6b (manual, 18:45:10) | pair DST(0,0)+Require | +0.46 s | wezterm | **CRASH** (~2.5 s later), core pid 1285 |
| R7 (manual, 18:47:39) | DST(0,0) only (no-op), no Require | +0.49 s | wezterm | **CRASH** (~2.5 s later), core pid 1308 |

(R1 18:01 and R6 18:42 were honest aborts: focus/mint gate design error — the ibus engine mint belongs to gnome-shell's shared text-input bridge context and fires on daemon/ibus restart, not on window spawn. No emission, no crash.)

**Trigger statement:** engine-emitted **DeleteSurroundingText, any parameters (no-op 0,0 ≡ real 0,6; Require NOT needed)** landing on the mutter text-input-v3 bridge context within ≲1.5 s of its focus_in in a fresh ibus/daemon generation (the "virgin window" — before any surrounding-text answer has flowed through) segfaults gnome-shell. On a warm context DST is harmless (R3, R4, R1b). Require alone is safe even in the crash window (R5b).

## 2. Crash evidence — three independent cores, one frame

Cores retained on the stand in /var/lib/systemd/coredump/ (compressed, dated):
- core.gnome-shell.1000.43fce0….1292.1792174164000000.zst (18:09:26, pair, e2e R2)
- core.gnome-shell.1000.764ad0….1285.1792176310000000.zst (18:45:12, pair fast R6b)
- core.gnome-shell.1000.a2664d….1308.1792176459000000.zst (18:47:39, DST-only fast R7)

Stack (gdb + libmutter-14-0-dbgsym 46.2-1ubuntu0.24.04.18, build-id 359f1d9859505199d2d89b2564bc3121702b2edc; identical in all three):

```
#4  __strlen_avx2 (libc.so.6)
#5  meta_wayland_text_input_focus_delete_surrounding (focus=…, offset=0, len=0)
        at ../src/wayland/meta-wayland-text-input.c:254   [libmutter-14.so.0 + 0x15a605]
        locals: start = 0x0; text_input = 0x…; resource = <optimized>
#6  clutter_input_focus_delete_surrounding (len=0, offset=0)
        at ../clutter/clutter/clutter-input-focus.c:315   [libmutter-clutter-14.so.0]
#7  meta_wayland_text_input_handle_event                   [libmutter-14.so.0 + 0x162da7]
#8  meta_wayland_seat_handle_event_internal                [libmutter-14.so.0]
#9  meta_wayland_event_handler_handle_event                [libmutter-14.so.0]
#11 _clutter_event_process_filters                         [libmutter-clutter-14.so.0]
#12 clutter_stage_handle_event … g_main_loop_run (main thread)
```

strlen faulting pointer: **rdi = 0x50c** — a wild pointer. Mechanism: the handler computes
`start = surrounding.text + cursor + offset` with `surrounding.text == NULL` (the wezterm
text-input-v3 client never sent set_surrounding_text in the virgin window) and an
uninitialized garbage `cursor` (0x50c) → `strlen(NULL + garbage)` → SIGSEGV. gnome-shell
logs «GNOME Shell crashed with signal 11» + a degenerate GJS frame (ui/init.js:21), the
whole Wayland session tears down (graphical-session.target stops; goswitchd unit dies
with it — the e2e runs' "P1 action: exit status 1" is the session already dead).

**Crashing component: mutter 46.2 (libmutter-14), src/wayland/meta-wayland-text-input.c:254 —
not ibus-daemon, not the client, not goswitchd.** goswitchd's emission is only the trigger;
the defect is the missing NULL/garbage guard in mutter's delete_surrounding handler.

Journal anchors: R2 daemon log preserved at stand:~/bisect/R2-daemon.log (ask 18:09:24.562 →
zero surrounding pushes → focus churn 18:09:26.824 → ibus bus closed 18:09:26.861).

## 3. Verdict for round-4

**(a) Is DeleteSurroundingText usable as the terminal write primitive on THIS stack?**
NO. On mutter 46.2-1ubuntu0.24.04.18 / gnome-shell 46.0 / GNOME Wayland (the stand AND
plausibly any GNOME 46 desktop with text-input-v3 clients), DST is a compositor-killer in
the virgin window (≲1.5 s after a context generation restart: ibus restart, daemon
restart, fresh session). Parameters do not matter — even the no-op DST(0,0) crashes. A
round-4 probe shape that emits DST at probe time (right after fixture spawn/focus) sits
exactly inside the killing window. DST-based terminal delivery is not shippable on this
stack unless guarded to never fire inside the virgin window — a guarantee a layout
corrector cannot honestly make.

**(b) Is RequireSurroundingText alone safe?** YES by this crash data: Require-only fired
at the exact crash timing (R5b, +0.44 s, fresh generation) and survived; all three cores
have exclusively the delete_surrounding frame; the Require path
(clutter_input_focus_request_surrounding → text-input get_surrounding_text) does not
dereference the virgin state. This reconciles with production: the GTK verify path has
issued Require for weeks with zero shell crashes. Caveat kept honest: Require toward a
virgin wezterm-class context is ANSWERLESS (R2 shows 2.3 s of silence) — safe, but it
proves nothing about receipt for non-GTK clients; the surrounding push evidence for
wezterm only exists after the context warmed (R1b pushes).

**(c) What write primitive and delivery-proof shape remains viable for round-4?**
- Write primitive: ForwardKeyEvent (the proven probe-forward channel; the P3 drain used it
  across green sessions) — NOT DST.
- Delivery proof: pty-oracle deltas (the frozen byte-exact pty_hex/pty_delta_hex oracles)
  and daemon journal counters — NOT Require-reflection (P1/P2/P3 reflection predicates are
  unprovable for wezterm-class clients in the probe window anyway, and DST is off the
  table). Receipt (P2) without Require is not directly formulable — route receipt evidence
  through the pty oracle + journal delivery marks, and mark the Require-reflection branch
  as GTK-only.
- Upstream note: the defect is mutter's; a fixed mutter (>46.2; check 46.3+/47+
  surrounding-text rework) may re-open the DST route — re-test before ever re-enabling.

**Production reconciliation (explicit, per the session mandate):** the owner's production
build v1.1.1-142 does not exercise the stand's probe path; nothing from this session was
or can be injected on the host (host stayed read-only). One honest code-level flag for the
owner to reconcile: the in-tree terminal branch still carries a DST primitive (actor.go
CORR-07 «one DeleteSurroundingText exactly over the token+tail range», the level-1 ladder
ADR-003) — on this stack such an emission is safe ONLY outside the virgin window (the
host's warm sessions never hit it, which is consistent with zero host crashes), but the
round-4 plan should treat DST-for-terminals as deprecated on GNOME 46 mutter.

## 4. Harness changes (committed, scope 10-13)

Split-emission debug verbs so the signals are separable (the bisection instrument):
- `goswitchctl probe-delete <offset> <nchars>` — DST only, no Require
- `goswitchctl probe-require` — Require only, no DST
- wired ctl → ctlsvc (fail-closed under non-debug) → actor → engine, TDD RED→GREEN,
  `go build ./...`, `go test -race ./...`, `golangci-lint run` all green.

## 5. Stand restoration (leave it better than found)

- Final state: fresh Wayland session (stand@seat0), goswitchd active and registered
  (engines:2), engine goswitch-en activated, kernel 6.8.0-147 (default), no stray
  wezterm/gnome-terminal processes, disk back to 6.7G used (32G free).
- Installed on the stand (kept deliberately): systemd-coredump + gdb + binutils — the
  stand now retains dated gnome-shell cores for future crashes (three kept as evidence).
- Removed: /tmp/gs-core (528 MB), mutter dbgsym ddeb + extraction, /usr/lib/debug
  symbols; /tmp e2e daemon log preserved at ~/bisect/R2-daemon.log.
- Left on the stand: ~/bisect/{run.sh,run2.sh,run3.sh} — the parameterized bisection
  drivers (fresh-mint discipline, focus_in gate + ydotool center-click nudge per recipe
  10-08).

## 6. Run ledger (honest accounting)

Valid matrix runs: 6 (R2 e2e repro, R3, R4, R5b, R6b, R7) + 2 invalidated manual attempts
(R1, R1b — gate design, emissions rode the warm context) + 1 gate-design abort (R6, no
emission). Reboots: 3 (after R2-crash discovered dead session, R6b, R7). Every run's dated
row lives in docs/adr/adr-008-spike-log.md, section «Фаза 10 round-3 caps-проба доставки
(план 10-13)» — outcomes/counters/latencies only, no field content (D-20/D-21).
