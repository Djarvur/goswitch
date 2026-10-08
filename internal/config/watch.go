package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// defaultPollInterval paces the config re-read (D-32 REV, the owner's
// 2026-10-08 decision): inotify is retired. The desktop's Electron apps
// exhaust the system inotify-instance budget — a daemon that loses the
// init race silently loses hot reload — and the rename-semantics watch
// (directory + debounce) existed solely to survive editors. A 2 s stat
// poll has no such failure mode, drops the fsnotify dependency, and the
// application model is unchanged: the fold still rides the next keypress
// (03-04), so the interval never lands inside a gesture.
const defaultPollInterval = 2 * time.Second

// FileStamp is the change fingerprint of the watched document: mtime plus
// size. A human-paced edit always moves both; a poll that sees an equal
// stamp skips the parse entirely.
type FileStamp struct {
	ModTime time.Time
	Size    int64
}

// stater is the stat seam: production os.Stat, the corpus scripts stamps
// (the EventSource seam's replacement — the headless corpus owns the
// clock and the fingerprints, no real filesystem timing involved).
type stater func(path string) (FileStamp, error)

// statStamp is the production stater.
func statStamp(path string) (FileStamp, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return FileStamp{}, err //nolint:wrapcheck // the caller's error paths wrap their own contexts
	}

	return FileStamp{ModTime: fi.ModTime(), Size: fi.Size()}, nil
}

// watchOpts carry the constructor's knobs; the zero value is filled with
// the daemon defaults (production stat, Load parser, 2 s poll).
type watchOpts struct {
	interval time.Duration
	load     func(path string) (*Config, error)
	stat     stater
	ticks    <-chan time.Time // the test clock; nil = the production ticker
}

// Option customizes NewWatcher (the test seams ride the same surface).
type Option func(*watchOpts)

// WithInterval overrides the poll period.
func WithInterval(d time.Duration) Option {
	return func(o *watchOpts) { o.interval = d }
}

// WithLoader replaces the parse function (counted loads, scripted results).
func WithLoader(load func(path string) (*Config, error)) Option {
	return func(o *watchOpts) { o.load = load }
}

// WithStat replaces the production stat (scripted fingerprints).
func WithStat(stat func(path string) (FileStamp, error)) Option {
	return func(o *watchOpts) { o.stat = stat }
}

// WithTicks injects the poll clock: every receive is one poll attempt.
// Production leaves it nil and owns a real ticker.
func WithTicks(ticks <-chan time.Time) Option {
	return func(o *watchOpts) { o.ticks = ticks }
}

// Watcher publishes the last-good config: a stat poll (mtime + size)
// re-reads the document when its fingerprint moves, at most once per
// interval. A valid parse is stored atomically; a rejected one keeps the
// last-good snapshot serving (D-32) with the error exposed for status —
// and the rejected fingerprint is consumed too, so one broken edit is
// ONE WARN, not a per-tick chorus.
type Watcher struct {
	path     string
	opts     watchOpts
	reloadMu sync.Mutex // serializes re-parses (poll, Reload)
	seenMu   sync.Mutex // guards seen against the poll/Reload race
	seen     FileStamp  // the fingerprint the served snapshot matches
	current  atomic.Pointer[Config]
	lastErr  atomic.Pointer[error]
}

// NewWatcher loads the config at path once (a refusal fails construction —
// an explicit -config must yield a working config) and then polls the
// document until ctx is done. An absent file at construction is legal
// (the adopt path's green defaults): the poll picks the document up when
// it appears.
func NewWatcher(ctx context.Context, path string, opts ...Option) (*Watcher, error) {
	o := watchOpts{
		interval: defaultPollInterval,
		load:     Load,
		stat:     statStamp,
	}
	for _, opt := range opts {
		opt(&o)
	}

	w := &Watcher{path: path, opts: o}
	cfg, err := o.load(path)
	if err != nil {
		return nil, fmt.Errorf("initial config load: %w", err)
	}
	w.current.Store(cfg)
	w.seen, _ = o.stat(path) // absent at start is legal: zero stamp, the poll catches the appearance

	if o.ticks != nil {
		go w.loop(ctx, o.ticks, nil)
	} else {
		t := time.NewTicker(o.interval)
		go w.loop(ctx, t.C, t.Stop)
	}

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

// ConfigPath returns the path the watcher serves and polls — the actor's
// status snapshot lifts it into the D-32 fields (INST-02).
func (w *Watcher) ConfigPath() string {
	return w.path
}

// Reload forces an immediate synchronous re-parse of the config document —
// the control surface's reload (INST-02): a valid document is published
// (the reply names the applied change), a rejected one keeps the last-good
// snapshot serving (D-32) with the error returned and exposed through
// LastError. The polled fingerprint is refreshed either way, so the next
// tick never re-parses what this call already answered for.
func (w *Watcher) Reload() (string, error) {
	w.reloadMu.Lock()
	defer w.reloadMu.Unlock()

	reply, err := w.reparse()
	w.refreshSeen()

	return reply, err
}

// loop drains the poll clock until ctx is done; every goroutine has an
// exit condition (go-ultimate concurrency hygiene).
func (w *Watcher) loop(ctx context.Context, ticks <-chan time.Time, stop func()) {
	if stop != nil {
		defer stop()
	}

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ticks:
			if !ok {
				return
			}
			w.poll()
		}
	}
}

// poll is one tick: a stat whose fingerprint moved since the served
// snapshot triggers exactly one re-parse; an equal stamp, a stat error
// (an absent document between edits is normal for the adopt path) and an
// in-flight Reload are all quiet no-ops.
func (w *Watcher) poll() {
	w.seenMu.Lock()
	seen := w.seen
	w.seenMu.Unlock()

	stamp, err := w.opts.stat(w.path)
	if err != nil || stamp == seen {
		return
	}

	w.reloadMu.Lock()
	defer w.reloadMu.Unlock()

	// Reload may have answered for this change while the stat ran.
	w.seenMu.Lock()
	unchanged := w.seen == stamp
	w.seenMu.Unlock()
	if unchanged {
		return
	}

	_, _ = w.reparse() // the outcome is observable state, not a return value here
	w.remember(stamp)
}

// remember publishes the fingerprint of the document the current snapshot
// answers for.
func (w *Watcher) remember(stamp FileStamp) {
	w.seenMu.Lock()
	defer w.seenMu.Unlock()

	w.seen = stamp
}

// refreshSeen re-stats and publishes the fingerprint under the reload
// lock's caller — Reload's twin of remember (the file may not stat, e.g.
// the document vanished between the ctl call and the read: the poll picks
// it up on its next tick).
func (w *Watcher) refreshSeen() {
	if stamp, err := w.opts.stat(w.path); err == nil {
		w.remember(stamp)
	}
}

// reparse is the ONE reload core: a valid document replaces the served
// snapshot atomically and logs the applied window (the reload-application
// record the live cases gate on — the same content-free shape as the
// startup "config loaded"); a rejected one WARNs ("config reload
// rejected", D-32) and leaves the last-good in place with the error
// exposed through LastError — status only, never section content
// (T-03-02-04). The caller holds reloadMu.
func (w *Watcher) reparse() (string, error) {
	cfg, err := w.opts.load(w.path)
	if err != nil {
		slog.Warn("config reload rejected", "error", err)
		w.lastErr.Store(&err)

		return "", fmt.Errorf("reload config %s: %w", w.path, err)
	}
	w.current.Store(cfg)
	slog.Info("config reloaded", "tap_window_ms", cfg.Timeouts.TapWindowMs)
	var none error
	w.lastErr.Store(&none)

	return fmt.Sprintf("applied (tap_window_ms=%d)", cfg.Timeouts.TapWindowMs), nil
}
