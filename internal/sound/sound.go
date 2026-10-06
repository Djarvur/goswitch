// Package sound plays the daemon's acoustic feedback (the 2026-10-05
// sound redesign, owner «делай», todo «Звук»): the GNOME sound theme's
// tones through a PERSISTENT native sound-server connection, the
// synthesized sines as the fallback whenever theme PCM is not ready.
// Best-effort by contract — the tones never block, error or otherwise
// alter the gesture that triggered them; muted sound holds NO
// sound-server connection (lazy dial on enable, close on disable); every
// failure is ONE WARN per episode and a quiet no-op. The theme settings
// ride a supervised gsettings watcher fully async from playback — the
// play path never waits for a settings read or a decode. The tones know
// nothing about text — no word content ever reaches this package or its
// logs (D-20/D-21).
package sound

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// The gsettings seam's pinned vocabulary (T-SQU-04: schema and key
// constants only — no shell, no interpolated user data; the monitor
// reads, never writes).
const (
	binGSettings     = "gsettings"
	gsGet            = "get"
	gsMonitor        = "monitor"
	gsSchema         = "org.gnome.desktop.sound"
	gsKeyTheme       = "theme-name"
	gsKeyEventSounds = "event-sounds"
	gsValueTrue      = "true"
)

// The degradation reasons of the warn-once episodes (the closed
// vocabulary the corpus pins — reason slugs and library errors only,
// never user context, D-20/D-21).
const (
	reasonConnectFailed    = "connect-failed"    // the sound server refused the dial
	reasonStreamFailed     = "stream-failed"     // the playback stream could not open or write
	reasonThemeUnavailable = "theme-unavailable" // the gsettings path or the theme read failed
)

// The tone math of the synthesized fallback (the old bundled-tone pins
// carried over; pcm.go's synthTone consumes them) plus the canonical
// stream format the synth opens before the first decoded theme PCM
// arrives.
const (
	toneRate        = 22050.0 // samples per second — the canonical low-fidelity PCM rate
	toneSeconds     = 0.06    // a short confirmation click — tens of milliseconds
	toneAmplitude   = 0.5     // half full scale — audible, comfortably free of clipping
	maxSample       = 32767.0 // the 16-bit full scale the amplitude scales
	freqFlip        = 880.0   // the flip tone's pitch (A5)
	freqAutoCorrect = 1320.0  // the autocorrect tone's pitch (E6) — audibly distinct
	twoPi           = 2 * math.Pi
	synthChannels   = 1 // the canonical synth's mono
	stereoChannels  = 2 // the theme reality's stereo (Yaru)
)

// eventSoundsDisabled is the PD-1 INFO: the system-wide event-sounds key
// turned off — an observed preference (goswitch tones ARE event sounds),
// not a degradation.
const eventSoundsDisabled = "event sounds disabled system-wide"

// monitorRestartPause paces a monitor that cannot start (a missing binary
// or schema) — supervision with a pause, never a busy retry, never a
// settings-read ticker (PD-2).
const monitorRestartPause = 5 * time.Second

// errEmptySetting is a settings read that came back empty — the schema's
// key answered nothing usable.
var errEmptySetting = errors.New("settings key answered empty")

// toneKind distinguishes the two feedback tones — the synthesized
// fallback's pitches.
type toneKind int

const (
	toneFlip toneKind = iota
	toneAutoCorrect
)

// toneRequest is one enqueued tone: the sound-theme event to resolve and
// the kind that names the fallback pitch.
type toneRequest struct {
	event string
	kind  toneKind
}

// commandRunner is the gsettings seam: one pinned-argv execution shape
// for the reads and the supervised monitor (T-SQU-04 — the argv carries
// schema and key constants only).
type commandRunner interface {
	output(argv []string) (string, error)
	stream(argv []string) (io.ReadCloser, func() error, error)
}

// execRunner is the production runner: pinned-argv commands, no shell.
type execRunner struct{}

func (execRunner) output(argv []string) (string, error) {
	//nolint:noctx // a deadline-free settings read — the install package's pinned gsettings shape
	// #nosec G204 -- the argv carries only the pinned schema/key constants (T-SQU-04); no shell, no user data
	out, err := exec.Command(argv[0], argv[1:]...).Output()
	if err != nil {
		return "", fmt.Errorf("read settings: %w", err)
	}

	return string(out), nil
}

func (execRunner) stream(argv []string) (io.ReadCloser, func() error, error) {
	//nolint:noctx // the supervised monitor — its own exit is the restart trigger, nothing kills a running monitor
	// #nosec G204 -- the argv carries only the pinned schema constants (T-SQU-04); the monitor reads, never writes
	cmd := exec.Command(argv[0], argv[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("pipe settings monitor: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start settings monitor: %w", err)
	}

	return stdout, cmd.Wait, nil
}

// Player plays the two tones of the acoustic feedback: Flip for every
// script flip, AutoCorrect for every fired correction — from the
// desktop's sound theme through a persistent native sound-server
// connection, the synthesized sines whenever theme PCM is not ready.
// Start launches the playback worker and the gsettings watcher (PD-3:
// the watcher runs from daemon start regardless of any muted state);
// Stop closes the connection — the mute lifecycle (requirement 3). The
// seam methods are a mutex read plus a non-blocking enqueue: the hot
// path of the gesture never waits for a sound.
type Player struct {
	flipEvent string // the flip tone's sound-theme event (the schema constant)

	mu      sync.Mutex
	acEvent string          // the autocorrect tone's event (the document's effective value; hot — WR-02)
	warned  map[string]bool // one WARN per episode per reason — a healthy tone reopens the budget

	// The seams (production in New, the corpus's fakes in tests).
	dial     dialer
	runner   commandRunner
	cache    *pcmCache
	resolver *themeResolver

	// currentTheme carries the watcher's latest theme name — the worker's
	// cache key and the re-pin's prefetch partner (the watcher writes, the
	// worker and the setter read).
	currentTheme atomic.Value // string

	toneCh       chan toneRequest // buffered 1 — the drop-if-busy enqueue
	control      chan struct{}    // buffered 1 — the mute lifecycle's stop signal
	startOnce    sync.Once
	monitorPause time.Duration // the supervision pause between unstartable monitors (PD-2)
}

// New returns the production player for the two sound-theme events: the
// flip event is the schema constant (config.DefaultSoundFlipEvent — NOT
// configurable, the owner asked for the switch and the distinct
// autocorrect event only), the autocorrect event the document's
// effective value.
func New(flipEvent, autocorrectEvent string) *Player {
	p := &Player{
		flipEvent:    flipEvent,
		acEvent:      autocorrectEvent,
		warned:       make(map[string]bool),
		dial:         dialPulse,
		runner:       execRunner{},
		resolver:     newFileResolver(),
		toneCh:       make(chan toneRequest, 1),
		control:      make(chan struct{}, 1),
		monitorPause: monitorRestartPause,
	}
	p.currentTheme.Store(fallbackTheme)
	p.cache = newPCMCache(p.decodeThemeEvent)

	return p
}

// newFileResolver builds the production XDG resolver over the real data
// directories and the real filesystem.
func newFileResolver() *themeResolver {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "" // no home — the system data dirs still carry the themes
	}

	return newThemeResolver(openFile, xdgSoundRoots(os.Getenv("XDG_DATA_HOME"), os.Getenv("XDG_DATA_DIRS"), home))
}

// openFile is the production opener — the resolver's real-filesystem arm.
func openFile(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open theme file: %w", err)
	}

	return f, nil
}

// Start launches the playback worker and the gsettings watcher —
// idempotent. The connection itself is dialed lazily by the first tone;
// the watcher's reads and decodes live entirely off the play path.
func (p *Player) Start() {
	p.startOnce.Do(func() {
		go p.watchSettings()
		go p.work()
	})
}

// Flip plays the flip tone — fire-and-forget: a field read and a
// non-blocking enqueue.
func (p *Player) Flip() {
	p.enqueue(toneRequest{event: p.flipEvent, kind: toneFlip})
}

// AutoCorrect plays the autocorrect tone — the same fire-and-forget path
// with the document's effective event (read under the mutex: the fold's
// SetAutocorrectEvent may land on any goroutine, WR-02).
func (p *Player) AutoCorrect() {
	p.mu.Lock()
	event := p.acEvent
	p.mu.Unlock()
	p.enqueue(toneRequest{event: event, kind: toneAutoCorrect})
}

// SetAutocorrectEvent re-pins the autocorrect tone's document event
// (WR-02: the key folds live — the D-32 contract holds without a
// restart) and prefetches the new (theme, event) pair asynchronously —
// a tone arriving before the prefetch lands plays the synthesized tone,
// never a wait.
func (p *Player) SetAutocorrectEvent(event string) {
	p.mu.Lock()
	p.acEvent = event
	p.mu.Unlock()
	theme, _ := p.currentTheme.Load().(string)
	go p.cache.load(theme, []string{p.flipEvent, event})
}

// Stop closes the sound-server connection — the mute lifecycle's close
// half (requirement 3, PD-1's master-mute reading): muted means NO
// connection at all. The signal is buffered and the call returns without
// waiting for a mid-tone write; repeated Stop is a no-op; the next tone
// after re-enable dials again (the lazy connect on enable).
func (p *Player) Stop() {
	select {
	case p.control <- struct{}{}:
	default:
	}
}

// decodeThemeEvent resolves one (theme, event) pair against the XDG data
// dirs and decodes it — the cache's production loader seam; it runs
// strictly off the play path (at Start, on a theme change, on a re-pin).
func (p *Player) decodeThemeEvent(theme, event string) (*pcm, error) {
	rc, err := p.resolver.resolve(theme, event)
	if err != nil {
		return nil, fmt.Errorf("resolve theme event: %w", err)
	}
	defer func() { _ = rc.Close() }()

	return decodeOggVorbis(rc)
}

// enqueue hands one tone to the worker — drop-if-busy: a busy worker
// drops the tone, the gesture never waits (the T-07-08-02 pin).
func (p *Player) enqueue(req toneRequest) {
	select {
	case p.toneCh <- req:
	default:
	}
}

// playback is the worker's connection state: the lazily dialed conn, the
// format-following stream, and the mute flag. Owned exclusively by the
// worker goroutine.
type playback struct {
	conn   audioConn
	stream audioStream
	format streamFormat
	live   bool
}

// close tears the whole playback down — idempotent, safe on a
// never-dialed state (requirement 3's letter). The stream drains first:
// a mid-tone close lets the tone finish, a recreation leaves no stale
// audio behind.
func (pb *playback) close() {
	if pb.stream != nil {
		pb.stream.drain()
		pb.stream.close()
		pb.stream = nil
	}
	if pb.conn != nil {
		pb.conn.close()
		pb.conn = nil
	}
	pb.live = false
}

// work is the single playback worker: it owns the connection lifecycle,
// the cache lookup (miss → synthesized fallback) and the mute fold.
// Tones cannot arrive while the actor's gate is off, so the first tone
// after a stop IS the re-enable — it dials again (requirement 3's lazy
// connect on enable).
func (p *Player) work() {
	pb := &playback{}
	for {
		select {
		case <-p.control:
			pb.close() // the mute lifecycle: the connection goes away entirely
		case req := <-p.toneCh:
			p.foldTone(pb, req)
		}
	}
}

// foldTone plays one enqueued tone: a cache hit streams the decoded theme
// PCM at its own format, a miss synthesizes at the stream's — never a
// wait for a decode or a settings read (requirement 2).
func (p *Player) foldTone(pb *playback, req toneRequest) {
	theme, _ := p.currentTheme.Load().(string)
	if decoded, ok := p.cache.lookup(theme, req.event); ok {
		p.playPCM(pb, decoded)

		return
	}
	p.playSynth(pb, req.kind)
}

// playPCM streams one decoded theme event, recreating the stream when
// the PCM's format changed (the locked «формат стрима следует за
// декодированным PCM» rule). A healthy tone closes the warn episode.
func (p *Player) playPCM(pb *playback, decoded *pcm) {
	if !p.ensureStream(pb, decoded.format) {
		return
	}
	p.write(pb, decoded.data)
}

// playSynth streams the fallback tone at the stream's own format — the
// canonical mono shape when no decoded PCM has opened one yet.
func (p *Player) playSynth(pb *playback, kind toneKind) {
	format := streamFormat{Rate: int(toneRate), Channels: synthChannels}
	if pb.live {
		format = pb.format // the stream's format — the synth never resamples
	}
	data, err := synthTone(freqOf(kind), format)
	if err != nil {
		p.warn(reasonStreamFailed, err)

		return
	}
	if !p.ensureStream(pb, format) {
		return
	}
	p.write(pb, data)
}

// freqOf names the fallback pitch of a tone kind.
func freqOf(kind toneKind) float64 {
	if kind == toneAutoCorrect {
		return freqAutoCorrect
	}

	return freqFlip
}

// ensureStream dials lazily and keeps the stream's format following the
// tone's — the dial and any recreation live here, OFF the gesture path.
// false = the tone is dropped with its episode WARN (connect or stream).
func (p *Player) ensureStream(pb *playback, format streamFormat) bool {
	if pb.conn == nil {
		conn, err := p.dial()
		if err != nil {
			p.warn(reasonConnectFailed, err)

			return false
		}
		pb.conn = conn
		pb.live = false
	}
	if pb.live && pb.format == format {
		return true
	}
	if pb.stream != nil {
		pb.stream.drain()
		pb.stream.close()
		pb.stream = nil
	}
	pb.live = false
	stream, err := pb.conn.openStream(format)
	if err != nil {
		p.warn(reasonStreamFailed, err)
		pb.close() // a conn that cannot stream is dead weight — the next tone dials fresh

		return false
	}
	pb.stream = stream
	pb.format = format
	pb.live = true

	return true
}

// write hands one tone to the stream — the one write; a failure warns its
// own episode and tears the connection down (the next tone starts fresh).
func (p *Player) write(pb *playback, data []byte) {
	if err := pb.stream.play(data); err != nil {
		p.warn(reasonStreamFailed, err)
		pb.close()

		return
	}
	p.closeEpisode()
}

// watchSettings tracks the desktop's sound settings (PD-2): a FULL read
// of both keys at every (re)start — the start read and the catch-up for
// anything missed while a monitor was down — then the supervised
// monitor; a monitor exit restarts the cycle. The watcher runs from
// daemon start regardless of any muted state (PD-3: settings
// infrastructure, never a sound-server connection).
func (p *Player) watchSettings() {
	for {
		p.reread()
		stdout, wait, err := p.runner.stream([]string{binGSettings, gsMonitor, gsSchema})
		if err != nil {
			// A monitor that cannot start gets a pause, never a busy
			// retry — supervision with backoff; the re-read above already
			// applied the last-known truth.
			p.warn(reasonThemeUnavailable, err)
			time.Sleep(p.monitorPause)

			continue
		}
		p.scanMonitor(stdout, wait)
	}
}

// reread performs the FULL two-key read (PD-2's catch-up): the theme
// name and the event-sounds switch, applying both. An unavailable
// gsettings path degrades to the fallback theme; an unreadable
// event-sounds key assumes true (the GNOME default, PD-1) — both in the
// theme-unavailable episode's one WARN.
func (p *Player) reread() {
	theme, err := p.readThemeName()
	if err != nil {
		p.warn(reasonThemeUnavailable, err)
		theme = fallbackTheme // the degradation: the spec's fallback carries playback
	}
	p.applyTheme(theme)
	on, err := p.readEventSounds()
	if err != nil {
		p.warn(reasonThemeUnavailable, err)
	}
	if !on {
		p.Stop() // the startup state simply begins muted — no transition INFO
	}
}

// scanMonitor consumes one monitor process's lines until it exits: a
// theme line triggers the async rebuild + atomic swap (the next tone's
// theme moves without any synchronous wait); an event-sounds=false line
// applies PD-1. The caller's loop restarts with the full re-read.
func (p *Player) scanMonitor(stdout io.ReadCloser, wait func() error) {
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if theme, ok := parseThemeLine(line); ok {
			p.applyTheme(theme)

			continue
		}
		if on, ok := parseEventSoundsLine(line); ok && !on {
			// PD-1: system event sounds off is the master mute's twin —
			// an observed preference, one INFO, the connection closes,
			// re-enable redials lazily.
			slog.Info(eventSoundsDisabled)
			p.Stop()
		}
	}
	_ = stdout.Close()
	_ = wait() // the exit itself is the supervision's restart trigger
}

// readThemeName reads the desktop's theme name once.
func (p *Player) readThemeName() (string, error) {
	out, err := p.runner.output([]string{binGSettings, gsGet, gsSchema, gsKeyTheme})
	if err != nil {
		return "", fmt.Errorf("read theme name: %w", err)
	}
	theme := unquoteSetting(out)
	if theme == "" {
		return "", fmt.Errorf("read theme name: %w", errEmptySetting)
	}

	return theme, nil
}

// readEventSounds reads the system-wide event-sounds switch.
func (p *Player) readEventSounds() (bool, error) {
	out, err := p.runner.output([]string{binGSettings, gsGet, gsSchema, gsKeyEventSounds})
	if err != nil {
		return true, fmt.Errorf("read event-sounds: %w", err)
	}

	return strings.Contains(out, gsValueTrue), nil
}

// applyTheme publishes a theme observation: the name becomes the worker's
// cache key and the (theme, events) decode runs asynchronously — the
// atomic swap lands when it lands; the play path never waits for it.
func (p *Player) applyTheme(theme string) {
	p.currentTheme.Store(theme)
	p.mu.Lock()
	ac := p.acEvent
	p.mu.Unlock()
	go p.cache.load(theme, []string{p.flipEvent, ac})
}

// parseThemeLine reads one monitor line's theme-name value.
func parseThemeLine(line string) (theme string, ok bool) {
	_, value, found := strings.Cut(line, gsKeyTheme+":")
	if !found {
		return "", false
	}
	theme = unquoteSetting(value)

	return theme, theme != ""
}

// parseEventSoundsLine reads one monitor line's event-sounds value.
func parseEventSoundsLine(line string) (on, ok bool) {
	_, value, found := strings.Cut(line, gsKeyEventSounds+":")
	if !found {
		return true, false
	}

	return strings.Contains(value, gsValueTrue), true
}

// unquoteSetting strips the quoting gsettings wraps its values in.
func unquoteSetting(value string) string {
	return strings.Trim(strings.TrimSpace(value), `'"`)
}

// closeEpisode reopens the warn budget: a healthy tone proves the sound
// stack — a LATER failure warns again (the appidWarned reset form: a
// healthy answer closes the episode).
func (p *Player) closeEpisode() {
	p.mu.Lock()
	defer p.mu.Unlock()

	clear(p.warned)
}

// warn records one degradation WARN for the reason — ONE per episode
// (the actor's warnAppid form): the first failure of an episode warns,
// the rest stay quiet. The record names the reason and the transport
// error only — never user context (D-20/D-21).
func (p *Player) warn(reason string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.warned[reason] {
		return
	}
	p.warned[reason] = true
	if err == nil {
		slog.Warn("sound unavailable", "reason", reason)

		return
	}
	slog.Warn("sound unavailable", "reason", reason, "error", err)
}
