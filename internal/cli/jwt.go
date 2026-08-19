package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/haritsAchmad/devkit-playground/internal/output"
	jwtinspect "github.com/haritsAchmad/devkit-playground/internal/tools/jwt"
)

const jwtUsage = `Usage:
  devkit [--json] jwt inspect [token]

Arguments:
  token         compact JWT; omit to read from standard input (recommended)

Notes:
  Inspection decodes header and claims but NEVER verifies the signature.
  Prefer stdin because command arguments may be stored in shell history.
`

type jwtData struct {
	Header            map[string]any `json:"header"`
	Claims            map[string]any `json:"claims"`
	Algorithm         string         `json:"algorithm"`
	SignatureVerified bool           `json:"signature_verified"`
}

func runJWT(args []string, stdin io.Reader, stdout, stderr io.Writer, jsonMode bool) int {
	if len(args) == 0 {
		return writeFailure(stdout, stderr, jsonMode, "jwt", "invalid_usage", "jwt requires the inspect subcommand", ExitUsage)
	}
	if args[0] == "--help" {
		return writeJWTHelp(stdout, jsonMode)
	}
	if args[0] != "inspect" {
		return writeFailure(stdout, stderr, jsonMode, "jwt", "unknown_subcommand", "unknown jwt subcommand", ExitUsage)
	}

	return runJWTInspect(args[1:], stdin, stdout, stderr, jsonMode)
}

func runJWTInspect(args []string, stdin io.Reader, stdout, stderr io.Writer, jsonMode bool) int {
	flags := flag.NewFlagSet("jwt inspect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeJWTHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, "jwt inspect", "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() > 1 {
		return writeFailure(stdout, stderr, jsonMode, "jwt inspect", "invalid_usage", "jwt inspect accepts at most one token", ExitUsage)
	}

	token, readErr := readJWTToken(flags.Args(), stdin)
	if readErr != nil {
		return writeFailure(stdout, stderr, jsonMode, "jwt inspect", "input_read_failed", "could not read JWT input", ExitOperation)
	}
	result, err := jwtinspect.Inspect(token)
	if err != nil {
		return writeFailure(stdout, stderr, jsonMode, "jwt inspect", "invalid_token", safeJWTError(err), ExitData)
	}

	data := jwtData{
		Header:            result.Header,
		Claims:            result.Claims,
		Algorithm:         result.Algorithm,
		SignatureVerified: false,
	}
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "jwt inspect", data); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	return writeJWTHuman(stdout, data)
}

func readJWTToken(args []string, stdin io.Reader) (string, error) {
	if len(args) == 1 {
		return strings.TrimSpace(args[0]), nil
	}

	raw, err := io.ReadAll(io.LimitReader(stdin, jwtinspect.MaxTokenSize+1))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}

func safeJWTError(err error) string {
	switch {
	case errors.Is(err, jwtinspect.ErrTokenTooLarge):
		return "JWT exceeds maximum size of 1 MiB"
	case errors.Is(err, jwtinspect.ErrInvalidStructure):
		return "token must contain three compact JWT segments"
	case errors.Is(err, jwtinspect.ErrInvalidEncoding):
		return "JWT contains malformed base64url"
	case errors.Is(err, jwtinspect.ErrInvalidJSON):
		return "JWT header and payload must be JSON objects"
	case errors.Is(err, jwtinspect.ErrMissingAlgorithm):
		return "JWT header must contain a non-empty alg value"
	default:
		return "JWT could not be inspected"
	}
}

func writeJWTHuman(stdout io.Writer, data jwtData) int {
	header, err := marshalIndented(data.Header)
	if err != nil {
		return ExitInternal
	}
	claims, err := marshalIndented(data.Claims)
	if err != nil {
		return ExitInternal
	}

	_, err = fmt.Fprintf(stdout, "WARNING: signature not verified; treat all claims as untrusted.\n\nAlgorithm: %s\n\nHeader:\n%s\n\nClaims:\n%s\n", data.Algorithm, header, claims)
	if err != nil {
		return ExitInternal
	}
	return ExitSuccess
}

func marshalIndented(value any) (string, error) {
	var destination strings.Builder
	encoder := json.NewEncoder(&destination)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return "", err
	}
	return strings.TrimSuffix(destination.String(), "\n"), nil
}

func writeJWTHelp(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "jwt inspect help", map[string]string{"usage": jwtUsage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	if _, err := io.WriteString(stdout, jwtUsage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}
