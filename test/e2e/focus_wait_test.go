package main

// The focus-detector corpus (plan 09-02): the detector keys on daemon
// focus_in pairs ABOVE A PRE-SPAWN BASELINE — the baseline must be
// captured before the fixture spawns, because a terminal-child window
// usually self-focuses within milliseconds and an owner click lands
// seconds later; a baseline taken after the wait starts sits above the
// very pairs it waits for and the case hangs on its own marker (the
// 2026-10-14 live finding on the zenity control).

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// seedFocusJournal writes a journal with n focus_in records.
func seedFocusJournal(t *testing.T, n int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "journal")
	b := make([]byte, 0, n*64)
	for range n {
		b = append(b, []byte(`{"level":"INFO","msg":"focus_in","engine":"goswitch-ru"}`+"\n")...)
	}
	if err := os.WriteFile(path, b, probeFilePerm); err != nil {
		t.Fatalf("seed journal: %v", err)
	}

	return path
}

// TestWaitFocusInFrom_EarlyPairsFire pins the zenity-control fix: pairs
// that landed between the spawn and the wait (baseline 0, pairs already in
// the journal) fire the detector IMMEDIATELY — the self-focus is the
// engagement, never invisible.
func TestWaitFocusInFrom_EarlyPairsFire(t *testing.T) {
	s := &stand{logPath: seedFocusJournal(t, 3)}

	start := time.Now()
	if err := s.waitFocusInFrom(context.Background(), "case", "zenity control dialog", 0); err != nil {
		t.Fatalf("waitFocusInFrom = %v with pairs above the baseline, want nil", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("waitFocusInFrom took %v with the pairs already in the journal", elapsed)
	}
}

// TestWaitFocusInFrom_NewPairsFire pins the live-arrival half: a pair that
// lands during the wait (the owner's click, the map focus) resolves it.
func TestWaitFocusInFrom_NewPairsFire(t *testing.T) {
	path := seedFocusJournal(t, 0)
	s := &stand{logPath: path}
	go func() {
		time.Sleep(100 * time.Millisecond)
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, probeFilePerm)
		if err == nil {
			_, _ = f.WriteString(`{"level":"INFO","msg":"focus_in","engine":"goswitch-ru"}` + "\n")
			_ = f.Close()
		}
	}()

	if err := s.waitFocusInFrom(context.Background(), "case", "surface", 0); err != nil {
		t.Fatalf("waitFocusInFrom = %v with a pair landing mid-wait, want nil", err)
	}
}

// TestWaitFocusInFrom_StalePairsBelowBaselineWait pins the discipline the
// fix must not lose: pairs BELOW the caller's baseline are old state — the
// wait keeps waiting (bounded here by the aborted context, not the
// 30-minute budget).
func TestWaitFocusInFrom_StalePairsBelowBaselineWait(t *testing.T) {
	s := &stand{logPath: seedFocusJournal(t, 5)}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	if err := s.waitFocusInFrom(ctx, "case", "surface", 5); err == nil {
		t.Fatal("waitFocusInFrom = nil with no pairs above the baseline, want the wait to continue")
	}
	if elapsed := time.Since(start); elapsed < 150*time.Millisecond {
		t.Fatalf("waitFocusInFrom returned after %v — the stale pairs leaked through", elapsed)
	}
}

// TestFocusBaselineSnapshotsCount pins the baseline primitive: it is the
// current focus_in record count of the pinned journal.
func TestFocusBaselineSnapshotsCount(t *testing.T) {
	s := &stand{logPath: seedFocusJournal(t, 7)}
	if got := s.focusBaseline(); got != 7 {
		t.Fatalf("focusBaseline = %d, want 7", got)
	}
}
