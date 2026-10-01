package install_test

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/Djarvur/goswitch/engine"
	"github.com/Djarvur/goswitch/internal/install"
)

// The corpus's desktop stand-ins: the gsettings raw values (the exact shape
// gsettings get prints and set accepts) and the list-engine reply that
// proves the component entered the registry. ownerSwitchBindings is the
// live switch-input-source binding of the owner's desktop (verified
// 2026-09-27) — the value install snapshots and uninstall restores.
const (
	ownerSources                = `[('xkb', 'us'), ('xkb', 'ru')]`
	singleUSSources             = `[('xkb', 'us')]`
	goswitchSources             = `[('ibus', 'goswitch-en')]`
	fallbackSources             = `[('xkb', 'us')]`
	wrappedSources              = `[('ibus', 'goswitch-en'), ('ibus', 'goswitch-ru')]`
	sourcesUSFR                 = `[('xkb', 'us'), ('xkb', 'fr')]`
	ownerSwitchBindings         = `['<Super>space', 'XF86Keyboard']`
	ownerSwitchBindingsBackward = `['<Shift><Super>space', '<Shift>XF86Keyboard']`
	// junkSwitchBinding is the junk live value the over-install's
	// switch-binding gets answer — it must never reach the state file nor
	// any `gsettings set` (ADR-006: install writes nothing to the chords).
	junkSwitchBinding              = `[]`
	fallbackSwitchBindings         = `['<Super>space', 'XF86Keyboard']`
	fallbackSwitchBindingsBackward = `['<Shift><Super>space', '<Shift>XF86Keyboard']`
	listEngineOut                  = "goswitch-en - goswitch English (US)\ngoswitch-ru - goswitch Русская\n"
)

// The pinned binary/operation names the corpus asserts on (goconst: named
// once instead of repeated literals).
const (
	binGSettings = "gsettings"
	binIbus      = "ibus"
	binSystemctl = "systemctl"
	opWriteCache = "write-cache"
	opRestart    = "restart"
	opGet        = "get"
	opSet        = "set"
	// opDaemonReload is the systemctl daemon-reload argv the sequence and
	// the rollback corpora pin (goconst: named once).
	opDaemonReload = "--user daemon-reload"
	engineENName   = "goswitch-en"
	engineRUName   = "goswitch-ru"
)

// The gsettings schema keys of the input sources (D-40's single-owner
// takeover writes and D-42's restore rewrites the sources; the
// wm.keybindings schema hosts the GNOME layout-switch chords install must
// leave untouched).
const (
	gsettingsSchema = "org.gnome.desktop.input-sources"
	gsettingsKey    = "sources"
	// gsettingsKeybindingsSchema hosts the GNOME layout-switch chords (live
	// finding 2026-09-27: NOT desktop.input-sources — that schema carries
	// no switch key at all). ADR-006 two-source: install leaves both
	// bindings untouched — the chords cycle the two goswitch engines as a
	// first-class visible switch — uninstall restores the saved values.
	gsettingsKeybindingsSchema = "org.gnome.desktop.wm.keybindings"
	gsettingsKeySwitch         = "switch-input-source"
	gsettingsKeySwitchBackward = "switch-input-source-backward"
	// derivedActivation is the engine name the first ('xkb', 'us') tuple of
	// ownerSources restores to (the fallbackEngine derivation).
	derivedActivation = "xkb:us::eng"
)

// instCall is one recorded subprocess invocation: argv plus the env the
// child ran with (the env is the observable surface of the Pitfall 1 pin —
// every write-cache must carry IBUS_COMPONENT_PATH).
type instCall struct {
	name string
	args []string
	env  []string
}

// fakeRunner records every subprocess the installer launches and answers
// through a per-test stub (default: the happy-path desktop above).
type fakeRunner struct {
	mu    sync.Mutex
	calls []instCall
	stub  func(name string, args []string) ([]byte, error)
}

// run records the invocation and answers it through the stub.
func (f *fakeRunner) run(_ context.Context, name string, args, env []string, _ []byte) ([]byte, error) {
	f.mu.Lock()
	f.calls = append(f.calls, instCall{name: name, args: slices.Clone(args), env: slices.Clone(env)})
	stub := f.stub
	f.mu.Unlock()
	if stub != nil {
		return stub(name, args)
	}

	return defaultReply(name, args)
}

// snapshot copies the recorded calls under the guard.
func (f *fakeRunner) snapshot() []instCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	return slices.Clone(f.calls)
}

// defaultReply answers the happy-path desktop: gsettings get returns the
// owner sources and the owner's live switch binding, ibus list-engine
// already lists goswitch.
func defaultReply(name string, args []string) ([]byte, error) {
	gsettingsGet := func(key string) bool {
		return name == binGSettings && len(args) == 3 && args[0] == opGet && args[2] == key
	}
	switch {
	case gsettingsGet(gsettingsKey):
		return []byte(ownerSources), nil
	case gsettingsGet(gsettingsKeySwitch):
		return []byte(ownerSwitchBindings), nil
	case gsettingsGet(gsettingsKeySwitchBackward):
		return []byte(ownerSwitchBindingsBackward), nil
	case name == binIbus && len(args) > 0 && args[0] == "list-engine":
		return []byte(listEngineOut), nil
	}

	return nil, nil
}

// newInstaller builds the installer over the fake runner, a fake $HOME and
// a self dir whose goswitchd stand-in exists. The registry probe defaults
// to hit (the happy-path desktop) — the retry corpus scripts miss/hit by
// overriding it.
func newInstaller(t *testing.T, f *fakeRunner, home, selfDir string, engines []string) *install.Installer {
	t.Helper()

	return install.New(
		install.WithRunner(f.run),
		install.WithHome(home),
		install.WithSelfDir(selfDir),
		install.WithActiveEngines(func(context.Context) ([]string, error) {
			return engines, nil
		}),
		install.WithRegistryProbe(func() bool { return true }),
	)
}

// selfDirWithDaemon returns a temp dir holding a goswitchd stand-in binary
// — the directory the unit's ExecStart resolves to.
func selfDirWithDaemon(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "goswitchd"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write goswitchd stand-in: %v", err)
	}

	return dir
}

// installPaths are the three $HOME artifacts of an install.
func installPaths(home string) (xmlPath, unitPath, statePath string) {
	return filepath.Join(home, ".config", "ibus", "component", "goswitch.xml"),
		filepath.Join(home, ".config", "systemd", "user", "goswitchd.service"),
		filepath.Join(home, ".local", "share", "goswitch", "install-state.json")
}

// runInstall is the corpus's one-line driver: install and fail the test on
// any error.
func runInstall(t *testing.T, i *install.Installer) []string {
	t.Helper()
	report, err := i.Install(context.Background())
	if err != nil {
		t.Fatalf("Install() err = %v, want nil (report %v)", err, report)
	}

	return report
}

// assertCallSequence pins the exact subprocess order of a phase — the
// D-39/D-42 sequences are the contract the CLI user depends on.
func assertCallSequence(t *testing.T, calls []instCall, want []struct{ name, args string }) {
	t.Helper()
	if len(calls) != len(want) {
		t.Fatalf("subprocess calls = %d, want %d; got %s", len(calls), len(want), joinCalls(calls))
	}
	for i, w := range want {
		if calls[i].name != w.name || strings.Join(calls[i].args, " ") != w.args {
			t.Errorf("call %d = %s %q, want %s %q", i, calls[i].name, calls[i].args, w.name, w.args)
		}
	}
}

// joinCalls renders the recorded calls for failure diagnostics.
func joinCalls(calls []instCall) string {
	parts := make([]string, 0, len(calls))
	for _, c := range calls {
		parts = append(parts, c.name+" "+strings.Join(c.args, " "))
	}

	return strings.Join(parts, " | ")
}

// assertPerm pins a file's permission bits.
func assertPerm(t *testing.T, path string, want fs.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if perm := info.Mode().Perm(); perm != want {
		t.Errorf("%s perm = %o, want %o", path, perm, want)
	}
}

// assertOnlyEntry pins that dir holds exactly one entry named want — the
// no-leftover-temps half of the atomic-write pin.
func assertOnlyEntry(t *testing.T, dir, want string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	if len(entries) != 1 || entries[0].Name() != want {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("dir %s entries = %v, want exactly [%s] (no temp leftovers)", dir, names, want)
	}
}

// TestInstall_Sequence pins the D-54/D-39/D-40 subprocess order end to end
// over the fake desktop: sources + switch-binding snapshot (the pair check
// gates the whole install BEFORE any mutation, D-54) → state save → XML →
// env-carrying write-cache → list-engine verification → unit →
// daemon-reload → enable + restart → ibus restart → live-registration wait →
// live sources re-read → the COMPUTED two-source wrapper (never a
// constant) → engine activation — with NO write to the wm.keybindings
// schema at all (ADR-006 two-source: the GNOME switch chords stay live and
// cycle the two goswitch engines).
func TestInstall_Sequence(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{"xkb:us::eng", engineENName})

	report := runInstall(t, i)

	// The live-true order (first install-cycle run): the cache-file probe
	// verifies write-cache immediately (list-engine stays stale until the
	// restart), and the daemon-view list-engine gate runs AFTER the restart.
	// The takeover re-reads the live value and writes the wrapper computed
	// from the user's own us/ru pair (D-54) — never the old constant.
	assertCallSequence(t, f.snapshot(), []struct{ name, args string }{
		{binGSettings, "get " + gsettingsSchema + " " + gsettingsKey},
		{binGSettings, "get " + gsettingsKeybindingsSchema + " " + gsettingsKeySwitch},
		{binGSettings, "get " + gsettingsKeybindingsSchema + " " + gsettingsKeySwitchBackward},
		{binIbus, opWriteCache},
		{binIbus, opRestart},
		{binIbus, opListEngine},
		{binSystemctl, opDaemonReload},
		{binSystemctl, "--user enable goswitchd"},
		{binSystemctl, "--user restart goswitchd"},
		{binGSettings, "get " + gsettingsSchema + " " + gsettingsKey},
		{binGSettings, "set " + gsettingsSchema + " " + gsettingsKey + " " + wrappedSources},
		{binIbus, "engine goswitch-en"},
	})

	// ADR-006 two-source: install records NO `gsettings set` against the
	// wm.keybindings schema — the GNOME switch chords stay untouched.
	assertSwitchBindingsUntouched(t, f.snapshot())

	xmlPath, unitPath, statePath := installPaths(home)
	assertSequenceState(t, xmlPath, unitPath, statePath, report)
}

// assertSequenceState pins the file/report tail of TestInstall_Sequence:
// all three artifacts exist, the state carries the read switch bindings
// VERBATIM (decoded — json.Marshal HTML-escapes the '<'/'>' of the raw
// binding bytes; the restore path reads the fields back through the same
// decode), and the report names both the preserved chords (left untouched,
// saved for uninstall) and the two-source wrap verdict.
func assertSequenceState(t *testing.T, xmlPath, unitPath, statePath string, report []string) {
	t.Helper()
	for _, p := range []string{xmlPath, unitPath, statePath} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("install artifact %s missing: %v", p, err)
		}
	}
	var st struct {
		Sources           string `json:"sources"`
		SwitchInputSource string `json:"switch_input_source"`
	}
	if err := json.Unmarshal([]byte(readAll(t, statePath)), &st); err != nil {
		t.Fatalf("parse install state: %v", err)
	}
	if st.SwitchInputSource != ownerSwitchBindings {
		t.Errorf("state switch_input_source = %q, want the pre-install binding VERBATIM %q",
			st.SwitchInputSource, ownerSwitchBindings)
	}
	var stBw struct {
		SwitchInputSourceBackward string `json:"switch_input_source_backward"`
	}
	if err := json.Unmarshal([]byte(readAll(t, statePath)), &stBw); err != nil {
		t.Fatalf("parse install state (backward): %v", err)
	}
	if stBw.SwitchInputSourceBackward != ownerSwitchBindingsBackward {
		t.Errorf("state switch_input_source_backward = %q, want the pre-install binding VERBATIM %q",
			stBw.SwitchInputSourceBackward, ownerSwitchBindingsBackward)
	}
	if !slices.ContainsFunc(report, func(line string) bool {
		return strings.Contains(line, "switch-input-source: left untouched (saved for uninstall)")
	}) {
		t.Errorf("report %v does not name the preserved switch binding", report)
	}
	if !slices.ContainsFunc(report, func(line string) bool {
		return strings.Contains(line, "sources: wrapped ("+engineENName+", "+engineRUName+")")
	}) {
		t.Errorf("report %v does not name the wrapped engines", report)
	}
}

// TestInstall_SequenceSingleSource pins the owner decision 2026-09-30 (one
// source in the GNOME switcher) on the FULL sequence: a desktop carrying a
// single xkb 'us' source installs into exactly [('ibus', 'goswitch-en')] —
// one `gsettings set sources` carrying the single goswitch engine, the
// explicit engine activation kept, the switch chords still untouched, and
// the report naming the one wrapped engine.
func TestInstall_SequenceSingleSource(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	f.stub = func(name string, args []string) ([]byte, error) {
		if name == binGSettings && len(args) == 3 && args[0] == opGet && args[2] == gsettingsKey {
			return []byte(singleUSSources), nil
		}

		return defaultReply(name, args)
	}
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})

	report := runInstall(t, i)

	// The same live-true order as the pair desktop — only the takeover's
	// set carries the single-source wrapper.
	assertCallSequence(t, f.snapshot(), []struct{ name, args string }{
		{binGSettings, "get " + gsettingsSchema + " " + gsettingsKey},
		{binGSettings, "get " + gsettingsKeybindingsSchema + " " + gsettingsKeySwitch},
		{binGSettings, "get " + gsettingsKeybindingsSchema + " " + gsettingsKeySwitchBackward},
		{binIbus, opWriteCache},
		{binIbus, opRestart},
		{binIbus, opListEngine},
		{binSystemctl, opDaemonReload},
		{binSystemctl, "--user enable goswitchd"},
		{binSystemctl, "--user restart goswitchd"},
		{binGSettings, "get " + gsettingsSchema + " " + gsettingsKey},
		{binGSettings, "set " + gsettingsSchema + " " + gsettingsKey + " " + goswitchSources},
		{binIbus, "engine goswitch-en"},
	})

	// ADR-006: install records NO `gsettings set` against the wm.keybindings
	// schema — the single-source model changes nothing about the chords.
	assertSwitchBindingsUntouched(t, f.snapshot())

	if !slices.ContainsFunc(report, func(line string) bool {
		return strings.Contains(line, "sources: wrapped ("+engineENName+")")
	}) {
		t.Errorf("report %v does not name the single wrapped engine", report)
	}

	// The single-source desktop's pre-install value is the uninstall
	// restore's material — the state file carries it verbatim.
	_, _, statePath := installPaths(home)
	if got := readAll(t, statePath); !strings.Contains(got, singleUSSources) {
		t.Errorf("state %q does not carry the single-source pre-install value %q", got, singleUSSources)
	}
}

// assertSwitchBindingsUntouched pins the ADR-006 two-source verdict on the
// recorded calls: install records NO `gsettings set` against
// gsettingsKeybindingsSchema — neither switch-input-source nor
// switch-input-source-backward is written, so the GNOME switch chords stay
// live and cycle the two goswitch engines as a first-class visible switch
// (the daemon follows external flips via the 05-04 sync listener).
func assertSwitchBindingsUntouched(t *testing.T, calls []instCall) {
	t.Helper()
	for _, c := range calls {
		if c.name == binGSettings && len(c.args) >= 3 && c.args[0] == opSet && c.args[1] == gsettingsKeybindingsSchema {
			t.Errorf("install recorded gsettings set %s %s — the GNOME switch chords must stay"+
				" untouched (ADR-006 two-source)", c.args[1], c.args[2])
		}
	}
}

// TestInstall_EveryWriteCacheCarriesEnv pins Pitfall 1 as a unit-testable
// contract: EVERY recorded `ibus write-cache` runs with
// IBUS_COMPONENT_PATH pointing at the user component dir — a bare
// write-cache evicts user components from the registry cache (live-verified
// research).
func TestInstall_EveryWriteCacheCarriesEnv(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})

	runInstall(t, i)

	// The env REPLACES the scan path: the value must carry the user dir
	// AND the system dir (a user-only value strips the system components
	// from the registry — live finding, first install-cycle run).
	want := "IBUS_COMPONENT_PATH=" + filepath.Join(home, ".config", "ibus", "component") +
		":/usr/share/ibus/component"
	caches := 0
	for _, c := range f.snapshot() {
		if c.name != "ibus" || len(c.args) == 0 || c.args[0] != "write-cache" {
			continue
		}
		caches++
		if !slices.Contains(c.env, want) {
			t.Errorf("write-cache env lacks %q (got %v) — every cache write must pin the component path", want, c.env)
		}
	}
	if caches == 0 {
		t.Fatal("no ibus write-cache recorded — install must refresh the registry cache")
	}
}

// TestInstall_ComponentXMLMirrorsWireIdentity pins Pattern 2: the XML
// byte-carries the SAME identity goswitchd registers at runtime (built here
// through engine.NewComponent/NewEngineDesc so a wire change fails this
// corpus), the <exec> stays EMPTY (systemd is the sole supervisor — a
// binary path would double-spawn against the ctlsvc single-instance guard),
// and the file lands 0644 with no temp leftovers (atomic write).
func TestInstall_ComponentXMLMirrorsWireIdentity(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})

	runInstall(t, i)

	xmlPath, _, _ := installPaths(home)
	data, err := os.ReadFile(xmlPath)
	if err != nil {
		t.Fatalf("read component XML: %v", err)
	}
	xml := string(data)

	wireEngines := []engine.EngineDesc{
		engine.NewEngineDesc("goswitch-en", "goswitch English (US)", "en", "us", "en"),
		engine.NewEngineDesc("goswitch-ru", "goswitch Русская", "ru", "ru", "ru"),
	}
	wire := engine.NewComponent(wireEngines)
	for _, want := range []string{
		wire.ComponentName, wire.Description, wire.Version, wire.License, wire.Author, wire.Homepage,
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("component XML misses the wire identity value %q", want)
		}
	}
	for _, e := range wireEngines {
		for _, want := range []string{e.EngineName, e.LongName, e.Language, e.Layout} {
			if !strings.Contains(xml, want) {
				t.Errorf("component XML misses the wire engine value %q", want)
			}
		}
		// Owner decision 1 (quick plan 260927-way): every engine desc names
		// the mode-indicator property — the XML mirrors the wire
		// EngineDesc.icon_prop_key end to end.
		if e.IconPropKey != "InputMode" {
			t.Errorf("wire engine %q IconPropKey = %q, want %q", e.EngineName, e.IconPropKey, "InputMode")
		}
		if !strings.Contains(xml, "<icon_prop_key>InputMode</icon_prop_key>") {
			t.Errorf("component XML misses the panel icon key <icon_prop_key>InputMode</icon_prop_key>" +
				" — it must reach the on-disk identity")
		}
	}
	if !strings.Contains(xml, "<exec></exec>") {
		t.Error("component XML <exec> is not empty — systemd must stay the sole supervisor (Pattern 2)")
	}
	if strings.Contains(xml, "goswitchd") {
		t.Error("component XML mentions goswitchd — an exec path would double-spawn the daemon")
	}

	assertPerm(t, xmlPath, 0o644)
	assertOnlyEntry(t, filepath.Dir(xmlPath), "goswitch.xml")
}

// TestInstall_UnitAbsoluteExecStart pins the ASVS V14 unit contract: the
// ExecStart is the ABSOLUTE goswitchd path next to the running binary (no
// %h, no $PATH lookup, no shell interpolation), while the dist template's
// ordering semantics survive (PartOf/After/Restart/WantedBy).
func TestInstall_UnitAbsoluteExecStart(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	selfDir := selfDirWithDaemon(t)
	i := newInstaller(t, f, home, selfDir, []string{engineENName})

	runInstall(t, i)

	_, unitPath, _ := installPaths(home)
	data, err := os.ReadFile(unitPath)
	if err != nil {
		t.Fatalf("read user unit: %v", err)
	}
	unit := string(data)

	wantExec := "ExecStart=" + filepath.Join(selfDir, "goswitchd")
	if !strings.Contains(unit, wantExec) {
		t.Errorf("unit lacks the absolute %s (got %q)", wantExec, unit)
	}
	for _, want := range []string{
		"PartOf=graphical-session.target",
		"After=org.freedesktop.IBus.session.GNOME.service",
		"Restart=on-failure",
		"WantedBy=graphical-session.target",
	} {
		if !strings.Contains(unit, want) {
			t.Errorf("unit lost the dist-template semantic %q", want)
		}
	}
	if strings.Contains(unit, "%h") || strings.Contains(unit, "${") {
		t.Error("unit carries %h or shell interpolation — the resolved absolute path is the V14 contract")
	}

	assertPerm(t, unitPath, 0o644)
}

// TestInstall_DaemonBinaryMissing pins the preflight refusal: no goswitchd
// next to the binary → a named error with a fix hint and ZERO subprocess
// calls (the desktop is untouched).
func TestInstall_DaemonBinaryMissing(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	selfDir := t.TempDir() // no goswitchd inside
	i := newInstaller(t, f, home, selfDir, []string{engineENName})

	report, err := i.Install(context.Background())
	if err == nil {
		t.Fatalf("Install() over a missing goswitchd = (%v, nil), want a named error", report)
	}
	if !strings.Contains(err.Error(), "goswitchd") {
		t.Errorf("error %q does not name goswitchd for the fix hint", err)
	}
	if calls := f.snapshot(); len(calls) != 0 {
		t.Errorf("missing goswitchd still ran %d subprocess calls (%s) — the refusal must precede any desktop change",
			len(calls), joinCalls(calls))
	}
}

// TestInstall_StateFileOutsideConfigDir pins Pitfall 8: the saved sources
// live in ~/.local/share/goswitch/install-state.json (0600) — NEVER in
// ~/.config/goswitch, which D-42 preserves on default uninstall and --purge
// deletes (a backup there could not survive the uninstall that needs it).
func TestInstall_StateFileOutsideConfigDir(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})

	runInstall(t, i)

	_, _, statePath := installPaths(home)
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read install state: %v", err)
	}
	if !strings.Contains(string(data), ownerSources) {
		t.Errorf("state %q does not carry the pre-install sources VERBATIM (%q)", data, ownerSources)
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "goswitch", "install-state.json")); !os.IsNotExist(err) {
		t.Error("state found inside ~/.config/goswitch — Pitfall 8 places it in ~/.local/share/goswitch")
	}
	assertPerm(t, statePath, 0o600)
}

// TestInstall_SecondInstallKeepsOriginalBackup pins the idempotency core:
// an existing state file is never overwritten — the FIRST install's backup
// is sacred, so repeated installs cannot destroy the pre-goswitch desktop.
// The over-install reads a junk live value for the switch-binding gets and
// must never save it over the original — nor write it anywhere (ADR-006:
// install touches no chord). The 05-02 upgrade form extends the rule
// further: the takeover still runs and computes the two-source wrapper FROM
// THE SAVED ORIGINAL (never from the live goswitch-only value), while the
// backup stays byte-intact.
func TestInstall_SecondInstallKeepsOriginalBackup(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})

	runInstall(t, i)

	_, _, statePath := installPaths(home)
	// The planted original is the owner's real pair — shape-valid, so the
	// D-54 upgrade path may compute the wrapper from it.
	marker := `{"sources":"` + ownerSources + `"}`
	if err := os.WriteFile(statePath, []byte(marker), 0o600); err != nil {
		t.Fatalf("plant the marker state: %v", err)
	}

	// The second install sees the post-takeover desktop (goswitch-only
	// sources) and a junk live value for the switch-binding gets — exactly
	// the state it must NOT save over the marker.
	f.stub = func(name string, args []string) ([]byte, error) {
		if name == binGSettings && len(args) == 3 && args[0] == opGet {
			if args[2] == gsettingsKey {
				return []byte(goswitchSources), nil
			}
			if args[2] == gsettingsKeySwitch || args[2] == gsettingsKeySwitchBackward {
				return []byte(junkSwitchBinding), nil
			}
		}

		return defaultReply(name, args)
	}
	runInstall(t, i)

	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("re-read state: %v", err)
	}
	if string(data) != marker {
		t.Errorf("second install overwrote the original backup: state = %q, want %q byte-intact"+
			" (the junk live binding must never be saved over it)", data, marker)
	}

	// ADR-006: NEITHER install writes a chord — the junk live binding value
	// must never reach `gsettings set` either.
	assertSwitchBindingsUntouched(t, f.snapshot())

	// The upgrade form (D-54): the takeover computes the wrapper from the
	// saved original and writes it — the desktop leaves the phase-4
	// single-owner form for the two-source wrapper in this same install.
	assertWrapperFromSavedOriginal(t, f.snapshot())
}

// assertWrapperFromSavedOriginal pins the D-54 upgrade verdict: exactly the
// wrapper computed from the SAVED original reaches `gsettings set …
// sources`, never the live goswitch-only value and never the old constant.
func assertWrapperFromSavedOriginal(t *testing.T, calls []instCall) {
	t.Helper()
	wrapperSet := false
	for _, c := range calls {
		if c.name == binGSettings && len(c.args) == 4 && c.args[0] == opSet && c.args[2] == gsettingsKey {
			wrapperSet = true
			if c.args[3] != wrappedSources {
				t.Errorf("upgrade set %q, want the wrapper computed from the SAVED original %q",
					c.args[3], wrappedSources)
			}
		}
	}
	if !wrapperSet {
		t.Error("second install recorded no sources set — the upgrade must write the two-source wrapper")
	}
}

// TestUninstall_FullRollback pins the D-42 sequence: stop → disable → unit
// removed → daemon-reload → XML removed → env write-cache → ibus restart →
// sources restored VERBATIM → best-effort activation of the restored source
// → state removed; the user's ~/.config/goswitch is untouched.
func TestUninstall_FullRollback(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})

	runInstall(t, i)

	userCfgDir := filepath.Join(home, ".config", "goswitch")
	if err := os.MkdirAll(userCfgDir, 0o755); err != nil {
		t.Fatalf("mkdir user config: %v", err)
	}
	userCfg := filepath.Join(userCfgDir, "config.yaml")
	if err := os.WriteFile(userCfg, []byte("hotkeys: {}\n"), 0o600); err != nil {
		t.Fatalf("write user config: %v", err)
	}

	report, err := i.Uninstall(context.Background(), false)
	if err != nil {
		t.Fatalf("Uninstall() err = %v (report %v), want the full rollback to succeed", err, report)
	}

	// The verbatim snapshot's restore: the saved switch bindings go back
	// BEFORE the state file dies, and the report names the restored values.
	assertCallSequence(t, uninstallCalls(t, f), []struct{ name, args string }{
		{binSystemctl, "--user stop goswitchd"},
		{binSystemctl, "--user disable goswitchd"},
		{binSystemctl, opDaemonReload},
		{binIbus, opWriteCache},
		{binIbus, opRestart},
		{binGSettings, "set " + gsettingsSchema + " " + gsettingsKey + " " + ownerSources},
		{binIbus, "engine " + derivedActivation},
		{binGSettings, "set " + gsettingsKeybindingsSchema + " " + gsettingsKeySwitch + " " + ownerSwitchBindings},
		{binGSettings, "set " + gsettingsKeybindingsSchema + " " +
			gsettingsKeySwitchBackward + " " + ownerSwitchBindingsBackward},
	})
	if !slices.ContainsFunc(report, func(line string) bool {
		return strings.Contains(line, "switch-input-source: restored "+ownerSwitchBindings)
	}) {
		t.Errorf("report %v does not name the restored switch binding", report)
	}
	if !slices.ContainsFunc(report, func(line string) bool {
		return strings.Contains(line, "switch-input-source-backward: restored "+ownerSwitchBindingsBackward)
	}) {
		t.Errorf("report %v does not name the restored backward switch binding", report)
	}

	xmlPath, unitPath, statePath := installPaths(home)
	for _, p := range []string{xmlPath, unitPath, statePath} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("uninstall artifact %s still present", p)
		}
	}
	if _, err := os.Stat(userCfg); err != nil {
		t.Errorf("user config %s touched by default uninstall (D-42 preserves it): %v", userCfg, err)
	}
}

// installCallCount is the subprocess count of one happy-path install (the
// FullRollback corpus asserts the uninstall suffix of the recording): three
// gsettings gets (sources + the two switch bindings) at snapshot time, two
// ibus cache steps, three systemctl steps, list-engine, the takeover's
// live sources re-read, the wrapper set (ADR-006: no chord writes),
// engine activation.
const installCallCount = 12

// uninstallCalls returns the recording suffix after one happy-path install
// — the uninstall phase's own calls. Fails the test when install itself did
// not run (the corpus never asserts a phase that never executed).
func uninstallCalls(t *testing.T, f *fakeRunner) []instCall {
	t.Helper()
	calls := f.snapshot()
	if len(calls) < installCallCount {
		t.Fatalf("install phase recorded %d calls, want >= %d — install did not run (%s)",
			len(calls), installCallCount, joinCalls(calls))
	}

	return calls[installCallCount:]
}

// corruptSwitchCase is one planted-state case of
// TestUninstall_CorruptStateFallsBack: the state file content and the
// expected restore targets of both bindings (wantFallback pins the visible
// substitution in the report).
type corruptSwitchCase struct {
	plant        string
	wantSources  string
	wantSwitch   string
	wantSwitchBw string
	wantFallback bool
}

// corruptSwitchCases is the planted-state table — extracted so the test
// itself stays under the funlen ceiling.
func corruptSwitchCases() map[string]corruptSwitchCase {
	return map[string]corruptSwitchCase{
		"garbage value": {
			plant:        `{"sources":"garbage"}`,
			wantSources:  fallbackSources,
			wantSwitch:   fallbackSwitchBindings, // field absent → distro default
			wantSwitchBw: fallbackSwitchBindingsBackward,
			wantFallback: true,
		},
		"unparsable": {
			plant:        "\x00not json",
			wantSources:  fallbackSources,
			wantSwitch:   fallbackSwitchBindings,
			wantSwitchBw: fallbackSwitchBindingsBackward,
			wantFallback: true,
		},
		"switch field absent (pre-batch JSON)": {
			plant:        `{"sources":"` + ownerSources + `"}`,
			wantSources:  ownerSources, // valid sources restore VERBATIM
			wantSwitch:   fallbackSwitchBindings,
			wantSwitchBw: fallbackSwitchBindingsBackward,
			wantFallback: true,
		},
		"switch field garbage": {
			plant:        `{"sources":"` + ownerSources + `","switch_input_source":"garbage"}`,
			wantSources:  ownerSources,
			wantSwitch:   fallbackSwitchBindings,
			wantSwitchBw: fallbackSwitchBindingsBackward,
			wantFallback: true,
		},
		"backward field garbage": {
			plant: `{"sources":"` + ownerSources + `","switch_input_source":"` + ownerSwitchBindings +
				`","switch_input_source_backward":"garbage"}`,
			wantSources:  ownerSources,
			wantSwitch:   fallbackSwitchBindings,
			wantSwitchBw: fallbackSwitchBindingsBackward,
			wantFallback: true,
		},
	}
}

// TestUninstall_CorruptStateFallsBack pins the ASVS V5 guard (T-04-01-02):
// state values that fail the shape check never reach gsettings verbatim —
// the restore substitutes the safe fallbacks, REPORTS them, and the
// rollback still completes (a corrupt backup must not brick the keyboard
// or abort the uninstall). The same discipline covers the switch bindings:
// a state file whose switch field is absent (a pre-batch JSON) or malformed
// restores the DISTRO DEFAULT binding with the substitution reported.
func TestUninstall_CorruptStateFallsBack(t *testing.T) {
	for name, tc := range corruptSwitchCases() {
		t.Run(name, func(t *testing.T) {
			f := &fakeRunner{}
			home := t.TempDir()
			i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})
			runInstall(t, i)

			_, _, statePath := installPaths(home)
			if err := os.WriteFile(statePath, []byte(tc.plant), 0o600); err != nil {
				t.Fatalf("plant the corrupt state: %v", err)
			}

			report, err := i.Uninstall(context.Background(), false)
			if err != nil {
				t.Fatalf("Uninstall() over a corrupt state err = %v — must not abort (report %v)",
					err, report)
			}

			wantSets := map[string]string{
				gsettingsKey:               tc.wantSources,
				gsettingsKeySwitch:         tc.wantSwitch,
				gsettingsKeySwitchBackward: tc.wantSwitchBw,
			}
			for _, c := range uninstallCalls(t, f) {
				if c.name != binGSettings || len(c.args) != 4 || c.args[0] != "set" {
					continue
				}
				want, ok := wantSets[c.args[2]]
				if !ok {
					continue
				}
				if c.args[3] != want {
					t.Errorf("restore of %s set %q, want %q", c.args[2], c.args[3], want)
				}
				delete(wantSets, c.args[2])
			}
			for key := range wantSets {
				t.Errorf("no gsettings set recorded for %s — the fallback must still restore a usable binding", key)
			}
			if tc.wantFallback && !slices.ContainsFunc(report, func(line string) bool {
				return strings.Contains(line, "fallback")
			}) {
				t.Errorf("report %v does not mention the fallback — the substitution must be visible", report)
			}
		})
	}
}

// readAll is the corpus's file-content reader.
func readAll(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return string(data)
}

// TestInstall_OverInstallIdempotent pins the install-over-install
// contract: the second install over an ALREADY installed desktop (the
// saved state must not become the goswitch-only takeover value) succeeds,
// keeps the FIRST install's backup, and rewrites the XML/unit atomically
// with the same content.
func TestInstall_OverInstallIdempotent(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})

	runInstall(t, i)

	xmlPath, unitPath, statePath := installPaths(home)
	firstXML, firstUnit := readAll(t, xmlPath), readAll(t, unitPath)

	// The second install sees the post-takeover desktop (goswitch-only
	// sources) — exactly what must NOT replace the saved backup.
	f.stub = func(name string, args []string) ([]byte, error) {
		if name == binGSettings && len(args) == 3 && args[0] == opGet && args[2] == gsettingsKey {
			return []byte(goswitchSources), nil
		}

		return defaultReply(name, args)
	}
	runInstall(t, i)

	if got := readAll(t, statePath); !strings.Contains(got, ownerSources) {
		t.Errorf("second install re-saved the state to %q, want the original %q kept", got, ownerSources)
	}
	if got := readAll(t, xmlPath); got != firstXML {
		t.Error("second install rewrote the component XML with different content")
	}
	if got := readAll(t, unitPath); got != firstUnit {
		t.Error("second install rewrote the unit with different content")
	}
}

// TestInstall_WriteCacheRetryOnce pins Pitfall 4's remedy: a cache probe
// miss triggers EXACTLY ONE retry write-cache (both attempts env-carrying),
// a hit after the retry succeeds, and a double miss fails with the named
// error after exactly two attempts.
func TestInstall_WriteCacheRetryOnce(t *testing.T) {
	t.Run("retry once then hit", func(t *testing.T) {
		f := &fakeRunner{}
		home := t.TempDir()
		probes := 0
		i := install.New(
			install.WithRunner(f.run),
			install.WithHome(home),
			install.WithSelfDir(selfDirWithDaemon(t)),
			install.WithActiveEngines(func(context.Context) ([]string, error) {
				return []string{engineENName}, nil
			}),
			install.WithRegistryProbe(func() bool {
				probes++

				return probes >= 2 // first probe misses, the retry's probe hits
			}),
		)
		runInstall(t, i)

		wantEnv := "IBUS_COMPONENT_PATH=" + filepath.Join(home, ".config", "ibus", "component") +
			":/usr/share/ibus/component"
		caches := 0
		for _, c := range f.snapshot() {
			if c.name != binIbus || len(c.args) == 0 || c.args[0] != opWriteCache {
				continue
			}
			caches++
			if !slices.Contains(c.env, wantEnv) {
				t.Errorf("retry write-cache env lacks %q (got %v)", wantEnv, c.env)
			}
		}
		if caches != 2 {
			t.Errorf("write-cache ran %d times, want exactly 2 (the first try + one retry)", caches)
		}
	})

	t.Run("retry exhausted fails named", func(t *testing.T) {
		f := &fakeRunner{}
		home := t.TempDir()
		i := install.New(
			install.WithRunner(f.run),
			install.WithHome(home),
			install.WithSelfDir(selfDirWithDaemon(t)),
			install.WithActiveEngines(func(context.Context) ([]string, error) {
				return []string{engineENName}, nil
			}),
			install.WithRegistryProbe(func() bool { return false }),
		)
		if _, err := i.Install(context.Background()); err == nil {
			t.Fatal("Install() over a never-hit cache = nil error, want the named registry failure")
		}
		caches := 0
		for _, c := range f.snapshot() {
			if c.name == binIbus && len(c.args) > 0 && c.args[0] == opWriteCache {
				caches++
			}
		}
		if caches != 2 {
			t.Errorf("write-cache ran %d times, want exactly 2 (never a third retry)", caches)
		}
	})
}

// TestInstall_UnitHijackOverwritten pins the ASVS V14 guard (T-04-01-01):
// a foreign pre-created goswitchd.service does not survive install — the
// install atomically overwrites it AND verifies the written content
// (read-back), REPORTING the verification verdict; a post-write divergence
// must be an error, never a silent hijack.
func TestInstall_UnitHijackOverwritten(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})

	runInstall(t, i)

	_, unitPath, _ := installPaths(home)
	foreign := "[Service]\nExecStart=/usr/bin/evil --keylog\n"
	if err := os.WriteFile(unitPath, []byte(foreign), 0o600); err != nil {
		t.Fatalf("plant the hijacked unit: %v", err)
	}

	selfDir := filepath.Join(home, "self")
	if err := os.MkdirAll(selfDir, 0o755); err != nil {
		t.Fatalf("mkdir self: %v", err)
	}
	if err := os.WriteFile(filepath.Join(selfDir, "goswitchd"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write self daemon: %v", err)
	}
	i2 := install.New(
		install.WithRunner(f.run),
		install.WithHome(home),
		install.WithSelfDir(selfDir),
		install.WithActiveEngines(func(context.Context) ([]string, error) {
			return []string{engineENName}, nil
		}),
		install.WithRegistryProbe(func() bool { return true }),
	)
	report, err := i2.Install(context.Background())
	if err != nil {
		t.Fatalf("Install() over a hijacked unit err = %v, want the atomic overwrite to succeed", err)
	}

	unit := readAll(t, unitPath)
	if strings.Contains(unit, "evil") {
		t.Errorf("hijacked unit content survived install: %q", unit)
	}
	if !strings.Contains(unit, "ExecStart="+filepath.Join(selfDir, "goswitchd")) {
		t.Errorf("unit lacks the restored absolute ExecStart: %q", unit)
	}
	if !slices.ContainsFunc(report, func(line string) bool { return strings.Contains(line, "verified") }) {
		t.Errorf("report %v carries no read-back verification verdict — the ASVS V14 check must be observable", report)
	}
}

// TestUninstall_Idempotent pins the rollback's tolerance: a second
// uninstall over already-removed artifacts is a green no-op — every
// removal/stop/disable step tolerates the already-gone state.
func TestUninstall_Idempotent(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})

	runInstall(t, i)

	if _, err := i.Uninstall(context.Background(), false); err != nil {
		t.Fatalf("first Uninstall() err = %v, want nil", err)
	}
	report, err := i.Uninstall(context.Background(), false)
	if err != nil {
		t.Fatalf("second Uninstall() err = %v, want a green no-op (report %v)", err, report)
	}

	xmlPath, unitPath, statePath := installPaths(home)
	for _, p := range []string{xmlPath, unitPath, statePath} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("artifact %s reappeared after the second uninstall", p)
		}
	}
}

// TestUninstall_PurgeRemovesUserConfig pins the D-42 purge semantics: the
// default uninstall preserves ~/.config/goswitch (user's own work),
// --purge removes it AND the state directory.
func TestUninstall_PurgeRemovesUserConfig(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})
	userCfg := filepath.Join(home, ".config", "goswitch", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(userCfg), 0o755); err != nil {
		t.Fatalf("mkdir user config: %v", err)
	}
	if err := os.WriteFile(userCfg, []byte("hotkeys: {}\n"), 0o600); err != nil {
		t.Fatalf("write user config: %v", err)
	}

	runInstall(t, i)
	if _, err := i.Uninstall(context.Background(), false); err != nil {
		t.Fatalf("default Uninstall() err = %v, want nil", err)
	}
	if _, err := os.Stat(userCfg); err != nil {
		t.Errorf("default uninstall removed the user's %s — D-42 preserves it", userCfg)
	}

	runInstall(t, i)
	if _, err := i.Uninstall(context.Background(), true); err != nil {
		t.Fatalf("purge Uninstall() err = %v, want nil", err)
	}
	for _, dir := range []string{
		filepath.Join(home, ".config", "goswitch"),
		filepath.Join(home, ".local", "share", "goswitch"),
	} {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Errorf("purge left %s behind — --purge removes it (D-42)", dir)
		}
	}
}

// countSourcesSets counts the recorded `gsettings set … sources` calls.
func countSourcesSets(calls []instCall) int {
	n := 0
	for _, c := range calls {
		if c.name == binGSettings && len(c.args) == 4 && c.args[0] == opSet && c.args[2] == gsettingsKey {
			n++
		}
	}

	return n
}

// hasEngineActivation reports whether the recording carries the explicit
// `ibus engine goswitch-en` activation (the Pitfall-6 half of the
// takeover contract).
func hasEngineActivation(calls []instCall) bool {
	return slices.ContainsFunc(calls, func(c instCall) bool {
		return c.name == binIbus && len(c.args) == 2 && c.args[0] == "engine" && c.args[1] == engineENName
	})
}

// TestInstall_TakeoverSkipsIdenticalWrite pins Pitfall 6 as a unit
// contract: a live value already equal to the computed wrapper receives
// ZERO `gsettings set … sources` calls — a value-identical write
// live-resets the global engine and breaks input right after install —
// while the rest of the sequence (the explicit engine activation among
// it) still runs in its order.
func TestInstall_TakeoverSkipsIdenticalWrite(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})

	runInstall(t, i)
	callsBefore := len(f.snapshot())

	// The follow-up install sees the ALREADY-wrapped desktop: its live
	// value equals the wrapper computed from the saved original — the
	// sources set must be skipped entirely.
	f.stub = func(name string, args []string) ([]byte, error) {
		if name == binGSettings && len(args) == 3 && args[0] == opGet && args[2] == gsettingsKey {
			return []byte(wrappedSources), nil
		}

		return defaultReply(name, args)
	}
	runInstall(t, i)

	second := f.snapshot()[callsBefore:]
	if n := countSourcesSets(second); n != 0 {
		t.Errorf("value-identical takeover issued %d gsettings set sources calls, want 0 (Pitfall 6)", n)
	}
	if !hasEngineActivation(second) {
		t.Error("identical takeover dropped the ibus engine goswitch-en activation — the sequence keeps it")
	}
}

// isMutatingCall reports whether one recorded subprocess mutates the
// desktop: a gsettings set, any systemctl call, an ibus write-cache or
// restart — the atomic-refusal probe of TestInstall_RefusalBeforeAnyWrite.
func isMutatingCall(c instCall) bool {
	gsettingsSet := c.name == binGSettings && len(c.args) > 0 && c.args[0] == opSet
	ibusMutation := c.name == binIbus && len(c.args) > 0 && (c.args[0] == opWriteCache || c.args[0] == opRestart)

	return gsettingsSet || c.name == binSystemctl || ibusMutation
}

// TestInstall_RefusalBeforeAnyWrite pins the atomic refusal gate: a
// desktop goswitch cannot own ENTIRELY — a foreign residue beside the
// wrappable layout (us+fr) — refuses BEFORE any mutating call — zero
// gsettings set, zero systemctl, zero ibus write-cache/restart — and the
// state file is never created, so a refused install leaves the desktop
// byte-identical.
func TestInstall_RefusalBeforeAnyWrite(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	f.stub = func(name string, args []string) ([]byte, error) {
		if name == binGSettings && len(args) == 3 && args[0] == opGet && args[2] == gsettingsKey {
			return []byte(sourcesUSFR), nil
		}

		return defaultReply(name, args)
	}
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})

	report, err := i.Install(context.Background())
	if err == nil {
		t.Fatalf("Install() over the foreign-residue desktop = (%v, nil), want the residue refusal", report)
	}
	if !strings.Contains(err.Error(), "fix") {
		t.Errorf("refusal %q carries no fix hint", err)
	}
	if idx := slices.IndexFunc(f.snapshot(), isMutatingCall); idx >= 0 {
		c := f.snapshot()[idx]
		t.Errorf("refused install still mutated the desktop: %s %v", c.name, c.args)
	}
	_, _, statePath := installPaths(home)
	if _, serr := os.Stat(statePath); !os.IsNotExist(serr) {
		t.Error("refused install created the state file — the refusal must precede the backup write")
	}
}
