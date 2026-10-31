package main

// The menu-blocklist live proof (plan 11-05, MENU-01): the «Добавить <app>
// в чёрный список» item driven over the SAME DBusMenu Event channel the
// menu-v2 case pins (no ydotool click, no mouse — the 07-RESEARCH Q1b
// channel proven live rc=0). The case closes the whole live loop on the
// VM stand: a fresh fixture maps AFTER the daemon's lazy observer is live
// (the macr-per-app discipline), the observed bridge identity arms the
// item (the GetGroupProperties poll reads the same snapshot the click's
// guard reads, S1 AT REQUEST TIME), ONE Event click on id 16 lands the
// regexp.QuoteMeta pattern into autocorrect.apps_blocklist (the 11-03
// writer round-trip — the byte footprint is EXACTLY one swapped line,
// proven live), the strict config.Load accepts the result, the enabled
// layer abstains app-blocked on the fixture's field (the blocklist-silent
// oracle machine of 07-06 reused verbatim), and the SECOND click on the
// now-covered item leaves the document BYTE-identical (D-01; the writer's
// dedup is the second belt, the file-level no-op is the pinned oracle).

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/Djarvur/goswitch/internal/appid"
	// The strict-load oracle rides the daemon's OWN Load semantics (S5) —
	// aliased: the stand's CLI config struct owns the plain name.
	gsconfig "github.com/Djarvur/goswitch/internal/config"
)

const (
	// menuIDBlocklistApp pins the item's id (internal/indicator/menu.go,
	// phase 11 — the next free id AFTER the canonical set; a renumber
	// breaks this case loudly, by design, the menu-v2 pin form).
	menuIDBlocklistApp = int32(16)

	// caseMenuBlocklist is the case's registry name — the error messages'
	// prefix (the caseMenuV2 form: one case, one name).
	caseMenuBlocklist = "menu-blocklist"

	// blocklistWait bounds every blocklist-step oracle (the arm poll, the
	// click → document settle); the apply is SYNCHRONOUS inside the click
	// (the 11-04 pin — Event returns after the dispatch's write+reload),
	// the budget only absorbs the poll's read skew.
	blocklistWait = 10 * time.Second
)

// blocklistMenuConfigDoc is the case's complete config document (the
// 03-02 no-overlay rule, the menuConfigDoc form): autocorrect ON with an
// EMPTY apps_blocklist — the silence oracle needs the enabled layer, and
// the click's append must be the document's FIRST blocklist entry (the
// «exactly one element» oracle needs the clean start). Everything else is
// at the case-standard values; the sound section is explicit.
const blocklistMenuConfigDoc = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: false # no corrections in this case — the abstention stays countable
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
autocorrect:
  enabled: true
  apps_blocklist: []
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
sound:
  enabled: false
`

// parseBlocklistDoc decodes the persisted document's autocorrect block
// (the parseMenuToggles partial-decoder form): the enabled scalar and the
// apps_blocklist list — everything else in the document is the daemon's
// business, not this case's.
func parseBlocklistDoc(data []byte) (enabled bool, list []string, err error) {
	var doc struct {
		Autocorrect struct {
			Enabled       bool     `yaml:"enabled"`
			AppsBlocklist []string `yaml:"apps_blocklist"`
		} `yaml:"autocorrect"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false, nil, fmt.Errorf("parse blocklist document: %w", err)
	}

	return doc.Autocorrect.Enabled, doc.Autocorrect.AppsBlocklist, nil
}

// waitBlocklistDoc polls the persisted document until the blocklist list
// reads exactly want — the file oracle of the click (the waitMenuToggles
// form, instance-exact through the pattern).
func (s *stand) waitBlocklistDoc(ctx context.Context, cfgPath string, want []string) error {
	deadline := time.Now().Add(blocklistWait)
	var last string
	for {
		data, err := os.ReadFile(cfgPath)
		if err == nil {
			_, list, perr := parseBlocklistDoc(data)
			switch {
			case perr != nil:
				last = "parse error: " + perr.Error()
			case slices.Equal(list, want):
				return nil
			default:
				last = fmt.Sprintf("%q", list)
			}
		} else {
			last = err.Error()
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s: persisted document did not settle to %q (last %s)",
				caseMenuBlocklist, want, last)
		}
		if err := sleepCtx(ctx, menuStatusPoll); err != nil {
			return err
		}
	}
}

// lineMultiset counts the document's lines — the input of the byte
// footprint oracle. Only the trailing newline's empty tail is dropped;
// interior empty lines count (a blank-line change must fail the oracle).
func lineMultiset(data []byte) map[string]int {
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	counts := make(map[string]int, len(lines))
	for _, line := range lines {
		counts[line]++
	}

	return counts
}

// assertOnlyBlocklistLine pins the click's byte footprint (T-11-05-01):
// before/after differ in EXACTLY one swapped line pair — the empty
// `apps_blocklist: []` line out, the one-element line carrying the
// QuoteMeta literal in — and every other line is byte-identical (the
// 11-03 Node round-trip discipline, proven live). The exact literals are
// deterministic here: the case writes the before-document itself and the
// writer's encoder pins the 2-space indent.
func assertOnlyBlocklistLine(before, after []byte, pattern string) error {
	bm, am := lineMultiset(before), lineMultiset(after)
	var added, removed []string
	for line, n := range am {
		for range n - bm[line] {
			added = append(added, line)
		}
	}
	for line, n := range bm {
		for range n - am[line] {
			removed = append(removed, line)
		}
	}

	wantAdded := "  apps_blocklist: [" + pattern + "]"
	if len(added) != 1 || added[0] != wantAdded {
		return fmt.Errorf("%s: byte footprint: added lines %q, want exactly [%q]",
			caseMenuBlocklist, added, wantAdded)
	}
	if len(removed) != 1 || removed[0] != "  apps_blocklist: []" {
		return fmt.Errorf("%s: byte footprint: removed lines %q, want exactly [\"  apps_blocklist: []\"]",
			caseMenuBlocklist, removed)
	}

	return nil
}

// waitBlocklistArmed polls the LIVE menu until the blocklist item answers
// with the OBSERVED identity in its label — the daemon-side arm gate: the
// freshness snapshot the click's guard reads (GetGroupProperties serves
// from the same snapshot AT REQUEST TIME, the S1 discipline). The unknown
// state's STATIC label never carries a name, so the identity substring is
// the precise arming predicate; covered is impossible here (the
// document's list is empty before the click).
func waitBlocklistArmed(ctx context.Context, app string) error {
	deadline := time.Now().Add(blocklistWait)
	var last string
	for {
		out, err := runCmd(ctx, "gdbus", "call", "--session",
			"--dest", menuBusName,
			"--object-path", menuObjPath,
			"--method", "com.canonical.dbusmenu.GetGroupProperties",
			"["+strconv.FormatInt(int64(menuIDBlocklistApp), 10)+"]", "[]")
		if err == nil {
			last = out
			if strings.Contains(out, app) {
				return nil
			}
		} else {
			last = err.Error()
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s: blocklist item never armed with %q (last %s)",
				caseMenuBlocklist, app, last)
		}
		if err := sleepCtx(ctx, menuStatusPoll); err != nil {
			return err
		}
	}
}

// startBlocklistDaemon writes the case's document and runs the -config
// spawn ladder (the startAutocorrectConfigDaemon form) — returning the
// temp config path, the file oracle's target (the case owns the document,
// never the user's real config.yaml).
func startBlocklistDaemon(ctx context.Context, s *stand, doc string) (string, error) {
	if err := s.activateGoswitch(ctx); err != nil {
		return "", err
	}
	cfgPath := filepath.Join(s.tmpDir, "menu-blocklist.yaml")
	if err := os.WriteFile(cfgPath, []byte(doc), configFilePerm); err != nil {
		return "", fmt.Errorf("%s: write temp config: %w", caseMenuBlocklist, err)
	}
	if err := s.restartDaemonWithArgs("-config", cfgPath); err != nil {
		return "", err
	}
	for _, mark := range []string{configLoadedMark, componentRegisteredMark, ctlListeningMark} {
		if err := s.waitForLog(ctx, mark, registrationWait); err != nil {
			return "", fmt.Errorf("%s daemon -config spawn (%s): %w", caseMenuBlocklist, mark, err)
		}
	}

	return cfgPath, s.activateGoswitchRetry(ctx)
}

// runMenuBlocklist drives the whole MENU-01 live loop over the Event
// channel: observe → arm → click → document/Load/silence oracles → the
// covered second click's byte-identical no-op (D-01). Each round in its
// own helper (the case body stays the step list).
func runMenuBlocklist(ctx context.Context, s *stand) error {
	cfgPath, err := startBlocklistDaemon(ctx, s, blocklistMenuConfigDoc)
	if err != nil {
		return err
	}
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}
	witness, app, err := blocklistObserveFixture(ctx, s)
	if err != nil {
		return err
	}
	defer s.reapPasswordFixture() // the observation surface lives through the silence round
	if err := blocklistClickRound(ctx, s, ctlBin, cfgPath, app); err != nil {
		return err
	}
	// The silence round — the blocklist-silent oracle machine of 07-06
	// verbatim: the typed word survives (witness count + verbatim readback
	// + the PLAIN stdout oracle), one app-blocked abstention, fired=0,
	// zero correction records.
	if err := s.acTypeAndWaitKeys(ctx); err != nil {
		return err
	}
	if err := s.acBlocklistOracle(ctx, ctlBin, caseMenuBlocklist, witness); err != nil {
		return err
	}
	if err := blocklistNoopRound(ctx, s, cfgPath, app); err != nil {
		return err
	}

	fmt.Printf("%s: click id %d landed the QuoteMeta pattern (one swapped line, strict Load), "+
		"the layer abstained app-blocked in the fixture, the covered second click was byte-identical — green\n",
		caseMenuBlocklist, menuIDBlocklistApp)

	return nil
}

// blocklistObserverRound is the observer-start round in the VM-stand form
// (the 10-08 machine-activation discipline): the zenity control spawns, the
// daemon focus_in gate PROVOKES the pair with the ydotool center-click
// nudge, and the single boundary-free keystroke starts the daemon's lazy
// a11y observer without moving any autocorrect counter. The host-era
// acStartObserverRound (waitZenityEntry) rides the 30-minute human-click
// manual-activation budget — a budget this unattended venue cannot spend
// (the 11-05 live finding: the first run hung the whole case watchdog on
// it); the proceed signal stays the daemon's own focus_in.
func (s *stand) blocklistObserverRound(ctx context.Context, caseName string) error {
	focusBase := s.focusBaseline()
	if err := s.startZenity(ctx); err != nil {
		return err
	}
	defer s.reapZenity()
	if err := s.waitFocusInFrom(ctx, caseName, "zenity observer round", focusBase); err != nil {
		return err
	}

	return s.injectText(ctx, "x")
}

// blocklistObserveFixture runs the observation prologue (the
// runAutocorrectBlocklistSilent shape): the observer-start round wakes the
// daemon's lazy a11y source (the 07-06 live finding — the daemon's
// per-key-event fold starts the source there), the fresh fixture's
// map-focus gain feeds BOTH observers (the case-side identity read and the
// daemon's freshness hook), and the arm poll leaves the item ACTIVE for
// the click. The started fixture is the caller's to reap (the case's
// deferred reap) — it must stay focused through the silence round.
func blocklistObserveFixture(ctx context.Context, s *stand) (witness, app string, err error) {
	if err := s.blocklistObserverRound(ctx, caseMenuBlocklist); err != nil {
		return "", "", err
	}
	obs, err := appid.Start(ctx)
	if err != nil {
		return "", "", fmt.Errorf("%s: start case observer: %w", caseMenuBlocklist, err)
	}
	if err := s.startPasswordFixture(ctx); err != nil {
		return "", "", err
	}
	if witness, err = s.waitFixtureWitness(ctx, acRoleTextWitness, 0, true); err != nil {
		return "", "", err
	}
	if app, err = waitFocusedApp(ctx, obs); err != nil {
		return "", "", fmt.Errorf("%s: observe fixture identity: %w", caseMenuBlocklist, err)
	}
	if app == "" {
		return "", "", fmt.Errorf("%s: no bridge identity observed for the fixture window", caseMenuBlocklist)
	}
	fmt.Printf("%s: observed fixture identity %q\n", caseMenuBlocklist, app)
	if err := waitBlocklistArmed(ctx, app); err != nil {
		return "", "", err
	}

	return witness, app, nil
}

// blocklistNoopRound is the D-01 no-op: the second click on the
// now-covered item must leave the document BYTE-identical (the Event is
// synchronous; the covered state was folded by the first click's reload;
// the writer's dedup is the second belt — the file-level equality is the
// pinned oracle).
func blocklistNoopRound(ctx context.Context, s *stand, cfgPath, app string) error {
	stable, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("%s: read document before the second click: %w", caseMenuBlocklist, err)
	}
	if err := eventClick(ctx, menuIDBlocklistApp); err != nil {
		return err
	}
	reclicked, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("%s: read document after the second click: %w", caseMenuBlocklist, err)
	}
	if !bytes.Equal(stable, reclicked) {
		return fmt.Errorf("%s: covered second click changed the document (D-01 violated): %d → %d bytes",
			caseMenuBlocklist, len(stable), len(reclicked))
	}

	return s.waitBlocklistDoc(ctx, cfgPath, []string{regexp.QuoteMeta(app)})
}

// blocklistClickRound is THE click with its full oracle belt: the byte
// snapshot before, the id-16 Event, the document settling to exactly the
// QuoteMeta pattern, the one-swapped-line byte footprint, the strict Load
// and the applied-and-still-enabled status. The Event call is
// synchronous: when it returns, the dispatch (QuoteMeta → writer → sync
// reload → fold) has completed daemon-side (the synchronous-apply pin of
// case_menu.go) — the poll budgets absorb only read skew.
func blocklistClickRound(ctx context.Context, s *stand, ctlBin, cfgPath, app string) error {
	before, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("%s: read document before the click: %w", caseMenuBlocklist, err)
	}
	if err := eventClick(ctx, menuIDBlocklistApp); err != nil {
		return err
	}
	pattern := regexp.QuoteMeta(app)
	if err := s.waitBlocklistDoc(ctx, cfgPath, []string{pattern}); err != nil {
		return err
	}
	after, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("%s: read document after the click: %w", caseMenuBlocklist, err)
	}
	if err := assertOnlyBlocklistLine(before, after, pattern); err != nil {
		return err
	}
	if _, err := gsconfig.Load(cfgPath); err != nil {
		return fmt.Errorf("%s: strict Load refused the appended document: %w", caseMenuBlocklist, err)
	}
	// The apply is synchronous AND the enabled scalar untouched: the same
	// click's reload folded the document (the silence round's counters
	// read the applied state), the layer stays ON.
	if err := menuStatusHas(ctx, ctlBin, "autocorrect_enabled=true"); err != nil {
		return fmt.Errorf("%s: applied status after the click: %w", caseMenuBlocklist, err)
	}

	return nil
}
