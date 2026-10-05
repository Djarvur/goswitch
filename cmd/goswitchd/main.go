// Command goswitchd is the goswitch IBus engine daemon. Phase 1 ships the
// observer skeleton: register two engines, trace keys behind -debug,
// consume nothing, survive ibus-daemon restarts. Phase 3 adds the YAML
// config: -config loads the startup values (the tap window reaches the
// FSM) and starts the hot-reload watcher on the daemon's signal context.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/Djarvur/goswitch/engine"
	"github.com/Djarvur/goswitch/internal/activate"
	"github.com/Djarvur/goswitch/internal/config"
	"github.com/Djarvur/goswitch/internal/ctlsvc"
	"github.com/Djarvur/goswitch/internal/hotkey"
	"github.com/Djarvur/goswitch/internal/indicator"
	"github.com/Djarvur/goswitch/internal/logging"
	"github.com/Djarvur/goswitch/internal/session"
	"github.com/Djarvur/goswitch/internal/sound"
)

func main() {
	debug := flag.Bool("debug", false, "enable key tracing (logs every keystroke — passwords become visible)")
	configPath := flag.String("config", "", "path to the YAML config file (empty = built-in defaults)")
	showVersion := flag.Bool("version", false, "print the build version and exit")
	flag.Parse()

	// The version query never reaches the daemon: print and exit 0 before
	// run() — no signal handling, no bus, no engine (D-37).
	if *showVersion {
		runVersion(os.Stdout)
		os.Exit(0)
	}

	// No defer here: os.Exit skips defers, so signal handling lives in run().
	if err := run(context.Background(), *debug, *configPath); err != nil {
		fmt.Fprintf(os.Stderr, "goswitchd: %v\n", err)
		os.Exit(1)
	}
}

// run holds the daemon lifecycle: signal context, structured logging,
// config loading/watching and the engine registration loop. All testable
// logic lives below main.
func run(ctx context.Context, debug bool, configPath string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	logging.Setup(os.Stderr, debug)

	lc, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	if lc.fromFile {
		// INFO contract (T-03-02-04): the path and the applied window only —
		// never section content.
		slog.Info("config loaded", "path", lc.path, "tap_window_ms", lc.cfg.Timeouts.TapWindowMs)
	} else {
		slog.Info("config defaults", "path", lc.path)
	}
	var watcher *config.Watcher
	if lc.path != "" {
		w, werr := startConfigWatcher(ctx, lc, configPath)
		if werr != nil {
			return fmt.Errorf("watch config: %w", werr)
		}
		watcher = w
	}

	actor := newActor(lc.cfg, watcher)
	startCtl(ctx, actor, watcher, lc.cfg, lc.path)
	if err := engine.Run(ctx, engineConfig(actor)); err != nil {
		return fmt.Errorf("engine run: %w", err)
	}

	return nil
}

// The org.freedesktop.Notifications surface the status menu item posts
// through — pure D-Bus on the SAME ctl connection, no subprocess.
const (
	notificationName = "org.freedesktop.Notifications"
	notificationPath = dbus.ObjectPath("/org/freedesktop/Notifications")
)

// startCtl runs the control service on the session bus — the daemon's
// SECOND godbus connection (the engine rides the private IBus socket,
// this one the session bus), started and stopped on the daemon's signal
// context. The startup config and the FACTUAL config path (explicit
// -config or the adopted default, 07-03) ride along: the menu's toggles and
// «Настройки…» read and write THAT document. A ctl failure NEVER kills the
// daemon: the desktop's input rides this process, so the error is logged
// and the daemon keeps serving without the control surface (the appid
// degradation precedent).
func startCtl(ctx context.Context, actor *session.Actor, watcher *config.Watcher, cfg config.Config, cfgPath string) {
	var reload ctlsvc.Reloader
	if watcher != nil {
		reload = watcher // nil without a resolvable path: ReloadConfig answers "no config file"
	}
	go func() {
		deps := ctlsvc.Deps{Status: actor, Reload: reload, Correct: actor}
		// The tray indicator rides the ctl connection (quick plan
		// 260930-pf6): one bus name more visible at org.djarvur.goswitch,
		// attached after the name and the ctl object go live. Attach errors
		// are WARNed by Run and serving continues — the daemon types on
		// without its tray icon. The display installs before engine.Run, so
		// the icon is live from the first generation; SetModeDisplay's
		// install push covers any ordering skew.
		//
		// The interactive surface — menu v2 (plan 07-05, the canonical
		// layout 287999c): the EN/RU radio pair flips through the actor's
		// public SwitchMode (the SAME flipTo path as every other gesture —
		// ADR-006, never a second mechanism), both checkmark toggles persist
		// through the 07-03 writer (snapshot → invert → write → menu push →
		// synchronous reload; the 200 ms echo re-applies the same values
		// harmlessly), «Настройки…» ensures the document then xdg-opens it
		// no-pipes, «О программе» posts the mode+version notification
		// (counts/states only — T-03-06-03), «Перечитать конфиг» drives the
		// watcher's Reload and is served greyed only when no config path
		// resolved (the degenerate no-HOME case). The menu's INITIAL state
		// rides the STARTUP config (the truth the actor re-folds per event)
		// and SetMenuSync's install push shows the current mode. The
		// supervisor rides the serve context OnConn receives — the tray
		// dies with the daemon, and the icon self-heals when the shell's
		// watcher appears late or evicts the item.
		deps.OnConn = func(connCtx context.Context, conn *dbus.Conn) error {
			// The menu rides an ATOMIC slot (WR-01): Attach exports the menu
			// object BEFORE `item.Menu()` is stored, so a click's dispatch
			// goroutine can read the slot in flight with the store — a plain
			// pointer write raced that read (UB, invisible to the corpus
			// until TestMenuTogglePushRacesAttachStore). Load() == nil keeps
			// a pre-store click a harmless no-op.
			var menuSlot atomic.Pointer[indicator.Menu]
			editor := &configEditor{ensure: config.EnsureDocument}
			var syncReload func()
			if reload != nil {
				syncReload = func() { reloadConfig(reload) }
			}
			// The toggles' apply is SYNCHRONOUS end to end (the 07-05 pin,
			// live-pinned by the 07-06 menu-v2 case): the plain reload only
			// re-stores the watcher snapshot — the fold puts the applied
			// values in force in the same click, no key event in between.
			// The «Перечитать конфиг» item and the CLI reload keep the
			// 03-04 per-event succession (their pinned semantics).
			toggleReload := syncReload
			if reload != nil {
				toggleReload = func() {
					syncReload()
					actor.FoldAppliedConfig()
				}
			}
			toggles := newMenuToggles(actor, &menuSlot, cfgPath, toggleReload)
			cb := indicator.Callbacks{
				Toggle:            actor.ToggleMode,
				Status:            func() { notifyStatus(conn, actor) },
				Switch:            actor.SwitchMode,
				ToggleAutocorrect: toggles.ac.flip,
				ToggleSound:       toggles.sound.flip,
				Settings:          func() { editor.open(cfgPath) },
				About:             func() { notifyStatus(conn, actor) },
			}
			if syncReload != nil {
				cb.Reload = syncReload
			}
			item := indicator.Attach(conn, ctlsvc.BusName, cb)
			menuSlot.Store(item.Menu())
			menu := menuSlot.Load()
			// The initial menu state: version, both toggles, the raw key
			// names — the menu dedupes later identical pushes from the fold.
			menu.SetVersion(version)
			menu.SetAutocorrectEnabled(cfg.Autocorrect.Enabled)
			menu.SetSoundEnabled(cfg.Sound.EffectiveEnabled())
			menu.SetKeys(cfg.Hotkeys.TapKey, cfg.Hotkeys.WordLayoutCombo, cfg.Hotkeys.ModeSwitchChord)
			actor.SetMenuSync(menu) // the install push covers ordering skew, as SetModeDisplay
			// The sound sink (plan 07-08): the flip tone rides the schema
			// constant, the autocorrect tone starts at the startup
			// document's effective event — the actor's fold re-pushes every
			// changed event to the installed sink (WR-02: the key folds
			// live), so no reload branch exists here either (the D-32
			// contour — the wiring installs once).
			actor.SetSoundSink(sound.New(config.DefaultSoundFlipEvent, cfg.Sound.EffectiveAutocorrectEvent()))
			actor.SetModeDisplay(item)
			go item.Supervise(connCtx, conn)

			return nil
		}
		if err := ctlsvc.Run(ctx, deps); err != nil {
			slog.Error("ctl service stopped", "error", err)
		}
	}()
}

// menuToggles carries the two persisted-toggle compositions of the menu
// (plan 07-05): the autocorrect and the sound click both ride the SAME
// composition — snapshot → invert → write → push → reload — over their own
// writer and their own status truth.
type menuToggles struct {
	ac    *configToggle
	sound *configToggle
}

// newMenuToggles wires the compositions: the applied value reads from the
// actor's status snapshot, the write goes through the 07-03 writer on the
// FACTUAL config path, the push lands on the menu's ATOMIC slot (nil until
// Attach — the Load guard keeps a click before the store a harmless no-op,
// and the store races no dispatch read — WR-01), the synchronous apply
// rides the reload (nil without a config source — the flip's own guard
// handles it).
func newMenuToggles(actor *session.Actor, menu *atomic.Pointer[indicator.Menu], cfgPath string, syncReload func()) menuToggles {
	return menuToggles{
		ac: &configToggle{
			path:  cfgPath,
			read:  func() bool { return actor.StatusSnapshot().AutoCorrectEnabled },
			write: config.SetAutocorrectEnabled,
			push: func(on bool) {
				if m := menu.Load(); m != nil {
					m.SetAutocorrectEnabled(on)
				}
			},
			reload: syncReload,
		},
		sound: &configToggle{
			path:  cfgPath,
			read:  func() bool { return actor.StatusSnapshot().SoundEnabled },
			write: config.SetSoundEnabled,
			push: func(on bool) {
				if m := menu.Load(); m != nil {
					m.SetSoundEnabled(on)
				}
			},
			reload: syncReload,
		},
	}
}

// notifyStatus posts ONE desktop notification with the current mode and the
// build version from the actor's snapshot — counts/states only, never user
// text (T-03-06-03 canon). An error is a WARN, never a panic: the
// notification daemon is a same-session, same-uid surface, and its failure
// costs nothing but the message (the recover shim around menu dispatch
// bounds whatever else escapes).
func notifyStatus(conn *dbus.Conn, actor *session.Actor) {
	st := actor.StatusSnapshot()
	body := "Mode: " + st.Mode + "\nVersion: " + st.Version
	call := conn.Object(notificationName, notificationPath).
		Call("org.freedesktop.Notifications.Notify", 0,
			"goswitch", uint32(0), "", "goswitch", body,
			[]string{}, map[string]dbus.Variant{}, int32(-1))
	if call.Err != nil {
		slog.Warn("status notification failed", "component", "tray indicator", "error", call.Err)
	}
}

// procStarter is the started-process seam of the settings launcher (plan
// 07-05): Start brings the editor up, Wait reaps it. *exec.Cmd is the
// production value; the corpus installs a fake.
type procStarter interface {
	Start() error
	Wait() error
}

// newEditorProc builds the editor launch: xdg-open with the config path as
// its ONLY argument — never a shell, never user data (T-07-05-01) — and NO
// pipes: the fork-shaped grandchild (a GUI editor outliving the daemon)
// would deadlock a piped Run on its write-ends (the wl-copy precedent,
// clipboard.go:141-150). The package var is the corpus seam.
//
//nolint:gochecknoglobals // the launcher's test seam (the D-37 stamping precedent)
var newEditorProc = func(bin, path string) procStarter {
	return exec.Command(bin, path) // nil Stdout/Stderr: the no-pipes fork-shaped form (T-07-05-02)
}

// configEditor opens the config document in the desktop editor
// («Настройки…», plan 07-05): the document is ensured first (an editor on
// an empty buffer would strict-decode-refuse on save, research Q7), then
// xdg-open launches on the pinned path. A failed start is ONE WARN per
// editor lifetime and the menu keeps working; $EDITOR is deliberately not
// consulted — the daemon has no controlling tty (research Q7).
type configEditor struct {
	ensure func(path string) error
	mu     sync.Mutex
	warned bool
}

// open runs one ensure+launch episode. The Start error path keeps the
// one-WARN-per-lifetime discipline (the emitWarned precedent); the launch
// itself is fire-and-forget — the reaper goroutine has no deadline, nothing
// kills a running editor.
func (e *configEditor) open(path string) {
	if err := e.ensure(path); err != nil {
		slog.Warn("config ensure failed", "component", "tray indicator", "error", err)

		return
	}
	cmd := newEditorProc("xdg-open", path)
	if err := cmd.Start(); err != nil {
		e.mu.Lock()
		first := !e.warned
		e.warned = true
		e.mu.Unlock()
		if first {
			slog.Warn("config editor start failed", "component", "tray indicator", "error", err)
		}

		return
	}
	go func() { _ = cmd.Wait() }() // reap without a deadline — nothing kills a running editor
}

// configToggle is one persisted menu toggle (plan 07-05): the click reads
// the APPLIED value from the actor's status snapshot, inverts it, writes it
// through the config writer (the 07-03 Node round-trip), pushes the menu at
// the click and applies the document synchronously through the reload —
// never a write-only toggle (research Q3b anti-pattern).
type configToggle struct {
	path   string
	read   func() bool
	write  func(path string, on bool) error
	push   func(on bool)
	reload func() // nil without a config source — impossible after adopt (07-03), guarded anyway
}

// flip executes one click in the pinned order: snapshot → write → menu
// push → reload. A failed write leaves everything untouched (one WARN —
// the state did not change); a missing reload keeps the write AND the push
// (the file is the truth, the status catches up through the next fold) and
// WARNs the deferred apply. The 200 ms echo reload re-applies the same
// values harmlessly (the idempotent fold, 07-03).
func (t *configToggle) flip() {
	on := !t.read()
	if err := t.write(t.path, on); err != nil {
		slog.Warn("config toggle write failed", "component", "tray indicator", "error", err)

		return
	}
	t.push(on)
	if t.reload == nil {
		slog.Warn("config toggle apply deferred", "component", "tray indicator", "reason", "no config source")

		return
	}
	t.reload()
}

// reloadConfig drives the forced synchronous re-read and logs the outcome —
// the menu item's mirror of goswitchctl reload: WARN on the rejection (the
// last-good snapshot keeps serving, D-32), INFO with the applied message on
// success (the 03-07 log-form canon).
func reloadConfig(r ctlsvc.Reloader) {
	msg, err := r.Reload()
	if err != nil {
		slog.Warn("config reload rejected", "error", err)

		return
	}
	slog.Info("config reloaded", "applied", msg)
}

// loadedConfig is loadConfig's explicit result: the effective config, the
// factual file path (empty only in the no-HOME degenerate case) and
// whether the values came from a file — the adopt+watch wiring consumes
// all three; an explicit result, never global state.
type loadedConfig struct {
	cfg      config.Config
	path     string
	fromFile bool
}

// adoptLoad is the watcher loader for the ADOPTED default path: a missing
// file serves the built-in defaults (the 04-02 green-defaults contract —
// the watcher constructs before the first toggle creates the document),
// a broken one rejects the reload and keeps the last-good snapshot
// serving (D-32).
func adoptLoad(path string) (*config.Config, error) {
	cfg, err := config.Load(path)
	if err == nil {
		return cfg, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		d := config.Defaults()

		return &d, nil
	}

	return nil, fmt.Errorf("load config %s: %w", path, err)
}

// configDirPerm is the adopted config DIRECTORY's mode (mnd) —
// owner-private like the file (the configPerm precedent).
const configDirPerm = 0o700

// startConfigWatcher starts the hot-reload watcher on the FACTUAL config
// path — unconditionally whenever a path resolved (adopt+watch, the owner
// decision: the file the first toggle creates is picked up by the same
// watch, and an explicit -config keeps its hot reload as today). An
// explicit -config keeps the plain Load contract; the adopted path serves
// the defaults until the document exists (adoptLoad), and its parent
// directory is created first — a fresh install has neither file nor
// directory, the watcher watches the DIRECTORY, and refusing the start
// over a missing empty directory would break the 04-02 green-defaults
// contract (the FILE itself is still never generated at start). A watcher
// failure is a start refusal, as with -config today.
func startConfigWatcher(ctx context.Context, lc loadedConfig, explicit string) (*config.Watcher, error) {
	loader := config.Load
	if explicit == "" {
		loader = adoptLoad
		if err := os.MkdirAll(filepath.Dir(lc.path), configDirPerm); err != nil {
			return nil, fmt.Errorf("create config directory: %w", err)
		}
	}

	w, err := config.NewWatcher(ctx, lc.path, config.WithLoader(loader))
	if err != nil {
		return nil, fmt.Errorf("start config watcher: %w", err)
	}

	return w, nil
}

// loadConfig resolves the startup configuration: an explicit -config must
// load and validate (a refusal is a visible start error, never silent
// defaults); without the flag the daemon ADOPTS the default path — an
// existing document loads with the same contract as an explicit -config
// (a broken file refuses the start loudly, T-07-03-03: the divergence
// with the disk is invisible otherwise), an absent one stays the green
// defaults start of plan 04-02 with the default path returned for the
// watcher (nothing is created at start — generation is a user action).
// The degenerate no-HOME environment gets the defaults without a watcher.
func loadConfig(explicit string) (loadedConfig, error) {
	path, adopt := explicit, false
	if path == "" {
		path = config.DefaultPath()
		if path == "" {
			return loadedConfig{cfg: config.Defaults()}, nil
		}
		adopt = true
	}
	if adopt {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			return loadedConfig{cfg: config.Defaults(), path: path}, nil
		}
	}
	cfg, err := config.Load(path)
	if err != nil {
		return loadedConfig{}, fmt.Errorf("load config: %w", err)
	}

	return loadedConfig{cfg: *cfg, path: path, fromFile: true}, nil
}

// newActor builds the session actor with the startup config applied (the
// FSM's tap window from the config, the correction options through
// SetOptions; an attached watcher has PRIORITY per event — the succession
// of plan 03-04: SetOptions remains the no-config surface, the snapshot
// wins once a source exists). The actor is built here so the control
// service can hold it from the start (INST-02); the daemon's build version
// (D-37) is pinned so status identifies the running build.
func newActor(cfg config.Config, watcher *config.Watcher) *session.Actor {
	window := time.Duration(cfg.Timeouts.TapWindowMs) * time.Millisecond
	actor := session.NewActor(window)
	actor.SetVersion(version)
	// The mode-switch chord binding resolves here (owner decision 2, quick
	// plan 260927-way): validated documents never error and Defaults always
	// parses, so a failure is a wiring bug — WARNed and the chord stays
	// disabled (the zero Binding), never a daemon start failure.
	chord, err := hotkey.ParseBinding(cfg.Hotkeys.ModeSwitchChord)
	if err != nil {
		slog.Warn("mode switch chord disabled", "reason", err.Error())
	}
	actor.SetOptions(session.Options{
		BackspaceCap:        cfg.Correction.BackspaceCap,
		ClipboardRung:       cfg.Correction.ClipboardRung,
		FlipAfterCorrection: cfg.Correction.FlipAfterCorrection,
		ModeSwitchChord:     chord,
		// The sound switch of the startup document (plan 07-08): the
		// owner's default ON through the Effective truth — an explicit
		// enabled: false mutes from the first gesture; a attached source
		// re-folds it live per event (the hot-reload contour).
		SoundEnabled: cfg.Sound.EffectiveEnabled(),
	})
	if watcher != nil {
		actor.AttachConfig(watcher)
		// The startup document is in force AT STARTUP (plan 07-06, the
		// restart-survival pin): the fold the first key event would run
		// runs now — the status and the menu report the document truth
		// from the first read, never a stale off (the 03-04 per-event
		// succession is unchanged after it).
		actor.FoldAppliedConfig()
	}

	return actor
}

// engineConfig builds the registration payload: one component, two
// engines — the D-01 experiment needs both from day one. The FSM's tap
// window flows from the actor's startup wiring above (SWCH-04/D-35):
// without -config the built-in default equals hotkey.DefaultWindow
// (pinned by config.TestDefaults).
//
// The sync loop of ADR-006 (05-04, criterion 3) wires both directions:
// OnGlobalEngine feeds every observed engine name (GlobalEngineChanged AND
// FocusIn) into the actor's SyncEngine — the daemon FOLLOWS the factual
// engine, never fights it; BindGlobalEngine keeps the generation-scoped
// GetGlobalEngine reader, and PostRegister hands it to activate.IfOwned so
// a reactivation after (re)registration prefers the FACTUAL engine over
// the dead gsettings current key (Pitfall 3). Serve binds the reader
// BEFORE PostRegister fires, so every generation's reactivation consults
// its own generation's reader.
func engineConfig(actor *session.Actor) engine.Config {
	engines := []engine.EngineDesc{
		engine.NewEngineDesc("goswitch-en", "goswitch English (US)", "en", "us", "en"),
		engine.NewEngineDesc("goswitch-ru", "goswitch Русская", "ru", "ru", "ru"),
	}

	// readGlobalEngine is rebound on every generation before PostRegister;
	// nil only before the first bind ever — a generation's PostRegister
	// cannot run ahead of its own bind (the serve order).
	var readGlobalEngine func(ctx context.Context) (string, bool)

	return engine.Config{
		Component: engine.NewComponent(engines),
		Engines:   engines,
		Handler:   actor, // decides consumption (RU script mode) at the key; tap decisions at window expiry.
		// The sync-listener input: the engine adapter calls it with the
		// observed wire name; the actor's SyncEngine corrects only its
		// internal mode, never the bus (single-writer, T-05-04-01).
		OnGlobalEngine: actor.SyncEngine,
		PostRegister: func(ctx context.Context, _ int) {
			activate.IfOwned(ctx, activate.NewExecRunner(), readGlobalEngine)
		},
		// BindSwitcher hands the generation-scoped SetGlobalEngine closure
		// to the actor (D-52): every flip gesture leaves the daemon through
		// THIS closure — the engine-truth flip of ADR-006 — rebound on every
		// reconnecting generation by the engine's serve cycle. The method
		// value is wiring before engine.Run; the actor's nil-switcher
		// degradation (one WARN, the chord-parse precedent below) keeps the
		// daemon starting even if the seam never arrives.
		BindSwitcher: actor.SetSwitcher,
		// BindGlobalEngine adapts the engine reader's (string, error) to
		// IfOwned's (string, bool): an unreadable or empty name is the cold
		// bus, and IfOwned falls back to the current-index derivation.
		BindGlobalEngine: func(get func(ctx context.Context) (string, error)) {
			readGlobalEngine = func(ctx context.Context) (string, bool) {
				name, err := get(ctx)
				if err != nil || name == "" {
					return "", false
				}

				return name, true
			}
		},
	}
}
