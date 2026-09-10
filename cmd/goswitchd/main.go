// Command goswitchd is the goswitch IBus engine daemon. Phase 1 ships the
// observer skeleton: register two engines, trace keys behind -debug,
// consume nothing, survive ibus-daemon restarts.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Djarvur/goswitch/engine"
	"github.com/Djarvur/goswitch/internal/logging"
)

func main() {
	debug := flag.Bool("debug", false, "enable key tracing (logs every keystroke — passwords become visible)")
	flag.Parse()

	// No defer here: os.Exit skips defers, so signal handling lives in run().
	if err := run(context.Background(), *debug); err != nil {
		fmt.Fprintf(os.Stderr, "goswitchd: %v\n", err)
		os.Exit(1)
	}
}

// run holds the daemon lifecycle: signal context, structured logging and
// the engine registration loop. All testable logic lives below main.
func run(ctx context.Context, debug bool) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	logging.Setup(os.Stderr, debug)

	if err := engine.Run(ctx, engineConfig()); err != nil {
		return fmt.Errorf("engine run: %w", err)
	}

	return nil
}

// engineConfig builds the Phase 1 registration payload: one component,
// two engines — the D-01 experiment needs both from day one.
func engineConfig() engine.Config {
	engines := []engine.EngineDesc{
		engine.NewEngineDesc("goswitch-en", "goswitch English (US)", "en", "us", "en"),
		engine.NewEngineDesc("goswitch-ru", "goswitch Русская", "ru", "ru", "ru"),
	}

	return engine.Config{
		Component: engine.NewComponent(engines),
		Engines:   engines,
		Handler:   nil, // Phase 1: pure observer, no key consumption.
	}
}
