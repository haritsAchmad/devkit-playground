package cli

import (
	"encoding/json"
	"net"
	"strconv"
	"strings"
	"testing"
)

func TestRunPortInspectJSONReportsInUse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen fixture: %v", err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port

	stdout, stderr, exitCode := runForTest(t, "--json", "port", "inspect", strconv.Itoa(port))
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	var envelope struct {
		Command string          `json:"command"`
		OK      bool            `json:"ok"`
		Data    portInspectData `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "port inspect" || !envelope.OK || envelope.Data.Port != port || envelope.Data.State != "in_use" {
		t.Errorf("envelope = %+v, want occupied port result", envelope)
	}
	if envelope.Data.PID != nil || envelope.Data.Process != nil || envelope.Data.OwnerInspection != "not_supported" {
		t.Errorf("owner data = %+v, want explicit unsupported owner lookup", envelope.Data)
	}
}

func TestRunPortInspectHumanReportsAvailable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve fixture: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatalf("release fixture: %v", err)
	}

	stdout, stderr, exitCode := runForTest(t, "port", "inspect", strconv.Itoa(port))
	if exitCode != ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, ExitSuccess, stderr)
	}
	if !strings.Contains(stdout, "State: available") || !strings.Contains(stdout, "Process owner: unavailable") {
		t.Errorf("stdout = %q, want available state and owner limitation", stdout)
	}
}

func TestRunPortInspectRejectsInvalidInput(t *testing.T) {
	for _, args := range [][]string{
		{"port", "inspect", "not-a-port"},
		{"port", "inspect", "70000"},
		{"port", "inspect", "--host", "localhost", "8080"},
	} {
		stdout, stderr, exitCode := runForTest(t, args...)
		if exitCode != ExitUsage || stdout != "" || stderr == "" {
			t.Errorf("args %v: stdout/stderr/code = %q/%q/%d, want usage error", args, stdout, stderr, exitCode)
		}
	}
}

func TestRunPortInspectHelp(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "port", "inspect", "--help")
	if exitCode != ExitSuccess || stderr != "" || !strings.Contains(stdout, "port inspect [--host IP] <port>") {
		t.Errorf("stdout/stderr/code = %q/%q/%d, want port inspect help", stdout, stderr, exitCode)
	}
}
