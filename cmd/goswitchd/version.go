package main

import (
	"fmt"
	"io"
)

// Build identity of goswitchd (D-37): goreleaser stamps main.version,
// main.commit and main.date with its default ldflags on release builds;
// `go install` builds stay unstamped and honestly report the dev fallback.
// The build version is also threaded into the session actor so the control
// status identifies the running build.
//
//nolint:gochecknoglobals // package vars are the goreleaser -X stamping contract (D-37)
var (
	version = "dev"
	commit  = ""
	date    = ""
)

// versionString renders the -version line: "goswitchd <version>" plus the
// commit and date when stamped, without hanging punctuation on dev builds —
// the honest identification a bug report is read against (D-37).
func versionString() string {
	s := "goswitchd " + version
	if commit != "" {
		s += " " + commit
	}
	if date != "" {
		s += " " + date
	}

	return s
}

// runVersion prints the version line to w — the whole behavior of the
// -version flag: main calls it and exits 0 before run() ever starts the
// daemon, so a version query costs nothing and touches nothing.
func runVersion(w io.Writer) {
	_, _ = fmt.Fprintln(w, versionString())
}
