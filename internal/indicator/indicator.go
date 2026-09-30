package indicator

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/godbus/dbus/v5"
)

// The SNI wire constants: the watcher's well-known name and object path,
// the item's object path and interface, the Properties surface the
// extension reads the item through, and the one signal of the v0 surface.
const (
	watcherName = "org.kde.StatusNotifierWatcher"
	watcherPath = dbus.ObjectPath("/StatusNotifierWatcher")
	itemPath    = dbus.ObjectPath("/StatusNotifierItem")
	sniIface    = "org.kde.StatusNotifierItem"
	// propertiesIface is the org.freedesktop.DBus.Properties interface the
	// item exports itself under — godbus routes method calls strictly by
	// (path, interface), so serving Get/GetAll means exporting the item a
	// second time under this name (the prop-package route, hand-rolled to
	// keep the corpus hermetic).
	propertiesIface = "org.freedesktop.DBus.Properties"
	signalNewIcon   = "NewIcon"
)

// The served SNI v0 property surface: icon-only (no Menu, no ToolTip —
// the extension renders the icon alone, owner-verified on this desktop).
const (
	propCategory   = "Category"
	propID         = "Id"
	propTitle      = "Title"
	propStatus     = "Status"
	propIconName   = "IconName"
	propIconPixmap = "IconPixmap"
	propWindowID   = "WindowId"
)

// The property VALUES the item serves — static identity, the SNI spec's
// ApplicationStatus idiom.
const (
	sniCategory     = "ApplicationStatus"
	sniStatusActive = "Active"
	serviceID       = "goswitch"
)

// The D-Bus standard error names the Properties surface answers refusals
// with.
const (
	errNameUnknownProperty = "org.freedesktop.DBus.Error.UnknownProperty"
	errNameUnknownIface    = "org.freedesktop.DBus.Error.UnknownInterface"
	errNameReadOnly        = "org.freedesktop.DBus.Error.PropertyReadOnly"
)

// errNoWatcher is the absent-watcher verdict of the probe — the reason the
// attach degradation names (the daemon's one WARN).
var errNoWatcher = errors.New("no StatusNotifierWatcher on the session bus")

// Watcher probes the org.kde.StatusNotifierWatcher owner and registers the
// item with it — the daemon's point-of-use view of the watcher surface
// (the fakeSink precedent: the corpus runs without a live bus).
type Watcher interface {
	// NameHasOwner reports the presence of the watcher's well-known name.
	NameHasOwner() (bool, error)
	// RegisterStatusNotifierItem registers the item's service with the
	// watcher (the SNI registration act).
	RegisterStatusNotifierItem(service string) error
}

// Emitter sends one signal on the daemon's session-bus connection — the
// NewIcon channel of the v0 surface.
type Emitter interface {
	// Emit sends one signal of the named interface at the object path.
	Emit(objectPath dbus.ObjectPath, iface, signal string, args ...any) error
}

// exporter publishes one object under one interface on the connection —
// the *dbus.Conn Export surface narrowed to what the item needs.
type exporter interface {
	Export(v any, path dbus.ObjectPath, iface string) error
}

// connWatcher is the Watcher over the daemon's live session-bus
// connection: the probe rides the bus object, the register call the
// watcher's.
type connWatcher struct {
	conn *dbus.Conn
}

// NameHasOwner probes the well-known name via the bus object.
func (w connWatcher) NameHasOwner() (bool, error) {
	var has bool
	if err := w.conn.BusObject().Call("org.freedesktop.DBus.NameHasOwner", 0, watcherName).Store(&has); err != nil {
		return false, fmt.Errorf("probe %s: %w", watcherName, err)
	}

	return has, nil
}

// RegisterStatusNotifierItem registers the service with the watcher.
func (w connWatcher) RegisterStatusNotifierItem(service string) error {
	call := w.conn.Object(watcherName, watcherPath).
		Call("org.kde.StatusNotifierWatcher.RegisterStatusNotifierItem", 0, service)
	if call.Err != nil {
		return fmt.Errorf("register %s with %s: %w", service, watcherName, call.Err)
	}

	return nil
}

// connEmitter is the Emitter over the daemon's live connection.
type connEmitter struct {
	conn *dbus.Conn
}

// Emit sends the signal on the connection.
func (e connEmitter) Emit(objectPath dbus.ObjectPath, iface, signal string, args ...any) error {
	if err := e.conn.Emit(objectPath, iface+"."+signal, args...); err != nil {
		return fmt.Errorf("emit %s.%s: %w", iface, signal, err)
	}

	return nil
}

// Item is the exported tray object — the org.kde.StatusNotifierItem at
// /StatusNotifierItem under the daemon's well-known name, and the actor's
// ModeDisplay seam. godbus dispatches property reads on their own
// goroutines, so the pixmap swap and every read go through the item's own
// mutex (T-Q6-04: no path leads back into the actor — no lock cycle).
type Item struct {
	mu           sync.Mutex
	pix          Pixmap  // the pixmap the Properties surface serves
	emitter      Emitter // the NewIcon channel (always non-nil at construction)
	path         dbus.ObjectPath
	disabled     bool // permanent inert state: a degraded attach or a failed emit
	attachWarned bool // one WARN per degradation type per item lifetime
	emitWarned   bool
	symbolWarned bool
}

// Attach builds the tray item on conn — the daemon's EXISTING session-bus
// connection (never a second Connect, never a second RequestName: the item
// is reachable at <service>/StatusNotifierItem, and service is the
// well-known name the control service already owns). Attach ALWAYS
// succeeds: every degradation is a one-WARN inert display inside the item.
func Attach(conn *dbus.Conn, service string) *Item {
	return attach(connWatcher{conn: conn}, connEmitter{conn: conn}, conn, service)
}

// attach is the testable core of Attach: probe → export → register, one
// shot, off the key path. The export precedes the register so the
// watcher's immediate property reads after the registration find the
// object served.
func attach(w Watcher, em Emitter, exp exporter, service string) *Item {
	it := &Item{
		emitter: em,
		path:    itemPath,
	}
	if pm, ok := PixmapFor(symbolEN); ok {
		it.pix = pm // EN at start (ADR-001); the install-push corrects any skew
	}
	if err := registerItem(w, exp, it, service); err != nil {
		it.disabled = true
		it.attachWarned = true
		slog.Warn("tray indicator attach failed", "component", "tray indicator", "error", err)
	}

	return it
}

// registerItem runs the one-shot attach-time registration sequence: the
// watcher probe first (absent — the indicator is a permanent no-op, v0
// never re-checks), then the double export, then the register. The EXPORTED
// value is the live item itself — a throwaway instance would serve an empty
// pixmap to every property read.
func registerItem(w Watcher, exp exporter, it *Item, service string) error {
	has, err := w.NameHasOwner()
	if err != nil {
		return fmt.Errorf("probe %s: %w", watcherName, err)
	}
	if !has {
		return errNoWatcher
	}
	if err := exp.Export(it, itemPath, sniIface); err != nil {
		return fmt.Errorf("export %s: %w", sniIface, err)
	}
	if err := exp.Export(it, itemPath, propertiesIface); err != nil {
		return fmt.Errorf("export %s: %w", propertiesIface, err)
	}

	if err := w.RegisterStatusNotifierItem(service); err != nil {
		return fmt.Errorf("register %s: %w", service, err)
	}

	return nil
}

// ModeChanged implements the actor's ModeDisplay seam (the observer is
// invoked last in every mode record, under the actor's mutex). A known
// symbol swaps the served pixmap and emits NewIcon — a queued message
// send, no round trip, so the call stays quick (T-Q6-03). An unknown
// symbol is one WARN no-op; an emit error is one WARN and the display
// self-disables (the dead-connection case after ctl shutdown).
func (it *Item) ModeChanged(symbol string) {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.disabled {
		return
	}
	pm, ok := PixmapFor(symbol)
	if !ok {
		if !it.symbolWarned {
			it.symbolWarned = true
			slog.Warn("tray indicator unknown mode symbol", "component", "tray indicator", "symbol", symbol)
		}

		return
	}
	it.pix = pm
	if err := it.emitter.Emit(it.path, sniIface, signalNewIcon); err != nil {
		it.disabled = true
		if !it.emitWarned {
			it.emitWarned = true
			slog.Warn("tray indicator emit failed", "component", "tray indicator", "error", err)
		}
	}
}

// Get serves org.freedesktop.DBus.Properties.Get for the item — the
// dispatch re-entry the SNI extension drives on its own goroutine. Error
// returns carry the zero Variant: MakeVariant(nil) would panic in the
// signature computation, and the value is unspecified whenever the error
// is non-nil anyway.
func (it *Item) Get(iface, property string) (dbus.Variant, *dbus.Error) {
	if iface != sniIface {
		return dbus.Variant{}, dbus.NewError(errNameUnknownIface, []any{iface})
	}

	props := it.properties()
	v, ok := props[property]
	if !ok {
		return dbus.Variant{}, dbus.NewError(errNameUnknownProperty, []any{property})
	}

	return v, nil
}

// GetAll serves org.freedesktop.DBus.Properties.GetAll for the item.
func (it *Item) GetAll(iface string) (map[string]dbus.Variant, *dbus.Error) {
	if iface != sniIface {
		return nil, dbus.NewError(errNameUnknownIface, []any{iface})
	}

	return it.properties(), nil
}

// Set refuses: every SNI property of the item is read-only — the mode
// changes ride ModeChanged, never a property write.
func (it *Item) Set(iface, property string, _ dbus.Variant) *dbus.Error {
	return dbus.NewError(errNameReadOnly, []any{property + " on " + iface + " is read-only"})
}

// properties snapshots the served property set under the mutex — the one
// map both Get and GetAll read from.
func (it *Item) properties() map[string]dbus.Variant {
	it.mu.Lock()
	defer it.mu.Unlock()

	return map[string]dbus.Variant{
		propCategory:   dbus.MakeVariant(sniCategory),
		propID:         dbus.MakeVariant(serviceID),
		propTitle:      dbus.MakeVariant(serviceID),
		propStatus:     dbus.MakeVariant(sniStatusActive),
		propIconName:   dbus.MakeVariant(""),
		propIconPixmap: dbus.MakeVariant([]Pixmap{it.pix}),
		propWindowID:   dbus.MakeVariant(int32(0)),
	}
}
