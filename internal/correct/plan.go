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
// surrounding-text capability bit (CORR-07, ADR-003). GEOMETRY: deleting
// exactly the token while the tail is non-empty is impossible — after the
// deletion the cursor lands AFTER the tail and the commit would insert the
// corrected word behind it ("abc  привет" with the order lost); the D-13
// pin ("ghbdtn " → "привет ") allows only deleting token+tail and
// recommitting converted+tail. BOTH levels therefore replace token+tail:
// Level1 with a single DeleteSurroundingText range, Level2 with
// token+tail Backspaces and the tail re-committed. Wiring — plans 02-03
// (Level1) and 02-05 (Level2).
func BuildPlan(token, tail, converted []rune, caps uint32) Plan {
	n := len(token) + len(tail) // runes, never bytes (ADR-003)
	commit := make([]rune, 0, len(converted)+len(tail))
	commit = append(commit, converted...)
	commit = append(commit, tail...)
	if caps&CapSurroundingText != 0 {
		return Plan{
			Level: Level1,
			// #nosec G115 -- a phrase since the last hard reset is
			// bounded by human typing, far below 2^31 runes.
			Offset: -int32(n),
			// #nosec G115 -- same phrase-length bound as Offset.
			NChars: uint32(n),
			Commit: commit,
		}
	}

	return Plan{Level: Level2, Backspaces: n, Commit: commit}
}
