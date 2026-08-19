// Package uuid generates cryptographically secure UUID version 4 values.
package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

const (
	DefaultCount = 1
	MaxCount     = 1000
)

var ErrInvalidCount = errors.New("invalid UUID count")

// Generate returns count UUID version 4 strings using the system's secure
// random source.
func Generate(count int) ([]string, error) {
	return generate(count, rand.Reader)
}

func generate(count int, random io.Reader) ([]string, error) {
	if count < 1 || count > MaxCount {
		return nil, fmt.Errorf("%w: must be between 1 and %d", ErrInvalidCount, MaxCount)
	}

	values := make([]string, count)
	buffer := make([]byte, 16)
	for index := range values {
		if _, err := io.ReadFull(random, buffer); err != nil {
			return nil, fmt.Errorf("read secure random bytes: %w", err)
		}

		buffer[6] = (buffer[6] & 0x0f) | 0x40
		buffer[8] = (buffer[8] & 0x3f) | 0x80
		values[index] = format(buffer)
	}

	return values, nil
}

func format(value []byte) string {
	formatted := make([]byte, 36)
	hex.Encode(formatted[0:8], value[0:4])
	formatted[8] = '-'
	hex.Encode(formatted[9:13], value[4:6])
	formatted[13] = '-'
	hex.Encode(formatted[14:18], value[6:8])
	formatted[18] = '-'
	hex.Encode(formatted[19:23], value[8:10])
	formatted[23] = '-'
	hex.Encode(formatted[24:36], value[10:16])
	return string(formatted)
}
