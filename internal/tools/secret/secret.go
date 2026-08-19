// Package secret generates cryptographically secure encoded secrets.
package secret

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

const (
	DefaultLength = 32
	MaxLength     = 4096
)

type Encoding string

const (
	EncodingBase64URL Encoding = "base64url"
	EncodingHex       Encoding = "hex"
)

var (
	ErrInvalidLength   = errors.New("invalid secret length")
	ErrInvalidEncoding = errors.New("invalid secret encoding")
)

type Result struct {
	Value    string
	Encoding Encoding
	Bytes    int
}

// Generate returns a secret produced with the system's secure random source.
func Generate(length int, encoding Encoding) (Result, error) {
	return generate(length, encoding, rand.Reader)
}

func generate(length int, encoding Encoding, random io.Reader) (Result, error) {
	if length < 1 || length > MaxLength {
		return Result{}, fmt.Errorf("%w: must be between 1 and %d bytes", ErrInvalidLength, MaxLength)
	}
	if encoding != EncodingBase64URL && encoding != EncodingHex {
		return Result{}, fmt.Errorf("%w: must be %q or %q", ErrInvalidEncoding, EncodingBase64URL, EncodingHex)
	}

	raw := make([]byte, length)
	if _, err := io.ReadFull(random, raw); err != nil {
		return Result{}, fmt.Errorf("read secure random bytes: %w", err)
	}

	value := base64.RawURLEncoding.EncodeToString(raw)
	if encoding == EncodingHex {
		value = hex.EncodeToString(raw)
	}

	return Result{Value: value, Encoding: encoding, Bytes: length}, nil
}
