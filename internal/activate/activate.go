// Package activate re-activates the goswitch engine as the global IBus
// input source after the daemon's own (re)registration — but only when
// goswitch already owns the current GNOME input source. This is the
// daemon-restart recovery contract (SY8): a restart of goswitchd
// (systemd unit Restart=on-failure) or of ibus-daemon must restore the
// exact input state the user had, unattended — registration alone leaves
// keystrokes routing through plain XKB until something activates the
// engine.
//
// The single decision is ownership: the current index into
// org.gnome.desktop.input-sources sources must point at a `goswitch-`
// engine. A foreign current source (xkb or another ibus engine) is never
// hijacked; malformed or unavailable desktop state degrades to a log line.
// Nothing here is ever fatal: IfOwned returns nothing and never panics —
// every failure is a DEBUG/WARN entry in the journal.
package activate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// The pinned binary names and desktop keys (mirroring install.go: fixed
// names from PATH, the arguments carry no user content — T-SY8-01).
const (
	binGSettings    = "gsettings"
	binIbus         = "ibus"
	gsettingsSchema = "org.gnome.desktop.input-sources"
	keySources      = "sources"
	keyCurrent      = "current"

	// enginePrefix marks the goswitch-owned engines in the sources list.
	enginePrefix = "goswitch-"
)

// cmdTimeout bounds every subprocess call (mirrors install.cmdTimeout;
// defined locally — the daemon must not import internal/install, whose
// Runner seam serves the install corpus — T-SY8-02).
const cmdTimeout = 10 * time.Second

// The default retry knobs: up to three `ibus engine` attempts half a
// second apart. Package vars so the corpus can shrink them (t.Cleanup
// restore); the package's tests never run in parallel.
const (
	defaultActivationAttempts = 3
	defaultActivationDelay    = 500 * time.Millisecond
)

var (
	//nolint:gochecknoglobals // test seam: the retry corpus overrides the knobs (t.Cleanup restore)
	activationAttempts = defaultActivationAttempts
	//nolint:gochecknoglobals // test seam: same override contract as activationAttempts
	activationRetryDelay = defaultActivationDelay

	// tupleRe matches one GVariant 2-tuple of single-quoted strings —
	// the shape gsettings prints for input sources: ('xkb', 'us').
	tupleRe = regexp.MustCompile(`\('([^']*)',\s*'([^']*)'\)`)
)

// Runner executes one subprocess (gsettings, ibus): the seam the unit
// corpus drives with a recording fake — name and args are the observable
// surface — backed in production by os/exec.CommandContext.
type Runner func(ctx context.Context, name string, args []string) ([]byte, error)

// NewExecRunner returns the production Runner: one subprocess through
// os/exec.CommandContext under the per-call timeout. Combined stderr folds
// into the error (a bare exit code names nothing); stdout returns whole.
func NewExecRunner() Runner {
	return func(ctx context.Context, name string, args []string) ([]byte, error) {
		ctx, cancel := context.WithTimeout(ctx, cmdTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, name, args...)
		var out, errOut bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errOut
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("%s %s: %w: %s",
				name, strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
		}

		return out.Bytes(), nil
	}
}

// SourceTuple is one parsed element of the GNOME input-sources list: the
// GVariant 2-tuple ('kind', 'id') — e.g. ('xkb', 'us') or
// ('ibus', 'goswitch-en'). Kind is the source type (xkb, ibus), ID the
// layout or engine name.
type SourceTuple struct {
	Kind string
	ID   string
}

// errMalformedSources is the parse-failure sentinel of the canonical
// sources parser (err113): every refusal names what deviated.
var errMalformedSources = errors.New("malformed input sources")

// ParseSourceTuples reads one gsettings sources output strictly as a
// GVariant text array of 2-tuples of single-quoted strings — the canonical
// parser every consumer (the daemon's ownership verdict, the installer's
// sources wrap) shares, so the wire shape is parsed in exactly one place.
// Any deviation — unbalanced brackets, no tuples, an empty list, garbage
// between tuples — is an error, never a partial parse.
func ParseSourceTuples(raw string) ([]SourceTuple, error) {
	s := strings.TrimSpace(raw)
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return nil, fmt.Errorf("%w: not a bracketed list: %q", errMalformedSources, s)
	}
	body := s[1 : len(s)-1]
	matches := tupleRe.FindAllStringSubmatch(body, -1)
	if matches == nil {
		return nil, fmt.Errorf("%w: no source tuples in %q", errMalformedSources, s)
	}
	if rest := tupleRe.ReplaceAllString(body, ""); strings.IndexFunc(rest, func(r rune) bool {
		return r != ',' && !unicode.IsSpace(r)
	}) >= 0 {
		return nil, fmt.Errorf("%w: garbage between tuples in %q", errMalformedSources, s)
	}
	tuples := make([]SourceTuple, 0, len(matches))
	for _, m := range matches {
		tuples = append(tuples, SourceTuple{Kind: m[1], ID: m[2]})
	}

	return tuples, nil
}

// IfOwned re-activates the engine that owns the current GNOME input
// source: `ibus engine <name>` for the goswitch engine at the current
// index of the sources list. Never returns an error and never panics —
// every outcome lands in the log (INFO success, DEBUG skip/failure
// attempts, WARN exhausted retries). A foreign or malformed current source
// receives no `ibus engine` call at all (T-SY8-01).
func IfOwned(ctx context.Context, run Runner) {
	if err := ctx.Err(); err != nil {
		slog.Debug("engine reactivation skipped", "reason", "context done", "error", err)

		return
	}

	sources, err := readKey(ctx, run, keySources)
	if err != nil {
		slog.Debug("engine reactivation skipped", "reason", "gsettings unavailable", "error", err)

		return
	}
	current, err := readKey(ctx, run, keyCurrent)
	if err != nil {
		slog.Debug("engine reactivation skipped", "reason", "gsettings unavailable", "error", err)

		return
	}

	names := parseSources(sources)
	index, err := parseCurrent(current)
	if err != nil {
		slog.Debug("engine reactivation skipped", "reason", "malformed current", "error", err)

		return
	}
	if int64(index) >= int64(len(names)) {
		slog.Debug("engine reactivation skipped", "reason", "current index out of range",
			"index", index, "sources", len(names))

		return
	}
	if !strings.HasPrefix(names[index], enginePrefix) {
		slog.Debug("engine reactivation skipped", "reason", "current input source is foreign",
			"engine", names[index])

		return
	}

	reactivate(ctx, run, names[index])
}

// readKey reads one org.gnome.desktop.input-sources key through the runner.
func readKey(ctx context.Context, run Runner, key string) ([]byte, error) {
	return run(ctx, binGSettings, []string{"get", gsettingsSchema, key})
}

// parseSources reads the sources output strictly, returning the engine/
// source NAME (the second element) of every tuple — the thin adapter over
// the canonical ParseSourceTuples the IfOwned verdict consumes. Any
// deviation is malformed: nil.
func parseSources(out []byte) []string {
	tuples, err := ParseSourceTuples(string(out))
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(tuples))
	for _, t := range tuples {
		names = append(names, t.ID)
	}

	return names
}

// currentPrefix is the literal type word gsettings prints before the
// current index; errMalformedCurrent is its parse-failure sentinel.
const currentPrefix = "uint32"

var errMalformedCurrent = errors.New("malformed current")

// parseCurrent reads the current-index output strictly as `uint32 <N>`.
func parseCurrent(out []byte) (uint32, error) {
	value := strings.TrimSpace(string(out))
	fields := strings.Fields(value)
	if len(fields) != 2 || fields[0] != currentPrefix {
		return 0, fmt.Errorf("%w: %q", errMalformedCurrent, value)
	}
	n, err := strconv.ParseUint(fields[1], 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%w %q: %w", errMalformedCurrent, value, err)
	}

	return uint32(n), nil
}

// reactivate runs the bounded retry loop for one owned engine name.
// Intermediate attempt failures stay DEBUG; success logs INFO; exhausted
// retries warn with the last error. Every step respects ctx — a cancelled
// context aborts the loop immediately, never sleeping past cancellation.
func reactivate(ctx context.Context, run Runner, name string) {
	var lastErr error
	for attempt := 1; ; attempt++ {
		if err := ctx.Err(); err != nil {
			slog.Debug("engine reactivation aborted", "engine", name, "attempt", attempt, "error", err)

			return
		}
		if _, err := run(ctx, binIbus, []string{"engine", name}); err != nil {
			lastErr = err
			slog.Debug("engine activation attempt failed", "engine", name, "attempt", attempt, "error", err)
		} else {
			slog.Info("engine reactivated", "engine", name)

			return
		}
		if attempt >= activationAttempts {
			break
		}
		if !sleep(ctx, activationRetryDelay) {
			slog.Debug("engine reactivation aborted", "engine", name, "attempt", attempt, "error", ctx.Err())

			return
		}
	}
	slog.Warn("engine reactivation failed", "engine", name, "attempts", activationAttempts, "error", lastErr)
}

// sleep waits d or until ctx is done; it reports whether the full delay
// elapsed.
func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
