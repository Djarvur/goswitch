package main

// The registration-wait corpus (plan 09-02): the preflight's
// engine-registered check is a retry-until-ready wait, not a fixed cliff —
// and when the daemon process DIES before registering (the name-conflict
// death: "another goswitchd instance already registered"), the wait must
// fail fast with the journal's actual error instead of burning the whole
// budget on a generic timeout (the 2026-10-13 finding).

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLastJournalError pins the diagnostic extraction: the LAST ERROR-level
// journal record is what the fail-fast wait reports (the R1-era generic
// timeout cost a whole batch of triage), and a clean journal yields "".
func TestLastJournalError(t *testing.T) {
	journal := strings.Join([]string{
		`{"time":"t1","level":"INFO","msg":"config loaded"}`,
		`{"time":"t2","level":"ERROR","msg":"another goswitchd instance already registered","reply":2}`,
		`{"time":"t3","level":"INFO","msg":"lifecycle","kind":"focus_in"}`,
	}, "\n")
	got := lastErrorLine(journal)
	if !strings.Contains(got, "another goswitchd instance already registered") {
		t.Fatalf("lastErrorLine = %q, want the last ERROR record", got)
	}
	if got := lastErrorLine(`{"level":"INFO","msg":"clean"}`); got != "" {
		t.Fatalf("lastErrorLine on a clean journal = %q, want empty", got)
	}
}

// startTestDaemon spawns a real quick process as the stand's daemon so the
// liveness probe sees a genuinely exited child (headless: /bin/true).
func startTestDaemon(t *testing.T, s *stand) {
	t.Helper()
	//nolint:noctx // the test's daemon probe: the process must outlive Start, Wait is reaped
	cmd := exec.Command("true")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start test daemon: %v", err)
	}
	s.daemon = cmd
	// Reap it so the process does not linger as a zombie for the test run.
	go func() { _ = cmd.Wait() }()
}

// TestWaitForLogWhileAlive_DeadDaemonFailsFast pins the fail-fast half: a
// daemon that exits before the mark lands produces the dead-daemon error
// with the journal's error record inside — promptly, not after the budget.
func TestWaitForLogWhileAlive_DeadDaemonFailsFast(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal")
	journal := `{"level":"INFO","msg":"config loaded"}` + "\n" +
		`{"level":"ERROR","msg":"another goswitchd instance already registered","reply":2}` + "\n"
	if err := os.WriteFile(path, []byte(journal), probeFilePerm); err != nil {
		t.Fatalf("seed journal: %v", err)
	}
	s := &stand{logPath: path}
	startTestDaemon(t, s)

	start := time.Now()
	err := s.waitForLogWhileAlive(context.Background(), componentRegisteredMark, 30*time.Second)
	if err == nil {
		t.Fatal("waitForLogWhileAlive = nil for a dead daemon without the mark, want error")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("waitForLogWhileAlive took %v — the fail-fast is not fast", elapsed)
	}
	const journalErr = "another goswitchd instance already registered"
	if !strings.Contains(err.Error(), "exited") || !strings.Contains(err.Error(), journalErr) {
		t.Fatalf("waitForLogWhileAlive error %q does not name the death and the journal error", err.Error())
	}
}

// TestWaitForLogWhileAlive_MarkLands pins the success half: the mark in the
// journal resolves the wait even while the daemon lives.
func TestWaitForLogWhileAlive_MarkLands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal")
	if err := os.WriteFile(path, []byte(componentRegisteredMark+"\n"), probeFilePerm); err != nil {
		t.Fatalf("seed journal: %v", err)
	}
	s := &stand{logPath: path}
	startTestDaemon(t, s)

	if err := s.waitForLogWhileAlive(context.Background(), componentRegisteredMark, 2*time.Second); err != nil {
		t.Fatalf("waitForLogWhileAlive = %v, want nil once the mark lands", err)
	}
}
