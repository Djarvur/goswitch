package hotkey

import (
	"errors"
	"fmt"
	"strings"
)

// ParseBinding refusals — package sentinels (the err113 discipline of the
// strict lint: no dynamically created error values), wrapped with the
// offending token and binding.
var (
	errEmptyBinding = errors.New("binding name is empty")
	errUnknownToken = errors.New("not a name of the closed binding tables")
)

// Binding is a resolved hotkey binding: the keyval of the bound key plus
// the full modifier state expected at its key event (each "+"-joined token
// contributes its family bit — a Shift_R press carries MaskShift on the
// wire, a held Control_R adds MaskControl).
type Binding struct {
	Keyval  uint32
	ModMask uint32
}

// Keyvals of the bindable keys, verified verbatim against
// /usr/include/ibus-1.0/ibuskeysyms.h:188-199 — values are never written
// from memory (the 02-05 class trap: a name is not a key). Duplicated into
// this pure package instead of importing the D-Bus-bound engine — the
// precedent is KeyvalShiftR (fsm.go); the dependency direction stays
// engine → hotkey.
const (
	KeyvalShiftL = 0xffe1 // Shift_L, ibuskeysyms.h:188.
	KeyvalCtrlL  = 0xffe3 // Control_L, ibuskeysyms.h:190.
	KeyvalCtrlR  = 0xffe4 // Control_R, ibuskeysyms.h:191.
	KeyvalAltL   = 0xffe9 // Alt_L, ibuskeysyms.h:196.
	KeyvalAltR   = 0xffea // Alt_R, ibuskeysyms.h:197.
	KeyvalSuperL = 0xffeb // Super_L, ibuskeysyms.h:198.
	KeyvalSuperR = 0xffec // Super_R, ibuskeysyms.h:199.
)

// Modifier masks of the binding grammar, duplicated from engine/keys.go:5-12
// (values verbatim from /usr/include/ibus-1.0/ibustypes.h:70-77,97) — hotkey
// is a pure package and must not import the D-Bus-bound engine (the same
// duplication precedent as KeyvalShiftR).
const (
	MaskShift   = 1 << 0
	MaskControl = 1 << 2
	MaskMod1    = 1 << 3 // Alt.
	MaskMod4    = 1 << 6 // Super.
)

// keyNameValues is the closed table of bindable KEY names (the last "+"
// token of a binding). A function, not a package-level var — the strict
// lint forbids mutable globals (the matrix.go:37-39 idiom).
func keyNameValues() map[string]uint32 {
	return map[string]uint32{
		"shift_l": KeyvalShiftL,
		"shift_r": KeyvalShiftR,
		"ctrl_l":  KeyvalCtrlL,
		"ctrl_r":  KeyvalCtrlR,
		"alt_l":   KeyvalAltL,
		"alt_r":   KeyvalAltR,
		"super_l": KeyvalSuperL,
		"super_r": KeyvalSuperR,
	}
}

// modifierMasks is the closed table of MODIFIER names (every "+" token
// before the key). A function for the same no-mutable-globals reason.
func modifierMasks() map[string]uint32 {
	return map[string]uint32{
		"shift": MaskShift,
		"ctrl":  MaskControl,
		"alt":   MaskMod1,
		"super": MaskMod4,
	}
}

// keyFamilyMasks maps each bindable key to the modifier bit its own press
// carries on the wire (pressing Shift_R sets the Shift bit in the IBus
// state word). A function for the same no-mutable-globals reason.
func keyFamilyMasks() map[string]uint32 {
	return map[string]uint32{
		"shift_l": MaskShift,
		"shift_r": MaskShift,
		"ctrl_l":  MaskControl,
		"ctrl_r":  MaskControl,
		"alt_l":   MaskMod1,
		"alt_r":   MaskMod1,
		"super_l": MaskMod4,
		"super_r": MaskMod4,
	}
}

// ParseBinding resolves a binding name against the closed name tables:
// "+"-joined tokens with the KEY as the last token, the rest modifiers.
// ModMask is the full modifier state of the bound key's event — the held
// modifiers OR the key's own family bit. An unknown or empty token rejects
// the whole binding (D-33: never a silent fallback).
func ParseBinding(name string) (Binding, error) {
	if name == "" {
		return Binding{}, errEmptyBinding
	}

	tokens := strings.Split(name, "+")
	keyName := tokens[len(tokens)-1]

	keyval, ok := keyNameValues()[keyName]
	if !ok {
		return Binding{}, fmt.Errorf("key %q in binding %q: %w", keyName, name, errUnknownToken)
	}

	mask := keyFamilyMasks()[keyName]
	for _, mod := range tokens[:len(tokens)-1] {
		m, ok := modifierMasks()[mod]
		if !ok {
			return Binding{}, fmt.Errorf("modifier %q in binding %q: %w", mod, name, errUnknownToken)
		}
		mask |= m
	}

	return Binding{Keyval: keyval, ModMask: mask}, nil
}
