package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"strconv"
	"sync"
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
// to register and the optional seams.
type Config struct {
	Component Component
	Engines   []EngineDesc
	Handler   EventHandler // may be nil — observer-only engine.

	// PostRegister is the optional self-reactivation seam (SY8): invoked
	// synchronously after EVERY successful RegisterComponent — generation 0
	// and every reconnect generation. nil = no-op. It must be cheap and
	// non-fatal: serve runs it inline before waitBusLoss, so a hang here
	// stalls the generation (the activate package's 10 s per-call timeout
	// bounds it).
	PostRegister func(ctx context.Context, generation int)

	// BindSwitcher is the optional flip seam (D-52): invoked once per
	// connection generation, right beside PostRegister, with the closure
	// that calls org.freedesktop.IBus.SetGlobalEngine on THIS generation's
	// connection. A switcher captured from a dead generation must never
	// survive the reconnect — the rebinding HERE is the guarantee
	// (T-05-03-03). nil = no-op (the PostRegister contract). The closure is
	// invoked by the actor on the flip path, never by the serve loop: a slow
	// bus can only stall a flip, never the registration cycle, and its
	// failures are journal lines (switch_engine), never generation errors.
	BindSwitcher func(flip func(ctx context.Context, engineName string) error)

	// OnGlobalEngine is the sync-listener input (05-04 criterion 3): invoked
	// with the engine name of every org.freedesktop.IBus.GlobalEngineChanged
	// signal on this connection AND of every FocusIn — the daemon follows
	// the FACTUAL active engine instead of trusting gsettings. nil = no-op.
	// It is called from the signal dispatcher's goroutine and from engine
	// D-Bus handlers: the callee must be cheap and never panic (the
	// FocusIn path is recoverHandler-wrapped). It NEVER flips anything
	// itself — the sync corrects only the receiver's internal state, never
	// the bus (single-writer, T-05-04-01).
	OnGlobalEngine func(engineName string)

	// BindGlobalEngine is the generation-scoped GetGlobalEngine reader seam
	// (05-04, Pitfall 3): invoked once per connection generation — bound
	// BEFORE PostRegister, so the same generation's reactivation can consult
	// it — with the closure that reads the FACTUAL active engine off THIS
	// generation's connection. A reader captured from a dead generation must
	// never survive the reconnect. nil = no-op (the PostRegister contract).
	BindGlobalEngine func(get func(ctx context.Context) (string, error))
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
	conn, addr, err := dialIbus()
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			slog.Warn("closing ibus connection", "error", closeErr)
		}
	}()

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

	f := &factory{conn: conn, handler: cfg.Handler, onGlobalEngine: cfg.OnGlobalEngine}
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
	logRegistered(generation, len(cfg.Engines))

	// The sync listener's plumbing lands BEFORE the post-registration seams:
	// anything ibus-daemon emits while they run (a re-registration can flip
	// the global engine) lands in the buffer and dispatches — the sync
	// listener must not lose the earliest facts.
	signals, err := subscribeGlobalEngine(ctx, cfg, conn, ibus)
	if err != nil {
		return err
	}
	if cfg.PostRegister != nil {
		cfg.PostRegister(ctx, generation)
	}
	if cfg.BindSwitcher != nil {
		cfg.BindSwitcher(newSwitcher(ibus))
	}

	return dispatchSignals(ctx, cfg, signals)
}

// logRegistered records the registration outcome: the first registration
// and every reconnect are distinct facts of the serve cycle.
func logRegistered(generation, engines int) {
	if generation == 0 {
		slog.Info("component registered", "engines", engines)

		return
	}
	slog.Info("re-registered", "generation", generation, "engines", engines)
}

// dialIbus performs one full connect handshake of a serve generation:
// address discovery (never cached — the socket path changes with every
// ibus-daemon generation), the dial with the storm-safe signal handler (it
// owns the sync listener's drop policy, T-05-04-02, and the bus-loss
// verdict via Terminate), EXTERNAL auth and Hello. The address rides along
// for the connected journal record.
func dialIbus() (*dbus.Conn, string, error) {
	addr, err := Discover()
	if err != nil {
		return nil, "", fmt.Errorf("discover ibus address: %w", err)
	}
	conn, err := dbus.Dial(addr, dbus.WithSignalHandler(newStormHandler()))
	if err != nil {
		return nil, "", fmt.Errorf("dial ibus bus: %w", err)
	}
	if err = conn.Auth(userAuth()); err != nil {
		return nil, "", fmt.Errorf("dbus auth: %w", err)
	}
	if err = conn.Hello(); err != nil {
		return nil, "", fmt.Errorf("dbus hello: %w", err)
	}

	return conn, addr, nil
}

// subscribeGlobalEngine wires the sync listener into one connection
// generation: the match rule that makes ibus-daemon forward
// GlobalEngineChanged to THIS connection (spike P2: AddMatch works on the
// ibus socket; the shell subscribes the same way), the buffered signal
// channel, and the generation-scoped GetGlobalEngine reader seam — bound
// here so the same generation's reactivation (IfOwned, Pitfall 3) consults
// the factual engine, never a stale closure from a dead generation. An
// AddMatch refusal fails the generation: a bus that refuses match rules
// cannot carry the sync listener.
func subscribeGlobalEngine(
	ctx context.Context, cfg *Config, conn *dbus.Conn, ibus dbus.BusObject,
) (<-chan *dbus.Signal, error) {
	if err := conn.AddMatchSignalContext(ctx,
		dbus.WithMatchObjectPath(ibusPath),
		dbus.WithMatchInterface(ibusService),
	); err != nil {
		return nil, fmt.Errorf("subscribe global engine changes: %w", err)
	}
	signals := make(chan *dbus.Signal, signalBufferSize)
	conn.Signal(signals)
	if cfg.BindGlobalEngine != nil {
		cfg.BindGlobalEngine(newGlobalEngineReader(ibus))
	}

	return signals, nil
}

// newSwitcher builds the generation-scoped SetGlobalEngine closure — the
// D-52 flip act (wire signature SetGlobalEngine(in s engine_name),
// live-verified; the RegisterComponent call shape above is the in-repo
// precedent). The journal record names the engine — a config literal
// (D-20/D-21) — at INFO on success and WARN with the cause on failure: a
// flip loss must stay observable, and it is NEVER a generation error. The
// caller's context carries the deadline (the actor bounds it; the engine
// layer only honors it).
func newSwitcher(ibus dbus.BusObject) func(ctx context.Context, engineName string) error {
	return func(ctx context.Context, engineName string) error {
		err := ibus.CallWithContext(ctx, ibusService+".SetGlobalEngine", 0, engineName).Err
		if err != nil {
			slog.Warn("switch_engine", "engine", engineName, "err", err)

			return fmt.Errorf("set global engine: %w", err)
		}
		slog.Info("switch_engine", "engine", engineName)

		return nil
	}
}

// dispatchSignals consumes the registered signal stream until ctx is
// cancelled (returns nil) or godbus closes the channel — the bus-loss
// verdict (errBusClosed), byte-identical to the pre-phase drain. A
// GlobalEngineChanged signal on the IBus path carries the FACTUAL active
// engine (criterion 3): its name funnels into Config.OnGlobalEngine, the
// single sync input the daemon follows the bus through. Every other signal
// is ignored — the path and the interface.member must name the IBus service
// exactly, so foreign or spoofed traffic cannot move the daemon's state
// (T-05-04-03). The dispatcher never touches the bus itself: the sync
// corrects only the receiver's internal state (T-05-04-01).
func dispatchSignals(ctx context.Context, cfg *Config, signals <-chan *dbus.Signal) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case sig, ok := <-signals:
			if !ok {
				return errBusClosed
			}
			if cfg.OnGlobalEngine == nil {
				continue
			}
			if name, ok := globalEngineChangedName(sig); ok {
				cfg.OnGlobalEngine(name)
			}
		}
	}
}

// globalEngineChangedName extracts the engine-name payload of a
// GlobalEngineChanged signal — only when BOTH the object path and the
// interface.member name the IBus service; anything else is not the sync
// listener's business (T-05-04-03).
func globalEngineChangedName(sig *dbus.Signal) (string, bool) {
	if sig == nil || sig.Path != ibusPath || sig.Name != ibusService+".GlobalEngineChanged" {
		return "", false
	}
	if len(sig.Body) == 0 {
		return "", false
	}
	name, ok := sig.Body[0].(string)

	return name, ok
}

// stormHandler is the signal handler of the daemon's ibus connection: a
// strict non-blocking deliver. godbus's default handler defers a BLOCKING
// goroutine per signal that does not fit a full channel (v5.2.2
// default_handler.go deliver → deferredDeliver) — a signal storm would
// park one goroutine and queue one *Signal per overflow: unbounded memory
// on the daemon's hot path. Here a full registered channel means the
// signal is DROPPED and counted, warned about exactly once (the
// switcherWarned precedent — a storm must not flood the journal): memory
// stays bounded by signalBufferSize (T-05-04-02). Terminate closes every
// registered channel — the bus-loss verdict (errBusClosed) rides on this,
// mirroring the default handler's Terminate.
type stormHandler struct {
	mu     sync.Mutex
	chans  []chan<- *dbus.Signal
	warned bool
}

// newStormHandler builds the per-generation signal handler.
func newStormHandler() *stormHandler { return &stormHandler{} }

// DeliverSignal implements dbus.SignalHandler: non-blocking fan-out, drop
// with one bounded WARN on a full channel.
func (h *stormHandler) DeliverSignal(_, _ string, signal *dbus.Signal) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ch := range h.chans {
		select {
		case ch <- signal:
		default:
			if !h.warned {
				h.warned = true
				slog.Warn("signal dropped", "capacity", signalBufferSize)
			}
		}
	}
}

// AddSignal implements the SignalRegistrar seam conn.Signal dispatches to.
func (h *stormHandler) AddSignal(ch chan<- *dbus.Signal) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.chans = append(h.chans, ch)
}

// RemoveSignal implements the SignalRegistrar seam conn.RemoveSignal
// dispatches to.
func (h *stormHandler) RemoveSignal(ch chan<- *dbus.Signal) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i := len(h.chans) - 1; i >= 0; i-- {
		if h.chans[i] == ch {
			h.chans = append(h.chans[:i], h.chans[i+1:]...)
		}
	}
}

// Terminate implements dbus.Terminator — conn.Close calls it exactly once.
// Closing every registered channel is the bus-loss verdict itself.
func (h *stormHandler) Terminate() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ch := range h.chans {
		close(ch)
	}
	h.chans = nil
}

// errGlobalEngineShape is the parse-failure sentinel of the GetGlobalEngine
// reply (err113): the factual engine could not be derived from the wire
// shape the daemon saw.
var errGlobalEngineShape = errors.New("unexpected global engine wire shape")

// newGlobalEngineReader builds the generation-scoped GetGlobalEngine reader
// — the FACTUAL active engine, the truth GNOME 46 never writes into the
// gsettings `current` key (Pitfall 3). The call copies the RegisterComponent
// shape (the in-repo precedent for calls TO ibus-daemon); the reply's
// variant-wrapped EngineDesc carries the engine name at positional field 2
// of the types.go wire order (Name, Attachments, EngineName) — decoded
// defensively like decodeIBusText, never a second wire parser. The caller's
// context carries the deadline (the activate package bounds it; the engine
// layer only honors it).
func newGlobalEngineReader(ibus dbus.BusObject) func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		call := ibus.CallWithContext(ctx, ibusService+".GetGlobalEngine", 0)
		if call.Err != nil {
			return "", fmt.Errorf("get global engine: %w", call.Err)
		}
		desc, ok := call.Body[0].(dbus.Variant)
		if !ok {
			return "", fmt.Errorf("global engine reply: %w", errGlobalEngineShape)
		}
		name, ok := engineDescName(desc)
		if !ok {
			return "", fmt.Errorf("global engine desc: %w", errGlobalEngineShape)
		}

		return name, nil
	}
}

// engineDescName extracts the engine name from a variant-wrapped serialized
// EngineDesc — positional field 2 of the types.go wire order. godbus
// decodes STRUCT generically into []any (the typed struct never appears on
// the incoming path), so the fields are extracted defensively and any other
// shape is refused.
func engineDescName(desc dbus.Variant) (string, bool) {
	fields, ok := desc.Value().([]any)
	if !ok || len(fields) < 3 {
		return "", false
	}
	name, ok := fields[2].(string)

	return name, ok
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
