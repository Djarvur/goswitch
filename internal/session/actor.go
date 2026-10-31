// Package session hosts the daemon's event consumers: the actor that
// serializes engine events into the hotkey FSM, feeds the typed-phrase
// buffer and runs the two-phase word correction (CORR-01).
package session

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/Djarvur/goswitch/internal/appid"
	"github.com/Djarvur/goswitch/internal/clipboard"
	"github.com/Djarvur/goswitch/internal/config"
	"github.com/Djarvur/goswitch/internal/correct"
	"github.com/Djarvur/goswitch/internal/detect"
	"github.com/Djarvur/goswitch/internal/engine"
	"github.com/Djarvur/goswitch/internal/hotkey"
	"github.com/Djarvur/goswitch/internal/layouts"
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
// keycode half of the level-2 replay's press+release PAIRS
// ForwardKeyEvent(KeyBackSpace, 14, 0)/(KeyBackSpace, 14, MaskRelease)
// (RESEARCH Pattern 1; the keyval half is engine.KeyBackSpace).
const backSpaceKeycode = 14

// terminalChunkKeys is the chunk ceiling of the terminal write branch
// (plan 10-04, TERM-02, ROADMAP criterion 4): the BackSpace replay is
// emitted in bursts of at most this many PAIRS (round-6: one pair = one
// press + one release forward of the same key) with a context re-check
// between chunks — a mid-burst focus loss can never cost more than one
// chunk, and the IME queue's loss window stays bounded (T-10-04-03).
const terminalChunkKeys = 8

// forwardEmissionsPerKey is the round-6 pair arity: one BackSpace key rides
// exactly TWO ForwardKeyEvent emissions — the press and the release that
// disarms the client's own key repeat. The delivery marks report the wire
// emission count (keys×arity) alongside the rune-true key count, and every
// pacing-budget arithmetic (the burst's own quanta, the key-hold bound)
// multiplies by it.
const forwardEmissionsPerKey = 2

// focusFlapWindow is the flap-pair window of the DEFERRED verify-dead wipe
// (plan 10-07, G-10-1 cause A): a real FocusOut whose FocusIn comes back
// within this window is the host kernel's same-engine focus flap — the
// journal of 2026-10-13 16:20:05 shows pairs 1–3 ms apart, four within
// 6 ms — while a real window change costs orders of magnitude more. The
// wipe the pair used to execute immediately waits the window out; the
// completing FocusIn cancels it and the learned terminal class survives
// the burst. An internal constant with a journal provenance, never a
// config key: the flap geometry is a property of the churn defect, not a
// user preference.
const focusFlapWindow = 10 * time.Millisecond

// level1ConfirmBudget is the confirmation budget of the two-phase level-1
// delivery (plan 10-11, round-2 G-10-1, the 09:39 host rows): the level-1
// ladder's commit phase waits this long for the client's proof that the
// deletion applied (a push or a Require answer whose text-before-cursor no
// longer ends with token+tail), then degrades to today's delete+commit pair
// with a loud INFO mark — never a «deletion without replacement». Origin of
// the value: a healthy client's confirming push lands ~16 ms after the
// emission (the E3 -debug host trace of 2026-10-13), and the x4 headroom
// rides out the churn-delayed clients the host journals show; an internal
// constant with a journal provenance, never a config key (the
// flipConfirmGrace precedent) — the fallback guarantees that an oversized
// budget degrades to today's semantics, not to a worse one.
const level1ConfirmBudget = 60 * time.Millisecond

// flipConfirmGrace is the grace deadline of the terminal branch's deferred
// correction flip (plan 10-08, G-10-1 cause B, the E3 verdict of 2026-10-13):
// a healthy confirming push lands ~16 ms after the commit (the E3 -debug
// trace), while the churn-delayed client's next keystroke is unmeasurable
// from the daemon — the deadline carries a x20 headroom, so the mode rule
// can never be lost silently yet never fires into a live delivery wait. An
// internal constant with a journal provenance, never a config key.
const flipConfirmGrace = 300 * time.Millisecond

// flipAftermathWindow is the re-mint-aftermath window of the client Reset
// carve-out (Round 2 of debug session orca-terminal-second-shift): a
// successful SetGlobalEngine makes ibus re-mint engine objects, and the slow
// client (Electron ibus-IM, the journal's Orca IDE terminal) reconnects its
// IM context ASYNCHRONOUSLY, delivering a Reset on the re-minted engine
// object well after the synthetic focus pair — the green106 acceptance row
// of 2026-10-31 13:55 shows the echo's wipe 923 ms after the flip (the
// machine-verified journal elimination: all 22 wipe rows of the day ran the
// Reset branch). A Reset inside the window is the re-mint's own echo — the
// same client field re-engaging, not a context change — and keeps the
// learned class (the synthetic-FocusOut discipline); past the window a
// Reset is REAL and wipes immediately, byte as before (fail-closed for a
// genuine context change). The value doubles the observed ~0.9 s reconnect
// for slower clients and stays an internal constant with a journal
// provenance, never a config key (the focusFlapWindow precedent).
const flipAftermathWindow = 2 * time.Second

// defaultLevel2ForwardGap is the DEFAULT pacing quantum between the write
// branch's wire emissions (plan 10-05, the 2026-10-12 VTE live finding):
// the client drains forwarded keys lazily — they queue as GDK events —
// while the commit lands synchronously, so a tight burst reorders and
// corrupts the line («ghbdtn» came back «ghприв»: four erases, the
// commit, two late erases). A sleep before EVERY forward and before the
// commit lets the client's queue drain between emissions; the order on
// the wire and the chunk geometry stay untouched (the op-log pins hold).
// The value lives on the actor (NewActor plants the default) so the
// -debug stand matrix can override it (plan 10-17 Task 3) — never from
// config.
const defaultLevel2ForwardGap = 20 * time.Millisecond

// defaultTerminalDrainSettle is the DEFAULT bounded delivery window the
// terminal branch holds between its last BackSpace forward and the commit
// (plan 10-17, the round-4 machine). Origin (the journal rows, never a
// config key — the level1ConfirmBudget precedent): the 19:21:26 host row
// (DEBUG-term-host-live-failures.md, round-2 block) reports done +144 ms
// with the whole 6×20 ms package behind it — the host client was STILL
// DRAINING its forward queue when the commit landed; the healthy client's
// confirming push (the E3 -debug trace of 2026-10-13) lands ~16 ms after
// the emission. The window gives the queue an honest term to drain before
// the replacement is committed; past the window the machine claims NO
// persistence — every outcome beyond it carries the loud delivery marks
// instead of a field promise (the honest uncertainty of the round-4
// mandate). The value lives on the actor (NewActor plants the default) so
// the -debug stand matrix can override it (plan 10-17 Task 3) — never
// from config.
const defaultTerminalDrainSettle = 150 * time.Millisecond

// keyHoldSlack is the scheduling slack of the key-hold bound (CR-02): the
// parked keystroke's wait is terminalDrainSettle + the machine's whole
// pacing budget (cap×2×gap — round-6: press AND release quanta per key; a
// keystroke may catch the FIRST chunk-boundary Gosched window of a
// maximum-cap burst) plus this quantum for the commit's
// own bookkeeping and the goroutine handoffs. A bound, never an eternity:
// past it the hold degrades loudly to arrival processing (keys are never
// lost — the owner's typing is sacred), accepting the pre-fix inversion for
// that pathological overrun only.
const keyHoldSlack = 250 * time.Millisecond

// switchTimeout bounds the SetGlobalEngine round trip on the flip path —
// the FAILURE-path wedge guard (Pitfall 4, T-05-03-01), not a slice of the
// <50 ms SPEC reaction budget: it bounds only how long a wedged ibus-daemon
// may hold the flip. The measured success RTT is ~41-45 ms (the 05-05 live
// proofs and the 2026-09-30 journal — the old 40 ms deadline fired before
// completion on every healthy flip), so 150 ms sits above the measured
// maximum: every healthy flip completes inside it, and a wedged bus still
// costs at most the deadline. G-6-1: the deadline stays that wedge guard —
// it is deliberately NOT the fix for the factory self-deadlock; the
// re-entrant CreateEngine path was freed separately, by taking AttachEngine
// off the actor mutex (the lock-free emitter slot below), never by trimming
// or moving this deadline.
const switchTimeout = 150 * time.Millisecond

// autoRoleTimeout bounds the live GetRole round trip of one autocorrect
// confirm (plan 06-06, Pitfall 5): the decision runs OUTSIDE the actor
// mutex with its own deadline, so a stuck a11y object can never stall the
// keystroke path. The measured unix-socket RTT is sub-millisecond — the
// deadline is a wedge guard, budgeted at half of SPEC §5's 50 ms reaction
// ceiling.
const autoRoleTimeout = 25 * time.Millisecond

// acRoleWarnRearm is the re-arm window of the sticky role-source WARN
// (plan 10-07, G-10-1 missing #3): the 2026-10-13 host journal shows the
// 11:00:16 episode consuming the one WARN and staying silent the WHOLE DAY
// while the source stayed broken (autocorrect_fired=0 all day) — a lasting
// outage must not dissolve into a single absorbed line. Errors continuing
// past half an hour sound the WARN again (one line per window, never one
// per query — a flap-bursting source keeps its journal clean); every error
// between the lines still grows the status-visible counter.
const acRoleWarnRearm = 30 * time.Minute

// The AT-SPI text-input roles the D-53 role gate allows: text box (61),
// entry (79) and document text (94) — the live-verified enum of
// internal/appid (gi Atspi 2.52.0, 06-RESEARCH Q4). password-text (40),
// terminal (60) and every other value refuse. The gate compares NUMBERS
// only: role-name strings localize ("text box" vs the gi-nick "text").
// acRoleTerminal is the terminal class of the D-55 veto (Phase 10) — the
// first source of the classification seam (internal/appid RoleTerminal).
const (
	acRoleText         uint32 = 61
	acRoleEntry        uint32 = 79
	acRoleDocumentText uint32 = 94
	acRoleTerminal     uint32 = 60
)

// The autocorrect abstention vocabulary (plan 06-06, D-53 fail-closed):
// every refusal of the boundary/confirm conjunction counts exactly one of
// these slugs; the arming gate additionally reuses the detector's own
// closed reasons (internal/detect). Closed and word-free (D-20/D-21,
// Pitfall 6).
const (
	acReasonAppBlocked = "app-blocked"
	acReasonNoCaps     = "no-caps"
	// acReasonRoleAmbiguous is the CR-01 tightening (07-REVIEW, the
	// owner-sanctioned narrow exception to the unknown-passes rule): the
	// ambiguous role 61 answered while the identity is UNKNOWN.
	acReasonRoleAmbiguous   = "role-ambiguous"
	acReasonRoleUnknown     = "role-unknown"
	acReasonRoleForbidden   = "role-forbidden"
	acReasonRoleTimeout     = "role-timeout"
	acReasonRecheckDisabled = "recheck-disabled"
	acReasonPayloadStale    = "payload-stale"
	// acReasonTerminalVeto is the arming veto of the terminal class
	// (Phase 10, TERM-04/D-55): the boundary refuses before the detector
	// runs — the invariant's literal placement, never the blocklist.
	acReasonTerminalVeto = "terminal-veto"
)

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

// The mode symbols the observers receive (owner decision 1): the panel
// glyph and the menu's radio mark are the same wire-literal — one truth,
// two consumers (quick plan 260930-pf6, plan 07-05).
const (
	symbolEN = "en"
	symbolRU = "ru"
)

// emitterSlot wraps the actor's emitter sink for the atomic pointer: Go
// atomics store pointers, not interface values, so the engine.Emitter rides
// in a one-field wrapper (G-6-1).
type emitterSlot struct {
	eng engine.Emitter
}

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
	mu         sync.Mutex
	rungMu     sync.Mutex // serializes the clipboard rung (WR-01: the rung runs off the actor mutex)
	fsm        *hotkey.FSM
	window     time.Duration
	verifyWait time.Duration // the ADR-004 budget in force (CR-02: fed by timeouts.verify_wait_ms)
	start      time.Time
	timer      *time.Timer
	buf        *correct.Buffer
	caps       uint32
	// engSlot is the emitter sink of the engine minted for the input
	// context — an ATOMIC slot, deliberately outside the actor mutex: the
	// engine factory calls AttachEngine re-entrantly DURING a flip (ibus
	// answers SetGlobalEngine by calling CreateEngine), and that call must
	// never wait for a.mu (G-6-1). Reads go through the emitter() snapshot
	// accessor; nil before the first mint, exactly like the plain field it
	// replaced.
	engSlot atomic.Pointer[emitterSlot]
	surr    []rune // cached text-before-cursor from the latest client push
	// surrSeen is the DST guard's generation-scoped materialization evidence
	// (plan 10-16, the 2026-10-16 bisection): set together with the
	// surrounding cache — EVERY push or materializing Require answer of the
	// current generation proves the composer-side surrounding state exists —
	// and cleared exactly where the cache dies (the FocusOut/Reset branch,
	// a.surr=nil + ctxEpoch++). surrEng binds the evidence to the emitter it
	// was seen on (the flapEng interface-identity form: engine objects are
	// minted per input context), so a churn-missed re-mint cannot inherit a
	// dead generation's evidence. On mutter 46.2 a DeleteSurroundingText
	// toward a never-materialized context segfaults gnome-shell (strlen over
	// the virgin state, meta-wayland-text-input.c:254 — ANY parameters, the
	// ~1.5 s virgin window; three cores, one frame), so the correction
	// machine emits DST only over this evidence (executeCorrection).
	surrSeen    bool
	surrEng     engine.Emitter
	sel         selectionState
	pending     *pendingFix
	after       *pendingAfter
	verifyEpoch uint64 // monotonic verify-after round tag (stale-timer guard)
	// pendingEpoch is the pre-correction round tag counter (plan 10-10,
	// WR-02): armed with every pending fix and captured by the deadline's
	// callback — the VerifyExpiry stale-timer guard. The verifyEpoch
	// discipline: a monotonic tag, read and written under the mutex.
	pendingEpoch uint64
	// ctxEpoch is the input-context generation tag (plan 10-04): the
	// context-death branch of HandleLifecycle bumps it on every
	// FocusOut/Reset, and the terminal write branch compares it between
	// chunks — a focus loss landing in the chunk-boundary window stops the
	// round without completion. The verifyEpoch discipline: a monotonic
	// tag, read and written under the mutex.
	ctxEpoch     uint64               // monotonic input-context generation (chunk re-check)
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
	// corrDstGuarded counts the DST guard's diversions (plan 10-16, the
	// 2026-10-16 bisection): level-1 rounds whose DeleteSurroundingText was
	// WITHHELD for the missing generation-scoped materialization evidence
	// and served by the forward branch instead (the T-10-16-04 visibility —
	// the owner must see WHY a replacement went forwards). Counts only,
	// never content (D-20/D-21).
	corrDstGuarded int
	// Autocorrect counters (plan 06-06, D-54): the fired corrections, the
	// fail-closed abstentions with their reason slugs — counts and closed
	// slugs only, never the typed or corrected word (D-20/D-21, Pitfall 6).
	// acGeneration tags armed payloads so a newer boundary supersedes a
	// mid-flight confirm; acAppWarned/acRoleWarned keep the one-WARN
	// discipline per degradation episode (the appidWarned precedent).
	acFired      int
	acAbstained  int
	acReasons    map[string]int
	acGeneration uint64
	acAppWarned  bool
	acRoleWarned bool
	appid        AppidSource
	appidStarted bool
	// acBlocklist is the compiled view of the document's blocklist — the
	// gate matches the identity against these, never against the raw
	// strings (one compile per list change, not per boundary). Names is
	// the raw list the cache was built from, the macrLettersName
	// parse-cache form: rebuilt only when the document's list changed.
	acBlocklistNames []string
	acBlocklist      []*regexp.Regexp
	// acTerminals is the compiled view of the document's terminal-class
	// identity patterns (correction.terminals, Phase 10/D-55) — the same
	// one-compile-per-list-change cache discipline as acBlocklist. Names
	// is the raw list the cache was built from (the parse-cache form).
	acTerminalNames []string
	acTerminals     []*regexp.Regexp
	// blApp/blFresh are the freshness snapshot of the focused-app identity
	// for the blocklist menu item (phase 11, MENU-01 — D-03 «демон не
	// угадывает»): the appid observer's CACHE keeps the «last known»
	// identity through focus losses, so the menu must NOT read it directly —
	// a stale name would render as the active item of the WRONG window and a
	// click would record garbage. blApp is the last GAIN identity delivered
	// by the WithGainHook (PushFocusedApp); blFresh dies on every REAL
	// FocusOut (no synthetic flip credit, G-2-3) — freshness, not the cache,
	// decides. Accessed only under a.mu.
	blApp   string
	blFresh bool
	// verifyDead is the behavioral terminal class of the LIVE input
	// context (RESEARCH Pattern 1, source 3): set when a verify round of
	// this context expired by budget while the context was still alive
	// (the wezterm signature — the surrounding-text bit is announced, the
	// pushes never come), cleared by the context's death (HandleLifecycle
	// FocusOut/Reset — the same lifecycle that retires the pending round;
	// the own flip's synthetic FocusOut is not a death — the class rides
	// the G-2-3 carve-out). A5 insurance: every transition logs one
	// source-neutral INFO marker, word-free (D-20/D-21).
	verifyDead bool
	// detectCheck is the boundary's detector seam (plan 10-03): the
	// indirection the veto corpus spies through — UseDetector installs a
	// counting double, nil keeps the real detect.Check (the production
	// wiring never touches the seam).
	detectCheck detectorFunc
	// The deferred verify-dead wipe of the flap window (plan 10-07,
	// G-10-1 cause A): a REAL FocusOut arms focusFlapWindow once — the
	// wipe executes at expiry unless a FocusIn cancelled it first (the
	// same-engine flap pair). A Reset wipes immediately, byte as before;
	// a synthetic FocusOut (the own flip's credit) wipes nothing — the
	// class rides (orca-terminal-second-shift). flapEpoch is the
	// stale-deadline guard (the afterExpiry form: a monotonic tag compared
	// under the mutex — a re-armed window or a completed pair turns a
	// queued callback stale); flapStop is the armed timer's stop; flapAfter
	// is the timer seam (the level2Sleep precedent: the production arm is
	// time.AfterFunc, the corpus installs recorders).
	flapEpoch uint64
	flapStop  func()
	flapAfter func(d time.Duration, fn func()) (stop func())
	// The deferred correction flip of the terminal write branch (plan 10-08,
	// G-10-1 cause B, the E3 verdict): the post-correction mode flip is
	// ARMED, not executed — SetGlobalEngine 2–8 ms after the commit orphans
	// the in-flight commit under host churn (the client processes the commit
	// late, the re-minted context discards the stale engine's signal, the
	// committed word is lost). The armed flip fires on the FIRST delivery
	// confirmation — a confirming push, the next keystroke, or the grace
	// deadline — and NO release site sits on the FocusOut lifecycle (a flap
	// pair must not be able to reproduce E3; the mode rule is global, not
	// context-scoped). flipEpoch is the stale-deadline guard (the afterExpiry
	// discipline), flipStop cancels the armed deadline at release, and
	// flipAfter is the timer seam (the flapAfter form: the production arm is
	// time.AfterFunc, the corpus installs recorders).
	flipPending *deferredFlip
	flipEpoch   uint64
	flipStop    func()
	flipAfter   func(d time.Duration, fn func()) (stop func())
	// flipAftermathStop/flipAftermathEpoch hold the re-mint-aftermath window
	// of the daemon's own flip (Round 2, orca-terminal-second-shift): armed
	// by every successful flipTo round trip next to the synthetic-lifecycle
	// credit, spent by the client's late Reset inside the window (the
	// re-mint echo — the learned class keeps), expired by the deadline (a
	// Reset from then on is REAL — the immediate wipe stands), cancelled by
	// a real FocusOut. Single slot, the flap-window discipline (T-10-07-04):
	// a newer flip re-arms the same slot, never a second timer. Accessed
	// only under a.mu — flipTo's callers and HandleLifecycle serialize on
	// it (the flipCredit form).
	flipAftermathStop  func()
	flipAftermathEpoch uint64
	// flipAftermathAfter is the timer seam of the re-mint-aftermath window
	// (the flipAfter form: the production arm is time.AfterFunc
	// (defaultFlapAfter); the corpus installs a deterministic recorder via
	// UseFlipAftermathTimer).
	flipAftermathAfter func(d time.Duration, fn func()) (stop func())
	// verifyAfter is the pre-correction verify-deadline seam (plan 10-10,
	// WR-02, the flipAfter form): arm(d, fn) arms one verify budget and
	// returns the stop function canceling a still-pending firing. The
	// production arm is time.AfterFunc (defaultFlapAfter); the corpus
	// installs a deterministic recorder. The daemon path never calls it.
	verifyAfter func(d time.Duration, fn func()) (stop func())
	// l1 is the open confirmation window of the two-phase level-1 delivery
	// (plan 10-11, round-2 G-10-1, the 09:39 host rows): the level-1 ladder
	// emitted its DeleteSurroundingText and the commit phase waits for the
	// client's proof the deletion applied. Nil whenever no ladder is
	// mid-flight. l1Epoch is the fallback budget's stale-deadline guard
	// (the afterExpiry discipline) and l1After the timer seam (the
	// flipAfter form).
	l1      *level1Confirm
	l1Epoch uint64
	l1After func(d time.Duration, fn func()) (stop func())
	// flapEng is the emitter snapshot of the flap window (plan 10-11,
	// WR-03 of the 10-REVIEW): captured by the REAL FocusOut that arms the
	// deferred verify-dead wipe, compared by the completing FocusIn — the
	// wipe cancels ONLY for the same pair of engine objects (the journal's
	// 1–3 ms same-engine flap pairs); a different object is a real window
	// switch completing inside the 10 ms window and the class must die
	// with the old context (the invariant «класс живого контекста»).
	// Cleared together with the window slot (cancel/expiry); accessed only
	// under a.mu (HandleLifecycle serializes on it). The comparison is
	// interface identity (==): both live implementations are pointers
	// (*engine.Engine, the corpus *fakeSink).
	flapEng engine.Emitter
	// level2ForwardGap and terminalDrainSettle are the write branch's
	// pacing values (plan 10-05 and plan 10-17): NewActor plants the
	// journal-proven defaults; ONLY the daemon's -debug wiring overrides
	// them (the stand pacing matrix, plan 10-17 Task 3) — never config,
	// never a second writer. Read under a.mu by the machine.
	level2ForwardGap    time.Duration
	terminalDrainSettle time.Duration
	// level2Busy is the write branch's re-entrancy guard (plan 10-10,
	// WR-01): executeLevel2 checks-and-sets it BEFORE the first destructive
	// emission and clears it on EVERY exit (the backspace-cap refusal, the
	// context-changed stop, the success return). A second correction entry
	// through the chunk-boundary mutex window (a CorrectNow over ctl, a
	// fired autocorrect confirm, a tap-series expiry) refuses with the
	// correction-busy slug — never a second BuildPlan replayed over the
	// same un-replaced buffer, which in a terminal erases text beyond
	// intent. The flag never crosses rounds: cleared at each exit, the
	// next executeLevel2 passes.
	level2Busy bool
	// machineDone is the close-on-exit channel of the in-flight delivery
	// machine (CR-02 of the 10-REVIEW, the round-4 fix session): armed
	// together with level2Busy (machineArm), closed on EVERY machine exit
	// (machineSettle — the backspace-cap refusal, the free abort, the
	// success return). The key-hold path waits on it bounded, so a keystroke
	// that arrived inside the unlocked drain-settle window can never ride
	// the wire ahead of the machine's pending commit. Accessed only under
	// a.mu; busy==true implies machineDone!=nil.
	machineDone chan struct{}
	// emissionSeq is the monotonic identity of the terminal delivery bursts
	// (the round-5 instrument, the 2026-10-17 wezterm double-delete hunt):
	// executeLevel2's terminal branch assigns the next value at the burst's
	// burst-window open and every «correction delivery» row of that burst
	// carries it. The journal juxtaposed against a pty log stays decisive:
	// one seq in the daemon journal while the pty shows two delete runs ⇒
	// the client replayed its queued forwards; two seqs on the wire ⇒ the
	// daemon fired twice. A counter, never a key — content-free by
	// construction (D-20/D-21). Accessed only under a.mu.
	emissionSeq uint64
	// deliveryStart is the wall-clock anchor of the in-flight burst's
	// timeline: every delivery row's t_rel_ms is relative to it — the
	// per-emission timing the aggregate rows never carried. It stands after
	// the burst closes: the deferred flip releases later and stamps its own
	// captured anchor (deferredFlip.seq/start), never a newer burst's.
	// Accessed only under a.mu.
	deliveryStart time.Time
	// holdQueue is the FIFO of key events parked while the delivery machine
	// is in flight (CR-02): typed-layout input (printables and BackSpace)
	// must not hit the wire before the pending commit — the parked events
	// process post-settle in arrival order and their transits re-emit from
	// the daemon (HandleKey consumed them at arrival, so ibus's own
	// forwarding never happened). Gestures, bare modifiers and the reset
	// keys never park — their semantics belong to the arrival instant. The
	// queue drains on the goroutine of the FIRST parked key (holdDrain is
	// the single-flight flag); accessed only under a.mu.
	holdQueue []engine.EngineEvent
	holdDrain bool
	// role is the live AT-SPI role seam of the autocorrect policy (plan
	// 06-06): the SAME observer as appid when the concrete source implements
	// RoleSource, or a test double installed via UseRole.
	level2Sleep func(time.Duration) // the write branch's pacing seam (UseLevel2Sleeper)
	role        RoleSource
	// acRoleErrors counts EVERY role-request error of the live episode
	// (plan 10-07, G-10-1 missing #3): the 2026-10-13 host journal shows
	// the 11:00:16 WARN consumed the episode and the broken source stayed
	// silent the whole day — the counter makes that machine-visible in
	// goswitchctl status. A healthy answer closes the episode and zeroes
	// it (the :2741 semantics extended). acRoleWarnedAt anchors the sticky
	// WARN's re-arm window; roleClock is the episode's time seam (the
	// flapAfter form: the production clock is time.Now, the corpus
	// installs a controllable double).
	acRoleErrors   int
	acRoleWarnedAt time.Time
	roleClock      func() time.Time
	appidWarned    bool // one WARN per degradation episode — a broken source must not spam per keystroke
	startAppid     func() (AppidSource, error)
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
	// flipCredit is the synthetic-lifecycle credit of the daemon's own flip
	// (ADR-004 amendment 2026-10-07, G-2-3): a successful SetGlobalEngine
	// round trip inside flipTo makes ibus re-mint engine objects, and the
	// engine receives a synthetic focus_out/focus_in pair. flipTo arms one
	// credit per successful round trip; HandleLifecycle's FocusOut branch
	// consumes exactly one and skips ONLY the buffer hard-reset — a
	// synthetic transition of the own flip is not «смена фокуса окна». A
	// failed round trip (the WARN form) or the nil-seam degradation arms
	// nothing, so a real focus loss keeps its full reset. Accessed only
	// under a.mu — flipTo's callers and HandleLifecycle serialize on it.
	flipCredit int
	// display is the tray-indicator seam (quick 260930-pf6): the ModeDisplay
	// the daemon wiring installs via SetModeDisplay; nil = no display, every
	// mode change stays invisible to it. The display degrades itself — the
	// actor adds no error handling around the call.
	display ModeDisplay
	// menuSync is the tray-menu state seam (plan 07-05): the MenuSync the
	// daemon wiring installs via SetMenuSync; nil = no menu, every push
	// stays a no-op. The menu degrades itself (the display precedent).
	menuSync MenuSync
	// soundSink is the acoustic-feedback seam (plan 07-08): the SoundSink
	// the daemon wiring installs via SetSoundSink; nil = no sink, every
	// tone stays a silent no-op. The sink degrades itself (the menu
	// precedent).
	soundSink SoundSink
	// soundACEvent is the autocorrect tone's event last seen in a folded
	// document (WR-02): the dedupe key of the sink pushes and the value a
	// late SetSoundSink self-syncs to (the SetMenuSync install-push
	// precedent).
	soundACEvent string
	// a11ySink is the accessibility-magic seam (plan 08-05): the A11ySink
	// the daemon wiring installs via SetA11ySink; nil = no sink, every
	// push stays a silent no-op. The sink degrades itself (the menu
	// precedent).
	a11ySink A11ySink
	// a11yActive is the a11y state last seen in a folded document (plan
	// 08-05): the diff gate of the sink pushes and the value a late
	// SetA11ySink self-syncs to (the soundACEvent precedent).
	a11yActive bool
	// soundEnabled is the applied sound switch folded from the snapshot
	// (plan 07-05) — the EffectiveEnabled truth the status token serves.
	soundEnabled bool
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
	// AutoCorrectBlocklist carries the D-53 revision's regex patterns: a
	// SUBSTRING match (RE2 MatchString, ^…$ anchoring explicit) against the
	// focused app's bridge-namespace identity FORBIDS the correction; an
	// empty list forbids nothing. The thresholds mirror detect.Params and
	// the config 06-04 defaults (4/2.0/1.0 — change the places together).
	AutoCorrectEnabled    bool
	AutoCorrectBlocklist  []string
	AutoCorrectMinWordLen int
	AutoCorrectMargin     float64
	AutoCorrectFloor      float64
	// Terminals carries the document's terminal-class identity patterns
	// (correction.terminals, Phase 10/D-55): a substring RE2 match against
	// the focused app's bridge-namespace identity MARKS the context
	// terminal — the classification seam's second source. The zero value
	// (an empty list) marks nothing — the old behavior, the D-54
	// precedent.
	Terminals []string
	// SoundEnabled is the acoustic-feedback switch of the owner's «Звуки
	// при переключении» decision (plan 07-08): DEFAULT ON — only an
	// explicit enabled: false silences the tones. Fed from the document's
	// Sound.EffectiveEnabled through applySnapshot (hot reload) or the
	// startup wiring's SetOptions; the gate sits ahead of the sink, so a
	// muted daemon never consults it.
	SoundEnabled bool
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

// correctionRange parameterizes the correction pipeline by its range (one
// pipeline, different ranges): the word path (Double) supplies the token,
// its boundary tail and ReplaceToken; the phrase path (Triple) supplies the
// whole phrase since the last hard reset, no tail and ReplacePhrase (D-25).
// The selection path (03-03) supplies the client-reported selection — the
// range lives in the surrounding text, not the buffer, so its replace half
// is the buffer's honest death (HardReset) instead of an edit.
type correctionRange struct {
	token   []rune
	tail    []rune
	replace func(converted []rune)
	// acBoundary marks a range armed by the autocorrect word boundary
	// (plan 06-07 live finding): the separator is being typed in the same
	// event that fires the layer, so the client answers the pre-correction
	// verify's RequireSurroundingText BEFORE the tail reaches it — the
	// first answer is the pre-boundary state, an early report and not a
	// mismatch. The manual paths leave it false: their verify semantics
	// stay byte-as-today.
	acBoundary bool
}

// acPayload is one armed boundary decision (plan 06-06): the correction
// range snapshot (token+tail+replace) taken under the mutex at the
// boundary, the detector's direction, the typed word the verdict was about
// and the arming generation. app is the bridge-namespace identity the
// blocklist verdict was evaluated for — EMPTY when no identity was visible
// at arm (the 07-04 inversion: unknown passes without an app; the confirm
// consults the LIVE identity against the blocklist — the WR-01 equality
// removal and the
// confirm rework are plan 07-04). expectToken/expectTail snapshot the buffer as
// the arming event LEAVES it (the reset branch of feedKey hard-resets
// between the capture and the launch — CR-01): the confirm's under-the-
// mutex revalidation compares the live buffer against them, so the
// boundary's own mutation is accepted and every later one (an interleaved
// keystroke, a Backspace, a focus loss) refuses the payload fail-closed.
// It travels to the confirm goroutine; the word stays in memory only — no
// log and no status surface ever sees it (D-20/D-21).
type acPayload struct {
	rng         correctionRange
	dir         correct.Dir
	word        []rune
	gen         uint64
	app         string
	expectToken []rune
	expectTail  []rune
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
	stop      func()
	sel       *selectionSpec
	// epoch is the round's tag of the WR-02 fix (plan 10-10): the monotonic
	// counter's value at arm — the armed deadline's callback carries it, and
	// VerifyExpiry no-ops unless the live pending round owns the same tag
	// (the afterExpiry form). A stale deadline from a resolved round can
	// never consume the next round nor teach the behavioral class for a
	// context that never missed a budget.
	epoch uint64
	// cacheNil snapshots the push cache at arm time (plan 10-08, the
	// owner-approved client-drift amendment of 2026-10-13): the word round
	// armed over an EMPTY cache — the client has never reported a field
	// state in this context episode (an honest GTK client pushes on every
	// change). Paired with an answer that does not hold the typed stretch,
	// that is the stale-cache terminal signature (pendingStaleCache).
	cacheNil bool
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
	// noRung marks a terminal-branch round (plan 10-04, ADR-008 D4, owner
	// verdict OQ 2 «contract»): the mismatch verdict must never arm the
	// clipboard rung — a Ctrl+V chord into a shell is the SIGINT class.
	// Forbidden BY CONSTRUCTION, not by config: the round carries no paste
	// payload at all.
	noRung bool
}

// deferredFlip is the armed post-correction mode flip (plan 10-08, G-10-1
// cause B, the E3 verdict): the SET-semantics target — the corrected
// text's script, the owner rule «режим следует за скриптом результата» —
// plus the armed slot's epoch. The epoch makes a stale grace deadline from
// a released or re-armed slot a no-op (the afterExpiry discipline).
// terminal marks the branch that armed the slot (plan 10-17): the
// terminal flip ignores the verify outcomes and arms no grace deadline —
// only the next keystroke or a confirming push releases it — while the
// level-1 flip keeps the full 10-08 machine byte-as-today. armedAt feeds
// the terminal release's delivery-mark latency. gesture marks a PARKED
// gesture-driven flip (CR-01 of the 10-REVIEW, the round-4 fix session):
// a flip that arrived while the delivery machine was in flight and must
// not SetGlobalEngine mid-window — the owner's explicit command, armed
// regardless of flip_after_correction and superseding the correction's
// own post-correction arming (never a silent drop).
type deferredFlip struct {
	target   scriptMode
	epoch    uint64
	terminal bool
	gesture  bool
	armedAt  time.Time
	// The arming burst's identity (the round-5 instrument): a flip released
	// seconds later stamps ITS OWN burst's emission-seq and timeline
	// anchor, never whatever burst (if any) happens to be in flight at the
	// release instant — the juxtaposition contract of emission-seq.
	seq   uint64
	start time.Time
}

// level1Confirm is the open confirmation window of the two-phase level-1
// delivery (plan 10-11, round-2 G-10-1, the 09:39 host rows): the ladder's
// DeleteSurroundingText is on the wire while the commit phase waits for the
// client's proof the deletion applied — a push or Require answer whose
// text-before-cursor no longer ends with match (token+tail). Everything the
// completion needs is captured HERE, at arm: the emitter snapshot rides the
// whole window (the :3444 one-snapshot precedent), so a flush after the
// context's death or re-mint can never pour the text into a NEW
// application's engine — the stale object is discarded by the client, the
// same mechanic that was the E3 vector now working as the fuse. rng carries
// the replace hook and the tail; level/token feed the frozen done row;
// armed keeps the latency origin. epoch is the stale-deadline guard of the
// fallback budget (the afterExpiry discipline); stop cancels the armed
// budget at flush.
type level1Confirm struct {
	eng       engine.Emitter
	commit    []rune
	rng       correctionRange
	converted []rune
	match     []rune
	armed     time.Time
	level     correct.Level
	epoch     uint64
	stop      func()
}

// NewActor returns an actor deciding tap series inside the given
// disambiguation window (D-05). The daemon passes hotkey.DefaultWindow.
// The series key starts at the documented default (shift_r); an attached
// config source re-resolves it from the first event on (CR-01).
func NewActor(window time.Duration) *Actor {
	a := &Actor{
		fsm:        hotkey.NewFSM(window, hotkey.KeyvalShiftR),
		window:     window,
		verifyWait: defaultVerifyWait,
		tapKeyval:  hotkey.KeyvalShiftR,
		start:      time.Now(),
		buf:        correct.NewBuffer(),
		clip:       clipboard.New(),
		// The write branch's journal-proven pacing defaults (plan 10-17
		// Task 3): only the daemon's -debug wiring overrides them.
		level2ForwardGap:    defaultLevel2ForwardGap,
		terminalDrainSettle: defaultTerminalDrainSettle,
		// The flap-window timer's production arm (plan 10-07): a real
		// time.AfterFunc deadline, stoppable — the corpus replaces it via
		// UseFlapTimer with a deterministic recorder.
		flapAfter: defaultFlapAfter,
		// The level-1 confirm budget's production arm (plan 10-11): the
		// defaultFlapAfter form — the corpus replaces it via UseLevel1Timer.
		l1After: defaultFlapAfter,
		// The role-episode clock's production form (plan 10-07): the real
		// wall clock; the corpus installs a controllable double via
		// UseRoleClock.
		roleClock: time.Now,
	}
	// The lazy per-app observer's production starter: the a11y dial of
	// internal/appid on the daemon's lifetime (the clipboard rung's
	// context.Background precedent — the actor owns no shutdown path).
	// The WithGainHook wiring (phase 11, MENU-01) delivers every gain
	// identity to PushFocusedApp — the freshness machine of the blocklist
	// menu item lives inside the actor; main.go stays untouched. The
	// assignment sits OUTSIDE the composite literal: a closure referencing
	// `a` inside its own initializer is out of scope there.
	a.startAppid = func() (AppidSource, error) {
		return appid.Start(context.Background(), appid.WithGainHook(a.PushFocusedApp))
	}

	return a
}

// defaultFlapAfter arms one real focusFlapWindow deadline (the production
// flapAfter): the returned closure stops a still-pending firing; an already
// fired timer returns false harmlessly — the epoch guard owns the verdict.
func defaultFlapAfter(d time.Duration, fn func()) func() {
	t := time.AfterFunc(d, fn)

	return func() { t.Stop() }
}

// SetOptions stores the correction options (the startup wiring feeds them
// from config.Load; the snapshot consumption on hot reload is plan 03-04).
// The caller is the daemon or a test — the actor reads the options under
// its mutex at use time.
func (a *Actor) SetOptions(o Options) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.opts = o
	a.refreshACBlocklist(o.AutoCorrectBlocklist) // the no-config surface owns the same cache
	a.refreshACTerminals(o.Terminals)            // the terminal-class cache rides the same discipline
	// An enabled autocorrect layer or a non-empty macr list arrives through
	// this surface too.
	a.ensureAppid()
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

// detectorFunc is the boundary detector's signature — the named type of
// the plan 10-03 spy seam (the RoleCall precedent of internal/appid): one
// Check over the finished token with the baked dictionaries and thresholds.
type detectorFunc func(
	tok []rune, mode string, data detect.Data, tri detect.Trigrams, params detect.Params,
) detect.Verdict

// UseRole installs a prepared role source — the wiring/test seam of the
// D-53 role contour (the UseAppid mirror). The daemon path never calls it:
// the observer arrives through ensureAppid's lazy start and the RoleSource
// type assertion.
func (a *Actor) UseRole(rs RoleSource) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.role = rs
}

// UseLevel2Sleeper installs the write branch's pacing seam — the
// wiring/test seam of the UseRole form. The production sleeper is
// time.Sleep; the corpus installs recorders (the 2026-10-12 pacing fix).
func (a *Actor) UseLevel2Sleeper(s func(time.Duration)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.level2Sleep = s
}

// SetLevel2ForwardGap overrides the write branch's pacing quantum — the
// -debug stand matrix is the only caller (plan 10-17 Task 3, T-10-17-04):
// the daemon wiring feeds it from the debug-gated env override, never
// from config, and a production daemon never reaches it. The wiring runs
// before the engine serves.
func (a *Actor) SetLevel2ForwardGap(d time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.level2ForwardGap = d
}

// SetTerminalDrainSettle overrides the terminal branch's bounded delivery
// window — the SetLevel2ForwardGap form (plan 10-17 Task 3, T-10-17-04).
func (a *Actor) SetTerminalDrainSettle(d time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.terminalDrainSettle = d
}

// UseFlapTimer installs the flap-window timer seam (plan 10-07, the
// UseLevel2Sleeper form): arm(d, fn) arms one focusFlapWindow deadline and
// returns the stop function canceling a still-pending firing. The production
// arm is time.AfterFunc (defaultFlapAfter); the corpus installs a
// deterministic recorder. The daemon path never calls it.
func (a *Actor) UseFlapTimer(arm func(d time.Duration, fn func()) (stop func())) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.flapAfter = arm
}

// UseFlipTimer installs the deferred-flip grace seam of the terminal write
// branch (plan 10-08, the UseFlapTimer form): arm(d, fn) arms one grace
// deadline and returns the stop function canceling a still-pending firing.
// The production arm is time.AfterFunc (defaultFlapAfter); the corpus
// installs a deterministic recorder. The daemon path never calls it.
func (a *Actor) UseFlipTimer(arm func(d time.Duration, fn func()) (stop func())) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.flipAfter = arm
}

// UseLevel1Timer installs the level-1 confirm budget seam (plan 10-11, the
// UseFlipTimer form): arm(d, fn) arms one confirmation budget and returns
// the stop function canceling a still-pending firing. The production arm is
// time.AfterFunc (defaultFlapAfter); the corpus installs a deterministic
// recorder.
func (a *Actor) UseLevel1Timer(arm func(d time.Duration, fn func()) (stop func())) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.l1After = arm
}

// UseFlipAftermathTimer installs the re-mint-aftermath seam of the client
// Reset carve-out (Round 2, orca-terminal-second-shift, the UseFlipTimer
// form): arm(d, fn) arms one flipAftermathWindow deadline and returns the
// stop function canceling a still-pending firing. The production arm is
// time.AfterFunc (defaultFlapAfter); the corpus installs a deterministic
// recorder. The daemon path never calls it.
func (a *Actor) UseFlipAftermathTimer(arm func(d time.Duration, fn func()) (stop func())) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.flipAftermathAfter = arm
}

// UseVerifyTimer installs the pre-correction verify-deadline seam (plan
// 10-10, WR-02, the UseFlipTimer form): arm(d, fn) arms one verify budget
// and returns the stop function canceling a still-pending firing. The
// production arm is time.AfterFunc (defaultFlapAfter); the corpus installs
// a deterministic recorder — the epoch-tagged callback stays captured
// behind the seam, so the deterministic injection fires the SAME closure
// the real timer would. The daemon path never calls it.
func (a *Actor) UseVerifyTimer(arm func(d time.Duration, fn func()) (stop func())) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.verifyAfter = arm
}

// UseRoleClock installs the role-episode clock seam (plan 10-07, the
// UseFlapTimer form): the clock timestamps the sticky role WARN's re-arm
// window. The production clock is time.Now; the corpus installs a
// controllable double. The daemon path never calls it.
func (a *Actor) UseRoleClock(now func() time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.roleClock = now
}

// UseDetector installs a boundary detector double — the plan 10-03 spy
// seam (the UseRole precedent): the veto corpus counts the calls through
// it to prove the detector never runs on a terminal class and runs exactly
// as before on a GTK class. nil restores the real detect.Check. The daemon
// path never calls it.
func (a *Actor) UseDetector(fn detectorFunc) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.detectCheck = fn
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
// counters with their skip-reason breakdown, the MACR interception
// counters, the autocorrect layer's state and counters (plan 06-06, D-54)
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
	// The autocorrect layer's state and counters (plan 06-06, D-54): the
	// in-force switch and the D-53 outcome counts with their closed-slug
	// breakdown — counts and states only, never the word (D-20/D-21).
	AutoCorrectEnabled     bool
	AutoCorrectFired       int
	AutoCorrectAbstained   int
	AutoCorrectSkipReasons map[string]int
	// AutoCorrectRoleErrors is the live role-request error count (plan
	// 10-07, G-10-1 missing #3): every error of the current degradation
	// episode is counted — the 11:00:16 episode stayed silent for the rest
	// of the day behind its single WARN. A healthy answer zeroes it.
	AutoCorrectRoleErrors int
	// CorrDstGuarded is the DST guard's diversion counter (plan 10-16, the
	// 2026-10-16 bisection): level-1 corrections rerouted to the forward
	// branch for the missing generation-scoped materialization evidence —
	// the machine-visible WHY of a forwards-delivered replacement
	// (T-10-16-04). Counts only (D-20/D-21).
	CorrDstGuarded int
	// SoundEnabled is the applied sound switch (plan 07-05) — the
	// EffectiveEnabled truth of the applied snapshot (an absent section
	// reads ON, the owner's default), the `sound_enabled` status token the
	// menu composition and the e2e oracle read.
	SoundEnabled bool
	ConfigPath   string
	ConfigValid  bool
	ConfigError  string
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
		Mode:                  symbolEN,
		Engine:                engineNameOf(a.mode),
		CorrectionsDone:       a.corrDone,
		CorrectionsSkipped:    a.corrSkipped,
		SkipReasons:           maps.Clone(a.skipReasons),
		SuperIntercepted:      a.macrIntercepted,
		SuperUpstreamConsumed: a.macrConsumed,
		AutoCorrectEnabled:    a.opts.AutoCorrectEnabled,
		AutoCorrectFired:      a.acFired,
		AutoCorrectAbstained:  a.acAbstained,
		SoundEnabled:          a.soundEnabled,

		AutoCorrectSkipReasons: maps.Clone(a.acReasons),
		AutoCorrectRoleErrors:  a.acRoleErrors,
		CorrDstGuarded:         a.corrDstGuarded,
		ConfigValid:            true, // built-in defaults, or a source without a status surface
	}
	if a.mode == modeRU {
		st.Mode = symbolRU
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
//
// CR-02 (the 10-REVIEW, the round-4 fix session): typed-layout input
// arriving while the delivery machine is in flight must not hit the wire
// before the machine's pending commit — the round-3 machine held the mutex
// through burst+commit so the serialization guaranteed the order, and the
// drain window released it, making the inversion («xпривет » instead of
// «привет x») routine for typing continued right after the gesture. The
// hold path parks the event (the arrival verdict is CONSUMED — ibus's own
// forwarding never happens) and processes it post-settle, re-emitting its
// transit from the daemon: wire order = deletes → commit → held keys, FIFO.
func (a *Actor) HandleKey(ev engine.EngineEvent) (consume bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if (a.level2Busy && holdableKey(ev)) || len(a.holdQueue) > 0 {
		return a.holdDuringDelivery(ev)
	}

	return a.handleKeyArrived(ev)
}

// HandleLifecycle implements engine.EventHandler: FocusOut and Reset disarm
// a pending series (the input context is gone — a decision there would be
// garbage), hard-reset the phrase buffer (CORR-09) and retire any open
// correction round — its verify answer belongs to an input context that no
// longer exists; everything else is a DEBUG trace. ADR-004 amendment
// 2026-10-07 (G-2-3): a FocusOut carrying the own flip's synthetic-lifecycle
// credit (a successful SetGlobalEngine makes ibus re-mint engine objects)
// skips the buffer hard-reset AND the behavioral-class wipe — the synthetic
// transition of the daemon's own flip is not a real focus change; a real
// focus loss still resets. Round 2 amendment 2026-10-31
// (orca-terminal-second-shift): a Reset arriving within the
// flipAftermathWindow of the daemon's own flip is the slow client's echo of
// the same re-mint (the IM context reconnecting asynchronously on the new
// engine object) and is equated to the synthetic pair — the buffer resets,
// the behavioral class survives; a Reset outside the window wipes
// immediately, byte as before.
func (a *Actor) HandleLifecycle(kind engine.LifecycleKind) {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch kind {
	case engine.LifecycleFocusOut, engine.LifecycleReset:
		// ADR-004 amendment 2026-10-07 (G-2-3): a FocusOut carrying the own
		// flip's synthetic-lifecycle credit is not «смена фокуса окна» —
		// exactly one credit is consumed and the BUFFER hard-reset plus the
		// BEHAVIORAL-CLASS wipe are skipped, so a word typed across the
		// daemon's own flip corrects as a whole and the client's learned
		// verify behavior survives its own engine's re-mint (debug session
		// orca-terminal-second-shift). Every other side effect below is
		// context-scoped (ibus re-minted the engine object) and stays
		// correct: the FSM reset, the surrounding-cache clear, the
		// pending-resolution retirement, the clearAfter, the combo/MACR
		// cleanup, the timer stop. A Reset kind never consumes a credit —
		// the observed synthetic pair is focus_out/focus_in; a spent or
		// absent credit means a REAL focus loss and takes the full reset.
		// Round 2 (orca-terminal-second-shift): the client's echo Reset of
		// the re-mint is equated to the synthetic pair for the BEHAVIORAL
		// CLASS only — it spends the flipAftermathWindow slot instead of a
		// credit (the pair below consumed the credit already); see the
		// wipe-discipline switch.
		synthetic := false
		if kind == engine.LifecycleFocusOut && a.flipCredit > 0 {
			a.flipCredit--
			synthetic = true
		}
		// The blocklist item's freshness (phase 11, MENU-01 — D-03): a REAL
		// FocusOut is the window switch the cache kept the identity through —
		// the item must not render the previous window's name as active, so
		// the freshness dies and the menu learns immediately. The own flip's
		// synthetic FocusOut is not a window switch (G-2-3) and a Reset is
		// not a window switch either (the observed synthetic pair is
		// focus_out/focus_in) — neither touches the freshness.
		if kind == engine.LifecycleFocusOut && !synthetic {
			a.blFresh = false
			a.pushBlocklistAppLocked()
		}
		// The two-phase level-1 ladder's flush site (plan 10-11): the input
		// context is leaving — the ladder completes over the window's
		// CAPTURED emitter snapshot, never the current emitter() (the stale
		// object is discarded by the client). Byte-parity with today's pair:
		// delete and commit are already both on the wire by the time a
		// FocusOut could observe them. The verify-after the flush arms
		// retires with the context below (clearAfter) — the same retirement
		// a post-commit lifecycle executes today.
		a.flushLevel1Confirm()
		a.fsm.Feed(hotkey.Reset{}, a.elapsed())
		if !synthetic {
			a.buf.HardReset()
		}
		// the caches belong to the input context that just left
		a.surr = nil
		a.sel = selectionState{}
		// The DST guard's materialization evidence dies WITH the cache
		// (plan 10-16): the generation that proved the composer-side
		// surrounding state exists is gone, and a fresh generation starts
		// unevidenced — the virgin-window rule.
		a.surrSeen = false
		a.surrEng = nil
		// the context generation moves with it (plan 10-04): the terminal
		// write branch reads this tag between chunks — any death of the
		// context inside the burst's boundary windows stops the round
		a.ctxEpoch++
		// The behavioral terminal class (Phase 10) rides its own wipe
		// discipline — deferred for a real FocusOut, immediate for a Reset,
		// nothing for the own flip's synthetic pair and its client echo
		// (verifyDeadOnContextDeath carries the full contract).
		a.verifyDeadOnContextDeath(kind, synthetic)
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
		if kind == engine.LifecycleFocusIn {
			// Plan 10-07 (G-10-1 cause A): the flap pair completed within
			// the window — the context never left; the learned class
			// survives and the pending wipe dies with the window. Plan
			// 10-11 (WR-03 of the 10-REVIEW): «completed» now means the
			// SAME pair of engine objects — the wipe cancels only when the
			// FocusIn's emitter is the one whose FocusOut armed the window
			// (the journal's 1–3 ms same-engine pairs). A different object
			// is a REAL window switch that happened to complete inside the
			// 10 ms window: the wipe is left to expire and the class dies
			// with the old context — the freshly focused context classifies
			// from zero. A nil snapshot (no window armed) cancels nothing.
			if eng := a.emitter(); a.flapEng != nil && a.flapEng == eng {
				a.cancelVerifyDeadWipe()
			}
		}
		// The panel indicator self-heals on every focus gain (owner
		// decision 1, quick plan 260927-way): engine objects are minted
		// per input context, so a freshly minted context re-asserts the
		// CURRENT mode symbol — it must not resurrect the factory's
		// initial EN registration while the actor sits in another mode.
		if eng := a.emitter(); kind == engine.LifecycleFocusIn && eng != nil {
			eng.UpdateModeSymbol(a.modeSymbol())
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
//
// G-6-1 contract: the engine factory calls this method from a RE-ENTRANT
// CreateEngine DURING a flip (ibus-daemon answers the flip's SetGlobalEngine
// by synchronously minting the target engine, and the flip's await holds
// a.mu) — the method must answer without ever waiting on the actor mutex.
// The swap is one atomic store; there are no read-modify-write races
// (out-of-order stores between concurrent CreateEngine calls keep the
// pre-fix equivalent ordering, and the corpus runs under -race).
func (a *Actor) AttachEngine(eng engine.Emitter) {
	a.engSlot.Store(&emitterSlot{eng: eng})
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

// SetMenuSync installs the menu seam — the SetModeDisplay mirror (plan
// 07-05). The menu immediately receives the CURRENT mode and the CURRENT
// corrections counter (the Status row's truth from the first paint): the
// startup install shows the initial state, and a late install (after flips,
// corrections or syncs) self-syncs to the factual state instead of waiting
// for the next change (the SetSoundSink install-push precedent). The
// applied config truth reaches the menu through the fold (applySnapshot)
// and the click composition's direct pushes.
func (a *Actor) SetMenuSync(ms MenuSync) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.menuSync = ms
	if ms != nil {
		ms.SetMode(a.modeSymbol())
		ms.SetCorrections(a.corrDone)
		// The blocklist item's startup install push (the RESEARCH push
		// moments, the SetCorrections install form): a late attach — a menu
		// rebuilt after an IBus restart — receives the CURRENT freshness
		// state instead of serving the unknown default until the next gain.
		a.pushBlocklistAppLocked()
	}
}

// autocorrectBlocklistCeil is the autocorrect.apps_blocklist schema ceiling
// (config.MaxAutocorrectBlocklist, plan 11-03) restated as the session
// package constant the full-state computation reads: the menu's
// заполненность (full) renders as soon as the list length reaches it. The
// single-source reconciliation to the exported constant is one point edit
// in plan 11-04 (Task 2) — not smeared across 11-02.
const autocorrectBlocklistCeil = 64

// PushFocusedApp receives one focus GAIN identity from the appid observer's
// freshness hook (phase 11, MENU-01 — wired via appid.WithGainHook in the
// default startAppid): the snapshot identity and its freshness are stored
// under the actor's mutex and the blocklist menu item's state is pushed.
// The hook runs on the observer's goroutine OUTSIDE the observer's mutex
// (Pitfall 9) — taking a.mu here keeps the one-way actor.mu → sink.mu
// ordering; the menu dedupes identical pushes, so the menuSync call is the
// documented quick, non-blocking sink (the MenuSync contract).
func (a *Actor) PushFocusedApp(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if name == "" {
		return // a gain without an identity is nothing to remember (D-03)
	}
	a.blApp = name
	a.blFresh = true
	a.pushBlocklistAppLocked()
}

// blocklistStateLocked computes the blocklist menu item's state triple from
// the actor's snapshot: app+fresh (the freshness machine — empty and false
// when unknown, the D-03 static state), covered (any compiled blocklist
// pattern matches the FRESH identity — the same re.MatchString discipline
// as autoCorrectBoundary's gate, never the raw strings) and full (the
// document's list length reached the ceiling). The caller holds the mutex.
func (a *Actor) blocklistStateLocked() (app string, fresh, covered, full bool) {
	full = len(a.opts.AutoCorrectBlocklist) >= autocorrectBlocklistCeil
	fresh = a.blFresh && a.blApp != ""
	if !fresh {
		return "", false, false, full
	}
	for _, re := range a.acBlocklist {
		if re.MatchString(a.blApp) {
			return a.blApp, true, true, full
		}
	}

	return a.blApp, true, false, full
}

// pushBlocklistAppLocked hands the blocklist item's computed state to the
// menu — the pushMenuSync discipline: synchronous under the actor's mutex,
// quick, never panics or blocks; the menu dedupes identical values. A nil
// menuSync (no menu attached) is the silent no-op. The caller holds the
// mutex.
func (a *Actor) pushBlocklistAppLocked() {
	if a.menuSync == nil {
		return
	}
	app, fresh, covered, full := a.blocklistStateLocked()
	a.menuSync.SetBlocklistApp(app, fresh, covered, full)
}

// SoundSink is the acoustic-feedback seam (plan 07-08, the owner's «Звуки
// при переключении»): Flip plays the flip tone, AutoCorrect the distinct
// autocorrect tone, SetAutocorrectEvent re-pins that tone's document event
// (WR-02: the key folds live like every other document truth — the D-32
// contract holds without a restart). Defined at the point of use; the
// interface travels with the consumer (the AppidSource precedent). The
// implementation is fire-and-forget by contract — it must never block the
// actor's hot path and must never panic; every playback failure is the
// sink's own best-effort episode (one WARN), never an actor error.
type SoundSink interface {
	Flip()
	AutoCorrect()
	SetAutocorrectEvent(event string)
}

// SoundStopper is the mute lifecycle's optional capability (261006-squ,
// requirement 3): the sound switch's ON→OFF fold transition stops the
// sink — muted sound holds NO sound-server connection; re-enabling dials
// lazily on the first tone. Optional by design: a sink without it folds
// silently (the type-assertion miss is a no-op). Defined at the point of
// use, next to SoundSink; Stop must be quick and re-entrant-safe under
// the actor's mutex (the actor.mu → sink.mu one-way ordering holds).
type SoundStopper interface {
	Stop()
}

// SetSoundSink installs the sound seam — the SetMenuSync mirror (plan
// 07-08). nil = no sink: every tone stays a silent no-op (the menu nil
// form — a degradation, never an error). A non-nil install self-syncs the
// last folded autocorrect event (the SetMenuSync install-push precedent,
// WR-02): a sink attached after a fold — the OnConn ordering — still
// serves the document truth.
func (a *Actor) SetSoundSink(s SoundSink) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.soundSink = s
	if s != nil && a.soundACEvent != "" {
		s.SetAutocorrectEvent(a.soundACEvent)
	}
}

// A11ySink is the accessibility-magic seam (plan 08-05, D-8-3): the fold
// pushes the a11y section's EffectiveEnabled() truth to the desktop
// reconciler.
// Defined at the point of use; the interface travels with the consumer
// (the SoundSink precedent). The implementation is fire-and-forget by
// contract — it must never block the actor's hot path (the apply series
// runs on the sink's own serialization, outside the actor mutex, WR-01)
// and must never panic; every failure is the sink's own best-effort
// episode (one WARN), never an actor error.
type A11ySink interface {
	Apply(active bool)
}

// SetA11ySink installs the accessibility-magic seam (plan 08-05) — the
// SetSoundSink mirror. nil = no sink: every push stays a silent no-op (the
// menu nil form — a degradation, never an error). A non-nil install
// self-syncs the last folded a11y state (the SetSoundSink install-push
// precedent): a sink attached after a fold — the OnConn ordering, where
// the startup FoldAppliedConfig has already run — still serves the folded
// truth immediately, so the D-8-3 reconcile-at-start never depends on the
// install order.
func (a *Actor) SetA11ySink(s A11ySink) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.a11ySink = s
	if s != nil && a.a11yActive {
		s.Apply(true)
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
// moot. The epoch argument is the armed round's tag (WR-02, plan 10-10): a
// callback whose deadline belonged to a RESOLVED round is a quiet no-op —
// it must never consume the next pending round nor teach the behavioral
// class for a context that never missed a budget (the afterExpiry form).
// Tests inject it directly (deterministic deadline).
func (a *Actor) VerifyExpiry(epoch uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.applySnapshot() // one config read per event (CONF-02, Pattern 2)

	if a.pending == nil || a.pending.epoch != epoch {
		return // already settled, or a stale deadline from a resolved round (WR-02) — a no-op
	}
	// The behavioral terminal class (Phase 10, RESEARCH Pattern 1 source
	// 3): the budget expired while the context is still alive — the
	// pending guard above already retired every context-death case, so
	// this round's expiry is the wezterm signature learned on the living
	// context. The class persists for the context's remaining life and is
	// cleared by HandleLifecycle (FocusOut/Reset).
	a.setVerifyDead(true)
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

// SwitchMode flips the script mode to the NAMED target (plan 07-05): the
// menu's EN/RU radio pair is ONE MORE GESTURE on the single flipTo
// execution path (ADR-006) — never a second flip mechanism, never a
// gsettings or SetGlobalEngine shortcut (the SWCH regression guard: the
// D-36 record order and the flip guard discipline are flipTo's, untouched).
// The same-target case lands on flipTo's no-op guard — never a duplicated
// check; an unknown symbol is a silent no-op.
func (a *Actor) SwitchMode(target string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var m scriptMode
	switch target {
	case symbolRU:
		m = modeRU
	case symbolEN:
		m = modeEN
	default:
		return
	}
	a.flipTo(m)
}

// FoldAppliedConfig re-reads the attached config source NOW — the menu
// toggle's synchronous-apply seam (plan 07-06, the 07-05 pin): the
// toggle's reload re-stores the watcher snapshot and this fold puts the
// applied values in force without waiting for the next key event. The
// 03-04 per-event succession is unchanged for every other consumer (the
// CLI reload, the echo); without a source the fold is a quiet no-op.
func (a *Actor) FoldAppliedConfig() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cfgSrc == nil {
		return
	}
	a.applySnapshot()
}

// armVerifyDeadline arms the pending round's verify budget through the
// timer seam (the caller holds the mutex, pending != nil): the callback is
// wrapped with the round's epoch tag (WR-02, plan 10-10) — the afterExpiry
// form on both arm points (selection and word). The tag assignment and the
// closure capture happen under the same critical section as the arm, so no
// interleaving can pair a callback with a foreign round.
func (a *Actor) armVerifyDeadline() {
	a.pendingEpoch++
	a.pending.epoch = a.pendingEpoch
	epoch := a.pending.epoch
	arm := a.verifyAfter
	if arm == nil {
		arm = defaultFlapAfter
	}
	a.pending.stop = arm(a.verifyWait, func() { a.VerifyExpiry(epoch) })
}

// armVerifyDeadWipe re-arms the single deferred wipe deadline of the flap
// window (the caller holds the mutex): a repeated FocusOut re-arms the SAME
// window instead of breeding timers (T-10-07-04), and the epoch advance
// turns a deadline from a superseded window stale (the afterExpiry form).
func (a *Actor) armVerifyDeadWipe() {
	a.cancelVerifyDeadWipe() // one window at a time, never a second timer
	// The window's context identity (plan 10-11, WR-03): one snapshot per
	// window — a repeated FocusOut re-arms the snapshot together with the
	// window (the internal cancel cleared it first).
	a.flapEng = a.emitter()
	epoch := a.flapEpoch
	arm := a.flapAfter
	if arm == nil {
		arm = defaultFlapAfter
	}
	a.flapStop = arm(focusFlapWindow, func() { a.verifyDeadWipeExpiry(epoch) })
}

// cancelVerifyDeadWipe stops a pending deferred wipe (the caller holds the
// mutex): the epoch advance turns an already-queued callback stale, so a
// racing expiry is a no-op — the completing SAME-ENGINE FocusIn and every
// immediate wipe (Reset, the own flip's synthetic pair) land here. The
// window's emitter snapshot clears with the slot (plan 10-11, WR-03).
func (a *Actor) cancelVerifyDeadWipe() {
	if a.flapStop != nil {
		a.flapStop()
		a.flapStop = nil
	}
	a.flapEpoch++
	a.flapEng = nil
}

// verifyDeadWipeExpiry is the deferred wipe's deadline callback: the flap
// window closed without a FocusIn — the focus loss is REAL and the wipe
// executes (a stale epoch — a newer FocusIn/FocusOut owns the window — is
// a no-op; the class learns and clears on transitions only, so the host's
// flap bursts never spam the journal).
func (a *Actor) verifyDeadWipeExpiry(epoch uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.flapEpoch != epoch {
		return // stale deadline — a newer FocusIn/FocusOut owns the window
	}
	a.flapStop = nil
	a.flapEng = nil // the window's context identity dies with the window (plan 10-11, WR-03)
	a.setVerifyDead(false)
}

// verifyDeadOnContextDeath applies the behavioral terminal class's wipe
// discipline of a context death (the caller holds the mutex; kind is a
// FocusOut or a Reset, synthetic marks the own flip's credit FocusOut). A
// fresh context starts unclassified — the next expired verify round
// re-learns the class. Plan 10-07 (G-10-1 cause A): a REAL FocusOut defers
// the wipe by focusFlapWindow — the host kernel's same-engine flap pairs
// (1–3 ms in the 16:20:05 journal) wiped the learned class before the next
// gesture could ride it, and wezterm's only classification source died with
// it. A Reset wipes immediately, byte as before — with the Round 2
// aftermath carve-out below. A SYNTHETIC FocusOut (the own flip's re-mint)
// wipes NOTHING (debug session orca-terminal-second-shift, green106
// 2026-10-31, 10/10 journal repros): the re-mint serves the SAME client
// field — the push-less verify behavior is a property of the app, not of
// the engine object — and wiping it made every terminal correction re-pay a
// verify round it always loses (the stationary «only the second
// double-shift works» cycle). The class rides the G-2-3 carve-out: the
// daemon's own flip is not a focus change. Round 2 (the owner acceptance of
// 2026-10-31 13:55, «работает со второго», kept by the client Reset despite
// the synthetic carve-out): the slow client reconnects its IM context
// asynchronously after the re-mint and delivers a Reset on the NEW engine
// object — the acceptance row's echo landed 923 ms after the flip and wiped
// the class through the Reset arm (the machine-verified journal
// elimination: all 22 wipe rows of the day ran the Reset branch). A Reset
// inside the flipAftermathWindow is the re-mint's own echo and rides the
// same carve-out; a real context change still clears: a real FocusOut
// cancels the aftermath window and takes the deferred wipe, and a Reset
// past the window wipes immediately, byte as before.
func (a *Actor) verifyDeadOnContextDeath(kind engine.LifecycleKind, synthetic bool) {
	switch {
	case kind == engine.LifecycleReset && a.flipAftermathStop != nil:
		// The re-mint's echo Reset: the same discipline as the synthetic
		// FocusOut — cancel the pending wipe, keep the class, spend the
		// window. Every other Reset side effect in HandleLifecycle (the
		// buffer hard-reset, the cache clears, the epoch bump, the pending
		// retirement) stays a full context reset — only the behavioral
		// class survives. The INFO row is the passive verification marker:
		// a `terminal class verify_dead=false` row between flip-released
		// and the next correction means a wipe OUTSIDE the aftermath — the
		// regression signature.
		a.cancelFlipAftermath()
		a.cancelVerifyDeadWipe()
		slog.Info("lifecycle", "kind", kind.String(), "aftermath", true)
	case kind == engine.LifecycleReset:
		a.cancelVerifyDeadWipe()
		a.setVerifyDead(false)
	case synthetic:
		// The pending window's snapshot is the old engine object — it can
		// never complete after the re-mint; cancel it, keep the class. The
		// aftermath window arms on: the echo Reset may still be seconds
		// away.
		a.cancelVerifyDeadWipe()
	default:
		// A REAL FocusOut ends the re-mint aftermath: a Reset arriving
		// after it is a genuine context change and wipes as before.
		a.cancelFlipAftermath()
		a.armVerifyDeadWipe()
	}
}

// armFlipAftermath re-arms the single re-mint-aftermath window of the
// daemon's own flip (the caller holds the mutex; the armVerifyDeadWipe
// single-window discipline): every successful flip round trip re-arms the
// same slot, so the client's echo Reset spends the CURRENT flip's
// aftermath whenever it lands.
func (a *Actor) armFlipAftermath() {
	a.cancelFlipAftermath()
	a.flipAftermathEpoch++
	epoch := a.flipAftermathEpoch
	arm := a.flipAftermathAfter
	if arm == nil {
		arm = defaultFlapAfter
	}
	a.flipAftermathStop = arm(flipAftermathWindow, func() { a.flipAftermathExpiry(epoch) })
}

// cancelFlipAftermath stops the pending re-mint-aftermath window (the
// caller holds the mutex): the client's echo Reset inside the window spent
// it, a real FocusOut ended it, or a newer flip re-arms the slot. The epoch
// advance turns an already-queued callback stale (the cancelVerifyDeadWipe
// form).
func (a *Actor) cancelFlipAftermath() {
	if a.flipAftermathStop != nil {
		a.flipAftermathStop()
		a.flipAftermathStop = nil
	}
	a.flipAftermathEpoch++
}

// flipAftermathExpiry is the aftermath window's deadline callback: the slow
// client never echoed — a Reset from now on is REAL and wipes immediately
// (a stale epoch — a newer window owns the slot — is a no-op, the
// verifyDeadWipeExpiry form).
func (a *Actor) flipAftermathExpiry(epoch uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.flipAftermathEpoch != epoch {
		return // stale deadline — a newer window owns the slot
	}
	a.flipAftermathStop = nil
}

// armLevel1Confirm opens the two-phase level-1 confirmation window (the
// caller holds the mutex, plan 10-11): the ladder's DeleteSurroundingText
// is already on the wire and the commit phase waits — either for the
// client's proof the deletion applied (the handleSurroundingLocked confirm
// check) or for this budget's deadline, whose fallback reproduces today's
// delete+commit pair with a loud INFO mark. One window at a time: arming
// cancels a still-pending budget of a superseded window (the
// armVerifyDeadWipe form) and the epoch advance turns its queued callback
// stale.
func (a *Actor) armLevel1Confirm(w *level1Confirm) {
	if a.l1 != nil && a.l1.stop != nil {
		a.l1.stop()
	}
	a.l1Epoch++
	w.epoch = a.l1Epoch
	a.l1 = w
	epoch := a.l1Epoch
	arm := a.l1After
	if arm == nil {
		arm = defaultFlapAfter
	}
	w.stop = arm(level1ConfirmBudget, func() { a.level1ConfirmExpiry(epoch) })
}

// level1ConfirmExpiry is the confirmation budget's deadline callback: the
// client stayed silent the whole budget, so the ladder degrades to today's
// delete+commit pair (the fallback is the CURRENT semantics, never a
// «deletion without replacement») with the loud INFO outcome mark —
// push-spontaneous clients confirm in milliseconds, so a timeout row is
// rare and worth hearing (T-10-11-04). A stale epoch (the window was
// flushed or re-armed) is a no-op — the afterExpiry form.
func (a *Actor) level1ConfirmExpiry(epoch uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.l1 == nil || a.l1.epoch != epoch {
		return // stale deadline — the window was flushed or re-armed
	}
	slog.Info("correction confirm", "outcome", "timeout") // loud and rare — the loud-mark fallback
	a.flushLevel1Confirm()
}

// flushLevel1Confirm is the ONE completion point of the two-phase level-1
// ladder (the caller holds the mutex, plan 10-11): the replacement commits
// through the window's CAPTURED emitter snapshot — never the current
// emitter(), so a flush after the context's death or re-mint cannot pour
// the text into another application's engine (T-10-11-02; the stale object
// is discarded by the client) — then the post-commit half runs in the
// today's order: buffer replace, the frozen done row, the DEFERRED flip
// (the 10-08 machine — the level-1 flip waits for its own delivery
// confirmation), the combo settle and the verify-after arm. Idempotent by
// the nil-check: every flush site (the confirm check, the budget fallback,
// the feedKey entry, FocusOut/Reset, a new gesture's entry) passes through
// untouched when no ladder is mid-flight.
func (a *Actor) flushLevel1Confirm() {
	if a.l1 == nil {
		return
	}
	w := a.l1
	a.l1 = nil
	if w.stop != nil {
		w.stop()
		w.stop = nil
	}
	w.eng.CommitText(engine.NewIBusText(string(w.commit)))
	w.rng.replace(w.converted) // the buffer keeps mirroring the field — repeat converts back
	a.logCorrectionDone(w.level, w.rng.token, w.converted, time.Since(w.armed))
	// The owner rule 2026-09-28, DEFERRED (plan 10-11): the mode follows
	// the converted script once the delivery confirms (the 10-08 machine).
	a.deferCorrectionFlip(w.converted, false) // the level-1 flip keeps the full 10-08 machine (plan 10-17)
	a.settleCombo()                           // D-36: the flip lands strictly after the settled completion record
	a.armAfterVerify(w.converted, w.rng.tail)
}

// soundSinkFor returns the installed sink when the sound switch is ON —
// the SINGLE gate of both tones (plan 07-08): with the switch off the sink
// is never even consulted, so no binary lookup and no spawn ever run for a
// muted daemon. nil (no sink installed, or the switch off) is the silent
// degradation. The caller holds the mutex.
//
//nolint:ireturn // the seam accessor hands the interface back (the emitter() precedent)
func (a *Actor) soundSinkFor() SoundSink {
	if !a.opts.SoundEnabled {
		return nil
	}

	return a.soundSink
}

// MenuSync is the tray-menu state seam (plan 07-05): the pushes the menu
// snapshot needs to mirror the actor — the mode symbol, the applied config
// truth and the corrections counter the Status info row renders (owner UAT
// 2026-10-05). Defined at the point of use; the interface travels with
// the consumer (the AppidSource precedent). The pushes ride the SAME
// discipline as ModeDisplay: synchronous under the actor's mutex, the
// implementation must stay quick (the menu dedupes identical values, so a
// redundant push is a cheap no-op) and must never panic or block.
type MenuSync interface {
	SetMode(symbol string)
	SetAutocorrectEnabled(on bool)
	SetSoundEnabled(on bool)
	SetKeys(tap, combo, chord string)
	SetCorrections(n int)
	SetBlocklistApp(app string, fresh, covered, full bool)
}

// emitter returns the attached emitter sink — the snapshot accessor of the
// atomic slot. Nil before the first mint, exactly like the plain a.eng
// field it replaced; every former a.eng read takes a local snapshot through
// it at the same computation point (the snapshots are atomic, so the
// mutex-holding discipline around them is unchanged).
//
//nolint:ireturn // the slot stores the emitter seam; the accessor hands the interface back (G-6-1)
func (a *Actor) emitter() engine.Emitter {
	if slot := a.engSlot.Load(); slot != nil {
		return slot.eng
	}

	return nil
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

// deferCorrectionFlip arms a post-correction mode flip instead of
// executing it (plan 10-08, G-10-1 cause B, the E3 verdict): the decision
// reuses settleCorrectionFlip's EXACT guards — a combo gesture keeps its
// own settleCombo flip, the off gate arms nothing, and a target equal to
// the current mode is a no-op flip, so nothing is armed. The terminal
// flag (plan 10-17) selects the release machine: the TERMINAL flip arms
// NO grace deadline — SetGlobalEngine must be impossible between the
// commit and the flip's honest release cause (the next keystroke or a
// confirming push; the round-2 hypothesis: the +104 ms flip re-mints the
// engine and drives the client to flush its forward queue) — and it is
// immune to the verify outcomes. The level-1 flip (flushLevel1Confirm)
// keeps the full 10-08 machine byte-as-today, the grace deadline its last
// release site (T-10-08-02). The owner rule 2026-09-28 «режим следует за
// скриптом результата» is deferred, not cancelled: the flip still follows
// the result — on the next gesture; the release-site narrowing is
// owner-validated in 10-18. One slot at a time — arming cancels a
// still-pending deadline (a second correction can reach this site only
// off the keystroke path, e.g. CorrectNow, and its own flip supersedes
// the pending one). The caller holds the mutex.
func (a *Actor) deferCorrectionFlip(converted []rune, terminal bool) {
	if a.comboPending || !a.opts.FlipAfterCorrection {
		return
	}
	if flip := a.flipPending; flip != nil && flip.gesture {
		// CR-01: a gesture-parked flip owns the slot — the owner's explicit
		// command arrived LATER than the correction's own mode rule and
		// supersedes it; re-arming here would silently drop the parked
		// gesture (the slot is one, the gesture wins, the drop is the one
		// outcome the mandate forbids). The parked flip keeps its own
		// release causes (the next keystroke or a confirming push).
		return
	}
	target := scriptOf(converted)
	if target == a.mode {
		return
	}
	a.cancelFlipGrace()
	a.flipEpoch++
	a.flipPending = &deferredFlip{
		target:   target,
		epoch:    a.flipEpoch,
		terminal: terminal,
		armedAt:  time.Now(),
		seq:      a.emissionSeq,
		start:    a.deliveryStart,
	}
	if terminal {
		a.deliveryMark("flip-armed")

		return
	}
	epoch := a.flipEpoch
	arm := a.flipAfter
	if arm == nil {
		arm = defaultFlapAfter
	}
	a.flipStop = arm(flipConfirmGrace, func() { a.flipGraceExpiry(epoch) })
}

// releaseDeferredFlip executes an armed deferred flip exactly once (the
// caller holds the mutex): flipTo's semantics are reused WHOLE — the D-36
// record order (mode record → switcher → panel symbol) is flipTo's own,
// never duplicated — and the slot clears before the call, so every release
// site (the confirming push, the next keystroke, the grace deadline, the
// verify settlements) is idempotent. A nil slot is a no-op: the release
// sites on the shared paths (settleAfter, afterExpiry, feedKey) pass
// through untouched for the GTK corrections and the flip-disabled
// terminals. A released TERMINAL flip carries its delivery mark — the
// release reason and the pending latency, outcomes only, never field
// content (D-20/D-21); the level-1 flip's journal stays byte-as-today.
// The caller holds the mutex.
func (a *Actor) releaseDeferredFlip(reason string) {
	if a.flipStop != nil {
		a.flipStop()
		a.flipStop = nil
	}
	if a.flipPending == nil {
		return
	}
	flip := a.flipPending
	a.flipPending = nil
	if flip.terminal {
		a.deliveryMarkBurst(flip.seq, flip.start, "flip-released",
			"reason", reason, "latency_ms", time.Since(flip.armedAt).Milliseconds())
	}
	a.flipTo(flip.target)
}

// releaseDeferredFlipOnVerifyOutcome is the verify-outcome release site,
// branchy under plan 10-17: a TERMINAL deferred flip IGNORES the verify
// outcomes — the verify machinery judges the field, the mode rule waits
// for its own cause (the next keystroke or a confirming push; a
// SetGlobalEngine riding a verify settlement is exactly the round-2 loss
// geometry) — while every other flip keeps the byte-as-today release. The
// caller holds the mutex.
func (a *Actor) releaseDeferredFlipOnVerifyOutcome() {
	if flip := a.flipPending; flip != nil && flip.terminal {
		return
	}
	a.releaseDeferredFlip("verify")
}

// cancelFlipGrace stops a pending grace deadline (the caller holds the
// mutex): the epoch advance turns an already-queued callback stale — the
// afterExpiry form.
func (a *Actor) cancelFlipGrace() {
	if a.flipStop != nil {
		a.flipStop()
		a.flipStop = nil
	}
	a.flipEpoch++
}

// flipGraceExpiry is the armed flip's grace deadline: no confirming push and
// no keystroke arrived, so the mode rule must not wait forever — the slot
// cannot stick (T-10-08-02), the flip fires. A stale epoch (the flip was
// already released, or a newer arm owns the slot) is a no-op — the
// afterExpiry form.
func (a *Actor) flipGraceExpiry(epoch uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.flipPending == nil || a.flipPending.epoch != epoch {
		return // stale deadline — the flip was released or re-armed
	}
	a.releaseDeferredFlip("grace")
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
	a.opts.Terminals = snap.Correction.Terminals // the terminal-class list folds live (Phase 10, D-55)
	a.applySnapshotBindings(snap)
	a.opts.MACREnabled = snap.MACR.Enabled
	a.opts.MACRApps = snap.MACR.Apps
	a.opts.MACRAltModifier = snap.MACR.AltModifier
	if snap.MACR.Letters != a.macrLettersName {
		a.macrLettersName = snap.MACR.Letters
		a.opts.MACRLetters = parseMACRLetters(snap.MACR.Letters)
	}
	// The autocorrect section folds live (plan 06-06, the MACR-fold
	// precedent): the D-54 defaults are off, so a document without the
	// section keeps the boundary byte-as-today; the thresholds mirror
	// detect.Params (the config 06-04 pair — change the places together).
	// The blocklist recompiles only when the document's pattern list
	// changed (one compile per edit, not per boundary).
	a.opts.AutoCorrectEnabled = snap.Autocorrect.Enabled
	a.opts.AutoCorrectBlocklist = snap.Autocorrect.AppsBlocklist
	a.opts.AutoCorrectMinWordLen = snap.Autocorrect.MinWordLen
	a.opts.AutoCorrectMargin = snap.Autocorrect.TrigramMargin
	a.opts.AutoCorrectFloor = snap.Autocorrect.TrigramFloor
	a.refreshACBlocklist(snap.Autocorrect.AppsBlocklist)
	a.refreshACTerminals(snap.Correction.Terminals)
	a.ensureAppid() // the per-app list or the enabled autocorrect layer may have appeared with this document
	// The sound switch folds live (plan 07-08, the autocorrect-fold
	// precedent): the owner's default ON is the EffectiveEnabled truth —
	// an absent section reads ON, only an explicit enabled: false mutes,
	// and the next gesture carries the change (the D-32 hot reload).
	a.foldSound(snap.Sound.EffectiveEnabled())
	a.pushA11y(snap) // the a11y-magic fold rides the same per-fold push chain (plan 08-05)
	a.pushMenuSync(snap)
}

// applySnapshotBindings re-resolves the NAME-keyed hotkey bindings of the
// snapshot (the word-layout combo, the mode-switch chord) — the fold half
// of applySnapshot, extracted whole: a binding re-resolves only when the
// document's binding NAME changed; a name that fails to parse — impossible
// from a validated document — keeps the last-good binding (the D-32
// discipline). The caller holds the mutex.
func (a *Actor) applySnapshotBindings(snap config.Config) {
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
}

// foldSound applies the snapshot's sound switch (plan 07-08's fold, the
// 261006-squ mute lifecycle): the owner's default ON is the
// EffectiveEnabled truth, and the ON→OFF TRANSITION is the close trigger
// — the installed sink stops (muted sound holds no sound-server
// connection; the lazy redial rides the next tone). The gate in
// soundSinkFor keeps its exact nil form — the fold transition closes,
// the gate remains the single play gate. The caller holds the mutex.
func (a *Actor) foldSound(enabled bool) {
	if a.opts.SoundEnabled && !enabled {
		if stopper, ok := a.soundSink.(SoundStopper); ok {
			stopper.Stop() // quick and re-entrant-safe under the sink's own mutex
		}
	}
	a.opts.SoundEnabled = enabled
}

// pushA11y hands the reconciler the folded a11y state (plan 08-05, D-8-3
// — the pushMenuSync form; revised by the owner's 2026-10-06 decision,
// plan 08-09): the fold reads ONLY snap.A11y.EffectiveEnabled() — the
// single definition of the desired a11y state, never recomputed here —
// and pushes it to the sink only on a CHANGE of that boolean (the
// refreshACBlocklist diff-gate form): a repeated fold of the same state
// never touches the sink, so dconf churn and journal noise are impossible
// at any fold rate. The push rides the same synchronous under-the-mutex
// discipline as the sound pushes — legal because the sink contract is
// fire-and-forget (a field write plus its own spawn, 08-03): the
// subprocess series serializes on the reconciler's OWN mutex, never under
// the actor's (WR-01) — the actor.mu → sink.mu ordering is one-way (the
// sink never calls the actor back, the pushMenuSync comment's form).
// The caller holds the mutex.
func (a *Actor) pushA11y(snap config.Config) {
	if active := snap.A11y.EffectiveEnabled(); active != a.a11yActive {
		a.a11yActive = active
		if a.a11ySink != nil {
			a.a11ySink.Apply(active)
		}
	}
}

// pushMenuSync hands the menu the applied config truth (plan 07-05): both
// toggles' applied values, the RAW key names and the corrections counter
// the Status row renders (owner UAT 2026-10-05 — the counter rides the
// same per-fold push, so the row catches up with the next fold after a
// correction lands). The sound value is the EffectiveEnabled truth — an
// absent section reads ON (the owner's default-ON verdict) — and the
// effective autocorrect event rides the same fold to the sound sink on
// change (WR-02). The menu dedupes identical pushes, so the per-fold push
// of unchanged values is cheap. The caller holds the mutex.
func (a *Actor) pushMenuSync(snap config.Config) {
	a.soundEnabled = snap.Sound.EffectiveEnabled()
	// The autocorrect tone's event folds live too (WR-02): every changed
	// effective value is pushed to the sink — the actor keeps the
	// last-pushed value as the dedupe key (the menu's dedupe discipline),
	// so an unchanged document costs one comparison. The push rides the
	// same synchronous under-the-mutex discipline as the menu pushes; the
	// sink's setter is a field write behind its own mutex, and the
	// actor.mu → sink.mu ordering is one-way (the sink never calls back).
	if event := snap.Sound.EffectiveAutocorrectEvent(); event != a.soundACEvent {
		a.soundACEvent = event
		if a.soundSink != nil {
			a.soundSink.SetAutocorrectEvent(event)
		}
	}
	if a.menuSync != nil {
		a.menuSync.SetAutocorrectEnabled(snap.Autocorrect.Enabled)
		a.menuSync.SetSoundEnabled(a.soundEnabled)
		a.menuSync.SetKeys(snap.Hotkeys.TapKey, snap.Hotkeys.WordLayoutCombo, snap.Hotkeys.ModeSwitchChord)
		a.menuSync.SetCorrections(a.corrDone)
	}
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
	// The DST guard's materialization evidence (plan 10-16, the 2026-10-16
	// bisection): EVERY push or materializing Require answer of the current
	// generation proves the composer-side surrounding state exists — the
	// bisection made materialization, not emission parameters or timing, the
	// crash discriminator (R1b warm survive vs R2/R6b/R7 virgin crash). The
	// evidence binds to THIS generation's emitter (the flapEng identity
	// form), so a churn-missed re-mint cannot inherit a dead generation's
	// proof; it dies with the cache in the FocusOut/Reset branch.
	a.surrSeen = true
	a.surrEng = a.emitter()
	// -debug only, same sensitivity class as the key trace (diagnosis mode
	// is opt-in and short-lived); the live hunt needs the pushed field
	// text to see what a correction actually left behind (2026-09-28).
	slog.Debug("surrounding push", "text", text,
		"cursor", cursorPos, "anchor", anchorPos)
	// The level-1 confirm check (plan 10-11, the two-phase delivery): a
	// ladder is mid-confirmation — its delete is on the wire and this push
	// is the client's report. A text-before-cursor that no longer ends
	// with token+tail PROVES the deletion applied and reported — the commit
	// phase runs now, through the window's captured snapshot. A
	// still-matching push is the client's stale pre-delete state — the
	// window lives to its budget. The window is separate state from the
	// pending round (resolvePending happened in executeCorrection), so the
	// check sits ahead of the pending branch and consumes the push: the
	// verify-after the flush arms settles on the client's NEXT report, not
	// on this pre-commit one.
	if a.l1 != nil {
		if !correct.MatchesSuffix(a.surr, a.l1.match) {
			slog.Debug("correction confirm", "outcome", "confirmed")
			a.flushLevel1Confirm()
		}

		return nil, false
	}
	if a.pending != nil {
		p := a.pending
		switch {
		case a.pendingVerdict():
			if pendingStaleCache(p) {
				// The answer-only route (plan 10-10, G-10-1 round 2, the
				// 09:39 host rows): the same stale-cache signature as the
				// mismatch branch below, now on the MATCH verdict — the
				// client never pushed a field state (empty cache at arm)
				// and its single answer holds what the daemon watched being
				// typed. The 2026-10-13 owner-approved semantics extend
				// from the answer's CONTENT (10-08) to its CHANNEL: the
				// host's wezterm/GNOME-Terminal clients answered fresh and
				// matching, yet level-1 (done in 1.9–2.4 ms) is exactly
				// where both 10-07/10-08 fixes physically do not stand.
				// Byte-the-same form as the mismatch branch: the class
				// learns, the round routes to the terminal write branch.
				a.resolvePending()
				a.setVerifyDead(true)
				a.executeLevel2(p.rng, p.converted, p.armed, true)

				return nil, false
			}
			a.executeCorrection()
		case a.pendingEarlyAnswer():
			// The pre-boundary early answer of an autocorrect round: the
			// client has not applied the separator yet — the round stays
			// open within its budget for the post-boundary push (06-07
			// live finding; VerifyExpiry keeps the fail-closed bound).
		default:
			if pendingStaleCache(p) {
				// The 2026-10-13 client-drift amendment (plan 10-08, owner
				// verdict, G-10-1): the client never reported a field state
				// (empty cache at arm) and its only answer does not hold
				// what the daemon watched being typed — the stale-cache
				// terminal signature. The behavioral class learns and the
				// round routes to the terminal write branch; the word
				// corrects on this round instead of dying a GTK mismatch.
				a.resolvePending()
				a.setVerifyDead(true)
				a.executeLevel2(p.rng, p.converted, p.armed, true)

				return nil, false
			}
			a.resolvePending()
			a.skipCorrection("verify-mismatch")
			a.settleCombo() // the word half settled by refusal — the D-36 flip still fires
		}

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

// pendingStaleCache reports the stale-cache terminal signature of the
// 2026-10-13 client-drift amendment (plan 10-08, owner verdict, G-10-1),
// extended by plan 10-10 to BOTH verdict branches: the round armed over an
// EMPTY push cache — the client has never reported a field state in this
// context episode (an honest GTK client pushes on every change), the
// ANSWER-ONLY push channel — whatever its single answer holds. Content
// does not distinguish: a fresh matching answer routed level-1 on the host
// (the 09:39 rows — done in 1.9–2.4 ms, the branch where the 10-07/10-08
// fixes physically do not stand), a stale non-matching one died a GTK
// verify-mismatch (the 0x29-drifted wezterm nightly). The channel
// distinguishes, per the owner's approved semantics («обещает окружение,
// но отвечает протухшим кэшем = всё равно терминал»). Cache evidence (the
// pushes tracked the typing) keeps the byte-as-today fail-closed world;
// selection rounds are range-anchored (Pitfall 6) and autocorrect boundary
// rounds carry their own CR-02 fail-closed semantics — both excluded.
func pendingStaleCache(p *pendingFix) bool {
	return p.sel == nil && !p.rng.acBoundary && p.cacheNil
}

// pendingEarlyAnswer reports whether the fresh push is an armed
// autocorrect round's PRE-boundary state — the honest client state the
// verify's RequireSurroundingText answer arrives in (plan 06-07 live
// finding): the Require is served before the separator key reaches the
// client, so the first answer holds the word but not yet its tail —
// whatever the cache's own freshness was at arm time. The round stays
// open within its budget; the in-flight separator's post-boundary push
// settles it, and VerifyExpiry keeps the fail-closed bound. A foreign
// answer (neither the match nor the token-only pre-state) still concludes
// the mismatch. The caller holds the mutex and pending != nil.
func (a *Actor) pendingEarlyAnswer() bool {
	return a.pending.sel == nil && a.pending.rng.acBoundary &&
		correct.MatchesSuffix(a.surr, a.pending.rng.token) &&
		!correct.MatchesSuffix(a.surr, a.pending.match)
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
	eng := a.emitter() // one atomic snapshot for both emitter reads
	if a.afterVerdict(round.expected, round.start, round.atRange) {
		a.clearAfter()
		slog.Debug("correction verify", "outcome", "match")
		a.releaseDeferredFlip("push") // the confirming push: delivery proven — the deferred flip fires now

		return nil, false
	}
	if round.mismatches == 0 {
		round.mismatches = 1
		a.rearmAfter(round)
		slog.Debug("correction verify", "outcome", "stale-retry")
		eng.RequireSurroundingText()

		return nil, false
	}
	a.clearAfter()
	slog.Info("correction verify", "outcome", "mismatch")
	a.releaseDeferredFlipOnVerifyOutcome() // branchy (plan 10-17): the terminal flip ignores the verify outcomes
	if round.noRung {
		// The terminal branch (plan 10-04, ADR-008 D4, owner verdict OQ 2
		// «contract»): the mismatch LOGS and STOPS — the round carries no
		// paste payload, so the rung condition is never even evaluated
		// (Pitfall 3: the paste chord is forbidden by construction, under
		// every configuration).
		return nil, false
	}

	return round.paste, a.opts.ClipboardRung && a.clip != nil && eng != nil
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
	a.releaseDeferredFlipOnVerifyOutcome() // branchy (plan 10-17): the terminal flip ignores the verify outcomes
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
	a.emitter().RequireSurroundingText()
}

// armAfterVerifyTerminal starts the verify-after round of the terminal
// write branch (plan 10-04 Task 3, owner verdict OQ 2 «contract»): the
// SAME epoch, verify budget and quiet-timeout machinery as armAfterVerify —
// the client that does push (VTE) settles honestly, a push-less context
// dies by budget — but the round carries NO paste payload: a mismatch can
// only log and stop (settleAfter's noRung branch), the clipboard rung is
// forbidden by construction, never by config (ADR-008 D4). The caller
// holds the mutex.
func (a *Actor) armAfterVerifyTerminal(converted, tail []rune) {
	a.verifyEpoch++
	a.clearAfter() // never two open rounds
	a.after = &pendingAfter{
		expected: concatRunes(converted, tail),
		epoch:    a.verifyEpoch,
		noRung:   true,
	}
	epoch := a.verifyEpoch
	a.after.deadline = time.AfterFunc(a.verifyWait, func() { a.afterExpiry(epoch) })
	a.emitter().RequireSurroundingText()
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
	a.emitter().RequireSurroundingText()
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
	// The rung's own snapshots: clip under a brief lock (the verdict armed
	// it under the same guard one call earlier), eng from the atomic slot
	// without any lock (G-6-1); a mid-window engine swap only retargets the
	// burst to the live sink.
	a.mu.Lock()
	clip := a.clip
	a.mu.Unlock()
	eng := a.emitter()
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
	if !a.opts.MACREnabled || !a.macrTargetActive() || a.emitter() == nil {
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
	if eng := a.emitter(); a.macrPendingKeyval != 0 && eng != nil {
		pending := a.macrPendingKeyval
		a.macrPendingKeyval = 0
		forwardCtrlLetter(eng, pending, a.opts.MACRAltModifier)

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
// NON-EMPTY macr list is in force OR autocorrect is ENABLED (the 07-04
// condition — the blocklist negatives need an identity to refuse on, and
// the confirm's live role query needs the observer as its RoleSource), and
// only once per actor (ADR-005 a: the a11y bus is connected IFF per-app
// scope or the enabled autocorrect layer can need it). A failed start WARNs
// and leaves the source nil — the quiet unknown-identity pass-through
// covers the gap for good (one documented degradation, not a retry loop).
// The caller holds the mutex.
func (a *Actor) ensureAppid() {
	if a.appidStarted || (len(a.opts.MACRApps) == 0 && !a.opts.AutoCorrectEnabled) {
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
// G-6-1: the factory path (CreateEngine → AttachEngine) is freed from a.mu
// by contract — AttachEngine stores the emitter slot atomically, so ibus's
// re-entrant CreateEngine during THIS await no longer self-deadlocks on the
// actor mutex. The await itself stays under the mutex: the "sync, not the
// async WR-01-handoff" pin stands — the D-36 record order and the final
// correctness of rapid flips hold by the synchronous serialization.
//
// ADR-004 amendment 2026-10-07 (G-2-3): a SUCCESSFUL switcher round trip
// arms one synthetic-lifecycle credit — ibus answers the flip by re-minting
// engine objects, and the synthetic focus_out/focus_in pair then reaches
// HandleLifecycle, which consumes the credit and skips exactly the buffer
// hard-reset (the daemon's own flip is not a focus change; the word typed
// across it corrects as a whole). A WARNed failure or the nil-seam
// degradation arms nothing — the next FocusOut stays a real focus loss.
// Round 2 (orca-terminal-second-shift): the same round trip also arms the
// flipAftermathWindow slot — the slow client's echo Reset of the re-mint
// lands within it and keeps the behavioral class (see HandleLifecycle).
//
// A flipTo to the CURRENT mode is a no-op: the SET-semantics site
// (settleCorrectionFlip) names its target from the corrected text's script,
// and a same-script correction flip changes nothing.
//
// CR-01 (the 10-REVIEW, the round-4 fix session): while the delivery
// machine is in flight (level2Busy — the drain-settle window holds the
// mutex RELEASED, so every gesture entry converges here INSIDE the window),
// flipTo must not execute: SetGlobalEngine between the forwards and the
// still-pending commit re-mints the engine objects and orphans the commit
// riding the captured snapshot — the E3 verdict, «удаление без замены»,
// the one outcome the ADR-008 amendment (6) forbids. The flip PARKS into
// the deferred slot with the terminal release discipline (parkGestureFlip):
// the mode switch happens, just after the correction settles. The gate
// covers every gesture entry by construction — the combo's busy-refusal
// settleCombo, the mode-switch chord, the Single-tap expiry, the tray
// ToggleMode/SwitchMode all reach this one seam.
func (a *Actor) flipTo(target scriptMode) {
	if target == a.mode {
		return
	}
	if a.level2Busy {
		a.parkGestureFlip(target)

		return
	}
	a.mode = target
	slog.Info("mode", "to", a.modeSymbol())

	if sw := a.switcher; sw != nil {
		ctx, cancel := context.WithTimeout(context.Background(), switchTimeout)
		defer cancel()
		if err := sw(ctx, engineNameOf(target)); err != nil {
			slog.Warn("engine switch failed", "engine", engineNameOf(target), "error", err)
		} else {
			// ADR-004 amendment 2026-10-07 (G-2-3): a SUCCESSFUL
			// SetGlobalEngine round trip makes ibus re-mint engine objects —
			// a synthetic focus_out/focus_in pair is on its way to the
			// engine. Arm exactly one credit; the FocusOut branch of
			// HandleLifecycle consumes it and skips exactly the buffer
			// hard-reset (the daemon's own flip is not a focus change). A
			// failed round trip arms nothing — the buffer then treats the
			// next FocusOut as a real focus loss.
			a.flipCredit++
			// Round 2 (orca-terminal-second-shift): the re-mint aftermath
			// window arms with the credit — the slow client reconnects its
			// IM context asynchronously and its echo Reset lands seconds
			// after the synthetic pair (the green106 row: +923 ms).
			a.armFlipAftermath()
		}
	} else if !a.switcherWarned {
		a.switcherWarned = true
		slog.Warn("switcher unavailable", "engine", engineNameOf(target))
	}

	if eng := a.emitter(); eng != nil {
		eng.UpdateModeSymbol(a.modeSymbol())
	}

	// The display observer fires LAST — after the panel symbol, the D-36
	// order with the display appended (quick 260930-pf6). The indicator
	// degrades itself; the actor adds no error handling around the call.
	if a.display != nil {
		a.display.ModeChanged(a.modeSymbol())
	}
	// The menu observer fires with the display — the second appended
	// observer of the same record (plan 07-05): the radio pair re-marks
	// itself from the same push the icon uses. The menu dedupes identical
	// symbols; the actor adds no error handling around the call.
	if a.menuSync != nil {
		a.menuSync.SetMode(a.modeSymbol())
	}
	// The sound observer fires LAST — after the display and the menu, the
	// D-36 order with every observer appended (plan 07-08): the tone
	// confirms a flip that fully happened (the same-target guard above
	// already silenced the no-ops). The single gate consults the sink only
	// when the sound switch is on; the sink is fire-and-forget by
	// contract, and the actor adds no error handling around the call.
	if s := a.soundSinkFor(); s != nil {
		s.Flip()
	}
}

// parkGestureFlip arms a gesture-driven flip that arrived while the delivery
// machine is in flight (CR-01 of the 10-REVIEW): the flip rides the EXISTING
// deferred-flip slot with the TERMINAL release discipline — no grace
// deadline, immune to the verify outcomes, released by the next keystroke or
// a confirming push — so the SetGlobalEngine fires only after the pending
// commit landed (the machine's own exit clears level2Busy before the release
// sites can run). NOT gated on flip_after_correction: the gesture flip is the
// owner's explicit command — settleCombo's own flip never consulted the gate
// and neither does the park. The target arrives as a scriptMode (the gesture
// entries are toggles, not converted text). The slot is one: a second park
// supersedes the first (the LATEST gesture is the owner's intent), and the
// machine's post-correction arming skips a gesture-occupied slot
// (deferCorrectionFlip) — the gesture is never silently dropped. The caller
// holds the mutex.
func (a *Actor) parkGestureFlip(target scriptMode) {
	if target == a.mode {
		return
	}
	a.cancelFlipGrace()
	a.flipEpoch++
	a.flipPending = &deferredFlip{
		target:   target,
		epoch:    a.flipEpoch,
		terminal: true,
		gesture:  true,
		armedAt:  time.Now(),
		seq:      a.emissionSeq,
		start:    a.deliveryStart,
	}
	// The loud boundary (D-20/D-21): outcomes only, never content — the
	// journal's proof that the gesture was captured, not lost.
	a.deliveryMark("flip-parked")
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
	if eng := a.emitter(); eng != nil {
		eng.UpdateModeSymbol(a.modeSymbol())
	}
	// The display observer fires last, mirroring flipTo (quick 260930-pf6):
	// the icon follows the FACTUAL engine, not only the daemon's own flips.
	if a.display != nil {
		a.display.ModeChanged(a.modeSymbol())
	}
	// The menu observer fires with the display, mirroring flipTo (plan
	// 07-05): the radio pair follows the FACTUAL engine too.
	if a.menuSync != nil {
		a.menuSync.SetMode(a.modeSymbol())
	}
}

// modeSymbol returns the panel symbol of the current script mode — the
// glyph the mode-indicator property carries and the value both observers
// receive (owner decision 1; plan 07-05). The caller holds the mutex.
func (a *Actor) modeSymbol() string {
	if a.mode == modeRU {
		return symbolRU
	}

	return symbolEN
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
	// The two-phase level-1 ladder's flush site (plan 10-11): the
	// replacement is on the wire before the next input processes — a
	// keystroke arriving without the client's push is the delivery proof
	// of last resort. The flip the flush arms releases right below
	// (strictly before the key's processing — the 10-08 release order).
	// A nil window is a no-op.
	a.flushLevel1Confirm()

	// The deferred terminal flip's second confirmation site (plan 10-08, E3;
	// plan 10-17 made it the FIRST cause): the first keystroke after the
	// correction proves the commit delivered — release STRICTLY before
	// processing, so the key already types in the RESULT mode. A nil slot
	// is a no-op.
	a.releaseDeferredFlip("keystroke")

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
		// decision is captured BEFORE the reset reads the finished word and
		// launched AFTER the hard reset — the payload's staleness
		// expectation is therefore the emptied buffer (CR-01), so the
		// confirm accepts the boundary's own reset and refuses anything
		// typed since. The arming never touches the transit or the reset
		// semantics. The FocusOut lifecycle is NOT a boundary — the word
		// was not finished by a keystroke (06-RESEARCH Q5) — so autocorrect
		// stays silent there.
		payload, armed := a.autoCorrectBoundary()
		a.buf.HardReset()
		a.armAutoCorrect(payload, armed)

		return false
	case !printableKeyval(ev.Keyval) || ev.Mods&comboMask != 0:
		// Non-printables (bare modifiers included) and true combos put no
		// rune in the field — the buffer must not see them either.
		return false
	case a.mode == modeRU:
		// #nosec G115 -- printableKeyval bounds the keyval below
		// 0xFE00, so the uint32→rune conversion cannot overflow.
		r := rune(ev.Keyval)
		eng := a.emitter()
		if ru, ok := layouts.ENToRU[r]; ok && ru != r && eng != nil {
			eng.CommitText(engine.NewIBusText(string(ru)))
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
		payload, armed := a.autoCorrectBoundary()
		a.armAutoCorrect(payload, armed)
	}
}

// armAutoCorrect launches the confirm for an already-captured boundary
// decision (plan 06-06): the generation is tagged under the mutex so a
// newer boundary supersedes a mid-flight confirm, and the confirm goroutine
// leaves the keystroke path immediately — the armed-payload discipline of
// the clipboard rung (T-06-06-04/T-06-06-05). The staleness expectation is
// snapshotted HERE, after the boundary's own mutation completed (the reset
// branch of feedKey hard-resets between the capture and this launch —
// CR-01): the confirm revalidates the live buffer against it, so the
// boundary's own effect passes and every later buffer change refuses the
// payload fail-closed. The caller holds the mutex.
func (a *Actor) armAutoCorrect(payload acPayload, armed bool) {
	if !armed {
		return
	}
	payload.expectToken = a.buf.Token()
	payload.expectTail = a.buf.Tail()
	a.acGeneration++
	payload.gen = a.acGeneration
	go a.autoConfirm(payload)
}

// autoCorrectBoundary decides one word boundary under the CHEAP half of
// the D-53 conjunction (the caller holds the mutex — feedKey): enabled,
// the surrounding-text capability, the focused app's bridge-namespace
// identity against the BLOCKLIST (any regex pattern matching the identity
// FORBIDS the correction; an empty list forbids nothing) and the
// detector's CONFIDENT wrong-layout verdict on the finished token. The
// 07-04 revision (locked 07-CONTEXT) inverts EXACTLY ONE segment: an
// UNKNOWN identity — no source started, or no focus event yet — is NOT a
// prohibition: the identity is omitted from the payload (no app field),
// the detector still runs, and the confirm gate plus the payload-stale
// buffer geometry carry the safety story. A KNOWN identity matching any
// pattern refuses with app-blocked. Role, caps and the detector stay
// fail-closed — the pinned direction is
// TestAutoCorrect_IdentityUnknownNotProhibition. The expensive half — the
// live GetRole — never runs here: the armed payload hands the decision
// outside the mutex (the off-mutex discipline, T-06-06-04). The off state
// is the zero-behavior invariant: no counter, no record — the boundary is
// byte-as-today (D-54).
func (a *Actor) autoCorrectBoundary() (acPayload, bool) {
	if !a.opts.AutoCorrectEnabled {
		return acPayload{}, false
	}
	if a.caps&correct.CapSurroundingText == 0 {
		// D-53 condition 3: without the bit the verify rung is not
		// applicable — the word is never touched.
		a.recordACAbstain(acReasonNoCaps)

		return acPayload{}, false
	}
	// The 07-04 inversion: an unknown identity PASSES — only a KNOWN
	// identity can be forbidden, the payload simply carries no app.
	app, appKnown := a.acFocusedApp()
	if appKnown {
		// The blocklist matches by regex SUBSTRING over the identity (RE2
		// — linear, no catastrophic backtracking; ^…$ anchoring is the
		// document's explicit choice) — the compiled cache from
		// applySnapshot, never the raw strings. Any matching pattern
		// forbids.
		for _, re := range a.acBlocklist {
			if re.MatchString(app) {
				a.recordACAbstain(acReasonAppBlocked)

				return acPayload{}, false
			}
		}
	}
	// The D-55 terminal veto (TERM-04, the invariant's literal placement):
	// the terminal class is checked at the ARMING point, ahead of the
	// detector — a terminal context never arms the automatic layer. The
	// class does not depend on identity observability (fail-closed: the
	// behavioral verify-dead source alone vetoes an invisible context),
	// and it deliberately does NOT live in the blocklist — a blocklist
	// cannot exclude the invisible (wezterm is outside the a11y bus). The
	// confirm's role switch below stays as the second belt, untouched.
	if a.classifyTerminal() {
		a.recordACAbstain(acReasonTerminalVeto)

		return acPayload{}, false
	}
	tok := a.buf.Token()
	v := a.detectVerdict(tok)
	if !v.WrongLayout || !v.Confident {
		// D-53 condition 4 (with the detector's own length gate): an unsure
		// verdict, a dictionary veto or a short token stays untouched — the
		// detector's closed reason slug is the only trace.
		a.recordACAbstain(v.Reason)

		return acPayload{}, false
	}

	return acPayload{
		rng: correctionRange{
			token:      tok,
			tail:       a.buf.Tail(),
			replace:    a.buf.ReplaceToken,
			acBoundary: true,
		},
		dir:  v.Dir,
		word: tok,
		app:  app,
	}, true
}

// classifyTerminal answers the one question both consumers of the
// terminal class share — the arming veto (autoCorrectBoundary, plan 10-03)
// and the terminal write pipeline (plan 10-04): is this input context a
// terminal? One seam, three sources, cheap-to-expensive (RESEARCH
// Pattern 1):
//
//  1. the behavioral verify-dead class — a verify round of this context
//     expired by budget while the context lived (learned in VerifyExpiry,
//     cleared by HandleLifecycle); it needs no identity at all;
//  2. the document's terminal identity patterns (correction.terminals) —
//     substring RE2 against the observable identity; an invisible identity
//     answers nothing from this source (the quiet norm, not an error);
//  3. the live AT-SPI terminal role 60, queried at the decision moment
//     through the RoleSource under the autoRoleTimeout deadline — the
//     T-10-03-03 wedge guard: a stuck a11y object costs the same bounded
//     budget the confirm already pays, never an unbounded stall. A role
//     ERROR answers "not terminal" here — the confirm's role-unknown
//     discipline stays the authority on an unusable source.
//
// The caller holds the mutex.
func (a *Actor) classifyTerminal() bool {
	if a.verifyDead {
		return true
	}
	if app, known := a.acFocusedApp(); known && a.matchACTerminal(app) {
		return true
	}
	if rs := a.role; rs != nil {
		ctx, cancel := context.WithTimeout(context.Background(), autoRoleTimeout)
		defer cancel()
		role, err := rs.Role(ctx)
		if err == nil && role == acRoleTerminal {
			return true
		}
		if err != nil {
			a.warnACRole(err)
		}
	}

	return false
}

// matchACTerminal reports whether the identity matches any compiled
// terminal-class pattern (substring RE2 — the blocklist semantics, the
// author anchors). The caller holds the mutex.
func (a *Actor) matchACTerminal(app string) bool {
	for _, re := range a.acTerminals {
		if re.MatchString(app) {
			return true
		}
	}

	return false
}

// detectVerdict runs the boundary's detector over the finished token — the
// threshold plumbing and the seam indirection in one place (the boundary
// stays under the complexity ceiling). A nil seam is the real detect.Check.
func (a *Actor) detectVerdict(tok []rune) detect.Verdict {
	check := a.detectCheck
	if check == nil {
		check = detect.Check
	}

	return check(tok, a.modeSymbol(),
		detect.Data{RU: layouts.DictRU, EN: layouts.DictEN},
		detect.Trigrams{RU: layouts.TriRU, EN: layouts.TriEN},
		detect.Params{
			MinWordLen:    a.opts.AutoCorrectMinWordLen,
			TrigramMargin: a.opts.AutoCorrectMargin,
			TrigramFloor:  a.opts.AutoCorrectFloor,
		})
}

// refreshACBlocklist rebuilds the compiled blocklist cache only when the
// raw pattern LIST changed (the macrLettersName parse-cache form) — one
// compile per document edit or SetOptions install, never per boundary.
// The caller holds the mutex.
func (a *Actor) refreshACBlocklist(names []string) {
	if slices.Equal(names, a.acBlocklistNames) {
		return
	}
	a.acBlocklistNames = names
	a.acBlocklist = compileACPatterns(names, "blocklist")
}

// refreshACTerminals rebuilds the compiled terminal-class cache only when
// the raw pattern LIST changed (the refreshACBlocklist form) — one compile
// per document edit or SetOptions install, never per boundary. The caller
// holds the mutex.
func (a *Actor) refreshACTerminals(names []string) {
	if slices.Equal(names, a.acTerminalNames) {
		return
	}
	a.acTerminalNames = names
	a.acTerminals = compileACPatterns(names, "terminal-class")
}

// compileACPatterns compiles one document pattern list for the boundary
// gates (the blocklist, the terminal class). A compile failure is
// impossible from a validated document (config.Validate refuses it at
// Load, D-33); the defensive branch WARNs once per rebuild and treats the
// whole list as EMPTY — nothing matches (the failure falls open toward
// the role gate, never around it).
func compileACPatterns(patterns []string, source string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			slog.Warn("autocorrect pattern rejected at fold", "source", source, "error", err)

			return nil
		}
		compiled = append(compiled, re)
	}

	return compiled
}

// setVerifyDead flips the behavioral terminal class of the live input
// context and records the transition with ONE source-neutral INFO marker —
// the A5 insurance (a flapping class must leave a trail), word-free
// (D-20/D-21): the marker names neither the typed word nor the source that
// produced the verdict. The caller holds the mutex.
func (a *Actor) setVerifyDead(dead bool) {
	if a.verifyDead == dead {
		return
	}
	a.verifyDead = dead
	slog.Info("terminal class", "verify_dead", dead)
}

// acFocusedApp resolves the focused app identity for the boundary and
// confirm gates. The 07-04 warn discipline (locked 07-CONTEXT): a WARN
// fires ONLY on a SOURCE ERROR — the source is installed but its answer
// broke (the one-per-episode warnAppid form). A MISSING identity — no
// source started, or no focus event yet — is the quiet norm under the
// blocklist semantics (nothing is visible to forbid, the unknown passes
// both gates) and never warns; a debug record is its only trace. The
// caller holds the mutex.
func (a *Actor) acFocusedApp() (string, bool) {
	if a.appid == nil {
		slog.Debug("autocorrect app identity unknown", "reason", "no source started")

		return "", false
	}
	app, err := a.appid.FocusedApp()
	if err != nil {
		a.warnACApp(err)

		return "", false
	}
	if app == "" {
		slog.Debug("autocorrect app identity unknown", "reason", "no focus event yet")

		return "", false
	}
	a.acAppWarned = false // a healthy answer closes the episode

	return app, true
}

// warnACApp records one autocorrect identity-source ERROR — the source is
// installed but its answer broke: the WARN fires once per degradation
// episode (the warnAppid form — per boundary would spam the journal of a
// broken source). A missing identity never reaches here: that is the
// quiet unknown of acFocusedApp (the 07-04 warn discipline). The caller
// holds the mutex.
func (a *Actor) warnACApp(err error) {
	if a.acAppWarned {
		return
	}
	a.acAppWarned = true
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

// autoConfirm resolves one armed autocorrect payload OFF the actor mutex
// (plan 06-06, the D-53 role contour): one LIVE GetRole round trip under
// the hard autoRoleTimeout deadline — the stored pair is the address of
// the query, never a cached answer (D-53/ADR-007) — then the verdict lands
// back under the mutex. The re-entry re-checks enabled (a reload may have
// switched the layer off between the boundary and this confirm — the
// stale-payload guard, T-06-06-06) and the generation (a newer boundary
// owns the decision); the fired branch additionally revalidates the armed
// payload against the live buffer (CR-01 — the generation advances only at
// boundaries, so a non-boundary mutation inside the role-RTT window must
// be caught by content) and consults the blocklist on the CURRENT focused
// identity (07-04: the confirm gate is blocklist-only, owner variant 1 —
// the ADR-007 amendment removed the arming/confirm equality; a focus
// switch inside the window surfaces as payload-stale via the buffer, or as
// a blocklist hit on the new app). Any
// unknown means SILENCE with its counted slug —
// the fail-closed direction INVERTED from macrTargetActive's degradation
// (no rung upward, ADR-007): a role error or the deadline is
// role-unknown/role-timeout with one WARN per episode, a non-text role
// (password 40, terminal 60, anything unlisted) is role-forbidden, and the
// ambiguous text role 61 with an UNKNOWN identity is role-ambiguous (the
// CR-01 tightening — a GTK3 password field reports 61 too, ADR-007's live
// fact). Only the full conjunction FIRES: the counter moves and the armed
// range launches THE one correction pipeline (startRangeCorrection — no
// second replacement mechanism, Pitfall 7). No branch ever logs the word
// (D-20/D-21).
func (a *Actor) autoConfirm(payload acPayload) {
	a.mu.Lock()
	rs := a.role
	a.mu.Unlock()
	if rs == nil {
		a.mu.Lock()
		a.warnACRole(nil)
		a.recordACAbstain(acReasonRoleUnknown)
		a.mu.Unlock()

		return
	}

	// The live call: outside the actor mutex, under its own deadline —
	// a stuck a11y object costs the budget, never a keystroke stall.
	ctx, cancel := context.WithTimeout(context.Background(), autoRoleTimeout)
	defer cancel()
	role, err := rs.Role(ctx)

	a.mu.Lock()
	defer a.mu.Unlock()
	if payload.gen != a.acGeneration {
		return // superseded mid-flight by a newer boundary — the stale verdict drops
	}
	if !a.opts.AutoCorrectEnabled {
		a.recordACAbstain(acReasonRecheckDisabled)

		return
	}
	if err != nil {
		reason := acReasonRoleUnknown
		if errors.Is(err, context.DeadlineExceeded) {
			reason = acReasonRoleTimeout
		}
		a.warnACRole(err)
		a.recordACAbstain(reason)

		return
	}
	a.acRoleWarned = false // a healthy answer closes the episode
	a.acRoleErrors = 0     // …and zeroes its error counter (plan 10-07): a recovered source reads as zero errors
	switch role {
	case acRoleText, acRoleEntry, acRoleDocumentText:
		// CR-01 + the 07-04 confirm form: the fired threshold re-checks the
		// armed payload's preconditions under the mutex — the window
		// between the boundary and this verdict must not move the field
		// out from under the payload, and the current app must not sit on
		// the blocklist (a focus move lands as payload-stale, or as a
		// blocklist hit on the new app). The role verdict rides along:
		// the ambiguous 61 answers only for a KNOWN identity (the CR-01
		// tightening — role-ambiguous otherwise).
		if !a.acConfirmRefusals(payload, role) {
			return
		}
		a.acFired++
		slog.Info("autocorrect", "reason", "fired")
		// The fired tone (plan 07-08): the distinct autocorrect event from
		// the fired point — the abstentions and the off state never reach
		// it (the single gate ahead of the sink). Fire-and-forget; the
		// actor adds no error handling around the call.
		if s := a.soundSinkFor(); s != nil {
			s.AutoCorrect()
		}
		a.startRangeCorrection(payload.rng)
	default:
		a.recordACAbstain(acReasonRoleForbidden)
	}
}

// acConfirmRefusals re-checks the armed payload's preconditions under the
// mutex at the confirm's fired threshold. CR-01: the buffer must still be
// what the arming event left — any mutation since (an interleaved
// keystroke, a Backspace, a second reset, a focus loss) means the armed
// range no longer describes the field, and executing it would delete the
// wrong runes and corrupt the mirror (T-06-06-06 extended past the
// generation guard). The identity check is BLOCKLIST-ONLY on the CURRENT
// identity (07-04, owner variant 1 — the ADR-007 amendment): a known
// focused app matching any compiled pattern refuses with app-blocked; an
// unknown identity passes — with ONE owner-sanctioned narrow exception
// (the CR-01 tightening): the ambiguous role 61 with an UNKNOWN identity
// refuses as role-ambiguous. Role 61 is the one value GTK3 password fields
// report exactly like text boxes (ADR-007's live fact), so with no
// identity visible the field cannot be told apart from a password and the
// fail-closed direction wins; the unambiguous allowed roles (79, 94) keep
// the locked "unknown passes" scenario, and the role-unknown/role-timeout
// disciplines are untouched (TestAutoConfirm_RoleTextAmbiguity pins the
// pair). The arming/confirm equality of WR-01 (the 06-REVIEW fix) is
// deliberately REMOVED with this revision: under the live role query it is
// redundant — the verdict is taken for the object the query answers NOW,
// and a focus move inside the window surfaces as payload-stale (FocusOut
// hard-resets the buffer; the replacement pipeline re-verifies the field
// before any Delete) — while kept conditionally it would break the locked
// scenario "unknown at arm, learned by confirm" with a phantom mismatch (a
// payload armed without an app can never equality-match a learned
// identity). Every refusal counts its closed slug and reports false; the
// caller holds the mutex.
func (a *Actor) acConfirmRefusals(payload acPayload, role uint32) bool {
	if !slices.Equal(a.buf.Token(), payload.expectToken) ||
		!slices.Equal(a.buf.Tail(), payload.expectTail) {
		a.recordACAbstain(acReasonPayloadStale)

		return false
	}
	// The identity resolves ONCE for both checks below — the same live
	// answer the boundary gate reads.
	app, known := a.acFocusedApp()
	// The CR-01 tightening: the ambiguous role with an UNKNOWN identity
	// cannot be told apart from a GTK3 password field — refuse fail-closed.
	// A KNOWN identity keeps the blocklist verdict (the owner's control).
	if role == acRoleText && !known {
		a.recordACAbstain(acReasonRoleAmbiguous)

		return false
	}
	// The 07-04 confirm form: the blocklist consults the CURRENT identity;
	// unknown passes, a matching known app forbids.
	if known {
		for _, re := range a.acBlocklist {
			if re.MatchString(app) {
				a.recordACAbstain(acReasonAppBlocked)

				return false
			}
		}
	}

	return true
}

// warnACRole records one live role-source failure (plan 10-07, G-10-1
// missing #3): the ERROR COUNT is loud — every role-request error of the
// degradation episode grows the status-visible counter (the 11:00:16
// episode stayed silent-broken behind its single WARN for the rest of the
// day; counts never hide again) — while the WARN LINE stays sticky: one
// line per episode, re-armed after acRoleWarnRearm of continued breakage
// (a lasting outage sounds again, a flap-bursting source never spams).
// A nil error (the no-source degradation, autoConfirm's rs==nil arm)
// warns by the same stickiness but counts nothing — no role request was
// made. The error is transport-class only — it never carries the word
// (D-20/D-21). The caller holds the mutex.
func (a *Actor) warnACRole(err error) {
	if err != nil {
		a.acRoleErrors++
	}
	now := a.roleClock()
	if a.acRoleWarned && now.Sub(a.acRoleWarnedAt) < acRoleWarnRearm {
		return // the episode's one line stands; the counter above carries the rest
	}
	a.acRoleWarned = true
	a.acRoleWarnedAt = now
	if err == nil {
		slog.Warn("autocorrect role unavailable")

		return
	}
	slog.Warn("autocorrect role unavailable", "error", err)
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
// the buffer holds); otherwise the WORD path of Phase 2 over the convertible
// stretch (SPEC §4.2, plan 10-02): the token plus the adjacent runes that
// carry a table position — «ytn?» corrects whole to «нет,», never «нет?» —
// with the replace half covering the whole run. The autocorrect payload
// keeps consuming the bare Token() (D-13/D-14 unchanged). The caller holds
// the mutex.
func (a *Actor) startCorrection() {
	// The two-phase level-1 ladder's flush site (plan 10-11): the next
	// gesture starts with a clean ladder — a superseded window must not
	// pour its stale replacement over the new round's field state. A nil
	// window is a no-op.
	a.flushLevel1Confirm()
	if sel, runes, ok := a.activeSelection(); ok {
		a.startSelectionCorrection(sel, runes)

		return
	}
	a.startRangeCorrection(correctionRange{
		token:   a.buf.ConvertibleStretch(),
		tail:    a.buf.StretchTail(),
		replace: a.buf.ReplaceStretch,
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
// run pipeline as the word and the phrase (the same ConvertRuns over the
// range), with the same refusal vocabulary and the same two-phase
// verification; the pre-check is range-anchored (VerifyRangeAt — the
// selection may sit on either side of the cursor, Pitfall 6), and the
// freshest spontaneous push is checked first, exactly as for the word. The
// caller holds the mutex.
func (a *Actor) startSelectionCorrection(sel selectionSpec, runes []rune) {
	armed := time.Now()
	eng := a.emitter() // one atomic snapshot for both emitter reads
	converted, changed, ok := correct.ConvertRuns(runes)
	if !ok {
		a.skipCorrection(refusalReason(runes))
		a.settleCombo()

		return
	}
	if !changed {
		// D-24: the success-without-changes outcome — a SUCCESSFUL
		// operation without changes; nothing to replace, nothing to verify.
		// The discriminator names it changed=false (plan 10-04).
		a.countCorrectionDone()
		slog.Info("correction", "outcome", "done", "changed", false)
		a.settleCombo()

		return
	}
	if eng == nil {
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
	a.armVerifyDeadline()
	eng.RequireSurroundingText()
}

// startPhraseCorrection launches the PHRASE correction of the Triple
// decision (CORR-02, D-25): the range is the whole buffer since the last
// hard reset — words and separators alike, no tail, no artificial length
// cap. The same pipeline runs over it (D-23). The caller holds the mutex.
func (a *Actor) startPhraseCorrection() {
	// The ladder flush of plan 10-11 extended to the phrase gesture: the
	// same «next gesture starts with a clean ladder» invariant — a Triple
	// decision landing inside a confirm window must not race the stale
	// replacement. A nil window is a no-op.
	a.flushLevel1Confirm()
	a.startRangeCorrection(correctionRange{
		token:   a.buf.Phrase(),
		replace: a.buf.ReplacePhrase,
	})
}

// startRangeCorrection is the ONE correction pipeline of the daemon,
// parameterized by its range: convert the range through ConvertRuns
// (homogeneous wholesale, mixed by per-character inversion — spec-delta
// 2026-10-06), choose the ladder level by the caps bit, then verify against
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
// a miss does the Require round-trip run inside the verify budget. A
// boundary-armed payload skips the cached shortcut entirely (CR-02): its
// arming keystroke is an in-flight field mutation, so the cache is stale by
// construction and the round trip is mandatory.
func (a *Actor) startRangeCorrection(rng correctionRange) {
	armed := time.Now()
	eng := a.emitter() // one atomic snapshot for both emitter reads
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
		// D-24: the success-without-changes outcome — a SUCCESSFUL
		// operation without changes (the owner's choice over a refusal and
		// over a WARN); nothing to replace, nothing to verify. The
		// discriminator names it changed=false (plan 10-04) — the same
		// grep-form, the attribute extends.
		a.countCorrectionDone()
		slog.Info("correction", "outcome", "done", "changed", false)
		a.settleCombo()

		return
	}
	if eng == nil {
		a.skipCorrection("no-engine")
		a.settleCombo()

		return
	}
	a.resolvePending() // a second Double/Triple supersedes the stale round
	a.clearAfter()     // …and the previous round's verify-after with it
	// TERM-02 routing (plan 10-04, owner verdict OQ 1 «unified»): the
	// terminal class takes its own write branch BEFORE the caps fork — no
	// pending fix, no RequireSurroundingText, no app verification (ADR-004
	// never answers in a terminal: the pushes do not come). The
	// classification is the SAME seam the arming veto consults — the veto
	// and the write pipeline cannot diverge by construction.
	terminal := a.classifyTerminal()
	if a.caps&correct.CapSurroundingText == 0 || terminal {
		// The write branch: the chunked level-2 replay. It serves the
		// legacy no-caps degradation (the ADR-003 ladder level, no
		// verification — the documented ADR-004 degradation, the buffer plus
		// the explicit CORR-09 reset triggers are the only synchronization)
		// AND the whole terminal class — against a terminal the buffer is
		// the only target source (the stretch of plan 10-02), the pending
		// machine is never armed. Only the terminal class verifies after
		// the fact (Task 3); the legacy path arms nothing, byte as today.
		a.executeLevel2(rng, converted, armed, terminal)

		return
	}
	a.pending = &pendingFix{
		rng:       rng,
		converted: converted,
		match:     concatRunes(rng.token, rng.tail),
		armed:     armed,
		cacheNil:  a.surr == nil, // the stale-cache evidence of the 2026-10-13 amendment
	}
	// CR-02: a boundary-armed payload NEVER trusts the cached push — the
	// arming keystroke (Enter/Tab/Escape) is itself an in-flight
	// field-mutating or focus-moving key, so the cache is the PRE-reset
	// state by construction, and a delete+commit on it can land in a
	// moved-on or entirely different input context. Always require the
	// fresh answer: the pre-boundary early-answer machinery (cbfa46c)
	// waits out the in-flight separator, and a post-reset answer names the
	// field as the reset key left it — a mismatch refuses fail-closed
	// (ADR-004). For the space boundary the round is one extra
	// sub-millisecond trip; for the reset boundaries it is the only guard.
	if rng.acBoundary || !correct.MatchesSuffix(a.surr, a.pending.match) {
		a.armVerifyDeadline()
		eng.RequireSurroundingText()

		return
	}
	a.executeCorrection() // cached push is the freshest report
}

// executeCorrection runs the level-1 ladder plan of the settled pending fix
// — one DeleteSurroundingText exactly over the token+tail range (CORR-07),
// gated by the DST guard (plan 10-16: only over the generation-scoped
// materialization evidence, else the forward branch) — and arms the
// two-phase confirmation window (plan 10-11, round-2 G-10-1,
// the 09:39 host rows): the commit phase does NOT follow the deletion on
// the wire. The host client's processing reorders the tight pair (the A2
// class: the insertion lands first and is visible, the late deletion erases
// the fresh word — «вижу привет, и он удаляется»), so the replacement waits
// for the client's proof the deletion applied; flushLevel1Confirm is the
// ONE completion point, reached by the deletion proof, the budget fallback
// (today's pair, loudly marked) or a flush site. The emitter snapshot is
// captured here and rides the whole window (:3444 precedent). A selection
// fix runs its own geometry instead (executeSelectionCorrection). The
// caller holds the mutex and pending != nil.
func (a *Actor) executeCorrection() {
	p := a.pending
	a.pending = nil
	if p.stop != nil {
		p.stop()
	}
	if p.sel != nil {
		a.executeSelectionCorrection(p)

		return
	}
	eng := a.emitter() // one atomic snapshot for the whole ladder emission
	if !a.surrSeen || a.surrEng != eng {
		// The DST guard (plan 10-16, the 2026-10-16 bisection verdict): the
		// surrounding state of THIS generation was never materialized in the
		// composer — no spontaneous push, no materializing answer — and on
		// mutter 46.2 a DeleteSurroundingText toward such a context segfaults
		// gnome-shell (strlen over the virgin state,
		// meta-wayland-text-input.c:254; ANY parameters, the ~1.5 s virgin
		// window — a guarantee a layout corrector cannot honestly make, so
		// the machine makes it structurally). The emission is withheld and
		// the replacement rides the PROVEN forward branch — the existing
		// executeLevel2 form (chunks ≤ terminalChunkKeys, level2ForwardGap
		// pacing, the ctxEpoch re-check, one commit; the WR-01 busy guard
		// covers the window) with no new transport code (verdict (c)). This
		// round was classified NON-terminal to reach the level-1 ladder, so
		// the branch runs in its non-terminal form. The boundary is loud and
		// content-free (D-20/D-21) and the diversion is counted for the
		// status surface (T-10-16-04). RequireSurroundingText is never
		// suppressed and the two-phase machine of a materialized context is
		// byte-untouched (TestDstGuard_MaterializedContextByteUnchanged).
		a.corrDstGuarded++
		slog.Info("correction", "outcome", "dst-guarded",
			"latency_ms", time.Since(p.armed).Milliseconds())
		a.executeLevel2(p.rng, p.converted, p.armed, false)

		return
	}
	plan := correct.BuildPlan(p.rng.token, p.rng.tail, p.converted, a.caps, a.backspaceCap())
	eng.DeleteSurroundingText(plan.Offset, plan.NChars)
	a.armLevel1Confirm(&level1Confirm{
		eng:       eng,
		commit:    plan.Commit,
		rng:       p.rng,
		converted: p.converted,
		match:     concatRunes(p.rng.token, p.rng.tail),
		armed:     p.armed,
		level:     plan.Level,
	})
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
	a.emitter().CommitText(engine.NewIBusText(string(p.converted)))
	a.buf.HardReset() // the buffer can no longer mirror the replaced field
	a.logCorrectionDone(correct.Level1, p.rng.token, p.converted, time.Since(p.armed))
	a.settleCorrectionFlip(p.converted) // owner rule 2026-09-28: the mode follows the converted script
	a.settleCombo()                     // D-36: the flip lands strictly after the settled completion record
	a.armAfterVerifyRange(p.converted, sel.start)
}

// executeLevel2 runs the ladder's Backspace level: the plan's Backspaces —
// counted in runes by BuildPlan, the tail included — replayed as
// press+release PAIRS of ForwardKeyEvent(BackSpace) (round-6: a press-only
// forward arms the client's own key repeat) in chunks of at most
// terminalChunkKeys PAIRS with context re-checks, followed by ONE commit of
// the converted token plus the tail. The burst→commit order is the wire contract (research A4: the
// daemon preserves it for every client). The write branch serves two
// callers: the legacy no-caps degradation (ADR-003) and — owner verdict
// OQ 1 «unified» — the whole terminal class, which may ANNOUNCE the
// surrounding-text bit (VTE's 0x29 lie, the wezterm bit-set signature
// with no pushes ever coming): the level-2 geometry is therefore forced by
// masking the bit, and BuildPlan itself stays untouched (the D-13 tail
// semantics ride verbatim). The terminal argument marks the class-routed
// call: only it gets the round-4 delivery machine (plan 10-17 — the
// drain-settle bounded window before the commit, the narrowed flip
// release sites, the INFO delivery marks) and arms the verify-after round
// (armAfterVerifyTerminal — no paste payload, plan 10-04 Task 3); the
// legacy degradation keeps today's byte-exact shape, one pacing gap
// before the commit, no marks, no deferred flip. The context discipline
// is plan 10-17's: BEFORE the first destructive key a generation change
// is a free stop (nothing destroyed — nothing to replace); AFTER it the
// commit is NOT cancellable («удаление без замены» is forbidden — the
// ADR-008 amendment (6)) — the replacement rides the captured emitter
// snapshot behind the loud mid-flow mark. The caller holds the mutex.
func (a *Actor) executeLevel2(rng correctionRange, converted []rune, armed time.Time, terminal bool) {
	// The WR-01 re-entrancy guard (plan 10-10): checked-and-set BEFORE the
	// plan build and the first forward — a second entry while a burst is in
	// flight (the chunk-boundary window and the terminal branch's unlocked
	// delivery wait release the mutex) refuses here, honestly and
	// non-destructively. The guard covers the WHOLE machine window
	// (T-10-17-01): chunks, the drain-settle wait, every exit.
	if a.level2Busy {
		a.skipCorrection("correction-busy")
		a.settleCombo()

		return
	}
	a.machineArm()     // the WR-01 guard arms the key-hold's close-on-exit channel with it (CR-02)
	eng := a.emitter() // one atomic snapshot for the whole burst and its commit
	plan := correct.BuildPlan(rng.token, rng.tail, converted, a.caps&^correct.CapSurroundingText, a.backspaceCap())
	if plan.Level == correct.LevelNone {
		// D-27: the Backspace series would exceed the cap and this client
		// has no DeleteSurroundingText — refuse silently, not one deletion.
		a.machineSettle() // the guard clears on EVERY exit (WR-01); the held keys ride free
		a.skipCorrection("backspace-cap")
		a.settleCombo()

		return
	}
	gen := a.ctxEpoch
	sleep := a.level2Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	if terminal {
		// The burst identity (the round-5 instrument): the open row births
		// the burst — its emission-seq rides every following delivery row,
		// its close pairs on EVERY exit (the free abort closes with
		// forwards=0, the commit path with the emitted count). The frozen
		// grep-forms keep matching; the attributes extend, never rename.
		// Round-6: every BackSpace rides a press+release pair, so the
		// emitted FORWARD count is 2×keys — the rows carry both ("keys" is
		// the rune-true BackSpace count, "forwards" the wire emissions).
		a.emissionSeq++
		a.deliveryStart = time.Now()
		a.deliveryMark("burst-window", "phase", "open",
			"forwards", forwardEmissionsPerKey*plan.Backspaces, "keys", plan.Backspaces)
	}
	if a.replayBurst(eng, plan.Backspaces, gen, sleep) {
		if terminal {
			a.deliveryMark("burst-window", "phase", "close", "forwards", 0, "outcome", "aborted")
		}

		return // the free stop BEFORE the first forward — the round is settled
	}
	if terminal {
		// The drain-settle bounded window (plan 10-17): the replacement
		// commits only after the window that gives the client's forward
		// queue an honest term to drain (the constant's journal provenance
		// above). The WR-01 guard covers the wait (T-10-17-01); a
		// generation change inside it is loud, never an abort.
		a.drainSettleWindow(gen, sleep)
		// The delivery marks (plan 10-17): the host's ONLY observability of
		// the machine — outcomes, counters and latencies, never field
		// content (D-20/D-21). The frozen grep-forms
		// («msg correction, outcome done») keep matching; the attributes
		// extend, never rename.
		a.deliveryMark("chunks-sent",
			"forwards", forwardEmissionsPerKey*plan.Backspaces, "keys", plan.Backspaces,
			"pacing_ms", (forwardEmissionsPerKey * time.Duration(plan.Backspaces) * a.level2ForwardGap).Milliseconds())
	} else {
		sleep(a.level2ForwardGap) // the legacy branch, byte as today
	}
	eng.CommitText(engine.NewIBusText(string(plan.Commit)))
	if terminal {
		a.deliveryMark("commit-sent", "latency_ms", time.Since(armed).Milliseconds())
		a.deliveryMark("burst-window", "phase", "close",
			"forwards", forwardEmissionsPerKey*plan.Backspaces, "keys", plan.Backspaces, "outcome", "committed")
	}
	rng.replace(converted)
	a.logCorrectionDone(plan.Level, rng.token, converted, time.Since(armed))
	a.machineSettle() // the guard clears on the success exit (WR-01); the held keys ride free
	if terminal {
		// E3 delivery-order safety (plan 10-08, G-10-1 cause B), narrowed by
		// plan 10-17: the SetGlobalEngine of the post-correction flip may
		// not follow the commit — the flip ARMS and fires ONLY on the next
		// keystroke or a confirming push; the grace deadline and the verify
		// outcomes release nothing (the release sites inside the delivery
		// window are exactly the round-2 loss geometry).
		a.deferCorrectionFlip(converted, true)
	} else {
		a.settleCorrectionFlip(converted) // owner rule 2026-09-28: the mode follows the converted script
	}
	a.settleCombo()
	if terminal {
		// The terminal branch verifies after the fact (plan 10-04 Task 3):
		// the client that does push settles the round honestly, a push-less
		// one dies by budget — and a mismatch can only log and stop, the
		// round carries no paste (armAfterVerifyTerminal).
		a.armAfterVerifyTerminal(converted, rng.tail)
	}
}

// chunkBoundaryMoved yields the actor mutex for one scheduling quanta and
// reports whether the input context died in the window — the ctxEpoch tag
// bumped by HandleLifecycle's FocusOut/Reset branch (the verifyEpoch
// discipline: a monotonic tag compared under the mutex). A second boundary
// always catches a callback that missed the first window. The caller holds
// the mutex on both entry and return. (The abort semantics around this
// window are plan 10-17's: a free stop only BEFORE the first forward;
// after it the round continues to its commit, loudly marked.)
func (a *Actor) chunkBoundaryMoved(gen uint64) bool {
	a.mu.Unlock()
	runtime.Gosched()
	a.mu.Lock()

	return a.ctxEpoch != gen
}

// freeAbortBeforeFirstForward is the round's LAST free stop (plan 10-17):
// a generation change observed before the first destructive key ends the
// round without a commit, without a completion and without a repair — the
// context that must receive the replacement no longer exists and nothing
// has been destroyed yet. The chunk-boundary window releases the mutex so
// the lifecycle callback can land. The caller holds the mutex on both
// entry and return; a true answer means the round is ALREADY settled (the
// caller just returns).
func (a *Actor) freeAbortBeforeFirstForward(gen uint64) bool {
	if !a.chunkBoundaryMoved(gen) {
		return false
	}
	slog.Info("correction", "outcome", "context-changed")
	a.machineSettle() // the guard clears on EVERY exit (WR-01); the held keys ride free
	a.settleCombo()

	return true
}

// replayBurst emits the plan's BackSpace forwards — chunked at
// terminalChunkKeys (PAIRS per chunk, round-6) with the context re-checks
// between chunks (plan 10-04) — and reports whether the round took the free
// stop BEFORE its first forward (plan 10-17). Every BackSpace rides a
// press+release PAIR on the pacing seam (round-6, the ADR-008 amendment
// (7)): a press-only forward ARMS the client's own key repeat — wezterm
// e2a98e1b sets KeyRepeatState on every repeatable-key press
// (window.rs:1194-1206) and disarms it ONLY on the same key's release,
// another key's press or a keyboard focus Enter/Leave; with no release the
// armed repeat fires after the GNOME repeat-delay (default 500 ms) at ~30 Hz
// and deletes the committed word — the host defect «появилось и исчезло»
// (.planning/debug/2026-10-19-wezterm-repeat-sourcedig.md). The pair shape
// is ProbeForward's discipline (probe.go:179-183): press (state 0), one
// pacing gap, release (state MaskRelease) — the release lands one gap tick
// (20 ms default) after its press, far inside the 500 ms arming window, and
// the inter-PAIR pacing keeps the level2ForwardGap rhythm. A generation
// change after the first destructive key is loud but no longer fatal (the
// chunk-boundary mark): the round can only go forward to its commit on the
// captured snapshot — the old criterion-4 stop (no commit after forwards)
// is replaced by the ADR-008 amendment (6): a deletion without its
// replacement is the one outcome the machine may never produce. The caller
// holds the mutex.
func (a *Actor) replayBurst(eng engine.Emitter, backspaces int, gen uint64, sleep func(time.Duration)) bool {
	emitted := 0
	for sent := 0; sent < backspaces; {
		chunk := min(terminalChunkKeys, backspaces-sent)
		for range chunk {
			sleep(a.level2ForwardGap)
			// The free-abort window (plan 10-17): BEFORE the first
			// destructive key a generation change still stops the round for
			// free — nothing has been destroyed, so nothing needs replacing.
			if emitted == 0 && a.freeAbortBeforeFirstForward(gen) {
				return true
			}
			eng.ForwardKeyEvent(engine.KeyBackSpace, backSpaceKeycode, 0)
			sleep(a.level2ForwardGap)
			eng.ForwardKeyEvent(engine.KeyBackSpace, backSpaceKeycode, engine.MaskRelease)
			emitted++
		}
		sent += chunk
		if sent >= backspaces {
			break
		}
		// The chunk-boundary re-check: a generation change here is LOUD but
		// no longer fatal — see replayBurst's contract above.
		if a.chunkBoundaryMoved(gen) {
			a.deliveryMark("context-changed-midflow")
		}
	}

	return false
}

// drainSettleWindow holds the terminal branch's bounded delivery wait
// (plan 10-17, T-10-17-05): the actor mutex is RELEASED for the window —
// a second correction entry lands here and is refused by the WR-01 guard
// (T-10-17-01, level2Busy covers the whole window), a real lifecycle
// callback lands freely, and the epoch re-check marks it loud WITHOUT
// aborting: after the first forward the commit is not cancellable (the
// ADR-008 amendment (6)). The settle constant bounds the wait — the
// window can never hold the machine hostage. The caller holds the mutex
// on both entry and return; sleep is the branch's pacing seam.
func (a *Actor) drainSettleWindow(gen uint64, sleep func(time.Duration)) {
	// WR-06 (the 10-REVIEW): the settle value is read UNDER the mutex — the
	// field's documented discipline («Read under a.mu by the machine»);
	// reading it after the Unlock was a live data-race against the setter
	// for any future runtime caller (WR-01's race class).
	settle := a.terminalDrainSettle
	a.mu.Unlock()
	sleep(settle)
	a.mu.Lock()
	if a.ctxEpoch != gen {
		a.deliveryMark("context-changed-midflow")
	}
}

// machineArm raises the WR-01 re-entrancy guard and arms the key-hold's
// close-on-exit channel with it (CR-02): one raise, one channel, both under
// the mutex the caller holds. machineSettle is its mirror — every machine
// exit (the backspace-cap refusal, the free abort, the success return)
// closes the channel so the keystrokes parked inside the unlocked window
// ride free the moment the machine settles. The caller holds the mutex.
func (a *Actor) machineArm() {
	a.level2Busy = true
	a.machineDone = make(chan struct{})
}

func (a *Actor) machineSettle() {
	a.level2Busy = false
	if a.machineDone != nil {
		close(a.machineDone)
		a.machineDone = nil
	}
}

// deliveryMark logs one content-free «correction delivery» row of the
// in-flight burst (the round-5 instrument): the stage and the row's own
// attributes go FIRST — the frozen grep-forms keep their adjacency (the
// e2e corpus pins «stage flip-released, reason keystroke» byte-exact) —
// and the burst identity closes the row: emission-seq, the burst's
// monotonic counter, and t_rel_ms, the offset from the burst's timeline
// anchor. Outcomes and counters only, never field content (D-20/D-21).
// The caller holds the mutex.
func (a *Actor) deliveryMark(stage string, extra ...any) {
	a.deliveryMarkBurst(a.emissionSeq, a.deliveryStart, stage, extra...)
}

// deliveryMarkBurst is deliveryMark with an explicit burst identity: the
// deferred-flip release sites stamp the burst the flip armed in — a newer
// burst may already have replaced the actor fields by the release instant.
// The caller holds the mutex.
func (a *Actor) deliveryMarkBurst(seq uint64, start time.Time, stage string, extra ...any) {
	rel := int64(0)
	if !start.IsZero() {
		rel = time.Since(start).Milliseconds()
	}
	args := make([]any, 0, len(extra)+deliveryMarkFixedAttrs)
	args = append(args, "stage", stage)
	args = append(args, extra...)
	args = append(args, "emission-seq", seq, "t_rel_ms", rel)
	slog.Info("correction delivery", args...)
}

// deliveryMarkFixedAttrs is the capacity headroom of the burst rows for
// the instrument's own key/value pairs (stage, emission-seq, t_rel_ms).
const deliveryMarkFixedAttrs = 6

// holdableKey reports whether the event is TYPED-LAYOUT input the hold path
// must park while the delivery machine is in flight (CR-02): printable
// characters (the runes the buffer tracks and the field receives — the RU
// branch's own daemon-side commit included) and BackSpace (the buffer's
// erase — a mid-window pop would also corrupt the pending round's replace
// geometry). Gestures, bare modifiers and the reset keys never park: their
// semantics (the busy refusal, the chord, the FSM, the CORR-09 boundary)
// belong to the arrival instant, and a parked combo would re-enter the
// correction pipeline post-settle over the already-replaced buffer.
func holdableKey(ev engine.EngineEvent) bool {
	if ev.Keyval == engine.KeyBackSpace {
		return true
	}

	return printableKeyval(ev.Keyval) && ev.Mods&comboMask == 0
}

// holdDuringDelivery parks ev while the delivery machine is in flight and
// drains the parked queue once it settles (CR-02). The arrival verdict is
// always TRUE — consumed: ibus's own forwarding never happens, the daemon
// owes the wire the event after the machine's commit (reemitTransit). The
// first parked event's goroutine drains the whole FIFO (holdDrain is the
// single-flight flag); later arrivals append and return — the drainer
// processes them in arrival order. Every wait is bounded
// (awaitDeliverySettle): past the bound the hold degrades loudly to arrival
// processing — keys are never lost, the owner's typing is sacred. The
// caller holds the mutex on entry and on return.
func (a *Actor) holdDuringDelivery(ev engine.EngineEvent) bool {
	a.holdQueue = append(a.holdQueue, ev)
	if a.holdDrain {
		return true // a drainer already owns the parked queue
	}
	a.holdDrain = true
	defer func() { a.holdDrain = false }()
	for len(a.holdQueue) > 0 {
		if a.level2Busy && !a.awaitDeliverySettle() {
			slog.Warn("key hold bound expired", "parked", len(a.holdQueue))
		}
		ev := a.holdQueue[0]
		a.holdQueue = a.holdQueue[1:]
		consume := a.handleKeyArrived(ev)
		a.reemitTransit(ev, consume)
	}

	return true
}

// awaitDeliverySettle waits (bounded) for the in-flight delivery machine to
// settle: the mutex is RELEASED for the wait — the machine needs it to
// finish (its pacing sleeps hold it) — and re-taken on return. The bound
// covers the machine's whole worst case (the drain window plus a full-cap
// burst's pacing — round-6: every key rides TWO pacing quanta, press and
// release — plus one slack quantum; a keystroke may catch the first
// chunk-boundary Gosched window of a maximum burst), so an expiry is a
// pathological overrun, never routine. Returns whether the machine settled.
// The caller holds the mutex on entry and on return.
func (a *Actor) awaitDeliverySettle() bool {
	done := a.machineDone
	bound := a.terminalDrainSettle +
		forwardEmissionsPerKey*time.Duration(a.backspaceCap())*a.level2ForwardGap + keyHoldSlack
	a.mu.Unlock()
	timer := time.NewTimer(bound)
	select {
	case <-done:
	case <-timer.C:
	}
	timer.Stop()
	a.mu.Lock()

	return !a.level2Busy
}

// handleKeyArrived is HandleKey's arrival body minus the hold gate — the
// form the drainer runs for every parked event post-settle (the exact
// semantics the key would have received at arrival on an idle machine).
// The caller holds the mutex.
func (a *Actor) handleKeyArrived(ev engine.EngineEvent) (consume bool) {
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

// reemitTransit puts a held key's transit back on the wire AFTER the
// machine's commit (CR-02): HandleKey consumed the event at arrival, so
// ibus's own forwarding never happened — the daemon owes the client the
// key. A consumed verdict needs nothing (the RU branch already committed
// the rune post-settle; gestures emit nothing); a nil emitter fails closed
// (no engine — no wire to owe). The release bit rides the wire state word,
// the modsMask decoding convention. The caller holds the mutex.
func (a *Actor) reemitTransit(ev engine.EngineEvent, consume bool) {
	if consume {
		return
	}
	eng := a.emitter()
	if eng == nil {
		return
	}
	mods := ev.Mods
	if ev.Release {
		mods |= engine.MaskRelease
	}
	eng.ForwardKeyEvent(ev.Keyval, ev.Keycode, mods)
}

// logCorrectionDone counts and writes the completion pair of one ladder
// execution: the INFO counter (D-20 — outcome only, never the word) with
// the D-24 discriminator (plan 10-04: an executed correction is
// changed=true — the frozen stand grep-forms «msg correction, outcome done»
// keep matching, the attribute extends, never renames) and the DEBUG detail
// record whose FIRST attribute after msg is the ladder level — the exact
// form the e2e matrix greps to pin the ACTUAL level a client got (D-21).
// The caller holds the mutex.
func (a *Actor) logCorrectionDone(level correct.Level, token, converted []rune, latency time.Duration) {
	a.corrDone++
	slog.Info("correction", "outcome", "done", "changed", true)
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
	if a.pending.stop != nil {
		a.pending.stop()
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
// without Latin or Cyrillic letters has nothing to convert; anything else
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
