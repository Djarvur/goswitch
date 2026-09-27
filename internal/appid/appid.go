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
// StateChanged signals whose first body string is "focused" and whose
// second int32 is 1 on a gain; the emitting object paths are namespaced
// by the application bridge (/org/gnome/Zenity/a11y/<uuid>) — that
// namespace IS the identity this observer reports ("org.gnome.Zenity",
// the vocabulary macr.apps matches and goswitchctl will display).
package appid

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
)

// The a11y bus's D-Bus coordinates: the well-known name and object of the
// bus service on the SESSION bus, and the event the observer subscribes
// to on the a11y bus itself.
const (
	a11yBusName   = "org.a11y.Bus"
	a11yBusPath   = "/org/a11y/bus"
	eventIface    = "org.a11y.atspi.Event.Object"
	stateChanged  = "StateChanged"
	signalBufSize = 16
)

// ErrBusClosed reports that the observer's a11y bus connection went away —
// FocusedApp's error case (the caller degrades to the global rule).
var ErrBusClosed = busClosedError{}

// busClosedError is the sentinel's concrete type (errors.Is-friendly
// without a mutable global error value).
type busClosedError struct{}

func (busClosedError) Error() string { return "app identity observer: a11y bus connection closed" }

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
func New(ctx context.Context, feed <-chan *dbus.Signal, cleanup func()) *Observer {
	obs := &Observer{cleanup: cleanup}
	go obs.run(ctx, feed)

	return obs
}

// focusGain extracts the application identity from one StateChanged
// signal: a "focused" GAIN on a bridge-namespaced path. Losses (state 0)
// and paths without the bridge marker report gained=false — the cache
// keeps the last known identity through them (an unmarked node carries no
// identity to refresh with).
func focusGain(sig *dbus.Signal) (string, bool) {
	if sig.Name != eventIface+"."+stateChanged || len(sig.Body) < 2 {
		return "", false
	}
	kind, ok := sig.Body[0].(string)
	if !ok || kind != "focused" {
		return "", false
	}
	state, ok := sig.Body[1].(int32)
	if !ok || state != 1 {
		return "", false
	}
	name := bridgeApp(sig.Path)

	return name, name != ""
}

// bridgeApp maps a bridge-namespaced accessible path to the application
// identity: /org/gnome/Zenity/a11y/<uuid> → org.gnome.Zenity. Paths
// without the bridge marker (older atk trees' /root nodes) report "" —
// the observer only tracks bridge-marked applications (the documented
// A4 scope; an unmarked app leaves the cache at the last known value).
func bridgeApp(path dbus.ObjectPath) string {
	s := string(path)
	i := strings.Index(s, "/a11y/")
	if i <= 0 {
		return ""
	}

	return strings.ReplaceAll(s[1:i], "/", ".")
}

// Start dials the accessibility bus and returns its running observer: the
// session bus answers org.a11y.Bus.GetAddress, a second connection dials
// that socket (EXTERNAL auth with the decimal uid — the same SO_PEERCRED
// idiom as the IBus socket, engine/conn.go), subscribes to the
// StateChanged signals and feeds the observer. Any failure is returned,
// never panicked (T-03-05-02: the a11y bus is a secondary source; its
// absence degrades the per-app lists, it never takes the daemon down).
// The connection lives until ctx dies (the daemon's lifetime).
func Start(ctx context.Context) (*Observer, error) {
	sess, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("connect session bus: %w", err)
	}
	defer func() { _ = sess.Close() }() // the address round trip only — the observer never keeps it

	var addr string
	if err := sess.Object(a11yBusName, a11yBusPath).
		CallWithContext(ctx, a11yBusName+".GetAddress", 0).Store(&addr); err != nil {
		return nil, fmt.Errorf("get a11y bus address: %w", err)
	}

	conn, err := dbus.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("dial a11y bus: %w", err)
	}
	if err := conn.Auth([]dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("a11y dbus auth: %w", err)
	}
	if err := conn.Hello(); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("a11y dbus hello: %w", err)
	}
	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface(eventIface),
		dbus.WithMatchMember(stateChanged),
	); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("subscribe to a11y focus events: %w", err)
	}

	feed := make(chan *dbus.Signal, signalBufSize)
	conn.Signal(feed)

	return New(ctx, feed, func() { _ = conn.Close() }), nil
}

// FocusedApp returns the cached identity of the focused application (the
// bridge namespace of the freshest focused node — "" before the first
// event) or ErrBusClosed once the feed is gone.
func (o *Observer) FocusedApp() (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.app, o.err
}

// run is the observer's event loop: every focus GAIN on a
// bridge-namespaced path refreshes the identity cache; a closed feed or a
// cancelled context retires the observer with ErrBusClosed and runs the
// cleanup exactly once (the conn.go:132-145 loop idiom).
func (o *Observer) run(ctx context.Context, feed <-chan *dbus.Signal) {
	defer func() {
		if o.cleanup != nil {
			o.cleanup()
		}
		o.mu.Lock()
		o.err = ErrBusClosed
		o.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case sig, ok := <-feed:
			if !ok {
				return
			}
			if name, gained := focusGain(sig); gained {
				o.mu.Lock()
				o.app = name
				o.mu.Unlock()
			}
		}
	}
}
