// Package filehash calculates cryptographic hashes from streams.
package filehash

import (
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
)

type Algorithm string

const (
	AlgorithmSHA256 Algorithm = "sha256"
	AlgorithmSHA512 Algorithm = "sha512"
)

var ErrUnsupportedAlgorithm = errors.New("unsupported hash algorithm")
var ErrInvalidDigest = errors.New("invalid expected digest")

type Result struct {
	Algorithm Algorithm
	Digest    string
	Bytes     int64
}

type Verification struct {
	Result
	Match bool
}

// Sum streams input through the selected hash algorithm.
func Sum(input io.Reader, algorithm Algorithm) (Result, error) {
	hasher, err := newHasher(algorithm)
	if err != nil {
		return Result{}, err
	}

	bytesRead, err := io.Copy(hasher, input)
	if err != nil {
		return Result{}, fmt.Errorf("read hash input: %w", err)
	}

	return Result{
		Algorithm: algorithm,
		Digest:    hex.EncodeToString(hasher.Sum(nil)),
		Bytes:     bytesRead,
	}, nil
}

// Verify streams input and compares its digest with an expected hexadecimal value.
func Verify(input io.Reader, algorithm Algorithm, expected string) (Verification, error) {
	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return Verification{}, fmt.Errorf("%w: must be hexadecimal", ErrInvalidDigest)
	}
	hasher, err := newHasher(algorithm)
	if err != nil {
		return Verification{}, err
	}
	if len(expectedBytes) != hasher.Size() {
		return Verification{}, fmt.Errorf("%w: must contain %d hexadecimal characters", ErrInvalidDigest, hasher.Size()*2)
	}
	bytesRead, err := io.Copy(hasher, input)
	if err != nil {
		return Verification{}, fmt.Errorf("read hash input: %w", err)
	}
	actualBytes := hasher.Sum(nil)
	return Verification{
		Result: Result{
			Algorithm: algorithm,
			Digest:    hex.EncodeToString(actualBytes),
			Bytes:     bytesRead,
		},
		Match: subtle.ConstantTimeCompare(actualBytes, expectedBytes) == 1,
	}, nil
}

func newHasher(algorithm Algorithm) (hash.Hash, error) {
	switch algorithm {
	case AlgorithmSHA256:
		return sha256.New(), nil
	case AlgorithmSHA512:
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("%w: must be %q or %q", ErrUnsupportedAlgorithm, AlgorithmSHA256, AlgorithmSHA512)
	}
}
