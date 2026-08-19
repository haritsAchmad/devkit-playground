package secret

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestGenerateBase64URL(t *testing.T) {
	result, err := generate(4, EncodingBase64URL, bytes.NewReader([]byte{0xfb, 0xff, 0xef, 0x01}))
	if err != nil {
		t.Fatalf("generate() error = %v", err)
	}

	if result.Value != "-__vAQ" {
		t.Errorf("value = %q, want %q", result.Value, "-__vAQ")
	}
	if result.Encoding != EncodingBase64URL || result.Bytes != 4 {
		t.Errorf("result metadata = %+v, want base64url and 4 bytes", result)
	}
}

func TestGenerateHex(t *testing.T) {
	result, err := generate(4, EncodingHex, bytes.NewReader([]byte{0x00, 0xab, 0xcd, 0xff}))
	if err != nil {
		t.Fatalf("generate() error = %v", err)
	}

	if result.Value != "00abcdff" {
		t.Errorf("value = %q, want %q", result.Value, "00abcdff")
	}
}

func TestGenerateRejectsInvalidLength(t *testing.T) {
	for _, length := range []int{-1, 0, MaxLength + 1} {
		_, err := generate(length, EncodingBase64URL, bytes.NewReader(nil))
		if !errors.Is(err, ErrInvalidLength) {
			t.Errorf("generate(%d) error = %v, want ErrInvalidLength", length, err)
		}
	}
}

func TestGenerateRejectsInvalidEncoding(t *testing.T) {
	_, err := generate(DefaultLength, Encoding("wat"), bytes.NewReader(nil))
	if !errors.Is(err, ErrInvalidEncoding) {
		t.Errorf("generate() error = %v, want ErrInvalidEncoding", err)
	}
}

func TestGenerateReportsRandomSourceFailure(t *testing.T) {
	_, err := generate(1, EncodingBase64URL, errorReader{})
	if err == nil {
		t.Fatal("generate() error = nil, want random source failure")
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
