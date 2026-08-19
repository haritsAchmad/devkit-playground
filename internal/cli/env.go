package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/haritsAchmad/devkit-playground/internal/output"
	"github.com/haritsAchmad/devkit-playground/internal/tools/envdiff"
)

const envUsage = `Usage:
  devkit [--json] env diff <reference-file> <target-file>

Arguments:
  reference-file    expected environment keys, typically .env.example
  target-file       environment keys to compare, typically .env

Notes:
  Only key names are compared. Values are never included in output.
  Blank lines, comments, and an optional export prefix are supported.
  Each assignment must be contained on one line.
`

type envCounts struct {
	Missing int `json:"missing"`
	Extra   int `json:"extra"`
}

type envData struct {
	Missing []string  `json:"missing"`
	Extra   []string  `json:"extra"`
	Counts  envCounts `json:"counts"`
}

func runEnv(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	if len(args) == 0 {
		return writeFailure(stdout, stderr, jsonMode, "env", "invalid_usage", "env requires the diff subcommand", ExitUsage)
	}
	if args[0] == "--help" {
		return writeEnvHelp(stdout, jsonMode)
	}
	if args[0] != "diff" {
		return writeFailure(stdout, stderr, jsonMode, "env", "unknown_subcommand", "unknown env subcommand", ExitUsage)
	}
	return runEnvDiff(args[1:], stdout, stderr, jsonMode)
}

func runEnvDiff(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	flags := flag.NewFlagSet("env diff", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeEnvHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, "env diff", "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() != 2 {
		return writeFailure(stdout, stderr, jsonMode, "env diff", "invalid_usage", "env diff requires reference and target files", ExitUsage)
	}

	reference, err := os.Open(flags.Arg(0))
	if err != nil {
		return writeFailure(stdout, stderr, jsonMode, "env diff", "file_read_failed", "could not open reference environment file", ExitOperation)
	}
	defer reference.Close()

	target, err := os.Open(flags.Arg(1))
	if err != nil {
		return writeFailure(stdout, stderr, jsonMode, "env diff", "file_read_failed", "could not open target environment file", ExitOperation)
	}
	defer target.Close()

	result, err := envdiff.Compare(reference, target)
	if err != nil {
		switch {
		case errors.Is(err, envdiff.ErrDuplicateKey):
			return writeFailure(stdout, stderr, jsonMode, "env diff", "duplicate_key", err.Error(), ExitData)
		case errors.Is(err, envdiff.ErrMalformedLine):
			return writeFailure(stdout, stderr, jsonMode, "env diff", "invalid_env", err.Error(), ExitData)
		default:
			return writeFailure(stdout, stderr, jsonMode, "env diff", "input_read_failed", "could not read environment files", ExitOperation)
		}
	}

	data := envData{
		Missing: result.Missing,
		Extra:   result.Extra,
		Counts:  envCounts{Missing: len(result.Missing), Extra: len(result.Extra)},
	}
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "env diff", data); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	return writeEnvHuman(stdout, data)
}

func writeEnvHuman(stdout io.Writer, data envData) int {
	if len(data.Missing) == 0 && len(data.Extra) == 0 {
		if _, err := io.WriteString(stdout, "Environment keys match.\n"); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	if _, err := fmt.Fprintln(stdout, "Missing in target:"); err != nil {
		return ExitInternal
	}
	if len(data.Missing) == 0 {
		if _, err := fmt.Fprintln(stdout, "  (none)"); err != nil {
			return ExitInternal
		}
	} else {
		for _, key := range data.Missing {
			if _, err := fmt.Fprintf(stdout, "  %s\n", key); err != nil {
				return ExitInternal
			}
		}
	}

	if _, err := fmt.Fprintln(stdout, "Extra in target:"); err != nil {
		return ExitInternal
	}
	if len(data.Extra) == 0 {
		if _, err := fmt.Fprintln(stdout, "  (none)"); err != nil {
			return ExitInternal
		}
	} else {
		for _, key := range data.Extra {
			if _, err := fmt.Fprintf(stdout, "  %s\n", key); err != nil {
				return ExitInternal
			}
		}
	}
	return ExitSuccess
}

func writeEnvHelp(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "env diff help", map[string]string{"usage": envUsage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if _, err := io.WriteString(stdout, envUsage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}
