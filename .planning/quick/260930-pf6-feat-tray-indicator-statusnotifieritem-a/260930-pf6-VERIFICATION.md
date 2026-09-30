---
phase: 260930-pf6-feat-tray-indicator-statusnotifieritem-a
verified: 2026-09-30T17:43:38Z
status: human_needed
score: 5/5 must-haves verified programmatically; owner pixel confirmation pending
covered_files:
  - .planning/quick/260930-pf6-feat-tray-indicator-statusnotifieritem-a/260930-pf6-PLAN.md
  - internal/indicator/pixmap.go
  - internal/indicator/pixmap_test.go
  - internal/indicator/indicator.go
  - internal/indicator/indicator_test.go
  - internal/indicator/testdata/en.golden
  - internal/indicator/testdata/ru.golden
  - internal/session/actor.go
  - internal/session/actor_test.go
  - internal/ctlsvc/ctlsvc.go
  - internal/ctlsvc/ctlsvc_test.go
  - cmd/goswitchd/main.go
covered_digest: "v1:sha256:b20f220b72a13b3ad60937a5020da5fd23e7fcb733ded26ed92334c8d0c436d3"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: none
  note: "Initial verification. No SUMMARY.md existed at verification time (plan output contract); verification ran directly against the codebase and the live bus."
gaps:
  - truth: "Live acceptance on the desktop: after restart, busctl shows the item registered and the owner sees the EN/RU icon in the top bar change on Super+Space"
    status: failed
    reason: "Registration does not stick. The ubuntu-appindicators watcher (gnome-shell, /usr/share/gnome-shell/extensions/ubuntu-appindicators@ubuntu.com) requires NEEDED_PROPERTIES = ['Id', 'Menu'] (appIndicator.js:400). Our item answers UnknownProperty for Menu (omitted by v0 design), so _checkNeededProperties fails after 3 retries 1s apart, _setupProxy logs 'While initalizing proxy for org.djarvur.goswitch: Gio.DBusError: ...UnknownProperty: Menu' (journal 20:29:38, 3s after attach) and calls this.destroy() — _onIndicatorDestroyed removes the item from RegisteredStatusNotifierItems and takes the icon off the panel. Independently reproduced by this verifier: 6 min after the 20:29:35 restart, busctl RegisteredStatusNotifierItems lists only chromium/update-notifier items. The daemon logs no WARN because the RegisterStatusNotifierItem call itself succeeded — the loss is entirely host-side, and it is deterministic on every attach."
    artifacts:
      - path: internal/indicator/indicator.go
        issue: "The served property set (Category, Id, Title, Status, IconName, IconPixmap, WindowId) omits Menu; Get/GetAll answer org.freedesktop.DBus.Error.UnknownProperty for it, which this desktop's GNOME appindicators watcher treats as a fatal proxy-setup failure and destroys the item ~3-4s after registration"
    missing:
      - "Serve a Menu property of type 'o' (object path); the extension's own icon-only sentinel '/NO_DBUSMENU' (appIndicator.js:570) is the designed answer — no DBusMenu object is needed"
      - "Re-run live acceptance after the fix: systemctl --user restart goswitchd; busctl RegisteredStatusNotifierItems must still contain org.djarvur.goswitch after 10s+; then the owner visual check (below)"
human_verification:
  - test: "After the Menu gap is fixed and goswitchd restarted: owner confirms the EN/RU icon is visible in the GNOME top bar and changes on Super+Space (and on a goswitchctl-driven external sync)"
    expected: "The icon persists in the panel (not just for the first ~3s), shows EN after an EN flip and RU after an RU flip, including sync-corrected modes"
    why_human: "Pixel visibility, panel placement, and visual flip behavior in the live GNOME shell are owner-perception; no programmatic check reads the rendered panel"
coincidental_reliance_items:
  - truth: "The daemon publishes an SNI item ... registered with org.kde.StatusNotifierWatcher at attach"
    reason: undeclared-precondition
    harden: "A nil error from RegisterStatusNotifierItem proves only that the call was accepted, not that the host kept the item: the host's own property demands (NEEDED_PROPERTIES) are an undeclared precondition of a STICKY registration. The gap fix (serve Menu) should promote 'the host accepts the item surface' from assumption to a stated invariant — and the live check should assert registration persistence past the watcher's proxy-setup window, not just an immediate list hit."
---

# Quick Task 260930-pf6: Tray Indicator (StatusNotifierItem) Verification Report

**Task Goal:** feat: tray indicator (StatusNotifierItem/AppIndicator) in goswitchd — daemon publishes EN/RU icon on the session bus, updated at flipTo+syncMode via a mode-observer seam; in-code ARGB32 pixmaps (zero deps, no icon files), graceful degradation when no SNI watcher; wired in cmd/goswitchd
**Verified:** 2026-09-30T17:43:38Z
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SNI item published on the EXISTING ctl connection: properties answerable at /StatusNotifierItem under org.djarvur.goswitch, registered with the watcher at attach — no second bus name, no second connection | ✓ VERIFIED (see coincidental-reliance note) | `indicator.Attach(conn, service)` takes the caller's conn; no ConnectSessionBus/RequestName anywhere in internal/indicator (grep clean). registerItem: probe → double export (sniIface + Properties) → register, pinned by TestAttachRegistersAndExports (register calls == [service], both exports, initial EN pixmap). Live: `busctl --user get-property org.djarvur.goswitch /StatusNotifierItem org.kde.StatusNotifierItem Id` → `s "goswitch"` (this verifier, item live on the bus); the extension processed the registration (icon created at 20:29:35). |
| 2 | Deterministic in-code ARGB32 renderer (EN/RU, white glyphs on transparent bg) pinned by goldens — zero icon files, zero new deps, zero config keys, no SPEC/ADR edits | ✓ VERIFIED | pixmap.go pure Go, no D-Bus imports; TestPixmapForGolden byte-exact vs testdata/{en,ru}.golden with standard -update flag + intra-test repeat-call determinism; TestPixmapTwoTonePalette (A=0xFF glyph / A=0x00 bg), TestPixmapCanvasIsFixed, TestPixmapSymbolsDiffer. `git diff 1833643..HEAD -- go.mod go.sum` EMPTY; full diff stat = exactly the 11 planned files; no internal/config or installer changes. |
| 3 | Every mode change reaches the display: flipTo and syncMode invoke ModeChanged strictly AFTER UpdateModeSymbol; SetModeDisplay pushes the current mode at install (late install self-syncs) | ✓ VERIFIED | actor.go:1422-1424 (flipTo) and :1466-1468 (syncMode) — nil-guarded call after UpdateModeSymbol; SetModeDisplay actor.go:623-631 pushes modeSymbol at install. Behavior pinned by passing tests: TestActor_FlipInvokesDisplayLast and TestActor_SyncDriftInvokesDisplay assert exact op order ["symbol:ru","display:ru"]; TestActor_SetModeDisplayPushesCurrentMode asserts install push AND late install after flips. Live support: orchestrator-observed IconPixmap md5 flips en→ru→en on real engine flips. |
| 4 | Absent watcher / register failure / failed emit each degrade to exactly one WARN and permanent silent no-op; daemon never blocks; full actor corpus green with nil display | ✓ VERIFIED | TestAttachWatcherAbsentIsInert (1 WARN, zero registrations, zero exports, post-degradation ModeChanged silent, still 1 WARN), TestAttachProbeFailureIsInert (probe error → same inert path — the shared registerItem-error degradation), TestAttachEmitFailureSelfDisables (exactly 1 emit attempt, exactly 1 WARN, then silence). Emit is a queued send, no round trip (T-Q6-03). Session corpus passes unchanged with nil display (full -race suite green). |
| 5 | Live acceptance: after restart, busctl shows the item registered and the owner sees the EN/RU icon change on Super+Space | ✗ FAILED | Registration is accepted but does not STICK: the watcher destroys the item ~3-4s after attach because the item refuses the required Menu property. Independent busctl 6 min post-restart: org.djarvur.goswitch ABSENT from RegisteredStatusNotifierItems; journal: `gnome-shell[2449574]: While initalizing proxy for org.djarvur.goswitch: Gio.DBusError: GDBus.Error:org.freedesktop.DBus.Error.UnknownProperty: Menu` at 20:29:38 (3s after the 20:29:35 attach); extension source confirms destroy-on-_checkNeededProperties-failure. The orchestrator's positive snapshot was taken inside the ~3s registration window — the steady state does not hold. Owner visual check is moot until registration persists. |

**Score:** 4/5 truths verified (0 present-but-behavior-unverified)

The orchestrator's live evidence (item in the list, md5 flips, no WARNs) is real but was a window observation: deploy commit a23dc82 landed 20:29:28, service restart 20:29:35, extension destroy 20:29:38. Daemon-side facts it attests (properties answerable, pixmap updates on flips) were independently re-confirmed by this verifier via direct property reads; the registration presence did not survive the window.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/indicator/pixmap.go` | Deterministic EN/RU ARGB32 renderer | ✓ VERIFIED | 121 lines, pure, closed-input API, golden-pinned |
| `internal/indicator/testdata/en.golden` + `ru.golden` | Determinism pins | ✓ VERIFIED | 2312 bytes each, byte-exact comparison, -update regen contract |
| `internal/indicator/indicator.go` | Watcher probe, register, exported item, ModeChanged | ✓ VERIFIED (artifact) | 281 lines; all three degradation paths + double export + mutex-guarded pixmap swap; substantive 12-test corpus |
| `internal/session/actor.go` | ModeDisplay + SetModeDisplay + observer-last | ✓ VERIFIED | Seam mirrors SetSwitcher; install push; observer last in both sites |
| `internal/ctlsvc/ctlsvc.go` | OnConn in Deps, WARN-and-continue | ✓ VERIFIED | Called once after RequestName+Export, before serve wait; error WARNed, serving continues (wire-verified against a real test bus) |
| `cmd/goswitchd/main.go` | Indicator from ctl conn, SetModeDisplay before engine.Run | ✓ VERIFIED | startCtl (line 78) precedes engine.Run (79); OnConn closure attaches with ctlsvc.BusName |

All artifacts: exists + substantive + wired. No stubs, no orphans.

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| actor.flipTo / syncMode | ModeDisplay.ModeChanged | observer last, after UpdateModeSymbol, both sites | ✓ WIRED | actor.go:1422-1424, :1466-1468; order pinned by two passing tests |
| indicator.ModeChanged | pixmap swap + NewIcon emit | indicator-own mutex, /StatusNotifierItem | ✓ WIRED | indicator.go:206-230; TestAttachModeChangedEmitsAndServesPins (exact emit path + served pixmap swap) |
| ctlsvc.Run | OnConn(conn) → indicator.Attach + actor.SetModeDisplay | post-export hook, before engine's first generation | ✓ WIRED | ctlsvc.go:216-220; main.go:78/106-110; TestRun_OnConnHookCalledOnce (real bus, exactly-once, serving answers after hook) |
| renderer output | golden files | byte-stable determinism (SPEC §7.1) | ✓ WIRED | TestPixmapForGolden byte-exact; repeated-call stability asserted in-test |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| IconPixmap property | it.pix | PixmapFor(symbolEN/RU) in-code render, swapped by ModeChanged | Yes | ✓ FLOWING |
| ModeChanged symbol | a.modeSymbol() | actor's factual mode (flipTo/syncMode/settle records) | Yes | ✓ FLOWING |

No static or mock-fed values in the rendered path.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Three touched packages under -race | `go test -race -count=1 ./internal/indicator/ ./internal/session/ ./internal/ctlsvc/` | all ok | ✓ PASS |
| Full CI gate (build+vet+lint+test -race all pkgs+tidy-diff) | `mise run ci` | all 13 packages ok, gates green | ✓ PASS |
| Item properties answerable on live bus | `busctl --user get-property org.djarvur.goswitch /StatusNotifierItem org.kde.StatusNotifierItem Id` | `s "goswitch"` | ✓ PASS |
| Watcher steady-state registration | `busctl --user get-property org.kde.StatusNotifierWatcher ... RegisteredStatusNotifierItems` | goswitch ABSENT (3 unrelated items only) | ✗ FAIL |

### Probe Execution

No scripts/*/tests/probe-*.sh declared or conventional for this task; the unit corpora + mise run ci + live busctl checks above are the executable evidence.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| QUICK-260930-pf6-tray-indicator | 260930-pf6-PLAN.md | Tray indicator per plan success criteria | ⚠️ PARTIAL | Daemon-side criteria fully met (zero deps/config/files, observer seam, degradations, D-36 order intact — scope guard diff = exactly 11 files). Live criterion fails: icon does not persist (Menu gap). |

Not orphaned: REQUIREMENTS.md has no phase-mapped entry for this quick ID; the plan's own success criteria are the contract.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none) | — | TBD/FIXME/TODO/PLACEHOLDER scan clean; gofmt clean; no empty implementations, no hardcoded empties in the rendered path | — | — |

Note: the deployed binary is version "dev" (no ldflags stamping) — pre-existing condition, not introduced by this task.

### Human Verification Required

1. **Owner visual acceptance (after the Menu gap fix + daemon restart)**
   **Test:** Confirm the EN/RU icon is visible in the GNOME top bar and changes on Super+Space (and on an external sync via goswitchctl).
   **Expected:** Icon persists in the panel well beyond the first seconds; EN after EN flips, RU after RU flips including sync corrections.
   **Why human:** Pixel visibility and flip perception in the live shell are owner-perception; no programmatic read of the rendered panel exists.

### Gaps Summary

One gap, one root cause, and it is small and precisely localized. The entire daemon-side deliverable is real and verified: deterministic golden-pinned renderer, one-connection SNI item, observer-last mode seam with install push, one-WARN degradation corpus, clean wiring, zero scope creep (diff = exactly the 11 planned files; go.mod untouched), full CI green. Five unit-level behavior contracts are pinned by substantive tests, and the live property reads plus orchestrator-observed md5 flips confirm the daemon side on the real bus.

What fails is the last mile: the desktop's ubuntu-appindicators watcher demands `Id` and `Menu` (NEEDED_PROPERTIES, appIndicator.js:400), retries each 3× at 1s, then destroy()s the indicator on failure — so the item is evicted from RegisteredStatusNotifierItems ~3-4s after every attach and the icon never stays in the panel. The plan's "Menu omitted in v0 (extension renders icon-only, owner-verified)" premise is falsified by this extension version; the extension even ships the designed icon-only answer — the `/NO_DBUSMENU` sentinel it checks at appIndicator.js:570. Serving `Menu` as object path `/NO_DBUSMENU` (a one-property addition to the item's property set, plus its test) should flip _checkNeededProperties to green and make the registration stick.

Not deferred: no later phase covers this (quick task; the fix is in this task's own artifact).

Process note: the plan's output contract (`260930-pf6-SUMMARY.md`) had not been created at verification time; this report verifies against the codebase and live bus directly.

---

_Verified: 2026-09-30T17:43:38Z_
_Verifier: Claude (gsd-verifier)_

## Re-verification addendum (2026-09-30, post-fix, orchestrator)

The gap was closed by commit 9063703: the item now serves `Menu` as the
`/NO_DBUSMENU` object path — the extension's own icon-only sentinel
(appIndicator.js `menuPath`), which satisfies `NEEDED_PROPERTIES = ['Id',
'Menu']` without a DBusMenu tree. RED (TestItemProperties pins the sentinel,
GetAll = 8) → GREEN, full `mise run ci` green.

Live re-acceptance on the desktop (daemon rebuilt from 9063703, unit
restarted 20:5x):
- `RegisteredStatusNotifierItems` contains `org.djarvur.goswitch` at t+1s,
  t+5s, t+12s, t+25s after restart — the registration STICKS; gnome-shell's
  journal logs no destroy chain for goswitch.
- `IconPixmap` property readback flips deterministically en→ru→en on
  `ibus engine` flips (md5 cce7bd2d24ee ↔ 57014693527e) — the ModeChanged
  path is live on the real bus.
- The daemon journal carries no indicator WARNs (the only WARNs are the
  expected `mode corrected` records of external flips).

Remaining: the owner's pixel-level confirmation that the EN/RU icon is
visible in the top bar and changes on Super+Space (human_needed, the
plan's own verify gate).
