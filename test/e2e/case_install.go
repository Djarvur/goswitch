package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Djarvur/goswitch/internal/activate"
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
	if err := installSourcesWrapped(ctx, s); err != nil {
		return err
	}
	if err := installUnitActive(ctx); err != nil {
		return err
	}
	if err := installRegistered(ctx); err != nil {
		return err
	}
	// The D-41 self-check runs right after the install it audits (INST-01,
	// the roadmap's №1 criterion): the registration wait above guarantees
	// the unit daemon is far enough into run() that its ctlsvc name (the
	// audit's version step) is already on the session bus.
	if err := installSelfcheck(ctx, ctlBin); err != nil {
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

// The D-54 wrap enum the install-cycle oracle restates deliberately (NOT
// imported from the installer — an oracle sharing the production wrap code
// could not catch a wrap bug): the xkb layout → goswitch engine mapping
// and the rendered tuple shapes.
const (
	oracleKindXKB       = "xkb"
	oracleLayoutUS      = "us"
	oracleLayoutRU      = "ru"
	oracleEngineTupleEN = "('ibus', 'goswitch-en')"
	oracleEngineTupleRU = "('ibus', 'goswitch-ru')"
)

// carriesOraclePair reports whether the raw line carries the wrappable
// xkb us/ru pair (the oracle's own pair check over the canonical parser).
func carriesOraclePair(raw string) bool {
	tuples, err := activate.ParseSourceTuples(raw)
	if err != nil {
		return false
	}
	hasUS, hasRU := false, false
	for _, t := range tuples {
		switch {
		case t.Kind == oracleKindXKB && t.ID == oracleLayoutUS:
			hasUS = true
		case t.Kind == oracleKindXKB && t.ID == oracleLayoutRU:
			hasRU = true
		}
	}

	return hasUS && hasRU
}

// readStateSources returns the SAVED original sources from the installer's
// state file — the true pre-goswitch pair when the case runs over an
// already-goswitch desktop (the D-54 upgrade form).
func readStateSources() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	path := filepath.Join(home, ".local", "share", "goswitch", "install-state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	var st struct {
		Sources string `json:"sources"`
	}
	if err := json.Unmarshal(data, &st); err != nil {
		return "", fmt.Errorf("parse %s: %w", path, err)
	}

	return st.Sources, nil
}

// wrapForCheck recomputes the D-54 wrapper from a raw sources line: the
// canonical parser (internal/activate) for the wire shape, the oracle's
// own closed enum for the mapping, positions preserved, everything outside
// the pair transiting verbatim.
func wrapForCheck(raw string) (string, error) {
	tuples, err := activate.ParseSourceTuples(raw)
	if err != nil {
		return "", fmt.Errorf("parse %q: %w", raw, err)
	}
	hasUS, hasRU := false, false
	parts := make([]string, 0, len(tuples))
	for _, t := range tuples {
		switch {
		case t.Kind == oracleKindXKB && t.ID == oracleLayoutUS:
			hasUS = true
			parts = append(parts, oracleEngineTupleEN)
		case t.Kind == oracleKindXKB && t.ID == oracleLayoutRU:
			hasRU = true
			parts = append(parts, oracleEngineTupleRU)
		default:
			parts = append(parts, fmt.Sprintf("('%s', '%s')", t.Kind, t.ID))
		}
	}
	if !hasUS || !hasRU {
		return "", errors.New("no wrappable xkb us/ru pair")
	}

	return "[" + strings.Join(parts, ", ") + "]", nil
}

// installSourcesWrapped machine-proves the D-54 criterion 1 on the live
// desktop: right after install the gsettings sources equal the wrapper
// computed from the desk's OWN pair — both goswitch tuples in the pair's
// positions, every other entry transiting verbatim. The pair source is
// the case snapshot when it carries the xkb us/ru pair, and the SAVED
// ORIGINAL from the state file otherwise (the D-54 upgrade form, Pitfall
// 7: an already-goswitch desktop wraps from its saved pre-install pair).
func installSourcesWrapped(ctx context.Context, s *stand) error {
	pairSource := s.snap.sources
	if !carriesOraclePair(pairSource) {
		saved, err := readStateSources()
		if err != nil {
			return fmt.Errorf("install-cycle: resolve the desk's pre-install pair: %w", err)
		}
		pairSource = saved
	}
	wrapped, err := wrapForCheck(pairSource)
	if err != nil {
		return fmt.Errorf("install-cycle: the desk's pair does not wrap: %w", err)
	}
	now, err := runCmd(ctx, "gsettings", "get", gsettingsSchema, keySources)
	if err != nil {
		return fmt.Errorf("install-cycle: read sources after install: %w", err)
	}
	if now != wrapped {
		return fmt.Errorf("install-cycle: sources after install %q, want the computed wrapper %q", now, wrapped)
	}

	return nil
}

// installSelfcheck proves the D-41 audit rides the installed desktop: the
// freshly built CLI exits 0 and prints the six ok verdicts in order.
func installSelfcheck(ctx context.Context, ctlBin string) error {
	out, errOut, err := runCtl(ctx, ctlBin, "selfcheck")
	if err != nil {
		return fmt.Errorf("install-cycle: selfcheck exited non-zero (out %q err %q): %w", out, errOut, err)
	}
	for _, want := range []string{
		"ok version",
		"ok component-visible",
		"ok unit-active",
		"ok engine-registered",
		"ok config",
		"ok input-source",
	} {
		if !strings.Contains(out, want) {
			return fmt.Errorf("install-cycle: selfcheck output misses %q (out %q)", want, out)
		}
	}

	return nil
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
	if state != unitActiveState {
		return fmt.Errorf("install-cycle: unit state %q, want %s", state, unitActiveState)
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
// the input sources read back equal to the desktop's original list — the
// case snapshot, or the SAVED ORIGINAL from the state file when the case
// ran over an already-goswitch desktop (the D-54 upgrade form: the honest
// verbatim restore is the pre-goswitch pair, read BEFORE the uninstall
// consumes the state file).
func runUninstallAndVerify(ctx context.Context, ctlBin string, s *stand) error {
	restoreTarget := s.snap.sources
	if !carriesOraclePair(restoreTarget) {
		if saved, err := readStateSources(); err == nil {
			restoreTarget = saved
		}
	}
	if out, errOut, err := runCtl(ctx, ctlBin, "uninstall"); err != nil {
		return fmt.Errorf("install-cycle: uninstall exited non-zero (out %q err %q): %w", out, errOut, err)
	}

	state, err := runCmd(ctx, "systemctl", "--user", "is-active", "goswitchd")
	if err == nil && state == unitActiveState {
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
	if now != restoreTarget {
		return fmt.Errorf("%w: original %q, now %q", errInstallSourcesDrift, restoreTarget, now)
	}

	return nil
}
