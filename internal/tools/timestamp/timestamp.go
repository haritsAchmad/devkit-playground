// Package timestamp converts explicit timestamp formats into deterministic UTC facts.
package timestamp

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type InputFormat string

const (
	FormatUnix    InputFormat = "unix"
	FormatUnixMS  InputFormat = "unix-ms"
	FormatRFC3339 InputFormat = "rfc3339"
)

var (
	ErrUnsupportedFormat = errors.New("unsupported timestamp format")
	ErrInvalidValue      = errors.New("invalid timestamp value")
	ErrOutOfRange        = errors.New("timestamp is outside the supported RFC3339 year range")
)

type Result struct {
	InputFormat          InputFormat
	UTC                  time.Time
	UnixSeconds          int64
	UnixMilliseconds     int64
	SubsecondNanoseconds int
}

// Convert parses value in the explicit input format and returns UTC representations.
func Convert(value string, format InputFormat) (Result, error) {
	var parsed time.Time
	switch format {
	case FormatUnix:
		seconds, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return Result{}, fmt.Errorf("%w: unix input must be an integer", ErrInvalidValue)
		}
		parsed = time.Unix(seconds, 0)
	case FormatUnixMS:
		milliseconds, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return Result{}, fmt.Errorf("%w: unix-ms input must be an integer", ErrInvalidValue)
		}
		parsed = time.UnixMilli(milliseconds)
	case FormatRFC3339:
		valueTime, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return Result{}, fmt.Errorf("%w: input must use RFC3339", ErrInvalidValue)
		}
		parsed = valueTime
	default:
		return Result{}, fmt.Errorf("%w: must be %q, %q, or %q", ErrUnsupportedFormat, FormatUnix, FormatUnixMS, FormatRFC3339)
	}

	utc := parsed.UTC()
	if utc.Year() < 0 || utc.Year() > 9999 {
		return Result{}, ErrOutOfRange
	}
	return Result{
		InputFormat:          format,
		UTC:                  utc,
		UnixSeconds:          utc.Unix(),
		UnixMilliseconds:     utc.UnixMilli(),
		SubsecondNanoseconds: utc.Nanosecond(),
	}, nil
}
