// Package hotkey implements the Right Shift tap state machine: the
// classic-with-waiting scheme with a 300 ms disambiguation window whose
// decisions fire only when the window expires after the last tap.
package hotkey

import "time"

// DefaultWindow is the tap disambiguation window (D-05): after every tap
// the machine waits this long before deciding; YAML-configurable from
// Phase 3.
const DefaultWindow = 300 * time.Millisecond

// KeyvalShiftR is the IBus keyval of the right Shift key — the documented
// default tap key of NewFSM (/usr/include/ibus-1.0/ibuskeysyms.h — verified
// verbatim).
const KeyvalShiftR = 0xffe2

// MaxTaps caps a series: a fourth tap inside the window stays at three so
// Triple fires exactly once (D-04).
const MaxTaps = 3

// Action is a tap-series decision, fired at window expiry: Single switches
// the layout, Double corrects the last word, Triple the last phrase.
type Action int

// Tap-series decisions.
const (
	Single Action = iota + 1
	Double
	Triple
)

// Event kinds fed into the FSM by the adapter.
type (
	// KeyPress is a press of the key with the given keyval.
	KeyPress struct{ Keyval uint32 }
	// KeyRelease is a release of the key with the given keyval.
	KeyRelease struct{ Keyval uint32 }
	// TimerExpired re-enters the machine when the adapter's time.AfterFunc
	// for the armed window fires.
	TimerExpired struct{}
	// Reset disarms everything; the adapter feeds it on FocusOut/Reset.
	Reset struct{}
)

// Event is the sum of the FSM input kinds.
type Event = any

// FSM is the Right Shift tap state machine: pure and deterministic — no
// goroutines, no real clock, no channels. The adapter feeds key events with
// injected timestamps and re-enters TimerExpired when its window timer
// fires. The series key is configurable through the constructor (D-35:
// mechanics only — the corpus of fsm_test.go is the executed ADR-002 and
// its expectations are untouchable).
type FSM struct {
	window    time.Duration
	tapKeyval uint32
	taps      int
	held      bool
	sawKey    bool
	lastTap   time.Duration
}

// NewFSM returns an FSM deciding tap series of the given key inside the
// given disambiguation window.
func NewFSM(window time.Duration, tapKeyval uint32) *FSM {
	return &FSM{window: window, tapKeyval: tapKeyval}
}

// Feed advances the machine by one event at the given injected time and
// returns the decision — exactly one Action — when a window expires over a
// live series; every other event returns no actions.
func (f *FSM) Feed(ev Event, now time.Duration) []Action {
	switch e := ev.(type) {
	case KeyPress:
		f.keyPress(e)
	case KeyRelease:
		f.keyRelease(e, now)
	case TimerExpired:
		return f.expired(now)
	case Reset:
		f.disarm()
	}

	return nil
}

// keyPress arms a hold of the configured tap key or, for any other key,
// either marks modifier use (while held) or silently cancels a pending
// series.
func (f *FSM) keyPress(e KeyPress) {
	if e.Keyval == f.tapKeyval {
		f.held = true
		f.sawKey = false

		return
	}
	if f.held {
		f.sawKey = true

		return
	}
	f.taps = 0
}

// keyRelease routes tap-key releases through the tap logic; any other
// release between taps silently cancels a pending series.
func (f *FSM) keyRelease(e KeyRelease, now time.Duration) {
	if e.Keyval == f.tapKeyval {
		f.releaseShift(now)

		return
	}
	if !f.held {
		f.taps = 0
	}
}

// releaseShift settles a Shift_R hold: a modifier use is not a tap and kills
// the pending series; a clean tap counts (capped at MaxTaps) and re-arms the
// window from now.
func (f *FSM) releaseShift(now time.Duration) {
	if !f.held {
		return // release without a matching press — ignore
	}
	f.held = false
	if f.sawKey {
		f.sawKey = false
		f.taps = 0

		return
	}
	if f.taps > 0 && now-f.lastTap > f.window {
		f.taps = 0 // the gap broke the series: a stale count never fires
	}
	if f.taps < MaxTaps {
		f.taps++
	}
	f.lastTap = now
	// The adapter arms time.AfterFunc(window) and re-enters TimerExpired.
}

// expired decides at window end: a live series whose last tap is at least
// one window old fires its action exactly once; a stale timer (an older
// window's AfterFunc racing a later tap) is a no-op.
func (f *FSM) expired(now time.Duration) []Action {
	if f.taps == 0 || now-f.lastTap < f.window {
		return nil
	}
	action := Action(f.taps)
	f.taps = 0

	return []Action{action}
}

// disarm clears every bit of a pending series, including a held Shift.
func (f *FSM) disarm() {
	f.taps = 0
	f.held = false
	f.sawKey = false
}
