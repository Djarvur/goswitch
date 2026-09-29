package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/Djarvur/goswitch/engine"
)

// Switch-spike (plan 05-01, SWCH-03 criterion 2): the live reversible
// two-source window that settles the phase's central question BEFORE any
// dependent daemon code — does GNOME Shell 46 follow an external
// SetGlobalEngine when the sources list carries TWO goswitch engines
// (indicator, XKB group, GlobalEngineChanged delivery, self-echo). The
// case snapshots the desktop and the unit daemon, stops the unit
// (single-instance: the stand daemon must own the engines), writes the
// two-source list, runs the P1-P4 probe matrix and restores the snapshot
// in a defer; every probe leaves an observable verdict row in
// switch-spike-report.txt. The typed probe word never reaches the report
// or the logs (D-20/D-21): its script class does.

// The two-source window value: the FULL sources list, not a prefix — the
// snapshot is restored verbatim by the defer, never merged.
const twoSources = "[('ibus', 'goswitch-en'), ('ibus', 'goswitch-ru')]"

// The engine names the spike flips between (config-level literals — never
// user data).
const (
	spikeEngineEN = "goswitch-en"
	spikeEngineRU = "goswitch-ru"
)

// unitActiveState is the systemd state word the spike's unit snapshot
// treats as "restore after the window".
const unitActiveState = "active"

// spikeDetailLimit caps an error row's reason so the report stays
// one-line-per-probe.
const spikeDetailLimit = 120

// The spike's own godbus connection targets the same IBus service and
// object path the daemon side uses (literals here — the engine package's
// constants are unexported). The component's well-known name is the
// matrix's goswitchComponentName mirror.
const (
	ibusDBusService = "org.freedesktop.IBus"
	ibusDBusPath    = dbus.ObjectPath("/org/freedesktop/IBus")
)

// Spike timing: the signal waits ride the d01 probe budget; the
// observation pauses are the owner's invitation to watch the indicator
// before each flip (SWCH-03 is eyes-only); the unit and name-release
// budgets bound the reversible window's setup and teardown.
const (
	spikeObservePause = 5 * time.Second
	spikeStateWait    = 15 * time.Second
	spikeSpawnDelay   = 3 * time.Second
	spikeSpawnTry     = 3
	spikeSignalBuffer = 16
)

// modeRecordMark is the greppable prefix of the daemon's mode record —
// narrow enough to never match the other "mode"-named records.
const modeRecordMark = `"msg":"mode","to":`

// spikeReportPath is the spike's verdict report at the repo root (the
// perf-report.txt/e2e-report.txt convention).
const spikeReportPath = "switch-spike-report.txt"

// spikeIndicatorLine is the report's final line: the owner's indicator
// verdict (follows / not-follows — eyes only, SWCH-03), filled at the
// Task-2 checkpoint; the case always writes the unknown placeholder.
const spikeIndicatorLine = "indicator-verdict: unknown"

// The closed verdict set of every spike line.
const (
	verdictFollows    = "follows"
	verdictNotFollows = "not-follows"
	verdictError      = "error"
)

// Readback classes of the P3 typing rounds (D-20: the typed word itself
// never reaches the report — its script class does).
const (
	readbackCyrillic = "cyrillic"
	readbackLatin    = "latin"
)

// spikeProbe is one probe's rendered verdict row.
type spikeProbe struct {
	ID       string
	Name     string
	Observed string
	Verdict  string
	Detail   string
}

// spikeProbeSpec pins the probe plan: fixed, complete, ordered.
type spikeProbeSpec struct {
	ID   string
	Name string
}

// closedVerdict reports whether v belongs to the closed verdict set.
func closedVerdict(v string) bool {
	switch v {
	case verdictFollows, verdictNotFollows, verdictError:
		return true
	}

	return false
}

// newSpikeProbe builds a verdict row, enforcing the closed set at
// construction so a report line can never carry an out-of-vocabulary
// verdict.
func newSpikeProbe(id, name, observed, verdict, detail string) (spikeProbe, error) {
	if !closedVerdict(verdict) {
		return spikeProbe{}, fmt.Errorf("verdict %q outside the closed set {follows, not-follows, error}", verdict)
	}

	return spikeProbe{ID: id, Name: name, Observed: observed, Verdict: verdict, Detail: detail}, nil
}

// renderVerdictTable renders one (probe | observed | verdict) line per
// probe; the error verdict carries its one-line reason.
func renderVerdictTable(probes []spikeProbe) string {
	var b strings.Builder
	for _, p := range probes {
		line := p.ID + " " + p.Name + " | " + p.Observed + " | " + p.Verdict
		if p.Verdict == verdictError && p.Detail != "" {
			line += " (" + strings.Join(strings.Fields(p.Detail), " ") + ")"
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

// spikeProbePlan returns the fixed probe sequence — P1..P4, complete, in
// the research-prescribed order; the case never reorders or drops one.
func spikeProbePlan() []spikeProbeSpec {
	return []spikeProbeSpec{
		{ID: "P1", Name: "shell-activation"},
		{ID: "P2", Name: "external-flip-signal"},
		{ID: "P3", Name: "xkb-truth"},
		{ID: "P4", Name: "self-echo"},
	}
}

// renderSpikeReport renders the whole report: the verdict table plus the
// final indicator-verdict line (the owner's checkpoint fills it; the case
// always writes the unknown placeholder).
func renderSpikeReport(probes []spikeProbe) string {
	return renderVerdictTable(probes) + spikeIndicatorLine + "\n"
}

// xkbVerdictFromClasses derives the P3 XKB verdict from the EN-pinned
// readback class ONLY: cyrillic means the XKB group followed the external
// flip, latin means it did not. The RU-pinned class is the safety-net
// witness (cyrillic expected in BOTH XKB worlds — the modeRU commit branch
// converts latin keyvals itself) and is never a discriminator.
func xkbVerdictFromClasses(enClass, ruClass string) (string, error) {
	if !isReadbackClass(ruClass) {
		return "", fmt.Errorf("ru readback class %q outside {%s, %s}", ruClass, readbackCyrillic, readbackLatin)
	}
	switch enClass {
	case readbackCyrillic:
		return verdictFollows, nil
	case readbackLatin:
		return verdictNotFollows, nil
	default:
		return "", fmt.Errorf("en readback class %q outside {%s, %s}", enClass, readbackCyrillic, readbackLatin)
	}
}

// isReadbackClass reports whether v belongs to the closed readback class set.
func isReadbackClass(v string) bool {
	switch v {
	case readbackCyrillic, readbackLatin:
		return true
	}

	return false
}

// mustProbe builds a row from an in-vocabulary verdict; the fallback row
// keeps the closed-set invariant without ever panicking the stand.
func mustProbe(id, name, observed, verdict string) spikeProbe {
	p, err := newSpikeProbe(id, name, observed, verdict, "")
	if err != nil {
		return spikeProbe{ID: id, Name: name, Observed: observed, Verdict: verdictError,
			Detail: "internal verdict vocabulary error"}
	}

	return p
}

// errorProbe builds the error row for a probe whose observation could not
// be made; the reason is flattened to one line (the report stays row-per-probe).
func errorProbe(id, name, observed string, err error) spikeProbe {
	detail := "unavailable"
	if err != nil {
		detail = err.Error()
	}
	if len(detail) > spikeDetailLimit {
		detail = detail[:spikeDetailLimit]
	}
	p, _ := newSpikeProbe(id, name, observed, verdictError, detail)

	return p
}

// spikeSession is the harness's own godbus connection to the ibus socket:
// the subscription that P2/P4 read GlobalEngineChanged from and the
// endpoint P3/P4 flip engines through — the daemon-path call shape the
// phase's mechanism would use.
type spikeSession struct {
	conn    *dbus.Conn
	signals chan *dbus.Signal
}

// runSwitchSpike executes the spike matrix on the live desktop. Standalone
// case: it owns its daemon lifecycle around the unit daemon's (single
// instance — the stand daemon must own the engines), so setupStand built
// the binary and snapshotted the desktop but spawned nothing.
func runSwitchSpike(ctx context.Context, s *stand) error {
	unit := unitState(ctx)
	fmt.Printf("switch-spike: unit goswitchd state %q\n", unit)

	probes := make([]spikeProbe, 0, len(spikeProbePlan()))
	reported := false
	// Registered first → runs last: the report always lands, with error
	// rows for any probe that did not complete (none lost before the write).
	defer func() {
		if reported {
			return
		}
		if werr := writeSpikeReport(probes); werr != nil {
			fmt.Fprintf(os.Stderr, "e2e: switch-spike deferred report write: %v\n", werr)
		}
	}()

	if unit == unitActiveState {
		if _, err := runCmd(ctx, "systemctl", "--user", "stop", "goswitchd"); err != nil {
			return fmt.Errorf("switch-spike: stop unit daemon: %w", err)
		}
	}
	// Registered second → runs after the probes: the reversible window's
	// restore — sources from the snapshot (write-only-on-diff), stand
	// daemon killed, unit daemon back to its snapshotted state. A fresh
	// bounded context inside the method: this must also run when the
	// watchdog cancelled ctx.
	defer func() { s.restoreSpikeWindow(ctx, unit) }()

	if err := waitNameFree(ctx); err != nil {
		return err
	}
	if err := s.startDaemonRegistered(ctx); err != nil {
		return err
	}

	sess, err := spikeSubscribe()
	if err != nil {
		return err
	}
	defer func() {
		if cerr := sess.conn.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "e2e: switch-spike close ibus probe connection: %v\n", cerr)
		}
	}()

	// The two-source window: write-only-on-diff (a value-identical write
	// still triggers GNOME's re-evaluation and live-unsets the global
	// engine — Pitfall 6), with the P1 focus baselines taken BEFORE the
	// write so the shell's activation is measured as NEW records.
	enBase := s.countSub(focusInEN)
	ruBase := s.countSub(focusInRU)
	wrote, err := s.writeSourcesIfDiffers(ctx, twoSources)
	if err != nil {
		return err
	}

	// Argument evaluation order IS the probe order (P1..P4).
	probes = append(probes,
		s.spikeProbeShellActivation(ctx, enBase, ruBase, wrote),
		s.spikeProbeExternalFlipSignal(ctx, sess),
		s.spikeProbeXKBTruth(ctx, sess),
		s.spikeProbeSelfEcho(ctx, sess),
	)

	if err := writeSpikeReport(probes); err != nil {
		return err
	}
	reported = true

	return nil
}

// unitState snapshots the unit daemon's state as "active" or "inactive" —
// only the active state is restored verbatim; every other systemd state
// (failed, activating) is left for the owner's session manager.
func unitState(ctx context.Context) string {
	if state, err := runCmd(ctx, "systemctl", "--user", "is-active", "goswitchd"); err == nil {
		return state
	}

	return "inactive"
}

// restoreSpikeDesktop puts the reversible window back: gsettings from the
// snapshot (the stand's write-only-on-diff restore), the stand daemon
// killed (the unit needs the component name free), the unit daemon back
// to its snapshotted state. The engine re-assertion stays with run()'s
// teardown, which runs after this and never targets a goswitch engine
// unless the snapshot itself carries one.
func (s *stand) restoreSpikeWindow(ctx context.Context, unit string) {
	// The window's restore must run even when the watchdog cancelled the
	// case context — WithoutCancel keeps the values, drops the deadline.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), teardownTimeout)
	defer cancel()
	s.restoreGsettings(ctx)
	s.stopDaemon()
	restoreUnit(ctx, unit)
}

// restoreUnit brings the unit daemon back when the snapshot caught it
// active; best-effort with a bounded wait — a failing restore is a loud
// stderr note, never a panic on the teardown path.
func restoreUnit(ctx context.Context, want string) {
	if want != unitActiveState {
		return
	}
	// A unit left in the failed state refuses start; reset is harmless otherwise.
	_, _ = runCmd(ctx, "systemctl", "--user", "reset-failed", "goswitchd")
	if _, err := runCmd(ctx, "systemctl", "--user", "start", "goswitchd"); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: switch-spike restore unit start: %v\n", err)

		return
	}
	deadline := time.Now().Add(spikeStateWait)
	for {
		if unitState(ctx) == unitActiveState {
			return
		}
		if time.Now().After(deadline) {
			fmt.Fprintf(os.Stderr, "e2e: switch-spike unit not active %v after restore\n", spikeStateWait)

			return
		}
		if err := sleepCtx(ctx, logPollInterval); err != nil {
			return
		}
	}
}

// waitNameFree blocks until the component's well-known name is claimable —
// the stopped unit daemon must have released it before the stand daemon's
// fatal-on-conflict RequestName. Probed non-destructively: DoNotQueue
// without ReplaceExisting never disturbs a live holder.
func waitNameFree(ctx context.Context) error {
	addr, err := engine.Discover()
	if err != nil {
		return fmt.Errorf("switch-spike: discover ibus: %w", err)
	}
	deadline := time.Now().Add(spikeStateWait)
	for {
		free, perr := probeNameFree(addr)
		if perr != nil {
			return fmt.Errorf("switch-spike: name probe: %w", perr)
		}
		if free {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("switch-spike: %s still held %v after the unit stop —"+
				" the previous owner has not released the component name", goswitchComponentName, spikeStateWait)
		}
		if err := sleepCtx(ctx, logPollInterval); err != nil {
			return err
		}
	}
}

// probeNameFree claims the component name with DoNotQueue and releases it
// back when won — PrimaryOwner means free, EXISTS means held.
func probeNameFree(addr string) (bool, error) {
	conn, err := dbus.Dial(addr)
	if err != nil {
		return false, fmt.Errorf("dial ibus bus: %w", err)
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "e2e: close ibus name probe connection: %v\n", cerr)
		}
	}()
	if err := conn.Auth([]dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}); err != nil {
		return false, fmt.Errorf("dbus auth: %w", err)
	}
	if err := conn.Hello(); err != nil {
		return false, fmt.Errorf("dbus hello: %w", err)
	}
	reply, err := conn.RequestName(goswitchComponentName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return false, fmt.Errorf("request name: %w", err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return false, nil
	}
	if _, err := conn.ReleaseName(goswitchComponentName); err != nil {
		return false, fmt.Errorf("release name: %w", err)
	}

	return true, nil
}

// startDaemonRegistered spawns the stand daemon and waits for its
// component registration; a name-conflict death (the unit daemon's release
// still in flight) is retried with a fresh spawn, bounded.
func (s *stand) startDaemonRegistered(ctx context.Context) error {
	var lastErr error
	for range spikeSpawnTry {
		if err := s.startDaemonArgs(); err != nil {
			return err
		}
		err := s.waitForLog(ctx, componentRegisteredMark, registrationWait)
		if err == nil {
			return nil
		}
		lastErr = err
		if s.countSub("another goswitchd instance already registered") == 0 {
			return fmt.Errorf("switch-spike: stand daemon registration: %w", lastErr)
		}
		s.stopDaemon()
		if err := sleepCtx(ctx, spikeSpawnDelay); err != nil {
			return err
		}
	}

	return fmt.Errorf("switch-spike: stand daemon did not register after %d attempts: %w", spikeSpawnTry, lastErr)
}

// spikeSubscribe dials the ibus socket with the harness's own connection
// and subscribes to the IBus-interface signals at the IBus object path —
// the AddMatch shape the research's Pattern 3 prescribes (assumption A2
// — AddMatch works on the ibus socket — is itself under probe here).
func spikeSubscribe() (*spikeSession, error) {
	addr, err := engine.Discover()
	if err != nil {
		return nil, fmt.Errorf("switch-spike: discover ibus: %w", err)
	}
	conn, err := dbus.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("switch-spike: dial ibus bus: %w", err)
	}
	if err := conn.Auth([]dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("switch-spike: dbus auth: %w", err)
	}
	if err := conn.Hello(); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("switch-spike: dbus hello: %w", err)
	}
	if err := conn.AddMatchSignal(
		dbus.WithMatchObjectPath(ibusDBusPath),
		dbus.WithMatchInterface(ibusDBusService),
	); err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("switch-spike: AddMatchSignal on the ibus socket: %w", err)
	}
	ch := make(chan *dbus.Signal, spikeSignalBuffer)
	conn.Signal(ch)

	return &spikeSession{conn: conn, signals: ch}, nil
}

// writeSourcesIfDiffers writes the two-source list only when the live
// value differs; the bool reports whether a write happened (P1 annotates
// its observation when the desktop was already two-source).
func (s *stand) writeSourcesIfDiffers(ctx context.Context, value string) (bool, error) {
	now, err := runCmd(ctx, "gsettings", "get", gsettingsSchema, keySources)
	if err != nil {
		return false, fmt.Errorf("switch-spike: read sources: %w", err)
	}
	if now == value {
		return false, nil
	}
	if _, err := runCmd(ctx, "gsettings", "set", gsettingsSchema, keySources, value); err != nil {
		return false, fmt.Errorf("switch-spike: set two-source list: %w", err)
	}

	return true, nil
}

// spikeProbeShellActivation answers P1: which engine did the shell
// activate after the two-source write — the observed `ibus engine`
// readback plus the NEW focus_in records of the written engines (the
// d01-observable discipline: journal and readback, never a dconf value).
func (s *stand) spikeProbeShellActivation(ctx context.Context, enBase, ruBase int, wrote bool) spikeProbe {
	const (
		id   = "P1"
		name = "shell-activation"
	)
	_ = s.waitForAnyNew(ctx, []string{focusInEN, focusInRU}, []int{enBase, ruBase}, d01ProbeWait)
	gotEN := s.countSub(focusInEN) - enBase
	gotRU := s.countSub(focusInRU) - ruBase
	readback, rbErr := runCmd(ctx, "ibus", "engine")
	if rbErr != nil {
		return errorProbe(id, name, "readback=unreadable", rbErr)
	}
	observed := fmt.Sprintf("readback=%s focus_in(en)=+%d focus_in(ru)=+%d", readback, gotEN, gotRU)
	if !wrote {
		observed += " (sources already two-source, write skipped)"
	}
	switch {
	case gotEN+gotRU >= 1 && isGoswitchEngine(readback):
		return mustProbe(id, name, observed, verdictFollows)
	case gotEN+gotRU == 0:
		return mustProbe(id, name, observed+" no-new-goswitch-focus within "+d01ProbeWait.String(), verdictNotFollows)
	default:
		return mustProbe(id, name, observed, verdictNotFollows)
	}
}

// spikeProbeExternalFlipSignal answers P2: does an externally-issued flip
// (`ibus engine goswitch-ru` — a probe subprocess, not the daemon path)
// deliver GlobalEngineChanged to the harness's own subscribed connection.
func (s *stand) spikeProbeExternalFlipSignal(ctx context.Context, sess *spikeSession) spikeProbe {
	const (
		id   = "P2"
		name = "external-flip-signal"
	)
	drainSignals(sess.signals)
	if err := observePause(ctx, id); err != nil {
		return errorProbe(id, name, "pause aborted", err)
	}
	if _, err := runCmd(ctx, "ibus", "engine", spikeEngineRU); err != nil {
		return errorProbe(id, name, "flip=failed", err)
	}
	readback, err := runCmd(ctx, "ibus", "engine")
	if err != nil {
		return errorProbe(id, name, "readback=unreadable", err)
	}
	got, delivered, werr := waitEngineChanged(sess.signals, d01ProbeWait)
	if werr != nil {
		return errorProbe(id, name, "flip="+spikeEngineRU+" readback="+readback, werr)
	}
	observed := "flip=" + spikeEngineRU + " readback=" + readback
	if !delivered {
		return mustProbe(id, name, observed+" signal=none within "+d01ProbeWait.String(), verdictNotFollows)
	}
	observed += " signal=GlobalEngineChanged(" + got + ")"
	if got == spikeEngineRU {
		return mustProbe(id, name, observed, verdictFollows)
	}

	return mustProbe(id, name, observed, verdictNotFollows)
}

// spikeProbeXKBTruth answers P3 — the ONLY XKB discriminator: with the
// daemon's mode pinned EN journally, the typed word's script class tells
// whether the XKB group followed the external flips (native cyrillic =
// followed, latin = not). The follow-up RU-pinned round witnesses the
// modeRU safety net (cyrillic in both XKB worlds) and never feeds the
// verdict.
func (s *stand) spikeProbeXKBTruth(ctx context.Context, sess *spikeSession) spikeProbe {
	const (
		id   = "P3"
		name = "xkb-truth"
	)
	for _, flip := range []string{spikeEngineEN, spikeEngineRU} {
		if err := observePause(ctx, id+" flip "+flip); err != nil {
			return errorProbe(id, name, "pause aborted", err)
		}
		if err := setGlobalEngineOn(ctx, sess.conn, flip); err != nil {
			return errorProbe(id, name, "readback=unreadable", err)
		}
		readback, err := runCmd(ctx, "ibus", "engine")
		if err != nil {
			return errorProbe(id, name, "readback=unreadable after flip "+flip, err)
		}
		if readback != flip {
			return errorProbe(id, name, "readback="+readback,
				fmt.Errorf("flip %s: readback shows %s", flip, readback))
		}
	}
	enClass, err := s.spikeTypeClass(ctx, "en")
	if err != nil {
		return errorProbe(id, name, "en_pinned=unreadable", err)
	}
	ruClass, err := s.spikeTypeClass(ctx, "ru")
	if err != nil {
		return errorProbe(id, name, "en_pinned="+enClass+" ru_pinned=unreadable", err)
	}
	observed := "en_pinned=" + enClass + " ru_pinned=" + ruClass
	verdict, verr := xkbVerdictFromClasses(enClass, ruClass)
	if verr != nil {
		return errorProbe(id, name, observed, verr)
	}

	return mustProbe(id, name, observed, verdict)
}

// spikeProbeSelfEcho answers P4: does the INITIATOR of a SetGlobalEngine
// receive its own GlobalEngineChanged — the echo-suppression input for
// the phase's sync-listener design.
func (s *stand) spikeProbeSelfEcho(ctx context.Context, sess *spikeSession) spikeProbe {
	const (
		id   = "P4"
		name = "self-echo"
	)
	drainSignals(sess.signals)
	if err := observePause(ctx, id); err != nil {
		return errorProbe(id, name, "pause aborted", err)
	}
	if err := setGlobalEngineOn(ctx, sess.conn, spikeEngineEN); err != nil {
		return errorProbe(id, name, "flip=failed", err)
	}
	got, delivered, werr := waitEngineChanged(sess.signals, d01ProbeWait)
	if werr != nil {
		return errorProbe(id, name, "flip="+spikeEngineEN, werr)
	}
	if !delivered {
		return mustProbe(id, name, "flip="+spikeEngineEN+" initiator-signal=none within "+d01ProbeWait.String(),
			verdictNotFollows)
	}

	return mustProbe(id, name, "flip="+spikeEngineEN+" initiator-signal=GlobalEngineChanged("+got+")", verdictFollows)
}

// spikeTypeClass runs one typing round on a fresh zenity entry: pin the
// daemon's mode to wantMode through injected Shift_R taps (journally —
// the fresh daemon's zero-value EN default is never trusted), type the
// probe word, settle the field, read it back and reduce it to its script
// class (D-20: the word itself stays out of the report).
func (s *stand) spikeTypeClass(ctx context.Context, wantMode string) (string, error) {
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return "", fmt.Errorf("switch-spike entry surface: %w", err)
	}
	if kind != surfaceZenity {
		_ = s.closeEntrySurface(ctx, kind)

		return "", errors.New("switch-spike needs the zenity entry surface (locked-session fallback engaged?)")
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()

	if err := s.spikePinMode(ctx, wantMode); err != nil {
		return "", err
	}
	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return "", fmt.Errorf("switch-spike typing: %w", err)
	}
	if err := s.waitZenityChars(ctx, len([]rune(wordProbeEN))); err != nil {
		return "", fmt.Errorf("switch-spike typed word settle: %w", err)
	}
	out, err := s.closeZenity(ctx)
	if err != nil {
		return "", fmt.Errorf("switch-spike readback: %w", err)
	}
	if out == "" {
		return "", errors.New("switch-spike: empty readback")
	}
	if hasCyrillic(out) {
		return readbackCyrillic, nil
	}

	return readbackLatin, nil
}

// spikePinMode drives the daemon's mode to wantMode with injected Shift_R
// taps, gated on the daemon's OWN mode records — the journal is the pin's
// proof, never the assumed boot default. Bounded at two taps.
func (s *stand) spikePinMode(ctx context.Context, wantMode string) error {
	const maxTaps = 2
	mark := modeRecordMark + `"` + wantMode + `"`
	for range maxTaps {
		last, err := s.lastModeMark()
		if err != nil {
			return err
		}
		if last == wantMode {
			return nil
		}
		base := s.countSub(modeRecordMark)
		if err := s.injectKeys(ctx, "Shift_R"); err != nil {
			return fmt.Errorf("switch-spike mode pin tap: %w", err)
		}
		if err := s.waitForNew(ctx, modeRecordMark, base+1, decisionWait); err != nil {
			return fmt.Errorf("switch-spike mode pin: tap produced no mode record: %w", err)
		}
	}
	last, err := s.lastModeMark()
	if err != nil {
		return err
	}
	if last != wantMode {
		return fmt.Errorf("switch-spike mode pin: %d taps left the mode at %q (want %q, mark %s)",
			maxTaps, last, wantMode, mark)
	}

	return nil
}

// lastModeMark scans the daemon log backwards for the newest mode record
// and returns its target ("en"/"ru"); an empty string means no record yet
// — the fresh daemon's mode is unproven, and the pin taps will prove it.
func (s *stand) lastModeMark() (string, error) {
	lines := s.logLines()
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if !strings.Contains(line, modeRecordMark) {
			continue
		}
		switch {
		case strings.Contains(line, `"to":"en"`):
			return "en", nil
		case strings.Contains(line, `"to":"ru"`):
			return "ru", nil
		default:
			return "", fmt.Errorf("switch-spike: unparsable mode record %q", line)
		}
	}

	return "", nil
}

// setGlobalEngineOn issues the daemon-path flip over the harness's own
// connection — the exact call shape the phase's switcher seam would use.
func setGlobalEngineOn(ctx context.Context, conn *dbus.Conn, engineName string) error {
	call := conn.Object(ibusDBusService, ibusDBusPath).
		CallWithContext(ctx, ibusDBusService+".SetGlobalEngine", 0, engineName)
	if call.Err != nil {
		return fmt.Errorf("SetGlobalEngine(%s): %w", engineName, call.Err)
	}

	return nil
}

// waitEngineChanged waits for one GlobalEngineChanged signal (any engine
// name) on the subscribed channel; the bool reports delivery, the string
// the delivered engine name.
func waitEngineChanged(ch <-chan *dbus.Signal, timeout time.Duration) (engineName string, delivered bool, err error) {
	deadline := time.After(timeout)
	for {
		select {
		case sig, ok := <-ch:
			if !ok {
				return "", false, errors.New("switch-spike: signal channel closed (ibus connection lost)")
			}
			if sig == nil || sig.Name != ibusDBusService+".GlobalEngineChanged" || len(sig.Body) == 0 {
				continue
			}
			engineName, _ = sig.Body[0].(string)

			return engineName, true, nil
		case <-deadline:
			return "", false, nil
		}
	}
}

// drainSignals drops every buffered signal — each signal probe baselines
// its channel before the action so only fresh deliveries count.
func drainSignals(ch <-chan *dbus.Signal) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}

// observePause prints the owner's watching invitation and pauses — the
// indicator verdict is eyes-only (SWCH-03), and the pause is the window
// the owner watches in.
func observePause(ctx context.Context, what string) error {
	fmt.Printf("switch-spike: %s — НАБЛЮДАЙТЕ индикатор верхней панели: флип через %v\n", what, spikeObservePause)

	return sleepCtx(ctx, spikeObservePause)
}

// isGoswitchEngine reports whether name is one of the two goswitch
// engines (config-level literals — never user data).
func isGoswitchEngine(name string) bool {
	return name == spikeEngineEN || name == spikeEngineRU
}

// waitForAnyNew blocks until ANY of the marks gains new occurrences past
// its baseline — P1's either-engine activation wait.
func (s *stand) waitForAnyNew(ctx context.Context, subs []string, bases []int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		for i, sub := range subs {
			if s.countSub(sub) > bases[i] {
				return nil
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout after %v: no new records among %v", timeout, subs)
		}
		if err := sleepCtx(ctx, logPollInterval); err != nil {
			return err
		}
	}
}

// writeSpikeReport renders the report from the fixed plan — every plan
// entry lands exactly once, probes that did not complete become error
// rows — and writes it to the repo root.
func writeSpikeReport(probes []spikeProbe) error {
	rows := make([]spikeProbe, 0, len(spikeProbePlan()))
	for _, spec := range spikeProbePlan() {
		row, found := findProbe(probes, spec.ID)
		if !found {
			row = mustProbe(spec.ID, spec.Name, "not reached", verdictError)
			row.Detail = "probe did not complete"
		}
		rows = append(rows, row)
	}
	if err := os.WriteFile(spikeReportPath, []byte(renderSpikeReport(rows)), configFilePerm); err != nil {
		return fmt.Errorf("write %s: %w", spikeReportPath, err)
	}

	return nil
}

// findProbe returns the recorded row for an ID.
func findProbe(probes []spikeProbe, id string) (spikeProbe, bool) {
	for _, p := range probes {
		if p.ID == id {
			return p, true
		}
	}

	return spikeProbe{}, false
}
