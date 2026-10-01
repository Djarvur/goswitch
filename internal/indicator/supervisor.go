package indicator

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

// The supervisor's wire and pacing constants: the bus signal the event path
// reacts to, the health-check period (a compile-time constant — no config
// keys, ever) and the signal buffer that keeps the connection's delivery
// goroutines from piling up between beats.
const (
	// dbusSignalIface and signalNameOwnerChanged name the broadcast the
	// event path rides: the watcher appearing on the bus (the 2026-10-01
	// boot race) is a NameOwnerChanged for its well-known name.
	dbusSignalIface        = "org.freedesktop.DBus"
	signalNameOwnerChanged = "NameOwnerChanged"

	// healthInterval is the poll path's beat: a silent eviction self-heals
	// within one interval.
	healthInterval = 30 * time.Second

	// supervisorSignalBuffer backs the NameOwnerChanged channel — the
	// watcher-name broadcast is rare, the buffer only covers a burst
	// between two run-loop iterations.
	supervisorSignalBuffer = 4

	// ownerChangedBodyParts is the NameOwnerChanged body arity: name, old
	// owner, new owner — the defensive guard's shape check.
	ownerChangedBodyParts = 3
)

// The loss verdicts the reason field of the loss WARN names.
const (
	reasonProbeFailed = "watcher probe failed"
	reasonWatcherGone = "watcher gone"
	reasonItemsFailed = "registered-items read failed"
	reasonItemEvicted = "item missing from the watcher's registry"
)

// supervisor owns the item's watcher lifecycle (quick plan 261001-fg3): the
// v0 one-shot attach became a re-attachable one — the event path reacts to
// the watcher appearing (NameOwnerChanged), the poll path re-checks the
// registration every health beat. Every field except the channels is fixed
// at construction (attach); the channels are synthetic in the corpus and
// live in Supervise. run's select loop is the only goroutine touching lost.
type supervisor struct {
	item    *Item
	w       Watcher
	exp     exporter
	service string
	signals <-chan *dbus.Signal
	ticks   <-chan time.Time
	lost    bool // the healthy→lost transition gates the one-WARN discipline
}

// run drives the supervisor until ctx is done — the goroutine's exit
// condition (the daemon passes its signal context, so the supervisor dies
// with the daemon). Ticks drive the health check, signals the owner-changed
// re-attach; both paths share the single-writer loop, so lost needs no lock.
func (s *supervisor) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.ticks:
			s.check()
		case sig, ok := <-s.signals:
			if !ok {
				s.signals = nil // closed: stop reacting, keep the health beats

				continue
			}
			s.onWatcherSignal(sig)
		}
	}
}

// check is the poll path's beat: probe the watcher, read its registry, and
// re-attach when our service is missing from the list — the silent-eviction
// verdict. A present item is the silent steady state; a stale lost flag
// clears silently (we did not revive it, we merely observe health).
func (s *supervisor) check() {
	has, err := s.w.NameHasOwner()
	if err != nil {
		s.markLost(reasonProbeFailed, err)

		return
	}
	if !has {
		s.markLost(reasonWatcherGone, nil)

		return
	}
	items, err := s.w.RegisteredStatusNotifierItems()
	if err != nil {
		s.markLost(reasonItemsFailed, err)

		return
	}
	if !itemsContainService(items, s.service) {
		s.markLost(reasonItemEvicted, nil)
		s.tryReattach()

		return
	}
	if s.lost {
		s.lost = false
	}
}

// onWatcherSignal is the event path: react only to
// org.freedesktop.DBus.NameOwnerChanged for the watcher name with a
// NON-EMPTY new owner. The live AddMatch filter does the bulk of this
// filtering; the re-check is the defensive corpus-pinned guard (a replayed
// or foreign signal must not drive a registration).
func (s *supervisor) onWatcherSignal(sig *dbus.Signal) {
	if sig == nil || sig.Name != dbusSignalIface+"."+signalNameOwnerChanged {
		return
	}
	if len(sig.Body) != ownerChangedBodyParts {
		return
	}
	name, _ := sig.Body[0].(string)
	newOwner, _ := sig.Body[2].(string)
	if name != watcherName || newOwner == "" {
		return
	}

	s.tryReattach()
}

// tryReattach re-runs the FULL registration sequence (probe → exports →
// register; the re-export of the live objects is idempotent replacement in
// godbus). Success revives the item — and logs exactly one INFO when a loss
// was on record. Failure is silent: the retry happens on the NEXT event or
// beat, never inline (T-FG3-02 — no tight loops).
func (s *supervisor) tryReattach() {
	if err := registerItem(s.w, s.exp, s.item, s.service); err != nil {
		return
	}
	wasLost := s.lost
	s.lost = false
	s.item.revive()
	if wasLost {
		slog.Info("tray indicator re-registered", "component", "tray indicator")
	}
}

// markLost records the healthy→lost transition: exactly one WARN with the
// component token and the verdict in the reason field, and the item drops
// to the unregistered state (ModeChanged stops emitting until a revival).
// Subsequent failed beats find lost already set — silent.
func (s *supervisor) markLost(reason string, err error) {
	if s.lost {
		return
	}
	s.lost = true
	s.item.markUnregistered()
	if err != nil {
		slog.Warn("tray indicator lost", "component", "tray indicator", "reason", reason, "error", err)

		return
	}
	slog.Warn("tray indicator lost", "component", "tray indicator", "reason", reason)
}

// itemsContainService reports whether the watcher's registry carries our
// registration. The ubuntu-appindicators watcher lists every item as
// <bus name><object path> ("org.djarvur.goswitch/StatusNotifierItem"), so
// the check accepts the bare service name or the service-prefixed path.
func itemsContainService(items []string, service string) bool {
	return slices.ContainsFunc(items, func(entry string) bool {
		return entry == service || strings.HasPrefix(entry, service+"/")
	})
}

// Supervise runs the supervisor's live wiring on the daemon's EXISTING ctl
// connection and BLOCKS until ctx is done — the daemon calls it as
//
//	go item.Supervise(ctx, conn)
//
// with the Run context received through OnConn, so the supervisor rides the
// daemon's signal context and dies with it. The AddMatch failure degrades
// to the poll path with one WARN (the event path is an optimization; the
// ~30s beat self-heals whatever the event path would have caught).
func (it *Item) Supervise(ctx context.Context, conn *dbus.Conn) {
	if it.sup == nil {
		return
	}
	match := []dbus.MatchOption{
		dbus.WithMatchInterface(dbusSignalIface),
		dbus.WithMatchMember(signalNameOwnerChanged),
		dbus.WithMatchArg(0, watcherName),
	}
	if err := conn.AddMatchSignal(match...); err != nil {
		slog.Warn("tray indicator owner-changed match failed",
			"component", "tray indicator", "error", err)
		match = nil // no match, no best-effort removal at teardown
	}
	signals := make(chan *dbus.Signal, supervisorSignalBuffer)
	conn.Signal(signals)
	ticker := time.NewTicker(healthInterval)
	defer func() {
		ticker.Stop()
		conn.RemoveSignal(signals)
		if match != nil {
			_ = conn.RemoveMatchSignal(match...) // best-effort teardown
		}
	}()

	it.sup.signals = signals
	it.sup.ticks = ticker.C
	it.sup.run(ctx)
}
