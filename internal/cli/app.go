package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/haritsAchmad/devkit-playground/internal/output"
)

const (
	ExitSuccess   = 0
	ExitUsage     = 2
	ExitData      = 3
	ExitOperation = 4
	ExitInternal  = 70
)

const usage = `DevKit is an offline-first developer toolbox.

Usage:
  devkit [global flags] <command> [arguments]

Global flags:
  --json       emit structured JSON
  --help       show help
  --version    show the DevKit version

Commands:
  uuid         generate cryptographically secure UUIDs
`

// Run executes the CLI and returns the process exit code.
func Run(args []string, stdout, stderr io.Writer, version string) int {
	jsonMode, remaining, parseErr := parseGlobalFlags(args)
	if parseErr != nil {
		return writeFailure(stdout, stderr, jsonMode, "", "invalid_usage", parseErr.Error(), ExitUsage)
	}

	if len(remaining) == 0 {
		return writeHelp(stdout, jsonMode)
	}

	command := remaining[0]
	switch command {
	case "help":
		if len(remaining) != 1 {
			return writeFailure(stdout, stderr, jsonMode, command, "invalid_usage", "help does not accept arguments yet", ExitUsage)
		}
		return writeHelp(stdout, jsonMode)
	case "version":
		if len(remaining) != 1 {
			return writeFailure(stdout, stderr, jsonMode, command, "invalid_usage", "version does not accept arguments", ExitUsage)
		}
		return writeVersion(stdout, jsonMode, version)
	case "uuid":
		return runUUID(remaining[1:], stdout, stderr, jsonMode)
	default:
		message := fmt.Sprintf("unknown command %q", command)
		return writeFailure(stdout, stderr, jsonMode, command, "unknown_command", message, ExitUsage)
	}
}

func parseGlobalFlags(args []string) (jsonMode bool, remaining []string, err error) {
	for index, arg := range args {
		switch arg {
		case "--json":
			jsonMode = true
		case "--help":
			return jsonMode, []string{"help"}, nil
		case "--version":
			return jsonMode, []string{"version"}, nil
		default:
			if strings.HasPrefix(arg, "-") {
				return jsonMode, nil, fmt.Errorf("unknown global flag %q", arg)
			}
			return jsonMode, args[index:], nil
		}
	}

	return jsonMode, nil, nil
}

func writeHelp(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "help", map[string]string{"usage": usage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	if _, err := io.WriteString(stdout, usage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}

func writeVersion(stdout io.Writer, jsonMode bool, version string) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "version", map[string]string{"version": version}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	if _, err := fmt.Fprintf(stdout, "devkit %s\n", version); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}

func writeFailure(stdout, stderr io.Writer, jsonMode bool, command, code, message string, exitCode int) int {
	if jsonMode {
		if err := output.WriteJSONError(stdout, command, code, message); err != nil {
			return ExitInternal
		}
		return exitCode
	}

	if _, err := fmt.Fprintf(stderr, "devkit: %s\n", message); err != nil {
		return ExitInternal
	}
	return exitCode
}
