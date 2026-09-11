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
