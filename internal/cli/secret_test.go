package cli

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

func TestRunSecretDefaultsToBase64URL(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "secret")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	value := strings.TrimSuffix(stdout, "\n")
	if !regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`).MatchString(value) {
		t.Errorf("secret = %q, want 32-byte unpadded base64url", value)
	}
}

func TestRunSecretHex(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "secret", "--length", "4", "--encoding", "hex")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}\n$`).MatchString(stdout) {
		t.Errorf("stdout = %q, want four-byte hex secret", stdout)
	}
}

func TestRunSecretJSON(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "--json", "secret", "--length", "8")

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
			Secret   string `json:"secret"`
			Encoding string `json:"encoding"`
			Bytes    int    `json:"bytes"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "secret" || !envelope.OK {
		t.Errorf("envelope = %+v, want successful secret result", envelope)
	}
	if envelope.Data.Encoding != "base64url" || envelope.Data.Bytes != 8 || envelope.Data.Secret == "" {
		t.Errorf("data = %+v, want eight-byte base64url secret", envelope.Data)
	}
}

func TestRunSecretRejectsInvalidLength(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "secret", "--length", "0")

	if exitCode != ExitUsage {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitUsage)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "must be between 1 and 4096 bytes") {
		t.Errorf("stderr = %q, want length range", stderr)
	}
}

func TestRunSecretRejectsInvalidEncodingAsJSON(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "--json", "secret", "--encoding", "plain")

	if exitCode != ExitUsage {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitUsage)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	var envelope struct {
		OK    bool `json:"ok"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.OK || envelope.Error.Code != "invalid_encoding" {
		t.Errorf("envelope = %+v, want invalid_encoding failure", envelope)
	}
	if strings.Contains(envelope.Error.Message, "generated-value") {
		t.Error("error unexpectedly contains a secret value")
	}
}

func TestRunSecretHelp(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "secret", "--help")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitSuccess)
	}
	if !strings.Contains(stdout, "--encoding base64url|hex") {
		t.Errorf("stdout = %q, want secret usage", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}
