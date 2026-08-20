package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/haritsAchmad/devkit-playground/internal/output"
	"github.com/haritsAchmad/devkit-playground/internal/tools/textinspect"
)

const textInspectUsage = `Usage:
  devkit [--json] text inspect <path>

Arguments:
  path                  regular text file to inspect

Line analysis is available for valid UTF-8, including UTF-8 with a BOM.
Unicode analysis reports non-ASCII code points and flags suspicious characters.
Other encodings are identified conservatively and do not receive line counts.
`

type textLineData struct {
	Style        string `json:"style"`
	LineCount    int64  `json:"line_count"`
	LF           int64  `json:"lf"`
	CRLF         int64  `json:"crlf"`
	CR           int64  `json:"cr"`
	FinalNewline bool   `json:"final_newline"`
}

type textInspectData struct {
	Path            string           `json:"path"`
	Bytes           int64            `json:"bytes"`
	Encoding        string           `json:"encoding"`
	BOM             string           `json:"bom"`
	UTF8Valid       bool             `json:"utf8_valid"`
	LineAnalysis    *textLineData    `json:"line_analysis"`
	UnicodeAnalysis *textUnicodeData `json:"unicode_analysis"`
}

type textUnicodeFindingData struct {
	Line       int64  `json:"line"`
	Column     int64  `json:"column"`
	Character  string `json:"character"`
	CodePoint  string `json:"code_point"`
	Kind       string `json:"kind"`
	Suspicious bool   `json:"suspicious"`
}

type textUnicodeData struct {
	NonASCII          int64                    `json:"non_ascii"`
	Suspicious        int64                    `json:"suspicious"`
	Findings          []textUnicodeFindingData `json:"findings"`
	FindingsTruncated bool                     `json:"findings_truncated"`
}

func runText(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	if len(args) == 0 {
		return writeFailure(stdout, stderr, jsonMode, "text", "invalid_usage", "text requires the inspect subcommand", ExitUsage)
	}
	if args[0] == "--help" {
		return writeTextInspectHelp(stdout, jsonMode)
	}
	if args[0] != "inspect" {
		return writeFailure(stdout, stderr, jsonMode, "text", "unknown_subcommand", "unknown text subcommand", ExitUsage)
	}
	return runTextInspect(args[1:], stdout, stderr, jsonMode)
}

func runTextInspect(args []string, stdout, stderr io.Writer, jsonMode bool) int {
	flags := flag.NewFlagSet("text inspect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return writeTextInspectHelp(stdout, jsonMode)
		}
		return writeFailure(stdout, stderr, jsonMode, "text inspect", "invalid_usage", err.Error(), ExitUsage)
	}
	if flags.NArg() != 1 {
		return writeFailure(stdout, stderr, jsonMode, "text inspect", "invalid_usage", "text inspect requires exactly one path", ExitUsage)
	}

	path := flags.Arg(0)
	info, err := os.Lstat(path)
	if err != nil {
		return writeFailure(stdout, stderr, jsonMode, "text inspect", "file_read_failed", "could not inspect input file", ExitOperation)
	}
	if !info.Mode().IsRegular() {
		return writeFailure(stdout, stderr, jsonMode, "text inspect", "not_regular_file", "input path is not a regular file", ExitData)
	}
	file, err := os.Open(path)
	if err != nil {
		return writeFailure(stdout, stderr, jsonMode, "text inspect", "file_read_failed", "could not open input file", ExitOperation)
	}
	defer file.Close()

	result, err := textinspect.Inspect(file)
	if err != nil {
		return writeFailure(stdout, stderr, jsonMode, "text inspect", "file_read_failed", "could not read input file", ExitOperation)
	}
	data := textInspectData{
		Path: path, Bytes: result.Bytes, Encoding: result.Encoding,
		BOM: result.BOM, UTF8Valid: result.UTF8Valid,
	}
	if result.LineAnalysis != nil {
		data.LineAnalysis = &textLineData{
			Style: result.LineAnalysis.Style, LineCount: result.LineAnalysis.LineCount,
			LF: result.LineAnalysis.LF, CRLF: result.LineAnalysis.CRLF, CR: result.LineAnalysis.CR,
			FinalNewline: result.LineAnalysis.FinalNewline,
		}
	}
	if result.UnicodeAnalysis != nil {
		data.UnicodeAnalysis = &textUnicodeData{
			NonASCII:          result.UnicodeAnalysis.NonASCII,
			Suspicious:        result.UnicodeAnalysis.Suspicious,
			FindingsTruncated: result.UnicodeAnalysis.FindingsTruncated,
			Findings:          make([]textUnicodeFindingData, 0, len(result.UnicodeAnalysis.Findings)),
		}
		for _, finding := range result.UnicodeAnalysis.Findings {
			data.UnicodeAnalysis.Findings = append(data.UnicodeAnalysis.Findings, textUnicodeFindingData{
				Line: finding.Line, Column: finding.Column, Character: finding.Rune,
				CodePoint: finding.CodePoint, Kind: finding.Kind, Suspicious: finding.Suspicious,
			})
		}
	}

	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "text inspect", data); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	return writeTextInspectHuman(stdout, data)
}

func writeTextInspectHuman(stdout io.Writer, data textInspectData) int {
	if _, err := fmt.Fprintf(stdout, "Path: %s\nBytes: %d\nEncoding: %s\nBOM: %s\nValid UTF-8: %t\n", data.Path, data.Bytes, data.Encoding, data.BOM, data.UTF8Valid); err != nil {
		return ExitInternal
	}
	if data.LineAnalysis == nil {
		if _, err := io.WriteString(stdout, "Line analysis: unavailable\n"); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	_, err := fmt.Fprintf(stdout, "Newline style: %s\nLines: %d\nLF: %d\nCRLF: %d\nCR: %d\nFinal newline: %t\n",
		data.LineAnalysis.Style, data.LineAnalysis.LineCount, data.LineAnalysis.LF,
		data.LineAnalysis.CRLF, data.LineAnalysis.CR, data.LineAnalysis.FinalNewline)
	if err != nil {
		return ExitInternal
	}
	if data.UnicodeAnalysis == nil {
		return ExitSuccess
	}
	if _, err := fmt.Fprintf(stdout, "Non-ASCII characters: %d\nSuspicious Unicode: %d\n", data.UnicodeAnalysis.NonASCII, data.UnicodeAnalysis.Suspicious); err != nil {
		return ExitInternal
	}
	for _, finding := range data.UnicodeAnalysis.Findings {
		marker := "info"
		if finding.Suspicious {
			marker = "suspicious"
		}
		if _, err := fmt.Fprintf(stdout, "  %d:%d  %s  %q  %s  [%s]\n", finding.Line, finding.Column, finding.CodePoint, finding.Character, finding.Kind, marker); err != nil {
			return ExitInternal
		}
	}
	if data.UnicodeAnalysis.FindingsTruncated {
		if _, err := io.WriteString(stdout, "  ... findings truncated\n"); err != nil {
			return ExitInternal
		}
	}
	return ExitSuccess
}

func writeTextInspectHelp(stdout io.Writer, jsonMode bool) int {
	if jsonMode {
		if err := output.WriteJSONSuccess(stdout, "text inspect help", map[string]string{"usage": textInspectUsage}); err != nil {
			return ExitInternal
		}
		return ExitSuccess
	}
	if _, err := io.WriteString(stdout, textInspectUsage); err != nil {
		return ExitInternal
	}
	return ExitSuccess
}
