package jsonutil

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestParsePreservesLargeNumber(t *testing.T) {
	document, err := Parse(strings.NewReader(`{"value":9007199254740993}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	object, ok := document.Value.(map[string]any)
	if !ok {
		t.Fatalf("value type = %T, want object", document.Value)
	}
	number, ok := object["value"].(json.Number)
	if !ok || number.String() != "9007199254740993" {
		t.Errorf("number = %v, want precision-preserving value", object["value"])
	}
}

func TestPretty(t *testing.T) {
	document, err := Parse(strings.NewReader(`{"name":"Ada","items":[1,2]}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	formatted, err := document.Pretty()
	if err != nil {
		t.Fatalf("Pretty() error = %v", err)
	}
	if !strings.Contains(string(formatted), "\n  \"name\": \"Ada\"") {
		t.Errorf("Pretty() = %q, want indented JSON", formatted)
	}
}

func TestMinify(t *testing.T) {
	document, err := Parse(strings.NewReader("{\n  \"name\": \"Ada\"\n}"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	formatted, err := document.Minify()
	if err != nil {
		t.Fatalf("Minify() error = %v", err)
	}
	if string(formatted) != `{"name":"Ada"}` {
		t.Errorf("Minify() = %q, want compact JSON", formatted)
	}
}

func TestParseRejectsInvalidOrTrailingJSON(t *testing.T) {
	for _, input := range []string{"", `{`, `{} {}`, `{} trailing`} {
		_, err := Parse(strings.NewReader(input))
		if !errors.Is(err, ErrInvalidJSON) {
			t.Errorf("Parse(%q) error = %v, want ErrInvalidJSON", input, err)
		}
	}
}

func TestParseReportsReadFailure(t *testing.T) {
	_, err := Parse(errorReader{})
	if err == nil || errors.Is(err, ErrInvalidJSON) {
		t.Errorf("Parse() error = %v, want non-JSON read failure", err)
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
