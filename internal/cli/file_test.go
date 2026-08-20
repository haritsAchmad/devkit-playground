package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFileInspectHuman(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.7\nfixture"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	stdout, stderr, exitCode := runForTest(t, "file", "inspect", path)
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	for _, expected := range []string{"Detected MIME: application/pdf", "Extension check: match", "SHA-256:"} {
		if !strings.Contains(stdout, expected) {
			t.Errorf("stdout = %q, want %q", stdout, expected)
		}
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRunFileInspectJSONReportsMismatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.pdf.exe")
	content := []byte("%PDF-1.7\nfixture")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	stdout, stderr, exitCode := runForTest(t, "--json", "file", "inspect", path)
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	var envelope struct {
		Command string          `json:"command"`
		OK      bool            `json:"ok"`
		Data    fileInspectData `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "file inspect" || !envelope.OK {
		t.Errorf("envelope = %+v, want successful file inspection", envelope)
	}
	if envelope.Data.Path != path || envelope.Data.Extension != ".exe" || envelope.Data.DetectedMIME != "application/pdf" || envelope.Data.ExtensionCheck != "mismatch" {
		t.Errorf("data = %+v, want PDF/extension mismatch metadata", envelope.Data)
	}
	if envelope.Data.SizeBytes != int64(len(content)) || len(envelope.Data.SHA256) != 64 {
		t.Errorf("data = %+v, want size and SHA-256", envelope.Data)
	}
}

func TestRunFileInspectRejectsDirectoryWithoutExposingPathInError(t *testing.T) {
	path := t.TempDir()
	stdout, stderr, exitCode := runForTest(t, "--json", "file", "inspect", path)
	if exitCode != ExitData {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitData)
	}
	if stderr != "" || strings.Contains(stdout, path) {
		t.Errorf("stdout/stderr = %q/%q, want path-safe JSON error", stdout, stderr)
	}
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Error.Code != "not_regular_file" {
		t.Errorf("error code = %q, want not_regular_file", envelope.Error.Code)
	}
}

func TestRunFileInspectHelp(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "file", "inspect", "--help")
	if exitCode != ExitSuccess || stderr != "" || !strings.Contains(stdout, "file inspect <path>") {
		t.Errorf("stdout/stderr/code = %q/%q/%d, want file inspect help", stdout, stderr, exitCode)
	}
}

func TestRunFileHelp(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "file", "--help")
	if exitCode != ExitSuccess || stderr != "" || !strings.Contains(stdout, "file inspect <path>") {
		t.Errorf("stdout/stderr/code = %q/%q/%d, want file help", stdout, stderr, exitCode)
	}
}
