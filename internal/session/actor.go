// Package session hosts the daemon's event consumers: the actor that
// serializes engine events into the hotkey FSM and logs its decisions.
//
// This is the RED-phase stub (D-07): no FSM, no timer, no decision logging —
// the corpus in actor_test.go fails on its assertions against it.
package session

import (
	"time"

	"github.com/Djarvur/goswitch/engine"
)

// Actor is the stub the failing corpus is written against.
type Actor struct{}

// NewActor returns a stub actor; window is not used yet.
func NewActor(window time.Duration) *Actor {
	return &Actor{}
}

// HandleKey is a no-op stub.
func (a *Actor) HandleKey(ev engine.EngineEvent) {}

// HandleLifecycle is a no-op stub.
func (a *Actor) HandleLifecycle(kind engine.LifecycleKind) {}

// Expiry is a no-op stub.
func (a *Actor) Expiry() {}

// ExpiryAt is a no-op stub.
func (a *Actor) ExpiryAt(now time.Duration) {}
