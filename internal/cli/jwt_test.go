package cli

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	jwtinspect "github.com/haritsAchmad/devkit-playground/internal/tools/jwt"
)

func TestRunJWTInspectReadsStdin(t *testing.T) {
	token := testJWT(`{"alg":"HS256","typ":"JWT"}`, `{"sub":"123","name":"Ada"}`, "private-signature")
	stdout, stderr, exitCode := runForTestWithInput(t, token+"\n", "jwt", "inspect")

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "WARNING: signature not verified") || !strings.Contains(stdout, `"name": "Ada"`) {
		t.Errorf("stdout = %q, want warning and decoded claims", stdout)
	}
	if strings.Contains(stdout, token) || strings.Contains(stdout, "private-signature") {
		t.Error("human output exposes raw token or signature")
	}
}

func TestRunJWTInspectJSON(t *testing.T) {
	token := testJWT(`{"alg":"RS256"}`, `{"sub":"123","exp":1893456000}`, "signature")
	stdout, stderr, exitCode := runForTest(t, "--json", "jwt", "inspect", token)

	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}

	var envelope struct {
		Command string  `json:"command"`
		OK      bool    `json:"ok"`
		Data    jwtData `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "jwt inspect" || !envelope.OK {
		t.Errorf("envelope = %+v, want successful JWT result", envelope)
	}
	if envelope.Data.Algorithm != "RS256" || envelope.Data.SignatureVerified {
		t.Errorf("data = %+v, want unverified RS256 metadata", envelope.Data)
	}
	encodedSignature := base64.RawURLEncoding.EncodeToString([]byte("signature"))
	if strings.Contains(stdout, token) || strings.Contains(stdout, encodedSignature) {
		t.Error("JSON output exposes raw token or signature")
	}
}

func TestRunJWTInspectRejectsMalformedTokenWithoutEcho(t *testing.T) {
	rawToken := "sensitive-malformed-token"
	stdout, stderr, exitCode := runForTest(t, "--json", "jwt", "inspect", rawToken)

	if exitCode != ExitData {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitData)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
	if strings.Contains(stdout, rawToken) {
		t.Error("JSON error exposes raw token")
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
	if envelope.OK || envelope.Error.Code != "invalid_token" {
		t.Errorf("envelope = %+v, want invalid_token", envelope)
	}
}

func TestRunJWTRejectsUnknownSubcommand(t *testing.T) {
	_, stderr, exitCode := runForTest(t, "jwt", "verify")
	if exitCode != ExitUsage {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitUsage)
	}
	if !strings.Contains(stderr, "unknown jwt subcommand") {
		t.Errorf("stderr = %q, want unknown subcommand", stderr)
	}
}

func TestRunJWTInspectHelp(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "jwt", "inspect", "--help")
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d", exitCode, ExitSuccess)
	}
	if !strings.Contains(stdout, "NEVER verifies the signature") {
		t.Errorf("stdout = %q, want verification warning", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestRunJWTRejectsOversizedRawStdinBeforeTrimming(t *testing.T) {
	input := testJWT(`{"alg":"HS256"}`, `{}`, "signature") + strings.Repeat(" ", jwtinspect.MaxTokenSize)
	stdout, stderr, exitCode := runForTestWithInput(t, input, "--json", "jwt", "inspect")
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
	if envelope.Error.Code != "invalid_token" || !strings.Contains(stdout, "exceeds maximum size") {
		t.Errorf("stdout = %q, want oversized invalid_token", stdout)
	}
}

func testJWT(header, claims, signature string) string {
	encode := base64.RawURLEncoding.EncodeToString
	return encode([]byte(header)) + "." + encode([]byte(claims)) + "." + encode([]byte(signature))
}
