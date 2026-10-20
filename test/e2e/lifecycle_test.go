package main

// The stale-holder stop corpus (plan 09-02): a SIGTERM'd stand daemon can
// stall in its own shutdown path and keep the component name held — the
// stop must escalate to SIGKILL after a short grace, because the socket
// close is what releases the name (the 2026-10-13 live finding: the wait
// burned its whole budget behind a TERM'd holder).

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

// spawnStaleHolder starts a process whose command line matches the stand
// daemon pattern and which IGNORES SIGTERM — the stuck-in-shutdown holder.
// The bash exec -a form sets argv[0] so pgrep -f matches the pattern; the
// trap swallows TERM so only SIGKILL stops it. The returned channel closes
// when the process is reaped (a killed-but-unreaped child still shows in
// the process table).
func spawnStaleHolder(t *testing.T) (<-chan struct{}, *exec.Cmd) {
	t.Helper()
	//nolint:noctx // the corpus's own holder: stopped (SIGKILLed) within the test
	cmd := exec.Command("bash", "-c", `trap "" TERM; exec -a "goswitchd -debug" sleep 30`)
	if err := cmd.Start(); err != nil {
		t.Fatalf("spawn stale holder: %v", err)
	}
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()

	return done, cmd
}

// TestStopStaleStandDaemon_TermThenKill pins the escalation: a TERM-proof
// holder is gone after stopStaleStandDaemonWith — SIGKILLed inside the
// grace plus a small poll margin, never left holding the name.
func TestStopStaleStandDaemon_TermThenKill(t *testing.T) {
	done, holder := spawnStaleHolder(t)
	defer func() { _ = holder.Process.Kill() }()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := stopStaleStandDaemonWith(ctx, 500*time.Millisecond); err != nil {
		t.Fatalf("stopStaleStandDaemonWith: %v", err)
	}

	select {
	case <-done:
		return
	case <-time.After(3 * time.Second):
		t.Fatal("the TERM-proof holder survived stopStaleStandDaemonWith — no escalation")
	}
}

// TestStopStaleStandDaemon_NoHolderIsQuiet pins the idempotent half: with
// no matching process the stop returns nil without delay.
func TestStopStaleStandDaemon_NoHolderIsQuiet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	if err := stopStaleStandDaemonWith(ctx, 500*time.Millisecond); err != nil {
		t.Fatalf("stopStaleStandDaemonWith = %v with no holder, want nil", err)
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("stopStaleStandDaemonWith took %v with no holder", elapsed)
	}
}
