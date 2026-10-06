//nolint:testpackage // drives the unexported resolver seams — the sound corpus in-package precedent
package sound

import (
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
)

// The corpus's fixture vocabulary: two ordered XDG roots and the event
// names the daemon plays (the schema constant flip event and the config
// autocorrect event default).
const (
	rootHome = "/home/u/.local/share/sounds"
	rootData = "/usr/share/sounds"

	fsEventBell    = "bell"
	fsEventMessage = "message"
)

// errFakeFsMissing is the opener double's miss — the "no such file" of the
// in-memory theme filesystem.
var errFakeFsMissing = errors.New("theme file not present in the fake filesystem")

// fakeThemeFS is the fake-filesystem seam: a path-to-content map behind the
// resolver's opener — no real disk anywhere in the unit corpus.
type fakeThemeFS map[string]string

// open is the resolver's opener double: a mapped path answers a reader
// over its content, everything else the miss.
//
//nolint:ireturn // the fake hands the opened-reader interface back (the production seam's shape)
func (fs fakeThemeFS) open(path string) (io.ReadCloser, error) {
	if data, ok := fs[path]; ok {
		return io.NopCloser(strings.NewReader(data)), nil
	}

	return nil, errFakeFsMissing
}

// drainThemeFile reads one opened theme file to its end and closes it.
func drainThemeFile(t *testing.T, rc io.ReadCloser) string {
	t.Helper()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read theme file: %v", err)
	}
	if err := rc.Close(); err != nil {
		t.Errorf("close theme file: %v", err)
	}

	return string(data)
}

// yaruIndex is the Yaru index.theme form the owner's desktop carries:
// one [Sound Theme] section with a single Directories entry.
const yaruIndex = "[Sound Theme]\nDirectories=stereo\n"

// TestResolver_DataHomeBeatsDataDirs pins the discovery order (XDG Sound
// Theme spec): $XDG_DATA_HOME/sounds is searched BEFORE $XDG_DATA_DIRS/sounds
// — when both roots carry the same theme file, the data-home copy wins.
func TestResolver_DataHomeBeatsDataDirs(t *testing.T) {
	fs := fakeThemeFS{
		rootHome + "/yaru/index.theme":                  yaruIndex,
		rootHome + "/yaru/stereo/" + fsEventBell + ".oga": "HOME",
		rootData + "/yaru/index.theme":                  yaruIndex,
		rootData + "/yaru/stereo/" + fsEventBell + ".oga": "DATA-DIRS",
	}
	r := newThemeResolver(fs.open, []string{rootHome, rootData})

	rc, err := r.resolve("yaru", fsEventBell)
	if err != nil {
		t.Fatalf("resolve yaru/bell: %v", err)
	}
	if got := drainThemeFile(t, rc); got != "HOME" {
		t.Errorf("resolved content = %q, want the data-home copy —"+
			" $XDG_DATA_HOME/sounds must be searched before $XDG_DATA_DIRS/sounds", got)
	}
}

// TestResolver_FallsThroughToDataDirs pins the second rung of the
// discovery order: a theme file only present under $XDG_DATA_DIRS/sounds
// still resolves.
func TestResolver_FallsThroughToDataDirs(t *testing.T) {
	fs := fakeThemeFS{
		rootData + "/yaru/index.theme":                  yaruIndex,
		rootData + "/yaru/stereo/" + fsEventBell + ".oga": "DATA-DIRS",
	}
	r := newThemeResolver(fs.open, []string{rootHome, rootData})

	rc, err := r.resolve("yaru", fsEventBell)
	if err != nil {
		t.Fatalf("resolve yaru/bell: %v", err)
	}
	if got := drainThemeFile(t, rc); got != "DATA-DIRS" {
		t.Errorf("resolved content = %q, want the data-dirs copy", got)
	}
}

// TestResolver_ExtensionPriority pins the locked extension rungs: within a
// theme's directory the event resolves as <event>.oga, then .ogg, then
// .wav — each lower rung only when the higher ones are absent.
func TestResolver_ExtensionPriority(t *testing.T) {
	index := map[string]string{rootData + "/yaru/index.theme": yaruIndex}
	oga := map[string]string{rootData + "/yaru/stereo/" + fsEventBell + ".oga": "OGA"}
	ogg := map[string]string{rootData + "/yaru/stereo/" + fsEventBell + ".ogg": "OGG"}
	wav := map[string]string{rootData + "/yaru/stereo/" + fsEventBell + ".wav": "WAV"}

	cells := []struct {
		name string
		fs   map[string]string
		want string
	}{
		{"oga beats ogg and wav", mergeFiles(index, oga, ogg, wav), "OGA"},
		{"ogg beats wav without oga", mergeFiles(index, ogg, wav), "OGG"},
		{"wav is the last rung", mergeFiles(index, wav), "WAV"},
	}
	for _, cell := range cells {
		t.Run(cell.name, func(t *testing.T) {
			r := newThemeResolver(fakeThemeFS(cell.fs).open, []string{rootData})
			rc, err := r.resolve("yaru", fsEventBell)
			if err != nil {
				t.Fatalf("resolve yaru/bell: %v", err)
			}
			if got := drainThemeFile(t, rc); got != cell.want {
				t.Errorf("resolved content = %q, want %q — the extension priority broke", got, cell.want)
			}
		})
	}
}

// mergeFiles merges the corpus's path maps into one fake filesystem.
func mergeFiles(maps ...map[string]string) map[string]string {
	out := make(map[string]string)
	for _, m := range maps {
		for path, data := range m {
			out[path] = data
		}
	}

	return out
}

// TestResolver_InheritsChainFirstHitWins pins the Inherits chain: the
// resolver walks theme → its inherits, and the FIRST hit across the whole
// chain wins — the theme's own file beats anything inherited.
func TestResolver_InheritsChainFirstHitWins(t *testing.T) {
	const customIndex = "[Sound Theme]\nDirectories=stereo\nInherits=yaru\n"

	t.Run("own file beats the inherited one", func(t *testing.T) {
		fs := fakeThemeFS{
			rootData + "/custom/index.theme":                  customIndex,
			rootData + "/custom/stereo/" + fsEventBell + ".oga": "CUSTOM",
			rootData + "/yaru/index.theme":                    yaruIndex,
			rootData + "/yaru/stereo/" + fsEventBell + ".oga": "INHERITED",
		}
		r := newThemeResolver(fs.open, []string{rootData})

		rc, err := r.resolve("custom", fsEventBell)
		if err != nil {
			t.Fatalf("resolve custom/bell: %v", err)
		}
		if got := drainThemeFile(t, rc); got != "CUSTOM" {
			t.Errorf("resolved content = %q, want the theme's own file — the chain's first hit wins", got)
		}
	})

	t.Run("empty theme falls through to its inherit", func(t *testing.T) {
		fs := fakeThemeFS{
			rootData + "/custom/index.theme":                  customIndex,
			rootData + "/yaru/index.theme":                    yaruIndex,
			rootData + "/yaru/stereo/" + fsEventBell + ".oga": "INHERITED",
		}
		r := newThemeResolver(fs.open, []string{rootData})

		rc, err := r.resolve("custom", fsEventBell)
		if err != nil {
			t.Fatalf("resolve custom/bell: %v", err)
		}
		if got := drainThemeFile(t, rc); got != "INHERITED" {
			t.Errorf("resolved content = %q, want the inherited file", got)
		}
	})
}

// TestResolver_InheritsCycleTerminatesAtFallback pins the cycle safety: a
// theme pair inheriting each other terminates through the visited set and
// the walk still reaches the spec's fallback theme — a hostile index.theme
// must never hang the resolver (T-SQU-01).
func TestResolver_InheritsCycleTerminatesAtFallback(t *testing.T) {
	fs := fakeThemeFS{
		rootData + "/a/index.theme": "[Sound Theme]\nDirectories=stereo\nInherits=b\n",
		rootData + "/b/index.theme": "[Sound Theme]\nDirectories=stereo\nInherits=a\n",
		rootData + "/" + fallbackTheme + "/index.theme": yaruIndex,
		rootData + "/" + fallbackTheme + "/stereo/" + fsEventBell + ".ogg": "FALLBACK",
	}
	r := newThemeResolver(fs.open, []string{rootData})

	rc, err := r.resolve("a", fsEventBell)
	if err != nil {
		t.Fatalf("resolve a/bell through the cycle: %v", err)
	}
	if got := drainThemeFile(t, rc); got != "FALLBACK" {
		t.Errorf("resolved content = %q, want the fallback theme's file", got)
	}
}

// TestResolver_EmptyDirectoriesYieldsNothing pins the defensive index
// parse: a theme with absent or empty Directories contributes nothing to
// the search (the walk moves on down the chain) and never panics.
func TestResolver_EmptyDirectoriesYieldsNothing(t *testing.T) {
	fs := fakeThemeFS{
		rootData + "/hollow/index.theme": "[Sound Theme]\nInherits=yaru\n",
		rootData + "/bare/index.theme":   "[Sound Theme]\nDirectories=\nInherits=yaru\n",
		rootData + "/yaru/index.theme":   yaruIndex,
		rootData + "/yaru/stereo/" + fsEventBell + ".oga": "YARU",
	}
	r := newThemeResolver(fs.open, []string{rootData})

	for _, theme := range []string{"hollow", "bare"} {
		rc, err := r.resolve(theme, fsEventBell)
		if err != nil {
			t.Fatalf("resolve %s/bell: %v", theme, err)
		}
		if got := drainThemeFile(t, rc); got != "YARU" {
			t.Errorf("%s: resolved content = %q, want the chain's answering theme", theme, got)
		}
	}
}

// TestResolver_TotalMissIsAnError pins the miss contract: a total miss
// across the whole chain is a resolvable sentinel error — the caller
// synthesizes (never a panic, never a wait).
func TestResolver_TotalMissIsAnError(t *testing.T) {
	r := newThemeResolver(fakeThemeFS{}.open, []string{rootData})

	if _, err := r.resolve("yaru", fsEventBell); !errors.Is(err, errThemeEventMissing) {
		t.Fatalf("total miss err = %v, want errThemeEventMissing", err)
	}
}

// TestParseSoundThemeIndex pins the small INI-section reader: the
// [Sound Theme] section's Directories (the single-entry Yaru form, comma-
// and semicolon-separated lists accepted defensively) and the Inherits
// chain; sections other than [Sound Theme] contribute nothing.
func TestParseSoundThemeIndex(t *testing.T) {
	cells := []struct {
		name string
		doc  string
		want soundThemeIndex
	}{
		{
			name: "yaru single-entry form",
			doc:  yaruIndex,
			want: soundThemeIndex{directories: []string{"stereo"}},
		},
		{
			name: "comma-separated list",
			doc:  "[Sound Theme]\nDirectories=stereo,surround\n",
			want: soundThemeIndex{directories: []string{"stereo", "surround"}},
		},
		{
			name: "semicolon-separated list",
			doc:  "[Sound Theme]\nDirectories=stereo;surround;4.0\n",
			want: soundThemeIndex{directories: []string{"stereo", "surround", "4.0"}},
		},
		{
			name: "mixed separators with spaces and a trailing separator",
			doc:  "[Sound Theme]\nDirectories=stereo, surround;\n",
			want: soundThemeIndex{directories: []string{"stereo", "surround"}},
		},
		{
			name: "inherits chain",
			doc:  "[Sound Theme]\nDirectories=stereo\nInherits=yaru,freedesktop\n",
			want: soundThemeIndex{directories: []string{"stereo"}, inherits: []string{"yaru", fallbackTheme}},
		},
		{
			name: "foreign section contributes nothing",
			doc:  "[X-Whatever]\nDirectories=stereo\nInherits=yaru\n",
			want: soundThemeIndex{},
		},
		{
			name: "empty document",
			doc:  "",
			want: soundThemeIndex{},
		},
	}
	for _, cell := range cells {
		t.Run(cell.name, func(t *testing.T) {
			got := parseSoundThemeIndex([]byte(cell.doc))
			if !slices.Equal(got.directories, cell.want.directories) {
				t.Errorf("directories = %q, want %q", got.directories, cell.want.directories)
			}
			if !slices.Equal(got.inherits, cell.want.inherits) {
				t.Errorf("inherits = %q, want %q", got.inherits, cell.want.inherits)
			}
		})
	}
}

// TestXDGSoundRoots pins the root list's shape: $XDG_DATA_HOME/sounds (or
// its $HOME default) BEFORE the $XDG_DATA_DIRS/sounds entries (or the
// XDG-spec default pair).
func TestXDGSoundRoots(t *testing.T) {
	t.Run("spec defaults", func(t *testing.T) {
		got := xdgSoundRoots("", "", "/home/u")
		want := []string{rootHome, "/usr/local/share/sounds", rootData}
		if !slices.Equal(got, want) {
			t.Errorf("roots = %q, want %q", got, want)
		}
	})

	t.Run("explicit environment", func(t *testing.T) {
		got := xdgSoundRoots("/x/home", "/a:/b", "/home/u")
		want := []string{"/x/home/sounds", "/a/sounds", "/b/sounds"}
		if !slices.Equal(got, want) {
			t.Errorf("roots = %q, want %q", got, want)
		}
	})
}
