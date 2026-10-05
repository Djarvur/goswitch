// Package sound plays the daemon's acoustic feedback (plan 07-08, owner
// decision «Звуки при переключении», 07-CONTEXT): a short tone on EVERY
// script flip and a distinct tone on every autocorrect fire, played by a
// best-effort subprocess — the desktop's canberra-gtk-play preferred, a
// paplay fallback over bundled tones — that can never block, error or
// otherwise alter the gesture that triggered it: every failure is ONE WARN
// per episode and a quiet no-op. The tones know nothing about text — no
// word content ever reaches this package or its logs (D-20/D-21).
package sound

import (
	"log/slog"
	"os"
	"os/exec"
	"sync"
)

// The pinned player binaries (T-07-08-01: fixed names from PATH, arguments
// are flags and event names only — never user content; no shell anywhere).
const (
	binCanberra = "canberra-gtk-play"
	binPaplay   = "paplay"
)

// The degradation reasons of the warn-once episodes (the closed vocabulary
// the corpus pins — reasons and binary names only, never user context,
// D-20/D-21). The paplay-fallback and tone-cache reasons arrive with
// task 2's fallback.
const reasonStartFailed = "start-failed"

// The tone-cache shape of the paplay fallback (the bundled tones live in
// the user's cache dir — owner-private directory, owner-only files; the
// GREEN of task 2 owns the synthesis itself).
const (
	cacheSubdir  = "goswitch"
	toneDirPerm  = 0o700
	toneFilePerm = 0o600
)

// playProc is one spawned player child — the corpus seam (the daemon
// launcher's procStarter form): Start brings the player up, Wait reaps it.
type playProc interface {
	Start() error
	Wait() error
}

// newPlayProc builds one player child: the pinned binary with NOTHING
// attached to Stdin/Stdout/Stderr — the no-pipes fork shape (the wl-copy
// precedent, clipboard.go:141-150): a piped descriptor would make a reaper
// wait on a fork-shaped grandchild's write-ends (T-07-08-03).
//
//nolint:ireturn // the seam hands the spawned-child interface back (the emitter() precedent)
func newPlayProc(name string, args []string) playProc {
	//nolint:noctx // deadline-free by design — Start plus a reaped Wait; nothing kills a playing tone (research Q7)
	return exec.Command(name, args...) // Stdin/Stdout/Stderr stay nil — no pipes
}

// Player plays the two tones of the acoustic feedback (plan 07-08): Flip
// for every script flip, AutoCorrect for every fired correction. The
// backend decision is made at the FIRST tone and cached — never a LookPath
// per flip (T-07-08-02); every failure is one WARN per episode and a quiet
// no-op, never an error to the gesture that triggered the tone.
type Player struct {
	flipEvent string // the flip tone's sound-theme event (the schema constant)
	acEvent   string // the autocorrect tone's event (the document's effective value)
	lookPath  func(string) (string, error)
	newProc   func(string, []string) playProc
	cacheDir  func() (string, error)

	mu      sync.Mutex // serializes the decision cache and the warn episode
	decided bool       // the backend decision exists (made at the first tone)
	backend string     // the decided binary, "" when canberra is not installed (the task-2 fallback pending)
	warned  bool       // one start-failure WARN per episode — a broken desktop must not spam per flip
}

// New returns the production player for the two sound-theme events: the
// flip event is the schema constant (config.DefaultSoundFlipEvent — NOT
// configurable, the owner asked for the switch and the distinct
// autocorrect event only), the autocorrect event the document's effective
// value.
func New(flipEvent, autocorrectEvent string) *Player {
	return &Player{
		flipEvent: flipEvent,
		acEvent:   autocorrectEvent,
		lookPath:  exec.LookPath,
		newProc:   newPlayProc,
		cacheDir:  os.UserCacheDir,
	}
}

// Flip plays the flip tone: fire-and-forget — the synchronous cost is the
// argv choice and the child's Start; the playback itself lives in the
// child, reaped by its own goroutine. The hot path of the gesture never
// waits for a sound.
func (p *Player) Flip() {
	p.play(p.flipEvent)
}

// AutoCorrect plays the autocorrect tone — the same fire-and-forget path
// with the document's effective event.
func (p *Player) AutoCorrect() {
	p.play(p.acEvent)
}

// play launches one tone's player child: the canberra path pins the argv
// [canberra-gtk-play -i <event>] — the event is ONE argv element, no shell
// (T-07-08-01). A failed Start is one WARN per episode and a quiet no-op;
// a successful Start hands the child to its reaper goroutine — a later
// non-zero exit (an event the desktop's sound theme lacks) stays silent,
// it is a theme gap, not a daemon failure.
func (p *Player) play(event string) {
	backend := p.backendOf()
	if backend == "" {
		return // the tracer's stub branch: the paplay fallback is task 2 — a silent no-op
	}
	cmd := p.newProc(backend, []string{"-i", event})
	if err := cmd.Start(); err != nil {
		p.warnStart(err)

		return
	}
	go func() { _ = cmd.Wait() }() // reap without a deadline — nothing kills a playing tone
}

// backendOf resolves the playback backend in force: decided at the FIRST
// tone and cached — a LookPath per flip would put a filesystem probe on
// every gesture (T-07-08-02). With canberra absent the tracer's stub
// branch answers "" (the task-2 fallback completes the choice); the caller
// reads the answer without holding the mutex.
func (p *Player) backendOf() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.decided {
		p.decided = true
		if _, err := p.lookPath(binCanberra); err == nil {
			p.backend = binCanberra
		}
	}

	return p.backend
}

// warnStart records one start failure: ONE WARN per episode (the actor's
// warnAppid form — per flip would spam the journal of a broken desktop).
// The record names the reason, the binary and the transport error only.
func (p *Player) warnStart(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.warned {
		return
	}
	p.warned = true
	slog.Warn("sound unavailable", "reason", reasonStartFailed, "binary", binCanberra, "error", err)
}
