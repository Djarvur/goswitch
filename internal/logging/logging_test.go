package logging_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/Djarvur/goswitch/internal/logging"
)

// parseLines decodes every log line as a JSON object.
func parseLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	out := []map[string]any{}
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}

		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("log line is not a JSON object: %q: %v", line, err)
		}
		out = append(out, rec)
	}

	return out
}

// resetLogger restores a harmless default logger once a test replaced it.
func resetLogger(t *testing.T) {
	t.Helper()

	t.Cleanup(func() {
		slog.SetDefault(slog.New(slog.NewTextHandler(&nullWriter{}, nil)))
	})
}

type nullWriter struct{}

func (nullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// TestLogging_JSONShape pins the structured-log contract (INST-04): every
// record is a JSON object with time, level and msg, and all four levels
// are emitted when DEBUG is enabled.
func TestLogging_JSONShape(t *testing.T) {
	resetLogger(t)

	var buf bytes.Buffer
	logging.Setup(&buf, true)

	slog.Error("adapter failure")
	slog.Warn("reconnecting")
	slog.Info("component registered")
	slog.Debug("key")

	records := parseLines(t, &buf)
	if len(records) != 4 {
		t.Fatalf("got %d records, want 4:\n%s", len(records), buf.String())
	}

	wantLevels := []string{"ERROR", "WARN", "INFO", "DEBUG"}
	for i, rec := range records {
		for _, key := range []string{"time", "level", "msg"} {
			if _, ok := rec[key]; !ok {
				t.Errorf("record %d lacks key %q: %v", i, key, rec)
			}
		}
		if got := rec["level"]; got != wantLevels[i] {
			t.Errorf("record %d level = %v, want %v", i, got, wantLevels[i])
		}
	}
}

// TestLogging_LevelFiltering pins that Setup(w, false) drops DEBUG records.
func TestLogging_LevelFiltering(t *testing.T) {
	resetLogger(t)

	var buf bytes.Buffer
	logging.Setup(&buf, false)

	slog.Error("adapter failure")
	slog.Warn("reconnecting")
	slog.Info("component registered")
	slog.Debug("key")

	records := parseLines(t, &buf)
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3 (DEBUG filtered out):\n%s", len(records), buf.String())
	}
	for _, rec := range records {
		if rec["level"] == "DEBUG" {
			t.Errorf("DEBUG record present with debug=false: %v", rec)
		}
	}
}

// TestLogging_KeyTraceGated pins the privacy rule (INST-04 / T-01-01): the
// key trace is emitted only with debug enabled — with debug off, nothing
// is written; with debug on, the record carries the full key payload.
func TestLogging_KeyTraceGated(t *testing.T) {
	resetLogger(t)

	keyTrace := func() {
		slog.Debug("key",
			"keyval", "0x67",
			"keycode", 38,
			"release", false,
			"mods", "0x0")
	}

	var off bytes.Buffer
	logging.Setup(&off, false)
	keyTrace()
	if off.Len() != 0 {
		t.Errorf("key trace written with debug off (privacy violation, T-01-01):\n%s", off.String())
	}

	var on bytes.Buffer
	logging.Setup(&on, true)
	keyTrace()

	records := parseLines(t, &on)
	if len(records) != 1 {
		t.Fatalf("got %d records with debug on, want 1:\n%s", len(records), on.String())
	}
	rec := records[0]
	if rec["msg"] != "key" {
		t.Errorf("msg = %v, want key", rec["msg"])
	}
	for _, key := range []string{"keyval", "keycode", "release", "mods"} {
		if _, ok := rec[key]; !ok {
			t.Errorf("key-trace record lacks %q: %v", key, rec)
		}
	}
}
