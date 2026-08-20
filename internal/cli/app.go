package cli

import (
	"errors"
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
  uuid              generate cryptographically secure UUIDs
  secret            generate cryptographically secure secrets
  hash              hash a file or standard input
  hash verify       verify a file or standard input checksum
  jwt inspect       decode JWT header and claims without verification
  json pretty       format JSON for readability
  json minify       remove insignificant JSON whitespace
  env diff          compare dotenv key sets without exposing values
  file inspect      inspect a regular file's type, size, extension, and SHA-256
  repo inspect      detect repository metadata without reading file contents
  text inspect      inspect encoding, BOM, line endings, and line counts
  port inspect      inspect local TCP port availability
  timestamp convert convert explicit timestamp formats to UTC
  base64 encode     encode file or standard input as Base64
  base64 decode     decode Base64 to original bytes
  help              show help
  version           show the DevKit version
`

// Run executes the CLI and returns the process exit code.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer, version string) int {
	jsonMode, remaining, parseErr := parseGlobalFlags(args)
	if parseErr != nil {
		return writeFailure(stdout, stderr, jsonMode, "global", "invalid_usage", parseErr.Error(), ExitUsage)
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
	case "secret":
		return runSecret(remaining[1:], stdout, stderr, jsonMode)
	case "hash":
		return runHash(remaining[1:], stdin, stdout, stderr, jsonMode)
	case "jwt":
		return runJWT(remaining[1:], stdin, stdout, stderr, jsonMode)
	case "json":
		return runJSON(remaining[1:], stdin, stdout, stderr, jsonMode)
	case "env":
		return runEnv(remaining[1:], stdout, stderr, jsonMode)
	case "file":
		return runFile(remaining[1:], stdout, stderr, jsonMode)
	case "repo":
		return runRepo(remaining[1:], stdout, stderr, jsonMode)
	case "text":
		return runText(remaining[1:], stdout, stderr, jsonMode)
	case "port":
		return runPort(remaining[1:], stdout, stderr, jsonMode)
	case "timestamp":
		return runTimestamp(remaining[1:], stdout, stderr, jsonMode)
	case "base64":
		return runBase64(remaining[1:], stdin, stdout, stderr, jsonMode)
	default:
		message := fmt.Sprintf("unknown command %q", command)
		return writeFailure(stdout, stderr, jsonMode, command, "unknown_command", message, ExitUsage)
	}
}

func parseGlobalFlags(args []string) (jsonMode bool, remaining []string, err error) {
	action := ""
	for index, arg := range args {
		switch arg {
		case "--json":
			jsonMode = true
		case "--help":
			if action == "version" {
				return jsonMode, nil, errors.New("global flags --help and --version cannot be combined")
			}
			action = "help"
		case "--version":
			if action == "help" {
				return jsonMode, nil, errors.New("global flags --help and --version cannot be combined")
			}
			action = "version"
		default:
			if strings.HasPrefix(arg, "-") {
				return jsonMode, nil, fmt.Errorf("unknown global flag %q", arg)
			}
			if action != "" {
				return jsonMode, []string{action}, nil
			}
			return jsonMode, args[index:], nil
		}
	}

	if action != "" {
		return jsonMode, []string{action}, nil
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
