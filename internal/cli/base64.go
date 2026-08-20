package cli

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"github.com/haritsAchmad/devkit-playground/internal/output"
	"github.com/haritsAchmad/devkit-playground/internal/tools/base64util"
)

const base64Usage = `Usage:
  devkit [--json] base64 encode [--variant standard|url] [--padding padded|raw] [file]
  devkit [--json] base64 decode [--variant standard|url] [--padding padded|raw] [file]

Arguments:
  file                  input file; omit or use - to read standard input

Flags:
  --variant NAME        standard or url (default standard)
  --padding MODE        padded or raw (default padded)

Input is limited to 16 MiB. Human decode output writes the exact decoded bytes.
`

type base64Data struct {
	Operation      string `json:"operation"`
	Variant        string `json:"variant"`
	Padding        string `json:"padding"`
	InputBytes     int    `json:"input_bytes"`
	OutputBytes    int    `json:"output_bytes"`
	Value          string `json:"value"`
	Representation string `json:"representation"`
	Source         string `json:"source"`
}

func runBase64(args []string, stdin io.Reader, stdout, stderr io.Writer, jsonMode bool) int {
	if len(args) == 0 {
		return writeFailure(stdout, stderr, jsonMode, "base64", "invalid_usage", "base64 requires encode or decode", ExitUsage)
	}
	if args[0] == "--help" {
		return writeBase64Help(stdout, jsonMode)
	}
	operation := base64util.Operation(args[0])
	if operation != base64util.OperationEncode && operation != base64util.OperationDecode {
		return writeFailure(stdout, stderr, jsonMode, "base64", "unknown_subcommand", "unknown base64 subcommand", ExitUsage)
	}
	return runBase64Transform(args[1:], stdin, stdout, stderr, jsonMode, operation)
}

func runBase64Transform(args []string, stdin io.Reader, stdout, stderr io.Writer, jsonMode bool, operation base64util.Operation) int {
	command := "base64 " + string(operation)
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	variant := flags.String("variant", string(base64util.VariantStandard), "Base64 alphabet")
	padding := flags.String("padding", string(base64util.PaddingPadded), "Base64 padding mode")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeBase64Help(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, command, "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() > 1 {
		return writeFailure(stdout, stderr, jsonMode, command, "invalid_usage", command+" accepts at most one file", ExitUsage)
	}

	input := stdin
	source := "stdin"
	var file *os.File
	if flags.NArg() == 1 && flags.Arg(0) != "-" {
		var err error
		file, err = os.Open(flags.Arg(0))
		if err != nil {
			return writeFailure(stdout, stderr, jsonMode, command, "file_read_failed", "could not open input file", ExitOperation)
		}
		defer file.Close()
		input = file
		source = "file"
	}

	result, err := base64util.Transform(input, operation, base64util.Variant(*variant), base64util.Padding(*padding))
	if err != nil {
		switch {
		case errors.Is(err, base64util.ErrUnsupportedVariant):
			return writeFailure(stdout, stderr, jsonMode, command, "unsupported_variant", err.Error(), ExitUsage)
		case errors.Is(err, base64util.ErrUnsupportedPadding):
			return writeFailure(stdout, stderr, jsonMode, command, "unsupported_padding", err.Error(), ExitUsage)
		case errors.Is(err, base64util.ErrInputTooLarge):
			return writeFailure(stdout, stderr, jsonMode, command, "input_too_large", "Base64 input exceeds maximum size of 16 MiB", ExitData)
		case errors.Is(err, base64util.ErrInvalidEncoding):
			return writeFailure(stdout, stderr, jsonMode, command, "invalid_base64", "input is not valid for the selected Base64 variant and padding", ExitData)
		default:
			return writeFailure(stdout, stderr, jsonMode, command, "input_read_failed", "could not read Base64 input", ExitOperation)
		}
	}

	if jsonMode {
		data := base64JSONData(result, source)
		if err := output.WriteJSONSuccess(stdout, command, data); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if operation == base64util.OperationEncode {
		if _, err := fmt.Fprintf(stdout, "%s\n", result.Value); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if _, err := stdout.Write(result.Value); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}

func base64JSONData(result base64util.Result, source string) base64Data {
	value := string(result.Value)
	representation := "utf-8"
	if result.Operation == base64util.OperationDecode && !utf8.Valid(result.Value) {
		value = base64.StdEncoding.EncodeToString(result.Value)
		representation = "base64"
	}
	return base64Data{
		Operation: string(result.Operation), Variant: string(result.Variant), Padding: string(result.Padding),
		InputBytes: result.InputBytes, OutputBytes: result.OutputBytes,
		Value: value, Representation: representation, Source: source,
	}
}

func writeBase64Help(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "base64 help", map[string]string{"usage": base64Usage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if _, err := io.WriteString(stdout, base64Usage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}
