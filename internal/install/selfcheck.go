package install

import (
	"context"
	"io"
)

// CtlStatus reports the daemon's control-service status line (the ctlsvc
// Status reply) — the live probe behind the selfcheck's version step
// (D-41 step 1). The corpus replaces it; the production backing dials the
// session bus.
type CtlStatus func(ctx context.Context) (string, error)

// Selfcheck runs the six-step D-41 live audit — version → component →
// unit → engine → config → source — and prints one verdict line per step:
// "ok NAME" or "FAIL NAME: <fix hint>". The first red verdict ends the run
// (fail-fast, the e2e-preflight discipline) and the error carries it.
// STUB (RED of plan 04-02 Task 2): the audit does nothing yet.
func (i *Installer) Selfcheck(_ context.Context, _ io.Writer) error {
	return nil
}
