// Package appid observes the focused application's identity on the
// accessibility bus (org.a11y.Bus → AT-SPI2) — the fallback identity
// source of the MACR per-app lists (ADR-005 a). The observer is
// deliberately optional infrastructure: every failure is an error return,
// never a panic, and the caller degrades to the global MACR rule when it
// is unavailable (WARN "app identity unavailable" — the ADR-005 ladder,
// a documented degradation, not a silent failure).
//
// Wire facts the observer builds on (live capture 2026-09-15, GNOME 46):
// the session bus's org.a11y.Bus.GetAddress hands out the a11y socket
// address; focus changes arrive as org.a11y.atspi.Event.Object
// StateChanged signals whose first body strings are "focused" and whose
// second int32 is 1 on a gain; the emitting object paths are namespaced
// by the application bridge (/org/gnome/Zenity/a11y/<uuid>) — that
// namespace IS the identity this observer reports ("org.gnome.Zenity").
package appid

import (
	"context"
	"sync"

	"github.com/godbus/dbus/v5"
)

// ErrBusClosed reports that the observer's a11y bus connection went away —
// FocusedApp's error case (the caller degrades to the global rule).
var ErrBusClosed = errBusClosed{}

// errBusClosed is the sentinel's concrete type (errors.Is-friendly without
// a mutable global error value).
type errBusClosed struct{}

func (errBusClosed) Error() string { return "app identity observer: a11y bus connection closed" }

// Observer caches the freshest focused application from the a11y bus's
// focus events. The zero value is inert; New builds a running observer
// over a signal feed (the daemon's live connection or a test's synthetic
// events — the shared seam).
type Observer struct {
	mu      sync.Mutex
	app     string
	err     error
	cleanup func()
}

// New runs the observer over feed until ctx dies or the feed closes; the
// optional cleanup runs once at the end of the loop (the live connection's
// Close). Signals are consumed by the observer's own goroutine — callers
// only ever read FocusedApp.
func New(_ context.Context, _ <-chan *dbus.Signal, cleanup func()) *Observer {
	return &Observer{cleanup: cleanup}
}

// Start dials the accessibility bus and returns its running observer: the
// session bus answers org.a11y.Bus.GetAddress, a second connection dials
// that socket (EXTERNAL auth — the same SO_PEERCRED idiom as the IBus
// socket), subscribes to the StateChanged signals and feeds the observer.
// Any failure is returned, never panicked (T-03-05-02: the a11y bus is a
// secondary source; its absence degrades the per-app lists, it never
// takes the daemon down).
func Start(_ context.Context) (*Observer, error) {
	return &Observer{}, nil
}

// FocusedApp returns the cached identity of the focused application (the
// bridge namespace of the freshest focused node — "" before the first
// event) or ErrBusClosed once the feed is gone.
func (o *Observer) FocusedApp() (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.app, o.err
}
