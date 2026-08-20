package base64util

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestTransformEncodesVariants(t *testing.T) {
	cases := []struct {
		variant Variant
		padding Padding
		want    string
	}{
		{VariantStandard, PaddingPadded, "+/8="},
		{VariantStandard, PaddingRaw, "+/8"},
		{VariantURL, PaddingPadded, "-_8="},
		{VariantURL, PaddingRaw, "-_8"},
	}
	for _, test := range cases {
		result, err := Transform(bytes.NewReader([]byte{0xfb, 0xff}), OperationEncode, test.variant, test.padding)
		if err != nil {
			t.Fatalf("Transform() error = %v", err)
		}
		if string(result.Value) != test.want {
			t.Errorf("value = %q, want %q", result.Value, test.want)
		}
	}
}

func TestTransformDecodesSurroundingWhitespace(t *testing.T) {
	result, err := Transform(strings.NewReader("  aGVsbG8=\r\n"), OperationDecode, VariantStandard, PaddingPadded)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}
	if string(result.Value) != "hello" || result.OutputBytes != 5 {
		t.Errorf("result = %+v, want decoded hello", result)
	}
}

func TestTransformRejectsInvalidInputWithoutEchoingIt(t *testing.T) {
	secret := "not-base64-secret"
	_, err := Transform(strings.NewReader(secret), OperationDecode, VariantStandard, PaddingPadded)
	if !errors.Is(err, ErrInvalidEncoding) {
		t.Fatalf("Transform() error = %v, want ErrInvalidEncoding", err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Error("error exposes invalid input")
	}
}

func TestTransformRejectsOversizedInput(t *testing.T) {
	_, err := Transform(strings.NewReader(strings.Repeat("a", MaxInputSize+1)), OperationEncode, VariantStandard, PaddingPadded)
	if !errors.Is(err, ErrInputTooLarge) {
		t.Errorf("Transform() error = %v, want ErrInputTooLarge", err)
	}
}

func FuzzRoundTrip(f *testing.F) {
	f.Add([]byte("hello"), false, false)
	f.Add([]byte{0x00, 0xfb, 0xff}, true, true)
	f.Fuzz(func(t *testing.T, input []byte, urlVariant, rawPadding bool) {
		if len(input) > 1<<20 {
			t.Skip()
		}
		variant := VariantStandard
		if urlVariant {
			variant = VariantURL
		}
		padding := PaddingPadded
		if rawPadding {
			padding = PaddingRaw
		}
		encoded, err := Transform(bytes.NewReader(input), OperationEncode, variant, padding)
		if err != nil {
			t.Fatalf("encode: %v", err)
		}
		decoded, err := Transform(bytes.NewReader(encoded.Value), OperationDecode, variant, padding)
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if !bytes.Equal(decoded.Value, input) {
			t.Fatal("round trip changed input")
		}
	})
}
