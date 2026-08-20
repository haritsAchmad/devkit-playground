package timestamp

import (
	"errors"
	"testing"
	"time"
)

func TestConvertUnix(t *testing.T) {
	result, err := Convert("0", FormatUnix)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}
	if result.UTC.Format(time.RFC3339Nano) != "1970-01-01T00:00:00Z" || result.UnixSeconds != 0 || result.UnixMilliseconds != 0 {
		t.Errorf("result = %+v, want Unix epoch", result)
	}
}

func TestConvertUnixMillisecondsPreservesSubsecond(t *testing.T) {
	result, err := Convert("1720000000123", FormatUnixMS)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}
	if result.UnixMilliseconds != 1720000000123 || result.SubsecondNanoseconds != 123000000 {
		t.Errorf("result = %+v, want millisecond precision", result)
	}
}

func TestConvertRFC3339NormalizesToUTC(t *testing.T) {
	result, err := Convert("2026-08-20T10:30:15.123456789+07:00", FormatRFC3339)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}
	if got := result.UTC.Format(time.RFC3339Nano); got != "2026-08-20T03:30:15.123456789Z" {
		t.Errorf("UTC = %q, want normalized timestamp", got)
	}
	if result.SubsecondNanoseconds != 123456789 {
		t.Errorf("SubsecondNanoseconds = %d, want 123456789", result.SubsecondNanoseconds)
	}
}

func TestConvertRejectsInvalidInput(t *testing.T) {
	if _, err := Convert("wat", FormatUnix); !errors.Is(err, ErrInvalidValue) {
		t.Errorf("Convert() error = %v, want ErrInvalidValue", err)
	}
	if _, err := Convert("0", "auto"); !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("Convert() error = %v, want ErrUnsupportedFormat", err)
	}
}

func FuzzConvert(f *testing.F) {
	f.Add("0", "unix")
	f.Add("2026-08-20T10:30:15+07:00", "rfc3339")
	f.Fuzz(func(t *testing.T, value, format string) {
		_, _ = Convert(value, InputFormat(format))
	})
}
