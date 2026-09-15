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
	// LevelNone marks a plan that could not be built: the only reachable
	// level was the Backspace replay and the series exceeded the cap
	// (D-27) — the actor refuses with the backspace-cap reason and deletes
	// nothing.
	LevelNone Level = iota
	// Level1 deletes the range through DeleteSurroundingText on clients
	// that report surrounding text.
	Level1
	// Level2 replays Backspace key events for clients without the bit.
	Level2
)

// DefaultBackspaceCap is the default bound of one level-2 Backspace series
// (D-27: a configurable ~50-press ceiling; the YAML wiring arrives in plan
// 03-04). A range longer than the cap never replays — the correction takes
// the silent backspace-cap refusal instead of hammering the key bus.
const DefaultBackspaceCap = 50

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
// token+tail Backspaces and the tail re-committed. The Backspace series is
// bounded by backspaceCap (D-27): a range longer than the cap does not
// build level 2 at all — the plan comes back LevelNone and the actor takes
// the silent refusal. Wiring — plans 02-03 (Level1), 02-05 (Level2) and
// 03-01 (the cap).
func BuildPlan(token, tail, converted []rune, caps uint32, backspaceCap int) Plan {
	n := len(token) + len(tail) // runes, never bytes (ADR-003)
	commit := make([]rune, 0, len(converted)+len(tail))
	commit = append(commit, converted...)
	commit = append(commit, tail...)
	if caps&CapSurroundingText != 0 {
		return Plan{
			Level: Level1,
			// #nosec G115 -- a phrase since the last hard reset is
			// bounded by human typing, far below 2^31 runes; level 1 is
			// one DeleteSurroundingText call, so the D-27 Backspace cap
			// does not bound it (Pitfall 9: the client's surrounding
			// window may still refuse the range — that is the verify
			// step's problem, not the plan's).
			Offset: -int32(n),
			// #nosec G115 -- same phrase-length bound as Offset.
			NChars: uint32(n),
			Commit: commit,
		}
	}
	if n > backspaceCap {
		// D-27: the Backspace series would exceed the cap and the client
		// cannot delete — level 2 is not built at all, and the actor takes
		// the silent backspace-cap refusal instead of a destructive burst.
		return Plan{Level: LevelNone}
	}

	return Plan{Level: Level2, Backspaces: n, Commit: commit}
}
