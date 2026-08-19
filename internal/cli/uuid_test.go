package cli

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

func TestRunUUIDHuman(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "uuid", "--count", "2")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("UUID count = %d, want 2; stdout = %q", len(lines), stdout)
	}
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	for _, line := range lines {
		if !pattern.MatchString(line) {
			t.Errorf("UUID = %q, want UUID v4", line)
		}
	}
}

func TestRunUUIDJSON(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "--json", "uuid", "--count", "2")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	var envelope struct {
		Command string `json:"command"`
		OK      bool   `json:"ok"`
		Data    struct {
			UUIDs []string `json:"uuids"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "uuid" || !envelope.OK || len(envelope.Data.UUIDs) != 2 {
		t.Errorf("envelope = %+v, want two UUIDs", envelope)
	}
}

func TestRunUUIDRejectsInvalidCount(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "uuid", "--count", "0")

	if exitCode != ExitUsage {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitUsage)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "must be between 1 and 1000") {
		t.Errorf("stderr = %q, want count range", stderr)
	}
}

func TestRunUUIDRejectsInvalidCountAsJSON(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "--json", "uuid", "--count", "1001")

	if exitCode != ExitUsage {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitUsage)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	var envelope struct {
		OK    bool `json:"ok"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.OK || envelope.Error.Code != "invalid_count" {
		t.Errorf("envelope = %+v, want invalid_count failure", envelope)
	}
}

func TestRunUUIDHelp(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "uuid", "--help")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitSuccess)
	}
	if !strings.Contains(stdout, "devkit [--json] uuid [--count N]") {
		t.Errorf("stdout = %q, want UUID usage", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}
