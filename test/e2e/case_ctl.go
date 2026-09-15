package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// The ctl-smoke case's config windows: the spawn document and the restored
// one differ so "applied" is a visible change, not a no-op re-parse of the
// same bytes (mnd: named once).
const (
	ctlWindowStart    = 300
	ctlWindowRestored = 450
)

// ctlConfigYAML is the COMPLETE config document of the ctl-smoke case (the
// 03-02 strict-parse rule: no defaults overlay) with the given tap window.
func ctlConfigYAML(windowMs int) string {
	return fmt.Sprintf(`hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: %d
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
`, windowMs)
}

// ctlBrokenYAML carries one unknown key — the strict decoder names it in
// the rejection (D-33), which the reload reply and the status string must
// both surface (D-32: the last-good error is visible from OUTSIDE the
// daemon).
const ctlBrokenYAML = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_mss: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
`

// ctlListeningMark is the daemon's grep-stable readiness record of the
// control service — the gate before the first goswitchctl call.
const ctlListeningMark = `"msg":"ctl service listening"`

// errCtlZenityNeeded is the surface precondition (err113: static).
var errCtlZenityNeeded = errors.New("ctl-smoke needs the zenity entry surface (locked-session fallback engaged?)")

// errCtlEmptyReply is the forced correction's empty-reply failure (err113).
var errCtlEmptyReply = errors.New("ctl-smoke correct reply is empty, want the acknowledgment")

// buildCtl compiles the control client fresh so the stand always tests the
// current tree (the buildDaemon discipline).
func (s *stand) buildCtl(ctx context.Context) (string, error) {
	bin := filepath.Join(s.tmpDir, "goswitchctl")
	cmd := exec.CommandContext(ctx, "go", "build", "-o", bin, "./cmd/goswitchctl")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build goswitchctl: %w: %s", err, strings.TrimSpace(out.String()))
	}

	return bin, nil
}

// runCtl runs the control client capturing BOTH streams without failing on
// a non-zero exit — the broken-reload step asserts on the expected failure
// itself (the exit contract is part of the oracle).
func runCtl(ctx context.Context, bin string, args ...string) (out, errOut string, err error) {
	ctx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	return stdout.String(), stderr.String(), err
}

// runCtlSmoke proves the whole control surface against a LIVE daemon
// (INST-02): goswitchctl status (human + --json) reports the mode and the
// config string; a broken config's reload answers with the rejection and
// the LAST-GOOD ERROR STAYS VISIBLE in status while the daemon keeps
// serving corrections under the last-good snapshot (D-32 end to end);
// correct forces the word pipeline without any tap; the restored config
// reloads with "applied".
func runCtlSmoke(ctx context.Context, s *stand) error {
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}
	cfgPath := filepath.Join(s.tmpDir, "ctl-smoke.yaml")
	if err := os.WriteFile(cfgPath, []byte(ctlConfigYAML(ctlWindowStart)), configFilePerm); err != nil {
		return fmt.Errorf("ctl-smoke: write temp config: %w", err)
	}
	if err := s.restartDaemonWithArgs("-config", cfgPath); err != nil {
		return err
	}
	for _, mark := range []string{
		`"msg":"config loaded"`,
		componentRegisteredMark,
		ctlListeningMark,
	} {
		if err := s.waitForLog(ctx, mark, registrationWait); err != nil {
			return fmt.Errorf("ctl-smoke daemon spawn (%s): %w", mark, err)
		}
	}

	if err := ctlSmokeStatus(ctx, ctlBin, cfgPath); err != nil {
		return err
	}
	if err := ctlSmokeBrokenReload(ctx, ctlBin, cfgPath); err != nil {
		return err
	}
	if err := ctlSmokeForcedCorrect(ctx, ctlBin, s); err != nil {
		return err
	}

	return ctlSmokeRestoreReload(ctx, ctlBin, cfgPath, s)
}

// ctlSmokeStatus proves step 1: the human report carries the mode and the
// config string; --json is valid single-line JSON with the same facts.
func ctlSmokeStatus(ctx context.Context, ctlBin, cfgPath string) error {
	out, _, err := runCtl(ctx, ctlBin, "status")
	if err != nil {
		return fmt.Errorf("ctl-smoke status: %w", err)
	}
	if !strings.Contains(out, "mode=en") || !strings.Contains(out, "config_path="+cfgPath) {
		return fmt.Errorf("ctl-smoke status output %q missing mode or config path", out)
	}
	jout, _, err := runCtl(ctx, ctlBin, "status", "--json")
	if err != nil {
		return fmt.Errorf("ctl-smoke status --json: %w", err)
	}
	if !json.Valid([]byte(jout)) || !strings.Contains(jout, `"mode":"en"`) || !strings.Contains(jout, `"config_path"`) {
		return fmt.Errorf("ctl-smoke status --json output %q is not the JSON form of the same facts", jout)
	}

	return nil
}

// ctlSmokeBrokenReload proves step 2 (D-32 end to end): the broken
// document's reload exits non-zero and names the unknown key, and the
// LAST-GOOD ERROR is visible in status while the daemon keeps the
// last-good snapshot serving.
func ctlSmokeBrokenReload(ctx context.Context, ctlBin, cfgPath string) error {
	if err := os.WriteFile(cfgPath, []byte(ctlBrokenYAML), configFilePerm); err != nil {
		return fmt.Errorf("ctl-smoke: break config: %w", err)
	}
	rout, rerrOut, rerr := runCtl(ctx, ctlBin, "reload")
	if rerr == nil {
		return fmt.Errorf("ctl-smoke broken reload exited 0 (out %q) — the rejection must be a non-zero exit", rout)
	}
	if combined := rout + rerrOut; !strings.Contains(combined, "verify_wait_mss") {
		return fmt.Errorf("ctl-smoke broken reload output %q names no unknown key", combined)
	}
	sout, _, err := runCtl(ctx, ctlBin, "status")
	if err != nil {
		return fmt.Errorf("ctl-smoke status after broken reload: %w", err)
	}
	if !strings.Contains(sout, "config_valid=false") || !strings.Contains(sout, "verify_wait_mss") {
		return fmt.Errorf("ctl-smoke status %q missing the D-32 last-good error", sout)
	}

	return nil
}

// ctlSmokeForcedCorrect proves step 3: correct forces the word pipeline on
// a typed word with NO tap — under the last-good snapshot (the document on
// disk is still the broken one), the settled correction shows in the
// status counters, and both oracles (the AT-SPI readback and the zenity
// stdout) hold the converted word.
func ctlSmokeForcedCorrect(ctx context.Context, ctlBin string, s *stand) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.closeEntrySurface(ctx, kind) }()
	if kind != surfaceZenity {
		return errCtlZenityNeeded
	}
	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"key"`, minKeyEvents, keyWait); err != nil {
		return fmt.Errorf("ctl-smoke key visibility: %w", err)
	}
	cout, _, err := runCtl(ctx, ctlBin, "correct")
	if err != nil {
		return fmt.Errorf("ctl-smoke correct: %w", err)
	}
	if strings.TrimSpace(cout) == "" {
		return errCtlEmptyReply
	}
	if err := s.waitForLog(ctx, `"msg":"correction","outcome":"done"`, correctionWait); err != nil {
		return fmt.Errorf("ctl-smoke forced correction: %w", err)
	}
	if err := s.waitZenityChars(ctx, len([]rune(wordResultRU))); err != nil {
		return fmt.Errorf("ctl-smoke applied correction: %w", err)
	}
	// The correction counter is part of the status surface (ADR-005 b.2's
	// visibility sibling): the settled correction shows in the next report.
	fout, _, err := runCtl(ctx, ctlBin, "status")
	if err != nil {
		return fmt.Errorf("ctl-smoke status after correction: %w", err)
	}
	if !strings.Contains(fout, "corrections_done=1") {
		return fmt.Errorf("ctl-smoke status %q missing corrections_done=1 after the settled correction", fout)
	}
	entryOut, err := s.closeZenity(ctx)
	if err != nil {
		return err
	}
	if entryOut != wordResultRU {
		return fmt.Errorf("ctl-smoke oracle: entry printed %q, want %q", entryOut, wordResultRU)
	}

	return nil
}

// ctlSmokeRestoreReload proves step 4: the restored config reloads — the
// reply says applied, the status is valid again, and the daemon's own
// watcher record confirms the publication.
func ctlSmokeRestoreReload(ctx context.Context, ctlBin, cfgPath string, s *stand) error {
	if err := os.WriteFile(cfgPath, []byte(ctlConfigYAML(ctlWindowRestored)), configFilePerm); err != nil {
		return fmt.Errorf("ctl-smoke: restore config: %w", err)
	}
	aout, _, err := runCtl(ctx, ctlBin, "reload")
	if err != nil {
		return fmt.Errorf("ctl-smoke restored reload: %w", err)
	}
	if !strings.Contains(aout, "applied") {
		return fmt.Errorf("ctl-smoke restored reload reply %q missing applied", aout)
	}
	vout, _, err := runCtl(ctx, ctlBin, "status")
	if err != nil {
		return fmt.Errorf("ctl-smoke status after restore: %w", err)
	}
	if !strings.Contains(vout, "config_valid=true") {
		return fmt.Errorf("ctl-smoke status %q missing config_valid=true after the applied reload", vout)
	}
	if err := s.waitForLog(ctx, `"msg":"config reloaded"`, decisionWait); err != nil {
		return fmt.Errorf("ctl-smoke reloaded record: %w", err)
	}

	return nil
}
