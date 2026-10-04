package indicator

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"

	"github.com/godbus/dbus/v5"
)

// The SNI wire constants: the watcher's well-known name and object path,
// the item's object path and interface, the Properties surface the
// extension reads the item through, and the one signal of the v0 surface.
const (
	watcherName = "org.kde.StatusNotifierWatcher"
	watcherPath = dbus.ObjectPath("/StatusNotifierWatcher")
	// watcherIface is the watcher's D-Bus interface — the prefix the
	// RegisteredStatusNotifierItems property read goes through. The bus
	// name and the interface share the spelling; the item's sniIface is a
	// different string.
	watcherIface        = "org.kde.StatusNotifierWatcher"
	propRegisteredItems = "RegisteredStatusNotifierItems"
	itemPath            = dbus.ObjectPath("/StatusNotifierItem")
	sniIface            = "org.kde.StatusNotifierItem"
	// propertiesIface is the org.freedesktop.DBus.Properties interface the
	// item exports itself under — godbus routes method calls strictly by
	// (path, interface), so serving Get/GetAll means exporting the item a
	// second time under this name (the prop-package route, hand-rolled to
	// keep the corpus hermetic).
	propertiesIface = "org.freedesktop.DBus.Properties"
	signalNewIcon   = "NewIcon"
	// menuPath and menuIface locate the DBusMenu object — exported at /Menu
	// under com.canonical.dbusmenu once the menu exports succeed; the value
	// the Menu property serves before that is the sentinel below.
	menuPath  = dbus.ObjectPath("/Menu")
	menuIface = "com.canonical.dbusmenu"
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
	propMenu       = "Menu"
)

// The property VALUES the item serves — static identity, the SNI spec's
// ApplicationStatus idiom.
const (
	sniCategory     = "ApplicationStatus"
	sniStatusActive = "Active"
	serviceID       = "goswitch"

	// menuNoDBusMenu is the SNI icon-only sentinel: the ubuntu-appindicators
	// watcher lists Menu among NEEDED_PROPERTIES and DESTROYS an item that
	// refuses it, while this exact object path disables the DBusMenu setup
	// (appIndicator.js menuPath) — served, never dereferenced.
	menuNoDBusMenu = "/NO_DBUSMENU"
)

// The D-Bus standard error names the Properties surface answers refusals
// with.
const (
	errNameUnknownProperty = "org.freedesktop.DBus.Error.UnknownProperty"
	errNameUnknownIface    = "org.freedesktop.DBus.Error.UnknownInterface"
	errNameReadOnly        = "org.freedesktop.DBus.Error.PropertyReadOnly"
	errNameFailed          = "org.freedesktop.DBus.Error.Failed"
)

// errNoWatcher is the absent-watcher verdict of the probe — the reason the
// attach degradation names (the daemon's one WARN).
var errNoWatcher = errors.New("no StatusNotifierWatcher on the session bus")

// errItemsType is the registry-read type-mismatch verdict: the
// RegisteredStatusNotifierItems property arrived as something other than
// the as-typed string list (never seen on the wire — the wrap names the
// property for the log).
var errItemsType = errors.New("unexpected registered-items variant payload")

// Callbacks carries the daemon's interactive surface — small func fields the
// daemon wiring fills (quick plan 261001-fg3). Reload nil means the daemon
// runs without -config (the menu serves the reload item disabled); a nil
// Toggle or Status makes the corresponding gesture a contained no-op. The
// menu v2 fields (plan 07-05) ride the same nil-safe discipline: Switch is
// the EN/RU radio pair's target flip (the actor's SwitchMode), the two
// Toggle* fields the persisted config toggles, Settings the config-editor
// launcher, About the version notification. The callbacks run on godbus
// dispatch goroutines and must carry their own synchronization (the actor's
// mutex).
type Callbacks struct {
	Toggle func()
	Status func()
	Reload func()
	// The menu v2 gestures (plan 07-05) — each a silent no-op without its
	// callback.
	Switch            func(target string)
	ToggleAutocorrect func()
	ToggleSound       func()
	Settings          func()
	About             func()
}

// Watcher probes the org.kde.StatusNotifierWatcher owner, registers the
// item with it and reads the watcher's live item registry — the daemon's
// point-of-use view of the watcher surface (the fakeSink precedent: the
// corpus runs without a live bus).
type Watcher interface {
	// NameHasOwner reports the presence of the watcher's well-known name.
	NameHasOwner() (bool, error)
	// RegisterStatusNotifierItem registers the item's service with the
	// watcher (the SNI registration act).
	RegisterStatusNotifierItem(service string) error
	// RegisteredStatusNotifierItems reads the watcher's item list — the
	// supervisor's silent-eviction check (our service missing from the
	// list is the eviction verdict).
	RegisteredStatusNotifierItems() ([]string, error)
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

// RegisteredStatusNotifierItems reads the watcher's item registry — the
// as-typed as-list property on the watcher object.
func (w connWatcher) RegisteredStatusNotifierItems() ([]string, error) {
	v, err := w.conn.Object(watcherName, watcherPath).
		GetProperty(watcherIface + "." + propRegisteredItems)
	if err != nil {
		return nil, fmt.Errorf("read %s.%s: %w", watcherIface, propRegisteredItems, err)
	}
	items, ok := v.Value().([]string)
	if !ok {
		return nil, fmt.Errorf("read %s.%s: %w", watcherIface, propRegisteredItems, errItemsType)
	}

	return items, nil
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
	mu             sync.Mutex
	pix            Pixmap  // the pixmap the Properties surface serves
	emitter        Emitter // the NewIcon channel (always non-nil at construction)
	path           dbus.ObjectPath
	registered     bool // the watcher holds the item — supervisor-managed
	emitDead       bool // a failed emit — permanent per connection
	attachWarned   bool // one WARN per degradation type per item lifetime
	emitWarned     bool
	symbolWarned   bool
	menuWarned     bool
	activateWarned bool
	cb             Callbacks   // the interactive surface; immutable after attach
	menuPath       string      // the served Menu value: /Menu or the sentinel
	sup            *supervisor // the lifecycle owner; born with the item
}

// Attach builds the tray item on conn — the daemon's EXISTING session-bus
// connection (never a second Connect, never a second RequestName: the item
// is reachable at <service>/StatusNotifierItem, and service is the
// well-known name the control service already owns). Attach ALWAYS
// succeeds: every degradation is a one-WARN inert display inside the item.
func Attach(conn *dbus.Conn, service string, cb Callbacks) *Item {
	return attach(connWatcher{conn: conn}, connEmitter{conn: conn}, conn, service, cb)
}

// attach is the testable core of Attach: probe → export → register, one
// shot, off the key path. The export precedes the register so the
// watcher's immediate property reads after the registration find the
// object served. A failed registration is no longer the last word — the
// supervisor born with the item re-checks (quick plan 261001-fg3).
func attach(w Watcher, em Emitter, exp exporter, service string, cb Callbacks) *Item {
	it := &Item{
		emitter:  em,
		path:     itemPath,
		cb:       cb,
		menuPath: menuNoDBusMenu, // icon-only until the menu exports succeed
	}
	if pm, ok := PixmapFor(symbolEN); ok {
		it.pix = pm // EN at start (ADR-001); the install-push corrects any skew
	}
	it.sup = &supervisor{item: it, w: w, exp: exp, service: service}
	if err := registerItem(w, exp, it, service); err != nil {
		it.attachWarned = true
		slog.Warn("tray indicator attach failed", "component", "tray indicator", "error", err)

		return it
	}
	it.registered = true

	return it
}

// registerItem runs the attach-time registration sequence: the watcher
// probe first, then the menu exports (their outcome decides the Menu
// property BEFORE the item is registered — the watcher's immediate property
// reads must find the decision), then the item's double export, then the
// register. The EXPORTED value is the live item itself — a throwaway
// instance would serve an empty pixmap to every property read. A menu
// export failure degrades to the sentinel and CONTINUES; everything else
// fails the sequence (the supervisor re-runs it).
func registerItem(w Watcher, exp exporter, it *Item, service string) error {
	has, err := w.NameHasOwner()
	if err != nil {
		return fmt.Errorf("probe %s: %w", watcherName, err)
	}
	if !has {
		return errNoWatcher
	}

	menu := &Menu{cb: it.cb}
	if err := exp.Export(menu, menuPath, menuIface); err != nil {
		it.menuFailed(fmt.Errorf("export %s: %w", menuIface, err))
	} else if err := exp.Export(menu, menuPath, propertiesIface); err != nil {
		it.menuFailed(fmt.Errorf("export %s: %w", propertiesIface, err))
	} else {
		it.mu.Lock()
		it.menuPath = string(menuPath)
		it.mu.Unlock()
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
// symbol ALWAYS swaps the served pixmap — the icon stays fresh through a
// lost registration period — and, while the item is registered, emits
// NewIcon: a queued message send, no round trip, so the call stays quick
// (T-Q6-03). An unregistered item emits nothing (no shell listener), an
// unknown symbol is one WARN no-op, and an emit error is one WARN with the
// display self-disabling (the dead-connection case after ctl shutdown).
func (it *Item) ModeChanged(symbol string) {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.emitDead {
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
	if it.registered {
		it.emitNewIcon()
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

// Activate implements the SNI Activate method — the menu-less environment's
// path into the toggle (quick plan 261001-fg3): the same Callbacks.Toggle
// the menu's first item drives. A nil toggle is a one-WARN no-op (the
// emitWarned discipline pattern); the recover shim contains anything a
// toggle panics with (T-FG3-01).
func (it *Item) Activate(_, _ int32) (err *dbus.Error) {
	defer recoverMenuCall("Activate", &err)

	it.mu.Lock()
	toggle := it.cb.Toggle
	firstNil := toggle == nil && !it.activateWarned
	if firstNil {
		it.activateWarned = true
	}
	it.mu.Unlock()

	if toggle == nil {
		if firstNil {
			slog.Warn("tray indicator activate unavailable",
				"component", "tray indicator", "reason", "no toggle callback")
		}

		return nil
	}
	toggle()

	return nil
}

// emitNewIcon sends the NewIcon signal; a failed emit is the one-WARN
// self-disable — permanent per connection. The caller holds the mutex.
func (it *Item) emitNewIcon() {
	if err := it.emitter.Emit(it.path, sniIface, signalNewIcon); err != nil {
		it.emitDead = true
		if !it.emitWarned {
			it.emitWarned = true
			slog.Warn("tray indicator emit failed", "component", "tray indicator", "error", err)
		}
	}
}

// revive marks the item registered with the watcher again and refreshes the
// shell's copy with ONE NewIcon emit — the pix is already current (the
// lost-period ModeChanged calls kept swapping it); a failed revival emit
// follows the emitDead path. The supervisor calls it after a successful
// re-attach.
func (it *Item) revive() {
	it.mu.Lock()
	defer it.mu.Unlock()

	it.registered = true
	it.emitNewIcon()
}

// markUnregistered clears the watcher-registration state — the supervisor
// calls it when a beat proves the watcher no longer holds the item.
// ModeChanged keeps the pixmap fresh but stops emitting (an unregistered
// item has no shell listener).
func (it *Item) markUnregistered() {
	it.mu.Lock()
	defer it.mu.Unlock()

	it.registered = false
}

// menuFailed degrades the menu decision to the icon-only sentinel: exactly
// one WARN per item lifetime, and the registration CONTINUES — the
// sentinel is what keeps the watcher from destroying an icon-only item.
func (it *Item) menuFailed(err error) {
	it.mu.Lock()
	defer it.mu.Unlock()

	it.menuPath = menuNoDBusMenu
	if it.menuWarned {
		return
	}
	it.menuWarned = true
	slog.Warn("tray indicator menu export failed", "component", "tray indicator", "error", err)
}

// recoverMenuCall contains a panic raised anywhere below a menu click or an
// item Activate: the panic is swallowed and logged at ERROR with its stack,
// and the caller answers with a D-Bus error — a click can never kill the
// daemon (the ctlsvc recoverMethod precedent, T-FG3-01).
func recoverMenuCall(method string, dbusErr **dbus.Error) {
	if r := recover(); r != nil {
		slog.Error("menu call panic contained",
			"method", method,
			"panic", fmt.Sprint(r),
			"stack", string(debug.Stack()))
		*dbusErr = dbus.NewError(errNameFailed, []any{"internal error in " + method})
	}
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
		propMenu:       dbus.MakeVariant(dbus.ObjectPath(it.menuPath)),
	}
}
