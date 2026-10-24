---
status: diagnosed
trigger: "Nightly e2e-matrix schedule (removed same day, PR #22) crashed the owner's live gnome-shell ~every run: SIGSEGV in libgobject+0x3cd61 during the idle-desktop AT-SPI preflight, session teardown, owner logged out. 6 confirmed crashes 2026-10-14..10-21."
created: 2026-10-23T17:05:00Z
updated: 2026-10-23T18:30:00Z
symptoms_prefilled: true
goal: find_root_cause_only
---

## Current Focus

hypothesis: CONFIRMED — use-after-free in the atk-adaptor object cache of at-spi2-core 2.52.0 (Ubuntu 24.04 / GNOME 46). Any AT-SPI client enumerating the tree (GetItems) makes gnome-shell's embedded atk-bridge iterate `SpiCache->objects`, a raw-pointer hash WITHOUT weak refs; an AtkObject finalized without the deregister signal leaves a dangling pointer; `g_object_ref` on it kills the whole shell.
test: symbolized gdb backtrace of the host's own crash core (apport .crash 2026-10-21 08:56:47, unpacked and analyzed on the stand VM with exact-version ddebs) + source quotes at-spi2-core 2.52.0
expecting: (confirmed) frames g_object_ref ← g_hash_table_foreach(ref_accessible_hf) ← spi_cache_foreach ← impl_GetItems ← droute handle_message
next_action: none — root cause delivered; mitigations shipped same day (nightly cron removed PR #22, runner moved to goswitch-stand VM)

## Root cause (source-level)

gnome-shell 46.0-0ubuntu6~24.04.15, libatk-bridge2.0-0t64 2.52.0-1build1, libglib2.0-0t64 2.80.0-6ubuntu3.9 (host green106; VM matches byte-for-byte).

Stack (gdb + ddebs, core from 2026-10-21 08:56:47 MSK, host pid 931524):

```
#0 g_type_check_instance_is_fundamentally_a  gobject/gtype.c:4153     (freed instance 0x606ef6010e60)
#1 g_object_ref (_object=0x606ef6010e60)     gobject/gobject.c:4199
#2 g_hash_table_foreach (func=ref_accessible_hf)  glib/ghash.c:2117
#3 spi_cache_foreach                         atk-adaptor/accessible-cache.c:423
#4 impl_GetItems                             atk-adaptor/adaptors/cache-adaptor.c:327
#5 handle_other path=/org/a11y/atspi/cache member=GetItems   droute/droute.c:558
#6 handle_message (a11y bus dispatch)        droute/droute.c:605
```

Source, at-spi2-core 2.52.0:

- `atk-adaptor/accessible-cache.c:112` — `cache->objects = g_hash_table_new (g_direct_hash, g_direct_equal);` — raw GObject pointers as keys, no key-destroy, **no weak references anywhere in the file** (grep weak: 0 hits).
- `atk-adaptor/accessible-cache.c:421-424` — `spi_cache_foreach` = bare `g_hash_table_foreach (cache->objects, func, data)`.
- `atk-adaptor/adaptors/cache-adaptor.c:327` — `impl_GetItems` does `spi_cache_foreach (spi_global_cache, ref_accessible_hf, NULL);` before serializing — the first ref of a dangling entry faults here.
- Eviction relies solely on the destroy-signal path `remove_object` (accessible-cache.c:160-177); an object finalized without that signal stays dangling.

Upstream fix: commit `d442ee182e` «atk-adaptor: Have the SpiCache keep its own weak ref» (2025-01-17; in 2.55.2/2.56, NEWS: «Attempt to fix a use after free in the atk bridge»). Commit message names the mechanism verbatim: «Previously, SpiCache relied on the code in accessible-register.c to send a signal when the object went away. At least in theory, this could cause a dangling pointer…». Follow-ups: c4ddb104fb (2025-11-11), 8db3f580bd (2026-06-09). Ubuntu 24.04 stays on 2.52.0 → defect present on any GNOME 46 desktop.

Trigger side (goswitch): `test/e2e/focus_helper.py focused-app` (idle-desktop preflight step of e2e-matrix) walks the whole AT-SPI tree → GetItems per app → crash on the owner's long-lived, app-polluted session. goswitch code is NOT involved (crash lands 6–12 s into the job, before any daemon/matrix step). Repro attempt on a fresh VM session: 25 walks + zenity churn — shell survived (fresh cache has no dangling entries yet); the host core is the authoritative evidence.

## Evidence artifacts

- Host: `/var/crash/_usr_bin_gnome-shell.1001.crash` (2026-10-21 08:56:47, 67 MB apport text, readable by owner uid; Signal 11).
- VM goswitch-stand: `~/host-gs-crash.crash`, unpacked `~/hostcore/` (CoreDump 801 MB; bt.txt/bt2.txt/bt3.txt gdb runs).
- ddebs installed on VM (exact versions, from Launchpad +files): libglib2.0-0t64-dbgsym 2.80.0-6ubuntu3.9, at-spi2-core-dbgsym + libatk1.0-0t64-dbgsym + libatk-bridge2.0-0t64-dbgsym 2.52.0-1build1. Note: libatk-bridge symbols live in the SEPARATE package libatk-bridge2.0-0t64-dbgsym, not in at-spi2-core-dbgsym.
- at-spi2-core 2.52.0 source: /tmp/at-spi2-core-2.52.0 on the VM (download.gnome.org tarball).
- Crash history on host: 2026-10-14, 10-15, 10-17, 10-19, 10-20 (journal signature g_object_ref×2 → fault libgobject+0x3cd61) + 2026-10-21 (core above). The 2026-10-16 stand cores are a DIFFERENT defect (mutter DST, bisected same day).
- VM coredumps 10-16 (3× mutter DST) + 10-19 (SIGABRT mutter text-input assertion) — unrelated to this UAF.

## Mitigations shipped 2026-10-23

1. Nightly schedule removed from e2e-matrix.yml (PR #22, merged 1db5fc5); workflow re-enabled — dispatch-only.
2. Runner moved host green106 → VM goswitch-stand (actions/runner v2.338.1, labels self-hosted,Linux,X64,gnome, user unit in the VM graphical session, GDM autologin already on). Host unit stopped+disabled, files kept, server registration kept (offline).
3. Verification dispatch run (workflow_dispatch, v4, fresh_session=false) — see run 39990616519.

Reversal/notes: to re-enable host runner — `systemctl --user enable --now goswitch-ci-runner` on green106 (registration retained). VM shell crash during a run = job aborts (runner dies with session), autologin rebuilds the session; systemd-coredump on the VM captures the core.

## Addendum 2026-10-24 (verification outcome — honest halt)

CI pipeline on the VM works end-to-end (checkout → mise → daemon build/spawn → report upload; run 40035875529 proved every step except the case bodies). The matrix cases themselves hang at the drivers' focus gates on this venue:

- fresh-rebooted virgin session, zero probes before the run: `word-en-ru` still fails at the 3-min case watchdog (log /tmp/case1.log on the VM);
- venue properties measured 2026-10-24: no map-focus to spawned zenity (known since plan 10-08 for terminals); `grab-input-pid` pokes do not move focus (host-proven recovery); frame-level grabFocus returns `atspi_error (1)`; screen-center and frame-geometry clicks do not grant focus; on probed sessions the shell's search entry holds keyboard focus and accumulates injected keystrokes — probing pollutes the desktop, only virgin-boot runs are meaningful;
- matrix v1–v4 was pinned green ONLY on the host desktop historically; it has never been green on the VM.

Open work (needs its own plan): generalize the phase-10 click-provocation into the drivers' generic focus gates (waitZenityEntry class), or find a working activation primitive for this venue; then re-pin the double run. Mitigations landed meanwhile: job timeout 60→150 min (VM is ~3× slower on GUI fixtures; the 60-min wall killed run #1 mid-matrix).

