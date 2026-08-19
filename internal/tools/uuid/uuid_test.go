package uuid

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"testing"
)

func TestGenerateProducesVersion4Variant1UUIDs(t *testing.T) {
	values, err := generate(2, bytes.NewReader(make([]byte, 32)))
	if err != nil {
		t.Fatalf("generate() error = %v", err)
	}

	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	for _, value := range values {
		if !pattern.MatchString(value) {
			t.Errorf("generate() value = %q, want UUID v4", value)
		}
	}
}

func TestGenerateRejectsInvalidCount(t *testing.T) {
	for _, count := range []int{-1, 0, MaxCount + 1} {
		_, err := generate(count, bytes.NewReader(nil))
		if !errors.Is(err, ErrInvalidCount) {
			t.Errorf("generate(%d) error = %v, want ErrInvalidCount", count, err)
		}
	}
}

func TestGenerateReportsRandomSourceFailure(t *testing.T) {
	_, err := generate(1, errorReader{})
	if err == nil {
		t.Fatal("generate() error = nil, want random source failure")
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
