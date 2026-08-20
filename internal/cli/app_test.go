package cli

import (
	"bytes"
	"encoding/json"
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

func TestRunTreatsGlobalFlagAfterCommandAsCommandArgument(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "uuid", "--json")
	if exitCode != ExitUsage {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitUsage)
	}
	if stdout != "" || !strings.Contains(stderr, "flag provided but not defined") {
		t.Errorf("stdout/stderr = %q/%q, want command-level invalid flag error", stdout, stderr)
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
