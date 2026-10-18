package main

// The host self-drive driver (the round-5 instrument, the 2026-10-17
// wezterm double-delete hunt): it drives a test wezterm PANE THE OWNER
// OPENED MANUALLY on the live host — no stand daemon, no gsettings writes,
// no windows spawned. The owner types ONE command in the dedicated pane
// (script -q /tmp/gs-pane.log); the driver then:
//
//  1. proves the pane's liveness and focus with a MARKER before every
//     injection burst: one harmless DIGIT via the in-repo uinput injector
//     (digits are identical keystrokes in EN and RU — the 2026-10-18 first
//     battery's "." landed as «ю» in the RU-mode pane and the check grepped
//     a ghost), a 300 ms wait, and a check that the script log's new bytes
//     received it — a missing marker is a loud SKIP row and an aborted
//     battery, the real sequence is never typed blind (the marker
//     discipline);
//  2. normalizes the pane's IBus MODE at battery start: the corpus
//     (ghbdtn) only corrects from EN — the first live battery typed it
//     blind into an RU pane and got Cyrillic garbage. The daemon's mode
//     rides goswitchctl status; a non-en mode gets the double Right Shift
//     chord on an EMPTY buffer (pure flip), a re-read, up to twice — a
//     mode that never reads en aborts the battery (no typing in an
//     unverified layout), every step a content-free row;
//  3. runs the battery: one probe run, N baseline runs, N timing-axis
//     runs (a typing→gesture wait before the chord), each run a marker →
//     Ctrl-U line clear (the marker AND any stray line junk gone — the
//     corpus always types on a clean line) → ghbdtn + the double Right
//     Shift correction gesture, parsed from the NEW log bytes for the
//     choreography (erase→insert, and the second delete that is the
//     client-replay defect);
//  4. observes (never induces) focus churn: the daemon journal's
//     lifecycle rows in the run window ride the verdict row;
//  5. appends a dated verdict row per run to stdout and the verdicts log.
//
// The rows carry counts and booleans only — the pane holds the test word
// alone (the marker discipline guarantees focus), and the daemon rows are
// content-free (D-20/D-21). The driver never injects anything at build,
// import or test time: live use is orchestrator-invoked (mise task
// test-selfdrive runs the DRY-RUN only; the -selfdrive LIVE form is a
// deliberate operator act).

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Djarvur/goswitch/internal/uinject"
)

// The driver's vocabulary: markers, paths, budgets. Every value is a
// behavioral constant of the battery, named once.
const (
	// ssMarker is the layout-independent focus proof: a DIGIT, identical
	// keystrokes in EN and RU, so the probe cannot produce Cyrillic garbage
	// when the pane's IBus mode differs from the battery's assumption.
	ssMarker = "7"
	ssWordRU = "привет" // the correction corpus's RU half (the commit needle)
	ssUnit   = "goswitchd"

	// ssOwnerCmd is THE owner instruction: the one command to type in the
	// dedicated pane so the driver's log watcher has a pty transcript.
	ssOwnerCmd = "script -q /tmp/gs-pane.log"

	ssDefaultPaneLog    = "/tmp/gs-pane.log"
	ssDefaultOut        = "/tmp/gs-selfdrive-verdicts.log"
	ssDefaultSettle     = 3 * time.Second
	ssDefaultRuns       = 3
	ssDefaultTimingWait = 5 * time.Second

	// ssMarkerWait is the marker's echo budget: the pane must prove the
	// focus within it, or the battery aborts (never types blind).
	ssMarkerWait = 1500 * time.Millisecond
	// ssBetweenRuns lets the previous run's deferred flip and verify window
	// settle before the next marker.
	ssBetweenRuns = 2 * time.Second
	// ssLiveCountdown is the abort window before the FIRST injection.
	ssLiveCountdown = 3 * time.Second
	// ssJournalLayout is systemd's --since/--timestamp form.
	ssJournalLayout = "2006-01-02 15:04:05"

	// ssOutFilePerm is the verdicts log's mode (owner-only: /tmp).
	ssOutFilePerm = 0o600

	// ssSyntheticClear is the dry-run vectors' Ctrl-U clear echo length: a
	// NON-trivial erase run, so the executable self-check pins the lenient
	// clear skip (marker + junk), not just the single-erase shape.
	ssSyntheticClear = 3
)

// The parser's byte vocabulary of a `script` transcript.
const (
	ssByteBS    = 0x08 // the cooked ERASE echo opens with backspace
	ssByteSpace = 0x20 // the "\b \b" triple's middle byte
	ssByteDEL   = 0x7f // the RAW capture's erase byte
	ssByteESC   = 0x1b // a terminal sequence's opener — neutral
	ssByteCSI   = '['  // the CSI introducer after ESC

	// ssBSEchoSpan is the "\b \b" triple's length: ONE erased character.
	ssBSEchoSpan = 3
	// ssESCSeqMin is the shortest terminal sequence (ESC + final).
	ssESCSeqMin = 2
	// ssCSIEndFirst/ssCSIEndLast bound the CSI final-byte range (@–~).
	ssCSIEndFirst = 0x40
	ssCSIEndLast  = 0x7e
)

// ssEvKind is the kind of one parsed pane-log event.
type ssEvKind int

const (
	ssEvText  ssEvKind = iota // printable output (typed echo, committed text)
	ssEvErase                 // erased characters (the forwarded BackSpaces)
)

// ssEvent is one merged run of the pane log: consecutive same-kind output
// is a single event (six erase echoes read E6, never six events).
type ssEvent struct {
	kind  ssEvKind
	runes int
}

// ssChoreo is the classified shape of one run window.
type ssChoreo int

const (
	ssChoreoEmpty      ssChoreo = iota // nothing arrived (the pane died silently)
	ssChoreoNoGesture                  // the word echoed, no correction fired
	ssChoreoEraseOnly                  // forwards landed, no commit within the settle
	ssChoreoClean                      // erase → insert, and nothing after
	ssChoreoReplay                     // a SECOND delete after the insert — the defect
	ssChoreoUnexpected                 // a foreign shape — the events string says what
)

// ssVerdictEmpty is the empty-window outcome's word (goconst).
const ssVerdictEmpty = "empty"

// word renders the outcome for the verdict row.
func (c ssChoreo) word() string {
	switch c {
	case ssChoreoEmpty:
		return ssVerdictEmpty
	case ssChoreoNoGesture:
		return "no-correction"
	case ssChoreoEraseOnly:
		return "erase-only"
	case ssChoreoClean:
		return "clean"
	case ssChoreoReplay:
		return "replay"
	case ssChoreoUnexpected:
		return "unexpected"
	default:
		return "unexpected"
	}
}

// ssRunReport is one run window's verdict payload: the classification, the
// second-delete flag (the defect itself) and the symbolic event string —
// the juxtaposition input against the daemon journal's emission-seq rows.
type ssRunReport struct {
	choreo    ssChoreo
	secondDel bool
	events    string
}

// ssSymbolic renders events as "T1 E1 T6 E6" (T=text runes, E=erases).
func ssSymbolic(events []ssEvent) string {
	parts := make([]string, 0, len(events))
	for _, e := range events {
		k := "T"
		if e.kind == ssEvErase {
			k = "E"
		}
		parts = append(parts, k+strconv.Itoa(e.runes))
	}

	return strings.Join(parts, " ")
}

// parsePaneEvents parses a raw `script -q` fragment into merged erase/text
// runs. The erase vocabulary: the cooked "\b \b" triple, a bare backspace,
// the raw capture's 0x7f — each ONE erased character. Terminal escape
// sequences, CR/LF and the bell are neutral: they never become events and
// never split a text run into phantom choreography. Stray non-UTF-8 bytes
// are neutral too.
func parsePaneEvents(frag []byte) []ssEvent {
	var events []ssEvent
	pending := 0 // runes of the current text run
	flushText := func() {
		if pending > 0 {
			events = append(events, ssEvent{kind: ssEvText, runes: pending})
			pending = 0
		}
	}
	appendErase := func() {
		flushText()
		if n := len(events); n > 0 && events[n-1].kind == ssEvErase {
			events[n-1].runes++

			return
		}
		events = append(events, ssEvent{kind: ssEvErase, runes: 1})
	}
	for i := 0; i < len(frag); {
		switch b := frag[i]; {
		case b == ssByteBS:
			// The cooked erase echo is the "\b \b" triple — ONE erase;
			// consume the triple, count one.
			if i+ssBSEchoSpan <= len(frag) && frag[i+1] == ssByteSpace && frag[i+2] == ssByteBS {
				i += ssBSEchoSpan - 1
			}
			appendErase()
			i++
		case b == ssByteDEL:
			appendErase()
			i++
		case b == ssByteESC:
			i += ssEscapeLen(frag[i:])
		case b < ssByteSpace:
			i++ // \r, \n, bell — neutral, never a choreography event
		default:
			r, size := utf8.DecodeRune(frag[i:])
			if r == utf8.RuneError && size <= 1 {
				i++

				continue
			}
			pending++
			i += size
		}
	}
	flushText()

	return events
}

// ssEscapeLen returns the length of the terminal sequence at the head of
// seq (the ESC opcode is neutral, whatever it encodes).
func ssEscapeLen(seq []byte) int {
	if len(seq) < ssESCSeqMin {
		return len(seq)
	}
	if seq[1] == ssByteCSI {
		for j := ssESCSeqMin; j < len(seq); j++ {
			if seq[j] >= ssCSIEndFirst && seq[j] <= ssCSIEndLast {
				return j + 1
			}
		}

		return len(seq)
	}

	return ssESCSeqMin
}

// analyzeRunWindow classifies one run window's events (the window opens at
// the marker's echo, taken from the log offset BEFORE the marker went out).
// The expected shape after the lenient marker skip: the word echo (T6),
// the correction's erase run (E6), the committed word (T6) — and, on the
// defect, a SECOND erase run. Anything else is named honestly.
func analyzeRunWindow(events []ssEvent, wordRunes int) ssRunReport {
	rep := ssRunReport{events: ssSymbolic(events)}
	i := ssSkipMarker(events, utf8.RuneCountInString(ssMarker))
	if i >= len(events) {
		rep.choreo = ssChoreoEmpty

		return rep
	}
	if events[i].kind != ssEvText || events[i].runes != wordRunes {
		rep.choreo = ssChoreoUnexpected

		return rep
	}

	return ssClassifyChoreo(rep, events, i+1, wordRunes)
}

// ssSkipMarker leniently skips the window's marker prefix: the marker's
// echo and the Ctrl-U line clear that follows it — ONE erase run of ANY
// length, because the clear wipes the marker plus whatever stray junk
// shared the line (the first battery's single marker BackSpace is gone).
// A prompt redraw before the echo is neutral bytes the parser already
// dropped; an absent marker skip leaves the classification to the
// choreography zone (honest, never a crash).
func ssSkipMarker(events []ssEvent, markerRunes int) int {
	i := 0
	if i < len(events) && events[i].kind == ssEvText && events[i].runes == markerRunes {
		i++
	}
	if i < len(events) && events[i].kind == ssEvErase && events[i].runes >= 1 {
		i++
	}

	return i
}

// ssClassifyChoreo walks the choreography zone after the word echo: erase
// (the correction), the committed word — and anything beyond it, where a
// second erase is THE defect (the client replayed its queued forwards).
func ssClassifyChoreo(rep ssRunReport, events []ssEvent, i, wordRunes int) ssRunReport {
	if i >= len(events) {
		rep.choreo = ssChoreoNoGesture

		return rep
	}
	if events[i].kind != ssEvErase {
		rep.choreo = ssChoreoUnexpected

		return rep
	}
	i++ // the correction's erase run
	if i >= len(events) {
		rep.choreo = ssChoreoEraseOnly

		return rep
	}
	if events[i].kind != ssEvText || events[i].runes != wordRunes {
		rep.choreo = ssChoreoUnexpected

		return rep
	}
	i++ // the committed word
	if i >= len(events) {
		rep.choreo = ssChoreoClean

		return rep
	}
	for ; i < len(events); i++ {
		if events[i].kind == ssEvErase {
			rep.secondDel = true
			rep.choreo = ssChoreoReplay

			return rep
		}
	}
	rep.choreo = ssChoreoUnexpected // text after the commit without an erase

	return rep
}

// ssDriver is the live battery's state (the config rides verbatim).
type ssDriver struct {
	cfg config
}

// runSelfDrive is the -selfdrive entry point: preconditions (fail-closed),
// then the dry-run self-check or the LIVE battery. Returns the exit code.
func runSelfDrive(ctx context.Context, cfg config) int {
	d := &ssDriver{cfg: cfg}
	fmt.Printf("=== selfdrive: host self-drive driver (dry-run=%v) ===\n", cfg.ssDry)
	fmt.Printf("pane log: %s\nverdicts: %s\nsettle: %v; baseline=%d; timing=%d (typing→gesture wait %v); pacing=%dms\n",
		cfg.ssPaneLog, cfg.ssOut, cfg.ssSettle, cfg.ssBaseline, cfg.ssTiming, cfg.ssTimingWait, cfg.pacing)
	if err := d.preconditions(ctx); err != nil {
		if !cfg.ssDry {
			fmt.Fprintf(os.Stderr, "selfdrive: FAIL precondition: %v\n", err)

			return 1
		}
		fmt.Printf("selfdrive: precondition (dry-run, report-only): %v\n", err)
	}
	if cfg.ssDry {
		return d.selfCheck()
	}

	return d.battery(ctx)
}

// preconditions fail-closed: the uinput control node exists (the stand
// user must be in the input group; the injector itself opens only at
// battery start — never in the dry-run), the daemon journal is
// reachable, the mode reader (goswitchctl) is resolvable, and the pane log
// the owner's `script` command opened exists — without it the driver has
// nothing to watch and says so with the exact owner instruction.
func (d *ssDriver) preconditions(ctx context.Context) error {
	if _, err := os.Stat(uinject.DevicePath); err != nil {
		return fmt.Errorf("%s is unavailable: %w (the stand user must be in the input group)", uinject.DevicePath, err)
	}
	if _, err := runCmd(ctx, "journalctl", "--user", "-u", ssUnit, "-n", "1", "--no-pager"); err != nil {
		return fmt.Errorf("the %s user unit is unreachable via journalctl: %w", ssUnit, err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}
	if _, err := ssCtlPath(home); err != nil {
		return fmt.Errorf("the daemon-mode reader is unresolvable: %w", err)
	}
	if _, err := os.Stat(d.cfg.ssPaneLog); err != nil {
		return fmt.Errorf("the pane log %s does not exist — the owner must FIRST type in the dedicated pane: %s",
			d.cfg.ssPaneLog, ssOwnerCmd)
	}

	return nil
}

// selfCheck is the dry-run body: the parser's built-in vectors must
// classify exactly — no injections, no writes, exit non-zero on a
// mismatch (the mise test-selfdrive task runs THIS).
func (d *ssDriver) selfCheck() int {
	words := utf8.RuneCountInString(corpusLatin)
	clearEcho, echoN := ssEchoFrag(ssSyntheticClear), ssEchoFrag(words) // the Ctrl-U echo (marker + junk erased)
	vectors := []struct {
		name string
		frag []byte
		want ssChoreo
	}{
		{"clean", ssFrag(ssMarker, clearEcho, corpusLatin, echoN, ssWordRU), ssChoreoClean},
		{"replay", ssFrag(ssMarker, clearEcho, corpusLatin, echoN, ssWordRU, echoN), ssChoreoReplay},
		{"no-correction", ssFrag(ssMarker, clearEcho, corpusLatin), ssChoreoNoGesture},
	}
	bad := 0
	for _, v := range vectors {
		got := analyzeRunWindow(parsePaneEvents(v.frag), words)
		verdict := "ok"
		if got.choreo != v.want {
			verdict = "MISMATCH"
			bad++
		}
		fmt.Printf("selfdrive: vector %-14s choreo=%-14s want=%-14s %s\n",
			v.name, got.choreo.word(), v.want.word(), verdict)
	}
	if bad > 0 {
		fmt.Fprintf(os.Stderr, "selfdrive: %d parser vector(s) MISMATCHED\n", bad)

		return 1
	}
	fmt.Printf("selfdrive: dry-run green — parser vectors pass, preconditions reported; NO injections performed\n")

	return 0
}

// battery is the LIVE body: the probe run, the baseline axis, the timing
// axis. A marker failure aborts the whole battery — the one unforgivable
// outcome is typing the real sequence at a window that is not the pane.
func (d *ssDriver) battery(ctx context.Context) int {
	fmt.Printf("selfdrive: LIVE — the uinput injector types into the FOCUSED window; the dedicated pane (%s) "+
		"must hold focus. Ctrl-C aborts. Starting in %v...\n", d.cfg.ssPaneLog, ssLiveCountdown)
	if err := sleepCtx(ctx, ssLiveCountdown); err != nil {
		return 1
	}
	// The layout gate BEFORE any battery keystroke: typing blind in an
	// unverified layout is the one outcome the marker discipline cannot
	// catch (the marker itself lands wrong), so the mode normalizes here.
	if err := d.normalizeMode(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "selfdrive: BATTERY ABORTED (no further injections): %v\n", err)

		return 1
	}
	// The injector opens AFTER the layout gate: an unavailable or unwritable
	// uinput node is a loud abort row BEFORE any typing (fail-closed, never
	// a partial battery).
	inj, err := uinject.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "selfdrive: BATTERY ABORTED (no further injections): uinject: %v\n", err)

		return 1
	}
	defer func() {
		if cerr := inj.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "selfdrive: uinject close: %v\n", cerr)
		}
	}()
	axes := []struct {
		name    string
		runs    int
		preWait time.Duration
	}{
		{name: "probe", runs: 1},
		{name: "baseline", runs: d.cfg.ssBaseline},
		{name: "timing", runs: d.cfg.ssTiming, preWait: d.cfg.ssTimingWait},
	}
	for _, ax := range axes {
		for i := 1; i <= ax.runs; i++ {
			if err := d.runRow(ctx, inj, ax.name, i, ax.runs, ax.preWait); err != nil {
				fmt.Fprintf(os.Stderr, "selfdrive: BATTERY ABORTED (no further injections): %v\n", err)

				return 1
			}
			if err := sleepCtx(ctx, ssBetweenRuns); err != nil {
				return 1
			}
		}
	}
	fmt.Printf("selfdrive: battery complete — verdict rows above and in %s\n", d.cfg.ssOut)

	return 0
}

// runRow drives one battery run: the marker discipline, the Ctrl-U line
// clear, the word, the (optional) timing wait, the gesture, the settle,
// and the verdict row parsed from the log bytes the run itself produced.
func (d *ssDriver) runRow(ctx context.Context, inj *uinject.Injector,
	axis string, i, of int, preWait time.Duration,
) error {
	tag := "axis=" + axis + " run=" + strconv.Itoa(i) + "/" + strconv.Itoa(of)
	offset := ssFileSize(d.cfg.ssPaneLog)
	from := time.Now()
	if err := d.marker(ctx, inj, tag, offset); err != nil {
		return err
	}
	// The line clear (readline Ctrl-U): erases the marker AND any stray
	// owner or previous-run junk — the corpus always types on a clean line,
	// and the marker's own BackSpace step died with it (the clear already
	// wiped the marker).
	if err := ssCtrlU(ctx, inj, d.cfg.pacing); err != nil {
		return fmt.Errorf("%s: line clear: %w", tag, err)
	}
	if err := ssType(ctx, inj, d.cfg.pacing, corpusLatin); err != nil {
		return fmt.Errorf("%s: type %s: %w", tag, corpusLatin, err)
	}
	if err := sleepCtx(ctx, preWait); err != nil {
		return fmt.Errorf("%s: timing wait: %w", tag, err)
	}
	// The correction gesture: the double Right Shift chord, the same
	// window-paced keycodes the e2e corpus drives (injectKeys Shift_R ×2).
	if err := ssKeys(ctx, inj, d.cfg.pacing*tapWindowFactor, "ShiftRight", "ShiftRight"); err != nil {
		return fmt.Errorf("%s: double gesture: %w", tag, err)
	}
	if err := sleepCtx(ctx, d.cfg.ssSettle); err != nil {
		return fmt.Errorf("%s: settle: %w", tag, err)
	}
	frag, _, err := ssReadFrom(d.cfg.ssPaneLog, offset)
	if err != nil {
		return fmt.Errorf("%s: %w", tag, err)
	}
	rep := analyzeRunWindow(parsePaneEvents(frag), utf8.RuneCountInString(corpusLatin))
	pairs, note := ssFocusPairs(ctx, from, time.Now())
	row := tag +
		" choreo=" + rep.choreo.word() +
		" second-delete=" + strconv.FormatBool(rep.secondDel) +
		" events=" + strconv.Quote(rep.events) +
		" focus-pairs=" + strconv.Itoa(pairs)
	if note != "" {
		row += " note=" + strconv.Quote(note)
	}
	d.emit(row)

	return nil
}

// marker is the safety core: one harmless digit, a bounded wait, and
// the pane log's new bytes must receive it — else the battery aborts with
// the loud skip row BEFORE the real sequence (never typed blind).
func (d *ssDriver) marker(ctx context.Context, inj *uinject.Injector, tag string, offset int64) error {
	if err := ssType(ctx, inj, d.cfg.pacing, ssMarker); err != nil {
		return fmt.Errorf("%s: marker type: %w", tag, err)
	}
	if err := sleepCtx(ctx, ssMarkerWait); err != nil {
		return fmt.Errorf("%s: marker wait: %w", tag, err)
	}
	if ssTailContains(d.cfg.ssPaneLog, offset, ssMarker) {
		return nil
	}
	d.emit(tag + " marker=MISSING verdict=skip-abort focus=elsewhere")

	return fmt.Errorf("%s: the marker never reached the pane log — focus is NOT the dedicated pane; "+
		"battery aborted before the real sequence", tag)
}

// emit prints one dated verdict row to stdout and appends it to the
// verdicts log (an append failure degrades loudly, never silently).
func (d *ssDriver) emit(row string) {
	line := time.Now().Format(time.RFC3339) + " " + row
	fmt.Println("selfdrive: " + line)
	f, err := os.OpenFile(d.cfg.ssOut, os.O_APPEND|os.O_CREATE|os.O_WRONLY, ssOutFilePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "selfdrive: append %s: %v\n", d.cfg.ssOut, err)

		return
	}
	defer func() { _ = f.Close() }()
	if _, werr := fmt.Fprintln(f, line); werr != nil {
		fmt.Fprintf(os.Stderr, "selfdrive: write %s: %v\n", d.cfg.ssOut, werr)
	}
}

// ssFocusPairs counts the daemon journal's focus_in rows in the run window
// (observation only — focus churn is never induced by this driver). The
// honest note: without -debug the lifecycle rows are DEBUG-only and the
// count reads zero.
func ssFocusPairs(ctx context.Context, from, to time.Time) (pairs int, note string) {
	out, err := runCmd(ctx, "journalctl", "--user", "-u", ssUnit,
		"--since", from.Format(ssJournalLayout), "--until", to.Format(ssJournalLayout),
		"-o", "cat", "--no-pager")
	if err != nil {
		return 0, "journalctl unavailable: " + boundRowText(err.Error())
	}
	if !strings.Contains(out, `"msg":"lifecycle"`) {
		return 0, "no lifecycle rows at INFO (focus rows are DEBUG-only)"
	}

	return strings.Count(out, `"kind":"focus_in"`), ""
}

// ssType types text through the in-repo uinput injector with the run's
// pacing (the stand's injectText form, aimed at the FOCUSED window).
func ssType(ctx context.Context, inj *uinject.Injector, pacingMs int, text string) error {
	if err := inj.Type(ctx, text, pacingMs); err != nil {
		return fmt.Errorf("inject %q: %w", text, err)
	}

	return nil
}

// ssCtrlUSteps is the explicit Ctrl-U chord: the modifier down first, the
// letter in, the letter out, the modifier out last — a true simultaneous
// chord (the ydotool "ctrl+u" form it replaces), each step its own
// press/release with the driver's pacing between.
//
//nolint:gochecknoglobals // the chord's frozen step order, pinned by the corpus
var ssCtrlUSteps = []ssChordStep{
	{name: ssKeyLeftCtrl, down: true},
	{name: "u", down: true},
	{name: "u", down: false},
	{name: ssKeyLeftCtrl, down: false},
}

// ssKeyLeftCtrl is the chord modifier's name in the injector's table
// (named once — the clear and the corpus both say it).
const ssKeyLeftCtrl = "LeftCtrl"

// ssChordStep is one explicit chord step: which key, down or up.
type ssChordStep struct {
	name string
	down bool
}

// ssCtrlU clears the readline line: the explicit LeftCtrl+u chord, each
// step paced so the terminal sees a clean kill-line.
func ssCtrlU(ctx context.Context, inj *uinject.Injector, pacingMs int) error {
	for _, step := range ssCtrlUSteps {
		var err error
		if step.down {
			err = inj.Press(step.name)
		} else {
			err = inj.Release(step.name)
		}
		if err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}
		if err := sleepCtx(ctx, time.Duration(pacingMs)*time.Millisecond); err != nil {
			return fmt.Errorf("pace %s: %w", step.name, err)
		}
	}

	return nil
}

// ssKeys taps key names sequentially with the tap-window pacing (the
// stand's injectKeys form: pacing×3 so a multi-tap fits the 300 ms window).
func ssKeys(ctx context.Context, inj *uinject.Injector, pacingMs int, keys ...string) error {
	if err := inj.Chord(ctx, keys, pacingMs); err != nil {
		return fmt.Errorf("inject keys %v: %w", keys, err)
	}

	return nil
}

// ssEchoFrag builds the cooked erase echo of n erased characters (the
// synthetic-fragment and self-check form).
func ssEchoFrag(n int) string {
	return strings.Repeat("\b \b", n)
}

// ssFrag assembles a synthetic pane-log fragment from its pieces.
func ssFrag(parts ...string) []byte {
	return []byte(strings.Join(parts, ""))
}

// ssFileSize returns the pane log's current size (0 when absent).
func ssFileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}

	return fi.Size()
}

// ssTailContains reports whether the log bytes past offset received the
// needle — the marker's liveness proof.
func ssTailContains(path string, offset int64, needle string) bool {
	data, err := os.ReadFile(path)
	if err != nil || int64(len(data)) <= offset {
		return false
	}

	return strings.Contains(string(data[offset:]), needle)
}

// ssReadFrom returns the log's growth past offset and the new size; a
// SHRUNK log means the pane's script session restarted — an honest error,
// never a mis-attributed window.
func ssReadFrom(path string, offset int64) (data []byte, size int64, err error) {
	data, rerr := os.ReadFile(path)
	if rerr != nil {
		return nil, offset, fmt.Errorf("read pane log %s: %w", path, rerr)
	}
	if int64(len(data)) < offset {
		return nil, offset, fmt.Errorf("pane log %s shrank (%d < %d bytes) — the pane's script session restarted",
			path, len(data), offset)
	}

	return data[offset:], int64(len(data)), nil
}

// The battery's layout gate: the corpus (ghbdtn) only corrects from EN —
// the 2026-10-18 first live battery typed it blind into an RU-mode pane
// (Cyrillic garbage, no correction). The daemon's mode rides goswitchctl
// status, resolved the plans' way.
const (
	// ssModeEN is the daemon mode the battery normalizes to before typing.
	ssModeEN = "en"
	// ssModeAttempts is the flip budget: a failed flip is retried once,
	// then the battery aborts (never types blind in an unverified layout).
	ssModeAttempts = 2
	// ssModeSettle lets a flip reach the daemon before the re-read.
	ssModeSettle = 600 * time.Millisecond
)

// ssCtlPath resolves the goswitchctl binary the plans resolve it: on PATH
// first, else the newest-by-mtime binary among the installed releases.
func ssCtlPath(home string) (string, error) {
	if p, err := exec.LookPath("goswitchctl"); err == nil {
		return p, nil
	}

	return ssNewestReleaseCtl(home)
}

// ssNewestReleaseCtl picks the newest-by-mtime goswitchctl among
// ~/.local/share/goswitch/releases/*/goswitchctl (the plans'
// `ls -t .../releases/*/goswitchctl | head -1`); none is a loud error,
// never a guessed version.
func ssNewestReleaseCtl(home string) (string, error) {
	pattern := filepath.Join(home, ".local", "share", "goswitch", "releases", "*", "goswitchctl")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob %s: %w", pattern, err)
	}
	best, bestMod := "", time.Time{}
	for _, p := range paths {
		fi, statErr := os.Stat(p)
		if statErr != nil {
			continue
		}
		if best == "" || fi.ModTime().After(bestMod) {
			best, bestMod = p, fi.ModTime()
		}
	}
	if best == "" {
		return "", fmt.Errorf("goswitchctl is not on PATH and no release ships one under %s", pattern)
	}

	return best, nil
}

// ssParseMode extracts the daemon's mode token (mode=en|ru) from a
// goswitchctl status report; empty when the report carries no mode field
// (fail-closed: no guessed mode, the gate refuses to proceed).
func ssParseMode(status string) string {
	for _, field := range strings.Fields(status) {
		if token, ok := strings.CutPrefix(field, "mode="); ok {
			return token
		}
	}

	return ""
}

// ssDaemonMode reads the daemon's current mode through goswitchctl status.
func ssDaemonMode(ctx context.Context) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	ctl, err := ssCtlPath(home)
	if err != nil {
		return "", err
	}
	out, err := runCmd(ctx, ctl, "status")
	if err != nil {
		return "", fmt.Errorf("goswitchctl status: %w", err)
	}
	mode := ssParseMode(out)
	if mode == "" {
		return "", fmt.Errorf("goswitchctl status carries no mode field: %s", boundRowText(out))
	}

	return mode, nil
}

// normalizeMode is the battery's layout gate: the daemon's mode rides
// goswitchctl status; a non-en mode gets GNOME's own input-source switch
// (gsettings current → the xkb 'us' source — the daemon follows GNOME and
// NEVER flips on the chord inside a terminal: an empty-buffer double
// Shift_R there is action n:2 «correction skipped: empty-buffer», pinned
// live 2026-10-18 23:38), one settle, a re-read, up to ssModeAttempts
// times. A mode that never reads en aborts the battery before any
// typing, and every step rides a content-free row.
func (d *ssDriver) normalizeMode(ctx context.Context) error {
	mode, err := ssDaemonMode(ctx)
	if err != nil {
		return fmt.Errorf("mode read: %w", err)
	}
	for attempt := 1; mode != ssModeEN && attempt <= ssModeAttempts; attempt++ {
		d.emit("mode-normalization mode=" + mode + " attempt=" + strconv.Itoa(attempt) +
			" action=gsettings-switch (GNOME input-source → us, content-free)")
		if err := ssSwitchGNOMEToEN(ctx); err != nil {
			return fmt.Errorf("mode switch: %w", err)
		}
		if err := sleepCtx(ctx, ssModeSettle); err != nil {
			return fmt.Errorf("mode switch settle: %w", err)
		}
		if mode, err = ssDaemonMode(ctx); err != nil {
			return fmt.Errorf("mode re-read: %w", err)
		}
	}
	if mode != ssModeEN {
		d.emit("mode-normalization FAILED mode=" + mode + " verdict=abort (no typing in an unverified layout)")

		return fmt.Errorf("the daemon mode never read %s after %d switch attempt(s) (last read: %s) — "+
			"battery aborted before any typing", ssModeEN, ssModeAttempts, mode)
	}
	d.emit("mode-normalization mode=en verdict=proceed")

	return nil
}

// ssSwitchToEN switches the global IBus engine to the daemon's EN engine
// directly (`ibus engine goswitch-en`). Pinned live 2026-10-18 23:47: the
// gsettings `current` route does NOT move IBus on this stack (GNOME's
// current changed, the engine stayed goswitch-ru, the daemon saw nothing),
// while the direct IBus switch is instant and daemon-observable (engine
// created → mode to en → «mode corrected»).
func ssSwitchGNOMEToEN(ctx context.Context) error {
	if _, err := runCmd(ctx, "ibus", "engine", "goswitch-en"); err != nil {
		return fmt.Errorf("ibus engine goswitch-en: %w", err)
	}

	return nil
}
