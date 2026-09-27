package engine

import "github.com/godbus/dbus/v5"

// Wire-struct field order is a contract: godbus marshals exported fields in
// declaration order and ibus-daemon deserializes positionally, so a reorder
// is a silent protocol break. The orders below are ported from the verified
// goibus/libibus field maps in 01-RESEARCH.md §Pattern 1 — do not reorder.

// Component mirrors the serialized IBusComponent wire struct (12 fields).
type Component struct {
	Name          string                  // "IBusComponent".
	Attachments   map[string]dbus.Variant // {}.
	ComponentName string                  // "org.freedesktop.IBus.goswitch".
	Description   string
	Version       string
	License       string // "MIT".
	Author        string
	Homepage      string
	Exec          string // "" — programmatic registration only, never spawned.
	Textdomain    string
	ObservedPaths []dbus.Variant // empty av.
	EngineList    []dbus.Variant // MakeVariant of each EngineDesc value.
}

// EngineDesc mirrors the serialized IBusEngineDesc wire struct (18 fields:
// s, a{sv}, s×8, u, s×7 — Layout at index 9 drives mutter XKB for the
// two-engine experiment, Symbol at index 12 is the GNOME indicator label).
type EngineDesc struct {
	Name          string                  // "IBusEngineDesc".
	Attachments   map[string]dbus.Variant // {}.
	EngineName    string                  // "goswitch-en" / "goswitch-ru".
	LongName      string
	Description   string
	Language      string // "en" / "ru".
	License       string // "MIT".
	Author        string
	Icon          string
	Layout        string // "us" / "ru".
	Rank          uint32
	Hotkeys       string
	Symbol        string // "en" / "ru".
	Setup         string
	LayoutVariant string
	LayoutOption  string
	Version       string
	Textdomain    string
}

// AttrList mirrors the serialized IBusAttrList wire struct; Attributes is
// an av of variant-wrapped serialized IBusAttribute values — ibusattrlist.h
// keeps a GArray of attribute OBJECTS, and the daemon's CommitText parser
// reads the child as "av": an 'au' there trips its format assertion and
// crashes the 1.5.29 daemon (live-verified 2026-09-14, journal SEGV).
type AttrList struct {
	Name        string // "IBusAttrList".
	Attachments map[string]dbus.Variant
	Attributes  []dbus.Variant // empty av for plain commits
}

// IBusText mirrors the serialized IBusText wire struct; AttrList carries a
// variant-wrapped AttrList value.
type IBusText struct {
	Name        string // "IBusText".
	Attachments map[string]dbus.Variant
	Text        string
	AttrList    dbus.Variant
}

// NewIBusText builds a plain-text IBusText payload with an empty attribute
// list — the shape CommitText expects for unstyled commits.
func NewIBusText(text string) IBusText {
	return IBusText{
		Name:        "IBusText",
		Attachments: map[string]dbus.Variant{},
		Text:        text,
		AttrList: dbus.MakeVariant(AttrList{
			Name:        "IBusAttrList",
			Attachments: map[string]dbus.Variant{},
			Attributes:  []dbus.Variant{},
		}),
	}
}

// NewComponent builds the goswitch IBusComponent wire payload with the
// given engine descriptions.
func NewComponent(engines []EngineDesc) Component {
	list := make([]dbus.Variant, 0, len(engines))
	for i := range engines {
		list = append(list, dbus.MakeVariant(engines[i]))
	}

	return Component{
		Name:          "IBusComponent",
		Attachments:   map[string]dbus.Variant{},
		ComponentName: "org.freedesktop.IBus.goswitch",
		Description:   "goswitch layout engine",
		Version:       "0.1.0",
		License:       "MIT",
		Author:        "Djarvur",
		Homepage:      "https://github.com/Djarvur/goswitch",
		Exec:          "",
		Textdomain:    "",
		ObservedPaths: []dbus.Variant{},
		EngineList:    list,
	}
}

// NewEngineDesc builds one goswitch IBusEngineDesc wire payload.
func NewEngineDesc(name, longName, language, layout, symbol string) EngineDesc {
	return EngineDesc{
		Name:          "IBusEngineDesc",
		Attachments:   map[string]dbus.Variant{},
		EngineName:    name,
		LongName:      longName,
		Description:   "goswitch layout engine",
		Language:      language,
		License:       "MIT",
		Author:        "Djarvur",
		Icon:          "",
		Layout:        layout,
		Rank:          0,
		Hotkeys:       "",
		Symbol:        symbol,
		Setup:         "",
		LayoutVariant: "",
		LayoutOption:  "",
		Version:       "0.1.0",
		Textdomain:    "",
	}
}
