package main

import (
	"context"
	"fmt"
)

// smokeWord is the probe word of the surface smoke cases: injected through
// the physical path, it must read back verbatim. These cases prove the
// DRIVER (spawn, witness, readback, close) — correction wiring arrives in
// plans 02-03..02-05, which reuse the drivers as-is.
const smokeWord = "ghbdtn"

// surfaceDriver bundles the four operations of one surface driver so the
// smoke runner (and the 02-06 matrix after it) consumes drivers opaquely.
type surfaceDriver struct {
	app string
	// open spawns the surface's process; close reaps it unconditionally.
	open  func(context.Context) error
	close func()
	// await is the input-focus gate: wantChars is the expected field
	// length (0 before injection, the probe-word length after it).
	await func(context.Context, int) error
	// read returns the surface input's current text.
	read func(context.Context) (string, error)
}

// runSurfaceSmoke proves one surface driver on a live session: the
// goswitch engine is active (observer mode — phase 2 correction is not
// wired yet), the surface opens its own window, the witness verifies its
// input owns focus BEFORE any injection, and the focused readback returns
// the probe word. The gate runs a second time after injection so a focus
// that drifted mid-case fails before the readback reads someone else's
// node.
func runSurfaceSmoke(ctx context.Context, s *stand, d surfaceDriver) error {
	if err := s.activateGoswitch(ctx); err != nil {
		return err
	}
	if err := d.open(ctx); err != nil {
		return err
	}
	defer d.close()

	if err := d.await(ctx, 0); err != nil {
		return err
	}
	if err := s.injectText(ctx, smokeWord); err != nil {
		return err
	}
	if err := d.await(ctx, len(smokeWord)); err != nil {
		return fmt.Errorf("%s surface lost the injected field: %w", d.app, err)
	}
	got, err := d.read(ctx)
	if err != nil {
		return err
	}
	if got != smokeWord {
		return fmt.Errorf("%s readback %q, want %q", d.app, got, smokeWord)
	}

	return nil
}

// runChromiumSmoke proves the Chromium surface (D-17): fresh instance,
// witness-gated injection, focused readback, PID close.
func runChromiumSmoke(ctx context.Context, s *stand) error {
	return runSurfaceSmoke(ctx, s, surfaceDriver{
		app:   "chromium",
		open:  s.startChromium,
		await: s.waitChromiumInput,
		read:  s.readChromiumText,
		close: s.closeChromium,
	})
}

// runGTESmoke proves the gnome-text-editor surface (the GtkSourceView
// multiline class of the matrix): standalone instance with an isolated
// data dir, witness-gated injection, focused readback, PID close.
func runGTESmoke(ctx context.Context, s *stand) error {
	return runSurfaceSmoke(ctx, s, surfaceDriver{
		app:   gteAppName,
		open:  s.startGTE,
		await: s.waitGTEInput,
		read:  s.readGTEText,
		close: s.closeGTE,
	})
}
