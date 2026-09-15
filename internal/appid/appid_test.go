package appid_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/Djarvur/goswitch/internal/appid"
)

// appidObserverWait budgets one synthetic event's propagation through the
// observer's own goroutine (the poll-until idiom — never a fixed sleep).
const appidObserverWait = 2 * time.Second

// focusSignal builds one a11y StateChanged("focused") signal of the live
// wire shape (capture 2026-09-15): interface org.a11y.atspi.Event.Object,
// member StateChanged, body ("focused", gain, detail, properties).
func focusSignal(path string, gain int32) *dbus.Signal {
	return &dbus.Signal{
		Path: dbus.ObjectPath(path),
		Name: "org.a11y.atspi.Event.Object.StateChanged",
		Body: []any{"focused", gain, int32(0), dbus.MakeVariant("0")},
	}
}

// waitApp polls the observer until its identity equals want or the
// deadline passes.
func waitApp(t *testing.T, obs *appid.Observer, want string) bool {
	t.Helper()

	deadline := time.Now().Add(appidObserverWait)
	for {
		got, err := obs.FocusedApp()
		if err == nil && got == want {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// TestAppid_FocusedAppByFakeBus pins the observer over a synthetic feed
// (the A4 seam): a focused=1 event on a bridge-namespaced path caches the
// application identity; a loss (focused=0) and a non-bridge path leave the
// cache alone; a closed feed turns FocusedApp into ErrBusClosed — errors,
// never panics.
func TestAppid_FocusedAppByFakeBus(t *testing.T) {
	t.Run("focused gain caches the bridge identity", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		feed := make(chan *dbus.Signal, 4)
		obs := appid.New(ctx, feed, nil)

		feed <- focusSignal("/org/gnome/Zenity/a11y/6c20c914_1a35", 1)
		if !waitApp(t, obs, "org.gnome.Zenity") {
			app, err := obs.FocusedApp()
			t.Fatalf("focused identity = %q, err %v; want org.gnome.Zenity", app, err)
		}
	})

	t.Run("loss and non-bridge paths keep the cache", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		feed := make(chan *dbus.Signal, 4)
		obs := appid.New(ctx, feed, nil)

		feed <- focusSignal("/org/gnome/gnome-text-editor/a11y/n1", 1)
		if !waitApp(t, obs, "org.gnome.gnome-text-editor") {
			t.Fatalf("first identity never landed")
		}
		feed <- focusSignal("/org/gnome/gnome-text-editor/a11y/n1", 0) // loss
		feed <- focusSignal("/root", 1)                                // not bridge-namespaced
		time.Sleep(50 * time.Millisecond)                              // let the loop see both
		if got, _ := obs.FocusedApp(); got != "org.gnome.gnome-text-editor" {
			t.Errorf("identity after loss/non-bridge events = %q, want the cached one", got)
		}
	})

	t.Run("closed feed is ErrBusClosed, not a panic", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cleaned := make(chan struct{})
		feed := make(chan *dbus.Signal)
		obs := appid.New(ctx, feed, func() { close(cleaned) })

		close(feed)
		select {
		case <-cleaned:
		case <-time.After(appidObserverWait):
			t.Fatalf("cleanup never ran after the feed closed")
		}
		if _, err := obs.FocusedApp(); !errors.Is(err, appid.ErrBusClosed) {
			t.Errorf("FocusedApp error after close = %v, want ErrBusClosed", err)
		}
	})
}

// TestAppid_StartBusUnavailable pins the no-panic contract of the live
// dialer: with no reachable session bus, Start returns an error — the
// ADR-005 degradation ladder's raw material.
func TestAppid_StartBusUnavailable(t *testing.T) {
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/nonexistent/goswitch-appid-test-bus")

	if _, err := appid.Start(context.Background()); err == nil {
		t.Fatalf("Start over an unreachable bus = nil error, want the dial failure (never a panic)")
	}
}
