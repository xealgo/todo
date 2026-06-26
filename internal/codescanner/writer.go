package codescanner

import (
	"fmt"
	"regexp"

	"github.com/fatih/color"
)

// ResultsWriter interface used to output the file scan results to different destinations (e.g., console, file, etc.)
// and or formats such as markdown, html, pdf, csv, etc.
type ResultsWriter interface {
	Write([]ScanResult) error
}

type SimpleConsoleWriter struct {
	//
}

var _ ResultsWriter = (*SimpleConsoleWriter)(nil)

var commentKeywordStripper = regexp.MustCompile(`^\s*//+\s*\w+:\s*`)
var commentSlashStripper = regexp.MustCompile(`/{2,3}\s*`)

// NewSimpleConsoleWriter creates a new instance of SimpleConsoleWriter.
func NewSimpleConsoleWriter() *SimpleConsoleWriter {
	return &SimpleConsoleWriter{}
}

// Write outputs the scan results to the console in a simple format.
func (w *SimpleConsoleWriter) Write(results []ScanResult) error {
	if len(results) == 0 {
		fmt.Println("No results found.")
	}

	resultsGroup := make(map[string][]ScanResult)
	filesSeen := make(map[string]struct{})

	for _, r := range results {
		resultsGroup[r.Filename] = append(resultsGroup[r.Filename], r)
	}

	for _, r := range results {
		if _, exists := filesSeen[r.Filename]; exists {
			continue
		}

		filesSeen[r.Filename] = struct{}{}

		entries, exists := resultsGroup[r.Filename]
		if !exists {
			continue
		}

		color.Blue(fmt.Sprintf("%s:\n", r.Filename))
		for _, entry := range entries {
			comment := commentKeywordStripper.ReplaceAllString(entry.Comment, "")
			comment = commentSlashStripper.ReplaceAllString(comment, "")

			fmt.Printf("\t%s: %s\n\t%s\n", color.GreenString(entry.Keyword), comment, color.MagentaString("LN: %d", entry.Linenum))
		}
	}

	return nil
}
