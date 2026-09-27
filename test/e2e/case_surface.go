package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

// runGeditSmoke proves the gedit surface (04-05, D-47): the GTK3-generation
// editor, standalone instance with both XDG dirs isolated, witness-gated
// injection, pid-keyed readback, PID close — then prints the surface's
// spike verdicts (caps, anchor, actual correction level) the matrix-v3
// composition consumes.
func runGeditSmoke(ctx context.Context, s *stand) error {
	if err := runSurfaceSmoke(ctx, s, surfaceDriver{
		app:   geditAppName,
		open:  s.startGedit,
		await: s.waitGeditInput,
		read:  s.readGeditText,
		close: s.closeGedit,
	}); err != nil {
		return err
	}

	return s.spikeVerdictsV3(ctx, selectSurfaceDriver{
		surface:   matrixSurfaceGedit,
		start:     s.startGedit,
		close:     s.closeGedit,
		waitChars: s.waitGeditInput,
		readback:  s.readGeditText,
	})
}

// runChromiumX11Smoke proves the chromium-x11 surface (04-05, D-47): the
// same binary in the --ozone-platform=x11 window mode, fresh profile,
// pid-keyed witness/readback, PID close — then the same spike verdicts,
// pinned SEPARATELY from the Wayland surface (the IM path differs).
func runChromiumX11Smoke(ctx context.Context, s *stand) error {
	if err := runSurfaceSmoke(ctx, s, surfaceDriver{
		app:   chromiumX11AppName + " (x11)",
		open:  s.startChromiumX11,
		await: s.waitChromiumX11Input,
		read:  s.readChromiumX11Text,
		close: s.closeChromiumX11,
	}); err != nil {
		return err
	}

	return s.spikeVerdictsV3(ctx, selectSurfaceDriver{
		surface:   surfaceChromiumX11Name,
		start:     s.startChromiumX11,
		close:     s.closeChromiumX11,
		waitChars: s.waitChromiumX11Input,
		readback:  s.readChromiumX11Text,
	})
}

// spikeVerdictsV3 prints the 04-05 Task-1 verdict records for one surface:
// the caps bitmap its client pushed (the ladder-level input), the ACTUAL
// correction level and field outcome under a clean double tap
// (spikeCorrectionProbe), and the anchor verdict of the 03-03 probe
// (whether a selection push with anchor != cursor ever arrives, and under
// which select-all name). The case passes with whatever the desktop
// produces — the verdicts feed the matrix-v3 composition, they assert
// nothing (the select-smoke discipline).
func (s *stand) spikeVerdictsV3(ctx context.Context, d selectSurfaceDriver) error {
	fmt.Printf("v3-spike %-13s caps=%s\n", d.surface, s.lastCapsValue())
	if err := s.spikeCorrectionProbe(ctx, d); err != nil {
		return err
	}

	return s.runSelectSpike(ctx, d)
}

// spikeCorrectionProbe types the probe word into a fresh instance of the
// surface and fires the double tap — the CLEAN correction record the
// level verdict needs (the select probe's failed candidates destroy the
// field, so its own readback says nothing about the ladder). The readback
// after the tap is printed verbatim: whatever the desktop produced.
func (s *stand) spikeCorrectionProbe(ctx context.Context, d selectSurfaceDriver) error {
	if err := d.start(ctx); err != nil {
		return err
	}
	defer d.close()
	if err := d.waitChars(ctx, 0); err != nil {
		return fmt.Errorf("v3-spike %s correction surface: %w", d.surface, err)
	}
	if err := s.injectText(ctx, wordProbeEN); err != nil {
		return err
	}
	if err := d.waitChars(ctx, len([]rune(wordProbeEN))); err != nil {
		return fmt.Errorf("v3-spike %s correction field: %w", d.surface, err)
	}
	if err := s.injectKeys(ctx, "Shift_R", "Shift_R"); err != nil {
		return err
	}
	if err := s.waitForLog(ctx, `"msg":"action","n":2`, decisionWait); err != nil {
		return fmt.Errorf("v3-spike %s double-tap decision: %w", d.surface, err)
	}
	_ = s.waitForLog(ctx, `"msg":"correction","outcome":"done"`, correctionWait)
	if err := sleepCtx(ctx, selectSettleWait); err != nil {
		return err
	}
	got, err := d.readback(ctx)
	if err != nil {
		return fmt.Errorf("v3-spike %s correction readback: %w", d.surface, err)
	}
	fmt.Printf("v3-spike %-13s level=%s readback=%q\n", d.surface, s.lastCorrectionLevel(), got)

	return nil
}

// lastCapsValue parses the caps bitmap of the NEWEST capabilities record
// in the daemon log — the surface's own ladder-level input (ADR-003).
func (s *stand) lastCapsValue() string {
	lines := s.logLines()
	for i := len(lines) - 1; i >= 0; i-- {
		if !strings.Contains(lines[i], `"msg":"capabilities"`) {
			continue
		}
		var rec struct {
			Caps string `json:"caps"`
		}
		if err := json.Unmarshal([]byte(lines[i]), &rec); err == nil && rec.Caps != "" {
			return rec.Caps
		}
	}

	return "(none)"
}

// lastCorrectionLevel extracts the ladder level of the NEWEST DEBUG
// correction record in the daemon log — the actual level the double tap
// ran at. Grep-shaped, like verifyMatrixCase's oracle: the level is the
// record's FIRST attribute (the actor pins the form), and JSON decoding
// is NOT an option here — slog's own "level" key and the record's collide
// in one object, and the decoder refuses the duplicate string/int pair.
func (s *stand) lastCorrectionLevel() string {
	lines := s.logLines()
	const mark = `"msg":"correction","level":`
	for i := len(lines) - 1; i >= 0; i-- {
		idx := strings.Index(lines[i], mark)
		if idx < 0 {
			continue
		}
		rest := lines[i][idx+len(mark):]
		if end := strings.IndexByte(rest, ','); end >= 0 {
			rest = rest[:end]
		}

		return strings.TrimSpace(rest)
	}

	return "(none)"
}
