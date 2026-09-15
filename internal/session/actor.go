// Package session hosts the daemon's event consumers: the actor that
// serializes engine events into the hotkey FSM, feeds the typed-phrase
// buffer and runs the two-phase word correction (CORR-01).
package session

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/Djarvur/goswitch/engine"
	"github.com/Djarvur/goswitch/internal/clipboard"
	"github.com/Djarvur/goswitch/internal/config"
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
	mu           sync.Mutex
	fsm          *hotkey.FSM
	window       time.Duration
	start        time.Time
	timer        *time.Timer
	buf          *correct.Buffer
	caps         uint32
	eng          engine.Emitter
	surr         []rune // cached text-before-cursor from the latest client push
	sel          selectionState
	pending      *pendingFix
	after        *pendingAfter
	verifyEpoch  uint64               // monotonic verify-after round tag (stale-timer guard)
	mode         scriptMode           // output-script state, EN at start (ADR-001 Option B)
	opts         Options              // correction tuning (D-27 cap, D-28 rung switch, D-36 combo)
	clip         *clipboard.Clipboard // the rung's wl-clipboard client
	comboPending bool                 // a word-layout combo awaits its word pipeline's settlement (D-36)
	// cfgSrc is the live config source (the 03-02 watcher's Snapshot
	// contract); nil on the no-config path, where SetOptions and the
	// built-in defaults govern.
	cfgSrc    interface{ Snapshot() config.Config }
	comboName string // the resolved document's combo binding name (parse cache)
	// MACR state (plan 03-05, ADR-005): the Super-hold window of the
	// consumed-upstream detect (b.2) with its letter witness, the pending
	// remap awaiting the hold's end, the layer's counters (the goswitchctl
	// status surface of 03-06) and the letters parse cache of the snapshot
	// consumption.
	macrSuperHeld     bool
	macrSawLetter     bool
	macrPendingKeyval uint32
	macrIntercepted   int
	macrConsumed      int
	macrLettersName   string
}

// Options is the correction-tuning surface of the actor (plan 03-03): the
// D-27 Backspace series cap and the D-28 opt-in clipboard rung switch —
// OFF at the zero value, which is the default configuration — plus the
// D-36 word-layout combo binding (the zero Binding selects the built-in
// Shift+Control_R default) and the MACR-01 Super→Ctrl layer of ADR-005
// (plan 03-05): OFF at the zero value, with an empty letter set, no per-app
// list and NO alternative modifier (b.3 — not introduced by default).
type Options struct {
	BackspaceCap    int
	ClipboardRung   bool
	WordLayoutCombo hotkey.Binding
	MACREnabled     bool
	MACRLetters     map[rune]bool
	MACRApps        []string
	MACRAltModifier string
}

// MACRStats are the Super→Ctrl layer's counters (ADR-005 b.2) — the status
// surface the 03-06 goswitchctl reads; the field names are the contract.
type MACRStats struct {
	SuperIntercepted int
	ConsumedUpstream int
}

// defaultComboBinding is the built-in word-layout combo (D-36/SPEC §4.1
// "Shift+RightCtrl"): wire-equal to hotkey.ParseBinding("shift+ctrl_r") —
// the config default of 03-02 — under that package's ModMask semantics
// (the full modifier state of the bound key's event: the held Shift OR the
// Control_R press's own family bit). A literal, not a ParseBinding call,
// so the built-in path needs no error handling; the equivalence is pinned
// by the combo corpus against both spellings of the same gesture.
func defaultComboBinding() hotkey.Binding {
	return hotkey.Binding{Keyval: hotkey.KeyvalCtrlR, ModMask: hotkey.MaskShift | hotkey.MaskControl}
}

// selectionState is the selection half of the latest surrounding-text push
// (D-30): the full text the client reported plus BOTH wire positions. A
// selection is active exactly when cursor != anchor — the anchor is the
// selection boundary the client pushes with every text change (GTK's
// IMContext, ibusengine.h:430); a selection may sit on either side of the
// cursor, hence the full text, not just the before-cursor prefix.
type selectionState struct {
	full   []rune
	cursor uint32
	anchor uint32
}

// correctionRange parameterizes the correction pipeline by its range (D-23 —
// one pipeline, different ranges): the word path (Double) supplies the token,
// its boundary tail and ReplaceToken; the phrase path (Triple) supplies the
// whole phrase since the last hard reset, no tail and ReplacePhrase (D-25).
// The selection path (03-03) supplies the client-reported selection — the
// range lives in the surrounding text, not the buffer, so its replace half
// is the buffer's honest death (HardReset) instead of an edit.
type correctionRange struct {
	token   []rune
	tail    []rune
	replace func(converted []rune)
}

// selectionSpec is the wire geometry of a selection correction (D-30,
// Pitfall 6): the selected half-open range [start,end) inside the
// client-reported text and the cursor position the deletion offset is
// signed against — a selection left of the cursor deletes with a negative
// offset, one right of the cursor with a positive one.
type selectionSpec struct {
	start  uint32
	end    uint32
	cursor uint32
}

// pendingFix is the state of a correction between the Double/Triple decision
// and the surrounding-text verdict: the range under correction (token, its
// boundary tail), what to commit (the converted token), the range the
// verification must see at the end of the text (token+tail — exactly what the
// ladder deletes) and the verify deadline. A selection correction carries
// its geometry in sel instead of the buffer's replace bookkeeping.
type pendingFix struct {
	rng       correctionRange
	converted []rune
	match     []rune
	armed     time.Time
	deadline  *time.Timer
	sel       *selectionSpec
}

// pendingAfter is the state of the verify-after round — the post-correction
// check that compensates the ack-less DeleteSurroundingText (ADR-003): the
// suffix the client's fresh surrounding text must END with (converted+tail)
// and the round's epoch. The epoch makes a stale deadline from a superseded
// round a no-op: the timer callback re-checks the tag under the mutex. A
// selection round checks the expected text AT the range position (atRange)
// instead of at the text's end — Pitfall 6 again. paste carries the rung's
// payload (the converted text, D-28) for the mismatch branch.
type pendingAfter struct {
	expected []rune
	paste    []rune
	start    uint32
	atRange  bool
	epoch    uint64
	deadline *time.Timer
}

// NewActor returns an actor deciding tap series inside the given
// disambiguation window (D-05). The daemon passes hotkey.DefaultWindow.
func NewActor(window time.Duration) *Actor {
	return &Actor{
		fsm:    hotkey.NewFSM(window, hotkey.KeyvalShiftR),
		window: window,
		start:  time.Now(),
		buf:    correct.NewBuffer(),
		clip:   clipboard.New(),
	}
}

// SetOptions stores the correction options (the startup wiring feeds them
// from config.Load; the snapshot consumption on hot reload is plan 03-04).
// The caller is the daemon or a test — the actor reads the options under
// its mutex at use time.
func (a *Actor) SetOptions(o Options) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.opts = o
}

// AttachConfig connects the live config source (the 03-02 watcher's
// Snapshot contract). The actor reads one snapshot per event — the window
// for arming NEW series, the options and the combo binding — while an
// already-armed timer keeps its own deadline (Pitfall 8: a reload never
// re-arms a live timer).
func (a *Actor) AttachConfig(src interface{ Snapshot() config.Config }) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.cfgSrc = src
}

// UseClipboard replaces the actor's clipboard client — the test seam of the
// D-28 rung (the daemon keeps the production wl-clipboard client NewActor
// wired).
func (a *Actor) UseClipboard(c *clipboard.Clipboard) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.clip = c
}

// MACRCounters snapshots the Super→Ctrl layer's counters — the goswitchctl
// status surface of plan 03-06 (ADR-005 b.2: every intercepted combination
// and every upstream-consumed one is countable, never silent).
func (a *Actor) MACRCounters() MACRStats {
	a.mu.Lock()
	defer a.mu.Unlock()

	return MACRStats{SuperIntercepted: a.macrIntercepted, ConsumedUpstream: a.macrConsumed}
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

	a.applySnapshot() // one config read per event (CONF-02, Pattern 2)

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
		a.macrKeyRelease(ev)

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
		// the caches belong to the input context that just left
		a.surr = nil
		a.sel = selectionState{}
		a.resolvePending()
		a.clearAfter()
		// A combo retired by the context's death never settles its word
		// half — the flip is dropped with it (cleared WITHOUT settleCombo:
		// the gesture belonged to the input context that just left).
		a.comboPending = false
		// A pending MACR remap dies with its context the same way: the
		// deferred burst belongs to the input context that held the chord.
		a.macrSuperHeld = false
		a.macrPendingKeyval = 0
		if a.timer != nil {
			a.timer.Stop()
			a.timer = nil
		}
	case engine.LifecycleFocusIn, engine.LifecycleEnable, engine.LifecycleDisable:
		slog.Debug("lifecycle", "kind", kind.String())
	}
}

// HandleSurroundingText implements engine.EventHandler: the arriving text
// updates the surrounding cache — the before-cursor prefix for the suffix
// verification and the full text with BOTH positions for the selection
// detection (D-30) — and settles whichever round is open. A pending
// correction (ADR-004 pre-check) executes only when the runes before the
// cursor END with the correction range (token+tail — exactly what the
// ladder deletes, CORR-07); otherwise it aborts silently ("abort, не
// мусорить": not one character is touched). A pending verify-after
// (ADR-003/ADR-004 post-check) compares the suffix against the replacement
// the correction left behind: a mismatch — the client ignored the deletion,
// Chromium ibus#2354, Pitfall 3 — counts as an INFO record and is NEVER
// followed by an automatic repair.
func (a *Actor) HandleSurroundingText(text string, cursorPos, anchorPos uint32) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.applySnapshot() // one config read per event (CONF-02, Pattern 2)

	a.surr = beforeCursor(text, cursorPos)
	a.sel = selectionState{full: []rune(text), cursor: cursorPos, anchor: anchorPos}
	if a.pending != nil {
		if !a.pendingVerdict() {
			a.resolvePending()
			slog.Info("correction skipped", "reason", "verify-mismatch")
			a.settleCombo() // the word half settled by refusal — the D-36 flip still fires

			return
		}
		a.executeCorrection()

		return
	}
	if a.after != nil {
		expected := a.after.expected
		paste := a.after.paste
		start, atRange := a.after.start, a.after.atRange
		a.clearAfter()
		if !a.afterVerdict(expected, start, atRange) {
			slog.Info("correction verify", "outcome", "mismatch")
			a.runClipboardRung(paste)

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

	a.applySnapshot() // one config read per event (CONF-02, Pattern 2)

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

	a.applySnapshot() // one config read per event (CONF-02, Pattern 2)

	if a.pending == nil {
		return // already settled — a stale timer is a no-op
	}
	a.resolvePending()
	a.clearAfter()
	slog.Info("correction skipped", "reason", "verify-timeout")
	a.settleCombo() // the round closed by timeout — the combo's flip still fires
}

// effectiveCombo resolves the combo binding in force: the configured
// binding, or the built-in default when none was fed (the zero Binding of
// the no-config path). The caller holds the mutex.
func (a *Actor) effectiveCombo() hotkey.Binding {
	if a.opts.WordLayoutCombo.Keyval != 0 {
		return a.opts.WordLayoutCombo
	}

	return defaultComboBinding()
}

// settleCombo fires the combo's layout flip: the word half of the gesture
// has SETTLED — its completion or refusal record is already in the log —
// so the D-36 order (correct the word FIRST, switch the layout SECOND)
// holds on every pipeline outcome. The caller holds the mutex and must
// call this only AFTER the settled record.
func (a *Actor) settleCombo() {
	if !a.comboPending {
		return
	}
	a.comboPending = false
	a.flipScript()
}

// applySnapshot reads the live config source ONCE and folds the document
// into the actor (CONF-02, Pattern 2: a value into local state — a pointer
// is never held across events). The snapshot's tap window reaches both the
// timer arming of NEW series (armTimer reads a.window) and the FSM's gap
// and expiry checks (SetWindow); an already-armed timer keeps its own
// deadline — a reload never re-arms or cancels a live decision (Pitfall 8).
// The options overwrite whatever SetOptions fed (an attached source has
// priority — the documented succession: SetOptions stays the surface of
// the no-config path and the tests). The combo binding is re-resolved only
// when the document's binding NAME changed; a name that fails to parse —
// impossible from a validated document — keeps the last-good binding (the
// D-32 discipline). The caller holds the mutex.
func (a *Actor) applySnapshot() {
	if a.cfgSrc == nil {
		return
	}
	snap := a.cfgSrc.Snapshot()

	if w := time.Duration(snap.Timeouts.TapWindowMs) * time.Millisecond; w != a.window {
		a.window = w
		a.fsm.SetWindow(w)
	}
	a.opts.BackspaceCap = snap.Correction.BackspaceCap
	a.opts.ClipboardRung = snap.Correction.ClipboardRung
	if name := snap.Hotkeys.WordLayoutCombo; name != a.comboName {
		if binding, err := hotkey.ParseBinding(name); err == nil {
			a.opts.WordLayoutCombo = binding
			a.comboName = name
		}
	}
	a.opts.MACREnabled = snap.MACR.Enabled
	a.opts.MACRApps = snap.MACR.Apps
	a.opts.MACRAltModifier = snap.MACR.AltModifier
	if snap.MACR.Letters != a.macrLettersName {
		a.macrLettersName = snap.MACR.Letters
		a.opts.MACRLetters = parseMACRLetters(snap.MACR.Letters)
	}
}

// pendingVerdict checks the pending fix's range against the fresh push: the
// word and phrase paths need the text before the cursor to END with their
// range (token+tail), the selection path needs its runes AT the reported
// position — the selection may sit on either side of the cursor (Pitfall 6).
// The caller holds the mutex and pending != nil.
func (a *Actor) pendingVerdict() bool {
	if a.pending.sel == nil {
		return correct.MatchesSuffix(a.surr, a.pending.match)
	}

	return correct.VerifyRangeAt(a.sel.full, a.pending.sel.start, a.pending.match)
}

// afterVerdict checks the verify-after expectation against the fresh push:
// the suffix rule for word/phrase corrections, the range rule for selection
// corrections. The caller holds the mutex (and has already retired the
// round — the arguments carry its state).
func (a *Actor) afterVerdict(expected []rune, start uint32, atRange bool) bool {
	if !atRange {
		return correct.MatchesSuffix(a.surr, expected)
	}

	return correct.VerifyRangeAt(a.sel.full, start, expected)
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
		paste:    converted,
		epoch:    a.verifyEpoch,
	}
	epoch := a.verifyEpoch
	a.after.deadline = time.AfterFunc(verifyWait, func() { a.afterExpiry(epoch) })
	a.eng.RequireSurroundingText()
}

// armAfterVerifyRange starts the range-anchored verify-after round of a
// selection correction (the caller holds the mutex): the client's fresh
// push must hold the converted text AT the selection position — the range
// may sit on either side of the cursor, so the suffix rule of the word path
// would look at the wrong end of the text (Pitfall 6).
func (a *Actor) armAfterVerifyRange(converted []rune, start uint32) {
	a.verifyEpoch++
	a.clearAfter() // never two open rounds
	a.after = &pendingAfter{
		expected: converted,
		paste:    converted,
		start:    start,
		atRange:  true,
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

// ctrlVKeycodes are the physical (evdev) keycodes of the Ctrl+V forward
// burst — KEY_LEFTCTRL=29, KEY_V=47 (linux/event-codes.h), the same keycode
// discipline as backSpaceKeycode=14; ctrlRKeycode is the right-Ctrl half
// of the MACR alt_modifier burst — KEY_RIGHTCTRL=97.
const (
	ctrlLKeycode = 29
	ctrlVKeycode = 47
	ctrlRKeycode = 97
)

// runClipboardRung runs the D-28 opt-in clipboard rung after a verify-after
// mismatch (ADR-003 rung C, ordered after the commit/delete rungs): save
// the user's clipboard byte-exactly, put the converted replacement in
// through the stdin-only wl-copy, replay the Ctrl+V burst over whatever the
// client still holds selected, and restore the saved state best-effort — a
// restore failure is a WARN (D-29), never an operation error. Clipboard
// CONTENT is never logged at any level (D-20/D-21 extension), and the rung
// is OFF unless Options.ClipboardRung armed it (the default configuration
// never touches the user's clipboard). The caller holds the mutex.
func (a *Actor) runClipboardRung(paste []rune) {
	if !a.opts.ClipboardRung || a.clip == nil {
		return
	}
	// The clipboard client bounds every subprocess with its own deadline
	// (T-03-03-05) — a background context is the rung's lifetime.
	ctx := context.Background()
	saved, had, err := a.clip.Save(ctx)
	if err != nil {
		slog.Info("correction skipped", "reason", "clipboard-unavailable")

		return
	}
	if err := a.clip.Set(ctx, []byte(string(paste))); err != nil {
		slog.Info("correction skipped", "reason", "clipboard-unavailable")

		return
	}
	forwardCtrlV(a.eng)
	if err := a.clip.Restore(ctx, saved, had); err != nil {
		// D-29: best-effort — the replacement already landed in the field;
		// a failed restore is logged (error only, never content), not raised.
		slog.Warn("clipboard restore failed", "error", err)
	}
}

// forwardCtrlV replays the paste over the active selection: the owner
// prototype's validated four-event sequence — Control_L press bare, v press
// under Control, v release, Control_L release (punto_engine.py:301-304).
func forwardCtrlV(eng engine.Emitter) {
	eng.ForwardKeyEvent(engine.KeyControlL, ctrlLKeycode, 0)
	eng.ForwardKeyEvent(engine.KeyV, ctrlVKeycode, engine.MaskControl)
	eng.ForwardKeyEvent(engine.KeyV, ctrlVKeycode, engine.MaskControl|engine.MaskRelease)
	eng.ForwardKeyEvent(engine.KeyControlL, ctrlLKeycode, engine.MaskControl|engine.MaskRelease)
}

// macrIntercept is the MACR-01 branch (ADR-005, Pattern 6): a configured
// letter press carrying Mod4 — the wire truth of the live probe
// (2026-09-15: super+b arrives as keyval 0x62 with Mod4|NumLock, in EVERY
// internal mode) — is consumed and its remap deferred to the hold's end:
// the Ctrl+letter burst is emitted by macrKeyRelease when the Super key
// itself is released (the live truth of the first implementation: a burst
// forwarded while the physical Super is still held reaches the client as a
// Ctrl+Super chord — the client's own modifier tracking pollutes the
// synthetic event, the binding never matches; the proven forwardCtrlV of
// the clipboard rung fires with clean state for exactly this reason). The
// Super press itself opens the hold window of the consumed-upstream detect
// and any letter press under Mod4 marks the hold as witnessed (the chord's
// letter reached the IME — nothing was swallowed upstream, intercepted or
// not). The INFO record names the CONFIG letter only (D-20: config letters
// are not user data; keystroke content never enters the log). The caller
// holds the mutex; releases never reach this branch (HandleKey routes them
// to macrKeyRelease).
func (a *Actor) macrIntercept(ev engine.EngineEvent) bool {
	if isSuperKeyval(ev.Keyval) {
		a.macrSuperHeld = true
		a.macrSawLetter = false

		return false
	}
	if ev.Mods&engine.MaskMod4 == 0 {
		return false
	}
	if isLetterKeyval(ev.Keyval) {
		a.macrSawLetter = true
	}
	if !a.opts.MACREnabled || !a.macrTargetActive() || a.eng == nil {
		return false
	}
	// #nosec G115 -- isLetterKeyval bounds the keyval to 'a'..'z', so the
	// uint32→rune conversion cannot overflow.
	if !a.opts.MACRLetters[rune(ev.Keyval)] {
		return false
	}
	a.macrIntercepted++
	a.macrPendingKeyval = ev.Keyval
	// #nosec G115 -- same bound: the letter grammar of the config schema.
	slog.Info("super intercept", "key", string(rune(ev.Keyval)))

	return true
}

// macrKeyRelease closes the Super hold (ADR-005 b.2) and delivers the
// pending remap: the Ctrl+letter burst is replayed HERE, after the
// physical Super left the keyboard, so the client receives it with clean
// modifier state. A Super release with NO letter press seen while held
// means the chord was consumed before the IME — a mutter keybinding took
// it (the live probe's Super+a/toggle-application-view shape: the press
// reached the engine, the shell surface then swallowed the rest) — and
// WARNs with the grep-stable reason; the detect is armed only while the
// layer is enabled, so a MACR-off desktop never warns on its native Super
// use (the overview toggle among them). The caller holds the mutex.
func (a *Actor) macrKeyRelease(ev engine.EngineEvent) {
	if !isSuperKeyval(ev.Keyval) || !a.macrSuperHeld {
		return
	}
	a.macrSuperHeld = false
	if a.macrPendingKeyval != 0 && a.eng != nil {
		pending := a.macrPendingKeyval
		a.macrPendingKeyval = 0
		forwardCtrlLetter(a.eng, pending, a.opts.MACRAltModifier)

		return
	}
	if !a.opts.MACREnabled || a.macrSawLetter {
		return
	}
	a.macrConsumed++
	slog.Warn("super combo skipped", "reason", "consumed-upstream")
}

// macrTargetActive reports whether the MACR rule governs the current
// focus: with an empty per-app list the rule is global (ADR-005 a); a
// non-empty list defers to the app-identity observer (plan 03-05 Task 3) —
// an unavailable observer degrades to the global rule, never to silence
// (the ADR-005 ladder). The caller holds the mutex.
func (a *Actor) macrTargetActive() bool {
	return len(a.opts.MACRApps) == 0
}

// isSuperKeyval reports whether keyval is one of the Super modifier
// keysyms (the press/release pair the consumed-upstream detect tracks).
func isSuperKeyval(keyval uint32) bool {
	return keyval == engine.KeySuperL || keyval == engine.KeySuperR
}

// isLetterKeyval reports whether the keyval is a lowercase Latin letter —
// the MACR letter grammar of the config schema (single a-z tokens, 03-02).
func isLetterKeyval(keyval uint32) bool {
	return keyval >= 'a' && keyval <= 'z'
}

// parseMACRLetters resolves the config's comma-joined letter tokens into
// the interception set. The schema validated every token already
// (03-02) — anything else is defensively skipped, never a failure: an
// empty set simply intercepts nothing.
func parseMACRLetters(letters string) map[rune]bool {
	set := make(map[rune]bool)
	for _, tok := range strings.Split(letters, ",") {
		if len(tok) == 1 {
			set[rune(tok[0])] = true
		}
	}

	return set
}

// forwardCtrlLetter replays the Super→Ctrl remap (MACR-01): the owner
// prototype's four-event shape (forwardCtrlV's discipline) with the letter
// in place of v and the modifier key from the config — Control_L by
// default; a configured alt_modifier (ADR-005 b.3) swaps the modifier key
// only, the burst shape never changes and the default stays empty.
func forwardCtrlLetter(eng engine.Emitter, keyval uint32, altModifier string) {
	modKeyval, modKeycode := uint32(engine.KeyControlL), uint32(ctrlLKeycode)
	if altModifier == "ctrl_r" {
		modKeyval, modKeycode = engine.KeyControlR, ctrlRKeycode
	}
	letterCode := letterKeycode(keyval)
	eng.ForwardKeyEvent(modKeyval, modKeycode, 0)
	eng.ForwardKeyEvent(keyval, letterCode, engine.MaskControl)
	eng.ForwardKeyEvent(keyval, letterCode, engine.MaskControl|engine.MaskRelease)
	eng.ForwardKeyEvent(modKeyval, modKeycode, engine.MaskControl|engine.MaskRelease)
}

// letterKeycode returns the evdev keycode of a Latin letter — the keycode
// half of the Ctrl+letter forward burst. Zero for anything else (the
// interception set is schema-validated a-z, so unreachable in practice).
func letterKeycode(keyval uint32) uint32 {
	if keyval < 'a' || keyval > 'z' {
		return 0
	}

	return macrLetterKeycodes()[keyval-'a']
}

// macrLetterKeycodes returns the evdev keycode table of the Latin letters
// indexed by position ('a'=0): KEY_A..KEY_Z of
// /usr/include/linux/input-event-codes.h, verified verbatim — never from
// memory, the 02-05 class trap; the live probe of 2026-09-15 confirmed
// a=30 and b=48 on the wire.
func macrLetterKeycodes() [26]uint32 {
	return [26]uint32{
		30, 48, 46, 32, 18, 33, 34, 35, 23, 36, 37, 38, 50,
		49, 24, 25, 16, 19, 31, 20, 22, 47, 17, 45, 21, 44,
	}
}

// backspaceCap resolves the effective D-27 cap: the configured value, or
// the documented default when no options were fed (the zero Options of a
// bare NewActor — the no-config path and the unit corpus). The caller holds
// the mutex.
func (a *Actor) backspaceCap() int {
	if a.opts.BackspaceCap > 0 {
		return a.opts.BackspaceCap
	}

	return correct.DefaultBackspaceCap
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
// it). The MACR interception (MACR-01, ADR-005, Pattern 6) is recognized
// FIRST of all, above the combo and every mode branch: a configured letter
// press carrying Mod4 — the press-side wire truth of the live probe: the
// letter arrives Latin with Mod4 in EVERY internal mode — is consumed and
// replayed as the Ctrl+letter forward burst, so the RU commit can never
// fire on a Super chord and the buffer is never fed. The word-layout combo
// (D-36) is recognized next, above every mode branch: a press of the bound
// combo key under its bound HELD modifiers — the press's state word
// carries only the modifiers held before the key (live finding
// 2026-09-15: a Control_R press under Shift arrives with Shift|NumLock,
// its own Control bit rides only on the release), so the match compares
// Binding.ModMask with the key's own family bit cleared — latch-tolerant
// through &, the NumLock precedent of 02-04 — kills the tap series with a
// deliberate Reset (Pitfall 4: left to the FSM the Control_R press would
// silently die as modifier use) and launches the word pipeline of the
// Double semantics with the flip deferred to its settlement
// (comboPending/settleCombo). The combo press itself transits: a bare
// modifier chord puts no rune in the field, and the transit keeps the
// client's press/release pairing intact. The CORR-09 reset keyvals (Enter
// and its keypad variant, Tab, Escape) end the phrase instead of feeding
// it. The caller holds the mutex.
func (a *Actor) feedKey(ev engine.EngineEvent) bool {
	if a.macrIntercept(ev) {
		return true
	}

	if c := a.effectiveCombo(); ev.Keyval == c.Keyval {
		// The HELD subset of the binding: a press never carries the key's
		// own family bit, so it is cleared from the required mask.
		if held := c.ModMask &^ hotkey.FamilyMask(c.Keyval); ev.Mods&held == held {
			a.fsm.Feed(hotkey.Reset{}, a.elapsed())
			slog.Info("combo", "kind", "word-layout")
			a.comboPending = true
			a.startCorrection()

			return false
		}
	}

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

// startCorrection launches the correction of the Double decision: with an
// active selection in the last surrounding push the range is that selection
// (CORR-03, D-30 — the selection takes precedence over the buffer, whatever
// the buffer holds); otherwise the WORD path of Phase 2, exactly as before.
// The caller holds the mutex.
func (a *Actor) startCorrection() {
	if sel, runes, ok := a.activeSelection(); ok {
		a.startSelectionCorrection(sel, runes)

		return
	}
	a.startRangeCorrection(correctionRange{
		token:   a.buf.Token(),
		tail:    a.buf.Tail(),
		replace: a.buf.ReplaceToken,
	})
}

// activeSelection resolves the selection half of the last surrounding push
// into a correction range (D-30): active exactly when the client reported
// cursor != anchor, clamped to the text it positions into — a pair the text
// cannot hold describes no selection, and the word path keeps the decision
// (T-03-03-03: client positions are untrusted input). The caller holds the
// mutex.
func (a *Actor) activeSelection() (selectionSpec, []rune, bool) {
	start, end, active := correct.SelectionRange(a.sel.cursor, a.sel.anchor)
	if !active {
		return selectionSpec{}, nil, false
	}
	runes := a.sel.full
	if int(end) > len(runes) {
		// #nosec G115 -- len(runes) is bounded by a real input field, far
		// below 2^31 runes; on the 64-bit target an int always holds a uint32.
		end = uint32(len(runes)) // clamp to the reported text (beforeCursor idiom)
	}
	// #nosec G115 -- both index a real input field, far below 2^31 runes.
	if int(start) >= int(end) {
		return selectionSpec{}, nil, false // collapsed against the text — no range
	}

	return selectionSpec{start: start, end: end, cursor: a.sel.cursor}, runes[start:end], true
}

// startSelectionCorrection launches the SELECTION correction (CORR-03,
// D-30): the range is the client-reported selection, converted by the same
// run pipeline as the word and the phrase (D-23 — ConvertRuns over the
// range), with the same refusal vocabulary and the same two-phase
// verification; the pre-check is range-anchored (VerifyRangeAt — the
// selection may sit on either side of the cursor, Pitfall 6), and the
// freshest spontaneous push is checked first, exactly as for the word. The
// caller holds the mutex.
func (a *Actor) startSelectionCorrection(sel selectionSpec, runes []rune) {
	armed := time.Now()
	converted, changed, ok := correct.ConvertRuns(runes)
	if !ok {
		slog.Info("correction skipped", "reason", refusalReason(runes))
		a.settleCombo()

		return
	}
	if !changed {
		// D-24: every letter of the range is already in the anchor layout —
		// a SUCCESSFUL operation without changes; nothing to replace,
		// nothing to verify.
		slog.Info("correction", "outcome", "done")
		a.settleCombo()

		return
	}
	if a.eng == nil {
		slog.Info("correction skipped", "reason", "no-engine")
		a.settleCombo()

		return
	}
	a.resolvePending() // a second Double supersedes the stale round
	a.clearAfter()     // …and the previous round's verify-after with it
	// A selection exists only because a surrounding push carried it — the
	// client has the capability bit by construction; the caps branch of the
	// word path (level 2 without verification) cannot apply here.
	a.pending = &pendingFix{
		rng:       correctionRange{token: runes},
		converted: converted,
		match:     runes,
		armed:     armed,
		sel:       &sel,
	}
	if correct.VerifyRangeAt(a.sel.full, sel.start, runes) {
		a.executeCorrection() // cached push is the freshest report

		return
	}
	a.pending.deadline = time.AfterFunc(verifyWait, a.VerifyExpiry)
	a.eng.RequireSurroundingText()
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
// parameterized by its range (D-23): convert the range run-wise through
// ConvertRuns (homogeneous wholesale, mixed by the last-letter anchor —
// D-22/D-23), choose the ladder level by the caps bit, then verify against
// the freshest surrounding text (ADR-004 — the verification covers the
// whole range, token+tail, exactly what the ladder deletes). Every refusal
// logs its D-20 reason at INFO — without the range's contents — and touches
// nothing. The caller holds the mutex.
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
		a.settleCombo() // no word — the flip is the combo's primary intent

		return
	}
	converted, changed, ok := correct.ConvertRuns(rng.token)
	if !ok {
		slog.Info("correction skipped", "reason", refusalReason(rng.token))
		a.settleCombo()

		return
	}
	if !changed {
		// D-24: every letter of the range is already in the anchor layout —
		// a SUCCESSFUL operation without changes (the owner's choice over a
		// refusal and over a WARN); nothing to replace, nothing to verify.
		slog.Info("correction", "outcome", "done")
		a.settleCombo()

		return
	}
	if a.eng == nil {
		slog.Info("correction skipped", "reason", "no-engine")
		a.settleCombo()

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
// (ADR-003/ADR-004). A selection fix runs its own geometry instead
// (executeSelectionCorrection). The caller holds the mutex and pending !=
// nil.
func (a *Actor) executeCorrection() {
	p := a.pending
	a.pending = nil
	if p.deadline != nil {
		p.deadline.Stop()
	}
	if p.sel != nil {
		a.executeSelectionCorrection(p)

		return
	}
	plan := correct.BuildPlan(p.rng.token, p.rng.tail, p.converted, a.caps, a.backspaceCap())
	a.eng.DeleteSurroundingText(plan.Offset, plan.NChars)
	a.eng.CommitText(engine.NewIBusText(string(plan.Commit)))
	p.rng.replace(p.converted) // the buffer keeps mirroring the field — repeat converts back
	logCorrectionDone(plan.Level, p.rng.token, p.converted, time.Since(p.armed))
	a.settleCombo() // D-36: the flip lands strictly after the settled completion record
	a.armAfterVerify(p.converted, p.rng.tail)
}

// executeSelectionCorrection runs the ladder over the selection range: one
// DeleteSurroundingText whose offset is SIGNED by the cursor's side of the
// selection (Pitfall 6: start−cursor — negative when the range sits left of
// the cursor, zero-or-positive when right of it) and whose nchars is exactly
// the range length — not a rune more (CORR-07 precision applied to the
// selection) — one commit of the converted range. The buffer dies instead
// of being edited: a selection may span text the buffer never mirrored, so
// from here on it cannot stand in for the field. The verify-after round is
// range-anchored. The caller holds the mutex and pending != nil.
func (a *Actor) executeSelectionCorrection(p *pendingFix) {
	sel := p.sel
	// #nosec G115 -- both positions index a real input field, far below
	// 2^31 runes; on the 64-bit target an int always holds a uint32.
	offset := int32(sel.start) - int32(sel.cursor)
	nchars := sel.end - sel.start
	a.eng.DeleteSurroundingText(offset, nchars)
	a.eng.CommitText(engine.NewIBusText(string(p.converted)))
	a.buf.HardReset() // the buffer can no longer mirror the replaced field
	logCorrectionDone(correct.Level1, p.rng.token, p.converted, time.Since(p.armed))
	a.settleCombo() // D-36: the flip lands strictly after the settled completion record
	a.armAfterVerifyRange(p.converted, sel.start)
}

// executeLevel2 runs the ladder's Backspace level on a client without the
// surrounding-text capability (ADR-003): plan.Backspaces replayed
// ForwardKeyEvent(BackSpace) — counted in runes by BuildPlan, the tail
// included — followed by ONE commit of the converted token plus the tail.
// The burst→commit order is the wire contract (research A4: the daemon
// preserves it for every client). The caller holds the mutex.
func (a *Actor) executeLevel2(rng correctionRange, converted []rune, armed time.Time) {
	plan := correct.BuildPlan(rng.token, rng.tail, converted, a.caps, a.backspaceCap())
	if plan.Level == correct.LevelNone {
		// D-27: the Backspace series would exceed the cap and this client
		// has no DeleteSurroundingText — refuse silently, not one deletion.
		slog.Info("correction skipped", "reason", "backspace-cap")
		a.settleCombo()

		return
	}
	for range plan.Backspaces {
		a.eng.ForwardKeyEvent(engine.KeyBackSpace, backSpaceKeycode, 0)
	}
	a.eng.CommitText(engine.NewIBusText(string(plan.Commit)))
	rng.replace(converted)
	logCorrectionDone(plan.Level, rng.token, converted, time.Since(armed))
	a.settleCombo()
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

// refusalReason labels a ConvertRuns refusal for the D-20 skip log: a range
// without Latin or Cyrillic letters has nothing to anchor on; anything else
// failed on a letter missing from the layout tables. A diagnostic label
// only — the refusal decision itself belongs to correct.ConvertRuns and is
// not duplicated here.
func refusalReason(token []rune) string {
	for _, r := range token {
		if unicode.Is(unicode.Latin, r) || unicode.Is(unicode.Cyrillic, r) {
			return "convert-failed"
		}
	}

	return "no-letters"
}
