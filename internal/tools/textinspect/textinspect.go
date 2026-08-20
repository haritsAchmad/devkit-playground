// Package textinspect reports encoding and line-ending facts about text streams.
package textinspect

import (
	"bufio"
	"errors"
	"io"
	"unicode/utf8"
)

type LineAnalysis struct {
	Style        string
	LineCount    int64
	LF           int64
	CRLF         int64
	CR           int64
	FinalNewline bool
}

type Result struct {
	Bytes        int64
	Encoding     string
	BOM          string
	UTF8Valid    bool
	LineAnalysis *LineAnalysis
}

// Inspect streams input and reports UTF encoding and newline metadata.
func Inspect(input io.Reader) (Result, error) {
	reader := bufio.NewReader(input)
	prefix, err := reader.Peek(4)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, bufio.ErrBufferFull) {
		return Result{}, err
	}
	bom, bomBytes := detectBOM(prefix)

	result := Result{BOM: bom, UTF8Valid: true}
	var lf, crlf, cr int64
	var pendingCR, hasContent bool
	var lastRune rune
	firstRune := true
	for {
		runeValue, size, readErr := reader.ReadRune()
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return Result{}, readErr
		}
		result.Bytes += int64(size)
		if runeValue == utf8.RuneError && size == 1 {
			result.UTF8Valid = false
		}
		if firstRune && bomBytes == 3 && runeValue == '\ufeff' {
			firstRune = false
			continue
		}
		firstRune = false
		hasContent = true
		lastRune = runeValue

		if pendingCR {
			if runeValue == '\n' {
				crlf++
				pendingCR = false
				continue
			}
			cr++
			pendingCR = false
		}
		switch runeValue {
		case '\r':
			pendingCR = true
		case '\n':
			lf++
		}
	}
	if pendingCR {
		cr++
	}

	result.Encoding = detectedEncoding(bom, result.UTF8Valid)
	if !result.UTF8Valid || (bom != "none" && bom != "utf-8") {
		return result, nil
	}
	analysis := LineAnalysis{LF: lf, CRLF: crlf, CR: cr}
	analysis.Style = newlineStyle(lf, crlf, cr)
	analysis.FinalNewline = hasContent && (lastRune == '\n' || lastRune == '\r')
	if hasContent {
		analysis.LineCount = lf + crlf + cr
		if !analysis.FinalNewline {
			analysis.LineCount++
		}
	}
	result.LineAnalysis = &analysis
	return result, nil
}

func detectBOM(prefix []byte) (string, int) {
	switch {
	case len(prefix) >= 4 && prefix[0] == 0x00 && prefix[1] == 0x00 && prefix[2] == 0xfe && prefix[3] == 0xff:
		return "utf-32be", 4
	case len(prefix) >= 4 && prefix[0] == 0xff && prefix[1] == 0xfe && prefix[2] == 0x00 && prefix[3] == 0x00:
		return "utf-32le", 4
	case len(prefix) >= 3 && prefix[0] == 0xef && prefix[1] == 0xbb && prefix[2] == 0xbf:
		return "utf-8", 3
	case len(prefix) >= 2 && prefix[0] == 0xfe && prefix[1] == 0xff:
		return "utf-16be", 2
	case len(prefix) >= 2 && prefix[0] == 0xff && prefix[1] == 0xfe:
		return "utf-16le", 2
	default:
		return "none", 0
	}
}

func detectedEncoding(bom string, validUTF8 bool) string {
	if bom != "none" {
		return bom
	}
	if validUTF8 {
		return "utf-8"
	}
	return "unknown"
}

func newlineStyle(lf, crlf, cr int64) string {
	kinds := 0
	style := "none"
	for _, value := range []struct {
		count int64
		name  string
	}{{lf, "lf"}, {crlf, "crlf"}, {cr, "cr"}} {
		if value.count > 0 {
			kinds++
			style = value.name
		}
	}
	if kinds > 1 {
		return "mixed"
	}
	return style
}
