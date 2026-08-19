package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRunEnvDiffHuman(t *testing.T) {
	reference := writeEnvFixture(t, ".env.example", "DATABASE_URL=example\nAPP_ENV=dev\nCACHE_URL=example\n")
	target := writeEnvFixture(t, ".env", "APP_ENV=prod\nLOCAL_DEBUG=true\n")

	stdout, stderr, exitCode := runForTest(t, "env", "diff", reference, target)
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	const expected = "Missing in target:\n  CACHE_URL\n  DATABASE_URL\nExtra in target:\n  LOCAL_DEBUG\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
	if strings.Contains(stdout, "example") || strings.Contains(stdout, "prod") {
		t.Error("human output exposes environment values")
	}
}

func TestRunEnvDiffJSON(t *testing.T) {
	reference := writeEnvFixture(t, "reference.env", "DATABASE_URL=secret\nAPP_ENV=dev\n")
	target := writeEnvFixture(t, "target.env", "APP_ENV=prod\nLOCAL_DEBUG=true\n")

	stdout, stderr, exitCode := runForTest(t, "--json", "env", "diff", reference, target)
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if strings.Contains(stdout, "secret") || strings.Contains(stdout, "prod") {
		t.Error("JSON output exposes environment values")
	}

	var envelope struct {
		Command string  `json:"command"`
		OK      bool    `json:"ok"`
		Data    envData `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "env diff" || !envelope.OK {
		t.Errorf("envelope = %+v, want successful env diff", envelope)
	}
	if !reflect.DeepEqual(envelope.Data.Missing, []string{"DATABASE_URL"}) || !reflect.DeepEqual(envelope.Data.Extra, []string{"LOCAL_DEBUG"}) {
		t.Errorf("data = %+v, want key differences", envelope.Data)
	}
	if envelope.Data.Counts.Missing != 1 || envelope.Data.Counts.Extra != 1 {
		t.Errorf("counts = %+v, want one missing and one extra", envelope.Data.Counts)
	}
}

func TestRunEnvDiffMatch(t *testing.T) {
	reference := writeEnvFixture(t, "reference.env", "APP_ENV=dev\n")
	target := writeEnvFixture(t, "target.env", "APP_ENV=prod\n")

	stdout, stderr, exitCode := runForTest(t, "env", "diff", reference, target)
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if stdout != "Environment keys match.\n" {
		t.Errorf("stdout = %q, want matching message", stdout)
	}
}

func TestRunEnvDiffRejectsDuplicateWithoutValueLeak(t *testing.T) {
	secretValue := "private-value"
	reference := writeEnvFixture(t, "reference.env", "TOKEN=first\nTOKEN="+secretValue+"\n")
	target := writeEnvFixture(t, "target.env", "TOKEN=target\n")

	stdout, stderr, exitCode := runForTest(t, "--json", "env", "diff", reference, target)
	if exitCode != ExitData {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitData)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if strings.Contains(stdout, secretValue) {
		t.Error("JSON error exposes environment value")
	}

	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Error.Code != "duplicate_key" {
		t.Errorf("error code = %q, want duplicate_key", envelope.Error.Code)
	}
}

func TestRunEnvDiffMissingFileDoesNotExposePath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "private-name.env")
	target := writeEnvFixture(t, "target.env", "APP_ENV=dev\n")
	stdout, stderr, exitCode := runForTest(t, "--json", "env", "diff", missing, target)

	if exitCode != ExitOperation {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitOperation)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if strings.Contains(stdout, missing) {
		t.Error("JSON error exposes environment file path")
	}
}

func TestRunEnvDiffHelp(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "env", "diff", "--help")
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitSuccess)
	}
	if !strings.Contains(stdout, "Values are never included in output") {
		t.Errorf("stdout = %q, want env diff safety note", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func writeEnvFixture(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}
