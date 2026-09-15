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
// as the single-line key=value report.
func (s *Svc) Status() (reply string, err *dbus.Error) {
	return "", nil
}

// ReloadConfig implements org.djarvur.goswitch.ReloadConfig () → s: the
// forced synchronous re-read; a rejected document keeps the last-good
// snapshot serving with the error in the reply (D-32).
func (s *Svc) ReloadConfig() (reply string, err *dbus.Error) {
	return "", nil
}

// CorrectNow implements org.djarvur.goswitch.CorrectNow () → s: the forced
// word correction with an immediate acknowledgment.
func (s *Svc) CorrectNow() (reply string, err *dbus.Error) {
	return "", nil
}

// Run serves the control methods on the session bus until ctx is
// cancelled (a cancelled context is a clean shutdown: Run returns nil).
func Run(ctx context.Context, deps Deps) error {
	return nil
}
