package indicator

import (
	"github.com/godbus/dbus/v5"
)

// The DBusMenu wire constants: the property names the extension reads, the
// static values they serve, the event the menu dispatches on, the static
// layout ids (root, toggle FIRST, status, reload) and the three items'
// owner-facing labels.
const (
	menuPropVersion = "Version"
	menuPropStatus  = "Status"
	menuPropLabel   = "label"
	menuPropType    = "type"
	menuPropEnabled = "enabled"

	menuTypeStandard = "standard"
	menuStatusNormal = "normal"

	// menuVersion is the com.canonical.dbusmenu protocol version — 3, the
	// shape the toolkit's appindicator bridge implements.
	menuVersion = uint32(3)

	// menuRevision is the layout revision — a constant: the menu is static,
	// nothing ever invalidates it.
	menuRevision = uint32(1)

	eventClicked = "clicked"

	menuIDRoot   = int32(0)
	menuIDToggle = int32(1)
	menuIDStatus = int32(2)
	menuIDReload = int32(3)

	labelToggle = "Переключить раскладку"
	labelStatus = "Статус"
	labelReload = "Перечитать конфиг"
)

// menuLayout is one GetLayout node in the dbusmenu v3 wire shape (ia{sv}av):
// id, properties, children — each child a variant wrapping the same struct
// type. The field order IS the wire order (godbus computes the signature by
// reflection — pinned by the corpus).
type menuLayout struct {
	ID    int32
	Props map[string]dbus.Variant
	Kids  []dbus.Variant
}

// menuItemProps is one GetGroupProperties answer in the (ia{sv}) wire shape.
type menuItemProps struct {
	ID    int32
	Props map[string]dbus.Variant
}

// Menu is the com.canonical.dbusmenu object at /Menu — the tray's
// interactive surface (quick plan 261001-fg3): on GNOME's
// ubuntu-appindicators ANY click opens the menu, so the menu IS the
// interaction surface and the toggle is its FIRST item. The layout is
// static (a constant revision, AboutToShow always false); every served
// value is immutable, so the struct carries no mutex of its own — the
// callbacks it dispatches into own their synchronization (the actor's
// mutex), and the recover shim bounds whatever escapes them (T-FG3-01).
type Menu struct {
	cb Callbacks
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

// Set refuses: every menu property is read-only — the layout is static.
func (m *Menu) Set(iface, property string, _ dbus.Variant) *dbus.Error {
	return dbus.NewError(errNameReadOnly, []any{property + " on " + iface + " is read-only"})
}

// GetLayout implements com.canonical.dbusmenu.GetLayout: the static
// three-item tree. parentID 0 serves the root with the three kids (the
// toggle FIRST — the interaction surface); any other parent id serves that
// item with no kids; the depth argument is ignored — the menu is flat and
// full. The kid variants wrap CONCRETE menuLayout values (the zero-Variant
// discipline forbids MakeVariant(nil), and error paths keep the zero
// Variant).
func (m *Menu) GetLayout(parentID, _ int32, _ []string) (uint32, menuLayout, *dbus.Error) {
	layout := menuLayout{ID: parentID, Props: map[string]dbus.Variant{}}
	if parentID != menuIDRoot {
		layout.Props = menuProps(parentID, m.cb)

		return menuRevision, layout, nil
	}
	layout.Kids = []dbus.Variant{
		dbus.MakeVariant(menuLayout{ID: menuIDToggle, Props: menuProps(menuIDToggle, m.cb)}),
		dbus.MakeVariant(menuLayout{ID: menuIDStatus, Props: menuProps(menuIDStatus, m.cb)}),
		dbus.MakeVariant(menuLayout{ID: menuIDReload, Props: menuProps(menuIDReload, m.cb)}),
	}

	return menuRevision, layout, nil
}

// GetGroupProperties implements com.canonical.dbusmenu.GetGroupProperties:
// the (id, props) pairs for the requested known ids in request order —
// unknown ids are skipped.
func (m *Menu) GetGroupProperties(ids []int32, _ []string) ([]menuItemProps, *dbus.Error) {
	out := make([]menuItemProps, 0, len(ids))
	for _, id := range ids {
		if id == menuIDToggle || id == menuIDStatus || id == menuIDReload {
			out = append(out, menuItemProps{ID: id, Props: menuProps(id, m.cb)})
		}
	}

	return out, nil
}

// AboutToShow implements com.canonical.dbusmenu.AboutToShow: always false —
// the menu is static, nothing is ever pending.
func (m *Menu) AboutToShow(_ int32) (bool, *dbus.Error) {
	return false, nil
}

// Event implements com.canonical.dbusmenu.Event: the click dispatch. Only
// the "clicked" event carries an action: 1 toggles the mode (through the
// SAME Callbacks.Toggle the item's Activate drives), 2 posts the status
// notification, 3 reloads the config — each a silent no-op without its
// callback (the disabled reload item). Unknown ids and unknown event ids
// are ignored. The recover shim means a panic below a click can never kill
// the daemon (T-FG3-01).
func (m *Menu) Event(id int32, eventID string, _ dbus.Variant, _ uint32) (err *dbus.Error) {
	defer recoverMenuCall("Event", &err)

	if eventID != eventClicked {
		return nil
	}
	switch id {
	case menuIDToggle:
		if m.cb.Toggle != nil {
			m.cb.Toggle()
		}
	case menuIDStatus:
		if m.cb.Status != nil {
			m.cb.Status()
		}
	case menuIDReload:
		if m.cb.Reload != nil {
			m.cb.Reload()
		}
	}

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

// menuProps builds one item's served property set: the minimal
// (label, type, enabled) triple the extension renders. The reload item is
// enabled only when a reload callback exists (the daemon runs without
// -config otherwise) — served greyed, never dead.
func menuProps(id int32, cb Callbacks) map[string]dbus.Variant {
	label := ""
	enabled := true
	switch id {
	case menuIDToggle:
		label = labelToggle
	case menuIDStatus:
		label = labelStatus
	case menuIDReload:
		label = labelReload
		enabled = cb.Reload != nil
	}

	return map[string]dbus.Variant{
		menuPropLabel:   dbus.MakeVariant(label),
		menuPropType:    dbus.MakeVariant(menuTypeStandard),
		menuPropEnabled: dbus.MakeVariant(enabled),
	}
}
