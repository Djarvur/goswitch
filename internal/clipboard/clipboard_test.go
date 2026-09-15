package clipboard_test

import (
	"context"
	"fmt"
	"os/exec"
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
	defer f.mu.Unlock()
	f.calls = append(f.calls, clipCall{name: name, args: args, stdin: stdin, ctxHasDeadline: hasDeadline})
	if f.hangUntilCtx {
		<-ctx.Done()

		return nil, fmt.Errorf("subprocess canceled: %w", ctx.Err())
	}
	if name == "wl-paste" && f.exitErr != nil {
		return nil, f.exitErr
	}

	return f.reply, nil
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
	// *exec.ExitError, the exact type Save classifies as "empty".
	f := &fakeRunner{exitErr: &exec.ExitError{}}
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
