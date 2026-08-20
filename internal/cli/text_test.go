package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTextInspectHuman(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mixed.txt")
	if err := os.WriteFile(path, []byte("first\r\nsecond\nthird"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	stdout, stderr, exitCode := runForTest(t, "text", "inspect", path)
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	for _, expected := range []string{"Encoding: utf-8", "Newline style: mixed", "Lines: 3", "Final newline: false"} {
		if !strings.Contains(stdout, expected) {
			t.Errorf("stdout = %q, want %q", stdout, expected)
		}
	}
}

func TestRunTextInspectJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "windows.txt")
	content := []byte("first\r\nsecond\r\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	stdout, stderr, exitCode := runForTest(t, "--json", "text", "inspect", path)
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	var envelope struct {
		Command string          `json:"command"`
		OK      bool            `json:"ok"`
		Data    textInspectData `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "text inspect" || !envelope.OK || envelope.Data.Path != path || envelope.Data.Bytes != int64(len(content)) {
		t.Errorf("envelope = %+v, want successful text inspection", envelope)
	}
	if envelope.Data.LineAnalysis == nil || envelope.Data.LineAnalysis.Style != "crlf" || envelope.Data.LineAnalysis.LineCount != 2 || !envelope.Data.LineAnalysis.FinalNewline {
		t.Errorf("line analysis = %+v, want two final CRLF lines", envelope.Data.LineAnalysis)
	}
}

func TestRunTextInspectRejectsDirectoryWithoutExposingPath(t *testing.T) {
	path := t.TempDir()
	stdout, stderr, exitCode := runForTest(t, "--json", "text", "inspect", path)
	if exitCode != ExitData || stderr != "" || strings.Contains(stdout, path) {
		t.Errorf("stdout/stderr/code = %q/%q/%d, want path-safe regular-file error", stdout, stderr, exitCode)
	}
}

func TestRunTextInspectHelp(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "text", "inspect", "--help")
	if exitCode != ExitSuccess || stderr != "" || !strings.Contains(stdout, "text inspect <path>") {
		t.Errorf("stdout/stderr/code = %q/%q/%d, want text inspect help", stdout, stderr, exitCode)
	}
}
