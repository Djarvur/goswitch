package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// d01-probe knobs and identifiers: the probe matrix of the D-01 experiment
// (plan 01-04, Task 1). Every probe's success criterion is OBSERVABLE state
// (focus movement of our engines in the daemon log plus the script of
// subsequently typed text) — never a dconf value (Pitfall 3).
const (
	d01Journal   = "docs/adr/d01-experiment-log.md"
	d01Text      = "ghbdtn" // typed after each probe; Cyrillic output = RU XKB applied
	d01ProbeWait = 6 * time.Second
	// goswitchSourcesPrefix is prepended to the snapshotted sources value so
	// both goswitch engines enter the GNOME source list at indices 0 and 1.
	goswitchSourcesPrefix = "[('ibus', 'goswitch-en'), ('ibus', 'goswitch-ru')]"
	// goswitchRUIndex is the index of goswitch-ru in the probe sources list.
	goswitchRUIndex = "uint32 1"
	probeUUID       = "goswitch-probe@localhost"
)

// focus log substrings — the engine adapter names the engine in every
// lifecycle record, so the observation can tell WHICH engine took focus.
const (
	focusOutEN = `"msg":"focus_out","engine":"goswitch-en"`
	focusInEN  = `"msg":"focus_in","engine":"goswitch-en"`
	focusInRU  = `"msg":"focus_in","engine":"goswitch-ru"`
)

// Permissions for the probe's user-local artifacts and the journal (gosec
// floors: 0750 dirs, 0600 files — nothing here needs to be wider).
const (
	probeDirPerm  os.FileMode = 0o750
	probeFilePerm os.FileMode = 0o600
)

// runD01Probe executes the D-01 probe matrix (CONTEXT D-01): can the daemon
// command GNOME to switch the active input source? Probes run in the
// research-prescribed order — gsettings write, synthetic GNOME switcher
// hotkeys, IBus SetGlobalEngine, org.gnome.Shell.Eval — and, only when none
// of those switched, the time-boxed micro-extension probe. Each probe
// appends a verdict row to the experiment journal; the case PASSES when
// every executed probe left a verdict and a decision is derivable. Probes
// are live-session integration checks, not TDD (D-07); the canonical run is
// mise run e2e-d01 (D-09).
func runD01Probe(ctx context.Context, s *stand) error {
	if err := resetToGoswitchEN(ctx, s); err != nil {
		return err
	}

	switched := false
	for _, probe := range []struct {
		name string
		run  func(context.Context, *stand) (bool, error)
	}{
		{name: "gsettings", run: probeGsettings},
		{name: "ydotool super+space", run: probeYdotoolSuperSpace},
		{name: "ydotool alt+shift_l", run: probeYdotoolAltShift},
		{name: "set-global-engine", run: probeSetGlobalEngine},
		{name: "shell-eval", run: probeShellEval},
	} {
		fmt.Printf("d01: probe %s: running\n", probe.name)
		didSwitch, err := probe.run(ctx, s)
		if err != nil {
			return fmt.Errorf("probe %s: %w", probe.name, err)
		}
		switched = switched || didSwitch
		if err := resetToGoswitchEN(ctx, s); err != nil {
			return fmt.Errorf("reset after probe %s: %w", probe.name, err)
		}
	}

	if !switched {
		didSwitch, err := probeMicroExtension(ctx)
		if err != nil {
			return fmt.Errorf("probe micro-extension: %w", err)
		}
		switched = switched || didSwitch
	}

	return appendJournalRow("Result "+time.Now().Format("15:04:05"), "—",
		fmt.Sprintf("switched probes: %v", switched), decisionFor(switched))
}

// decisionFor derives the D-01 outcome from the probe matrix: any switched
// probe makes two-engine available; otherwise Option B (the owner-proven
// internal flip) — a result, not a failure.
func decisionFor(switched bool) string {
	if switched {
		return "DECISION: two-engine mechanism available (see ADR-001)"
	}

	return "DECISION: Option B — internal flip (the owner prototype pattern; see ADR-001)"
}

// probeGsettings — probe 1: write the gsettings pair (sources list with both
// goswitch engines, current pointed at goswitch-ru) and observe. The
// research's expected failure mode: GNOME 46's shell holds live
// InputSourceManager state and ignores external writes at runtime.
func probeGsettings(ctx context.Context, s *stand) (bool, error) {
	sources := withGoswitchSources(s.snap.sources)
	action := func() error {
		if _, err := runCmd(ctx, "gsettings", "set", gsettingsSchema, keySources, sources); err != nil {
			return fmt.Errorf("set sources: %w", err)
		}
		if _, err := runCmd(ctx, "gsettings", "set", gsettingsSchema, keyCurrent, goswitchRUIndex); err != nil {
			return fmt.Errorf("set current: %w", err)
		}

		return nil
	}
	obs := s.observeSwitch(ctx, action)

	return obs.record("1. gsettings "+time.Now().Format("15:04:05"),
		"gsettings set org.gnome.desktop.input-sources sources/current → goswitch-ru")
}

// probeYdotoolSuperSpace — probe 2a: the synthetic Super+Space, one of the
// live-verified GNOME switch-input-source bindings.
func probeYdotoolSuperSpace(ctx context.Context, s *stand) (bool, error) {
	obs := s.observeSwitch(ctx, func() error {
		return s.pressKey(ctx, "super+space")
	})

	return obs.record("2a. ydotool super+space "+time.Now().Format("15:04:05"), "ydotool key super+space")
}

// probeYdotoolAltShift — probe 2b: the synthetic Alt+Shift_L, the other
// live-verified switch-input-source binding.
func probeYdotoolAltShift(ctx context.Context, s *stand) (bool, error) {
	obs := s.observeSwitch(ctx, func() error {
		return s.pressKey(ctx, "alt+Shift_L")
	})

	return obs.record("2b. ydotool alt+shift_l "+time.Now().Format("15:04:05"), "ydotool key alt+Shift_L")
}

// probeSetGlobalEngine — probe 3: IBus SetGlobalEngine (the `ibus engine`
// CLI). The research predicts the engine changes without the XKB group
// following (keyboard.js 46 applies layouts only on its own
// InputSource.activate) — the probe records exactly that split.
func probeSetGlobalEngine(ctx context.Context, s *stand) (bool, error) {
	obs := s.observeSwitch(ctx, func() error {
		if _, err := runCmd(ctx, "ibus", "engine", "goswitch-ru"); err != nil {
			return fmt.Errorf("ibus engine goswitch-ru: %w", err)
		}

		return nil
	})

	return obs.record("3. set-global-engine "+time.Now().Format("15:04:05"), "ibus engine goswitch-ru")
}

// probeShellEval — probe 4: org.gnome.Shell.Eval with an activate() script.
// GNOME 41+ gates Eval behind unsafe mode; unsafe mode is NEVER enabled
// (T-04-01, D-02), so the expected verdict is unavailable with the refusal
// as the observed. Should Eval nevertheless run, the observation flow still
// measures the switch.
func probeShellEval(ctx context.Context, s *stand) (bool, error) {
	const script = "Main.getInputSourceManager().inputSources[1].activate()"
	var reply string
	action := func() error {
		out, err := runCmd(ctx, "gdbus", "call", "--session",
			"--dest", "org.gnome.Shell", "--object-path", "/org/gnome/Shell",
			"--method", "org.gnome.Shell.Eval", script)
		reply = fmt.Sprintf("reply %q err %v", out, err)

		// A refused Eval is a verdict, not a case failure.
		return nil
	}
	obs := s.observeSwitch(ctx, action)
	obs.extra = reply

	// GNOME 41+ gates Eval behind unsafe mode: the call itself succeeds with
	// success=false. That is an unavailable mechanism (unsafe mode is never
	// enabled — D-02/T-04-01), not a not-switched observation.
	override := ""
	if strings.Contains(reply, "(false,") {
		override = "Verdict: unavailable — Shell.Eval отклонил вычисление (unsafe mode выключен;" +
			" включать запрещено — D-02/T-04-01)"
	}

	return obs.recordVerdict("4. shell eval "+time.Now().Format("15:04:05"),
		"gdbus call org.gnome.Shell.Eval 'Main.getInputSourceManager().inputSources[1].activate()'", override)
}

// probeMicroExtension — probe 5 (conditional, time-boxed): a minimal
// user-local shell extension that would call InputSourceManager.activate(N)
// on enable. Loading a NEW extension requires a GNOME Shell restart, which
// on Wayland means logging the owner out — never performed — so the honest
// verdict is unavailable. The extension files are created in the user
// directory (no root) and removed after the probe either way (T-04-04).
func probeMicroExtension(ctx context.Context) (bool, error) {
	dir, err := microExtensionDir()
	if err != nil {
		return false, err
	}
	if err := writeMicroExtension(dir); err != nil {
		_ = os.RemoveAll(dir)

		return false, err
	}
	defer func() { _ = os.RemoveAll(dir) }()

	list, _ := runCmd(ctx, "/usr/bin/gnome-extensions", "list")
	observed := fmt.Sprintf("extension written to %s; running shell ignores new extensions without a restart"+
		" (gnome-extensions list shows goswitch-probe: %v);"+
		" enabling requires a GNOME Shell restart = re-login (Wayland)",
		dir, strings.Contains(list, "goswitch-probe"))

	return false, appendJournalRow("5. micro-extension "+time.Now().Format("15:04:05"),
		"mkdir ~/.local/share/gnome-shell/extensions/"+probeUUID+" (~30 lines JS)",
		observed, "Verdict: unavailable — требует ре-логина сессии (не выполнен)")
}

// microExtensionDir resolves the user-local extension directory.
func microExtensionDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home: %w", err)
	}

	return filepath.Join(home, ".local", "share", "gnome-shell", "extensions", probeUUID), nil
}

// writeMicroExtension writes the metadata and the ~15-line probe extension.
func writeMicroExtension(dir string) error {
	if err := os.MkdirAll(dir, probeDirPerm); err != nil {
		return fmt.Errorf("create extension dir: %w", err)
	}
	metadata := `{"name":"goswitch-probe","uuid":"` + probeUUID +
		`","shell-version":["46"],"description":"D-01 experiment probe (removed after the run)"}` +
		"\n"
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte(metadata), probeFilePerm); err != nil {
		return fmt.Errorf("write metadata.json: %w", err)
	}
	const extension = `// goswitch D-01 probe: activate input source 1 when enabled.
const Main = imports.ui.main;

function enable() {
    const ism = Main.getInputSourceManager();
    if (ism && ism.inputSources.length > 1) {
        ism.inputSources[1].activate();
    }
}

function disable() {}
`
	if err := os.WriteFile(filepath.Join(dir, "extension.js"), []byte(extension), probeFilePerm); err != nil {
		return fmt.Errorf("write extension.js: %w", err)
	}

	return nil
}

// resetToGoswitchEN puts goswitch-en back as the active engine (IBus
// SetGlobalEngine — the activation GNOME 46 honors) and waits for its
// focus_in so the next probe starts from a known state. An engine already
// active is left alone: re-setting it would be a no-op that never emits a
// new focus_in, and the wait would time out on nothing.
func resetToGoswitchEN(ctx context.Context, s *stand) error {
	if cur, err := runCmd(ctx, "ibus", "engine"); err == nil && strings.TrimSpace(cur) == "goswitch-en" {
		return nil
	}
	before := s.countSub(focusInEN)
	if _, err := runCmd(ctx, "ibus", "engine", "goswitch-en"); err != nil {
		return fmt.Errorf("reset to goswitch-en: %w", err)
	}

	return s.waitForNew(ctx, focusInEN, before+1, focusWait)
}

// withGoswitchSources prepends both goswitch engines to a snapshotted
// sources value, keeping the original entries after them. The snapshot is
// the literal gsettings print (`[('xkb', 'us'), ('xkb', 'ru')]`), so both
// brackets must be stripped before re-wrapping — a stray bracket makes the
// whole write a GVariant syntax error and the probe would measure nothing.
func withGoswitchSources(snapshot string) string {
	body := strings.TrimSpace(snapshot)
	body = strings.TrimSpace(strings.TrimPrefix(body, "["))
	body = strings.TrimSpace(strings.TrimSuffix(body, "]"))
	if body == "" {
		return goswitchSourcesPrefix
	}

	return strings.TrimSuffix(goswitchSourcesPrefix, "]") + ", " + body + "]"
}

// d01Obs is one probe's observable outcome: how many NEW focus records each
// of our engines logged after the probe action, plus the readback of the
// text typed afterwards and any probe-specific extra observation.
type d01Obs struct {
	outEN int
	inRU  int
	typed bool // whether the text probe ran at all (action failures skip it)
	text  string
	extra string
}

// switched is the two-observation criterion: the FocusOut(en) → FocusIn(ru)
// pair in the daemon log AND Cyrillic script from the subsequent typing
// (ghbdtn → привет means the RU XKB group is actually applied).
func (o d01Obs) switched() bool {
	return o.outEN >= 1 && o.inRU >= 1 && hasCyrillic(o.text)
}

// record appends the probe's journal row and reports the switched verdict.
func (o d01Obs) record(probe, cmd string) (bool, error) {
	return o.recordVerdict(probe, cmd, "")
}

// recordVerdict appends the row with an optional verdict override — the
// Shell.Eval probe reports unavailable (mechanism refused), not
// not-switched (mechanism exercised and failed).
func (o d01Obs) recordVerdict(probe, cmd, override string) (bool, error) {
	observed := fmt.Sprintf("focus_out(en)+%d, focus_in(ru)+%d", o.outEN, o.inRU)
	if o.typed {
		observed += fmt.Sprintf("; typed %q -> %q", d01Text, o.text)
	}
	if o.extra != "" {
		observed += "; " + o.extra
	}
	verdict := "Verdict: not-switched"
	if o.switched() {
		verdict = "Verdict: switched (двойное наблюдение: фокус-пара + кириллический текст)"
	}
	if override != "" {
		verdict = override
	}

	return o.switched(), appendJournalRow(probe, cmd, observed, verdict)
}

// observeSwitch runs one probe action inside the uniform observation flow:
// baseline focus counters, action, bounded wait for a NEW goswitch-ru
// focus_in, then typing the probe text into the stand's entry surface and
// reading it back (zenity prints its text on close — the full-script oracle;
// the locked-session shell fallback yields a length-only note).
func (s *stand) observeSwitch(ctx context.Context, action func() error) d01Obs {
	var obs d01Obs

	kind, err := s.openEntrySurface(ctx)
	if err != nil {
		obs.extra = "surface unavailable: " + err.Error()

		return obs
	}
	outBefore := s.countSub(focusOutEN)
	ruBefore := s.countSub(focusInRU)

	if err := action(); err != nil {
		obs.extra = "action failed: " + err.Error()
		_ = s.closeEntrySurface(ctx, kind)

		return obs
	}

	// Bounded wait for the switch signal; a timeout is the observation.
	if err := s.waitForNew(ctx, focusInRU, ruBefore+1, d01ProbeWait); err != nil {
		obs.extra = strings.TrimSpace("no new focus_in(goswitch-ru) within " + d01ProbeWait.String() +
			"; " + obs.extra)
	}
	obs.outEN = s.countSub(focusOutEN) - outBefore
	obs.inRU = s.countSub(focusInRU) - ruBefore

	if err := s.injectText(ctx, d01Text); err != nil {
		obs.extra = strings.TrimSpace("typing failed: " + err.Error() + "; " + obs.extra)
	}
	obs.text = s.readBackText(ctx, kind)
	obs.typed = true

	return obs
}

// readBackText collects the typed text from the entry surface. The zenity
// oracle is the full script (Enter commits and prints it); the shell-entry
// fallback can only witness length, which the note says explicitly.
func (s *stand) readBackText(ctx context.Context, kind surfaceKind) string {
	switch kind {
	case surfaceZenity:
		out, err := s.closeZenity(ctx)
		if err != nil {
			return fmt.Sprintf("(close failed: %v)", err)
		}

		return out
	case surfaceShell:
		after, err := s.witnessChars(ctx)
		_ = s.pressKey(ctx, "Escape")

		return fmt.Sprintf("(length-only oracle: chars=%d, err=%v)", after, err)
	case surfaceChromium:
		// Driver-managed surface (surface.go): readback goes through the
		// focused-text bridge, not the entry-surface oracle.
		return ""
	}

	return ""
}

// hasCyrillic reports whether s contains any Cyrillic rune.
func hasCyrillic(s string) bool {
	for _, r := range s {
		if r >= 0x0400 && r <= 0x04FF {
			return true
		}
	}

	return false
}

// appendJournalRow appends one evidence row to the D-01 experiment journal
// (T-04-03: the journal is written by the case itself, command + observed +
// verdict per line).
func appendJournalRow(probe, cmd, observed, verdict string) error {
	f, err := os.OpenFile(d01Journal, os.O_APPEND|os.O_CREATE|os.O_WRONLY, probeFilePerm)
	if err != nil {
		return fmt.Errorf("open journal %s: %w", d01Journal, err)
	}
	defer f.Close() //nolint:errcheck // append-only evidence log; the row write's error is the one that matters
	observed = strings.NewReplacer("|", "/").Replace(observed)
	row := fmt.Sprintf("| %s | `%s` | %s | %s |\n", probe, cmd, observed, verdict)
	if _, err := f.WriteString(row); err != nil {
		return fmt.Errorf("append journal row: %w", err)
	}

	return nil
}
