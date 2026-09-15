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
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// cmdTimeout bounds every wl-clipboard subprocess (T-03-03-05: a wedged
// wl-copy must fail the rung, not hang the correction path). wl-copy forks
// and serves the clipboard in the background — its Run returning after the
// fork is normal, only the fork itself sits inside this budget.
const cmdTimeout = 1500 * time.Millisecond // the owner prototype's subprocess timeout

// The pinned wl-clipboard flags (man wl-clipboard 2.2.1, RESEARCH §4):
// --no-newline reads the clipboard byte-exactly, --trim-newline keeps the
// pipe's trailing newline out of the copy, --clear empties it.
const (
	flagNoNewline   = "--no-newline"
	flagTrimNewline = "--trim-newline"
	flagClear       = "--clear"
)

// The pinned binary names (T-03-03-04: fixed names from PATH, arguments
// are flags only — never user content).
const (
	binPaste = "wl-paste"
	binCopy  = "wl-copy"
)

// Runner executes one wl-clipboard subprocess: the seam the unit corpus
// drives with a recording fake (argv + stdin are the observable surface),
// backed in production by os/exec.CommandContext.
type Runner func(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error)

// Clipboard is the wl-clipboard client of the rung.
type Clipboard struct {
	run Runner
}

// New returns the production client over os/exec; the options replace the
// runner (the test seam).
func New(opts ...func(*Clipboard)) *Clipboard {
	c := &Clipboard{run: execRunner}
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
// A non-zero wl-paste exit is the EMPTY clipboard — a state, not a failure:
// had=false, no error, and the restore later clears instead of pasting
// (Pitfall 3). Everything else (a cancelled context, a missing binary) is
// an error the caller turns into the rung's quiet refusal.
func (c *Clipboard) Save(ctx context.Context) (saved []byte, had bool, err error) {
	out, err := c.call(ctx, binPaste, []string{flagNoNewline}, nil)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, false, nil // wl-paste's empty-clipboard exit
		}

		return nil, false, fmt.Errorf("clipboard save: %w", err)
	}

	return out, true, nil
}

// Set puts content into the clipboard: the bytes travel through the
// subprocess STDIN pipe only, never argv (T-03-03-01 — argv is visible in
// ps and user content parses as flags). wl-copy forks and serves the
// clipboard in the background, so the subprocess finishing is its normal
// shape, not an error.
func (c *Clipboard) Set(ctx context.Context, content []byte) error {
	if _, err := c.call(ctx, binCopy, []string{flagTrimNewline}, content); err != nil {
		return fmt.Errorf("clipboard set: %w", err)
	}

	return nil
}

// Clear empties the clipboard (wl-copy --clear) — the restore path of an
// originally-empty buffer (Pitfall 3): an empty original must not come back
// as pasted text.
func (c *Clipboard) Clear(ctx context.Context) error {
	if _, err := c.call(ctx, binCopy, []string{flagClear}, nil); err != nil {
		return fmt.Errorf("clipboard clear: %w", err)
	}

	return nil
}

// Restore puts the saved state back: the bytes for a non-empty original, a
// clear for an empty one. The error is returned to the caller, which
// classifies it as best-effort (a WARN, D-29 — never an operation error).
func (c *Clipboard) Restore(ctx context.Context, saved []byte, had bool) error {
	if !had {
		return c.Clear(ctx)
	}

	return c.Set(ctx, saved)
}

// call bounds one subprocess with the package deadline (T-03-03-05) and
// hands it to the runner — the context the runner sees always carries a
// cancellation deadline.
func (c *Clipboard) call(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	return c.run(ctx, name, args, stdin)
}

// execRunner runs one subprocess through os/exec: the content rides the
// stdin pipe (never argv), and the caller's deadline-bounded context kills
// a wedged process (CommandContext). wl-copy is the fork-shaped exception:
// it forks a background grandchild that serves the clipboard indefinitely,
// and a piped stdout/stderr would make Run() wait for that grandchild's
// pipe write-ends to close — a deadlock (live finding of the first
// select-clipboard run, 2026-09-15: the rung hung the full watchdog
// budget). wl-copy is silent on success, so it gets NO pipes and its error
// carries the exit status alone; wl-paste exits when done and its stdout is
// captured byte-exactly.
func execRunner(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	if name == binCopy {
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
		}

		return nil, nil
	}
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}

	return out.Bytes(), nil
}
