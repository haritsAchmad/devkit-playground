package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/haritsAchmad/devkit-playground/internal/output"
	"github.com/haritsAchmad/devkit-playground/internal/tools/secret"
)

const secretUsage = `Usage:
  devkit [--json] secret [--length N] [--encoding base64url|hex]

Flags:
  --length N           number of random bytes (default 32, maximum 4096)
  --encoding FORMAT    output encoding: base64url or hex (default base64url)
  --help               show help for secret
`

type secretData struct {
	Secret   string `json:"secret"`
	Encoding string `json:"encoding"`
	Bytes    int    `json:"bytes"`
}

func runSecret(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	flags := flag.NewFlagSet("secret", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	length := flags.Int("length", secret.DefaultLength, "number of random bytes")
	encoding := flags.String("encoding", string(secret.EncodingBase64URL), "output encoding")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeSecretHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, "secret", "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() != 0 {
		message := fmt.Sprintf("secret does not accept positional arguments: %q", flags.Arg(0))
		return writeFailure(stdout, stderr, jsonMode, "secret", "invalid_usage", message, ExitUsage)
	}

	result, err := secret.Generate(*length, secret.Encoding(*encoding))
	if err != nil {
		switch {
		case errors.Is(err, secret.ErrInvalidLength):
			return writeFailure(stdout, stderr, jsonMode, "secret", "invalid_length", err.Error(), ExitUsage)
		case errors.Is(err, secret.ErrInvalidEncoding):
			return writeFailure(stdout, stderr, jsonMode, "secret", "invalid_encoding", err.Error(), ExitUsage)
		default:
			return writeFailure(stdout, stderr, jsonMode, "secret", "secure_random_failed", "could not generate secure secret", ExitOperation)
		}
	}

	if jsonMode {
		data := secretData{Secret: result.Value, Encoding: string(result.Encoding), Bytes: result.Bytes}
		if err := output.WriteJSONSuccess(stdout, "secret", data); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	if _, err := fmt.Fprintln(stdout, result.Value); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}

func writeSecretHelp(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "secret help", map[string]string{"usage": secretUsage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	if _, err := io.WriteString(stdout, secretUsage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}
