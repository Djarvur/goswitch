//nolint:testpackage // drives the unexported retry knobs — the sanctioned in-package corpus seam
package activate

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

// Corpus literals (goconst): the gsettings shapes repeat across tables.
const (
	sourcesOwnerSingle = `[('ibus', 'goswitch-en')]`
	sourcesOwnerMid    = `[('xkb', 'us'), ('ibus', 'goswitch-ru')]`
	sourcesForeignXKB  = `[('xkb', 'us'), ('xkb', 'ru')]`
	sourcesForeignIBus = `[('ibus', 'mozc'), ('ibus', 'goswitch-en')]`
	sourcesGarbage     = `garbage`
	current0           = "uint32 0"
	current1           = "uint32 1"
	currentOutOfRange  = "uint32 5"
	currentNegative    = "uint32 -1"
	currentAlpha       = "uint32 abc"

	callEN = "ibus engine goswitch-en"
	callRU = "ibus engine goswitch-ru"
)

var (
	errGSettingsDown = errors.New("gsettings is down")
	errIBusDown      = errors.New("ibus refused")
)

// fakeRunner records every subprocess IfOwned makes and answers from a
// per-case script — the seam standing in for the real gsettings/ibus
// binaries (the install fakeRunner pattern, without env/stdin).
type fakeRunner struct {
	mu    sync.Mutex
	calls []string // rendered "name args..." per call, in order
	reply func(name string, args []string) ([]byte, error)
}

func (f *fakeRunner) run(_ context.Context, name string, args []string) ([]byte, error) {
	f.mu.Lock()
	f.calls = append(f.calls, name+" "+strings.Join(args, " "))
	f.mu.Unlock()

	return f.reply(name, args)
}

// ibusCalls copies the recorded ibus calls under the guard.
func (f *fakeRunner) ibusCalls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	var out []string
	for _, c := range f.calls {
		if strings.HasPrefix(c, binIbus+" ") {
			out = append(out, c)
		}
	}

	return out
}

// totalCalls reports how many subprocesses the fake served.
func (f *fakeRunner) totalCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return len(f.calls)
}

// stubDesktop answers the two gsettings reads from fixed strings and the
// ibus activation per attempt: ibusErrs[i] answers attempt i (nil =
// success); the last entry repeats for further attempts.
func stubDesktop(sources, current string, ibusErrs ...error) func(string, []string) ([]byte, error) {
	attempt := 0

	return func(name string, args []string) ([]byte, error) {
		if name == binGSettings {
			if slices.Contains(args, keySources) {
				return []byte(sources), nil
			}

			return []byte(current), nil
		}
		if len(ibusErrs) == 0 {
			return []byte(""), nil
		}
		i := min(attempt, len(ibusErrs)-1)
		attempt++

		return []byte(""), ibusErrs[i]
	}
}

// stubGSettingsDown fails exactly one of the two gsettings reads — the
// headless / missing-binary degradation path.
func stubGSettingsDown(failSources bool) func(string, []string) ([]byte, error) {
	return func(name string, args []string) ([]byte, error) {
		if name != binGSettings {
			return []byte(""), nil
		}
		isSources := slices.Contains(args, keySources)
		if isSources == failSources {
			return nil, errGSettingsDown
		}
		if isSources {
			return []byte(sourcesOwnerSingle), nil
		}

		return []byte(current0), nil
	}
}

// withRetry overrides the retry knobs for one test; t.Cleanup restores
// them. The package's tests never run in parallel (the knobs are process
// globals), keeping -race deterministic.
func withRetry(t *testing.T, attempts int, delay time.Duration) {
	t.Helper()

	oldAttempts, oldDelay := activationAttempts, activationRetryDelay
	if attempts > 0 {
		activationAttempts = attempts
	}
	if delay > 0 {
		activationRetryDelay = delay
	}
	t.Cleanup(func() {
		activationAttempts, activationRetryDelay = oldAttempts, oldDelay
	})
}

// runIfOwned drives IfOwned through the fake on the background context and
// returns the elapsed wall time and the recorded ibus calls.
func runIfOwned(t *testing.T, reply func(string, []string) ([]byte, error),
) (elapsed time.Duration, ibusCalls []string) {
	t.Helper()

	f := &fakeRunner{reply: reply}
	start := time.Now()
	IfOwned(context.Background(), f.run)
	elapsed = time.Since(start)

	return elapsed, f.ibusCalls()
}

// assertIBus compares the recorded ibus calls with the expected list.
func assertIBus(t *testing.T, got, want []string) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Errorf("ibus calls = %q, want %q", got, want)
	}
}

// TestIfOwnedActivatesOwned pins the owned path: the CURRENT input source
// carries a goswitch engine, so exactly one `ibus engine <name>` call goes
// out with the full current name (not always goswitch-en).
func TestIfOwnedActivatesOwned(t *testing.T) {
	tests := []struct {
		name    string
		sources string
		current string
		want    []string
	}{
		{
			name:    "single source",
			sources: sourcesOwnerSingle,
			current: current0,
			want:    []string{callEN},
		},
		{
			name:    "goswitch mid-list current",
			sources: sourcesOwnerMid,
			current: current1,
			want:    []string{callRU},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			withRetry(t, 0, time.Millisecond)

			_, got := runIfOwned(t, stubDesktop(tc.sources, tc.current))

			assertIBus(t, got, tc.want)
		})
	}
}

// TestIfOwnedForeignCurrentSkipped pins the no-hijack guard: a non-goswitch
// current source never receives an activation, even when goswitch sits in
// the list elsewhere.
func TestIfOwnedForeignCurrentSkipped(t *testing.T) {
	tests := []struct {
		name    string
		sources string
		current string
	}{
		{name: "xkb current", sources: sourcesForeignXKB, current: current0},
		{name: "foreign ibus engine current", sources: sourcesForeignIBus, current: current0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			withRetry(t, 0, time.Millisecond)

			_, got := runIfOwned(t, stubDesktop(tc.sources, tc.current))

			assertIBus(t, got, nil)
		})
	}
}

// TestIfOwnedMalformedSkipped pins the strict parsing: any deviation from
// the expected shapes (out-of-range index, garbage or empty sources,
// negative/non-numeric/empty current) skips activation — no panic, no call.
func TestIfOwnedMalformedSkipped(t *testing.T) {
	tests := []struct {
		name    string
		sources string
		current string
	}{
		{name: "current out of range", sources: sourcesForeignXKB, current: currentOutOfRange},
		{name: "garbage sources", sources: sourcesGarbage, current: current0},
		{name: "empty sources", sources: `[]`, current: current0},
		{name: "negative current", sources: sourcesOwnerSingle, current: currentNegative},
		{name: "non-numeric current", sources: sourcesOwnerSingle, current: currentAlpha},
		{name: "empty current", sources: sourcesOwnerSingle, current: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			withRetry(t, 0, time.Millisecond)

			_, got := runIfOwned(t, stubDesktop(tc.sources, tc.current))

			assertIBus(t, got, nil)
		})
	}
}

// TestIfOwnedGSettingsFailureSkipped pins the degradation path: a failed
// gsettings read (headless CI, missing binary) skips activation and returns
// normally — the daemon keeps serving.
func TestIfOwnedGSettingsFailureSkipped(t *testing.T) {
	tests := []struct {
		name        string
		failSources bool
	}{
		{name: "sources read fails", failSources: true},
		{name: "current read fails", failSources: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			withRetry(t, 0, time.Millisecond)

			_, got := runIfOwned(t, stubGSettingsDown(tc.failSources))

			assertIBus(t, got, nil)
		})
	}
}

// TestIfOwnedActivationRetry pins the bounded retry: every failure warns
// after exactly activationAttempts calls and returns normally; a flaky
// first attempt succeeds on the retry.
func TestIfOwnedActivationRetry(t *testing.T) {
	t.Run("every attempt fails", func(t *testing.T) {
		const attempts = 3
		withRetry(t, attempts, time.Millisecond)

		_, got := runIfOwned(t, stubDesktop(sourcesOwnerSingle, current0, errIBusDown))

		assertIBus(t, got, []string{callEN, callEN, callEN})
	})

	t.Run("first attempt flaky", func(t *testing.T) {
		withRetry(t, 0, time.Millisecond)

		_, got := runIfOwned(t, stubDesktop(sourcesOwnerSingle, current0, errIBusDown, nil))

		assertIBus(t, got, []string{callEN, callEN})
	})
}

// TestIfOwnedCancelledContext pins the prompt return: a pre-cancelled
// context makes no subprocess calls and never sleeps through the retry
// delay (the default 500 ms delay stays in force — the wall bound proves
// the loop never sleeps past cancellation).
func TestIfOwnedCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	f := &fakeRunner{reply: stubDesktop(sourcesOwnerSingle, current0)}
	start := time.Now()
	IfOwned(ctx, f.run)
	elapsed := time.Since(start)

	if elapsed > 400*time.Millisecond {
		t.Errorf("IfOwned with a pre-cancelled context took %s — it slept past cancellation", elapsed)
	}
	if total := f.totalCalls(); total != 0 {
		t.Errorf("subprocess calls = %d, want 0", total)
	}
}

// TestParseSourceTuples pins the canonical strict parser of the GNOME
// input-sources list (05-02): the real gsettings output parses into typed
// tuples in order, and any deviation — a garbage remainder between tuples,
// unbalanced brackets, an empty list — is an error, never a partial parse.
func TestParseSourceTuples(t *testing.T) {
	t.Run("real gsettings output parses in order", func(t *testing.T) {
		tuples, err := ParseSourceTuples(`[('xkb', 'us'), ('xkb', 'ru')]`)
		if err != nil {
			t.Fatalf("ParseSourceTuples() err = %v, want nil", err)
		}
		want := []SourceTuple{{Kind: "xkb", ID: "us"}, {Kind: "xkb", ID: "ru"}}
		if !slices.Equal(tuples, want) {
			t.Errorf("tuples = %v, want %v", tuples, want)
		}
	})

	t.Run("garbage remainder between tuples is an error", func(t *testing.T) {
		tuples, err := ParseSourceTuples(`[('xkb', 'us') junk]`)
		if err == nil {
			t.Fatalf("ParseSourceTuples() = %v with nil error — want the remainder-check refusal", tuples)
		}
	})

	t.Run("unbalanced input is an error", func(t *testing.T) {
		if _, err := ParseSourceTuples(`('xkb', 'us')`); err == nil {
			t.Fatal("ParseSourceTuples() err = nil — want the unbalanced-input refusal")
		}
	})

	t.Run("empty list is an error", func(t *testing.T) {
		if _, err := ParseSourceTuples(`[]`); err == nil {
			t.Fatal("ParseSourceTuples() err = nil — want the empty-list refusal")
		}
	})
}
