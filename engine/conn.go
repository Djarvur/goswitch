package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"strconv"
	"time"

	"github.com/godbus/dbus/v5"
)

// IBus bus constants.
const (
	ibusService = "org.freedesktop.IBus"
	ibusPath    = dbus.ObjectPath("/org/freedesktop/IBus")
)

// errNotPrimaryOwner reports that another goswitchd already holds the
// component's well-known name — the single-instance guard (T-01-03).
var errNotPrimaryOwner = errors.New("another instance owns the goswitch name")

// errBusClosed reports that the private IBus connection went away
// (ibus-daemon exit or restart).
var errBusClosed = errors.New("ibus bus connection closed")

// Config carries everything Run needs: the component identity, the engines
// to register and the optional event seam.
type Config struct {
	Component Component
	Engines   []EngineDesc
	Handler   EventHandler // may be nil — observer-only engine.
}

// signalBufferSize keeps the registered signal channel from dropping into
// godbus's deferred-delivery path during signal bursts.
const signalBufferSize = 16

// Run connects to the private IBus bus, registers the component and serves
// engine traffic until ctx is cancelled. On bus loss the whole cycle
// repeats from address discovery — the socket path changes on every
// ibus-daemon generation, so the address is never cached (INTEG-04). A
// cancelled context is a clean shutdown: Run returns nil.
func Run(ctx context.Context, cfg Config) error {
	for generation := 0; ; generation++ {
		err := serve(ctx, &cfg, generation)
		if errors.Is(err, errNotPrimaryOwner) {
			return err
		}
		if err != nil {
			slog.Warn("ibus connection lost, reconnecting", "error", err)
		}

		if !backoff(ctx) {
			return nil
		}
	}
}

// serve performs one full connect→register→serve cycle. Errors it returns
// are transient (bus loss, ibus-daemon not up yet) except
// errNotPrimaryOwner, which Run treats as fatal.
func serve(ctx context.Context, cfg *Config, generation int) error {
	addr, err := Discover()
	if err != nil {
		return fmt.Errorf("discover ibus address: %w", err)
	}
	conn, err := dbus.Dial(addr)
	if err != nil {
		return fmt.Errorf("dial ibus bus: %w", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			slog.Warn("closing ibus connection", "error", closeErr)
		}
	}()

	if err = conn.Auth(userAuth()); err != nil {
		return fmt.Errorf("dbus auth: %w", err)
	}
	if err = conn.Hello(); err != nil {
		return fmt.Errorf("dbus hello: %w", err)
	}

	// Single-instance guard. org.freedesktop.IBus itself is reserved by
	// the daemon and refused to clients (verified live on the target bus),
	// so the component's own name carries the guard — the same pattern the
	// owner's Python prototype used. ReplaceExisting lets a reconnecting
	// generation retake its own name from a zombie connection.
	reply, err := conn.RequestName(cfg.Component.ComponentName, dbus.NameFlagReplaceExisting)
	if err != nil {
		return fmt.Errorf("request name %s: %w", cfg.Component.ComponentName, err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		slog.Error("another goswitchd instance already registered", "reply", int(reply))

		return fmt.Errorf("%w: got reply %d", errNotPrimaryOwner, int(reply))
	}
	slog.Info("connected", "address", addr)

	f := &factory{conn: conn, handler: cfg.Handler}
	if err = conn.Export(f, factoryPath, ifaceFactory); err != nil {
		return fmt.Errorf("export factory: %w", err)
	}

	component := cfg.Component
	component.EngineList = make([]dbus.Variant, 0, len(cfg.Engines))
	for i := range cfg.Engines {
		component.EngineList = append(component.EngineList, dbus.MakeVariant(cfg.Engines[i]))
	}
	ibus := conn.Object(ibusService, ibusPath)
	call := ibus.CallWithContext(ctx, ibusService+".RegisterComponent", 0, dbus.MakeVariant(component))
	if call.Err != nil {
		return fmt.Errorf("register component: %w", call.Err)
	}
	if generation == 0 {
		slog.Info("component registered", "engines", len(cfg.Engines))
	} else {
		slog.Info("re-registered", "generation", generation, "engines", len(cfg.Engines))
	}

	return waitBusLoss(ctx, conn)
}

// waitBusLoss blocks until ctx is cancelled (returns nil) or the signal
// channel closes — godbus closes channels registered through Signal() when
// the connection terminates, which is the bus-loss verdict. Signals that
// do arrive are drained: they carry no meaning for the tracer.
func waitBusLoss(ctx context.Context, conn *dbus.Conn) error {
	signals := make(chan *dbus.Signal, signalBufferSize)
	conn.Signal(signals)
	for {
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-signals:
			if !ok {
				return errBusClosed
			}
		}
	}
}

// backoff sleeps 1-2 s with jitter so a restarting ibus-daemon has time to
// come up, and reports whether the sleep completed (false means ctx was
// cancelled first).
func backoff(ctx context.Context) bool {
	delay := time.Second + rand.N(time.Second)
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// userAuth returns the D-Bus auth mechanisms to try: EXTERNAL with the
// decimal uid. godbus v5.2.2 ships no DBUS_COOKIE_SHA1 implementation for
// Unix (its default auth set is EXTERNAL alone), and the IBus socket
// authenticates via SO_PEERCRED, so EXTERNAL is the complete set.
func userAuth() []dbus.Auth {
	return []dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}
}
