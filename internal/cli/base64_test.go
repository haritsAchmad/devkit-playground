package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBase64EncodeStdin(t *testing.T) {
	stdout, stderr, exitCode := runForTestWithInput(t, "hello", "base64", "encode")
	if exitCode != ExitSuccess || stderr != "" || stdout != "aGVsbG8=\n" {
		t.Errorf("stdout/stderr/code = %q/%q/%d, want encoded hello", stdout, stderr, exitCode)
	}
}

func TestRunBase64DecodeFileAsJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "encoded.txt")
	if err := os.WriteFile(path, []byte("aGVsbG8="), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	stdout, stderr, exitCode := runForTest(t, "--json", "base64", "decode", path)
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	var envelope struct {
		Command string     `json:"command"`
		OK      bool       `json:"ok"`
		Data    base64Data `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "base64 decode" || !envelope.OK || envelope.Data.Value != "hello" || envelope.Data.Representation != "utf-8" || envelope.Data.Source != "file" {
		t.Errorf("envelope = %+v, want decoded UTF-8 file", envelope)
	}
	if strings.Contains(stdout, path) {
		t.Error("JSON output exposes input file path")
	}
}

func TestRunBase64DecodeBinaryJSONUsesCanonicalBase64Transport(t *testing.T) {
	stdout, stderr, exitCode := runForTestWithInput(t, "//4=", "--json", "base64", "decode")
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	var envelope struct {
		Data base64Data `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Data.Value != "//4=" || envelope.Data.Representation != "base64" || envelope.Data.OutputBytes != 2 {
		t.Errorf("data = %+v, want canonical binary transport", envelope.Data)
	}
}

func TestRunBase64URLRawRoundTrip(t *testing.T) {
	encoded, stderr, exitCode := runForTestWithInput(t, string([]byte{0xfb, 0xff}), "base64", "encode", "--variant", "url", "--padding", "raw")
	if exitCode != ExitSuccess || stderr != "" || encoded != "-_8\n" {
		t.Fatalf("encoded/stderr/code = %q/%q/%d, want URL raw Base64", encoded, stderr, exitCode)
	}
	decoded, stderr, exitCode := runForTestWithInput(t, strings.TrimSpace(encoded), "base64", "decode", "--variant", "url", "--padding", "raw")
	if exitCode != ExitSuccess || stderr != "" || decoded != string([]byte{0xfb, 0xff}) {
		t.Errorf("decoded/stderr/code = %q/%q/%d, want original bytes", decoded, stderr, exitCode)
	}
}

func TestRunBase64RejectsInvalidInputWithoutEchoingIt(t *testing.T) {
	secret := "not-base64-private-value"
	stdout, stderr, exitCode := runForTestWithInput(t, secret, "--json", "base64", "decode")
	if exitCode != ExitData || stderr != "" || strings.Contains(stdout, secret) {
		t.Errorf("stdout/stderr/code = %q/%q/%d, want safe invalid Base64 error", stdout, stderr, exitCode)
	}
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Error.Code != "invalid_base64" {
		t.Errorf("error code = %q, want invalid_base64", envelope.Error.Code)
	}
}

func TestRunBase64Help(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "base64", "--help")
	if exitCode != ExitSuccess || stderr != "" || !strings.Contains(stdout, "base64 encode") {
		t.Errorf("stdout/stderr/code = %q/%q/%d, want Base64 help", stdout, stderr, exitCode)
	}
}
