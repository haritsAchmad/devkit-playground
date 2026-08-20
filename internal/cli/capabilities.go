package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/haritsAchmad/devkit-playground/internal/output"
)

const capabilitiesSchemaVersion = "1"

const capabilitiesUsage = `Usage:
  devkit [--json] capabilities [--category NAME]

Flags:
  --category NAME       compare, generate, inspect, integrity, system, or transform

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
	{Name: "capabilities", Category: "system", Summary: "report structured command capabilities", InputSources: []string{"arguments"}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
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
	{Name: "text inspect", Category: "inspect", Summary: "inspect encoding, line endings, and Unicode anomalies", InputSources: []string{"file"}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "timestamp convert", Category: "transform", Summary: "convert explicit timestamp formats to UTC", InputSources: []string{"arguments"}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "uuid", Category: "generate", Summary: "generate cryptographically secure UUIDs", InputSources: []string{"arguments"}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
	{Name: "version", Category: "system", Summary: "show the DevKit version", InputSources: []string{}, OutputSensitivity: "normal", SideEffects: "none", Offline: true, SupportsJSON: true},
}

var capabilityCategories = map[string]struct{}{
	"compare": {}, "generate": {}, "inspect": {},
	"integrity": {}, "system": {}, "transform": {},
}

func runCapabilities(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	flags := flag.NewFlagSet("capabilities", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	category := flags.String("category", "", "capability category")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeCapabilitiesHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, "capabilities", "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() != 0 {
		return writeFailure(stdout, stderr, jsonMode, "capabilities", "invalid_usage", "capabilities does not accept arguments", ExitUsage)
	}
	commands, validCategory := filterCapabilities(*category)
	if !validCategory {
		return writeFailure(stdout, stderr, jsonMode, "capabilities", "invalid_category", "unknown capability category", ExitUsage)
	}

	data := capabilitiesData{CapabilitiesSchemaVersion: capabilitiesSchemaVersion, Commands: commands}
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "capabilities", data); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if _, err := fmt.Fprintf(stdout, "DevKit capabilities (%d):\n", len(commands)); err != nil {
		return ExitInternal
	}
	grouped := make(map[string][]capability)
	for _, item := range commands {
		grouped[item.Category] = append(grouped[item.Category], item)
	}
	categories := make([]string, 0, len(grouped))
	for name := range grouped {
		categories = append(categories, name)
	}
	sort.Strings(categories)
	for _, name := range categories {
		if _, err := fmt.Fprintf(stdout, "\n%s:\n", strings.ToUpper(name[:1])+name[1:]); err != nil {
			return ExitInternal
		}
		for _, item := range grouped[name] {
			if _, err := fmt.Fprintf(stdout, "  %-18s %s\n", item.Name, item.Summary); err != nil {
				return ExitInternal
			}
		}
	}
	if _, err := io.WriteString(stdout, "\nUse `devkit <command> --help` for details or `devkit --json capabilities` for machine-readable metadata.\n"); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}

func filterCapabilities(category string) ([]capability, bool) {
	if category == "" {
		return commandCapabilities, true
	}
	if _, valid := capabilityCategories[category]; !valid {
		return nil, false
	}
	result := make([]capability, 0)
	for _, item := range commandCapabilities {
		if item.Category == category {
			result = append(result, item)
		}
	}
	return result, true
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
