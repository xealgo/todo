package scanner

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"
)

// DefaultKeywords defines the default keywords that the scanner will look for in the codebase.
var DefaultKeywords = []string{"TODO:", "HACK:", "NOTE:", "FIX:", "BUG:", "REFACTOR:", "TEMPORARY:"}

// Result data after scanning the codebase for specific comments.
type ScanResult struct {
	Filename string
	Linenum  int
	Keyword  string
	Comment  string
}

// ScanReport aggregates the results of a scan, including the total number of findings, the individual results, and the
// keywords that were searched for.
type ScanReport struct {
	Total    int
	Results  []ScanResult
	Keywords []string
}

// Scanner scans the codebase for specific comments.
type Scanner struct {
	keywords         [][]byte
	numWorkers       int
	srcPath          string
	targetFileSuffix string
	generators       []ResultsGenerator
}

// NewScanner initializes a new scanner with the given keywords.
func NewScanner(srcPath string) *Scanner {
	scanner := &Scanner{
		numWorkers:       runtime.NumCPU(),
		srcPath:          srcPath,
		targetFileSuffix: ".cs",
		generators:       []ResultsGenerator{},
	}

	scanner.SetKeywords(DefaultKeywords...)

	return scanner
}

// WithGenerator adds a ResultsGenerator to the scanner, allowing for flexible output of scan results.
func (s *Scanner) WithGenerator(g ResultsGenerator) *Scanner {
	s.generators = append(s.generators, g)
	return s
}

// SetNumWorkers sets the number of workers for the scanner, ensuring it is within a valid range.
func (s *Scanner) SetNumWorkers(n int) error {
	if n <= 0 {
		return nil
	}

	if n > 100 {
		return fmt.Errorf("number of workers must be between 1 and 100, got %d", n)
	}

	s.numWorkers = n
	return nil
}

// SetKeywords sets the keywords for the scanner, ensuring both lowercase and uppercase versions are stored.
func (s *Scanner) SetKeywords(keywords ...string) {
	s.keywords = make([][]byte, len(keywords)*2)

	for i, kw := range keywords {
		keyword := kw

		// Add the trailing : if it doesn't already have it.
		if !strings.HasSuffix(kw, ":") {
			keyword = kw + ":"
		}

		s.keywords[i*2] = []byte(strings.ToLower(keyword))
		s.keywords[i*2+1] = []byte(strings.ToUpper(keyword))
	}
}

// File enumeration for debugging or listing files before committing to the scan.
func (s *Scanner) Enumerate(ctx context.Context) ([]string, error) {
	results := make([]string, 0, 128)

	err := s.scanFiles(ctx, &results)
	if err != nil {
		return results, fmt.Errorf("failed to enumerate source files: %w", err)
	}

	return results, nil
}

// Process scans the codebase for the specified keywords and returns the results.
func (s *Scanner) Process(ctx context.Context, timeoutSeconds time.Duration) error {
	timeout, cancel := context.WithTimeout(ctx, timeoutSeconds)
	defer cancel()

	files := make([]string, 0, 128)

	if err := s.scanFiles(timeout, &files); err != nil {
		return fmt.Errorf("failed to scan source files in path %s: %w", s.srcPath, err)
	}

	numFiles := len(files)

	if numFiles == 0 {
		return fmt.Errorf("no source files found in %s", s.srcPath)
	}

	jobsChan := make(chan string, numFiles)
	errChan := make(chan error, numFiles)
	// We'll assume there are no more than 512 TODO comments in the codebase..
	// but even if there are, this should be plenty large enough to prevent blocking.
	outChan := make(chan ScanResult, 512)

	go func() {
		defer close(jobsChan)
		for _, file := range files {
			select {
			case jobsChan <- file:
			case <-timeout.Done():
				return
			}
		}
	}()

	var wg sync.WaitGroup

	for i := 0; i < s.numWorkers; i++ {
		wg.Add(1)
		go s.worker(timeout, &wg, jobsChan, outChan, errChan)
	}

	var finalResults []ScanResult
	doneCollecting := make(chan struct{})

	go func() {
		for res := range outChan {
			finalResults = append(finalResults, res)
		}
		close(doneCollecting)
	}()

	var workerErrors []error
	errorsDone := make(chan struct{})

	go func() {
		for err := range errChan {
			workerErrors = append(workerErrors, err)
		}
		close(errorsDone)
	}()

	wg.Wait()
	close(errChan)
	close(outChan)

	<-doneCollecting
	<-errorsDone

	slices.SortFunc(finalResults, func(a ScanResult, b ScanResult) int {
		return strings.Compare(a.Filename, b.Filename)
	})

	uniqueKeywordsFound := make(map[string]struct{})
	keywords := []string{}

	for _, r := range finalResults {
		if _, exits := uniqueKeywordsFound[r.Keyword]; !exits {
			keywords = append(keywords, r.Keyword)
		}

		uniqueKeywordsFound[r.Keyword] = struct{}{}
	}

	report := ScanReport{
		Total:    len(finalResults),
		Keywords: keywords,
		Results:  finalResults,
	}

	var writerErrors []error

	if len(s.generators) == 0 {
		writerErrors = append(writerErrors, fmt.Errorf("no output writers available"))
	} else {
		for _, w := range s.generators {
			if err := w.Generate(timeout, report); err != nil {
				writerErrors = append(writerErrors, fmt.Errorf("error writing result: %w", err))
			}
		}
	}

	allErrors := errors.Join(workerErrors...)
	allErrors = errors.Join(allErrors, errors.Join(writerErrors...))

	return allErrors
}

// worker processes files from the jobs channel, scanning each for the specified keywords and sending results to the out channel.
func (s *Scanner) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan string, out chan<- ScanResult, errs chan<- error) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			select {
			case errs <- ctx.Err():
			default:
			}
			return
		case file, ok := <-jobs:
			if !ok {
				return
			}

			results, err := s.ScanFile(file)
			if err != nil {
				select {
				case errs <- fmt.Errorf("error scanning file %s: %w", file, err):
				default:
				}
				continue
			}

			for _, result := range results {
				select {
				case out <- result:
				case <-ctx.Done():
					select {
					case errs <- ctx.Err():
					default:
					}
					return
				}
			}
		}
	}
}

// ScanFile scans a single file for the specified keywords and returns the results.
func (s *Scanner) ScanFile(filePath string) ([]ScanResult, error) {
	var results []ScanResult

	file, err := os.Open(path.Join(s.srcPath, filePath))
	if err != nil {
		return results, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()

		for _, keyword := range s.keywords {
			if bytes.Contains(line, keyword) {
				comment := strings.TrimSpace(string(line))
				startLine := lineNum

				// TODO: This is pretty basic and not very optimized. It needs to be able to
				// ensure we don't bleed over into any other standard C# doc keywords, tags, or
				// carries over to the actual function. This should be removed from here and
				// written as a pure function so that we can test it in isolation.
				for scanner.Scan() {
					lineNum++
					nextLine := scanner.Bytes()
					trimmed := bytes.TrimSpace(nextLine)

					// Check if it's a comment continuation (starts with //)
					if !bytes.HasPrefix(trimmed, []byte("//")) || bytes.HasPrefix(trimmed, []byte("/// <")) {
						break
					}

					hasKeyword := false
					for _, kw := range s.keywords {
						if bytes.Contains(trimmed, kw) {
							hasKeyword = true
							break
						}
					}

					if hasKeyword {
						break
					}

					comment += " " + strings.TrimSpace(string(nextLine))
				}

				// TODO: Maybe convert keywords to a struct to avoid the additional cast.
				results = append(results, ScanResult{
					Filename: filePath,
					Linenum:  startLine,
					Keyword:  strings.TrimSuffix(string(keyword), ":"),
					Comment:  comment,
				})

				break
			}
		}
	}

	if scanner.Err() != nil {
		return results, fmt.Errorf("error scanning file %s: %w", filePath, scanner.Err())
	}

	return results, nil
}

// scanFiles scans for all files matching the target file suffix in the given source path and writes the path results
// to the provided channel.
func (s *Scanner) scanFiles(ctx context.Context, results *[]string) error {
	info, err := os.Stat(s.srcPath)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s a file, not a directory", s.srcPath)
		}
	} else if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%s does not exist", s.srcPath)
	} else {
		return fmt.Errorf("error scanning %s: %w", s.srcPath, err)
	}

	err = filepath.WalkDir(s.srcPath, func(path string, d os.DirEntry, err error) error {
		ctxErr := ctx.Err()
		if ctxErr != nil {
			// Stop walking if the context is cancelled
			return filepath.SkipAll
		}

		if err != nil {
			return fmt.Errorf("error scanning %s: %w", path, err)
		}

		if !d.IsDir() && strings.HasSuffix(path, s.targetFileSuffix) {
			relPath, err := filepath.Rel(s.srcPath, path)
			if err != nil {
				return err
			}
			*results = append(*results, relPath)
		}

		return nil
	})

	ctxErr := ctx.Err()
	if ctxErr != nil {
		return fmt.Errorf("scanning interrupted: %w", ctxErr)
	}

	return err
}
