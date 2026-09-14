// Package session hosts the daemon's event consumers: the actor that
// serializes engine events into the hotkey FSM, feeds the typed-phrase
// buffer and runs the two-phase word correction (CORR-01).
package session

import (
	"log/slog"
	"sync"
	"time"
	"unicode"

	"github.com/Djarvur/goswitch/engine"
	"github.com/Djarvur/goswitch/internal/correct"
	"github.com/Djarvur/goswitch/internal/hotkey"
	"github.com/Djarvur/goswitch/layouts"
)

// verifyWait is the ADR-004 budget for a fresh SetSurroundingText to arrive
// after RequireSurroundingText (resolved planning question 4: ~100 ms — the
// local D-Bus RTT is 1-10 ms, a slow client gets the rest; past the budget
// the correction is silently dropped, never guessed).
const verifyWait = 100 * time.Millisecond

// comboMask isolates the true key combinations whose characters never
// reach the field (Ctrl shortcuts, Alt menus, Super shell keys): a press
// carrying any of them must not feed the buffer and must never be consumed.
// Latch-state modifiers (CapsLock, NumLock) ride along with every key on a
// lit desktop and stay allowed — the keyval is already the XKB-translated
// character, so it keeps mirroring the field (live finding of the tracer
// run, 2026-09-14: every injected letter arrived with mods 0x10 — NumLock —
// and a literal Shift-only guard starved the buffer). The same guard governs
// RU-mode consumption: the plan's literal mods&^MaskShift==0 is superseded
// by this live-proven mask, or every NumLock-lit keystroke would transit
// Latin and the flip would be dead on the real desktop.
const comboMask = engine.MaskControl | engine.MaskMod1 | engine.MaskMod4

// scriptMode is the daemon's output-script state (ADR-001 Option B): the
// flip toggles it on every Single decision while the session's XKB group
// stays untouched — the mode is pure daemon state, the rune the field
// receives comes from the engine's own commits in RU mode. EN is the start
// state.
type scriptMode int

// Script modes of the flip.
const (
	modeEN scriptMode = iota
	modeRU
)

// Actor implements engine.EventHandler for the daemon: every key and
// lifecycle event is serialized through one mutex into the pure hotkey FSM,
// printable presses feed the typed-phrase buffer (CORR-09 resets on
// FocusOut/Reset), and the tap-window deadline is carried by a single
// time.AfterFunc timer that re-enters the FSM at expiry. A Double decision
// starts the ADR-004 verification: the correction range (token+tail) is
// checked against the freshest surrounding text the client reported — the
// cached spontaneous push, or a fresh answer to RequireSurroundingText
// within verifyWait for clients that implement the round trip. The
// ProcessKeyEvent answer never waits (ADR-004, Pitfall 4). A Single
// decision flips the internal script mode (ADR-001 Option B, plan 02-04):
// in RU mode clean printable presses are consumed and their Cyrillic runes
// committed through layouts.ENToRU, in EN mode everything transits as in
// Phase 1.
//
// godbus dispatches every D-Bus method call on its own goroutine, so the
// mutex is the actor's single entry point: without it the FSM state would
// race between concurrent ProcessKeyEvent calls.
type Actor struct {
	mu      sync.Mutex
	fsm     *hotkey.FSM
	window  time.Duration
	start   time.Time
	timer   *time.Timer
	buf     *correct.Buffer
	caps    uint32
	eng     engine.Emitter
	surr    []rune // cached text-before-cursor from the latest client push
	pending *pendingFix
	mode    scriptMode // output-script state, EN at start (ADR-001 Option B)
}

// pendingFix is the state of a correction between the Double decision and
// the surrounding-text verdict: what to replace (token, its boundary tail),
// what to commit (the converted token), the range the verification must see
// at the end of the text (token+tail — exactly what the ladder deletes) and
// the verify deadline.
type pendingFix struct {
	token     []rune
	tail      []rune
	converted []rune
	match     []rune
	armed     time.Time
	deadline  *time.Timer
}

// NewActor returns an actor deciding tap series inside the given
// disambiguation window (D-05). The daemon passes hotkey.DefaultWindow.
func NewActor(window time.Duration) *Actor {
	return &Actor{
		fsm:    hotkey.NewFSM(window),
		window: window,
		start:  time.Now(),
		buf:    correct.NewBuffer(),
	}
}

// HandleKey implements engine.EventHandler: the decoded event is fed into
// the FSM under the mutex and a Shift_R release re-arms the deadline timer.
// Decisions never fire here — only at window expiry (D-04). The returned
// verdict is the consumption decision of the script mode: in RU mode a
// clean printable press whose key maps to a different rune is consumed
// after committing the Cyrillic rune (the owner-prototype pattern,
// punto_engine.py:394-414); everything else — EN mode, identical mappings,
// unmapped keys, combos, bare modifiers — transits. The buffer is fed
// script-true in EVERY branch that puts a rune in the field (plan 02-04,
// A6): the committed Cyrillic rune on a RU commit, the original keyval on
// every transit the client will insert. A Backspace press pops the buffer
// honestly, so it keeps mirroring the field (Pitfall 7).
func (a *Actor) HandleKey(ev engine.EngineEvent) (consume bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if ev.Release {
		a.fsm.Feed(hotkey.KeyRelease{Keyval: ev.Keyval}, a.elapsed())
		if ev.Keyval == hotkey.KeyvalShiftR {
			// Only a Shift_R release can move the series deadline (the FSM
			// counts taps on clean releases); re-arming on any other event
			// would either extend the deadline from the wrong instant or
			// arm a timer over a cancelled series. A stale expiry is a
			// no-op in the FSM, so over-arming is harmless, under-arming
			// would silently drop the decision.
			a.armTimer()
		}

		return false
	}
	a.fsm.Feed(hotkey.KeyPress{Keyval: ev.Keyval}, a.elapsed())

	return a.feedKey(ev)
}

// HandleLifecycle implements engine.EventHandler: FocusOut and Reset disarm
// a pending series (the input context is gone — a decision there would be
// garbage) and hard-reset the phrase buffer (CORR-09); everything else is a
// DEBUG trace.
func (a *Actor) HandleLifecycle(kind engine.LifecycleKind) {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch kind {
	case engine.LifecycleFocusOut, engine.LifecycleReset:
		a.fsm.Feed(hotkey.Reset{}, a.elapsed())
		a.buf.HardReset()
		a.surr = nil // the cache belongs to the input context that just left
		if a.timer != nil {
			a.timer.Stop()
			a.timer = nil
		}
	case engine.LifecycleFocusIn, engine.LifecycleEnable, engine.LifecycleDisable:
		slog.Debug("lifecycle", "kind", kind.String())
	}
}

// HandleSurroundingText implements engine.EventHandler: the arriving text
// updates the surrounding cache, and when a correction is pending it also
// settles it (ADR-004). The runes before the cursor must END with the
// correction range (token+tail — exactly what the ladder deletes, CORR-07);
// a match executes the plan, a mismatch aborts silently ("abort, не
// мусорить": not one character is touched).
func (a *Actor) HandleSurroundingText(text string, cursorPos uint32) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.surr = beforeCursor(text, cursorPos)
	if a.pending == nil {
		return
	}
	if !correct.MatchesSuffix(a.surr, a.pending.match) {
		a.resolvePending()
		slog.Info("correction skipped", "reason", "verify-mismatch")

		return
	}
	a.executeCorrection()
}

// HandleCapabilities implements engine.EventHandler: the capability bitmap
// of the current input context — the input of the ADR-003 ladder choice.
func (a *Actor) HandleCapabilities(caps uint32) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.caps = caps
}

// AttachEngine implements engine.EventHandler: the emitter sink of the
// engine minted for the input context.
func (a *Actor) AttachEngine(eng engine.Emitter) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.eng = eng
}

// Expiry is the timer callback: time.AfterFunc(window) re-enters here when
// the disambiguation window closes.
func (a *Actor) Expiry() {
	a.ExpiryAt(a.elapsed())
}

// ExpiryAt feeds the FSM a window expiry at the given logical time, logs
// every decision as {"msg":"action","n":N} — the e2e stand greps this exact
// shape — and dispatches it: Double starts the correction pipeline, Single
// flips the script mode (plan 02-04), Triple stays deferred to the phrase
// (02-05). Tests inject the logical time directly (deterministic expiry);
// the daemon path always goes through Expiry's real clock.
func (a *Actor) ExpiryAt(now time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.timer = nil
	for _, action := range a.fsm.Feed(hotkey.TimerExpired{}, now) {
		slog.Info("action", "n", int(action))
		switch action {
		case hotkey.Double:
			a.startCorrection()
		case hotkey.Single:
			a.flipScript()
		case hotkey.Triple:
			slog.Debug("action deferred", "n", int(action)) // 02-05 phrase
		}
	}
}

// VerifyExpiry is the verify-deadline timer callback: the verifyWait budget
// for a fresh SetSurroundingText closed without an answer — the correction
// is dropped silently (ADR-004, Pitfall 4: the wait lives in a timer, never
// in a handler). Tests inject it directly (deterministic deadline).
func (a *Actor) VerifyExpiry() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.pending == nil {
		return // already settled — a stale timer is a no-op
	}
	a.resolvePending()
	slog.Info("correction skipped", "reason", "verify-timeout")
}

// flipScript toggles the internal output-script mode (ADR-001 Option B):
// the switch is daemon state only — the session's XKB group is never
// touched and the flip itself emits nothing on the sink. The INFO mode
// record is the e2e sequencing contract: the stand waits for it after a
// single tap before typing in the new script, because the flip fires at
// window expiry, not inside the tap (Pitfall 5). The caller holds the
// mutex.
func (a *Actor) flipScript() {
	if a.mode == modeEN {
		a.mode = modeRU
		slog.Info("mode", "to", "ru")

		return
	}
	a.mode = modeEN
	slog.Info("mode", "to", "en")
}

// feedKey decides one press: whether the engine consumes the key and which
// rune the buffer takes — the script-true invariant made branch-local (a
// rune that reaches the field also reaches the buffer, whatever delivered
// it). The caller holds the mutex.
func (a *Actor) feedKey(ev engine.EngineEvent) bool {
	switch {
	case ev.Keyval == engine.KeyBackSpace:
		a.buf.Backspace()

		return false
	case !printableKeyval(ev.Keyval) || ev.Mods&comboMask != 0:
		// Non-printables (bare modifiers included) and true combos put no
		// rune in the field — the buffer must not see them either.
		return false
	case a.mode == modeRU:
		// #nosec G115 -- printableKeyval bounds the keyval below
		// 0xFE00, so the uint32→rune conversion cannot overflow.
		r := rune(ev.Keyval)
		if ru, ok := layouts.ENToRU[r]; ok && ru != r && a.eng != nil {
			a.eng.CommitText(engine.NewIBusText(string(ru)))
			a.buf.Push(ru) // the committed rune is what the field now holds

			return true
		}
		// Identical map (digits, parentheses…) or unmapped key: transit —
		// the client inserts the original rune, so that rune is the field
		// truth the buffer must mirror.
		a.buf.Push(r)

		return false
	default:
		// EN mode: Phase 1 semantics — transit, buffer fed as typed.
		// #nosec G115 -- printableKeyval bounds the keyval below
		// 0xFE00, so the uint32→rune conversion cannot overflow.
		a.buf.Push(rune(ev.Keyval))

		return false
	}
}

// startCorrection launches a correction: classify the token, convert it,
// choose the ladder level by the caps bit, then verify against the freshest
// surrounding text (ADR-004). Every refusal logs its D-20 reason at INFO —
// without the word's contents — and touches nothing. The caller holds the
// mutex.
//
// Transport adaptation (live finding, 2026-09-14): neither GTK nor the
// mutter input context answers RequireSurroundingText — clients push
// SetSurroundingText spontaneously on every text change (the pushes at
// cursor 1..6 during typing), exactly how the owner's prototype works.
// The freshest spontaneous push is therefore checked first (a push after
// the last keystroke is the freshest state the client can report); only on
// a miss does the Require round-trip run inside the verifyWait budget.
func (a *Actor) startCorrection() {
	token := a.buf.Token()
	if len(token) == 0 {
		slog.Info("correction skipped", "reason", "empty-buffer")

		return
	}
	dir, ok := correct.Detect(token)
	if !ok {
		slog.Info("correction skipped", "reason", tokenRefusal(token))

		return
	}
	converted, ok := correct.Convert(token, dir)
	if !ok {
		slog.Info("correction skipped", "reason", "convert-failed")

		return
	}
	if a.eng == nil {
		slog.Info("correction skipped", "reason", "no-engine")

		return
	}
	if a.caps&correct.CapSurroundingText == 0 {
		// No surrounding text, no verification — level 2 arrives in plan
		// 02-05; until then the correction is refused, never guessed.
		slog.Info("correction skipped", "reason", "no-surrounding")

		return
	}
	tail := a.buf.Tail()
	a.resolvePending() // a second Double supersedes the stale round
	a.pending = &pendingFix{
		token:     token,
		tail:      tail,
		converted: converted,
		match:     concatRunes(token, tail),
		armed:     time.Now(),
	}
	if correct.MatchesSuffix(a.surr, a.pending.match) {
		a.executeCorrection() // cached push is the freshest report

		return
	}
	a.pending.deadline = time.AfterFunc(verifyWait, a.VerifyExpiry)
	a.eng.RequireSurroundingText()
}

// executeCorrection runs the ladder plan of the settled pending fix — one
// DeleteSurroundingText exactly over the token+tail range, one commit of
// the converted token plus the tail (CORR-07) — and replaces the token in
// the buffer so a repeated correction converts back. The caller holds the
// mutex and pending != nil.
func (a *Actor) executeCorrection() {
	p := a.pending
	a.pending = nil
	if p.deadline != nil {
		p.deadline.Stop()
	}
	plan := correct.BuildPlan(p.token, p.tail, p.converted, a.caps)
	a.eng.DeleteSurroundingText(plan.Offset, plan.NChars)
	a.eng.CommitText(engine.NewIBusText(string(plan.Commit)))
	a.buf.ReplaceToken(p.converted) // the buffer keeps mirroring the field — repeat converts back
	slog.Info("correction", "outcome", "done")
	slog.Debug("correction",
		"level", int(plan.Level), // first attribute after msg — the matrix greps this exact form (D-21)
		"runes", len(p.token),
		"source", string(p.token),
		"result", string(p.converted),
		"latency_ms", time.Since(p.armed).Milliseconds())
}

// resolvePending retires the pending fix together with its deadline timer;
// the caller holds the mutex.
func (a *Actor) resolvePending() {
	if a.pending == nil {
		return
	}
	if a.pending.deadline != nil {
		a.pending.deadline.Stop()
	}
	a.pending = nil
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

// printableKeyval reports whether the keyval carries a printable character
// (the Unicode range below the keypad/function block): presses in that
// range land in the field as characters.
func printableKeyval(keyval uint32) bool {
	return keyval >= 0x20 && keyval < 0xFE00
}

// beforeCursor returns the rune slice of text up to the cursor position,
// clamped to the text length.
func beforeCursor(text string, cursorPos uint32) []rune {
	runes := []rune(text)
	// #nosec G115 -- len(runes) is bounded by a real input field, far
	// below 2^31 runes; on the 64-bit target an int always holds a uint32.
	if int(cursorPos) < len(runes) {
		return runes[:int(cursorPos)]
	}

	return runes
}

// concatRunes joins two rune slices.
func concatRunes(head, tail []rune) []rune {
	joined := make([]rune, 0, len(head)+len(tail))
	joined = append(joined, head...)

	return append(joined, tail...)
}

// tokenRefusal labels a direction-less token for the D-20 skip log: letters
// of both scripts are a mixed word (D-16), anything else has no letters at
// all. A diagnostic label only — the refusal decision itself belongs to
// correct.Detect and is not duplicated here.
func tokenRefusal(token []rune) string {
	var latin, cyrillic bool
	for _, r := range token {
		switch {
		case unicode.Is(unicode.Latin, r):
			latin = true
		case unicode.Is(unicode.Cyrillic, r):
			cyrillic = true
		}
	}
	if latin && cyrillic {
		return "mixed-script"
	}

	return "no-letters"
}
