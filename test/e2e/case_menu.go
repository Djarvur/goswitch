package main

// The menu-v2 live proof (plan 07-06): the tray menu v2 driven over the
// DBusMenu Event channel — no ydotool, no mouse (07-RESEARCH Q1b: the
// channel was proven live on the working daemon, rc=0). The case clicks
// the pinned item ids (internal/indicator/menu.go, plan 07-05 — a layout
// change breaks this case loudly), asserts the journal mode records of the
// EN/RU radio pair, both persisted toggles (autocorrect id 4, sound id 5:
// the document on disk AND the ctl status flip in the same click — the
// synchronous-apply pin), and the RESTART survival of both enabled values
// (the locked «переживает перезапуск», automated). The sound oracle
// discipline: playback is NEVER asserted live (unreadable from the stand);
// the engine-invocation proof is the 07-08 unit corpus on the runner seam —
// this case proves ONLY the toggle's persist/status mechanics.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// The pinned DBusMenu surface and item ids — the daemon's own exported
// object (internal/indicator/menu.go, plan 07-05). A menu re-layout that
// moves these ids breaks this case loudly, by design (the key_links pin).
const (
	menuBusName = "org.djarvur.goswitch"
	menuObjPath = "/Menu"
	menuIface   = "com.canonical.dbusmenu"

	// menuEventMethod is the click dispatch of the canonical set — the
	// grep-stable pin of the wire method the case drives.
	menuEventMethod = "com.canonical.dbusmenu.Event"

	menuIDEN          = int32(1)
	menuIDRU          = int32(2)
	menuIDACToggle    = int32(4)
	menuIDSoundToggle = int32(5)

	// caseMenuV2 is the case's registry name — the error messages' prefix
	// (a package const, not a threaded parameter: one case, one name).
	caseMenuV2 = "menu-v2"
)

// menuWait bounds every menu-step oracle (click → record/file/status); the
// toggle apply is synchronous (the 07-05 pin), the budget only absorbs the
// read-after-write skew. menuStatusPoll is the poll interval.
const (
	menuWait       = 5 * time.Second
	menuStatusPoll = 100 * time.Millisecond
)

// menuConfigDoc is the case's complete config document (the 03-02
// no-overlay rule), starting BOTH toggles OFF: the clicks invert the
// APPLIED value (the 07-05 configToggle), so enabled:false starts give the
// on-click oracles an unambiguous direction — and the sound section is
// EXPLICIT (the existing-section form, plan 07-06): a default-ON absent
// section would make the first click's write the section's birth, and the
// oracle would read the inversion through EffectiveEnabled's nil-means-on
// rule instead of the document.
const menuConfigDoc = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: false # no corrections in this case — the mode records stay the radio pair's
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
autocorrect:
  enabled: false
  apps_blocklist: []
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
sound:
  enabled: false
`

// parseMenuToggles decodes the persisted document's two toggle-owned
// scalars: autocorrect.enabled and sound.enabled (the sound value is the
// DOCUMENT's — the case drives explicit sections, never the nil-means-on
// default). The decode is deliberately partial: everything else in the
// document is the daemon's business, not the menu case's.
func parseMenuToggles(data []byte) (ac, sound bool, err error) {
	var doc struct {
		Autocorrect struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"autocorrect"`
		Sound struct {
			Enabled *bool `yaml:"enabled"`
		} `yaml:"sound"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false, false, fmt.Errorf("parse menu toggles: %w", err)
	}
	sound = doc.Sound.Enabled != nil && *doc.Sound.Enabled

	return doc.Autocorrect.Enabled, sound, nil
}

// waitMenuToggles polls the persisted document until both toggle scalars
// read exactly want — the disk half of the synchronous-apply oracle.
func (s *stand) waitMenuToggles(ctx context.Context, cfgPath string, ac, sound bool) error {
	want := fmt.Sprintf("autocorrect=%t sound=%t", ac, sound)
	deadline := time.Now().Add(menuWait)
	var last string
	for {
		data, err := os.ReadFile(cfgPath)
		if err == nil {
			gotAC, gotSound, perr := parseMenuToggles(data)
			if perr == nil {
				last = fmt.Sprintf("autocorrect=%t sound=%t", gotAC, gotSound)
				if gotAC == ac && gotSound == sound {
					return nil
				}
			} else {
				last = "parse error: " + perr.Error()
			}
		} else {
			last = err.Error()
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("persisted document did not settle to %s (last %s)", want, last)
		}
		if err := sleepCtx(ctx, menuStatusPoll); err != nil {
			return err
		}
	}
}

// menuStatusHas polls goswitchctl status until every want token is present
// — the status half of the synchronous-apply oracle (the apply happens
// inside the click; the poll absorbs only the ctl round-trip skew).
func menuStatusHas(ctx context.Context, ctlBin string, want ...string) error {
	deadline := time.Now().Add(menuWait)
	var last string
	for {
		out, _, err := runCtl(ctx, ctlBin, "status")
		if err == nil {
			last = out
			missing := false
			for _, token := range want {
				if !strings.Contains(out, token) {
					missing = true

					break
				}
			}
			if !missing {
				return nil
			}
		} else {
			last = err.Error()
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("menu status %q missing one of %v", last, want)
		}
		if err := sleepCtx(ctx, menuStatusPoll); err != nil {
			return err
		}
	}
}

// startMenuDaemon writes the case's document and runs the -config spawn
// ladder (the startAutocorrectConfigDaemon form): activate → write →
// restart → the three readiness marks → re-activate.
func startMenuDaemon(ctx context.Context, s *stand, doc string) (string, error) {
	if err := s.activateGoswitch(ctx); err != nil {
		return "", err
	}
	cfgPath := filepath.Join(s.tmpDir, "menu.yaml")
	if err := os.WriteFile(cfgPath, []byte(doc), configFilePerm); err != nil {
		return "", fmt.Errorf("menu-v2: write temp config: %w", err)
	}

	return cfgPath, spawnMenuDaemon(ctx, s, cfgPath)
}

// spawnMenuDaemon restarts the daemon on an EXISTING document (the
// restart-survival step: the toggled file on disk is the thing being
// proven — never rewritten) and waits the readiness marks + re-activates.
func spawnMenuDaemon(ctx context.Context, s *stand, cfgPath string) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	if err := s.restartDaemonWithArgs("-config", cfgPath); err != nil {
		return err
	}
	for _, mark := range []string{configLoadedMark, componentRegisteredMark, ctlListeningMark} {
		if err := s.waitForLog(ctx, mark, registrationWait); err != nil {
			return fmt.Errorf("menu-v2 daemon -config spawn (%s): %w", mark, err)
		}
	}

	return s.activateGoswitch(ctx)
}

// runMenuV2 drives the menu v2 over the Event channel: the radio pair,
// both persisted toggles on, the restart survival, both toggles back off —
// each round in its own helper (the case body stays the step list).
func runMenuV2(ctx context.Context, s *stand) error {
	const caseName = caseMenuV2
	cfgPath, err := startMenuDaemon(ctx, s, menuConfigDoc)
	if err != nil {
		return err
	}
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}
	if err := menuRadioRound(ctx, s); err != nil {
		return err
	}
	// (в) the autocorrect toggle: file AND status flip in the same click,
	// the sound value UNTOUCHED (the intermediate state is part of the
	// oracle).
	if err := menuToggleStep(ctx, s, ctlBin, cfgPath, menuIDACToggle, true, false); err != nil {
		return err
	}
	// (в-бис) the sound toggle: same mechanism, same synchronous apply.
	if err := menuToggleStep(ctx, s, ctlBin, cfgPath, menuIDSoundToggle, true, true); err != nil {
		return err
	}
	// (г) the restart survival: the same -config PATH, a fresh daemon, the
	// toggled file on disk NEVER rewritten — BOTH enabled values read back
	// from disk and status.
	if err := spawnMenuDaemon(ctx, s, cfgPath); err != nil {
		return err
	}
	if err := s.waitMenuToggles(ctx, cfgPath, true, true); err != nil {
		return fmt.Errorf("%s: restart survival (document): %w", caseName, err)
	}
	if err := menuStatusHas(ctx, ctlBin, "autocorrect_enabled=true", "sound_enabled=true"); err != nil {
		return fmt.Errorf("%s: restart survival (status): %w", caseName, err)
	}
	// (д) the reverse clicks: both scalars back to false on disk and in
	// the status — the intermediate state after the first click is
	// false/TRUE (the sound value is still on; the mirror image of the
	// on-round's true/false).
	if err := menuToggleStep(ctx, s, ctlBin, cfgPath, menuIDACToggle, false, true); err != nil {
		return err
	}
	if err := menuToggleStep(ctx, s, ctlBin, cfgPath, menuIDSoundToggle, false, false); err != nil {
		return err
	}

	fmt.Printf("%s: EN/RU radio records, both toggles persist+apply synchronously, "+
		"both enabled values survived the restart, reverse clicks restored the off state — green\n", caseName)

	return nil
}

// menuRadioRound clicks the EN/RU radio pair and gates the journal mode
// records (the byte-stable flipTo records, ADR-006 — the SAME path every
// gesture shares, the SWCH regression guard). The first click drives RU:
// the daemon starts in EN and flipTo's same-target guard would no-op an EN
// click first.
func menuRadioRound(ctx context.Context, s *stand) error {
	ruMark := modeRecordMark + `"ru"`
	enMark := modeRecordMark + `"en"`
	ruBase := s.countSub(ruMark)
	if err := eventClick(ctx, menuIDRU); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, ruMark, ruBase+1, decisionWait); err != nil {
		return fmt.Errorf("%s: RU click mode record: %w", caseMenuV2, err)
	}
	enBase := s.countSub(enMark)
	if err := eventClick(ctx, menuIDEN); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, enMark, enBase+1, decisionWait); err != nil {
		return fmt.Errorf("%s: EN click mode record: %w", caseMenuV2, err)
	}

	return nil
}

// menuToggleStep is ONE toggle click with its full synchronous-apply
// oracle: after the Event returns, the persisted document AND the ctl
// status must read exactly the step's want values (the 07-05 pin — the
// apply happens inside the click, never a write-only toggle).
func menuToggleStep(ctx context.Context, s *stand, ctlBin, cfgPath string, id int32, ac, sound bool) error {
	if err := eventClick(ctx, id); err != nil {
		return err
	}
	if err := s.waitMenuToggles(ctx, cfgPath, ac, sound); err != nil {
		return fmt.Errorf("%s: toggle persist (Event %d): %w", caseMenuV2, id, err)
	}
	if err := menuStatusHas(ctx, ctlBin,
		"autocorrect_enabled="+strconv.FormatBool(ac),
		"sound_enabled="+strconv.FormatBool(sound)); err != nil {
		return fmt.Errorf("%s: toggle status (Event %d): %w", caseMenuV2, id, err)
	}

	return nil
}

// eventClick delivers one DBusMenu click over the session bus — the
// research Q1b channel (Event on the daemon's OWN exported /Menu object,
// proven live rc=0), never a blind injection. The Event call is
// synchronous: when it returns, the click's dispatch (flip, persist,
// apply) has completed daemon-side.
func eventClick(ctx context.Context, id int32) error {
	_, err := runCmd(ctx, "gdbus", "call", "--session",
		"--dest", menuBusName,
		"--object-path", menuObjPath,
		"--method", menuEventMethod,
		strconv.FormatInt(int64(id), 10), "clicked", "<0>", "0")
	if err != nil {
		return fmt.Errorf("menu Event(%d): %w", id, err)
	}

	return nil
}
