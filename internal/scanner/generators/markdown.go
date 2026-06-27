package generators

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xealgo/todo/internal/scanner"
)

// MarkdownGenerator writes scan results to a markdown file
type MarkdownGenerator struct {
	SourcePath string
	DestPath   string
}

var _ scanner.ResultsGenerator = (*MarkdownGenerator)(nil)

// NewMarkdownGenerator creates a new MarkdownWriter
func NewMarkdownGenerator(sourcePath string, destPath string) *MarkdownGenerator {
	return &MarkdownGenerator{
		SourcePath: sourcePath,
		DestPath:   destPath,
	}
}

// Generates the markdown artifact from the provided scan report.
func (mw *MarkdownGenerator) Generate(ctx context.Context, report scanner.ScanReport) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("code-scan-results_%s.md", timestamp)
	outputPath := filepath.Join(mw.DestPath, filename)

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write header
	fmt.Fprintf(writer, "# Code Comment Scan Results\n\n")
	fmt.Fprintf(writer, "**Directory Scanned:** `%s`  \n", mw.SourcePath)
	fmt.Fprintf(writer, "**Scan Date:** %s  \n", time.Now().Format("January 2, 2006 at 3:04 PM"))
	fmt.Fprintf(writer, "**Total Findings:** %d\n\n", report.Total)
	fmt.Fprintf(writer, "---\n\n")

	// Group by type
	byType := make(map[string][]scanner.ScanResult)
	for _, f := range report.Results {
		byType[f.Keyword] = append(byType[f.Keyword], f)
	}

	// Write summary
	fmt.Fprintf(writer, "## Summary\n\n")
	for _, keyword := range report.Keywords {
		if err = ctx.Err(); err != nil {
			return err
		}

		typeKey := strings.TrimSuffix(keyword, ":")
		count := len(byType[typeKey])
		if count > 0 {
			fmt.Fprintf(writer, "- **%s:** %d\n", typeKey, count)
		}
	}
	fmt.Fprintf(writer, "\n---\n\n")

	// Write detailed findings
	fmt.Fprintf(writer, "## Detailed Findings\n\n")
	for _, keyword := range report.Keywords {
		if err = ctx.Err(); err != nil {
			return err
		}

		typeKey := strings.TrimSuffix(keyword, ":")
		items := byType[typeKey]
		if len(items) == 0 {
			continue
		}

		fmt.Fprintf(writer, "### %s (%d)\n\n", typeKey, len(items))
		for _, item := range items {
			fmt.Fprintf(writer, "**File:** `%s` (Line %d)\n", item.Filename, item.Linenum)
			fmt.Fprintf(writer, "```\n%s\n```\n\n", item.Comment)
		}
	}

	fmt.Printf("✅ Results saved to: %s\n", filename)
	return nil
}
