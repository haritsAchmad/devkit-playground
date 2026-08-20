package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRepoInspectJSON(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{".git", "migrations"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0o700); err != nil {
			t.Fatalf("create directory: %v", err)
		}
	}
	for _, name := range []string{"go.mod", "go.sum", "Dockerfile"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("contents are not inspected"), 0o600); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
	}

	stdout, stderr, exitCode := runForTest(t, "--json", "repo", "inspect", root)
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	var envelope struct {
		Command string          `json:"command"`
		OK      bool            `json:"ok"`
		Data    repoInspectData `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "repo inspect" || !envelope.OK || envelope.Data.Path != root || !envelope.Data.GitRepository {
		t.Errorf("envelope = %+v, want successful repository inspection", envelope)
	}
	if len(envelope.Data.Projects) != 1 || envelope.Data.Projects[0].Ecosystem != "go" || len(envelope.Data.DockerFiles) != 1 || len(envelope.Data.MigrationDirs) != 1 {
		t.Errorf("data = %+v, want detected Go, Docker, and migration metadata", envelope.Data)
	}
	if strings.Contains(stdout, "contents are not inspected") {
		t.Error("JSON output exposes file contents")
	}
}

func TestRunRepoInspectHumanDefaultsToCurrentDirectory(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "repo", "inspect")
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	for _, expected := range []string{"Path: .", "Git repository:", "Projects:"} {
		if !strings.Contains(stdout, expected) {
			t.Errorf("stdout = %q, want %q", stdout, expected)
		}
	}
}

func TestRunRepoInspectRejectsFileWithoutExposingPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private-name")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	stdout, stderr, exitCode := runForTest(t, "--json", "repo", "inspect", path)
	if exitCode != ExitData || stderr != "" || strings.Contains(stdout, path) {
		t.Errorf("stdout/stderr/code = %q/%q/%d, want path-safe not-directory error", stdout, stderr, exitCode)
	}
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Error.Code != "not_directory" {
		t.Errorf("error code = %q, want not_directory", envelope.Error.Code)
	}
}

func TestRunRepoInspectHelp(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "repo", "inspect", "--help")
	if exitCode != ExitSuccess || stderr != "" || !strings.Contains(stdout, "repo inspect [path]") {
		t.Errorf("stdout/stderr/code = %q/%q/%d, want repo inspect help", stdout, stderr, exitCode)
	}
}
