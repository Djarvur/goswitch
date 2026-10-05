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
	"os"
	"os/exec"
)

// The pinned player binaries (T-07-08-01: fixed names from PATH, arguments
// are flags and event names only — never user content; no shell anywhere).
const (
	binCanberra = "canberra-gtk-play"
	binPaplay   = "paplay"
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
// for every script flip, AutoCorrect for every fired correction. The seam
// fields (lookPath/newProc/cacheDir) are the schema shape the corpus
// drives; the playback path itself arrives with the tracer's GREEN, the
// paplay fallback with task 2.
type Player struct {
	flipEvent string // the flip tone's sound-theme event (the schema constant)
	acEvent   string // the autocorrect tone's event (the document's effective value)
	lookPath  func(string) (string, error)
	newProc   func(string, []string) playProc
	cacheDir  func() (string, error)
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

// Flip plays the flip tone. RED stub — the tracer's GREEN lands the
// canberra fork here.
func (p *Player) Flip() {}

// AutoCorrect plays the autocorrect tone. RED stub — a silent no-op.
func (p *Player) AutoCorrect() {}
