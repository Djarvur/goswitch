package engine_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/godbus/dbus/v5"

	"github.com/Djarvur/goswitch/engine"
	"github.com/Djarvur/goswitch/internal/logging"
)

// recordingHandler captures everything the engine funnels through the seam.
type recordingHandler struct {
	events   []engine.EngineEvent
	kinds    []engine.LifecycleKind
	texts    []string
	cursors  []uint32
	anchors  []uint32
	caps     []uint32
	attached []engine.Emitter
}

func (h *recordingHandler) HandleKey(ev engine.EngineEvent) bool {
	h.events = append(h.events, ev)

	return false
}

func (h *recordingHandler) HandleLifecycle(kind engine.LifecycleKind) {
	h.kinds = append(h.kinds, kind)
}

func (h *recordingHandler) HandleSurroundingText(text string, cursorPos, anchorPos uint32) {
	h.texts = append(h.texts, text)
	h.cursors = append(h.cursors, cursorPos)
	h.anchors = append(h.anchors, anchorPos)
}

func (h *recordingHandler) HandleCapabilities(caps uint32) {
	h.caps = append(h.caps, caps)
}

func (h *recordingHandler) AttachEngine(eng engine.Emitter) {
	h.attached = append(h.attached, eng)
}

// panicHandler injects a panic into every seam call (INTEG-05 test double).
type panicHandler struct{}

func (panicHandler) HandleKey(engine.EngineEvent) bool {
	panic("injected handler panic")
}

func (panicHandler) HandleLifecycle(engine.LifecycleKind) {
	panic("injected handler panic")
}

func (panicHandler) HandleSurroundingText(string, uint32, uint32) {}

func (panicHandler) HandleCapabilities(uint32) {}

func (panicHandler) AttachEngine(engine.Emitter) {}

// discardLogger resets the default slog logger after a test replaced it.
func discardLogger(t *testing.T) {
	t.Helper()

	t.Cleanup(func() {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	})
}

// TestEngine_ProcessKeyEventReturnsFalse pins the decline half of the
// consume contract (formerly the Phase 1 observer contract, INTEG-02): with
// a handler that declines, every event — bare modifiers included — transits
// and decodes exactly once on the way through.
func TestEngine_ProcessKeyEventReturnsFalse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		keyval      uint32
		keycode     uint32
		state       uint32
		wantRelease bool
		wantMods    uint32
	}{
		{"shift_r press", engine.KeyShiftR, 62, 0x1, false, 0x1},
		{"shift_r release", engine.KeyShiftR, 62, 0x1 | engine.MaskRelease, true, 0x1},
		{"letter press", 0x67, 38, 0, false, 0},
		{"letter release", 0x67, 38, engine.MaskRelease, true, 0},
		{"shift+letter", 0x47, 38, engine.MaskShift, false, engine.MaskShift},
		{"ctrl+letter", 0x67, 38, engine.MaskControl, false, engine.MaskControl},
		{"alt+shift+letter", 0x67, 38, engine.MaskShift | engine.MaskMod1, false, engine.MaskShift | engine.MaskMod1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := &recordingHandler{}
			eng := engine.NewEngine(rec, "goswitch-en")

			handled, err := eng.ProcessKeyEvent(tt.keyval, tt.keycode, tt.state)
			if err != nil {
				t.Fatalf("ProcessKeyEvent() err = %v, want nil", err)
			}
			if handled {
				t.Errorf("ProcessKeyEvent() = true, want false (observer mode, INTEG-02)")
			}
			if len(rec.events) != 1 {
				t.Fatalf("handler got %d events, want 1", len(rec.events))
			}

			ev := rec.events[0]
			if ev.Keyval != tt.keyval {
				t.Errorf("Keyval = 0x%x, want 0x%x", ev.Keyval, tt.keyval)
			}
			if ev.Keycode != tt.keycode {
				t.Errorf("Keycode = %d, want %d", ev.Keycode, tt.keycode)
			}
			if ev.Release != tt.wantRelease {
				t.Errorf("Release = %v, want %v", ev.Release, tt.wantRelease)
			}
			if ev.Mods != tt.wantMods {
				t.Errorf("Mods = 0x%x, want 0x%x", ev.Mods, tt.wantMods)
			}
		})
	}
}

// consumeStubHandler is the configurable decision double of the consume
// contract: HandleKey answers the configured verdict, nothing else.
type consumeStubHandler struct {
	consume bool
}

func (h *consumeStubHandler) HandleKey(engine.EngineEvent) bool { return h.consume }

func (h *consumeStubHandler) HandleLifecycle(engine.LifecycleKind) {}

func (h *consumeStubHandler) HandleSurroundingText(string, uint32, uint32) {}

func (h *consumeStubHandler) HandleCapabilities(uint32) {}

func (h *consumeStubHandler) AttachEngine(engine.Emitter) {}

// TestEngine_ProcessKeyEventConsumePropagation pins the 02-04 contract that
// replaces the Phase 1 observer-false table: the handler's verdict IS the
// method's answer — true means consumed (the RU-mode commit path), false
// means transit; a nil handler stays a pure observer.
func TestEngine_ProcessKeyEventConsumePropagation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler engine.EventHandler
		want    bool
	}{
		{"handler consumes", &consumeStubHandler{consume: true}, true},
		{"handler declines", &consumeStubHandler{consume: false}, false},
		{"nil handler observes", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			eng := engine.NewEngine(tt.handler, "goswitch-en")
			handled, err := eng.ProcessKeyEvent(0x67, 38, 0)
			if err != nil {
				t.Fatalf("ProcessKeyEvent() err = %v, want nil", err)
			}
			if handled != tt.want {
				t.Errorf("ProcessKeyEvent() = %v, want %v (handler verdict must propagate)", handled, tt.want)
			}
		})
	}
}

// ibusTextVariant builds the wire shape of a SetSurroundingText payload the
// way godbus decodes it on the incoming path: the IBusText struct as a
// positional []any (STRUCT never decodes into the typed struct).
func ibusTextVariant(t *testing.T, text string) dbus.Variant {
	t.Helper()

	sig, err := dbus.ParseSignature("(sa{sv}sv)")
	if err != nil {
		t.Fatalf("parse struct signature: %v", err)
	}
	attrSig, err := dbus.ParseSignature("(sa{sv}av)")
	if err != nil {
		t.Fatalf("parse attrlist signature: %v", err)
	}

	return dbus.MakeVariantWithSignature([]any{
		"IBusText",
		map[string]dbus.Variant{},
		text,
		dbus.MakeVariantWithSignature([]any{
			"IBusAttrList",
			map[string]dbus.Variant{},
			[]dbus.Variant{},
		}, attrSig),
	}, sig)
}

// TestEngine_SurroundingTextForward pins the incoming decode of the
// surrounding text (plan 02-03) and the anchorPos seam widening (plan 03-03,
// Pitfall 1): a struct-shaped payload is decoded and forwarded with text and
// BOTH positions — cursor_pos and anchor_pos reach the handler verbatim, the
// selection anchor no longer dies at the seam (D-30); every other shape is
// dropped without a call. Capabilities forward likewise.
func TestEngine_SurroundingTextForward(t *testing.T) {
	t.Parallel()

	rec := &recordingHandler{}
	eng := engine.NewEngine(rec, "goswitch-en")

	if err := eng.SetSurroundingText(ibusTextVariant(t, "ghbdtn"), 6, 0); err != nil {
		t.Fatalf("SetSurroundingText() err = %v, want nil", err)
	}
	if err := eng.SetSurroundingText(dbus.MakeVariant("not a struct"), 1, 1); err != nil {
		t.Fatalf("SetSurroundingText(malformed) err = %v, want nil", err)
	}

	if len(rec.texts) != 1 || rec.texts[0] != "ghbdtn" {
		t.Errorf("forwarded texts = %q, want exactly [ghbdtn]", rec.texts)
	}
	if len(rec.cursors) != 1 || rec.cursors[0] != 6 {
		t.Errorf("forwarded cursors = %v, want exactly [6]", rec.cursors)
	}
	if len(rec.anchors) != 1 || rec.anchors[0] != 0 {
		t.Errorf("forwarded anchors = %v, want exactly [0] — the selection anchor"+
			" must survive the seam (D-30)", rec.anchors)
	}

	if err := eng.SetCapabilities(engine.CapSurroundingText); err != nil {
		t.Fatalf("SetCapabilities() err = %v, want nil", err)
	}
	if len(rec.caps) != 1 || rec.caps[0] != engine.CapSurroundingText {
		t.Errorf("forwarded caps = %v, want exactly [CapSurroundingText]", rec.caps)
	}
}

// TestEmitters_DetachedQuiet pins the detached-engine contract of every
// emitter: with no connection bound (headless construction, the state every
// unit test creates) the calls are quiet returns — nothing panics, nothing
// is expected on the wire.
func TestEmitters_DetachedQuiet(t *testing.T) {
	t.Parallel()

	eng := engine.NewEngine(nil, "goswitch-en")
	eng.RequireSurroundingText()
	eng.DeleteSurroundingText(-6, 6)
	eng.CommitText(engine.NewIBusText("quiet"))
	eng.ForwardKeyEvent(engine.KeyBackSpace, 14, 0)
}

// TestEngine_DecodeState pins the one-shot decode of the raw IBus state
// word: bit 30 is release, the low byte is the modifier mask.
func TestEngine_DecodeState(t *testing.T) {
	t.Parallel()

	rec := &recordingHandler{}
	eng := engine.NewEngine(rec, "goswitch-en")

	if _, err := eng.ProcessKeyEvent(0x67, 38, 0xff|engine.MaskRelease); err != nil {
		t.Fatalf("ProcessKeyEvent() err = %v, want nil", err)
	}

	if len(rec.events) != 1 {
		t.Fatalf("handler got %d events, want 1", len(rec.events))
	}
	ev := rec.events[0]
	if !ev.Release {
		t.Errorf("Release = false, want true for state bit 30")
	}
	if ev.Mods != 0xff {
		t.Errorf("Mods = 0x%x, want 0xff (masked to the low byte)", ev.Mods)
	}
}

// TestEngine_KeyConstants pins the keyval/mask/capability constants
// against the verified IBus tables (/usr/include/ibus-1.0/*).
func TestEngine_KeyConstants(t *testing.T) {
	t.Parallel()

	keyvals := map[string]uint32{
		"KeyBackSpace": engine.KeyBackSpace,
		"KeyTab":       engine.KeyTab,
		"KeyReturn":    engine.KeyReturn,
		"KeyEscape":    engine.KeyEscape,
		"KeyKPEnter":   engine.KeyKPEnter,
		"KeyShiftL":    engine.KeyShiftL,
		"KeyShiftR":    engine.KeyShiftR,
		"KeyControlL":  engine.KeyControlL,
		"KeySpace":     engine.KeySpace,
	}
	wantKeyvals := map[string]uint32{
		"KeyBackSpace": 0xff08,
		"KeyTab":       0xff09,
		"KeyReturn":    0xff0d,
		"KeyEscape":    0xff1b,
		"KeyKPEnter":   0xff8b,
		"KeyShiftL":    0xffe1,
		"KeyShiftR":    0xffe2,
		"KeyControlL":  0xffe3,
		"KeySpace":     0x020,
	}
	for name, want := range wantKeyvals {
		if got := keyvals[name]; got != want {
			t.Errorf("%s = 0x%x, want 0x%x", name, got, want)
		}
	}

	masks := map[string]uint32{
		"MaskShift":          engine.MaskShift,
		"MaskControl":        engine.MaskControl,
		"MaskMod1":           engine.MaskMod1,
		"MaskMod4":           engine.MaskMod4,
		"MaskMod5":           engine.MaskMod5,
		"MaskRelease":        engine.MaskRelease,
		"CapPreeditText":     engine.CapPreeditText,
		"CapSurroundingText": engine.CapSurroundingText,
		"CapSyncProcessKey":  engine.CapSyncProcessKey,
	}
	wantMasks := map[string]uint32{
		"MaskShift":          1 << 0,
		"MaskControl":        1 << 2,
		"MaskMod1":           1 << 3,
		"MaskMod4":           1 << 6,
		"MaskMod5":           1 << 7,
		"MaskRelease":        1 << 30,
		"CapPreeditText":     1 << 0,
		"CapSurroundingText": 1 << 5,
		"CapSyncProcessKey":  1 << 7,
	}
	for name, want := range wantMasks {
		if got := masks[name]; got != want {
			t.Errorf("%s = 0x%x, want 0x%x", name, got, want)
		}
	}
}

// allDBusHandlers returns one invocation for every exported D-Bus handler
// of the engine object — running them to completion is the containment
// assertion: an escaping panic would fail the test by propagating.
func allDBusHandlers(eng *engine.Engine) map[string]func() {
	return map[string]func(){
		"ProcessKeyEvent":    func() { _, _ = eng.ProcessKeyEvent(engine.KeyShiftR, 62, 0) },
		"SetCursorLocation":  func() { _ = eng.SetCursorLocation(1, 2, 3, 4) },
		"SetSurroundingText": func() { _ = eng.SetSurroundingText(dbus.MakeVariant("x"), 0, 0) },
		"SetCapabilities":    func() { _ = eng.SetCapabilities(engine.CapSurroundingText) },
		"FocusIn":            func() { _ = eng.FocusIn() },
		"FocusOut":           func() { _ = eng.FocusOut() },
		"Reset":              func() { _ = eng.Reset() },
		"Enable":             func() { _ = eng.Enable() },
		"Disable":            func() { _ = eng.Disable() },
		"PageUp":             func() { _ = eng.PageUp() },
		"PageDown":           func() { _ = eng.PageDown() },
		"CursorUp":           func() { _ = eng.CursorUp() },
		"CursorDown":         func() { _ = eng.CursorDown() },
		"CandidateClicked":   func() { _ = eng.CandidateClicked(0, 1, 0) },
		"PropertyActivate":   func() { _ = eng.PropertyActivate("p", 0) },
		"PropertyShow":       func() { _ = eng.PropertyShow("p") },
		"PropertyHide":       func() { _ = eng.PropertyHide("p") },
		"Destroy":            func() { _ = eng.Destroy() },
		"Get":                func() { _, _ = eng.Get("x") },
		"GetAll":             func() { _, _ = eng.GetAll("x") },
		"Set":                func() { _ = eng.Set("x", dbus.MakeVariant("x")) },
	}
}

// TestEngine_RecoverContainsPanic proves the recover shim (INTEG-05): a
// panic injected through the seam never escapes a D-Bus handler, the
// replies stay safe, and the containment lands in the log at ERROR.
func TestEngine_RecoverContainsPanic(t *testing.T) {
	discardLogger(t)

	var logBuf bytes.Buffer
	logging.Setup(&logBuf, true)
	eng := engine.NewEngine(panicHandler{}, "goswitch-en")

	handled, keyErr := eng.ProcessKeyEvent(engine.KeyShiftR, 62, 0)
	if keyErr != nil {
		t.Errorf("ProcessKeyEvent() err = %v, want nil safe reply", keyErr)
	}
	if handled {
		t.Errorf("ProcessKeyEvent() = true, want false safe reply")
	}

	value, getErr := eng.Get("x")
	if getErr != nil {
		t.Errorf("Get() err = %v, want nil safe reply", getErr)
	}
	if value.Value() != "" {
		t.Errorf("Get() = %v, want an empty safe reply", value.Value())
	}
	props, getAllErr := eng.GetAll("x")
	if getAllErr != nil {
		t.Errorf("GetAll() err = %v, want nil safe reply", getAllErr)
	}
	if len(props) != 0 {
		t.Errorf("GetAll() = %v, want an empty safe reply", props)
	}

	for name, call := range allDBusHandlers(eng) {
		call() // a panic escaping here fails the test by propagating
		_ = name
	}

	contained := strings.Count(logBuf.String(), `"msg":"handler panic contained"`)
	if want := 6; contained < want { // ProcessKeyEvent + 5 lifecycle funnels
		t.Errorf("log has %d contained-panic records, want at least %d:\n%s", contained, want, logBuf.String())
	}

	for _, line := range strings.Split(strings.TrimSpace(logBuf.String()), "\n") {
		if !strings.Contains(line, "handler panic contained") {
			continue // ordinary operation lines are expected in the same stream
		}

		var rec struct {
			Level string `json:"level"`
		}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("log line is not JSON: %q: %v", line, err)
		}
		if rec.Level != "ERROR" {
			t.Errorf("panic containment logged at level %q, want ERROR: %q", rec.Level, line)
		}
	}
}
