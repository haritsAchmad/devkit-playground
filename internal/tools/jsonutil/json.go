// Package jsonutil validates and formats JSON documents.
package jsonutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var ErrInvalidJSON = errors.New("invalid JSON")

type Document struct {
	Value any
	raw   []byte
}

// Parse reads exactly one JSON value and preserves number precision.
func Parse(input io.Reader) (Document, error) {
	raw, err := io.ReadAll(input)
	if err != nil {
		return Document{}, fmt.Errorf("read JSON input: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return Document{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	if err := ensureEnd(decoder); err != nil {
		return Document{}, fmt.Errorf("%w: trailing data", ErrInvalidJSON)
	}

	return Document{Value: value, raw: bytes.TrimSpace(raw)}, nil
}

// Pretty returns the document with two-space indentation.
func (document Document) Pretty() ([]byte, error) {
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, document.raw, "", "  "); err != nil {
		return nil, fmt.Errorf("format JSON: %w", err)
	}
	return formatted.Bytes(), nil
}

// Minify returns the document without insignificant whitespace.
func (document Document) Minify() ([]byte, error) {
	var formatted bytes.Buffer
	if err := json.Compact(&formatted, document.raw); err != nil {
		return nil, fmt.Errorf("minify JSON: %w", err)
	}
	return formatted.Bytes(), nil
}

func ensureEnd(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("additional JSON value")
		}
		return err
	}
	return nil
}
