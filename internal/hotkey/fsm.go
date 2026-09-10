// Package hotkey implements the Right Shift tap state machine: the
// classic-with-waiting scheme with a 300 ms disambiguation window whose
// decisions fire only when the window expires after the last tap.
package hotkey

import "time"

// DefaultWindow is the tap disambiguation window (D-05): after every tap
// the machine waits this long before deciding; YAML-configurable from
// Phase 3.
const DefaultWindow = 300 * time.Millisecond

// KeyvalShiftR is the IBus keyval of the right Shift key
// (/usr/include/ibus-1.0/ibuskeysyms.h — verified verbatim).
const KeyvalShiftR = 0xffe2

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
// fires.
type FSM struct {
	window time.Duration
}

// NewFSM returns an FSM with the given tap disambiguation window.
func NewFSM(window time.Duration) *FSM {
	return &FSM{window: window}
}

// Feed advances the machine by one event at the given injected time and
// returns the decision — exactly one Action — when a window expires over a
// live series; every other event returns no actions.
func (f *FSM) Feed(_ Event, _ time.Duration) []Action {
	return nil
}
