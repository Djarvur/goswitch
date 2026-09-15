package main

// The YAML case-matrix runner (TEST-04, D-18): declarative "typing →
// expectation" cases on the live desktop, driven through the phase-1 stand
// primitives (injectText/injectKeys/pressKey/waitForLog, the 02-02 surface
// drivers). Case isolation is an explicit step of the run loop — every case
// gets the FULL stand cycle (setupStand → case → teardown), so the daemon's
// scriptMode and buffer never leak between cases and the file order cannot
// influence any outcome.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/godbus/dbus/v5"
	"gopkg.in/yaml.v3"

	"github.com/Djarvur/goswitch/engine"
)

// signalProbe is the POSIX liveness probe (signal 0 delivers nothing).
const signalProbe = syscall.Signal(0)

// The closed vocabularies of the case schema: anything outside these sets
// fails validation with the field named — a typo must never silently
// execute (T-02-06-01). Exposed as functions: package-level vars would be
// mutable globals under the strict lint.
const (
	matrixSurfaceZenity = "zenity"
	matrixSurfaceGTE    = "gnome-text-editor"
	surfaceChromiumName = "chromium"
)

// matrixLevelMax is the deepest ladder level (ADR-003); expect_level above
// it can never be observed and is rejected.
const matrixLevelMax = 2

func matrixSurfaces() []string {
	return []string{matrixSurfaceZenity, surfaceChromiumName, matrixSurfaceGTE}
}

func matrixModes() []string {
	return []string{"en", "ru"}
}

func matrixTaps() []string {
	return []string{"single", "double", "triple"}
}

// matrixKeyNames maps the schema's key names to the ydotool key names the
// physical path actually honors. Two of the four mappings are live-proven
// corrections: ydotool 0.1.8 resolves "space" to the physical S key and
// "Escape" to the physical E key (first-letter fallback, findings 02-03 and
// 02-05), so the separator rides the typing path and Escape goes by its
// lowercase table name.
func matrixKeyNames() map[string]string {
	return map[string]string{
		"Escape": "esc",
		"Enter":  "enter",
		"Tab":    "tab",
	}
}

// matrixSpaceKey rides the typing path instead of the key table (see
// matrixKeyNames).
const matrixSpaceKey = "space"

// goswitchComponentName mirrors engine/types.go's DefaultConfig literal
// (the IBus name every stand daemon registers under; not an exported
// engine constant).
const goswitchComponentName = "org.freedesktop.IBus.goswitch"

// matrixReportPath is the file artifact the run duplicates its report into
// (CI evidence, plan 02-07); gitignored.
const matrixReportPath = "e2e-report.txt"

// matrixStep is one step of a case: exactly one of the four fields is set
// (validated), the step kind it names decides the runner's action.
type matrixStep struct {
	Type  string `yaml:"type"`
	Key   string `yaml:"key"`
	Tap   string `yaml:"tap"`
	Focus string `yaml:"focus"`
}

// matrixCase is one "typing → expectation" case of the YAML matrix
// (02-RESEARCH Pattern 5 schema). ExpectLevel 0 means "do not pin the
// ladder level"; 1 and 2 pin the ACTUAL level the daemon chose (ADR-003).
type matrixCase struct {
	Name        string       `yaml:"name"`
	Surface     string       `yaml:"surface"`
	Mode        string       `yaml:"mode"`
	Steps       []matrixStep `yaml:"steps"`
	ExpectText  string       `yaml:"expect_text"`
	ExpectLevel uint8        `yaml:"expect_level"`
}

// loadMatrixCases decodes the multi-document YAML matrix strictly: a case
// with an unknown field or a missing mandatory element fails with the
// file's case number and name — never a silent skip (KnownFields, 02-06
// must-have).
func loadMatrixCases(data []byte) ([]matrixCase, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var cases []matrixCase
	for i := 1; ; i++ {
		var c matrixCase
		err := dec.Decode(&c)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode case #%d: %w", i, err)
		}
		if err := c.validate(); err != nil {
			return nil, fmt.Errorf("case #%d: %w", i, err)
		}
		cases = append(cases, c)
	}
	if len(cases) == 0 {
		return nil, errors.New("no cases in matrix file")
	}

	return cases, nil
}

// validate enforces the mandatory fields and the closed vocabularies.
func (c matrixCase) validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	surfaces := matrixSurfaces()
	if !slices.Contains(surfaces, c.Surface) {
		return fmt.Errorf("surface %q must be one of %s", c.Surface, strings.Join(surfaces, "|"))
	}
	modes := matrixModes()
	if !slices.Contains(modes, c.Mode) {
		return fmt.Errorf("mode %q must be one of %s", c.Mode, strings.Join(modes, "|"))
	}
	if len(c.Steps) == 0 {
		return errors.New("at least one step is required")
	}
	for i, st := range c.Steps {
		if err := st.validate(); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}
	if c.ExpectText == "" {
		return errors.New("expect_text is required")
	}
	if c.ExpectLevel > matrixLevelMax {
		return fmt.Errorf("expect_level %d must be 0 (skip), 1 or %d", c.ExpectLevel, matrixLevelMax)
	}

	return nil
}

// validate enforces the "exactly one step kind" rule and the tap/focus/key
// vocabularies.
func (s matrixStep) validate() error {
	set := 0
	for _, v := range []string{s.Type, s.Key, s.Tap, s.Focus} {
		if v != "" {
			set++
		}
	}
	if set != 1 {
		return fmt.Errorf("exactly one of type|key|tap|focus is required, got %d", set)
	}
	switch {
	case s.Tap != "":
		taps := matrixTaps()
		if !slices.Contains(taps, s.Tap) {
			return fmt.Errorf("tap %q must be one of %s", s.Tap, strings.Join(taps, "|"))
		}
	case s.Key != "":
		if s.Key != matrixSpaceKey {
			if _, ok := matrixKeyNames()[s.Key]; !ok {
				allowed := append(slices.Sorted(maps.Keys(matrixKeyNames())), matrixSpaceKey)

				return fmt.Errorf("key %q must be one of %s", s.Key, strings.Join(allowed, "|"))
			}
		}
	case s.Focus != "":
		surfaces := matrixSurfaces()
		if !slices.Contains(surfaces, s.Focus) {
			return fmt.Errorf("focus %q must be one of %s", s.Focus, strings.Join(surfaces, "|"))
		}
	}

	return nil
}

// matrixEntry is one case outcome: err == nil means PASS.
type matrixEntry struct {
	name string
	err  error
}

// matrixReport accumulates per-case outcomes and computes the run's exit
// code (TEST-04: non-zero on any FAIL).
type matrixReport struct {
	entries []matrixEntry
}

// add records one case outcome.
func (r *matrixReport) add(name string, err error) {
	r.entries = append(r.entries, matrixEntry{name: name, err: err})
}

// passed counts PASS entries.
func (r *matrixReport) passed() int {
	n := 0
	for _, e := range r.entries {
		if e.err == nil {
			n++
		}
	}

	return n
}

// failed counts FAIL entries.
func (r *matrixReport) failed() int {
	n := 0
	for _, e := range r.entries {
		if e.err != nil {
			n++
		}
	}

	return n
}

// exitCode computes the run's exit code: 1 when any case failed, else 0.
func (r *matrixReport) exitCode() int {
	if r.failed() > 0 {
		return 1
	}

	return 0
}

// lines renders the per-case report lines: "PASS <name>" or
// "FAIL <name>: <reason>" (the text-mismatch reason reads
// "expected %q, got %q").
func (r *matrixReport) lines() []string {
	lines := make([]string, 0, len(r.entries))
	for _, e := range r.entries {
		if e.err != nil {
			lines = append(lines, fmt.Sprintf("FAIL %s: %v", e.name, e.err))

			continue
		}
		lines = append(lines, "PASS "+e.name)
	}

	return lines
}

// runMatrixFile is the -matrix entry: load and validate the whole file
// BEFORE touching the live desktop, preflight the environment once, then
// run every case in its own full stand cycle. The report file is written
// even when the run aborts early (Ctrl-C): the artifact is the evidence.
func runMatrixFile(ctx context.Context, cfg config, path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL matrix: read %s: %v\n", path, err)

		return 1
	}
	cases, err := loadMatrixCases(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL matrix: %s: %v\n", path, err)

		return 1
	}
	if err := matrixPreflight(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL matrix: %v\n", err)

		return 1
	}

	var report matrixReport
	defer writeMatrixReport(path, &report)

	for _, c := range cases {
		if cerr := ctx.Err(); cerr != nil {
			report.add(c.Name, fmt.Errorf("run aborted: %w", cerr))

			break
		}
		err := runMatrixCaseIsolated(ctx, cfg, c)
		report.add(c.Name, err)
		if err != nil {
			fmt.Printf("FAIL %s: %v\n", c.Name, err)

			continue
		}
		fmt.Printf("PASS %s\n", c.Name)
	}
	fmt.Printf("matrix: %d/%d PASS (report: %s)\n", report.passed(), len(report.entries), matrixReportPath)

	return report.exitCode()
}

// matrixPreflight runs the stand-independent environment checks once per
// matrix run (the surface drivers themselves prove their binaries inside
// the cases that use them). Same table shape as preflight.go.
//
// The goswitch-slot checks exist because a 16-cycle run multiplies any
// leftover: a live goswitchd process or a stale ibus-side name entry (LIVE
// FINDING 2026-09-15: ibus-daemon 1.5.29 keeps the name of an abruptly
// killed engine connection — an interrupted matrix run leaves one, and
// every following daemon loses the single-instance guard with reply 2)
// would fail all sixteen cases with a message that never names the cause.
func matrixPreflight(ctx context.Context, cfg config) error {
	checks := []struct {
		name string
		run  func(context.Context, *stand) error
	}{
		{"injection-selftest", checkInjectionSelfTest},
		{"ibus-address", checkIbusAddress},
		{"uinput-writable", checkUinputWritable},
		{"python-gi", checkPythonGI},
		{"goswitch-process-absent", checkNoGoswitchProcess},
		{"goswitch-name-free", checkGoswitchNameFree},
	}
	s := &stand{cfg: cfg}
	for _, check := range checks {
		if err := check.run(ctx, s); err != nil {
			return fmt.Errorf("preflight %s: %w", check.name, err)
		}
		fmt.Printf("preflight %s: ok\n", check.name)
	}

	return nil
}

// checkNoGoswitchProcess refuses to start when a goswitchd process is
// already running: with a daemon alive, every matrix case would lose the
// single-instance guard — and a second live stand would inject into the
// same desktop (two parallel stands are forbidden, plan prohibition).
func checkNoGoswitchProcess(ctx context.Context, _ *stand) error {
	out, err := runCmd(ctx, "pgrep", "-x", "goswitchd")
	if err == nil && out != "" {
		return fmt.Errorf("a goswitchd process is already running (pid(s) %s)"+
			" — another stand or an installed daemon; stop it first", out)
	}

	return nil
}

// checkGoswitchNameFree refuses to start when the IBus bus already has an
// owner for the goswitch component name while NO goswitchd process exists
// — the stale-registration state an interrupted run leaves behind (the
// owner connection is gone, ibus-daemon still hands out reply 2). The
// remedy is the ibus-native `ibus restart` (the stand's own ibus-restart
// case proves the desktop survives it).
func checkGoswitchNameFree(ctx context.Context, _ *stand) error {
	owner, err := goswitchNameOwner(ctx)
	if err != nil {
		return fmt.Errorf("probe name ownership: %w", err)
	}
	if owner == "" {
		return nil
	}

	return errors.New("ibus-daemon reports an owner of " + goswitchComponentName +
		" but checkNoGoswitchProcess found no live process — a stale registration" +
		" from an interrupted run; run `ibus restart` to reap it, then rerun")
}

// goswitchNameOwner asks the IBus bus who owns the goswitch component
// name: "" when the name is free, the owner's unique name when held, an
// error only when the probe itself cannot run.
func goswitchNameOwner(ctx context.Context) (string, error) {
	addr, err := engine.Discover()
	if err != nil {
		return "", fmt.Errorf("discover ibus address: %w", err)
	}
	conn, err := dbus.Dial(addr)
	if err != nil {
		return "", fmt.Errorf("dial ibus bus: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if err := conn.Auth([]dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getpid()))}); err != nil {
		return "", fmt.Errorf("dbus auth: %w", err)
	}
	if err := conn.Hello(); err != nil {
		return "", fmt.Errorf("dbus hello: %w", err)
	}
	var owner string
	call := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").
		CallWithContext(ctx, "org.freedesktop.DBus.GetNameOwner", 0, goswitchComponentName)
	if call.Err != nil {
		// ibus-daemon words the free-name answer its own way (live-probed:
		// "Can not get name owner of …: no such name"); match both the
		// standard error name and the observed text.
		msg := strings.ToLower(call.Err.Error())
		if strings.Contains(call.Err.Error(), "NameHasNoOwner") || strings.Contains(msg, "no such name") {
			return "", nil // free — the expected state before any case runs
		}

		return "", fmt.Errorf("GetNameOwner: %w", call.Err)
	}
	if err := call.Store(&owner); err != nil {
		return "", fmt.Errorf("store GetNameOwner reply: %w", err)
	}

	return owner, nil
}

// matrixReportMode is the report file's permission (owner-only: the report
// contains the typed probe words).
const matrixReportMode = 0o600

// writeMatrixReport dumps the run's report lines to matrixReportPath — the
// CI artifact of plan 02-07. Best-effort: a write failure is a stderr note,
// never a crash of the reporting path itself.
func writeMatrixReport(path string, report *matrixReport) {
	var b strings.Builder
	fmt.Fprintf(&b, "goswitch e2e matrix — %s — %s\n", path, time.Now().Format(time.RFC3339))
	for _, line := range report.lines() {
		fmt.Fprintln(&b, line)
	}
	fmt.Fprintf(&b, "summary: %d/%d PASS\n", report.passed(), len(report.entries))
	if err := os.WriteFile(matrixReportPath, []byte(b.String()), matrixReportMode); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: write %s: %v\n", matrixReportPath, err)
	}
}

// witnessProbeTimeout bounds a single desktop-tree probe: a wedged AT-SPI
// walk must cost seconds, not the whole cmdTimeout budget (live finding
// 02-06: right after a case's teardown killed its surfaces, the registry
// briefly lists their app entries and a walk into one blocks until the
// registry reaps it).
const witnessProbeTimeout = 4 * time.Second

// witnessQuieceAttempts caps how many probe windows the quiesce gate waits
// for the a11y registry to finish reaping killed surfaces.
const witnessQuieceAttempts = 2

// matrixQuiesce waits until the desktop's AT-SPI tree answers a witness
// probe promptly again. Killed surfaces (the per-case teardown SIGKILLs
// chromium/GTE/zenity) leave app entries the at-spi registry reaps
// asynchronously; a tree walk started in that window blocks far past any
// sane poll and surfaces as "signal: killed" under the command deadline —
// failing healthy cases that merely opened a surface at the wrong moment
// (live finding 02-06, shuffled-order run). Waiting for the heal is
// correct: the wedge always clears on its own.
func matrixQuiesce(ctx context.Context, cfg config) error {
	overall := time.Now().Add(witnessQuieceAttempts * surfaceFocusWait)
	var lastErr error
	for {
		probe, cancel := context.WithTimeout(ctx, witnessProbeTimeout)
		_, err := runCmd(probe, "/usr/bin/python3", cfg.helper, "witness")
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err
		if time.Now().After(overall) {
			return fmt.Errorf("a11y witness not answering within %v: %w",
				witnessQuieceAttempts*surfaceFocusWait, lastErr)
		}
		if serr := sleepCtx(ctx, witnessPoll); serr != nil {
			return serr
		}
	}
}

// runMatrixCaseIsolated is the CASE ISOLATION step of the matrix (02-06
// must-have): the case runs inside a complete stand cycle — fresh daemon,
// fresh log, fresh desktop snapshot — so daemon-carried state (scriptMode,
// the correction buffer) dies with the cycle and no case can observe
// another's leftovers. Word-ru-en and word-mixed leave the daemon in RU
// mode; sharing a daemon would fail every later en case.
func runMatrixCaseIsolated(ctx context.Context, cfg config, c matrixCase) error {
	s, err := setupStand(ctx, cfg)
	if err != nil {
		return fmt.Errorf("stand setup: %w", err)
	}
	caseErr := func() error {
		if rerr := s.waitForLog(ctx, "component registered", registrationWait); rerr != nil {
			if s.countSub("another goswitchd instance already registered") > 0 {
				return fmt.Errorf("daemon registration: the goswitch name is taken"+
					" (single-instance guard fired; a concurrent stand or a stale ibus"+
					" registration — `ibus restart` reaps the latter): %w", rerr)
			}

			return fmt.Errorf("daemon registration: %w", rerr)
		}
		// The previous case's teardown may have left the a11y registry
		// mid-reap: gate on the desktop tree answering again before the
		// case's witness polls start.
		if qerr := matrixQuiesce(ctx, s.cfg); qerr != nil {
			return fmt.Errorf("desktop a11y quiesce: %w", qerr)
		}

		return runCaseWatchdog(ctx, c.Name, func(ctx context.Context, s *stand) error {
			return runMatrixCase(ctx, s, c)
		}, s)
	}()
	if caseErr != nil {
		s.printLogExcerpt()
	}
	//nolint:contextcheck // teardown and its restore verification must NOT
	// inherit the case context — they exist to run precisely when that
	// context is already cancelled (watchdog expiry, Ctrl-C), and both
	// build their own bounded contexts internally.
	s.teardown()
	//nolint:contextcheck // see above: the verification belongs to teardown.
	if rerr := s.verifyRestored(); rerr != nil {
		return errors.Join(caseErr, fmt.Errorf("teardown restore verification: %w", rerr))
	}

	return caseErr
}

// runMatrixCase drives one case on a live stand: engine activation, the
// primary surface with its witness gate, the optional RU-mode flip (AFTER
// the surface — the flip tap must land in the stand's own focused context,
// the proven 02-04 order; before the surface the tap reaches no engine
// context at all, live finding 02-06), every step in file order, then the
// expectation checks.
func runMatrixCase(ctx context.Context, s *stand, c matrixCase) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	if err := openMatrixSurface(ctx, s, c.Surface); err != nil {
		return err
	}
	if c.Mode == "ru" {
		// The daemon always boots in EN mode: a ru case flips first, gated
		// on the mode record (the flip fires at window expiry, Pitfall 5).
		if err := s.flipToRU(ctx, c.Name); err != nil {
			return err
		}
	}
	for i, st := range c.Steps {
		if err := runMatrixStep(ctx, s, c, st); err != nil {
			return fmt.Errorf("step %d (%s): %w", i+1, stepSummary(st), err)
		}
	}

	return verifyMatrixCase(ctx, s, c)
}

// stepSummary names a step for FAIL lines.
func stepSummary(st matrixStep) string {
	switch {
	case st.Type != "":
		return "type " + st.Type
	case st.Key != "":
		return "key " + st.Key
	case st.Tap != "":
		return "tap " + st.Tap
	case st.Focus != "":
		return "focus " + st.Focus
	}

	return "empty"
}

// openMatrixSurface spawns the case's primary surface through the existing
// drivers and holds on the witness gate (empty field, own focus) BEFORE
// any injection — text never reaches the owner's windows (T-02-06-02).
func openMatrixSurface(ctx context.Context, s *stand, name string) error {
	switch name {
	case matrixSurfaceZenity:
		if err := s.startZenity(ctx); err != nil {
			return err
		}
		if err := s.waitZenityEntry(ctx); err != nil {
			s.reapZenity()

			return err
		}

		return s.waitZenityChars(ctx, 0)
	case surfaceChromiumName:
		if err := s.startChromium(ctx); err != nil {
			return err
		}

		return s.waitChromiumInput(ctx, 0)
	case matrixSurfaceGTE:
		if err := s.startGTE(ctx); err != nil {
			return err
		}

		return s.waitGTEInput(ctx, 0)
	}

	return fmt.Errorf("unknown surface %q", name)
}

// runMatrixStep executes exactly one step kind.
func runMatrixStep(ctx context.Context, s *stand, c matrixCase, st matrixStep) error {
	switch {
	case st.Type != "":
		return s.injectText(ctx, st.Type)
	case st.Key != "":
		return pressMatrixKey(ctx, s, st.Key)
	case st.Tap != "":
		return tapMatrix(ctx, s, st.Tap)
	case st.Focus != "":
		return focusMatrixSurface(ctx, s, c, st.Focus)
	}

	return errors.New("empty step") // unreachable: validation rejects it
}

// pressMatrixKey maps a schema key name onto the physical path. The space
// separator rides the typing path (ydotool 0.1.8 resolves the NAME "space"
// to the physical S key, live finding 02-03); the other names go through
// the ydotool key table under their working names ("Escape"→"esc", live
// finding 02-05).
func pressMatrixKey(ctx context.Context, s *stand, key string) error {
	if key == matrixSpaceKey {
		return s.injectText(ctx, " ")
	}
	name, ok := matrixKeyNames()[key]
	if !ok {
		return fmt.Errorf("unknown key %q", key)
	}

	return s.pressKey(ctx, name)
}

// tapMatrix injects K Right Shift taps inside the disambiguation window and
// gates on the FSM decision record BEFORE returning — the RU part of a case
// must never be typed before a single-tap flip has actually fired
// (Pitfall 5). A single tap is the script flip: its mode record is the
// additional gate.
func tapMatrix(ctx context.Context, s *stand, tap string) error {
	var n int
	switch tap {
	case "single":
		n = 1
	case "double":
		n = 2
	case "triple":
		n = 3
	default:
		return fmt.Errorf("unknown tap %q", tap)
	}
	if err := s.injectKeys(ctx, slices.Repeat([]string{"Shift_R"}, n)...); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, fmt.Sprintf(`"msg":"action","n":%d`, n), decisionWait); err != nil {
		return fmt.Errorf("tap %s decision: %w", tap, err)
	}
	if n == 1 {
		if err := s.waitForLog(ctx, `"msg":"mode"`, decisionWait); err != nil {
			return fmt.Errorf("single-tap script flip: %w", err)
		}
	}

	return nil
}

// focusMatrixSurface makes name the focused input surface — the focus-reset
// class of CORR-09 needs a REAL focus flip between two stand-owned
// surfaces. A surface not yet open is spawned fresh (a freshly mapped
// window takes focus, phase-01 finding); an already-open one is re-focused
// through the pid-keyed grabFocus poke.
//
// The zenity spawn path gates on `focused-inputs` ENUMERATION, not the
// single-witness poll: the surface being left (e.g. the stand's chromium)
// can keep a stale AT-SPI FOCUSED bit for beats after losing compositor
// focus (live finding 02-06, reset-focus run), and the witness would keep
// naming it while zenity already owns the real keyboard focus. The
// already-open path gates on the case's expected rune count: the v1 focus
// steps always return to a field that holds the typed word, and the count
// is what distinguishes the page input from the browser's own focused
// chrome (an omnibox that took the Tab sits at 0 chars and must not pass).
func focusMatrixSurface(ctx context.Context, s *stand, c matrixCase, name string) error {
	if pid := matrixSurfacePID(s, name); pid != 0 {
		return s.waitMatrixInputPid(ctx, pid, utf8.RuneCountInString(c.ExpectText))
	}
	if name == matrixSurfaceZenity {
		return openZenityFocused(ctx, s)
	}

	return openMatrixSurface(ctx, s, name)
}

// openZenityFocused spawns a zenity entry and gates on it appearing among
// the focused input candidates, poking its input node past the grace —
// the focus-flip variant of the zenity driver for cases that already have
// another surface open.
func openZenityFocused(ctx context.Context, s *stand) error {
	if err := s.startZenity(ctx); err != nil {
		return err
	}
	pid := strconv.Itoa(s.zenity.Process.Pid)
	deadline := time.Now().Add(surfaceFocusWait)
	nextGrab := time.Now().Add(zenityGrabGrace)
	for {
		out, err := runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "focused-inputs")
		if err != nil {
			s.reapZenity()

			return fmt.Errorf("focused-inputs gate: %w", err)
		}
		for _, line := range strings.Split(out, "\n") {
			if strings.HasPrefix(line, matrixSurfaceZenity+":") {
				return nil
			}
		}
		if time.Now().After(deadline) {
			s.reapZenity()

			return fmt.Errorf("zenity entry did not take focus (focused inputs %q)", out)
		}
		if time.Now().After(nextGrab) {
			// Best-effort poke: the refused grab still re-issues the
			// activation request (02-02 finding).
			_, _ = runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "grab-input-pid", pid)
			nextGrab = time.Now().Add(zenityGrabEvery)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			s.reapZenity()

			return err
		}
	}
}

// matrixSurfacePID returns the stand's spawned process id of a surface
// (0 when the stand has none open under that name).
func matrixSurfacePID(s *stand, name string) int {
	switch name {
	case matrixSurfaceZenity:
		if s.zenity != nil && s.zenity.Process != nil {
			return s.zenity.Process.Pid
		}
	case surfaceChromiumName:
		if s.chromium != nil && s.chromium.Process != nil {
			return s.chromium.Process.Pid
		}
	case matrixSurfaceGTE:
		if s.gte != nil && s.gte.Process != nil {
			return s.gte.Process.Pid
		}
	}

	return 0
}

// waitMatrixInputPid polls until the pid's input surface holds focus with
// exactly wantChars runes — the waitGTEInput recovery pattern generalized
// for the matrix, where a case's own keys (Tab) or focus steps can move
// focus off the field mid-case. Past the grace the loop pokes the input
// node with the pid-keyed grabFocus: the call's refusal still re-issues
// the window activation (02-02 finding). The char count is load-bearing
// on chromium: an omnibox that swallowed the Tab holds 0 chars and must
// not pass for the page input.
func (s *stand) waitMatrixInputPid(ctx context.Context, pid, wantChars int) error {
	if pid == 0 {
		return errors.New("no surface process to witness")
	}
	pidStr := strconv.Itoa(pid)
	want := ":chars=" + strconv.Itoa(wantChars)
	deadline := time.Now().Add(surfaceFocusWait)
	nextGrab := time.Now().Add(gteGrabGrace)
	var last string
	for {
		out, err := runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "focused-input-pid", pidStr)
		if err != nil {
			return fmt.Errorf("focused-input-pid gate: %w", err)
		}
		last = out
		if strings.HasSuffix(out, want) {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("surface pid %s input not focused with %s (witness %q)", pidStr, want, last)
		}
		if time.Now().After(nextGrab) {
			// Best-effort poke: a no-op while the app tree is not up. The
			// char filter is what aims the grab at the page input on
			// chromium (the post-Tab omnibox is the app's first focusable
			// input and would otherwise swallow every poke).
			_, _ = runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "grab-input-pid", pidStr, strconv.Itoa(wantChars))
			nextGrab = time.Now().Add(gteGrabEvery)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// verifyMatrixCase checks the expectations: the ACTUAL ladder level from
// the daemon's -debug record (ADR-003) and the field content rune for
// rune. The verify-after match record is deliberately NOT a matrix
// oracle — LIVE FINDING (2026-09-15, the first 02-06 matrix runs): the
// daemon's verify-after round settles on the FIRST surrounding-text push
// after the correction, and a client may push the INTERMEDIATE
// post-delete pre-commit state (cursor 0), producing a false mismatch
// while the field itself ends up correct; the race hit zenity and
// gnome-text-editor on consecutive runs (nondeterministic). The
// content-exact readback below is the ground truth; the match contract
// lives in the unit corpus (TestActor_VerifyAfterLevel1) and the
// ladder-chromium case; the stale-answer race itself is a deferred item
// (internal/session, outside this plan's file scope).
func verifyMatrixCase(ctx context.Context, s *stand, c matrixCase) error {
	if c.ExpectLevel > 0 {
		mark := fmt.Sprintf(`"msg":"correction","level":%d`, c.ExpectLevel)
		if err := s.waitForLog(ctx, mark, correctionWait); err != nil {
			return fmt.Errorf("expected actual ladder level %d: %w", c.ExpectLevel, err)
		}
	}

	return verifyMatrixText(ctx, s, c)
}

// verifyMatrixText compares the field content with expect_text rune for
// rune. The readback rides the AT-SPI bridge with settle polling (the
// bridge can lag the field by a beat, 02-03 finding); a zenity case whose
// steps closed the dialog falls back to the stdout close-oracle.
func verifyMatrixText(ctx context.Context, s *stand, c matrixCase) error {
	want := c.ExpectText
	runes := utf8.RuneCountInString(want)
	switch c.Surface {
	case matrixSurfaceZenity:
		if zenityAlive(s) {
			if err := s.waitZenityText(ctx, want); err != nil {
				return err
			}
			out, err := s.closeZenity(ctx)
			if err != nil {
				return err
			}
			// The delivery oracle: zenity prints the entry text on OK. The
			// printed line loses boundary whitespace, so it pins the word
			// only when the expectation itself carries none — the readback
			// above already pinned the exact content.
			if out != want && strings.TrimSpace(want) == want {
				return fmt.Errorf("expected %q, got %q (stdout oracle)", want, out)
			}

			return nil
		}
		// Steps closed the dialog (Enter): the collected stdout IS the oracle.
		out := strings.TrimSpace(s.zenityOut.String())
		if out != strings.TrimSpace(want) {
			return fmt.Errorf("expected %q, got %q (closed-dialog stdout)", strings.TrimSpace(want), out)
		}

		return nil
	case surfaceChromiumName:
		if err := s.waitMatrixInputPid(ctx, matrixSurfacePID(s, surfaceChromiumName), runes); err != nil {
			return fmt.Errorf("applied state: %w", err)
		}

		return waitMatrixReadback(ctx, want, s.readChromiumText)
	case matrixSurfaceGTE:
		if err := s.waitGTEInput(ctx, runes); err != nil {
			return fmt.Errorf("applied state: %w", err)
		}

		return waitMatrixReadback(ctx, want, s.readGTEText)
	}

	return fmt.Errorf("unknown surface %q", c.Surface)
}

// zenityAlive reports whether the spawned zenity entry is still running
// (a step's Enter closes the dialog; the process then exits by itself).
func zenityAlive(s *stand) bool {
	return s.zenity != nil && s.zenity.Process != nil &&
		s.zenity.Process.Signal(signalProbe) == nil
}

// waitMatrixReadback polls a readback function until it returns exactly
// want (settle-riding, never a single read) and reports the plan's FAIL
// wording on a mismatch.
func waitMatrixReadback(ctx context.Context, want string, read func(context.Context) (string, error)) error {
	deadline := time.Now().Add(witnessWait)
	var got string
	for {
		out, err := read(ctx)
		if err == nil {
			got = out
			if out == want {
				return nil
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("expected %q, got %q", want, got)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}
