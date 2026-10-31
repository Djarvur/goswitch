package appid_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/Djarvur/goswitch/internal/appid"
)

// errSeamDead is the failing seam's sentinel of the LiveCallShape corpus
// (package level — the errors.New discipline).
var errSeamDead = errors.New("app identity test seam: dead")

// appidObserverWait budgets one synthetic event's propagation through the
// observer's own goroutine (the poll-until idiom — never a fixed sleep).
const appidObserverWait = 2 * time.Second

// The corpus's synthetic sender/path pairs — the live wire shapes (the
// bridge namespace /org/gnome/<App>/a11y/<uuid> IS the identity; the sender
// is the emitting connection's unique bus name, delivered as a plain string
// by godbus and stored typed by the observer).
const (
	senderZenity = dbus.Sender(":1.90")
	senderEditor = dbus.Sender(":1.91")

	zenityPath = "/org/gnome/Zenity/a11y/6c20c914_1a35"
	editorPath = "/org/gnome/gnome-text-editor/a11y/n1"
)

// focusSignal builds one a11y StateChanged("focused") signal of the live
// wire shape (capture 2026-09-15): interface org.a11y.atspi.Event.Object,
// member StateChanged, body ("focused", gain, detail, properties). The
// sender rides the signal like the bus delivers it — a plain string.
func focusSignal(sender dbus.Sender, path string, gain int32) *dbus.Signal {
	return &dbus.Signal{
		Sender: string(sender),
		Path:   dbus.ObjectPath(path),
		Name:   "org.a11y.atspi.Event.Object.StateChanged",
		Body:   []any{"focused", gain, int32(0), dbus.MakeVariant("0")},
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

		feed <- focusSignal(senderZenity, zenityPath, 1)
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

		feed <- focusSignal(senderEditor, editorPath, 1)
		if !waitApp(t, obs, "org.gnome.gnome-text-editor") {
			t.Fatalf("first identity never landed")
		}
		feed <- focusSignal(senderEditor, editorPath, 0) // loss
		feed <- focusSignal(senderEditor, "/root", 1)    // not bridge-namespaced
		time.Sleep(50 * time.Millisecond)                // let the loop see both
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

// TestObserver_StoresFocusPair pins the address storage of the D-53 role
// policy: a focus GAIN stores (sender, path) beside the app name under the
// same mutex, and the injected Role seam sees exactly that pair.
func TestObserver_StoresFocusPair(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	feed := make(chan *dbus.Signal, 4)

	var (
		gotSender dbus.Sender
		gotPath   dbus.ObjectPath
		called    bool
	)
	obs := appid.New(ctx, feed, nil, appid.WithRoleCall(
		func(_ context.Context, sender dbus.Sender, path dbus.ObjectPath) (uint32, error) {
			called = true
			gotSender, gotPath = sender, path

			return 0, nil
		},
	))

	feed <- focusSignal(senderZenity, zenityPath, 1)
	if !waitApp(t, obs, "org.gnome.Zenity") {
		app, err := obs.FocusedApp()
		t.Fatalf("focused identity = %q, err %v; want org.gnome.Zenity", app, err)
	}
	if _, err := obs.Role(ctx); err != nil {
		t.Fatalf("Role over the stored pair errored: %v", err)
	}
	if !called || gotSender != senderZenity || gotPath != dbus.ObjectPath(zenityPath) {
		t.Errorf("role seam saw (%q, %q) called=%v, want the gain's (%q, %q)",
			gotSender, gotPath, called, senderZenity, zenityPath)
	}
}

// TestObserver_LossKeepsPair pins the cache-keep semantics of the stored
// pair (the mirror of FocusedApp): a loss (state 0) and a non-bridge gain
// leave the stored address alone — a later Role still asks the LAST gain's
// object.
func TestObserver_LossKeepsPair(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	feed := make(chan *dbus.Signal, 4)

	var gotPath dbus.ObjectPath
	obs := appid.New(ctx, feed, nil, appid.WithRoleCall(
		func(_ context.Context, _ dbus.Sender, path dbus.ObjectPath) (uint32, error) {
			gotPath = path

			return 0, nil
		},
	))

	feed <- focusSignal(senderZenity, zenityPath, 1)
	if !waitApp(t, obs, "org.gnome.Zenity") {
		t.Fatalf("first identity never landed")
	}
	feed <- focusSignal(senderZenity, zenityPath, 0) // loss
	feed <- focusSignal(senderZenity, "/root", 1)    // not bridge-namespaced
	time.Sleep(50 * time.Millisecond)                // let the loop see both (the documented drain)
	_, _ = obs.Role(ctx)
	if gotPath != dbus.ObjectPath(zenityPath) {
		t.Errorf("role seam path after loss/non-bridge = %q, want the kept gain pair %q",
			gotPath, zenityPath)
	}
}

// TestObserver_FreshObserverUnknown pins the fail-closed zero value: Role on
// an observer that never saw a gain reports the unknown-role error — never
// (0, nil) — and the seam is not consulted for an answer it cannot own.
func TestObserver_FreshObserverUnknown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	feed := make(chan *dbus.Signal, 4)

	called := false
	obs := appid.New(ctx, feed, nil, appid.WithRoleCall(
		func(_ context.Context, _ dbus.Sender, _ dbus.ObjectPath) (uint32, error) {
			called = true

			return 61, nil
		},
	))

	role, err := obs.Role(ctx)
	if err == nil {
		t.Fatalf("Role on a fresh observer = (%d, nil), want an unknown-role error", role)
	}
	if role != 0 {
		t.Errorf("role value on the unknown error = %d, want 0", role)
	}
	if called {
		t.Errorf("role seam consulted on a fresh observer, want the unknown refused first")
	}
}

// TestRole_EnumConstants pins the live-verified AT-SPI role enum values (gi
// Atspi 2.52.0, target GNOME 46 desktop, 2026-09-27 — 06-RESEARCH Q4): the
// D-53 policy compares by NUMBER only, because role-name strings are
// localized and diverge between toolkits for the same enum value.
func TestRole_EnumConstants(t *testing.T) {
	for _, tc := range []struct {
		name string
		got  uint32
		want uint32
	}{
		{"RolePasswordText", appid.RolePasswordText, 40},
		{"RoleTerminal", appid.RoleTerminal, 60},
		{"RoleText", appid.RoleText, 61},
		{"RoleEntry", appid.RoleEntry, 79},
		{"RoleDocumentText", appid.RoleDocumentText, 94},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %d, want %d (gi Atspi 2.52.0 live pin)", tc.name, tc.got, tc.want)
		}
	}
}

// TestRole_EmptyPathUnknown pins the unknown-role sentinel on the empty
// address: no gain at all, or a gain on a path without the bridge marker —
// either way the pair is the zero value and Role reports ErrRoleUnknown
// (fail-closed: the consumer's policy stays silent).
func TestRole_EmptyPathUnknown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	feed := make(chan *dbus.Signal, 4)

	obs := appid.New(ctx, feed, nil, appid.WithRoleCall(
		func(_ context.Context, _ dbus.Sender, _ dbus.ObjectPath) (uint32, error) {
			return 61, nil
		},
	))

	if _, err := obs.Role(ctx); !errors.Is(err, appid.ErrRoleUnknown) {
		t.Fatalf("Role on a fresh observer = %v, want ErrRoleUnknown", err)
	}

	feed <- focusSignal(senderZenity, "/root", 1) // a gain without the bridge marker
	time.Sleep(50 * time.Millisecond)
	if _, err := obs.Role(ctx); !errors.Is(err, appid.ErrRoleUnknown) {
		t.Errorf("Role after a non-bridge gain = %v, want ErrRoleUnknown", err)
	}
}

// TestRole_LiveCallShape pins the live-query shape (D-53): the injected seam
// is called with exactly the caller's ctx and the (sender, path) pair of the
// last gain, its enum answer IS Role's answer, and a seam failure keeps its
// identity through Role's error (errors.Is).
func TestRole_LiveCallShape(t *testing.T) {
	const roleText = 61

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	feed := make(chan *dbus.Signal, 4)

	var (
		gotCtx    context.Context
		gotSender dbus.Sender
		gotPath   dbus.ObjectPath
		calls     int
	)
	obs := appid.New(ctx, feed, nil, appid.WithRoleCall(
		//nolint:fatcontext // the seam's ctx param IS the assertion: ctx must reach the wire call (D-53)
		func(seamCtx context.Context, sender dbus.Sender, path dbus.ObjectPath) (uint32, error) {
			calls++
			gotCtx, gotSender, gotPath = seamCtx, sender, path

			return roleText, nil
		},
	))

	feed <- focusSignal(senderEditor, editorPath, 1)
	if !waitApp(t, obs, "org.gnome.gnome-text-editor") {
		t.Fatalf("identity never landed")
	}

	role, err := obs.Role(ctx)
	if err != nil || role != roleText {
		t.Fatalf("Role = (%d, %v), want (%d, nil)", role, err, roleText)
	}
	if calls != 1 || gotCtx != ctx || gotSender != senderEditor || gotPath != dbus.ObjectPath(editorPath) {
		t.Errorf("seam call (ctx=%v, %q, %q) calls=%d, want exactly (caller ctx, %q, %q)",
			gotCtx != nil, gotSender, gotPath, calls, senderEditor, editorPath)
	}

	// A synthetic feed serves exactly one observer — the failing seam gets
	// its own.
	deadFeed := make(chan *dbus.Signal, 4)
	dead := appid.New(ctx, deadFeed, nil, appid.WithRoleCall(
		func(_ context.Context, _ dbus.Sender, _ dbus.ObjectPath) (uint32, error) {
			return 0, errSeamDead
		},
	))
	deadFeed <- focusSignal(senderEditor, editorPath, 1)
	if !waitApp(t, dead, "org.gnome.gnome-text-editor") {
		t.Fatalf("failing seam's identity never landed")
	}
	if _, err := dead.Role(ctx); !errors.Is(err, errSeamDead) {
		t.Errorf("Role over a failing seam = %v, want the seam error (errors.Is)", err)
	}
}

// TestRole_NilSeamUnknown pins the fail-closed zero seam: an observer built
// by New without Start has no live role seam — even with a stored gain pair,
// Role reports ErrRoleUnknown (the zero value = the consumer stays silent).
func TestRole_NilSeamUnknown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	feed := make(chan *dbus.Signal, 4)

	obs := appid.New(ctx, feed, nil) // no WithRoleCall: the zero seam

	feed <- focusSignal(senderZenity, zenityPath, 1)
	if !waitApp(t, obs, "org.gnome.Zenity") {
		t.Fatalf("identity never landed")
	}

	if _, err := obs.Role(ctx); !errors.Is(err, appid.ErrRoleUnknown) {
		t.Fatalf("Role without a live seam = %v, want ErrRoleUnknown", err)
	}
}

// TestRole_ConcurrentWithGain drives concurrent focus gains against
// concurrent Role queries under -race: the pair snapshot and the seam call
// never race, the stored address never empties (cache-keep), and every
// query over a stored pair succeeds.
func TestRole_ConcurrentWithGain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	feed := make(chan *dbus.Signal, 128)

	obs := appid.New(ctx, feed, nil, appid.WithRoleCall(
		func(_ context.Context, _ dbus.Sender, _ dbus.ObjectPath) (uint32, error) {
			return 61, nil
		},
	))

	feed <- focusSignal(senderZenity, zenityPath, 1) // pre-store: cache-keep keeps the pair non-empty
	if !waitApp(t, obs, "org.gnome.Zenity") {
		t.Fatalf("identity never landed")
	}

	paths := [2]string{zenityPath, editorPath}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := range 50 {
			feed <- focusSignal(senderZenity, paths[i%2], 1)
		}
	}()
	go func() {
		defer wg.Done()
		for range 50 {
			if _, err := obs.Role(ctx); err != nil {
				t.Errorf("concurrent Role = %v, want the seam answer", err)
			}
		}
	}()
	wg.Wait()

	if role, err := obs.Role(ctx); err != nil || role != 61 {
		t.Fatalf("final Role = (%d, %v), want (61, nil)", role, err)
	}
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

// TestObserver_GainHook pins the freshness hook of the blocklist menu item
// (phase 11, MENU-01): every focus GAIN delivers the identity to the
// installed hook — the actor's PushFocusedApp lives on this seam — a LOSS
// never calls it (freshness rides gains only), and the hook may read the
// observer back WITHOUT deadlocking: the call happens OUTSIDE the observer's
// mutex (Pitfall 9 — the hook under o.mu entering the actor's mutex would
// invert the one-way actor.mu → sink.mu ordering). The -race corpus of the
// three-package verify watches every one of these crossings.
func TestObserver_GainHook(t *testing.T) {
	t.Run("gain delivers the identity to the hook", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		feed := make(chan *dbus.Signal, 4)
		got := make(chan string, 1)
		appid.New(ctx, feed, nil, appid.WithGainHook(func(name string) {
			got <- name
		}))

		feed <- focusSignal(senderZenity, zenityPath, 1)
		select {
		case name := <-got:
			if name != "org.gnome.Zenity" {
				t.Errorf("gain hook name = %q, want org.gnome.Zenity", name)
			}
		case <-time.After(appidObserverWait):
			t.Fatalf("the gain hook never fired for a bridge-namespaced gain")
		}
	})

	t.Run("loss events never call the hook", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		feed := make(chan *dbus.Signal, 4)
		got := make(chan string, 1)
		appid.New(ctx, feed, nil, appid.WithGainHook(func(name string) {
			got <- name
		}))

		feed <- focusSignal(senderEditor, editorPath, 0) // a loss
		feed <- focusSignal(senderEditor, "/root", 1)    // not bridge-namespaced — no identity
		time.Sleep(50 * time.Millisecond)                // let the loop see both
		if len(got) != 0 {
			t.Errorf("the hook fired %d time(s) on loss/non-identity events, want never", len(got))
		}
	})

	t.Run("hook may read the observer without deadlocking", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		feed := make(chan *dbus.Signal, 4)
		var obs *appid.Observer
		seen := make(chan string, 1)
		obs = appid.New(ctx, feed, nil, appid.WithGainHook(func(name string) {
			// Pitfall 9: the hook runs OUTSIDE the observer's mutex — a read
			// of the observer FROM the hook must return, never deadlock.
			if app, err := obs.FocusedApp(); err == nil {
				seen <- app
			}
		}))

		feed <- focusSignal(senderEditor, editorPath, 1)
		select {
		case app := <-seen:
			if app != "org.gnome.gnome-text-editor" {
				t.Errorf("observer read from the hook = %q, want the fresh gain's identity", app)
			}
		case <-time.After(appidObserverWait):
			t.Fatalf("the hook blocked reading the observer — the mutex inversion of Pitfall 9")
		}
	})
}
