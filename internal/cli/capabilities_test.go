package cli

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestRunCapabilitiesJSONIsStableAndComplete(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "--json", "capabilities")
	if exitCode != ExitSuccess || stderr != "" {
		t.Fatalf("stdout/stderr/code = %q/%q/%d, want JSON success", stdout, stderr, exitCode)
	}
	var envelope struct {
		Command string           `json:"command"`
		OK      bool             `json:"ok"`
		Data    capabilitiesData `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Command != "capabilities" || !envelope.OK || envelope.Data.CapabilitiesSchemaVersion != "1" {
		t.Errorf("envelope = %+v, want versioned capabilities success", envelope)
	}
	if !reflect.DeepEqual(envelope.Data.Commands, commandCapabilities) {
		t.Error("JSON capabilities differ from the internal manifest")
	}

	names := make([]string, 0, len(envelope.Data.Commands))
	seen := make(map[string]struct{})
	validCategories := map[string]bool{"compare": true, "generate": true, "inspect": true, "integrity": true, "system": true, "transform": true}
	validSources := map[string]bool{"arguments": true, "directory": true, "file": true, "stdin": true}
	validSensitivity := map[string]bool{"normal": true, "requested_content": true, "sensitive": true}
	validEffects := map[string]bool{"none": true, "temporary_local_bind": true}
	for _, item := range envelope.Data.Commands {
		names = append(names, item.Name)
		if _, exists := seen[item.Name]; exists {
			t.Errorf("duplicate capability name %q", item.Name)
		}
		seen[item.Name] = struct{}{}
		if !item.Offline || !item.SupportsJSON || item.InputSources == nil {
			t.Errorf("capability = %+v, want offline, JSON, and explicit input sources", item)
		}
		if !validCategories[item.Category] || !validSensitivity[item.OutputSensitivity] || !validEffects[item.SideEffects] {
			t.Errorf("capability = %+v, contains an unknown metadata value", item)
		}
		for _, source := range item.InputSources {
			if !validSources[source] {
				t.Errorf("capability %q has unknown input source %q", item.Name, source)
			}
		}
		if !strings.Contains(usage, item.Name) {
			t.Errorf("root help does not list capability %q", item.Name)
		}
	}
	sorted := append([]string(nil), names...)
	sort.Strings(sorted)
	if !reflect.DeepEqual(names, sorted) {
		t.Errorf("capability names = %v, want deterministic lexical order", names)
	}
}

func TestRunCapabilitiesHuman(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "capabilities")
	if exitCode != ExitSuccess || stderr != "" {
		t.Fatalf("stdout/stderr/code = %q/%q/%d, want human success", stdout, stderr, exitCode)
	}
	for _, expected := range []string{"DevKit capabilities:", "jwt inspect", "sensitive", "temporary_local_bind"} {
		if !strings.Contains(stdout, expected) {
			t.Errorf("stdout = %q, want %q", stdout, expected)
		}
	}
}

func TestRunCapabilitiesRejectsArguments(t *testing.T) {
	stdout, stderr, exitCode := runForTest(t, "--json", "capabilities", "extra")
	if exitCode != ExitUsage || stderr != "" {
		t.Fatalf("stdout/stderr/code = %q/%q/%d, want JSON usage error", stdout, stderr, exitCode)
	}
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	if envelope.Error.Code != "invalid_usage" {
		t.Errorf("error code = %q, want invalid_usage", envelope.Error.Code)
	}
}
