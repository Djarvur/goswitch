package install_test

import (
	"context"
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
// proves the component entered the registry.
const (
	ownerSources    = `[('xkb', 'us'), ('xkb', 'ru')]`
	goswitchSources = `[('ibus', 'goswitch-en')]`
	fallbackSources = `[('xkb', 'us')]`
	listEngineOut   = "goswitch-en - goswitch English (US)\ngoswitch-ru - goswitch Русская\n"
	markerSources   = "MARKER-ORIGINAL"
)

// The pinned binary/operation names the corpus asserts on (goconst: named
// once instead of repeated literals).
const (
	binGSettings = "gsettings"
	binIbus      = "ibus"
	binSystemctl = "systemctl"
	opWriteCache = "write-cache"
	engineENName = "goswitch-en"
)

// The gsettings schema key of the input sources (D-40's single-owner
// takeover writes and D-42's restore rewrites it).
const (
	gsettingsSchema = "org.gnome.desktop.input-sources"
	gsettingsKey    = "sources"
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
// owner sources, ibus list-engine already lists goswitch.
func defaultReply(name string, args []string) ([]byte, error) {
	if name == binGSettings && len(args) == 3 && args[0] == "get" && args[2] == gsettingsKey {
		return []byte(ownerSources), nil
	}
	if name == binIbus && len(args) > 0 && args[0] == "list-engine" {
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

// TestInstall_Sequence pins the D-39/D-40 subprocess order end to end over
// the fake desktop: sources snapshot → state save → XML → env-carrying
// write-cache → list-engine verification → unit → daemon-reload →
// enable --now → ibus restart → live-registration wait → single-owner
// sources set → engine activation.
func TestInstall_Sequence(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{"xkb:us::eng", engineENName})

	runInstall(t, i)

	// The live-true order (first install-cycle run): the cache-file probe
	// verifies write-cache immediately (list-engine stays stale until the
	// restart), and the daemon-view list-engine gate runs AFTER the restart.
	assertCallSequence(t, f.snapshot(), []struct{ name, args string }{
		{binGSettings, "get " + gsettingsSchema + " " + gsettingsKey},
		{binIbus, opWriteCache},
		{binSystemctl, "--user daemon-reload"},
		{binSystemctl, "--user enable --now goswitchd"},
		{binIbus, "restart"},
		{binIbus, "list-engine"},
		{binGSettings, "set " + gsettingsSchema + " " + gsettingsKey + " " + goswitchSources},
		{binIbus, "engine goswitch-en"},
	})

	xmlPath, unitPath, statePath := installPaths(home)
	for _, p := range []string{xmlPath, unitPath, statePath} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("install artifact %s missing: %v", p, err)
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
func TestInstall_SecondInstallKeepsOriginalBackup(t *testing.T) {
	f := &fakeRunner{}
	home := t.TempDir()
	i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})

	runInstall(t, i)

	_, _, statePath := installPaths(home)
	marker := `{"sources":"` + markerSources + `"}`
	if err := os.WriteFile(statePath, []byte(marker), 0o600); err != nil {
		t.Fatalf("plant the marker state: %v", err)
	}

	// The second install sees the post-takeover desktop (goswitch-only
	// sources) — exactly the state it must NOT save over the marker.
	f.stub = func(name string, args []string) ([]byte, error) {
		if name == "gsettings" && len(args) == 3 && args[0] == "get" && args[2] == gsettingsKey {
			return []byte(goswitchSources), nil
		}

		return defaultReply(name, args)
	}
	runInstall(t, i)

	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("re-read state: %v", err)
	}
	if !strings.Contains(string(data), markerSources) {
		t.Errorf("second install overwrote the original backup: state = %q, want %q intact",
			data, markerSources)
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

	assertCallSequence(t, uninstallCalls(t, f), []struct{ name, args string }{
		{binSystemctl, "--user stop goswitchd"},
		{binSystemctl, "--user disable goswitchd"},
		{binSystemctl, "--user daemon-reload"},
		{binIbus, opWriteCache},
		{binIbus, "restart"},
		{binGSettings, "set " + gsettingsSchema + " " + gsettingsKey + " " + ownerSources},
		{binIbus, "engine " + derivedActivation},
	})

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
// FullRollback corpus asserts the uninstall suffix of the recording).
const installCallCount = 8

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

// TestUninstall_CorruptStateFallsBack pins the ASVS V5 guard: a state file
// whose value fails the shape check never reaches gsettings verbatim — the
// restore substitutes the safe xkb fallback, REPORTS it, and the rollback
// still completes (a corrupt backup must not brick the keyboard or abort
// the uninstall).
func TestUninstall_CorruptStateFallsBack(t *testing.T) {
	for name, plant := range map[string]string{
		"garbage value": `{"sources":"garbage"}`,
		"unparsable":    "\x00not json",
	} {
		t.Run(name, func(t *testing.T) {
			f := &fakeRunner{}
			home := t.TempDir()
			i := newInstaller(t, f, home, selfDirWithDaemon(t), []string{engineENName})
			runInstall(t, i)

			_, _, statePath := installPaths(home)
			if err := os.WriteFile(statePath, []byte(plant), 0o600); err != nil {
				t.Fatalf("plant the corrupt state: %v", err)
			}

			report, err := i.Uninstall(context.Background(), false)
			if err != nil {
				t.Fatalf("Uninstall() over a corrupt state err = %v — must not abort (report %v)",
					err, report)
			}

			var restore *instCall
			for _, c := range uninstallCalls(t, f) {
				if c.name == binGSettings && len(c.args) == 4 && c.args[0] == "set" {
					restore = &c

					break
				}
			}
			if restore == nil {
				t.Fatal("no gsettings set recorded — the fallback must still restore a usable source")
			}
			if restore.args[3] != fallbackSources {
				t.Errorf("restore set %q, want the safe fallback %q", restore.args[3], fallbackSources)
			}
			if !slices.ContainsFunc(report, func(line string) bool { return strings.Contains(line, "fallback") }) {
				t.Errorf("report %v does not mention the fallback — the substitution must be visible", report)
			}
		})
	}
}
