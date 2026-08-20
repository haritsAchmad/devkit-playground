package fileinspect

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestInspectDetectsMatchingPDF(t *testing.T) {
	content := "%PDF-1.7\nfixture"
	result, err := Inspect(strings.NewReader(content), "report.PDF")
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}

	wantHash := sha256.Sum256([]byte(content))
	if result.Name != "report.PDF" || result.Extension != ".pdf" || result.SizeBytes != int64(len(content)) {
		t.Errorf("result = %+v, want PDF file metadata", result)
	}
	if result.DetectedMIME != "application/pdf" || result.ExtensionCheck != ExtensionMatch {
		t.Errorf("detection = %q/%q, want application/pdf/match", result.DetectedMIME, result.ExtensionCheck)
	}
	if result.SHA256 != hex.EncodeToString(wantHash[:]) {
		t.Errorf("SHA256 = %q, want %x", result.SHA256, wantHash)
	}
}

func TestInspectDetectsExtensionMismatch(t *testing.T) {
	result, err := Inspect(strings.NewReader("%PDF-1.7\nfixture"), "report.exe")
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.ExtensionCheck != ExtensionMismatch {
		t.Errorf("ExtensionCheck = %q, want mismatch", result.ExtensionCheck)
	}
}

func TestInspectUsesUnknownWhenTypeHasNoStableMapping(t *testing.T) {
	result, err := Inspect(strings.NewReader("plain text"), "notes.txt")
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.ExtensionCheck != ExtensionUnknown {
		t.Errorf("ExtensionCheck = %q, want unknown", result.ExtensionCheck)
	}
}

func FuzzInspect(f *testing.F) {
	f.Add([]byte("%PDF-1.7\nfixture"), "report.pdf")
	f.Add([]byte{}, "empty")
	f.Fuzz(func(t *testing.T, content []byte, name string) {
		result, err := Inspect(bytes.NewReader(content), name)
		if err != nil {
			t.Fatalf("Inspect() error = %v", err)
		}
		if result.SizeBytes != int64(len(content)) {
			t.Fatalf("SizeBytes = %d, want %d", result.SizeBytes, len(content))
		}
	})
}
