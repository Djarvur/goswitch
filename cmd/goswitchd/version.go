// Build identity of goswitchd (D-37): goreleaser stamps main.version,
// main.commit and main.date with its default ldflags on release builds;
// `go install` builds stay unstamped and honestly report the dev fallback.
// The build version is also threaded into the session actor so the control
// status identifies the running build.
package main

import "io"

// version, commit and date are the stamped build coordinates: goreleaser's
// default ldflags inject -X main.version -X main.commit -X main.date
// (research Pattern 3 — custom ldflags are NOT needed); dev builds keep the
// "dev" fallback and the empty commit/date.
var (
	version = "dev"
	commit  = ""
	date    = ""
)

// versionString renders the -version line: "goswitchd <version>" plus the
// commit and date when stamped, without hanging punctuation on dev builds.
// STUB (RED of plan 04-02 Task 1): the real formatting lands in GREEN.
func versionString() string {
	return ""
}

// runVersion prints the version line to w — the whole behavior of the
// -version flag: main calls it and exits 0 before run() ever starts the
// daemon. STUB (RED of plan 04-02 Task 1): prints nothing yet.
func runVersion(w io.Writer) {}
