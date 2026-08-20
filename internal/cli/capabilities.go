package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/haritsAchmad/devkit-playground/internal/output"
)

const capabilitiesSchemaVersion = "1"

const capabilitiesUsage = `Usage:
  devkit [--json] capabilities

Reports the stable names and operational traits of DevKit commands. The JSON
form is intended for programs and agents that should not parse human help text.
`

type capability struct {
	Name              string   `json:"name"`
	Category          string   `json:"category"`
	Summary           string   `json:"summary"`
	InputSources      []string `json:"input_sources"`
	OutputSensitivity string   `json:"output_sensitivity"`
	SideEffects       string   `json:"side_effects"`
	Offline           bool     `json:"offline"`
	SupportsJSON      bool     `json:"supports_json"`
}

type capabilitiesData struct {
	CapabilitiesSchemaVersion string       `json:"capabilities_schema_version"`
	Commands                  []capability `json:"commands"`
}

var commandCapabilities = []capability{
	{Name: "base64 decode", Category: "transform", Summary: "decode Base64 to original bytes", InputSources: []string{"file", "stdin"}, OutputSensitivity: "requested_content", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "base64 encode", Category: "transform", Summary: "encode file or standard input as Base64", InputSources: []string{"file", "stdin"}, OutputSensitivity: "requested_content", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "capabilities", Category: "system", Summary: "report structured command capabilities", InputSources: []string{}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "env diff", Category: "compare", Summary: "compare dotenv key sets without exposing values", InputSources: []string{"file"}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "file inspect", Category: "inspect", Summary: "inspect file type, size, extension, and SHA-256", InputSources: []string{"file"}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "hash", Category: "integrity", Summary: "hash a file or standard input", InputSources: []string{"file", "stdin"}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "hash verify", Category: "integrity", Summary: "verify a file or standard input checksum", InputSources: []string{"arguments", "file", "stdin"}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "help", Category: "system", Summary: "show human or JSON help", InputSources: []string{}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "json minify", Category: "transform", Summary: "remove insignificant JSON whitespace", InputSources: []string{"file", "stdin"}, OutputSensitivity: "requested_content", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "json pretty", Category: "transform", Summary: "format JSON for readability", InputSources: []string{"file", "stdin"}, OutputSensitivity: "requested_content", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "jwt inspect", Category: "inspect", Summary: "decode unverified JWT header and claims", InputSources: []string{"arguments", "stdin"}, OutputSensitivity: "sensitive", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "port inspect", Category: "inspect", Summary: "inspect local TCP port availability", InputSources: []string{"arguments"}, OutputSensitivity: "normal", SideEffects: "temporary_local_bind", Offline: true, SupportsJSON: true},
	{Name: "repo inspect", Category: "inspect", Summary: "detect repository metadata without reading contents", InputSources: []string{"directory"}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "secret", Category: "generate", Summary: "generate cryptographically secure secrets", InputSources: []string{"arguments"}, OutputSensitivity: "sensitive", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "text inspect", Category: "inspect", Summary: "inspect encoding, BOM, and line endings", InputSources: []string{"file"}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "timestamp convert", Category: "transform", Summary: "convert explicit timestamp formats to UTC", InputSources: []string{"arguments"}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "uuid", Category: "generate", Summary: "generate cryptographically secure UUIDs", InputSources: []string{"arguments"}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "version", Category: "system", Summary: "show the DevKit version", InputSources: []string{}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
}

func runCapabilities(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	flags := flag.NewFlagSet("capabilities", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeCapabilitiesHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, "capabilities", "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() != 0 {
		return writeFailure(stdout, stderr, jsonMode, "capabilities", "invalid_usage", "capabilities does not accept arguments", ExitUsage)
	}

	data := capabilitiesData{CapabilitiesSchemaVersion: capabilitiesSchemaVersion, Commands: commandCapabilities}
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "capabilities", data); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if _, err := fmt.Fprintln(stdout, "DevKit capabilities:"); err != nil {
		return ExitInternal
	}
	for _, item := range commandCapabilities {
		inputs := "none"
		if len(item.InputSources) > 0 {
			inputs = strings.Join(item.InputSources, ", ")
		}
		if _, err := fmt.Fprintf(stdout, "  %s — %s\n    category: %s; inputs: %s; output: %s; effects: %s\n", item.Name, item.Summary, item.Category, inputs, item.OutputSensitivity, item.SideEffects); err != nil {
			return ExitInternal
		}
	}
	return ExitSuccess
}

func writeCapabilitiesHelp(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "capabilities help", map[string]string{"usage": capabilitiesUsage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if _, err := io.WriteString(stdout, capabilitiesUsage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}
