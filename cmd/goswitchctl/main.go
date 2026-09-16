// Command goswitchctl is the daemon's control client (INST-02) and the
// installer's entry point (INST-01): status (with --json), reload and
// correct drive the control service on the D-Bus session bus; install and
// uninstall [--purge] run the user-space lifecycle (all logic lives in
// internal/install — the client stays flag parsing and printing only).
// Stdlib only (the project convention: no CLI framework), all dispatch in
// run so the exit contract lives in one place.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/Djarvur/goswitch/internal/install"
)

// The control service's D-Bus coordinates (internal/ctlsvc owns the service
// side; the literals stay here so the control branches link no daemon
// packages — a thin client).
const (
	ctlBusName = "org.djarvur.goswitch"
	ctlPath    = "/org/djarvur/goswitch"
)

// ctlCallTimeout bounds every D-Bus call: a wedged session bus must fail
// the named subcommand, not hang the terminal.
const ctlCallTimeout = 5 * time.Second

// installTimeout is the install/uninstall budget — NOT ctlCallTimeout:
// the lifecycle spans an ibus restart and a unit start (research: 5 s is
// too small for that chain).
const installTimeout = 120 * time.Second

// errDaemonNotRunning names the friendly verdict for a missing name owner
// — the acceptance-pinned wording.
var errDaemonNotRunning = errors.New("daemon not running? (org.djarvur.goswitch is not on the session bus)")

// errUsage is the subcommand grammar (err113: static).
var errUsage = errors.New("usage: goswitchctl status [--json] | reload | correct | install | uninstall [--purge]")

// configErrorKey terminates the status token grammar: its value is the
// LAST field and may contain spaces (a flattened parse error), so the
// parser must treat everything after it as the value.
const configErrorKey = "config_error="

func main() {
	// No defer here: os.Exit skips defers, so error reporting lives in main
	// itself (the goswitchd skeleton).
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "goswitchctl: %v\n", err)
		os.Exit(1)
	}
}

// run dispatches the subcommand and prints the reply; a non-nil error is
// the non-zero exit contract (a rejected reload, an unreachable daemon).
// The D-Bus branches carry ctlCallTimeout; install/uninstall get their OWN
// installTimeout budget (they never dial the daemon).
func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errUsage
	}

	switch args[0] {
	case "status":
		ctx, cancel := context.WithTimeout(ctx, ctlCallTimeout)
		defer cancel()

		return runStatus(ctx, args[1:])
	case "reload":
		ctx, cancel := context.WithTimeout(ctx, ctlCallTimeout)
		defer cancel()

		reply, err := callMethod(ctx, "ReloadConfig")
		if err != nil {
			return fmt.Errorf("reload: %w", err)
		}
		printLine(reply)
	case "correct":
		ctx, cancel := context.WithTimeout(ctx, ctlCallTimeout)
		defer cancel()

		reply, err := callMethod(ctx, "CorrectNow")
		if err != nil {
			return fmt.Errorf("correct: %w", err)
		}
		printLine(reply)
	case "install":
		return runInstallCmd(ctx)
	case "uninstall":
		return runUninstallCmd(ctx, args[1:])
	default:
		return fmt.Errorf("unknown subcommand %q: %w", args[0], errUsage)
	}

	return nil
}

// runInstallCmd executes the D-39 install under its own budget and prints
// the step-by-step report.
func runInstallCmd(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, installTimeout)
	defer cancel()

	report, err := install.Install(ctx)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	printReport(report)

	return nil
}

// runUninstallCmd executes the D-42 rollback under its own budget;
// --purge additionally removes the user's ~/.config/goswitch data (D-42).
func runUninstallCmd(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	purge := fs.Bool("purge", false, "also remove the user's ~/.config/goswitch data (default: preserved)")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("uninstall flags: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, installTimeout)
	defer cancel()

	report, err := install.Uninstall(ctx, *purge)
	if err != nil {
		return fmt.Errorf("uninstall: %w", err)
	}
	printReport(report)

	return nil
}

// printReport writes the lifecycle report line by line — the CLI's
// contract IS terminal output, so the write error is explicitly discarded
// (one place, not per-line).
func printReport(report []string) {
	for _, line := range report {
		printLine(line)
	}
}

// printLine writes one reply line to stdout — the CLI's contract IS
// terminal output, so the write error is explicitly discarded here (one
// place, not four).
func printLine(reply string) {
	_, _ = fmt.Fprintln(os.Stdout, reply)
}

// runStatus prints the daemon state report — verbatim (the key=value canon
// is human-readable as printed) or as single-line JSON under --json.
func runStatus(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "single-line JSON output")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("status flags: %w", err)
	}

	reply, err := callMethod(ctx, "Status")
	if err != nil {
		return fmt.Errorf("status: %w", err)
	}
	if *asJSON {
		printLine(statusJSON(reply))

		return nil
	}
	printLine(reply)

	return nil
}

// callMethod performs one control call and returns the string reply. A
// name without an owner is the friendly daemon-not-running verdict; every
// other failure wraps the D-Bus error as-is.
func callMethod(ctx context.Context, method string) (string, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return "", fmt.Errorf("connect session bus: %w", err)
	}
	defer func() { _ = conn.Close() }()

	var reply string
	call := conn.Object(ctlBusName, dbus.ObjectPath(ctlPath)).
		CallWithContext(ctx, ctlBusName+"."+method, 0)
	if call.Err != nil {
		if isNoOwner(call.Err) {
			return "", errDaemonNotRunning
		}

		return "", fmt.Errorf("call %s: %w", method, call.Err)
	}
	if err := call.Store(&reply); err != nil {
		return "", fmt.Errorf("store %s reply: %w", method, err)
	}

	return reply, nil
}

// isNoOwner reports whether the call error is the bus's "no such name"
// verdict — the daemon is not running (or released the name on exit).
func isNoOwner(err error) bool {
	var derr dbus.Error
	if !errors.As(err, &derr) {
		return false
	}

	return derr.Name == "org.freedesktop.DBus.Error.ServiceUnknown" ||
		derr.Name == "org.freedesktop.DBus.Error.NameHasNoOwner"
}

// parseStatusLine splits the status canon into its key=value pairs: every
// token before config_error is one space-free pair; everything after the
// config_error key is that field's value (spaces allowed — the flattened
// parse error).
func parseStatusLine(line string) map[string]string {
	pairs := map[string]string{}
	head := line
	if i := strings.Index(line, configErrorKey); i >= 0 {
		pairs["config_error"] = line[i+len(configErrorKey):]
		head = line[:i]
	}
	for _, tok := range strings.Fields(head) {
		if k, v, ok := strings.Cut(tok, "="); ok {
			pairs[k] = v
		}
	}

	return pairs
}

// statusJSON renders the status line as single-line JSON — the --json
// contract. Values are typed best-effort: integers stay numbers, the
// boolean literals stay booleans, everything else is a string.
func statusJSON(line string) string {
	pairs := parseStatusLine(line)
	obj := make(map[string]any, len(pairs))
	for k, v := range pairs {
		obj[k] = jsonValue(v)
	}
	data, err := json.Marshal(obj)
	if err != nil {
		// A map of string keys to string/int/bool values cannot fail to
		// marshal; the branch exists so the error is never swallowed.
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}

	return string(data)
}

// jsonValue types one token value: number, boolean or string.
func jsonValue(v string) any {
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	if v == "true" {
		return true
	}
	if v == "false" {
		return false
	}

	return v
}
