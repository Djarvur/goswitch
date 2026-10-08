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

// TestRun_ProbeForwardCli pins the probe verb's client half (plan 09-02):
// exactly one positional keyspec (anything else is the bare errUsage, the
// identity comparison keeps the unknown-subcommand wrapper out), and a
// valid spec dials ProbeForward through the seam — the dial itself is
// swapped out so the corpus never touches a live bus.
func TestRun_ProbeForwardCli(t *testing.T) {
	t.Run("no argument is the usage refusal", func(t *testing.T) {
		err := run(context.Background(), []string{"probe-forward"})
		if err != errUsage {
			t.Fatalf("probe-forward with no argument = %v, want the bare errUsage", err)
		}
		if !strings.Contains(err.Error(), "probe-forward <keyspec>") {
			t.Errorf("errUsage %q missing the probe-forward segment", err.Error())
		}
	})
	t.Run("two arguments are the usage refusal", func(t *testing.T) {
		err := run(context.Background(), []string{"probe-forward", "x", "y"})
		if err != errUsage {
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

		if err := run(context.Background(), []string{"probe-forward", "shift+Home"}); err != nil {
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
			return "", errors.New("probe disabled")
		}
		t.Cleanup(func() { probeForwardCall = prev })

		err := run(context.Background(), []string{"probe-forward", "x"})
		if err == nil || !strings.Contains(err.Error(), "probe-forward") {
			t.Fatalf("probe-forward dial error = %v, want a wrap naming the subcommand", err)
		}
	})
}
