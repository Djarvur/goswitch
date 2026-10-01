// Package session hosts the daemon's event consumers: the actor that
// serializes engine events into the hotkey FSM, feeds the typed-phrase
// buffer and runs the two-phase word correction (CORR-01).
package session

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/Djarvur/goswitch/engine"
	"github.com/Djarvur/goswitch/internal/appid"
	"github.com/Djarvur/goswitch/internal/clipboard"
	"github.com/Djarvur/goswitch/internal/config"
	"github.com/Djarvur/goswitch/internal/correct"
	"github.com/Djarvur/goswitch/internal/detect"
	"github.com/Djarvur/goswitch/internal/hotkey"
	"github.com/Djarvur/goswitch/layouts"
)

// defaultVerifyWait is the ADR-004 verify budget the actor STARTS with
// (resolved planning question 4: ~100 ms — the local D-Bus RTT is 1-10 ms,
// a slow client gets the rest; past the budget the correction is silently
// dropped, never guessed). The budget in force is actor state: the config
// document's timeouts.verify_wait_ms feeds it through applySnapshot
// (CR-02 — the knob is wiring-live, not validation-only).
const defaultVerifyWait = 100 * time.Millisecond

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

// switchTimeout bounds the SetGlobalEngine round trip on the flip path —
// the FAILURE-path wedge guard (Pitfall 4, T-05-03-01), not a slice of the
// <50 ms SPEC reaction budget: it bounds only how long a wedged ibus-daemon
// may hold the flip. The measured success RTT is ~41-45 ms (the 05-05 live
// proofs and the 2026-09-30 journal — the old 40 ms deadline fired before
// completion on every healthy flip), so 150 ms sits above the measured
// maximum: every healthy flip completes inside it, and a wedged bus still
// costs at most the deadline.
const switchTimeout = 150 * time.Millisecond

// autoRoleTimeout bounds the live GetRole round trip of one autocorrect
// confirm (plan 06-06, Pitfall 5): the decision runs OUTSIDE the actor
// mutex with its own deadline, so a stuck a11y object can never stall the
// keystroke path. The measured unix-socket RTT is sub-millisecond — the
// deadline is a wedge guard, budgeted at half of SPEC §5's 50 ms reaction
// ceiling.
const autoRoleTimeout = 25 * time.Millisecond

// The autocorrect abstention vocabulary (plan 06-06, D-53 fail-closed):
// every refusal of the boundary/confirm conjunction counts exactly one of
// these slugs; the arming gate additionally reuses the detector's own
// closed reasons (internal/detect). Closed and word-free (D-20/D-21,
// Pitfall 6).
const (
	acReasonNoApps          = "no-apps"
	acReasonNoCaps          = "no-caps"
	acReasonAppUnknown      = "app-unknown"
	acReasonAppNotListed    = "app-not-listed"
	acReasonRoleUnknown     = "role-unknown"
	acReasonRoleForbidden   = "role-forbidden"
	acReasonRoleTimeout     = "role-timeout"
	acReasonRecheckDisabled = "recheck-disabled"
)

// Task-2 keepalive: the confirm goroutine spends this budget on the live
// GetRole ctx (the D-53 role contour of plan 06-06 Task 2).
var _ = autoRoleTimeout

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
// within the verify budget (timeouts.verify_wait_ms) for clients that
// implement the round trip. The
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
	rungMu       sync.Mutex // serializes the clipboard rung (WR-01: the rung runs off the actor mutex)
	fsm          *hotkey.FSM
	window       time.Duration
	verifyWait   time.Duration // the ADR-004 budget in force (CR-02: fed by timeouts.verify_wait_ms)
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
	cfgSrc interface{ Snapshot() config.Config }
	// tapKeyval is the series key in force — the FSM's key and HandleKey's
	// timer-arming comparison both read it (CR-01: the documented
	// hotkeys.tap_key is wiring-live, not validation-only). tapKeyName is
	// its parse cache, the comboName precedent; comboName is the resolved
	// document's combo binding name; chordName is the mode-switch chord's
	// parse cache (the same precedent, with the empty-disables twist).
	tapKeyval  uint32
	tapKeyName string
	comboName  string
	chordName  string
	// MACR state (plan 03-05, ADR-005): the Super-hold window of the
	// consumed-upstream detect (b.2) with its letter witness, the pending
	// remap awaiting the hold's end, the layer's counters (the goswitchctl
	// status surface of 03-06) and the letters parse cache of the snapshot
	// consumption. The per-app identity source (ADR-005 a) starts lazily —
	// only once a non-empty macr.apps list is in force.
	macrSuperHeld     bool
	macrSawLetter     bool
	macrPendingKeyval uint32
	macrIntercepted   int
	macrConsumed      int
	macrLettersName   string
	// Correction outcome counters — the goswitchctl status surface of plan
	// 03-06 (INST-02): completed corrections (the D-24 success-without-change
	// included) and refusals with their D-20 reason breakdown. Counts only,
	// never content.
	corrDone    int
	corrSkipped int
	skipReasons map[string]int
	// Autocorrect counters (plan 06-06, D-54): the fired corrections, the
	// fail-closed abstentions with their reason slugs — counts and closed
	// slugs only, never the typed or corrected word (D-20/D-21, Pitfall 6).
	// acGeneration tags armed payloads so a newer boundary supersedes a
	// mid-flight confirm; acAppWarned keeps the one-WARN discipline per
	// degradation episode (the appidWarned precedent).
	acFired      int
	acAbstained  int
	acReasons    map[string]int
	acGeneration uint64
	acAppWarned  bool
	appid        AppidSource
	appidStarted bool
	// role is the live AT-SPI role seam of the autocorrect policy (plan
	// 06-06): the SAME observer as appid when the concrete source implements
	// RoleSource, or a test double installed via UseRole.
	role         RoleSource
	appidWarned  bool // one WARN per degradation episode — a broken source must not spam per keystroke
	startAppid   func() (AppidSource, error)
	// version is the daemon's build identity (D-37) pinned at construction;
	// the status snapshot lifts it so goswitchctl status identifies the
	// running build.
	version string
	// switcher is the generation-scoped SetGlobalEngine seam (D-52) the
	// daemon wiring hands over via SetSwitcher before engine.Run; nil = the
	// internal-flip degradation (one WARN per episode at first use).
	switcher func(ctx context.Context, engineName string) error
	// switcherWarned keeps the nil-seam degradation at one WARN per episode —
	// a broken seam must not spam the journal per flip (the appidWarned
	// precedent).
	switcherWarned bool
	// display is the tray-indicator seam (quick 260930-pf6): the ModeDisplay
	// the daemon wiring installs via SetModeDisplay; nil = no display, every
	// mode change stays invisible to it. The display degrades itself — the
	// actor adds no error handling around the call.
	display ModeDisplay
}

// Options is the correction-tuning surface of the actor (plan 03-03): the
// D-27 Backspace series cap and the D-28 opt-in clipboard rung switch —
// OFF at the zero value, which is the default configuration — the
// post-correction script flip of owner decision 2 (260927-vu8):
// FlipAfterCorrection is OFF at the zero value (the unit corpus and the
// no-SetOptions path) and ON when fed from the built-in defaults, whose
// config.Defaults sets flip_after_correction true (the daemon wiring
// passes it through SetOptions; an attached document overrides it live per
// applySnapshot) — plus the D-36 word-layout combo binding (the zero
// Binding selects the built-in Shift+Control_R default), the mode-switch
// chord binding of owner decision 2 (quick plan 260927-way: the ZERO
// Binding is the DISABLED state — unlike the combo there is no built-in
// fallback; the daemon wiring feeds config.Defaults' super+space through
// SetOptions and an attached document folds it live, empty = off) and the
// MACR-01 Super→Ctrl layer of ADR-005 (plan 03-05): OFF at the zero value,
// with an empty letter set, no per-app list and NO alternative modifier
// (b.3 — not introduced by default).
type Options struct {
	BackspaceCap        int
	ClipboardRung       bool
	FlipAfterCorrection bool
	WordLayoutCombo     hotkey.Binding
	ModeSwitchChord     hotkey.Binding
	MACREnabled         bool
	MACRLetters         map[rune]bool
	MACRApps            []string
	MACRAltModifier     string
	// The autocorrect layer (plan 06-06, D-53/D-54): OFF at the zero value —
	// the unit corpus and the no-SetOptions path — exactly like MACR above.
	// The white list matches the focused app's bridge-namespace EXACTLY (no
	// prefix merging); the thresholds mirror detect.Params and the config
	// 06-04 defaults (4/2.0/1.0 — change the places together).
	AutoCorrectEnabled    bool
	AutoCorrectApps       []string
	AutoCorrectMinWordLen int
	AutoCorrectMargin     float64
	AutoCorrectFloor      float64
}

// MACRStats are the Super→Ctrl layer's counters (ADR-005 b.2) — the status
// surface the 03-06 goswitchctl reads; the field names are the contract.
type MACRStats struct {
	SuperIntercepted int
	ConsumedUpstream int
}

// AutoCorrectStats are the autocorrect layer's counters (plan 06-06, D-54)
// — the goswitchctl status surface; the field names are the contract.
// Reasons maps each fail-closed abstention slug to its count; the slugs
// come from a closed vocabulary and never carry the word (D-20/D-21).
type AutoCorrectStats struct {
	Fired     int
	Abstained int
	Reasons   map[string]int
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

// acPayload is one armed boundary decision (plan 06-06): the correction
// range snapshot (token+tail+replace) taken under the mutex at the
// boundary, the detector's direction, the typed word the verdict was about
// and the arming generation. It travels to the confirm goroutine; the word
// stays in memory only — no log and no status surface ever sees it
// (D-20/D-21).
type acPayload struct {
	rng  correctionRange
	dir  correct.Dir
	word []rune
	gen  uint64
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
// payload (the converted text, D-28) for the mismatch branch. mismatches
// counts the stale answers of the WR-05 debounce: the first mismatching
// push re-requires instead of concluding (it may be the client's
// INTERMEDIATE post-delete pre-commit state), the second concludes.
type pendingAfter struct {
	expected   []rune
	paste      []rune
	start      uint32
	atRange    bool
	epoch      uint64
	mismatches int
	deadline   *time.Timer
}

// NewActor returns an actor deciding tap series inside the given
// disambiguation window (D-05). The daemon passes hotkey.DefaultWindow.
// The series key starts at the documented default (shift_r); an attached
// config source re-resolves it from the first event on (CR-01).
func NewActor(window time.Duration) *Actor {
	return &Actor{
		fsm:        hotkey.NewFSM(window, hotkey.KeyvalShiftR),
		window:     window,
		verifyWait: defaultVerifyWait,
		tapKeyval:  hotkey.KeyvalShiftR,
		start:      time.Now(),
		buf:        correct.NewBuffer(),
		clip:       clipboard.New(),
		// The lazy per-app observer's production starter: the a11y dial of
		// internal/appid on the daemon's lifetime (the clipboard rung's
		// context.Background precedent — the actor owns no shutdown path).
		startAppid: func() (AppidSource, error) {
			return appid.Start(context.Background())
		},
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
	a.ensureAppid() // a non-empty app list arrives through this surface too
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

// AppidSource is the per-app identity seam of the MACR layer (ADR-005 a):
// FocusedApp returns the a11y identity of the focused application — the
// live observer of internal/appid or a test double. Defined at the point
// of use; the interface travels with the consumer.
type AppidSource interface {
	FocusedApp() (string, error)
}

// UseClipboard replaces the actor's clipboard client — the test seam of the
// D-28 rung (the daemon keeps the production wl-clipboard client NewActor
// wired).
func (a *Actor) UseClipboard(c *clipboard.Clipboard) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.clip = c
}

// UseAppid installs a prepared identity source — the wiring/test seam of
// the per-app decision (the lazy start never runs afterwards).
func (a *Actor) UseAppid(src AppidSource) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.appid = src
	a.appidStarted = true
}

// UseAppidStarter swaps the lazy-start seam of the per-app observer (the
// production starter dials the a11y bus; the tests install counters and
// failures).
func (a *Actor) UseAppidStarter(fn func() (AppidSource, error)) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.startAppid = fn
}

// RoleSource is the live AT-SPI role seam of the autocorrect policy (plan
// 06-06, D-53 condition 2): Role queries the role of the focused object
// NOW — the live call at the decision moment, never a cached answer
// (ADR-007). The observer of internal/appid implements it; a test double
// stands in for the corpus. Defined at the point of use; the interface
// travels with the consumer (the AppidSource precedent).
type RoleSource interface {
	Role(ctx context.Context) (uint32, error)
}

// UseRole installs a prepared role source — the wiring/test seam of the
// D-53 role contour (the UseAppid mirror). The daemon path never calls it:
// the observer arrives through ensureAppid's lazy start and the RoleSource
// type assertion.
func (a *Actor) UseRole(rs RoleSource) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.role = rs
}

// SetVersion pins the daemon's build identity into the status snapshot
// (D-37): the daemon calls it once at construction with its stamped (or
// dev-fallback) version, and StatusSnapshot lifts it into the report.
func (a *Actor) SetVersion(v string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.version = v
}

// MACRCounters snapshots the Super→Ctrl layer's counters — the goswitchctl
// status surface of plan 03-06 (ADR-005 b.2: every intercepted combination
// and every upstream-consumed one is countable, never silent).
func (a *Actor) MACRCounters() MACRStats {
	a.mu.Lock()
	defer a.mu.Unlock()

	return MACRStats{SuperIntercepted: a.macrIntercepted, ConsumedUpstream: a.macrConsumed}
}

// AutoCorrectCounters snapshots the autocorrect layer's counters (plan
// 06-06, D-54 — the MACRCounters form): fired decisions, fail-closed
// abstentions and their closed-slug reasons. Counts and slugs only — the
// word itself never leaves the actor (D-20/D-21, Pitfall 6).
func (a *Actor) AutoCorrectCounters() AutoCorrectStats {
	a.mu.Lock()
	defer a.mu.Unlock()

	return AutoCorrectStats{Fired: a.acFired, Abstained: a.acAbstained, Reasons: maps.Clone(a.acReasons)}
}

// Status is the daemon state snapshot for the control surface (INST-02):
// the build identity (D-37), the internal mode, the correction outcome
// counters with their skip-reason breakdown, the MACR interception counters
// and the config-source status (D-32 last-good visibility). Counts and
// states ONLY — never typed or corrected text (D-20, T-03-06-03).
type Status struct {
	Version               string
	Mode                  string
	Engine                string // the active engine's wire name (D-52, D-20: a config literal)
	CorrectionsDone       int
	CorrectionsSkipped    int
	SkipReasons           map[string]int
	SuperIntercepted      int
	SuperUpstreamConsumed int
	ConfigPath            string
	ConfigValid           bool
	ConfigError           string
}

// configStatus is the optional status surface of a config source: the
// served document's path and the last rejected reload's error. The 03-02
// watcher implements it; a source without it reports no config fields.
type configStatus interface {
	ConfigPath() string
	LastError() error
}

// StatusSnapshot returns the daemon state for goswitchctl status (INST-02):
// filled under the mutex from the counters the actor already keeps — never
// a second FSM, never the buffer's contents (T-03-06-03). The config fields
// come from the attached source's optional status surface (the watcher);
// with no source the built-in defaults are in force and the snapshot says
// so with an empty path.
func (a *Actor) StatusSnapshot() Status {
	a.mu.Lock()
	defer a.mu.Unlock()

	st := Status{
		Version:               a.version,
		Mode:                  "en",
		Engine:                engineNameOf(a.mode),
		CorrectionsDone:       a.corrDone,
		CorrectionsSkipped:    a.corrSkipped,
		SkipReasons:           maps.Clone(a.skipReasons),
		SuperIntercepted:      a.macrIntercepted,
		SuperUpstreamConsumed: a.macrConsumed,
		ConfigValid:           true, // built-in defaults, or a source without a status surface
	}
	if a.mode == modeRU {
		st.Mode = "ru"
	}
	if cs, ok := a.cfgSrc.(configStatus); ok {
		st.ConfigPath = cs.ConfigPath()
		if err := cs.LastError(); err != nil {
			st.ConfigValid = false
			st.ConfigError = err.Error()
		}
	}

	return st
}

// correctStartedReply is the forced correction's immediate acknowledgment:
// the pipeline is armed — its settlement (the ADR-004 verification, the
// ladder execution, the D-20 refusals) is asynchronous and lands in the
// status counters and the log records, exactly as for a tap-launched
// correction.
const correctStartedReply = "correction started (word pipeline armed; the outcome lands in the counters and the log)"

// CorrectNow forces the word-correction pipeline — the control surface's
// forced correction (INST-02, the Q5 word semantics): the same internal
// point the Double decision dispatches (startCorrection), launched with no
// tap and no FSM round trip, under the actor's mutex with one config
// snapshot folded in. The reply is immediate (see correctStartedReply).
func (a *Actor) CorrectNow() string {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.applySnapshot()
	a.startCorrection()

	return correctStartedReply
}

// HandleKey implements engine.EventHandler: the decoded event is fed into
// the FSM under the mutex and a release of the CONFIGURED tap key re-arms
// the deadline timer. Decisions never fire here — only at window expiry
// (D-04). The returned
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
		if ev.Keyval == a.tapKeyval {
			// Only a release of the CONFIGURED tap key can move the series
			// deadline (the FSM counts taps on clean releases); re-arming
			// on any other event would either extend the deadline from the
			// wrong instant or arm a timer over a cancelled series. A stale
			// expiry is a no-op in the FSM, so over-arming is harmless,
			// under-arming would silently drop the decision.
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
		if kind == engine.LifecycleFocusIn && a.eng != nil {
			// The panel indicator self-heals on every focus gain (owner
			// decision 1, quick plan 260927-way): engine objects are minted
			// per input context, so a freshly minted context re-asserts the
			// CURRENT mode symbol — it must not resurrect the factory's
			// initial EN registration while the actor sits in another mode.
			a.eng.UpdateModeSymbol(a.modeSymbol())
		}
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
// followed by an automatic repair. The mismatch verdict arms the clipboard
// rung OUTSIDE the mutex (WR-01): the rung's three 1.5 s subprocesses must
// never stall the keystroke path, so the locked core below only hands the
// rung payload to the caller.
func (a *Actor) HandleSurroundingText(text string, cursorPos, anchorPos uint32) {
	if paste, armed := a.handleSurroundingLocked(text, cursorPos, anchorPos); armed {
		a.runClipboardRung(paste)
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

// SetSwitcher installs the generation-scoped SetGlobalEngine seam (D-52) —
// the BindSwitcher mirror of AttachEngine: the daemon wiring calls it before
// engine.Run, and every reconnecting generation replaces the closure. A nil
// sw is the no-seam state — the internal flip with its one-WARN degradation.
func (a *Actor) SetSwitcher(sw func(ctx context.Context, engineName string) error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.switcher = sw
	a.switcherWarned = false // a fresh generation opens a fresh degradation episode
}

// ModeDisplay is one mode-observer slot: the tray indicator (the SNI item)
// is the v0 implementation. The observer fires LAST in every mode record —
// mode record → switcher → panel symbol → display (the D-36 order with the
// display appended) — synchronously under the actor's mutex; the
// implementation must stay quick (a pointer swap plus a queued signal, no
// round trip) and must never panic or block.
type ModeDisplay interface {
	ModeChanged(symbol string)
}

// SetModeDisplay installs the display seam — the SetSwitcher mirror. The
// display immediately receives the CURRENT mode: the startup install shows
// the initial mode, and a late install (after flips or syncs) self-syncs to
// the factual state instead of waiting for the next change.
func (a *Actor) SetModeDisplay(md ModeDisplay) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.display = md
	if md != nil {
		md.ModeChanged(a.modeSymbol())
	}
}

// SyncEngine pulls the daemon under the FACTUAL active engine (05-04,
// criterion 3): every GlobalEngineChanged on the daemon's ibus connection
// and every FocusIn lands here with the observed wire name. The enum is
// closed: goswitch-en → EN, goswitch-ru → RU. A name matching the current
// mode is the daemon's OWN flip echoed back (spike P4: the initiator
// receives its own signal) — a confirmation without a record. A drift
// corrects the internal mode with the byte-stable mode record (the e2e
// oracle) plus a WARN. Anything else is a foreign engine — the honest exit
// from under goswitch (a third source or another IME) — WARNed with the
// state untouched (T-05-04-03).
//
// SyncEngine NEVER calls the switcher and NEVER flips the bus: the
// correction moves the internal mode only, so the daemon stays the single
// writer of the active source — a flip loop (daemon flips → echo →
// correction → flip) is impossible by construction (T-05-04-01: goswitch
// flips, the shell's indicator clicks are followed, never fought). The
// caller runs off the key path (the signal dispatcher, engine D-Bus
// handlers); the mutex hold is record-keeping only.
func (a *Actor) SyncEngine(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch name {
	case engine.NameEN:
		a.syncMode(modeEN, name)
	case engine.NameRU:
		a.syncMode(modeRU, name)
	default:
		slog.Warn("foreign engine", "engine", name)
	}
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
			a.flipTo(oppositeMode(a.mode))
		case hotkey.Triple:
			a.startPhraseCorrection()
		}
	}
}

// VerifyExpiry is the verify-deadline timer callback of the pre-correction
// round: the verify budget closed without a fresh SetSurroundingText
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
	a.skipCorrection("verify-timeout")
	a.settleCombo() // the round closed by timeout — the combo's flip still fires
}

// ToggleMode flips the script mode through the SAME flipTo execution path
// as every other gesture (quick plan 261001-fg3) — the public entry the
// tray's interactive surface lands in: the menu's first item and the item's
// SNI Activate. The target is oppositeMode(a.mode), so flipTo's same-target
// guard is unreachable by construction; everything else — the byte-stable
// mode record, the switcher leg, the panel symbol, the display observer
// last (D-36) — comes from flipTo untouched.
func (a *Actor) ToggleMode() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.flipTo(oppositeMode(a.mode))
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
	a.flipTo(oppositeMode(a.mode))
}

// settleCorrectionFlip applies owner decision 2 (quick plan 260927-vu8,
// 2026-09-27): after a successful correction that CHANGED the text, the
// internal script mode flips — the mode follows the correction (the owner
// works in a correct-then-type flow). Exactly two sites call it — the
// word/phrase level-1 success (executeCorrection) and the level-2 success
// (executeLevel2) — each AFTER logCorrectionDone and BEFORE settleCombo,
// so the settled done record precedes every flip (the D-36 order extended)
// and the comboPending guard reads the flag before settleCombo clears it.
// The helper no-ops when the gesture was a combo (the combo's own
// settleCombo flip covers it — exactly one flip per gesture) and when
// correction.flip_after_correction is off (the zero-value Options of the
// unit corpus and the no-SetOptions path stays off; the daemon's built-in
// defaults feed it ON — CorrectNow's forced word pipeline flips on success
// exactly like the double tap it reuses) and after a settled CHANGED
// selection correction (owner report 2026-09-28: «при коррекции выделения
// — нет» fixed). The D-24 done-without-change outcome and the
// skipCorrection refusals never reach it. The caller holds the mutex.
func (a *Actor) settleCorrectionFlip(converted []rune) {
	if a.comboPending || !a.opts.FlipAfterCorrection {
		return
	}
	// The mode is SET to the converted text's script, never toggled (owner
	// rule 2026-09-28): a cyr→lat correction leaves the layout Latin, a
	// lat→cyr one makes it Cyrillic — the layout you now intend to type in.
	// In the two-engine form this makes the RESULT script's ENGINE active
	// (ADR-006 criterion 4); flipTo's same-target guard keeps a same-script
	// correction a no-op.
	a.flipTo(scriptOf(converted))
}

// scriptOf reports the script of the first letter of the converted range —
// the layout the corrected text now lives in. A letterless range (digits,
// punctuation only) keeps the current mode.
func scriptOf(converted []rune) scriptMode {
	for _, r := range converted {
		if unicode.Is(unicode.Cyrillic, r) {
			return modeRU
		}
		if unicode.Is(unicode.Latin, r) {
			return modeEN
		}
	}

	return modeEN // letterless — the caller's transition is false either way
}

// skipCorrection records one D-20 refusal: the INFO reason record (one
// line, the reason slug, nothing touched) and the counters the control
// surface reports. The caller holds the mutex.
func (a *Actor) skipCorrection(reason string) {
	a.corrSkipped++
	if a.skipReasons == nil {
		a.skipReasons = make(map[string]int)
	}
	a.skipReasons[reason]++
	slog.Info("correction skipped", "reason", reason)
}

// countCorrectionDone records one completed correction — the D-24
// success-without-change counts as done too (outcome done either way).
// The caller holds the mutex.
func (a *Actor) countCorrectionDone() {
	a.corrDone++
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
	if w := time.Duration(snap.Timeouts.VerifyWaitMs) * time.Millisecond; w != a.verifyWait && w > 0 {
		// CR-02: the ADR-004 budget is document state — an already-armed
		// deadline keeps its own budget (the Pitfall-8 rule), the next
		// round arms at the new one. The >0 guard keeps a zero-value
		// document on the last-good budget (validation already enforces
		// (0, 2000]).
		a.verifyWait = w
	}
	if name := snap.Hotkeys.TapKey; name != a.tapKeyName {
		// CR-01: the tap key is wiring-live, not validation-only — a
		// changed name re-resolves through the same last-good discipline
		// as the combo below: a name that fails to parse (impossible from
		// a validated document) keeps the last-good key.
		if binding, err := hotkey.ParseBinding(name); err == nil {
			a.tapKeyName = name
			a.tapKeyval = binding.Keyval
			a.fsm.SetTapKey(binding.Keyval)
		}
	}
	a.opts.BackspaceCap = snap.Correction.BackspaceCap
	a.opts.ClipboardRung = snap.Correction.ClipboardRung
	a.opts.FlipAfterCorrection = snap.Correction.FlipAfterCorrection
	if name := snap.Hotkeys.WordLayoutCombo; name != a.comboName {
		if binding, err := hotkey.ParseBinding(name); err == nil {
			a.opts.WordLayoutCombo = binding
			a.comboName = name
		}
	}
	if name := snap.Hotkeys.ModeSwitchChord; name != a.chordName {
		// The EMPTY-DISABLES twist: "" is the chord's off state, not a
		// failed parse — the fold must clear the binding on the empty
		// document value (never keep last-good), while a NON-empty name
		// that fails to parse — impossible from a validated document —
		// keeps the last-good chord (the D-32 discipline).
		if name == "" {
			a.chordName = name
			a.opts.ModeSwitchChord = hotkey.Binding{}
		} else if binding, err := hotkey.ParseBinding(name); err == nil {
			a.opts.ModeSwitchChord = binding
			a.chordName = name
		}
	}
	a.opts.MACREnabled = snap.MACR.Enabled
	a.opts.MACRApps = snap.MACR.Apps
	a.opts.MACRAltModifier = snap.MACR.AltModifier
	if snap.MACR.Letters != a.macrLettersName {
		a.macrLettersName = snap.MACR.Letters
		a.opts.MACRLetters = parseMACRLetters(snap.MACR.Letters)
	}
	a.ensureAppid() // the per-app list may have appeared with this document
}

// handleSurroundingLocked is the mutex-held core of HandleSurroundingText.
// It returns the rung payload and whether the mismatch verdict ARMED the
// rung (ClipboardRung on, a client and an engine present) — the caller then
// runs the subprocess sequence outside the mutex (WR-01), against its own
// snapshots.
func (a *Actor) handleSurroundingLocked(text string, cursorPos, anchorPos uint32) ([]rune, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.applySnapshot() // one config read per event (CONF-02, Pattern 2)

	a.surr = beforeCursor(text, cursorPos)
	a.sel = selectionState{full: []rune(text), cursor: cursorPos, anchor: anchorPos}
	// -debug only, same sensitivity class as the key trace (diagnosis mode
	// is opt-in and short-lived); the live hunt needs the pushed field
	// text to see what a correction actually left behind (2026-09-28).
	slog.Debug("surrounding push", "text", text,
		"cursor", cursorPos, "anchor", anchorPos)
	if a.pending != nil {
		if !a.pendingVerdict() {
			a.resolvePending()
			a.skipCorrection("verify-mismatch")
			a.settleCombo() // the word half settled by refusal — the D-36 flip still fires

			return nil, false
		}
		a.executeCorrection()

		return nil, false
	}
	if a.after != nil {
		return a.settleAfter()
	}

	return nil, false
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

// settleAfter settles the open verify-after round on the fresh push (the
// caller holds the mutex and a.after != nil). The WR-05 debounce: the
// FIRST mismatching answer re-requires instead of concluding — a client
// may push the INTERMEDIATE post-delete pre-commit state (the live finding
// of the first matrix runs), so one stale answer proves nothing; the
// second stale answer concludes, and a matching one settles quietly. This
// is what keeps a false mismatch from routing the clipboard rung into a
// duplicate paste of the just-applied correction. Returns the rung payload
// and whether the mismatch verdict ARMED the rung.
func (a *Actor) settleAfter() ([]rune, bool) {
	round := a.after
	if a.afterVerdict(round.expected, round.start, round.atRange) {
		a.clearAfter()
		slog.Debug("correction verify", "outcome", "match")

		return nil, false
	}
	if round.mismatches == 0 {
		round.mismatches = 1
		a.rearmAfter(round)
		slog.Debug("correction verify", "outcome", "stale-retry")
		a.eng.RequireSurroundingText()

		return nil, false
	}
	a.clearAfter()
	slog.Info("correction verify", "outcome", "mismatch")

	return round.paste, a.opts.ClipboardRung && a.clip != nil && a.eng != nil
}

// rearmAfter re-arms the round's deadline for the debounce retry (the
// caller holds the mutex): the same verify budget and the SAME epoch — an
// expiry from a superseded round stays a no-op, and a client that never
// answers settles as the quiet timeout (no verdict, no rung).
func (a *Actor) rearmAfter(round *pendingAfter) {
	if round.deadline != nil {
		round.deadline.Stop()
	}
	epoch := round.epoch
	round.deadline = time.AfterFunc(a.verifyWait, func() { a.afterExpiry(epoch) })
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
// ends with it and lands in the INFO mismatch counter (Pitfall 3) — but
// only after the WR-05 debounce: the FIRST mismatching answer re-requires
// (it may be the intermediate post-delete pre-commit state), the second
// concludes.
func (a *Actor) armAfterVerify(converted, tail []rune) {
	a.verifyEpoch++
	a.clearAfter() // never two open rounds
	a.after = &pendingAfter{
		expected: concatRunes(converted, tail),
		paste:    converted,
		epoch:    a.verifyEpoch,
	}
	epoch := a.verifyEpoch
	a.after.deadline = time.AfterFunc(a.verifyWait, func() { a.afterExpiry(epoch) })
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
	a.after.deadline = time.AfterFunc(a.verifyWait, func() { a.afterExpiry(epoch) })
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
// only runs armed (the caller checked ClipboardRung and the snapshots).
//
// WR-01: the rung runs WITHOUT the actor mutex — up to three 1.5 s
// subprocesses must never stall the <50 ms keystroke path that serializes
// on it. The rung touches no FSM/buffer state: it works on the payload and
// its own emitter/clipboard snapshots, and takes a.mu only for the brief
// counter updates (recordSkip). The emitter calls are wire-safe off-mutex
// (godbus Emit is goroutine-safe). rungMu serializes the rungs themselves:
// a second mismatch concluding while one rung is mid-flight skips with the
// busy reason — never blocks, never interleaves a second wl-copy/wl-paste
// pair into the first round-trip.
func (a *Actor) runClipboardRung(paste []rune) {
	// The rung's own snapshots, read under a brief lock (the verdict armed
	// it under the same guard one call earlier; a mid-window engine swap
	// only retargets the burst to the live sink).
	a.mu.Lock()
	eng, clip := a.eng, a.clip
	a.mu.Unlock()
	if eng == nil || clip == nil {
		return
	}
	if !a.rungMu.TryLock() {
		a.recordSkip("clipboard-rung-busy")

		return
	}
	defer a.rungMu.Unlock()

	// The clipboard client bounds every subprocess with its own deadline
	// (T-03-03-05) — a background context is the rung's lifetime.
	ctx := context.Background()
	saved, had, err := clip.Save(ctx)
	if err != nil {
		a.recordSkip("clipboard-unavailable")

		return
	}
	if err := clip.Set(ctx, []byte(string(paste))); err != nil {
		a.recordSkip("clipboard-unavailable")

		return
	}
	forwardCtrlV(eng)
	if err := clip.Restore(ctx, saved, had); err != nil {
		// D-29: best-effort — the replacement already landed in the field;
		// a failed restore is logged (error only, never content), not raised.
		slog.Warn("clipboard restore failed", "error", err)
	}
}

// recordSkip records one D-20 refusal from the mutex-free rung: a brief
// re-acquisition of the actor mutex around the counter (skipCorrection's
// contract — the caller holds the mutex).
func (a *Actor) recordSkip(reason string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.skipCorrection(reason)
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
// focus: an empty app list is the GLOBAL rule (ADR-005 a); a non-empty
// list needs the focused application to be IN it — the identity source
// decides, and its unavailability (never started, or erroring) degrades
// to the GLOBAL rule with the WARN "app identity unavailable" (the
// ADR-005 ladder: the observer's failure never switches the layer off).
// A source that has not yet seen a focus event answers "" — the per-app
// scope stays conservative (transit) until an identity is known. The
// caller holds the mutex.
func (a *Actor) macrTargetActive() bool {
	if len(a.opts.MACRApps) == 0 {
		return true
	}
	if a.appid == nil {
		a.warnAppid(nil)

		return true // the degradation rung: the global rule stays in force
	}
	app, err := a.appid.FocusedApp()
	if err != nil {
		a.warnAppid(err)

		return true // the degradation rung (ADR-005 ladder)
	}
	a.appidWarned = false // a healthy answer closes the degradation episode
	if app == "" {
		return false // identity not observed yet — the per-app scope is conservative
	}

	return slices.Contains(a.opts.MACRApps, app)
}

// warnAppid records one identity-source failure: the WARN fires once per
// degradation episode (per keystroke would spam the log of a broken
// source). The caller holds the mutex.
func (a *Actor) warnAppid(err error) {
	if a.appidWarned {
		return
	}
	a.appidWarned = true
	if err == nil {
		slog.Warn("app identity unavailable")

		return
	}
	slog.Warn("app identity unavailable", "error", err)
}

// ensureAppid lazily starts the per-app identity source: only once a
// NON-EMPTY app list is in force, only once per actor (ADR-005 a: the
// a11y bus is connected IFF per-app lists exist). A failed start WARNs
// and leaves the source nil — the global rule covers the gap for good
// (one documented degradation, not a retry loop). The caller holds the
// mutex.
func (a *Actor) ensureAppid() {
	if a.appidStarted || (len(a.opts.MACRApps) == 0 && len(a.opts.AutoCorrectApps) == 0) {
		return
	}
	a.appidStarted = true
	src, err := a.startAppid()
	if err != nil {
		slog.Warn("app identity unavailable", "error", err)

		return
	}
	a.appid = src
	if rs, ok := src.(RoleSource); ok {
		a.role = rs // one Observer, two seams (the plan-06-06 role contour)
	}
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

// flipTo executes one flip to the target mode — the single execution path of
// EVERY flip gesture (the Single decision at window expiry, the combo's
// settleCombo, the correction flip's SET semantics via settleCorrectionFlip,
// the mode-switch chord). The ADR-001 Option B internal flip became the
// ADR-006 engine-truth flip (D-52): the internal mode stands switched
// IMMEDIATELY — it stays the single writer of the script state and the RU
// commit branch keeps producing the script — and the actual engine switch
// leaves through the SetGlobalEngine seam, so the next keystroke enters the
// new engine and GNOME sees a first-class switch. The record order is the
// D-36 extension, pinned by the corpus: the byte-stable mode record first
// (the e2e oracle — the stand waits for it after a single tap before typing
// in the new script, Pitfall 5), then the switcher call, then the
// panel-symbol update — exactly one UpdateModeSymbol per flip, strictly
// after the log record (owner decision 1, quick plan 260927-way).
//
// The switcher call runs under a.mu (every caller holds it) with a hard
// switchTimeout deadline — the deliberate choice of the two mechanisms the
// phase research offered (deadline OR off-mutex handoff, Pitfall 4 /
// T-05-03-01): a wedged ibus-daemon can cost a flip at most switchTimeout,
// never an unbounded keystroke stall, because the real seam honors the
// context deadline (the godbus CallWithContext round trip). The deadline is
// the wedge guard only — it sits above the measured live round trip
// (~41-45 ms), so it fires on a wedged bus, never on a healthy flip. The async
// WR-01-rung alternative was rejected for THIS path: the flip must be
// synchronous so the record order stays deterministic and rapid flips keep
// the final bus state equal to the final mode. A failure WARNs with the
// engine name — the internal flip already stands, typing continues in the
// new mode (criterion 3); a nil seam degrades to the internal flip with
// exactly one WARN per episode. The caller holds the mutex.
//
// A flipTo to the CURRENT mode is a no-op: the SET-semantics site
// (settleCorrectionFlip) names its target from the corrected text's script,
// and a same-script correction flip changes nothing.
func (a *Actor) flipTo(target scriptMode) {
	if target == a.mode {
		return
	}
	a.mode = target
	slog.Info("mode", "to", a.modeSymbol())

	if sw := a.switcher; sw != nil {
		ctx, cancel := context.WithTimeout(context.Background(), switchTimeout)
		defer cancel()
		if err := sw(ctx, engineNameOf(target)); err != nil {
			slog.Warn("engine switch failed", "engine", engineNameOf(target), "error", err)
		}
	} else if !a.switcherWarned {
		a.switcherWarned = true
		slog.Warn("switcher unavailable", "engine", engineNameOf(target))
	}

	if a.eng != nil {
		a.eng.UpdateModeSymbol(a.modeSymbol())
	}

	// The display observer fires LAST — after the panel symbol, the D-36
	// order with the display appended (quick 260930-pf6). The indicator
	// degrades itself; the actor adds no error handling around the call.
	if a.display != nil {
		a.display.ModeChanged(a.modeSymbol())
	}
}

// oppositeMode is the toggle target of the gesture flips (the Single
// decision, the combo's settle, the mode-switch chord); the SET-semantics
// site names its target directly (settleCorrectionFlip).
func oppositeMode(m scriptMode) scriptMode {
	if m == modeEN {
		return modeRU
	}

	return modeEN
}

// engineNameOf returns the wire name of a script mode's engine (D-52): the
// mode's engine IS the flip target; the names are the engine package's
// literals — config constants, never user data (D-20).
func engineNameOf(m scriptMode) string {
	if m == modeRU {
		return engine.NameRU
	}

	return engine.NameEN
}

// syncMode applies one observed goswitch engine name under the mutex: the
// same mode is a silent confirmation (the own-flip echo); a drift records
// the byte-stable mode record first, the correction WARN second, then the
// panel symbol follows the corrected mode — the flipTo order without the
// switcher leg.
func (a *Actor) syncMode(target scriptMode, name string) {
	if target == a.mode {
		return
	}
	a.mode = target
	slog.Info("mode", "to", a.modeSymbol())
	slog.Warn("mode corrected", "engine", name)
	if a.eng != nil {
		a.eng.UpdateModeSymbol(a.modeSymbol())
	}
	// The display observer fires last, mirroring flipTo (quick 260930-pf6):
	// the icon follows the FACTUAL engine, not only the daemon's own flips.
	if a.display != nil {
		a.display.ModeChanged(a.modeSymbol())
	}
}

// modeSymbol returns the panel symbol of the current script mode — the
// glyph the mode-indicator property carries (owner decision 1). The caller
// holds the mutex.
func (a *Actor) modeSymbol() string {
	if a.mode == modeRU {
		return "ru"
	}

	return "en"
}

// feedKey decides one press: whether the engine consumes the key and which
// rune the buffer takes — the script-true invariant made branch-local (a
// rune that reaches the field also reaches the buffer, whatever delivered
// it). The branch order is the branch's contract. The MACR interception
// (MACR-01, ADR-005, Pattern 6) is recognized FIRST of all, above the
// chords and every mode branch: a configured letter press carrying Mod4 —
// the press-side wire truth of the live probe: the letter arrives Latin
// with Mod4 in EVERY internal mode — is consumed and replayed as the
// Ctrl+letter forward burst, so the RU commit can never fire on a Super
// chord and the buffer is never fed. The word-layout combo (D-36) is
// recognized next, above every mode branch: a press of the bound combo key
// under its bound HELD modifiers — the press's state word carries only the
// modifiers held before the key (live finding 2026-09-15: a Control_R
// press under Shift arrives with Shift|NumLock, its own Control bit rides
// only on the release), so the match compares Binding.ModMask with the
// key's own family bit cleared — latch-tolerant through &, the NumLock
// precedent of 02-04 — kills the tap series with a deliberate Reset
// (Pitfall 4: left to the FSM the Control_R press would silently die as
// modifier use) and launches the word pipeline of the Double semantics
// with the flip deferred to its settlement (comboPending/settleCombo).
// The combo press itself transits: a bare modifier chord puts no rune in
// the field, and the transit keeps the client's press/release pairing
// intact. The mode-switch chord (owner decision 2, quick plan 260927-way)
// sits AFTER the combo (the combo wins a collision) and BEFORE the mode
// branches: a press of the bound chord key under its held modifiers flips
// the script mode IMMEDIATELY, consumes the press (no space lands in the
// field), kills any pending tap series, and marks the Super hold as
// WITNESSED — goswitch itself consumed the chord, nothing went upstream,
// so the MACR consumed-upstream detect stays silent (the MACR layer is
// unaffected: space is not a letter, macrIntercept owns the hold window).
// The CORR-09 reset keyvals (Enter and its keypad variant, Tab, Escape)
// end the phrase instead of feeding it. The caller holds the mutex.
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

	if a.modeSwitchChord(ev) {
		return true
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
		// The reset is ALSO a word boundary (plan 06-06): the autocorrect
		// decision arms BEFORE the reset reads the finished word, and the
		// arming never touches the transit or the reset semantics. The
		// FocusOut lifecycle is NOT a boundary — the word was not finished
		// by a keystroke (06-RESEARCH Q5) — so autocorrect stays silent
		// there.
		a.armAutoCorrect()
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
			a.pushAndArm(ru) // the committed rune is what the field now holds

			return true
		}
		// Identical map (digits, parentheses…) or unmapped key: transit —
		// the client inserts the original rune, so that rune is the field
		// truth the buffer must mirror.
		a.pushAndArm(r)

		return false
	default:
		// EN mode: Phase 1 semantics — transit, buffer fed as typed.
		// #nosec G115 -- printableKeyval bounds the keyval below
		// 0xFE00, so the uint32→rune conversion cannot overflow.
		a.pushAndArm(rune(ev.Keyval))

		return false
	}
}

// pushAndArm feeds one typed rune to the buffer and arms the autocorrect
// decision at a word boundary (the three Push sites of feedKey — plan
// 06-06). The boundary never changes the caller's consume verdict: the
// rune-feeding semantics of every branch stay byte-identical, and the
// armed decision executes asynchronously (armAutoCorrect).
func (a *Actor) pushAndArm(r rune) {
	if a.buf.PushFeed(r) {
		a.armAutoCorrect()
	}
}

// armAutoCorrect runs the cheap-gate boundary decision and launches the
// confirm for an armed payload (plan 06-06): the boundary's generation is
// tagged under the mutex so a newer boundary supersedes a mid-flight
// confirm, and the confirm goroutine leaves the keystroke path
// immediately — the armed-payload discipline of the clipboard rung
// (T-06-06-04/T-06-06-05). The caller holds the mutex.
func (a *Actor) armAutoCorrect() {
	payload, armed := a.autoCorrectBoundary()
	if !armed {
		return
	}
	a.acGeneration++
	payload.gen = a.acGeneration
	go a.autoConfirm(payload)
}

// autoCorrectBoundary decides one word boundary under the CHEAP half of
// the D-53 conjunction (the caller holds the mutex — feedKey): enabled, a
// non-empty white list, the surrounding-text capability, the focused app's
// exact bridge-namespace identity and the detector's CONFIDENT
// wrong-layout verdict on the finished token. Every refusal counts its
// reason and refuses the payload; any unknown means SILENCE — the
// fail-closed direction INVERTED from macrTargetActive's degradation (no
// rung upward, ADR-007). The expensive half — the live GetRole — never
// runs here: the armed payload hands the decision outside the mutex (the
// off-mutex discipline, T-06-06-04). The off state is the zero-behavior
// invariant: no counter, no record — the boundary is byte-as-today (D-54).
func (a *Actor) autoCorrectBoundary() (acPayload, bool) {
	if !a.opts.AutoCorrectEnabled {
		return acPayload{}, false
	}
	if len(a.opts.AutoCorrectApps) == 0 {
		a.recordACAbstain(acReasonNoApps)

		return acPayload{}, false
	}
	if a.caps&correct.CapSurroundingText == 0 {
		// D-53 condition 3: without the bit the verify rung is not
		// applicable — the word is never touched.
		a.recordACAbstain(acReasonNoCaps)

		return acPayload{}, false
	}
	app, ok := a.acFocusedApp()
	if !ok {
		a.recordACAbstain(acReasonAppUnknown)

		return acPayload{}, false
	}
	// The white list is an EXACT bridge-namespace match: no prefix or
	// suffix merging (org.gnome.ZenityX never matches org.gnome.Zenity).
	if !slices.Contains(a.opts.AutoCorrectApps, app) {
		a.recordACAbstain(acReasonAppNotListed)

		return acPayload{}, false
	}
	tok := a.buf.Token()
	v := detect.Check(tok, a.modeSymbol(),
		detect.Data{RU: layouts.DictRU, EN: layouts.DictEN},
		detect.Trigrams{RU: layouts.TriRU, EN: layouts.TriEN},
		detect.Params{
			MinWordLen:    a.opts.AutoCorrectMinWordLen,
			TrigramMargin: a.opts.AutoCorrectMargin,
			TrigramFloor:  a.opts.AutoCorrectFloor,
		})
	if !v.WrongLayout || !v.Confident {
		// D-53 condition 4 (with the detector's own length gate): an unsure
		// verdict, a dictionary veto or a short token stays untouched — the
		// detector's closed reason slug is the only trace.
		a.recordACAbstain(v.Reason)

		return acPayload{}, false
	}

	return acPayload{
		rng: correctionRange{
			token:   tok,
			tail:    a.buf.Tail(),
			replace: a.buf.ReplaceToken,
		},
		dir:  v.Dir,
		word: tok,
	}, true
}

// acFocusedApp resolves the focused app identity for the boundary gate: a
// missing source is an UNKNOWN (fail-closed — never the MACR degradation),
// reported with the one-WARN-per-episode discipline of warnAppid. The
// caller holds the mutex.
func (a *Actor) acFocusedApp() (string, bool) {
	if a.appid == nil {
		a.warnACApp(nil)

		return "", false
	}
	app, err := a.appid.FocusedApp()
	if err != nil {
		a.warnACApp(err)

		return "", false
	}
	a.acAppWarned = false // a healthy answer closes the episode

	return app, true
}

// warnACApp records one autocorrect identity-source failure: the WARN
// fires once per degradation episode (the warnAppid form — per boundary
// would spam the journal of a broken source). The caller holds the mutex.
func (a *Actor) warnACApp(err error) {
	if a.acAppWarned {
		return
	}
	a.acAppWarned = true
	if err == nil {
		slog.Warn("autocorrect app identity unavailable")

		return
	}
	slog.Warn("autocorrect app identity unavailable", "error", err)
}

// recordACAbstain counts one fail-closed abstention (the skipCorrection
// form): the counter, the closed-slug reason and the INFO record — the
// word never appears on any level (D-20/D-21, Pitfall 6). The caller holds
// the mutex.
func (a *Actor) recordACAbstain(reason string) {
	a.acAbstained++
	if a.acReasons == nil {
		a.acReasons = make(map[string]int)
	}
	a.acReasons[reason]++
	slog.Info("autocorrect skipped", "reason", reason)
}

// autoConfirm executes one armed autocorrect payload — the Task-1
// scaffold: the armed range launches THE one correction pipeline
// (startRangeCorrection — no second replacement mechanism, Pitfall 7) from
// the confirm goroutine, off the keystroke path. A newer boundary
// supersedes a mid-flight confirm by generation: the stale one drops. The
// D-53 role contour (the live GetRole gate spending autoRoleTimeout, the
// fail-closed verdicts, the fired counter) arrives with Task 2.
func (a *Actor) autoConfirm(payload acPayload) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if payload.gen != a.acGeneration {
		return // a newer boundary owns the decision now
	}
	a.startRangeCorrection(payload.rng)
}

// modeSwitchChord is the mode-switch chord branch of feedKey (owner
// decision 2, quick plan 260927-way): a press of the bound chord key under
// its held modifiers — the same press-side held-mask rule as the combo
// above (space has no family bit, so held is exactly the chord's Mod4) —
// flips the script mode IMMEDIATELY and reports the press consumed. The
// Super hold is marked WITNESSED (goswitch consumed the chord itself,
// nothing was swallowed upstream — the MACR consumed-upstream detect must
// not fire on the release), any pending tap series dies with the
// deliberate Reset (Pitfall 4), and the panel symbol rides the flip. A
// zero chord binding (disabled) never matches. The caller holds the mutex.
func (a *Actor) modeSwitchChord(ev engine.EngineEvent) bool {
	c := a.opts.ModeSwitchChord
	if c.Keyval == 0 || ev.Keyval != c.Keyval {
		return false
	}
	if held := c.ModMask &^ hotkey.FamilyMask(c.Keyval); ev.Mods&held != held {
		return false
	}
	a.macrSawLetter = true
	a.fsm.Feed(hotkey.Reset{}, a.elapsed())
	slog.Info("combo", "kind", "mode-switch-chord")
	a.flipTo(oppositeMode(a.mode))

	return true
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
		a.skipCorrection(refusalReason(runes))
		a.settleCombo()

		return
	}
	if !changed {
		// D-24: every letter of the range is already in the anchor layout —
		// a SUCCESSFUL operation without changes; nothing to replace,
		// nothing to verify.
		a.countCorrectionDone()
		slog.Info("correction", "outcome", "done")
		a.settleCombo()

		return
	}
	if a.eng == nil {
		a.skipCorrection("no-engine")
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
	a.pending.deadline = time.AfterFunc(a.verifyWait, a.VerifyExpiry)
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
// a miss does the Require round-trip run inside the verify budget.
func (a *Actor) startRangeCorrection(rng correctionRange) {
	armed := time.Now()
	if len(rng.token) == 0 {
		a.skipCorrection("empty-buffer")
		a.settleCombo() // no word — the flip is the combo's primary intent

		return
	}
	converted, changed, ok := correct.ConvertRuns(rng.token)
	if !ok {
		a.skipCorrection(refusalReason(rng.token))
		a.settleCombo()

		return
	}
	if !changed {
		// D-24: every letter of the range is already in the anchor layout —
		// a SUCCESSFUL operation without changes (the owner's choice over a
		// refusal and over a WARN); nothing to replace, nothing to verify.
		a.countCorrectionDone()
		slog.Info("correction", "outcome", "done")
		a.settleCombo()

		return
	}
	if a.eng == nil {
		a.skipCorrection("no-engine")
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
	a.pending.deadline = time.AfterFunc(a.verifyWait, a.VerifyExpiry)
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
	a.logCorrectionDone(plan.Level, p.rng.token, p.converted, time.Since(p.armed))
	a.settleCorrectionFlip(p.converted) // owner rule 2026-09-28: the mode follows the converted script
	a.settleCombo()                     // D-36: the flip lands strictly after the settled completion record
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
	// NO DeleteSurroundingText here (2026-09-28 fix): the client's
	// selection is ACTIVE — a CommitText replaces an active selection by
	// the OS input-method contract (GTK, Chromium/Electron, Qt alike), so
	// the former delete+commit pair double-applied and corrupted the field
	// around the selection (live hunt 2026-09-28: Chromium answered with a
	// cursor position no correct delete+commit ordering can produce, the
	// verify-after timed out — 25→24-rune correction). Geometry only,
	// never field content (D-20/D-21).
	slog.Debug("selection geometry",
		"cursor", sel.cursor, "anchor", a.sel.anchor,
		"start", sel.start, "end", sel.end,
		"field_len", len(a.sel.full))
	a.eng.CommitText(engine.NewIBusText(string(p.converted)))
	a.buf.HardReset() // the buffer can no longer mirror the replaced field
	a.logCorrectionDone(correct.Level1, p.rng.token, p.converted, time.Since(p.armed))
	a.settleCorrectionFlip(p.converted) // owner rule 2026-09-28: the mode follows the converted script
	a.settleCombo()                     // D-36: the flip lands strictly after the settled completion record
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
		a.skipCorrection("backspace-cap")
		a.settleCombo()

		return
	}
	for range plan.Backspaces {
		a.eng.ForwardKeyEvent(engine.KeyBackSpace, backSpaceKeycode, 0)
	}
	a.eng.CommitText(engine.NewIBusText(string(plan.Commit)))
	rng.replace(converted)
	a.logCorrectionDone(plan.Level, rng.token, converted, time.Since(armed))
	a.settleCorrectionFlip(converted) // owner rule 2026-09-28: the mode follows the converted script
	a.settleCombo()
}

// logCorrectionDone counts and writes the completion pair of one ladder
// execution: the INFO counter (D-20 — outcome only, never the word) and
// the DEBUG detail record whose FIRST attribute after msg is the ladder
// level — the exact form the e2e matrix greps to pin the ACTUAL level a
// client got (D-21). The caller holds the mutex.
func (a *Actor) logCorrectionDone(level correct.Level, token, converted []rune, latency time.Duration) {
	a.corrDone++
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
