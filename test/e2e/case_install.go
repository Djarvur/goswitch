package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// The install-cycle case drives the REAL user-space installer (INST-01,
// D-39/D-40/D-42) through the freshly built goswitchctl: install puts the
// component XML, the user unit and the takeover sources in place, the unit
// daemon registers live, and uninstall rolls the whole desktop back. The
// stand spawns NO daemon of its own for this case (single-instance: the
// ctlsvc name AND the IBus registration belong to the unit daemon) — the
// standalone flag in the case registry switches the stand-daemon spawn and
// the daemon preflight checks off.

// errInstallSourcesDrift is the uninstall oracle's mismatch (err113).
var (
	errInstallSourcesDrift = errors.New("install-cycle: sources after uninstall differ from the desktop snapshot")
	errInstallListedEngine = errors.New("install-cycle: goswitch-en still in ibus list-engine after uninstall")
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

// runInstallCycle proves the full install lifecycle against the live
// desktop: install → unit active → registry + live registration →
// uninstall → artifacts gone, sources back to the snapshot (machine
// check). The deferred best-effort uninstall is the Pitfall 9 insurance:
// a case failing mid-install must not leave goswitch owning the desktop.
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
	if err := runUninstallAndVerify(ctx, ctlBin, s); err != nil {
		return err
	}

	return nil
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
