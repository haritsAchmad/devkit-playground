package output

import (
	"encoding/json"
	"io"
)

const SchemaVersion = "1"

// Error describes a stable, machine-readable command failure.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type envelope struct {
	SchemaVersion string `json:"schema_version"`
	Command       string `json:"command"`
	OK            bool   `json:"ok"`
	Data          any    `json:"data,omitempty"`
	Error         *Error `json:"error,omitempty"`
}

// WriteJSONSuccess writes one successful JSON envelope followed by a newline.
func WriteJSONSuccess(w io.Writer, command string, data any) error {
	return json.NewEncoder(w).Encode(envelope{
		SchemaVersion: SchemaVersion,
		Command:       command,
		OK:            true,
		Data:          data,
	})
}

// WriteJSONError writes one failed JSON envelope followed by a newline.
func WriteJSONError(w io.Writer, command, code, message string) error {
	return json.NewEncoder(w).Encode(envelope{
		SchemaVersion: SchemaVersion,
		Command:       command,
		OK:            false,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	})
}
