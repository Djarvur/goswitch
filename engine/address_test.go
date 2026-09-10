package engine_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Djarvur/goswitch/engine"
)

// writeBusFile creates one fake IBus bus file with the given address and
// modification time, and returns its path.
func writeBusFile(t *testing.T, dir, name, addr string, when time.Time) {
	t.Helper()

	path := filepath.Join(dir, name)
	content := "IBUS_ADDRESS=" + addr + "\nIBUS_DAEMON_PID=1\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write bus file %s: %v", path, err)
	}
	if err := os.Chtimes(path, when, when); err != nil {
		t.Fatalf("chtimes bus file %s: %v", path, err)
	}
}

// busDirWith creates an isolated HOME whose ~/.config/ibus/bus holds the
// given files, and points the environment at it.
func busDirWith(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	dir := filepath.Join(home, ".config", "ibus", "bus")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir bus dir: %v", err)
	}

	t.Setenv("HOME", home)

	return dir
}

func TestAddress_EnvPrecedence(t *testing.T) {
	dir := busDirWith(t)
	base := time.Now().Add(-time.Hour)
	writeBusFile(t, dir, "mid-unix-wayland-0", "unix:path=/tmp/env-should-win", base)
	writeBusFile(t, dir, "mid2-unix-wayland-0", "unix:path=/tmp/fresh-should-lose", base.Add(time.Minute))
	t.Setenv("IBUS_ADDRESS", "unix:path=/tmp/from-env")

	addr, err := engine.Discover()
	if err != nil {
		t.Fatalf("Discover() unexpected err: %v", err)
	}
	if addr != "unix:path=/tmp/from-env" {
		t.Errorf("Discover() = %q, want the IBUS_ADDRESS env value to take precedence", addr)
	}
}

func TestAddress_PicksFreshest(t *testing.T) {
	base := time.Now().Add(-2 * time.Hour)

	t.Run("newest among matching suffix", func(t *testing.T) {
		dir := busDirWith(t)
		t.Setenv("WAYLAND_DISPLAY", "wayland-0")
		writeBusFile(t, dir, "mid-unix-wayland-0", "unix:path=/tmp/old-wayland", base)
		writeBusFile(t, dir, "mid3-unix-wayland-0", "unix:path=/tmp/new-wayland", base.Add(time.Minute))

		addr, err := engine.Discover()
		if err != nil {
			t.Fatalf("Discover() unexpected err: %v", err)
		}
		if addr != "unix:path=/tmp/new-wayland" {
			t.Errorf("Discover() = %q, want the freshest matching file", addr)
		}
	})

	t.Run("foreign suffix ignored even when fresher", func(t *testing.T) {
		dir := busDirWith(t)
		t.Setenv("WAYLAND_DISPLAY", "wayland-0")
		writeBusFile(t, dir, "mid-unix-wayland-0", "unix:path=/tmp/old-wayland", base)
		writeBusFile(t, dir, "mid-unix-0", "unix:path=/tmp/fresh-x11", base.Add(time.Hour))

		addr, err := engine.Discover()
		if err != nil {
			t.Fatalf("Discover() unexpected err: %v", err)
		}
		if addr != "unix:path=/tmp/old-wayland" {
			t.Errorf("Discover() = %q, want the wayland file even though the X11 file is fresher", addr)
		}
	})

	t.Run("X11 suffix when only DISPLAY is set", func(t *testing.T) {
		dir := busDirWith(t)
		t.Setenv("WAYLAND_DISPLAY", "") // inherited from the live session otherwise
		t.Setenv("DISPLAY", ":0")
		writeBusFile(t, dir, "mid-unix-wayland-0", "unix:path=/tmp/fresh-wayland", base.Add(time.Hour))
		writeBusFile(t, dir, "mid-unix-0", "unix:path=/tmp/old-x11", base)

		addr, err := engine.Discover()
		if err != nil {
			t.Fatalf("Discover() unexpected err: %v", err)
		}
		if addr != "unix:path=/tmp/old-x11" {
			t.Errorf("Discover() = %q, want the X11 file selected by DISPLAY", addr)
		}
	})
}

func TestAddress_EmptyDirError(t *testing.T) {
	t.Run("empty dir", func(t *testing.T) {
		busDirWith(t)

		addr, err := engine.Discover()
		if err == nil {
			t.Fatalf("Discover() = %q, want a named error for an empty bus dir", addr)
		}
		if addr != "" {
			t.Errorf("Discover() addr = %q, want empty string alongside the error", addr)
		}
		if !errors.Is(err, engine.ErrNoAddress) {
			t.Errorf("Discover() err = %v, want errors.Is(err, engine.ErrNoAddress)", err)
		}
	})

	t.Run("missing dir", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home) // no bus dir at all

		addr, err := engine.Discover()
		if err == nil {
			t.Fatalf("Discover() = %q, want a named error for a missing bus dir", addr)
		}
		if addr != "" {
			t.Errorf("Discover() addr = %q, want empty string alongside the error", addr)
		}
		if !errors.Is(err, engine.ErrNoAddress) {
			t.Errorf("Discover() err = %v, want errors.Is(err, engine.ErrNoAddress)", err)
		}
	})
}
