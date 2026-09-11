package correct

// CapSurroundingText is IBUS_CAP_SURROUNDING_TEXT (1<<5,
// ibustypes.h:119-126), mirrored from engine/keys.go: the pure package must
// not import the D-Bus-bound engine package (Phase 1 dependency direction —
// engine imports pure packages, never the reverse).
const CapSurroundingText = 1 << 5

// Level identifies a step of the ADR-003 replacement capability ladder.
type Level int

// Replacement ladder levels (ADR-003).
const (
	// Level1 deletes the range through DeleteSurroundingText on clients
	// that report surrounding text.
	Level1 Level = iota + 1
	// Level2 replays Backspace key events for clients without the bit.
	Level2
)

// Plan is the replacement geometry of one correction (CORR-07): every count
// is in runes, never bytes (ADR-003 — a Cyrillic rune is 2 UTF-8 bytes, and
// a byte count would erase twice the text).
type Plan struct {
	Level      Level
	Offset     int32
	NChars     uint32
	Backspaces int
	Commit     []rune
}

// BuildPlan computes the replacement geometry for the token, its boundary
// tail and the converted token, choosing the ladder level by the
// surrounding-text capability bit (CORR-07, ADR-003).
func BuildPlan(token, tail, converted []rune, caps uint32) Plan {
	_, _, _ = token, tail, converted
	_ = caps

	return Plan{}
}
