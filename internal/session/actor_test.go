package session_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Djarvur/goswitch/internal/clipboard"
	"github.com/Djarvur/goswitch/internal/config"
	"github.com/Djarvur/goswitch/internal/ctlsvc"
	"github.com/Djarvur/goswitch/internal/detect"
	"github.com/Djarvur/goswitch/internal/engine"
	"github.com/Djarvur/goswitch/internal/hotkey"
	"github.com/Djarvur/goswitch/internal/session"
)

// farWindow is large enough that the actor's real AfterFunc can never fire
// during a test: expiry is injected deterministically via ExpiryAt instead.
const farWindow = time.Hour

// expiryAfterWindow is a logical timestamp safely past the last tap's
// deadline: the actor's clock reads microseconds at most during a test, so
// taps land well inside the window and this value is well past it.
const expiryAfterWindow = 2 * farWindow

// syncBuffer is a mutex-guarded log sink: the actor's timer callback logs
// from its own goroutine while the test reads.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

// Write appends under the guard. bytes.Buffer.Write is documented to
// always return a nil error, so there is nothing to propagate.
func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	n, _ := b.buf.Write(p)

	return n, nil
}

// String snapshots the buffer under the guard.
func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

// captureLogs redirects the process default logger into a guarded buffer.
func captureLogs(t *testing.T) *syncBuffer {
	t.Helper()
	buf := &syncBuffer{}
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, nil)))

	return buf
}

// tapShift feeds one clean Shift_R press/release pair.
func tapShift(a *session.Actor) {
	a.HandleKey(engine.EngineEvent{Keyval: hotkey.KeyvalShiftR})
	a.HandleKey(engine.EngineEvent{Keyval: hotkey.KeyvalShiftR, Release: true})
}

// tapKeySeries feeds one clean press/release pair of an arbitrary key — the
// tap-key reload corpus drives keys other than the corpus Shift_R.
func tapKeySeries(a *session.Actor, keyval uint32) {
	a.HandleKey(engine.EngineEvent{Keyval: keyval})
	a.HandleKey(engine.EngineEvent{Keyval: keyval, Release: true})
}

// countActions counts decision records in the captured log.
func countActions(buf *syncBuffer) int {
	return strings.Count(buf.String(), `"msg":"action"`)
}

// Corpus words of the correction pipeline (the 02-01 corpus pair — named
// once, goconst). mismatchLine is the non-matching surrounding push of the
// refusal corpus — it ends with neither layout's token. mixedRU is the
// per-character inversion of the mixed token "gfb"+wordRU — the re-pinned
// matrix-v1 word-mixed expectation (2026-10-06 owner verdict).
const (
	wordEN       = "ghbdtn"
	wordRU       = "привет"
	mixedRU      = "паиghbdtn"
	mismatchLine = "abc другойтекст"
)

// deleteCall is one recorded DeleteSurroundingText emission.
type deleteCall struct {
	offset int32
	nchars uint32
}

// forwardCall is one recorded ForwardKeyEvent emission with its exact
// arguments (the level-2 Backspace burst, plan 02-05).
type forwardCall struct {
	keyval  uint32
	keycode uint32
	state   uint32
}

// fakeSink is the test double of engine.Emitter: every emitter call is
// recorded under a mutex — the observable surface the correction pipeline
// drives (the guarded-sink style of syncBuffer) — and the unified op log
// pins the ORDER of the calls (the burst→commit sequence of the ladder).
// forwardHook, when set, fires inside ForwardKeyEvent so a test can
// interleave the sink's records with another guarded recorder (the
// clipboard rung's subprocess log); modeHook plays the same role for
// UpdateModeSymbol (the panel-symbol corpus pins the emit-after-log order).
type fakeSink struct {
	mu          sync.Mutex
	requires    int
	deletes     []deleteCall
	commits     []string
	forwards    []forwardCall
	modes       []string
	ops         []string
	forwardHook func(forwardCall)
	commitHook  func(text string)
	modeHook    func(symbol string)
}

// RequireSurroundingText records the verification request.
func (f *fakeSink) RequireSurroundingText() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.requires++
	f.ops = append(f.ops, "require")
}

// DeleteSurroundingText records the ladder deletion with its exact range.
func (f *fakeSink) DeleteSurroundingText(offset int32, nchars uint32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.deletes = append(f.deletes, deleteCall{offset: offset, nchars: nchars})
	f.ops = append(f.ops, fmt.Sprintf("delete(%d,%d)", offset, nchars))
}

// CommitText records the committed text payload.
func (f *fakeSink) CommitText(text engine.IBusText) {
	f.mu.Lock()
	f.commits = append(f.commits, text.Text)
	f.ops = append(f.ops, "commit("+text.Text+")")
	hook := f.commitHook
	f.mu.Unlock()

	if hook != nil {
		hook(text.Text)
	}
}

// ForwardKeyEvent records the replayed key event with its exact arguments.
func (f *fakeSink) ForwardKeyEvent(keyval, keycode, state uint32) {
	f.mu.Lock()
	call := forwardCall{keyval: keyval, keycode: keycode, state: state}
	f.forwards = append(f.forwards, call)
	f.ops = append(f.ops, fmt.Sprintf("forward(%d,%d,%d)", keyval, keycode, state))
	hook := f.forwardHook
	f.mu.Unlock()

	if hook != nil {
		hook(call)
	}
}

// UpdateModeSymbol records the panel-symbol update with its argument — the
// mode-indicator emit of owner decision 1 (quick plan 260927-way).
func (f *fakeSink) UpdateModeSymbol(symbol string) {
	f.mu.Lock()
	f.modes = append(f.modes, symbol)
	f.ops = append(f.ops, "mode("+symbol+")")
	hook := f.modeHook
	f.mu.Unlock()

	if hook != nil {
		hook(symbol)
	}
}

// modeSymbols snapshots the recorded panel-symbol updates.
func (f *fakeSink) modeSymbols() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]string(nil), f.modes...)
}

// requireCount snapshots the verification-request count.
func (f *fakeSink) requireCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.requires
}

// deleteCalls snapshots the recorded deletions.
func (f *fakeSink) deleteCalls() []deleteCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]deleteCall(nil), f.deletes...)
}

// commitTexts snapshots the recorded commits.
func (f *fakeSink) commitTexts() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]string(nil), f.commits...)
}

// forwardCalls snapshots the recorded key-event replays.
func (f *fakeSink) forwardCalls() []forwardCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]forwardCall(nil), f.forwards...)
}

// opLog snapshots the unified emitter-call sequence in arrival order.
func (f *fakeSink) opLog() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]string(nil), f.ops...)
}

// wiredActor returns an actor with the fake sink attached and the
// surrounding-text capability announced — the minimal live pipeline state.
func wiredActor() (*session.Actor, *fakeSink) {
	return wiredActorCaps(engine.CapSurroundingText)
}

// wiredActorCaps is wiredActor with an explicit capability bitmap — the
// level-2 corpus announces a client WITHOUT the surrounding-text bit.
func wiredActorCaps(caps uint32) (*session.Actor, *fakeSink) {
	a := session.NewActor(farWindow)
	sink := &fakeSink{}
	a.AttachEngine(sink)
	a.HandleCapabilities(caps)

	return a, sink
}

// typeWord feeds printable key presses the way the engine delivers them:
// one press/release pair per rune, keyval = the rune itself.
func typeWord(a *session.Actor, word string) {
	for _, r := range word {
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r)})
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r), Release: true})
	}
}

// captureLogsLevel redirects the default logger at the given level: the
// correction corpus asserts both the INFO contract (D-20 — reasons without
// word contents) and the DEBUG record shape (D-21 — level first).
func captureLogsLevel(t *testing.T, level slog.Level) *syncBuffer {
	t.Helper()
	buf := &syncBuffer{}
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: level})))

	return buf
}

// TestActor_DoubleTapCorrects pins the tracer pipeline end to end (CORR-01,
// CORR-07 level 1): printable presses feed the buffer, a double tap at
// window expiry starts the two-phase correction — RequireSurroundingText on
// the sink, nothing destructive yet — and the matching SetSurroundingText
// resolves it: exactly one DeleteSurroundingText(-6,6) (the token, no tail)
// and one CommitText("привет"), plus the INFO completion record.
func TestActor_DoubleTapCorrects(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	typeWord(a, wordEN)
	honestPush(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != 1 {
		t.Fatalf("RequireSurroundingText calls after double tap = %d, want 1", got)
	}
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Fatalf("two-phase violation: %d deletions before surrounding text, want 0", len(calls))
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Fatalf("two-phase violation: %d commits before surrounding text, want 0", len(texts))
	}

	// The cursor at the end of the line sees the whole text.
	a.HandleSurroundingText("abc "+wordEN, runeLen("abc "+wordEN), runeLen("abc "+wordEN))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -6, nchars: 6}) {
		t.Fatalf("deletions = %+v, want exactly one (-6,6) — the token range, no tail", calls)
	}
	// The two-phase completion (plan 10-11): the client reports the
	// deletion applied (the field is back to «abc ») — only then does the
	// replacement commit.
	a.HandleSurroundingText("abc ", 4, 4)
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != wordRU {
		t.Fatalf("commits = %q, want exactly one %q", texts, wordRU)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done"`) {
		t.Errorf("INFO completion record missing; log:\n%s", buf.String())
	}
}

// runeLen is the cursor position at the end of s — the surrounding-text
// argument shape every test uses (self-computed, never hand-counted).
func runeLen(s string) uint32 {
	return uint32(len([]rune(s)))
}

// honestPush is the migration idiom of plan 10-10: one spontaneous client
// push before the gesture — the push channel is ALIVE, the round arms over
// cache evidence (cacheNil=false). The nil-cache form of a matched round is
// the answer-only terminal signature now (TestReplayHost0939_*); every pin
// of the byte-as-today level-1 world drives the honest-GTK world.
func honestPush(a *session.Actor) {
	a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))
}

// TestActor_CachedSurroundingCorrects pins the live transport behavior of
// GTK/mutter clients (live finding 2026-09-14): they never answer
// RequireSurroundingText — they push SetSurroundingText spontaneously after
// every keystroke. The correction must verify against that cached push and
// execute at the Double decision itself; the only Require the actor makes
// afterwards is the verify-after round of plan 02-05 (exactly one), never a
// pre-correction round.
func TestActor_CachedSurroundingCorrects(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	for i, r := range wordEN {
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r)})
		a.HandleSurroundingText(wordEN[:i+1], uint32(i+1), uint32(i+1)) // the client's spontaneous push
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r), Release: true})
	}
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != 0 {
		t.Errorf("require calls before the completion = %d, want 0 — the verify-after arms at the "+
			"flush (plan 10-11), never a pre-correction one", got)
	}
	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -6, nchars: 6}) {
		t.Fatalf("deletions = %+v, want exactly one (-6,6)", calls)
	}
	// The two-phase completion (plan 10-11): the client reports the
	// deletion applied (the field is empty) — the replacement commits and
	// the verify-after Require rides the flush.
	a.HandleSurroundingText("", 0, 0)
	if got := sink.requireCount(); got != 1 {
		t.Errorf("require calls after the completion = %d, want exactly 1 — the verify-after round", got)
	}
	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != wordRU {
		t.Fatalf("commits = %q, want one %q", texts, wordRU)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done"`) {
		t.Errorf("INFO completion record missing; log:\n%s", buf.String())
	}
}

// TestActor_LatchedModsStillFeed pins the live finding of the tracer run
// (2026-09-14): on a NumLock-lit desktop every letter arrives with the
// NumLock latch (mods 0x10) in the IBus state word — latch-state modifiers
// must not starve the buffer, the correction has to run exactly as with
// bare keys (Ctrl/Alt/Super combos still never feed it).
func TestActor_LatchedModsStillFeed(t *testing.T) {
	const numLockLatch = 0x10 // IBUS_MOD2_MASK: latched NumLock (live-observed)
	a, sink := wiredActor()

	for _, r := range wordEN {
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r), Mods: numLockLatch})
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r), Mods: numLockLatch, Release: true})
	}
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText(wordEN, runeLen(wordEN), runeLen(wordEN))

	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != wordRU {
		t.Fatalf("commits under NumLock = %q, want one %q", texts, wordRU)
	}

	// The negative control: a Ctrl-modified press is a combo, not a
	// character — it must not feed the buffer either.
	b, sink2 := wiredActor()
	b.HandleKey(engine.EngineEvent{Keyval: uint32('c'), Mods: engine.MaskControl})
	b.HandleKey(engine.EngineEvent{Keyval: uint32('c'), Mods: engine.MaskControl, Release: true})
	tapShift(b)
	tapShift(b)
	b.ExpiryAt(expiryAfterWindow)
	if got := sink2.requireCount(); got != 0 {
		t.Errorf("Ctrl-combo press fed the buffer (started verification %d times), want 0", got)
	}
}

// TestActor_VerifyPaths pins the ADR-004 abort discipline of the two-phase
// verification: a suffix mismatch and the verify timeout each skip the
// correction with the exact INFO reason — and not one of them deletes or
// commits anything ("abort, не мусорить").
func TestActor_VerifyPaths(t *testing.T) {
	t.Run("mismatch", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()

		typeWord(a, wordEN)
		// The honest-GTK world (plan 10-08 amendment migration, owner
		// verdict 2026-10-13): the client PUSHED its field state during the
		// typing — the cache carries evidence the verify channel tracks.
		// The nil-cache form of this scenario is the stale-cache terminal
		// signature now (TestTerminalBranch_StaleCacheAnswerRoutesTerminal).
		a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
		mismatch := mismatchLine
		a.HandleSurroundingText(mismatch, runeLen(mismatch), runeLen(mismatch)) // does not end with the token

		if !strings.Contains(buf.String(), `"reason":"verify-mismatch"`) {
			t.Errorf("verify-mismatch record missing; log:\n%s", buf.String())
		}
		if calls := sink.deleteCalls(); len(calls) != 0 {
			t.Errorf("mismatch deleted %+v — abort, не мусорить", calls)
		}
		if texts := sink.commitTexts(); len(texts) != 0 {
			t.Errorf("mismatch committed %q — abort, не мусорить", texts)
		}

		// A late surrounding text (the client answering after the verdict)
		// must not resurrect the correction: the pending fix is gone.
		a.HandleSurroundingText("abc "+wordEN, runeLen("abc "+wordEN), runeLen("abc "+wordEN))
		if calls := sink.deleteCalls(); len(calls) != 0 {
			t.Errorf("late surrounding text deleted %+v after the mismatch verdict", calls)
		}
		if texts := sink.commitTexts(); len(texts) != 0 {
			t.Errorf("late surrounding text committed %q after the mismatch verdict", texts)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()
		clock := &flipClock{} // the verify-deadline recorder (plan 10-10, UseVerifyTimer)
		a.UseVerifyTimer(clock.arm)

		typeWord(a, wordEN)
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
		clock.fire() // deterministic injection of the 100 ms deadline

		if !strings.Contains(buf.String(), `"reason":"verify-timeout"`) {
			t.Errorf("verify-timeout record missing; log:\n%s", buf.String())
		}
		if calls := sink.deleteCalls(); len(calls) != 0 {
			t.Errorf("timeout deleted %+v — abort, не мусорить", calls)
		}
		if texts := sink.commitTexts(); len(texts) != 0 {
			t.Errorf("timeout committed %q — abort, не мусорить", texts)
		}
	})
}

// TestActor_TokenRefusals pins the pipeline-entry refusals of the D-20
// vocabulary: a letterless token skips with the exact INFO reason and never
// even starts the verification round. (The mixed-script subtest is
// superseded by the D-16→D-23 succession — TestActor_MixedWordConverts-
// ForeignRuns converts the foreign runs; the 02-03 "no surrounding
// capability" refusal pin is superseded by TestActor_Level2NoCaps: a
// no-caps client is no longer refused — ladder level 2 executes.)
func TestActor_TokenRefusals(t *testing.T) {
	t.Run("no letters", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()

		typeWord(a, "2026") // digits only: no direction
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)

		if !strings.Contains(buf.String(), `"reason":"no-letters"`) {
			t.Errorf("no-letters record missing; log:\n%s", buf.String())
		}
		if got := sink.requireCount(); got != 0 {
			t.Errorf("letterless token started verification %d times, want 0", got)
		}
	})
}

// TestActor_EmptyBufferNoDestructive pins the empty-input edge of CORR-07:
// a double tap with no token in the buffer skips with the empty-buffer
// reason and makes zero calls of any kind on the sink.
func TestActor_EmptyBufferNoDestructive(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if !strings.Contains(buf.String(), `"reason":"empty-buffer"`) {
		t.Errorf("empty-buffer record missing; log:\n%s", buf.String())
	}
	if got := sink.requireCount(); got != 0 {
		t.Errorf("empty buffer asked for surrounding text %d times, want 0", got)
	}
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Errorf("empty buffer deleted %+v, want nothing", calls)
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Errorf("empty buffer committed %q, want nothing", texts)
	}
}

// pressSpace feeds one space press/release pair — the D-13 separator.
func pressSpace(a *session.Actor) {
	a.HandleKey(engine.EngineEvent{Keyval: engine.KeySpace})
	a.HandleKey(engine.EngineEvent{Keyval: engine.KeySpace, Release: true})
}

// TestActor_AfterSpaceCorrects pins the D-13 tail geometry on the actor:
// the word already separated by a space is corrected TOGETHER with the
// separator — exactly one DeleteSurroundingText(-7,7) (token+tail) and one
// CommitText("привет ") — never the token alone, which would land the
// correction behind the surviving space (Pitfall 1).
func TestActor_AfterSpaceCorrects(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	typeWord(a, wordEN)
	pressSpace(a)
	honestPush(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != 1 {
		t.Fatalf("RequireSurroundingText calls after double tap = %d, want 1", got)
	}

	line := "abc " + wordEN + " "
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -7, nchars: 7}) {
		t.Fatalf("deletions = %+v, want exactly one (-7,7) — token+tail, D-13", calls)
	}
	// The two-phase completion (plan 10-11): the client reports the
	// deletion applied («abc ») — the replacement commits.
	a.HandleSurroundingText("abc ", 4, 4)
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != wordRU+" " {
		t.Fatalf("commits = %q, want exactly one %q", texts, wordRU+" ")
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done"`) {
		t.Errorf("INFO completion record missing; log:\n%s", buf.String())
	}
}

// TestActor_DoubleTapPunctuationStretch pins the convertible stretch end to
// end (SPEC §4.2, plan 10-02): a word typed with its trailing table
// punctuation («ytn?») corrects WHOLE on the GTK cached-push path — one
// DeleteSurroundingText over the 4-rune stretch (not the 3-rune token) and
// one CommitText of «нет,». «нет?» — the punctuation riding the tail as
// typed — is the defect this pins out.
func TestActor_DoubleTapPunctuationStretch(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	const typed = "ytn?"
	for i, r := range typed {
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r)})
		a.HandleSurroundingText(typed[:i+1], uint32(i+1), uint32(i+1)) // the client's spontaneous push
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r), Release: true})
	}
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != 0 {
		t.Errorf("require calls before the completion = %d, want 0 — the verify-after arms at the "+
			"flush (plan 10-11), never a pre-correction one", got)
	}
	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -4, nchars: 4}) {
		t.Fatalf("deletions = %+v, want exactly one (-4,4) — the whole stretch, not the 3-rune token", calls)
	}
	// The two-phase completion (plan 10-11): the client reports the
	// deletion applied (the field is empty) — the replacement commits and
	// the verify-after Require rides the flush.
	a.HandleSurroundingText("", 0, 0)
	if got := sink.requireCount(); got != 1 {
		t.Errorf("require calls after the completion = %d, want exactly 1 — the verify-after round", got)
	}
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != "нет," {
		t.Fatalf("commits = %q, want exactly one %q — «ytn?» corrects to «нет,», never «нет?»", texts, "нет,")
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done"`) {
		t.Errorf("INFO completion record missing; log:\n%s", buf.String())
	}
}

// TestActor_ToggleRepeat pins the ReplaceToken toggle invariant through the
// whole pipeline: after a successful correction the buffer holds the
// converted word, so a repeated double tap converts it BACK — the field
// mirrors what the buffer believes (plan 02-01 invariant, wired by 02-03).
func TestActor_ToggleRepeat(t *testing.T) {
	a, sink := wiredActor()

	// First correction: ghbdtn → привет.
	typeWord(a, wordEN)
	honestPush(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText(wordEN, runeLen(wordEN), runeLen(wordEN))
	a.HandleSurroundingText("", 0, 0) // the deletion proof completes the ladder (plan 10-11)

	// Repeat: the buffer holds привет, the field too.
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText(wordRU, runeLen(wordRU), runeLen(wordRU))
	a.HandleSurroundingText("", 0, 0) // the deletion proof completes the ladder (plan 10-11)

	texts := sink.commitTexts()
	if len(texts) != 2 || texts[0] != wordRU || texts[1] != wordEN {
		t.Fatalf("commits = %q, want [%s %s] — the second double tap toggles back", texts, wordRU, wordEN)
	}
	calls := sink.deleteCalls()
	if len(calls) != 2 {
		t.Fatalf("deletions = %+v, want exactly two (one per correction)", calls)
	}
	for i, want := range []deleteCall{{offset: -6, nchars: 6}, {offset: -6, nchars: 6}} {
		if calls[i] != want {
			t.Errorf("deletion %d = %+v, want %+v", i, calls[i], want)
		}
	}
}

// TestActor_DebugCorrectionRecord pins the D-21 log contract of a
// successful correction: the DEBUG record carries the ladder level as its
// first attribute after msg (the matrix greps the exact form), with the
// source, result and latency — and no INFO record in the whole stream
// contains a word of the correction (D-20).
func TestActor_DebugCorrectionRecord(t *testing.T) {
	buf := captureLogsLevel(t, slog.LevelDebug)
	a, sink := wiredActor()

	typeWord(a, wordEN)
	honestPush(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText(wordEN, runeLen(wordEN), runeLen(wordEN))
	a.HandleSurroundingText("abc ", 4, 4) // the deletion proof completes the ladder (plan 10-11)

	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != wordRU {
		t.Fatalf("commits = %q, want one %q (precondition of the log pin)", texts, wordRU)
	}

	logged := buf.String()
	if !strings.Contains(logged, `"msg":"correction","level":1`) {
		t.Errorf("DEBUG correction record with level-first attribute missing; log:\n%s", logged)
	}
	for _, want := range []string{
		`"source":"` + wordEN + `"`,
		`"result":"` + wordRU + `"`,
		`"latency_ms"`,
	} {
		if !strings.Contains(logged, want) {
			t.Errorf("DEBUG correction record misses %s; log:\n%s", want, logged)
		}
	}

	// D-20: every INFO record is word-free.
	for _, line := range strings.Split(logged, "\n") {
		if !strings.Contains(line, `"level":"INFO"`) {
			continue
		}
		if strings.Contains(line, wordEN) || strings.Contains(line, wordRU) {
			t.Errorf("INFO record leaks correction contents: %s", line)
		}
	}
}

// phraseEN is the tracer phrase of plan 03-01: two SPEC words and the
// separator between them — the smallest phrase that is MORE than a token.
const phraseEN = wordEN + " " + wordEN

// phraseRU is the converted expectation of the same phrase: Convert passes
// the separator identically (CORR-05).
const phraseRU = wordRU + " " + wordRU

// TestActor_TripleTapCorrectsPhrase pins the phrase tracer end to end
// (CORR-02, D-25): printable presses feed the buffer with a whole phrase, a
// triple tap at window expiry starts the same two-phase pipeline as the word
// — RequireSurroundingText on the sink, nothing destructive yet — and the
// matching SetSurroundingText resolves it over the WHOLE phrase range:
// exactly one DeleteSurroundingText(-13,13) (both words and the separator,
// no tail) and one CommitText("привет привет"), plus the INFO completion
// record.
func TestActor_TripleTapCorrectsPhrase(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	typeWord(a, wordEN)
	pressSpace(a)
	typeWord(a, wordEN)
	honestPush(a)
	tapShift(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != 1 {
		t.Fatalf("RequireSurroundingText calls after triple tap = %d, want 1", got)
	}
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Fatalf("two-phase violation: %d deletions before surrounding text, want 0", len(calls))
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Fatalf("two-phase violation: %d commits before surrounding text, want 0", len(texts))
	}

	line := "abc " + phraseEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -13, nchars: 13}) {
		t.Fatalf("deletions = %+v, want exactly one (-13,13) — the whole phrase range", calls)
	}
	// The two-phase completion (plan 10-11): the client reports the
	// deletion applied («abc ») — the replacement commits.
	a.HandleSurroundingText("abc ", 4, 4)
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != phraseRU {
		t.Fatalf("commits = %q, want exactly one %q", texts, phraseRU)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done"`) {
		t.Errorf("INFO completion record missing; log:\n%s", buf.String())
	}
}

// TestActor_TripleTapEmptyBuffer pins the D-20/D-25 refusal of the phrase
// path: a triple tap with an empty buffer skips with the empty-buffer reason
// and makes zero calls of any kind on the sink — the phrase correction never
// guesses at an empty range.
func TestActor_TripleTapEmptyBuffer(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	tapShift(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if !strings.Contains(buf.String(), `"reason":"empty-buffer"`) {
		t.Errorf("empty-buffer record missing; log:\n%s", buf.String())
	}
	if got := sink.requireCount(); got != 0 {
		t.Errorf("empty buffer asked for surrounding text %d times, want 0", got)
	}
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Errorf("empty buffer deleted %+v, want nothing", calls)
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Errorf("empty buffer committed %q, want nothing", texts)
	}
}

// TestActor_PhraseVerifyMismatch pins the ADR-004 abort discipline of the
// phrase range (T-03-01-01): a surrounding text whose head does not match the
// phrase — the client reports a field the buffer does not mirror — skips the
// correction with the verify-mismatch reason and touches nothing ("abort, не
// мусорить"), and a late matching push cannot resurrect the round.
func TestActor_PhraseVerifyMismatch(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	typeWord(a, wordEN)
	pressSpace(a)
	typeWord(a, wordEN)
	// The honest-GTK world (plan 10-08 amendment migration, owner verdict
	// 2026-10-13): the pushed cache carries tracking evidence — see the word
	// path's mismatch cell.
	a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))
	tapShift(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	mismatch := mismatchLine
	a.HandleSurroundingText(mismatch, runeLen(mismatch), runeLen(mismatch)) // does not end with the phrase

	if !strings.Contains(buf.String(), `"reason":"verify-mismatch"`) {
		t.Errorf("verify-mismatch record missing; log:\n%s", buf.String())
	}
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Errorf("mismatch deleted %+v — abort, не мусорить", calls)
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Errorf("mismatch committed %q — abort, не мусорить", texts)
	}

	// A late matching push must not resurrect the correction: the pending
	// fix is gone with the verdict.
	line := "abc " + phraseEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Errorf("late surrounding text deleted %+v after the mismatch verdict", calls)
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Errorf("late surrounding text committed %q after the mismatch verdict", texts)
	}
}

// backSpaceOp is the unified op-log entry of one level-2 Backspace replay's
// PRESS half: keyval 0xff08, physical keycode 14, no modifiers (RESEARCH
// Pattern 1 — the wire contract the actor must emit verbatim).
func backSpaceOp() string {
	return fmt.Sprintf("forward(%d,%d,%d)", engine.KeyBackSpace, 14, 0)
}

// backSpaceReleaseOp is the RELEASE half of the same replay (round-6): the
// same key with exactly MaskRelease — the emission that disarms the
// client's own key repeat (probe.go's ProbeForward discipline).
func backSpaceReleaseOp() string {
	return fmt.Sprintf("forward(%d,%d,%d)", engine.KeyBackSpace, 14, engine.MaskRelease)
}

// backSpaceWireCall is the recorded-argument form of the same replay's
// press half.
func backSpaceWireCall() forwardCall {
	return forwardCall{keyval: engine.KeyBackSpace, keycode: 14, state: 0}
}

// backSpaceReleaseWireCall is the recorded-argument release half (round-6).
func backSpaceReleaseWireCall() forwardCall {
	return forwardCall{keyval: engine.KeyBackSpace, keycode: 14, state: engine.MaskRelease}
}

// backSpacePairOps is the op-log pair of ONE replayed BackSpace: the press
// immediately followed by its release (round-6 pair discipline).
func backSpacePairOps() []string {
	return []string{backSpaceOp(), backSpaceReleaseOp()}
}

// assertBackSpacePairs fails unless calls is the exact PAIR sequence: one
// press (state 0) followed by one release (state MaskRelease) per expected
// BackSpace key — the round-6 wire contract of the whole write branch.
func assertBackSpacePairs(t *testing.T, calls []forwardCall, keys int, where string) {
	t.Helper()

	if len(calls) != 2*keys {
		t.Fatalf("%s: forward count = %d, want %d (a press+release pair per BackSpace)",
			where, len(calls), 2*keys)
	}
	for i := 0; i < len(calls); i += 2 {
		if calls[i] != backSpaceWireCall() {
			t.Fatalf("%s: pair[%d] press = %+v, want %+v", where, i/2, calls[i], backSpaceWireCall())
		}
		if calls[i+1] != backSpaceReleaseWireCall() {
			t.Fatalf("%s: pair[%d] release = %+v, want %+v — the release disarms the client's "+
				"key repeat", where, i/2, calls[i+1], backSpaceReleaseWireCall())
		}
	}
}

// burstOps builds the unified op-log suffix of a level-2 correction: n
// Backspace press+release pairs followed by the single replacement commit.
func burstOps(n int, commit string) []string {
	ops := make([]string, 0, 2*n+1)
	for range n {
		ops = append(ops, backSpacePairOps()...)
	}

	return append(ops, "commit("+commit+")")
}

// TestActor_Level2NoCaps pins the level-2 ladder on a client WITHOUT the
// surrounding-text bit (CORR-07, the ADR-004 degradation): verification is
// impossible, so the correction trusts the buffer plus the explicit CORR-09
// reset triggers and executes IMMEDIATELY at the Double decision — zero
// RequireSurroundingText calls, exactly six BackSpace press+release PAIRS
// (one pair per RUNE of the token, Pitfall 2; round-6) and one CommitText
// of the converted word, the burst→commit order pinned by the unified op
// log. This test supersedes the 02-03 "no-surrounding" refusal pin: level 2
// IS the replacement for that skip.
func TestActor_Level2NoCaps(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActorCaps(engine.CapPreeditText) // no CapSurroundingText

	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != 0 {
		t.Fatalf("no-caps client was asked for surrounding text %d times, want 0 — verification impossible", got)
	}
	if got, want := sink.opLog(), burstOps(len([]rune(wordEN)), wordRU); !slices.Equal(got, want) {
		t.Fatalf("op log = %q, want the 6-forward burst then one commit — burst→commit order", got)
	}
	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != wordRU {
		t.Errorf("commits = %q, want exactly one %q", texts, wordRU)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done"`) {
		t.Errorf("INFO completion record missing; log:\n%s", buf.String())
	}
}

// TestActor_Level2WithTailAndRunes pins the level-2 geometry through the
// real RU feeding branches (Pitfall 2): "ghbdtn" typed in RU mode commits
// «привет» rune by rune (the 02-04 script-true buffer) and the space
// transits, so the buffer holds "привет " — SIX token runes plus the
// one-rune tail. The RU→EN correction replays exactly SEVEN Backspaces
// (runes, never bytes — привет is 12 UTF-8 bytes and a byte count would
// erase twice the text) and commits "ghbdtn " with the tail re-committed.
func TestActor_Level2WithTailAndRunes(t *testing.T) {
	a, sink := wiredActorCaps(engine.CapPreeditText)

	flipMode(a)         // EN → RU: the Cyrillic buffer is fed by the 02-04 commits
	typeWord(a, wordEN) // commits привет; buffer script-true
	pressSpace(a)       // the D-13 separator: buffer "привет "
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	const tailRunes = 1 // the separator space
	tokenRunes := len([]rune(wordRU))
	if got := sink.requireCount(); got != 0 {
		t.Fatalf("no-caps client was asked for surrounding text %d times, want 0", got)
	}
	assertBackSpacePairs(t, sink.forwardCalls(), tokenRunes+tailRunes,
		"token+tail in RUNES, not bytes — one press+release pair per rune")
	ops := sink.opLog()
	burst := ops[len(ops)-(2*(tokenRunes+tailRunes)+1):] // the 7 pairs + the commit
	if want := burstOps(tokenRunes+tailRunes, wordEN+" "); !slices.Equal(burst, want) {
		t.Fatalf("op-log tail = %q, want %q — burst→commit order", burst, want)
	}
	if texts := sink.commitTexts(); texts[len(texts)-1] != wordEN+" " {
		t.Errorf("correction commit = %q, want %q (converted+tail)", texts[len(texts)-1], wordEN+" ")
	}
}

// TestActor_Level2ForwardsComeInPressReleasePairs pins the pair discipline
// of the level-2 write branch (the round-6 sourcedig verdict,
// .planning/debug/2026-10-19-wezterm-repeat-sourcedig.md): replayBurst emits
// ForwardKeyEvent(BackSpace) press-ONLY (state=0, MaskRelease never set), and
// wezterm e2a98e1b arms its CLIENT-SIDE auto-repeat on every repeatable-key
// press (window.rs:1194-1206) — a repeat disarmed ONLY by the same key's
// release, another key's press, or a keyboard focus Enter/Leave. With no
// releases the armed repeat fires after the GNOME repeat-delay (default
// 500 ms, delivered via wl_keyboard.repeat_info) at ~30 Hz and deletes the
// freshly committed word — the host defect «появилось и исчезло» (the 21:56
// video, the 19:03 selftype). The wire contract is therefore ProbeForward's
// discipline (probe.go:179-183, the forwardCtrlV form actor.go:2669-2672):
// every press forward is followed by a release forward of the SAME key, the
// release carrying exactly MaskRelease (no other state bits) — no client may
// leave the burst with an armed repeat. RED truth: on the unchanged HEAD the
// burst emits press-only forwards — the count and the pair form both fail
// (evidence-20261019-10-17-red-pairs.json).
func TestActor_Level2ForwardsComeInPressReleasePairs(t *testing.T) {
	a, sink := wiredActorCaps(engine.CapPreeditText) // no CapSurroundingText — the write branch

	typeWord(a, wordEN) // six runes — six BackSpace pairs on the wire
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	forwards := sink.forwardCalls()
	if len(forwards) != 2*len([]rune(wordEN)) {
		t.Fatalf("forward count = %d, want %d — every BackSpace must ride a press+release PAIR",
			len(forwards), 2*len([]rune(wordEN)))
	}
	for i := 0; i < len(forwards); i += 2 {
		press, release := forwards[i], forwards[i+1]
		if press != backSpaceWireCall() {
			t.Errorf("pair[%d] press = %+v, want %+v", i/2, press, backSpaceWireCall())
		}
		wantRelease := forwardCall{keyval: engine.KeyBackSpace, keycode: 14, state: engine.MaskRelease}
		if release != wantRelease {
			t.Errorf("pair[%d] release = %+v, want %+v — the release is what disarms the "+
				"client's key repeat", i/2, release, wantRelease)
		}
		if press.keyval != release.keyval || press.keycode != release.keycode {
			t.Errorf("pair[%d] = press %+v vs release %+v — the release must answer the SAME key",
				i/2, press, release)
		}
	}
	// The op log keeps the pair ADJACENT: a release that trails a foreign
	// emission cannot answer the press it must disarm.
	ops := sink.opLog()
	for i := 0; i < 2*len([]rune(wordEN)); i += 2 {
		if ops[i] != backSpaceOp() || !strings.HasPrefix(ops[i+1], fmt.Sprintf(
			"forward(%d,%d,", engine.KeyBackSpace, 14)) {
			t.Fatalf("op log pair %d = %q,%q — the release must sit immediately after its press",
				i/2, ops[i], ops[i+1])
		}
	}
}

// verifyAfterSettled runs one complete level-1 correction and returns the
// actor at the moment the verify-after round is pending.
func verifyAfterSettled(t *testing.T) (*session.Actor, *fakeSink, *syncBuffer) {
	t.Helper()
	buf := captureLogs(t)
	a, sink := wiredActor()

	typeWord(a, wordEN)
	honestPush(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	line := "abc " + wordEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line)) // settles the pre-correction round
	// The two-phase completion (plan 10-11): the client's confirming push
	// completes the ladder — the verify-after arms here (the require
	// counter's second call).
	a.HandleSurroundingText("abc ", 4, 4)

	return a, sink, buf
}

// TestActor_VerifyAfterLevel1 pins the post-correction check of the level-1
// ladder (ADR-003/ADR-004: DeleteSurroundingText is ack-less on 1.5.29, the
// verify-after is the compensation): the correction itself asks for a fresh
// surrounding text — the require counter grows past the pre-correction
// round — and the answer decides. A field ending with the expected
// replacement is quiet; a field that does not is DEBOUNCED first (WR-05:
// one re-require round, because the push may be the intermediate
// post-delete pre-commit state) and only the second stale answer raises
// the INFO mismatch counter exactly once, NEVER followed by a second
// correction ("abort, не мусорить" — no auto-repair, the residual risk
// stays documented in ADR-003).
func TestActor_VerifyAfterLevel1(t *testing.T) {
	t.Run("match is quiet", func(t *testing.T) {
		a, sink, buf := verifyAfterSettled(t)

		if got := sink.requireCount(); got != 2 {
			t.Fatalf("require calls after the correction = %d, want 2 (pre-correction round + verify-after)", got)
		}
		// The fresh verify answer: the field holds the replacement.
		after := "abc " + wordRU
		a.HandleSurroundingText(after, runeLen(after), runeLen(after))

		if strings.Contains(buf.String(), `"msg":"correction verify","outcome":"mismatch"`) {
			t.Errorf("a matching field raised a mismatch; log:\n%s", buf.String())
		}
		if calls := sink.deleteCalls(); len(calls) != 1 {
			t.Errorf("matching verify-after re-corrected: deletions = %+v, want 1", calls)
		}
		if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != wordRU {
			t.Errorf("matching verify-after re-corrected: commits = %q", texts)
		}
	})

	t.Run("mismatch counts once without repair", func(t *testing.T) {
		afterMismatchDebouncedOnce(t)
	})

	t.Run("the intermediate post-delete push is debounced into a match", func(t *testing.T) {
		afterIntermediatePushDebounced(t)
	})
}

// afterMismatchDebouncedOnce is the concluding half of the WR-05 protocol:
// the first stale answer re-requires (no verdict), the second concludes the
// mismatch exactly once with no repair, and the quenched round ignores
// everything after it.
func afterMismatchDebouncedOnce(t *testing.T) {
	t.Helper()
	a, sink, buf := verifyAfterSettled(t)

	// The correction never landed in the field (the plan's oracle): the
	// client reports the pre-correction text — the suffix no longer
	// matches the expected replacement.
	uncorrected := "abc " + wordEN
	a.HandleSurroundingText(uncorrected, runeLen(uncorrected), runeLen(uncorrected))
	// WR-05: the FIRST stale answer is debounced, not concluded — the
	// round issues one more Require instead of verdicting on a push
	// that may predate the commit.
	if got := strings.Count(buf.String(), `"msg":"correction verify","outcome":"mismatch"`); got != 0 {
		t.Fatalf("the first stale answer concluded: mismatch counter = %d, want 0; log:\n%s", got, buf.String())
	}
	if got := sink.requireCount(); got != 3 {
		t.Fatalf("require calls after the debounced answer = %d, want 3 (pre-round + verify-after + retry)", got)
	}

	// The SECOND stale answer concludes exactly once.
	a.HandleSurroundingText(uncorrected, runeLen(uncorrected), runeLen(uncorrected))
	if got := strings.Count(buf.String(), `"msg":"correction verify","outcome":"mismatch"`); got != 1 {
		t.Fatalf("mismatch counter = %d, want exactly 1; log:\n%s", got, buf.String())
	}
	if calls := sink.deleteCalls(); len(calls) != 1 {
		t.Errorf("mismatch triggered a repair: deletions = %+v, want 1 — no auto-repeat", calls)
	}
	if texts := sink.commitTexts(); len(texts) != 1 {
		t.Errorf("mismatch triggered a repair: commits = %q, want 1", texts)
	}

	// pendingAfter is quenched: a further push does nothing at all.
	a.HandleSurroundingText(wordRU, runeLen(wordRU), runeLen(wordRU))
	if got := strings.Count(buf.String(), `"msg":"correction verify","outcome":"mismatch"`); got != 1 {
		t.Errorf("quenched pendingAfter re-fired: mismatch counter = %d, want 1", got)
	}
	if texts := sink.commitTexts(); len(texts) != 1 {
		t.Errorf("quenched pendingAfter re-corrected: commits = %q, want 1", texts)
	}
}

// afterIntermediatePushDebounced is the race half of the WR-05 protocol:
// the client pushes the INTERMEDIATE post-delete pre-commit state — the
// old text — and only then the committed one; the verdict waits out the
// stale answer (with the clipboard rung on, a false mismatch would paste a
// duplicate of the just-applied correction into the field).
func afterIntermediatePushDebounced(t *testing.T) {
	t.Helper()
	a, sink, buf := verifyAfterSettled(t)

	uncorrected := "abc " + wordEN
	a.HandleSurroundingText(uncorrected, runeLen(uncorrected), runeLen(uncorrected))
	if strings.Contains(buf.String(), `"msg":"correction verify","outcome":"mismatch"`) {
		t.Fatalf("the intermediate push raised a mismatch; log:\n%s", buf.String())
	}

	after := "abc " + wordRU
	a.HandleSurroundingText(after, runeLen(after), runeLen(after))
	if strings.Contains(buf.String(), `"msg":"correction verify","outcome":"mismatch"`) {
		t.Errorf("the settled field still raised a mismatch; log:\n%s", buf.String())
	}
	if calls := sink.deleteCalls(); len(calls) != 1 {
		t.Errorf("the debounced round re-corrected: deletions = %+v, want 1", calls)
	}
	if texts := sink.commitTexts(); len(texts) != 1 {
		t.Errorf("the debounced round re-corrected: commits = %q, want 1", texts)
	}
}

// TestActor_HardResetKeyvals pins the keyval half of CORR-09: Enter and its
// keypad variant, Tab and Escape each hard-reset the phrase buffer on
// press — after any of them a double tap finds an EMPTY buffer and skips
// with the empty-buffer reason, zero calls of any kind on the sink. The
// reset is engine state, never consumption: the key transits (the client
// sees its Enter/Tab/Escape exactly as before). The keypad Enter is covered
// explicitly — 0xff8b is a distinct keyval that must ride the same table.
func TestActor_HardResetKeyvals(t *testing.T) {
	for _, tc := range []struct {
		name   string
		keyval uint32
	}{
		{"Return", engine.KeyReturn},
		{"KP_Enter", engine.KeyKPEnter},
		{"Tab", engine.KeyTab},
		{"Escape", engine.KeyEscape},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buf := captureLogs(t)
			a, sink := wiredActor()

			typeWord(a, wordEN)
			if consume := a.HandleKey(engine.EngineEvent{Keyval: tc.keyval}); consume {
				t.Errorf("%s press: consume = true, want false — a reset transits", tc.name)
			}
			a.HandleKey(engine.EngineEvent{Keyval: tc.keyval, Release: true})
			tapShift(a)
			tapShift(a)
			a.ExpiryAt(expiryAfterWindow)

			if !strings.Contains(buf.String(), `"reason":"empty-buffer"`) {
				t.Errorf("%s left a correctable token — buffer not hard-reset; log:\n%s", tc.name, buf.String())
			}
			if got := sink.requireCount(); got != 0 {
				t.Errorf("%s: sink saw %d RequireSurroundingText calls, want 0", tc.name, got)
			}
			if calls := sink.deleteCalls(); len(calls) != 0 {
				t.Errorf("%s: sink saw deletions %+v, want none", tc.name, calls)
			}
			if texts := sink.commitTexts(); len(texts) != 0 {
				t.Errorf("%s: sink saw commits %q, want none", tc.name, texts)
			}
			if got := sink.forwardCalls(); len(got) != 0 {
				t.Errorf("%s: sink saw forwards %+v, want none", tc.name, got)
			}
		})
	}
}

// TestBuffer_ResetByFocusOut pins the lifecycle half of CORR-09 (wired
// since 02-03, pinned here as part of the reset corpus): FocusOut and Reset
// both hard-reset the buffer — the phrase belongs to the input context that
// just left.
func TestBuffer_ResetByFocusOut(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind engine.LifecycleKind
	}{
		{"FocusOut", engine.LifecycleFocusOut},
		{"Reset", engine.LifecycleReset},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buf := captureLogs(t)
			a, sink := wiredActor()

			typeWord(a, wordEN)
			a.HandleLifecycle(tc.kind)
			tapShift(a)
			tapShift(a)
			a.ExpiryAt(expiryAfterWindow)

			if !strings.Contains(buf.String(), `"reason":"empty-buffer"`) {
				t.Errorf("%s left a correctable token — buffer not hard-reset; log:\n%s", tc.name, buf.String())
			}
			if got := sink.requireCount(); got != 0 {
				t.Errorf("%s: sink saw %d RequireSurroundingText calls, want 0", tc.name, got)
			}
		})
	}
}

// TestBuffer_CtrlIsolation pins the combo isolation of the buffer (Pitfall 7
// measure, implemented by the 02-04 comboMask guard — the corpus pins it as
// part of the reset semantics): a Ctrl-modified press is a shortcut, not
// text — in BOTH script modes the buffer must not change, and a Ctrl+Shift
// chord is a combo all the same.
func TestBuffer_CtrlIsolation(t *testing.T) {
	assertEmpty := func(t *testing.T, a *session.Actor, sink *fakeSink, logs *syncBuffer, label string) {
		t.Helper()
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
		if !strings.Contains(logs.String(), `"reason":"empty-buffer"`) {
			t.Errorf("%s: buffer was fed by a combo (no empty-buffer skip); log:\n%s", label, logs.String())
		}
		if got := sink.requireCount(); got != 0 {
			t.Errorf("%s: sink saw %d RequireSurroundingText calls, want 0", label, got)
		}
	}

	t.Run("EN ctrl combo", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor() // starts in EN

		a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskControl})
		a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskControl, Release: true})
		assertEmpty(t, a, sink, buf, "EN Ctrl+a")
	})

	t.Run("RU ctrl combo", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()
		flipMode(a) // EN → RU

		a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskControl})
		a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskControl, Release: true})
		assertEmpty(t, a, sink, buf, "RU Ctrl+a")
	})

	t.Run("ctrl shift chord", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()

		a.HandleKey(engine.EngineEvent{Keyval: uint32('c'), Mods: engine.MaskControl | engine.MaskShift})
		a.HandleKey(engine.EngineEvent{Keyval: uint32('c'), Mods: engine.MaskControl | engine.MaskShift, Release: true})
		assertEmpty(t, a, sink, buf, "Ctrl+Shift+c")
	})
}

// flipMode delivers one clean Shift_R tap and expires the window: the
// runtime path of the Single decision — the flip fires at expiry, not inside
// HandleKey (D-04).
func flipMode(a *session.Actor) {
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
}

// TestActor_FlipOnSingle pins the ADR-001 Option B flip: a single-tap series
// resolved at window expiry switches the actor's internal script mode EN→RU
// and back RU→EN, each switch logging the exact INFO mode record — the flip
// is still purely internal toward the FIELD (the XKB group of the session is
// not touched, D-01 verdict; zero CommitText/DeleteSurroundingText/
// ForwardKeyEvent on a bare flip), but each switch now refreshes the panel
// symbol (owner decision 1, quick plan 260927-way): exactly one
// UpdateModeSymbol per flip, "ru" then "en" — reconciled deliberately from
// the vu8-era zero-sink-call pin, decision 1 cited.
func TestActor_FlipOnSingle(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	flipMode(a) // EN → RU
	flipMode(a) // RU → EN

	logged := buf.String()
	if !strings.Contains(logged, `"msg":"mode","to":"ru"`) {
		t.Errorf("first flip: mode-to-ru record missing; log:\n%s", logged)
	}
	if !strings.Contains(logged, `"msg":"mode","to":"en"`) {
		t.Errorf("second flip: mode-to-en record missing; log:\n%s", logged)
	}
	if strings.Index(logged, `"to":"ru"`) > strings.Index(logged, `"to":"en"`) {
		t.Errorf("mode records out of order (ru must precede en); log:\n%s", logged)
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru", "en"}) {
		t.Errorf("panel symbols = %q, want exactly [ru en] (one per flip, owner decision 1)", got)
	}
	if got := sink.requireCount(); got != 0 {
		t.Errorf("flip made %d RequireSurroundingText calls, want 0", got)
	}
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Errorf("flip deleted %+v, want nothing", calls)
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Errorf("flip committed %q, want nothing", texts)
	}
	if calls := sink.forwardCalls(); len(calls) != 0 {
		t.Errorf("flip forwarded %+v, want nothing", calls)
	}
}

// TestActor_FlipEmitsPanelSymbol pins the ORDER of the mode-indicator emit
// (owner decision 1, quick plan 260927-way): the panel symbol update lands
// strictly AFTER the flip's INFO mode record — the e2e sequencing contract
// extended, the mode log record stays the flip's observable — and the
// flip-back emits "en" with the same shape.
func TestActor_FlipEmitsPanelSymbol(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	var emittedAfterLog bool
	sink.modeHook = func(symbol string) {
		// At the emit instant the flip's own INFO record must already be in
		// the log — the emit fires after it, under the same mutex hold.
		emittedAfterLog = strings.Contains(buf.String(), `"msg":"mode","to":"`+symbol+`"`)
	}

	flipMode(a) // EN → RU
	if !emittedAfterLog {
		t.Errorf("UpdateModeSymbol(ru) fired before the mode log record — the D-36-style order is violated; log:\n%s",
			buf.String())
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru"}) {
		t.Fatalf("panel symbols after the first flip = %q, want exactly [ru]", got)
	}

	flipMode(a) // RU → EN
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru", "en"}) {
		t.Errorf("panel symbols after the flip-back = %q, want exactly [ru en]", got)
	}
}

// TestActor_FocusInReassertsModeSymbol pins the self-heal of the indicator
// (owner decision 1, quick plan 260927-way): engine objects are minted per
// input context, so every FocusIn re-emits the CURRENT mode symbol — a newly
// minted engine context must not resurrect the stale EN registration of the
// factory's RegisterProperties while the actor sits in RU.
func TestActor_FocusInReassertsModeSymbol(t *testing.T) {
	a, sink := wiredActor()

	a.HandleLifecycle(engine.LifecycleFocusIn) // EN at start: re-assert "en"
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"en"}) {
		t.Fatalf("FocusIn symbols at start = %q, want exactly [en]", got)
	}

	flipMode(a) // EN → RU
	a.HandleLifecycle(engine.LifecycleFocusIn)
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"en", "ru", "ru"}) {
		t.Errorf("FocusIn symbols after the flip = %q, want [en ru ru] — the fresh context must see RU", got)
	}
}

// TestActor_RUConsumesPrintable pins the RU-mode table consumption (T-02-04-01
// mitigation): a clean printable press is consumed and the mapped Cyrillic
// rune is committed — 'g'→"п", '['→"х", '?'→"," (the shift level already
// rides in the keyval; both table levels are exercised by the plan corpus).
func TestActor_RUConsumesPrintable(t *testing.T) {
	a, sink := wiredActor()
	flipMode(a)

	for _, tc := range []struct {
		key  rune
		want string
	}{
		{'g', "п"},
		{'[', "х"},
		{'?', ","},
	} {
		if consume := a.HandleKey(engine.EngineEvent{Keyval: uint32(tc.key)}); !consume {
			t.Errorf("RU press %q: consume = false, want true", tc.key)
		}
	}
	if texts := sink.commitTexts(); len(texts) != 3 || texts[0] != "п" || texts[1] != "х" || texts[2] != "," {
		t.Fatalf("RU commits = %q, want [п х ,]", texts)
	}
}

// TestActor_RUTransitUnmapped pins the unmapped transit of RU mode: a key
// outside layouts.ENToRU (Return) is neither consumed nor committed and does
// not feed the buffer — its handling belongs to the client (and the
// correction reset wiring arrives in plan 02-05).
func TestActor_RUTransitUnmapped(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	flipMode(a)

	if consume := a.HandleKey(engine.EngineEvent{Keyval: engine.KeyReturn}); consume {
		t.Errorf("RU press Return: consume = true, want false (unmapped transit)")
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Fatalf("RU unmapped commits = %q, want none", texts)
	}

	// The buffer must not have been fed: a double tap finds it empty.
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	if !strings.Contains(buf.String(), `"reason":"empty-buffer"`) {
		t.Errorf("empty-buffer record missing — Return fed the buffer; log:\n%s", buf.String())
	}
}

// TestActor_RUCtrlModifiedNotCommitted pins the combo guard of RU mode: a
// Ctrl-modified press is a shortcut, not text — not consumed, not committed,
// never fed into the buffer (the buffer mirrors the field, and a combo
// inserts nothing).
func TestActor_RUCtrlModifiedNotCommitted(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	flipMode(a)

	if consume := a.HandleKey(engine.EngineEvent{Keyval: uint32('g'), Mods: engine.MaskControl}); consume {
		t.Errorf("RU Ctrl+g: consume = true, want false (combo, not text)")
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Fatalf("RU Ctrl+g commits = %q, want none", texts)
	}

	// The buffer must stay empty: a double tap finds nothing to correct.
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	if !strings.Contains(buf.String(), `"reason":"empty-buffer"`) {
		t.Errorf("empty-buffer record missing — combo fed the buffer; log:\n%s", buf.String())
	}
}

// TestActor_ENTransitUnchanged pins the EN mode as the untouched Phase 1
// path: a clean printable press transits (consume=false, zero commits) but
// still feeds the buffer — the correction pipeline then works on it exactly
// as before the flip mode existed.
func TestActor_ENTransitUnchanged(t *testing.T) {
	a, sink := wiredActor() // starts in EN

	if consume := a.HandleKey(engine.EngineEvent{Keyval: uint32('g')}); consume {
		t.Errorf("EN press 'g': consume = true, want false (transit)")
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Fatalf("EN transit commits = %q, want none", texts)
	}

	// Phase 1 semantics: the transited rune is in the buffer, so the
	// double tap corrects it through the whole pipeline.
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	if got := sink.requireCount(); got != 1 {
		t.Fatalf("RequireSurroundingText calls = %d, want 1 (transit must feed the buffer)", got)
	}
	a.HandleSurroundingText("g", runeLen("g"), runeLen("g"))
	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != "п" {
		t.Fatalf("EN-transit correction commits = %q, want one [п]", texts)
	}
}

// TestActor_RUScriptTrueAllBranches pins the script-true invariant (T-02-04-03
// mitigation) on EVERY printing branch of RU mode: a rune that lands in the
// field lands in the buffer. The committed branch feeds the committed
// Cyrillic rune ('g'→'п', '['→'х'); the identical-map branch transits
// ('2' maps to itself — pinned choice: no consume, no commit) but the client
// inserts the original rune, so the buffer takes it too. The correction on
// "п2х" converting back to "g2[" proves the buffer held the field runes.
func TestActor_RUScriptTrueAllBranches(t *testing.T) {
	a, sink := wiredActor()
	flipMode(a)

	typeWord(a, "g2[")

	// Only the two mapped runes were committed — '2' is the pinned
	// identical-map transit.
	if texts := sink.commitTexts(); len(texts) != 2 || texts[0] != "п" || texts[1] != "х" {
		t.Fatalf("RU branch commits = %q, want [п х] ('2' transits)", texts)
	}

	// The buffer must hold the field runes "п2х": the double-tap
	// correction converts the whole token back through RUToEN.
	honestPush(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText("п2х", runeLen("п2х"), runeLen("п2х"))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -3, nchars: 3}) {
		t.Fatalf("deletions = %+v, want exactly one (-3,3) — the whole script-true token", calls)
	}
	a.HandleSurroundingText("", 0, 0) // the deletion proof completes the ladder (plan 10-11)
	if texts := sink.commitTexts(); len(texts) != 3 || texts[2] != "g2[" {
		t.Fatalf("correction commits = %q, want the last one to be [g2[]", texts)
	}
}

// TestActor_RUDigitsFullPipeline is the edge-probe desync pin (T-02-04-03):
// digits inside a RU-typed word reach the field through the identical-map
// transit and MUST reach the buffer as well — otherwise the RU→EN correction
// range would overrun the token. "ghbdtn2026" typed in RU mode: the engine
// commits "привет", the digits transit, the buffer token is "привет2026"
// (10 runes, NOT 14 bytes), and the correction deletes exactly 10 runes and
// commits "ghbdtn2026".
func TestActor_RUDigitsFullPipeline(t *testing.T) {
	a, sink := wiredActor()
	flipMode(a)

	typeWord(a, "ghbdtn2026")

	if texts := sink.commitTexts(); len(texts) != 6 || texts[0]+texts[1]+texts[2]+texts[3]+texts[4]+texts[5] != wordRU {
		t.Fatalf("RU typing commits = %q, want the six runes of %q (digits transit)", texts, wordRU)
	}

	honestPush(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText(wordRU+"2026", runeLen(wordRU+"2026"), runeLen(wordRU+"2026"))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -10, nchars: 10}) {
		t.Fatalf("deletions = %+v, want exactly one (-10,10) — 10 runes, not 14 bytes", calls)
	}
	a.HandleSurroundingText("", 0, 0) // the deletion proof completes the ladder (plan 10-11)
	if texts := sink.commitTexts(); len(texts) != 7 || texts[6] != wordEN+"2026" {
		t.Fatalf("correction commits = %q, want the last one to be [%s2026]", texts, wordEN)
	}
}

// TestActor_ScriptTrueBuffer pins the full RU→EN correction on the fake
// sink: RU-typed letters enter the buffer as the committed Cyrillic runes,
// so the double tap converts "привет" back to "ghbdtn" — the second
// correction direction of CORR-01, enabled by the flip.
func TestActor_ScriptTrueBuffer(t *testing.T) {
	a, sink := wiredActor()
	flipMode(a)

	typeWord(a, wordEN)

	if texts := sink.commitTexts(); len(texts) != 6 || texts[0]+texts[1]+texts[2]+texts[3]+texts[4]+texts[5] != wordRU {
		t.Fatalf("RU typing commits = %q, want the six runes of %q", texts, wordRU)
	}

	honestPush(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText(wordRU, runeLen(wordRU), runeLen(wordRU))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -6, nchars: 6}) {
		t.Fatalf("deletions = %+v, want exactly one (-6,6)", calls)
	}
	a.HandleSurroundingText("", 0, 0) // the deletion proof completes the ladder (plan 10-11)
	if texts := sink.commitTexts(); len(texts) != 7 || texts[6] != wordEN {
		t.Fatalf("RU→EN correction commits = %q, want the last one to be [%s]", texts, wordEN)
	}
}

// TestActor_MixedWordInvertsPerChar pins the per-character inversion of the
// mixed word (spec-delta 2026-10-06; owner verdict, UAT of phase 6,
// 2026-10-03, commit 7dd46e9 — supersedes the D-16→D-23 foreign-run
// succession of plan 03-01): a mixed word assembled the way the desktop
// produces it — "gfb" typed in EN (transit), then the flip, then the rest
// typed in RU (committed Cyrillic runes) — inverts EVERY letter through
// its own script's table: the whole token range is deleted (-9,9) and
// "паиghbdtn" is committed, every letter flipped to the other layout.
func TestActor_MixedWordInvertsPerChar(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	typeWord(a, "gfb")  // EN part: transit, buffer fed as typed
	flipMode(a)         // EN → RU
	typeWord(a, wordEN) // RU part: commits "привет", buffer script-true

	// The RU part must have arrived as Cyrillic commits — the mixed word
	// "gfbпривет" is what the field holds.
	if texts := sink.commitTexts(); len(texts) != 6 || texts[0]+texts[1]+texts[2]+texts[3]+texts[4]+texts[5] != wordRU {
		t.Fatalf("RU typing commits = %q, want the six runes of %q", texts, wordRU)
	}

	honestPush(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != 1 {
		t.Fatalf("mixed word started verification %d times, want 1 — the run conversion runs the pipeline", got)
	}
	if strings.Contains(buf.String(), `"reason":"mixed-script"`) {
		t.Errorf("mixed-script refusal fired — the D-16→D-23 succession removed it; log:\n%s", buf.String())
	}

	// The verification sees the whole mixed token — exactly what the ladder
	// deletes — and the settle commits the per-character inversion.
	field := "gfb" + wordRU
	a.HandleSurroundingText(field, runeLen(field), runeLen(field))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -9, nchars: 9}) {
		t.Fatalf("deletions = %+v, want exactly one (-9,9) — the whole mixed token range", calls)
	}
	a.HandleSurroundingText("", 0, 0) // the deletion proof completes the ladder (plan 10-11)
	texts := sink.commitTexts()
	if len(texts) != 7 || texts[6] != mixedRU {
		t.Fatalf("correction commits = %q, want the last one to be [%s]", texts, mixedRU)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done"`) {
		t.Errorf("INFO completion record missing; log:\n%s", buf.String())
	}
}

// The own-flip corpus (ADR-004 amendment 2026-10-07, G-2-3): the daemon's
// OWN engine flip — the SetGlobalEngine round trip inside flipTo through
// the D-52 seam — makes ibus re-mint engine objects, which delivers a
// SYNTHETIC focus_out/focus_in lifecycle pair to the engine. A synthetic
// transition is not «смена фокуса окна»: the correction buffer survives
// the daemon's own flip so a word typed across it corrects as a whole,
// while a REAL focus loss keeps the hard reset.

// TestActor_OwnFlipMixedWordCorrectsWhole pins the live failure signature
// of UAT phase 2 test 3 (matrix v1 word-mixed, run 2026-10-07) as FIXED:
// "gfb" typed in EN (transit), the daemon's own flip EN→RU (the stub
// switcher's SetGlobalEngine round trip returns nil), the synthetic
// lifecycle pair the live bus delivers, the rest typed in RU (six Cyrillic
// commits) — the double-tap correction then sees the WHOLE mixed token:
// exactly one RequireSurroundingText, one deletion over (-9,9) and the
// per-character inversion «паиghbdtn» — the re-pinned matrix row's
// semantics. The pre-amendment actor hard-resets the buffer on the
// synthetic FocusOut and corrects only the post-flip token: (-6,6) and
// «ghbdtn» — the live failure signature.
func TestActor_OwnFlipMixedWordCorrectsWhole(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	probe := &switchProbe{}
	a.SetSwitcher(probe.switcher)

	typeWord(a, "gfb") // EN part: transit, buffer fed as typed
	flipMode(a)        // EN → RU: flipTo runs, the stub switcher's round trip returns nil
	// The live bus answer to the daemon's own flip: ibus re-mints the
	// engine objects and the engine receives the synthetic pair.
	a.HandleLifecycle(engine.LifecycleFocusOut)
	a.HandleLifecycle(engine.LifecycleFocusIn)
	typeWord(a, wordEN) // RU part: commits "привет", buffer script-true

	if texts := sink.commitTexts(); len(texts) != 6 || texts[0]+texts[1]+texts[2]+texts[3]+texts[4]+texts[5] != wordRU {
		t.Fatalf("RU typing commits = %q, want the six runes of %q", texts, wordRU)
	}

	honestPush(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != 1 {
		t.Fatalf("mixed word across the own flip started verification %d times, want 1", got)
	}

	// The field holds the whole mixed token — exactly what the correction
	// must replace.
	field := "gfb" + wordRU
	a.HandleSurroundingText(field, runeLen(field), runeLen(field))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -9, nchars: 9}) {
		t.Fatalf("deletions = %+v, want exactly one (-9,9) — the whole mixed token; the synthetic"+
			" FocusOut of the own flip must not hard-reset the buffer", calls)
	}
	a.HandleSurroundingText("", 0, 0) // the deletion proof completes the ladder (plan 10-11)
	texts := sink.commitTexts()
	if len(texts) != 7 || texts[6] != mixedRU {
		t.Fatalf("correction commits = %q, want the last one to be [паиghbdtn]", texts)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done"`) {
		t.Errorf("INFO completion record missing; log:\n%s", buf.String())
	}
}

// TestActor_OwnFlipCreditConsumedOnce pins the over-suppression guard: the
// synthetic-lifecycle credit of the own flip is consumed EXACTLY once — a
// SECOND FocusOut after the synthetic one is a real context loss and must
// hard-reset the buffer (the double tap then yields the empty-buffer
// refusal, no verification round at all).
func TestActor_OwnFlipCreditConsumedOnce(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	probe := &switchProbe{}
	a.SetSwitcher(probe.switcher)

	flipMode(a)                                 // EN → RU: the successful round trip arms exactly one credit
	a.HandleLifecycle(engine.LifecycleFocusOut) // the synthetic pair's FocusOut consumes it
	a.HandleLifecycle(engine.LifecycleFocusIn)

	typeWord(a, wordEN) // a fresh word in the buffer
	// A SECOND FocusOut with the credit spent: a REAL focus loss.
	a.HandleLifecycle(engine.LifecycleFocusOut)

	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if !strings.Contains(buf.String(), `"reason":"empty-buffer"`) {
		t.Errorf("second FocusOut left a correctable token — a spent credit must not suppress a"+
			" real focus loss; log:\n%s", buf.String())
	}
	if got := sink.requireCount(); got != 0 {
		t.Errorf("sink saw %d RequireSurroundingText calls, want 0", got)
	}
}

// TestActor_OwnFlipFailedSwitcherArmsNothing pins the arming condition: a
// flip whose SetGlobalEngine round trip FAILED (the WARN form) delivers no
// engine re-mint, so no credit is armed — the next FocusOut is treated as
// a real focus loss and hard-resets the buffer.
func TestActor_OwnFlipFailedSwitcherArmsNothing(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.SetSwitcher(func(_ context.Context, _ string) error { return errSwitchInjected })

	flipMode(a) // EN → RU: the seam fails (WARN), no credit armed
	a.HandleLifecycle(engine.LifecycleFocusOut)

	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if !strings.Contains(buf.String(), `"reason":"empty-buffer"`) {
		t.Errorf("FocusOut after a failed-switcher flip left a correctable token — a failed"+
			" SetGlobalEngine must arm no credit; log:\n%s", buf.String())
	}
	if got := sink.requireCount(); got != 0 {
		t.Errorf("sink saw %d RequireSurroundingText calls, want 0", got)
	}
}

// TestActor_PhraseMixedCorrects pins D-26 on the sink (under the
// spec-delta of 2026-10-06): a phrase with words in different layouts —
// "ghbdtn" typed in EN, the flip, " привет" committed in RU — corrects
// through the SAME per-character inversion as the mixed word; there is no
// separate phrase-level script semantics. The triple tap deletes the whole
// phrase (-13,13) and commits "привет ghbdtn" — every letter flipped.
func TestActor_PhraseMixedCorrects(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	typeWord(a, wordEN) // EN word: transit, buffer fed as typed
	pressSpace(a)       // the phrase separator
	flipMode(a)         // EN → RU
	typeWord(a, wordEN) // RU word: commits "привет", buffer script-true

	honestPush(a)
	tapShift(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != 1 {
		t.Fatalf("mixed phrase started verification %d times, want 1", got)
	}

	field := wordEN + " " + wordRU
	a.HandleSurroundingText(field, runeLen(field), runeLen(field))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -13, nchars: 13}) {
		t.Fatalf("deletions = %+v, want exactly one (-13,13) — the whole mixed phrase range", calls)
	}
	a.HandleSurroundingText("", 0, 0) // the deletion proof completes the ladder (plan 10-11)
	texts := sink.commitTexts()
	if len(texts) != 7 || texts[6] != wordRU+" "+wordEN {
		t.Fatalf("correction commits = %q, want the last one to be [%s]", texts, wordRU+" "+wordEN)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done"`) {
		t.Errorf("INFO completion record missing; log:\n%s", buf.String())
	}
}

// Combo corpus (plan 03-04, SWCH-02/D-36): a press of the configured combo
// key under its configured HELD modifiers — the default Control_R under a
// held Shift, matched against Binding.ModMask with the key's own family
// bit cleared (the live wire truth: a press's state word carries only the
// modifiers held before the key) — corrects the last word through the
// double-tap pipeline and THEN flips the script mode, in that fixed order
// (D-36: the SPEC action name is "correct the word AND switch the
// layout"). The combo kills the tap series with a deliberate Reset
// (Pitfall 4: the pre-combo FSM silently swallowed the gesture as modifier
// use) and never feeds the buffer.

// pressComboDefault feeds the default combo shape: Shift held (press only,
// never released), then a Control_R press whose state word carries the HELD
// Shift plus the NumLock latch — the live wire truth (2026-09-15: a press
// carries only the modifiers held before the key; the key's own Control bit
// appears on its release) under the latch-tolerant mask compare of the
// 02-04 precedent.
func pressComboDefault(a *session.Actor) {
	const numLockLatch = 0x10 // IBUS_MOD2_MASK: latched NumLock (live-observed)
	a.HandleKey(engine.EngineEvent{Keyval: hotkey.KeyvalShiftR})
	a.HandleKey(engine.EngineEvent{
		Keyval: hotkey.KeyvalCtrlR,
		Mods:   engine.MaskShift | numLockLatch,
	})
}

// TestActor_ComboWordThenFlip pins the D-36 contract end to end on the fake
// sink: the combo launches the word pipeline (Require, nothing destructive,
// settle on the surrounding push: one delete over the token range and one
// commit of the converted word), the mode flip follows AFTER the settled
// completion record — and NO tap-series decision ever fires (the series
// died with the deliberate Reset, Pitfall 4).
func TestActor_ComboWordThenFlip(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	typeWord(a, wordEN)
	honestPush(a)
	pressComboDefault(a)

	if got := sink.requireCount(); got != 1 {
		t.Fatalf("combo require calls = %d, want 1 (the pre-correction round)", got)
	}
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Fatalf("two-phase violation: %d deletions before surrounding text, want 0", len(calls))
	}

	line := "abc " + wordEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -6, nchars: 6}) {
		t.Fatalf("deletions = %+v, want exactly one (-6,6) — the word range of the Double semantics", calls)
	}
	a.HandleSurroundingText("abc ", 4, 4) // the deletion proof completes the ladder (plan 10-11)
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != wordRU {
		t.Fatalf("commits = %q, want exactly one %q", texts, wordRU)
	}

	logged := buf.String()
	if !strings.Contains(logged, `"msg":"combo","kind":"word-layout"`) {
		t.Errorf("combo entry record missing; log:\n%s", logged)
	}
	if got := countActions(buf); got != 0 {
		t.Errorf("tap-series decisions after a combo = %d, want 0 — the series died with the Reset (Pitfall 4)", got)
	}
	// D-36 order pin: the flip's mode record lands strictly AFTER the word
	// pipeline's settled completion record.
	iDone := strings.Index(logged, `"msg":"correction","outcome":"done"`)
	iMode := strings.Index(logged, `"msg":"mode"`)
	if iDone < 0 || iMode < 0 {
		t.Fatalf("completion (%d) or mode (%d) record missing; log:\n%s", iDone, iMode, logged)
	}
	if iDone > iMode {
		t.Errorf("mode flip preceded the settled word correction (done@%d > mode@%d) — D-36 order", iDone, iMode)
	}
}

// TestActor_ComboEmptyBufferStillFlips pins the combo's discretion on an
// empty buffer: the word pipeline refuses with empty-buffer, and the flip
// STILL fires — switching is the primary intent when there is no word —
// strictly AFTER the refusal record (the D-36 order holds on both
// outcomes).
func TestActor_ComboEmptyBufferStillFlips(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	pressComboDefault(a)

	logged := buf.String()
	iRefusal := strings.Index(logged, `"reason":"empty-buffer"`)
	iMode := strings.Index(logged, `"msg":"mode"`)
	if iRefusal < 0 || iMode < 0 {
		t.Fatalf("empty-buffer refusal (%d) or mode flip (%d) missing; log:\n%s", iRefusal, iMode, logged)
	}
	if iRefusal > iMode {
		t.Errorf("flip preceded the word refusal (refusal@%d > mode@%d) — D-36 order violated", iRefusal, iMode)
	}
	if got := sink.requireCount(); got != 0 {
		t.Errorf("empty-buffer combo asked for surrounding text %d times, want 0", got)
	}
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Errorf("empty-buffer combo deleted %+v, want nothing", calls)
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Errorf("empty-buffer combo committed %q, want nothing", texts)
	}
}

// TestActor_ComboConfigurableBinding pins the D-31 renavigation of the
// combo: with word_layout_combo "alt+ctrl_l" fed through Options (the
// config string resolved by the same hotkey.ParseBinding the schema uses),
// the combo fires on a Ctrl_L press under Alt — and the DEFAULT
// Shift+Control_R shape no longer does.
func TestActor_ComboConfigurableBinding(t *testing.T) {
	t.Run("rebound combo fires", func(t *testing.T) {
		buf := captureLogs(t)
		binding, err := hotkey.ParseBinding("alt+ctrl_l")
		if err != nil {
			t.Fatalf("parse alt+ctrl_l: %v", err)
		}
		a, sink := wiredActor()
		a.SetOptions(session.Options{WordLayoutCombo: binding})

		typeWord(a, wordEN)
		a.HandleKey(engine.EngineEvent{Keyval: hotkey.KeyvalAltL}) // Alt held
		a.HandleKey(engine.EngineEvent{
			Keyval: hotkey.KeyvalCtrlL,
			Mods:   engine.MaskMod1, // held Alt only — the press never carries Ctrl_L's own bit
		})

		if !strings.Contains(buf.String(), `"msg":"combo","kind":"word-layout"`) {
			t.Errorf("rebound combo record missing; log:\n%s", buf.String())
		}
		if got := sink.requireCount(); got != 1 {
			t.Fatalf("rebound combo require calls = %d, want 1", got)
		}
		line := "abc " + wordEN
		a.HandleSurroundingText(line, runeLen(line), runeLen(line))
		if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != wordRU {
			t.Fatalf("rebound combo commits = %q, want one %q", texts, wordRU)
		}
	})

	t.Run("default shape no longer fires", func(t *testing.T) {
		buf := captureLogs(t)
		binding, err := hotkey.ParseBinding("alt+ctrl_l")
		if err != nil {
			t.Fatalf("parse alt+ctrl_l: %v", err)
		}
		a, sink := wiredActor()
		a.SetOptions(session.Options{WordLayoutCombo: binding})

		pressComboDefault(a)

		if strings.Contains(buf.String(), `"msg":"combo"`) {
			t.Errorf("default combo fired under a rebound binding; log:\n%s", buf.String())
		}
		if strings.Contains(buf.String(), `"msg":"mode"`) {
			t.Errorf("default combo flipped the mode under a rebound binding; log:\n%s", buf.String())
		}
		if got := sink.requireCount(); got != 0 {
			t.Errorf("default combo started the pipeline %d times under a rebound binding, want 0", got)
		}
	})
}

// TestActor_ComboDoesNotFeedBuffer pins the comboMask invariant against the
// new branch: a Super-modified letter and the combo key press itself put no
// rune in the field, so neither may feed the buffer — a double tap after
// them finds it empty (the combo fired on the empty buffer and flipped the
// mode; the empty-buffer skip below is the buffer oracle either way).
func TestActor_ComboDoesNotFeedBuffer(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4})
	a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4, Release: true})
	pressComboDefault(a) // the combo key press itself must not feed the buffer

	if got := sink.requireCount(); got != 0 {
		t.Fatalf("combo-key presses started verification %d times, want 0 — buffer stayed empty", got)
	}

	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if !strings.Contains(buf.String(), `"reason":"empty-buffer"`) {
		t.Errorf("empty-buffer record missing — a combo-path key fed the buffer; log:\n%s", buf.String())
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Errorf("commits = %q, want none — combo-path keys never print", texts)
	}
}

// Selection-correction corpus (plan 03-03, D-30): the range is the
// selection the client reported — [min(cursor,anchor), max(cursor,anchor))
// of the last surrounding push — in EITHER geometric direction (Pitfall 6),
// converted by the same run pipeline (D-23) and deleted EXACTLY over the
// range, never a rune more (CORR-07 precision applied to selections).

// TestActor_DoubleTapSelectionCorrects pins the selection branch of the
// Double decision (CORR-03, D-30): the field "ghbdtn привет" with the range
// [0,6) selected RIGHT-TO-LEFT (cursor=6, anchor=0 — the anchor wire pair of
// the ctrl+a class) corrects exactly the selected ghbdtn: one
// DeleteSurroundingText(-6,6) — the selection geometry, signed by the
// cursor's side — one CommitText("привет"), and the tail " привет" is NOT
// deleted. The buffer stays empty throughout: the selection range comes
// from the client push and takes precedence over the word path.
func TestActor_DoubleTapSelectionCorrects(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	field := wordEN + " " + wordRU
	a.HandleSurroundingText(field, 6, 0) // selection [0,6), right-to-left
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	// No DeleteSurroundingText on the selection path (2026-09-28 fix): the
	// commit itself replaces the ACTIVE selection — the delete+commit pair
	// double-applied in Chromium (live corruption hunt, verify timeout).
	calls := sink.deleteCalls()
	if len(calls) != 0 {
		t.Fatalf("deletions = %+v, want none — the commit replaces the active selection", calls)
	}
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != wordRU {
		t.Fatalf("commits = %q, want exactly one %q — only the selected range converts", texts, wordRU)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done"`) {
		t.Errorf("INFO completion record missing; log:\n%s", buf.String())
	}
	if got := sink.requireCount(); got != 1 {
		t.Errorf("require calls = %d, want exactly 1 (the verify-after round)", got)
	}
}

// TestActor_SelectionLeftToRight pins the positive-offset half of the
// selection geometry (Pitfall 6): the same range [0,6) reported
// LEFT-TO-RIGHT (cursor=0, anchor=6 — the GTE live shape of the spike)
// deletes with offset 0 — the range sits entirely right of the cursor — and
// nchars is still exactly the range length.
func TestActor_SelectionLeftToRight(t *testing.T) {
	a, sink := wiredActor()

	field := wordEN + " " + wordRU
	a.HandleSurroundingText(field, 0, 6) // selection [0,6), left-to-right
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	calls := sink.deleteCalls()
	if len(calls) != 0 {
		t.Fatalf("deletions = %+v, want none — the commit replaces the active selection", calls)
	}
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != wordRU {
		t.Fatalf("commits = %q, want exactly one %q", texts, wordRU)
	}
}

// TestActor_NoSelectionStillWord pins the D-30 continuity: a push with
// anchor == cursor (no selection observable) keeps the Double decision on
// the WORD path exactly as Phase 2 — the token geometry from the buffer,
// not the selection geometry.
func TestActor_NoSelectionStillWord(t *testing.T) {
	a, sink := wiredActor()

	typeWord(a, wordEN)
	line := "abc " + wordEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line)) // anchor == cursor
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -6, nchars: 6}) {
		t.Fatalf("deletions = %+v, want exactly one (-6,6) — the token geometry of Phase 2", calls)
	}
	a.HandleSurroundingText("abc ", 4, 4) // the deletion proof completes the ladder (plan 10-11)
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != wordRU {
		t.Fatalf("commits = %q, want exactly one %q", texts, wordRU)
	}
}

// TestActor_SelectionMixedConverts pins the mixed selection under the
// spec-delta of 2026-10-06: the selected mixed text "gfbпривет" inverts
// per character — gfb→паи, привет→ghbdtn — through the SAME ConvertRuns
// the word and phrase paths use.
func TestActor_SelectionMixedConverts(t *testing.T) {
	a, sink := wiredActor()

	field := "gfb" + wordRU
	a.HandleSurroundingText(field, runeLen(field), 0) // whole field selected
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	calls := sink.deleteCalls()
	if len(calls) != 0 {
		t.Fatalf("deletions = %+v, want none — the commit replaces the active selection", calls)
	}
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != mixedRU {
		t.Fatalf("commits = %q, want exactly one [%s] — the selection inverts per character", texts, mixedRU)
	}
}

// TestActor_SelectionNoLetters pins the D-20 refusal of the selection path:
// a selected range without Latin or Cyrillic letters has nothing to anchor
// on — the skip carries the no-letters reason and touches nothing.
func TestActor_SelectionNoLetters(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	a.HandleSurroundingText("2026", 4, 0) // digits only, selected
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if !strings.Contains(buf.String(), `"reason":"no-letters"`) {
		t.Errorf("no-letters record missing; log:\n%s", buf.String())
	}
	if got := sink.requireCount(); got != 0 {
		t.Errorf("letterless selection started verification %d times, want 0", got)
	}
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Errorf("letterless selection deleted %+v, want nothing", calls)
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Errorf("letterless selection committed %q, want nothing", texts)
	}
}

// Clipboard-rung corpus (plan 03-03, D-28/D-29): the rung is OPT-IN (zero
// clipboard subprocesses at the default configuration) and fires only on a
// verify-after mismatch, in the ADR-003 rung-C order — Save → Set(converted)
// → the Ctrl+V forward burst → best-effort Restore — with the clipboard
// content never reaching any log record.

// ownerClipStr is the fake user clipboard the rung must save and restore
// byte-exactly (D-29) — a const; the []byte conversions live at the use
// sites.
const ownerClipStr = "goswitch-e2e-owner-clip"

// errSimulated is the fake runner's failure sentinel (the err113
// discipline: no dynamically built error values).
var errSimulated = errors.New("simulated subprocess failure")

// forwardMark is the sequence-log tag of one forwarded key event (goconst).
const forwardMark = "forward"

// ctrlVKeycode and ctrlLKeycode are the physical (evdev) keycodes of the
// Ctrl+V forward burst — KEY_V=47, KEY_LEFTCTRL=29 (linux/event-codes.h),
// the same keycode discipline as backSpaceKeycode=14.
const (
	ctrlVKeycode = 47
	ctrlLKeycode = 29
)

// ctrlVWireCalls is the exact four-event forward sequence of the paste
// (the owner prototype's validated order): Control_L press bare, v press
// under Control, v release, Control_L release.
func ctrlVWireCalls() []forwardCall {
	return []forwardCall{
		{keyval: engine.KeyControlL, keycode: ctrlLKeycode, state: 0},
		{keyval: engine.KeyV, keycode: ctrlVKeycode, state: engine.MaskControl},
		{keyval: engine.KeyV, keycode: ctrlVKeycode, state: engine.MaskControl | engine.MaskRelease},
		{keyval: engine.KeyControlL, keycode: ctrlLKeycode, state: engine.MaskControl | engine.MaskRelease},
	}
}

// clipRunner is the recording wl-clipboard subprocess fake of the actor
// corpus: every invocation appends to one guarded sequence log AND fires
// the optional hook, so a test can interleave the subprocess records with
// the sink's forward hook and pin the rung's full order. reply is what
// wl-paste returns (the user's clipboard); failCopyFrom makes every wl-copy
// call FROM that 1-based ordinal fail — 2 fails exactly the restore (the
// set succeeded), the D-29 degradation shape.
type clipRunner struct {
	mu           sync.Mutex
	seq          []string
	reply        []byte
	failCopyFrom int
	copyCalls    int
	hook         func(line string)
}

// run records and answers one subprocess.
func (r *clipRunner) run(_ context.Context, name string, args []string, stdin []byte) ([]byte, error) {
	r.mu.Lock()
	line := fmt.Sprintf("%s %s <- %q", name, strings.Join(args, " "), stdin)
	r.seq = append(r.seq, line)
	hook := r.hook
	if name == "wl-copy" {
		r.copyCalls++
		if r.failCopyFrom > 0 && r.copyCalls >= r.failCopyFrom {
			r.mu.Unlock()

			return nil, fmt.Errorf("%s: %w", name, errSimulated)
		}
	}
	reply := r.reply
	r.mu.Unlock()
	if hook != nil {
		hook(line)
	}
	if name == "wl-paste" {
		return reply, nil
	}

	return nil, nil
}

// sequence snapshots the recorded subprocess log.
func (r *clipRunner) sequence() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.seq...)
}

// settledWordCorrection drives one settled level-1 word correction and
// delivers the verify-after MISMATCH verdict through the WR-05 debounced
// protocol — the state the rung keys on (the client ignored the deletion;
// the field still holds the pre-correction text on BOTH answers, so the
// debounce's retry also mismatches and the second push concludes).
func settledWordCorrection(a *session.Actor) {
	typeWord(a, wordEN)
	honestPush(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	line := "abc " + wordEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line)) // settles the pre-round
	// The two-phase completion (plan 10-11): the client's confirming push
	// (the deletion applied — the field is back to «abc ») completes the
	// ladder; the rung keys on the verify-after stale answers below.
	a.HandleSurroundingText("abc ", 4, 4)
	uncorrected := "abc " + wordEN
	// Stale answer 1: the debounce re-requires instead of verdicting.
	a.HandleSurroundingText(uncorrected, runeLen(uncorrected), runeLen(uncorrected))
	// Stale answer 2: the mismatch verdict — and the rung.
	a.HandleSurroundingText(uncorrected, runeLen(uncorrected), runeLen(uncorrected))
}

// TestActor_ClipboardRungDisabledByDefault pins the D-28 opt-in: at the
// default configuration (ClipboardRung false — the zero Options) a
// verify-after mismatch launches ZERO clipboard subprocesses — the rung
// must never touch the user's clipboard unconfigured.
func TestActor_ClipboardRungDisabledByDefault(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	runner := &clipRunner{reply: []byte(ownerClipStr)}
	a.UseClipboard(clipboard.New(clipboard.WithRunner(runner.run)))

	settledWordCorrection(a)

	if got := strings.Count(buf.String(), `"msg":"correction verify","outcome":"mismatch"`); got != 1 {
		t.Fatalf("mismatch counter = %d, want 1 (precondition of the pin); log:\n%s", got, buf.String())
	}
	if seq := runner.sequence(); len(seq) != 0 {
		t.Errorf("clipboard subprocesses at default config = %q, want none (D-28 opt-in)", seq)
	}
	if got := len(sink.forwardCalls()); got != 0 {
		t.Errorf("forward events = %d, want 0 — no paste may be replayed unconfigured", got)
	}
}

// TestActor_ClipboardRungAfterMismatch pins the rung's full contract
// (D-28/D-29): with ClipboardRung enabled a verify-after mismatch runs the
// ADR-003 rung-C sequence — wl-paste save, wl-copy set of the CONVERTED
// text, the four-event Ctrl+V forward burst, wl-copy restore of the saved
// bytes — in exactly that order (the interleaved sequence log), with the
// clipboard content never appearing in any INFO record. A restore failure
// is a WARN (D-29), never an operation error: the mismatch counter stays at
// one and the sequence completes.
func TestActor_ClipboardRungAfterMismatch(t *testing.T) {
	t.Run("full sequence in order, no content in INFO", func(t *testing.T) {
		rungFullSequenceInOrder(t)
	})

	t.Run("restore failure is a WARN, not an error", func(t *testing.T) {
		rungRestoreFailureIsWarn(t)
	})
}

// rungFullSequenceInOrder pins the rung's contract on one guarded sequence
// log across BOTH recorders (subprocesses numbered in arrival order — the
// seq assertions name which is which — interleaved with the forward events)
// plus the exact Ctrl+V wire shape and the INFO content prohibition.
func rungFullSequenceInOrder(t *testing.T) {
	t.Helper()
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.SetOptions(session.Options{ClipboardRung: true})
	runner := &clipRunner{reply: []byte(ownerClipStr)}
	a.UseClipboard(clipboard.New(clipboard.WithRunner(runner.run)))

	var seqMu sync.Mutex
	full := []string{}
	record := func(line string) {
		seqMu.Lock()
		defer seqMu.Unlock()
		full = append(full, line)
	}
	clipCalls := 0
	sink.forwardHook = func(forwardCall) { record(forwardMark) }
	runner.hook = func(string) {
		clipCalls++
		record(fmt.Sprintf("clip:%d", clipCalls))
	}

	settledWordCorrection(a)

	want := []string{
		"clip:1", "clip:2", forwardMark, forwardMark, forwardMark, forwardMark, "clip:3",
	}
	if !slices.Equal(full, want) {
		t.Fatalf("rung sequence = %q, want %q (save → set → Ctrl+V burst → restore)", full, want)
	}
	seq := runner.sequence()
	if len(seq) != 3 {
		t.Fatalf("clipboard subprocesses = %q, want exactly 3 (save, set, restore)", seq)
	}
	if !strings.Contains(seq[0], "wl-paste --no-newline") {
		t.Errorf("first subprocess = %q, want the wl-paste save", seq[0])
	}
	if !strings.Contains(seq[1], "wl-copy --trim-newline") || !strings.Contains(seq[1], wordRU) {
		t.Errorf("set subprocess = %q, want wl-copy of the converted %q through stdin", seq[1], wordRU)
	}
	if !strings.Contains(seq[2], "wl-copy --trim-newline") || !strings.Contains(seq[2], ownerClipStr) {
		t.Errorf("restore subprocess = %q, want wl-copy of the saved owner bytes", seq[2])
	}
	forwards := sink.forwardCalls()
	if !slices.Equal(forwards, ctrlVWireCalls()) {
		t.Fatalf("forward events = %+v, want the exact Ctrl+V burst %+v", forwards, ctrlVWireCalls())
	}
	assertNoClipContentInINFO(t, buf.String())
}

// rungRestoreFailureIsWarn pins the D-29 degradation: the restore's wl-copy
// fails AFTER the paste fired — a WARN lands, the mismatch verdict stays at
// one, and the burst already delivered its four events.
func rungRestoreFailureIsWarn(t *testing.T) {
	t.Helper()
	// INFO verdict + WARN degradation in one stream.
	buf := captureLogsLevel(t, slog.LevelDebug)
	a, sink := wiredActor()
	a.SetOptions(session.Options{ClipboardRung: true})
	runner := &clipRunner{reply: []byte(ownerClipStr), failCopyFrom: 2} // set ok, restore fails
	a.UseClipboard(clipboard.New(clipboard.WithRunner(runner.run)))

	settledWordCorrection(a)

	if !strings.Contains(buf.String(), "clipboard restore failed") {
		t.Errorf("restore-failure WARN missing; log:\n%s", buf.String())
	}
	if got := strings.Count(buf.String(), `"msg":"correction verify","outcome":"mismatch"`); got != 1 {
		t.Errorf("mismatch counter = %d, want 1 — the rung degrades, never re-fires the verdict", got)
	}
	// The paste already fired before the failed restore: the full burst.
	if got := len(sink.forwardCalls()); got != len(ctrlVWireCalls()) {
		t.Errorf("forward events = %d, want %d — the degradation costs only the restore", got, len(ctrlVWireCalls()))
	}
}

// assertNoClipContentInINFO pins the D-20/D-21 extension over the rung: no
// INFO record carries clipboard content — neither the replacement nor the
// owner bytes.
func assertNoClipContentInINFO(t *testing.T, logged string) {
	t.Helper()
	for _, line := range strings.Split(logged, "\n") {
		if !strings.Contains(line, `"level":"INFO"`) {
			continue
		}
		if strings.Contains(line, wordRU) || strings.Contains(line, ownerClipStr) {
			t.Errorf("INFO record leaks clipboard content: %s", line)
		}
	}
}

// TestActor_ClipboardRungOffKeyPath pins WR-01: the rung's subprocess
// sequence NEVER holds the actor mutex — while a wedged wl-paste parks the
// rung, keystrokes keep deciding through the FSM, and a SECOND mismatch
// concluding in that window skips with a busy reason instead of blocking
// or interleaving a second wl-copy/wl-paste pair into the first round-trip.
func TestActor_ClipboardRungOffKeyPath(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.SetOptions(session.Options{ClipboardRung: true})
	release := make(chan struct{})
	saveStarted := make(chan struct{})
	var saveOnce sync.Once
	a.UseClipboard(clipboard.New(clipboard.WithRunner(
		func(_ context.Context, name string, _ []string, _ []byte) ([]byte, error) {
			if name != "wl-paste" {
				return nil, nil
			}
			saveOnce.Do(func() { close(saveStarted) })
			<-release // the wedged wl-paste: parks until the test frees it

			return []byte(ownerClipStr), nil
		})))

	rungDone := make(chan struct{})
	go func() {
		defer close(rungDone)
		settledWordCorrection(a) // ends in the mismatch push → the rung parks in Save
	}()
	select {
	case <-saveStarted:
	case <-time.After(reloadWait):
		t.Fatalf("the rung never started its save; log:\n%s", buf.String())
	}

	// A key event must complete while the rung is parked — the daemon's
	// <50 ms key contract cannot queue behind three 1.5 s subprocesses.
	// (A modifier pair: the same HandleKey → mutex → FSM path as any
	// keystroke, without polluting the buffer the next round settles.)
	keyDone := make(chan struct{})
	go func() {
		defer close(keyDone)
		tapKeySeries(a, hotkey.KeyvalAltL)
	}()
	select {
	case <-keyDone:
	case <-time.After(reloadWait):
		t.Fatalf("a key event blocked behind the parked rung — the rung must not hold the actor mutex; log:\n%s",
			buf.String())
	}

	// A SECOND full correction settling into a mismatch while the first
	// rung is parked must conclude with the busy skip — never block, never
	// paste a second payload. A Reset first: the executed round replaced
	// the buffer's token, and the fresh round needs a fresh word.
	a.HandleLifecycle(engine.LifecycleReset)
	settledWordCorrection(a)
	if !strings.Contains(buf.String(), `"reason":"clipboard-rung-busy"`) {
		t.Errorf("the concurrent rung attempt did not skip busy; log:\n%s", buf.String())
	}
	if got := len(sink.forwardCalls()); got != 0 {
		t.Errorf("forward events while parked = %d, want 0 — the busy rung must not paste", got)
	}

	close(release)
	select {
	case <-rungDone:
	case <-time.After(reloadWait):
		t.Fatalf("the parked rung never completed after the release; log:\n%s", buf.String())
	}
	// The one live rung finished its full contract: exactly one Ctrl+V
	// burst and the mismatch verdict counted once.
	if got := len(sink.forwardCalls()); got != len(ctrlVWireCalls()) {
		t.Errorf("forward events after the release = %d, want the single full burst %d", got, len(ctrlVWireCalls()))
	}
	if got := strings.Count(buf.String(), `"msg":"correction verify","outcome":"mismatch"`); got != 2 {
		t.Errorf("mismatch verdicts = %d, want 2 (one per settled round); log:\n%s", got, buf.String())
	}
}

// TestActor_KeyEventsFeedFSM pins the single-tap path: a clean Shift_R
// press/release pair produces exactly one decision record, n=1, at window
// expiry — never inside HandleKey.
func TestActor_KeyEventsFeedFSM(t *testing.T) {
	buf := captureLogs(t)
	a := session.NewActor(farWindow)

	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	logged := buf.String()
	if got := strings.Count(logged, `"msg":"action"`); got != 1 {
		t.Fatalf("action records after single tap = %d, want 1; log:\n%s", got, logged)
	}
	if !strings.Contains(logged, `"n":1`) {
		t.Errorf("single-tap decision n=1 missing; log:\n%s", logged)
	}
}

// TestActor_DoubleTriple pins series counting: two and three taps each
// produce their decision exactly once, and a second expiry is a no-op.
func TestActor_DoubleTriple(t *testing.T) {
	for _, tc := range []struct {
		taps int
		want string
	}{
		{taps: 2, want: `"n":2`},
		{taps: 3, want: `"n":3`},
	} {
		buf := captureLogs(t)
		a := session.NewActor(farWindow)

		for range tc.taps {
			tapShift(a)
		}
		a.ExpiryAt(expiryAfterWindow)
		a.ExpiryAt(3 * farWindow) // stale second expiry must not re-fire

		logged := buf.String()
		if got := countActions(buf); got != 1 {
			t.Fatalf("%d taps: action records = %d, want 1; log:\n%s", tc.taps, got, logged)
		}
		if !strings.Contains(logged, tc.want) {
			t.Errorf("%d taps: decision %s missing; log:\n%s", tc.taps, tc.want, logged)
		}
	}
}

// TestActor_FocusOutDisarms pins the disarm path: the same tap series that
// fires without interruption stays silent when FocusOut lands mid-series.
func TestActor_FocusOutDisarms(t *testing.T) {
	// Positive control first: without FocusOut the double tap fires n=2.
	buf := captureLogs(t)
	hot := session.NewActor(farWindow)
	tapShift(hot)
	tapShift(hot)
	hot.ExpiryAt(expiryAfterWindow)
	if got := countActions(buf); got != 1 || !strings.Contains(buf.String(), `"n":2`) {
		t.Fatalf("positive control: expected one n=2 action, got %d; log:\n%s", got, buf.String())
	}

	// FocusOut mid-series: no decision, timer disarmed.
	buf = captureLogs(t)
	cold := session.NewActor(farWindow)
	tapShift(cold)
	tapShift(cold)
	cold.HandleLifecycle(engine.LifecycleFocusOut)
	cold.ExpiryAt(expiryAfterWindow)

	if got := countActions(buf); got != 0 {
		t.Fatalf("actions after FocusOut = %d, want 0; log:\n%s", got, buf.String())
	}
}

// TestActor_Serialization pins the single entry point: a burst of Shift_R
// taps from several goroutines loses no series (exactly one decision fires,
// n between 1 and 3) and runs clean under -race.
func TestActor_Serialization(t *testing.T) {
	buf := captureLogs(t)
	a := session.NewActor(farWindow)

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 25 {
				tapShift(a)
			}
		}()
	}
	wg.Wait()
	a.ExpiryAt(expiryAfterWindow)

	logged := buf.String()
	if got := countActions(buf); got != 1 {
		t.Fatalf("actions after parallel burst = %d, want exactly 1; log:\n%s", got, logged)
	}
	for _, want := range []string{`"n":1`, `"n":2`, `"n":3`} {
		if strings.Contains(logged, want) {
			return
		}
	}
	t.Errorf("burst decision n in 1..3 missing; log:\n%s", logged)
}

// TestActor_WindowTimerFiresAutomatically pins the real-timer wiring: with a
// short window the AfterFunc path — not the test — delivers the expiry and
// the decision appears in the log on its own.
func TestActor_WindowTimerFiresAutomatically(t *testing.T) {
	buf := captureLogs(t)
	a := session.NewActor(8 * time.Millisecond)

	tapShift(a)
	tapShift(a)

	deadline := time.Now().Add(2 * time.Second)
	for countActions(buf) < 1 && time.Now().Before(deadline) {
		time.Sleep(2 * time.Millisecond)
	}

	logged := buf.String()
	if got := countActions(buf); got != 1 {
		t.Fatalf("timer-fired actions = %d, want 1; log:\n%s", got, logged)
	}
	if !strings.Contains(logged, `"n":2`) {
		t.Errorf("timer-fired decision n=2 missing; log:\n%s", logged)
	}
}

// Hot-reload timing constants: the corpus drives the actor's REAL AfterFunc
// wiring (the reload lands between armings), so budgets are wall-clock.
const (
	// reloadWait budgets one real-timer window decision.
	reloadWait = 2 * time.Second
	// newWindowWait budgets the new-window series decision.
	newWindowWait = 2 * time.Second
	// newWindowCeiling sits below the OLD 300 ms window: a decision inside
	// it proves the 150 ms reload window armed the new series.
	newWindowCeiling = 250 * time.Millisecond
)

// Hot-reload corpus (plan 03-04, CONF-02/Pitfall 8): the actor consumes
// config snapshots live — ONE Snapshot() read per event, the window
// reaching the ARMING of new series (an already-armed timer keeps its own
// deadline; a reload never re-arms or cancels it), the options and the
// combo binding picked up by the next event after a change.

// reloadSource is the fake config source of the hot-reload corpus: a
// guarded config.Config standing in for the watcher's Snapshot() contract
// (03-02 pinned the production last-good behavior; what is under test here
// is the CONSUMPTION). calls counts Snapshot() invocations — the
// one-read-per-event pin.
type reloadSource struct {
	mu    sync.Mutex
	cfg   config.Config
	calls int
}

// Snapshot records the read and returns the current document.
func (s *reloadSource) Snapshot() config.Config {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls++

	return s.cfg
}

// set replaces the served document (the watcher's valid-reload step).
func (s *reloadSource) set(cfg config.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cfg = cfg
}

// count snapshots the read counter.
func (s *reloadSource) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.calls
}

// reloadCfg builds a complete valid document with the given window and
// combo name over Defaults() — the same shape a validated -config file
// yields.
func reloadCfg(tapWindowMs int, combo string) config.Config {
	cfg := config.Defaults()
	cfg.Timeouts.TapWindowMs = tapWindowMs
	cfg.Hotkeys.WordLayoutCombo = combo

	return cfg
}

// waitActionsUntil polls the captured log until it holds want action
// records or the deadline passes (the real-timer polling idiom of
// TestActor_WindowTimerFiresAutomatically — the reload corpus drives the
// actor's own AfterFunc wiring, so expiry is NOT injected).
func waitActionsUntil(t *testing.T, buf *syncBuffer, want int, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for {
		if countActions(buf) >= want {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// waitLogUntil polls the captured log until it holds want occurrences of
// mark or the deadline passes — waitActionsUntil generalized off the action
// counter for the real-timer records that are not decisions.
func waitLogUntil(t *testing.T, buf *syncBuffer, mark string, want int, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for {
		if strings.Count(buf.String(), mark) >= want {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// TestActor_HotReloadWindowNewSeries pins the Pitfall-8 window rule: a
// series armed BEFORE a reload expires by the window it was armed with
// (the reload neither cancels nor loses the armed decision), and a series
// armed AFTER the reload expires by the NEW window — the snapshot's
// tap_window_ms reaches both the actor's timer arming and the FSM's gap
// and expiry checks.
func TestActor_HotReloadWindowNewSeries(t *testing.T) {
	t.Run("armed series survives the reload, exactly one decision", func(t *testing.T) {
		buf := captureLogs(t)
		a := session.NewActor(farWindow)
		src := &reloadSource{cfg: reloadCfg(300, "shift+ctrl_r")}
		a.AttachConfig(src)

		tapShift(a)
		src.set(reloadCfg(150, "shift+ctrl_r")) // reload lands mid-window

		if !waitActionsUntil(t, buf, 1, reloadWait) {
			t.Fatalf("the armed series never decided after the reload; log:\n%s", buf.String())
		}
		if got := countActions(buf); got != 1 {
			t.Errorf("actions after reload = %d, want exactly 1 (no re-arm, no loss)", got)
		}
	})

	t.Run("new series uses the new window", func(t *testing.T) {
		buf := captureLogs(t)
		a := session.NewActor(farWindow)
		src := &reloadSource{cfg: reloadCfg(300, "shift+ctrl_r")}
		a.AttachConfig(src)

		tapShift(a)
		if !waitActionsUntil(t, buf, 1, reloadWait) {
			t.Fatalf("first series (old window) never decided; log:\n%s", buf.String())
		}

		src.set(reloadCfg(150, "shift+ctrl_r"))
		t0 := time.Now()
		tapShift(a)
		// The constructor window is farWindow (1h): only the per-event
		// snapshot consumption can arm the 150 ms decision — and it must
		// arrive well inside the old 300 ms window.
		if !waitActionsUntil(t, buf, 2, newWindowWait) {
			t.Fatalf("new series never decided; log:\n%s", buf.String())
		}
		if d := time.Since(t0); d > newWindowCeiling {
			t.Errorf("new-series decision took %v, want < %v — the 150 ms window governs",
				d, newWindowCeiling)
		}
	})
}

// TestActor_HotReloadOptionsAndCombo pins the non-window consumption: the
// Backspace cap, the clipboard-rung switch and the combo binding of a
// changed snapshot are picked up by the NEXT event — SetOptions remains
// the no-source surface, but an attached source has priority.
func TestActor_HotReloadOptionsAndCombo(t *testing.T) {
	t.Run("backspace cap reaches the next BuildPlan", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActorCaps(engine.CapPreeditText) // level 2: the cap bites
		src := &reloadSource{cfg: reloadCfg(300, "shift+ctrl_r")}
		src.cfg.Correction.BackspaceCap = 3
		a.AttachConfig(src)

		typeWord(a, wordEN) // 6 runes > cap 3
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
		if !strings.Contains(buf.String(), `"reason":"backspace-cap"`) {
			t.Fatalf("over-cap burst was not refused; log:\n%s", buf.String())
		}
		if got := len(sink.forwardCalls()); got != 0 {
			t.Errorf("over-cap burst replayed %d Backspaces, want 0", got)
		}

		over := config.Defaults()
		over.Timeouts.TapWindowMs = 300
		over.Correction.BackspaceCap = 50
		src.set(over)
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
		if got := len(sink.forwardCalls()); got != 2*len([]rune(wordEN)) {
			t.Errorf("post-reload burst replayed %d forwards, want %d (pairs for %d Backspaces) — "+
				"the new cap governs", got, 2*len([]rune(wordEN)), len([]rune(wordEN)))
		}
	})

	t.Run("combo binding follows the snapshot", func(t *testing.T) {
		a, sink := wiredActor()
		src := &reloadSource{cfg: reloadCfg(300, "alt+ctrl_l")}
		a.AttachConfig(src)

		typeWord(a, wordEN)
		a.HandleKey(engine.EngineEvent{Keyval: hotkey.KeyvalAltL}) // Alt held
		a.HandleKey(engine.EngineEvent{Keyval: hotkey.KeyvalCtrlL, Mods: engine.MaskMod1})
		if got := sink.requireCount(); got != 1 {
			t.Fatalf("rebound combo require calls = %d, want 1", got)
		}
		line := "abc " + wordEN
		a.HandleSurroundingText(line, runeLen(line), runeLen(line))
		if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != wordRU {
			t.Fatalf("rebound combo commits = %q, want one %q", texts, wordRU)
		}

		// Reload back to the default binding: the alt shape must die with
		// the old snapshot, the default Shift+Control_R shape must fire.
		// Require budget: 2 spent by the first correction (pre-round +
		// verify-after), 1 more by the rebound combo's pre-round.
		src.set(reloadCfg(300, "shift+ctrl_r"))
		a.HandleKey(engine.EngineEvent{Keyval: hotkey.KeyvalShiftR})
		a.HandleKey(engine.EngineEvent{Keyval: hotkey.KeyvalCtrlR, Mods: engine.MaskShift})
		if got := sink.requireCount(); got != 3 {
			t.Fatalf("default-binding combo require calls = %d, want 3 — the rebind follows the snapshot", got)
		}
	})
}

// TestActor_HotReloadTapKey pins the CR-01 wiring: the document's
// hotkeys.tap_key reaches every consumer that makes it live — the FSM's
// series key, HandleKey's timer-arming release check, and the reload seam
// (a changed name re-resolves on the next event, the comboName precedent).
// While the document names shift_l, a Shift_L series decides through the
// REAL timer (expiry never injected) and Shift_R is a plain modifier; a
// reload back to shift_r swaps the shapes.
func TestActor_HotReloadTapKey(t *testing.T) {
	t.Run("the configured key's series decides through the real timer", func(t *testing.T) {
		buf := captureLogs(t)
		a := session.NewActor(farWindow)
		cfg := reloadCfg(150, "shift+ctrl_r")
		cfg.Hotkeys.TapKey = "shift_l"
		src := &reloadSource{cfg: cfg}
		a.AttachConfig(src)

		t0 := time.Now()
		tapKeySeries(a, hotkey.KeyvalShiftL)
		tapKeySeries(a, hotkey.KeyvalShiftL)
		// Only the release of the CONFIGURED key arming the actor timer can
		// decide here: the constructor window is farWindow (1 h) and expiry
		// is never injected — and the decision must land inside the old
		// 300 ms default too, proving the snapshot's 150 ms window armed it.
		if !waitActionsUntil(t, buf, 1, reloadWait) {
			t.Fatalf("the configured tap key's series never decided; log:\n%s", buf.String())
		}
		if d := time.Since(t0); d > newWindowCeiling {
			t.Errorf("configured-key decision took %v, want < %v — the snapshot's 150 ms window governs",
				d, newWindowCeiling)
		}
		if got := countActions(buf); got != 1 {
			t.Errorf("actions = %d, want exactly 1 (double tap)", got)
		}
	})

	t.Run("the reload swaps the deciding key", func(t *testing.T) {
		buf := captureLogs(t)
		a := session.NewActor(farWindow)
		src := &reloadSource{cfg: reloadCfg(300, "shift+ctrl_r")}
		a.AttachConfig(src)

		tapKeySeries(a, hotkey.KeyvalShiftL) // not the document's key
		a.ExpiryAt(expiryAfterWindow)
		if got := countActions(buf); got != 0 {
			t.Fatalf("Shift_L series decided under a shift_r document; log:\n%s", buf.String())
		}

		cfgL := reloadCfg(300, "shift+ctrl_r")
		cfgL.Hotkeys.TapKey = "shift_l"
		src.set(cfgL) // reload: shift_l now decides
		tapKeySeries(a, hotkey.KeyvalShiftL)
		a.ExpiryAt(expiryAfterWindow)
		if got := countActions(buf); got != 1 {
			t.Fatalf("Shift_L series after the tap-key reload: actions = %d, want 1; log:\n%s", got, buf.String())
		}

		src.set(reloadCfg(300, "shift+ctrl_r")) // back to the default key
		tapKeySeries(a, hotkey.KeyvalShiftL)
		a.ExpiryAt(expiryAfterWindow)
		if got := countActions(buf); got != 1 {
			t.Fatalf("the replaced key still decides after the reload back: actions = %d, want 1; log:\n%s",
				got, buf.String())
		}
	})
}

// TestActor_HotReloadVerifyWait pins the CR-02 wiring: the document's
// timeouts.verify_wait_ms is the actor's verify budget — the pending-fix
// deadline and both verify-after deadlines run on it, and the snapshot's
// value changes it live. The pre-round miss arms the deadline at the
// RELOADED budget: the verify-timeout skip must not land inside the old
// 100 ms default, and must land by the new budget.
func TestActor_HotReloadVerifyWait(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	cfg := reloadCfg(300, "shift+ctrl_r")
	cfg.Timeouts.VerifyWaitMs = 1200
	src := &reloadSource{cfg: cfg}
	a.AttachConfig(src)

	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	// No push ever arrives: the Double's pre-round miss armed the pending
	// deadline, and only VerifyExpiry can close it (the real AfterFunc).
	const timeoutMark = `"reason":"verify-timeout"`
	if waitLogUntil(t, buf, timeoutMark, 1, 400*time.Millisecond) {
		t.Fatalf("verify-timeout landed inside the old 100 ms budget"+
			" — the snapshot's verify_wait_ms is not live; log:\n%s", buf.String())
	}
	if !waitLogUntil(t, buf, timeoutMark, 1, reloadWait) {
		t.Fatalf("the reloaded 1200 ms budget never closed the round; log:\n%s", buf.String())
	}
}

// TestActor_HotReloadInvalidKeepsLastGood pins the consumption side of the
// D-32 last-good contract: while the source keeps serving the last valid
// document (the watcher's invalid-reload behavior, pinned in 03-02), the
// actor's behavior is unchanged — and every EVENT reads the snapshot
// exactly once (the Pattern-2 rule: a value per event, never a pointer
// held across the mutex).
func TestActor_HotReloadInvalidKeepsLastGood(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	src := &reloadSource{cfg: reloadCfg(300, "shift+ctrl_r")}
	a.AttachConfig(src)

	typeWord(a, wordEN)
	if got, want := src.count(), 2*len(wordEN); got != want {
		t.Errorf("Snapshot() reads after typing %d runes = %d, want %d (one per key event)", len(wordEN), got, want)
	}

	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText(wordEN, runeLen(wordEN), runeLen(wordEN))
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done"`) {
		t.Errorf("correction under a last-good source did not complete; log:\n%s", buf.String())
	}
	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != wordRU {
		t.Fatalf("commits = %q, want one %q — the served document is the behavior", texts, wordRU)
	}
}

// MACR corpus (plan 03-05 Task 2, MACR-01/ADR-005): the Super→Ctrl remap
// layer. The interception branch sits in feedKey ABOVE every mode branch
// (Pattern 6): a configured letter press carrying Mod4 is consumed and
// replayed as the owner prototype's Ctrl+letter forward burst — in ANY
// internal mode, feeding no rune to the buffer, printing nothing. The
// consumed-upstream detect (ADR-005 b.2) keys on the live-proven wire
// shape: the Super press arrives bare, its release still carries Mod4, and
// a shell-consumed chord delivers NO letter press at all between them.

// The physical keycodes of the MACR burst's modifier and letter halves
// (/usr/include/linux/input-event-codes.h — KEY_LEFTCTRL, KEY_RIGHTCTRL,
// KEY_A, KEY_B; the same keycode discipline as backSpaceKeycode).
const (
	macrCtrlLKeycode = 29
	macrCtrlRKeycode = 97
	macrKeyAKeycode  = 30
)

// enableMACR feeds the MACR options surface with the given letter set —
// the config document's macr section (03-02 schema) through the same
// SetOptions path the combo corpus uses.
func enableMACR(a *session.Actor, letters string) {
	set := make(map[rune]bool)
	for _, r := range letters {
		set[r] = true
	}
	a.SetOptions(session.Options{MACREnabled: true, MACRLetters: set})
}

// releaseSuper feeds the Super release the wire delivers: still carrying
// Mod4 — the family-bit truth of the modifiers (live trace 2026-09-15:
// 0xffeb release with mods 0x50). The pending remap rides this event.
func releaseSuper(a *session.Actor) {
	a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL, Mods: engine.MaskMod4, Release: true})
}

// TestActor_MACRIntercepts pins the interception itself (MACR-01,
// ADR-005): super+a (a press carrying Mod4 — the press-side wire truth of
// the live trace) is CONSUMED and its remap rides the SUPER RELEASE: the
// Ctrl+letter burst fires only once the physical Super left the keyboard,
// because a client still holding Super reads the synthetic chord as
// Ctrl+Super and drops the binding (live finding 2026-09-15 — the first
// implementation burst at the press and the live cut never landed).
// Nothing prints (the branch sits above the mode branches); the INFO
// record names the config letter only (D-20); the counter grows at the
// press. The alt_modifier subtest pins the b.3 mechanics (empty default =
// Control_L; a configured name swaps the modifier key only).
func TestActor_MACRIntercepts(t *testing.T) {
	t.Run("super+a consumed, the Ctrl+a burst rides the Super release", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()
		enableMACR(a, "a")

		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
		consumed := a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4})
		if !consumed {
			t.Fatalf("super+a press was not consumed — the MACR branch must own it (MACR-01)")
		}
		if got := len(sink.forwardCalls()); got != 0 {
			t.Fatalf("burst before the Super release = %d events, want 0 — the client still holds Super", got)
		}
		if got := a.MACRCounters(); got.SuperIntercepted != 1 {
			t.Errorf("SuperIntercepted at the press = %d, want 1 (the counter is the press's)", got.SuperIntercepted)
		}
		releaseSuper(a)

		want := []forwardCall{
			{keyval: engine.KeyControlL, keycode: macrCtrlLKeycode, state: 0},
			{keyval: uint32('a'), keycode: macrKeyAKeycode, state: engine.MaskControl},
			{keyval: uint32('a'), keycode: macrKeyAKeycode, state: engine.MaskControl | engine.MaskRelease},
			{keyval: engine.KeyControlL, keycode: macrCtrlLKeycode, state: engine.MaskControl | engine.MaskRelease},
		}
		if got := sink.forwardCalls(); !slices.Equal(got, want) {
			t.Fatalf("forward burst = %+v, want exactly the prototype Ctrl+letter sequence %+v", got, want)
		}
		if texts := sink.commitTexts(); len(texts) != 0 {
			t.Fatalf("MACR interception committed %q — the branch sits ABOVE the mode branches, nothing may print",
				texts)
		}
		if !strings.Contains(buf.String(), `"msg":"super intercept","key":"a"`) {
			t.Errorf("interception INFO record missing (config letter + counters only, D-20); log:\n%s", buf.String())
		}
		if got := a.MACRCounters(); got.SuperIntercepted != 1 || got.ConsumedUpstream != 0 {
			t.Errorf("counters = %+v, want {SuperIntercepted:1 ConsumedUpstream:0}", got)
		}
	})
}

// TestActor_MACRAltModifier pins the b.3 mechanics (ADR-005): the
// alternative modifier is NOT introduced by default (empty = Control_L,
// pinned by TestActor_MACRIntercepts' plain burst); a configured name
// swaps the burst's modifier key only — the four-event shape never
// changes.
func TestActor_MACRAltModifier(t *testing.T) {
	a, sink := wiredActor()
	a.SetOptions(session.Options{
		MACREnabled:     true,
		MACRLetters:     map[rune]bool{'a': true},
		MACRAltModifier: "ctrl_r",
	})

	a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
	_ = a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4})
	releaseSuper(a)

	calls := sink.forwardCalls()
	if len(calls) != 4 {
		t.Fatalf("alt-modifier burst = %d events, want the 4-event shape", len(calls))
	}
	if calls[0].keyval != engine.KeyControlR || calls[3].keyval != engine.KeyControlR {
		t.Fatalf("alt-modifier burst ends = %x/%x, want Control_R (0xffe4) at both ends",
			calls[0].keyval, calls[3].keyval)
	}
	if calls[0].keycode != macrCtrlRKeycode {
		t.Errorf("Control_R keycode = %d, want %d (KEY_RIGHTCTRL)", calls[0].keycode, macrCtrlRKeycode)
	}
}

// TestActor_MACRConfigSnapshot pins the CONF-02 consumption of the macr
// section: an attached source folds enabled/letters into the options —
// the same one-snapshot-per-event rule as the combo and the caps.
func TestActor_MACRConfigSnapshot(t *testing.T) {
	a, sink := wiredActor()
	cfg := config.Defaults()
	cfg.MACR.Enabled = true
	cfg.MACR.Letters = "a"
	a.AttachConfig(&reloadSource{cfg: cfg})

	a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
	if consumed := a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4}); !consumed {
		t.Fatalf("super+a under a config snapshot was not consumed — the macr section must fold like the rest")
	}
	releaseSuper(a)
	if got := len(sink.forwardCalls()); got != 4 {
		t.Fatalf("snapshot-fed interception forwarded %d events, want the 4-event burst", got)
	}
}

// TestActor_MACRDisabledByDefault pins the off-by-default contract: the
// zero Options (the no-config daemon) never touches a Super chord — no
// burst, no consumed-upstream WARN (the native Super use must not spam),
// and even under an enabled layer only the configured letter set is
// intercepted.
func TestActor_MACRDisabledByDefault(t *testing.T) {
	t.Run("off: super+a transits, no burst, no warn", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()

		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
		if consumed := a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4}); consumed {
			t.Fatalf("super+a consumed with MACR off — the default layer must never touch the keyboard")
		}
		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL, Mods: engine.MaskMod4, Release: true})

		if got := len(sink.forwardCalls()); got != 0 {
			t.Errorf("MACR-off daemon forwarded %+v, want none", got)
		}
		if strings.Contains(buf.String(), `"msg":"super combo skipped"`) {
			t.Errorf("consumed-upstream WARN with MACR off — spam on every native Super use; log:\n%s", buf.String())
		}
		if got := a.MACRCounters(); got.SuperIntercepted != 0 || got.ConsumedUpstream != 0 {
			t.Errorf("counters = %+v, want the zero value with the layer off", got)
		}
	})

	t.Run("letter outside the set transits under enabled", func(t *testing.T) {
		a, sink := wiredActor()
		enableMACR(a, "a")

		if consumed := a.HandleKey(engine.EngineEvent{Keyval: uint32('b'), Mods: engine.MaskMod4}); consumed {
			t.Fatalf("super+b consumed with letters \"a\" — only the configured set is intercepted (ADR-005 b)")
		}
		if got := len(sink.forwardCalls()); got != 0 {
			t.Errorf("out-of-set letter forwarded %+v, want none", got)
		}
	})
}

// TestActor_MACRBufferNotFed pins the comboMask invariant against the MACR
// branch: the intercepted chord puts no rune in the field, so the buffer
// must stay empty — a double tap after it finds nothing and refuses with
// empty-buffer (the same buffer oracle as the combo corpus).
func TestActor_MACRBufferNotFed(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	enableMACR(a, "a")

	a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
	_ = a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4})
	a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL, Mods: engine.MaskMod4, Release: true})

	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if !strings.Contains(buf.String(), `"reason":"empty-buffer"`) {
		t.Fatalf("empty-buffer record missing — the intercepted letter fed the buffer; log:\n%s", buf.String())
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Errorf("commits = %q, want none — the interception puts no rune in the field", texts)
	}
}

// TestActor_MACRConsumedUpstream pins the b.2 detect: a Super press/release
// pair with NO letter press in between (the shell consumed the chord before
// the IME) WARNs with the grep-stable reason and counts; a letter that DID
// arrive — intercepted or transited — means nothing was consumed upstream,
// so no WARN; the two counters grow independently.
func TestActor_MACRConsumedUpstream(t *testing.T) {
	t.Run("bare super press/release warns and counts", func(t *testing.T) {
		buf := captureLogs(t)
		a, _ := wiredActor()
		enableMACR(a, "a")

		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL, Mods: engine.MaskMod4, Release: true})

		if !strings.Contains(buf.String(), `"msg":"super combo skipped","reason":"consumed-upstream"`) {
			t.Fatalf("consumed-upstream WARN missing after a bare Super press/release; log:\n%s", buf.String())
		}
		if got := a.MACRCounters(); got.ConsumedUpstream != 1 || got.SuperIntercepted != 0 {
			t.Errorf("counters = %+v, want {SuperIntercepted:0 ConsumedUpstream:1}", got)
		}
	})

	t.Run("a seen letter during the hold is not consumed-upstream", func(t *testing.T) {
		buf := captureLogs(t)
		a, _ := wiredActor()
		enableMACR(a, "a")

		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
		_ = a.HandleKey(engine.EngineEvent{Keyval: uint32('x'), Mods: engine.MaskMod4}) // transits: not in the set
		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL, Mods: engine.MaskMod4, Release: true})

		if strings.Contains(buf.String(), `"msg":"super combo skipped"`) {
			t.Errorf("a transited super+x misreported as consumed-upstream; log:\n%s", buf.String())
		}
		if got := a.MACRCounters(); got.ConsumedUpstream != 0 {
			t.Errorf("ConsumedUpstream = %d, want 0 — the letter reached the IME", got.ConsumedUpstream)
		}
	})

	t.Run("an intercepted hold does not warn; the counters grow separately", func(t *testing.T) {
		buf := captureLogs(t)
		a, _ := wiredActor()
		enableMACR(a, "a")

		// First the solo pair: one legitimate warn.
		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL, Mods: engine.MaskMod4, Release: true})
		// Then a hold that intercepts: the release must NOT warn again.
		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
		_ = a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4})
		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL, Mods: engine.MaskMod4, Release: true})

		if got := strings.Count(buf.String(), `"msg":"super combo skipped"`); got != 1 {
			t.Errorf("super-combo-skipped records = %d, want exactly 1 (the intercepted hold must not warn); log:\n%s",
				got, buf.String())
		}
		if got := a.MACRCounters(); got.SuperIntercepted != 1 || got.ConsumedUpstream != 1 {
			t.Errorf("counters = %+v, want {SuperIntercepted:1 ConsumedUpstream:1} — the counters grow separately", got)
		}
	})
}

// TestActor_MACRRULayoutStillIntercepts pins the branch ORDER (Pattern 6):
// with the internal mode flipped to RU — where a plain 'a' press would be
// consumed and committed as 'ф' — super+a is still intercepted FIRST: the
// burst carries the LATIN keyval and the RU branch never fires (the plan's
// live trace proved the interception input is layout-independent).
func TestActor_MACRRULayoutStillIntercepts(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	if !strings.Contains(buf.String(), `"msg":"mode","to":"ru"`) {
		t.Fatalf("RU flip missing before the interception round; log:\n%s", buf.String())
	}

	enableMACR(a, "a")
	a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
	consumed := a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4})
	if !consumed {
		t.Fatalf("super+a in RU mode was not consumed — the MACR branch sits above the mode branches")
	}
	releaseSuper(a)

	calls := sink.forwardCalls()
	if len(calls) != 4 || calls[1].keyval != uint32('a') {
		t.Fatalf("burst = %+v, want the Ctrl+LATIN-a shape — the RU translation must never reach it", calls)
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Fatalf("RU mode committed %q on an intercepted chord — the branch order is violated", texts)
	}
}

// Per-app MACR corpus (plan 03-05 Task 3, ADR-005 a): the interception
// scope of a non-empty macr.apps list is the focused application — the
// a11y identity source decides, and its FAILURE degrades to the global
// rule with a WARN (the ADR-005 ladder: the observer's absence never
// switches the layer off). The observer starts lazily — only when a
// non-empty list is in force.

// macrListedApp is the per-app corpus's list entry (goconst: one name,
// four uses).
const macrListedApp = "org.gnome.gnome-text-editor"

// macrZenityApp is the corpus's unlisted focused app (the out-of-list
// side and the lazy-start rounds' list entry).
const macrZenityApp = "org.gnome.Zenity"

// The per-app corpus's failure sentinels (err113: static, not dynamic).
var (
	errAppidBusDead = errors.New("a11y bus dead")
	errAppidNoBus   = errors.New("no a11y bus")
)

// fakeAppid is the identity-source test double: a fixed answer or a fixed
// failure.
type fakeAppid struct {
	app string
	err error
}

// FocusedApp answers the double's programmed identity.
func (f fakeAppid) FocusedApp() (string, error) {
	return f.app, f.err
}

// TestActor_MACRPerAppMatch pins the scoping itself: with apps
// ["org.gnome.gnome-text-editor"] in force, a focused app IN the list is
// intercepted and one OUTSIDE it transits — the global rule is OFF while
// the per-app list scopes the layer.
func TestActor_MACRPerAppMatch(t *testing.T) {
	t.Run("focused app in the list intercepts", func(t *testing.T) {
		a, sink := wiredActor()
		a.UseAppid(fakeAppid{app: macrListedApp})
		a.SetOptions(session.Options{
			MACREnabled: true,
			MACRLetters: map[rune]bool{'a': true},
			MACRApps:    []string{macrListedApp},
		})

		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
		if consumed := a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4}); !consumed {
			t.Fatalf("super+a with the focused app IN the list was not consumed")
		}
		releaseSuper(a)
		if got := len(sink.forwardCalls()); got != 4 {
			t.Errorf("in-list burst = %d events, want the 4-event shape", got)
		}
	})

	t.Run("focused app outside the list transits", func(t *testing.T) {
		a, sink := wiredActor()
		a.UseAppid(fakeAppid{app: macrZenityApp})
		a.SetOptions(session.Options{
			MACREnabled: true,
			MACRLetters: map[rune]bool{'a': true},
			MACRApps:    []string{macrListedApp},
		})

		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
		if consumed := a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4}); consumed {
			t.Fatalf("super+a with the focused app OUTSIDE the list was consumed — the list scopes the layer")
		}
		releaseSuper(a)
		if got := len(sink.forwardCalls()); got != 0 {
			t.Errorf("out-of-list burst = %d events, want none", got)
		}
	})
}

// TestActor_MACRAppidDegradation pins the ladder's key rung (ADR-005): an
// identity source that answers with an error does NOT switch the layer
// off — the interception falls back to the GLOBAL rule and the failure is
// visible as the WARN "app identity unavailable" (a documented
// degradation, never a silent one).
func TestActor_MACRAppidDegradation(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{err: errAppidBusDead})
	a.SetOptions(session.Options{
		MACREnabled: true,
		MACRLetters: map[rune]bool{'a': true},
		MACRApps:    []string{macrListedApp},
	})

	a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
	if consumed := a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4}); !consumed {
		t.Fatalf("super+a under a failed identity source was not consumed — degradation is the GLOBAL rule, not off")
	}
	releaseSuper(a)
	if got := len(sink.forwardCalls()); got != 4 {
		t.Errorf("degraded-mode burst = %d events, want the 4-event shape (the global rule)", got)
	}
	if !strings.Contains(buf.String(), `"msg":"app identity unavailable"`) {
		t.Errorf("degradation WARN missing; log:\n%s", buf.String())
	}
}

// TestActor_MACRGlobalWhenNoApps pins the zero-connection rung of the
// lazy start (ADR-005 a): with no per-app list the observer NEVER starts
// and the global rule decides alone.
func TestActor_MACRGlobalWhenNoApps(t *testing.T) {
	a, sink := wiredActor()
	starts := 0
	a.UseAppidStarter(func() (session.AppidSource, error) {
		starts++

		return fakeAppid{app: "never"}, nil
	})
	a.SetOptions(session.Options{
		MACREnabled: true,
		MACRLetters: map[rune]bool{'a': true},
	})

	a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
	if consumed := a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4}); !consumed {
		t.Fatalf("super+a under the global rule was not consumed")
	}
	releaseSuper(a)
	if starts != 0 {
		t.Errorf("observer started %d times with an empty app list, want 0 — the a11y bus is per-app-only", starts)
	}
	if got := len(sink.forwardCalls()); got != 4 {
		t.Errorf("global burst = %d events, want the 4-event shape", got)
	}
}

// TestActor_MACRAppidLazyStart pins the lazy start itself (ADR-005 a): a
// non-empty list starts the observer exactly once (no per-event restarts),
// and a failed start degrades with the WARN while the global rule stays
// in force.
func TestActor_MACRAppidLazyStart(t *testing.T) {
	t.Run("non-empty apps start the observer exactly once", func(t *testing.T) {
		a, _ := wiredActor()
		starts := 0
		a.UseAppidStarter(func() (session.AppidSource, error) {
			starts++

			return fakeAppid{app: macrZenityApp}, nil
		})

		a.SetOptions(session.Options{
			MACREnabled: true,
			MACRLetters: map[rune]bool{'a': true},
			MACRApps:    []string{macrZenityApp},
		})
		if starts != 1 {
			t.Fatalf("observer started %d times on the first non-empty list, want 1", starts)
		}
		a.HandleKey(engine.EngineEvent{Keyval: uint32('x')})
		a.HandleKey(engine.EngineEvent{Keyval: uint32('x'), Release: true})
		if starts != 1 {
			t.Errorf("observer started %d times after further events, want still 1 (one lazy start)", starts)
		}
	})

	t.Run("start failure degrades to the global rule with a WARN", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()
		a.UseAppidStarter(func() (session.AppidSource, error) {
			return nil, errAppidNoBus
		})
		a.SetOptions(session.Options{
			MACREnabled: true,
			MACRLetters: map[rune]bool{'a': true},
			MACRApps:    []string{macrZenityApp},
		})

		a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL})
		if consumed := a.HandleKey(engine.EngineEvent{Keyval: uint32('a'), Mods: engine.MaskMod4}); !consumed {
			t.Fatalf("super+a after a failed observer start was not consumed — the global rule stays in force")
		}
		releaseSuper(a)
		if got := len(sink.forwardCalls()); got != 4 {
			t.Errorf("failed-start burst = %d events, want the 4-event shape (the global rule)", got)
		}
		if !strings.Contains(buf.String(), `"msg":"app identity unavailable"`) {
			t.Errorf("failed-start WARN missing; log:\n%s", buf.String())
		}
	})
}

// fakeCfgStatus is the config-source double of the snapshot corpus: a
// healthy Defaults-shaped snapshot with the corpus window (so applySnapshot
// never disturbs the farWindow timer isolation) plus the optional status
// surface — the D-32 fields StatusSnapshot must lift into the snapshot.
type fakeCfgStatus struct {
	err error
}

// Snapshot returns the corpus document with the farWindow-equivalent tap
// window — a smaller window would re-arm the actor's real AfterFunc with a
// test-hostile deadline.
func (f fakeCfgStatus) Snapshot() config.Config {
	c := config.Defaults()
	c.Timeouts.TapWindowMs = int(farWindow.Milliseconds())

	return c
}

// ConfigPath names the served document.
func (f fakeCfgStatus) ConfigPath() string { return "/tmp/goswitch-status-snapshot.yaml" }

// LastError is the D-32 surface: nil while the served snapshot is valid.
func (f fakeCfgStatus) LastError() error { return f.err }

// errStatusRejected is the corpus's static rejection (err113): the D-32
// error a rejected reload leaves on the source.
var errStatusRejected = errors.New("decode config: field verify_wait_mss not found")

// TestActor_StatusSnapshot pins the snapshot's base state (INST-02): the
// fresh actor reports EN mode with zero counters and no config path, and
// the single-tap flip reaches the mode field. No typed or corrected text
// appears anywhere in the snapshot (T-03-06-03).
func TestActor_StatusSnapshot(t *testing.T) {
	t.Run("fresh actor: EN mode, zero counters, no config", func(t *testing.T) {
		a, _ := wiredActor()
		st := a.StatusSnapshot()
		if st.Mode != "en" {
			t.Errorf("fresh mode = %q, want en", st.Mode)
		}
		if st.CorrectionsDone != 0 || st.CorrectionsSkipped != 0 || st.SkipReasons != nil {
			t.Errorf("fresh correction counters = %+v, want all zero", st)
		}
		if st.SuperIntercepted != 0 || st.SuperUpstreamConsumed != 0 {
			t.Errorf("fresh MACR counters = %+v, want all zero", st)
		}
		if st.ConfigPath != "" {
			t.Errorf("no-config snapshot carries path %q", st.ConfigPath)
		}
	})

	t.Run("mode flip reaches the snapshot", func(t *testing.T) {
		a, _ := wiredActor()
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
		if st := a.StatusSnapshot(); st.Mode != "ru" {
			t.Errorf("mode after the single-tap flip = %q, want ru", st.Mode)
		}
	})
}

// TestActor_StatusCounters pins the counter half of the snapshot: the
// counters increment through the real pipeline paths — a settled
// correction counts done, a refusal counts skipped with its D-20 reason.
func TestActor_StatusCounters(t *testing.T) {
	t.Run("settled correction counts done", func(t *testing.T) {
		a, sink := wiredActor()
		typeWord(a, wordEN)
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
		a.HandleSurroundingText("abc "+wordEN, runeLen("abc "+wordEN), runeLen("abc "+wordEN))

		if got := sink.commitTexts(); len(got) != 1 || got[0] != wordRU {
			t.Fatalf("correction commits = %q, want one %q (the pipeline precondition)", got, wordRU)
		}
		st := a.StatusSnapshot()
		if st.CorrectionsDone != 1 || st.CorrectionsSkipped != 0 {
			t.Errorf("counters after one settled correction = done %d skipped %d, want 1/0",
				st.CorrectionsDone, st.CorrectionsSkipped)
		}
	})

	t.Run("empty-buffer refusal counts skipped with its reason", func(t *testing.T) {
		a, _ := wiredActor()
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)

		st := a.StatusSnapshot()
		if st.CorrectionsDone != 0 || st.CorrectionsSkipped != 1 {
			t.Errorf("counters after the empty-buffer refusal = done %d skipped %d, want 0/1",
				st.CorrectionsDone, st.CorrectionsSkipped)
		}
		if st.SkipReasons["empty-buffer"] != 1 {
			t.Errorf("skip reasons = %v, want empty-buffer:1", st.SkipReasons)
		}
	})
}

// TestActor_StatusConfigFields pins the D-32 half of the snapshot: the
// attached source's path and last error lift into the status fields — the
// healthy document and the rejected reload (the error text included).
func TestActor_StatusConfigFields(t *testing.T) {
	a, _ := wiredActor()
	a.AttachConfig(fakeCfgStatus{})
	a.HandleKey(engine.EngineEvent{Keyval: uint32('x')}) // one event folds the source in

	st := a.StatusSnapshot()
	if st.ConfigPath != "/tmp/goswitch-status-snapshot.yaml" {
		t.Errorf("config path = %q, want the source's path", st.ConfigPath)
	}
	if !st.ConfigValid || st.ConfigError != "" {
		t.Errorf("healthy config status = valid %t error %q, want true/empty", st.ConfigValid, st.ConfigError)
	}

	a.AttachConfig(fakeCfgStatus{err: errStatusRejected})
	a.HandleKey(engine.EngineEvent{Keyval: uint32('y')})

	st = a.StatusSnapshot()
	if st.ConfigValid {
		t.Error("config valid after a rejected reload = true, want false")
	}
	if !strings.Contains(st.ConfigError, "verify_wait_mss") {
		t.Errorf("config error = %q, want the rejection naming the key", st.ConfigError)
	}
}

// TestStatusCarriesVersion pins the D-37 build identity in the status
// snapshot: the version the daemon pinned at construction (its stamped or
// dev-fallback build) reaches goswitchctl status through StatusSnapshot.
func TestStatusCarriesVersion(t *testing.T) {
	a, _ := wiredActor()
	a.SetVersion("v1.2.3")

	if st := a.StatusSnapshot(); st.Version != "v1.2.3" {
		t.Errorf("snapshot Version = %q, want v1.2.3", st.Version)
	}
}

// The flip_after_correction corpus of quick plan 260927-vu8 (owner decision
// 2, 2026-09-27): a successful correction that CHANGED the text flips the
// internal script mode exactly once, strictly after the settled done record
// (the D-36 order extended); the D-24 done-without-change outcome and every
// refusal never flip; a combo gesture still flips exactly once — the
// correction flip never doubles it. The gate is Options.FlipAfterCorrection,
// fed from correction.flip_after_correction (default ON via config.Defaults,
// OFF at the zero value).

// flipOnOptions is the Options shape under test: the post-correction flip
// enabled.
func flipOnOptions() session.Options {
	return session.Options{FlipAfterCorrection: true}
}

// countModeRecords counts the INFO mode records in the captured log — the
// observable of every flip (the Single tap, the combo settle, the
// post-correction flip).
func countModeRecords(buf *syncBuffer) int {
	return strings.Count(buf.String(), `"msg":"mode"`)
}

// settleWordCorrection drives one full word correction on a: type wordEN,
// double tap, settle on the matching surrounding push. Returns the log's
// done-record presence to the caller's judgment.
func settleWordCorrection(a *session.Actor) {
	typeWord(a, wordEN)
	honestPush(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	line := "abc " + wordEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line)) // settles the pre-round
	// The two-phase completion (plan 10-11): the client's confirming push
	// completes the ladder — the commit phase runs on the deletion proof.
	a.HandleSurroundingText("abc ", 4, 4)
}

// TestActor_FlipAfterWordCorrection pins the owner decision 2 word path: a
// settled CHANGED word correction flips the internal script mode exactly
// once, strictly after the done record — and the flip itself is pure daemon
// state: zero extra calls on the sink beyond the established pipeline shape
// (one delete, one commit, the pre-round plus the verify-after Require).
func TestActor_FlipAfterWordCorrection(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.SetOptions(flipOnOptions())

	settleWordCorrection(a)

	logged := buf.String()
	if !strings.Contains(logged, `"msg":"correction","outcome":"done"`) {
		t.Fatalf("INFO completion record missing; log:\n%s", logged)
	}
	// The deferred level-1 flip (plan 10-11) fires on its delivery proof:
	// the verify-after's confirming push (the field ends with converted).
	line := "abc " + wordRU
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))
	logged = buf.String()
	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("mode records after a changed word correction = %d, want exactly 1; log:\n%s", got, logged)
	}
	if !strings.Contains(logged, `"msg":"mode","to":"ru"`) {
		t.Errorf("mode record is not the EN→RU flip; log:\n%s", logged)
	}
	// D-36 order extended: the flip lands strictly AFTER the settled done.
	iDone := strings.Index(logged, `"msg":"correction","outcome":"done"`)
	iMode := strings.Index(logged, `"msg":"mode"`)
	if iDone > iMode {
		t.Errorf("mode flip preceded the settled done record (done@%d > mode@%d) — D-36 order extended", iDone, iMode)
	}
	// The flip itself makes zero FIELD calls — the pipeline shape is exactly
	// the pre-existing one (mode is internal state); reconciled deliberately
	// from the vu8-era zero-sink-call pin (owner decision 1, quick plan
	// 260927-way): the flip's ONLY sink call is the panel-symbol update.
	if calls := sink.deleteCalls(); len(calls) != 1 || calls[0] != (deleteCall{offset: -6, nchars: 6}) {
		t.Errorf("deletions = %+v, want exactly one (-6,6) — the flip adds none", calls)
	}
	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != wordRU {
		t.Errorf("commits = %q, want exactly one %q — the flip commits nothing", texts, wordRU)
	}
	if got := sink.requireCount(); got != 2 {
		t.Errorf("require calls = %d, want 2 (pre-round + verify-after) — the flip requires nothing", got)
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru"}) {
		t.Errorf("panel symbols = %q, want exactly [ru] (owner decision 1: the indicator follows the flip)", got)
	}
}

// TestActor_FlipAfterPhraseCorrection pins the owner decision 2 phrase path:
// the same single flip strictly after the done record, over a triple tap.
func TestActor_FlipAfterPhraseCorrection(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.SetOptions(flipOnOptions())

	typeWord(a, wordEN)
	pressSpace(a)
	typeWord(a, wordEN)
	honestPush(a)
	tapShift(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	line := "abc " + phraseEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))
	a.HandleSurroundingText("abc ", 4, 4) // the deletion proof completes the ladder (plan 10-11)
	// The deferred level-1 flip (plan 10-11) fires on its delivery proof.
	line = "abc " + phraseRU
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))

	logged := buf.String()
	if !strings.Contains(logged, `"msg":"correction","outcome":"done"`) {
		t.Fatalf("INFO completion record missing; log:\n%s", logged)
	}
	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("mode records after a changed phrase correction = %d, want exactly 1; log:\n%s", got, logged)
	}
	iDone := strings.Index(logged, `"msg":"correction","outcome":"done"`)
	iMode := strings.Index(logged, `"msg":"mode"`)
	if iDone > iMode {
		t.Errorf("mode flip preceded the settled phrase done (done@%d > mode@%d)", iDone, iMode)
	}
	if calls := sink.deleteCalls(); len(calls) != 1 || calls[0] != (deleteCall{offset: -13, nchars: 13}) {
		t.Errorf("deletions = %+v, want exactly one (-13,13)", calls)
	}
}

// TestActor_FlipAfterLevel2Correction pins the owner decision 2 level-2
// path: a Backspace-level success flips after its done record too — the
// ladder level does not matter, the CHANGED settlement does.
func TestActor_FlipAfterLevel2Correction(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActorCaps(0) // no surrounding-text capability: ladder level 2
	a.SetOptions(flipOnOptions())

	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow) // level 2 executes at the decision itself

	logged := buf.String()
	if !strings.Contains(logged, `"msg":"correction","outcome":"done"`) {
		t.Fatalf("INFO completion record missing; log:\n%s", logged)
	}
	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("mode records after a changed level-2 correction = %d, want exactly 1; log:\n%s", got, logged)
	}
	iDone := strings.Index(logged, `"msg":"correction","outcome":"done"`)
	iMode := strings.Index(logged, `"msg":"mode"`)
	if iDone > iMode {
		t.Errorf("mode flip preceded the settled level-2 done (done@%d > mode@%d)", iDone, iMode)
	}
}

// TestActor_FlipGateOff pins the off reading of the gate: at the zero-value
// Options (the bare wiredActor path — every pre-existing test's world) and
// at the explicit false, a settled CHANGED correction logs done and NO mode
// record.
func TestActor_FlipGateOff(t *testing.T) {
	t.Run("zero-value options (no SetOptions)", func(t *testing.T) {
		buf := captureLogs(t)
		a, _ := wiredActor()

		settleWordCorrection(a)

		logged := buf.String()
		if !strings.Contains(logged, `"msg":"correction","outcome":"done"`) {
			t.Fatalf("INFO completion record missing; log:\n%s", logged)
		}
		if got := countModeRecords(buf); got != 0 {
			t.Errorf("mode records with the gate off = %d, want 0; log:\n%s", got, logged)
		}
	})

	t.Run("explicit false", func(t *testing.T) {
		buf := captureLogs(t)
		a, _ := wiredActor()
		a.SetOptions(session.Options{FlipAfterCorrection: false})

		settleWordCorrection(a)

		logged := buf.String()
		if !strings.Contains(logged, `"msg":"correction","outcome":"done"`) {
			t.Fatalf("INFO completion record missing; log:\n%s", logged)
		}
		if got := countModeRecords(buf); got != 0 {
			t.Errorf("mode records with flip_after_correction: false = %d, want 0; log:\n%s", got, logged)
		}
	})
}

// refusalFlipProbe runs one correction attempt that must refuse, with the
// flip ON, and pins the refusal half of owner decision 2: the D-20 reason
// record lands and the mode is never flipped.
func refusalFlipProbe(t *testing.T, wantReason string, run func(a *session.Actor)) {
	t.Helper()

	buf := captureLogs(t)
	a, _ := wiredActor()
	a.SetOptions(flipOnOptions())
	run(a)

	logged := buf.String()
	if !strings.Contains(logged, `"reason":"`+wantReason+`"`) {
		t.Errorf("%s record missing; log:\n%s", wantReason, logged)
	}
	if got := countModeRecords(buf); got != 0 {
		t.Errorf("the %s refusal flipped the mode %d times, want 0", wantReason, got)
	}
}

// TestActor_RefusalsNeverFlip pins the refusal half of owner decision 2:
// with the option ON, every D-20 refusal and only refusals leave the mode
// untouched — empty-buffer, no-letters, verify-mismatch and verify-timeout
// each log their reason and zero mode records.
func TestActor_RefusalsNeverFlip(t *testing.T) {
	t.Run("empty buffer", func(t *testing.T) {
		refusalFlipProbe(t, "empty-buffer", func(a *session.Actor) {
			tapShift(a)
			tapShift(a)
			a.ExpiryAt(expiryAfterWindow)
		})
	})

	t.Run("no letters", func(t *testing.T) {
		refusalFlipProbe(t, "no-letters", func(a *session.Actor) {
			typeWord(a, "2026")
			tapShift(a)
			tapShift(a)
			a.ExpiryAt(expiryAfterWindow)
		})
	})

	t.Run("verify mismatch", func(t *testing.T) {
		refusalFlipProbe(t, "verify-mismatch", func(a *session.Actor) {
			typeWord(a, wordEN)
			// The honest-GTK world (plan 10-08 amendment migration): the
			// cache carries tracking evidence — the nil-cache form is the
			// stale-cache terminal signature now.
			a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))
			tapShift(a)
			tapShift(a)
			a.ExpiryAt(expiryAfterWindow)
			a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))
		})
	})

	t.Run("verify timeout", func(t *testing.T) {
		refusalFlipProbe(t, "verify-timeout", func(a *session.Actor) {
			clock := &flipClock{} // the verify-deadline recorder (plan 10-10, UseVerifyTimer)
			a.UseVerifyTimer(clock.arm)
			typeWord(a, wordEN)
			tapShift(a)
			tapShift(a)
			a.ExpiryAt(expiryAfterWindow)
			clock.fire() // deterministic injection of the 100 ms deadline
		})
	})
}

// Mode-switch chord corpus (owner decision 2, quick plan 260927-way):
// Super+Space becomes a configurable goswitch-owned chord
// (hotkeys.mode_switch_chord, default super+space, empty = disabled). The
// live trace 2026-09-27 20:31 is the wire truth: Super+Space arrives as the
// SPACE keyval with mods 0x50 — Mod4|NumLock — and (with GNOME's own
// switch-input-source binding cleared) reaches the daemon's key path. The
// chord branch sits AFTER the word-layout combo and BEFORE the mode
// branches; the MACR layer is unaffected: space is not a letter, and the
// chord marks the Super hold as witnessed so the consumed-upstream detect
// stays silent.

// numLockLatch is the IBUS_MOD2_MASK latched NumLock bit every chord press
// carries on this desktop (the live trace 0x50 = Mod4|NumLock).
const numLockLatch = 0x10

// superSpaceChord returns the corpus's chord binding: a literal wire-equal
// to hotkey.ParseBinding("super+space") — the defaultComboBinding idiom
// (a literal, no corpus error handling). Keyval 0x020 with Mod4 as the FULL
// ModMask: space carries no family bit, so the held mask is exactly Mod4.
func superSpaceChord() hotkey.Binding {
	return hotkey.Binding{Keyval: 0x020, ModMask: hotkey.MaskMod4}
}

// pressSuperSpace feeds the chord press exactly as the wire delivers it and
// returns the consumption verdict.
func pressSuperSpace(a *session.Actor) bool {
	return a.HandleKey(engine.EngineEvent{Keyval: engine.KeySpace, Mods: engine.MaskMod4 | numLockLatch})
}

// TestActor_SuperSpaceChordFlipsMode pins the chord itself: a Super+Space
// press flips the script mode IMMEDIATELY (no ExpiryAt — the chord is not a
// tap), CONSUMES the press (no space lands in the field), emits exactly one
// panel-symbol update through the flip body (vu8 Task 1) and commits nothing; the
// buffer stays clean — typing "abc" afterwards yields the token "abc" with
// no leading space.
func TestActor_SuperSpaceChordFlipsMode(t *testing.T) {
	buf := captureLogsLevel(t, slog.LevelDebug)
	a, sink := wiredActor()
	a.SetOptions(session.Options{ModeSwitchChord: superSpaceChord()})

	if consume := pressSuperSpace(a); !consume {
		t.Fatalf("super+space press transited — the chord must own and consume it (owner decision 2)")
	}
	logged := buf.String()
	if !strings.Contains(logged, `"msg":"combo","kind":"mode-switch-chord"`) {
		t.Errorf("chord entry record missing; log:\n%s", logged)
	}
	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("mode records after the chord press = %d, want exactly 1 (immediate, no expiry)", got)
	}
	if !strings.Contains(logged, `"msg":"mode","to":"ru"`) {
		t.Errorf("the chord did not flip EN→RU; log:\n%s", logged)
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru"}) {
		t.Errorf("panel symbols = %q, want exactly [ru] (vu8 Task 1 through the flip body)", got)
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Errorf("chord commits = %q, want none — the press never lands in the field", texts)
	}

	// The buffer stays clean: flip back with a second chord (RU → EN) and
	// type — the chord press fed no rune, so the corrected token is exactly
	// "abc" with no leading space.
	pressSuperSpace(a)
	if got := countModeRecords(buf); got != 2 {
		t.Fatalf("mode records after the flip-back = %d, want 2; log:\n%s", got, buf.String())
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru", "en"}) {
		t.Errorf("panel symbols = %q, want exactly [ru en]", got)
	}
	typeWord(a, "abc")
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	line := "done abc"
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))
	if !strings.Contains(buf.String(), `"source":"abc"`) {
		t.Errorf("the corrected token is not the clean \"abc\" — a space leaked into the buffer; log:\n%s",
			buf.String())
	}
}

// TestActor_SuperSpaceChordKillsPendingTap pins the Pitfall-4 Reset: a tap
// series armed when the chord press lands DIES with the chord's deliberate
// FSM Reset — after the window expires exactly ONE mode record exists (the
// chord's), and no tap decision ever fires.
func TestActor_SuperSpaceChordKillsPendingTap(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	a.SetOptions(session.Options{ModeSwitchChord: superSpaceChord()})

	tapShift(a) // series armed — the single-tap decision is pending
	pressSuperSpace(a)
	a.ExpiryAt(expiryAfterWindow) // the armed window closes on a dead series

	if got := countActions(buf); got != 0 {
		t.Errorf("actions after the killed series = %d, want 0 — the pending tap decision must die", got)
	}
	if got := countModeRecords(buf); got != 1 {
		t.Errorf("mode records = %d, want exactly 1 (the chord's flip, never a tap flip)", got)
	}
}

// TestActor_SuperSpaceChordWitnessesSuperHold pins the MACR contract of the
// chord (decision 2: space is not a letter — the interception order must
// not disturb MACR): with MACR enabled, a bare Super press followed by the
// chord press leaves BOTH counters at zero and logs no consumed-upstream
// WARN — the chord marked the hold as witnessed (goswitch itself consumed
// the press, nothing was swallowed upstream). The existing MACR corpus
// (super+b intercepted, burst on release) passes unmodified.
func TestActor_SuperSpaceChordWitnessesSuperHold(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.SetOptions(session.Options{
		MACREnabled:     true,
		MACRLetters:     map[rune]bool{'b': true},
		ModeSwitchChord: superSpaceChord(),
	})

	a.HandleKey(engine.EngineEvent{Keyval: engine.KeySuperL}) // the hold opens
	if consume := pressSuperSpace(a); !consume {
		t.Fatalf("super+space transited under MACR — the chord branch must sit above the mode branches")
	}
	releaseSuper(a)

	if got := a.MACRCounters(); got.SuperIntercepted != 0 || got.ConsumedUpstream != 0 {
		t.Errorf("counters = %+v, want zeros — space is not a MACR letter and the hold WAS witnessed", got)
	}
	if strings.Contains(buf.String(), `"msg":"super combo skipped"`) {
		t.Errorf("consumed-upstream WARN after a chord-witnessed hold; log:\n%s", buf.String())
	}
	if calls := sink.forwardCalls(); len(calls) != 0 {
		t.Errorf("chord forwarded %+v, want none — the chord is not a MACR remap", calls)
	}
}

// TestActor_SuperSpaceChordGateOff pins the off readings: at the zero-value
// Options (the no-config world) the chord press transits untouched — no
// consumption, no mode record, no sink call — and with the chord set, a
// Mod2-only space (NumLock without Mod4) never matches.
func TestActor_SuperSpaceChordGateOff(t *testing.T) {
	t.Run("zero-value options", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()

		if consume := pressSuperSpace(a); consume {
			t.Fatalf("chord press consumed with no chord configured — the zero value must be disabled")
		}
		if got := countModeRecords(buf); got != 0 {
			t.Errorf("mode records = %d, want 0; log:\n%s", got, buf.String())
		}
		if got := len(sink.modeSymbols()); got != 0 {
			t.Errorf("panel symbols = %d, want 0 — a transiting space emits nothing", got)
		}
	})

	t.Run("Mod2-only space never matches", func(t *testing.T) {
		buf := captureLogs(t)
		a, _ := wiredActor()
		a.SetOptions(session.Options{ModeSwitchChord: superSpaceChord()})

		if consume := a.HandleKey(engine.EngineEvent{Keyval: engine.KeySpace, Mods: numLockLatch}); consume {
			t.Fatalf("Mod2-only space consumed — the held mask requires Mod4")
		}
		if got := countModeRecords(buf); got != 0 {
			t.Errorf("mode records = %d, want 0 — NumLock-only space is a keystroke, not the chord; log:\n%s",
				got, buf.String())
		}
	})
}

// TestActor_SuperSpaceChordLosesToCombo pins the branch order: with
// ModeSwitchChord EQUAL to the word-layout combo binding, the combo branch
// wins (the chord branch sits after it) — the word pipeline launches and
// the gesture still settles with exactly one flip, never a double.
func TestActor_SuperSpaceChordLosesToCombo(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	combo := reloadCfg(300, "shift+ctrl_r").Hotkeys.WordLayoutCombo
	binding, err := hotkey.ParseBinding(combo)
	if err != nil {
		t.Fatalf("ParseBinding(%q): %v", combo, err)
	}
	a.SetOptions(session.Options{WordLayoutCombo: binding, ModeSwitchChord: binding})

	typeWord(a, wordEN)
	pressComboDefault(a)

	if got := sink.requireCount(); got != 1 {
		t.Fatalf("combo require calls = %d, want 1 — the combo branch must own the press", got)
	}
	line := "abc " + wordEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))
	if got := countModeRecords(buf); got != 1 {
		t.Errorf("mode records = %d, want exactly 1 — no double flip on a chord/combo collision", got)
	}
}

// TestActor_HotReloadModeSwitchChord pins the CONF-02 wiring of the chord
// (the applySnapshot fold with the EMPTY-DISABLES twist): the document's
// mode_switch_chord is folded per snapshot — "super+space" arms the chord,
// an explicit "" DISABLES it (transit, not last-good), and a document back
// restores it. No restart, no SetOptions.
func TestActor_HotReloadModeSwitchChord(t *testing.T) {
	a, _ := wiredActor()
	src := &reloadSource{cfg: reloadCfg(300, "shift+ctrl_r")} // the default document carries super+space
	a.AttachConfig(src)

	if consume := pressSuperSpace(a); !consume {
		t.Fatalf("chord press transited under the default document — the chord must be live")
	}

	src.set(func() config.Config {
		cfg := reloadCfg(300, "shift+ctrl_r")
		cfg.Hotkeys.ModeSwitchChord = "" // explicit disable

		return cfg
	}())
	if consume := pressSuperSpace(a); consume {
		t.Fatalf("chord consumed after the empty reload — empty must DISABLE, not keep last-good")
	}

	src.set(reloadCfg(300, "shift+ctrl_r"))
	if consume := pressSuperSpace(a); !consume {
		t.Fatalf("chord press transited after the restore reload — the chord must come back")
	}
}

// TestActor_ComboStillSingleFlipWithCorrectionFlip pins the single-flip
// contract of the gesture: with the option ON, a combo over a correctable
// word STILL produces exactly one mode record — the combo's settleCombo
// flip; the correction flip skips (the comboPending guard) so the gesture
// never doubles. TestActor_ComboWordThenFlip keeps its zero-Options world.
func TestActor_ComboStillSingleFlipWithCorrectionFlip(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.SetOptions(flipOnOptions())

	typeWord(a, wordEN)
	pressComboDefault(a)
	line := "abc " + wordEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))

	logged := buf.String()
	if !strings.Contains(logged, `"msg":"correction","outcome":"done"`) {
		t.Fatalf("INFO completion record missing; log:\n%s", logged)
	}
	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("combo correction mode records = %d, want exactly 1 (single flip per gesture); log:\n%s", got, logged)
	}
	iDone := strings.Index(logged, `"msg":"correction","outcome":"done"`)
	iMode := strings.Index(logged, `"msg":"mode"`)
	if iDone > iMode {
		t.Errorf("combo flip preceded the settled done (done@%d > mode@%d) — D-36 order", iDone, iMode)
	}
	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != wordRU {
		t.Errorf("commits = %q, want exactly one %q", texts, wordRU)
	}
}

// pressEnter feeds the CORR-09 reset trigger: Enter hard-resets the phrase
// buffer — the way the field ends a correction episode between episodes.
func pressEnter(a *session.Actor) {
	a.HandleKey(engine.EngineEvent{Keyval: engine.KeyReturn})
}

// flipDoc builds a reload document with the given flip_after_correction
// value over the documented defaults.
func flipDoc(on bool) config.Config {
	cfg := reloadCfg(300, "shift+ctrl_r")
	cfg.Correction.FlipAfterCorrection = on

	return cfg
}

// flipSrc builds the reload source of the flip corpus serving the given
// flip_after_correction value over the documented defaults.
func flipSrc(on bool) *reloadSource {
	return &reloadSource{cfg: flipDoc(on)}
}

// assertNoFlipAfterLastDone pins that no mode record follows the LAST done
// record in logged — the reloaded-off gate must stop the flip.
func assertNoFlipAfterLastDone(t *testing.T, logged string) {
	t.Helper()

	iLastDone := strings.LastIndex(logged, `"msg":"correction","outcome":"done"`)
	iLastMode := strings.LastIndex(logged, `"msg":"mode"`)
	if iLastMode > iLastDone {
		t.Errorf("the post-reload correction flipped (last mode@%d > last done@%d)", iLastMode, iLastDone)
	}
}

// settleRUWordCorrection drives one full word correction episode typed in
// RU mode: the RU commits feed the buffer script-true Cyrillic, so the
// field push carries wordRU and the correction converts RU→EN.
func settleRUWordCorrection(a *session.Actor) {
	pressEnter(a)
	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	line := "abc " + wordRU
	a.HandleSurroundingText(line, runeLen(line), runeLen(line)) // settles the pre-round
	a.HandleSurroundingText("abc ", 4, 4)                       // the deletion proof completes the ladder (plan 10-11)
}

// TestActor_HotReloadFlipOn pins the flip-on document's first half: the
// attached document's correction.flip_after_correction=true flips a settled
// CHANGED correction — the deferred level-1 flip (plan 10-11) fires on its
// delivery proof.
func TestActor_HotReloadFlipOn(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	a.AttachConfig(flipSrc(true))

	settleWordCorrection(a)
	line := "abc " + wordRU
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))

	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("mode records under a flip-on document = %d, want exactly 1; log:\n%s", got, buf.String())
	}
}

// TestActor_HotReloadFlipAfterCorrection pins the CONF-02 wiring of the flip
// gate: the document's correction.flip_after_correction is folded per
// snapshot, so a reload true→false stops the flip for the NEXT correction
// and false→true restores it — no restart, no SetOptions.
func TestActor_HotReloadFlipAfterCorrection(t *testing.T) {
	t.Run("true to false stops the flip", func(t *testing.T) {
		buf := captureLogs(t)
		a, _ := wiredActor()
		src := flipSrc(true)
		a.AttachConfig(src)

		settleWordCorrection(a) // EN mode: done + one flip (the mode is now RU)
		// The deferred level-1 flip (plan 10-11) fires on its delivery proof.
		line := "abc " + wordRU
		a.HandleSurroundingText(line, runeLen(line), runeLen(line))
		if got := countModeRecords(buf); got != 1 {
			t.Fatalf("mode records before the reload = %d, want exactly 1; log:\n%s", got, buf.String())
		}

		off := flipDoc(false)
		src.set(off)
		settleRUWordCorrection(a)

		logged := buf.String()
		if got := strings.Count(logged, `"msg":"correction","outcome":"done"`); got != 2 {
			t.Fatalf("done records = %d, want 2; log:\n%s", got, logged)
		}
		if got := countModeRecords(buf); got != 1 {
			t.Errorf("mode records after the true→false reload = %d, want 1 (the pre-reload flip only)", got)
		}
		assertNoFlipAfterLastDone(t, logged)
	})

	t.Run("false to true restores the flip", func(t *testing.T) {
		buf := captureLogs(t)
		a, _ := wiredActor()
		src := flipSrc(false) // explicit off — the built-in default is ON
		a.AttachConfig(src)

		settleWordCorrection(a)
		if got := countModeRecords(buf); got != 0 {
			t.Fatalf("mode records before the reload = %d, want 0; log:\n%s", got, buf.String())
		}

		on := flipDoc(true)
		src.set(on)
		pressEnter(a)
		settleWordCorrection(a) // the mode is still EN (no flip happened above)
		// The deferred level-1 flip (plan 10-11) fires on its delivery proof.
		line := "abc " + wordRU
		a.HandleSurroundingText(line, runeLen(line), runeLen(line))

		logged := buf.String()
		if got := countModeRecords(buf); got != 1 {
			t.Fatalf("mode records after the false→true reload = %d, want exactly 1; log:\n%s", got, logged)
		}
		iLastDone := strings.LastIndex(logged, `"msg":"correction","outcome":"done"`)
		iLastMode := strings.LastIndex(logged, `"msg":"mode"`)
		if iLastDone > iLastMode {
			t.Errorf("the restored flip preceded the done record (done@%d > mode@%d)", iLastDone, iLastMode)
		}
	})
}

// TestActor_FlipAfterSelectionCorrection extends the owner decision-2 flip
// to the SELECTION path (owner report 2026-09-28): a settled CHANGED
// selection correction flips the internal script
// mode exactly once, strictly after the done record, mirroring the word
// path. The combo path still flips via settleCombo exactly once.
func TestActor_FlipAfterSelectionCorrection(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.SetOptions(flipOnOptions())

	field := wordEN + " " + wordRU
	a.HandleSurroundingText(field, 6, 0) // selection [0,6), right-to-left
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	logged := buf.String()
	if !strings.Contains(logged, `"msg":"correction","outcome":"done"`) {
		t.Fatalf("INFO completion record missing; log:\n%s", logged)
	}
	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("mode records after a changed selection correction = %d, want exactly 1; log:\n%s", got, logged)
	}
	if !strings.Contains(logged, `"msg":"mode","to":"ru"`) {
		t.Errorf("mode record is not the EN→RU flip; log:\n%s", logged)
	}
	iDone := strings.Index(logged, `"msg":"correction","outcome":"done"`)
	iMode := strings.Index(logged, `"msg":"mode"`)
	if iDone > iMode {
		t.Errorf("mode flip preceded the settled done record (done@%d > mode@%d)", iDone, iMode)
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru"}) {
		t.Errorf("panel symbols = %q, want exactly [ru]", got)
	}
}

// TestActor_FlipDirectionFollowsConvertedScript pins the owner rule
// (2026-09-28): after a correction the mode is SET to the script of the
// CONVERTED result, not toggled — correcting a Cyrillic selection while
// in EN mode leaves the mode EN (the text is Latin now), and correcting a
// Latin word to Cyrillic makes it RU.
func TestActor_FlipDirectionFollowsConvertedScript(t *testing.T) {
	t.Run("cyr-to-lat while EN stays EN silently", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()
		a.SetOptions(flipOnOptions())

		field := wordRU + " " + wordRU
		a.HandleSurroundingText(field, 6, 0) // selection [0,6) of привет
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)

		if got := countModeRecords(buf); got != 0 {
			t.Fatalf("mode records = %d, want 0 (set-by-script, not a toggle); log:\n%s",
				got, buf.String())
		}
		if got := sink.modeSymbols(); len(got) != 0 {
			t.Errorf("panel symbols = %q, want none (no transition)", got)
		}
	})

	t.Run("cyr-to-lat while RU flips to EN", func(t *testing.T) {
		buf := captureLogs(t)
		a, _ := wiredActor()
		a.SetOptions(flipOnOptions())

		// Enter RU mode first (single tap = manual toggle).
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)

		field := wordRU + " " + wordRU
		a.HandleSurroundingText(field, 6, 0)
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)

		if got := countModeRecords(buf); got != 2 {
			t.Fatalf("mode records = %d, want 2 (manual + correction); log:\n%s", got, buf.String())
		}
		if !strings.Contains(buf.String(), `"msg":"mode","to":"en"`) {
			t.Errorf("the correction's record must set en (the converted script); log:\n%s", buf.String())
		}
	})
}

// The switcher-seam corpus of plan 05-03 (D-52): every flip gesture executes
// org.freedesktop.IBus.SetGlobalEngine THROUGH the injected seam — the mode
// record stays the byte-stable e2e oracle, the switch lands strictly between
// the mode record and the panel-symbol emit (the D-36 extension), a failing
// seam is a WARN that never stops typing, and the snapshot names the active
// engine. The seam is a hard-deadline call: a wedged bus can cost a flip at
// most switchTimeout, never an unbounded stall.

// errSwitchInjected is the failing seam's static cause (err113) — the corpus
// injects it to pin the WARN-not-fatal contract.
var errSwitchInjected = errors.New("injected switch failure")

// flipBudget is the switchTimeout contract from the outside: the switcher's
// context deadline is at most this far from the call instant. It mirrors
// switchTimeout and sits ABOVE the measured live SetGlobalEngine round trip,
// so the deadline guards the wedge path without firing on healthy flips.
const flipBudget = 150 * time.Millisecond

// liveSwitchRTTMax is the measured maximum of a real SetGlobalEngine round
// trip: every live proof lands at ~41-45 ms (the 05-05 flip proofs and the
// 2026-09-30 journal — the old 40 ms deadline fired before completion on
// every flip). The seam-entry budget must sit strictly above it.
const liveSwitchRTTMax = 45 * time.Millisecond

// attachBudget bounds the independent AttachEngine witness of G-6-1: the
// factory path is an atomic store and must return in microseconds — a
// quarter second sits orders of magnitude above any scheduling delay and
// far below the flip deadline it must never consume.
const attachBudget = 250 * time.Millisecond

// switchProbe is the switcher-seam double: it records every call's target
// and context deadline — the flip path's observable — with the hook form of
// fakeSink's modeHook (the corpus interleaves the call with the journal and
// the sink's op log).
type switchProbe struct {
	mu     sync.Mutex
	names  []string
	onCall func(name string)
}

// switcher is the SetSwitcher-shaped recording function.
func (p *switchProbe) switcher(_ context.Context, name string) error {
	p.mu.Lock()
	p.names = append(p.names, name)
	hook := p.onCall
	p.mu.Unlock()

	if hook != nil {
		hook(name)
	}

	return nil
}

// targets snapshots the recorded switch targets.
func (p *switchProbe) targets() []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	return append([]string(nil), p.names...)
}

// TestActor_FlipRoutesThroughSwitcher pins the D-52 execution: a single-tap
// flip routes through the seam with the TARGET mode's engine — goswitch-ru
// on EN→RU, goswitch-en on the flip back — while the byte-stable mode
// records stay the flip's observable and the panel emit stays one per flip.
func TestActor_FlipRoutesThroughSwitcher(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	probe := &switchProbe{}
	a.SetSwitcher(probe.switcher)

	flipMode(a) // EN → RU
	flipMode(a) // RU → EN

	if got := probe.targets(); !slices.Equal(got, []string{engine.NameRU, engine.NameEN}) {
		t.Errorf("switch targets = %q, want exactly [%s %s] — the target mode's engine per flip",
			got, engine.NameRU, engine.NameEN)
	}
	logged := buf.String()
	if !strings.Contains(logged, `"msg":"mode","to":"ru"`) || !strings.Contains(logged, `"msg":"mode","to":"en"`) {
		t.Errorf("byte-stable mode records missing; log:\n%s", logged)
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru", "en"}) {
		t.Errorf("panel symbols = %q, want exactly [ru en] — one emit per flip", got)
	}
	if got := sink.requireCount(); got != 0 {
		t.Errorf("flip made %d RequireSurroundingText calls, want 0", got)
	}
}

// TestActor_FlipOrderPinned pins the flip's record order (the D-36
// extension): the byte-stable mode record lands strictly BEFORE the switcher
// call, the switcher call strictly BEFORE the panel-symbol emit; on the combo
// gesture the switch lands strictly after the settled done record.
func TestActor_FlipOrderPinned(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	probe := &switchProbe{}
	var swSawModeRecord, symbolSawSwitch bool
	probe.onCall = func(name string) {
		to := "en"
		if name == engine.NameRU {
			to = "ru"
		}
		swSawModeRecord = strings.Contains(buf.String(), `"msg":"mode","to":"`+to+`"`)
	}
	sink.modeHook = func(_ string) {
		symbolSawSwitch = len(probe.targets()) == 1 && swSawModeRecord
	}
	a.SetSwitcher(probe.switcher)

	flipMode(a) // EN → RU
	if !swSawModeRecord {
		t.Errorf("the switcher fired before the mode record — order is mode → switch; log:\n%s", buf.String())
	}
	if !symbolSawSwitch {
		t.Error("UpdateModeSymbol fired before the switcher call — the pinned order is switch → emit")
	}

	// The combo leg: the word half settles FIRST (its done record is in the
	// log), then the flip — mode record, then the switch (D-36 extended).
	buf2 := captureLogs(t)
	comboDone, comboMode := false, false
	probe2 := &switchProbe{}
	probe2.onCall = func(_ string) {
		logged := buf2.String()
		comboDone = strings.Contains(logged, `"msg":"correction","outcome":"done"`)
		comboMode = strings.Contains(logged, `"msg":"mode","to":"ru"`)
	}
	a2, _ := wiredActor()
	a2.SetSwitcher(probe2.switcher)
	typeWord(a2, wordEN)
	pressComboDefault(a2)
	line := "abc " + wordEN
	a2.HandleSurroundingText(line, runeLen(line), runeLen(line))
	if !comboDone || !comboMode {
		t.Errorf("combo switch fired without its predecessors (done %t, mode %t) — the D-36 order is"+
			" done → mode → switch", comboDone, comboMode)
	}
	if got := probe2.targets(); !slices.Equal(got, []string{engine.NameRU}) {
		t.Errorf("combo switch targets = %q, want exactly [%s] — one flip per gesture", got, engine.NameRU)
	}
}

// TestActor_CorrectionFlipSetsResultEngine pins criterion 4 in two-engine
// form (the SET semantics): the engine of the RESULT script becomes active —
// ghbdtn→привет activates goswitch-ru, привет→ghbdtn-обратно activates
// goswitch-en — the "режим = скрипт результата" rule carried by the seam.
func TestActor_CorrectionFlipSetsResultEngine(t *testing.T) {
	a, _ := wiredActor()
	a.SetOptions(flipOnOptions())
	probe := &switchProbe{}
	a.SetSwitcher(probe.switcher)

	settleWordCorrection(a) // lat→cyr: the result script is Cyrillic
	// The deferred level-1 flips (plan 10-11) fire on their delivery proofs.
	line := "abc " + wordRU
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))
	if got := probe.targets(); !slices.Equal(got, []string{engine.NameRU}) {
		t.Fatalf("after ghbdtn→привет the switch targets = %q, want exactly [%s] (the result script's engine)",
			got, engine.NameRU)
	}

	settleRUWordCorrection(a) // cyr→lat: привет corrects back to ghbdtn
	line = "abc " + wordEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))
	if got := probe.targets(); !slices.Equal(got, []string{engine.NameRU, engine.NameEN}) {
		t.Errorf("after привет→ghbdtn the switch targets = %q, want [%s %s]", got, engine.NameRU, engine.NameEN)
	}
}

// TestActor_SwitcherFailureWarnsNotFatal pins the WARN-not-fatal contract
// (criterion 3): a failing seam logs the engine-name WARN, the daemon's mode
// is ALREADY switched, the panel symbol still updates and the next keys are
// handled in the new mode — typing never stops.
func TestActor_SwitcherFailureWarnsNotFatal(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.SetSwitcher(func(_ context.Context, _ string) error { return errSwitchInjected })

	flipMode(a) // EN → RU — the bus call fails, the internal flip stands

	logged := buf.String()
	if !strings.Contains(logged, `"msg":"engine switch failed","engine":"`+engine.NameRU+`"`) {
		t.Errorf("failure WARN with the engine name missing; log:\n%s", logged)
	}
	if !strings.Contains(logged, `"msg":"mode","to":"ru"`) {
		t.Errorf("mode record missing — the internal flip must stand; log:\n%s", logged)
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru"}) {
		t.Errorf("panel symbols = %q, want exactly [ru] — the emit survives the failure", got)
	}
	if consume := a.HandleKey(engine.EngineEvent{Keyval: uint32('g')}); !consume {
		t.Fatal("post-failure RU press transited — the mode did not switch internally")
	}
	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != "п" {
		t.Errorf("post-failure commits = %q, want exactly [п] — typing continues in the new mode", texts)
	}
}

// TestActor_SwitcherDeadlineBounded pins the flip's deadline discipline
// (T-05-03-01, Pitfall 4): the switcher receives a context with a hard
// deadline that sits ABOVE the measured live SetGlobalEngine round trip and
// at most flipBudget — every healthy flip completes inside it, and a seam
// stuck until that deadline stalls the actor only boundedly — a Status
// snapshot served THROUGH the stalled flip completes within the budget, so
// the keystroke path can never hang on a wedged bus.
func TestActor_SwitcherDeadlineBounded(t *testing.T) {
	a, _ := wiredActor()
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	remainingCh := make(chan time.Duration, 1)
	a.SetSwitcher(func(ctx context.Context, _ string) error {
		remaining := 2 * flipBudget // no deadline at all — the failing probe value
		if dl, ok := ctx.Deadline(); ok {
			remaining = time.Until(dl) // measured at entry — the original budget's shape
		}
		remainingCh <- remaining
		entered <- struct{}{}
		select {
		case <-release:
			return nil
		case <-ctx.Done(): // the seam honors the deadline — the real closure's shape
			return ctx.Err()
		}
	})

	tapShift(a)
	go a.ExpiryAt(expiryAfterWindow) // the flip runs on the timer goroutine's path
	select {
	case <-entered:
	case <-time.After(flipBudget + 5*time.Second):
		t.Fatal("the flip never called the switcher — the seam is not wired into the flip path")
	}
	if remaining := <-remainingCh; remaining <= liveSwitchRTTMax || remaining > flipBudget {
		t.Errorf("switcher context deadline budget = %v at entry, want in (%v, %v] — above the measured live"+
			" SetGlobalEngine round trip (a smaller deadline fires before every real flip completes),"+
			" at most the flip budget",
			remaining, liveSwitchRTTMax, flipBudget)
	}

	// While the seam is stuck, the stall is bounded: a Status snapshot
	// served THROUGH the stalled flip completes within the budget's slack —
	// the mutex-holding call can cost a bounded slice, never a hang.
	snapDone := make(chan struct{})
	go func() {
		_ = a.StatusSnapshot()
		close(snapDone)
	}()
	select {
	case <-snapDone:
	case <-time.After(flipBudget + 500*time.Millisecond):
		t.Error("StatusSnapshot blocked past the switch budget — the flip can stall the keystroke path unboundedly")
	}
	close(release)

	// The flip concluded (deadline error or late success); the mode stands
	// switched and the actor keeps serving.
	if st := a.StatusSnapshot(); st.Mode != "ru" {
		t.Errorf("mode after the stalled flip = %q, want ru — the internal flip stands", st.Mode)
	}
}

// TestActor_NilSwitcherInternalFlip pins the nil-seam degradation: with no
// switcher bound the flip stays the internal mode switch with exactly ONE
// WARN at first use, the panel symbol rides, and typing works — the missing
// seam degrades the flip, never the daemon.
func TestActor_NilSwitcherInternalFlip(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	flipMode(a) // EN → RU — no switcher bound
	flipMode(a) // RU → EN

	logged := buf.String()
	if got := strings.Count(logged, `"msg":"switcher unavailable"`); got != 1 {
		t.Errorf("nil-switcher WARNs = %d, want exactly 1 (one WARN per degradation episode); log:\n%s", got, logged)
	}
	if got := countModeRecords(buf); got != 2 {
		t.Errorf("mode records = %d, want 2 — the internal flip stands", got)
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru", "en"}) {
		t.Errorf("panel symbols = %q, want exactly [ru en]", got)
	}
	if consume := a.HandleKey(engine.EngineEvent{Keyval: uint32('g')}); consume {
		t.Error("EN-mode press consumed without a switcher — the internal mode must still govern typing")
	}
}

// TestActor_ReentrantAttachDuringFlip pins the live G-6-1 form: ibus-daemon
// answers a flip's SetGlobalEngine by synchronously calling CreateEngine on
// the daemon's factory, and the factory calls handler.AttachEngine — the
// seam closure here plays that whole chain. The flip must complete WITHOUT
// consuming its switchTimeout deadline; while AttachEngine still waited on
// the actor mutex, the re-entrant call deadlocked the flip goroutine and
// the deadline burned to the WARN ("engine created" landed ~1 ms after the
// abort on every flip of the diagnosing journal).
func TestActor_ReentrantAttachDuringFlip(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.SetSwitcher(func(_ context.Context, _ string) error {
		// The live re-entrant chain — SetGlobalEngine → CreateEngine →
		// AttachEngine — runs while the flip's await holds the actor mutex.
		a.AttachEngine(sink)

		return nil
	})

	tapShift(a) // one clean tap: the Single decision fires at window expiry
	start := time.Now()
	flipDone := make(chan struct{})
	go func() {
		defer close(flipDone)
		a.ExpiryAt(expiryAfterWindow) // EN → RU
	}()
	select {
	case <-flipDone:
	case <-time.After(flipBudget + 5*time.Second):
		t.Fatal("the flip never completed — the re-entrant AttachEngine deadlocked on the actor mutex (G-6-1)")
	}
	if elapsed := time.Since(start); elapsed >= flipBudget {
		t.Errorf("the flip consumed %v — at the switchTimeout deadline; the factory path must complete"+
			" without spending the deadline", elapsed)
	}
	if strings.Contains(buf.String(), `"msg":"engine switch failed"`) {
		t.Errorf("the flip WARNed — the SetGlobalEngine round trip did not complete inside the"+
			" deadline; log:\n%s", buf.String())
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru"}) {
		t.Errorf("panel symbols = %q, want exactly [ru] — flipTo must run past the re-entrant attach", got)
	}
}

// TestActor_AttachEngineWhileFlipInFlight pins the independent-call form of
// G-6-1: while the flip's SetGlobalEngine await is blocked mid-round-trip
// (holding the actor mutex), an AttachEngine from a foreign goroutine — the
// factory's dispatch goroutine in the live daemon — must still return; the
// factory path never waits on the actor mutex.
func TestActor_AttachEngineWhileFlipInFlight(t *testing.T) {
	a, _ := wiredActor()
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	a.SetSwitcher(func(_ context.Context, _ string) error {
		entered <- struct{}{}
		<-release // the await stuck mid-round-trip — the slow-ibus-daemon shape

		return nil
	})

	tapShift(a)
	flipDone := make(chan struct{})
	go func() {
		defer close(flipDone)
		a.ExpiryAt(expiryAfterWindow) // EN → RU — the await holds the actor mutex
	}()
	select {
	case <-entered:
	case <-time.After(flipBudget + 5*time.Second):
		t.Fatal("the flip never reached the switcher seam — the seam is not wired into the flip path")
	}

	liveSink := &fakeSink{}
	attachDone := make(chan struct{})
	go func() {
		defer close(attachDone)
		a.AttachEngine(liveSink) // the factory's re-entrant path, off the actor mutex
	}()
	select {
	case <-attachDone:
	case <-time.After(attachBudget):
		t.Fatalf("AttachEngine blocked past %v while the flip's SetGlobalEngine await is still in"+
			" flight — it waits on the actor mutex (G-6-1)", attachBudget)
	}
	close(release) // the seam is released only AFTER the attach assertion
	select {
	case <-flipDone:
	case <-time.After(flipBudget + 5*time.Second):
		t.Fatal("the flip never concluded after the seam's release")
	}
}

// TestStatus_ReportsActiveEngine pins the D-52 status surface: the snapshot
// names the ACTIVE engine — the mode's engine literal (a config constant,
// D-20), never user text.
func TestStatus_ReportsActiveEngine(t *testing.T) {
	a, _ := wiredActor()
	if st := a.StatusSnapshot(); st.Engine != engine.NameEN {
		t.Errorf("fresh snapshot engine = %q, want %s", st.Engine, engine.NameEN)
	}

	flipMode(a)
	if st := a.StatusSnapshot(); st.Engine != engine.NameRU {
		t.Errorf("post-flip snapshot engine = %q, want %s", st.Engine, engine.NameRU)
	}
}

// TestActor_SyncEngineFollowsEngine pins the sync semantics (05-04,
// criterion 3): an observed engine name matching the current mode is the
// daemon's OWN flip echoed back (spike P4) — a silent confirmation without
// a record; a drift corrects the internal mode with the byte-stable mode
// record (the e2e oracle) plus a WARN, and the panel symbol follows the
// corrected mode.
func TestActor_SyncEngineFollowsEngine(t *testing.T) {
	t.Run("same mode confirms without a record", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()

		a.SyncEngine(engine.NameEN) // EN at start: the daemon's own state echoed back

		logged := buf.String()
		if strings.Contains(logged, `"msg":"mode"`) {
			t.Errorf("same-mode sync wrote a mode record — the echo must confirm silently; log:\n%s", logged)
		}
		if strings.Contains(logged, `"level":"WARN"`) {
			t.Errorf("same-mode sync warned; log:\n%s", logged)
		}
		if st := a.StatusSnapshot(); st.Mode != "en" || st.Engine != engine.NameEN {
			t.Errorf("snapshot mode/engine = %q/%q, want en/%s", st.Mode, st.Engine, engine.NameEN)
		}
		if got := sink.modeSymbols(); len(got) != 0 {
			t.Errorf("same-mode sync emitted %v panel symbols, want none", got)
		}
	})

	t.Run("drift corrects with the mode record and a WARN", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()

		a.SyncEngine(engine.NameRU) // an external flip (indicator click): EN → RU

		logged := buf.String()
		if !strings.Contains(logged, `"msg":"mode","to":"ru"`) {
			t.Errorf("byte-stable mode record missing; log:\n%s", logged)
		}
		if !strings.Contains(logged, `"msg":"mode corrected"`) {
			t.Errorf("correction WARN missing; log:\n%s", logged)
		}
		if st := a.StatusSnapshot(); st.Mode != "ru" || st.Engine != engine.NameRU {
			t.Errorf("snapshot mode/engine = %q/%q, want ru/%s", st.Mode, st.Engine, engine.NameRU)
		}
		if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru"}) {
			t.Errorf("panel symbols after the correction = %q, want exactly [ru]", got)
		}
	})
}

// TestActor_SyncEngineForeignWarns pins the honest exit from under goswitch
// (T-05-04-03): a third-party engine name — another ibus engine, an xkb
// source — warns and leaves the mode untouched; the daemon owns no state it
// cannot observe.
func TestActor_SyncEngineForeignWarns(t *testing.T) {
	for _, name := range []string{"anthy", "xkb:fr"} {
		t.Run(name, func(t *testing.T) {
			buf := captureLogs(t)
			a, sink := wiredActor()

			a.SyncEngine(name)

			logged := buf.String()
			if !strings.Contains(logged, `"msg":"foreign engine"`) {
				t.Errorf("foreign-engine WARN missing; log:\n%s", logged)
			}
			if strings.Contains(logged, `"msg":"mode"`) {
				t.Errorf("foreign sync moved the mode; log:\n%s", logged)
			}
			if st := a.StatusSnapshot(); st.Mode != "en" {
				t.Errorf("mode = %q, want en — foreign names must not touch the state", st.Mode)
			}
			if got := sink.modeSymbols(); len(got) != 0 {
				t.Errorf("foreign sync emitted %v panel symbols, want none", got)
			}
		})
	}
}

// TestActor_SyncEngineNeverSwitches pins the flip-loop prohibition
// (T-05-04-01): NO sync input ever reaches the switcher — the correction
// moves the internal mode only; the daemon stays the single writer of the
// bus, the shell's indicator clicks are followed, never fought.
func TestActor_SyncEngineNeverSwitches(t *testing.T) {
	a, _ := wiredActor()
	probe := &switchProbe{}
	a.SetSwitcher(probe.switcher)

	a.SyncEngine(engine.NameEN)
	a.SyncEngine(engine.NameRU)
	a.SyncEngine("anthy")

	if got := probe.targets(); len(got) != 0 {
		t.Errorf("SyncEngine called the switcher with %q — a flip loop is built on this; want zero calls", got)
	}
}

// The op-log labels of the display-order corpus: the sink's modeHook records
// the panel-symbol emit, fakeDisplay the display call — the D-36 order reads
// through these exact strings.
const (
	opSymbolRU  = "symbol:ru"
	opDisplayRU = "display:ru"
	opSymbolEN  = "symbol:en"
	opDisplayEN = "display:en"
)

// fakeDisplay is the actor's ModeDisplay double (quick plan 260930-pf6):
// every ModeChanged recorded under a mutex, with an optional hook the order
// pins interleave against the sink's modeHook.
type fakeDisplay struct {
	mu     sync.Mutex
	calls  []string
	onCall func(symbol string)
}

// ModeChanged records the symbol and plays the hook.
func (d *fakeDisplay) ModeChanged(symbol string) {
	d.mu.Lock()
	d.calls = append(d.calls, symbol)
	hook := d.onCall
	d.mu.Unlock()
	if hook != nil {
		hook(symbol)
	}
}

// symbols snapshots the recorded sequence.
func (d *fakeDisplay) symbols() []string {
	d.mu.Lock()
	defer d.mu.Unlock()

	return append([]string(nil), d.calls...)
}

// TestActor_SetModeDisplayPushesCurrentMode pins the install contract: the
// display receives the CURRENT mode at SetModeDisplay — startup with the
// initial mode AND late install after flips — and every later flip reaches
// it.
func TestActor_SetModeDisplayPushesCurrentMode(t *testing.T) {
	a, _ := wiredActor()
	disp := &fakeDisplay{}

	a.SetModeDisplay(disp)
	if got, want := disp.symbols(), []string{"en"}; !slices.Equal(got, want) {
		t.Errorf("install push = %q, want exactly [en] — the display must self-sync at install", got)
	}

	flipMode(a) // EN → RU
	if got, want := disp.symbols(), []string{"en", "ru"}; !slices.Equal(got, want) {
		t.Errorf("symbols after the flip = %q, want [en ru]", got)
	}

	// Late install on a drifted actor: the fresh display gets the CURRENT
	// mode, not the startup default.
	flipMode(a) // RU → EN
	late := &fakeDisplay{}
	a.SetModeDisplay(late)
	if got, want := late.symbols(), []string{"en"}; !slices.Equal(got, want) {
		t.Errorf("late install push = %q, want exactly [en]", got)
	}
}

// TestActor_FlipInvokesDisplayLast extends the D-36 order pin: the display
// observer fires strictly AFTER the panel-symbol emit — mode record →
// switcher → panel symbol → display.
func TestActor_FlipInvokesDisplayLast(t *testing.T) {
	a, sink := wiredActor()
	var opMu sync.Mutex
	var ops []string
	sink.modeHook = func(symbol string) {
		opMu.Lock()
		ops = append(ops, "symbol:"+symbol)
		opMu.Unlock()
	}
	disp := &fakeDisplay{}
	disp.onCall = func(symbol string) {
		opMu.Lock()
		ops = append(ops, "display:"+symbol)
		opMu.Unlock()
	}
	a.SetModeDisplay(disp)
	opMu.Lock()
	ops = ops[:0] // the install push is pinned elsewhere; only the flip order matters here
	opMu.Unlock()

	flipMode(a) // EN → RU

	opMu.Lock()
	defer opMu.Unlock()
	want := []string{opSymbolRU, opDisplayRU}
	if !slices.Equal(ops, want) {
		t.Errorf("flip op order = %q, want exactly %q — the display fires last", ops, want)
	}
}

// TestActor_ToggleModeFlips pins the public toggle gesture (quick plan
// 261001-fg3 — the tray menu's first item and the item's SNI Activate land
// here): ToggleMode drives the SAME flipTo execution path as every other
// gesture — two toggles yield the byte-stable mode records to=ru then to=en,
// switch targets exactly goswitch-ru then goswitch-en, panel symbols ru then
// en, and the ops strictly symbol:ru → display:ru → symbol:en → display:en
// (the observer LAST — the D-36 order intact).
func TestActor_ToggleModeFlips(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	probe := &switchProbe{}
	var opMu sync.Mutex
	var ops []string
	sink.modeHook = func(symbol string) {
		opMu.Lock()
		ops = append(ops, "symbol:"+symbol)
		opMu.Unlock()
	}
	disp := &fakeDisplay{}
	disp.onCall = func(symbol string) {
		opMu.Lock()
		ops = append(ops, "display:"+symbol)
		opMu.Unlock()
	}
	a.SetSwitcher(probe.switcher)
	a.SetModeDisplay(disp)
	opMu.Lock()
	ops = ops[:0] // the install push is pinned elsewhere; only the toggle order matters here
	opMu.Unlock()

	a.ToggleMode() // EN → RU
	a.ToggleMode() // RU → EN

	if got := probe.targets(); !slices.Equal(got, []string{engine.NameRU, engine.NameEN}) {
		t.Errorf("switch targets = %q, want exactly [%s %s] — the target mode's engine per toggle",
			got, engine.NameRU, engine.NameEN)
	}
	logged := buf.String()
	if !strings.Contains(logged, `"msg":"mode","to":"ru"`) || !strings.Contains(logged, `"msg":"mode","to":"en"`) {
		t.Errorf("byte-stable mode records missing; log:\n%s", logged)
	}
	if strings.Index(logged, `"to":"ru"`) > strings.Index(logged, `"to":"en"`) {
		t.Errorf("mode records out of order (ru must precede en); log:\n%s", logged)
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru", "en"}) {
		t.Errorf("panel symbols = %q, want exactly [ru en] — one emit per toggle", got)
	}
	opMu.Lock()
	defer opMu.Unlock()
	want := []string{opSymbolRU, opDisplayRU, opSymbolEN, opDisplayEN}
	if !slices.Equal(ops, want) {
		t.Errorf("toggle op order = %q, want exactly %q — the display fires last", ops, want)
	}
}

// TestActor_SyncDriftInvokesDisplay mirrors the order pin for syncMode: a
// drift correction reaches the display too — the icon follows the FACTUAL
// engine, not only the daemon's own flips.
func TestActor_SyncDriftInvokesDisplay(t *testing.T) {
	a, sink := wiredActor()
	var opMu sync.Mutex
	var ops []string
	sink.modeHook = func(symbol string) {
		opMu.Lock()
		ops = append(ops, "symbol:"+symbol)
		opMu.Unlock()
	}
	disp := &fakeDisplay{}
	disp.onCall = func(symbol string) {
		opMu.Lock()
		ops = append(ops, "display:"+symbol)
		opMu.Unlock()
	}
	a.SetModeDisplay(disp)
	opMu.Lock()
	ops = ops[:0] // the install push is pinned elsewhere; only the drift order matters here
	opMu.Unlock()

	a.SyncEngine(engine.NameRU) // an external flip: EN → RU

	opMu.Lock()
	defer opMu.Unlock()
	want := []string{opSymbolRU, opDisplayRU}
	if !slices.Equal(ops, want) {
		t.Errorf("drift op order = %q, want exactly %q — the display fires last on sync too", ops, want)
	}
}

// The menu-sync corpus of plan 07-05: the tray-menu state seam — SwitchMode
// (the radio pair's flip through the SAME flipTo path) and the observer
// pushes (the mode on every flip/sync observer-last; the applied config
// truth on every snapshot fold).

// fakeMenuSync is the actor's MenuSync double (plan 07-05): every push
// recorded under a mutex, with an op-log hook the order pins interleave
// against the display's.
type fakeMenuSync struct {
	mu          sync.Mutex
	modes       []string
	ac          []bool
	sound       []bool
	keys        []string
	corrections []int
	blocklist   []blocklistPush
	onMode      func(symbol string)
}

// SetMode records the mode push.
func (f *fakeMenuSync) SetMode(symbol string) {
	f.mu.Lock()
	f.modes = append(f.modes, symbol)
	hook := f.onMode
	f.mu.Unlock()
	if hook != nil {
		hook(symbol)
	}
}

// SetAutocorrectEnabled records the applied autocorrect push.
func (f *fakeMenuSync) SetAutocorrectEnabled(on bool) {
	f.mu.Lock()
	f.ac = append(f.ac, on)
	f.mu.Unlock()
}

// SetSoundEnabled records the applied sound push.
func (f *fakeMenuSync) SetSoundEnabled(on bool) {
	f.mu.Lock()
	f.sound = append(f.sound, on)
	f.mu.Unlock()
}

// SetKeys records the raw key names push.
func (f *fakeMenuSync) SetKeys(tap, combo, chord string) {
	f.mu.Lock()
	f.keys = append(f.keys, tap+"|"+combo+"|"+chord)
	f.mu.Unlock()
}

// SetCorrections records the corrections-counter push (owner UAT
// 2026-10-05: the Status info row renders the live count).
func (f *fakeMenuSync) SetCorrections(n int) {
	f.mu.Lock()
	f.corrections = append(f.corrections, n)
	f.mu.Unlock()
}

// blocklistPush is one recorded SetBlocklistApp push — the blocklist menu
// item's full state triple (phase 11, MENU-01).
type blocklistPush struct {
	app     string
	fresh   bool
	covered bool
	full    bool
}

// SetBlocklistApp records the blocklist-item state push (MENU-01: the item
// lives from the identity's freshness — gain pushes fresh, the tracked
// surface's own blur (the empty identity) clears it).
func (f *fakeMenuSync) SetBlocklistApp(app string, fresh, covered, full bool) {
	f.mu.Lock()
	f.blocklist = append(f.blocklist, blocklistPush{app, fresh, covered, full})
	f.mu.Unlock()
}

// menuModes snapshots the recorded mode pushes.
func (f *fakeMenuSync) menuModes() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]string(nil), f.modes...)
}

// menuAC snapshots the recorded autocorrect pushes.
func (f *fakeMenuSync) menuAC() []bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]bool(nil), f.ac...)
}

// menuSound snapshots the recorded sound pushes.
func (f *fakeMenuSync) menuSound() []bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]bool(nil), f.sound...)
}

// menuKeys snapshots the recorded key-name pushes.
func (f *fakeMenuSync) menuKeys() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]string(nil), f.keys...)
}

// menuCorrections snapshots the recorded corrections-counter pushes.
func (f *fakeMenuSync) menuCorrections() []int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]int(nil), f.corrections...)
}

// menuBlocklist snapshots the recorded blocklist-item pushes.
func (f *fakeMenuSync) menuBlocklist() []blocklistPush {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]blocklistPush(nil), f.blocklist...)
}

// blTestApp is the blocklist corpus's test identity — the goconst form
// (the literal repeats across the freshness/fold/accessor corpus).
const blTestApp = "gedit"

// TestActor_BlocklistItemFreshPushes pins the freshness machine's PUSH
// half (MENU-01, D-03 «демон не угадывает»): the install push reports the
// initial UNKNOWN (fresh=false, no name) and a PushFocusedApp gain pushes
// the identity with fresh=true.
func TestActor_BlocklistItemFreshPushes(t *testing.T) {
	t.Run("install push reports the initial unknown", func(t *testing.T) {
		a, _ := wiredActor()
		ms := &fakeMenuSync{}
		a.SetMenuSync(ms)

		pushes := ms.menuBlocklist()
		if len(pushes) != 1 {
			t.Fatalf("install pushes = %d, want exactly 1 (the startup self-sync)", len(pushes))
		}
		if pushes[0] != (blocklistPush{}) {
			t.Errorf("install push = %+v, want the zero unknown state (D-03)", pushes[0])
		}
	})

	t.Run("gain pushes the fresh identity", func(t *testing.T) {
		a, _ := wiredActor()
		ms := &fakeMenuSync{}
		a.SetMenuSync(ms)

		a.PushFocusedApp(blTestApp)

		pushes := ms.menuBlocklist()
		if len(pushes) != 2 {
			t.Fatalf("pushes after the gain = %d, want install + gain", len(pushes))
		}
		want := blocklistPush{app: blTestApp, fresh: true}
		if pushes[1] != want {
			t.Errorf("gain push = %+v, want %+v", pushes[1], want)
		}
	})
}

// TestActor_BlocklistItemFreshLosses pins the freshness machine's LOSS
// half (plan 11-05, the VM live finding): the tracked surface's OWN blur —
// the empty identity from the a11y loss report — de-arms and pushes the
// unknown state (the stale cache name never reaches the menu); the
// ENGINE's focus churn — real or synthetic — never touches the freshness
// (the click that grants a fresh window its focus pairs a real focus_out
// with the focus_in, and that churn trails the a11y gain: a reset here
// erased freshness the same surface still held). A loss on an
// already-unknown state is the silent no-op.
func TestActor_BlocklistItemFreshLosses(t *testing.T) {
	t.Run("blur loss de-arms and pushes the unknown", func(t *testing.T) {
		a, _ := wiredActor()
		ms := &fakeMenuSync{}
		a.SetMenuSync(ms)

		a.PushFocusedApp(blTestApp)
		a.PushFocusedApp("") // the tracked surface's blur (the a11y loss report)

		pushes := ms.menuBlocklist()
		if len(pushes) != 3 {
			t.Fatalf("pushes after the blur loss = %d, want install + gain + de-arm", len(pushes))
		}
		if pushes[2].fresh {
			t.Errorf("post-loss fresh = true, want false — freshness died with the surface")
		}
		if pushes[2].app != "" {
			t.Errorf("post-loss name = %q, want empty — "+
				"the stale cache name must not reach the menu (D-03)", pushes[2].app)
		}
	})

	t.Run("engine focus churn keeps freshness", func(t *testing.T) {
		a, _ := wiredActor()
		ms := &fakeMenuSync{}
		a.SetMenuSync(ms)
		a.SetSwitcher(func(_ context.Context, _ string) error { return nil })

		a.PushFocusedApp(blTestApp)
		flipMode(a)                                 // the successful round trip arms exactly one credit
		a.HandleLifecycle(engine.LifecycleFocusOut) // the synthetic pair's FocusOut consumes it
		a.HandleLifecycle(engine.LifecycleFocusOut) // a REAL loss — the engine has no freshness vote

		pushes := ms.menuBlocklist()
		if len(pushes) != 2 {
			t.Fatalf("the engine churn pushed %d extra states, want none; pushes: %+v",
				len(pushes)-2, pushes)
		}
		if !pushes[1].fresh {
			t.Errorf("freshness after the engine churn = false, want the gain's true — " +
				"the de-arm is the a11y blur's, never the engine's (11-05)")
		}
	})

	t.Run("loss on the unknown state is the silent no-op", func(t *testing.T) {
		a, _ := wiredActor()
		ms := &fakeMenuSync{}
		a.SetMenuSync(ms)

		a.PushFocusedApp("") // nothing armed — nothing to de-arm

		pushes := ms.menuBlocklist()
		if len(pushes) != 1 {
			t.Fatalf("pushes after the idle loss = %d, want exactly the install push", len(pushes))
		}
	})
}

// TestActor_BlocklistFoldCovered pins the fold's covered push moment ((в)
// of the RESEARCH push moments): a document whose apps_blocklist GREW
// recomputes covered and pushes SetBlocklistApp WITHOUT any click — the
// checkmark follows the document. The covered match is the GATE's
// discipline: any compiled pattern matched by re.MatchString against the
// FRESH identity.
func TestActor_BlocklistFoldCovered(t *testing.T) {
	a, _ := wiredActor()
	ms := &fakeMenuSync{}
	a.SetMenuSync(ms)
	src := &reloadSource{cfg: config.Defaults()}
	a.AttachConfig(src)

	a.PushFocusedApp(blTestApp)
	last := ms.menuBlocklist()
	if len(last) == 0 || last[len(last)-1].covered {
		t.Fatalf("pre-fold push = %+v, want covered=false before the document grows", last)
	}

	cfg := config.Defaults()
	cfg.Autocorrect.AppsBlocklist = []string{blTestApp}
	src.set(cfg)
	a.FoldAppliedConfig()

	pushes := ms.menuBlocklist()
	if len(pushes) != len(last)+1 {
		t.Fatalf("the fold pushed %d states (had %d), want exactly one more", len(pushes)-len(last), len(last))
	}
	want := blocklistPush{app: blTestApp, fresh: true, covered: true}
	if pushes[len(pushes)-1] != want {
		t.Errorf("fold push = %+v, want %+v — covered without a click", pushes[len(pushes)-1], want)
	}
}

// TestActor_BlocklistFoldFull pins the fold's full push moment: a
// ceiling-length document flips full WITHOUT a click — the visible fill
// state follows the document (S9), the identity matching no pattern stays
// uncovered.
func TestActor_BlocklistFoldFull(t *testing.T) {
	a, _ := wiredActor()
	ms := &fakeMenuSync{}
	a.SetMenuSync(ms)
	src := &reloadSource{cfg: config.Defaults()}
	a.AttachConfig(src)

	a.PushFocusedApp("org.gnome.other")

	full := make([]string, config.MaxAutocorrectBlocklist)
	for i := range full {
		full[i] = fmt.Sprintf("pattern-%02d", i)
	}
	cfg := config.Defaults()
	cfg.Autocorrect.AppsBlocklist = full
	src.set(cfg)
	a.FoldAppliedConfig()

	pushes := ms.menuBlocklist()
	if len(pushes) == 0 {
		t.Fatal("no pushes recorded")
	}
	last := pushes[len(pushes)-1]
	if !last.full {
		t.Errorf("fold push = %+v, want full=true at the 64-pattern document", last)
	}
	if last.covered {
		t.Errorf("fold push = %+v, want covered=false — the identity matches no pattern", last)
	}
}

// TestActor_BlocklistFoldFreshnessBeatsCache pins the fold's D-03 side: a
// blur loss (the tracked surface's own) cleared the freshness, so the fold
// over a document with the identity's pattern pushes the UNKNOWN state
// anyway — freshness, not the cache, decides.
func TestActor_BlocklistFoldFreshnessBeatsCache(t *testing.T) {
	a, _ := wiredActor()
	ms := &fakeMenuSync{}
	a.SetMenuSync(ms)
	src := &reloadSource{cfg: config.Defaults()}
	a.AttachConfig(src)

	a.PushFocusedApp(blTestApp)
	a.PushFocusedApp("") // the blur loss clears freshness

	cfg := config.Defaults()
	cfg.Autocorrect.AppsBlocklist = []string{blTestApp}
	src.set(cfg)
	a.FoldAppliedConfig()

	pushes := ms.menuBlocklist()
	if len(pushes) == 0 {
		t.Fatal("no pushes recorded")
	}
	last := pushes[len(pushes)-1]
	if last.fresh || last.covered || last.app != "" {
		t.Errorf("fold push after the real FocusOut = %+v, want the unknown state — "+
			"freshness, not the cache, decides", last)
	}
}

// TestActor_BlocklistAppState pins the composition accessor's identity
// half (the ONLY state channel of the click composition, plan 11-04):
// known requires FRESHNESS — a stale cache name never reads as known and
// never leaks to the composition (D-03).
func TestActor_BlocklistAppState(t *testing.T) {
	t.Run("unknown before any gain", func(t *testing.T) {
		a, _ := wiredActor()

		app, known, covered, full := a.BlocklistAppState()
		if app != "" || known || covered || full {
			t.Errorf("initial state = (%q, %t, %t, %t), want the zero unknown", app, known, covered, full)
		}
	})

	t.Run("known requires freshness", func(t *testing.T) {
		a, _ := wiredActor()
		a.PushFocusedApp(blTestApp)
		a.PushFocusedApp("") // the blur loss

		app, known, _, _ := a.BlocklistAppState()
		if known || app != "" {
			t.Errorf("state after the blur loss = (%q, %t), want unknown — "+
				"the stale name must not leak to the composition", app, known)
		}
	})

	t.Run("known fresh uncovered", func(t *testing.T) {
		a, _ := wiredActor()
		a.PushFocusedApp(blTestApp)

		app, known, covered, full := a.BlocklistAppState()
		if app != blTestApp || !known || covered || full {
			t.Errorf("state = (%q, %t, %t, %t), want (%s, true, false, false)",
				app, known, covered, full, blTestApp)
		}
	})
}

// TestActor_BlocklistAppFoldedState pins the composition accessor's
// document half: covered rides the gate's match discipline against the
// compiled patterns, full rides the document length at the ceiling.
func TestActor_BlocklistAppFoldedState(t *testing.T) {
	t.Run("covered by the document's pattern", func(t *testing.T) {
		a, _ := wiredActor()
		src := &reloadSource{cfg: config.Defaults()}
		a.AttachConfig(src)
		a.PushFocusedApp(blTestApp)
		cfg := config.Defaults()
		cfg.Autocorrect.AppsBlocklist = []string{"^" + blTestApp + "$"}
		src.set(cfg)
		a.FoldAppliedConfig()

		_, known, covered, _ := a.BlocklistAppState()
		if !known || !covered {
			t.Errorf("state = (known %t, covered %t), want known+covered by the anchored pattern", known, covered)
		}
	})

	t.Run("full at the ceiling document", func(t *testing.T) {
		a, _ := wiredActor()
		src := &reloadSource{cfg: config.Defaults()}
		a.AttachConfig(src)
		a.PushFocusedApp("org.gnome.other")

		full := make([]string, config.MaxAutocorrectBlocklist)
		for i := range full {
			full[i] = fmt.Sprintf("pattern-%02d", i)
		}
		cfg := config.Defaults()
		cfg.Autocorrect.AppsBlocklist = full
		src.set(cfg)
		a.FoldAppliedConfig()

		_, known, covered, fullState := a.BlocklistAppState()
		if !known || covered || !fullState {
			t.Errorf("state = (known %t, covered %t, full %t), want known, uncovered, full", known, covered, fullState)
		}
	})
}

// TestActor_EnsureAppidUnconditional pins the Pitfall-2 cure: the lazy
// start's condition (a non-empty macr list OR enabled autocorrect) must NOT
// gate the blocklist item's observer — with autocorrect OFF (the D-54
// default) and an EMPTY macr list, EnsureAppid starts the observer anyway,
// or the item would stay dead forever on a default install.
func TestActor_EnsureAppidUnconditional(t *testing.T) {
	a, _ := wiredActor()
	starts := 0
	a.UseAppidStarter(func() (session.AppidSource, error) {
		starts++

		return fakeAppid{app: "org.gnome.Zenity"}, nil
	})
	// AutoCorrectEnabled=false and MACR empty — the ensureAppid condition
	// would refuse.

	a.EnsureAppid()

	if starts != 1 {
		t.Fatalf("observer starts after EnsureAppid = %d, want exactly 1 — "+
			"the item must live without autocorrect (Pitfall 2)", starts)
	}
	// Idempotent: the once-per-actor discipline stands.
	a.EnsureAppid()
	if starts != 1 {
		t.Errorf("observer starts after the second EnsureAppid = %d, want still 1", starts)
	}
}

// TestActor_SwitchModeFlipsToTarget pins the radio pair's gesture (plan
// 07-05): SwitchMode drives the SAME flipTo execution path — the target
// mode's switch, the byte-stable mode record, one panel symbol — and the
// same-target case lands on flipTo's no-op guard: zero extra records, zero
// extra switch calls.
func TestActor_SwitchModeFlipsToTarget(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	probe := &switchProbe{}
	a.SetSwitcher(probe.switcher)

	a.SwitchMode("ru")
	a.SwitchMode("ru") // the same-target no-op

	if got := probe.targets(); !slices.Equal(got, []string{engine.NameRU}) {
		t.Errorf("switch targets = %q, want exactly [%s] — one flip to the named target", got, engine.NameRU)
	}
	logged := buf.String()
	if strings.Count(logged, `"msg":"mode","to":"ru"`) != 1 {
		t.Errorf("mode records = %q, want exactly one to=ru (the same-target click is a no-op)", logged)
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru"}) {
		t.Errorf("panel symbols = %q, want exactly [ru]", got)
	}
}

// TestActor_MenuSyncSetModeOnFlip pins the mode push of the menu seam: the
// install receives the CURRENT mode (the SetModeDisplay mirror), every flip
// and every sync drift pushes the new symbol observer-last — after the
// display, exactly as ModeChanged (the D-36 order with both observers
// appended).
func TestActor_MenuSyncSetModeOnFlip(t *testing.T) {
	a, _ := wiredActor()
	var opMu sync.Mutex
	var ops []string
	record := func(op string) {
		opMu.Lock()
		ops = append(ops, op)
		opMu.Unlock()
	}
	msync := &fakeMenuSync{}
	msync.onMode = func(symbol string) { record("menu:" + symbol) }
	disp := &fakeDisplay{}
	disp.onCall = func(symbol string) { record("display:" + symbol) }

	// The install self-syncs: the menu receives the CURRENT mode, exactly
	// like the display (the SetModeDisplay mirror).
	a.SetMenuSync(msync)
	a.SetModeDisplay(disp)
	if got := msync.menuModes(); !slices.Equal(got, []string{"en"}) {
		t.Errorf("install mode pushes = %q, want exactly [en] — the current mode", got)
	}

	opMu.Lock()
	ops = nil // the install pushes are pinned above; only the change order matters
	opMu.Unlock()

	a.ToggleMode()              // EN → RU
	a.SyncEngine(engine.NameEN) // a drift back: RU → EN

	if got := msync.menuModes(); !slices.Equal(got, []string{"en", "ru", "en"}) {
		t.Errorf("menu mode pushes = %q, want [en ru en] — the install plus one per flip and per drift", got)
	}
	opMu.Lock()
	defer opMu.Unlock()
	want := []string{opDisplayRU, "menu:ru", opDisplayEN, "menu:en"}
	if !slices.Equal(ops, want) {
		t.Errorf("menu op order = %q, want exactly %q — the menu fires AFTER the display, observer-last", ops, want)
	}
}

// TestActor_MenuSyncApplySnapshotPushes pins the fold push: applySnapshot
// hands the menu the APPLIED config truth — the autocorrect switch, the
// raw key names and the corrections counter (the Status row's live value,
// owner UAT 2026-10-05) — once per fold, whatever the document says.
func TestActor_MenuSyncApplySnapshotPushes(t *testing.T) {
	a, _ := wiredActor()
	cfg := config.Defaults()
	cfg.Autocorrect.Enabled = true
	src := &reloadSource{cfg: cfg}
	a.AttachConfig(src)
	msync := &fakeMenuSync{}
	a.SetMenuSync(msync)

	a.ExpiryAt(expiryAfterWindow) // one applySnapshot fold

	if got := msync.menuAC(); !slices.Equal(got, []bool{true}) {
		t.Errorf("autocorrect pushes = %v, want exactly [true]", got)
	}
	if got := msync.menuKeys(); !slices.Equal(got, []string{"shift_r|shift+ctrl_r|super+space"}) {
		t.Errorf("key pushes = %q, want the raw config names of the defaults document", got)
	}
	if got := msync.menuCorrections(); !slices.Equal(got, []int{0, 0}) {
		t.Errorf("corrections pushes = %v, want the install push then the fold push, both 0", got)
	}

	// The document flips the switch off: the next fold pushes the applied
	// false (the actor hands CONFIG truth, never a cache).
	cfg.Autocorrect.Enabled = false
	src.set(cfg)
	a.ExpiryAt(expiryAfterWindow)
	if got := msync.menuAC(); !slices.Equal(got, []bool{true, false}) {
		t.Errorf("autocorrect pushes after the flip = %v, want exactly [true false]", got)
	}
}

// TestActor_MenuSyncApplySnapshotPushesSound pins the sound push (plan
// 07-05, the owner's default-ON verdict): a document WITHOUT the sound
// section pushes true (the EffectiveEnabled reading), an explicit
// enabled:false pushes false, and the status snapshot carries the same
// truth for the `sound_enabled` token.
func TestActor_MenuSyncApplySnapshotPushesSound(t *testing.T) {
	a, _ := wiredActor()
	cfg := config.Defaults()
	cfg.Sound = config.Sound{} // the absent section: decodes as enabled (default ON)
	src := &reloadSource{cfg: cfg}
	a.AttachConfig(src)
	msync := &fakeMenuSync{}
	a.SetMenuSync(msync)

	a.ExpiryAt(expiryAfterWindow)

	if got := msync.menuSound(); !slices.Equal(got, []bool{true}) {
		t.Errorf("sound pushes = %v, want exactly [true] — the absent section reads enabled", got)
	}
	if st := a.StatusSnapshot(); !st.SoundEnabled {
		t.Error("status snapshot SoundEnabled = false after the fold, want true")
	}

	off := false
	cfg.Sound = config.Sound{Enabled: &off}
	src.set(cfg)
	a.ExpiryAt(expiryAfterWindow)
	if got := msync.menuSound(); !slices.Equal(got, []bool{true, false}) {
		t.Errorf("sound pushes after the explicit off = %v, want exactly [true false]", got)
	}
	if st := a.StatusSnapshot(); st.SoundEnabled {
		t.Error("status snapshot SoundEnabled = true after the explicit off, want false")
	}
}

// TestActor_NilMenuSyncNoOp pins the nil-seam degradation: without
// SetMenuSync every gesture works and every push point is a no-op — zero
// panics, zero behavior drift.
func TestActor_NilMenuSyncNoOp(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	a.AttachConfig(&reloadSource{cfg: config.Defaults()})

	a.ToggleMode()
	a.ExpiryAt(expiryAfterWindow)

	if !strings.Contains(buf.String(), `"msg":"mode","to":"ru"`) {
		t.Errorf("the flip did not happen without a menu sync; log:\n%s", buf.String())
	}
}

// The autocorrect corpus of plan 06-06 (blocklist re-pinned by 07-02):
// the word boundary of feedKey, the cheap-gate arming and the fail-closed
// silence. The fixture app is a bridge-namespace literal distinct from
// the MACR corpus names.

// acListedApp is the fixture app identity of the autocorrect corpus.
const acListedApp = "org.gnome.gedit"

// acBlockMiss is the corpus's blocklist pattern: an ANCHORED prefix that
// never matches acListedApp — the fire paths' Options (a blocklist entry
// that forbids nothing on this corpus).
const acBlockMiss = `^com\.google\.Chrome`

// The detector thresholds of the corpus — the config 06-04 defaults (the
// detect.DefaultParams mirror: change the places together).
const (
	acMinWordLen    = 4
	acTrigramMargin = 2.0
	acTrigramFloor  = 1.0
)

// acPollBudget bounds every eventually poll of the async confirm — orders
// of magnitude above the goroutine's real latency, far below any test
// timeout.
const acPollBudget = 2 * time.Second

// acOptions is the enabled autocorrect Options of the corpus: a blocklist
// pattern that does NOT match the fixture app, so every fire path fires.
func acOptions() session.Options {
	return session.Options{
		AutoCorrectEnabled:    true,
		AutoCorrectBlocklist:  []string{acBlockMiss},
		AutoCorrectMinWordLen: acMinWordLen,
		AutoCorrectMargin:     acTrigramMargin,
		AutoCorrectFloor:      acTrigramFloor,
	}
}

// acBlockedOptions is the blocklist-match cell's configuration: the
// pattern is a bare SUBSTRING of the fixture identity (no anchors — the
// MatchString substring semantics pinned at the actor level).
func acBlockedOptions() session.Options {
	opts := acOptions()
	opts.AutoCorrectBlocklist = []string{"gedit"}

	return opts
}

// eventually polls cond until it holds or the budget lapses — the poll-
// until idiom of the appid corpus, never a fixed sleep: the autocorrect
// confirm runs on its own goroutine, so its visible effects land
// asynchronously.
func eventually(t *testing.T, cond func() bool, msg string) {
	t.Helper()
	for deadline := time.Now().Add(acPollBudget); time.Now().Before(deadline); {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal(msg)
}

// spaceKey is the boundary carrier of the corpus (the plan's
// geometrically-deterministic fires vehicle; the reset keys pin the
// arming-before-reset instead).
func spaceKey() engine.EngineEvent {
	return engine.EngineEvent{Keyval: uint32(' ')}
}

// fireAutocorrect drives one full autocorrect round: install the
// allowed-role double, type the token, land the separator boundary, settle
// the armed pipeline's pre-correction verify with the freshest push, then
// settle the verify-after round — the same dance the Double corpus
// performs, driven asynchronously (the confirm goroutine).
func fireAutocorrect(t *testing.T, a *session.Actor, sink *fakeSink, token, converted string) {
	t.Helper()
	fireAutocorrectRole(t, a, sink, token, converted, acRoleAllowed)
}

// fireAutocorrectRole is fireAutocorrect with the confirm's live role
// verdict chosen by the caller: acRoleAllowed (61 — the ambiguous text box)
// for the known-identity fire paths, acRoleAllowedEntry (79) where the
// identity stays UNKNOWN (the CR-01 tightening refuses 61 without one).
func fireAutocorrectRole(t *testing.T, a *session.Actor, sink *fakeSink, token, converted string, role uint32) {
	t.Helper()
	a.UseRole(&fakeRole{role: role})
	typeWord(a, token)
	if a.HandleKey(spaceKey()) {
		t.Fatal("the boundary separator was consumed — the decision never changes consumption at the boundary")
	}
	eventually(t, func() bool { return sink.requireCount() >= 1 },
		"the pipeline never armed after the boundary (no RequireSurroundingText)")
	tokenEnd := uint32(len([]rune(token)) + 1)
	a.HandleSurroundingText(token+" ", tokenEnd, tokenEnd)
	a.HandleSurroundingText("", 0, 0) // the deletion proof completes the ladder (plan 10-11)
	eventually(t, func() bool { return sink.requireCount() >= 2 },
		"the verify-after round never armed (the correction never executed)")
	convEnd := uint32(len([]rune(converted)) + 1)
	a.HandleSurroundingText(converted+" ", convEnd, convEnd)
}

// TestActor_AutocorrectDisabledSilent pins the D-54 off invariant: with the
// zero Options (the default configuration) a word boundary is a NO-OP —
// the same keystrokes leave the same emitter trace as before the feature
// existed and no counter moves.
func TestActor_AutocorrectDisabledSilent(t *testing.T) {
	a, sink := wiredActor()
	typeWord(a, wordEN)
	if a.HandleKey(spaceKey()) {
		t.Fatal("the separator was consumed with autocorrect off")
	}
	if ops := sink.opLog(); len(ops) != 0 {
		t.Errorf("emitter ops with autocorrect off = %v, want none — the off state is byte-as-today", ops)
	}
	if got := a.AutoCorrectCounters(); got.Fired != 0 || got.Abstained != 0 || len(got.Reasons) != 0 {
		t.Errorf("counters with autocorrect off = %+v, want all zero — the off state counts nothing (D-54)", got)
	}
}

// TestActor_AutocorrectBoundaryArms pins the armed decision: with the
// cheap gates green (enabled, the listed focused app, the
// surrounding-text cap, a confident detector verdict) the separator
// boundary launches THE one correction pipeline — the key transits, the
// level-1 ladder deletes exactly token+tail and commits the conversion
// plus the tail.
func TestActor_AutocorrectBoundaryArms(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.SetOptions(acOptions())

	fireAutocorrect(t, a, sink, wordEN, wordRU)

	deletes := sink.deleteCalls()
	if len(deletes) != 1 || deletes[0].offset != -7 || deletes[0].nchars != 7 {
		t.Fatalf("deletes = %+v, want exactly [{-7 7}] — the range is token+tail", deletes)
	}
	a.HandleSurroundingText("", 0, 0) // the deletion proof completes the ladder (plan 10-11)
	if commits := sink.commitTexts(); len(commits) != 1 || commits[0] != wordRU+" " {
		t.Fatalf("commits = %q, want [%q] — the conversion plus the tail", commits, wordRU+" ")
	}
	if got := a.AutoCorrectCounters(); got.Abstained != 0 {
		t.Errorf("abstained after a fired boundary = %d, want 0", got.Abstained)
	}
}

// TestActor_AutocorrectEarlyVerifyAnswerWaits pins the 06-07 live finding
// (06-RESEARCH/live GTK4 fixture): the boundary fires ON the separator
// keypress, and the pre-correction verify's RequireSurroundingText is
// served by the client BEFORE the separator key reaches it — by wire
// ordering the first verify answer is necessarily the PRE-boundary state
// (token without the tail). That answer is an early report, not a
// mismatch: the round must stay open within its budget and settle on the
// post-boundary push (the client applying the separator). The old shape —
// mismatch on the first answer — killed every fired correction on such
// clients with verify-mismatch before the real state ever arrived.
func TestActor_AutocorrectEarlyVerifyAnswerWaits(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.UseRole(&fakeRole{role: acRoleAllowed})
	a.SetOptions(acOptions())

	typeWord(a, wordEN)
	if a.HandleKey(spaceKey()) {
		t.Fatal("the boundary separator was consumed — the decision never changes consumption at the boundary")
	}
	eventually(t, func() bool { return sink.requireCount() >= 1 },
		"the pipeline never armed after the boundary (no RequireSurroundingText)")
	opsAfterArm := len(sink.opLog())

	// The live first answer: the client's pre-boundary state — the
	// separator has not been applied client-side yet.
	preEnd := uint32(len([]rune(wordEN)))
	a.HandleSurroundingText(wordEN, preEnd, preEnd)

	if got := len(sink.opLog()); got != opsAfterArm {
		t.Fatalf("emitter ops after the pre-boundary answer = %v, want none new —"+
			" the early answer must not resolve the round", sink.opLog())
	}
	if got := strings.Count(buf.String(), `"reason":"verify-mismatch"`); got != 0 {
		t.Fatalf("verify-mismatch records after the pre-boundary answer = %d, want 0 —"+
			" the early answer is not a mismatch", got)
	}
	if got := a.AutoCorrectCounters(); got.Fired != 1 || got.Abstained != 0 {
		t.Fatalf("counters after the pre-boundary answer = %+v, want fired=1 abstained=0 — the round stays open", got)
	}

	// The post-boundary push (the client applied the separator) settles
	// the round through the normal pipeline.
	tokenEnd := uint32(len([]rune(wordEN)) + 1)
	a.HandleSurroundingText(wordEN+" ", tokenEnd, tokenEnd)

	deletes := sink.deleteCalls()
	if len(deletes) != 1 || deletes[0].offset != -7 || deletes[0].nchars != 7 {
		t.Fatalf("deletes = %+v, want exactly [{-7 7}] — the range is token+tail", deletes)
	}
	a.HandleSurroundingText("", 0, 0) // the deletion proof completes the ladder (plan 10-11)
	if commits := sink.commitTexts(); len(commits) != 1 || commits[0] != wordRU+" " {
		t.Fatalf("commits = %q, want [%q] — the conversion plus the tail", commits, wordRU+" ")
	}

	// The verify-after round settles exactly as in the fireAutocorrect
	// dance.
	eventually(t, func() bool { return sink.requireCount() >= 2 },
		"the verify-after round never armed (the correction never executed)")
	convEnd := uint32(len([]rune(wordRU)) + 1)
	a.HandleSurroundingText(wordRU+" ", convEnd, convEnd)
}

// TestActor_AutocorrectForeignVerifyAnswerMismatch pins the tolerance's
// guard: the early-answer wait covers ONLY the token-only pre-boundary
// state (the separator in flight) — a foreign push inside the verify
// budget still concludes verify-mismatch with zero emitter ops, the
// fail-closed direction untouched (D-53).
func TestActor_AutocorrectForeignVerifyAnswerMismatch(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.UseRole(&fakeRole{role: acRoleAllowed})
	a.SetOptions(acOptions())

	typeWord(a, wordEN)
	if a.HandleKey(spaceKey()) {
		t.Fatal("the boundary separator was consumed — the decision never changes consumption at the boundary")
	}
	eventually(t, func() bool { return sink.requireCount() >= 1 },
		"the pipeline never armed after the boundary (no RequireSurroundingText)")

	// A foreign answer: neither the token-only pre-state nor the settled
	// token+tail — the field moved on to something else entirely.
	a.HandleSurroundingText("xx", 2, 2)

	if got := strings.Count(buf.String(), `"reason":"verify-mismatch"`); got != 1 {
		t.Fatalf("verify-mismatch records after the foreign answer = %d, want exactly 1", got)
	}
	if got := len(sink.deleteCalls()) + len(sink.commitTexts()); got != 0 {
		t.Fatalf("correction ops after the foreign answer = %d, want 0 — the round concluded by refusal", got)
	}
}

// TestActor_ResetKeyBoundary pins the reset side of the boundary (CORR-09
// preserved): the decision arms BEFORE the hard reset reads the finished
// word, Enter transits exactly as before, and the buffer is genuinely
// empty after the reset — a Double finds no word and refuses once with
// the empty-buffer reason.
func TestActor_ResetKeyBoundary(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.UseRole(&fakeRole{role: acRoleAllowed}) // the armed decision fires through the pipeline
	a.SetOptions(acOptions())

	typeWord(a, wordEN)
	if a.HandleKey(engine.EngineEvent{Keyval: engine.KeyReturn}) {
		t.Fatal("Enter was consumed — the reset is engine state, never consumption")
	}
	eventually(t, func() bool { return sink.requireCount() >= 1 },
		"the boundary did not arm before the reset")

	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := strings.Count(buf.String(), `"reason":"empty-buffer"`); got != 1 {
		t.Errorf("empty-buffer refusals after the Double = %d, want exactly 1 — CORR-09 cleared the buffer", got)
	}
	if got := len(sink.deleteCalls()) + len(sink.commitTexts()); got != 0 {
		t.Errorf("correction ops after the Double = %d, want 0 — the reset emptied the buffer", got)
	}
}

// TestActor_BackspaceNoBoundary pins that a pop never arms the decision:
// the Backspace press (and the letters after it) produce no boundary, no
// arming and no counters.
func TestActor_BackspaceNoBoundary(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.SetOptions(acOptions())

	typeWord(a, wordEN)
	a.HandleKey(engine.EngineEvent{Keyval: engine.KeyBackSpace})
	a.HandleKey(engine.EngineEvent{Keyval: uint32('x')})

	if got := sink.requireCount(); got != 0 {
		t.Errorf("requires after the pop = %d, want 0 — a pop never arms the decision", got)
	}
	if got := a.AutoCorrectCounters(); got.Fired != 0 || got.Abstained != 0 {
		t.Errorf("counters after the pop = %+v, want zero — the pop is not a boundary", got)
	}
}

// TestActor_FocusOutNoBoundary pins the research-Q5 scope: FocusOut ends
// the phrase (CORR-09) but is NOT a word boundary — the word was not
// finished by a keystroke, so the autocorrect layer stays silent there.
func TestActor_FocusOutNoBoundary(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.UseRole(&fakeRole{role: acRoleAllowed})
	a.SetOptions(acOptions())

	typeWord(a, wordEN)
	a.HandleLifecycle(engine.LifecycleFocusOut)

	if got := sink.requireCount(); got != 0 {
		t.Errorf("requires after FocusOut = %d, want 0 — a focus loss never arms the decision", got)
	}
	if got := a.AutoCorrectCounters(); got.Fired != 0 || got.Abstained != 0 {
		t.Errorf("counters after FocusOut = %+v, want zero", got)
	}
}

// TestActor_ManualOverrideAfterAutocorrect pins criterion 4: the manual
// path rides above the autocorrect layer and never consults its detector —
// after the silent ghbdtn→привет correction a Double Shift converts привет
// back to ghbdtn (a dictionary-valid word the autocorrect layer itself
// would veto).
func TestActor_ManualOverrideAfterAutocorrect(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.SetOptions(acOptions())

	fireAutocorrect(t, a, sink, wordEN, wordRU)

	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText("", 0, 0) // the deletion proof completes the ladder (plan 10-11)

	commits := sink.commitTexts()
	if len(commits) != 2 || commits[1] != wordEN+" " {
		t.Fatalf("commits = %q, want [%q %q] — the manual round converted привет back", commits, wordRU+" ", wordEN+" ")
	}
}

// The role-contour corpus of plan 06-06 Task 2 (D-53): the live GetRole
// gate runs OFF the actor mutex under a hard deadline, the fail-closed
// silence counts every cell, and the fired decision re-enters THE one
// pipeline.

// The role enum of the corpus (the live-verified AT-SPI values —
// 06-RESEARCH Q4; internal/appid pins the same numbers).
const (
	acRoleAllowed           uint32 = 61 // text box
	acRoleForbiddenEntry    uint32 = 40 // password text
	acRoleForbiddenTerminal uint32 = 60 // terminal
)

// acReasonFired is the fired record's reason literal (the INFO log class
// of the fired decision); acReasonRoleForbidden is the shared matrix slug;
// acReasonAppBlocked is the 07-02 blocklist slug (pinned by the e2e oracle
// name ac_skip_app_blocked); acReasonNoCaps and acReasonTrigramUnsure are
// the security-cell slugs the inverted corpus reuses across tables.
const (
	acReasonFired         = "fired"
	acReasonRoleForbidden = "role-forbidden"
	acReasonAppBlocked    = "app-blocked"
	acReasonNoCaps        = "no-caps"
	acReasonTrigramUnsure = "trigram-unsure"
	// acReasonTerminalVeto is the plan 10-03 arming-veto slug (TERM-04,
	// D-55): the terminal class refuses at the ARMING point, ahead of the
	// detector. The literal is pinned here independently — the test greps
	// the wire form, not the identifier.
	acReasonTerminalVeto = "terminal-veto"
)

// acUnsureWord is the corpus's OOV token ("москва" typed in EN): a
// both-dictionary miss the detector answers unsure with.
const acUnsureWord = "vjcrdf"

// boundaryWord types one wrong-layout token and lands its separator — the
// raw boundary of the counter/warn corpora (no verify dance: the cells
// abstain, or the caller polls the specific counter).
func boundaryWord(t *testing.T, a *session.Actor, token string) {
	t.Helper()
	typeWord(a, token)
	if a.HandleKey(spaceKey()) {
		t.Fatalf("the boundary separator was consumed for %q", token)
	}
}

// awaitRoleStart waits until the double's live call begins, bounded by the
// poll budget — a RED-safe form of <-started (the contour's absence is a
// test failure, never a hang).
func awaitRoleStart(t *testing.T, role *fakeRole) {
	t.Helper()
	select {
	case <-role.started:
	case <-time.After(acPollBudget):
		t.Fatal("the role source was never consulted — the D-53 contour is absent")
	}
}

// fakeRole is the role-source test double (the fakeAppid form): a fixed
// answer or a fixed failure, with optional call gating — started closes on
// entry (proof the live call began), release blocks the return until
// closed (the deadline and mutex-freedom cells). The double honors the
// caller's ctx exactly like the godbus CallWithContext round trip.
type fakeRole struct {
	role    uint32
	err     error
	started chan struct{}
	release chan struct{}
	once    sync.Once

	mu    sync.Mutex
	calls int
}

// Role answers the double's programmed verdict.
func (f *fakeRole) Role(ctx context.Context) (uint32, error) {
	f.mu.Lock()
	f.calls++
	role, err := f.role, f.err
	started, release := f.started, f.release
	f.mu.Unlock()
	if started != nil {
		f.once.Do(func() { close(started) })
	}
	if release != nil {
		select {
		case <-release:
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}

	return role, err
}

// set re-programs the double between boundaries (the episode-swap corpus).
func (f *fakeRole) set(role uint32, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.role, f.err = role, err
}

// callCount snapshots the consultation counter — the anchor corpus proves
// the live source is never consulted when the identity already classified.
func (f *fakeRole) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.calls
}

// assertSilence checks the shared silence invariants of one matrix cell:
// nothing fired, no correction op touched the field, exactly one
// abstention with the cell's reason slug.
func assertSilence(t *testing.T, a *session.Actor, sink *fakeSink, reason string) {
	t.Helper()
	counters := a.AutoCorrectCounters()
	if counters.Fired != 0 {
		t.Errorf("fired = %d, want 0 — the cell must stay silent", counters.Fired)
	}
	if counters.Abstained != 1 {
		t.Errorf("abstained = %d, want 1", counters.Abstained)
	}
	if counters.Reasons[reason] != 1 {
		t.Errorf("reasons = %v, want [%q:1]", counters.Reasons, reason)
	}
	if got := len(sink.deleteCalls()) + len(sink.commitTexts()) + len(sink.forwardCalls()); got != 0 {
		t.Errorf("correction ops fired in a silence cell (%d) — the typed word must stay in the field", got)
	}
}

// TestAutoCorrect_FiresThroughPipeline pins the resolve path: with every
// gate green and the live role in the allowed set, the fired decision
// re-enters THE one correction pipeline (the level-1 ladder — no own
// delete/commit replay), counts Fired, and records the fired INFO class
// WITHOUT the word (D-20/D-21).
func TestAutoCorrect_FiresThroughPipeline(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.SetOptions(acOptions())

	fireAutocorrect(t, a, sink, wordEN, wordRU)

	deletes := sink.deleteCalls()
	if len(deletes) != 1 || deletes[0].offset != -7 || deletes[0].nchars != 7 {
		t.Fatalf("deletes = %+v, want exactly [{-7 7}] — the ladder level 1 ran", deletes)
	}
	if commits := sink.commitTexts(); len(commits) != 1 || commits[0] != wordRU+" " {
		t.Fatalf("commits = %q, want [%q]", commits, wordRU+" ")
	}
	if forwards := sink.forwardCalls(); len(forwards) != 0 {
		t.Errorf("forwards = %d, want 0 — the fired decision adds no key replay of its own", len(forwards))
	}
	if got := a.AutoCorrectCounters(); got.Fired != 1 || got.Abstained != 0 {
		t.Errorf("counters = %+v, want {Fired:1 Abstained:0}", got)
	}
	logged := buf.String()
	if !strings.Contains(logged, `"msg":"autocorrect"`) || !strings.Contains(logged, `"reason":"`+acReasonFired+`"`) {
		t.Errorf("the fired INFO record is missing; log:\n%s", logged)
	}
	for _, word := range []string{wordEN, wordRU} {
		if strings.Contains(logged, word) {
			t.Errorf("the log leaked %q (D-20/D-21); log:\n%s", word, logged)
		}
	}
}

// silenceMatrixCells is the D-53 conjunction matrix (plan 06-06, inverted
// by 07-04 per the locked 07-CONTEXT semantics): every refusal cell with
// its expected reason slug, plus the one FIRE cell — an UNKNOWN app
// identity is NOT a prohibition (the identity is the only conjunction
// segment whose unknown passes; role/caps/detector stay fail-closed). The
// role cells conclude on their own goroutine (async); the cheap-gate cells
// abstain synchronously.
func silenceMatrixCells() []struct {
	name    string
	token   string
	caps    uint32
	appid   func(a *session.Actor)
	role    *fakeRole
	options func() session.Options
	reason  string
	async   bool
	fired   bool
} {
	return []struct {
		name    string
		token   string
		caps    uint32
		appid   func(a *session.Actor)
		role    *fakeRole
		options func() session.Options
		reason  string
		async   bool
		fired   bool
	}{
		{
			name: "role 40 password text is forbidden", token: wordEN,
			role: &fakeRole{role: acRoleForbiddenEntry}, reason: acReasonRoleForbidden, async: true,
		},
		{
			// The D-55 migration (Phase 10): the terminal verdict moved from
			// the confirm's role switch to the ARMING point — role 60 now
			// abstains synchronously with terminal-veto, ahead of the
			// detector (the confirm switch stays as the second belt).
			name: "role 60 terminal vetoes at arming", token: wordEN,
			role: &fakeRole{role: acRoleForbiddenTerminal}, reason: acReasonTerminalVeto,
		},
		{
			name: "role error is unknown", token: wordEN,
			role: &fakeRole{err: errAppidBusDead}, reason: "role-unknown", async: true,
		},
		{
			name: "role deadline times out", token: wordEN,
			role: &fakeRole{role: acRoleAllowed, release: make(chan struct{})}, reason: "role-timeout", async: true,
		},
		{
			// The 07-04 inversion (locked 07-CONTEXT): an UNKNOWN identity
			// is NOT a prohibition — the arm gate passes the payload
			// WITHOUT the app field, the detector runs, and with a healthy
			// UNAMBIGUOUS role (entry 79) the word FIRES. The ambiguous 61
			// with an unknown identity is the CR-01 refusal (its own test).
			name: "unknown identity is not a prohibition: no source fires", token: wordEN,
			appid: func(a *session.Actor) {
				a.UseAppidStarter(func() (session.AppidSource, error) { return nil, errAppidNoBus })
			},
			role:  &fakeRole{role: acRoleAllowedEntry},
			fired: true,
		},
		{
			// The 07-02 blocklist polarity: a KNOWN identity matching a
			// pattern forbids the correction (app-blocked) — replacing the
			// white-list "unlisted app" cell.
			name: "blocklist match forbids a known app", token: wordEN,
			appid:   func(a *session.Actor) { a.UseAppid(fakeAppid{app: acListedApp}) },
			options: acBlockedOptions,
			reason:  acReasonAppBlocked,
		},
		{name: "no surrounding-text cap", token: wordEN, caps: 0, reason: acReasonNoCaps},
		{name: "detector unsure on a both-dictionary miss", token: acUnsureWord, reason: acReasonTrigramUnsure},
		{name: "short word abstains", token: "ok", reason: "abstain-short"},
	}
}

// TestAutoCorrect_SilenceMatrix pins the conjunction cell by cell: every
// SECURITY segment — a forbidden role (password 40, terminal 60), a role
// error, a role deadline, a missing capability, an unsure verdict, a short
// word — stays silent (zero correction ops, the word stays in the field)
// and counts exactly one abstention with its reason slug, while the
// identity segment's unknown FIRES for the unambiguous roles (the 07-04
// inversion: unknown is not a prohibition — the pinned direction; the
// ambiguous 61 + unknown cell refuses as role-ambiguous per the CR-01
// tightening, pinned separately).
func TestAutoCorrect_SilenceMatrix(t *testing.T) {
	for _, tc := range silenceMatrixCells() {
		t.Run(tc.name, func(t *testing.T) {
			caps := tc.caps
			if caps == 0 && tc.reason != acReasonNoCaps {
				caps = engine.CapSurroundingText
			}
			a, sink := wiredActorCaps(caps)
			if tc.appid != nil {
				tc.appid(a)
			} else {
				a.UseAppid(fakeAppid{app: acListedApp})
			}
			if tc.role != nil {
				a.UseRole(tc.role)
			}
			if tc.options != nil {
				a.SetOptions(tc.options())
			} else {
				a.SetOptions(acOptions())
			}

			boundaryWord(t, a, tc.token)
			if tc.fired {
				eventually(t, func() bool {
					c := a.AutoCorrectCounters()

					return c.Fired == 1 && c.Abstained == 0
				}, "the unknown identity never fired — unknown is not a prohibition (07-CONTEXT locked)")
				if got := a.AutoCorrectCounters(); got.Fired != 1 {
					t.Errorf("fired = %d, want exactly 1", got.Fired)
				}

				return
			}
			if tc.async {
				eventually(t, func() bool { return a.AutoCorrectCounters().Abstained >= 1 },
					"the confirm never concluded")
			}
			assertSilence(t, a, sink, tc.reason)
		})
	}
}

// TestAutoCorrect_UnknownIdentityFires pins the 07-04 arm inversion
// (locked 07-CONTEXT): with NO identity source at all and a healthy
// UNAMBIGUOUS text role (entry 79), the wrong-layout word IS corrected —
// the identity is simply omitted from the payload (no app), the detector
// runs, and the full pipeline fires. The old D-53-unknown silence is gone:
// an invisible app is not a forbidden app. The ambiguous role 61 cell is
// the CR-01 exception — TestAutoConfirm_RoleTextAmbiguity pins it.
func TestAutoCorrect_UnknownIdentityFires(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppidStarter(func() (session.AppidSource, error) { return nil, errAppidNoBus })
	a.SetOptions(acOptions())

	fireAutocorrectRole(t, a, sink, wordEN, wordRU, acRoleAllowedEntry)

	deletes := sink.deleteCalls()
	if len(deletes) != 1 || deletes[0].offset != -7 || deletes[0].nchars != 7 {
		t.Fatalf("deletes = %+v, want exactly [{-7 7}] — the correction ran without any identity", deletes)
	}
	if commits := sink.commitTexts(); len(commits) != 1 || commits[0] != wordRU+" " {
		t.Fatalf("commits = %q, want [%q] — the conversion plus the tail", commits, wordRU+" ")
	}
	if got := a.AutoCorrectCounters(); got.Fired != 1 || got.Abstained != 0 {
		t.Errorf("counters = %+v, want {Fired:1 Abstained:0} — the unknown identity never refused", got)
	}
}

// TestAutoCorrect_KnownNotBlockedFires pins the positive blocklist cell
// (07-04, the mirror of the match refusal): a KNOWN identity that matches
// NO pattern leaves every gate green — the word fires exactly as before
// the blocklist existed. Green-by-design continuity pin against the 07-02
// corpus (the fire paths never depended on the list).
func TestAutoCorrect_KnownNotBlockedFires(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.SetOptions(acOptions())

	fireAutocorrect(t, a, sink, wordEN, wordRU)

	if got := a.AutoCorrectCounters(); got.Fired != 1 || got.Abstained != 0 {
		t.Errorf("counters = %+v, want {Fired:1 Abstained:0} — a non-blocked app fires", got)
	}
	if got := len(sink.deleteCalls()); got != 1 {
		t.Errorf("deletes = %d, want 1 — the known-not-blocked path runs the ladder", got)
	}
}

// TestAutoCorrect_IdentityUnknownNotProhibition pins the DIRECTION of the
// 07-04 inversion: ONLY the app-identity segment softened. With the
// identity unknown, every other unknown of the conjunction keeps its
// fail-closed refusal — a role error is role-unknown, a role deadline is
// role-timeout, a missing surrounding-text capability is no-caps, an
// unsure detector verdict is trigram-unsure. The inversion must never
// creep into the safety segments (T-07-04-02).
func TestAutoCorrect_IdentityUnknownNotProhibition(t *testing.T) {
	cases := []struct {
		name   string
		caps   uint32
		role   *fakeRole
		token  string
		reason string
		async  bool
	}{
		{
			name: "role error stays role-unknown", token: wordEN,
			role: &fakeRole{err: errAppidBusDead}, reason: "role-unknown", async: true,
		},
		{
			name: "role deadline stays role-timeout", token: wordEN,
			role: &fakeRole{role: acRoleAllowed, release: make(chan struct{})}, reason: "role-timeout", async: true,
		},
		{
			name: "missing caps stays no-caps", token: wordEN,
			caps: 0, reason: acReasonNoCaps,
		},
		{
			name: "unsure detector stays trigram-unsure", token: acUnsureWord,
			reason: acReasonTrigramUnsure,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			caps := tc.caps
			if caps == 0 && tc.reason != acReasonNoCaps {
				caps = engine.CapSurroundingText
			}
			a, sink := wiredActorCaps(caps)
			a.UseAppidStarter(func() (session.AppidSource, error) { return nil, errAppidNoBus })
			if tc.role != nil {
				a.UseRole(tc.role)
			}
			a.SetOptions(acOptions())

			boundaryWord(t, a, tc.token)
			if tc.async {
				eventually(t, func() bool { return a.AutoCorrectCounters().Abstained >= 1 },
					"the confirm never concluded")
			}
			assertSilence(t, a, sink, tc.reason)
		})
	}
}

// TestAutoCorrect_WarnOncePerEpisode pins the warn discipline (the
// warnAppid form): a dead role source warns exactly ONCE per episode while
// every boundary still counts its reason; a healthy answer closes the
// episode, and the next failing episode warns once again.
func TestAutoCorrect_WarnOncePerEpisode(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	role := &fakeRole{err: errAppidBusDead}
	a.UseRole(role)
	a.SetOptions(acOptions())

	boundaryWord(t, a, wordEN)
	eventually(t, func() bool { return a.AutoCorrectCounters().Abstained == 1 },
		"the first failing boundary never counted")
	boundaryWord(t, a, "cnhjrf")
	eventually(t, func() bool { return a.AutoCorrectCounters().Abstained == 2 },
		"the second failing boundary never counted")
	if got := strings.Count(buf.String(), `"msg":"autocorrect role unavailable"`); got != 1 {
		t.Fatalf("role WARNs in the first episode = %d, want exactly 1; log:\n%s", got, buf.String())
	}

	role.set(acRoleAllowed, nil) // a healthy answer closes the episode
	boundaryWord(t, a, "cegth")
	eventually(t, func() bool { return a.AutoCorrectCounters().Fired == 1 },
		"the healthy boundary never fired")

	role.set(0, errAppidBusDead) // a fresh failing episode
	boundaryWord(t, a, "vtyz")
	eventually(t, func() bool { return a.AutoCorrectCounters().Abstained == 3 },
		"the third failing boundary never counted")
	if got := strings.Count(buf.String(), `"msg":"autocorrect role unavailable"`); got != 2 {
		t.Errorf("role WARNs after the fresh episode = %d, want exactly 2; log:\n%s", got, buf.String())
	}
	if got := a.AutoCorrectCounters().Reasons["role-unknown"]; got != 3 {
		t.Errorf("role-unknown count = %d, want 3 — every boundary counted", got)
	}
}

// acRoleWarnRearmQuiet mirrors the production re-arm window of the sticky
// role WARN (plan 10-07): errors resuming after this silence sound the WARN
// AGAIN — a lasting outage must not dissolve into one morning line (the
// 11:00:16 episode stayed silent the whole day). The test advances past the
// window instead of sleeping it.
const acRoleWarnRearmQuiet = 31 * time.Minute

// newFakeClock returns the role-episode clock's controllable double: Now
// answers the advanced instant, advance moves it without sleeping.
func newFakeClock() *fakeClock {
	return &fakeClock{t: time.Now()}
}

type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

// Now answers the clock's current instant (the seam's signature).
func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.t
}

// advance moves the clock forward without sleeping.
func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

// TestAutoCorrect_RoleErrorsCounted pins the counter half of the plan-10-07
// role diagnostics (G-10-1 missing #3): EVERY role-request error of the
// episode grows the status-visible counter even while the WARN keeps its
// stickiness (one line per episode — the journal never spams; the counter
// never hides). A failing boundary counts twice: the boundary's own
// classify query and the confirm's live verdict.
func TestAutoCorrect_RoleErrorsCounted(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	role := &fakeRole{err: errAppidBusDead}
	a.UseRole(role)
	a.SetOptions(acOptions())

	boundaryWord(t, a, wordEN)
	eventually(t, func() bool { return a.StatusSnapshot().AutoCorrectRoleErrors == 2 },
		"the first failing boundary never counted its role errors")
	boundaryWord(t, a, "cnhjrf")
	eventually(t, func() bool { return a.StatusSnapshot().AutoCorrectRoleErrors == 4 },
		"the second failing boundary never counted its role errors")
	if got := strings.Count(buf.String(), `"msg":"autocorrect role unavailable"`); got != 1 {
		t.Errorf("role WARNs inside one episode = %d, want exactly 1 — the stickiness is kept; log:\n%s",
			got, buf.String())
	}
}

// TestAutoCorrect_RoleWarnRearmsAfterQuietWindow pins the re-arm half of
// the plan-10-07 role diagnostics: errors continuing after the quiet window
// sound the WARN AGAIN — a lasting outage is loud every window, never one
// absorbed morning line (11:00:16). Inside the window the stickiness holds.
func TestAutoCorrect_RoleWarnRearmsAfterQuietWindow(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	clock := newFakeClock()
	a.UseRoleClock(clock.Now)
	a.UseAppid(fakeAppid{app: acListedApp})
	role := &fakeRole{err: errAppidBusDead}
	a.UseRole(role)
	a.SetOptions(acOptions())

	boundaryWord(t, a, wordEN)
	eventually(t, func() bool { return a.StatusSnapshot().AutoCorrectRoleErrors == 2 },
		"the first failing boundary never counted")
	if got := strings.Count(buf.String(), `"msg":"autocorrect role unavailable"`); got != 1 {
		t.Fatalf("role WARNs at the first error = %d, want exactly 1; log:\n%s", got, buf.String())
	}

	clock.advance(acRoleWarnRearmQuiet) // the quiet interval passes — the outage is still on
	boundaryWord(t, a, "cnhjrf")
	eventually(t, func() bool { return a.StatusSnapshot().AutoCorrectRoleErrors == 4 },
		"the post-window boundary never counted")
	if got := strings.Count(buf.String(), `"msg":"autocorrect role unavailable"`); got != 2 {
		t.Errorf("role WARNs after the quiet window = %d, want 2 — "+
			"a lasting outage must sound again; log:\n%s", got, buf.String())
	}
}

// TestAutoCorrect_RoleHealthyClosesEpisodeAndZeroes pins the episode close
// (the :2741 semantics extended by plan 10-07): a healthy role answer
// closes the episode AND zeroes the consecutive-error counter — a recovered
// source reads as zero errors, and the next failing boundary opens a fresh
// episode (WARN again immediately, counter from zero).
func TestAutoCorrect_RoleHealthyClosesEpisodeAndZeroes(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	role := &fakeRole{err: errAppidBusDead}
	a.UseRole(role)
	a.SetOptions(acOptions())

	boundaryWord(t, a, wordEN)
	eventually(t, func() bool { return a.StatusSnapshot().AutoCorrectRoleErrors == 2 },
		"the failing boundary never counted")
	if got := strings.Count(buf.String(), `"msg":"autocorrect role unavailable"`); got != 1 {
		t.Fatalf("role WARNs at the first error = %d, want exactly 1; log:\n%s", got, buf.String())
	}

	role.set(acRoleAllowed, nil) // a healthy answer closes the episode
	boundaryWord(t, a, "cegth")
	eventually(t, func() bool { return a.AutoCorrectCounters().Fired == 1 },
		"the healthy boundary never fired")
	if got := a.StatusSnapshot().AutoCorrectRoleErrors; got != 0 {
		t.Errorf("role-error counter after the healthy answer = %d, want 0 — "+
			"the healthy answer zeroes the episode counter", got)
	}

	role.set(0, errAppidBusDead) // a fresh failing episode
	boundaryWord(t, a, "vtyz")
	eventually(t, func() bool { return a.StatusSnapshot().AutoCorrectRoleErrors == 2 },
		"the fresh episode never counted")
	if got := strings.Count(buf.String(), `"msg":"autocorrect role unavailable"`); got != 2 {
		t.Errorf("role WARNs after the fresh episode = %d, want exactly 2; log:\n%s", got, buf.String())
	}
}

// TestAutoCorrect_RoleDeadlineBounded pins the off-mutex deadline
// discipline (T-06-06-04): while the live role call blocks, the keystroke
// path stays fully live (HandleKey completes — the mutex never leaks into
// the call), and the confirm concludes on its own deadline with the
// role-timeout abstention — no pipeline launched.
func TestAutoCorrect_RoleDeadlineBounded(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	role := &fakeRole{role: acRoleAllowed, started: make(chan struct{}), release: make(chan struct{})}
	a.UseRole(role)
	a.SetOptions(acOptions())

	boundaryWord(t, a, wordEN)
	awaitRoleStart(t, role)

	keyDone := make(chan bool, 1)
	go func() { keyDone <- a.HandleKey(engine.EngineEvent{Keyval: uint32('z')}) }()
	select {
	case <-keyDone:
	case <-time.After(acPollBudget):
		t.Fatal("HandleKey blocked while the role call was in flight — the mutex leaked into the live call")
	}

	eventually(t, func() bool { return a.AutoCorrectCounters().Abstained == 1 },
		"the confirm never concluded after the role deadline")
	counters := a.AutoCorrectCounters()
	if counters.Reasons["role-timeout"] != 1 {
		t.Errorf("reasons = %v, want [role-timeout:1]", counters.Reasons)
	}
	if got := sink.requireCount(); got != 0 {
		t.Errorf("requires = %d, want 0 — a timed-out decision never launches the pipeline", got)
	}
}

// TestAutoCorrect_ReEntryRechecks pins the stale-payload guard
// (T-06-06-06): the config flips OFF between the armed boundary and the
// confirm's verdict — the re-check under the mutex silences the confirm
// (no correction, one recheck-disabled abstention).
func TestAutoCorrect_ReEntryRechecks(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	role := &fakeRole{role: acRoleAllowed, started: make(chan struct{}), release: make(chan struct{})}
	a.UseRole(role)
	a.SetOptions(acOptions())

	boundaryWord(t, a, wordEN)
	awaitRoleStart(t, role)

	a.SetOptions(session.Options{}) // the hot-reload shift lands between arm and verdict
	close(role.release)

	eventually(t, func() bool { return a.AutoCorrectCounters().Abstained == 1 },
		"the confirm never concluded")
	counters := a.AutoCorrectCounters()
	if counters.Fired != 0 || counters.Reasons["recheck-disabled"] != 1 {
		t.Errorf("counters = %+v, want {Fired:0, recheck-disabled:1} — the stale payload never corrects", counters)
	}
	if got := len(sink.deleteCalls()) + len(sink.commitTexts()); got != 0 {
		t.Errorf("correction ops = %d, want 0", got)
	}
}

// The review-fix corpus of 06-REVIEW (CR-01/CR-02) with the 07-04
// confirm-gate rework on top: the async confirm must never execute against
// state that moved on inside the role-RTT window (CR-01/CR-02), and the
// confirm's identity check is BLOCKLIST-ONLY on the CURRENT identity
// (owner variant 1 — the arming/confirm equality of WR-01 is deliberately
// REMOVED: a focus move inside the window is caught by the buffer's own
// payload-stale geometry, FocusOut hard-resets the buffer). Every test
// drives the interleaving deterministically — arm the payload, block the
// live role call on the double's release channel, mutate the interleaved
// state, then let the confirm conclude — no sleeps: the gating is the
// double's own channels (the awaitRoleStart precedent).

// TestAutoConfirm_RevalidatesArmedPayload pins the CR-01 fix: a
// token-capable keystroke inside the role-RTT window moves the buffer off
// the armed payload while the cached surrounding push still ends with the
// armed token+tail (the keystroke's own client push is in flight) — the
// cached-push fast path would fire the delete+commit against the stale
// geometry and corrupt the field and the mirror. The confirm must
// re-validate the armed payload against the live buffer under the mutex and
// refuse fail-closed: silence + the payload-stale counter, the word never
// logged (D-20/D-21).
func TestAutoConfirm_RevalidatesArmedPayload(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	role := &fakeRole{role: acRoleAllowed, started: make(chan struct{}), release: make(chan struct{})}
	a.UseRole(role)
	a.SetOptions(acOptions())

	boundaryWord(t, a, wordEN)
	awaitRoleStart(t, role)

	// The client's post-separator push lands while the role round trip is
	// in flight: the cache ends with the armed token+tail — the exact
	// freshness the cached fast path trusts.
	tokenEnd := uint32(len([]rune(wordEN)) + 1)
	a.HandleSurroundingText(wordEN+" ", tokenEnd, tokenEnd)

	// The interleaved keystroke: the daemon has processed it, the client's
	// own push for it is still in flight — the buffer has moved on, the
	// cache has not.
	a.HandleKey(engine.EngineEvent{Keyval: uint32('x')})
	close(role.release)

	eventually(t, func() bool {
		c := a.AutoCorrectCounters()

		return c.Abstained >= 1 || c.Fired >= 1
	}, "the confirm never concluded")
	if c := a.AutoCorrectCounters(); c.Fired != 0 {
		t.Fatalf("fired = %d — the moved-on buffer must never fire the armed payload (CR-01)", c.Fired)
	}
	assertSilence(t, a, sink, "payload-stale")
	logged := buf.String()
	for _, word := range []string{wordEN, wordRU} {
		if strings.Contains(logged, word) {
			t.Errorf("the log leaked %q (D-20/D-21); log:\n%s", word, logged)
		}
	}
}

// TestAutoConfirm_ResetBoundaryAlwaysVerifies pins the CR-02 fix: a
// reset-boundary payload (Enter/Tab/Escape) must NEVER take the cached-push
// fast path — the arming keystroke is itself an in-flight field-mutating or
// focus-moving key, so the cached push is the PRE-reset state by
// construction and a delete+commit on it can land in a moved-on or entirely
// different input context. The payload always runs the pre-correction
// verify round, whose answer names the field as the reset key left it: a
// stale answer refuses fail-closed (verify-mismatch), never a
// wrong-context write (ADR-004).
func TestAutoConfirm_ResetBoundaryAlwaysVerifies(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	role := &fakeRole{role: acRoleAllowed, started: make(chan struct{}), release: make(chan struct{})}
	a.UseRole(role)
	a.SetOptions(acOptions())

	typeWord(a, wordEN)
	// The client's last spontaneous push before the reset key: the cache
	// ends with the armed word — the pre-reset state the cached fast path
	// must not trust for a boundary-armed payload.
	end := uint32(len([]rune(wordEN)))
	a.HandleSurroundingText(wordEN, end, end)

	if a.HandleKey(engine.EngineEvent{Keyval: engine.KeyEscape}) {
		t.Fatal("the reset key was consumed — the reset is engine state, never consumption")
	}
	awaitRoleStart(t, role)
	close(role.release)

	// The conclusion must be the verify round's Require — never the cached
	// delete+commit.
	eventually(t, func() bool {
		return sink.requireCount() >= 1 || len(sink.deleteCalls())+len(sink.commitTexts()) > 0
	}, "the reset-boundary payload never concluded")
	if got := len(sink.deleteCalls()) + len(sink.commitTexts()); got != 0 {
		t.Fatalf("correction ops without a verify round = %d — "+
			"the cached push must never shortcut a reset boundary (CR-02)", got)
	}

	// The verify answer names the field AFTER the reset key's own effect —
	// the single-line entry consumed/submitted: the armed word is gone.
	// The round refuses fail-closed with zero field edits.
	a.HandleSurroundingText("", 0, 0)

	if got := strings.Count(buf.String(), `"reason":"verify-mismatch"`); got != 1 {
		t.Errorf("verify-mismatch records after the post-reset answer = %d, want exactly 1; log:\n%s",
			got, buf.String())
	}
	if got := len(sink.deleteCalls()) + len(sink.commitTexts()); got != 0 {
		t.Errorf("correction ops after the post-reset answer = %d, want 0 — the stale word stays untouched", got)
	}
}

// TestAutoCorrect_ConfirmBlocklistRefuses pins the 07-04 confirm form
// (owner variant 1, locked 07-CONTEXT): the payload was armed while the
// identity was UNKNOWN, and by confirm time the identity has appeared AND
// matches the blocklist. The confirm evaluates the CURRENT identity — the
// refusal is app-blocked, the field is untouched. The removed arming/
// confirm equality would have produced a phantom mismatch here instead
// (the armed payload carries no app): the locked scenario "unknown at
// arm, learned by confirm" requires its absence.
func TestAutoCorrect_ConfirmBlocklistRefuses(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppidStarter(func() (session.AppidSource, error) { return nil, errAppidNoBus })
	role := &fakeRole{role: acRoleAllowed, started: make(chan struct{}), release: make(chan struct{})}
	a.UseRole(role)
	a.SetOptions(acBlockedOptions())

	boundaryWord(t, a, wordEN) // armed at UNKNOWN identity — the blocklist sees nothing yet
	awaitRoleStart(t, role)

	// The identity is learned inside the role-RTT window and MATCHES the
	// list: the current identity is exactly what the confirm must consult.
	a.UseAppid(fakeAppid{app: acListedApp})
	close(role.release)

	eventually(t, func() bool {
		c := a.AutoCorrectCounters()

		return c.Abstained >= 1 || c.Fired >= 1
	}, "the confirm never concluded")
	if c := a.AutoCorrectCounters(); c.Fired != 0 {
		t.Fatalf("fired = %d — a blocklisted app must never fire at confirm", c.Fired)
	}
	assertSilence(t, a, sink, acReasonAppBlocked)
}

// TestAutoCorrect_ConfirmUnknownPasses pins the confirm-side mirror of the
// arm inversion: an identity that stays UNKNOWN through the whole episode
// never refuses the confirm on the identity alone — unknown is not a
// prohibition on either gate (locked 07-CONTEXT). The role here is the
// UNAMBIGUOUS entry (79): the ambiguous 61 with an unknown identity is the
// CR-01 refusal (TestAutoConfirm_RoleTextAmbiguity). With a healthy role
// the word fires.
func TestAutoCorrect_ConfirmUnknownPasses(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppidStarter(func() (session.AppidSource, error) { return nil, errAppidNoBus })
	role := &fakeRole{role: acRoleAllowedEntry, started: make(chan struct{}), release: make(chan struct{})}
	a.UseRole(role)
	a.SetOptions(acOptions())

	boundaryWord(t, a, wordEN) // unknown at arm — the payload carries no app
	awaitRoleStart(t, role)
	close(role.release) // still unknown at confirm — the pass-through (role 79, unambiguous)

	eventually(t, func() bool { return a.AutoCorrectCounters().Fired == 1 },
		"the never-known identity refused the confirm — unknown must pass")
	if c := a.AutoCorrectCounters(); c.Abstained != 0 {
		t.Errorf("abstained = %d, want 0 — the unknown identity never refuses", c.Abstained)
	}

	// The fired confirm launches THE pipeline: settle its pre-correction
	// verify with the field's post-boundary state — the confirmed word
	// runs the ladder exactly like any other fired round.
	eventually(t, func() bool { return sink.requireCount() >= 1 },
		"the pipeline never armed after the confirm")
	tokenEnd := uint32(len([]rune(wordEN)) + 1)
	a.HandleSurroundingText(wordEN+" ", tokenEnd, tokenEnd)

	if got := len(sink.deleteCalls()); got != 1 {
		t.Errorf("deletes = %d, want 1 — the confirmed word runs the ladder", got)
	}
}

// acRoleAllowedEntry is the ENTRY role (79) of the ambiguity pair corpus:
// an allowed role that is NOT the ambiguous one — the "unknown passes"
// direction keeps firing on it under the CR-01 tightening.
const acRoleAllowedEntry uint32 = 79

// TestAutoConfirm_RoleTextAmbiguity pins the CR-01 tightening (07-REVIEW,
// owner-sanctioned narrow exception to the "unknown is not a prohibition"
// rule): role 61 is the AMBIGUOUS role — GTK3 password fields report it
// exactly like text boxes (ADR-007's live fact) — so a confirm resolving
// role 61 while the identity is UNKNOWN refuses fail-closed with the
// role-ambiguous slug. A KNOWN identity keeps the locked semantics: role 61
// fires and the blocklist decides.
func TestAutoConfirm_RoleTextAmbiguity(t *testing.T) {
	t.Run("unknown identity with role 61 refuses", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()
		a.UseAppidStarter(func() (session.AppidSource, error) { return nil, errAppidNoBus })
		role := &fakeRole{role: acRoleAllowed, started: make(chan struct{}), release: make(chan struct{})}
		a.UseRole(role)
		a.SetOptions(acOptions())

		boundaryWord(t, a, wordEN) // armed at the UNKNOWN identity
		awaitRoleStart(t, role)
		close(role.release) // the live role answers 61 — the identity never became known

		eventually(t, func() bool { return a.AutoCorrectCounters().Abstained >= 1 },
			"the confirm never concluded")
		if c := a.AutoCorrectCounters(); c.Fired != 0 {
			t.Fatalf("fired = %d — the ambiguous role 61 with an UNKNOWN identity must refuse (CR-01)", c.Fired)
		}
		assertSilence(t, a, sink, "role-ambiguous")
		logged := buf.String()
		for _, word := range []string{wordEN, wordRU} {
			if strings.Contains(logged, word) {
				t.Errorf("the log leaked %q (D-20/D-21); log:\n%s", word, logged)
			}
		}
	})

	t.Run("known identity with role 61 fires", func(t *testing.T) {
		a, sink := wiredActor()
		a.UseAppid(fakeAppid{app: acListedApp})
		role := &fakeRole{role: acRoleAllowed, started: make(chan struct{}), release: make(chan struct{})}
		a.UseRole(role)
		a.SetOptions(acOptions())

		boundaryWord(t, a, wordEN)
		awaitRoleStart(t, role)
		close(role.release)

		eventually(t, func() bool { return a.AutoCorrectCounters().Fired == 1 },
			"role 61 with a KNOWN identity refused — the blocklist decides, the ambiguity must not")
		eventually(t, func() bool { return sink.requireCount() >= 1 },
			"the pipeline never armed after the fired confirm")
		tokenEnd := uint32(len([]rune(wordEN)) + 1)
		a.HandleSurroundingText(wordEN+" ", tokenEnd, tokenEnd)
	})
}

// TestAutoCorrect_CountersAndReasons pins the counter surface over a mixed
// scenario: fired and every abstention class accumulate in
// AutoCorrectCounters — counts and closed slugs only, no word in the log.
func TestAutoCorrect_CountersAndReasons(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	role := &fakeRole{role: acRoleAllowed}
	a.UseRole(role)
	a.SetOptions(acOptions())

	boundaryWord(t, a, "ok")     // short → abstain-short (synchronous)
	boundaryWord(t, a, "vjcrdf") // unsure → trigram-unsure (synchronous)
	if got := a.AutoCorrectCounters(); got.Abstained != 2 {
		t.Fatalf("abstained after the sync cells = %d, want 2", got.Abstained)
	}

	fireAutocorrect(t, a, sink, wordEN, wordRU)      // fired (role 61)
	a.UseRole(&fakeRole{role: acRoleForbiddenEntry}) // fireAutocorrect installed its own double
	boundaryWord(t, a, "cnhjrf")                     // role 40 → role-forbidden
	eventually(t, func() bool { return a.AutoCorrectCounters().Abstained == 3 },
		"the forbidden-role boundary never counted")

	counters := a.AutoCorrectCounters()
	want := map[string]int{"abstain-short": 1, acReasonTrigramUnsure: 1, acReasonRoleForbidden: 1}
	if counters.Fired != 1 || counters.Abstained != 3 || !maps.Equal(counters.Reasons, want) {
		t.Errorf("counters = {Fired:%d Abstained:%d Reasons:%v}, want {Fired:1 Abstained:3 Reasons:%v}",
			counters.Fired, counters.Abstained, counters.Reasons, want)
	}
	logged := buf.String()
	for _, word := range []string{wordEN, wordRU, "cnhjrf", "vjcrdf"} {
		if strings.Contains(logged, word) {
			t.Errorf("the log leaked %q (D-20/D-21)", word)
		}
	}
}

// The status/fold corpus of plan 06-06 Task 3 (D-54): the autocorrect
// state and counters ride the status snapshot, the config section folds
// live through applySnapshot, and the ctl line carries the new tokens.

// acEnabledCfg is a defaults-based document with the autocorrect section
// ACTIVE on a blocklist pattern that never matches the fixture app (the
// fire paths). The post-correction flip is pinned OFF here: the fold
// corpus asserts counters, not the owner's mode rule, and a mid-test
// script flip would retarget the detector's mode side.
func acEnabledCfg() config.Config {
	cfg := config.Defaults()
	cfg.Correction.FlipAfterCorrection = false
	cfg.Autocorrect = config.Autocorrect{
		Enabled:       true,
		AppsBlocklist: []string{acBlockMiss},
		MinWordLen:    acMinWordLen,
		TrigramMargin: acTrigramMargin,
		TrigramFloor:  acTrigramFloor,
	}

	return cfg
}

// TestStatus_AutocorrectFields pins the snapshot surface (D-54): the
// in-force switch, the fired/abstained counts and the CLONED reason map —
// zeros with the layer off, the real counts after a fired round, and a
// snapshot mutation never reaches the actor's own counters.
func TestStatus_AutocorrectFields(t *testing.T) {
	a, sink := wiredActor()
	st := a.StatusSnapshot()
	off := st.AutoCorrectEnabled || st.AutoCorrectFired != 0 ||
		st.AutoCorrectAbstained != 0 || len(st.AutoCorrectSkipReasons) != 0
	if off {
		t.Fatalf("off-state snapshot = %+v, want the zero autocorrect state", st)
	}

	a.UseAppid(fakeAppid{app: acListedApp})
	a.SetOptions(acOptions())
	fireAutocorrect(t, a, sink, wordEN, wordRU)

	st = a.StatusSnapshot()
	if !st.AutoCorrectEnabled || st.AutoCorrectFired != 1 || st.AutoCorrectAbstained != 0 {
		t.Fatalf("snapshot after the fired round = %+v, want {enabled true, fired 1, abstained 0}", st)
	}

	// One abstention gives the reason map a carried entry; the clone probe
	// mutates the snapshot's map and demands the actor's own counters
	// untouched (maps.Clone semantics — a nil source stays nil).
	boundaryWord(t, a, "ok")
	st = a.StatusSnapshot()
	if st.AutoCorrectAbstained != 1 || st.AutoCorrectSkipReasons["abstain-short"] != 1 {
		t.Fatalf("snapshot after the abstention = %+v, want [abstain-short:1]", st)
	}
	st.AutoCorrectSkipReasons["mutated"] = 99
	again := a.StatusSnapshot()
	if again.AutoCorrectSkipReasons["mutated"] != 0 || again.AutoCorrectSkipReasons["abstain-short"] != 1 {
		t.Errorf("the snapshot map is not a clone: %v", again.AutoCorrectSkipReasons)
	}
}

// TestActor_AppidStartsForAutocorrect pins the 07-04 observer start
// condition: autocorrect ENABLED starts the a11y observer even with an
// EMPTY blocklist and no MACR list — the blocklist negatives need the
// identity to refuse on, and the confirm's live role query needs the
// observer as its RoleSource. With the layer off and no lists the start
// count stays zero (the off state is byte-as-today, D-54).
func TestActor_AppidStartsForAutocorrect(t *testing.T) {
	t.Run("enabled autocorrect starts the observer with no lists", func(t *testing.T) {
		a, _ := wiredActor()
		starts := 0
		a.UseAppidStarter(func() (session.AppidSource, error) {
			starts++

			return fakeAppid{app: acListedApp}, nil
		})

		a.SetOptions(session.Options{AutoCorrectEnabled: true})

		if starts != 1 {
			t.Fatalf("observer starts = %d, want exactly 1 — enabled autocorrect needs the identity source", starts)
		}
	})
	t.Run("off with no lists never starts", func(t *testing.T) {
		a, _ := wiredActor()
		starts := 0
		a.UseAppidStarter(func() (session.AppidSource, error) {
			starts++

			return fakeAppid{app: acListedApp}, nil
		})

		a.SetOptions(session.Options{})

		if starts != 0 {
			t.Fatalf("observer starts = %d, want 0 — the off state is byte-as-today (D-54)", starts)
		}
	})
}

// TestApplySnapshot_AutocorrectFold pins the live fold (the MACR-fold
// precedent): the attached document's autocorrect section governs the
// boundary gates — the feature appears and disappears LIVE on reload, the
// folded MinWordLen gates the detector, and the blocklist starts the
// identity observer (and recompiles its pattern cache per edit).
func TestApplySnapshot_AutocorrectFold(t *testing.T) {
	a, sink := wiredActor()
	starts := 0
	a.UseAppidStarter(func() (session.AppidSource, error) {
		starts++

		return fakeAppid{app: acListedApp}, nil
	})
	src := &reloadSource{cfg: acEnabledCfg()}
	a.AttachConfig(src)
	a.UseRole(&fakeRole{role: acRoleAllowed})

	fireAutocorrect(t, a, sink, wordEN, wordRU)
	if starts != 1 {
		t.Errorf("the folded blocklist started the observer %d times, want exactly 1", starts)
	}
	if got := a.AutoCorrectCounters(); got.Fired != 1 {
		t.Fatalf("counters after the enabled fold = %+v, want Fired 1", got)
	}

	// The section disappears live: the same boundary is now the off
	// no-op (zero behavior, zero counters).
	cfg := config.Defaults()
	src.set(cfg)
	a.UseRole(&fakeRole{role: acRoleAllowed})
	typeWord(a, "cegth")
	a.HandleKey(spaceKey())
	if got := a.AutoCorrectCounters(); got.Fired != 1 || got.Abstained != 0 {
		t.Errorf("counters after the disabled fold = %+v, want unchanged — the off fold is byte-as-today", got)
	}

	// The thresholds fold too: min_word_len 16 refuses the six-rune word
	// with the detector's own slug.
	cfg = acEnabledCfg()
	cfg.Autocorrect.MinWordLen = 16
	src.set(cfg)
	a.UseRole(&fakeRole{role: acRoleAllowed})
	typeWord(a, "vtyz")
	a.HandleKey(spaceKey())
	if got := a.AutoCorrectCounters().Reasons["abstain-short"]; got != 1 {
		t.Errorf("abstain-short after the threshold fold = %d, want 1", got)
	}
}

// TestCtlStatus_EndToEnd drives the real actor's snapshot through the
// control line: a fired autocorrect shows as autocorrect_enabled=true with
// the fired count — the goswitchctl grammar needs no change (the generic
// key=value parser).
func TestCtlStatus_EndToEnd(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.SetOptions(acOptions())
	fireAutocorrect(t, a, sink, wordEN, wordRU)

	svc := ctlsvc.NewSvc(ctlsvc.Deps{Status: a})
	reply, err := svc.Status()
	if err != nil {
		t.Fatalf("Status error = %v, want nil", err)
	}
	for _, want := range []string{"autocorrect_enabled=true", "autocorrect_fired=1"} {
		if !strings.Contains(reply, want) {
			t.Errorf("Status reply %q missing %q", reply, want)
		}
	}
}

// The sound-sink corpus of plan 07-08: the acoustic-feedback seam — ONE
// Flip per flipTo execution path (the single flip point: the Single
// decision, SwitchMode, settleCombo, settleCorrectionFlip all land here),
// one AutoCorrect per fired correction, and the D-36 order pin (the tone
// fires after the mode record's observers).

// fakeSoundSink is the actor's SoundSink double (plan 07-08): both tones
// counted under a mutex, the re-pinned autocorrect event recorded (WR-02),
// with an optional flip hook the order pins interleave against the
// display/menu hooks. It also carries the SoundStopper capability (the
// 261006-squ mute lifecycle) — the stop counter is the fold pins' oracle.
type fakeSoundSink struct {
	mu     sync.Mutex
	flips  int
	fires  int
	event  string
	onFlip func()
	stops  int
}

// Flip records one flip tone and plays the hook.
func (f *fakeSoundSink) Flip() {
	f.mu.Lock()
	f.flips++
	hook := f.onFlip
	f.mu.Unlock()
	if hook != nil {
		hook()
	}
}

// AutoCorrect records one autocorrect tone.
func (f *fakeSoundSink) AutoCorrect() {
	f.mu.Lock()
	f.fires++
	f.mu.Unlock()
}

// SetAutocorrectEvent records the folded event push (WR-02). The session
// corpus pins the propagation from the daemon side (main_test's wiring
// corpus); here the double only satisfies the interface.
func (f *fakeSoundSink) SetAutocorrectEvent(event string) {
	f.mu.Lock()
	f.event = event
	f.mu.Unlock()
}

// Stop records the mute lifecycle's close — the fold's ON→OFF transition
// is the close trigger (requirement 3: muted means NO sound-server
// connection).
func (f *fakeSoundSink) Stop() {
	f.mu.Lock()
	f.stops++
	f.mu.Unlock()
}

func (f *fakeSoundSink) flipCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.flips
}

func (f *fakeSoundSink) fireCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.fires
}

func (f *fakeSoundSink) stopCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.stops
}

// flipSoundOn returns the bare sound-on Options — the flip-tone corpus
// configuration (autocorrect stays off; flips are what these cells drive).
func flipSoundOn() session.Options {
	return session.Options{SoundEnabled: true}
}

// TestActor_FlipSoundsSink pins the single flip point (plan 07-08): a
// gesture flip (the Single decision at expiry) and a menu flip
// (SwitchMode — the SAME flipTo path) each sound exactly one Flip; the
// same-target click lands on flipTo's no-op guard and stays silent.
func TestActor_FlipSoundsSink(t *testing.T) {
	a, _ := wiredActor()
	sink := &fakeSoundSink{}
	a.SetOptions(flipSoundOn())
	a.SetSoundSink(sink)

	flipMode(a)        // the Single decision at expiry — EN → RU
	a.SwitchMode("en") // the menu's radio pair — the same flipTo path
	a.SwitchMode("en") // the same-target no-op — silence

	if got := sink.flipCount(); got != 2 {
		t.Errorf("flip tones after a gesture flip + a menu flip + a same-target no-op = %d, want exactly 2", got)
	}
}

// TestActor_SameTargetFlipSilent pins the guard's silence: a flipTo to the
// CURRENT mode is a no-op — no record, no tone (the owner's rule: a no-op
// flip does not sound).
func TestActor_SameTargetFlipSilent(t *testing.T) {
	a, _ := wiredActor()
	sink := &fakeSoundSink{}
	a.SetOptions(flipSoundOn())
	a.SetSoundSink(sink)

	a.SwitchMode("en") // already EN — the same-target guard

	if got := sink.flipCount(); got != 0 {
		t.Errorf("flip tones after a same-target flip = %d, want exactly 0", got)
	}
}

// TestActor_SoundSinkAfterModeRecord extends the D-36 order pin (plan
// 07-08): the tone fires LAST — after the mode record's panel symbol, the
// display and the menu observers — on the same synchronous push chain.
func TestActor_SoundSinkAfterModeRecord(t *testing.T) {
	a, sink := wiredActor()
	var opMu sync.Mutex
	var ops []string
	record := func(op string) {
		opMu.Lock()
		ops = append(ops, op)
		opMu.Unlock()
	}
	sink.modeHook = func(symbol string) { record("symbol:" + symbol) }
	disp := &fakeDisplay{}
	disp.onCall = func(symbol string) { record("display:" + symbol) }
	msync := &fakeMenuSync{}
	msync.onMode = func(symbol string) { record("menu:" + symbol) }
	sound := &fakeSoundSink{}
	sound.onFlip = func() { record("sound:flip") }
	a.SetOptions(flipSoundOn())
	a.SetMenuSync(msync)
	a.SetModeDisplay(disp)
	a.SetSoundSink(sound)
	opMu.Lock()
	ops = nil // the install pushes are pinned elsewhere; only the flip order matters here
	opMu.Unlock()

	flipMode(a) // EN → RU

	opMu.Lock()
	defer opMu.Unlock()
	want := []string{opSymbolRU, "display:ru", "menu:ru", "sound:flip"}
	if !slices.Equal(ops, want) {
		t.Errorf("flip op order = %q, want exactly %q — the tone fires after every observer, D-36 intact", ops, want)
	}
}

// The sound-gating corpus of plan 07-08 (task 3): the fired tone, the
// single gate ahead of the sink, the snapshot fold and the nil-sink
// degradation.

// soundOn returns the enabled-autocorrect Options with the sound switch ON
// — the fire-tone cells' configuration (acOptions plus the switch).
func soundOn() session.Options {
	opts := acOptions()
	opts.SoundEnabled = true

	return opts
}

// TestActor_AutoCorrectFireSoundsSink pins the fired tone (plan 07-08): a
// fired correction sounds exactly one AutoCorrect from the fired point —
// the abstentions never reach the sink.
func TestActor_AutoCorrectFireSoundsSink(t *testing.T) {
	a, sink := wiredActor()
	sound := &fakeSoundSink{}
	a.UseAppid(fakeAppid{app: acListedApp})
	a.SetOptions(soundOn())
	a.SetSoundSink(sound)

	fireAutocorrect(t, a, sink, wordEN, wordRU)

	if got := sound.fireCount(); got != 1 {
		t.Errorf("autocorrect tones after a fired correction = %d, want exactly 1", got)
	}
}

// TestActor_AutoCorrectAbstainSilent pins the abstention's silence (plan
// 07-08): an app-blocked boundary refuses before the fired point — zero
// tones, the fail-closed abstention stays the only trace.
func TestActor_AutoCorrectAbstainSilent(t *testing.T) {
	a, _ := wiredActor()
	sound := &fakeSoundSink{}
	a.UseAppid(fakeAppid{app: acListedApp})
	a.SetOptions(acBlockedOptions()) // the fixture app matches the blocklist
	a.SetSoundSink(sound)

	typeWord(a, wordEN)
	a.HandleKey(spaceKey())

	if got := sound.fireCount(); got != 0 {
		t.Errorf("autocorrect tones after an app-blocked abstention = %d, want 0", got)
	}
	if got := a.AutoCorrectCounters().Reasons[acReasonAppBlocked]; got != 1 {
		t.Errorf("app-blocked abstentions = %d, want 1 — the refusal must remain counted", got)
	}
}

// TestActor_SoundDisabledNoSink pins the single gate (plan 07-08): with
// Options.SoundEnabled off, NO gesture reaches the sink — the gate sits
// ahead of the sink, so no binary lookup ever runs for a muted daemon.
func TestActor_SoundDisabledNoSink(t *testing.T) {
	a, sink := wiredActor()
	sound := &fakeSoundSink{}
	a.UseAppid(fakeAppid{app: acListedApp})
	opts := soundOn()
	opts.SoundEnabled = false
	a.SetOptions(opts)
	a.SetSoundSink(sound)

	fireAutocorrect(t, a, sink, wordEN, wordRU) // a live fire — must stay silent
	flipMode(a)                                 // a live flip — must stay silent

	if got := sound.flipCount(); got != 0 {
		t.Errorf("flip tones with the sound off = %d, want 0", got)
	}
	if got := sound.fireCount(); got != 0 {
		t.Errorf("autocorrect tones with the sound off = %d, want 0", got)
	}
}

// TestActor_NilSoundSinkNoOp pins the nil degradation (the SetMenuSync nil
// form): no sink installed — every flip and every fire runs byte-as-today,
// the tones silently absent.
func TestActor_NilSoundSinkNoOp(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.SetOptions(soundOn())

	fireAutocorrect(t, a, sink, wordEN, wordRU)
	flipMode(a)

	if got := a.AutoCorrectCounters().Fired; got != 1 {
		t.Errorf("fired corrections with no sink = %d, want 1 — the correction is unaffected", got)
	}
	if !strings.Contains(buf.String(), `"msg":"mode","to":"ru"`) {
		t.Errorf("mode record missing with no sink; log:\n%s", buf.String())
	}
}

// TestActor_SoundFoldGatesFromSnapshot pins the fold (plan 07-08): the
// applied document's sound switch gates the tones — default ON lands as
// ON, an explicit enabled: false mutes the next flip WITHOUT a restart,
// and a reload back to ON sounds again (the D-32 hot-reload contour).
func TestActor_SoundFoldGatesFromSnapshot(t *testing.T) {
	a, _ := wiredActor()
	sound := &fakeSoundSink{}
	cfg := config.Defaults()
	src := &reloadSource{cfg: cfg}
	a.AttachConfig(src)
	a.SetOptions(soundOn())
	a.SetSoundSink(sound)

	flipMode(a) // the defaults document reads ON — the tone fires
	if got := sound.flipCount(); got != 1 {
		t.Fatalf("flip tones under the default document = %d, want 1 — sound is default ON", got)
	}

	off := false
	cfg.Sound.Enabled = &off
	src.set(cfg)
	flipMode(a) // the muted document — the fold gates before the sink
	if got := sound.flipCount(); got != 1 {
		t.Errorf("flip tones after the muted reload = %d, want still 1 — the gate folds from the snapshot", got)
	}

	on := true
	cfg.Sound.Enabled = &on
	src.set(cfg)
	flipMode(a) // re-enabled — the tone returns
	if got := sound.flipCount(); got != 2 {
		t.Errorf("flip tones after the re-enabled reload = %d, want 2", got)
	}
}

// TestActor_SoundFoldOffStopsSink pins the mute lifecycle's close half
// (261006-squ, requirement 3): a fold that transitions the sound switch
// ON→OFF calls the installed sink's stop EXACTLY ONCE — muted sound means
// NO sound-server connection, the lazy redial rides the next tone.
func TestActor_SoundFoldOffStopsSink(t *testing.T) {
	a, _ := wiredActor()
	sound := &fakeSoundSink{}
	cfg := config.Defaults()
	src := &reloadSource{cfg: cfg}
	a.AttachConfig(src)
	a.SetOptions(soundOn())
	a.SetSoundSink(sound)

	flipMode(a) // ON — the fold has no transition to close
	if got := sound.stopCount(); got != 0 {
		t.Fatalf("stops after an ON fold = %d, want 0", got)
	}

	off := false
	cfg.Sound.Enabled = &off
	src.set(cfg)
	flipMode(a) // the ON→OFF transition — the close trigger
	if got := sound.stopCount(); got != 1 {
		t.Errorf("stops after the ON→OFF fold = %d, want exactly 1 — the connection must close", got)
	}

	flipMode(a) // still OFF — no further transition
	if got := sound.stopCount(); got != 1 {
		t.Errorf("stops after a second muted fold = %d, want still 1 — only the transition closes", got)
	}
}

// TestActor_SoundFoldStaysOffStopsNothingExtra pins the OFF→OFF fold: a
// document that arrives already muted stops nothing — the close is the
// TRANSITION's, never the state's.
func TestActor_SoundFoldStaysOffStopsNothingExtra(t *testing.T) {
	a, _ := wiredActor()
	sound := &fakeSoundSink{}
	cfg := config.Defaults()
	off := false
	cfg.Sound.Enabled = &off
	src := &reloadSource{cfg: cfg}
	a.AttachConfig(src)
	opts := soundOn()
	opts.SoundEnabled = false
	a.SetOptions(opts)
	a.SetSoundSink(sound)

	flipMode(a) // OFF→OFF — nothing to close
	if got := sound.stopCount(); got != 0 {
		t.Errorf("stops after an OFF→OFF fold = %d, want 0", got)
	}

	on := true
	cfg.Sound.Enabled = &on
	src.set(cfg)
	flipMode(a) // OFF→ON — the lazy redial rides the next tone, not a stop
	if got := sound.stopCount(); got != 0 {
		t.Errorf("stops after the OFF→ON fold = %d, want 0 — enable never stops", got)
	}
}

// TestActor_SinkWithoutStopCapabilityFoldsSilently pins the optional
// capability: a sink that only satisfies SoundSink (type-assertion miss
// on SoundStopper) folds silently — the mute close degrades to a no-op,
// byte-as-today for every pre-existing sink.
type bareSoundSink struct {
	mu    sync.Mutex
	flips int
}

func (b *bareSoundSink) Flip() {
	b.mu.Lock()
	b.flips++
	b.mu.Unlock()
}

func (*bareSoundSink) AutoCorrect() {}

func (*bareSoundSink) SetAutocorrectEvent(string) {}

func TestActor_SinkWithoutStopCapabilityFoldsSilently(t *testing.T) {
	a, _ := wiredActor()
	sound := &bareSoundSink{}
	cfg := config.Defaults()
	src := &reloadSource{cfg: cfg}
	a.AttachConfig(src)
	a.SetOptions(soundOn())
	a.SetSoundSink(sound)

	off := false
	cfg.Sound.Enabled = &off
	src.set(cfg)
	flipMode(a) // the ON→OFF transition against a stop-less sink — no panic, the fold proceeds

	if got := sound.flips; got != 0 {
		t.Errorf("flip tones after the muted fold = %d, want 0 — the gate held", got)
	}
}

// TestActor_NilSinkFoldSilent pins the nil degradation through the mute
// fold: no sink installed — the ON→OFF transition folds byte-as-today,
// silently.
func TestActor_NilSinkFoldSilent(t *testing.T) {
	a, _ := wiredActor()
	cfg := config.Defaults()
	src := &reloadSource{cfg: cfg}
	a.AttachConfig(src)
	a.SetOptions(soundOn())

	off := false
	cfg.Sound.Enabled = &off
	src.set(cfg)
	flipMode(a) // the ON→OFF transition with no sink — no panic

	if got := a.StatusSnapshot().SoundEnabled; got {
		t.Errorf("sound option after the muted fold = true, want false — the fold applied")
	}
}

// TestActor_FoldAppliedConfig pins the menu toggle's synchronous-apply
// seam (plan 07-06, the 07-05 pin live-proven by the menu-v2 case): the
// toggle's reload re-stores the watcher snapshot, and FoldAppliedConfig
// puts the applied values in force WITHOUT waiting for the next key event
// — the fold is the ONE application path (the 03-04 succession), only
// invoked eagerly. A sourceless actor folds as a quiet no-op.
func TestActor_FoldAppliedConfig(t *testing.T) {
	a, _ := wiredActor()
	cfg := config.Defaults()
	on := true
	cfg.Autocorrect.Enabled = false
	src := &reloadSource{cfg: cfg}
	a.AttachConfig(src)
	opts := soundOn()
	opts.AutoCorrectEnabled = false
	opts.SoundEnabled = false
	a.SetOptions(opts)

	cfg.Autocorrect.Enabled = true
	cfg.Sound.Enabled = &on
	src.set(cfg)
	a.FoldAppliedConfig() // the toggle's reload fold — no key event in between

	st := a.StatusSnapshot()
	if !st.AutoCorrectEnabled {
		t.Errorf("autocorrect option after the eager fold = false, want true — the apply must not wait for a key event")
	}
	if !st.SoundEnabled {
		t.Errorf("sound option after the eager fold = false, want true")
	}

	bare, _ := wiredActor()
	bare.SetOptions(opts)
	bare.FoldAppliedConfig()
	if st := bare.StatusSnapshot(); st.AutoCorrectEnabled {
		t.Errorf("autocorrect option after a sourceless fold = true, want the SetOptions value false")
	}
}

// The a11y-sink corpus of plan 08-05, rewritten by the 2026-10-06 owner
// revision (plan 08-09): the accessibility-magic seam — the fold delivers
// the a11y section's EffectiveEnabled() truth to the sink ONLY on a change
// of that boolean (the refreshACBlocklist diff-gate form), a late install
// self-syncs the last folded state (the SetSoundSink precedent), and a nil
// sink stays a silent no-op. There is no app list in the schema any more:
// the fold reads the single global switch (nil decodes ON — the owner's
// default-ON verdict, D-8-6/REV).

// fakeA11ySink is the actor's A11ySink double (plan 08-05): every pushed
// state recorded under a mutex — the fold's diff-gate observability.
type fakeA11ySink struct {
	mu     sync.Mutex
	pushes []bool
}

// Apply records one pushed a11y state.
func (f *fakeA11ySink) Apply(active bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.pushes = append(f.pushes, active)
}

// pushed returns the recorded push history.
func (f *fakeA11ySink) pushed() []bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return slices.Clone(f.pushes)
}

// a11yDoc returns the defaults document with the a11y switch set to the
// given value — the fold corpus's feed (a complete valid document, the
// reloadCfg shape). The pointer form mirrors the schema: an explicit bool
// is the document's own verdict, the nil pointer the absent-key shape.
func a11yDoc(enabled bool) config.Config {
	cfg := config.Defaults()
	cfg.A11y.Enabled = &enabled

	return cfg
}

// TestActor_A11yDefaultsDocumentPushesTrue pins the default-ON fold (the
// owner's 2026-10-06 revision in action, D-8-6/REV): the fold of the
// built-in defaults document — whose a11y switch decodes ON — delivers
// exactly one Apply(true) to the reconciler.
func TestActor_A11yDefaultsDocumentPushesTrue(t *testing.T) {
	a, _ := wiredActor()
	sink := &fakeA11ySink{}
	src := &reloadSource{cfg: config.Defaults()}
	a.AttachConfig(src)
	a.SetA11ySink(sink)

	a.FoldAppliedConfig()

	if got := sink.pushed(); !slices.Equal(got, []bool{true}) {
		t.Errorf("a11y pushes after the defaults fold = %v, want exactly [true] (default ON)", got)
	}
}

// TestActor_A11yFoldTriggersSinkOnActivation pins the activation push
// (plan 08-05): a fold of a document with the switch enabled delivers
// exactly one Apply(true), and a REPEATED fold of the same state stays
// silent — the diff gate keeps dconf churn and journal noise at zero at
// any fold rate (the startup fold of newActor's FoldAppliedConfig reaches
// the reconciler the same way).
func TestActor_A11yFoldTriggersSinkOnActivation(t *testing.T) {
	a, _ := wiredActor()
	sink := &fakeA11ySink{}
	cfg := a11yDoc(true)
	src := &reloadSource{cfg: cfg}
	a.AttachConfig(src)
	a.SetA11ySink(sink)

	a.FoldAppliedConfig() // the activation — the only push
	src.set(cfg)
	a.FoldAppliedConfig() // the same state — silence

	if got := sink.pushed(); !slices.Equal(got, []bool{true}) {
		t.Errorf("a11y pushes after the active and a repeated fold = %v, want exactly [true] — "+
			"the gate is on the boolean", got)
	}
}

// TestActor_A11yExplicitFalsePushesFalse pins the deactivation push (plan
// 08-05): a fold of a document with an EXPLICIT enabled: false after an
// active one delivers Apply(false) — the only OFF shape of the revised
// section. The 08-03 reconciler itself decides that a false touches
// nothing (D-8-4, pinned by its own corpus); the actor's job is only to
// deliver the change.
func TestActor_A11yExplicitFalsePushesFalse(t *testing.T) {
	a, _ := wiredActor()
	sink := &fakeA11ySink{}
	src := &reloadSource{cfg: a11yDoc(true)}
	a.AttachConfig(src)
	a.SetA11ySink(sink)

	a.FoldAppliedConfig() // active
	src.set(a11yDoc(false))
	a.FoldAppliedConfig() // explicit false — the change must propagate

	if got := sink.pushed(); !slices.Equal(got, []bool{true, false}) {
		t.Errorf("a11y pushes after activate then explicit false = %v, want exactly [true false]", got)
	}
}

// TestActor_A11ySinkInstallSelfSyncs pins the late install (plan 08-05,
// the SetSoundSink self-sync form): the startup fold runs with NO sink
// installed (the OnConn ordering — the wiring attaches later), and the
// SetA11ySink install pushes the last folded state immediately, so the
// reconcile-at-start does not depend on the install order.
func TestActor_A11ySinkInstallSelfSyncs(t *testing.T) {
	a, _ := wiredActor()
	sink := &fakeA11ySink{}
	src := &reloadSource{cfg: a11yDoc(true)}
	a.AttachConfig(src)

	a.FoldAppliedConfig() // the startup fold — no sink installed yet

	a.SetA11ySink(sink) // the late install self-syncs the folded state

	if got := sink.pushed(); !slices.Equal(got, []bool{true}) {
		t.Errorf("a11y pushes after a late install over an active fold = %v, want exactly [true] — "+
			"the install pushes immediately", got)
	}
}

// TestActor_A11yNilSinkNoOp pins the nil degradation (the SetMenuSync nil
// form): folds with no sink installed run byte-as-today — no panic, no
// effect, the silent absence of pushes.
func TestActor_A11yNilSinkNoOp(t *testing.T) {
	a, _ := wiredActor()
	src := &reloadSource{cfg: a11yDoc(true)}
	a.AttachConfig(src)

	a.FoldAppliedConfig() // active with no sink — silent
	src.set(a11yDoc(false))
	a.FoldAppliedConfig() // deactivated with no sink — still silent

	if st := a.StatusSnapshot(); !st.SoundEnabled {
		t.Errorf("status unreadable after nil-sink folds — the folds must run byte-as-today " +
			"(the defaults document reads sound ON)")
	}
}

// The terminal-class corpus of plan 10-03 Task 1 (TERM-04/D-55): the
// behavioral verify-dead source of the classification seam lives and dies
// with the input context — learned when a verify round of the LIVE context
// expires by budget, cleared by the context's death (FocusOut/Reset) — and
// every class transition writes ONE source-neutral INFO marker (the A5
// insurance) without a word of content (D-20/D-21).

// teachVerifyDead drives one manual Double whose pre-correction verify
// budget expires while the context is still alive — the exact behavioral
// signature the seam learns (RESEARCH Pattern 1, source 3). The callers
// own every assertion; this helper only drives the deterministic round
// through the verify-deadline seam (plan 10-10, UseVerifyTimer): the
// recorder fires the SAME epoch-tagged closure the real timer would.
func teachVerifyDead(t *testing.T, a *session.Actor, word string) {
	t.Helper()
	clock := &flipClock{} // reused as the one-shot verify-deadline recorder (the arm/fire/stop shape)
	a.UseVerifyTimer(clock.arm)
	typeWord(a, word)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	clock.fire() // deterministic injection of the 100 ms deadline
}

// TestTerminalClass_VerifyDeadLearnedOnVerifyTimeout pins the learning: a
// verify round that expires by budget while the context is still alive
// marks the context verify-dead — exactly one transition marker — and the
// class PERSISTS: a second expired round on the same context is no news
// (no second marker).
func TestTerminalClass_VerifyDeadLearnedOnVerifyTimeout(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()

	teachVerifyDead(t, a, wordEN)

	if got := strings.Count(buf.String(), `"reason":"verify-timeout"`); got != 1 {
		t.Fatalf("verify-timeout records = %d, want 1 — the round must still drop exactly as before", got)
	}
	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 1 {
		t.Fatalf("terminal-class markers after the first verify-timeout = %d, want exactly 1; log:\n%s",
			got, buf.String())
	}
	if !strings.Contains(buf.String(), `"verify_dead":true`) {
		t.Errorf("the learned-class marker is missing; log:\n%s", buf.String())
	}

	// A second round on the same context. Plan 10-04 migration (owner
	// verdict OQ 1 «unified»): a LEARNED verify-dead context routes the
	// next Double through the terminal write branch — the verify round
	// never arms again, so a second expiry is unreachable; the marker
	// discipline stands unchanged: no transition, no marker.
	typeWord(a, "cnhjrf")
	a.HandleKey(spaceKey())
	teachVerifyDead(t, a, wordEN)

	if got := strings.Count(buf.String(), `"reason":"verify-timeout"`); got != 1 {
		t.Fatalf("verify-timeout records after the second round = %d, want 1 — the learned class routes "+
			"the round through the terminal branch (plan 10-04), nothing re-arms the verify machine", got)
	}
	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 1 {
		t.Errorf("terminal-class markers after the second round = %d, want still 1 — "+
			"no transition, no marker", got)
	}
}

// flapClock is the deterministic double of the flap-window timer seam
// (plan 10-07, the UseFlapTimer form): arm records the arm/stop counts and
// holds the single pending callback; fire() plays the window's expiry, and
// stop() cancels it — the FocusIn and immediate-wipe paths run through the
// same stop closure the actor holds.
type flapClock struct {
	mu      sync.Mutex
	armed   int
	stopped int
	pending func()
}

// arm records one armed window and returns the actor's stop closure.
func (f *flapClock) arm(_ time.Duration, fn func()) func() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.armed++
	f.pending = fn

	return func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.stopped++
		f.pending = nil
	}
}

// fire plays the window expiry: a cancelled or superseded window holds no
// callback, so a stale fire is a no-op — exactly the production guard.
func (f *flapClock) fire() {
	f.mu.Lock()
	fn := f.pending
	f.pending = nil
	f.mu.Unlock()
	if fn != nil {
		fn()
	}
}

// counts snapshots the arm/stop counters.
func (f *flapClock) counts() (armed, stopped int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.armed, f.stopped
}

// TestTerminalClass_VerifyDeadSurvivesFlapPair pins the flap tolerance of
// plan 10-07 (G-10-1, cause A): a REAL FocusOut followed by a FocusIn
// within the flap window — the host kernel's same-engine focus pairs
// (journal 2026-10-13 16:20:05: pairs 1–3 ms apart, four within 6 ms) —
// must NOT wipe the learned terminal class: the wipe is deferred by the
// window and the completing FocusIn cancels it. The class stays alive as a
// fact, not just unlogged: the next expired round routes through the
// terminal write branch (no verify round, no second verify-timeout — the
// plan-10-04 routing contract).
func TestTerminalClass_VerifyDeadSurvivesFlapPair(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()

	teachVerifyDead(t, a, wordEN)
	a.HandleLifecycle(engine.LifecycleFocusOut)
	a.HandleLifecycle(engine.LifecycleFocusIn) // the pair completes inside the window

	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 1 {
		t.Fatalf("terminal-class markers after the flap pair = %d, want exactly 1 (learn only) — "+
			"an instant same-engine pair must not wipe the learned class (G-10-1); log:\n%s",
			got, buf.String())
	}

	// The class is functionally alive: the next Double rides it through the
	// terminal branch — the verify machine never arms again (plan 10-04).
	typeWord(a, "cnhjrf")
	a.HandleKey(spaceKey())
	teachVerifyDead(t, a, wordEN)

	if got := strings.Count(buf.String(), `"reason":"verify-timeout"`); got != 1 {
		t.Fatalf("verify-timeout records after the flap pair = %d, want 1 — the learned class must "+
			"still route the next gesture through the terminal branch; log:\n%s", got, buf.String())
	}
}

// TestTerminalClass_VerifyDeadClearedByFocusOut pins the clearing side of
// the lifecycle (the plan's two separate clearing tests, FocusOut first),
// re-pinned by plan 10-07 on the deferred form: the wipe a REAL FocusOut
// executes is the window's expiry — no FocusIn inside focusFlapWindow
// means a real focus loss, and the class clears exactly as before (one
// clearing marker, and the cleared context learns afresh on its next
// expired round — the state, not just the log, was reset).
func TestTerminalClass_VerifyDeadClearedByFocusOut(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	clock := &flapClock{}
	a.UseFlapTimer(clock.arm)

	teachVerifyDead(t, a, wordEN)
	a.HandleLifecycle(engine.LifecycleFocusOut)
	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 1 {
		t.Fatalf("terminal-class markers before the window expiry = %d, want 1 — "+
			"the wipe waits the flap window out; log:\n%s", got, buf.String())
	}
	clock.fire() // the window closed without a FocusIn — a real focus loss

	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 2 {
		t.Fatalf("terminal-class markers after the window expiry = %d, want exactly 2 (learn + clear); log:\n%s",
			got, buf.String())
	}
	if !strings.Contains(buf.String(), `"verify_dead":false`) {
		t.Errorf("the cleared-class marker is missing; log:\n%s", buf.String())
	}

	// The fresh context learns afresh: the second expiry transitions again.
	typeWord(a, "cnhjrf")
	a.HandleKey(spaceKey())
	teachVerifyDead(t, a, wordEN)

	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 3 {
		t.Errorf("terminal-class markers after the re-learn = %d, want 3 — "+
			"the real focus loss really cleared the class, not just logged", got)
	}
}

// TestTerminalClass_VerifyDeadSurvivesSyntheticFocusOut pins the G-2-3
// carve-out's behavioral-class arm (debug session orca-terminal-second-shift,
// green106 2026-10-31, re-pinned from the plan 10-07 immediate-wipe form the
// live journal falsified): a FocusOut carrying the own flip's synthetic
// credit is NOT a context death — the re-mint serves the same client field,
// the class rides, and the wipe never defers either (no window armed, no
// stop pending).
func TestTerminalClass_VerifyDeadSurvivesSyntheticFocusOut(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	clock := &flapClock{}
	a.UseFlapTimer(clock.arm)
	a.SetSwitcher(func(_ context.Context, _ string) error { return nil })

	teachVerifyDead(t, a, wordEN)
	flipMode(a)                                 // EN → RU: the successful round trip arms exactly one credit
	a.HandleLifecycle(engine.LifecycleFocusOut) // the synthetic pair's FocusOut consumes it

	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 1 {
		t.Fatalf("terminal-class markers after the synthetic FocusOut = %d, want exactly 1 (learn only) — "+
			"the own flip's re-mint is not a context death; log:\n%s", got, buf.String())
	}
	if armed, stopped := clock.counts(); armed != 0 {
		t.Errorf("flap windows armed on the synthetic path = %d, want 0 — "+
			"the synthetic path never defers either", armed)
	} else if stopped != 0 {
		t.Errorf("flap windows stopped on the synthetic path = %d, want 0", stopped)
	}

	// The class is functionally alive: the next Double rides it through the
	// terminal branch — the verify machine never arms again (plan 10-04).
	typeWord(a, "cnhjrf")
	a.HandleKey(spaceKey())
	teachVerifyDead(t, a, wordEN)

	if got := strings.Count(buf.String(), `"reason":"verify-timeout"`); got != 1 {
		t.Fatalf("verify-timeout records after the synthetic FocusOut = %d, want 1 — the learned class must "+
			"still route the next gesture through the terminal branch; log:\n%s", got, buf.String())
	}
}

// TestTerminalClass_VerifyDeadSurvivesOwnFlip pins the live-journal finding
// of debug session orca-terminal-second-shift (green106, 2026-10-31
// 01:02:39..01:34:34, 10/10 repros): the daemon's OWN flip — the deferred
// terminal flip a successful terminal correction arms, released by the
// user's next keystroke — re-mints the engine object, and the synthetic
// focus_out must NOT wipe the learned behavioral class. The SetGlobalEngine
// re-mint serves the SAME client field (the journal's Orca IDE terminal,
// push-less): the client's verify behavior is unchanged by the daemon's own
// flip — the G-2-3 principle the buffer hard-reset already rides. The wipe
// is the stationary «correction lands only on the second double-shift»
// cycle: every successful correction erases the class, so the next one
// re-pays a verify round a push-less terminal always loses.
func TestTerminalClass_VerifyDeadSurvivesOwnFlip(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	clock := &flapClock{}
	a.UseFlapTimer(clock.arm)
	a.SetOptions(session.Options{FlipAfterCorrection: true})
	a.SetSwitcher(func(_ context.Context, _ string) error { return nil })

	// Gesture 1 — the verify round expires on the living context: the class
	// is learned, the round skipped (the journal's first press).
	teachVerifyDead(t, a, wordEN)
	if got := strings.Count(buf.String(), `"reason":"verify-timeout"`); got != 1 {
		t.Fatalf("verify-timeout records after the teaching round = %d, want 1; log:\n%s",
			got, buf.String())
	}

	// Gesture 2 — rides the learned class: the terminal write branch
	// delivers, and the terminal flip arms (the release waits for the next
	// keystroke — the journal's flip-armed).
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	if !strings.Contains(buf.String(), `"msg":"correction delivery","stage":"flip-armed"`) {
		t.Fatalf("the terminal flip never armed — gesture 2 must deliver through the "+
			"terminal write branch off the learned class; log:\n%s", buf.String())
	}

	// The user keeps typing: the keystroke releases the flip (the
	// SetGlobalEngine round trip succeeds and arms exactly one credit), and
	// the re-mint's synthetic focus_out/focus_in pair arrives.
	a.HandleKey(spaceKey())                     // the flip's release keystroke
	a.HandleLifecycle(engine.LifecycleFocusOut) // the synthetic pair's FocusOut consumes the credit
	a.HandleLifecycle(engine.LifecycleFocusIn)

	// Gesture 3 — the NEXT correction (the journal's first press of the next
	// pair): the learned class must still route it through the terminal
	// branch — no verify round re-arms (the helper's deadline fire is a
	// no-op then), no second verify-timeout, the delivery commits.
	teachVerifyDead(t, a, wordEN)

	logged := buf.String()
	if got := strings.Count(logged, `"reason":"verify-timeout"`); got != 1 {
		t.Fatalf("verify-timeout records after the own flip = %d, want 1 — the own flip's re-mint "+
			"must not wipe the learned class (the two-press cycle); log:\n%s", got, logged)
	}
	if got := strings.Count(logged, `"msg":"terminal class"`); got != 1 {
		t.Fatalf("terminal-class markers = %d, want exactly 1 (learn only) — the own flip is not a "+
			"context change for the behavioral class; log:\n%s", got, logged)
	}
	if got := strings.Count(logged, `"msg":"correction","outcome":"done","changed":true`); got != 2 {
		t.Fatalf("done records = %d, want 2 — gestures 2 and 3 both deliver; log:\n%s", got, logged)
	}
	if armed, stopped := clock.counts(); armed != 0 || stopped != 0 {
		t.Errorf("flap windows armed/stopped = %d/%d, want 0/0 — the synthetic path never defers",
			armed, stopped)
	}
}

// TestTerminalClass_VerifyDeadSurvivesClientResetAfterOwnFlip pins the Round
// 2 finding of debug session orca-terminal-second-shift (green106, owner
// acceptance 2026-10-31 13:55 «работает со второго», the Round 1 fix
// 946b4cd already in main): the synthetic carve-out survives the re-mint's
// focus pair, but ~0.9 s later the SLOW CLIENT (Electron ibus-IM, the
// journal's Orca IDE terminal) reconnects its IM context and delivers a
// Reset on the re-minted engine object — and the Reset branch wipes the
// class the carve-out had just saved (the acceptance row's wipe 41.874,
// 923 ms after the flip, followed by the same-engine flap pair; the
// machine-verified journal elimination: all 22 wipe rows of the day ran the
// Reset branch). Within the re-mint aftermath window the Reset is the
// re-mint's own echo — the same client field re-engaging: the class
// survives and the next correction rides it on the FIRST press.
func TestTerminalClass_VerifyDeadSurvivesClientResetAfterOwnFlip(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	aftermath := &flapClock{}
	a.UseFlipAftermathTimer(aftermath.arm)
	a.SetOptions(session.Options{FlipAfterCorrection: true})
	a.SetSwitcher(func(_ context.Context, _ string) error { return nil })

	// Gesture 1 — the verify round expires on the living context: the class
	// is learned, the round skipped (the journal's first press).
	teachVerifyDead(t, a, wordEN)
	if got := strings.Count(buf.String(), `"reason":"verify-timeout"`); got != 1 {
		t.Fatalf("verify-timeout records after the teaching round = %d, want 1; log:\n%s",
			got, buf.String())
	}

	// Gesture 2 — rides the learned class: the terminal write branch
	// delivers, and the terminal flip arms (the release waits for the next
	// keystroke — the journal's flip-armed).
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	if !strings.Contains(buf.String(), `"msg":"correction delivery","stage":"flip-armed"`) {
		t.Fatalf("the terminal flip never armed — gesture 2 must deliver through the "+
			"terminal write branch off the learned class; log:\n%s", buf.String())
	}

	// The keystroke releases the flip (a successful SetGlobalEngine round
	// trip: one credit AND the aftermath window), the re-mint's synthetic
	// pair arrives — and then the client's late Reset lands INSIDE the
	// aftermath window (the journal's +923 ms echo).
	a.HandleKey(spaceKey())
	a.HandleLifecycle(engine.LifecycleFocusOut) // the synthetic pair's FocusOut consumes the credit
	a.HandleLifecycle(engine.LifecycleFocusIn)
	if armed, _ := aftermath.counts(); armed != 1 {
		t.Fatalf("aftermath windows armed by the flip = %d, want 1 — the successful round trip "+
			"arms the echo window next to the credit", armed)
	}
	a.HandleLifecycle(engine.LifecycleReset) // the client's re-mint echo, inside the window

	// Gesture 3 — the NEXT correction (the journal's press after the echo):
	// the class survived the client Reset — no verify round re-arms, no
	// second verify-timeout, the delivery commits.
	teachVerifyDead(t, a, wordEN)

	logged := buf.String()
	if got := strings.Count(logged, `"reason":"verify-timeout"`); got != 1 {
		t.Fatalf("verify-timeout records after the client Reset = %d, want 1 — the re-mint "+
			"aftermath Reset must not wipe the learned class (the two-press cycle); log:\n%s",
			got, logged)
	}
	if got := strings.Count(logged, `"msg":"terminal class"`); got != 1 {
		t.Fatalf("terminal-class markers = %d, want exactly 1 (learn only) — the aftermath "+
			"Reset is not a context change for the behavioral class; log:\n%s", got, logged)
	}
	if got := strings.Count(logged, `"msg":"correction","outcome":"done","changed":true`); got != 2 {
		t.Fatalf("done records = %d, want 2 — gestures 2 and 3 both deliver; log:\n%s", got, logged)
	}
}

// TestTerminalClass_VerifyDeadClearedByResetOutsideAftermath pins the
// fail-closed arm of the Round 2 carve-out: a Reset arriving AFTER the
// re-mint aftermath window closed is a REAL context change and wipes the
// learned class immediately, byte as before — the next correction re-pays
// the verify round a push-less terminal always loses. The aftermath window
// must never grow into a class-immunity blanket: its only tenant is the
// re-mint echo of the daemon's own flip.
func TestTerminalClass_VerifyDeadClearedByResetOutsideAftermath(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	aftermath := &flapClock{}
	a.UseFlipAftermathTimer(aftermath.arm)
	a.SetOptions(session.Options{FlipAfterCorrection: true})
	a.SetSwitcher(func(_ context.Context, _ string) error { return nil })

	teachVerifyDead(t, a, wordEN) // the class learns (the first press)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow) // gesture 2 delivers, the terminal flip arms
	a.HandleKey(spaceKey())       // the flip's release keystroke: credit + aftermath window
	a.HandleLifecycle(engine.LifecycleFocusOut)
	a.HandleLifecycle(engine.LifecycleFocusIn)

	aftermath.fire()                         // the window closed — the echo never came
	a.HandleLifecycle(engine.LifecycleReset) // a REAL Reset: the immediate wipe stands

	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 2 {
		t.Fatalf("terminal-class markers after the outside Reset = %d, want exactly 2 (learn + clear); log:\n%s",
			got, buf.String())
	}
	if !strings.Contains(buf.String(), `"verify_dead":false`) {
		t.Errorf("the cleared-class marker is missing; log:\n%s", buf.String())
	}

	// The fresh context learns afresh: the next correction re-pays the
	// verify round (the fail-closed cost of a real context change).
	typeWord(a, "cnhjrf")
	a.HandleKey(spaceKey())
	teachVerifyDead(t, a, wordEN)

	if got := strings.Count(buf.String(), `"reason":"verify-timeout"`); got != 2 {
		t.Errorf("verify-timeout records = %d, want 2 — the outside Reset really cleared the "+
			"class and the next correction re-pays the verify round (fail-closed kept); log:\n%s",
			got, buf.String())
	}
}

// TestTerminalClass_FlapWindowRearmsNotBreeds pins the single-window
// discipline (T-10-07-04): a second real FocusOut while a window is pending
// RE-ARMS the same window (arm 2, stop 1), the completing FocusIn cancels
// it (stop 2, nothing pending), and a stale expiry afterwards wipes nothing
// — the class survives.
func TestTerminalClass_FlapWindowRearmsNotBreeds(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	clock := &flapClock{}
	a.UseFlapTimer(clock.arm)

	teachVerifyDead(t, a, wordEN)
	a.HandleLifecycle(engine.LifecycleFocusOut)
	a.HandleLifecycle(engine.LifecycleFocusOut) // a flap burst: re-arm, never a second timer
	a.HandleLifecycle(engine.LifecycleFocusIn)  // the pair completes

	if armed, stopped := clock.counts(); armed != 2 || stopped != 2 {
		t.Fatalf("flap windows armed/stopped = %d/%d, want 2/2 — "+
			"a repeated FocusOut re-arms the same window, the FocusIn cancels it", armed, stopped)
	}
	clock.fire() // a stale expiry — the cancelled window owns nothing

	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 1 {
		t.Errorf("terminal-class markers after the burst + stale fire = %d, want 1 — "+
			"the cancelled window wipes nothing; log:\n%s", got, buf.String())
	}
}

// TestTerminalBranch_FlapPairBetweenChunksCommits pins the rewritten
// context discipline (plan 10-17, the ADR-008 amendment (6)): the
// generation observability survives — a flap pair landing in a
// chunk-boundary window still bumps ctxEpoch — but after the first
// destructive key the round can only go FORWARD: the burst completes, the
// commit follows on the captured snapshot, and the boundary is loud (the
// mid-flow delivery mark). The free stop survives only BEFORE the first
// forward (TestReplayWezterm1921_ContextChangeBeforeFirstForwardStops).
func TestTerminalBranch_FlapPairBetweenChunksCommits(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	teachTerminalClass(t, a)

	typeWord(a, termWord20)
	var landed sync.Once
	sent := atomic.Int32{}
	sink.forwardHook = func(forwardCall) {
		// 2×termChunkKeys emissions = the FIRST chunk's pairs complete
		// (round-6: one key = press + release) — the boundary moment.
		if sent.Add(1) == 2*termChunkKeys {
			landed.Do(func() {
				go func() {
					a.HandleLifecycle(engine.LifecycleFocusOut)
					// the pair completes — the class survives, the generation moved
					a.HandleLifecycle(engine.LifecycleFocusIn)
				}()
			})
		}
	}
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if n := len(sink.forwardCalls()); n != 2*len([]rune(termWord20)) {
		t.Fatalf("forwards = %d, want %d — a mid-flow generation change must not stop the burst "+
			"after the first destructive key", n, 2*len([]rune(termWord20)))
	}
	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != termWord20Converted {
		t.Fatalf("commits = %q, want exactly one %q — a deletion without its replacement is the one "+
			"outcome the machine may never produce", texts, termWord20Converted)
	}
	logged := buf.String()
	if !strings.Contains(logged, `"msg":"correction delivery","stage":"context-changed-midflow"`) {
		t.Errorf("the loud mid-flow boundary is missing — a silent generation change inside the "+
			"round is the quiet-loss class; log:\n%s", logged)
	}
	if !strings.Contains(logged, `"msg":"correction","outcome":"done","changed":true`) {
		t.Errorf("the round never logged its completion; log:\n%s", logged)
	}
}

// TestTerminalClass_VerifyDeadClearedByReset pins the second clearing arm
// (the plan demands FocusOut AND Reset as separate tests): an engine Reset
// retires the context exactly like a focus loss.
func TestTerminalClass_VerifyDeadClearedByReset(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()

	teachVerifyDead(t, a, wordEN)
	a.HandleLifecycle(engine.LifecycleReset)

	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 2 {
		t.Fatalf("terminal-class markers after the Reset = %d, want exactly 2 (learn + clear); log:\n%s",
			got, buf.String())
	}
	if !strings.Contains(buf.String(), `"verify_dead":false`) {
		t.Errorf("the cleared-class marker is missing; log:\n%s", buf.String())
	}

	typeWord(a, "cnhjrf")
	a.HandleKey(spaceKey())
	teachVerifyDead(t, a, wordEN)

	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 3 {
		t.Errorf("terminal-class markers after the re-learn = %d, want 3 — "+
			"the Reset really cleared the class, not just logged", got)
	}
}

// TestTerminalClass_NeverDeadContextClearsSilently pins the transition-only
// discipline from the other side: clearing a context that was never dead is
// no news — zero markers.
func TestTerminalClass_NeverDeadContextClearsSilently(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()

	a.HandleLifecycle(engine.LifecycleFocusOut)
	a.HandleLifecycle(engine.LifecycleReset)

	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 0 {
		t.Errorf("terminal-class markers on a never-dead context = %d, want 0 — "+
			"clearing a clear state must stay silent", got)
	}
}

// TestTerminalClass_StaleVerifyExpiryLearnsNothing pins the stale-timer
// guard: the round settled in time (the client answered), so the deadline
// callback is a no-op — no timeout record, no learned class. The "expired
// budget while the context is still alive" signature requires the budget
// to actually expire.
func TestTerminalClass_StaleVerifyExpiryLearnsNothing(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	typeWord(a, wordEN)
	// The honest-GTK world (plan 10-10 migration, the VerifyPaths form):
	// the client pushed its field state during typing — cache evidence at
	// arm. The nil-cache form of this scenario is the answer-only terminal
	// signature now (TestReplayHost0939_MatchAnswerRoutesTerminalBranch).
	a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	// The fresh push is the freshest report — the round executes.
	end := uint32(len([]rune("abc " + wordEN)))
	a.HandleSurroundingText("abc "+wordEN, end, end)
	a.HandleSurroundingText("abc ", 4, 4) // the deletion proof completes the ladder (plan 10-11)
	if got := len(sink.commitTexts()); got != 1 {
		t.Fatalf("commits after the cached push = %d, want 1 — the round never executed", got)
	}

	a.VerifyExpiry(0) // the stale deadline — a settled round teaches nothing (any epoch: pending is nil)

	if got := strings.Count(buf.String(), `"reason":"verify-timeout"`); got != 0 {
		t.Errorf("verify-timeout records after the stale deadline = %d, want 0", got)
	}
	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 0 {
		t.Errorf("terminal-class markers after the stale deadline = %d, want 0 — "+
			"a settled round teaches nothing", got)
	}
}

// TestTerminalClass_MarkerWordFree pins the privacy half of the marker
// contract (D-20/D-21): the class transitions never carry the typed word
// nor the conversion — counts and the boolean state only.
func TestTerminalClass_MarkerWordFree(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()

	teachVerifyDead(t, a, wordEN)
	a.HandleLifecycle(engine.LifecycleFocusOut)

	logged := buf.String()
	for _, word := range []string{wordEN, wordRU} {
		if strings.Contains(logged, word) {
			t.Errorf("the log leaked %q (D-20/D-21); log:\n%s", word, logged)
		}
	}
}

// The arming-veto corpus of plan 10-03 Task 2 (TERM-04/D-55, ROADMAP
// criterion 5): the terminal class refuses at the ARMING point —
// autoCorrectBoundary, after the blocklist segment, AHEAD of the detector
// — from all three sources (role 60, correction.terminals, verify-dead),
// independent of identity observability (fail-closed), with the confirm's
// role switch untouched as the second belt.

// fakeDetector is the boundary detector's test spy (the plan 10-03 seam,
// UseDetector): it counts the calls and answers with the real Check — the
// veto corpus proves the detector never runs on a terminal class and runs
// exactly as before on a GTK class through the same counter.
type fakeDetector struct {
	mu    sync.Mutex
	count int
}

// check counts one detector call and delegates to the real Check.
func (f *fakeDetector) check(
	tok []rune, mode string, d detect.Data, tri detect.Trigrams, p detect.Params,
) detect.Verdict {
	f.mu.Lock()
	f.count++
	f.mu.Unlock()

	return detect.Check(tok, mode, d, tri, p)
}

// calls snapshots the counter.
func (f *fakeDetector) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.count
}

// TestAutoCorrect_TerminalVetoBeforeDetector pins the D-55 placement: a
// live terminal role 60 vetoes at the ARMING point — the detector is never
// consulted (spy: zero calls), exactly one terminal-veto abstention is
// counted, nothing is armed (the confirm never exists).
func TestAutoCorrect_TerminalVetoBeforeDetector(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.UseRole(&fakeRole{role: acRoleForbiddenTerminal})
	a.SetOptions(acOptions())
	detector := &fakeDetector{}
	a.UseDetector(detector.check)

	boundaryWord(t, a, wordEN)

	if got := detector.calls(); got != 0 {
		t.Fatalf("detector calls on a terminal class = %d, want 0 — the veto precedes detect.Check (D-55)", got)
	}
	assertSilence(t, a, sink, acReasonTerminalVeto)
	if got := sink.requireCount(); got != 0 {
		t.Errorf("requires = %d, want 0 — a vetoed boundary never arms a verify round", got)
	}
}

// acDefaultAnchorApp is the bridge-namespace identity of GNOME Terminal —
// the process name the default anchor's substring `gnome-terminal` must
// find (the substring match is the blocklist semantics, no re-anchoring).
const acDefaultAnchorApp = "gnome-terminal-server"

// TestAutoCorrect_TerminalVetoDefaultAnchor pins the default identity
// anchor of plan 10-07 (G-10-1, missing #2): a defaults document (no
// correction.terminals key — the shape on the host AND every stand case)
// carries the anchor [gnome-terminal, wezterm], and a focused identity
// matching it classifies the context terminal BEFORE the expensive sources
// — zero role queries (the live a11y round trip never runs), zero detector
// calls (the veto precedes detect.Check, D-55).
func TestAutoCorrect_TerminalVetoDefaultAnchor(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acDefaultAnchorApp})
	role := &fakeRole{role: acRoleAllowed} // 61 — not a terminal by role; the anchor must answer first
	a.UseRole(role)
	detector := &fakeDetector{}
	a.UseDetector(detector.check)
	src := &reloadSource{cfg: acEnabledCfg()} // the defaults document — the anchor rides the existing fold
	a.AttachConfig(src)

	boundaryWord(t, a, wordEN)

	if got := role.callCount(); got != 0 {
		t.Fatalf("role queries on an identity-anchored terminal = %d, want 0 — "+
			"the default anchor must classify before the expensive live source", got)
	}
	if got := detector.calls(); got != 0 {
		t.Fatalf("detector calls on an identity-anchored terminal = %d, want 0 — "+
			"the veto precedes detect.Check (D-55)", got)
	}
	assertSilence(t, a, sink, acReasonTerminalVeto)
}

// TestAutoCorrect_TerminalVetoPatternMatch pins the configured-identity
// source: an identity matching correction.terminals marks the context
// terminal even with a text-role answer (61) — substring RE2 semantics,
// the author anchors explicitly (the blocklist precedent).
func TestAutoCorrect_TerminalVetoPatternMatch(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.UseRole(&fakeRole{role: acRoleAllowed}) // 61 — a text box, not a terminal by role
	opts := acOptions()
	opts.Terminals = []string{`^org\.gnome\.gedit`}
	a.SetOptions(opts)

	boundaryWord(t, a, wordEN)

	assertSilence(t, a, sink, acReasonTerminalVeto)
}

// TestAutoCorrect_DefaultAnchorDoesNotInventIdentity pins the quiet norm
// of the anchor's source 2: an INVISIBLE identity (wezterm sits outside the
// a11y bus — no source, no focus event) answers nothing from the anchor —
// the anchor never invents a terminal class from silence; the GTK path
// (role query, detector, fire) runs exactly as before.
func TestAutoCorrect_DefaultAnchorDoesNotInventIdentity(t *testing.T) {
	a, _ := wiredActor() // no appid source installed — the identity is unobservable
	// The start fails honestly (the errAppidNoBus degradation): no observer,
	// the role seam keeps the fake — the identity stays invisible for both
	// the anchor and the confirm.
	a.UseAppidStarter(func() (session.AppidSource, error) { return nil, errAppidNoBus })
	detector := &fakeDetector{}
	a.UseDetector(detector.check)
	a.UseRole(&fakeRole{role: acRoleAllowedEntry}) // 79 — an unambiguous allowed role fires with an unknown identity
	src := &reloadSource{cfg: acEnabledCfg()}      // the anchor is in force, the identity is not
	a.AttachConfig(src)

	boundaryWord(t, a, wordEN)
	eventually(t, func() bool { return a.AutoCorrectCounters().Fired == 1 },
		"an invisible identity must not be classified terminal by the anchor — the round must fire")

	if got := detector.calls(); got != 1 {
		t.Errorf("detector calls on an invisible identity = %d, want exactly 1 — "+
			"the anchor must stay quiet when nothing is observable", got)
	}
}

// TestAutoCorrect_TerminalVetoIdentityUnobservable pins the fail-closed
// core of the invariant (D-55 verbatim: the veto «не зависит от
// наблюдаемости identity»): the identity is invisible (no a11y source, the
// role source errors) and the behavioral verify-dead class alone — learned
// on the manual round whose verify budget expired with the context alive —
// vetoes the next arming.
func TestAutoCorrect_TerminalVetoIdentityUnobservable(t *testing.T) {
	a, sink := wiredActor()
	a.UseAppidStarter(func() (session.AppidSource, error) { return nil, errAppidNoBus })
	a.UseRole(&fakeRole{err: errAppidBusDead})

	// The teaching round: the layer is OFF — the manual Double's verify
	// budget expires with the context alive, the class is learned.
	teachVerifyDead(t, a, wordEN)
	requiresAfterTeach := sink.requireCount()
	// The separator clears the taught word while the layer is still off —
	// the veto round must start from an honest single-word buffer.
	a.HandleKey(spaceKey())

	// The veto round: terminal class with ZERO observable identity.
	a.SetOptions(acOptions())
	boundaryWord(t, a, "cnhjrf")

	assertSilence(t, a, sink, acReasonTerminalVeto)
	if got := sink.requireCount(); got != requiresAfterTeach {
		t.Errorf("requires = %d, want %d — the vetoed round never arms regardless of the class source",
			got, requiresAfterTeach)
	}
}

// TestAutoCorrect_TerminalVetoGTKUnchanged pins the byte-as-today side
// (the veto adds one refusal, it moves nothing else): a GTK context —
// role 61, no terminal patterns, live class — runs the detector exactly as
// before the veto existed (one call) and the armed pipeline fires.
func TestAutoCorrect_TerminalVetoGTKUnchanged(t *testing.T) {
	a, _ := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	detector := &fakeDetector{}
	a.UseDetector(detector.check)
	a.UseRole(&fakeRole{role: acRoleAllowed})
	a.SetOptions(acOptions())

	boundaryWord(t, a, wordEN)
	eventually(t, func() bool { return a.AutoCorrectCounters().Fired == 1 },
		"the GTK round never fired — the veto must not touch the healthy path")

	if got := detector.calls(); got != 1 {
		t.Errorf("detector calls on a GTK class = %d, want exactly 1 — the same count as before the veto existed", got)
	}
}

// TestAutoCorrect_TerminalVetoSingleSlug pins the abstention accounting
// (T-10-03-04): one vetoed arming counts exactly one terminal-veto record
// — and because the vetoed boundary arms nothing, the confirm stage (whose
// role-60 branch is the second belt) never exists to duplicate it.
func TestAutoCorrect_TerminalVetoSingleSlug(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.UseAppid(fakeAppid{app: acListedApp})
	a.UseRole(&fakeRole{role: acRoleForbiddenTerminal}) // the second belt would say role-forbidden — if it ran
	a.SetOptions(acOptions())

	boundaryWord(t, a, wordEN)

	if got := sink.requireCount(); got != 0 {
		t.Fatalf("requires = %d, want 0 — nothing armed, the confirm goroutine never exists", got)
	}
	if got := strings.Count(buf.String(), `"reason":"`+acReasonTerminalVeto+`"`); got != 1 {
		t.Errorf("terminal-veto records = %d, want exactly 1; log:\n%s", got, buf.String())
	}
	assertSilence(t, a, sink, acReasonTerminalVeto)
}

// The terminal write-branch corpus of plan 10-04 (TERM-02, ROADMAP
// criterion 4): the terminal class routes to its OWN write branch at
// startRangeCorrection — before the caps fork, no pending fix, no
// RequireSurroundingText, no app verification (ADR-004 never answers
// here) — and the shared level-2 burst is CHUNKED: at most 8 BackSpace
// forwards per chunk, a context re-check between chunks, ONE commit at
// the end. The abort is free before the first BackSpace.

// termWord20 is the 20-rune convertible token of the chunk matrix
// (wordEN ×3 + "gh") and its wholesale EN→RU conversion — 20 backspaces
// the chunk ceiling splits 8+8+4.
const (
	termWord20          = wordEN + wordEN + wordEN + "gh"
	termWord20Converted = wordRU + wordRU + wordRU + "пр"
)

// termChunkKeys is the chunk ceiling of the terminal write branch — the
// production constant under test (the corpus mirrors it in termChunkOps).
const termChunkKeys = 8

// termChunkOps builds the exact unified op-log of a chunked terminal
// correction: chunks of at most 8 BackSpace forwards, then the single
// replacement commit.
func termChunkOps(total int, commit string) []string {
	ops := make([]string, 0, 2*total+1)
	for sent := 0; sent < total; {
		chunk := min(termChunkKeys, total-sent)
		for range chunk {
			ops = append(ops, backSpacePairOps()...)
		}
		sent += chunk
	}

	return append(ops, "commit("+commit+")")
}

// teachTerminalClass prepares a verify-dead context and leaves the buffer
// empty — the deterministic terminal-class setup the routing corpus drives
// (the teaching round's verify budget expires with the context alive).
func teachTerminalClass(t *testing.T, a *session.Actor) {
	t.Helper()
	teachVerifyDead(t, a, wordEN)
	a.HandleKey(spaceKey())
}

// TestTerminalBranch_RoutesChunkedNoPending pins the routing and the burst
// shape: a verify-dead context with the surrounding-text bit ANNOUNCED (the
// wezterm signature) skips the pending machine entirely — zero deletions,
// no pre-burst verification — and the 20-rune token replays as exactly 20
// BackSpace forwards in 8+8+4 chunks with ONE commit, the done INFO
// carrying the changed=true discriminator (D-24). The single require of the
// round is the verify-after ARM (Task 3) and lands strictly after the
// commit.
func TestTerminalBranch_RoutesChunkedNoPending(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor() // the surrounding-text bit is SET — the class routes, not the bit
	teachTerminalClass(t, a)
	requiresBefore := sink.requireCount()
	opsBefore := len(sink.opLog()) // the teaching round's own records stay out of the comparison

	typeWord(a, termWord20)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != requiresBefore+1 {
		t.Fatalf("requires = %d, want %d — exactly the verify-after arm, never the pending machine",
			got, requiresBefore+1)
	}
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Errorf("deletions = %+v, want 0 — the terminal replay is BackSpace-only", calls)
	}
	want := append(termChunkOps(len([]rune(termWord20)), termWord20Converted), "require")
	if got := sink.opLog()[opsBefore:]; !slices.Equal(got, want) {
		t.Fatalf("round op log = %q, want the 8+8+4 chunk shape + the verify-after arm %q", got, want)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done","changed":true`) {
		t.Errorf("the done INFO lost the changed=true discriminator (D-24); log:\n%s", buf.String())
	}
}

// TestTerminalBranch_VTERoleUnifiedChunks pins the owner verdict OQ 1
// «unified»: a role-60 context goes through the SAME terminal write branch
// — the working level-1 path with app verification is deliberately
// sacrificed, the unified channel is chunked level-2 for BOTH terminal
// classes.
func TestTerminalBranch_VTERoleUnifiedChunks(t *testing.T) {
	a, sink := wiredActor()
	a.UseRole(&fakeRole{role: acRoleForbiddenTerminal})

	typeWord(a, termWord20)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	// Task 3 semantics: the single require is the verify-after arm, after
	// the commit — the pending machine stays unarmed (unified routing).
	if got := sink.requireCount(); got != 1 {
		t.Fatalf("requires = %d, want 1 — unified routing: no pending machine, just the verify-after arm", got)
	}
	want := append(termChunkOps(len([]rune(termWord20)), termWord20Converted), "require")
	if got := sink.opLog(); !slices.Equal(got, want) {
		t.Fatalf("op log = %q, want the 8+8+4 chunk shape + the verify-after arm %q", got, want)
	}
}

// TestTerminalBranch_FocusLossBetweenChunksStops pins the criterion-4 stop:
// a FocusOut landing between chunks (the forward hook plays the real
// lifecycle callback, which serializes on the actor mutex and can only be
// observed in a chunk-boundary or delivery-wait window) is LOUD but no
// longer fatal (plan 10-17, the ADR-008 amendment (6)): after the first
// destructive key the burst completes, the commit follows on the CAPTURED
// emitter snapshot (the re-minted context stays byte-silent — the
// TestLevel1ConfirmFlush_FocusOutUsesCapturedSnapshot form), and the
// mid-flow delivery mark records the boundary. The commit's own
// BackSpace-only shape is pinned on the way.
func TestTerminalBranch_FocusLossAfterFirstForwardCommits(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	teachTerminalClass(t, a)
	fresh := &fakeSink{}

	typeWord(a, termWord20)
	var landed sync.Once
	sent := atomic.Int32{}
	sink.forwardHook = func(forwardCall) {
		// 2×termChunkKeys emissions = the FIRST chunk's pairs complete
		// (round-6) — the mid-flow moment after the first destructive key.
		if sent.Add(1) == 2*termChunkKeys {
			landed.Do(func() {
				go func() {
					a.HandleLifecycle(engine.LifecycleFocusOut) // the generation moves mid-flow
					a.AttachEngine(fresh)                       // the context re-mints onto a new object
				}()
			})
		}
	}
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	forwards := sink.forwardCalls()
	if len(forwards) != 2*len([]rune(termWord20)) {
		t.Fatalf("forwards = %d, want %d — a mid-flow generation change must not stop the burst",
			len(forwards), 2*len([]rune(termWord20)))
	}
	assertBackSpacePairs(t, forwards, len([]rune(termWord20)), "the commit's own replay")
	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != termWord20Converted {
		t.Fatalf("commits = %q, want exactly one %q on the captured snapshot — the deletion already "+
			"on the wire must always be followed by its replacement", texts, termWord20Converted)
	}
	if got := fresh.commitTexts(); len(got) != 0 {
		t.Errorf("commits into the RE-MINTED engine = %q, want none — the commit rides the snapshot "+
			"captured at the round's start", got)
	}
	logged := buf.String()
	if !strings.Contains(logged, `"msg":"correction delivery","stage":"context-changed-midflow"`) {
		t.Errorf("the loud mid-flow boundary is missing; log:\n%s", logged)
	}
	if !strings.Contains(logged, `"msg":"correction","outcome":"done","changed":true`) {
		t.Errorf("the round never logged its completion; log:\n%s", logged)
	}
}

// TestTerminalBranch_FreeAbortBeforeBurst pins the free abort: every
// refusal that can happen before the first BackSpace — empty-buffer,
// convert-failed, no-engine, the D-27 cap — leaves the field untouched:
// zero forwards, the skipCorrection branches unchanged.
func TestTerminalBranch_FreeAbortBeforeBurst(t *testing.T) {
	t.Run("empty-buffer", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()

		tapShift(a) // no word — the Double finds an empty buffer
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)

		if got := len(sink.forwardCalls()); got != 0 {
			t.Errorf("forwards on an empty buffer = %d, want 0", got)
		}
		if !strings.Contains(buf.String(), `"reason":"empty-buffer"`) {
			t.Errorf("empty-buffer skip record missing; log:\n%s", buf.String())
		}
	})

	t.Run("convert-failed", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()

		// The phrase path carries the unmappable letter: «é» has no key
		// position, so the word stretch stops before it — but the Triple's
		// phrase is the whole buffer, and ConvertRuns refuses it wholesale
		// (T-03-01-03: no partial conversion ever leaks).
		typeWord(a, wordEN+"é")
		tapShift(a)
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)

		if got := len(sink.forwardCalls()); got != 0 {
			t.Errorf("forwards on an unmappable phrase = %d, want 0", got)
		}
		if !strings.Contains(buf.String(), `"reason":"convert-failed"`) {
			t.Errorf("convert-failed skip record missing; log:\n%s", buf.String())
		}
	})

	t.Run("no-engine", func(t *testing.T) {
		runNoEngineCell(t)
	})

	t.Run("backspace-cap", func(t *testing.T) {
		runBackspaceCapCell(t)
	})
}

// runNoEngineCell drives the no-engine free-abort cell (the subtest body —
// factored out to keep the table function under the complexity ceiling).
func runNoEngineCell(t *testing.T) {
	t.Helper()
	buf := captureLogs(t)
	a := session.NewActor(farWindow)
	a.HandleCapabilities(engine.CapSurroundingText) // no engine minted

	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if !strings.Contains(buf.String(), `"reason":"no-engine"`) {
		t.Errorf("no-engine skip record missing; log:\n%s", buf.String())
	}
}

// runBackspaceCapCell drives the D-27 free-abort cell of the terminal
// branch: the cap refusal lives inside executeLevel2 — the new code path —
// so it is proven against a LEARNED terminal class.
func runBackspaceCapCell(t *testing.T) {
	t.Helper()
	buf := captureLogs(t)
	a, sink := wiredActor()
	teachTerminalClass(t, a)
	a.SetOptions(session.Options{BackspaceCap: 5}) // the 6-rune word exceeds the cap

	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := len(sink.forwardCalls()); got != 0 {
		t.Errorf("forwards on a capped replay = %d, want 0 — D-27 refuses before the burst", got)
	}
	if !strings.Contains(buf.String(), `"reason":"backspace-cap"`) {
		t.Errorf("backspace-cap skip record missing; log:\n%s", buf.String())
	}
}

// TestTerminalBranch_LegacyNoCapsChunkedToo pins the shared write path: a
// legacy client WITHOUT the surrounding-text bit keeps its immediate
// no-verification level-2 correction (the ADR-004 degradation, byte as
// today) — and the chunking applies to it too, the same executeLevel2.
func TestTerminalBranch_LegacyNoCapsChunkedToo(t *testing.T) {
	a, sink := wiredActorCaps(engine.CapPreeditText) // no CapSurroundingText, no terminal class
	legacyWord := strings.Repeat(wordEN, 2)          // 12 runes — two chunks

	typeWord(a, legacyWord)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != 0 {
		t.Errorf("requires = %d, want 0 — the legacy path stays verification-free", got)
	}
	want := termChunkOps(len([]rune(legacyWord)), strings.Repeat(wordRU, 2))
	if got := sink.opLog(); !slices.Equal(got, want) {
		t.Fatalf("op log = %q, want the chunked legacy burst %q", got, want)
	}
}

// TestTerminalBranch_TailRidesReplay pins the D-13 invariant on the
// terminal replay: the table-less tail (the space) is deleted with the
// token and recommitted AS TYPED — BuildPlan untouched, «ghbdtn » still
// becomes «привет » through the chunked branch.
func TestTerminalBranch_TailRidesReplay(t *testing.T) {
	a, sink := wiredActor()
	teachTerminalClass(t, a)
	opsBefore := len(sink.opLog()) // the teaching round's own records stay out

	typeWord(a, wordEN+" ")
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	want := append(termChunkOps(len(wordEN)+1, wordRU+" "), "require") // the verify-after arm (Task 3)
	if got := sink.opLog()[opsBefore:]; !slices.Equal(got, want) {
		t.Fatalf("round op log = %q, want the token+tail replay %q", got, want)
	}
}

// The terminal verify-after corpus of plan 10-04 Task 3 (ADR-008 D4, owner
// verdict OQ 2 «contract»): the terminal branch verifies its correction
// after the fact on the SAME epoch/debounce/quiet-timeout machine — but the
// clipboard rung is forbidden BY CONSTRUCTION, never by config: the round
// carries no paste payload, so a mismatch can only log and stop. The
// mid-line semantics stop at the D7 contract boundary: buffer trust holds
// only for the type-at-cursor flow (the poison/primary mechanisms are
// deliberately absent).

// TestTerminalBranch_VerifyAfterNoRungOnMismatch pins the construction ban:
// a terminal correction whose verify-after MISMATCHES never produces a
// Ctrl+V forward and never spawns a clipboard subprocess — even with
// clipboard_rung: true and a healthy client installed. The WR-05 debounce
// keeps its shape (the first mismatching push re-requires), the second
// concludes into the INFO mismatch and a full stop.
func TestTerminalBranch_VerifyAfterNoRungOnMismatch(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.UseRole(&fakeRole{role: acRoleForbiddenTerminal}) // unified routing: terminal branch
	a.SetOptions(session.Options{ClipboardRung: true})  // the config MUST NOT matter
	runner := &clipRunner{reply: []byte("sentinel")}
	a.UseClipboard(clipboard.New(clipboard.WithRunner(runner.run)))

	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != 1 {
		t.Fatalf("requires after the terminal correction = %d, want 1 — the verify-after round armed", got)
	}

	// WR-05: the first mismatching answer re-requires, the second concludes.
	mismatch := "abc другойтекст"
	a.HandleSurroundingText(mismatch, runeLen(mismatch), runeLen(mismatch))
	if got := sink.requireCount(); got != 2 {
		t.Fatalf("requires after the first stale answer = %d, want 2 — the debounce re-required", got)
	}
	a.HandleSurroundingText(mismatch, runeLen(mismatch), runeLen(mismatch))

	assertBackSpacePairs(t, sink.forwardCalls(), len([]rune(wordEN)),
		"a terminal mismatch must never forward a paste chord (ADR-008 D4) — BackSpace pairs only")
	if seq := runner.sequence(); len(seq) != 0 {
		t.Errorf("clipboard subprocesses = %q, want none — the rung is forbidden by construction", seq)
	}
	if !strings.Contains(buf.String(), `"msg":"correction verify","outcome":"mismatch"`) {
		t.Errorf("the mismatch INFO record is missing; log:\n%s", buf.String())
	}
}

// TestTerminalBranch_VerifyDeadQuietTimeout pins the quiet timeout of a
// verify-dead context: the pushes never come, the round dies by budget —
// the timeout record, no paste load, no repeated requires, and the stale
// teaching deadline stays a no-op (epoch discipline).
func TestTerminalBranch_VerifyDeadQuietTimeout(t *testing.T) {
	buf := captureLogsLevel(t, slog.LevelDebug)
	a, sink := wiredActor()
	cfg := reloadCfg(300, "shift+ctrl_r")
	cfg.Timeouts.VerifyWaitMs = 30
	a.AttachConfig(&reloadSource{cfg: cfg})

	teachTerminalClass(t, a)
	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != 2 {
		t.Fatalf("requires = %d, want 2 — the teaching round's (1) plus the verify-after arm (1), nothing else", got)
	}
	if !waitLogUntil(t, buf, `"outcome":"timeout"`, 1, 2*time.Second) {
		t.Fatalf("the quiet timeout never landed; log:\n%s", buf.String())
	}
	if got := strings.Count(buf.String(), `"outcome":"mismatch"`); got != 0 {
		t.Errorf("mismatch records on a quiet context = %d, want 0", got)
	}
	if got := sink.requireCount(); got != 2 {
		t.Errorf("requires after the timeout = %d, want 2 — no repeats", got)
	}
	assertBackSpacePairs(t, sink.forwardCalls(), len([]rune(wordEN)),
		"no paste load on the timeout path — BackSpace pairs only")
}

// TestTerminalBranch_VerifyAfterQuietConfirm pins the honest-client side: a
// fresh push whose suffix is the converted token (+tail) settles the round
// quietly — a DEBUG match record, no mismatch, no further emitter traffic.
func TestTerminalBranch_VerifyAfterQuietConfirm(t *testing.T) {
	buf := captureLogsLevel(t, slog.LevelDebug)
	a, sink := wiredActor()
	a.UseRole(&fakeRole{role: acRoleForbiddenTerminal})

	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	line := "abc " + wordRU
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))

	if !strings.Contains(buf.String(), `"msg":"correction verify","outcome":"match"`) {
		t.Errorf("the quiet match record is missing; log:\n%s", buf.String())
	}
	if got := strings.Count(buf.String(), `"outcome":"mismatch"`); got != 0 {
		t.Errorf("mismatch records on a matching push = %d, want 0", got)
	}
	if got := sink.requireCount(); got != 1 {
		t.Errorf("requires = %d, want 1 — the match settles without a re-require", got)
	}
	if got := len(sink.forwardCalls()); got != 2*len(wordEN) {
		t.Errorf("forwards = %d, want %d — the confirmation emits nothing", got, 2*len(wordEN))
	}
}

// TestTerminalBranch_Level2PacesWire pins the paced wire discipline of the
// write branch (the 2026-10-12 VTE live finding, plan 10-05): a tight
// ForwardKeyEvent burst REORDERS against the CommitText inside the client's
// IM queue — the client drains forwarded keys lazily (they queue as GDK
// events) while the commit lands synchronously, so the live row turned
// "ghbdtn" into "ghприв" (four erases, the commit, two late erases). The
// branch must sleep its pacing quantum BEFORE EVERY emission — round-6:
// every BackSpace is a press+release PAIR, so the merged wire timeline is
// sleep→press→sleep→release→…→sleep→commit through the actor's sleeper
// seam (the UseLevel2Sleeper form); the release's own pacing tick keeps it
// one gap (20 ms default) behind its press — far inside the 500 ms window
// in which a client's armed key repeat would fire.
func TestTerminalBranch_Level2PacesWire(t *testing.T) {
	a, sink := wiredActorCaps(engine.CapPreeditText) // no CapSurroundingText — the write branch

	const gap = 20 * time.Millisecond
	var wire []string
	record := func(op string) {
		wire = append(wire, op)
	}
	a.UseLevel2Sleeper(func(d time.Duration) {
		record(fmt.Sprintf("sleep(%s)", d))
	})
	sink.forwardHook = func(c forwardCall) {
		if c.state&engine.MaskRelease != 0 {
			record(backSpaceReleaseOp())

			return
		}
		record(backSpaceOp())
	}
	sink.commitHook = func(text string) { record("commit(" + text + ")") }

	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	want := make([]string, 0, 4*len([]rune(wordEN))+2)
	for range len([]rune(wordEN)) {
		want = append(want, fmt.Sprintf("sleep(%s)", gap), backSpaceOp(),
			fmt.Sprintf("sleep(%s)", gap), backSpaceReleaseOp())
	}
	want = append(want, fmt.Sprintf("sleep(%s)", gap), "commit("+wordRU+")")
	if !slices.Equal(wire, want) {
		t.Fatalf("wire timeline = %q, want the paced %q", wire, want)
	}
}

// The deferred-flip corpus of plan 10-08 (G-10-1 cause B, the E3 verdict of
// 2026-10-13 18:33 MSK): the terminal branch's post-correction flip DEFERS
// until the commit's delivery is confirmed — a confirming push, the next
// keystroke, or the grace deadline — because the flip's SetGlobalEngine
// re-mints the engine object 2–8 ms after the commit and, under host focus
// churn, the late-processing client discards the stale engine's signal: the
// committed word is LOST (the E3 push 16 ms after the commit showed an EMPTY
// field). The flip decision reuses settleCorrectionFlip's exact guards (the
// combo gesture and the off gate keep their byte-as-today paths), each
// release site fires flipTo exactly once, and no release site sits on the
// FocusOut lifecycle — a flap pair must not be able to reproduce E3.

// termFlipRUOp is the released flip's panel-symbol op — the deferred flip's
// observable in the op log (the corpus names it once, goconst).
const termFlipRUOp = "mode(ru)"

// requireOp is the RequireSurroundingText op-log entry (goconst).
const requireOp = "require"

// flipClock is the deterministic double of the deferred-flip grace seam
// (plan 10-08, the flapClock form): arm records the armed deadlines and
// holds the single pending callback; fire() plays the grace expiry, and the
// stop closure the actor holds cancels a released flip's pending callback —
// a stale fire (already released, re-armed slot) is the production no-op.
type flipClock struct {
	mu        sync.Mutex
	durations []time.Duration
	pending   func()
}

// arm records one armed grace deadline and returns the actor's stop closure.
func (f *flipClock) arm(d time.Duration, fn func()) func() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.durations = append(f.durations, d)
	f.pending = fn

	return func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.pending = nil
	}
}

// fire plays the grace expiry: a released flip holds no callback, so a
// stale fire is a no-op — exactly the production guard.
func (f *flipClock) fire() {
	f.mu.Lock()
	fn := f.pending
	f.pending = nil
	f.mu.Unlock()
	if fn != nil {
		fn()
	}
}

// armedDurations snapshots the armed grace deadlines (armed count + the
// pinned constant in one read).
func (f *flipClock) armedDurations() []time.Duration {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]time.Duration(nil), f.durations...)
}

// hasPending reports whether a grace deadline is still armed — the quiet-
// timeout cell's proof that the flip fired from the timeout release, not
// from the grace.
func (f *flipClock) hasPending() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.pending != nil
}

// deferFlipActor wires the corpus actor: the terminal branch via the role
// double (the TestTerminalBranch_VTERoleUnifiedChunks form — a fresh buffer,
// no teaching leftovers), the flip gate ON, the grace under the corpus
// clock.
func deferFlipActor() (*session.Actor, *fakeSink, *flipClock) {
	a, sink := wiredActor()
	a.UseRole(&fakeRole{role: acRoleForbiddenTerminal})
	a.SetOptions(flipOnOptions())
	clock := &flipClock{}
	a.UseFlipTimer(clock.arm)

	return a, sink, clock
}

// deferFlipRound drives one terminal word correction: type wordEN, double
// tap, expiry — the branch commits and DEFERS its flip.
func deferFlipRound(a *session.Actor) {
	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
}

// TestTerminalBranch_FlipDeferredUntilConfirmed pins the E3 fix's core: a
// flip-enabled terminal correction emits burst+commit+verify-arm and NOTHING
// else — no SetGlobalEngine, no panel emit between the commit and the
// delivery confirmation; the confirming push (the field text ends with
// converted+tail — the SAME match criterion the verify-after reads) fires
// the flip exactly once, in the D-36 order (mode record → switcher → panel
// emit), and the release itself emits no field traffic beyond the flip's
// panel emit.
func TestTerminalBranch_FlipDeferredUntilConfirmed(t *testing.T) {
	buf := captureLogs(t)
	a, sink, _ := deferFlipActor()
	probe := &switchProbe{}
	var swSawMode bool
	probe.onCall = func(string) {
		swSawMode = strings.Contains(buf.String(), `"msg":"mode","to":"ru"`)
	}
	a.SetSwitcher(probe.switcher)

	deferFlipRound(a)

	want := append(termChunkOps(len([]rune(wordEN)), wordRU), "require")
	if got := sink.opLog(); !slices.Equal(got, want) {
		t.Fatalf("round op log = %q, want the burst+commit+verify-arm shape %q — no switch/panel "+
			"traffic may sit between the commit and the delivery confirmation (E3, plan 10-08)", got, want)
	}
	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("mode records before the confirming push = %d, want 0 — the flip defers; log:\n%s",
			got, buf.String())
	}
	if got := probe.targets(); len(got) != 0 {
		t.Errorf("SetGlobalEngine calls before the confirmation = %q, want none — the flip must not "+
			"orphan the in-flight commit", got)
	}

	line := "abc " + wordRU
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))

	logged := buf.String()
	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("mode records after the confirming push = %d, want exactly 1; log:\n%s", got, logged)
	}
	if !strings.Contains(logged, `"msg":"mode","to":"ru"`) {
		t.Errorf("the released flip is not the EN→RU set; log:\n%s", logged)
	}
	if got := probe.targets(); !slices.Equal(got, []string{engine.NameRU}) {
		t.Errorf("switch targets after the confirmation = %q, want exactly [%s] — one flip, target engine",
			got, engine.NameRU)
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru"}) {
		t.Errorf("panel symbols = %q, want exactly [ru] — one emit per released flip", got)
	}
	if !swSawMode {
		t.Errorf("the switcher fired before the mode record — the D-36 order survives the deferral")
	}
	if got := sink.requireCount(); got != 1 {
		t.Errorf("requires = %d, want 1 — the release requires nothing", got)
	}
	if got := len(sink.forwardCalls()); got != 2*len([]rune(wordEN)) {
		t.Errorf("forwards = %d, want %d — the confirmation replayed nothing", got, 2*len([]rune(wordEN)))
	}
}

// TestTerminalBranch_FlipReleasedByNextKey pins the push-less client's
// confirmation (the wezterm class): the first keystroke after the correction
// releases the flip STRICTLY before the key is processed — the key already
// types in the RESULT mode (the owner rule «режим следует за скриптом
// результата» is preserved by the deferral, not lost to it).
func TestTerminalBranch_FlipReleasedByNextKey(t *testing.T) {
	buf := captureLogs(t)
	a, sink, _ := deferFlipActor()

	deferFlipRound(a)

	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("mode records before the next key = %d, want 0 — the flip waits; log:\n%s",
			got, buf.String())
	}

	a.HandleKey(engine.EngineEvent{Keyval: 'x'}) // EN-mode keyval — released first, consumed in RU

	ops := sink.opLog()
	n := len(ops)
	if n < 2 || ops[n-2] != termFlipRUOp || ops[n-1] != "commit(ч)" {
		t.Fatalf("tail of the op log = %q, want [... mode(ru) commit(ч)] — the release must precede "+
			"the key's processing, the key types in the result mode", ops[max(n-3, 0):])
	}
	if got := countModeRecords(buf); got != 1 {
		t.Errorf("mode records after the key = %d, want exactly 1 — one flip per correction; log:\n%s",
			got, buf.String())
	}
}

// TestTerminalBranch_FlipArmsNoGrace pins the narrowed release machine
// (plan 10-17): the terminal flip arms NO grace deadline — SetGlobalEngine
// must be impossible between the commit and the flip's honest release
// cause — and a grace fire (a stale callback from a level-1 machine that
// shared the slot form) is a no-op. The mode rule is never lost: it
// releases exactly once on the NEXT keystroke, loudly marked.
func TestTerminalBranch_FlipArmsNoGrace(t *testing.T) {
	buf := captureLogs(t)
	a, _, clock := deferFlipActor()

	deferFlipRound(a)

	if armed := clock.armedDurations(); len(armed) != 0 {
		t.Fatalf("the terminal flip armed %d grace deadlines, want 0 — a grace release inside the "+
			"delivery window is the round-2 loss geometry (the +104 ms flip)", len(armed))
	}
	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("mode records with the flip pending = %d, want 0; log:\n%s", got, buf.String())
	}

	clock.fire() // stale — nothing was armed, the slot must not move

	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("a grace fire released the flip: %d mode records, want 0 — the deadline is not a "+
			"release cause anymore; log:\n%s", got, buf.String())
	}

	a.HandleKey(engine.EngineEvent{Keyval: 'x'}) // the next keystroke — the honest cause

	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("mode records after the next keystroke = %d, want exactly 1 — the rule waits for "+
			"its cause, then fires once; log:\n%s", got, buf.String())
	}
	if !strings.Contains(buf.String(), `"msg":"correction delivery","stage":"flip-released","reason":"keystroke"`) {
		t.Errorf("the released flip carries no delivery mark; log:\n%s", buf.String())
	}
}

// TestTerminalBranch_FlipImmuneToVerifyMismatch pins the mismatch half of
// the narrowed release machine (plan 10-17): the concluding mismatch push
// does NOT release the terminal deferred flip — a SetGlobalEngine riding a
// verify settlement is exactly the round-2 loss geometry — while the 10-04
// mismatch semantics stay byte-unchanged (log and stop, no rung, no
// repair). The mode rule is not lost: the next keystroke releases it.
func TestTerminalBranch_FlipImmuneToVerifyMismatch(t *testing.T) {
	buf := captureLogs(t)
	a, _, _ := deferFlipActor()

	deferFlipRound(a)

	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("mode records before the mismatch push = %d, want 0; log:\n%s", got, buf.String())
	}

	// WR-05: the first mismatching answer re-quires — a stale-retry push
	// is NOT a settlement and must not release the flip early.
	a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))
	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("a stale-retry push released the flip early: %d mode records, want 0; log:\n%s",
			got, buf.String())
	}
	a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))

	logged := buf.String()
	if !strings.Contains(logged, `"msg":"correction verify","outcome":"mismatch"`) {
		t.Errorf("the 10-04 mismatch INFO record is missing (byte-unchanged semantics); log:\n%s", logged)
	}
	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("the concluding mismatch released the flip: %d mode records, want 0 — the verify "+
			"outcome is not a release cause (plan 10-17); log:\n%s", got, logged)
	}

	a.HandleKey(engine.EngineEvent{Keyval: 'x'}) // the honest cause still releases

	if got := countModeRecords(buf); got != 1 {
		t.Errorf("mode records after the next keystroke = %d, want exactly 1 — the rule waits, "+
			"then fires once; log:\n%s", got, buf.String())
	}
}

// TestTerminalBranch_FlipImmuneToVerifyTimeout pins the timeout half of
// the narrowed release machine (plan 10-17): the quiet verify-after
// timeout does NOT release the terminal deferred flip, no grace exists to
// paper over it, and the next keystroke releases the rule exactly once.
func TestTerminalBranch_FlipImmuneToVerifyTimeout(t *testing.T) {
	buf := captureLogsLevel(t, slog.LevelDebug)
	a, _, clock := deferFlipActor()
	cfg := reloadCfg(300, "shift+ctrl_r")
	cfg.Timeouts.VerifyWaitMs = 30
	a.AttachConfig(&reloadSource{cfg: cfg})

	deferFlipRound(a)

	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("mode records before the quiet timeout = %d, want 0; log:\n%s", got, buf.String())
	}

	if !waitLogUntil(t, buf, `"outcome":"timeout"`, 1, 2*time.Second) {
		t.Fatalf("the quiet timeout never landed; log:\n%s", buf.String())
	}
	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("the quiet timeout released the flip: %d mode records, want 0 — the timeout is "+
			"not a release cause (plan 10-17); log:\n%s", got, buf.String())
	}
	clock.fire() // stale — no grace was armed either

	if got := countModeRecords(buf); got != 0 {
		t.Errorf("a grace fire released the flip after the timeout: %d mode records, want 0", got)
	}

	a.HandleKey(engine.EngineEvent{Keyval: 'x'})

	if got := countModeRecords(buf); got != 1 {
		t.Errorf("mode records after the next keystroke = %d, want exactly 1; log:\n%s",
			got, buf.String())
	}
}

// TestTerminalBranch_FlipFalseArmsNothing pins the frozen-stand regression
// guard: with the off gate (the zero value every stand config document and
// the whole unit corpus carries) the machine never arms — the round's op log
// stays byte-identical to the frozen shape.
func TestTerminalBranch_FlipFalseArmsNothing(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.UseRole(&fakeRole{role: acRoleForbiddenTerminal})
	clock := &flipClock{}
	a.UseFlipTimer(clock.arm)
	a.SetOptions(session.Options{FlipAfterCorrection: false})

	deferFlipRound(a)

	if armed := clock.armedDurations(); len(armed) != 0 {
		t.Fatalf("flip_after_correction=false armed %d deferred flips, want 0 — the frozen stand "+
			"rows' world must stay byte-identical", len(armed))
	}
	if got := countModeRecords(buf); got != 0 {
		t.Errorf("mode records with the gate off = %d, want 0; log:\n%s", got, buf.String())
	}
	want := append(termChunkOps(len([]rune(wordEN)), wordRU), "require")
	if got := sink.opLog(); !slices.Equal(got, want) {
		t.Fatalf("op log = %q, want the byte-identical frozen shape %q", got, want)
	}
}

// TestTerminalBranch_ComboFlipImmediateNotDeferred pins the combo path:
// a combo gesture's terminal correction keeps its OWN settleCombo flip —
// immediate, exactly one, the deferred machine never arms (the D-36 combo
// semantics byte-as-today).
func TestTerminalBranch_ComboFlipImmediateNotDeferred(t *testing.T) {
	buf := captureLogs(t)
	a, _, clock := deferFlipActor()

	typeWord(a, wordEN)
	pressComboDefault(a) // the word half runs the terminal branch under comboPending

	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("combo mode records = %d, want exactly 1 fired immediately — the combo's own "+
			"settleCombo flip, byte as today; log:\n%s", got, buf.String())
	}
	if armed := clock.armedDurations(); len(armed) != 0 {
		t.Errorf("deferred flips armed on the combo path = %d, want 0 — the combo flip stays immediate",
			len(armed))
	}
	logged := buf.String()
	iDone := strings.Index(logged, `"msg":"correction","outcome":"done"`)
	iMode := strings.Index(logged, `"msg":"mode"`)
	if iDone > iMode {
		t.Errorf("combo flip preceded the settled done (done@%d > mode@%d) — the D-36 order", iDone, iMode)
	}
}

// TestTerminalBranch_FlipSurvivesFocusOutGraceReleases pins the lifecycle
// neutrality: the deferred flip is the GLOBAL mode rule, not context state —
// a FocusOut/FocusIn flap pair (the host churn) must not release it (a
// SetGlobalEngine inside the churn window is exactly the E3 reproduction),
// and no grace exists to paper over it (plan 10-17): the next keystroke
// after the pair releases the flip exactly once.
func TestTerminalBranch_FlipSurvivesFocusOutGraceReleases(t *testing.T) {
	buf := captureLogs(t)
	a, _, clock := deferFlipActor()

	deferFlipRound(a)

	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("mode records before the flap pair = %d, want 0 — the flip waits; log:\n%s",
			got, buf.String())
	}

	a.HandleLifecycle(engine.LifecycleFocusOut)
	a.HandleLifecycle(engine.LifecycleFocusIn)

	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("the flap pair released the flip: %d mode records, want 0 — no release site may sit "+
			"on the lifecycle (E3 must stay unreproducible); log:\n%s", got, buf.String())
	}
	if clock.hasPending() {
		t.Error("the terminal flip armed a grace deadline — the narrowed machine must not (plan 10-17)")
	}

	a.HandleKey(engine.EngineEvent{Keyval: 'x'}) // the honest cause survives the pair

	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("mode records after the keystroke = %d, want exactly 1 — the rule survives the "+
			"flap pair; log:\n%s", got, buf.String())
	}
}

// TestTerminalBranch_SecondCorrectionReleasesPendingFlip pins the exactly-
// one-flip-per-correction handoff: the second correction releases the first
// round's pending flip synchronously (strictly before its own first wire
// emission), then defers its own — and the confirming push releases that one
// too. No flip is lost, none doubles.
func TestTerminalBranch_SecondCorrectionReleasesPendingFlip(t *testing.T) {
	buf := captureLogs(t)
	a, sink, _ := deferFlipActor()

	deferFlipRound(a)

	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("mode records after round 1 = %d, want 0 — the flip waits; log:\n%s", got, buf.String())
	}

	opsAfterRound1 := len(sink.opLog())
	tapShift(a) // the first keystroke of the new gesture releases round 1's flip
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	ops := sink.opLog()
	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("mode records after round 2 = %d, want exactly 1 — round 1's flip released before "+
			"round 2 started, round 2's own flip still deferred; log:\n%s", got, buf.String())
	}
	// The release flip sits strictly before round 2's burst; the round-1
	// buffer was replaced with wordRU, so round 2 converts it back to wordEN
	// — the opposite direction, the same chunked shape.
	want := append([]string{termFlipRUOp}, termChunkOps(len([]rune(wordEN)), wordEN)...)
	want = append(want, "require")
	if !slices.Equal(ops[opsAfterRound1:], want) {
		t.Fatalf("round-2 op log = %q, want the release flip %q seated strictly before round 2's "+
			"chunked burst %q", ops[opsAfterRound1:], []string{termFlipRUOp},
			append(termChunkOps(len([]rune(wordEN)), wordEN), "require"))
	}

	line := "abc " + wordEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))

	if got := countModeRecords(buf); got != 2 {
		t.Fatalf("mode records after round 2's confirmation = %d, want exactly 2 — one flip per "+
			"correction, none lost; log:\n%s", got, buf.String())
	}
	if !strings.Contains(buf.String(), `"msg":"mode","to":"en"`) {
		t.Errorf("round 2's flip is not the RU→EN set; log:\n%s", buf.String())
	}
}

// The stale-cache amendment corpus (plan 10-08, the owner-approved Rule 4
// amendment of 2026-10-13, G-10-1): the stand's wezterm-nightly context
// drifted from the pinned 0x9 signature (no surrounding bit, no pushes) to
// the 0x29 class — the bit IS announced and the client ANSWERS
// RequireSurroundingText with its stale cache (never the typed stretch).
// The old pending machine concluded such rounds as GTK verify-mismatches,
// the behavioral class never learned, and the context never corrected. The
// owner's verdict: a client that announced the bit and answers with a cache
// that does not hold what the daemon watched being typed IS the terminal
// context — the class learns and the round routes to the terminal write
// branch. The fail-closed invariant survives for contexts WITH cache
// evidence: a client whose pushes tracked the typing (the honest GTK world)
// keeps the byte-as-today verify-mismatch abort.

// TestTerminalBranch_StaleCacheAnswerRoutesTerminal pins the amendment's
// core: a nil-cache word round (the client never pushed a field state —
// the pending armed over the empty cache and required) whose only answer
// does not hold the typed stretch LEARNS the behavioral class and executes
// the terminal write branch ON THIS ROUND — done+level-2, no GTK
// verify-mismatch, no repair-free abort of a correctable word.
func TestTerminalBranch_StaleCacheAnswerRoutesTerminal(t *testing.T) {
	buf := captureLogsLevel(t, slog.LevelDebug)
	a, sink := wiredActor() // caps announced; NO role source — unclassified context

	typeWord(a, wordEN) // transit; the client never pushed — nil cache
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow) // pending armed over the nil cache; the Require is out

	if got := sink.opLog(); !slices.Equal(got, []string{requireOp}) {
		t.Fatalf("op log before the answer = %q, want exactly [%s] — nothing may precede the verdict", got, requireOp)
	}

	a.HandleSurroundingText("x", 1, 1) // the stale cache: not the typed stretch

	logged := buf.String()
	if !strings.Contains(logged, `"msg":"terminal class","verify_dead":true`) {
		t.Errorf("the stale-cache signature did not learn the behavioral class; log:\n%s", logged)
	}
	if !strings.Contains(logged, `"msg":"correction","outcome":"done","changed":true`) {
		t.Fatalf("the stale-routed round never completed — the word was abandoned; log:\n%s", logged)
	}
	if !strings.Contains(logged, `"msg":"correction","level":2`) {
		t.Errorf("the stale-routed round missed the terminal write branch (level-2 record); log:\n%s", logged)
	}
	if strings.Contains(logged, `"reason":"verify-mismatch"`) {
		t.Errorf("the stale-cache answer concluded a GTK mismatch — the amended signature routes terminal\n%s", logged)
	}
	burst := burstOps(len([]rune(wordEN)), wordRU)
	want := append([]string{requireOp}, burst...)
	want = append(want, requireOp)
	if got := sink.opLog(); !slices.Equal(got, want) {
		t.Fatalf("op log = %q, want the arm + the chunked terminal write %q + the verify-after arm", got, burst)
	}
}

// TestTerminalBranch_StaleCacheLearnsForNextGesture pins the class
// persistence: after the stale-cache round, the NEXT gesture routes through
// the terminal write branch AT THE DECISION — no pending machine, no
// Require before the burst, the same shape the verify-timeout-taught class
// produces.
func TestTerminalBranch_StaleCacheLearnsForNextGesture(t *testing.T) {
	a, sink := wiredActor()

	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText("x", 1, 1) // the stale-cache answer: learn + write

	opsBefore := len(sink.opLog())
	a.HandleKey(spaceKey()) // the boundary-free separator; the buffer keeps mirroring
	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	want := append(burstOps(len([]rune(wordEN)), wordRU), requireOp)
	if got := sink.opLog()[opsBefore:]; !slices.Equal(got, want) {
		t.Fatalf("round-2 op log = %q, want the decision-time terminal burst + the verify-after arm %q — "+
			"the learned class must route at the decision, never back into the pending machine", got, want)
	}
}

// TestTerminalBranch_HonestPushStillFailClosed pins the invariant half of
// the amendment: a context WITH cache evidence — the pushes tracked the
// typing (the honest GTK world) — whose fresh answer does not hold the
// token keeps the byte-as-today fail-closed verify-mismatch: no learning,
// no write, not one destructive emission.
func TestTerminalBranch_HonestPushStillFailClosed(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	typeWord(a, wordEN)
	// The honest tracking push: the cache carries the client's field state.
	a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow) // pending armed over the NON-nil cache
	// The fresh honest answer — still not the typed token.
	a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))

	logged := buf.String()
	if !strings.Contains(logged, `"reason":"verify-mismatch"`) {
		t.Errorf("the honest-GTK mismatch lost its fail-closed record; log:\n%s", logged)
	}
	if strings.Contains(logged, `"msg":"terminal class"`) {
		t.Errorf("a tracking client learned the terminal class — cache evidence keeps fail-closed\n%s", logged)
	}
	if got := len(sink.forwardCalls()); got != 0 {
		t.Errorf("the honest mismatch forwarded %+v — abort, не мусорить", sink.forwardCalls())
	}
	if got := len(sink.commitTexts()); got != 0 {
		t.Errorf("the honest mismatch committed %q — abort, не мусорить", sink.commitTexts())
	}
}

// TestTerminalBranch_StaleRouteFlipStillDeferred pins the Task-1 E3
// contract on the amendment path: the stale-routed terminal write with the
// flip gate ON arms the deferred flip — SetGlobalEngine does not follow the
// commit — and the confirming push releases it exactly once.
func TestTerminalBranch_StaleRouteFlipStillDeferred(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.SetOptions(flipOnOptions())

	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText("x", 1, 1) // the stale-cache answer routes the write

	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("mode records after the stale-routed write = %d, want 0 — the flip defers on the "+
			"amendment path too (E3); log:\n%s", got, buf.String())
	}
	// The narrowed machine (plan 10-17): the arm is observable through its
	// delivery mark and the absent mode record — no grace deadline backs
	// the terminal slot.
	if !strings.Contains(buf.String(), `"msg":"correction delivery","stage":"flip-armed"`) {
		t.Fatalf("the stale-routed write armed no deferred flip — the round must defer its mode "+
			"rule (E3, plan 10-17); log:\n%s", buf.String())
	}

	line := "abc " + wordRU
	a.HandleSurroundingText(line, runeLen(line), runeLen(line)) // the confirming push

	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("mode records after the confirming push = %d, want exactly 1; log:\n%s", got, buf.String())
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru"}) {
		t.Errorf("panel symbols = %q, want exactly [ru]", got)
	}
}

// The plan 10-10 corpus (round-2 G-10-1 routing closure): the host
// re-acceptance rows of 2026-10-14 09:39–09:41 MSK (DEBUG-term-host-live-
// failures.md, Closure validation) replayed on the unit seams. On the host
// BOTH gestures ran LEVEL-1 (done in 1.9–2.4 ms, instant flip +0.03 ms)
// because terminal classification never engaged: the a11y identity is
// unavailable (role-unknown, ac_role_errors=47), the behavioral class is
// unlearned at gesture time, and the correction.terminals anchor has no
// identity source. The one live discriminator left in the journals is the
// ANSWER-ONLY push channel: the client never reported a field state in the
// context's whole life (the round arms over the empty cache — cacheNil) and
// answers Require with fresh state holding the typed word (the
// pendingVerdict MATCH branch — the branch the 10-08 amendment does not
// cover). The owner's approved stale-cache semantics (2026-10-13,
// «обещает окружение, но отвечает протухшим кэшем = всё равно терминал»)
// extends from the answer's CONTENT (10-08, the mismatch branch) to the
// answer's CHANNEL (plan 10-10, both verdict branches). The fixture pins
// decisions and op order; the journal latencies (1,9–2,4 мс) live only in
// origin comments — no timings are synthesized.

// TestReplayHost0939_MatchAnswerRoutesTerminalBranch replays the 09:39:41
// row: a fresh context whose role source errors at the decision, whose
// appid identity is unknown, whose client never pushed (nil cache at arm)
// and whose single Require answer is FRESH state holding the typed word —
// the match verdict. RED truth: the terminal write branch must receive
// control (learn + level-2 burst + the deferred flip + the noRung
// verify-after arm), level-1 must not run.
func TestReplayHost0939_MatchAnswerRoutesTerminalBranch(t *testing.T) {
	buf := captureLogsLevel(t, slog.LevelDebug)
	a, sink := wiredActor() // the caps bit announced — the 09:39 clients carry correct.CapSurroundingText
	// The host a11y state on the unit seams: the role source is consulted
	// at the decision and ALWAYS errors (ac_role_errors grows — 47 in the
	// host window); no appid source is started — the identity is unknown
	// and the config anchor answers nothing.
	a.UseRole(&fakeRole{err: errAppidBusDead})
	a.SetOptions(flipOnOptions()) // the host config: flip_after_correction=true (mode +0,03 ms after done)
	probe := &switchProbe{}
	a.SetSwitcher(probe.switcher)

	// The typing window: the client NEVER pushes — the round arms over the
	// empty cache (otherwise the cached-push instant path had answered with
	// a.surr != nil and the host rows would look different).
	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow) // action n=2 — the Require is out, nothing else

	if got := sink.opLog(); !slices.Equal(got, []string{requireOp}) {
		t.Fatalf("ops before the answer = %q, want exactly [%s] — the replay arms the pending round", got, requireOp)
	}

	// The client's single answer: FRESH state holding the typed word — the
	// pendingVerdict MATCH branch (the host done row +1,9–2,4 ms ran
	// level-1 here; the route must take this branch to the terminal write).
	end := runeLen(wordEN)
	a.HandleSurroundingText(wordEN, end, end)

	logged := buf.String()
	if !strings.Contains(logged, `"msg":"terminal class","verify_dead":true`) {
		t.Errorf("the answer-only signature did not learn the behavioral class; log:\n%s", logged)
	}
	if !strings.Contains(logged, `"msg":"correction","outcome":"done","changed":true`) {
		t.Fatalf("the replayed round never completed — the word was abandoned; log:\n%s", logged)
	}
	if !strings.Contains(logged, `"msg":"correction","level":2`) {
		t.Errorf("no level-2 DEBUG record — the terminal write branch did not receive control "+
			"(the host row ran level-1 here); log:\n%s", logged)
	}
	// The write shape: BackSpace burst + ONE commit — no level-1
	// DeleteSurroundingText+Commit pair (the host's «появляется и
	// удаляется» inversion lives on that pair).
	want := append([]string{requireOp}, burstOps(len([]rune(wordEN)), wordRU)...)
	want = append(want, requireOp) // the terminal verify-after arm (noRung)
	if got := sink.opLog(); !slices.Equal(got, want) {
		t.Fatalf("op log = %q, want the arm + the chunked terminal write %q + the verify-after arm — "+
			"the terminal branch must receive control, level-1 must not run", got, want)
	}
	// Defer-until-confirmed rides the routed round (the 10-08 form): no
	// flip traffic before the delivery confirmation — the host's instant
	// flip (+0,03 ms) is the level-1 misroute signature, never this path's.
	if got := countModeRecords(buf); got != 0 {
		t.Errorf("mode records before the confirming push = %d, want 0 — the routed round defers "+
			"its flip; log:\n%s", got, logged)
	}
	if got := probe.targets(); len(got) != 0 {
		t.Errorf("SetGlobalEngine before the confirmation = %q, want none (E3)", got)
	}
	if got := a.StatusSnapshot().AutoCorrectRoleErrors; got == 0 {
		t.Errorf("ac_role_errors = 0 — the replay must reproduce the host's erroring role source")
	}

	// The confirming push releases the deferred flip exactly once.
	line := "abc " + wordRU
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))
	if got := countModeRecords(buf); got != 1 {
		t.Errorf("mode records after the confirming push = %d, want exactly 1; log:\n%s", got, buf.String())
	}
	if got := probe.targets(); !slices.Equal(got, []string{engine.NameRU}) {
		t.Errorf("switch targets = %q, want exactly [%s]", got, engine.NameRU)
	}
}

// TestReplayHost0939_RouteLearnsVetoSeam pins the D-55 consequence of the
// route: the class the route learns silences the autocorrect boundary of
// the SAME context — classifyTerminal is the ONE seam both consumers read
// (the write pipeline and the arming veto), so they cannot diverge by
// construction (the plan's prohibition: the veto follows the same seam).
func TestReplayHost0939_RouteLearnsVetoSeam(t *testing.T) {
	a, _ := wiredActor()
	a.UseRole(&fakeRole{err: errAppidBusDead})
	a.SetOptions(session.Options{FlipAfterCorrection: true, AutoCorrectEnabled: true})

	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	end := runeLen(wordEN)
	a.HandleSurroundingText(wordEN, end, end) // the answer-only match route — the class learns

	a.HandleKey(spaceKey()) // a boundary in the routed context — the veto fires before the detector

	if got := a.AutoCorrectCounters().Reasons[acReasonTerminalVeto]; got != 1 {
		t.Errorf("terminal-veto abstentions after the route = %d, want 1 — the learned class must "+
			"silence the autocorrect boundary through the same seam", got)
	}
}

// TestReplayHost0939_RejectionSignaturesAbsent pins the replay's fidelity
// to the 09:39 block, NOT the 16:20 one: the window carries NO
// verify-timeout, NO verify-mismatch, and NO flap wipe — the learned class
// survives (exactly one terminal-class marker, zero verify_dead:false
// transitions), and the erroring role source is part of the reproduced
// state (ac_role_errors grows).
func TestReplayHost0939_RejectionSignaturesAbsent(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	a.UseRole(&fakeRole{err: errAppidBusDead})

	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	end := runeLen(wordEN)
	a.HandleSurroundingText(wordEN, end, end)

	logged := buf.String()
	for _, rejection := range []string{`"reason":"verify-timeout"`, `"reason":"verify-mismatch"`} {
		if got := strings.Count(logged, rejection); got != 0 {
			t.Errorf("%s rows in the replay window = %d, want 0 — the replay must reproduce the "+
				"09:39 block (clean of rejection signatures), not the 16:20 one", rejection, got)
		}
	}
	if got := strings.Count(logged, `"verify_dead":false`); got != 0 {
		t.Errorf("class wipes in the replay window = %d, want 0 — no flap pairs in the 09:39 window", got)
	}
	if got := strings.Count(logged, `"msg":"terminal class"`); got != 1 {
		t.Errorf("terminal-class markers = %d, want exactly 1 (the route's learn)", got)
	}
	if got := a.StatusSnapshot().AutoCorrectRoleErrors; got == 0 {
		t.Errorf("ac_role_errors = 0 — the 09:39 signature includes the erroring role source (47 in the host window)")
	}
}

// TestReplayHost0939_SpontaneousPushClientNotRouted pins the route's
// boundary from the honest side: a client whose push channel is ALIVE
// (spontaneous pushes tracked the typing — cache evidence at arm,
// cacheNil=false) keeps the byte-as-today world on the match verdict too —
// a fresh matching answer settles level-1, the route never fires, no class
// learns, no burst replays.
func TestReplayHost0939_SpontaneousPushClientNotRouted(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor() // identity/role unknown — no sources installed

	typeWord(a, wordEN)
	// The honest client's spontaneous push: the field state tracked during
	// typing — the cache carries evidence the push channel is alive.
	a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow) // the pending armed over the NON-nil cache
	end := runeLen("abc " + wordEN)
	a.HandleSurroundingText("abc "+wordEN, end, end) // the fresh matching answer
	a.HandleSurroundingText("abc ", 4, 4)            // the deletion proof completes the level-1 ladder (plan 10-11)

	logged := buf.String()
	if !strings.Contains(logged, `"msg":"correction","outcome":"done","changed":true`) {
		t.Fatalf("the honest round never completed; log:\n%s", logged)
	}
	if strings.Contains(logged, `"msg":"terminal class"`) {
		t.Errorf("a push-tracking client learned the terminal class — the route must key on the "+
			"answer-only CHANNEL, never on content alone; log:\n%s", logged)
	}
	if calls := sink.deleteCalls(); len(calls) != 1 {
		t.Errorf("deletions = %+v, want exactly one level-1 delete — the honest world stays byte-as-today", calls)
	}
	if got := len(sink.forwardCalls()); got != 0 {
		t.Errorf("forwards = %d, want 0 — no terminal burst on the honest world", got)
	}
}

// timerRecorder is the verify-deadline seam double of the WR-02 corpus: it
// records EVERY armed callback in arrival order, and its stop is a
// deliberate no-op — the production Stop is a no-op against an ALREADY
// FIRED deadline, which is exactly the callback the WR-02 window leaves
// runnable past the resolution of its own round.
type timerRecorder struct {
	mu  sync.Mutex
	fns []func()
}

// arm records one armed callback and returns the no-op stop (see above).
func (r *timerRecorder) arm(_ time.Duration, fn func()) func() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fns = append(r.fns, fn)

	return func() {} // Stop after fire: the callback stays runnable — the race under test
}

// nth hands out the n-th armed callback (0-based).
func (r *timerRecorder) nth(n int) func() {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.fns[n]
}

// TestVerifyExpiryStaleRound_DoesNotEatNextRound pins the WR-02 fix (the
// 10-REVIEW scenario): round 1's deadline already fired and waits on the
// actor mutex; the main path resolves round 1 and arms round 2 (Stop — a
// no-op against the fired callback); the stale callback then runs. Today
// it sees round 2's pending, consumes it (verify-timeout) and TEACHES the
// behavioral class for a context that never missed a budget — whose blast
// radius grew with 10-08 (the false class now also arms the deferred-flip
// machine). The pin: a stale callback is a no-op, round 2 lives, no class.
func TestVerifyExpiryStaleRound_DoesNotEatNextRound(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	arms := &timerRecorder{}
	a.UseVerifyTimer(arms.arm)

	// Round 1 in the honest-push world (cache evidence — a mismatching
	// answer resolves it byte-today, no terminal route involved).
	typeWord(a, wordEN)
	a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow) // round 1 armed — its callback is recorded
	stale := arms.nth(0)

	a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine)) // round 1 resolved
	if !strings.Contains(buf.String(), `"reason":"verify-mismatch"`) {
		t.Fatalf("round 1 never resolved by mismatch — the scenario is not set up; log:\n%s", buf.String())
	}

	// Round 2 arms while the stale callback is already runnable (the fired
	// deadline's Stop lost the race — the WR-02 window).
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	stale() // the fired-and-blocked callback finally runs against round 2

	logged := buf.String()
	if strings.Contains(logged, `"reason":"verify-timeout"`) {
		t.Errorf("the stale deadline consumed the live round (verify-timeout) — the WR-02 hole; log:\n%s", logged)
	}
	if strings.Contains(logged, `"msg":"terminal class"`) {
		t.Errorf("the stale deadline taught the behavioral class for a context that never missed a "+
			"budget — the WR-02 blast radius; log:\n%s", logged)
	}

	// Round 2 lives: the fresh matching answer settles it (level-1 — the
	// honest world), byte-as-today.
	end := runeLen("abc " + wordEN)
	a.HandleSurroundingText("abc "+wordEN, end, end)
	a.HandleSurroundingText("abc ", 4, 4) // the deletion proof completes the ladder (plan 10-11)
	if texts := sink.commitTexts(); len(texts) != 1 {
		t.Errorf("commits after round 2's answer = %d, want exactly 1 — round 2 must survive the "+
			"stale deadline", len(texts))
	}
	if calls := sink.deleteCalls(); len(calls) != 1 {
		t.Errorf("deletions = %d, want exactly 1 — round 2 settles level-1 in the honest world", len(calls))
	}
}

// TestCorrectionBusy_SecondEntrySkipsNotSecondBurst pins the WR-01 guard:
// while a level-2 burst is IN FLIGHT (the pacing seam parked, the mutex
// released at the chunk boundary), a second correction entry — CorrectNow
// over ctl — must refuse with the correction-busy slug and settle the
// combo instead of executing a second BuildPlan over the same un-replaced
// buffer: exactly ONE burst, exactly ONE commit, exactly ONE done.
func TestCorrectionBusy_SecondEntrySkipsNotSecondBurst(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.UseRole(&fakeRole{role: acRoleForbiddenTerminal}) // both entries route the terminal branch at the decision

	// The pacing seam parks the burst inside the FIRST chunk (under the
	// mutex) until the second entry is blocked on the actor mutex — the
	// in-flight window WR-01 describes.
	started := make(chan struct{})
	release := make(chan struct{})
	var sleeps atomic.Int64
	a.UseLevel2Sleeper(func(time.Duration) {
		if sleeps.Add(1) == 1 {
			close(started)
			<-release
		}
	})

	typeWord(a, termWord20) // 20 runes → 8+8+4 chunks: two boundary windows
	burstDone := make(chan struct{})
	go func() {
		defer close(burstDone)
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
	}()
	<-started // the burst is in flight, holding the mutex

	second := make(chan struct{})
	go func() {
		close(second)
		a.CorrectNow() // the ctl entry: blocks on the mutex until a boundary window
	}()
	<-second
	time.Sleep(20 * time.Millisecond) // past the 1 ms starvation threshold — the handoff is direct
	close(release)

	select {
	case <-burstDone:
	case <-time.After(5 * time.Second):
		t.Fatal("the burst never settled — deadlock in the busy interleave")
	}

	logged := buf.String()
	if !strings.Contains(logged, `"reason":"correction-busy"`) {
		t.Errorf("no correction-busy refusal — the second entry executed instead of skipping; log:\n%s", logged)
	}
	if texts := sink.commitTexts(); len(texts) != 1 {
		t.Errorf("commits = %d, want exactly 1 — a second destructive burst replayed over the same "+
			"un-replaced buffer (WR-01)", len(texts))
	}
	if got := len(sink.forwardCalls()); got != 2*len([]rune(termWord20)) {
		t.Errorf("forwards = %d, want exactly %d — one burst, never two", got, 2*len([]rune(termWord20)))
	}
	if got := strings.Count(logged, `"msg":"correction","outcome":"done","changed":true`); got != 1 {
		t.Errorf("done records = %d, want exactly 1; log:\n%s", got, logged)
	}
}

// The plan 10-11 corpus (round-2 G-10-1 delivery closure): the host
// re-acceptance rows of 2026-10-14 09:39 (DEBUG-term-host-live-failures.md,
// Closure validation — verdict REJECTED, «вижу привет, и он удаляется» in
// BOTH terminals) replayed on the unit seams. Both host gestures rode
// LEVEL-1 (done in 1,9–2,4 ms), where executeCorrection emits
// DeleteSurroundingText and CommitText back-to-back with no confirming
// boundary — and the host client's processing reorders the tight pair (the
// A2 reorder class documented live by the 10-05 rows for the
// forwards/commit pair): the insertion lands first and is VISIBLE, the
// late deletion erases the fresh word. The fix under test: a two-phase
// level-1 — delete, then the commit ONLY after the client proves the
// deletion applied (a push or Require answer no longer ending with
// token+tail), with a budget fallback that degrades to today's pair with a
// loud INFO mark; the level-1 flip defers onto the 10-08 machine. The
// journal latencies (1,9–2,4 мс; the E3 +16 ms push) live only in origin
// comments — no timings are synthesized.
//
// Level-1 OLD-semantics pins inventoried for migration (each keeps its
// meaning; the commit phase moves behind the confirm window and the
// verify-after arm moves into the flush): TestActor_DoubleTapCorrects,
// TestActor_CachedSurroundingCorrects, TestActor_AfterSpaceCorrects,
// TestActor_DoubleTapPunctuationStretch, TestActor_ToggleRepeat,
// TestActor_DebugCorrectionRecord, TestActor_TripleTapCorrectsPhrase,
// TestActor_PhraseMixedCorrects, TestActor_VerifyAfterLevel1,
// TestActor_MixedWordInvertsPerChar, TestActor_OwnFlipMixedWordCorrectsWhole,
// TestActor_ComboWordThenFlip, TestActor_ComboStillSingleFlipWithCorrectionFlip,
// TestActor_FlipAfterWordCorrection, TestActor_FlipAfterPhraseCorrection,
// TestActor_FlipDirectionFollowsConvertedScript,
// TestActor_CorrectionFlipSetsResultEngine, TestActor_FlipOrderPinned,
// TestActor_FlipSoundsSink, TestActor_ClipboardRungAfterMismatch,
// TestActor_SwitcherFailureWarnsNotFatal, TestActor_SwitcherDeadlineBounded,
// TestActor_NilSwitcherInternalFlip, TestActor_ReentrantAttachDuringFlip,
// TestActor_AttachEngineWhileFlipInFlight,
// TestReplayHost0939_SpontaneousPushClientNotRouted, and the autocorrect
// pipeline rounds that settle a word correction level-1
// (TestAutoCorrect_FiresThroughPipeline et al.). Selection, level-2,
// combo-flip and flip=false worlds are byte-frozen and stay untouched.

// level1Budget is the pinned confirm budget of the two-phase level-1
// delivery: the production constant's corpus mirror — the healthy E3 push
// ×4 headroom, never a config key. (The terminal flip's own grace — the
// old termFlipGrace mirror — is gone with plan 10-17's narrowed release
// machine: the terminal slot arms no deadline at all.)
const level1Budget = 60 * time.Millisecond

// invertingField is the model-B simulated field (plan 10-11): a consumer
// that applies the sink's ops to a string state with the A2 reorder class —
// within ONE arrival batch the COMMIT lands before the DELETE (the
// host-class client processes the tight pair inverted, the cursor riding
// the committed text), so the deletion erases the fresh word. Across
// batches every received op is applied — the client is late, never lossy.
// The delete model: the range ends at the (end-of-text) cursor, nchars
// runes come off the tail.
type invertingField struct {
	text []rune
}

// apply replays one arrival batch in the client's inverted order.
func (f *invertingField) apply(ops []string) {
	var commits []string
	var deletes []deleteCall
	for _, op := range ops {
		switch {
		case strings.HasPrefix(op, "commit("):
			commits = append(commits, strings.TrimSuffix(strings.TrimPrefix(op, "commit("), ")"))
		case strings.HasPrefix(op, "delete("):
			var off int32
			var n uint32
			if _, err := fmt.Sscanf(op, "delete(%d,%d)", &off, &n); err != nil {
				panic("unparseable delete op: " + op)
			}
			deletes = append(deletes, deleteCall{offset: off, nchars: n})
		}
	}
	for _, c := range commits {
		f.text = append(f.text, []rune(c)...)
	}
	for _, d := range deletes {
		if int(d.nchars) > len(f.text) {
			panic("delete beyond the simulated field")
		}
		f.text = f.text[:len(f.text)-int(d.nchars)]
	}
}

// typeCachedWord types wordEN with the spontaneously-pushing client's push
// after every keystroke — the cache tracks the run (cacheNil=false at arm),
// so the 10-10 route never fires and the gesture rides the cached-push
// instant level-1: the model-B world of the 09:39 rows.
func typeCachedWord(a *session.Actor) {
	for i, r := range wordEN {
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r)})
		line := wordEN[:i+1]
		a.HandleSurroundingText(line, runeLen(line), runeLen(line))
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r), Release: true})
	}
}

// TestTwoPhaseLevel1_InvertingConsumerKeepsReplacement replays the 09:39
// «вижу привет, и он удаляется» on the model-B consumer: the
// spontaneously-pushing client (identity/role unavailable, no flap pairs)
// whose gesture lands the cached instant level-1. RED truth: the commit
// shares the delete's arrival batch and the inverting consumer ends with
// the replacement ERASED. Required: the commit is absent from the wire
// right after the delete; it appears only after the client's confirming
// push (its state no longer ends with token+tail); the simulated field
// holds «привет» at the end — under ANY client processing order.
func TestTwoPhaseLevel1_InvertingConsumerKeepsReplacement(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	// The 09:39 host identity: the role source errors (ac_role_errors), no
	// appid — nothing classifies the context terminal; the client pushes,
	// so the answer-only route stays closed.
	a.UseRole(&fakeRole{err: errAppidBusDead})
	field := &invertingField{text: []rune(wordEN)}

	typeCachedWord(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow) // action n=2 — the cached match arms the instant level-1

	// Arrival batch 1: what followed the decision. The commit must not
	// share the delete's batch — that co-arrival IS the A2 inversion.
	n0 := len(sink.opLog())
	batch1 := sink.opLog()
	if !slices.Contains(batch1, "delete(-6,6)") {
		t.Fatalf("no deletion in the gesture batch: %q — the model-B round must run level-1", batch1)
	}
	if slices.Contains(batch1, "commit("+wordRU+")") {
		t.Fatalf("CommitText shares the delete's arrival batch (%q) — the A2 inversion erases the "+
			"fresh word on a reordering client; the commit phase must wait for the deletion proof", batch1)
	}
	field.apply(batch1)

	// The client processed the deletion and pushes the post-delete state —
	// the confirming push. Only now may the commit phase run.
	a.HandleSurroundingText("", 0, 0)
	batch2 := sink.opLog()[n0:]
	if !slices.Contains(batch2, "commit("+wordRU+")") {
		t.Fatalf("no commit after the confirming push: %q — the ladder must complete once the "+
			"deletion is proven applied", batch2)
	}
	field.apply(batch2)

	if got := string(field.text); got != wordRU {
		t.Fatalf("simulated field after the inverted processing = %q, want %q — the replacement must "+
			"survive ANY client processing order", got, wordRU)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done","changed":true`) {
		t.Errorf("INFO completion record missing; log:\n%s", buf.String())
	}
}

// TestTwoPhaseLevel1_SilentClientDegradesLoud pins the fallback: a client
// that stays silent the whole budget degrades to TODAY'S pair (delete+commit
// — never a «deletion without replacement»), with the loud INFO outcome
// slug in the journal, and the verify-after round armed on the exit — the
// checks never disappear, whatever the client does.
func TestTwoPhaseLevel1_SilentClientDegradesLoud(t *testing.T) {
	buf := captureLogsLevel(t, slog.LevelDebug)
	a, sink := wiredActor()
	clock := &flipClock{} // the level1ConfirmBudget recorder (the flipClock form)
	a.UseLevel1Timer(clock.arm)

	typeCachedWord(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if calls := sink.deleteCalls(); len(calls) != 1 {
		t.Fatalf("deletions after the gesture = %d, want exactly 1", len(calls))
	}
	if got := sink.commitTexts(); len(got) != 0 {
		t.Fatalf("commits right after the delete = %q, want none — the commit phase waits for the "+
			"deletion proof (the silent client has not given it yet)", got)
	}
	armed := clock.armedDurations()
	if len(armed) != 1 || armed[0] != level1Budget {
		t.Fatalf("confirm budgets armed = %v, want exactly [%v] — the fallback deadline is the "+
			"silent client's only completion path", armed, level1Budget)
	}
	if got := sink.requireCount(); got != 0 {
		t.Fatalf("requires before the fallback = %d, want 0 — the verify-after arms at the flush, "+
			"not at the delete", got)
	}

	clock.fire() // the budget closed with no proof — today's pair, loudly marked

	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != wordRU {
		t.Fatalf("commits after the fallback = %q, want exactly one %q — the silent client degrades "+
			"to the today's semantics, never a deletion without replacement", texts, wordRU)
	}
	if !strings.Contains(buf.String(), `"msg":"correction confirm","outcome":"timeout"`) {
		t.Errorf("the loud confirm-timeout INFO slug is missing; log:\n%s", buf.String())
	}
	if got := sink.requireCount(); got != 1 {
		t.Errorf("requires after the fallback = %d, want exactly 1 — the verify-after arms on the "+
			"exit, the checks never disappear", got)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done","changed":true`) {
		t.Errorf("INFO completion record missing after the fallback; log:\n%s", buf.String())
	}
}

// TestTwoPhaseLevel1_FlipDeferredUntilConfirmingPush pins the level-1 flip
// moving onto the 10-08 machine: after the delete there is NO switcher
// traffic until the delivery is proven; the confirming push flushes the
// commit and ARMS the flip (no traffic yet); the verify-after's confirming
// push (the match) releases it exactly once (the FlipDeferredUntilConfirmed
// form). The owner rule «режим следует за скриптом результата» is deferred,
// not cancelled.
func TestTwoPhaseLevel1_FlipDeferredUntilConfirmingPush(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.SetOptions(flipOnOptions()) // the owner's config: flip follows the correction
	probe := &switchProbe{}
	a.SetSwitcher(probe.switcher)

	typeCachedWord(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	// RED truth: today the level-1 flip is IMMEDIATE — the mode record and
	// the SetGlobalEngine sit right next to the done row (the 09:39 rows'
	// «mode to ru +0,03 ms»), the exact orphaning geometry E3 convicted.
	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("mode records right after the delete = %d, want 0 — the level-1 flip must defer "+
			"onto the 10-08 machine; log:\n%s", got, buf.String())
	}
	if got := probe.targets(); len(got) != 0 {
		t.Fatalf("SetGlobalEngine before any delivery proof = %q, want none", got)
	}

	// The confirming push completes the ladder; the flip is armed, not fired.
	a.HandleSurroundingText("", 0, 0)
	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("mode records after the confirming push = %d, want 0 — the flip waits for its own "+
			"delivery confirmation; log:\n%s", got, buf.String())
	}
	if got := probe.targets(); len(got) != 0 {
		t.Fatalf("SetGlobalEngine after the ladder flush but before the flip proof = %q, want none", got)
	}

	// The verify-after's confirming push (the field ends with converted) —
	// the same match criterion the 10-08 release reads — fires the flip.
	line := wordRU
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))
	if got := countModeRecords(buf); got != 1 {
		t.Fatalf("mode records after the flip proof = %d, want exactly 1; log:\n%s", got, buf.String())
	}
	if got := probe.targets(); !slices.Equal(got, []string{engine.NameRU}) {
		t.Errorf("switch targets = %q, want exactly [%s]", got, engine.NameRU)
	}
	if got := sink.modeSymbols(); !slices.Equal(got, []string{"ru"}) {
		t.Errorf("panel symbols = %q, want exactly [ru]", got)
	}
}

// TestTwoPhaseLevel1_FlipReleasedByNextKey pins the push-less client's
// level-1 release site (the wezterm-class silence): the first keystroke
// after the delete flushes the ladder (the replacement is on the wire
// before the next input) and releases the flip STRICTLY before the key is
// processed — the key already types in the RESULT mode (the
// FlipReleasedByNextKey form, plan 10-11's feedKey entry).
func TestTwoPhaseLevel1_FlipReleasedByNextKey(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	a.SetOptions(flipOnOptions())
	probe := &switchProbe{}
	a.SetSwitcher(probe.switcher)

	typeCachedWord(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("mode records before the next key = %d, want 0 — the flip waits; log:\n%s",
			got, buf.String())
	}
	wireBefore := len(sink.opLog())
	all := sink.opLog()
	if !slices.Contains(all, "delete(-6,6)") {
		t.Fatalf("no deletion before the keystroke: %q — the model-B round must run level-1", all)
	}

	a.HandleKey(engine.EngineEvent{Keyval: 'x'}) // EN-mode keyval — the flush and the release first

	ops := sink.opLog()[wireBefore:]
	join := func(ops []string) string { return strings.Join(ops, " ") }
	comIdx := slices.Index(ops, "commit("+wordRU+")")
	modeIdx := slices.Index(ops, termFlipRUOp)
	keyIdx := slices.Index(ops, "commit(ч)")
	if comIdx < 0 || modeIdx < 0 || keyIdx < 0 {
		t.Fatalf("flush-at-keystroke op batch = %s — want the flushed commit, the "+
			"released flip and the key's own commit all present", join(ops))
	}
	if comIdx >= modeIdx || modeIdx >= keyIdx {
		t.Fatalf("op batch order = %s — want delete < flushed commit < mode(ru) < commit(ч): the "+
			"replacement is on the wire before the next input and the release precedes the key's "+
			"processing (the key types in the result mode)", join(ops))
	}
	if got := countModeRecords(buf); got != 1 {
		t.Errorf("mode records after the key = %d, want exactly 1 — one flip per correction; log:\n%s",
			got, buf.String())
	}
}

// TestFlapWindowContextIdentity_RealSwitchWipesClass pins the WR-03 fix
// (10-REVIEW :119-148): a REAL window switch completing inside the flap
// window — between the FocusOut and the FocusIn the input context is
// re-minted (AttachEngine of a NEW engine object) — must NOT cancel the
// deferred wipe: the learned class dies with the OLD context, the fresh
// context starts unclassified (the invariant «класс живого контекста»).
func TestFlapWindowContextIdentity_RealSwitchWipesClass(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()
	clock := &flapClock{}
	a.UseFlapTimer(clock.arm)

	teachVerifyDead(t, a, wordEN)
	a.HandleLifecycle(engine.LifecycleFocusOut) // arms the window over THIS context's emitter
	a.AttachEngine(&fakeSink{})                 // the context was re-minted between the pair
	a.HandleLifecycle(engine.LifecycleFocusIn)  // a DIFFERENT engine object — no cancellation
	clock.fire()                                // the window expires — the class wipes

	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 2 {
		t.Fatalf("terminal-class markers after the real switch = %d, want exactly 2 (learn + clear) — "+
			"a different engine object in the FocusIn must not cancel the wipe (WR-03); log:\n%s",
			got, buf.String())
	}
	if !strings.Contains(buf.String(), `"verify_dead":false`) {
		t.Errorf("the cleared-class marker is missing — the class leaked into the new context; log:\n%s",
			buf.String())
	}

	// The fresh context learns afresh: the state, not just the log, reset.
	typeWord(a, "cnhjrf")
	a.HandleKey(spaceKey())
	teachVerifyDead(t, a, wordEN)
	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 3 {
		t.Errorf("terminal-class markers after the re-learn = %d, want 3 — the new context must "+
			"classify from zero", got)
	}
}

// TestFlapWindowContextIdentity_SameEnginePairSurvives is the protective
// pin riding the WR-03 fix: the same-engine flap pair (the journal's 1–3 ms
// churn pairs, one engine object on both halves) cancels the wipe exactly
// as before — the 10-07 fix does not regress (the
// TestTerminalClass_VerifyDeadSurvivesFlapPair control, re-expressed on the
// identity seam; it must stay green UNCHANGED).
func TestFlapWindowContextIdentity_SameEnginePairSurvives(t *testing.T) {
	buf := captureLogs(t)
	a, _ := wiredActor()

	teachVerifyDead(t, a, wordEN)
	a.HandleLifecycle(engine.LifecycleFocusOut)
	a.HandleLifecycle(engine.LifecycleFocusIn) // the same engine object — the flap pair completes

	if got := strings.Count(buf.String(), `"msg":"terminal class"`); got != 1 {
		t.Fatalf("terminal-class markers after the same-engine pair = %d, want exactly 1 (learn "+
			"only) — the pair must cancel the wipe as in 10-07; log:\n%s", got, buf.String())
	}
	if strings.Contains(buf.String(), `"verify_dead":false`) {
		t.Errorf("the class was wiped by a same-engine flap pair — the 10-07 fix regressed; log:\n%s",
			buf.String())
	}
}

// TestLevel1ConfirmFlush_FocusOutUsesCapturedSnapshot pins the flush
// safety contract (T-10-11-02): the ladder's completion commits through the
// emitter snapshot CAPTURED AT ARM — never the current emitter(). A context
// re-mint between the arm and the flush (the churn reality) must not pour
// the replacement into the NEW application's engine: the stale object goes
// to the old sink, the new sink stays byte-silent.
func TestLevel1ConfirmFlush_FocusOutUsesCapturedSnapshot(t *testing.T) {
	a, sink := wiredActor()

	typeCachedWord(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.commitTexts(); len(got) != 0 {
		t.Fatalf("commits right after the delete = %q, want none — the ladder is mid-confirmation", got)
	}

	fresh := &fakeSink{}
	a.AttachEngine(fresh)                       // the context was re-minted after the arm
	a.HandleLifecycle(engine.LifecycleFocusOut) // the flush runs — over the CAPTURED snapshot

	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != wordRU {
		t.Fatalf("commits on the captured engine = %q, want exactly one %q — the flush must complete "+
			"the ladder over the armed snapshot (byte-parity: both ops already on the wire today)", texts, wordRU)
	}
	if got := fresh.commitTexts(); len(got) != 0 {
		t.Fatalf("commits into the RE-MINTED engine = %q, want none — a flush must never read the "+
			"current emitter(), the text would land in another application (T-10-11-02)", got)
	}
}

// The DST-guard corpus (plan 10-16, round-4 gap G-10-1, the 2026-10-16
// bisection): the stand bisection proved DeleteSurroundingText with ANY
// parameters toward a text-input-v3 bridge context whose surrounding state
// was never materialized — no spontaneous push, no materializing answer in
// the context's CURRENT generation — segfaults gnome-shell (mutter 46.2,
// strlen over the virgin surrounding state, meta-wayland-text-input.c:254;
// three cores, one frame — R2/R6b/R7). The guard rule under test: the
// correction machine emits DST only with the generation-scoped evidence of
// materialization; without it the replacement rides the proven forward
// branch (the executeLevel2 form — chunked BackSpace forwards + one commit,
// the write primitive verdict (c) reuses) behind a loud, content-free
// boundary. The guard is safety, not a behavior change: an honestly pushing
// client keeps the byte-as-today two-phase level-1 (TestDstGuard_
// MaterializedContextByteUnchanged), RequireSurroundingText is never
// suppressed (safe by R5b), and the selection path stays untouched.
// RED truth (strict TDD, directive 1): on the unguarded tree
// executeCorrection emits DST unconditionally — a delete lands on the wire
// toward a never-materialized context, and both negative tests fail on that
// observable (evidence-20261016-10-16-red-task1.json).

// TestDstGuard_NeverMaterializedContextGetsForwards pins the guard's core
// rule: a level-1 round toward a context whose CURRENT generation never
// materialized its surrounding state (no push, no answer — a fresh engine
// object with the push channel silent) gets ZERO DeleteSurroundingText on
// the wire; the replacement arrives the forward way (the burstOps shape:
// chunked BackSpace forwards, then ONE commit, in order) behind the loud
// content-free diversion boundary (D-20/D-21). RED truth: the unguarded
// executeCorrection rides the previous generation's stale cache straight
// into the virgin context — the delete lands (the gnome-shell killer).
func TestDstGuard_NeverMaterializedContextGetsForwards(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	// Generation 1: ONE spontaneous push materializes the cache on THIS
	// engine object.
	line := "abc " + wordEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))
	// The context generation is re-minted WITHOUT a lifecycle pair (engine
	// objects are minted per input context; the host churn can swallow the
	// focus pair): a fresh VIRGIN generation on a new engine object. The
	// cache of the dead generation survives the mint — the exact residual
	// route the guard exists for.
	fresh := &fakeSink{}
	a.AttachEngine(fresh)
	// The gesture types on the fresh generation; the push channel stays
	// silent the whole time — nothing materializes the fresh context.
	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow) // the stale cache matches — the level-1 round

	if calls := fresh.deleteCalls(); len(calls) != 0 {
		t.Fatalf("DeleteSurroundingText toward the never-materialized context = %+v, want none — "+
			"DST into a fresh generation is the compositor killer (bisection 2026-10-16, "+
			"R2/R6b/R7: any parameters, ≲1.5 s window)", calls)
	}
	if got, want := fresh.opLog(), burstOps(len([]rune(wordEN)), wordRU); !slices.Equal(got, want) {
		t.Fatalf("op log on the fresh engine = %q, want the forward-branch replacement %q — "+
			"forwards then one commit, in order (verdict (c))", got, want)
	}
	logged := buf.String()
	if !strings.Contains(logged, `"msg":"correction","outcome":"dst-guarded"`) {
		t.Errorf("the loud diversion boundary is missing from the journal; log:\n%s", logged)
	}
	if strings.Contains(logged, wordEN) || strings.Contains(logged, wordRU) {
		t.Errorf("the guard boundary carries field content (D-20/D-21 — outcomes and counters only); "+
			"log:\n%s", logged)
	}
	// The diversion is machine-visible in the status surface (T-10-16-04,
	// the ac_role_errors form).
	if got := a.StatusSnapshot().CorrDstGuarded; got != 1 {
		t.Errorf("corr_dst_guarded counter = %d, want 1 — the diversion must be countable", got)
	}
	// The dead generation's engine stays untouched by the fresh round.
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Errorf("deletions on the previous generation's engine = %+v, want none", calls)
	}
}

// TestDstGuard_MaterializedContextByteUnchanged is the protective pin riding
// the guard: an honestly pushing client (the pushes track the typing on THIS
// generation's engine — the materialization evidence is live) keeps the
// byte-as-today two-phase level-1 — exactly one delete over the token+tail
// range (the CORR-07 geometry), the commit waiting behind the 10-11
// confirmation machine, ZERO forwards, the guard silent. Green on the
// unguarded tree by construction; it must STAY green once the guard exists —
// safety must never cost the honest clients their delivery shape.
func TestDstGuard_MaterializedContextByteUnchanged(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	typeCachedWord(a) // the pushes track the typing — the evidence is live
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow) // the fresh cached match — the instant level-1

	if calls := sink.deleteCalls(); len(calls) != 1 || calls[0] != (deleteCall{offset: -6, nchars: 6}) {
		t.Fatalf("deletions on the materialized context = %+v, want exactly one delete(-6,6) over "+
			"token+tail (CORR-07)", calls)
	}
	if got := sink.commitTexts(); len(got) != 0 {
		t.Fatalf("commits before the deletion proof = %q, want none — the two-phase machine (plan "+
			"10-11) must not be diverted by the guard", got)
	}
	if strings.Contains(buf.String(), `"outcome":"dst-guarded"`) {
		t.Errorf("the guard fired on a materialized context — safety degraded an honest client; "+
			"log:\n%s", buf.String())
	}

	// The confirming push completes the ladder byte-as-today.
	a.HandleSurroundingText("", 0, 0)
	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != wordRU {
		t.Fatalf("commits after the confirming push = %q, want exactly one %q", texts, wordRU)
	}
	if got := len(sink.forwardCalls()); got != 0 {
		t.Errorf("forwards on the honest world = %d, want 0 — the forward branch must never touch "+
			"a materialized context", got)
	}
	// The guard never fired — the counter reads zero on the honest world.
	if got := a.StatusSnapshot().CorrDstGuarded; got != 0 {
		t.Errorf("corr_dst_guarded on the honest world = %d, want 0", got)
	}
}

// TestDstGuard_EvidenceDiesWithGeneration pins the generation scoping of the
// materialization evidence: it must live and die WITH the context's cache.
// The push materializes on generation 1; the FocusOut/Reset branch (форма
// ~:1166 — a.surr=nil, ctxEpoch++) kills both; a post-reset push re-arms the
// evidence BOUND to that generation's engine object; and the churn-missed
// re-mint onto a fresh virgin engine object must NOT inherit it. RED truth:
// the unguarded tree knows no evidence at all — the fresh round rides the
// stale cache and the delete lands on the virgin context.
func TestDstGuard_EvidenceDiesWithGeneration(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	// Generation 1: the push materializes cache and evidence together.
	line := "abc " + wordEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))
	// The generation dies (the FocusOut/Reset form): cache and evidence die
	// with it.
	a.HandleLifecycle(engine.LifecycleFocusOut)
	// Generation 2: the client pushes again — the evidence re-arms, bound
	// to THIS generation's engine object...
	a.HandleSurroundingText(line, runeLen(line), runeLen(line))
	// ...and the context is re-minted WITHOUT a lifecycle pair: a fresh
	// VIRGIN generation whose engine object never saw a push. The evidence
	// of the dead object must not vouch for it.
	fresh := &fakeSink{}
	a.AttachEngine(fresh)
	// The fresh round rides the stale cache of the dead generation.
	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if calls := fresh.deleteCalls(); len(calls) != 0 {
		t.Fatalf("DST reached the fresh generation after the evidence died = %+v, want none — "+
			"evidence that outlives its generation is the virgin-window hole (T-10-16-02)", calls)
	}
	if got, want := fresh.opLog(), burstOps(len([]rune(wordEN)), wordRU); !slices.Equal(got, want) {
		t.Fatalf("op log on the fresh engine = %q, want the forward-branch replacement %q", got, want)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"dst-guarded"`) {
		t.Errorf("the loud diversion boundary is missing from the journal; log:\n%s", buf.String())
	}
	// The dead generation's engine stays untouched by the fresh round.
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Errorf("deletions on the dead generation's engine = %+v, want none", calls)
	}
}

// The round-4 replay corpus (plan 10-17, gap G-10-1, the owner mandate of
// 2026-10-14/16 «вижу исправленное слово, после чего оно удаляется»): the
// REAL host journal rows of 2026-10-14 19:21:26 (DEBUG-term-host-live-
// failures.md round-2 block — terminal level-2, done +144 ms, deferred flip
// +104 ms, the word seen then erased) and 19:36:00 (done +76 ms, the
// second sample) replayed on the unit seams — RED before the round-4
// machine exists (the 2026-10-12 lesson: a fix is proven by replaying the
// literal rows BEFORE any stand run and any owner question), GREEN after.
//
// The consumer model (the TwoPhaseLevel1_InvertingConsumer form, adapted to
// the level-2 reorder class): an answer-only terminal client whose
// CommitText applies SYNCHRONOUSLY while every ForwardKeyEvent(BackSpace)
// queues and applies LATE (the host client-drain class — the 2026-10-12
// live finding «ghbdtn»→«ghприв»). The client's clock is the machine's own
// pacing (the level2Sleep seam records every sleep duration), so a
// per-forward processing lag is expressed deterministically: a forward
// applies when the machine's elapsed pacing time reaches arrival+lag.

// replayDrainedLag is the client lag INSIDE the machine's delivery window:
// the E3 healthy push class (16 ms) at the host-churn scale — a lag the
// window must absorb for the word to survive by construction.
const replayDrainedLag = 60 * time.Millisecond

// replayBeyondLag is the client lag BEYOND the whole correction package:
// the 19:21:26 row class (done +144 ms with the entire 6×20 ms package
// behind it — the host client was still draining when the commit landed).
// Past the window the machine claims NO persistence — the zone's contract
// is loudness (delivery marks on every boundary), never a field promise.
const replayBeyondLag = 300 * time.Millisecond

// replayTailWord is the replayed row's field content: «ghbdtn» (6 runes)
// plus the trailing space of the UAT recipe — 7 BackSpaces, commit
// «привет » (the D-13 tail rides the replay, the TailRidesReplay form).
const replayTailWord = wordEN + " "

// replayTailConverted is the same range after the EN→RU conversion.
const replayTailConverted = wordRU + " "

// The replay matrix's zone and release-cause tags (goconst — named once).
const (
	replayZoneDrained = "drained"
	replayZoneBeyond  = "beyond"

	releasePush      = "push"
	releaseKeystroke = "keystroke"
	releaseGrace     = "grace"
	releaseTimeout   = "timeout"
	releaseMismatch  = "mismatch"
)

// replayWezterm is the answer-only terminal client of the 19:21:26 class:
// the field starts at the typed stretch, commits apply synchronously,
// forwarded BackSpaces queue and apply late by the fixed lag. The clock
// advances only by the machine's own pacing sleeps (the level2Sleep seam)
// plus explicit advance() calls — deterministic, no real timers.
type replayWezterm struct {
	field    []rune
	lag      time.Duration
	clock    time.Duration
	arrivals []time.Duration
}

// tick advances the client clock by one machine sleep and applies every
// forward whose lag has elapsed (the sleeper seam callback).
func (c *replayWezterm) tick(d time.Duration) {
	c.clock += d
	c.applyDue()
}

// forward records one queued BackSpace PRESS at the current machine time
// (the sink's forwardHook callback). Round-6 pair discipline: the release
// half of a pair is a wire no-op for the terminal's field — a real client
// erases on the keypress and ignores the release — so only presses join
// the lag queue (a pair-unaware model would erase twice per key).
func (c *replayWezterm) forward(call forwardCall) {
	if call.state&engine.MaskRelease != 0 {
		return
	}
	c.arrivals = append(c.arrivals, c.clock)
}

// commit applies the due forwards FIRST, then the replacement
// synchronously — the client class the host rows convict (the commit
// never waits for the queue).
func (c *replayWezterm) commit(text string) {
	c.applyDue()
	c.field = append(c.field, []rune(text)...)
}

// advance moves the client clock forward after the round (the post-commit
// queue drain) and applies whatever became due.
func (c *replayWezterm) advance(d time.Duration) {
	c.tick(d)
}

// applyDue erases one trailing rune per due BackSpace (the cursor sits at
// the end of the simulated field).
func (c *replayWezterm) applyDue() {
	due := 0
	for _, at := range c.arrivals {
		if c.clock >= at+c.lag {
			due++
		}
	}
	if due == 0 {
		return
	}
	c.arrivals = c.arrivals[due:]
	for range due {
		if len(c.field) > 0 {
			c.field = c.field[:len(c.field)-1]
		}
	}
}

// replayRowEnv is the wired environment of one replayed row — the pieces
// the cells assert on (the struct keeps replayRow's result count civil).
type replayRowEnv struct {
	a      *session.Actor
	sink   *fakeSink
	probe  *switchProbe
	clock  *flipClock
	client *replayWezterm
	buf    *syncBuffer
}

// replayRow drives one literal 19:21:26-class row: the learned terminal
// class (the host row's own state — «выученный класс терминала жив на
// хосте впервые»), the typed «ghbdtn » stretch, the double Right Shift
// gesture, the terminal level-2 write.
func replayRow(t *testing.T, lag time.Duration) *replayRowEnv {
	t.Helper()
	buf := captureLogs(t) // capture BEFORE the teach round: the cells assert the row's own records
	a, sink := wiredActor()
	teachTerminalClass(t, a)
	a.SetOptions(flipOnOptions()) // the owner's config: flip follows the correction
	probe := &switchProbe{}
	a.SetSwitcher(probe.switcher)
	clock := &flipClock{}
	a.UseFlipTimer(clock.arm)
	client := &replayWezterm{field: []rune(replayTailWord), lag: lag}
	a.UseLevel2Sleeper(client.tick)
	sink.forwardHook = func(call forwardCall) { client.forward(call) }
	sink.commitHook = client.commit

	typeWord(a, replayTailWord)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	return &replayRowEnv{a: a, sink: sink, probe: probe, clock: clock, client: client, buf: buf}
}

// refuteModeRecord fails the cell the moment a mode record lands — the
// deterministic refutation of an early flip release (the grace fire, the
// verify outcomes): the round-4 machine's flip must never ride them.
func refuteModeRecord(t *testing.T, buf *syncBuffer) {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if got := countModeRecords(buf); got != 0 {
			t.Fatalf("a mode record released the flip outside its cause (%d records) — SetGlobalEngine "+
				"fired between the commit and the honest release; log:\n%s", got, buf.String())
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// replayCell is one matrix cell: the client lag zone × the flip release
// cause the cell drives after the round.
type replayCell struct {
	name    string
	lag     time.Duration
	release string
	zone    string
}

// replayMatrixCells is the zoned matrix of the 19:21:26 row: inside the
// machine's delivery window the field must hold the word under ANY release
// cause; beyond it the contract is loudness, never persistence.
func replayMatrixCells() []replayCell {
	return []replayCell{
		{"drained-push", replayDrainedLag, releasePush, replayZoneDrained},
		{"drained-keystroke", replayDrainedLag, releaseKeystroke, replayZoneDrained},
		{"drained-grace-dead", replayDrainedLag, releaseGrace, replayZoneDrained},
		{"drained-verify-timeout", replayDrainedLag, releaseTimeout, replayZoneDrained},
		{"drained-verify-mismatch", replayDrainedLag, releaseMismatch, replayZoneDrained},
		{"beyond-push-loud", replayBeyondLag, releasePush, replayZoneBeyond},
		{"beyond-keystroke-loud", replayBeyondLag, releaseKeystroke, replayZoneBeyond},
	}
}

// replayAssertRoundForm pins the row's form signature: the terminal branch
// executed (the done row), the flip is still deferred (no mode record —
// the +104 ms flip of the row rode a release cause, never the commit),
// and no SetGlobalEngine sits between the commit and the release cause.
func replayAssertRoundForm(t *testing.T, env *replayRowEnv) {
	t.Helper()
	if !strings.Contains(env.buf.String(), `"msg":"correction","outcome":"done","changed":true`) {
		t.Fatalf("the replayed round never completed; log:\n%s", env.buf.String())
	}
	if got := countModeRecords(env.buf); got != 0 {
		t.Fatalf("mode records right after the round = %d, want 0 — the flip is deferred; log:\n%s",
			got, env.buf.String())
	}
	if got := env.probe.targets(); len(got) != 0 {
		t.Fatalf("SetGlobalEngine between the commit and the release cause = %q, want none", got)
	}
}

// replayAssertDeliveryMarks pins the marks' presence and order: chunks-sent
// → commit-sent → flip-armed on every round (a silent boundary is the
// three-round lesson).
func replayAssertDeliveryMarks(t *testing.T, buf *syncBuffer) {
	t.Helper()
	logged := buf.String()
	iChunks := strings.Index(logged, `"msg":"correction delivery","stage":"chunks-sent"`)
	iCommit := strings.Index(logged, `"msg":"correction delivery","stage":"commit-sent"`)
	iArmed := strings.Index(logged, `"msg":"correction delivery","stage":"flip-armed"`)
	if iChunks < 0 || iCommit < 0 || iArmed < 0 {
		t.Fatalf("delivery marks missing (chunks@%d commit@%d flip-armed@%d) — a silent boundary "+
			"is the three-round lesson; log:\n%s", iChunks, iCommit, iArmed, logged)
	}
	if iChunks >= iCommit || iCommit >= iArmed {
		t.Fatalf("delivery marks out of order (chunks@%d commit@%d flip-armed@%d); log:\n%s",
			iChunks, iCommit, iArmed, logged)
	}
}

// replayAssertZone is the zone half of the contract, asserted BEFORE the
// release cause runs (a keystroke cause types its own character, which is
// not the correction's doing): inside the window the field keeps the word
// under ANY cause (the mandate's core promise); beyond it persistence is
// NOT claimed — the model drains to where the host's client lands and the
// loudness assertions own the contract.
func replayAssertZone(t *testing.T, cell replayCell, client *replayWezterm) {
	t.Helper()
	if cell.zone == replayZoneDrained {
		client.advance(time.Second)
		if got := string(client.field); got != replayTailConverted {
			t.Fatalf("simulated field after the drained round = %q, want %q — the client lag "+
				"reordered the commit against the queued forwards (the owner's «вижу привет — "+
				"и он удаляется»)", got, replayTailConverted)
		}

		return
	}
	client.advance(2 * replayBeyondLag)
}

// replayReleaseWiring is the fresh wiring of the timeout cell's own round:
// the short verify budget attached BEFORE the gesture (the quiet-timeout
// corpus form) so the arm reads the folded 30 ms.
func replayReleaseWiring(t *testing.T) *replayRowEnv {
	t.Helper()
	buf := captureLogsLevel(t, slog.LevelDebug) // the quiet verify row is DEBUG-level
	a, sink := wiredActor()
	teachTerminalClass(t, a)
	a.SetOptions(flipOnOptions())
	cfg := reloadCfg(300, "shift+ctrl_r")
	cfg.Timeouts.VerifyWaitMs = 30
	a.AttachConfig(&reloadSource{cfg: cfg}) // the NEXT event folds the budget
	probe := &switchProbe{}
	a.SetSwitcher(probe.switcher)
	clock := &flipClock{}
	a.UseFlipTimer(clock.arm)
	client := &replayWezterm{field: []rune(replayTailWord), lag: replayDrainedLag}
	a.UseLevel2Sleeper(client.tick)
	sink.forwardHook = func(call forwardCall) { client.forward(call) }
	sink.commitHook = client.commit

	return &replayRowEnv{a: a, sink: sink, probe: probe, clock: clock, client: client, buf: buf}
}

// replayRunTimeoutCause is the timeout cell's own round (the folded 30 ms
// budget of replayReleaseWiring): the quiet timeout lands loudly and
// releases nothing — the verify outcome is not a release cause (plan 10-17).
func replayRunTimeoutCause(t *testing.T) {
	t.Helper()
	env := replayReleaseWiring(t)
	typeWord(env.a, replayTailWord)
	tapShift(env.a)
	tapShift(env.a)
	env.a.ExpiryAt(expiryAfterWindow)
	if !waitLogUntil(t, env.buf, `"outcome":"timeout"`, 1, 2*time.Second) {
		t.Fatalf("the quiet timeout never landed; log:\n%s", env.buf.String())
	}
	refuteModeRecord(t, env.buf)
	if got := env.probe.targets(); len(got) != 0 {
		t.Fatalf("SetGlobalEngine after the verify timeout = %q, want none — the timeout is "+
			"not a release cause anymore", got)
	}
}

// replayRunReleaseCause drives the cell's release cause and asserts its
// immediate contract: the honest causes (the confirming push, the next
// keystroke) release the flip exactly once; the grace fire and the verify
// outcomes release NOTHING anymore.
func replayRunReleaseCause(t *testing.T, cell replayCell, env *replayRowEnv) {
	t.Helper()
	switch cell.release {
	case releasePush:
		line := replayTailConverted
		env.a.HandleSurroundingText(line, runeLen(line), runeLen(line)) // the confirming push
		env.client.field = []rune(line)
		if got := countModeRecords(env.buf); got != 1 {
			t.Fatalf("mode records after the confirming push = %d, want exactly 1; log:\n%s",
				got, env.buf.String())
		}
		if got := env.probe.targets(); !slices.Equal(got, []string{engine.NameRU}) {
			t.Errorf("switch targets = %q, want exactly [%s]", got, engine.NameRU)
		}
	case releaseKeystroke:
		env.a.HandleKey(engine.EngineEvent{Keyval: 'x'}) // released first, consumed in RU
		if got := countModeRecords(env.buf); got != 1 {
			t.Fatalf("mode records after the next keystroke = %d, want exactly 1; log:\n%s",
				got, env.buf.String())
		}
		if got := env.probe.targets(); !slices.Equal(got, []string{engine.NameRU}) {
			t.Errorf("switch targets = %q, want exactly [%s]", got, engine.NameRU)
		}
	case releaseGrace:
		// The round-4 rule: the terminal flip arms NO grace deadline — the
		// slot waits for the next keystroke or a confirming push, and a
		// grace fire is a stale no-op.
		if armed := env.clock.armedDurations(); len(armed) != 0 {
			t.Fatalf("deferred flips armed a grace deadline = %v, want none — the flip must not "+
				"release inside the delivery window (the round-2 hypothesis)", armed)
		}
		env.clock.fire()
		refuteModeRecord(t, env.buf)
		if got := env.probe.targets(); len(got) != 0 {
			t.Fatalf("SetGlobalEngine after a grace fire = %q, want none — the grace is not a "+
				"release cause anymore", got)
		}
	case releaseTimeout:
		replayRunTimeoutCause(t)
	case releaseMismatch:
		// WR-05: the first mismatching answer re-requires, the second
		// concludes — and the concluding mismatch must NOT release the
		// terminal flip.
		env.a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))
		env.a.HandleSurroundingText(mismatchLine, runeLen(mismatchLine), runeLen(mismatchLine))
		if !strings.Contains(env.buf.String(), `"msg":"correction verify","outcome":"mismatch"`) {
			t.Fatalf("the mismatch never concluded; log:\n%s", env.buf.String())
		}
		refuteModeRecord(t, env.buf)
		if got := env.probe.targets(); len(got) != 0 {
			t.Fatalf("SetGlobalEngine after the mismatch = %q, want none — the mismatch is not "+
				"a release cause anymore", got)
		}
	}
}

// replayAssertReleaseMarks is the release side of the loudness contract:
// the released flip carries its delivery mark after the cause, and the
// INFO journal carries outcomes only, never field content (D-20/D-21).
// The timeout cell captures at DEBUG (the quiet verify row is
// DEBUG-level by contract) — the content-free pin is the INFO cells' and
// is skipped there.
func replayAssertReleaseMarks(t *testing.T, cell replayCell, buf *syncBuffer) {
	t.Helper()
	if cell.release == releaseTimeout {
		return
	}
	logged := buf.String()
	iArmed := strings.Index(logged, `"msg":"correction delivery","stage":"flip-armed"`)
	iReleased := strings.Index(logged, `"msg":"correction delivery","stage":"flip-released"`)
	switch cell.release {
	case releasePush, releaseKeystroke:
		if iReleased < 0 || iReleased < iArmed {
			t.Fatalf("the released flip carries no delivery mark after its cause "+
				"(flip-released@%d, flip-armed@%d); log:\n%s", iReleased, iArmed, logged)
		}
	default:
		if iReleased >= 0 {
			t.Fatalf("a flip-released mark exists without an honest release cause; log:\n%s", logged)
		}
	}
	if strings.Contains(logged, wordEN) || strings.Contains(logged, wordRU) {
		t.Errorf("the delivery marks carry field content (D-20/D-21); log:\n%s", logged)
	}
}

// TestReplayWezterm1921_ZonedDeliveryMatrix is the zoned matrix of the
// 19:21:26 row: client lag (inside vs beyond the machine's delivery
// window) × flip release cause (confirming push / next keystroke / grace
// fire / verify timeout / verify mismatch). Contracts: (а) inside the
// window the field holds «привет » under ANY release cause; (б) beyond it
// persistence is NOT claimed but every boundary is loud (delivery marks,
// no silent paths); (в) at NO matrix point does SetGlobalEngine fire
// between the commit and the flip's honest release cause — a grace release
// or a verify-outcome release is a failure.
func TestReplayWezterm1921_ZonedDeliveryMatrix(t *testing.T) {
	for _, cell := range replayMatrixCells() {
		t.Run(cell.name, func(t *testing.T) {
			env := replayRow(t, cell.lag)

			replayAssertRoundForm(t, env)
			replayAssertDeliveryMarks(t, env.buf)
			replayAssertZone(t, cell, env.client)
			replayRunReleaseCause(t, cell, env)
			replayAssertReleaseMarks(t, cell, env.buf)
		})
	}
}

// TestReplayWezterm1921_FlipFalseByteStable is pin (г): with the off gate
// (the stand configs' world) the whole matrix stays byte-stable — zero
// armed flips, zero switch traffic, the frozen burst+commit+verify-arm
// shape, in BOTH lag zones.
func TestReplayWezterm1921_FlipFalseByteStable(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	teachTerminalClass(t, a)
	a.SetOptions(session.Options{FlipAfterCorrection: false})
	clock := &flipClock{}
	a.UseFlipTimer(clock.arm)
	client := &replayWezterm{field: []rune(replayTailWord), lag: replayDrainedLag}
	a.UseLevel2Sleeper(client.tick)
	sink.forwardHook = func(call forwardCall) { client.forward(call) }
	sink.commitHook = client.commit
	opsBefore := len(sink.opLog()) // the teaching round's own records stay out

	typeWord(a, replayTailWord)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if armed := clock.armedDurations(); len(armed) != 0 {
		t.Fatalf("flip_after_correction=false armed %d deferred flips, want 0", len(armed))
	}
	if got := countModeRecords(buf); got != 0 {
		t.Errorf("mode records with the gate off = %d, want 0; log:\n%s", got, buf.String())
	}
	want := append(burstOps(len([]rune(replayTailWord)), replayTailConverted), "require")
	if got := sink.opLog()[opsBefore:]; !slices.Equal(got, want) {
		t.Errorf("op log = %q, want the byte-stable frozen shape %q", got, want)
	}
	client.advance(time.Second)
	if got := string(client.field); got != replayTailConverted {
		t.Errorf("simulated field = %q, want %q — the off-gate world keeps the delivery contract",
			got, replayTailConverted)
	}
}

// generationBumpHook returns the forwardHook that moves the input-context
// generation after the FIRST forward — the FocusOut form plus the re-mint
// onto a fresh engine object — the mid-flow cell's driver (the flap-pair
// corpus's hook form).
func generationBumpHook(a *session.Actor, fresh *fakeSink, client *replayWezterm) func(forwardCall) {
	var landed sync.Once
	sent := atomic.Int32{}

	return func(call forwardCall) {
		if sent.Add(1) == 1 {
			landed.Do(func() {
				go func() {
					a.HandleLifecycle(engine.LifecycleFocusOut) // the generation moves mid-flow
					a.AttachEngine(fresh)                       // the context re-mints onto a new object
				}()
			})
		}
		client.forward(call)
	}
}

// TestReplayWezterm1921_ContextChangeAfterFirstForwardCommits is cell (д)
// of the matrix: the context generation moves AFTER the first forward (the
// FocusOut form between chunks — a real window switch) and the commit must
// NOT be cancelled: the replacement lands on the emitter snapshot CAPTURED
// at the round's start (the fresh re-minted engine stays byte-silent) with
// a loud mid-flow mark. RED truth: the current chunk-boundary discipline
// («criterion 4») stops the round WITHOUT the commit — the deletions are
// on the wire and the replacement never follows (the «удаление без
// замены» the mandate forbids). The multi-chunk form (termWord20) is the
// reachable shape of the boundary window on every tree.
func TestReplayWezterm1921_ContextChangeAfterFirstForwardCommits(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	teachTerminalClass(t, a)
	a.SetOptions(flipOnOptions())
	probe := &switchProbe{}
	a.SetSwitcher(probe.switcher)
	fresh := &fakeSink{}
	client := &replayWezterm{field: []rune(termWord20), lag: replayDrainedLag}
	a.UseLevel2Sleeper(client.tick)
	sink.forwardHook = generationBumpHook(a, fresh, client)
	sink.commitHook = client.commit

	typeWord(a, termWord20)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if n := len(sink.forwardCalls()); n != 2*len([]rune(termWord20)) {
		t.Fatalf("forwards = %d, want %d — a mid-flow generation change must not stop the burst after "+
			"the first destructive key", n, 2*len([]rune(termWord20)))
	}
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != termWord20Converted {
		t.Fatalf("commits = %q, want exactly one %q on the captured snapshot — a mid-flow context "+
			"change must never leave the field deleted without its replacement", texts, termWord20Converted)
	}
	if got := fresh.commitTexts(); len(got) != 0 {
		t.Errorf("commits into the RE-MINTED engine = %q, want none — the commit rides the captured "+
			"snapshot, never the current emitter()", got)
	}
	logged := buf.String()
	if !strings.Contains(logged, `"msg":"correction delivery","stage":"context-changed-midflow"`) {
		t.Errorf("the loud mid-flow boundary is missing — a silent context change inside the window "+
			"is the quiet-loss class; log:\n%s", logged)
	}
	if !strings.Contains(logged, `"msg":"correction","outcome":"done","changed":true`) {
		t.Errorf("the round never logged its completion; log:\n%s", logged)
	}
	client.advance(time.Second)
	if got := string(client.field); got != termWord20Converted {
		t.Errorf("simulated field = %q, want %q", got, termWord20Converted)
	}
	// The flip still waits for its honest cause.
	if got := countModeRecords(buf); got != 0 {
		t.Errorf("mode records before the release = %d, want 0; log:\n%s", got, logged)
	}
	if got := probe.targets(); len(got) != 0 {
		t.Errorf("SetGlobalEngine before any release cause = %q, want none", got)
	}
}

// TestReplayWezterm1921_ContextChangeBeforeFirstForwardStops pins the
// preserved abort: a generation change landing BEFORE the first forward is
// still a free stop — zero forwards, zero commits, the context-changed
// boundary (nothing has been destroyed, so nothing needs replacing). RED
// truth: the pre-first-forward discipline does not exist on HEAD — the
// burst replays anyway.
func TestReplayWezterm1921_ContextChangeBeforeFirstForwardStops(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	teachTerminalClass(t, a)
	client := &replayWezterm{field: []rune(termWord20), lag: replayDrainedLag}
	var landed sync.Once
	var sleeps atomic.Int64
	a.UseLevel2Sleeper(func(d time.Duration) {
		if sleeps.Add(1) == 1 {
			landed.Do(func() { go a.HandleLifecycle(engine.LifecycleFocusOut) })
		}
		client.tick(d)
	})
	sink.forwardHook = func(call forwardCall) { client.forward(call) }
	sink.commitHook = client.commit

	typeWord(a, termWord20)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := len(sink.forwardCalls()); got != 0 {
		t.Errorf("forwards = %d, want 0 — a generation change before the first forward is a free "+
			"stop, the burst must never start", got)
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Errorf("commits = %q, want none — nothing was destroyed, nothing commits", texts)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"context-changed"`) {
		t.Errorf("the round-stop INFO boundary is missing; log:\n%s", buf.String())
	}
	if got := strings.Count(buf.String(), `"outcome":"done"`); got != 0 {
		t.Errorf("a stopped round logged a done outcome; log:\n%s", buf.String())
	}
}

// TestReplayWezterm1921_BusyGuardCoversDeliveryWindow pins T-10-17-01: the
// WR-01 re-entrancy guard covers the WHOLE machine window including the
// delivery wait — a second correction entry that lands while the round is
// parked inside it refuses with the correction-busy slug instead of
// building a second plan over the same un-replaced buffer.
func TestReplayWezterm1921_BusyGuardCoversDeliveryWindow(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	teachTerminalClass(t, a)

	// Park the machine inside the delivery window: the word is 7 BackSpace
	// PAIRS (round-6: 14 pacing sleeps — press and release quanta per key),
	// the window's own wait is the 15th sleep (the ONLY burst-adjacent sleep
	// that releases the actor mutex — parking on a mid-burst sleep would
	// hold the lock and starve the second entry instead of refusing it).
	started := make(chan struct{})
	release := make(chan struct{})
	var sleeps atomic.Int64
	a.UseLevel2Sleeper(func(time.Duration) {
		if sleeps.Add(1) == int64(2*len([]rune(replayTailWord))+1) {
			close(started)
			<-release
		}
	})

	burstDone := make(chan struct{})
	go func() {
		defer close(burstDone)
		typeWord(a, replayTailWord)
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
	}()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("the machine never parked inside its delivery window — the bounded wait does not " +
			"exist on this tree")
	}

	second := make(chan struct{})
	go func() {
		close(second)
		a.CorrectNow() // the ctl entry: must refuse against the in-flight window
	}()
	<-second
	if !waitLogUntil(t, buf, `"reason":"correction-busy"`, 1, 2*time.Second) {
		t.Fatalf("the second entry never refused — WR-01 does not cover the delivery window; log:\n%s",
			buf.String())
	}
	close(release)

	select {
	case <-burstDone:
	case <-time.After(5 * time.Second):
		t.Fatal("the burst never settled — deadlock in the busy interleave")
	}
	if texts := sink.commitTexts(); len(texts) != 1 || texts[0] != replayTailConverted {
		t.Errorf("commits = %q, want exactly one %q — a second destructive burst must never ride the "+
			"window", texts, replayTailConverted)
	}
	if got := len(sink.forwardCalls()); got != 2*len([]rune(replayTailWord)) {
		t.Errorf("forwards = %d, want exactly %d — one burst, never two", got, 2*len([]rune(replayTailWord)))
	}
}

// TestReplayHost1936_SecondSampleKeepsWord replays the second host sample
// (2026-10-14 19:36:00 — done +76 ms, a signature line, not a full block):
// the same consumer at the row's faster client lag, one zone (inside the
// machine's delivery window), the signature pins of the form — the done
// row, the chunked terminal write, the deferred flip released exactly once
// by the confirming push, and the delivery marks — with the field holding
// «привет » at the end.
func TestReplayHost1936_SecondSampleKeepsWord(t *testing.T) {
	const hostSampleLag = 76 * time.Millisecond // the row's own done latency class
	env := replayRow(t, hostSampleLag)

	logged := env.buf.String()
	if !strings.Contains(logged, `"msg":"correction","outcome":"done","changed":true`) {
		t.Fatalf("the replayed round never completed; log:\n%s", logged)
	}
	// The terminal write's shape: seven BackSpace forwards, ONE commit, the
	// verify-after arm — never the level-1 pair. Compared as the op log's
	// SUFFIX: the teaching round's own records stay out.
	want := append(burstOps(len([]rune(replayTailWord)), replayTailConverted), "require")
	got := env.sink.opLog()
	if len(got) < len(want) || !slices.Equal(got[len(got)-len(want):], want) {
		t.Fatalf("op log tail = %q, want the terminal write %q", got, want)
	}
	if got := countModeRecords(env.buf); got != 0 {
		t.Fatalf("mode records before the release = %d, want 0 — the flip is deferred; log:\n%s",
			got, logged)
	}
	if got := env.probe.targets(); len(got) != 0 {
		t.Fatalf("SetGlobalEngine before the release cause = %q, want none", got)
	}
	replayAssertDeliveryMarks(t, env.buf)

	line := replayTailConverted
	env.a.HandleSurroundingText(line, runeLen(line), runeLen(line)) // the confirming push
	env.client.field = []rune(line)
	env.client.advance(time.Second)

	if got := string(env.client.field); got != replayTailConverted {
		t.Fatalf("simulated field after the row = %q, want %q — the second host sample loses the "+
			"word too", got, replayTailConverted)
	}
	if got := countModeRecords(env.buf); got != 1 {
		t.Errorf("mode records after the confirming push = %d, want exactly 1; log:\n%s",
			got, env.buf.String())
	}
	if got := env.probe.targets(); !slices.Equal(got, []string{engine.NameRU}) {
		t.Errorf("switch targets = %q, want exactly [%s]", got, engine.NameRU)
	}
}

// drainPark is the parked-round handle of parkRoundInDrainWindow: release
// unblocks the machine's window wait; wait fails the cell if the round never
// settles after it.
type drainPark struct {
	release func()
	wait    func(t *testing.T)
}

// parkRoundInDrainWindow drives one full terminal round (round types the
// stretch and the gesture) and parks the machine INSIDE the drain window:
// the word is 7 forwards (7 pacing sleeps), the window's own wait is the
// 8th sleep — it blocks until release. The shared interleave rig of the
// CR-01/CR-02 cells.
func parkRoundInDrainWindow(t *testing.T, a *session.Actor, round func()) *drainPark {
	t.Helper()
	started := make(chan struct{})
	release := make(chan struct{})
	var sleeps atomic.Int64
	a.UseLevel2Sleeper(func(time.Duration) {
		// Round-6 pair arithmetic: the 7-key burst sleeps 2×7 pacing quanta
		// (press + release per key, mutex HELD); the settle window's own
		// sleep is the 15th — the only one that releases the mutex, so the
		// park lets the corpus's lifecycle/combo traffic actually land.
		if sleeps.Add(1) == int64(2*len([]rune(replayTailWord))+1) {
			close(started)
			<-release
		}
	})
	done := make(chan struct{})
	go func() {
		defer close(done)
		round()
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("the machine never parked inside its drain window — the bounded wait does not " +
			"exist on this tree")
	}

	return &drainPark{
		release: func() { close(release) },
		wait: func(t *testing.T) {
			t.Helper()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("the round never settled — deadlock in the drain-window interleave")
			}
		},
	}
}

// TestReplayWezterm1921_GestureFlipDuringWindowParksNotSwitches is corpus
// cell (α) of the round-4 review (CR-01, 10-REVIEW): a gesture-driven flip
// arriving INSIDE the drain-settle window must not SetGlobalEngine between
// the forwards and the still-pending commit — the re-mint orphans the commit
// on the captured snapshot (the E3 verdict, «удаление без замены», the one
// outcome the ADR-008 amendment (6) forbids). The review's most probable
// entry drives the cell: the word-layout combo pressed mid-window refuses by
// the WR-01 guard (correction-busy) and its own settleCombo reaches flipTo —
// the gate parks the flip into the deferred slot with the terminal release
// discipline: the mode switch happens AFTER the correction settles, on the
// next keystroke or a confirming push. Never a silent drop. RED truth:
// flipTo executes immediately — the switch lands mid-window, before the
// commit, with no park boundary in the journal.
func TestReplayWezterm1921_GestureFlipDuringWindowParksNotSwitches(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()
	teachTerminalClass(t, a)
	// The gesture flip is isolated: the correction arms nothing itself.
	a.SetOptions(session.Options{FlipAfterCorrection: false})
	probe := &switchProbe{}
	a.SetSwitcher(probe.switcher)

	park := parkRoundInDrainWindow(t, a, func() {
		typeWord(a, replayTailWord)
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
	})

	pressComboDefault(a) // the combo arrives INSIDE the window: busy refusal → settleCombo → flipTo

	if got := probe.targets(); len(got) != 0 {
		t.Fatalf("SetGlobalEngine fired inside the drain window = %q, want none — a gesture flip must "+
			"park until the correction settles (CR-01, the E3 loss); log:\n%s", got, buf.String())
	}
	if got := countModeRecords(buf); got != 0 {
		t.Fatalf("mode records inside the window = %d, want 0; log:\n%s", got, buf.String())
	}
	park.release()
	park.wait(t)

	// The correction itself landed — the deletion-without-replacement is
	// absent at the unit seam: exactly one burst, the replacement first.
	if got := len(sink.forwardCalls()); got != 2*len([]rune(replayTailWord)) {
		t.Errorf("forwards = %d, want exactly %d — one burst", got, 2*len([]rune(replayTailWord)))
	}
	texts := sink.commitTexts()
	if len(texts) < 1 || texts[0] != replayTailConverted {
		t.Errorf("commits = %q, want the replacement %q first", texts, replayTailConverted)
	}
	// The parked flip still waits for its honest cause — nothing released it.
	if got := probe.targets(); len(got) != 0 {
		t.Errorf("SetGlobalEngine before any release cause = %q, want none", got)
	}
	if !strings.Contains(buf.String(), `"msg":"correction delivery","stage":"flip-parked"`) {
		t.Errorf("the park boundary is missing from the journal; log:\n%s", buf.String())
	}

	// The honest release cause: the next keystroke — the mode switch happens
	// NOW, after the correction landed (the owner's intent preserved).
	typeWord(a, "x")
	if got := probe.targets(); !slices.Equal(got, []string{engine.NameRU}) {
		t.Errorf("switch targets after the release keystroke = %q, want exactly [%s] — the parked "+
			"gesture flip must still happen (never a silent drop)", got, engine.NameRU)
	}
	if !strings.Contains(buf.String(), `"stage":"flip-released","reason":"keystroke"`) {
		t.Errorf("the released mark is missing; log:\n%s", buf.String())
	}
}

// TestReplayWezterm1921_KeysDuringWindowRideAfterCommit is corpus cell (β)
// of the round-4 review (CR-02): keystrokes processed inside the unlocked
// drain-settle window hit the wire BEFORE the pending commit — the
// daemon-side inversion («xпривет » instead of «привет x», deterministic on
// every interleave of fast typing after the gesture). The hold path parks
// typed-layout input (printables and BackSpace — the keys the buffer tracks)
// for the machine's settle and re-emits the transits AFTER the commit:
// wire order = deletes → commit → held keys, FIFO. RED truth: the keystroke
// processes at arrival (verdict false, ibus forwards it immediately) and no
// daemon emission of it exists at all.
func TestReplayWezterm1921_KeysDuringWindowRideAfterCommit(t *testing.T) {
	a, sink := wiredActor()
	teachTerminalClass(t, a)
	a.SetOptions(session.Options{FlipAfterCorrection: false})

	park := parkRoundInDrainWindow(t, a, func() {
		typeWord(a, replayTailWord)
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
	})

	// Two keystrokes land inside the window — the routine interleave of fast
	// typing continued right after the gesture (the ~100-150 ms key pitch vs
	// the 150 ms window).
	v1 := make(chan bool, 1)
	v2 := make(chan bool, 1)
	go func() { v1 <- a.HandleKey(engine.EngineEvent{Keyval: 'x', Keycode: 45}) }()
	time.Sleep(20 * time.Millisecond) // the first key parks and blocks on the machine's settle
	go func() { v2 <- a.HandleKey(engine.EngineEvent{Keyval: 'y', Keycode: 21}) }()
	time.Sleep(20 * time.Millisecond) // the second key parks behind it (FIFO)

	park.release()
	park.wait(t)
	for i, ch := range []chan bool{v1, v2} {
		select {
		case c := <-ch:
			if !c {
				t.Errorf("held keystroke #%d arrival verdict = false, want true (consumed — the daemon "+
					"owes the wire emission)", i+1)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("held keystroke #%d never returned — the hold path deadlocked", i+1)
		}
	}

	// Wire order: the machine's deletes, the commit (and the verify-after
	// arm riding the same exit) first — the held keys strictly after, FIFO.
	// Never «xпривет».
	want := append(burstOps(len([]rune(replayTailWord)), replayTailConverted),
		"require", "forward(120,45,0)", "forward(121,21,0)")
	got := sink.opLog()
	if len(got) < len(want) || !slices.Equal(got[len(got)-len(want):], want) {
		t.Errorf("op log tail = %q, want the held-key order %q (deletes → commit → held keys)", got, want)
	}
}
