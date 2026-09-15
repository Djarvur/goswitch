package hotkey

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

// ParseBinding resolves a binding name against the closed name tables:
// "+"-joined tokens with the KEY as the last token, the rest modifiers.
// An unknown token anywhere rejects the whole binding (D-33).
func ParseBinding(name string) (Binding, error) {
	_ = name // RED stub: the corpus pins the real resolution behavior.

	return Binding{}, nil
}
