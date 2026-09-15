package config

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

// defaultDebounce coalesces an editor's save burst (write-temp + rename +
// chained writes) into one re-parse — 200 ms, inside the STACK 150–300 ms
// band.
const defaultDebounce = 200 * time.Millisecond

// EventSource is the filesystem-event surface the watcher consumes — an
// interface so the corpus drives the loop with synthetic events headlessly
// (no real inotify; the Wave 0 gap note of the phase research).
type EventSource interface {
	Events() <-chan fsnotify.Event
	Errors() <-chan error
	Add(dir string) error
	Close() error
}

// inotifySource adapts the real fsnotify.Watcher to EventSource.
type inotifySource struct{ w *fsnotify.Watcher }

func (s inotifySource) Events() <-chan fsnotify.Event { return s.w.Events }
func (s inotifySource) Errors() <-chan error          { return s.w.Errors }

func (s inotifySource) Add(dir string) error {
	if err := s.w.Add(dir); err != nil {
		return fmt.Errorf("fsnotify add %s: %w", dir, err)
	}

	return nil
}

func (s inotifySource) Close() error {
	if err := s.w.Close(); err != nil {
		return fmt.Errorf("fsnotify close: %w", err)
	}

	return nil
}

// watchOpts carry the constructor's knobs; the zero value is filled with
// the daemon defaults (real inotify source, Load parser, 200 ms debounce).
type watchOpts struct {
	debounce time.Duration
	load     func(path string) (*Config, error)
	source   func() (EventSource, error)
}

// Option customizes NewWatcher (the test seams ride the same surface).
type Option func(*watchOpts)

// WithDebounce overrides the reload debounce.
func WithDebounce(d time.Duration) Option {
	return func(o *watchOpts) { o.debounce = d }
}

// WithLoader replaces the parse function (counted loads, scripted results).
func WithLoader(load func(path string) (*Config, error)) Option {
	return func(o *watchOpts) { o.load = load }
}

// WithSource replaces the real inotify source (synthetic event feeding).
func WithSource(makeSource func() (EventSource, error)) Option {
	return func(o *watchOpts) { o.source = makeSource }
}

// Watcher publishes the last-good config: fsnotify watches the config's
// DIRECTORY (atomic-rename editors swap the file's inode — watching the
// path itself loses the file), events for other names are ignored, and a
// debounce timer coalesces bursts into one full re-parse. A valid parse is
// stored atomically; a rejected one keeps the last-good snapshot serving
// (D-32) with the error exposed for status.
type Watcher struct {
	path     string
	fileName string
	opts     watchOpts
	reloadMu sync.Mutex // serializes debounce re-parses
	current  atomic.Pointer[Config]
	lastErr  atomic.Pointer[error]
	timer    *time.Timer // event-loop goroutine only (the armTimer discipline)
}

// NewWatcher loads the config at path once (a refusal fails construction —
// an explicit -config must yield a working config) and then watches the
// file's directory until ctx is done.
func NewWatcher(ctx context.Context, path string, opts ...Option) (*Watcher, error) {
	o := watchOpts{
		debounce: defaultDebounce,
		load:     Load,
		source: func() (EventSource, error) {
			fw, err := fsnotify.NewWatcher()
			if err != nil {
				return nil, err //nolint:wrapcheck // plain construction error, wrapped by the caller
			}

			return inotifySource{w: fw}, nil
		},
	}
	for _, opt := range opts {
		opt(&o)
	}

	w := &Watcher{path: path, fileName: filepath.Base(path), opts: o}
	cfg, err := o.load(path)
	if err != nil {
		return nil, fmt.Errorf("initial config load: %w", err)
	}
	w.current.Store(cfg)

	src, err := o.source()
	if err != nil {
		return nil, fmt.Errorf("create fsnotify watcher: %w", err)
	}
	// The DIRECTORY, not the file: vim/kwrite/gedit save via write-temp +
	// rename, so the watched inode changes under a file watch (STACK).
	if err := src.Add(filepath.Dir(path)); err != nil {
		_ = src.Close()

		return nil, fmt.Errorf("watch config directory: %w", err)
	}

	go w.loop(ctx, src)

	return w, nil
}

// Snapshot returns the effective configuration — a value copy of the
// last-good load. The signature is the pinned producer contract: the
// consumer of plan 03-04 matches interface{ Snapshot() config.Config }.
func (w *Watcher) Snapshot() Config {
	return *w.current.Load()
}

// LastError returns the error of the last rejected reload (nil while the
// served snapshot is valid) — the status surface of plan 03-06.
func (w *Watcher) LastError() error {
	if p := w.lastErr.Load(); p != nil {
		return *p
	}

	return nil
}

// Reload forces an immediate synchronous re-parse of the config document —
// the control surface's reload (INST-02): a valid document is published
// (the reply names the applied change), a rejected one keeps the last-good
// snapshot serving (D-32) with the error returned and exposed through
// LastError.
func (w *Watcher) Reload() (string, error) {
	return "", nil
}

// loop drains the event source until ctx is done; every goroutine has an
// exit condition (go-ultimate concurrency hygiene).
func (w *Watcher) loop(ctx context.Context, src EventSource) {
	defer func() { _ = src.Close() }() // teardown: the close error carries no signal

	for {
		select {
		case <-ctx.Done():
			if w.timer != nil {
				w.timer.Stop()
			}

			return
		case err, ok := <-src.Errors():
			if !ok {
				return
			}
			slog.Warn("config watch error", "error", err)
		case ev, ok := <-src.Events():
			if !ok {
				return
			}
			w.handle(ev)
		}
	}
}

// handle re-arms the debounce timer for events of the watched file only:
// other names in the directory and non-trigger operations are ignored, a
// fresh event inside the window recharges the timer (the armTimer
// discipline of actor.go:572-577 — stop the old AfterFunc, arm a new one).
func (w *Watcher) handle(ev fsnotify.Event) {
	if filepath.Base(ev.Name) != w.fileName {
		return
	}
	if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
		return
	}
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(w.opts.debounce, w.reload)
}

// reload re-parses the config: a valid document replaces the served
// snapshot atomically and logs the applied window (the reload-application
// record the live cases gate on — the same content-free shape as the
// startup "config loaded"); a rejected one WARNs ("config reload
// rejected", D-32) and leaves the last-good in place with the error
// exposed through LastError — status only, never section content
// (T-03-02-04).
func (w *Watcher) reload() {
	w.reloadMu.Lock()
	defer w.reloadMu.Unlock()

	cfg, err := w.opts.load(w.path)
	if err != nil {
		slog.Warn("config reload rejected", "error", err)
		w.lastErr.Store(&err)

		return
	}
	w.current.Store(cfg)
	slog.Info("config reloaded", "tap_window_ms", cfg.Timeouts.TapWindowMs)
	var none error
	w.lastErr.Store(&none)
}
