package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRunTimestampConvertHumanDefaultsToUnix(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "timestamp", "convert", "0")
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if !strings.Contains(stdout, "UTC: 1970-01-01T00:00:00Z") || !strings.Contains(stdout, "Unix milliseconds: 0") {
		t.Errorf("stdout = %q, want Unix epoch representations", stdout)
	}
}

func TestRunTimestampConvertRFC3339AsJSON(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "--json", "timestamp", "convert", "--from", "rfc3339", "2026-08-20T10:30:15.123+07:00")
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	var envelope struct {
		Command string        `json:"command"`
		OK      bool          `json:"ok"`
		Data    timestampData `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "timestamp convert" || !envelope.OK || envelope.Data.InputFormat != "rfc3339" || envelope.Data.UTC != "2026-08-20T03:30:15.123Z" {
		t.Errorf("envelope = %+v, want normalized RFC3339 timestamp", envelope)
	}
	if envelope.Data.SubsecondNanoseconds != 123000000 {
		t.Errorf("subsecond nanoseconds = %d, want 123000000", envelope.Data.SubsecondNanoseconds)
	}
}

func TestRunTimestampConvertMapsErrors(t *testing.T) {
	cases := []struct {
		args     []string
		exitCode int
		code     string
	}{
		{args: []string{"timestamp", "convert", "--from", "auto", "0"}, exitCode: ExitUsage, code: "unsupported_format"},
		{args: []string{"timestamp", "convert", "wat"}, exitCode: ExitData, code: "invalid_timestamp"},
	}
	for _, test := range cases {
		stdout, stderr, exitCode := runForTest(t, append([]string{"--json"}, test.args...)...)
		if exitCode != test.exitCode || stderr != "" {
			t.Errorf("args %v: stderr/code = %q/%d, want %d", test.args, stderr, exitCode, test.exitCode)
		}
		var envelope struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
			t.Fatalf("stdout is not JSON: %v", err)
		}
		if envelope.Error.Code != test.code {
			t.Errorf("error code = %q, want %q", envelope.Error.Code, test.code)
		}
	}
}

func TestRunTimestampHelp(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "timestamp", "convert", "--help")
	if exitCode != ExitSuccess || stderr != "" || !strings.Contains(stdout, "timestamp convert [--from unix|unix-ms|rfc3339] <value>") {
		t.Errorf("stdout/stderr/code = %q/%q/%d, want timestamp help", stdout, stderr, exitCode)
	}
}
