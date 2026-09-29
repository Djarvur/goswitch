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

// EngineDesc mirrors the serialized IBusEngineDesc wire struct (19 fields:
// s, a{sv}, s×8, u, s×8 — Layout at index 9 drives mutter XKB for the
// two-engine experiment, Symbol at index 12 is the GNOME indicator label,
// and IconPropKey appended LAST (19th, ibusenginedesc.h:335 appended-field
// order) names the engine property whose SYMBOL becomes the panel's dynamic
// icon — the mode-indicator key of owner decision 1, quick plan 260927-way).
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
	IconPropKey   string // modePropKey — the dynamic panel icon property.
}

// NameEN and NameRU are the wire names of the goswitch engine pair — the
// single source the flip path derives its SetGlobalEngine target from
// (D-52). The values are config-level literals (D-20/D-21): a flip journal
// record names the engine, never any user text.
const (
	NameEN = "" // RED stub — GREEN pins "goswitch-en"
	NameRU = "" // RED stub — GREEN pins "goswitch-ru"
)

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

// Property mirrors the serialized IBusProperty wire struct — the panel
// property of the mode indicator (owner decision 1, quick plan 260927-way).
// Field order is the ctor order of ibusproperty.h:159-167 (key, type,
// label, icon, tooltip, sensitive, visible, state, prop_list) with Symbol
// APPENDED LAST — the appended-field convention of every IBus struct
// (verified against the goibus property.go serialization, the
// STACK-sanctioned read-only wire reference). Label/Tooltip/SubProps/Symbol
// are variant-wrapped IBusText/PropList values, the IBusText.AttrList
// precedent.
type Property struct {
	Name        string                  // "IBusProperty".
	Attachments map[string]dbus.Variant // {}.
	Key         string                  // modePropKey — the panel icon key.
	Type        uint32                  // PropTypeNormal (ibusproperty.h:79).
	Label       dbus.Variant            // variant-wrapped IBusText.
	Icon        string                  // "" — the symbol carries the mode.
	Tooltip     dbus.Variant            // variant-wrapped IBusText.
	Sensitive   bool
	Visible     bool
	State       uint32 // PropStateUnchecked (ibusproperty.h:109).
	SubProps    dbus.Variant
	Symbol      dbus.Variant // variant-wrapped IBusText — the panel icon glyph.
}

// PropList mirrors the serialized IBusPropList wire struct — the container
// RegisterProperties carries and Property.SubProps wraps.
type PropList struct {
	Name        string                  // "IBusPropList".
	Attachments map[string]dbus.Variant // {}.
	Properties  []dbus.Variant          // variant-wrapped Property values.
}

// The property constants (ibusproperty.h — values verified verbatim) and
// the mode-indicator identity: the key string is the EngineDesc
// icon_prop_key the GNOME panel resolves to the dynamic icon property, the
// initial symbol is the daemon's start mode (ADR-001: EN at start).
const (
	PropTypeNormal     = 0 // ibusproperty.h:79 "PROP_TYPE_NORMAL = 0".
	PropStateUnchecked = 0 // ibusproperty.h:109 "PROP_STATE_UNCHECKED = 0".

	modePropKey       = "InputMode"
	initialModeSymbol = "en"
)

// NewModeProperty builds the mode-indicator panel property payload: a
// plain normal property (NOT a toggle — PropertyActivate stays a stub,
// nothing may invite clicking) whose Label AND Symbol carry the given
// script symbol, with an empty sub-property list.
func NewModeProperty(symbol string) Property {
	return Property{
		Name:        "IBusProperty",
		Attachments: map[string]dbus.Variant{},
		Key:         modePropKey,
		Type:        PropTypeNormal,
		Label:       dbus.MakeVariant(NewIBusText(symbol)),
		Icon:        "",
		Tooltip:     dbus.MakeVariant(NewIBusText("")),
		Sensitive:   true,
		Visible:     true,
		State:       PropStateUnchecked,
		SubProps: dbus.MakeVariant(PropList{
			Name:        "IBusPropList",
			Attachments: map[string]dbus.Variant{},
			Properties:  []dbus.Variant{},
		}),
		Symbol: dbus.MakeVariant(NewIBusText(symbol)),
	}
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
		IconPropKey:   modePropKey,
	}
}
