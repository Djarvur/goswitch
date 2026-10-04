package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Djarvur/goswitch/internal/appid"
)

// The autocorrect live proofs (plan 06-07, criterion 5, MACR-ACL, criterion
// 2's unknown→silence): three cases driving the 06-06 actor contour on the
// real desktop —
//
//	autocorrect-fires           the word corrects WITHOUT any hotkey in the
//	                            fixture's TEXT field (counter + readback +
//	                            stdout oracle + the manual-double regression);
//	autocorrect-password-silent the same enabled layer stays SILENT on the
//	                            fixture's GtkPasswordEntry (AT-SPI
//	                            PASSWORD_TEXT=40 — the role half of D-53),
//	                            proven by three independent oracles;
//	autocorrect-terminal-silent the terminal negative rides wezterm's ABSENCE
//	                            from the a11y bus (06-RESEARCH Q6): counters
//	                            are the only oracle, the app gate refuses.
//
// The white-list half of the conjunction uses the OBSERVED bridge-namespace
// identity (the macr-per-app discipline — never guessed): a case-side
// appid.Observer reads the identity of the fixture's fresh map-focus event
// and that exact string becomes the config's apps entry. The GTK4 fixture
// pins its own prgname, so the identity is deterministic
// (org.gtk.application.gs_e2e_password) and NEVER the generic python3
// namespace. Chromium's <input type=password> was pinned LIVE (2026-10-01,
// research A2): google-chrome 153 exposes it as AT-SPI PASSWORD_TEXT (40)
// too — the matrix carries no chromium role row (the driver has no
// password-surface step), the GTK4 fixture stays the role negative carrier.

const (
	// acWord is the injected probe: the canonical wrong-layout word plus
	// the space separator — the separator is the word boundary the
	// detector's decision rides on (Buffer.PushFeed).
	acWord = "ghbdtn "
	// acWordRU is the corrected field content: the level-1 ladder's
	// delete+commit with the trailing separator riding inside the D-13
	// replacement range. The fixture's stdout oracle (get_text, byte-exact)
	// pins it; the AT-SPI LINE readback cannot — this client's
	// line-granularity answer drops an engine-committed trailing separator
	// (the readback class v3 documents for gedit), so the readback
	// oracles compare the word alone (acWordReadback/acWordRUReadback).
	acWordRU         = "привет "
	acWordReadback   = "ghbdtn"
	acWordRUReadback = "привет"

	// The AT-SPI role labels the pid-keyed witness reports (the
	// focus_helper INPUT_ROLES naming): the fixture's GtkEntry answers
	// TEXT, its GtkPasswordEntry PASSWORD_TEXT (06-RESEARCH Q4, live).
	acRoleTextWitness     = "TEXT"
	acRolePasswordWitness = "PASSWORD_TEXT"

	// The grep-stable daemon records of the 06-06 contour: the fired INFO
	// (slog.Info("autocorrect", "reason", "fired")), the abstention INFO
	// ("autocorrect skipped", reason slug — role-forbidden/app-not-listed
	// of the actor's closed slug vocabulary) and the settled correction.
	acFiredMark      = `"msg":"autocorrect","reason":"fired"`
	acAbstainFmt     = `"msg":"autocorrect skipped","reason":"%s"`
	correctionAnyMar = `"msg":"correction"`
	correctionDoneMa = `"msg":"correction","outcome":"done"`
	actionDoubleMark = `"msg":"action","n":2`

	// acReasonRoleForbidden mirrors the actor's closed abstention slug
	// (internal/session actor.go) — the reason the password case's oracle
	// requires.
	acReasonRoleForbidden = "role-forbidden"

	// acFixtureTimeoutSec is the fixture's lifetime argument: long enough
	// for the whole observation/injection round, short enough that the
	// natural (non-killed) exit prints the FINAL:/PLAIN: oracle lines
	// within the case's watchdog. acFixtureExitGrace is the extra wait
	// past the timeout before the exit is declared lost.
	acFixtureTimeoutSec = 25
	acFixtureExitGrace  = 10 * time.Second

	// acTerminalGuardApp is the terminal case's white-list entry: a
	// sentinel bridge namespace nothing on the desktop can report — the
	// layer demonstrably runs against an ACTIVE (non-empty) list while the
	// desktop state stays out of the oracle.
	acTerminalGuardApp = "org.gnome.GoswitchE2eAbsent"

	// acTerminalSettle is the bounded wait after the terminal injection:
	// an IM-routed key event reaches the daemon log within milliseconds
	// (the control round calibrates the latency), so a quiet 2 s window
	// is the absence proof's budget.
	acTerminalSettle = 2 * time.Second

	// acControlKeyEvents is the positive control's own floor: ONE
	// keystroke yields press+release (2 records) above the round's base —
	// minKeyEvents calibrates the 7-keystroke probe word, not the control
	// (06-07 live finding: the absolute 6 never admits a single-key round).
	acControlKeyEvents = 2
)

// weztermBin is the terminal surface of the terminal-silent negative — the
// owner's actual terminal, fully absent from the a11y bus (06-RESEARCH Q4:
// zero focus events, zero roles — the unknown→silence premise).
const weztermBin = "wezterm"

// Static failure shapes of the fixture oracle (err113-relaxed package; the
// named forms keep the diagnostics greppable).
var (
	errFixtureGone   = errors.New("no password fixture process")
	errFixtureOracle = errors.New("password fixture stdout carries no FINAL:/PLAIN: oracle lines")
)

// autocorrectConfigTmpl is the cases' complete config document (the 03-02
// no-overlay rule): all five sections explicit, the layer ON with the
// observed bridge identity as the single white-list entry, the 06-05 corpus
// thresholds and flip_after_correction pinned false — the manual-double
// regression round of the fires case needs the mode kept at EN (a
// post-correction flip would make the corrected word dictionary-valid in
// the flipped mode and the manual gesture a no-op).
const autocorrectConfigTmpl = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
  flip_after_correction: false # the manual-double regression round needs the EN mode kept
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
autocorrect:
  enabled: true
  apps: ["%s"]
  min_word_len: 4
  trigram_margin: 2.0
  trigram_floor: 1.0
`

// startAutocorrectConfigDaemon writes the case's config document (the
// observed identity as the white-list entry), restarts the daemon on it and
// re-activates the engine — the -config spawn ladder of the config cases
// (startMacrConfigDaemon form) extended with the ctl readiness mark the
// status oracles need.
func startAutocorrectConfigDaemon(ctx context.Context, s *stand, app string) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	cfgPath := filepath.Join(s.tmpDir, "autocorrect.yaml")
	doc := fmt.Sprintf(autocorrectConfigTmpl, app)
	if err := os.WriteFile(cfgPath, []byte(doc), configFilePerm); err != nil {
		return fmt.Errorf("autocorrect: write temp config: %w", err)
	}
	if err := s.restartDaemonWithArgs("-config", cfgPath); err != nil {
		return err
	}
	for _, mark := range []string{configLoadedMark, componentRegisteredMark, ctlListeningMark} {
		if err := s.waitForLog(ctx, mark, registrationWait); err != nil {
			return fmt.Errorf("autocorrect daemon -config spawn (%s): %w", mark, err)
		}
	}

	return s.activateGoswitch(ctx)
}

// startPasswordFixture spawns the GTK4 fixture (a stand-owned surface —
// never an owner window); its stdout collects the FINAL:/PLAIN: oracle
// lines the natural (timeout) exit prints.
func (s *stand) startPasswordFixture(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "/usr/bin/python3", "test/e2e/fixtures/password_entry.py",
		strconv.Itoa(acFixtureTimeoutSec))
	out := &bytes.Buffer{}
	cmd.Stdout = out
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start password fixture: %w", err)
	}
	s.pwFixture = cmd
	s.pwFixtureOut = out

	return nil
}

// fixturePid returns the fixture's process id (0 when none is running).
func (s *stand) fixturePid() int {
	if s.pwFixture == nil || s.pwFixture.Process == nil {
		return 0
	}

	return s.pwFixture.Process.Pid
}

// reapPasswordFixture force-kills a leftover fixture (teardown and error
// paths) — a kill before the natural exit silences the stdout oracle,
// which is exactly why the cases wait the timeout out on the happy path.
func (s *stand) reapPasswordFixture() {
	if s.pwFixture == nil {
		return
	}
	if s.pwFixture.Process != nil {
		_ = s.pwFixture.Process.Kill()
	}
	_ = s.pwFixture.Wait()
	s.pwFixture = nil
	s.pwFixtureOut = nil
}

// waitFixtureWitness polls the pid-keyed witness until the fixture's
// focused input answers exactly "<role>:chars=<chars>" — the gate that
// makes injection safe and the silent cases' witness oracle. A missing
// focus or role is a FAIL, never a skip (T-06-07-01). Past the grace the
// loop pokes the fixture's first input node with the pid-keyed grabFocus
// (the fresh-map focus-stealing recovery) — only while the fresh TEXT
// focus is the target; the post-Tab password waits run pokeless (the poke
// targets the TEXT node and would fight the very focus being waited on).
func (s *stand) waitFixtureWitness(ctx context.Context, role string, chars int, poke bool) (string, error) {
	pid := s.fixturePid()
	if pid == 0 {
		return "", errFixtureGone
	}
	pidStr := strconv.Itoa(pid)
	want := role + ":chars=" + strconv.Itoa(chars)
	deadline := time.Now().Add(surfaceFocusWait)
	nextGrab := time.Now().Add(zenityGrabGrace)
	var last string
	for {
		out, err := runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "focused-input-pid", pidStr)
		if err != nil {
			return "", fmt.Errorf("fixture witness gate: %w", err)
		}
		last = out
		if out == want {
			return out, nil
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("fixture input did not settle to %q (witness %q)", want, last)
		}
		if poke && time.Now().After(nextGrab) {
			// Best-effort poke: the refused grab still re-issues the
			// activation request (the 02-02 finding).
			_, _ = runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "grab-input-pid", pidStr)
			nextGrab = time.Now().Add(zenityGrabEvery)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return "", err
		}
	}
}

// waitFixtureText polls the pid-keyed focused-text readback until the
// fixture's focused field holds exactly want — the settle-riding oracle of
// waitMatrixReadback, instance-exact through the pid.
func (s *stand) waitFixtureText(ctx context.Context, want string) error {
	pid := s.fixturePid()
	if pid == 0 {
		return errFixtureGone
	}
	deadline := time.Now().Add(witnessWait)
	var last string
	for {
		out, err := s.readFocusedTextPid(ctx, pid)
		if err == nil {
			last = out
			if out == want {
				return nil
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("fixture field did not settle to %q (readback %q)", want, last)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// waitFixtureExit waits out the fixture's natural timeout exit (the GLib
// timeout prints the FINAL:/PLAIN: oracle — a kill would silence it) and
// returns the two oracle lines: FINAL is the password field, PLAIN the
// text field.
func (s *stand) waitFixtureExit(ctx context.Context) (final, plain string, err error) {
	if s.pwFixture == nil || s.pwFixture.Process == nil {
		return "", "", errFixtureGone
	}
	done := make(chan struct{})
	go func() {
		_ = s.pwFixture.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		return "", "", fmt.Errorf("fixture exit wait aborted: %w", ctx.Err())
	case <-time.After(acFixtureTimeoutSec*time.Second + acFixtureExitGrace):
		s.reapPasswordFixture()

		return "", "", errors.New("password fixture did not exit within its timeout budget")
	}
	final, plain, perr := parseFixtureOracle(s.pwFixtureOut.String())
	s.pwFixture = nil
	s.pwFixtureOut = nil

	return final, plain, perr
}

// parseFixtureOracle splits the fixture's stdout oracle lines. FINAL: may
// legitimately print an EMPTY password field (the silent case's expected
// actual) — the missing-oracle verdict keys on the line's presence, never
// on its content.
func parseFixtureOracle(out string) (final, plain string, err error) {
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "FINAL:"):
			final = strings.TrimPrefix(line, "FINAL:")
		case strings.HasPrefix(line, "PLAIN:"):
			plain = strings.TrimPrefix(line, "PLAIN:")
		}
	}
	if !strings.Contains(out, "FINAL:") || !strings.Contains(out, "PLAIN:") {
		return "", "", errFixtureOracle
	}

	return final, plain, nil
}

// observeFixtureApp reads the bridge-namespace identity the a11y observer
// gleans from the fixture's fresh map-focus event — the exact vocabulary
// the daemon's own observer reports and the white list must match. The
// observation fixture is reaped before the caller spawns the injection
// surface: the daemon's lazy observer starts at the config fold and only a
// POST-fold fresh map feeds it (the macr-per-app lesson — an
// already-focused window never emits a gain, and the poke proved useless).
func observeFixtureApp(ctx context.Context, s *stand, caseName string) (string, error) {
	obs, err := appid.Start(ctx)
	if err != nil {
		return "", fmt.Errorf("%s: start case observer: %w", caseName, err)
	}
	if err := s.startPasswordFixture(ctx); err != nil {
		return "", err
	}
	defer s.reapPasswordFixture()
	if _, err := s.waitFixtureWitness(ctx, acRoleTextWitness, 0, true); err != nil {
		return "", err
	}
	app, err := waitFocusedApp(ctx, obs)
	if err != nil {
		return "", fmt.Errorf("%s: observe fixture identity: %w", caseName, err)
	}
	if app == "" {
		return "", fmt.Errorf("%s: no bridge identity observed for the fixture window", caseName)
	}
	fmt.Printf("%s: observed fixture identity %q\n", caseName, app)

	return app, nil
}

// acStartObserverRound drives the daemon's FIRST key event on a stand-owned
// zenity: the per-event snapshot read is where applySnapshot folds the
// autocorrect section AND ensureAppid lazily starts the a11y observer —
// both post-date any surface that mapped before them, so the identity-
// bearing injection surface must map only AFTER this round (the
// macr-per-app surface-C lesson: a fresh map-focus gain is what feeds the
// freshly started observer; an already-mapped window never re-emits one).
// The keystroke is a single letter with no separator — boundary-free, so
// no autocorrect counter moves.
func (s *stand) acStartObserverRound(ctx context.Context) error {
	if err := s.startZenity(ctx); err != nil {
		return err
	}
	defer s.reapZenity()
	if err := s.waitZenityEntry(ctx); err != nil {
		return err
	}

	return s.injectText(ctx, "x")
}

// ctlStatusHas runs goswitchctl status and requires every want token — the
// D-54 counters/status-surface oracle.
func ctlStatusHas(ctx context.Context, ctlBin string, want ...string) error {
	out, _, err := runCtl(ctx, ctlBin, "status")
	if err != nil {
		return fmt.Errorf("autocorrect status: %w", err)
	}
	for _, token := range want {
		if !strings.Contains(out, token) {
			return fmt.Errorf("autocorrect status %q missing %q", out, token)
		}
	}

	return nil
}

// acTypeAndWaitKeys injects the probe word plus separator and gates on the
// key records the daemon traced — the injection-visibility floor every
// case shares.
func (s *stand) acTypeAndWaitKeys(ctx context.Context) error {
	base := s.countSub(`"msg":"key"`)
	if err := s.injectText(ctx, acWord); err != nil {
		return err
	}

	return s.waitForNew(ctx, `"msg":"key"`, base+minKeyEvents, keyWait)
}

// assertAcSilent is the shared negative oracle of the silent cases: zero
// fired records and zero correction records in the daemon log — the
// nothing-entered-the-ladder half of the silence proof.
func (s *stand) assertAcSilent(caseName string) error {
	if got := s.countSub(acFiredMark); got != 0 {
		return fmt.Errorf("%s: %d fired records — the field must never fire", caseName, got)
	}
	if got := s.countSub(correctionAnyMar); got != 0 {
		return fmt.Errorf("%s: %d correction records — nothing may enter the ladder", caseName, got)
	}

	return nil
}

// assertFixtureExitOracles checks the fixture's natural-exit stdout oracle
// lines against the case's expectation (the delivery half of the proofs —
// verbatim unchanged for the silent cases, corrected/re-converted for the
// fires case).
func assertFixtureExitOracles(caseName, final, plain, wantFinal, wantPlain string) error {
	if final != wantFinal {
		return fmt.Errorf("%s: FINAL oracle %q, want %q", caseName, final, wantFinal)
	}
	if plain != wantPlain {
		return fmt.Errorf("%s: PLAIN oracle %q, want %q", caseName, plain, wantPlain)
	}

	return nil
}

// runAutocorrectFires proves the automatic pipeline end to end on a live
// desktop (criterion 5): with the layer enabled and the fixture's observed
// identity whitelisted, injecting the wrong-layout word plus its separator
// into the fixture's TEXT field corrects the word WITHOUT any hotkey. The
// fired record and the D-54 counters prove the decision; the AT-SPI
// readback and the fixture's PLAIN stdout oracle prove the delivery; the
// manual Double Right Shift still converts the corrected word back — the
// CORR-01 hotkey regression on the same surface.
func runAutocorrectFires(ctx context.Context, s *stand) error {
	const caseName = "autocorrect-fires"
	app, err := observeFixtureApp(ctx, s, caseName)
	if err != nil {
		return err
	}
	if err := startAutocorrectConfigDaemon(ctx, s, app); err != nil {
		return err
	}
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}
	// The observer-start round must precede the injection surface: the
	// daemon's lazy observer starts at its first key event, and the fresh
	// fixture map is the gain it learns from.
	if err := s.acStartObserverRound(ctx); err != nil {
		return err
	}
	// The injection surface: a FRESH fixture instance — its map-focus gain
	// postdates the observer start and is what feeds the daemon's cache.
	if err := s.startPasswordFixture(ctx); err != nil {
		return err
	}
	defer s.reapPasswordFixture()
	if _, err := s.waitFixtureWitness(ctx, acRoleTextWitness, 0, true); err != nil {
		return err
	}

	if err := s.acTypeAndWaitKeys(ctx); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, acFiredMark, correctionWait); err != nil {
		return fmt.Errorf("%s: fired record: %w", caseName, err)
	}
	if err := s.waitForLog(ctx, correctionDoneMa, correctionWait); err != nil {
		return fmt.Errorf("%s: settled correction: %w", caseName, err)
	}
	if err := s.waitFixtureText(ctx, acWordRUReadback); err != nil {
		return fmt.Errorf("%s: corrected readback: %w", caseName, err)
	}
	if err := ctlStatusHas(ctx, ctlBin, "autocorrect_enabled=true", "autocorrect_fired=1"); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}

	return s.acManualDoubleRound(ctx, caseName)
}

// acManualDoubleRound is the additive CORR-01 regression of the fires case:
// the manual Double Right Shift still converts on the same surface after an
// automatic correction — the hotkey path is not displaced by the layer —
// and the fixture's stdout oracle closes the round (PLAIN holds the
// manually re-converted word, FINAL the untouched empty password field).
func (s *stand) acManualDoubleRound(ctx context.Context, caseName string) error {
	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, actionDoubleMark, decisionWait); err != nil {
		return fmt.Errorf("%s: manual double decision: %w", caseName, err)
	}
	if err := s.waitFixtureText(ctx, acWordReadback); err != nil {
		return fmt.Errorf("%s: manual conversion readback: %w", caseName, err)
	}
	final, plain, err := s.waitFixtureExit(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	if err := assertFixtureExitOracles(caseName, final, plain, "", acWord); err != nil {
		return err
	}
	fmt.Printf("%s: fired, corrected without a hotkey, manual double converted back — oracles green\n", caseName)

	return nil
}

// runAutocorrectPasswordSilent proves the role half of the D-53
// conjunction live (criterion 2): the same enabled layer and white-listed
// app, the injection aimed at the fixture's GtkPasswordEntry (a Tab moves
// the GTK focus; the fresh gain also re-points the daemon's (sender, path)
// pair, so the live GetRole at the boundary queries the PASSWORD node).
// Three independent proofs (T-06-07-01): the pid-keyed witness holds
// PASSWORD_TEXT with an UNCHANGED observed char count across the
// injection, the fixture's FINAL stdout oracle holds the injected content
// verbatim, and the counters show fired=0 with exactly one
// role-forbidden abstention. A missing focus/role witness is a FAIL, not
// a skip.
func runAutocorrectPasswordSilent(ctx context.Context, s *stand) error {
	const caseName = "autocorrect-password-silent"
	app, err := observeFixtureApp(ctx, s, caseName)
	if err != nil {
		return err
	}
	if err := startAutocorrectConfigDaemon(ctx, s, app); err != nil {
		return err
	}
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}
	// Same observer-start discipline as the fires case: the fixture below
	// maps only after the daemon's observer is live.
	if err := s.acStartObserverRound(ctx); err != nil {
		return err
	}
	if err := s.startPasswordFixture(ctx); err != nil {
		return err
	}
	defer s.reapPasswordFixture()
	if _, err := s.waitFixtureWitness(ctx, acRoleTextWitness, 0, true); err != nil {
		return err
	}
	before, err := s.acFocusPasswordField(ctx, caseName)
	if err != nil {
		return err
	}
	if err := s.acSilentInjectionRound(ctx, caseName); err != nil {
		return err
	}
	// The after-witness: the password field holds EXACTLY what was typed
	// (the injected rune count) — the content itself is pinned byte-exact
	// by the FINAL stdout oracle below; the count additionally rules out
	// any extra text. A correction would still be 7 runes («привет ») —
	// the daemon-side silence counters and the verbatim FINAL are the
	// distinguishing proofs.
	after, err := s.waitFixtureWitness(ctx, acRolePasswordWitness, utf8.RuneCountInString(acWord), false)
	if err != nil {
		return fmt.Errorf("%s: typed-content witness: %w", caseName, err)
	}
	if err := s.assertAcSilent(caseName); err != nil {
		return err
	}
	if err := ctlStatusHas(ctx, ctlBin, "autocorrect_enabled=true", "autocorrect_fired=0",
		"ac_skip_role_forbidden=1"); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	final, plain, err := s.waitFixtureExit(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	if err := assertFixtureExitOracles(caseName, final, plain, acWord, ""); err != nil {
		return err
	}
	fmt.Printf("%s: password silent — witness %q→%q, fired=0, one role-forbidden abstention\n", caseName, before, after)

	return nil
}

// acFocusPasswordField moves the fixture's GTK focus from the text entry to
// the password entry (a Tab — the fresh gain also re-points the daemon's
// (sender, path) pair, so the live GetRole at the boundary queries the
// PASSWORD node) and gates the focused-password witness — the returned
// line is the "before" record of the unchanged-content proof.
func (s *stand) acFocusPasswordField(ctx context.Context, caseName string) (string, error) {
	if err := s.pressKey(ctx, "tab"); err != nil {
		return "", err
	}
	before, err := s.waitFixtureWitness(ctx, acRolePasswordWitness, 0, false)
	if err != nil {
		return "", fmt.Errorf("%s: password focus witness: %w", caseName, err)
	}

	return before, nil
}

// acSilentInjectionRound injects the probe at the password field and gates
// the role-forbidden abstention — the daemon's decision record of the
// silence.
func (s *stand) acSilentInjectionRound(ctx context.Context, caseName string) error {
	if err := s.acTypeAndWaitKeys(ctx); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, fmt.Sprintf(acAbstainFmt, acReasonRoleForbidden), decisionWait); err != nil {
		return fmt.Errorf("%s: role-forbidden abstention: %w", caseName, err)
	}

	return nil
}

// startWezterm spawns a fresh terminal window (--always-new-process: never
// delegate to an owner instance — the chromium-profile isolation lesson).
// The window takes focus on map like any fresh surface and is invisible to
// the a11y bus, which is exactly the property the negative pins.
func (s *stand) startWezterm(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, weztermBin, "start", "--always-new-process")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", weztermBin, err)
	}
	s.wezterm = cmd

	return nil
}

// closeWezterm terminates the spawned terminal by pid — a fresh window has
// no state to save. Doubles as the teardown reap; a no-op once closed.
func (s *stand) closeWezterm() {
	if s.wezterm == nil {
		return
	}
	if s.wezterm.Process != nil {
		_ = s.wezterm.Process.Kill()
	}
	_ = s.wezterm.Wait()
	s.wezterm = nil
}

// runAutocorrectTerminalSilent proves the terminal negative (criterion 2's
// unknown→silence) on the only oracle a terminal offers — counters
// (06-RESEARCH Q6). CORRECTED LIVE FINDING (2026-10-02, this case's runs):
// the first pin ("terminal typing produces ZERO engine events") held only
// in some runs — wezterm DOES register an IBus input context when focused
// (caps 0x9, no surrounding-text bit) and its typing RACES the IME arming:
// one run delivered all 7 keystrokes + the boundary through the daemon's
// key trace, another delivered zero records. BOTH shapes are safe by the
// same daemon property, which is what the oracle set now pins, with a
// positive CONTROL in the same case against deafness-by-breakage:
//
//	control — the same daemon, the same injector, the IM-routed zenity:
//	          the keystroke appears in the key trace;
//	terminal — wezterm focused (the witness must stop naming our zenity —
//	         the only fresh surface was the terminal): the injection may
//	         reach the daemon through wezterm's IM context or not at all —
//	         EITHER way the layer stays silent: zero fired records, zero
//	         correction records; a boundary that does arrive is refused
//	         fail-closed (the 0x9 context has no surrounding-text bit —
//	         D-53 condition 3, the no-caps slug; abstention records carry
//	         the closed slug vocabulary, unit-pinned in internal/session).
//
// The status closes the case: autocorrect_enabled=true (the negative runs
// against an ACTIVE layer) and autocorrect_fired=0. The stale-cache shape
// (a whitelisted app focused BEFORE the terminal) is a different scenario
// and deliberately out of this case's scope.
func runAutocorrectTerminalSilent(ctx context.Context, s *stand) error {
	const caseName = "autocorrect-terminal-silent"
	if err := startAutocorrectConfigDaemon(ctx, s, acTerminalGuardApp); err != nil {
		return err
	}
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}
	if err := s.startZenity(ctx); err != nil {
		return err
	}
	defer s.reapZenity()
	if err := s.waitZenityEntry(ctx); err != nil {
		return err
	}
	if err := s.acControlKeystroke(ctx, caseName); err != nil {
		return err
	}
	if err := s.startWezterm(ctx); err != nil {
		return err
	}
	defer s.closeWezterm()
	if err := s.acWaitWeztermFocus(ctx, caseName); err != nil {
		return err
	}
	if err := s.acTerminalOracle(ctx, ctlBin, caseName); err != nil {
		return err
	}
	fmt.Printf("%s: terminal silent — fired=0, zero correction records under both IME delivery modes\n", caseName)

	return nil
}

// acControlKeystroke is the positive control: one keystroke on the
// IM-routed zenity is visible to the daemon (the injector and the engine
// are alive) — the deafness counter-evidence the terminal round needs.
func (s *stand) acControlKeystroke(ctx context.Context, caseName string) error {
	base := s.countSub(`"msg":"key"`)
	if err := s.injectText(ctx, "x"); err != nil {
		return err
	}
	if err := s.waitForNew(ctx, `"msg":"key"`, base+acControlKeyEvents, keyWait); err != nil {
		return fmt.Errorf("%s: control keystroke visibility: %w", caseName, err)
	}

	return nil
}

// acWaitWeztermFocus is the focus proof: the witness must stop naming our
// zenity (stale AT-SPI FOCUSED bits clear within beats — the 02-06
// lesson) — a blind injection into an unknown surface is forbidden.
func (s *stand) acWaitWeztermFocus(ctx context.Context, caseName string) error {
	deadline := time.Now().Add(witnessWait)
	for {
		witness, werr := s.focusWitness(ctx)
		if werr != nil {
			return werr
		}
		if !strings.HasPrefix(witness, "zenity:") {
			fmt.Printf("%s: witness %q — wezterm is off the a11y bus, the negative witness is the expected actual\n",
				caseName, witness)

			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s: our zenity still holds focus — the terminal did not take it (witness %q)",
				caseName, witness)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}

	return nil
}

// acTerminalOracle runs the terminal round's oracle set: after the settle
// budget the counters close the case — zero fired records, zero correction
// records, active-layer status. The injection's key trace may or may not
// reach the daemon (wezterm's IME arming races the injection — the
// corrected live finding in runAutocorrectTerminalSilent); a boundary that
// does arrive is refused fail-closed (no-caps: the 0x9 context has no
// surrounding-text bit), and the abstention records carry the closed slug
// vocabulary — silence under BOTH delivery modes is the pinned property.
func (s *stand) acTerminalOracle(ctx context.Context, ctlBin, caseName string) error {
	if err := s.injectText(ctx, acWord); err != nil {
		return err
	}
	if err := sleepCtx(ctx, acTerminalSettle); err != nil {
		return err
	}
	if err := s.assertAcSilent(caseName); err != nil {
		return err
	}
	if err := ctlStatusHas(ctx, ctlBin, "autocorrect_enabled=true", "autocorrect_fired=0"); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}

	return nil
}
