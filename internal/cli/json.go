package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/haritsAchmad/devkit-playground/internal/output"
	"github.com/haritsAchmad/devkit-playground/internal/tools/jsonutil"
)

const jsonUsage = `Usage:
  devkit [--json] json pretty [file]
  devkit [--json] json minify [file]

Arguments:
  file         JSON file to process; omit or use - to read standard input

Notes:
  Inputs are never modified in place.
  In --json mode, the parsed value is returned in data.value.
`

type jsonData struct {
	Value any `json:"value"`
}

func runJSON(args []string, stdin io.Reader, stdout, stderr io.Writer, jsonMode bool) int {
	if len(args) == 0 {
		return writeFailure(stdout, stderr, jsonMode, "json", "invalid_usage", "json requires the pretty or minify subcommand", ExitUsage)
	}
	if args[0] == "--help" {
		return writeJSONHelp(stdout, jsonMode)
	}
	if args[0] != "pretty" && args[0] != "minify" {
		return writeFailure(stdout, stderr, jsonMode, "json", "unknown_subcommand", "unknown json subcommand", ExitUsage)
	}

	return runJSONFormat(args[0], args[1:], stdin, stdout, stderr, jsonMode)
}

func runJSONFormat(operation string, args []string, stdin io.Reader, stdout, stderr io.Writer, jsonMode bool) int {
	command := "json " + operation
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeJSONHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, command, "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() > 1 {
		return writeFailure(stdout, stderr, jsonMode, command, "invalid_usage", command+" accepts at most one file", ExitUsage)
	}

	input := stdin
	var closeInput func() error
	if flags.NArg() == 1 && flags.Arg(0) != "-" {
		file, err := os.Open(flags.Arg(0))
		if err != nil {
			return writeFailure(stdout, stderr, jsonMode, command, "file_read_failed", "could not open JSON input file", ExitOperation)
		}
		input = file
		closeInput = file.Close
	}
	if closeInput != nil {
		defer closeInput()
	}

	document, err := jsonutil.Parse(input)
	if err != nil {
		if errors.Is(err, jsonutil.ErrInputTooLarge) {
			return writeFailure(stdout, stderr, jsonMode, command, "input_too_large", "JSON input exceeds maximum size of 16 MiB", ExitData)
		}
		if errors.Is(err, jsonutil.ErrInvalidJSON) {
			return writeFailure(stdout, stderr, jsonMode, command, "invalid_json", "input must contain exactly one valid JSON value", ExitData)
		}
		return writeFailure(stdout, stderr, jsonMode, command, "input_read_failed", "could not read JSON input", ExitOperation)
	}

	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, command, jsonData{Value: document.Value}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	formatted, err := formatJSONDocument(document, operation)
	if err != nil {
		return ExitInternal
	}
	if _, err := fmt.Fprintln(stdout, string(formatted)); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}

func formatJSONDocument(document jsonutil.Document, operation string) ([]byte, error) {
	if operation == "pretty" {
		return document.Pretty()
	}
	return document.Minify()
}

func writeJSONHelp(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "json help", map[string]string{"usage": jsonUsage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	if _, err := io.WriteString(stdout, jsonUsage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}
