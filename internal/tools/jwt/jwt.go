// Package jwtinspect decodes JWT metadata without verifying its signature.
package jwtinspect

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const MaxTokenSize = 1 << 20

var (
	ErrTokenTooLarge    = errors.New("JWT exceeds maximum size")
	ErrInvalidStructure = errors.New("invalid JWT structure")
	ErrInvalidEncoding  = errors.New("invalid JWT base64url encoding")
	ErrInvalidJSON      = errors.New("invalid JWT JSON")
	ErrMissingAlgorithm = errors.New("JWT header is missing algorithm")
)

type Result struct {
	Header    map[string]any
	Claims    map[string]any
	Algorithm string
}

// Inspect decodes a compact JWT without validating or trusting its signature.
func Inspect(token string) (Result, error) {
	if len(token) > MaxTokenSize {
		return Result{}, ErrTokenTooLarge
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" {
		return Result{}, fmt.Errorf("%w: token must contain three segments", ErrInvalidStructure)
	}

	header, err := decodeObject(parts[0], "header")
	if err != nil {
		return Result{}, err
	}
	claims, err := decodeObject(parts[1], "payload")
	if err != nil {
		return Result{}, err
	}

	algorithm, ok := header["alg"].(string)
	if !ok || algorithm == "" {
		return Result{}, ErrMissingAlgorithm
	}

	return Result{Header: header, Claims: claims, Algorithm: algorithm}, nil
}

func decodeObject(segment, name string) (map[string]any, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return nil, fmt.Errorf("%w: malformed %s segment", ErrInvalidEncoding, name)
	}

	decoder := json.NewDecoder(bytes.NewReader(decoded))
	decoder.UseNumber()
	var value map[string]any
	if err := decoder.Decode(&value); err != nil || value == nil {
		return nil, fmt.Errorf("%w: %s must be a JSON object", ErrInvalidJSON, name)
	}
	if err := ensureJSONEnd(decoder); err != nil {
		return nil, fmt.Errorf("%w: %s contains trailing data", ErrInvalidJSON, name)
	}
	return value, nil
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("additional JSON value")
		}
		return err
	}
	return nil
}
