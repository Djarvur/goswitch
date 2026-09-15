package config_test

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/Djarvur/goswitch/internal/config"
)

// fakeSource feeds synthetic events without inotify — the EventSource seam
// (Wave 0 gap note: the corpus runs headless, no real filesystem watch).
type fakeSource struct {
	events     chan fsnotify.Event
	errors     chan error
	addedDirs  []string
	closeGuard sync.Mutex
	closed     bool
}

func newFakeSource() *fakeSource {
	return &fakeSource{
		events: make(chan fsnotify.Event, 16),
		errors: make(chan error, 16),
	}
}

func (s *fakeSource) Events() <-chan fsnotify.Event { return s.events }
func (s *fakeSource) Errors() <-chan error          { return s.errors }

func (s *fakeSource) Add(dir string) error {
	s.closeGuard.Lock()
	defer s.closeGuard.Unlock()
	s.addedDirs = append(s.addedDirs, dir)

	return nil
}

func (s *fakeSource) Close() error {
	s.closeGuard.Lock()
	defer s.closeGuard.Unlock()
	s.closed = true

	return nil
}

func (s *fakeSource) isClosed() bool {
	s.closeGuard.Lock()
	defer s.closeGuard.Unlock()

	return s.closed
}

// send delivers one synthetic event without blocking (buffered channel).
func (s *fakeSource) send(name string, op fsnotify.Op) {
	s.events <- fsnotify.Event{Name: name, Op: op}
}

// watchFixture wires a watcher over the fake source with a scripted
// loader; the path never touches the real filesystem.
type watchFixture struct {
	src   *fakeSource
	loads atomic.Int32
}

// newWatcher starts a watcher on a fast debounce over synthetic plumbing.
func newWatchFixture(t *testing.T, results func(call int32) (*config.Config, error)) (*watchFixture, *config.Watcher) {
	t.Helper()

	fx := &watchFixture{src: newFakeSource()}
	loader := func(path string) (*config.Config, error) {
		return results(fx.loads.Add(1))
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	w, err := config.NewWatcher(ctx, "/tmp/goswitch-test/goswitch.yaml",
		config.WithSource(func() (config.EventSource, error) { return fx.src, nil }),
		config.WithLoader(loader),
		config.WithDebounce(10*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}

	return fx, w
}

// docWith returns a valid config with the given tap window.
func docWith(tapWindowMs int) *config.Config {
	c := validDoc()
	c.Timeouts.TapWindowMs = tapWindowMs

	return &c
}

// waitUntil polls cond until it holds or the 2 s deadline passes; it
// returns cond's final value.
func waitUntil(cond func() bool) bool {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(2 * time.Millisecond)
	}

	return cond()
}

// syncBuffer is a mutex-guarded log sink (the actor_test idiom).
type syncBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

// captureLogs redirects the process default logger into a guarded buffer.
// The capturing test cannot run parallel: the default logger is
// process-global.
func captureLogs(t *testing.T) *syncBuffer {
	t.Helper()

	buf := &syncBuffer{}
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, nil)))
	t.Cleanup(func() { slog.SetDefault(old) })

	return buf
}

// TestWatch_ValidRewritePublishes pins the happy reload path (CONF-02):
// a Write event for the watched file, after the debounce, publishes the
// NEW config through Snapshot.
func TestWatch_ValidRewritePublishes(t *testing.T) {
	t.Parallel()

	fx, w := newWatchFixture(t, func(call int32) (*config.Config, error) {
		if call == 1 {
			return docWith(300), nil
		}

		return docWith(450), nil
	})
	if got := w.Snapshot().Timeouts.TapWindowMs; got != 300 {
		t.Fatalf("initial snapshot tap_window_ms = %d, want 300", got)
	}

	fx.src.send("/tmp/goswitch-test/goswitch.yaml", fsnotify.Write)

	if !waitUntil(func() bool { return w.Snapshot().Timeouts.TapWindowMs == 450 }) {
		t.Fatalf("snapshot tap_window_ms = %d after rewrite, want 450", w.Snapshot().Timeouts.TapWindowMs)
	}
	if err := w.LastError(); err != nil {
		t.Errorf("LastError = %v after a valid reload, want nil", err)
	}
}

// TestWatch_InvalidRewriteKeepsLastGood pins D-32: a broken edit WARNs
// ("config reload rejected"), the last-good snapshot keeps serving and
// LastError carries the rejection for the status surface (plan 03-06).
func TestWatch_InvalidRewriteKeepsLastGood(t *testing.T) {
	buf := captureLogs(t)

	fx, w := newWatchFixture(t, func(call int32) (*config.Config, error) {
		if call == 1 {
			return docWith(300), nil
		}

		return nil, errors.New("decode config: line 2: unknown field timeots")
	})
	fx.src.send("/tmp/goswitch-test/goswitch.yaml", fsnotify.Write)

	if !waitUntil(func() bool { return w.LastError() != nil }) {
		t.Fatal("LastError never set after an invalid rewrite")
	}
	if got := w.Snapshot().Timeouts.TapWindowMs; got != 300 {
		t.Errorf("snapshot tap_window_ms = %d after a rejected reload, want the last-good 300", got)
	}
	if !strings.Contains(buf.String(), `"msg":"config reload rejected"`) {
		t.Errorf("log %q misses the WARN record config reload rejected", buf.String())
	}
}

// TestWatch_IgnoresIrrelevantEvents pins the filters: events for other
// file names in the watched directory and non-trigger operations on the
// watched file (Chmod) never cause a re-parse.
func TestWatch_IgnoresIrrelevantEvents(t *testing.T) {
	t.Parallel()

	fx, _ := newWatchFixture(t, func(int32) (*config.Config, error) {
		return docWith(300), nil
	})

	fx.src.send("/tmp/goswitch-test/other.yaml", fsnotify.Write)
	fx.src.send("/tmp/goswitch-test/goswitch.yaml", fsnotify.Chmod)

	time.Sleep(80 * time.Millisecond) // 8x debounce: any re-parse would land
	if got := fx.loads.Load(); got != 1 {
		t.Errorf("re-parses = %d after irrelevant events, want 1 (the initial load only)", got)
	}
}

// TestWatch_DebounceCoalesces pins the debounce contract: three events in
// one window coalesce into exactly ONE re-parse (the counter pin).
func TestWatch_DebounceCoalesces(t *testing.T) {
	t.Parallel()

	fx, _ := newWatchFixture(t, func(int32) (*config.Config, error) {
		return docWith(300), nil
	})

	for range 3 {
		fx.src.send("/tmp/goswitch-test/goswitch.yaml", fsnotify.Write)
	}

	if !waitUntil(func() bool { return fx.loads.Load() >= 2 }) {
		t.Fatalf("re-parse never fired: loads = %d", fx.loads.Load())
	}
	time.Sleep(80 * time.Millisecond) // 8x debounce: a second re-parse would land
	if got := fx.loads.Load(); got != 2 {
		t.Errorf("re-parses after a 3-event burst = %d, want exactly 2 (initial + one coalesced)", got)
	}
}

// TestWatch_RemoveAndRenameTrigger pins the atomic-rename coverage: Remove
// and Rename events for the watched name each trigger a re-parse (an
// editor's save is a Rename inside the same directory).
func TestWatch_RemoveAndRenameTrigger(t *testing.T) {
	t.Parallel()

	fx, _ := newWatchFixture(t, func(int32) (*config.Config, error) {
		return docWith(300), nil
	})

	fx.src.send("/tmp/goswitch-test/goswitch.yaml", fsnotify.Remove)
	if !waitUntil(func() bool { return fx.loads.Load() == 2 }) {
		t.Fatalf("loads after Remove = %d, want 2", fx.loads.Load())
	}

	fx.src.send("/tmp/goswitch-test/goswitch.yaml", fsnotify.Rename)
	if !waitUntil(func() bool { return fx.loads.Load() == 3 }) {
		t.Fatalf("loads after Rename = %d, want 3", fx.loads.Load())
	}
}

// TestWatch_SourceErrorWarns pins the Errors-channel drain: a source error
// is a WARN, never a crash of the watch loop.
func TestWatch_SourceErrorWarns(t *testing.T) {
	buf := captureLogs(t)

	fx, _ := newWatchFixture(t, func(int32) (*config.Config, error) {
		return docWith(300), nil
	})

	fx.src.errors <- errors.New("inotify queue overflow")

	if !waitUntil(func() bool { return strings.Contains(buf.String(), `"msg":"config watch error"`) }) {
		t.Error("log misses the WARN record config watch error")
	}
}

// TestWatch_WatchesDirectoryNotFile pins the STACK pattern: the watch is
// added on filepath.Dir (atomic-rename editors swap the inode).
func TestWatch_WatchesDirectoryNotFile(t *testing.T) {
	t.Parallel()

	fx, _ := newWatchFixture(t, func(int32) (*config.Config, error) {
		return docWith(300), nil
	})

	fx.src.closeGuard.Lock()
	defer fx.src.closeGuard.Unlock()
	if len(fx.src.addedDirs) != 1 || fx.src.addedDirs[0] != "/tmp/goswitch-test" {
		t.Errorf("watched dirs = %v, want exactly the config's directory /tmp/goswitch-test", fx.src.addedDirs)
	}
}

// TestWatch_ContextCancelClosesSource pins goroutine hygiene: canceling
// the context stops the loop and closes the source — no leak under -race.
func TestWatch_ContextCancelClosesSource(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	fx := &fakeSource{
		events: make(chan fsnotify.Event, 16),
		errors: make(chan error, 16),
	}
	loader := func(string) (*config.Config, error) { return docWith(300), nil }

	w, err := config.NewWatcher(ctx, "/tmp/goswitch-test/goswitch.yaml",
		config.WithSource(func() (config.EventSource, error) { return fx, nil }),
		config.WithLoader(loader),
		config.WithDebounce(10*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	_ = w

	cancel()

	if !waitUntil(fx.isClosed) {
		t.Error("source not closed after ctx cancel — the loop goroutine leaked")
	}
}

// TestWatch_InitialLoadFailureRefuses pins the start contract: a config
// that cannot load fails construction — an explicit -config must yield a
// working config, never silent defaults.
func TestWatch_InitialLoadFailureRefuses(t *testing.T) {
	t.Parallel()

	_, err := config.NewWatcher(context.Background(), "/tmp/goswitch-test/goswitch.yaml",
		config.WithSource(func() (config.EventSource, error) { return newFakeSource(), nil }),
		config.WithLoader(func(string) (*config.Config, error) {
			return nil, errors.New("decode config: unknown field")
		}),
	)
	if err == nil {
		t.Fatal("NewWatcher succeeded on an unloadable config, want a construction refusal")
	}
}
