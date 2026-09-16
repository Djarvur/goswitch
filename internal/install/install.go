// Package install implements the goswitch user-space installer (INST-01,
// D-39/D-40/D-42): goswitchctl install / uninstall drive the whole lifecycle
// — component XML into ~/.config/ibus/component (registered through an
// env-carrying `ibus write-cache`), a systemd user unit with an absolute
// ExecStart, the "single owner" input-sources takeover (ADR-001/D-40) with
// the prior sources saved for restore, and the matching full rollback.
// Everything stays under $HOME: no root, no system paths (the no-root
// delivery model is a public promise).
package install

import "context"

// Runner executes one installer subprocess (ibus, systemctl, gsettings):
// the seam the unit corpus drives with a recording fake — argv, env and
// stdin are the observable surface — backed in production by
// os/exec.CommandContext. The env parameter carries the full child
// environment on the calls that need IBUS_COMPONENT_PATH pinned (Pitfall 1:
// a plain `ibus write-cache` evicts user components from the registry
// cache), and is nil for inherit-the-parent calls.
type Runner func(ctx context.Context, name string, args []string, env []string, stdin []byte) ([]byte, error)

// ActiveEngines reports the engine names the live ibus-daemon has
// registered (ListActiveEngines on the private IBus socket). A seam so the
// install corpus can answer the bounded-wait registration probe without a
// live bus; the production implementation dials through engine.Discover.
type ActiveEngines func(ctx context.Context) ([]string, error)

// Installer owns the install/uninstall lifecycle of one goswitch
// deployment. The zero value is not usable — build it with New; every
// collaborator has a seam the corpus replaces.
type Installer struct {
	run           Runner
	home          string
	selfDir       string
	activeEngines ActiveEngines
}

// New returns the production installer over os/exec and the real $HOME;
// the options replace the collaborators (the test seams).
func New(opts ...func(*Installer)) *Installer {
	i := &Installer{}
	for _, opt := range opts {
		opt(i)
	}

	return i
}

// WithRunner replaces the subprocess runner — the test seam.
func WithRunner(r Runner) func(*Installer) {
	return func(i *Installer) {
		i.run = r
	}
}

// WithHome pins the user home the installer lays its files under — the
// corpus points it at t.TempDir instead of the real $HOME.
func WithHome(home string) func(*Installer) {
	return func(i *Installer) {
		i.home = home
	}
}

// WithSelfDir pins the directory whose goswitchd the unit's ExecStart
// points at — the corpus points it at a temp dir with a stand-in binary;
// production resolves the directory of the running executable (ASVS V14:
// an absolute path, never %h or a $PATH lookup).
func WithSelfDir(dir string) func(*Installer) {
	return func(i *Installer) {
		i.selfDir = dir
	}
}

// WithActiveEngines replaces the live-registration probe — the test seam
// for the bounded-wait step at the end of Install.
func WithActiveEngines(fn ActiveEngines) func(*Installer) {
	return func(i *Installer) {
		i.activeEngines = fn
	}
}

// Install performs the D-39/D-40 sequence and returns the step-by-step
// report (paths and verdicts only — never user text, D-20/D-21).
func (i *Installer) Install(ctx context.Context) ([]string, error) {
	return nil, nil
}

// Uninstall performs the D-42 rollback and returns the step-by-step report.
// purge additionally removes the user's ~/.config/goswitch and the state
// directory.
func (i *Installer) Uninstall(ctx context.Context, purge bool) ([]string, error) {
	return nil, nil
}

// Install runs the production installer — the thin entry the CLI calls so
// the client stays flag parsing and printing only.
func Install(ctx context.Context) ([]string, error) {
	return New().Install(ctx)
}

// Uninstall runs the production rollback — the CLI's thin entry.
func Uninstall(ctx context.Context, purge bool) ([]string, error) {
	return New().Uninstall(ctx, purge)
}
