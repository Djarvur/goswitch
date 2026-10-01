//nolint:testpackage // drives the unexported supervisor seam — the sanctioned in-package corpus seam
package indicator

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

// errItemsInjected is the registry-read failure the supervisor corpus
// injects (err113).
var errItemsInjected = errors.New("registered-items read exploded")

// watcherItem is the live wire shape of a registry entry: the
// ubuntu-appindicators watcher lists every item as <bus name><object path>.
const watcherItem = testService + "/StatusNotifierItem"

// ownerChangedSignal is the full member name of the bus signal the
// supervisor reacts to.
const ownerChangedSignal = dbusSignalIface + "." + signalNameOwnerChanged

// attachSupervised builds a healthy item plus its supervisor — the corpus
// drives the unexported supervisor with synthetic channels (no real time,
// no live bus).
func attachSupervised() (*Item, *supervisor, *fakeWatcher, *fakeEmitter, *fakeExporter) {
	item, w, em, exp := attachOK()

	return item, item.sup, w, em, exp
}

// TestSupervisorSteadyStateSilent pins the healthy-state discipline: a tick
// that finds our item in the watcher's registry logs NOTHING, re-registers
// nothing and emits nothing — the silent steady state.
func TestSupervisorSteadyStateSilent(t *testing.T) {
	buf := captureLogs(t)
	item, sup, w, em, _ := attachSupervised()
	w.setItems([]string{watcherItem})

	sup.check()
	sup.check() // any number of healthy beats stay silent

	if logged := buf.String(); logged != "" {
		t.Errorf("steady-state beats logged:\n%s", logged)
	}
	if got := len(w.registerCalls()) - 1; got != 0 { // -1: the attach-time registration
		t.Errorf("steady state re-registered %d extra times", got)
	}
	if got := len(em.emitCalls()); got != 0 {
		t.Errorf("steady state emitted %v", got)
	}
	if item == nil {
		t.Fatal("supervised item is nil")
	}
}

// TestSupervisorRecognizesBareServiceItem pins the registry-membership
// grammar: the watcher may list the item under the bare service name (the
// path-suffixed live shape is pinned by the steady-state corpus) — both
// count as present.
func TestSupervisorRecognizesBareServiceItem(t *testing.T) {
	buf := captureLogs(t)
	item, sup, w, em, _ := attachSupervised()
	w.setItems([]string{testService})

	sup.check()

	if logged := buf.String(); logged != "" {
		t.Errorf("present item (bare service form) logged:\n%s", logged)
	}
	if got := len(em.emitCalls()); got != 0 {
		t.Errorf("present item produced %d emits, want none", got)
	}
	if item == nil {
		t.Fatal("supervised item is nil")
	}
}

// TestSupervisorEvictionHealsOnTick pins the silent-eviction self-heal (the
// ~30s poll path): the beat that finds our service missing logs exactly one
// WARN, re-runs the FULL registration (probe → exports → register), revives
// the item with one NewIcon and logs exactly one INFO — and the next
// healthy beat is silent again.
func TestSupervisorEvictionHealsOnTick(t *testing.T) {
	buf := captureLogs(t)
	item, sup, w, em, exp := attachSupervised()
	w.setItems(nil) // the watcher silently dropped us

	sup.check()

	logged := buf.String()
	if got := strings.Count(logged, `"level":"WARN"`); got != 1 {
		t.Errorf("eviction warned %d times, want exactly one; log:\n%s", got, logged)
	}
	if !strings.Contains(logged, `"component":"tray indicator"`) {
		t.Errorf("loss WARN missing the component token; log:\n%s", logged)
	}
	if !strings.Contains(logged, `"msg":"tray indicator re-registered"`) {
		t.Errorf("revival INFO missing; log:\n%s", logged)
	}
	if got := strings.Count(logged, `"level":"INFO"`); got != 1 {
		t.Errorf("revival logged %d INFO records, want exactly one; log:\n%s", got, logged)
	}
	if got := w.registerCalls(); !equalStrings(got, []string{testService, testService}) {
		t.Errorf("registrations = %q, want the attach-time one plus the re-attach", got)
	}
	wantExports := []exportCall{
		{path: menuPath, iface: menuIface},
		{path: menuPath, iface: propertiesIface},
		{path: itemPath, iface: sniIface},
		{path: itemPath, iface: propertiesIface},
		{path: menuPath, iface: menuIface},
		{path: menuPath, iface: propertiesIface},
		{path: itemPath, iface: sniIface},
		{path: itemPath, iface: propertiesIface},
	}
	if got := exp.exportCalls(); !equalExports(got, wantExports) {
		t.Errorf("exports = %v, want the full sequence (menu then item) for attach + re-attach", got)
	}
	if calls := em.emitCalls(); len(calls) != 1 || calls[0].signal != signalNewIcon {
		t.Errorf("revival emits = %v, want exactly one NewIcon", calls)
	}

	// The watcher re-listed us: the next beat is the silent steady state.
	w.setItems([]string{watcherItem})
	sup.check()
	logged2 := buf.String()
	if strings.Count(logged2, `"level":"WARN"`) != 1 || strings.Count(logged2, `"level":"INFO"`) != 1 {
		t.Errorf("post-heal beat logged again; log:\n%s", logged2)
	}
	if item == nil {
		t.Fatal("supervised item is nil")
	}
}

// TestSupervisorWatcherGoneTickDoesNotRegister pins the watcher-gone
// verdict: a beat whose probe says the watcher is absent logs exactly one
// WARN and makes NO registration attempt — the re-attach waits for the
// watcher's return (the tick or the owner-changed event).
func TestSupervisorWatcherGoneTickDoesNotRegister(t *testing.T) {
	buf := captureLogs(t)
	item, sup, w, em, _ := attachSupervised()
	w.setOwner(false)

	sup.check()

	logged := buf.String()
	if got := strings.Count(logged, `"level":"WARN"`); got != 1 {
		t.Errorf("watcher-gone warned %d times, want exactly one; log:\n%s", got, logged)
	}
	if !strings.Contains(logged, `"reason":`) {
		t.Errorf("loss WARN missing the reason field; log:\n%s", logged)
	}
	if got := len(w.registerCalls()) - 1; got != 0 {
		t.Errorf("watcher-gone tick registered %d extra times, want 0", got)
	}
	if got := len(em.emitCalls()); got != 0 {
		t.Errorf("watcher-gone tick emitted %v", got)
	}
	if item == nil {
		t.Fatal("supervised item is nil")
	}
}

// TestSupervisorOwnerChangedReattaches pins the event path: a
// NameOwnerChanged for the watcher name with a NON-EMPTY new owner re-runs
// the full registration — even with the attach-time health intact (the
// fresh watcher knows nothing of us yet).
func TestSupervisorOwnerChangedReattaches(t *testing.T) {
	buf := captureLogs(t)
	item, sup, w, em, _ := attachSupervised()
	w.setItems(nil) // the fresh watcher's registry is empty

	sup.onWatcherSignal(&dbus.Signal{
		Name: ownerChangedSignal,
		Body: []any{watcherName, "", watcherName},
	})

	if got := w.registerCalls(); !equalStrings(got, []string{testService, testService}) {
		t.Errorf("registrations = %q, want the attach-time one plus the event re-attach", got)
	}
	if calls := em.emitCalls(); len(calls) != 1 || calls[0].signal != signalNewIcon {
		t.Errorf("revival emits = %v, want exactly one NewIcon", calls)
	}
	if logged := buf.String(); strings.Contains(logged, `"level":"INFO"`) {
		t.Errorf("event re-attach without a prior loss logged an INFO; log:\n%s", logged)
	}
	if item == nil {
		t.Fatal("supervised item is nil")
	}
}

// TestSupervisorOwnerChangedGuards pins the defensive re-check the live
// match cannot fully replace: foreign signals, foreign names, malformed
// bodies and an empty new owner never trigger a re-attach.
func TestSupervisorOwnerChangedGuards(t *testing.T) {
	buf := captureLogs(t)
	_, sup, w, em, _ := attachSupervised()
	baseline := len(w.registerCalls())

	for name, sig := range map[string]*dbus.Signal{
		"nil signal": nil,
		"foreign member": {
			Name: dbusSignalIface + ".NameAcquired", Body: []any{watcherName, "", watcherName},
		},
		"foreign name": {
			Name: ownerChangedSignal,
			Body: []any{"org.freedesktop.Notifications", "", "org.freedesktop.Notifications"},
		},
		"two-part body":   {Name: ownerChangedSignal, Body: []any{watcherName, ""}},
		"empty new owner": {Name: ownerChangedSignal, Body: []any{watcherName, watcherName, ""}},
		"non-string body": {Name: ownerChangedSignal, Body: []any{watcherName, "", 42}},
	} {
		sup.onWatcherSignal(sig)
		if got := len(w.registerCalls()); got != baseline {
			t.Errorf("%s triggered a re-attach (%d registrations, baseline %d)", name, got, baseline)
		}
	}
	if got := len(em.emitCalls()); got != 0 {
		t.Errorf("guarded signals emitted %v", got)
	}
	if logged := buf.String(); logged != "" {
		t.Errorf("guarded signals logged:\n%s", logged)
	}
}

// TestSupervisorFailedReattachRetriesNextTick pins the retry discipline: a
// failing re-attach is SILENT (the loss WARN stands alone), and the retry
// happens on the NEXT beat — never inline. Across the failed and the healed
// beats the journal carries exactly one WARN and exactly one INFO.
func TestSupervisorFailedReattachRetriesNextTick(t *testing.T) {
	buf := captureLogs(t)
	item, sup, w, _, _ := attachSupervised()
	w.setItems(nil)
	w.registerErr = errWatchInjected

	sup.check() // eviction discovered, re-attach fails silently

	logged := buf.String()
	if got := strings.Count(logged, `"level":"WARN"`); got != 1 {
		t.Errorf("failed re-attach warned %d times, want exactly the one loss WARN; log:\n%s", got, logged)
	}
	if strings.Contains(logged, `"level":"INFO"`) {
		t.Errorf("failed re-attach logged an INFO; log:\n%s", logged)
	}

	w.registerErr = nil
	sup.check() // the next beat heals

	logged2 := buf.String()
	if got := strings.Count(logged2, `"level":"WARN"`); got != 1 {
		t.Errorf("total WARNs across the failed and healed beats = %d, want exactly one", got)
	}
	if got := strings.Count(logged2, `"level":"INFO"`); got != 1 {
		t.Errorf("total INFOs across the failed and healed beats = %d, want exactly one", got)
	}
	if got := w.registerCalls(); !equalStrings(got, []string{testService, testService, testService}) {
		t.Errorf("registrations = %q, want attach + failed re-attach + healed re-attach", got)
	}
	if item == nil {
		t.Fatal("supervised item is nil")
	}
}

// TestSupervisorRevivalRefreshesIcon pins the icon contract across a lost
// period: ModeChanged during the lost period keeps the served pixmap fresh
// WITHOUT emitting (an unregistered item has no shell listener); the
// revival fires exactly one NewIcon and the served pixmap is current.
func TestSupervisorRevivalRefreshesIcon(t *testing.T) {
	buf := captureLogs(t)
	item, sup, w, em, _ := attachSupervised()
	w.setItems(nil)
	w.registerErr = errWatchInjected

	sup.check() // lost: the revival is deferred

	item.ModeChanged(symbolRU)
	if got := len(em.emitCalls()); got != 0 {
		t.Errorf("lost-period ModeChanged emitted %v — the item is unregistered", got)
	}
	v, derr := item.Get(sniIface, propIconPixmap)
	if derr != nil || !variantIsPixmap(v, symbolRU) {
		t.Errorf("lost-period IconPixmap = (%v, %v), want the fresh RU composition", v.Value(), derr)
	}

	w.registerErr = nil
	sup.check() // revival

	if calls := em.emitCalls(); len(calls) != 1 || calls[0].signal != signalNewIcon {
		t.Errorf("revival emits = %v, want exactly one NewIcon", calls)
	}
	v2, derr := item.Get(sniIface, propIconPixmap)
	if derr != nil || !variantIsPixmap(v2, symbolRU) {
		t.Errorf("post-revival IconPixmap = (%v, %v), want the fresh RU composition", v2.Value(), derr)
	}
	if got := strings.Count(buf.String(), `"level":"INFO"`); got != 1 {
		t.Errorf("revival logged %d INFO records, want exactly one", got)
	}
}

// TestSupervisorConsecutiveFailuresOneWarn pins the log discipline under a
// persistently broken watcher: any number of consecutive failed beats
// produce exactly ONE WARN total and no INFO — the retry waits for the
// next beat instead of looping.
func TestSupervisorConsecutiveFailuresOneWarn(t *testing.T) {
	buf := captureLogs(t)
	item, sup, w, _, _ := attachSupervised()
	w.itemsErr = errItemsInjected
	w.registerErr = errWatchInjected

	for range 3 {
		sup.check() // the registry read fails
	}
	w.itemsErr = nil
	w.setOwner(false)
	for range 2 {
		sup.check() // the probe fails
	}

	if got := strings.Count(buf.String(), `"level":"WARN"`); got != 1 {
		t.Errorf("five failed beats warned %d times, want exactly one", got)
	}
	if strings.Contains(buf.String(), `"level":"INFO"`) {
		t.Error("failed beats logged an INFO")
	}
	if item == nil {
		t.Fatal("supervised item is nil")
	}
}

// TestSupervisorRunExitsOnContextDone pins the goroutine exit condition:
// run returns when the context is cancelled — the daemon's shutdown must
// never strand the supervisor.
func TestSupervisorRunExitsOnContextDone(t *testing.T) {
	item, sup, w, _, _ := attachSupervised()
	_ = item
	w.setItems([]string{watcherItem})
	sup.signals = make(chan *dbus.Signal, 1)
	sup.ticks = make(chan time.Time)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		sup.run(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("run did not exit on ctx done")
	}
}

// TestSupervisorRunConsumesBeats pins the select loop's duty: a tick sent
// while run is active is PROCESSED (the eviction heals) before the exit —
// the synthetic-channel drive of the production loop.
func TestSupervisorRunConsumesBeats(t *testing.T) {
	buf := captureLogs(t)
	_, sup, w, em, _ := attachSupervised()
	w.setItems(nil) // evicted behind run's back
	ticks := make(chan time.Time)
	sup.ticks = ticks
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		sup.run(ctx)
		close(done)
	}()

	ticks <- time.Now() // blocks until run receives the beat
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("run did not exit on ctx done")
	}

	if got := w.registerCalls(); !equalStrings(got, []string{testService, testService}) {
		t.Errorf("registrations = %q, want attach + the beat's re-attach", got)
	}
	logged := buf.String()
	if strings.Count(logged, `"level":"WARN"`) != 1 || strings.Count(logged, `"level":"INFO"`) != 1 {
		t.Errorf("healed beat log discipline broken (one WARN + one INFO wanted); log:\n%s", logged)
	}
	if calls := em.emitCalls(); len(calls) != 1 || calls[0].signal != signalNewIcon {
		t.Errorf("revival emits = %v, want exactly one NewIcon", calls)
	}
}
