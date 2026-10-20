package main

// The marker replay corpus (plan 09-02, the hard gate): the focus marker
// is proven against the REAL journal records of today's live runs — the
// verbatim focus_out/focus_in (and lifecycle DEBUG) lines from the pinned
// journals, replayed through the fixed detector. The zenity-control
// regression (2026-10-14): the pairs landed between the spawn and the
// wait's baseline, so the case hung on its own marker — the replay pins
// that the marker FIRES on exactly those records, in both arrival modes
// (already in the journal at detector entry, and landing mid-wait), and
// the negative: with no fresh pairs above the baseline it must NOT fire.
// The records are focus/lifecycle lines only — never text-bearing records
// (the encoding discipline).

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// arrival modes of the replayed records.
const (
	arrivalPre = "pre" // the records are already in the journal at detector entry (the self-focus beat the wait)
	arrivalMid = "mid" // the records land during the wait (the owner's click)
)

// replayRealRecords is the verbatim real-record set per case, transcribed
// from the pinned journals (gs-t04.log, gs-flip2.log, gs-smoke3.log,
// gs-caps2.log — the 2026-10-12/13 runs); one string per record, joined
// with newlines when replayed.
func replayRealRecords() map[string][]string {
	return map[string][]string{
		"term04-negative": {
			`{"time":"2026-10-12T01:30:16.273169011+03:00","level":"INFO",` +
				`"msg":"focus_out","engine":"goswitch-ru"}`,
			`{"time":"2026-10-12T01:30:16.274553129+03:00","level":"INFO",` +
				`"msg":"focus_in","engine":"goswitch-ru"}`,
			`{"time":"2026-10-12T01:30:16.274623821+03:00","level":"DEBUG",` +
				`"msg":"lifecycle","kind":"focus_in"}`,
			`{"time":"2026-10-12T01:30:16.326275855+03:00","level":"INFO",` +
				`"msg":"focus_out","engine":"goswitch-ru"}`,
			`{"time":"2026-10-12T01:30:16.327412139+03:00","level":"INFO",` +
				`"msg":"focus_in","engine":"goswitch-ru"}`,
			`{"time":"2026-10-12T01:30:16.327456628+03:00","level":"DEBUG",` +
				`"msg":"lifecycle","kind":"focus_in"}`,
			`{"time":"2026-10-12T01:30:17.482580111+03:00","level":"INFO",` +
				`"msg":"focus_out","engine":"goswitch-ru"}`,
			`{"time":"2026-10-12T01:30:17.483323141+03:00","level":"INFO",` +
				`"msg":"focus_in","engine":"goswitch-ru"}`,
		},
		"wezterm-flip": {
			`{"time":"2026-10-11T22:26:19.318727135+03:00","level":"INFO",` +
				`"msg":"focus_out","engine":"goswitch-en"}`,
			`{"time":"2026-10-11T22:26:19.320655732+03:00","level":"INFO",` +
				`"msg":"focus_in","engine":"goswitch-en"}`,
		},
		"spike-smoke": {
			`{"time":"2026-10-11T22:25:24.384618428+03:00","level":"INFO",` +
				`"msg":"focus_in","engine":"goswitch-en"}`,
			`{"time":"2026-10-11T22:25:24.627504483+03:00","level":"INFO",` +
				`"msg":"focus_out","engine":"goswitch-en"}`,
			`{"time":"2026-10-11T22:25:24.629161939+03:00","level":"INFO",` +
				`"msg":"focus_in","engine":"goswitch-en"}`,
		},
		"spike-caps": {
			`{"time":"2026-10-11T22:29:52.424701631+03:00","level":"INFO",` +
				`"msg":"focus_in","engine":"goswitch-en"}`,
			`{"time":"2026-10-11T22:30:00.572138467+03:00","level":"INFO",` +
				`"msg":"focus_in","engine":"goswitch-en"}`,
		},
	}
}

// seedReplayJournal writes the pre-spawn journal state (no pairs yet — the
// daemon's startup records only) and returns the path plus the baseline
// the fixed caller would capture before spawning.
func seedReplayJournal(t *testing.T) (st *stand, focusBase int) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "journal")
	pre := `{"level":"INFO","msg":"config loaded","path":"/tmp/x/config.yaml"}` + "\n" +
		`{"level":"INFO","msg":"ctl service listening","name":"org.djarvur.goswitch"}` + "\n"
	if err := os.WriteFile(path, []byte(pre), probeFilePerm); err != nil {
		t.Fatalf("seed journal: %v", err)
	}
	st = &stand{logPath: path}
	focusBase = st.focusBaseline()

	return st, focusBase
}

// appendReplayRecords lands the case's real records into its journal —
// the spawn/click moment, replayed.
func appendReplayRecords(t *testing.T, s *stand, caseName string) {
	t.Helper()
	f, err := os.OpenFile(s.logPath, os.O_APPEND|os.O_WRONLY, probeFilePerm)
	if err != nil {
		t.Fatalf("open journal: %v", err)
	}
	defer func() { _ = f.Close() }()
	record := strings.Join(replayRealRecords()[caseName], "\n") + "\n"
	if _, err := f.WriteString(record); err != nil {
		t.Fatalf("append records: %v", err)
	}
}

// TestReplayMarkers_OnRealRecords runs the hard gate across every case's
// real records: with the baseline captured pre-spawn, the marker FIRES on
// exactly those records — in both arrival modes per case (the zenity
// self-focus beat the wait; the flip/smoke/caps pairs landed during it).
func TestReplayMarkers_OnRealRecords(t *testing.T) {
	for _, tc := range []struct {
		caseName string
		surface  string
		arrival  string
	}{
		{"term04-negative", "zenity control dialog", arrivalPre},
		{"wezterm-flip", "wezterm-flip раунд 1 (xim)", arrivalMid},
		{"spike-smoke", "spike-smoke fixture", arrivalMid},
		{"spike-caps", "gnome-terminal window", arrivalPre},
	} {
		s, focusBase := seedReplayJournal(t)
		switch tc.arrival {
		case arrivalPre:
			appendReplayRecords(t, s, tc.caseName)
		case arrivalMid:
			go func() {
				time.Sleep(100 * time.Millisecond)
				appendReplayRecords(t, s, tc.caseName)
			}()
		}
		done := make(chan error, 1)
		go func() {
			done <- s.waitFocusInFrom(context.Background(), tc.caseName, tc.surface, focusBase)
		}()
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("%s (%s arrival): the marker did not fire on the real records: %v",
					tc.caseName, tc.arrival, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("%s (%s arrival): the marker never fired on the real records", tc.caseName, tc.arrival)
		}
	}
}

// TestReplayMarker_Negative_NoFreshPairsDoesNotFire pins the negative half
// of the gate: with the baseline EQUAL to the journal's record count — no
// fresh pairs above the spawn baseline — the marker must keep waiting
// (bounded by the aborted context), never fire on stale state.
func TestReplayMarker_Negative_NoFreshPairsDoesNotFire(t *testing.T) {
	s := &stand{logPath: seedFocusJournal(t, 4)}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	if err := s.waitFocusInFrom(ctx, "case", "surface", 4); err == nil {
		t.Fatal("the marker fired on stale pairs below the baseline — the gate is void")
	}
	if elapsed := time.Since(start); elapsed < 150*time.Millisecond {
		t.Fatalf("the marker fired after %v on stale state", elapsed)
	}
}
