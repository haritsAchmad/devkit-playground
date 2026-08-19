package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriteJSONSuccess(t *testing.T) {
	var destination bytes.Buffer

	err := WriteJSONSuccess(&destination, "version", map[string]string{"version": "dev"})
	if err != nil {
		t.Fatalf("WriteJSONSuccess() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(destination.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if got["schema_version"] != SchemaVersion {
		t.Errorf("schema_version = %v, want %q", got["schema_version"], SchemaVersion)
	}
	if got["command"] != "version" {
		t.Errorf("command = %v, want version", got["command"])
	}
	if got["ok"] != true {
		t.Errorf("ok = %v, want true", got["ok"])
	}
	if _, exists := got["error"]; exists {
		t.Error("successful envelope contains error")
	}
	if destination.Bytes()[destination.Len()-1] != '\n' {
		t.Error("JSON output does not end with a newline")
	}
}

func TestWriteJSONError(t *testing.T) {
	var destination bytes.Buffer

	err := WriteJSONError(&destination, "wat", "unknown_command", "unknown command")
	if err != nil {
		t.Fatalf("WriteJSONError() error = %v", err)
	}

	var got struct {
		OK    bool   `json:"ok"`
		Data  any    `json:"data"`
		Error *Error `json:"error"`
	}
	if err := json.Unmarshal(destination.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if got.OK {
		t.Error("ok = true, want false")
	}
	if got.Data != nil {
		t.Errorf("data = %v, want nil", got.Data)
	}
	if got.Error == nil || got.Error.Code != "unknown_command" {
		t.Errorf("error = %+v, want unknown_command", got.Error)
	}
}
