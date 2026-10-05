// Package a11y brings the desktop's accessibility state to the desired
// state of the config's a11y section (D-8-3, the daemon-reconcile): the
// global toolkit-accessibility key through gsettings (D-8-6, the owner's
// manual ZCode precedent) — a best-effort series of subprocesses that can
// never block, error or otherwise alter the fold that triggered it: the
// desired state is pinned and the series runs serialized on the
// reconciler's OWN mutex, outside the actor mutex (WR-01); every failure
// is ONE WARN per episode with a closed reason and a quiet continuation.
// Deactivation never touches the desktop (D-8-4: a manually enabled key
// is indistinguishable from a goswitch-enabled one — the only revert is
// the uninstaller's). Application names never reach this package at all:
// the config list only decides WHETHER the fixed actions run, the argv is
// rendered from package literals alone (D-20/D-21, T-08-03-01).
package a11y

import (
	"context"
	"os/exec"
	"sync"
)

// The pinned gsettings coordinates of the global a11y key (D-8-6: the
// exact lever the owner enabled manually for ZCode) and the two verbs.
// The reconcile argv is rendered ONLY from these literals — the config's
// app list never becomes an argument (T-08-03-01, the install.go
// argv-discipline canon).
const (
	binGSettings            = "gsettings"
	schemaGnomeInterface    = "org.gnome.desktop.interface"
	keyToolkitAccessibility = "toolkit-accessibility"
	valTrue                 = "true"
	verbGet                 = "get"
	verbSet                 = "set"
)

// Runner executes one gsettings subprocess: the seam the unit corpus
// drives with a recording fake (the argv is the observable surface),
// backed in production by NewExecRunner. Reads are deadline-bounded with
// the stdout captured; writes are the no-pipes fire-and-forget form.
type Runner func(ctx context.Context, name string, args []string) ([]byte, error)

// StatusSetter sets the session-scoped org.a11y.Status.IsEnabled belt —
// the seam the belt corpus drives with a recording fake (plan 08-03
// Task 2 lands the call site and the production godbus adapter).
type StatusSetter func(ctx context.Context) error

// Reconciler serializes the a11y series: the desired state is pinned by
// Apply, the apply series runs on the reconciler's own mutex — a one-way
// ordering with the actor's (the sink never calls the actor back, WR-01)
// — so a config fold never waits for a subprocess.
type Reconciler struct {
	runner Runner
	status StatusSetter

	mu          sync.Mutex      // the series' serialization — never the actor's mutex
	desired     bool            // the pinned desired state (the zero value is off, D-54)
	lastApplied bool            // the diff gate: the desired state the series last converged to
	warned      map[string]bool // one WARN per episode per reason — a healthy activation reopens
}

// New returns the reconciler over the given seams: NewExecRunner for the
// gsettings series and NewDBusStatusSetter for the belt are the
// production pair (the names are the 08-05 wiring contract).
func New(runner Runner, status StatusSetter) *Reconciler {
	return &Reconciler{
		runner: runner,
		status: status,
		warned: make(map[string]bool),
	}
}

// Apply pins the desired a11y state and spawns the apply series —
// fire-and-forget by contract: the fold's caller never waits for a
// subprocess (WR-01). A repeated Apply with the same desired state is a
// no-op; the series itself re-checks the diff under the mutex, so a
// stale series converges to the newest desired state without races.
func (r *Reconciler) Apply(active bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if active == r.desired {
		return
	}
	r.desired = active
}

// newSetCmd builds one gsettings set child: the pinned binary with
// NOTHING attached to Stdin/Stdout/Stderr — the no-pipes fork shape (the
// wl-copy precedent, the sound.go newPlayProc form): a piped descriptor
// would make a reaper wait on a fork-shaped grandchild's write-ends.
func newSetCmd(name string, args ...string) *exec.Cmd { //nolint:ireturn // exec.Cmd is the pinned concrete child type (the newPlayProc precedent)
	//nolint:noctx // deadline-free by design — Start plus a reaped Wait; nothing kills a set (the sound.go form)
	return exec.Command(name, args...) // Stdin/Stdout/Stderr stay nil — no pipes
}
