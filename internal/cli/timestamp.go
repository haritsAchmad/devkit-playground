package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/haritsAchmad/devkit-playground/internal/output"
	"github.com/haritsAchmad/devkit-playground/internal/tools/timestamp"
)

const timestampConvertUsage = `Usage:
  devkit [--json] timestamp convert [--from unix|unix-ms|rfc3339] <value>

Arguments:
  value                 timestamp value in the selected input format

Flags:
  --from FORMAT         unix, unix-ms, or rfc3339 (default unix)

Output is normalized to UTC and does not depend on the machine timezone.
`

type timestampData struct {
	InputFormat          string `json:"input_format"`
	UTC                  string `json:"utc"`
	UnixSeconds          int64  `json:"unix_seconds"`
	UnixMilliseconds     int64  `json:"unix_milliseconds"`
	SubsecondNanoseconds int    `json:"subsecond_nanoseconds"`
}

func runTimestamp(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	if len(args) == 0 {
		return writeFailure(stdout, stderr, jsonMode, "timestamp", "invalid_usage", "timestamp requires the convert subcommand", ExitUsage)
	}
	if args[0] == "--help" {
		return writeTimestampHelp(stdout, jsonMode)
	}
	if args[0] != "convert" {
		return writeFailure(stdout, stderr, jsonMode, "timestamp", "unknown_subcommand", "unknown timestamp subcommand", ExitUsage)
	}
	return runTimestampConvert(args[1:], stdout, stderr, jsonMode)
}

func runTimestampConvert(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	flags := flag.NewFlagSet("timestamp convert", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	inputFormat := flags.String("from", string(timestamp.FormatUnix), "input timestamp format")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeTimestampHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, "timestamp convert", "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() != 1 {
		return writeFailure(stdout, stderr, jsonMode, "timestamp convert", "invalid_usage", "timestamp convert requires exactly one value", ExitUsage)
	}

	result, err := timestamp.Convert(flags.Arg(0), timestamp.InputFormat(*inputFormat))
	if err != nil {
		switch {
		case errors.Is(err, timestamp.ErrUnsupportedFormat):
			return writeFailure(stdout, stderr, jsonMode, "timestamp convert", "unsupported_format", err.Error(), ExitUsage)
		case errors.Is(err, timestamp.ErrOutOfRange):
			return writeFailure(stdout, stderr, jsonMode, "timestamp convert", "timestamp_out_of_range", err.Error(), ExitData)
		default:
			return writeFailure(stdout, stderr, jsonMode, "timestamp convert", "invalid_timestamp", err.Error(), ExitData)
		}
	}
	data := timestampData{
		InputFormat: string(result.InputFormat), UTC: result.UTC.Format(time.RFC3339Nano),
		UnixSeconds: result.UnixSeconds, UnixMilliseconds: result.UnixMilliseconds,
		SubsecondNanoseconds: result.SubsecondNanoseconds,
	}
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "timestamp convert", data); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if _, err := fmt.Fprintf(stdout, "UTC: %s\nUnix seconds: %d\nUnix milliseconds: %d\nSubsecond nanoseconds: %d\n", data.UTC, data.UnixSeconds, data.UnixMilliseconds, data.SubsecondNanoseconds); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}

func writeTimestampHelp(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "timestamp convert help", map[string]string{"usage": timestampConvertUsage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if _, err := io.WriteString(stdout, timestampConvertUsage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}
