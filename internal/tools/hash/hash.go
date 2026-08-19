// Package filehash calculates cryptographic hashes from streams.
package filehash

import (
	"crypto/sha256"
	"crypto/sha512"
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

type Result struct {
	Algorithm Algorithm
	Digest    string
	Bytes     int64
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
