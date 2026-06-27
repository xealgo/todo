package generators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/fatih/color"
	"github.com/xealgo/todo/internal/scanner"
)

type ConsoleWriter struct {
	//
}

var _ scanner.ResultsGenerator = (*ConsoleWriter)(nil)

var commentKeywordStripper = regexp.MustCompile(`^\s*//+\s*\w+:\s*`)
var commentSlashStripper = regexp.MustCompile(`/{2,3}\s*`)

// NewConsoleWriter creates a new instance of ConsoleWriter.
func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{}
}

// Generate outputs the scan results to the console / terminal.
func (w *ConsoleWriter) Generate(ctx context.Context, report scanner.ScanReport) error {
	if report.Total == 0 {
		fmt.Println("No results found.")
	}

	resultsGroup := make(map[string][]scanner.ScanResult)
	filesSeen := make(map[string]struct{})

	for _, r := range report.Results {
		resultsGroup[r.Filename] = append(resultsGroup[r.Filename], r)
	}

	for _, r := range report.Results {
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
