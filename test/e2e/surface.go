package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Chromium-surface constants (plan 02-02). The binary is pinned, not probed
// per run: on the target machine google-chrome (deb 153.x) is the matrix
// Chromium surface and snap chromium the fallback (research A3); swapping
// binaries is a one-constant change here.
const (
	// chromiumBin is the Chromium-class binary the stand launches.
	chromiumBin = "google-chrome"
	// chromiumAppName is the AT-SPI application name the pinned binary
	// exposes (live-verified 2026-09-12; snap chromium names itself
	// differently — re-pin both constants together if the binary changes).
	chromiumAppName = "Google Chrome"
	// chromiumX11AppName is the AT-SPI application name of the SAME binary
	// in the --ozone-platform=x11 window mode (04-05 spike, live-verified
	// 2026-09-16: the ATK bridge over XWayland names the app identically —
	// re-pin all three constants together if the binary changes).
	chromiumX11AppName = "Google Chrome"
)

// fixtureInputPath is the stand's input page for Chromium-surface cases.
const fixtureInputPath = "test/e2e/fixtures/input.html"

// gteBin is the GtkSourceView multiline surface of the case matrix — the
// verbatim addressee of the phase-2 success criterion 1.
const gteBin = "gnome-text-editor"

// gteAppName is the AT-SPI application name the editor exposes
// (live-verified 2026-09-14; the name matches the binary).
const gteAppName = "gnome-text-editor"

// geditBin is the GTK3-generation editor surface of matrix v3 (plan 04-05,
// D-47): its IM/AT-SPI stack is a genuinely different generation from the
// GTK4 gnome-text-editor (libgedit-gtksourceview-300 — the GTK3 fork), the
// very difference the roadmap's criterion 2 adds the surface for.
const geditBin = "gedit"

// geditAppName is the AT-SPI application name the editor exposes
// (04-05 spike, live-verified 2026-09-16: the GTK3 bridge names the app
// after the binary — witness line "gedit:TEXT:chars=N"; the pid-keyed
// gates stay the instance-exact pair since the name is shared with any
// concurrently running owner instance).
const geditAppName = "gedit"

// surfaceFocusWait budgets how long a spawned surface may take from process
// start to a focused, empty input node in the AT-SPI tree: a cold Chromium
// start needs several seconds before its renderer accessibility tree is up.
const surfaceFocusWait = 15 * time.Second

// The surfaceKind enumeration of case_m1.go extended from this file so
// that case_m1.go itself stays untouched.
const (
	// surfaceChromium is the stand-spawned fresh Chromium instance.
	surfaceChromium surfaceKind = iota + 2
	// surfaceGnomeTextEditor is the stand-spawned standalone
	// gnome-text-editor document window.
	surfaceGnomeTextEditor
	// surfaceGedit is the stand-spawned standalone gedit window (04-05).
	surfaceGedit
	// surfaceChromiumX11 is the stand-spawned fresh Chromium instance in
	// the x11/XWayland window mode (04-05).
	surfaceChromiumX11
)

// startChromium spawns a fresh Chromium instance on the stand's fixture
// page. Two flag groups are load-bearing:
//   - --user-data-dir inside the stand's temp dir keeps the instance off
//     the owner's profile AND makes it a separate process — without a
//     distinct data dir the new process hands its URL to the owner's
//     running instance, so PID-based closing would be impossible;
//   - --force-renderer-accessibility: a fresh instance leaves its AT-SPI
//     tree empty without it even while an AT client polls (live-verified
//     2026-09-12), so the witness would never see the page input.
func (s *stand) startChromium(ctx context.Context) error {
	url, err := fixtureInputURL()
	if err != nil {
		return err
	}
	profile := filepath.Join(s.tmpDir, "chromium-profile")
	cmd := exec.CommandContext(ctx, chromiumBin,
		"--new-window",
		"--user-data-dir="+profile,
		"--no-first-run",
		"--no-default-browser-check",
		"--force-renderer-accessibility",
		url,
	)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", chromiumBin, err)
	}
	s.chromium = cmd

	return nil
}

// waitChromiumInput polls until the page input of the spawned instance is
// the focused input surface with exactly wantChars characters — the gate
// that makes injection and readback safe. Called with 0 before injection
// (empty field) and with the probe-word length after it.
func (s *stand) waitChromiumInput(ctx context.Context, wantChars int) error {
	return s.waitInputFocus(ctx, chromiumAppName, wantChars, surfaceFocusWait)
}

// readChromiumText reads the page input's text through the AT-SPI helper.
// The readback is focused-node based, not app-name based: the desktop's
// other "Google Chrome" instances answer to the same app name
// (live-verified 2026-09-12), so `text <app>` would read the owner's
// browser instead of the stand's page input.
func (s *stand) readChromiumText(ctx context.Context) (string, error) {
	return s.readFocusedText(ctx)
}

// closeChromium terminates the instance by killing its own PID — the
// fresh-profile instance has nothing to save, so there is no graceful
// path. Doubles as the reap for teardown and error paths; a no-op once
// closed.
func (s *stand) closeChromium() {
	if s.chromium == nil {
		return
	}
	if s.chromium.Process != nil {
		_ = s.chromium.Process.Kill()
	}
	_ = s.chromium.Wait()
	s.chromium = nil
}

// startChromiumX11 spawns a fresh Chromium instance in the x11/XWayland
// window mode (plan 04-05, D-47) on the stand's fixture page. The flag set
// is startChromium's verbatim plus exactly one mode flag — every existing
// flag stays load-bearing for the same reasons (fresh profile → separate
// process and PID-closeable; renderer accessibility → live AT-SPI tree):
//
//   - --ozone-platform=x11 forces the XWayland surface: the desktop's
//     Chrome 153 defaults to native Wayland (04-05 research, /proc fd
//     probe), so the x11 IM path (XIM/XWayland bridge) is a DIFFERENT
//     surface tier, not a re-run of the Wayland one.
//
// 04-05 spike verdicts (live 2026-09-16, the x11-smoke case + probes):
// the instance is born as its own process tree, its AT-SPI app name is
// still "Google Chrome" (the pid-keyed gates are the instance-exact
// discriminators), and injected keys reach the fixture page input. The IM
// verdict is NEGATIVE and the matrix composes it as such: the x11 input
// context declares caps 0x29 but NEVER pushes surrounding text (zero
// pushes during typing; the GTK_IM_MODULE=ibus probe changed nothing), so
// the correction ladder never executes there (pre-correction verify
// expires — "correction skipped, verify-timeout"); in RU mode the engine
// commits land in the OMNIBOX (the URL bar — a 63-char witness), so the
// surface composes as EN-mode transit rows only (see matrix-v3.yaml).
func (s *stand) startChromiumX11(ctx context.Context) error {
	url, err := fixtureInputURL()
	if err != nil {
		return err
	}
	profile := filepath.Join(s.tmpDir, "chromium-x11-profile")
	cmd := exec.CommandContext(ctx, chromiumBin,
		"--new-window",
		"--user-data-dir="+profile,
		"--no-first-run",
		"--no-default-browser-check",
		"--force-renderer-accessibility",
		"--ozone-platform=x11",
		url,
	)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s (x11): %w", chromiumBin, err)
	}
	s.chromiumX11 = cmd

	return nil
}

// waitChromiumX11Input polls until the x11-mode instance's page input is
// the focused input surface with exactly wantChars characters. The gate is
// PID-keyed — not app-name keyed like the Wayland surface's: the x11
// instance shares "Google Chrome" with every owner instance too (04-05
// spike), and the pid is the only instance-exact discriminator.
func (s *stand) waitChromiumX11Input(ctx context.Context, wantChars int) error {
	if s.chromiumX11 == nil || s.chromiumX11.Process == nil {
		return errors.New("no chromium-x11 instance to witness")
	}

	return s.awaitPidInput(ctx, chromiumX11AppName+" (x11)", s.chromiumX11.Process.Pid, wantChars)
}

// readChromiumX11Text reads the x11-mode instance's page input through the
// pid-keyed focused-text bridge — instance-exact, pairing with
// waitChromiumX11Input (the app name is shared with the owner's browser).
func (s *stand) readChromiumX11Text(ctx context.Context) (string, error) {
	if s.chromiumX11 == nil || s.chromiumX11.Process == nil {
		return "", errors.New("no chromium-x11 instance to read")
	}

	return s.readFocusedTextPid(ctx, s.chromiumX11.Process.Pid)
}

// closeChromiumX11 terminates the x11-mode instance by killing its own
// PID — the fresh-profile instance has nothing to save. Doubles as the
// reap for teardown and error paths; a no-op once closed.
func (s *stand) closeChromiumX11() {
	if s.chromiumX11 == nil {
		return
	}
	if s.chromiumX11.Process != nil {
		_ = s.chromiumX11.Process.Kill()
	}
	_ = s.chromiumX11.Wait()
	s.chromiumX11 = nil
}

// startGTE spawns a standalone gnome-text-editor with a new unsaved
// document. Three live-verified facts shape the spawn (2026-09-14):
//   - gnome-text-editor is single-instance per session bus: without
//     --standalone a spawn into a running owner instance delegates and
//     exits, leaving the window in a PID the stand cannot kill;
//   - the 46.3 build restores the owner's previous session (windows from
//     ~/.local/share/org.gnome.TextEditor/session.gvariant) even with
//     --ignore-session — the REAL isolation is XDG_DATA_HOME: with no
//     session file to read, nothing is restored (--ignore-session stays
//     as the second line of defense for builds that honor it);
//   - the editor autosaves unsaved documents as drafts (3 s) into
//     XDG_DATA_HOME — the isolated dir keeps the probe word out of the
//     owner's draft store; it dies with the stand's temp dir.
func (s *stand) startGTE(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, gteBin, "--standalone", "--ignore-session")
	cmd.Env = withEnv("XDG_DATA_HOME", filepath.Join(s.tmpDir, "gte-data"))
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", gteBin, err)
	}
	s.gte = cmd

	return nil
}

// GTE witness/recovery timing: the pid-keyed poll first watches whether
// the new window takes focus on its own; past the grace the loop pokes
// the editor with the pid-keyed input-node grabFocus every grabEvery
// until the deadline — each poke re-issues the activation request with a
// timestamp that outranks the current focus-stealing denial.
const (
	gteGrabGrace = 2 * time.Second
	gteGrabEvery = 3 * time.Second
)

// GTE readback retry: bounded re-reads for transient a11y walk misses.
const (
	gteReadAttempts   = 3
	gteReadRetryDelay = 250 * time.Millisecond
)

// waitGTEInput polls until the standalone editor's document — identified
// by the spawned instance's PID, because the AT-SPI name is shared with
// any concurrently running owner instance — is the focused input surface
// with exactly wantChars characters.
//
// The pid-keyed loop is awaitPidInput; its grab-poke recovery path exists
// for the GTE window specifically (live-verified 2026-09-14): mutter's
// focus-stealing prevention denies focus to the tokenless editor window
// whenever a real input event — even pointer motion — follows its map
// request, so the window may NEVER take focus on its own.
func (s *stand) waitGTEInput(ctx context.Context, wantChars int) error {
	if s.gte == nil || s.gte.Process == nil {
		return errors.New("no gnome-text-editor instance to witness")
	}

	return s.awaitPidInput(ctx, gteAppName, s.gte.Process.Pid, wantChars)
}

// readGTEText reads the focused document's text through the helper's
// pid-keyed focused-text bridge — instance-exact, pairing with
// waitGTEInput (readFocusedTextPid owns the bounded re-reads).
func (s *stand) readGTEText(ctx context.Context) (string, error) {
	if s.gte == nil || s.gte.Process == nil {
		return "", errors.New("no gnome-text-editor instance to read")
	}

	return s.readFocusedTextPid(ctx, s.gte.Process.Pid)
}

// closeGTE terminates the editor by killing its own PID — the unsaved
// document dies without any save dialog under SIGKILL, and its draft
// autosave lives in the stand's temp dir. Doubles as the reap for
// teardown and error paths; a no-op once closed.
func (s *stand) closeGTE() {
	if s.gte == nil {
		return
	}
	if s.gte.Process != nil {
		_ = s.gte.Process.Kill()
	}
	_ = s.gte.Wait()
	s.gte = nil
}

// startGedit spawns a standalone gedit with a new empty document (plan
// 04-05, D-47). The 04-05 spike (live-verified 2026-09-16) pinned the
// isolation set — the GTE treatment transfers almost verbatim:
//   - gedit 46.2 HAS -s/--standalone (the plan's "no --standalone" was
//     wrong — research A4 resolved by probe): without it a spawn into a
//     running owner instance delegates and exits, leaving the window in a
//     PID the stand cannot kill;
//   - XDG_DATA_HOME redirects gedit's state (recent files, session data)
//     into the stand's temp dir — the probe word never reaches the owner's
//     stores and dies with the stand;
//   - XDG_CONFIG_HOME must join the redirection: gedit's spell-check
//     (enchant) writes user dictionaries into it — the spike caught the
//     en_US.dic/en_US.exc files materializing in the redirected dir (the
//     owner's real dictionary stays untouched).
//
// Behavior verdicts of the same spike: EN-mode typing transits and pushes
// surrounding text; the word correction runs the level-1 ladder live
// ("correction","level":1 — caps 0x29, delete+commit applied and read
// back); a ctrl+a never produces an anchor push (cursor == anchor in every
// push — the zenity-class D-30 degradation); in RU mode the engine-consumed
// keys' commits DO NOT render (the document stays empty, caps drop to 0x9)
// — the matrix composes gedit as EN-mode rows only (matrix-v3.yaml header).
func (s *stand) startGedit(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, geditBin, "--standalone")
	cmd.Env = withEnvList(os.Environ(),
		envPair{"XDG_DATA_HOME", filepath.Join(s.tmpDir, "gedit-data")},
		envPair{"XDG_CONFIG_HOME", filepath.Join(s.tmpDir, "gedit-config")},
	)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", geditBin, err)
	}
	s.gedit = cmd

	return nil
}

// waitGeditInput polls until the standalone editor's document — identified
// by the spawned instance's PID, because the AT-SPI name is shared with
// any concurrently running owner instance — is the focused input surface
// with exactly wantChars characters (awaitPidInput; the gedit window takes
// focus on its own map — 04-05 spike — and the poke path is the same
// recovery the GTE window needs).
func (s *stand) waitGeditInput(ctx context.Context, wantChars int) error {
	if s.gedit == nil || s.gedit.Process == nil {
		return errors.New("no gedit instance to witness")
	}

	return s.awaitPidInput(ctx, geditAppName, s.gedit.Process.Pid, wantChars)
}

// readGeditText reads the focused document's text through the helper's
// pid-keyed focused-text bridge — instance-exact, pairing with
// waitGeditInput (readFocusedTextPid owns the bounded re-reads).
func (s *stand) readGeditText(ctx context.Context) (string, error) {
	if s.gedit == nil || s.gedit.Process == nil {
		return "", errors.New("no gedit instance to read")
	}

	return s.readFocusedTextPid(ctx, s.gedit.Process.Pid)
}

// closeGedit terminates the editor by killing its own PID — the unsaved
// document dies without any save dialog under SIGKILL, and any state it
// keeps lives in the stand's redirected XDG dirs. Doubles as the reap for
// teardown and error paths; a no-op once closed.
func (s *stand) closeGedit() {
	if s.gedit == nil {
		return
	}
	if s.gedit.Process != nil {
		_ = s.gedit.Process.Kill()
	}
	_ = s.gedit.Wait()
	s.gedit = nil
}

// withEnv returns the current process environment with key overridden to
// value — later duplicates of key removed, so glibc's first-match getenv
// cannot resurrect a stale assignment in the spawned child.
func withEnv(key, value string) []string {
	return withEnvList(os.Environ(), envPair{key, value})
}

// envPair is one key=value override for withEnvList.
type envPair struct {
	key   string
	value string
}

// withEnvList returns env with every pair applied — each key's later
// duplicates removed, so overrides chain (the gedit driver needs two XDG
// redirections in one child; calling the single-pair form twice would
// restart from os.Environ() and lose the first override).
func withEnvList(env []string, pairs ...envPair) []string {
	for _, p := range pairs {
		prefix := p.key + "="
		next := make([]string, 0, len(env)+1)
		for _, kv := range env {
			if !strings.HasPrefix(kv, prefix) {
				next = append(next, kv)
			}
		}
		next = append(next, prefix+p.value)
		env = next
	}

	return env
}

// fixtureInputURL resolves the fixture page to an absolute file:// URL.
func fixtureInputURL() (string, error) {
	abs, err := filepath.Abs(fixtureInputPath)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", fixtureInputPath, err)
	}

	return "file://" + abs, nil
}

// waitInputFocus polls the AT-SPI helper until exactly the named app's
// input is the focused input surface with wantChars characters — and no
// shell entry is. AT-SPI focus state can lag the compositor: a shell
// overlay (overview search and friends surface as PASSWORD_TEXT) may hold
// REAL keyboard focus while the case's freshly opened window still reports
// its node focused (live-verified 2026-09-12 — an injection under that
// lag typed into the shell's entry). The gate therefore enumerates every
// focused candidate and refuses while the shell's own entry is among
// them. The chars check matters for apps whose AT-SPI name the owner's
// own instances share (Google Chrome): the stand's surfaces always start
// empty and carry exactly the probe word after injection.
func (s *stand) waitInputFocus(ctx context.Context, app string, wantChars int, timeout time.Duration) error {
	prefix := app + ":"
	want := ":chars=" + strconv.Itoa(wantChars)
	deadline := time.Now().Add(timeout)
	var last string
	for {
		out, err := runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "focused-inputs")
		if err != nil {
			return fmt.Errorf("focused-inputs gate: %w", err)
		}
		last = out
		found := false
		for _, line := range strings.Split(out, "\n") {
			switch {
			case line == "":
				continue
			case strings.HasPrefix(line, shellAppPrefix):
				return fmt.Errorf("shell entry holds focus while waiting for %s (focused %q) —"+
					" desktop overlay state, rerun on an idle desktop", app, out)
			case strings.HasPrefix(line, prefix) && strings.HasSuffix(line, want):
				found = true
			}
		}
		if found {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s input not focused with %s (focused %q)", app, want, last)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// readFocusedText reads the text of the currently focused input surface
// through the helper's focused-text bridge (see readChromiumText for why
// the focused node, not the app name, is the key).
func (s *stand) readFocusedText(ctx context.Context) (string, error) {
	out, err := runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "focused-text")
	if err != nil {
		return "", fmt.Errorf("focused-text readback: %w", err)
	}

	return out, nil
}

// awaitPidInput is the shared witness loop of the pid-keyed surface
// drivers (GTE, gedit, chromium-x11): it polls until the application with
// the given pid owns the focused input surface with exactly wantChars
// characters.
//
// Recovery path (live-verified 2026-09-14 on GTE): mutter's
// focus-stealing prevention can deny focus to a freshly mapped window
// whenever a real input event follows its map request. Past the grace
// the loop pokes the pid-keyed input node with grabFocus — the AT-SPI
// call itself reports refused, but the widget grab re-issues the window
// activation with a timestamp that outranks the denial; the loop keeps
// poking until the witness clears or the surfaceFocusWait budget runs
// out. Each poke is a best-effort no-op while the app tree is not up.
func (s *stand) awaitPidInput(ctx context.Context, app string, pid, wantChars int) error {
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
			return fmt.Errorf("%s (pid %s) input not focused with %s (witness %q)",
				app, pidStr, want, last)
		}
		if time.Now().After(nextGrab) {
			_, _ = runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "grab-input-pid", pidStr)
			nextGrab = time.Now().Add(gteGrabEvery)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// readFocusedTextPid reads the text of the pid's focused input surface
// through the helper's pid-keyed focused-text bridge — the instance-exact
// readback paired with awaitPidInput. The read is retried a few times: a
// single a11y walk can transiently miss the node while the witness
// polling keeps the bus busy (live-verified 2026-09-14: witness passed,
// one-shot read found no focused surface).
func (s *stand) readFocusedTextPid(ctx context.Context, pid int) (string, error) {
	pidStr := strconv.Itoa(pid)
	var lastErr error
	for range gteReadAttempts {
		out, err := runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "focused-text-pid", pidStr)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if err := sleepCtx(ctx, gteReadRetryDelay); err != nil {
			return "", err
		}
	}

	return "", fmt.Errorf("focused-text-pid readback: %w", lastErr)
}
