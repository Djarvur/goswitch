// Package a11y brings the desktop's accessibility state to the desired
// state of the config's a11y section (D-8-3, the daemon-reconcile): the
// global toolkit-accessibility key through gsettings (D-8-6, the owner's
// manual ZCode precedent) — a best-effort series of subprocesses that can
// never block, error or otherwise alter the fold that triggered it: the
// desired state is pinned and the series runs serialized on the
// reconciler's OWN mutex, outside the actor mutex (WR-01 — no subprocess
// ever rides the keystroke hot path); every failure is ONE WARN per
// episode with a closed reason and a quiet continuation, never an error
// to the caller. Deactivation never touches the desktop (D-8-4: a
// manually enabled key is indistinguishable from a goswitch-enabled one —
// the only revert is the uninstaller's). Application names never reach
// this package at all: the config list only decides WHETHER the fixed
// actions run, the argv is rendered from package literals alone
// (D-20/D-21, T-08-03-01 — the ASVS V5 argv discipline).
package a11y

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"
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

// The degradation reasons of the warn-once episodes (the closed vocabulary
// the corpus pins — reasons only, never user context, D-20/D-21).
const (
	reasonKeyRead = "key-read-failed"  // the gsettings get probe failed — the set still fires
	reasonKeySet  = "key-write-failed" // the gsettings set child failed to start
)

// readTimeout bounds the gsettings get probe (the clipboard cmdTimeout
// form): a wedged probe must fail the read, not hold the series.
const readTimeout = 1500 * time.Millisecond

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
// no-op (the diff gate keeps dconf churn and log noise at zero); the
// series itself re-checks the diff under the mutex, so a stale series
// converges to the newest desired state without races.
func (r *Reconciler) Apply(active bool) {
	r.mu.Lock()
	if active == r.desired {
		r.mu.Unlock()

		return
	}
	r.desired = active
	r.mu.Unlock()

	go r.run() // the series on the reconciler's own serialization — never the caller's stack
}

// run converges the desktop to the pinned desired state, the diff gate
// first: a deactivation touches NOTHING (D-8-4 — the key is never
// removed, zero subprocesses); an activation runs one apply episode. The
// whole series holds the reconciler's mutex — the serialization that
// keeps Apply(true)→Apply(false)→Apply(true) converging to the last
// state, with a stale series superseded before it fires a subprocess.
func (r *Reconciler) run() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.desired == r.lastApplied {
		return // the stale-series gate: a newer Apply supersedes this run
	}
	if !r.desired {
		r.lastApplied = false // zero subprocesses — deactivation never touches the desktop (D-8-4)

		return
	}
	r.episode()
}

// episode is one activation series: read-verify-then-set on the global
// key (research Pitfall 4 — the key already true means ZERO subprocess
// writes, no dconf churn), fail-toward-desired on a failed read, and the
// warn budget closed only by a clean activation. lastApplied lands true
// whatever the degradation: the desired state is pinned and the next
// desired change re-runs the series.
func (r *Reconciler) episode() {
	healthy := true

	value, err := r.read()
	if err != nil {
		r.warn(reasonKeyRead, err) // ONE WARN — the set still fires (fail toward desired)
		healthy = false
	}
	if err != nil || value != valTrue {
		if !r.setKey() {
			healthy = false
		}
	}
	r.lastApplied = true
	if healthy {
		r.closeEpisode() // a clean activation proves the stack — a LATER failure warns again
	}
}

// read probes the live key through the deadline-bounded Runner and trims
// the answer to the exact form gsettings prints ("true"/"false" — the
// install.go readSources trim discipline).
func (r *Reconciler) read() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	out, err := r.runner(ctx, binGSettings, []string{verbGet, schemaGnomeInterface, keyToolkitAccessibility})
	if err != nil {
		return "", fmt.Errorf("read %s: %w", keyToolkitAccessibility, err)
	}

	return strings.TrimSpace(string(out)), nil
}

// setKey fires the set through the Runner seam — in production the
// no-pipes child reaped by its own goroutine (newSetCmd/NewExecRunner). A
// failed start is one WARN and a false answer: the episode stays open.
func (r *Reconciler) setKey() bool {
	if _, err := r.runner(context.Background(), binGSettings, []string{verbSet, schemaGnomeInterface, keyToolkitAccessibility, valTrue}); err != nil {
		r.warn(reasonKeySet, err)

		return false
	}

	return true
}

// closeEpisode reopens the warn budget (the sound.go closeEpisode form):
// a healthy activation proves the gsettings stack — a LATER failure
// warns again. The caller holds the mutex.
func (r *Reconciler) closeEpisode() {
	clear(r.warned)
}

// warn records ONE degradation WARN per episode per reason (the
// sound.go warn form): the first failure of an episode warns, the rest
// stay quiet. The record names the closed reason and the transport error
// only — never user context (D-20/D-21). The caller holds the mutex.
func (r *Reconciler) warn(reason string, err error) {
	if r.warned[reason] {
		return
	}
	r.warned[reason] = true
	if err == nil {
		slog.Warn("a11y reconcile degraded", "reason", reason)

		return
	}
	slog.Warn("a11y reconcile degraded", "reason", reason, "error", err)
}

// newSetCmd builds one gsettings set child: the pinned binary with
// NOTHING attached to Stdin/Stdout/Stderr — the no-pipes fork shape (the
// wl-copy precedent, the sound.go newPlayProc form): a piped descriptor
// would make a reaper wait on a fork-shaped grandchild's write-ends.
func newSetCmd(name string, args ...string) *exec.Cmd { //nolint:ireturn // exec.Cmd is the pinned concrete child type (the newPlayProc precedent)
	//nolint:noctx // deadline-free by design — Start plus a reaped Wait; nothing kills a set (the sound.go form)
	return exec.Command(name, args...) // Stdin/Stdout/Stderr stay nil — no pipes
}

// NewExecRunner returns the production Runner (the name is the 08-05
// wiring contract — internal/a11y is not expected to change after this
// plan): reads go through exec.CommandContext (the caller's deadline
// SIGKILLs a wedged probe; stdout captured byte-exactly, stderr folded
// into the error — the clipboard execRunner form), writes are the
// no-pipes child reaped by its own goroutine (the sound.go reaped form —
// nothing kills a set, and a failed start is the returned error).
func NewExecRunner() Runner { //nolint:ireturn // the seam hands the runner closure back (the seam-type precedent)
	return func(ctx context.Context, name string, args []string) ([]byte, error) {
		if len(args) > 0 && args[0] == verbSet {
			cmd := newSetCmd(name, args...)
			if err := cmd.Start(); err != nil {
				return nil, fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
			}
			go func() { _ = cmd.Wait() }() // reap without a deadline — nothing kills a set

			return nil, nil
		}
		cmd := exec.CommandContext(ctx, name, args...)
		var out, errOut bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errOut
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
		}

		return out.Bytes(), nil
	}
}
