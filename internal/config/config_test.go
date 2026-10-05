package config_test

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Djarvur/goswitch/internal/config"
	"github.com/Djarvur/goswitch/internal/hotkey"
)

// Corpus literals named once for the strict lint's goconst (the 02-01
// wordEN/wordRU idiom); shared with load_test.go of this test package.
const (
	defTapKey        = "shift_r"
	defCombo         = "shift+ctrl_r"
	defChord         = "super+space"
	fieldTapWindow   = "timeouts.tap_window_ms"
	fieldVerifyWait  = "timeouts.verify_wait_ms"
	fieldTapKey      = "hotkeys.tap_key"
	fieldChord       = "hotkeys.mode_switch_chord"
	fieldMACRLetters = "macr.letters"
	fieldMACRApps    = "macr.apps"
	fieldMACRAltMod  = "macr.alt_modifier"

	fieldAutoCorrectWordLn = "autocorrect.min_word_len"
	fieldAutoCorrectMargin = "autocorrect.trigram_margin"
	fieldAutoCorrectFloor  = "autocorrect.trigram_floor"

	// fieldAutoCorrectBlocklist is the D-53 revision's blocklist key.
	fieldAutoCorrectBlocklist = "autocorrect.apps_blocklist"

	// fieldA11yApps is the a11y section's app-list key (D-8-1/D-8-5).
	fieldA11yApps = "a11y.apps"

	// appGedit is the corpus's bridge-namespace example app.
	appGedit = "org.gnome.Gedit"

	// Blocklist corpus patterns (D-53): a bare substring and an explicit
	// anchor — both legal RE2, both pinned by the compile corpus.
	acBlockSubstr = "chrom"
	acBlockAnchor = `^org\.gnome\.`

	// A11y corpus patterns (D-8-5, the blocklist pair's analog): a bare
	// substring and an explicit anchor — both legal RE2 — plus the
	// non-compilable broken pattern.
	a11ySubstr = "zcode"
	a11yAnchor = `^org\.gnome\.Terminal$`
	a11yBroken = "["
)

// validDoc is the complete schema document every Validate case starts from
// (the plan's config_schema corpus, all sections, all keys).
func validDoc() config.Config {
	return config.Config{
		Hotkeys:    config.Hotkeys{TapKey: defTapKey, WordLayoutCombo: defCombo},
		Timeouts:   config.Timeouts{TapWindowMs: 300, VerifyWaitMs: 100},
		Correction: config.Correction{BackspaceCap: 50, ClipboardRung: false},
		MACR:       config.MACR{Enabled: false, Letters: "", Apps: nil, AltModifier: ""},
		Autocorrect: config.Autocorrect{
			Enabled: false, AppsBlocklist: nil, MinWordLen: 4, TrigramMargin: 2.0, TrigramFloor: 1.0,
		},
	}
}

// activeAutocorrectDoc returns the complete document with the autocorrect
// layer ACTIVE — enabled with a non-empty blocklist and the documented
// thresholds — the baseline the autocorrect range and ceiling cases mutate
// (a dormant section never reaches the threshold checks: the zero value is
// the off state, D-54).
func activeAutocorrectDoc() config.Config {
	cfg := validDoc()
	cfg.Autocorrect = config.Autocorrect{
		Enabled:       true,
		AppsBlocklist: []string{appGedit},
		MinWordLen:    4,
		TrigramMargin: 2.0,
		TrigramFloor:  1.0,
	}

	return cfg
}

// activeA11yDoc returns the complete document with the a11y section
// ACTIVE — enabled with a non-empty list — the baseline the a11y corpus
// mutates (the zero section is the off state, D-54).
func activeA11yDoc() config.Config {
	cfg := validDoc()
	cfg.A11y = config.A11y{Enabled: true, Apps: []string{a11ySubstr}}

	return cfg
}

// TestDefaults pins the documented defaults (config_schema): the tap window
// is ADR-002's DefaultWindow expressed in milliseconds, the Backspace cap
// is D-27's 50, the clipboard rung is OFF by default (D-28) and MACR never
// activates on its own — the alternative modifier is not introduced until
// the live probe clears it (ADR-005 b.3).
func TestDefaults(t *testing.T) {
	t.Parallel()

	got := config.Defaults()
	if got.Hotkeys.TapKey != defTapKey {
		t.Errorf("hotkeys.tap_key = %q, want %q", got.Hotkeys.TapKey, defTapKey)
	}
	if got.Hotkeys.WordLayoutCombo != defCombo {
		t.Errorf("hotkeys.word_layout_combo = %q, want %q", got.Hotkeys.WordLayoutCombo, defCombo)
	}
	if got.Hotkeys.ModeSwitchChord != defChord {
		t.Errorf("%s = %q, want %q (owner decision 2: goswitch owns Super+Space)",
			fieldChord, got.Hotkeys.ModeSwitchChord, defChord)
	}
	if got.Timeouts.TapWindowMs != 300 {
		t.Errorf("%s = %d, want 300", fieldTapWindow, got.Timeouts.TapWindowMs)
	}
	if time.Duration(got.Timeouts.TapWindowMs)*time.Millisecond != hotkey.DefaultWindow {
		t.Errorf("%s ≠ hotkey.DefaultWindow (%v)", fieldTapWindow, hotkey.DefaultWindow)
	}
	if got.Timeouts.VerifyWaitMs != 100 {
		t.Errorf("%s = %d, want 100", fieldVerifyWait, got.Timeouts.VerifyWaitMs)
	}
	if got.Correction.BackspaceCap != 50 {
		t.Errorf("correction.backspace_cap = %d, want 50 (D-27)", got.Correction.BackspaceCap)
	}
	if got.Correction.ClipboardRung {
		t.Error("correction.clipboard_rung = true, want false (D-28 opt-in)")
	}
	if !got.Correction.FlipAfterCorrection {
		// Owner decision 2: the mode follows a changed correction.
		t.Error("correction.flip_after_correction = false, want true")
	}
	if got.MACR.Enabled {
		t.Error("macr.enabled = true, want false")
	}
	if got.MACR.AltModifier != "" {
		t.Errorf("%s = %q, want empty (ADR-005 b.3: not introduced)", fieldMACRAltMod, got.MACR.AltModifier)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Defaults() does not validate: %v", err)
	}
}

// assertRejected fails unless cfg violates validation with the given field
// named in the error — the shared shape of every whole-config rejection.
func assertRejected(t *testing.T, name string, cfg config.Config, wantField string) {
	t.Helper()

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("%s accepted, want a whole-config rejection", name)
	}
	if !strings.Contains(err.Error(), wantField) {
		t.Errorf("error %q does not name the field %q", err, wantField)
	}
}

// TestValidate_TimeoutRanges pins the millisecond ceilings (T-03-02-02,
// ASVS V5): a giant or non-positive window/verify-wait rejects the whole
// config with the field named.
func TestValidate_TimeoutRanges(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		mutate    func(*config.Config)
		wantField string
	}{
		{
			name:      "tap window zero",
			mutate:    func(c *config.Config) { c.Timeouts.TapWindowMs = 0 },
			wantField: fieldTapWindow,
		},
		{
			name:      "tap window negative",
			mutate:    func(c *config.Config) { c.Timeouts.TapWindowMs = -1 },
			wantField: fieldTapWindow,
		},
		{
			name:      "tap window above ceiling",
			mutate:    func(c *config.Config) { c.Timeouts.TapWindowMs = 2001 },
			wantField: fieldTapWindow,
		},
		{
			name:      "verify wait zero",
			mutate:    func(c *config.Config) { c.Timeouts.VerifyWaitMs = 0 },
			wantField: fieldVerifyWait,
		},
		{
			name:      "verify wait above ceiling",
			mutate:    func(c *config.Config) { c.Timeouts.VerifyWaitMs = 2001 },
			wantField: fieldVerifyWait,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := validDoc()
			tc.mutate(&cfg)
			assertRejected(t, tc.name, cfg, tc.wantField)
		})
	}
}

// TestValidate_CorrectionRanges pins the Backspace cap range (D-27) and
// the MACR app-list ceiling: boundary violations reject the whole config
// with the field named.
func TestValidate_CorrectionRanges(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		mutate    func(*config.Config)
		wantField string
	}{
		{
			name:      "backspace cap zero",
			mutate:    func(c *config.Config) { c.Correction.BackspaceCap = 0 },
			wantField: "correction.backspace_cap",
		},
		{
			name:      "backspace cap above ceiling",
			mutate:    func(c *config.Config) { c.Correction.BackspaceCap = 501 },
			wantField: "correction.backspace_cap",
		},
		{
			name: "macr apps above ceiling",
			mutate: func(c *config.Config) {
				c.MACR.Apps = make([]string, 65)
			},
			wantField: fieldMACRApps,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := validDoc()
			tc.mutate(&cfg)
			assertRejected(t, tc.name, cfg, tc.wantField)
		})
	}
}

// TestValidate_BoundariesAccepted pins that the ceilings are inclusive:
// 2000 ms windows, a 500-press cap and a 64-entry app list all validate.
func TestValidate_BoundariesAccepted(t *testing.T) {
	t.Parallel()

	cfg := validDoc()
	cfg.Timeouts.TapWindowMs = 2000
	cfg.Timeouts.VerifyWaitMs = 2000
	cfg.Correction.BackspaceCap = 500
	cfg.MACR.Apps = make([]string, 64)
	if err := cfg.Validate(); err != nil {
		t.Errorf("ceiling values rejected: %v", err)
	}
}

// TestValidate_Bindings pins the binding resolution path: both key names
// resolve through hotkey.ParseBinding and an unknown token rejects the
// whole config with the field named.
func TestValidate_Bindings(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		mutate    func(*config.Config)
		wantField string
	}{
		{
			name:      "tap key unknown",
			mutate:    func(c *config.Config) { c.Hotkeys.TapKey = "shift_x" },
			wantField: fieldTapKey,
		},
		{
			name:      "tap key empty",
			mutate:    func(c *config.Config) { c.Hotkeys.TapKey = "" },
			wantField: fieldTapKey,
		},
		{
			// WR-02: tap semantics are key-only — the FSM tracks one keyval
			// and no modifier state, so a modifier-bearing tap_key would
			// "apply" while its modifier tokens are silently ignored.
			name:      "tap key combo (modifier-bearing)",
			mutate:    func(c *config.Config) { c.Hotkeys.TapKey = "ctrl+shift_r" },
			wantField: fieldTapKey,
		},
		{
			name:      "tap key redundant family modifier",
			mutate:    func(c *config.Config) { c.Hotkeys.TapKey = "shift+shift_r" },
			wantField: fieldTapKey,
		},
		{
			name:      "combo key unknown",
			mutate:    func(c *config.Config) { c.Hotkeys.WordLayoutCombo = "shift+ctrl_q" },
			wantField: "hotkeys.word_layout_combo",
		},
		{
			name:      "combo modifier unknown",
			mutate:    func(c *config.Config) { c.Hotkeys.WordLayoutCombo = "shif+ctrl_r" },
			wantField: "hotkeys.word_layout_combo",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := validDoc()
			tc.mutate(&cfg)
			assertRejected(t, tc.name, cfg, tc.wantField)
		})
	}
}

// TestValidate_MACRLetters pins the letter grammar: required once the
// layer is on, only single lowercase a–z tokens allowed (the Cyrillic
// keyboard has no business here — MACR forwards Latin Ctrl-shortcuts).
func TestValidate_MACRLetters(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		mutate func(*config.Config)
	}{
		{
			name: "enabled without letters",
			mutate: func(c *config.Config) {
				c.MACR.Enabled = true
			},
		},
		{
			name: "cyrillic letter",
			mutate: func(c *config.Config) {
				c.MACR.Enabled = true
				c.MACR.Letters = "a,я"
			},
		},
		{
			name: "multi-letter token",
			mutate: func(c *config.Config) {
				c.MACR.Enabled = true
				c.MACR.Letters = "ab,z"
			},
		},
		{
			name: "uppercase letter",
			mutate: func(c *config.Config) {
				c.MACR.Enabled = true
				c.MACR.Letters = "A"
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := validDoc()
			tc.mutate(&cfg)
			assertRejected(t, tc.name, cfg, fieldMACRLetters)
		})
	}
}

// TestValidate_MACRAltModifier pins the alternative-modifier closed set
// (ADR-005 b.3): ctrl_l/ctrl_r/empty validate, everything else — including
// a real modifier name that is not a candidate — rejects.
func TestValidate_MACRAltModifier(t *testing.T) {
	t.Parallel()

	cfg := validDoc()
	cfg.MACR.AltModifier = "shift"
	assertRejected(t, "alt modifier not a candidate", cfg, fieldMACRAltMod)

	accepted := validDoc()
	accepted.MACR = config.MACR{
		Enabled:     true,
		Letters:     "a,z",
		Apps:        []string{"google-chrome"},
		AltModifier: "ctrl_l",
	}
	if err := accepted.Validate(); err != nil {
		t.Errorf("valid MACR block rejected: %v", err)
	}
}

// TestValidate_AutocorrectRanges pins the threshold ranges on an ACTIVE
// section (enabled with a non-empty white list — the only shape the
// thresholds can bite in): min_word_len in [2, 16], both trigram
// thresholds positive with margin ≥ floor, every rejection naming its
// field; the boundaries 2/16 and margin == floor validate.
func TestValidate_AutocorrectRanges(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		mutate    func(*config.Config)
		wantField string
	}{
		{
			name:      "min word len zero",
			mutate:    func(c *config.Config) { c.Autocorrect.MinWordLen = 0 },
			wantField: fieldAutoCorrectWordLn,
		},
		{
			name:      "min word len below floor",
			mutate:    func(c *config.Config) { c.Autocorrect.MinWordLen = 1 },
			wantField: fieldAutoCorrectWordLn,
		},
		{
			name:      "min word len above ceiling",
			mutate:    func(c *config.Config) { c.Autocorrect.MinWordLen = 17 },
			wantField: fieldAutoCorrectWordLn,
		},
		{
			name:      "trigram margin zero",
			mutate:    func(c *config.Config) { c.Autocorrect.TrigramMargin = 0 },
			wantField: fieldAutoCorrectMargin,
		},
		{
			name:      "trigram margin negative",
			mutate:    func(c *config.Config) { c.Autocorrect.TrigramMargin = -1.5 },
			wantField: fieldAutoCorrectMargin,
		},
		{
			name:      "trigram margin below floor",
			mutate:    func(c *config.Config) { c.Autocorrect.TrigramMargin = 0.5 },
			wantField: fieldAutoCorrectMargin,
		},
		{
			name:      "trigram floor zero",
			mutate:    func(c *config.Config) { c.Autocorrect.TrigramFloor = 0 },
			wantField: fieldAutoCorrectFloor,
		},
		{
			name:      "trigram floor negative",
			mutate:    func(c *config.Config) { c.Autocorrect.TrigramFloor = -0.5 },
			wantField: fieldAutoCorrectFloor,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := activeAutocorrectDoc()
			tc.mutate(&cfg)
			assertRejected(t, tc.name, cfg, tc.wantField)
		})
	}
}

// TestValidate_AutocorrectRangeBoundaries pins that the ranges are
// inclusive: min_word_len at both ends (2 and 16) and margin == floor
// validate on an active section.
func TestValidate_AutocorrectRangeBoundaries(t *testing.T) {
	t.Parallel()

	boundary := activeAutocorrectDoc()
	boundary.Autocorrect.MinWordLen = 2
	boundary.Autocorrect.TrigramMargin = 1.0
	boundary.Autocorrect.TrigramFloor = 1.0
	if err := boundary.Validate(); err != nil {
		t.Errorf("min_word_len 2 with margin == floor rejected: %v", err)
	}
	boundary.Autocorrect.MinWordLen = 16
	if err := boundary.Validate(); err != nil {
		t.Errorf("min_word_len 16 rejected: %v", err)
	}
}

// TestValidate_AutocorrectDormantShapesValid pins the D-54 off states at
// the validation level: the zero section (a document without the section)
// and enabled-with-an-empty-list are both VALID — an empty white list
// silences the feature (T-06-04-03), it never widens it, so no combination
// of absent or empty values can turn autocorrect on.
func TestValidate_AutocorrectDormantShapesValid(t *testing.T) {
	t.Parallel()

	zero := validDoc()
	zero.Autocorrect = config.Autocorrect{}
	if err := zero.Validate(); err != nil {
		t.Errorf("zero autocorrect section rejected: %v", err)
	}

	// Re-pinned by plan 07-02 (the ACTIVE condition — Q4c): an enabled
	// section IS the active state (an empty blocklist forbids nothing), so
	// the thresholds are mandatory — the 06-04 zero-threshold shape is now
	// a loud refusal naming its field.
	enabledEmpty := validDoc()
	enabledEmpty.Autocorrect = config.Autocorrect{Enabled: true, AppsBlocklist: []string{}}
	assertRejected(t, "enabled without thresholds", enabledEmpty, fieldAutoCorrectWordLn)
}

// TestValidate_EnabledRequiresThresholds pins the ACTIVE condition
// (T-07-02-05): enabled: true with zero thresholds would fire the
// detector with a degenerate min_word_len=0 — validation refuses loudly,
// every rejection naming its own field.
func TestValidate_EnabledRequiresThresholds(t *testing.T) {
	t.Parallel()

	t.Run("min word len zero", func(t *testing.T) {
		t.Parallel()

		cfg := validDoc()
		cfg.Autocorrect = config.Autocorrect{Enabled: true}
		assertRejected(t, "enabled zero thresholds", cfg, fieldAutoCorrectWordLn)
	})

	t.Run("trigram thresholds zero", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name      string
			mutate    func(*config.Autocorrect)
			wantField string
		}{
			{
				name:      "trigram margin zero",
				mutate:    func(a *config.Autocorrect) { a.TrigramMargin = 0 },
				wantField: fieldAutoCorrectMargin,
			},
			{
				name:      "trigram floor zero",
				mutate:    func(a *config.Autocorrect) { a.TrigramFloor = 0 },
				wantField: fieldAutoCorrectFloor,
			},
		} {
			cfg := validDoc()
			cfg.Autocorrect = config.Autocorrect{
				Enabled:       true,
				MinWordLen:    4,
				TrigramMargin: 2.0,
				TrigramFloor:  1.0,
			}
			tc.mutate(&cfg.Autocorrect)
			assertRejected(t, tc.name, cfg, tc.wantField)
		}
	})
}

// TestValidate_EnabledEmptyBlocklistWithThresholdsValid pins the new
// active shape: enabled + a NIL blocklist + the documented thresholds is
// VALID — an empty blocklist forbids nothing, so the layer fires wherever
// the field's role gate lets it through (Q4c).
func TestValidate_EnabledEmptyBlocklistWithThresholdsValid(t *testing.T) {
	t.Parallel()

	cfg := validDoc()
	cfg.Autocorrect = config.Autocorrect{
		Enabled:       true,
		AppsBlocklist: nil,
		MinWordLen:    4,
		TrigramMargin: 2.0,
		TrigramFloor:  1.0,
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("enabled with an empty blocklist and thresholds rejected: %v", err)
	}
}

// TestValidate_DisabledZeroThresholdsValid pins the D-54 dormant shape:
// enabled: false with zero thresholds stays VALID — the threshold checks
// never run on a dormant layer, so the zero value remains the off state.
func TestValidate_DisabledZeroThresholdsValid(t *testing.T) {
	t.Parallel()

	cfg := validDoc()
	cfg.Autocorrect = config.Autocorrect{Enabled: false}
	if err := cfg.Validate(); err != nil {
		t.Errorf("dormant section with zero thresholds rejected: %v", err)
	}
}

// TestValidate_BlocklistRegexCompile pins the D-53 compile gate: every
// blocklist pattern is compiled by regexp on Load — the only place a
// broken regex is visible before runtime — and a non-compilable pattern
// rejects the whole config with the field AND the element index named
// (D-33). A bare substring and an explicit anchor both compile.
func TestValidate_BlocklistRegexCompile(t *testing.T) {
	t.Parallel()

	broken := activeAutocorrectDoc()
	broken.Autocorrect.AppsBlocklist = []string{appGedit, "["}
	err := broken.Validate()
	if err == nil {
		t.Fatal("broken blocklist pattern accepted, want a whole-config rejection")
	}
	if !strings.Contains(err.Error(), fieldAutoCorrectBlocklist) || !strings.Contains(err.Error(), "[1]") {
		t.Errorf("error %q does not name %q and the element index [1]", err, fieldAutoCorrectBlocklist)
	}

	valid := activeAutocorrectDoc()
	valid.Autocorrect.AppsBlocklist = []string{acBlockSubstr, acBlockAnchor}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid blocklist patterns rejected: %v", err)
	}
}

// TestValidate_BlocklistEmptyPattern pins the empty-pattern trap
// (T-07-02-03): an empty or blank pattern is a VALID RE2 that matches
// EVERYTHING — a silent "block everywhere" — so validation refuses it
// with its own sentinel and the index, before any compile attempt.
func TestValidate_BlocklistEmptyPattern(t *testing.T) {
	t.Parallel()

	for _, pattern := range []string{"", "   "} {
		cfg := activeAutocorrectDoc()
		cfg.Autocorrect.AppsBlocklist = []string{appGedit, pattern}
		err := cfg.Validate()
		if err == nil {
			t.Fatalf("blank blocklist pattern %q accepted, want a whole-config rejection", pattern)
		}
		if !strings.Contains(err.Error(), fieldAutoCorrectBlocklist) || !strings.Contains(err.Error(), "[1]") {
			t.Errorf("error %q does not name %q and the element index [1]", err, fieldAutoCorrectBlocklist)
		}
		if !strings.Contains(err.Error(), "matches everything") {
			t.Errorf("error %q does not cite the empty-pattern trap", err)
		}
	}
}

// TestValidate_BlocklistCeilUnconditional pins the blocklist DoS ceiling
// (T-07-02-02, the T-06-04-01 precedent): 65 patterns reject the whole
// config even with the layer DORMANT — the ceiling does not depend on
// enabled — with the field and the 64-entry limit named; 64 validate in
// both the dormant and the active shape.
func TestValidate_BlocklistCeilUnconditional(t *testing.T) {
	t.Parallel()

	over := validDoc()
	over.Autocorrect.AppsBlocklist = make([]string, 65)
	err := over.Validate()
	if err == nil {
		t.Fatal("65 blocklist patterns accepted while dormant, want a whole-config rejection")
	}
	if !strings.Contains(err.Error(), fieldAutoCorrectBlocklist) || !strings.Contains(err.Error(), "64") {
		t.Errorf("error %q does not name %q and the 64-entry limit", err, fieldAutoCorrectBlocklist)
	}

	overActive := activeAutocorrectDoc()
	overActive.Autocorrect.AppsBlocklist = make([]string, 65)
	if err := overActive.Validate(); err == nil {
		t.Fatal("65 blocklist patterns accepted while active, want a whole-config rejection")
	}

	atCeil := validDoc()
	atCeil.Autocorrect.AppsBlocklist = slices.Repeat([]string{acBlockSubstr}, 64)
	if err := atCeil.Validate(); err != nil {
		t.Errorf("64-entry blocklist rejected: %v", err)
	}
}

// TestValidate_BlocklistOrderIrrelevant pins the blocklist's set
// semantics (D-53): the same patterns in a different order validate
// identically — any match forbids, so order carries no meaning.
func TestValidate_BlocklistOrderIrrelevant(t *testing.T) {
	t.Parallel()

	first := activeAutocorrectDoc()
	first.Autocorrect.AppsBlocklist = []string{acBlockSubstr, appGedit, acBlockAnchor}
	second := activeAutocorrectDoc()
	second.Autocorrect.AppsBlocklist = []string{acBlockAnchor, acBlockSubstr, appGedit}

	if err := first.Validate(); err != nil {
		t.Errorf("first order rejected: %v", err)
	}
	if err := second.Validate(); err != nil {
		t.Errorf("second order rejected: %v", err)
	}
}

// TestDefaults_AutocorrectBlocklistNil pins the D-54 default-off shape
// under the D-53 schema: the feature ships disabled with a nil blocklist
// and the start thresholds 4/2.0/1.0 — nothing activates on its own, and
// the defaults still validate as a whole.
func TestDefaults_AutocorrectBlocklistNil(t *testing.T) {
	t.Parallel()

	got := config.Defaults().Autocorrect
	if got.Enabled {
		t.Error("Defaults().Autocorrect.Enabled = true, want false (D-54 default off)")
	}
	if got.AppsBlocklist != nil {
		t.Errorf("Defaults().Autocorrect.AppsBlocklist = %v, want nil (the off state)", got.AppsBlocklist)
	}
	if got.MinWordLen != 4 {
		t.Errorf("Defaults().Autocorrect.MinWordLen = %d, want 4", got.MinWordLen)
	}
	if got.TrigramMargin != 2.0 {
		t.Errorf("Defaults().Autocorrect.TrigramMargin = %v, want 2.0", got.TrigramMargin)
	}
	if got.TrigramFloor != 1.0 {
		t.Errorf("Defaults().Autocorrect.TrigramFloor = %v, want 1.0", got.TrigramFloor)
	}
	if err := config.Defaults().Validate(); err != nil {
		t.Errorf("Defaults() does not validate with the autocorrect section: %v", err)
	}
}

// TestDefaults_SoundOn pins the sound section's built-in defaults (owner
// decision «Звуки при переключении»): the sounds ship ON and the
// autocorrect tone carries the default event name — both constants
// verified present in this desktop's Yaru sound theme — and the defaults
// still validate as a whole (the section carries no validation).
func TestDefaults_SoundOn(t *testing.T) {
	t.Parallel()

	got := config.Defaults().Sound
	if !got.EffectiveEnabled() {
		t.Error("Defaults().Sound.EffectiveEnabled() = false, want true (default ON — owner decision)")
	}
	if got.EffectiveAutocorrectEvent() != config.DefaultSoundAutocorrectEvent {
		t.Errorf(
			"Defaults().Sound.EffectiveAutocorrectEvent() = %q, want %q",
			got.EffectiveAutocorrectEvent(), config.DefaultSoundAutocorrectEvent,
		)
	}
	flipEvent, corrEvent := config.DefaultSoundFlipEvent, config.DefaultSoundAutocorrectEvent
	if flipEvent != "bell" || corrEvent != "message" {
		t.Errorf("sound event constants = %q/%q, want the verified Yaru names bell/message", flipEvent, corrEvent)
	}
	if err := config.Defaults().Validate(); err != nil {
		t.Errorf("Defaults() does not validate with the sound section: %v", err)
	}
}

// TestValidate_A11yCeilingUnconditional pins the a11y app-list DoS ceiling
// (T-08-02-01, the T-06-04-01/maxAutocorrectBlocklist precedent): 65
// patterns reject the whole config even with the section DORMANT — the
// ceiling is UNCONDITIONAL, it never keys on enabled — with the field and
// the 64-entry limit named; 64 validate in both the dormant and the active
// shape.
func TestValidate_A11yCeilingUnconditional(t *testing.T) {
	t.Parallel()

	over := validDoc()
	over.A11y = config.A11y{Enabled: false, Apps: make([]string, 65)}
	err := over.Validate()
	if err == nil {
		t.Fatal("65 a11y patterns accepted while dormant, want a whole-config rejection")
	}
	if !strings.Contains(err.Error(), fieldA11yApps) || !strings.Contains(err.Error(), "64") {
		t.Errorf("error %q does not name %q and the 64-entry limit", err, fieldA11yApps)
	}

	atCeilDormant := validDoc()
	atCeilDormant.A11y = config.A11y{Enabled: false, Apps: slices.Repeat([]string{a11ySubstr}, 64)}
	if err := atCeilDormant.Validate(); err != nil {
		t.Errorf("64-entry dormant a11y list rejected: %v", err)
	}

	atCeilActive := activeA11yDoc()
	atCeilActive.A11y.Apps = slices.Repeat([]string{a11ySubstr}, 64)
	if err := atCeilActive.Validate(); err != nil {
		t.Errorf("64-entry active a11y list rejected: %v", err)
	}
}

// TestValidate_A11yBlankPattern pins the empty-pattern trap (T-08-02-02,
// the T-07-02-03 blocklist precedent): an empty or blank pattern is a VALID
// RE2 that matches EVERYTHING — silent magic "for every application at
// once" — so validation refuses it with its own sentinel and the index,
// before any compile attempt.
func TestValidate_A11yBlankPattern(t *testing.T) {
	t.Parallel()

	for _, pattern := range []string{"", "   "} {
		cfg := activeA11yDoc()
		cfg.A11y.Apps = []string{a11ySubstr, pattern}
		err := cfg.Validate()
		if err == nil {
			t.Fatalf("blank a11y pattern %q accepted, want a whole-config rejection", pattern)
		}
		if !strings.Contains(err.Error(), fieldA11yApps) || !strings.Contains(err.Error(), "[1]") {
			t.Errorf("error %q does not name %q and the element index [1]", err, fieldA11yApps)
		}
		if !strings.Contains(err.Error(), "matches everything") {
			t.Errorf("error %q does not cite the empty-pattern trap", err)
		}
	}
}

// TestValidate_A11yRegexCompile pins the D-8-5 compile gate: every a11y
// pattern is compiled by regexp on Load — the only place a broken regex is
// visible before runtime — and a non-compilable pattern rejects the whole
// config with the field AND the element index named (D-33). A bare
// substring and an explicit anchor both compile.
func TestValidate_A11yRegexCompile(t *testing.T) {
	t.Parallel()

	broken := activeA11yDoc()
	broken.A11y.Apps = []string{a11ySubstr, a11yBroken}
	err := broken.Validate()
	if err == nil {
		t.Fatal("broken a11y pattern accepted, want a whole-config rejection")
	}
	if !strings.Contains(err.Error(), fieldA11yApps) || !strings.Contains(err.Error(), "[1]") {
		t.Errorf("error %q does not name %q and the element index [1]", err, fieldA11yApps)
	}

	valid := activeA11yDoc()
	valid.A11y.Apps = []string{a11ySubstr, a11yAnchor}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid a11y patterns rejected: %v", err)
	}
}

// TestValidate_A11yOrderIrrelevant pins the a11y list's set semantics
// (D-8-5, the D-53 blocklist precedent): the same patterns in a different
// order validate identically — any match enables, so order carries no
// meaning.
func TestValidate_A11yOrderIrrelevant(t *testing.T) {
	t.Parallel()

	first := activeA11yDoc()
	first.A11y.Apps = []string{a11ySubstr, appGedit, a11yAnchor}
	second := activeA11yDoc()
	second.A11y.Apps = []string{a11yAnchor, a11ySubstr, appGedit}

	if err := first.Validate(); err != nil {
		t.Errorf("first order rejected: %v", err)
	}
	if err := second.Validate(); err != nil {
		t.Errorf("second order rejected: %v", err)
	}
}

// TestA11yActive pins the ACTIVE semantics of the section — the single
// definition every consumer reads (the actor fold of plan 08-05):
// enabled AND a non-empty list. enabled with an empty list is a valid
// document but NOT active (nothing to apply); the zero value is off.
func TestA11yActive(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		a    config.A11y
		want bool
	}{
		{name: "zero value is off", a: config.A11y{}, want: false},
		{name: "enabled with an empty list", a: config.A11y{Enabled: true, Apps: []string{}}, want: false},
		{name: "enabled with a non-empty list", a: config.A11y{Enabled: true, Apps: []string{a11ySubstr}}, want: true},
		{
			name: "disabled with a non-empty list",
			a:    config.A11y{Enabled: false, Apps: []string{a11ySubstr}},
			want: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.a.Active(); got != tc.want {
				t.Errorf("A11y(%+v).Active() = %v, want %v", tc.a, got, tc.want)
			}
		})
	}
}

// TestDefaults_A11yOff pins the D-54 default-off shape for the a11y
// section: the magic ships disabled with a nil list — nothing activates on
// its own — and the defaults still validate as a whole.
func TestDefaults_A11yOff(t *testing.T) {
	t.Parallel()

	got := config.Defaults().A11y
	if got.Enabled {
		t.Error("Defaults().A11y.Enabled = true, want false (D-54 default off)")
	}
	if got.Apps != nil {
		t.Errorf("Defaults().A11y.Apps = %v, want nil (the off state)", got.Apps)
	}
	if err := config.Defaults().Validate(); err != nil {
		t.Errorf("Defaults() does not validate with the a11y section: %v", err)
	}
}

// chordDocYAML renders a complete document with the given
// hotkeys.mode_switch_chord YAML value — the decode corpus's template (the
// clipboard_rung idiom: full document, one key under test).
func chordDocYAML(value string) string {
	return `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
  mode_switch_chord: ` + value + `
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
}

// chordDocNoKeyYAML is the same document WITHOUT the mode_switch_chord key
// — every pre-existing document's shape.
const chordDocNoKeyYAML = `hotkeys:
  tap_key: shift_r
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

// TestLoad_ModeSwitchChord pins the decode/validation surface of the chord
// key (owner decision 2, quick plan 260927-way): an explicit value decodes
// and validates, an EXPLICIT empty string decodes "" and validates (chord
// disabled — the missing-key compatibility rule of the vu8 batch, the
// macr.alt_modifier empty precedent, so every pre-existing document loads
// unchanged), and a garbage value refuses the WHOLE document (D-33).
func TestLoad_ModeSwitchChord(t *testing.T) {
	t.Parallel()

	t.Run("explicit value decodes and validates", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.Load(writeConfig(t, chordDocYAML("ctrl+space")))
		if err != nil {
			t.Fatalf("Load(ctrl+space doc): %v", err)
		}
		if cfg.Hotkeys.ModeSwitchChord != "ctrl+space" {
			t.Errorf("%s = %q, want %q", fieldChord, cfg.Hotkeys.ModeSwitchChord, "ctrl+space")
		}
	})

	t.Run("explicit empty disables", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.Load(writeConfig(t, chordDocYAML(`""`)))
		if err != nil {
			t.Fatalf("Load(empty doc): %v", err)
		}
		if cfg.Hotkeys.ModeSwitchChord != "" {
			t.Errorf("%s = %q, want %q (empty = the chord is disabled)", fieldChord, cfg.Hotkeys.ModeSwitchChord, "")
		}
	})

	t.Run("document without the key loads unchanged", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.Load(writeConfig(t, chordDocNoKeyYAML))
		if err != nil {
			t.Fatalf("Load(pre-batch doc): %v", err)
		}
		if cfg.Hotkeys.ModeSwitchChord != "" {
			t.Errorf("%s = %q, want %q — a missing key decodes disabled, never defaulted",
				fieldChord, cfg.Hotkeys.ModeSwitchChord, "")
		}
	})

	t.Run("garbage refuses the whole document", func(t *testing.T) {
		t.Parallel()

		_, err := config.Load(writeConfig(t, chordDocYAML("super+spaces")))
		if err == nil {
			t.Fatalf("Load(garbage chord) = nil error, want a whole-document rejection (D-33)")
		}
		if !strings.Contains(err.Error(), fieldChord) {
			t.Errorf("error %q does not name the field %q", err, fieldChord)
		}
	})
}
