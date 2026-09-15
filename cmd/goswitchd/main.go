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

	"github.com/Djarvur/goswitch/engine"
	"github.com/Djarvur/goswitch/internal/config"
	"github.com/Djarvur/goswitch/internal/logging"
	"github.com/Djarvur/goswitch/internal/session"
)

func main() {
	debug := flag.Bool("debug", false, "enable key tracing (logs every keystroke — passwords become visible)")
	configPath := flag.String("config", "", "path to the YAML config file (empty = built-in defaults)")
	flag.Parse()

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
	if err := watchConfig(ctx, configPath); err != nil {
		return err
	}

	if err := engine.Run(ctx, engineConfig(cfg)); err != nil {
		return fmt.Errorf("engine run: %w", err)
	}

	return nil
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

// watchConfig starts the hot-reload watcher on the daemon's signal context
// (D-32): valid re-edits publish new snapshots, rejected ones keep the
// last-good serving. The snapshots await their actor consumer — plan 03-04
// wires them into the tap window, the Backspace cap, the clipboard rung
// flag and the combo binding; the watcher's own lifecycle ends with ctx.
func watchConfig(ctx context.Context, path string) error {
	if path == "" {
		return nil
	}
	if _, err := config.NewWatcher(ctx, path); err != nil {
		return fmt.Errorf("watch config: %w", err)
	}

	return nil
}

// engineConfig builds the registration payload: one component, two
// engines — the D-01 experiment needs both from day one. The FSM's tap
// window flows from the config (SWCH-04/D-35): without -config the
// built-in default equals hotkey.DefaultWindow (pinned by
// config.TestDefaults). The correction options — the D-27 Backspace cap
// and the D-28 clipboard rung switch — are fed at startup the same way
// (the snapshot consumption on hot reload is plan 03-04).
func engineConfig(cfg config.Config) engine.Config {
	engines := []engine.EngineDesc{
		engine.NewEngineDesc("goswitch-en", "goswitch English (US)", "en", "us", "en"),
		engine.NewEngineDesc("goswitch-ru", "goswitch Русская", "ru", "ru", "ru"),
	}
	window := time.Duration(cfg.Timeouts.TapWindowMs) * time.Millisecond
	actor := session.NewActor(window)
	actor.SetOptions(session.Options{
		BackspaceCap:  cfg.Correction.BackspaceCap,
		ClipboardRung: cfg.Correction.ClipboardRung,
	})

	return engine.Config{
		Component: engine.NewComponent(engines),
		Engines:   engines,
		Handler:   actor, // decides consumption (RU script mode) at the key; tap decisions at window expiry.
	}
}
