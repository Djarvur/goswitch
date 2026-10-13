package main

// The phase-10 VTE live word case (plan 10-05, TERM-02): inside the VM stand
// (the vm-primary venue verdict) a fresh gnome-terminal window per row, the
// arming gate by the daemon key-trace/pty probe (never AT-SPI — Pitfall 7),
// then the Double Right Shift gesture and the pty-hex oracle — the terminal
// write branch must replace ghbdtn with привет in the captured bytes. Four
// rows: word-bash, word-zsh (pane program zsh -f, same capture protocol),
// mid-line (owner verdict OQ 2 «contract»: the word typed straight at the
// cursor behind a prefix separator) and clip-frozen (the CLIPBOARD sentinel
// byte-untouched by the terminal correction — ADR-008 D4/D5). Every row
// lands a dated verdict row in the phase journal with the venue marker; a
// failure lands the honest-abort row and fails the case (exit contract) —
// never a retry loop.

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// The case's journal section and frozen grep marks: the section header text
// is the contract (the journal owns it; TestSpikeJournalSectionsExist pins
// it), the level-2 mark pins the TERMINAL write branch — the DEBUG
// correction record's level attribute, the matrix's ExpectLevel form
// (Level2 is the chunked branch; the GTK pending machine logs Level1 and
// must never answer here).
const (
	secTermLive = "Фаза 10 живые ряды VTE (term-word-vte / term-negative-witnesses)"

	termLevel2Mark = `"msg":"correction","level":2`

	// The mid-line row's typed prefix and its hex form: the prefix rides the
	// same typing flow (contract-valid — the cursor never moves), and the
	// canonical capture must hold prefix+привет byte-adjacent after the
	// correction erases exactly the trailing word.
	termMidPrefix    = "echo "
	termMidPrefixHex = "6563686f20"

	// termCaptureWait bounds the canonical-line flush poll after the row's
	// line-ending keystroke (the committed text sits in the tty line buffer
	// until a newline delivers it to the capture program).
	termCaptureWait = 5 * time.Second

	// termRowTextBound caps one deferred-row error text (a full hex dump
	// would flood the journal cell).
	termRowTextBound = 200

	// armProbeKeys is the arming gate's key-record floor: the probe letter
	// plus its erase, press+release each (4 records above the baseline).
	armProbeKeys = 4

	// termArmGateTries bounds the arming ladder (SPEC §7.2 «ограниченные
	// ретраи», the wezterm family's form): the uinput churn of the first
	// ydotool calls can silently eat whole key calls (the 2026-10-13
	// fresh-boot finding), so the probe retries until the floor is met.
	termArmGateTries = 3
)

// termRow appends one dated verdict row to the case's journal section (the
// appendSpikeRow form; the label carries the row tag, the time the date).
func termRow(label, cmd, observed, verdict string) error {
	return appendSpikeRow(secTermLive, label+" "+time.Now().Format("15:04:05"), cmd, observed, verdict)
}

// boundRowText caps one deferred-row error text.
func boundRowText(text string) string {
	if len(text) > termRowTextBound {
		return text[:termRowTextBound]
	}

	return text
}

// venueMarker reports the execution venue for the journal rows (the plan's
// vm|host marker): systemd-detect-virt is the honest in-VM source —
// /tmp/stand-env.sh lives on the orchestrating host, never inside the VM
// the case runs in. The tool exits 1 exactly on "not virtualized", so the
// host answer rides the error contract.
func venueMarker(ctx context.Context) string {
	out, err := runCmd(ctx, "systemd-detect-virt")
	switch {
	case err == nil && out != "" && out != "none":
		return "vm (" + out + ")"
	case err == nil, strings.Contains(err.Error(), "exit status 1"):
		return "host"
	default:
		return "unknown"
	}
}

// startVTETerm spawns a fresh gnome-terminal window running the given pane
// program (--window: never a tab in an owner window). The pane command line
// embeds the fixture's unique tmp paths — the teardown kill marker.
//
//nolint:noctx // the window must outlive the case context: the caller closes it
func startVTETerm(pane []string) (*vteCapture, error) {
	argv := append([]string{"--window", "--"}, pane...)
	cmd := exec.Command(vteBin, argv...)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start %s pane %v: %w", vteBin, pane, err)
	}

	return &vteCapture{cmd: cmd}, nil
}

// killTermPane stops the fixture's pane processes by the unique tmp marker
// the pane command line carries (the pid-точечный discipline: the pattern
// never matches an owner window; the ctrl+d EOF chord is not traced on the
// stand — venue-mechanics — so the kill is the teardown fallback).
//
//nolint:noctx // best-effort teardown reap; the process calls outlive the case context harmlessly
func killTermPane(marker string) {
	out, err := exec.Command("pgrep", "-f", marker).Output()
	if err != nil {
		return
	}
	for _, pid := range strings.Fields(string(out)) {
		_ = exec.Command("kill", pid).Run()
	}
}

// termFixture is one open fixture round: the capture paths and the window
// handle (the pane's death closes the window; the client is a zombie reap).
type termFixture struct {
	cap       *vteCapture
	readyPath string
	ptyPath   string
}

// close tears the fixture down: the pane kill (the EOF chord is unreliable
// on the stand) then the client reap. Safe on the nil/error paths.
func (tf *termFixture) close() {
	if tf == nil {
		return
	}
	killTermPane(tf.readyPath)
	tf.cap.close()
}

// capturePane builds the pane program of a cat-capture fixture: the shell
// argv prefix ("bash -c" / "zsh -f -c") plus the readiness marker and the
// capture redirect — the startVTECapture protocol, pane shell parameterized.
func capturePane(shell ...string) func(ready, pty string) []string {
	return func(ready, pty string) []string {
		return append(shell, "echo ready > "+ready+"; cat > "+pty)
	}
}

// openTermFixture acquires one VTE round: the spawn, the pane-ready proof,
// and the daemon focus_in gate above the spawn baseline (the unified
// detector — map focus usually self-fires on the stand's stock kernel).
func (s *stand) openTermFixture(
	ctx context.Context, caseName, tag string, pane func(ready, pty string) []string,
) (*termFixture, error) {
	readyPath := filepath.Join(s.tmpDir, tag+"-ready")
	ptyPath := filepath.Join(s.tmpDir, tag+"-in")
	focusBase := s.focusBaseline()
	vcap, err := startVTETerm(pane(readyPath, ptyPath))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", caseName, err)
	}
	tf := &termFixture{cap: vcap, readyPath: readyPath, ptyPath: ptyPath}
	if werr := waitPaneReady(ctx, caseName, readyPath); werr != nil {
		tf.close()

		return nil, fmt.Errorf("%s: %w", caseName, werr)
	}
	fmt.Printf("%s: %s pane ready — awaiting daemon focus_in (map focus usually self-fires)\n", caseName, tag)
	if werr := s.waitFocusInFrom(ctx, caseName, tag+" gnome-terminal pane", focusBase); werr != nil {
		tf.close()

		return nil, fmt.Errorf("%s: %w", caseName, werr)
	}

	return tf, nil
}

// termArmGate is the rows' arming gate — BY THE DAEMON KEY-TRACE (the
// plan's gate, never an AT-SPI witness): one boundary-free probe letter
// typed and erased again (BackSpace is the venue-proven named key —
// venue-mechanics 2026-10-12; the named Enter forms are NOT resolved by
// this stand's ydotool and type their first letter instead, the run-3
// finding). The self-erasing pair leaves the tty line buffer empty, so no
// probe bytes ride into the row oracles. Window delivery itself is the
// venue-probe fact (r1+r2, full channel) plus each row's own pty-hex
// oracle — a broken delivery fails the row honestly.
func (s *stand) termArmGate(ctx context.Context, caseName string) error {
	for attempt := 1; attempt <= termArmGateTries; attempt++ {
		base := s.countSub(keyTraceMark)
		if err := s.injectText(ctx, "x"); err != nil {
			return fmt.Errorf("%s: arming probe type: %w", caseName, err)
		}
		if err := s.injectKeys(ctx, "BackSpace"); err != nil {
			return fmt.Errorf("%s: arming probe erase: %w", caseName, err)
		}
		if err := s.waitForNew(ctx, keyTraceMark, base+armProbeKeys, keyWait); err == nil {
			return nil
		}
		fmt.Printf("%s: arming attempt %d/%d below the key-trace floor — retrying\n",
			caseName, attempt, termArmGateTries)
	}

	return fmt.Errorf("%s: arming gate: key-trace floor unmet after %d attempts", caseName, termArmGateTries)
}

// termObserverRound warms the daemon's a11y observer path on a throwaway
// VTE window BEFORE the row fixtures map: the enabled-layer daemon starts
// the observer with its config (ensureAppid at boot), and a fixture that
// maps after it feeds the (sender, path) pair the live role query addresses
// — the 07-06 observer-start discipline, stand gates only (focus_in +
// key-trace, never an AT-SPI witness).
func (s *stand) termObserverRound(ctx context.Context, caseName string) error {
	tf, err := s.openTermFixture(ctx, caseName, "obs", capturePane("bash", "-c"))
	if err != nil {
		return err
	}
	defer tf.close()

	return s.termArmGate(ctx, caseName)
}

// termWordRound drives one manual word-correction round into the focused
// fixture: the typed text (gated by the key-trace), the Double Right Shift
// gesture, and the three daemon marks — the double decision, the executed
// correction (done) and the TERMINAL branch's level-2 DEBUG record. The
// final newline flushes the tty line buffer into the capture (venue
// mechanics: Enter rides ydotool type, never a named key).
func (s *stand) termWordRound(ctx context.Context, caseName, text string) error {
	keyBase := s.countSub(keyTraceMark)
	if err := s.injectText(ctx, text); err != nil {
		return fmt.Errorf("%s: type %q: %w", caseName, text, err)
	}
	if err := s.waitForNew(ctx, keyTraceMark, keyBase+minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("%s: key visibility: %w", caseName, err)
	}
	actionBase := s.countSub(actionDoubleMark)
	doneBase := s.countSub(correctionDoneMa)
	levelBase := s.countSub(termLevel2Mark)
	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return fmt.Errorf("%s: double gesture: %w", caseName, err)
	}
	if err := s.waitForNew(ctx, actionDoubleMark, actionBase+1, decisionWait); err != nil {
		return fmt.Errorf("%s: double-tap decision: %w", caseName, err)
	}
	if err := s.waitForNew(ctx, correctionDoneMa, doneBase+1, correctionWait); err != nil {
		return fmt.Errorf("%s: correction done: %w", caseName, err)
	}
	if err := s.waitForNew(ctx, termLevel2Mark, levelBase+1, correctionWait); err != nil {
		return fmt.Errorf("%s: terminal branch level-2 mark: %w", caseName, err)
	}

	return s.injectText(ctx, "\n")
}

// waitForCaptureHex polls the capture until the arming-stripped hex carries
// want — the pty oracle's settle form. The last hex rides the error for the
// journal row's observables.
//
//nolint:unparam // the timeout is the frozen termCaptureWait by design; the parameter keeps the helper general
func waitForCaptureHex(ctx context.Context, path, want string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	var hexr string
	for {
		if data, err := os.ReadFile(path); err == nil {
			hexr = hex.EncodeToString(stripArmingProbe(data))
			if strings.Contains(hexr, want) {
				return hexr, nil
			}
		}
		if time.Now().After(deadline) {
			return hexr, fmt.Errorf("capture %s never contained hex %s within %v (last hex %s)",
				path, want, timeout, hexr)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return hexr, err
		}
	}
}

// termWordRowCore is the shared word-row body: fixture, arming gate, the
// typed-text round and the pty-hex oracle. The row lands pass or FAIL with
// the hex; a FAIL also fails the case (the exit contract).
func (s *stand) termWordRowCore(
	ctx context.Context, caseName, rowTag string, pane func(ready, pty string) []string, text, wantHex string,
) error {
	tf, err := s.openTermFixture(ctx, caseName, rowTag, pane)
	if err != nil {
		return err
	}
	defer tf.close()
	if err := s.termArmGate(ctx, caseName); err != nil {
		return err
	}
	if err := s.termWordRound(ctx, caseName, text); err != nil {
		return err
	}
	hexr, cerr := waitForCaptureHex(ctx, tf.ptyPath, wantHex, termCaptureWait)
	observed := fmt.Sprintf("venue: %s; pty_hex=%s", venueMarker(ctx), hexr)
	cmd := "armGate + type " + text + " + double Shift_R + pty-hex oracle"
	if cerr == nil {
		return termRow(rowTag, cmd, observed,
			"Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона)")
	}
	if rerr := termRow(rowTag, cmd, observed,
		"Verdict: FAIL (pty-hex не содержит ожидаемые байты)"); rerr != nil {
		return rerr
	}

	return cerr
}

// termRowClipFrozen is the frozen CLIPBOARD row (criterion 3): a sentinel of
// the GS-SPIKE class rides CLIPBOARD through the whole correction round and
// wl-paste must read it back byte-identical — the terminal write branch
// never touches CLIPBOARD (noRung verify-after, ADR-008 D4/D5). The
// sentinel overwrite of the pre-state is documented here in the row (D5).
func (s *stand) termRowClipFrozen(ctx context.Context, caseName string) error {
	const rowTag = "clip-frozen"
	tf, err := s.openTermFixture(ctx, caseName, rowTag, capturePane("bash", "-c"))
	if err != nil {
		return err
	}
	defer tf.close()
	if err := s.termArmGate(ctx, caseName); err != nil {
		return err
	}
	if err := clipCopy(ctx, primarySentinelMark); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	before := clipReadNoNewline(ctx)
	if before != primarySentinelMark {
		return fmt.Errorf("%s: CLIPBOARD pre-state %q, want sentinel %q", caseName, before, primarySentinelMark)
	}
	if err := s.termWordRound(ctx, caseName, corpusLatin); err != nil {
		return err
	}
	hexr, cerr := waitForCaptureHex(ctx, tf.ptyPath, corpusHexRU, termCaptureWait)
	after := clipReadNoNewline(ctx)
	observed := fmt.Sprintf("venue: %s; clip_before=%s; clip_after=%s; pty_hex=%s"+
		" (CLIPBOARD перезаписан сентинелом — синтетика, пред-стейт задокументирован затиранием)",
		venueMarker(ctx), before, after, hexr)
	cmd := "printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline"
	if cerr == nil && after == primarySentinelMark {
		return termRow(rowTag, cmd, observed,
			"Verdict: pass (CLIPBOARD байт-нетронут терминальной коррекцией, слово исправлено)")
	}
	if rerr := termRow(rowTag, cmd, observed, "Verdict: FAIL (клипборд изменён или слово не исправлено)"); rerr != nil {
		return rerr
	}
	if cerr != nil {
		return cerr
	}

	return fmt.Errorf("%s: CLIPBOARD changed: %q, want sentinel %q", caseName, after, primarySentinelMark)
}

// clipCopy sets a known CLIPBOARD pre-state (the primarySentinel form,
// CLIPBOARD target).
//
//nolint:noctx // best-effort clipboard write; bounded by the caller's step
func clipCopy(ctx context.Context, mark string) error {
	cmd := exec.Command("bash", "-c", "printf %s "+mark+" | wl-copy")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("set CLIPBOARD sentinel: %w", err)
	}

	return sleepCtx(ctx, chordSettle)
}

// clipReadNoNewline reads CLIPBOARD without the trailing newline; an empty
// or failed read is a defined state, not an error (Pitfall 5).
func clipReadNoNewline(ctx context.Context) string {
	out, err := runCmd(ctx, "wl-paste", "--no-newline")
	if err != nil {
		return ""
	}

	return out
}

// runTermWordVTE drives the phase-10 VTE word rows on the stand: the
// enabled-layer daemon (the observer path the terminal classification
// reads), the observer-start round, then the four rows — bash, zsh,
// mid-line (contract semantics) and the frozen CLIPBOARD sentinel. A
// failure lands the honest-abort row and fails the case (exit contract).
func runTermWordVTE(ctx context.Context, s *stand) error {
	const caseName = "term-word-vte"
	err := runTermWordRounds(ctx, s)
	if err == nil {
		fmt.Printf("%s: bash/zsh words, mid-line and clip-frozen rows green (rows in %s)\n", caseName, spikeJournal)

		return nil
	}
	obs := "case aborted: " + boundRowText(err.Error())
	if rerr := termRow(caseName, "VTE word round", obs, "Verdict: aborted"); rerr != nil {
		fmt.Fprintf(os.Stderr, "e2e: %s deferred journal row: %v\n", caseName, rerr)
	}

	return err
}

// runTermWordRounds is the row program of the word case (extracted so the
// honest-abort wrapper above stays under the complexity budget).
func runTermWordRounds(ctx context.Context, s *stand) error {
	const caseName = "term-word-vte"
	if err := startAutocorrectConfigDaemon(ctx, s, autocorrectConfigDoc); err != nil {
		return err
	}
	if err := s.termObserverRound(ctx, caseName); err != nil {
		return err
	}
	if err := s.termWordRowCore(ctx, caseName, "word-bash",
		capturePane("bash", "-c"), corpusLatin, corpusHexRU); err != nil {
		return err
	}
	if err := s.termWordRowCore(ctx, caseName, "word-zsh",
		capturePane("zsh", "-f", "-c"), corpusLatin, corpusHexRU); err != nil {
		return err
	}
	if err := s.termWordRowCore(ctx, caseName, "mid-line",
		capturePane("bash", "-c"), termMidPrefix+corpusLatin, termMidPrefixHex+corpusHexRU); err != nil {
		return err
	}

	return s.termRowClipFrozen(ctx, caseName)
}
