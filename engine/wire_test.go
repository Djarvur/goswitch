package engine_test

import (
	"reflect"
	"testing"

	"github.com/Djarvur/goswitch/engine"
)

// assertFieldOrder fails the test unless typ's exported field sequence
// equals want exactly.
func assertFieldOrder(t *testing.T, typ reflect.Type, want []string) {
	t.Helper()

	if got := typ.NumField(); got != len(want) {
		t.Errorf("%s has %d fields, want %d — the wire contract changed", typ.Name(), got, len(want))
	}

	for i, name := range want {
		if got := typ.Field(i).Name; got != name {
			t.Errorf("%s field %d = %q, want %q — field order is the wire contract", typ.Name(), i, got, name)
		}
	}
}

// TestWire_FieldOrder pins the serialized field order of the IBus wire
// structs: godbus marshals exported fields in declaration order and
// ibus-daemon deserializes positionally, so any reorder is a silent
// protocol break. Expected sequences per 01-RESEARCH.md §Pattern 1.
func TestWire_FieldOrder(t *testing.T) {
	t.Parallel()

	// header opens every serialized IBus struct: the magic name string and
	// the attachments map.
	header := []string{"Name", "Attachments"}
	withHeader := func(fields ...string) []string {
		return append(append([]string{}, header...), fields...)
	}

	componentWant := withHeader(
		"ComponentName", // "org.freedesktop.IBus.goswitch"
		"Description",
		"Version",
		"License",
		"Author",
		"Homepage",
		"Exec", // "" — programmatic registration only
		"Textdomain",
		"ObservedPaths", // av
		"EngineList",    // av
	)

	engineWant := withHeader(
		"EngineName", // "goswitch-en" / "goswitch-ru"
		"LongName",
		"Description",
		"Language",
		"License",
		"Author",
		"Icon",
		"Layout", // index 9 — drives mutter XKB
		"Rank",   // u
		"Hotkeys",
		"Symbol", // index 12 — GNOME indicator label
		"Setup",
		"LayoutVariant",
		"LayoutOption",
		"Version",
		"Textdomain",
	)

	textWant := withHeader(
		"Text",
		"AttrList",
	)

	t.Run("Component", func(t *testing.T) {
		t.Parallel()

		assertFieldOrder(t, reflect.TypeOf(engine.Component{}), componentWant)
	})

	t.Run("EngineDesc", func(t *testing.T) {
		t.Parallel()

		typ := reflect.TypeOf(engine.EngineDesc{})
		assertFieldOrder(t, typ, engineWant)
		if got := typ.Field(9).Name; got != "Layout" {
			t.Errorf("EngineDesc field 9 = %q, want Layout (drives mutter XKB)", got)
		}
		if got := typ.Field(12).Name; got != "Symbol" {
			t.Errorf("EngineDesc field 12 = %q, want Symbol (GNOME indicator label)", got)
		}
	})

	t.Run("IBusText", func(t *testing.T) {
		t.Parallel()

		assertFieldOrder(t, reflect.TypeOf(engine.IBusText{}), textWant)
	})
}
