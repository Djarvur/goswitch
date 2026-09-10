package hotkey_test

import (
	"testing"
	"time"

	"github.com/Djarvur/goswitch/internal/hotkey"
)

// Gaps around the 300 ms window edge (D-05) and misc feed offsets, named so
// every corpus case states its timing contract.
const (
	gapJustInside  = 299 * time.Millisecond
	gapAtEdge      = 300 * time.Millisecond
	gapJustOutside = 301 * time.Millisecond

	glitchAt    = 1 * time.Millisecond
	letterAt    = 50 * time.Millisecond
	secondTapAt = 100 * time.Millisecond
)

// tap feeds a full Shift_R tap (press and release) at time t.
func tap(f *hotkey.FSM, t time.Duration) []hotkey.Action {
	f.Feed(hotkey.KeyPress{Keyval: hotkey.KeyvalShiftR}, t)

	return f.Feed(hotkey.KeyRelease{Keyval: hotkey.KeyvalShiftR}, t)
}

// pressLetter feeds a press of a non-shift key at time t.
func pressLetter(f *hotkey.FSM, t time.Duration) []hotkey.Action {
	return f.Feed(hotkey.KeyPress{Keyval: 0x61}, t) // 'a'
}

// actionsEqual reports whether got matches want exactly.
func actionsEqual(got, want []hotkey.Action) bool {
	if len(got) != len(want) {
		return false
	}
	for i, a := range got {
		if a != want[i] {
			return false
		}
	}

	return true
}

// TestFSM_SingleAtWindowExpiry pins the classic-with-waiting scheme (D-04):
// nothing fires on the tap itself — press and release return no action — and
// the single decision arrives only when the window expires after it.
func TestFSM_SingleAtWindowExpiry(t *testing.T) {
	t.Parallel()

	f := hotkey.NewFSM(hotkey.DefaultWindow)
	if acts := tap(f, 0); len(acts) != 0 {
		t.Fatalf("tap at 0 returned %v, want no action until the window expires", acts)
	}
	acts := f.Feed(hotkey.TimerExpired{}, hotkey.DefaultWindow)
	if !actionsEqual(acts, []hotkey.Action{hotkey.Single}) {
		t.Fatalf("TimerExpired at window end = %v, want [Single]", acts)
	}
	if acts := f.Feed(hotkey.TimerExpired{}, 2*hotkey.DefaultWindow); len(acts) != 0 {
		t.Errorf("second TimerExpired = %v, want none: the series is consumed", acts)
	}
}

// TestFSM_DoubleAndTriple pins the tap counts: two taps inside the window
// decide Double, three decide Triple — always at window expiry.
func TestFSM_DoubleAndTriple(t *testing.T) {
	t.Parallel()

	t.Run("two taps gap 299ms", func(t *testing.T) {
		t.Parallel()
		f := hotkey.NewFSM(hotkey.DefaultWindow)
		tap(f, 0)
		tap(f, gapJustInside)
		acts := f.Feed(hotkey.TimerExpired{}, gapJustInside+hotkey.DefaultWindow)
		if !actionsEqual(acts, []hotkey.Action{hotkey.Double}) {
			t.Errorf("TimerExpired = %v, want [Double]", acts)
		}
	})

	t.Run("two taps gap 300ms", func(t *testing.T) {
		t.Parallel()
		f := hotkey.NewFSM(hotkey.DefaultWindow)
		tap(f, 0)
		tap(f, gapAtEdge)
		acts := f.Feed(hotkey.TimerExpired{}, gapAtEdge+hotkey.DefaultWindow)
		if !actionsEqual(acts, []hotkey.Action{hotkey.Double}) {
			t.Errorf("TimerExpired = %v, want [Double]", acts)
		}
	})

	t.Run("three taps", func(t *testing.T) {
		t.Parallel()
		f := hotkey.NewFSM(hotkey.DefaultWindow)
		tap(f, 0)
		tap(f, letterAt)
		tap(f, 2*letterAt)
		acts := f.Feed(hotkey.TimerExpired{}, 2*letterAt+hotkey.DefaultWindow)
		if !actionsEqual(acts, []hotkey.Action{hotkey.Triple}) {
			t.Errorf("TimerExpired = %v, want [Triple]", acts)
		}
	})
}

// TestFSM_WindowEdges pins the window boundary: a gap of 299 or 300 ms
// continues the series, a gap of 301 ms starts a new one — the first tap
// does not continue into the second.
func TestFSM_WindowEdges(t *testing.T) {
	t.Parallel()

	t.Run("gap 301ms starts a new series", func(t *testing.T) {
		t.Parallel()
		f := hotkey.NewFSM(hotkey.DefaultWindow)
		tap(f, 0)
		tap(f, gapJustOutside)
		acts := f.Feed(hotkey.TimerExpired{}, gapJustOutside+hotkey.DefaultWindow)
		if !actionsEqual(acts, []hotkey.Action{hotkey.Single}) {
			t.Errorf("TimerExpired = %v, want [Single]: the 301 ms gap broke the series", acts)
		}
	})

	t.Run("gap 299ms continues the series", func(t *testing.T) {
		t.Parallel()
		f := hotkey.NewFSM(hotkey.DefaultWindow)
		tap(f, 0)
		tap(f, gapJustInside)
		acts := f.Feed(hotkey.TimerExpired{}, gapJustInside+hotkey.DefaultWindow)
		if !actionsEqual(acts, []hotkey.Action{hotkey.Double}) {
			t.Errorf("TimerExpired = %v, want [Double]: the 299 ms gap continued the series", acts)
		}
	})
}

// TestFSM_ModifierUse pins the prototype's rshift_saw_key rule: Shift_R held
// through another key was used as a modifier — its release is not a tap and
// no series forms.
func TestFSM_ModifierUse(t *testing.T) {
	t.Parallel()

	f := hotkey.NewFSM(hotkey.DefaultWindow)
	f.Feed(hotkey.KeyPress{Keyval: hotkey.KeyvalShiftR}, 0)
	if acts := pressLetter(f, letterAt); len(acts) != 0 {
		t.Fatalf("letter press returned %v, want no action", acts)
	}
	if acts := f.Feed(hotkey.KeyRelease{Keyval: hotkey.KeyvalShiftR}, 2*letterAt); len(acts) != 0 {
		t.Fatalf("modifier-use release returned %v, want no action", acts)
	}
	if acts := f.Feed(hotkey.TimerExpired{}, 3*hotkey.DefaultWindow); len(acts) != 0 {
		t.Errorf("TimerExpired = %v, want none: a modifier use never forms a series", acts)
	}
}

// TestFSM_InterveningKeyCancels pins the silent cancel: a non-shift key
// between two taps kills the pending series — at the canceled series' window
// expiry nothing fires, and the later tap decides alone as a single.
func TestFSM_InterveningKeyCancels(t *testing.T) {
	t.Parallel()

	f := hotkey.NewFSM(hotkey.DefaultWindow)
	tap(f, 0)
	pressLetter(f, letterAt)
	tap(f, 2*secondTapAt)
	// The first series' window (armed at t=0) expired at 300 ms: silent.
	if acts := f.Feed(hotkey.TimerExpired{}, hotkey.DefaultWindow); len(acts) != 0 {
		t.Fatalf("TimerExpired at the canceled series' window = %v, want none", acts)
	}
	acts := f.Feed(hotkey.TimerExpired{}, 2*secondTapAt+hotkey.DefaultWindow)
	if !actionsEqual(acts, []hotkey.Action{hotkey.Single}) {
		t.Errorf("TimerExpired = %v, want [Single]: the second tap starts a fresh series", acts)
	}
}

// TestFSM_FourthTapStaysAtThree pins the cap: a fourth tap within the window
// does not wrap or stack — exactly one Triple fires at expiry.
func TestFSM_FourthTapStaysAtThree(t *testing.T) {
	t.Parallel()

	f := hotkey.NewFSM(hotkey.DefaultWindow)
	tap(f, 0)
	tap(f, letterAt)
	tap(f, 2*letterAt)
	tap(f, 4*letterAt)
	lastTap := 4 * letterAt
	acts := f.Feed(hotkey.TimerExpired{}, lastTap+hotkey.DefaultWindow)
	if !actionsEqual(acts, []hotkey.Action{hotkey.Triple}) {
		t.Fatalf("TimerExpired after 4 taps = %v, want [Triple] exactly once", acts)
	}
	if acts := f.Feed(hotkey.TimerExpired{}, lastTap+2*hotkey.DefaultWindow); len(acts) != 0 {
		t.Errorf("second TimerExpired = %v, want none: the decision fired exactly once", acts)
	}
}

// TestFSM_FocusOutDisarms pins the Reset event (the adapter feeds it on
// FocusOut/Reset from IBus): inside a series it disarms the window, and it
// also clears a held Shift so its later release is not a tap.
func TestFSM_FocusOutDisarms(t *testing.T) {
	t.Parallel()

	t.Run("reset inside series", func(t *testing.T) {
		t.Parallel()
		f := hotkey.NewFSM(hotkey.DefaultWindow)
		tap(f, 0)
		if acts := f.Feed(hotkey.Reset{}, letterAt); len(acts) != 0 {
			t.Fatalf("Reset returned %v, want no action", acts)
		}
		if acts := f.Feed(hotkey.TimerExpired{}, 2*hotkey.DefaultWindow); len(acts) != 0 {
			t.Errorf("TimerExpired after Reset = %v, want none: the window is disarmed", acts)
		}
	})

	t.Run("press held across reset", func(t *testing.T) {
		t.Parallel()
		f := hotkey.NewFSM(hotkey.DefaultWindow)
		f.Feed(hotkey.KeyPress{Keyval: hotkey.KeyvalShiftR}, 0)
		f.Feed(hotkey.Reset{}, letterAt)
		if acts := f.Feed(hotkey.KeyRelease{Keyval: hotkey.KeyvalShiftR}, 2*letterAt); len(acts) != 0 {
			t.Fatalf("release after Reset returned %v, want no action", acts)
		}
		if acts := f.Feed(hotkey.TimerExpired{}, 2*hotkey.DefaultWindow); len(acts) != 0 {
			t.Errorf("TimerExpired = %v, want none: the held press died with the Reset", acts)
		}
	})
}

// TestFSM_ReleaseWithoutPress pins that a Shift_R release with no matching
// press is ignored entirely.
func TestFSM_ReleaseWithoutPress(t *testing.T) {
	t.Parallel()

	f := hotkey.NewFSM(hotkey.DefaultWindow)
	if acts := f.Feed(hotkey.KeyRelease{Keyval: hotkey.KeyvalShiftR}, 0); len(acts) != 0 {
		t.Fatalf("release without press returned %v, want no action", acts)
	}
	if acts := f.Feed(hotkey.TimerExpired{}, hotkey.DefaultWindow); len(acts) != 0 {
		t.Errorf("TimerExpired = %v, want none: the stray release formed no series", acts)
	}
	if acts := tap(f, letterAt); len(acts) != 0 {
		t.Fatalf("clean tap after stray release returned %v, want no action yet", acts)
	}
	acts := f.Feed(hotkey.TimerExpired{}, letterAt+hotkey.DefaultWindow)
	if !actionsEqual(acts, []hotkey.Action{hotkey.Single}) {
		t.Errorf("TimerExpired = %v, want [Single]: the FSM still works after the stray release", acts)
	}
}

// TestFSM_ShiftGlitch pins the ibus#2600 shape — Shift_R down, a letter 1 ms
// later, Shift_R up — as modifier use with no action, after which the FSM
// keeps deciding normally.
func TestFSM_ShiftGlitch(t *testing.T) {
	t.Parallel()

	f := hotkey.NewFSM(hotkey.DefaultWindow)
	f.Feed(hotkey.KeyPress{Keyval: hotkey.KeyvalShiftR}, 0)
	pressLetter(f, glitchAt)
	if acts := f.Feed(hotkey.KeyRelease{Keyval: hotkey.KeyvalShiftR}, 2*glitchAt); len(acts) != 0 {
		t.Fatalf("glitch release returned %v, want no action", acts)
	}
	if acts := f.Feed(hotkey.TimerExpired{}, hotkey.DefaultWindow+2*glitchAt); len(acts) != 0 {
		t.Errorf("TimerExpired = %v, want none: a glitch is modifier use, not a tap", acts)
	}
	if acts := tap(f, letterAt); len(acts) != 0 {
		t.Fatalf("clean tap after the glitch returned %v, want no action yet", acts)
	}
	acts := f.Feed(hotkey.TimerExpired{}, letterAt+hotkey.DefaultWindow)
	if !actionsEqual(acts, []hotkey.Action{hotkey.Single}) {
		t.Errorf("TimerExpired = %v, want [Single]: the FSM recovers after a glitch", acts)
	}
}

// TestFSM_BurstOfTen pins the machine against cycling: ten taps inside the
// window still decide Triple once, never a wrap-around or a repeated fire.
func TestFSM_BurstOfTen(t *testing.T) {
	t.Parallel()

	f := hotkey.NewFSM(hotkey.DefaultWindow)
	step := 30 * time.Millisecond
	var lastTap time.Duration
	for i := range 10 {
		lastTap = time.Duration(i) * step
		if acts := tap(f, lastTap); len(acts) != 0 {
			t.Fatalf("tap %d returned %v, want no action before expiry", i+1, acts)
		}
	}
	acts := f.Feed(hotkey.TimerExpired{}, lastTap+hotkey.DefaultWindow)
	if !actionsEqual(acts, []hotkey.Action{hotkey.Triple}) {
		t.Fatalf("TimerExpired after 10 taps = %v, want [Triple]", acts)
	}
	if acts := f.Feed(hotkey.TimerExpired{}, lastTap+2*hotkey.DefaultWindow); len(acts) != 0 {
		t.Errorf("second TimerExpired = %v, want none: no cycling", acts)
	}
}

// TestFSM_DefaultWindow pins the D-05 default: 300 ms.
func TestFSM_DefaultWindow(t *testing.T) {
	t.Parallel()

	if hotkey.DefaultWindow != 300*time.Millisecond {
		t.Errorf("DefaultWindow = %v, want 300ms", hotkey.DefaultWindow)
	}
}
