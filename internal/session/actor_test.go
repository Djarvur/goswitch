package session_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Djarvur/goswitch/engine"
	"github.com/Djarvur/goswitch/internal/clipboard"
	"github.com/Djarvur/goswitch/internal/config"
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

// countActions counts decision records in the captured log.
func countActions(buf *syncBuffer) int {
	return strings.Count(buf.String(), `"msg":"action"`)
}

// Corpus words of the correction pipeline (the 02-01 corpus pair — named
// once, goconst).
const (
	wordEN = "ghbdtn"
	wordRU = "привет"
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
// clipboard rung's subprocess log).
type fakeSink struct {
	mu          sync.Mutex
	requires    int
	deletes     []deleteCall
	commits     []string
	forwards    []forwardCall
	ops         []string
	forwardHook func(forwardCall)
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
	defer f.mu.Unlock()

	f.commits = append(f.commits, text.Text)
	f.ops = append(f.ops, "commit("+text.Text+")")
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

	if got := sink.requireCount(); got != 1 {
		t.Errorf("require calls = %d, want exactly 1 — the verify-after round, never a pre-correction one", got)
	}
	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -6, nchars: 6}) {
		t.Fatalf("deletions = %+v, want exactly one (-6,6)", calls)
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
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
		mismatch := "abc другойтекст"
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

		typeWord(a, wordEN)
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
		a.VerifyExpiry() // deterministic injection of the 100 ms deadline

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
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != wordRU+" " {
		t.Fatalf("commits = %q, want exactly one %q", texts, wordRU+" ")
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
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText(wordEN, runeLen(wordEN), runeLen(wordEN))

	// Repeat: the buffer holds привет, the field too.
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText(wordRU, runeLen(wordRU), runeLen(wordRU))

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
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText(wordEN, runeLen(wordEN), runeLen(wordEN))

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
	tapShift(a)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	mismatch := "abc другойтекст"
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

// backSpaceOp is the unified op-log entry of one level-2 Backspace replay:
// keyval 0xff08, physical keycode 14, no modifiers (RESEARCH Pattern 1 — the
// wire contract the actor must emit verbatim).
func backSpaceOp() string {
	return fmt.Sprintf("forward(%d,%d,%d)", engine.KeyBackSpace, 14, 0)
}

// backSpaceWireCall is the recorded-argument form of the same replay.
func backSpaceWireCall() forwardCall {
	return forwardCall{keyval: engine.KeyBackSpace, keycode: 14, state: 0}
}

// burstOps builds the unified op-log suffix of a level-2 correction: n
// Backspace forwards followed by the single replacement commit.
func burstOps(n int, commit string) []string {
	ops := make([]string, 0, n+1)
	for range n {
		ops = append(ops, backSpaceOp())
	}

	return append(ops, "commit("+commit+")")
}

// TestActor_Level2NoCaps pins the level-2 ladder on a client WITHOUT the
// surrounding-text bit (CORR-07, the ADR-004 degradation): verification is
// impossible, so the correction trusts the buffer plus the explicit CORR-09
// reset triggers and executes IMMEDIATELY at the Double decision — zero
// RequireSurroundingText calls, exactly six ForwardKeyEvent(BackSpace)
// replays (one per RUNE of the token, Pitfall 2) and one CommitText of the
// converted word, the burst→commit order pinned by the unified op log. This
// test supersedes the 02-03 "no-surrounding" refusal pin: level 2 IS the
// replacement for that skip.
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
	forwards := sink.forwardCalls()
	if len(forwards) != tokenRunes+tailRunes {
		t.Fatalf("forward count = %d, want %d (token+tail in RUNES, not bytes)", len(forwards), tokenRunes+tailRunes)
	}
	for i, call := range forwards {
		if want := backSpaceWireCall(); call != want {
			t.Errorf("forward %d = %+v, want %+v", i, call, want)
		}
	}
	ops := sink.opLog()
	burst := ops[len(ops)-(tokenRunes+tailRunes+1):] // the 7 forwards + the commit
	if want := burstOps(tokenRunes+tailRunes, wordEN+" "); !slices.Equal(burst, want) {
		t.Fatalf("op-log tail = %q, want %q — burst→commit order", burst, want)
	}
	if texts := sink.commitTexts(); texts[len(texts)-1] != wordEN+" " {
		t.Errorf("correction commit = %q, want %q (converted+tail)", texts[len(texts)-1], wordEN+" ")
	}
}

// TestActor_VerifyAfterLevel1 pins the post-correction check of the level-1
// ladder (ADR-003/ADR-004: DeleteSurroundingText is ack-less on 1.5.29, the
// verify-after is the compensation): the correction itself asks for a fresh
// surrounding text — the require counter grows past the pre-correction
// round — and the answer decides. A field ending with the expected
// replacement is quiet; a field that does not raises the INFO mismatch
// counter exactly once and is NEVER followed by a second correction
// ("abort, не мусорить" — no auto-repair, the residual risk stays
// documented in ADR-003).
func TestActor_VerifyAfterLevel1(t *testing.T) {
	// settled runs one complete level-1 correction and returns the actor at
	// the moment the verify-after round is pending.
	settled := func(t *testing.T) (*session.Actor, *fakeSink, *syncBuffer) {
		t.Helper()
		buf := captureLogs(t)
		a, sink := wiredActor()

		typeWord(a, wordEN)
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)
		line := "abc " + wordEN
		a.HandleSurroundingText(line, runeLen(line), runeLen(line)) // settles the pre-correction round

		return a, sink, buf
	}

	t.Run("match is quiet", func(t *testing.T) {
		a, sink, buf := settled(t)

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
		a, sink, buf := settled(t)

		// The correction never landed in the field (the plan's oracle): the
		// client reports the pre-correction text — the suffix no longer
		// matches the expected replacement.
		uncorrected := "abc " + wordEN
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
	})
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
// and back RU→EN, each switch logging the exact INFO mode record — and the
// flip is purely internal: zero calls of any kind on the sink (the XKB group
// of the session is not touched, D-01 verdict).
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
	if got := sink.requireCount(); got != 0 {
		t.Errorf("flip made %d RequireSurroundingText calls, want 0", got)
	}
	if calls := sink.deleteCalls(); len(calls) != 0 {
		t.Errorf("flip deleted %+v, want nothing", calls)
	}
	if texts := sink.commitTexts(); len(texts) != 0 {
		t.Errorf("flip committed %q, want nothing", texts)
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
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText("п2х", runeLen("п2х"), runeLen("п2х"))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -3, nchars: 3}) {
		t.Fatalf("deletions = %+v, want exactly one (-3,3) — the whole script-true token", calls)
	}
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

	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText(wordRU+"2026", runeLen(wordRU+"2026"), runeLen(wordRU+"2026"))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -10, nchars: 10}) {
		t.Fatalf("deletions = %+v, want exactly one (-10,10) — 10 runes, not 14 bytes", calls)
	}
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

	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	a.HandleSurroundingText(wordRU, runeLen(wordRU), runeLen(wordRU))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -6, nchars: 6}) {
		t.Fatalf("deletions = %+v, want exactly one (-6,6)", calls)
	}
	if texts := sink.commitTexts(); len(texts) != 7 || texts[6] != wordEN {
		t.Fatalf("RU→EN correction commits = %q, want the last one to be [%s]", texts, wordEN)
	}
}

// TestActor_MixedWordConvertsForeignRuns pins the D-16→D-23 succession
// (plan 03-01 task 2): a mixed word assembled the way the desktop produces
// it — "gfb" typed in EN (transit), then the flip, then the rest typed in
// RU (committed Cyrillic runes) — is no longer refused wholesale; the
// run-wise conversion converts ONLY the foreign Latin run: the whole token
// range is deleted (-9,9) and "паипривет" is committed, the Cyrillic run
// re-committed unchanged. The Phase 2 refusal pin (TestActor_MixedWord-
// Untouched, D-16) is superseded by this contract.
func TestActor_MixedWordConvertsForeignRuns(t *testing.T) {
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
	// deletes — and the settle commits the run-converted replacement.
	field := "gfb" + wordRU
	a.HandleSurroundingText(field, runeLen(field), runeLen(field))

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -9, nchars: 9}) {
		t.Fatalf("deletions = %+v, want exactly one (-9,9) — the whole mixed token range", calls)
	}
	texts := sink.commitTexts()
	if len(texts) != 7 || texts[6] != "паи"+wordRU {
		t.Fatalf("correction commits = %q, want the last one to be [%s]", texts, "паи"+wordRU)
	}
	if !strings.Contains(buf.String(), `"msg":"correction","outcome":"done"`) {
		t.Errorf("INFO completion record missing; log:\n%s", buf.String())
	}
}

// TestActor_PhraseMixedCorrects pins D-26 on the sink: a phrase with words
// in different layouts — "ghbdtn" typed in EN, the flip, " привет"
// committed in RU — corrects through the SAME run semantics as the mixed
// word (anchor = last letter of the phrase, foreign runs converted, own
// runs untouched); there is no separate phrase-level script semantics. The
// triple tap deletes the whole phrase (-13,13) and commits "привет привет".
func TestActor_PhraseMixedCorrects(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	typeWord(a, wordEN) // EN word: transit, buffer fed as typed
	pressSpace(a)       // the phrase separator
	flipMode(a)         // EN → RU
	typeWord(a, wordEN) // RU word: commits "привет", buffer script-true

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
	texts := sink.commitTexts()
	if len(texts) != 7 || texts[6] != wordRU+" "+wordRU {
		t.Fatalf("correction commits = %q, want the last one to be [%s]", texts, wordRU+" "+wordRU)
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

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -6, nchars: 6}) {
		t.Fatalf("deletions = %+v, want exactly one (-6,6) — the selection range, tail not touched", calls)
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
	if len(calls) != 1 || calls[0] != (deleteCall{offset: 0, nchars: 6}) {
		t.Fatalf("deletions = %+v, want exactly one (0,6) — offset 0, the range is right of the cursor", calls)
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
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != wordRU {
		t.Fatalf("commits = %q, want exactly one %q", texts, wordRU)
	}
}

// TestActor_SelectionMixedConverts pins D-23 on the selection range: the
// selected mixed text "gfbпривет" converts by the last-letter anchor — the
// foreign Latin run gfb→паи, the Cyrillic run recommitted unchanged —
// through the SAME ConvertRuns the word and phrase paths use.
func TestActor_SelectionMixedConverts(t *testing.T) {
	a, sink := wiredActor()

	field := "gfb" + wordRU
	a.HandleSurroundingText(field, runeLen(field), 0) // whole field selected
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	calls := sink.deleteCalls()
	if len(calls) != 1 || calls[0] != (deleteCall{offset: -9, nchars: 9}) {
		t.Fatalf("deletions = %+v, want exactly one (-9,9) — the whole selected range", calls)
	}
	texts := sink.commitTexts()
	if len(texts) != 1 || texts[0] != "паи"+wordRU {
		t.Fatalf("commits = %q, want exactly one [%s] — run conversion inside the selection", texts, "паи"+wordRU)
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
// delivers the verify-after MISMATCH push — the state the rung keys on (the
// client ignored the deletion; the field still holds the pre-correction
// text).
func settledWordCorrection(a *session.Actor) {
	typeWord(a, wordEN)
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)
	line := "abc " + wordEN
	a.HandleSurroundingText(line, runeLen(line), runeLen(line)) // settles the pre-round
	uncorrected := "abc " + wordEN
	a.HandleSurroundingText(uncorrected, runeLen(uncorrected), runeLen(uncorrected)) // verify-after mismatch
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
		if got := len(sink.forwardCalls()); got != len([]rune(wordEN)) {
			t.Errorf("post-reload burst replayed %d Backspaces, want %d — the new cap governs",
				got, len([]rune(wordEN)))
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
