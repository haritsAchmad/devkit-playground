// Package fileinspect reports deterministic metadata about regular files.
package fileinspect

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

const sniffSize = 512

type ExtensionCheck string

const (
	ExtensionMatch    ExtensionCheck = "match"
	ExtensionMismatch ExtensionCheck = "mismatch"
	ExtensionUnknown  ExtensionCheck = "unknown"
)

type Result struct {
	Name           string
	Extension      string
	SizeBytes      int64
	DetectedMIME   string
	ExtensionCheck ExtensionCheck
	SHA256         string
}

// Inspect reads a file stream once and reports content-derived metadata.
func Inspect(input io.Reader, name string) (Result, error) {
	prefix := make([]byte, sniffSize)
	read, err := io.ReadFull(input, prefix)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return Result{}, err
	}
	prefix = prefix[:read]

	hasher := sha256.New()
	size, err := io.Copy(hasher, io.MultiReader(bytes.NewReader(prefix), input))
	if err != nil {
		return Result{}, err
	}

	extension := strings.ToLower(filepath.Ext(name))
	detectedMIME := http.DetectContentType(prefix)
	return Result{
		Name:           filepath.Base(name),
		Extension:      extension,
		SizeBytes:      size,
		DetectedMIME:   detectedMIME,
		ExtensionCheck: checkExtension(extension, detectedMIME),
		SHA256:         hex.EncodeToString(hasher.Sum(nil)),
	}, nil
}

func checkExtension(extension, detectedMIME string) ExtensionCheck {
	known := map[string][]string{
		"application/pdf":              {".pdf"},
		"application/zip":              {".zip"},
		"application/x-gzip":           {".gz"},
		"application/x-rar-compressed": {".rar"},
		"application/x-7z-compressed":  {".7z"},
		"image/gif":                    {".gif"},
		"image/jpeg":                   {".jpeg", ".jpg"},
		"image/png":                    {".png"},
		"image/webp":                   {".webp"},
	}

	extensions, ok := known[detectedMIME]
	if !ok || extension == "" {
		return ExtensionUnknown
	}
	for _, candidate := range extensions {
		if extension == candidate {
			return ExtensionMatch
		}
	}
	return ExtensionMismatch
}
