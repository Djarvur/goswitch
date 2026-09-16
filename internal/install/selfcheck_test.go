package install_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Djarvur/goswitch/internal/install"
)

// The selfcheck corpus's stand-ins: the healthy control-status reply (the
// ctlsvc wire canon with the D-37 version leading), the fix hints every
// red verdict must carry (named once for goconst), and the env key the
// repair write-cache must pin (Pitfall 1).
const (
	ctlStatusHealthy = "version=dev mode=en corrections_done=0 config=none"
	ctlStatusNoVer   = "mode=en corrections_done=0 config=none"

	hintUnitStart  = "systemctl --user start goswitchd"
	hintUnitEnable = "systemctl --user enable --now goswitchd"
	hintJournal    = "journalctl --user -u goswitchd"
	hintInstall    = "goswitchctl install"

	envComponentPath = "IBUS_COMPONENT_PATH"
	opListEngine     = "list-engine"

	// stepComponentName mirrors the audit's step name for the corpus's
	// transparency pin (the install-package const is unexported).
	stepComponentName = "component-visible"
)

// selfcheckConfigYAML is a COMPLETE valid config document (the 03-02
// strict-parse shape: every section present, no unknown keys) for the
// config-valid green path.
const selfcheckConfigYAML = `hotkeys:
  tap_key: shift_r
  word_layout_combo: shift+ctrl_r
timeouts:
  tap_window_ms: 300
  verify_wait_ms: 100
correction:
  backspace_cap: 50
  clipboard_rung: false
macr:
  enabled: false
  letters: ""
  apps: []
  alt_modifier: ""
`

// errFakeCtlDown is the corpus's static daemon-silent error (err113).
var errFakeCtlDown = errors.New("name has no owner")

// selfcheckGreenStub answers the healthy desktop for the audit: the unit
// active and the single-owner sources set; list-engine falls through to
// defaultReply's registry hit.
func selfcheckGreenStub(name string, args []string) ([]byte, error) {
	if name == binSystemctl && len(args) == 3 && args[1] == "is-active" {
		return []byte("active"), nil
	}
	if name == binGSettings && len(args) == 3 && args[0] == opGet {
		return []byte(goswitchSources), nil
	}

	return defaultReply(name, args)
}

// redUnitStub answers an inactive unit over the otherwise-green desktop.
func redUnitStub(name string, args []string) ([]byte, error) {
	if name == binSystemctl && len(args) == 3 && args[1] == "is-active" {
		return []byte("inactive"), nil
	}

	return selfcheckGreenStub(name, args)
}

// redSourcesStub answers the pre-install owner sources over the
// otherwise-green desktop.
func redSourcesStub(name string, args []string) ([]byte, error) {
	if name == binGSettings && len(args) == 3 && args[0] == opGet {
		return []byte(ownerSources), nil
	}

	return selfcheckGreenStub(name, args)
}

// newSelfchecker builds an installer for the selfcheck corpus: fake runner,
// the given $HOME, canned ctlStatus and activeEngines — no live bus, no
// real desktop behind the audit.
func newSelfchecker(
	t *testing.T, f *fakeRunner, home, status string, statusErr error, engines []string,
) *install.Installer {
	t.Helper()

	return install.New(
		install.WithRunner(f.run),
		install.WithHome(home),
		install.WithCtlStatus(func(context.Context) (string, error) { return status, statusErr }),
		install.WithActiveEngines(func(context.Context) ([]string, error) {
			return engines, nil
		}),
		install.WithRegistryProbe(func() bool { return true }),
	)
}

// assertSelfcheckLines pins the full green transcript: the six D-41 verdicts
// in order, one line each.
func assertSelfcheckLines(t *testing.T, out string) {
	t.Helper()

	want := []string{
		"ok version",
		"ok component-visible",
		"ok unit-active",
		"ok engine-registered",
		"ok config (defaults: no config file)",
		"ok input-source",
	}
	got := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if !slices.Equal(got, want) {
		t.Errorf("selfcheck lines = %q, want %q", got, want)
	}
}

// assertLastLineFail pins the fail-fast discipline: the run stops at the
// first red verdict — the FAIL line is the last line printed.
func assertLastLineFail(t *testing.T, out string) {
	t.Helper()

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[len(lines)-1], "FAIL ") {
		t.Errorf("last line of %q is not a FAIL verdict (fail-fast broken)", out)
	}
}

// runRed drives one red scenario: the audit must fail, name the step in a
// FAIL line carrying the fix hint, and stop there.
func runRed(t *testing.T, i *install.Installer, wantFail, wantHint string) {
	t.Helper()

	var buf bytes.Buffer
	if err := i.Selfcheck(context.Background(), &buf); err == nil {
		t.Fatalf("Selfcheck error = nil, want the %q red verdict to fail the run", wantFail)
	}
	out := buf.String()
	if !strings.Contains(out, wantFail) {
		t.Errorf("selfcheck output %q misses %q", out, wantFail)
	}
	if !strings.Contains(out, wantHint) {
		t.Errorf("selfcheck output %q misses the fix hint %q", out, wantHint)
	}
	assertLastLineFail(t, out)
}

// countWriteCaches snapshots the recorded calls and counts the repair
// write-cache invocations.
func countWriteCaches(f *fakeRunner) int {
	n := 0
	for _, c := range f.snapshot() {
		if c.name == binIbus && len(c.args) > 0 && c.args[0] == opWriteCache {
			n++
		}
	}

	return n
}

// TestSelfcheck_AllGreen pins the green transcript: the six D-41 steps in
// order (version → component → unit → engine → config → source), each an
// "ok" line, and a nil error.
func TestSelfcheck_AllGreen(t *testing.T) {
	f := &fakeRunner{stub: selfcheckGreenStub}
	i := newSelfchecker(t, f, t.TempDir(), ctlStatusHealthy, nil, []string{engineENName})

	var buf bytes.Buffer
	if err := i.Selfcheck(context.Background(), &buf); err != nil {
		t.Fatalf("Selfcheck error = %v, want nil", err)
	}
	assertSelfcheckLines(t, buf.String())
}

// TestSelfcheck_ComponentRepair pins the repair path (research Pitfall 1,
// T-04-02-01): the first list-engine answer misses goswitch → the audit
// re-runs the env-carrying write-cache EXACTLY once → the re-check is green,
// transparently (one ok verdict, no repair chatter).
func TestSelfcheck_ComponentRepair(t *testing.T) {
	f := &fakeRunner{}
	listProbes := 0
	f.stub = func(name string, args []string) ([]byte, error) {
		if name == binIbus && len(args) > 0 && args[0] == opListEngine {
			listProbes++
			if listProbes == 1 {
				return []byte("other-engine - Not Goswitch\n"), nil
			}

			return []byte(listEngineOut), nil
		}

		return selfcheckGreenStub(name, args)
	}
	i := newSelfchecker(t, f, t.TempDir(), ctlStatusHealthy, nil, []string{engineENName})

	var buf bytes.Buffer
	if err := i.Selfcheck(context.Background(), &buf); err != nil {
		t.Fatalf("Selfcheck error = %v, want nil (the repair must recover the step)", err)
	}
	verdict := "ok component-visible"
	if out := buf.String(); strings.Count(out, stepComponentName) != 1 || !strings.Contains(out, verdict) {
		t.Errorf("selfcheck output %q must carry exactly one transparent %q verdict", out, verdict)
	}
	if repairs := countWriteCaches(f); repairs != 1 {
		t.Errorf("repair write-cache count = %d, want exactly 1 (T-04-02-01: never a loop)", repairs)
	}
	for _, c := range f.snapshot() {
		if c.name == binIbus && len(c.args) > 0 && c.args[0] == opWriteCache &&
			!slices.ContainsFunc(c.env, func(kv string) bool { return strings.HasPrefix(kv, envComponentPath+"=") }) {
			t.Errorf("repair write-cache env %v misses %s (Pitfall 1)", c.env, envComponentPath)
		}
	}
}

// TestSelfcheck_ComponentRedAfterRepair pins the exhausted repair: both
// list-engine answers miss goswitch → one write-cache repair, then the FAIL
// verdict with the install hint and a failed run — never a retry loop.
func TestSelfcheck_ComponentRedAfterRepair(t *testing.T) {
	f := &fakeRunner{}
	f.stub = func(name string, args []string) ([]byte, error) {
		if name == binIbus && len(args) > 0 && args[0] == opListEngine {
			return []byte("other-engine - Not Goswitch\n"), nil
		}

		return selfcheckGreenStub(name, args)
	}
	i := newSelfchecker(t, f, t.TempDir(), ctlStatusHealthy, nil, []string{engineENName})

	var buf bytes.Buffer
	err := i.Selfcheck(context.Background(), &buf)
	if err == nil {
		t.Fatal("Selfcheck error = nil, want the exhausted component step to fail the run")
	}
	out := buf.String()
	if !strings.Contains(out, "FAIL component-visible") || !strings.Contains(out, hintInstall) {
		t.Errorf("selfcheck output %q must carry FAIL component-visible with the %q hint", out, hintInstall)
	}
	if repairs := countWriteCaches(f); repairs != 1 {
		t.Errorf("repair write-cache count = %d, want exactly 1 (T-04-02-01: never a loop)", repairs)
	}
	assertLastLineFail(t, out)
}

// TestSelfcheck_EachRedPath pins every step's red verdict with its fix
// hint: the daemon silent, the unit inactive, the engine unregistered, the
// config invalid (path + parse reason) and the sources not the goswitch
// owner — each fails the run and stops it (fail-fast).
func TestSelfcheck_EachRedPath(t *testing.T) {
	t.Run("version: daemon not answering", func(t *testing.T) {
		f := &fakeRunner{stub: selfcheckGreenStub}
		i := newSelfchecker(t, f, t.TempDir(), "", errFakeCtlDown, []string{engineENName})
		runRed(t, i, "FAIL version", hintUnitStart)
	})

	t.Run("unit inactive", func(t *testing.T) {
		f := &fakeRunner{stub: redUnitStub}
		i := newSelfchecker(t, f, t.TempDir(), ctlStatusHealthy, nil, []string{engineENName})
		runRed(t, i, "FAIL unit-active", hintUnitEnable)
	})

	t.Run("engine not registered", func(t *testing.T) {
		f := &fakeRunner{stub: selfcheckGreenStub}
		i := newSelfchecker(t, f, t.TempDir(), ctlStatusHealthy, nil, []string{"other-engine"})
		runRed(t, i, "FAIL engine-registered", hintJournal)
	})

	t.Run("config invalid: path and parse reason", func(t *testing.T) {
		runConfigRed(t)
	})

	t.Run("input source not the goswitch owner", func(t *testing.T) {
		f := &fakeRunner{stub: redSourcesStub}
		i := newSelfchecker(t, f, t.TempDir(), ctlStatusHealthy, nil, []string{engineENName})
		runRed(t, i, "FAIL input-source", hintInstall)
	})
}

// runConfigRed drives the config step's red scenario: an existing document
// that fails the strict decode fails the run, with the path and the parse
// reason in the verdict.
func runConfigRed(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	cfgDir := filepath.Join(home, ".config", "goswitch")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("not_a_known_key: true\n"), 0o600); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}
	f := &fakeRunner{stub: selfcheckGreenStub}
	i := newSelfchecker(t, f, home, ctlStatusHealthy, nil, []string{engineENName})

	var buf bytes.Buffer
	if err := i.Selfcheck(context.Background(), &buf); err == nil {
		t.Fatal("Selfcheck error = nil, want the invalid config to fail the run")
	}
	out := buf.String()
	if !strings.Contains(out, "FAIL config") || !strings.Contains(out, cfgPath) {
		t.Errorf("selfcheck output %q must carry FAIL config with the path %q", out, cfgPath)
	}
	if !strings.Contains(out, "not_a_known_key") {
		t.Errorf("selfcheck output %q must name the parse reason", out)
	}
	assertLastLineFail(t, out)
}

// TestSelfcheck_ConfigNoFileGreen pins the no-file green (the 04-planner
// decision: install generates no config — the daemon runs on the built-in
// defaults), plus the valid-explicit-config green with its path.
func TestSelfcheck_ConfigNoFileGreen(t *testing.T) {
	f := &fakeRunner{stub: selfcheckGreenStub}
	i := newSelfchecker(t, f, t.TempDir(), ctlStatusHealthy, nil, []string{engineENName})

	var buf bytes.Buffer
	if err := i.Selfcheck(context.Background(), &buf); err != nil {
		t.Fatalf("Selfcheck error = %v, want nil", err)
	}
	if out := buf.String(); !strings.Contains(out, "ok config (defaults: no config file)") {
		t.Errorf("selfcheck output %q must carry the no-file defaults verdict", out)
	}

	t.Run("valid explicit config is green with its path", func(t *testing.T) {
		home := t.TempDir()
		cfgPath := filepath.Join(home, ".config", "goswitch", "config.yaml")
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
			t.Fatalf("mkdir config dir: %v", err)
		}
		if err := os.WriteFile(cfgPath, []byte(selfcheckConfigYAML), 0o600); err != nil {
			t.Fatalf("write valid config: %v", err)
		}
		f := &fakeRunner{stub: selfcheckGreenStub}
		i := newSelfchecker(t, f, home, ctlStatusHealthy, nil, []string{engineENName})

		var buf bytes.Buffer
		if err := i.Selfcheck(context.Background(), &buf); err != nil {
			t.Fatalf("Selfcheck error = %v, want nil", err)
		}
		if out := buf.String(); !strings.Contains(out, "ok config "+cfgPath) {
			t.Errorf("selfcheck output %q must carry the config path verdict", out)
		}
	})
}

// TestSelfcheck_StatusProbeSeam pins the version step's probe: the verdict
// follows the CtlStatus seam's answer — a reply without the version token
// is red, the healthy reply is green; a silent daemon fails with the
// unit-start hint (covered against the live desktop too — D-41 step 1).
func TestSelfcheck_StatusProbeSeam(t *testing.T) {
	t.Run("reply without version= is red", func(t *testing.T) {
		f := &fakeRunner{stub: selfcheckGreenStub}
		i := newSelfchecker(t, f, t.TempDir(), ctlStatusNoVer, nil, []string{engineENName})
		// The plan pins the unit-start hint for the SILENT daemon (the
		// EachRedPath case); a pre-version reply names its own fix — the
		// unit restart that puts a stamping-capable daemon in place.
		runRed(t, i, "FAIL version", "systemctl --user restart goswitchd")
	})

	t.Run("healthy reply is green", func(t *testing.T) {
		f := &fakeRunner{stub: selfcheckGreenStub}
		i := newSelfchecker(t, f, t.TempDir(), ctlStatusHealthy, nil, []string{engineENName})

		var buf bytes.Buffer
		if err := i.Selfcheck(context.Background(), &buf); err != nil {
			t.Fatalf("Selfcheck error = %v, want nil", err)
		}
		if out := buf.String(); !strings.Contains(out, "ok version") {
			t.Errorf("selfcheck output %q misses the ok version verdict", out)
		}
	})
}
