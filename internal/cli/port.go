package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/haritsAchmad/devkit-playground/internal/output"
	"github.com/haritsAchmad/devkit-playground/internal/tools/portinspect"
)

const portInspectUsage = `Usage:
  devkit [--json] port inspect [--host IP] <port>

Arguments:
  port                  TCP port from 1 through 65535

Flags:
  --host IP             local IPv4 or IPv6 address (default 127.0.0.1)

The probe briefly binds the address without connecting to a target service.
Process-owner lookup is not available in the portable implementation.
`

type portInspectData struct {
	Host            string  `json:"host"`
	Port            int     `json:"port"`
	Protocol        string  `json:"protocol"`
	State           string  `json:"state"`
	Reason          *string `json:"reason"`
	PID             *int    `json:"pid"`
	Process         *string `json:"process"`
	OwnerInspection string  `json:"owner_inspection"`
}

func runPort(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	if len(args) == 0 {
		return writeFailure(stdout, stderr, jsonMode, "port", "invalid_usage", "port requires the inspect subcommand", ExitUsage)
	}
	if args[0] == "--help" {
		return writePortInspectHelp(stdout, jsonMode)
	}
	if args[0] != "inspect" {
		return writeFailure(stdout, stderr, jsonMode, "port", "unknown_subcommand", "unknown port subcommand", ExitUsage)
	}
	return runPortInspect(args[1:], stdout, stderr, jsonMode)
}

func runPortInspect(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	flags := flag.NewFlagSet("port inspect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	host := flags.String("host", "127.0.0.1", "local IP address")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writePortInspectHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, "port inspect", "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() != 1 {
		return writeFailure(stdout, stderr, jsonMode, "port inspect", "invalid_usage", "port inspect requires exactly one port", ExitUsage)
	}
	port, err := strconv.Atoi(flags.Arg(0))
	if err != nil {
		return writeFailure(stdout, stderr, jsonMode, "port inspect", "invalid_port", "port must be an integer from 1 through 65535", ExitUsage)
	}

	result, err := portinspect.Inspect(*host, port)
	if err != nil {
		switch {
		case errors.Is(err, portinspect.ErrInvalidHost):
			return writeFailure(stdout, stderr, jsonMode, "port inspect", "invalid_host", err.Error(), ExitUsage)
		case errors.Is(err, portinspect.ErrInvalidPort):
			return writeFailure(stdout, stderr, jsonMode, "port inspect", "invalid_port", err.Error(), ExitUsage)
		default:
			return writeFailure(stdout, stderr, jsonMode, "port inspect", "probe_failed", "could not inspect local port", ExitOperation)
		}
	}
	data := portInspectData{
		Host: result.Host, Port: result.Port, Protocol: result.Protocol,
		State: string(result.State), OwnerInspection: "not_supported",
	}
	if result.Reason != "" {
		data.Reason = &result.Reason
	}
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "port inspect", data); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	if _, err := fmt.Fprintf(stdout, "Host: %s\nPort: %d\nProtocol: %s\nState: %s\n", data.Host, data.Port, data.Protocol, data.State); err != nil {
		return ExitInternal
	}
	if data.Reason != nil {
		if _, err := fmt.Fprintf(stdout, "Reason: %s\n", *data.Reason); err != nil {
			return ExitInternal
		}
	}
	if _, err := io.WriteString(stdout, "Process owner: unavailable (portable inspection)\n"); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}

func writePortInspectHelp(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "port inspect help", map[string]string{"usage": portInspectUsage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if _, err := io.WriteString(stdout, portInspectUsage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}
