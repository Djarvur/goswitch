// The XDG sound-theme resolver (todo «Звук», owner «делай» 2026-10-05):
// theme files live under the XDG data directories as
// <root>/sounds/<theme>/index.theme plus <theme>/<directory>/<event>.oga
// — the resolver walks a theme's Inherits chain (cycle-safe, terminating
// at the spec's fallback theme) and answers the FIRST hit with the locked
// extension priority. Everything here is a contained value: a hostile
// index.theme resolves to a miss, never a panic (T-SQU-01); the argv-free,
// shell-free opener carries pinned paths derived from XDG constants and
// schema event names only (T-SQU-04).

package sound

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// fallbackTheme is the XDG Sound Theme spec's fallback theme — the chain's
// terminal rung when neither the theme nor its inherits carry the event.
const fallbackTheme = "freedesktop"

// errThemeEventMissing is the total-miss sentinel: neither the theme, nor
// its inherits, nor the fallback theme carry the event — the caller
// synthesizes instead of waiting (never an error on the gesture path).
var errThemeEventMissing = errors.New("sound theme event not found")

// opener opens one theme file by absolute path — the fake-filesystem seam
// of the corpus (no real disk in unit tests).
type opener func(path string) (io.ReadCloser, error)

// themeExtensions is the locked extension priority of the event lookup
// within a theme's directory.
//
//nolint:gochecknoglobals // an immutable fixed-size table — the constant form Go refuses for arrays
var themeExtensions = [3]string{".oga", ".ogg", ".wav"}

// xdgRootsCapacity is the usual root-list size (home + the two default
// data dirs) — the slice's allocation hint.
const xdgRootsCapacity = 3

// themeResolver resolves (theme, event) pairs against an ordered list of
// XDG sound roots through the opener seam.
type themeResolver struct {
	open  opener
	roots []string // $XDG_DATA_HOME/sounds before the $XDG_DATA_DIRS/sounds entries
}

// newThemeResolver builds one resolver over the given ordered roots — the
// order IS the XDG discovery order (xdgSoundRoots builds it).
func newThemeResolver(open opener, roots []string) *themeResolver {
	return &themeResolver{open: open, roots: roots}
}

// xdgSoundRoots builds the ordered XDG sound roots: $XDG_DATA_HOME/sounds
// (default $HOME/.local/share/sounds) BEFORE the $XDG_DATA_DIRS/sounds
// entries (default /usr/local/share:/usr/share) — the XDG Sound Theme
// spec's discovery order.
func xdgSoundRoots(dataHome, dataDirs, home string) []string {
	roots := make([]string, 0, xdgRootsCapacity)
	if dataHome != "" {
		roots = append(roots, filepath.Join(dataHome, "sounds"))
	} else if home != "" {
		roots = append(roots, filepath.Join(home, ".local", "share", "sounds"))
	}
	if dataDirs == "" {
		dataDirs = "/usr/local/share:/usr/share"
	}
	for _, dir := range strings.Split(dataDirs, ":") {
		if dir != "" {
			roots = append(roots, filepath.Join(dir, "sounds"))
		}
	}

	return roots
}

// resolve returns one opened theme file for the (theme, event) pair: the
// walk runs theme → its Inherits chain → the spec's fallback theme,
// cycle-safe via a visited set (a hostile chain terminates, T-SQU-01);
// within each theme the directories carry the locked extension priority.
// The first hit across the whole chain wins; a total miss is the
// errThemeEventMissing sentinel — the caller synthesizes instead of
// waiting.
func (r *themeResolver) resolve(theme, event string) (io.ReadCloser, error) {
	visited := make(map[string]bool)
	queue := []string{theme}
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if visited[name] {
			continue
		}
		visited[name] = true
		index, ok := r.loadIndex(name)
		if !ok {
			continue // the theme's index lives nowhere — the chain moves on
		}
		if rc := r.probe(name, index.directories, event); rc != nil {
			return rc, nil
		}
		queue = append(queue, index.inherits...)
	}
	if !visited[fallbackTheme] {
		if index, ok := r.loadIndex(fallbackTheme); ok {
			if rc := r.probe(fallbackTheme, index.directories, event); rc != nil {
				return rc, nil
			}
		}
	}

	return nil, fmt.Errorf("resolve theme event %s/%s: %w", theme, event, errThemeEventMissing)
}

// loadIndex reads one theme's index.theme from the FIRST root carrying it
// — the first index.theme wins; a theme without any index contributes
// nothing to the search.
func (r *themeResolver) loadIndex(theme string) (soundThemeIndex, bool) {
	for _, root := range r.roots {
		rc, err := r.open(filepath.Join(root, theme, "index.theme"))
		if err != nil {
			continue
		}
		data, readErr := io.ReadAll(rc)
		_ = rc.Close()
		if readErr != nil {
			continue // a hostile/rotten index is a contained miss (T-SQU-01)
		}

		return parseSoundThemeIndex(data), true
	}

	return soundThemeIndex{}, false
}

// probe returns the first event file answering within one theme's
// directories — the locked extension priority per directory, roots in
// discovery order; nil when the theme carries no such event.
func (r *themeResolver) probe(theme string, directories []string, event string) io.ReadCloser {
	for _, dir := range directories {
		for _, root := range r.roots {
			for _, ext := range themeExtensions {
				rc, err := r.open(filepath.Join(root, theme, dir, event+ext))
				if err == nil {
					return rc
				}
			}
		}
	}

	return nil
}

// soundThemeIndex is the [Sound Theme] section's payload the resolver
// consumes: the theme's directory list and its Inherits chain.
type soundThemeIndex struct {
	directories []string
	inherits    []string
}

// parseSoundThemeIndex reads the [Sound Theme] section of one index.theme:
// the Directories list (the single-entry Yaru form; comma- and
// semicolon-separated lists accepted defensively) and the Inherits chain.
// Anything outside the section — and every malformed line — is skipped: a
// hostile index parses to an empty payload, never a panic (T-SQU-01).
func parseSoundThemeIndex(data []byte) soundThemeIndex {
	var (
		index     soundThemeIndex
		inSection bool
	)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			inSection = line == "[Sound Theme]"

			continue
		}
		if !inSection {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		switch strings.TrimSpace(key) {
		case "Directories":
			index.directories = splitThemeList(value)
		case "Inherits":
			index.inherits = splitThemeList(value)
		}
	}

	return index
}

// splitThemeList splits one index.theme list value on its separators
// (comma and semicolon accepted alike, the Yaru single-entry form being
// the degenerate case) and drops the empties.
func splitThemeList(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';'
	})
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if field = strings.TrimSpace(field); field != "" {
			out = append(out, field)
		}
	}

	return out
}
