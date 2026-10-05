//nolint:testpackage // drives the unexported runner/LookPath seams — the menu_test in-package corpus precedent
package sound

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

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
// no shell). startErr programs every child's Start failure — the broken
// launch path of the warn-once cell.
type fakeBuilder struct {
	mu       sync.Mutex
	argv     [][]string
	procs    []*fakeProc
	startErr error
}

// newProc is the Player.newProc replacement: record the argv, hand back a
// fresh fake child.
//
//nolint:ireturn // the fake hands the spawned-child interface back (the production seam's shape)
func (b *fakeBuilder) newProc(name string, args []string) playProc {
	proc := &fakeProc{startErr: b.startErr}
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

// TestPlayer_SetAutocorrectEvent pins the hot event re-pin (WR-02): a
// setter call re-arms the autocorrect tone — the next spawn carries the
// NEW event as its argv element — while the flip tone stays on the schema
// constant.
func TestPlayer_SetAutocorrectEvent(t *testing.T) {
	b := &fakeBuilder{}
	p := New(config.DefaultSoundFlipEvent, config.DefaultSoundAutocorrectEvent)
	p.lookPath = func(name string) (string, error) { return "/usr/bin/" + name, nil }
	p.newProc = b.newProc

	const changed = "system-message"
	p.SetAutocorrectEvent(changed)

	p.AutoCorrect()
	p.Flip()

	calls := b.calls()
	if len(calls) != 2 {
		t.Fatalf("runner calls after one autocorrect tone and one flip = %d, want exactly 2", len(calls))
	}
	if want := []string{binCanberra, "-i", changed}; !slices.Equal(calls[0], want) {
		t.Errorf("autocorrect argv after the re-pin = %q, want exactly %q — the re-pinned event rides the next tone", calls[0], want)
	}
	if want := []string{binCanberra, "-i", config.DefaultSoundFlipEvent}; !slices.Equal(calls[1], want) {
		t.Errorf("flip argv after the re-pin = %q, want exactly %q — the flip tone never moves", calls[1], want)
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

// The task-2 corpus: the paplay fallback with bundled tones, the warn-once
// episodes and the full no-pipes/non-blocking discipline (plan 07-08).

// errBinaryMissing is the LookPath double's miss — the "binary not
// installed" state of the corpus cells (the exec.Error shape is irrelevant
// here: the player only reads the error's presence).
var errBinaryMissing = errors.New("executable file not found in $PATH")

// errSimulatedStart is the Start double's failure — the launch-failure
// state of the warn-once cell.
var errSimulatedStart = errors.New("simulated start failure")

// fakeLookPath returns a LookPath double answering found for the names in
// the set — the binary-mount state of the corpus cell.
func fakeLookPath(found map[string]bool) func(string) (string, error) {
	return func(name string) (string, error) {
		if found[name] {
			return "/usr/bin/" + name, nil
		}

		return "", errBinaryMissing
	}
}

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
// buffer — the WARN counts of the episode cells read it.
func captureSoundLogs(t *testing.T) *logBuffer {
	t.Helper()
	buf := &logBuffer{}
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, nil)))

	return buf
}

// pollBudget bounds every poll of an asynchronous effect — orders of
// magnitude above the reaper's real latency, far below any test timeout.
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

func (f *fakeProc) startCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.starts
}

func (f *fakeProc) waitCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.waits
}

// TestPlayer_CanberraPreferred pins the preference (plan 07-08): with both
// players installed BOTH tones ride canberra-gtk-play -i <event> — the
// themed desktop sounds, paplay never consulted.
func TestPlayer_CanberraPreferred(t *testing.T) {
	b := &fakeBuilder{}
	p := New(config.DefaultSoundFlipEvent, config.DefaultSoundAutocorrectEvent)
	p.lookPath = fakeLookPath(map[string]bool{binCanberra: true, binPaplay: true})
	p.newProc = b.newProc

	p.Flip()
	p.AutoCorrect()

	calls := b.calls()
	if len(calls) != 2 {
		t.Fatalf("runner calls = %d, want exactly 2 (both tones)", len(calls))
	}
	if !slices.Equal(calls[0], []string{binCanberra, "-i", config.DefaultSoundFlipEvent}) {
		t.Errorf("flip argv = %q, want exactly [%s -i %s]", calls[0], binCanberra, config.DefaultSoundFlipEvent)
	}
	if !slices.Equal(calls[1], []string{binCanberra, "-i", config.DefaultSoundAutocorrectEvent}) {
		t.Errorf("autocorrect argv = %q, want exactly [%s -i %s]",
			calls[1], binCanberra, config.DefaultSoundAutocorrectEvent)
	}
}

// TestPlayer_PaplayFallback pins the fallback (plan 07-08): canberra
// absent, paplay present — both tones go through paplay over BUNDLED tone
// files materialized in the user's cache dir (RIFF/PCM, owner-only perms,
// audibly distinct pitches), the event name irrelevant to the fork.
func TestPlayer_PaplayFallback(t *testing.T) {
	root := t.TempDir()
	b := &fakeBuilder{}
	p := New(config.DefaultSoundFlipEvent, config.DefaultSoundAutocorrectEvent)
	p.lookPath = fakeLookPath(map[string]bool{binPaplay: true})
	p.newProc = b.newProc
	p.cacheDir = func() (string, error) { return root, nil }

	p.Flip()
	p.AutoCorrect()

	calls := b.calls()
	if len(calls) != 2 {
		t.Fatalf("runner calls = %d, want exactly 2 (both tones through paplay)", len(calls))
	}
	flipWav := filepath.Join(root, "goswitch", "flip.wav")
	acWav := filepath.Join(root, "goswitch", "autocorrect.wav")
	if !slices.Equal(calls[0], []string{binPaplay, flipWav}) {
		t.Errorf("flip argv = %q, want exactly [%s %s]", calls[0], binPaplay, flipWav)
	}
	if !slices.Equal(calls[1], []string{binPaplay, acWav}) {
		t.Errorf("autocorrect argv = %q, want exactly [%s %s]", calls[1], binPaplay, acWav)
	}
	flipData, err := os.ReadFile(flipWav)
	if err != nil {
		t.Fatalf("read flip tone: %v", err)
	}
	acData, err := os.ReadFile(acWav)
	if err != nil {
		t.Fatalf("read autocorrect tone: %v", err)
	}
	for name, data := range map[string][]byte{"flip": flipData, "autocorrect": acData} {
		if !bytes.HasPrefix(data, []byte("RIFF")) {
			t.Errorf("%s tone lacks the RIFF magic: %x", name, data[:4])
		}
		info, err := os.Stat(filepath.Join(root, "goswitch", name+".wav"))
		if err != nil {
			t.Fatalf("stat %s tone: %v", name, err)
		}
		if info.Mode().Perm() != toneFilePerm {
			t.Errorf("%s tone perms = %o, want %o — owner-only", name, info.Mode().Perm(), toneFilePerm)
		}
	}
	if bytes.Equal(flipData, acData) {
		t.Errorf("the two tones are byte-identical — the fallback must carry audibly distinct pitches")
	}
}

// TestPlayer_BothMissingSilent pins the best-effort degradation (plan
// 07-08): neither player installed — zero runner calls ever, and ONE WARN
// for the episode no matter how many tones are asked for (never an error,
// never spam).
func TestPlayer_BothMissingSilent(t *testing.T) {
	buf := captureSoundLogs(t)
	b := &fakeBuilder{}
	p := New(config.DefaultSoundFlipEvent, config.DefaultSoundAutocorrectEvent)
	p.lookPath = fakeLookPath(nil)
	p.newProc = b.newProc

	p.Flip()
	p.AutoCorrect()
	p.Flip()

	if calls := b.calls(); len(calls) != 0 {
		t.Errorf("runner calls with no player installed = %d, want 0 — the sound is off, not broken", len(calls))
	}
	if got := strings.Count(buf.String(), "sound unavailable"); got != 1 {
		t.Errorf("missing-player WARNs = %d, want exactly 1 per episode", got)
	}
}

// TestPlayer_StartFailureWarnOnce pins the episode discipline on a broken
// launch path: every Start fails — every tone is still ATTEMPTED (a no-op
// degradation, never a permanent disable) and exactly ONE WARN fires for
// the episode (per flip would spam the journal of a broken desktop).
func TestPlayer_StartFailureWarnOnce(t *testing.T) {
	buf := captureSoundLogs(t)
	b := &fakeBuilder{startErr: errSimulatedStart}
	p := New(config.DefaultSoundFlipEvent, config.DefaultSoundAutocorrectEvent)
	p.lookPath = fakeLookPath(map[string]bool{binCanberra: true})
	p.newProc = b.newProc

	p.Flip()
	p.AutoCorrect()
	p.Flip()
	p.AutoCorrect()

	if calls := b.calls(); len(calls) != 4 {
		t.Errorf("start attempts = %d, want 4 — a failed launch degrades, never disables", len(calls))
	}
	if got := strings.Count(buf.String(), "sound unavailable"); got != 1 {
		t.Errorf("start-failure WARNs = %d, want exactly 1 per episode", got)
	}
}

// TestPlayer_NonBlocking pins the fire-and-forget contract (plan 07-08,
// the <50 ms budget): Flip returns while the child is still playing — the
// synchronous cost ends at Start, the reaper runs on its own goroutine and
// collects the child once its playback ends.
func TestPlayer_NonBlocking(t *testing.T) {
	release := make(chan struct{})
	child := &fakeProc{onWait: func() { <-release }}
	p := New(config.DefaultSoundFlipEvent, config.DefaultSoundAutocorrectEvent)
	p.lookPath = fakeLookPath(map[string]bool{binCanberra: true})
	p.newProc = func(string, []string) playProc { return child }

	done := make(chan struct{})
	go func() {
		p.Flip()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(pollBudget):
		close(release)
		t.Fatal("Flip blocked on the playing child — the tone call must be fire-and-forget")
	}
	if got := child.startCount(); got != 1 {
		t.Fatalf("starts = %d, want 1 — the tone was launched", got)
	}
	close(release) // the child's playback ends — the reaper collects it
	if !poll(func() bool { return child.waitCount() == 1 }) {
		t.Error("the reaper never collected the child — a leaked tone goroutine")
	}
}

// TestPlayer_WavDeterministic pins the bundled tones' determinism (plan
// 07-08): the synthesis is byte-stable across runs, and an existing
// size-matching cache file is REUSED — never rewritten per flip.
func TestPlayer_WavDeterministic(t *testing.T) {
	generate := func(root string) []byte {
		t.Helper()
		b := &fakeBuilder{}
		p := New(config.DefaultSoundFlipEvent, config.DefaultSoundAutocorrectEvent)
		p.lookPath = fakeLookPath(map[string]bool{binPaplay: true})
		p.newProc = b.newProc
		p.cacheDir = func() (string, error) { return root, nil }
		p.Flip()
		data, err := os.ReadFile(filepath.Join(root, "goswitch", "flip.wav"))
		if err != nil {
			t.Fatalf("read generated tone: %v", err)
		}

		return data
	}

	first := generate(t.TempDir())
	second := generate(t.TempDir())
	if !bytes.Equal(first, second) {
		t.Errorf("tone synthesis is not deterministic — %d vs %d bytes differ", len(first), len(second))
	}

	// An existing valid (size-matching) file survives every later play.
	root := t.TempDir()
	path := filepath.Join(root, "goswitch", "flip.wav")
	if err := os.MkdirAll(filepath.Dir(path), toneDirPerm); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}
	plant := bytes.Repeat([]byte{0x55}, len(first))
	if err := os.WriteFile(path, plant, toneFilePerm); err != nil {
		t.Fatalf("plant cache: %v", err)
	}
	generate(root)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cache after replay: %v", err)
	}
	if !bytes.Equal(got, plant) {
		t.Errorf("the cached tone was rewritten — an existing valid file must be reused, not regenerated")
	}
}
