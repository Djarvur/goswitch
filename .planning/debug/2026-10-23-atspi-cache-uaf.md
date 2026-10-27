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

## Addendum 2026-10-27 (venue verdict — the matrix hang, fully root-caused)

Further bisection on the VM (frame timeline via `virsh screenshot` as the primary oracle; pixel-diff vs known plain-desktop/overview reference frames):

1. **A window mapping on the virgin desktop opens the Activities Overview by itself.** A bare `zenity --entry` spawned via systemd-run (no daemon, no driver) flips the desktop to the overview (frame diff 7.4, the exact signature of the in-case flip); the window lives as a thumbnail and can never take keyboard focus. Exonerated by direct experiments: preflights (the only injected key in a whole case is the preflight Shift_R, keycode 54 — evtest under a root systemd unit, the earlier "no keys captured" results were a broken capture: evtest was not even installed), daemon spawn, `ibus engine goswitch-en` activation, `setxkbmap` (plain and with options — desktop byte-identical), the desktop snapshot (reads only).
2. **A single click at screen center closes the overview and the window takes real focus** — typed text verifiably lands in the entry ("Ghbdtn"/"vbhn" visible on frame). The phase-10 click-provocation pattern (nudgeFocusClick) is the right primitive for the focus gates.
3. **The AT-SPI oracle is dead on this venue regardless of focus**: with GTK_MODULES=gail:atk-bridge present in the app env and real keyboard focus, `focus_helper.py witness`/`text zenity` see neither the FOCUSED state nor the entry content (the app node carries a bare frame, no children). On the owner's host the same tree works. GTK4 a11y content publishing is broken on this VM's stack (shell on swrast — "Failed to initialize glamor, falling back to sw" in the shell journal).

Mouse fixes that made the click experiments possible (landed on the VM, 2026-10-25): udev rule `/etc/udev/rules.d/99-ydotool-mouse.rules` tagging the ydotool/ydotoold uinput devices `ID_INPUT_MOUSE=1` (without it libinput/mutter ignored the pointer axis and buttons entirely — keys still worked); persistent `~/.config/systemd/user/ydotoold.service` (WantedBy=default.target). Tooling notes: ydotool 0.1.8 `mousemove` is RELATIVE and needs `--` before negative deltas; anchor resets must use the bottom-right corner (the top-left pass opens the overview via the hot corner); `virsh screenshot` (hypervisor-level) is the trustworthy desktop oracle — the AT-SPI Overview SHOWING flag is not (it reports SHOWING on a plain fresh desktop).

Consequence for the matrix: v1–v4 cases gate on the AT-SPI witness and read text back through AT-SPI — on this venue both are unworkable even with the click fix. Making the matrix green on the VM needs an oracle rework (zenity stdout readback + daemon-journal focus_in gates, the phase-10 pattern), which is a planned task of its own, not a hotfix.

