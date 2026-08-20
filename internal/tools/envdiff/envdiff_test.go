package envdiff

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestCompareReturnsSortedKeyDifferences(t *testing.T) {
	reference := strings.NewReader("# required\nDATABASE_URL=secret\nAPP_ENV=dev\nCACHE_URL=secret\n")
	target := strings.NewReader("APP_ENV=prod\nLOCAL_DEBUG=true\nexport EXTRA_KEY=value\n")

	result, err := Compare(reference, target)
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if !reflect.DeepEqual(result.Missing, []string{"CACHE_URL", "DATABASE_URL"}) {
		t.Errorf("missing = %v, want sorted missing keys", result.Missing)
	}
	if !reflect.DeepEqual(result.Extra, []string{"EXTRA_KEY", "LOCAL_DEBUG"}) {
		t.Errorf("extra = %v, want sorted extra keys", result.Extra)
	}
}

func TestCompareIgnoresValues(t *testing.T) {
	result, err := Compare(strings.NewReader("TOKEN=reference-secret"), strings.NewReader("TOKEN=target-secret"))
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if len(result.Missing) != 0 || len(result.Extra) != 0 {
		t.Errorf("result = %+v, want matching key sets", result)
	}
}

func TestCompareRejectsDuplicateKeyWithoutExposingValue(t *testing.T) {
	secretValue := "do-not-expose"
	_, err := Compare(strings.NewReader("TOKEN=first\nTOKEN="+secretValue), strings.NewReader(""))
	if !errors.Is(err, ErrDuplicateKey) {
		t.Fatalf("Compare() error = %v, want ErrDuplicateKey", err)
	}
	if strings.Contains(err.Error(), secretValue) {
		t.Error("error exposes environment value")
	}
}

func TestCompareRejectsMalformedLine(t *testing.T) {
	_, err := Compare(strings.NewReader("NOT_AN_ASSIGNMENT"), strings.NewReader(""))
	if !errors.Is(err, ErrMalformedLine) {
		t.Errorf("Compare() error = %v, want ErrMalformedLine", err)
	}
}

func TestCompareReportsReadFailure(t *testing.T) {
	_, err := Compare(errorReader{}, strings.NewReader(""))
	if err == nil || errors.Is(err, ErrMalformedLine) {
		t.Errorf("Compare() error = %v, want read failure", err)
	}
}

func FuzzCompare(f *testing.F) {
	f.Add("APP_ENV=dev\n", "APP_ENV=prod\n")
	f.Add("", "")
	f.Fuzz(func(t *testing.T, reference, target string) {
		_, _ = Compare(strings.NewReader(reference), strings.NewReader(target))
	})
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
