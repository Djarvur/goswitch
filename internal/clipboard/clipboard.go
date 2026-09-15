// Package clipboard implements the D-28/D-29 clipboard round-trip rung of
// the ADR-003 ladder: save the user's clipboard byte-exactly, put the
// replacement in through wl-copy (content ONLY via the stdin pipe — argv
// leaks into ps and parses as flags, the security-table anti-pattern), and
// restore the saved content best-effort afterwards (an empty original
// restores through wl-copy --clear; a failed restore is the caller's WARN,
// never an operation error). Clipboard CONTENT never appears in any log at
// any level (D-20/D-21 extension).
package clipboard

import (
	"context"
	"time"
)

// cmdTimeout bounds every wl-clipboard subprocess (T-03-03-05: a wedged
// wl-copy must fail the rung, not hang the correction path). wl-copy forks
// and serves the clipboard in the background — its Run returning after the
// fork is normal, only the fork itself sits inside this budget.
const cmdTimeout = 1500 * time.Millisecond // the owner prototype's subprocess timeout

// Runner executes one wl-clipboard subprocess: the seam the unit corpus
// drives with a recording fake (argv + stdin are the observable surface),
// backed in production by os/exec.CommandContext.
type Runner func(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error)

// Clipboard is the wl-clipboard client of the rung.
type Clipboard struct {
	run Runner
}

// New returns the production client over os/exec (the RED form wires no
// runner yet — the stubs never launch a subprocess).
func New(opts ...func(*Clipboard)) *Clipboard {
	c := &Clipboard{}
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithRunner replaces the subprocess runner — the test seam.
func WithRunner(r Runner) func(*Clipboard) {
	return func(c *Clipboard) {
		c.run = r
	}
}

// Save reads the current clipboard byte-exactly (wl-paste --no-newline).
//
// RED stub (plan 03-03 task 3): returns the empty shape so the corpus fails
// on assertions.
func (c *Clipboard) Save(_ context.Context) (saved []byte, had bool, err error) {
	return nil, false, nil
}

// Set puts content into the clipboard — the bytes travel through the
// subprocess stdin ONLY, never argv (T-03-03-01).
//
// RED stub (plan 03-03 task 3): a no-op so the corpus fails on assertions.
func (c *Clipboard) Set(_ context.Context, _ []byte) error {
	return nil
}

// Clear empties the clipboard (wl-copy --clear) — the restore path of an
// originally-empty buffer (Pitfall 3).
//
// RED stub (plan 03-03 task 3): a no-op so the corpus fails on assertions.
func (c *Clipboard) Clear(_ context.Context) error {
	return nil
}

// Restore puts the saved state back: the bytes for a non-empty original, a
// clear for an empty one. The error is returned to the caller, which
// classifies it as best-effort (a WARN, D-29).
//
// RED stub (plan 03-03 task 3): a no-op so the corpus fails on assertions.
func (c *Clipboard) Restore(_ context.Context, _ []byte, _ bool) error {
	return nil
}
