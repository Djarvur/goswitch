package session_test

import (
	"bytes"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Djarvur/goswitch/engine"
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

// fakeSink is the test double of engine.Emitter: every emitter call is
// recorded under a mutex — the observable surface the correction pipeline
// drives (the guarded-sink style of syncBuffer).
type fakeSink struct {
	mu       sync.Mutex
	requires int
	deletes  []deleteCall
	commits  []string
}

// RequireSurroundingText records the verification request.
func (f *fakeSink) RequireSurroundingText() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.requires++
}

// DeleteSurroundingText records the ladder deletion with its exact range.
func (f *fakeSink) DeleteSurroundingText(offset int32, nchars uint32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.deletes = append(f.deletes, deleteCall{offset: offset, nchars: nchars})
}

// CommitText records the committed text payload.
func (f *fakeSink) CommitText(text engine.IBusText) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.commits = append(f.commits, text.Text)
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

// wiredActor returns an actor with the fake sink attached and the
// surrounding-text capability announced — the minimal live pipeline state.
func wiredActor() (*session.Actor, *fakeSink) {
	a := session.NewActor(farWindow)
	sink := &fakeSink{}
	a.AttachEngine(sink)
	a.HandleCapabilities(engine.CapSurroundingText)

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
	a.HandleSurroundingText("abc "+wordEN, runeLen("abc "+wordEN))

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
// execute at the Double decision itself, with no post-decision answer.
func TestActor_CachedSurroundingCorrects(t *testing.T) {
	buf := captureLogs(t)
	a, sink := wiredActor()

	for i, r := range wordEN {
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r)})
		a.HandleSurroundingText(wordEN[:i+1], uint32(i+1)) // the client's spontaneous push
		a.HandleKey(engine.EngineEvent{Keyval: uint32(r), Release: true})
	}
	tapShift(a)
	tapShift(a)
	a.ExpiryAt(expiryAfterWindow)

	if got := sink.requireCount(); got != 0 {
		t.Errorf("cache hit still asked for surrounding text %d times, want 0", got)
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
	a.HandleSurroundingText(wordEN, runeLen(wordEN))

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
		a.HandleSurroundingText(mismatch, runeLen(mismatch)) // does not end with the token

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
		a.HandleSurroundingText("abc "+wordEN, runeLen("abc "+wordEN))
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
// vocabulary: a mixed-script token (D-16), a letterless token and a client
// without the surrounding-text capability each skip with the exact INFO
// reason and never even start the verification round.
func TestActor_TokenRefusals(t *testing.T) {
	t.Run("mixed script", func(t *testing.T) {
		buf := captureLogs(t)
		a, sink := wiredActor()

		typeWord(a, "gfb"+wordRU) // letters of both scripts: D-16 silent refusal
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)

		if !strings.Contains(buf.String(), `"reason":"mixed-script"`) {
			t.Errorf("mixed-script record missing; log:\n%s", buf.String())
		}
		if got := sink.requireCount(); got != 0 {
			t.Errorf("mixed token started verification %d times, want 0", got)
		}
	})

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

	t.Run("no surrounding capability", func(t *testing.T) {
		buf := captureLogs(t)
		a := session.NewActor(farWindow)
		sink := &fakeSink{}
		a.AttachEngine(sink)
		a.HandleCapabilities(engine.CapPreeditText) // no CapSurroundingText

		typeWord(a, wordEN)
		tapShift(a)
		tapShift(a)
		a.ExpiryAt(expiryAfterWindow)

		if !strings.Contains(buf.String(), `"reason":"no-surrounding"`) {
			t.Errorf("no-surrounding record missing; log:\n%s", buf.String())
		}
		if got := sink.requireCount(); got != 0 {
			t.Errorf("no-cap client was asked for surrounding text %d times, want 0", got)
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
	a.HandleSurroundingText(wordEN, runeLen(wordEN))

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
