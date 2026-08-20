package textinspect

import (
	"bytes"
	"reflect"
	"testing"
)

func TestInspectReportsMixedUTF8Lines(t *testing.T) {
	input := []byte("first\r\nsecond\nthird\rlast")
	result, err := Inspect(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.Bytes != int64(len(input)) || result.Encoding != "utf-8" || result.BOM != "none" || !result.UTF8Valid {
		t.Errorf("result = %+v, want UTF-8 byte metadata", result)
	}
	want := &LineAnalysis{Style: "mixed", LineCount: 4, LF: 1, CRLF: 1, CR: 1, FinalNewline: false}
	if !reflect.DeepEqual(result.LineAnalysis, want) {
		t.Errorf("LineAnalysis = %+v, want %+v", result.LineAnalysis, want)
	}
}

func TestInspectReportsUTF8BOMAndFinalNewline(t *testing.T) {
	input := append([]byte{0xef, 0xbb, 0xbf}, []byte("hello\r\n")...)
	result, err := Inspect(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.Encoding != "utf-8" || result.BOM != "utf-8" || result.LineAnalysis == nil {
		t.Fatalf("result = %+v, want UTF-8 BOM with line analysis", result)
	}
	if result.LineAnalysis.Style != "crlf" || result.LineAnalysis.LineCount != 1 || !result.LineAnalysis.FinalNewline {
		t.Errorf("LineAnalysis = %+v, want one final CRLF line", result.LineAnalysis)
	}
}

func TestInspectDoesNotGuessLinesForUnsupportedEncoding(t *testing.T) {
	input := []byte{0xff, 0xfe, 'a', 0, '\n', 0}
	result, err := Inspect(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.Encoding != "utf-16le" || result.BOM != "utf-16le" || result.UTF8Valid || result.LineAnalysis != nil {
		t.Errorf("result = %+v, want conservative UTF-16LE detection", result)
	}
}

func TestInspectEmptyInput(t *testing.T) {
	result, err := Inspect(bytes.NewReader(nil))
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	want := &LineAnalysis{Style: "none"}
	if result.Bytes != 0 || result.Encoding != "utf-8" || !reflect.DeepEqual(result.LineAnalysis, want) {
		t.Errorf("result = %+v, want empty UTF-8 text", result)
	}
}

func TestInspectReportsUnicodeFindings(t *testing.T) {
	input := []byte("café\npаssword\nhello\u200bworld\ncafe\u0301\nhello\u00a0world\nToday\u02bcs")
	result, err := Inspect(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.UnicodeAnalysis == nil {
		t.Fatal("UnicodeAnalysis = nil, want findings")
	}
	if result.UnicodeAnalysis.NonASCII != 6 || result.UnicodeAnalysis.Suspicious != 5 {
		t.Errorf("UnicodeAnalysis = %+v, want 6 non-ASCII and 5 suspicious", result.UnicodeAnalysis)
	}
	wantKinds := []string{"non_ascii", "cyrillic_letter", "invisible_format", "combining_mark", "unusual_whitespace", "modifier_letter"}
	for index, want := range wantKinds {
		if got := result.UnicodeAnalysis.Findings[index].Kind; got != want {
			t.Errorf("finding %d kind = %q, want %q", index, got, want)
		}
	}
	if finding := result.UnicodeAnalysis.Findings[2]; finding.Line != 3 || finding.Column != 6 || finding.CodePoint != "U+200B" {
		t.Errorf("zero-width finding = %+v, want line 3 column 6 U+200B", finding)
	}
}

func TestInspectTreatsCommonTypographyAsInformational(t *testing.T) {
	result, err := Inspect(bytes.NewReader([]byte("words—more\nToday’s")))
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.UnicodeAnalysis == nil || len(result.UnicodeAnalysis.Findings) != 2 {
		t.Fatalf("UnicodeAnalysis = %+v, want two findings", result.UnicodeAnalysis)
	}
	if result.UnicodeAnalysis.Findings[0].Suspicious {
		t.Errorf("em dash = %+v, want informational typography", result.UnicodeAnalysis.Findings[0])
	}
	if !result.UnicodeAnalysis.Findings[1].Suspicious {
		t.Errorf("curly apostrophe = %+v, want suspicious lookalike", result.UnicodeAnalysis.Findings[1])
	}
}

func FuzzInspect(f *testing.F) {
	f.Add([]byte("first\r\nsecond\n"))
	f.Add([]byte{0xff, 0xfe, 'x', 0})
	f.Fuzz(func(t *testing.T, input []byte) {
		result, err := Inspect(bytes.NewReader(input))
		if err != nil {
			t.Fatalf("Inspect() error = %v", err)
		}
		if result.Bytes != int64(len(input)) {
			t.Fatalf("Bytes = %d, want %d", result.Bytes, len(input))
		}
	})
}
