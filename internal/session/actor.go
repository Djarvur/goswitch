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

// backSpaceKeycode is the physical keycode of the Backspace key — the
// keycode half of the level-2 replay ForwardKeyEvent(KeyBackSpace, 14, 0)
// (RESEARCH Pattern 1; the keyval half is engine.KeyBackSpace).
const backSpaceKeycode = 14

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
	mu          sync.Mutex
	fsm         *hotkey.FSM
	window      time.Duration
	start       time.Time
	timer       *time.Timer
	buf         *correct.Buffer
	caps        uint32
	eng         engine.Emitter
	surr        []rune // cached text-before-cursor from the latest client push
	pending     *pendingFix
	after       *pendingAfter
	verifyEpoch uint64     // monotonic verify-after round tag (stale-timer guard)
	mode        scriptMode // output-script state, EN at start (ADR-001 Option B)
}

// correctionRange parameterizes the correction pipeline by its range (D-23 —
// one pipeline, different ranges): the word path (Double) supplies the token,
// its boundary tail and ReplaceToken; the phrase path (Triple) supplies the
// whole phrase since the last hard reset, no tail and ReplacePhrase (D-25).
// The selection branch of later plans joins the same seam.
type correctionRange struct {
	token   []rune
	tail    []rune
	replace func(converted []rune)
}

// pendingFix is the state of a correction between the Double/Triple decision
// and the surrounding-text verdict: the range under correction (token, its
// boundary tail), what to commit (the converted token), the range the
// verification must see at the end of the text (token+tail — exactly what the
// ladder deletes) and the verify deadline.
type pendingFix struct {
	rng       correctionRange
	converted []rune
	match     []rune
	armed     time.Time
	deadline  *time.Timer
}

// pendingAfter is the state of the verify-after round — the post-correction
// check that compensates the ack-less DeleteSurroundingText (ADR-003): the
// suffix the client's fresh surrounding text must END with (converted+tail)
// and the round's epoch. The epoch makes a stale deadline from a superseded
// round a no-op: the timer callback re-checks the tag under the mutex.
type pendingAfter struct {
	expected []rune
	epoch    uint64
	deadline *time.Timer
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
// garbage), hard-reset the phrase buffer (CORR-09) and retire any open
// correction round — its verify answer belongs to an input context that no
// longer exists; everything else is a DEBUG trace.
func (a *Actor) HandleLifecycle(kind engine.LifecycleKind) {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch kind {
	case engine.LifecycleFocusOut, engine.LifecycleReset:
		a.fsm.Feed(hotkey.Reset{}, a.elapsed())
		a.buf.HardReset()
		a.surr = nil // the cache belongs to the input context that just left
		a.resolvePending()
		a.clearAfter()
		if a.timer != nil {
			a.timer.Stop()
			a.timer = nil
		}
	case engine.LifecycleFocusIn, engine.LifecycleEnable, engine.LifecycleDisable:
		slog.Debug("lifecycle", "kind", kind.String())
	}
}

// HandleSurroundingText implements engine.EventHandler: the arriving text
// updates the surrounding cache and settles whichever round is open. A
// pending correction (ADR-004 pre-check) executes only when the runes before
// the cursor END with the correction range (token+tail — exactly what the
// ladder deletes, CORR-07); otherwise it aborts silently ("abort, не
// мусорить": not one character is touched). A pending verify-after
// (ADR-003/ADR-004 post-check) compares the suffix against the replacement
// the correction left behind: a mismatch — the client ignored the deletion,
// Chromium ibus#2354, Pitfall 3 — counts as an INFO record and is NEVER
// followed by an automatic repair.
func (a *Actor) HandleSurroundingText(text string, cursorPos uint32) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.surr = beforeCursor(text, cursorPos)
	if a.pending != nil {
		if !correct.MatchesSuffix(a.surr, a.pending.match) {
			a.resolvePending()
			slog.Info("correction skipped", "reason", "verify-mismatch")

			return
		}
		a.executeCorrection()

		return
	}
	if a.after != nil {
		expected := a.after.expected
		a.clearAfter()
		if !correct.MatchesSuffix(a.surr, expected) {
			slog.Info("correction verify", "outcome", "mismatch")

			return
		}
		slog.Debug("correction verify", "outcome", "match")
	}
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
// shape — and dispatches it: Double starts the word-correction pipeline,
// Single flips the script mode (plan 02-04), Triple starts the same pipeline
// over the whole phrase (plan 03-01, CORR-02). Tests inject the logical time
// directly (deterministic expiry); the daemon path always goes through
// Expiry's real clock.
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
			a.startPhraseCorrection()
		}
	}
}

// VerifyExpiry is the verify-deadline timer callback of the pre-correction
// round: the verifyWait budget for a fresh SetSurroundingText closed without
// an answer — the correction is dropped silently (ADR-004, Pitfall 4: the
// wait lives in a timer, never in a handler). The timeout also retires any
// pendingAfter: a round that never executed leaves its would-be post-check
// moot. Tests inject it directly (deterministic deadline).
func (a *Actor) VerifyExpiry() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.pending == nil {
		return // already settled — a stale timer is a no-op
	}
	a.resolvePending()
	a.clearAfter()
	slog.Info("correction skipped", "reason", "verify-timeout")
}

// afterExpiry is the verify-after deadline: the client never answered the
// post-correction Require, so the check is dropped quietly — the
// compensation is best-effort and never a repair (ADR-003 residual risk).
// The epoch tag makes a deadline from a superseded round a no-op.
func (a *Actor) afterExpiry(epoch uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.after == nil || a.after.epoch != epoch {
		return // stale deadline — a newer round owns the state
	}
	a.clearAfter()
	slog.Debug("correction verify", "outcome", "timeout")
}

// armAfterVerify starts the verify-after round (the caller holds the
// mutex): the actor itself asks for a fresh surrounding text and expects a
// suffix of converted+tail — the replacement the ladder just wrote. A
// client that silently ignored the deletion reports a text that no longer
// ends with it and lands in the INFO mismatch counter (Pitfall 3).
func (a *Actor) armAfterVerify(converted, tail []rune) {
	a.verifyEpoch++
	a.clearAfter() // never two open rounds
	a.after = &pendingAfter{
		expected: concatRunes(converted, tail),
		epoch:    a.verifyEpoch,
	}
	epoch := a.verifyEpoch
	a.after.deadline = time.AfterFunc(verifyWait, func() { a.afterExpiry(epoch) })
	a.eng.RequireSurroundingText()
}

// clearAfter retires the verify-after round together with its deadline
// timer; the caller holds the mutex.
func (a *Actor) clearAfter() {
	if a.after == nil {
		return
	}
	if a.after.deadline != nil {
		a.after.deadline.Stop()
	}
	a.after = nil
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
// it). The CORR-09 reset keyvals (Enter and its keypad variant, Tab,
// Escape) end the phrase instead of feeding it. The caller holds the mutex.
func (a *Actor) feedKey(ev engine.EngineEvent) bool {
	switch {
	case ev.Keyval == engine.KeyBackSpace:
		a.buf.Backspace()

		return false
	case isResetKeyval(ev.Keyval):
		// CORR-09: the commit/abort keys end the phrase — the buffer dies
		// with its word's context. The reset is engine state, never
		// consumption: the key transits so the client sees its Enter, Tab
		// or Escape exactly as before.
		a.buf.HardReset()

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

// startCorrection launches the WORD correction of the Double decision: the
// range is the token with its boundary tail, exactly as in Phase 2 (D-23).
// The caller holds the mutex.
func (a *Actor) startCorrection() {
	a.startRangeCorrection(correctionRange{
		token:   a.buf.Token(),
		tail:    a.buf.Tail(),
		replace: a.buf.ReplaceToken,
	})
}

// startPhraseCorrection launches the PHRASE correction of the Triple
// decision (CORR-02, D-25): the range is the whole buffer since the last
// hard reset — words and separators alike, no tail, no artificial length
// cap. The same pipeline runs over it (D-23). The caller holds the mutex.
func (a *Actor) startPhraseCorrection() {
	a.startRangeCorrection(correctionRange{
		token:   a.buf.Phrase(),
		replace: a.buf.ReplacePhrase,
	})
}

// startRangeCorrection is the ONE correction pipeline of the daemon,
// parameterized by its range (D-23): classify the range, convert it, choose
// the ladder level by the caps bit, then verify against the freshest
// surrounding text (ADR-004 — the verification covers the whole range,
// token+tail, exactly what the ladder deletes). Every refusal logs its D-20
// reason at INFO — without the range's contents — and touches nothing. The
// tracer direction is homogeneous-script: Detect refuses a range with
// letters of both scripts; the run-wise mixed conversion of Phase 3 Task 2
// replaces this step for every range. The caller holds the mutex.
//
// Transport adaptation (live finding, 2026-09-14): neither GTK nor the
// mutter input context answers RequireSurroundingText — clients push
// SetSurroundingText spontaneously on every text change (the pushes at
// cursor 1..6 during typing), exactly how the owner's prototype works.
// The freshest spontaneous push is therefore checked first (a push after
// the last keystroke is the freshest state the client can report); only on
// a miss does the Require round-trip run inside the verifyWait budget.
func (a *Actor) startRangeCorrection(rng correctionRange) {
	armed := time.Now()
	if len(rng.token) == 0 {
		slog.Info("correction skipped", "reason", "empty-buffer")

		return
	}
	dir, ok := correct.Detect(rng.token)
	if !ok {
		slog.Info("correction skipped", "reason", tokenRefusal(rng.token))

		return
	}
	converted, ok := correct.Convert(rng.token, dir)
	if !ok {
		slog.Info("correction skipped", "reason", "convert-failed")

		return
	}
	if a.eng == nil {
		slog.Info("correction skipped", "reason", "no-engine")

		return
	}
	a.resolvePending() // a second Double/Triple supersedes the stale round
	a.clearAfter()     // …and the previous round's verify-after with it
	if a.caps&correct.CapSurroundingText == 0 {
		// Ladder level 2 (ADR-003): no surrounding text, no verification —
		// the documented ADR-004 degradation. The buffer plus the explicit
		// CORR-09 reset triggers are the only synchronization the daemon
		// has, so the correction executes immediately, never guessed twice.
		a.executeLevel2(rng, converted, armed)

		return
	}
	a.pending = &pendingFix{
		rng:       rng,
		converted: converted,
		match:     concatRunes(rng.token, rng.tail),
		armed:     armed,
	}
	if correct.MatchesSuffix(a.surr, a.pending.match) {
		a.executeCorrection() // cached push is the freshest report

		return
	}
	a.pending.deadline = time.AfterFunc(verifyWait, a.VerifyExpiry)
	a.eng.RequireSurroundingText()
}

// executeCorrection runs the level-1 ladder plan of the settled pending fix
// — one DeleteSurroundingText exactly over the token+tail range, one commit
// of the converted token plus the tail (CORR-07) — replaces the range in
// the buffer so a repeated correction converts back, and arms the
// verify-after round: the deletion is ack-less on 1.5.29, so the correction
// itself asks for the surrounding text back and checks the suffix
// (ADR-003/ADR-004). The caller holds the mutex and pending != nil.
func (a *Actor) executeCorrection() {
	p := a.pending
	a.pending = nil
	if p.deadline != nil {
		p.deadline.Stop()
	}
	plan := correct.BuildPlan(p.rng.token, p.rng.tail, p.converted, a.caps, correct.DefaultBackspaceCap)
	a.eng.DeleteSurroundingText(plan.Offset, plan.NChars)
	a.eng.CommitText(engine.NewIBusText(string(plan.Commit)))
	p.rng.replace(p.converted) // the buffer keeps mirroring the field — repeat converts back
	logCorrectionDone(plan.Level, p.rng.token, p.converted, time.Since(p.armed))
	a.armAfterVerify(p.converted, p.rng.tail)
}

// executeLevel2 runs the ladder's Backspace level on a client without the
// surrounding-text capability (ADR-003): plan.Backspaces replayed
// ForwardKeyEvent(BackSpace) — counted in runes by BuildPlan, the tail
// included — followed by ONE commit of the converted token plus the tail.
// The burst→commit order is the wire contract (research A4: the daemon
// preserves it for every client). The caller holds the mutex.
func (a *Actor) executeLevel2(rng correctionRange, converted []rune, armed time.Time) {
	plan := correct.BuildPlan(rng.token, rng.tail, converted, a.caps, correct.DefaultBackspaceCap)
	for range plan.Backspaces {
		a.eng.ForwardKeyEvent(engine.KeyBackSpace, backSpaceKeycode, 0)
	}
	a.eng.CommitText(engine.NewIBusText(string(plan.Commit)))
	rng.replace(converted)
	logCorrectionDone(plan.Level, rng.token, converted, time.Since(armed))
}

// logCorrectionDone writes the completion pair of one ladder execution: the
// INFO counter (D-20 — outcome only, never the word) and the DEBUG detail
// record whose FIRST attribute after msg is the ladder level — the exact
// form the e2e matrix greps to pin the ACTUAL level a client got (D-21).
func logCorrectionDone(level correct.Level, token, converted []rune, latency time.Duration) {
	slog.Info("correction", "outcome", "done")
	slog.Debug("correction",
		"level", int(level),
		"runes", len(token),
		"source", string(token),
		"result", string(converted),
		"latency_ms", latency.Milliseconds())
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

// isResetKeyval reports whether keyval is one of the CORR-09 hard-reset
// triggers observable at the IME level: Enter and its keypad variant, Tab
// and Escape (ADR-004 — the mouse click is not observable here and is
// replaced by the pre-correction verification, not by this table).
func isResetKeyval(keyval uint32) bool {
	switch keyval {
	case engine.KeyReturn, engine.KeyKPEnter, engine.KeyTab, engine.KeyEscape:
		return true
	default:
		return false
	}
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
