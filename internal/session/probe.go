package session

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Djarvur/goswitch/internal/engine"
)

// Probe keyspec refusals — the named errors of the strict keyspec grammar
// (plan 09-02, ASVS V5: every cause distinct and errors.Is-checkable).
var (
	ErrProbeEmptySpec   = errors.New("probe keyspec is empty")
	ErrProbeSpecFormat  = errors.New("probe keyspec must be [mod+]{Home|End|Left|Right|BackSpace} or one key")
	ErrProbeKeyUnknown  = errors.New("probe key unknown")
	ErrProbeModUnknown  = errors.New("probe modifier unknown")
	ErrProbeKeyUnmapped = errors.New("probe key has no physical keycode mapping")
)

// probeKeysyms maps the named keyspec keys to their engine keyvals.
//
//nolint:gochecknoglobals // the keyspec grammar's static lookup tables (read-only)
var probeKeysyms = map[string]uint32{
	"home":      engine.KeyHome,
	"end":       engine.KeyEnd,
	"left":      engine.KeyLeft,
	"right":     engine.KeyRight,
	"backspace": engine.KeyBackSpace, // the drain probe's replay key (plan 10-13 P3)
}

// probeModifiers maps the keyspec modifier tokens to their IBus masks.
//
//nolint:gochecknoglobals // the keyspec grammar's static lookup tables (read-only)
var probeModifiers = map[string]uint32{
	"shift": engine.MaskShift,
	"ctrl":  engine.MaskControl,
}

// Physical (evdev) keycodes of the probe bursts — the forwardCtrlV keycode
// discipline (KEY_LEFTCTRL=29, KEY_V=47, linux/event-codes.h) continued:
// every value below is transcribed from
// /usr/include/linux/input-event-codes.h verified verbatim (never from
// memory): KEY_HOME=102:177, KEY_END=107:182, KEY_LEFT=105:180,
// KEY_RIGHT=106:181, KEY_1=2:77, KEY_0=11:86, KEY_SPACE=57:132. The letter
// map below is positional (keyboard rows Q..P = 16..25, A..L = 30..38,
// Z..M = 44..50 — lines 91-125 of the header), NOT alphabetical: the
// corpus caught the +offset shortcut live (KEY_X=45, not 30+23).
const (
	probeKeycodeHome      = 102
	probeKeycodeEnd       = 107
	probeKeycodeLeft      = 105
	probeKeycodeRight     = 106
	probeKeycode1         = 2
	probeKeycode0         = 11
	probeKeycodeSpace     = 57
	probeKeycodeBackspace = 14 // KEY_BACKSPACE=14 — the level-2 replay code (input-event-codes.h; backSpaceKeycode)
)

// probeLetterKeycodes is the physical evdev code of every letter key on the
// us layout (input-event-codes.h lines 91-125, transcribed verbatim — see
// the const block above for the never-from-memory rule). The literals ARE
// the transcription, cited once here; they are not magic numbers to hide.
//
//nolint:gochecknoglobals,mnd // the transcribed header table, cited verbatim (read-only)
var probeLetterKeycodes = map[rune]uint32{
	'q': 16, 'w': 17, 'e': 18, 'r': 19, 't': 20, 'y': 21, 'u': 22, 'i': 23, 'o': 24, 'p': 25,
	'a': 30, 's': 31, 'd': 32, 'f': 33, 'g': 34, 'h': 35, 'j': 36, 'k': 37, 'l': 38,
	'z': 44, 'x': 45, 'c': 46, 'v': 47, 'b': 48, 'n': 49, 'm': 50,
}

// ParseProbeKey parses the probe verb's keyspec — the strict grammar
// [mod+]{Home|End|Left|Right} or exactly one printable key with optional
// modifiers (plan 09-02): modifiers are shift/ctrl, case-insensitive; named
// keys fold case; a printable key is one ASCII letter, digit or space — the
// classes with a physical evdev keycode (the burst answers for its
// modifiers explicitly, so "A" carries NO auto-Shift). Every refusal is a
// named, errors.Is-checkable error (ASVS V5).
func ParseProbeKey(spec string) (keyval, keycode, mods uint32, err error) {
	if spec == "" {
		return 0, 0, 0, ErrProbeEmptySpec
	}
	tokens := strings.Split(spec, "+")
	for _, tok := range tokens[:len(tokens)-1] {
		if tok == "" { // "shift++Home" / "+Home": an empty modifier slot
			return 0, 0, 0, ErrProbeSpecFormat
		}
		mask, ok := probeModifiers[strings.ToLower(tok)]
		if !ok {
			return 0, 0, 0, fmt.Errorf("%w: %q", ErrProbeModUnknown, tok)
		}
		mods |= mask
	}
	key := tokens[len(tokens)-1]
	// A multi-word token ("Home End") is a form error; the single space
	// token is the space KEY itself, so the check needs both words AND a
	// space inside the token.
	if key == "" || (len([]rune(key)) > 1 && strings.ContainsRune(key, ' ')) {
		return 0, 0, 0, ErrProbeSpecFormat
	}
	if kv, ok := probeKeysyms[strings.ToLower(key)]; ok {
		return kv, probeKeysymKeycode(kv), mods, nil
	}
	runes := []rune(key)
	if len(runes) != 1 {
		return 0, 0, 0, fmt.Errorf("%w: %q", ErrProbeKeyUnknown, key)
	}

	kc, ok := probePrintableKeycode(runes[0])
	if !ok {
		return 0, 0, 0, fmt.Errorf("%w: %q", ErrProbeKeyUnmapped, key)
	}

	return uint32(runes[0]), kc, mods, nil
}

// probeKeysymKeycode returns the physical keycode of a named probe key.
func probeKeysymKeycode(keyval uint32) uint32 {
	switch keyval {
	case engine.KeyHome:
		return probeKeycodeHome
	case engine.KeyEnd:
		return probeKeycodeEnd
	case engine.KeyLeft:
		return probeKeycodeLeft
	case engine.KeyRight:
		return probeKeycodeRight
	case engine.KeyBackSpace:
		return probeKeycodeBackspace
	}

	return 0
}

// probePrintableKeycode maps one printable rune to its physical evdev
// keycode: letters (the positional map above — case-folded, the keycode is
// the key, not the shifted glyph), digits (KEY_1 base, KEY_0 capping the
// row) and space. Everything else — punctuation and every non-ASCII or
// non-printable rune — has no mapped physical key on the probe's us layout
// and is refused.
func probePrintableKeycode(r rune) (uint32, bool) {
	if r >= 'A' && r <= 'Z' {
		r += 'a' - 'A'
	}
	if kc, ok := probeLetterKeycodes[r]; ok {
		return kc, true
	}
	switch {
	case r >= '1' && r <= '9':
		return probeKeycode1 + uint32(r-'1'), true
	case r == '0':
		return probeKeycode0, true
	case r == ' ':
		return probeKeycodeSpace, true
	}

	return 0, false
}

// ProbeForward emits the synthetic press/release pair for spec through the
// live engine emitter — the daemon channel of the select-chord probes
// (plan 09-02): ForwardKeyEvent is a signal only the engine's own
// connection can emit, no external tool substitutes. The shape mirrors
// forwardCtrlV's pair discipline (press with the event mods, release with
// MaskRelease); each emission is journalled at DEBUG — the -debug key-trace
// class («passwords become visible»), never the keyspec argument itself.
// Without an emitter the probe fails closed (WR-07 of the 10-REVIEW) — the
// family contract of the other four verbs; ProbeForward was the only one
// without the guard and panicked on the nil interface instead.
func ProbeForward(eng engine.Emitter, spec string) error {
	keyval, keycode, mods, err := ParseProbeKey(spec)
	if err != nil {
		return err
	}
	if eng == nil {
		return ErrProbeNoEngine
	}
	eng.ForwardKeyEvent(keyval, keycode, mods)
	slog.Debug("probe forward",
		"keyval", fmt.Sprintf("0x%x", keyval), "keycode", keycode,
		"mods", fmt.Sprintf("0x%x", mods), "release", false)
	eng.ForwardKeyEvent(keyval, keycode, mods|engine.MaskRelease)
	slog.Debug("probe forward",
		"keyval", fmt.Sprintf("0x%x", keyval), "keycode", keycode,
		"mods", fmt.Sprintf("0x%x", mods|engine.MaskRelease), "release", true)

	return nil
}

// ErrProbeNoEngine is the delivery probes' fail-closed state name (plan
// 10-13): the actor has no attached input-context emitter yet — nothing can
// be emitted.
var ErrProbeNoEngine = errors.New("probe has no engine (no input context attached)")

// ProbeDeleteSurrounding emits the delivery probe's erase-then-ask pair
// (plan 10-13): exactly DeleteSurroundingText(offset, nchars) followed by
// RequireSurroundingText on the live engine emitter — the single ordered
// IM channel the round-3 delivery-capability probe measures (no external
// tool can substitute: the signals ride the engine's own connection). One
// DEBUG row per call carries the parameters and the outcome — never the
// field content (the «passwords become visible» class of the probe family;
// the content lives only in the stand's pty oracle). Without an emitter
// (no input context minted yet) the probe fails closed.
func ProbeDeleteSurrounding(eng engine.Emitter, offset int32, nchars uint32) error {
	if eng == nil {
		return ErrProbeNoEngine
	}
	eng.DeleteSurroundingText(offset, nchars)
	eng.RequireSurroundingText()
	slog.Debug("probe delete-surrounding", "offset", offset, "nchars", nchars)

	return nil
}

// ProbeDelete emits the crash-bisection probe's DST-only half (the
// 2026-10-16 stand finding: gnome-shell dies on the paired emission — the
// bisection needs the signals separable): exactly
// DeleteSurroundingText(offset, nchars), NO RequireSurroundingText. The
// fail-closed and content disciplines are ProbeDeleteSurrounding's
// verbatim.
func ProbeDelete(eng engine.Emitter, offset int32, nchars uint32) error {
	if eng == nil {
		return ErrProbeNoEngine
	}
	eng.DeleteSurroundingText(offset, nchars)
	slog.Debug("probe delete", "offset", offset, "nchars", nchars)

	return nil
}

// ProbeRequire emits the crash-bisection probe's Require-only half: exactly
// RequireSurroundingText — the surrounding-text request alone, no delete
// riding it. The fail-closed and content disciplines are
// ProbeDeleteSurrounding's verbatim.
func ProbeRequire(eng engine.Emitter) error {
	if eng == nil {
		return ErrProbeNoEngine
	}
	eng.RequireSurroundingText()
	slog.Debug("probe require")

	return nil
}

// ProbeCommit emits the delivery probe's commit (plan 10-13): one
// CommitText carrying the text on the live engine emitter. The DEBUG row
// carries the rune count, never the text (D-20/D-21 — same content
// discipline as ProbeDeleteSurrounding). Without an emitter the probe
// fails closed.
func ProbeCommit(eng engine.Emitter, text string) error {
	if eng == nil {
		return ErrProbeNoEngine
	}
	eng.CommitText(engine.NewIBusText(text))
	slog.Debug("probe commit", "runes", len([]rune(text)))

	return nil
}

// ProbeEmitter returns the engine emitter of the currently attached input
// context — the probe verb's sink (plan 09-02): nil before the first mint,
// exactly like emitter().
//
//nolint:ireturn // the probe seam hands the emitter interface back (the emitter() precedent)
func (a *Actor) ProbeEmitter() engine.Emitter {
	if slot := a.engSlot.Load(); slot != nil {
		return slot.eng
	}

	return nil
}
