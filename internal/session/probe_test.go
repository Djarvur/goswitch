package session_test

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/Djarvur/goswitch/internal/engine"
	"github.com/Djarvur/goswitch/internal/session"
)

// The probe corpus's evdev keycodes were transcribed INDEPENDENTLY from
// /usr/include/linux/input-event-codes.h (the never-from-memory rule: the
// corpus must catch a wrong constant, not restate it) — KEY_HOME=102:177,
// KEY_LEFT=105:180, KEY_RIGHT=106:181, KEY_END=107:182, KEY_A=30:105,
// KEY_X=45:120, KEY_1=2:77, KEY_SPACE=57:132 (line numbers in the header).
const (
	tcKeyHome  = 102
	tcKeyEnd   = 107
	tcKeyLeft  = 105
	tcKeyRight = 106
	tcKeyA     = 30
	tcKeyX     = 45
	tcKey1     = 2
	tcKeySpace = 57
)

// probeEvent is one recorded ForwardKeyEvent call.
type probeEvent struct {
	keyval  uint32
	keycode uint32
	state   uint32
}

// probeSink is the engine.Emitter double of the probe corpus: every
// ForwardKeyEvent recorded under a mutex (the guarded-sink style of the
// session corpus), everything else a no-op — the probe emits nothing else.
type probeSink struct {
	mu     sync.Mutex
	events []probeEvent
}

func (p *probeSink) ForwardKeyEvent(keyval, keycode, state uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.events = append(p.events, probeEvent{keyval: keyval, keycode: keycode, state: state})
}

func (p *probeSink) RequireSurroundingText()             {}
func (p *probeSink) DeleteSurroundingText(int32, uint32) {}
func (p *probeSink) CommitText(engine.IBusText)          {}
func (p *probeSink) UpdateModeSymbol(string)             {}

// snapshot returns the recorded events.
func (p *probeSink) snapshot() []probeEvent {
	p.mu.Lock()
	defer p.mu.Unlock()

	return append([]probeEvent(nil), p.events...)
}

// TestParseProbeKey_Named pins the named-key half of the keyspec grammar:
// the four navigation keys, modifier normalization (case-insensitive
// modifiers, any combination) and case-folding on named keys only.
func TestParseProbeKey_Named(t *testing.T) {
	for _, tc := range []struct {
		spec    string
		wantKV  uint32
		wantKC  uint32
		wantMod uint32
	}{
		{spec: "Home", wantKV: engine.KeyHome, wantKC: tcKeyHome},
		{spec: "End", wantKV: engine.KeyEnd, wantKC: tcKeyEnd},
		{spec: "Left", wantKV: engine.KeyLeft, wantKC: tcKeyLeft},
		{spec: "Right", wantKV: engine.KeyRight, wantKC: tcKeyRight},
		{spec: "home", wantKV: engine.KeyHome, wantKC: tcKeyHome},
		{spec: "SHIFT+Home", wantKV: engine.KeyHome, wantKC: tcKeyHome, wantMod: engine.MaskShift},
		{spec: "ctrl+shift+Left", wantKV: engine.KeyLeft, wantKC: tcKeyLeft,
			wantMod: engine.MaskControl | engine.MaskShift},
	} {
		t.Run(tc.spec, func(t *testing.T) {
			kv, kc, mods, err := session.ParseProbeKey(tc.spec)
			if err != nil {
				t.Fatalf("ParseProbeKey(%q) error = %v, want nil", tc.spec, err)
			}
			if kv != tc.wantKV || kc != tc.wantKC || mods != tc.wantMod {
				t.Errorf("ParseProbeKey(%q) = (0x%x, %d, 0x%x), want (0x%x, %d, 0x%x)",
					tc.spec, kv, kc, mods, tc.wantKV, tc.wantKC, tc.wantMod)
			}
		})
	}
}

// TestParseProbeKey_Printable pins the printable half: one ASCII key maps to
// its Latin-1 keyval and its physical evdev keycode, and modifiers are NEVER
// inferred — "A" carries no auto-Shift (the synthetic pair answers for its
// modifiers explicitly).
func TestParseProbeKey_Printable(t *testing.T) {
	for _, tc := range []struct {
		spec   string
		wantKV uint32
		wantKC uint32
	}{
		{spec: "x", wantKV: 0x78, wantKC: tcKeyX},
		{spec: "A", wantKV: 0x41, wantKC: tcKeyA},
		{spec: "5", wantKV: 0x35, wantKC: tcKey1 + 4},
		{spec: " ", wantKV: 0x20, wantKC: tcKeySpace},
	} {
		t.Run(fmt.Sprintf("%q", tc.spec), func(t *testing.T) {
			kv, kc, mods, err := session.ParseProbeKey(tc.spec)
			if err != nil {
				t.Fatalf("ParseProbeKey(%q) error = %v, want nil", tc.spec, err)
			}
			if kv != tc.wantKV || kc != tc.wantKC || mods != 0 {
				t.Errorf("ParseProbeKey(%q) = (0x%x, %d, 0x%x), want (0x%x, %d, 0x0)",
					tc.spec, kv, kc, mods, tc.wantKV, tc.wantKC)
			}
		})
	}
}

// TestParseProbeKey_Rejects pins the refusal corpus (ASVS V5): every cause
// carries its own named error, checkable with errors.Is.
func TestParseProbeKey_Rejects(t *testing.T) {
	for _, tc := range []struct {
		spec    string
		wantErr error
	}{
		{spec: "", wantErr: session.ErrProbeEmptySpec},
		{spec: "Home End", wantErr: session.ErrProbeSpecFormat},
		{spec: "shift+", wantErr: session.ErrProbeSpecFormat},
		{spec: "+Home", wantErr: session.ErrProbeSpecFormat},
		{spec: "Frobnicate", wantErr: session.ErrProbeKeyUnknown},
		// "q+shift": the trailing modifier sits in the KEY slot and the
		// leading "q" in the MODIFIER slot — the parser names the offending
		// modifier-slot token (the refusal class of the nested-modifier form).
		{spec: "q+shift", wantErr: session.ErrProbeModUnknown},
		{spec: "hyper+Home", wantErr: session.ErrProbeModUnknown},
		{spec: "ы", wantErr: session.ErrProbeKeyUnmapped},
		{spec: "\t", wantErr: session.ErrProbeKeyUnmapped},
	} {
		t.Run(fmt.Sprintf("%q", tc.spec), func(t *testing.T) {
			kv, kc, mods, err := session.ParseProbeKey(tc.spec)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ParseProbeKey(%q) error = %v, want %v", tc.spec, err, tc.wantErr)
			}
			if kv != 0 || kc != 0 || mods != 0 {
				t.Errorf("ParseProbeKey(%q) returned values (0x%x, %d, 0x%x) with an error", tc.spec, kv, kc, mods)
			}
		})
	}
}

// TestProbeForward_EmitsPressRelease pins the emission shape: exactly one
// press/release pair through the emitter, the keycode by the forwardCtrlV
// physical-code convention, the release carrying MaskRelease.
func TestProbeForward_EmitsPressRelease(t *testing.T) {
	sink := &probeSink{}
	if err := session.ProbeForward(sink, "x"); err != nil {
		t.Fatalf("ProbeForward(x) error = %v, want nil", err)
	}
	events := sink.snapshot()
	if len(events) != 2 {
		t.Fatalf("ProbeForward(x) emitted %d events, want exactly 2", len(events))
	}
	want := []probeEvent{
		{keyval: 0x78, keycode: tcKeyX, state: 0},
		{keyval: 0x78, keycode: tcKeyX, state: engine.MaskRelease},
	}
	for i, w := range want {
		if events[i] != w {
			t.Errorf("event[%d] = %+v, want %+v", i, events[i], w)
		}
	}
}

// TestProbeForward_RejectsInvalidSpec pins the fail-before-emit order: an
// invalid keyspec errors and leaves the emitter untouched.
func TestProbeForward_RejectsInvalidSpec(t *testing.T) {
	sink := &probeSink{}
	if err := session.ProbeForward(sink, "Frobnicate"); err == nil {
		t.Fatal("ProbeForward(Frobnicate) error = nil, want the parse refusal")
	}
	if got := len(sink.snapshot()); got != 0 {
		t.Errorf("invalid spec emitted %d events, want 0", got)
	}
}

// TestProbeForward_Logged pins the emission journal: a structured
// "probe forward" record per event (press + release) carrying keyval, mods
// and the release flag — never the keyspec content itself.
func TestProbeForward_Logged(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	sink := &probeSink{}
	if err := session.ProbeForward(sink, "x"); err != nil {
		t.Fatalf("ProbeForward(x) error = %v, want nil", err)
	}
	out := buf.String()
	if got := strings.Count(out, `"msg":"probe forward"`); got != 2 {
		t.Errorf("log carries %d probe forward records, want 2 (press + release): %s", got, out)
	}
	for _, want := range []string{`"keyval":"0x78"`, `"release":false`, `"release":true`} {
		if !strings.Contains(out, want) {
			t.Errorf("probe forward log missing %s: %s", want, out)
		}
	}
	if strings.Contains(out, `"spec"`) {
		t.Errorf("probe forward log leaks the keyspec argument: %s", out)
	}
}

// The delivery-probe corpus (plan 10-13): ProbeDeleteSurrounding must emit
// EXACTLY DeleteSurroundingText(offset, nchars) then RequireSurroundingText
// (the ordered IM channel the routing verdict rides), ProbeCommit exactly
// one CommitText — both fail-closed without an engine emitter, both
// journalled at DEBUG without the field content.

// probeOpSink is the delivery-probe emitter double: every emission recorded
// IN ORDER under a mutex — the op log pins the DST→Require order.
type probeOpSink struct {
	mu      sync.Mutex
	ops     []string
	commits []string
}

func (p *probeOpSink) ForwardKeyEvent(keyval, keycode, state uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ops = append(p.ops, fmt.Sprintf("forward(0x%x,%d,0x%x)", keyval, keycode, state))
}

func (p *probeOpSink) RequireSurroundingText() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ops = append(p.ops, "require")
}

func (p *probeOpSink) DeleteSurroundingText(offset int32, nchars uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ops = append(p.ops, fmt.Sprintf("delete(%d,%d)", offset, nchars))
}

func (p *probeOpSink) CommitText(text engine.IBusText) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ops = append(p.ops, "commit")
	p.commits = append(p.commits, text.Text)
}

func (p *probeOpSink) UpdateModeSymbol(string) {}

// opsSnapshot returns the ordered op log and the committed texts.
func (p *probeOpSink) opsSnapshot() ([]string, []string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return append([]string(nil), p.ops...), append([]string(nil), p.commits...)
}

// TestProbeDeleteSurrounding_NilEngineFailClosed pins the fail-closed half:
// without an emitter the probe errors (the named ErrProbeNoEngine) and can
// emit nothing.
func TestProbeDeleteSurrounding_NilEngineFailClosed(t *testing.T) {
	if err := session.ProbeDeleteSurrounding(nil, 0, 6); !errors.Is(err, session.ErrProbeNoEngine) {
		t.Fatalf("ProbeDeleteSurrounding(nil) error = %v, want ErrProbeNoEngine", err)
	}
}

// TestProbeDeleteSurrounding_EmitsDeleteThenRequire pins the emission
// order: exactly DeleteSurroundingText(offset, nchars) followed by
// RequireSurroundingText — nothing else, never reordered.
func TestProbeDeleteSurrounding_EmitsDeleteThenRequire(t *testing.T) {
	sink := &probeOpSink{}
	if err := session.ProbeDeleteSurrounding(sink, 0, 6); err != nil {
		t.Fatalf("ProbeDeleteSurrounding(0, 6) error = %v, want nil", err)
	}
	ops, _ := sink.opsSnapshot()
	want := []string{"delete(0,6)", "require"}
	if !slices.Equal(ops, want) {
		t.Fatalf("ops = %v, want exactly %v (DST before Require)", ops, want)
	}
}

// TestProbeCommit_NilEngineFailClosed pins the commit probe's fail-closed
// half.
func TestProbeCommit_NilEngineFailClosed(t *testing.T) {
	if err := session.ProbeCommit(nil, "x"); !errors.Is(err, session.ErrProbeNoEngine) {
		t.Fatalf("ProbeCommit(nil) error = %v, want ErrProbeNoEngine", err)
	}
}

// TestProbeCommit_EmitsOneCommit pins the commit shape: exactly one
// CommitText carrying the text, no other emission.
func TestProbeCommit_EmitsOneCommit(t *testing.T) {
	sink := &probeOpSink{}
	if err := session.ProbeCommit(sink, "привет"); err != nil {
		t.Fatalf("ProbeCommit error = %v, want nil", err)
	}
	ops, commits := sink.opsSnapshot()
	if !slices.Equal(ops, []string{"commit"}) {
		t.Fatalf("ops = %v, want exactly [commit]", ops)
	}
	if len(commits) != 1 || commits[0] != "привет" {
		t.Fatalf("commits = %q, want exactly one привет", commits)
	}
}

// TestProbeVerbs_Logged pins the DEBUG journal of both probes: parameters
// and rune counts only — never the committed text (the «passwords become
// visible» discipline of the probe family).
func TestProbeVerbs_Logged(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	sink := &probeOpSink{}
	if err := session.ProbeDeleteSurrounding(sink, 0, 6); err != nil {
		t.Fatalf("ProbeDeleteSurrounding error = %v, want nil", err)
	}
	if err := session.ProbeCommit(sink, "привет"); err != nil {
		t.Fatalf("ProbeCommit error = %v, want nil", err)
	}
	out := buf.String()
	for _, want := range []string{
		`"msg":"probe delete-surrounding"`, `"offset":0`, `"nchars":6`,
		`"msg":"probe commit"`, `"runes":6`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("probe log missing %s: %s", want, out)
		}
	}
	if strings.Contains(out, "привет") {
		t.Errorf("probe log leaks the committed text: %s", out)
	}
}
