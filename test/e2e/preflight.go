package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"syscall"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/Djarvur/goswitch/engine"
)

// wOK is the POSIX W_OK access mode for syscall.Access (x/sys is outside
// the dependency budget).
const wOK = 0x2

// registrationWait is how long the daemon may take from process start to
// the component-registered log record.
const registrationWait = 10 * time.Second

// preflight runs the fail-fast environment checks (TEST-02). Each check
// carries a named diagnostic so a missing prerequisite prints one line a
// human can act on before the stand exits 1. Standalone cases (no stand
// daemon — install-cycle) run the environment core only: the daemon-bound
// and surface-driver checks presuppose the stand's own daemon/surfaces.
func preflight(ctx context.Context, s *stand) error {
	type check struct {
		name string
		run  func(context.Context, *stand) error
	}
	checks := []check{
		{"injection-selftest", checkInjectionSelfTest},
		{"ibus-address", checkIbusAddress},
		{"uinput-writable", checkUinputWritable},
		{"python-gi", checkPythonGI},
	}
	if !s.standalone {
		checks = append(checks,
			check{"engine-registered", checkEngineRegistered},
			check{"daemon-log-heartbeat", checkLogHeartbeat},
			check{"chromium-launch", checkChromiumLaunch},
			check{"gnome-text-editor-launch", checkGnomeTextEditorLaunch},
			check{"gedit-launch", checkGeditLaunch},
		)
	}
	for _, c := range checks {
		if err := c.run(ctx, s); err != nil {
			return fmt.Errorf("preflight %s: %w", c.name, err)
		}
		fmt.Printf("preflight %s: ok\n", c.name)
	}

	return nil
}

// checkInjectionSelfTest proves the injector path early (research A4:
// ydotool 0.1.8's binary and permissions were verified, its injection was
// not — until here). A bare Shift_R tap is inert in any window, and at this
// point goswitch is not the active source yet, so the tap stays out of the
// daemon log too.
func checkInjectionSelfTest(ctx context.Context, s *stand) error {
	delay := strconv.Itoa(s.cfg.pacing)
	if _, err := runCmd(ctx, "ydotool", "key", "--key-delay", delay, "Shift_R"); err != nil {
		return fmt.Errorf("%w (is /dev/uinput writable and ydotool 0.1.8 installed?)", err)
	}

	return nil
}

// checkIbusAddress proves an IBus session is reachable.
func checkIbusAddress(ctx context.Context, _ *stand) error {
	addr, err := runCmd(ctx, "ibus", "address")
	if err != nil {
		return fmt.Errorf("%w (is an IBus/GNOME session running?)", err)
	}
	if addr == "" {
		return errors.New("ibus address is empty (is an IBus/GNOME session running?)")
	}

	return nil
}

// checkUinputWritable mirrors `test -w /dev/uinput`: the injector device
// needs the input group.
func checkUinputWritable(_ context.Context, _ *stand) error {
	if err := syscall.Access("/dev/uinput", wOK); err != nil {
		return fmt.Errorf("/dev/uinput is not writable: %w (add the user to the input group and re-login)", err)
	}

	return nil
}

// checkPythonGI proves the AT-SPI helper's interpreter works: PATH python3
// is linuxbrew without gi, /usr/bin/python3 is mandatory (01-PATTERNS).
func checkPythonGI(ctx context.Context, _ *stand) error {
	if _, err := runCmd(ctx, "/usr/bin/python3", "-c", "import gi"); err != nil {
		return fmt.Errorf("%w (install python3-gi and gir1.2-atspi-2.0)", err)
	}

	return nil
}

// checkEngineRegistered proves the component is registered daemon-side.
// `ibus list-engine` cannot see programmatic registrations on IBus 1.5.29 —
// it reads the XML registry only (01-01 live finding) — so this check waits
// for the registration log line and then queries ListActiveEngines on the
// private bus.
func checkEngineRegistered(ctx context.Context, s *stand) error {
	if err := s.waitForLog(ctx, componentRegisteredMark, registrationWait); err != nil {
		return fmt.Errorf("%w (did the daemon reach RegisterComponent?)", err)
	}
	found, err := listActiveEnginesContain(ctx, "goswitch-en")
	if err != nil {
		return err
	}
	if !found {
		return errors.New("goswitch-en is not in the daemon-side ListActiveEngines reply")
	}

	return nil
}

// checkLogHeartbeat proves the daemon log is alive: both startup records
// arrived, so the file is the assertion surface the cases grep.
func checkLogHeartbeat(_ context.Context, s *stand) error {
	for _, want := range []string{`"msg":"connected"`, componentRegisteredMark} {
		if s.countSub(want) == 0 {
			return fmt.Errorf("daemon log misses startup record %q", want)
		}
	}

	return nil
}

// checkChromiumLaunch proves the Chromium surface driver's happy path: the
// pinned binary exists, a fresh instance opens the fixture page, the page
// input takes focus per the witness, and the PID close cleans up. One
// actionable diagnostic line when the binary is missing (plan 02-02).
func checkChromiumLaunch(ctx context.Context, s *stand) error {
	if _, err := exec.LookPath(chromiumBin); err != nil {
		return fmt.Errorf("%s not in PATH: %w (install google-chrome or chromium)", chromiumBin, err)
	}
	if err := s.startChromium(ctx); err != nil {
		return err
	}
	if err := s.waitChromiumInput(ctx, 0); err != nil {
		s.closeChromium()

		return fmt.Errorf("%w (does the fixture window open and take focus?)", err)
	}
	s.closeChromium()

	return nil
}

// checkGnomeTextEditorLaunch proves the GTE surface driver's happy path:
// the binary exists, the standalone instance with its isolated data dir
// opens a new empty document, the witness sees the editor focused, and
// the PID close cleans up. One actionable diagnostic line when the binary
// is missing (plan 02-02).
func checkGnomeTextEditorLaunch(ctx context.Context, s *stand) error {
	if _, err := exec.LookPath(gteBin); err != nil {
		return fmt.Errorf("%s not in PATH: %w (install gnome-text-editor)", gteBin, err)
	}
	if err := s.startGTE(ctx); err != nil {
		return err
	}
	if err := s.waitGTEInput(ctx, 0); err != nil {
		s.closeGTE()

		return fmt.Errorf("%w (does the new-document window open and take focus?)", err)
	}
	s.closeGTE()

	return nil
}

// checkGeditLaunch proves the gedit surface driver's happy path (plan
// 04-05): the binary exists, the standalone instance with its isolated
// XDG dirs opens an empty document, the witness sees the editor focused,
// and the PID close cleans up. One actionable diagnostic line when the
// binary is missing — gedit is NOT part of the default Ubuntu GNOME
// install, so its absence is the expected failure this check exists to
// name (the research Fall-6 warning-sign detector).
func checkGeditLaunch(ctx context.Context, s *stand) error {
	if _, err := exec.LookPath(geditBin); err != nil {
		return fmt.Errorf("%s not in PATH: %w (sudo apt install gedit — universe repo)", geditBin, err)
	}
	if err := s.startGedit(ctx); err != nil {
		return err
	}
	if err := s.waitGeditInput(ctx, 0); err != nil {
		s.closeGedit()

		return fmt.Errorf("%w (does the new-document window open and take focus?)", err)
	}
	s.closeGedit()

	return nil
}

// listActiveEnginesContain dials the private IBus bus and reports whether
// the ListActiveEngines reply mentions name anywhere in its variant tree.
func listActiveEnginesContain(ctx context.Context, name string) (bool, error) {
	addr, err := engine.Discover()
	if err != nil {
		return false, fmt.Errorf("discover ibus address: %w", err)
	}
	conn, err := dbus.Dial(addr)
	if err != nil {
		return false, fmt.Errorf("dial ibus bus: %w", err)
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "e2e: close ibus probe connection: %v\n", cerr)
		}
	}()
	if err := conn.Auth([]dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}); err != nil {
		return false, fmt.Errorf("dbus auth: %w", err)
	}
	if err := conn.Hello(); err != nil {
		return false, fmt.Errorf("dbus hello: %w", err)
	}
	call := conn.Object("org.freedesktop.IBus", "/org/freedesktop/IBus").
		CallWithContext(ctx, "org.freedesktop.IBus.ListActiveEngines", 0)
	if call.Err != nil {
		return false, fmt.Errorf("ListActiveEngines: %w", call.Err)
	}

	return bodyMentions(call.Body, name), nil
}

// bodyMentions walks a D-Bus reply body looking for one string: engine
// descs are nested variants, and matching the string anywhere keeps the
// check independent of the exact wire shape.
func bodyMentions(body []any, want string) bool {
	var walk func(v any) bool
	walk = func(v any) bool {
		switch t := v.(type) {
		case string:
			return t == want
		case dbus.Variant:
			return walk(t.Value())
		case []any:
			return slices.ContainsFunc(t, walk)
		case []dbus.Variant:
			return slices.ContainsFunc(t, func(item dbus.Variant) bool { return walk(item) })
		}

		return false
	}

	return slices.ContainsFunc(body, walk)
}
