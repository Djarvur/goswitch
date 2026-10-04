package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Djarvur/goswitch/internal/config"
)

// writerDocYAML is a live hand-tuned user document: a header comment and
// inline comments on the flip target AND on a neighbouring key — exactly
// the byte shapes the Node round-trip must survive (research Q3a).
const writerDocYAML = `# goswitch config — hand-tuned
hotkeys:
  tap_key: shift_r # single tap flips the layout
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
autocorrect:
  enabled: false # the toggle target — comments must survive
  apps_blocklist: []
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
`

// writerBrokenYAML is not YAML — the broken-source corpus demands an error
// and a byte-untouched file.
const writerBrokenYAML = "hotkeys: [unclosed"

// writeWriterDoc writes the corpus to config.yaml in a temp dir and returns
// its path — the writer under test owns the file from then on.
func writeWriterDoc(t *testing.T, corpus string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(corpus), 0o600); err != nil {
		t.Fatalf("write corpus: %v", err)
	}

	return path
}

// TestWriter_FlipRoundTrip pins the tracer's one path — disk → Node → disk
// → Load: the flip lands in the file and every value case reads BACK
// through the strict Load (Pitfall 6 — a test without the reverse read
// stays green on a writer that changes nothing).
func TestWriter_FlipRoundTrip(t *testing.T) {
	t.Run("off to on", func(t *testing.T) {
		path := writeWriterDoc(t, writerDocYAML)
		if err := config.SetAutocorrectEnabled(path, true); err != nil {
			t.Fatalf("SetAutocorrectEnabled(true): %v", err)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatalf("reload flipped config: %v", err)
		}
		if !cfg.Autocorrect.Enabled {
			t.Error("reloaded autocorrect.enabled = false, want true — the flip never reached the disk")
		}
	})

	t.Run("on to off", func(t *testing.T) {
		path := writeWriterDoc(t, writerDocYAML)
		if err := config.SetAutocorrectEnabled(path, true); err != nil {
			t.Fatalf("SetAutocorrectEnabled(true): %v", err)
		}
		if err := config.SetAutocorrectEnabled(path, false); err != nil {
			t.Fatalf("SetAutocorrectEnabled(false): %v", err)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatalf("reload flipped config: %v", err)
		}
		if cfg.Autocorrect.Enabled {
			t.Error("reloaded autocorrect.enabled = true, want false — the flip back never reached the disk")
		}
	})
}

// TestWriter_CommentsSurvive pins the round-trip's user contract: the
// header comment, the flipped key's inline comment and a neighbouring
// section's inline comment are all still in the file after the write.
func TestWriter_CommentsSurvive(t *testing.T) {
	path := writeWriterDoc(t, writerDocYAML)
	if err := config.SetAutocorrectEnabled(path, true); err != nil {
		t.Fatalf("SetAutocorrectEnabled(true): %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	out := string(raw)
	for _, marker := range []string{
		"# goswitch config — hand-tuned",
		"# the toggle target — comments must survive",
		"# single tap flips the layout",
	} {
		if !strings.Contains(out, marker) {
			t.Errorf("result lost the comment %q — the writer regenerated the document", marker)
		}
	}
}

// TestWriter_IdempotentFlip pins the repeated same-value toggle: no error,
// and the file content is unchanged by the second write.
func TestWriter_IdempotentFlip(t *testing.T) {
	path := writeWriterDoc(t, writerDocYAML)
	if err := config.SetAutocorrectEnabled(path, true); err != nil {
		t.Fatalf("SetAutocorrectEnabled(true): %v", err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read first result: %v", err)
	}
	if err := config.SetAutocorrectEnabled(path, true); err != nil {
		t.Fatalf("idempotent SetAutocorrectEnabled(true): %v", err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read second result: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Error("the second same-value flip changed the bytes — the writer is not idempotent")
	}
}

// TestWriter_AtomicWrite pins the write discipline: no .tmp artifacts stay
// in the directory after the rename, and the file mode is owner-only 0600
// (the configFilePerm precedent of test/e2e/case_select.go, T-07-03-02).
func TestWriter_AtomicWrite(t *testing.T) {
	path := writeWriterDoc(t, writerDocYAML)
	if err := config.SetAutocorrectEnabled(path, true); err != nil {
		t.Fatalf("SetAutocorrectEnabled(true): %v", err)
	}
	dir := filepath.Dir(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read directory: %v", err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp") {
			t.Errorf("temp artifact %s left in the config directory — the rename never completed", e.Name())
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat result: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("result mode = %v, want -rw------- (0600)", got)
	}
}

// TestWriter_BrokenSourceUntouched pins the failure atomicity: a
// non-parseable source document yields an error and leaves the original
// file byte-equal — a broken user document is never silently rewritten.
func TestWriter_BrokenSourceUntouched(t *testing.T) {
	path := writeWriterDoc(t, writerBrokenYAML)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read broken source: %v", err)
	}
	if err := config.SetAutocorrectEnabled(path, true); err == nil {
		t.Fatal("SetAutocorrectEnabled on a broken document returned nil — the broken YAML went unnoticed")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read untouched source: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Error("the broken source file changed — a failed write must leave the original byte-equal")
	}
}
