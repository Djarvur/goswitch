package install

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/godbus/dbus/v5"

	"github.com/Djarvur/goswitch/internal/config"
	"github.com/Djarvur/goswitch/internal/ctlsvc"
)

// configFileName is the conventional config document the config step
// validates (the planner-pinned path: ~/.config/goswitch/config.yaml).
const configFileName = "config.yaml"

// stepComponent is the D-41 step name of the registry view — named once
// (the repair path returns it twice).
const stepComponent = "component-visible"

// Static red verdicts (err113): every one names the failing step and the
// fix — the D-41 contract that a red selfcheck line is actionable without
// reading logs or the issue tracker.
var (
	errCtlNoAnswer = errors.New(
		"the daemon's control service did not answer — fix: systemctl --user start goswitchd")
	errCtlNoVersion = errors.New(
		"daemon status carries no version (an upgrade left an old goswitchd running?) — " +
			"fix: systemctl --user restart goswitchd")
	errComponentMissing = errors.New(
		"goswitch is not visible to ibus list-engine (even after one env-carrying cache repair) — " +
			"fix: goswitchctl install")
	errUnitInactive = errors.New(
		"the goswitchd user unit is not active — fix: systemctl --user enable --now goswitchd")
	errEngineMissing = errors.New(
		"goswitch-en is not registered with the live ibus-daemon (cache-visible ≠ daemon-visible) — " +
			"fix: journalctl --user -u goswitchd")
	errSourceNotOwner = errors.New(
		"the input sources are not the goswitch single owner — fix: goswitchctl install")
)

// Selfcheck runs the production D-41 audit — the CLI's thin entry (the
// client stays flag parsing and printing only).
func Selfcheck(ctx context.Context, w io.Writer) error {
	return New().Selfcheck(ctx, w)
}

// CtlStatus reports the daemon's control-service status line (the ctlsvc
// Status reply) — the live probe behind the selfcheck's version step
// (D-41 step 1). The corpus replaces it; the production backing dials the
// session bus.
type CtlStatus func(ctx context.Context) (string, error)

// Selfcheck runs the six-step D-41 live audit — version → component →
// unit → engine → config → source — printing one verdict line per step:
// "ok NAME" or "FAIL NAME: <fix hint>". The first red verdict ends the run
// (fail-fast, the e2e-preflight discipline) and is also returned as the
// error, so the CLI exit is non-zero on any red. The output carries paths,
// verdicts and parse reasons only — never user content (D-20/D-21).
func (i *Installer) Selfcheck(ctx context.Context, w io.Writer) error {
	for _, c := range []struct {
		name string
		run  func(context.Context) (string, error)
	}{
		{"version", i.checkVersion},
		{stepComponent, i.checkComponentVisible},
		{"unit-active", i.checkUnitActive},
		{"engine-registered", i.checkEngineRegistered},
		{"config", i.checkConfig},
		{"input-source", i.checkInputSource},
	} {
		ok, err := c.run(ctx)
		if err != nil {
			// One verdict per line: a multi-line cause (the YAML decoder's
			// error is) is whitespace-flattened, the renderStatus canon.
			// The write error is deliberately discarded — the CLI's contract
			// IS terminal output (the printLine precedent).
			_, _ = fmt.Fprintf(w, "FAIL %s: %s\n", c.name, strings.Join(strings.Fields(err.Error()), " "))

			return fmt.Errorf("%s: %w", c.name, err)
		}
		_, _ = fmt.Fprintf(w, "ok %s\n", ok)
	}

	return nil
}

// checkVersion proves the daemon identifies its build (D-41 step 1, D-37):
// the control service answers a status line that carries the version
// token — a silent (or absent) daemon is the first red a broken install
// produces.
func (i *Installer) checkVersion(ctx context.Context) (string, error) {
	reply, err := i.ctlStatus(ctx)
	if err != nil {
		return "", errCtlNoAnswer
	}
	if !strings.Contains(reply, "version=") {
		return "", errCtlNoVersion
	}

	return "version", nil
}

// checkComponentVisible proves the component entered the registry view the
// `ibus list-engine` reads (D-41 step 2). On a miss the audit repairs ONCE
// by re-running the env-carrying write-cache (research Pitfall 1: a plain
// cache run evicts user components; T-04-02-01: never a retry loop) and
// re-checks — an exhausted repair is the red verdict with the install hint.
func (i *Installer) checkComponentVisible(ctx context.Context) (string, error) {
	if i.listEngineHasGoswitch(ctx) {
		return stepComponent, nil
	}
	if err := i.writeCache(ctx); err != nil {
		return "", fmt.Errorf("%w: cache repair failed: %w", errComponentMissing, err)
	}
	if i.listEngineHasGoswitch(ctx) {
		return stepComponent, nil
	}

	return "", errComponentMissing
}

// checkUnitActive proves the systemd user unit runs (D-41 step 3).
func (i *Installer) checkUnitActive(ctx context.Context) (string, error) {
	out, err := i.call(ctx, binSystemctl, "--user", "is-active", daemonBinary)
	if err != nil || strings.TrimSpace(string(out)) != "active" {
		return "", errUnitInactive
	}

	return "unit-active", nil
}

// checkEngineRegistered proves the LIVE daemon-side registration
// (D-41 step 4) — ListActiveEngines on the private IBus socket, the
// deliberately DIFFERENT view from step 2's registry (cache-visible ≠
// daemon-visible, research Pitfall 2; T-04-02-03: both steps are required).
func (i *Installer) checkEngineRegistered(ctx context.Context) (string, error) {
	engines, err := i.activeEngines(ctx)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errEngineMissing, err)
	}
	if !slices.Contains(engines, engineEN) {
		return "", errEngineMissing
	}

	return "engine-registered", nil
}

// checkConfig proves the config document loads (D-41 step 5). A missing
// file is GREEN — install deliberately generates no config, so the daemon
// runs on the built-in defaults until the first user document (the
// 04-planner decision; no file can trip the strict decode, D-33). An
// existing file must load: the red verdict names the path and the parse
// reason (D-20/D-21 allow both — never the document's content).
func (i *Installer) checkConfig(_ context.Context) (string, error) {
	path := i.path(userCfgDirRel, configFileName)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return "config (defaults: no config file)", nil
	} else if err != nil {
		return "", fmt.Errorf("stat %s: %w", path, err)
	}
	if _, err := config.Load(path); err != nil {
		return "", fmt.Errorf("%s: %w (fix the document or remove it to run on defaults)", path, err)
	}

	return "config " + path, nil
}

// checkInputSource proves the D-40 single-owner takeover holds (D-41
// step 6): the gsettings sources carry the goswitch engine.
func (i *Installer) checkInputSource(ctx context.Context) (string, error) {
	out, err := i.call(ctx, binGSettings, "get", gsettingsSchema, gsettingsKey)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errSourceNotOwner, err)
	}
	if !strings.Contains(string(out), "('ibus', '"+engineEN+"')") {
		return "", errSourceNotOwner
	}

	return "input-source", nil
}

// probeCtlStatus is the production CtlStatus: one Status call to the
// daemon's control service on the session bus (the client's callMethod
// shape). Every failure — no bus, no name owner, wire error — is an error:
// the audit renders it as the red verdict with the unit-start hint.
func probeCtlStatus(ctx context.Context) (string, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return "", fmt.Errorf("connect session bus: %w", err)
	}
	defer func() { _ = conn.Close() }() // read-only probe: the close error carries no signal

	var reply string
	call := conn.Object(ctlsvc.BusName, ctlsvc.ObjectPath).
		CallWithContext(ctx, ctlsvc.BusName+".Status", 0)
	if call.Err != nil {
		return "", fmt.Errorf("call Status: %w", call.Err)
	}
	if err := call.Store(&reply); err != nil {
		return "", fmt.Errorf("store Status reply: %w", err)
	}

	return reply, nil
}
