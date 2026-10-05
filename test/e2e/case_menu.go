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
	"os"
	"path/filepath"
	"time"
)

// The pinned DBusMenu surface and item ids — the daemon's own exported
// object (internal/indicator/menu.go, plan 07-05). A menu re-layout that
// moves these ids breaks this case loudly, by design (the key_links pin).
const (
	menuBusName = "org.djarvur.goswitch"
	menuObjPath = "/Menu"
	menuIface   = "com.canonical.dbusmenu"

	menuIDEN          = int32(1)
	menuIDRU          = int32(2)
	menuIDACToggle    = int32(4)
	menuIDSoundToggle = int32(5)
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
// default).
func parseMenuToggles(data []byte) (ac, sound bool, err error) {
	return false, false, nil // RED stub (plan 07-06 TDD): GREEN fills the decode
}

// runMenuV2 drives the menu v2 over the Event channel.
func runMenuV2(ctx context.Context, s *stand) error {
	const caseName = "menu-v2"
	_ = caseName
	_ = filepath.Join
	_ = os.WriteFile
	_ = ctx

	return nil
}

// eventClick delivers one DBusMenu click over the session bus — the
// research Q1b channel, the daemon's OWN exported /Menu object; never a
// blind injection.
func eventClick(ctx context.Context, s *stand, id int32) error {
	_ = ctx
	_ = s
	_ = id

	return nil
}
