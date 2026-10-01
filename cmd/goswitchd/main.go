// Command goswitchd is the goswitch IBus engine daemon. Phase 1 ships the
// observer skeleton: register two engines, trace keys behind -debug,
// consume nothing, survive ibus-daemon restarts. Phase 3 adds the YAML
// config: -config loads the startup values (the tap window reaches the
// FSM) and starts the hot-reload watcher on the daemon's signal context.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
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

	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	if configPath != "" {
		// INFO contract (T-03-02-04): the path and the applied window only —
		// never section content.
		slog.Info("config loaded", "path", configPath, "tap_window_ms", cfg.Timeouts.TapWindowMs)
	}
	var watcher *config.Watcher
	if configPath != "" {
		w, werr := config.NewWatcher(ctx, configPath)
		if werr != nil {
			return fmt.Errorf("watch config: %w", werr)
		}
		watcher = w
	}

	actor := newActor(cfg, watcher)
	startCtl(ctx, actor, watcher)
	if err := engine.Run(ctx, engineConfig(actor)); err != nil {
		return fmt.Errorf("engine run: %w", err)
	}

	return nil
}

// startCtl runs the control service on the session bus — the daemon's
// SECOND godbus connection (the engine rides the private IBus socket,
// this one the session bus), started and stopped on the daemon's signal
// context. A ctl failure NEVER kills the daemon: the desktop's input rides
// this process, so the error is logged and the daemon keeps serving
// without the control surface (the appid degradation precedent).
func startCtl(ctx context.Context, actor *session.Actor, watcher *config.Watcher) {
	var reload ctlsvc.Reloader
	if watcher != nil {
		reload = watcher // nil without -config: ReloadConfig answers "no config file"
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
		deps.OnConn = func(conn *dbus.Conn) error {
			// Attach always succeeds: every degradation (no watcher on the
			// bus, a failed register or export) is a one-WARN inert display
			// inside the item. The zero-value Callbacks keeps this task's
			// call site behaviorally inert — the real wiring lands with the
			// daemon task (quick plan 261001-fg3).
			actor.SetModeDisplay(indicator.Attach(conn, ctlsvc.BusName, indicator.Callbacks{}))

			return nil
		}
		if err := ctlsvc.Run(ctx, deps); err != nil {
			slog.Error("ctl service stopped", "error", err)
		}
	}()
}

// loadConfig resolves the startup configuration: an explicit -config must
// load and validate (a refusal is a visible start error, never silent
// defaults); without the flag the documented built-in defaults apply — the
// tap window 300 ms of ADR-002 — so the Phase 2 behavior is preserved
// unchanged (SWCH-04/D-35).
func loadConfig(path string) (config.Config, error) {
	if path == "" {
		return config.Defaults(), nil
	}
	cfg, err := config.Load(path)
	if err != nil {
		return config.Config{}, fmt.Errorf("load config: %w", err)
	}

	return *cfg, nil
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
	})
	if watcher != nil {
		actor.AttachConfig(watcher)
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
