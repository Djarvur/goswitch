//nolint:testpackage // binds unexported conn/path — the sanctioned in-package emission seam (02-PATTERNS § engine)
package engine

// Wire pins of the engine emitters (plan 02-03). This file is deliberately
// IN-PACKAGE: e.conn/e.path are unexported and bound only in
// factory.CreateEngine, so an external test cannot reach an engine that
// emits (02-PATTERNS § engine — "emission tests need an in-package seam").
// The seam is godbus's WithOutgoingInterceptor: every Emit call hands the
// outgoing message to the interceptor synchronously before it is written,
// so the signal name and argument order are observable without a bus
// (godbus ships no server-side SASL, so a paired in-memory connection is
// not an option).

import (
	"io"
	"sync"
	"testing"

	"github.com/godbus/dbus/v5"
)

// capturedSignal is one outgoing D-Bus signal observed through the
// interceptor: its object path, interface, member and positional body.
type capturedSignal struct {
	path   dbus.ObjectPath
	iface  string
	member string
	body   []any
}

// emitRecorder observes every signal an Engine emits over an intercepted
// connection. The transport underneath discards writes and parks reads —
// nothing ever answers, the interceptor is the only observer.
type emitRecorder struct {
	mu      sync.Mutex
	signals []capturedSignal
	conn    *dbus.Conn
}

// newEmitRecorder builds the intercepted connection; it is closed with the
// test.
func newEmitRecorder(t *testing.T) *emitRecorder {
	t.Helper()

	rec := &emitRecorder{}
	transport := quietTransport{closed: make(chan struct{})}
	conn, err := dbus.NewConn(transport, dbus.WithOutgoingInterceptor(func(msg *dbus.Message) {
		if msg.Type != dbus.TypeSignal {
			return
		}
		rec.mu.Lock()
		defer rec.mu.Unlock()
		rec.signals = append(rec.signals, capturedSignal{
			path:   variantPath(msg.Headers[dbus.FieldPath]),
			iface:  variantString(msg.Headers[dbus.FieldInterface]),
			member: variantString(msg.Headers[dbus.FieldMember]),
			body:   msg.Body,
		})
	}))
	if err != nil {
		t.Fatalf("build intercepted connection: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	rec.conn = conn

	return rec
}

// snapshot copies the captured signals under the guard.
func (r *emitRecorder) snapshot() []capturedSignal {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]capturedSignal(nil), r.signals...)
}

// last returns the most recently captured signal, failing the test when
// nothing was emitted.
func (r *emitRecorder) last(t *testing.T) capturedSignal {
	t.Helper()

	sigs := r.snapshot()
	if len(sigs) == 0 {
		t.Fatal("no signal was emitted")
	}

	return sigs[len(sigs)-1]
}

// wantEngineSignal fails unless sig is a signal on the engine interface at
// the expected path with the expected member.
func wantEngineSignal(t *testing.T, sig capturedSignal, member string) {
	t.Helper()

	if sig.iface != ifaceEngine {
		t.Errorf("%s iface = %q, want %q", member, sig.iface, ifaceEngine)
	}
	if sig.member != member {
		t.Errorf("member = %q, want %q", sig.member, member)
	}
	if sig.path != emitterPath {
		t.Errorf("%s path = %q, want %q", member, sig.path, emitterPath)
	}
}

// quietTransport is the far end of the intercepted connection: writes are
// discarded, reads park until Close — the messages matter only to the
// interceptor, never to a peer.
type quietTransport struct {
	closed chan struct{}
}

// Write discards the payload; godbus requires the full-length write.
func (q quietTransport) Write(p []byte) (int, error) { return len(p), nil }

// Read parks until the connection closes, then reports EOF so godbus
// read-loops (if any) terminate.
func (q quietTransport) Read([]byte) (int, error) {
	<-q.closed

	return 0, io.EOF
}

// Close unblocks Read exactly once.
func (q quietTransport) Close() error {
	select {
	case <-q.closed:
	default:
		close(q.closed)
	}

	return nil
}

// variantString extracts the string payload of a header variant.
func variantString(v dbus.Variant) string {
	if s, ok := v.Value().(string); ok {
		return s
	}

	return ""
}

// variantPath extracts the object-path payload of a header variant.
func variantPath(v dbus.Variant) dbus.ObjectPath {
	if p, ok := v.Value().(dbus.ObjectPath); ok {
		return p
	}

	return ""
}

// emitterPath is the object path the pinned engines are bound to.
const emitterPath = dbus.ObjectPath("/org/freedesktop/IBus/Engine/goswitch/1")

// newBoundEngine returns an engine bound to a fresh intercepted
// connection — one isolated recorder per caller, so parallel subtests never
// share capture state.
func newBoundEngine(t *testing.T) (*Engine, *emitRecorder) {
	t.Helper()

	rec := newEmitRecorder(t)
	eng := NewEngine(nil, "goswitch-en")
	eng.conn = rec.conn
	eng.path = emitterPath

	return eng, rec
}

// TestEmitters_DeleteAndRequire pins the wire contract of the ladder
// emitters (ADR-003): signal interface, member, path and positional
// arguments — DeleteSurroundingText(offset int32, nchars uint32) in that
// order, RequireSurroundingText with no body, CommitText with the
// variant-wrapped IBusText — all on the org.freedesktop.IBus.Engine
// interface, symmetric with the day-one CommitText emitter.
func TestEmitters_DeleteAndRequire(t *testing.T) {
	t.Parallel()

	t.Run("delete surrounding text", func(t *testing.T) {
		t.Parallel()

		eng, rec := newBoundEngine(t)
		eng.DeleteSurroundingText(-6, 6)
		sig := rec.last(t)
		wantEngineSignal(t, sig, "DeleteSurroundingText")
		if len(sig.body) != 2 {
			t.Fatalf("body has %d args, want 2: %+v", len(sig.body), sig.body)
		}
		if off, ok := sig.body[0].(int32); !ok || off != -6 {
			t.Errorf("arg0 = %#v, want int32(-6)", sig.body[0])
		}
		if n, ok := sig.body[1].(uint32); !ok || n != 6 {
			t.Errorf("arg1 = %#v, want uint32(6)", sig.body[1])
		}
	})

	t.Run("require surrounding text", func(t *testing.T) {
		t.Parallel()

		eng, rec := newBoundEngine(t)
		eng.RequireSurroundingText()
		sig := rec.last(t)
		wantEngineSignal(t, sig, "RequireSurroundingText")
		if len(sig.body) != 0 {
			t.Errorf("body = %+v, want none", sig.body)
		}
	})

	t.Run("commit text", func(t *testing.T) {
		t.Parallel()

		eng, rec := newBoundEngine(t)
		eng.CommitText(NewIBusText("привет"))
		sig := rec.last(t)
		wantEngineSignal(t, sig, "CommitText")
		if len(sig.body) != 1 {
			t.Fatalf("body has %d args, want 1: %+v", len(sig.body), sig.body)
		}
		variant, ok := sig.body[0].(dbus.Variant)
		if !ok {
			t.Fatalf("arg0 = %T, want dbus.Variant", sig.body[0])
		}
		if got := variant.Signature().String(); got != "(sa{sv}sv)" {
			t.Errorf("CommitText payload signature = %q, want (sa{sv}sv)", got)
		}
		text, ok := variant.Value().(IBusText)
		if !ok {
			t.Fatalf("variant payload = %T, want engine.IBusText", variant.Value())
		}
		if text.Text != "привет" {
			t.Errorf("text = %q, want %q", text.Text, "привет")
		}
		if got := text.AttrList.Signature().String(); got != "(sa{sv}av)" {
			t.Errorf("AttrList signature = %q, want (sa{sv}av) — an 'au' there crashes ibus-daemon 1.5.29", got)
		}
	})
}

// TestEmitters_ForwardKeyEvent pins the wire contract of the level-2 ladder
// emitter (plan 02-05, ADR-003): a ForwardKeyEvent(keyval, keycode, state)
// signal on the engine interface with the three uint32 arguments in exactly
// that positional order — the Backspace burst the no-surrounding clients
// receive before the replacement commit.
func TestEmitters_ForwardKeyEvent(t *testing.T) {
	t.Parallel()

	eng, rec := newBoundEngine(t)
	eng.ForwardKeyEvent(KeyBackSpace, 14, 0)
	sig := rec.last(t)
	wantEngineSignal(t, sig, "ForwardKeyEvent")
	if len(sig.body) != 3 {
		t.Fatalf("body has %d args, want 3: %+v", len(sig.body), sig.body)
	}
	for i, want := range []uint32{KeyBackSpace, 14, 0} {
		if got, ok := sig.body[i].(uint32); !ok || got != want {
			t.Errorf("arg%d = %#v, want uint32(%d)", i, sig.body[i], want)
		}
	}
}
