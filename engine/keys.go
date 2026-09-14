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
	KeySpace    = 0x020
)

// modsMask isolates the low modifier byte of the IBus state word; the
// release bit lives in bit 30 and the rest is reserved.
const modsMask = 0xff
