package config_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Djarvur/goswitch/internal/config"
)

// Fixture refusals — package sentinels (the err113 discipline of the
// strict lint: no dynamically created error values).
var (
	errFixtureDecode  = errors.New("decode config: unknown field")
	errFixtureNoStamp = errors.New("stat: no stamp scripted for the path")
)

// fakeStat scripts the stat seam: per-path fingerprints, absence = the
// stat error. bump moves a path's stamp the way a real edit moves mtime.
type fakeStat struct {
	mu     sync.Mutex
	stamps map[string]config.FileStamp
}

func newFakeStat(paths ...string) *fakeStat {
	s := &fakeStat{stamps: map[string]config.FileStamp{}}
	for _, p := range paths {
		s.stamps[p] = config.FileStamp{ModTime: time.Unix(0, 0), Size: 10}
	}

	return s
}

func (s *fakeStat) stat(path string) (config.FileStamp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.stamps[path]
	if !ok {
		return config.FileStamp{}, errFixtureNoStamp
	}

	return st, nil
}

func (s *fakeStat) set(path string, st config.FileStamp) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stamps[path] = st
}

func (s *fakeStat) bump(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := s.stamps[path]
	st.ModTime = st.ModTime.Add(time.Second)
	st.Size++
	s.stamps[path] = st
}

// pollFixture wires a watcher over the fake stat with a scripted loader
// and an owned tick channel; the path never touches the real filesystem.
type pollFixture struct {
	stat   *fakeStat
	ticks  chan time.Time
	loads  atomic.Int32
	cancel context.CancelFunc
}

func newPollFixture(
	t *testing.T, path string, results func(call int32) (*config.Config, error),
) (*pollFixture, *config.Watcher) {
	t.Helper()

	fx := &pollFixture{
		stat:  newFakeStat(path),
		ticks: make(chan time.Time, 8),
	}
	loader := func(path string) (*config.Config, error) {
		return results(fx.loads.Add(1))
	}

	ctx, cancel := context.WithCancel(context.Background())
	fx.cancel = cancel
	t.Cleanup(cancel)

	w, err := config.NewWatcher(ctx, path,
		config.WithStat(fx.stat.stat),
		config.WithLoader(loader),
		config.WithTicks(fx.ticks),
	)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}

	return fx, w
}

// tick delivers one poll tick.
func (fx *pollFixture) tick() { fx.ticks <- time.Now() }

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

	n, _ := b.buf.Write(p) // strings.Builder.Write never fails

	return n, nil
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

// TestPoll_ChangePublishes pins the happy reload path (CONF-02, the
// polling form): a stat whose fingerprint moved publishes the NEW config
// through Snapshot on the next tick.
func TestPoll_ChangePublishes(t *testing.T) {
	t.Parallel()

	fx, w := newPollFixture(t, "/tmp/goswitch-test/goswitch.yaml", func(call int32) (*config.Config, error) {
		if call == 1 {
			return docWith(300), nil
		}

		return docWith(450), nil
	})
	if got := w.Snapshot().Timeouts.TapWindowMs; got != 300 {
		t.Fatalf("initial snapshot tap_window_ms = %d, want 300", got)
	}

	fx.stat.bump("/tmp/goswitch-test/goswitch.yaml")
	fx.tick()

	if !waitUntil(func() bool { return w.Snapshot().Timeouts.TapWindowMs == 450 }) {
		t.Fatalf("snapshot tap_window_ms = %d after the change, want 450", w.Snapshot().Timeouts.TapWindowMs)
	}
	if err := w.LastError(); err != nil {
		t.Errorf("LastError = %v after a valid reload, want nil", err)
	}
}

// TestPoll_UnchangedStampSkips pins the poll's economy: ticks over an
// unchanged fingerprint never parse — the stat compare is the whole cost
// of a quiet config.
func TestPoll_UnchangedStampSkips(t *testing.T) {
	t.Parallel()

	fx, _ := newPollFixture(t, "/tmp/goswitch-test/goswitch.yaml", func(int32) (*config.Config, error) {
		return docWith(300), nil
	})

	for range 3 {
		fx.tick()
	}
	time.Sleep(50 * time.Millisecond) // a late re-parse would land

	if got := fx.loads.Load(); got != 1 {
		t.Errorf("loads after 3 quiet ticks = %d, want 1 (the initial load only)", got)
	}
}

// TestPoll_InvalidChangeKeepsLastGoodAndRecovers pins D-32 end to end: a
// rejected fingerprint consumes itself (ONE WARN, no per-tick chorus),
// the last-good keeps serving, and the repaired document applies on the
// next tick without a restart.
func TestPoll_InvalidChangeKeepsLastGoodAndRecovers(t *testing.T) {
	buf := captureLogs(t)

	fx, w := newPollFixture(t, "/tmp/goswitch-test/goswitch.yaml", func(call int32) (*config.Config, error) {
		switch call {
		case 1:
			return docWith(300), nil
		case 2:
			return nil, errFixtureDecode
		default:
			return docWith(500), nil
		}
	})

	fx.stat.bump("/tmp/goswitch-test/goswitch.yaml")
	fx.tick()
	if !waitUntil(func() bool { return w.LastError() != nil }) {
		t.Fatal("LastError never set after the rejected change")
	}
	if got := w.Snapshot().Timeouts.TapWindowMs; got != 300 {
		t.Errorf("snapshot tap_window_ms = %d after the rejection, want the last-good 300", got)
	}
	if !strings.Contains(buf.String(), `"msg":"config reload rejected"`) {
		t.Errorf("log %q misses the WARN record config reload rejected", buf.String())
	}

	// The rejected fingerprint is consumed: quiet ticks add no attempts.
	before := fx.loads.Load()
	fx.tick()
	time.Sleep(30 * time.Millisecond)
	if got := fx.loads.Load(); got != before {
		t.Errorf("loads after a quiet tick over the rejected fingerprint = %d, want %d", got, before)
	}

	// The repair applies without a restart.
	fx.stat.bump("/tmp/goswitch-test/goswitch.yaml")
	fx.tick()
	if !waitUntil(func() bool { return w.Snapshot().Timeouts.TapWindowMs == 500 }) {
		t.Fatalf("snapshot tap_window_ms = %d after the repair, want 500", w.Snapshot().Timeouts.TapWindowMs)
	}
	if err := w.LastError(); err != nil {
		t.Errorf("LastError = %v after the repaired reload, want nil", err)
	}
}

// TestPoll_AbsentFileQuietThenAppears pins the adopt path's lifecycle:
// a document that is absent at construction (the green defaults start)
// polls quietly, and its later appearance applies on the next tick.
func TestPoll_AbsentFileQuietThenAppears(t *testing.T) {
	t.Parallel()

	path := "/tmp/goswitch-test/late.yaml"
	fx := &pollFixture{
		stat:  newFakeStat(), // the path is absent: no stamp registered
		ticks: make(chan time.Time, 8),
	}
	loader := func(string) (*config.Config, error) {
		fx.loads.Add(1)

		return docWith(370), nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	w, err := config.NewWatcher(ctx, path,
		config.WithStat(fx.stat.stat),
		config.WithLoader(loader),
		config.WithTicks(fx.ticks),
	)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}

	fx.tick()
	time.Sleep(30 * time.Millisecond) // an absent-document poll must stay quiet
	if got := fx.loads.Load(); got != 1 {
		t.Fatalf("loads over an absent document = %d, want 1 (the initial load only)", got)
	}

	fx.stat.set(path, config.FileStamp{ModTime: time.Unix(1, 0), Size: 10})
	fx.tick()
	if !waitUntil(func() bool { return w.Snapshot().Timeouts.TapWindowMs == 370 }) {
		t.Fatalf("snapshot tap_window_ms = %d after the document appeared, want 370", w.Snapshot().Timeouts.TapWindowMs)
	}
}

// TestWatch_ReloadRefreshesFingerprint pins the ctl-path contract on the
// poller: Reload answers for the document's current fingerprint, so the
// next tick over the SAME content adds no re-parse.
func TestWatch_ReloadRefreshesFingerprint(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "goswitch.yaml")
	if err := os.WriteFile(path, []byte(autocorrectDocYAML("2.0")), 0o600); err != nil {
		t.Fatalf("write initial config: %v", err)
	}

	ticks := make(chan time.Time, 8)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	w, err := config.NewWatcher(ctx, path,
		config.WithLoader(config.Load),
		config.WithTicks(ticks),
	)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}

	// The document changes under the served snapshot; Reload answers
	// without waiting for a tick.
	if err := os.WriteFile(path, []byte(autocorrectDocYAML("2.5")), 0o600); err != nil {
		t.Fatalf("write changed config: %v", err)
	}
	if _, err := w.Reload(); err != nil {
		t.Fatalf("Reload: %v", err)
	}
	if got := w.Snapshot().Autocorrect.TrigramMargin; got != 2.5 {
		t.Fatalf("snapshot trigram_margin after Reload = %v, want 2.5", got)
	}

	ticks <- time.Now()
	time.Sleep(50 * time.Millisecond) // a double answer would land
	if got := w.Snapshot().Autocorrect.TrigramMargin; got != 2.5 {
		t.Errorf("snapshot trigram_margin drifted after the tick: %v", got)
	}
}

// autocorrectDocYAML renders the complete document with the given
// autocorrect trigram margin — the broken-autocorrect-block corpus's
// template (the chordDocYAML idiom: full document, one value under test).
func autocorrectDocYAML(margin string) string {
	return `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
  mode_switch_chord: super+space
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
autocorrect:
  enabled: true
  apps_blocklist: ["` + appGedit + `"]
  min_word_len: 4
  trigram_margin: ` + margin + `
  trigram_floor: 1.0
`
}

// TestWatch_BrokenAutocorrectBlockKeepsLastGood pins the D-32/D-33
// propagation to the new section WITHOUT any watcher code: a broken edit
// of the autocorrect block (margin below the floor) invalidates the WHOLE
// document — the reload is rejected, the last-good snapshot keeps serving,
// the WARN "config reload rejected" lands — and the repaired document
// applies without a restart. The REAL parser runs (config.Load), so the
// property under test is the schema's, not a scripted loader's.
func TestWatch_BrokenAutocorrectBlockKeepsLastGood(t *testing.T) {
	buf := captureLogs(t)

	path := filepath.Join(t.TempDir(), "goswitch.yaml")
	if err := os.WriteFile(path, []byte(autocorrectDocYAML("2.0")), 0o600); err != nil {
		t.Fatalf("write initial config: %v", err)
	}

	ticks := make(chan time.Time, 8)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	w, err := config.NewWatcher(ctx, path,
		config.WithLoader(config.Load),
		config.WithTicks(ticks),
	)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	if got := w.Snapshot().Autocorrect.TrigramMargin; got != 2.0 {
		t.Fatalf("initial snapshot trigram_margin = %v, want 2.0", got)
	}

	// The broken edit: margin below the floor invalidates the whole file.
	if err := os.WriteFile(path, []byte(autocorrectDocYAML("0.5")), 0o600); err != nil {
		t.Fatalf("write broken config: %v", err)
	}
	ticks <- time.Now()
	if !waitUntil(func() bool { return w.LastError() != nil }) {
		t.Fatal("LastError never set after a broken autocorrect edit")
	}
	if got := w.Snapshot().Autocorrect.TrigramMargin; got != 2.0 {
		t.Errorf("snapshot trigram_margin = %v after the rejection, want the last-good 2.0", got)
	}
	if !strings.Contains(buf.String(), `"msg":"config reload rejected"`) {
		t.Errorf("log %q misses the WARN record config reload rejected", buf.String())
	}

	// The repair applies without a restart.
	if err := os.WriteFile(path, []byte(autocorrectDocYAML("2.5")), 0o600); err != nil {
		t.Fatalf("write repaired config: %v", err)
	}
	ticks <- time.Now()
	if !waitUntil(func() bool { return w.Snapshot().Autocorrect.TrigramMargin == 2.5 }) {
		t.Fatalf("snapshot trigram_margin = %v after the repair, want 2.5", w.Snapshot().Autocorrect.TrigramMargin)
	}
	if err := w.LastError(); err != nil {
		t.Errorf("LastError = %v after the repaired reload, want nil", err)
	}
}

// blocklistDocYAML renders the complete document with the given
// autocorrect.apps_blocklist pattern — the broken-pattern reload corpus's
// template (the autocorrectDocYAML idiom: full document, one value under
// test).
func blocklistDocYAML(pattern string) string {
	return `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
  mode_switch_chord: super+space
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
autocorrect:
  enabled: true
  apps_blocklist: ["` + pattern + `"]
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
`
}

// TestWatch_BrokenBlocklistPatternKeepsLastGood pins the D-32/D-33
// propagation to the blocklist patterns WITHOUT any watcher code: a
// reload carrying a non-compilable pattern invalidates the WHOLE document
// (the compile gate runs on Load) — the reload is rejected, the last-good
// snapshot keeps serving, the WARN "config reload rejected" lands — and
// the repaired pattern applies without a restart. The REAL parser runs
// (config.Load), so the property under test is the schema's.
func TestWatch_BrokenBlocklistPatternKeepsLastGood(t *testing.T) {
	buf := captureLogs(t)

	path := filepath.Join(t.TempDir(), "goswitch.yaml")
	if err := os.WriteFile(path, []byte(blocklistDocYAML("gedit")), 0o600); err != nil {
		t.Fatalf("write initial config: %v", err)
	}

	ticks := make(chan time.Time, 8)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	w, err := config.NewWatcher(ctx, path,
		config.WithLoader(config.Load),
		config.WithTicks(ticks),
	)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	if got := w.Snapshot().Autocorrect.AppsBlocklist; len(got) != 1 || got[0] != "gedit" {
		t.Fatalf("initial snapshot blocklist = %v, want [gedit]", got)
	}

	// The broken edit: a non-compilable pattern invalidates the whole file.
	if err := os.WriteFile(path, []byte(blocklistDocYAML("[")), 0o600); err != nil {
		t.Fatalf("write broken config: %v", err)
	}
	ticks <- time.Now()
	if !waitUntil(func() bool { return w.LastError() != nil }) {
		t.Fatal("LastError never set after a broken blocklist edit")
	}
	if got := w.Snapshot().Autocorrect.AppsBlocklist; len(got) != 1 || got[0] != "gedit" {
		t.Errorf("snapshot blocklist = %v after the rejection, want the last-good [gedit]", got)
	}
	if !strings.Contains(buf.String(), `"msg":"config reload rejected"`) {
		t.Errorf("log %q misses the WARN record config reload rejected", buf.String())
	}

	// The repair applies without a restart.
	if err := os.WriteFile(path, []byte(blocklistDocYAML("chrom")), 0o600); err != nil {
		t.Fatalf("write repaired config: %v", err)
	}
	ticks <- time.Now()
	if !waitUntil(func() bool {
		got := w.Snapshot().Autocorrect.AppsBlocklist

		return len(got) == 1 && got[0] == "chrom"
	}) {
		t.Fatalf("snapshot blocklist = %v after the repair, want [chrom]", w.Snapshot().Autocorrect.AppsBlocklist)
	}
	if err := w.LastError(); err != nil {
		t.Errorf("LastError = %v after the repaired reload, want nil", err)
	}
}

// TestWatch_A11yUnknownKeyKeepsLastGood pins the D-32/D-33 propagation to
// the a11y section WITHOUT any watcher code (the 07-02 blocklist-case
// precedent, retargeted by the 2026-10-06 owner revision): a reload
// carrying an unknown key inside a11y invalidates the WHOLE document —
// the reload is rejected, the last-good snapshot keeps serving, the WARN
// "config reload rejected" lands — and the repaired document applies
// without a restart. The REAL parser runs (config.Load), so the property
// under test is the schema's.
func TestWatch_A11yUnknownKeyKeepsLastGood(t *testing.T) {
	buf := captureLogs(t)

	path := filepath.Join(t.TempDir(), "goswitch.yaml")
	if err := os.WriteFile(path, []byte(a11yDocYAML("  enabled: true\n")), 0o600); err != nil {
		t.Fatalf("write initial config: %v", err)
	}

	ticks := make(chan time.Time, 8)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	w, err := config.NewWatcher(ctx, path,
		config.WithLoader(config.Load),
		config.WithTicks(ticks),
	)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	if got := w.Snapshot().A11y.EffectiveEnabled(); !got {
		t.Fatalf("initial snapshot a11y effective switch = %v, want true (default ON)", got)
	}

	// The broken edit: an unknown key inside a11y invalidates the whole file.
	if err := os.WriteFile(path, []byte(a11yDocYAML("  enabled: true\n  per_app_levels: wat\n")), 0o600); err != nil {
		t.Fatalf("write broken config: %v", err)
	}
	ticks <- time.Now()
	if !waitUntil(func() bool { return w.LastError() != nil }) {
		t.Fatal("LastError never set after a broken a11y edit")
	}
	if got := w.Snapshot().A11y.EffectiveEnabled(); !got {
		t.Errorf("snapshot a11y effective switch = %v after the rejection, want the last-good true", got)
	}
	if !strings.Contains(buf.String(), `"msg":"config reload rejected"`) {
		t.Errorf("log %q misses the WARN record config reload rejected", buf.String())
	}

	// The repair applies without a restart: the only OFF shape.
	if err := os.WriteFile(path, []byte(a11yDocYAML("  enabled: false\n")), 0o600); err != nil {
		t.Fatalf("write repaired config: %v", err)
	}
	ticks <- time.Now()
	if !waitUntil(func() bool { return !w.Snapshot().A11y.EffectiveEnabled() }) {
		t.Fatalf("snapshot a11y effective switch = %v after the repair, want false",
			w.Snapshot().A11y.EffectiveEnabled())
	}
	if err := w.LastError(); err != nil {
		t.Errorf("LastError = %v after the repaired reload, want nil", err)
	}
}

// TestWatch_ContextCancelStopsPolling pins goroutine hygiene: after ctx
// cancellation the loop is gone — ticks over a changed fingerprint
// produce no work, no panic, no leak under -race.
func TestWatch_ContextCancelStopsPolling(t *testing.T) {
	t.Parallel()

	fx, _ := newPollFixture(t, "/tmp/goswitch-test/goswitch.yaml", func(int32) (*config.Config, error) {
		return docWith(300), nil
	})

	fx.cancel()
	time.Sleep(20 * time.Millisecond) // the loop's exit lands

	fx.stat.bump("/tmp/goswitch-test/goswitch.yaml")
	fx.tick()
	time.Sleep(30 * time.Millisecond) // a live loop would re-parse here

	if got := fx.loads.Load(); got != 1 {
		t.Errorf("loads after cancel + changed fingerprint + tick = %d, want 1 (the initial load)", got)
	}
}

// TestWatch_InitialLoadFailureRefuses pins the start contract: a config
// that cannot load fails construction — an explicit -config must yield a
// working config, never silent defaults.
func TestWatch_InitialLoadFailureRefuses(t *testing.T) {
	t.Parallel()

	ticks := make(chan time.Time, 1)
	_, err := config.NewWatcher(context.Background(), "/tmp/goswitch-test/goswitch.yaml",
		config.WithStat(newFakeStat("/tmp/goswitch-test/goswitch.yaml").stat),
		config.WithLoader(func(string) (*config.Config, error) {
			return nil, errFixtureDecode
		}),
		config.WithTicks(ticks),
	)
	if err == nil {
		t.Fatal("NewWatcher succeeded on an unloadable config, want a construction refusal")
	}
}
