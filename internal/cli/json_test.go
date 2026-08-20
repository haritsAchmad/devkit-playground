package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/haritsAchmad/devkit-playground/internal/tools/jsonutil"
)

func TestRunJSONPrettyReadsStdin(t *testing.T) {
	stdout, stderr, exitCode := runForTestWithInput(t, `{"name":"Ada","items":[1,2]}`, "json", "pretty")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if !strings.Contains(stdout, "\n  \"name\": \"Ada\"") {
		t.Errorf("stdout = %q, want indented JSON", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRunJSONMinifyReadsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.json")
	input := "{\n  \"name\": \"Ada\",\n  \"active\": true\n}"
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stdout, stderr, exitCode := runForTest(t, "json", "minify", path)
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if stdout != `{"name":"Ada","active":true}`+"\n" {
		t.Errorf("stdout = %q, want minified JSON", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRunJSONModePreservesLargeNumber(t *testing.T) {
	stdout, stderr, exitCode := runForTestWithInput(t, `{"value":9007199254740993}`, "--json", "json", "pretty")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, `"value":9007199254740993`) {
		t.Errorf("stdout = %q, want exact large number", stdout)
	}

	var envelope struct {
		Command string `json:"command"`
		OK      bool   `json:"ok"`
		Data    struct {
			Value map[string]json.Number `json:"value"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(strings.NewReader(stdout))
	decoder.UseNumber()
	if err := decoder.Decode(&envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "json pretty" || !envelope.OK || envelope.Data.Value["value"].String() != "9007199254740993" {
		t.Errorf("envelope = %+v, want precise JSON value", envelope)
	}
}

func TestRunJSONRejectsTrailingData(t *testing.T) {
	stdout, stderr, exitCode := runForTestWithInput(t, `{} {}`, "--json", "json", "minify")

	if exitCode != ExitData {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitData)
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
	if envelope.OK || envelope.Error.Code != "invalid_json" {
		t.Errorf("envelope = %+v, want invalid_json", envelope)
	}
}

func TestRunJSONMissingFileDoesNotExposePath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "private-name.json")
	stdout, stderr, exitCode := runForTest(t, "--json", "json", "pretty", missing)

	if exitCode != ExitOperation {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitOperation)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if strings.Contains(stdout, missing) {
		t.Error("JSON error exposes input file path")
	}
}

func TestRunJSONHelp(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "json", "pretty", "--help")
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitSuccess)
	}
	if !strings.Contains(stdout, "json minify [file]") {
		t.Errorf("stdout = %q, want JSON usage", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRunJSONRejectsOversizedInput(t *testing.T) {
	stdout, stderr, exitCode := runForTestWithInput(t, strings.Repeat(" ", jsonutil.MaxInputSize+1), "--json", "json", "pretty")
	if exitCode != ExitData || stderr != "" {
		t.Fatalf("stdout/stderr/code = %q/%q/%d, want JSON data error", stdout, stderr, exitCode)
	}
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Error.Code != "input_too_large" {
		t.Errorf("error code = %q, want input_too_large", envelope.Error.Code)
	}
}
