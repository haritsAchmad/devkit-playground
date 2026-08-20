package filehash

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestSumSHA256(t *testing.T) {
	result, err := Sum(strings.NewReader("hello"), AlgorithmSHA256)
	if err != nil {
		t.Fatalf("Sum() error = %v", err)
	}

	const expected = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if result.Digest != expected {
		t.Errorf("digest = %q, want %q", result.Digest, expected)
	}
	if result.Bytes != 5 || result.Algorithm != AlgorithmSHA256 {
		t.Errorf("result metadata = %+v, want sha256 and 5 bytes", result)
	}
}

func TestSumSHA512(t *testing.T) {
	result, err := Sum(strings.NewReader("hello"), AlgorithmSHA512)
	if err != nil {
		t.Fatalf("Sum() error = %v", err)
	}

	const expected = "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"
	if result.Digest != expected {
		t.Errorf("digest = %q, want %q", result.Digest, expected)
	}
}

func TestSumRejectsUnsupportedAlgorithm(t *testing.T) {
	_, err := Sum(strings.NewReader("hello"), Algorithm("md5"))
	if !errors.Is(err, ErrUnsupportedAlgorithm) {
		t.Errorf("Sum() error = %v, want ErrUnsupportedAlgorithm", err)
	}
}

func TestVerifyMatchesCaseInsensitiveDigest(t *testing.T) {
	verification, err := Verify(strings.NewReader("hello"), AlgorithmSHA256, "2CF24DBA5FB0A30E26E83B2AC5B9E29E1B161E5C1FA7425E73043362938B9824")
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !verification.Match || verification.Bytes != 5 {
		t.Errorf("verification = %+v, want matching five-byte input", verification)
	}
}

func TestVerifyReportsMismatch(t *testing.T) {
	verification, err := Verify(strings.NewReader("hello"), AlgorithmSHA256, strings.Repeat("0", 64))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if verification.Match {
		t.Error("Match = true, want false")
	}
}

func TestVerifyRejectsInvalidDigest(t *testing.T) {
	for _, expected := range []string{"not-hex", "abcd"} {
		if _, err := Verify(strings.NewReader("hello"), AlgorithmSHA256, expected); !errors.Is(err, ErrInvalidDigest) {
			t.Errorf("Verify(%q) error = %v, want ErrInvalidDigest", expected, err)
		}
	}
}

func TestSumReportsReadFailure(t *testing.T) {
	_, err := Sum(errorReader{}, AlgorithmSHA256)
	if err == nil {
		t.Fatal("Sum() error = nil, want read failure")
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
