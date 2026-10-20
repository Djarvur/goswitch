package clipboard_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/Djarvur/goswitch/internal/clipboard"
)

// ownerClip is the stand-in for the user's clipboard content: a probe value
// with no edge whitespace, so runCmd-style trimming cannot eat part of it.
const ownerClip = "goswitch-e2e-owner-clip"

// replacement is the rung's replacement payload (the converted text the
// correction would paste).
const replacement = "привет"

// binCopyName is the recorded name of the copy binary (goconst).
const binCopyName = "wl-copy"

// clipCall is one recorded subprocess invocation of the fake runner: the
// argv and the stdin bytes are the observable surface (T-03-03-01: the
// content must ride stdin, never argv).
type clipCall struct {
	name           string
	args           []string
	stdin          []byte
	ctxHasDeadline bool
}

// fakeRunner records every subprocess the client launches and answers
// wl-paste with the configured clipboard state (reply, exitErr — the
// empty-clipboard shape is wl-paste's non-zero exit).
type fakeRunner struct {
	mu      sync.Mutex
	calls   []clipCall
	reply   []byte
	exitErr error
	// hangUntilCtx makes the runner honor context cancellation like a real
	// wedged subprocess (T-03-03-05).
	hangUntilCtx bool
}

// run records the invocation and answers it.
func (f *fakeRunner) run(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error) {
	_, hasDeadline := ctx.Deadline()
	f.mu.Lock()
	f.calls = append(f.calls, clipCall{name: name, args: args, stdin: stdin, ctxHasDeadline: hasDeadline})
	hang := f.hangUntilCtx
	exitErr := f.exitErr
	reply := f.reply
	f.mu.Unlock()
	if hang {
		// The real deadline-kill shape (CR-03): the subprocess is RUNNING
		// when the context's deadline fires — os/exec's Cancel SIGKILLs it
		// and Wait reports the genuine *exec.ExitError ("signal: killed",
		// ExitCode -1). The child must be started BEFORE the wait so the
		// kill — never a start-time refusal — shapes the returned error.
		cmd := exec.CommandContext(ctx, "sleep", "5")
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("subprocess start: %w", err)
		}
		<-ctx.Done()
		waitErr := cmd.Wait()
		if waitErr != nil {
			return nil, fmt.Errorf("subprocess wait: %w", waitErr)
		}

		return nil, nil
	}
	if name == "wl-paste" && exitErr != nil {
		return nil, exitErr
	}

	return reply, nil
}

// snapshot copies the recorded calls under the guard.
func (f *fakeRunner) snapshot() []clipCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	return slices.Clone(f.calls)
}

// newClient builds the client over the fake runner.
func newClient(f *fakeRunner) *clipboard.Clipboard {
	return clipboard.New(clipboard.WithRunner(f.run))
}

// TestClipboard_SaveSetRestore pins the round-trip wire discipline of the
// rung (D-28, Pitfall 3, T-03-03-01): Save reads wl-paste --no-newline
// byte-exactly; Set writes wl-copy --trim-newline with the content in STDIN
// ONLY — the payload must never appear in argv (ps leakage / flag parsing);
// Restore re-writes the saved bytes through the same stdin path, and an
// originally-empty clipboard (wl-paste exit != 0) restores through
// wl-copy --clear instead of pasting anything.
func TestClipboard_SaveSetRestore(t *testing.T) {
	t.Parallel()

	t.Run("round trip restores the owner bytes", func(t *testing.T) {
		t.Parallel()
		roundTripRestoresOwnerBytes(t)
	})

	t.Run("empty clipboard clears on restore", func(t *testing.T) {
		t.Parallel()
		emptyClipboardClearsOnRestore(t)
	})
}

// roundTripRestoresOwnerBytes is the non-empty round-trip body: the three
// subprocesses in order with the exact flags, the content riding stdin.
func roundTripRestoresOwnerBytes(t *testing.T) {
	t.Helper()
	f := &fakeRunner{reply: []byte(ownerClip)}
	c := newClient(f)

	saved, had, err := c.Save(context.Background())
	if err != nil {
		t.Fatalf("Save() err = %v, want nil", err)
	}
	if !had {
		t.Fatal("Save() had = false over a non-empty clipboard, want true")
	}
	if string(saved) != ownerClip {
		t.Fatalf("Save() = %q, want the byte-exact %q", saved, ownerClip)
	}
	if err := c.Set(context.Background(), []byte(replacement)); err != nil {
		t.Fatalf("Set() err = %v, want nil", err)
	}
	if err := c.Restore(context.Background(), saved, had); err != nil {
		t.Fatalf("Restore() err = %v, want nil", err)
	}

	assertRoundTripWire(t, f.snapshot())
}

// assertRoundTripWire pins the three recorded subprocesses: exact names and
// flags, the content in stdin and never in argv (T-03-03-01).
func assertRoundTripWire(t *testing.T, calls []clipCall) {
	t.Helper()
	if len(calls) != 3 {
		t.Fatalf("subprocess calls = %d, want 3 (paste, copy, copy); got %+v", len(calls), calls)
	}
	paste, set, restore := calls[0], calls[1], calls[2]
	if paste.name != "wl-paste" || !slices.Equal(paste.args, []string{"--no-newline"}) {
		t.Errorf("Save ran %s %v, want wl-paste [--no-newline] (byte-exact read, Pitfall 3)", paste.name, paste.args)
	}
	if set.name != binCopyName || !slices.Equal(set.args, []string{"--trim-newline"}) {
		t.Errorf("Set ran %s %v, want wl-copy [--trim-newline]", set.name, set.args)
	}
	if string(set.stdin) != replacement {
		t.Errorf("Set stdin = %q, want the replacement there", set.stdin)
	}
	if slices.Contains(set.args, replacement) {
		t.Errorf("Set argv %q contains the clipboard content — stdin only, never argv (T-03-03-01)", set.args)
	}
	if restore.name != binCopyName || !slices.Equal(restore.args, []string{"--trim-newline"}) {
		t.Errorf("Restore ran %s %v, want wl-copy [--trim-newline]", restore.name, restore.args)
	}
	if string(restore.stdin) != ownerClip {
		t.Errorf("Restore stdin = %q, want the saved owner bytes", restore.stdin)
	}
}

// emptyClipboardClearsOnRestore is the empty-original body: wl-paste's
// non-zero exit reads as an empty state (no error), and the restore CLEARS.
func emptyClipboardClearsOnRestore(t *testing.T) {
	t.Helper()
	// The empty-buffer shape: wl-paste's non-zero exit — a genuine
	// *exec.ExitError with a POSITIVE exit code, forged from a real child
	// (the zero-value ExitError carries no ProcessState and cannot answer
	// ExitCode).
	f := &fakeRunner{exitErr: cleanPasteExitErr(t)}
	c := newClient(f)

	saved, had, err := c.Save(context.Background())
	if err != nil {
		t.Fatalf("Save() over an empty clipboard err = %v, want nil (empty is a state, not a failure)", err)
	}
	if had {
		t.Fatal("Save() had = true over an empty clipboard, want false")
	}
	if len(saved) != 0 {
		t.Fatalf("Save() = %q over an empty clipboard, want no bytes", saved)
	}
	if err := c.Restore(context.Background(), saved, had); err != nil {
		t.Fatalf("Restore() err = %v, want nil", err)
	}

	calls := f.snapshot()
	if len(calls) != 2 {
		t.Fatalf("subprocess calls = %d, want 2 (paste, clear); got %+v", len(calls), calls)
	}
	clearCall := calls[1]
	if clearCall.name != binCopyName || !slices.Equal(clearCall.args, []string{"--clear"}) {
		t.Errorf("Restore ran %s %v over an empty original, want wl-copy [--clear] (Pitfall 3)",
			clearCall.name, clearCall.args)
	}
	if len(clearCall.stdin) != 0 {
		t.Errorf("Restore stdin = %q on clear, want none", clearCall.stdin)
	}
}

// cleanPasteExitErr is wl-paste's genuine empty-clipboard exit: a real
// subprocess exiting non-zero (exit status 1) — the state Save must keep
// reading as EMPTY. Forged from a real child because the wire shape
// matters: *exec.ExitError with a positive ExitCode.
func cleanPasteExitErr(t *testing.T) error {
	t.Helper()
	err := exec.CommandContext(context.Background(), "false").Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() <= 0 {
		t.Fatalf("false exit = %v, want *exec.ExitError with a positive exit code", err)
	}

	return exitErr
}

// killedPasteErr reproduces the real deadline-kill shape (CR-03): a
// subprocess KILLED by exec.CommandContext's Cancel reports
// *exec.ExitError "signal: killed" with ExitCode -1 — the exact type and
// state the rung's own 1500 ms deadline produces against a wedged
// wl-paste. Forged from a real killed child (start first, cancel after):
// the kill must shape the error, never a start-time refusal.
func killedPasteErr(t *testing.T) error {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "sleep", "5")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start the kill-shape child: %v", err)
	}
	cancel()
	err := cmd.Wait()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() >= 0 {
		t.Fatalf("killed child error = %v, want *exec.ExitError with a negative exit code (signal kill)", err)
	}

	return exitErr
}

// TestClipboard_KilledNotEmpty pins the CR-03 data-loss guard: a wl-paste
// KILLED by a signal — above all the rung's own 1500 ms deadline SIGKILL —
// must surface as an ERROR, never as the empty clipboard. Read as empty,
// the restore would run wl-copy --clear over a non-empty clipboard the
// hung subprocess simply failed to report.
func TestClipboard_KilledNotEmpty(t *testing.T) {
	t.Parallel()

	t.Run("signal-killed wl-paste is an error, not an empty read", func(t *testing.T) {
		t.Parallel()
		f := &fakeRunner{exitErr: killedPasteErr(t)}
		c := newClient(f)

		saved, had, err := c.Save(context.Background())
		if err == nil {
			t.Fatalf("Save() over a killed wl-paste = (%q, %v, nil), want an error"+
				" — a signal kill must abort the round-trip, never read as empty", saved, had)
		}
		if had {
			t.Error("Save() had = true over a killed wl-paste, want false")
		}
	})
}

// TestClipboard_DeadlineKillNotEmpty pins the CR-03 production path
// end-to-end: a wl-paste that hangs past the package's own 1500 ms budget
// is SIGKILLed by exec.CommandContext, and the kill surfaces through the
// PRODUCTION runner as an error — the exact misclassification that would
// clear the user's clipboard on restore. Serial: t.Setenv cannot run in
// parallel tests, and the case spends the real cmdTimeout budget.
func TestClipboard_DeadlineKillNotEmpty(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "wl-paste")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexec /bin/sleep 5\n"), 0o755); err != nil {
		t.Fatalf("write the hanging wl-paste: %v", err)
	}
	t.Setenv("PATH", dir)

	c := clipboard.New()
	saved, had, err := c.Save(context.Background())
	if err == nil {
		t.Fatalf("Save() over the deadline-killed wl-paste = (%q, %v, nil), want an error"+
			" — the rung's own kill must never read as the empty clipboard", saved, had)
	}
	if had {
		t.Error("Save() had = true over the deadline-killed wl-paste, want false")
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Errorf("Save() error = %v, want it to carry the killed subprocess's *exec.ExitError", err)
	}
}

// TestClipboard_Timeouts pins the DoS guard of the rung (T-03-03-05): every
// subprocess call carries its own context deadline (a wedged wl-clipboard
// must fail inside the budget, never hang the correction), and a cancelled
// context surfaces as an ERROR — never as a silent had=false empty read,
// which would throw the user's clipboard away.
func TestClipboard_Timeouts(t *testing.T) {
	t.Parallel()

	t.Run("every call carries a deadline", func(t *testing.T) {
		t.Parallel()
		f := &fakeRunner{reply: []byte(ownerClip)}
		c := newClient(f)

		saved, had, err := c.Save(context.Background())
		if err != nil || !had {
			t.Fatalf("Save() = (%q, %v, %v), want a successful read", saved, had, err)
		}
		if err := c.Set(context.Background(), []byte(replacement)); err != nil {
			t.Fatalf("Set() err = %v, want nil", err)
		}
		if err := c.Restore(context.Background(), saved, had); err != nil {
			t.Fatalf("Restore() err = %v, want nil", err)
		}
		for i, call := range f.snapshot() {
			if !call.ctxHasDeadline {
				t.Errorf("call %d (%s) has no context deadline — every subprocess must be bounded", i, call.name)
			}
		}
	})

	t.Run("cancelled context is an error, not an empty read", func(t *testing.T) {
		t.Parallel()
		f := &fakeRunner{hangUntilCtx: true}
		c := newClient(f)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, _, err := c.Save(ctx); err == nil {
			t.Error("Save(cancelled ctx) err = nil, want the context error surfaced")
		}
		if err := c.Set(ctx, []byte(replacement)); err == nil {
			t.Error("Set(cancelled ctx) err = nil, want the context error surfaced")
		}
		if err := c.Restore(ctx, nil, false); err == nil {
			t.Error("Restore(cancelled ctx) err = nil, want the context error surfaced")
		}
	})
}

// TestClipboard_NewProductionRunner proves the production constructor wires
// a runnable client: no options means the os/exec runner — Save actually
// launches wl-paste (and fails cleanly when the binary is absent from PATH,
// which the test enforces by clearing it: a missing binary must be an
// error, never a silent success). Serial: t.Setenv cannot run in parallel
// tests.
func TestClipboard_NewProductionRunner(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // no wl-paste in an empty dir

	c := clipboard.New()
	if _, _, err := c.Save(context.Background()); err == nil {
		t.Error("Save() with wl-paste absent from PATH err = nil, want a named subprocess error")
	}
	if err := c.Set(context.Background(), []byte(replacement)); err == nil {
		t.Error("Set() with wl-copy absent from PATH err = nil, want a named subprocess error")
	}
	if err := c.Clear(context.Background()); err == nil {
		t.Error("Clear() with wl-copy absent from PATH err = nil, want a named subprocess error")
	}
	// The absent-binary error must carry the command name for diagnostics.
	if _, _, err := c.Save(context.Background()); err == nil || !strings.Contains(err.Error(), "wl-paste") {
		t.Errorf("Save() error = %v, want it to name wl-paste", err)
	}
}

// binPasteName is the recorded name of the paste binary (the binCopyName
// idiom).
const binPasteName = "wl-paste"

// TestClipboard_Primary pins the PRIMARY read primitive of plan 10-02
// (ADR-008 D5: the wezterm leg reads PRIMARY, never CLIPBOARD): the read
// runs wl-paste with EXACTLY [--primary, --no-newline] — pinned literal
// flags, no payload ever rides argv — an empty selection is the same
// non-error state as Save's, a genuine failure is a wrapped error, every
// call carries the rung's deadline, and a signal-killed wl-paste is an
// ERROR, never an empty read (the CR-03 misclassification would make the
// caller trust a selection the subprocess merely failed to report).
func TestClipboard_Primary(t *testing.T) {
	t.Parallel()

	t.Run("reads wl-paste --primary --no-newline", func(t *testing.T) {
		t.Parallel()
		f := &fakeRunner{reply: []byte(ownerClip)}
		c := newClient(f)

		saved, had, err := c.Primary(context.Background())
		if err != nil {
			t.Fatalf("Primary() err = %v, want nil", err)
		}
		if !had {
			t.Fatal("Primary() had = false over a non-empty selection, want true")
		}
		if string(saved) != ownerClip {
			t.Fatalf("Primary() = %q, want the byte-exact %q", saved, ownerClip)
		}
		calls := f.snapshot()
		if len(calls) != 1 {
			t.Fatalf("subprocess calls = %d, want 1; got %+v", len(calls), calls)
		}
		call := calls[0]
		if call.name != binPasteName || !slices.Equal(call.args, []string{"--primary", "--no-newline"}) {
			t.Errorf("Primary ran %s %v, want %s [--primary --no-newline] (pinned literal flags)",
				call.name, call.args, binPasteName)
		}
		if len(call.stdin) != 0 {
			t.Errorf("Primary stdin = %q, want none — a read carries no payload", call.stdin)
		}
	})

	t.Run("empty selection is a state, not an error", func(t *testing.T) {
		t.Parallel()
		f := &fakeRunner{exitErr: cleanPasteExitErr(t)}
		c := newClient(f)

		saved, had, err := c.Primary(context.Background())
		if err != nil {
			t.Fatalf("Primary() over an empty selection err = %v, want nil (empty is a state, not a failure)", err)
		}
		if had {
			t.Error("Primary() had = true over an empty selection, want false")
		}
		if len(saved) != 0 {
			t.Errorf("Primary() = %q over an empty selection, want no bytes", saved)
		}
	})

	t.Run("signal-killed wl-paste is a wrapped error", func(t *testing.T) {
		t.Parallel()
		testPrimaryKilledNotEmpty(t)
	})
}

// testPrimaryKilledNotEmpty is the killed-wl-paste body: the CR-03 guard —
// a signal kill is an ERROR carrying the clipboard primary prefix, never
// the empty selection.
func testPrimaryKilledNotEmpty(t *testing.T) {
	t.Helper()

	f := &fakeRunner{exitErr: killedPasteErr(t)}
	c := newClient(f)

	saved, had, err := c.Primary(context.Background())
	if err == nil {
		t.Fatalf("Primary() over a killed wl-paste = (%q, %v, nil), want an error"+
			" — a signal kill must never read as the empty selection", saved, had)
	}
	if !strings.Contains(err.Error(), "clipboard primary") {
		t.Errorf("error %q does not carry the clipboard primary prefix", err)
	}
	if had {
		t.Error("Primary() had = true over a failed read, want false")
	}
}

// TestClipboard_PrimaryKilledNotEmpty pins the CR-03 guard over the PRIMARY
// read (the Save corpus's KilledNotEmpty mirror): a wl-paste KILLED by a
// signal — the rung's own deadline included — is an ERROR, never the empty
// selection; read as empty, the caller would trust a selection the
// subprocess merely failed to report.
func TestClipboard_PrimaryKilledNotEmpty(t *testing.T) {
	t.Parallel()

	testPrimaryKilledNotEmpty(t)
}

// TestClipboard_PrimaryDeadline pins the DoS guard over the PRIMARY read
// (T-03-03-05, the Save corpus's deadline discipline): the read is bounded
// by the same rung deadline — a wedged wl-paste fails inside the budget,
// never hangs the correction path.
func TestClipboard_PrimaryDeadline(t *testing.T) {
	t.Parallel()

	f := &fakeRunner{reply: []byte(ownerClip)}
	c := newClient(f)
	if _, _, err := c.Primary(context.Background()); err != nil {
		t.Fatalf("Primary() err = %v, want nil", err)
	}
	for i, call := range f.snapshot() {
		if !call.ctxHasDeadline {
			t.Errorf("call %d (%s) has no context deadline — every subprocess must be bounded",
				i, call.name)
		}
	}
}
