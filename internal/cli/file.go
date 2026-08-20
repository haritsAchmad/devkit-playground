package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/haritsAchmad/devkit-playground/internal/output"
	"github.com/haritsAchmad/devkit-playground/internal/tools/fileinspect"
)

const fileInspectUsage = `Usage:
  devkit [--json] file inspect <path>

Arguments:
  path                  regular file to inspect

The detected MIME type is based on content sniffing. Extension checks are
reported as match, mismatch, or unknown; unknown avoids unsupported guesses.
`

type fileInspectData struct {
	Path           string `json:"path"`
	Name           string `json:"name"`
	Extension      string `json:"extension"`
	SizeBytes      int64  `json:"size_bytes"`
	DetectedMIME   string `json:"detected_mime"`
	ExtensionCheck string `json:"extension_check"`
	SHA256         string `json:"sha256"`
}

func runFile(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	if len(args) == 0 {
		return writeFailure(stdout, stderr, jsonMode, "file", "invalid_usage", "file requires the inspect subcommand", ExitUsage)
	}
	if args[0] == "--help" {
		return writeFileInspectHelp(stdout, jsonMode)
	}
	if args[0] != "inspect" {
		return writeFailure(stdout, stderr, jsonMode, "file", "unknown_subcommand", "unknown file subcommand", ExitUsage)
	}
	return runFileInspect(args[1:], stdout, stderr, jsonMode)
}

func runFileInspect(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	flags := flag.NewFlagSet("file inspect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeFileInspectHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, "file inspect", "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() != 1 {
		return writeFailure(stdout, stderr, jsonMode, "file inspect", "invalid_usage", "file inspect requires exactly one path", ExitUsage)
	}

	path := flags.Arg(0)
	info, err := os.Lstat(path)
	if err != nil {
		return writeFailure(stdout, stderr, jsonMode, "file inspect", "file_read_failed", "could not inspect input file", ExitOperation)
	}
	if !info.Mode().IsRegular() {
		return writeFailure(stdout, stderr, jsonMode, "file inspect", "not_regular_file", "input path is not a regular file", ExitData)
	}
	file, err := os.Open(path)
	if err != nil {
		return writeFailure(stdout, stderr, jsonMode, "file inspect", "file_read_failed", "could not open input file", ExitOperation)
	}
	defer file.Close()

	result, err := fileinspect.Inspect(file, info.Name())
	if err != nil {
		return writeFailure(stdout, stderr, jsonMode, "file inspect", "file_read_failed", "could not read input file", ExitOperation)
	}
	data := fileInspectData{
		Path: path, Name: result.Name, Extension: result.Extension,
		SizeBytes: result.SizeBytes, DetectedMIME: result.DetectedMIME,
		ExtensionCheck: string(result.ExtensionCheck), SHA256: result.SHA256,
	}

	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "file inspect", data); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	if _, err := fmt.Fprintf(stdout, "Path: %s\nName: %s\nExtension: %s\nSize: %d bytes\nDetected MIME: %s\nExtension check: %s\nSHA-256: %s\n",
		data.Path, data.Name, displayNone(data.Extension), data.SizeBytes, data.DetectedMIME, data.ExtensionCheck, data.SHA256); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}

func displayNone(value string) string {
	if value == "" {
		return "(none)"
	}
	return value
}

func writeFileInspectHelp(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "file inspect help", map[string]string{"usage": fileInspectUsage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if _, err := io.WriteString(stdout, fileInspectUsage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}
