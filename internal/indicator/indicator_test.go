//nolint:testpackage // drives the unexported attach seam and wire constants — the sanctioned in-package corpus seam
package indicator

import (
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync"
	"bytes"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

// The tray corpus constants (quick plan 260930-pf6) — the wire names the
// tests pin and the well-known service name the daemon registers under.
const (
	testService = "org.djarvur.goswitch"
	testSymbol  = "ru"
)

// errWatchInjected is the probe/register failure the degradation corpus
// injects (err113).
var errWatchInjected = errors.New("watcher exploded")

// errEmitInjected is the emit failure of the self-disable corpus.
var errEmitInjected = errors.New("emit exploded")

// syncBuffer is the guarded log buffer of the degradation corpus (the
// session corpus's captureLogs discipline, local copy — dupl is relaxed in
// tests).
type syncBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

// Write appends to the buffer under the guard.
func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

// String snapshots the buffer under the guard.
func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

// captureLogs redirects the process default logger into a guarded buffer.
func captureLogs(t *testing.T) *syncBuffer {
	t.Helper()
	buf := &syncBuffer{}
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, nil)))
	t.Cleanup(func() { slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil))) })

	return buf
}

// fakeWatcher is the watcher-probe double: canned owner presence and
// register verdicts, with the registration calls recorded under a mutex.
type fakeWatcher struct {
	mu          sync.Mutex
	owner       bool
	probeErr    error
	registerErr error
	registered  []string
}

// NameHasOwner answers the canned probe.
func (f *fakeWatcher) NameHasOwner() (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.owner, f.probeErr
}

// RegisterStatusNotifierItem records the service and answers the canned
// verdict.
func (f *fakeWatcher) RegisterStatusNotifierItem(service string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.registered = append(f.registered, service)

	return f.registerErr
}

// registerCalls snapshots the recorded registrations.
func (f *fakeWatcher) registerCalls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]string(nil), f.registered...)
}

// emitCall is one recorded signal emission.
type emitCall struct {
	path   dbus.ObjectPath
	iface  string
	signal string
	args   []any
}

// fakeEmitter is the signal-emitter double: every Emit recorded under a
// mutex, with an injected failure for the self-disable corpus.
type fakeEmitter struct {
	mu    sync.Mutex
	calls []emitCall
	err   error
}

// Emit records the signal and answers the canned verdict.
func (f *fakeEmitter) Emit(objectPath dbus.ObjectPath, iface, signal string, args ...any) error {
	f.mu.Lock()
	f.calls = append(f.calls, emitCall{path: objectPath, iface: iface, signal: signal, args: args})
	err := f.err
	f.mu.Unlock()

	return err
}

// emitCalls snapshots the recorded emissions.
func (f *fakeEmitter) emitCalls() []emitCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]emitCall(nil), f.calls...)
}

// exportCall is one recorded object export.
type exportCall struct {
	path  dbus.ObjectPath
	iface string
}

// fakeExporter is the object-export double: every Export recorded under a
// mutex, with an injected failure for the degradation corpus.
type fakeExporter struct {
	mu    sync.Mutex
	calls []exportCall
	err   error
}

// Export records the object and answers the canned verdict.
func (f *fakeExporter) Export(_ any, path dbus.ObjectPath, iface string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, exportCall{path: path, iface: iface})

	return f.err
}

// exportCalls snapshots the recorded exports.
func (f *fakeExporter) exportCalls() []exportCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]exportCall(nil), f.calls...)
}

// attachOK builds an item over healthy doubles: the watcher owns the name,
// registration and exports succeed. The corpus's happy-path base.
func attachOK() (*Item, *fakeWatcher, *fakeEmitter, *fakeExporter) {
	w, em, exp := &fakeWatcher{owner: true}, &fakeEmitter{}, &fakeExporter{}

	return attach(w, em, exp, testService), w, em, exp
}

// TestAttachWatcherAbsentIsInert pins the absent-watcher degradation: one
// WARN, no registration, no export — Attach succeeds and the indicator is
// a permanent no-op that never re-checks (v0).
func TestAttachWatcherAbsentIsInert(t *testing.T) {
	buf := captureLogs(t)
	w, em, exp := &fakeWatcher{}, &fakeEmitter{}, &fakeExporter{}

	item := attach(w, em, exp, testService)

	if item == nil {
		t.Fatal("attach returned nil — Attach must succeed even on degradation")
	}
	logged := buf.String()
	if !strings.Contains(logged, `"component":"tray indicator"`) {
		t.Errorf("absent-watcher WARN missing the component token; log:\n%s", logged)
	}
	if got := strings.Count(logged, `"level":"WARN"`); got != 1 {
		t.Errorf("absent watcher warned %d times, want exactly one; log:\n%s", got, logged)
	}
	if got := w.registerCalls(); len(got) != 0 {
		t.Errorf("absent watcher still received registrations %q — the probe must gate the register", got)
	}
	if got := exp.exportCalls(); len(got) != 0 {
		t.Errorf("absent watcher still exported %v — an inert display exports nothing", got)
	}

	item.ModeChanged(testSymbol) // must be a silent no-op
	if got := em.emitCalls(); len(got) != 0 {
		t.Errorf("inert display emitted %v on ModeChanged — a degraded item never emits", got)
	}
	if logged2 := buf.String(); strings.Count(logged2, `"level":"WARN"`) != 1 {
		t.Errorf("post-degradation ModeChanged warned again; log:\n%s", logged2)
	}
}

// TestAttachProbeFailureIsInert pins the probe-error degradation: a broken
// probe is the same one-WARN inert display — the watcher's reply is never
// trusted beyond success/failure (T-Q6-02).
func TestAttachProbeFailureIsInert(t *testing.T) {
	buf := captureLogs(t)
	w, em, exp := &fakeWatcher{probeErr: errWatchInjected}, &fakeEmitter{}, &fakeExporter{}

	item := attach(w, em, exp, testService)

	if item == nil {
		t.Fatal("attach returned nil — Attach must succeed even on degradation")
	}
	if got := strings.Count(buf.String(), `"level":"WARN"`); got != 1 {
		t.Errorf("probe failure warned %d times, want exactly one; log:\n%s", got, buf.String())
	}
	if got := w.registerCalls(); len(got) != 0 {
		t.Errorf("failed probe still registered %q", got)
	}
	item.ModeChanged(testSymbol)
	if got := em.emitCalls(); len(got) != 0 {
		t.Errorf("inert display emitted %v — a degraded item never emits", got)
	}
}

// TestAttachRegistersAndExports pins the healthy attach: the watcher probe
// passes, the item is exported at /StatusNotifierItem under BOTH the SNI
// interface and the Properties surface (the export precedes the register —
// the watcher's immediate property reads must find the object), and the
// registration names exactly the daemon's well-known service.
func TestAttachRegistersAndExports(t *testing.T) {
	buf := captureLogs(t)
	item, w, _, exp := attachOK()

	if got := strings.Count(buf.String(), `"level":"WARN"`); got != 0 {
		t.Errorf("healthy attach warned %d times; log:\n%s", got, buf.String())
	}
	if got := w.registerCalls(); !equalStrings(got, []string{testService}) {
		t.Errorf("registrations = %q, want exactly [%s] once", got, testService)
	}
	wantExports := []exportCall{
		{path: itemPath, iface: sniIface},
		{path: itemPath, iface: propertiesIface},
	}
	if got := exp.exportCalls(); !equalExports(got, wantExports) {
		t.Errorf("exports = %v, want the item at %s under %s and %s", got, itemPath, sniIface, propertiesIface)
	}

	// The served pixmap starts at EN (ADR-001) — the install-push corrects
	// any skew immediately after attach.
	v, derr := item.Get(sniIface, propIconPixmap)
	if derr != nil {
		t.Fatalf("initial IconPixmap Get error = %v, want nil", derr)
	}
	if !variantIsPixmap(v, testSymbolEN) {
		t.Error("initial IconPixmap is not the EN composition — the attach default must be EN")
	}
}

// TestAttachModeChangedEmitsAndServesPins the mode-change contract: a known
// symbol swaps the served pixmap and emits NewIcon on the item path; an
// unknown symbol is one WARN no-op that changes nothing.
func TestAttachModeChangedEmitsAndServesPins(t *testing.T) {
	buf := captureLogs(t)
	item, _, em, _ := attachOK()

	item.ModeChanged(testSymbol)

	calls := em.emitCalls()
	if len(calls) != 1 {
		t.Fatalf("ModeChanged emitted %d signals, want exactly one NewIcon", len(calls))
	}
	call := calls[0]
	if call.path != itemPath || call.iface != sniIface || call.signal != signalNewIcon {
		t.Errorf("emit = %s %s.%s, want %s %s.NewIcon", call.path, call.iface, call.signal, itemPath, sniIface)
	}
	v, derr := item.Get(sniIface, propIconPixmap)
	if derr != nil {
		t.Fatalf("IconPixmap Get after ModeChanged error = %v, want nil", derr)
	}
	if !variantIsPixmap(v, testSymbol) {
		t.Error("IconPixmap after ModeChanged(ru) is not the RU composition")
	}

	// The unknown-symbol branch: one WARN, no emit, the served pixmap unchanged.
	item.ModeChanged("xx")
	if got := len(em.emitCalls()); got != 1 {
		t.Errorf("unknown symbol produced %d emits, want none", got)
	}
	if !strings.Contains(buf.String(), `"component":"tray indicator"`) {
		t.Errorf("unknown-symbol WARN missing the component token; log:\n%s", buf.String())
	}
	if v2, derr := item.Get(sniIface, propIconPixmap); derr != nil || !variantIsPixmap(v2, testSymbol) {
		t.Error("unknown symbol changed the served pixmap — the last good icon must stand")
	}
}

// TestAttachEmitFailureSelfDisables pins the dead-connection degradation
// (the ctl-shutdown case): the first emit error is one WARN and every
// subsequent ModeChanged is a silent no-op.
func TestAttachEmitFailureSelfDisables(t *testing.T) {
	buf := captureLogs(t)
	w, em, exp := &fakeWatcher{owner: true}, &fakeEmitter{err: errEmitInjected}, &fakeExporter{}
	item := attach(w, em, exp, testService)

	item.ModeChanged(testSymbol) // the emit fails
	item.ModeChanged(testSymbolEN)
	item.ModeChanged("xx")

	if got := len(em.emitCalls()); got != 1 {
		t.Errorf("emit attempts = %d, want exactly 1 — after the failure the display must be silent", got)
	}
	logged := buf.String()
	if !strings.Contains(logged, `"component":"tray indicator"`) {
		t.Errorf("emit-failure WARN missing the component token; log:\n%s", logged)
	}
	if got := strings.Count(logged, `"level":"WARN"`); got != 1 {
		t.Errorf("emit failure warned %d times, want exactly one; log:\n%s", got, logged)
	}
}

// TestItemProperties pins the served property surface: the SNI v0 set
// (icon-only — no Menu, no ToolTip), the unknown-property and
// unknown-interface refusals, and the read-only verdict of Set.
func TestItemProperties(t *testing.T) {
	item, _, em, _ := attachOK()

	v, derr := item.Get(sniIface, propIconPixmap)
	if derr != nil || !variantIsPixmap(v, testSymbolEN) {
		t.Fatalf("IconPixmap = (%v, %v), want the EN pixmap", v, derr)
	}
	for prop, want := range map[string]string{
		propCategory: "ApplicationStatus",
		propID:       serviceID,
		propTitle:    serviceID,
		propStatus:   "Active",
		propIconName: "",
	} {
		got, derr := item.Get(sniIface, prop)
		if derr != nil || got.Value() != want {
			t.Errorf("Get(%s) = (%v, %v), want %q", prop, got.Value(), derr, want)
		}
	}
	got, derr := item.Get(sniIface, propWindowID)
	if derr != nil || got.Value() != int32(0) {
		t.Errorf("Get(WindowId) = (%v, %v), want 0", got.Value(), derr)
	}

	if _, derr := item.Get(sniIface, "Nope"); derr == nil {
		t.Error("Get of an unknown property returned no error")
	}
	if _, derr := item.GetAll("org.freedesktop.DBus.Error"); derr == nil {
		t.Error("GetAll of a foreign interface returned no error")
	}
	all, derr := item.GetAll(sniIface)
	if derr != nil {
		t.Fatalf("GetAll error = %v, want nil", derr)
	}
	for _, want := range []string{propCategory, propID, propTitle, propStatus, propIconName, propIconPixmap, propWindowID} {
		if _, ok := all[want]; !ok {
			t.Errorf("GetAll is missing %s", want)
		}
	}
	if len(all) != 7 {
		t.Errorf("GetAll serves %d properties, want exactly 7 (icon-only v0 surface)", len(all))
	}
	if derr := item.Set(sniIface, propStatus, dbus.MakeVariant("Passive")); derr == nil {
		t.Error("Set returned no error — every item property is read-only")
	}

	// The ModeDisplay contract drives nothing here: ModeChanged's emit is
	// counted above; this corpus only re-checks that property reads and
	// emits coexist under the same mutex (the concurrent hammer below is
	// the race detector's real subject).
	_ = em
}

// TestItemConcurrentModeAndReads hammers the item's mutex the way godbus
// dispatch does: property reads on foreign goroutines while ModeChanged
// swaps the pixmap — the race detector must stay quiet (T-Q6-04).
func TestItemConcurrentModeAndReads(t *testing.T) {
	item, _, em, _ := attachOK()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 200 {
			_, _ = item.GetAll(sniIface)
			_, _ = item.Get(sniIface, propIconPixmap)
		}
	}()
	for i := range 200 {
		item.ModeChanged([]string{testSymbolEN, testSymbol}[i%2])
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent reader never finished")
	}
	if got := len(em.emitCalls()); got != 200 {
		t.Errorf("emits = %d, want 200 (one per flip)", got)
	}
}

// variantIsPixmap reports whether the property variant carries the
// composition of the given symbol — the (iiay) body compared byte for byte.
func variantIsPixmap(v dbus.Variant, symbol string) bool {
	pixmaps, ok := v.Value().([]Pixmap)
	if !ok || len(pixmaps) != 1 {
		return false
	}
	want, ok := PixmapFor(symbol)
	if !ok {
		return false
	}
	pm := pixmaps[0]

	return pm.Width == want.Width && pm.Height == want.Height && bytes.Equal(pm.Pixels, want.Pixels)
}

// equalStrings reports slice equality without importing slices for two
// call sites (the corpus stays on the stdlib basics).
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// equalExports reports export-call equality in order.
func equalExports(a, b []exportCall) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
