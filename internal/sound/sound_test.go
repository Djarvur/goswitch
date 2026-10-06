//nolint:testpackage // drives the unexported dialer/runner seams — the sound corpus in-package precedent
package sound

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

// The corpus's fixture vocabulary: the watched theme name, the theme a
// monitor line switches to, the re-pinned autocorrect event, and the
// simulated seam failures.
const (
	watchedTheme   = "Yaru"
	changedTheme   = "Kokon"
	rePinnedEvent  = "system-message"
	eventSoundsOff = "event sounds disabled system-wide"
)

// errSimulatedDial is the dial double's failure — the server-unavailable
// state of the connect-failed episode cell.
var errSimulatedDial = errors.New("simulated dial failure")

// errSimulatedStream is the stream-creation double's failure — the broken
// stream state of the stream-failed episode cell.
var errSimulatedStream = errors.New("simulated stream failure")

// errSimulatedGet is the gsettings double's failure — the missing-binary
// state of the theme-unavailable cell.
var errSimulatedGet = errors.New("simulated gsettings failure")

// pollBudget bounds every poll of an asynchronous effect — orders of
// magnitude above the worker's real latency, far below any test timeout.
const pollBudget = 2 * time.Second

// poll runs cond until it holds or the budget lapses — the poll-until
// idiom, never a fixed sleep.
func poll(cond func() bool) bool {
	for deadline := time.Now().Add(pollBudget); time.Now().Before(deadline); {
		if cond() {
			return true
		}
		time.Sleep(time.Millisecond)
	}

	return cond()
}

// staysFalse pins the absence of an effect: cond must hold false for a
// short settling window (a no-op assertion, never a fixed-sleep oracle).
func staysFalse(t *testing.T, what string, cond func() bool) {
	t.Helper()
	for range settleWindows {
		if cond() {
			t.Fatalf("%s happened — must stay absent", what)
		}
		time.Sleep(settleStep)
	}
}

const (
	settleWindows = 25
	settleStep    = 2 * time.Millisecond
)

// logBuffer is a guarded log sink — the captureLogs form of the session
// corpus, local to the sound corpus.
type logBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *logBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *logBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

// captureSoundLogs redirects the process default logger into a guarded
// buffer — the WARN/INFO counts of the episode cells read it.
func captureSoundLogs(t *testing.T) *logBuffer {
	t.Helper()
	buf := &logBuffer{}
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, nil)))

	return buf
}

// fakeConn is one dialed sound-server connection: it records the stream
// formats it opened and its own closes — the mute lifecycle's oracle.
type fakeConn struct {
	mu        sync.Mutex
	formats   []streamFormat
	streams   []*fakeStream
	streamErr error
	closes    int
}

// openStream mirrors the library's contract: stream creation calls Start
// (blocked until the reader serves the first buffer) before returning.
//
//nolint:ireturn // the fake hands the connection-stream interface back (the production seam's shape)
func (c *fakeConn) openStream(format streamFormat) (audioStream, error) {
	c.mu.Lock()
	if c.streamErr != nil {
		err := c.streamErr
		c.mu.Unlock()

		return nil, err
	}
	s := &fakeStream{format: format, reader: newToneReader()}
	c.streams = append(c.streams, s)
	c.formats = append(c.formats, format)
	c.mu.Unlock()

	s.start() // the prototype's SIGQUIT lesson lives here: this blocks on the reader

	return s, nil
}

func (c *fakeConn) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closes++
}

func (c *fakeConn) closeCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.closes
}

func (c *fakeConn) lastStream() *fakeStream {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.streams) == 0 {
		return nil
	}

	return c.streams[len(c.streams)-1]
}

// fakeStream is one opened playback stream: it records writes, formats,
// drains and closes, and its start mirrors the real library — blocked
// until the toneReader serves the first buffer.
type fakeStream struct {
	format  streamFormat
	reader  *toneReader
	playErr error

	mu     sync.Mutex
	writes [][]byte
	drains int
	closes int
}

// start is the blocking contract: the real Start() hangs until the first
// buffer reaches the server, so the fake consumes one reader callback —
// a toneReader that lost its silence-default hangs this call and the
// whole corpus with it (the pin).
func (s *fakeStream) start() {
	buf := make([]byte, silenceChunk)
	_, _ = s.reader.Read(buf)
}

func (s *fakeStream) play(data []byte) error {
	s.mu.Lock()
	s.writes = append(s.writes, data)
	err := s.playErr
	s.mu.Unlock()
	s.reader.feed(data)

	return err
}

func (s *fakeStream) drain() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.drains++
}

func (s *fakeStream) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closes++
}

func (s *fakeStream) writeCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.writes)
}

// allWrites snapshots every recorded write in order.
func (s *fakeStream) allWrites() [][]byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([][]byte(nil), s.writes...)
}

func (s *fakeStream) closeCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.closes
}

// fakeDialer is the dial seam: a programmed error sequence (nil = success,
// each success opens a fresh fakeConn), a per-connection stream failure,
// and a wedge channel for the non-blocking pin.
type fakeDialer struct {
	mu        sync.Mutex
	errs      []error // popped per dial; empty = every dial succeeds
	streamErr error   // every opened conn's stream creation fails
	dials     int
	conns     []*fakeConn
	block     chan struct{} // non-nil = every dial wedges forever
}

//nolint:ireturn // the fake hands the connection interface back (the production seam's shape)
func (d *fakeDialer) dial() (audioConn, error) {
	if d.block != nil {
		<-d.block // the wedged dial of the non-blocking pin — never completes
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.dials++
	if len(d.errs) > 0 {
		err := d.errs[0]
		d.errs = d.errs[1:]
		if err != nil {
			return nil, err
		}
		// A nil entry programs a SUCCESSFUL dial — fall through to the
		// fresh connection.
	}
	c := &fakeConn{streamErr: d.streamErr}
	d.conns = append(d.conns, c)

	return c, nil
}

func (d *fakeDialer) dialCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.dials
}

func (d *fakeDialer) lastConn() *fakeConn {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.conns) == 0 {
		return nil
	}

	return d.conns[len(d.conns)-1]
}

// lastWriteCount is the last conn's last stream's write count (0 before
// anything dialed) — the poll closures read it without nil chains.
func (d *fakeDialer) lastWriteCount() int {
	c := d.lastConn()
	if c == nil {
		return 0
	}
	s := c.lastStream()
	if s == nil {
		return 0
	}

	return s.writeCount()
}

// fakeMonitor is one supervised monitor process: the watcher reads its
// lines, its exit triggers the PD-2 catch-up re-read and the restart.
type fakeMonitor struct {
	lines  chan string
	closed chan struct{}
	once   sync.Once
}

func newFakeMonitor() *fakeMonitor {
	return &fakeMonitor{
		lines:  make(chan string),
		closed: make(chan struct{}),
	}
}

// emit pushes one monitor output line (the wire form gsettings monitor
// writes on a settings change).
func (m *fakeMonitor) emit(line string) {
	m.lines <- line
}

// exit ends the monitor process: the reader answers EOF and the wait
// function returns — the supervision's restart trigger.
func (m *fakeMonitor) exit() {
	m.once.Do(func() {
		close(m.lines)
		close(m.closed)
	})
}

// reader is the monitor's stdout — lines drained through a pipe-shaped
// reader, EOF at exit.
func (m *fakeMonitor) reader() io.ReadCloser {
	return &chanReader{lines: m.lines}
}

func (m *fakeMonitor) wait() error {
	<-m.closed

	return nil
}

// chanReader drains a line channel through the io.ReadCloser shape the
// watcher's scanner consumes.
type chanReader struct {
	lines <-chan string
	buf   []byte
}

func (r *chanReader) Read(p []byte) (int, error) {
	for len(r.buf) == 0 {
		line, ok := <-r.lines
		if !ok {
			return 0, io.EOF
		}
		r.buf = append(r.buf, line...)
		r.buf = append(r.buf, '\n')
	}
	n := copy(p, r.buf)
	r.buf = r.buf[n:]

	return n, nil
}

func (r *chanReader) Close() error { return nil }

// fakeRunner is the gsettings seam: a programmed get answer per key, a
// queue of monitor processes per stream call (idle monitors beyond the
// queue — every cell that never touches the monitor still supervises one).
type fakeRunner struct {
	mu        sync.Mutex
	gets      [][]string
	getErr    error
	theme     string
	sounds    string
	monitors  []*fakeMonitor
	streamErr error
	active    *fakeMonitor // the monitor the watcher is currently consuming
}

func (r *fakeRunner) output(argv []string) (string, error) {
	r.mu.Lock()
	r.gets = append(r.gets, argv)
	theme, sounds, getErr := r.theme, r.sounds, r.getErr
	r.mu.Unlock()
	if getErr != nil {
		return "", getErr
	}
	if len(argv) > 0 && argv[len(argv)-1] == gsKeyTheme {
		return "'" + theme + "'", nil
	}
	if len(argv) > 0 && argv[len(argv)-1] == gsKeyEventSounds {
		return sounds, nil
	}

	return "", errSimulatedGet
}

func (r *fakeRunner) stream(argv []string) (io.ReadCloser, func() error, error) {
	r.mu.Lock()
	r.gets = append(r.gets, argv)
	if r.streamErr != nil {
		err := r.streamErr
		r.mu.Unlock()

		return nil, nil, err
	}
	var m *fakeMonitor
	if len(r.monitors) > 0 {
		m = r.monitors[0]
		r.monitors = r.monitors[1:]
	} else {
		// No programmed monitor left: run forever silent — the idle
		// supervision of every cell that never touches the monitor.
		m = newFakeMonitor()
	}
	r.active = m
	r.mu.Unlock()

	return m.reader(), m.wait, nil
}

// monitor hands the watcher's active monitor to the corpus — the emit/
// exit handle of whichever process stream() started last.
func (r *fakeRunner) monitor() *fakeMonitor {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.active
}

func (r *fakeRunner) getCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.gets)
}

// awaitMonitor waits until the watcher's supervision reached the given
// monitor process — every emit/exit assertion must hold this fence.
func awaitMonitor(t *testing.T, r *fakeRunner, m *fakeMonitor) {
	t.Helper()
	if !poll(func() bool { return r.monitor() == m }) {
		t.Fatal("the watcher never reached the programmed monitor")
	}
}

// newSoundTestPlayer builds one player over the corpus seams: a failing
// decode (every tone synthesized), a successful dial, the watched theme
// answered by the fake runner.
func newSoundTestPlayer() (*Player, *fakeDialer, *fakeRunner) {
	p := New("bell", "message")
	d := &fakeDialer{}
	r := &fakeRunner{theme: watchedTheme, sounds: "true"}
	p.dial = d.dial
	p.runner = r
	p.cache = newPCMCache(func(string, string) (*pcm, error) { return nil, errDecodeRefused })
	p.currentTheme.Store(watchedTheme)

	return p, d, r
}

// TestPlayer_LazyConnectDialsOncePerConnectedPeriod pins requirement 3's
// lazy half and the prototype's economics: construction and Start dial
// NOTHING; the first tone dials exactly once; a burst of tones during one
// connected period reuses the client.
func TestPlayer_LazyConnectDialsOncePerConnectedPeriod(t *testing.T) {
	p, d, _ := newSoundTestPlayer()
	p.Start()

	if got := d.dialCount(); got != 0 {
		t.Fatalf("dials after construction + Start = %d, want 0 — the connection is lazy", got)
	}

	p.Flip()
	if !poll(func() bool {
		c := d.lastConn()

		return d.dialCount() == 1 && c != nil && c.lastStream().writeCount() == 1
	}) {
		t.Fatal("the first tone never dialed and played")
	}

	p.AutoCorrect()
	if !poll(func() bool { return d.lastWriteCount() == 2 }) {
		t.Fatal("the second tone never played")
	}
	p.Flip()
	if !poll(func() bool { return d.lastWriteCount() == 3 }) {
		t.Fatal("the third tone never played")
	}
	if got := d.dialCount(); got != 1 {
		t.Errorf("dials after a three-tone burst = %d, want exactly 1 — the client is reused", got)
	}
}

// TestPlayer_FlipNonBlockingUnderWedgedWorker pins the drop-if-busy
// contract (the T-07-08-02 pin carried over): a wedged worker — a dial
// that never completes — must not stall the seam call.
func TestPlayer_FlipNonBlockingUnderWedgedWorker(t *testing.T) {
	p, _, _ := newSoundTestPlayer()
	p.dial = (&fakeDialer{block: make(chan struct{})}).dial
	p.Start()

	done := make(chan struct{})
	go func() {
		p.Flip()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(pollBudget):
		t.Fatal("Flip blocked on the wedged worker — the seam call must answer immediately")
	}
}

// TestPlayer_StopMuteLifecycle pins requirement 3's close half (PD-1's
// master-mute reading): Stop closes client+stream exactly once, repeated
// Stop is a no-op, the muted period dials nothing, and the first tone
// after re-enable dials again.
func TestPlayer_StopMuteLifecycle(t *testing.T) {
	p, d, _ := newSoundTestPlayer()
	p.Start()
	p.Flip()
	if !poll(func() bool {
		c := d.lastConn()

		return d.dialCount() == 1 && c != nil && c.lastStream().writeCount() == 1
	}) {
		t.Fatal("the pre-mute tone never played")
	}
	c := d.lastConn()

	p.Stop()
	if !poll(func() bool { return c.closeCount() == 1 && c.lastStream().closeCount() == 1 }) {
		t.Fatal("Stop never closed the client and stream")
	}

	p.Stop() // repeated Stop is a no-op
	staysFalse(t, "a second close", func() bool { return c.closeCount() != 1 })

	staysFalse(t, "a dial during the muted period", func() bool { return d.dialCount() != 1 })

	p.Flip() // the re-enable tone
	if !poll(func() bool { return d.dialCount() == 2 }) {
		t.Fatal("the post-mute tone never redialed — the lazy connect on enable is broken")
	}
	fresh := d.lastConn()
	if !poll(func() bool { return fresh.lastStream().writeCount() == 1 }) {
		t.Fatal("the post-mute tone never played")
	}
}

// TestPlayer_ConnectFailedWarnOnceThenReopen pins the degradation
// discipline: a dial failure is ONE warn-once WARN per episode (reason
// connect-failed), quiet afterwards, a later successful tone closes the
// episode (a subsequent failure warns again), and no failure ever
// surfaces an error to the caller.
func TestPlayer_ConnectFailedWarnOnceThenReopen(t *testing.T) {
	buf := captureSoundLogs(t)
	p, _, _ := newSoundTestPlayer()
	fd := &fakeDialer{errs: []error{errSimulatedDial, errSimulatedDial, nil, errSimulatedDial}}
	p.dial = fd.dial
	p.Start()

	p.Flip()
	if !poll(func() bool { return fd.dialCount() == 1 }) {
		t.Fatal("the first failed dial never happened")
	}
	p.Flip()
	if !poll(func() bool { return fd.dialCount() == 2 }) {
		t.Fatal("the second tone never attempted its dial — a failed episode degrades, never disables")
	}
	if got := strings.Count(buf.String(), "sound unavailable"); got != 1 {
		t.Errorf("connect-failed WARNs = %d, want exactly 1 per episode", got)
	}

	p.Flip() // the successful tone closes the episode
	if !poll(func() bool { return fd.dialCount() == 3 && fd.lastWriteCount() == 1 }) {
		t.Fatal("the successful tone never played")
	}

	// A later failure warns AGAIN — the episode reopened. The healthy
	// connection first goes away the way it does in production (the mute
	// close), so the next tone is a fresh dial into the programmed
	// failure.
	p.Stop()
	if !poll(func() bool { return fd.lastConn().closeCount() == 1 }) {
		t.Fatal("the mute close never retired the healthy connection")
	}
	p.Flip()
	if !poll(func() bool { return fd.dialCount() == 4 }) {
		t.Fatal("the post-episode failure never attempted its dial")
	}
	if got := strings.Count(buf.String(), "sound unavailable"); got != 2 {
		t.Errorf("WARNs after the episode reopened = %d, want exactly 2", got)
	}
	if got := strings.Count(buf.String(), reasonConnectFailed); got != 2 {
		t.Errorf("connect-failed reasons in the log = %d, want 2", got)
	}
}

// TestPlayer_StreamFailureOwnEpisode pins the stream-failed episode: a
// broken stream creation warns once under its OWN reason and the next
// tone starts fresh — the failure never surfaces an error and never
// disables playback.
func TestPlayer_StreamFailureOwnEpisode(t *testing.T) {
	buf := captureSoundLogs(t)
	p, d, _ := newSoundTestPlayer()
	d.streamErr = errSimulatedStream
	p.Start()

	p.Flip()
	if !poll(func() bool { return d.dialCount() == 1 }) {
		t.Fatal("the first tone never dialed")
	}
	p.Flip()
	if !poll(func() bool { return d.dialCount() == 2 }) {
		t.Fatal("the second tone never redialed — a stream failure must start fresh on the next tone")
	}
	if got := strings.Count(buf.String(), reasonStreamFailed); got != 1 {
		t.Errorf("stream-failed reasons = %d, want exactly 1 per episode", got)
	}
}

// TestPlayer_WatcherThemeCatchUp pins the watcher's full cycle (PD-2):
// Start reads the theme once and loads it; a monitor theme line triggers
// the async rebuild and atomic swap (the next tone's theme moves without
// any synchronous wait); a monitor exit restarts the monitor AND performs
// the FULL re-read of both keys.
func TestPlayer_WatcherThemeCatchUp(t *testing.T) {
	var mu sync.Mutex
	decodes := make(map[string]int)
	decode := func(theme, event string) (*pcm, error) {
		mu.Lock()
		decodes[theme+"/"+event]++
		mu.Unlock()

		return nil, errDecodeRefused
	}

	first := newFakeMonitor()
	idle := newFakeMonitor()
	p, _, r := newSoundTestPlayer()
	p.cache = newPCMCache(decode)
	r.mu.Lock()
	r.monitors = []*fakeMonitor{first, idle}
	r.mu.Unlock()
	p.Start()

	if !poll(func() bool {
		mu.Lock()
		defer mu.Unlock()

		return decodes[watchedTheme+"/"+fsEventBell] > 0
	}) {
		t.Fatal("Start never loaded the watched theme — the initial get is missing")
	}
	awaitMonitor(t, r, first)

	first.emit("theme-name: '" + changedTheme + "'")
	if !poll(func() bool {
		mu.Lock()
		defer mu.Unlock()

		return decodes[changedTheme+"/"+fsEventBell] > 0
	}) {
		t.Fatal("the theme line never triggered the async rebuild — the swap is broken")
	}

	gets := r.getCount()
	first.exit()
	if !poll(func() bool { return r.getCount() >= gets+2 }) {
		t.Fatal("the monitor restart never performed the full two-key re-read (PD-2 catch-up)")
	}
	awaitMonitor(t, r, idle) // the restart supervises the next process
}

// TestPlayer_WatcherEventSoundsMute pins PD-1: an event-sounds=false
// monitor line closes the connection with ONE INFO and re-enable redials
// lazily on the next tone.
func TestPlayer_WatcherEventSoundsMute(t *testing.T) {
	buf := captureSoundLogs(t)
	p, d, r := newSoundTestPlayer()
	p.Start()
	p.Flip()
	if !poll(func() bool {
		c := d.lastConn()

		return d.dialCount() == 1 && c != nil && c.lastStream().writeCount() == 1
	}) {
		t.Fatal("the pre-mute tone never played")
	}
	c := d.lastConn()
	if !poll(func() bool { return r.monitor() != nil }) {
		t.Fatal("the watcher never started the monitor")
	}
	mon := r.monitor()

	mon.emit("event-sounds: false")
	if !poll(func() bool { return c.closeCount() == 1 }) {
		t.Fatal("the event-sounds=false transition never closed the connection (PD-1)")
	}
	if got := strings.Count(buf.String(), eventSoundsOff); got != 1 {
		t.Errorf("event-sounds INFO lines = %d, want exactly 1", got)
	}

	mon.emit("event-sounds: true")
	staysFalse(t, "an eager redial on re-enable", func() bool { return d.dialCount() != 1 })

	p.Flip()
	if !poll(func() bool { return d.dialCount() == 2 }) {
		t.Fatal("the post-re-enable tone never redialed")
	}
}

// TestPlayer_EventSoundsMuteDropsTonesWithoutRedial pins the muted arm's
// suppression (PD-1's effective play: config sound AND event-sounds): a
// tone arriving while the system event-sounds key is off — the config
// switch still on, the actor gate open — is DROPPED WITHOUT a redial;
// the muted state ends only when the key turns true again, and the next
// tone dials and plays.
func TestPlayer_EventSoundsMuteDropsTonesWithoutRedial(t *testing.T) {
	p, d, r := newSoundTestPlayer()
	p.Start()
	p.Flip()
	if !poll(func() bool { return d.dialCount() == 1 && d.lastWriteCount() == 1 }) {
		t.Fatal("the pre-mute tone never played")
	}
	if !poll(func() bool { return r.monitor() != nil }) {
		t.Fatal("the watcher never started the monitor")
	}
	mon := r.monitor()

	mon.emit("event-sounds: false")
	if !poll(func() bool { return d.lastConn().closeCount() == 1 }) {
		t.Fatal("the event-sounds=false transition never closed the connection (PD-1)")
	}

	p.Flip() // a tone under the system mute — dropped WITHOUT a redial
	staysFalse(t, "a redial under the system mute", func() bool { return d.dialCount() != 1 })
	staysFalse(t, "a played tone under the system mute", func() bool { return d.lastWriteCount() != 1 })

	mon.emit("event-sounds: true") // the muted state ends
	p.Flip()
	if !poll(func() bool { return d.dialCount() == 2 && d.lastWriteCount() == 2 }) {
		t.Fatal("the post-unmute tone never dialed and played — the lazy reconnect is broken")
	}
}

// TestPlayer_EventSoundsMutedAtStartDialsNothing pins the startup arm:
// the start re-read finding event-sounds=false mutes from the first tone
// — nothing ever dials until the key turns true, then the lazy redial
// rides the next tone.
func TestPlayer_EventSoundsMutedAtStartDialsNothing(t *testing.T) {
	p, d, r := newSoundTestPlayer()
	r.sounds = "false"
	p.Start()
	if !poll(func() bool { return r.monitor() != nil }) {
		t.Fatal("the watcher never started the monitor")
	}

	p.Flip() // the system mute held from the start read — dropped, no dial
	staysFalse(t, "a dial under the startup system mute", func() bool { return d.dialCount() != 0 })
	staysFalse(t, "a played tone under the startup system mute", func() bool { return d.lastWriteCount() != 0 })

	r.monitor().emit("event-sounds: true")
	p.Flip()
	if !poll(func() bool { return d.dialCount() == 1 && d.lastWriteCount() == 1 }) {
		t.Fatal("the post-unmute tone never dialed and played")
	}
}

// TestPlayer_GSettingsMissingFallsBack pins the PD-2 degradation: the
// whole gsettings path unavailable — the theme degrades to the XDG
// fallback theme with ONE warn-once WARN and playback carries on.
func TestPlayer_GSettingsMissingFallsBack(t *testing.T) {
	buf := captureSoundLogs(t)
	var mu sync.Mutex
	decodes := make(map[string]int)
	decode := func(theme, event string) (*pcm, error) {
		mu.Lock()
		decodes[theme+"/"+event]++
		mu.Unlock()

		return nil, errDecodeRefused
	}
	p, _, _ := newSoundTestPlayer()
	p.cache = newPCMCache(decode)
	p.runner = &fakeRunner{getErr: errSimulatedGet}
	p.Start()

	if !poll(func() bool {
		mu.Lock()
		defer mu.Unlock()

		return decodes[fallbackTheme+"/"+fsEventBell] > 0
	}) {
		t.Fatal("Start never loaded the fallback theme — the gsettings degradation is broken")
	}
	if got := strings.Count(buf.String(), reasonThemeUnavailable); got != 1 {
		t.Errorf("theme-unavailable WARNs = %d, want exactly 1", got)
	}
}

// TestPlayer_SetAutocorrectEventPrefetch pins the WR-02 re-pin: the
// setter re-pins the event and prefetches the new (theme, event) pair
// asynchronously — the call itself never waits on the prefetch (a tone
// arriving before the prefetch lands plays the synthesized tone).
func TestPlayer_SetAutocorrectEventPrefetch(t *testing.T) {
	release := make(chan struct{})
	var mu sync.Mutex
	decodes := make(map[string]int)
	decode := func(theme, event string) (*pcm, error) {
		if event == rePinnedEvent {
			<-release // the wedged prefetch — the setter must not wait for it
		}
		mu.Lock()
		decodes[theme+"/"+event]++
		mu.Unlock()

		return nil, errDecodeRefused
	}
	p, _, _ := newSoundTestPlayer()
	p.cache = newPCMCache(decode)
	p.Start()

	done := make(chan struct{})
	go func() {
		p.SetAutocorrectEvent(rePinnedEvent)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(pollBudget):
		close(release)
		t.Fatal("SetAutocorrectEvent blocked on the prefetch — the re-pin must be fire-and-forget")
	}
	close(release)
	if !poll(func() bool {
		mu.Lock()
		defer mu.Unlock()

		return decodes[watchedTheme+"/"+rePinnedEvent] > 0
	}) {
		t.Fatal("the re-pin never prefetched the new (theme, event) pair")
	}
}

// TestPlayer_EventRePinReroutesTheTone pins the seam's event routing
// after the rewrite (the old SetAutocorrectEvent argv pin carried over):
// the flip tone rides the schema constant's pitch, and a re-pinned
// autocorrect event changes the NEXT tone — the cache misses on the new
// event, so the synthesized fallback's audibly distinct pitch carries it.
func TestPlayer_EventRePinReroutesTheTone(t *testing.T) {
	p, d, _ := newSoundTestPlayer()
	p.Start()

	p.Flip()
	if !poll(func() bool { return d.lastWriteCount() == 1 }) {
		t.Fatal("the flip tone never played")
	}
	wantFlip, err := synthTone(freqFlip, streamFormat{Rate: int(toneRate), Channels: synthChannels})
	if err != nil {
		t.Fatalf("synth reference tone: %v", err)
	}
	if got := d.lastConn().lastStream().allWrites()[0]; !bytes.Equal(got, wantFlip) {
		t.Error("the flip tone is not the flip pitch — the seam routing broke")
	}

	p.SetAutocorrectEvent(rePinnedEvent)
	p.AutoCorrect()
	if !poll(func() bool { return d.lastWriteCount() == 2 }) {
		t.Fatal("the re-pinned autocorrect tone never played")
	}
	wantAC, err := synthTone(freqAutoCorrect, streamFormat{Rate: int(toneRate), Channels: synthChannels})
	if err != nil {
		t.Fatalf("synth reference tone: %v", err)
	}
	if got := d.lastConn().lastStream().allWrites()[1]; !bytes.Equal(got, wantAC) {
		t.Error("the autocorrect tone is not the autocorrect pitch — the re-pin never rerouted")
	}
}
