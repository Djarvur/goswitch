//nolint:testpackage // drives the unexported menu seam and wire constants — the sanctioned in-package corpus seam
package indicator

import (
	"strings"
	"testing"

	"github.com/godbus/dbus/v5"
)

// TestMenuProperties pins the org.canonical.dbusmenu v3 property surface:
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

// TestMenuLayout pins the static layout: parentId 0 serves the root with
// the three kids in order — «Переключить раскладку» FIRST (on GNOME's
// ubuntu-appindicators ANY click opens the menu, so the menu IS the
// interaction surface), then «Статус», then «Перечитать конфиг» — with the
// exact labels, the enabled flags (reload greyed without callbacks), the
// flat depth-ignoring shape, and the (ia{sv}av) wire signature.
func TestMenuLayout(t *testing.T) {
	m := &Menu{} // zero callbacks: the reload item must come out disabled

	if got := dbus.Signature(menuLayout{}); got != dbus.Signature("(ia{sv}av)") {
		t.Fatalf("menuLayout wire signature = %s, want (ia{sv}av)", got)
	}

	rev, root, derr := m.GetLayout(menuIDRoot, 0, nil)
	if derr != nil {
		t.Fatalf("GetLayout(0) error = %v, want nil", derr)
	}
	if rev != menuRevision {
		t.Errorf("layout revision = %d, want the static %d", rev, menuRevision)
	}
	if root.Id != menuIDRoot || len(root.Kids) != 3 {
		t.Fatalf("root layout = (id %d, %d kids), want id 0 with exactly 3 kids", root.Id, len(root.Kids))
	}
	wantKids := []struct {
		id      int32
		label   string
		enabled bool
	}{
		{menuIDToggle, labelToggle, true},
		{menuIDStatus, labelStatus, true},
		{menuIDReload, labelReload, false},
	}
	for i, want := range wantKids {
		kid, ok := root.Kids[i].Value().(menuLayout)
		if !ok {
			t.Fatalf("kid %d does not wrap a menuLayout: %T", i, root.Kids[i].Value())
		}
		if kid.Id != want.id {
			t.Errorf("kid %d id = %d, want %d (the toggle must be FIRST)", i, kid.Id, want.id)
		}
		if kid.Props[menuPropLabel].Value() != want.label {
			t.Errorf("kid %d label = %v, want %q", i, kid.Props[menuPropLabel].Value(), want.label)
		}
		if kid.Props[menuPropType].Value() != menuTypeStandard {
			t.Errorf("kid %d type = %v, want %q", i, kid.Props[menuPropType].Value(), menuTypeStandard)
		}
		if kid.Props[menuPropEnabled].Value() != want.enabled {
			t.Errorf("kid %d enabled = %v, want %t", i, kid.Props[menuPropEnabled].Value(), want.enabled)
		}
		if len(kid.Kids) != 0 {
			t.Errorf("kid %d carries %d kids, want none — the menu is flat", i, len(kid.Kids))
		}
	}

	// The depth argument is ignored: the menu is flat and full.
	_, deep, derr := m.GetLayout(menuIDRoot, 9, nil)
	if derr != nil || len(deep.Kids) != 3 {
		t.Errorf("deep GetLayout changed the shape (err %v, %d kids)", derr, len(deep.Kids))
	}

	// A non-root parent id serves that item alone with no kids.
	_, one, derr := m.GetLayout(menuIDToggle, 0, nil)
	if derr != nil {
		t.Fatalf("GetLayout(1) error = %v, want nil", derr)
	}
	if one.Id != menuIDToggle || one.Props[menuPropLabel].Value() != labelToggle || len(one.Kids) != 0 {
		t.Errorf("GetLayout(1) = (id %d, label %v, %d kids), want the toggle item alone",
			one.Id, one.Props[menuPropLabel].Value(), len(one.Kids))
	}
}

// TestMenuGetGroupProperties pins the (ia{sv}) batch read: the requested
// known ids come back with their props in order; the reload item is
// disabled without callbacks; unknown ids are skipped.
func TestMenuGetGroupProperties(t *testing.T) {
	m := &Menu{} // zero callbacks

	if got := dbus.Signature(menuItemProps{}); got != dbus.Signature("(ia{sv})") {
		t.Fatalf("menuItemProps wire signature = %s, want (ia{sv})", got)
	}

	props, derr := m.GetGroupProperties([]int32{menuIDToggle, menuIDReload, 99}, nil)
	if derr != nil {
		t.Fatalf("GetGroupProperties error = %v, want nil", derr)
	}
	if len(props) != 2 {
		t.Fatalf("GetGroupProperties returned %d entries, want exactly 2 (the unknown id skipped)", len(props))
	}
	if props[0].Id != menuIDToggle || props[0].Props[menuPropLabel].Value() != labelToggle {
		t.Errorf("entry 0 = (id %d, label %v), want the toggle item", props[0].Id, props[0].Props[menuPropLabel].Value())
	}
	if props[1].Id != menuIDReload || props[1].Props[menuPropEnabled].Value() != false {
		t.Errorf("entry 1 = (id %d, enabled %v), want the reload item disabled without callbacks",
			props[1].Id, props[1].Props[menuPropEnabled].Value())
	}
}

// TestMenuAboutToShow pins the static-menu answer: nothing is ever pending.
func TestMenuAboutToShow(t *testing.T) {
	m := &Menu{}

	update, derr := m.AboutToShow(menuIDRoot)
	if derr != nil || update {
		t.Errorf("AboutToShow = (%t, %v), want (false, nil) — the menu is static", update, derr)
	}
}

// TestMenuEventDispatch pins the clicked dispatch: id 1 → Toggle, 2 →
// Status, 3 → Reload; a nil callback is a silent no-op (the disabled reload
// item); unknown ids and unknown event ids are ignored.
func TestMenuEventDispatch(t *testing.T) {
	toggles, statuses, reloads := 0, 0, 0
	m := &Menu{cb: Callbacks{
		Toggle: func() { toggles++ },
		Status: func() { statuses++ },
		Reload: func() { reloads++ },
	}}

	for _, tc := range []struct {
		id      int32
		eventID string
		want    func() int
		name    string
	}{
		{menuIDToggle, eventClicked, func() int { return toggles }, "toggle click"},
		{menuIDStatus, eventClicked, func() int { return statuses }, "status click"},
		{menuIDReload, eventClicked, func() int { return reloads }, "reload click"},
	} {
		if derr := m.Event(tc.id, tc.eventID, dbus.Variant{}, 0); derr != nil {
			t.Fatalf("%s error = %v, want nil", tc.name, derr)
		}
		if got := tc.want(); got != 1 {
			t.Errorf("%s fired %d times, want exactly one", tc.name, got)
		}
	}

	// A nil callback is a silent no-op; unknown ids and event ids ignored.
	nilMenu := &Menu{}
	if derr := nilMenu.Event(menuIDToggle, eventClicked, dbus.Variant{}, 0); derr != nil {
		t.Errorf("nil-callback Event error = %v, want nil", derr)
	}
	if derr := m.Event(99, eventClicked, dbus.Variant{}, 0); derr != nil {
		t.Errorf("unknown-id Event error = %v, want nil", derr)
	}
	if derr := m.Event(menuIDToggle, "opened", dbus.Variant{}, 0); derr != nil {
		t.Errorf("unknown-event Event error = %v, want nil", derr)
	}
	if toggles != 1 || statuses != 1 || reloads != 1 {
		t.Errorf("ignored events fired callbacks (toggles %d, statuses %d, reloads %d)",
			toggles, statuses, reloads)
	}
}

// TestMenuEventPanicContained pins the recover shim: a panicking callback is
// contained — Event answers with a D-Bus error instead of killing the
// daemon, and the ERROR record carries the stack (T-FG3-01, the ctlsvc
// recoverMethod precedent).
func TestMenuEventPanicContained(t *testing.T) {
	buf := captureLogs(t)
	m := &Menu{cb: Callbacks{Toggle: func() { panic("boom") }}}

	derr := m.Event(menuIDToggle, eventClicked, dbus.Variant{}, 0)

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

// TestItemActivateInvokesToggle pins the menu-less path: the SNI Activate
// method drives the same Callbacks.Toggle exactly once.
func TestItemActivateInvokesToggle(t *testing.T) {
	w, em, exp := &fakeWatcher{owner: true}, &fakeEmitter{}, &fakeExporter{}
	calls := 0
	item := attach(w, em, exp, testService, Callbacks{Toggle: func() { calls++ }})

	if derr := item.Activate(10, 20); derr != nil {
		t.Fatalf("Activate error = %v, want nil", derr)
	}
	if calls != 1 {
		t.Errorf("Activate fired the toggle %d times, want exactly one", calls)
	}
}

// TestItemActivateNilToggleWarnsOnce pins the nil-callback degradation:
// without a toggle the Activate is a no-op with exactly one WARN per item
// lifetime (the emitWarned discipline pattern).
func TestItemActivateNilToggleWarnsOnce(t *testing.T) {
	buf := captureLogs(t)
	item, _, _, _ := attachOK()

	if derr := item.Activate(0, 0); derr != nil {
		t.Fatalf("nil-toggle Activate error = %v, want nil", derr)
	}
	if derr := item.Activate(1, 1); derr != nil {
		t.Fatalf("second nil-toggle Activate error = %v, want nil", derr)
	}
	if got := strings.Count(buf.String(), `"level":"WARN"`); got != 1 {
		t.Errorf("nil-toggle Activate warned %d times, want exactly one; log:\n%s", got, buf.String())
	}
}
