// Package engine implements the IBus input-method engine wire adapter: it
// speaks the private IBus D-Bus protocol (registration, factory, per-input
// context engine object) and funnels decoded key and lifecycle events into
// the EventHandler seam.
package engine

import (
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/godbus/dbus/v5"
)

// Wire interfaces the engine object is exported on — all three at once, as
// ibus-daemon dispatches calls on whichever interface it discovered the
// method on.
const (
	ifaceEngine  = "org.freedesktop.IBus.Engine"
	ifaceService = "org.freedesktop.IBus.Service"
	ifaceProps   = "org.freedesktop.DBus.Properties"
)

// LifecycleKind enumerates the engine lifecycle events forwarded to the
// EventHandler seam.
type LifecycleKind int

// Lifecycle kinds delivered to HandleLifecycle.
const (
	LifecycleFocusIn LifecycleKind = iota
	LifecycleFocusOut
	LifecycleReset
	LifecycleEnable
	LifecycleDisable
)

// String returns the stable label used in log records.
func (k LifecycleKind) String() string {
	switch k {
	case LifecycleFocusIn:
		return "focus_in"
	case LifecycleFocusOut:
		return "focus_out"
	case LifecycleReset:
		return "reset"
	case LifecycleEnable:
		return "enable"
	case LifecycleDisable:
		return "disable"
	default:
		return fmt.Sprintf("lifecycle(%d)", int(k))
	}
}

// EngineEvent is the decoded form of a ProcessKeyEvent call. The adapter
// decodes the raw IBus state word exactly once; consumers never re-parse
// raw state.
type EngineEvent struct {
	Keyval  uint32
	Keycode uint32
	Release bool
	Mods    uint32
}

// Emitter is the outgoing-signal surface of an Engine — the ladder
// primitives the correction pipeline drives (ADR-003). Declared here
// because EventHandler.AttachEngine hands the freshly minted engine to the
// handler as this seam, not as the concrete object: consumers (and their
// test doubles) program against the four emitters alone.
type Emitter interface {
	RequireSurroundingText()
	DeleteSurroundingText(offset int32, nchars uint32)
	ForwardKeyEvent(keyval, keycode, state uint32)
	CommitText(text IBusText)
}

// EventHandler is the seam the rest of the daemon plugs into (hotkey FSM,
// buffers, correction). A nil handler is legal: the engine then only
// observes and logs. HandleKey's bool is the consumption verdict: true means
// the handler already delivered the key's character (RU-mode commit) and the
// client must not insert anything.
type EventHandler interface {
	HandleKey(ev EngineEvent) (consume bool)
	HandleLifecycle(kind LifecycleKind)
	// HandleSurroundingText delivers the surrounding text the client
	// reported (the runes before the anchor at cursorPos) — the input of
	// the ADR-004 pre-correction verification.
	HandleSurroundingText(text string, cursorPos uint32)
	// HandleCapabilities delivers the per-input-context capability bitmap
	// (SetCapabilities), the input of the ADR-003 ladder-level choice.
	HandleCapabilities(caps uint32)
	// AttachEngine binds the emitter sink of the engine minted for the
	// input context: the handler's corrections drive it.
	AttachEngine(eng Emitter)
}

// Engine is the per-input-context D-Bus object exported on the three
// interfaces above. godbus dispatches every method call on its own
// goroutine, so every exported handler carries a recover shim (INTEG-05):
// one unrecovered panic would kill the process and, through ibus-daemon's
// lifecycle coupling, desktop-wide input.
type Engine struct {
	conn    *dbus.Conn
	path    dbus.ObjectPath
	name    string
	handler EventHandler
	caps    uint32
}

// NewEngine creates an engine object for the named engine (e.g.
// "goswitch-en"); the name rides along into lifecycle log records so an
// observer can tell WHICH of the registered engines took focus — the
// observable the D-01 experiment asserts on. The handler may be nil. The
// connection and object path are bound at export time; a detached engine
// (no connection) simply skips signal emission, which keeps the object
// constructible in headless unit tests.
func NewEngine(handler EventHandler, name string) *Engine {
	return &Engine{handler: handler, name: name}
}

// recoverHandler contains a panic raised anywhere below a D-Bus handler:
// the panic is swallowed and logged at ERROR, and the caller returns its
// zero values (a safe reply — never a D-Bus error, which ibus-daemon would
// treat as engine failure).
func recoverHandler(method string, dbusErr **dbus.Error) {
	if r := recover(); r != nil {
		slog.Error("handler panic contained",
			"method", method,
			"panic", fmt.Sprint(r),
			"stack", string(debug.Stack()))
		*dbusErr = nil
	}
}

// decodeEvent converts the raw ProcessKeyEvent arguments into an
// EngineEvent: bit 30 is the release flag, the low byte carries the
// modifier mask.
func decodeEvent(keyval, keycode, state uint32) EngineEvent {
	return EngineEvent{
		Keyval:  keyval,
		Keycode: keycode,
		Release: state&MaskRelease != 0,
		Mods:    state & modsMask,
	}
}

// ProcessKeyEvent implements org.freedesktop.IBus.Engine.ProcessKeyEvent
// (u keyval, u keycode, u state) → b. It decodes the event, traces it at
// DEBUG and forwards it to the handler: the handler's verdict IS the answer
// — consumption is decided by the EventHandler (the RU script mode of Phase
// 2 commits a Cyrillic rune and consumes the key), the transport only
// returns it. A nil handler stays a pure observer (transit, the Phase 1
// contract), and bare modifier presses are never consumed — the actor
// declines every non-printable.
func (e *Engine) ProcessKeyEvent(keyval, keycode, state uint32) (handled bool, err *dbus.Error) {
	defer recoverHandler("ProcessKeyEvent", &err)

	ev := decodeEvent(keyval, keycode, state)
	slog.Debug("key",
		"keyval", fmt.Sprintf("0x%x", ev.Keyval),
		"keycode", ev.Keycode,
		"release", ev.Release,
		"mods", fmt.Sprintf("0x%x", ev.Mods))
	var consume bool
	if e.handler != nil {
		consume = e.handler.HandleKey(ev)
	}

	return consume, nil
}

// SetCursorLocation implements org.freedesktop.IBus.Engine.SetCursorLocation
// (i,i,i,i): a no-op in Phase 1.
func (e *Engine) SetCursorLocation(x, y, width, height int32) (err *dbus.Error) {
	defer recoverHandler("SetCursorLocation", &err)

	return nil
}

// SetSurroundingText implements
// org.freedesktop.IBus.Engine.SetSurroundingText (v text, u cursor_pos,
// u anchor_pos): the variant-wrapped IBusText is decoded once here and the
// text with its cursor position is funneled to the handler — the input of
// the ADR-004 pre-correction verification. Logged at DEBUG, contents never
// recorded at INFO (D-20). A payload that does not decode as IBusText is
// dropped: the correction waiting on it then times out silently.
func (e *Engine) SetSurroundingText(text dbus.Variant, cursorPos, anchorPos uint32) (err *dbus.Error) {
	defer recoverHandler("SetSurroundingText", &err)
	slog.Debug("surrounding_text", "cursor_pos", cursorPos, "anchor_pos", anchorPos)
	if e.handler != nil {
		if t, ok := decodeIBusText(text); ok {
			e.handler.HandleSurroundingText(t.Text, cursorPos)
		}
	}

	return nil
}

// decodeIBusText decodes a variant-wrapped IBusText wire struct. godbus
// decodes STRUCT generically into []any — the typed struct never appears on
// the incoming path — so the positional fields (types.go field order is the
// wire contract) are extracted defensively and any other shape is refused.
func decodeIBusText(text dbus.Variant) (IBusText, bool) {
	fields, ok := text.Value().([]any)
	if !ok || len(fields) < 4 {
		return IBusText{}, false
	}
	name, ok := fields[0].(string)
	if !ok {
		return IBusText{}, false
	}
	attachments, ok := fields[1].(map[string]dbus.Variant)
	if !ok {
		return IBusText{}, false
	}
	value, ok := fields[2].(string)
	if !ok {
		return IBusText{}, false
	}
	attrList, ok := fields[3].(dbus.Variant)
	if !ok {
		return IBusText{}, false
	}

	return IBusText{Name: name, Attachments: attachments, Text: value, AttrList: attrList}, true
}

// SetCapabilities implements org.freedesktop.IBus.Engine.SetCapabilities
// (u caps): the capability bitmap is stored per input context and funneled
// to the handler — the input of the ADR-003 ladder-level choice.
func (e *Engine) SetCapabilities(caps uint32) (err *dbus.Error) {
	defer recoverHandler("SetCapabilities", &err)
	e.caps = caps
	slog.Debug("capabilities", "caps", fmt.Sprintf("0x%x", caps))
	if e.handler != nil {
		e.handler.HandleCapabilities(caps)
	}

	return nil
}

// FocusIn implements org.freedesktop.IBus.Engine.FocusIn ().
func (e *Engine) FocusIn() (err *dbus.Error) {
	defer recoverHandler("FocusIn", &err)
	slog.Info("focus_in", "engine", e.name)
	e.lifecycle(LifecycleFocusIn)

	return nil
}

// FocusOut implements org.freedesktop.IBus.Engine.FocusOut ().
func (e *Engine) FocusOut() (err *dbus.Error) {
	defer recoverHandler("FocusOut", &err)
	slog.Info("focus_out", "engine", e.name)
	e.lifecycle(LifecycleFocusOut)

	return nil
}

// Reset implements org.freedesktop.IBus.Engine.Reset ().
func (e *Engine) Reset() (err *dbus.Error) {
	defer recoverHandler("Reset", &err)
	e.lifecycle(LifecycleReset)

	return nil
}

// Enable implements org.freedesktop.IBus.Engine.Enable ().
func (e *Engine) Enable() (err *dbus.Error) {
	defer recoverHandler("Enable", &err)
	e.lifecycle(LifecycleEnable)

	return nil
}

// Disable implements org.freedesktop.IBus.Engine.Disable ().
func (e *Engine) Disable() (err *dbus.Error) {
	defer recoverHandler("Disable", &err)
	e.lifecycle(LifecycleDisable)

	return nil
}

// PageUp implements org.freedesktop.IBus.Engine.PageUp ().
func (e *Engine) PageUp() (err *dbus.Error) {
	defer recoverHandler("PageUp", &err)

	return nil
}

// PageDown implements org.freedesktop.IBus.Engine.PageDown ().
func (e *Engine) PageDown() (err *dbus.Error) {
	defer recoverHandler("PageDown", &err)

	return nil
}

// CursorUp implements org.freedesktop.IBus.Engine.CursorUp ().
func (e *Engine) CursorUp() (err *dbus.Error) {
	defer recoverHandler("CursorUp", &err)

	return nil
}

// CursorDown implements org.freedesktop.IBus.Engine.CursorDown ().
func (e *Engine) CursorDown() (err *dbus.Error) {
	defer recoverHandler("CursorDown", &err)

	return nil
}

// CandidateClicked implements org.freedesktop.IBus.Engine.CandidateClicked
// (u,u,u).
func (e *Engine) CandidateClicked(index, button, state uint32) (err *dbus.Error) {
	defer recoverHandler("CandidateClicked", &err)

	return nil
}

// PropertyActivate implements org.freedesktop.IBus.Engine.PropertyActivate
// (s,u).
func (e *Engine) PropertyActivate(name string, state uint32) (err *dbus.Error) {
	defer recoverHandler("PropertyActivate", &err)

	return nil
}

// PropertyShow implements org.freedesktop.IBus.Engine.PropertyShow (s).
func (e *Engine) PropertyShow(name string) (err *dbus.Error) {
	defer recoverHandler("PropertyShow", &err)

	return nil
}

// PropertyHide implements org.freedesktop.IBus.Engine.PropertyHide (s).
func (e *Engine) PropertyHide(name string) (err *dbus.Error) {
	defer recoverHandler("PropertyHide", &err)

	return nil
}

// Destroy implements org.freedesktop.IBus.Service.Destroy (): the object
// is unexported from all three interfaces.
func (e *Engine) Destroy() (err *dbus.Error) {
	defer recoverHandler("Destroy", &err)
	if e.conn != nil {
		for _, iface := range []string{ifaceEngine, ifaceService, ifaceProps} {
			_ = e.conn.Export(nil, e.path, iface)
		}
	}

	return nil
}

// Get implements org.freedesktop.DBus.Properties.Get (s) → v: engines
// expose no properties in Phase 1.
func (e *Engine) Get(name string) (value dbus.Variant, err *dbus.Error) {
	defer recoverHandler("Get", &err)

	return dbus.MakeVariant(""), nil
}

// GetAll implements org.freedesktop.DBus.Properties.GetAll (s) → a{sv}.
func (e *Engine) GetAll(interfaceName string) (props map[string]dbus.Variant, err *dbus.Error) {
	defer recoverHandler("GetAll", &err)

	return map[string]dbus.Variant{}, nil
}

// Set implements org.freedesktop.DBus.Properties.Set (s,v).
func (e *Engine) Set(name string, value dbus.Variant) (err *dbus.Error) {
	defer recoverHandler("Set", &err)

	return nil
}

// CommitText emits the org.freedesktop.IBus.Engine.CommitText signal — the
// text-injection primitive. The adapter ships the emitter from day one;
// correction logic wires it in Phase 2.
func (e *Engine) CommitText(text IBusText) {
	if e.conn == nil {
		return
	}
	if err := e.conn.Emit(e.path, ifaceEngine+".CommitText", dbus.MakeVariant(text)); err != nil {
		slog.Error("commit text emit failed", "error", err)
	}
}

// DeleteSurroundingText emits the org.freedesktop.IBus.Engine.
// DeleteSurroundingText signal — the level-1 ladder primitive (ADR-003):
// the client is asked to delete nchars runes before (negative offset) or
// after the cursor. Fire-and-forget on IBus 1.5.29 — the ADR-004
// verification is the compensation, not an ack.
func (e *Engine) DeleteSurroundingText(offset int32, nchars uint32) {
	if e.conn == nil {
		return
	}
	if err := e.conn.Emit(e.path, ifaceEngine+".DeleteSurroundingText", offset, nchars); err != nil {
		slog.Error("delete surrounding emit failed", "error", err)
	}
}

// RequireSurroundingText emits the org.freedesktop.IBus.Engine.
// RequireSurroundingText signal — the client answers with a fresh
// SetSurroundingText, which is the input of the pre-correction
// verification (ADR-004).
func (e *Engine) RequireSurroundingText() {
	if e.conn == nil {
		return
	}
	if err := e.conn.Emit(e.path, ifaceEngine+".RequireSurroundingText"); err != nil {
		slog.Error("require surrounding emit failed", "error", err)
	}
}

// ForwardKeyEvent emits the org.freedesktop.IBus.Engine.ForwardKeyEvent
// signal — the level-2 ladder primitive (ADR-003): the key event is replayed
// to the client as if the user pressed it (the Backspace burst before the
// replacement commit). Fire-and-forget like every engine signal — the
// ADR-004 verification is the compensation, never an expected ack.
func (e *Engine) ForwardKeyEvent(keyval, keycode, state uint32) {
	if e.conn == nil {
		return
	}
	if err := e.conn.Emit(e.path, ifaceEngine+".ForwardKeyEvent", keyval, keycode, state); err != nil {
		slog.Error("forward key event emit failed", "error", err)
	}
}

// lifecycle forwards a lifecycle event to the handler, if installed.
func (e *Engine) lifecycle(kind LifecycleKind) {
	if e.handler != nil {
		e.handler.HandleLifecycle(kind)
	}
}
