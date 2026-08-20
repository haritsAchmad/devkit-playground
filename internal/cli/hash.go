package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/haritsAchmad/devkit-playground/internal/output"
	filehash "github.com/haritsAchmad/devkit-playground/internal/tools/hash"
)

const hashUsage = `Usage:
  devkit [--json] hash [--algorithm sha256|sha512] [file]
  devkit [--json] hash verify --expected HEX [--algorithm sha256|sha512] [file]

Arguments:
  file                  file to hash; omit or use - to read standard input

Flags:
  --algorithm NAME      hash algorithm: sha256 or sha512 (default sha256)
  --expected HEX        expected digest for hash verify
  --help                show help for hash
`

type hashData struct {
	Algorithm string `json:"algorithm"`
	Digest    string `json:"digest"`
	Bytes     int64  `json:"bytes"`
	Source    string `json:"source"`
}

type hashVerifyData struct {
	Algorithm string `json:"algorithm"`
	Digest    string `json:"digest"`
	Bytes     int64  `json:"bytes"`
	Source    string `json:"source"`
	Verified  bool   `json:"verified"`
}

func runHash(args []string, stdin io.Reader, stdout, stderr io.Writer, jsonMode bool) int {
	if len(args) > 0 && args[0] == "verify" {
		return runHashVerify(args[1:], stdin, stdout, stderr, jsonMode)
	}
	flags := flag.NewFlagSet("hash", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	algorithm := flags.String("algorithm", string(filehash.AlgorithmSHA256), "hash algorithm")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeHashHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, "hash", "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() > 1 {
		return writeFailure(stdout, stderr, jsonMode, "hash", "invalid_usage", "hash accepts at most one file", ExitUsage)
	}

	input := stdin
	source := "stdin"
	var closeInput func() error
	if flags.NArg() == 1 && flags.Arg(0) != "-" {
		file, err := os.Open(flags.Arg(0))
		if err != nil {
			return writeFailure(stdout, stderr, jsonMode, "hash", "file_read_failed", "could not open input file", ExitOperation)
		}
		input = file
		source = "file"
		closeInput = file.Close
	}
	if closeInput != nil {
		defer closeInput()
	}

	result, err := filehash.Sum(input, filehash.Algorithm(*algorithm))
	if err != nil {
		if errors.Is(err, filehash.ErrUnsupportedAlgorithm) {
			return writeFailure(stdout, stderr, jsonMode, "hash", "unsupported_algorithm", err.Error(), ExitUsage)
		}
		return writeFailure(stdout, stderr, jsonMode, "hash", "input_read_failed", "could not read hash input", ExitOperation)
	}

	if jsonMode {
		data := hashData{Algorithm: string(result.Algorithm), Digest: result.Digest, Bytes: result.Bytes, Source: source}
		if err := output.WriteJSONSuccess(stdout, "hash", data); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	if _, err := fmt.Fprintln(stdout, result.Digest); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}

func runHashVerify(args []string, stdin io.Reader, stdout, stderr io.Writer, jsonMode bool) int {
	flags := flag.NewFlagSet("hash verify", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	algorithm := flags.String("algorithm", string(filehash.AlgorithmSHA256), "hash algorithm")
	expected := flags.String("expected", "", "expected hexadecimal digest")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeHashHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, "hash verify", "invalid_usage", err.Error(), ExitUsage)
	}
	if *expected == "" {
		return writeFailure(stdout, stderr, jsonMode, "hash verify", "invalid_usage", "hash verify requires --expected", ExitUsage)
	}
	if flags.NArg() > 1 {
		return writeFailure(stdout, stderr, jsonMode, "hash verify", "invalid_usage", "hash verify accepts at most one file", ExitUsage)
	}

	input := stdin
	source := "stdin"
	var closeInput func() error
	if flags.NArg() == 1 && flags.Arg(0) != "-" {
		file, err := os.Open(flags.Arg(0))
		if err != nil {
			return writeFailure(stdout, stderr, jsonMode, "hash verify", "file_read_failed", "could not open input file", ExitOperation)
		}
		input = file
		source = "file"
		closeInput = file.Close
	}
	if closeInput != nil {
		defer closeInput()
	}

	verification, err := filehash.Verify(input, filehash.Algorithm(*algorithm), *expected)
	if err != nil {
		switch {
		case errors.Is(err, filehash.ErrUnsupportedAlgorithm):
			return writeFailure(stdout, stderr, jsonMode, "hash verify", "unsupported_algorithm", err.Error(), ExitUsage)
		case errors.Is(err, filehash.ErrInvalidDigest):
			return writeFailure(stdout, stderr, jsonMode, "hash verify", "invalid_digest", err.Error(), ExitUsage)
		default:
			return writeFailure(stdout, stderr, jsonMode, "hash verify", "input_read_failed", "could not read hash input", ExitOperation)
		}
	}
	if !verification.Match {
		return writeFailure(stdout, stderr, jsonMode, "hash verify", "checksum_mismatch", "checksum does not match", ExitData)
	}

	data := hashVerifyData{
		Algorithm: string(verification.Algorithm), Digest: verification.Digest,
		Bytes: verification.Bytes, Source: source, Verified: true,
	}
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "hash verify", data); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if _, err := fmt.Fprintf(stdout, "Verified %s: %s\n", data.Algorithm, data.Digest); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}

func writeHashHelp(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "hash help", map[string]string{"usage": hashUsage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}

	if _, err := io.WriteString(stdout, hashUsage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}
