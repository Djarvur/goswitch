package indicator

import (
	"sync"

	"github.com/godbus/dbus/v5"
)

// The DBusMenu wire constants: the property names the extension reads, the
// static values they serve, the event the menu dispatches on, the pinned
// layout ids of the canonical menu (07-CONTEXT «Approved Menu Layout»,
// 287999c — the e2e cases assert these exact ids) and the owner-facing
// labels. The old «Переключить раскладку» item is gone — replaced by the
// EN/RU radio pair.
const (
	menuPropVersion = "Version"
	menuPropStatus  = "Status"
	menuPropLabel   = "label"
	menuPropType    = "type"
	menuPropEnabled = "enabled"
	// The dynamic-menu additions (the renderer's property dictionary,
	// dbusMenu.js:74-90): the radio/checkmark ornaments (dbusMenu.js:731-739)
	// and the visibility switch (dbusMenu.js:782-784).
	menuPropVisible     = "visible"
	menuPropToggleType  = "toggle-type"
	menuPropToggleState = "toggle-state"

	menuTypeStandard  = "standard"
	menuTypeSeparator = "separator"
	menuStatusNormal  = "normal"

	menuToggleRadio     = "radio"
	menuToggleCheckmark = "checkmark"

	// menuVersion is the com.canonical.dbusmenu protocol version — 3, the
	// shape the toolkit's appindicator bridge implements.
	menuVersion = uint32(3)

	// menuRevisionBase is the revision the FIRST published layout carries;
	// the counter below bumps only on a change of the item SET (a macro
	// row appearing or disappearing), never on a property update.
	menuRevisionBase = uint32(1)

	eventClicked = "clicked"

	// The dbusmenu v3 signal names of the dynamic menu:
	// ItemsPropertiesUpdated is the ONLY channel of property dynamics (the
	// renderer flags updates behind a closed menu and replays on open),
	// LayoutUpdated forces the full re-read that follows a set change.
	signalItemsPropsUpdated = "ItemsPropertiesUpdated"
	signalLayoutUpdated     = "LayoutUpdated"

	menuIDRoot        = int32(0)
	menuIDEN          = int32(1)
	menuIDRU          = int32(2)
	menuIDSep1        = int32(3)
	menuIDACToggle    = int32(4)
	menuIDSoundToggle = int32(5)
	menuIDMacro1      = int32(6)
	menuIDMacro2      = int32(7)
	menuIDMacro3      = int32(8)
	menuIDMacroCombo  = int32(9)
	menuIDMacroChord  = int32(10)
	menuIDSep2        = int32(11)
	menuIDSettings    = int32(12)
	menuIDAbout       = int32(13)
	menuIDStatus      = int32(14)
	menuIDReload      = int32(15)

	labelEN          = "EN"
	labelRU          = "RU"
	labelACToggle    = "Автокоррекция"
	labelSoundToggle = "Звук"
	labelSettings    = "Настройки…"
	labelAbout       = "О программе"
	labelStatus      = "Статус"
	labelReload      = "Перечитать конфиг"

	// menuIDToggle, menuIDStatus and menuIDReload's v1 positions and the
	// labelToggle literal are gone with the static menu; the ids above are
	// the canonical layout's contract.
)

// removedProps is one entry of the ItemsPropertiesUpdated removed-props
// argument — the a(ias) wire shape. Every emit carries an EMPTY SLICE
// LITERAL of this type, never nil: the empty literal keeps the reflected
// signature a(ias) on the wire (Pitfall 7 — a nil argument breaks the
// godbus signature computation and panics the first live emit).
type removedProps struct {
	ID    int32
	Names []string
}

// menuLayout is one GetLayout node in the dbusmenu v3 wire shape (ia{sv}av):
// id, properties, children — each child a variant wrapping the same struct
// type. The field order IS the wire order (godbus computes the signature by
// reflection — pinned by the corpus).
type menuLayout struct {
	ID    int32
	Props map[string]dbus.Variant
	Kids  []dbus.Variant
}

// menuItemProps is one GetGroupProperties answer in the (ia{sv}) wire shape —
// and the per-item delta entry of ItemsPropertiesUpdated's first argument
// (the same (ia{sv}) pair on the update channel, research Pattern 1).
type menuItemProps struct {
	ID    int32
	Props map[string]dbus.Variant
}

// Menu is the com.canonical.dbusmenu object at /Menu — the tray's
// interactive surface (quick plan 261001-fg3): on GNOME's
// ubuntu-appindicators ANY click opens the menu, so the menu IS the
// interaction surface. Menu v2 (plan 07-05) serves the canonical layout
// from a MUTEX-GUARDED STATE SNAPSHOT (mode, both toggles, the raw key
// names, the version) that the actor and the daemon wiring push into; the
// callbacks it dispatches into own their synchronization (the actor's
// mutex), and the recover shim bounds whatever escapes them (T-FG3-01).
type Menu struct {
	cb Callbacks
	em Emitter // the signal channel — nil before wiring, emits degrade to no-ops

	mu           sync.Mutex
	mode         string // "en"/"ru" — the radio pair's mark
	acEnabled    bool   // the autocorrect toggle's applied value
	soundEnabled bool   // the sound toggle's applied value (default ON)
	tapKey       string // the RAW config names — the menu renders (mnemonics doubled)
	wordCombo    string
	modeChord    string
	version      string
	revision     uint32 // bumps ONLY on a change of the item set
	emitDead     bool   // a failed emit — permanent per connection (the Item discipline)
	emitWarned   bool   // one WARN per emit-degradation episode
}

// newMenu builds the menu with its startup snapshot: EN active (ADR-001 —
// the install push corrects any skew), sound ON (the owner's default-ON
// verdict — an absent section reads enabled), autocorrect OFF (D-54) and
// the revision at the base the first published layout carries.
func newMenu(cb Callbacks, em Emitter) *Menu {
	return &Menu{
		cb:           cb,
		em:           em,
		mode:         symbolEN,
		soundEnabled: true,
		revision:     menuRevisionBase,
	}
}

// Get serves org.freedesktop.DBus.Properties.Get for the menu — the
// dispatch re-entry the extension drives on its own goroutine. Error
// returns carry the zero Variant: MakeVariant(nil) would panic in the
// signature computation, and the value is unspecified whenever the error
// is non-nil anyway.
func (m *Menu) Get(iface, property string) (dbus.Variant, *dbus.Error) {
	if iface != menuIface {
		return dbus.Variant{}, dbus.NewError(errNameUnknownIface, []any{iface})
	}
	props := m.properties()
	v, ok := props[property]
	if !ok {
		return dbus.Variant{}, dbus.NewError(errNameUnknownProperty, []any{property})
	}

	return v, nil
}

// GetAll serves org.freedesktop.DBus.Properties.GetAll for the menu.
func (m *Menu) GetAll(iface string) (map[string]dbus.Variant, *dbus.Error) {
	if iface != menuIface {
		return nil, dbus.NewError(errNameUnknownIface, []any{iface})
	}

	return m.properties(), nil
}

// Set refuses: every menu property is read-only — state flows in through
// the setters, never a property write.
func (m *Menu) Set(iface, property string, _ dbus.Variant) *dbus.Error {
	return dbus.NewError(errNameReadOnly, []any{property + " on " + iface + " is read-only"})
}

// SetMode installs the observed mode symbol into the snapshot — the actor's
// menu-sync push (observer-last, the ModeChanged mirror). RED stub: the
// production emit discipline lands in GREEN.
func (m *Menu) SetMode(_ string) {}

// SetAutocorrectEnabled installs the applied autocorrect value — the
// applySnapshot fold and the click composition's push. RED stub.
func (m *Menu) SetAutocorrectEnabled(_ bool) {}

// SetSoundEnabled installs the applied sound value (the EffectiveEnabled
// truth — an absent section reads ON). RED stub.
func (m *Menu) SetSoundEnabled(_ bool) {}

// SetKeys installs the raw config key names of the macro rows — the actor
// hands CONFIG truth, the menu renders (the mnemonic doubling lives here).
// RED stub.
func (m *Menu) SetKeys(_, _, _ string) {}

// SetVersion installs the build identity (D-37) the About label serves.
// RED stub.
func (m *Menu) SetVersion(_ string) {}

// layoutRevision snapshots the revision counter — the corpus's read of the
// bump discipline (set changes only).
func (m *Menu) layoutRevision() uint32 {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.revision
}

// GetLayout implements com.canonical.dbusmenu.GetLayout. RED stub: the
// canonical layout is not served yet — the corpus proves its absence.
func (m *Menu) GetLayout(parentID, _ int32, _ []string) (uint32, menuLayout, *dbus.Error) {
	return m.revision, menuLayout{ID: parentID, Props: map[string]dbus.Variant{}}, nil
}

// GetGroupProperties implements com.canonical.dbusmenu.GetGroupProperties.
// RED stub: no item is served yet.
func (m *Menu) GetGroupProperties(_ []int32, _ []string) ([]menuItemProps, *dbus.Error) {
	return []menuItemProps{}, nil
}

// AboutToShow implements com.canonical.dbusmenu.AboutToShow: always false —
// honest signals keep the renderer current; nothing is ever pending here.
func (m *Menu) AboutToShow(_ int32) (bool, *dbus.Error) {
	return false, nil
}

// Event implements com.canonical.dbusmenu.Event: the click dispatch. RED
// stub: no item is dispatched yet.
func (m *Menu) Event(_ int32, _ string, _ dbus.Variant, _ uint32) (err *dbus.Error) {
	defer recoverMenuCall("Event", &err)

	return nil
}

// properties snapshots the served property set — the dbusmenu v3 minimum
// (IconThemePath skipped per the minimal brief).
func (m *Menu) properties() map[string]dbus.Variant {
	return map[string]dbus.Variant{
		menuPropVersion: dbus.MakeVariant(menuVersion),
		menuPropStatus:  dbus.MakeVariant(menuStatusNormal),
	}
}
