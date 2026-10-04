package indicator

import (
	"fmt"
	"log/slog"
	"strings"
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
	// the counter on Menu bumps only on a change of the item SET (a macro
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

	// The macro-row wrappers (canonical layout, research A4): RU text around
	// the RAW config key names — the menu renders, the actor never
	// transforms.
	macroComboWhat = " — слово и язык"
	macroChordWhat = " — смена режима"

	// The tap-series counts of the macro rows (canonical: 1× язык, 2×
	// слово, 3× фраза).
	macroTapLanguage = 1
	macroTapWord     = 2
	macroTapPhrase   = 3
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
// property answers are computed from the snapshot AT REQUEST TIME (the
// Item.properties precedent), and every snapshot change emits
// ItemsPropertiesUpdated through the menu's emitter — the renderer flags
// updates behind a closed menu and replays them on open (research Q1). The
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

// SetMode installs the observed mode symbol — the actor's menu-sync push
// (observer-last, the ModeChanged mirror): the radio pair re-marks itself
// from the same truth the icon uses. An identical value is a silent no-op
// (the actor pushes per record, the menu dedupes). One signal carries BOTH
// radio deltas and an empty removed slice (Pitfall 7).
func (m *Menu) SetMode(symbol string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if symbol == m.mode {
		return
	}
	m.mode = symbol
	m.emitSignal(signalItemsPropsUpdated,
		[]menuItemProps{
			m.stateDelta(menuIDEN, m.mode == symbolEN),
			m.stateDelta(menuIDRU, m.mode == symbolRU),
		},
		[]removedProps{})
}

// SetAutocorrectEnabled installs the applied autocorrect value — the
// applySnapshot fold and the click composition's push. One toggle-state
// delta, no revision bump.
func (m *Menu) SetAutocorrectEnabled(on bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if on == m.acEnabled {
		return
	}
	m.acEnabled = on
	m.emitSignal(signalItemsPropsUpdated,
		[]menuItemProps{m.stateDelta(menuIDACToggle, on)},
		[]removedProps{})
}

// SetSoundEnabled installs the applied sound value — the EffectiveEnabled
// truth of the applied snapshot (an absent section reads ON, the owner's
// default) pushed by the same fold, and the click composition's push. One
// toggle-state delta, no revision bump.
func (m *Menu) SetSoundEnabled(on bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if on == m.soundEnabled {
		return
	}
	m.soundEnabled = on
	m.emitSignal(signalItemsPropsUpdated,
		[]menuItemProps{m.stateDelta(menuIDSoundToggle, on)},
		[]removedProps{})
}

// SetKeys installs the raw config key names of the macro rows — the actor
// hands CONFIG truth, the menu renders (the mnemonic doubling lives in the
// label builders). A binding entering or leaving is a change of the item
// SET: the row disappears from the layout («отключённый режим — строка
// исчезает», canonical layout), the revision bumps and ONE LayoutUpdated
// forces the renderer's full re-read. Label-only changes ride
// ItemsPropertiesUpdated; identical values emit nothing at all (the actor
// pushes per fold, the menu dedupes).
func (m *Menu) SetKeys(tap, combo, chord string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tap == m.tapKey && combo == m.wordCombo && chord == m.modeChord {
		return
	}
	oldTap, oldCombo, oldChord := m.tapKey, m.wordCombo, m.modeChord
	m.tapKey, m.wordCombo, m.modeChord = tap, combo, chord
	if (combo == "") != (oldCombo == "") || (chord == "") != (oldChord == "") {
		m.revision++
		m.emitSignal(signalLayoutUpdated, m.revision, menuIDRoot)

		return
	}
	var deltas []menuItemProps
	if tap != oldTap {
		for id := menuIDMacro1; id <= menuIDMacro3; id++ {
			deltas = append(deltas, m.labelDelta(id))
		}
	}
	if combo != oldCombo && combo != "" {
		deltas = append(deltas, m.labelDelta(menuIDMacroCombo))
	}
	if chord != oldChord && chord != "" {
		deltas = append(deltas, m.labelDelta(menuIDMacroChord))
	}
	m.emitSignal(signalItemsPropsUpdated, deltas, []removedProps{})
}

// SetVersion installs the build identity (D-37) the About label serves —
// the wiring seeds it at attach. One label delta.
func (m *Menu) SetVersion(v string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if v == m.version {
		return
	}
	m.version = v
	m.emitSignal(signalItemsPropsUpdated,
		[]menuItemProps{m.labelDelta(menuIDAbout)},
		[]removedProps{})
}

// GetLayout implements com.canonical.dbusmenu.GetLayout: the canonical
// tree served from the snapshot AT REQUEST TIME. parentID 0 serves the
// root with the visible kids in canonical order (a macro row with an empty
// binding drops OUT of the layout — the disabled-mode reading); any other
// parent id serves that item alone with no kids; the depth argument is
// ignored — the menu is flat and full. The kid variants wrap CONCRETE
// menuLayout values (the zero-Variant discipline forbids MakeVariant(nil)).
func (m *Menu) GetLayout(parentID, _ int32, _ []string) (uint32, menuLayout, *dbus.Error) {
	layout := menuLayout{ID: parentID, Props: map[string]dbus.Variant{}}
	m.mu.Lock()
	defer m.mu.Unlock()

	if parentID != menuIDRoot {
		if p := m.itemPropsLocked(parentID); p != nil {
			layout.Props = p
		}

		return m.revision, layout, nil
	}
	for _, id := range m.rootKidsLocked() {
		layout.Kids = append(layout.Kids,
			dbus.MakeVariant(menuLayout{ID: id, Props: m.itemPropsLocked(id)}))
	}

	return m.revision, layout, nil
}

// GetGroupProperties implements com.canonical.dbusmenu.GetGroupProperties:
// the (id, props) pairs for the requested known ids in request order —
// unknown ids are skipped; a hidden row answers visible:false (the
// renderer may still ask — the defensive branch the corpus pins).
func (m *Menu) GetGroupProperties(ids []int32, _ []string) ([]menuItemProps, *dbus.Error) {
	out := make([]menuItemProps, 0, len(ids))
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, id := range ids {
		if id < menuIDEN || id > menuIDReload {
			continue
		}
		out = append(out, menuItemProps{ID: id, Props: m.itemPropsLocked(id)})
	}

	return out, nil
}

// AboutToShow implements com.canonical.dbusmenu.AboutToShow: always false —
// honest signals keep the renderer current; nothing is ever pending here.
func (m *Menu) AboutToShow(_ int32) (bool, *dbus.Error) {
	return false, nil
}

// Event implements com.canonical.dbusmenu.Event: the click dispatch of the
// canonical set — the EN/RU pair through the SAME Switch seam every gesture
// shares (the actor's flipTo path, ADR-006 — never a second flip
// mechanism), both persisted toggles, the settings launcher, the version
// notification, the status notification and the config reload. Each is a
// silent no-op without its callback. Unknown ids and unknown event ids are
// ignored. The recover shim means a panic below a click can never kill the
// daemon (T-FG3-01, T-07-05-04).
func (m *Menu) Event(id int32, eventID string, _ dbus.Variant, _ uint32) (err *dbus.Error) {
	defer recoverMenuCall("Event", &err)

	if eventID != eventClicked {
		return nil
	}
	m.dispatchClick(id)

	return nil
}

// dispatchClick fires one clicked item's callback — every nil callback is a
// silent no-op (the disabled item's gesture). The two radio halves share
// the Switch seam; every other item owns its callback.
//
//nolint:cyclop // one case per canonical item — the dispatch table is the truth
func (m *Menu) dispatchClick(id int32) {
	switch id {
	case menuIDEN, menuIDRU:
		if m.cb.Switch != nil {
			target := symbolEN
			if id == menuIDRU {
				target = symbolRU
			}
			m.cb.Switch(target)
		}
	case menuIDACToggle:
		if m.cb.ToggleAutocorrect != nil {
			m.cb.ToggleAutocorrect()
		}
	case menuIDSoundToggle:
		if m.cb.ToggleSound != nil {
			m.cb.ToggleSound()
		}
	case menuIDSettings:
		if m.cb.Settings != nil {
			m.cb.Settings()
		}
	case menuIDAbout:
		if m.cb.About != nil {
			m.cb.About()
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
}

// layoutRevision snapshots the revision counter — the corpus's read of the
// bump discipline (set changes only).
func (m *Menu) layoutRevision() uint32 {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.revision
}

// rootKidsLocked returns the visible item ids in canonical order — a macro
// row with an EMPTY binding is absent from the layout entirely. The caller
// holds the mutex.
func (m *Menu) rootKidsLocked() []int32 {
	ids := []int32{
		menuIDEN, menuIDRU, menuIDSep1, menuIDACToggle, menuIDSoundToggle,
		menuIDMacro1, menuIDMacro2, menuIDMacro3,
	}
	if m.wordCombo != "" {
		ids = append(ids, menuIDMacroCombo)
	}
	if m.modeChord != "" {
		ids = append(ids, menuIDMacroChord)
	}

	return append(ids, menuIDSep2, menuIDSettings, menuIDAbout, menuIDStatus, menuIDReload)
}

// itemPropsLocked is the per-id property builder — the canonical layout's
// ornaments and greyness read straight off the snapshot. The caller holds
// the mutex; an unknown id serves nil (the callers skip it).
func (m *Menu) itemPropsLocked(id int32) map[string]dbus.Variant {
	switch id {
	case menuIDEN, menuIDRU:
		return m.radioProps(id)
	case menuIDACToggle, menuIDSoundToggle:
		return m.checkProps(id)
	case menuIDSep1, menuIDSep2:
		return separatorProps()
	case menuIDMacro1, menuIDMacro2, menuIDMacro3:
		return m.macroTapProps(id)
	case menuIDMacroCombo:
		return bindingProps(m.wordCombo, macroComboWhat)
	case menuIDMacroChord:
		return bindingProps(m.modeChord, macroChordWhat)
	case menuIDSettings:
		return standardProps(labelSettings, true)
	case menuIDAbout:
		return standardProps(aboutLabel(m.version), true)
	case menuIDStatus:
		return standardProps(labelStatus, true)
	case menuIDReload:
		return standardProps(labelReload, m.cb.Reload != nil)
	}

	return nil
}

// radioProps builds one EN/RU radio item: the mark IS the current mode.
func (m *Menu) radioProps(id int32) map[string]dbus.Variant {
	if id == menuIDRU {
		return toggleProps(labelRU, menuToggleRadio, m.mode == symbolRU)
	}

	return toggleProps(labelEN, menuToggleRadio, m.mode == symbolEN)
}

// checkProps builds one checkmark toggle: the applied value from the
// snapshot — autocorrect (D-54 default off) and sound (default ON).
func (m *Menu) checkProps(id int32) map[string]dbus.Variant {
	if id == menuIDSoundToggle {
		return toggleProps(labelSoundToggle, menuToggleCheckmark, m.soundEnabled)
	}

	return toggleProps(labelACToggle, menuToggleCheckmark, m.acEnabled)
}

// macroTapProps builds one tap-series info row: «N× <tap> — <что>» with the
// RAW key name rendered through the mnemonic doubling. The three rows are
// ALWAYS present — tap_key is schema-required (validation refuses an empty
// name at Load).
func (m *Menu) macroTapProps(id int32) map[string]dbus.Variant {
	n, what := macroTapSeries(id)

	return macroProps(fmt.Sprintf("%d× %s — %s", n, doubleUnderscores(m.tapKey), what))
}

// macroTapSeries returns one tap-series row's count and tail in the
// canonical reading: one tap switches the language, two corrects the word,
// three the phrase.
func macroTapSeries(id int32) (count int, what string) {
	switch id {
	case menuIDMacro2:
		return macroTapWord, "слово"
	case menuIDMacro3:
		return macroTapPhrase, "фраза"
	}

	return macroTapLanguage, "язык"
}

// bindingProps builds one conditional binding row: an EMPTY binding is the
// disabled state — visible:false (asked defensively; the layout omits the
// row entirely).
func bindingProps(binding, what string) map[string]dbus.Variant {
	if binding == "" {
		return hiddenProps()
	}

	return macroProps(doubleUnderscores(binding) + what)
}

// properties snapshots the served property set — the dbusmenu v3 minimum
// (IconThemePath skipped per the minimal brief).
func (m *Menu) properties() map[string]dbus.Variant {
	return map[string]dbus.Variant{
		menuPropVersion: dbus.MakeVariant(menuVersion),
		menuPropStatus:  dbus.MakeVariant(menuStatusNormal),
	}
}

// stateDelta is one toggle-state delta — the ItemsPropertiesUpdated payload
// of an ornament change.
func (m *Menu) stateDelta(id int32, on bool) menuItemProps {
	return menuItemProps{ID: id, Props: map[string]dbus.Variant{
		menuPropToggleState: dbus.MakeVariant(stateInt(on)),
	}}
}

// labelDelta is one label delta — the ItemsPropertiesUpdated payload of a
// text change, computed from the current snapshot. The caller holds the
// mutex.
func (m *Menu) labelDelta(id int32) menuItemProps {
	return menuItemProps{ID: id, Props: map[string]dbus.Variant{
		menuPropLabel: m.itemPropsLocked(id)[menuPropLabel],
	}}
}

// emitSignal sends one menu signal — a queued send, no round trip; a failed
// emit is the one-WARN self-disable (the Item's dead-connection discipline:
// the menu keeps SERVING correct state, it only stops pushing). A nil
// emitter is the pre-wiring degradation: the snapshot still updates, the
// push is a no-op. The caller holds the mutex.
func (m *Menu) emitSignal(signal string, args ...any) {
	if m.em == nil || m.emitDead {
		return
	}
	if err := m.em.Emit(menuPath, menuIface, signal, args...); err != nil {
		m.emitDead = true
		if !m.emitWarned {
			m.emitWarned = true
			slog.Warn("tray menu emit failed", "component", "tray indicator", "error", err)
		}
	}
}

// baseProps is the property set every standard item shares — visible is
// served explicitly on every live item (the renderer defaults it true; the
// explicit value keeps the hidden branch the only visible:false carrier).
func baseProps(label string, enabled, visible bool) map[string]dbus.Variant {
	return map[string]dbus.Variant{
		menuPropLabel:   dbus.MakeVariant(label),
		menuPropType:    dbus.MakeVariant(menuTypeStandard),
		menuPropEnabled: dbus.MakeVariant(enabled),
		menuPropVisible: dbus.MakeVariant(visible),
	}
}

// standardProps is a clickable standard item.
func standardProps(label string, enabled bool) map[string]dbus.Variant {
	return baseProps(label, enabled, true)
}

// toggleProps adds the radio/checkmark ornament pair to a standard item —
// the renderer's _updateOrnament reads exactly these two keys
// (dbusMenu.js:731-739).
func toggleProps(label, kind string, on bool) map[string]dbus.Variant {
	p := standardProps(label, true)
	p[menuPropToggleType] = dbus.MakeVariant(kind)
	p[menuPropToggleState] = dbus.MakeVariant(stateInt(on))

	return p
}

// macroProps is a grey info row — enabled:false, never dead: it carries the
// live gesture reference the menu exists to show (canonical layout).
func macroProps(label string) map[string]dbus.Variant {
	return baseProps(label, false, true)
}

// hiddenProps is the defensive answer about a row whose binding is empty.
func hiddenProps() map[string]dbus.Variant {
	return baseProps("", false, false)
}

// separatorProps is a type=separator row.
func separatorProps() map[string]dbus.Variant {
	return map[string]dbus.Variant{
		menuPropLabel:   dbus.MakeVariant(""),
		menuPropType:    dbus.MakeVariant(menuTypeSeparator),
		menuPropEnabled: dbus.MakeVariant(true),
		menuPropVisible: dbus.MakeVariant(true),
	}
}

// stateInt maps a boolean onto the toggle-state wire type (i).
func stateInt(on bool) int32 {
	if on {
		return 1
	}

	return 0
}

// doubleUnderscores escapes GTK mnemonics for the renderer: dbusMenu.js
// cuts single underscores with /_([^_])/, '$1' (Pitfall 1, VERIFIED
// dbusMenu.js:746), so every key name is doubled before it enters a label
// (shift_r serves as shift__r and RENDERS shift_r).
func doubleUnderscores(name string) string {
	return strings.ReplaceAll(name, "_", "__")
}

// aboutLabel renders the About item: the version rides the label (the
// canonical layout shows it inline), the notification carries the rest.
func aboutLabel(v string) string {
	if v == "" {
		return labelAbout
	}

	return labelAbout + " — goswitch " + v
}
