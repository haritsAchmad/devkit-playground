package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestRunHelpByDefault(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t)

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitSuccess)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("stdout = %q, want usage", stdout)
	}
	for _, command := range []string{"hash verify", "base64 decode", "help", "version"} {
		if !strings.Contains(stdout, command) {
			t.Errorf("stdout = %q, want listed command %q", stdout, command)
		}
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRunVersion(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "--version")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitSuccess)
	}
	if stdout != "devkit test-version\n" {
		t.Errorf("stdout = %q, want version", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRunJSONVersion(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "--json", "--version")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitSuccess)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	var envelope struct {
		Command string `json:"command"`
		OK      bool   `json:"ok"`
		Data    struct {
			Version string `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "version" || !envelope.OK || envelope.Data.Version != "test-version" {
		t.Errorf("envelope = %+v, want successful version result", envelope)
	}
}

func TestRunGlobalJSONFlagOrderDoesNotChangeHelpOrVersionMode(t *testing.T) {
	for _, args := range [][]string{{"--help", "--json"}, {"--version", "--json"}} {
		stdout, stderr, exitCode := runForTest(t, args...)
		if exitCode != ExitSuccess || stderr != "" {
			t.Fatalf("args %v: stdout/stderr/code = %q/%q/%d, want JSON success", args, stdout, stderr, exitCode)
		}
		var envelope struct {
			OK bool `json:"ok"`
		}
		if err := json.Unmarshal([]byte(stdout), &envelope); err != nil || !envelope.OK {
			t.Errorf("args %v: stdout = %q, want successful JSON envelope; error = %v", args, stdout, err)
		}
	}
}

func TestRunRejectsConflictingGlobalActions(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "--json", "--help", "--version")
	if exitCode != ExitUsage || stderr != "" {
		t.Fatalf("stdout/stderr/code = %q/%q/%d, want JSON usage error", stdout, stderr, exitCode)
	}
	var envelope struct {
		Command string `json:"command"`
		OK      bool   `json:"ok"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "global" || envelope.OK || envelope.Error.Code != "invalid_usage" {
		t.Errorf("envelope = %+v, want global invalid_usage", envelope)
	}
}

func TestRunUnknownCommandHuman(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "wat")

	if exitCode != ExitUsage {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitUsage)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if stderr != "devkit: unknown command \"wat\"\n" {
		t.Errorf("stderr = %q, want unknown command error", stderr)
	}
}

func TestRunUnknownCommandJSON(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "--json", "wat")

	if exitCode != ExitUsage {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitUsage)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	var envelope struct {
		Command string `json:"command"`
		OK      bool   `json:"ok"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "wat" || envelope.OK || envelope.Error.Code != "unknown_command" {
		t.Errorf("envelope = %+v, want unknown_command failure", envelope)
	}
}

func TestRunRejectsUnknownGlobalFlag(t *testing.T) {
	_, stderr, exitCode := runForTest(t, "--wat")

	if exitCode != ExitUsage {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitUsage)
	}
	if !strings.Contains(stderr, "unknown global flag") {
		t.Errorf("stderr = %q, want unknown global flag error", stderr)
	}
}

func TestRunUnknownGlobalFlagAsJSONUsesGlobalCommand(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "--json", "--wat")
	if exitCode != ExitUsage || stderr != "" {
		t.Fatalf("stdout/stderr/code = %q/%q/%d, want JSON usage error", stdout, stderr, exitCode)
	}
	var envelope struct {
		Command string `json:"command"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "global" || envelope.Error.Code != "invalid_usage" {
		t.Errorf("envelope = %+v, want global invalid_usage", envelope)
	}
}

func TestRunTreatsGlobalFlagAfterCommandAsCommandArgument(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "uuid", "--json")
	if exitCode != ExitUsage {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitUsage)
	}
	if stdout != "" || !strings.Contains(stderr, "flag provided but not defined") {
		t.Errorf("stdout/stderr = %q/%q, want command-level invalid flag error", stdout, stderr)
	}
}

func TestRunEveryCommandHelpAsJSON(t *testing.T) {
	cases := []struct {
		args    []string
		command string
	}{
		{[]string{"uuid", "--help"}, "uuid help"},
		{[]string{"secret", "--help"}, "secret help"},
		{[]string{"hash", "--help"}, "hash help"},
		{[]string{"hash", "verify", "--help"}, "hash help"},
		{[]string{"jwt", "inspect", "--help"}, "jwt inspect help"},
		{[]string{"json", "pretty", "--help"}, "json help"},
		{[]string{"env", "diff", "--help"}, "env diff help"},
		{[]string{"file", "inspect", "--help"}, "file inspect help"},
		{[]string{"repo", "inspect", "--help"}, "repo inspect help"},
		{[]string{"text", "inspect", "--help"}, "text inspect help"},
		{[]string{"port", "inspect", "--help"}, "port inspect help"},
		{[]string{"timestamp", "convert", "--help"}, "timestamp convert help"},
		{[]string{"base64", "encode", "--help"}, "base64 help"},
		{[]string{"capabilities", "--help"}, "capabilities help"},
	}
	for _, test := range cases {
		t.Run(test.command, func(t *testing.T) {
			args := append([]string{"--json"}, test.args...)
			stdout, stderr, exitCode := runForTest(t, args...)
			if exitCode != ExitSuccess || stderr != "" {
				t.Fatalf("stdout/stderr/code = %q/%q/%d, want JSON help success", stdout, stderr, exitCode)
			}
			var envelope struct {
				Command string `json:"command"`
				OK      bool   `json:"ok"`
				Data    struct {
					Usage string `json:"usage"`
				} `json:"data"`
			}
			decoder := json.NewDecoder(strings.NewReader(stdout))
			if err := decoder.Decode(&envelope); err != nil {
				t.Fatalf("stdout is not JSON: %v", err)
			}
			var extra any
			if err := decoder.Decode(&extra); err != io.EOF {
				t.Fatalf("stdout contains trailing data: %v", err)
			}
			if envelope.Command != test.command || !envelope.OK || envelope.Data.Usage == "" {
				t.Errorf("envelope = %+v, want %s help", envelope, test.command)
			}
		})
	}
}

func runForTest(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	return runForTestWithInput(t, "", args...)
}

func runForTestWithInput(t *testing.T, input string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	exitCode = Run(args, strings.NewReader(input), &stdoutBuffer, &stderrBuffer, "test-version")
	return stdoutBuffer.String(), stderrBuffer.String(), exitCode
}
