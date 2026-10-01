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
// bus service on the SESSION bus, the event the observer subscribes to on
// the a11y bus itself, and the Accessible role query the Role seam makes.
// getRoleMethod is pinned as the full wire literal: the answer is the uint32
// role enum (u), compared by NUMBER only.
const (
	a11yBusName   = "org.a11y.Bus"
	a11yBusPath   = "/org/a11y/bus"
	eventIface    = "org.a11y.atspi.Event.Object"
	stateChanged  = "StateChanged"
	getRoleMethod = "org.a11y.atspi.Accessible.GetRole"
	signalBufSize = 16
)

// ErrBusClosed reports that the observer's a11y bus connection went away —
// FocusedApp's error case (the caller degrades to the global rule).
var ErrBusClosed = busClosedError{}

// busClosedError is the sentinel's concrete type (errors.Is-friendly
// without a mutable global error value).
type busClosedError struct{}

func (busClosedError) Error() string { return "app identity observer: a11y bus connection closed" }

// ErrRoleUnknown reports that the observer cannot name the role of the
// focused object — no focus gain seen yet, no live role seam (the zero
// value / tests without Start), or the bus call failed. The D-53 policy
// treats every unknown as silence.
var ErrRoleUnknown = roleUnknownError{}

// roleUnknownError is the sentinel's concrete type (errors.Is-friendly
// without a mutable global error value) — the ErrBusClosed discipline.
type roleUnknownError struct{}

func (roleUnknownError) Error() string {
	return "app identity observer: focused object role unknown"
}

// The AT-SPI role enum values the autocorrect policy gates on (gi Atspi
// 2.52.0, live-verified on the target GNOME 46 desktop 2026-09-27 —
// 06-RESEARCH Q4). Roles compare by NUMBER only: role-name strings are
// localized and diverge between toolkits for the same enum value (61 is
// "text box" over the wire and "text" in the GI registry).
const (
	RolePasswordText uint32 = 40 // password text — the autocorrect policy never corrects it
	RoleTerminal     uint32 = 60 // terminal — never corrected
	RoleText         uint32 = 61 // text box / text — correctable text input
	RoleEntry        uint32 = 79 // entry — correctable text input
	RoleDocumentText uint32 = 94 // document text — correctable text input
)

// RoleCall is the observer's live role seam: one
// org.a11y.atspi.Accessible.GetRole round trip over the a11y bus per
// invocation, made at the caller's decision moment (D-53 — the role is a
// fresh answer, never a cache read). Start installs the real bus closure;
// tests inject recorders.
type RoleCall func(ctx context.Context, sender dbus.Sender, path dbus.ObjectPath) (uint32, error)

// Option customizes an Observer at construction time, before the run loop
// starts — the injection seam of New.
type Option func(*Observer)

// WithRoleCall installs the live role seam (Start's real GetRole closure or
// a test's recorder). Without it Role reports the unknown-role error — the
// zero value stays fail-closed (D-53: every unknown silences the consumer).
func WithRoleCall(fn RoleCall) Option {
	return func(o *Observer) { o.roleCall = fn }
}

// Observer caches the freshest focused application from the a11y bus's
// focus events, together with the (sender, path) address of the focused
// object — the destination of Role's live role query. The zero value is
// inert; New builds a running observer over a signal feed (the daemon's
// live connection or a test's synthetic events — the shared seam).
type Observer struct {
	mu          sync.Mutex
	app         string
	focusSender dbus.Sender
	focusPath   dbus.ObjectPath
	roleCall    RoleCall
	err         error
	cleanup     func()
}

// New runs the observer over feed until ctx dies or the feed closes; the
// optional cleanup runs once at the end of the loop (the live connection's
// Close). Signals are consumed by the observer's own goroutine — callers
// only ever read FocusedApp and Role.
func New(ctx context.Context, feed <-chan *dbus.Signal, cleanup func(), opts ...Option) *Observer {
	obs := &Observer{cleanup: cleanup}
	for _, opt := range opts {
		opt(obs)
	}
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

	// The live role seam: one GetRole round trip over the observer's own
	// connection per Role call — the GetAddress precedent's call shape
	// (CallWithContext().Store()), on the a11y object the focus event named.
	// The wire layer wraps the transport error (the seam's answer IS Role's
	// error); a dead object behind the path is an error return — the D-53
	// policy silences, never panics.
	return New(ctx, feed, func() { _ = conn.Close() }, WithRoleCall(
		func(ctx context.Context, sender dbus.Sender, path dbus.ObjectPath) (uint32, error) {
			var role uint32
			if err := conn.Object(string(sender), path).
				CallWithContext(ctx, getRoleMethod, 0).Store(&role); err != nil {
				return 0, fmt.Errorf("get role: %w", err)
			}

			return role, nil
		},
	)), nil
}

// FocusedApp returns the cached identity of the focused application (the
// bridge namespace of the freshest focused node — "" before the first
// event) or ErrBusClosed once the feed is gone.
func (o *Observer) FocusedApp() (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.app, o.err
}

// Role queries the AT-SPI role of the object behind the stored focus pair —
// one live GetRole round trip at the caller's decision moment (D-53: the
// stored pair is the address of the query, never a cached answer). The role
// comes back as the uint32 AT-SPI enum; any unknown (no gain seen, no live
// seam, a failed call) is an error the consumer's policy treats as silence —
// the seam's transport failure arrives wrapped by the wire layer, the
// unknown-role cases here by the sentinel. Role never logs — the decision
// and any warn-once discipline are the consumer's business.
func (o *Observer) Role(ctx context.Context) (uint32, error) {
	o.mu.Lock()
	sender, path := o.focusSender, o.focusPath
	o.mu.Unlock()

	if path == "" || o.roleCall == nil {
		return 0, fmt.Errorf("role: %w", ErrRoleUnknown)
	}

	return o.roleCall(ctx, sender, path)
}

// run is the observer's event loop: every focus GAIN on a bridge-namespaced
// path refreshes the identity cache and stores the (sender, path) address of
// the focused object in one critical section; losses and non-bridge paths
// keep both (the cache-keep semantics mirror FocusedApp — an event without a
// bridge identity carries nothing to refresh with). A closed feed or a
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
				o.focusSender = dbus.Sender(sig.Sender)
				o.focusPath = sig.Path
				o.mu.Unlock()
			}
		}
	}
}
