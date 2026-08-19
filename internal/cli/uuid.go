package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/haritsAchmad/devkit-playground/internal/output"
	"github.com/haritsAchmad/devkit-playground/internal/tools/uuid"
)

const uuidUsage = `Usage:
  devkit [--json] uuid [--count N]

Flags:
  --count N     number of UUIDs to generate (default 1, maximum 1000)
  --help        show help for uuid
`

func runUUID(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	flags := flag.NewFlagSet("uuid", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	count := flags.Int("count", uuid.DefaultCount, "number of UUIDs to generate")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeUUIDHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, "uuid", "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() != 0 {
		message := fmt.Sprintf("uuid does not accept positional arguments: %q", flags.Arg(0))
		return writeFailure(stdout, stderr, jsonMode, "uuid", "invalid_usage", message, ExitUsage)
	}

	values, err := uuid.Generate(*count)
	if err != nil {
		if errors.Is(err, uuid.ErrInvalidCount) {
			return writeFailure(stdout, stderr, jsonMode, "uuid", "invalid_count", err.Error(), ExitUsage)
		}
		return writeFailure(stdout, stderr, jsonMode, "uuid", "secure_random_failed", "could not generate secure UUID", ExitOperation)
	}

	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "uuid", map[string]any{"uuids": values}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	if _, err := fmt.Fprintln(stdout, strings.Join(values, "\n")); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}

func writeUUIDHelp(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "uuid help", map[string]string{"usage": uuidUsage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	if _, err := io.WriteString(stdout, uuidUsage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}
