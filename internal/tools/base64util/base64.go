// Package base64util performs bounded Base64 encoding and decoding.
package base64util

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

const MaxInputSize = 16 << 20

type Operation string
type Variant string
type Padding string

const (
	OperationEncode Operation = "encode"
	OperationDecode Operation = "decode"
	VariantStandard Variant   = "standard"
	VariantURL      Variant   = "url"
	PaddingPadded   Padding   = "padded"
	PaddingRaw      Padding   = "raw"
)

var (
	ErrUnsupportedOperation = errors.New("unsupported Base64 operation")
	ErrUnsupportedVariant   = errors.New("unsupported Base64 variant")
	ErrUnsupportedPadding   = errors.New("unsupported Base64 padding")
	ErrInputTooLarge        = errors.New("Base64 input exceeds maximum size")
	ErrInvalidEncoding      = errors.New("invalid Base64 input")
)

type Result struct {
	Value       []byte
	InputBytes  int
	OutputBytes int
	Operation   Operation
	Variant     Variant
	Padding     Padding
}

// Transform reads bounded input and applies the selected Base64 operation.
func Transform(input io.Reader, operation Operation, variant Variant, padding Padding) (Result, error) {
	encoding, err := selectEncoding(variant, padding)
	if err != nil {
		return Result{}, err
	}
	raw, err := io.ReadAll(io.LimitReader(input, MaxInputSize+1))
	if err != nil {
		return Result{}, fmt.Errorf("read Base64 input: %w", err)
	}
	if len(raw) > MaxInputSize {
		return Result{}, ErrInputTooLarge
	}

	var output []byte
	switch operation {
	case OperationEncode:
		output = make([]byte, encoding.EncodedLen(len(raw)))
		encoding.Encode(output, raw)
	case OperationDecode:
		trimmed := bytes.TrimSpace(raw)
		output = make([]byte, encoding.DecodedLen(len(trimmed)))
		written, err := encoding.Decode(output, trimmed)
		if err != nil {
			return Result{}, ErrInvalidEncoding
		}
		output = output[:written]
	default:
		return Result{}, fmt.Errorf("%w: must be %q or %q", ErrUnsupportedOperation, OperationEncode, OperationDecode)
	}

	return Result{
		Value: output, InputBytes: len(raw), OutputBytes: len(output),
		Operation: operation, Variant: variant, Padding: padding,
	}, nil
}

func selectEncoding(variant Variant, padding Padding) (*base64.Encoding, error) {
	var encoding *base64.Encoding
	switch variant {
	case VariantStandard:
		encoding = base64.StdEncoding
	case VariantURL:
		encoding = base64.URLEncoding
	default:
		return nil, fmt.Errorf("%w: must be %q or %q", ErrUnsupportedVariant, VariantStandard, VariantURL)
	}
	switch padding {
	case PaddingPadded:
		return encoding, nil
	case PaddingRaw:
		return encoding.WithPadding(base64.NoPadding), nil
	default:
		return nil, fmt.Errorf("%w: must be %q or %q", ErrUnsupportedPadding, PaddingPadded, PaddingRaw)
	}
}
