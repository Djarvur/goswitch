package appid

import (
	"context"
	"errors"
	"testing"

	"github.com/godbus/dbus/v5"
)

// errRegistryDenied is the refused-registry sentinel of the interest corpus
// (the err113 discipline).
var errRegistryDenied = errors.New("app identity test registry: denied")

// fakeRegistry records the registry calls of the interest seam and can
// refuse them (the dbus.BusObject shape, recording form).
type fakeRegistry struct {
	calls   []string
	args    [][]interface{}
	failErr error
}

func (f *fakeRegistry) Call(method string, _ dbus.Flags, args ...interface{}) *dbus.Call {
	return f.CallWithContext(nil, method, 0, args...)
}

func (f *fakeRegistry) CallWithContext(_ context.Context, method string, _ dbus.Flags, args ...interface{}) *dbus.Call {
	f.calls = append(f.calls, method)
	f.args = append(f.args, args)

	return &dbus.Call{Err: f.failErr}
}

func (f *fakeRegistry) Go(method string, _ dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
	call := f.CallWithContext(nil, method, 0, args...)
	go func() { ch <- call }()

	return call
}

func (f *fakeRegistry) GoWithContext(ctx context.Context, method string, _ dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
	call := f.CallWithContext(ctx, method, 0, args...)
	go func() { ch <- call }()

	return call
}

func (f *fakeRegistry) AddMatchSignal(_ string, _ string, _ ...dbus.MatchOption) *dbus.Call {
	return &dbus.Call{}
}

func (f *fakeRegistry) RemoveMatchSignal(_ string, _ string, _ ...dbus.MatchOption) *dbus.Call {
	return &dbus.Call{}
}

func (f *fakeRegistry) GetProperty(string) (dbus.Variant, error) {
	return dbus.Variant{}, errRegistryDenied
}

func (f *fakeRegistry) StoreProperty(string, interface{}) error {
	return errRegistryDenied
}

func (f *fakeRegistry) SetProperty(string, interface{}) error {
	return errRegistryDenied
}

func (f *fakeRegistry) Destination() string {
	return a11yRegistryName
}

func (f *fakeRegistry) Path() dbus.ObjectPath {
	return a11yRegistryPath
}

// TestRegisterEventInterest_PinsRegistryCall pins the event-interest seam:
// exactly one RegisterEvent call to the at-spi registry object carrying the
// state-changed event name — the registration every real a11y consumer
// performs (the application bridges consult the registry's registered-event
// list and emit signals ONLY while some listener holds one; the 2026-10-12
// stand finding: a bare session emitted zero StateChanged signals until a
// listener registered, then the signals flowed) — and the refusal path
// returns the registry error wrapped, named.
func TestRegisterEventInterest_PinsRegistryCall(t *testing.T) {
	reg := &fakeRegistry{}
	if err := registerEventInterest(reg, eventStateInterest); err != nil {
		t.Fatalf("registerEventInterest = %v, want nil", err)
	}
	if len(reg.calls) != 1 || reg.calls[0] != a11yRegistryName+".RegisterEvent" {
		t.Fatalf("registry calls %v, want exactly [%s.RegisterEvent]", reg.calls, a11yRegistryName)
	}
	if len(reg.args) != 1 || len(reg.args[0]) != 1 || reg.args[0][0] != eventStateInterest {
		t.Fatalf("registry args %v, want exactly [%q]", reg.args, eventStateInterest)
	}

	reg.failErr = errRegistryDenied
	if err := registerEventInterest(reg, eventStateInterest); !errors.Is(err, errRegistryDenied) {
		t.Fatalf("registerEventInterest over a refused registry = %v, want the wrapped registry error", err)
	}
}
