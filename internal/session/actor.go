// Package session hosts the daemon's event consumers: the actor that
// serializes engine events into the hotkey FSM and logs its decisions.
package session

import (
	"log/slog"
	"sync"
	"time"

	"github.com/Djarvur/goswitch/engine"
	"github.com/Djarvur/goswitch/internal/hotkey"
)

// Actor implements engine.EventHandler for the Phase 1 daemon: every key and
// lifecycle event is serialized through one mutex into the pure hotkey FSM,
// and the tap-window deadline is carried by a single time.AfterFunc timer
// that re-enters the FSM at expiry.
//
// godbus dispatches every D-Bus method call on its own goroutine, so the
// mutex is the actor's single entry point: without it the FSM state would
// race between concurrent ProcessKeyEvent calls.
type Actor struct {
	mu     sync.Mutex
	fsm    *hotkey.FSM
	window time.Duration
	start  time.Time
	timer  *time.Timer
}

// NewActor returns an actor deciding tap series inside the given
// disambiguation window (D-05). The daemon passes hotkey.DefaultWindow.
func NewActor(window time.Duration) *Actor {
	return &Actor{
		fsm:    hotkey.NewFSM(window),
		window: window,
		start:  time.Now(),
	}
}

// HandleKey implements engine.EventHandler: the decoded event is fed into
// the FSM under the mutex and a Shift_R release re-arms the deadline timer.
// Decisions never fire here — only at window expiry (D-04).
func (a *Actor) HandleKey(ev engine.EngineEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if ev.Release {
		a.fsm.Feed(hotkey.KeyRelease{Keyval: ev.Keyval}, a.elapsed())
	} else {
		a.fsm.Feed(hotkey.KeyPress{Keyval: ev.Keyval}, a.elapsed())
	}
	if ev.Release && ev.Keyval == hotkey.KeyvalShiftR {
		// Only a Shift_R release can move the series deadline (the FSM
		// counts taps on clean releases); re-arming on any other event
		// would either extend the deadline from the wrong instant or
		// arm a timer over a cancelled series. A stale expiry is a
		// no-op in the FSM, so over-arming is harmless, under-arming
		// would silently drop the decision.
		a.armTimer()
	}
}

// HandleLifecycle implements engine.EventHandler: FocusOut and Reset disarm
// a pending series (the input context is gone — a decision there would be
// garbage), everything else is a DEBUG trace.
func (a *Actor) HandleLifecycle(kind engine.LifecycleKind) {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch kind {
	case engine.LifecycleFocusOut, engine.LifecycleReset:
		a.fsm.Feed(hotkey.Reset{}, a.elapsed())
		if a.timer != nil {
			a.timer.Stop()
			a.timer = nil
		}
	case engine.LifecycleFocusIn, engine.LifecycleEnable, engine.LifecycleDisable:
		slog.Debug("lifecycle", "kind", kind.String())
	}
}

// Expiry is the timer callback: time.AfterFunc(window) re-enters here when
// the disambiguation window closes.
func (a *Actor) Expiry() {
	a.ExpiryAt(a.elapsed())
}

// ExpiryAt feeds the FSM a window expiry at the given logical time and logs
// every decision as {"msg":"action","n":N} — the e2e stand greps this exact
// shape. Tests inject the logical time directly (deterministic expiry); the
// daemon path always goes through Expiry's real clock.
func (a *Actor) ExpiryAt(now time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.timer = nil
	for _, action := range a.fsm.Feed(hotkey.TimerExpired{}, now) {
		slog.Info("action", "n", int(action))
	}
}

// armTimer replaces the deadline timer; the caller holds the mutex.
func (a *Actor) armTimer() {
	if a.timer != nil {
		a.timer.Stop()
	}
	a.timer = time.AfterFunc(a.window, a.Expiry)
}

// elapsed returns the actor's logical clock; the caller holds the mutex
// (start itself is immutable, so Expiry's unlocked read is safe).
func (a *Actor) elapsed() time.Duration {
	return time.Since(a.start)
}
