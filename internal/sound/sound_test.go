//nolint:testpackage // drives the unexported runner/LookPath seams — the menu_test in-package corpus precedent
package sound

import (
	"os/exec"
	"slices"
	"sync"
	"testing"

	"github.com/Djarvur/goswitch/internal/config"
)

// fakeProc is one child-process double of the corpus (the daemon
// launcher's procStarter fake): Start records the launch and returns the
// programmed result, Wait signals the reap — optionally blocking until
// released, the playing child of the non-blocking pin.
type fakeProc struct {
	mu       sync.Mutex
	starts   int
	waits    int
	startErr error
	onWait   func()
}

// Start records one launch and answers the programmed result.
func (f *fakeProc) Start() error {
	f.mu.Lock()
	f.starts++
	f.mu.Unlock()

	return f.startErr
}

// Wait records one reap and plays the blocking hook first.
func (f *fakeProc) Wait() error {
	if hook := f.waitHook(); hook != nil {
		hook()
	}
	f.mu.Lock()
	f.waits++
	f.mu.Unlock()

	return nil
}

func (f *fakeProc) waitHook() func() {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.onWait
}

// fakeBuilder is the process-builder double: every spawn is recorded as its
// full argv — the observable surface the corpus pins (one event element,
// no shell).
type fakeBuilder struct {
	mu    sync.Mutex
	argv  [][]string
	procs []*fakeProc
}

// newProc is the Player.newProc replacement: record the argv, hand back a
// fresh fake child.
//
//nolint:ireturn // the fake hands the spawned-child interface back (the production seam's shape)
func (b *fakeBuilder) newProc(name string, args []string) playProc {
	proc := &fakeProc{}
	b.mu.Lock()
	b.argv = append(b.argv, append([]string{name}, args...))
	b.procs = append(b.procs, proc)
	b.mu.Unlock()

	return proc
}

// calls snapshots every spawned argv in order.
func (b *fakeBuilder) calls() [][]string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return append([][]string(nil), b.argv...)
}

// TestPlayer_FlipCanberraArgv pins the tracer's wire form (plan 07-08):
// one flip tone goes through canberra-gtk-play with the event as ONE
// argv element — no shell, no pipes, the fork-shaped no-pipes form.
func TestPlayer_FlipCanberraArgv(t *testing.T) {
	b := &fakeBuilder{}
	p := New(config.DefaultSoundFlipEvent, config.DefaultSoundAutocorrectEvent)
	p.lookPath = func(name string) (string, error) { return "/usr/bin/" + name, nil }
	p.newProc = b.newProc

	p.Flip()

	calls := b.calls()
	if len(calls) != 1 {
		t.Fatalf("runner calls after one flip = %d, want exactly 1", len(calls))
	}
	want := []string{binCanberra, "-i", config.DefaultSoundFlipEvent}
	if !slices.Equal(calls[0], want) {
		t.Errorf("flip argv = %q, want exactly %q — the event is one argv element, never a shell", calls[0], want)
	}
}

// TestPlayer_AutoCorrectCanberraArgv pins the second tone's wire form: the
// autocorrect tone rides the SAME no-pipes canberra path with the
// document's effective event.
func TestPlayer_AutoCorrectCanberraArgv(t *testing.T) {
	b := &fakeBuilder{}
	p := New(config.DefaultSoundFlipEvent, config.DefaultSoundAutocorrectEvent)
	p.lookPath = func(name string) (string, error) { return "/usr/bin/" + name, nil }
	p.newProc = b.newProc

	p.AutoCorrect()

	calls := b.calls()
	if len(calls) != 1 {
		t.Fatalf("runner calls after one autocorrect tone = %d, want exactly 1", len(calls))
	}
	want := []string{binCanberra, "-i", config.DefaultSoundAutocorrectEvent}
	if !slices.Equal(calls[0], want) {
		t.Errorf("autocorrect argv = %q, want exactly %q", calls[0], want)
	}
}

// TestPlayer_NoPipesForm pins the production builder's descriptor
// discipline (T-07-08-03, the wl-copy live finding): the spawned child
// carries NOTHING on Stdin/Stdout/Stderr — a piped descriptor would hang
// the reaper on a fork-shaped grandchild's write-ends.
func TestPlayer_NoPipesForm(t *testing.T) {
	cmd, ok := newPlayProc(binCanberra, []string{"-i", "bell"}).(*exec.Cmd)
	if !ok {
		t.Fatalf("production builder returned %T, want *exec.Cmd", cmd)
	}
	if cmd.Stdin != nil || cmd.Stdout != nil || cmd.Stderr != nil {
		t.Errorf("pipes attached: stdin=%v stdout=%v stderr=%v — want all nil (the no-pipes fork shape)",
			cmd.Stdin, cmd.Stdout, cmd.Stderr)
	}
}
