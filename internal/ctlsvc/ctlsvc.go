// Package ctlsvc is the daemon's control service on the D-Bus session bus
// (INST-02): a second godbus connection of the same process (the engine
// rides the private IBus socket, this one the session bus — the
// ARCHITECTURE bus map) that owns the well-known name org.djarvur.goswitch
// and exports the methods the goswitchctl client drives: Status,
// ReloadConfig and CorrectNow. The name request is the single-instance
// guard and the session bus's SO_PEERCRED isolation bounds every caller to
// the same uid (T-03-06-04); every exported method opens with a recover
// shim so a panic below a control call can never kill the daemon — the
// desktop's input rides the same process (INTEG-05 continuation).
package ctlsvc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"

	"github.com/godbus/dbus/v5"

	"github.com/Djarvur/goswitch/internal/session"
)

// Wire constants (ARCHITECTURE-pinned, the CLI's grep-pinned contract):
// the well-known name on the session bus and the object path the three
// control methods are exported on.
const (
	BusName    = "org.djarvur.goswitch"
	ObjectPath = dbus.ObjectPath("/org/djarvur/goswitch")
)

// ErrNotPrimaryOwner reports that another process of the same user already
// holds the control name — the single-instance guard's sentinel (the
// conn.go:24 idiom, checked with errors.Is by the daemon's caller).
var ErrNotPrimaryOwner = errors.New("another instance owns the goswitch control name")

// errFailed is the error name every ctl failure reply carries — the client
// is goswitchctl, an error reply is information, never a lifecycle event
// (unlike the IBus side, where a reply must stay zero-valued).
const errFailed = "org.freedesktop.DBus.Error.Failed"

// StatusSnapshotProvider reports the daemon state — the session actor's
// snapshot under its mutex (never a second FSM).
type StatusSnapshotProvider interface {
	StatusSnapshot() session.Status
}

// Reloader forces a synchronous config re-read — the watcher's Reload (the
// publication-or-last-good core shared with the debounce path, D-32).
type Reloader interface {
	Reload() (string, error)
}

// Corrector launches the forced word correction — the actor's CorrectNow
// (the Double decision's internal point, without a tap).
type Corrector interface {
	CorrectNow() string
}

// Deps carries the daemon surfaces the control methods drive — small
// single-method interfaces at the point of use. Reload may be nil: the
// no-config daemon reports "no config file" on ReloadConfig.
type Deps struct {
	Status  StatusSnapshotProvider
	Reload  Reloader
	Correct Corrector
}

// Svc is the object exported at ObjectPath on BusName; godbus dispatches
// each method call on its own goroutine, so the deps interfaces re-enter
// the actor under its own mutex (T-03-06-02: the shim contains whatever
// races or panics that invites).
type Svc struct {
	deps Deps
}

// NewSvc builds the exported control object.
func NewSvc(deps Deps) *Svc {
	return &Svc{deps: deps}
}

// Status implements org.djarvur.goswitch.Status () → s: the daemon state
// as the single-line key=value report (renderStatus — the wire canon of
// the status surface, counts and states only, T-03-06-03).
func (s *Svc) Status() (reply string, err *dbus.Error) {
	defer recoverMethod("Status", &err)

	return renderStatus(s.deps.Status.StatusSnapshot()), nil
}

// ReloadConfig implements org.djarvur.goswitch.ReloadConfig () → s: the
// forced synchronous re-read. A valid document applies (the reply names
// the change); a rejected one keeps the last-good serving and the reply
// error carries the rejection (D-32 — the same error is then visible in
// Status). A nil Reloader is the no-config daemon: nothing to reread.
func (s *Svc) ReloadConfig() (reply string, err *dbus.Error) {
	defer recoverMethod("ReloadConfig", &err)

	if s.deps.Reload == nil {
		return "", &dbus.Error{Name: errFailed, Body: []any{"no config file (built-in defaults in force)"}}
	}
	msg, reloadErr := s.deps.Reload.Reload()
	if reloadErr != nil {
		return "", &dbus.Error{Name: errFailed, Body: []any{reloadErr.Error()}}
	}

	return msg, nil
}

// CorrectNow implements org.djarvur.goswitch.CorrectNow () → s: the forced
// word correction (the Q5 word semantics) with its immediate
// acknowledgment — the pipeline's settlement lands in the counters and the
// log, the reply never waits for it.
func (s *Svc) CorrectNow() (reply string, err *dbus.Error) {
	defer recoverMethod("CorrectNow", &err)

	return s.deps.Correct.CorrectNow(), nil
}

// renderStatus formats the snapshot as the Status method's single-line
// key=value report — the wire canon: human-readable as printed, machine-
// parseable by goswitchctl's --json. The build version (D-37) LEADS the
// line, so a bug report's first token identifies the build; the
// config_error token is always LAST and its value whitespace-flattened
// (a multi-line parse error must not break the token grammar); skip
// reasons flatten their dashes so every key is one token. Counts and
// states only — never user text (T-03-06-03).
func renderStatus(st session.Status) string {
	tokens := []string{
		"version=" + st.Version,
		"mode=" + st.Mode,
		"corrections_done=" + strconv.Itoa(st.CorrectionsDone),
		"corrections_skipped=" + strconv.Itoa(st.CorrectionsSkipped),
	}
	for _, reason := range slices.Sorted(maps.Keys(st.SkipReasons)) {
		tokens = append(tokens, "skip_"+strings.ReplaceAll(reason, "-", "_")+
			"="+strconv.Itoa(st.SkipReasons[reason]))
	}
	tokens = append(tokens,
		"super_intercepted="+strconv.Itoa(st.SuperIntercepted),
		"super_upstream_consumed="+strconv.Itoa(st.SuperUpstreamConsumed),
	)
	if st.ConfigPath == "" {
		tokens = append(tokens, "config=none")
	} else {
		tokens = append(tokens,
			"config_path="+st.ConfigPath,
			"config_valid="+strconv.FormatBool(st.ConfigValid),
			"config_error="+strings.Join(strings.Fields(st.ConfigError), " "),
		)
	}

	return strings.Join(tokens, " ")
}

// recoverMethod contains a panic raised anywhere below a control method:
// the panic is swallowed and logged at ERROR with its stack, and the
// caller answers with a D-Bus error instead (T-03-06-02, the INTEG-05
// continuation — the daemon keeps serving desktop input through the
// engine connection whatever a control call broke).
func recoverMethod(method string, dbusErr **dbus.Error) {
	if r := recover(); r != nil {
		slog.Error("ctl method panic contained",
			"method", method,
			"panic", fmt.Sprint(r),
			"stack", string(debug.Stack()))
		*dbusErr = &dbus.Error{Name: errFailed, Body: []any{"internal error in " + method}}
	}
}

// Run serves the control methods on the session bus until ctx is
// cancelled (a cancelled context is a clean shutdown: Run returns nil and
// the deferred Close releases the name). The name request is the
// single-instance guard: only a RequestNameReplyPrimaryOwner reply lets
// the service serve — anything else is ErrNotPrimaryOwner (V4/T-03-06-01;
// the flags mirror conn.go:93-101, where no AllowReplacement on our side
// means a second live instance can never replace us, and a dead
// connection's name auto-release covers the restart case).
func Run(ctx context.Context, deps Deps) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("connect session bus: %w", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			slog.Warn("closing ctl connection", "error", closeErr)
		}
	}()

	reply, err := conn.RequestName(BusName, dbus.NameFlagReplaceExisting)
	if err != nil {
		return fmt.Errorf("request name %s: %w", BusName, err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		slog.Error("another goswitchd instance already registered", "reply", int(reply))

		return fmt.Errorf("%w: got reply %d", ErrNotPrimaryOwner, int(reply))
	}
	if err := conn.Export(NewSvc(deps), ObjectPath, BusName); err != nil {
		return fmt.Errorf("export control object: %w", err)
	}
	slog.Info("ctl service listening", "name", BusName)

	<-ctx.Done()

	return nil
}
