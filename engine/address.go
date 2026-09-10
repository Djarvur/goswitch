package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// errNoAddress reports that no usable IBus address file was found.
var errNoAddress = errors.New("no usable ibus address file")

// errNoAddressLine reports a bus file without an IBUS_ADDRESS entry.
var errNoAddressLine = errors.New("no IBUS_ADDRESS line")

// Discover returns the address of the live IBus daemon's private socket.
//
// Precedence: the IBUS_ADDRESS environment variable, else the newest file
// in ~/.config/ibus/bus/. The daemon writes a fresh socket per generation
// (stale files accumulate), so callers must re-discover on every connection
// attempt and never cache the result.
func Discover() (string, error) {
	if addr := os.Getenv("IBUS_ADDRESS"); addr != "" {
		return addr, nil
	}
	dir := filepath.Join(os.Getenv("HOME"), ".config", "ibus", "bus")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", dir, err)
	}
	var (
		best     string
		bestWhen time.Time
	)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		addr, err := readAddressFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		if best == "" || info.ModTime().After(bestWhen) {
			best, bestWhen = addr, info.ModTime()
		}
	}
	if best == "" {
		return "", fmt.Errorf("%w in %s", errNoAddress, dir)
	}

	return best, nil
}

// readAddressFile parses an IBus bus file and returns its IBUS_ADDRESS
// value (unix:path=...,guid=...).
func readAddressFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if v, ok := strings.CutPrefix(line, "IBUS_ADDRESS="); ok {
			return v, nil
		}
	}

	return "", fmt.Errorf("%w in %s", errNoAddressLine, path)
}
