package main

import (
	"context"
	"fmt"
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

// surfaceFocusWait budgets how long a spawned surface may take from process
// start to a focused, empty input node in the AT-SPI tree: a cold Chromium
// start needs several seconds before its renderer accessibility tree is up.
const surfaceFocusWait = 15 * time.Second

// surfaceChromium is the stand-spawned fresh Chromium instance — the
// surfaceKind enumeration of case_m1.go extended from this file so that
// case_m1.go itself stays untouched.
const surfaceChromium surfaceKind = iota + 2

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
