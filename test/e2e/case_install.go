package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// installWait bounds the status-readiness poll of the installed unit
// daemon (the ctlsvc answers on the session bus right after the daemon's
// start; the poll rides out the enable --now → ibus restart window).
const installWait = 10 * time.Second

// The install-cycle case drives the REAL user-space installer (INST-01,
// D-39/D-40/D-42) through the freshly built goswitchctl and proves the
// FULL product lifecycle on one live desktop: install → correction
// (ghbdtn→привет) THROUGH THE INSTALLED UNIT DAEMON → status → harmless
// reinstall → second correction → uninstall → desktop restored. The stand
// spawns NO daemon of its own (single-instance: the ctlsvc name AND the
// IBus registration belong to the unit daemon) — the standalone flag in
// the case registry switches the stand-daemon spawn and the daemon
// preflight checks off, and every oracle is log-independent (the unit
// daemon writes journald, not the stand's log pipe).

// errInstallSourcesDrift is the uninstall oracle's mismatch (err113).
var (
	errInstallSourcesDrift = errors.New("install-cycle: sources after uninstall differ from the desktop snapshot")
	errInstallListedEngine = errors.New("install-cycle: goswitch-en still in ibus list-engine after uninstall")
	errInstallNeedsZenity  = errors.New(
		"install-cycle correction needs the zenity entry surface (locked-session fallback engaged?)")
)

// installArtifactPaths returns the three $HOME artifacts the installer
// owns (the REAL user paths — the case drives the production installer).
func installArtifactPaths() (xmlPath, unitPath, statePath string) {
	home, err := os.UserHomeDir()
	if err != nil {
		// The desktop precondition guarantees $HOME; an unresolvable home
		// fails the case through the caller's own path checks.
		home = ""
	}

	return filepath.Join(home, ".config", "ibus", "component", "goswitch.xml"),
		filepath.Join(home, ".config", "systemd", "user", "goswitchd.service"),
		filepath.Join(home, ".local", "share", "goswitch", "install-state.json")
}

// runInstallCycle proves the FULL product lifecycle against the live
// desktop: install → unit active → registry + live registration → word
// correction through the INSTALLED unit daemon → status → harmless
// reinstall → second correction → uninstall → artifacts gone, sources
// back to the snapshot (machine check). The deferred best-effort
// uninstall is the Pitfall 9 insurance: a case failing mid-install must
// not leave goswitch owning the desktop.
func runInstallCycle(ctx context.Context, s *stand) error {
	if err := buildDaemon(ctx, s.daemonBin); err != nil {
		return err
	}
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}
	// The unit's ExecStart resolves NEXT TO goswitchctl — both binaries
	// land in the stand's tmp dir, so the installed daemon is the tree's.
	xmlPath, unitPath, _ := installArtifactPaths()

	//nolint:contextcheck // the cleanup must outlive the case context (watchdog cancellation)
	defer func() {
		if _, err := os.Stat(unitPath); err != nil {
			if _, err := os.Stat(xmlPath); err != nil {
				return // nothing of ours left on the desktop
			}
		}
		bctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()
		if _, _, uerr := runCtl(bctx, ctlBin, "uninstall"); uerr != nil {
			fmt.Fprintf(os.Stderr, "e2e: install-cycle best-effort uninstall: %v\n", uerr)
		}
	}()

	if out, errOut, err := runCtl(ctx, ctlBin, "install"); err != nil {
		return fmt.Errorf("install-cycle: install exited non-zero (out %q err %q): %w", out, errOut, err)
	}
	if err := installUnitActive(ctx); err != nil {
		return err
	}
	if err := installRegistered(ctx); err != nil {
		return err
	}
	if err := installCorrection(ctx, s, "first correction"); err != nil {
		return err
	}
	if err := installCtlStatus(ctx, ctlBin); err != nil {
		return err
	}
	if out, errOut, err := runCtl(ctx, ctlBin, "install"); err != nil {
		return fmt.Errorf("install-cycle: reinstall exited non-zero (out %q err %q): %w", out, errOut, err)
	}
	if err := installCorrection(ctx, s, "post-reinstall correction"); err != nil {
		return err
	}

	return runUninstallAndVerify(ctx, ctlBin, s)
}

// installCorrection runs one full word-correction gesture against the
// INSTALLED unit daemon: zenity entry, "ghbdtn" through the physical
// path, double Right Shift, and the LOG-INDEPENDENT oracles — the
// content-exact AT-SPI readback settles at "привет" and the zenity stdout
// prints it on OK (the unit daemon writes journald, so the daemon-log
// gates of the stand-spawned cases are unavailable by design).
func installCorrection(ctx context.Context, s *stand, step string) error {
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return fmt.Errorf("install-cycle %s: %w", step, err)
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errInstallNeedsZenity
	}

	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return fmt.Errorf("install-cycle %s: %w", step, err)
	}
	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return fmt.Errorf("install-cycle %s: %w", step, err)
	}
	if err := s.waitZenityText(ctx, wordResultRU); err != nil {
		return fmt.Errorf("install-cycle %s: %w", step, err)
	}
	out, err := s.closeZenity(ctx)
	if err != nil {
		return fmt.Errorf("install-cycle %s: %w", step, err)
	}
	if out != wordResultRU {
		return fmt.Errorf("install-cycle %s oracle: entry printed %q, want %q", step, out, wordResultRU)
	}

	return nil
}

// installCtlStatus proves the INST-02 surface rides the installed unit:
// goswitchctl status answers with the mode report (the ctlsvc name is
// owned by the unit daemon), polled to ride out the unit start.
func installCtlStatus(ctx context.Context, ctlBin string) error {
	deadline := time.Now().Add(installWait)
	for {
		out, _, err := runCtl(ctx, ctlBin, "status")
		if err == nil && strings.Contains(out, "mode=") {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("install-cycle: status never answered with mode= (out %q): %w", out, err)
		}
		if serr := sleepCtx(ctx, witnessPoll); serr != nil {
			return serr
		}
	}
}

// installUnitActive proves the unit side of the install: the unit daemon
// answers systemd as active.
func installUnitActive(ctx context.Context) error {
	state, err := runCmd(ctx, "systemctl", "--user", "is-active", "goswitchd")
	if err != nil {
		return fmt.Errorf("install-cycle: unit not active after install (%s): %w", state, err)
	}
	if state != "active" {
		return fmt.Errorf("install-cycle: unit state %q, want active", state)
	}

	return nil
}

// installRegistered proves BOTH registration views: the XML registry
// (ibus list-engine) and the live daemon (ListActiveEngines).
func installRegistered(ctx context.Context) error {
	listed, err := runCmd(ctx, "ibus", "list-engine")
	if err != nil {
		return fmt.Errorf("install-cycle: ibus list-engine: %w", err)
	}
	if !strings.Contains(listed, "goswitch-en") {
		return fmt.Errorf("install-cycle: ibus list-engine output misses goswitch-en (%q)", listed)
	}
	found, err := listActiveEnginesContain(ctx, "goswitch-en")
	if err != nil {
		return fmt.Errorf("install-cycle: live registration probe: %w", err)
	}
	if !found {
		return errors.New("install-cycle: goswitch-en not in the daemon-side ListActiveEngines reply")
	}

	return nil
}

// runUninstallAndVerify proves the D-42 rollback on the live desktop: the
// unit stops, the artifacts disappear, the registry drops goswitch, and
// the input sources read back equal to the desktop snapshot.
func runUninstallAndVerify(ctx context.Context, ctlBin string, s *stand) error {
	if out, errOut, err := runCtl(ctx, ctlBin, "uninstall"); err != nil {
		return fmt.Errorf("install-cycle: uninstall exited non-zero (out %q err %q): %w", out, errOut, err)
	}

	if state, err := runCmd(ctx, "systemctl", "--user", "is-active", "goswitchd"); err == nil && state == "active" {
		return fmt.Errorf("install-cycle: unit still active after uninstall (%s)", state)
	}
	xmlPath, unitPath, statePath := installArtifactPaths()
	for _, p := range []string{xmlPath, unitPath, statePath} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			return fmt.Errorf("install-cycle: artifact %s still present after uninstall", p)
		}
	}
	listed, err := runCmd(ctx, "ibus", "list-engine")
	if err == nil && strings.Contains(listed, "goswitch-en") {
		return errInstallListedEngine
	}
	now, err := runCmd(ctx, "gsettings", "get", gsettingsSchema, keySources)
	if err != nil {
		return fmt.Errorf("install-cycle: read sources after uninstall: %w", err)
	}
	if now != s.snap.sources {
		return fmt.Errorf("%w: snapshot %q, now %q", errInstallSourcesDrift, s.snap.sources, now)
	}

	return nil
}
