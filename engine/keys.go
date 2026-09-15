package engine

// Modifier masks decoded from the IBus key-event state word
// (/usr/include/ibus-1.0/ibustypes.h:70-77,97 — values verified verbatim).
const (
	MaskShift   = 1 << 0
	MaskControl = 1 << 2
	MaskMod1    = 1 << 3 // Alt.
	MaskMod4    = 1 << 6 // Super.
	MaskMod5    = 1 << 7 // Level3/AltGr.
	MaskRelease = 1 << 30
)

// Client capability bits reported through SetCapabilities
// (ibustypes.h:119-126).
const (
	CapPreeditText     = 1 << 0
	CapSurroundingText = 1 << 5
	CapSyncProcessKey  = 1 << 7
)

// Keyvals used by the daemon (/usr/include/ibus-1.0/ibuskeysyms.h — values
// verified verbatim).
const (
	KeyBackSpace = 0xff08
	KeyTab       = 0xff09
	KeyReturn    = 0xff0d
	KeyEscape    = 0xff1b
	// KeyKPEnter is the keypad variant of Enter (XK_KP_Enter) — a CORR-09
	// reset trigger in its own right: the numpad Enter must clear the buffer
	// exactly like the main-row one (plan 02-05).
	KeyKPEnter  = 0xff8b
	KeyShiftL   = 0xffe1
	KeyShiftR   = 0xffe2
	KeyControlL = 0xffe3
	// KeyControlR is the right Ctrl key — the default combo key of the
	// word-layout gesture (SWCH-02/D-36; ibuskeysyms.h:191
	// "#define IBUS_KEY_Control_R 0xffe4", verified verbatim — never from
	// memory, the 02-05 class trap). The resolved binding itself lives in
	// the pure hotkey package as KeyvalCtrlR.
	KeyControlR = 0xffe4
	// KeySuperL/KeySuperR are the Super/Windows modifier keysyms
	// (ibuskeysyms.h "#define IBUS_KEY_Super_L 0xffeb" / Super_R 0xffec,
	// verified verbatim) — the press/release pair the MACR layer tracks for
	// its consumed-upstream detect (ADR-005 b.1/b.2).
	KeySuperL = 0xffeb
	KeySuperR = 0xffec
	KeySpace  = 0x020
	// KeyV is the Latin-1 'v' — the paste half of the clipboard rung's
	// Ctrl+V forward burst (plan 03-03, D-28).
	KeyV = 0x076
)

// modsMask isolates the low modifier byte of the IBus state word; the
// release bit lives in bit 30 and the rest is reserved.
const modsMask = 0xff
