//nolint:testpackage // drives the unexported menu seam and wire constants — the sanctioned in-package corpus seam
package indicator

import (
	"strings"
	"testing"

	"github.com/godbus/dbus/v5"
)

// TestMenuProperties pins the com.canonical.dbusmenu v3 property surface:
// Version 3, Status "normal" (IconThemePath skipped per the minimal brief),
// the read-only refusal of Set, and the standard unknown-property and
// unknown-interface error names.
func TestMenuProperties(t *testing.T) {
	m := &Menu{}

	v, derr := m.Get(menuIface, menuPropVersion)
	if derr != nil || v.Value() != menuVersion {
		t.Errorf("Get(Version) = (%v, %v), want %d", v.Value(), derr, menuVersion)
	}
	if v, derr := m.Get(menuIface, menuPropStatus); derr != nil || v.Value() != menuStatusNormal {
		t.Errorf("Get(Status) = (%v, %v), want %q", v.Value(), derr, menuStatusNormal)
	}
	if _, derr := m.Get(menuIface, "Nope"); derr == nil {
		t.Error("Get of an unknown property returned no error")
	}
	if _, derr := m.Get(sniIface, menuPropVersion); derr == nil {
		t.Error("Get of a foreign interface returned no error")
	}
	if derr := m.Set(menuIface, menuPropVersion, dbus.MakeVariant(uint32(9))); derr == nil {
		t.Error("Set returned no error — every menu property is read-only")
	}
	all, derr := m.GetAll(menuIface)
	if derr != nil {
		t.Fatalf("GetAll error = %v, want nil", derr)
	}
	for _, want := range []string{menuPropVersion, menuPropStatus} {
		if _, ok := all[want]; !ok {
			t.Errorf("GetAll is missing %s", want)
		}
	}
	if len(all) != 2 {
		t.Errorf("GetAll serves %d properties, want exactly 2 (the dbusmenu v3 minimum)", len(all))
	}
}

// menuKidWant is one layout node's expected identity: the id, the label,
// the item type, the enabled and visible flags and the toggle ornament
// (type + state) — the full renderer dictionary the extension reads.
type menuKidWant struct {
	id          int32
	label       string
	typ         string
	enabled     bool
	visible     bool
	toggleType  string
	toggleState int32
}

// checkMenuKid asserts one layout node against its expected identity and
// the flat no-kids shape.
func checkMenuKid(t *testing.T, node menuLayout, want menuKidWant) {
	t.Helper()

	if node.ID != want.id {
		t.Errorf("node id = %d, want %d", node.ID, want.id)
	}
	if node.Props[menuPropLabel].Value() != want.label {
		t.Errorf("node %d label = %v, want %q", node.ID, node.Props[menuPropLabel].Value(), want.label)
	}
	if node.Props[menuPropType].Value() != want.typ {
		t.Errorf("node %d type = %v, want %q", node.ID, node.Props[menuPropType].Value(), want.typ)
	}
	if node.Props[menuPropEnabled].Value() != want.enabled {
		t.Errorf("node %d enabled = %v, want %t", node.ID, node.Props[menuPropEnabled].Value(), want.enabled)
	}
	if node.Props[menuPropVisible].Value() != want.visible {
		t.Errorf("node %d visible = %v, want %t", node.ID, node.Props[menuPropVisible].Value(), want.visible)
	}
	// The absent ornament keys read as their zero values — the renderer's
	// PropertyStore defaults (toggle-type '' → ornament NONE, state 0).
	gotType, _ := node.Props[menuPropToggleType].Value().(string)
	if gotType != want.toggleType {
		t.Errorf("node %d toggle-type = %q, want %q", node.ID, gotType, want.toggleType)
	}
	gotState, _ := node.Props[menuPropToggleState].Value().(int32)
	if gotState != want.toggleState {
		t.Errorf("node %d toggle-state = %d, want %d", node.ID, gotState, want.toggleState)
	}
	if len(node.Kids) != 0 {
		t.Errorf("node carries %d kids, want none — the menu is flat", len(node.Kids))
	}
}

// menuFixture builds a fully seeded v2 menu over a fresh fake emitter —
// the canonical state the layout pins read: RU active, autocorrect ON,
// sound ON, the built-in key names, the dev version. The seeding emits are
// drained so every test starts from a quiet emitter.
func menuFixture(t *testing.T) (*Menu, *fakeEmitter) {
	t.Helper()

	em := &fakeEmitter{}
	m := newMenu(Callbacks{}, em)
	m.SetMode(symbolRU)
	m.SetAutocorrectEnabled(true)
	m.SetKeys("shift_r", "shift+ctrl_r", "super+space")
	m.SetVersion("dev")
	drainEmits(em)

	return m, em
}

// drainEmits drops every recorded emission — the per-test isolation of the
// signal corpus.
func drainEmits(em *fakeEmitter) {
	em.mu.Lock()
	em.calls = nil
	em.mu.Unlock()
}

// TestMenuLayout pins the canonical layout (07-CONTEXT «Approved Menu
// Layout», 287999c): the EN/RU radio pair marked with the current mode,
// the checkmark toggles «Автокоррекция» and «Звук» ADJACENT (the canonical
// order), the five grey macro info rows with live key names (the
// underscores doubled — Pitfall 1), two separators, the two standard items
// (Settings, Reload) and the two GREY live info rows — About (the build
// identity) and Status (the compact snapshot summary, owner UAT
// 2026-10-05) — every id pinned by constant, the (ia{sv}av) wire signature
// and the depth-ignoring flat shape intact.
func TestMenuLayout(t *testing.T) {
	m, _ := menuFixture(t)

	if got := dbus.SignatureOf(menuLayout{}); got != dbus.ParseSignatureMust("(ia{sv}av)") {
		t.Fatalf("menuLayout wire signature = %s, want (ia{sv}av)", got)
	}

	rev, root, derr := m.GetLayout(menuIDRoot, 0, nil)
	if derr != nil {
		t.Fatalf("GetLayout(0) error = %v, want nil", derr)
	}
	// The fixture's chord binding ENTERED after construction — a set change,
	// the one legal bump: the seeded canonical state carries revision 2.
	if rev != menuRevisionBase+1 {
		t.Errorf("layout revision = %d, want %d (the chord's appearance bumped it once)", rev, menuRevisionBase+1)
	}
	if root.ID != menuIDRoot {
		t.Fatalf("root layout id = %d, want %d", root.ID, menuIDRoot)
	}
	wantKids := []menuKidWant{
		{menuIDEN, labelEN, menuTypeStandard, true, true, menuToggleRadio, 0},
		{menuIDRU, labelRU, menuTypeStandard, true, true, menuToggleRadio, 1},
		{menuIDSep1, "", menuTypeSeparator, true, true, "", 0},
		{menuIDACToggle, labelACToggle, menuTypeStandard, true, true, menuToggleCheckmark, 1},
		{menuIDSoundToggle, labelSoundToggle, menuTypeStandard, true, true, menuToggleCheckmark, 1},
		{menuIDMacro1, "1× shift__r — язык", menuTypeStandard, false, true, "", 0},
		{menuIDMacro2, "2× shift__r — слово", menuTypeStandard, false, true, "", 0},
		{menuIDMacro3, "3× shift__r — фраза", menuTypeStandard, false, true, "", 0},
		{menuIDMacroCombo, "shift+ctrl__r — слово и язык", menuTypeStandard, false, true, "", 0},
		{menuIDMacroChord, "super+space — смена режима", menuTypeStandard, false, true, "", 0},
		{menuIDSep2, "", menuTypeSeparator, true, true, "", 0},
		{menuIDSettings, labelSettings, menuTypeStandard, true, true, "", 0},
		{menuIDAbout, "goswitch dev", menuTypeStandard, false, true, "", 0},
		{menuIDStatus, "RU · автокоррекция вкл · звук вкл · испр. 0", menuTypeStandard, false, true, "", 0},
		{menuIDReload, labelReload, menuTypeStandard, false, true, "", 0},
	}
	if len(root.Kids) != len(wantKids) {
		t.Fatalf("root layout = %d kids, want exactly %d (the canonical set)", len(root.Kids), len(wantKids))
	}
	for i, want := range wantKids {
		kid, ok := root.Kids[i].Value().(menuLayout)
		if !ok {
			t.Fatalf("kid %d does not wrap a menuLayout: %T", i, root.Kids[i].Value())
		}
		checkMenuKid(t, kid, want)
	}

	// The depth argument is ignored: the menu is flat and full.
	_, deep, derr := m.GetLayout(menuIDRoot, 9, nil)
	if derr != nil || len(deep.Kids) != len(wantKids) {
		t.Errorf("deep GetLayout changed the shape (err %v, %d kids)", derr, len(deep.Kids))
	}

	// A non-root parent id serves that item alone with no kids.
	_, one, derr := m.GetLayout(menuIDACToggle, 0, nil)
	if derr != nil {
		t.Fatalf("GetLayout(%d) error = %v, want nil", menuIDACToggle, derr)
	}
	checkMenuKid(t, one, wantKids[3])
}

// TestMenuMacroVisibility pins the disappearing-row contract (canonical
// layout: «отключённый режим — строка исчезает»): a macro row with an EMPTY
// binding is absent from the layout entirely (visible:false served
// defensively if asked), its appearance/disappearance is a SET change —
// LayoutUpdated plus a revision bump — and a SetKeys with unchanged values
// emits nothing at all. The combo branch is defensive (the schema refuses
// an empty word_layout_combo on Load) and is exercised here DIRECTLY.
func TestMenuMacroVisibility(t *testing.T) {
	em := &fakeEmitter{}
	m := newMenu(Callbacks{Reload: func() {}}, em)

	// The chord binding appears: a set change — the revision bumps and the
	// row enters the layout.
	m.SetKeys("shift_r", "shift+ctrl_r", "super+space")
	_, root, _ := m.GetLayout(menuIDRoot, 0, nil)
	if rev := m.layoutRevision(); rev != menuRevisionBase+1 {
		t.Errorf("revision after the chord appeared = %d, want %d (the set changed)", rev, menuRevisionBase+1)
	}
	if !rootHasID(root, menuIDMacroChord) {
		t.Error("the chord row is missing from the layout despite the non-empty binding")
	}

	// The chord binding empties: the row DISAPPEARS from the layout — the
	// canonical disabled-mode reading.
	m.SetKeys("shift_r", "shift+ctrl_r", "")
	_, root, _ = m.GetLayout(menuIDRoot, 0, nil)
	if rootHasID(root, menuIDMacroChord) {
		t.Error("the chord row is still served despite the empty binding — it must disappear")
	}
	// Asked defensively, the hidden row answers visible:false.
	props, derr := m.GetGroupProperties([]int32{menuIDMacroChord}, nil)
	if derr != nil || len(props) != 1 {
		t.Fatalf("GetGroupProperties(hidden chord) = (%v, %v), want one entry", props, derr)
	}
	if props[0].Props[menuPropVisible].Value() != false {
		t.Errorf("hidden chord visible = %v, want false", props[0].Props[menuPropVisible].Value())
	}

	// The defensive combo branch: an empty combo binding hides its row the
	// same way (unreachable through Load, corpus-tested directly).
	m.SetKeys("shift_r", "", "")
	_, root, _ = m.GetLayout(menuIDRoot, 0, nil)
	if rootHasID(root, menuIDMacroCombo) {
		t.Error("the combo row is still served despite the empty binding")
	}
	if rootHasID(root, menuIDMacroChord) {
		t.Error("the chord row returned despite the empty binding")
	}
	if rev := m.layoutRevision(); rev != menuRevisionBase+3 {
		t.Errorf("revision after both rows disappeared = %d, want %d", rev, menuRevisionBase+3)
	}

	// The tap-series rows NEVER disappear (tap_key is schema-required): an
	// empty name still serves the three rows.
	if !rootHasID(root, menuIDMacro1) || !rootHasID(root, menuIDMacro2) || !rootHasID(root, menuIDMacro3) {
		t.Error("the tap-series rows must always be present")
	}

	// A SetKeys with unchanged values is a silent no-op — the actor pushes
	// per fold, the menu dedupes.
	drainEmits(em)
	m.SetKeys("shift_r", "", "")
	if got := len(em.emitCalls()); got != 0 {
		t.Errorf("unchanged SetKeys emitted %d signals, want none", got)
	}
}

// rootHasID reports whether the root layout serves the item id.
func rootHasID(root menuLayout, id int32) bool {
	for _, kid := range root.Kids {
		if node, ok := kid.Value().(menuLayout); ok && node.ID == id {
			return true
		}
	}

	return false
}

// signalDeltas extracts the updated-props payload of one recorded emit —
// the []menuItemProps the corpus asserts on.
func signalDeltas(t *testing.T, call emitCall) []menuItemProps {
	t.Helper()

	deltas, ok := call.args[0].([]menuItemProps)
	if !ok {
		t.Fatalf("updated-args type = %T, want []menuItemProps", call.args[0])
	}

	return deltas
}

// signalRemoved extracts and validates the removed-props payload: the EMPTY
// non-nil slice literal the wire shape demands (Pitfall 7).
func signalRemoved(t *testing.T, call emitCall) []removedProps {
	t.Helper()

	removed, ok := call.args[1].([]removedProps)
	if !ok {
		t.Fatalf("removed-args type = %T, want []removedProps (never nil)", call.args[1])
	}
	if removed == nil || len(removed) != 0 {
		t.Errorf("removed = (nil: %t, len %d), want an EMPTY non-nil slice literal (Pitfall 7)",
			removed == nil, len(removed))
	}

	return removed
}

// TestMenuWireSignatures pins the dynamic channel's payload signatures
// (Pitfall 7): the godbus reflection computes a(ia{sv}) for the updated
// props and a(ias) for the removed props — the wire shapes the renderer
// decodes.
func TestMenuWireSignatures(t *testing.T) {
	if got := dbus.SignatureOf([]menuItemProps{}); got != dbus.ParseSignatureMust("a(ia{sv})") {
		t.Fatalf("updated-props signature = %s, want a(ia{sv})", got)
	}
	if got := dbus.SignatureOf([]removedProps{}); got != dbus.ParseSignatureMust("a(ias)") {
		t.Fatalf("removed-props signature = %s, want a(ias)", got)
	}
}

// TestMenuSignalFormPins pins the dynamic channel's byte form (research
// Pattern 1, Pitfall 7): the SetMode push is EXACTLY ONE
// ItemsPropertiesUpdated carrying BOTH radio deltas, the Status info row's
// label delta (the mode rides the summary too — owner UAT 2026-10-05) and
// an empty removed slice — the a(ias) signature the godbus reflection
// computes from the literal — and a property update NEVER bumps the layout
// revision.
func TestMenuSignalFormPins(t *testing.T) {
	em := &fakeEmitter{}
	m := newMenu(Callbacks{}, em)

	m.SetMode(symbolRU)
	calls := em.emitCalls()
	if len(calls) != 1 {
		t.Fatalf("SetMode emitted %d signals, want exactly one", len(calls))
	}
	call := calls[0]
	if call.path != menuPath || call.iface != menuIface || call.signal != signalItemsPropsUpdated {
		t.Errorf("emit = %s %s.%s, want %s %s.ItemsPropertiesUpdated",
			call.path, call.iface, call.signal, menuPath, menuIface)
	}
	if len(call.args) != 2 {
		t.Fatalf("signal arity = %d, want 2 (updated, removed)", len(call.args))
	}
	deltas := signalDeltas(t, call)
	if len(deltas) != 3 || deltas[0].ID != menuIDEN || deltas[1].ID != menuIDRU || deltas[2].ID != menuIDStatus {
		t.Fatalf("SetMode deltas = %+v, want the EN/RU toggle-state pair plus the Status label", deltas)
	}
	for _, d := range deltas[:2] {
		if d.Props[menuPropToggleState].Value() == nil {
			t.Errorf("delta %d carries no toggle-state", d.ID)
		}
	}
	if deltas[2].Props[menuPropLabel].Value() == nil {
		t.Errorf("delta %d carries no label", deltas[2].ID)
	}
	signalRemoved(t, call)
	if rev := m.layoutRevision(); rev != menuRevisionBase {
		t.Errorf("revision after a property update = %d, want the unchanged %d", rev, menuRevisionBase)
	}
}

// TestMenuToggleAndVersionDeltas pins the push discipline of the toggles
// and the version: each toggle setter emits exactly one signal with its
// ornament delta plus the Status row's label delta (the summary follows the
// applied values — owner UAT 2026-10-05); the version emits exactly the
// About row's label delta.
func TestMenuToggleAndVersionDeltas(t *testing.T) {
	em := &fakeEmitter{}
	m := newMenu(Callbacks{}, em)

	m.SetAutocorrectEnabled(true)
	m.SetSoundEnabled(false)
	calls := em.emitCalls()
	if len(calls) != 2 {
		t.Fatalf("toggle pushes emitted %d signals, want exactly 2", len(calls))
	}
	first := signalDeltas(t, calls[0])
	if len(first) != 2 || first[0].ID != menuIDACToggle || first[1].ID != menuIDStatus {
		t.Errorf("autocorrect deltas = %v, want ids %d then %d", calls[0].args[0], menuIDACToggle, menuIDStatus)
	}
	second := signalDeltas(t, calls[1])
	if len(second) != 2 || second[0].ID != menuIDSoundToggle || second[1].ID != menuIDStatus {
		t.Errorf("sound deltas = %v, want ids %d then %d", calls[1].args[0], menuIDSoundToggle, menuIDStatus)
	}
	signalRemoved(t, calls[0])

	drainEmits(em)
	m.SetVersion("1.2.3")
	calls = em.emitCalls()
	if len(calls) != 1 {
		t.Fatalf("SetVersion emitted %d signals, want exactly one", len(calls))
	}
	about := signalDeltas(t, calls[0])
	if len(about) != 1 || about[0].ID != menuIDAbout {
		t.Errorf("version delta = %v, want exactly id %d", calls[0].args[0], menuIDAbout)
	}
	if label, _ := about[0].Props[menuPropLabel].Value().(string); label != "goswitch 1.2.3" {
		t.Errorf("About label = %q, want the greyed identity row goswitch 1.2.3", label)
	}
}

// TestMenuStatusRowDeltas pins the Status info row's live label (owner UAT
// 2026-10-05): the corrections push emits exactly one Status label delta
// with the fresh count, identical values emit nothing, and the served row
// renders the snapshot compactly — mode, both toggles, the counter.
func TestMenuStatusRowDeltas(t *testing.T) {
	em := &fakeEmitter{}
	m := newMenu(Callbacks{}, em)

	m.SetCorrections(5)
	calls := em.emitCalls()
	if len(calls) != 1 {
		t.Fatalf("SetCorrections emitted %d signals, want exactly one", len(calls))
	}
	deltas := signalDeltas(t, calls[0])
	if len(deltas) != 1 || deltas[0].ID != menuIDStatus {
		t.Fatalf("corrections delta = %v, want exactly id %d", calls[0].args[0], menuIDStatus)
	}
	if label, _ := deltas[0].Props[menuPropLabel].Value().(string); !strings.Contains(label, "испр. 5") {
		t.Errorf("Status label = %q, want the counter inside", label)
	}
	signalRemoved(t, calls[0])

	// The actor pushes per fold, the menu dedupes.
	drainEmits(em)
	m.SetCorrections(5)
	if got := len(em.emitCalls()); got != 0 {
		t.Errorf("unchanged SetCorrections emitted %d signals, want none", got)
	}

	// The served row composes the whole summary from the snapshot.
	m.SetMode(symbolRU)
	props, derr := m.GetGroupProperties([]int32{menuIDStatus}, nil)
	if derr != nil || len(props) != 1 {
		t.Fatalf("GetGroupProperties(Status) = (%v, %v), want one entry", props, derr)
	}
	want := "RU · автокоррекция выкл · звук вкл · испр. 5"
	if label, _ := props[0].Props[menuPropLabel].Value().(string); label != want {
		t.Errorf("Status label = %q, want %q", label, want)
	}
	if props[0].Props[menuPropEnabled].Value() != false {
		t.Error("the Status info row must answer disabled — a grey row, never clickable")
	}
}

// TestMenuUnderscoreDoubling pins the mnemonic discipline (Pitfall 1): the
// renderer cuts single underscores with /_([^_])/, '$1' (dbusMenu.js:746),
// so every key name enters its label DOUBLED — shift_r serves as shift__r.
func TestMenuUnderscoreDoubling(t *testing.T) {
	m, _ := menuFixture(t)

	_, root, derr := m.GetLayout(menuIDRoot, 0, nil)
	if derr != nil {
		t.Fatalf("GetLayout error = %v, want nil", derr)
	}
	var labels []string
	for _, kid := range root.Kids {
		if node, ok := kid.Value().(menuLayout); ok {
			if l, ok := node.Props[menuPropLabel].Value().(string); ok {
				labels = append(labels, l)
			}
		}
	}
	joined := strings.Join(labels, "|")
	if !strings.Contains(joined, "shift__r") {
		t.Errorf("no doubled shift__r in the labels %q — the mnemonic would eat the underscore", joined)
	}
	if strings.Contains(joined, "shift_r") {
		t.Errorf("a raw single-underscore shift_r leaked into the labels %q", joined)
	}
	if !strings.Contains(joined, "1× shift__r — язык") {
		t.Errorf("the tap-series label is not the canonical «1× <tap> — язык» form: %q", joined)
	}
}

// TestMenuNilEmitterDegradation pins the nil-emitter degradation: the
// setters update the snapshot WITHOUT emitting and never panic — a menu
// built before wiring stays a correct display, the Status row included.
func TestMenuNilEmitterDegradation(t *testing.T) {
	m := newMenu(Callbacks{}, nil)

	m.SetMode(symbolRU)
	m.SetAutocorrectEnabled(true)
	m.SetSoundEnabled(false)
	m.SetKeys("shift_r", "shift+ctrl_r", "super+space")
	m.SetVersion("dev")
	m.SetCorrections(7)

	_, root, derr := m.GetLayout(menuIDRoot, 0, nil)
	if derr != nil {
		t.Fatalf("GetLayout error = %v, want nil", derr)
	}
	if !rootHasID(root, menuIDRU) || !rootHasID(root, menuIDMacroChord) {
		t.Error("the snapshot was not updated without an emitter — the display must stay correct")
	}
	if rev := m.layoutRevision(); rev != menuRevisionBase+1 {
		t.Errorf("revision = %d, want %d — the chord appearance is a set change even unemitted",
			rev, menuRevisionBase+1)
	}
	props, derr := m.GetGroupProperties([]int32{menuIDStatus}, nil)
	if derr != nil || len(props) != 1 {
		t.Fatalf("GetGroupProperties(Status) = (%v, %v), want one entry", props, derr)
	}
	want := "RU · автокоррекция вкл · звук выкл · испр. 7"
	if label, _ := props[0].Props[menuPropLabel].Value().(string); label != want {
		t.Errorf("Status label without an emitter = %q, want %q", label, want)
	}
}

// TestMenuGetGroupProperties pins the (ia{sv}) batch read: the requested
// known ids come back with their props in order — the ornaments included —
// the reload item is disabled without callbacks, unknown ids are skipped,
// and a hidden row answers visible:false.
func TestMenuGetGroupProperties(t *testing.T) {
	m, _ := menuFixture(t)

	if got := dbus.SignatureOf(menuItemProps{}); got != dbus.ParseSignatureMust("(ia{sv})") {
		t.Fatalf("menuItemProps wire signature = %s, want (ia{sv})", got)
	}

	props, derr := m.GetGroupProperties([]int32{menuIDRU, menuIDACToggle, menuIDReload, 99}, nil)
	if derr != nil {
		t.Fatalf("GetGroupProperties error = %v, want nil", derr)
	}
	if len(props) != 3 {
		t.Fatalf("GetGroupProperties returned %d entries, want exactly 3 (the unknown id skipped)", len(props))
	}
	if props[0].ID != menuIDRU || props[0].Props[menuPropToggleState].Value() != int32(1) {
		t.Errorf("entry 0 = (id %d, state %v), want the marked RU radio",
			props[0].ID, props[0].Props[menuPropToggleState].Value())
	}
	if props[1].ID != menuIDACToggle || props[1].Props[menuPropToggleType].Value() != menuToggleCheckmark {
		t.Errorf("entry 1 = (id %d, toggle-type %v), want the autocorrect checkmark",
			props[1].ID, props[1].Props[menuPropToggleType].Value())
	}
	if props[2].ID != menuIDReload || props[2].Props[menuPropEnabled].Value() != false {
		t.Errorf("entry 2 = (id %d, enabled %v), want the reload item disabled without callbacks",
			props[2].ID, props[2].Props[menuPropEnabled].Value())
	}
}

// TestMenuAboutToShow pins the honest-signal answer: nothing is ever
// pending — the signals keep the renderer current.
func TestMenuAboutToShow(t *testing.T) {
	m, _ := menuFixture(t)

	update, derr := m.AboutToShow(menuIDRoot)
	if derr != nil || update {
		t.Errorf("AboutToShow = (%t, %v), want (false, nil)", update, derr)
	}
}

// TestMenuEventDispatch pins the clicked dispatch of the canonical set:
// id 1 → Switch("en"), 2 → Switch("ru"), 4 → ToggleAutocorrect,
// 5 → ToggleSound, 12 → Settings, 15 → Reload — each exactly once. The
// About and Status rows are greyed info rows (owner UAT 2026-10-05): they
// carry no callbacks and fire nothing.
func TestMenuEventDispatch(t *testing.T) {
	var switched []string
	acs, sounds, settings, reloads := 0, 0, 0, 0
	m := &Menu{cb: Callbacks{
		Switch:            func(target string) { switched = append(switched, target) },
		ToggleAutocorrect: func() { acs++ },
		ToggleSound:       func() { sounds++ },
		Settings:          func() { settings++ },
		Reload:            func() { reloads++ },
	}}

	for _, tc := range []struct {
		id      int32
		eventID string
	}{
		{menuIDEN, eventClicked},
		{menuIDRU, eventClicked},
		{menuIDACToggle, eventClicked},
		{menuIDSoundToggle, eventClicked},
		{menuIDSettings, eventClicked},
		{menuIDReload, eventClicked},
	} {
		if derr := m.Event(tc.id, tc.eventID, dbus.Variant{}, 0); derr != nil {
			t.Fatalf("click id %d error = %v, want nil", tc.id, derr)
		}
	}
	if got := strings.Join(switched, ","); got != "en,ru" {
		t.Errorf("switch targets = %q, want exactly en,ru", got)
	}
	if acs != 1 || sounds != 1 || settings != 1 || reloads != 1 {
		t.Errorf("click counts = (ac %d, sound %d, settings %d, reload %d), want all 1",
			acs, sounds, settings, reloads)
	}
}

// TestMenuEventNilAndUnknownSilent pins the no-op discipline: a nil callback
// is a silent no-op on every clickable id, the greyed About/Status info rows
// fire nothing at all, and unknown ids and unknown event ids are ignored —
// nothing fires anywhere.
func TestMenuEventNilAndUnknownSilent(t *testing.T) {
	var switched []string
	acs, sounds, settings, reloads := 0, 0, 0, 0
	m := &Menu{cb: Callbacks{
		Switch:            func(target string) { switched = append(switched, target) },
		ToggleAutocorrect: func() { acs++ },
		ToggleSound:       func() { sounds++ },
		Settings:          func() { settings++ },
		Reload:            func() { reloads++ },
	}}

	nilMenu := newMenu(Callbacks{}, nil)
	for _, id := range []int32{
		menuIDEN, menuIDRU, menuIDACToggle, menuIDSoundToggle,
		menuIDSettings, menuIDAbout, menuIDStatus,
	} {
		if derr := nilMenu.Event(id, eventClicked, dbus.Variant{}, 0); derr != nil {
			t.Errorf("nil-callback Event(%d) error = %v, want nil", id, derr)
		}
	}
	if derr := m.Event(99, eventClicked, dbus.Variant{}, 0); derr != nil {
		t.Errorf("unknown-id Event error = %v, want nil", derr)
	}
	if derr := m.Event(menuIDEN, "opened", dbus.Variant{}, 0); derr != nil {
		t.Errorf("unknown-event Event error = %v, want nil", derr)
	}
	if len(switched) != 0 || acs != 0 || sounds != 0 || settings != 0 || reloads != 0 {
		t.Error("ignored events fired callbacks")
	}
}

// TestMenuEventPanicContained pins the recover shim: a panicking click
// callback is contained — Event answers with a D-Bus error instead of
// killing the daemon, and the ERROR record carries the stack (T-FG3-01).
func TestMenuEventPanicContained(t *testing.T) {
	buf := captureLogs(t)
	m := &Menu{cb: Callbacks{Switch: func(string) { panic("boom") }}}

	derr := m.Event(menuIDEN, eventClicked, dbus.Variant{}, 0)

	if derr == nil {
		t.Fatal("panicking callback produced no error reply — the daemon would die on a click")
	}
	logged := buf.String()
	if !strings.Contains(logged, `"level":"ERROR"`) {
		t.Errorf("panic containment ERROR record missing; log:\n%s", logged)
	}
	if !strings.Contains(logged, "stack") {
		t.Errorf("panic ERROR record missing the stack; log:\n%s", logged)
	}
}

// TestMenuWireNamesPinned pins the DBusMenu wire constants against the REAL
// protocol literals — the 261001-fg3 lesson: a self-consistent corpus cannot
// catch a wrong interface NAME (the fakes compare the same constant), only a
// literal pin can. The DBusMenu protocol lives at com.canonical.dbusmenu
// (the org. prefix variant does not exist on any stack we target), the
// signal members carry their real dbusmenu v3 spellings, and the SNI
// surface stays org.kde.*.
func TestMenuWireNamesPinned(t *testing.T) {
	if menuIface != "com.canonical.dbusmenu" {
		t.Errorf("menuIface = %q, want the real protocol name com.canonical.dbusmenu", menuIface)
	}
	if signalItemsPropsUpdated != "ItemsPropertiesUpdated" || signalLayoutUpdated != "LayoutUpdated" {
		t.Errorf("signal names = %q/%q, want the real dbusmenu v3 members",
			signalItemsPropsUpdated, signalLayoutUpdated)
	}
	if watcherName != "org.kde.StatusNotifierWatcher" {
		t.Errorf("watcherName = %q, want org.kde.StatusNotifierWatcher", watcherName)
	}
	if sniIface != "org.kde.StatusNotifierItem" {
		t.Errorf("sniIface = %q, want org.kde.StatusNotifierItem", sniIface)
	}
	if string(itemPath) != "/StatusNotifierItem" || string(menuPath) != "/Menu" {
		t.Errorf("item/menu paths = %q/%q, want /StatusNotifierItem and /Menu", itemPath, menuPath)
	}
}
