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

	"github.com/Djarvur/goswitch/internal/engine"
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

// The flip-marks vocabulary of plan 05-05: the closed flip-target set and
// the daemon journal marks whose ORDER is the only observable form of the
// switching act (D-34 heritage: the daemon's journal, never dconf). The
// flipTo record order is mode first (actor mutex), switch_engine second
// (the SetGlobalEngine leg, engine/conn.go) — both records carry engine
// names and mode symbols only (D-20/D-21).

// The closed flip-target vocabulary of the marks oracle.
const (
	flipTargetEN = "en"
	flipTargetRU = "ru"
)

// switchEngineMarkPrefix is the stable prefix of the daemon's switch_engine
// record — the per-target marks append the engine literal and the closing
// quote.
const switchEngineMarkPrefix = `"msg":"switch_engine","engine":"`

// hasSwitchAttemptMark reports whether the journal line carries a
// switch_engine record of either goswitch engine (INFO success or WARN
// refusal — the any-outcome form; see flipEngineMark for the live-latency
// reality that makes the WARN the common record on a real desktop).
func hasSwitchAttemptMark(line string) bool {
	for _, name := range goswitchEngineNames() {
		if strings.Contains(line, switchEngineMarkPrefix+name+`"`) {
			return true
		}
	}

	return false
}

// flipEngineName returns the goswitch engine name a flip target's
// switch_engine record carries; an unknown target is an error — the marks
// never guess.
func flipEngineName(target string) (string, error) {
	switch target {
	case flipTargetEN:
		return spikeEngineEN, nil
	case flipTargetRU:
		return spikeEngineRU, nil
	}

	return "", fmt.Errorf("flip target %q outside {%s, %s}", target, flipTargetEN, flipTargetRU)
}

// flipModeMark returns the daemon's mode-record mark for a flip target.
func flipModeMark(target string) (string, error) {
	if _, err := flipEngineName(target); err != nil {
		return "", err
	}

	return modeRecordMark + `"` + target + `"`, nil
}

// flipEngineMark returns the ANY-OUTCOME mark of the switch_engine record
// for a flip target's engine — the journal line the bound switcher's
// SetGlobalEngine leg writes on success (INFO) AND on a timed-out switch
// (the same prefix under a WARN with a trailing "err" attribute). LIVE
// FINDING (the first 05-05 runs, deterministic): the live engine-creation
// latency is ~41-42 ms, so the daemon's 40 ms switch deadline fires first
// on EVERY live flip — the record is the WARN form while the switch still
// lands, inside the <50 ms budget (the engine-created record trails by
// ~1 ms). The pair oracle therefore pins the ATTEMPT in the pinned order,
// and the `ibus engine` readback in the round arbitrates completion —
// the two together are the observable form of the switching act.
func flipEngineMark(target string) (string, error) {
	engineName, err := flipEngineName(target)
	if err != nil {
		return "", err
	}

	return switchEngineMarkPrefix + engineName + `"`, nil
}

// goswitchEngineNames returns both goswitch engine literals — the closed
// set the matrix driver's additive oracle scans for (engine names are
// config constants, D-20).
func goswitchEngineNames() []string {
	return []string{spikeEngineEN, spikeEngineRU}
}

// engineReactivatedMark is the prefix of the daemon's engine-reactivation
// record (INFO "engine reactivated") — the journal proof that IfOwned
// brought an owned engine back after a (re)registration.
const engineReactivatedMark = `"msg":"engine reactivated","engine":"`

// lastReactivatedEngine scans the daemon log backwards for the newest
// engine-reactivation record and returns the engine name it carried —
// the FACTUAL engine the post-restart IfOwned chose (the empty string
// means no reactivation record yet).
func (s *stand) lastReactivatedEngine() (string, error) {
	lines := s.logLines()
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if !strings.Contains(line, engineReactivatedMark) {
			continue
		}
		rest := line[strings.Index(line, engineReactivatedMark)+len(engineReactivatedMark):]
		end := strings.Index(rest, `"`)
		if end < 0 {
			return "", fmt.Errorf("unparsable reactivation record %q", line)
		}

		return rest[:end], nil
	}

	return "", nil
}

// flipMarksPaired verifies the daemon journal's two-record form of a bus
// flip toward target: the mode record for the target and, STRICTLY AFTER
// it (the offsets of the NEWEST occurrences), the switch_engine record
// naming the target's engine. The daemon writes the pair only for its own
// flips — a sync correction (an external flip the daemon follows) writes
// the mode record alone — so the pair is what makes the switching ACT
// observable in the criterion-6 live cases (D-34 heritage: the daemon's
// journal, never dconf).
func flipMarksPaired(journal, target string) error {
	modeMark, err := flipModeMark(target)
	if err != nil {
		return err
	}
	engineMark, err := flipEngineMark(target)
	if err != nil {
		return err
	}
	iMode := strings.LastIndex(journal, modeMark)
	if iMode < 0 {
		return fmt.Errorf("flip pair (%s): the journal holds no mode record %s", target, modeMark)
	}
	iEngine := strings.LastIndex(journal, engineMark)
	if iEngine < 0 {
		return fmt.Errorf("flip pair (%s): the journal holds no switch_engine record %s", target, engineMark)
	}
	if iEngine < iMode {
		return fmt.Errorf("flip pair (%s): the switch_engine record (offset %d) precedes its mode record (offset %d)"+
			" — the flipTo order is mode first, SetGlobalEngine second", target, iEngine, iMode)
	}

	return nil
}

// modeFollowedOnly verifies the journal form of a SYNC correction toward
// target: the mode record moved to the target while NO switch_engine
// record followed the newest mode record. The daemon's sync listener
// follows the factual engine and never flips the bus in response (the
// single-writer guard, T-05-04-01) — any switch record after the
// correction (WARN refusals included) would be the daemon fighting the
// desktop for the engine truth.
func modeFollowedOnly(journal, target string) error {
	modeMark, err := flipModeMark(target)
	if err != nil {
		return err
	}
	engineMark, err := flipEngineMark(target)
	if err != nil {
		return err
	}
	iMode := strings.LastIndex(journal, modeMark)
	if iMode < 0 {
		return fmt.Errorf("sync follow (%s): the journal holds no mode record %s", target, modeMark)
	}
	if iEngine := strings.LastIndex(journal, engineMark); iEngine > iMode {
		return fmt.Errorf("sync follow (%s): a switch_engine record (offset %d) followed the mode correction"+
			" (offset %d) — the daemon must follow, never flip", target, iEngine, iMode)
	}

	return nil
}

// extFlipSettleWait is the bounded grace the sync case waits after the
// mode correction before asserting ABSENCE (no switch_engine record may
// follow the correction). Absence has no record to wait for, so the
// negative oracle settles a grace instead: were the single-writer guard
// broken, the daemon's illegal flip would write its record within the
// flipTo budget (the 40 ms switch deadline) — well inside the grace.
const extFlipSettleWait = 2 * time.Second

// runTwoSourceFlip proves the phase's central live truth (criterion 6) in
// BOTH directions: inside the 05-01 reversible two-source window a single
// Shift_R tap flips ON THE BUS — the daemon journal carries the two-record
// pair in the pinned order (mode, then switch_engine naming the goswitch
// engine), and the `ibus engine` readback confirms the factual engine
// after every flip. No Super+Space injection anywhere (Pitfall 8): the
// flips ride Shift_R taps in the stand's own focused entry.
func runTwoSourceFlip(ctx context.Context, s *stand) error {
	unit := unitState(ctx)
	fmt.Printf("two-source-flip: unit goswitchd state %q\n", unit)

	if unit == unitActiveState {
		if _, err := runCmd(ctx, "systemctl", "--user", "stop", "goswitchd"); err != nil {
			return fmt.Errorf("two-source-flip: stop unit daemon: %w", err)
		}
	}
	defer func() { s.restoreSpikeWindow(ctx, unit) }()

	if err := waitNameFree(ctx); err != nil {
		return err
	}
	if err := s.startDaemonRegistered(ctx); err != nil {
		return err
	}
	if _, err := s.writeSourcesIfDiffers(ctx, twoSources); err != nil {
		return err
	}

	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("two-source-flip needs the zenity entry surface (locked-session fallback engaged?)")
	}

	// The rounds assume the EN starting side; the pin is journal-gated,
	// never the assumed boot default (the spikePinMode discipline).
	if err := s.spikePinMode(ctx, flipTargetEN); err != nil {
		return fmt.Errorf("two-source-flip en pin: %w", err)
	}
	for _, target := range []string{flipTargetRU, flipTargetEN} {
		if err := s.busFlipRound(ctx, target); err != nil {
			return fmt.Errorf("two-source-flip %s: %w", target, err)
		}
	}

	return nil
}

// busFlipRound drives ONE Shift_R tap through the focused entry and gates
// the bus flip's complete observable form: the NEW mode record, the NEW
// switch_engine record naming the target's engine, the pair ORDER pinned
// by the pure oracle over the whole journal, and the `ibus engine`
// readback confirming the factual engine moved with the flip.
func (s *stand) busFlipRound(ctx context.Context, target string) error {
	modeMark, err := flipModeMark(target)
	if err != nil {
		return err
	}
	engineMark, err := flipEngineMark(target)
	if err != nil {
		return err
	}
	modeBase := s.countSub(modeMark)
	engineBase := s.countSub(engineMark)
	if err := s.injectKeys(ctx, "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, modeMark, modeBase+1, decisionWait); err != nil {
		return fmt.Errorf("mode record: %w", err)
	}
	if err := s.waitForNew(ctx, engineMark, engineBase+1, decisionWait); err != nil {
		return fmt.Errorf("switch_engine record (the SetGlobalEngine leg never landed): %w", err)
	}
	if err := flipMarksPaired(s.logText(), target); err != nil {
		return err
	}
	readback, err := runCmd(ctx, "ibus", "engine")
	if err != nil {
		return fmt.Errorf("engine readback: %w", err)
	}
	want, err := flipEngineName(target)
	if err != nil {
		return err
	}
	if readback != want {
		return fmt.Errorf("bus flip unproven: `ibus engine` readback %q, want %q", readback, want)
	}

	return nil
}

// runExternalFlipSync proves criterion 3 live: an EXTERNAL flip (`ibus
// engine <name>` from a stand subprocess — the simulation of an indicator
// click or another desktop actor) PULLS the daemon along — the daemon's
// mode follows with the byte-stable mode correction record while the
// daemon itself never writes a switch_engine record (it follows, it does
// not flip: the single-writer guard's live face). Both directions run
// inside the reversible two-source window; no Super+Space injection
// (Pitfall 8).
func runExternalFlipSync(ctx context.Context, s *stand) error {
	unit := unitState(ctx)
	fmt.Printf("external-flip-sync: unit goswitchd state %q\n", unit)

	if unit == unitActiveState {
		if _, err := runCmd(ctx, "systemctl", "--user", "stop", "goswitchd"); err != nil {
			return fmt.Errorf("external-flip-sync: stop unit daemon: %w", err)
		}
	}
	defer func() { s.restoreSpikeWindow(ctx, unit) }()

	if err := waitNameFree(ctx); err != nil {
		return err
	}
	if err := s.startDaemonRegistered(ctx); err != nil {
		return err
	}
	if _, err := s.writeSourcesIfDiffers(ctx, twoSources); err != nil {
		return err
	}

	for _, target := range []string{flipTargetRU, flipTargetEN} {
		if err := s.externalSyncRound(ctx, target); err != nil {
			return fmt.Errorf("external-flip-sync %s: %w", target, err)
		}
	}

	return nil
}

// externalSyncRound drives ONE external flip against the live entry and
// gates the follow's complete observable form: the factual engine moved
// (readback), the daemon's mode followed with a NEW mode correction
// record, and after a bounded settle the journal holds NO switch_engine
// record after the correction — the daemon followed, it never flipped.
func (s *stand) externalSyncRound(ctx context.Context, target string) error {
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("external-flip-sync needs the zenity entry surface (locked-session fallback engaged?)")
	}

	engineName, err := flipEngineName(target)
	if err != nil {
		return err
	}
	modeMark, err := flipModeMark(target)
	if err != nil {
		return err
	}
	modeBase := s.countSub(modeMark)
	if _, err := runCmd(ctx, "ibus", "engine", engineName); err != nil {
		return fmt.Errorf("external flip: %w", err)
	}
	readback, err := runCmd(ctx, "ibus", "engine")
	if err != nil {
		return fmt.Errorf("engine readback: %w", err)
	}
	if readback != engineName {
		return fmt.Errorf("external flip unproven: readback %q, want %q", readback, engineName)
	}
	if err := s.waitForNew(ctx, modeMark, modeBase+1, decisionWait); err != nil {
		return fmt.Errorf("the daemon did not follow the external flip with a mode correction: %w", err)
	}
	if err := sleepCtx(ctx, extFlipSettleWait); err != nil {
		return err
	}
	last, lerr := s.lastModeMark()
	if lerr != nil {
		return lerr
	}
	if last != target {
		return fmt.Errorf("mode did not settle on the factual engine: newest mode record %q, want %q", last, target)
	}

	return modeFollowedOnly(s.logText(), target)
}

// requireGoswitchDesktop fails fast when the live sources list carries NO
// goswitch engine — the ibus-restart extension presupposes a
// goswitch-owned desktop of ANY count: with one source the flip mechanism
// still works (SetGlobalEngine targets any registered engine; the
// component registers both engines on the IBus bus regardless of the
// sources list; v1.0.0 ran a single [('ibus','goswitch-en')] source with
// internal flips).
func (s *stand) requireGoswitchDesktop(ctx context.Context) error {
	out, err := runCmd(ctx, "gsettings", "get", gsettingsSchema, keySources)
	if err != nil {
		return fmt.Errorf("goswitch-desktop check: read sources: %w", err)
	}
	for _, engine := range []string{spikeEngineEN, spikeEngineRU} {
		if strings.Contains(out, engine) {
			return nil
		}
	}

	return fmt.Errorf("goswitch-desktop check: sources %q carry no goswitch engine — install first", out)
}

// The flip-keystroke round timing (plan 06-10): the case's essence is the
// letter with NO pause between the mode record and the keystroke — flipTo
// writes the record BEFORE the switcher call, so the letter lands inside
// the switching window, the exact window the pre-fix factory self-block
// stretched into a key swallower (G-6-1). The only pauses ride AFTER the
// letter: the tap FSM's 300 ms window must expire before the next round's
// tap (a tap inside the window is a double-tap correction gesture, not a
// flip) and the AT-SPI bridge needs a beat before the tail readback — the
// probe's 0.35 s + 0.2 s form.
const (
	flipKeystrokeReps = 3 // rounds per direction: 6 daemon flips per run, 12 per double green
	flipKeyFsmWait    = 350 * time.Millisecond
	flipKeySettle     = 200 * time.Millisecond
)

// flipKeystrokeKey is the injected physical key of every round — the
// probe's 'd': EN mode transits it, RU mode commits its ЙЦУКЕН twin.
const flipKeystrokeKey = "d"

// expectedFlipLetter returns the field-tail letter the immediate keystroke
// must leave for a flip toward target — the probe's per-direction
// expectations raised to the pinned oracle: EN mode transits the key, RU
// mode commits the same key position's Cyrillic rune.
func expectedFlipLetter(target string) (string, error) {
	switch target {
	case flipTargetEN:
		return flipKeystrokeKey, nil
	case flipTargetRU:
		return "в", nil
	}

	return "", fmt.Errorf("flip target %q outside {%s, %s}", target, flipTargetEN, flipTargetRU)
}

// switchEngineWarnLine reports whether one journal line is the WARN form of
// the switch_engine record — the deadline-abort shape: the INFO success
// line shares the msg/engine shape and never carries the err member (the
// record form is pinned by the conn_switcher corpus).
func switchEngineWarnLine(line string) bool {
	return strings.Contains(line, `"msg":"switch_engine"`) && strings.Contains(line, `"err"`)
}

// switchEngineWarnCount counts the journal's WARN-form switch_engine
// records so a round can gate on NEW aborts since its own baseline — the
// round-scoped counting that keeps an environmental WARN outside the
// rounds from gating the case (T-06-10-02).
func (s *stand) switchEngineWarnCount() int {
	n := 0
	for _, line := range s.logLines() {
		if switchEngineWarnLine(line) {
			n++
		}
	}

	return n
}

// runFlipKeystroke proves the closed gap G-6-1 LIVE on the wire: after a
// daemon flip (a single Shift_R tap) the FIRST letter injected with no
// pause reaches the field — in both directions, deterministically. The
// mechanics are the owner's /tmp/gsy-settle2.sh probe (5/5 losses pre-fix)
// raised to a stand case: flip through the daemon → wait for the mode
// record (the e2e oracle, Pitfall 5) → inject 'd' IMMEDIATELY → read the
// field tail. The setup follows runTwoSourceFlip's reversible window but
// never rewrites the sources list: the flip is a SetGlobalEngine and works
// with any registered engine (ADR-006 Amendment 2026-09-30) — the probe
// ran on the owner's own single-source desk.
func runFlipKeystroke(ctx context.Context, s *stand) error {
	unit := unitState(ctx)
	fmt.Printf("flip-keystroke: unit goswitchd state %q\n", unit)

	if unit == unitActiveState {
		if _, err := runCmd(ctx, "systemctl", "--user", "stop", "goswitchd"); err != nil {
			return fmt.Errorf("flip-keystroke: stop unit daemon: %w", err)
		}
	}
	defer func() { s.restoreSpikeWindow(ctx, unit) }()

	if err := waitNameFree(ctx); err != nil {
		return err
	}
	if err := s.startDaemonRegistered(ctx); err != nil {
		return err
	}

	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errors.New("flip-keystroke needs the zenity entry surface (locked-session fallback engaged?)")
	}

	// The rounds need a goswitch engine GLOBAL: the unit stop leaves the
	// global engine unset (ibus-daemon unsets it when an engine connection
	// closes), and taps route nowhere until one is activated — the probe's
	// own precondition ("поле в фокусе, движок: …"; the owner's session
	// always carried one). The activation is the external SetGlobalEngine
	// of the word cases, and the sync correction it provokes doubles as
	// the journal's first mode record.
	if err := s.activateGoswitch(ctx); err != nil {
		return fmt.Errorf("flip-keystroke engine activation: %w", err)
	}
	// The mode pin stays journal-gated (spikePinMode discipline): the
	// activation's own correction record usually settles it with zero taps.
	if err := s.spikePinMode(ctx, flipTargetEN); err != nil {
		return fmt.Errorf("flip-keystroke en pin: %w", err)
	}

	for rep := range flipKeystrokeReps {
		// The ru→en direction rides the pair's second flip: after the
		// en→ru round the journal-pinned mode is RU, so the next tap
		// returns EN (the pin is read from the journal, never assumed).
		if err := s.flipKeystrokeRound(ctx, flipTargetRU); err != nil {
			return fmt.Errorf("flip-keystroke rep %d en→ru: %w", rep+1, err)
		}
		if err := s.flipKeystrokeRound(ctx, flipTargetEN); err != nil {
			return fmt.Errorf("flip-keystroke rep %d ru→en: %w", rep+1, err)
		}
		// The busFlipRound form: after every pair the `ibus engine`
		// readback confirms the FACTUAL engine moved with the last flip.
		readback, err := runCmd(ctx, "ibus", "engine")
		if err != nil {
			return fmt.Errorf("flip-keystroke rep %d engine readback: %w", rep+1, err)
		}
		want, err := flipEngineName(flipTargetEN)
		if err != nil {
			return err
		}
		if readback != want {
			return fmt.Errorf("flip-keystroke rep %d: `ibus engine` readback %q after the pair, want %q",
				rep+1, readback, want)
		}
		fmt.Printf("flip-keystroke: rep %d — both letters landed, factual engine %q\n", rep+1, readback)
	}

	return nil
}

// flipKeystrokeRound drives ONE daemon flip plus the immediate letter and
// gates the round's full observable form: the NEW mode record (the e2e
// oracle the keystroke waits on — Pitfall 5), the letter injected with NO
// pause after that record, the field tail settling to the direction's
// letter, the flipMarksPaired order over the whole journal, and ZERO new
// deadline-abort switch_engine WARN records since the round's baseline —
// a WARN here means the factory self-block is back (or the bus wedged;
// the case fails with the journal excerpt and the stand's restart
// discipline applies).
func (s *stand) flipKeystrokeRound(ctx context.Context, target string) error {
	last, err := s.lastModeMark()
	if err != nil {
		return err
	}
	if last == target {
		return fmt.Errorf("flip-keystroke %s: the journal's newest mode record is already %q —"+
			" the tap would not flip", target, target)
	}
	modeMark, err := flipModeMark(target)
	if err != nil {
		return err
	}
	letter, err := expectedFlipLetter(target)
	if err != nil {
		return err
	}
	modeBase := s.countSub(modeMark)
	warnBase := s.switchEngineWarnCount()

	// The daemon flip: a single Shift_R tap (SWCH-01's physical gesture).
	if err := s.injectKeys(ctx, "Shift_R"); err != nil {
		return err
	}
	// The mode record is the keystroke's starting gun: flipTo writes it
	// BEFORE the switcher call, so from here the letter races the
	// switching window on purpose.
	if err := s.waitForNew(ctx, modeMark, modeBase+1, decisionWait); err != nil {
		return fmt.Errorf("mode record: %w", err)
	}
	// THE letter — injected with NO pause (T-06-10-01): any settle here
	// would weaken the oracle into every existing flip case's shape.
	if err := s.injectKeys(ctx, flipKeystrokeKey); err != nil {
		return err
	}
	// The only pauses ride after the letter (see the timing constants).
	if err := sleepCtx(ctx, flipKeyFsmWait); err != nil {
		return err
	}
	if err := sleepCtx(ctx, flipKeySettle); err != nil {
		return err
	}
	if err := s.waitZenityTail(ctx, letter); err != nil {
		return fmt.Errorf("flip-keystroke %s: immediate letter lost: %w", target, err)
	}
	fmt.Printf("flip-keystroke: %s — immediate %q at the field tail\n", target, letter)
	if err := flipMarksPaired(s.logText(), target); err != nil {
		return err
	}
	if warns := s.switchEngineWarnCount(); warns != warnBase {
		return fmt.Errorf("flip-keystroke %s: %d new deadline-abort switch_engine WARN record(s) in the round"+
			" (base %d, now %d) — the factory self-block is back or the bus wedged",
			target, warns-warnBase, warnBase, warns)
	}

	return nil
}

// waitZenityTail polls the content-exact readback until the focused entry's
// text ends with want — the tail oracle over the accumulating field (each
// round appends its letter and the previous round's letter is always the
// other one, so a stale tail can never satisfy the poll). The AT-SPI bridge
// can lag the delivery by a beat; the poll rides the lag out instead of a
// one-shot read (the waitZenityText discipline).
func (s *stand) waitZenityTail(ctx context.Context, want string) error {
	deadline := time.Now().Add(witnessWait)
	var last string
	for {
		out, err := s.readFocusedTextRaw(ctx)
		if err != nil {
			return err
		}
		last = out
		if strings.HasSuffix(out, want) {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("field tail never became %q (readback %q)", want, last)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}
