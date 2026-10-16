package main

// Validation audit (phase 04) gap test: the client half of the D-37 wire
// contract. The daemon leads its status line with version= (pinned server-side
// by internal/ctlsvc TestRenderStatusVersionToken); these tests pin that the
// generic client parser and the --json contract surface that token WITHOUT
// client edits — a fixed-token whitelist regression here would silently drop
// the build identity from `goswitchctl status` / `status --json`.

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

// statusVersionLine mirrors renderStatus output: version= leads, space-free
// key=value pairs precede config_error, and config_error closes the line with
// its whitespace-flattened value.
const statusVersionLine = "version=dev mode=en corrections_done=3 corrections_skipped=1 " +
	"skip_verify_timeout=2 super_intercepted=0 super_upstream_consumed=0 " +
	"config_path=/home/u/.config/goswitch/config.yaml config_valid=false " +
	"config_error=decode config: boom with spaces"

func TestParseStatusLine_ExtractsVersionToken(t *testing.T) {
	got := parseStatusLine(statusVersionLine)

	want := map[string]string{
		"version":                 "dev",
		"mode":                    "en",
		"corrections_done":        "3",
		"corrections_skipped":     "1",
		"skip_verify_timeout":     "2",
		"super_intercepted":       "0",
		"super_upstream_consumed": "0",
		"config_path":             "/home/u/.config/goswitch/config.yaml",
		"config_valid":            "false",
		"config_error":            "decode config: boom with spaces",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseStatusLine mismatch:\n got %#v\nwant %#v", got, want)
	}
}

func TestParseStatusLine_StampedVersionToken(t *testing.T) {
	// The release channel stamps the goreleaser default (tag without 'v',
	// 04-03/04-07 pin): the client parser must surface it verbatim.
	line := "version=1.0.0 mode=ru corrections_done=0 corrections_skipped=0 config=none"
	got := parseStatusLine(line)
	if got["version"] != "1.0.0" {
		// Stamped release builds identify themselves via the status token.
		t.Errorf("parseStatusLine version = %q, want %q", got["version"], "1.0.0")
	}
}

func TestStatusJSON_VersionTokenSurvives(t *testing.T) {
	// The --json contract: version stays a string, counters are numbers —
	// the D-37 token must ride through without client-side changes.
	var obj map[string]any
	if err := json.Unmarshal([]byte(statusJSON(statusVersionLine)), &obj); err != nil {
		t.Fatalf("statusJSON produced invalid JSON: %v", err)
	}

	if v, ok := obj["version"].(string); !ok || v != "dev" {
		t.Errorf("json version = %#v, want string %q", obj["version"], "dev")
	}
	if v, ok := obj["corrections_done"].(float64); !ok || v != 3 {
		t.Errorf("json corrections_done = %#v, want number 3", obj["corrections_done"])
	}
	if v, ok := obj["config_error"].(string); !ok || v != "decode config: boom with spaces" {
		t.Errorf("json config_error = %#v, want the flattened parse reason", obj["config_error"])
	}
}

// cliProbeForward is the probe subcommand's name in the CLI corpus (goconst).
const cliProbeForward = "probe-forward"

// errStubProbeDisabled is the fake seam's static refusal (err113).
var errStubProbeDisabled = errors.New("probe disabled")

// TestRun_ProbeForwardCli pins the probe verb's client half (plan 09-02):
// exactly one positional keyspec (anything else is the bare errUsage — the
// message-identity comparison keeps the unknown-subcommand wrapper out),
// and a valid spec dials ProbeForward through the seam — the dial itself is
// swapped out so the corpus never touches a live bus.
func TestRun_ProbeForwardCli(t *testing.T) {
	t.Run("no argument is the usage refusal", func(t *testing.T) {
		err := run(context.Background(), []string{cliProbeForward})
		if err == nil || err.Error() != errUsage.Error() {
			t.Fatalf("probe-forward with no argument = %v, want the bare errUsage", err)
		}
		if !strings.Contains(err.Error(), "probe-forward <keyspec>") {
			t.Errorf("errUsage %q missing the probe-forward segment", err.Error())
		}
	})
	t.Run("two arguments are the usage refusal", func(t *testing.T) {
		err := run(context.Background(), []string{cliProbeForward, "x", "y"})
		if err == nil || err.Error() != errUsage.Error() {
			t.Fatalf("probe-forward with two arguments = %v, want the bare errUsage", err)
		}
	})
	t.Run("valid spec dials ProbeForward", func(t *testing.T) {
		var dialedMethod string
		var dialedArgs []any
		prev := probeForwardCall
		probeForwardCall = func(_ context.Context, method string, args ...any) (string, error) {
			dialedMethod = method
			dialedArgs = args

			return "probe forwarded", nil
		}
		t.Cleanup(func() { probeForwardCall = prev })

		if err := run(context.Background(), []string{cliProbeForward, "shift+Home"}); err != nil {
			t.Fatalf("probe-forward shift+Home error = %v, want nil", err)
		}
		if dialedMethod != "ProbeForward" {
			t.Errorf("seam dialed method %q, want ProbeForward", dialedMethod)
		}
		if len(dialedArgs) != 1 || dialedArgs[0] != "shift+Home" {
			t.Errorf("seam dialed args %#v, want exactly [shift+Home]", dialedArgs)
		}
	})
	t.Run("dial error names the subcommand", func(t *testing.T) {
		prev := probeForwardCall
		probeForwardCall = func(context.Context, string, ...any) (string, error) {
			return "", errStubProbeDisabled
		}
		t.Cleanup(func() { probeForwardCall = prev })

		err := run(context.Background(), []string{cliProbeForward, "x"})
		if err == nil || !strings.Contains(err.Error(), cliProbeForward) {
			t.Fatalf("probe-forward dial error = %v, want a wrap naming the subcommand", err)
		}
	})
}

// The delivery-probe verb names (plan 10-13) — the goconst form of
// cliProbeForward.
const (
	cliProbeDeleteSurrounding = "probe-delete-surrounding"
	cliProbeCommit            = "probe-commit"
)

// TestRun_ProbeDeliveryCliUsage pins the delivery probes' usage grammar
// (plan 10-13): the exact positional shapes (anything else is the bare
// errUsage — the message-identity comparison keeps the unknown-subcommand
// wrapper out) and the usage segments.
func TestRun_ProbeDeliveryCliUsage(t *testing.T) {
	t.Run("no arguments are the usage refusal", func(t *testing.T) {
		for _, verb := range []string{cliProbeDeleteSurrounding, cliProbeCommit} {
			err := run(context.Background(), []string{verb})
			if err == nil || err.Error() != errUsage.Error() {
				t.Fatalf("%s with no argument = %v, want the bare errUsage", verb, err)
			}
		}
		if seg := cliProbeDeleteSurrounding + " <offset> <nchars>"; !strings.Contains(errUsage.Error(), seg) {
			t.Errorf("errUsage %q missing the %q segment", errUsage.Error(), seg)
		}
		if seg := cliProbeCommit + " <text>"; !strings.Contains(errUsage.Error(), seg) {
			t.Errorf("errUsage %q missing the %q segment", errUsage.Error(), seg)
		}
	})
	t.Run("wrong argument counts are the usage refusal", func(t *testing.T) {
		for _, argv := range [][]string{
			{cliProbeDeleteSurrounding, "0"},
			{cliProbeDeleteSurrounding, "0", "6", "7"},
			{cliProbeCommit, "a", "b"},
		} {
			err := run(context.Background(), argv)
			if err == nil || err.Error() != errUsage.Error() {
				t.Fatalf("%v = %v, want the bare errUsage", argv, err)
			}
		}
	})
}

// TestRun_ProbeDeliveryCliDials pins the dials of both delivery probes
// (plan 10-13): the seam carries TYPED wire arguments (int32/uint32 — the
// D-Bus (i u) signature), and every refusal names the subcommand.
func TestRun_ProbeDeliveryCliDials(t *testing.T) {
	t.Run("valid arguments dial the probe methods", func(t *testing.T) {
		var dialedMethod string
		var dialedArgs []any
		prev := probeForwardCall
		probeForwardCall = func(_ context.Context, method string, args ...any) (string, error) {
			dialedMethod = method
			dialedArgs = args

			return "probe emitted", nil
		}
		t.Cleanup(func() { probeForwardCall = prev })

		if err := run(context.Background(), []string{cliProbeDeleteSurrounding, "-3", "6"}); err != nil {
			t.Fatalf("probe-delete-surrounding -3 6 error = %v, want nil", err)
		}
		if dialedMethod != "ProbeDeleteSurrounding" {
			t.Errorf("seam dialed method %q, want ProbeDeleteSurrounding", dialedMethod)
		}
		if len(dialedArgs) != 2 || dialedArgs[0] != int32(-3) || dialedArgs[1] != uint32(6) {
			t.Errorf("seam dialed args %#v, want [int32(-3) uint32(6)]", dialedArgs)
		}

		if err := run(context.Background(), []string{cliProbeCommit, "привет"}); err != nil {
			t.Fatalf("probe-commit привет error = %v, want nil", err)
		}
		if dialedMethod != "ProbeCommit" {
			t.Errorf("seam dialed method %q, want ProbeCommit", dialedMethod)
		}
		if len(dialedArgs) != 1 || dialedArgs[0] != "привет" {
			t.Errorf("seam dialed args %#v, want exactly [привет]", dialedArgs)
		}
	})
}

// TestRun_ProbeDeliveryCliRefusals pins the refusal corpus of the delivery
// probes (plan 10-13): every parse/arity/dial failure names the subcommand
// (ASVS V5 — every cause distinct).
func TestRun_ProbeDeliveryCliRefusals(t *testing.T) {
	t.Run("non-integer offset names the subcommand", func(t *testing.T) {
		err := run(context.Background(), []string{cliProbeDeleteSurrounding, "six", "6"})
		if err == nil || !strings.Contains(err.Error(), cliProbeDeleteSurrounding) {
			t.Fatalf("probe-delete-surrounding six 6 = %v, want a parse refusal naming the verb", err)
		}
	})
	t.Run("negative nchars names the subcommand", func(t *testing.T) {
		err := run(context.Background(), []string{cliProbeDeleteSurrounding, "0", "-6"})
		if err == nil || !strings.Contains(err.Error(), cliProbeDeleteSurrounding) {
			t.Fatalf("probe-delete-surrounding 0 -6 = %v, want a refusal naming the verb", err)
		}
	})
	t.Run("dial error names the subcommand", func(t *testing.T) {
		prev := probeForwardCall
		probeForwardCall = func(context.Context, string, ...any) (string, error) {
			return "", errStubProbeDisabled
		}
		t.Cleanup(func() { probeForwardCall = prev })

		err := run(context.Background(), []string{cliProbeCommit, "x"})
		if err == nil || !strings.Contains(err.Error(), cliProbeCommit) {
			t.Fatalf("probe-commit dial error = %v, want a wrap naming the subcommand", err)
		}
	})
}

// The split-emission probe verb names (the 2026-10-16 crash bisection) —
// the goconst form of the delivery verbs.
const (
	cliProbeDelete  = "probe-delete"
	cliProbeRequire = "probe-require"
)

// TestRun_ProbeSplitCliUsage pins the split probes' usage grammar: the
// exact positional shapes and the usage segments.
func TestRun_ProbeSplitCliUsage(t *testing.T) {
	t.Run("missing positionals are the usage refusal", func(t *testing.T) {
		err := run(context.Background(), []string{cliProbeDelete})
		if err == nil || err.Error() != errUsage.Error() {
			t.Fatalf("%s with no argument = %v, want the bare errUsage", cliProbeDelete, err)
		}
		if seg := cliProbeDelete + " <offset> <nchars>"; !strings.Contains(errUsage.Error(), seg) {
			t.Errorf("errUsage %q missing the %q segment", errUsage.Error(), seg)
		}
		if seg := cliProbeRequire; !strings.Contains(errUsage.Error(), seg) {
			t.Errorf("errUsage %q missing the %q segment", errUsage.Error(), seg)
		}
	})
	t.Run("wrong argument counts are the usage refusal", func(t *testing.T) {
		for _, argv := range [][]string{
			{cliProbeDelete, "0"},
			{cliProbeDelete, "0", "6", "7"},
			{cliProbeRequire, "0", "0"},
		} {
			err := run(context.Background(), argv)
			if err == nil || err.Error() != errUsage.Error() {
				t.Fatalf("%v = %v, want the bare errUsage", argv, err)
			}
		}
	})
}

// TestRun_ProbeSplitCliDials pins the dials of the split probes: the
// DST-only verb carries TYPED (i u) arguments, the Require-only verb dials
// bare, and every refusal names the subcommand.
func TestRun_ProbeSplitCliDials(t *testing.T) {
	var dialedMethod string
	var dialedArgs []any
	prev := probeForwardCall
	probeForwardCall = func(_ context.Context, method string, args ...any) (string, error) {
		dialedMethod = method
		dialedArgs = args

		return "probe emitted", nil
	}
	t.Cleanup(func() { probeForwardCall = prev })

	if err := run(context.Background(), []string{cliProbeDelete, "-3", "6"}); err != nil {
		t.Fatalf("probe-delete -3 6 error = %v, want nil", err)
	}
	if dialedMethod != "ProbeDelete" {
		t.Errorf("seam dialed method %q, want ProbeDelete", dialedMethod)
	}
	if len(dialedArgs) != 2 || dialedArgs[0] != int32(-3) || dialedArgs[1] != uint32(6) {
		t.Errorf("seam dialed args %#v, want [int32(-3) uint32(6)]", dialedArgs)
	}

	if err := run(context.Background(), []string{cliProbeRequire}); err != nil {
		t.Fatalf("probe-require error = %v, want nil", err)
	}
	if dialedMethod != "ProbeRequire" {
		t.Errorf("seam dialed method %q, want ProbeRequire", dialedMethod)
	}
	if len(dialedArgs) != 0 {
		t.Errorf("seam dialed args %#v, want none", dialedArgs)
	}

	probeForwardCall = func(context.Context, string, ...any) (string, error) {
		return "", errStubProbeDisabled
	}
	for _, argv := range [][]string{{cliProbeDelete, "0", "6"}, {cliProbeRequire}} {
		err := run(context.Background(), argv)
		if err == nil || !strings.Contains(err.Error(), argv[0]) {
			t.Fatalf("%v dial error = %v, want a wrap naming the subcommand", argv, err)
		}
	}
}
