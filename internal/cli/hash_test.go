package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHashReadsStdin(t *testing.T) {
	stdout, stderr, exitCode := runForTestWithInput(t, "hello", "hash")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	const expected = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824\n"
	if stdout != expected {
		t.Errorf("stdout = %q, want %q", stdout, expected)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRunHashReadsFileAsJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stdout, stderr, exitCode := runForTest(t, "--json", "hash", "--algorithm", "sha512", path)
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	var envelope struct {
		Command string   `json:"command"`
		OK      bool     `json:"ok"`
		Data    hashData `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "hash" || !envelope.OK {
		t.Errorf("envelope = %+v, want successful hash result", envelope)
	}
	if envelope.Data.Algorithm != "sha512" || envelope.Data.Bytes != 5 || envelope.Data.Source != "file" {
		t.Errorf("data = %+v, want sha512 file metadata", envelope.Data)
	}
	if strings.Contains(stdout, path) {
		t.Error("JSON output exposes input file path")
	}
}

func TestRunHashDashReadsStdin(t *testing.T) {
	stdout, stderr, exitCode := runForTestWithInput(t, "hello", "hash", "-")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if !strings.HasPrefix(stdout, "2cf24dba") {
		t.Errorf("stdout = %q, want SHA-256 digest", stdout)
	}
}

func TestRunHashRejectsUnsupportedAlgorithm(t *testing.T) {
	stdout, stderr, exitCode := runForTestWithInput(t, "hello", "hash", "--algorithm", "md5")

	if exitCode != ExitUsage {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitUsage)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "must be \"sha256\" or \"sha512\"") {
		t.Errorf("stderr = %q, want supported algorithm error", stderr)
	}
}

func TestRunHashMissingFileAsJSON(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.txt")
	stdout, stderr, exitCode := runForTest(t, "--json", "hash", missing)

	if exitCode != ExitOperation {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitOperation)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if strings.Contains(stdout, missing) {
		t.Error("JSON error exposes input file path")
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
	if envelope.OK || envelope.Error.Code != "file_read_failed" {
		t.Errorf("envelope = %+v, want file_read_failed", envelope)
	}
}

func TestRunHashHelp(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "hash", "--help")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitSuccess)
	}
	if !strings.Contains(stdout, "hash [--algorithm sha256|sha512] [file]") {
		t.Errorf("stdout = %q, want hash usage", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}
