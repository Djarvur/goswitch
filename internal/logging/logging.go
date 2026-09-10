// Package logging configures the daemon's structured JSON logger (INST-04).
//
// Level contract:
//   - ERROR: adapter failures, re-registration exhaustion.
//   - WARN: reconnects, capability anomalies.
//   - INFO: lifecycle (connected, component registered, engine created,
//     focus in/out, re-registered). Buffer contents are never logged at
//     INFO — lengths only.
//   - DEBUG: key trace. Enabled exclusively by the -debug flag because
//     keystroke logs capture passwords.
package logging

import (
	"io"
	"log/slog"
)

// Setup installs a slog JSON handler on w as the process default logger.
// debug switches the level from INFO to DEBUG; it defaults to off in the
// caller because the DEBUG key trace records every keystroke.
func Setup(w io.Writer, debug bool) {
	level := new(slog.LevelVar)
	if debug {
		level.Set(slog.LevelDebug)
	} else {
		level.Set(slog.LevelInfo)
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})))
}
