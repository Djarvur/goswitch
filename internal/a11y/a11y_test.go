//nolint:testpackage // drives the unexported series state and set-builder — the sound_test in-package corpus precedent
package a11y

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"
)

// The corpus' timing constants: the async series converges in
// milliseconds on fakes, the deadlines only bound a wedged run's report.
const (
	testDeadline  = 2 * time.Second          // the convergence deadline — polling, never a sleep-based race
	testSettleGap = 100 * time.Millisecond // the diff-gate's no-op settle window
)

// fakeRunner is the recording double of the Runner seam (the argv is the
// observable surface): every invocation's full argv is recorded, the get
// answer is canned, and blockGet, when non-nil, stalls the FIRST get
// until released — the non-blocking pin's wedged probe.
type fakeRunner struct {
	mu       sync.Mutex
	argvs    [][]string
	gets     int
	sets     int
	value    string
	err      error
	blockGet chan struct{}
}

// run is the Runner seam double: the argv lands in the record, gets
// answer the canned value/error, sets answer clean.
func (f *fakeRunner) run(ctx context.Context, name string, args []string) ([]byte, error) {
	f.mu.Lock()
	f.argvs = append(f.argvs, append([]string{name}, args...))
	isGet := len(args) > 0 && args[0] == verbGet
	if isGet {
		f.gets++
	} else {
		f.sets++
	}
	value, err, release := f.value, f.err, f.blockGet
	if release != nil {
		f.blockGet = nil // the first get only
	}
	f.mu.Unlock()

	if release != nil {
		<-release
	}
	if isGet {
		return []byte(value), err
	}

	return nil, nil
}

// counts returns the recorded get/set call totals.
func (f *fakeRunner) counts() (gets, sets int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.gets, f.sets
}

// total returns the recorded invocation count across both verbs.
func (f *fakeRunner) total() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return len(f.argvs)
}

// argvSnapshot returns a copy of the recorded argv list.
func (f *fakeRunner) argvSnapshot() [][]string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return slices.Clone(f.argvs)
}

// failReads programs the get answer's failure (the fail-toward-desired
// corpus).
func (f *fakeRunner) failReads(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.err = err
}

// fakeStatus records the belt's activations (the StatusSetter seam
// double): a count per call and a programmable refusal.
type fakeStatus struct {
	mu    sync.Mutex
	calls int
	err   error
}

// set is the StatusSetter seam double: count the call, answer the canned
// error.
func (f *fakeStatus) set(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls++

	return f.err
}

// total returns the recorded belt-call count.
func (f *fakeStatus) total() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.calls
}

// failSets programs the belt's refusal.
func (f *fakeStatus) fail(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.err = err
}

// waitConverged polls the reconciler's applied state until it equals
// want — the series is async by contract, so tests converge on the
// observable state instead of sleeping. Fails the test on timeout.
func waitConverged(t *testing.T, r *Reconciler, want bool) {
	t.Helper()
	for deadline := time.Now().Add(testDeadline); time.Now().Before(deadline); {
		r.mu.Lock()
		got := r.lastApplied
		r.mu.Unlock()
		if got == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("series did not converge to lastApplied=%v within the deadline", want)
}

// argvLiterals is the closed vocabulary every recorded argv element must
// belong to — the argv discipline's test form: the config's app list has
// no path into this set, hence none into any command (T-08-03-01).
func argvLiterals() map[string]bool {
	return map[string]bool{
		binGSettings:            true,
		verbGet:                 true,
		verbSet:                 true,
		schemaGnomeInterface:    true,
		keyToolkitAccessibility: true,
		valTrue:                 true,
	}
}

// TestA11yWireLiterals pins the literal VALUES behind the argv
// vocabulary: a renamed constant would silently change the command the
// daemon renders.
func TestA11yWireLiterals(t *testing.T) {
	if binGSettings != "gsettings" {
		t.Errorf("binGSettings = %q, want gsettings", binGSettings)
	}
	if schemaGnomeInterface != "org.gnome.desktop.interface" {
		t.Errorf("schemaGnomeInterface = %q, want org.gnome.desktop.interface", schemaGnomeInterface)
	}
	if keyToolkitAccessibility != "toolkit-accessibility" {
		t.Errorf("keyToolkitAccessibility = %q, want toolkit-accessibility", keyToolkitAccessibility)
	}
	if valTrue != "true" {
		t.Errorf("valTrue = %q, want true", valTrue)
	}
	if verbGet != "get" || verbSet != "set" {
		t.Errorf("verbs = %q/%q, want get/set", verbGet, verbSet)
	}
}

// TestA11y_ActivateSetsKeyWhenFalse pins the activation series' full
// observable shape: exactly one read-verify probe and exactly one set,
// each with its literal argv (the key was false — the set must fire).
func TestA11y_ActivateSetsKeyWhenFalse(t *testing.T) {
	runner := &fakeRunner{value: "false\n"}
	r := New(runner.run, (&fakeStatus{}).set)

	r.Apply(true)
	waitConverged(t, r, true)

	gets, sets := runner.counts()
	if gets != 1 {
		t.Errorf("get calls = %d, want exactly 1", gets)
	}
	if sets != 1 {
		t.Errorf("set calls = %d, want exactly 1", sets)
	}
	wantGet := []string{binGSettings, verbGet, schemaGnomeInterface, keyToolkitAccessibility}
	wantSet := []string{binGSettings, verbSet, schemaGnomeInterface, keyToolkitAccessibility, valTrue}
	if !slices.Equal(runner.argvs[0], wantGet) {
		t.Errorf("get argv = %v, want %v", runner.argvs[0], wantGet)
	}
	if !slices.Equal(runner.argvs[1], wantSet) {
		t.Errorf("set argv = %v, want %v", runner.argvs[1], wantSet)
	}
}

// TestA11y_AlreadyTrueNoSet pins read-verify-then-set (research Pitfall
// 4): the key already true means the probe ran and NO set fired — dconf
// churn from a reconcile is forbidden.
func TestA11y_AlreadyTrueNoSet(t *testing.T) {
	runner := &fakeRunner{value: "true\n"}
	r := New(runner.run, (&fakeStatus{}).set)

	r.Apply(true)
	waitConverged(t, r, true)

	gets, sets := runner.counts()
	if gets != 1 {
		t.Errorf("get calls = %d, want exactly 1 (the verify probe must run)", gets)
	}
	if sets != 0 {
		t.Errorf("set calls = %d, want 0 — the key is already true (no dconf churn)", sets)
	}
}

// TestA11y_DeactivateNeverTouchesKey pins D-8-4: deactivation runs ZERO
// subprocesses — neither the probe nor the set — both on a fresh
// reconciler and after an activation (a manually enabled key is
// indistinguishable from a goswitch-enabled one; the only revert is the
// uninstaller's).
func TestA11y_DeactivateNeverTouchesKey(t *testing.T) {
	runner := &fakeRunner{value: "false\n"}
	r := New(runner.run, (&fakeStatus{}).set)

	r.Apply(false)
	gets, sets := runner.counts()
	if gets != 0 || sets != 0 {
		t.Fatalf("fresh deactivation ran subprocesses: gets=%d sets=%d, want 0/0", gets, sets)
	}

	r.Apply(true)
	waitConverged(t, r, true)
	before := runner.total()
	r.Apply(false)
	waitConverged(t, r, false)

	if after := runner.total(); after != before {
		t.Errorf("deactivation changed the runner call count: %d → %d, want unchanged", before, after)
	}
}

// TestA11y_FixedArgvOnlyLiterals pins the argv discipline end to end:
// every argv the series ever renders consists of package literals alone
// (T-08-03-01 — the config's app list has no path into argv), and the
// diff gate keeps a repeated Apply(true) at zero new calls.
func TestA11y_FixedArgvOnlyLiterals(t *testing.T) {
	runner := &fakeRunner{value: "false\n"}
	r := New(runner.run, (&fakeStatus{}).set)

	r.Apply(true)
	waitConverged(t, r, true)

	literals := argvLiterals()
	recorded := runner.argvSnapshot()
	if len(recorded) == 0 {
		t.Fatal("no argv recorded — the series never ran")
	}
	for i, argv := range recorded {
		for _, element := range argv {
			if !literals[element] {
				t.Errorf("argv[%d] carries a non-literal element %q: %v", i, element, argv)
			}
		}
	}

	before := runner.total()
	r.Apply(true) // lastApplied is already true — the diff gate must record nothing
	time.Sleep(testSettleGap)
	if after := runner.total(); after != before {
		t.Errorf("repeated Apply(true) recorded %d → %d argvs, want unchanged (diff gate)", before, after)
	}
}

// TestA11y_ApplyDoesNotBlockAndConverges pins the fire-and-forget
// contract (WR-01): Apply returns while the series' probe is still
// wedged; after the release the series converges, and the sequence
// Apply(true)→Apply(false)→Apply(true) converges to the LAST desired
// state under -race.
func TestA11y_ApplyDoesNotBlockAndConverges(t *testing.T) {
	release := make(chan struct{})
	runner := &fakeRunner{value: "false\n", blockGet: release}
	r := New(runner.run, (&fakeStatus{}).set)

	applied := make(chan struct{})
	go func() {
		r.Apply(true)
		close(applied)
	}()
	select {
	case <-applied:
	case <-time.After(testDeadline):
		close(release) // let the wedged series finish before the failure reports
		t.Fatal("Apply blocked on the apply series — the fire-and-forget contract is broken")
	}
	close(release)
	waitConverged(t, r, true)

	r.Apply(false)
	waitConverged(t, r, false)
	r.Apply(true)
	waitConverged(t, r, true)
}

// TestA11y_SetCmdNoPipesForm pins the set child's fork shape: the
// no-pipes builder hands back an exec.Cmd with NOTHING attached to the
// standard streams (the wl-copy deadlock precedent) and the fixed
// literal argv.
func TestA11y_SetCmdNoPipesForm(t *testing.T) {
	cmd := newSetCmd(binGSettings, verbSet, schemaGnomeInterface, keyToolkitAccessibility, valTrue)
	if cmd.Stdin != nil || cmd.Stdout != nil || cmd.Stderr != nil {
		t.Error("set command carries pipes — the no-pipes fork shape is broken (the wl-copy precedent)")
	}
	want := []string{binGSettings, verbSet, schemaGnomeInterface, keyToolkitAccessibility, valTrue}
	if !slices.Equal(cmd.Args, want) {
		t.Errorf("set argv = %v, want %v", cmd.Args, want)
	}
}
