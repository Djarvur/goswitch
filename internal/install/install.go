// Package install implements the goswitch user-space installer (INST-01,
// D-39/D-40/D-42): goswitchctl install / uninstall drive the whole lifecycle
// — component XML into ~/.config/ibus/component (registered through an
// env-carrying `ibus write-cache`), a systemd user unit with an absolute
// ExecStart, the "single owner" input-sources takeover (ADR-001/D-40) with
// the prior sources saved for restore, and the matching full rollback.
// Everything stays under $HOME: no root, no system paths (the no-root
// delivery model is a public promise).
package install

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/Djarvur/goswitch/engine"
)

// Timing knobs: every value is a behavioral constant, named once. cmdTimeout
// bounds ONE subprocess (a wedged systemctl/ibus must fail the named step,
// not hang the terminal); the WHOLE install/uninstall budget is the CLI's
// installTimeout — the deadline here is per call. registrationWait bounds
// the live-registration probe (the daemon observes the cache-written
// component only after an ibus restart — research Pitfall 2),
// registrationPoll is its quantum.
const (
	cmdTimeout       = 10 * time.Second
	registrationWait = 20 * time.Second
	registrationPoll = 500 * time.Millisecond
)

// File permission bits (ASVS V14: explicit, never derived).
const (
	permPrivate = 0o600 // install-state.json — the restore key of the desktop
	permPublic  = 0o644 // component XML and the user unit
	permDir     = 0o755 // created directories (XDG-shape defaults)
)

// The $HOME-relative paths this package owns. The state location is a
// binding contract (Pitfall 8): NEVER under ~/.config/goswitch — D-42
// preserves that directory on default uninstall, so a backup there would
// outlive the uninstall that needs it.
const (
	componentDirRel = ".config/ibus/component"
	componentFile   = "goswitch.xml"
	unitDirRel      = ".config/systemd/user"
	unitFile        = "goswitchd.service"
	stateDirRel     = ".local/share/goswitch"
	stateFile       = "install-state.json"
	userCfgDirRel   = ".config/goswitch"
	cacheDirRel     = ".cache/ibus/bus"
	cacheFile       = "registry"
)

// Pinned binary names and wire identity. The engine names/identity mirror
// engine.NewComponent/NewEngineDesc (cmd/goswitchd/main.go builds the SAME
// payloads at runtime) — the XML must byte-match the wire component or the
// registry and the live component fork (research Pattern 2).
const (
	binIbus      = "ibus"
	binSystemctl = "systemctl"
	binGSettings = "gsettings"

	daemonBinary     = "goswitchd"
	engineEN         = "goswitch-en"
	engineRU         = "goswitch-ru"
	envComponentPath = "IBUS_COMPONENT_PATH"

	gsettingsSchema = "org.gnome.desktop.input-sources"
	gsettingsKey    = "sources"
)

// ownerSourcesSet is the D-40 single-owner takeover value: goswitch becomes
// the ONLY input source. fallbackSources is the ASVS V5 safe restore when
// the saved state cannot be trusted — a plain xkb US keyboard, so an
// uninstall never leaves the desktop without input. systemComponentDir is
// the FHS ibus component dir: the env-carrying write-cache must keep it in
// the scan path (see componentPathEnv).
const (
	ownerSourcesSet    = "[('ibus', 'goswitch-en')]"
	fallbackSources    = "[('xkb', 'us')]"
	systemComponentDir = "/usr/share/ibus/component"
)

// Static errors (err113): every one names the failing step and the fix.
var (
	errHomeUnknown = errors.New("user home directory is not resolvable (is $HOME set?)")
	errSelfUnknown = errors.New("executable directory is not resolvable (run the installed goswitchctl binary)")

	errDaemonMissing = errors.New(
		"goswitchd not found next to goswitchctl (install both binaries into the same directory)")
	errCacheNoGosw = errors.New("goswitch is not in the ibus registry cache after write-cache " +
		"(does the installed ibus honor IBUS_COMPONENT_PATH?)")
	errListNoGosw = errors.New("goswitch is not visible to ibus list-engine after the restart " +
		"(cache-written component did not reach the registry view)")
	errNotRegistered = errors.New(
		"goswitch-en did not register with the live ibus-daemon in time (did the unit start and register?)")
	errRestoreFailed = errors.New("gsettings rejected the restored sources (saved value failed the shape check?)")
	errWriteVerify   = errors.New("written file content mismatch after read-back (possible tampering — ASVS V14)")
)

// Runner executes one installer subprocess (ibus, systemctl, gsettings):
// the seam the unit corpus drives with a recording fake — argv, env and
// stdin are the observable surface — backed in production by
// os/exec.CommandContext. The env parameter carries the full child
// environment on the calls that need IBUS_COMPONENT_PATH pinned (Pitfall 1:
// a plain `ibus write-cache` evicts user components from the registry
// cache), and is nil for inherit-the-parent calls.
type Runner func(ctx context.Context, name string, args []string, env []string, stdin []byte) ([]byte, error)

// ActiveEngines reports the engine names the live ibus-daemon has
// registered (ListActiveEngines on the private IBus socket). A seam so the
// install corpus can answer the bounded-wait registration probe without a
// live bus; the production implementation dials through engine.Discover.
type ActiveEngines func(ctx context.Context) ([]string, error)

// RegistryProbe reports whether the goswitch component is present in the
// ibus registry cache — the IMMEDIATE post-write-cache verification view
// (the live daemon's list-engine cannot see a cache-written component
// until the restart later in the sequence — live finding of the first
// install-cycle run, research A6). Production reads the cache file; the
// corpus replaces the probe to script miss/hit.
type RegistryProbe func() bool

// Installer owns the install/uninstall lifecycle of one goswitch
// deployment. The zero value is not usable — build it with New; every
// collaborator has a seam the corpus replaces.
type Installer struct {
	run           Runner
	home          string
	selfDir       string
	activeEngines ActiveEngines
	registryProbe RegistryProbe
	// ctlStatus is the selfcheck's version-step probe of the daemon's
	// control service (D-41 step 1); the corpus answers without a live bus.
	ctlStatus CtlStatus
}

// installState is the on-disk restore contract between install and
// uninstall: the pre-install gsettings sources string, verbatim.
type installState struct {
	Sources string `json:"sources"`
}

// componentXML is the rendered component document: the wire identity of
// engine.NewComponent in ibus's on-disk XML shape. <exec> stays EMPTY —
// systemd is the sole supervisor (an exec path would double-spawn the
// daemon against the ctlsvc single-instance guard).
type componentXML struct {
	XMLName     xml.Name    `xml:"component"`
	Name        string      `xml:"name"`
	Description string      `xml:"description"`
	Exec        string      `xml:"exec"`
	Version     string      `xml:"version"`
	Author      string      `xml:"author"`
	License     string      `xml:"license"`
	Homepage    string      `xml:"homepage"`
	Textdomain  string      `xml:"textdomain"`
	Engines     []xmlEngine `xml:"engines>engine"`
}

// xmlEngine is one engine entry of componentXML, mirroring EngineDesc.
type xmlEngine struct {
	Name        string `xml:"name"`
	Language    string `xml:"language"`
	Layout      string `xml:"layout"`
	LongName    string `xml:"longname"`
	Description string `xml:"description"`
	Symbol      string `xml:"symbol"`
	Rank        uint32 `xml:"rank"`
}

// New returns the production installer over os/exec and the real $HOME;
// the options replace the collaborators (the test seams). Unset
// collaborators resolve to their production backing here, once.
func New(opts ...func(*Installer)) *Installer {
	i := &Installer{}
	for _, opt := range opts {
		opt(i)
	}
	if i.run == nil {
		i.run = execRunner
	}
	if i.home == "" {
		i.home, _ = os.UserHomeDir() // empty → Install/Uninstall refuse with errHomeUnknown
	}
	if i.selfDir == "" {
		if self, err := os.Executable(); err == nil {
			i.selfDir = filepath.Dir(self)
		} // empty → Install refuses with errSelfUnknown
	}
	if i.activeEngines == nil {
		i.activeEngines = probeActiveEngines
	}
	if i.registryProbe == nil {
		i.registryProbe = i.probeRegistryCache
	}
	if i.ctlStatus == nil {
		i.ctlStatus = probeCtlStatus
	}

	return i
}

// WithRunner replaces the subprocess runner — the test seam.
func WithRunner(r Runner) func(*Installer) {
	return func(i *Installer) {
		i.run = r
	}
}

// WithHome pins the user home the installer lays its files under — the
// corpus points it at t.TempDir instead of the real $HOME.
func WithHome(home string) func(*Installer) {
	return func(i *Installer) {
		i.home = home
	}
}

// WithSelfDir pins the directory whose goswitchd the unit's ExecStart
// points at — the corpus points it at a temp dir with a stand-in binary;
// production resolves the directory of the running executable (ASVS V14:
// an absolute path, never %h or a $PATH lookup).
func WithSelfDir(dir string) func(*Installer) {
	return func(i *Installer) {
		i.selfDir = dir
	}
}

// WithActiveEngines replaces the live-registration probe — the test seam
// for the bounded-wait step at the end of Install.
func WithActiveEngines(fn ActiveEngines) func(*Installer) {
	return func(i *Installer) {
		i.activeEngines = fn
	}
}

// WithRegistryProbe replaces the registry-cache presence probe — the test
// seam for the write-cache verification (miss/hit scripting).
func WithRegistryProbe(fn RegistryProbe) func(*Installer) {
	return func(i *Installer) {
		i.registryProbe = fn
	}
}

// WithCtlStatus replaces the control-status probe of the selfcheck's
// version step (D-41) — the test seam: the corpus answers without a live
// session bus.
func WithCtlStatus(fn CtlStatus) func(*Installer) {
	return func(i *Installer) {
		i.ctlStatus = fn
	}
}

// Install performs the D-39/D-40 sequence and returns the step-by-step
// report (paths and verdicts only — never user text, D-20/D-21):
// preflight the daemon binary → save prior sources → component XML →
// env-carrying write-cache + registry verification → user unit →
// daemon-reload → enable --now → ibus restart → bounded-wait live
// registration → single-owner sources → engine activation.
func (i *Installer) Install(ctx context.Context) ([]string, error) {
	daemonPath, err := i.resolveDaemon()
	if err != nil {
		return nil, err
	}
	if _, err := i.saveState(ctx); err != nil {
		return nil, err
	}
	if err := i.writeComponent(); err != nil {
		return nil, err
	}
	if err := i.registerCache(ctx); err != nil {
		return nil, err
	}
	if err := i.writeUnit(daemonPath); err != nil {
		return nil, err
	}
	if err := i.startUnit(ctx); err != nil {
		return nil, err
	}
	if _, err := i.call(ctx, binIbus, "restart"); err != nil {
		return nil, fmt.Errorf("ibus restart: %w", err)
	}
	if err := i.waitListEngine(ctx); err != nil {
		return nil, err
	}
	if err := i.waitRegistration(ctx); err != nil {
		return nil, err
	}
	if err := i.takeoverSources(ctx); err != nil {
		return nil, err
	}
	if err := i.activateEngine(ctx, engineEN); err != nil {
		return nil, err
	}

	return i.report(daemonPath), nil
}

// Uninstall performs the D-42 rollback and returns the step-by-step report.
// purge additionally removes the user's ~/.config/goswitch and the state
// directory. Every step is idempotent-tolerant: uninstalling an already
// removed installation is a green no-op.
func (i *Installer) Uninstall(ctx context.Context, purge bool) ([]string, error) {
	if i.home == "" {
		return nil, errHomeUnknown
	}

	// stop/disable tolerate the already-gone unit (idempotent uninstall).
	_ = i.unitStep(ctx, "stop")
	_ = i.unitStep(ctx, "disable")

	unitPath := i.path(unitDirRel, unitFile)
	xmlPath := i.path(componentDirRel, componentFile)
	statePath := i.path(stateDirRel, stateFile)
	for _, p := range []string{unitPath, xmlPath} {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("remove %s: %w", p, err)
		}
	}
	if _, err := i.call(ctx, binSystemctl, "--user", "daemon-reload"); err != nil {
		return nil, fmt.Errorf("systemctl daemon-reload: %w", err)
	}
	if err := i.writeCache(ctx); err != nil {
		return nil, err
	}
	if _, err := i.call(ctx, binIbus, "restart"); err != nil {
		return nil, fmt.Errorf("ibus restart: %w", err)
	}

	lines, err := i.restoreSources(ctx)
	if err != nil {
		return nil, err
	}
	if err := os.Remove(statePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("remove %s: %w", statePath, err)
	}
	if purge {
		lines, err = i.purgeDirs(lines)
		if err != nil {
			return nil, err
		}
	}

	return append(lines, "state: "+statePath+" removed"), nil
}

// Install runs the production installer — the thin entry the CLI calls so
// the client stays flag parsing and printing only.
func Install(ctx context.Context) ([]string, error) {
	return New().Install(ctx)
}

// Uninstall runs the production rollback — the CLI's thin entry.
func Uninstall(ctx context.Context, purge bool) ([]string, error) {
	return New().Uninstall(ctx, purge)
}

// resolveDaemon preflights the unit's ExecStart target: goswitchd must
// exist NEXT TO the running goswitchctl (production selfDir). The refusal
// precedes ANY subprocess — a broken installation attempt must not touch
// the desktop.
func (i *Installer) resolveDaemon() (string, error) {
	if i.home == "" {
		return "", errHomeUnknown
	}
	if i.selfDir == "" {
		return "", errSelfUnknown
	}
	path := filepath.Join(i.selfDir, daemonBinary)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("%w: %s", errDaemonMissing, path)
	}

	return path, nil
}

// saveState reads the current sources and stores them verbatim — unless a
// state file already exists, in which case it is LEFT UNTOUCHED: the FIRST
// install's backup is sacred (an install-over-install must never save the
// post-takeover goswitch-only desktop over the owner's original values).
func (i *Installer) saveState(ctx context.Context) (string, error) {
	out, err := i.call(ctx, binGSettings, "get", gsettingsSchema, gsettingsKey)
	if err != nil {
		return "", fmt.Errorf("read current sources: %w", err)
	}
	prior := strings.TrimSpace(string(out))

	path := i.path(stateDirRel, stateFile)
	if _, err := os.Stat(path); err == nil {
		return prior, nil // idempotent backup: the original state stays
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat state file: %w", err)
	}
	data, err := json.Marshal(installState{Sources: prior})
	if err != nil {
		return "", fmt.Errorf("marshal install state: %w", err)
	}
	if err := writeAtomic(path, data, permPrivate); err != nil {
		return "", err
	}

	return prior, nil
}

// writeComponent renders and atomically installs the component XML, then
// verifies the written content by read-back.
func (i *Installer) writeComponent() error {
	data, err := renderComponentXML()
	if err != nil {
		return err
	}

	return writeVerified(i.path(componentDirRel, componentFile), data, permPublic)
}

// registerCache refreshes the registry cache and verifies the component
// entered it by probing the CACHE FILE — the only view that is current
// before the sequence's later ibus restart (the live finding: list-engine
// reads the daemon's startup registry and stays stale until then). EVERY
// write-cache carries IBUS_COMPONENT_PATH (Pitfall 1: a plain write-cache
// evicts user components). The verify-then-retry-once loop covers the
// same-second mtime flake (Pitfall 4, remedy per research A3).
func (i *Installer) registerCache(ctx context.Context) error {
	if err := i.writeCache(ctx); err != nil {
		return err
	}
	if i.registryProbe() {
		return nil
	}
	// Exactly one retry — the observed same-second cache race (Pitfall 4).
	if err := i.writeCache(ctx); err != nil {
		return err
	}
	if i.registryProbe() {
		return nil
	}

	return errCacheNoGosw
}

// writeCache runs one env-carrying registry refresh.
func (i *Installer) writeCache(ctx context.Context) error {
	_, err := i.callEnv(ctx, withEnv(envComponentPath, i.componentPathEnv()), binIbus, "write-cache")
	if err != nil {
		return fmt.Errorf("ibus write-cache: %w", err)
	}

	return nil
}

// componentPathEnv builds the IBUS_COMPONENT_PATH value: the user
// component dir FIRST, then the system dir. The env var REPLACES the scan
// path entirely — a user-only value strips every system component (Config,
// Panel, Simple) from the registry cache, and the next daemon start dies
// without its config component ("Can not execute default config program",
// live finding of the first install-cycle run). The ':' delimiter is the
// documented separator (ibusregistry.h).
func (i *Installer) componentPathEnv() string {
	return i.path(componentDirRel, "") + ":" + systemComponentDir
}

// waitListEngine polls the daemon's registry view until the goswitch
// component is listed. `ibus restart` is ASYNCHRONOUS — the CLI returns
// before the fresh daemon has re-read the registry cache (live finding of
// the second install-cycle run: an immediate one-shot list-engine raced
// the restart and read the dying/not-yet-up daemon) — so the gate waits.
func (i *Installer) waitListEngine(ctx context.Context) error {
	deadline := time.Now().Add(registrationWait)
	for {
		if i.listEngineHasGoswitch(ctx) {
			return nil
		}
		if time.Now().After(deadline) {
			return errListNoGosw
		}
		if err := sleepCtx(ctx, registrationPoll); err != nil {
			return err
		}
	}
}

// listEngineHasGoswitch probes the daemon's registry view (ibus
// list-engine) for the goswitch component's engines — valid ONLY after the
// restart refreshed the daemon's startup registry; an unreadable view
// reads as not-visible (the retry, then the named failure, report it).
func (i *Installer) listEngineHasGoswitch(ctx context.Context) bool {
	out, err := i.call(ctx, binIbus, "list-engine")
	if err != nil {
		return false
	}

	return strings.Contains(string(out), engineEN) || strings.Contains(string(out), engineRU)
}

// probeRegistryCache is the production RegistryProbe: the cache is a
// binary GVariant blob owned by ibus, so the probe greps it (read-only),
// never parses it.
func (i *Installer) probeRegistryCache() bool {
	data, err := os.ReadFile(i.path(cacheDirRel, cacheFile))
	if err != nil {
		return false
	}

	return strings.Contains(string(data), engineEN) || strings.Contains(string(data), engineRU)
}

// writeUnit renders and atomically installs the systemd user unit, then
// verifies the written content by read-back.
func (i *Installer) writeUnit(daemonPath string) error {
	return writeVerified(i.path(unitDirRel, unitFile), renderUnit(daemonPath), permPublic)
}

// startUnit reloads the user manager and enables+starts the unit.
func (i *Installer) startUnit(ctx context.Context) error {
	if _, err := i.call(ctx, binSystemctl, "--user", "daemon-reload"); err != nil {
		return fmt.Errorf("systemctl daemon-reload: %w", err)
	}
	_, err := i.call(ctx, binSystemctl, "--user", "enable", "--now", daemonBinary)
	if err != nil {
		return fmt.Errorf("systemctl enable --now %s: %w (is the user systemd manager running?)", daemonBinary, err)
	}

	return nil
}

// waitRegistration polls ListActiveEngines until the daemon-side
// registration is observable, bounded by registrationWait (the cache
// becomes daemon-visible only after the ibus restart the sequence ran).
func (i *Installer) waitRegistration(ctx context.Context) error {
	deadline := time.Now().Add(registrationWait)
	for {
		engines, err := i.activeEngines(ctx)
		if err == nil && slices.Contains(engines, engineEN) {
			return nil
		}
		if time.Now().After(deadline) {
			if err != nil {
				return fmt.Errorf("%w: last probe: %w", errNotRegistered, err)
			}

			return errNotRegistered
		}
		if err := sleepCtx(ctx, registrationPoll); err != nil {
			return err
		}
	}
}

// takeoverSources writes the D-40 single-owner value.
func (i *Installer) takeoverSources(ctx context.Context) error {
	_, err := i.call(ctx, binGSettings, "set", gsettingsSchema, gsettingsKey, ownerSourcesSet)
	if err != nil {
		return fmt.Errorf("set input sources to %s: %w", ownerSourcesSet, err)
	}

	return nil
}

// activateEngine sets the global engine (SetGlobalEngine via `ibus engine`):
// the activation path the GNOME shell actually honors (the runtime
// gsettings `current` write is ignored — Phase 1 live finding).
func (i *Installer) activateEngine(ctx context.Context, name string) error {
	if _, err := i.call(ctx, binIbus, "engine", name); err != nil {
		return fmt.Errorf("activate engine %s: %w", name, err)
	}

	return nil
}

// report assembles the install step-by-step verdicts (paths and verdicts
// only).
func (i *Installer) report(daemonPath string) []string {
	return []string{
		"daemon: " + daemonPath,
		"state: " + i.path(stateDirRel, stateFile),
		"component: " + i.path(componentDirRel, componentFile) + " content verified (read-back)",
		"registry: goswitch visible after write-cache",
		"unit: " + i.path(unitDirRel, unitFile) + " content verified (read-back)",
		"unit: daemon-reload + enable --now done",
		"engine: registered live (ListActiveEngines)",
		"sources: single owner " + ownerSourcesSet,
		"engine: activated " + engineEN,
	}
}

// unitStep runs one stop/disable; the already-gone unit is not an error
// (systemd names it "not loaded"/"not found").
func (i *Installer) unitStep(ctx context.Context, op string) error {
	_, err := i.call(ctx, binSystemctl, "--user", op, daemonBinary)
	if err != nil && unitGone(err) {
		return nil
	}

	return err
}

// restoreSources puts the saved sources back. The saved value is
// shape-validated BEFORE it reaches gsettings (ASVS V5/T-04-01-02: a
// corrupt or injected state file must never brick the keyboard) — anything
// not shaped like a GVariant array restores the safe xkb fallback, and the
// substitution is REPORTED, never silent.
func (i *Installer) restoreSources(ctx context.Context) ([]string, error) {
	value, trusted := savedSources(i.path(stateDirRel, stateFile))
	if !trusted {
		value = fallbackSources
	}
	if _, err := i.call(ctx, binGSettings, "set", gsettingsSchema, gsettingsKey, value); err != nil {
		return nil, fmt.Errorf("%w: set %s: %w", errRestoreFailed, value, err)
	}

	lines := []string{"sources: restored " + value}
	if !trusted {
		lines = []string{
			"sources: fallback " + fallbackSources +
				" (saved state unreadable or malformed — safe default applied)",
		}
	}

	// Best-effort activation of the restored source (the fallbackEngine
	// derivation): a failure here degrades the report, never the rollback.
	return i.appendActivation(ctx, lines, value), nil
}

// appendActivation derives and activates the restored source's xkb engine,
// degrading to a report line when the activation fails.
func (i *Installer) appendActivation(ctx context.Context, lines []string, value string) []string {
	name := derivedEngine(value)
	if name == "" {
		return lines
	}
	if err := i.activateEngine(ctx, name); err != nil {
		return append(lines, "engine: activation of "+name+" skipped ("+err.Error()+")")
	}

	return append(lines, "engine: activated "+name)
}

// purgeDirs removes the user-owned config dir and the state dir (--purge,
// D-42: the default uninstall preserves ~/.config/goswitch).
func (i *Installer) purgeDirs(lines []string) ([]string, error) {
	for _, dir := range []string{i.path(userCfgDirRel, ""), i.path(stateDirRel, "")} {
		if err := os.RemoveAll(dir); err != nil {
			return nil, fmt.Errorf("purge %s: %w", dir, err)
		}
		lines = append(lines, "purge: "+dir+" removed")
	}

	return lines, nil
}

// path joins a $HOME-relative contract path; an empty name yields the dir.
func (i *Installer) path(dirRel, name string) string {
	return filepath.Join(i.home, dirRel, name)
}

// call bounds one subprocess with the package deadline and hands it to the
// runner — the context the runner sees always carries a cancellation
// deadline (the whole install budget lives at the CLI layer).
func (i *Installer) call(ctx context.Context, name string, args ...string) ([]byte, error) {
	return i.callEnv(ctx, nil, name, args...)
}

// callEnv is call with an explicit child environment (nil = inherit).
func (i *Installer) callEnv(ctx context.Context, env []string, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	return i.run(ctx, name, args, env, nil)
}

// renderComponentXML serializes the component document. The payload is
// fixed repo identity (never user content), so a plain marshal is safe.
func renderComponentXML() ([]byte, error) {
	doc := componentXML{
		Name:        "org.freedesktop.IBus.goswitch",
		Description: "goswitch layout engine",
		Exec:        "", // systemd is the sole supervisor — never spawn from the registry
		Version:     "0.1.0",
		Author:      "Djarvur",
		License:     "MIT",
		Homepage:    "https://github.com/Djarvur/goswitch",
		Engines:     []xmlEngine{},
	}
	engines := wireEngines()
	for idx := range engines {
		e := engines[idx]
		doc.Engines = append(doc.Engines, xmlEngine{
			Name:        e.EngineName,
			Language:    e.Language,
			Layout:      e.Layout,
			LongName:    e.LongName,
			Description: e.Description,
			Symbol:      e.Symbol,
			Rank:        e.Rank,
		})
	}
	data, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("render component XML: %w", err)
	}

	return append([]byte(xml.Header), data...), nil
}

// wireEngines builds the engine identities exactly as cmd/goswitchd does
// (the single wire source both sides must mirror).
func wireEngines() []engine.EngineDesc {
	return []engine.EngineDesc{
		engine.NewEngineDesc("goswitch-en", "goswitch English (US)", "en", "us", "en"),
		engine.NewEngineDesc("goswitch-ru", "goswitch Русская", "ru", "ru", "ru"),
	}
}

// renderUnit renders the systemd user unit: the dist template's ordering
// semantics with the RESOLVED absolute ExecStart (ASVS V14 — no %h, no
// $PATH lookup, no shell interpolation).
func renderUnit(daemonPath string) []byte {
	return []byte(fmt.Sprintf(`# goswitchd user unit — rendered by goswitchctl install (INST-01, D-39).
# Ordering: start only inside a graphical session, after the GNOME IBus user
# service — goswitchd registers with ibus-daemon and retries until it answers.
[Unit]
Description=goswitch layout engine
PartOf=graphical-session.target
After=org.freedesktop.IBus.session.GNOME.service

[Service]
ExecStart=%s
# Covers uncatchable deaths too (kill -9): systemd respawns the daemon, the
# desktop keeps its input because ibus falls back to the previous engine.
Restart=on-failure

[Install]
WantedBy=graphical-session.target
`, daemonPath))
}

// savedSources reads the state file and returns its sources value only
// when the file parses AND the value passes the shape check.
func savedSources(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var st installState
	if err := json.Unmarshal(data, &st); err != nil {
		return "", false
	}
	v := strings.TrimSpace(st.Sources)
	if !strings.HasPrefix(v, "[") || !strings.HasSuffix(v, "]") || len(v) <= len("[]") {
		return "", false
	}

	return v, true
}

// derivedEngine derives the xkb engine name of the first ('xkb', layout)
// tuple — the best-effort activation target of a restored desktop
// (the stand's fallbackEngine shape).
func derivedEngine(sources string) string {
	marker := "'xkb'"
	at := strings.Index(sources, marker)
	if at < 0 {
		return ""
	}
	rest := sources[at+len(marker):]
	start := strings.Index(rest, "'")
	if start < 0 {
		return ""
	}
	end := strings.Index(rest[start+1:], "'")
	if end < 0 {
		return ""
	}

	return "xkb:" + rest[start+1:start+1+end] + "::eng"
}

// unitGone reports the tolerated not-installed verdicts of systemctl.
func unitGone(err error) bool {
	msg := err.Error()

	return strings.Contains(msg, "not loaded") ||
		strings.Contains(msg, "not found") ||
		strings.Contains(msg, "does not exist")
}

// sleepCtx sleeps in abortable chunks so Ctrl-C and the CLI budget
// interrupt waits.
func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("wait aborted: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}

// withEnv builds the child environment with key pinned to value, removing
// any inherited assignment of the same key so a stale value cannot
// resurrect (the surface.go withEnv discipline).
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

// writeVerified writes atomically and then re-reads the file, comparing
// byte-exactly — the ASVS V14 hijack guard (T-04-01-01): a foreign
// pre-created unit or any post-write divergence fails the install with a
// named error instead of surviving silently.
func writeVerified(path string, data []byte, perm os.FileMode) error {
	if err := writeAtomic(path, data, perm); err != nil {
		return err
	}

	return verifyWritten(path, data)
}

// verifyWritten re-reads path and requires the exact written bytes.
func verifyWritten(path string, want []byte) error {
	got, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read back %s: %w", path, err)
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("%w: %s", errWriteVerify, path)
	}

	return nil
}

// writeAtomic writes data to path through a temp file in the SAME
// directory plus a rename — a reader (or the systemd manager) never
// observes a half-written XML/unit/state file, and the temp name cannot
// survive as a leftover.
func writeAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, permDir); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }() // no-op after a successful rename

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()

		return fmt.Errorf("write %s: %w", name, err)
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()

		return fmt.Errorf("chmod %s: %w", name, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", name, err)
	}
	if err := os.Rename(name, path); err != nil {
		return fmt.Errorf("rename %s to %s: %w", name, path, err)
	}

	return nil
}

// execRunner runs one subprocess through os/exec: stdout/stderr captured,
// env applied when given, and the caller's deadline-bounded context kills
// a wedged process (CommandContext). The error carries argv + stderr so
// the CLI verdict names the failing step.
func execRunner(ctx context.Context, name string, args, env []string, stdin []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if env != nil {
		cmd.Env = env
	}
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}

	return out.Bytes(), nil
}

// probeActiveEngines is the production ActiveEngines: ListActiveEngines on
// the private IBus socket (the preflight probe's dial shape), collecting
// every string in the reply's variant tree — matching is by name presence.
// The engine import is deliberate: Discover owns the address file
// semantics, and install is the same module's delivery surface.
func probeActiveEngines(ctx context.Context) ([]string, error) {
	addr, err := engine.Discover()
	if err != nil {
		return nil, fmt.Errorf("discover ibus address: %w", err)
	}
	conn, err := dbus.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("dial ibus bus: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if err := conn.Auth([]dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}); err != nil {
		return nil, fmt.Errorf("dbus auth: %w", err)
	}
	if err := conn.Hello(); err != nil {
		return nil, fmt.Errorf("dbus hello: %w", err)
	}
	call := conn.Object("org.freedesktop.IBus", "/org/freedesktop/IBus").
		CallWithContext(ctx, "org.freedesktop.IBus.ListActiveEngines", 0)
	if call.Err != nil {
		return nil, fmt.Errorf("ListActiveEngines: %w", call.Err)
	}

	return collectStrings(call.Body), nil
}

// collectStrings walks a D-Bus reply body collecting every string: engine
// descs are nested variants, and name-presence matching keeps the probe
// independent of the exact wire shape.
func collectStrings(body []any) []string {
	out := make([]string, 0, len(body))
	var walk func(v any)
	walk = func(v any) {
		switch t := v.(type) {
		case string:
			out = append(out, t)
		case dbus.Variant:
			walk(t.Value())
		case []any:
			for _, item := range t {
				walk(item)
			}
		case []dbus.Variant:
			for _, item := range t {
				walk(item)
			}
		}
	}
	for _, v := range body {
		walk(v)
	}

	return out
}
