package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
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

// writerDocNoAutocorrectYAML is a complete pre-autocorrect document with a
// marker comment inside the hotkeys section — the insertion corpus proves
// the rest of the document (marker included) survives a section insert.
const writerDocNoAutocorrectYAML = `# goswitch config — hand-tuned
hotkeys:
  tap_key: shift_r # marker — another section's comment must survive
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
`

// writerSoundDocYAML carries a sound section with an inline marker on its
// enabled key — the sound-flip corpus reads it back after the write.
const writerSoundDocYAML = `# goswitch config — hand-tuned
hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
autocorrect:
  enabled: false # the autocorrect marker
  apps_blocklist: []
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
sound:
  enabled: true # the sound marker — comments must survive
  autocorrect_event: message
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

// assertCreatedDefaults fails unless the created document carries the
// Defaults()-equivalent values — the shared tail of the creation corpus.
func assertCreatedDefaults(t *testing.T, cfg *config.Config) {
	t.Helper()

	if cfg.Autocorrect.MinWordLen != 4 || cfg.Autocorrect.TrigramMargin != 2.0 || cfg.Autocorrect.TrigramFloor != 1.0 {
		t.Errorf(
			"created thresholds = %d/%v/%v, want the Defaults 4/2.0/1.0",
			cfg.Autocorrect.MinWordLen, cfg.Autocorrect.TrigramMargin, cfg.Autocorrect.TrigramFloor,
		)
	}
	if len(cfg.Autocorrect.AppsBlocklist) != 0 {
		t.Errorf("created blocklist = %v, want empty", cfg.Autocorrect.AppsBlocklist)
	}
	if cfg.Hotkeys.TapKey != "shift_r" || cfg.Hotkeys.WordLayoutCombo != "shift+ctrl_r" {
		t.Errorf("created hotkeys = %v, want the Defaults bindings", cfg.Hotkeys)
	}
	if cfg.Hotkeys.ModeSwitchChord != "super+space" {
		t.Errorf("created mode_switch_chord = %q, want super+space", cfg.Hotkeys.ModeSwitchChord)
	}
	if cfg.Timeouts.TapWindowMs != 300 || cfg.Timeouts.VerifyWaitMs != 100 {
		t.Errorf("created timeouts = %v, want 300/100", cfg.Timeouts)
	}
	if cfg.Correction.BackspaceCap != 50 || cfg.Correction.ClipboardRung || !cfg.Correction.FlipAfterCorrection {
		t.Errorf("created correction = %v, want the Defaults 50/false/true", cfg.Correction)
	}
}

// TestWriter_CreatesFullDocumentOnAbsent pins the ensure semantics of the
// first toggle (owner decision, adopt+watch): on a machine without a
// config file the autocorrect toggle CREATES the complete
// Defaults()-equivalent document with the toggle value in place — and the
// created document passes the strict Load with every section complete
// (the completeness rule of 03-02: Load overlays nothing).
func TestWriter_CreatesFullDocumentOnAbsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml") // absent — the test never creates it
	if err := config.SetAutocorrectEnabled(path, true); err != nil {
		t.Fatalf("SetAutocorrectEnabled(true) on an absent file: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load created document: %v", err)
	}
	if !cfg.Autocorrect.Enabled {
		t.Error("created autocorrect.enabled = false, want the toggle value true")
	}
	assertCreatedDefaults(t, cfg)
}

// TestWriter_EnsureDocumentCreatesAndNoOps pins EnsureDocument: creation
// of the same full document when the file is absent (enabled:false), and a
// byte-level no-op once it exists — «Настройки…» never rewrites a live
// document.
func TestWriter_EnsureDocumentCreatesAndNoOps(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml") // absent
	if err := config.EnsureDocument(path); err != nil {
		t.Fatalf("EnsureDocument on an absent file: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load ensured document: %v", err)
	}
	if cfg.Autocorrect.Enabled {
		t.Error("ensured autocorrect.enabled = true, want false (the default-off document)")
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ensured document: %v", err)
	}
	if err := config.EnsureDocument(path); err != nil {
		t.Fatalf("EnsureDocument on an existing file: %v", err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after re-ensure: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Error("EnsureDocument rewrote an existing document — the contract is a byte-level no-op")
	}
}

// TestWriter_InsertsMissingSection pins the insertion branch: a valid
// document WITHOUT the autocorrect section gains a WHOLE five-key mapping
// (enabled + thresholds + empty blocklist) and the rest of the document —
// another section's marker comment included — survives.
func TestWriter_InsertsMissingSection(t *testing.T) {
	path := writeWriterDoc(t, writerDocNoAutocorrectYAML)
	if err := config.SetAutocorrectEnabled(path, true); err != nil {
		t.Fatalf("SetAutocorrectEnabled(true) without the section: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load after insert: %v", err)
	}
	if !cfg.Autocorrect.Enabled {
		t.Error("inserted autocorrect.enabled = false, want true")
	}
	if cfg.Autocorrect.MinWordLen != 4 || cfg.Autocorrect.TrigramMargin != 2.0 || cfg.Autocorrect.TrigramFloor != 1.0 {
		t.Errorf(
			"inserted thresholds = %d/%v/%v, want the Defaults 4/2.0/1.0 — the section must be complete",
			cfg.Autocorrect.MinWordLen, cfg.Autocorrect.TrigramMargin, cfg.Autocorrect.TrigramFloor,
		)
	}
	if len(cfg.Autocorrect.AppsBlocklist) != 0 {
		t.Errorf("inserted blocklist = %v, want the empty list", cfg.Autocorrect.AppsBlocklist)
	}
	if cfg.Timeouts.TapWindowMs != 300 || cfg.Hotkeys.TapKey != "shift_r" {
		t.Errorf("the rest of the document changed: hotkeys/timeouts = %v/%v", cfg.Hotkeys, cfg.Timeouts)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after insert: %v", err)
	}
	if !strings.Contains(string(raw), "# marker — another section's comment must survive") {
		t.Error("the hotkeys marker comment was lost — the insert regenerated the document")
	}
}

// TestWriter_SoundFlipReverseRead pins the «Звук» toggle's persist: the
// same Node round-trip flips sound.enabled, the reverse read goes through
// Load's EffectiveEnabled, and the marker comments survive both flips.
func TestWriter_SoundFlipReverseRead(t *testing.T) {
	path := writeWriterDoc(t, writerSoundDocYAML)
	if err := config.SetSoundEnabled(path, false); err != nil {
		t.Fatalf("SetSoundEnabled(false): %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load after sound off: %v", err)
	}
	if cfg.Sound.EffectiveEnabled() {
		t.Error("sound EffectiveEnabled = true, want false after the flip")
	}
	if err := config.SetSoundEnabled(path, true); err != nil {
		t.Fatalf("SetSoundEnabled(true): %v", err)
	}
	cfg, err = config.Load(path)
	if err != nil {
		t.Fatalf("load after sound on: %v", err)
	}
	if !cfg.Sound.EffectiveEnabled() || cfg.Sound.Enabled == nil || !*cfg.Sound.Enabled {
		t.Errorf("sound after flip back = %v, want encoded true", cfg.Sound.Enabled)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after sound flips: %v", err)
	}
	out := string(raw)
	for _, marker := range []string{
		"# the sound marker — comments must survive",
		"# the autocorrect marker",
	} {
		if !strings.Contains(out, marker) {
			t.Errorf("result lost the comment %q", marker)
		}
	}
}

// TestWriter_SoundInsertMissingSection pins the sound insertion branch: a
// document without the sound section gains a WHOLE two-key mapping
// (enabled at the toggle value + autocorrect_event from the Defaults) and
// the rest of the document survives.
func TestWriter_SoundInsertMissingSection(t *testing.T) {
	path := writeWriterDoc(t, writerDocYAML) // no sound section
	if err := config.SetSoundEnabled(path, false); err != nil {
		t.Fatalf("SetSoundEnabled(false) without the section: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load after sound insert: %v", err)
	}
	if cfg.Sound.EffectiveEnabled() {
		t.Error("inserted sound.enabled reads ON, want the toggle value false")
	}
	if cfg.Sound.AutocorrectEvent != config.DefaultSoundAutocorrectEvent {
		t.Errorf(
			"inserted autocorrect_event = %q, want the Defaults %q",
			cfg.Sound.AutocorrectEvent, config.DefaultSoundAutocorrectEvent,
		)
	}
	if cfg.Autocorrect.Enabled {
		t.Error("the autocorrect section changed during the sound insert")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after sound insert: %v", err)
	}
	if !strings.Contains(string(raw), "# the toggle target — comments must survive") {
		t.Error("the autocorrect marker comment was lost — the insert regenerated the document")
	}
}

// TestWriter_FullDocumentStrictLoad pins the created document's validity:
// the ensure branch's document decodes under the strict KnownFields decode
// without unknown-field refusals, carries the owner's default-ON sound
// switch encoded, and honors the toggle argument for autocorrect.
func TestWriter_FullDocumentStrictLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml") // absent
	if err := config.SetAutocorrectEnabled(path, true); err != nil {
		t.Fatalf("SetAutocorrectEnabled(true) on an absent file: %v", err)
	}
	cfg, err := config.Load(path) // strict KnownFields decode — the assertion is the success
	if err != nil {
		t.Fatalf("strict load of the created document: %v", err)
	}
	if cfg.Sound.Enabled == nil || !*cfg.Sound.Enabled {
		t.Errorf("created sound.enabled = %v, want encoded true (the owner's default ON)", cfg.Sound.Enabled)
	}
	if !cfg.Autocorrect.Enabled {
		t.Error("created autocorrect.enabled = false, want the toggle argument true")
	}
}

// writerDocBlocklistYAML is the append corpus: the autocorrect section with
// ONE listed pattern (a block sequence, not the empty flow `[]`) and inline
// comments on the toggle AND on the listed pattern — the golden-diff corpus
// proves the write adds exactly one list line and regenerates nothing.
const writerDocBlocklistYAML = `# goswitch config — hand-tuned
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
  apps_blocklist:
    - org.gnome.Gedit # the listed pattern — the append joins below it
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
`

// writerDocNoBlocklistKeyYAML carries a complete Load-valid document whose
// autocorrect section lacks the apps_blocklist key — the absent-key corpus:
// the append creates the key (an autocorrect-only document would be refused
// by the strict pre-write validation, D-33 — the daemon never accepted
// partial documents).
const writerDocNoBlocklistKeyYAML = `# goswitch config — hand-tuned
hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
autocorrect:
  enabled: false # the section exists, the blocklist key does not
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
`

// writerDocNullBlocklistYAML is a complete Load-valid document with an
// explicit null blocklist key — the null corpus: the append replaces the
// null with a real sequence.
const writerDocNullBlocklistYAML = `# goswitch config — hand-tuned
hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: true
autocorrect:
  enabled: false
  apps_blocklist:
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
`

// writerDocScalarDocumentYAML is a valid YAML document whose root is a
// scalar — there is no mapping to find a section in.
const writerDocScalarDocumentYAML = "just a scalar document\n"

// writerDocSeqDocumentYAML is a valid YAML document whose root is a
// sequence — also not a mapping.
const writerDocSeqDocumentYAML = "- one\n- two\n"

// writerDocUnknownKeyYAML parses cleanly as a Node tree, but its EDITED
// result would fail the strict KnownFields decode — the constructed-broken-
// result corpus: the append must refuse before any bytes land (D-33 — the
// writer never persists a document the strict decoder rejects).
const writerDocUnknownKeyYAML = `# goswitch config — hand-tuned
bogus_section:
  future_key: 1
autocorrect:
  enabled: false
  apps_blocklist: []
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
`

// goldenAddedLines walks the after-lines against the before-lines assuming
// an ordered insert (the writer's minimal-write contract) and returns the
// inserted lines — the golden-diff helper of the append corpus. A removed,
// reordered or edited existing line fails the test.
func goldenAddedLines(t *testing.T, before, after []byte) []string {
	t.Helper()

	beforeLines := strings.Split(string(before), "\n")
	afterLines := strings.Split(string(after), "\n")
	var added []string
	consumed := 0
	for _, line := range afterLines {
		if consumed < len(beforeLines) && line == beforeLines[consumed] {
			consumed++

			continue
		}
		added = append(added, line)
	}
	if consumed != len(beforeLines) {
		t.Fatalf("golden diff broken: the write moved %d existing lines", len(beforeLines)-consumed)
	}

	return added
}

// assertAppendCommentsSurvive fails unless the append result still carries
// the corpus comment markers — the shared tail of the comment corpus.
func assertAppendCommentsSurvive(t *testing.T, raw []byte) {
	t.Helper()

	out := string(raw)
	for _, marker := range []string{
		"# goswitch config — hand-tuned",
		"# single tap flips the layout",
		"# the listed pattern — the append joins below it",
	} {
		if !strings.Contains(out, marker) {
			t.Errorf("result lost the comment %q — the writer regenerated the document", marker)
		}
	}
}

// The append corpus's own names (the goconst discipline — the shared
// appGedit and acBlockSubstr live in config_test.go of this test package).
const (
	appWezterm   = "wezterm"
	blankPattern = "   "
)

// TestWriter_AppendRoundTrip pins the append's one path — disk → Node →
// disk → Load: the pattern lands in the file and the whole document reads
// BACK through the strict Load (the FlipRoundTrip form; a test without the
// reverse read stays green on a writer that changes nothing).
func TestWriter_AppendRoundTrip(t *testing.T) {
	t.Run("joins the listed pattern", func(t *testing.T) {
		path := writeWriterDoc(t, writerDocBlocklistYAML)
		if err := config.AppendAutocorrectBlocklist(path, appWezterm); err != nil {
			t.Fatalf("AppendAutocorrectBlocklist: %v", err)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatalf("reload appended config: %v", err)
		}
		if !slices.Equal(cfg.Autocorrect.AppsBlocklist, []string{appGedit, appWezterm}) {
			t.Errorf("reloaded blocklist = %v, want original entry + append", cfg.Autocorrect.AppsBlocklist)
		}
		if cfg.Autocorrect.Enabled {
			t.Error("the append switched autocorrect.enabled on — blocking an app never enables the layer (D-54)")
		}
	})

	t.Run("fills the empty list", func(t *testing.T) {
		path := writeWriterDoc(t, writerDocYAML)
		if err := config.AppendAutocorrectBlocklist(path, acBlockSubstr); err != nil {
			t.Fatalf("AppendAutocorrectBlocklist: %v", err)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatalf("reload appended config: %v", err)
		}
		if !slices.Equal(cfg.Autocorrect.AppsBlocklist, []string{acBlockSubstr}) {
			t.Errorf("reloaded blocklist = %v, want [chrom]", cfg.Autocorrect.AppsBlocklist)
		}
	})
}

// TestWriter_AppendGoldenDiff pins the minimal-write contract (roadmap
// criterion 2): the ONLY change the append makes to the document is the one
// new sequence element — every existing line survives in place, comments
// included.
func TestWriter_AppendGoldenDiff(t *testing.T) {
	path := writeWriterDoc(t, writerDocBlocklistYAML)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read before: %v", err)
	}
	if err := config.AppendAutocorrectBlocklist(path, appWezterm); err != nil {
		t.Fatalf("AppendAutocorrectBlocklist: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after: %v", err)
	}
	added := goldenAddedLines(t, before, after)
	if len(added) != 1 {
		t.Fatalf("the append changed %d lines (%v), want exactly one new list line", len(added), added)
	}
	if strings.TrimSpace(added[0]) != "- "+appWezterm {
		t.Errorf("the new line = %q, want the plain block-sequence item - %s", added[0], appWezterm)
	}
	assertAppendCommentsSurvive(t, after)
}

// TestWriter_AppendAbsentFile pins the ensure branch (A4): on a machine
// without a config file the append CREATES the complete Defaults()-equivalent
// document with the pattern in place, the layer still at its default OFF —
// and the created document passes the strict Load.
func TestWriter_AppendAbsentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml") // absent — the test never creates it
	if err := config.AppendAutocorrectBlocklist(path, acBlockSubstr); err != nil {
		t.Fatalf("AppendAutocorrectBlocklist on an absent file: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load created document: %v", err)
	}
	if !slices.Equal(cfg.Autocorrect.AppsBlocklist, []string{acBlockSubstr}) {
		t.Errorf("created blocklist = %v, want [chrom]", cfg.Autocorrect.AppsBlocklist)
	}
	if cfg.Autocorrect.Enabled {
		t.Error("created autocorrect.enabled = true, want the D-54 default false — the append never enables the layer")
	}
	if cfg.Autocorrect.MinWordLen != 4 || cfg.Autocorrect.TrigramMargin != 2.0 || cfg.Autocorrect.TrigramFloor != 1.0 {
		t.Errorf(
			"created thresholds = %d/%v/%v, want the Defaults 4/2.0/1.0",
			cfg.Autocorrect.MinWordLen, cfg.Autocorrect.TrigramMargin, cfg.Autocorrect.TrigramFloor,
		)
	}
	if cfg.Hotkeys.TapKey != defTapKey || cfg.Timeouts.TapWindowMs != 300 {
		t.Errorf("the rest of the created document changed: hotkeys/timeouts = %v/%v", cfg.Hotkeys, cfg.Timeouts)
	}
}

// TestWriter_AppendAbsentKeyAndNull pins the two partial-shape branches: a
// section WITHOUT the apps_blocklist key gains the key with the pattern, and
// an explicit null key becomes a real sequence — both strict-Load green.
func TestWriter_AppendAbsentKeyAndNull(t *testing.T) {
	t.Run("absent key", func(t *testing.T) {
		path := writeWriterDoc(t, writerDocNoBlocklistKeyYAML)
		if err := config.AppendAutocorrectBlocklist(path, acBlockSubstr); err != nil {
			t.Fatalf("AppendAutocorrectBlocklist without the key: %v", err)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatalf("reload after key insert: %v", err)
		}
		if !slices.Equal(cfg.Autocorrect.AppsBlocklist, []string{acBlockSubstr}) {
			t.Errorf("blocklist after key insert = %v, want [chrom]", cfg.Autocorrect.AppsBlocklist)
		}
	})

	t.Run("explicit null", func(t *testing.T) {
		path := writeWriterDoc(t, writerDocNullBlocklistYAML)
		if err := config.AppendAutocorrectBlocklist(path, acBlockSubstr); err != nil {
			t.Fatalf("AppendAutocorrectBlocklist on a null key: %v", err)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatalf("reload after null replace: %v", err)
		}
		if !slices.Equal(cfg.Autocorrect.AppsBlocklist, []string{acBlockSubstr}) {
			t.Errorf("blocklist after null replace = %v, want [chrom]", cfg.Autocorrect.AppsBlocklist)
		}
	})
}

// TestWriter_AppendInsertsMissingSection pins the insertion branch: a valid
// document WITHOUT the autocorrect section gains a WHOLE five-key mapping
// (enabled at the D-54 default + thresholds from the Defaults) with the
// pattern in place, and the rest of the document — another section's marker
// comment included — survives.
func TestWriter_AppendInsertsMissingSection(t *testing.T) {
	path := writeWriterDoc(t, writerDocNoAutocorrectYAML)
	if err := config.AppendAutocorrectBlocklist(path, acBlockSubstr); err != nil {
		t.Fatalf("AppendAutocorrectBlocklist without the section: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load after insert: %v", err)
	}
	if !slices.Equal(cfg.Autocorrect.AppsBlocklist, []string{acBlockSubstr}) {
		t.Errorf("inserted blocklist = %v, want [chrom]", cfg.Autocorrect.AppsBlocklist)
	}
	if cfg.Autocorrect.Enabled {
		t.Error("inserted autocorrect.enabled = true, want the D-54 default false — the append never enables the layer")
	}
	if cfg.Autocorrect.MinWordLen != 4 || cfg.Autocorrect.TrigramMargin != 2.0 || cfg.Autocorrect.TrigramFloor != 1.0 {
		t.Errorf(
			"inserted thresholds = %d/%v/%v, want the Defaults 4/2.0/1.0 — the section must be complete",
			cfg.Autocorrect.MinWordLen, cfg.Autocorrect.TrigramMargin, cfg.Autocorrect.TrigramFloor,
		)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after insert: %v", err)
	}
	if !strings.Contains(string(raw), "# marker — another section's comment must survive") {
		t.Error("the hotkeys marker comment was lost — the insert regenerated the document")
	}
}

// TestWriter_AppendDedupNoOp pins the dedup guard (D-01's mechanical
// protection): appending a pattern already in the list is an idempotent
// success and the file bytes are unchanged — no write happens at all.
func TestWriter_AppendDedupNoOp(t *testing.T) {
	path := writeWriterDoc(t, writerDocBlocklistYAML)
	if err := config.AppendAutocorrectBlocklist(path, appGedit); err != nil {
		t.Fatalf("AppendAutocorrectBlocklist (already listed): %v", err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read first result: %v", err)
	}
	if err := config.AppendAutocorrectBlocklist(path, appGedit); err != nil {
		t.Fatalf("idempotent AppendAutocorrectBlocklist: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read second result: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Error("the repeated append changed the bytes — the dedup must be a byte-level no-op")
	}
}

// TestWriter_AppendAtomicWrite pins the append's write discipline: no .tmp
// artifacts stay in the directory after the rename, and the file mode is
// owner-only 0600 (the TestWriter_AtomicWrite form, T-07-03-02).
func TestWriter_AppendAtomicWrite(t *testing.T) {
	path := writeWriterDoc(t, writerDocYAML)
	if err := config.AppendAutocorrectBlocklist(path, acBlockSubstr); err != nil {
		t.Fatalf("AppendAutocorrectBlocklist: %v", err)
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

// appendRefusedUntouched is the shared refusal oracle of the broken-source
// corpus: the append on the corpus returns an error and the file stays
// byte-equal — the failure-atomicity tail every subtest proves. The error
// is returned for the caller's message assertions.
func appendRefusedUntouched(t *testing.T, corpus string) error {
	t.Helper()

	path := writeWriterDoc(t, corpus)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	appendErr := config.AppendAutocorrectBlocklist(path, acBlockSubstr)
	if appendErr == nil {
		t.Fatal("the append returned nil — the broken source went unnoticed")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read untouched source: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Error("the source file changed — a failed write must leave the original byte-equal")
	}

	return appendErr
}

// TestWriter_AppendBrokenSourceUntouched pins the failure atomicity of the
// append: a source the writer cannot land (unparseable YAML, a non-mapping
// root, or an edited RESULT the strict decoder would reject) yields an error
// and leaves the original file byte-equal.
func TestWriter_AppendBrokenSourceUntouched(t *testing.T) {
	t.Run("unparseable yaml", func(t *testing.T) {
		_ = appendRefusedUntouched(t, writerBrokenYAML)
	})

	t.Run("scalar document", func(t *testing.T) {
		_ = appendRefusedUntouched(t, writerDocScalarDocumentYAML)
	})

	t.Run("sequence document", func(t *testing.T) {
		_ = appendRefusedUntouched(t, writerDocSeqDocumentYAML)
	})

	t.Run("edited result fails the strict load", func(t *testing.T) {
		err := appendRefusedUntouched(t, writerDocUnknownKeyYAML)
		if !strings.Contains(err.Error(), "validate") {
			t.Errorf("error %q does not name the strict-validation refusal", err)
		}
	})
}

// TestWriter_AppendEmptyPatternRefused pins the empty-pattern guard: an
// empty or blank pattern is a VALID RE2 that matches everything — the strict
// Load would reject the whole edited document, so the append refuses before
// any bytes land and the file stays byte-equal (T-07-02-03's trap, the
// pre-write mirror).
func TestWriter_AppendEmptyPatternRefused(t *testing.T) {
	for _, pattern := range []string{"", blankPattern} {
		path := writeWriterDoc(t, writerDocBlocklistYAML)
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read source: %v", err)
		}
		if err := config.AppendAutocorrectBlocklist(path, pattern); err == nil {
			t.Fatalf("blank pattern %q accepted, want a refusal before the write", pattern)
		}
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read untouched source: %v", err)
		}
		if !bytes.Equal(before, after) {
			t.Errorf("the append of %q changed the file — the refusal must keep the bytes", pattern)
		}
	}
}
