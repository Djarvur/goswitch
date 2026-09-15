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
)

// fixtureInputPath is the stand's input page for Chromium-surface cases.
const fixtureInputPath = "test/e2e/fixtures/input.html"

// gteBin is the GtkSourceView multiline surface of the case matrix — the
// verbatim addressee of the phase-2 success criterion 1.
const gteBin = "gnome-text-editor"

// gteAppName is the AT-SPI application name the editor exposes
// (live-verified 2026-09-14; the name matches the binary).
const gteAppName = "gnome-text-editor"

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
// Recovery path (live-verified 2026-09-14): mutter's focus-stealing
// prevention denies focus to the tokenless editor window whenever a real
// input event — even pointer motion — follows its map request, so the
// window may NEVER take focus on its own. The helper's pid-keyed
// grabFocus on the editor's input node makes GTK4 re-issue the window
// activation as a side effect of the widget grab (the AT-SPI call itself
// reports refused — the grab still works), and the loop keeps poking
// until the witness clears or the surfaceFocusWait budget runs out.
func (s *stand) waitGTEInput(ctx context.Context, wantChars int) error {
	if s.gte == nil || s.gte.Process == nil {
		return errors.New("no gnome-text-editor instance to witness")
	}
	pid := strconv.Itoa(s.gte.Process.Pid)
	want := ":chars=" + strconv.Itoa(wantChars)
	deadline := time.Now().Add(surfaceFocusWait)
	nextGrab := time.Now().Add(gteGrabGrace)
	var last string
	for {
		out, err := runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "focused-input-pid", pid)
		if err != nil {
			return fmt.Errorf("focused-input-pid gate: %w", err)
		}
		last = out
		if strings.HasSuffix(out, want) {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s (pid %s) input not focused with %s (witness %q)",
				gteAppName, pid, want, last)
		}
		if time.Now().After(nextGrab) {
			// Best-effort poke: the refused call still triggers the fresh
			// activation request; a no-op when the app tree is not up yet.
			_, _ = runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "grab-input-pid", pid)
			nextGrab = time.Now().Add(gteGrabEvery)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// readGTEText reads the focused document's text through the helper's
// pid-keyed focused-text bridge — instance-exact, pairing with
// waitGTEInput. The read is retried a few times: a single a11y walk can
// transiently miss the node while the 100 ms witness polling keeps the
// bus busy (live-verified 2026-09-14: witness passed, one-shot read
// found no focused surface).
func (s *stand) readGTEText(ctx context.Context) (string, error) {
	if s.gte == nil || s.gte.Process == nil {
		return "", errors.New("no gnome-text-editor instance to read")
	}
	pid := strconv.Itoa(s.gte.Process.Pid)
	var lastErr error
	for range gteReadAttempts {
		out, err := runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "focused-text-pid", pid)
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

// withEnv returns the current process environment with key overridden to
// value — later duplicates of key removed, so glibc's first-match getenv
// cannot resurrect a stale assignment in the spawned child.
func withEnv(key, value string) []string {
	prefix := key + "="
	env := make([]string, 0, len(os.Environ())+1)
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, prefix) {
			env = append(env, kv)
		}
	}

	return append(env, prefix+value)
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
